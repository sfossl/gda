package rest

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiterAcquireImmediate(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()

	if err := limiter.Acquire(ctx, "POST:/channels/1/messages"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRateLimiterBucketWait(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()
	route := "POST:/channels/2/messages"

	header := make(http.Header)
	header.Set("X-RateLimit-Limit", "2")
	header.Set("X-RateLimit-Remaining", "0")
	header.Set("X-RateLimit-Reset-After", "0.05")
	limiter.Update(route, header)

	start := time.Now()
	if err := limiter.Acquire(ctx, route); err != nil {
		t.Fatalf("acquire error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Fatalf("expected delay >= 40ms, got %v", elapsed)
	}
}

func TestRateLimiterContextCancel(t *testing.T) {
	limiter := NewRateLimiter()
	route := "POST:/channels/3/messages"

	header := make(http.Header)
	header.Set("X-RateLimit-Remaining", "0")
	header.Set("X-RateLimit-Reset-After", "5.0")
	limiter.Update(route, header)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := limiter.Acquire(ctx, route)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestRateLimiterGlobalLock(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()

	header := make(http.Header)
	header.Set("X-RateLimit-Global", "true")
	header.Set("Retry-After", "0.05")
	limiter.Update("any_route", header)

	start := time.Now()
	if err := limiter.Acquire(ctx, "POST:/other/route"); err != nil {
		t.Fatalf("acquire error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Fatalf("expected global delay >= 40ms, got %v", elapsed)
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()
	route := "POST:/channels/4/messages"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = limiter.Acquire(ctx, route)
		}()
	}
	wg.Wait()
}

func TestRateLimiterUpdateWithRetryAfter(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()
	route := "POST:/channels/5/messages"

	limiter.UpdateWithRetryAfter(route, 0.05, false)

	start := time.Now()
	if err := limiter.Acquire(ctx, route); err != nil {
		t.Fatalf("acquire error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Fatalf("expected delay >= 40ms, got %v", elapsed)
	}

	limiter.UpdateWithRetryAfter("any_route", 0.05, true)
	start = time.Now()
	if err := limiter.Acquire(ctx, "POST:/channels/6/messages"); err != nil {
		t.Fatalf("acquire error: %v", err)
	}
	elapsed = time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Fatalf("expected global delay >= 40ms, got %v", elapsed)
	}
}

func TestRateLimiterNoThunderingHerd(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()
	route := "POST:/channels/7/messages"

	header := make(http.Header)
	header.Set("X-RateLimit-Limit", "1")
	header.Set("X-RateLimit-Remaining", "0")
	header.Set("X-RateLimit-Reset-After", "0.05")
	limiter.Update(route, header)

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)
	var t1, t2 time.Duration
	var firstAcquired atomic.Bool

	go func() {
		defer wg.Done()
		_ = limiter.Acquire(ctx, route)
		t1 = time.Since(start)
		if firstAcquired.CompareAndSwap(false, true) {
			h := make(http.Header)
			h.Set("X-RateLimit-Limit", "1")
			h.Set("X-RateLimit-Remaining", "1")
			limiter.Update(route, h)
		}
	}()

	go func() {
		defer wg.Done()
		_ = limiter.Acquire(ctx, route)
		t2 = time.Since(start)
		if firstAcquired.CompareAndSwap(false, true) {
			h := make(http.Header)
			h.Set("X-RateLimit-Limit", "1")
			h.Set("X-RateLimit-Remaining", "1")
			limiter.Update(route, h)
		}
	}()

	wg.Wait()

	if t1 < 40*time.Millisecond || t2 < 40*time.Millisecond {
		t.Fatalf("expected both to wait >= 40ms, got t1=%v, t2=%v", t1, t2)
	}
}

