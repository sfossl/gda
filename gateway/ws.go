package gateway

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sfossl/gda/types"
)

const (
	wsGUID           = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	wsOpContinuation = 0x0
	wsOpText         = 0x1
	wsOpBinary       = 0x2
	wsOpClose        = 0x8
	wsOpPing         = 0x9
	wsOpPong         = 0xa
	wsMaxPayloadLen  = types.MaxPayloadSize
)

var (
	ErrWSClosed       = errors.New("gda/gateway: websocket connection is closed")
	ErrWSPayloadLimit = errors.New("gda/gateway: websocket frame payload exceeds 10MB limit")
)

type CloseError struct {
	Code   int
	Reason string
}

func (e CloseError) Error() string {
	return fmt.Sprintf("websocket closed: %d %s", e.Code, e.Reason)
}

type Conn struct {
	netConn net.Conn
	reader  *bufio.Reader
	writeMu sync.Mutex
	closed  atomic.Bool
}

func Dial(ctx context.Context, endpointURL string) (*Conn, error) {
	u, err := url.Parse(endpointURL)
	if err != nil {
		return nil, err
	}

	host := u.Host
	port := "443"
	useTLS := true

	if strings.Contains(host, ":") {
		h, p, err := net.SplitHostPort(host)
		if err == nil {
			host = h
			port = p
		}
	}

	if u.Scheme == "ws" {
		useTLS = false
		if port == "443" {
			port = "80"
		}
	}

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	addr := net.JoinHostPort(host, port)
	var netConn net.Conn

	if useTLS {
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config: &tls.Config{
				ServerName: host,
			},
		}
		netConn, err = tlsDialer.DialContext(ctx, "tcp", addr)
	} else {
		netConn, err = dialer.DialContext(ctx, "tcp", addr)
	}

	if err != nil {
		return nil, err
	}

	rawKey := make([]byte, 16)
	if _, err := rand.Read(rawKey); err != nil {
		_ = netConn.Close()
		return nil, err
	}
	secKey := base64.StdEncoding.EncodeToString(rawKey)

	reqPath := u.Path
	if reqPath == "" {
		reqPath = "/"
	}
	if u.RawQuery != "" {
		reqPath += "?" + u.RawQuery
	}

	reqBuf := bytes.NewBuffer(nil)
	fmt.Fprintf(reqBuf, "GET %s HTTP/1.1\r\n", reqPath)
	fmt.Fprintf(reqBuf, "Host: %s\r\n", u.Host)
	fmt.Fprintf(reqBuf, "Upgrade: websocket\r\n")
	fmt.Fprintf(reqBuf, "Connection: Upgrade\r\n")
	fmt.Fprintf(reqBuf, "Sec-WebSocket-Key: %s\r\n", secKey)
	fmt.Fprintf(reqBuf, "Sec-WebSocket-Version: 13\r\n")
	fmt.Fprintf(reqBuf, "User-Agent: %s\r\n\r\n", types.UserAgent)

	_ = netConn.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := netConn.Write(reqBuf.Bytes()); err != nil {
		_ = netConn.Close()
		return nil, err
	}

	reader := bufio.NewReader(netConn)
	resp, err := http.ReadResponse(reader, &http.Request{Method: "GET"})
	if err != nil {
		_ = netConn.Close()
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		_ = netConn.Close()
		return nil, fmt.Errorf("gda/gateway: handshake failed with status %d", resp.StatusCode)
	}

	h := sha1.New()
	h.Write([]byte(secKey + wsGUID))
	expectedAccept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if resp.Header.Get("Sec-WebSocket-Accept") != expectedAccept {
		_ = netConn.Close()
		return nil, errors.New("gda/gateway: invalid sec-websocket-accept header")
	}

	_ = netConn.SetDeadline(time.Time{})

	return &Conn{
		netConn: netConn,
		reader:  reader,
	}, nil
}

func (c *Conn) WriteText(ctx context.Context, payload []byte) error {
	return c.writeFrame(ctx, wsOpText, payload)
}

func (c *Conn) writeFrame(ctx context.Context, opcode byte, payload []byte) error {
	if opcode != wsOpClose && c.closed.Load() {
		return ErrWSClosed
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	header := make([]byte, 0, 14)
	header = append(header, 0x80|opcode)

	length := len(payload)
	switch {
	case length < 126:
		header = append(header, 0x80|byte(length))
	case length <= 65535:
		header = append(header, 0x80|126, byte(length>>8), byte(length))
	default:
		header = append(header, 0x80|127)
		var lenBuf [8]byte
		binary.BigEndian.PutUint64(lenBuf[:], uint64(length))
		header = append(header, lenBuf[:]...)
	}

	var mask [4]byte
	if _, err := rand.Read(mask[:]); err != nil {
		return err
	}
	header = append(header, mask[:]...)

	maskedPayload := make([]byte, length)
	maskKey := binary.LittleEndian.Uint32(mask[:])
	mask64 := uint64(maskKey) | (uint64(maskKey) << 32)
	i := 0
	for ; i+8 <= length; i += 8 {
		val := binary.LittleEndian.Uint64(payload[i : i+8])
		binary.LittleEndian.PutUint64(maskedPayload[i:i+8], val^mask64)
	}
	for ; i < length; i++ {
		maskedPayload[i] = payload[i] ^ mask[i%4]
	}

	buf := append(header, maskedPayload...)
	_ = c.netConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := c.netConn.Write(buf)
	return err
}

func (c *Conn) ReadText(ctx context.Context) ([]byte, error) {
	var fragBuf []byte
	for {
		if c.closed.Load() {
			return nil, ErrWSClosed
		}

		b0, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		fin := (b0 & 0x80) != 0
		opcode := b0 & 0x0f

		b1, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		isMasked := (b1 & 0x80) != 0
		lenByte := b1 & 0x7f

		var payloadLen uint64
		switch {
		case lenByte < 126:
			payloadLen = uint64(lenByte)
		case lenByte == 126:
			var l [2]byte
			if _, err := io.ReadFull(c.reader, l[:]); err != nil {
				return nil, err
			}
			payloadLen = uint64(binary.BigEndian.Uint16(l[:]))
		default:
			var l [8]byte
			if _, err := io.ReadFull(c.reader, l[:]); err != nil {
				return nil, err
			}
			payloadLen = binary.BigEndian.Uint64(l[:])
		}

		if payloadLen > wsMaxPayloadLen || uint64(len(fragBuf))+payloadLen > wsMaxPayloadLen {
			return nil, ErrWSPayloadLimit
		}

		var mask [4]byte
		if isMasked {
			if _, err := io.ReadFull(c.reader, mask[:]); err != nil {
				return nil, err
			}
		}

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return nil, err
		}

		if isMasked {
			for i := uint64(0); i < payloadLen; i++ {
				payload[i] ^= mask[i%4]
			}
		}

		switch opcode {
		case wsOpText, wsOpBinary:
			if fin {
				return payload, nil
			}
			fragBuf = append(fragBuf, payload...)
		case wsOpContinuation:
			fragBuf = append(fragBuf, payload...)
			if fin {
				res := fragBuf
				fragBuf = nil
				return res, nil
			}
		case wsOpPing:
			_ = c.writeFrame(ctx, wsOpPong, payload)
		case wsOpClose:
			c.closed.Store(true)
			var code int
			var reason string
			if len(payload) >= 2 {
				code = int(binary.BigEndian.Uint16(payload[:2]))
				reason = string(payload[2:])
			}
			_ = c.netConn.Close()
			return nil, CloseError{Code: code, Reason: reason}
		}
	}
}

func (c *Conn) Close(code int, reason string) error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload[:2], uint16(code))
	copy(payload[2:], reason)

	_ = c.writeFrame(context.Background(), wsOpClose, payload)
	return c.netConn.Close()
}
