package gateway

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrHeartbeatACKTimeout = errors.New("bfrm/gateway: heartbeat ack timeout (zombie socket)")
)

type HeartbeatManager struct {
	interval    time.Duration
	ackReceived atomic.Bool
	sendFunc    func(*int64) error
	getSeqFunc  func() *int64
	stopOnce    sync.Once
	stopChan    chan struct{}
	running     atomic.Bool
}

func NewHeartbeatManager(interval time.Duration, sendFunc func(*int64) error, getSeqFunc func() *int64) *HeartbeatManager {
	hm := &HeartbeatManager{
		interval:   interval,
		sendFunc:   sendFunc,
		getSeqFunc: getSeqFunc,
		stopChan:   make(chan struct{}),
	}
	hm.ackReceived.Store(true)
	return hm
}

func (h *HeartbeatManager) Acknowledge() {
	h.ackReceived.Store(true)
}

func (h *HeartbeatManager) computeJitter() time.Duration {
	var b [8]byte
	_, _ = rand.Read(b[:])
	n := binary.LittleEndian.Uint64(b[:])
	fraction := float64(n%1000) / 1000.0
	return time.Duration(float64(h.interval) * fraction)
}

func (h *HeartbeatManager) Start(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)

	if !h.running.CompareAndSwap(false, true) {
		close(errChan)
		return errChan
	}

	go func() {
		defer close(errChan)
		defer h.running.Store(false)

		jitter := h.computeJitter()
		initialTimer := time.NewTimer(jitter)

		select {
		case <-initialTimer.C:
		case <-h.stopChan:
			initialTimer.Stop()
			return
		case <-ctx.Done():
			initialTimer.Stop()
			return
		}

		h.ackReceived.Store(false)
		if err := h.sendFunc(h.getSeqFunc()); err != nil {
			errChan <- err
			return
		}

		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !h.ackReceived.Load() {
					errChan <- ErrHeartbeatACKTimeout
					return
				}
				h.ackReceived.Store(false)
				if err := h.sendFunc(h.getSeqFunc()); err != nil {
					errChan <- err
					return
				}

			case <-h.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return errChan
}

func (h *HeartbeatManager) Stop() {
	h.stopOnce.Do(func() {
		close(h.stopChan)
	})
}

