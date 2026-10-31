# API Reference

Complete reference of exported packages, types, constructors, and functions in `github.com/sfossl/gda`.

## Client

### Constructors
```go
func New(token string, opts ...Option) (*Client, error)
func NewWithConfig(cfg Config) (*Client, error)
func DefaultConfig(token string) Config
```

### Functional Options
```go
func WithAppID(appID string) Option
func WithPublicKey(publicKey string) Option
func WithIntents(intents types.Intent) Option
func WithShard(shardID, shardCount int) Option
func WithHTTPTimeout(d time.Duration) Option
func WithMaxRetries(retries int) Option
func WithLargeThreshold(threshold int) Option
func WithAutoDefer(autoDefer bool) Option
func WithCompression(compress bool) Option
```

### Methods
```go
func (c *Client) Start(ctx context.Context) error
func (c *Client) Close() error
func (c *Client) Latency() time.Duration
func (c *Client) OnReady(h func(context.Context, *types.ReadyEvent))
func (c *Client) OnMessageCreate(h func(context.Context, *types.Message))
func (c *Client) On(event types.EventName, h gateway.RawHandler)
func (c *Client) OnCommand(name string, h interaction.Handler)
func (c *Client) OnSubcommand(command, subcommand string, h interaction.Handler)
func (c *Client) OnSubcommandGroup(command, group, subcommand string, h interaction.Handler)
func (c *Client) OnComponent(customID string, h interaction.Handler)
func (c *Client) OnModal(customID string, h interaction.Handler)
func (c *Client) UpdatePresence(ctx context.Context, presence types.PresenceUpdate) error
func (c *Client) TriggerTyping(ctx context.Context, channelID types.Snowflake) error
func (c *Client) ValidateEventIntents(event types.EventName) error
func (c *Client) WebhookHandler() (http.Handler, error)
func (c *Client) String() string
```

## Gateway

### `gateway.Client`
```go
func NewClient(cfg Config, dispatcher *EventDispatcher) *Client
func (c *Client) Connect(ctx context.Context) error
func (c *Client) Close() error
func (c *Client) Latency() time.Duration
func (c *Client) State() *SessionState
func (c *Client) Dispatcher() *EventDispatcher
func (c *Client) UpdatePresence(ctx context.Context, presence types.PresenceUpdate) error
func (c *Client) String() string
```

### `gateway.EventDispatcher`
```go
func NewEventDispatcher(concurrency ...int) *EventDispatcher
func (d *EventDispatcher) On(event types.EventName, h RawHandler)
func (d *EventDispatcher) OnReady(h func(context.Context, *types.ReadyEvent))
func (d *EventDispatcher) OnMessageCreate(h func(context.Context, *types.Message))
func (d *EventDispatcher) OnInteractionCreate(h func(context.Context, *types.Interaction))
func (d *EventDispatcher) Dispatch(ctx context.Context, event types.EventName, data []byte)
```

## REST

### `rest.Client` Methods
```go
// execution
func (c *Client) Do(ctx context.Context, req Request) ([]byte, error)
func (c *Client) CloseIdleConnections()

// messages & channels
func (c *Client) SendMessage(ctx context.Context, channelID types.Snowflake, payload MessagePayload) (*types.Message, error)
func (c *Client) EditMessage(ctx context.Context, channelID, messageID types.Snowflake, payload MessagePayload) (*types.Message, error)
func (c *Client) DeleteMessage(ctx context.Context, channelID, messageID types.Snowflake, reason ...string) error
func (c *Client) GetMessage(ctx context.Context, channelID, messageID types.Snowflake) (*types.Message, error)
func (c *Client) GetChannel(ctx context.Context, channelID types.Snowflake) (*types.Channel, error)
func (c *Client) CreateGuildChannel(ctx context.Context, guildID types.Snowflake, payload CreateChannelPayload, reason ...string) (*types.Channel, error)
func (c *Client) ModifyChannel(ctx context.Context, channelID types.Snowflake, payload ModifyChannelPayload, reason ...string) (*types.Channel, error)
func (c *Client) DeleteChannel(ctx context.Context, channelID types.Snowflake, reason ...string) error
func (c *Client) CreateDM(ctx context.Context, recipientID types.Snowflake) (*types.Channel, error)
func (c *Client) EditChannelPermissions(ctx context.Context, channelID, overwriteID types.Snowflake, overwrite types.PermissionOverwrite) error
func (c *Client) DeleteChannelPermission(ctx context.Context, channelID, overwriteID types.Snowflake) error
func (c *Client) TriggerTyping(ctx context.Context, channelID types.Snowflake) error

// users
func (c *Client) GetCurrentUser(ctx context.Context) (*types.User, error)
func (c *Client) GetUser(ctx context.Context, userID types.Snowflake) (*types.User, error)

// guilds & members
func (c *Client) GetGuild(ctx context.Context, guildID types.Snowflake) (*types.Guild, error)
func (c *Client) GetGuildMember(ctx context.Context, guildID, userID types.Snowflake) (*types.Member, error)
func (c *Client) ModifyGuildMember(ctx context.Context, guildID, userID types.Snowflake, payload ModifyMemberPayload) (*types.Member, error)
func (c *Client) AddGuildMemberRole(ctx context.Context, guildID, userID, roleID types.Snowflake, reason ...string) error
func (c *Client) RemoveGuildMemberRole(ctx context.Context, guildID, userID, roleID types.Snowflake, reason ...string) error
func (c *Client) KickMember(ctx context.Context, guildID, userID types.Snowflake, reason string) error
func (c *Client) BanMember(ctx context.Context, guildID, userID types.Snowflake, deleteMessageSeconds int, reason string) error
func (c *Client) UnbanMember(ctx context.Context, guildID, userID types.Snowflake, reason string) error
func (c *Client) TimeoutMember(ctx context.Context, guildID, userID types.Snowflake, until *time.Time, reason string) error

// roles
func (c *Client) GetGuildRoles(ctx context.Context, guildID types.Snowflake) ([]types.Role, error)
func (c *Client) CreateGuildRole(ctx context.Context, guildID types.Snowflake, payload RolePayload, reason ...string) (*types.Role, error)
func (c *Client) ModifyGuildRole(ctx context.Context, guildID, roleID types.Snowflake, payload RolePayload, reason ...string) (*types.Role, error)
func (c *Client) DeleteGuildRole(ctx context.Context, guildID, roleID types.Snowflake, reason ...string) error

// reactions & pins
func (c *Client) CreateReaction(ctx context.Context, channelID, messageID types.Snowflake, emoji string) error
func (c *Client) DeleteOwnReaction(ctx context.Context, channelID, messageID types.Snowflake, emoji string) error
func (c *Client) DeleteUserReaction(ctx context.Context, channelID, messageID types.Snowflake, emoji string, userID types.Snowflake) error
func (c *Client) PinMessage(ctx context.Context, channelID, messageID types.Snowflake, reason ...string) error
func (c *Client) UnpinMessage(ctx context.Context, channelID, messageID types.Snowflake, reason ...string) error

// interaction responses & followups
func (c *Client) CreateInteractionResponse(ctx context.Context, interactionID types.Snowflake, token string, response types.InteractionResponse) error
func (c *Client) CreateFollowupMessage(ctx context.Context, appID types.Snowflake, token string, payload MessagePayload) (*types.Message, error)
func (c *Client) EditOriginalInteractionResponse(ctx context.Context, appID types.Snowflake, token string, payload MessagePayload) (*types.Message, error)
func (c *Client) DeleteOriginalInteractionResponse(ctx context.Context, appID types.Snowflake, token string) error

// application commands sync
func (c *Client) BulkOverwriteGlobalCommands(ctx context.Context, appID types.Snowflake, commands []types.ApplicationCommand) ([]types.ApplicationCommand, error)
func (c *Client) BulkOverwriteGuildCommands(ctx context.Context, appID, guildID types.Snowflake, commands []types.ApplicationCommand) ([]types.ApplicationCommand, error)
```

## Interactions

### `interaction.Router`
```go
func NewRouter() *Router
func (r *Router) Use(mw ...Middleware)
func (r *Router) OnCommand(name string, h Handler)
func (r *Router) OnSubcommand(command, subcommand string, h Handler)
func (r *Router) OnSubcommandGroup(command, group, subcommand string, h Handler)
func (r *Router) OnComponent(customID string, h Handler)
func (r *Router) OnModal(customID string, h Handler)
func (r *Router) SetDefault(h Handler)
func (r *Router) Handle(ctx *Context) error
func (r *Router) WebhookHandler(restClient *rest.Client, autoDefer ...bool) http.Handler
```

### `interaction.Context`
```go
func NewContext(ctx context.Context, in *types.Interaction, restClient *rest.Client) *Context
func NewGatewayContext(ctx context.Context, in *types.Interaction, restClient *rest.Client) *Context
func (c *Context) Context() context.Context
func (c *Context) Reply(res Response) error
func (c *Context) Defer(ephemeral bool) error
func (c *Context) DeferUpdate() error
func (c *Context) UpdateOriginal(res Response) error
func (c *Context) Modal(m builder.Modal) error
func (c *Context) Autocomplete(choices []types.Choice) error
func (c *Context) FocusedOption() *types.ApplicationCommandOption
func (c *Context) Followup(res Response) (*types.Message, error)
func (c *Context) ResponseChannel() <-chan types.InteractionResponse
func (c *Context) CommandName() string
func (c *Context) ComponentCustomID() string
func (c *Context) ComponentParam() string
func (c *Context) User() *types.User
func (c *Context) Member() *types.Member
func (c *Context) ChannelID() types.Snowflake
func (c *Context) GuildID() *types.Snowflake
```

### Signature Verification
```go
func ParsePublicKey(publicKeyHex string) (ed25519.PublicKey, error)
func VerifySignature(publicKey ed25519.PublicKey, signatureHex, timestamp string, body []byte) bool
func VerifyMiddleware(publicKey ed25519.PublicKey) func(http.Handler) http.Handler
```

## Rate Limiter

```go
func NewRateLimiter() *RateLimiter
func (l *RateLimiter) Acquire(ctx context.Context, routeKey string) error
func (l *RateLimiter) Update(routeKey string, header http.Header)
func (l *RateLimiter) UpdateWithRetryAfter(routeKey string, retryAfter float64, isGlobal bool)
```

## Builders

### `EmbedBuilder`
```go
func NewEmbed() *EmbedBuilder
func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder
func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder
func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder
func (b *EmbedBuilder) SetColor(color int) *EmbedBuilder
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder
func (b *EmbedBuilder) SetFooter(text string, iconURL ...string) *EmbedBuilder
func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder
func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder
func (b *EmbedBuilder) SetAuthor(name string, urls ...string) *EmbedBuilder
func (b *EmbedBuilder) SetTimestamp(t time.Time) *EmbedBuilder
func (b *EmbedBuilder) Build() (types.Embed, error)
```

### `ButtonBuilder` and `ActionRow`
```go
func NewButton() *ButtonBuilder
func (b *ButtonBuilder) SetStyle(style ButtonStyle) *ButtonBuilder
func (b *ButtonBuilder) SetCustomID(customID string) *ButtonBuilder
func (b *ButtonBuilder) SetLabel(label string) *ButtonBuilder
func (b *ButtonBuilder) SetURL(url string) *ButtonBuilder
func (b *ButtonBuilder) SetDisabled(disabled bool) *ButtonBuilder
func (b *ButtonBuilder) SetEmoji(name string, id *types.Snowflake, animated bool) *ButtonBuilder
func (b *ButtonBuilder) Build() (types.Component, error)

func NewActionRow(components ...types.Component) (types.Component, error)
```

### `SelectMenuBuilder`
```go
func NewStringSelect(customID string) *SelectMenuBuilder
func NewUserSelect(customID string) *SelectMenuBuilder
func NewRoleSelect(customID string) *SelectMenuBuilder
func NewChannelSelect(customID string) *SelectMenuBuilder
func NewMentionableSelect(customID string) *SelectMenuBuilder
func (b *SelectMenuBuilder) SetPlaceholder(placeholder string) *SelectMenuBuilder
func (b *SelectMenuBuilder) SetValues(min, max int) *SelectMenuBuilder
func (b *SelectMenuBuilder) SetDisabled(disabled bool) *SelectMenuBuilder
func (b *SelectMenuBuilder) AddOption(label, value, description string, isDefault bool) *SelectMenuBuilder
func (b *SelectMenuBuilder) Build() (types.Component, error)
```

### `ModalBuilder` and `TextInputBuilder`
```go
func NewTextInput(customID, label string, style TextInputStyle) *TextInputBuilder
func (b *TextInputBuilder) SetPlaceholder(placeholder string) *TextInputBuilder
func (b *TextInputBuilder) SetValue(value string) *TextInputBuilder
func (b *TextInputBuilder) SetLength(min, max int) *TextInputBuilder
func (b *TextInputBuilder) Build() (types.Component, error)

func NewModal(customID, title string) *ModalBuilder
func (b *ModalBuilder) AddTextInput(input *TextInputBuilder) *ModalBuilder
func (b *ModalBuilder) Build() (Modal, error)
```

## Errors

```go
func IsNotFound(err error) bool
func IsRateLimited(err error) bool
func IsUnauthorized(err error) bool
func IsForbidden(err error) bool
func HasErrorCode(err error, code int) bool

// Sentinel errors
var ErrMissingToken error
var ErrInvalidShard error
var ErrMissingMessageIntent error
var ErrMissingPublicKey error
var ErrPayloadTooLarge error
var ErrAlreadyResponded error
var ErrModalAfterDefer error
var ErrNoRestClient error
```

Previous: [Types & Models](types.md)
