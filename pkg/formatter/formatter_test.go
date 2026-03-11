package formatter

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
	"testing"
)

func TestFormatRoute(t *testing.T) {
	route := &ast.Route{
		Method: ast.Get,
		Path:   "/api/users",
		Body: []ast.Statement{
			ast.AssignStatement{
				Target: "users",
				Value:  ast.ArrayExpr{Elements: []ast.Expr{}},
			},
			ast.ReturnStatement{
				Value: ast.VariableExpr{Name: "users"},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{route},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "@ GET /api/users") {
		t.Errorf("Compact output should contain '@ GET /api/users', got: %s", compact)
	}
	if !strings.Contains(compact, "$ users = []") {
		t.Errorf("Compact output should contain '$ users = []', got: %s", compact)
	}
	if !strings.Contains(compact, "> users") {
		t.Errorf("Compact output should contain '> users', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "route GET /api/users") {
		t.Errorf("Expanded output should contain 'route GET /api/users', got: %s", expanded)
	}
	if !strings.Contains(expanded, "let users = []") {
		t.Errorf("Expanded output should contain 'let users = []', got: %s", expanded)
	}
	if !strings.Contains(expanded, "return users") {
		t.Errorf("Expanded output should contain 'return users', got: %s", expanded)
	}
}

func TestFormatTypeDef(t *testing.T) {
	typeDef := &ast.TypeDef{
		Name: "User",
		Fields: []ast.Field{
			{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
			{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
			{Name: "email", TypeAnnotation: ast.StringType{}, Required: false},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{typeDef},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, ": User {") {
		t.Errorf("Compact output should contain ': User {', got: %s", compact)
	}
	if !strings.Contains(compact, "id: int!") {
		t.Errorf("Compact output should contain 'id: int!', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "type User {") {
		t.Errorf("Expanded output should contain 'type User {', got: %s", expanded)
	}
}

func TestFormatCommand(t *testing.T) {
	cmd := &ast.Command{
		Name: "hello",
		Params: []ast.CommandParam{
			{Name: "name", Type: ast.StringType{}, Required: true},
		},
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "message", Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "Hello"}}},
					},
				},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{cmd},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "! hello") {
		t.Errorf("Compact output should contain '! hello', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "command hello") {
		t.Errorf("Expanded output should contain 'command hello', got: %s", expanded)
	}
}

func TestFormatCronTask(t *testing.T) {
	cron := &ast.CronTask{
		Schedule: "0 0 * * *",
		Name:     "daily_task",
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "done", Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
					},
				},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{cron},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "* \"0 0 * * *\" daily_task") {
		t.Errorf("Compact output should contain '* \"0 0 * * *\" daily_task', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "cron \"0 0 * * *\" daily_task") {
		t.Errorf("Expanded output should contain 'cron \"0 0 * * *\" daily_task', got: %s", expanded)
	}
}

func TestFormatEventHandler(t *testing.T) {
	event := &ast.EventHandler{
		EventType: "user.created",
		Async:     true,
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "handled", Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
					},
				},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{event},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "~ \"user.created\" async") {
		t.Errorf("Compact output should contain '~ \"user.created\" async', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "handle \"user.created\" async") {
		t.Errorf("Expanded output should contain 'handle \"user.created\" async', got: %s", expanded)
	}
}

func TestFormatQueueWorker(t *testing.T) {
	queue := &ast.QueueWorker{
		QueueName:   "email.send",
		Concurrency: 5,
		MaxRetries:  3,
		Timeout:     30,
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "sent", Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
					},
				},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{queue},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "& \"email.send\"") {
		t.Errorf("Compact output should contain '& \"email.send\"', got: %s", compact)
	}
	if !strings.Contains(compact, "+ concurrency(5)") {
		t.Errorf("Compact output should contain '+ concurrency(5)', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "queue \"email.send\"") {
		t.Errorf("Expanded output should contain 'queue \"email.send\"', got: %s", expanded)
	}
	if !strings.Contains(expanded, "middleware concurrency(5)") {
		t.Errorf("Expanded output should contain 'middleware concurrency(5)', got: %s", expanded)
	}
}

func TestFormatIfStatement(t *testing.T) {
	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			ast.IfStatement{
				Condition: ast.BinaryOpExpr{
					Op:    ast.Gt,
					Left:  ast.VariableExpr{Name: "x"},
					Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				},
				ThenBlock: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "positive"}},
					},
				},
				ElseBlock: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "non-positive"}},
					},
				},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{route},
	}

	formatter := New(Expanded)
	output := formatter.Format(module)

	if !strings.Contains(output, "if x > 0 {") {
		t.Errorf("Output should contain 'if x > 0 {', got: %s", output)
	}
	if !strings.Contains(output, "} else {") {
		t.Errorf("Output should contain '} else {', got: %s", output)
	}
}

func TestFormatValidationStatement(t *testing.T) {
	route := &ast.Route{
		Method: ast.Post,
		Path:   "/api/users",
		Body: []ast.Statement{
			ast.ValidationStatement{
				Call: ast.FunctionCallExpr{
					Name: "validateEmail",
					Args: []ast.Expr{
						ast.VariableExpr{Name: "email"},
					},
				},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{route},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "? validateEmail(email)") {
		t.Errorf("Compact output should contain '? validateEmail(email)', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "validate validateEmail(email)") {
		t.Errorf("Expanded output should contain 'validate validateEmail(email)', got: %s", expanded)
	}
}

func TestFormatMiddlewareAndInjection(t *testing.T) {
	route := &ast.Route{
		Method: ast.Get,
		Path:   "/api/admin",
		Auth: &ast.AuthConfig{
			AuthType: "jwt",
			Required: true,
		},
		RateLimit: &ast.RateLimit{
			Requests: 100,
			Window:   "min",
		},
		Injections: []ast.Injection{
			{Name: "db", Type: ast.DatabaseType{}},
		},
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{Fields: []ast.ObjectField{}},
			},
		},
	}

	module := &ast.Module{
		Items: []ast.Item{route},
	}

	// Test compact mode
	compactFormatter := New(Compact)
	compact := compactFormatter.Format(module)

	if !strings.Contains(compact, "+ auth(jwt)") {
		t.Errorf("Compact output should contain '+ auth(jwt)', got: %s", compact)
	}
	if !strings.Contains(compact, "+ ratelimit(100/min)") {
		t.Errorf("Compact output should contain '+ ratelimit(100/min)', got: %s", compact)
	}
	if !strings.Contains(compact, "% db: Database") {
		t.Errorf("Compact output should contain '%% db: Database', got: %s", compact)
	}

	// Test expanded mode
	expandedFormatter := New(Expanded)
	expanded := expandedFormatter.Format(module)

	if !strings.Contains(expanded, "middleware auth(jwt)") {
		t.Errorf("Expanded output should contain 'middleware auth(jwt)', got: %s", expanded)
	}
	if !strings.Contains(expanded, "middleware ratelimit(100/min)") {
		t.Errorf("Expanded output should contain 'middleware ratelimit(100/min)', got: %s", expanded)
	}
	if !strings.Contains(expanded, "use db: Database") {
		t.Errorf("Expanded output should contain 'use db: Database', got: %s", expanded)
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\nworld", "hello\\nworld"},
		{"tab\there", "tab\\there"},
		{`quote"here`, `quote\"here`},
		{`back\slash`, `back\\slash`},
	}

	for _, tt := range tests {
		result := escapeString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// strPtr is a helper for creating string pointers in test data.
func strPtr(s string) *string {
	return &s
}

// formatViaModule is a helper that wraps an item in a module and formats it.
func formatViaModule(mode Mode, items ...ast.Item) string {
	module := &ast.Module{Items: items}
	f := New(mode)
	return f.Format(module)
}

// formatRoute is a helper that wraps statements inside a Route and formats them.
func formatRouteBody(mode Mode, stmts ...ast.Statement) string {
	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body:   stmts,
	}
	return formatViaModule(mode, route)
}

func TestFormatFunction_Compact(t *testing.T) {
	fn := &ast.Function{
		Name: "add",
		Params: []ast.Field{
			{Name: "a", TypeAnnotation: ast.IntType{}},
			{Name: "b", TypeAnnotation: ast.IntType{}},
		},
		ReturnType: ast.IntType{},
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  ast.VariableExpr{Name: "a"},
					Right: ast.VariableExpr{Name: "b"},
				},
			},
		},
	}
	result := formatViaModule(Compact, fn)
	if !strings.Contains(result, "= add(a: int, b: int) -> int {") {
		t.Errorf("Compact function should use '= ' prefix, got: %s", result)
	}
	if !strings.Contains(result, "> a + b") {
		t.Errorf("Body should contain return, got: %s", result)
	}
}

func TestFormatFunction_Expanded(t *testing.T) {
	fn := &ast.Function{
		Name: "greet",
		Params: []ast.Field{
			{Name: "name", TypeAnnotation: ast.StringType{}},
		},
		ReturnType: ast.StringType{},
		Body: []ast.Statement{
			ast.ReturnStatement{Value: ast.VariableExpr{Name: "name"}},
		},
	}
	result := formatViaModule(Expanded, fn)
	if !strings.Contains(result, "func greet(name: str) -> str {") {
		t.Errorf("Expanded function should use 'func ' prefix, got: %s", result)
	}
}

func TestFormatFunction_WithTypeParams(t *testing.T) {
	fn := &ast.Function{
		Name:       "identity",
		TypeParams: []ast.TypeParameter{{Name: "T"}, {Name: "U"}},
		Params:     []ast.Field{{Name: "val"}},
		Body: []ast.Statement{
			ast.ReturnStatement{Value: ast.VariableExpr{Name: "val"}},
		},
	}
	result := formatViaModule(Expanded, fn)
	if !strings.Contains(result, "func identity<T, U>(val)") {
		t.Errorf("Function with type params should include <T, U>, got: %s", result)
	}
}

func TestFormatFunction_NoReturnType(t *testing.T) {
	fn := &ast.Function{
		Name:   "doSomething",
		Params: []ast.Field{{Name: "x"}},
		Body:   []ast.Statement{},
	}
	result := formatViaModule(Expanded, fn)
	if strings.Contains(result, "->") {
		t.Errorf("Should not contain '->' when no return type, got: %s", result)
	}
}

func TestFormatWebSocketRoute_AllEventTypes(t *testing.T) {
	ws := &ast.WebSocketRoute{
		Path: "/ws/chat",
		Events: []ast.WebSocketEvent{
			{EventType: ast.WSEventConnect, Body: []ast.Statement{
				ast.ExpressionStatement{Expr: ast.FunctionCallExpr{Name: "log", Args: []ast.Expr{ast.LiteralExpr{Value: ast.StringLiteral{Value: "connected"}}}}},
			}},
			{EventType: ast.WSEventDisconnect, Body: []ast.Statement{}},
			{EventType: ast.WSEventMessage, Body: []ast.Statement{}},
			{EventType: ast.WSEventError, Body: []ast.Statement{}},
		},
	}
	compact := formatViaModule(Compact, ws)
	if !strings.Contains(compact, "@ WS /ws/chat {") {
		t.Errorf("Compact ws route should start with '@ WS', got: %s", compact)
	}
	for _, ev := range []string{"on connect {", "on disconnect {", "on message {", "on error {"} {
		if !strings.Contains(compact, ev) {
			t.Errorf("Should contain '%s', got: %s", ev, compact)
		}
	}
	expanded := formatViaModule(Expanded, ws)
	if !strings.Contains(expanded, "route WS /ws/chat {") {
		t.Errorf("Expanded ws route should use 'route WS', got: %s", expanded)
	}
}

func TestFormatImport_Simple(t *testing.T) {
	imp := &ast.ImportStatement{Path: "utils/helpers"}
	result := formatViaModule(Compact, imp)
	if !strings.Contains(result, `import "utils/helpers"`) {
		t.Errorf("Simple import should format correctly, got: %s", result)
	}
}

func TestFormatImport_WithAlias(t *testing.T) {
	imp := &ast.ImportStatement{Path: "utils/helpers", Alias: "h"}
	result := formatViaModule(Compact, imp)
	if !strings.Contains(result, `import "utils/helpers" as h`) {
		t.Errorf("Import with alias should format correctly, got: %s", result)
	}
}

func TestFormatImport_Selective(t *testing.T) {
	imp := &ast.ImportStatement{
		Path:      "math",
		Selective: true,
		Names: []ast.ImportName{
			{Name: "sqrt"},
			{Name: "pow", Alias: "power"},
		},
	}
	result := formatViaModule(Compact, imp)
	if !strings.Contains(result, `from "math" import { sqrt, pow as power }`) {
		t.Errorf("Selective import should format correctly, got: %s", result)
	}
}

func TestFormatModuleDecl(t *testing.T) {
	m := &ast.ModuleDecl{Name: "myapp"}
	result := formatViaModule(Compact, m)
	if !strings.Contains(result, `module "myapp"`) {
		t.Errorf("Module decl should format correctly, got: %s", result)
	}
}

func TestFormatMacroDef(t *testing.T) {
	macro := &ast.MacroDef{Name: "log", Params: []string{"level", "msg"}}
	result := formatViaModule(Compact, macro)
	if !strings.Contains(result, "macro! log(level, msg) {") {
		t.Errorf("Macro def should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "# ... macro body ...") {
		t.Errorf("Macro body placeholder should be present, got: %s", result)
	}
}

func TestFormatMacroDef_NoParams(t *testing.T) {
	macro := &ast.MacroDef{Name: "timestamp", Params: []string{}}
	result := formatViaModule(Compact, macro)
	if !strings.Contains(result, "macro! timestamp() {") {
		t.Errorf("Macro with no params should format correctly, got: %s", result)
	}
}

func TestFormatReassign(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "x", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}},
		ast.ReassignStatement{Target: "x", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}}},
	)
	if !strings.Contains(result, "$ x = 1") {
		t.Errorf("Should contain assign, got: %s", result)
	}
	// Reassign should not have $ prefix
	lines := strings.Split(result, "\n")
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "x = 2" {
			found = true
		}
	}
	if !found {
		t.Errorf("Should contain reassign 'x = 2' without $, got: %s", result)
	}
}

func TestFormatReassign_Pointer(t *testing.T) {
	result := formatRouteBody(Expanded,
		&ast.ReassignStatement{Target: "count", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}}},
	)
	if !strings.Contains(result, "count = 5") {
		t.Errorf("Pointer reassign should format correctly, got: %s", result)
	}
}

func TestFormatWhile(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.WhileStatement{
			Condition: ast.BinaryOpExpr{
				Op: ast.Lt, Left: ast.VariableExpr{Name: "i"}, Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			Body: []ast.Statement{
				ast.ReassignStatement{Target: "i", Value: ast.BinaryOpExpr{
					Op: ast.Add, Left: ast.VariableExpr{Name: "i"}, Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				}},
			},
		},
	)
	if !strings.Contains(result, "while i < 10 {") {
		t.Errorf("While loop should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "i = i + 1") {
		t.Errorf("While body should contain reassign, got: %s", result)
	}
}

func TestFormatWhile_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.WhileStatement{
			Condition: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			Body: []ast.Statement{
				ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}}},
			},
		},
	)
	if !strings.Contains(result, "while true {") {
		t.Errorf("Pointer while should format correctly, got: %s", result)
	}
}

func TestFormatFor_ValueOnly(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.ForStatement{
			ValueVar: "item",
			Iterable: ast.VariableExpr{Name: "items"},
			Body: []ast.Statement{
				ast.ExpressionStatement{Expr: ast.FunctionCallExpr{Name: "process", Args: []ast.Expr{ast.VariableExpr{Name: "item"}}}},
			},
		},
	)
	if !strings.Contains(result, "for item in items {") {
		t.Errorf("For with value only should format correctly, got: %s", result)
	}
}

func TestFormatFor_KeyAndValue(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ForStatement{
			KeyVar: "idx", ValueVar: "val",
			Iterable: ast.VariableExpr{Name: "arr"},
			Body:     []ast.Statement{},
		},
	)
	if !strings.Contains(result, "for idx, val in arr {") {
		t.Errorf("For with key+value should format correctly, got: %s", result)
	}
}

func TestFormatFor_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.ForStatement{
			ValueVar: "x", Iterable: ast.VariableExpr{Name: "list"},
			Body: []ast.Statement{},
		},
	)
	if !strings.Contains(result, "for x in list {") {
		t.Errorf("Pointer for should format correctly, got: %s", result)
	}
}

func TestFormatSwitch_WithDefault(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.SwitchStatement{
			Value: ast.VariableExpr{Name: "status"},
			Cases: []ast.SwitchCase{
				{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 200}}, Body: []ast.Statement{
					ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "ok"}}},
				}},
				{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 404}}, Body: []ast.Statement{
					ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "not found"}}},
				}},
			},
			Default: []ast.Statement{
				ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "error"}}},
			},
		},
	)
	if !strings.Contains(result, "switch status {") {
		t.Errorf("Switch should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "case 200 {") {
		t.Errorf("Should contain case 200, got: %s", result)
	}
	if !strings.Contains(result, "case 404 {") {
		t.Errorf("Should contain case 404, got: %s", result)
	}
	if !strings.Contains(result, "default {") {
		t.Errorf("Should contain default block, got: %s", result)
	}
}

func TestFormatSwitch_NoDefault(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.SwitchStatement{
			Value: ast.VariableExpr{Name: "x"},
			Cases: []ast.SwitchCase{
				{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}, Body: []ast.Statement{}},
			},
		},
	)
	if !strings.Contains(result, "switch x {") {
		t.Errorf("Switch should format, got: %s", result)
	}
	if strings.Contains(result, "default") {
		t.Errorf("Should not contain default, got: %s", result)
	}
}

func TestFormatSwitch_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.SwitchStatement{
			Value: ast.VariableExpr{Name: "y"},
			Cases: []ast.SwitchCase{
				{Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "a"}}, Body: []ast.Statement{}},
			},
		},
	)
	if !strings.Contains(result, "switch y {") {
		t.Errorf("Pointer switch should format correctly, got: %s", result)
	}
}

func TestFormatDbQuery_Compact(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.DbQueryStatement{
			Var: "users", Query: "SELECT * FROM users WHERE id = ?",
			Params: []ast.Expr{ast.VariableExpr{Name: "userId"}},
		},
	)
	if !strings.Contains(result, `$ users = db.query("SELECT * FROM users WHERE id = ?", userId)`) {
		t.Errorf("Compact db query should use $ prefix, got: %s", result)
	}
}

func TestFormatDbQuery_Expanded(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.DbQueryStatement{Var: "results", Query: "SELECT * FROM items"},
	)
	if !strings.Contains(result, `let results = db.query("SELECT * FROM items")`) {
		t.Errorf("Expanded db query should use let prefix, got: %s", result)
	}
}

func TestFormatDbQuery_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.DbQueryStatement{Var: "data", Query: "SELECT 1"},
	)
	if !strings.Contains(result, `$ data = db.query("SELECT 1")`) {
		t.Errorf("Pointer db query should format correctly, got: %s", result)
	}
}

func TestFormatExpr_UnaryOps(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "neg", Value: ast.UnaryOpExpr{Op: ast.Neg, Right: ast.VariableExpr{Name: "x"}}},
		ast.AssignStatement{Target: "notVal", Value: ast.UnaryOpExpr{Op: ast.Not, Right: ast.VariableExpr{Name: "flag"}}},
	)
	if !strings.Contains(result, "let neg = -x") {
		t.Errorf("Unary neg should format as -x, got: %s", result)
	}
	if !strings.Contains(result, "let notVal = !flag") {
		t.Errorf("Unary not should format as !flag, got: %s", result)
	}
}

func TestFormatExpr_UnaryOp_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "neg", Value: &ast.UnaryOpExpr{Op: ast.Neg, Right: ast.VariableExpr{Name: "y"}}},
	)
	if !strings.Contains(result, "$ neg = -y") {
		t.Errorf("Pointer unary should format correctly, got: %s", result)
	}
}

func TestFormatExpr_FieldAccess(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "name", Value: ast.FieldAccessExpr{Object: ast.VariableExpr{Name: "user"}, Field: "name"}},
	)
	if !strings.Contains(result, "let name = user.name") {
		t.Errorf("Field access should format as user.name, got: %s", result)
	}
}

func TestFormatExpr_FieldAccess_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "v", Value: &ast.FieldAccessExpr{Object: ast.VariableExpr{Name: "obj"}, Field: "val"}},
	)
	if !strings.Contains(result, "$ v = obj.val") {
		t.Errorf("Pointer field access should format correctly, got: %s", result)
	}
}

func TestFormatExpr_ArrayIndex(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "first", Value: ast.ArrayIndexExpr{
			Array: ast.VariableExpr{Name: "items"}, Index: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		}},
	)
	if !strings.Contains(result, "let first = items[0]") {
		t.Errorf("Array index should format as items[0], got: %s", result)
	}
}

func TestFormatExpr_ArrayIndex_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "el", Value: &ast.ArrayIndexExpr{
			Array: ast.VariableExpr{Name: "arr"}, Index: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		}},
	)
	if !strings.Contains(result, "$ el = arr[1]") {
		t.Errorf("Pointer array index should format correctly, got: %s", result)
	}
}

func TestFormatExpr_FunctionCallPointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ExpressionStatement{Expr: &ast.FunctionCallExpr{
			Name: "doStuff", Args: []ast.Expr{ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}}},
		}},
	)
	if !strings.Contains(result, "doStuff(42)") {
		t.Errorf("Pointer function call should format correctly, got: %s", result)
	}
}

func TestFormatExpr_ObjectPointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: &ast.ObjectExpr{
			Fields: []ast.ObjectField{{Key: "a", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}}},
		}},
	)
	if !strings.Contains(result, "{a: 1}") {
		t.Errorf("Pointer object should format correctly, got: %s", result)
	}
}

func TestFormatExpr_ArrayPointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: &ast.ArrayExpr{
			Elements: []ast.Expr{ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}, ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}}},
		}},
	)
	if !strings.Contains(result, "[1, 2]") {
		t.Errorf("Pointer array should format correctly, got: %s", result)
	}
}

func TestFormatExpr_LiteralPointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 99}}},
	)
	if !strings.Contains(result, "> 99") {
		t.Errorf("Pointer literal should format correctly, got: %s", result)
	}
}

func TestFormatExpr_VariablePointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: &ast.VariableExpr{Name: "foo"}},
	)
	if !strings.Contains(result, "> foo") {
		t.Errorf("Pointer variable should format correctly, got: %s", result)
	}
}

func TestFormatExpr_BinaryOpPointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: &ast.BinaryOpExpr{Op: ast.Mul, Left: ast.VariableExpr{Name: "a"}, Right: ast.VariableExpr{Name: "b"}}},
	)
	if !strings.Contains(result, "> a * b") {
		t.Errorf("Pointer binary op should format correctly, got: %s", result)
	}
}

func TestFormatExpr_AwaitExpr(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "result", Value: ast.AwaitExpr{
			Expr: ast.FunctionCallExpr{Name: "fetchData", Args: []ast.Expr{}},
		}},
	)
	if !strings.Contains(result, "let result = await fetchData()") {
		t.Errorf("Await should format correctly, got: %s", result)
	}
}

func TestFormatExpr_AwaitPointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "r", Value: &ast.AwaitExpr{Expr: ast.VariableExpr{Name: "future"}}},
	)
	if !strings.Contains(result, "$ r = await future") {
		t.Errorf("Pointer await should format correctly, got: %s", result)
	}
}

func TestFormatFunctionCall_WithTypeArgs(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.ExpressionStatement{Expr: ast.FunctionCallExpr{
			Name: "parse", TypeArgs: []ast.Type{ast.IntType{}, ast.StringType{}},
			Args: []ast.Expr{ast.VariableExpr{Name: "data"}},
		}},
	)
	if !strings.Contains(result, "parse<int, str>(data)") {
		t.Errorf("Function call with type args should format correctly, got: %s", result)
	}
}

func TestFormatFunctionCall_NoArgs(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ExpressionStatement{Expr: ast.FunctionCallExpr{Name: "now", Args: []ast.Expr{}}},
	)
	if !strings.Contains(result, "now()") {
		t.Errorf("No-arg function call should format correctly, got: %s", result)
	}
}

func TestFormatFunctionCall_MultipleArgs(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ExpressionStatement{Expr: ast.FunctionCallExpr{
			Name: "range",
			Args: []ast.Expr{
				ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
				ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
		}},
	)
	if !strings.Contains(result, "range(0, 10, 2)") {
		t.Errorf("Multiple arg function call should format correctly, got: %s", result)
	}
}

func TestFormatLambda_ExprBody(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "double", Value: ast.LambdaExpr{
			Params: []ast.Field{{Name: "x", TypeAnnotation: ast.IntType{}}},
			Body: ast.BinaryOpExpr{
				Op: ast.Mul, Left: ast.VariableExpr{Name: "x"}, Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
		}},
	)
	if !strings.Contains(result, "(x: int) => x * 2") {
		t.Errorf("Lambda with expr body should format correctly, got: %s", result)
	}
}

func TestFormatLambda_BlockBody(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "fn", Value: ast.LambdaExpr{
			Params: []ast.Field{{Name: "a"}, {Name: "b"}},
			Block: []ast.Statement{
				ast.AssignStatement{Target: "sum", Value: ast.BinaryOpExpr{
					Op: ast.Add, Left: ast.VariableExpr{Name: "a"}, Right: ast.VariableExpr{Name: "b"},
				}},
				ast.ReturnStatement{Value: ast.VariableExpr{Name: "sum"}},
			},
		}},
	)
	if !strings.Contains(result, "(a, b) => {") {
		t.Errorf("Lambda with block body should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "let sum = a + b") {
		t.Errorf("Lambda block should contain statements, got: %s", result)
	}
}

func TestFormatLambda_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "inc", Value: &ast.LambdaExpr{
			Params: []ast.Field{{Name: "n"}},
			Body: ast.BinaryOpExpr{
				Op: ast.Add, Left: ast.VariableExpr{Name: "n"}, Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
		}},
	)
	if !strings.Contains(result, "(n) => n + 1") {
		t.Errorf("Pointer lambda should format correctly, got: %s", result)
	}
}

func TestFormatLambda_NoParams(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "getZero", Value: ast.LambdaExpr{
			Params: []ast.Field{},
			Body:   ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		}},
	)
	if !strings.Contains(result, "() => 0") {
		t.Errorf("Lambda with no params should format as () => ..., got: %s", result)
	}
}

func TestFormatMatch_Cases(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "label", Value: ast.MatchExpr{
			Value: ast.VariableExpr{Name: "code"},
			Cases: []ast.MatchCase{
				{Pattern: ast.LiteralPattern{Value: ast.IntLiteral{Value: 200}}, Body: ast.LiteralExpr{Value: ast.StringLiteral{Value: "OK"}}},
				{Pattern: ast.LiteralPattern{Value: ast.IntLiteral{Value: 404}}, Body: ast.LiteralExpr{Value: ast.StringLiteral{Value: "Not Found"}}},
				{Pattern: ast.WildcardPattern{}, Body: ast.LiteralExpr{Value: ast.StringLiteral{Value: "Unknown"}}},
			},
		}},
	)
	if !strings.Contains(result, "match code {") {
		t.Errorf("Match should start correctly, got: %s", result)
	}
	if !strings.Contains(result, `200 => "OK"`) {
		t.Errorf("Match case 200 missing, got: %s", result)
	}
	if !strings.Contains(result, `_ => "Unknown"`) {
		t.Errorf("Wildcard case missing, got: %s", result)
	}
}

func TestFormatMatch_WithGuard(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "msg", Value: ast.MatchExpr{
			Value: ast.VariableExpr{Name: "x"},
			Cases: []ast.MatchCase{
				{
					Pattern: ast.VariablePattern{Name: "n"},
					Guard:   ast.BinaryOpExpr{Op: ast.Gt, Left: ast.VariableExpr{Name: "n"}, Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}}},
					Body:    ast.LiteralExpr{Value: ast.StringLiteral{Value: "positive"}},
				},
				{Pattern: ast.WildcardPattern{}, Body: ast.LiteralExpr{Value: ast.StringLiteral{Value: "non-positive"}}},
			},
		}},
	)
	if !strings.Contains(result, `n when n > 0 => "positive"`) {
		t.Errorf("Match with guard should format correctly, got: %s", result)
	}
}

func TestFormatMatch_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "r", Value: &ast.MatchExpr{
			Value: ast.VariableExpr{Name: "val"},
			Cases: []ast.MatchCase{
				{Pattern: ast.WildcardPattern{}, Body: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}}},
			},
		}},
	)
	if !strings.Contains(result, "match val {") {
		t.Errorf("Pointer match should format correctly, got: %s", result)
	}
}

func TestFormatAsync_Body(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "future", Value: ast.AsyncExpr{
			Body: []ast.Statement{
				ast.AssignStatement{Target: "data", Value: ast.FunctionCallExpr{Name: "fetch", Args: []ast.Expr{}}},
				ast.ReturnStatement{Value: ast.VariableExpr{Name: "data"}},
			},
		}},
	)
	if !strings.Contains(result, "async {") {
		t.Errorf("Async should start with 'async {', got: %s", result)
	}
	if !strings.Contains(result, "let data = fetch()") {
		t.Errorf("Async body should contain statements, got: %s", result)
	}
}

func TestFormatAsync_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.AssignStatement{Target: "f", Value: &ast.AsyncExpr{
			Body: []ast.Statement{
				ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}}},
			},
		}},
	)
	if !strings.Contains(result, "async {") {
		t.Errorf("Pointer async should format correctly, got: %s", result)
	}
}

func TestFormatPattern_ObjectPattern(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "result", Value: ast.MatchExpr{
			Value: ast.VariableExpr{Name: "obj"},
			Cases: []ast.MatchCase{{
				Pattern: ast.ObjectPattern{Fields: []ast.ObjectPatternField{
					{Key: "name"},
					{Key: "age", Pattern: ast.VariablePattern{Name: "a"}},
				}},
				Body: ast.VariableExpr{Name: "a"},
			}},
		}},
	)
	if !strings.Contains(result, "{name, age: a}") {
		t.Errorf("Object pattern should format correctly, got: %s", result)
	}
}

func TestFormatPattern_ArrayPatternWithRest(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "result", Value: ast.MatchExpr{
			Value: ast.VariableExpr{Name: "arr"},
			Cases: []ast.MatchCase{{
				Pattern: ast.ArrayPattern{
					Elements: []ast.Pattern{ast.VariablePattern{Name: "head"}},
					Rest:     strPtr("tail"),
				},
				Body: ast.VariableExpr{Name: "head"},
			}},
		}},
	)
	if !strings.Contains(result, "[head, ...tail]") {
		t.Errorf("Array pattern with rest should format correctly, got: %s", result)
	}
}

func TestFormatPattern_ArrayPatternNoRest(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "r", Value: ast.MatchExpr{
			Value: ast.VariableExpr{Name: "pair"},
			Cases: []ast.MatchCase{{
				Pattern: ast.ArrayPattern{
					Elements: []ast.Pattern{ast.VariablePattern{Name: "a"}, ast.VariablePattern{Name: "b"}},
				},
				Body: ast.VariableExpr{Name: "a"},
			}},
		}},
	)
	if !strings.Contains(result, "[a, b]") {
		t.Errorf("Array pattern without rest should format correctly, got: %s", result)
	}
}

func TestFormatPattern_ArrayEmptyWithRest(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.AssignStatement{Target: "r", Value: ast.MatchExpr{
			Value: ast.VariableExpr{Name: "list"},
			Cases: []ast.MatchCase{{
				Pattern: ast.ArrayPattern{Elements: []ast.Pattern{}, Rest: strPtr("all")},
				Body:    ast.VariableExpr{Name: "all"},
			}},
		}},
	)
	if !strings.Contains(result, "[...all]") {
		t.Errorf("Empty array pattern with rest should format correctly, got: %s", result)
	}
}

func TestFormatType_AllBasicTypes(t *testing.T) {
	fields := []ast.Field{
		{Name: "a", TypeAnnotation: ast.IntType{}},
		{Name: "b", TypeAnnotation: ast.StringType{}},
		{Name: "c", TypeAnnotation: ast.BoolType{}},
		{Name: "d", TypeAnnotation: ast.FloatType{}},
		{Name: "e", TypeAnnotation: ast.DatabaseType{}},
		{Name: "f", TypeAnnotation: ast.NamedType{Name: "User"}},
	}
	td := &ast.TypeDef{Name: "AllTypes", Fields: fields}
	result := formatViaModule(Compact, td)
	for _, expect := range []string{"a: int", "b: str", "c: bool", "d: float", "e: Database", "f: User"} {
		if !strings.Contains(result, expect) {
			t.Errorf("Should contain '%s', got: %s", expect, result)
		}
	}
}

func TestFormatType_ArrayType(t *testing.T) {
	td := &ast.TypeDef{Name: "T", Fields: []ast.Field{
		{Name: "ids", TypeAnnotation: ast.ArrayType{ElementType: ast.IntType{}}},
	}}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "ids: int[]") {
		t.Errorf("Array type should format as 'int[]', got: %s", result)
	}
}

func TestFormatType_OptionalType(t *testing.T) {
	td := &ast.TypeDef{Name: "T", Fields: []ast.Field{
		{Name: "bio", TypeAnnotation: ast.OptionalType{InnerType: ast.StringType{}}},
	}}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "bio: str?") {
		t.Errorf("Optional type should format as 'str?', got: %s", result)
	}
}

func TestFormatType_GenericType(t *testing.T) {
	td := &ast.TypeDef{Name: "T", Fields: []ast.Field{
		{Name: "map", TypeAnnotation: ast.GenericType{
			BaseType: ast.NamedType{Name: "Map"},
			TypeArgs: []ast.Type{ast.StringType{}, ast.IntType{}},
		}},
	}}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "map: Map<str, int>") {
		t.Errorf("Generic type should format correctly, got: %s", result)
	}
}

func TestFormatType_TypeParameterType(t *testing.T) {
	td := &ast.TypeDef{Name: "Box", TypeParams: []ast.TypeParameter{{Name: "T"}},
		Fields: []ast.Field{{Name: "value", TypeAnnotation: ast.TypeParameterType{Name: "T"}}},
	}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "value: T") {
		t.Errorf("TypeParameterType should format as name, got: %s", result)
	}
}

func TestFormatType_FunctionType(t *testing.T) {
	td := &ast.TypeDef{Name: "T", Fields: []ast.Field{
		{Name: "cb", TypeAnnotation: ast.FunctionType{
			ParamTypes: []ast.Type{ast.StringType{}, ast.IntType{}},
			ReturnType: ast.BoolType{},
		}},
	}}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "cb: (str, int) -> bool") {
		t.Errorf("Function type should format correctly, got: %s", result)
	}
}

func TestFormatType_UnionType(t *testing.T) {
	td := &ast.TypeDef{Name: "T", Fields: []ast.Field{
		{Name: "val", TypeAnnotation: ast.UnionType{
			Types: []ast.Type{ast.StringType{}, ast.IntType{}, ast.BoolType{}},
		}},
	}}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "val: str | int | bool") {
		t.Errorf("Union type should format with pipes, got: %s", result)
	}
}

func TestFormatType_FutureType(t *testing.T) {
	td := &ast.TypeDef{Name: "T", Fields: []ast.Field{
		{Name: "task", TypeAnnotation: ast.FutureType{ResultType: ast.StringType{}}},
	}}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, "task: Future<str>") {
		t.Errorf("Future type should format correctly, got: %s", result)
	}
}

func TestFormatObject_MoreThanThreeFields(t *testing.T) {
	result := formatRouteBody(Expanded,
		ast.ReturnStatement{Value: ast.ObjectExpr{Fields: []ast.ObjectField{
			{Key: "a", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}},
			{Key: "b", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}}},
			{Key: "c", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}}},
			{Key: "d", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}}},
		}}},
	)
	// With >3 fields, multi-line format is used with trailing commas (except last)
	if !strings.Contains(result, "a: 1,") {
		t.Errorf("Multi-line object should have trailing commas, got: %s", result)
	}
	if !strings.Contains(result, "d: 4") {
		t.Errorf("Last field should be present, got: %s", result)
	}
}

func TestFormatObject_ThreeFieldsInline(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: ast.ObjectExpr{Fields: []ast.ObjectField{
			{Key: "r", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 255}}},
			{Key: "g", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 128}}},
			{Key: "b", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}}},
		}}},
	)
	if !strings.Contains(result, "{r: 255, g: 128, b: 0}") {
		t.Errorf("3-field object should be inline, got: %s", result)
	}
}

func TestFormatArray_MultipleElements(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: ast.ArrayExpr{Elements: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
		}}},
	)
	if !strings.Contains(result, "[1, 2, 3]") {
		t.Errorf("Array should format with commas, got: %s", result)
	}
}

func TestFormatArray_SingleElement(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: ast.ArrayExpr{Elements: []ast.Expr{
			ast.LiteralExpr{Value: ast.StringLiteral{Value: "only"}},
		}}},
	)
	if !strings.Contains(result, `["only"]`) {
		t.Errorf("Single element array should format correctly, got: %s", result)
	}
}

func TestFormatLiteral_Float(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.FloatLiteral{Value: 3.14}}},
	)
	if !strings.Contains(result, "> 3.14") {
		t.Errorf("Float literal should format correctly, got: %s", result)
	}
}

func TestFormatLiteral_Null(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.NullLiteral{}}},
	)
	if !strings.Contains(result, "> null") {
		t.Errorf("Null literal should format as 'null', got: %s", result)
	}
}

func TestFormatLiteral_BoolFalse(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}}},
	)
	if !strings.Contains(result, "> false") {
		t.Errorf("Bool false should format correctly, got: %s", result)
	}
}

func TestFormatWsSend(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.WsSendStatement{
			Client: ast.VariableExpr{Name: "client"}, Message: ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello"}},
		},
	)
	if !strings.Contains(result, `ws.send(client, "hello")`) {
		t.Errorf("WsSend should format correctly, got: %s", result)
	}
}

func TestFormatWsSend_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.WsSendStatement{Client: ast.VariableExpr{Name: "c"}, Message: ast.VariableExpr{Name: "msg"}},
	)
	if !strings.Contains(result, "ws.send(c, msg)") {
		t.Errorf("Pointer WsSend should format correctly, got: %s", result)
	}
}

func TestFormatWsBroadcast(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.WsBroadcastStatement{Message: ast.LiteralExpr{Value: ast.StringLiteral{Value: "update"}}},
	)
	if !strings.Contains(result, `ws.broadcast("update")`) {
		t.Errorf("WsBroadcast should format correctly, got: %s", result)
	}
}

func TestFormatWsBroadcast_WithExcept(t *testing.T) {
	except := ast.Expr(ast.VariableExpr{Name: "sender"})
	result := formatRouteBody(Compact,
		ast.WsBroadcastStatement{
			Message: ast.LiteralExpr{Value: ast.StringLiteral{Value: "msg"}}, Except: &except,
		},
	)
	if !strings.Contains(result, `ws.broadcast("msg", except: sender)`) {
		t.Errorf("WsBroadcast with except should format correctly, got: %s", result)
	}
}

func TestFormatWsBroadcast_Pointer(t *testing.T) {
	except := ast.Expr(ast.VariableExpr{Name: "me"})
	result := formatRouteBody(Compact,
		&ast.WsBroadcastStatement{
			Message: ast.LiteralExpr{Value: ast.StringLiteral{Value: "hi"}}, Except: &except,
		},
	)
	if !strings.Contains(result, `ws.broadcast("hi", except: me)`) {
		t.Errorf("Pointer WsBroadcast should format correctly, got: %s", result)
	}
}

func TestFormatWsClose(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.WsCloseStatement{
			Client: ast.VariableExpr{Name: "client"},
			Reason: ast.LiteralExpr{Value: ast.StringLiteral{Value: "timeout"}},
		},
	)
	if !strings.Contains(result, `ws.close(client, "timeout")`) {
		t.Errorf("WsClose with reason should format correctly, got: %s", result)
	}
}

func TestFormatWsClose_NoReason(t *testing.T) {
	result := formatRouteBody(Compact,
		ast.WsCloseStatement{Client: ast.VariableExpr{Name: "client"}},
	)
	if !strings.Contains(result, "ws.close(client)") {
		t.Errorf("WsClose without reason should format correctly, got: %s", result)
	}
}

func TestFormatWsClose_Pointer(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.WsCloseStatement{
			Client: ast.VariableExpr{Name: "c"},
			Reason: ast.LiteralExpr{Value: ast.StringLiteral{Value: "done"}},
		},
	)
	if !strings.Contains(result, `ws.close(c, "done")`) {
		t.Errorf("Pointer WsClose should format correctly, got: %s", result)
	}
}

func TestFormatStatement_PointerTypes(t *testing.T) {
	// Test pointer versions of assign, return, if, expression, validation
	result := formatRouteBody(Compact,
		&ast.AssignStatement{Target: "z", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}}},
		&ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}}},
	)
	if !strings.Contains(result, "$ z = 5") {
		t.Errorf("Pointer assign should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "> 42") {
		t.Errorf("Pointer return should format correctly, got: %s", result)
	}
}

func TestFormatStatement_PointerIf(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.IfStatement{
			Condition: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			ThenBlock: []ast.Statement{ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}}},
		},
	)
	if !strings.Contains(result, "if true {") {
		t.Errorf("Pointer if should format correctly, got: %s", result)
	}
}

func TestFormatStatement_PointerExpression(t *testing.T) {
	result := formatRouteBody(Compact,
		&ast.ExpressionStatement{Expr: ast.FunctionCallExpr{Name: "hello", Args: []ast.Expr{}}},
	)
	if !strings.Contains(result, "hello()") {
		t.Errorf("Pointer expression stmt should format correctly, got: %s", result)
	}
}

func TestFormatStatement_PointerValidation(t *testing.T) {
	result := formatRouteBody(Expanded,
		&ast.ValidationStatement{
			Call: ast.FunctionCallExpr{Name: "check", Args: []ast.Expr{ast.VariableExpr{Name: "x"}}},
		},
	)
	if !strings.Contains(result, "validate check(x)") {
		t.Errorf("Pointer validation should format correctly, got: %s", result)
	}
}

func TestFormatTypeDef_WithConstrainedTypeParams(t *testing.T) {
	td := &ast.TypeDef{
		Name: "Result",
		TypeParams: []ast.TypeParameter{
			{Name: "T"},
			{Name: "E", Constraint: ast.NamedType{Name: "Error"}},
		},
		Fields: []ast.Field{
			{Name: "ok", TypeAnnotation: ast.TypeParameterType{Name: "T"}},
		},
	}
	result := formatViaModule(Compact, td)
	if !strings.Contains(result, ": Result<T, E: Error> {") {
		t.Errorf("TypeDef with constrained type params should format correctly, got: %s", result)
	}
}

func TestFormatRoute_QueryParamsAndReturnType(t *testing.T) {
	route := &ast.Route{
		Method: ast.Get, Path: "/api/users",
		QueryParams: []ast.QueryParamDecl{
			{Name: "page", Type: ast.IntType{}, Required: true},
			{Name: "limit", Type: ast.IntType{}, Default: ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}}},
		},
		ReturnType: ast.NamedType{Name: "UserList"},
		Body: []ast.Statement{
			ast.ReturnStatement{Value: ast.ObjectExpr{Fields: []ast.ObjectField{}}},
		},
	}
	result := formatViaModule(Expanded, route)
	if !strings.Contains(result, "?page: int!") {
		t.Errorf("Required query param should have !, got: %s", result)
	}
	if !strings.Contains(result, "?limit: int = 20") {
		t.Errorf("Query param with default should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "-> UserList") {
		t.Errorf("Return type should format correctly, got: %s", result)
	}
}

func TestFormatCommand_DescriptionAndFlags(t *testing.T) {
	cmd := &ast.Command{
		Name: "deploy", Description: "Deploy the application",
		Params: []ast.CommandParam{
			{Name: "env", Type: ast.StringType{}, Required: true},
			{Name: "verbose", IsFlag: true, Type: ast.BoolType{}, Default: ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}}},
		},
		Body: []ast.Statement{},
	}
	result := formatViaModule(Compact, cmd)
	if !strings.Contains(result, `! deploy "Deploy the application"`) {
		t.Errorf("Command with description should format correctly, got: %s", result)
	}
	if !strings.Contains(result, "--verbose: bool = false") {
		t.Errorf("Flag param should have -- prefix, got: %s", result)
	}
	if !strings.Contains(result, "env: str!") {
		t.Errorf("Required param should have !, got: %s", result)
	}
}

func TestFormat_MultipleItemsSeparator(t *testing.T) {
	module := &ast.Module{Items: []ast.Item{
		&ast.ModuleDecl{Name: "app"},
		&ast.ImportStatement{Path: "utils"},
	}}
	result := New(Compact).Format(module)
	if !strings.Contains(result, "module \"app\"\n\nimport \"utils\"") {
		t.Errorf("Items should be separated by blank lines, got: %s", result)
	}
}

func TestFormatCronTask_WithInjections(t *testing.T) {
	cron := &ast.CronTask{
		Schedule:   "*/5 * * * *",
		Injections: []ast.Injection{{Name: "db", Type: ast.DatabaseType{}}},
		Body: []ast.Statement{
			ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
		},
	}
	compact := formatViaModule(Compact, cron)
	if !strings.Contains(compact, "% db: Database") {
		t.Errorf("Cron injections should format with %%, got: %s", compact)
	}
	expanded := formatViaModule(Expanded, cron)
	if !strings.Contains(expanded, "use db: Database") {
		t.Errorf("Expanded cron injections should use 'use', got: %s", expanded)
	}
}

func TestFormatEventHandler_WithInjections(t *testing.T) {
	eh := &ast.EventHandler{
		EventType:  "order.paid",
		Injections: []ast.Injection{{Name: "db", Type: ast.DatabaseType{}}},
		Body: []ast.Statement{
			ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
		},
	}
	compact := formatViaModule(Compact, eh)
	if !strings.Contains(compact, "% db: Database") {
		t.Errorf("Event handler injections should format, got: %s", compact)
	}
	expanded := formatViaModule(Expanded, eh)
	if !strings.Contains(expanded, "use db: Database") {
		t.Errorf("Expanded event handler injections should use 'use', got: %s", expanded)
	}
}

func TestFormatAllBinaryOps(t *testing.T) {
	ops := []struct {
		op       ast.BinOp
		expected string
	}{
		{ast.Add, "+"}, {ast.Sub, "-"}, {ast.Mul, "*"}, {ast.Div, "/"},
		{ast.Eq, "=="}, {ast.Ne, "!="}, {ast.Lt, "<"}, {ast.Le, "<="},
		{ast.Gt, ">"}, {ast.Ge, ">="}, {ast.And, "&&"}, {ast.Or, "||"},
	}
	for _, tt := range ops {
		result := formatRouteBody(Expanded,
			ast.AssignStatement{Target: "r", Value: ast.BinaryOpExpr{
				Op: tt.op, Left: ast.VariableExpr{Name: "a"}, Right: ast.VariableExpr{Name: "b"},
			}},
		)
		expected := "a " + tt.expected + " b"
		if !strings.Contains(result, expected) {
			t.Errorf("Binary op %s should produce '%s', got: %s", tt.expected, expected, result)
		}
	}
}

func TestFormatQueueWorker_WithInjections(t *testing.T) {
	qw := &ast.QueueWorker{
		QueueName:  "tasks",
		Injections: []ast.Injection{{Name: "cache", Type: ast.NamedType{Name: "Redis"}}},
		Body: []ast.Statement{
			ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
		},
	}
	compact := formatViaModule(Compact, qw)
	if !strings.Contains(compact, "% cache: Redis") {
		t.Errorf("Queue worker injections should format, got: %s", compact)
	}
	expanded := formatViaModule(Expanded, qw)
	if !strings.Contains(expanded, "use cache: Redis") {
		t.Errorf("Expanded queue worker injections should use 'use', got: %s", expanded)
	}
}
