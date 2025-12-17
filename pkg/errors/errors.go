package errors

import (
	"fmt"
)

// FormatError formats any error with colors and context
// This is the main public interface for error formatting
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Handle our custom error types
	switch e := err.(type) {
	case *CompileError:
		return e.FormatError(true)
	case *RuntimeError:
		return e.FormatError(true)
	default:
		// For unknown error types, just return the error message
		return fmt.Sprintf("%sError:%s %s\n", Bold+Red, Reset, err.Error())
	}
}

// WithLineInfo wraps an error with line and column information
func WithLineInfo(err error, line, col int, source string) error {
	if err == nil {
		return nil
	}

	// Extract source snippet around the line
	snippet := ExtractSourceSnippet(source, line)

	// If it's already a CompileError, update it
	if ce, ok := err.(*CompileError); ok {
		ce.Line = line
		ce.Column = col
		ce.SourceSnippet = snippet
		return ce
	}

	// Otherwise, create a new CompileError
	return &CompileError{
		Message:       err.Error(),
		Line:          line,
		Column:        col,
		SourceSnippet: snippet,
		ErrorType:     "Error",
	}
}

// WithSuggestion wraps an error with a helpful suggestion
func WithSuggestion(err error, suggestion string) error {
	if err == nil {
		return nil
	}

	// If it's a CompileError, add the suggestion
	if ce, ok := err.(*CompileError); ok {
		ce.Suggestion = suggestion
		return ce
	}

	// If it's a RuntimeError, add the suggestion
	if re, ok := err.(*RuntimeError); ok {
		re.Suggestion = suggestion
		return re
	}

	// For other errors, create a CompileError with the suggestion
	return &CompileError{
		Message:    err.Error(),
		Suggestion: suggestion,
		ErrorType:  "Error",
	}
}

// WithFileName adds a filename to an error
func WithFileName(err error, fileName string) error {
	if err == nil {
		return nil
	}

	// If it's a CompileError, add the filename
	if ce, ok := err.(*CompileError); ok {
		ce.FileName = fileName
		return ce
	}

	// For other errors, create a CompileError with the filename
	return &CompileError{
		Message:   err.Error(),
		FileName:  fileName,
		ErrorType: "Error",
	}
}

// Wrap wraps an error with complete context information
func Wrap(err error, line, col int, source, fileName, suggestion string) error {
	if err == nil {
		return nil
	}

	snippet := ExtractSourceSnippet(source, line)

	return &CompileError{
		Message:       err.Error(),
		Line:          line,
		Column:        col,
		SourceSnippet: snippet,
		FileName:      fileName,
		Suggestion:    suggestion,
		ErrorType:     "Error",
	}
}
