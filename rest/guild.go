package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sfossl/gda/types"
)

type ModifyMemberPayload struct {
	Nick                       *string            `json:"nick,omitempty"`
	Roles                      *[]types.Snowflake `json:"roles,omitempty"`
	Mute                       *bool              `json:"mute,omitempty"`
	Deaf                       *bool              `json:"deaf,omitempty"`
	ChannelID                  *types.Snowflake   `json:"channel_id,omitempty"`
	CommunicationDisabledUntil *time.Time         `json:"communication_disabled_until,omitempty"`
}

type RolePayload struct {
	Name        *string           `json:"name,omitempty"`
	Permissions *types.Permission `json:"permissions,omitempty"`
	Color       *int              `json:"color,omitempty"`
	Hoist       *bool             `json:"hoist,omitempty"`
	Mentionable *bool             `json:"mentionable,omitempty"`
}

func (c *Client) GetGuild(ctx context.Context, guildID types.Snowflake) (*types.Guild, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetGuild(guildID),
	})
	if err != nil {
		return nil, err
	}
	var g types.Guild
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (c *Client) GetGuildMember(ctx context.Context, guildID, userID types.Snowflake) (*types.Member, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetGuildMember(guildID, userID),
	})
	if err != nil {
		return nil, err
	}
	var m types.Member
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *Client) ModifyGuildMember(ctx context.Context, guildID, userID types.Snowflake, payload ModifyMemberPayload) (*types.Member, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointModifyGuildMember(guildID, userID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}
	var m types.Member
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *Client) AddGuildMemberRole(ctx context.Context, guildID, userID, roleID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointAddGuildMemberRole(guildID, userID, roleID),
	})
	return err
}

func (c *Client) RemoveGuildMemberRole(ctx context.Context, guildID, userID, roleID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointRemoveGuildMemberRole(guildID, userID, roleID),
	})
	return err
}

func (c *Client) KickMember(ctx context.Context, guildID, userID types.Snowflake, reason string) error {
	var headers http.Header
	if reason != "" {
		headers = http.Header{"X-Audit-Log-Reason": []string{reason}}
	}
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointRemoveGuildMember(guildID, userID),
		Headers:  headers,
	})
	return err
}

func (c *Client) BanMember(ctx context.Context, guildID, userID types.Snowflake, deleteMessageSeconds int, reason string) error {
	var headers http.Header
	if reason != "" {
		headers = http.Header{"X-Audit-Log-Reason": []string{reason}}
	}
	var body any
	if deleteMessageSeconds > 0 {
		body = struct {
			DeleteMessageSeconds int `json:"delete_message_seconds"`
		}{
			DeleteMessageSeconds: deleteMessageSeconds,
		}
	}
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointCreateGuildBan(guildID, userID),
		Body:     body,
		Headers:  headers,
	})
	return err
}

func (c *Client) UnbanMember(ctx context.Context, guildID, userID types.Snowflake, reason string) error {
	var headers http.Header
	if reason != "" {
		headers = http.Header{"X-Audit-Log-Reason": []string{reason}}
	}
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointRemoveGuildBan(guildID, userID),
		Headers:  headers,
	})
	return err
}

func (c *Client) TimeoutMember(ctx context.Context, guildID, userID types.Snowflake, until *time.Time, reason string) error {
	var headers http.Header
	if reason != "" {
		headers = http.Header{"X-Audit-Log-Reason": []string{reason}}
	}
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointModifyGuildMember(guildID, userID),
		Body: ModifyMemberPayload{
			CommunicationDisabledUntil: until,
		},
		Headers: headers,
	})
	return err
}

func (c *Client) GetGuildRoles(ctx context.Context, guildID types.Snowflake) ([]types.Role, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetGuildRoles(guildID),
	})
	if err != nil {
		return nil, err
	}
	var roles []types.Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (c *Client) CreateGuildRole(ctx context.Context, guildID types.Snowflake, payload RolePayload) (*types.Role, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointCreateGuildRole(guildID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}
	var role types.Role
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (c *Client) ModifyGuildRole(ctx context.Context, guildID, roleID types.Snowflake, payload RolePayload) (*types.Role, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointModifyGuildRole(guildID, roleID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}
	var role types.Role
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (c *Client) DeleteGuildRole(ctx context.Context, guildID, roleID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteGuildRole(guildID, roleID),
	})
	return err
}

