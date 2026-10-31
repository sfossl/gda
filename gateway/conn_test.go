package gateway

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestGatewayConnectAndIdentify(t *testing.T) {
	var identifyReceived sync.WaitGroup
	identifyReceived.Add(1)

	var readyHandled sync.WaitGroup
	readyHandled.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()

		hello := types.GatewayPayload{
			Op: types.OpcodeHello,
			D:  []byte(`{"heartbeat_interval":50}`),
		}
		helloData, _ := json.Marshal(hello)
		_ = sConn.WriteText(helloData)

		idMsg, err := sConn.ReadText()
		if err != nil {
			return
		}
		var idPayload types.GatewayPayload
		_ = json.Unmarshal(idMsg, &idPayload)
		if idPayload.Op == types.OpcodeIdentify {
			identifyReceived.Done()
		}

		ready := types.GatewayPayload{
			Op: types.OpcodeDispatch,
			S:  new(int64),
			T:  new(types.EventName),
			D:  []byte(`{"session_id":"sess_mock","resume_gateway_url":"wss://gateway.discord.gg","user":{"id":"1","username":"bot"}}`),
		}
		*ready.S = 1
		*ready.T = types.EventReady
		readyData, _ := json.Marshal(ready)
		_ = sConn.WriteText(readyData)

		time.Sleep(100 * time.Millisecond)
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	dispatcher := NewEventDispatcher()
	dispatcher.OnReady(func(ctx context.Context, e *types.ReadyEvent) {
		readyHandled.Done()
	})

	client := NewClient(Config{
		Token:      "token_123",
		GatewayURL: wsURL,
		Intents:    types.IntentsDefault,
	}, dispatcher)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = client.Connect(ctx)
	}()

	identifyReceived.Wait()
	readyHandled.Wait()

	if client.State().SessionID() != "sess_mock" {
		t.Fatalf("expected session sess_mock, got %s", client.State().SessionID())
	}
	_ = client.Close()
}

func TestGatewayResume(t *testing.T) {
	var resumeReceived sync.WaitGroup
	resumeReceived.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()

		hello := types.GatewayPayload{
			Op: types.OpcodeHello,
			D:  []byte(`{"heartbeat_interval":50}`),
		}
		helloData, _ := json.Marshal(hello)
		_ = sConn.WriteText(helloData)

		resumeMsg, err := sConn.ReadText()
		if err != nil {
			return
		}
		var p types.GatewayPayload
		_ = json.Unmarshal(resumeMsg, &p)
		if p.Op == types.OpcodeResume {
			resumeReceived.Done()
		}

		resumed := types.GatewayPayload{
			Op: types.OpcodeDispatch,
			S:  new(int64),
			T:  new(types.EventName),
			D:  []byte(`{}`),
		}
		*resumed.S = 2
		*resumed.T = types.EventResumed
		resumedData, _ := json.Marshal(resumed)
		_ = sConn.WriteText(resumedData)

		time.Sleep(100 * time.Millisecond)
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	dispatcher := NewEventDispatcher()
	client := NewClient(Config{
		Token:      "token_123",
		GatewayURL: wsURL,
	}, dispatcher)

	client.State().SetReady("sess_resumable", wsURL)
	client.State().SetSequence(1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = client.Connect(ctx)
	}()

	resumeReceived.Wait()

	if client.State().SessionID() != "sess_resumable" {
		t.Fatalf("expected session preserved")
	}
	_ = client.Close()
}

func TestGatewayCloseTerminatesConnect(t *testing.T) {
	var identifyReceived sync.WaitGroup
	identifyReceived.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()

		hello := types.GatewayPayload{
			Op: types.OpcodeHello,
			D:  []byte(`{"heartbeat_interval":50}`),
		}
		helloData, _ := json.Marshal(hello)
		_ = sConn.WriteText(helloData)
		_, _ = sConn.ReadText()
		identifyReceived.Done()
		time.Sleep(1 * time.Second)
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	client := NewClient(Config{
		Token:      "token_123",
		GatewayURL: wsURL,
	}, NewEventDispatcher())

	done := make(chan error, 1)
	go func() {
		done <- client.Connect(context.Background())
	}()

	identifyReceived.Wait()
	_ = client.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil error on clean close, got: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Connect did not exit on Close()")
	}
}

func TestGatewayCancelBeforeHelloTerminatesConnect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()
		time.Sleep(2 * time.Second)
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	client := NewClient(Config{
		Token:      "token_123",
		GatewayURL: wsURL,
	}, NewEventDispatcher())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- client.Connect(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Connect did not exit on ctx cancel before Hello")
	}
}

func TestGatewayZlibCompression(t *testing.T) {
	var identifyReceived sync.WaitGroup
	identifyReceived.Add(1)

	var readyHandled sync.WaitGroup
	readyHandled.Add(1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "compress=zlib-stream") {
			t.Errorf("expected compress=zlib-stream in query, got: %s", r.URL.RawQuery)
		}

		sConn, err := acceptMockWS(w, r)
		if err != nil {
			return
		}
		defer sConn.Close()

		var buf bytes.Buffer
		zw, _ := zlib.NewWriterLevel(&buf, zlib.DefaultCompression)

		hello := types.GatewayPayload{
			Op: types.OpcodeHello,
			D:  []byte(`{"heartbeat_interval":50}`),
		}
		helloData, _ := json.Marshal(hello)
		_, _ = zw.Write(helloData)
		_ = zw.Flush()

		_ = sConn.WriteBinary(buf.Bytes())
		buf.Reset()

		idMsg, err := sConn.ReadText()
		if err != nil {
			return
		}
		var idPayload types.GatewayPayload
		_ = json.Unmarshal(idMsg, &idPayload)
		if idPayload.Op == types.OpcodeIdentify {
			identifyReceived.Done()
		}

		ready := types.GatewayPayload{
			Op: types.OpcodeDispatch,
			S:  new(int64),
			T:  new(types.EventName),
			D:  []byte(`{"session_id":"sess_zlib","resume_gateway_url":"wss://gateway.discord.gg","user":{"id":"1","username":"bot"}}`),
		}
		*ready.S = 1
		*ready.T = types.EventReady
		readyData, _ := json.Marshal(ready)
		_, _ = zw.Write(readyData)
		_ = zw.Flush()

		_ = sConn.WriteBinary(buf.Bytes())
		_ = zw.Close()

		time.Sleep(100 * time.Millisecond)
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	dispatcher := NewEventDispatcher()
	dispatcher.OnReady(func(ctx context.Context, r *types.ReadyEvent) {
		if r.SessionID == "sess_zlib" {
			readyHandled.Done()
		}
	})

	client := NewClient(Config{
		Token:      "token_123",
		GatewayURL: wsURL,
		Compress:   true,
	}, dispatcher)

	go func() {
		_ = client.Connect(context.Background())
	}()

	identifyReceived.Wait()
	readyHandled.Wait()

	if client.State().SessionID() != "sess_zlib" {
		t.Fatalf("expected session ID sess_zlib, got: %s", client.State().SessionID())
	}
	_ = client.Close()
}

func TestGatewayClient_String(t *testing.T) {
	var nilClient *Client
	if nilClient.String() != "<nil>" {
		t.Fatalf("expected <nil>, got %s", nilClient.String())
	}

	emptyClient := NewClient(Config{}, NewEventDispatcher())
	if !strings.Contains(emptyClient.String(), "[EMPTY]") {
		t.Fatalf("expected [EMPTY], got %s", emptyClient.String())
	}

	client := NewClient(Config{
		Token:      "super-secret-gateway-token",
		ShardID:    1,
		ShardCount: 4,
	}, NewEventDispatcher())

	str := client.String()
	if strings.Contains(str, "super-secret-gateway-token") {
		t.Fatalf("gateway token leaked in String(): %s", str)
	}
	if !strings.Contains(str, "[REDACTED]") {
		t.Fatalf("missing [REDACTED]: %s", str)
	}
	if !strings.Contains(str, "1/4") {
		t.Fatalf("missing shard info: %s", str)
	}
}

func TestGatewayResumptionURLConstruction(t *testing.T) {
	client := NewClient(Config{
		Token:    "tok",
		Compress: true,
	}, NewEventDispatcher())

	client.State().SetReady("sess", "wss://resume.discord.gg")
	client.State().SetSequence(42)

	endpointURL := client.cfg.GatewayURL
	if client.state.CanResume() && client.state.ResumeGatewayURL() != "" {
		endpointURL = client.state.ResumeGatewayURL()
	}
	u, err := url.Parse(endpointURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	q := u.Query()
	q.Set("v", types.APIVersion)
	q.Set("encoding", "json")
	if client.cfg.Compress {
		q.Set("compress", "zlib-stream")
	}
	u.RawQuery = q.Encode()

	finalURL := u.String()
	if !strings.HasPrefix(finalURL, "wss://resume.discord.gg?") {
		t.Fatalf("unexpected prefix: %s", finalURL)
	}
	if !strings.Contains(finalURL, "compress=zlib-stream") {
		t.Fatalf("missing compress param: %s", finalURL)
	}
	if !strings.Contains(finalURL, "v=10") {
		t.Fatalf("missing v param: %s", finalURL)
	}
}

func TestSendHeartbeat_PayloadStructure(t *testing.T) {
	var seq int64 = 42
	var d json.RawMessage
	if &seq != nil {
		d = []byte(strconv.FormatInt(seq, 10))
	}
	p := types.GatewayPayload{
		Op: types.OpcodeHeartbeat,
		D:  d,
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(b), `"op":1`) {
		t.Fatalf("missing op 1: %s", string(b))
	}
	if !strings.Contains(string(b), `"d":42`) {
		t.Fatalf("missing d 42: %s", string(b))
	}

	pNull := types.GatewayPayload{
		Op: types.OpcodeHeartbeat,
		D:  []byte("null"),
	}
	bNull, _ := json.Marshal(pNull)
	if !strings.Contains(string(bNull), `"d":null`) {
		t.Fatalf("missing d null: %s", string(bNull))
	}
}
