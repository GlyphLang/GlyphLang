// Package docs embeds documentation files needed at runtime,
// so the compiled binary can serve them without a source checkout.
package docs

import _ "embed"

// NotationSpec is the full GlyphLang notation specification, served to
// AI agents via the MCP syntax_reference tool.
//
//go:embed GLYPH_NOTATION_SPEC.md
var NotationSpec string
