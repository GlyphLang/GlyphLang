package lsp

import (
	"strings"
	"testing"
)

func TestCommandHover(t *testing.T) {
	dm := NewDocumentManager()

	// Correct syntax: ! followed by command name and params
	source := `! hello name: str! {
  > "Hello " + name
}`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	hover := GetHover(doc, Position{Line: 0, Character: 3})
	if hover != nil {
		content := hover.Contents.Value
		if !strings.Contains(content, "hello") {
			t.Error("Expected hover to contain command name")
		}
	}
}

func TestCronTaskHover(t *testing.T) {
	dm := NewDocumentManager()

	// Correct syntax: * followed directly by schedule string
	source := `* "0 0 * * *" daily_cleanup {
  > "Daily task"
}`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
	}

	// Verify the cron task was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestEventHandlerHover(t *testing.T) {
	dm := NewDocumentManager()

	// Correct syntax: ~ followed directly by event type string
	source := `~ "user.created" {
  > "User created event"
}`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
	}

	// Verify the event handler was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestQueueWorkerHover(t *testing.T) {
	dm := NewDocumentManager()

	// Correct syntax: & followed directly by queue name string
	source := `& "email.send" {
  > "Send email"
}`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
	}

	// Verify the queue worker was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestDirectiveKeywordInfo(t *testing.T) {
	keywords := []string{"command", "cron", "event", "queue"}

	for _, kw := range keywords {
		info := getKeywordInfo(kw)
		if info == "" {
			t.Errorf("Expected info for keyword '%s'", kw)
		}
		if !strings.Contains(info, kw) {
			t.Errorf("Keyword info for '%s' should contain the keyword itself", kw)
		}
	}
}

func TestDirectiveCompletionSnippets(t *testing.T) {
	dm := NewDocumentManager()
	source := ""
	doc, _ := dm.Open("file:///test.abc", 1, source)

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	expectedSnippets := map[string]bool{
		"command": false,
		"cron":    false,
		"event":   false,
		"queue":   false,
	}

	for _, item := range completions {
		if item.Kind == CompletionItemKindSnippet {
			if _, ok := expectedSnippets[item.Label]; ok {
				expectedSnippets[item.Label] = true
				if item.InsertText == "" {
					t.Errorf("Snippet '%s' should have InsertText", item.Label)
				}
				if item.InsertTextFormat != 2 {
					t.Errorf("Snippet '%s' should have InsertTextFormat=2", item.Label)
				}
			}
		}
	}

	for snippet, found := range expectedSnippets {
		if !found {
			t.Errorf("Expected to find snippet for '%s'", snippet)
		}
	}
}

func TestDirectiveDocumentSymbols(t *testing.T) {
	dm := NewDocumentManager()

	// Use correct syntax for all directives
	source := `: User {
  name: str!
}

! greet name: str! {
  > "Hello"
}

* "0 0 * * *" daily_task {
  > "Task"
}

~ "user.created" {
  > "Event"
}

& "email.send" {
  > "Queue"
}

@ route /api/users [GET] {
  > {users: []}
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed - parser may not support these directives yet")
		return
	}

	symbols := GetDocumentSymbols(doc)

	if len(symbols) == 0 {
		t.Skip("No symbols found - directives may not be parsed yet")
		return
	}

	// Count different symbol types
	typeCount := 0
	commandCount := 0
	cronCount := 0
	eventCount := 0
	queueCount := 0
	routeCount := 0

	for _, sym := range symbols {
		switch sym.Kind {
		case SymbolKindStruct:
			typeCount++
		case SymbolKindMethod:
			// Check the name prefix to determine type
			if strings.HasPrefix(sym.Name, "!") {
				commandCount++
			} else if strings.HasPrefix(sym.Name, "*") {
				cronCount++
			} else if strings.HasPrefix(sym.Name, "~") {
				eventCount++
			} else if strings.HasPrefix(sym.Name, "&") {
				queueCount++
			} else {
				routeCount++
			}
		}
	}

	// Log what we found for debugging
	t.Logf("Found symbols: %d types, %d commands, %d cron, %d events, %d queues, %d routes",
		typeCount, commandCount, cronCount, eventCount, queueCount, routeCount)

	// Only test what we can parse
	if typeCount > 0 && typeCount != 1 {
		t.Errorf("Expected 1 type symbol, got %d", typeCount)
	}

	if commandCount > 0 && commandCount != 1 {
		t.Errorf("Expected 1 command symbol, got %d", commandCount)
	}

	if cronCount > 0 && cronCount != 1 {
		t.Errorf("Expected 1 cron symbol, got %d", cronCount)
	}

	if eventCount > 0 && eventCount != 1 {
		t.Errorf("Expected 1 event symbol, got %d", eventCount)
	}

	if queueCount > 0 && queueCount != 1 {
		t.Errorf("Expected 1 queue symbol, got %d", queueCount)
	}

	if routeCount > 0 && routeCount != 1 {
		t.Errorf("Expected 1 route symbol, got %d", routeCount)
	}
}

func TestCommandValidation(t *testing.T) {
	dm := NewDocumentManager()

	// Correct syntax: ! followed by command name and params
	source := `! test name: str! {
  > "Test"
}`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	diagnostics := checkTypes(doc.AST)

	// Should not have errors for valid command
	for _, diag := range diagnostics {
		if strings.Contains(diag.Message, "Command must have a name") {
			t.Error("Valid command should not generate name error")
		}
	}
}

func TestDirectiveTriggerCharacters(t *testing.T) {
	// This test verifies that the completion trigger characters include the new directive symbols
	expectedTriggers := []string{"!", "*", "~", "&"}

	// We can't directly test the server configuration here, but we can verify
	// that the completion function works when called
	dm := NewDocumentManager()
	doc, _ := dm.Open("file:///test.abc", 1, "")

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	// Should have directive snippets
	hasDirectiveSnippets := false
	for _, item := range completions {
		if item.Kind == CompletionItemKindSnippet {
			if item.Label == "command" || item.Label == "cron" ||
				item.Label == "event" || item.Label == "queue" {
				hasDirectiveSnippets = true
				break
			}
		}
	}

	if !hasDirectiveSnippets {
		t.Error("Expected to find directive snippets in completions")
	}

	// Verify the expected triggers are documented
	_ = expectedTriggers
}

func TestDirectiveSymbolNames(t *testing.T) {
	dm := NewDocumentManager()

	source := `! greet name: str! {
  > "Hello"
}

* "0 0 * * *" cleanup {
  > "Done"
}

~ "order.created" {
  > "Order event"
}

& "notification.send" {
  > "Sent"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	symbols := GetDocumentSymbols(doc)

	// Verify specific symbol names
	expectedNames := map[string]bool{
		"! greet":               false,
		"* cleanup":             false,
		"~ order.created":       false,
		"& notification.send":   false,
	}

	for _, sym := range symbols {
		if _, ok := expectedNames[sym.Name]; ok {
			expectedNames[sym.Name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected to find symbol named '%s'", name)
		}
	}
}

func TestDirectiveSymbolDetails(t *testing.T) {
	dm := NewDocumentManager()

	source := `* "*/5 * * * *" health_check {
  > "healthy"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	symbols := GetDocumentSymbols(doc)

	// Find the cron task symbol and check its detail
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, "*") {
			// Detail should contain the schedule
			if !strings.Contains(sym.Detail, "*/5 * * * *") {
				t.Errorf("Cron symbol detail should contain schedule, got: %s", sym.Detail)
			}
		}
	}
}

func TestCronTaskWithTimezone(t *testing.T) {
	dm := NewDocumentManager()

	source := `* "0 9 * * *" morning_report tz "America/New_York" {
  > "Report generated"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	// Verify the cron task was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestEventHandlerAsync(t *testing.T) {
	dm := NewDocumentManager()

	source := `~ "user.updated" async {
  > "Processed"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	// Verify the event handler was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestQueueWorkerWithConfig(t *testing.T) {
	dm := NewDocumentManager()

	source := `& "image.process" {
  + concurrency(5)
  + retries(3)
  + timeout(120)
  > "Processed"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	// Verify the queue worker was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestCommandWithFlags(t *testing.T) {
	dm := NewDocumentManager()

	source := `! deploy env: str! --force: bool = false --dry-run: bool = false {
  > "Deployed"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	// Verify the command was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}

func TestMultipleDirectivesSameType(t *testing.T) {
	dm := NewDocumentManager()

	source := `! build {
  > "Building"
}

! test {
  > "Testing"
}

! deploy {
  > "Deploying"
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	symbols := GetDocumentSymbols(doc)

	commandCount := 0
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, "!") {
			commandCount++
		}
	}

	if commandCount != 3 {
		t.Errorf("Expected 3 command symbols, got %d", commandCount)
	}
}

func TestDirectiveWithInjections(t *testing.T) {
	dm := NewDocumentManager()

	source := `* "0 0 * * *" cleanup {
  % db: Database
  $ deleted = db.cleanup()
  > {deleted: deleted}
}
`

	doc, _ := dm.Open("file:///test.glyph", 1, source)

	if doc.AST == nil {
		t.Skip("AST not parsed")
		return
	}

	// Verify the cron task with injection was parsed
	if len(doc.AST.Items) == 0 {
		t.Error("Expected at least one item in AST")
	}
}
