package builder

import (
	"strings"
	"testing"
	"github.com/sfossl/bfrm/types"
)

func TestModalBuilderValid(t *testing.T) {
	nameInput := NewTextInput("user_name", "Your Name", TextInputStyleShort).
		SetPlaceholder("Alice")
	feedbackInput := NewTextInput("user_feedback", "Feedback", TextInputStyleParagraph)

	modal, err := NewModal("feedback_modal", "User Feedback").
		AddTextInput(nameInput).
		AddTextInput(feedbackInput).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if modal.CustomID != "feedback_modal" {
		t.Fatalf("unexpected custom_id: %s", modal.CustomID)
	}
	if modal.Title != "User Feedback" {
		t.Fatalf("unexpected title: %s", modal.Title)
	}
	if len(modal.Components) != 2 {
		t.Fatalf("expected 2 action rows, got %d", len(modal.Components))
	}
	if modal.Components[0].Type != types.ComponentTypeActionRow {
		t.Fatalf("expected action row wrapper")
	}
}

func TestModalBuilderLimits(t *testing.T) {
	_, err := NewModal("", "Title").Build()
	if err != ErrModalCustomIDRequired {
		t.Fatalf("expected ErrModalCustomIDRequired, got %v", err)
	}

	_, err = NewModal("id", "").Build()
	if err != ErrModalTitleRequired {
		t.Fatalf("expected ErrModalTitleRequired, got %v", err)
	}

	longTitle := strings.Repeat("t", 46)
	_, err = NewModal("id", longTitle).Build()
	if err != ErrModalTitleTooLong {
		t.Fatalf("expected ErrModalTitleTooLong, got %v", err)
	}

	_, err = NewModal("id", "Title").Build()
	if err != ErrModalNoComponents {
		t.Fatalf("expected ErrModalNoComponents, got %v", err)
	}

	m := NewModal("id", "Title")
	for i := 0; i < 5; i++ {
		m.AddTextInput(NewTextInput("id", "label", TextInputStyleShort))
	}
	m.AddTextInput(NewTextInput("id6", "label", TextInputStyleShort))
	_, err = m.Build()
	if err != ErrModalTooManyComponents {
		t.Fatalf("expected ErrModalTooManyComponents, got %v", err)
	}
}

func TestTextInputBuilderLimits(t *testing.T) {
	_, err := NewTextInput("", "label", TextInputStyleShort).Build()
	if err != ErrTextInputCustomIDRequired {
		t.Fatalf("expected ErrTextInputCustomIDRequired, got %v", err)
	}

	_, err = NewTextInput("id", "", TextInputStyleShort).Build()
	if err != ErrTextInputLabelRequired {
		t.Fatalf("expected ErrTextInputLabelRequired, got %v", err)
	}

	longLabel := strings.Repeat("l", 46)
	_, err = NewTextInput("id", longLabel, TextInputStyleShort).Build()
	if err != ErrTextInputLabelTooLong {
		t.Fatalf("expected ErrTextInputLabelTooLong, got %v", err)
	}
}

