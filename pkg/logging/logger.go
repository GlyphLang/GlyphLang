package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogFormat represents the output format for logs
type LogFormat int

const (
	// TextFormat outputs human-readable text logs
	TextFormat LogFormat = iota
	// JSONFormat outputs structured JSON logs
	JSONFormat
)

// LogEntry represents a single log entry with all metadata
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Caller     string                 `json:"caller,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	// MinLevel is the minimum level to log (default: INFO)
	MinLevel LogLevel
	// Format is the output format (default: TextFormat)
	Format LogFormat
	// IncludeCaller includes file and line number in logs
	IncludeCaller bool
	// IncludeStackTrace includes stack trace for ERROR and FATAL logs
	IncludeStackTrace bool
	// BufferSize is the size of the async log buffer (default: 1000)
	BufferSize int
	// Outputs are the writers to send logs to
	Outputs []io.Writer
	// MaxFileSize is the maximum size in bytes before rotation (0 = no rotation)
	MaxFileSize int64
	// MaxBackups is the maximum number of old log files to keep
	MaxBackups int
	// FilePath is the path to the log file (empty = no file logging)
	FilePath string
}

// Logger is the main logging instance
type Logger struct {
	config     LoggerConfig
	buffer     chan *LogEntry
	wg         sync.WaitGroup
	mu         sync.Mutex
	stopped    bool
	fileWriter *rotatingFileWriter
	// For Sync() support
	syncCh chan chan struct{}
}

// rotatingFileWriter handles log file rotation
type rotatingFileWriter struct {
	mu          sync.Mutex
	file        *os.File
	path        string
	size        int64
	maxSize     int64
	maxBackups  int
	currentSize int64
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config LoggerConfig) (*Logger, error) {
	// Set defaults
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if len(config.Outputs) == 0 {
		config.Outputs = []io.Writer{os.Stdout}
	}

	logger := &Logger{
		config:  config,
		buffer:  make(chan *LogEntry, config.BufferSize),
		stopped: false,
		syncCh:  make(chan chan struct{}, 1),
	}

	// Setup file rotation if file path is provided
	if config.FilePath != "" {
		fw, err := newRotatingFileWriter(config.FilePath, config.MaxFileSize, config.MaxBackups)
		if err != nil {
			return nil, fmt.Errorf("failed to create file writer: %w", err)
		}
		logger.fileWriter = fw
		logger.config.Outputs = append(logger.config.Outputs, fw)
	}

	// Start async log processor
	logger.wg.Add(1)
	go logger.processLogs()

	return logger, nil
}

// newRotatingFileWriter creates a new rotating file writer
func newRotatingFileWriter(path string, maxSize int64, maxBackups int) (*rotatingFileWriter, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	return &rotatingFileWriter{
		file:        file,
		path:        path,
		maxSize:     maxSize,
		maxBackups:  maxBackups,
		currentSize: info.Size(),
	}, nil
}

// Write implements io.Writer for rotatingFileWriter
func (w *rotatingFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation is needed
	if w.maxSize > 0 && w.currentSize+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

// rotate rotates the log file
func (w *rotatingFileWriter) rotate() error {
	// Close current file
	if err := w.file.Close(); err != nil {
		return err
	}

	// Rotate existing backup files
	for i := w.maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Rename current file to .1
	if err := os.Rename(w.path, fmt.Sprintf("%s.1", w.path)); err != nil {
		return err
	}

	// Create new file
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.currentSize = 0
	return nil
}

// Close closes the rotating file writer
func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

// processLogs processes log entries from the buffer asynchronously
func (l *Logger) processLogs() {
	defer l.wg.Done()

	for {
		select {
		case entry, ok := <-l.buffer:
			if !ok {
				// Buffer closed, drain any pending sync requests
				select {
				case done := <-l.syncCh:
					close(done)
				default:
				}
				return
			}
			l.writeLog(entry)
		case done := <-l.syncCh:
			// Drain all pending entries before signaling done
			draining := true
			for draining {
				select {
				case entry := <-l.buffer:
					l.writeLog(entry)
				default:
					draining = false
				}
			}
			close(done)
		}
	}
}

// writeLog writes a log entry to all outputs
func (l *Logger) writeLog(entry *LogEntry) {
	var output string

	if l.config.Format == JSONFormat {
		bytes, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
			return
		}
		output = string(bytes) + "\n"
	} else {
		// Text format
		output = l.formatTextLog(entry)
	}

	// Write to all outputs
	for _, w := range l.config.Outputs {
		if _, err := w.Write([]byte(output)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
		}
	}
}

// formatTextLog formats a log entry as human-readable text
func (l *Logger) formatTextLog(entry *LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, fmt.Sprintf("[%s]", entry.Level))

	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.RequestID))
	}

	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Caller))
	}

	parts = append(parts, entry.Message)

	// Add fields
	if len(entry.Fields) > 0 {
		fieldsStr := ""
		for k, v := range entry.Fields {
			if fieldsStr != "" {
				fieldsStr += ", "
			}
			fieldsStr += fmt.Sprintf("%s=%v", k, v)
		}
		parts = append(parts, fmt.Sprintf("{%s}", fieldsStr))
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}

	// Add stack trace if present
	if entry.StackTrace != "" {
		result += "\n" + entry.StackTrace
	}

	return result + "\n"
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}, requestID string) {
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	// Check if we should log this level
	if level < l.config.MinLevel {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   msg,
		RequestID: requestID,
		Fields:    fields,
	}

	// Add caller information if configured
	if l.config.IncludeCaller {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			entry.Caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		}
	}

	// Add stack trace for ERROR and FATAL if configured
	if l.config.IncludeStackTrace && (level == ERROR || level == FATAL) {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		entry.StackTrace = string(buf[:n])
	}

	// Send to buffer (non-blocking with fallback to direct write)
	select {
	case l.buffer <- entry:
		// Successfully buffered
	default:
		// Buffer full, write synchronously
		l.writeLog(entry)
	}

	// Exit on FATAL
	if level == FATAL {
		l.Close()
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg, nil, "")
}

// DebugWithFields logs a debug message with additional fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields, "")
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(INFO, msg, nil, "")
}

// InfoWithFields logs an info message with additional fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields, "")
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WARN, msg, nil, "")
}

// WarnWithFields logs a warning message with additional fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields, "")
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(ERROR, msg, nil, "")
}

// ErrorWithFields logs an error message with additional fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields, "")
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(msg string) {
	l.log(FATAL, msg, nil, "")
}

// FatalWithFields logs a fatal message with additional fields and exits
func (l *Logger) FatalWithFields(msg string, fields map[string]interface{}) {
	l.log(FATAL, msg, fields, "")
}

// Sync flushes all pending log entries and waits for them to be written.
// This is useful in tests to ensure all logs have been processed before
// reading from output buffers.
func (l *Logger) Sync() {
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	done := make(chan struct{})
	l.syncCh <- done
	<-done
}

// Close gracefully shuts down the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		return nil
	}
	l.stopped = true
	l.mu.Unlock()

	// Close buffer and wait for processing to complete
	close(l.buffer)
	l.wg.Wait()

	// Close file writer if present
	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}

	return nil
}

// WithRequestID creates a new ContextLogger with a request ID
func (l *Logger) WithRequestID(requestID string) *ContextLogger {
	return &ContextLogger{
		logger:    l,
		requestID: requestID,
		fields:    make(map[string]interface{}),
	}
}

// WithFields creates a new ContextLogger with fields
func (l *Logger) WithFields(fields map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger: l,
		fields: fields,
	}
}

// NewRequestID generates a new UUID for request tracking
func NewRequestID() string {
	return uuid.New().String()
}

// ContextLogger is a logger with pre-configured context (request ID and fields)
type ContextLogger struct {
	logger    *Logger
	requestID string
	fields    map[string]interface{}
	mu        sync.Mutex
}

// WithField adds a field to the context logger
func (cl *ContextLogger) WithField(key string, value interface{}) *ContextLogger {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	newFields := make(map[string]interface{}, len(cl.fields)+1)
	for k, v := range cl.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &ContextLogger{
		logger:    cl.logger,
		requestID: cl.requestID,
		fields:    newFields,
	}
}

// WithFields adds multiple fields to the context logger
func (cl *ContextLogger) WithFields(fields map[string]interface{}) *ContextLogger {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	newFields := make(map[string]interface{}, len(cl.fields)+len(fields))
	for k, v := range cl.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &ContextLogger{
		logger:    cl.logger,
		requestID: cl.requestID,
		fields:    newFields,
	}
}

// mergeFields merges the context fields with additional fields
func (cl *ContextLogger) mergeFields(additional map[string]interface{}) map[string]interface{} {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if additional == nil {
		return cl.fields
	}

	merged := make(map[string]interface{}, len(cl.fields)+len(additional))
	for k, v := range cl.fields {
		merged[k] = v
	}
	for k, v := range additional {
		merged[k] = v
	}
	return merged
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(msg string) {
	cl.logger.log(DEBUG, msg, cl.fields, cl.requestID)
}

// DebugWithFields logs a debug message with additional fields
func (cl *ContextLogger) DebugWithFields(msg string, fields map[string]interface{}) {
	cl.logger.log(DEBUG, msg, cl.mergeFields(fields), cl.requestID)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(msg string) {
	cl.logger.log(INFO, msg, cl.fields, cl.requestID)
}

// InfoWithFields logs an info message with additional fields
func (cl *ContextLogger) InfoWithFields(msg string, fields map[string]interface{}) {
	cl.logger.log(INFO, msg, cl.mergeFields(fields), cl.requestID)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(msg string) {
	cl.logger.log(WARN, msg, cl.fields, cl.requestID)
}

// WarnWithFields logs a warning message with additional fields
func (cl *ContextLogger) WarnWithFields(msg string, fields map[string]interface{}) {
	cl.logger.log(WARN, msg, cl.mergeFields(fields), cl.requestID)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(msg string) {
	cl.logger.log(ERROR, msg, cl.fields, cl.requestID)
}

// ErrorWithFields logs an error message with additional fields
func (cl *ContextLogger) ErrorWithFields(msg string, fields map[string]interface{}) {
	cl.logger.log(ERROR, msg, cl.mergeFields(fields), cl.requestID)
}

// Fatal logs a fatal message with context and exits
func (cl *ContextLogger) Fatal(msg string) {
	cl.logger.log(FATAL, msg, cl.fields, cl.requestID)
}

// FatalWithFields logs a fatal message with additional fields and exits
func (cl *ContextLogger) FatalWithFields(msg string, fields map[string]interface{}) {
	cl.logger.log(FATAL, msg, cl.mergeFields(fields), cl.requestID)
}

// Default logger instance
var defaultLogger *Logger
var defaultLoggerMu sync.Mutex

// InitDefaultLogger initializes the default logger with the given configuration
func InitDefaultLogger(config LoggerConfig) error {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()

	if defaultLogger != nil {
		defaultLogger.Close()
	}

	logger, err := NewLogger(config)
	if err != nil {
		return err
	}

	defaultLogger = logger
	return nil
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() *Logger {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()

	if defaultLogger == nil {
		// Create a basic default logger if none exists
		defaultLogger, _ = NewLogger(LoggerConfig{
			MinLevel: INFO,
			Format:   TextFormat,
		})
	}

	return defaultLogger
}

// Convenience functions for the default logger

// Debug logs a debug message using the default logger
func Debug(msg string) {
	GetDefaultLogger().Debug(msg)
}

// Info logs an info message using the default logger
func Info(msg string) {
	GetDefaultLogger().Info(msg)
}

// Warn logs a warning message using the default logger
func Warn(msg string) {
	GetDefaultLogger().Warn(msg)
}

// Error logs an error message using the default logger
func Error(msg string) {
	GetDefaultLogger().Error(msg)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(msg string) {
	GetDefaultLogger().Fatal(msg)
}

// WithRequestID creates a context logger with a request ID
func WithRequestID(requestID string) *ContextLogger {
	return GetDefaultLogger().WithRequestID(requestID)
}

// WithFields creates a context logger with fields
func WithFields(fields map[string]interface{}) *ContextLogger {
	return GetDefaultLogger().WithFields(fields)
}
