# Rate Limits

The `gda/rest` package includes an internal rate limiter (`RateLimiter`) designed for Discord's header-based rate limiting model, ensuring concurrent requests do not trigger HTTP 429 limits or IP blocks.

## Tiers

Discord enforces rate limits across two tiers:

- **Per-Route Buckets**: Individual limits per endpoint path and major parameter (e.g. sending messages to channel A has a different bucket than channel B).
- **Global Limits**: A hard limit across the entire bot token (typically 50 requests per second across all routes).

## Hashing

Discord returns dynamic bucket hashes in HTTP response headers:

- `X-RateLimit-Bucket`: Unique hash representing the internal rate-limit group.
- `X-RateLimit-Remaining`: Requests left before exhausting the bucket.
- `X-RateLimit-Reset-After`: Seconds until the bucket resets.
- `X-RateLimit-Global`: Set to `true` if the application hit the global ceiling.

Because Discord's internal bucket hashes are initially unknown, `gda` utilizes a two-tier registry:

1. Initial requests route by normalized route path (e.g. `POST:/channels/123/messages`).
2. When Discord returns `X-RateLimit-Bucket`, `gda` maps that route path to the returned bucket hash.
3. Subsequent requests sharing that hash synchronize against the same bucket mutex and token counter.
4. Idle buckets inactive for over 5 minutes are automatically cleaned up during throttled periodic sweeps.

## Acquire

Before sending any HTTP request, `rest.Client.Do` synchronizes against the internal rate limiter:

```go
err := limiter.Acquire(ctx, req.Endpoint.Key)
```

1. **Global Lock Check**: If a global 429 was encountered, `Acquire` pauses until `globalReset` has passed.
2. **Bucket Lock Check**: The bucket mutex is acquired. If `remaining <= 0` and the current time is before `resetTime`, the caller pauses until the reset timestamp.
3. **Optimistic Decrement**: Once the wait completes, `remaining` is decremented by 1 and the request proceeds.
4. **Context Cancellation**: If `ctx.Done()` fires while waiting, the pause aborts immediately and returns `ctx.Err()`.

## Backoff

If a request still returns HTTP 429:

- Response headers are parsed for `Retry-After` and `X-RateLimit-Global`.
- The bucket (or global tracker) is updated with the backoff period.
- `gda` automatically pauses and retries the request up to `Config.MaxRetries` times without bubbling an error to your application.

Previous: [Interactions](interactions.md)  
Next: [Builders](builders.md)
