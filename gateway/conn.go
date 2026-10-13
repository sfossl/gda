package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sfossl/bfrm/types"
)

const (
	DefaultGatewayURL = "wss://gateway.discord.gg/?v=" + types.APIVersion + "&encoding=json"
)

var (
	ErrFatalGatewayClose = errors.New("bfrm/gateway: non-reconnectable gateway close code")
)

type Config struct {
	Token          string
	Intents        types.Intent
	ShardID        int
	ShardCount     int
	LargeThreshold int
	GatewayURL     string
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
		return errors.New("bfrm/gateway: websocket connection is nil")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return conn.WriteText(ctx, data)
}

func (c *Client) sendHeartbeat(ctx context.Context, seq *int64) error {
	p := types.GatewayPayload{
		Op: types.OpcodeHeartbeat,
	}
	if seq != nil {
		p.S = seq
	}
	return c.sendJSON(ctx, p)
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
			Browser: "bfrm",
			Device:  "bfrm",
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

func (c *Client) Connect(ctx context.Context) error {
	backoff := 1 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		endpointURL := c.cfg.GatewayURL
		if c.state.CanResume() && c.state.ResumeGatewayURL() != "" {
			endpointURL = strings.TrimRight(c.state.ResumeGatewayURL(), "/") + "/?v=" + types.APIVersion + "&encoding=json"
		}

		c.state.SetState(StateConnecting)
		conn, err := Dial(ctx, endpointURL)
		if err != nil {
			select {
			case <-time.After(backoff):
				backoff *= 2
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		c.mu.Lock()
		c.wsConn = conn
		c.mu.Unlock()

		connErr := c.runConnection(ctx, conn)

		c.mu.Lock()
		if c.hm != nil {
			c.hm.Stop()
			c.hm = nil
		}
		c.wsConn = nil
		c.mu.Unlock()

		_ = conn.Close(1000, "")

		if connErr != nil {
			var closeErr CloseError
			if errors.As(connErr, &closeErr) {
				if !types.IsReconnectable(closeErr.Code) {
					c.state.SetState(StateClosed)
					return fmt.Errorf("%w: %d %s", ErrFatalGatewayClose, closeErr.Code, closeErr.Reason)
				}
			}
		}

		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Client) runConnection(ctx context.Context, conn *Conn) error {
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	for {
		data, err := conn.ReadText(connCtx)
		if err != nil {
			return err
		}

		var payload types.GatewayPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
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
			return errors.New("bfrm/gateway: discord requested reconnect")

		case types.OpcodeInvalidSession:
			var resumable bool
			_ = json.Unmarshal(payload.D, &resumable)
			if !resumable {
				c.state.Reset()
			}
			return errors.New("bfrm/gateway: invalid session")

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

	if c.hm != nil {
		c.hm.Stop()
		c.hm = nil
	}
	if c.wsConn != nil {
		err := c.wsConn.Close(1000, "client closing")
		c.wsConn = nil
		c.state.SetState(StateClosed)
		return err
	}
	return nil
}
