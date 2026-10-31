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
	if err != ErrNoRestClient {
		t.Fatalf("expected ErrNoRestClient, got %v", err)
	}

	ctx2 := NewGatewayContext(context.Background(), in, nil)
	if err := ctx2.Defer(false); err != ErrNoRestClient {
		t.Fatalf("expected ErrNoRestClient on Defer, got %v", err)
	}

	ctx3 := NewGatewayContext(context.Background(), in, nil)
	if err := ctx3.DeferUpdate(); err != ErrNoRestClient {
		t.Fatalf("expected ErrNoRestClient on DeferUpdate, got %v", err)
	}

	ctx4 := NewGatewayContext(context.Background(), in, nil)
	if err := ctx4.UpdateOriginal(Response{Content: "up"}); err != ErrNoRestClient {
		t.Fatalf("expected ErrNoRestClient on UpdateOriginal, got %v", err)
	}

	ctx5 := NewGatewayContext(context.Background(), in, nil)
	if err := ctx5.Modal(builder.Modal{CustomID: "m", Title: "t"}); err != ErrNoRestClient {
		t.Fatalf("expected ErrNoRestClient on Modal, got %v", err)
	}

	ctx6 := NewGatewayContext(context.Background(), in, nil)
	if err := ctx6.Autocomplete([]types.Choice{{Name: "a", Value: "1"}}); err != ErrNoRestClient {
		t.Fatalf("expected ErrNoRestClient on Autocomplete, got %v", err)
	}
}

func TestContext_FocusedOptionNested(t *testing.T) {
	in := &types.Interaction{
		Data: &types.InteractionData{
			Options: []types.ApplicationCommandOption{
				{
					Name: "sub",
					Type: int(types.OptionTypeSubcommand),
					Options: []types.ApplicationCommandOption{
						{
							Name:    "target",
							Focused: true,
						},
					},
				},
			},
		},
	}
	ctx := NewContext(context.Background(), in, nil)
	f := ctx.FocusedOption()
	if f == nil || f.Name != "target" {
		t.Fatalf("expected focused option target, got %+v", f)
	}
}
