package apikey

import (
	"crypto/subtle"
	"fmt"
	"strings"
	"sync"
)

// KeyInfo holds metadata about an API key.
type KeyInfo struct {
	ID          string
	Key         string
	Name        string
	Permissions []string
}

// LookupFunc is a function that validates an API key and returns its metadata.
// Returns nil if the key is not found.
type LookupFunc func(key string) *KeyInfo

// Validator validates API keys from requests.
type Validator struct {
	mu         sync.RWMutex
	staticKeys map[string]*KeyInfo
	lookupFunc LookupFunc
	headerName string
	queryParam string
}

// Config holds configuration for the API key validator.
type Config struct {
	// StaticKeys is a list of valid API keys for simple use cases.
	StaticKeys []string
	// LookupFunc is called to validate keys dynamically (e.g., from a database).
	LookupFunc LookupFunc
	// HeaderName is the HTTP header to extract the key from.
	// Default: "X-API-Key". Set to "Authorization" for Bearer token format.
	HeaderName string
	// QueryParam is the query parameter to extract the key from (optional).
	QueryParam string
}

// NewValidator creates a new API key validator.
func NewValidator(cfg Config) *Validator {
	v := &Validator{
		staticKeys: make(map[string]*KeyInfo),
		lookupFunc: cfg.LookupFunc,
		headerName: cfg.HeaderName,
		queryParam: cfg.QueryParam,
	}

	if v.headerName == "" {
		v.headerName = "X-API-Key"
	}

	for i, key := range cfg.StaticKeys {
		v.staticKeys[key] = &KeyInfo{
			ID:   fmt.Sprintf("static-%d", i),
			Key:  key,
			Name: fmt.Sprintf("static-key-%d", i),
		}
	}

	return v
}

// ExtractKey extracts an API key from a request's headers and query parameters.
// Headers are checked first, then query parameters. The first non-empty value
// found is returned. For the Authorization header, only the "Bearer" scheme is
// accepted and the token is extracted after validating the scheme prefix.
func (v *Validator) ExtractKey(headers map[string]string, queryParams map[string]string) string {
	if headerVal, ok := headers[v.headerName]; ok && headerVal != "" {
		if v.headerName == "Authorization" {
			// Only accept Bearer scheme for Authorization header
			if strings.HasPrefix(headerVal, "Bearer ") {
				token := strings.TrimPrefix(headerVal, "Bearer ")
				if token != "" {
					return token
				}
			}
			// Reject non-Bearer Authorization values
			return ""
		}
		return headerVal
	}

	// Try query parameter as fallback
	if v.queryParam != "" {
		if qp, ok := queryParams[v.queryParam]; ok && qp != "" {
			return qp
		}
	}

	return ""
}

// Validate checks if a key is valid and returns its metadata.
// Static keys are compared using constant-time comparison to prevent
// timing attacks. The lookup function is called if no static key matches.
func (v *Validator) Validate(key string) (*KeyInfo, error) {
	if key == "" {
		return nil, fmt.Errorf("empty API key")
	}

	// Check static keys using constant-time comparison to prevent timing attacks.
	// We iterate all keys regardless of match to avoid early-exit leaks.
	v.mu.RLock()
	var matched *KeyInfo
	for staticKey, info := range v.staticKeys {
		if subtle.ConstantTimeCompare([]byte(key), []byte(staticKey)) == 1 {
			matched = info
		}
	}
	v.mu.RUnlock()

	if matched != nil {
		return matched, nil
	}

	// Try lookup function for database-backed validation
	if v.lookupFunc != nil {
		info := v.lookupFunc(key)
		if info != nil {
			return info, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}

// AddKey adds a static key at runtime.
func (v *Validator) AddKey(key string, info *KeyInfo) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.staticKeys[key] = info
}

// RemoveKey removes a static key at runtime.
func (v *Validator) RemoveKey(key string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.staticKeys, key)
}

// HeaderName returns the configured header name.
func (v *Validator) HeaderName() string {
	return v.headerName
}

// QueryParam returns the configured query parameter name.
func (v *Validator) QueryParam() string {
	return v.queryParam
}
