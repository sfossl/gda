package gda

import (
	"errors"
	"fmt"
	"time"

	"github.com/sfossl/gda/types"
)

var (
	ErrMissingToken = errors.New("bot token is required")
	ErrInvalidShard = errors.New("shard id must be non-negative and less than shard count")
)

type Config struct {
	Token          string
	AppID          string
	PublicKey      string
	Intents        types.Intent
	ShardID        int
	ShardCount     int
	HTTPTimeout    time.Duration
	MaxRetries     int
	LargeThreshold int
	AutoDefer      bool
	Compress       bool
}

func DefaultConfig(token string) Config {
	return Config{
		Token:          token,
		Intents:        types.IntentsDefault,
		ShardID:        0,
		ShardCount:     1,
		HTTPTimeout:    10 * time.Second,
		MaxRetries:     3,
		LargeThreshold: 50,
		AutoDefer:      false,
	}
}

func (c *Config) Validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}
	if c.ShardCount <= 0 {
		c.ShardCount = 1
	}
	if c.ShardID < 0 || c.ShardID >= c.ShardCount {
		return ErrInvalidShard
	}
	if c.HTTPTimeout <= 0 {
		c.HTTPTimeout = 10 * time.Second
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3
	}
	c.LargeThreshold = min(max(c.LargeThreshold, 50), 250)
	if c.Intents == 0 {
		c.Intents = types.IntentsDefault
	}
	return nil
}

func (c Config) String() string {
	redacted := "[REDACTED]"
	if c.Token == "" {
		redacted = "[EMPTY]"
	}
	return fmt.Sprintf("Config{Token:%s, AppID:%s, Intents:%d, Shards:%d/%d}",
		redacted, c.AppID, c.Intents, c.ShardID, c.ShardCount)
}
