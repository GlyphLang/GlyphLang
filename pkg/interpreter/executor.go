package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"fmt"
	"strings"
)

// returnValue is a special error type to handle return statements
type returnValue struct {
	value interface{}
}

func (r *returnValue) Error() string {
	return "return"
}

// breakValue is a special error type to handle break statements (same pattern as returnValue above)
type breakValue struct{}

func (b *breakValue) Error() string {
	return "break"
}

// continueValue is a special error type to handle continue statements (same pattern as returnValue above)
type continueValue struct{}

func (c *continueValue) Error() string {
	return "continue"
}

// AssertionError represents a failed test assertion
type AssertionError struct {
	Message string
}

func (a *AssertionError) Error() string {
	return a.Message
}

// IsAssertionError checks if an error is an assertion error
func IsAssertionError(err error) bool {
	_, ok := err.(*AssertionError)
	return ok
}

// ValidationError is a special error type for validation failures
// This error indicates a 400 Bad Request should be returned
type ValidationError struct {
	Message string
}

func (v *ValidationError) Error() string {
	return v.Message
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// ExecuteStatement executes a single statement
func (i *Interpreter) ExecuteStatement(stmt Statement, env *Environment) (interface{}, error) {
	switch s := stmt.(type) {
	case AssignStatement:
		return i.executeAssign(s, env)

	case DbQueryStatement:
		return i.executeDbQuery(s, env)

	case ReturnStatement:
		return i.executeReturn(s, env)

	case IfStatement:
		return i.executeIf(s, env)

	case WhileStatement:
		return i.executeWhile(s, env)

	case ForStatement:
		return i.executeFor(s, env)

	case SwitchStatement:
		return i.executeSwitch(s, env)

	case ValidationStatement:
		return i.executeValidation(s, env)

	case ReassignStatement:
		return i.executeReassign(s, env)

	case YieldStatement:
		return i.executeYield(s, env)

	case AssertStatement:
		return i.executeAssert(s, env)

	case IndexAssignStatement:
		return i.executeIndexAssign(s, env)

	case ExpressionStatement:
		return i.EvaluateExpression(s.Expr, env)

	case BreakStatement:
		return nil, &breakValue{}

	case ContinueStatement:
		return nil, &continueValue{}

	default:
		return nil, fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// executeStatements executes a list of statements
func (i *Interpreter) executeStatements(stmts []Statement, env *Environment) (interface{}, error) {
	var result interface{}
	var err error

	for _, stmt := range stmts {
		result, err = i.ExecuteStatement(stmt, env)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

// executeAssign executes a variable assignment
func (i *Interpreter) executeAssign(stmt AssignStatement, env *Environment) (interface{}, error) {
	// Handle dot-notation field assignment (e.g., obj.field = value)
	if parts := strings.SplitN(stmt.Target, ".", 2); len(parts) == 2 {
		return i.executeFieldAssign(parts[0], parts[1], stmt.Value, env)
	}

	// Check for redeclaration in current scope (issue #70)
	// Variables declared with $ cannot be redeclared in the same scope
	if env.HasLocal(stmt.Target) {
		return nil, fmt.Errorf("cannot redeclare variable '%s' in the same scope", stmt.Target)
	}

	value, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}

	// If variable exists in any scope (including parent), update it
	// Otherwise, define a new variable in current scope
	if env.Has(stmt.Target) {
		env.Set(stmt.Target, value)
	} else {
		env.Define(stmt.Target, value)
	}
	return value, nil
}

// executeFieldAssign handles dot-notation field assignment like obj.field = value
func (i *Interpreter) executeFieldAssign(objName, fieldPath string, valueExpr Expr, env *Environment) (interface{}, error) {
	objVal, err := env.Get(objName)
	if err != nil {
		return nil, fmt.Errorf("cannot assign to field of undeclared variable '%s'", objName)
	}

	obj, ok := objVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot assign to field of non-object variable '%s' (type %T)", objName, objVal)
	}

	value, err := i.EvaluateExpression(valueExpr, env)
	if err != nil {
		return nil, err
	}

	// Handle nested field access (e.g., obj.a.b)
	parts := strings.Split(fieldPath, ".")
	current := obj
	for _, part := range parts[:len(parts)-1] {
		next, exists := current[part]
		if !exists {
			return nil, fmt.Errorf("field '%s' does not exist on object", part)
		}
		nextObj, ok := next.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot access field on non-object value at '%s'", part)
		}
		current = nextObj
	}

	current[parts[len(parts)-1]] = value
	return value, nil
}

// executeReassign executes a variable reassignment (without $ prefix)
func (i *Interpreter) executeReassign(stmt ReassignStatement, env *Environment) (interface{}, error) {
	// Handle dot-notation field reassignment (e.g., obj.field = value)
	if parts := strings.SplitN(stmt.Target, ".", 2); len(parts) == 2 {
		return i.executeFieldAssign(parts[0], parts[1], stmt.Value, env)
	}

	// Check that the variable exists (must be previously declared)
	if !env.Has(stmt.Target) {
		return nil, fmt.Errorf("cannot assign to undeclared variable '%s'", stmt.Target)
	}

	// Check if target is a constant (immutable)
	if i.IsConstant(stmt.Target) {
		return nil, fmt.Errorf("cannot reassign constant '%s'", stmt.Target)
	}

	value, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}

	// Update the existing variable
	env.Set(stmt.Target, value)
	return value, nil
}

// executeDbQuery executes a database query (mocked for now)
func (i *Interpreter) executeDbQuery(stmt DbQueryStatement, env *Environment) (interface{}, error) {
	// Evaluate parameters
	params := make([]interface{}, len(stmt.Params))
	for idx, param := range stmt.Params {
		val, err := i.EvaluateExpression(param, env)
		if err != nil {
			return nil, err
		}
		params[idx] = val
	}

	// Mock database query result
	// In a real implementation, this would execute the query against a database
	result := map[string]interface{}{
		"query":  stmt.Query,
		"params": params,
		"result": "mock_db_result",
	}

	env.Define(stmt.Var, result)
	return result, nil
}

// executeReturn executes a return statement
func (i *Interpreter) executeReturn(stmt ReturnStatement, env *Environment) (interface{}, error) {
	value, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}

	// Return a special error to signal a return statement
	return value, &returnValue{value: value}
}

// executeIf executes an if statement
func (i *Interpreter) executeIf(stmt IfStatement, env *Environment) (interface{}, error) {
	condition, err := i.EvaluateExpression(stmt.Condition, env)
	if err != nil {
		return nil, err
	}

	// Check if condition is a boolean
	condBool, ok := condition.(bool)
	if !ok {
		return nil, fmt.Errorf("if condition must be a boolean, got %T", condition)
	}

	// Create a new environment for the if block
	blockEnv := NewChildEnvironment(env)

	if condBool {
		return i.executeStatements(stmt.ThenBlock, blockEnv)
	} else if stmt.ElseBlock != nil {
		return i.executeStatements(stmt.ElseBlock, blockEnv)
	}

	return nil, nil
}

// maxWhileIterations is the safety limit for while loops to prevent infinite loops.
const maxWhileIterations = 1_000_000

// executeWhile executes a while loop
func (i *Interpreter) executeWhile(stmt WhileStatement, env *Environment) (interface{}, error) {
	var result interface{}

	// Execute the loop until the condition is false
	for iterations := 0; ; iterations++ {
		if iterations >= maxWhileIterations {
			return nil, fmt.Errorf("while loop exceeded maximum iterations (%d)", maxWhileIterations)
		}

		// Evaluate condition in parent environment (can access loop variables from previous iterations)
		condition, err := i.EvaluateExpression(stmt.Condition, env)
		if err != nil {
			return nil, err
		}

		// Check if condition is a boolean
		condBool, ok := condition.(bool)
		if !ok {
			return nil, fmt.Errorf("while condition must be a boolean, got %T", condition)
		}

		// Exit loop if condition is false
		if !condBool {
			break
		}

		// Create a fresh environment for each iteration
		loopEnv := NewChildEnvironment(env)

		// Execute loop body
		result, err = i.executeStatements(stmt.Body, loopEnv)
		if err != nil {
			if _, isBreak := err.(*breakValue); isBreak {
				break
			}
			if _, isContinue := err.(*continueValue); isContinue {
				continue
			}
			// Check if it's a return statement
			if _, isReturn := err.(*returnValue); isReturn {
				return result, err
			}
			return nil, err
		}
	}

	return result, nil
}

// executeFor executes a for loop
func (i *Interpreter) executeFor(stmt ForStatement, env *Environment) (interface{}, error) {
	// Evaluate the iterable expression
	iterable, err := i.EvaluateExpression(stmt.Iterable, env)
	if err != nil {
		return nil, err
	}

	var result interface{}

	// Check if iterable is an array (slice)
	if arr, ok := iterable.([]interface{}); ok {
		// Iterate over array
		for index, element := range arr {
			// Create a fresh environment for each iteration
			loopEnv := NewChildEnvironment(env)

			// Define the loop variables for this iteration
			if stmt.KeyVar != "" {
				loopEnv.Define(stmt.KeyVar, int64(index))
			}
			loopEnv.Define(stmt.ValueVar, element)

			// Execute loop body
			result, err = i.executeStatements(stmt.Body, loopEnv)
			if err != nil {
				if _, isBreak := err.(*breakValue); isBreak {
					return result, nil
				}
				if _, isContinue := err.(*continueValue); isContinue {
					continue
				}
				// Check if it's a return statement
				if _, isReturn := err.(*returnValue); isReturn {
					return result, err
				}
				return nil, err
			}
		}
	} else if obj, ok := iterable.(map[string]interface{}); ok {
		// Iterate over object/map
		for key, value := range obj {
			// Create a fresh environment for each iteration
			loopEnv := NewChildEnvironment(env)

			// Define the loop variables for this iteration
			if stmt.KeyVar != "" {
				loopEnv.Define(stmt.KeyVar, key)
				loopEnv.Define(stmt.ValueVar, value)
			} else {
				// If no key variable, iterate over values
				loopEnv.Define(stmt.ValueVar, value)
			}

			// Execute loop body
			result, err = i.executeStatements(stmt.Body, loopEnv)
			if err != nil {
				if _, isBreak := err.(*breakValue); isBreak {
					return result, nil
				}
				if _, isContinue := err.(*continueValue); isContinue {
					continue
				}
				// Check if it's a return statement
				if _, isReturn := err.(*returnValue); isReturn {
					return result, err
				}
				return nil, err
			}
		}
	} else {
		return nil, fmt.Errorf("for loop iterable must be an array or object, got %T", iterable)
	}

	return result, nil
}

// executeSwitch executes a switch statement
func (i *Interpreter) executeSwitch(stmt SwitchStatement, env *Environment) (interface{}, error) {
	// Evaluate the switch value
	switchValue, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}

	// Try to match each case
	for _, caseClause := range stmt.Cases {
		// Evaluate the case value
		caseValue, err := i.EvaluateExpression(caseClause.Value, env)
		if err != nil {
			return nil, err
		}

		// Check if values match
		if i.valuesEqual(switchValue, caseValue) {
			// Create a new environment for the case block
			caseEnv := NewChildEnvironment(env)

			// Execute the case body
			result, err := i.executeStatements(caseClause.Body, caseEnv)
			if err != nil {
				return result, err
			}

			// In GLYPH, switch cases don't fall through by default
			return result, nil
		}
	}

	// If no case matched and there's a default block, execute it
	if stmt.Default != nil && len(stmt.Default) > 0 {
		defaultEnv := NewChildEnvironment(env)
		return i.executeStatements(stmt.Default, defaultEnv)
	}

	return nil, nil
}

// valuesEqual compares two values for equality
func (i *Interpreter) valuesEqual(a, b interface{}) bool {
	// Handle nil values
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare based on type
	switch aVal := a.(type) {
	case int64:
		if bVal, ok := b.(int64); ok {
			return aVal == bVal
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal == bVal
		}
		// Allow comparison between int64 and float64
		if bVal, ok := b.(int64); ok {
			return aVal == float64(bVal)
		}
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	}

	// For int64/float64 cross-comparison
	if aVal, ok := a.(int64); ok {
		if bVal, ok := b.(float64); ok {
			return float64(aVal) == bVal
		}
	}

	return false
}

// executeValidation executes a validation statement
func (i *Interpreter) executeValidation(stmt ValidationStatement, env *Environment) (interface{}, error) {
	// Evaluate the validation function call
	result, err := i.evaluateFunctionCall(stmt.Call, env)
	if err != nil {
		// If the validation function itself errors, return a validation error
		return nil, &ValidationError{Message: err.Error()}
	}

	// Check if the result is a boolean
	if boolResult, ok := result.(bool); ok {
		if !boolResult {
			// Validation failed - return a validation error
			return nil, &ValidationError{Message: fmt.Sprintf("validation failed: %s", stmt.Call.Name)}
		}
		// Validation passed, return nil (no error)
		return nil, nil
	}

	// Check if the result is an error (validation functions can return errors)
	if err, ok := result.(error); ok {
		if err != nil {
			return nil, &ValidationError{Message: err.Error()}
		}
		// No error means validation passed
		return nil, nil
	}

	// If result is nil, treat as successful validation
	if result == nil {
		return nil, nil
	}

	// Unexpected return type
	return nil, &ValidationError{Message: fmt.Sprintf("validation function %s returned unexpected type %T", stmt.Call.Name, result)}
}

// SSEWriter is the interface that yield statements use to send events.
// It is injected into the environment as "__sse_writer" for SSE routes.
type SSEWriter interface {
	SendEvent(data interface{}, eventType string) error
}

// executeYield handles a yield statement by sending an SSE event.
func (i *Interpreter) executeYield(stmt YieldStatement, env *Environment) (interface{}, error) {
	value, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate yield expression: %w", err)
	}

	writer, err := env.Get("__sse_writer")
	if err != nil {
		return nil, fmt.Errorf("yield can only be used inside an SSE route")
	}

	sseWriter, ok := writer.(SSEWriter)
	if !ok {
		return nil, fmt.Errorf("invalid SSE writer in environment")
	}

	if err := sseWriter.SendEvent(value, stmt.EventType); err != nil {
		return nil, fmt.Errorf("failed to send SSE event: %w", err)
	}

	return nil, nil
}

// executeAssert executes an assertion statement
func (i *Interpreter) executeAssert(stmt AssertStatement, env *Environment) (interface{}, error) {
	result, err := i.EvaluateExpression(stmt.Condition, env)
	if err != nil {
		return nil, &AssertionError{
			Message: fmt.Sprintf("assertion error: %v", err),
		}
	}

	boolResult, ok := result.(bool)
	if !ok {
		return nil, &AssertionError{
			Message: fmt.Sprintf("assertion condition must evaluate to boolean, got %T", result),
		}
	}

	if !boolResult {
		message := "assertion failed"
		if stmt.Message != nil {
			msg, err := i.EvaluateExpression(stmt.Message, env)
			if err == nil {
				if msgStr, ok := msg.(string); ok {
					message = msgStr
				}
			}
		}
		return nil, &AssertionError{Message: message}
	}

	return true, nil
}

// executeIndexAssign handles assignment to indexed targets: arr[0] = value, obj.field[0] = value
func (i *Interpreter) executeIndexAssign(stmt IndexAssignStatement, env *Environment) (interface{}, error) {
	value, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}
	return i.assignToTarget(stmt.Target, value, env)
}

// assignToTarget recursively resolves the l-value target and performs the mutation
func (i *Interpreter) assignToTarget(target Expr, value interface{}, env *Environment) (interface{}, error) {
	switch t := target.(type) {
	case ArrayIndexExpr:
		container, err := i.EvaluateExpression(t.Array, env)
		if err != nil {
			return nil, err
		}
		indexVal, err := i.EvaluateExpression(t.Index, env)
		if err != nil {
			return nil, err
		}

		switch c := container.(type) {
		case []interface{}:
			var index int64
			switch idx := indexVal.(type) {
			case int64:
				index = idx
			case int:
				index = int64(idx)
			default:
				return nil, fmt.Errorf("array index must be an integer, got %T", indexVal)
			}
			if index < 0 || int(index) >= len(c) {
				return nil, fmt.Errorf("array index out of bounds: %d (length: %d)", index, len(c))
			}
			c[index] = value
			return value, nil

		case map[string]interface{}:
			keyStr, ok := indexVal.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be a string, got %T", indexVal)
			}
			c[keyStr] = value
			return value, nil

		default:
			return nil, fmt.Errorf("cannot index-assign to %T", container)
		}

	case FieldAccessExpr:
		obj, err := i.EvaluateExpression(t.Object, env)
		if err != nil {
			return nil, err
		}
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot assign field '%s' on %T", t.Field, obj)
		}
		objMap[t.Field] = value
		return value, nil

	default:
		return nil, fmt.Errorf("invalid assignment target: %T", target)
	}
}
