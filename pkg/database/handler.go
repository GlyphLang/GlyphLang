package database

import (
	"context"
	"fmt"
	"sync"
)

// Handler manages database connections and operations for the interpreter
type Handler struct {
	db     Database
	mu     sync.RWMutex
	tables map[string]*TableHandler
	ctx    context.Context
}

// NewHandler creates a new database handler
func NewHandler(db Database) *Handler {
	return &Handler{
		db:     db,
		tables: make(map[string]*TableHandler),
		ctx:    context.Background(),
	}
}

// NewHandlerFromString creates a new handler from a connection string
func NewHandlerFromString(connStr string) (*Handler, error) {
	db, err := NewDatabaseFromString(connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return NewHandler(db), nil
}

// Table returns a table handler for the given table name
func (h *Handler) Table(name string) *TableHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler, ok := h.tables[name]; ok {
		return handler
	}

	handler := &TableHandler{
		db:   h.db,
		orm:  NewORM(h.db, name),
		name: name,
		ctx:  h.ctx,
	}

	h.tables[name] = handler
	return handler
}

// Close closes the database connection
func (h *Handler) Close() error {
	return h.db.Close()
}

// TableHandler provides high-level database operations for GLYPH
type TableHandler struct {
	db   Database
	orm  *ORM
	name string
	ctx  context.Context
}

// All retrieves all records from the table
func (t *TableHandler) All() ([]map[string]interface{}, error) {
	return t.orm.FindAll(t.ctx)
}

// Get retrieves a single record by ID
func (t *TableHandler) Get(id interface{}) (map[string]interface{}, error) {
	result, err := t.orm.FindByID(t.ctx, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates a new record
func (t *TableHandler) Create(data map[string]interface{}) (map[string]interface{}, error) {
	return t.orm.Create(t.ctx, data)
}

// Update updates a record by ID
func (t *TableHandler) Update(id interface{}, data map[string]interface{}) (map[string]interface{}, error) {
	return t.orm.Update(t.ctx, id, data)
}

// Delete deletes a record by ID
func (t *TableHandler) Delete(id interface{}) error {
	return t.orm.Delete(t.ctx, id)
}

// Count counts records matching a condition
func (t *TableHandler) Count(column string, value interface{}) (int64, error) {
	return t.orm.Count(t.ctx, WhereCondition{
		Column:   column,
		Operator: "=",
		Value:    value,
	})
}

// CountWhere counts records matching multiple conditions
func (t *TableHandler) CountWhere(conditions ...interface{}) (int64, error) {
	whereConds := make([]WhereCondition, 0)

	// Parse conditions in pairs: column, value, column, value, ...
	for i := 0; i < len(conditions); i += 2 {
		if i+1 >= len(conditions) {
			return 0, fmt.Errorf("invalid where conditions: expected pairs of column, value")
		}

		column, ok := conditions[i].(string)
		if !ok {
			return 0, fmt.Errorf("expected string for column name, got %T", conditions[i])
		}

		whereConds = append(whereConds, WhereCondition{
			Column:   column,
			Operator: "=",
			Value:    conditions[i+1],
		})
	}

	return t.orm.Count(t.ctx, whereConds...)
}

// Filter retrieves records matching a condition
func (t *TableHandler) Filter(column string, value interface{}) ([]map[string]interface{}, error) {
	qb := t.orm.NewQueryBuilder().WhereEq(column, value)
	return qb.Get(t.ctx)
}

// Query executes a raw SQL query
func (t *TableHandler) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return t.orm.Query(t.ctx, query, args...)
}

// NextId returns the count of existing records plus one as an approximation.
// For true sequence values, use database-specific sequence queries.
func (t *TableHandler) NextId() (result int64) {
	defer func() {
		if r := recover(); r != nil {
			result = 1
		}
	}()
	count, err := t.orm.Count(t.ctx)
	if err != nil {
		return 1
	}
	return count + 1
}

// Length returns the total count of records
func (t *TableHandler) Length() (int64, error) {
	return t.orm.Count(t.ctx)
}

// Exists checks if a record exists
func (t *TableHandler) Exists(column string, value interface{}) (bool, error) {
	return t.orm.Exists(t.ctx, WhereCondition{
		Column:   column,
		Operator: "=",
		Value:    value,
	})
}

// Where creates a query builder with a WHERE condition
func (t *TableHandler) Where(column string, operator string, value interface{}) *QueryBuilder {
	return t.orm.NewQueryBuilder().Where(column, operator, value)
}

// FindWhere retrieves records matching conditions
func (t *TableHandler) FindWhere(column string, value interface{}) ([]map[string]interface{}, error) {
	return t.Filter(column, value)
}

// First retrieves the first record
func (t *TableHandler) First() (map[string]interface{}, error) {
	return t.orm.NewQueryBuilder().Limit(1).First(t.ctx)
}

// Last retrieves the last record (ordered by ID descending)
func (t *TableHandler) Last() (map[string]interface{}, error) {
	return t.orm.NewQueryBuilder().OrderBy("id", "DESC").Limit(1).First(t.ctx)
}

// MockDatabase represents a mock database for testing without actual DB connection
type MockDatabase struct {
	data map[string][]map[string]interface{}
	mu   sync.RWMutex
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		data: make(map[string][]map[string]interface{}),
	}
}

// Table returns a mock table handler
func (m *MockDatabase) Table(name string) *MockTableHandler {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[name]; !ok {
		m.data[name] = make([]map[string]interface{}, 0)
	}

	return &MockTableHandler{
		db:   m,
		name: name,
	}
}

// MockTableHandler provides mock database operations
type MockTableHandler struct {
	db   *MockDatabase
	name string
}

// All retrieves all records
func (m *MockTableHandler) All() []interface{} {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	data := m.db.data[m.name]
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = v
	}
	return result
}

// Get retrieves a record by ID
func (m *MockTableHandler) Get(id interface{}) interface{} {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	for _, record := range m.db.data[m.name] {
		if record["id"] == id {
			return record
		}
	}
	return nil
}

// Create creates a new record
func (m *MockTableHandler) Create(data map[string]interface{}) map[string]interface{} {
	m.db.mu.Lock()
	defer m.db.mu.Unlock()

	// Auto-generate ID if not provided
	if _, ok := data["id"]; !ok {
		data["id"] = int64(len(m.db.data[m.name]) + 1)
	}

	m.db.data[m.name] = append(m.db.data[m.name], data)
	return data
}

// Update updates a record by ID
func (m *MockTableHandler) Update(id interface{}, data map[string]interface{}) map[string]interface{} {
	m.db.mu.Lock()
	defer m.db.mu.Unlock()

	for i, record := range m.db.data[m.name] {
		if record["id"] == id {
			// Merge data
			for k, v := range data {
				record[k] = v
			}
			m.db.data[m.name][i] = record
			return record
		}
	}
	return nil
}

// Delete deletes a record by ID
func (m *MockTableHandler) Delete(id interface{}) bool {
	m.db.mu.Lock()
	defer m.db.mu.Unlock()

	for i, record := range m.db.data[m.name] {
		if record["id"] == id {
			m.db.data[m.name] = append(m.db.data[m.name][:i], m.db.data[m.name][i+1:]...)
			return true
		}
	}
	return false
}

// Count counts records matching a condition
func (m *MockTableHandler) Count(column string, value interface{}) int64 {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	count := int64(0)
	for _, record := range m.db.data[m.name] {
		if record[column] == value {
			count++
		}
	}
	return count
}

// CountWhere counts records matching multiple conditions.
func (m *MockTableHandler) CountWhere(column1 string, value1 interface{}, column2 string, value2 interface{}) int64 {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	count := int64(0)
	for _, record := range m.db.data[m.name] {
		if record[column1] == value1 && record[column2] == value2 {
			count++
		}
	}
	return count
}

// Filter retrieves records matching a condition
func (m *MockTableHandler) Filter(column string, value interface{}) []interface{} {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	result := make([]interface{}, 0)
	for _, record := range m.db.data[m.name] {
		if record[column] == value {
			result = append(result, record)
		}
	}
	return result
}

// NextId returns the next available ID.
func (m *MockTableHandler) NextId() int64 {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	return int64(len(m.db.data[m.name]) + 1)
}

// Length returns the total count of records
func (m *MockTableHandler) Length() int64 {
	m.db.mu.RLock()
	defer m.db.mu.RUnlock()

	return int64(len(m.db.data[m.name]))
}
