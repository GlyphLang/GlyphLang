package sse

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter_SetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, err := NewWriter(rec)
	require.NoError(t, err)
	require.NotNil(t, sw)

	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWriter_SendData_String(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.SendData("hello world")
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "data: hello world\n")
	assert.Contains(t, body, "id: 1\n")
}

func TestWriter_SendData_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.SendData(map[string]interface{}{"count": 42})
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, `data: {"count":42}`)
}

func TestWriter_Send_WithEventType(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.Send(Event{
		Type: "notification",
		Data: "new message",
	})
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "event: notification\n")
	assert.Contains(t, body, "data: new message\n")
}

func TestWriter_Send_WithCustomID(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.Send(Event{
		ID:   "custom-123",
		Data: "test",
	})
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "id: custom-123\n")
}

func TestWriter_Send_WithRetry(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.Send(Event{
		Data:  "reconnect test",
		Retry: 5000,
	})
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "retry: 5000\n")
}

func TestWriter_Send_AutoIncrementID(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	_ = sw.SendData("first")
	_ = sw.SendData("second")
	_ = sw.SendData("third")

	body := rec.Body.String()
	assert.Contains(t, body, "id: 1\n")
	assert.Contains(t, body, "id: 2\n")
	assert.Contains(t, body, "id: 3\n")
}

func TestWriter_SendComment(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.SendComment("keep-alive")
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, ": keep-alive\n")
}

func TestWriter_Send_MultilineData(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	err := sw.SendData("line1\nline2\nline3")
	require.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "data: line1\n")
	assert.Contains(t, body, "data: line2\n")
	assert.Contains(t, body, "data: line3\n")
}

func TestWriter_Send_EndsWithBlankLine(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	_ = sw.SendData("test")

	body := rec.Body.String()
	// Events must end with \n\n (blank line separates events)
	assert.True(t, strings.HasSuffix(body, "\n\n"))
}

func TestWriter_MultipleEvents(t *testing.T) {
	rec := httptest.NewRecorder()
	sw, _ := NewWriter(rec)

	_ = sw.Send(Event{Type: "start", Data: "beginning"})
	_ = sw.Send(Event{Type: "progress", Data: map[string]interface{}{"percent": 50}})
	_ = sw.Send(Event{Type: "end", Data: "done"})

	body := rec.Body.String()
	assert.Contains(t, body, "event: start\n")
	assert.Contains(t, body, "event: progress\n")
	assert.Contains(t, body, "event: end\n")
	assert.Contains(t, body, "data: beginning\n")
	assert.Contains(t, body, "data: done\n")
}
