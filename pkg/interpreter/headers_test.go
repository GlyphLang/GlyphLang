package interpreter

import (
	"testing"

	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInterpreter_HeadersAccessible verifies that route handlers can read
// request headers via the 'headers' built-in variable (#241).
func TestInterpreter_HeadersAccessible(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/api/echo",
		Method: Get,
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "auth", Value: FieldAccessExpr{
							Object: VariableExpr{Name: "headers"},
							Field:  "Authorization",
						}},
					},
				},
			},
		},
	}

	request := &Request{
		Path:   "/api/echo",
		Method: "GET",
		Headers: map[string]string{
			"Authorization": "Bearer tok123",
			"Content-Type":  "application/json",
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)

	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Bearer tok123", body["auth"])
}

// TestInterpreter_HeadersMissing verifies that accessing an absent header
// returns nil rather than an error.
func TestInterpreter_HeadersMissing(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/api/echo",
		Method: Get,
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "val", Value: FieldAccessExpr{
							Object: VariableExpr{Name: "headers"},
							Field:  "X-Missing",
						}},
					},
				},
			},
		},
	}

	request := &Request{
		Path:    "/api/echo",
		Method:  "GET",
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)

	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Nil(t, body["val"])
}
