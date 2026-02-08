package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxConnectionsPerHub != 10000 {
		t.Errorf("MaxConnectionsPerHub = %d, want 10000", cfg.MaxConnectionsPerHub)
	}
	if cfg.MaxConnectionsPerRoom != 1000 {
		t.Errorf("MaxConnectionsPerRoom = %d, want 1000", cfg.MaxConnectionsPerRoom)
	}
	if !cfg.EnableHeartbeat {
		t.Error("EnableHeartbeat should be true by default")
	}
	if cfg.HeartbeatInterval != 30*time.Second {
		t.Errorf("HeartbeatInterval = %v, want 30s", cfg.HeartbeatInterval)
	}
	if cfg.HeartbeatTimeout != 90*time.Second {
		t.Errorf("HeartbeatTimeout = %v, want 90s", cfg.HeartbeatTimeout)
	}
	if cfg.PongWaitTimeout != 60*time.Second {
		t.Errorf("PongWaitTimeout = %v, want 60s", cfg.PongWaitTimeout)
	}
	if cfg.MaxMissedPongs != 3 {
		t.Errorf("MaxMissedPongs = %d, want 3", cfg.MaxMissedPongs)
	}
	if !cfg.EnableReconnection {
		t.Error("EnableReconnection should be true by default")
	}
	if cfg.ReconnectionTimeout != 30*time.Second {
		t.Errorf("ReconnectionTimeout = %v, want 30s", cfg.ReconnectionTimeout)
	}
	if cfg.MaxReconnectionTime != 5*time.Minute {
		t.Errorf("MaxReconnectionTime = %v, want 5m", cfg.MaxReconnectionTime)
	}
	if !cfg.PreserveClientState {
		t.Error("PreserveClientState should be true by default")
	}
	if cfg.MessageQueueSize != 256 {
		t.Errorf("MessageQueueSize = %d, want 256", cfg.MessageQueueSize)
	}
	if cfg.MessageQueueStrategy != QueueStrategyDropOldest {
		t.Errorf("MessageQueueStrategy = %q, want %q", cfg.MessageQueueStrategy, QueueStrategyDropOldest)
	}
	if cfg.MaxMessageSize != 512*1024 {
		t.Errorf("MaxMessageSize = %d, want %d", cfg.MaxMessageSize, 512*1024)
	}
	if cfg.WriteWait != 10*time.Second {
		t.Errorf("WriteWait = %v, want 10s", cfg.WriteWait)
	}
	if cfg.ReadWait != 60*time.Second {
		t.Errorf("ReadWait = %v, want 60s", cfg.ReadWait)
	}
	if !cfg.EnableMetrics {
		t.Error("EnableMetrics should be true by default")
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Run("zero values get defaults", func(t *testing.T) {
		cfg := &Config{}
		err := cfg.Validate()
		if err != nil {
			t.Fatalf("Validate() returned error: %v", err)
		}

		if cfg.HeartbeatInterval != 30*time.Second {
			t.Errorf("HeartbeatInterval = %v, want 30s", cfg.HeartbeatInterval)
		}
		if cfg.HeartbeatTimeout != 90*time.Second {
			t.Errorf("HeartbeatTimeout = %v, want 90s", cfg.HeartbeatTimeout)
		}
		if cfg.PongWaitTimeout != 60*time.Second {
			t.Errorf("PongWaitTimeout = %v, want 60s", cfg.PongWaitTimeout)
		}
		if cfg.MaxMissedPongs != 3 {
			t.Errorf("MaxMissedPongs = %d, want 3", cfg.MaxMissedPongs)
		}
		if cfg.MessageQueueSize != 256 {
			t.Errorf("MessageQueueSize = %d, want 256", cfg.MessageQueueSize)
		}
		if cfg.MaxMessageSize != 512*1024 {
			t.Errorf("MaxMessageSize = %d, want %d", cfg.MaxMessageSize, 512*1024)
		}
		if cfg.WriteWait != 10*time.Second {
			t.Errorf("WriteWait = %v, want 10s", cfg.WriteWait)
		}
		if cfg.ReadWait != 60*time.Second {
			t.Errorf("ReadWait = %v, want 60s", cfg.ReadWait)
		}
		if cfg.MessageQueueStrategy != QueueStrategyDropOldest {
			t.Errorf("MessageQueueStrategy = %q, want %q", cfg.MessageQueueStrategy, QueueStrategyDropOldest)
		}
	})

	t.Run("negative values get defaults", func(t *testing.T) {
		cfg := &Config{
			HeartbeatInterval: -1 * time.Second,
			MaxMissedPongs:    -5,
			MessageQueueSize:  -10,
			MaxMessageSize:    -100,
		}
		cfg.Validate()

		if cfg.HeartbeatInterval != 30*time.Second {
			t.Errorf("HeartbeatInterval = %v, want 30s", cfg.HeartbeatInterval)
		}
		if cfg.MaxMissedPongs != 3 {
			t.Errorf("MaxMissedPongs = %d, want 3", cfg.MaxMissedPongs)
		}
		if cfg.MessageQueueSize != 256 {
			t.Errorf("MessageQueueSize = %d, want 256", cfg.MessageQueueSize)
		}
		if cfg.MaxMessageSize != 512*1024 {
			t.Errorf("MaxMessageSize = %d, want %d", cfg.MaxMessageSize, 512*1024)
		}
	})

	t.Run("valid custom values preserved", func(t *testing.T) {
		cfg := &Config{
			HeartbeatInterval: 45 * time.Second,
			HeartbeatTimeout:  120 * time.Second,
			PongWaitTimeout:   90 * time.Second,
			MaxMissedPongs:    5,
			MessageQueueSize:  512,
			MaxMessageSize:    1024 * 1024,
			WriteWait:         20 * time.Second,
			ReadWait:          120 * time.Second,
		}
		cfg.Validate()

		if cfg.HeartbeatInterval != 45*time.Second {
			t.Errorf("HeartbeatInterval should be preserved, got %v", cfg.HeartbeatInterval)
		}
		if cfg.MaxMissedPongs != 5 {
			t.Errorf("MaxMissedPongs should be preserved, got %d", cfg.MaxMissedPongs)
		}
	})
}

func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		host           string
		expected       bool
	}{
		{
			name:           "no origin header allows request",
			allowedOrigins: nil,
			origin:         "",
			host:           "example.com",
			expected:       true,
		},
		{
			name:           "wildcard allows any origin",
			allowedOrigins: []string{"*"},
			origin:         "http://evil.com",
			host:           "example.com",
			expected:       true,
		},
		{
			name:           "exact origin match",
			allowedOrigins: []string{"http://trusted.com"},
			origin:         "http://trusted.com",
			host:           "example.com",
			expected:       true,
		},
		{
			name:           "case insensitive origin match",
			allowedOrigins: []string{"http://Trusted.Com"},
			origin:         "http://trusted.com",
			host:           "example.com",
			expected:       true,
		},
		{
			name:           "origin not in allowed list",
			allowedOrigins: []string{"http://trusted.com"},
			origin:         "http://evil.com",
			host:           "example.com",
			expected:       false,
		},
		{
			name:           "same origin allowed when no origins configured",
			allowedOrigins: nil,
			origin:         "http://example.com",
			host:           "example.com",
			expected:       true,
		},
		{
			name:           "same origin with port",
			allowedOrigins: nil,
			origin:         "http://example.com:8080",
			host:           "example.com:8080",
			expected:       true,
		},
		{
			name:           "cross-origin rejected when no origins configured",
			allowedOrigins: nil,
			origin:         "http://evil.com",
			host:           "example.com",
			expected:       false,
		},
		{
			name:           "multiple allowed origins",
			allowedOrigins: []string{"http://a.com", "http://b.com", "http://c.com"},
			origin:         "http://b.com",
			host:           "example.com",
			expected:       true,
		},
		{
			name:           "multiple allowed origins - not matching",
			allowedOrigins: []string{"http://a.com", "http://b.com"},
			origin:         "http://c.com",
			host:           "example.com",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				AllowedOrigins: tt.allowedOrigins,
			}

			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			req.Host = tt.host

			result := cfg.CheckOrigin(req)
			if result != tt.expected {
				t.Errorf("CheckOrigin() = %v, want %v", result, tt.expected)
			}
		})
	}
}
