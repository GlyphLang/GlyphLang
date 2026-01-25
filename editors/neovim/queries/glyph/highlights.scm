; Glyph Tree-sitter Highlights
; Note: This file is for future tree-sitter integration
; Currently, use the vim syntax file for highlighting

; Comments
(comment) @comment

; Strings
(string) @string
(escape_sequence) @string.escape

; Numbers
(number) @number
(float) @float

; Booleans
(true) @boolean
(false) @boolean
(null) @constant.builtin

; HTTP methods
(http_method) @keyword
(websocket_keyword) @keyword

; Route marker
(route_marker) @punctuation.special

; Path
(path) @string.special

; Type definitions
(type_marker) @punctuation.special
(type_name) @type
(type_annotation) @type

; Variable declaration
(variable_marker) @punctuation.special
(identifier) @variable

; Return statement
(return_marker) @punctuation.special

; Decorators
(decorator_marker) @punctuation.special
(decorator_name) @function.macro

; Injection
(injection_marker) @punctuation.special

; Validation
(validation_marker) @punctuation.special

; Keywords
[
  "if"
  "else"
  "while"
  "for"
  "in"
  "switch"
  "case"
  "default"
  "on"
  "connect"
  "message"
  "disconnect"
  "fn"
  "return"
] @keyword

; Types
[
  "int"
  "str"
  "bool"
  "float"
  "timestamp"
  "List"
  "Map"
] @type.builtin

; Operators
[
  "+"
  "-"
  "*"
  "/"
  "%"
  "="
  "=="
  "!="
  "<"
  ">"
  "<="
  ">="
  "&&"
  "||"
  "!"
] @operator

; Punctuation
[
  "{"
  "}"
  "("
  ")"
  "["
  "]"
] @punctuation.bracket

[
  ":"
  ","
  "."
] @punctuation.delimiter

; Function calls
(function_call
  name: (identifier) @function.call)

; Field access
(field_access
  field: (identifier) @property)
