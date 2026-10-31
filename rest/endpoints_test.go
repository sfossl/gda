package rest

import (
	"testing"

	"github.com/sfossl/gda/types"
)

func TestEndpointsFormatting(t *testing.T) {
	dm := EndpointCreateDM()
	if dm.Method != "POST" || dm.Path != "/users/@me/channels" {
		t.Fatalf("unexpected DM endpoint: %+v", dm)
	}
	if dm.Key != "POST:/users/@me/channels" {
		t.Fatalf("unexpected DM key: %s", dm.Key)
	}

	channelID := types.Snowflake(123456789)
	send := EndpointSendMessage(channelID)
	if send.Path != "/channels/123456789/messages" {
		t.Fatalf("unexpected send path: %s", send.Path)
	}
	if send.Key != "POST:/channels/123456789/messages" {
		t.Fatalf("unexpected send key: %s", send.Key)
	}

	msgID := types.Snowflake(987654321)
	edit := EndpointEditMessage(channelID, msgID)
	if edit.Path != "/channels/123456789/messages/987654321" {
		t.Fatalf("unexpected edit path: %s", edit.Path)
	}
	if edit.Key != "PATCH:/channels/123456789/messages" {
		t.Fatalf("unexpected edit key (should not include messageID): %s", edit.Key)
	}

	del := EndpointDeleteMessage(channelID, msgID)
	if del.Path != "/channels/123456789/messages/987654321" {
		t.Fatalf("unexpected delete path: %s", del.Path)
	}
	if del.Key != "DELETE:/channels/123456789/messages" {
		t.Fatalf("unexpected delete key: %s", del.Key)
	}

	appID := types.Snowflake(555555555)
	guildID := types.Snowflake(666666666)
	cmd := EndpointBulkOverwriteGuildCommands(appID, guildID)
	if cmd.Path != "/applications/555555555/guilds/666666666/commands" {
		t.Fatalf("unexpected guild command path: %s", cmd.Path)
	}
	if cmd.Key != "PUT:/applications/555555555/guilds/666666666/commands" {
		t.Fatalf("unexpected guild command key: %s", cmd.Key)
	}
}
