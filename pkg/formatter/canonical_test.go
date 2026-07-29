package formatter

import (
	"strings"
	"testing"
)

func TestCanonicalizeSource_Reindent(t *testing.T) {
	input := "@ GET /users -> List[User] {\n\tdb <- Database\n      > db.users.Find()\n}\n"
	expected := "@ GET /users -> List[User] {\n  db <- Database\n  > db.users.Find()\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_NestedBlocks(t *testing.T) {
	input := ": User {\nid: int!\nprofile: {\nname: str!\n}\n}\n"
	expected := ": User {\n  id: int!\n  profile: {\n    name: str!\n  }\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_CommentsPreserved(t *testing.T) {
	input := "# top comment with { brace\n@ GET /x {\n// inner comment\n$ y = 1  # trailing comment\n}\n"
	expected := "# top comment with { brace\n@ GET /x {\n  // inner comment\n  $ y = 1  # trailing comment\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_StringsUntouched(t *testing.T) {
	input := "@ GET /x {\n$ s = \"a { brace # hash // slash\"\n$ q = 'don\\'t } close'\n}\n"
	expected := "@ GET /x {\n  $ s = \"a { brace # hash // slash\"\n  $ q = 'don\\'t } close'\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_BlankLines(t *testing.T) {
	input := "\n\n@ GET /a {\n> 1\n}\n\n\n\n@ GET /b {\n> 2\n}\n\n\n"
	expected := "@ GET /a {\n  > 1\n}\n\n@ GET /b {\n  > 2\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_StripsBOM(t *testing.T) {
	input := "\ufeff# comment\n$ x = 1\n"
	expected := "# comment\n$ x = 1\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_CRLF(t *testing.T) {
	input := "@ GET /x {\r\n> 1\r\n}\r\n"
	expected := "@ GET /x {\n  > 1\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_TrailingWhitespaceAndNewline(t *testing.T) {
	input := "$ x = 1   \n$ y = 2\t"
	expected := "$ x = 1\n$ y = 2\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_EmptyFile(t *testing.T) {
	if got := CanonicalizeSource(""); got != "" {
		t.Errorf("empty input should stay empty, got %q", got)
	}
	if got := CanonicalizeSource("\n\n\n"); got != "" {
		t.Errorf("blank-only input should become empty, got %q", got)
	}
}

func TestCanonicalizeSource_CloserDedents(t *testing.T) {
	input := "@ GET /x {\nif ok {\n> 1\n} else {\n> 2\n}\n}\n"
	expected := "@ GET /x {\n  if ok {\n    > 1\n  } else {\n    > 2\n  }\n}\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_UnbalancedClampsAtZero(t *testing.T) {
	input := "}\n}\n$ x = 1\n"
	expected := "}\n}\n$ x = 1\n"
	if got := CanonicalizeSource(input); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCanonicalizeSource_Idempotent(t *testing.T) {
	inputs := []string{
		"@ GET /users -> List[User] {\n\tdb <- Database\n> db.users.Find()\n}\n",
		"# comment\n\n\n: User {\nid: int!\n}\n",
		"@ GET /x {\r\n$ s = \"str with } brace\"\r\n}\r\n",
	}
	for _, input := range inputs {
		once := CanonicalizeSource(input)
		twice := CanonicalizeSource(once)
		if once != twice {
			t.Errorf("not idempotent for %q:\nonce:  %q\ntwice: %q", input, once, twice)
		}
	}
}

func TestCanonicalizeSource_OneLineChangeOneLineDiff(t *testing.T) {
	base := "@ GET /a {\n  > 1\n}\n\n@ GET /b {\n  > 2\n}\n"
	edited := strings.Replace(base, "> 2", "> 3", 1)
	gotBase := CanonicalizeSource(base)
	gotEdited := CanonicalizeSource(edited)
	baseLines := strings.Split(gotBase, "\n")
	editedLines := strings.Split(gotEdited, "\n")
	if len(baseLines) != len(editedLines) {
		t.Fatalf("line counts differ: %d vs %d", len(baseLines), len(editedLines))
	}
	diff := 0
	for i := range baseLines {
		if baseLines[i] != editedLines[i] {
			diff++
		}
	}
	if diff != 1 {
		t.Errorf("expected exactly 1 differing line, got %d", diff)
	}
}
