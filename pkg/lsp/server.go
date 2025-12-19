package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Server represents the LSP server
type Server struct {
	reader          *bufio.Reader
	writer          io.Writer
	docManager      *DocumentManager
	clientCaps      *ClientCapabilities
	initialized     bool
	shutdownRequest bool
	logger          *log.Logger
	mu              sync.Mutex
}

// NewServer creates a new LSP server
func NewServer(reader io.Reader, writer io.Writer, logFile string) *Server {
	var logger *log.Logger
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger = log.New(f, "[LSP] ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	return &Server{
		reader:     bufio.NewReader(reader),
		writer:     writer,
		docManager: NewDocumentManager(),
		logger:     logger,
	}
}

// Start starts the LSP server
func (s *Server) Start() error {
	s.logger.Println("LSP Server starting...")

	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				s.logger.Println("Client disconnected")
				return nil
			}
			s.logger.Printf("Error reading message: %v", err)
			return err
		}

		if s.shutdownRequest {
			s.logger.Println("Shutdown requested, exiting")
			return nil
		}

		// Handle message
		if err := s.handleMessage(msg); err != nil {
			s.logger.Printf("Error handling message: %v", err)
		}
	}
}

// readMessage reads a JSON-RPC message from the client
func (s *Server) readMessage() (json.RawMessage, error) {
	// Read headers
	headers := make(map[string]string)
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Get content length
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %v", err)
	}

	// Read content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, content); err != nil {
		return nil, err
	}

	s.logger.Printf("Received: %s", string(content))

	return json.RawMessage(content), nil
}

// writeMessage writes a JSON-RPC message to the client
func (s *Server) writeMessage(msg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	s.logger.Printf("Sending: %s", string(content))

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	if _, err := s.writer.Write([]byte(header)); err != nil {
		return err
	}

	if _, err := s.writer.Write(content); err != nil {
		return err
	}

	return nil
}

// handleMessage dispatches a message to the appropriate handler
func (s *Server) handleMessage(msg json.RawMessage) error {
	// Try to parse as request
	var req Request
	if err := json.Unmarshal(msg, &req); err == nil && req.Method != "" {
		return s.handleRequest(&req)
	}

	// Try to parse as notification
	var notif Notification
	if err := json.Unmarshal(msg, &notif); err == nil && notif.Method != "" {
		return s.handleNotification(&notif)
	}

	return fmt.Errorf("unknown message type")
}

// handleRequest handles a JSON-RPC request
func (s *Server) handleRequest(req *Request) error {
	s.logger.Printf("Handling request: %s", req.Method)

	var result interface{}
	var err error

	switch req.Method {
	case "initialize":
		result, err = s.handleInitialize(req.Params)
	case "shutdown":
		result, err = s.handleShutdown()
	case "textDocument/hover":
		result, err = s.handleHover(req.Params)
	case "textDocument/completion":
		result, err = s.handleCompletion(req.Params)
	case "textDocument/definition":
		result, err = s.handleDefinition(req.Params)
	case "textDocument/references":
		result, err = s.handleReferences(req.Params)
	case "textDocument/documentSymbol":
		result, err = s.handleDocumentSymbol(req.Params)
	default:
		err = fmt.Errorf("method not found: %s", req.Method)
	}

	// Send response
	var resp Response
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	if err != nil {
		resp.Error = &RPCError{
			Code:    MethodNotFound,
			Message: err.Error(),
		}
	} else {
		resp.Result = result
	}

	return s.writeMessage(&resp)
}

// handleNotification handles a JSON-RPC notification
func (s *Server) handleNotification(notif *Notification) error {
	s.logger.Printf("Handling notification: %s", notif.Method)

	switch notif.Method {
	case "initialized":
		return s.handleInitialized()
	case "textDocument/didOpen":
		return s.handleDidOpen(notif.Params)
	case "textDocument/didChange":
		return s.handleDidChange(notif.Params)
	case "textDocument/didClose":
		return s.handleDidClose(notif.Params)
	case "exit":
		os.Exit(0)
	default:
		s.logger.Printf("Unknown notification: %s", notif.Method)
	}

	return nil
}

// LSP Request Handlers

func (s *Server) handleInitialize(params json.RawMessage) (*InitializeResult, error) {
	var initParams InitializeParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		return nil, err
	}

	s.clientCaps = &initParams.Capabilities
	s.initialized = true

	s.logger.Printf("Client: %v", initParams.ClientInfo)
	s.logger.Printf("Root URI: %s", initParams.RootURI)

	return &InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    TextDocumentSyncKindFull,
				Save: &SaveOptions{
					IncludeText: false,
				},
			},
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", ":", "@", "/", "!", "*", "~", "&"},
				ResolveProvider:   false,
			},
			HoverProvider:          true,
			DefinitionProvider:     true,
			ReferencesProvider:     true,
			DocumentSymbolProvider: true,
		},
		ServerInfo: &ServerInfo{
			Name:    "glyph-lsp",
			Version: "0.1.0",
		},
	}, nil
}

func (s *Server) handleShutdown() (interface{}, error) {
	s.logger.Println("Shutdown requested")
	s.shutdownRequest = true
	return nil, nil
}

func (s *Server) handleInitialized() error {
	s.logger.Println("Client initialized")
	return nil
}

func (s *Server) handleHover(params json.RawMessage) (*Hover, error) {
	var hoverParams TextDocumentPositionParams
	if err := json.Unmarshal(params, &hoverParams); err != nil {
		return nil, err
	}

	doc, exists := s.docManager.Get(hoverParams.TextDocument.URI)
	if !exists {
		return nil, nil
	}

	return GetHover(doc, hoverParams.Position), nil
}

func (s *Server) handleCompletion(params json.RawMessage) ([]CompletionItem, error) {
	var compParams TextDocumentPositionParams
	if err := json.Unmarshal(params, &compParams); err != nil {
		return nil, err
	}

	doc, exists := s.docManager.Get(compParams.TextDocument.URI)
	if !exists {
		return []CompletionItem{}, nil
	}

	return GetCompletion(doc, compParams.Position), nil
}

func (s *Server) handleDefinition(params json.RawMessage) ([]Location, error) {
	var defParams TextDocumentPositionParams
	if err := json.Unmarshal(params, &defParams); err != nil {
		return nil, err
	}

	doc, exists := s.docManager.Get(defParams.TextDocument.URI)
	if !exists {
		return []Location{}, nil
	}

	return GetDefinition(doc, defParams.Position), nil
}

func (s *Server) handleDocumentSymbol(params json.RawMessage) ([]DocumentSymbol, error) {
	var symbolParams DocumentSymbolParams
	if err := json.Unmarshal(params, &symbolParams); err != nil {
		return nil, err
	}

	doc, exists := s.docManager.Get(symbolParams.TextDocument.URI)
	if !exists {
		return []DocumentSymbol{}, nil
	}

	return GetDocumentSymbols(doc), nil
}

func (s *Server) handleReferences(params json.RawMessage) ([]Location, error) {
	var refParams ReferenceParams
	if err := json.Unmarshal(params, &refParams); err != nil {
		return nil, err
	}

	doc, exists := s.docManager.Get(refParams.TextDocument.URI)
	if !exists {
		return []Location{}, nil
	}

	return GetReferences(doc, refParams.Position, refParams.Context.IncludeDeclaration), nil
}

// LSP Notification Handlers

func (s *Server) handleDidOpen(params json.RawMessage) error {
	var openParams DidOpenTextDocumentParams
	if err := json.Unmarshal(params, &openParams); err != nil {
		return err
	}

	doc, err := s.docManager.Open(
		openParams.TextDocument.URI,
		openParams.TextDocument.Version,
		openParams.TextDocument.Text,
	)
	if err != nil {
		return err
	}

	// Publish diagnostics
	return s.publishDiagnostics(doc)
}

func (s *Server) handleDidChange(params json.RawMessage) error {
	var changeParams DidChangeTextDocumentParams
	if err := json.Unmarshal(params, &changeParams); err != nil {
		return err
	}

	doc, err := s.docManager.Update(
		changeParams.TextDocument.URI,
		changeParams.TextDocument.Version,
		changeParams.ContentChanges,
	)
	if err != nil {
		return err
	}

	// Publish diagnostics
	return s.publishDiagnostics(doc)
}

func (s *Server) handleDidClose(params json.RawMessage) error {
	var closeParams DidCloseTextDocumentParams
	if err := json.Unmarshal(params, &closeParams); err != nil {
		return err
	}

	return s.docManager.Close(closeParams.TextDocument.URI)
}

// Helper methods

func (s *Server) publishDiagnostics(doc *Document) error {
	diagnostics := GetDiagnostics(doc)

	notification := Notification{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
	}

	diagParams := PublishDiagnosticsParams{
		URI:         doc.URI,
		Version:     doc.Version,
		Diagnostics: diagnostics,
	}

	paramsJSON, err := json.Marshal(diagParams)
	if err != nil {
		return err
	}

	notification.Params = paramsJSON

	return s.writeMessage(&notification)
}
