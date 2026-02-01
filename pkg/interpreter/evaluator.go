package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"fmt"
	"strings"
	"unicode"
)

// capitalizeFirst capitalizes only the first letter of a string, preserving the rest.
// This properly handles camelCase: "countWhere" -> "CountWhere", "nextId" -> "NextId"
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// EvaluateExpression evaluates an expression and returns its value
func (i *Interpreter) EvaluateExpression(expr Expr, env *Environment) (interface{}, error) {
	i.evalDepth++
	if i.evalDepth > maxEvalDepth {
		i.evalDepth--
		return nil, fmt.Errorf("maximum evaluation depth exceeded (%d levels)", maxEvalDepth)
	}
	defer func() { i.evalDepth-- }()
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

	case LambdaExpr:
		return i.evaluateLambdaExpr(e, env)

	case PipeExpr:
		return i.evaluatePipeExpr(e, env)

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
	case NullLiteral:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported literal type: %T", lit)
	}
}

// evaluateBinaryOp evaluates a binary operation
func (i *Interpreter) evaluateBinaryOp(expr BinaryOpExpr, env *Environment) (interface{}, error) {
	// Short-circuit evaluation for logical operators
	if expr.Op == And {
		left, err := i.EvaluateExpression(expr.Left, env)
		if err != nil {
			return nil, err
		}
		leftBool, ok := left.(bool)
		if !ok {
			return nil, fmt.Errorf("logical AND operator requires boolean operands, got %T", left)
		}
		if !leftBool {
			return false, nil
		}
		right, err := i.EvaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rightBool, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("logical AND operator requires boolean operands, got %T", right)
		}
		return rightBool, nil
	}
	if expr.Op == Or {
		left, err := i.EvaluateExpression(expr.Left, env)
		if err != nil {
			return nil, err
		}
		leftBool, ok := left.(bool)
		if !ok {
			return nil, fmt.Errorf("logical OR operator requires boolean operands, got %T", left)
		}
		if leftBool {
			return true, nil
		}
		right, err := i.EvaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rightBool, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("logical OR operator requires boolean operands, got %T", right)
		}
		return rightBool, nil
	}

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

	// Handle database table access (db.tablename) using reflection
	// Check if obj has a Table(string) method
	if HasMethod(obj, "Table") {
		result, callErr := CallMethod(obj, "Table", expr.Field)
		if callErr == nil {
			return result, nil
		}
	}

	// Handle Result type field access
	if result, ok := obj.(*ResultValue); ok {
		switch expr.Field {
		case "ok":
			return result.IsOk(), nil
		case "error", "err":
			return result.ErrValue(), nil
		case "value":
			return result.OkValue(), nil
		case "isOk":
			return result.IsOk(), nil
		case "isErr":
			return result.IsErr(), nil
		default:
			return nil, fmt.Errorf("Result has no field '%s'", expr.Field)
		}
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
	// Handle built-in functions via dispatch table
	if fn, ok := builtinFuncs[expr.Name]; ok {
		return fn(i, expr.Args, env)
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

		// Handle Result type methods (result.map, result.isOk, etc.)
		if result, ok := obj.(*ResultValue); ok {
			return i.evaluateResultMethod(result, methodName, args, env)
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
		// Capitalize first letter only, preserving camelCase (e.g., "countWhere" -> "CountWhere")
		capitalizedName := capitalizeFirst(methodName)
		return CallMethod(obj, capitalizedName, args...)
	}

	// Check if it's a user-defined function
	fn, err := env.Get(expr.Name)
	if err != nil {
		// Function not found - check if first arg is an object with this method
		// This handles the parser's transformation of obj.method() -> method(obj)
		if len(expr.Args) > 0 {
			firstArg, evalErr := i.EvaluateExpression(expr.Args[0], env)
			if evalErr == nil && firstArg != nil {
				methodName := capitalizeFirst(expr.Name)
				if HasMethod(firstArg, methodName) {
					// Evaluate remaining arguments
					args := make([]interface{}, len(expr.Args)-1)
					for idx, arg := range expr.Args[1:] {
						val, argErr := i.EvaluateExpression(arg, env)
						if argErr != nil {
							return nil, argErr
						}
						args[idx] = val
					}
					return CallMethod(firstArg, methodName, args...)
				}
			}
		}
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

// ApplyTypeDefaults applies default values from a TypeDef to an object
// Returns a new object with defaults applied for missing fields
func (i *Interpreter) ApplyTypeDefaults(obj map[string]interface{}, typeDef TypeDef, env *Environment) (map[string]interface{}, error) {
	// Create a copy of the object to avoid modifying the original
	result := make(map[string]interface{})
	for k, v := range obj {
		result[k] = v
	}

	// Apply defaults for missing fields
	for _, field := range typeDef.Fields {
		if _, exists := result[field.Name]; !exists && field.Default != nil {
			// Evaluate the default expression
			defaultVal, err := i.EvaluateExpression(field.Default, env)
			if err != nil {
				return nil, fmt.Errorf("error evaluating default for field %s: %v", field.Name, err)
			}
			result[field.Name] = defaultVal
		}
	}

	return result, nil
}

// executeFunction executes a user-defined function
func (i *Interpreter) executeFunction(fn Function, args []Expr, env *Environment) (interface{}, error) {
	// Create a new environment for the function
	fnEnv := NewChildEnvironment(env)

	// Count required parameters (those marked required without defaults)
	requiredCount := 0
	for _, param := range fn.Params {
		if param.Required && param.Default == nil {
			requiredCount++
		}
	}

	// Validate argument count
	if len(args) < requiredCount {
		return nil, fmt.Errorf("function %s expects at least %d arguments, got %d", fn.Name, requiredCount, len(args))
	}
	if len(args) > len(fn.Params) {
		return nil, fmt.Errorf("function %s expects at most %d arguments, got %d", fn.Name, len(fn.Params), len(args))
	}

	// Evaluate arguments and bind to parameters
	for idx, param := range fn.Params {
		var argVal interface{}
		var err error

		if idx < len(args) {
			// Argument was provided
			argVal, err = i.EvaluateExpression(args[idx], env)
			if err != nil {
				return nil, err
			}
		} else if param.Default != nil {
			// Use default value
			argVal, err = i.EvaluateExpression(param.Default, fnEnv)
			if err != nil {
				return nil, fmt.Errorf("error evaluating default for parameter %s: %v", param.Name, err)
			}
		} else if !param.Required {
			// Optional parameter without default gets nil
			argVal = nil
		} else {
			return nil, fmt.Errorf("missing required argument %s in function %s", param.Name, fn.Name)
		}

		// Validate argument type matches parameter type annotation
		// Optional parameters can be nil without type checking
		skipTypeCheck := argVal == nil && !param.Required
		if param.TypeAnnotation != nil && !skipTypeCheck {
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
	var objMap map[string]interface{}

	// Handle ResultValue by exposing it as a map with "ok"/"error" keys
	if result, ok := value.(*ResultValue); ok {
		if result.IsOk() {
			objMap = map[string]interface{}{
				"ok":    result.value,
				"value": result.value,
			}
		} else {
			objMap = map[string]interface{}{
				"error": result.value,
				"err":   result.value,
			}
		}
	} else {
		var ok bool
		objMap, ok = value.(map[string]interface{})
		if !ok {
			return false, nil
		}
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

// LambdaClosure represents a lambda expression with its captured environment
type LambdaClosure struct {
	Lambda LambdaExpr
	Env    *Environment
}

// evaluateLambdaExpr creates a closure from a lambda expression
func (i *Interpreter) evaluateLambdaExpr(expr LambdaExpr, env *Environment) (interface{}, error) {
	return &LambdaClosure{
		Lambda: expr,
		Env:    env,
	}, nil
}

// evaluatePipeExpr evaluates a pipe expression: left |> right
// The left value is piped as the first argument to the right function
func (i *Interpreter) evaluatePipeExpr(expr PipeExpr, env *Environment) (interface{}, error) {
	// Evaluate the left side to get the value being piped
	leftVal, err := i.EvaluateExpression(expr.Left, env)
	if err != nil {
		return nil, err
	}

	// Handle the right side based on its type
	switch right := expr.Right.(type) {
	case VariableExpr:
		// Simple function reference: x |> f
		// Look up the function and call it with leftVal as the first argument
		fn, err := env.Get(right.Name)
		if err != nil {
			return nil, fmt.Errorf("undefined function: %s", right.Name)
		}
		return i.callWithPipedArg(fn, leftVal, nil, env)

	case FunctionCallExpr:
		// Function call with args: x |> f(a, b)
		// Prepend leftVal to the existing arguments
		fn, err := env.Get(right.Name)
		if err != nil {
			return nil, fmt.Errorf("undefined function: %s", right.Name)
		}
		return i.callWithPipedArg(fn, leftVal, right.Args, env)

	case LambdaExpr:
		// Inline lambda: x |> (v -> v * 2)
		closure := &LambdaClosure{Lambda: right, Env: env}
		return i.callWithPipedArg(closure, leftVal, nil, env)

	default:
		// Try to evaluate it as an expression that results in a callable
		rightVal, err := i.EvaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		return i.callWithPipedArg(rightVal, leftVal, nil, env)
	}
}

// callWithPipedArg calls a function with the piped value as the first argument
func (i *Interpreter) callWithPipedArg(fn interface{}, pipedVal interface{}, extraArgs []Expr, env *Environment) (interface{}, error) {
	switch f := fn.(type) {
	case Function:
		// Build argument list: piped value first, then extra args
		argVals := []interface{}{pipedVal}
		for _, argExpr := range extraArgs {
			argVal, err := i.EvaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			argVals = append(argVals, argVal)
		}
		return i.executeFunctionWithValues(f, argVals, env)

	case *LambdaClosure:
		// Build argument list for lambda
		argVals := []interface{}{pipedVal}
		for _, argExpr := range extraArgs {
			argVal, err := i.EvaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			argVals = append(argVals, argVal)
		}
		return i.callLambdaClosure(f, argVals)

	case func(args ...interface{}) (interface{}, error):
		// Built-in variadic function
		argVals := []interface{}{pipedVal}
		for _, argExpr := range extraArgs {
			argVal, err := i.EvaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			argVals = append(argVals, argVal)
		}
		return f(argVals...)

	default:
		return nil, fmt.Errorf("pipe target is not a function: %T", fn)
	}
}

// executeFunctionWithValues executes a user-defined function with pre-evaluated argument values
func (i *Interpreter) executeFunctionWithValues(fn Function, argVals []interface{}, env *Environment) (interface{}, error) {
	// Create a new environment for the function
	fnEnv := NewChildEnvironment(env)

	// Count required parameters (those marked required without defaults)
	requiredCount := 0
	for _, param := range fn.Params {
		if param.Required && param.Default == nil {
			requiredCount++
		}
	}

	// Validate argument count
	if len(argVals) < requiredCount {
		return nil, fmt.Errorf("function %s expects at least %d arguments, got %d", fn.Name, requiredCount, len(argVals))
	}
	if len(argVals) > len(fn.Params) {
		return nil, fmt.Errorf("function %s expects at most %d arguments, got %d", fn.Name, len(fn.Params), len(argVals))
	}

	// Bind arguments to parameters
	for idx, param := range fn.Params {
		var argVal interface{}
		var err error

		if idx < len(argVals) {
			// Argument was provided
			argVal = argVals[idx]
		} else if param.Default != nil {
			// Use default value
			argVal, err = i.EvaluateExpression(param.Default, fnEnv)
			if err != nil {
				return nil, fmt.Errorf("error evaluating default for parameter %s: %v", param.Name, err)
			}
		} else if !param.Required {
			// Optional parameter without default gets nil
			argVal = nil
		} else {
			return nil, fmt.Errorf("missing required argument %s in function %s", param.Name, fn.Name)
		}

		// Validate argument type matches parameter type annotation
		skipTypeCheck := argVal == nil && !param.Required
		if param.TypeAnnotation != nil && !skipTypeCheck {
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

// callLambdaClosure executes a lambda closure with the given arguments
func (i *Interpreter) callLambdaClosure(closure *LambdaClosure, args []interface{}) (interface{}, error) {
	// Create a new environment for the lambda execution
	lambdaEnv := NewChildEnvironment(closure.Env)

	// Bind parameters to arguments
	for idx, param := range closure.Lambda.Params {
		if idx < len(args) {
			lambdaEnv.Define(param.Name, args[idx])
		} else if param.Default != nil {
			// Use default value if provided
			defVal, err := i.EvaluateExpression(param.Default, lambdaEnv)
			if err != nil {
				return nil, err
			}
			lambdaEnv.Define(param.Name, defVal)
		} else if !param.Required {
			lambdaEnv.Define(param.Name, nil)
		} else {
			return nil, fmt.Errorf("missing required argument: %s", param.Name)
		}
	}

	// Execute the lambda body
	if closure.Lambda.Body != nil {
		return i.EvaluateExpression(closure.Lambda.Body, lambdaEnv)
	}

	// Execute block body
	if len(closure.Lambda.Block) > 0 {
		result, err := i.executeStatements(closure.Lambda.Block, lambdaEnv)
		if err != nil {
			if retErr, ok := err.(*returnValue); ok {
				return retErr.value, nil
			}
			return nil, err
		}
		return result, nil
	}

	return nil, nil
}

// evaluateResultMethod handles method calls on ResultValue instances.
func (i *Interpreter) evaluateResultMethod(result *ResultValue, method string, args []interface{}, env *Environment) (interface{}, error) {
	switch method {
	case "isOk":
		return result.IsOk(), nil

	case "isErr":
		return result.IsErr(), nil

	case "unwrap":
		return result.Unwrap()

	case "unwrapOr":
		if len(args) != 1 {
			return nil, fmt.Errorf("unwrapOr() expects 1 argument, got %d", len(args))
		}
		return result.UnwrapOr(args[0]), nil

	case "unwrapErr":
		return result.UnwrapErr()

	case "map":
		if len(args) != 1 {
			return nil, fmt.Errorf("map() expects 1 argument (a function), got %d", len(args))
		}
		if !result.IsOk() {
			return result, nil
		}
		mapped, err := i.callFnArg(args[0], result.value, env)
		if err != nil {
			return nil, fmt.Errorf("Result.map: %w", err)
		}
		return NewOk(mapped), nil

	case "mapErr":
		if len(args) != 1 {
			return nil, fmt.Errorf("mapErr() expects 1 argument (a function), got %d", len(args))
		}
		if result.IsOk() {
			return result, nil
		}
		mapped, err := i.callFnArg(args[0], result.value, env)
		if err != nil {
			return nil, fmt.Errorf("Result.mapErr: %w", err)
		}
		return NewErr(mapped), nil

	case "andThen":
		if len(args) != 1 {
			return nil, fmt.Errorf("andThen() expects 1 argument (a function), got %d", len(args))
		}
		if !result.IsOk() {
			return result, nil
		}
		chained, err := i.callFnArg(args[0], result.value, env)
		if err != nil {
			return nil, fmt.Errorf("Result.andThen: %w", err)
		}
		if _, ok := chained.(*ResultValue); ok {
			return chained, nil
		}
		return NewOk(chained), nil

	case "orElse":
		if len(args) != 1 {
			return nil, fmt.Errorf("orElse() expects 1 argument (a function), got %d", len(args))
		}
		if result.IsOk() {
			return result, nil
		}
		fallback, err := i.callFnArg(args[0], result.value, env)
		if err != nil {
			return nil, fmt.Errorf("Result.orElse: %w", err)
		}
		if _, ok := fallback.(*ResultValue); ok {
			return fallback, nil
		}
		return NewOk(fallback), nil

	default:
		return nil, fmt.Errorf("Result has no method '%s'", method)
	}
}

// callFnArg calls a function-like value with a single argument.
// Supports Function AST nodes and LambdaClosure values.
func (i *Interpreter) callFnArg(fn interface{}, arg interface{}, env *Environment) (interface{}, error) {
	switch f := fn.(type) {
	case Function:
		fnEnv := NewChildEnvironment(env)
		if len(f.Params) > 0 {
			fnEnv.Define(f.Params[0].Name, arg)
		}
		result, err := i.executeStatements(f.Body, fnEnv)
		if err != nil {
			if retErr, ok := err.(*returnValue); ok {
				return retErr.value, nil
			}
			return nil, err
		}
		return result, nil
	case *Function:
		return i.callFnArg(*f, arg, env)
	case *LambdaClosure:
		fnEnv := NewChildEnvironment(f.Env)
		if len(f.Lambda.Params) > 0 {
			fnEnv.Define(f.Lambda.Params[0].Name, arg)
		}
		if f.Lambda.Body != nil {
			return i.EvaluateExpression(f.Lambda.Body, fnEnv)
		}
		if len(f.Lambda.Block) > 0 {
			result, err := i.executeStatements(f.Lambda.Block, fnEnv)
			if err != nil {
				if retErr, ok := err.(*returnValue); ok {
					return retErr.value, nil
				}
				return nil, err
			}
			return result, nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("expected a function, got %T", fn)
	}
}
