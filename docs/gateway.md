# Gateway WebSocket

The `gda/gateway` package implements a persistent WebSocket client for Discord's Gateway v10 protocol, managing connection lifecycles, zlib decompression, heartbeat loops, resumes, and concurrent event dispatching.

## Lifecycle

```
[Disconnected]
      │
      ▼
Dial WebSocket (wss://gateway.discord.gg/?v=10&encoding=json)
      │
      ▼
Receive Opcode 10 (HELLO) ───> Start Heartbeat Ticker
      │
Is existing session?
├── Yes ──> Send Opcode 6 (RESUME) ───> Receive Opcode 0 (RESUMED)
└── No  ──> Send Opcode 2 (IDENTIFY) ──> Receive Opcode 0 (READY)
      │
      ▼
[Connected & Streaming Events]
```

- `client.Start(ctx)` initiates the dial and runs the read loop.
- If the WebSocket drops with a recoverable code (e.g. 4000, 4008 or network drops), the client reconnects and resumes the session using the cached `session_id` and sequence counter.
- Non-recoverable codes (e.g. 4004 Authentication Failed, 4013/4014 Invalid Intents) terminate immediately and return an error.

## Heartbeat

The Gateway maintains a heartbeat manager that verifies connection health and measures network ping:

- **Interval**: Derived from `heartbeat_interval` sent by Discord in the `HELLO` payload.
- **Jitter**: Initial heartbeat applies random jitter to prevent thundering herd spikes on Discord's gateway servers.
- **ACK Tracking**: Sends Opcode 1 (`HEARTBEAT`). Discord replies with Opcode 11 (`HEARTBEAT_ACK`).
- **Round-Trip Ping**: Duration between sending heartbeat and receiving ACK is tracked atomically, available via `client.Latency()`.
- **Zombie Detection**: If an outgoing heartbeat fires but the previous ACK was never received, the connection is treated as dead and reconnects.

## Events

Incoming Gateway events (Opcode 0 `DISPATCH`) are handled via `gateway.EventDispatcher`.

### Typed Handlers

```go
client.OnReady(func(ctx context.Context, e *types.ReadyEvent) {
	fmt.Printf("ready: %s (%s)\n", e.User.Username, e.SessionID)
})

client.OnMessageCreate(func(ctx context.Context, m *types.Message) {
	if m.Author.Bot {
		return
	}
	fmt.Printf("message: %s: %s\n", m.Author.Username, m.Content)
})
```

### Raw Handlers

For Discord events without a dedicated helper, register a raw handler:

```go
client.On(types.EventGuildCreate, func(ctx context.Context, data []byte) {
	var guild types.Guild
	if err := json.Unmarshal(data, &guild); err == nil {
		fmt.Printf("guild: %s (%s)\n", guild.Name, guild.ID)
	}
})
```

Each event is dispatched in its own goroutine bounded by a concurrency semaphore, preventing slow event handlers from blocking the WebSocket read loop.

## Presence

Update bot status and activity over the active WebSocket:

```go
idleSince := int(time.Now().UnixMilli())

err := client.UpdatePresence(ctx, types.PresenceUpdate{
	Since:  &idleSince,
	Status: types.StatusOnline,
	AFK:    false,
	Activities: []types.Activity{
		{
			Name: "status",
			Type: types.ActivityTypeGame,
		},
	},
})
```

Previous: [Getting Started](getting-started.md)  
Next: [REST Client](rest.md)
