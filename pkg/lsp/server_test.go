package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Test helpers

type testClient struct {
	input  *bytes.Buffer
	output *bytes.Buffer
}

func newTestClient() *testClient {
	return &testClient{
		input:  &bytes.Buffer{},
		output: &bytes.Buffer{},
	}
}

func (tc *testClient) sendRequest(id int, method string, params interface{}) {
	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}

	if params != nil {
		paramsJSON, _ := json.Marshal(params)
		req.Params = paramsJSON
	}

	tc.writeMessage(req)
}

func (tc *testClient) sendNotification(method string, params interface{}) {
	notif := Notification{
		JSONRPC: "2.0",
		Method:  method,
	}

	if params != nil {
		paramsJSON, _ := json.Marshal(params)
		notif.Params = paramsJSON
	}

	tc.writeMessage(notif)
}

func (tc *testClient) writeMessage(msg interface{}) {
	content, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	tc.input.WriteString(header)
	tc.input.Write(content)
}

func (tc *testClient) readResponse() (*Response, error) {
	// Read headers
	for {
		line, err := tc.output.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "" {
			break
		}
	}

	// For now, just parse the response manually
	// In a real test, we'd parse the Content-Length and read exact bytes
	decoder := json.NewDecoder(tc.output)
	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Protocol Tests

func TestServerInitialize(t *testing.T) {
	tc := newTestClient()
	server := NewServer(tc.input, tc.output, "")

	// Send initialize request
	initParams := InitializeParams{
		ProcessID: 1234,
		RootURI:   "file:///test",
		Capabilities: ClientCapabilities{
			TextDocument: &TextDocumentClientCapabilities{},
		},
	}

	tc.sendRequest(1, "initialize", initParams)

	// Manually handle the message
	msg, err := server.readMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var req Request
	if err := json.Unmarshal(msg, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if req.Method != "initialize" {
		t.Errorf("Expected method 'initialize', got '%s'", req.Method)
	}

	// JSON unmarshals numbers as float64, so compare with float64(1)
	if req.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", req.ID)
	}
}

func TestServerShutdown(t *testing.T) {
	tc := newTestClient()
	server := NewServer(tc.input, tc.output, "")

	// Send shutdown request
	tc.sendRequest(2, "shutdown", nil)

	msg, err := server.readMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var req Request
	if err := json.Unmarshal(msg, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if req.Method != "shutdown" {
		t.Errorf("Expected method 'shutdown', got '%s'", req.Method)
	}

	// Handle shutdown
	server.handleRequest(&req)

	if !server.shutdownRequest {
		t.Error("Server should be marked for shutdown")
	}
}

func TestMessageParsing(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantType string
	}{
		{
			name:     "Request",
			message:  `{"jsonrpc":"2.0","id":1,"method":"test"}`,
			wantType: "request",
		},
		{
			name:     "Notification",
			message:  `{"jsonrpc":"2.0","method":"test"}`,
			wantType: "notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req Request
			var notif Notification

			// Try to unmarshal as request first - if ID is present, it's a request
			if err := json.Unmarshal([]byte(tt.message), &req); err == nil && req.Method != "" && req.ID != nil {
				if tt.wantType != "request" {
					t.Errorf("Expected %s, got request", tt.wantType)
				}
			} else if err := json.Unmarshal([]byte(tt.message), &notif); err == nil && notif.Method != "" {
				if tt.wantType != "notification" {
					t.Errorf("Expected %s, got notification", tt.wantType)
				}
			}
		})
	}
}

func TestRPCError(t *testing.T) {
	err := &RPCError{
		Code:    MethodNotFound,
		Message: "Method not found",
	}

	if err.Code != MethodNotFound {
		t.Errorf("Expected code %d, got %d", MethodNotFound, err.Code)
	}

	if err.Message != "Method not found" {
		t.Errorf("Expected message 'Method not found', got '%s'", err.Message)
	}
}

func TestResponseWithError(t *testing.T) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      1,
		Error: &RPCError{
			Code:    InternalError,
			Message: "Internal error",
		},
	}

	if resp.Error == nil {
		t.Fatal("Expected error to be set")
	}

	if resp.Error.Code != InternalError {
		t.Errorf("Expected code %d, got %d", InternalError, resp.Error.Code)
	}

	if resp.Result != nil {
		t.Error("Expected result to be nil when error is set")
	}
}

func TestResponseWithResult(t *testing.T) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]string{"status": "ok"},
	}

	if resp.Result == nil {
		t.Fatal("Expected result to be set")
	}

	if resp.Error != nil {
		t.Error("Expected error to be nil when result is set")
	}
}

func TestInitializeParams(t *testing.T) {
	params := InitializeParams{
		ProcessID: 1234,
		RootURI:   "file:///workspace",
		ClientInfo: &ClientInfo{
			Name:    "Test Client",
			Version: "1.0.0",
		},
		Capabilities: ClientCapabilities{
			TextDocument: &TextDocumentClientCapabilities{
				Synchronization: &TextDocumentSyncClientCapabilities{
					DynamicRegistration: true,
				},
			},
		},
	}

	if params.ProcessID != 1234 {
		t.Errorf("Expected ProcessID 1234, got %d", params.ProcessID)
	}

	if params.RootURI != "file:///workspace" {
		t.Errorf("Expected RootURI 'file:///workspace', got '%s'", params.RootURI)
	}

	if params.ClientInfo == nil {
		t.Fatal("Expected ClientInfo to be set")
	}

	if params.ClientInfo.Name != "Test Client" {
		t.Errorf("Expected client name 'Test Client', got '%s'", params.ClientInfo.Name)
	}
}

func TestInitializeResult(t *testing.T) {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    TextDocumentSyncKindFull,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
		},
		ServerInfo: &ServerInfo{
			Name:    "Test Server",
			Version: "1.0.0",
		},
	}

	if result.Capabilities.TextDocumentSync == nil {
		t.Fatal("Expected TextDocumentSync to be set")
	}

	if !result.Capabilities.HoverProvider {
		t.Error("Expected HoverProvider to be true")
	}

	if !result.Capabilities.DefinitionProvider {
		t.Error("Expected DefinitionProvider to be true")
	}

	if result.ServerInfo == nil {
		t.Fatal("Expected ServerInfo to be set")
	}

	if result.ServerInfo.Name != "Test Server" {
		t.Errorf("Expected server name 'Test Server', got '%s'", result.ServerInfo.Name)
	}
}

func TestTextDocumentItem(t *testing.T) {
	doc := TextDocumentItem{
		URI:        "file:///test.glyph",
		LanguageID: "glyph",
		Version:    1,
		Text:       ": User { name: str! }",
	}

	if doc.URI != "file:///test.glyph" {
		t.Errorf("Expected URI 'file:///test.glyph', got '%s'", doc.URI)
	}

	if doc.LanguageID != "glyph" {
		t.Errorf("Expected languageId 'glyph', got '%s'", doc.LanguageID)
	}

	if doc.Version != 1 {
		t.Errorf("Expected version 1, got %d", doc.Version)
	}
}

func TestPosition(t *testing.T) {
	pos := Position{Line: 5, Character: 10}

	if pos.Line != 5 {
		t.Errorf("Expected line 5, got %d", pos.Line)
	}

	if pos.Character != 10 {
		t.Errorf("Expected character 10, got %d", pos.Character)
	}
}

func TestRange(t *testing.T) {
	r := Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: 0, Character: 10},
	}

	if r.Start.Line != 0 || r.Start.Character != 0 {
		t.Error("Start position incorrect")
	}

	if r.End.Line != 0 || r.End.Character != 10 {
		t.Error("End position incorrect")
	}
}

func TestLocation(t *testing.T) {
	loc := Location{
		URI: "file:///test.glyph",
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 0, Character: 10},
		},
	}

	if loc.URI != "file:///test.glyph" {
		t.Errorf("Expected URI 'file:///test.glyph', got '%s'", loc.URI)
	}
}

func TestDiagnostic(t *testing.T) {
	diag := Diagnostic{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 0, Character: 5},
		},
		Severity: DiagnosticSeverityError,
		Source:   "glyph",
		Message:  "Syntax error",
	}

	if diag.Severity != DiagnosticSeverityError {
		t.Errorf("Expected severity Error, got %d", diag.Severity)
	}

	if diag.Source != "glyph" {
		t.Errorf("Expected source 'glyph', got '%s'", diag.Source)
	}

	if diag.Message != "Syntax error" {
		t.Errorf("Expected message 'Syntax error', got '%s'", diag.Message)
	}
}

func TestCompletionItem(t *testing.T) {
	item := CompletionItem{
		Label:  "route",
		Kind:   CompletionItemKindKeyword,
		Detail: "Route definition",
	}

	if item.Label != "route" {
		t.Errorf("Expected label 'route', got '%s'", item.Label)
	}

	if item.Kind != CompletionItemKindKeyword {
		t.Errorf("Expected kind Keyword, got %d", item.Kind)
	}
}

func TestHover(t *testing.T) {
	hover := Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: "**route** - Defines an HTTP route",
		},
	}

	if hover.Contents.Kind != "markdown" {
		t.Errorf("Expected kind 'markdown', got '%s'", hover.Contents.Kind)
	}

	if hover.Contents.Value == "" {
		t.Error("Expected non-empty content")
	}
}

func TestSymbolInformation(t *testing.T) {
	sym := SymbolInformation{
		Name: "User",
		Kind: SymbolKindStruct,
		Location: Location{
			URI: "file:///test.glyph",
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 5, Character: 0},
			},
		},
	}

	if sym.Name != "User" {
		t.Errorf("Expected name 'User', got '%s'", sym.Name)
	}

	if sym.Kind != SymbolKindStruct {
		t.Errorf("Expected kind Struct, got %d", sym.Kind)
	}
}

func TestDocumentSymbol(t *testing.T) {
	sym := DocumentSymbol{
		Name:   "User",
		Kind:   SymbolKindStruct,
		Detail: "Type definition",
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 5, Character: 0},
		},
		SelectionRange: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 0, Character: 4},
		},
		Children: []DocumentSymbol{
			{
				Name:   "name",
				Kind:   SymbolKindField,
				Detail: "str",
				Range: Range{
					Start: Position{Line: 1, Character: 2},
					End:   Position{Line: 1, Character: 15},
				},
				SelectionRange: Range{
					Start: Position{Line: 1, Character: 2},
					End:   Position{Line: 1, Character: 6},
				},
			},
		},
	}

	if sym.Name != "User" {
		t.Errorf("Expected name 'User', got '%s'", sym.Name)
	}

	if len(sym.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(sym.Children))
	}

	if sym.Children[0].Name != "name" {
		t.Errorf("Expected child name 'name', got '%s'", sym.Children[0].Name)
	}
}

func TestDidOpenParams(t *testing.T) {
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.glyph",
			LanguageID: "glyph",
			Version:    1,
			Text:       ": User { name: str! }",
		},
	}

	if params.TextDocument.URI != "file:///test.glyph" {
		t.Errorf("Expected URI 'file:///test.glyph', got '%s'", params.TextDocument.URI)
	}

	if params.TextDocument.Version != 1 {
		t.Errorf("Expected version 1, got %d", params.TextDocument.Version)
	}
}

func TestDidChangeParams(t *testing.T) {
	params := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{
				URI: "file:///test.glyph",
			},
			Version: 2,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{
				Text: ": User { name: str!, age: int! }",
			},
		},
	}

	if params.TextDocument.Version != 2 {
		t.Errorf("Expected version 2, got %d", params.TextDocument.Version)
	}

	if len(params.ContentChanges) != 1 {
		t.Errorf("Expected 1 change, got %d", len(params.ContentChanges))
	}
}

func TestDidCloseParams(t *testing.T) {
	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{
			URI: "file:///test.glyph",
		},
	}

	if params.TextDocument.URI != "file:///test.glyph" {
		t.Errorf("Expected URI 'file:///test.glyph', got '%s'", params.TextDocument.URI)
	}
}

func TestPublishDiagnosticsParams(t *testing.T) {
	params := PublishDiagnosticsParams{
		URI:     "file:///test.glyph",
		Version: 1,
		Diagnostics: []Diagnostic{
			{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 5},
				},
				Severity: DiagnosticSeverityError,
				Message:  "Syntax error",
			},
		},
	}

	if params.URI != "file:///test.glyph" {
		t.Errorf("Expected URI 'file:///test.glyph', got '%s'", params.URI)
	}

	if len(params.Diagnostics) != 1 {
		t.Errorf("Expected 1 diagnostic, got %d", len(params.Diagnostics))
	}
}
