# GlyphLang VS Code Extension

Syntax highlighting and language support for Glyph (`.glyph`) and expanded Glyph (`.glyphx`) files.

## Features

- Syntax highlighting for `.glyph` (compact symbol syntax)
- Syntax highlighting for `.glyphx` (expanded keyword syntax)
- LSP integration for diagnostics, hover, completions, and go-to-definition
- Auto-bracket closing and indentation
- Code folding
- Commands to convert between `.glyph` and `.glyphx` formats

## Keyword Mapping

The `.glyphx` format uses human-readable keywords instead of symbols:

| Symbol (`.glyph`) | Keyword (`.glyphx`) | Purpose              |
|--------------------|---------------------|----------------------|
| `@`                | `route`             | Route definition     |
| `:`                | `type`              | Type definition      |
| `$`                | `let`               | Variable declaration |
| `>`                | `return`            | Return statement     |
| `+`                | `middleware`         | Middleware           |
| `%`                | `use`               | Service injection    |
| `~`                | `handle`            | Event handler        |
| `*`                | `cron`              | Scheduled task       |
| `!`                | `command`            | CLI command          |
| `&`                | `queue`             | Queue worker         |
| `=`                | `func`              | Function definition  |
| `?`                | `validate`          | Validation           |
| `<`                | `expects`           | Expected input       |

## Installation

### From Source

1. Copy this `vscode` directory to your VS Code extensions folder:

   ```bash
   # macOS/Linux
   cp -r editors/vscode ~/.vscode/extensions/glyphlang

   # Windows
   xcopy /E editors\vscode %USERPROFILE%\.vscode\extensions\glyphlang
   ```

2. Restart VS Code

### LSP Setup

The extension uses the Glyph CLI's built-in language server. Ensure `glyph` is in your PATH:

```bash
go install github.com/glyphlang/glyph/cmd/glyph@latest
```

The LSP server is started automatically when opening `.glyph` or `.glyphx` files.

## Configuration

| Setting                    | Default  | Description                              |
|----------------------------|----------|------------------------------------------|
| `glyph.lsp.enabled`       | `true`   | Enable the Glyph Language Server         |
| `glyph.lsp.path`          | `glyph`  | Path to the glyph CLI binary             |
| `glyph.autoExpandOnOpen`  | `false`  | Auto-expand .glyph to .glyphx on open    |
| `glyph.autoCompactOnSave` | `false`  | Auto-compact .glyphx to .glyph on save   |

## Commands

- **Glyph: Expand to .glyphx** - Convert current `.glyph` file to expanded syntax
- **Glyph: Compact to .glyph** - Convert current `.glyphx` file to compact syntax
