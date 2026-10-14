package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sfossl/gda/types"
)

type Client struct {
	token      string
	appID      string
	httpClient *http.Client
	limiter    *RateLimiter
	maxRetries int
	baseURL    string
}

type ClientOption func(*Client)

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

func WithRateLimiter(l *RateLimiter) ClientOption {
	return func(c *Client) {
		if l != nil {
			c.limiter = l
		}
	}
}

func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		if url != "" {
			c.baseURL = url
		}
	}
}

func WithMaxRetries(retries int) ClientOption {
	return func(c *Client) {
		if retries > 0 {
			c.maxRetries = retries
		}
	}
}

func WithAppID(appID string) ClientOption {
	return func(c *Client) {
		c.appID = appID
	}
}

func NewClient(token string, opts ...ClientOption) *Client {
	c := &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		limiter:    NewRateLimiter(),
		maxRetries: 3,
		baseURL:    BaseURL,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) CloseIdleConnections() {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
}

type Request struct {
	Endpoint Endpoint
	Body     any
	Headers  http.Header
}

type MessagePayload struct {
	Content    string            `json:"content,omitempty"`
	Embeds     []types.Embed     `json:"embeds,omitempty"`
	Components []types.Component `json:"components,omitempty"`
	Flags      types.MessageFlag `json:"flags,omitempty"`
	TTS        bool              `json:"tts,omitempty"`
}

func (c *Client) Do(ctx context.Context, req Request) ([]byte, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.limiter.Acquire(ctx, req.Endpoint.Key); err != nil {
			return nil, err
		}

		var reader io.Reader
		if bodyBytes != nil {
			reader = bytes.NewReader(bodyBytes)
		}

		url := c.baseURL + req.Endpoint.Path
		httpReq, err := http.NewRequestWithContext(ctx, req.Endpoint.Method, url, reader)
		if err != nil {
			return nil, err
		}

		if c.token != "" {
			httpReq.Header.Set("Authorization", "Bot "+c.token)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", types.UserAgent)

		for k, values := range req.Headers {
			for _, v := range values {
				httpReq.Header.Add(k, v)
			}
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if attempt < c.maxRetries {
				continue
			}
			return nil, err
		}

		c.limiter.Update(req.Endpoint.Key, resp.Header)

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, types.MaxPayloadSize))
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			var rateLimitResp struct {
				RetryAfter float64 `json:"retry_after"`
				Global     bool    `json:"global"`
			}
			if err := json.Unmarshal(respBody, &rateLimitResp); err == nil && rateLimitResp.RetryAfter > 0 {
				c.limiter.UpdateWithRetryAfter(req.Endpoint.Key, rateLimitResp.RetryAfter, rateLimitResp.Global)
			}
			if attempt < c.maxRetries {
				continue
			}
		}

		if resp.StatusCode >= 500 && attempt < c.maxRetries {
			select {
			case <-time.After(time.Duration(1<<attempt) * 50 * time.Millisecond):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}

		if resp.StatusCode >= 400 {
			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Code != 0 {
				apiErr.HTTPStatus = resp.StatusCode
				return nil, &apiErr
			}
			return nil, &APIError{
				HTTPStatus: resp.StatusCode,
				Message:    fmt.Sprintf("http error status %d: %s", resp.StatusCode, string(respBody)),
			}
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("discord rest: max retries reached for %s", req.Endpoint.Key)
}

func (c *Client) CreateDM(ctx context.Context, recipientID types.Snowflake) (*types.Channel, error) {
	payload := struct {
		RecipientID types.Snowflake `json:"recipient_id"`
	}{
		RecipientID: recipientID,
	}

	data, err := c.Do(ctx, Request{
		Endpoint: EndpointCreateDM(),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}

	var ch types.Channel
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (c *Client) SendMessage(ctx context.Context, channelID types.Snowflake, payload MessagePayload) (*types.Message, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointSendMessage(channelID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}

	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) EditMessage(ctx context.Context, channelID, messageID types.Snowflake, payload MessagePayload) (*types.Message, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointEditMessage(channelID, messageID),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}

	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) DeleteMessage(ctx context.Context, channelID, messageID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteMessage(channelID, messageID),
	})
	return err
}

func (c *Client) CreateInteractionResponse(ctx context.Context, interactionID types.Snowflake, token string, response types.InteractionResponse) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointInteractionCallback(interactionID, token),
		Body:     response,
	})
	return err
}

func (c *Client) CreateFollowupMessage(ctx context.Context, appID types.Snowflake, token string, payload MessagePayload) (*types.Message, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointFollowupMessage(appID, token),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}

	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) EditOriginalInteractionResponse(ctx context.Context, appID types.Snowflake, token string, payload MessagePayload) (*types.Message, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointEditOriginalInteraction(appID, token),
		Body:     payload,
	})
	if err != nil {
		return nil, err
	}

	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) DeleteOriginalInteractionResponse(ctx context.Context, appID types.Snowflake, token string) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointDeleteOriginalInteraction(appID, token),
	})
	return err
}

func (c *Client) TriggerTyping(ctx context.Context, channelID types.Snowflake) error {
	_, err := c.Do(ctx, Request{
		Endpoint: EndpointTriggerTyping(channelID),
	})
	return err
}

func (c *Client) BulkOverwriteGlobalCommands(ctx context.Context, appID types.Snowflake, commands []types.ApplicationCommand) ([]types.ApplicationCommand, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointBulkOverwriteGlobalCommands(appID),
		Body:     commands,
	})
	if err != nil {
		return nil, err
	}

	var registered []types.ApplicationCommand
	if err := json.Unmarshal(data, &registered); err != nil {
		return nil, err
	}
	return registered, nil
}

func (c *Client) BulkOverwriteGuildCommands(ctx context.Context, appID, guildID types.Snowflake, commands []types.ApplicationCommand) ([]types.ApplicationCommand, error) {
	data, err := c.Do(ctx, Request{
		Endpoint: EndpointBulkOverwriteGuildCommands(appID, guildID),
		Body:     commands,
	})
	if err != nil {
		return nil, err
	}

	var registered []types.ApplicationCommand
	if err := json.Unmarshal(data, &registered); err != nil {
		return nil, err
	}
	return registered, nil
}


