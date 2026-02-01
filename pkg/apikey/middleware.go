package apikey

import (
	"context"
	"net/http"
)

// contextKey is an unexported type for context keys to prevent collisions.
type contextKey struct{ name string }

// APIKeyInfoKey is the context key for storing validated KeyInfo.
var APIKeyInfoKey = &contextKey{"apikey-info"}

// GetKeyInfo retrieves the validated KeyInfo from the request context.
// Returns nil if no key info is present (middleware did not run).
func GetKeyInfo(r *http.Request) *KeyInfo {
	if info, ok := r.Context().Value(APIKeyInfoKey).(*KeyInfo); ok {
		return info
	}
	return nil
}

// Middleware returns an HTTP middleware that validates API keys.
// On successful validation, the KeyInfo is stored in the request context
// (not headers, which can be spoofed by clients).
func Middleware(validator *Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := make(map[string]string)
			headers[validator.HeaderName()] = r.Header.Get(validator.HeaderName())

			queryParams := make(map[string]string)
			if qp := validator.QueryParam(); qp != "" {
				queryParams[qp] = r.URL.Query().Get(qp)
			}

			key := validator.ExtractKey(headers, queryParams)
			if key == "" {
				http.Error(w, `{"error":"missing API key"}`, http.StatusUnauthorized)
				return
			}

			info, err := validator.Validate(key)
			if err != nil {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			// Store key info in request context (not headers) to prevent
			// client-side spoofing of identity headers.
			ctx := context.WithValue(r.Context(), APIKeyInfoKey, info)
			r = r.WithContext(ctx)

			// Also set headers for backward compatibility, but strip them
			// from incoming requests first to prevent spoofing.
			r.Header.Del("X-APIKey-ID")
			r.Header.Del("X-APIKey-Name")
			r.Header.Set("X-APIKey-ID", info.ID)
			r.Header.Set("X-APIKey-Name", info.Name)

			next.ServeHTTP(w, r)
		})
	}
}
