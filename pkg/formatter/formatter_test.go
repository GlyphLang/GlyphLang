package formatter

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

func TestFormatRoute(t *testing.T) {
	route := &interpreter.Route{
		Method: interpreter.Get,
		Path:   "/api/users",
		Body: []interpreter.Statement{
			interpreter.AssignStatement{
				Target: "users",
				Value:  interpreter.ArrayExpr{Elements: []interpreter.Expr{}},
			},
			interpreter.ReturnStatement{
				Value: interpreter.VariableExpr{Name: "users"},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{route},
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
	typeDef := &interpreter.TypeDef{
		Name: "User",
		Fields: []interpreter.Field{
			{Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true},
			{Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true},
			{Name: "email", TypeAnnotation: interpreter.StringType{}, Required: false},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{typeDef},
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
	cmd := &interpreter.Command{
		Name: "hello",
		Params: []interpreter.CommandParam{
			{Name: "name", Type: interpreter.StringType{}, Required: true},
		},
		Body: []interpreter.Statement{
			interpreter.ReturnStatement{
				Value: interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "message", Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello"}}},
					},
				},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{cmd},
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
	cron := &interpreter.CronTask{
		Schedule: "0 0 * * *",
		Name:     "daily_task",
		Body: []interpreter.Statement{
			interpreter.ReturnStatement{
				Value: interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "done", Value: interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}},
					},
				},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{cron},
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
	event := &interpreter.EventHandler{
		EventType: "user.created",
		Async:     true,
		Body: []interpreter.Statement{
			interpreter.ReturnStatement{
				Value: interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "handled", Value: interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}},
					},
				},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{event},
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

	if !strings.Contains(expanded, "event \"user.created\" async") {
		t.Errorf("Expanded output should contain 'event \"user.created\" async', got: %s", expanded)
	}
}

func TestFormatQueueWorker(t *testing.T) {
	queue := &interpreter.QueueWorker{
		QueueName:   "email.send",
		Concurrency: 5,
		MaxRetries:  3,
		Timeout:     30,
		Body: []interpreter.Statement{
			interpreter.ReturnStatement{
				Value: interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "sent", Value: interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}},
					},
				},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{queue},
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
	route := &interpreter.Route{
		Method: interpreter.Get,
		Path:   "/test",
		Body: []interpreter.Statement{
			interpreter.IfStatement{
				Condition: interpreter.BinaryOpExpr{
					Op:    interpreter.Gt,
					Left:  interpreter.VariableExpr{Name: "x"},
					Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
				},
				ThenBlock: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "positive"}},
					},
				},
				ElseBlock: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "non-positive"}},
					},
				},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{route},
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
	route := &interpreter.Route{
		Method: interpreter.Post,
		Path:   "/api/users",
		Body: []interpreter.Statement{
			interpreter.ValidationStatement{
				Call: interpreter.FunctionCallExpr{
					Name: "validateEmail",
					Args: []interpreter.Expr{
						interpreter.VariableExpr{Name: "email"},
					},
				},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{route},
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
	route := &interpreter.Route{
		Method: interpreter.Get,
		Path:   "/api/admin",
		Auth: &interpreter.AuthConfig{
			AuthType: "jwt",
			Required: true,
		},
		RateLimit: &interpreter.RateLimit{
			Requests: 100,
			Window:   "min",
		},
		Injections: []interpreter.Injection{
			{Name: "db", Type: interpreter.DatabaseType{}},
		},
		Body: []interpreter.Statement{
			interpreter.ReturnStatement{
				Value: interpreter.ObjectExpr{Fields: []interpreter.ObjectField{}},
			},
		},
	}

	module := &interpreter.Module{
		Items: []interpreter.Item{route},
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
	if !strings.Contains(expanded, "inject db: Database") {
		t.Errorf("Expanded output should contain 'inject db: Database', got: %s", expanded)
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
