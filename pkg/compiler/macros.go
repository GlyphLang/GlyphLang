package compiler

import (
	"fmt"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// MacroExpander handles compile-time macro expansion
type MacroExpander struct {
	macros map[string]*interpreter.MacroDef
}

// NewMacroExpander creates a new macro expander
func NewMacroExpander() *MacroExpander {
	return &MacroExpander{
		macros: make(map[string]*interpreter.MacroDef),
	}
}

// RegisterMacro registers a macro definition
func (e *MacroExpander) RegisterMacro(macro *interpreter.MacroDef) {
	e.macros[macro.Name] = macro
}

// GetMacro retrieves a registered macro by name
func (e *MacroExpander) GetMacro(name string) (*interpreter.MacroDef, bool) {
	macro, ok := e.macros[name]
	return macro, ok
}

// ExpandModule expands all macros in a module
// This performs compile-time macro expansion before the module is compiled
func (e *MacroExpander) ExpandModule(module *interpreter.Module) (*interpreter.Module, error) {
	// First pass: register all macro definitions
	for _, item := range module.Items {
		if macro, ok := item.(*interpreter.MacroDef); ok {
			e.RegisterMacro(macro)
		}
	}

	// Second pass: expand all macro invocations
	expandedItems := make([]interpreter.Item, 0, len(module.Items))
	for _, item := range module.Items {
		// Skip macro definitions - they don't produce code
		if _, ok := item.(*interpreter.MacroDef); ok {
			continue
		}

		// Expand macro invocations
		if inv, ok := item.(*interpreter.MacroInvocation); ok {
			expanded, err := e.ExpandMacroInvocation(inv)
			if err != nil {
				return nil, err
			}
			// Add all expanded items
			for _, exp := range expanded {
				if expItem, ok := exp.(interpreter.Item); ok {
					expandedItems = append(expandedItems, expItem)
				}
			}
			continue
		}

		// Recursively expand macros in other items
		expandedItem, err := e.expandItem(item)
		if err != nil {
			return nil, err
		}
		expandedItems = append(expandedItems, expandedItem)
	}

	return &interpreter.Module{Items: expandedItems}, nil
}

// ExpandMacroInvocation expands a single macro invocation
func (e *MacroExpander) ExpandMacroInvocation(inv *interpreter.MacroInvocation) ([]interpreter.Node, error) {
	macro, ok := e.macros[inv.Name]
	if !ok {
		return nil, fmt.Errorf("undefined macro: %s", inv.Name)
	}

	if len(inv.Args) != len(macro.Params) {
		return nil, fmt.Errorf("macro %s expects %d arguments, got %d",
			inv.Name, len(macro.Params), len(inv.Args))
	}

	// Create substitution map
	substitutions := make(map[string]interpreter.Expr)
	for i, param := range macro.Params {
		substitutions[param] = inv.Args[i]
	}

	// Expand macro body with substitutions
	expandedNodes := make([]interpreter.Node, 0, len(macro.Body))
	for _, node := range macro.Body {
		expanded, err := e.substituteNode(node, substitutions)
		if err != nil {
			return nil, err
		}
		// Recursively expand any nested macro invocations
		if expandedInv, ok := expanded.(*interpreter.MacroInvocation); ok {
			nested, err := e.ExpandMacroInvocation(expandedInv)
			if err != nil {
				return nil, err
			}
			expandedNodes = append(expandedNodes, nested...)
		} else {
			expandedNodes = append(expandedNodes, expanded)
		}
	}

	return expandedNodes, nil
}

// expandItem recursively expands macros in an item
func (e *MacroExpander) expandItem(item interpreter.Item) (interpreter.Item, error) {
	switch it := item.(type) {
	case *interpreter.Route:
		expandedBody, err := e.expandStatements(it.Body)
		if err != nil {
			return nil, err
		}
		return &interpreter.Route{
			Path:        it.Path,
			Method:      it.Method,
			ReturnType:  it.ReturnType,
			Auth:        it.Auth,
			RateLimit:   it.RateLimit,
			Injections:  it.Injections,
			QueryParams: it.QueryParams,
			Body:        expandedBody,
		}, nil

	case *interpreter.Command:
		expandedBody, err := e.expandStatements(it.Body)
		if err != nil {
			return nil, err
		}
		return &interpreter.Command{
			Name:        it.Name,
			Description: it.Description,
			Params:      it.Params,
			ReturnType:  it.ReturnType,
			Body:        expandedBody,
		}, nil

	case *interpreter.CronTask:
		expandedBody, err := e.expandStatements(it.Body)
		if err != nil {
			return nil, err
		}
		return &interpreter.CronTask{
			Name:       it.Name,
			Schedule:   it.Schedule,
			Timezone:   it.Timezone,
			Retries:    it.Retries,
			Injections: it.Injections,
			Body:       expandedBody,
		}, nil

	case *interpreter.EventHandler:
		expandedBody, err := e.expandStatements(it.Body)
		if err != nil {
			return nil, err
		}
		return &interpreter.EventHandler{
			EventType:  it.EventType,
			Async:      it.Async,
			Injections: it.Injections,
			Body:       expandedBody,
		}, nil

	case *interpreter.QueueWorker:
		expandedBody, err := e.expandStatements(it.Body)
		if err != nil {
			return nil, err
		}
		return &interpreter.QueueWorker{
			QueueName:   it.QueueName,
			Concurrency: it.Concurrency,
			MaxRetries:  it.MaxRetries,
			Timeout:     it.Timeout,
			Injections:  it.Injections,
			Body:        expandedBody,
		}, nil

	default:
		return item, nil
	}
}

// expandStatements expands macros in a list of statements
func (e *MacroExpander) expandStatements(stmts []interpreter.Statement) ([]interpreter.Statement, error) {
	result := make([]interpreter.Statement, 0, len(stmts))
	for _, stmt := range stmts {
		// Check if statement is a macro invocation
		if inv, ok := stmt.(interpreter.MacroInvocation); ok {
			expanded, err := e.ExpandMacroInvocation(&inv)
			if err != nil {
				return nil, err
			}
			for _, node := range expanded {
				if s, ok := node.(interpreter.Statement); ok {
					result = append(result, s)
				}
			}
			continue
		}

		expandedStmt, err := e.expandStatement(stmt)
		if err != nil {
			return nil, err
		}
		result = append(result, expandedStmt)
	}
	return result, nil
}

// expandStatement recursively expands macros in a statement
func (e *MacroExpander) expandStatement(stmt interpreter.Statement) (interpreter.Statement, error) {
	switch s := stmt.(type) {
	case interpreter.IfStatement:
		thenBlock, err := e.expandStatements(s.ThenBlock)
		if err != nil {
			return nil, err
		}
		elseBlock, err := e.expandStatements(s.ElseBlock)
		if err != nil {
			return nil, err
		}
		return interpreter.IfStatement{
			Condition: s.Condition,
			ThenBlock: thenBlock,
			ElseBlock: elseBlock,
		}, nil

	case interpreter.WhileStatement:
		body, err := e.expandStatements(s.Body)
		if err != nil {
			return nil, err
		}
		return interpreter.WhileStatement{
			Condition: s.Condition,
			Body:      body,
		}, nil

	case interpreter.ForStatement:
		body, err := e.expandStatements(s.Body)
		if err != nil {
			return nil, err
		}
		return interpreter.ForStatement{
			KeyVar:   s.KeyVar,
			ValueVar: s.ValueVar,
			Iterable: s.Iterable,
			Body:     body,
		}, nil

	case interpreter.SwitchStatement:
		expandedCases := make([]interpreter.SwitchCase, len(s.Cases))
		for i, c := range s.Cases {
			body, err := e.expandStatements(c.Body)
			if err != nil {
				return nil, err
			}
			expandedCases[i] = interpreter.SwitchCase{
				Value: c.Value,
				Body:  body,
			}
		}
		defaultBody, err := e.expandStatements(s.Default)
		if err != nil {
			return nil, err
		}
		return interpreter.SwitchStatement{
			Value:   s.Value,
			Cases:   expandedCases,
			Default: defaultBody,
		}, nil

	default:
		return stmt, nil
	}
}

// substituteNode performs parameter substitution in a node
func (e *MacroExpander) substituteNode(node interpreter.Node, subs map[string]interpreter.Expr) (interpreter.Node, error) {
	switch n := node.(type) {
	case *interpreter.Route:
		expandedBody, err := e.substituteStatements(n.Body, subs)
		if err != nil {
			return nil, err
		}
		// Substitute in path (for dynamic route generation)
		path := e.substituteString(n.Path, subs)
		return &interpreter.Route{
			Path:        path,
			Method:      n.Method,
			ReturnType:  n.ReturnType,
			Auth:        n.Auth,
			RateLimit:   n.RateLimit,
			Injections:  n.Injections,
			QueryParams: n.QueryParams,
			Body:        expandedBody,
		}, nil

	case interpreter.AssignStatement:
		subExpr, err := e.substituteExpr(n.Value, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.AssignStatement{
			Target: e.substituteString(n.Target, subs),
			Value:  subExpr,
		}, nil

	case interpreter.ReturnStatement:
		subExpr, err := e.substituteExpr(n.Value, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.ReturnStatement{Value: subExpr}, nil

	case interpreter.IfStatement:
		cond, err := e.substituteExpr(n.Condition, subs)
		if err != nil {
			return nil, err
		}
		thenBlock, err := e.substituteStatements(n.ThenBlock, subs)
		if err != nil {
			return nil, err
		}
		elseBlock, err := e.substituteStatements(n.ElseBlock, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.IfStatement{
			Condition: cond,
			ThenBlock: thenBlock,
			ElseBlock: elseBlock,
		}, nil

	case interpreter.WhileStatement:
		cond, err := e.substituteExpr(n.Condition, subs)
		if err != nil {
			return nil, err
		}
		body, err := e.substituteStatements(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.WhileStatement{
			Condition: cond,
			Body:      body,
		}, nil

	case interpreter.ForStatement:
		iter, err := e.substituteExpr(n.Iterable, subs)
		if err != nil {
			return nil, err
		}
		body, err := e.substituteStatements(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.ForStatement{
			KeyVar:   n.KeyVar,
			ValueVar: n.ValueVar,
			Iterable: iter,
			Body:     body,
		}, nil

	case interpreter.ExpressionStatement:
		subExpr, err := e.substituteExpr(n.Expr, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.ExpressionStatement{Expr: subExpr}, nil

	case *interpreter.MacroInvocation:
		// Substitute in macro invocation arguments
		subArgs := make([]interpreter.Expr, len(n.Args))
		for i, arg := range n.Args {
			subArg, err := e.substituteExpr(arg, subs)
			if err != nil {
				return nil, err
			}
			subArgs[i] = subArg
		}
		return &interpreter.MacroInvocation{
			Name: n.Name,
			Args: subArgs,
		}, nil

	default:
		return node, nil
	}
}

// substituteStatements performs substitution in a list of statements
func (e *MacroExpander) substituteStatements(stmts []interpreter.Statement, subs map[string]interpreter.Expr) ([]interpreter.Statement, error) {
	result := make([]interpreter.Statement, len(stmts))
	for i, stmt := range stmts {
		subNode, err := e.substituteNode(stmt, subs)
		if err != nil {
			return nil, err
		}
		result[i] = subNode.(interpreter.Statement)
	}
	return result, nil
}

// substituteExpr performs parameter substitution in an expression
func (e *MacroExpander) substituteExpr(expr interpreter.Expr, subs map[string]interpreter.Expr) (interpreter.Expr, error) {
	switch ex := expr.(type) {
	case interpreter.VariableExpr:
		// Check if this variable should be substituted
		if sub, ok := subs[ex.Name]; ok {
			return sub, nil
		}
		return ex, nil

	case interpreter.BinaryOpExpr:
		left, err := e.substituteExpr(ex.Left, subs)
		if err != nil {
			return nil, err
		}
		right, err := e.substituteExpr(ex.Right, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.BinaryOpExpr{
			Op:    ex.Op,
			Left:  left,
			Right: right,
		}, nil

	case interpreter.UnaryOpExpr:
		right, err := e.substituteExpr(ex.Right, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.UnaryOpExpr{
			Op:    ex.Op,
			Right: right,
		}, nil

	case interpreter.FunctionCallExpr:
		subArgs := make([]interpreter.Expr, len(ex.Args))
		for i, arg := range ex.Args {
			subArg, err := e.substituteExpr(arg, subs)
			if err != nil {
				return nil, err
			}
			subArgs[i] = subArg
		}
		return interpreter.FunctionCallExpr{
			Name: ex.Name,
			Args: subArgs,
		}, nil

	case interpreter.FieldAccessExpr:
		obj, err := e.substituteExpr(ex.Object, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.FieldAccessExpr{
			Object: obj,
			Field:  ex.Field,
		}, nil

	case interpreter.ArrayIndexExpr:
		arr, err := e.substituteExpr(ex.Array, subs)
		if err != nil {
			return nil, err
		}
		idx, err := e.substituteExpr(ex.Index, subs)
		if err != nil {
			return nil, err
		}
		return interpreter.ArrayIndexExpr{
			Array: arr,
			Index: idx,
		}, nil

	case interpreter.ObjectExpr:
		subFields := make([]interpreter.ObjectField, len(ex.Fields))
		for i, field := range ex.Fields {
			subVal, err := e.substituteExpr(field.Value, subs)
			if err != nil {
				return nil, err
			}
			subFields[i] = interpreter.ObjectField{
				Key:   field.Key,
				Value: subVal,
			}
		}
		return interpreter.ObjectExpr{Fields: subFields}, nil

	case interpreter.ArrayExpr:
		subElems := make([]interpreter.Expr, len(ex.Elements))
		for i, elem := range ex.Elements {
			subElem, err := e.substituteExpr(elem, subs)
			if err != nil {
				return nil, err
			}
			subElems[i] = subElem
		}
		return interpreter.ArrayExpr{Elements: subElems}, nil

	case interpreter.LiteralExpr:
		// Check if string literal contains parameter references for string interpolation
		if strLit, ok := ex.Value.(interpreter.StringLiteral); ok {
			newVal := e.substituteString(strLit.Value, subs)
			return interpreter.LiteralExpr{
				Value: interpreter.StringLiteral{Value: newVal},
			}, nil
		}
		return ex, nil

	case interpreter.UnquoteExpr:
		// Evaluate the unquote expression - substitute and return the inner expr
		return e.substituteExpr(ex.Expr, subs)

	default:
		return expr, nil
	}
}

// substituteString performs string interpolation for ${param} syntax
func (e *MacroExpander) substituteString(s string, subs map[string]interpreter.Expr) string {
	result := s
	for param, expr := range subs {
		// Handle ${param} interpolation
		placeholder := "${" + param + "}"
		if litExpr, ok := expr.(interpreter.LiteralExpr); ok {
			if strLit, ok := litExpr.Value.(interpreter.StringLiteral); ok {
				result = replaceAll(result, placeholder, strLit.Value)
			} else if intLit, ok := litExpr.Value.(interpreter.IntLiteral); ok {
				result = replaceAll(result, placeholder, fmt.Sprintf("%d", intLit.Value))
			}
		} else if varExpr, ok := expr.(interpreter.VariableExpr); ok {
			result = replaceAll(result, placeholder, varExpr.Name)
		}
	}
	return result
}

// replaceAll replaces all occurrences of old with new in s
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
