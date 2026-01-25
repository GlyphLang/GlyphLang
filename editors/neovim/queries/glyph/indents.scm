; Glyph Tree-sitter Indentation Queries

; Indent after opening braces
[
  "{"
  "["
  "("
] @indent.begin

; Dedent at closing braces
[
  "}"
  "]"
  ")"
] @indent.end

; Indent after route definitions
(route_definition
  body: (_) @indent.begin)

; Indent after type definitions
(type_definition
  body: (_) @indent.begin)

; Indent after function definitions
(function_definition
  body: (_) @indent.begin)

; Indent after control flow
(if_statement
  consequence: (_) @indent.begin
  alternative: (_)? @indent.begin)

(while_statement
  body: (_) @indent.begin)

(for_statement
  body: (_) @indent.begin)

(switch_statement
  body: (_) @indent.begin)

; Indent after WebSocket event handlers
(websocket_event
  body: (_) @indent.begin)

; Branch nodes (else, case, default)
[
  "else"
  "case"
  "default"
] @indent.branch
