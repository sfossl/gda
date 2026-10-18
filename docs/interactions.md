# Interactions

The `gda/interaction` package provides an HTTP interaction router and Ed25519 webhook verification engine for handling Discord slash commands, buttons, and modals directly over HTTP without a Gateway connection.

## Overview

Discord sends interactions (application commands, components, and modals) as HTTP `POST` requests to your registered Interactions Endpoint URL. Every incoming request must have its Ed25519 signature verified before execution.

## Server

Set up an interaction webhook server using standard `net/http`:

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sfossl/gda"
	"github.com/sfossl/gda/interaction"
)

func main() {
	client, err := gda.New(
		os.Getenv("TOKEN"),
		gda.WithAppID(os.Getenv("APP_ID")),
		gda.WithPublicKey(os.Getenv("PUBLIC_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}

	client.OnCommand("ping", func(ctx *interaction.Context) error {
		return ctx.Reply(interaction.Response{
			Content: "pong",
		})
	})

	handler, err := client.WebhookHandler()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/interactions", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`client.WebhookHandler()` automatically:
- Verifies Ed25519 signatures on raw request bodies.
- Validates timestamp clock skew to prevent replay attacks.
- Acknowledges Discord `PING` interactions (type 1) with `PONG`.
- Routes commands, subcommands, buttons, and modals.

## Verification

If you are writing a custom HTTP handler without `gda.Client`, use `VerifyMiddleware` directly:

```go
pubKey, err := interaction.ParsePublicKey(os.Getenv("PUBLIC_KEY"))
if err != nil {
	log.Fatal(err)
}

router := interaction.NewRouter()
webhookHandler := router.WebhookHandler(restClient, false)

http.Handle("/interactions", interaction.VerifyMiddleware(pubKey)(webhookHandler))
```

## Router

The interaction router maps command names, subcommand paths, and component custom IDs to handlers:

```go
// top-level command: /help
router.OnCommand("help", func(ctx *interaction.Context) error {
	return ctx.Reply(interaction.Response{Content: "help menu"})
})

// subcommand: /settings role
router.OnSubcommand("settings", "role", func(ctx *interaction.Context) error {
	return ctx.Reply(interaction.Response{Content: "role settings"})
})

// button or select component by custom id
router.OnComponent("btn_confirm", func(ctx *interaction.Context) error {
	return ctx.Reply(interaction.Response{Content: "confirmed!"})
})

// modal submission by custom id
router.OnModal("modal_feedback", func(ctx *interaction.Context) error {
	return ctx.Reply(interaction.Response{Content: "feedback received"})
})
```

## Responses

Inside an interaction handler, respond immediately or defer:

```go
// immediate response
err := ctx.Reply(interaction.Response{
	Content:   "Processed successfully",
	Ephemeral: true,
})

// defer response (gives up to 15 minutes to edit via REST)
err = ctx.Defer(false)
```

Previous: [REST Client](rest.md)  
Next: [Rate Limits](rate-limits.md)

