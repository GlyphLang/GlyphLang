" Vim syntax file
" Language: Glyph
" Maintainer: GlyphLang Team
" Latest Revision: 2024

if exists("b:current_syntax")
  finish
endif

" Comments
syn match glyphComment "#.*$" contains=glyphTodo
syn keyword glyphTodo TODO FIXME XXX NOTE contained

" Strings
syn region glyphString start=/"/ skip=/\\"/ end=/"/ contains=glyphEscape
syn match glyphEscape "\\." contained

" Numbers
syn match glyphNumber "\<\d\+\>"
syn match glyphFloat "\<\d\+\.\d*\>"

" Booleans and null
syn keyword glyphBoolean true false
syn keyword glyphNull null

" HTTP methods (route definitions)
syn keyword glyphMethod GET POST PUT DELETE PATCH OPTIONS HEAD
syn keyword glyphMethod ws

" Route definition
syn match glyphRouteMarker "^@" nextgroup=glyphMethod skipwhite
syn match glyphPath "/[a-zA-Z0-9_/:*-]*" contained

" Type definition
syn match glyphTypeMarker "^:" nextgroup=glyphTypeName skipwhite
syn match glyphTypeName "\<[A-Z][a-zA-Z0-9_]*\>" contained

" Variable declaration
syn match glyphVarDecl "\$\s*\w\+" contains=glyphVarMarker,glyphIdentifier
syn match glyphVarMarker "\$" contained

" Return statement
syn match glyphReturn ">\s*"

" Decorators/Middleware
syn match glyphDecorator "+\s*\w\+\s*([^)]*)" contains=glyphDecoratorMarker,glyphDecoratorName
syn match glyphDecoratorMarker "+" contained
syn match glyphDecoratorName "\w\+" contained

" Dependency injection
syn match glyphInjection "%\s*\w\+\s*:" contains=glyphInjectionMarker
syn match glyphInjectionMarker "%" contained

" Validation
syn match glyphValidation "?\s*\w\+" contains=glyphValidationMarker
syn match glyphValidationMarker "?" contained

" Keywords
syn keyword glyphKeyword if else while for in switch case default
syn keyword glyphKeyword on connect message disconnect
syn keyword glyphKeyword fn return

" Built-in types
syn keyword glyphType int str bool float timestamp List Map
syn match glyphTypeRequired "!" contained

" Operators
syn match glyphOperator "[-+*/%=<>!&|]"
syn match glyphOperator "=="
syn match glyphOperator "!="
syn match glyphOperator "<="
syn match glyphOperator ">="
syn match glyphOperator "&&"
syn match glyphOperator "||"

" Brackets and delimiters
syn match glyphDelimiter "[{}()\[\]:,]"

" Function calls
syn match glyphFunctionCall "\<\w\+\s*(" contains=glyphFunction
syn match glyphFunction "\<\w\+\>" contained

" Field access
syn match glyphFieldAccess "\.\w\+"

" Highlighting links
hi def link glyphComment Comment
hi def link glyphTodo Todo
hi def link glyphString String
hi def link glyphEscape SpecialChar
hi def link glyphNumber Number
hi def link glyphFloat Float
hi def link glyphBoolean Boolean
hi def link glyphNull Constant
hi def link glyphMethod Keyword
hi def link glyphRouteMarker Special
hi def link glyphPath String
hi def link glyphTypeMarker Special
hi def link glyphTypeName Type
hi def link glyphVarDecl Identifier
hi def link glyphVarMarker Special
hi def link glyphReturn Special
hi def link glyphDecorator PreProc
hi def link glyphDecoratorMarker Special
hi def link glyphDecoratorName Function
hi def link glyphInjection PreProc
hi def link glyphInjectionMarker Special
hi def link glyphValidation PreProc
hi def link glyphValidationMarker Special
hi def link glyphKeyword Keyword
hi def link glyphType Type
hi def link glyphTypeRequired Special
hi def link glyphOperator Operator
hi def link glyphDelimiter Delimiter
hi def link glyphFunction Function
hi def link glyphFieldAccess Identifier

let b:current_syntax = "glyph"
