package types

import "io"

type File struct {
	Name        string
	Description string
	Reader      io.Reader
}

type Guild struct {
	ID                          Snowflake  `json:"id"`
	Name                        string     `json:"name"`
	Icon                        *string    `json:"icon,omitempty"`
	Description                 *string    `json:"description,omitempty"`
	OwnerID                     Snowflake  `json:"owner_id"`
	Permissions                 Permission `json:"permissions,omitempty"`
	AfkChannelID                *Snowflake `json:"afk_channel_id,omitempty"`
	AfkTimeout                  int        `json:"afk_timeout"`
	VerificationLevel           int        `json:"verification_level"`
	DefaultMessageNotifications int        `json:"default_message_notifications"`
	ExplicitContentFilter       int        `json:"explicit_content_filter"`
	Roles                       []Role     `json:"roles,omitempty"`
	Features                    []string   `json:"features,omitempty"`
	MFALevel                    int        `json:"mfa_level"`
	SystemChannelID             *Snowflake `json:"system_channel_id,omitempty"`
	RulesChannelID              *Snowflake `json:"rules_channel_id,omitempty"`
	PublicUpdatesChannelID      *Snowflake `json:"public_updates_channel_id,omitempty"`
	PreferredLocale             string     `json:"preferred_locale"`
	VanityURLCode               *string    `json:"vanity_url_code,omitempty"`
	ApproximateMemberCount      int        `json:"approximate_member_count,omitempty"`
}

type Ban struct {
	Reason *string `json:"reason,omitempty"`
	User   User    `json:"user"`
}

type PermissionOverwrite struct {
	ID    Snowflake  `json:"id"`
	Type  int        `json:"type"`
	Allow Permission `json:"allow"`
	Deny  Permission `json:"deny"`
}
