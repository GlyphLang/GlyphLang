package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"fmt"
	"strings"
	"time"
)

// builtinFunc is the signature for all builtin function implementations.
type builtinFunc func(i *Interpreter, args []Expr, env *Environment) (interface{}, error)

// builtinFuncs is the dispatch table mapping builtin function names to their implementations.
// Initialized in init() to avoid initialization cycle with evaluateFunctionCall.
var builtinFuncs map[string]builtinFunc

func init() {
	builtinFuncs = map[string]builtinFunc{
		"time.now":   builtinTimeNow,
		"now":        builtinNow,
		"Ok":         builtinOk,
		"Err":        builtinErr,
		"upper":      builtinUpper,
		"lower":      builtinLower,
		"trim":       builtinTrim,
		"split":      builtinSplit,
		"join":       builtinJoin,
		"contains":   builtinContains,
		"replace":    builtinReplace,
		"substring":  builtinSubstring,
		"length":     builtinLength,
		"startsWith": builtinStartsWith,
		"endsWith":   builtinEndsWith,
		"indexOf":    builtinIndexOf,
		"charAt":     builtinCharAt,
		"parseInt":   builtinParseInt,
		"parseFloat": builtinParseFloat,
		"toString":   builtinToString,
		"abs":        builtinAbs,
		"min":        builtinMin,
		"max":        builtinMax,
	}
}

func builtinTimeNow(_ *Interpreter, _ []Expr, _ *Environment) (interface{}, error) {
	return time.Now().Unix(), nil
}

func builtinNow(_ *Interpreter, _ []Expr, _ *Environment) (interface{}, error) {
	return time.Now().Unix(), nil
}

func builtinOk(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Ok() expects 1 argument, got %d", len(args))
	}
	val, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	return NewOk(val), nil
}

func builtinErr(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Err() expects 1 argument, got %d", len(args))
	}
	val, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	return NewErr(val), nil
}

func builtinUpper(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Convert string to uppercase
	if len(args) != 1 {
		return nil, fmt.Errorf("upper() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := arg.(string)
	if !ok {
		return nil, fmt.Errorf("upper() expects a string argument, got %T", arg)
	}
	return strings.ToUpper(str), nil
}

func builtinLower(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Convert string to lowercase
	if len(args) != 1 {
		return nil, fmt.Errorf("lower() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := arg.(string)
	if !ok {
		return nil, fmt.Errorf("lower() expects a string argument, got %T", arg)
	}
	return strings.ToLower(str), nil
}

func builtinTrim(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Remove leading/trailing whitespace
	if len(args) != 1 {
		return nil, fmt.Errorf("trim() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := arg.(string)
	if !ok {
		return nil, fmt.Errorf("trim() expects a string argument, got %T", arg)
	}
	return strings.TrimSpace(str), nil
}

func builtinSplit(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Split string into array
	if len(args) != 2 {
		return nil, fmt.Errorf("split() expects 2 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("split() expects first argument to be a string, got %T", strArg)
	}
	delimArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	delim, ok := delimArg.(string)
	if !ok {
		return nil, fmt.Errorf("split() expects second argument to be a string, got %T", delimArg)
	}
	parts := strings.Split(str, delim)
	result := make([]interface{}, len(parts))
	for idx, part := range parts {
		result[idx] = part
	}
	return result, nil
}

func builtinJoin(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Join array into string
	if len(args) != 2 {
		return nil, fmt.Errorf("join() expects 2 arguments, got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("join() expects first argument to be an array, got %T", arrArg)
	}
	delimArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	delim, ok := delimArg.(string)
	if !ok {
		return nil, fmt.Errorf("join() expects second argument to be a string, got %T", delimArg)
	}
	strParts := make([]string, len(arr))
	for idx, elem := range arr {
		strParts[idx] = fmt.Sprintf("%v", elem)
	}
	return strings.Join(strParts, delim), nil
}

func builtinContains(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Check if string contains substring
	if len(args) != 2 {
		return nil, fmt.Errorf("contains() expects 2 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("contains() expects first argument to be a string, got %T", strArg)
	}
	substrArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	substr, ok := substrArg.(string)
	if !ok {
		return nil, fmt.Errorf("contains() expects second argument to be a string, got %T", substrArg)
	}
	return strings.Contains(str, substr), nil
}

func builtinReplace(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Replace occurrences in string
	if len(args) != 3 {
		return nil, fmt.Errorf("replace() expects 3 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("replace() expects first argument to be a string, got %T", strArg)
	}
	oldArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	old, ok := oldArg.(string)
	if !ok {
		return nil, fmt.Errorf("replace() expects second argument to be a string, got %T", oldArg)
	}
	newArg, err := i.EvaluateExpression(args[2], env)
	if err != nil {
		return nil, err
	}
	new, ok := newArg.(string)
	if !ok {
		return nil, fmt.Errorf("replace() expects third argument to be a string, got %T", newArg)
	}
	return strings.ReplaceAll(str, old, new), nil
}

func builtinSubstring(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Get substring
	if len(args) != 3 {
		return nil, fmt.Errorf("substring() expects 3 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("substring() expects first argument to be a string, got %T", strArg)
	}
	startArg, err := i.EvaluateExpression(args[1], env)
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
	endArg, err := i.EvaluateExpression(args[2], env)
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
}

func builtinLength(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Get length of string or array
	if len(args) != 1 {
		return nil, fmt.Errorf("length() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
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
}

func builtinStartsWith(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Check if string starts with prefix
	if len(args) != 2 {
		return nil, fmt.Errorf("startsWith() expects 2 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("startsWith() expects first argument to be a string, got %T", strArg)
	}
	prefixArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	prefix, ok := prefixArg.(string)
	if !ok {
		return nil, fmt.Errorf("startsWith() expects second argument to be a string, got %T", prefixArg)
	}
	return strings.HasPrefix(str, prefix), nil
}

func builtinEndsWith(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Check if string ends with suffix
	if len(args) != 2 {
		return nil, fmt.Errorf("endsWith() expects 2 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("endsWith() expects first argument to be a string, got %T", strArg)
	}
	suffixArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	suffix, ok := suffixArg.(string)
	if !ok {
		return nil, fmt.Errorf("endsWith() expects second argument to be a string, got %T", suffixArg)
	}
	return strings.HasSuffix(str, suffix), nil
}

func builtinIndexOf(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Find first occurrence of substring
	if len(args) != 2 {
		return nil, fmt.Errorf("indexOf() expects 2 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("indexOf() expects first argument to be a string, got %T", strArg)
	}
	substrArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	substr, ok := substrArg.(string)
	if !ok {
		return nil, fmt.Errorf("indexOf() expects second argument to be a string, got %T", substrArg)
	}
	return int64(strings.Index(str, substr)), nil
}

func builtinCharAt(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Get character at index
	if len(args) != 2 {
		return nil, fmt.Errorf("charAt() expects 2 arguments, got %d", len(args))
	}
	strArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	str, ok := strArg.(string)
	if !ok {
		return nil, fmt.Errorf("charAt() expects first argument to be a string, got %T", strArg)
	}
	indexArg, err := i.EvaluateExpression(args[1], env)
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
}

func builtinParseInt(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Parse string to integer
	if len(args) != 1 {
		return nil, fmt.Errorf("parseInt() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
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
}

func builtinParseFloat(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Parse string to float
	if len(args) != 1 {
		return nil, fmt.Errorf("parseFloat() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
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
}

func builtinToString(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Convert value to string
	if len(args) != 1 {
		return nil, fmt.Errorf("toString() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%v", arg), nil
}

func builtinAbs(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Absolute value
	if len(args) != 1 {
		return nil, fmt.Errorf("abs() expects 1 argument, got %d", len(args))
	}
	arg, err := i.EvaluateExpression(args[0], env)
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
}

func builtinMin(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Minimum of two values
	if len(args) != 2 {
		return nil, fmt.Errorf("min() expects 2 arguments, got %d", len(args))
	}
	leftArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	rightArg, err := i.EvaluateExpression(args[1], env)
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
}

func builtinMax(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	// Maximum of two values
	if len(args) != 2 {
		return nil, fmt.Errorf("max() expects 2 arguments, got %d", len(args))
	}
	leftArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	rightArg, err := i.EvaluateExpression(args[1], env)
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
