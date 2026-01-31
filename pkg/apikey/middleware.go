package apikey

import (
	"net/http"
)

// Middleware returns an HTTP middleware that validates API keys.
// On successful validation, the KeyInfo is stored in the request context
// and the key metadata is made available to downstream handlers.
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

			// Store key info in request header for downstream access.
			// This avoids context key collisions and works with any middleware chain.
			r.Header.Set("X-APIKey-ID", info.ID)
			r.Header.Set("X-APIKey-Name", info.Name)

			next.ServeHTTP(w, r)
		})
	}
}
