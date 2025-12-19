package lsp

import (
	"fmt"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// GetDiagnostics returns diagnostics for a document
func GetDiagnostics(doc *Document) []Diagnostic {
	var diagnostics []Diagnostic

	// Convert parse errors to diagnostics
	for _, err := range doc.Errors {
		diagnostic := Diagnostic{
			Range: Range{
				Start: Position{
					Line:      err.Line - 1, // LSP is 0-based
					Character: err.Column - 1,
				},
				End: Position{
					Line:      err.Line - 1,
					Character: err.Column,
				},
			},
			Severity: DiagnosticSeverityError,
			Source:   "glyph",
			Message:  err.Message,
		}

		if err.Hint != "" {
			diagnostic.Message = fmt.Sprintf("%s\n\nHint: %s", err.Message, err.Hint)
		}

		diagnostics = append(diagnostics, diagnostic)
	}

	// Type checking errors (if AST is available)
	if doc.AST != nil {
		typeErrors := checkTypes(doc.AST)
		for _, err := range typeErrors {
			diagnostics = append(diagnostics, err)
		}

		// Add optimizer hints
		optimizerHints := getOptimizerHints(doc.AST)
		diagnostics = append(diagnostics, optimizerHints...)
	}

	return diagnostics
}

// GetHover returns hover information at a position
func GetHover(doc *Document, pos Position) *Hover {
	if doc.AST == nil {
		return nil
	}

	// Get word at position
	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Check if it's a keyword
	if info := getKeywordInfo(word); info != "" {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: info,
			},
		}
	}

	// Check if it's a type definition
	for _, item := range doc.AST.Items {
		if typeDef, ok := item.(*interpreter.TypeDef); ok {
			if typeDef.Name == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatTypeDefHover(typeDef),
					},
				}
			}
		}
	}

	// Check if it's a route
	for _, item := range doc.AST.Items {
		if route, ok := item.(*interpreter.Route); ok {
			// Check if position is in route path or body
			if strings.Contains(route.Path, word) {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatRouteHover(route),
					},
				}
			}
		}
	}

	// Check if it's a command (pointer or value type)
	for _, item := range doc.AST.Items {
		if cmd, ok := item.(*interpreter.Command); ok {
			if cmd.Name == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatCommandHover(cmd),
					},
				}
			}
		} else if cmd, ok := item.(interpreter.Command); ok {
			if cmd.Name == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatCommandHover(&cmd),
					},
				}
			}
		}
	}

	// Check if it's a cron task (pointer or value type)
	for _, item := range doc.AST.Items {
		if cron, ok := item.(*interpreter.CronTask); ok {
			if cron.Name == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatCronTaskHover(cron),
					},
				}
			}
		} else if cron, ok := item.(interpreter.CronTask); ok {
			if cron.Name == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatCronTaskHover(&cron),
					},
				}
			}
		}
	}

	// Check if it's an event handler (pointer or value type)
	for _, item := range doc.AST.Items {
		if event, ok := item.(*interpreter.EventHandler); ok {
			if event.EventType == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatEventHandlerHover(event),
					},
				}
			}
		} else if event, ok := item.(interpreter.EventHandler); ok {
			if event.EventType == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatEventHandlerHover(&event),
					},
				}
			}
		}
	}

	// Check if it's a queue worker (pointer or value type)
	for _, item := range doc.AST.Items {
		if queue, ok := item.(*interpreter.QueueWorker); ok {
			if queue.QueueName == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatQueueWorkerHover(queue),
					},
				}
			}
		} else if queue, ok := item.(interpreter.QueueWorker); ok {
			if queue.QueueName == word {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatQueueWorkerHover(&queue),
					},
				}
			}
		}
	}

	// Check if it's a built-in type
	if info := getBuiltInTypeInfo(word); info != "" {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: info,
			},
		}
	}

	return nil
}

// GetCompletion returns completion items at a position
func GetCompletion(doc *Document, pos Position) []CompletionItem {
	var items []CompletionItem

	// Add keywords
	keywords := []string{
		"route", "if", "else", "while", "for", "in", "switch", "case", "default",
		"true", "false",
	}

	for _, kw := range keywords {
		items = append(items, CompletionItem{
			Label:  kw,
			Kind:   CompletionItemKindKeyword,
			Detail: "Keyword",
		})
	}

	// Add built-in types
	types := []string{"int", "str", "string", "bool", "float"}
	for _, t := range types {
		items = append(items, CompletionItem{
			Label:  t,
			Kind:   CompletionItemKindClass,
			Detail: "Built-in type",
		})
	}

	// Add HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, m := range methods {
		items = append(items, CompletionItem{
			Label:  m,
			Kind:   CompletionItemKindKeyword,
			Detail: "HTTP method",
		})
	}

	// Add defined types
	if doc.AST != nil {
		for _, item := range doc.AST.Items {
			if typeDef, ok := item.(*interpreter.TypeDef); ok {
				items = append(items, CompletionItem{
					Label:  typeDef.Name,
					Kind:   CompletionItemKindStruct,
					Detail: "Type definition",
				})
			}
		}
	}

	// Add route snippets
	items = append(items, CompletionItem{
		Label:            "route-get",
		Kind:             CompletionItemKindSnippet,
		Detail:           "GET route",
		InsertText:       "@ route /${1:path} [GET]\n  > {${2:response}}",
		InsertTextFormat: 2, // Snippet
	})

	items = append(items, CompletionItem{
		Label:            "route-post",
		Kind:             CompletionItemKindSnippet,
		Detail:           "POST route",
		InsertText:       "@ route /${1:path} [POST]\n  > {${2:response}}",
		InsertTextFormat: 2,
	})

	// Add CLI command snippets
	items = append(items, CompletionItem{
		Label:            "command",
		Kind:             CompletionItemKindSnippet,
		Detail:           "CLI command",
		InsertText:       "! command ${1:name} ${2:param}: ${3:str}!\n  > ${4:\"result\"}",
		InsertTextFormat: 2,
	})

	// Add cron task snippets
	items = append(items, CompletionItem{
		Label:            "cron",
		Kind:             CompletionItemKindSnippet,
		Detail:           "Scheduled task",
		InsertText:       "* cron \"${1:0 0 * * *}\"\n  > ${2:\"task executed\"}",
		InsertTextFormat: 2,
	})

	// Add event handler snippets
	items = append(items, CompletionItem{
		Label:            "event",
		Kind:             CompletionItemKindSnippet,
		Detail:           "Event handler",
		InsertText:       "~ event \"${1:event.type}\"\n  > ${2:\"event handled\"}",
		InsertTextFormat: 2,
	})

	// Add queue worker snippets
	items = append(items, CompletionItem{
		Label:            "queue",
		Kind:             CompletionItemKindSnippet,
		Detail:           "Queue worker",
		InsertText:       "& queue \"${1:queue.name}\"\n  > ${2:\"message processed\"}",
		InsertTextFormat: 2,
	})

	items = append(items, CompletionItem{
		Label:            "typedef",
		Kind:             CompletionItemKindSnippet,
		Detail:           "Type definition",
		InsertText:       ": ${1:TypeName} {\n  ${2:field}: ${3:str}!\n}",
		InsertTextFormat: 2,
	})

	return items
}

// GetDefinition returns the definition location for a symbol at position
func GetDefinition(doc *Document, pos Position) []Location {
	if doc.AST == nil {
		return nil
	}

	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Find type definitions
	for _, item := range doc.AST.Items {
		if typeDef, ok := item.(*interpreter.TypeDef); ok {
			if typeDef.Name == word {
				// Return approximate location (we'd need position info in AST for exact location)
				return []Location{
					{
						URI: doc.URI,
						Range: Range{
							Start: Position{Line: 0, Character: 0},
							End:   Position{Line: 0, Character: len(word)},
						},
					},
				}
			}
		}
	}

	return nil
}

// GetDocumentSymbols returns document symbols for outline view
func GetDocumentSymbols(doc *Document) []DocumentSymbol {
	if doc.AST == nil {
		return nil
	}

	var symbols []DocumentSymbol

	for _, item := range doc.AST.Items {
		switch v := item.(type) {
		case *interpreter.TypeDef:
			// Create symbol for type definition
			symbol := DocumentSymbol{
				Name:   v.Name,
				Kind:   SymbolKindStruct,
				Detail: "Type definition",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}

			// Add fields as children
			for _, field := range v.Fields {
				fieldSymbol := DocumentSymbol{
					Name:   field.Name,
					Kind:   SymbolKindField,
					Detail: formatType(field.TypeAnnotation),
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
					SelectionRange: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
				}
				symbol.Children = append(symbol.Children, fieldSymbol)
			}

			symbols = append(symbols, symbol)

		case *interpreter.Route:
			// Create symbol for route
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("%s %s", v.Method, v.Path),
				Kind:   SymbolKindMethod,
				Detail: "Route handler",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case *interpreter.Command:
			// Create symbol for CLI command (pointer type)
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("! %s", v.Name),
				Kind:   SymbolKindMethod,
				Detail: "CLI command",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case interpreter.Command:
			// Create symbol for CLI command (value type for compatibility)
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("! %s", v.Name),
				Kind:   SymbolKindMethod,
				Detail: "CLI command",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case *interpreter.CronTask:
			// Create symbol for cron task (pointer type)
			name := v.Name
			if name == "" {
				name = v.Schedule
			}
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("* %s", name),
				Kind:   SymbolKindMethod,
				Detail: fmt.Sprintf("Cron task (%s)", v.Schedule),
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case interpreter.CronTask:
			// Create symbol for cron task (value type for compatibility)
			name := v.Name
			if name == "" {
				name = v.Schedule
			}
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("* %s", name),
				Kind:   SymbolKindMethod,
				Detail: fmt.Sprintf("Cron task (%s)", v.Schedule),
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case *interpreter.EventHandler:
			// Create symbol for event handler (pointer type)
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("~ %s", v.EventType),
				Kind:   SymbolKindMethod,
				Detail: "Event handler",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case interpreter.EventHandler:
			// Create symbol for event handler (value type for compatibility)
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("~ %s", v.EventType),
				Kind:   SymbolKindMethod,
				Detail: "Event handler",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case *interpreter.QueueWorker:
			// Create symbol for queue worker (pointer type)
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("& %s", v.QueueName),
				Kind:   SymbolKindMethod,
				Detail: "Queue worker",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case interpreter.QueueWorker:
			// Create symbol for queue worker
			symbol := DocumentSymbol{
				Name:   fmt.Sprintf("& %s", v.QueueName),
				Kind:   SymbolKindMethod,
				Detail: "Queue worker",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)

		case interpreter.Function:
			// Create symbol for function
			symbol := DocumentSymbol{
				Name:   v.Name,
				Kind:   SymbolKindFunction,
				Detail: "Function",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			}
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

// GetReferences returns all references to a symbol at position
func GetReferences(doc *Document, pos Position, includeDeclaration bool) []Location {
	if doc.AST == nil {
		return nil
	}

	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	var locations []Location

	// Find all references to this symbol in the AST
	for _, item := range doc.AST.Items {
		switch v := item.(type) {
		case *interpreter.Route:
			// Check if the word is a route parameter
			params := extractRouteParams(v.Path)
			for _, param := range params {
				if param == word && includeDeclaration {
					// This is the declaration in the route path
					locations = append(locations, Location{
						URI: doc.URI,
						Range: Range{
							Start: Position{Line: 0, Character: 0},
							End:   Position{Line: 0, Character: 0},
						},
					})
				}
			}

			// Check route body for variable references
			locations = append(locations, findReferencesInStatements(v.Body, word, doc.URI)...)

		case interpreter.Function:
			// Check function parameters
			for _, param := range v.Params {
				if param.Name == word && includeDeclaration {
					locations = append(locations, Location{
						URI: doc.URI,
						Range: Range{
							Start: Position{Line: 0, Character: 0},
							End:   Position{Line: 0, Character: 0},
						},
					})
				}
			}

			// Check function body
			locations = append(locations, findReferencesInStatements(v.Body, word, doc.URI)...)

		case *interpreter.TypeDef:
			// Check if this is the type definition
			if v.Name == word && includeDeclaration {
				locations = append(locations, Location{
					URI: doc.URI,
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
				})
			}
		}
	}

	return locations
}

// findReferencesInStatements finds all references to a symbol in a list of statements
func findReferencesInStatements(stmts []interpreter.Statement, symbol string, uri string) []Location {
	var locations []Location

	for _, stmt := range stmts {
		locations = append(locations, findReferencesInStatement(stmt, symbol, uri)...)
	}

	return locations
}

// findReferencesInStatement finds all references to a symbol in a statement
func findReferencesInStatement(stmt interpreter.Statement, symbol string, uri string) []Location {
	var locations []Location

	switch s := stmt.(type) {
	case *interpreter.AssignStatement:
		// Check if the assignment target is our symbol
		if s.Target == symbol {
			locations = append(locations, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})
		}
		// Check the value expression
		locations = append(locations, findReferencesInExpression(s.Value, symbol, uri)...)

	case interpreter.AssignStatement:
		if s.Target == symbol {
			locations = append(locations, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})
		}
		locations = append(locations, findReferencesInExpression(s.Value, symbol, uri)...)

	case *interpreter.ReturnStatement:
		locations = append(locations, findReferencesInExpression(s.Value, symbol, uri)...)

	case interpreter.ReturnStatement:
		locations = append(locations, findReferencesInExpression(s.Value, symbol, uri)...)

	case *interpreter.IfStatement:
		locations = append(locations, findReferencesInExpression(s.Condition, symbol, uri)...)
		locations = append(locations, findReferencesInStatements(s.ThenBlock, symbol, uri)...)
		locations = append(locations, findReferencesInStatements(s.ElseBlock, symbol, uri)...)

	case interpreter.IfStatement:
		locations = append(locations, findReferencesInExpression(s.Condition, symbol, uri)...)
		locations = append(locations, findReferencesInStatements(s.ThenBlock, symbol, uri)...)
		locations = append(locations, findReferencesInStatements(s.ElseBlock, symbol, uri)...)

	case *interpreter.WhileStatement:
		locations = append(locations, findReferencesInExpression(s.Condition, symbol, uri)...)
		locations = append(locations, findReferencesInStatements(s.Body, symbol, uri)...)

	case interpreter.WhileStatement:
		locations = append(locations, findReferencesInExpression(s.Condition, symbol, uri)...)
		locations = append(locations, findReferencesInStatements(s.Body, symbol, uri)...)
	}

	return locations
}

// findReferencesInExpression finds all references to a symbol in an expression
func findReferencesInExpression(expr interpreter.Expr, symbol string, uri string) []Location {
	var locations []Location

	switch e := expr.(type) {
	case *interpreter.VariableExpr:
		if e.Name == symbol {
			locations = append(locations, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})
		}

	case interpreter.VariableExpr:
		if e.Name == symbol {
			locations = append(locations, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			})
		}

	case *interpreter.BinaryOpExpr:
		locations = append(locations, findReferencesInExpression(e.Left, symbol, uri)...)
		locations = append(locations, findReferencesInExpression(e.Right, symbol, uri)...)

	case interpreter.BinaryOpExpr:
		locations = append(locations, findReferencesInExpression(e.Left, symbol, uri)...)
		locations = append(locations, findReferencesInExpression(e.Right, symbol, uri)...)

	case *interpreter.ObjectExpr:
		for _, field := range e.Fields {
			locations = append(locations, findReferencesInExpression(field.Value, symbol, uri)...)
		}

	case interpreter.ObjectExpr:
		for _, field := range e.Fields {
			locations = append(locations, findReferencesInExpression(field.Value, symbol, uri)...)
		}

	case *interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			locations = append(locations, findReferencesInExpression(elem, symbol, uri)...)
		}

	case interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			locations = append(locations, findReferencesInExpression(elem, symbol, uri)...)
		}

	case *interpreter.FieldAccessExpr:
		locations = append(locations, findReferencesInExpression(e.Object, symbol, uri)...)

	case interpreter.FieldAccessExpr:
		locations = append(locations, findReferencesInExpression(e.Object, symbol, uri)...)
	}

	return locations
}

// extractRouteParams extracts parameter names from a route path
func extractRouteParams(path string) []string {
	params := []string{}
	parts := []rune(path)

	for i := 0; i < len(parts); i++ {
		if parts[i] == ':' {
			paramStart := i + 1
			paramEnd := paramStart

			for paramEnd < len(parts) && parts[paramEnd] != '/' {
				paramEnd++
			}

			if paramEnd > paramStart {
				paramName := string(parts[paramStart:paramEnd])
				params = append(params, paramName)
			}

			i = paramEnd - 1
		}
	}

	return params
}

// Helper functions

// checkTypes performs basic type checking and returns diagnostics
func checkTypes(module *interpreter.Module) []Diagnostic {
	var diagnostics []Diagnostic

	// For now, just check for undefined types in fields
	knownTypes := make(map[string]bool)
	knownTypes["int"] = true
	knownTypes["str"] = true
	knownTypes["string"] = true
	knownTypes["bool"] = true
	knownTypes["float"] = true

	// Collect defined types
	for _, item := range module.Items {
		if typeDef, ok := item.(*interpreter.TypeDef); ok {
			knownTypes[typeDef.Name] = true
		}
	}

	// Check for undefined types
	for _, item := range module.Items {
		if typeDef, ok := item.(*interpreter.TypeDef); ok {
			for _, field := range typeDef.Fields {
				if namedType, ok := field.TypeAnnotation.(interpreter.NamedType); ok {
					if !knownTypes[namedType.Name] {
						diagnostics = append(diagnostics, Diagnostic{
							Range: Range{
								Start: Position{Line: 0, Character: 0},
								End:   Position{Line: 0, Character: 0},
							},
							Severity: DiagnosticSeverityWarning,
							Source:   "glyph",
							Message:  fmt.Sprintf("Undefined type: %s", namedType.Name),
						})
					}
				}
			}
		}
	}

	// Validate cron tasks (pointer or value type)
	for _, item := range module.Items {
		var schedule string
		if cron, ok := item.(*interpreter.CronTask); ok {
			schedule = cron.Schedule
		} else if cron, ok := item.(interpreter.CronTask); ok {
			schedule = cron.Schedule
		} else {
			continue
		}
		if schedule == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				Severity: DiagnosticSeverityError,
				Source:   "glyph",
				Message:  "Cron task must have a schedule expression",
			})
		}
	}

	// Validate event handlers (pointer or value type)
	for _, item := range module.Items {
		var eventType string
		if event, ok := item.(*interpreter.EventHandler); ok {
			eventType = event.EventType
		} else if event, ok := item.(interpreter.EventHandler); ok {
			eventType = event.EventType
		} else {
			continue
		}
		if eventType == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				Severity: DiagnosticSeverityError,
				Source:   "glyph",
				Message:  "Event handler must specify an event type",
			})
		}
	}

	// Validate queue workers (pointer or value type)
	for _, item := range module.Items {
		var queueName string
		if queue, ok := item.(*interpreter.QueueWorker); ok {
			queueName = queue.QueueName
		} else if queue, ok := item.(interpreter.QueueWorker); ok {
			queueName = queue.QueueName
		} else {
			continue
		}
		if queueName == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				Severity: DiagnosticSeverityError,
				Source:   "glyph",
				Message:  "Queue worker must specify a queue name",
			})
		}
	}

	// Validate commands (pointer or value type)
	for _, item := range module.Items {
		var cmdName string
		if cmd, ok := item.(*interpreter.Command); ok {
			cmdName = cmd.Name
		} else if cmd, ok := item.(interpreter.Command); ok {
			cmdName = cmd.Name
		} else {
			continue
		}
		if cmdName == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
				Severity: DiagnosticSeverityError,
				Source:   "glyph",
				Message:  "Command must have a name",
			})
		}
	}

	return diagnostics
}

// getKeywordInfo returns information about a keyword
func getKeywordInfo(keyword string) string {
	keywordDocs := map[string]string{
		"route":   "**route** - Defines an HTTP route handler\n\nExample:\n```glyph\n@ route /api/users [GET]\n  > {users: []}\n```",
		"command": "**command** - Defines a CLI command\n\nExample:\n```glyph\n! command hello name: str!\n  > \"Hello \" + name\n```",
		"cron":    "**cron** - Defines a scheduled task\n\nExample:\n```glyph\n* cron \"0 0 * * *\"\n  > \"Daily task executed\"\n```",
		"event":   "**event** - Defines an event handler\n\nExample:\n```glyph\n~ event \"user.created\"\n  > \"Handle user creation\"\n```",
		"queue":   "**queue** - Defines a message queue worker\n\nExample:\n```glyph\n& queue \"email.send\"\n  > \"Process email\"\n```",
		"if":      "**if** - Conditional statement\n\nExample:\n```glyph\nif condition {\n  > {success: true}\n}\n```",
		"else":    "**else** - Alternative branch for if statement",
		"while":   "**while** - Loop that executes while condition is true\n\nExample:\n```glyph\nwhile count < 10 {\n  $ count = count + 1\n}\n```",
		"for":     "**for** - Iterates over arrays or objects\n\nExample:\n```glyph\nfor item in items {\n  > item\n}\n```",
		"switch":  "**switch** - Multi-way branch statement\n\nExample:\n```glyph\nswitch value {\n  case 1 { > \"one\" }\n  default { > \"other\" }\n}\n```",
		"case":    "**case** - Branch in switch statement",
		"default": "**default** - Default branch in switch statement",
		"true":    "**true** - Boolean literal",
		"false":   "**false** - Boolean literal",
	}

	return keywordDocs[keyword]
}

// getBuiltInTypeInfo returns information about a built-in type
func getBuiltInTypeInfo(typeName string) string {
	typeDocs := map[string]string{
		"int":    "**int** - 64-bit signed integer",
		"str":    "**str** - UTF-8 string",
		"string": "**string** - UTF-8 string (alias for str)",
		"bool":   "**bool** - Boolean (true or false)",
		"float":  "**float** - 64-bit floating point number",
	}

	return typeDocs[typeName]
}

// formatTypeDefHover formats a type definition for hover display
func formatTypeDefHover(typeDef *interpreter.TypeDef) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Type: %s**\n\n", typeDef.Name))
	sb.WriteString("```glyph\n")
	sb.WriteString(fmt.Sprintf(": %s {\n", typeDef.Name))

	for _, field := range typeDef.Fields {
		required := ""
		if field.Required {
			required = "!"
		}
		sb.WriteString(fmt.Sprintf("  %s: %s%s\n", field.Name, formatType(field.TypeAnnotation), required))
	}

	sb.WriteString("}\n```")

	return sb.String()
}

// formatRouteHover formats a route for hover display
func formatRouteHover(route *interpreter.Route) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Route: %s %s**\n\n", route.Method, route.Path))

	if route.ReturnType != nil {
		sb.WriteString(fmt.Sprintf("Returns: `%s`\n\n", formatType(route.ReturnType)))
	}

	if route.Auth != nil {
		sb.WriteString(fmt.Sprintf("Auth: `%s`\n", route.Auth.AuthType))
	}

	if route.RateLimit != nil {
		sb.WriteString(fmt.Sprintf("Rate Limit: `%d/%s`\n", route.RateLimit.Requests, route.RateLimit.Window))
	}

	return sb.String()
}

// formatCommandHover formats a CLI command for hover display
func formatCommandHover(cmd *interpreter.Command) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**CLI Command: %s**\n\n", cmd.Name))

	if cmd.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", cmd.Description))
	}

	if len(cmd.Params) > 0 {
		sb.WriteString("Parameters:\n")
		for _, param := range cmd.Params {
			required := ""
			if param.Required {
				required = "!"
			}
			defaultVal := ""
			if param.Default != nil {
				defaultVal = " (optional)"
			}
			sb.WriteString(fmt.Sprintf("- `%s: %s%s`%s\n", param.Name, formatType(param.Type), required, defaultVal))
		}
		sb.WriteString("\n")
	}

	if cmd.ReturnType != nil {
		sb.WriteString(fmt.Sprintf("Returns: `%s`\n", formatType(cmd.ReturnType)))
	}

	return sb.String()
}

// formatCronTaskHover formats a cron task for hover display
func formatCronTaskHover(cron *interpreter.CronTask) string {
	var sb strings.Builder

	if cron.Name != "" {
		sb.WriteString(fmt.Sprintf("**Cron Task: %s**\n\n", cron.Name))
	} else {
		sb.WriteString("**Cron Task**\n\n")
	}

	sb.WriteString(fmt.Sprintf("Schedule: `%s`\n", cron.Schedule))

	if cron.Timezone != "" {
		sb.WriteString(fmt.Sprintf("Timezone: `%s`\n", cron.Timezone))
	}

	if cron.Retries > 0 {
		sb.WriteString(fmt.Sprintf("Retries: `%d`\n", cron.Retries))
	}

	return sb.String()
}

// formatEventHandlerHover formats an event handler for hover display
func formatEventHandlerHover(event *interpreter.EventHandler) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Event Handler: %s**\n\n", event.EventType))

	if event.Async {
		sb.WriteString("Mode: `async`\n")
	} else {
		sb.WriteString("Mode: `sync`\n")
	}

	return sb.String()
}

// formatQueueWorkerHover formats a queue worker for hover display
func formatQueueWorkerHover(queue *interpreter.QueueWorker) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Queue Worker: %s**\n\n", queue.QueueName))

	if queue.Concurrency > 0 {
		sb.WriteString(fmt.Sprintf("Concurrency: `%d`\n", queue.Concurrency))
	}

	if queue.MaxRetries > 0 {
		sb.WriteString(fmt.Sprintf("Max Retries: `%d`\n", queue.MaxRetries))
	}

	if queue.Timeout > 0 {
		sb.WriteString(fmt.Sprintf("Timeout: `%ds`\n", queue.Timeout))
	}

	return sb.String()
}

// formatType formats a type for display
func formatType(t interpreter.Type) string {
	switch v := t.(type) {
	case interpreter.IntType:
		return "int"
	case interpreter.StringType:
		return "str"
	case interpreter.BoolType:
		return "bool"
	case interpreter.FloatType:
		return "float"
	case interpreter.ArrayType:
		return fmt.Sprintf("[%s]", formatType(v.ElementType))
	case interpreter.OptionalType:
		return fmt.Sprintf("%s?", formatType(v.InnerType))
	case interpreter.NamedType:
		return v.Name
	default:
		return "unknown"
	}
}

// getOptimizerHints analyzes code and provides optimization suggestions
func getOptimizerHints(module *interpreter.Module) []Diagnostic {
	var diagnostics []Diagnostic

	// Analyze each route for optimization opportunities
	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			hints := analyzeRouteForOptimizations(route.Body)
			diagnostics = append(diagnostics, hints...)
		}
	}

	return diagnostics
}

// analyzeRouteForOptimizations looks for optimization opportunities in statements
func analyzeRouteForOptimizations(stmts []interpreter.Statement) []Diagnostic {
	var diagnostics []Diagnostic

	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *interpreter.AssignStatement:
			// Check for constant folding opportunities
			if hint := checkConstantFoldingOpportunity(s.Value); hint != "" {
				diagnostics = append(diagnostics, Diagnostic{
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
					Severity: DiagnosticSeverityHint,
					Source:   "glyph-optimizer",
					Message:  hint,
				})
			}

		case *interpreter.WhileStatement:
			// Check for loop invariant code
			if hint := checkLoopInvariants(s); hint != "" {
				diagnostics = append(diagnostics, Diagnostic{
					Range: Range{
						Start: Position{Line: 0, Character: 0},
						End:   Position{Line: 0, Character: 0},
					},
					Severity: DiagnosticSeverityInformation,
					Source:   "glyph-optimizer",
					Message:  hint,
				})
			}
		}
	}

	return diagnostics
}

// checkConstantFoldingOpportunity checks if an expression can be constant-folded
func checkConstantFoldingOpportunity(expr interpreter.Expr) string {
	if binOp, ok := expr.(*interpreter.BinaryOpExpr); ok {
		leftLit, leftIsLit := binOp.Left.(*interpreter.LiteralExpr)
		rightLit, rightIsLit := binOp.Right.(*interpreter.LiteralExpr)

		// Both operands are literals
		if leftIsLit && rightIsLit {
			return fmt.Sprintf("💡 Constant expression detected. The optimizer will fold this at compile-time (-O2)")
		}

		// Check for algebraic simplifications
		if leftIsLit {
			if intLit, ok := leftLit.Value.(interpreter.IntLiteral); ok {
				if intLit.Value == 0 && binOp.Op == interpreter.Add {
					return "💡 Adding zero has no effect. Use -O2 to remove redundant operations"
				}
				if intLit.Value == 1 && binOp.Op == interpreter.Mul {
					return "💡 Multiplying by one has no effect. Use -O2 to optimize"
				}
			}
		}
		if rightIsLit {
			if intLit, ok := rightLit.Value.(interpreter.IntLiteral); ok {
				if intLit.Value == 0 && binOp.Op == interpreter.Add {
					return "💡 Adding zero has no effect. Use -O2 to remove redundant operations"
				}
				if intLit.Value == 1 && binOp.Op == interpreter.Mul {
					return "💡 Multiplying by one has no effect. Use -O2 to optimize"
				}
				if intLit.Value == 2 && binOp.Op == interpreter.Mul {
					return "💡 Multiplying by 2 can be optimized to addition. Use -O3 for strength reduction"
				}
			}
		}
	}

	return ""
}

// checkLoopInvariants checks for loop invariant code motion opportunities
func checkLoopInvariants(whileStmt *interpreter.WhileStatement) string {
	// Count assignments that don't depend on loop variables
	invariantCount := 0
	totalCount := 0

	for _, stmt := range whileStmt.Body {
		if _, ok := stmt.(*interpreter.AssignStatement); ok {
			totalCount++
			// Simple heuristic: if it doesn't reference loop condition variables, it might be invariant
			// A real implementation would do proper data flow analysis
			invariantCount++ // Simplified for now
		}
	}

	if invariantCount > 0 && totalCount > 1 {
		return fmt.Sprintf("💡 This loop may contain invariant code. Use -O3 to enable Loop Invariant Code Motion")
	}

	return ""
}

// ========================================
// Rename Refactoring
// ========================================

// PrepareRename prepares a rename operation at the given position
func PrepareRename(doc *Document, pos Position) *PrepareRenameResult {
	if doc.AST == nil {
		return nil
	}

	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Check if it's a renameable symbol
	if !isRenameableSymbol(doc, word) {
		return nil
	}

	// Return the range and placeholder
	wordRange := doc.GetWordRangeAtPosition(pos)
	return &PrepareRenameResult{
		Range:       wordRange,
		Placeholder: word,
	}
}

// Rename performs a rename operation
func Rename(doc *Document, pos Position, newName string) *WorkspaceEdit {
	if doc.AST == nil {
		return nil
	}

	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Validate the new name
	if !isValidIdentifier(newName) {
		return nil
	}

	// Find all references to this symbol
	references := GetReferences(doc, pos, true)
	if len(references) == 0 {
		return nil
	}

	// Create text edits for all references
	edits := make([]TextEdit, 0, len(references))
	for _, ref := range references {
		// For each reference, create an edit to replace with new name
		edits = append(edits, TextEdit{
			Range:   ref.Range,
			NewText: newName,
		})
	}

	// Create workspace edit
	return &WorkspaceEdit{
		Changes: map[string][]TextEdit{
			doc.URI: edits,
		},
	}
}

// isRenameableSymbol checks if a symbol can be renamed
func isRenameableSymbol(doc *Document, word string) bool {
	// Keywords cannot be renamed
	keywords := map[string]bool{
		"route": true, "if": true, "else": true, "while": true,
		"for": true, "in": true, "switch": true, "case": true,
		"default": true, "true": true, "false": true, "null": true,
	}
	if keywords[word] {
		return false
	}

	// Built-in types cannot be renamed
	builtInTypes := map[string]bool{
		"int": true, "str": true, "string": true, "bool": true, "float": true,
	}
	if builtInTypes[word] {
		return false
	}

	// Check if it's a defined symbol (type, variable, route param)
	for _, item := range doc.AST.Items {
		if typeDef, ok := item.(*interpreter.TypeDef); ok {
			if typeDef.Name == word {
				return true
			}
			// Check fields
			for _, field := range typeDef.Fields {
				if field.Name == word {
					return true
				}
			}
		}
		if route, ok := item.(*interpreter.Route); ok {
			params := extractRouteParams(route.Path)
			for _, param := range params {
				if param == word {
					return true
				}
			}
		}
		if fn, ok := item.(interpreter.Function); ok {
			if fn.Name == word {
				return true
			}
			for _, param := range fn.Params {
				if param.Name == word {
					return true
				}
			}
		}
	}

	return true // Default: variables are renameable
}

// isValidIdentifier checks if a string is a valid identifier
func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Remaining characters must be alphanumeric or underscore
	for _, ch := range name[1:] {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}

	return true
}

// ========================================
// Code Actions
// ========================================

// GetCodeActions returns code actions for the given range and context
func GetCodeActions(doc *Document, rangeParam Range, context CodeActionContext) []CodeAction {
	var actions []CodeAction

	// Process diagnostics and generate quick fixes
	for _, diag := range context.Diagnostics {
		fixes := generateQuickFixes(doc, diag)
		actions = append(actions, fixes...)
	}

	// Add refactoring actions based on selection
	refactorActions := generateRefactorActions(doc, rangeParam)
	actions = append(actions, refactorActions...)

	// Add source actions
	sourceActions := generateSourceActions(doc)
	actions = append(actions, sourceActions...)

	return actions
}

// generateQuickFixes generates quick fix actions for a diagnostic
func generateQuickFixes(doc *Document, diag Diagnostic) []CodeAction {
	var fixes []CodeAction

	msg := diag.Message

	// Check for undefined type error
	if strings.Contains(msg, "Undefined type:") {
		typeName := extractTypeName(msg)
		if typeName != "" {
			// Suggest creating the type
			fixes = append(fixes, CodeAction{
				Title: fmt.Sprintf("Create type '%s'", typeName),
				Kind:  CodeActionKindQuickFix,
				Diagnostics: []Diagnostic{diag},
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						doc.URI: {
							{
								Range: Range{
									Start: Position{Line: 0, Character: 0},
									End:   Position{Line: 0, Character: 0},
								},
								NewText: fmt.Sprintf(": %s {\n  // Add fields here\n}\n\n", typeName),
							},
						},
					},
				},
			})
		}
	}

	// Check for missing route return
	if strings.Contains(msg, "missing return") || strings.Contains(msg, "no return") {
		fixes = append(fixes, CodeAction{
			Title: "Add return statement",
			Kind:  CodeActionKindQuickFix,
			Diagnostics: []Diagnostic{diag},
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					doc.URI: {
						{
							Range: Range{
								Start: diag.Range.End,
								End:   diag.Range.End,
							},
							NewText: "\n  > {status: \"ok\"}",
						},
					},
				},
			},
		})
	}

	// Check for typos in keywords
	if strings.Contains(msg, "Did you mean") {
		suggestion := extractSuggestion(msg)
		if suggestion != "" {
			fixes = append(fixes, CodeAction{
				Title: fmt.Sprintf("Change to '%s'", suggestion),
				Kind:  CodeActionKindQuickFix,
				IsPreferred: true,
				Diagnostics: []Diagnostic{diag},
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						doc.URI: {
							{
								Range:   diag.Range,
								NewText: suggestion,
							},
						},
					},
				},
			})
		}
	}

	return fixes
}

// generateRefactorActions generates refactoring actions for a selection
func generateRefactorActions(doc *Document, rangeParam Range) []CodeAction {
	var actions []CodeAction

	// Get selected text
	selectedText := doc.GetTextInRange(rangeParam)
	if selectedText == "" {
		return actions
	}

	// Extract to variable
	actions = append(actions, CodeAction{
		Title: "Extract to variable",
		Kind:  CodeActionKindRefactorExtract,
		Edit: &WorkspaceEdit{
			Changes: map[string][]TextEdit{
				doc.URI: {
					{
						Range: Range{
							Start: Position{Line: rangeParam.Start.Line, Character: 0},
							End:   Position{Line: rangeParam.Start.Line, Character: 0},
						},
						NewText: fmt.Sprintf("$ extracted = %s\n  ", selectedText),
					},
					{
						Range:   rangeParam,
						NewText: "extracted",
					},
				},
			},
		},
	})

	// Inline variable (if selection is a variable name)
	if isValidIdentifier(strings.TrimSpace(selectedText)) {
		actions = append(actions, CodeAction{
			Title: "Inline variable",
			Kind:  CodeActionKindRefactorInline,
			// This would require finding the variable definition and all usages
			// For now, just provide the action structure
		})
	}

	return actions
}

// generateSourceActions generates source-level actions
func generateSourceActions(doc *Document) []CodeAction {
	var actions []CodeAction

	// Add organize imports action (if applicable)
	// For Glyph, this could organize type definitions
	if doc.AST != nil {
		hasTypes := false
		hasRoutes := false
		for _, item := range doc.AST.Items {
			if _, ok := item.(*interpreter.TypeDef); ok {
				hasTypes = true
			}
			if _, ok := item.(*interpreter.Route); ok {
				hasRoutes = true
			}
		}

		if hasTypes && hasRoutes {
			actions = append(actions, CodeAction{
				Title: "Organize declarations (types first)",
				Kind:  CodeActionKindSourceOrganize,
				// This would reorganize the file to put types before routes
			})
		}
	}

	return actions
}

// extractTypeName extracts a type name from an error message
func extractTypeName(msg string) string {
	prefix := "Undefined type: "
	idx := strings.Index(msg, prefix)
	if idx == -1 {
		return ""
	}
	typeName := msg[idx+len(prefix):]
	// Remove any trailing punctuation
	typeName = strings.TrimRight(typeName, ".,!?;")
	return strings.TrimSpace(typeName)
}

// extractSuggestion extracts a suggestion from a "Did you mean" message
func extractSuggestion(msg string) string {
	// Look for pattern like "Did you mean 'xxx'?"
	prefix := "Did you mean '"
	idx := strings.Index(msg, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(msg[start:], "'")
	if end == -1 {
		return ""
	}
	return msg[start : start+end]
}

// ========================================
// Document Formatting
// ========================================

// FormatDocument formats the entire document
func FormatDocument(doc *Document, options FormattingOptions) []TextEdit {
	if doc.Content == "" {
		return nil
	}

	var edits []TextEdit
	lines := strings.Split(doc.Content, "\n")
	indent := "  " // Default 2 spaces
	if options.TabSize > 0 {
		if options.InsertSpaces {
			indent = strings.Repeat(" ", options.TabSize)
		} else {
			indent = "\t"
		}
	}

	indentLevel := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Decrease indent before closing braces
		if strings.HasPrefix(trimmed, "}") {
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
			}
		}

		// Calculate expected indentation
		expectedIndent := strings.Repeat(indent, indentLevel)
		expectedLine := expectedIndent + trimmed

		// Compare with actual
		if line != expectedLine && trimmed != "" {
			edits = append(edits, TextEdit{
				Range: Range{
					Start: Position{Line: i, Character: 0},
					End:   Position{Line: i, Character: len(line)},
				},
				NewText: expectedLine,
			})
		}

		// Increase indent after opening braces
		if strings.HasSuffix(trimmed, "{") {
			indentLevel++
		}
	}

	// Handle trailing newline
	if options.InsertFinalNewline {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			edits = append(edits, TextEdit{
				Range: Range{
					Start: Position{Line: len(lines) - 1, Character: len(lines[len(lines)-1])},
					End:   Position{Line: len(lines) - 1, Character: len(lines[len(lines)-1])},
				},
				NewText: "\n",
			})
		}
	}

	return edits
}

// ========================================
// Signature Help
// ========================================

// GetSignatureHelp returns signature help at the given position
func GetSignatureHelp(doc *Document, pos Position) *SignatureHelp {
	// Get the context around the cursor
	line := doc.GetLine(pos.Line)
	if line == "" {
		return nil
	}

	// Find if we're inside a function call
	fnName, paramIndex := findFunctionCallContext(line, pos.Character)
	if fnName == "" {
		return nil
	}

	// Get signature for known functions
	signature := getKnownFunctionSignature(fnName)
	if signature == nil {
		return nil
	}

	return &SignatureHelp{
		Signatures:      []SignatureInformation{*signature},
		ActiveSignature: 0,
		ActiveParameter: paramIndex,
	}
}

// findFunctionCallContext finds the function name and parameter index at a position
func findFunctionCallContext(line string, col int) (string, int) {
	// Look backwards from cursor to find opening parenthesis
	if col > len(line) {
		col = len(line)
	}

	parenDepth := 0
	commaCount := 0
	fnEnd := -1

	for i := col - 1; i >= 0; i-- {
		ch := line[i]
		switch ch {
		case ')':
			parenDepth++
		case '(':
			if parenDepth == 0 {
				fnEnd = i
				// Now find function name
				fnStart := i - 1
				for fnStart >= 0 && (isAlphaNumeric(line[fnStart]) || line[fnStart] == '_') {
					fnStart--
				}
				fnStart++
				if fnStart < fnEnd {
					return line[fnStart:fnEnd], commaCount
				}
				return "", -1
			}
			parenDepth--
		case ',':
			if parenDepth == 0 {
				commaCount++
			}
		}
	}

	return "", -1
}

// isAlphaNumeric checks if a byte is alphanumeric
func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// getKnownFunctionSignature returns the signature for known built-in functions
func getKnownFunctionSignature(fnName string) *SignatureInformation {
	signatures := map[string]SignatureInformation{
		"len": {
			Label:         "len(value: array | string) -> int",
			Documentation: "Returns the length of an array or string",
			Parameters: []ParameterInformation{
				{Label: "value", Documentation: "The array or string to measure"},
			},
		},
		"append": {
			Label:         "append(array: array, value: any) -> array",
			Documentation: "Appends a value to an array",
			Parameters: []ParameterInformation{
				{Label: "array", Documentation: "The array to append to"},
				{Label: "value", Documentation: "The value to append"},
			},
		},
		"print": {
			Label:         "print(value: any) -> void",
			Documentation: "Prints a value to the console",
			Parameters: []ParameterInformation{
				{Label: "value", Documentation: "The value to print"},
			},
		},
		"json_encode": {
			Label:         "json_encode(value: any) -> string",
			Documentation: "Encodes a value as JSON string",
			Parameters: []ParameterInformation{
				{Label: "value", Documentation: "The value to encode"},
			},
		},
		"json_decode": {
			Label:         "json_decode(json: string) -> any",
			Documentation: "Decodes a JSON string",
			Parameters: []ParameterInformation{
				{Label: "json", Documentation: "The JSON string to decode"},
			},
		},
		"parseInt": {
			Label:         "parseInt(value: string) -> int",
			Documentation: "Parses a string as an integer",
			Parameters: []ParameterInformation{
				{Label: "value", Documentation: "The string to parse"},
			},
		},
		"toString": {
			Label:         "toString(value: any) -> string",
			Documentation: "Converts a value to a string",
			Parameters: []ParameterInformation{
				{Label: "value", Documentation: "The value to convert"},
			},
		},
	}

	if sig, ok := signatures[fnName]; ok {
		return &sig
	}
	return nil
}
