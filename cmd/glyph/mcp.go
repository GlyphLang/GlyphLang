package main

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/glyphlang/glyph/docs"
	"github.com/glyphlang/glyph/pkg/validate"
)

// validateInput is the MCP tool input: inline source or a file path.
type validateInput struct {
	Source string `json:"source,omitempty" jsonschema:"GlyphLang source code to validate"`
	Path   string `json:"path,omitempty" jsonschema:"path to a .glyph file to validate (alternative to source)"`
}

type syntaxInput struct{}

type syntaxOutput struct {
	Spec string `json:"spec" jsonschema:"the full GlyphLang notation specification in Markdown"`
}

func mcpValidate(_ context.Context, _ *mcp.CallToolRequest, in validateInput) (*mcp.CallToolResult, *validate.ValidationResult, error) {
	source := in.Source
	path := in.Path
	if source == "" && path == "" {
		return nil, nil, fmt.Errorf("provide either source or path")
	}
	if source == "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", path, err)
		}
		source = string(data)
	}
	if path == "" {
		path = "input.glyph"
	}
	return nil, validate.NewValidator(source, path).Validate(), nil
}

func mcpSyntax(_ context.Context, _ *mcp.CallToolRequest, _ syntaxInput) (*mcp.CallToolResult, syntaxOutput, error) {
	return nil, syntaxOutput{Spec: docs.NotationSpec}, nil
}

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start a Model Context Protocol server over stdio",
		Long: `Starts an MCP server over stdio exposing GlyphLang tools to AI agents:

  validate_glyph    validate GlyphLang source and return structured errors
  syntax_reference  return the full GlyphLang notation specification

Register the server in an MCP-capable agent (e.g. Claude Code) with:
  claude mcp add glyphlang -- glyph mcp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := mcp.NewServer(&mcp.Implementation{Name: "glyphlang", Version: version}, nil)
			mcp.AddTool(server, &mcp.Tool{
				Name: "validate_glyph",
				Description: "Validate GlyphLang source code. Call this after writing or editing any " +
					"GlyphLang code, before presenting it, and fix every reported error. Returns " +
					"structured errors with locations and fix hints.",
			}, mcpValidate)
			mcp.AddTool(server, &mcp.Tool{
				Name: "syntax_reference",
				Description: "Get the full GlyphLang notation specification. Call this before writing " +
					"GlyphLang code if the syntax is not already in context.",
			}, mcpSyntax)
			return server.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}
}
