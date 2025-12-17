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

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
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

// extractQueryParams extracts query parameters from a URL path
func extractQueryParams(path string) map[string]interface{} {
	queryParams := make(map[string]interface{})

	// Find the query string portion
	idx := strings.Index(path, "?")
	if idx == -1 {
		return queryParams // No query string, return empty map
	}

	queryString := path[idx+1:]
	if queryString == "" {
		return queryParams
	}

	// Parse key=value pairs
	pairs := strings.Split(queryString, "&")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Try to parse as number, otherwise store as string
			// This allows GLYPHLANG code to use query.page > 0 checks
			if strings.Contains(value, ".") {
				// Try as float
				var floatVal float64
				if _, err := fmt.Sscanf(value, "%f", &floatVal); err == nil {
					queryParams[key] = floatVal
				} else {
					queryParams[key] = value
				}
			} else {
				// Try as int
				var intVal int
				if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
					queryParams[key] = intVal
				} else {
					queryParams[key] = value
				}
			}
		} else if len(parts) == 1 {
			// Key without value, set to empty string
			queryParams[parts[0]] = ""
		}
	}

	return queryParams
}
