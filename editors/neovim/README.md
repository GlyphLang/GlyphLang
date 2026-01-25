# Neovim Support for Glyph

This directory contains Neovim configuration files for Glyph syntax highlighting and LSP support.

## Installation

### Manual Installation

1. Copy the files to your Neovim runtime path:

```bash
# Create directories if they don't exist
mkdir -p ~/.config/nvim/syntax
mkdir -p ~/.config/nvim/ftdetect
mkdir -p ~/.config/nvim/ftplugin
mkdir -p ~/.config/nvim/queries/glyph

# Copy files
cp syntax/glyph.vim ~/.config/nvim/syntax/
cp ftdetect/glyph.vim ~/.config/nvim/ftdetect/
cp ftplugin/glyph.vim ~/.config/nvim/ftplugin/
cp queries/glyph/highlights.scm ~/.config/nvim/queries/glyph/
```

2. Add LSP configuration to your `init.lua`:

```lua
-- Option 1: Copy lsp.lua to your lua directory and require it
require('glyph-lsp').setup()

-- Option 2: Inline configuration
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.glyph then
  configs.glyph = {
    default_config = {
      cmd = { 'glyph', 'lsp' },
      filetypes = { 'glyph' },
      root_dir = lspconfig.util.root_pattern('.git', 'go.mod', 'main.glyph'),
      single_file_support = true,
    },
  }
end

lspconfig.glyph.setup({
  on_attach = function(client, bufnr)
    -- Your on_attach configuration
  end,
})
```

### Using lazy.nvim

Add to your plugin specifications:

```lua
{
  "GlyphLang/GlyphLang",
  config = function()
    -- Add runtime path for syntax files
    vim.opt.runtimepath:append(vim.fn.stdpath("data") .. "/lazy/GlyphLang/editors/neovim")
    
    -- Setup LSP
    dofile(vim.fn.stdpath("data") .. "/lazy/GlyphLang/editors/neovim/lsp.lua").setup()
  end,
}
```

### Using packer.nvim

```lua
use {
  'GlyphLang/GlyphLang',
  config = function()
    vim.opt.runtimepath:append(vim.fn.stdpath("data") .. "/site/pack/packer/start/GlyphLang/editors/neovim")
    dofile(vim.fn.stdpath("data") .. "/site/pack/packer/start/GlyphLang/editors/neovim/lsp.lua").setup()
  end
}
```

## Requirements

- Neovim >= 0.8.0
- [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig) for LSP support
- `glyph` CLI installed and in PATH

## Features

### Syntax Highlighting

The syntax file provides highlighting for:
- Comments (`#`)
- Route definitions (`@ GET /path`)
- Type definitions (`: TypeName`)
- Variable declarations (`$ var = value`)
- Variable reassignments (`var = value`)
- Return statements (`> expression`)
- Decorators/middleware (`+ auth(jwt)`)
- Dependency injection (`% db: Database`)
- Validation (`? validate()`)
- Control flow (`if`, `else`, `while`, `for`, `switch`)
- WebSocket events (`on connect`, `on message`, `on disconnect`)
- Strings, numbers, booleans
- Operators and delimiters

### LSP Features

When the Glyph LSP is configured:
- Go to definition (`gd`)
- Find references (`gr`)
- Hover documentation (`K`)
- Code actions (`<leader>ca`)
- Rename symbol (`<leader>rn`)
- Format document (`<leader>f`)

### File Type Settings

The ftplugin sets sensible defaults:
- 2-space indentation
- Comment string set to `#`
- Syntax-based folding

## Keybindings

Default LSP keybindings (can be customized in setup):

| Key | Action |
|-----|--------|
| `gd` | Go to definition |
| `gD` | Go to declaration |
| `gr` | Find references |
| `gi` | Go to implementation |
| `K` | Hover documentation |
| `<C-k>` | Signature help |
| `<leader>rn` | Rename symbol |
| `<leader>ca` | Code actions |
| `<leader>f` | Format document |

## Troubleshooting

### LSP not starting

1. Ensure `glyph` is installed and in PATH:
   ```bash
   glyph --version
   ```

2. Check LSP logs:
   ```vim
   :LspLog
   ```

3. Verify the LSP is attached:
   ```vim
   :LspInfo
   ```

### Syntax highlighting not working

1. Check filetype is detected:
   ```vim
   :set filetype?
   ```

2. Ensure syntax is enabled:
   ```vim
   :syntax on
   ```

## Contributing

Issues and pull requests are welcome at [GlyphLang/GlyphLang](https://github.com/GlyphLang/GlyphLang).
