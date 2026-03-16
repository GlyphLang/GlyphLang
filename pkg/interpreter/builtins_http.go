package interpreter

import (
	"fmt"

	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/httpclient"
)

// defaultHTTPHandler is a shared HTTP client handler for built-in http.* functions.
// It is lazily initialized on first use so there is no cost when HTTP builtins are unused.
var defaultHTTPHandler *httpclient.Handler

func getDefaultHTTPHandler() *httpclient.Handler {
	if defaultHTTPHandler == nil {
		defaultHTTPHandler = httpclient.NewHandler()
	}
	return defaultHTTPHandler
}

func init() {
	builtinFuncs["http.get"] = builtinHTTPGet
	builtinFuncs["http.post"] = builtinHTTPPost
	builtinFuncs["http.put"] = builtinHTTPPut
	builtinFuncs["http.patch"] = builtinHTTPPatch
	builtinFuncs["http.delete"] = builtinHTTPDelete
}

func builtinHTTPGet(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	return callHTTPBuiltin(i, "Get", args, env)
}

func builtinHTTPPost(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	return callHTTPBuiltin(i, "Post", args, env)
}

func builtinHTTPPut(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	return callHTTPBuiltin(i, "Put", args, env)
}

func builtinHTTPPatch(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	return callHTTPBuiltin(i, "Patch", args, env)
}

func builtinHTTPDelete(i *Interpreter, args []Expr, env *Environment) (interface{}, error) {
	return callHTTPBuiltin(i, "Delete", args, env)
}

// callHTTPBuiltin evaluates arguments and delegates to the HTTP handler.
// Accepts 1 argument: either a string URL or an options object.
func callHTTPBuiltin(i *Interpreter, method string, args []Expr, env *Environment) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("http.%s() expects 1-2 arguments, got %d", method, len(args))
	}

	// Evaluate the URL argument
	urlArg, err := i.EvaluateExpression(args[0], env)
	if err != nil {
		return nil, err
	}

	var requestArg interface{}

	if len(args) == 2 {
		// Two-arg form: http.post(url, {body: ..., headers: ...})
		optsArg, err := i.EvaluateExpression(args[1], env)
		if err != nil {
			return nil, err
		}
		optsMap, ok := optsArg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("http.%s() second argument must be an object, got %T", method, optsArg)
		}
		urlStr, ok := urlArg.(string)
		if !ok {
			return nil, fmt.Errorf("http.%s() first argument must be a string URL when using two arguments, got %T", method, urlArg)
		}
		optsMap["url"] = urlStr
		requestArg = optsMap
	} else {
		// Single-arg form: http.get(url) or http.get({url: ..., headers: ...})
		requestArg = urlArg
	}

	h := getDefaultHTTPHandler()
	return CallMethod(h, method, requestArg)
}
