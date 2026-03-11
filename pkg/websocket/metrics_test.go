package websocket

import (
	"sync"
	"testing"
	"time"
)

func TestNewMetrics_Initialization(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	// Verify all counters start at zero
	if m.GetActiveConnections() != 0 {
		t.Errorf("activeConnections = %d, want 0", m.GetActiveConnections())
	}
	if m.GetTotalConnections() != 0 {
		t.Errorf("totalConnections = %d, want 0", m.GetTotalConnections())
	}
	if m.GetTotalDisconnections() != 0 {
		t.Errorf("totalDisconnections = %d, want 0", m.GetTotalDisconnections())
	}
	if m.GetRejectedConnections() != 0 {
		t.Errorf("rejectedConnections = %d, want 0", m.GetRejectedConnections())
	}
	if m.GetMessagesSent() != 0 {
		t.Errorf("messagesSent = %d, want 0", m.GetMessagesSent())
	}
	if m.GetMessagesReceived() != 0 {
		t.Errorf("messagesReceived = %d, want 0", m.GetMessagesReceived())
	}
	if m.GetMessagesFailed() != 0 {
		t.Errorf("messagesFailed = %d, want 0", m.GetMessagesFailed())
	}
	if m.GetBytesSent() != 0 {
		t.Errorf("bytesSent = %d, want 0", m.GetBytesSent())
	}
	if m.GetBytesReceived() != 0 {
		t.Errorf("bytesReceived = %d, want 0", m.GetBytesReceived())
	}
	if m.GetReadErrors() != 0 {
		t.Errorf("readErrors = %d, want 0", m.GetReadErrors())
	}
	if m.GetWriteErrors() != 0 {
		t.Errorf("writeErrors = %d, want 0", m.GetWriteErrors())
	}
	if m.GetHandlerErrors() != 0 {
		t.Errorf("handlerErrors = %d, want 0", m.GetHandlerErrors())
	}
	if m.GetActiveRooms() != 0 {
		t.Errorf("activeRooms = %d, want 0", m.GetActiveRooms())
	}
	if m.GetMissedPongs() != 0 {
		t.Errorf("missedPongs = %d, want 0", m.GetMissedPongs())
	}
	if m.GetSuccessfulPongs() != 0 {
		t.Errorf("successfulPongs = %d, want 0", m.GetSuccessfulPongs())
	}
	if m.GetQueueOverflows() != 0 {
		t.Errorf("queueOverflows = %d, want 0", m.GetQueueOverflows())
	}
	if m.GetDroppedMessages() != 0 {
		t.Errorf("droppedMessages = %d, want 0", m.GetDroppedMessages())
	}

	// Enabled by default
	if !m.IsEnabled() {
		t.Error("metrics should be enabled by default")
	}

	// Start time should be recent
	if time.Since(m.GetStartTime()) > time.Second {
		t.Error("start time should be recent")
	}

	// Last message time should be zero
	if !m.GetLastMessageTime().IsZero() {
		t.Error("lastMessageTime should be zero initially")
	}

	// Uptime should be non-negative
	if m.GetUptime() < 0 {
		t.Errorf("uptime should be non-negative, got %v", m.GetUptime())
	}
}

func TestMetrics_IncrementDecrementConnections(t *testing.T) {
	m := NewMetrics()

	m.IncrementConnections()
	if got := m.GetActiveConnections(); got != 1 {
		t.Errorf("activeConnections after increment = %d, want 1", got)
	}
	if got := m.GetTotalConnections(); got != 1 {
		t.Errorf("totalConnections after increment = %d, want 1", got)
	}

	m.IncrementConnections()
	if got := m.GetActiveConnections(); got != 2 {
		t.Errorf("activeConnections after second increment = %d, want 2", got)
	}
	if got := m.GetTotalConnections(); got != 2 {
		t.Errorf("totalConnections after second increment = %d, want 2", got)
	}

	m.DecrementConnections()
	if got := m.GetActiveConnections(); got != 1 {
		t.Errorf("activeConnections after decrement = %d, want 1", got)
	}
	if got := m.GetTotalDisconnections(); got != 1 {
		t.Errorf("totalDisconnections after decrement = %d, want 1", got)
	}

	// Total connections should still be 2 (not affected by decrement)
	if got := m.GetTotalConnections(); got != 2 {
		t.Errorf("totalConnections should remain 2, got %d", got)
	}
}

func TestMetrics_IncrementRejectedConnections(t *testing.T) {
	m := NewMetrics()

	m.IncrementRejectedConnections()
	if got := m.GetRejectedConnections(); got != 1 {
		t.Errorf("rejectedConnections = %d, want 1", got)
	}

	m.IncrementRejectedConnections()
	m.IncrementRejectedConnections()
	if got := m.GetRejectedConnections(); got != 3 {
		t.Errorf("rejectedConnections = %d, want 3", got)
	}
}

func TestMetrics_IncrementMessagesSent(t *testing.T) {
	m := NewMetrics()

	m.IncrementMessagesSent(100)
	if got := m.GetMessagesSent(); got != 1 {
		t.Errorf("messagesSent = %d, want 1", got)
	}
	if got := m.GetBytesSent(); got != 100 {
		t.Errorf("bytesSent = %d, want 100", got)
	}

	m.IncrementMessagesSent(250)
	if got := m.GetMessagesSent(); got != 2 {
		t.Errorf("messagesSent = %d, want 2", got)
	}
	if got := m.GetBytesSent(); got != 350 {
		t.Errorf("bytesSent = %d, want 350", got)
	}

	// lastMessageTime should be updated
	if m.GetLastMessageTime().IsZero() {
		t.Error("lastMessageTime should be non-zero after sending")
	}
}

func TestMetrics_IncrementMessagesReceived(t *testing.T) {
	m := NewMetrics()

	m.IncrementMessagesReceived(200)
	if got := m.GetMessagesReceived(); got != 1 {
		t.Errorf("messagesReceived = %d, want 1", got)
	}
	if got := m.GetBytesReceived(); got != 200 {
		t.Errorf("bytesReceived = %d, want 200", got)
	}

	m.IncrementMessagesReceived(300)
	if got := m.GetMessagesReceived(); got != 2 {
		t.Errorf("messagesReceived = %d, want 2", got)
	}
	if got := m.GetBytesReceived(); got != 500 {
		t.Errorf("bytesReceived = %d, want 500", got)
	}

	// lastMessageTime should be updated
	if m.GetLastMessageTime().IsZero() {
		t.Error("lastMessageTime should be non-zero after receiving")
	}
}

func TestMetrics_IncrementMessagesFailed(t *testing.T) {
	m := NewMetrics()

	m.IncrementMessagesFailed()
	if got := m.GetMessagesFailed(); got != 1 {
		t.Errorf("messagesFailed = %d, want 1", got)
	}

	m.IncrementMessagesFailed()
	if got := m.GetMessagesFailed(); got != 2 {
		t.Errorf("messagesFailed = %d, want 2", got)
	}
}

func TestMetrics_IncrementReadErrors(t *testing.T) {
	m := NewMetrics()

	m.IncrementReadErrors()
	if got := m.GetReadErrors(); got != 1 {
		t.Errorf("readErrors = %d, want 1", got)
	}

	m.IncrementReadErrors()
	m.IncrementReadErrors()
	if got := m.GetReadErrors(); got != 3 {
		t.Errorf("readErrors = %d, want 3", got)
	}
}

func TestMetrics_IncrementWriteErrors(t *testing.T) {
	m := NewMetrics()

	m.IncrementWriteErrors()
	if got := m.GetWriteErrors(); got != 1 {
		t.Errorf("writeErrors = %d, want 1", got)
	}

	m.IncrementWriteErrors()
	if got := m.GetWriteErrors(); got != 2 {
		t.Errorf("writeErrors = %d, want 2", got)
	}
}

func TestMetrics_IncrementHandlerErrors(t *testing.T) {
	m := NewMetrics()

	m.IncrementHandlerErrors()
	if got := m.GetHandlerErrors(); got != 1 {
		t.Errorf("handlerErrors = %d, want 1", got)
	}

	m.IncrementHandlerErrors()
	m.IncrementHandlerErrors()
	if got := m.GetHandlerErrors(); got != 3 {
		t.Errorf("handlerErrors = %d, want 3", got)
	}
}

func TestMetrics_GetTotalErrors(t *testing.T) {
	m := NewMetrics()

	m.IncrementReadErrors()
	m.IncrementReadErrors()
	m.IncrementWriteErrors()
	m.IncrementHandlerErrors()
	m.IncrementHandlerErrors()
	m.IncrementHandlerErrors()

	if got := m.GetTotalErrors(); got != 6 {
		t.Errorf("totalErrors = %d, want 6", got)
	}
}

func TestMetrics_IncrementDecrementRooms(t *testing.T) {
	m := NewMetrics()

	m.IncrementRooms()
	if got := m.GetActiveRooms(); got != 1 {
		t.Errorf("activeRooms = %d, want 1", got)
	}

	m.IncrementRooms()
	m.IncrementRooms()
	if got := m.GetActiveRooms(); got != 3 {
		t.Errorf("activeRooms = %d, want 3", got)
	}

	m.DecrementRooms()
	if got := m.GetActiveRooms(); got != 2 {
		t.Errorf("activeRooms after decrement = %d, want 2", got)
	}

	m.DecrementRooms()
	m.DecrementRooms()
	if got := m.GetActiveRooms(); got != 0 {
		t.Errorf("activeRooms after all decrements = %d, want 0", got)
	}
}

func TestMetrics_IncrementMissedPongs(t *testing.T) {
	m := NewMetrics()

	m.IncrementMissedPongs()
	if got := m.GetMissedPongs(); got != 1 {
		t.Errorf("missedPongs = %d, want 1", got)
	}

	m.IncrementMissedPongs()
	m.IncrementMissedPongs()
	if got := m.GetMissedPongs(); got != 3 {
		t.Errorf("missedPongs = %d, want 3", got)
	}
}

func TestMetrics_IncrementSuccessfulPongs(t *testing.T) {
	m := NewMetrics()

	m.IncrementSuccessfulPongs()
	if got := m.GetSuccessfulPongs(); got != 1 {
		t.Errorf("successfulPongs = %d, want 1", got)
	}

	m.IncrementSuccessfulPongs()
	if got := m.GetSuccessfulPongs(); got != 2 {
		t.Errorf("successfulPongs = %d, want 2", got)
	}
}

func TestMetrics_IncrementQueueOverflows(t *testing.T) {
	m := NewMetrics()

	m.IncrementQueueOverflows()
	if got := m.GetQueueOverflows(); got != 1 {
		t.Errorf("queueOverflows = %d, want 1", got)
	}

	m.IncrementQueueOverflows()
	m.IncrementQueueOverflows()
	if got := m.GetQueueOverflows(); got != 3 {
		t.Errorf("queueOverflows = %d, want 3", got)
	}
}

func TestMetrics_IncrementDroppedMessages(t *testing.T) {
	m := NewMetrics()

	m.IncrementDroppedMessages()
	if got := m.GetDroppedMessages(); got != 1 {
		t.Errorf("droppedMessages = %d, want 1", got)
	}

	m.IncrementDroppedMessages()
	if got := m.GetDroppedMessages(); got != 2 {
		t.Errorf("droppedMessages = %d, want 2", got)
	}
}

func TestMetrics_EnableDisable(t *testing.T) {
	m := NewMetrics()

	if !m.IsEnabled() {
		t.Error("metrics should be enabled by default")
	}

	m.Disable()
	if m.IsEnabled() {
		t.Error("metrics should be disabled after Disable()")
	}

	// All increment operations should be no-ops when disabled
	m.IncrementConnections()
	m.IncrementMessagesSent(100)
	m.IncrementMessagesReceived(200)
	m.IncrementReadErrors()
	m.IncrementWriteErrors()
	m.IncrementHandlerErrors()
	m.IncrementRooms()
	m.IncrementMissedPongs()
	m.IncrementSuccessfulPongs()
	m.IncrementQueueOverflows()
	m.IncrementDroppedMessages()
	m.IncrementRejectedConnections()
	m.IncrementMessagesFailed()
	m.DecrementConnections()
	m.DecrementRooms()

	if m.GetActiveConnections() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetMessagesSent() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetMessagesReceived() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetReadErrors() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetWriteErrors() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetHandlerErrors() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetActiveRooms() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetMissedPongs() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetSuccessfulPongs() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetQueueOverflows() != 0 {
		t.Error("metrics should not change when disabled")
	}
	if m.GetDroppedMessages() != 0 {
		t.Error("metrics should not change when disabled")
	}

	m.Enable()
	if !m.IsEnabled() {
		t.Error("metrics should be enabled after Enable()")
	}

	// Now operations should work
	m.IncrementConnections()
	if m.GetActiveConnections() != 1 {
		t.Errorf("activeConnections = %d, want 1 after re-enabling", m.GetActiveConnections())
	}
}

func TestMetrics_PerConnectionMetrics(t *testing.T) {
	m := NewMetrics()

	// Register a connection
	m.RegisterConnection("conn-1")
	cm := m.GetConnectionMetrics("conn-1")
	if cm == nil {
		t.Fatal("expected connection metrics after registration, got nil")
	}
	if cm.MessagesSent != 0 || cm.MessagesReceived != 0 {
		t.Error("new connection metrics should start at zero")
	}
	if cm.ConnectedAt.IsZero() {
		t.Error("ConnectedAt should be set on registration")
	}

	// Increment per-connection metrics
	m.IncrementConnectionMessagesSent("conn-1", 50)
	m.IncrementConnectionMessagesReceived("conn-1", 75)
	m.IncrementConnectionErrors("conn-1")
	m.IncrementConnectionMissedPongs("conn-1")

	cm = m.GetConnectionMetrics("conn-1")
	if cm.MessagesSent != 1 {
		t.Errorf("connection MessagesSent = %d, want 1", cm.MessagesSent)
	}
	if cm.BytesSent != 50 {
		t.Errorf("connection BytesSent = %d, want 50", cm.BytesSent)
	}
	if cm.MessagesReceived != 1 {
		t.Errorf("connection MessagesReceived = %d, want 1", cm.MessagesReceived)
	}
	if cm.BytesReceived != 75 {
		t.Errorf("connection BytesReceived = %d, want 75", cm.BytesReceived)
	}
	if cm.Errors != 1 {
		t.Errorf("connection Errors = %d, want 1", cm.Errors)
	}
	if cm.MissedPongs != 1 {
		t.Errorf("connection MissedPongs = %d, want 1", cm.MissedPongs)
	}

	// Unregister
	m.UnregisterConnection("conn-1")
	if m.GetConnectionMetrics("conn-1") != nil {
		t.Error("connection metrics should be nil after unregistration")
	}
}

func TestMetrics_PerConnectionMetrics_NonexistentConn(t *testing.T) {
	m := NewMetrics()

	// Operations on nonexistent connections should not panic
	m.IncrementConnectionMessagesSent("nonexistent", 100)
	m.IncrementConnectionMessagesReceived("nonexistent", 200)
	m.IncrementConnectionErrors("nonexistent")
	m.IncrementConnectionMissedPongs("nonexistent")

	if m.GetConnectionMetrics("nonexistent") != nil {
		t.Error("should return nil for nonexistent connection")
	}
}

func TestMetrics_PerConnectionMetrics_DisabledNoOp(t *testing.T) {
	m := NewMetrics()
	m.Disable()

	// Register should be no-op when disabled
	m.RegisterConnection("disabled-conn")
	if m.GetConnectionMetrics("disabled-conn") != nil {
		t.Error("RegisterConnection should be no-op when disabled")
	}

	// Re-enable, register, then disable and test increments
	m.Enable()
	m.RegisterConnection("test-conn")
	m.Disable()

	m.IncrementConnectionMessagesSent("test-conn", 100)
	m.IncrementConnectionMessagesReceived("test-conn", 200)
	m.IncrementConnectionErrors("test-conn")
	m.IncrementConnectionMissedPongs("test-conn")

	m.Enable()
	cm := m.GetConnectionMetrics("test-conn")
	if cm == nil {
		t.Fatal("connection should still exist")
	}
	if cm.MessagesSent != 0 {
		t.Error("per-connection sent should be 0 when increments were disabled")
	}
	if cm.MessagesReceived != 0 {
		t.Error("per-connection received should be 0 when increments were disabled")
	}
	if cm.Errors != 0 {
		t.Error("per-connection errors should be 0 when increments were disabled")
	}
	if cm.MissedPongs != 0 {
		t.Error("per-connection missed pongs should be 0 when increments were disabled")
	}

	// Unregister when disabled should be no-op
	m.Disable()
	m.UnregisterConnection("test-conn")
	m.Enable()
	if m.GetConnectionMetrics("test-conn") == nil {
		t.Error("UnregisterConnection should be no-op when disabled")
	}
}

func TestMetrics_GetAllConnectionMetrics(t *testing.T) {
	m := NewMetrics()

	m.RegisterConnection("conn-a")
	m.RegisterConnection("conn-b")
	m.RegisterConnection("conn-c")

	m.IncrementConnectionMessagesSent("conn-a", 10)
	m.IncrementConnectionMessagesSent("conn-b", 20)

	all := m.GetAllConnectionMetrics()
	if len(all) != 3 {
		t.Errorf("expected 3 connection metrics, got %d", len(all))
	}

	if all["conn-a"] == nil || all["conn-a"].BytesSent != 10 {
		t.Error("conn-a metrics incorrect")
	}
	if all["conn-b"] == nil || all["conn-b"].BytesSent != 20 {
		t.Error("conn-b metrics incorrect")
	}
	if all["conn-c"] == nil {
		t.Error("conn-c should exist in all metrics")
	}
}

func TestMetrics_GetAllConnectionMetrics_ReturnsCopies(t *testing.T) {
	m := NewMetrics()
	m.RegisterConnection("conn-copy")
	m.IncrementConnectionMessagesSent("conn-copy", 42)

	all := m.GetAllConnectionMetrics()
	// Modify the returned copy -- should not affect internal state
	all["conn-copy"].MessagesSent = 999

	cm := m.GetConnectionMetrics("conn-copy")
	if cm.MessagesSent != 1 {
		t.Errorf("modifying returned copy should not affect internal state, got MessagesSent=%d", cm.MessagesSent)
	}
}

func TestMetrics_GetConnectionMetrics_ReturnsCopy(t *testing.T) {
	m := NewMetrics()
	m.RegisterConnection("conn-copy2")
	m.IncrementConnectionErrors("conn-copy2")

	cm := m.GetConnectionMetrics("conn-copy2")
	cm.Errors = 999

	cm2 := m.GetConnectionMetrics("conn-copy2")
	if cm2.Errors != 1 {
		t.Errorf("modifying returned copy should not affect internal state, got Errors=%d", cm2.Errors)
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := NewMetrics()

	// Populate metrics
	m.IncrementConnections()
	m.IncrementConnections()
	m.DecrementConnections()
	m.IncrementRejectedConnections()
	m.IncrementMessagesSent(100)
	m.IncrementMessagesReceived(200)
	m.IncrementMessagesFailed()
	m.IncrementReadErrors()
	m.IncrementWriteErrors()
	m.IncrementHandlerErrors()
	m.IncrementRooms()
	m.IncrementMissedPongs()
	m.IncrementSuccessfulPongs()
	m.IncrementQueueOverflows()
	m.IncrementDroppedMessages()
	m.RegisterConnection("conn-reset")

	// Reset
	m.Reset()

	// Verify all counters are back to zero
	if m.GetActiveConnections() != 0 {
		t.Error("activeConnections should be 0 after reset")
	}
	if m.GetTotalConnections() != 0 {
		t.Error("totalConnections should be 0 after reset")
	}
	if m.GetTotalDisconnections() != 0 {
		t.Error("totalDisconnections should be 0 after reset")
	}
	if m.GetRejectedConnections() != 0 {
		t.Error("rejectedConnections should be 0 after reset")
	}
	if m.GetMessagesSent() != 0 {
		t.Error("messagesSent should be 0 after reset")
	}
	if m.GetMessagesReceived() != 0 {
		t.Error("messagesReceived should be 0 after reset")
	}
	if m.GetMessagesFailed() != 0 {
		t.Error("messagesFailed should be 0 after reset")
	}
	if m.GetBytesSent() != 0 {
		t.Error("bytesSent should be 0 after reset")
	}
	if m.GetBytesReceived() != 0 {
		t.Error("bytesReceived should be 0 after reset")
	}
	if m.GetReadErrors() != 0 {
		t.Error("readErrors should be 0 after reset")
	}
	if m.GetWriteErrors() != 0 {
		t.Error("writeErrors should be 0 after reset")
	}
	if m.GetHandlerErrors() != 0 {
		t.Error("handlerErrors should be 0 after reset")
	}
	if m.GetActiveRooms() != 0 {
		t.Error("activeRooms should be 0 after reset")
	}
	if m.GetMissedPongs() != 0 {
		t.Error("missedPongs should be 0 after reset")
	}
	if m.GetSuccessfulPongs() != 0 {
		t.Error("successfulPongs should be 0 after reset")
	}
	if m.GetQueueOverflows() != 0 {
		t.Error("queueOverflows should be 0 after reset")
	}
	if m.GetDroppedMessages() != 0 {
		t.Error("droppedMessages should be 0 after reset")
	}
	if !m.GetLastMessageTime().IsZero() {
		t.Error("lastMessageTime should be zero after reset")
	}

	// Per-connection metrics should be cleared
	if m.GetConnectionMetrics("conn-reset") != nil {
		t.Error("per-connection metrics should be cleared after reset")
	}

	// Start time should be refreshed
	if time.Since(m.GetStartTime()) > time.Second {
		t.Error("start time should be refreshed after reset")
	}
}

func TestMetrics_GetSnapshot(t *testing.T) {
	m := NewMetrics()

	// Populate various metrics
	m.IncrementConnections()
	m.IncrementConnections()
	m.IncrementConnections()
	m.DecrementConnections()
	m.IncrementRejectedConnections()
	m.IncrementMessagesSent(100)
	m.IncrementMessagesSent(200)
	m.IncrementMessagesReceived(300)
	m.IncrementMessagesFailed()
	m.IncrementReadErrors()
	m.IncrementReadErrors()
	m.IncrementWriteErrors()
	m.IncrementHandlerErrors()
	m.IncrementRooms()
	m.IncrementRooms()
	m.IncrementMissedPongs()
	m.IncrementSuccessfulPongs()
	m.IncrementSuccessfulPongs()
	m.IncrementQueueOverflows()
	m.IncrementDroppedMessages()
	m.IncrementDroppedMessages()
	m.IncrementDroppedMessages()

	snapshot := m.GetSnapshot()
	if snapshot == nil {
		t.Fatal("GetSnapshot returned nil")
	}

	if snapshot.ActiveConnections != 2 {
		t.Errorf("ActiveConnections = %d, want 2", snapshot.ActiveConnections)
	}
	if snapshot.TotalConnections != 3 {
		t.Errorf("TotalConnections = %d, want 3", snapshot.TotalConnections)
	}
	if snapshot.TotalDisconnections != 1 {
		t.Errorf("TotalDisconnections = %d, want 1", snapshot.TotalDisconnections)
	}
	if snapshot.RejectedConnections != 1 {
		t.Errorf("RejectedConnections = %d, want 1", snapshot.RejectedConnections)
	}
	if snapshot.MessagesSent != 2 {
		t.Errorf("MessagesSent = %d, want 2", snapshot.MessagesSent)
	}
	if snapshot.MessagesReceived != 1 {
		t.Errorf("MessagesReceived = %d, want 1", snapshot.MessagesReceived)
	}
	if snapshot.MessagesFailed != 1 {
		t.Errorf("MessagesFailed = %d, want 1", snapshot.MessagesFailed)
	}
	if snapshot.BytesSent != 300 {
		t.Errorf("BytesSent = %d, want 300", snapshot.BytesSent)
	}
	if snapshot.BytesReceived != 300 {
		t.Errorf("BytesReceived = %d, want 300", snapshot.BytesReceived)
	}
	if snapshot.ReadErrors != 2 {
		t.Errorf("ReadErrors = %d, want 2", snapshot.ReadErrors)
	}
	if snapshot.WriteErrors != 1 {
		t.Errorf("WriteErrors = %d, want 1", snapshot.WriteErrors)
	}
	if snapshot.HandlerErrors != 1 {
		t.Errorf("HandlerErrors = %d, want 1", snapshot.HandlerErrors)
	}
	if snapshot.TotalErrors != 4 {
		t.Errorf("TotalErrors = %d, want 4", snapshot.TotalErrors)
	}
	if snapshot.ActiveRooms != 2 {
		t.Errorf("ActiveRooms = %d, want 2", snapshot.ActiveRooms)
	}
	if snapshot.MissedPongs != 1 {
		t.Errorf("MissedPongs = %d, want 1", snapshot.MissedPongs)
	}
	if snapshot.SuccessfulPongs != 2 {
		t.Errorf("SuccessfulPongs = %d, want 2", snapshot.SuccessfulPongs)
	}
	if snapshot.QueueOverflows != 1 {
		t.Errorf("QueueOverflows = %d, want 1", snapshot.QueueOverflows)
	}
	if snapshot.DroppedMessages != 3 {
		t.Errorf("DroppedMessages = %d, want 3", snapshot.DroppedMessages)
	}
	if snapshot.LastMessageTime.IsZero() {
		t.Error("LastMessageTime should be non-zero")
	}
	if snapshot.StartTime.IsZero() {
		t.Error("StartTime should be non-zero")
	}
	if snapshot.Uptime <= 0 {
		t.Error("Uptime should be positive")
	}

	// Rate calculations: MessagesPerSecond and ConnectionsPerSecond
	// With uptime > 0, these should be non-negative
	if snapshot.MessagesPerSecond < 0 {
		t.Error("MessagesPerSecond should be non-negative")
	}
	if snapshot.ConnectionsPerSecond < 0 {
		t.Error("ConnectionsPerSecond should be non-negative")
	}
}

func TestMetrics_GetSnapshot_ZeroUptime(t *testing.T) {
	// Immediately after creation, uptime is extremely small but > 0
	// so we just verify no division by zero panic and rates are >= 0
	m := NewMetrics()
	snapshot := m.GetSnapshot()
	if snapshot.MessagesPerSecond < 0 {
		t.Error("MessagesPerSecond should be non-negative even at tiny uptime")
	}
	if snapshot.ConnectionsPerSecond < 0 {
		t.Error("ConnectionsPerSecond should be non-negative even at tiny uptime")
	}
}

func TestMetrics_TimeMetrics(t *testing.T) {
	m := NewMetrics()

	startTime := m.GetStartTime()
	if startTime.IsZero() {
		t.Error("startTime should not be zero")
	}

	// lastMessageTime should be zero initially
	if !m.GetLastMessageTime().IsZero() {
		t.Error("lastMessageTime should be zero initially")
	}

	beforeSend := time.Now()
	m.IncrementMessagesSent(10)
	afterSend := time.Now()

	lastMsg := m.GetLastMessageTime()
	if lastMsg.Before(beforeSend) || lastMsg.After(afterSend) {
		t.Error("lastMessageTime should be between before and after send")
	}

	// Test that received messages also update the timestamp
	beforeRecv := time.Now()
	m.IncrementMessagesReceived(20)
	afterRecv := time.Now()

	lastMsg = m.GetLastMessageTime()
	if lastMsg.Before(beforeRecv) || lastMsg.After(afterRecv) {
		t.Error("lastMessageTime should be updated by received messages")
	}

	// Uptime should be positive
	if m.GetUptime() <= 0 {
		t.Error("uptime should be positive")
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup
	iterations := 100

	// Launch many goroutines that concurrently modify all metric types
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				m.IncrementConnections()
				m.DecrementConnections()
				m.IncrementRejectedConnections()
				m.IncrementMessagesSent(1)
				m.IncrementMessagesReceived(1)
				m.IncrementMessagesFailed()
				m.IncrementReadErrors()
				m.IncrementWriteErrors()
				m.IncrementHandlerErrors()
				m.IncrementRooms()
				m.DecrementRooms()
				m.IncrementMissedPongs()
				m.IncrementSuccessfulPongs()
				m.IncrementQueueOverflows()
				m.IncrementDroppedMessages()
			}
		}()
	}

	// Also concurrently read metrics
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = m.GetActiveConnections()
				_ = m.GetTotalConnections()
				_ = m.GetTotalDisconnections()
				_ = m.GetRejectedConnections()
				_ = m.GetMessagesSent()
				_ = m.GetMessagesReceived()
				_ = m.GetMessagesFailed()
				_ = m.GetBytesSent()
				_ = m.GetBytesReceived()
				_ = m.GetReadErrors()
				_ = m.GetWriteErrors()
				_ = m.GetHandlerErrors()
				_ = m.GetTotalErrors()
				_ = m.GetActiveRooms()
				_ = m.GetMissedPongs()
				_ = m.GetSuccessfulPongs()
				_ = m.GetQueueOverflows()
				_ = m.GetDroppedMessages()
				_ = m.GetLastMessageTime()
				_ = m.GetStartTime()
				_ = m.GetUptime()
				_ = m.GetSnapshot()
			}
		}()
	}

	wg.Wait()

	// After all goroutines complete:
	// 10 goroutines * 100 iterations each
	expectedTotal := int64(10 * iterations)

	// activeConnections should be 0 (same number of increments and decrements)
	if got := m.GetActiveConnections(); got != 0 {
		t.Errorf("activeConnections = %d, want 0 (increments == decrements)", got)
	}

	// totalConnections should be exactly expectedTotal
	if got := m.GetTotalConnections(); got != expectedTotal {
		t.Errorf("totalConnections = %d, want %d", got, expectedTotal)
	}

	// rejectedConnections should be exactly expectedTotal
	if got := m.GetRejectedConnections(); got != expectedTotal {
		t.Errorf("rejectedConnections = %d, want %d", got, expectedTotal)
	}

	// activeRooms should be 0 (same number of increments and decrements)
	if got := m.GetActiveRooms(); got != 0 {
		t.Errorf("activeRooms = %d, want 0", got)
	}
}

func TestMetrics_ConcurrentPerConnectionAccess(t *testing.T) {
	m := NewMetrics()
	m.RegisterConnection("concurrent-conn")

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				m.IncrementConnectionMessagesSent("concurrent-conn", 1)
				m.IncrementConnectionMessagesReceived("concurrent-conn", 1)
				m.IncrementConnectionErrors("concurrent-conn")
				m.IncrementConnectionMissedPongs("concurrent-conn")
			}
		}()
	}

	// Also concurrently read per-connection metrics
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = m.GetConnectionMetrics("concurrent-conn")
				_ = m.GetAllConnectionMetrics()
			}
		}()
	}

	wg.Wait()

	cm := m.GetConnectionMetrics("concurrent-conn")
	if cm == nil {
		t.Fatal("concurrent connection metrics should exist")
	}

	expectedTotal := int64(10 * iterations)
	if cm.Errors != expectedTotal {
		t.Errorf("connection Errors = %d, want %d", cm.Errors, expectedTotal)
	}
	if cm.MissedPongs != expectedTotal {
		t.Errorf("connection MissedPongs = %d, want %d", cm.MissedPongs, expectedTotal)
	}
}

func TestMetrics_ConcurrentRegisterUnregister(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup

	// Concurrently register and unregister connections
	for i := 0; i < 50; i++ {
		connID := "conn-" + string(rune('A'+i%26))
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			m.RegisterConnection(id)
			m.IncrementConnectionMessagesSent(id, 10)
			_ = m.GetConnectionMetrics(id)
			m.UnregisterConnection(id)
		}(connID)
	}

	wg.Wait()
	// Should not panic or deadlock
}

func TestMetrics_ConcurrentEnableDisable(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup

	// Concurrently enable/disable and do operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if idx%2 == 0 {
					m.Enable()
				} else {
					m.Disable()
				}
				m.IncrementConnections()
				m.DecrementConnections()
				_ = m.IsEnabled()
			}
		}(i)
	}

	wg.Wait()
	// Should not panic or deadlock
}

func TestMetrics_ResetThenConcurrentAccess(t *testing.T) {
	m := NewMetrics()

	// Populate, reset, then verify concurrent access after reset
	m.IncrementConnections()
	m.IncrementMessagesSent(100)
	m.RegisterConnection("pre-reset")
	m.Reset()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				m.IncrementConnections()
				m.DecrementConnections()
				m.IncrementMessagesSent(1)
				_ = m.GetSnapshot()
			}
		}()
	}

	wg.Wait()
	// Should not panic or deadlock after reset
}
