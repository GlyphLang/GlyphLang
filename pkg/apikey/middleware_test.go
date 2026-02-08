package apikey

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_ValidKeyWithContext(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"test-key-123"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := GetKeyInfo(r)
		if info == nil {
			t.Error("expected KeyInfo in context")
			return
		}
		if info.Key != "test-key-123" {
			t.Errorf("expected key 'test-key-123', got %q", info.Key)
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_MissingKeyReturns401(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"test-key-123"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for missing key")
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddleware_InvalidKeyReturns401(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"test-key-123"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid key")
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddleware_HeaderSpoofingProtection(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"test-key-123"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that spoofed headers were stripped and replaced with correct values
		id := r.Header.Get("X-APIKey-ID")
		name := r.Header.Get("X-APIKey-Name")

		if id == "spoofed-id" {
			t.Error("spoofed X-APIKey-ID header should have been stripped")
		}
		if name == "spoofed-name" {
			t.Error("spoofed X-APIKey-Name header should have been stripped")
		}
		if id == "" {
			t.Error("X-APIKey-ID should be set by middleware")
		}
		if name == "" {
			t.Error("X-APIKey-Name should be set by middleware")
		}

		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	// Attempt to spoof identity headers
	req.Header.Set("X-APIKey-ID", "spoofed-id")
	req.Header.Set("X-APIKey-Name", "spoofed-name")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_ContextKeyInfo(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"ctx-test-key"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := GetKeyInfo(r)
		if info == nil {
			t.Fatal("expected KeyInfo in context, got nil")
		}
		if info.Key != "ctx-test-key" {
			t.Errorf("expected key 'ctx-test-key', got %q", info.Key)
		}
		if info.ID == "" {
			t.Error("expected non-empty ID")
		}
		if info.Name == "" {
			t.Error("expected non-empty Name")
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("X-API-Key", "ctx-test-key")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)
}

func TestMiddleware_QueryParamKey(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"query-key-123"},
		QueryParam: "api_key",
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data?api_key=query-key-123", nil)
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_CustomHeaderName(t *testing.T) {
	validator := NewValidator(Config{
		StaticKeys: []string{"bearer-token-123"},
		HeaderName: "Authorization",
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(validator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", "Bearer bearer-token-123")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetKeyInfo_NoMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	info := GetKeyInfo(req)
	if info != nil {
		t.Errorf("expected nil KeyInfo when middleware hasn't run, got %+v", info)
	}
}
