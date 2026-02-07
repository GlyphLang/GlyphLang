package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"fmt"
	"strings"
)

const maxMacroExpansionDepth = 100

// expandMacro expands a macro invocation into a list of AST nodes.
func (i *Interpreter) expandMacro(inv *MacroInvocation) ([]Node, error) {
	return i.expandMacroWithDepth(inv, 0)
}

func (i *Interpreter) expandMacroWithDepth(inv *MacroInvocation, depth int) ([]Node, error) {
	if depth >= maxMacroExpansionDepth {
		return nil, fmt.Errorf("macro expansion depth limit exceeded (%d) while expanding %q",
			maxMacroExpansionDepth, inv.Name)
	}

	macro, ok := i.macros[inv.Name]
	if !ok {
		return nil, fmt.Errorf("undefined macro: %s", inv.Name)
	}

	if len(inv.Args) != len(macro.Params) {
		return nil, fmt.Errorf("macro %s expects %d arguments, got %d",
			inv.Name, len(macro.Params), len(inv.Args))
	}

	subs := make(map[string]Expr)
	for idx, param := range macro.Params {
		subs[param] = inv.Args[idx]
	}

	expanded := make([]Node, 0, len(macro.Body))
	for _, node := range macro.Body {
		sub, err := i.substituteNode(node, subs)
		if err != nil {
			return nil, err
		}
		if nestedInv, ok := sub.(*MacroInvocation); ok {
			nested, err := i.expandMacroWithDepth(nestedInv, depth+1)
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, nested...)
		} else {
			expanded = append(expanded, sub)
		}
	}

	return expanded, nil
}

// executeMacroInvocation expands and executes a macro invocation as a statement.
func (i *Interpreter) executeMacroInvocation(inv MacroInvocation, env *Environment) (interface{}, error) {
	expanded, err := i.expandMacro(&inv)
	if err != nil {
		return nil, err
	}
	var result interface{}
	for _, node := range expanded {
		stmt, ok := node.(Statement)
		if !ok {
			return nil, fmt.Errorf("macro %s expanded to non-statement node: %T", inv.Name, node)
		}
		result, err = i.ExecuteStatement(stmt, env)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// evaluateMacroInvocation expands and evaluates a macro invocation as an expression.
func (i *Interpreter) evaluateMacroInvocation(inv MacroInvocation, env *Environment) (interface{}, error) {
	expanded, err := i.expandMacro(&inv)
	if err != nil {
		return nil, err
	}
	var result interface{}
	for _, node := range expanded {
		switch n := node.(type) {
		case Statement:
			result, err = i.ExecuteStatement(n, env)
			if err != nil {
				return nil, err
			}
		case Expr:
			result, err = i.EvaluateExpression(n, env)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("macro %s expanded to unsupported node type: %T", inv.Name, node)
		}
	}
	return result, nil
}

// substituteNode performs parameter substitution in a node.
func (i *Interpreter) substituteNode(node Node, subs map[string]Expr) (Node, error) {
	switch n := node.(type) {
	case AssignStatement:
		subExpr, err := i.substituteExpr(n.Value, subs)
		if err != nil {
			return nil, err
		}
		return AssignStatement{
			Target: i.substituteString(n.Target, subs),
			Value:  subExpr,
		}, nil

	case ReassignStatement:
		subExpr, err := i.substituteExpr(n.Value, subs)
		if err != nil {
			return nil, err
		}
		return ReassignStatement{
			Target: i.substituteString(n.Target, subs),
			Value:  subExpr,
		}, nil

	case ReturnStatement:
		subExpr, err := i.substituteExpr(n.Value, subs)
		if err != nil {
			return nil, err
		}
		return ReturnStatement{Value: subExpr}, nil

	case IfStatement:
		cond, err := i.substituteExpr(n.Condition, subs)
		if err != nil {
			return nil, err
		}
		thenBlock, err := i.substituteStatements(n.ThenBlock, subs)
		if err != nil {
			return nil, err
		}
		elseBlock, err := i.substituteStatements(n.ElseBlock, subs)
		if err != nil {
			return nil, err
		}
		return IfStatement{
			Condition: cond,
			ThenBlock: thenBlock,
			ElseBlock: elseBlock,
		}, nil

	case WhileStatement:
		cond, err := i.substituteExpr(n.Condition, subs)
		if err != nil {
			return nil, err
		}
		body, err := i.substituteStatements(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return WhileStatement{
			Condition: cond,
			Body:      body,
		}, nil

	case ForStatement:
		iter, err := i.substituteExpr(n.Iterable, subs)
		if err != nil {
			return nil, err
		}
		body, err := i.substituteStatements(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return ForStatement{
			KeyVar:   n.KeyVar,
			ValueVar: n.ValueVar,
			Iterable: iter,
			Body:     body,
		}, nil

	case ExpressionStatement:
		subExpr, err := i.substituteExpr(n.Expr, subs)
		if err != nil {
			return nil, err
		}
		return ExpressionStatement{Expr: subExpr}, nil

	case *MacroInvocation:
		subArgs := make([]Expr, len(n.Args))
		for idx, arg := range n.Args {
			subArg, err := i.substituteExpr(arg, subs)
			if err != nil {
				return nil, err
			}
			subArgs[idx] = subArg
		}
		return &MacroInvocation{
			Name: n.Name,
			Args: subArgs,
		}, nil

	default:
		return node, nil
	}
}

// substituteStatements performs substitution in a list of statements.
func (i *Interpreter) substituteStatements(stmts []Statement, subs map[string]Expr) ([]Statement, error) {
	result := make([]Statement, 0, len(stmts))
	for _, stmt := range stmts {
		subNode, err := i.substituteNode(stmt, subs)
		if err != nil {
			return nil, err
		}
		subStmt, ok := subNode.(Statement)
		if !ok {
			return nil, fmt.Errorf("macro substitution produced non-statement node: %T", subNode)
		}
		result = append(result, subStmt)
	}
	return result, nil
}

// substituteExpr performs parameter substitution in an expression.
func (i *Interpreter) substituteExpr(expr Expr, subs map[string]Expr) (Expr, error) {
	switch ex := expr.(type) {
	case VariableExpr:
		if sub, ok := subs[ex.Name]; ok {
			return sub, nil
		}
		return ex, nil

	case BinaryOpExpr:
		left, err := i.substituteExpr(ex.Left, subs)
		if err != nil {
			return nil, err
		}
		right, err := i.substituteExpr(ex.Right, subs)
		if err != nil {
			return nil, err
		}
		return BinaryOpExpr{Op: ex.Op, Left: left, Right: right}, nil

	case UnaryOpExpr:
		right, err := i.substituteExpr(ex.Right, subs)
		if err != nil {
			return nil, err
		}
		return UnaryOpExpr{Op: ex.Op, Right: right}, nil

	case FunctionCallExpr:
		subArgs := make([]Expr, len(ex.Args))
		for idx, arg := range ex.Args {
			subArg, err := i.substituteExpr(arg, subs)
			if err != nil {
				return nil, err
			}
			subArgs[idx] = subArg
		}
		return FunctionCallExpr{Name: ex.Name, Args: subArgs}, nil

	case FieldAccessExpr:
		obj, err := i.substituteExpr(ex.Object, subs)
		if err != nil {
			return nil, err
		}
		return FieldAccessExpr{Object: obj, Field: ex.Field}, nil

	case ArrayIndexExpr:
		arr, err := i.substituteExpr(ex.Array, subs)
		if err != nil {
			return nil, err
		}
		idx, err := i.substituteExpr(ex.Index, subs)
		if err != nil {
			return nil, err
		}
		return ArrayIndexExpr{Array: arr, Index: idx}, nil

	case ObjectExpr:
		subFields := make([]ObjectField, len(ex.Fields))
		for idx, field := range ex.Fields {
			subVal, err := i.substituteExpr(field.Value, subs)
			if err != nil {
				return nil, err
			}
			subFields[idx] = ObjectField{Key: field.Key, Value: subVal}
		}
		return ObjectExpr{Fields: subFields}, nil

	case ArrayExpr:
		subElems := make([]Expr, len(ex.Elements))
		for idx, elem := range ex.Elements {
			subElem, err := i.substituteExpr(elem, subs)
			if err != nil {
				return nil, err
			}
			subElems[idx] = subElem
		}
		return ArrayExpr{Elements: subElems}, nil

	case LiteralExpr:
		if strLit, ok := ex.Value.(StringLiteral); ok {
			newVal := i.substituteString(strLit.Value, subs)
			return LiteralExpr{Value: StringLiteral{Value: newVal}}, nil
		}
		return ex, nil

	case UnquoteExpr:
		return i.substituteExpr(ex.Expr, subs)

	default:
		return expr, nil
	}
}

// substituteString performs ${param} interpolation in strings.
// Only literal and variable expressions can be interpolated into strings.
func (i *Interpreter) substituteString(s string, subs map[string]Expr) string {
	result := s
	for param, expr := range subs {
		placeholder := "${" + param + "}"
		if !strings.Contains(result, placeholder) {
			continue
		}
		switch e := expr.(type) {
		case LiteralExpr:
			switch lit := e.Value.(type) {
			case StringLiteral:
				result = strings.ReplaceAll(result, placeholder, lit.Value)
			case IntLiteral:
				result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%d", lit.Value))
			}
		case VariableExpr:
			result = strings.ReplaceAll(result, placeholder, e.Name)
		}
	}
	return result
}
