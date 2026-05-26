package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

var (
	ErrMissingBearerToken = errors.New("authorization bearer token is required")
	ErrInvalidBearerToken = errors.New("authorization bearer token is invalid")
)

type Authenticator interface {
	Authenticate(ctx context.Context, r *http.Request) (Identity, error)
}

type OIDCAuthenticator struct {
	verifier *oidc.IDTokenVerifier
}

func NewOIDCAuthenticator(ctx context.Context, issuerURL, audience string, allowInsecure bool) (*OIDCAuthenticator, error) {
	if err := validateOIDCEndpoint("issuer", issuerURL, allowInsecure); err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	var metadata struct {
		JWKSURL string `json:"jwks_uri"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("read OIDC discovery claims: %w", err)
	}
	if err := validateOIDCEndpoint("JWKS", metadata.JWKSURL, allowInsecure); err != nil {
		return nil, err
	}

	return &OIDCAuthenticator{
		verifier: provider.Verifier(&oidc.Config{ClientID: audience}),
	}, nil
}

func (a *OIDCAuthenticator) Authenticate(ctx context.Context, r *http.Request) (Identity, error) {
	rawToken, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return Identity{}, err
	}

	token, err := a.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrInvalidBearerToken, err)
	}
	if strings.TrimSpace(token.Subject) == "" || strings.TrimSpace(token.Issuer) == "" {
		return Identity{}, ErrInvalidBearerToken
	}

	var claims struct {
		Email string `json:"email"`
	}
	if err := token.Claims(&claims); err != nil {
		return Identity{}, fmt.Errorf("%w: parse claims: %v", ErrInvalidBearerToken, err)
	}

	return Identity{
		Issuer:  token.Issuer,
		Subject: token.Subject,
		Email:   strings.TrimSpace(claims.Email),
	}, nil
}

func bearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", ErrMissingBearerToken
	}
	return parts[1], nil
}

func validateOIDCEndpoint(name, rawURL string, allowInsecure bool) error {
	endpoint, err := url.Parse(rawURL)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") {
		return fmt.Errorf("OIDC %s URL must be a valid HTTP(S) URL", name)
	}
	if endpoint.Scheme != "https" && !allowInsecure {
		return fmt.Errorf("OIDC %s URL must use HTTPS unless insecure local OIDC is explicitly enabled", name)
	}
	return nil
}
