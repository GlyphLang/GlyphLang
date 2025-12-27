# GlyphLang Support for VS Code

This extension provides comprehensive language support for GlyphLang (formerly AIBC) - an AI-focused backend language with native compilation.

## Features

### Core Language Support
- **Syntax Highlighting**: Full syntax highlighting for GlyphLang constructs (.glyph and .glybc files)
- **IntelliSense**: Smart code completion for keywords, types, and snippets
- **Hover Information**: Type information and documentation on hover
- **Go to Definition**: Jump to type, variable, and route definitions
- **Find References**: Find all references to symbols across your code
- **Document Symbols**: Outline view showing types, routes, and functions
- **Diagnostics**: Real-time error checking and validation

### Optimizer Integration
- **Optimization Hints**: Real-time suggestions for code optimizations
  - Constant folding opportunities
  - Algebraic simplifications (x + 0, x * 1)
  - Strength reduction (x * 2 → x + x)
  - Loop invariant code motion
- **Configurable Optimization Levels**: Choose from O0 (none) to O3 (aggressive)

## Requirements

You must have the Glyph compiler installed and accessible in your PATH:

```bash
go install github.com/GlyphLang/GlyphLang/cmd/glyph@latest
```

Or specify the path to the Glyph executable in settings:

```json
{
  "glyph.lsp.path": "/path/to/glyph"
}
```

## Extension Settings

This extension contributes the following settings:

### Language Server
* `glyph.lsp.path`: Path to the Glyph LSP server executable (default: `glyph`)
* `glyph.lsp.logFile`: Path to LSP log file for debugging (optional)

### Compiler
* `glyph.compiler.optimizationLevel`: Compiler optimization level (default: `O2`)
  - `O0`: No optimization - fastest compilation
  - `O1`: Basic optimization - constant folding only
  - `O2`: Standard optimization - constant folding, algebraic simplification, dead code elimination
  - `O3`: Aggressive optimization - all optimizations including strength reduction and LICM

### Diagnostics
* `glyph.diagnostics.showOptimizerHints`: Show optimization hints in the editor (default: `true`)

## Usage

1. Open any `.glyph` or `.glybc` file
2. The extension will automatically start the GlyphLang Language Server
3. You'll get syntax highlighting, completions, diagnostics, and optimizer hints

### Keyboard Shortcuts

- **Go to Definition**: F12 or Ctrl+Click
- **Find References**: Shift+F12
- **Show Document Symbols**: Ctrl+Shift+O

## Example

```glyph
# Define a User type
: User {
  id: int!
  name: str!
  email: str!
  age: int!
}

# Define a REST API route with optimization opportunities
@route /api/users/:id [GET]
  $ user_id = :id
  $ base_age = 25

  # The LSP will suggest optimizations for these:
  $ constant_result = 2 + 3          # → Constant folding hint
  $ redundant = base_age + 0         # → Algebraic simplification hint
  $ doubled = base_age * 2           # → Strength reduction hint

  $ counter = 0
  while counter < 10 {
    $ invariant = 100 * 5            # → Loop invariant hint
    $ counter = counter + 1
  }

  > {
    id: user_id,
    age: constant_result,
    doubled: doubled
  }
```

## Language Server Protocol

This extension uses the Language Server Protocol (LSP) to provide intelligent language features. The LSP server is implemented in Go and runs as a separate process.

## Debugging

To enable LSP logging for debugging:

```json
{
  "glyph.lsp.logFile": "/tmp/glyph-lsp.log"
}
```

## Release Notes

### 1.0.0

Production-ready release with comprehensive language support:

**Language Features:**
- Full syntax highlighting for .glyph and .glybc files
- IntelliSense with smart code completion
- Hover information for types and variables
- Go to definition for types, variables, and routes
- Find all references across codebase
- Document symbols / outline view

**Optimizer Integration:**
- Real-time optimization hints
- Configurable optimization levels (O0-O3)
- Suggestions for:
  - Constant folding
  - Algebraic simplifications
  - Strength reduction
  - Loop invariant code motion

**Configuration:**
- Customizable LSP server path
- Debug logging support
- Toggle optimizer hints on/off

## Contributing

Contributions are welcome! Please visit https://github.com/GlyphLang/GlyphLang

## License

Apache License 2.0 - see [LICENSE](https://github.com/GlyphLang/GlyphLang/blob/main/LICENSE) for details.
