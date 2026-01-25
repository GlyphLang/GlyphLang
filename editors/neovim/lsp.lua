-- Glyph LSP Configuration for Neovim
-- Add this to your Neovim configuration (init.lua or lua/plugins/lsp.lua)

local M = {}

-- Setup function to configure Glyph LSP
function M.setup(opts)
  opts = opts or {}

  -- Check if lspconfig is available
  local ok, lspconfig = pcall(require, 'lspconfig')
  if not ok then
    vim.notify("nvim-lspconfig is required for Glyph LSP support", vim.log.levels.ERROR)
    return
  end

  local configs = require('lspconfig.configs')

  -- Register Glyph LSP if not already registered
  if not configs.glyph then
    configs.glyph = {
      default_config = {
        cmd = { 'glyph', 'lsp' },
        filetypes = { 'glyph' },
        root_dir = lspconfig.util.root_pattern('.git', 'go.mod', 'main.glyph'),
        single_file_support = true,
        settings = {},
      },
    }
  end

  -- Setup the LSP with user options
  lspconfig.glyph.setup(vim.tbl_deep_extend('force', {
    on_attach = function(client, bufnr)
      -- Enable completion triggered by <c-x><c-o>
      vim.bo[bufnr].omnifunc = 'v:lua.vim.lsp.omnifunc'

      -- Buffer local mappings
      local bufopts = { noremap = true, silent = true, buffer = bufnr }
      vim.keymap.set('n', 'gD', vim.lsp.buf.declaration, bufopts)
      vim.keymap.set('n', 'gd', vim.lsp.buf.definition, bufopts)
      vim.keymap.set('n', 'K', vim.lsp.buf.hover, bufopts)
      vim.keymap.set('n', 'gi', vim.lsp.buf.implementation, bufopts)
      vim.keymap.set('n', '<C-k>', vim.lsp.buf.signature_help, bufopts)
      vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, bufopts)
      vim.keymap.set('n', '<leader>ca', vim.lsp.buf.code_action, bufopts)
      vim.keymap.set('n', 'gr', vim.lsp.buf.references, bufopts)
      vim.keymap.set('n', '<leader>f', function()
        vim.lsp.buf.format { async = true }
      end, bufopts)

      -- Call user's on_attach if provided
      if opts.on_attach then
        opts.on_attach(client, bufnr)
      end
    end,
    capabilities = opts.capabilities,
  }, opts.server or {}))
end

-- Filetype detection
vim.filetype.add({
  extension = {
    glyph = 'glyph',
    glyphx = 'glyph',
  },
})

return M
