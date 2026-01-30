package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name           string
		err            *ValidationError
		expectedStatus int
		expectedType   string
		expectedMsg    string
	}{
		{
			name:           "basic validation error",
			err:            NewValidationError("email", "invalid email format"),
			expectedStatus: http.StatusBadRequest,
			expectedType:   "ValidationError",
			expectedMsg:    "email: invalid email format",
		},
		{
			name:           "validation error with details",
			err:            NewValidationErrorWithDetails("password", "password too weak", "must contain 8+ characters"),
			expectedStatus: http.StatusBadRequest,
			expectedType:   "ValidationError",
			expectedMsg:    "password: password too weak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("StatusCode() = %d, want %d", tt.err.StatusCode(), tt.expectedStatus)
			}

			resp := tt.err.ToResponse()
			if resp.Error != tt.expectedType {
				t.Errorf("Error type = %s, want %s", resp.Error, tt.expectedType)
			}
			if resp.Message != tt.expectedMsg {
				t.Errorf("Message = %s, want %s", resp.Message, tt.expectedMsg)
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name           string
		err            *NotFoundError
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "resource not found",
			err:            NewNotFoundError("User"),
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "User not found",
		},
		{
			name:           "custom not found message",
			err:            NewNotFoundErrorWithDetails("Post", "The requested blog post does not exist"),
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "The requested blog post does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("StatusCode() = %d, want %d", tt.err.StatusCode(), tt.expectedStatus)
			}

			resp := tt.err.ToResponse()
			if resp.Message != tt.expectedMsg {
				t.Errorf("Message = %s, want %s", resp.Message, tt.expectedMsg)
			}
		})
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("invalid token")

	if err.StatusCode() != http.StatusUnauthorized {
		t.Errorf("StatusCode() = %d, want %d", err.StatusCode(), http.StatusUnauthorized)
	}

	resp := err.ToResponse()
	if resp.Error != "UnauthorizedError" {
		t.Errorf("Error type = %s, want UnauthorizedError", resp.Error)
	}
	if resp.Details != "invalid token" {
		t.Errorf("Details = %s, want 'invalid token'", resp.Details)
	}
}

func TestForbiddenError(t *testing.T) {
	err := NewForbiddenError("delete user")

	if err.StatusCode() != http.StatusForbidden {
		t.Errorf("StatusCode() = %d, want %d", err.StatusCode(), http.StatusForbidden)
	}

	resp := err.ToResponse()
	if resp.Error != "ForbiddenError" {
		t.Errorf("Error type = %s, want ForbiddenError", resp.Error)
	}
}

func TestInternalError(t *testing.T) {
	tests := []struct {
		name        string
		err         *InternalError
		expectCause bool
	}{
		{
			name:        "basic internal error",
			err:         NewInternalError("database connection failed"),
			expectCause: false,
		},
		{
			name:        "internal error with cause",
			err:         NewInternalErrorWithCause("failed to process request", errors.New("nil pointer")),
			expectCause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode() != http.StatusInternalServerError {
				t.Errorf("StatusCode() = %d, want %d", tt.err.StatusCode(), http.StatusInternalServerError)
			}

			resp := tt.err.ToResponse()
			if tt.expectCause && resp.Details == "" {
				t.Error("Expected details to contain cause error")
			}
		})
	}
}

func TestConflictError(t *testing.T) {
	err := NewConflictError("User", "email already exists")

	if err.StatusCode() != http.StatusConflict {
		t.Errorf("StatusCode() = %d, want %d", err.StatusCode(), http.StatusConflict)
	}

	resp := err.ToResponse()
	if resp.Error != "ConflictError" {
		t.Errorf("Error type = %s, want ConflictError", resp.Error)
	}
	if resp.Message != "email already exists" {
		t.Errorf("Message = %s, want 'email already exists'", resp.Message)
	}
}

func TestBadRequestError(t *testing.T) {
	tests := []struct {
		name        string
		err         *BadRequestError
		expectCause bool
	}{
		{
			name:        "basic bad request",
			err:         NewBadRequestError("missing required field"),
			expectCause: false,
		},
		{
			name:        "bad request with cause",
			err:         NewBadRequestErrorWithCause("invalid JSON", errors.New("unexpected token")),
			expectCause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode() != http.StatusBadRequest {
				t.Errorf("StatusCode() = %d, want %d", tt.err.StatusCode(), http.StatusBadRequest)
			}

			resp := tt.err.ToResponse()
			if tt.expectCause && resp.Details == "" {
				t.Error("Expected details to contain cause error")
			}
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		err            HTTPError
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "validation error response",
			err:            NewValidationError("name", "required field"),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ValidationError",
		},
		{
			name:           "not found error response",
			err:            NewNotFoundError("Resource"),
			expectedStatus: http.StatusNotFound,
			expectedError:  "NotFoundError",
		},
		{
			name:           "internal error response",
			err:            NewInternalError("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "InternalError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteErrorResponse(w, tt.err)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.expectedStatus)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp.Error != tt.expectedError {
				t.Errorf("Error type = %s, want %s", resp.Error, tt.expectedError)
			}
			if resp.Status != tt.expectedStatus {
				t.Errorf("Response status = %d, want %d", resp.Status, tt.expectedStatus)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "HTTP error",
			err:            NewValidationError("field", "invalid"),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ValidationError",
		},
		{
			name:           "standard error converts to internal",
			err:            errors.New("generic error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "InternalError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.err)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.expectedStatus)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp.Error != tt.expectedError {
				t.Errorf("Error type = %s, want %s", resp.Error, tt.expectedError)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		defaultMsg   string
		expectedType string
		expectedCode int
	}{
		{
			name:         "nil error",
			err:          nil,
			defaultMsg:   "default",
			expectedType: "",
			expectedCode: 0,
		},
		{
			name:         "already HTTP error",
			err:          NewValidationError("field", "invalid"),
			defaultMsg:   "default",
			expectedType: "ValidationError",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "standard error wraps to internal",
			err:          errors.New("some error"),
			defaultMsg:   "failed to process",
			expectedType: "InternalError",
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.err, tt.defaultMsg)

			if tt.err == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if result.ErrorType() != tt.expectedType {
				t.Errorf("Error type = %s, want %s", result.ErrorType(), tt.expectedType)
			}
			if result.StatusCode() != tt.expectedCode {
				t.Errorf("Status code = %d, want %d", result.StatusCode(), tt.expectedCode)
			}
		})
	}
}

func TestIsHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "validation error is HTTP error",
			err:      NewValidationError("field", "invalid"),
			expected: true,
		},
		{
			name:     "internal error is HTTP error",
			err:      NewInternalError("error"),
			expected: true,
		},
		{
			name:     "standard error is not HTTP error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHTTPError(tt.err)
			if result != tt.expected {
				t.Errorf("IsHTTPError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "validation error",
			err:          NewValidationError("field", "invalid"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "not found error",
			err:          NewNotFoundError("Resource"),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "standard error defaults to 500",
			err:          errors.New("generic error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetStatusCode(tt.err)
			if code != tt.expectedCode {
				t.Errorf("GetStatusCode() = %d, want %d", code, tt.expectedCode)
			}
		})
	}
}

func TestErrorResponseJSON(t *testing.T) {
	err := NewValidationErrorWithDetails("email", "invalid format", "must be a valid email address")
	w := httptest.NewRecorder()
	WriteErrorResponse(w, err)

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// Verify all fields are properly serialized
	if resp.Status != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.Status, http.StatusBadRequest)
	}
	if resp.Error != "ValidationError" {
		t.Errorf("Error = %s, want ValidationError", resp.Error)
	}
	if resp.Message != "email: invalid format" {
		t.Errorf("Message = %s, want 'email: invalid format'", resp.Message)
	}
	if resp.Details != "must be a valid email address" {
		t.Errorf("Details = %s, want 'must be a valid email address'", resp.Details)
	}
}
