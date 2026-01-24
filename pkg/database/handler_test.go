package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockDatabase(t *testing.T) {
	db := NewMockDatabase()
	assert.NotNil(t, db)
	assert.NotNil(t, db.data)
}

func TestMockTableHandler_Create(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	t.Run("Create with ID", func(t *testing.T) {
		user := map[string]interface{}{
			"id":    int64(1),
			"name":  "John",
			"email": "john@example.com",
		}

		result := users.Create(user)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result["id"])
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, "john@example.com", result["email"])
	})

	t.Run("Create without ID (auto-generated)", func(t *testing.T) {
		user := map[string]interface{}{
			"name":  "Jane",
			"email": "jane@example.com",
		}

		result := users.Create(user)
		assert.NotNil(t, result)
		assert.NotNil(t, result["id"])
		assert.Equal(t, "Jane", result["name"])
	})
}

func TestMockTableHandler_Get(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Create a user
	user := map[string]interface{}{
		"id":    int64(1),
		"name":  "John",
		"email": "john@example.com",
	}
	users.Create(user)

	t.Run("Get existing user", func(t *testing.T) {
		result := users.Get(int64(1))
		assert.NotNil(t, result)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, int64(1), resultMap["id"])
		assert.Equal(t, "John", resultMap["name"])
	})

	t.Run("Get non-existing user", func(t *testing.T) {
		result := users.Get(int64(999))
		assert.Nil(t, result)
	})
}

func TestMockTableHandler_Update(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Create a user
	user := map[string]interface{}{
		"id":    int64(1),
		"name":  "John",
		"email": "john@example.com",
	}
	users.Create(user)

	t.Run("Update existing user", func(t *testing.T) {
		updates := map[string]interface{}{
			"name":  "John Doe",
			"email": "johndoe@example.com",
		}

		result := users.Update(int64(1), updates)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result["id"])
		assert.Equal(t, "John Doe", result["name"])
		assert.Equal(t, "johndoe@example.com", result["email"])
	})

	t.Run("Update non-existing user", func(t *testing.T) {
		updates := map[string]interface{}{
			"name": "Ghost",
		}

		result := users.Update(int64(999), updates)
		assert.Nil(t, result)
	})
}

func TestMockTableHandler_Delete(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Create users
	users.Create(map[string]interface{}{
		"id":   int64(1),
		"name": "John",
	})
	users.Create(map[string]interface{}{
		"id":   int64(2),
		"name": "Jane",
	})

	t.Run("Delete existing user", func(t *testing.T) {
		result := users.Delete(int64(1))
		assert.True(t, result)

		// Verify user is deleted
		getResult := users.Get(int64(1))
		assert.Nil(t, getResult)
	})

	t.Run("Delete non-existing user", func(t *testing.T) {
		result := users.Delete(int64(999))
		assert.False(t, result)
	})
}

func TestMockTableHandler_All(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	t.Run("Empty table", func(t *testing.T) {
		result := users.All()
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})

	t.Run("With records", func(t *testing.T) {
		users.Create(map[string]interface{}{
			"id":   int64(1),
			"name": "John",
		})
		users.Create(map[string]interface{}{
			"id":   int64(2),
			"name": "Jane",
		})

		result := users.All()
		assert.Equal(t, 2, len(result))
	})
}

func TestMockTableHandler_Count(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Create users with different statuses
	users.Create(map[string]interface{}{
		"id":     int64(1),
		"name":   "John",
		"status": "active",
	})
	users.Create(map[string]interface{}{
		"id":     int64(2),
		"name":   "Jane",
		"status": "active",
	})
	users.Create(map[string]interface{}{
		"id":     int64(3),
		"name":   "Bob",
		"status": "inactive",
	})

	t.Run("Count by status", func(t *testing.T) {
		count := users.Count("status", "active")
		assert.Equal(t, int64(2), count)
	})

	t.Run("Count by different status", func(t *testing.T) {
		count := users.Count("status", "inactive")
		assert.Equal(t, int64(1), count)
	})

	t.Run("Count non-existing", func(t *testing.T) {
		count := users.Count("status", "deleted")
		assert.Equal(t, int64(0), count)
	})
}

func TestMockTableHandler_CountWhere(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Create users
	users.Create(map[string]interface{}{
		"id":       int64(1),
		"name":     "John",
		"status":   "active",
		"verified": true,
	})
	users.Create(map[string]interface{}{
		"id":       int64(2),
		"name":     "Jane",
		"status":   "active",
		"verified": false,
	})
	users.Create(map[string]interface{}{
		"id":       int64(3),
		"name":     "Bob",
		"status":   "inactive",
		"verified": true,
	})

	t.Run("Count with multiple conditions", func(t *testing.T) {
		count := users.CountWhere("status", "active", "verified", true)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Count with no matches", func(t *testing.T) {
		count := users.CountWhere("status", "inactive", "verified", false)
		assert.Equal(t, int64(0), count)
	})
}

func TestMockTableHandler_Filter(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Create users
	users.Create(map[string]interface{}{
		"id":     int64(1),
		"name":   "John",
		"status": "active",
	})
	users.Create(map[string]interface{}{
		"id":     int64(2),
		"name":   "Jane",
		"status": "active",
	})
	users.Create(map[string]interface{}{
		"id":     int64(3),
		"name":   "Bob",
		"status": "inactive",
	})

	t.Run("Filter by status", func(t *testing.T) {
		result := users.Filter("status", "active")
		assert.Equal(t, 2, len(result))
	})

	t.Run("Filter with no matches", func(t *testing.T) {
		result := users.Filter("status", "deleted")
		assert.Equal(t, 0, len(result))
	})
}

func TestMockTableHandler_NextID(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	t.Run("Empty table", func(t *testing.T) {
		nextID := users.NextId()
		assert.Equal(t, int64(1), nextID)
	})

	t.Run("After creating records", func(t *testing.T) {
		users.Create(map[string]interface{}{"name": "John"})
		users.Create(map[string]interface{}{"name": "Jane"})

		nextID := users.NextId()
		assert.Equal(t, int64(3), nextID)
	})
}

func TestMockTableHandler_Length(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	t.Run("Empty table", func(t *testing.T) {
		length := users.Length()
		assert.Equal(t, int64(0), length)
	})

	t.Run("After creating records", func(t *testing.T) {
		users.Create(map[string]interface{}{"name": "John"})
		users.Create(map[string]interface{}{"name": "Jane"})

		length := users.Length()
		assert.Equal(t, int64(2), length)
	})
}

func TestNewHandler(t *testing.T) {
	mockDB := &MockDB{}
	handler := NewHandler(mockDB)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.db)
	assert.NotNil(t, handler.tables)
}

func TestHandler_Table(t *testing.T) {
	mockDB := &MockDB{}
	handler := NewHandler(mockDB)

	t.Run("Get table handler", func(t *testing.T) {
		usersTable := handler.Table("users")
		assert.NotNil(t, usersTable)
		assert.Equal(t, "users", usersTable.name)
	})

	t.Run("Get same table twice returns same handler", func(t *testing.T) {
		table1 := handler.Table("users")
		table2 := handler.Table("users")
		assert.Equal(t, table1, table2)
	})

	t.Run("Different tables return different handlers", func(t *testing.T) {
		usersTable := handler.Table("users")
		postsTable := handler.Table("posts")
		assert.NotEqual(t, usersTable, postsTable)
	})
}

func TestMultipleTables(t *testing.T) {
	db := NewMockDatabase()

	users := db.Table("users")
	posts := db.Table("posts")

	// Create data in different tables
	users.Create(map[string]interface{}{
		"id":   int64(1),
		"name": "John",
	})

	posts.Create(map[string]interface{}{
		"id":      int64(1),
		"title":   "First Post",
		"user_id": int64(1),
	})

	// Verify data is in correct tables
	assert.Equal(t, int64(1), users.Length())
	assert.Equal(t, int64(1), posts.Length())

	userRecord := users.Get(int64(1))
	assert.NotNil(t, userRecord)

	postRecord := posts.Get(int64(1))
	assert.NotNil(t, postRecord)
}

func TestConcurrentAccess(t *testing.T) {
	db := NewMockDatabase()
	users := db.Table("users")

	// Test that concurrent access doesn't cause data races
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 10; i++ {
			users.Create(map[string]interface{}{
				"id":   int64(i),
				"name": "User",
			})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 10; i++ {
			users.All()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify data integrity
	assert.Equal(t, int64(10), users.Length())
}
