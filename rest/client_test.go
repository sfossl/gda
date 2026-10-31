package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestClientSendMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot test-token" {
			t.Errorf("missing or bad authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") != types.UserAgent {
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

func TestClientTriggerTypingAndCommands(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/channels/555/typing":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "PUT" && r.URL.Path == "/applications/777/commands":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{
				{Name: "ping"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	err := client.TriggerTyping(context.Background(), types.Snowflake(555))
	if err != nil {
		t.Fatalf("unexpected typing error: %v", err)
	}

	cmds, err := client.BulkOverwriteGlobalCommands(context.Background(), types.Snowflake(777), []types.ApplicationCommand{
		{Name: "ping"},
	})
	if err != nil {
		t.Fatalf("unexpected bulk overwrite error: %v", err)
	}
	if len(cmds) != 1 || cmds[0].Name != "ping" {
		t.Fatalf("unexpected commands returned: %+v", cmds)
	}
}

func TestClientRateLimitResponseBody(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"You are being rate limited.","retry_after":0.05,"global":false}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"999"}`))
	}))
	defer ts.Close()

	client := NewClient("test-token", WithBaseURL(ts.URL))
	start := time.Now()
	_, err := client.Do(context.Background(), Request{
		Endpoint: EndpointGetChannel(types.Snowflake(999)),
	})
	if err != nil {
		t.Fatalf("unexpected error on retry: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatalf("expected backoff delay from body >= 40ms, got %v", time.Since(start))
	}
}

func TestSendMessageWithFiles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Fatalf("expected multipart content type, got: %s", contentType)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		payloadJSON := r.FormValue("payload_json")
		if !strings.Contains(payloadJSON, "test content") {
			t.Fatalf("missing payload_json content, got: %s", payloadJSON)
		}
		file, header, err := r.FormFile("files[0]")
		if err != nil {
			t.Fatalf("failed to read form file: %v", err)
		}
		defer file.Close()
		if header.Filename != "test.txt" {
			t.Fatalf("expected filename test.txt, got %s", header.Filename)
		}
		fileData, _ := io.ReadAll(file)
		if string(fileData) != "sample file data" {
			t.Fatalf("unexpected file data: %s", string(fileData))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"101","channel_id":"202","content":"test content","author":{"id":"1","username":"b","discriminator":"0"}}`))
	}))
	defer ts.Close()

	client := NewClient("token", WithBaseURL(ts.URL))
	msg, err := client.SendMessage(context.Background(), types.Snowflake(202), MessagePayload{
		Content: "test content",
		Files: []types.File{
			{
				Name:   "test.txt",
				Reader: strings.NewReader("sample file data"),
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != types.Snowflake(101) {
		t.Fatalf("unexpected message id: %s", msg.ID)
	}
}

func TestClient_StringRedaction(t *testing.T) {
	var nilClient *Client
	if nilClient.String() != "<nil>" {
		t.Fatalf("expected <nil>, got %s", nilClient.String())
	}

	emptyClient := NewClient("")
	if !strings.Contains(emptyClient.String(), "[EMPTY]") {
		t.Fatalf("expected [EMPTY], got %s", emptyClient.String())
	}

	client := NewClient("super-secret-token", WithAppID("app-999"))
	str := client.String()
	if strings.Contains(str, "super-secret-token") {
		t.Fatalf("leaked token in string: %s", str)
	}
	if !strings.Contains(str, "[REDACTED]") {
		t.Fatalf("missing [REDACTED]: %s", str)
	}
}

func TestClient_ProtectedHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot legitimate-token" {
			t.Fatalf("authorization header was compromised: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") != types.UserAgent {
			t.Fatalf("user agent header was compromised: %s", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("X-Audit-Log-Reason") != "valid moderation action" {
			t.Fatalf("custom audit header missing: %s", r.Header.Get("X-Audit-Log-Reason"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClient("legitimate-token", WithBaseURL(ts.URL))
	headers := make(http.Header)
	headers.Set("Authorization", "Malicious Injected Header")
	headers.Set("User-Agent", "Malicious UA")
	headers.Set("X-Audit-Log-Reason", "valid moderation action")

	_, err := client.Do(context.Background(), Request{
		Endpoint: Endpoint{Method: "POST", Path: "/test", Key: "POST /test"},
		Headers:  headers,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_FilePayloadTooLarge(t *testing.T) {
	client := NewClient("token")
	giantReader := io.LimitReader(strings.NewReader(strings.Repeat("a", 1024)), 11<<20)

	zeros := io.MultiReader(
		bytes.NewReader(make([]byte, 5<<20)),
		bytes.NewReader(make([]byte, 6<<20)),
	)
	_ = giantReader

	_, err := client.Do(context.Background(), Request{
		Endpoint: Endpoint{Method: "POST", Path: "/test", Key: "POST /test"},
		Files: []types.File{
			{Name: "huge.bin", Reader: zeros},
		},
	})

	if err != ErrPayloadTooLarge {
		t.Fatalf("expected ErrPayloadTooLarge, got %v", err)
	}
}
