package builder

import (
	"strings"
	"testing"
	"time"
)

func TestEmbedBuilderValid(t *testing.T) {
	now := time.Now()
	embed, err := NewEmbed().
		SetTitle("Task Reminder").
		SetDescription("Meeting at 3PM").
		SetColor(0x5865F2).
		SetTimestamp(now).
		SetAuthor("PWS", "https://example.com", "").
		SetFooter("Reminders Bot", "").
		AddField("Priority", "High", true).
		AddField("Status", "Pending", true).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embed.Title != "Task Reminder" {
		t.Fatalf("unexpected title: %s", embed.Title)
	}
	if embed.Color != 0x5865F2 {
		t.Fatalf("unexpected color: %d", embed.Color)
	}
	if len(embed.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(embed.Fields))
	}
}

func TestEmbedBuilderTitleLimit(t *testing.T) {
	longTitle := strings.Repeat("a", 257)
	_, err := NewEmbed().SetTitle(longTitle).Build()
	if err != ErrEmbedTitleTooLong {
		t.Fatalf("expected ErrEmbedTitleTooLong, got %v", err)
	}
}

func TestEmbedBuilderDescriptionLimit(t *testing.T) {
	longDesc := strings.Repeat("a", 4097)
	_, err := NewEmbed().SetDescription(longDesc).Build()
	if err != ErrEmbedDescriptionTooLong {
		t.Fatalf("expected ErrEmbedDescriptionTooLong, got %v", err)
	}
}

func TestEmbedBuilderFieldValidation(t *testing.T) {
	_, err := NewEmbed().AddField("", "val", false).Build()
	if err != ErrEmbedFieldNameEmpty {
		t.Fatalf("expected ErrEmbedFieldNameEmpty, got %v", err)
	}

	_, err = NewEmbed().AddField("name", "", false).Build()
	if err != ErrEmbedFieldValueEmpty {
		t.Fatalf("expected ErrEmbedFieldValueEmpty, got %v", err)
	}

	longName := strings.Repeat("n", 257)
	_, err = NewEmbed().AddField(longName, "val", false).Build()
	if err != ErrEmbedFieldNameTooLong {
		t.Fatalf("expected ErrEmbedFieldNameTooLong, got %v", err)
	}

	longVal := strings.Repeat("v", 1025)
	_, err = NewEmbed().AddField("name", longVal, false).Build()
	if err != ErrEmbedFieldValueTooLong {
		t.Fatalf("expected ErrEmbedFieldValueTooLong, got %v", err)
	}

	b := NewEmbed()
	for i := 0; i < 25; i++ {
		b.AddField("n", "v", false)
	}
	b.AddField("n26", "v26", false)
	_, err = b.Build()
	if err != ErrEmbedTooManyFields {
		t.Fatalf("expected ErrEmbedTooManyFields, got %v", err)
	}
}

func TestEmbedBuilderTotalLengthLimit(t *testing.T) {
	b := NewEmbed().
		SetTitle(strings.Repeat("t", 250)).
		SetDescription(strings.Repeat("d", 4000))

	for i := 0; i < 2; i++ {
		b.AddField(strings.Repeat("n", 200), strings.Repeat("v", 800), false)
	}

	_, err := b.Build()
	if err != ErrEmbedTotalLengthExceeded {
		t.Fatalf("expected ErrEmbedTotalLengthExceeded, got %v", err)
	}
}

