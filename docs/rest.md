# REST Client

The `gda/rest` package provides an HTTP/1.1 client for Discord's REST API v10, featuring keep-alive connection pooling, pre-emptive rate-limit bucket scheduling, 429 backoff, and automatic retries for transient 5xx server errors.

## Transport

Every `rest.Client` configures an optimized `http.Transport`:

- `MaxIdleConns: 64`
- `MaxIdleConnsPerHost: 32`
- `IdleConnTimeout: 90 * time.Second`
- `TLSHandshakeTimeout: 10 * time.Second`

This preserves persistent connections across sequential API calls.

## Messages

```go
// send message
msg, err := client.REST.SendMessage(ctx, channelID, rest.MessagePayload{
	Content: "content",
	Embeds:  []types.Embed{embed},
})

// edit message
edited, err := client.REST.EditMessage(ctx, channelID, messageID, rest.MessagePayload{
	Content: "updated",
})

// delete message
err := client.REST.DeleteMessage(ctx, channelID, messageID, "audit log reason")

// fetch message
msg, err := client.REST.GetMessage(ctx, channelID, messageID)
```

## Channels

```go
channel, err := client.REST.GetChannel(ctx, channelID)
err = client.REST.DeleteChannel(ctx, channelID, "audit log reason")

err = client.REST.EditChannelPermissions(ctx, channelID, overwriteID, overwrite)
err = client.REST.DeleteChannelPermission(ctx, channelID, overwriteID)

dm, err := client.REST.CreateDM(ctx, recipientUserID)
err = client.REST.TriggerTyping(ctx, channelID)
```

## Users

```go
botUser, err := client.REST.GetCurrentUser(ctx)
user, err := client.REST.GetUser(ctx, targetUserID)
```

## Members

```go
guild, err := client.REST.GetGuild(ctx, guildID)
member, err := client.REST.GetGuildMember(ctx, guildID, userID)

updated, err := client.REST.ModifyGuildMember(ctx, guildID, userID, rest.ModifyMemberPayload{
	Nick:  &newNick,
	Roles: &[]types.Snowflake{roleID1, roleID2},
})

err := client.REST.AddGuildMemberRole(ctx, guildID, userID, roleID, "reason")
err = client.REST.RemoveGuildMemberRole(ctx, guildID, userID, roleID, "reason")
```

## Moderation

```go
err := client.REST.KickMember(ctx, guildID, userID, "reason")
err = client.REST.BanMember(ctx, guildID, userID, 86400, "reason")
err = client.REST.UnbanMember(ctx, guildID, userID, "reason")

timeoutUntil := time.Now().Add(1 * time.Hour)
err = client.REST.TimeoutMember(ctx, guildID, userID, &timeoutUntil, "reason")
```

## Roles

```go
roles, err := client.REST.GetGuildRoles(ctx, guildID)

newRole, err := client.REST.CreateGuildRole(ctx, guildID, rest.RolePayload{
	Name:  "Moderator",
	Color: 0x3498DB,
})

err = client.REST.DeleteGuildRole(ctx, guildID, roleID)
```

## Interactions

```go
// create response
err := client.REST.CreateInteractionResponse(ctx, interactionID, token, response)

// followup message
followup, err := client.REST.CreateFollowupMessage(ctx, appID, token, payload)

// edit original response
msg, err := client.REST.EditOriginalInteractionResponse(ctx, appID, token, payload)

// delete original response
err = client.REST.DeleteOriginalInteractionResponse(ctx, appID, token)

// bulk overwrite application commands
commands, err := client.REST.BulkOverwriteGlobalCommands(ctx, appID, commandList)
```

Previous: [Gateway WebSocket](gateway.md)  
Next: [Interactions](interactions.md)
