package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
