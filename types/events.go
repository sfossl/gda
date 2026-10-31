package types

import (
	"encoding/json"
	"time"
)

type EventName string

const (
	EventReady                       EventName = "READY"
	EventResumed                     EventName = "RESUMED"
	EventMessageCreate               EventName = "MESSAGE_CREATE"
	EventMessageUpdate               EventName = "MESSAGE_UPDATE"
	EventMessageDelete               EventName = "MESSAGE_DELETE"
	EventInteractionCreate           EventName = "INTERACTION_CREATE"
	EventGuildCreate                 EventName = "GUILD_CREATE"
	EventGuildUpdate                 EventName = "GUILD_UPDATE"
	EventGuildDelete                 EventName = "GUILD_DELETE"
	EventChannelCreate               EventName = "CHANNEL_CREATE"
	EventChannelUpdate               EventName = "CHANNEL_UPDATE"
	EventChannelDelete               EventName = "CHANNEL_DELETE"
	EventVoiceStateUpdate            EventName = "VOICE_STATE_UPDATE"
	EventAutoModerationRuleCreate    EventName = "AUTO_MODERATION_RULE_CREATE"
	EventAutoModerationRuleUpdate    EventName = "AUTO_MODERATION_RULE_UPDATE"
	EventAutoModerationRuleDelete    EventName = "AUTO_MODERATION_RULE_DELETE"
	EventAutoModerationActionExecute EventName = "AUTO_MODERATION_ACTION_EXECUTION"
)

type GatewayPayload struct {
	Op GatewayOpcode   `json:"op"`
	D  json.RawMessage `json:"d,omitempty"`
	S  *int64          `json:"s,omitempty"`
	T  *EventName      `json:"t,omitempty"`
}

type IdentifyPayload struct {
	Token          string               `json:"token"`
	Properties     ConnectionProperties `json:"properties"`
	Compress       bool                 `json:"compress,omitempty"`
	LargeThreshold int                  `json:"large_threshold,omitempty"`
	Shard          *[2]int              `json:"shard,omitempty"`
	Intents        Intent               `json:"intents"`
}

type ConnectionProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

type ResumePayload struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int64  `json:"seq"`
}

type HelloPayload struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type ReadyEvent struct {
	V                int                `json:"v"`
	User             User               `json:"user"`
	Guilds           []UnavailableGuild `json:"guilds"`
	SessionID        string             `json:"session_id"`
	ResumeGatewayURL string             `json:"resume_gateway_url"`
	Shard            *[2]int            `json:"shard,omitempty"`
}

type UnavailableGuild struct {
	ID          Snowflake `json:"id"`
	Unavailable bool      `json:"unavailable"`
}

type User struct {
	ID            Snowflake `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	GlobalName    *string   `json:"global_name,omitempty"`
	Avatar        *string   `json:"avatar,omitempty"`
	Bot           bool      `json:"bot,omitempty"`
	System        bool      `json:"system,omitempty"`
}

type Member struct {
	User         *User       `json:"user,omitempty"`
	Nick         *string     `json:"nick,omitempty"`
	Avatar       *string     `json:"avatar,omitempty"`
	Roles        []Snowflake `json:"roles"`
	JoinedAt     time.Time   `json:"joined_at"`
	PremiumSince *time.Time  `json:"premium_since,omitempty"`
	Deaf         bool        `json:"deaf"`
	Mute         bool        `json:"mute"`
	Pending      bool        `json:"pending,omitempty"`
	Permissions  Permission  `json:"permissions,omitempty"`
}

type Channel struct {
	ID               Snowflake  `json:"id"`
	Type             int        `json:"type"`
	GuildID          *Snowflake `json:"guild_id,omitempty"`
	Position         *int       `json:"position,omitempty"`
	Name             *string    `json:"name,omitempty"`
	Topic            *string    `json:"topic,omitempty"`
	NSFW             bool       `json:"nsfw,omitempty"`
	LastMessageID    *Snowflake `json:"last_message_id,omitempty"`
	Bitrate          *int       `json:"bitrate,omitempty"`
	UserLimit        *int       `json:"user_limit,omitempty"`
	RateLimitPerUser *int       `json:"rate_limit_per_user,omitempty"`
	ParentID         *Snowflake `json:"parent_id,omitempty"`
}

type Message struct {
	ID              Snowflake        `json:"id"`
	ChannelID       Snowflake        `json:"channel_id"`
	GuildID         *Snowflake       `json:"guild_id,omitempty"`
	Author          User             `json:"author"`
	Member          *Member          `json:"member,omitempty"`
	Content         string           `json:"content"`
	Timestamp       time.Time        `json:"timestamp"`
	EditedTimestamp *time.Time       `json:"edited_timestamp,omitempty"`
	TTS             bool             `json:"tts"`
	MentionEveryone bool             `json:"mention_everyone"`
	Mentions        []User           `json:"mentions"`
	Embeds          []Embed          `json:"embeds"`
	Attachments     []Attachment     `json:"attachments"`
	Components      []Component      `json:"components"`
	Flags           MessageFlag      `json:"flags,omitempty"`
	Interaction     *MessageInteract `json:"interaction,omitempty"`
}

type MessageInteract struct {
	ID   Snowflake       `json:"id"`
	Type InteractionType `json:"type"`
	Name string          `json:"name"`
	User User            `json:"user"`
}

type Attachment struct {
	ID          Snowflake `json:"id"`
	Filename    string    `json:"filename"`
	Description *string   `json:"description,omitempty"`
	ContentType *string   `json:"content_type,omitempty"`
	Size        int       `json:"size"`
	URL         string    `json:"url"`
	ProxyURL    string    `json:"proxy_url"`
}

type Embed struct {
	Title       string       `json:"title,omitempty"`
	Type        string       `json:"type,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Timestamp   *time.Time   `json:"timestamp,omitempty"`
	Color       int          `json:"color,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Image       *EmbedMedia  `json:"image,omitempty"`
	Thumbnail   *EmbedMedia  `json:"thumbnail,omitempty"`
	Author      *EmbedAuthor `json:"author,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type EmbedFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedMedia struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

type EmbedAuthor struct {
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type ComponentType int

const (
	ComponentTypeActionRow         ComponentType = 1
	ComponentTypeButton            ComponentType = 2
	ComponentTypeStringSelect      ComponentType = 3
	ComponentTypeTextInput         ComponentType = 4
	ComponentTypeUserSelect        ComponentType = 5
	ComponentTypeRoleSelect        ComponentType = 6
	ComponentTypeMentionableSelect ComponentType = 7
	ComponentTypeChannelSelect     ComponentType = 8
)

type Component struct {
	Type         ComponentType   `json:"type"`
	CustomID     string          `json:"custom_id,omitempty"`
	Disabled     bool            `json:"disabled,omitempty"`
	Style        int             `json:"style,omitempty"`
	Label        string          `json:"label,omitempty"`
	Emoji        *ComponentEmoji `json:"emoji,omitempty"`
	URL          string          `json:"url,omitempty"`
	Options      []SelectOption  `json:"options,omitempty"`
	Placeholder  string          `json:"placeholder,omitempty"`
	MinValues    *int            `json:"min_values,omitempty"`
	MaxValues    *int            `json:"max_values,omitempty"`
	Value        string          `json:"value,omitempty"`
	MinLength    *int            `json:"min_length,omitempty"`
	MaxLength    *int            `json:"max_length,omitempty"`
	Required     *bool           `json:"required,omitempty"`
	ChannelTypes []int           `json:"channel_types,omitempty"`
	Components   []Component     `json:"components,omitempty"`
}

type ComponentEmoji struct {
	ID       *Snowflake `json:"id,omitempty"`
	Name     string     `json:"name,omitempty"`
	Animated bool       `json:"animated,omitempty"`
}

type SelectOption struct {
	Label       string          `json:"label"`
	Value       string          `json:"value"`
	Description string          `json:"description,omitempty"`
	Emoji       *ComponentEmoji `json:"emoji,omitempty"`
	Default     bool            `json:"default,omitempty"`
}

type InteractionType int

const (
	InteractionTypePing                           InteractionType = 1
	InteractionTypeApplicationCommand             InteractionType = 2
	InteractionTypeMessageComponent               InteractionType = 3
	InteractionTypeApplicationCommandAutocomplete InteractionType = 4
	InteractionTypeModalSubmit                    InteractionType = 5
)

type InteractionCallbackType int

const (
	CallbackTypePong                                 InteractionCallbackType = 1
	CallbackTypeChannelMessageWithSource             InteractionCallbackType = 4
	CallbackTypeDeferredChannelMessageWithSource     InteractionCallbackType = 5
	CallbackTypeDeferredUpdateMessage                InteractionCallbackType = 6
	CallbackTypeUpdateMessage                        InteractionCallbackType = 7
	CallbackTypeApplicationCommandAutocompleteResult InteractionCallbackType = 8
	CallbackTypeModal                                InteractionCallbackType = 9
)

type Interaction struct {
	ID             Snowflake        `json:"id"`
	AppID          Snowflake        `json:"application_id"`
	Type           InteractionType  `json:"type"`
	Data           *InteractionData `json:"data,omitempty"`
	GuildID        *Snowflake       `json:"guild_id,omitempty"`
	ChannelID      *Snowflake       `json:"channel_id,omitempty"`
	Member         *Member          `json:"member,omitempty"`
	User           *User            `json:"user,omitempty"`
	Token          string           `json:"token"`
	Version        int              `json:"version"`
	Message        *Message         `json:"message,omitempty"`
	AppPermissions *Permission      `json:"app_permissions,omitempty"`
	Locale         string           `json:"locale,omitempty"`
	GuildLocale    string           `json:"guild_locale,omitempty"`
}

type InteractionData struct {
	ID            Snowflake                  `json:"id,omitempty"`
	Name          string                     `json:"name,omitempty"`
	Type          int                        `json:"type,omitempty"`
	Resolved      *ResolvedData              `json:"resolved,omitempty"`
	Options       []ApplicationCommandOption `json:"options,omitempty"`
	GuildID       *Snowflake                 `json:"guild_id,omitempty"`
	CustomID      string                     `json:"custom_id,omitempty"`
	ComponentType ComponentType              `json:"component_type,omitempty"`
	Values        []string                   `json:"values,omitempty"`
	TargetID      *Snowflake                 `json:"target_id,omitempty"`
	Components    []Component                `json:"components,omitempty"`
}

type ResolvedData struct {
	Users       map[Snowflake]User       `json:"users,omitempty"`
	Members     map[Snowflake]Member     `json:"members,omitempty"`
	Roles       map[Snowflake]Role       `json:"roles,omitempty"`
	Channels    map[Snowflake]Channel    `json:"channels,omitempty"`
	Messages    map[Snowflake]Message    `json:"messages,omitempty"`
	Attachments map[Snowflake]Attachment `json:"attachments,omitempty"`
}

type Role struct {
	ID          Snowflake  `json:"id"`
	Name        string     `json:"name"`
	Color       int        `json:"color"`
	Hoist       bool       `json:"hoist"`
	Position    int        `json:"position"`
	Permissions Permission `json:"permissions"`
	Managed     bool       `json:"managed"`
	Mentionable bool       `json:"mentionable"`
}

type ApplicationCommandOptionType int

const (
	OptionTypeSubcommand      ApplicationCommandOptionType = 1
	OptionTypeSubcommandGroup ApplicationCommandOptionType = 2
	OptionTypeString          ApplicationCommandOptionType = 3
	OptionTypeInteger         ApplicationCommandOptionType = 4
	OptionTypeBoolean         ApplicationCommandOptionType = 5
	OptionTypeUser            ApplicationCommandOptionType = 6
	OptionTypeChannel         ApplicationCommandOptionType = 7
	OptionTypeRole            ApplicationCommandOptionType = 8
	OptionTypeMentionable     ApplicationCommandOptionType = 9
	OptionTypeNumber          ApplicationCommandOptionType = 10
	OptionTypeAttachment      ApplicationCommandOptionType = 11
)

type ApplicationCommandOption struct {
	Name         string                     `json:"name"`
	Type         int                        `json:"type"`
	Description  string                     `json:"description,omitempty"`
	Required     bool                       `json:"required,omitempty"`
	Choices      []Choice                   `json:"choices,omitempty"`
	Value        any                        `json:"value,omitempty"`
	Options      []ApplicationCommandOption `json:"options,omitempty"`
	ChannelTypes []int                      `json:"channel_types,omitempty"`
	MinValue     *float64                   `json:"min_value,omitempty"`
	MaxValue     *float64                   `json:"max_value,omitempty"`
	MinLength    *int                       `json:"min_length,omitempty"`
	MaxLength    *int                       `json:"max_length,omitempty"`
	Autocomplete bool                       `json:"autocomplete,omitempty"`
	Focused      bool                       `json:"focused,omitempty"`
}

type MessageFlag uint64

const (
	MessageFlagCrossposted                      MessageFlag = 1 << 0
	MessageFlagIsCrosspost                      MessageFlag = 1 << 1
	MessageFlagSuppressEmbeds                   MessageFlag = 1 << 2
	MessageFlagSourceMessageDeleted             MessageFlag = 1 << 3
	MessageFlagUrgent                           MessageFlag = 1 << 4
	MessageFlagHasThread                        MessageFlag = 1 << 5
	MessageFlagEphemeral                        MessageFlag = 1 << 6
	MessageFlagLoading                          MessageFlag = 1 << 7
	MessageFlagFailedToMentionSomeRolesInThread MessageFlag = 1 << 8
	MessageFlagSuppressNotifications            MessageFlag = 1 << 12
	MessageFlagIsVoiceMessage                   MessageFlag = 1 << 13
)

type InteractionResponse struct {
	Type InteractionCallbackType  `json:"type"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

type InteractionResponseData struct {
	TTS             bool        `json:"tts,omitempty"`
	Content         string      `json:"content,omitempty"`
	Embeds          []Embed     `json:"embeds,omitempty"`
	AllowedMentions any         `json:"allowed_mentions,omitempty"`
	Flags           MessageFlag `json:"flags,omitempty"`
	Components      []Component `json:"components,omitempty"`
	Choices         []Choice    `json:"choices,omitempty"`
	CustomID        string      `json:"custom_id,omitempty"`
	Title           string      `json:"title,omitempty"`
	Files           []File      `json:"-"`
}

type Choice struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type ActivityType int

const (
	ActivityTypeGame      ActivityType = 0
	ActivityTypeStreaming ActivityType = 1
	ActivityTypeListening ActivityType = 2
	ActivityTypeWatching  ActivityType = 3
	ActivityTypeCustom    ActivityType = 4
	ActivityTypeCompeting ActivityType = 5
)

type Activity struct {
	Name  string       `json:"name"`
	Type  ActivityType `json:"type"`
	URL   string       `json:"url,omitempty"`
	State string       `json:"state,omitempty"`
}

type StatusType string

const (
	StatusOnline    StatusType = "online"
	StatusDND       StatusType = "dnd"
	StatusIdle      StatusType = "idle"
	StatusInvisible StatusType = "invisible"
	StatusOffline   StatusType = "offline"
)

type PresenceUpdate struct {
	Since      *int64     `json:"since"`
	Activities []Activity `json:"activities"`
	Status     StatusType `json:"status"`
	AFK        bool       `json:"afk"`
}

type ApplicationCommand struct {
	ID                       *Snowflake                 `json:"id,omitempty"`
	AppID                    *Snowflake                 `json:"application_id,omitempty"`
	GuildID                  *Snowflake                 `json:"guild_id,omitempty"`
	Name                     string                     `json:"name"`
	Description              string                     `json:"description,omitempty"`
	Options                  []ApplicationCommandOption `json:"options,omitempty"`
	DefaultMemberPermissions *Permission                `json:"default_member_permissions,omitempty"`
	DMPermission             *bool                      `json:"dm_permission,omitempty"`
	NSFW                     *bool                      `json:"nsfw,omitempty"`
	Type                     int                        `json:"type,omitempty"`
}
