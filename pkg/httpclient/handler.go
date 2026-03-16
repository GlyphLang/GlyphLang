package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// maxResponseSize is the maximum response body size (50 MB).
const maxResponseSize = 50 << 20

// Handler provides HTTP client capabilities for GlyphLang applications.
// Methods are called via reflection from the interpreter (see database.go allowedMethods).
type Handler struct {
	client *http.Client
}

// NewHandler creates a new HTTP client handler with default settings.
// Redirects are followed by default (Go's http.Client default behavior).
func NewHandler() *Handler {
	return &Handler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewHandlerWithTimeout creates a handler with a custom timeout.
func NewHandlerWithTimeout(timeout time.Duration) *Handler {
	return &Handler{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs an HTTP GET request.
// Called from GlyphLang code: http.get("https://api.example.com/data")
// or http.get("https://api.example.com/data", {headers: {"Authorization": "Bearer token"}})
func (h *Handler) Get(args interface{}) (map[string]interface{}, error) {
	reqURL, opts, err := parseRequestArgs(args)
	if err != nil {
		return nil, fmt.Errorf("http.get: %w", err)
	}
	return h.doRequest("GET", reqURL, opts)
}

// Post performs an HTTP POST request.
// Called from GlyphLang code: http.post("https://api.example.com/submit", {body: payload, headers: {…}})
func (h *Handler) Post(args interface{}) (map[string]interface{}, error) {
	reqURL, opts, err := parseRequestArgs(args)
	if err != nil {
		return nil, fmt.Errorf("http.post: %w", err)
	}
	return h.doRequest("POST", reqURL, opts)
}

// Put performs an HTTP PUT request.
func (h *Handler) Put(args interface{}) (map[string]interface{}, error) {
	reqURL, opts, err := parseRequestArgs(args)
	if err != nil {
		return nil, fmt.Errorf("http.put: %w", err)
	}
	return h.doRequest("PUT", reqURL, opts)
}

// Patch performs an HTTP PATCH request.
func (h *Handler) Patch(args interface{}) (map[string]interface{}, error) {
	reqURL, opts, err := parseRequestArgs(args)
	if err != nil {
		return nil, fmt.Errorf("http.patch: %w", err)
	}
	return h.doRequest("PATCH", reqURL, opts)
}

// Delete performs an HTTP DELETE request.
func (h *Handler) Delete(args interface{}) (map[string]interface{}, error) {
	reqURL, opts, err := parseRequestArgs(args)
	if err != nil {
		return nil, fmt.Errorf("http.delete: %w", err)
	}
	return h.doRequest("DELETE", reqURL, opts)
}

// requestOptions holds parsed options for an HTTP request.
type requestOptions struct {
	Headers         map[string]string
	Body            interface{}
	Query           map[string]string
	Timeout         time.Duration
	FollowRedirects *bool
}

// parseRequestArgs extracts the URL and options from the arguments passed to an HTTP method.
// Accepts either a string URL or a map with "url" and optional "headers", "body", "query",
// "timeout", and "followRedirects" fields.
func parseRequestArgs(args interface{}) (string, *requestOptions, error) {
	opts := &requestOptions{}

	switch v := args.(type) {
	case string:
		if err := validateURL(v); err != nil {
			return "", nil, err
		}
		return v, opts, nil

	case map[string]interface{}:
		reqURL, ok := v["url"].(string)
		if !ok || reqURL == "" {
			return "", nil, fmt.Errorf("expected 'url' field as string")
		}
		if err := validateURL(reqURL); err != nil {
			return "", nil, err
		}

		if headers, ok := v["headers"].(map[string]interface{}); ok {
			opts.Headers = make(map[string]string, len(headers))
			for k, val := range headers {
				opts.Headers[k] = fmt.Sprintf("%v", val)
			}
		}

		if body, exists := v["body"]; exists {
			opts.Body = body
		}

		if query, ok := v["query"].(map[string]interface{}); ok {
			opts.Query = make(map[string]string, len(query))
			for k, val := range query {
				opts.Query[k] = fmt.Sprintf("%v", val)
			}
		}

		if timeout, ok := v["timeout"]; ok {
			switch t := timeout.(type) {
			case int64:
				opts.Timeout = time.Duration(t) * time.Millisecond
			case float64:
				opts.Timeout = time.Duration(t) * time.Millisecond
			}
		}

		if follow, ok := v["followRedirects"].(bool); ok {
			opts.FollowRedirects = &follow
		}

		return reqURL, opts, nil

	default:
		return "", nil, fmt.Errorf("expected string URL or object with 'url' field, got %T", args)
	}
}

// validateURL checks that the URL is well-formed and uses an allowed scheme.
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https schemes are allowed, got %q", u.Scheme)
	}
	return nil
}

// doRequest executes an HTTP request and returns the response as a Glyph-friendly map.
func (h *Handler) doRequest(method, reqURL string, opts *requestOptions) (map[string]interface{}, error) {
	// Append query parameters
	if len(opts.Query) > 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		q := u.Query()
		for k, v := range opts.Query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	// Prepare request body
	var bodyReader io.Reader
	if opts.Body != nil {
		switch b := opts.Body.(type) {
		case string:
			bodyReader = strings.NewReader(b)
		case []byte:
			bodyReader = bytes.NewReader(b)
		default:
			jsonBody, err := json.Marshal(b)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Content-Type for bodies that were JSON-encoded
	if opts.Body != nil {
		if _, isStr := opts.Body.(string); !isStr {
			if _, isBytes := opts.Body.([]byte); !isBytes {
				req.Header.Set("Content-Type", "application/json")
			}
		}
	}

	// Set custom headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Handle per-request options
	client := h.client
	if opts.Timeout > 0 || (opts.FollowRedirects != nil && !*opts.FollowRedirects) {
		// Create a derived client for per-request overrides
		clientCopy := *h.client
		if opts.Timeout > 0 {
			clientCopy.Timeout = opts.Timeout
		}
		if opts.FollowRedirects != nil && !*opts.FollowRedirects {
			clientCopy.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		client = &clientCopy
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Build response headers map
	respHeaders := make(map[string]interface{}, len(resp.Header))
	for k, vals := range resp.Header {
		if len(vals) == 1 {
			respHeaders[k] = vals[0]
		} else {
			iface := make([]interface{}, len(vals))
			for i, v := range vals {
				iface[i] = v
			}
			respHeaders[k] = iface
		}
	}

	result := map[string]interface{}{
		"status":  int64(resp.StatusCode),
		"headers": respHeaders,
		"body":    string(respBody),
		"ok":      resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	return result, nil
}
