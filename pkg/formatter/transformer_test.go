package formatter

import (
	"strings"
	"testing"
)

func TestExpandSource(t *testing.T) {
	input := `# Comment preserved
@ GET /api/users {
  $ users = db.find()
  > users
}`

	result := ExpandSource(input)

	// Check comment is preserved
	if !strings.Contains(result, "# Comment preserved") {
		t.Errorf("Comment should be preserved, got: %s", result)
	}

	// Check transformations
	if !strings.Contains(result, "route GET /api/users") {
		t.Errorf("@ should become 'route', got: %s", result)
	}
	if !strings.Contains(result, "let users") {
		t.Errorf("$ should become 'let', got: %s", result)
	}
	if !strings.Contains(result, "return users") {
		t.Errorf("> should become 'return', got: %s", result)
	}
}

func TestCompactSource(t *testing.T) {
	input := `# Comment preserved
route GET /api/users {
  let users = db.find()
  return users
}`

	result := CompactSource(input)

	// Check comment is preserved
	if !strings.Contains(result, "# Comment preserved") {
		t.Errorf("Comment should be preserved, got: %s", result)
	}

	// Check transformations
	if !strings.Contains(result, "@ GET /api/users") {
		t.Errorf("'route' should become @, got: %s", result)
	}
	if !strings.Contains(result, "$ users") {
		t.Errorf("'let' should become $, got: %s", result)
	}
	if !strings.Contains(result, "> users") {
		t.Errorf("'return' should become >, got: %s", result)
	}
}

func TestRoundTrip(t *testing.T) {
	original := `# This is a comment
: User {
  id: int!
  name: str!
}

@ GET /users {
  $ users = db.find()
  > users
}

@ POST /users {
  $ name = input.name
  $ user = db.create({name: name})
  > user
}
`

	// Expand then compact
	expanded := ExpandSource(original)
	compacted := CompactSource(expanded)

	// Check key elements are preserved
	if !strings.Contains(compacted, "# This is a comment") {
		t.Errorf("Comment should be preserved through round-trip")
	}
	if !strings.Contains(compacted, ": User {") {
		t.Errorf("Type definition should be preserved")
	}
	if !strings.Contains(compacted, "@ GET /users") {
		t.Errorf("Route should be preserved")
	}
	if !strings.Contains(compacted, "$ users = db.find()") {
		t.Errorf("Assignment should be preserved")
	}
}

func TestPlusOperatorNotTransformed(t *testing.T) {
	input := `@ GET /greet/:name {
  $ greeting = "Hello, " + name + "!"
  > greeting
}`

	result := ExpandSource(input)

	// + in expressions should NOT become middleware
	if strings.Contains(result, "middleware") {
		t.Errorf("+ in expression should not become 'middleware', got: %s", result)
	}
	if !strings.Contains(result, `"Hello, " + name + "!"`) {
		t.Errorf("String concatenation should be preserved, got: %s", result)
	}
}

func TestMiddlewareTransformed(t *testing.T) {
	input := `@ GET /api/admin {
  + auth(jwt)
  + ratelimit(100/min)
  > {ok: true}
}`

	result := ExpandSource(input)

	// + at line start should become middleware
	if !strings.Contains(result, "middleware auth(jwt)") {
		t.Errorf("+ at line start should become 'middleware', got: %s", result)
	}
	if !strings.Contains(result, "middleware ratelimit(100/min)") {
		t.Errorf("+ at line start should become 'middleware', got: %s", result)
	}
}

func TestAllKeywords(t *testing.T) {
	tests := []struct {
		compact  string
		expanded string
	}{
		{"@ GET /", "route GET /"},
		{": User {", "type User {"},
		{"$ x = 1", "let x = 1"},
		{"> x", "return x"},
		{"+ auth()", "middleware auth()"},
		{"% db: Database", "use db: Database"},
		{"~ \"event\" {", "handle \"event\" {"},
		{"* \"0 * * * *\" task {", "cron \"0 * * * *\" task {"},
		{"! cmd {", "command cmd {"},
		{"& \"queue\" {", "queue \"queue\" {"},
	}

	for _, tt := range tests {
		result := ExpandSource(tt.compact)
		if !strings.Contains(result, tt.expanded) {
			t.Errorf("ExpandSource(%q) should contain %q, got: %s", tt.compact, tt.expanded, result)
		}

		// Test reverse
		result2 := CompactSource(tt.expanded)
		if !strings.Contains(result2, tt.compact) {
			t.Errorf("CompactSource(%q) should contain %q, got: %s", tt.expanded, tt.compact, result2)
		}
	}
}

func TestStringsNotTransformed(t *testing.T) {
	input := `@ GET / {
  > {message: "Use @ for route and $ for let"}
}`

	result := ExpandSource(input)

	// Symbols inside strings should NOT be transformed
	if !strings.Contains(result, `"Use @ for route and $ for let"`) {
		t.Errorf("Symbols in strings should not be transformed, got: %s", result)
	}
}
