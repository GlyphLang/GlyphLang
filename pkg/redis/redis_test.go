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

func TestMockHandlerSetNX(t *testing.T) {
	mock := NewMockHandler()

	// First SetNX should succeed
	ok, err := mock.SetNX("key1", "value1")
	require.NoError(t, err)
	assert.True(t, ok)

	// Second SetNX on same key should fail
	ok, err = mock.SetNX("key1", "value2")
	require.NoError(t, err)
	assert.False(t, ok)

	// Original value should remain
	val, err := mock.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestMockHandlerIncrBy(t *testing.T) {
	mock := NewMockHandler()

	// IncrBy on non-existent key
	val, err := mock.IncrBy("counter", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)

	// IncrBy on existing key
	val, err = mock.IncrBy("counter", 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), val)

	// IncrBy with negative value
	val, err = mock.IncrBy("counter", -3)
	require.NoError(t, err)
	assert.Equal(t, int64(12), val)
}

func TestMockHandlerDecrBy(t *testing.T) {
	mock := NewMockHandler()

	// DecrBy on non-existent key
	val, err := mock.DecrBy("counter", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(-5), val)

	// DecrBy on existing key
	val, err = mock.DecrBy("counter", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(-8), val)
}

func TestMockHandlerTTL(t *testing.T) {
	mock := NewMockHandler()

	// TTL on non-existent key returns -2
	ttl, err := mock.TTL("nonexistent")
	require.NoError(t, err)
	assert.Equal(t, int64(-2), ttl)

	// TTL on existing key without expiration returns -1
	mock.Set("key1", "value1")
	ttl, err = mock.TTL("key1")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), ttl)
}

func TestMockHandlerHashOperations(t *testing.T) {
	mock := NewMockHandler()

	// HSet - create new field
	count, err := mock.HSet("user:1", "name", "Alice")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// HSet - update existing field
	count, err = mock.HSet("user:1", "name", "Bob")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// HGet - existing field
	val, err := mock.HGet("user:1", "name")
	require.NoError(t, err)
	assert.Equal(t, "Bob", val)

	// HGet - non-existent field
	val, err = mock.HGet("user:1", "email")
	require.NoError(t, err)
	assert.Nil(t, val)

	// HGet - non-existent hash
	val, err = mock.HGet("user:999", "name")
	require.NoError(t, err)
	assert.Nil(t, val)

	// HExists
	exists, err := mock.HExists("user:1", "name")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = mock.HExists("user:1", "email")
	require.NoError(t, err)
	assert.False(t, exists)

	// HSet multiple fields
	mock.HSet("user:1", "email", "bob@example.com")
	mock.HSet("user:1", "age", 30)

	// HGetAll
	all, err := mock.HGetAll("user:1")
	require.NoError(t, err)
	assert.Equal(t, 3, len(all))
	assert.Equal(t, "Bob", all["name"])
	assert.Equal(t, "bob@example.com", all["email"])
	assert.Equal(t, "30", all["age"])

	// HDel - delete one field
	deleted, err := mock.HDel("user:1", "age")
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// HDel - delete non-existent field
	deleted, err = mock.HDel("user:1", "phone")
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)

	// Verify deletion
	exists, err = mock.HExists("user:1", "age")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMockHandlerListOperations(t *testing.T) {
	mock := NewMockHandler()

	// LPush - prepend to empty list
	length, err := mock.LPush("mylist", "a")
	require.NoError(t, err)
	assert.Equal(t, int64(1), length)

	// LPush - prepend multiple
	length, err = mock.LPush("mylist", "b", "c")
	require.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// List should be: c, b, a
	items, err := mock.LRange("mylist", 0, -1)
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "b", "a"}, items)

	// RPush - append
	length, err = mock.RPush("mylist", "d")
	require.NoError(t, err)
	assert.Equal(t, int64(4), length)

	// List should be: c, b, a, d
	items, err = mock.LRange("mylist", 0, -1)
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "b", "a", "d"}, items)

	// LLen
	length, err = mock.LLen("mylist")
	require.NoError(t, err)
	assert.Equal(t, int64(4), length)

	// LPop
	val, err := mock.LPop("mylist")
	require.NoError(t, err)
	assert.Equal(t, "c", val)

	// RPop
	val, err = mock.RPop("mylist")
	require.NoError(t, err)
	assert.Equal(t, "d", val)

	// LRange with indices
	items, err = mock.LRange("mylist", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"b"}, items)

	// LPop/RPop on empty list
	mock2 := NewMockHandler()
	val, err = mock2.LPop("empty")
	require.NoError(t, err)
	assert.Nil(t, val)

	val, err = mock2.RPop("empty")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestMockHandlerSetOperations(t *testing.T) {
	mock := NewMockHandler()

	// SAdd - add new members
	added, err := mock.SAdd("myset", "a", "b", "c")
	require.NoError(t, err)
	assert.Equal(t, int64(3), added)

	// SAdd - add duplicate (should not increase count)
	added, err = mock.SAdd("myset", "a", "d")
	require.NoError(t, err)
	assert.Equal(t, int64(1), added) // Only "d" is new

	// SIsMember
	isMember, err := mock.SIsMember("myset", "a")
	require.NoError(t, err)
	assert.True(t, isMember)

	isMember, err = mock.SIsMember("myset", "z")
	require.NoError(t, err)
	assert.False(t, isMember)

	// SMembers
	members, err := mock.SMembers("myset")
	require.NoError(t, err)
	assert.Equal(t, 4, len(members))

	// SRem
	removed, err := mock.SRem("myset", "a", "b")
	require.NoError(t, err)
	assert.Equal(t, int64(2), removed)

	// Verify removal
	isMember, err = mock.SIsMember("myset", "a")
	require.NoError(t, err)
	assert.False(t, isMember)

	// SRem non-existent member
	removed, err = mock.SRem("myset", "z")
	require.NoError(t, err)
	assert.Equal(t, int64(0), removed)
}

func TestMockHandlerPublish(t *testing.T) {
	mock := NewMockHandler()

	// Publish returns 0 (no subscribers in mock)
	count, err := mock.Publish("channel", "message")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestMockHandlerErrorCases(t *testing.T) {
	mock := NewMockHandler()

	// Incr on non-integer value
	mock.Set("string_key", "not_a_number")
	_, err := mock.Incr("string_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an integer")

	// Decr on non-integer value
	_, err = mock.Decr("string_key")
	assert.Error(t, err)

	// IncrBy on non-integer value
	_, err = mock.IncrBy("string_key", 5)
	assert.Error(t, err)

	// DecrBy on non-integer value
	_, err = mock.DecrBy("string_key", 5)
	assert.Error(t, err)
}

func TestParseConnectionStringEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		connStr   string
		expectErr bool
		check     func(t *testing.T, config *Config)
	}{
		{
			name:    "default port when not specified",
			connStr: "redis://localhost",
			check: func(t *testing.T, config *Config) {
				assert.Equal(t, 6379, config.Port)
			},
		},
		{
			name:    "password only (no username)",
			connStr: "redis://:mypassword@localhost:6379",
			check: func(t *testing.T, config *Config) {
				assert.Equal(t, "mypassword", config.Password)
			},
		},
		{
			name:    "username treated as password",
			connStr: "redis://secrettoken@localhost:6379",
			check: func(t *testing.T, config *Config) {
				assert.Equal(t, "secrettoken", config.Password)
			},
		},
		{
			name:    "database 15 (max default)",
			connStr: "redis://localhost:6379/15",
			check: func(t *testing.T, config *Config) {
				assert.Equal(t, 15, config.DB)
			},
		},
		{
			name:    "custom timeout parameters",
			connStr: "redis://localhost:6379?dial_timeout=10s&read_timeout=5s&write_timeout=5s",
			check: func(t *testing.T, config *Config) {
				assert.Equal(t, 10*time.Second, config.DialTimeout)
				assert.Equal(t, 5*time.Second, config.ReadTimeout)
				assert.Equal(t, 5*time.Second, config.WriteTimeout)
			},
		},
		{
			name:      "empty scheme",
			connStr:   "://localhost:6379",
			expectErr: true,
		},
		{
			name:      "http scheme (invalid)",
			connStr:   "http://localhost:6379",
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
			if tt.check != nil {
				tt.check(t, config)
			}
		})
	}
}
