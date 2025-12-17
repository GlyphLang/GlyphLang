package interpreter

import "fmt"

// returnValue is a special error type to handle return statements
type returnValue struct {
	value interface{}
}

func (r *returnValue) Error() string {
	return "return"
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
	value, err := i.EvaluateExpression(stmt.Value, env)
	if err != nil {
		return nil, err
	}

	// Try to set existing variable first, otherwise define new one in current scope
	err = env.Set(stmt.Target, value)
	if err != nil {
		// Variable doesn't exist, define it in current scope
		env.Define(stmt.Target, value)
	}
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

// executeWhile executes a while loop
func (i *Interpreter) executeWhile(stmt WhileStatement, env *Environment) (interface{}, error) {
	var result interface{}

	// Create a new environment for the loop
	loopEnv := NewChildEnvironment(env)

	// Execute the loop until the condition is false
	for {
		condition, err := i.EvaluateExpression(stmt.Condition, loopEnv)
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

		// Execute loop body
		result, err = i.executeStatements(stmt.Body, loopEnv)
		if err != nil {
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
		// Create a single child environment for the loop
		loopEnv := NewChildEnvironment(env)
		
		// Iterate over array
		for index, element := range arr {
			// Update the loop variables for this iteration
			if stmt.KeyVar != "" {
				loopEnv.Define(stmt.KeyVar, int64(index))
			}
			loopEnv.Define(stmt.ValueVar, element)

			// Execute loop body
			result, err = i.executeStatements(stmt.Body, loopEnv)
			if err != nil {
				// Check if it's a return statement
				if _, isReturn := err.(*returnValue); isReturn {
					return result, err
				}
				return nil, err
			}
		}
	} else if obj, ok := iterable.(map[string]interface{}); ok {
		// Create a single child environment for the loop
		loopEnv := NewChildEnvironment(env)
		
		// Iterate over object/map
		for key, value := range obj {
			// Update the loop variables for this iteration
			if stmt.KeyVar != "" {
				loopEnv.Define(stmt.KeyVar, key)
			} else {
				// If no key variable, use the value variable for the key
				// This allows "for key in object" syntax
				loopEnv.Define(stmt.ValueVar, key)
				continue
			}
			loopEnv.Define(stmt.ValueVar, value)

			// Execute loop body
			result, err = i.executeStatements(stmt.Body, loopEnv)
			if err != nil {
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
