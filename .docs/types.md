# Types & Models

The `gda/types` package provides Go definitions for Discord API v10 primitives, bitmasks, and payload data structures.

## Snowflake

Discord uses 64-bit unsigned integers (`Snowflake`) for unique IDs. In JSON payloads, Discord serializes snowflakes as strings to avoid numeric precision loss.

```go
type Snowflake uint64
```

### Methods
- `s.String()`: Returns string representation (`"1386196193981497364"`).
- `s.Time()`: Extracts creation timestamp from the top 42 bits:
  ```go
  created := messageID.Time()
  fmt.Printf("message time: %s\n", created.Format(time.RFC3339))
  ```
- `s.IsValid()`: Returns `true` if `s > 0`.
- `types.ParseSnowflake(string)`: Parses string representation into a `types.Snowflake`.

## Intents

Bitwise flags passed to `gda.WithIntents(...)` declaring which events your bot intends to receive over the Gateway:

```go
const (
	IntentGuilds                      Intent = 1 << 0
	IntentGuildMembers                Intent = 1 << 1  // privileged
	IntentGuildModeration             Intent = 1 << 2
	IntentGuildEmojisAndStickers      Intent = 1 << 3
	IntentGuildIntegrations           Intent = 1 << 4
	IntentGuildWebhooks               Intent = 1 << 5
	IntentGuildInvites                Intent = 1 << 6
	IntentGuildVoiceStates            Intent = 1 << 7
	IntentGuildPresences              Intent = 1 << 8  // privileged
	IntentGuildMessages               Intent = 1 << 9
	IntentGuildMessageReactions       Intent = 1 << 10
	IntentGuildMessageTyping          Intent = 1 << 11
	IntentDirectMessages              Intent = 1 << 12
	IntentDirectMessageReactions      Intent = 1 << 13
	IntentDirectMessageTyping         Intent = 1 << 14
	IntentMessageContent              Intent = 1 << 15 // privileged
	IntentGuildScheduledEvents        Intent = 1 << 16
	IntentAutoModerationConfiguration Intent = 1 << 20
	IntentAutoModerationExecution     Intent = 1 << 21
	IntentGuildMessagePolls           Intent = 1 << 24
	IntentDirectMessagePolls          Intent = 1 << 25
)
```

Pre-defined bundles:
- `types.IntentsDefault`: Guilds, Messages, and Direct Messages.
- `types.IntentsNonPrivileged`: All non-privileged intents combined.

Bitmask methods:
```go
intents := types.IntentsDefault.Add(types.IntentMessageContent)
if intents.Has(types.IntentMessageContent) {
	// message content enabled
}
intents = intents.Remove(types.IntentDirectMessages)
```

## Permissions

Bitwise flags representing Discord member, role, and channel permissions:

```go
const (
	PermissionCreateInstantInvite  Permission = 1 << 0
	PermissionKickMembers          Permission = 1 << 1
	PermissionBanMembers           Permission = 1 << 2
	PermissionAdministrator        Permission = 1 << 3
	PermissionManageChannels       Permission = 1 << 4
	PermissionManageGuild          Permission = 1 << 5
	PermissionAddReactions         Permission = 1 << 6
	PermissionViewAuditLog         Permission = 1 << 7
	PermissionViewChannel          Permission = 1 << 10
	PermissionSendMessages         Permission = 1 << 11
	PermissionManageMessages       Permission = 1 << 13
	PermissionEmbedLinks           Permission = 1 << 14
	PermissionAttachFiles          Permission = 1 << 15
	PermissionReadMessageHistory   Permission = 1 << 16
	PermissionMentionEveryone      Permission = 1 << 17
	PermissionManageRoles          Permission = 1 << 28
	PermissionModerateMembers      Permission = 1 << 40
)
```

Calling `.Has(...)` automatically evaluates to `true` if `PermissionAdministrator` is present in the bitmask.

## Models

Primary Discord data models in `gda/types`:

- `types.Message`: Full message object (author, content, channel, embeds, components).
- `types.Channel`: Guild text, voice, category, or DM channels.
- `types.Guild`: Discord guild with member count, roles, and settings.
- `types.Member`: Guild member with nickname, roles, joined timestamp, and permissions.
- `types.User`: Discord user account (username, discriminator, avatar, bot flag).
- `types.Role`: Guild role with color, hoist, position, and permission bitmask.
- `types.Interaction`: Inbound slash command, component click, or modal submission payload.

Previous: [Builders](builders.md)  
Next: [API Reference](api-reference.md)
