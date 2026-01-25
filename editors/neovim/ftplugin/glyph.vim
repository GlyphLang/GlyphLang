" Filetype plugin for Glyph files
" Language: Glyph

if exists("b:did_ftplugin")
  finish
endif
let b:did_ftplugin = 1

" Set local options
setlocal commentstring=#\ %s
setlocal comments=:#
setlocal tabstop=2
setlocal shiftwidth=2
setlocal expandtab
setlocal autoindent
setlocal smartindent

" Folding
setlocal foldmethod=syntax
setlocal foldlevel=99

" Match pairs for % navigation
setlocal matchpairs+=<:>

" Undo settings when switching filetype
let b:undo_ftplugin = "setlocal commentstring< comments< tabstop< shiftwidth< expandtab< autoindent< smartindent< foldmethod< foldlevel< matchpairs<"
