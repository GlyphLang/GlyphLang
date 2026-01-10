package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisType tests the Redis type marker
func TestRedisType(t *testing.T) {
	redisType := RedisType{}
	redisType.isType() // Verify it implements Type interface
}

// TestRedisInjection tests Redis dependency injection in routes
func TestRedisInjection(t *testing.T) {
	interp := NewInterpreter()

	// Create a mock Redis handler
	mockRedis := map[string]interface{}{
		"testValue": "cached_data",
	}
	interp.SetRedisHandler(mockRedis)

	// Create a route with Redis injection
	route := &Route{
		Method: Get,
		Path:   "/cache",
		Injections: []Injection{
			{Name: "redis", Type: RedisType{}},
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "redis"}},
		},
	}

	// Execute the route
	result, err := interp.ExecuteRoute(route, &Request{
		Path:   "/cache",
		Method: "GET",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, mockRedis, result.Body)
}

// TestRedisInjectionInCronTask tests Redis injection in cron tasks
func TestRedisInjectionInCronTask(t *testing.T) {
	interp := NewInterpreter()

	// Create a mock Redis handler
	mockRedis := map[string]interface{}{
		"counter": int64(100),
	}
	interp.SetRedisHandler(mockRedis)

	// Create a cron task with Redis injection
	task := &CronTask{
		Name:     "cache_cleanup",
		Schedule: "0 * * * *",
		Injections: []Injection{
			{Name: "redis", Type: RedisType{}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "count",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "redis"},
					Field:  "counter",
				},
			},
			ReturnStatement{Value: VariableExpr{Name: "count"}},
		},
	}

	// Execute the task
	result, err := interp.ExecuteCronTask(task)

	require.NoError(t, err)
	assert.Equal(t, int64(100), result)
}

// TestBothDatabaseAndRedisInjection tests using both Database and Redis in same route
func TestBothDatabaseAndRedisInjection(t *testing.T) {
	interp := NewInterpreter()

	// Create mock handlers
	mockDB := map[string]interface{}{
		"tableName": "users",
	}
	mockRedis := map[string]interface{}{
		"cacheKey": "user_cache",
	}

	interp.SetDatabaseHandler(mockDB)
	interp.SetRedisHandler(mockRedis)

	// Create a route with both injections
	route := &Route{
		Method: Get,
		Path:   "/data",
		Injections: []Injection{
			{Name: "db", Type: DatabaseType{}},
			{Name: "redis", Type: RedisType{}},
		},
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "db", Value: VariableExpr{Name: "db"}},
						{Key: "redis", Value: VariableExpr{Name: "redis"}},
					},
				},
			},
		},
	}

	// Execute the route
	result, err := interp.ExecuteRoute(route, &Request{
		Path:   "/data",
		Method: "GET",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)

	body, ok := result.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, mockDB, body["db"])
	assert.Equal(t, mockRedis, body["redis"])
}

// TestRedisInjectionInEventHandler tests Redis injection in event handlers
func TestRedisInjectionInEventHandler(t *testing.T) {
	interp := NewInterpreter()

	// Create a mock Redis handler
	mockRedis := map[string]interface{}{
		"channel": "notifications",
	}
	interp.SetRedisHandler(mockRedis)

	// Create an event handler with Redis injection
	handler := &EventHandler{
		EventType: "cache.invalidate",
		Injections: []Injection{
			{Name: "redis", Type: RedisType{}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "channel",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "redis"},
					Field:  "channel",
				},
			},
			ReturnStatement{Value: VariableExpr{Name: "channel"}},
		},
	}

	// Execute the handler
	eventData := map[string]interface{}{"key": "user:123"}
	result, err := interp.ExecuteEventHandler(handler, eventData)

	require.NoError(t, err)
	assert.Equal(t, "notifications", result)
}

// TestRedisInjectionInQueueWorker tests Redis injection in queue workers
func TestRedisInjectionInQueueWorker(t *testing.T) {
	interp := NewInterpreter()

	// Create a mock Redis handler
	mockRedis := map[string]interface{}{
		"queueName": "email_queue",
	}
	interp.SetRedisHandler(mockRedis)

	// Create a queue worker with Redis injection
	worker := &QueueWorker{
		QueueName: "email.send",
		Injections: []Injection{
			{Name: "redis", Type: RedisType{}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "queue",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "redis"},
					Field:  "queueName",
				},
			},
			ReturnStatement{Value: VariableExpr{Name: "queue"}},
		},
	}

	// Execute the worker
	message := map[string]interface{}{"to": "user@example.com"}
	result, err := interp.ExecuteQueueWorker(worker, message)

	require.NoError(t, err)
	assert.Equal(t, "email_queue", result)
}
