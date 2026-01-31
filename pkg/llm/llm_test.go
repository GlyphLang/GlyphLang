package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHandler verifies handler creation with correct provider settings
func TestNewHandler(t *testing.T) {
	h := NewHandler("openai", "test-key")
	assert.Equal(t, ProviderOpenAI, h.provider)
	assert.Equal(t, "test-key", h.apiKey)
	assert.Equal(t, "https://api.openai.com/v1", h.baseURL)

	h2 := NewHandler("anthropic", "ant-key")
	assert.Equal(t, ProviderAnthropic, h2.provider)
	assert.Equal(t, "https://api.anthropic.com/v1", h2.baseURL)

	h3 := NewHandler("ollama", "")
	assert.Equal(t, ProviderOllama, h3.provider)
	assert.Equal(t, "http://localhost:11434/api", h3.baseURL)
}

// TestNewHandlerWithBaseURL verifies custom base URL override
func TestNewHandlerWithBaseURL(t *testing.T) {
	h := NewHandlerWithBaseURL("openai", "key", "http://localhost:8080")
	assert.Equal(t, "http://localhost:8080", h.baseURL)
}

// TestMockHandlerComplete verifies mock completion returns expected response
func TestMockHandlerComplete(t *testing.T) {
	mock := NewMockHandler()

	request := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		},
	}

	result, err := mock.Complete(request)
	require.NoError(t, err)
	assert.Equal(t, "Mock response to: Hello", result["content"])
	assert.Equal(t, "gpt-4", result["model"])
	assert.Equal(t, "stop", result["finish_reason"])
}

// TestMockHandlerCannedResponse verifies OnComplete sets canned responses
func TestMockHandlerCannedResponse(t *testing.T) {
	mock := NewMockHandler()
	mock.OnComplete("gpt-4", "Custom response")

	request := map[string]interface{}{
		"model": "gpt-4",
	}

	result, err := mock.Complete(request)
	require.NoError(t, err)
	assert.Equal(t, "Custom response", result["content"])
}

// TestMockHandlerChat verifies Chat is an alias for Complete
func TestMockHandlerChat(t *testing.T) {
	mock := NewMockHandler()

	request := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Test",
			},
		},
	}

	result, err := mock.Chat(request)
	require.NoError(t, err)
	assert.Contains(t, result["content"].(string), "Test")
}

// TestMockHandlerStream verifies streaming returns chunks
func TestMockHandlerStream(t *testing.T) {
	mock := NewMockHandler()
	mock.SetDefaultResponse("Hello world")

	request := map[string]interface{}{
		"model": "gpt-4",
	}

	chunks, err := mock.Stream(request)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks)

	// Last chunk should have done=true
	lastChunk := chunks[len(chunks)-1]
	assert.True(t, lastChunk["done"].(bool))
}

// TestMockHandlerEmbed verifies embedding returns vector
func TestMockHandlerEmbed(t *testing.T) {
	mock := NewMockHandler()

	request := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": "Hello world",
	}

	result, err := mock.Embed(request)
	require.NoError(t, err)
	assert.Equal(t, "text-embedding-ada-002", result["model"])

	vector, ok := result["vector"].([]interface{})
	require.True(t, ok)
	assert.Len(t, vector, 8)
}

// TestMockHandlerListModels verifies model listing
func TestMockHandlerListModels(t *testing.T) {
	mock := NewMockHandler()

	models, err := mock.ListModels()
	require.NoError(t, err)
	assert.Len(t, models, 3)
	assert.Equal(t, "gpt-4", models[0]["id"])
}

// TestMockHandlerTokenCount verifies token estimation
func TestMockHandlerTokenCount(t *testing.T) {
	mock := NewMockHandler()

	result, err := mock.TokenCount("Hello world test")
	require.NoError(t, err)
	assert.Equal(t, int64(4), result["estimate"])

	// Test with object input
	result2, err := mock.TokenCount(map[string]interface{}{
		"text": "Hello world test",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(4), result2["estimate"])
}

// TestMockHandlerCallLog verifies call tracking
func TestMockHandlerCallLog(t *testing.T) {
	mock := NewMockHandler()

	request := map[string]interface{}{"model": "gpt-4"}
	mock.Complete(request)
	mock.Complete(request)
	mock.Embed(map[string]interface{}{"model": "ada", "input": "test"})

	assert.Equal(t, 2, mock.CalledCount("Complete"))
	assert.Equal(t, 1, mock.CalledCount("Embed"))
	assert.Equal(t, 0, mock.CalledCount("Stream"))
	assert.Len(t, mock.Calls(), 3)
}

// TestParseCompletionRequest verifies request parsing
func TestParseCompletionRequest(t *testing.T) {
	request := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "system",
				"content": "You are helpful",
			},
			map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		},
		"temperature": 0.7,
		"max_tokens":  int64(100),
	}

	cr, err := parseCompletionRequest(request)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4", cr.Model)
	assert.Len(t, cr.Messages, 2)
	assert.Equal(t, "system", cr.Messages[0].Role)
	assert.Equal(t, "user", cr.Messages[1].Role)
	assert.Equal(t, 0.7, cr.Temperature)
	assert.Equal(t, 100, cr.MaxTokens)
}

// TestParseCompletionRequestInvalid verifies error on invalid input
func TestParseCompletionRequestInvalid(t *testing.T) {
	_, err := parseCompletionRequest("not a map")
	assert.Error(t, err)
}

// TestParseEmbeddingRequest verifies embedding request parsing
func TestParseEmbeddingRequest(t *testing.T) {
	request := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": "Hello world",
	}

	er, err := parseEmbeddingRequest(request)
	require.NoError(t, err)
	assert.Equal(t, "text-embedding-ada-002", er.Model)
	assert.Equal(t, "Hello world", er.Input)
}

// TestParseOpenAIResponse verifies response parsing
func TestParseOpenAIResponse(t *testing.T) {
	response := map[string]interface{}{
		"choices": []interface{}{
			map[string]interface{}{
				"message": map[string]interface{}{
					"content": "Hello!",
					"role":    "assistant",
				},
				"finish_reason": "stop",
			},
		},
		"model": "gpt-4",
		"usage": map[string]interface{}{
			"total_tokens": float64(42),
		},
	}

	body, _ := json.Marshal(response)
	result, err := parseOpenAICompletionResponse(body)
	require.NoError(t, err)
	assert.Equal(t, "Hello!", result["content"])
	assert.Equal(t, "gpt-4", result["model"])
	assert.Equal(t, "stop", result["finish_reason"])
	assert.Equal(t, float64(42), result["tokens_used"])
}

// TestParseOpenAIResponseError verifies error response parsing
func TestParseOpenAIResponseError(t *testing.T) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message": "Invalid API key",
			"type":    "invalid_request_error",
		},
	}

	body, _ := json.Marshal(response)
	_, err := parseOpenAICompletionResponse(body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid API key")
}

// TestParseAnthropicResponse verifies Anthropic response parsing
func TestParseAnthropicResponse(t *testing.T) {
	response := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "Hello from Claude!",
			},
		},
		"model":       "claude-3-opus-20240229",
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  float64(10),
			"output_tokens": float64(5),
		},
	}

	body, _ := json.Marshal(response)
	result, err := parseAnthropicResponse(body)
	require.NoError(t, err)
	assert.Equal(t, "Hello from Claude!", result["content"])
	assert.Equal(t, "claude-3-opus-20240229", result["model"])
	assert.Equal(t, int64(15), result["tokens_used"])
}

// TestParseSSEStream verifies SSE stream parsing
func TestParseSSEStream(t *testing.T) {
	stream := `data: {"choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":" world"}}]}

data: [DONE]
`

	chunks, err := parseSSEStream([]byte(stream))
	require.NoError(t, err)
	assert.Len(t, chunks, 3)
	assert.Equal(t, "Hello", chunks[0]["content"])
	assert.Equal(t, " world", chunks[1]["content"])
	assert.True(t, chunks[2]["done"].(bool))
}

// TestHandlerOpenAIComplete tests completion against a mock HTTP server
func TestHandlerOpenAIComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		resp := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "Hello from test!",
						"role":    "assistant",
					},
					"finish_reason": "stop",
				},
			},
			"model": "gpt-4",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	h := NewHandlerWithBaseURL("openai", "test-key", server.URL)

	result, err := h.Complete(map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hi",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello from test!", result["content"])
}

// TestHandlerAnthropicComplete tests Anthropic completion with mock server
func TestHandlerAnthropicComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		resp := map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Hello from Claude!",
				},
			},
			"model":       "claude-3-opus",
			"stop_reason": "end_turn",
			"usage":       map[string]interface{}{"input_tokens": float64(5), "output_tokens": float64(3)},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	h := NewHandlerWithBaseURL("anthropic", "test-key", server.URL)

	result, err := h.Complete(map[string]interface{}{
		"model": "claude-3-opus",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hi",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello from Claude!", result["content"])
}

// TestHandlerUnsupportedProvider verifies error on unknown provider
func TestHandlerUnsupportedProvider(t *testing.T) {
	h := NewHandler("unknown", "key")

	_, err := h.Complete(map[string]interface{}{"model": "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

// TestTokenCountEdgeCases verifies token count edge cases
func TestTokenCountEdgeCases(t *testing.T) {
	mock := NewMockHandler()

	// Short string
	result, err := mock.TokenCount("Hi")
	require.NoError(t, err)
	assert.Equal(t, int64(1), result["estimate"])

	// Empty string
	result, err = mock.TokenCount("")
	require.NoError(t, err)
	assert.Equal(t, int64(0), result["estimate"])

	// Invalid type
	_, err = mock.TokenCount(42)
	assert.Error(t, err)
}
