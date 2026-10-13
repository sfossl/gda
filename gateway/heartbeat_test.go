package gateway

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestHeartbeatNormal(t *testing.T) {
	var count int32
	sendFunc := func(seq *int64) error {
		atomic.AddInt32(&count, 1)
		return nil
	}
	getSeq := func() *int64 {
		s := int64(1)
		return &s
	}

	hm := NewHeartbeatManager(20*time.Millisecond, sendFunc, getSeq)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := hm.Start(ctx)

	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(15 * time.Millisecond)
			hm.Acknowledge()
		}
	}()

	time.Sleep(80 * time.Millisecond)
	hm.Stop()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
	}

	if atomic.LoadInt32(&count) < 2 {
		t.Fatalf("expected at least 2 heartbeats, got %d", atomic.LoadInt32(&count))
	}
}

func TestHeartbeatACKTimeout(t *testing.T) {
	sendFunc := func(seq *int64) error {
		return nil
	}
	getSeq := func() *int64 {
		return nil
	}

	hm := NewHeartbeatManager(15*time.Millisecond, sendFunc, getSeq)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := hm.Start(ctx)

	select {
	case err := <-errChan:
		if err != ErrHeartbeatACKTimeout {
			t.Fatalf("expected ErrHeartbeatACKTimeout, got %v", err)
		}
	case <-time.After(150 * time.Millisecond):
		t.Fatalf("heartbeat watchdog did not trigger within deadline")
	}
}

