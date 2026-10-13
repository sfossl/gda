package types

import (
	"strings"
	"testing"
)

func TestConstants(t *testing.T) {
	if APIVersion != "10" {
		t.Fatalf("expected APIVersion 10, got %s", APIVersion)
	}
	if !strings.HasPrefix(UserAgent, "DiscordBot") {
		t.Fatalf("unexpected user agent: %s", UserAgent)
	}
	if MaxPayloadSize != 10<<20 {
		t.Fatalf("unexpected max payload size: %d", MaxPayloadSize)
	}
}

