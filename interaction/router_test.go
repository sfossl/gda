package interaction

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestRouterCommand(t *testing.T) {
	router := NewRouter()
	var executed bool

	router.OnCommand("ping", func(ctx *Context) error {
		executed = true
		return ctx.Reply(Response{Content: "pong"})
	})

	in := &types.Interaction{
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "ping",
		},
	}
	ctx := NewContext(context.Background(), in, nil)
	if err := router.Handle(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Fatalf("expected handler to execute")
	}

	select {
	case res := <-ctx.ResponseChannel():
		if res.Data.Content != "pong" {
			t.Fatalf("expected content pong, got %s", res.Data.Content)
		}
	default:
		t.Fatalf("expected response")
	}
}

func TestRouterSubcommand(t *testing.T) {
	router := NewRouter()
	var executed bool

	router.OnSubcommand("pws", "status", func(ctx *Context) error {
		executed = true
		return ctx.Reply(Response{Content: "running"})
	})

	in := &types.Interaction{
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "pws",
			Options: []types.ApplicationCommandOption{
				{
					Name: "status",
					Type: 1,
				},
			},
		},
	}
	ctx := NewContext(context.Background(), in, nil)
	if err := router.Handle(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Fatalf("expected subcommand handler to execute")
	}
}

func TestRouterComponentPrefix(t *testing.T) {
	router := NewRouter()
	var paramReceived string

	router.OnComponent("task_done", func(ctx *Context) error {
		paramReceived = ctx.ComponentParam()
		return ctx.Reply(Response{Content: "done"})
	})

	in := &types.Interaction{
		Type: types.InteractionTypeMessageComponent,
		Data: &types.InteractionData{
			CustomID: "task_done:12345",
		},
	}
	ctx := NewContext(context.Background(), in, nil)
	if err := router.Handle(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paramReceived != "12345" {
		t.Fatalf("expected param 12345, got %s", paramReceived)
	}
}

func TestRouterMiddleware(t *testing.T) {
	router := NewRouter()
	var count int32

	router.Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			atomic.AddInt32(&count, 1)
			return next(ctx)
		}
	})

	router.OnCommand("test", func(ctx *Context) error {
		return nil
	})

	in := &types.Interaction{
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{Name: "test"},
	}
	_ = router.Handle(NewContext(context.Background(), in, nil))
	if atomic.LoadInt32(&count) != 1 {
		t.Fatalf("expected middleware count 1, got %d", atomic.LoadInt32(&count))
	}
}

func TestRouterWebhookHandler(t *testing.T) {
	router := NewRouter()
	router.OnCommand("hello", func(ctx *Context) error {
		return ctx.Reply(Response{Content: "world"})
	})

	h := router.WebhookHandler(nil)

	pingBody := []byte(`{"type":1}`)
	req := httptest.NewRequest("POST", "/interactions", bytes.NewReader(pingBody))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != `{"type":1}` {
		t.Fatalf("expected pong response, got %s", rr.Body.String())
	}

	cmdBody := []byte(`{"type":2,"data":{"name":"hello"}}`)
	req2 := httptest.NewRequest("POST", "/interactions", bytes.NewReader(cmdBody))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr2.Code)
	}

	var res types.InteractionResponse
	if err := json.NewDecoder(rr2.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Type != types.CallbackTypeChannelMessageWithSource {
		t.Fatalf("expected callback type 4, got %d", res.Type)
	}
	if res.Data.Content != "world" {
		t.Fatalf("expected world, got %s", res.Data.Content)
	}
}

