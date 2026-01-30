// Tests for OAuth2 provider management (pkg/oauth2/provider.go) and PKCE (pkg/oauth2/pkce.go).
// Implementation files verified: provider.go (Manager, ProviderConfig, TokenResponse, UserInfo,
// Session, PKCEParams, NewManager, RegisterProvider, GetProvider, GetProviders, AuthorizationURL,
// AuthorizationURLWithPKCE, ExchangeCode, FetchUserInfo, CreateSession, GenerateState,
// GoogleProvider, GitHubProvider, GenericOIDCProvider) and pkce.go (GeneratePKCE, VerifyPKCE).
package oauth2

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGoogleProvider() ProviderConfig {
	return GoogleProvider("test-client-id", "test-secret", "http://localhost:3000/auth/google/callback", nil)
}

func testGitHubProvider() ProviderConfig {
	return GitHubProvider("gh-client-id", "gh-secret", "http://localhost:3000/auth/github/callback", nil)
}

// TestNewManager tests creating a new OAuth2 manager
func TestNewManager(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)
	assert.Empty(t, m.GetProviders())
}

// TestRegisterProvider tests registering a provider
func TestRegisterProvider(t *testing.T) {
	m := NewManager()
	err := m.RegisterProvider(testGoogleProvider())
	require.NoError(t, err)

	p, ok := m.GetProvider("google")
	assert.True(t, ok)
	assert.Equal(t, "google", p.Name)
	assert.Equal(t, "test-client-id", p.ClientID)
}

// TestRegisterProviderValidation tests provider validation
func TestRegisterProviderValidation(t *testing.T) {
	m := NewManager()

	err := m.RegisterProvider(ProviderConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	err = m.RegisterProvider(ProviderConfig{Name: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client ID")

	err = m.RegisterProvider(ProviderConfig{Name: "test", ClientID: "id"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth URL")

	err = m.RegisterProvider(ProviderConfig{Name: "test", ClientID: "id", AuthURL: "http://auth"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token URL")
}

// TestGetProviders tests getting all providers (returns copy)
func TestGetProviders(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))
	require.NoError(t, m.RegisterProvider(testGitHubProvider()))

	providers := m.GetProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "google")
	assert.Contains(t, providers, "github")
}

// TestGetProviderNotFound tests getting a non-existent provider
func TestGetProviderNotFound(t *testing.T) {
	m := NewManager()
	_, ok := m.GetProvider("nonexistent")
	assert.False(t, ok)
}

// TestAuthorizationURL tests generating an authorization URL
func TestAuthorizationURL(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	authURL, err := m.AuthorizationURL("google", "test-state-123")
	require.NoError(t, err)

	u, err := url.Parse(authURL)
	require.NoError(t, err)

	assert.Equal(t, "accounts.google.com", u.Host)
	assert.Equal(t, "test-client-id", u.Query().Get("client_id"))
	assert.Equal(t, "code", u.Query().Get("response_type"))
	assert.Equal(t, "test-state-123", u.Query().Get("state"))
	assert.Contains(t, u.Query().Get("scope"), "openid")
	assert.Equal(t, "http://localhost:3000/auth/google/callback", u.Query().Get("redirect_uri"))
}

// TestAuthorizationURLUnknownProvider tests auth URL generation for unknown provider
func TestAuthorizationURLUnknownProvider(t *testing.T) {
	m := NewManager()
	_, err := m.AuthorizationURL("unknown", "state")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown provider")
}

// TestAuthorizationURLWithPKCE tests generating a PKCE-enhanced authorization URL
func TestAuthorizationURLWithPKCE(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	authURL, pkce, err := m.AuthorizationURLWithPKCE("google", "test-state")
	require.NoError(t, err)
	require.NotNil(t, pkce)

	u, err := url.Parse(authURL)
	require.NoError(t, err)

	assert.Equal(t, pkce.CodeChallenge, u.Query().Get("code_challenge"))
	assert.Equal(t, "S256", u.Query().Get("code_challenge_method"))
	assert.NotEmpty(t, pkce.CodeVerifier)
	assert.NotEmpty(t, pkce.CodeChallenge)
}

// TestExchangeCode tests token exchange with mock exchanger
func TestExchangeCode(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	m.TokenExchanger = func(provider *ProviderConfig, code string) (*TokenResponse, error) {
		assert.Equal(t, "google", provider.Name)
		assert.Equal(t, "auth-code-123", code)
		return &TokenResponse{
			AccessToken:  "access-token-xyz",
			RefreshToken: "refresh-token-xyz",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
		}, nil
	}

	resp, err := m.ExchangeCode("google", "auth-code-123")
	require.NoError(t, err)
	assert.Equal(t, "access-token-xyz", resp.AccessToken)
	assert.Equal(t, "refresh-token-xyz", resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)
}

// TestExchangeCodeNoExchanger tests token exchange without configured exchanger
func TestExchangeCodeNoExchanger(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	_, err := m.ExchangeCode("google", "code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token exchanger")
}

// TestExchangeCodeUnknownProvider tests token exchange with unknown provider
func TestExchangeCodeUnknownProvider(t *testing.T) {
	m := NewManager()
	_, err := m.ExchangeCode("unknown", "code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown provider")
}

// TestFetchUserInfo tests user info fetching with mock fetcher
func TestFetchUserInfo(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	m.UserInfoFetcher = func(provider *ProviderConfig, accessToken string) (*UserInfo, error) {
		assert.Equal(t, "access-token-xyz", accessToken)
		return &UserInfo{
			ID:    "user-123",
			Email: "user@example.com",
			Name:  "Test User",
		}, nil
	}

	info, err := m.FetchUserInfo("google", "access-token-xyz")
	require.NoError(t, err)
	assert.Equal(t, "user-123", info.ID)
	assert.Equal(t, "user@example.com", info.Email)
	assert.Equal(t, "google", info.Provider)
}

// TestFetchUserInfoNoFetcher tests user info without configured fetcher
func TestFetchUserInfoNoFetcher(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	_, err := m.FetchUserInfo("google", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no user info fetcher")
}

// TestCreateSession tests the full session creation flow
func TestCreateSession(t *testing.T) {
	m := NewManager()
	require.NoError(t, m.RegisterProvider(testGoogleProvider()))

	m.TokenExchanger = func(provider *ProviderConfig, code string) (*TokenResponse, error) {
		return &TokenResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
		}, nil
	}
	m.UserInfoFetcher = func(provider *ProviderConfig, accessToken string) (*UserInfo, error) {
		return &UserInfo{
			ID:    "user-1",
			Email: "test@example.com",
			Name:  "Test User",
		}, nil
	}

	session, err := m.CreateSession("google", "auth-code")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", session.UserInfo.Email)
	assert.Equal(t, "google", session.UserInfo.Provider)
	assert.Equal(t, "access-token", session.AccessToken)
	assert.Equal(t, "refresh-token", session.RefreshToken)
	assert.Equal(t, 3600, session.ExpiresIn)
}

// TestPKCEGenerate tests PKCE parameter generation
func TestPKCEGenerate(t *testing.T) {
	pkce, err := GeneratePKCE()
	require.NoError(t, err)
	assert.NotEmpty(t, pkce.CodeVerifier)
	assert.NotEmpty(t, pkce.CodeChallenge)
	assert.Equal(t, "S256", pkce.Method)
	assert.True(t, VerifyPKCE(pkce.CodeVerifier, pkce.CodeChallenge))
}

// TestPKCEVerify tests PKCE verification
func TestPKCEVerify(t *testing.T) {
	pkce, err := GeneratePKCE()
	require.NoError(t, err)

	assert.True(t, VerifyPKCE(pkce.CodeVerifier, pkce.CodeChallenge))
	assert.False(t, VerifyPKCE("wrong-verifier", pkce.CodeChallenge))
}

// TestPKCEUniqueness tests that each PKCE generation produces unique values
func TestPKCEUniqueness(t *testing.T) {
	pkce1, err := GeneratePKCE()
	require.NoError(t, err)
	pkce2, err := GeneratePKCE()
	require.NoError(t, err)

	assert.NotEqual(t, pkce1.CodeVerifier, pkce2.CodeVerifier)
	assert.NotEqual(t, pkce1.CodeChallenge, pkce2.CodeChallenge)
}

// TestGenerateState tests state parameter generation
func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	require.NoError(t, err)
	assert.NotEmpty(t, state1)

	state2, err := GenerateState()
	require.NoError(t, err)
	assert.NotEqual(t, state1, state2)
}

// TestGoogleProviderConfig tests pre-configured Google provider
func TestGoogleProviderConfig(t *testing.T) {
	p := GoogleProvider("id", "secret", "http://localhost/callback", nil)
	assert.Equal(t, "google", p.Name)
	assert.Equal(t, "id", p.ClientID)
	assert.Contains(t, p.AuthURL, "accounts.google.com")
	assert.Contains(t, p.Scopes, "openid")
}

// TestGitHubProviderConfig tests pre-configured GitHub provider
func TestGitHubProviderConfig(t *testing.T) {
	p := GitHubProvider("id", "secret", "http://localhost/callback", nil)
	assert.Equal(t, "github", p.Name)
	assert.Contains(t, p.AuthURL, "github.com")
	assert.Contains(t, p.Scopes, "user:email")
}

// TestGenericOIDCProviderConfig tests generic OIDC provider
func TestGenericOIDCProviderConfig(t *testing.T) {
	p := GenericOIDCProvider("myidp", "https://idp.example.com", "id", "secret", "http://localhost/callback", nil)
	assert.Equal(t, "myidp", p.Name)
	assert.Equal(t, "https://idp.example.com/authorize", p.AuthURL)
	assert.Equal(t, "https://idp.example.com/oauth/token", p.TokenURL)
	assert.Equal(t, "https://idp.example.com/userinfo", p.UserInfoURL)
}

// TestCustomScopes tests providers with custom scopes
func TestCustomScopes(t *testing.T) {
	gp := GoogleProvider("id", "secret", "http://localhost/callback", []string{"openid", "calendar"})
	assert.Equal(t, []string{"openid", "calendar"}, gp.Scopes)

	ghp := GitHubProvider("id", "secret", "http://localhost/callback", []string{"repo", "user"})
	assert.Equal(t, []string{"repo", "user"}, ghp.Scopes)
}
