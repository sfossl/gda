package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestChannelAndTicketOperations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/channels/400":
			ch := types.Channel{
				ID:   types.Snowflake(400),
				Type: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ch)

		case r.Method == "POST" && r.URL.Path == "/guilds/100/channels":
			var p CreateChannelPayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			ch := types.Channel{
				ID:   types.Snowflake(401),
				Name: &p.Name,
				Type: p.Type,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ch)

		case r.Method == "PATCH" && r.URL.Path == "/channels/401":
			var p ModifyChannelPayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			ch := types.Channel{
				ID:   types.Snowflake(401),
				Name: p.Name,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ch)

		case r.Method == "DELETE" && r.URL.Path == "/channels/401":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "PUT" && r.URL.Path == "/channels/401/permissions/500":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/channels/401/permissions/500":
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	ctx := context.Background()

	ch, err := client.GetChannel(ctx, types.Snowflake(400))
	if err != nil || ch.ID != types.Snowflake(400) {
		t.Fatalf("unexpected channel: %+v, err: %v", ch, err)
	}

	created, err := client.CreateGuildChannel(ctx, types.Snowflake(100), CreateChannelPayload{
		Name: "ticket-001",
		Type: 0,
	})
	if err != nil || created.Name == nil || *created.Name != "ticket-001" {
		t.Fatalf("unexpected created channel: %+v, err: %v", created, err)
	}

	newName := "ticket-closed"
	modified, err := client.ModifyChannel(ctx, types.Snowflake(401), ModifyChannelPayload{
		Name: &newName,
	})
	if err != nil || modified.Name == nil || *modified.Name != "ticket-closed" {
		t.Fatalf("unexpected modified channel: %+v, err: %v", modified, err)
	}

	if err := client.EditChannelPermissions(ctx, types.Snowflake(401), types.Snowflake(500), types.PermissionOverwrite{
		ID:    types.Snowflake(500),
		Type:  0,
		Allow: types.PermissionViewChannel,
		Deny:  types.PermissionSendMessages,
	}); err != nil {
		t.Fatalf("unexpected edit permissions error: %v", err)
	}

	if err := client.DeleteChannelPermission(ctx, types.Snowflake(401), types.Snowflake(500)); err != nil {
		t.Fatalf("unexpected delete permissions error: %v", err)
	}

	if err := client.DeleteChannel(ctx, types.Snowflake(401)); err != nil {
		t.Fatalf("unexpected delete channel error: %v", err)
	}
}

func TestMessageReactionsAndPins(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/channels/400/messages/600":
			msg := types.Message{
				ID:      types.Snowflake(600),
				Content: "sample message",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(msg)

		case r.Method == "PUT" && r.URL.Path == "/channels/400/pins/600":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/channels/400/pins/600":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "PUT" && r.URL.Path == "/channels/400/messages/600/reactions/\u2705/@me":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/channels/400/messages/600/reactions/\u2705/@me":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/channels/400/messages/600/reactions/\u2705/700":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "GET" && r.URL.Path == "/users/@me":
			u := types.User{
				ID:       types.Snowflake(888),
				Username: "selfbot",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(u)

		case r.Method == "GET" && r.URL.Path == "/users/999":
			u := types.User{
				ID:       types.Snowflake(999),
				Username: "otheruser",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(u)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	ctx := context.Background()

	msg, err := client.GetMessage(ctx, types.Snowflake(400), types.Snowflake(600))
	if err != nil || msg.Content != "sample message" {
		t.Fatalf("unexpected message: %+v, err: %v", msg, err)
	}

	if err := client.PinMessage(ctx, types.Snowflake(400), types.Snowflake(600)); err != nil {
		t.Fatalf("unexpected pin error: %v", err)
	}

	if err := client.UnpinMessage(ctx, types.Snowflake(400), types.Snowflake(600)); err != nil {
		t.Fatalf("unexpected unpin error: %v", err)
	}

	emoji := "\u2705"
	if err := client.CreateReaction(ctx, types.Snowflake(400), types.Snowflake(600), emoji); err != nil {
		t.Fatalf("unexpected create reaction error: %v", err)
	}

	if err := client.DeleteOwnReaction(ctx, types.Snowflake(400), types.Snowflake(600), emoji); err != nil {
		t.Fatalf("unexpected delete own reaction error: %v", err)
	}

	if err := client.DeleteUserReaction(ctx, types.Snowflake(400), types.Snowflake(600), emoji, types.Snowflake(700)); err != nil {
		t.Fatalf("unexpected delete user reaction error: %v", err)
	}

	me, err := client.GetCurrentUser(ctx)
	if err != nil || me.Username != "selfbot" {
		t.Fatalf("unexpected self user: %+v, err: %v", me, err)
	}

	user, err := client.GetUser(ctx, types.Snowflake(999))
	if err != nil || user.Username != "otheruser" {
		t.Fatalf("unexpected user: %+v, err: %v", user, err)
	}
}
