package compiler

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
)

// OptimizationLevel defines how aggressive optimization should be
type OptimizationLevel int

const (
	OptNone OptimizationLevel = iota
	OptBasic
	OptAggressive
)

// Optimizer performs optimization passes on AST
type Optimizer struct {
	level       OptimizationLevel
	constants   map[string]ast.Literal // Track constant values for variables
	expressions map[string]string      // Track expression -> variable name for CSE
	copies      map[string]string      // Track variable copies (x = y means copies[x] = y)
}

// NewOptimizer creates a new optimizer instance
func NewOptimizer(level OptimizationLevel) *Optimizer {
	return &Optimizer{
		level:       level,
		constants:   make(map[string]ast.Literal),
		expressions: make(map[string]string),
		copies:      make(map[string]string),
	}
}

// OptimizeExpression optimizes an expression
func (o *Optimizer) OptimizeExpression(expr ast.Expr) ast.Expr {
	if o.level == OptNone {
		return expr
	}

	switch e := expr.(type) {
	case *ast.VariableExpr:
		// Copy propagation: follow copy chains
		varName := e.Name
		if o.level >= OptBasic {
			// Follow copy chain to find the root variable
			visited := make(map[string]bool)
			for {
				if visited[varName] {
					// Cycle detected, stop
					break
				}
				visited[varName] = true

				if copyTarget, ok := o.copies[varName]; ok {
					varName = copyTarget
				} else {
					break
				}
			}
		}

		// Constant propagation: replace variable with constant if known
		if lit, ok := o.constants[varName]; ok {
			return &ast.LiteralExpr{Value: lit}
		}

		// Return the resolved variable (might be different due to copy propagation)
		if varName != e.Name {
			return &ast.VariableExpr{Name: varName}
		}
		return expr
	case *ast.BinaryOpExpr:
		return o.foldBinaryOp(e)
	case *ast.ObjectExpr:
		// Optimize object fields
		fields := make([]ast.ObjectField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = ast.ObjectField{
				Key:   field.Key,
				Value: o.OptimizeExpression(field.Value),
			}
		}
		return &ast.ObjectExpr{Fields: fields}
	case *ast.ArrayExpr:
		// Optimize array elements
		elements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = o.OptimizeExpression(elem)
		}
		return &ast.ArrayExpr{Elements: elements}
	case *ast.FieldAccessExpr:
		// Optimize the object expression
		return &ast.FieldAccessExpr{
			Object: o.OptimizeExpression(e.Object),
			Field:  e.Field,
		}
	default:
		return expr
	}
}

// OptimizeStatements optimizes a list of statements
func (o *Optimizer) OptimizeStatements(stmts []ast.Statement) []ast.Statement {
	if o.level == OptNone {
		return stmts
	}

	result := make([]ast.Statement, 0, len(stmts))
	reachedReturn := false

	for _, stmt := range stmts {
		// Dead code elimination: skip statements after return
		if reachedReturn {
			continue
		}

		switch s := stmt.(type) {
		case *ast.AssignStatement:
			// Optimize the value expression
			optimizedValue := o.OptimizeExpression(s.Value)

			// Copy propagation: track variable-to-variable assignments
			if varExpr, ok := optimizedValue.(*ast.VariableExpr); ok {
				o.copies[s.Target] = varExpr.Name
				// Invalidate constant and expression tracking for this variable
				delete(o.constants, s.Target)
			} else {
				// Not a copy, remove from copy tracking
				delete(o.copies, s.Target)

				// Common subexpression elimination (only for OptAggressive)
				if o.level >= OptAggressive {
					key := exprKey(optimizedValue)
					if key != "" {
						// Check if this expression was already computed
						if existingVar, ok := o.expressions[key]; ok {
							// Reuse the existing variable
							optimizedValue = &ast.VariableExpr{Name: existingVar}
							// Now it's a copy
							o.copies[s.Target] = existingVar
						} else {
							// Track this expression
							o.expressions[key] = s.Target
						}
					}
				}

				// Track constant assignments for constant propagation
				if litExpr, ok := optimizedValue.(*ast.LiteralExpr); ok {
					o.constants[s.Target] = litExpr.Value
				} else {
					// Non-constant assignment, invalidate any previous constant
					delete(o.constants, s.Target)
				}
			}

			optimized := &ast.AssignStatement{
				Target: s.Target,
				Value:  optimizedValue,
			}
			result = append(result, optimized)

		case *ast.ReassignStatement:
			// Optimize the value expression (same logic as AssignStatement)
			optimizedValue := o.OptimizeExpression(s.Value)

			// Copy propagation: track variable-to-variable assignments
			if varExpr, ok := optimizedValue.(*ast.VariableExpr); ok {
				o.copies[s.Target] = varExpr.Name
				// Invalidate constant and expression tracking for this variable
				delete(o.constants, s.Target)
			} else {
				// Not a copy, remove from copy tracking
				delete(o.copies, s.Target)

				// Common subexpression elimination (only for OptAggressive)
				if o.level >= OptAggressive {
					key := exprKey(optimizedValue)
					if key != "" {
						// Check if this expression was already computed
						if existingVar, ok := o.expressions[key]; ok {
							// Reuse the existing variable
							optimizedValue = &ast.VariableExpr{Name: existingVar}
							// Now it's a copy
							o.copies[s.Target] = existingVar
						} else {
							// Track this expression
							o.expressions[key] = s.Target
						}
					}
				}

				// Track constant assignments for constant propagation
				if litExpr, ok := optimizedValue.(*ast.LiteralExpr); ok {
					o.constants[s.Target] = litExpr.Value
				} else {
					// Non-constant assignment, invalidate any previous constant
					delete(o.constants, s.Target)
				}
			}

			optimized := &ast.ReassignStatement{
				Target: s.Target,
				Value:  optimizedValue,
			}
			result = append(result, optimized)

		case ast.ReassignStatement:
			// Same as *ast.ReassignStatement
			optimizedValue := o.OptimizeExpression(s.Value)

			if varExpr, ok := optimizedValue.(*ast.VariableExpr); ok {
				o.copies[s.Target] = varExpr.Name
				delete(o.constants, s.Target)
			} else {
				delete(o.copies, s.Target)

				if o.level >= OptAggressive {
					key := exprKey(optimizedValue)
					if key != "" {
						if existingVar, ok := o.expressions[key]; ok {
							optimizedValue = &ast.VariableExpr{Name: existingVar}
							o.copies[s.Target] = existingVar
						} else {
							o.expressions[key] = s.Target
						}
					}
				}

				if litExpr, ok := optimizedValue.(*ast.LiteralExpr); ok {
					o.constants[s.Target] = litExpr.Value
				} else {
					delete(o.constants, s.Target)
				}
			}

			result = append(result, &ast.ReassignStatement{
				Target: s.Target,
				Value:  optimizedValue,
			})

		case *ast.ReturnStatement:
			// Optimize return value
			optimized := &ast.ReturnStatement{
				Value: o.OptimizeExpression(s.Value),
			}
			result = append(result, optimized)
			reachedReturn = true

		case *ast.IfStatement:
			// Try to optimize the condition
			condition := o.OptimizeExpression(s.Condition)

			// Check if condition is a constant boolean
			if litExpr, ok := condition.(*ast.LiteralExpr); ok {
				if boolLit, ok := litExpr.Value.(ast.BoolLiteral); ok {
					// Constant condition - eliminate dead branch
					if boolLit.Value {
						// Condition is always true - use only then block
						result = append(result, o.OptimizeStatements(s.ThenBlock)...)
					} else {
						// Condition is always false - use only else block
						result = append(result, o.OptimizeStatements(s.ElseBlock)...)
					}
					continue
				}
			}

			// Not a constant condition - optimize both branches
			optimized := &ast.IfStatement{
				Condition: condition,
				ThenBlock: o.OptimizeStatements(s.ThenBlock),
				ElseBlock: o.OptimizeStatements(s.ElseBlock),
			}
			result = append(result, optimized)

		case *ast.WhileStatement:
			// First, invalidate constants for any variables modified in the loop body
			// because the loop may execute multiple times or not at all
			modifiedVars := getModifiedVariables(s.Body)
			for varName := range modifiedVars {
				delete(o.constants, varName)
				delete(o.copies, varName)
				delete(o.expressions, varName)
			}

			// Loop invariant code motion (OptAggressive only)
			var invariantStmts []ast.Statement
			var loopBody []ast.Statement

			if o.level >= OptAggressive {
				// Also check if condition uses any variables
				conditionVars := getUsedVariables(s.Condition)

				// Separate invariant assignments from loop-variant ones
				for _, bodyStmt := range s.Body {
					if assignStmt, ok := bodyStmt.(*ast.AssignStatement); ok {
						// Check if this assignment is loop-invariant
						// An assignment is invariant if:
						// 1. Its RHS doesn't depend on modified variables
						// 2. The target variable is not used in the loop condition
						if isExprInvariant(assignStmt.Value, modifiedVars) && !conditionVars[assignStmt.Target] {
							// This is loop-invariant, move it out
							invariantStmts = append(invariantStmts, assignStmt)
							continue
						}
					}
					loopBody = append(loopBody, bodyStmt)
				}
			} else {
				loopBody = s.Body
			}

			// Add invariant statements before the loop
			for _, invStmt := range invariantStmts {
				result = append(result, o.OptimizeStatements([]ast.Statement{invStmt})...)
			}

			// Optimize condition and remaining loop body
			optimized := &ast.WhileStatement{
				Condition: o.OptimizeExpression(s.Condition),
				Body:      o.OptimizeStatements(loopBody),
			}
			result = append(result, optimized)

		case *ast.ForStatement:
			// Invalidate constants for any variables modified in the for loop body
			// because the loop may execute multiple times or not at all
			modifiedVars := getModifiedVariables(s.Body)
			for varName := range modifiedVars {
				delete(o.constants, varName)
				delete(o.copies, varName)
				delete(o.expressions, varName)
			}
			// Also invalidate the loop variables themselves
			if s.KeyVar != "" {
				delete(o.constants, s.KeyVar)
				delete(o.copies, s.KeyVar)
			}
			delete(o.constants, s.ValueVar)
			delete(o.copies, s.ValueVar)
			// Add the for statement unchanged (could optimize body in future)
			result = append(result, s)

		case ast.ForStatement:
			// Same as *ast.ForStatement
			modifiedVars := getModifiedVariables(s.Body)
			for varName := range modifiedVars {
				delete(o.constants, varName)
				delete(o.copies, varName)
				delete(o.expressions, varName)
			}
			if s.KeyVar != "" {
				delete(o.constants, s.KeyVar)
				delete(o.copies, s.KeyVar)
			}
			delete(o.constants, s.ValueVar)
			delete(o.copies, s.ValueVar)
			result = append(result, &s)

		case *ast.SwitchStatement:
			// Invalidate constants for any variables modified in switch case bodies
			// because we don't know which case will execute at compile time
			for _, switchCase := range s.Cases {
				modifiedVars := getModifiedVariables(switchCase.Body)
				for varName := range modifiedVars {
					delete(o.constants, varName)
					delete(o.copies, varName)
					delete(o.expressions, varName)
				}
			}
			// Also invalidate variables modified in the default case
			if len(s.Default) > 0 {
				modifiedVars := getModifiedVariables(s.Default)
				for varName := range modifiedVars {
					delete(o.constants, varName)
					delete(o.copies, varName)
					delete(o.expressions, varName)
				}
			}
			result = append(result, s)

		case ast.SwitchStatement:
			// Same as *ast.SwitchStatement
			for _, switchCase := range s.Cases {
				modifiedVars := getModifiedVariables(switchCase.Body)
				for varName := range modifiedVars {
					delete(o.constants, varName)
					delete(o.copies, varName)
					delete(o.expressions, varName)
				}
			}
			if len(s.Default) > 0 {
				modifiedVars := getModifiedVariables(s.Default)
				for varName := range modifiedVars {
					delete(o.constants, varName)
					delete(o.copies, varName)
					delete(o.expressions, varName)
				}
			}
			result = append(result, &s)

		default:
			result = append(result, stmt)
		}
	}

	return result
}

// foldBinaryOp performs constant folding on binary operations
func (o *Optimizer) foldBinaryOp(expr *ast.BinaryOpExpr) ast.Expr {
	// First optimize operands recursively
	left := o.OptimizeExpression(expr.Left)
	right := o.OptimizeExpression(expr.Right)

	// Arithmetic identities: x op x = constant (for OptAggressive)
	// NOTE: These optimizations are only safe for integer types.
	// For floats, NaN != NaN (IEEE 754), and x/0 must produce an error.
	// We only apply these when both sides are provably integer-typed literals.
	if o.level >= OptAggressive {
		if areExprsEqual(left, right) && isProvablyIntegerExpr(left) {
			switch expr.Op {
			case ast.Sub:
				// x - x = 0 (safe for integers)
				return &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}}
			case ast.Eq:
				// x == x = true (safe for integers, not for NaN floats)
				return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}
			case ast.Ne:
				// x != x = false (safe for integers, not for NaN floats)
				return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}}
			case ast.Le, ast.Ge:
				// x <= x = true, x >= x = true (safe for integers)
				return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}
			case ast.Lt, ast.Gt:
				// x < x = false, x > x = false (safe for integers)
				return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}}
				// NOTE: x / x = 1 is NOT safe -- x could be 0 at runtime
			}
		}
	}

	// Try to extract literal values
	leftLit, leftIsLit := left.(*ast.LiteralExpr)
	rightLit, rightIsLit := right.(*ast.LiteralExpr)

	// Both operands are literals - fold the operation
	if leftIsLit && rightIsLit {
		return o.foldLiteralBinaryOp(expr.Op, leftLit, rightLit)
	}

	// Algebraic simplifications (only if one operand is literal)
	if leftIsLit || rightIsLit {
		if simplified := o.algebraicSimplify(expr.Op, left, right, leftIsLit); simplified != nil {
			return simplified
		}
	}

	// Return optimized but not fully folded expression
	return &ast.BinaryOpExpr{
		Op:    expr.Op,
		Left:  left,
		Right: right,
	}
}

// isProvablyIntegerExpr returns true if the expression is provably integer-typed
// (e.g., an integer literal). This is used to guard identity optimizations that
// are unsafe for floats (NaN) or zero-valued expressions.
func isProvablyIntegerExpr(e ast.Expr) bool {
	if lit, ok := e.(*ast.LiteralExpr); ok {
		if _, isInt := lit.Value.(ast.IntLiteral); isInt {
			return true
		}
	}
	return false
}

// areExprsEqual checks if two expressions are structurally equal
func areExprsEqual(a, b ast.Expr) bool {
	switch aExpr := a.(type) {
	case *ast.VariableExpr:
		if bExpr, ok := b.(*ast.VariableExpr); ok {
			return aExpr.Name == bExpr.Name
		}
	case *ast.LiteralExpr:
		if bExpr, ok := b.(*ast.LiteralExpr); ok {
			return literalsEqual(aExpr.Value, bExpr.Value)
		}
	}
	return false
}

// literalsEqual compares two literals for equality
func literalsEqual(a, b ast.Literal) bool {
	switch aLit := a.(type) {
	case ast.IntLiteral:
		if bLit, ok := b.(ast.IntLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case ast.FloatLiteral:
		if bLit, ok := b.(ast.FloatLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case ast.BoolLiteral:
		if bLit, ok := b.(ast.BoolLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case ast.StringLiteral:
		if bLit, ok := b.(ast.StringLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case ast.NullLiteral:
		_, ok := b.(ast.NullLiteral)
		return ok
	}
	return false
}

// foldLiteralBinaryOp folds binary operations on two literals
func (o *Optimizer) foldLiteralBinaryOp(op ast.BinOp, left, right *ast.LiteralExpr) ast.Expr {
	// Extract int literals
	leftInt, leftIsInt := left.Value.(ast.IntLiteral)
	rightInt, rightIsInt := right.Value.(ast.IntLiteral)

	// Extract float literals
	leftFloat, leftIsFloat := left.Value.(ast.FloatLiteral)
	rightFloat, rightIsFloat := right.Value.(ast.FloatLiteral)

	// Extract bool literals
	leftBool, leftIsBool := left.Value.(ast.BoolLiteral)
	rightBool, rightIsBool := right.Value.(ast.BoolLiteral)

	// Arithmetic operations on integers
	if leftIsInt && rightIsInt {
		var result int64
		switch op {
		case ast.Add:
			result = leftInt.Value + rightInt.Value
			return &ast.LiteralExpr{Value: ast.IntLiteral{Value: result}}
		case ast.Sub:
			result = leftInt.Value - rightInt.Value
			return &ast.LiteralExpr{Value: ast.IntLiteral{Value: result}}
		case ast.Mul:
			result = leftInt.Value * rightInt.Value
			return &ast.LiteralExpr{Value: ast.IntLiteral{Value: result}}
		case ast.Div:
			if rightInt.Value != 0 {
				result = leftInt.Value / rightInt.Value
				return &ast.LiteralExpr{Value: ast.IntLiteral{Value: result}}
			}
		}

		// Comparison operations on integers
		var boolResult bool
		switch op {
		case ast.Eq:
			boolResult = leftInt.Value == rightInt.Value
		case ast.Ne:
			boolResult = leftInt.Value != rightInt.Value
		case ast.Lt:
			boolResult = leftInt.Value < rightInt.Value
		case ast.Le:
			boolResult = leftInt.Value <= rightInt.Value
		case ast.Gt:
			boolResult = leftInt.Value > rightInt.Value
		case ast.Ge:
			boolResult = leftInt.Value >= rightInt.Value
		default:
			goto noFold
		}
		return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: boolResult}}
	}

	// Arithmetic operations on floats
	if leftIsFloat && rightIsFloat {
		var result float64
		switch op {
		case ast.Add:
			result = leftFloat.Value + rightFloat.Value
			return &ast.LiteralExpr{Value: ast.FloatLiteral{Value: result}}
		case ast.Sub:
			result = leftFloat.Value - rightFloat.Value
			return &ast.LiteralExpr{Value: ast.FloatLiteral{Value: result}}
		case ast.Mul:
			result = leftFloat.Value * rightFloat.Value
			return &ast.LiteralExpr{Value: ast.FloatLiteral{Value: result}}
		case ast.Div:
			if rightFloat.Value != 0 {
				result = leftFloat.Value / rightFloat.Value
				return &ast.LiteralExpr{Value: ast.FloatLiteral{Value: result}}
			}
		}

		// Comparison operations on floats
		var boolResult bool
		switch op {
		case ast.Eq:
			boolResult = leftFloat.Value == rightFloat.Value
		case ast.Ne:
			boolResult = leftFloat.Value != rightFloat.Value
		case ast.Lt:
			boolResult = leftFloat.Value < rightFloat.Value
		case ast.Le:
			boolResult = leftFloat.Value <= rightFloat.Value
		case ast.Gt:
			boolResult = leftFloat.Value > rightFloat.Value
		case ast.Ge:
			boolResult = leftFloat.Value >= rightFloat.Value
		default:
			goto noFold
		}
		return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: boolResult}}
	}

	// Boolean operations
	if leftIsBool && rightIsBool {
		var result bool
		switch op {
		case ast.And:
			result = leftBool.Value && rightBool.Value
		case ast.Or:
			result = leftBool.Value || rightBool.Value
		case ast.Eq:
			result = leftBool.Value == rightBool.Value
		case ast.Ne:
			result = leftBool.Value != rightBool.Value
		default:
			goto noFold
		}
		return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: result}}
	}

noFold:
	// Cannot fold - return original expression
	return &ast.BinaryOpExpr{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

// algebraicSimplify performs algebraic simplifications
func (o *Optimizer) algebraicSimplify(op ast.BinOp, left, right ast.Expr, leftIsLit bool) ast.Expr {
	var litExpr *ast.LiteralExpr
	var varExpr ast.Expr

	if leftIsLit {
		litExpr, _ = left.(*ast.LiteralExpr)
		varExpr = right
	} else {
		litExpr, _ = right.(*ast.LiteralExpr)
		varExpr = left
	}

	if litExpr == nil {
		return nil
	}

	// Check for numeric literals
	intLit, isInt := litExpr.Value.(ast.IntLiteral)
	floatLit, isFloat := litExpr.Value.(ast.FloatLiteral)
	boolLit, isBool := litExpr.Value.(ast.BoolLiteral)

	isZero := (isInt && intLit.Value == 0) || (isFloat && floatLit.Value == 0)
	isOne := (isInt && intLit.Value == 1) || (isFloat && floatLit.Value == 1)
	isTwo := (isInt && intLit.Value == 2) || (isFloat && floatLit.Value == 2)
	isTrue := isBool && boolLit.Value
	isFalse := isBool && !boolLit.Value

	switch op {
	case ast.Add:
		// x + 0 = x, 0 + x = x
		if isZero {
			return varExpr
		}
	case ast.Sub:
		// x - 0 = x
		if !leftIsLit && isZero {
			return varExpr
		}
	case ast.Mul:
		// x * 0 = 0, 0 * x = 0
		if isZero {
			return &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}}
		}
		// x * 1 = x, 1 * x = x
		if isOne {
			return varExpr
		}
		// Strength reduction: x * 2 = x + x, 2 * x = x + x
		// Addition is typically faster than multiplication
		if isTwo && o.level >= OptAggressive {
			return &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  varExpr,
				Right: varExpr,
			}
		}
	case ast.Div:
		// x / 1 = x
		if !leftIsLit && isOne {
			return varExpr
		}
	case ast.And:
		// true && x = x, x && true = x
		if isTrue {
			return varExpr
		}
		// false && x = false, x && false = false
		if isFalse {
			return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}}
		}
	case ast.Or:
		// true || x = true, x || true = true
		if isTrue {
			return &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}
		}
		// false || x = x, x || false = x
		if isFalse {
			return varExpr
		}
	}

	return nil
}

// exprKey creates a unique key for an expression for CSE
func exprKey(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.LiteralExpr:
		// Literals don't need CSE (they're already constants)
		return ""
	case *ast.VariableExpr:
		// Variables don't need CSE
		return ""
	case *ast.BinaryOpExpr:
		leftKey := exprKey(e.Left)
		rightKey := exprKey(e.Right)

		// Only create key if both operands have keys or are variables/literals
		leftVarOrLit := isVarOrLit(e.Left)
		rightVarOrLit := isVarOrLit(e.Right)

		if (leftKey != "" || leftVarOrLit) && (rightKey != "" || rightVarOrLit) {
			leftStr := leftKey
			if leftStr == "" {
				leftStr = exprToString(e.Left)
			}
			rightStr := rightKey
			if rightStr == "" {
				rightStr = exprToString(e.Right)
			}
			return fmt.Sprintf("(%v %s %s)", e.Op, leftStr, rightStr)
		}
		return ""
	default:
		return ""
	}
}

// isVarOrLit checks if expression is a variable or literal
func isVarOrLit(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.VariableExpr, *ast.LiteralExpr:
		return true
	default:
		return false
	}
}

// exprToString converts simple expressions to strings
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.VariableExpr:
		return fmt.Sprintf("var:%s", e.Name)
	case *ast.LiteralExpr:
		switch lit := e.Value.(type) {
		case ast.IntLiteral:
			return fmt.Sprintf("int:%d", lit.Value)
		case ast.FloatLiteral:
			return fmt.Sprintf("float:%f", lit.Value)
		case ast.BoolLiteral:
			return fmt.Sprintf("bool:%v", lit.Value)
		case ast.StringLiteral:
			return fmt.Sprintf("str:%s", lit.Value)
		}
	}
	return ""
}

// getModifiedVariables returns the set of variables modified by a list of statements
func getModifiedVariables(stmts []ast.Statement) map[string]bool {
	modified := make(map[string]bool)
	for _, stmt := range stmts {
		getModifiedVariablesInStmt(stmt, modified)
	}
	return modified
}

// getModifiedVariablesInStmt adds modified variables from a statement to the map
func getModifiedVariablesInStmt(stmt ast.Statement, modified map[string]bool) {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		modified[s.Target] = true
	case ast.AssignStatement:
		modified[s.Target] = true
	case *ast.ReassignStatement:
		modified[s.Target] = true
	case ast.ReassignStatement:
		modified[s.Target] = true
	case *ast.IfStatement:
		for _, thenStmt := range s.ThenBlock {
			getModifiedVariablesInStmt(thenStmt, modified)
		}
		for _, elseStmt := range s.ElseBlock {
			getModifiedVariablesInStmt(elseStmt, modified)
		}
	case *ast.WhileStatement:
		for _, bodyStmt := range s.Body {
			getModifiedVariablesInStmt(bodyStmt, modified)
		}
	case *ast.ForStatement:
		// Mark loop variables as modified
		modified[s.ValueVar] = true
		if s.KeyVar != "" {
			modified[s.KeyVar] = true
		}
		// Recursively check body for modified variables
		for _, bodyStmt := range s.Body {
			getModifiedVariablesInStmt(bodyStmt, modified)
		}
	case ast.ForStatement:
		// Same as *ast.ForStatement
		modified[s.ValueVar] = true
		if s.KeyVar != "" {
			modified[s.KeyVar] = true
		}
		for _, bodyStmt := range s.Body {
			getModifiedVariablesInStmt(bodyStmt, modified)
		}
	}
}

// getUsedVariables returns the set of variables used in an expression
func getUsedVariables(expr ast.Expr) map[string]bool {
	used := make(map[string]bool)
	getUsedVariablesInExpr(expr, used)
	return used
}

// getUsedVariablesInExpr adds used variables from an expression to the map
func getUsedVariablesInExpr(expr ast.Expr, used map[string]bool) {
	switch e := expr.(type) {
	case *ast.VariableExpr:
		used[e.Name] = true
	case *ast.BinaryOpExpr:
		getUsedVariablesInExpr(e.Left, used)
		getUsedVariablesInExpr(e.Right, used)
	case *ast.ObjectExpr:
		for _, field := range e.Fields {
			getUsedVariablesInExpr(field.Value, used)
		}
	case *ast.ArrayExpr:
		for _, elem := range e.Elements {
			getUsedVariablesInExpr(elem, used)
		}
	case *ast.FieldAccessExpr:
		getUsedVariablesInExpr(e.Object, used)
	}
}

// isExprInvariant checks if an expression is loop-invariant (doesn't depend on modified vars)
func isExprInvariant(expr ast.Expr, modifiedVars map[string]bool) bool {
	usedVars := getUsedVariables(expr)
	for varName := range usedVars {
		if modifiedVars[varName] {
			return false // Expression uses a modified variable
		}
	}
	return true
}

// ========================================
// Advanced Optimization: Peephole Optimizer
// ========================================

// PeepholePattern represents a pattern for peephole optimization
type PeepholePattern struct {
	Name        string
	Match       func([]ast.Statement, int) (bool, int) // returns (matched, length)
	Replace     func([]ast.Statement, int) []ast.Statement
	Description string
}

// peepholePatterns contains all peephole optimization patterns
var peepholePatterns = []PeepholePattern{
	// Pattern: Double negation - !!x -> x
	{
		Name:        "double_negation",
		Description: "Remove double negation",
		Match: func(stmts []ast.Statement, i int) (bool, int) {
			// This would need UnaryOpExpr support in the AST
			return false, 0
		},
		Replace: func(stmts []ast.Statement, i int) []ast.Statement {
			return nil
		},
	},
	// Pattern: Self assignment - x = x -> (remove)
	{
		Name:        "self_assignment",
		Description: "Remove self-assignment",
		Match: func(stmts []ast.Statement, i int) (bool, int) {
			if assign, ok := stmts[i].(*ast.AssignStatement); ok {
				if varExpr, ok := assign.Value.(*ast.VariableExpr); ok {
					if varExpr.Name == assign.Target {
						return true, 1
					}
				}
			}
			return false, 0
		},
		Replace: func(stmts []ast.Statement, i int) []ast.Statement {
			return nil // Remove the statement
		},
	},
	// Pattern: Consecutive assignments to same variable
	{
		Name:        "redundant_assignment",
		Description: "Remove redundant assignments to same variable",
		Match: func(stmts []ast.Statement, i int) (bool, int) {
			if i+1 >= len(stmts) {
				return false, 0
			}
			assign1, ok1 := stmts[i].(*ast.AssignStatement)
			assign2, ok2 := stmts[i+1].(*ast.AssignStatement)
			if ok1 && ok2 && assign1.Target == assign2.Target {
				// Check that first value doesn't have side effects
				if !exprHasSideEffects(assign1.Value) {
					return true, 2
				}
			}
			return false, 0
		},
		Replace: func(stmts []ast.Statement, i int) []ast.Statement {
			// Keep only the second assignment
			return []ast.Statement{stmts[i+1]}
		},
	},
}

// exprHasSideEffects checks if an expression might have side effects
func exprHasSideEffects(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.FunctionCallExpr:
		return true // Function calls may have side effects
	case ast.FunctionCallExpr:
		return true
	case *ast.BinaryOpExpr:
		return exprHasSideEffects(e.Left) || exprHasSideEffects(e.Right)
	case ast.BinaryOpExpr:
		return exprHasSideEffects(e.Left) || exprHasSideEffects(e.Right)
	case *ast.ObjectExpr:
		for _, field := range e.Fields {
			if exprHasSideEffects(field.Value) {
				return true
			}
		}
		return false
	case ast.ObjectExpr:
		for _, field := range e.Fields {
			if exprHasSideEffects(field.Value) {
				return true
			}
		}
		return false
	case *ast.ArrayExpr:
		for _, elem := range e.Elements {
			if exprHasSideEffects(elem) {
				return true
			}
		}
		return false
	case ast.ArrayExpr:
		for _, elem := range e.Elements {
			if exprHasSideEffects(elem) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// ApplyPeepholeOptimizations applies peephole optimizations to statements
func (o *Optimizer) ApplyPeepholeOptimizations(stmts []ast.Statement) []ast.Statement {
	if o.level < OptAggressive {
		return stmts
	}

	result := make([]ast.Statement, 0, len(stmts))
	i := 0

	for i < len(stmts) {
		matched := false

		for _, pattern := range peepholePatterns {
			if match, length := pattern.Match(stmts, i); match {
				replacement := pattern.Replace(stmts, i)
				if replacement != nil {
					result = append(result, replacement...)
				}
				i += length
				matched = true
				break
			}
		}

		if !matched {
			result = append(result, stmts[i])
			i++
		}
	}

	return result
}

// ========================================
// Advanced Optimization: Function Inlining
// ========================================

// InlineCandidate represents a function that can be inlined
type InlineCandidate struct {
	Name        string
	Params      []string
	Body        []ast.Statement
	BodySize    int
	CallCount   int
	IsRecursive bool
}

// FunctionInliner handles function inlining optimization
type FunctionInliner struct {
	candidates map[string]*InlineCandidate
	maxSize    int // Maximum body size for inlining (in statements)
}

// NewFunctionInliner creates a new function inliner
func NewFunctionInliner() *FunctionInliner {
	return &FunctionInliner{
		candidates: make(map[string]*InlineCandidate),
		maxSize:    10, // Default max size
	}
}

// AnalyzeFunction analyzes a function for inlining potential
func (fi *FunctionInliner) AnalyzeFunction(fn ast.Function) {
	// Count body size
	bodySize := countStatements(fn.Body)

	// Check for recursion
	isRecursive := containsCallTo(fn.Body, fn.Name)

	// Extract param names
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = p.Name
	}

	fi.candidates[fn.Name] = &InlineCandidate{
		Name:        fn.Name,
		Params:      params,
		Body:        fn.Body,
		BodySize:    bodySize,
		IsRecursive: isRecursive,
	}
}

// ShouldInline determines if a function should be inlined
func (fi *FunctionInliner) ShouldInline(name string) bool {
	candidate, ok := fi.candidates[name]
	if !ok {
		return false
	}

	// Don't inline recursive functions
	if candidate.IsRecursive {
		return false
	}

	// Don't inline large functions
	if candidate.BodySize > fi.maxSize {
		return false
	}

	return true
}

// InlineCall inlines a function call
func (fi *FunctionInliner) InlineCall(call *ast.FunctionCallExpr) []ast.Statement {
	candidate, ok := fi.candidates[call.Name]
	if !ok {
		return nil
	}

	// Create parameter bindings
	bindings := make(map[string]ast.Expr)
	for i, param := range candidate.Params {
		if i < len(call.Args) {
			bindings[param] = call.Args[i]
		}
	}

	// Substitute parameters in body
	inlinedBody := substituteParams(candidate.Body, bindings)

	return inlinedBody
}

// countStatements counts the number of statements recursively
func countStatements(stmts []ast.Statement) int {
	count := 0
	for _, stmt := range stmts {
		count++
		switch s := stmt.(type) {
		case *ast.IfStatement:
			count += countStatements(s.ThenBlock)
			count += countStatements(s.ElseBlock)
		case ast.IfStatement:
			count += countStatements(s.ThenBlock)
			count += countStatements(s.ElseBlock)
		case *ast.WhileStatement:
			count += countStatements(s.Body)
		case ast.WhileStatement:
			count += countStatements(s.Body)
		case *ast.ForStatement:
			count += countStatements(s.Body)
		case ast.ForStatement:
			count += countStatements(s.Body)
		}
	}
	return count
}

// containsCallTo checks if statements contain a call to a specific function
func containsCallTo(stmts []ast.Statement, fnName string) bool {
	for _, stmt := range stmts {
		if containsCallInStmt(stmt, fnName) {
			return true
		}
	}
	return false
}

// containsCallInStmt checks if a statement contains a call to a specific function
func containsCallInStmt(stmt ast.Statement, fnName string) bool {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		return containsCallInExpr(s.Value, fnName)
	case ast.AssignStatement:
		return containsCallInExpr(s.Value, fnName)
	case *ast.ReassignStatement:
		return containsCallInExpr(s.Value, fnName)
	case ast.ReassignStatement:
		return containsCallInExpr(s.Value, fnName)
	case *ast.ReturnStatement:
		return containsCallInExpr(s.Value, fnName)
	case ast.ReturnStatement:
		return containsCallInExpr(s.Value, fnName)
	case *ast.IfStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.ThenBlock, fnName) || containsCallTo(s.ElseBlock, fnName)
	case ast.IfStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.ThenBlock, fnName) || containsCallTo(s.ElseBlock, fnName)
	case *ast.WhileStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.Body, fnName)
	case ast.WhileStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.Body, fnName)
	case *ast.ExpressionStatement:
		return containsCallInExpr(s.Expr, fnName)
	case ast.ExpressionStatement:
		return containsCallInExpr(s.Expr, fnName)
	}
	return false
}

// containsCallInExpr checks if an expression contains a call to a specific function
func containsCallInExpr(expr ast.Expr, fnName string) bool {
	switch e := expr.(type) {
	case *ast.FunctionCallExpr:
		if e.Name == fnName {
			return true
		}
		for _, arg := range e.Args {
			if containsCallInExpr(arg, fnName) {
				return true
			}
		}
	case ast.FunctionCallExpr:
		if e.Name == fnName {
			return true
		}
		for _, arg := range e.Args {
			if containsCallInExpr(arg, fnName) {
				return true
			}
		}
	case *ast.BinaryOpExpr:
		return containsCallInExpr(e.Left, fnName) || containsCallInExpr(e.Right, fnName)
	case ast.BinaryOpExpr:
		return containsCallInExpr(e.Left, fnName) || containsCallInExpr(e.Right, fnName)
	case *ast.ObjectExpr:
		for _, field := range e.Fields {
			if containsCallInExpr(field.Value, fnName) {
				return true
			}
		}
	case ast.ObjectExpr:
		for _, field := range e.Fields {
			if containsCallInExpr(field.Value, fnName) {
				return true
			}
		}
	case *ast.ArrayExpr:
		for _, elem := range e.Elements {
			if containsCallInExpr(elem, fnName) {
				return true
			}
		}
	case ast.ArrayExpr:
		for _, elem := range e.Elements {
			if containsCallInExpr(elem, fnName) {
				return true
			}
		}
	case *ast.FieldAccessExpr:
		return containsCallInExpr(e.Object, fnName)
	case ast.FieldAccessExpr:
		return containsCallInExpr(e.Object, fnName)
	}
	return false
}

// substituteParams substitutes parameters with their values in statements
func substituteParams(stmts []ast.Statement, bindings map[string]ast.Expr) []ast.Statement {
	result := make([]ast.Statement, len(stmts))
	for i, stmt := range stmts {
		result[i] = substituteParamsInStmt(stmt, bindings)
	}
	return result
}

// substituteParamsInStmt substitutes parameters in a statement
func substituteParamsInStmt(stmt ast.Statement, bindings map[string]ast.Expr) ast.Statement {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		return &ast.AssignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case ast.AssignStatement:
		return &ast.AssignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case *ast.ReassignStatement:
		return &ast.ReassignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case ast.ReassignStatement:
		return &ast.ReassignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case *ast.ReturnStatement:
		return &ast.ReturnStatement{
			Value: substituteParamsInExpr(s.Value, bindings),
		}
	case ast.ReturnStatement:
		return &ast.ReturnStatement{
			Value: substituteParamsInExpr(s.Value, bindings),
		}
	case *ast.IfStatement:
		return &ast.IfStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			ThenBlock: substituteParams(s.ThenBlock, bindings),
			ElseBlock: substituteParams(s.ElseBlock, bindings),
		}
	case ast.IfStatement:
		return &ast.IfStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			ThenBlock: substituteParams(s.ThenBlock, bindings),
			ElseBlock: substituteParams(s.ElseBlock, bindings),
		}
	case *ast.WhileStatement:
		return &ast.WhileStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			Body:      substituteParams(s.Body, bindings),
		}
	case ast.WhileStatement:
		return &ast.WhileStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			Body:      substituteParams(s.Body, bindings),
		}
	default:
		return stmt
	}
}

// substituteParamsInExpr substitutes parameters in an expression
func substituteParamsInExpr(expr ast.Expr, bindings map[string]ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.VariableExpr:
		if replacement, ok := bindings[e.Name]; ok {
			return replacement
		}
		return expr
	case ast.VariableExpr:
		if replacement, ok := bindings[e.Name]; ok {
			return replacement
		}
		return expr
	case *ast.BinaryOpExpr:
		return &ast.BinaryOpExpr{
			Op:    e.Op,
			Left:  substituteParamsInExpr(e.Left, bindings),
			Right: substituteParamsInExpr(e.Right, bindings),
		}
	case ast.BinaryOpExpr:
		return &ast.BinaryOpExpr{
			Op:    e.Op,
			Left:  substituteParamsInExpr(e.Left, bindings),
			Right: substituteParamsInExpr(e.Right, bindings),
		}
	case *ast.ObjectExpr:
		fields := make([]ast.ObjectField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = ast.ObjectField{
				Key:   field.Key,
				Value: substituteParamsInExpr(field.Value, bindings),
			}
		}
		return &ast.ObjectExpr{Fields: fields}
	case ast.ObjectExpr:
		fields := make([]ast.ObjectField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = ast.ObjectField{
				Key:   field.Key,
				Value: substituteParamsInExpr(field.Value, bindings),
			}
		}
		return &ast.ObjectExpr{Fields: fields}
	case *ast.ArrayExpr:
		elements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = substituteParamsInExpr(elem, bindings)
		}
		return &ast.ArrayExpr{Elements: elements}
	case ast.ArrayExpr:
		elements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = substituteParamsInExpr(elem, bindings)
		}
		return &ast.ArrayExpr{Elements: elements}
	case *ast.FieldAccessExpr:
		return &ast.FieldAccessExpr{
			Object: substituteParamsInExpr(e.Object, bindings),
			Field:  e.Field,
		}
	case ast.FieldAccessExpr:
		return &ast.FieldAccessExpr{
			Object: substituteParamsInExpr(e.Object, bindings),
			Field:  e.Field,
		}
	case *ast.FunctionCallExpr:
		args := make([]ast.Expr, len(e.Args))
		for i, arg := range e.Args {
			args[i] = substituteParamsInExpr(arg, bindings)
		}
		return &ast.FunctionCallExpr{Name: e.Name, Args: args}
	case ast.FunctionCallExpr:
		args := make([]ast.Expr, len(e.Args))
		for i, arg := range e.Args {
			args[i] = substituteParamsInExpr(arg, bindings)
		}
		return &ast.FunctionCallExpr{Name: e.Name, Args: args}
	default:
		return expr
	}
}

// ========================================
// Advanced Optimization: Strength Reduction
// ========================================

// StrengthReduce applies strength reduction optimizations
func (o *Optimizer) StrengthReduce(expr ast.Expr) ast.Expr {
	if o.level < OptAggressive {
		return expr
	}

	binOp, ok := expr.(*ast.BinaryOpExpr)
	if !ok {
		return expr
	}

	// Optimize multiplication by powers of 2
	if binOp.Op == ast.Mul {
		if lit, ok := binOp.Right.(*ast.LiteralExpr); ok {
			if intLit, ok := lit.Value.(ast.IntLiteral); ok {
				if isPowerOfTwo(intLit.Value) && intLit.Value > 0 {
					shift := log2(intLit.Value)
					// x * 2^n -> x << n (represented as x + x repeated)
					// For now, convert x * 2 -> x + x, x * 4 -> (x + x) + (x + x)
					if shift == 1 {
						return &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  binOp.Left,
							Right: binOp.Left,
						}
					}
				}
			}
		}
	}

	// Optimize division by powers of 2
	if binOp.Op == ast.Div {
		if lit, ok := binOp.Right.(*ast.LiteralExpr); ok {
			if intLit, ok := lit.Value.(ast.IntLiteral); ok {
				if isPowerOfTwo(intLit.Value) && intLit.Value > 0 {
					// x / 2^n -> x >> n (right shift)
					// Note: This is only valid for unsigned integers
					// For now, we skip this optimization
				}
			}
		}
	}

	return expr
}

// isPowerOfTwo checks if n is a power of 2
func isPowerOfTwo(n int64) bool {
	return n > 0 && (n&(n-1)) == 0
}

// log2 returns the log base 2 of n (assumes n is a power of 2)
func log2(n int64) int {
	count := 0
	for n > 1 {
		n >>= 1
		count++
	}
	return count
}

// ========================================
// Optimization Statistics
// ========================================

// OptimizationStats tracks statistics about optimizations performed
type OptimizationStats struct {
	ConstantsFolded       int
	DeadCodeEliminated    int
	CopiesPropagated      int
	ExpressionsEliminated int
	LoopInvariants        int
	FunctionsInlined      int
	StrengthReductions    int
}

// GetStats returns optimization statistics
func (o *Optimizer) GetStats() OptimizationStats {
	return OptimizationStats{
		// These would be tracked during optimization
	}
}
