package llm

import (
	"fmt"
	"strings"
)

// MockHandler provides a mock LLM handler for testing and demos.
// All methods return deterministic responses without making API calls.
type MockHandler struct {
	completionResponses map[string]string
	defaultResponse     string
	callLog             []MockCall
}

// MockCall records a method call for verification in tests
type MockCall struct {
	Method string
	Args   []interface{}
}

// NewMockHandler creates a new mock LLM handler with default responses
func NewMockHandler() *MockHandler {
	return &MockHandler{
		completionResponses: map[string]string{},
		defaultResponse:     "This is a mock LLM response.",
		callLog:             []MockCall{},
	}
}

// OnComplete sets a canned response for a given model
func (m *MockHandler) OnComplete(model string, response string) {
	m.completionResponses[model] = response
}

// SetDefaultResponse sets the default response for all completions
func (m *MockHandler) SetDefaultResponse(response string) {
	m.defaultResponse = response
}

// Calls returns the call log for verification
func (m *MockHandler) Calls() []MockCall {
	return m.callLog
}

// CalledCount returns the number of times a method was called
func (m *MockHandler) CalledCount(method string) int {
	count := 0
	for _, call := range m.callLog {
		if call.Method == method {
			count++
		}
	}
	return count
}

// Complete returns a mock completion response
func (m *MockHandler) Complete(request interface{}) (map[string]interface{}, error) {
	m.callLog = append(m.callLog, MockCall{Method: "Complete", Args: []interface{}{request}})

	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for completion request")
	}

	model, _ := reqMap["model"].(string)

	response := m.defaultResponse
	if canned, ok := m.completionResponses[model]; ok {
		response = canned
	}

	// Build response from messages if present
	if messages, ok := reqMap["messages"].([]interface{}); ok {
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if role, ok := msgMap["role"].(string); ok && role == "user" {
					if content, ok := msgMap["content"].(string); ok {
						response = "Mock response to: " + content
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"content":       response,
		"model":         model,
		"finish_reason": "stop",
		"tokens_used":   int64(len(response) / 4),
	}, nil
}

// Chat is an alias for Complete
func (m *MockHandler) Chat(request interface{}) (map[string]interface{}, error) {
	m.callLog = append(m.callLog, MockCall{Method: "Chat", Args: []interface{}{request}})
	return m.Complete(request)
}

// Stream returns a mock streaming response as collected chunks
func (m *MockHandler) Stream(request interface{}) ([]map[string]interface{}, error) {
	m.callLog = append(m.callLog, MockCall{Method: "Stream", Args: []interface{}{request}})

	result, err := m.Complete(request)
	if err != nil {
		return nil, err
	}

	content, _ := result["content"].(string)
	words := strings.Fields(content)

	var chunks []map[string]interface{}
	for i, word := range words {
		chunks = append(chunks, map[string]interface{}{
			"content": word + " ",
			"done":    false,
		})
		_ = i
	}
	chunks = append(chunks, map[string]interface{}{
		"content": "",
		"done":    true,
	})

	return chunks, nil
}

// Embed returns a mock embedding vector
func (m *MockHandler) Embed(request interface{}) (map[string]interface{}, error) {
	m.callLog = append(m.callLog, MockCall{Method: "Embed", Args: []interface{}{request}})

	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for embedding request")
	}

	model, _ := reqMap["model"].(string)
	input, _ := reqMap["input"].(string)

	// Generate a deterministic mock vector based on input length
	vectorSize := 8
	vector := make([]interface{}, vectorSize)
	for i := range vector {
		vector[i] = float64(len(input)+i) / 100.0
	}

	return map[string]interface{}{
		"vector": vector,
		"model":  model,
	}, nil
}

// ListModels returns a mock list of available models
func (m *MockHandler) ListModels() ([]map[string]interface{}, error) {
	m.callLog = append(m.callLog, MockCall{Method: "ListModels"})

	return []map[string]interface{}{
		{"id": "gpt-4", "created": int64(1687882410)},
		{"id": "gpt-3.5-turbo", "created": int64(1677610602)},
		{"id": "text-embedding-ada-002", "created": int64(1671217299)},
	}, nil
}

// TokenCount estimates token count for text
func (m *MockHandler) TokenCount(text interface{}) (map[string]interface{}, error) {
	m.callLog = append(m.callLog, MockCall{Method: "TokenCount", Args: []interface{}{text}})

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
