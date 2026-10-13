package gda

import (
	"strings"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-token")
	if cfg.Token != "test-token" {
		t.Fatalf("expected test-token, got %s", cfg.Token)
	}
	if cfg.HTTPTimeout != 10*time.Second {
		t.Fatalf("expected 10s timeout, got %v", cfg.HTTPTimeout)
	}
	if cfg.Intents != types.IntentsDefault {
		t.Fatalf("expected IntentsDefault, got %d", cfg.Intents)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	empty := Config{}
	if err := empty.Validate(); err != ErrMissingToken {
		t.Fatalf("expected ErrMissingToken, got %v", err)
	}

	invalidShard := Config{
		Token:      "token",
		ShardID:    2,
		ShardCount: 2,
	}
	if err := invalidShard.Validate(); err != ErrInvalidShard {
		t.Fatalf("expected ErrInvalidShard, got %v", err)
	}

	bounds := Config{
		Token:          "token",
		LargeThreshold: 10,
	}
	if err := bounds.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bounds.LargeThreshold != 50 {
		t.Fatalf("expected clamp to 50, got %d", bounds.LargeThreshold)
	}

	boundsUpper := Config{
		Token:          "token",
		LargeThreshold: 500,
	}
	if err := boundsUpper.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if boundsUpper.LargeThreshold != 250 {
		t.Fatalf("expected clamp to 250, got %d", boundsUpper.LargeThreshold)
	}
}

func TestConfigStringRedaction(t *testing.T) {
	cfg := DefaultConfig("super-secret-token")
	str := cfg.String()
	if strings.Contains(str, "super-secret-token") {
		t.Fatalf("token exposed in String(): %s", str)
	}
	if !strings.Contains(str, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] in String(): %s", str)
	}
}

