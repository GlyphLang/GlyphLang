package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse represents a standardized JSON error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HTTPError is the base interface for all HTTP errors
type HTTPError interface {
	error
	StatusCode() int
	ErrorType() string
	ToResponse() *ErrorResponse
}

// BaseError provides common functionality for all error types
type BaseError struct {
	Code   int
	Type   string
	Msg    string
	Detail string
	Cause  error
}

func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Msg, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

func (e *BaseError) StatusCode() int {
	return e.Code
}

func (e *BaseError) ErrorType() string {
	return e.Type
}

func (e *BaseError) ToResponse() *ErrorResponse {
	resp := &ErrorResponse{
		Status:  e.Code,
		Error:   e.Type,
		Message: e.Msg,
	}

	// Only include the explicit developer-set detail, never raw error internals
	if e.Detail != "" {
		resp.Details = e.Detail
	}

	return resp
}

// ValidationError represents a 400 Bad Request error
type ValidationError struct {
	*BaseError
	Field string
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		BaseError: &BaseError{
			Code: http.StatusBadRequest,
			Type: "ValidationError",
			Msg:  message,
		},
		Field: field,
	}
}

// NewValidationErrorWithDetails creates a validation error with additional details
func NewValidationErrorWithDetails(field, message, details string) *ValidationError {
	return &ValidationError{
		BaseError: &BaseError{
			Code:   http.StatusBadRequest,
			Type:   "ValidationError",
			Msg:    message,
			Detail: details,
		},
		Field: field,
	}
}

// ToResponse overrides the base method to include field information
func (e *ValidationError) ToResponse() *ErrorResponse {
	resp := e.BaseError.ToResponse()
	if e.Field != "" {
		resp.Message = fmt.Sprintf("%s: %s", e.Field, resp.Message)
	}
	return resp
}

// NotFoundError represents a 404 Not Found error
type NotFoundError struct {
	*BaseError
	Resource string
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *NotFoundError {
	return &NotFoundError{
		BaseError: &BaseError{
			Code: http.StatusNotFound,
			Type: "NotFoundError",
			Msg:  fmt.Sprintf("%s not found", resource),
		},
		Resource: resource,
	}
}

// NewNotFoundErrorWithDetails creates a not found error with custom message
func NewNotFoundErrorWithDetails(resource, message string) *NotFoundError {
	return &NotFoundError{
		BaseError: &BaseError{
			Code: http.StatusNotFound,
			Type: "NotFoundError",
			Msg:  message,
		},
		Resource: resource,
	}
}

// UnauthorizedError represents a 401 Unauthorized error
type UnauthorizedError struct {
	*BaseError
	Reason string
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(reason string) *UnauthorizedError {
	return &UnauthorizedError{
		BaseError: &BaseError{
			Code: http.StatusUnauthorized,
			Type: "UnauthorizedError",
			Msg:  "Authentication required",
		},
		Reason: reason,
	}
}

// NewUnauthorizedErrorWithMessage creates an unauthorized error with custom message
func NewUnauthorizedErrorWithMessage(message, reason string) *UnauthorizedError {
	return &UnauthorizedError{
		BaseError: &BaseError{
			Code: http.StatusUnauthorized,
			Type: "UnauthorizedError",
			Msg:  message,
		},
		Reason: reason,
	}
}

// ToResponse overrides to include reason in details
func (e *UnauthorizedError) ToResponse() *ErrorResponse {
	resp := e.BaseError.ToResponse()
	if e.Reason != "" && resp.Details == "" {
		resp.Details = e.Reason
	}
	return resp
}

// ForbiddenError represents a 403 Forbidden error
type ForbiddenError struct {
	*BaseError
	Action string
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(action string) *ForbiddenError {
	return &ForbiddenError{
		BaseError: &BaseError{
			Code: http.StatusForbidden,
			Type: "ForbiddenError",
			Msg:  fmt.Sprintf("Access denied: %s", action),
		},
		Action: action,
	}
}

// InternalError represents a 500 Internal Server Error
type InternalError struct {
	*BaseError
}

// NewInternalError creates a new internal server error
func NewInternalError(message string) *InternalError {
	return &InternalError{
		BaseError: &BaseError{
			Code: http.StatusInternalServerError,
			Type: "InternalError",
			Msg:  message,
		},
	}
}

// NewInternalErrorWithCause creates an internal error with a cause
func NewInternalErrorWithCause(message string, cause error) *InternalError {
	return &InternalError{
		BaseError: &BaseError{
			Code:  http.StatusInternalServerError,
			Type:  "InternalError",
			Msg:   message,
			Cause: cause,
		},
	}
}

// ConflictError represents a 409 Conflict error
type ConflictError struct {
	*BaseError
	Resource string
}

// NewConflictError creates a new conflict error
func NewConflictError(resource, message string) *ConflictError {
	return &ConflictError{
		BaseError: &BaseError{
			Code: http.StatusConflict,
			Type: "ConflictError",
			Msg:  message,
		},
		Resource: resource,
	}
}

// BadRequestError represents a 400 Bad Request error (generic)
type BadRequestError struct {
	*BaseError
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{
		BaseError: &BaseError{
			Code: http.StatusBadRequest,
			Type: "BadRequestError",
			Msg:  message,
		},
	}
}

// NewBadRequestErrorWithCause creates a bad request error with cause
func NewBadRequestErrorWithCause(message string, cause error) *BadRequestError {
	return &BadRequestError{
		BaseError: &BaseError{
			Code:  http.StatusBadRequest,
			Type:  "BadRequestError",
			Msg:   message,
			Cause: cause,
		},
	}
}

// Helper Functions

// WriteErrorResponse writes an HTTPError as JSON response
func WriteErrorResponse(w http.ResponseWriter, err HTTPError) {
	response := err.ToResponse()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Status)
	json.NewEncoder(w).Encode(response)
}

// WriteError writes a generic error as JSON response
// If err implements HTTPError, uses that. Otherwise creates InternalError.
func WriteError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(HTTPError); ok {
		WriteErrorResponse(w, httpErr)
		return
	}

	// Fallback to internal error
	internalErr := NewInternalErrorWithCause("Internal server error", err)
	WriteErrorResponse(w, internalErr)
}

// SendHTTPError is a helper to send HTTPError responses in handlers
func SendHTTPError(ctx *Context, err HTTPError) error {
	response := err.ToResponse()
	ctx.StatusCode = response.Status
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	ctx.ResponseWriter.WriteHeader(response.Status)
	return json.NewEncoder(ctx.ResponseWriter).Encode(response)
}

// WrapError converts a standard error to an HTTPError based on context
// This is useful for converting errors from other packages
func WrapError(err error, defaultMsg string) HTTPError {
	if err == nil {
		return nil
	}

	// If already an HTTPError, return as-is
	if httpErr, ok := err.(HTTPError); ok {
		return httpErr
	}

	// Otherwise wrap as internal error
	return NewInternalErrorWithCause(defaultMsg, err)
}

// IsHTTPError checks if an error implements HTTPError interface
func IsHTTPError(err error) bool {
	_, ok := err.(HTTPError)
	return ok
}

// GetStatusCode extracts the status code from an error
// Returns 500 if not an HTTPError
func GetStatusCode(err error) int {
	if httpErr, ok := err.(HTTPError); ok {
		return httpErr.StatusCode()
	}
	return http.StatusInternalServerError
}
