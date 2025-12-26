package interpreter_test

import (
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
)

func TestFutureBasic(t *testing.T) {
	t.Run("resolve future", func(t *testing.T) {
		future := interpreter.NewFuture()

		// Resolve in a goroutine
		go func() {
			time.Sleep(10 * time.Millisecond)
			future.Resolve("hello")
		}()

		result, err := future.Await()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}
	})

	t.Run("reject future", func(t *testing.T) {
		future := interpreter.NewFuture()

		// Reject in a goroutine
		go func() {
			time.Sleep(10 * time.Millisecond)
			future.Reject(&interpreter.ValidationError{Message: "test error"})
		}()

		_, err := future.Await()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("future state transitions", func(t *testing.T) {
		future := interpreter.NewFuture()

		if !future.IsPending() {
			t.Error("new future should be pending")
		}
		if future.IsResolved() {
			t.Error("new future should not be resolved")
		}

		future.Resolve(42)

		if future.IsPending() {
			t.Error("resolved future should not be pending")
		}
		if !future.IsResolved() {
			t.Error("resolved future should be resolved")
		}
		if future.Value() != 42 {
			t.Errorf("expected value 42, got %v", future.Value())
		}
	})

	t.Run("future timeout", func(t *testing.T) {
		future := interpreter.NewFuture()

		// Don't resolve - should timeout
		_, err := future.AwaitWithTimeout(50 * time.Millisecond)
		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
	})
}

func TestAsyncExprParsing(t *testing.T) {
	t.Run("parse simple async block", func(t *testing.T) {
		input := `
@ GET /test
  $ result = async {
    $ x = 1
    > x
  }
  > result
`
		lexer := parser.NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("lexer error: %v", err)
		}

		p := parser.NewParser(tokens)
		module, err := p.Parse()
		if err != nil {
			t.Fatalf("parser error: %v", err)
		}

		if len(module.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(module.Items))
		}

		route, ok := module.Items[0].(*interpreter.Route)
		if !ok {
			t.Fatalf("expected Route, got %T", module.Items[0])
		}

		if len(route.Body) != 2 {
			t.Fatalf("expected 2 statements in route body, got %d", len(route.Body))
		}

		// First statement should be an assignment with async expr
		assign, ok := route.Body[0].(interpreter.AssignStatement)
		if !ok {
			t.Fatalf("expected AssignStatement, got %T", route.Body[0])
		}
		if assign.Target != "result" {
			t.Errorf("expected target 'result', got %s", assign.Target)
		}

		_, ok = assign.Value.(interpreter.AsyncExpr)
		if !ok {
			t.Errorf("expected AsyncExpr, got %T", assign.Value)
		}
	})

	t.Run("parse await expression", func(t *testing.T) {
		input := `
@ GET /test
  $ future = async { > 42 }
  $ value = await future
  > value
`
		lexer := parser.NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("lexer error: %v", err)
		}

		p := parser.NewParser(tokens)
		module, err := p.Parse()
		if err != nil {
			t.Fatalf("parser error: %v", err)
		}

		route, ok := module.Items[0].(*interpreter.Route)
		if !ok {
			t.Fatalf("expected Route, got %T", module.Items[0])
		}

		if len(route.Body) != 3 {
			t.Fatalf("expected 3 statements, got %d", len(route.Body))
		}

		// Second statement should be await
		assign, ok := route.Body[1].(interpreter.AssignStatement)
		if !ok {
			t.Fatalf("expected AssignStatement, got %T", route.Body[1])
		}

		awaitExpr, ok := assign.Value.(interpreter.AwaitExpr)
		if !ok {
			t.Fatalf("expected AwaitExpr, got %T", assign.Value)
		}

		varExpr, ok := awaitExpr.Expr.(interpreter.VariableExpr)
		if !ok {
			t.Fatalf("expected VariableExpr in await, got %T", awaitExpr.Expr)
		}
		if varExpr.Name != "future" {
			t.Errorf("expected variable 'future', got %s", varExpr.Name)
		}
	})
}

func TestAsyncExprEvaluation(t *testing.T) {
	t.Run("basic async/await", func(t *testing.T) {
		input := `
@ GET /test
  $ future = async {
    $ x = 10
    $ y = 20
    > x + y
  }
  $ result = await future
  > result
`
		lexer := parser.NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("lexer error: %v", err)
		}

		p := parser.NewParser(tokens)
		module, err := p.Parse()
		if err != nil {
			t.Fatalf("parser error: %v", err)
		}

		interp := interpreter.NewInterpreter()
		if err := interp.LoadModule(*module); err != nil {
			t.Fatalf("load module error: %v", err)
		}

		route, ok := module.Items[0].(*interpreter.Route)
		if !ok {
			t.Fatalf("expected Route, got %T", module.Items[0])
		}

		result, err := interp.ExecuteRouteSimple(route, nil)
		if err != nil {
			t.Fatalf("execution error: %v", err)
		}

		if result != int64(30) {
			t.Errorf("expected 30, got %v (type: %T)", result, result)
		}
	})

	t.Run("concurrent async operations", func(t *testing.T) {
		input := `
@ GET /test
  $ f1 = async { > 100 }
  $ f2 = async { > 200 }
  $ f3 = async { > 300 }
  $ v1 = await f1
  $ v2 = await f2
  $ v3 = await f3
  > v1 + v2 + v3
`
		lexer := parser.NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("lexer error: %v", err)
		}

		p := parser.NewParser(tokens)
		module, err := p.Parse()
		if err != nil {
			t.Fatalf("parser error: %v", err)
		}

		interp := interpreter.NewInterpreter()
		if err := interp.LoadModule(*module); err != nil {
			t.Fatalf("load module error: %v", err)
		}

		route, ok := module.Items[0].(*interpreter.Route)
		if !ok {
			t.Fatalf("expected Route, got %T", module.Items[0])
		}

		result, err := interp.ExecuteRouteSimple(route, nil)
		if err != nil {
			t.Fatalf("execution error: %v", err)
		}

		if result != int64(600) {
			t.Errorf("expected 600, got %v", result)
		}
	})

	t.Run("async with object return", func(t *testing.T) {
		input := `
@ GET /test
  $ future = async {
    > {name: "test", value: 42}
  }
  $ result = await future
  > result
`
		lexer := parser.NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("lexer error: %v", err)
		}

		p := parser.NewParser(tokens)
		module, err := p.Parse()
		if err != nil {
			t.Fatalf("parser error: %v", err)
		}

		interp := interpreter.NewInterpreter()
		if err := interp.LoadModule(*module); err != nil {
			t.Fatalf("load module error: %v", err)
		}

		route, ok := module.Items[0].(*interpreter.Route)
		if !ok {
			t.Fatalf("expected Route, got %T", module.Items[0])
		}

		result, err := interp.ExecuteRouteSimple(route, nil)
		if err != nil {
			t.Fatalf("execution error: %v", err)
		}

		obj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		if obj["name"] != "test" {
			t.Errorf("expected name 'test', got %v", obj["name"])
		}
		if obj["value"] != int64(42) {
			t.Errorf("expected value 42, got %v", obj["value"])
		}
	})
}

func TestFutureHelpers(t *testing.T) {
	t.Run("RunAsync helper", func(t *testing.T) {
		future := interpreter.RunAsync(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return "async result", nil
		})

		result, err := future.Await()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "async result" {
			t.Errorf("expected 'async result', got %v", result)
		}
	})

	t.Run("All futures", func(t *testing.T) {
		f1 := interpreter.NewFuture()
		f2 := interpreter.NewFuture()
		f3 := interpreter.NewFuture()

		go func() {
			f1.Resolve(1)
			f2.Resolve(2)
			f3.Resolve(3)
		}()

		results, err := interpreter.All(f1, f2, f3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		if results[0] != 1 || results[1] != 2 || results[2] != 3 {
			t.Errorf("unexpected results: %v", results)
		}
	})

	t.Run("Race futures", func(t *testing.T) {
		f1 := interpreter.NewFuture()
		f2 := interpreter.NewFuture()

		go func() {
			time.Sleep(50 * time.Millisecond)
			f1.Resolve("slow")
		}()

		go func() {
			time.Sleep(10 * time.Millisecond)
			f2.Resolve("fast")
		}()

		result, err := interpreter.Race(f1, f2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "fast" {
			t.Errorf("expected 'fast', got %v", result)
		}
	})

	t.Run("IsFuture check", func(t *testing.T) {
		future := interpreter.NewFuture()

		if !interpreter.IsFuture(future) {
			t.Error("expected IsFuture to return true for Future")
		}

		if interpreter.IsFuture("not a future") {
			t.Error("expected IsFuture to return false for string")
		}
	})
}
