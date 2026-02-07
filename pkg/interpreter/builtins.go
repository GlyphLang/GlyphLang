package interpreter

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/google/uuid"
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
		"randomInt":  builtinRandomInt,
		"generateId": builtinGenerateId,
		"append":     builtinAppend,
		"set":        builtinSet,
		"remove":     builtinRemove,
		"keys":       builtinKeys,
		"map":        builtinMap,
		"filter":     builtinFilter,
		"reduce":     builtinReduce,
		"find":       builtinFind,
		"some":       builtinSome,
		"every":      builtinEvery,
		"sort":       builtinSort,
		"reverse":    builtinReverse,
		"flat":       builtinFlat,
		"slice":      builtinSlice,
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
	runes := []rune(str)
	if int(end) > len(runes) {
		end = int64(len(runes))
	}
	if int(start) > len(runes) {
		start = int64(len(runes))
	}
	return string(runes[start:end]), nil
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
		return int64(len([]rune(v))), nil
	case []interface{}:
		return int64(len(v)), nil
	case map[string]interface{}:
		return int64(len(v)), nil
	default:
		return nil, fmt.Errorf("length() expects a string, array, or object argument, got %T", arg)
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
	byteIndex := strings.Index(str, substr)
	if byteIndex < 0 {
		return int64(-1), nil
	}
	// Convert byte offset to rune offset for Unicode consistency
	runeIndex := len([]rune(str[:byteIndex]))
	return int64(runeIndex), nil
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
	runes := []rune(str)
	if index < 0 || int(index) >= len(runes) {
		return nil, fmt.Errorf("charAt() index out of bounds: %d", index)
	}
	return string(runes[index]), nil
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
		if v == math.MinInt64 {
			return nil, fmt.Errorf("abs() overflow: cannot negate minimum int64 value")
		}
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case float64:
		return math.Abs(v), nil
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

func builtinRandomInt(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("randomInt() expects 2 arguments (min, max), got %d", len(args))
	}
	minArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	maxArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	minVal, ok := minArg.(int64)
	if !ok {
		return nil, fmt.Errorf("randomInt() expects integer arguments, got %T for min", minArg)
	}
	maxVal, ok := maxArg.(int64)
	if !ok {
		return nil, fmt.Errorf("randomInt() expects integer arguments, got %T for max", maxArg)
	}
	if minVal > maxVal {
		return nil, fmt.Errorf("randomInt() requires min <= max, got min=%d, max=%d", minVal, maxVal)
	}
	// #nosec G404 -- non-cryptographic PRNG intentional for general-purpose scripting use
	return minVal + rand.Int63n(maxVal-minVal+1), nil
}

func builtinGenerateId(_ *Interpreter, args []Expr, _ *Environment) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("generateId() expects 0 arguments, got %d", len(args))
	}
	return uuid.New().String(), nil
}

func builtinAppend(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("append() expects 2 arguments (array, item), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("append() expects first argument to be an array, got %T", arrArg)
	}
	item, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	return append(arr, item), nil
}

func builtinSet(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("set() expects 3 arguments (object, key, value), got %d", len(args))
	}
	objArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	obj, ok := objArg.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("set() expects first argument to be an object, got %T", objArg)
	}
	keyArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	key, ok := keyArg.(string)
	if !ok {
		return nil, fmt.Errorf("set() expects second argument to be a string key, got %T", keyArg)
	}
	value, err := i.EvaluateExpression(args[2], env)
	if err != nil {
		return nil, err
	}
	obj[key] = value
	return obj, nil
}

func builtinRemove(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("remove() expects 2 arguments (object, key), got %d", len(args))
	}
	objArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	obj, ok := objArg.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("remove() expects first argument to be an object, got %T", objArg)
	}
	keyArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	key, ok := keyArg.(string)
	if !ok {
		return nil, fmt.Errorf("remove() expects second argument to be a string key, got %T", keyArg)
	}
	delete(obj, key)
	return obj, nil
}

func builtinKeys(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("keys() expects 1 argument (object), got %d", len(args))
	}
	objArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	obj, ok := objArg.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("keys() expects an object argument, got %T", objArg)
	}
	keys := make([]interface{}, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	return keys, nil
}

// callCallable invokes a callable (LambdaClosure or Function) with the given arguments.
func (i *Interpreter) callCallable(fn interface{}, args []interface{}) (interface{}, error) {
	switch f := fn.(type) {
	case *LambdaClosure:
		return i.callLambdaClosure(f, args)
	case Function:
		fnEnv := NewChildEnvironment(NewEnvironment())
		for idx, param := range f.Params {
			if idx < len(args) {
				fnEnv.Define(param.Name, args[idx])
			}
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
		return i.callCallable(*f, args)
	default:
		return nil, fmt.Errorf("expected a function, got %T", fn)
	}
}

func builtinMap(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("map() expects 2 arguments (array, function), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("map() expects first argument to be an array, got %T", arrArg)
	}
	fnArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(arr))
	for idx, elem := range arr {
		val, err := i.callCallable(fnArg, []interface{}{elem})
		if err != nil {
			return nil, fmt.Errorf("map() callback error at index %d: %v", idx, err)
		}
		result[idx] = val
	}
	return result, nil
}

func builtinFilter(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("filter() expects 2 arguments (array, function), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter() expects first argument to be an array, got %T", arrArg)
	}
	fnArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0)
	for idx, elem := range arr {
		val, err := i.callCallable(fnArg, []interface{}{elem})
		if err != nil {
			return nil, fmt.Errorf("filter() callback error at index %d: %v", idx, err)
		}
		if truthy, ok := val.(bool); ok && truthy {
			result = append(result, elem)
		}
	}
	return result, nil
}

func builtinReduce(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("reduce() expects 3 arguments (array, function, initial), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("reduce() expects first argument to be an array, got %T", arrArg)
	}
	fnArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	acc, err := i.EvaluateExpression(args[2], env)
	if err != nil {
		return nil, err
	}
	for idx, elem := range arr {
		acc, err = i.callCallable(fnArg, []interface{}{acc, elem})
		if err != nil {
			return nil, fmt.Errorf("reduce() callback error at index %d: %v", idx, err)
		}
	}
	return acc, nil
}

func builtinFind(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("find() expects 2 arguments (array, function), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("find() expects first argument to be an array, got %T", arrArg)
	}
	fnArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	for idx, elem := range arr {
		val, err := i.callCallable(fnArg, []interface{}{elem})
		if err != nil {
			return nil, fmt.Errorf("find() callback error at index %d: %v", idx, err)
		}
		if truthy, ok := val.(bool); ok && truthy {
			return elem, nil
		}
	}
	return nil, nil
}

func builtinSome(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("some() expects 2 arguments (array, function), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("some() expects first argument to be an array, got %T", arrArg)
	}
	fnArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	for idx, elem := range arr {
		val, err := i.callCallable(fnArg, []interface{}{elem})
		if err != nil {
			return nil, fmt.Errorf("some() callback error at index %d: %v", idx, err)
		}
		if truthy, ok := val.(bool); ok && truthy {
			return true, nil
		}
	}
	return false, nil
}

func builtinEvery(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("every() expects 2 arguments (array, function), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("every() expects first argument to be an array, got %T", arrArg)
	}
	fnArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	for idx, elem := range arr {
		val, err := i.callCallable(fnArg, []interface{}{elem})
		if err != nil {
			return nil, fmt.Errorf("every() callback error at index %d: %v", idx, err)
		}
		if truthy, ok := val.(bool); !ok || !truthy {
			return false, nil
		}
	}
	return true, nil
}

func builtinSort(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("sort() expects 1-2 arguments (array[, comparator]), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("sort() expects first argument to be an array, got %T", arrArg)
	}
	result := make([]interface{}, len(arr))
	copy(result, arr)

	if len(args) == 2 {
		fnArg, err := i.EvaluateExpression(args[1], env)
		if err != nil {
			return nil, err
		}
		var sortErr error
		sort.SliceStable(result, func(a, b int) bool {
			if sortErr != nil {
				return false
			}
			val, err := i.callCallable(fnArg, []interface{}{result[a], result[b]})
			if err != nil {
				sortErr = err
				return false
			}
			switch v := val.(type) {
			case int64:
				return v < 0
			case float64:
				return v < 0
			case bool:
				return v
			default:
				sortErr = fmt.Errorf("sort() comparator must return a number or boolean, got %T", val)
				return false
			}
		})
		if sortErr != nil {
			return nil, sortErr
		}
	} else {
		var sortErr error
		sort.SliceStable(result, func(a, b int) bool {
			if sortErr != nil {
				return false
			}
			return defaultLess(result[a], result[b], &sortErr)
		})
		if sortErr != nil {
			return nil, sortErr
		}
	}
	return result, nil
}

func defaultLess(a, b interface{}, errOut *error) bool {
	switch av := a.(type) {
	case int64:
		if bv, ok := b.(int64); ok {
			return av < bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av < bv
		}
	case string:
		if bv, ok := b.(string); ok {
			return av < bv
		}
	}
	*errOut = fmt.Errorf("sort() cannot compare %T and %T", a, b)
	return false
}

func builtinReverse(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("reverse() expects 1 argument (array), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("reverse() expects an array argument, got %T", arrArg)
	}
	result := make([]interface{}, len(arr))
	for idx, elem := range arr {
		result[len(arr)-1-idx] = elem
	}
	return result, nil
}

func builtinFlat(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("flat() expects 1 argument (array), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("flat() expects an array argument, got %T", arrArg)
	}
	result := make([]interface{}, 0)
	for _, elem := range arr {
		if inner, ok := elem.([]interface{}); ok {
			result = append(result, inner...)
		} else {
			result = append(result, elem)
		}
	}
	return result, nil
}

func builtinSlice(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("slice() expects 3 arguments (array, start, end), got %d", len(args))
	}
	arrArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}
	arr, ok := arrArg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("slice() expects first argument to be an array, got %T", arrArg)
	}
	startArg, err := i.EvaluateExpression(args[1], env)
	if err != nil {
		return nil, err
	}
	start, ok := startArg.(int64)
	if !ok {
		return nil, fmt.Errorf("slice() expects second argument to be an integer, got %T", startArg)
	}
	endArg, err := i.EvaluateExpression(args[2], env)
	if err != nil {
		return nil, err
	}
	end, ok := endArg.(int64)
	if !ok {
		return nil, fmt.Errorf("slice() expects third argument to be an integer, got %T", endArg)
	}
	if start < 0 {
		start = 0
	}
	if end > int64(len(arr)) {
		end = int64(len(arr))
	}
	if start > end {
		return make([]interface{}, 0), nil
	}
	result := make([]interface{}, end-start)
	copy(result, arr[start:end])
	return result, nil
}
