package gda

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sfossl/gda/gateway"
	"github.com/sfossl/gda/interaction"
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrMissingPublicKey = errors.New("gda: public key is required for webhook interactions")
)

type Option func(*Config)

func WithAppID(appID string) Option {
	return func(c *Config) {
		c.AppID = appID
	}
}

func WithPublicKey(publicKey string) Option {
	return func(c *Config) {
		c.PublicKey = publicKey
	}
}

func WithIntents(intents types.Intent) Option {
	return func(c *Config) {
		c.Intents = intents
	}
}

func WithShard(shardID, shardCount int) Option {
	return func(c *Config) {
		c.ShardID = shardID
		c.ShardCount = shardCount
	}
}

func WithHTTPTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.HTTPTimeout = d
	}
}

func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

func WithLargeThreshold(threshold int) Option {
	return func(c *Config) {
		c.LargeThreshold = threshold
	}
}

func WithAutoDefer(autoDefer bool) Option {
	return func(c *Config) {
		c.AutoDefer = autoDefer
	}
}

type Client struct {
	Config     Config
	REST       *rest.Client
	Gateway    *gateway.Client
	Router     *interaction.Router
	Dispatcher *gateway.EventDispatcher
}

func New(token string, opts ...Option) (*Client, error) {
	cfg := DefaultConfig(token)
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewWithConfig(cfg)
}

func NewWithConfig(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	restOpts := []rest.ClientOption{
		rest.WithMaxRetries(cfg.MaxRetries),
		rest.WithAppID(cfg.AppID),
	}
	if cfg.HTTPTimeout > 0 {
		restOpts = append(restOpts, rest.WithHTTPClient(&http.Client{
			Timeout: cfg.HTTPTimeout,
		}))
	}
	restClient := rest.NewClient(cfg.Token, restOpts...)

	dispatcher := gateway.NewEventDispatcher()
	gatewayClient := gateway.NewClient(gateway.Config{
		Token:          cfg.Token,
		Intents:        cfg.Intents,
		ShardID:        cfg.ShardID,
		ShardCount:     cfg.ShardCount,
		LargeThreshold: cfg.LargeThreshold,
	}, dispatcher)

	router := interaction.NewRouter()

	client := &Client{
		Config:     cfg,
		REST:       restClient,
		Gateway:    gatewayClient,
		Router:     router,
		Dispatcher: dispatcher,
	}

	dispatcher.OnInteractionCreate(func(ctx context.Context, in *types.Interaction) {
		ictx := interaction.NewGatewayContext(ctx, in, client.REST)
		_ = client.Router.Handle(ictx)
	})

	return client, nil
}

func (c *Client) Start(ctx context.Context) error {
	return c.Gateway.Connect(ctx)
}

func (c *Client) Close() error {
	return c.Gateway.Close()
}

func (c *Client) WebhookHandler() (http.Handler, error) {
	if c.Config.PublicKey == "" {
		return nil, ErrMissingPublicKey
	}

	pubKey, err := interaction.ParsePublicKey(c.Config.PublicKey)
	if err != nil {
		return nil, err
	}

	handler := c.Router.WebhookHandler(c.REST)
	return interaction.VerifyMiddleware(pubKey)(handler), nil
}

func (c *Client) OnCommand(name string, h interaction.Handler) {
	c.Router.OnCommand(name, h)
}

func (c *Client) OnSubcommand(command, subcommand string, h interaction.Handler) {
	c.Router.OnSubcommand(command, subcommand, h)
}

func (c *Client) OnComponent(customID string, h interaction.Handler) {
	c.Router.OnComponent(customID, h)
}

func (c *Client) OnModal(customID string, h interaction.Handler) {
	c.Router.OnModal(customID, h)
}

func (c *Client) OnMessageCreate(h func(context.Context, *types.Message)) {
	c.Dispatcher.OnMessageCreate(h)
}

func (c *Client) OnReady(h func(context.Context, *types.ReadyEvent)) {
	c.Dispatcher.OnReady(h)
}

func (c *Client) On(event types.EventName, h gateway.RawHandler) {
	c.Dispatcher.On(event, h)
}

func (c *Client) String() string {
	return c.Config.String()
}

