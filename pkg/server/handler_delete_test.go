package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandlerDELETEWithJSONBody verifies DELETE handlers parse JSON bodies.
// RFC 7231 permits DELETE to carry a request body; see issue #239.
func TestHandlerDELETEWithJSONBody(t *testing.T) {
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: DELETE,
		Path:   "/api/games/:rc",
	})

	requestBody := map[string]interface{}{
		"session_token": "tok1",
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	req := httptest.NewRequest("DELETE", "/api/games/ABCD", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	bodyData, ok := response["body"].(map[string]interface{})
	require.True(t, ok, "expected DELETE request body to be parsed")
	assert.Equal(t, "tok1", bodyData["session_token"])
}

// TestHandlerDELETEWithEmptyBody verifies DELETE with no body does not error.
func TestHandlerDELETEWithEmptyBody(t *testing.T) {
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: DELETE,
		Path:   "/api/games/:rc",
	})

	req := httptest.NewRequest("DELETE", "/api/games/ABCD", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	_, hasBody := response["body"]
	assert.False(t, hasBody, "expected no body field for empty DELETE body")
}

// TestHandlerDELETENonJSONContentType verifies DELETE with a non-JSON
// Content-Type returns 400, matching POST behavior.
func TestHandlerDELETENonJSONContentType(t *testing.T) {
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: DELETE,
		Path:   "/api/games/:rc",
	})

	req := httptest.NewRequest("DELETE", "/api/games/ABCD", bytes.NewBufferString("some text"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	errVal, ok := response["error"].(bool)
	require.True(t, ok)
	assert.True(t, errVal)
	assert.Contains(t, response["message"], "invalid JSON body")
}
