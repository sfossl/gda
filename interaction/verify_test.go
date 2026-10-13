package interaction

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestVerifySignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	ts := "1600000000"
	body := []byte(`{"type":1}`)
	msg := append([]byte(ts), body...)
	sig := ed25519.Sign(priv, msg)
	sigHex := hex.EncodeToString(sig)

	if !VerifySignature(pub, sigHex, ts, body) {
		t.Fatalf("expected signature to verify")
	}

	tamperedBody := []byte(`{"type":2}`)
	if VerifySignature(pub, sigHex, ts, tamperedBody) {
		t.Fatalf("expected verification to fail with tampered body")
	}

	if VerifySignature(pub, sigHex, "1600000001", body) {
		t.Fatalf("expected verification to fail with tampered timestamp")
	}

	if VerifySignature(pub, "invalid_hex", ts, body) {
		t.Fatalf("expected verification to fail with invalid hex")
	}
}

func TestParsePublicKey(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubHex := hex.EncodeToString(pub)

	parsed, err := ParsePublicKey(pubHex)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}
	if !bytes.Equal(pub, parsed) {
		t.Fatalf("parsed key mismatch")
	}

	_, err = ParsePublicKey("short")
	if err != ErrInvalidPublicKey {
		t.Fatalf("expected ErrInvalidPublicKey, got %v", err)
	}
}

func TestVerifyMiddleware(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"type":1}`)
	msg := append([]byte(ts), body...)
	sigHex := hex.EncodeToString(ed25519.Sign(priv, msg))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := VerifyMiddleware(pub)(nextHandler)

	req := httptest.NewRequest("POST", "/interactions", bytes.NewReader(body))
	req.Header.Set(HeaderSignature, sigHex)
	req.Header.Set(HeaderTimestamp, ts)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	badReq := httptest.NewRequest("POST", "/interactions", bytes.NewReader(body))
	badReq.Header.Set(HeaderSignature, "bad_sig")
	badReq.Header.Set(HeaderTimestamp, ts)
	badRR := httptest.NewRecorder()
	handler.ServeHTTP(badRR, badReq)
	if badRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", badRR.Code)
	}

	noHeadersReq := httptest.NewRequest("POST", "/interactions", bytes.NewReader(body))
	noHeadersRR := httptest.NewRecorder()
	handler.ServeHTTP(noHeadersRR, noHeadersReq)
	if noHeadersRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", noHeadersRR.Code)
	}

	oldTs := "1600000000"
	oldMsg := append([]byte(oldTs), body...)
	oldSig := hex.EncodeToString(ed25519.Sign(priv, oldMsg))
	expiredReq := httptest.NewRequest("POST", "/interactions", bytes.NewReader(body))
	expiredReq.Header.Set(HeaderSignature, oldSig)
	expiredReq.Header.Set(HeaderTimestamp, oldTs)
	expiredRR := httptest.NewRecorder()
	handler.ServeHTTP(expiredRR, expiredReq)
	if expiredRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for expired timestamp, got %d", expiredRR.Code)
	}
}

