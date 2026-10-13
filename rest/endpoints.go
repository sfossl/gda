package rest

import (
	"fmt"

	"github.com/sfossl/bfrm/types"
)

const (
	APIVersion = "10"
	BaseURL    = "https://discord.com/api/v" + APIVersion
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

