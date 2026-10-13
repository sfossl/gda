package builder

import (
	"strings"
	"testing"

	"github.com/sfossl/bfrm/types"
)

func TestButtonBuilderValid(t *testing.T) {
	btn, err := NewButton().
		SetStyle(ButtonStyleSuccess).
		SetCustomID("snooze:15").
		SetLabel("Snooze 15m").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if btn.Type != types.ComponentTypeButton {
		t.Fatalf("expected button type, got %d", btn.Type)
	}
	if btn.Style != int(ButtonStyleSuccess) {
		t.Fatalf("expected style success, got %d", btn.Style)
	}
	if btn.CustomID != "snooze:15" {
		t.Fatalf("expected custom_id snooze:15, got %s", btn.CustomID)
	}
}

func TestButtonBuilderLink(t *testing.T) {
	btn, err := NewButton().
		SetStyle(ButtonStyleLink).
		SetURL("https://example.com").
		SetLabel("Open Dashboard").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if btn.URL != "https://example.com" {
		t.Fatalf("unexpected url: %s", btn.URL)
	}

	_, err = NewButton().SetStyle(ButtonStyleLink).Build()
	if err != ErrButtonURLRequired {
		t.Fatalf("expected ErrButtonURLRequired, got %v", err)
	}

	_, err = NewButton().
		SetStyle(ButtonStyleLink).
		SetURL("https://example.com").
		SetCustomID("link_id").
		Build()
	if err != ErrButtonURLNotAllowed {
		t.Fatalf("expected ErrButtonURLNotAllowed, got %v", err)
	}
}

func TestButtonBuilderLimits(t *testing.T) {
	_, err := NewButton().SetCustomID(strings.Repeat("c", 101)).Build()
	if err != ErrButtonCustomIDTooLong {
		t.Fatalf("expected ErrButtonCustomIDTooLong, got %v", err)
	}

	_, err = NewButton().SetCustomID("ok").SetLabel(strings.Repeat("l", 81)).Build()
	if err != ErrButtonLabelTooLong {
		t.Fatalf("expected ErrButtonLabelTooLong, got %v", err)
	}

	_, err = NewButton().SetStyle(ButtonStylePrimary).Build()
	if err != ErrButtonCustomIDRequired {
		t.Fatalf("expected ErrButtonCustomIDRequired, got %v", err)
	}
}

func TestSelectMenuBuilder(t *testing.T) {
	sel, err := NewStringSelect("select_roles").
		SetPlaceholder("Choose a role").
		AddOption("Admin", "role_admin", "Administrator", false).
		AddOption("Member", "role_member", "Normal member", true).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.Type != types.ComponentTypeStringSelect {
		t.Fatalf("expected string select type, got %d", sel.Type)
	}
	if len(sel.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(sel.Options))
	}

	userSel, err := NewUserSelect("select_user").Build()
	if err != nil {
		t.Fatalf("unexpected user select error: %v", err)
	}
	if userSel.Type != types.ComponentTypeUserSelect {
		t.Fatalf("expected user select type, got %d", userSel.Type)
	}

	_, err = NewStringSelect("").Build()
	if err != ErrSelectCustomIDRequired {
		t.Fatalf("expected ErrSelectCustomIDRequired, got %v", err)
	}
}

func TestActionRowBuilder(t *testing.T) {
	b1, _ := NewButton().SetCustomID("b1").SetLabel("1").Build()
	b2, _ := NewButton().SetCustomID("b2").SetLabel("2").Build()
	b3, _ := NewButton().SetCustomID("b3").SetLabel("3").Build()
	b4, _ := NewButton().SetCustomID("b4").SetLabel("4").Build()
	b5, _ := NewButton().SetCustomID("b5").SetLabel("5").Build()
	b6, _ := NewButton().SetCustomID("b6").SetLabel("6").Build()

	row, err := NewActionRow(b1, b2, b3, b4, b5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row.Type != types.ComponentTypeActionRow {
		t.Fatalf("expected action row type, got %d", row.Type)
	}
	if len(row.Components) != 5 {
		t.Fatalf("expected 5 components, got %d", len(row.Components))
	}

	_, err = NewActionRow(b1, b2, b3, b4, b5, b6)
	if err != ErrActionRowTooManyButtons {
		t.Fatalf("expected ErrActionRowTooManyButtons, got %v", err)
	}

	sel, _ := NewStringSelect("sel").AddOption("1", "1", "", false).Build()
	_, err = NewActionRow(sel, b1)
	if err != ErrActionRowSelectNotAlone {
		t.Fatalf("expected ErrActionRowSelectNotAlone, got %v", err)
	}

	txt := types.Component{Type: types.ComponentTypeTextInput, CustomID: "txt"}
	_, err = NewActionRow(txt, b1)
	if err != ErrActionRowSelectNotAlone {
		t.Fatalf("expected ErrActionRowSelectNotAlone for text input, got %v", err)
	}

	_, err = NewActionRow()
	if err != ErrActionRowEmpty {
		t.Fatalf("expected ErrActionRowEmpty, got %v", err)
	}
}

