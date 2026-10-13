package bfrm

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/bfrm/interaction"
	"github.com/sfossl/bfrm/types"
)

func TestClientOptions(t *testing.T) {
	client, err := New("test-token",
		WithAppID("app-123"),
		WithPublicKey("pub-key-hex"),
		WithIntents(types.IntentGuildMessages),
		WithShard(2, 5),
		WithHTTPTimeout(5*time.Second),
		WithMaxRetries(5),
		WithLargeThreshold(100),
		WithAutoDefer(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Config.AppID != "app-123" {
		t.Fatalf("expected app id app-123, got %s", client.Config.AppID)
	}
	if client.Config.PublicKey != "pub-key-hex" {
		t.Fatalf("expected public key pub-key-hex, got %s", client.Config.PublicKey)
	}
	if client.Config.Intents != types.IntentGuildMessages {
		t.Fatalf("expected intent GuildMessages, got %d", client.Config.Intents)
	}
	if client.Config.ShardID != 2 || client.Config.ShardCount != 5 {
		t.Fatalf("expected shard 2/5, got %d/%d", client.Config.ShardID, client.Config.ShardCount)
	}
	if client.Config.HTTPTimeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", client.Config.HTTPTimeout)
	}
	if client.Config.MaxRetries != 5 {
		t.Fatalf("expected max retries 5, got %d", client.Config.MaxRetries)
	}
	if client.Config.LargeThreshold != 100 {
		t.Fatalf("expected large threshold 100, got %d", client.Config.LargeThreshold)
	}
	if !client.Config.AutoDefer {
		t.Fatalf("expected auto defer true")
	}
}

func TestClientValidation(t *testing.T) {
	_, err := New("")
	if err != ErrMissingToken {
		t.Fatalf("expected ErrMissingToken, got %v", err)
	}

	_, err = New("token", WithShard(2, 2))
	if err != ErrInvalidShard {
		t.Fatalf("expected ErrInvalidShard, got %v", err)
	}
}

func TestClientGatewayInteractionBridge(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var called atomic.Bool
	client.OnCommand("ping", func(ctx *interaction.Context) error {
		called.Store(true)
		return nil
	})

	in := &types.Interaction{
		ID:   types.Snowflake(1),
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "ping",
		},
	}

	client.Dispatcher.Dispatch(context.Background(), types.EventInteractionCreate, []byte(`{"id":"1","type":2,"data":{"name":"ping"}}`))

	for i := 0; i < 50; i++ {
		if called.Load() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !called.Load() {
		t.Fatalf("expected command handler to be called via gateway interaction bridge")
	}

	_ = in
}

func TestClientWebhookHandler(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.WebhookHandler()
	if err != ErrMissingPublicKey {
		t.Fatalf("expected ErrMissingPublicKey, got %v", err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}
	pubHex := hex.EncodeToString(pub)

	clientWithKey, err := New("test-token", WithPublicKey(pubHex))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler, err := clientWithKey.WebhookHandler()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("POST", "/interactions", bytes.NewReader([]byte(`{"type":1}`)))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without signature, got %d", w.Code)
	}

	ts := "1700000000"
	body := []byte(`{"type":1}`)
	sigMsg := append([]byte(ts), body...)
	sig := ed25519.Sign(priv, sigMsg)

	reqValid := httptest.NewRequest("POST", "/interactions", bytes.NewReader(body))
	reqValid.Header.Set("X-Signature-Ed25519", hex.EncodeToString(sig))
	reqValid.Header.Set("X-Signature-Timestamp", ts)
	wValid := httptest.NewRecorder()
	handler.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid ping, got %d", wValid.Code)
	}
	if !bytes.Contains(wValid.Body.Bytes(), []byte(`"type":1`)) {
		t.Fatalf("expected ping response pong, got %s", wValid.Body.String())
	}
}

func TestClientLifecycleAndShortcuts(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var readyCalled atomic.Bool
	client.OnReady(func(ctx context.Context, r *types.ReadyEvent) {
		readyCalled.Store(true)
	})

	var msgCalled atomic.Bool
	client.OnMessageCreate(func(ctx context.Context, m *types.Message) {
		msgCalled.Store(true)
	})

	client.Dispatcher.Dispatch(context.Background(), types.EventReady, []byte(`{"v":10,"session_id":"s","resume_gateway_url":"wss://g","user":{"id":"1","username":"bot","discriminator":"0000"},"guilds":[]}`))
	client.Dispatcher.Dispatch(context.Background(), types.EventMessageCreate, []byte(`{"id":"10","channel_id":"20","content":"hello","author":{"id":"1","username":"u","discriminator":"0000"}}`))

	for i := 0; i < 50; i++ {
		if readyCalled.Load() && msgCalled.Load() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !readyCalled.Load() {
		t.Fatalf("expected OnReady callback")
	}
	if !msgCalled.Load() {
		t.Fatalf("expected OnMessageCreate callback")
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("unexpected error closing client: %v", err)
	}

	str := client.String()
	if str == "" || bytes.Contains([]byte(str), []byte("test-token")) {
		t.Fatalf("token must be redacted in String(), got %s", str)
	}
}

