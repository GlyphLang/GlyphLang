package interpreter

import (
	"fmt"
	"strings"
)

// EvaluateExpression evaluates an expression and returns its value
func (i *Interpreter) EvaluateExpression(expr Expr, env *Environment) (interface{}, error) {
	switch e := expr.(type) {
	case LiteralExpr:
		return i.evaluateLiteral(e.Value)

	case VariableExpr:
		return env.Get(e.Name)

	case BinaryOpExpr:
		return i.evaluateBinaryOp(e, env)

	case UnaryOpExpr:
		return i.evaluateUnaryOp(e, env)

	case FieldAccessExpr:
		return i.evaluateFieldAccess(e, env)

	case FunctionCallExpr:
		return i.evaluateFunctionCall(e, env)

	case ObjectExpr:
		return i.evaluateObjectExpr(e, env)

	case ArrayExpr:
		return i.evaluateArrayExpr(e, env)

	case AsyncExpr:
		return i.evaluateAsyncExpr(e, env)

	case AwaitExpr:
		return i.evaluateAwaitExpr(e, env)

	case ArrayIndexExpr:
		return i.evaluateArrayIndexExpr(e, env)

	case MatchExpr:
		return i.evaluateMatchExpr(e, env)

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// evaluateAsyncExpr executes an async block in a goroutine and returns a Future
func (i *Interpreter) evaluateAsyncExpr(expr AsyncExpr, env *Environment) (interface{}, error) {
	// Create a new Future to represent the pending result
	future := NewFuture()

	// Create a child environment for the async block
	// This captures the current scope for use in the goroutine
	asyncEnv := NewChildEnvironment(env)

	// Execute the async block in a separate goroutine
	go func() {
		// Execute the statements in the async block
		result, err := i.executeStatements(expr.Body, asyncEnv)

		if err != nil {
			// Check if it's a return value (which is normal for returning from async blocks)
			if retErr, ok := err.(*returnValue); ok {
				future.Resolve(retErr.value)
			} else {
				future.Reject(err)
			}
		} else {
			future.Resolve(result)
		}
	}()

	// Return the Future immediately (non-blocking)
	return future, nil
}

// evaluateAwaitExpr waits for a Future to complete and returns its value
func (i *Interpreter) evaluateAwaitExpr(expr AwaitExpr, env *Environment) (interface{}, error) {
	// Evaluate the expression to get the Future
	val, err := i.EvaluateExpression(expr.Expr, env)
	if err != nil {
		return nil, err
	}

	// Check if the value is a Future
	future, ok := val.(*Future)
	if !ok {
		return nil, fmt.Errorf("await requires a Future, got %T", val)
	}

	// Wait for the Future to complete and return its value
	return future.Await()
}

// evaluateArrayIndexExpr evaluates array indexing: array[index]
func (i *Interpreter) evaluateArrayIndexExpr(expr ArrayIndexExpr, env *Environment) (interface{}, error) {
	// Evaluate the array expression
	arrayVal, err := i.EvaluateExpression(expr.Array, env)
	if err != nil {
		return nil, err
	}

	// Evaluate the index expression
	indexVal, err := i.EvaluateExpression(expr.Index, env)
	if err != nil {
		return nil, err
	}

	// Handle array indexing
	if arr, ok := arrayVal.([]interface{}); ok {
		// Get the index as int64
		var index int64
		switch idx := indexVal.(type) {
		case int64:
			index = idx
		case int:
			index = int64(idx)
		default:
			return nil, fmt.Errorf("array index must be an integer, got %T", indexVal)
		}

		// Check bounds
		if index < 0 || int(index) >= len(arr) {
			return nil, fmt.Errorf("array index out of bounds: %d (length: %d)", index, len(arr))
		}

		return arr[index], nil
	}

	// Handle map/object indexing with string key
	if obj, ok := arrayVal.(map[string]interface{}); ok {
		keyStr, ok := indexVal.(string)
		if !ok {
			return nil, fmt.Errorf("map key must be a string, got %T", indexVal)
		}
		if val, exists := obj[keyStr]; exists {
			return val, nil
		}
		return nil, fmt.Errorf("key '%s' not found in object", keyStr)
	}

	return nil, fmt.Errorf("cannot index %T", arrayVal)
}

// evaluateLiteral converts a literal AST node to a runtime value
func (i *Interpreter) evaluateLiteral(lit Literal) (interface{}, error) {
	switch l := lit.(type) {
	case IntLiteral:
		return l.Value, nil
	case StringLiteral:
		return l.Value, nil
	case BoolLiteral:
		return l.Value, nil
	case FloatLiteral:
		return l.Value, nil
	default:
		return nil, fmt.Errorf("unsupported literal type: %T", lit)
	}
}

// evaluateBinaryOp evaluates a binary operation
func (i *Interpreter) evaluateBinaryOp(expr BinaryOpExpr, env *Environment) (interface{}, error) {
	left, err := i.EvaluateExpression(expr.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := i.EvaluateExpression(expr.Right, env)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case Add:
		return i.evaluateAdd(left, right)
	case Sub:
		return i.evaluateSub(left, right)
	case Mul:
		return i.evaluateMul(left, right)
	case Div:
		return i.evaluateDiv(left, right)
	case Eq:
		return i.evaluateEq(left, right)
	case Ne:
		return i.evaluateNe(left, right)
	case Lt:
		return i.evaluateLt(left, right)
	case Le:
		return i.evaluateLe(left, right)
	case Gt:
		return i.evaluateGt(left, right)
	case Ge:
		return i.evaluateGe(left, right)
	case And:
		return i.evaluateAnd(left, right)
	case Or:
		return i.evaluateOr(left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operator: %s", expr.Op)
	}
}

// evaluateUnaryOp evaluates a unary operation
func (i *Interpreter) evaluateUnaryOp(expr UnaryOpExpr, env *Environment) (interface{}, error) {
	right, err := i.EvaluateExpression(expr.Right, env)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case Not:
		boolVal, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("type error: logical NOT requires boolean operand, got %T", right)
		}
		return !boolVal, nil

	case Neg:
		switch v := right.(type) {
		case int64:
			return -v, nil
		case float64:
			return -v, nil
		default:
			return nil, fmt.Errorf("type error: unary negation requires numeric operand, got %T", right)
		}

	default:
		return nil, fmt.Errorf("unsupported unary operator: %s", expr.Op)
	}
}

// evaluateAdd handles addition and string concatenation
func (i *Interpreter) evaluateAdd(left, right interface{}) (interface{}, error) {
	// String concatenation
	if leftStr, ok := left.(string); ok {
		rightStr, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("cannot add string and %T", right)
		}
		return leftStr + rightStr, nil
	}

	// Numeric addition with automatic coercion
	coercedLeft, coercedRight, _ := CoerceNumeric(left, right)

	// Integer addition
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt + rightInt, nil
		}
	}

	// Float addition
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat + rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot add %T and %T", left, right)
}

// evaluateSub handles subtraction
func (i *Interpreter) evaluateSub(left, right interface{}) (interface{}, error) {
	// Numeric subtraction - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking for subtraction)
	if coerced {
		return nil, fmt.Errorf("cannot subtract %T and %T", left, right)
	}

	// Integer subtraction
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt - rightInt, nil
		}
	}

	// Float subtraction
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat - rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot subtract %T and %T", left, right)
}

// evaluateMul handles multiplication
func (i *Interpreter) evaluateMul(left, right interface{}) (interface{}, error) {
	// Numeric multiplication - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking for multiplication)
	if coerced {
		return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
	}

	// Integer multiplication
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt * rightInt, nil
		}
	}

	// Float multiplication
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat * rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
}

// evaluateDiv handles division
func (i *Interpreter) evaluateDiv(left, right interface{}) (interface{}, error) {
	// Numeric division - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking for division)
	if coerced {
		return nil, fmt.Errorf("cannot divide %T and %T", left, right)
	}

	// Integer division
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			if rightInt == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return leftInt / rightInt, nil
		}
	}

	// Float division
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			if rightFloat == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return leftFloat / rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot divide %T and %T", left, right)
}

// evaluateEq handles equality comparison
func (i *Interpreter) evaluateEq(left, right interface{}) (interface{}, error) {
	// Try numeric coercion first
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	if coerced {
		// Compare coerced numeric values
		return coercedLeft == coercedRight, nil
	}

	// For non-numeric types, compare directly
	return left == right, nil
}

// evaluateNe handles inequality comparison
func (i *Interpreter) evaluateNe(left, right interface{}) (interface{}, error) {
	result, err := i.evaluateEq(left, right)
	if err != nil {
		return false, err
	}
	if boolResult, ok := result.(bool); ok {
		return !boolResult, nil
	}
	return nil, fmt.Errorf("unexpected result type from equality comparison")
}

// evaluateLt handles less than comparison
func (i *Interpreter) evaluateLt(left, right interface{}) (interface{}, error) {
	// Numeric comparison - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking)
	if coerced {
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	}

	// Integer comparison
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt < rightInt, nil
		}
	}

	// Float comparison
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat < rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot compare %T and %T", left, right)
}

// evaluateLe handles less than or equal comparison
func (i *Interpreter) evaluateLe(left, right interface{}) (interface{}, error) {
	// Numeric comparison - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking)
	if coerced {
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	}

	// Integer comparison
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt <= rightInt, nil
		}
	}

	// Float comparison
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat <= rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot compare %T and %T", left, right)
}

// evaluateGt handles greater than comparison
func (i *Interpreter) evaluateGt(left, right interface{}) (interface{}, error) {
	// Numeric comparison - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking)
	if coerced {
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	}

	// Integer comparison
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt > rightInt, nil
		}
	}

	// Float comparison
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat > rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot compare %T and %T", left, right)
}

// evaluateGe handles greater than or equal comparison
func (i *Interpreter) evaluateGe(left, right interface{}) (interface{}, error) {
	// Numeric comparison - do not allow automatic coercion
	coercedLeft, coercedRight, coerced := CoerceNumeric(left, right)

	// If coercion happened, return error (strict type checking)
	if coerced {
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	}

	// Integer comparison
	if leftInt, ok := coercedLeft.(int64); ok {
		if rightInt, ok := coercedRight.(int64); ok {
			return leftInt >= rightInt, nil
		}
	}

	// Float comparison
	if leftFloat, ok := coercedLeft.(float64); ok {
		if rightFloat, ok := coercedRight.(float64); ok {
			return leftFloat >= rightFloat, nil
		}
	}

	return nil, fmt.Errorf("cannot compare %T and %T", left, right)
}

// evaluateAnd handles logical AND operation
func (i *Interpreter) evaluateAnd(left, right interface{}) (interface{}, error) {
	// Check that left is a boolean
	leftBool, ok := left.(bool)
	if !ok {
		return nil, fmt.Errorf("logical AND operator requires boolean operands, got %T", left)
	}

	// Check that right is a boolean
	rightBool, ok := right.(bool)
	if !ok {
		return nil, fmt.Errorf("logical AND operator requires boolean operands, got %T", right)
	}

	// Both operands must be true for AND to return true
	return leftBool && rightBool, nil
}

// evaluateOr handles logical OR operation
func (i *Interpreter) evaluateOr(left, right interface{}) (interface{}, error) {
	// Check that left is a boolean
	leftBool, ok := left.(bool)
	if !ok {
		return nil, fmt.Errorf("logical OR operator requires boolean operands, got %T", left)
	}

	// Check that right is a boolean
	rightBool, ok := right.(bool)
	if !ok {
		return nil, fmt.Errorf("logical OR operator requires boolean operands, got %T", right)
	}

	// At least one operand must be true for OR to return true
	return leftBool || rightBool, nil
}

// evaluateFieldAccess handles field access on objects (maps) and database handlers
func (i *Interpreter) evaluateFieldAccess(expr FieldAccessExpr, env *Environment) (interface{}, error) {
	obj, err := i.EvaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}

	// Handle database table access (db.tablename)
	// The db handler will have a Table() method that returns a table handler
	// This is accessed via reflection or type assertion
	if dbHandler, ok := obj.(interface{ Table(string) interface{} }); ok {
		return dbHandler.Table(expr.Field), nil
	}

	// Handle map field access
	if objMap, ok := obj.(map[string]interface{}); ok {
		if val, exists := objMap[expr.Field]; exists {
			return val, nil
		}
		return nil, fmt.Errorf("field %s not found on object", expr.Field)
	}

	return nil, fmt.Errorf("cannot access field %s on %T", expr.Field, obj)
}

// evaluateFunctionCall handles function calls and method calls
func (i *Interpreter) evaluateFunctionCall(expr FunctionCallExpr, env *Environment) (interface{}, error) {
	// Handle built-in functions first (before checking for method calls)
	switch expr.Name {
	case "time.now":
		// Return a mock timestamp for now
		return int64(1234567890), nil

	case "now":
		// Return a mock timestamp
		return int64(1234567890), nil

	case "upper":
		// Convert string to uppercase
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("upper() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("upper() expects a string argument, got %T", arg)
		}
		return strings.ToUpper(str), nil

	case "lower":
		// Convert string to lowercase
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("lower() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("lower() expects a string argument, got %T", arg)
		}
		return strings.ToLower(str), nil

	case "trim":
		// Remove leading/trailing whitespace
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("trim() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("trim() expects a string argument, got %T", arg)
		}
		return strings.TrimSpace(str), nil

	case "split":
		// Split string into array
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("split() expects 2 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("split() expects first argument to be a string, got %T", strArg)
		}
		delimArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		delim, ok := delimArg.(string)
		if !ok {
			return nil, fmt.Errorf("split() expects second argument to be a string, got %T", delimArg)
		}
		parts := strings.Split(str, delim)
		result := make([]interface{}, len(parts))
		for i, part := range parts {
			result[i] = part
		}
		return result, nil

	case "join":
		// Join array into string
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("join() expects 2 arguments, got %d", len(expr.Args))
		}
		arrArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		arr, ok := arrArg.([]interface{})
		if !ok {
			return nil, fmt.Errorf("join() expects first argument to be an array, got %T", arrArg)
		}
		delimArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		delim, ok := delimArg.(string)
		if !ok {
			return nil, fmt.Errorf("join() expects second argument to be a string, got %T", delimArg)
		}
		strParts := make([]string, len(arr))
		for i, elem := range arr {
			strParts[i] = fmt.Sprintf("%v", elem)
		}
		return strings.Join(strParts, delim), nil

	case "contains":
		// Check if string contains substring
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("contains() expects 2 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("contains() expects first argument to be a string, got %T", strArg)
		}
		substrArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		substr, ok := substrArg.(string)
		if !ok {
			return nil, fmt.Errorf("contains() expects second argument to be a string, got %T", substrArg)
		}
		return strings.Contains(str, substr), nil

	case "replace":
		// Replace occurrences in string
		if len(expr.Args) != 3 {
			return nil, fmt.Errorf("replace() expects 3 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("replace() expects first argument to be a string, got %T", strArg)
		}
		oldArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		old, ok := oldArg.(string)
		if !ok {
			return nil, fmt.Errorf("replace() expects second argument to be a string, got %T", oldArg)
		}
		newArg, err := i.EvaluateExpression(expr.Args[2], env)
		if err != nil {
			return nil, err
		}
		new, ok := newArg.(string)
		if !ok {
			return nil, fmt.Errorf("replace() expects third argument to be a string, got %T", newArg)
		}
		return strings.ReplaceAll(str, old, new), nil

	case "substring":
		// Get substring
		if len(expr.Args) != 3 {
			return nil, fmt.Errorf("substring() expects 3 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("substring() expects first argument to be a string, got %T", strArg)
		}
		startArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		start, ok := startArg.(int64)
		if !ok {
			// Try to convert int to int64
			if startInt, ok := startArg.(int); ok {
				start = int64(startInt)
			} else {
				return nil, fmt.Errorf("substring() expects second argument to be an integer, got %T", startArg)
			}
		}
		endArg, err := i.EvaluateExpression(expr.Args[2], env)
		if err != nil {
			return nil, err
		}
		end, ok := endArg.(int64)
		if !ok {
			// Try to convert int to int64
			if endInt, ok := endArg.(int); ok {
				end = int64(endInt)
			} else {
				return nil, fmt.Errorf("substring() expects third argument to be an integer, got %T", endArg)
			}
		}
		if start < 0 || end < 0 {
			return nil, fmt.Errorf("substring() indices must be non-negative")
		}
		if start > end {
			return nil, fmt.Errorf("substring() start index must be less than or equal to end index")
		}
		if int(end) > len(str) {
			end = int64(len(str))
		}
		if int(start) > len(str) {
			start = int64(len(str))
		}
		return str[start:end], nil

	case "length":
		// Get length of string or array
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("length() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		switch v := arg.(type) {
		case string:
			return int64(len(v)), nil
		case []interface{}:
			return int64(len(v)), nil
		default:
			return nil, fmt.Errorf("length() expects a string or array argument, got %T", arg)
		}

	case "startsWith":
		// Check if string starts with prefix
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("startsWith() expects 2 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("startsWith() expects first argument to be a string, got %T", strArg)
		}
		prefixArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		prefix, ok := prefixArg.(string)
		if !ok {
			return nil, fmt.Errorf("startsWith() expects second argument to be a string, got %T", prefixArg)
		}
		return strings.HasPrefix(str, prefix), nil

	case "endsWith":
		// Check if string ends with suffix
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("endsWith() expects 2 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("endsWith() expects first argument to be a string, got %T", strArg)
		}
		suffixArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		suffix, ok := suffixArg.(string)
		if !ok {
			return nil, fmt.Errorf("endsWith() expects second argument to be a string, got %T", suffixArg)
		}
		return strings.HasSuffix(str, suffix), nil

	case "indexOf":
		// Find first occurrence of substring
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("indexOf() expects 2 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("indexOf() expects first argument to be a string, got %T", strArg)
		}
		substrArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		substr, ok := substrArg.(string)
		if !ok {
			return nil, fmt.Errorf("indexOf() expects second argument to be a string, got %T", substrArg)
		}
		return int64(strings.Index(str, substr)), nil

	case "charAt":
		// Get character at index
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("charAt() expects 2 arguments, got %d", len(expr.Args))
		}
		strArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := strArg.(string)
		if !ok {
			return nil, fmt.Errorf("charAt() expects first argument to be a string, got %T", strArg)
		}
		indexArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		index, ok := indexArg.(int64)
		if !ok {
			return nil, fmt.Errorf("charAt() expects second argument to be an integer, got %T", indexArg)
		}
		if index < 0 || int(index) >= len(str) {
			return nil, fmt.Errorf("charAt() index out of bounds: %d", index)
		}
		return string(str[index]), nil

	case "parseInt":
		// Parse string to integer
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("parseInt() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("parseInt() expects a string argument, got %T", arg)
		}
		str = strings.TrimSpace(str)
		var result int64
		_, err = fmt.Sscanf(str, "%d", &result)
		if err != nil {
			return nil, fmt.Errorf("parseInt() failed to parse '%s': %v", str, err)
		}
		return result, nil

	case "parseFloat":
		// Parse string to float
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("parseFloat() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		str, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("parseFloat() expects a string argument, got %T", arg)
		}
		str = strings.TrimSpace(str)
		var result float64
		_, err = fmt.Sscanf(str, "%f", &result)
		if err != nil {
			return nil, fmt.Errorf("parseFloat() failed to parse '%s': %v", str, err)
		}
		return result, nil

	case "toString":
		// Convert value to string
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("toString() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("%v", arg), nil

	case "abs":
		// Absolute value
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("abs() expects 1 argument, got %d", len(expr.Args))
		}
		arg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		switch v := arg.(type) {
		case int64:
			if v < 0 {
				return -v, nil
			}
			return v, nil
		case float64:
			if v < 0 {
				return -v, nil
			}
			return v, nil
		default:
			return nil, fmt.Errorf("abs() expects a numeric argument, got %T", arg)
		}

	case "min":
		// Minimum of two values
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("min() expects 2 arguments, got %d", len(expr.Args))
		}
		leftArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		rightArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		switch l := leftArg.(type) {
		case int64:
			r, ok := rightArg.(int64)
			if !ok {
				return nil, fmt.Errorf("min() arguments must be same type")
			}
			if l < r {
				return l, nil
			}
			return r, nil
		case float64:
			r, ok := rightArg.(float64)
			if !ok {
				return nil, fmt.Errorf("min() arguments must be same type")
			}
			if l < r {
				return l, nil
			}
			return r, nil
		default:
			return nil, fmt.Errorf("min() expects numeric arguments, got %T", leftArg)
		}

	case "max":
		// Maximum of two values
		if len(expr.Args) != 2 {
			return nil, fmt.Errorf("max() expects 2 arguments, got %d", len(expr.Args))
		}
		leftArg, err := i.EvaluateExpression(expr.Args[0], env)
		if err != nil {
			return nil, err
		}
		rightArg, err := i.EvaluateExpression(expr.Args[1], env)
		if err != nil {
			return nil, err
		}
		switch l := leftArg.(type) {
		case int64:
			r, ok := rightArg.(int64)
			if !ok {
				return nil, fmt.Errorf("max() arguments must be same type")
			}
			if l > r {
				return l, nil
			}
			return r, nil
		case float64:
			r, ok := rightArg.(float64)
			if !ok {
				return nil, fmt.Errorf("max() arguments must be same type")
			}
			if l > r {
				return l, nil
			}
			return r, nil
		default:
			return nil, fmt.Errorf("max() expects numeric arguments, got %T", leftArg)
		}
	}

	// Check if this is a method call (contains a dot)
	if dotIdx := strings.Index(expr.Name, "."); dotIdx != -1 {
		// Split into object and method
		parts := strings.SplitN(expr.Name, ".", 2)
		objName := parts[0]
		methodPath := parts[1]

		// Get the object
		obj, err := env.Get(objName)
		if err != nil {
			return nil, fmt.Errorf("undefined object: %s", objName)
		}

		// Handle nested method calls (e.g., db.table.method())
		for strings.Contains(methodPath, ".") {
			parts := strings.SplitN(methodPath, ".", 2)
			fieldName := parts[0]
			methodPath = parts[1]

			// Access the field
			if dbHandler, ok := obj.(interface{ Table(string) interface{} }); ok {
				obj = dbHandler.Table(fieldName)
			} else if objMap, ok := obj.(map[string]interface{}); ok {
				if val, exists := objMap[fieldName]; exists {
					obj = val
				} else {
					return nil, fmt.Errorf("field %s not found", fieldName)
				}
			} else {
				return nil, fmt.Errorf("cannot access field %s on %T", fieldName, obj)
			}
		}

		// Now we have the final object and method name
		methodName := methodPath

		// Evaluate arguments
		args := make([]interface{}, len(expr.Args))
		for idx, arg := range expr.Args {
			val, err := i.EvaluateExpression(arg, env)
			if err != nil {
				return nil, err
			}
			args[idx] = val
		}

		// Handle special case for array.length()
		if methodName == "length" {
			if arr, ok := obj.([]interface{}); ok {
				return int64(len(arr)), nil
			}
		}

		// Check if obj is a map (module namespace) and methodName is a function
		if objMap, ok := obj.(map[string]interface{}); ok {
			if fn, exists := objMap[methodName]; exists {
				// If it's a Function, execute it
				if fnDef, ok := fn.(*Function); ok {
					return i.executeFunction(*fnDef, expr.Args, env)
				}
				if fnDef, ok := fn.(Function); ok {
					return i.executeFunction(fnDef, expr.Args, env)
				}
				// If it's something else callable, continue to reflection
			}
		}

		// Call the method using reflection
		return CallMethod(obj, strings.Title(methodName), args...)
	}

	// Check if it's a user-defined function
	fn, err := env.Get(expr.Name)
	if err != nil {
		return nil, fmt.Errorf("undefined function: %s", expr.Name)
	}

	// If it's a Function AST node, execute it
	if fnDef, ok := fn.(Function); ok {
		// Check if this is a generic function
		if len(fnDef.TypeParams) > 0 {
			return i.executeGenericFunction(fnDef, expr.TypeArgs, expr.Args, env)
		}
		return i.executeFunction(fnDef, expr.Args, env)
	}

	return nil, fmt.Errorf("not a function: %s", expr.Name)
}

// evaluateObjectExpr handles object literal expressions
func (i *Interpreter) evaluateObjectExpr(expr ObjectExpr, env *Environment) (interface{}, error) {
	obj := make(map[string]interface{})

	for _, field := range expr.Fields {
		// Evaluate the field value expression
		value, err := i.EvaluateExpression(field.Value, env)
		if err != nil {
			return nil, fmt.Errorf("error evaluating field %s: %v", field.Key, err)
		}
		obj[field.Key] = value
	}

	return obj, nil
}

// evaluateArrayExpr handles array literal expressions
func (i *Interpreter) evaluateArrayExpr(expr ArrayExpr, env *Environment) (interface{}, error) {
	arr := make([]interface{}, 0, len(expr.Elements))

	for _, elem := range expr.Elements {
		// Evaluate each element expression
		value, err := i.EvaluateExpression(elem, env)
		if err != nil {
			return nil, fmt.Errorf("error evaluating array element: %v", err)
		}
		arr = append(arr, value)
	}

	return arr, nil
}

// executeFunction executes a user-defined function
func (i *Interpreter) executeFunction(fn Function, args []Expr, env *Environment) (interface{}, error) {
	// Create a new environment for the function
	fnEnv := NewChildEnvironment(env)

	// Evaluate arguments and bind to parameters
	if len(args) != len(fn.Params) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d", fn.Name, len(fn.Params), len(args))
	}

	for idx, param := range fn.Params {
		argVal, err := i.EvaluateExpression(args[idx], env)
		if err != nil {
			return nil, err
		}

		// Validate argument type matches parameter type annotation
		if param.TypeAnnotation != nil {
			if err := i.typeChecker.CheckType(argVal, param.TypeAnnotation); err != nil {
				return nil, fmt.Errorf("argument %d (%s): %v", idx+1, param.Name, err)
			}
		}

		fnEnv.Define(param.Name, argVal)
	}

	// Execute function body
	result, err := i.executeStatements(fn.Body, fnEnv)
	if err != nil {
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	// Validate return value matches declared return type
	if fn.ReturnType != nil {
		if err := i.typeChecker.CheckType(result, fn.ReturnType); err != nil {
			return nil, fmt.Errorf("return type mismatch in function %s: %v", fn.Name, err)
		}
	}

	return result, nil
}

// executeGenericFunction executes a generic function with type arguments
func (i *Interpreter) executeGenericFunction(fn Function, typeArgs []Type, args []Expr, env *Environment) (interface{}, error) {
	// Evaluate all arguments first (we need values for type inference)
	argValues := make([]interface{}, len(args))
	for idx, arg := range args {
		val, err := i.EvaluateExpression(arg, env)
		if err != nil {
			return nil, err
		}
		argValues[idx] = val
	}

	// If type arguments are not provided, try to infer them
	var resolvedTypeArgs []Type
	if len(typeArgs) == 0 {
		inferred, err := i.typeChecker.InferTypeArguments(fn, argValues)
		if err != nil {
			return nil, fmt.Errorf("cannot infer type arguments for generic function %s: %v", fn.Name, err)
		}
		resolvedTypeArgs = inferred
	} else {
		// Validate that the correct number of type arguments were provided
		if len(typeArgs) != len(fn.TypeParams) {
			return nil, fmt.Errorf("generic function %s expects %d type arguments, got %d",
				fn.Name, len(fn.TypeParams), len(typeArgs))
		}
		resolvedTypeArgs = typeArgs
	}

	// Instantiate the generic function with resolved type arguments
	instantiatedFn, typeBindings, err := i.typeChecker.InstantiateGenericFunction(fn, resolvedTypeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate generic function %s: %v", fn.Name, err)
	}

	// Push type bindings onto the type scope for nested generic type resolution
	i.typeChecker.PushTypeScope(typeBindings)
	defer func() {
		names := make([]string, 0, len(typeBindings))
		for name := range typeBindings {
			names = append(names, name)
		}
		i.typeChecker.PopTypeScope(names)
	}()

	// Create a new environment for the function
	fnEnv := NewChildEnvironment(env)

	// Validate argument count
	if len(argValues) != len(instantiatedFn.Params) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d",
			fn.Name, len(instantiatedFn.Params), len(argValues))
	}

	// Bind arguments to parameters with type checking
	for idx, param := range instantiatedFn.Params {
		argVal := argValues[idx]

		// Validate argument type matches the instantiated parameter type
		if param.TypeAnnotation != nil {
			if err := i.typeChecker.CheckType(argVal, param.TypeAnnotation); err != nil {
				return nil, fmt.Errorf("argument %d (%s): %v", idx+1, param.Name, err)
			}
		}

		fnEnv.Define(param.Name, argVal)
	}

	// Execute function body
	result, err := i.executeStatements(fn.Body, fnEnv)
	if err != nil {
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	// Validate return value matches the instantiated return type
	if instantiatedFn.ReturnType != nil {
		if err := i.typeChecker.CheckType(result, instantiatedFn.ReturnType); err != nil {
			return nil, fmt.Errorf("return type mismatch in function %s: %v", fn.Name, err)
		}
	}

	return result, nil
}

// extractPathParams extracts path parameters from a route path
func extractPathParams(path string, actualPath string) (map[string]string, error) {
	// Remove query string from actualPath if present
	actualPathWithoutQuery := actualPath
	if idx := strings.Index(actualPath, "?"); idx != -1 {
		actualPathWithoutQuery = actualPath[:idx]
	}

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	actualParts := strings.Split(strings.Trim(actualPathWithoutQuery, "/"), "/")

	if len(pathParts) != len(actualParts) {
		return nil, fmt.Errorf("path mismatch: expected %s, got %s", path, actualPathWithoutQuery)
	}

	params := make(map[string]string)
	for i, part := range pathParts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params[paramName] = actualParts[i]
		} else if part != actualParts[i] {
			return nil, fmt.Errorf("path mismatch: expected %s, got %s", path, actualPathWithoutQuery)
		}
	}

	return params, nil
}

// extractQueryParams is deprecated - use ExtractRawQueryParams and ProcessQueryParams instead.
// Kept for backward compatibility with any code that may still reference it.

// evaluateMatchExpr evaluates a match expression
func (i *Interpreter) evaluateMatchExpr(expr MatchExpr, env *Environment) (interface{}, error) {
	// Evaluate the value to match against
	value, err := i.EvaluateExpression(expr.Value, env)
	if err != nil {
		return nil, err
	}

	// Try each case in order
	for _, matchCase := range expr.Cases {
		// Create a new environment for pattern bindings
		caseEnv := NewChildEnvironment(env)

		// Try to match the pattern
		matched, err := i.matchPattern(matchCase.Pattern, value, caseEnv)
		if err != nil {
			return nil, err
		}

		if matched {
			// Check guard condition if present
			if matchCase.Guard != nil {
				guardResult, err := i.EvaluateExpression(matchCase.Guard, caseEnv)
				if err != nil {
					return nil, err
				}

				guardBool, ok := guardResult.(bool)
				if !ok {
					return nil, fmt.Errorf("match guard must evaluate to boolean, got %T", guardResult)
				}

				if !guardBool {
					// Guard failed, try next case
					continue
				}
			}

			// Pattern matched (and guard passed if present), evaluate body
			return i.EvaluateExpression(matchCase.Body, caseEnv)
		}
	}

	// No pattern matched - return nil (non-exhaustive match)
	return nil, nil
}

// matchPattern attempts to match a value against a pattern
// Returns true if matched, and binds any pattern variables to the environment
func (i *Interpreter) matchPattern(pattern Pattern, value interface{}, env *Environment) (bool, error) {
	switch p := pattern.(type) {
	case LiteralPattern:
		return i.matchLiteralPattern(p, value)

	case VariablePattern:
		// Variable pattern always matches and binds the value
		env.Define(p.Name, value)
		return true, nil

	case WildcardPattern:
		// Wildcard always matches
		return true, nil

	case ObjectPattern:
		return i.matchObjectPattern(p, value, env)

	case ArrayPattern:
		return i.matchArrayPattern(p, value, env)

	default:
		return false, fmt.Errorf("unsupported pattern type: %T", pattern)
	}
}

// matchLiteralPattern matches a value against a literal pattern
func (i *Interpreter) matchLiteralPattern(pattern LiteralPattern, value interface{}) (bool, error) {
	// Get the literal value from the pattern
	patternValue, err := i.evaluateLiteral(pattern.Value)
	if err != nil {
		return false, err
	}

	// Compare values (using the same equality logic as ==)
	result, err := i.evaluateEq(patternValue, value)
	if err != nil {
		return false, nil // Type mismatch means no match
	}

	matched, ok := result.(bool)
	if !ok {
		return false, nil
	}

	return matched, nil
}

// matchObjectPattern matches a value against an object destructuring pattern
func (i *Interpreter) matchObjectPattern(pattern ObjectPattern, value interface{}, env *Environment) (bool, error) {
	// Value must be a map
	objMap, ok := value.(map[string]interface{})
	if !ok {
		return false, nil
	}

	// Match each field in the pattern
	for _, field := range pattern.Fields {
		// Check if the field exists in the object
		fieldValue, exists := objMap[field.Key]
		if !exists {
			return false, nil
		}

		if field.Pattern != nil {
			// Match the nested pattern
			matched, err := i.matchPattern(field.Pattern, fieldValue, env)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
		} else {
			// No nested pattern, bind field name as variable
			env.Define(field.Key, fieldValue)
		}
	}

	return true, nil
}

// matchArrayPattern matches a value against an array destructuring pattern
func (i *Interpreter) matchArrayPattern(pattern ArrayPattern, value interface{}, env *Environment) (bool, error) {
	// Value must be a slice
	arr, ok := value.([]interface{})
	if !ok {
		return false, nil
	}

	// If there's a rest pattern, we need at least len(Elements) elements
	// Otherwise, we need exactly len(Elements) elements
	if pattern.Rest != nil {
		if len(arr) < len(pattern.Elements) {
			return false, nil
		}
	} else {
		if len(arr) != len(pattern.Elements) {
			return false, nil
		}
	}

	// Match each element pattern
	for idx, elemPattern := range pattern.Elements {
		matched, err := i.matchPattern(elemPattern, arr[idx], env)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	// If there's a rest pattern, bind the remaining elements
	if pattern.Rest != nil {
		rest := arr[len(pattern.Elements):]
		env.Define(*pattern.Rest, rest)
	}

	return true, nil
}
