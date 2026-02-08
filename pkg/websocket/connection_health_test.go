package websocket

import (
	"sync"
	"testing"
	"time"
)

func TestConnection_IsHealthy_HeartbeatDisabled(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = false
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "healthy-no-hb",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  100,         // even many missed pongs
		lastPongTime: time.Time{}, // zero time
	}

	if !conn.IsHealthy() {
		t.Error("connection should always be healthy when heartbeat is disabled")
	}
}

func TestConnection_IsHealthy_HealthyConnection(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 3
	config.HeartbeatTimeout = 90 * time.Second
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "healthy-conn",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  0,
		lastPongTime: time.Now(),
	}

	if !conn.IsHealthy() {
		t.Error("connection with 0 missed pongs and recent pong should be healthy")
	}
}

func TestConnection_IsHealthy_AtMaxMissedPongs(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 3
	config.HeartbeatTimeout = 90 * time.Second
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	// Exactly at max -- should still be healthy (check is >)
	conn := &Connection{
		ID:           "at-max-pongs",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  3,
		lastPongTime: time.Now(),
	}

	if !conn.IsHealthy() {
		t.Error("connection at exactly MaxMissedPongs should still be healthy (check is strictly >)")
	}
}

func TestConnection_IsHealthy_ExceededMaxMissedPongs(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 3
	config.HeartbeatTimeout = 90 * time.Second
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "exceeded-pongs",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  4,
		lastPongTime: time.Now(),
	}

	if conn.IsHealthy() {
		t.Error("connection exceeding MaxMissedPongs should be unhealthy")
	}
}

func TestConnection_IsHealthy_PongTimeout(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 3
	config.HeartbeatTimeout = 100 * time.Millisecond
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "timeout-conn",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  0,
		lastPongTime: time.Now().Add(-10 * time.Second),
	}

	if conn.IsHealthy() {
		t.Error("connection with expired pong timeout should be unhealthy")
	}
}

func TestConnection_IsHealthy_BothConditionsUnhealthy(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 2
	config.HeartbeatTimeout = 100 * time.Millisecond
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "both-unhealthy",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  10,
		lastPongTime: time.Now().Add(-1 * time.Hour),
	}

	if conn.IsHealthy() {
		t.Error("connection failing both checks should be unhealthy")
	}
}

func TestConnection_GetMissedPongs(t *testing.T) {
	config := DefaultConfig()
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:          "pong-tracking",
		hub:         hub,
		send:        make(chan []byte, 256),
		Data:        make(map[string]interface{}),
		rooms:       make(map[string]bool),
		missedPongs: 0,
	}

	if got := conn.GetMissedPongs(); got != 0 {
		t.Errorf("GetMissedPongs() = %d, want 0", got)
	}

	// Simulate missed pongs by directly setting (since we are in the same package)
	conn.heartbeatMu.Lock()
	conn.missedPongs = 5
	conn.heartbeatMu.Unlock()

	if got := conn.GetMissedPongs(); got != 5 {
		t.Errorf("GetMissedPongs() = %d, want 5", got)
	}

	// Simulate pong received (reset)
	conn.heartbeatMu.Lock()
	conn.missedPongs = 0
	conn.heartbeatMu.Unlock()

	if got := conn.GetMissedPongs(); got != 0 {
		t.Errorf("GetMissedPongs() after reset = %d, want 0", got)
	}
}

func TestConnection_GetLastPongTime(t *testing.T) {
	config := DefaultConfig()
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	initialTime := time.Now()
	conn := &Connection{
		ID:           "pong-time",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		lastPongTime: initialTime,
	}

	got := conn.GetLastPongTime()
	if !got.Equal(initialTime) {
		t.Errorf("GetLastPongTime() = %v, want %v", got, initialTime)
	}

	// Simulate pong received
	newTime := time.Now().Add(5 * time.Second)
	conn.heartbeatMu.Lock()
	conn.lastPongTime = newTime
	conn.heartbeatMu.Unlock()

	got = conn.GetLastPongTime()
	if !got.Equal(newTime) {
		t.Errorf("GetLastPongTime() after update = %v, want %v", got, newTime)
	}
}

func TestConnection_GetLastPongTime_ZeroValue(t *testing.T) {
	config := DefaultConfig()
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "pong-zero",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		lastPongTime: time.Time{},
	}

	got := conn.GetLastPongTime()
	if !got.IsZero() {
		t.Errorf("GetLastPongTime() = %v, want zero time", got)
	}
}

func TestConnection_HealthMethods_ConcurrentAccess(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 5
	config.HeartbeatTimeout = 90 * time.Second
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "concurrent-health",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  0,
		lastPongTime: time.Now(),
	}

	var wg sync.WaitGroup

	// Writers: simulate heartbeat tracking
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				conn.heartbeatMu.Lock()
				conn.missedPongs++
				conn.heartbeatMu.Unlock()

				conn.heartbeatMu.Lock()
				conn.missedPongs = 0
				conn.lastPongTime = time.Now()
				conn.heartbeatMu.Unlock()
			}
		}()
	}

	// Readers: concurrently check health state
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = conn.GetMissedPongs()
				_ = conn.GetLastPongTime()
				_ = conn.IsHealthy()
			}
		}()
	}

	wg.Wait()
	// Should not have any data races (verified with -race flag)
}

func TestConnection_HealthTransition(t *testing.T) {
	config := DefaultConfig()
	config.EnableHeartbeat = true
	config.MaxMissedPongs = 2
	config.HeartbeatTimeout = 90 * time.Second
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:           "transition",
		hub:          hub,
		send:         make(chan []byte, 256),
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		missedPongs:  0,
		lastPongTime: time.Now(),
	}

	// Start healthy
	if !conn.IsHealthy() {
		t.Error("connection should start healthy")
	}

	// Miss some pongs but stay within limit
	conn.heartbeatMu.Lock()
	conn.missedPongs = 1
	conn.heartbeatMu.Unlock()

	if !conn.IsHealthy() {
		t.Error("connection should still be healthy with 1 missed pong (max 2)")
	}

	// Exceed limit
	conn.heartbeatMu.Lock()
	conn.missedPongs = 3
	conn.heartbeatMu.Unlock()

	if conn.IsHealthy() {
		t.Error("connection should be unhealthy with 3 missed pongs (max 2)")
	}

	// Simulate pong received (recovery)
	conn.heartbeatMu.Lock()
	conn.missedPongs = 0
	conn.lastPongTime = time.Now()
	conn.heartbeatMu.Unlock()

	if !conn.IsHealthy() {
		t.Error("connection should be healthy again after pong received")
	}
}

func TestConnection_NewConnection_LastPongTime(t *testing.T) {
	config := DefaultConfig()
	hub := NewHubWithConfig(config)
	go hub.Run()
	defer hub.Shutdown()

	before := time.Now()
	conn := NewConnection("new-conn", nil, hub)
	after := time.Now()

	pongTime := conn.GetLastPongTime()
	if pongTime.Before(before) || pongTime.After(after) {
		t.Error("NewConnection should set lastPongTime to approximately time.Now()")
	}

	if conn.GetMissedPongs() != 0 {
		t.Errorf("new connection should have 0 missed pongs, got %d", conn.GetMissedPongs())
	}
}
