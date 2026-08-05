package validate

import (
	"strings"
	"testing"
)

// TestValidateRejectsRedeclaration covers the gap that let broken examples ship:
// the compiler rejects a variable declared twice in one scope, but validation
// stopped before that pass, so a file validated clean and then failed on the
// first request to the route.
func TestValidateRejectsRedeclaration(t *testing.T) {
	src := `@ GET /api/count {
  $ total = 0
  $ total = 1
  > {total: total}
}`

	result := NewValidator(src, "test.glyph").Validate()

	if result.Valid {
		t.Fatal("redeclaring a variable in the same scope must not validate")
	}

	var found *ValidationError
	for _, e := range result.Errors {
		if strings.Contains(e.Message, "redeclare variable") {
			found = e
			break
		}
	}
	if found == nil {
		t.Fatalf("expected a redeclaration error, got %+v", result.Errors)
	}
	// The hint carries the fix, since the notation spec left the reassignment
	// form implicit and an agent reading only the error needs to know it.
	if !strings.Contains(found.FixHint, "name = expr") {
		t.Errorf("fix hint should show the reassignment form, got %q", found.FixHint)
	}
	if found.RelatedTo != "route GET /api/count" {
		t.Errorf("error should name the route, got %q", found.RelatedTo)
	}
}

// TestValidateAcceptsReassignment is the other half: the bare form is how a
// running total is written, and it must stay valid.
func TestValidateAcceptsReassignment(t *testing.T) {
	src := `@ GET /api/count {
  $ total = 0
  total = total + 1
  > {total: total}
}`

	result := NewValidator(src, "test.glyph").Validate()

	if !result.Valid {
		t.Fatalf("reassignment must validate, got errors: %+v", result.Errors)
	}
}

// TestValidateRejectsPathParamShadowing covers the second form the compiler
// rejects: a path parameter is already bound by the route pattern.
func TestValidateRejectsPathParamShadowing(t *testing.T) {
	src := `@ GET /api/factorial/:n {
  $ n = parseInt(n)
  > {n: n}
}`

	result := NewValidator(src, "test.glyph").Validate()

	if result.Valid {
		t.Fatal("shadowing a path parameter must not validate")
	}
	for _, e := range result.Errors {
		if strings.Contains(e.Message, "redeclare path parameter") {
			return
		}
	}
	t.Fatalf("expected a path parameter error, got %+v", result.Errors)
}
