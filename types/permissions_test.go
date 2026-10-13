package types

import (
	"encoding/json"
	"testing"
)

func TestPermissionsBitwise(t *testing.T) {
	p := PermissionSendMessages
	if !p.Has(PermissionSendMessages) {
		t.Fatalf("expected to have PermissionSendMessages")
	}
	if p.Has(PermissionEmbedLinks) {
		t.Fatalf("did not expect to have PermissionEmbedLinks")
	}

	p = p.Add(PermissionEmbedLinks, PermissionAttachFiles)
	if !p.Has(PermissionEmbedLinks) || !p.Has(PermissionAttachFiles) {
		t.Fatalf("expected to have added permissions")
	}

	p = p.Remove(PermissionEmbedLinks)
	if p.Has(PermissionEmbedLinks) {
		t.Fatalf("did not expect to have removed PermissionEmbedLinks")
	}
	if !p.Has(PermissionSendMessages) || !p.Has(PermissionAttachFiles) {
		t.Fatalf("expected retained permissions")
	}
}

func TestPermissionsAdministratorOverride(t *testing.T) {
	admin := PermissionAdministrator
	if !admin.Has(PermissionSendMessages) {
		t.Fatalf("admin should have PermissionSendMessages")
	}
	if !admin.Has(PermissionBanMembers) {
		t.Fatalf("admin should have PermissionBanMembers")
	}
	if !admin.Has(PermissionManageGuild) {
		t.Fatalf("admin should have PermissionManageGuild")
	}
}

func TestPermissionsJSON(t *testing.T) {
	type wrapper struct {
		Perms Permission `json:"permissions"`
	}

	w := wrapper{Perms: PermissionSendMessages | PermissionEmbedLinks}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed wrapper
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if parsed.Perms != w.Perms {
		t.Fatalf("expected %d, got %d", w.Perms, parsed.Perms)
	}

	var nullWrapper wrapper
	if err := json.Unmarshal([]byte(`{"permissions":null}`), &nullWrapper); err != nil {
		t.Fatalf("unmarshal null error: %v", err)
	}
	if nullWrapper.Perms != 0 {
		t.Fatalf("expected 0, got %d", nullWrapper.Perms)
	}
}

