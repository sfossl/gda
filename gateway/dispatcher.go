package gateway

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/sfossl/gda/types"
)

type RawHandler func(ctx context.Context, data []byte)

type EventDispatcher struct {
	mu        sync.RWMutex
	listeners map[types.EventName][]RawHandler
	sem       chan struct{}
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners: make(map[types.EventName][]RawHandler),
		sem:       make(chan struct{}, 1024),
	}
}

func (d *EventDispatcher) SetMaxConcurrency(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if n > 0 {
		d.sem = make(chan struct{}, n)
	}
}

func (d *EventDispatcher) On(event types.EventName, h RawHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners[event] = append(d.listeners[event], h)
}

func (d *EventDispatcher) OnReady(h func(context.Context, *types.ReadyEvent)) {
	d.On(types.EventReady, func(ctx context.Context, data []byte) {
		var ready types.ReadyEvent
		if err := json.Unmarshal(data, &ready); err == nil {
			h(ctx, &ready)
		}
	})
}

func (d *EventDispatcher) OnMessageCreate(h func(context.Context, *types.Message)) {
	d.On(types.EventMessageCreate, func(ctx context.Context, data []byte) {
		var msg types.Message
		if err := json.Unmarshal(data, &msg); err == nil {
			h(ctx, &msg)
		}
	})
}

func (d *EventDispatcher) OnInteractionCreate(h func(context.Context, *types.Interaction)) {
	d.On(types.EventInteractionCreate, func(ctx context.Context, data []byte) {
		var in types.Interaction
		if err := json.Unmarshal(data, &in); err == nil {
			h(ctx, &in)
		}
	})
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event types.EventName, data []byte) {
	d.mu.RLock()
	handlers, ok := d.listeners[event]
	if !ok || len(handlers) == 0 {
		d.mu.RUnlock()
		return
	}
	copied := make([]RawHandler, len(handlers))
	copy(copied, handlers)
	d.mu.RUnlock()

	for _, h := range copied {
		d.mu.RLock()
		sem := d.sem
		d.mu.RUnlock()

		go func(fn RawHandler) {
			if ctx.Err() != nil {
				return
			}
			if sem != nil {
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					return
				}
			}
			if ctx.Err() != nil {
				return
			}
			defer func() {
				_ = recover()
			}()
			fn(ctx, data)
		}(h)
	}
}

