package interaction

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/sfossl/gda/types"
)

var (
	ErrInvalidPublicKey = errors.New("gda/interaction: invalid ed25519 public key")
	ErrInvalidSignature = errors.New("gda/interaction: invalid ed25519 signature")
)

const (
	HeaderSignature = "X-Signature-Ed25519"
	HeaderTimestamp = "X-Signature-Timestamp"
)

func ParsePublicKey(publicKeyHex string) (ed25519.PublicKey, error) {
	keyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(keyBytes) != ed25519.PublicKeySize {
		return nil, ErrInvalidPublicKey
	}
	return ed25519.PublicKey(keyBytes), nil
}

func VerifySignature(publicKey ed25519.PublicKey, signatureHex, timestamp string, body []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}

	sig, err := hex.DecodeString(signatureHex)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}

	msg := make([]byte, 0, len(timestamp)+len(body))
	msg = append(msg, timestamp...)
	msg = append(msg, body...)

	return ed25519.Verify(publicKey, msg, sig)
}

func VerifyMiddleware(publicKey ed25519.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sig := r.Header.Get(HeaderSignature)
			ts := r.Header.Get(HeaderTimestamp)
			if sig == "" || ts == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			body, err := io.ReadAll(io.LimitReader(r.Body, types.MaxPayloadSize))
			_ = r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if !VerifySignature(publicKey, sig, ts, body) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
