package interpreter

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
