package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestGuildAndMemberOperations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/guilds/100":
			g := types.Guild{
				ID:   types.Snowflake(100),
				Name: "Test Guild",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(g)

		case r.Method == "GET" && r.URL.Path == "/guilds/100/members/200":
			m := types.Member{
				User: &types.User{
					ID:       types.Snowflake(200),
					Username: "alice",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(m)

		case r.Method == "PATCH" && r.URL.Path == "/guilds/100/members/200":
			var payload ModifyMemberPayload
			_ = json.NewDecoder(r.Body).Decode(&payload)
			m := types.Member{
				Nick: payload.Nick,
				User: &types.User{
					ID:       types.Snowflake(200),
					Username: "alice",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(m)

		case r.Method == "PUT" && r.URL.Path == "/guilds/100/members/200/roles/300":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/guilds/100/members/200/roles/300":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/guilds/100/members/200":
			if r.Header.Get("X-Audit-Log-Reason") != "rule violation" {
				t.Errorf("expected audit reason 'rule violation', got: %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "PUT" && r.URL.Path == "/guilds/100/bans/200":
			if r.Header.Get("X-Audit-Log-Reason") != "banned for spam" {
				t.Errorf("expected audit reason 'banned for spam', got: %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			var body struct {
				DeleteMessageSeconds int `json:"delete_message_seconds"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.DeleteMessageSeconds != 3600 {
				t.Errorf("expected delete seconds 3600, got: %d", body.DeleteMessageSeconds)
			}
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "DELETE" && r.URL.Path == "/guilds/100/bans/200":
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	ctx := context.Background()

	g, err := client.GetGuild(ctx, types.Snowflake(100))
	if err != nil || g.Name != "Test Guild" {
		t.Fatalf("unexpected guild: %+v, err: %v", g, err)
	}

	m, err := client.GetGuildMember(ctx, types.Snowflake(100), types.Snowflake(200))
	if err != nil || m.User.Username != "alice" {
		t.Fatalf("unexpected member: %+v, err: %v", m, err)
	}

	newNick := "AliceInChains"
	modified, err := client.ModifyGuildMember(ctx, types.Snowflake(100), types.Snowflake(200), ModifyMemberPayload{
		Nick: &newNick,
	})
	if err != nil || modified.Nick == nil || *modified.Nick != "AliceInChains" {
		t.Fatalf("unexpected modified member: %+v, err: %v", modified, err)
	}

	if err := client.AddGuildMemberRole(ctx, types.Snowflake(100), types.Snowflake(200), types.Snowflake(300)); err != nil {
		t.Fatalf("unexpected add role error: %v", err)
	}

	if err := client.RemoveGuildMemberRole(ctx, types.Snowflake(100), types.Snowflake(200), types.Snowflake(300)); err != nil {
		t.Fatalf("unexpected remove role error: %v", err)
	}

	if err := client.KickMember(ctx, types.Snowflake(100), types.Snowflake(200), "rule violation"); err != nil {
		t.Fatalf("unexpected kick error: %v", err)
	}

	if err := client.BanMember(ctx, types.Snowflake(100), types.Snowflake(200), 3600, "banned for spam"); err != nil {
		t.Fatalf("unexpected ban error: %v", err)
	}

	if err := client.UnbanMember(ctx, types.Snowflake(100), types.Snowflake(200), "pardon"); err != nil {
		t.Fatalf("unexpected unban error: %v", err)
	}

	until := time.Now().Add(10 * time.Minute)
	if err := client.TimeoutMember(ctx, types.Snowflake(100), types.Snowflake(200), &until, "cooling down"); err != nil {
		t.Fatalf("unexpected timeout error: %v", err)
	}
}

func TestGuildRoleOperations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/guilds/100/roles":
			roles := []types.Role{
				{ID: types.Snowflake(301), Name: "Admin"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(roles)

		case r.Method == "POST" && r.URL.Path == "/guilds/100/roles":
			var p RolePayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			name := "New Role"
			if p.Name != nil {
				name = *p.Name
			}
			role := types.Role{
				ID:   types.Snowflake(302),
				Name: name,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(role)

		case r.Method == "PATCH" && r.URL.Path == "/guilds/100/roles/302":
			var p RolePayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			role := types.Role{
				ID:   types.Snowflake(302),
				Name: *p.Name,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(role)

		case r.Method == "DELETE" && r.URL.Path == "/guilds/100/roles/302":
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	ctx := context.Background()

	roles, err := client.GetGuildRoles(ctx, types.Snowflake(100))
	if err != nil || len(roles) != 1 || roles[0].Name != "Admin" {
		t.Fatalf("unexpected roles: %+v, err: %v", roles, err)
	}

	roleName := "Moderator"
	created, err := client.CreateGuildRole(ctx, types.Snowflake(100), RolePayload{
		Name: &roleName,
	})
	if err != nil || created.Name != "Moderator" {
		t.Fatalf("unexpected created role: %+v, err: %v", created, err)
	}

	updatedName := "Senior Moderator"
	modified, err := client.ModifyGuildRole(ctx, types.Snowflake(100), types.Snowflake(302), RolePayload{
		Name: &updatedName,
	})
	if err != nil || modified.Name != "Senior Moderator" {
		t.Fatalf("unexpected modified role: %+v, err: %v", modified, err)
	}

	if err := client.DeleteGuildRole(ctx, types.Snowflake(100), types.Snowflake(302)); err != nil {
		t.Fatalf("unexpected delete role error: %v", err)
	}
}
