package interpreter

import (
	"fmt"
	"net/url"
)

// TextResponse signals that the route wants to send a plain text response.
type TextResponse struct {
	Body       string
	StatusCode int
}

// HTMLResponse signals that the route wants to send an HTML response.
type HTMLResponse struct {
	Body       string
	StatusCode int
}

// BlobResponse signals that the route wants to send raw bytes with a custom content type.
type BlobResponse struct {
	Data        []byte
	ContentType string
	StatusCode  int
}

// RedirectResponse signals that the route wants to send an HTTP redirect.
type RedirectResponse struct {
	URL        string
	StatusCode int
}

// ValidateRedirect checks that the URL is well-formed and the status code is a valid redirect code.
func ValidateRedirect(rawURL string, statusCode int) error {
	if _, err := url.Parse(rawURL); err != nil {
		return fmt.Errorf("invalid redirect URL: %w", err)
	}
	switch statusCode {
	case 301, 302, 307, 308:
		return nil
	default:
		return fmt.Errorf("invalid redirect status code %d: must be 301, 302, 307, or 308", statusCode)
	}
}
