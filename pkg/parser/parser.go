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
		case IMPORT:
			// import "path" or import "path" as alias
			item, err := p.parseImport()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case FROM:
			// from "path" import { name1, name2 }
			item, err := p.parseFromImport()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case MODULE:
			// module "name"
			item, err := p.parseModuleDecl()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case CONST:
			// const NAME = value or const NAME: Type = value
			item, err := p.parseConstDecl()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
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
		case MACRO:
			// macro! name(params) { body }
			item, err := p.parseMacroDef()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		case IDENT:
			// Check for "test" keyword (test blocks)
			if p.current().Literal == "test" {
				item, err := p.parseTestBlock()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else if p.current().Literal == "type" {
				// Check for "type" keyword as alternative syntax
				p.advance() // consume "type"
				item, err := p.parseTypeDefWithoutColon()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else if p.current().Literal == "contract" {
				p.advance() // consume "contract"
				item, err := p.parseContract()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else if p.current().Literal == "trait" {
				p.advance() // consume "trait"
				item, err := p.parseTrait()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else if p.peek(1).Type == BANG {
				// Macro invocation at top level: name!(args)
				item, err := p.parseMacroInvocation()
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			} else {
				return nil, p.errorWithHint(
					fmt.Sprintf("Unexpected token %s", p.current().Type),
					p.current(),
					"Top-level items must start with ':', '@', '!', '*', '~', '&', 'macro', 'contract', 'trait', 'import', 'from', 'module', 'const', or 'test'",
				)
			}
		case EOF:
			break
		default:
			return nil, p.errorWithHint(
				fmt.Sprintf("Unexpected token %s", p.current().Type),
				p.current(),
				"Top-level items must start with ':', '@', '!', '*', '~', '&', 'macro', 'contract', 'trait', 'import', 'from', 'module', 'const', or 'test'",
			)
		}
	}

	return &interpreter.Module{Items: items}, nil
}

// parseTypeDef parses a type definition: : TypeName { fields } or : TypeName impl Trait { fields, methods }
func (p *Parser) parseTypeDef() (interpreter.Item, error) {
	if err := p.expect(COLON); err != nil {
		return nil, err
	}

	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse optional generic type parameters
	typeParams, err := p.parseTypeParameters()
	if err != nil {
		return nil, err
	}

	// Parse optional trait implementations: impl Trait1, Trait2
	var traits []string
	if p.check(IDENT) && p.current().Literal == "impl" {
		p.advance() // consume "impl"
		for {
			traitName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			traits = append(traits, traitName)
			if !p.match(COMMA) {
				break
			}
		}
	}

	return p.parseTypeDefBody(name, typeParams, traits)
}

// parseTypeDefWithoutColon parses a type definition without leading colon: type TypeName { fields }
func (p *Parser) parseTypeDefWithoutColon() (interpreter.Item, error) {
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse optional generic type parameters
	typeParams, err := p.parseTypeParameters()
	if err != nil {
		return nil, err
	}

	// Parse optional trait implementations: impl Trait1, Trait2
	var traits []string
	if p.check(IDENT) && p.current().Literal == "impl" {
		p.advance() // consume "impl"
		for {
			traitName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			traits = append(traits, traitName)
			if !p.match(COMMA) {
				break
			}
		}
	}

	// Delegate to parseTypeDefBody (line 266) which handles fields, methods, and brace parsing
	return p.parseTypeDefBody(name, typeParams, traits)
}

// parseTypeDefBody parses the body of a type definition (shared between parseTypeDef and parseTypeDefWithoutColon).
// It handles fields, methods, and trait implementations.
func (p *Parser) parseTypeDefBody(name string, typeParams []interpreter.TypeParameter, traits []string) (interpreter.Item, error) {
	// Get type parameter names for field parsing context
	var typeParamNames []string
	for _, tp := range typeParams {
		typeParamNames = append(typeParamNames, tp.Name)
	}

	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var fields []interpreter.Field
	var methods []interpreter.MethodDef

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		// Check if this is a method definition: name(...) -> type { body }
		// Methods have an identifier followed by LPAREN
		if p.check(IDENT) && p.peek(1).Type == LPAREN {
			method, err := p.parseMethodDef(typeParamNames)
			if err != nil {
				return nil, err
			}
			methods = append(methods, method)
		} else {
			field, err := p.parseFieldWithContext(typeParamNames)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.TypeDef{
		Name:       name,
		TypeParams: typeParams,
		Fields:     fields,
		Traits:     traits,
		Methods:    methods,
	}, nil
}

// parseMethodDef parses a method definition: name(params) -> returnType { body }
func (p *Parser) parseMethodDef(typeParamNames []string) (interpreter.MethodDef, error) {
	name, err := p.expectIdent()
	if err != nil {
		return interpreter.MethodDef{}, err
	}

	if err := p.expect(LPAREN); err != nil {
		return interpreter.MethodDef{}, err
	}

	var params []interpreter.Field
	for !p.check(RPAREN) && !p.isAtEnd() {
		if len(params) > 0 {
			if err := p.expect(COMMA); err != nil {
				return interpreter.MethodDef{}, err
			}
		}
		field, err := p.parseFieldWithContext(typeParamNames)
		if err != nil {
			return interpreter.MethodDef{}, err
		}
		params = append(params, field)
	}

	if err := p.expect(RPAREN); err != nil {
		return interpreter.MethodDef{}, err
	}

	// Parse return type: -> type
	var returnType interpreter.Type
	if p.match(ARROW) {
		returnType, _, err = p.parseTypeWithContext(typeParamNames)
		if err != nil {
			return interpreter.MethodDef{}, err
		}
	}

	// Parse body
	if err := p.expect(LBRACE); err != nil {
		return interpreter.MethodDef{}, err
	}
	p.skipNewlines()

	var body []interpreter.Statement
	for !p.check(RBRACE) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return interpreter.MethodDef{}, err
		}
		body = append(body, stmt)
		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return interpreter.MethodDef{}, err
	}

	return interpreter.MethodDef{
		Name:       name,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// parseTrait parses a trait definition: trait Name { method signatures }
func (p *Parser) parseTrait() (interpreter.Item, error) {
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse optional generic type parameters
	typeParams, err := p.parseTypeParameters()
	if err != nil {
		return nil, err
	}

	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var methods []interpreter.TraitMethodSignature

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		method, err := p.parseTraitMethodSignature()
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.TraitDef{
		Name:       name,
		TypeParams: typeParams,
		Methods:    methods,
	}, nil
}

// parseTraitMethodSignature parses a trait method signature: name(params) -> returnType
func (p *Parser) parseTraitMethodSignature() (interpreter.TraitMethodSignature, error) {
	name, err := p.expectIdent()
	if err != nil {
		return interpreter.TraitMethodSignature{}, err
	}

	if err := p.expect(LPAREN); err != nil {
		return interpreter.TraitMethodSignature{}, err
	}

	var params []interpreter.Field
	for !p.check(RPAREN) && !p.isAtEnd() {
		if len(params) > 0 {
			if err := p.expect(COMMA); err != nil {
				return interpreter.TraitMethodSignature{}, err
			}
		}
		field, err := p.parseField()
		if err != nil {
			return interpreter.TraitMethodSignature{}, err
		}
		params = append(params, field)
	}

	if err := p.expect(RPAREN); err != nil {
		return interpreter.TraitMethodSignature{}, err
	}

	// Parse return type: -> type
	var returnType interpreter.Type
	if p.match(ARROW) {
		returnType, _, err = p.parseType()
		if err != nil {
			return interpreter.TraitMethodSignature{}, err
		}
	}

	return interpreter.TraitMethodSignature{
		Name:       name,
		Params:     params,
		ReturnType: returnType,
	}, nil
}

// parseField parses a field: name: type!
func (p *Parser) parseField() (interpreter.Field, error) {
	return p.parseFieldWithContext(nil)
}

// parseFieldWithContext parses a field with type parameter context: name: type! [= default]
func (p *Parser) parseFieldWithContext(typeParamNames []string) (interpreter.Field, error) {
	name, err := p.expectIdent()
	if err != nil {
		return interpreter.Field{}, err
	}

	if err := p.expect(COLON); err != nil {
		return interpreter.Field{}, err
	}

	typeAnnotation, required, err := p.parseTypeWithContext(typeParamNames)
	if err != nil {
		return interpreter.Field{}, err
	}

	// Parse validation annotations: @minLen(2) @email @pattern("[A-Z]")
	// Annotations are optional and follow the type declaration.
	// The Annotations field on interpreter.Field is defined in ast.go.
	var annotations []interpreter.FieldAnnotation
	for p.check(AT) {
		ann, parseErr := p.parseFieldAnnotation()
		if parseErr != nil {
			return interpreter.Field{}, parseErr
		}
		annotations = append(annotations, ann)
	}

	// Check for default value: = expr
	var defaultValue interpreter.Expr
	if p.check(EQUALS) {
		equalsToken := p.current()
		p.advance()
		defaultValue, err = p.parseExpr()
		if err != nil {
			return interpreter.Field{}, err
		}

		// Validate literal default values against the declared type
		if err := p.validateDefaultType(name, typeAnnotation, defaultValue, equalsToken); err != nil {
			return interpreter.Field{}, err
		}
	}

	return interpreter.Field{
		Name:           name,
		TypeAnnotation: typeAnnotation,
		Required:       required,
		Default:        defaultValue,
		Annotations:    annotations,
	}, nil
}

// parseFieldAnnotation parses a single validation annotation like @minLen(2)
// or @email. Returns an error if the annotation has invalid syntax.
func (p *Parser) parseFieldAnnotation() (interpreter.FieldAnnotation, error) {
	p.advance() // consume @

	name, err := p.expectIdent()
	if err != nil {
		return interpreter.FieldAnnotation{}, err
	}

	ann := interpreter.FieldAnnotation{Name: name}

	// Parse optional parameters: @minLen(2), @range(0, 150), @oneOf(["a", "b"])
	if p.check(LPAREN) {
		p.advance() // consume (
		for !p.check(RPAREN) && !p.isAtEnd() {
			tok := p.current()
			switch tok.Type {
			case INTEGER:
				val, parseErr := strconv.ParseInt(tok.Literal, 10, 64)
				if parseErr != nil {
					return interpreter.FieldAnnotation{}, fmt.Errorf("line %d: invalid integer in annotation @%s: %s", tok.Line, name, tok.Literal)
				}
				ann.Params = append(ann.Params, val)
				p.advance()
			case FLOAT:
				val, parseErr := strconv.ParseFloat(tok.Literal, 64)
				if parseErr != nil {
					return interpreter.FieldAnnotation{}, fmt.Errorf("line %d: invalid float in annotation @%s: %s", tok.Line, name, tok.Literal)
				}
				ann.Params = append(ann.Params, val)
				p.advance()
			case STRING:
				ann.Params = append(ann.Params, tok.Literal)
				p.advance()
			case LBRACKET:
				// Parse string array: ["a", "b", "c"]
				p.advance() // consume [
				var items []string
				for !p.check(RBRACKET) && !p.isAtEnd() {
					if p.check(COMMA) {
						p.advance()
						continue
					}
					if p.current().Type == STRING {
						items = append(items, p.current().Literal)
						p.advance()
					} else {
						return interpreter.FieldAnnotation{}, fmt.Errorf("line %d: expected string in annotation array, got %s", p.current().Line, p.current().Type.String())
					}
				}
				if p.check(RBRACKET) {
					p.advance() // consume ]
				}
				ann.Params = append(ann.Params, items)
			case COMMA:
				p.advance() // skip comma between parameters
			default:
				return interpreter.FieldAnnotation{}, fmt.Errorf("line %d: unexpected token in annotation @%s: %s", tok.Line, name, tok.Literal)
			}
		}
		if err := p.expect(RPAREN); err != nil {
			return interpreter.FieldAnnotation{}, err
		}
	}

	return ann, nil
}

// validateDefaultType validates that a literal default value matches the declared type
// For complex expressions (function calls, binary ops, etc.), validation is deferred to runtime
func (p *Parser) validateDefaultType(fieldName string, declaredType interpreter.Type, defaultExpr interpreter.Expr, tok Token) error {
	// Only validate literal expressions at parse time
	litExpr, ok := defaultExpr.(interpreter.LiteralExpr)
	if !ok {
		// Complex expressions (function calls, binary ops, etc.) - defer to runtime
		return nil
	}

	// Get the actual type from the declared type, unwrapping OptionalType if present
	actualDeclaredType := declaredType
	if optType, ok := declaredType.(interpreter.OptionalType); ok {
		actualDeclaredType = optType.InnerType
	}

	// Handle built-in types that are parsed as NamedType
	if namedType, ok := actualDeclaredType.(interpreter.NamedType); ok {
		switch namedType.Name {
		case "any", "object":
			// 'any' and 'object' accept any literal - skip validation
			return nil
		case "timestamp":
			// 'timestamp' is semantically an int - allow int literals
			// Also allow string literals for ISO date parsing at runtime
			switch litExpr.Value.(type) {
			case interpreter.IntLiteral, interpreter.StringLiteral:
				return nil
			}
			// Fall through to error for other literal types (bool, float, null)
		}
		// For user-defined types, fall through to validation
		// This catches obvious errors like `field: User = 42`
	}

	// Check the literal type against the declared type
	switch lit := litExpr.Value.(type) {
	case interpreter.IntLiteral:
		if _, ok := actualDeclaredType.(interpreter.IntType); !ok {
			return p.errorWithHint(
				fmt.Sprintf("default value type mismatch: field '%s' expects %s, got int", fieldName, typeToString(actualDeclaredType)),
				tok,
				"Ensure the default value matches the declared type",
			)
		}
	case interpreter.StringLiteral:
		if _, ok := actualDeclaredType.(interpreter.StringType); !ok {
			return p.errorWithHint(
				fmt.Sprintf("default value type mismatch: field '%s' expects %s, got string", fieldName, typeToString(actualDeclaredType)),
				tok,
				"Ensure the default value matches the declared type",
			)
		}
	case interpreter.BoolLiteral:
		if _, ok := actualDeclaredType.(interpreter.BoolType); !ok {
			return p.errorWithHint(
				fmt.Sprintf("default value type mismatch: field '%s' expects %s, got bool", fieldName, typeToString(actualDeclaredType)),
				tok,
				"Ensure the default value matches the declared type",
			)
		}
	case interpreter.FloatLiteral:
		if _, ok := actualDeclaredType.(interpreter.FloatType); !ok {
			return p.errorWithHint(
				fmt.Sprintf("default value type mismatch: field '%s' expects %s, got float", fieldName, typeToString(actualDeclaredType)),
				tok,
				"Ensure the default value matches the declared type",
			)
		}
	case interpreter.NullLiteral:
		// null is only valid for optional types
		if _, ok := declaredType.(interpreter.OptionalType); !ok {
			return p.errorWithHint(
				fmt.Sprintf("default value type mismatch: field '%s' expects %s, got null (null is only valid for optional types)", fieldName, typeToString(declaredType)),
				tok,
				"Use '?' to mark the field as optional if you want to allow null",
			)
		}
	default:
		_ = lit // Silence unused variable warning for unknown literal types
	}

	return nil
}

// typeToString returns a human-readable string representation of a type
func typeToString(t interpreter.Type) string {
	switch t := t.(type) {
	case interpreter.IntType:
		return "int"
	case interpreter.StringType:
		return "str"
	case interpreter.BoolType:
		return "bool"
	case interpreter.FloatType:
		return "float"
	case interpreter.ArrayType:
		return "[" + typeToString(t.ElementType) + "]"
	case interpreter.OptionalType:
		return typeToString(t.InnerType) + "?"
	case interpreter.NamedType:
		return t.Name
	case interpreter.GenericType:
		// Handle generic types like List<int>
		base := typeToString(t.BaseType)
		if len(t.TypeArgs) > 0 {
			args := make([]string, len(t.TypeArgs))
			for i, arg := range t.TypeArgs {
				args[i] = typeToString(arg)
			}
			return base + "<" + strings.Join(args, ", ") + ">"
		}
		return base
	case interpreter.TypeParameterType:
		return t.Name
	case interpreter.FunctionType:
		params := make([]string, len(t.ParamTypes))
		for i, param := range t.ParamTypes {
			params[i] = typeToString(param)
		}
		return "(" + strings.Join(params, ", ") + ") -> " + typeToString(t.ReturnType)
	default:
		return "unknown"
	}
}

// validateFunctionParams validates that required parameters come before optional ones
// This ensures positional argument passing works correctly
func (p *Parser) validateFunctionParams(params []interpreter.Field) error {
	sawOptional := false
	for _, param := range params {
		hasDefault := param.Default != nil
		isRequired := param.Required && !hasDefault

		if isRequired && sawOptional {
			return fmt.Errorf("required parameter '%s' cannot come after optional parameters", param.Name)
		}
		if hasDefault || !param.Required {
			sawOptional = true
		}
	}
	return nil
}

// parseType parses a type annotation
func (p *Parser) parseType() (interpreter.Type, bool, error) {
	return p.parseTypeWithContext(nil)
}

// parseSingleType parses a single type (without union handling) with optional type parameter context
// This is used within union type parsing to avoid nested unions
func (p *Parser) parseSingleType(typeParamNames []string) (interpreter.Type, error) {
	var baseType interpreter.Type

	// Check for function type: (T) -> U or (int, string) -> bool
	if p.check(LPAREN) {
		fnType, err := p.parseFunctionType(typeParamNames)
		if err != nil {
			return nil, err
		}
		baseType = fnType
	} else if p.check(LBRACKET) {
		// Array type: [T] or [int]
		p.advance() // consume [
		elemType, _, err := p.parseTypeWithContext(typeParamNames)
		if err != nil {
			return nil, err
		}
		if err := p.expect(RBRACKET); err != nil {
			return nil, err
		}
		baseType = interpreter.ArrayType{ElementType: elemType}
	} else if !p.check(IDENT) {
		return nil, p.typeError(
			fmt.Sprintf("Expected type name, but found %s", p.current().Type),
			p.current(),
		)
	} else {
		typeName := p.current().Literal
		p.advance()

		// Check if this is a type parameter reference
		isTypeParam := false
		for _, param := range typeParamNames {
			if param == typeName {
				isTypeParam = true
				break
			}
		}

		if isTypeParam {
			baseType = interpreter.TypeParameterType{Name: typeName}
		} else {
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
		}

		// Check for generic type arguments with angle brackets: List<int>, Map<string, User>
		if p.check(LESS) {
			typeArgs, err := p.parseTypeArguments(typeParamNames)
			if err != nil {
				return nil, err
			}
			baseType = interpreter.GenericType{
				BaseType: baseType,
				TypeArgs: typeArgs,
			}
		}

		// Check for generic type parameters with square brackets (e.g., List[str] or Map[str, str])
		if p.check(LBRACKET) {
			p.advance()

			// If there's a type inside brackets, it's a generic type with parameters
			if !p.check(RBRACKET) {
				var typeArgs []interpreter.Type
				for {
					argType, _, err := p.parseTypeWithContext(typeParamNames)
					if err != nil {
						return nil, err
					}
					typeArgs = append(typeArgs, argType)

					if !p.match(COMMA) {
						break
					}
				}

				if err := p.expect(RBRACKET); err != nil {
					return nil, err
				}

				baseType = interpreter.GenericType{
					BaseType: baseType,
					TypeArgs: typeArgs,
				}
			} else {
				// Empty brackets like int[] - treat as array type
				p.advance() // consume ]
				baseType = interpreter.ArrayType{ElementType: baseType}
			}
		}
	}

	// Check for optional marker (?)
	if p.check(QUESTION) {
		p.advance()
		baseType = interpreter.OptionalType{InnerType: baseType}
	}

	// Note: No union type handling here - that's handled in parseTypeWithContext

	return baseType, nil
}

// parseTypeWithContext parses a type annotation with optional type parameter context
// The typeParamNames parameter contains names of type parameters in scope (for generic definitions)
func (p *Parser) parseTypeWithContext(typeParamNames []string) (interpreter.Type, bool, error) {
	var baseType interpreter.Type
	required := false

	// Check for function type: (T) -> U or (int, string) -> bool
	if p.check(LPAREN) {
		fnType, err := p.parseFunctionType(typeParamNames)
		if err != nil {
			return nil, false, err
		}
		baseType = fnType
	} else if p.check(LBRACKET) {
		// Array type: [T] or [int]
		p.advance() // consume [
		elemType, _, err := p.parseTypeWithContext(typeParamNames)
		if err != nil {
			return nil, false, err
		}
		if err := p.expect(RBRACKET); err != nil {
			return nil, false, err
		}
		baseType = interpreter.ArrayType{ElementType: elemType}
	} else if !p.check(IDENT) {
		return nil, false, p.typeError(
			fmt.Sprintf("Expected type name, but found %s", p.current().Type),
			p.current(),
		)
	} else {
		typeName := p.current().Literal
		p.advance()

		// Check for qualified type name (e.g., module.TypeName)
		for p.check(DOT) {
			p.advance() // consume dot
			if !p.check(IDENT) {
				return nil, false, p.typeError(
					fmt.Sprintf("Expected type name after '.', but found %s", p.current().Type),
					p.current(),
				)
			}
			typeName = typeName + "." + p.current().Literal
			p.advance()
		}

		// Check if this is a type parameter reference
		isTypeParam := false
		for _, param := range typeParamNames {
			if param == typeName {
				isTypeParam = true
				break
			}
		}

		if isTypeParam {
			baseType = interpreter.TypeParameterType{Name: typeName}
		} else {
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
		}

		// Check for generic type arguments with angle brackets: List<int>, Map<string, User>
		if p.check(LESS) {
			typeArgs, err := p.parseTypeArguments(typeParamNames)
			if err != nil {
				return nil, false, err
			}
			baseType = interpreter.GenericType{
				BaseType: baseType,
				TypeArgs: typeArgs,
			}
		}

		// Check for generic type parameters with square brackets (e.g., List[str] or Map[str, str])
		if p.check(LBRACKET) {
			p.advance()

			// If there's a type inside brackets, it's a generic type with parameters
			if !p.check(RBRACKET) {
				var typeArgs []interpreter.Type
				for {
					argType, _, err := p.parseTypeWithContext(typeParamNames)
					if err != nil {
						return nil, false, err
					}
					typeArgs = append(typeArgs, argType)

					if !p.match(COMMA) {
						break
					}
				}

				if err := p.expect(RBRACKET); err != nil {
					return nil, false, err
				}

				baseType = interpreter.GenericType{
					BaseType: baseType,
					TypeArgs: typeArgs,
				}
			} else {
				// Empty brackets like int[] - treat as array type
				p.advance() // consume ]
				baseType = interpreter.ArrayType{ElementType: baseType}
			}
		}
	}

	// Check for optional marker (?)
	if p.check(QUESTION) {
		p.advance()
		baseType = interpreter.OptionalType{InnerType: baseType}
	}

	// Check for union types (e.g., User | Error)
	if p.check(PIPE) {
		types := []interpreter.Type{baseType}
		for p.check(PIPE) {
			p.advance() // consume |
			// Parse single type without union handling to avoid nested unions
			nextType, err := p.parseSingleType(typeParamNames)
			if err != nil {
				return nil, false, err
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

// parseTypeArguments parses generic type arguments: <int, string>
func (p *Parser) parseTypeArguments(typeParamNames []string) ([]interpreter.Type, error) {
	if !p.check(LESS) {
		return nil, nil
	}
	p.advance() // consume <

	var typeArgs []interpreter.Type
	for {
		argType, _, err := p.parseTypeWithContext(typeParamNames)
		if err != nil {
			return nil, err
		}
		typeArgs = append(typeArgs, argType)

		if !p.match(COMMA) {
			break
		}
	}

	if err := p.expect(GREATER); err != nil {
		return nil, err
	}

	return typeArgs, nil
}

// parseTypeParameters parses generic type parameter declarations: <T, U: Constraint>
func (p *Parser) parseTypeParameters() ([]interpreter.TypeParameter, error) {
	if !p.check(LESS) {
		return nil, nil
	}
	p.advance() // consume <

	var params []interpreter.TypeParameter
	for {
		if !p.check(IDENT) {
			return nil, p.typeError(
				fmt.Sprintf("Expected type parameter name, but found %s", p.current().Type),
				p.current(),
			)
		}

		paramName := p.current().Literal
		p.advance()

		var constraint interpreter.Type

		// Check for constraint: T: Comparable or T extends Comparable
		if p.check(COLON) {
			p.advance()
			// Get the names of already-parsed type parameters for context
			paramNames := make([]string, len(params)+1)
			for i, tp := range params {
				paramNames[i] = tp.Name
			}
			paramNames[len(params)] = paramName

			constraintType, _, err := p.parseTypeWithContext(paramNames)
			if err != nil {
				return nil, err
			}
			constraint = constraintType
		} else if p.check(IDENT) && p.current().Literal == "extends" {
			p.advance()
			// Get the names of already-parsed type parameters for context
			paramNames := make([]string, len(params)+1)
			for i, tp := range params {
				paramNames[i] = tp.Name
			}
			paramNames[len(params)] = paramName

			constraintType, _, err := p.parseTypeWithContext(paramNames)
			if err != nil {
				return nil, err
			}
			constraint = constraintType
		}

		params = append(params, interpreter.TypeParameter{
			Name:       paramName,
			Constraint: constraint,
		})

		if !p.match(COMMA) {
			break
		}
	}

	if err := p.expect(GREATER); err != nil {
		return nil, err
	}

	return params, nil
}

// parseFunctionType parses a function type signature: (T) -> U or (int, string) -> bool
func (p *Parser) parseFunctionType(typeParamNames []string) (interpreter.Type, error) {
	if err := p.expect(LPAREN); err != nil {
		return nil, err
	}

	var paramTypes []interpreter.Type

	// Parse parameter types
	if !p.check(RPAREN) {
		for {
			paramType, _, err := p.parseTypeWithContext(typeParamNames)
			if err != nil {
				return nil, err
			}
			paramTypes = append(paramTypes, paramType)

			if !p.match(COMMA) {
				break
			}
		}
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	// Expect arrow
	if err := p.expect(ARROW); err != nil {
		return nil, err
	}

	// Parse return type
	returnType, _, err := p.parseTypeWithContext(typeParamNames)
	if err != nil {
		return nil, err
	}

	return interpreter.FunctionType{
		ParamTypes: paramTypes,
		ReturnType: returnType,
	}, nil
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
	case "rpc", "grpc":
		return p.parseGRPC()
	case "query":
		return p.parseGraphQLResolver(interpreter.GraphQLQuery)
	case "mutation":
		return p.parseGraphQLResolver(interpreter.GraphQLMutation)
	case "subscription":
		return p.parseGraphQLResolver(interpreter.GraphQLSubscription)
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
	case "SSE":
		methodFromKeyword = interpreter.SSE
		hasMethodKeyword = true
	case "ROUTE":
		// Standard syntax
		hasMethodKeyword = false
	default:
		return nil, p.routeError(
			fmt.Sprintf("Expected 'route', 'ws', 'websocket', 'sse', 'rpc', 'query', 'mutation', 'subscription', or HTTP method after '@', but found '%s'", routeKw),
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
			// Accept both IDENT and keyword tokens as valid path segments
			if p.isPathSegmentToken() {
				pathBuilder.WriteString(p.current().Literal)
				p.advance()

				// Handle hyphenated path segments: /async-simple, /user-profile
				for p.check(MINUS) {
					pathBuilder.WriteByte('-')
					p.advance()
					if p.isPathSegmentToken() {
						pathBuilder.WriteString(p.current().Literal)
						p.advance()
					} else {
						break
					}
				}
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

	// Parse route body - braces required
	var auth *interpreter.AuthConfig
	var rateLimit *interpreter.RateLimit
	var injections []interpreter.Injection
	var queryParams []interpreter.QueryParamDecl
	var body []interpreter.Statement
	var inputType interpreter.Type

	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' to start route body",
			p.current(),
			"Route bodies must be enclosed in braces: @ GET /path { ... }",
		)
	}
	p.advance() // consume '{'
	p.skipNewlines()

	// Parse route body contents until '}'
	for !p.check(RBRACE) && !p.isAtEnd() {
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
			// Input binding: < input: Type
			p.advance()
			_, err := p.expectIdent() // consume "input" identifier
			if err != nil {
				return nil, err
			}
			if err := p.expect(COLON); err != nil {
				return nil, err
			}
			inputType, _, err = p.parseType()
			if err != nil {
				return nil, err
			}
			p.skipNewlines()

		case QUESTION:
			// Check if this is query param declaration (? name: type) or validation (? validate_fn())
			if p.peek(1).Type == IDENT && p.peek(2).Type == COLON {
				p.advance() // consume ?
				param, err := p.parseQueryParamDecl()
				if err != nil {
					return nil, err
				}
				queryParams = append(queryParams, param)
			} else {
				// Validation statement: ? validate_fn(args)
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				body = append(body, stmt)
			}
			p.skipNewlines()

		case DOLLAR, GREATER:
			// Statement
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

		case IDENT:
			// Check for control flow keywords
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
			p.skipNewlines()

		case WHILE, FOR, SWITCH:
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
			p.skipNewlines()

		case NEWLINE:
			p.advance()

		default:
			// Try to parse as statement
			stmt, err := p.parseStatement()
			if err != nil {
				// Skip unknown tokens
				p.advance()
				continue
			}
			body = append(body, stmt)
			p.skipNewlines()
		}
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.Route{
		Path:        path,
		Method:      method,
		InputType:   inputType,
		ReturnType:  returnType,
		Auth:        auth,
		RateLimit:   rateLimit,
		Injections:  injections,
		QueryParams: queryParams,
		Body:        body,
	}, nil
}

// parseQueryParamDecl parses a query parameter declaration: ? name: type [= default]
// Examples:
//
//	? page: int = 1
//	? q: str!
//	? tags: str[]
func (p *Parser) parseQueryParamDecl() (interpreter.QueryParamDecl, error) {
	name, err := p.expectIdent()
	if err != nil {
		return interpreter.QueryParamDecl{}, err
	}

	if err := p.expect(COLON); err != nil {
		return interpreter.QueryParamDecl{}, err
	}

	typeAnnotation, required, err := p.parseType()
	if err != nil {
		return interpreter.QueryParamDecl{}, err
	}

	// Check if it's an array type
	isArray := false
	if _, ok := typeAnnotation.(interpreter.ArrayType); ok {
		isArray = true
	}

	// Check for default value
	var defaultValue interpreter.Expr
	if p.check(EQUALS) {
		p.advance()
		defaultValue, err = p.parseExpr()
		if err != nil {
			return interpreter.QueryParamDecl{}, err
		}
	}

	return interpreter.QueryParamDecl{
		Name:     name,
		Type:     typeAnnotation,
		Required: required,
		Default:  defaultValue,
		IsArray:  isArray,
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

		// Check for "yield" keyword (SSE event emission)
		if p.current().Literal == "yield" {
			p.advance() // consume "yield"
			value, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			return interpreter.YieldStatement{Value: value}, nil
		}

		// Check for bare assignment (reassignment): identifier = expr
		// This allows updating existing variables without $ prefix
		// Only match direct assignment (not field access which could be method calls)
		if p.peek(1).Type == EQUALS {
			return p.parseReassignment()
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

	case ASSERT:
		return p.parseAssertStatement()

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

// parseReassignment parses a simple variable reassignment: identifier = expr
// Note: Field reassignment (obj.field = expr) uses the $ syntax: $ obj.field = expr
func (p *Parser) parseReassignment() (interpreter.Statement, error) {
	varName, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	if err := p.expect(EQUALS); err != nil {
		return nil, err
	}

	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return interpreter.ReassignStatement{
		Target: varName,
		Value:  value,
	}, nil
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
	return p.parsePipeExpr()
}

// parsePipeExpr parses pipe expressions (|>) with the lowest precedence
// Pipes are left-associative: a |> b |> c parses as ((a |> b) |> c)
func (p *Parser) parsePipeExpr() (interpreter.Expr, error) {
	left, err := p.parseBinaryExpr(0)
	if err != nil {
		return nil, err
	}

	for p.current().Type == PIPE_OP {
		p.advance() // consume |>
		right, err := p.parseBinaryExpr(0)
		if err != nil {
			return nil, err
		}
		left = interpreter.PipeExpr{
			Left:  left,
			Right: right,
		}
	}

	return left, nil
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

// parseCommandDefaultExpr parses a default value expression for CLI command parameters.
// This is a specialized version that stops when it encounters what looks like the next flag
// (--name or -name pattern), avoiding the ambiguity between binary minus and flag prefix.
func (p *Parser) parseCommandDefaultExpr() (interpreter.Expr, error) {
	return p.parseCommandDefaultBinaryExpr(0)
}

// parseCommandDefaultBinaryExpr parses binary expressions for command defaults,
// but treats MINUS as a binary operator only if not followed by MINUS or IDENT
// (which would indicate a new flag parameter like --host or -h).
func (p *Parser) parseCommandDefaultBinaryExpr(minPrecedence int) (interpreter.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		op, precedence := p.currentCommandDefaultBinaryOp()
		if precedence < minPrecedence {
			break
		}

		p.advance() // consume operator

		right, err := p.parseCommandDefaultBinaryExpr(precedence + 1)
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

// currentCommandDefaultBinaryOp returns the current token as a binary operator,
// but returns -1 precedence for MINUS when followed by MINUS or IDENT
// (indicating start of a new flag parameter).
func (p *Parser) currentCommandDefaultBinaryOp() (interpreter.BinOp, int) {
	switch p.current().Type {
	case PLUS:
		return interpreter.Add, 10
	case MINUS:
		// Check if this MINUS is part of a flag prefix (--flag or -flag)
		// If the next token is MINUS or IDENT, this is likely a new flag, not binary minus
		nextTok := p.peek(1)
		if nextTok.Type == MINUS || nextTok.Type == IDENT {
			// Stop parsing - this is the start of a new flag parameter
			return interpreter.BinOp(-1), -1
		}
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
		p.advance()                  // consume !
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

	case MATCH:
		return p.parseMatchExpr()

	case ASYNC:
		return p.parseAsyncExpr()

	case AWAIT:
		return p.parseAwaitExpr()

	default:
		return nil, p.expressionError(
			fmt.Sprintf("Unexpected token in expression: %s", p.current().Type),
			p.current(),
		)
	}
}

// parseAsyncExpr parses an async block: async { statements }
func (p *Parser) parseAsyncExpr() (interpreter.Expr, error) {
	// Consume "async" keyword
	if err := p.expect(ASYNC); err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Expect opening brace
	if err := p.expect(LBRACE); err != nil {
		return nil, p.errorWithHint(
			"Expected '{' after 'async'",
			p.current(),
			"async blocks must have a body: async { ... }",
		)
	}

	p.skipNewlines()

	// Parse statements in the async block
	var body []interpreter.Statement
	for !p.check(RBRACE) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
		p.skipNewlines()
	}

	// Expect closing brace
	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return interpreter.AsyncExpr{Body: body}, nil
}

// parseAwaitExpr parses an await expression: await expr
func (p *Parser) parseAwaitExpr() (interpreter.Expr, error) {
	// Consume "await" keyword
	if err := p.expect(AWAIT); err != nil {
		return nil, err
	}

	// Parse the expression to await
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return interpreter.AwaitExpr{Expr: expr}, nil
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

// isPathSegmentToken returns true if the current token can be used as a path segment
// This includes IDENT and keyword tokens (async, await, import, etc.) which should
// be treated as regular identifiers when they appear in route paths
func (p *Parser) isPathSegmentToken() bool {
	switch p.current().Type {
	case IDENT, ASYNC, AWAIT, IMPORT, FROM, AS, MODULE, MATCH, WHEN, MACRO, QUOTE,
		TRUE, FALSE, NULL, FOR, WHILE, SWITCH, CASE, DEFAULT, IN:
		return true
	default:
		return false
	}
}

// peek looks at a token at a given offset from current position (0 = current)
func (p *Parser) peek(offset int) Token {
	pos := p.position + offset
	if pos >= len(p.tokens) {
		return Token{Type: EOF}
	}
	return p.tokens[pos]
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

// parseCommand parses a CLI command or generic function: ! name<T>(params): ReturnType { body }
// If generic type parameters or parentheses are present, it's a function. Otherwise, it's a CLI command.
// Examples:
//
//	! hello name: str! --greeting: str = "Hello"  (command)
//	! map<T, U>(arr: [T], fn: (T) -> U): [U] { body }  (generic function)
//	! double(x: int): int { ... }  (regular function)
func (p *Parser) parseCommand() (interpreter.Item, error) {
	// Get name
	cmdName, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Check for generic type parameters - if present, this is a generic function
	if p.check(LESS) {
		return p.parseGenericFunction(cmdName)
	}

	// Also check for parenthesized parameters (non-generic function syntax)
	if p.check(LPAREN) {
		return p.parseRegularFunction(cmdName)
	}

	// Parse optional description string (for commands)
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
		// Use parseCommandDefaultExpr to properly handle the case where the next
		// token is a flag prefix (--flag or -flag), not a binary minus operator
		if p.check(EQUALS) {
			p.advance()
			defaultValue, err := p.parseCommandDefaultExpr()
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

	// Parse body - braces required
	var body []interpreter.Statement
	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' to start function body",
			p.current(),
			"Function bodies must be enclosed in braces: ! name(params) { ... }",
		)
	}
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

	return &interpreter.Command{
		Name:        cmdName,
		Description: description,
		Params:      params,
		ReturnType:  returnType,
		Body:        body,
	}, nil
}

// parseGenericFunction parses a generic function: ! name<T, U>(params): ReturnType { body }
func (p *Parser) parseGenericFunction(name string) (interpreter.Item, error) {
	// Parse type parameters
	typeParams, err := p.parseTypeParameters()
	if err != nil {
		return nil, err
	}

	// Get type parameter names for parsing context
	var typeParamNames []string
	for _, tp := range typeParams {
		typeParamNames = append(typeParamNames, tp.Name)
	}

	// Parse parameter list
	if err := p.expect(LPAREN); err != nil {
		return nil, err
	}

	var params []interpreter.Field
	if !p.check(RPAREN) {
		for {
			field, err := p.parseFieldWithContext(typeParamNames)
			if err != nil {
				return nil, err
			}
			params = append(params, field)

			if !p.match(COMMA) {
				break
			}
		}
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	// Validate parameter ordering (required params must come before optional ones)
	if err := p.validateFunctionParams(params); err != nil {
		return nil, err
	}

	// Parse optional return type : Type or -> Type
	var returnType interpreter.Type
	if p.check(COLON) || p.check(ARROW) {
		p.advance()
		returnType, _, err = p.parseTypeWithContext(typeParamNames)
		if err != nil {
			return nil, err
		}
	}

	p.skipNewlines()

	// Parse body - braces required
	var body []interpreter.Statement
	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' to start function body",
			p.current(),
			"Function bodies must be enclosed in braces: ! name<T>(params) { ... }",
		)
	}
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

	return &interpreter.Function{
		Name:       name,
		TypeParams: typeParams,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// parseRegularFunction parses a non-generic function: ! name(params): ReturnType { body }
func (p *Parser) parseRegularFunction(name string) (interpreter.Item, error) {
	// Parse parameter list
	if err := p.expect(LPAREN); err != nil {
		return nil, err
	}

	var params []interpreter.Field
	if !p.check(RPAREN) {
		for {
			field, err := p.parseField()
			if err != nil {
				return nil, err
			}
			params = append(params, field)

			if !p.match(COMMA) {
				break
			}
		}
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	// Validate parameter ordering (required params must come before optional ones)
	if err := p.validateFunctionParams(params); err != nil {
		return nil, err
	}

	// Parse optional return type : Type or -> Type
	var returnType interpreter.Type
	var err error
	if p.check(COLON) || p.check(ARROW) {
		p.advance()
		returnType, _, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	p.skipNewlines()

	// Parse body - braces required
	var body []interpreter.Statement
	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' to start function body",
			p.current(),
			"Function bodies must be enclosed in braces: ! name(params) { ... }",
		)
	}
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

	return &interpreter.Function{
		Name:       name,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
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

	// Check for async modifier (can be ASYNC token or IDENT with "async")
	var async bool
	if p.check(ASYNC) || (p.check(IDENT) && p.current().Literal == "async") {
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

// parseContract parses a contract definition: contract Name { @ METHOD /path -> Type }
func (p *Parser) parseContract() (interpreter.Item, error) {
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var endpoints []interpreter.ContractEndpoint

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(RBRACE) {
			break
		}

		// Expect @ METHOD /path -> ReturnType
		if err := p.expect(AT); err != nil {
			return nil, err
		}

		methodKw, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		var method interpreter.HttpMethod
		switch strings.ToUpper(methodKw) {
		case "GET":
			method = interpreter.Get
		case "POST":
			method = interpreter.Post
		case "PUT":
			method = interpreter.Put
		case "DELETE":
			method = interpreter.Delete
		case "PATCH":
			method = interpreter.Patch
		default:
			return nil, p.routeError(
				fmt.Sprintf("Expected HTTP method in contract endpoint, found '%s'", methodKw),
				p.tokens[p.position-1],
			)
		}

		// Parse path
		var pathBuilder strings.Builder
		for p.check(SLASH) || p.check(IDENT) || p.check(COLON) {
			if p.check(SLASH) {
				pathBuilder.WriteByte('/')
				p.advance()
				if p.check(COLON) {
					pathBuilder.WriteByte(':')
					p.advance()
				}
				if p.check(IDENT) {
					pathBuilder.WriteString(p.current().Literal)
					p.advance()
				}
			} else if p.check(IDENT) {
				pathBuilder.WriteString(p.current().Literal)
				p.advance()
			} else {
				break
			}
		}
		path := pathBuilder.String()

		// Parse return type: -> Type
		var returnType interpreter.Type
		if p.match(ARROW) {
			returnType, _, err = p.parseType()
			if err != nil {
				return nil, err
			}
			// Check for union types: Type | Type
			for p.match(PIPE) {
				nextType, _, err := p.parseType()
				if err != nil {
					return nil, err
				}
				if ut, ok := returnType.(interpreter.UnionType); ok {
					ut.Types = append(ut.Types, nextType)
					returnType = ut
				} else {
					returnType = interpreter.UnionType{Types: []interpreter.Type{returnType, nextType}}
				}
			}
		}

		endpoints = append(endpoints, interpreter.ContractEndpoint{
			Method:     method,
			Path:       path,
			ReturnType: returnType,
		})

		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.ContractDef{
		Name:      name,
		Endpoints: endpoints,
	}, nil
}

// parseGRPC parses a gRPC definition. Two forms:
// Service definition: @ rpc ServiceName { MethodName(InputType) -> ReturnType ... }
// Handler implementation: @ rpc MethodName(param: Type) -> ReturnType { body }
func (p *Parser) parseGRPC() (interpreter.Item, error) {
	name, err := p.expectIdent()
	if err != nil {
		return nil, p.errorWithHint(
			"Expected service or method name after 'rpc'",
			p.current(),
			"Example: @ rpc UserService { ... } or @ rpc GetUser(req: Request) -> Response { ... }",
		)
	}

	p.skipNewlines()

	// If next is '(' it's a handler; if '{' check if it's a service definition
	if p.check(LPAREN) {
		return p.parseGRPCHandler(name)
	}

	p.skipNewlines()

	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' for service definition or '(' for handler parameters",
			p.current(),
			fmt.Sprintf("Example: @ rpc %s { ... } or @ rpc %s(req: Type) -> Type { ... }", name, name),
		)
	}
	p.advance() // consume '{'
	p.skipNewlines()

	var methods []interpreter.GRPCMethod
	for !p.check(RBRACE) && !p.isAtEnd() {
		method, parseErr := p.parseGRPCMethodDecl()
		if parseErr != nil {
			return nil, parseErr
		}
		methods = append(methods, method)
		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.GRPCService{
		Name:    name,
		Methods: methods,
	}, nil
}

// parseGRPCMethodDecl parses a method declaration inside a gRPC service definition.
// Syntax: MethodName(InputType) -> ReturnType
//
//	MethodName(InputType) -> stream ReturnType
func (p *Parser) parseGRPCMethodDecl() (interpreter.GRPCMethod, error) {
	methodName, err := p.expectIdent()
	if err != nil {
		return interpreter.GRPCMethod{}, err
	}

	if err := p.expect(LPAREN); err != nil {
		return interpreter.GRPCMethod{}, err
	}

	streamType := interpreter.GRPCUnary
	if p.check(IDENT) && p.current().Literal == "stream" {
		streamType = interpreter.GRPCClientStream
		p.advance()
		p.skipNewlines()
	}

	inputType, _, err := p.parseType()
	if err != nil {
		return interpreter.GRPCMethod{}, err
	}

	if err := p.expect(RPAREN); err != nil {
		return interpreter.GRPCMethod{}, err
	}

	if err := p.expect(ARROW); err != nil {
		return interpreter.GRPCMethod{}, err
	}

	if p.check(IDENT) && p.current().Literal == "stream" {
		if streamType == interpreter.GRPCClientStream {
			streamType = interpreter.GRPCBidirectional
		} else {
			streamType = interpreter.GRPCServerStream
		}
		p.advance()
		p.skipNewlines()
	}

	returnType, _, err := p.parseType()
	if err != nil {
		return interpreter.GRPCMethod{}, err
	}

	return interpreter.GRPCMethod{
		Name:       methodName,
		InputType:  inputType,
		ReturnType: returnType,
		StreamType: streamType,
	}, nil
}

// parseGRPCHandler parses a gRPC handler: @ rpc MethodName(param: Type) -> ReturnType { body }
func (p *Parser) parseGRPCHandler(methodName string) (interpreter.Item, error) {
	if err := p.expect(LPAREN); err != nil {
		return nil, err
	}

	var params []interpreter.Field
	for !p.check(RPAREN) && !p.isAtEnd() {
		paramName, err := p.expectIdent()
		if err != nil {
			return nil, err
		}
		if err := p.expect(COLON); err != nil {
			return nil, err
		}
		paramType, required, err := p.parseType()
		if err != nil {
			return nil, err
		}
		params = append(params, interpreter.Field{
			Name:           paramName,
			TypeAnnotation: paramType,
			Required:       required,
		})
		if !p.check(RPAREN) {
			if err := p.expect(COMMA); err != nil {
				return nil, err
			}
		}
	}
	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	streamType := interpreter.GRPCUnary
	var returnType interpreter.Type
	var err error
	if p.check(ARROW) {
		p.advance()
		if p.check(IDENT) && p.current().Literal == "stream" {
			streamType = interpreter.GRPCServerStream
			p.advance()
			p.skipNewlines()
		}
		returnType, _, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	p.skipNewlines()

	var auth *interpreter.AuthConfig
	var injections []interpreter.Injection
	var body []interpreter.Statement

	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' to start gRPC handler body",
			p.current(),
			fmt.Sprintf("Example: @ rpc %s(req: Type) -> Type { ... }", methodName),
		)
	}
	p.advance()
	p.skipNewlines()

	for !p.check(RBRACE) && !p.isAtEnd() {
		switch p.current().Type {
		case PLUS:
			p.advance()
			mwName, mwErr := p.expectIdent()
			if mwErr != nil {
				return nil, mwErr
			}
			if mwName == "auth" {
				auth, err = p.parseAuthConfig()
				if err != nil {
					return nil, err
				}
			} else {
				if p.check(LPAREN) {
					p.advance()
					for !p.check(RPAREN) && !p.isAtEnd() {
						p.advance()
					}
					if pErr := p.expect(RPAREN); pErr != nil {
						return nil, pErr
					}
				}
			}
		case PERCENT:
			p.advance()
			injName, injErr := p.expectIdent()
			if injErr != nil {
				return nil, injErr
			}
			if err := p.expect(COLON); err != nil {
				return nil, err
			}
			injType, _, typeErr := p.parseType()
			if typeErr != nil {
				return nil, typeErr
			}
			injections = append(injections, interpreter.Injection{
				Name: injName,
				Type: injType,
			})
		default:
			stmt, stmtErr := p.parseStatement()
			if stmtErr != nil {
				return nil, stmtErr
			}
			body = append(body, stmt)
		}
		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.GRPCHandler{
		MethodName: methodName,
		Params:     params,
		ReturnType: returnType,
		StreamType: streamType,
		Auth:       auth,
		Injections: injections,
		Body:       body,
	}, nil
}

// parseGraphQLResolver parses a GraphQL resolver definition.
// Syntax: @ query fieldName(param: Type) -> ReturnType { body }
//
//	@ mutation fieldName(param: Type) -> ReturnType { body }
//	@ subscription fieldName -> ReturnType { body }
func (p *Parser) parseGraphQLResolver(opType interpreter.GraphQLOperationType) (interpreter.Item, error) {
	// Parse field name (required)
	fieldName, err := p.expectIdent()
	if err != nil {
		return nil, p.errorWithHint(
			"Expected field name for GraphQL resolver",
			p.current(),
			fmt.Sprintf("Example: @ %s user(id: int) -> User { ... }", opType),
		)
	}

	// Parse optional parameters: (param: Type, param2: Type)
	var params []interpreter.Field
	if p.check(LPAREN) {
		p.advance() // consume '('
		for !p.check(RPAREN) && !p.isAtEnd() {
			paramName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			if err := p.expect(COLON); err != nil {
				return nil, err
			}
			paramType, required, err := p.parseType()
			if err != nil {
				return nil, err
			}
			params = append(params, interpreter.Field{
				Name:           paramName,
				TypeAnnotation: paramType,
				Required:       required,
			})
			if !p.check(RPAREN) {
				if err := p.expect(COMMA); err != nil {
					return nil, err
				}
			}
		}
		if err := p.expect(RPAREN); err != nil {
			return nil, err
		}
	}

	// Parse optional return type: -> Type
	var returnType interpreter.Type
	if p.check(ARROW) {
		p.advance()
		returnType, _, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	p.skipNewlines()

	// Parse body block with optional auth and injections
	var auth *interpreter.AuthConfig
	var injections []interpreter.Injection
	var body []interpreter.Statement

	if !p.check(LBRACE) {
		return nil, p.errorWithHint(
			"Expected '{' to start resolver body",
			p.current(),
			fmt.Sprintf("Example: @ %s %s { ... }", opType, fieldName),
		)
	}
	p.advance() // consume '{'
	p.skipNewlines()

	for !p.check(RBRACE) && !p.isAtEnd() {
		switch p.current().Type {
		case PLUS:
			// Middleware: + auth(jwt)
			p.advance()
			middlewareName, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			if middlewareName == "auth" {
				auth, err = p.parseAuthConfig()
				if err != nil {
					return nil, err
				}
			} else {
				if p.check(LPAREN) {
					p.advance()
					for !p.check(RPAREN) && !p.isAtEnd() {
						p.advance()
					}
					p.expect(RPAREN)
				}
			}

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

		default:
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, stmt)
		}
		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.GraphQLResolver{
		Operation:  opType,
		FieldName:  fieldName,
		Params:     params,
		ReturnType: returnType,
		Auth:       auth,
		Injections: injections,
		Body:       body,
	}, nil
}

// parseImport parses an import statement: import "path" or import "path" as alias
// Examples:
//
//	import "./utils"
//	import "./models" as m
//	import "github.com/some/package"
func (p *Parser) parseImport() (interpreter.Item, error) {
	// Consume "import" keyword
	if err := p.expect(IMPORT); err != nil {
		return nil, err
	}

	// Expect import path string
	if !p.check(STRING) {
		return nil, p.errorWithHint(
			"Expected import path string",
			p.current(),
			"Example: import \"./utils\" or import \"./models\" as m",
		)
	}
	path := p.current().Literal
	p.advance()

	// Check for optional alias: as name
	var alias string
	if p.check(AS) {
		p.advance() // consume "as"
		aliasName, err := p.expectIdent()
		if err != nil {
			return nil, p.errorWithHint(
				"Expected alias name after 'as'",
				p.current(),
				"Example: import \"./utils\" as u",
			)
		}
		alias = aliasName
	}

	return &interpreter.ImportStatement{
		Path:      path,
		Alias:     alias,
		Selective: false,
		Names:     nil,
	}, nil
}

// parseFromImport parses a selective import: from "path" import { name1, name2 as alias, ... }
// Examples:
//
//	from "./utils" import { getAllUsers }
//	from "./models" import { User, Order as Ord }
func (p *Parser) parseFromImport() (interpreter.Item, error) {
	// Consume "from" keyword
	if err := p.expect(FROM); err != nil {
		return nil, err
	}

	// Expect import path string
	if !p.check(STRING) {
		return nil, p.errorWithHint(
			"Expected import path string",
			p.current(),
			"Example: from \"./utils\" import { funcName }",
		)
	}
	path := p.current().Literal
	p.advance()

	// Expect "import" keyword
	if err := p.expect(IMPORT); err != nil {
		return nil, p.errorWithHint(
			"Expected 'import' after path",
			p.current(),
			"Example: from \"./utils\" import { funcName }",
		)
	}

	// Expect opening brace
	if err := p.expect(LBRACE); err != nil {
		return nil, p.errorWithHint(
			"Expected '{' after 'import'",
			p.current(),
			"Example: from \"./utils\" import { funcName, other as o }",
		)
	}

	p.skipNewlines()

	// Parse import names
	var names []interpreter.ImportName

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		// Get the name
		name, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		importName := interpreter.ImportName{
			Name:  name,
			Alias: "",
		}

		// Check for alias: name as alias
		if p.check(AS) {
			p.advance() // consume "as"
			aliasName, err := p.expectIdent()
			if err != nil {
				return nil, p.errorWithHint(
					"Expected alias name after 'as'",
					p.current(),
					"Example: from \"./utils\" import { funcName as f }",
				)
			}
			importName.Alias = aliasName
		}

		names = append(names, importName)

		// Check for comma
		if !p.match(COMMA) {
			break
		}

		p.skipNewlines()
	}

	p.skipNewlines()

	// Expect closing brace
	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.ImportStatement{
		Path:      path,
		Alias:     "",
		Selective: true,
		Names:     names,
	}, nil
}

// parseModuleDecl parses a module declaration: module "name"
// Example: module "myapp/utils"
func (p *Parser) parseModuleDecl() (interpreter.Item, error) {
	// Consume "module" keyword
	if err := p.expect(MODULE); err != nil {
		return nil, err
	}

	// Expect module name string
	if !p.check(STRING) {
		return nil, p.errorWithHint(
			"Expected module name string",
			p.current(),
			"Example: module \"myapp/utils\"",
		)
	}
	name := p.current().Literal
	p.advance()

	return &interpreter.ModuleDecl{
		Name: name,
	}, nil
}

// parseConstDecl parses a constant declaration: const NAME = value or const NAME: Type = value
func (p *Parser) parseConstDecl() (interpreter.Item, error) {
	// Consume "const" keyword
	if err := p.expect(CONST); err != nil {
		return nil, err
	}

	// Get constant name
	name, err := p.expectIdent()
	if err != nil {
		return nil, p.errorWithHint(
			"Expected constant name after 'const'",
			p.current(),
			"Example: const MAX_SIZE = 100",
		)
	}

	// Check for optional type annotation
	var constType interpreter.Type
	if p.match(COLON) {
		constType, _, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	// Expect '='
	if err := p.expect(EQUALS); err != nil {
		return nil, p.errorWithHint(
			"Expected '=' after constant name",
			p.current(),
			"Example: const MAX_SIZE = 100 or const PI: float = 3.14159",
		)
	}

	// Parse value expression
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &interpreter.ConstDecl{
		Name:  name,
		Value: value,
		Type:  constType,
	}, nil
}

// parseMacroDef parses a macro definition: macro! name(params) { body }
// Example: macro! log(level, msg) { if level >= logLevel { print(msg) } }
func (p *Parser) parseMacroDef() (interpreter.Item, error) {
	// Consume "macro" keyword
	if err := p.expect(MACRO); err != nil {
		return nil, err
	}

	// Expect "!" after macro
	if err := p.expect(BANG); err != nil {
		return nil, p.errorWithHint(
			"Expected '!' after 'macro'",
			p.current(),
			"Macro definitions use syntax: macro! name(params) { body }",
		)
	}

	// Get macro name
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse parameters
	if err := p.expect(LPAREN); err != nil {
		return nil, p.errorWithHint(
			"Expected '(' after macro name",
			p.current(),
			"Example: macro! log(level, msg) { ... }",
		)
	}

	var params []string
	for !p.check(RPAREN) && !p.isAtEnd() {
		paramName, err := p.expectIdent()
		if err != nil {
			return nil, err
		}
		params = append(params, paramName)

		if !p.match(COMMA) {
			break
		}
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Parse body
	if err := p.expect(LBRACE); err != nil {
		return nil, p.errorWithHint(
			"Expected '{' for macro body",
			p.current(),
			"Macro body must be enclosed in braces",
		)
	}

	p.skipNewlines()

	// Parse macro body as a list of nodes (statements and items)
	var body []interpreter.Node
	for !p.check(RBRACE) && !p.isAtEnd() {
		node, err := p.parseMacroBodyNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return &interpreter.MacroDef{
		Name:   name,
		Params: params,
		Body:   body,
	}, nil
}

// parseMacroBodyNode parses a single node in a macro body
func (p *Parser) parseMacroBodyNode() (interpreter.Node, error) {
	switch p.current().Type {
	case AT:
		// Route definition inside macro
		item, err := p.parseRoute()
		if err != nil {
			return nil, err
		}
		return item.(*interpreter.Route), nil

	case COLON:
		// Type definition inside macro
		item, err := p.parseTypeDef()
		if err != nil {
			return nil, err
		}
		return item.(*interpreter.TypeDef), nil

	case DOLLAR, GREATER, QUESTION:
		// Statement
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return statementToNode(stmt), nil

	case IDENT:
		// Could be if, for, while, let, return, or macro invocation
		switch p.current().Literal {
		case "if", "for", "while", "let", "return":
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			return statementToNode(stmt), nil
		default:
			// Check for macro invocation: name!(args)
			if p.peek(1).Type == BANG {
				inv, err := p.parseMacroInvocation()
				if err != nil {
					return nil, err
				}
				return inv.(*interpreter.MacroInvocation), nil
			}
			// Try as expression statement
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			return statementToNode(stmt), nil
		}

	case WHILE, FOR, SWITCH:
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return statementToNode(stmt), nil

	case NEWLINE:
		p.advance()
		return nil, nil

	default:
		return nil, p.errorWithHint(
			fmt.Sprintf("Unexpected token in macro body: %s", p.current().Type),
			p.current(),
			"Macro body can contain routes, type definitions, and statements",
		)
	}
}

// statementToNode converts a Statement interface to a Node interface
// This is needed because Statement and Node are separate interfaces
func statementToNode(stmt interpreter.Statement) interpreter.Node {
	switch s := stmt.(type) {
	case interpreter.AssignStatement:
		return s
	case interpreter.ReassignStatement:
		return s
	case interpreter.ReturnStatement:
		return s
	case interpreter.IfStatement:
		return s
	case interpreter.WhileStatement:
		return s
	case interpreter.ForStatement:
		return s
	case interpreter.SwitchStatement:
		return s
	case interpreter.ExpressionStatement:
		return s
	case interpreter.ValidationStatement:
		return s
	case interpreter.DbQueryStatement:
		return s
	case interpreter.WsSendStatement:
		return s
	case interpreter.WsBroadcastStatement:
		return s
	case interpreter.WsCloseStatement:
		return s
	case interpreter.WebSocketEvent:
		return s
	case interpreter.MacroInvocation:
		return s
	default:
		// Return the statement as-is, concrete types implement Node
		return stmt.(interpreter.Node)
	}
}

// parseMacroInvocation parses a macro invocation: name!(args)
// Example: log!("INFO", "Server starting")
func (p *Parser) parseMacroInvocation() (interpreter.Item, error) {
	// Get macro name
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Expect "!"
	if err := p.expect(BANG); err != nil {
		return nil, err
	}

	// Parse arguments
	if err := p.expect(LPAREN); err != nil {
		return nil, p.errorWithHint(
			"Expected '(' after macro name",
			p.current(),
			"Example: log!(\"INFO\", \"message\")",
		)
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

	return &interpreter.MacroInvocation{
		Name: name,
		Args: args,
	}, nil
}

// parseQuoteExpr parses a quote expression: quote { ... }
// Example: quote { if x > 0 { return x } }
func (p *Parser) parseQuoteExpr() (interpreter.Expr, error) {
	// Consume "quote" keyword
	if err := p.expect(QUOTE); err != nil {
		return nil, err
	}

	// Expect opening brace
	if err := p.expect(LBRACE); err != nil {
		return nil, p.errorWithHint(
			"Expected '{' after 'quote'",
			p.current(),
			"Quote expressions use syntax: quote { ... }",
		)
	}

	p.skipNewlines()

	// Parse quoted body
	var body []interpreter.Node
	for !p.check(RBRACE) && !p.isAtEnd() {
		node, err := p.parseMacroBodyNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
		p.skipNewlines()
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return interpreter.QuoteExpr{Body: body}, nil
}

// parseMatchExpr parses a match expression
// Syntax: match value { pattern => result, pattern when guard => result, _ => default }
func (p *Parser) parseMatchExpr() (interpreter.Expr, error) {
	if err := p.expect(MATCH); err != nil {
		return nil, err
	}

	// Parse the value to match on
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.skipNewlines()

	// Expect opening brace
	if err := p.expect(LBRACE); err != nil {
		return nil, p.errorWithHint(
			"Expected '{' after match value",
			p.current(),
			"Match expressions require a block with cases: match x { ... }",
		)
	}

	p.skipNewlines()

	var cases []interpreter.MatchCase
	hasWildcard := false

	// Parse cases
	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		// Parse pattern
		pattern, err := p.parsePattern()
		if err != nil {
			return nil, err
		}

		// Check if this is a wildcard pattern for exhaustiveness
		if _, ok := pattern.(interpreter.WildcardPattern); ok {
			hasWildcard = true
		}

		// Parse optional guard: when condition
		var guard interpreter.Expr
		if p.check(WHEN) {
			p.advance()
			guard, err = p.parseExpr()
			if err != nil {
				return nil, err
			}
		}

		// Expect =>
		if err := p.expect(FATARROW); err != nil {
			return nil, p.errorWithHint(
				"Expected '=>' after pattern",
				p.current(),
				"Match cases use => to separate pattern from result: pattern => result",
			)
		}

		// Parse body expression
		body, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		cases = append(cases, interpreter.MatchCase{
			Pattern: pattern,
			Guard:   guard,
			Body:    body,
		})

		// Optional comma between cases
		p.match(COMMA)
		p.skipNewlines()
	}

	// Exhaustiveness warning (log if no wildcard/default case)
	if !hasWildcard && len(cases) > 0 {
		// In a production setting, this would emit a warning
		// For now, we'll just allow non-exhaustive matches
	}

	if err := p.expect(RBRACE); err != nil {
		return nil, err
	}

	return interpreter.MatchExpr{
		Value: value,
		Cases: cases,
	}, nil
}

// parsePattern parses a pattern for match expressions
func (p *Parser) parsePattern() (interpreter.Pattern, error) {
	switch p.current().Type {
	case INTEGER:
		// Literal integer pattern
		n, err := strconv.ParseInt(p.current().Literal, 10, 64)
		if err != nil {
			return nil, err
		}
		p.advance()
		return interpreter.LiteralPattern{Value: interpreter.IntLiteral{Value: n}}, nil

	case FLOAT:
		// Literal float pattern
		f, err := strconv.ParseFloat(p.current().Literal, 64)
		if err != nil {
			return nil, err
		}
		p.advance()
		return interpreter.LiteralPattern{Value: interpreter.FloatLiteral{Value: f}}, nil

	case STRING:
		// Literal string pattern
		s := p.current().Literal
		p.advance()
		return interpreter.LiteralPattern{Value: interpreter.StringLiteral{Value: s}}, nil

	case TRUE:
		p.advance()
		return interpreter.LiteralPattern{Value: interpreter.BoolLiteral{Value: true}}, nil

	case FALSE:
		p.advance()
		return interpreter.LiteralPattern{Value: interpreter.BoolLiteral{Value: false}}, nil

	case NULL:
		p.advance()
		return interpreter.LiteralPattern{Value: interpreter.NullLiteral{}}, nil

	case IDENT:
		name := p.current().Literal
		// Check for underscore wildcard
		if name == "_" {
			p.advance()
			return interpreter.WildcardPattern{}, nil
		}
		// Variable binding pattern
		p.advance()
		return interpreter.VariablePattern{Name: name}, nil

	case LBRACE:
		// Object destructuring pattern: {name, age} or {name: n, age: a}
		return p.parseObjectPattern()

	case LBRACKET:
		// Array destructuring pattern: [first, second] or [head, ...rest]
		return p.parseArrayPattern()

	default:
		return nil, p.errorWithHint(
			fmt.Sprintf("Unexpected token in pattern: %s", p.current().Type),
			p.current(),
			"Patterns can be literals, variables, _ (wildcard), {fields}, or [elements]",
		)
	}
}

// parseObjectPattern parses an object destructuring pattern: {name, age} or {name: n}
func (p *Parser) parseObjectPattern() (interpreter.Pattern, error) {
	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var fields []interpreter.ObjectPatternField

	for !p.check(RBRACE) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACE) {
			break
		}

		// Get field name
		fieldName, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		var fieldPattern interpreter.Pattern

		// Check for binding: fieldName: bindingName
		if p.check(COLON) {
			p.advance()
			fieldPattern, err = p.parsePattern()
			if err != nil {
				return nil, err
			}
		}
		// If no colon, fieldPattern stays nil and the field name is used as variable

		fields = append(fields, interpreter.ObjectPatternField{
			Key:     fieldName,
			Pattern: fieldPattern,
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

	return interpreter.ObjectPattern{Fields: fields}, nil
}

// parseArrayPattern parses an array destructuring pattern: [first, second] or [head, ...rest]
func (p *Parser) parseArrayPattern() (interpreter.Pattern, error) {
	if err := p.expect(LBRACKET); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var elements []interpreter.Pattern
	var rest *string

	for !p.check(RBRACKET) && !p.isAtEnd() {
		p.skipNewlines()

		if p.check(RBRACKET) {
			break
		}

		// Check for rest pattern: ...rest
		if p.check(DOTDOTDOT) {
			p.advance()
			restName, err := p.expectIdent()
			if err != nil {
				return nil, p.errorWithHint(
					"Expected identifier after ...",
					p.current(),
					"Rest patterns require a variable name: [...rest]",
				)
			}
			rest = &restName
			// Rest must be last element
			p.skipNewlines()
			break
		}

		// Parse regular element pattern
		elem, err := p.parsePattern()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		if !p.match(COMMA) {
			break
		}

		p.skipNewlines()
	}

	p.skipNewlines()
	if err := p.expect(RBRACKET); err != nil {
		return nil, err
	}

	return interpreter.ArrayPattern{
		Elements: elements,
		Rest:     rest,
	}, nil
}

// ParseExpression parses a single expression from the token stream.
// This is used by the REPL and other tools that need to parse expressions directly.
func (p *Parser) ParseExpression() (interpreter.Expr, error) {
	// Skip leading newlines
	p.skipNewlines()

	if p.isAtEnd() {
		return nil, fmt.Errorf("unexpected end of input")
	}

	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	// Skip trailing newlines
	p.skipNewlines()

	return expr, nil
}

// parseTestBlock parses a test block: test "name" { body }
// "test" is parsed as an IDENT (not a dedicated token) to avoid conflicts with
// identifiers like /test in route paths.
func (p *Parser) parseTestBlock() (interpreter.Item, error) {
	// Consume "test" identifier
	p.advance()

	// Parse test name (string literal)
	if !p.check(STRING) {
		return nil, p.errorWithHint(
			fmt.Sprintf("Expected test name (string), got %s", p.current().Type),
			p.current(),
			"Test blocks must have a name: test \"description\" { ... }",
		)
	}
	name := p.current().Literal
	p.advance()

	p.skipNewlines()

	// Parse body
	if err := p.expect(LBRACE); err != nil {
		return nil, err
	}

	p.skipNewlines()

	var body []interpreter.Statement
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

	return &interpreter.TestBlock{
		Name: name,
		Body: body,
	}, nil
}

// parseAssertStatement parses an assert statement: assert(condition) or assert(condition, "message")
func (p *Parser) parseAssertStatement() (interpreter.Statement, error) {
	if err := p.expect(ASSERT); err != nil {
		return nil, err
	}

	if err := p.expect(LPAREN); err != nil {
		return nil, p.errorWithHint(
			fmt.Sprintf("Expected '(' after assert, got %s", p.current().Type),
			p.current(),
			"Assert syntax: assert(condition) or assert(condition, \"message\")",
		)
	}

	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	var message interpreter.Expr
	if p.match(COMMA) {
		message, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if err := p.expect(RPAREN); err != nil {
		return nil, err
	}

	return interpreter.AssertStatement{
		Condition: condition,
		Message:   message,
	}, nil
}

// ParseStatement parses a single statement from the token stream.
// This is used by the REPL and other tools that need to parse statements directly.
func (p *Parser) ParseStatement() (interpreter.Statement, error) {
	// Skip leading newlines
	p.skipNewlines()

	if p.isAtEnd() {
		return nil, fmt.Errorf("unexpected end of input")
	}

	stmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	// Skip trailing newlines
	p.skipNewlines()

	return stmt, nil
}
