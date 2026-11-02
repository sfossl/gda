# Changelog

## v0.1.1 (2026-11-02)

- Add user and guild member avatar URL helpers (`AvatarURL`, `AvatarOptions`).
- Add animated avatar format detection (`.gif`) and default avatar index calculation.
- Add `GuildID` field to `types.Member`.

## v0.1.0 (2026-10-31)

- Initial stable release.
- Zero-dependency Discord API v10 driver using Go 1.24 standard library.
- Pre-flight route-bucketed rate limiter.
- Gateway WebSocket engine with heartbeat watchdog, session resume and zlib stream decompression.
- Unified interaction router for Gateway events and HTTPS webhooks with Ed25519 verification.
- REST client with multipart form data file upload support.
- Bounds-checked fluent builders for embeds, components and modals.
- Discord v10 data models, snowflake utilities, bitflags and wire schemas.

