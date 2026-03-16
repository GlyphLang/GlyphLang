package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler()
	assert.NotNil(t, h)
	assert.NotNil(t, h.client)
	assert.Equal(t, 30*time.Second, h.client.Timeout)
}

func TestNewHandlerWithTimeout(t *testing.T) {
	h := NewHandlerWithTimeout(10 * time.Second)
	assert.Equal(t, 10*time.Second, h.client.Timeout)
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/data", r.URL.Path)
		w.Header().Set("X-Custom", "test-value")
		w.WriteHeader(200)
		err := json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(server.URL + "/data")
	require.NoError(t, err)
	assert.Equal(t, int64(200), result["status"])
	assert.Equal(t, true, result["ok"])
	assert.Contains(t, result["body"].(string), "hello")

	headers := result["headers"].(map[string]interface{})
	assert.Equal(t, "test-value", headers["X-Custom"])
}

func TestGetWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer abc123", r.Header.Get("Authorization"))
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		w.WriteHeader(200)
		_, err := w.Write([]byte("ok"))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(map[string]interface{}{
		"url": server.URL + "/search",
		"headers": map[string]interface{}{
			"Authorization": "Bearer abc123",
		},
		"query": map[string]interface{}{
			"foo": "bar",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(200), result["status"])
	assert.Equal(t, "ok", result["body"])
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)
		assert.Equal(t, "test", payload["name"])

		w.WriteHeader(201)
		_, err = w.Write([]byte(`{"id": 1}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Post(map[string]interface{}{
		"url": server.URL + "/create",
		"body": map[string]interface{}{
			"name": "test",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(201), result["status"])
	assert.Equal(t, true, result["ok"])
}

func TestPostStringBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "raw string body", string(body))
		w.WriteHeader(200)
		_, err = w.Write([]byte("ok"))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Post(map[string]interface{}{
		"url":  server.URL,
		"body": "raw string body",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(200), result["status"])
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(200)
		_, err := w.Write([]byte("updated"))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Put(map[string]interface{}{
		"url":  server.URL + "/update",
		"body": map[string]interface{}{"name": "updated"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(200), result["status"])
}

func TestPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(200)
		_, err := w.Write([]byte("patched"))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Patch(map[string]interface{}{
		"url":  server.URL + "/patch",
		"body": map[string]interface{}{"field": "value"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(200), result["status"])
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Delete(server.URL + "/resource/1")
	require.NoError(t, err)
	assert.Equal(t, int64(204), result["status"])
	assert.Equal(t, true, result["ok"])
}

func TestRedirectFollowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/destination", http.StatusFound)
			return
		}
		w.WriteHeader(200)
		_, err := w.Write([]byte("final"))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(server.URL + "/redirect")
	require.NoError(t, err)
	assert.Equal(t, int64(200), result["status"])
	assert.Equal(t, "final", result["body"])
}

func TestRedirectDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/destination", http.StatusFound)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(map[string]interface{}{
		"url":             server.URL + "/redirect",
		"followRedirects": false,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(302), result["status"])
	assert.Equal(t, false, result["ok"])
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHandler()
	_, err := h.Get(map[string]interface{}{
		"url":     server.URL,
		"timeout": int64(50),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}

func TestErrorResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, err := w.Write([]byte("not found"))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(server.URL + "/missing")
	require.NoError(t, err)
	assert.Equal(t, int64(404), result["status"])
	assert.Equal(t, false, result["ok"])
	assert.Equal(t, "not found", result["body"])
}

func TestInvalidURL(t *testing.T) {
	h := NewHandler()
	_, err := h.Get("ftp://invalid.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only http and https")
}

func TestInvalidArgs(t *testing.T) {
	h := NewHandler()
	_, err := h.Get(42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected string URL or object")
}

func TestMissingURLField(t *testing.T) {
	h := NewHandler()
	_, err := h.Get(map[string]interface{}{
		"headers": map[string]interface{}{},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 'url' field")
}

func TestMultipleResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(server.URL)
	require.NoError(t, err)

	headers := result["headers"].(map[string]interface{})
	cookies, ok := headers["Set-Cookie"].([]interface{})
	require.True(t, ok)
	assert.Len(t, cookies, 2)
}

func TestParseRequestArgsString(t *testing.T) {
	reqURL, opts, err := parseRequestArgs("https://example.com")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", reqURL)
	assert.NotNil(t, opts)
}

func TestParseRequestArgsMap(t *testing.T) {
	reqURL, opts, err := parseRequestArgs(map[string]interface{}{
		"url":             "https://example.com/api",
		"headers":         map[string]interface{}{"X-Key": "val"},
		"body":            "data",
		"query":           map[string]interface{}{"q": "search"},
		"timeout":         int64(5000),
		"followRedirects": false,
	})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/api", reqURL)
	assert.Equal(t, "val", opts.Headers["X-Key"])
	assert.Equal(t, "data", opts.Body)
	assert.Equal(t, "search", opts.Query["q"])
	assert.Equal(t, 5*time.Second, opts.Timeout)
	assert.NotNil(t, opts.FollowRedirects)
	assert.False(t, *opts.FollowRedirects)
}

func TestValidateURL(t *testing.T) {
	assert.NoError(t, validateURL("http://example.com"))
	assert.NoError(t, validateURL("https://example.com"))
	assert.Error(t, validateURL("ftp://example.com"))
	assert.Error(t, validateURL("://bad"))
}

func TestLargeResponse(t *testing.T) {
	largeBody := strings.Repeat("x", 1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, err := w.Write([]byte(largeBody))
		require.NoError(t, err)
	}))
	defer server.Close()

	h := NewHandler()
	result, err := h.Get(server.URL)
	require.NoError(t, err)
	assert.Equal(t, largeBody, result["body"])
}

func TestFloatTimeout(t *testing.T) {
	_, opts, err := parseRequestArgs(map[string]interface{}{
		"url":     "https://example.com",
		"timeout": float64(1000),
	})
	require.NoError(t, err)
	assert.Equal(t, 1*time.Second, opts.Timeout)
}
