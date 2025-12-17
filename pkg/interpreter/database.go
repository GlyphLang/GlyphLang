package interpreter

import (
	"fmt"
	"reflect"
)

// CallMethod calls a method on an object using reflection
func CallMethod(obj interface{}, methodName string, args ...interface{}) (interface{}, error) {
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
