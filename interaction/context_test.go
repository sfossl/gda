package interaction

import (
	"context"
	"testing"

	"github.com/sfossl/gda/builder"
	"github.com/sfossl/gda/types"
)

func TestContextReply(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "tok",
	}
	ctx := NewContext(context.Background(), in, nil)

	err := ctx.Reply(Response{
		Content:   "pong",
		Ephemeral: true,
	})
	if err != nil {
		t.Fatalf("unexpected reply error: %v", err)
	}

	select {
	case res := <-ctx.ResponseChannel():
		if res.Type != types.CallbackTypeChannelMessageWithSource {
			t.Fatalf("expected callback type 4, got %d", res.Type)
		}
		if res.Data.Content != "pong" {
			t.Fatalf("expected content pong, got %s", res.Data.Content)
		}
		if res.Data.Flags != types.MessageFlagEphemeral {
			t.Fatalf("expected ephemeral flag")
		}
	default:
		t.Fatalf("expected response in channel")
	}

	err = ctx.Reply(Response{Content: "again"})
	if err != ErrAlreadyResponded {
		t.Fatalf("expected ErrAlreadyResponded, got %v", err)
	}
}

func TestContextDefer(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(101),
		Token: "tok",
	}
	ctx := NewContext(context.Background(), in, nil)

	err := ctx.Defer(false)
	if err != nil {
		t.Fatalf("unexpected defer error: %v", err)
	}

	select {
	case res := <-ctx.ResponseChannel():
		if res.Type != types.CallbackTypeDeferredChannelMessageWithSource {
			t.Fatalf("expected callback type 5, got %d", res.Type)
		}
	default:
		t.Fatalf("expected response in channel")
	}

	err = ctx.Modal(builder.Modal{})
	if err != ErrModalAfterDefer {
		t.Fatalf("expected ErrModalAfterDefer, got %v", err)
	}
}

func TestContextComponentParam(t *testing.T) {
	in := &types.Interaction{
		Data: &types.InteractionData{
			CustomID: "task_done:42",
		},
	}
	ctx := NewContext(context.Background(), in, nil)

	if ctx.ComponentCustomID() != "task_done:42" {
		t.Fatalf("unexpected custom id: %s", ctx.ComponentCustomID())
	}
	if ctx.ComponentParam() != "42" {
		t.Fatalf("expected param 42, got %s", ctx.ComponentParam())
	}
}

func TestContextUserAndMember(t *testing.T) {
	inUser := &types.Interaction{
		User: &types.User{
			ID:       types.Snowflake(123),
			Username: "alice",
		},
	}
	ctx := NewContext(context.Background(), inUser, nil)
	if ctx.User() == nil || ctx.User().Username != "alice" {
		t.Fatalf("expected user alice")
	}

	inMember := &types.Interaction{
		Member: &types.Member{
			User: &types.User{
				ID:       types.Snowflake(456),
				Username: "bob",
			},
		},
	}
	ctx2 := NewContext(context.Background(), inMember, nil)
	if ctx2.User() == nil || ctx2.User().Username != "bob" {
		t.Fatalf("expected member user bob")
	}
}

func TestGatewayContextReply(t *testing.T) {
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		Token: "tok",
	}
	ctx := NewGatewayContext(context.Background(), in, nil)
	if !ctx.gatewayMode {
		t.Fatalf("expected gatewayMode true")
	}
	if ctx.ResponseChannel() != nil {
		t.Fatalf("expected nil ResponseChannel in gateway mode")
	}

	err := ctx.Reply(Response{Content: "gateway reply"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = ctx.Reply(Response{Content: "second reply"})
	if err != ErrAlreadyResponded {
		t.Fatalf("expected ErrAlreadyResponded, got %v", err)
	}
}


