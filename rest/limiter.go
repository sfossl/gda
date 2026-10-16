package rest

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.RWMutex
	buckets     map[string]*bucket
	routeToHash map[string]string

	globalMu    sync.RWMutex
	globalReset time.Time
}

type bucket struct {
	mu        sync.Mutex
	remaining int
	limit     int
	resetTime time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets:     make(map[string]*bucket),
		routeToHash: make(map[string]string),
	}
}

func (l *RateLimiter) getBucket(routeKey string) *bucket {
	l.mu.RLock()
	key := routeKey
	if hash, ok := l.routeToHash[routeKey]; ok {
		key = hash
	}
	b, ok := l.buckets[key]
	l.mu.RUnlock()

	if ok {
		return b
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if hash, ok := l.routeToHash[routeKey]; ok {
		key = hash
	}
	if !ok {
		if len(l.buckets) > 64 {
			now := time.Now()
			for k, bkt := range l.buckets {
				if !bkt.resetTime.IsZero() && now.After(bkt.resetTime.Add(time.Minute)) {
					delete(l.buckets, k)
					for rk, h := range l.routeToHash {
						if h == k {
							delete(l.routeToHash, rk)
						}
					}
				}
			}
		}

		b = &bucket{
			remaining: 1,
			limit:     1,
		}
		l.buckets[key] = b
	}
	return b
}

func (l *RateLimiter) Acquire(ctx context.Context, routeKey string) error {
	for {
		l.globalMu.RLock()
		gReset := l.globalReset
		l.globalMu.RUnlock()

		now := time.Now()
		if !gReset.IsZero() && now.Before(gReset) {
			timer := time.NewTimer(gReset.Sub(now))
			select {
			case <-timer.C:
				continue
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}
		break
	}

	b := l.getBucket(routeKey)

	b.mu.Lock()
	defer b.mu.Unlock()

	for {
		now := time.Now()
		if !b.resetTime.IsZero() && now.After(b.resetTime) && b.limit > 0 {
			b.remaining = b.limit
			b.resetTime = time.Time{}
		}

		if b.remaining > 0 {
			b.remaining--
			return nil
		}

		wait := b.resetTime.Sub(now)
		if wait <= 0 {
			wait = 50 * time.Millisecond
		}

		b.mu.Unlock()
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			b.mu.Lock()
			if b.resetTime.IsZero() && b.limit > 0 {
				b.remaining = b.limit
			}
		case <-ctx.Done():
			timer.Stop()
			b.mu.Lock()
			return ctx.Err()
		}
	}
}

func (l *RateLimiter) Update(routeKey string, header http.Header) {
	if header == nil {
		return
	}

	if header.Get("X-RateLimit-Global") == "true" {
		retryAfterStr := header.Get("Retry-After")
		if retryAfter, err := strconv.ParseFloat(retryAfterStr, 64); err == nil {
			l.globalMu.Lock()
			l.globalReset = time.Now().Add(time.Duration(retryAfter * float64(time.Second)))
			l.globalMu.Unlock()
		}
		return
	}

	bucketHash := header.Get("X-RateLimit-Bucket")
	bKey := routeKey
	if bucketHash != "" {
		l.mu.Lock()
		l.routeToHash[routeKey] = bucketHash
		if _, exists := l.buckets[bucketHash]; !exists {
			if existing, ok := l.buckets[routeKey]; ok {
				l.buckets[bucketHash] = existing
			}
		}
		if routeKey != bucketHash {
			delete(l.buckets, routeKey)
		}
		bKey = bucketHash
		l.mu.Unlock()
	}

	b := l.getBucket(bKey)
	b.mu.Lock()
	defer b.mu.Unlock()

	if limStr := header.Get("X-RateLimit-Limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil {
			b.limit = lim
		}
	}

	if remStr := header.Get("X-RateLimit-Remaining"); remStr != "" {
		if rem, err := strconv.Atoi(remStr); err == nil {
			b.remaining = rem
		}
	}

	if resetAfterStr := header.Get("X-RateLimit-Reset-After"); resetAfterStr != "" {
		if resetAfter, err := strconv.ParseFloat(resetAfterStr, 64); err == nil {
			b.resetTime = time.Now().Add(time.Duration(resetAfter * float64(time.Second)))
		}
	} else if retryAfterStr := header.Get("Retry-After"); retryAfterStr != "" {
		if retryAfter, err := strconv.ParseFloat(retryAfterStr, 64); err == nil {
			b.resetTime = time.Now().Add(time.Duration(retryAfter * float64(time.Second)))
		}
	}
}

func (l *RateLimiter) UpdateWithRetryAfter(routeKey string, retryAfter float64, isGlobal bool) {
	reset := time.Now().Add(time.Duration(retryAfter * float64(time.Second)))
	if isGlobal {
		l.globalMu.Lock()
		l.globalReset = reset
		l.globalMu.Unlock()
		return
	}
	b := l.getBucket(routeKey)
	b.mu.Lock()
	b.remaining = 0
	b.resetTime = reset
	b.mu.Unlock()
}

