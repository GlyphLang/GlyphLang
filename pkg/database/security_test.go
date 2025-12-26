package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantErr   bool
		errString string
	}{
		{
			name:  "valid simple identifier",
			input: "users",
			want:  `"users"`,
		},
		{
			name:  "valid identifier with underscore",
			input: "user_accounts",
			want:  `"user_accounts"`,
		},
		{
			name:  "valid identifier starting with underscore",
			input: "_internal",
			want:  `"_internal"`,
		},
		{
			name:  "valid identifier with numbers",
			input: "table123",
			want:  `"table123"`,
		},
		{
			name:  "valid mixed case identifier",
			input: "UserAccounts",
			want:  `"UserAccounts"`,
		},
		{
			name:      "empty identifier",
			input:     "",
			wantErr:   true,
			errString: "identifier cannot be empty",
		},
		{
			name:    "SQL injection attempt with semicolon",
			input:   "users; DROP TABLE users;--",
			wantErr: true,
		},
		{
			name:    "SQL injection attempt with quotes",
			input:   "users'--",
			wantErr: true,
		},
		{
			name:    "SQL injection attempt with double quotes",
			input:   `users"--`,
			wantErr: true,
		},
		{
			name:    "identifier with spaces",
			input:   "user accounts",
			wantErr: true,
		},
		{
			name:    "identifier with hyphen",
			input:   "user-accounts",
			wantErr: true,
		},
		{
			name:    "identifier starting with number",
			input:   "123table",
			wantErr: true,
		},
		{
			name:    "identifier with parentheses",
			input:   "users()",
			wantErr: true,
		},
		{
			name:    "identifier with equal sign",
			input:   "users=1",
			wantErr: true,
		},
		{
			name:    "identifier with OR injection",
			input:   "users OR 1=1",
			wantErr: true,
		},
		{
			name:    "identifier with UNION injection",
			input:   "users UNION SELECT",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeIdentifier(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSanitizeIdentifiers(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr bool
	}{
		{
			name:  "valid identifiers",
			input: []string{"id", "name", "email"},
			want:  []string{`"id"`, `"name"`, `"email"`},
		},
		{
			name:  "single valid identifier",
			input: []string{"users"},
			want:  []string{`"users"`},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
		{
			name:    "one invalid identifier",
			input:   []string{"id", "name; DROP TABLE--", "email"},
			wantErr: true,
		},
		{
			name:    "first identifier invalid",
			input:   []string{"invalid'", "name", "email"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeIdentifiers(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid identifier",
			input:   "valid_table",
			wantErr: false,
		},
		{
			name:    "empty identifier",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			input:   "table;DROP",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeColumnType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid INTEGER type",
			input: "INTEGER",
			want:  "INTEGER",
		},
		{
			name:  "valid VARCHAR with length",
			input: "VARCHAR(255)",
			want:  "VARCHAR(255)",
		},
		{
			name:  "valid NUMERIC with precision",
			input: "NUMERIC(10, 2)",
			want:  "NUMERIC(10, 2)",
		},
		{
			name:  "valid TEXT type",
			input: "TEXT",
			want:  "TEXT",
		},
		{
			name:  "valid BOOLEAN type",
			input: "BOOLEAN",
			want:  "BOOLEAN",
		},
		{
			name:  "valid TIMESTAMP type",
			input: "TIMESTAMP",
			want:  "TIMESTAMP",
		},
		{
			name:  "valid JSON type",
			input: "JSON",
			want:  "JSON",
		},
		{
			name:  "valid JSONB type",
			input: "JSONB",
			want:  "JSONB",
		},
		{
			name:  "valid UUID type",
			input: "UUID",
			want:  "UUID",
		},
		{
			name:    "empty type",
			input:   "",
			wantErr: true,
		},
		{
			name:    "SQL injection in type",
			input:   "INTEGER; DROP TABLE users;--",
			wantErr: true,
		},
		{
			name:    "unknown type",
			input:   "UNKNOWNTYPE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizeColumnType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestErrInvalidIdentifier(t *testing.T) {
	// Verify the error message is informative
	assert.Contains(t, ErrInvalidIdentifier.Error(), "invalid identifier")
	assert.Contains(t, ErrInvalidIdentifier.Error(), "alphanumeric")
}
