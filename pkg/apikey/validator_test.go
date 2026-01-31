package apikey

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator_Defaults(t *testing.T) {
	v := NewValidator(Config{})
	assert.Equal(t, "X-API-Key", v.HeaderName())
	assert.Equal(t, "", v.QueryParam())
}

func TestNewValidator_CustomHeader(t *testing.T) {
	v := NewValidator(Config{HeaderName: "Authorization"})
	assert.Equal(t, "Authorization", v.HeaderName())
}

func TestNewValidator_StaticKeys(t *testing.T) {
	v := NewValidator(Config{
		StaticKeys: []string{"key-1", "key-2"},
	})

	info, err := v.Validate("key-1")
	require.NoError(t, err)
	assert.Equal(t, "static-0", info.ID)

	info, err = v.Validate("key-2")
	require.NoError(t, err)
	assert.Equal(t, "static-1", info.ID)
}

func TestValidate_EmptyKey(t *testing.T) {
	v := NewValidator(Config{StaticKeys: []string{"valid"}})
	_, err := v.Validate("")
	assert.EqualError(t, err, "empty API key")
}

func TestValidate_InvalidKey(t *testing.T) {
	v := NewValidator(Config{StaticKeys: []string{"valid"}})
	_, err := v.Validate("invalid")
	assert.EqualError(t, err, "invalid API key")
}

func TestValidate_LookupFunc(t *testing.T) {
	v := NewValidator(Config{
		LookupFunc: func(key string) *KeyInfo {
			if key == "db-key-123" {
				return &KeyInfo{ID: "db-1", Key: key, Name: "database-key"}
			}
			return nil
		},
	})

	info, err := v.Validate("db-key-123")
	require.NoError(t, err)
	assert.Equal(t, "db-1", info.ID)
	assert.Equal(t, "database-key", info.Name)

	_, err = v.Validate("nonexistent")
	assert.Error(t, err)
}

func TestValidate_StaticKeyBeforeLookup(t *testing.T) {
	lookupCalled := false
	v := NewValidator(Config{
		StaticKeys: []string{"static-key"},
		LookupFunc: func(key string) *KeyInfo {
			lookupCalled = true
			return nil
		},
	})

	info, err := v.Validate("static-key")
	require.NoError(t, err)
	assert.Equal(t, "static-0", info.ID)
	assert.False(t, lookupCalled)
}

func TestExtractKey_DefaultHeader(t *testing.T) {
	v := NewValidator(Config{})

	key := v.ExtractKey(
		map[string]string{"X-API-Key": "my-key"},
		map[string]string{},
	)
	assert.Equal(t, "my-key", key)
}

func TestExtractKey_AuthorizationBearer(t *testing.T) {
	v := NewValidator(Config{HeaderName: "Authorization"})

	key := v.ExtractKey(
		map[string]string{"Authorization": "Bearer token-123"},
		map[string]string{},
	)
	assert.Equal(t, "token-123", key)
}

func TestExtractKey_AuthorizationNonBearer(t *testing.T) {
	v := NewValidator(Config{HeaderName: "Authorization"})

	// Non-Bearer Authorization values are rejected
	key := v.ExtractKey(
		map[string]string{"Authorization": "Basic dXNlcjpwYXNz"},
		map[string]string{},
	)
	assert.Equal(t, "", key)
}

func TestExtractKey_QueryParam(t *testing.T) {
	v := NewValidator(Config{QueryParam: "api_key"})

	key := v.ExtractKey(
		map[string]string{},
		map[string]string{"api_key": "query-key"},
	)
	assert.Equal(t, "query-key", key)
}

func TestExtractKey_HeaderTakesPrecedence(t *testing.T) {
	v := NewValidator(Config{QueryParam: "api_key"})

	key := v.ExtractKey(
		map[string]string{"X-API-Key": "header-key"},
		map[string]string{"api_key": "query-key"},
	)
	assert.Equal(t, "header-key", key)
}

func TestExtractKey_EmptyHeader_FallsToQuery(t *testing.T) {
	v := NewValidator(Config{QueryParam: "api_key"})

	key := v.ExtractKey(
		map[string]string{"X-API-Key": ""},
		map[string]string{"api_key": "query-key"},
	)
	assert.Equal(t, "query-key", key)
}

func TestExtractKey_NoKeyFound(t *testing.T) {
	v := NewValidator(Config{})

	key := v.ExtractKey(
		map[string]string{},
		map[string]string{},
	)
	assert.Equal(t, "", key)
}

func TestAddKey(t *testing.T) {
	v := NewValidator(Config{})

	v.AddKey("new-key", &KeyInfo{ID: "new-1", Key: "new-key", Name: "added"})

	info, err := v.Validate("new-key")
	require.NoError(t, err)
	assert.Equal(t, "new-1", info.ID)
}

func TestRemoveKey(t *testing.T) {
	v := NewValidator(Config{StaticKeys: []string{"removable"}})

	_, err := v.Validate("removable")
	require.NoError(t, err)

	v.RemoveKey("removable")

	_, err = v.Validate("removable")
	assert.Error(t, err)
}

func TestMiddleware_ValidKey(t *testing.T) {
	v := NewValidator(Config{StaticKeys: []string{"valid-key"}})

	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "static-0", r.Header.Get("X-APIKey-ID"))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "valid-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddleware_MissingKey(t *testing.T) {
	v := NewValidator(Config{StaticKeys: []string{"valid-key"}})

	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/data", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing API key")
}

func TestMiddleware_InvalidKey(t *testing.T) {
	v := NewValidator(Config{StaticKeys: []string{"valid-key"}})

	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid API key")
}

func TestMiddleware_QueryParam(t *testing.T) {
	v := NewValidator(Config{
		StaticKeys: []string{"qp-key"},
		QueryParam: "token",
	})

	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/data?token=qp-key", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddleware_SetsKeyMetadata(t *testing.T) {
	v := NewValidator(Config{
		LookupFunc: func(key string) *KeyInfo {
			return &KeyInfo{ID: "user-42", Key: key, Name: "production"}
		},
	})

	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "user-42", r.Header.Get("X-APIKey-ID"))
		assert.Equal(t, "production", r.Header.Get("X-APIKey-Name"))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-API-Key", "any-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidate_KeyPermissions(t *testing.T) {
	v := NewValidator(Config{})
	v.AddKey("admin-key", &KeyInfo{
		ID:          "admin-1",
		Key:         "admin-key",
		Name:        "admin",
		Permissions: []string{"read", "write", "admin"},
	})

	info, err := v.Validate("admin-key")
	require.NoError(t, err)
	assert.Equal(t, []string{"read", "write", "admin"}, info.Permissions)
}

func TestExtractKey_BearerEmptyToken(t *testing.T) {
	v := NewValidator(Config{HeaderName: "Authorization"})

	// "Bearer " with nothing after should return empty
	key := v.ExtractKey(
		map[string]string{"Authorization": "Bearer "},
		map[string]string{},
	)
	assert.Equal(t, "", key)
}
