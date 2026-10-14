package gateway

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockWSServerConn struct {
	conn   net.Conn
	rw     *bufio.ReadWriter
	reader *bufio.Reader
}

func acceptMockWS(w http.ResponseWriter, r *http.Request) (*mockWSServerConn, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return nil, http.ErrNotSupported
	}

	netConn, rw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	secKey := r.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(secKey + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"

	if _, err := rw.WriteString(resp); err != nil {
		_ = netConn.Close()
		return nil, err
	}
	if err := rw.Flush(); err != nil {
		_ = netConn.Close()
		return nil, err
	}

	return &mockWSServerConn{
		conn:   netConn,
		rw:     rw,
		reader: bufio.NewReader(netConn),
	}, nil
}

func (s *mockWSServerConn) WriteText(payload []byte) error {
	header := make([]byte, 0, 10)
	header = append(header, 0x81)

	length := len(payload)
	if length < 126 {
		header = append(header, byte(length))
	} else if length <= 65535 {
		header = append(header, 126, byte(length>>8), byte(length))
	} else {
		header = append(header, 127)
		var lenBuf [8]byte
		binary.BigEndian.PutUint64(lenBuf[:], uint64(length))
		header = append(header, lenBuf[:]...)
	}

	buf := append(header, payload...)
	_, err := s.conn.Write(buf)
	return err
}

func (s *mockWSServerConn) WriteBinary(payload []byte) error {
	header := make([]byte, 0, 10)
	header = append(header, 0x82)

	length := len(payload)
	if length < 126 {
		header = append(header, byte(length))
	} else if length <= 65535 {
		header = append(header, 126, byte(length>>8), byte(length))
	} else {
		header = append(header, 127)
		var lenBuf [8]byte
		binary.BigEndian.PutUint64(lenBuf[:], uint64(length))
		header = append(header, lenBuf[:]...)
	}

	buf := append(header, payload...)
	_, err := s.conn.Write(buf)
	return err
}

func (s *mockWSServerConn) ReadText() ([]byte, error) {
	b0, err := s.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	_ = b0

	b1, err := s.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	isMasked := (b1 & 0x80) != 0
	lenByte := b1 & 0x7f

	var payloadLen uint64
	if lenByte < 126 {
		payloadLen = uint64(lenByte)
	} else if lenByte == 126 {
		var l [2]byte
		if _, err := io.ReadFull(s.reader, l[:]); err != nil {
			return nil, err
		}
		payloadLen = uint64(binary.BigEndian.Uint16(l[:]))
	} else {
		var l [8]byte
		if _, err := io.ReadFull(s.reader, l[:]); err != nil {
			return nil, err
		}
		payloadLen = binary.BigEndian.Uint64(l[:])
	}

	var mask [4]byte
	if isMasked {
		if _, err := io.ReadFull(s.reader, mask[:]); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(s.reader, payload); err != nil {
		return nil, err
	}

	if isMasked {
		for i := uint64(0); i < payloadLen; i++ {
			payload[i] ^= mask[i%4]
		}
	}

	return payload, nil
}

func (s *mockWSServerConn) Close() error {
	return s.conn.Close()
}

func TestNativeWSDialAndFrames(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()

		msg, err := sConn.ReadText()
		if err != nil {
			return
		}

		_ = sConn.WriteText([]byte("echo:" + string(msg)))
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	ctx := context.Background()
	clientConn, err := Dial(ctx, wsURL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer clientConn.Close(1000, "closing")

	if err := clientConn.WriteText(ctx, []byte("hello")); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reply, err := clientConn.ReadText(ctx)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if string(reply) != "echo:hello" {
		t.Fatalf("expected echo:hello, got %s", string(reply))
	}
}

func TestNativeWSFragmentedFrames(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()

		frame1 := []byte{0x01, 0x05, 'p', 'i', 'e', 'c', 'e'}
		frame2 := []byte{0x80, 0x05, '1', '2', '3', '4', '5'}
		_, _ = sConn.conn.Write(frame1)
		_, _ = sConn.conn.Write(frame2)
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	ctx := context.Background()
	clientConn, err := Dial(ctx, wsURL)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer clientConn.Close(1000, "closing")

	msg, err := clientConn.ReadText(ctx)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if string(msg) != "piece12345" {
		t.Fatalf("expected piece12345, got %s", string(msg))
	}
}


