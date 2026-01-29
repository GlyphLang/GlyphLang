package graphql

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Query represents a parsed GraphQL query.
type Query struct {
	OperationType string // "query" or "mutation"
	Name          string // Optional operation name
	Selections    []Selection
	Variables     map[string]interface{}
}

// Selection represents a field selection in a GraphQL query.
type Selection struct {
	Name       string
	Alias      string
	Args       map[string]interface{}
	Selections []Selection // Nested field selections
}

// EffectiveName returns the alias if set, otherwise the field name.
func (s Selection) EffectiveName() string {
	if s.Alias != "" {
		return s.Alias
	}
	return s.Name
}

// ParseQuery parses a GraphQL query string into a Query struct.
// Supports: query { ... }, mutation { ... }, { ... } (implicit query),
// field arguments, aliases, and nested selections.
func ParseQuery(input string) (*Query, error) {
	p := &queryParser{
		input: input,
		pos:   0,
	}
	return p.parse()
}

type queryParser struct {
	input string
	pos   int
}

func (p *queryParser) parse() (*Query, error) {
	p.skipWhitespace()

	q := &Query{
		OperationType: "query",
		Variables:     make(map[string]interface{}),
	}

	// Check for operation type keyword (query, mutation, subscription)
	word := p.peekWord()
	if word == "query" || word == "mutation" || word == "subscription" {
		q.OperationType = p.readWord()
		p.skipWhitespace()

		// Optional operation name
		if p.pos < len(p.input) && p.input[p.pos] != '{' && p.input[p.pos] != '(' {
			q.Name = p.readWord()
			p.skipWhitespace()
		}

		// Optional variable definitions (skip for now)
		if p.pos < len(p.input) && p.input[p.pos] == '(' {
			if err := p.skipBalanced('(', ')'); err != nil {
				return nil, err
			}
			p.skipWhitespace()
		}
	}

	// Parse selection set
	selections, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	q.Selections = selections

	return q, nil
}

func (p *queryParser) parseSelectionSet() ([]Selection, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, fmt.Errorf("expected '{' at position %d", p.pos)
	}
	p.pos++ // consume '{'
	p.skipWhitespace()

	var selections []Selection
	for p.pos < len(p.input) && p.input[p.pos] != '}' {
		sel, err := p.parseSelection()
		if err != nil {
			return nil, err
		}
		selections = append(selections, sel)
		p.skipWhitespace()
	}

	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of query, expected '}'")
	}
	p.pos++ // consume '}'

	return selections, nil
}

func (p *queryParser) parseSelection() (Selection, error) {
	p.skipWhitespace()

	name := p.readWord()
	if name == "" {
		return Selection{}, fmt.Errorf("expected field name at position %d", p.pos)
	}

	sel := Selection{Name: name}

	p.skipWhitespace()

	// Check for alias: aliasName: fieldName
	if p.pos < len(p.input) && p.input[p.pos] == ':' {
		p.pos++ // consume ':'
		p.skipWhitespace()
		sel.Alias = name
		sel.Name = p.readWord()
		p.skipWhitespace()
	}

	// Parse optional arguments
	if p.pos < len(p.input) && p.input[p.pos] == '(' {
		args, err := p.parseArguments()
		if err != nil {
			return Selection{}, err
		}
		sel.Args = args
		p.skipWhitespace()
	}

	// Parse optional nested selection set
	if p.pos < len(p.input) && p.input[p.pos] == '{' {
		nested, err := p.parseSelectionSet()
		if err != nil {
			return Selection{}, err
		}
		sel.Selections = nested
	}

	return sel, nil
}

func (p *queryParser) parseArguments() (map[string]interface{}, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		return nil, fmt.Errorf("expected '(' at position %d", p.pos)
	}
	p.pos++ // consume '('
	p.skipWhitespace()

	args := make(map[string]interface{})

	for p.pos < len(p.input) && p.input[p.pos] != ')' {
		// Parse argument name
		argName := p.readWord()
		if argName == "" {
			return nil, fmt.Errorf("expected argument name at position %d", p.pos)
		}

		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != ':' {
			return nil, fmt.Errorf("expected ':' after argument name at position %d", p.pos)
		}
		p.pos++ // consume ':'
		p.skipWhitespace()

		// Parse argument value
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		args[argName] = val

		p.skipWhitespace()
		// Skip optional comma
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			p.skipWhitespace()
		}
	}

	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of arguments, expected ')'")
	}
	p.pos++ // consume ')'

	return args, nil
}

func (p *queryParser) parseValue() (interface{}, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of input while parsing value")
	}

	ch := p.input[p.pos]

	// String value
	if ch == '"' {
		return p.readString()
	}

	// Boolean or null
	word := p.peekWord()
	switch word {
	case "true":
		p.readWord()
		return true, nil
	case "false":
		p.readWord()
		return false, nil
	case "null":
		p.readWord()
		return nil, nil
	}

	// Number (int or float)
	if ch == '-' || (ch >= '0' && ch <= '9') {
		return p.readNumber()
	}

	// Enum value (unquoted identifier)
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		return p.readWord(), nil
	}

	return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, p.pos)
}

func (p *queryParser) readString() (string, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '"' {
		return "", fmt.Errorf("expected '\"' at position %d", p.pos)
	}
	p.pos++ // consume opening quote

	var b strings.Builder
	for p.pos < len(p.input) && p.input[p.pos] != '"' {
		if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) {
			p.pos++
			switch p.input[p.pos] {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(p.input[p.pos])
			}
		} else {
			b.WriteByte(p.input[p.pos])
		}
		p.pos++
	}

	if p.pos >= len(p.input) {
		return "", fmt.Errorf("unterminated string")
	}
	p.pos++ // consume closing quote

	return b.String(), nil
}

func (p *queryParser) readNumber() (interface{}, error) {
	start := p.pos
	isFloat := false

	if p.pos < len(p.input) && p.input[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if p.pos < len(p.input) && p.input[p.pos] == '.' {
		isFloat = true
		p.pos++
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
			p.pos++
		}
	}

	numStr := p.input[start:p.pos]
	if isFloat {
		f, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float: %s", numStr)
		}
		return f, nil
	}
	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %s", numStr)
	}
	return n, nil
}

func (p *queryParser) readWord() string {
	start := p.pos
	for p.pos < len(p.input) && (unicode.IsLetter(rune(p.input[p.pos])) || unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '_') {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *queryParser) peekWord() string {
	saved := p.pos
	word := p.readWord()
	p.pos = saved
	return word
}

func (p *queryParser) skipWhitespace() {
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == ',' {
			p.pos++
		} else if ch == '#' {
			// Skip comments
			for p.pos < len(p.input) && p.input[p.pos] != '\n' {
				p.pos++
			}
		} else {
			break
		}
	}
}

func (p *queryParser) skipBalanced(open, close byte) error {
	if p.pos >= len(p.input) || p.input[p.pos] != open {
		return fmt.Errorf("expected '%c'", open)
	}
	depth := 1
	p.pos++
	for p.pos < len(p.input) && depth > 0 {
		if p.input[p.pos] == open {
			depth++
		} else if p.input[p.pos] == close {
			depth--
		}
		p.pos++
	}
	if depth != 0 {
		return fmt.Errorf("unbalanced '%c'", open)
	}
	return nil
}
