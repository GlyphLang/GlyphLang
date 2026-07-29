package formatter

import (
	"strings"
)

// CanonicalizeSource applies the canonical GlyphLang formatting rules to source
// code. It is deliberately zero-option and line-preserving: content within a
// line is never reflowed or reordered, so a one-line change stays a one-line
// diff. Comments and strings are preserved byte-for-byte.
//
// Rules:
//  1. Line endings are normalized to LF.
//  2. Trailing whitespace is stripped from every line.
//  3. Indentation is exactly 2 spaces per bracket depth ({, [, ( open; ), ], }
//     close). A line whose first character is a closer indents at depth-1.
//  4. At most one consecutive blank line; leading blank lines are removed.
//  5. The file ends with exactly one newline (an empty file stays empty).
func CanonicalizeSource(source string) string {
	source = strings.TrimPrefix(source, "\ufeff") // strip UTF-8 BOM; the lexer rejects it
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")

	lines := strings.Split(source, "\n")
	var out []string
	depth := 0
	blankRun := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			blankRun++
			// Rule 4: no leading blanks, max one consecutive blank.
			if len(out) > 0 && blankRun == 1 {
				out = append(out, "")
			}
			continue
		}
		blankRun = 0

		opens, closes, leadingCloses := scanBrackets(trimmed)

		indent := depth
		if leadingCloses > 0 {
			indent -= leadingCloses
		}
		if indent < 0 {
			indent = 0
		}

		out = append(out, strings.Repeat("  ", indent)+trimmed)

		depth += opens - closes
		if depth < 0 {
			depth = 0
		}
	}

	// Drop a trailing blank line left by rule 4.
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}

	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n") + "\n"
}

// scanBrackets counts bracket opens and closes on a single line, ignoring
// brackets inside string literals and comments. leadingCloses is the number of
// closing brackets before any other content, which dictates the dedent of the
// line itself. GlyphLang strings and comments never span lines, so no state
// carries between lines.
func scanBrackets(line string) (opens, closes, leadingCloses int) {
	leading := true
	i := 0
	n := len(line)
	for i < n {
		ch := line[i]

		// Line comments run to end of line.
		if ch == '#' || (ch == '/' && i+1 < n && line[i+1] == '/') {
			break
		}

		// Skip string literals, honoring escapes.
		if ch == '"' || ch == '\'' {
			quote := ch
			i++
			for i < n && line[i] != quote {
				if line[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			if i < n {
				i++
			}
			leading = false
			continue
		}

		switch ch {
		case '{', '[', '(':
			opens++
			leading = false
		case '}', ']', ')':
			closes++
			if leading {
				leadingCloses++
			}
		default:
			if ch != ' ' && ch != '\t' {
				leading = false
			}
		}
		i++
	}
	return opens, closes, leadingCloses
}
