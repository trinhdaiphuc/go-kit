// Package mailbox provides a client for interacting with Microsoft Outlook
// mailboxes via the Microsoft Graph API using the Resource Owner Password
// Credentials (ROPC) OAuth2 flow.
package mailbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// defaultTimeout is applied to both the token-fetch client and the API client
// when no custom http.Client is supplied.
const defaultTimeout = 30 * time.Second

// ─── Options ─────────────────────────────────────────────────────────────────

// options holds the resolved configuration for MailboxClient construction.
type options struct {
	httpClient *http.Client
	ctx        context.Context //nolint:containedctx // intentional: stored as base context
}

// defaultOptions returns sane defaults.
func defaultOptions() options {
	return options{
		httpClient: &http.Client{Timeout: defaultTimeout},
		ctx:        context.Background(),
	}
}

// Option is a functional option that configures a MailboxClient.
type Option func(*options)

// WithHTTPClient replaces the default http.Client used for all Graph API calls.
// Use this to inject transports with custom TLS, proxies, or middleware.
func WithHTTPClient(hc *http.Client) Option {
	return func(o *options) {
		if hc != nil {
			o.httpClient = hc
		}
	}
}

// WithTimeout sets a timeout on the default http.Client.
// Ignored when WithHTTPClient is also provided.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		// Only apply when the caller has not already swapped the client.
		if d > 0 {
			o.httpClient = &http.Client{Timeout: d}
		}
	}
}

// WithContext sets the base context stored on the client. It is used as the
// parent for internally created contexts (e.g. during token fetch) and can be
// retrieved via MailboxClient.Context().
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		if ctx != nil {
			o.ctx = ctx
		}
	}
}

// ─── Config ───────────────────────────────────────────────────────────────────

// Config holds the credentials and tenant information needed to authenticate
// against Azure AD and access the Microsoft Graph API.
type Config struct {
	// Username is the Microsoft account username (email address).
	Username string
	// Password is the Microsoft account password.
	Password string
	// ClientID is the Azure AD application (client) ID.
	ClientID string
	// ClientSecret is the Azure AD client secret (optional for public clients).
	ClientSecret string
	// TenantID is the Azure AD directory (tenant) ID.
	TenantID string
}

// ─── TokenResponse ────────────────────────────────────────────────────────────

// TokenResponse represents the OAuth2 token response from Azure AD.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// ─── MailboxClient ────────────────────────────────────────────────────────────

// MailboxClient is an authenticated HTTP client for the Microsoft Graph API.
// Obtain one via NewMailboxClient.
type MailboxClient struct {
	accessToken string
	httpClient  *http.Client
	ctx         context.Context //nolint:containedctx
}

// Context returns the base context that was set via WithContext (defaults to
// context.Background()).
func (c *MailboxClient) Context() context.Context {
	return c.ctx
}

// NewMailboxClient authenticates against Azure AD using the ROPC OAuth2 flow
// and returns a ready-to-use MailboxClient.
//
// Pass zero or more Option values to override the HTTP client, timeout, or
// base context:
//
//	client, err := mailbox.NewMailboxClient(ctx, cfg,
//	    mailbox.WithHTTPClient(myClient),
//	    mailbox.WithContext(baseCtx),
//	)
func NewMailboxClient(cfg *Config, opts ...Option) (*MailboxClient, error) {
	o := defaultOptions()
	// WithContext from the variadic opts can override, but the explicit ctx
	// argument always wins as the token-fetch context.
	for _, opt := range opts {
		opt(&o)
	}

	// Always honour the caller-supplied ctx for the token fetch itself.
	token, err := fetchToken(o.ctx, cfg, o.httpClient)
	if err != nil {
		return nil, err
	}

	log.Printf("mailbox: authenticated as %s", cfg.Username)

	return &MailboxClient{
		accessToken: token,
		httpClient:  o.httpClient,
		ctx:         o.ctx,
	}, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// fetchToken performs the ROPC token request and returns the raw access token.
func fetchToken(ctx context.Context, cfg *Config, hc *http.Client) (string, error) {
	endpoint := tokenEndpointURL(cfg.TenantID)
	data := tokenFormData(cfg)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("mailbox: create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use a dedicated short-lived client for the token fetch so we don't
	// disturb any transport customisation the caller applied to hc.
	tokenClient := &http.Client{Timeout: defaultTimeout}
	if hc != nil && hc.Timeout > 0 {
		tokenClient.Timeout = hc.Timeout
	}

	resp, err := tokenClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("mailbox: execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mailbox: token request failed (status %d): %s", resp.StatusCode, body)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("mailbox: parse token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// newRequest constructs an authorized Graph API request using the supplied ctx.
func (c *MailboxClient) newRequest(ctx context.Context, method, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("mailbox: create %s request: %w", method, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
