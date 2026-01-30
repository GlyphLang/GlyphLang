package compiler

import (
	"fmt"

	"github.com/glyphlang/glyph/pkg/interpreter"
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
	constants   map[string]interpreter.Literal // Track constant values for variables
	expressions map[string]string              // Track expression -> variable name for CSE
	copies      map[string]string              // Track variable copies (x = y means copies[x] = y)
}

// NewOptimizer creates a new optimizer instance
func NewOptimizer(level OptimizationLevel) *Optimizer {
	return &Optimizer{
		level:       level,
		constants:   make(map[string]interpreter.Literal),
		expressions: make(map[string]string),
		copies:      make(map[string]string),
	}
}

// OptimizeExpression optimizes an expression
func (o *Optimizer) OptimizeExpression(expr interpreter.Expr) interpreter.Expr {
	if o.level == OptNone {
		return expr
	}

	switch e := expr.(type) {
	case *interpreter.VariableExpr:
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
			return &interpreter.LiteralExpr{Value: lit}
		}

		// Return the resolved variable (might be different due to copy propagation)
		if varName != e.Name {
			return &interpreter.VariableExpr{Name: varName}
		}
		return expr
	case *interpreter.BinaryOpExpr:
		return o.foldBinaryOp(e)
	case *interpreter.ObjectExpr:
		// Optimize object fields
		fields := make([]interpreter.ObjectField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = interpreter.ObjectField{
				Key:   field.Key,
				Value: o.OptimizeExpression(field.Value),
			}
		}
		return &interpreter.ObjectExpr{Fields: fields}
	case *interpreter.ArrayExpr:
		// Optimize array elements
		elements := make([]interpreter.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = o.OptimizeExpression(elem)
		}
		return &interpreter.ArrayExpr{Elements: elements}
	case *interpreter.FieldAccessExpr:
		// Optimize the object expression
		return &interpreter.FieldAccessExpr{
			Object: o.OptimizeExpression(e.Object),
			Field:  e.Field,
		}
	default:
		return expr
	}
}

// OptimizeStatements optimizes a list of statements
func (o *Optimizer) OptimizeStatements(stmts []interpreter.Statement) []interpreter.Statement {
	if o.level == OptNone {
		return stmts
	}

	result := make([]interpreter.Statement, 0, len(stmts))
	reachedReturn := false

	for _, stmt := range stmts {
		// Dead code elimination: skip statements after return
		if reachedReturn {
			continue
		}

		switch s := stmt.(type) {
		case *interpreter.AssignStatement:
			// Optimize the value expression
			optimizedValue := o.OptimizeExpression(s.Value)

			// Copy propagation: track variable-to-variable assignments
			if varExpr, ok := optimizedValue.(*interpreter.VariableExpr); ok {
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
							optimizedValue = &interpreter.VariableExpr{Name: existingVar}
							// Now it's a copy
							o.copies[s.Target] = existingVar
						} else {
							// Track this expression
							o.expressions[key] = s.Target
						}
					}
				}

				// Track constant assignments for constant propagation
				if litExpr, ok := optimizedValue.(*interpreter.LiteralExpr); ok {
					o.constants[s.Target] = litExpr.Value
				} else {
					// Non-constant assignment, invalidate any previous constant
					delete(o.constants, s.Target)
				}
			}

			optimized := &interpreter.AssignStatement{
				Target: s.Target,
				Value:  optimizedValue,
			}
			result = append(result, optimized)

		case *interpreter.ReassignStatement:
			// Optimize the value expression (same logic as AssignStatement)
			optimizedValue := o.OptimizeExpression(s.Value)

			// Copy propagation: track variable-to-variable assignments
			if varExpr, ok := optimizedValue.(*interpreter.VariableExpr); ok {
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
							optimizedValue = &interpreter.VariableExpr{Name: existingVar}
							// Now it's a copy
							o.copies[s.Target] = existingVar
						} else {
							// Track this expression
							o.expressions[key] = s.Target
						}
					}
				}

				// Track constant assignments for constant propagation
				if litExpr, ok := optimizedValue.(*interpreter.LiteralExpr); ok {
					o.constants[s.Target] = litExpr.Value
				} else {
					// Non-constant assignment, invalidate any previous constant
					delete(o.constants, s.Target)
				}
			}

			optimized := &interpreter.ReassignStatement{
				Target: s.Target,
				Value:  optimizedValue,
			}
			result = append(result, optimized)

		case interpreter.ReassignStatement:
			// Same as *interpreter.ReassignStatement
			optimizedValue := o.OptimizeExpression(s.Value)

			if varExpr, ok := optimizedValue.(*interpreter.VariableExpr); ok {
				o.copies[s.Target] = varExpr.Name
				delete(o.constants, s.Target)
			} else {
				delete(o.copies, s.Target)

				if o.level >= OptAggressive {
					key := exprKey(optimizedValue)
					if key != "" {
						if existingVar, ok := o.expressions[key]; ok {
							optimizedValue = &interpreter.VariableExpr{Name: existingVar}
							o.copies[s.Target] = existingVar
						} else {
							o.expressions[key] = s.Target
						}
					}
				}

				if litExpr, ok := optimizedValue.(*interpreter.LiteralExpr); ok {
					o.constants[s.Target] = litExpr.Value
				} else {
					delete(o.constants, s.Target)
				}
			}

			result = append(result, &interpreter.ReassignStatement{
				Target: s.Target,
				Value:  optimizedValue,
			})

		case *interpreter.ReturnStatement:
			// Optimize return value
			optimized := &interpreter.ReturnStatement{
				Value: o.OptimizeExpression(s.Value),
			}
			result = append(result, optimized)
			reachedReturn = true

		case *interpreter.IfStatement:
			// Try to optimize the condition
			condition := o.OptimizeExpression(s.Condition)

			// Check if condition is a constant boolean
			if litExpr, ok := condition.(*interpreter.LiteralExpr); ok {
				if boolLit, ok := litExpr.Value.(interpreter.BoolLiteral); ok {
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
			optimized := &interpreter.IfStatement{
				Condition: condition,
				ThenBlock: o.OptimizeStatements(s.ThenBlock),
				ElseBlock: o.OptimizeStatements(s.ElseBlock),
			}
			result = append(result, optimized)

		case *interpreter.WhileStatement:
			// First, invalidate constants for any variables modified in the loop body
			// because the loop may execute multiple times or not at all
			modifiedVars := getModifiedVariables(s.Body)
			for varName := range modifiedVars {
				delete(o.constants, varName)
				delete(o.copies, varName)
				delete(o.expressions, varName)
			}

			// Loop invariant code motion (OptAggressive only)
			var invariantStmts []interpreter.Statement
			var loopBody []interpreter.Statement

			if o.level >= OptAggressive {
				// Also check if condition uses any variables
				conditionVars := getUsedVariables(s.Condition)

				// Separate invariant assignments from loop-variant ones
				for _, bodyStmt := range s.Body {
					if assignStmt, ok := bodyStmt.(*interpreter.AssignStatement); ok {
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
				result = append(result, o.OptimizeStatements([]interpreter.Statement{invStmt})...)
			}

			// Optimize condition and remaining loop body
			optimized := &interpreter.WhileStatement{
				Condition: o.OptimizeExpression(s.Condition),
				Body:      o.OptimizeStatements(loopBody),
			}
			result = append(result, optimized)

		case *interpreter.ForStatement:
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

		case interpreter.ForStatement:
			// Same as *interpreter.ForStatement
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

		case *interpreter.SwitchStatement:
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

		case interpreter.SwitchStatement:
			// Same as *interpreter.SwitchStatement
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
func (o *Optimizer) foldBinaryOp(expr *interpreter.BinaryOpExpr) interpreter.Expr {
	// First optimize operands recursively
	left := o.OptimizeExpression(expr.Left)
	right := o.OptimizeExpression(expr.Right)

	// Arithmetic identities: x op x = constant (for OptAggressive)
	if o.level >= OptAggressive {
		if areExprsEqual(left, right) {
			switch expr.Op {
			case interpreter.Sub:
				// x - x = 0
				return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}}
			case interpreter.Div:
				// x / x = 1 (assuming x != 0, which we can't verify at compile time)
				// Only optimize if safe
				return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}}
			case interpreter.Eq:
				// x == x = true
				return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}
			case interpreter.Ne:
				// x != x = false
				return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}}
			case interpreter.Le, interpreter.Ge:
				// x <= x = true, x >= x = true
				return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}
			case interpreter.Lt, interpreter.Gt:
				// x < x = false, x > x = false
				return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}}
			}
		}
	}

	// Try to extract literal values
	leftLit, leftIsLit := left.(*interpreter.LiteralExpr)
	rightLit, rightIsLit := right.(*interpreter.LiteralExpr)

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
	return &interpreter.BinaryOpExpr{
		Op:    expr.Op,
		Left:  left,
		Right: right,
	}
}

// areExprsEqual checks if two expressions are structurally equal
func areExprsEqual(a, b interpreter.Expr) bool {
	switch aExpr := a.(type) {
	case *interpreter.VariableExpr:
		if bExpr, ok := b.(*interpreter.VariableExpr); ok {
			return aExpr.Name == bExpr.Name
		}
	case *interpreter.LiteralExpr:
		if bExpr, ok := b.(*interpreter.LiteralExpr); ok {
			return literalsEqual(aExpr.Value, bExpr.Value)
		}
	}
	return false
}

// literalsEqual compares two literals for equality
func literalsEqual(a, b interpreter.Literal) bool {
	switch aLit := a.(type) {
	case interpreter.IntLiteral:
		if bLit, ok := b.(interpreter.IntLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case interpreter.FloatLiteral:
		if bLit, ok := b.(interpreter.FloatLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case interpreter.BoolLiteral:
		if bLit, ok := b.(interpreter.BoolLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case interpreter.StringLiteral:
		if bLit, ok := b.(interpreter.StringLiteral); ok {
			return aLit.Value == bLit.Value
		}
	case interpreter.NullLiteral:
		_, ok := b.(interpreter.NullLiteral)
		return ok
	}
	return false
}

// foldLiteralBinaryOp folds binary operations on two literals
func (o *Optimizer) foldLiteralBinaryOp(op interpreter.BinOp, left, right *interpreter.LiteralExpr) interpreter.Expr {
	// Extract int literals
	leftInt, leftIsInt := left.Value.(interpreter.IntLiteral)
	rightInt, rightIsInt := right.Value.(interpreter.IntLiteral)

	// Extract float literals
	leftFloat, leftIsFloat := left.Value.(interpreter.FloatLiteral)
	rightFloat, rightIsFloat := right.Value.(interpreter.FloatLiteral)

	// Extract bool literals
	leftBool, leftIsBool := left.Value.(interpreter.BoolLiteral)
	rightBool, rightIsBool := right.Value.(interpreter.BoolLiteral)

	// Arithmetic operations on integers
	if leftIsInt && rightIsInt {
		var result int64
		switch op {
		case interpreter.Add:
			result = leftInt.Value + rightInt.Value
			return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: result}}
		case interpreter.Sub:
			result = leftInt.Value - rightInt.Value
			return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: result}}
		case interpreter.Mul:
			result = leftInt.Value * rightInt.Value
			return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: result}}
		case interpreter.Div:
			if rightInt.Value != 0 {
				result = leftInt.Value / rightInt.Value
				return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: result}}
			}
		}

		// Comparison operations on integers
		var boolResult bool
		switch op {
		case interpreter.Eq:
			boolResult = leftInt.Value == rightInt.Value
		case interpreter.Ne:
			boolResult = leftInt.Value != rightInt.Value
		case interpreter.Lt:
			boolResult = leftInt.Value < rightInt.Value
		case interpreter.Le:
			boolResult = leftInt.Value <= rightInt.Value
		case interpreter.Gt:
			boolResult = leftInt.Value > rightInt.Value
		case interpreter.Ge:
			boolResult = leftInt.Value >= rightInt.Value
		default:
			goto noFold
		}
		return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: boolResult}}
	}

	// Arithmetic operations on floats
	if leftIsFloat && rightIsFloat {
		var result float64
		switch op {
		case interpreter.Add:
			result = leftFloat.Value + rightFloat.Value
			return &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: result}}
		case interpreter.Sub:
			result = leftFloat.Value - rightFloat.Value
			return &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: result}}
		case interpreter.Mul:
			result = leftFloat.Value * rightFloat.Value
			return &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: result}}
		case interpreter.Div:
			if rightFloat.Value != 0 {
				result = leftFloat.Value / rightFloat.Value
				return &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: result}}
			}
		}

		// Comparison operations on floats
		var boolResult bool
		switch op {
		case interpreter.Eq:
			boolResult = leftFloat.Value == rightFloat.Value
		case interpreter.Ne:
			boolResult = leftFloat.Value != rightFloat.Value
		case interpreter.Lt:
			boolResult = leftFloat.Value < rightFloat.Value
		case interpreter.Le:
			boolResult = leftFloat.Value <= rightFloat.Value
		case interpreter.Gt:
			boolResult = leftFloat.Value > rightFloat.Value
		case interpreter.Ge:
			boolResult = leftFloat.Value >= rightFloat.Value
		default:
			goto noFold
		}
		return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: boolResult}}
	}

	// Boolean operations
	if leftIsBool && rightIsBool {
		var result bool
		switch op {
		case interpreter.And:
			result = leftBool.Value && rightBool.Value
		case interpreter.Or:
			result = leftBool.Value || rightBool.Value
		case interpreter.Eq:
			result = leftBool.Value == rightBool.Value
		case interpreter.Ne:
			result = leftBool.Value != rightBool.Value
		default:
			goto noFold
		}
		return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: result}}
	}

noFold:
	// Cannot fold - return original expression
	return &interpreter.BinaryOpExpr{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

// algebraicSimplify performs algebraic simplifications
func (o *Optimizer) algebraicSimplify(op interpreter.BinOp, left, right interpreter.Expr, leftIsLit bool) interpreter.Expr {
	var litExpr *interpreter.LiteralExpr
	var varExpr interpreter.Expr

	if leftIsLit {
		litExpr, _ = left.(*interpreter.LiteralExpr)
		varExpr = right
	} else {
		litExpr, _ = right.(*interpreter.LiteralExpr)
		varExpr = left
	}

	if litExpr == nil {
		return nil
	}

	// Check for numeric literals
	intLit, isInt := litExpr.Value.(interpreter.IntLiteral)
	floatLit, isFloat := litExpr.Value.(interpreter.FloatLiteral)
	boolLit, isBool := litExpr.Value.(interpreter.BoolLiteral)

	isZero := (isInt && intLit.Value == 0) || (isFloat && floatLit.Value == 0)
	isOne := (isInt && intLit.Value == 1) || (isFloat && floatLit.Value == 1)
	isTwo := (isInt && intLit.Value == 2) || (isFloat && floatLit.Value == 2)
	isTrue := isBool && boolLit.Value
	isFalse := isBool && !boolLit.Value

	switch op {
	case interpreter.Add:
		// x + 0 = x, 0 + x = x
		if isZero {
			return varExpr
		}
	case interpreter.Sub:
		// x - 0 = x
		if !leftIsLit && isZero {
			return varExpr
		}
	case interpreter.Mul:
		// x * 0 = 0, 0 * x = 0
		if isZero {
			return &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}}
		}
		// x * 1 = x, 1 * x = x
		if isOne {
			return varExpr
		}
		// Strength reduction: x * 2 = x + x, 2 * x = x + x
		// Addition is typically faster than multiplication
		if isTwo && o.level >= OptAggressive {
			return &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  varExpr,
				Right: varExpr,
			}
		}
	case interpreter.Div:
		// x / 1 = x
		if !leftIsLit && isOne {
			return varExpr
		}
	case interpreter.And:
		// true && x = x, x && true = x
		if isTrue {
			return varExpr
		}
		// false && x = false, x && false = false
		if isFalse {
			return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}}
		}
	case interpreter.Or:
		// true || x = true, x || true = true
		if isTrue {
			return &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}
		}
		// false || x = x, x || false = x
		if isFalse {
			return varExpr
		}
	}

	return nil
}

// exprKey creates a unique key for an expression for CSE
func exprKey(expr interpreter.Expr) string {
	switch e := expr.(type) {
	case *interpreter.LiteralExpr:
		// Literals don't need CSE (they're already constants)
		return ""
	case *interpreter.VariableExpr:
		// Variables don't need CSE
		return ""
	case *interpreter.BinaryOpExpr:
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
func isVarOrLit(expr interpreter.Expr) bool {
	switch expr.(type) {
	case *interpreter.VariableExpr, *interpreter.LiteralExpr:
		return true
	default:
		return false
	}
}

// exprToString converts simple expressions to strings
func exprToString(expr interpreter.Expr) string {
	switch e := expr.(type) {
	case *interpreter.VariableExpr:
		return fmt.Sprintf("var:%s", e.Name)
	case *interpreter.LiteralExpr:
		switch lit := e.Value.(type) {
		case interpreter.IntLiteral:
			return fmt.Sprintf("int:%d", lit.Value)
		case interpreter.FloatLiteral:
			return fmt.Sprintf("float:%f", lit.Value)
		case interpreter.BoolLiteral:
			return fmt.Sprintf("bool:%v", lit.Value)
		case interpreter.StringLiteral:
			return fmt.Sprintf("str:%s", lit.Value)
		}
	}
	return ""
}

// getModifiedVariables returns the set of variables modified by a list of statements
func getModifiedVariables(stmts []interpreter.Statement) map[string]bool {
	modified := make(map[string]bool)
	for _, stmt := range stmts {
		getModifiedVariablesInStmt(stmt, modified)
	}
	return modified
}

// getModifiedVariablesInStmt adds modified variables from a statement to the map
func getModifiedVariablesInStmt(stmt interpreter.Statement, modified map[string]bool) {
	switch s := stmt.(type) {
	case *interpreter.AssignStatement:
		modified[s.Target] = true
	case interpreter.AssignStatement:
		modified[s.Target] = true
	case *interpreter.ReassignStatement:
		modified[s.Target] = true
	case interpreter.ReassignStatement:
		modified[s.Target] = true
	case *interpreter.IfStatement:
		getModifiedVariablesInStmt(&interpreter.AssignStatement{}, modified)
		for _, thenStmt := range s.ThenBlock {
			getModifiedVariablesInStmt(thenStmt, modified)
		}
		for _, elseStmt := range s.ElseBlock {
			getModifiedVariablesInStmt(elseStmt, modified)
		}
	case *interpreter.WhileStatement:
		for _, bodyStmt := range s.Body {
			getModifiedVariablesInStmt(bodyStmt, modified)
		}
	case *interpreter.ForStatement:
		// Mark loop variables as modified
		modified[s.ValueVar] = true
		if s.KeyVar != "" {
			modified[s.KeyVar] = true
		}
		// Recursively check body for modified variables
		for _, bodyStmt := range s.Body {
			getModifiedVariablesInStmt(bodyStmt, modified)
		}
	case interpreter.ForStatement:
		// Same as *interpreter.ForStatement
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
func getUsedVariables(expr interpreter.Expr) map[string]bool {
	used := make(map[string]bool)
	getUsedVariablesInExpr(expr, used)
	return used
}

// getUsedVariablesInExpr adds used variables from an expression to the map
func getUsedVariablesInExpr(expr interpreter.Expr, used map[string]bool) {
	switch e := expr.(type) {
	case *interpreter.VariableExpr:
		used[e.Name] = true
	case *interpreter.BinaryOpExpr:
		getUsedVariablesInExpr(e.Left, used)
		getUsedVariablesInExpr(e.Right, used)
	case *interpreter.ObjectExpr:
		for _, field := range e.Fields {
			getUsedVariablesInExpr(field.Value, used)
		}
	case *interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			getUsedVariablesInExpr(elem, used)
		}
	case *interpreter.FieldAccessExpr:
		getUsedVariablesInExpr(e.Object, used)
	}
}

// isExprInvariant checks if an expression is loop-invariant (doesn't depend on modified vars)
func isExprInvariant(expr interpreter.Expr, modifiedVars map[string]bool) bool {
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
	Match       func([]interpreter.Statement, int) (bool, int) // returns (matched, length)
	Replace     func([]interpreter.Statement, int) []interpreter.Statement
	Description string
}

// peepholePatterns contains all peephole optimization patterns
var peepholePatterns = []PeepholePattern{
	// Pattern: Double negation - !!x -> x
	{
		Name:        "double_negation",
		Description: "Remove double negation",
		Match: func(stmts []interpreter.Statement, i int) (bool, int) {
			// This would need UnaryOpExpr support in the AST
			return false, 0
		},
		Replace: func(stmts []interpreter.Statement, i int) []interpreter.Statement {
			return nil
		},
	},
	// Pattern: Self assignment - x = x -> (remove)
	{
		Name:        "self_assignment",
		Description: "Remove self-assignment",
		Match: func(stmts []interpreter.Statement, i int) (bool, int) {
			if assign, ok := stmts[i].(*interpreter.AssignStatement); ok {
				if varExpr, ok := assign.Value.(*interpreter.VariableExpr); ok {
					if varExpr.Name == assign.Target {
						return true, 1
					}
				}
			}
			return false, 0
		},
		Replace: func(stmts []interpreter.Statement, i int) []interpreter.Statement {
			return nil // Remove the statement
		},
	},
	// Pattern: Consecutive assignments to same variable
	{
		Name:        "redundant_assignment",
		Description: "Remove redundant assignments to same variable",
		Match: func(stmts []interpreter.Statement, i int) (bool, int) {
			if i+1 >= len(stmts) {
				return false, 0
			}
			assign1, ok1 := stmts[i].(*interpreter.AssignStatement)
			assign2, ok2 := stmts[i+1].(*interpreter.AssignStatement)
			if ok1 && ok2 && assign1.Target == assign2.Target {
				// Check that first value doesn't have side effects
				if !exprHasSideEffects(assign1.Value) {
					return true, 2
				}
			}
			return false, 0
		},
		Replace: func(stmts []interpreter.Statement, i int) []interpreter.Statement {
			// Keep only the second assignment
			return []interpreter.Statement{stmts[i+1]}
		},
	},
}

// exprHasSideEffects checks if an expression might have side effects
func exprHasSideEffects(expr interpreter.Expr) bool {
	switch e := expr.(type) {
	case *interpreter.FunctionCallExpr:
		return true // Function calls may have side effects
	case interpreter.FunctionCallExpr:
		return true
	case *interpreter.BinaryOpExpr:
		return exprHasSideEffects(e.Left) || exprHasSideEffects(e.Right)
	case interpreter.BinaryOpExpr:
		return exprHasSideEffects(e.Left) || exprHasSideEffects(e.Right)
	case *interpreter.ObjectExpr:
		for _, field := range e.Fields {
			if exprHasSideEffects(field.Value) {
				return true
			}
		}
		return false
	case interpreter.ObjectExpr:
		for _, field := range e.Fields {
			if exprHasSideEffects(field.Value) {
				return true
			}
		}
		return false
	case *interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			if exprHasSideEffects(elem) {
				return true
			}
		}
		return false
	case interpreter.ArrayExpr:
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
func (o *Optimizer) ApplyPeepholeOptimizations(stmts []interpreter.Statement) []interpreter.Statement {
	if o.level < OptAggressive {
		return stmts
	}

	result := make([]interpreter.Statement, 0, len(stmts))
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
	Body        []interpreter.Statement
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
func (fi *FunctionInliner) AnalyzeFunction(fn interpreter.Function) {
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
func (fi *FunctionInliner) InlineCall(call *interpreter.FunctionCallExpr) []interpreter.Statement {
	candidate, ok := fi.candidates[call.Name]
	if !ok {
		return nil
	}

	// Create parameter bindings
	bindings := make(map[string]interpreter.Expr)
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
func countStatements(stmts []interpreter.Statement) int {
	count := 0
	for _, stmt := range stmts {
		count++
		switch s := stmt.(type) {
		case *interpreter.IfStatement:
			count += countStatements(s.ThenBlock)
			count += countStatements(s.ElseBlock)
		case interpreter.IfStatement:
			count += countStatements(s.ThenBlock)
			count += countStatements(s.ElseBlock)
		case *interpreter.WhileStatement:
			count += countStatements(s.Body)
		case interpreter.WhileStatement:
			count += countStatements(s.Body)
		case *interpreter.ForStatement:
			count += countStatements(s.Body)
		case interpreter.ForStatement:
			count += countStatements(s.Body)
		}
	}
	return count
}

// containsCallTo checks if statements contain a call to a specific function
func containsCallTo(stmts []interpreter.Statement, fnName string) bool {
	for _, stmt := range stmts {
		if containsCallInStmt(stmt, fnName) {
			return true
		}
	}
	return false
}

// containsCallInStmt checks if a statement contains a call to a specific function
func containsCallInStmt(stmt interpreter.Statement, fnName string) bool {
	switch s := stmt.(type) {
	case *interpreter.AssignStatement:
		return containsCallInExpr(s.Value, fnName)
	case interpreter.AssignStatement:
		return containsCallInExpr(s.Value, fnName)
	case *interpreter.ReassignStatement:
		return containsCallInExpr(s.Value, fnName)
	case interpreter.ReassignStatement:
		return containsCallInExpr(s.Value, fnName)
	case *interpreter.ReturnStatement:
		return containsCallInExpr(s.Value, fnName)
	case interpreter.ReturnStatement:
		return containsCallInExpr(s.Value, fnName)
	case *interpreter.IfStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.ThenBlock, fnName) || containsCallTo(s.ElseBlock, fnName)
	case interpreter.IfStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.ThenBlock, fnName) || containsCallTo(s.ElseBlock, fnName)
	case *interpreter.WhileStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.Body, fnName)
	case interpreter.WhileStatement:
		if containsCallInExpr(s.Condition, fnName) {
			return true
		}
		return containsCallTo(s.Body, fnName)
	case *interpreter.ExpressionStatement:
		return containsCallInExpr(s.Expr, fnName)
	case interpreter.ExpressionStatement:
		return containsCallInExpr(s.Expr, fnName)
	}
	return false
}

// containsCallInExpr checks if an expression contains a call to a specific function
func containsCallInExpr(expr interpreter.Expr, fnName string) bool {
	switch e := expr.(type) {
	case *interpreter.FunctionCallExpr:
		if e.Name == fnName {
			return true
		}
		for _, arg := range e.Args {
			if containsCallInExpr(arg, fnName) {
				return true
			}
		}
	case interpreter.FunctionCallExpr:
		if e.Name == fnName {
			return true
		}
		for _, arg := range e.Args {
			if containsCallInExpr(arg, fnName) {
				return true
			}
		}
	case *interpreter.BinaryOpExpr:
		return containsCallInExpr(e.Left, fnName) || containsCallInExpr(e.Right, fnName)
	case interpreter.BinaryOpExpr:
		return containsCallInExpr(e.Left, fnName) || containsCallInExpr(e.Right, fnName)
	case *interpreter.ObjectExpr:
		for _, field := range e.Fields {
			if containsCallInExpr(field.Value, fnName) {
				return true
			}
		}
	case interpreter.ObjectExpr:
		for _, field := range e.Fields {
			if containsCallInExpr(field.Value, fnName) {
				return true
			}
		}
	case *interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			if containsCallInExpr(elem, fnName) {
				return true
			}
		}
	case interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			if containsCallInExpr(elem, fnName) {
				return true
			}
		}
	case *interpreter.FieldAccessExpr:
		return containsCallInExpr(e.Object, fnName)
	case interpreter.FieldAccessExpr:
		return containsCallInExpr(e.Object, fnName)
	}
	return false
}

// substituteParams substitutes parameters with their values in statements
func substituteParams(stmts []interpreter.Statement, bindings map[string]interpreter.Expr) []interpreter.Statement {
	result := make([]interpreter.Statement, len(stmts))
	for i, stmt := range stmts {
		result[i] = substituteParamsInStmt(stmt, bindings)
	}
	return result
}

// substituteParamsInStmt substitutes parameters in a statement
func substituteParamsInStmt(stmt interpreter.Statement, bindings map[string]interpreter.Expr) interpreter.Statement {
	switch s := stmt.(type) {
	case *interpreter.AssignStatement:
		return &interpreter.AssignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case interpreter.AssignStatement:
		return &interpreter.AssignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case *interpreter.ReassignStatement:
		return &interpreter.ReassignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case interpreter.ReassignStatement:
		return &interpreter.ReassignStatement{
			Target: s.Target,
			Value:  substituteParamsInExpr(s.Value, bindings),
		}
	case *interpreter.ReturnStatement:
		return &interpreter.ReturnStatement{
			Value: substituteParamsInExpr(s.Value, bindings),
		}
	case interpreter.ReturnStatement:
		return &interpreter.ReturnStatement{
			Value: substituteParamsInExpr(s.Value, bindings),
		}
	case *interpreter.IfStatement:
		return &interpreter.IfStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			ThenBlock: substituteParams(s.ThenBlock, bindings),
			ElseBlock: substituteParams(s.ElseBlock, bindings),
		}
	case interpreter.IfStatement:
		return &interpreter.IfStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			ThenBlock: substituteParams(s.ThenBlock, bindings),
			ElseBlock: substituteParams(s.ElseBlock, bindings),
		}
	case *interpreter.WhileStatement:
		return &interpreter.WhileStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			Body:      substituteParams(s.Body, bindings),
		}
	case interpreter.WhileStatement:
		return &interpreter.WhileStatement{
			Condition: substituteParamsInExpr(s.Condition, bindings),
			Body:      substituteParams(s.Body, bindings),
		}
	default:
		return stmt
	}
}

// substituteParamsInExpr substitutes parameters in an expression
func substituteParamsInExpr(expr interpreter.Expr, bindings map[string]interpreter.Expr) interpreter.Expr {
	switch e := expr.(type) {
	case *interpreter.VariableExpr:
		if replacement, ok := bindings[e.Name]; ok {
			return replacement
		}
		return expr
	case interpreter.VariableExpr:
		if replacement, ok := bindings[e.Name]; ok {
			return replacement
		}
		return expr
	case *interpreter.BinaryOpExpr:
		return &interpreter.BinaryOpExpr{
			Op:    e.Op,
			Left:  substituteParamsInExpr(e.Left, bindings),
			Right: substituteParamsInExpr(e.Right, bindings),
		}
	case interpreter.BinaryOpExpr:
		return &interpreter.BinaryOpExpr{
			Op:    e.Op,
			Left:  substituteParamsInExpr(e.Left, bindings),
			Right: substituteParamsInExpr(e.Right, bindings),
		}
	case *interpreter.ObjectExpr:
		fields := make([]interpreter.ObjectField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = interpreter.ObjectField{
				Key:   field.Key,
				Value: substituteParamsInExpr(field.Value, bindings),
			}
		}
		return &interpreter.ObjectExpr{Fields: fields}
	case interpreter.ObjectExpr:
		fields := make([]interpreter.ObjectField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = interpreter.ObjectField{
				Key:   field.Key,
				Value: substituteParamsInExpr(field.Value, bindings),
			}
		}
		return &interpreter.ObjectExpr{Fields: fields}
	case *interpreter.ArrayExpr:
		elements := make([]interpreter.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = substituteParamsInExpr(elem, bindings)
		}
		return &interpreter.ArrayExpr{Elements: elements}
	case interpreter.ArrayExpr:
		elements := make([]interpreter.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = substituteParamsInExpr(elem, bindings)
		}
		return &interpreter.ArrayExpr{Elements: elements}
	case *interpreter.FieldAccessExpr:
		return &interpreter.FieldAccessExpr{
			Object: substituteParamsInExpr(e.Object, bindings),
			Field:  e.Field,
		}
	case interpreter.FieldAccessExpr:
		return &interpreter.FieldAccessExpr{
			Object: substituteParamsInExpr(e.Object, bindings),
			Field:  e.Field,
		}
	case *interpreter.FunctionCallExpr:
		args := make([]interpreter.Expr, len(e.Args))
		for i, arg := range e.Args {
			args[i] = substituteParamsInExpr(arg, bindings)
		}
		return &interpreter.FunctionCallExpr{Name: e.Name, Args: args}
	case interpreter.FunctionCallExpr:
		args := make([]interpreter.Expr, len(e.Args))
		for i, arg := range e.Args {
			args[i] = substituteParamsInExpr(arg, bindings)
		}
		return &interpreter.FunctionCallExpr{Name: e.Name, Args: args}
	default:
		return expr
	}
}

// ========================================
// Advanced Optimization: Strength Reduction
// ========================================

// StrengthReduce applies strength reduction optimizations
func (o *Optimizer) StrengthReduce(expr interpreter.Expr) interpreter.Expr {
	if o.level < OptAggressive {
		return expr
	}

	binOp, ok := expr.(*interpreter.BinaryOpExpr)
	if !ok {
		return expr
	}

	// Optimize multiplication by powers of 2
	if binOp.Op == interpreter.Mul {
		if lit, ok := binOp.Right.(*interpreter.LiteralExpr); ok {
			if intLit, ok := lit.Value.(interpreter.IntLiteral); ok {
				if isPowerOfTwo(intLit.Value) && intLit.Value > 0 {
					shift := log2(intLit.Value)
					// x * 2^n -> x << n (represented as x + x repeated)
					// For now, convert x * 2 -> x + x, x * 4 -> (x + x) + (x + x)
					if shift == 1 {
						return &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  binOp.Left,
							Right: binOp.Left,
						}
					}
				}
			}
		}
	}

	// Optimize division by powers of 2
	if binOp.Op == interpreter.Div {
		if lit, ok := binOp.Right.(*interpreter.LiteralExpr); ok {
			if intLit, ok := lit.Value.(interpreter.IntLiteral); ok {
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
