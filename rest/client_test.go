package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/sfossl/bfrm/types"
)

func TestClientSendMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot test-token" {
			t.Errorf("missing or bad authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") != "DiscordBot (github.com/sfossl/bfrm, v0.1.0)" {
			t.Errorf("unexpected user agent: %s", r.Header.Get("User-Agent"))
		}
		if r.Method != "POST" || r.URL.Path != "/channels/123/messages" {
			t.Errorf("unexpected path: %s %s", r.Method, r.URL.Path)
		}

		resp := types.Message{
			ID:        types.Snowflake(456),
			ChannelID: types.Snowflake(123),
			Content:   "pong",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	msg, err := client.SendMessage(context.Background(), types.Snowflake(123), MessagePayload{
		Content: "ping",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != types.Snowflake(456) || msg.Content != "pong" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestClientRateLimitRetry(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := atomic.AddInt32(&attempts, 1)
		if att == 1 {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset-After", "0.05")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"You are being rate limited.","retry_after":0.05,"global":false}`))
			return
		}

		resp := types.Message{
			ID:      types.Snowflake(789),
			Content: "success after rate limit",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	msg, err := client.SendMessage(context.Background(), types.Snowflake(100), MessagePayload{
		Content: "retry test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content != "success after rate limit" {
		t.Fatalf("unexpected response: %s", msg.Content)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestClientAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":10003,"message":"Unknown Channel"}`))
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	_, err := client.SendMessage(context.Background(), types.Snowflake(999), MessagePayload{
		Content: "hello",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got %v", err)
	}
}

func TestClientCreateDM(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/users/@me/channels" {
			t.Errorf("unexpected path: %s %s", r.Method, r.URL.Path)
		}
		ch := types.Channel{
			ID:   types.Snowflake(111222),
			Type: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ch)
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	ch, err := client.CreateDM(context.Background(), types.Snowflake(333444))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.ID != types.Snowflake(111222) || ch.Type != 1 {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}

