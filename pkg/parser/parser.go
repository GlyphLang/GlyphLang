package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// Parser converts tokens to AST
type Parser struct {
	tokens   []Token
	position int
	source   string // Original source code for error messages
}

// NewParser creates a new Parser
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
		source:   "",
	}
}

// NewParserWithSource creates a new Parser with source code for better error messages
func NewParserWithSource(tokens []Token, source string) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
		source:   source,
	}
}

// Parse parses tokens into a Module
func (p *Parser) Parse() (*interpreter.Module, error) {
	var items []interpreter.Item

	for !p.isAtEnd() {
		// Skip newlines at top level
		if p.match(NEWLINE) {
			continue
		}

		if p.isAtEnd() {
			break
		}

		switch p.current().Type {
		case COLON:
			item, err := p.parseTypeDef()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case AT:
			item, err := p.parseRoute()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case BANG:
			// ! for CLI commands (more token-efficient than @ command)
			p.advance() // consume !
			item, err := p.parseCommand()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case STAR:
			// * for cron tasks (more token-efficient than @ cron)
			p.advance() // consume *
			item, err := p.parseCronTask()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case TILDE:
			// ~ for event handlers (more token-efficient than @ event)
			p.advance() // consume ~
			item, err := p.parseEventHandler()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case AMPERSAND:
			// & for queue workers (more token-efficient than @ queue)
			p.advance() // consume &
			item, err := p.parseQueueWorker()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case IDENT:
			// Check for "type" keyword as alternative syntax
			if p.current().Literal == "type" {
				p.advance() // consume "type"
				item, err := p.parseTypeDefWithoutColon()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else {
				return nil, p.errorWithHint(
					fmt.Sprintf("Unexpected token %s", p.current().Type),
					p.current(),
					"Top-level items must start with ':', '@', '!', '*', '~', or '&'",
				)
			}
		case EOF:
			break
		default:
			return nil, p.errorWithHint(
				fmt.Sprintf("Unexpected token %s", p.current().Type),
				p.current(),
				"Top-level items must start with ':', '@', '!', '*', '~', or '&'",
			)
		}
	}

	return &interpreter.Module{Items: items}, nil
}

// parseTypeDef parses a type definition: : TypeName { fields }
func (p *Parser) parseTypeDef() (interpreter.Item, error) {
	if err := p.expect(COLON); err != nil {
		return nil, err
	}

	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var fields []interpreter.Field

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.TypeDef{
		Name:   name,
		Fields: fields,
	}, nil
}

// parseTypeDefWithoutColon parses a type definition without leading colon: type TypeName { fields }
func (p *Parser) parseTypeDefWithoutColon() (interpreter.Item, error) {
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var fields []interpreter.Field

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.TypeDef{
		Name:   name,
		Fields: fields,
	}, nil
}

// parseField parses a field: name: type!
func (p *Parser) parseField() (interpreter.Field, error) {
	name, err := p.expectIdent()
	if err != nil {
		return interpreter.Field{}, err
	}

	if err := p.expect(COLON); err != nil {
		return interpreter.Field{}, err
	}

	typeAnnotation, required, err := p.parseType()
	if err != nil {
		return interpreter.Field{}, err
	}

	return interpreter.Field{
		Name:           name,
		TypeAnnotation: typeAnnotation,
		Required:       required,
	}, nil
}

// parseType parses a type annotation
func (p *Parser) parseType() (interpreter.Type, bool, error) {
	var baseType interpreter.Type
	required := false

	if !p.check(IDENT) {
		return nil, false, p.typeError(
			fmt.Sprintf("Expected type name, but found %s", p.current().Type),
			p.current(),
		)
	}

	typeName := p.current().Literal
	p.advance()

	switch typeName {
	case "int":
		baseType = interpreter.IntType{}
	case "str", "string":
		baseType = interpreter.StringType{}
	case "bool":
		baseType = interpreter.BoolType{}
	case "float":
		baseType = interpreter.FloatType{}
	default:
		baseType = interpreter.NamedType{Name: typeName}
	}

	// Check for generic type parameters (e.g., List[str] or Map[str, str])
	if p.check(LBRACKET) {
		p.advance()

		// If there's a type inside brackets, it's a generic type with parameters
		// For now, we'll just skip to the closing bracket and treat it as a named type
		if !p.check(RBRACKET) {
			// Skip everything until the closing bracket
			bracketDepth := 1
			for bracketDepth > 0 && !p.isAtEnd() {
				if p.check(LBRACKET) {
					bracketDepth++
				} else if p.check(RBRACKET) {
					bracketDepth--
					if bracketDepth == 0 {
						break
					}
				}
				p.advance()
			}

			if err := p.expect(RBRACKET); err != nil {
				return nil, false, err
			}
			// Keep baseType as-is (NamedType like "List" or "Map")
		} else {
			// Empty brackets like int[] - treat as array type
			p.advance() // consume ]
			baseType = interpreter.ArrayType{ElementType: baseType}
		}
	}

	// Check for union types (e.g., User | Error or Result<T, E>)
	if p.check(PIPE) {
		types := []interpreter.Type{baseType}
		for p.check(PIPE) {
			p.advance() // consume |

			// Parse the next type
			if !p.check(IDENT) {
				return nil, false, p.typeError(
					fmt.Sprintf("Expected type name after |, but found %s", p.current().Type),
					p.current(),
				)
			}

			nextTypeName := p.current().Literal
			p.advance()

			var nextType interpreter.Type
			switch nextTypeName {
			case "int":
				nextType = interpreter.IntType{}
			case "str", "string":
				nextType = interpreter.StringType{}
			case "bool":
				nextType = interpreter.BoolType{}
			case "float":
				nextType = interpreter.FloatType{}
			default:
				nextType = interpreter.NamedType{Name: nextTypeName}
			}

			// Handle generic type parameters for the next type
			if p.check(LBRACKET) {
				p.advance()
				if !p.check(RBRACKET) {
					bracketDepth := 1
					for bracketDepth > 0 && !p.isAtEnd() {
						if p.check(LBRACKET) {
							bracketDepth++
						} else if p.check(RBRACKET) {
							bracketDepth--
							if bracketDepth == 0 {
								break
							}
						}
						p.advance()
					}
					if err := p.expect(RBRACKET); err != nil {
						return nil, false, err
					}
				} else {
					p.advance()
					nextType = interpreter.ArrayType{ElementType: nextType}
				}
			}

			types = append(types, nextType)
		}

		baseType = interpreter.UnionType{Types: types}
	}

	// Check for required marker
	if p.check(BANG) {
		p.advance()
		required = true
	}

	return baseType, required, nil
}

// parseRoute parses a route definition or WebSocket route
func (p *Parser) parseRoute() (interpreter.Item, error) {
	if err := p.expect(AT); err != nil {
		return nil, err
	}

	// Expect "route", "ws", "websocket", or HTTP method keyword
	routeKw, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Dispatch to WebSocket parser if needed
	if routeKw == "ws" || routeKw == "websocket" {
		return p.parseWebSocketRoute()
	}

	// Dispatch to new directive parsers
	switch routeKw {
	case "command", "cmd":
		return p.parseCommand()
	case "cron", "schedule":
		return p.parseCronTask()
	case "event", "on":
		return p.parseEventHandler()
	case "queue", "worker":
		return p.parseQueueWorker()
	}

	// Check for HTTP method shorthand: @ GET /path
	var methodFromKeyword interpreter.HttpMethod
	var hasMethodKeyword bool
	switch strings.ToUpper(routeKw) {
	case "GET":
		methodFromKeyword = interpreter.Get
		hasMethodKeyword = true
	case "POST":
		methodFromKeyword = interpreter.Post
		hasMethodKeyword = true
	case "PUT":
		methodFromKeyword = interpreter.Put
		hasMethodKeyword = true
	case "DELETE":
		methodFromKeyword = interpreter.Delete
		hasMethodKeyword = true
	case "PATCH":
		methodFromKeyword = interpreter.Patch
		hasMethodKeyword = true
	case "ROUTE":
		// Standard syntax
		hasMethodKeyword = false
	default:
		return nil, p.routeError(
			fmt.Sprintf("Expected 'route', 'ws', 'websocket', or HTTP method after '@', but found '%s'", routeKw),
			p.tokens[p.position-1],
		)
	}

	// Parse path (can be /path or path)
	var path string
	if p.check(IDENT) {
		path = p.current().Literal
		p.advance()
	} else if p.check(SLASH) {
		// Build path from slash-separated identifiers and parameters
		var pathBuilder strings.Builder

		// Keep consuming path segments: /segment or /:param
		for p.check(SLASH) {
			pathBuilder.WriteByte('/')
			p.advance()

			// After slash, check if it's a parameter (:name) or regular segment (name)
			if p.check(COLON) {
				pathBuilder.WriteByte(':')
				p.advance()
			}

			// Get identifier (path segment or param name)
			if p.check(IDENT) {
				pathBuilder.WriteString(p.current().Literal)
				p.advance()
			} else {
				break
			}
		}

		path = pathBuilder.String()

		if path == "" {
			return nil, p.routeError(
				"Invalid or empty route path",
				p.current(),
			)
		}

		// Allow "/" as a valid root path
		if path == "/" {
			// This is valid, continue
		}
	} else {
		return nil, p.errorWithHint(
			fmt.Sprintf("Expected route path, but found %s", p.current().Type),
			p.current(),
			"Route paths must start with '/' (e.g., /api/users)",
		)
	}

	// Parse HTTP method
	var method interpreter.HttpMethod
	if hasMethodKeyword {
		// Method was specified as keyword: @ GET /path
		method = methodFromKeyword
	} else {
		// Default or parse from [METHOD] syntax
		method = interpreter.Get
		if p.check(LBRACKET) {
			p.advance()
			methodName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			method, err = p.parseHTTPMethod(methodName)
			if err != nil {
				return nil, err
			}
			if err := p.expect(RBRACKET); err != nil {
				return nil, err
			}
		}
	}

	// Parse optional return type -> Type
	var returnType interpreter.Type
	if p.check(ARROW) {
		p.advance()
		returnType, _, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	p.skipNewlines()

	// Parse route body
	var auth *interpreter.AuthConfig
	var rateLimit *interpreter.RateLimit
	var injections []interpreter.Injection
	var body []interpreter.Statement

	for !p.isAtEnd() {
		switch p.current().Type {
		case PLUS:
			// Middleware: + auth(jwt) or + ratelimit(100/min)
			p.advance()
			middlewareName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}

			switch middlewareName {
			case "auth":
				auth, err = p.parseAuthConfig()
				if err != nil {
					return nil, err
				}
			case "ratelimit":
				rateLimit, err = p.parseRateLimit()
				if err != nil {
					return nil, err
				}
			default:
				// Skip unknown middleware
				if p.check(LPAREN) {
					p.advance()
					for !p.check(RPAREN) && !p.isAtEnd() {
						p.advance()
					}
					p.expect(RPAREN)
				}
			}
			p.skipNewlines()

		case PERCENT:
			// Dependency injection: % db: Database
			p.advance()
			injName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			if err := p.expect(COLON); err != nil {
				return nil, err
			}
			injType, _, err := p.parseType()
			if err != nil {
				return nil, err
			}
			injections = append(injections, interpreter.Injection{
				Name: injName,
				Type: injType,
			})
			p.skipNewlines()

		case LESS:
			// Input binding: < input: Type (skip for now)
			p.advance()
			p.expectIdent()
			p.expect(COLON)
			p.parseType()
			p.skipNewlines()

		case BANG:
			// Validation: ! validate input { ... } (skip for now)
			p.advance()
			p.expectIdent()
			p.expectIdent()
			p.expect(LBRACE)
			for !p.check(RBRACE) && !p.isAtEnd() {
				p.advance()
			}
			p.expect(RBRACE)
			p.skipNewlines()

		case QUESTION:
			// Validation statement: ? validate_fn(args)
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

		case DOLLAR, GREATER:
			// Statement
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

			// Return is typically the last statement
			if _, ok := stmt.(interpreter.ReturnStatement); ok {
				break
			}

		case IDENT:
			// Check for "if" keyword - if statement in route body
			if p.current().Literal == "if" {
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				body = append(body, stmt)
				p.skipNewlines()
			} else {
				goto endBody
			}

		case WHILE:
			// While loop in route body
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

		case FOR:
			// For loop in route body
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

		case SWITCH:
			// Switch statement in route body
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

		case NEWLINE:
			p.advance()

		case LBRACE:
			// Block syntax: @ GET /path { statements }
			p.advance() // consume '{'
			p.skipNewlines()

			// Parse statements until '}'
			for !p.check(RBRACE) && !p.isAtEnd() {
				// Handle different statement types
				switch p.current().Type {
				case DOLLAR:
					stmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					body = append(body, stmt)
				case GREATER:
					stmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					body = append(body, stmt)
				case IDENT:
					// Could be 'if', 'while', 'for', 'let', 'return', or expression statement
					switch p.current().Literal {
					case "if", "while", "for", "switch", "let", "return":
						stmt, err := p.parseStatement()
						if err != nil {
							return nil, err
						}
						body = append(body, stmt)
					default:
						// Expression statement (like function calls)
						stmt, err := p.parseStatement()
						if err != nil {
							return nil, err
						}
						body = append(body, stmt)
					}
				case WHILE, FOR, SWITCH:
					stmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					body = append(body, stmt)
				case NEWLINE:
					p.advance()
					continue
				default:
					// Try to parse as statement
					stmt, err := p.parseStatement()
					if err != nil {
						// Skip unknown tokens
						p.advance()
						continue
					}
					body = append(body, stmt)
				}
				p.skipNewlines()
			}

			if err := p.expect(RBRACE); err != nil {
				return nil, err
			}
			goto endBody

		default:
			goto endBody
		}
	}
endBody:

	return &interpreter.Route{
		Path:       path,
		Method:     method,
		ReturnType: returnType,
		Auth:       auth,
		RateLimit:  rateLimit,
		Injections: injections,
		Body:       body,
	}, nil
}

// parseWebSocketRoute parses a WebSocket route definition
// Syntax: @ ws /path { on connect {...} on message {...} on disconnect {...} }
func (p *Parser) parseWebSocketRoute() (interpreter.Item, error) {
	// Parse path
	var path string
	if p.check(SLASH) {
		// Build path from slash-separated identifiers
		var pathBuilder strings.Builder

		for p.check(SLASH) {
			pathBuilder.WriteByte('/')
			p.advance()

			if p.check(COLON) {
				pathBuilder.WriteByte(':')
				p.advance()
			}

			if p.check(IDENT) {
				pathBuilder.WriteString(p.current().Literal)
				p.advance()
			} else {
				break
			}
		}

		path = pathBuilder.String()

		if path == "" {
			return nil, p.routeError(
				"Invalid or empty WebSocket path",
				p.current(),
			)
		}
	} else {
		return nil, p.errorWithHint(
			fmt.Sprintf("Expected WebSocket path, but found %s", p.current().Type),
			p.current(),
			"WebSocket paths must start with '/' (e.g., /ws/chat)",
		)
	}

	p.skipNewlines()

	// Expect opening brace
	if err := p.expect(LBRACE); err != nil {
		return nil, p.errorWithHint(
			"WebSocket route must have a body in braces",
			p.current(),
			"Example: @ ws /chat { on connect {...} on message {...} }",
		)
	}

	p.skipNewlines()

	var events []interpreter.WebSocketEvent

	// Parse event handlers: on connect {...}, on message {...}, on disconnect {...}
	for !p.check(RBRACE) && !p.isAtEnd() {
		if p.check(IDENT) && p.current().Literal == "on" {
			p.advance() // consume "on"

			// Get event name
			eventName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}

			// Map event name to type
			var eventType interpreter.WebSocketEventType
			switch eventName {
			case "connect":
				eventType = interpreter.WSEventConnect
			case "message":
				eventType = interpreter.WSEventMessage
			case "disconnect":
				eventType = interpreter.WSEventDisconnect
			case "error":
				eventType = interpreter.WSEventError
			default:
				return nil, p.errorWithHint(
					fmt.Sprintf("Unknown WebSocket event '%s'", eventName),
					p.tokens[p.position-1],
					"Valid events are: connect, message, disconnect, error",
				)
			}

			// Expect opening brace for handler body
			if err := p.expect(LBRACE); err != nil {
				return nil, p.errorWithHint(
					fmt.Sprintf("Expected '{' after 'on %s'", eventName),
					p.current(),
					fmt.Sprintf("Example: on %s { ... }", eventName),
				)
			}

			p.skipNewlines()

			// Parse handler body
			var handlerBody []interpreter.Statement
			for !p.check(RBRACE) && !p.isAtEnd() {
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				handlerBody = append(handlerBody, stmt)
				p.skipNewlines()
			}

			if err := p.expect(RBRACE); err != nil {
				return nil, err
			}

			// Add event handler
			events = append(events, interpreter.WebSocketEvent{
				EventType: eventType,
				Body:      handlerBody,
			})
		} else {
			break
		}

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.WebSocketRoute{
		Path:   path,
		Events: events,
	}, nil
}

// parseHTTPMethod parses an HTTP method string
func (p *Parser) parseHTTPMethod(method string) (interpreter.HttpMethod, error) {
	switch strings.ToUpper(method) {
	case "GET":
		return interpreter.Get, nil
	case "POST":
		return interpreter.Post, nil
	case "PUT":
		return interpreter.Put, nil
	case "DELETE":
		return interpreter.Delete, nil
	case "PATCH":
		return interpreter.Patch, nil
	default:
		return 0, p.errorWithHint(
			fmt.Sprintf("Unknown HTTP method: %s", method),
			p.tokens[p.position-1],
			"Valid HTTP methods are: GET, POST, PUT, DELETE, PATCH",
		)
	}
}

// parseAuthConfig parses auth middleware: auth(jwt)
func (p *Parser) parseAuthConfig() (*interpreter.AuthConfig, error) {
	if err := p.expect(LPAREN); err != nil {
		return nil, err
	}

	authType, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Skip any additional parameters
	for !p.check(RPAREN) && !p.isAtEnd() {
		p.advance()
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	return &interpreter.AuthConfig{
		AuthType: authType,
		Required: true,
	}, nil
}

// parseRateLimit parses rate limit middleware: ratelimit(100/min)
func (p *Parser) parseRateLimit() (*interpreter.RateLimit, error) {
	if err := p.expect(LPAREN); err != nil {
		return nil, err
	}

	var requests uint32
	var window string

	// Parse rate limit - can be "100/min" or 100 / min
	if p.current().Type == STRING {
		parts := strings.Split(p.current().Literal, "/")
		if len(parts) >= 1 {
			n, err := strconv.ParseUint(parts[0], 10, 32)
			if err != nil {
				requests = 100
			} else {
				requests = uint32(n)
			}
		}
		if len(parts) >= 2 {
			window = parts[1]
		} else {
			window = "min"
		}
		p.advance()
	} else if p.current().Type == INTEGER {
		n, err := strconv.ParseUint(p.current().Literal, 10, 32)
		if err != nil {
			return nil, err
		}
		requests = uint32(n)
		p.advance()

		if err := p.expect(SLASH); err != nil {
			return nil, err
		}

		window, err = p.expectIdent()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("expected rate limit value, got %s", p.current().Type)
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	return &interpreter.RateLimit{
		Requests: requests,
		Window:   window,
	}, nil
}

// parseStatement parses a statement
func (p *Parser) parseStatement() (interpreter.Statement, error) {
	switch p.current().Type {
	case QUESTION:
		// ? validate_fn(args)
		p.advance()

		// Expect function call
		funcName, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		if err := p.expect(LPAREN); err != nil {
			return nil, err
		}

		var args []interpreter.Expr
		for !p.check(RPAREN) && !p.isAtEnd() {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if !p.match(COMMA) {
				break
			}
		}

		if err := p.expect(RPAREN); err != nil {
			return nil, err
		}

		return interpreter.ValidationStatement{
			Call: interpreter.FunctionCallExpr{
				Name: funcName,
				Args: args,
			},
		}, nil

	case DOLLAR:
		// $ var = expr or $ obj.field = expr or $ var: Type = expr
		p.advance()
		varName, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		// Check for optional type annotation: var: Type
		if p.check(COLON) {
			p.advance() // consume colon
			// Parse and ignore type for now (not stored in AssignStatement)
			_, _, err := p.parseType()
			if err != nil {
				return nil, err
			}

			// If no assignment follows, treat as declaration only (use nil/default value)
			if !p.check(EQUALS) {
				// Variable declaration without initialization
				// Use a default value based on context (for now, use empty string)
				return interpreter.AssignStatement{
					Target: varName,
					Value:  interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: ""}},
				}, nil
			}
		}

		// Check for field access: obj.field or obj.field.subfield
		target := varName
		for p.check(DOT) {
			p.advance() // consume dot
			fieldName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			target = target + "." + fieldName
		}

		if err := p.expect(EQUALS); err != nil {
			return nil, err
		}

		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		return interpreter.AssignStatement{
			Target: target,
			Value:  value,
		}, nil

	case GREATER:
		// > expr
		p.advance()
		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		return interpreter.ReturnStatement{
			Value: value,
		}, nil

	case IDENT:
		// Check for "if" keyword
		if p.current().Literal == "if" {
			return p.parseIfStatement()
		}

		// Check for "let" keyword (alias for $)
		if p.current().Literal == "let" {
			p.advance() // consume "let"
			// Parse as assignment statement
			varName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}

			// Check for optional type annotation
			if p.check(COLON) {
				p.advance()
				_, _, err := p.parseType()
				if err != nil {
					return nil, err
				}

				if !p.check(EQUALS) {
					return interpreter.AssignStatement{
						Target: varName,
						Value:  interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: ""}},
					}, nil
				}
			}

			if err := p.expect(EQUALS); err != nil {
				return nil, err
			}

			value, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			return interpreter.AssignStatement{
				Target: varName,
				Value:  value,
			}, nil
		}

		// Check for "return" keyword (alias for >)
		if p.current().Literal == "return" {
			p.advance() // consume "return"
			value, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			return interpreter.ReturnStatement{Value: value}, nil
		}

		// Try to parse as expression statement (e.g., function call: ws.send(...))
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		// Only allow certain expressions as statements (function calls, method calls)
		switch expr.(type) {
		case interpreter.FunctionCallExpr, interpreter.FieldAccessExpr:
			return interpreter.ExpressionStatement{Expr: expr}, nil
		default:
			return nil, p.errorWithHint(
				fmt.Sprintf("Unexpected identifier in statement position: %s", p.current().Literal),
				p.current(),
				"Did you mean to assign a variable? Use '$ varName = value'",
			)
		}

	case WHILE:
		return p.parseWhileStatement()

	case SWITCH:
		return p.parseSwitchStatement()

	case FOR:
		return p.parseForStatement()

	default:
		return nil, p.errorWithHint(
			fmt.Sprintf("Expected statement, but found %s", p.current().Type),
			p.current(),
			"Statements must start with '$' (for variable assignment) or '>' (for return)",
		)
	}
}

// parseIfStatement parses an if statement: if condition { ... } else { ... }
func (p *Parser) parseIfStatement() (interpreter.Statement, error) {
	// Consume "if" keyword
	_, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse condition
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Parse then block
	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var thenBlock []interpreter.Statement
	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(RBRACE) {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		thenBlock = append(thenBlock, stmt)

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Parse optional else / else if block
	var elseBlock []interpreter.Statement
	if p.check(IDENT) && p.current().Literal == "else" {
		p.advance() // consume "else"
		p.skipNewlines()

		// Check for "else if"
		if p.check(IDENT) && p.current().Literal == "if" {
			// Parse as nested if statement
			ifStmt, err := p.parseIfStatement()
			if err != nil {
				return nil, err
			}
			elseBlock = []interpreter.Statement{ifStmt}
		} else {
			// Regular else block
			if err := p.expect(LBRACE); err != nil {
				return nil, err
			}

			p.skipNewlines()

			for !p.check(RBRACE) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(RBRACE) {
					break
				}

				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				elseBlock = append(elseBlock, stmt)

				p.skipNewlines()
			}

			if err := p.expect(RBRACE); err != nil {
				return nil, err
			}
		}
	}

	return interpreter.IfStatement{
		Condition: condition,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
	}, nil
}

// parseWhileStatement parses a while loop: while condition { ... }
func (p *Parser) parseWhileStatement() (interpreter.Statement, error) {
	// Consume "while" keyword
	if err := p.expect(WHILE); err != nil {
		return nil, err
	}

	// Parse condition
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Parse body block
	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var body []interpreter.Statement
	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(RBRACE) {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return interpreter.WhileStatement{
		Condition: condition,
		Body:      body,
	}, nil
}

// parseForStatement parses a for loop: for item in array { ... } or for key, value in object { ... }
func (p *Parser) parseForStatement() (interpreter.Statement, error) {
	// Consume "for" keyword
	if err := p.expect(FOR); err != nil {
		return nil, err
	}

	// Parse loop variable(s)
	var keyVar, valueVar string

	// Get first identifier
	firstIdent, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Check if there's a comma (two variables: key, value)
	if p.match(COMMA) {
		keyVar = firstIdent
		valueVar, err = p.expectIdent()
		if err != nil {
			return nil, err
		}
	} else {
		// Single variable (just value/item)
		valueVar = firstIdent
	}

	// Expect "in" keyword
	if err := p.expect(IN); err != nil {
		return nil, err
	}

	// Parse iterable expression
	iterable, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Parse body block
	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var body []interpreter.Statement
	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(RBRACE) {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return interpreter.ForStatement{
		KeyVar:   keyVar,
		ValueVar: valueVar,
		Iterable: iterable,
		Body:     body,
	}, nil
}

// parseSwitchStatement parses a switch statement: switch value { case val { ... } default { ... } }
func (p *Parser) parseSwitchStatement() (interpreter.Statement, error) {
	// Consume "switch" keyword
	if err := p.expect(SWITCH); err != nil {
		return nil, err
	}

	// Parse the value to switch on
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Expect opening brace
	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var cases []interpreter.SwitchCase
	var defaultBlock []interpreter.Statement

	// Parse cases and default
	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(RBRACE) {
			break
		}

		if p.check(CASE) {
			// Parse case
			p.advance() // consume "case"

			// Parse case value
			caseValue, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			p.skipNewlines()

			// Parse case body
			if err := p.expect(LBRACE); err != nil {
				return nil, err
			}

			p.skipNewlines()

			var caseBody []interpreter.Statement
			for !p.check(RBRACE) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(RBRACE) {
					break
				}

				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				caseBody = append(caseBody, stmt)

				p.skipNewlines()
			}

			if err := p.expect(RBRACE); err != nil {
				return nil, err
			}

			cases = append(cases, interpreter.SwitchCase{
				Value: caseValue,
				Body:  caseBody,
			})

			p.skipNewlines()

		} else if p.check(DEFAULT) {
			// Parse default case
			p.advance() // consume "default"

			p.skipNewlines()

			// Parse default body
			if err := p.expect(LBRACE); err != nil {
				return nil, err
			}

			p.skipNewlines()

			for !p.check(RBRACE) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(RBRACE) {
					break
				}

				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				defaultBlock = append(defaultBlock, stmt)

				p.skipNewlines()
			}

			if err := p.expect(RBRACE); err != nil {
				return nil, err
			}

			p.skipNewlines()

		} else {
			return nil, p.errorWithHint(
				fmt.Sprintf("Expected 'case' or 'default' in switch body, but found %s", p.current().Type),
				p.current(),
				"Switch statements must contain 'case' or 'default' blocks",
			)
		}
	}

	// Expect closing brace
	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return interpreter.SwitchStatement{
		Value:   value,
		Cases:   cases,
		Default: defaultBlock,
	}, nil
}

// parseExpr parses an expression with operator precedence
func (p *Parser) parseExpr() (interpreter.Expr, error) {
	return p.parseBinaryExpr(0)
}

// parseBinaryExpr parses binary expressions with precedence climbing
func (p *Parser) parseBinaryExpr(minPrecedence int) (interpreter.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		op, precedence := p.currentBinaryOp()
		if precedence < minPrecedence {
			break
		}

		p.advance() // consume operator

		right, err := p.parseBinaryExpr(precedence + 1)
		if err != nil {
			return nil, err
		}

		left = interpreter.BinaryOpExpr{
			Op:    op,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

// currentBinaryOp returns the current token as a binary operator with precedence
func (p *Parser) currentBinaryOp() (interpreter.BinOp, int) {
	switch p.current().Type {
	case PLUS:
		return interpreter.Add, 10
	case MINUS:
		return interpreter.Sub, 10
	case STAR:
		return interpreter.Mul, 20
	case SLASH:
		return interpreter.Div, 20
	case EQ_EQ:
		return interpreter.Eq, 5
	case NOT_EQ:
		return interpreter.Ne, 5
	case LESS:
		return interpreter.Lt, 5
	case LESS_EQ:
		return interpreter.Le, 5
	case GREATER:
		return interpreter.Gt, 5
	case GREATER_EQ:
		return interpreter.Ge, 5
	case AND:
		return interpreter.And, 3
	case OR:
		return interpreter.Or, 2
	default:
		return interpreter.BinOp(-1), -1
	}
}

// parseUnary parses unary expressions (!, -)
func (p *Parser) parseUnary() (interpreter.Expr, error) {
	// Check for unary NOT operator
	if p.check(BANG) {
		p.advance() // consume !
		right, err := p.parseUnary() // recursively parse for chained unary ops
		if err != nil {
			return nil, err
		}
		return interpreter.UnaryOpExpr{
			Op:    interpreter.Not,
			Right: right,
		}, nil
	}

	// Check for unary minus (negation)
	if p.check(MINUS) {
		// Only treat as unary minus if it's at the start of an expression
		// or after an operator (not after an identifier or literal)
		p.advance() // consume -
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return interpreter.UnaryOpExpr{
			Op:    interpreter.Neg,
			Right: right,
		}, nil
	}

	return p.parsePrimary()
}

// parsePrimary parses a primary expression
func (p *Parser) parsePrimary() (interpreter.Expr, error) {
	switch p.current().Type {
	case INTEGER:
		n, err := strconv.ParseInt(p.current().Literal, 10, 64)
		if err != nil {
			return nil, err
		}
		p.advance()
		return interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: n}}, nil

	case FLOAT:
		f, err := strconv.ParseFloat(p.current().Literal, 64)
		if err != nil {
			return nil, err
		}
		p.advance()
		return interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: f}}, nil

	case STRING:
		s := p.current().Literal
		p.advance()
		return interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: s}}, nil

	case TRUE:
		p.advance()
		return interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}, nil

	case FALSE:
		p.advance()
		return interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}}, nil

	case NULL:
		p.advance()
		return interpreter.LiteralExpr{Value: interpreter.NullLiteral{}}, nil

	case IDENT:
		name := p.current().Literal
		p.advance()

		// Check for field access: a.b.c
		if p.check(DOT) {
			return p.parseFieldAccess(name)
		}

		// Check for array indexing: a[0]
		if p.check(LBRACKET) {
			return p.parseArrayIndex(interpreter.VariableExpr{Name: name})
		}

		// Check for function call: f(...)
		if p.check(LPAREN) {
			p.advance()
			var args []interpreter.Expr

			for !p.check(RPAREN) && !p.isAtEnd() {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)

				if !p.match(COMMA) {
					break
				}
			}

			if err := p.expect(RPAREN); err != nil {
				return nil, err
			}

			return interpreter.FunctionCallExpr{
				Name: name,
				Args: args,
			}, nil
		}

		return interpreter.VariableExpr{Name: name}, nil

	case LBRACE:
		// Object literal: {key: value} or {:key = value}
		p.advance()
		p.skipNewlines()

		var fields []interpreter.ObjectField

		for !p.check(RBRACE) && !p.isAtEnd() {
			p.skipNewlines()

			if p.check(RBRACE) {
				break
			}

			var fieldName string
			var err error

			// Check for alternate syntax: :field = value
			if p.check(COLON) {
				p.advance() // consume colon
				fieldName, err = p.expectIdent()
				if err != nil {
					return nil, err
				}
				if err := p.expect(EQUALS); err != nil {
					return nil, err
				}
			} else {
				// Standard syntax: field: value
				fieldName, err = p.expectIdent()
				if err != nil {
					return nil, err
				}
				if err := p.expect(COLON); err != nil {
					return nil, err
				}
			}

			fieldValue, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			fields = append(fields, interpreter.ObjectField{
				Key:   fieldName,
				Value: fieldValue,
			})

			if !p.match(COMMA) {
				break
			}

			p.skipNewlines()
		}

		p.skipNewlines()
		if err := p.expect(RBRACE); err != nil {
			return nil, err
		}

		return interpreter.ObjectExpr{Fields: fields}, nil

	case LPAREN:
		// Grouped expression
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case LBRACKET:
		// Array literal: [1, 2, 3]
		p.advance()
		p.skipNewlines()

		var elements []interpreter.Expr

		for !p.check(RBRACKET) && !p.isAtEnd() {
			p.skipNewlines()

			if p.check(RBRACKET) {
				break
			}

			element, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			elements = append(elements, element)

			if !p.match(COMMA) {
				break
			}

			p.skipNewlines()
		}

		p.skipNewlines()
		if err := p.expect(RBRACKET); err != nil {
			return nil, err
		}

		return interpreter.ArrayExpr{Elements: elements}, nil

	default:
		return nil, p.expressionError(
			fmt.Sprintf("Unexpected token in expression: %s", p.current().Type),
			p.current(),
		)
	}
}

// parseFieldAccess parses field access: obj.field or obj.field.subfield
func (p *Parser) parseFieldAccess(base string) (interpreter.Expr, error) {
	var object interpreter.Expr = interpreter.VariableExpr{Name: base}

	for p.match(DOT) {
		field, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		// For now, treat method calls as function calls
		// TODO: Add proper method call support
		if p.check(LPAREN) {
			p.advance()
			var args []interpreter.Expr

			// Parse actual arguments (not including the object)
			for !p.check(RPAREN) && !p.isAtEnd() {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)

				if !p.match(COMMA) {
					break
				}
			}

			if err := p.expect(RPAREN); err != nil {
				return nil, err
			}

			// Check if the object is a simple variable (like "ws")
			// If so, create a qualified function name (ws.method_name)
			// This is needed for built-in namespaced functions
			if varExpr, ok := object.(interpreter.VariableExpr); ok {
				return interpreter.FunctionCallExpr{
					Name: varExpr.Name + "." + field,
					Args: args,
				}, nil
			}

			// For complex objects, add object as first arg
			allArgs := make([]interpreter.Expr, 0, len(args)+1)
			allArgs = append(allArgs, object)
			allArgs = append(allArgs, args...)

			return interpreter.FunctionCallExpr{
				Name: field,
				Args: allArgs,
			}, nil
		} else {
			object = interpreter.FieldAccessExpr{
				Object: object,
				Field:  field,
			}
		}

		// Check for array indexing after field access: obj.field[0]
		if p.check(LBRACKET) {
			object, err = p.parseArrayIndex(object)
			if err != nil {
				return nil, err
			}
		}
	}

	return object, nil
}

// parseArrayIndex parses array indexing: array[index] or array[index][index2]
func (p *Parser) parseArrayIndex(array interpreter.Expr) (interpreter.Expr, error) {
	for p.match(LBRACKET) {
		index, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		if err := p.expect(RBRACKET); err != nil {
			return nil, err
		}

		array = interpreter.ArrayIndexExpr{
			Array: array,
			Index: index,
		}
	}

	return array, nil
}

// Helper methods

func (p *Parser) current() Token {
	if p.position >= len(p.tokens) {
		return Token{Type: EOF}
	}
	return p.tokens[p.position]
}

func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.position++
	}
}

func (p *Parser) isAtEnd() bool {
	return p.position >= len(p.tokens) || p.current().Type == EOF
}

func (p *Parser) check(t TokenType) bool {
	return p.current().Type == t
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(t TokenType) error {
	if p.check(t) {
		p.advance()
		return nil
	}
	return p.expectError(t, p.current())
}

func (p *Parser) expectIdent() (string, error) {
	if p.current().Type != IDENT {
		return "", p.errorWithHint(
			fmt.Sprintf("Expected identifier, but found %s", p.current().Type),
			p.current(),
			"Identifiers must start with a letter or underscore, followed by letters, digits, or underscores",
		)
	}
	ident := p.current().Literal
	p.advance()
	return ident, nil
}

func (p *Parser) skipNewlines() {
	for p.match(NEWLINE) {
		// keep skipping
	}
}

// parseCommand parses a CLI command: @ command name [params] { body }
// Example: @ command hello name: str! --greeting: str = "Hello"
func (p *Parser) parseCommand() (interpreter.Item, error) {
	// Get command name
	cmdName, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse optional description string
	var description string
	if p.check(STRING) {
		description = p.current().Literal
		p.advance()
	}

	p.skipNewlines()

	// Parse parameters
	var params []interpreter.CommandParam

	// Parameters can be positional (name: type) or flags (--name: type)
	for !p.check(LBRACE) && !p.check(ARROW) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(LBRACE) || p.check(ARROW) {
			break
		}

		var param interpreter.CommandParam

		// Check for flag syntax: --name or -n
		if p.check(MINUS) {
			p.advance()
			if p.check(MINUS) {
				p.advance() // consume second -
			}
			param.IsFlag = true
		}

		// Get parameter name
		paramName, err := p.expectIdent()
		if err != nil {
			break // No more params
		}
		param.Name = paramName

		// Parse type annotation: name: type
		if p.check(COLON) {
			p.advance()
			paramType, required, err := p.parseType()
			if err != nil {
				return nil, err
			}
			param.Type = paramType
			param.Required = required
		}

		// Check for default value: = value
		if p.check(EQUALS) {
			p.advance()
			defaultValue, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			param.Default = defaultValue
		}

		params = append(params, param)
		p.skipNewlines()
	}

	// Parse optional return type -> Type
	var returnType interpreter.Type
	if p.check(ARROW) {
		p.advance()
		returnType, _, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	p.skipNewlines()

	// Parse body
	var body []interpreter.Statement
	if p.check(LBRACE) {
		p.advance()
		p.skipNewlines()

		for !p.check(RBRACE) && !p.isAtEnd() {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()
		}

		if err := p.expect(RBRACE); err != nil {
			return nil, err
		}
	} else {
		// Inline body (same as route)
		for !p.isAtEnd() && !p.check(AT) && !p.check(COLON) {
			if p.check(NEWLINE) {
				p.advance()
				continue
			}
			if p.check(DOLLAR) || p.check(GREATER) {
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				body = append(body, stmt)
				if _, ok := stmt.(interpreter.ReturnStatement); ok {
					break
				}
			} else {
				break
			}
			p.skipNewlines()
		}
	}

	return &interpreter.Command{
		Name:        cmdName,
		Description: description,
		Params:      params,
		ReturnType:  returnType,
		Body:        body,
	}, nil
}

// parseCronTask parses a cron scheduled task: @ cron "schedule" [name] { body }
// Example: @ cron "0 0 * * *" daily_cleanup { ... }
func (p *Parser) parseCronTask() (interpreter.Item, error) {
	// Get cron schedule (required)
	if !p.check(STRING) {
		return nil, p.errorWithHint(
			"Expected cron schedule string",
			p.current(),
			"Example: @ cron \"0 0 * * *\" { ... }",
		)
	}
	schedule := p.current().Literal
	p.advance()

	// Optional task name
	var name string
	if p.check(IDENT) {
		name = p.current().Literal
		p.advance()
	}

	// Optional timezone
	var timezone string
	if p.check(IDENT) && p.current().Literal == "tz" {
		p.advance()
		if p.check(STRING) {
			timezone = p.current().Literal
			p.advance()
		}
	}

	p.skipNewlines()

	// Parse injections and body
	var injections []interpreter.Injection
	var body []interpreter.Statement
	var retries int

	if p.check(LBRACE) {
		p.advance()
		p.skipNewlines()

		for !p.check(RBRACE) && !p.isAtEnd() {
			switch p.current().Type {
			case PERCENT:
				// Dependency injection: % db: Database
				p.advance()
				injName, err := p.expectIdent()
				if err != nil {
					return nil, err
				}
				if err := p.expect(COLON); err != nil {
					return nil, err
				}
				injType, _, err := p.parseType()
				if err != nil {
					return nil, err
				}
				injections = append(injections, interpreter.Injection{
					Name: injName,
					Type: injType,
				})

			case PLUS:
				// Middleware-like config: + retries(3)
				p.advance()
				configName, _ := p.expectIdent()
				if configName == "retries" && p.check(LPAREN) {
					p.advance()
					if p.check(INTEGER) {
						n, _ := strconv.Atoi(p.current().Literal)
						retries = n
						p.advance()
					}
					p.expect(RPAREN)
				}

			case DOLLAR, GREATER:
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				body = append(body, stmt)

			case IDENT:
				if p.current().Literal == "if" || p.current().Literal == "for" || p.current().Literal == "while" {
					stmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					body = append(body, stmt)
				} else {
					goto endCronBody
				}

			default:
				goto endCronBody
			}
			p.skipNewlines()
		}
	endCronBody:

		if err := p.expect(RBRACE); err != nil {
			return nil, err
		}
	}

	return &interpreter.CronTask{
		Name:       name,
		Schedule:   schedule,
		Timezone:   timezone,
		Retries:    retries,
		Injections: injections,
		Body:       body,
	}, nil
}

// parseEventHandler parses an event handler: @ event "event.type" { body }
// Example: @ event "user.created" { ... }
func (p *Parser) parseEventHandler() (interpreter.Item, error) {
	// Get event type (required)
	var eventType string
	if p.check(STRING) {
		eventType = p.current().Literal
		p.advance()
	} else if p.check(IDENT) {
		// Allow unquoted event types like user.created
		var builder strings.Builder
		builder.WriteString(p.current().Literal)
		p.advance()
		for p.check(DOT) {
			builder.WriteByte('.')
			p.advance()
			if p.check(IDENT) {
				builder.WriteString(p.current().Literal)
				p.advance()
			}
		}
		eventType = builder.String()
	} else {
		return nil, p.errorWithHint(
			"Expected event type",
			p.current(),
			"Example: @ event \"user.created\" { ... }",
		)
	}

	// Check for async modifier
	var async bool
	if p.check(IDENT) && p.current().Literal == "async" {
		async = true
		p.advance()
	}

	p.skipNewlines()

	// Parse injections and body
	var injections []interpreter.Injection
	var body []interpreter.Statement

	if p.check(LBRACE) {
		p.advance()
		p.skipNewlines()

		for !p.check(RBRACE) && !p.isAtEnd() {
			switch p.current().Type {
			case PERCENT:
				// Dependency injection
				p.advance()
				injName, err := p.expectIdent()
				if err != nil {
					return nil, err
				}
				if err := p.expect(COLON); err != nil {
					return nil, err
				}
				injType, _, err := p.parseType()
				if err != nil {
					return nil, err
				}
				injections = append(injections, interpreter.Injection{
					Name: injName,
					Type: injType,
				})

			case DOLLAR, GREATER:
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				body = append(body, stmt)

			case IDENT:
				if p.current().Literal == "if" || p.current().Literal == "for" || p.current().Literal == "while" {
					stmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					body = append(body, stmt)
				} else {
					goto endEventBody
				}

			default:
				goto endEventBody
			}
			p.skipNewlines()
		}
	endEventBody:

		if err := p.expect(RBRACE); err != nil {
			return nil, err
		}
	}

	return &interpreter.EventHandler{
		EventType:  eventType,
		Async:      async,
		Injections: injections,
		Body:       body,
	}, nil
}

// parseQueueWorker parses a queue worker: @ queue "queue.name" { body }
// Example: @ queue "email.send" { ... }
func (p *Parser) parseQueueWorker() (interpreter.Item, error) {
	// Get queue name (required)
	var queueName string
	if p.check(STRING) {
		queueName = p.current().Literal
		p.advance()
	} else if p.check(IDENT) {
		// Allow unquoted queue names
		var builder strings.Builder
		builder.WriteString(p.current().Literal)
		p.advance()
		for p.check(DOT) {
			builder.WriteByte('.')
			p.advance()
			if p.check(IDENT) {
				builder.WriteString(p.current().Literal)
				p.advance()
			}
		}
		queueName = builder.String()
	} else {
		return nil, p.errorWithHint(
			"Expected queue name",
			p.current(),
			"Example: @ queue \"email.send\" { ... }",
		)
	}

	p.skipNewlines()

	// Parse configuration and body
	var injections []interpreter.Injection
	var body []interpreter.Statement
	var concurrency, maxRetries, timeout int

	if p.check(LBRACE) {
		p.advance()
		p.skipNewlines()

		for !p.check(RBRACE) && !p.isAtEnd() {
			switch p.current().Type {
			case PERCENT:
				// Dependency injection
				p.advance()
				injName, err := p.expectIdent()
				if err != nil {
					return nil, err
				}
				if err := p.expect(COLON); err != nil {
					return nil, err
				}
				injType, _, err := p.parseType()
				if err != nil {
					return nil, err
				}
				injections = append(injections, interpreter.Injection{
					Name: injName,
					Type: injType,
				})

			case PLUS:
				// Configuration: + concurrency(5) + retries(3) + timeout(30)
				p.advance()
				configName, _ := p.expectIdent()
				if p.check(LPAREN) {
					p.advance()
					if p.check(INTEGER) {
						n, _ := strconv.Atoi(p.current().Literal)
						switch configName {
						case "concurrency":
							concurrency = n
						case "retries":
							maxRetries = n
						case "timeout":
							timeout = n
						}
						p.advance()
					}
					p.expect(RPAREN)
				}

			case DOLLAR, GREATER:
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				body = append(body, stmt)

			case IDENT:
				if p.current().Literal == "if" || p.current().Literal == "for" || p.current().Literal == "while" {
					stmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					body = append(body, stmt)
				} else {
					goto endQueueBody
				}

			default:
				goto endQueueBody
			}
			p.skipNewlines()
		}
	endQueueBody:

		if err := p.expect(RBRACE); err != nil {
			return nil, err
		}
	}

	return &interpreter.QueueWorker{
		QueueName:   queueName,
		Concurrency: concurrency,
		MaxRetries:  maxRetries,
		Timeout:     timeout,
		Injections:  injections,
		Body:        body,
	}, nil
}
