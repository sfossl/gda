package gateway

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestEventDispatcherTyped(t *testing.T) {
	d := NewEventDispatcher()

	var wg sync.WaitGroup
	wg.Add(3)

	var readySession string
	d.OnReady(func(ctx context.Context, e *types.ReadyEvent) {
		readySession = e.SessionID
		wg.Done()
	})

	var msgContent string
	d.OnMessageCreate(func(ctx context.Context, e *types.Message) {
		msgContent = e.Content
		wg.Done()
	})

	var customID string
	d.OnInteractionCreate(func(ctx context.Context, e *types.Interaction) {
		customID = e.Data.CustomID
		wg.Done()
	})

	ctx := context.Background()
	d.Dispatch(ctx, types.EventReady, []byte(`{"session_id":"s_123","user":{"id":"1","username":"bot"}}`))
	d.Dispatch(ctx, types.EventMessageCreate, []byte(`{"id":"2","content":"hello"}`))
	d.Dispatch(ctx, types.EventInteractionCreate, []byte(`{"id":"3","type":3,"data":{"custom_id":"btn_456"}}`))

	wg.Wait()

	if readySession != "s_123" {
		t.Fatalf("expected session s_123, got %s", readySession)
	}
	if msgContent != "hello" {
		t.Fatalf("expected message hello, got %s", msgContent)
	}
	if customID != "btn_456" {
		t.Fatalf("expected custom_id btn_456, got %s", customID)
	}
}

func TestEventDispatcherPanicRecovery(t *testing.T) {
	d := NewEventDispatcher()

	var wg sync.WaitGroup
	wg.Add(1)

	d.On(types.EventMessageCreate, func(ctx context.Context, data []byte) {
		panic("handler panic")
	})

	var executed bool
	d.On(types.EventMessageCreate, func(ctx context.Context, data []byte) {
		executed = true
		wg.Done()
	})

	d.Dispatch(context.Background(), types.EventMessageCreate, []byte(`{}`))
	wg.Wait()

	if !executed {
		t.Fatalf("expected second handler to execute despite panic in first")
	}
}

func TestEventDispatcherContextCancel(t *testing.T) {
	d := NewEventDispatcher()
	d.SetMaxConcurrency(1)

	started := make(chan struct{})
	block := make(chan struct{})
	d.On("EVENT_BLOCK", func(ctx context.Context, data []byte) {
		close(started)
		<-block
	})

	var secondExecuted atomic.Bool
	d.On("EVENT_SECOND", func(ctx context.Context, data []byte) {
		secondExecuted.Store(true)
	})

	d.Dispatch(context.Background(), "EVENT_BLOCK", nil)
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d.Dispatch(ctx, "EVENT_SECOND", nil)
	close(block)
	time.Sleep(20 * time.Millisecond)

	if secondExecuted.Load() {
		t.Fatalf("expected second handler to be aborted due to context cancellation")
	}
}

