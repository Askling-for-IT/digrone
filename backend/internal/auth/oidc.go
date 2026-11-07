package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"digrone/internal/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Client struct {
	OAuth2   oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

func NewClient(ctx context.Context, cfg config.Config) (*Client, error) {
	if cfg.OIDCIssuer == "" || cfg.OIDCClientSecret == "" {
		return nil, errors.New("OIDC_ISSUER and OIDC_CLIENT_SECRET are required")
	}
	// The provider retains this HTTP client for discovery and signing-key requests.
	ctx = oidc.ClientContext(ctx, &http.Client{Timeout: 10 * time.Second})
	provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
	if err != nil {
		return nil, err
	}
	return &Client{
		OAuth2: oauth2.Config{
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			RedirectURL:  cfg.OIDCRedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		Verifier: provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID}),
	}, nil
}
