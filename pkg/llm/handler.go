package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Provider represents an LLM provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderOllama    Provider = "ollama"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest contains parameters for a completion request
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// EmbeddingRequest contains parameters for an embedding request
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// Handler provides LLM capabilities for GlyphLang applications.
// Methods are called via reflection from the interpreter (see database.go allowedMethods).
type Handler struct {
	provider Provider
	apiKey   string
	baseURL  string
	client   *http.Client
}

// NewHandler creates a new LLM handler for the specified provider
func NewHandler(provider string, apiKey string) *Handler {
	h := &Handler{
		provider: Provider(provider),
		apiKey:   apiKey,
		client:   &http.Client{},
	}
	switch h.provider {
	case ProviderOpenAI:
		h.baseURL = "https://api.openai.com/v1"
	case ProviderAnthropic:
		h.baseURL = "https://api.anthropic.com/v1"
	case ProviderOllama:
		h.baseURL = "http://localhost:11434/api"
	}
	return h
}

// NewHandlerWithBaseURL creates a handler with a custom base URL (e.g. for Ollama or proxies).
// The base URL must use http or https scheme.
func NewHandlerWithBaseURL(provider string, apiKey string, baseURL string) (*Handler, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("only http and https schemes are allowed, got %q", u.Scheme)
	}
	h := NewHandler(provider, apiKey)
	h.baseURL = baseURL
	return h, nil
}

// Complete sends a non-streaming completion request and returns the response.
// Called from GlyphLang code: llm.Complete(request)
func (h *Handler) Complete(request interface{}) (map[string]interface{}, error) {
	req, err := parseCompletionRequest(request)
	if err != nil {
		return nil, fmt.Errorf("invalid completion request: %w", err)
	}

	switch h.provider {
	case ProviderOpenAI:
		return h.completeOpenAI(req)
	case ProviderAnthropic:
		return h.completeAnthropic(req)
	case ProviderOllama:
		return h.completeOllama(req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", h.provider)
	}
}

// Chat is an alias for Complete
func (h *Handler) Chat(request interface{}) (map[string]interface{}, error) {
	return h.Complete(request)
}

// Stream sends a streaming completion request and returns collected chunks.
// Called from GlyphLang code: llm.Stream(request)
func (h *Handler) Stream(request interface{}) ([]map[string]interface{}, error) {
	req, err := parseCompletionRequest(request)
	if err != nil {
		return nil, fmt.Errorf("invalid stream request: %w", err)
	}

	switch h.provider {
	case ProviderOpenAI:
		return h.streamOpenAI(req)
	case ProviderAnthropic:
		return h.streamAnthropic(req)
	case ProviderOllama:
		return h.streamOllama(req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", h.provider)
	}
}

// Embed generates vector embeddings for the given input.
// Called from GlyphLang code: llm.Embed(request)
func (h *Handler) Embed(request interface{}) (map[string]interface{}, error) {
	req, err := parseEmbeddingRequest(request)
	if err != nil {
		return nil, fmt.Errorf("invalid embedding request: %w", err)
	}

	switch h.provider {
	case ProviderOpenAI:
		return h.embedOpenAI(req)
	default:
		return nil, fmt.Errorf("embeddings not supported for provider: %s", h.provider)
	}
}

// ListModels returns available models for the configured provider
func (h *Handler) ListModels() ([]map[string]interface{}, error) {
	switch h.provider {
	case ProviderOpenAI:
		return h.listModelsOpenAI()
	case ProviderOllama:
		return h.listModelsOllama()
	default:
		return nil, fmt.Errorf("model listing not supported for provider: %s", h.provider)
	}
}

// TokenCount estimates the number of tokens in a string (rough approximation)
func (h *Handler) TokenCount(text interface{}) (map[string]interface{}, error) {
	var s string
	switch v := text.(type) {
	case string:
		s = v
	case map[string]interface{}:
		t, ok := v["text"].(string)
		if !ok {
			return nil, fmt.Errorf("expected text field in request")
		}
		s = t
	default:
		return nil, fmt.Errorf("expected string or object with text field")
	}
	estimate := len(s) / 4
	if estimate == 0 && len(s) > 0 {
		estimate = 1
	}
	return map[string]interface{}{
		"estimate": int64(estimate),
		"text":     s,
	}, nil
}

// Request parsing helpers

func parseCompletionRequest(request interface{}) (*CompletionRequest, error) {
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for completion request")
	}

	cr := &CompletionRequest{}

	if model, ok := reqMap["model"].(string); ok {
		cr.Model = model
	}

	if messages, ok := reqMap["messages"].([]interface{}); ok {
		for _, m := range messages {
			msgMap, ok := m.(map[string]interface{})
			if !ok {
				continue
			}
			msg := Message{}
			if role, ok := msgMap["role"].(string); ok {
				msg.Role = role
			}
			if content, ok := msgMap["content"].(string); ok {
				msg.Content = content
			}
			cr.Messages = append(cr.Messages, msg)
		}
	}

	if temp, ok := reqMap["temperature"].(float64); ok {
		cr.Temperature = temp
	}
	if maxTokens, ok := reqMap["max_tokens"].(int64); ok {
		cr.MaxTokens = int(maxTokens)
	}

	return cr, nil
}

func parseEmbeddingRequest(request interface{}) (*EmbeddingRequest, error) {
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for embedding request")
	}

	er := &EmbeddingRequest{}
	if model, ok := reqMap["model"].(string); ok {
		er.Model = model
	}
	if input, ok := reqMap["input"].(string); ok {
		er.Input = input
	}

	return er, nil
}

// OpenAI provider implementations

func (h *Handler) completeOpenAI(req *CompletionRequest) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	respBody, err := h.doHTTPRequest("POST", h.baseURL+"/chat/completions", body, h.openAIHeaders())
	if err != nil {
		return nil, err
	}

	return parseOpenAICompletionResponse(respBody)
}

func (h *Handler) streamOpenAI(req *CompletionRequest) ([]map[string]interface{}, error) {
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	respBody, err := h.doHTTPRequest("POST", h.baseURL+"/chat/completions", body, h.openAIHeaders())
	if err != nil {
		return nil, err
	}

	return parseSSEStream(respBody)
}

func (h *Handler) embedOpenAI(req *EmbeddingRequest) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	respBody, err := h.doHTTPRequest("POST", h.baseURL+"/embeddings", body, h.openAIHeaders())
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	data, ok := resp["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("unexpected embedding response format")
	}

	first, ok := data[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected embedding data format")
	}

	embedding, ok := first["embedding"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing embedding vector in response")
	}

	vector := make([]interface{}, len(embedding))
	copy(vector, embedding)
	return map[string]interface{}{
		"vector": vector,
		"model":  req.Model,
	}, nil
}

func (h *Handler) listModelsOpenAI() ([]map[string]interface{}, error) {
	respBody, err := h.doHTTPRequest("GET", h.baseURL+"/models", nil, h.openAIHeaders())
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	var models []map[string]interface{}
	if data, ok := resp["data"].([]interface{}); ok {
		for _, m := range data {
			if model, ok := m.(map[string]interface{}); ok {
				models = append(models, map[string]interface{}{
					"id":      model["id"],
					"created": model["created"],
				})
			}
		}
	}

	return models, nil
}

func parseOpenAICompletionResponse(body []byte) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		return nil, fmt.Errorf("OpenAI API error: %v", errObj["message"])
	}

	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("unexpected OpenAI response format: no choices")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected message format in choice")
	}

	result := map[string]interface{}{
		"content":       message["content"],
		"model":         resp["model"],
		"finish_reason": choice["finish_reason"],
	}

	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		result["tokens_used"] = usage["total_tokens"]
	}

	return result, nil
}

// Anthropic provider implementations

func (h *Handler) completeAnthropic(req *CompletionRequest) (map[string]interface{}, error) {
	messages := make([]map[string]interface{}, 0, len(req.Messages))
	var systemMsg string

	for _, m := range req.Messages {
		if m.Role == "system" {
			systemMsg = m.Content
			continue
		}
		messages = append(messages, map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	body := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}
	if systemMsg != "" {
		body["system"] = systemMsg
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	} else {
		body["max_tokens"] = 1024
	}

	respBody, err := h.doHTTPRequest("POST", h.baseURL+"/messages", body, h.anthropicHeaders())
	if err != nil {
		return nil, err
	}

	return parseAnthropicResponse(respBody)
}

func (h *Handler) streamAnthropic(req *CompletionRequest) ([]map[string]interface{}, error) {
	result, err := h.completeAnthropic(req)
	if err != nil {
		return nil, err
	}
	return []map[string]interface{}{
		{"content": result["content"], "done": true},
	}, nil
}

func parseAnthropicResponse(body []byte) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	if errType, ok := resp["type"].(string); ok && errType == "error" {
		if errObj, ok := resp["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("Anthropic API error: %v", errObj["message"])
		}
		return nil, fmt.Errorf("Anthropic API error: unknown")
	}

	content := ""
	if contentBlocks, ok := resp["content"].([]interface{}); ok {
		for _, block := range contentBlocks {
			b, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := b["text"].(string); ok {
				content += text
			}
		}
	}

	result := map[string]interface{}{
		"content":       content,
		"model":         resp["model"],
		"finish_reason": resp["stop_reason"],
	}

	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		input, _ := usage["input_tokens"].(float64)
		output, _ := usage["output_tokens"].(float64)
		result["tokens_used"] = int64(input + output)
	}

	return result, nil
}

// Ollama provider implementations

func (h *Handler) completeOllama(req *CompletionRequest) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"model":  req.Model,
		"stream": false,
	}
	if len(req.Messages) > 0 {
		body["messages"] = req.Messages
	}
	if req.Temperature > 0 {
		body["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	respBody, err := h.doHTTPRequest("POST", h.baseURL+"/chat", body, nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	content := ""
	if message, ok := resp["message"].(map[string]interface{}); ok {
		if c, ok := message["content"].(string); ok {
			content = c
		}
	}

	return map[string]interface{}{
		"content":       content,
		"model":         resp["model"],
		"finish_reason": "stop",
	}, nil
}

func (h *Handler) streamOllama(req *CompletionRequest) ([]map[string]interface{}, error) {
	result, err := h.completeOllama(req)
	if err != nil {
		return nil, err
	}
	return []map[string]interface{}{
		{"content": result["content"], "done": true},
	}, nil
}

func (h *Handler) listModelsOllama() ([]map[string]interface{}, error) {
	respBody, err := h.doHTTPRequest("GET", h.baseURL+"/tags", nil, nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	var models []map[string]interface{}
	if modelList, ok := resp["models"].([]interface{}); ok {
		for _, m := range modelList {
			if model, ok := m.(map[string]interface{}); ok {
				models = append(models, map[string]interface{}{
					"id":   model["name"],
					"size": model["size"],
				})
			}
		}
	}

	return models, nil
}

// Unified HTTP helper

func (h *Handler) doHTTPRequest(method, url string, body interface{}, headers map[string]string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBody))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	const maxLLMResponseSize = 50 << 20 // 50 MB
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxLLMResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (h *Handler) openAIHeaders() map[string]string {
	headers := map[string]string{}
	if h.apiKey != "" {
		headers["Authorization"] = "Bearer " + h.apiKey
	}
	return headers
}

func (h *Handler) anthropicHeaders() map[string]string {
	headers := map[string]string{
		"anthropic-version": "2023-06-01",
	}
	if h.apiKey != "" {
		headers["x-api-key"] = h.apiKey
	}
	return headers
}

func parseSSEStream(body []byte) ([]map[string]interface{}, error) {
	var chunks []map[string]interface{}
	lines := strings.Split(string(body), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			chunks = append(chunks, map[string]interface{}{
				"content": "",
				"done":    true,
			})
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}

		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}

		delta, ok := choice["delta"].(map[string]interface{})
		if !ok {
			continue
		}

		content := ""
		if c, ok := delta["content"].(string); ok {
			content = c
		}
		chunks = append(chunks, map[string]interface{}{
			"content": content,
			"done":    false,
		})
	}

	return chunks, nil
}
