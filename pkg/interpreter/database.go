package interpreter

import (
	"fmt"
	"reflect"
)

// allowedMethods is a whitelist of safe methods that can be called via reflection
var allowedMethods = map[string]bool{
	// Database/ORM methods
	"Get":       true,
	"Find":      true,
	"Create":    true,
	"Update":    true,
	"Delete":    true,
	"Query":     true,
	"First":     true,
	"All":       true,
	"Where":     true,
	"Count":     true,
	"Save":      true,
	"Insert":    true,
	"Select":    true,
	"Limit":     true,
	"Offset":    true,
	"Order":     true,
	"Filter":     true,
	"Table":      true,
	"CountWhere": true,
	"NextId":     true,
	"Length":     true,
	// Common safe methods
	"String": true,
	"Int":    true,
	"Bool":   true,
	"Float":  true,
	"Len":    true,
	"IsZero": true,
}

// CallMethod calls a method on an object using reflection
// Only methods in the allowedMethods whitelist can be called for security
func CallMethod(obj interface{}, methodName string, args ...interface{}) (interface{}, error) {
	// Check if the method is in the whitelist
	if !allowedMethods[methodName] {
		return nil, fmt.Errorf("method %s is not allowed", methodName)
	}

	// Get the value and type of the object
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type()

	// Find the method
	method := objValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found on type %s", methodName, objType)
	}

	// Prepare arguments
	methodArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		methodArgs[i] = reflect.ValueOf(arg)
	}

	// Call the method
	results := method.Call(methodArgs)

	// Handle return values
	if len(results) == 0 {
		return nil, nil
	}

	// If the last result is an error, check it
	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !lastResult.IsNil() {
			return nil, lastResult.Interface().(error)
		}
		// If there are other results, return the first one
		if len(results) > 1 {
			return results[0].Interface(), nil
		}
		return nil, nil
	}

	// Return the first result
	return results[0].Interface(), nil
}

// HasMethod checks if an object has a method
func HasMethod(obj interface{}, methodName string) bool {
	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(methodName)
	return method.IsValid()
}

// GetMethodNames returns all method names of an object
func GetMethodNames(obj interface{}) []string {
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type()

	var methods []string
	for i := 0; i < objType.NumMethod(); i++ {
		methods = append(methods, objType.Method(i).Name)
	}

	return methods
}
