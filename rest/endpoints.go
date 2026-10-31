package rest

import (
	"fmt"
	"net/url"

	"github.com/sfossl/gda/types"
)

const (
	BaseURL = "https://discord.com/api/v" + types.APIVersion
)

type Endpoint struct {
	Method string
	Path   string
	Key    string
}

func EndpointCreateDM() Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   "/users/@me/channels",
		Key:    "POST:/users/@me/channels",
	}
}

func EndpointGetChannel(channelID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/channels/%s", channelID),
		Key:    fmt.Sprintf("GET:/channels/%s", channelID),
	}
}

func EndpointSendMessage(channelID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/channels/%s/messages", channelID),
		Key:    fmt.Sprintf("POST:/channels/%s/messages", channelID),
	}
}

func EndpointEditMessage(channelID, messageID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PATCH",
		Path:   fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID),
		Key:    fmt.Sprintf("PATCH:/channels/%s/messages", channelID),
	}
}

func EndpointDeleteMessage(channelID, messageID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID),
		Key:    fmt.Sprintf("DELETE:/channels/%s/messages", channelID),
	}
}

func EndpointTriggerTyping(channelID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/channels/%s/typing", channelID),
		Key:    fmt.Sprintf("POST:/channels/%s/typing", channelID),
	}
}

func EndpointInteractionCallback(interactionID types.Snowflake, token string) Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/interactions/%s/%s/callback", interactionID, token),
		Key:    "POST:/interactions/callback",
	}
}

func EndpointFollowupMessage(appID types.Snowflake, token string) Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/webhooks/%s/%s", appID, token),
		Key:    fmt.Sprintf("POST:/webhooks/%s", appID),
	}
}

func EndpointEditOriginalInteraction(appID types.Snowflake, token string) Endpoint {
	return Endpoint{
		Method: "PATCH",
		Path:   fmt.Sprintf("/webhooks/%s/%s/messages/@original", appID, token),
		Key:    fmt.Sprintf("PATCH:/webhooks/%s/messages/@original", appID),
	}
}

func EndpointDeleteOriginalInteraction(appID types.Snowflake, token string) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/webhooks/%s/%s/messages/@original", appID, token),
		Key:    fmt.Sprintf("DELETE:/webhooks/%s/messages/@original", appID),
	}
}

func EndpointBulkOverwriteGlobalCommands(appID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/applications/%s/commands", appID),
		Key:    fmt.Sprintf("PUT:/applications/%s/commands", appID),
	}
}

func EndpointBulkOverwriteGuildCommands(appID, guildID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID),
		Key:    fmt.Sprintf("PUT:/applications/%s/guilds/%s/commands", appID, guildID),
	}
}

func EndpointGetGuild(guildID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/guilds/%s", guildID),
		Key:    fmt.Sprintf("GET:/guilds/%s", guildID),
	}
}

func EndpointGetGuildMember(guildID, userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/guilds/%s/members/%s", guildID, userID),
		Key:    fmt.Sprintf("GET:/guilds/%s/members", guildID),
	}
}

func EndpointModifyGuildMember(guildID, userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PATCH",
		Path:   fmt.Sprintf("/guilds/%s/members/%s", guildID, userID),
		Key:    fmt.Sprintf("PATCH:/guilds/%s/members", guildID),
	}
}

func EndpointAddGuildMemberRole(guildID, userID, roleID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID),
		Key:    fmt.Sprintf("PUT:/guilds/%s/members/roles", guildID),
	}
}

func EndpointRemoveGuildMemberRole(guildID, userID, roleID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID),
		Key:    fmt.Sprintf("DELETE:/guilds/%s/members/roles", guildID),
	}
}

func EndpointRemoveGuildMember(guildID, userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/guilds/%s/members/%s", guildID, userID),
		Key:    fmt.Sprintf("DELETE:/guilds/%s/members", guildID),
	}
}

func EndpointCreateGuildBan(guildID, userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/guilds/%s/bans/%s", guildID, userID),
		Key:    fmt.Sprintf("PUT:/guilds/%s/bans", guildID),
	}
}

func EndpointRemoveGuildBan(guildID, userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/guilds/%s/bans/%s", guildID, userID),
		Key:    fmt.Sprintf("DELETE:/guilds/%s/bans", guildID),
	}
}

func EndpointGetGuildRoles(guildID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/guilds/%s/roles", guildID),
		Key:    fmt.Sprintf("GET:/guilds/%s/roles", guildID),
	}
}

func EndpointCreateGuildRole(guildID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/guilds/%s/roles", guildID),
		Key:    fmt.Sprintf("POST:/guilds/%s/roles", guildID),
	}
}

func EndpointModifyGuildRole(guildID, roleID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PATCH",
		Path:   fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID),
		Key:    fmt.Sprintf("PATCH:/guilds/%s/roles", guildID),
	}
}

func EndpointDeleteGuildRole(guildID, roleID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID),
		Key:    fmt.Sprintf("DELETE:/guilds/%s/roles", guildID),
	}
}

func EndpointCreateGuildChannel(guildID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "POST",
		Path:   fmt.Sprintf("/guilds/%s/channels", guildID),
		Key:    fmt.Sprintf("POST:/guilds/%s/channels", guildID),
	}
}

func EndpointModifyChannel(channelID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PATCH",
		Path:   fmt.Sprintf("/channels/%s", channelID),
		Key:    fmt.Sprintf("PATCH:/channels/%s", channelID),
	}
}

func EndpointDeleteChannel(channelID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/channels/%s", channelID),
		Key:    fmt.Sprintf("DELETE:/channels/%s", channelID),
	}
}

func EndpointEditChannelPermissions(channelID, overwriteID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/channels/%s/permissions/%s", channelID, overwriteID),
		Key:    fmt.Sprintf("PUT:/channels/%s/permissions", channelID),
	}
}

func EndpointDeleteChannelPermission(channelID, overwriteID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/channels/%s/permissions/%s", channelID, overwriteID),
		Key:    fmt.Sprintf("DELETE:/channels/%s/permissions", channelID),
	}
}

func EndpointGetMessage(channelID, messageID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID),
		Key:    fmt.Sprintf("GET:/channels/%s/messages", channelID),
	}
}

func EndpointPinMessage(channelID, messageID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/channels/%s/pins/%s", channelID, messageID),
		Key:    fmt.Sprintf("PUT:/channels/%s/pins", channelID),
	}
}

func EndpointUnpinMessage(channelID, messageID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/channels/%s/pins/%s", channelID, messageID),
		Key:    fmt.Sprintf("DELETE:/channels/%s/pins", channelID),
	}
}

func EndpointCreateReaction(channelID, messageID types.Snowflake, emoji string) Endpoint {
	return Endpoint{
		Method: "PUT",
		Path:   fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, url.PathEscape(emoji)),
		Key:    fmt.Sprintf("PUT:/channels/%s/messages/reactions", channelID),
	}
}

func EndpointDeleteOwnReaction(channelID, messageID types.Snowflake, emoji string) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, url.PathEscape(emoji)),
		Key:    fmt.Sprintf("DELETE:/channels/%s/messages/reactions", channelID),
	}
}

func EndpointDeleteUserReaction(channelID, messageID types.Snowflake, emoji string, userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "DELETE",
		Path:   fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/%s", channelID, messageID, url.PathEscape(emoji), userID),
		Key:    fmt.Sprintf("DELETE:/channels/%s/messages/reactions", channelID),
	}
}

func EndpointGetCurrentUser() Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   "/users/@me",
		Key:    "GET:/users/@me",
	}
}

func EndpointGetUser(userID types.Snowflake) Endpoint {
	return Endpoint{
		Method: "GET",
		Path:   fmt.Sprintf("/users/%s", userID),
		Key:    "GET:/users",
	}
}
