package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPValidateSource(t *testing.T) {
	_, result, err := mcpValidate(context.Background(), nil, validateInput{
		Source: "@ GET /hello {\n  > {message: \"hi\"}\n}\n",
	})
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, 1, result.Stats.Routes)
}

func TestMCPValidateInvalidSource(t *testing.T) {
	_, result, err := mcpValidate(context.Background(), nil, validateInput{
		Source: "@ GET /broken {\n  > \n",
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestMCPValidateNoInput(t *testing.T) {
	_, _, err := mcpValidate(context.Background(), nil, validateInput{})
	assert.Error(t, err)
}

func TestMCPSyntaxReference(t *testing.T) {
	_, out, err := mcpSyntax(context.Background(), nil, syntaxInput{})
	require.NoError(t, err)
	assert.Contains(t, out.Spec, "GlyphLang Notation Specification")
}
