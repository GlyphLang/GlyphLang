package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name:    "PostgreSQL connection string",
			connStr: "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres", cfg.Driver)
				assert.Equal(t, "localhost", cfg.Host)
				assert.Equal(t, 5432, cfg.Port)
				assert.Equal(t, "testdb", cfg.Database)
				assert.Equal(t, "user", cfg.Username)
				assert.Equal(t, "pass", cfg.Password)
				assert.Equal(t, "disable", cfg.SSLMode)
			},
		},
		{
			name:    "PostgreSQL without port",
			connStr: "postgres://user:pass@localhost/testdb",
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "postgres", cfg.Driver)
				assert.Equal(t, 5432, cfg.Port) // Default port
			},
		},
		{
			name:    "PostgreSQL without SSL mode",
			connStr: "postgres://user:pass@localhost:5432/testdb",
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "prefer", cfg.SSLMode) // Default SSL mode
			},
		},
		{
			name:    "Invalid connection string",
			connStr: "not-a-valid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseConnectionString(tt.connStr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestConnectionString(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "PostgreSQL connection string",
			config: &Config{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable",
		},
		{
			name: "MySQL connection string",
			config: &Config{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "user",
				Password: "pass",
			},
			want: "user:pass@tcp(localhost:3306)/testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ConnectionString()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "PostgreSQL database",
			config: &Config{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "Unsupported driver",
			config: &Config{
				Driver: "unsupported",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, db)
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	connStr := "postgres://user:pass@localhost/testdb"
	cfg, err := ParseConnectionString(connStr)
	require.NoError(t, err)

	assert.Equal(t, 25, cfg.MaxOpenConns)
	assert.Equal(t, 5, cfg.MaxIdleConns)
	assert.NotZero(t, cfg.ConnMaxLifetime)
	assert.NotZero(t, cfg.ConnMaxIdleTime)
}
