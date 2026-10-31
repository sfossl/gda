# Getting Started

`gda` connects your Go application to Discord's Gateway and REST API v10 using only the Go standard library.

## Install

Requires Go 1.24 or higher:

```bash
go get github.com/sfossl/gda
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sfossl/gda"
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func main() {
	token := os.Getenv("TOKEN")
	appID := os.Getenv("APP_ID")

	client, err := gda.New(
		token,
		gda.WithAppID(appID),
		gda.WithIntents(types.IntentsDefault|types.IntentMessageContent),
		gda.WithHTTPTimeout(15*time.Second),
		gda.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatalf("failed to initialize gda client: %v", err)
	}

	client.OnReady(func(ctx context.Context, e *types.ReadyEvent) {
		fmt.Printf("ready: %s (%s) guilds: %d\n", e.User.Username, e.User.ID, len(e.Guilds))
	})

	client.OnMessageCreate(func(ctx context.Context, m *types.Message) {
		if m.Author.Bot || m.Content != "hi" {
			return
		}

		gwLatency := client.Latency().Round(time.Millisecond)
		_, _ = client.REST.SendMessage(ctx, m.ChannelID, rest.MessagePayload{
			Content: fmt.Sprintf("hello from gda! (latency: %s)", gwLatency),
		})
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := client.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("runtime error: %v", err)
	}

	_ = client.Close()
}
```

## Options

| Option | Signature | Default | Description |
| :--- | :--- | :--- | :--- |
| `WithAppID` | `WithAppID(string)` | `""` | Sets Discord application ID for interaction endpoints. |
| `WithIntents` | `WithIntents(types.Intent)` | `IntentsDefault` | Bitwise gateway event subscription mask. |
| `WithPublicKey` | `WithPublicKey(string)` | `""` | Hex-encoded Ed25519 public key for webhook validation. |
| `WithShard` | `WithShard(id, count int)` | `0, 1` | Sets zero-indexed shard ID and total shard count. |
| `WithHTTPTimeout`| `WithHTTPTimeout(time.Duration)`| `10s` | Outbound HTTP request timeout for REST calls. |
| `WithMaxRetries` | `WithMaxRetries(int)` | `3` | Maximum automatic retries on HTTP 5xx errors. |
| `WithLargeThreshold` | `WithLargeThreshold(int)` | `50` | Member threshold (50 to 250) for offline member filtering. |
| `WithCompression` | `WithCompression(bool)` | `false` | Enables zlib stream compression on Gateway WebSocket. |
| `WithAutoDefer` | `WithAutoDefer(bool)` | `false` | Auto-defers incoming gateway interactions at 2.5 seconds. |

## Architecture

`gda.Client` coordinates four primary subsystems:

- **`client.REST`**: Access Discord REST API endpoints (channels, guilds, members, roles, messages).
- **`client.Gateway`**: Manages the persistent WebSocket connection, zlib decompression, heartbeat loop, and reconnection state.
- **`client.Dispatcher`**: Delivers incoming Gateway payloads concurrently to registered event listeners.
- **`client.Router`**: Interaction router for slash commands, buttons, and modals.

## Lifecycle

`client.Start(ctx)` opens the Gateway WebSocket, executes the identify or resume handshake, and blocks until the context is canceled or a fatal connection error occurs:

```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

if err := client.Start(ctx); err != nil && err != context.Canceled {
	log.Printf("gateway closed with error: %v", err)
}
```

Call `client.Close()` on shutdown to terminate the WebSocket cleanly with code `1000` and close idle HTTP connections:

```go
defer client.Close()
```

## Latency

`gda` measures Gateway WebSocket heartbeat round-trip ping automatically:

```go
latency := client.Latency()
fmt.Printf("gateway ping: %s\n", latency.Round(time.Millisecond))
```

Next: [Gateway WebSocket](gateway.md)
