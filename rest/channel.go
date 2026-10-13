package rest

import (
	"context"
	"encoding/json"

	"github.com/sfossl/gda/types"
)

type CreateChannelPayload struct {
	Name                 string                      `json:"name"`
	Type                 int                         `json:"type,omitempty"`
	Topic                string                      `json:"topic,omitempty"`
	Bitrate              int                         `json:"bitrate,omitempty"`
	UserLimit            int                         `json:"user_limit,omitempty"`
	ParentID             *types.Snowflake            `json:"parent_id,omitempty"`
	NSFW                 bool                        `json:"nsfw,omitempty"`
	Position             *int                        `json:"position,omitempty"`
	PermissionOverwrites []types.PermissionOverwrite `json:"permission_overwrites,omitempty"`
}

type ModifyChannelPayload struct {
	Name                 *string                     `json:"name,omitempty"`
	Type                 *int                        `json:"type,omitempty"`
	Position             *int                        `json:"position,omitempty"`
	Topic                *string                     `json:"topic,omitempty"`
	NSFW                 *bool                       `json:"nsfw,omitempty"`
	Bitrate              *int                        `json:"bitrate,omitempty"`
	UserLimit            *int                        `json:"user_limit,omitempty"`
	ParentID             *types.Snowflake            `json:"parent_id,omitempty"`
	PermissionOverwrites []types.PermissionOverwrite `json:"permission_overwrites,omitempty"`
}

func (c *Client) GetChannel(ctx context.Context, channelID types.Snowflake) (*types.Channel, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetChannel(channelID),
	})
	if err != nil {
		return nil, err
	}
	var ch types.Channel
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (c *Client) CreateGuildChannel(ctx context.Context, guildID types.Snowflake, payload CreateChannelPayload) (*types.Channel, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointCreateGuildChannel(guildID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}
	var ch types.Channel
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (c *Client) ModifyChannel(ctx context.Context, channelID types.Snowflake, payload ModifyChannelPayload) (*types.Channel, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointModifyChannel(channelID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}
	var ch types.Channel
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (c *Client) DeleteChannel(ctx context.Context, channelID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteChannel(channelID),
	})
	return err
}

func (c *Client) EditChannelPermissions(ctx context.Context, channelID, overwriteID types.Snowflake, overwrite types.PermissionOverwrite) error {
	payload := struct {
		Allow types.Permission `json:"allow"`
		Deny  types.Permission `json:"deny"`
		Type  int              `json:"type"`
	}{
		Allow: overwrite.Allow,
		Deny:  overwrite.Deny,
		Type:  overwrite.Type,
	}
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointEditChannelPermissions(channelID, overwriteID),
		Body:     payload,
	})
	return err
}

func (c *Client) DeleteChannelPermission(ctx context.Context, channelID, overwriteID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteChannelPermission(channelID, overwriteID),
	})
	return err
}

func (c *Client) GetMessage(ctx context.Context, channelID, messageID types.Snowflake) (*types.Message, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetMessage(channelID, messageID),
	})
	if err != nil {
		return nil, err
	}
	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) PinMessage(ctx context.Context, channelID, messageID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointPinMessage(channelID, messageID),
	})
	return err
}

func (c *Client) UnpinMessage(ctx context.Context, channelID, messageID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointUnpinMessage(channelID, messageID),
	})
	return err
}

func (c *Client) CreateReaction(ctx context.Context, channelID, messageID types.Snowflake, emoji string) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointCreateReaction(channelID, messageID, emoji),
	})
	return err
}

func (c *Client) DeleteOwnReaction(ctx context.Context, channelID, messageID types.Snowflake, emoji string) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteOwnReaction(channelID, messageID, emoji),
	})
	return err
}

func (c *Client) DeleteUserReaction(ctx context.Context, channelID, messageID types.Snowflake, emoji string, userID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteUserReaction(channelID, messageID, emoji, userID),
	})
	return err
}

func (c *Client) GetCurrentUser(ctx context.Context) (*types.User, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetCurrentUser(),
	})
	if err != nil {
		return nil, err
	}
	var u types.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *Client) GetUser(ctx context.Context, userID types.Snowflake) (*types.User, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointGetUser(userID),
	})
	if err != nil {
		return nil, err
	}
	var u types.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

