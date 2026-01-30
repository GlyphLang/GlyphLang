package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// ProviderConfig holds OAuth2 provider configuration
type ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	RedirectURL  string
	// PKCE support
	UsePKCE bool
}

// TokenResponse represents the response from a token exchange
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	IDToken      string
	Scope        string
}

// UserInfo represents user information from the OAuth2 provider
type UserInfo struct {
	ID       string
	Email    string
	Name     string
	Picture  string
	Provider string
	Raw      map[string]interface{} // Raw claims from the provider
}

// Session represents an authenticated user session
type Session struct {
	UserInfo     UserInfo
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// PKCEParams holds PKCE challenge parameters
type PKCEParams struct {
	CodeVerifier  string
	CodeChallenge string
	Method        string // "S256"
}

// Manager manages OAuth2 providers and authentication flows
type Manager struct {
	mu sync.RWMutex
	providers map[string]*ProviderConfig
	// TokenExchanger is a pluggable function for exchanging auth codes for tokens.
	// This allows injection of mock implementations for testing.
	TokenExchanger func(provider *ProviderConfig, code string) (*TokenResponse, error)
	// UserInfoFetcher is a pluggable function for fetching user info.
	UserInfoFetcher func(provider *ProviderConfig, accessToken string) (*UserInfo, error)
}

// NewManager creates a new OAuth2 manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]*ProviderConfig),
	}
}

// RegisterProvider registers an OAuth2 provider
func (m *Manager) RegisterProvider(config ProviderConfig) error {
	if config.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	if config.ClientID == "" {
		return fmt.Errorf("client ID is required for provider %s", config.Name)
	}
	if config.AuthURL == "" {
		return fmt.Errorf("auth URL is required for provider %s", config.Name)
	}
	if config.TokenURL == "" {
		return fmt.Errorf("token URL is required for provider %s", config.Name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[config.Name] = &config
	return nil
}

// GetProvider returns a registered provider by name
func (m *Manager) GetProvider(name string) (*ProviderConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.providers[name]
	return p, ok
}

// GetProviders returns a copy of all registered providers
func (m *Manager) GetProviders() map[string]ProviderConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]ProviderConfig, len(m.providers))
	for k, v := range m.providers {
		result[k] = *v
	}
	return result
}

// AuthorizationURL generates the OAuth2 authorization URL for a given provider
func (m *Manager) AuthorizationURL(providerName, state string) (string, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerName]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}

	u, err := url.Parse(provider.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL for provider %s: %w", providerName, err)
	}

	params := url.Values{}
	params.Set("client_id", provider.ClientID)
	params.Set("redirect_uri", provider.RedirectURL)
	params.Set("response_type", "code")
	params.Set("state", state)
	if len(provider.Scopes) > 0 {
		params.Set("scope", strings.Join(provider.Scopes, " "))
	}

	u.RawQuery = params.Encode()
	return u.String(), nil
}

// AuthorizationURLWithPKCE generates an authorization URL with PKCE parameters
func (m *Manager) AuthorizationURLWithPKCE(providerName, state string) (string, *PKCEParams, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerName]
	m.mu.RUnlock()
	if !ok {
		return "", nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	pkce, err := GeneratePKCE()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	u, err := url.Parse(provider.AuthURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid auth URL for provider %s: %w", providerName, err)
	}

	params := url.Values{}
	params.Set("client_id", provider.ClientID)
	params.Set("redirect_uri", provider.RedirectURL)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("code_challenge", pkce.CodeChallenge)
	params.Set("code_challenge_method", pkce.Method)
	if len(provider.Scopes) > 0 {
		params.Set("scope", strings.Join(provider.Scopes, " "))
	}

	u.RawQuery = params.Encode()
	return u.String(), pkce, nil
}

// ExchangeCode exchanges an authorization code for tokens using the configured TokenExchanger
func (m *Manager) ExchangeCode(providerName, code string) (*TokenResponse, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerName]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	if m.TokenExchanger == nil {
		return nil, fmt.Errorf("no token exchanger configured")
	}

	return m.TokenExchanger(provider, code)
}

// FetchUserInfo fetches user info using the configured UserInfoFetcher
func (m *Manager) FetchUserInfo(providerName, accessToken string) (*UserInfo, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerName]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	if m.UserInfoFetcher == nil {
		return nil, fmt.Errorf("no user info fetcher configured")
	}

	info, err := m.UserInfoFetcher(provider, accessToken)
	if err != nil {
		return nil, err
	}
	info.Provider = providerName
	return info, nil
}

// CreateSession creates a session from a token response and user info
func (m *Manager) CreateSession(providerName, code string) (*Session, error) {
	tokenResp, err := m.ExchangeCode(providerName, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	userInfo, err := m.FetchUserInfo(providerName, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("user info fetch failed: %w", err)
	}

	return &Session{
		UserInfo:     *userInfo,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// GenerateState generates a cryptographically random state parameter
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GoogleProvider returns a pre-configured Google OAuth2 provider
func GoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) ProviderConfig {
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}
	return ProviderConfig{
		Name:         "google",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
		Scopes:       scopes,
		RedirectURL:  redirectURL,
	}
}

// GitHubProvider returns a pre-configured GitHub OAuth2 provider
func GitHubProvider(clientID, clientSecret, redirectURL string, scopes []string) ProviderConfig {
	if len(scopes) == 0 {
		scopes = []string{"user:email"}
	}
	return ProviderConfig{
		Name:         "github",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scopes:       scopes,
		RedirectURL:  redirectURL,
	}
}

// GenericOIDCProvider returns a generic OIDC provider configuration
func GenericOIDCProvider(name, issuerURL, clientID, clientSecret, redirectURL string, scopes []string) ProviderConfig {
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}
	return ProviderConfig{
		Name:         name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      issuerURL + "/authorize",
		TokenURL:     issuerURL + "/oauth/token",
		UserInfoURL:  issuerURL + "/userinfo",
		Scopes:       scopes,
		RedirectURL:  redirectURL,
	}
}
