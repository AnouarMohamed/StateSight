package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOIDCAuthenticatorVerifiesMappedIdentityClaims(t *testing.T) {
	provider, privateKey := newTestOIDCProvider(t)
	authenticator, err := NewOIDCAuthenticator(context.Background(), provider.URL, "statesight-api", true)
	if err != nil {
		t.Fatalf("create authenticator: %v", err)
	}

	token := signTestToken(t, privateKey, provider.URL, "statesight-api", "operator-42", "operator@example.test")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	identity, err := authenticator.Authenticate(context.Background(), req)
	if err != nil {
		t.Fatalf("authenticate request: %v", err)
	}
	if identity.Issuer != provider.URL || identity.Subject != "operator-42" || identity.Email != "operator@example.test" {
		t.Fatalf("unexpected verified identity: %#v", identity)
	}
}

func TestOIDCAuthenticatorRejectsWrongAudience(t *testing.T) {
	provider, privateKey := newTestOIDCProvider(t)
	authenticator, err := NewOIDCAuthenticator(context.Background(), provider.URL, "statesight-api", true)
	if err != nil {
		t.Fatalf("create authenticator: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, privateKey, provider.URL, "another-api", "operator-42", ""))

	if _, err := authenticator.Authenticate(context.Background(), req); !errors.Is(err, ErrInvalidBearerToken) {
		t.Fatalf("expected invalid bearer token error, got %v", err)
	}
}

func TestOIDCAuthenticatorRequiresBearerToken(t *testing.T) {
	provider, _ := newTestOIDCProvider(t)
	authenticator, err := NewOIDCAuthenticator(context.Background(), provider.URL, "statesight-api", true)
	if err != nil {
		t.Fatalf("create authenticator: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req.Header.Set("Authorization", "Basic not-a-token")

	if _, err := authenticator.Authenticate(context.Background(), req); !errors.Is(err, ErrMissingBearerToken) {
		t.Fatalf("expected missing bearer token error, got %v", err)
	}
}

func TestValidateOIDCEndpointRejectsInsecureJWKSWithoutLocalOverride(t *testing.T) {
	if err := validateOIDCEndpoint("JWKS", "http://identity.example.test/keys", false); err == nil {
		t.Fatal("expected plain-HTTP JWKS endpoint to be rejected")
	}
	if err := validateOIDCEndpoint("JWKS", "http://127.0.0.1:5556/keys", true); err != nil {
		t.Fatalf("allow explicitly enabled local JWKS endpoint: %v", err)
	}
}

func newTestOIDCProvider(t *testing.T) (*httptest.Server, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}

	var provider *httptest.Server
	handler := http.NewServeMux()
	handler.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		writeTestJSON(t, w, map[string]any{
			"issuer":                                provider.URL,
			"authorization_endpoint":                provider.URL + "/authorize",
			"token_endpoint":                        provider.URL + "/token",
			"jwks_uri":                              provider.URL + "/keys",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	handler.HandleFunc("/keys", func(w http.ResponseWriter, _ *http.Request) {
		writeTestJSON(t, w, map[string]any{"keys": []map[string]string{{
			"kty": "RSA",
			"kid": "test-signing-key",
			"use": "sig",
			"alg": "RS256",
			"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
		}}})
	})
	provider = httptest.NewServer(handler)
	t.Cleanup(provider.Close)
	return provider, privateKey
}

func signTestToken(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, subject, email string) string {
	t.Helper()

	header := encodeTestJWTPart(t, map[string]string{"alg": "RS256", "kid": "test-signing-key", "typ": "JWT"})
	claims := encodeTestJWTPart(t, map[string]any{
		"iss":   issuer,
		"sub":   subject,
		"aud":   audience,
		"iat":   time.Now().Add(-time.Minute).Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
		"email": email,
	})
	unsigned := header + "." + claims
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func encodeTestJWTPart(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal token part: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write test provider response: %v", err)
	}
}
