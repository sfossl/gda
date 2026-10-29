package gateway

import (
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/sfossl/gda/types"
)

const (
	DefaultGatewayURL = "wss://gateway.discord.gg/?v=" + types.APIVersion + "&encoding=json"
)

var (
	ErrFatalGatewayClose = errors.New("non-reconnectable gateway close code")
)

type Config struct {
	Token          string
	Intents        types.Intent
	ShardID        int
	ShardCount     int
	LargeThreshold int
	GatewayURL     string
	Compress       bool
}

type Client struct {
	cfg        Config
	state      *SessionState
	dispatcher *EventDispatcher
	mu         sync.Mutex
	wsConn     *Conn
	hm         *HeartbeatManager
}

func NewClient(cfg Config, dispatcher *EventDispatcher) *Client {
	if cfg.GatewayURL == "" {
		cfg.GatewayURL = DefaultGatewayURL
	}
	if cfg.LargeThreshold <= 0 {
		cfg.LargeThreshold = 50
	}
	if cfg.ShardCount <= 0 {
		cfg.ShardCount = 1
	}
	return &Client{
		cfg:        cfg,
		state:      NewSessionState(),
		dispatcher: dispatcher,
	}
}

func (c *Client) State() *SessionState {
	return c.state
}

func (c *Client) Dispatcher() *EventDispatcher {
	return c.dispatcher
}

func (c *Client) sendJSON(ctx context.Context, v any) error {
	c.mu.Lock()
	conn := c.wsConn
	c.mu.Unlock()

	if conn == nil {
		return errors.New("websocket connection is nil")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return conn.WriteText(ctx, data)
}

func (c *Client) sendHeartbeat(ctx context.Context, seq *int64) error {
	var d json.RawMessage
	if seq != nil {
		d = []byte(strconv.FormatInt(*seq, 10))
	} else {
		d = []byte("null")
	}
	return c.sendJSON(ctx, types.GatewayPayload{
		Op: types.OpcodeHeartbeat,
		D:  d,
	})
}

func (c *Client) sendIdentify(ctx context.Context) error {
	p := types.GatewayPayload{
		Op: types.OpcodeIdentify,
	}
	var shard *[2]int
	if c.cfg.ShardCount > 1 {
		shard = &[2]int{c.cfg.ShardID, c.cfg.ShardCount}
	}

	identify := types.IdentifyPayload{
		Token: c.cfg.Token,
		Properties: types.ConnectionProperties{
			OS:      runtime.GOOS,
			Browser: "gda",
			Device:  "gda",
		},
		LargeThreshold: c.cfg.LargeThreshold,
		Shard:          shard,
		Intents:        c.cfg.Intents,
	}

	data, err := json.Marshal(identify)
	if err != nil {
		return err
	}
	p.D = data
	return c.sendJSON(ctx, p)
}

func (c *Client) sendResume(ctx context.Context) error {
	seq, _ := c.state.Sequence()
	resume := types.ResumePayload{
		Token:     c.cfg.Token,
		SessionID: c.state.SessionID(),
		Seq:       seq,
	}

	data, err := json.Marshal(resume)
	if err != nil {
		return err
	}

	p := types.GatewayPayload{
		Op: types.OpcodeResume,
		D:  data,
	}
	return c.sendJSON(ctx, p)
}

func (c *Client) UpdatePresence(ctx context.Context, presence types.PresenceUpdate) error {
	data, err := json.Marshal(presence)
	if err != nil {
		return err
	}
	return c.sendJSON(ctx, types.GatewayPayload{
		Op: types.OpcodePresenceUpdate,
		D:  data,
	})
}

func (c *Client) Connect(ctx context.Context) error {
	backoff := 1 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if c.state.State() == StateClosed {
			return nil
		}

		endpointURL := c.cfg.GatewayURL
		if c.state.CanResume() && c.state.ResumeGatewayURL() != "" {
			endpointURL = c.state.ResumeGatewayURL()
		}
		if u, err := url.Parse(endpointURL); err == nil {
			q := u.Query()
			q.Set("v", types.APIVersion)
			q.Set("encoding", "json")
			if c.cfg.Compress {
				q.Set("compress", "zlib-stream")
			}
			u.RawQuery = q.Encode()
			endpointURL = u.String()
		}

		c.state.SetState(StateConnecting)
		conn, err := Dial(ctx, endpointURL)
		if err != nil {
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
				backoff *= 2
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				continue
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}

		c.mu.Lock()
		c.wsConn = conn
		c.mu.Unlock()

		connectedAt := time.Now()
		connErr := c.runConnection(ctx, conn)
		if time.Since(connectedAt) > 10*time.Second || c.state.State() == StateReady {
			backoff = 1 * time.Second
		}

		c.mu.Lock()
		if c.hm != nil {
			c.hm.Stop()
			c.hm = nil
		}
		c.wsConn = nil
		c.mu.Unlock()

		_ = conn.Close(1000, "")

		if c.state.State() == StateClosed {
			return nil
		}

		if connErr != nil {
			var closeErr CloseError
			if errors.As(connErr, &closeErr) {
				if !types.IsReconnectable(closeErr.Code) {
					c.state.SetState(StateClosed)
					return fmt.Errorf("%w: %d %s", ErrFatalGatewayClose, closeErr.Code, closeErr.Reason)
				}
			}
		}

		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (c *Client) runConnection(ctx context.Context, conn *Conn) error {
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	stopWatcher := context.AfterFunc(connCtx, func() {
		_ = conn.Close(1000, "context done")
	})
	defer stopWatcher()

	var getPayload func() (*types.GatewayPayload, error)

	if c.cfg.Compress {
		pr, pw := io.Pipe()
		defer pr.Close()
		defer pw.Close()

		go func() {
			for {
				data, err := conn.ReadText(connCtx)
				if err != nil {
					_ = pw.CloseWithError(err)
					return
				}
				if _, err := pw.Write(data); err != nil {
					return
				}
			}
		}()

		zr, err := zlib.NewReader(pr)
		if err != nil {
			return err
		}
		defer zr.Close()

		dec := json.NewDecoder(zr)
		getPayload = func() (*types.GatewayPayload, error) {
			var p types.GatewayPayload
			if err := dec.Decode(&p); err != nil {
				return nil, err
			}
			return &p, nil
		}
	} else {
		getPayload = func() (*types.GatewayPayload, error) {
			for {
				data, err := conn.ReadText(connCtx)
				if err != nil {
					return nil, err
				}
				var p types.GatewayPayload
				if err := json.Unmarshal(data, &p); err != nil {
					continue
				}
				return &p, nil
			}
		}
	}

	for {
		payload, err := getPayload()
		if err != nil {
			return err
		}

		if payload.S != nil {
			c.state.SetSequence(*payload.S)
		}

		switch payload.Op {
		case types.OpcodeHello:
			var hello types.HelloPayload
			if err := json.Unmarshal(payload.D, &hello); err != nil {
				return err
			}

			interval := time.Duration(hello.HeartbeatInterval) * time.Millisecond
			hm := NewHeartbeatManager(
				interval,
				func(seq *int64) error {
					return c.sendHeartbeat(connCtx, seq)
				},
				func() *int64 {
					if seq, ok := c.state.Sequence(); ok {
						return &seq
					}
					return nil
				},
			)

			c.mu.Lock()
			if c.hm != nil {
				c.hm.Stop()
			}
			c.hm = hm
			c.mu.Unlock()

			errChan := hm.Start(connCtx)
			go func() {
				if err, ok := <-errChan; ok && err != nil {
					connCancel()
					_ = conn.Close(4000, "heartbeat ack timeout")
				}
			}()

			if c.state.CanResume() {
				c.state.SetState(StateResuming)
				if err := c.sendResume(connCtx); err != nil {
					return err
				}
			} else {
				if err := c.sendIdentify(connCtx); err != nil {
					return err
				}
			}

		case types.OpcodeHeartbeatACK:
			c.mu.Lock()
			hm := c.hm
			c.mu.Unlock()
			if hm != nil {
				hm.Acknowledge()
			}

		case types.OpcodeHeartbeat:
			var seqPtr *int64
			if seq, ok := c.state.Sequence(); ok {
				seqPtr = &seq
			}
			_ = c.sendHeartbeat(connCtx, seqPtr)

		case types.OpcodeReconnect:
			return errors.New("discord requested reconnect")

		case types.OpcodeInvalidSession:
			var resumable bool
			_ = json.Unmarshal(payload.D, &resumable)
			if !resumable {
				c.state.Reset()
			}
			return errors.New("invalid session")

		case types.OpcodeDispatch:
			if payload.T == nil {
				continue
			}

			switch *payload.T {
			case types.EventReady:
				var ready types.ReadyEvent
				if err := json.Unmarshal(payload.D, &ready); err == nil {
					c.state.SetReady(ready.SessionID, ready.ResumeGatewayURL)
				}
			case types.EventResumed:
				c.state.SetState(StateReady)
			}

			c.dispatcher.Dispatch(ctx, *payload.T, payload.D)
		}
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state.SetState(StateClosed)
	if c.hm != nil {
		c.hm.Stop()
		c.hm = nil
	}
	if c.wsConn != nil {
		err := c.wsConn.Close(1000, "client closing")
		c.wsConn = nil
		return err
	}
	return nil
}

func (c *Client) Latency() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hm != nil {
		return c.hm.Latency()
	}
	return 0
}

func (c *Client) String() string {
	if c == nil {
		return "<nil>"
	}
	redacted := "[REDACTED]"
	if c.cfg.Token == "" {
		redacted = "[EMPTY]"
	}
	return fmt.Sprintf("gateway.Client{Token:%s, Shard:%d/%d, State:%s}",
		redacted, c.cfg.ShardID, c.cfg.ShardCount, c.state.State())
}
