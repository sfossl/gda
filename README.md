# Go Discord API

## Overview

`gda` is a low-level Go driver for the Discord API v10 and Gateway WebSocket protocol, implemented using only the Go standard library. It provides direct access to event streams, REST endpoints, and interaction webhooks without forcing an opinionated application architecture.

## Ecosystem

`gda` is designed to be the protocol foundation:

- **`gda`**: Direct access to Gateway events, REST routes, and interaction handlers. You control socket lifecycles, HTTP servers, and execution flows directly.
- **[`gdk`](https://github.com/sfossl/gdk)**: A higher-level interaction framework built on top of `gda`. Choose `gdk` if you want automated slash command registration, cooldowns, role permission guards, and auto-deferrals.

## Install

```bash
go get github.com/sfossl/gda
```

## Example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sfossl/gda"
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func main() {
	client, err := gda.New(
		os.Getenv("TOKEN"),
		gda.WithIntents(types.IntentGuildMessages|types.IntentMessageContent),
	)
	if err != nil {
		log.Fatal(err)
	}

	client.OnReady(func(ctx context.Context, r *types.ReadyEvent) {
		fmt.Printf("Logged in as %s\n", r.User.Username)
	})

	client.OnMessageCreate(func(ctx context.Context, msg *types.Message) {
		if msg.Author.Bot {
			return
		}
		if msg.Content == "hi" {
			_, _ = client.REST.SendMessage(ctx, msg.ChannelID, rest.MessagePayload{
				Content: "hello from gda!",
			})
		}
	})

	if err := client.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

## Design

- **State Storage**: `gda` does not maintain global caches for guilds, channels, or members. State requirements vary drastically between a single-process bot and a multi-shard cluster, so caching strategy is left entirely to the application layer.
- **Audio**: Discord voice streaming requires Opus encoding and media transcoding. Rather than introducing CGO or external FFmpeg dependencies, `gda` provides voice protocol types and leaves media processing to dedicated audio tooling.
- **Dependencies**: The entire driver relies solely on standard library packages (`net/http`, `crypto/ed25519`, `compress/zlib`), ensuring fast builds, zero dependency drift, and simple cross-compilation.

## Docs

- [Getting Started](docs/getting-started.md)
- [Gateway WebSocket](docs/gateway.md)
- [REST Client](docs/rest.md)
- [Interactions](docs/interactions.md)
- [Rate Limits](docs/rate-limits.md)
- [Builders](docs/builders.md)
- [Types & Models](docs/types.md)
- [API Reference](docs/api-reference.md)
