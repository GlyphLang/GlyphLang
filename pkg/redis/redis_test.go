package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 2, config.MinIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.False(t, config.UseTLS)
	assert.False(t, config.ClusterMode)
	assert.False(t, config.SentinelMode)
}

func TestParseConnectionString(t *testing.T) {
	tests := []struct {
		name      string
		connStr   string
		expected  *Config
		expectErr bool
	}{
		{
			name:    "simple connection",
			connStr: "redis://localhost:6379",
			expected: &Config{
				Host:            "localhost",
				Port:            6379,
				DB:              0,
				UseTLS:          false,
				PoolSize:        10,
				MinIdleConns:    2,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
			},
		},
		{
			name:    "with password",
			connStr: "redis://:secretpass@localhost:6379",
			expected: &Config{
				Host:            "localhost",
				Port:            6379,
				Password:        "secretpass",
				DB:              0,
				UseTLS:          false,
				PoolSize:        10,
				MinIdleConns:    2,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
			},
		},
		{
			name:    "with database number",
			connStr: "redis://localhost:6379/5",
			expected: &Config{
				Host:            "localhost",
				Port:            6379,
				DB:              5,
				UseTLS:          false,
				PoolSize:        10,
				MinIdleConns:    2,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
			},
		},
		{
			name:    "with TLS (rediss)",
			connStr: "rediss://localhost:6379",
			expected: &Config{
				Host:            "localhost",
				Port:            6379,
				DB:              0,
				UseTLS:          true,
				PoolSize:        10,
				MinIdleConns:    2,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
			},
		},
		{
			name:    "with pool size query param",
			connStr: "redis://localhost:6379?pool_size=20",
			expected: &Config{
				Host:            "localhost",
				Port:            6379,
				DB:              0,
				UseTLS:          false,
				PoolSize:        20,
				MinIdleConns:    2,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
			},
		},
		{
			name:      "invalid scheme",
			connStr:   "postgres://localhost:6379",
			expectErr: true,
		},
		{
			name:      "invalid database number",
			connStr:   "redis://localhost:6379/abc",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseConnectionString(tt.connStr)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Host, config.Host)
			assert.Equal(t, tt.expected.Port, config.Port)
			assert.Equal(t, tt.expected.Password, config.Password)
			assert.Equal(t, tt.expected.DB, config.DB)
			assert.Equal(t, tt.expected.UseTLS, config.UseTLS)
			assert.Equal(t, tt.expected.PoolSize, config.PoolSize)
		})
	}
}

func TestConfigConnectionString(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "secret",
		DB:       2,
		UseTLS:   false,
	}

	connStr := config.ConnectionString()
	assert.Equal(t, "redis://:secret@localhost:6379/2", connStr)

	// With TLS
	config.UseTLS = true
	connStr = config.ConnectionString()
	assert.Equal(t, "rediss://:secret@localhost:6379/2", connStr)
}

func TestConfigAddress(t *testing.T) {
	config := &Config{
		Host: "redis.example.com",
		Port: 6380,
	}

	assert.Equal(t, "redis.example.com:6380", config.Address())
}

func TestMockHandler(t *testing.T) {
	mock := NewMockHandler()

	// Test Set and Get
	err := mock.Set("key1", "value1")
	require.NoError(t, err)

	val, err := mock.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test Get non-existent key
	val, err = mock.Get("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, val)

	// Test Exists
	exists, err := mock.Exists("key1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = mock.Exists("nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test Del
	count, err := mock.Del("key1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	exists, err = mock.Exists("key1")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test Incr
	val1, err := mock.Incr("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val1)

	val2, err := mock.Incr("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(2), val2)

	// Test Decr
	val3, err := mock.Decr("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val3)

	// Test Expire (no-op for mock)
	ok, err := mock.Expire("counter", 300)
	require.NoError(t, err)
	assert.True(t, ok)

	// Test Close
	err = mock.Close()
	require.NoError(t, err)
}

func TestMockHandlerConcurrency(t *testing.T) {
	mock := NewMockHandler()
	done := make(chan bool)

	// Run concurrent operations
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "key"
			mock.Set(key, id)
			mock.Get(key)
			mock.Incr("counter")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Counter should be 10
	val, err := mock.Get("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(10), val)
}
