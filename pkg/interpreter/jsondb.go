package interpreter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// JSONDatabase is a simple JSON file-based NoSQL database
type JSONDatabase struct {
	basePath string
	tables   map[string]*JSONTable
	mu       sync.RWMutex
}

// JSONTable represents a single table/collection
type JSONTable struct {
	name    string
	db      *JSONDatabase
	records []map[string]interface{}
	nextID  int
	mu      sync.RWMutex
}

// NewJSONDatabase creates a new JSON-based database
func NewJSONDatabase(basePath string) *JSONDatabase {
	db := &JSONDatabase{
		basePath: basePath,
		tables:   make(map[string]*JSONTable),
	}
	return db
}

// Table returns a table handler for the given table name
func (db *JSONDatabase) Table(name string) interface{} {
	db.mu.Lock()
	defer db.mu.Unlock()

	if table, exists := db.tables[name]; exists {
		return table
	}

	// Create new table and try to load from file
	table := &JSONTable{
		name:    name,
		db:      db,
		records: []map[string]interface{}{},
		nextID:  1,
	}

	// Try to load existing data
	table.load()

	db.tables[name] = table
	return table
}

// Load seeds the database from a JSON file
func (db *JSONDatabase) LoadSeed(seedPath string) error {
	data, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	var seedData map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("failed to parse seed file: %w", err)
	}

	for tableName, records := range seedData {
		table := db.Table(tableName).(*JSONTable)
		table.mu.Lock()
		table.records = records
		// Find max ID
		maxID := 0
		for _, record := range records {
			if id, ok := record["id"].(float64); ok {
				if int(id) > maxID {
					maxID = int(id)
				}
			}
		}
		table.nextID = maxID + 1
		table.mu.Unlock()
	}

	return nil
}

// load reads table data from JSON file
func (t *JSONTable) load() {
	if t.db.basePath == "" {
		return
	}

	filePath := filepath.Join(t.db.basePath, t.name+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return // File doesn't exist yet, that's fine
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return
	}

	t.records = records

	// Find max ID
	maxID := 0
	for _, record := range records {
		if id, ok := record["id"].(float64); ok {
			if int(id) > maxID {
				maxID = int(id)
			}
		}
	}
	t.nextID = maxID + 1
}

// save writes table data to JSON file
func (t *JSONTable) save() error {
	if t.db.basePath == "" {
		return nil // In-memory only mode
	}

	// Ensure directory exists
	if err := os.MkdirAll(t.db.basePath, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(t.db.basePath, t.name+".json")
	data, err := json.MarshalIndent(t.records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Find returns records matching the query
// If query matches exactly one record, returns that record directly
// If query matches multiple records, returns an array
// If no matches, returns nil
func (t *JSONTable) Find(query map[string]interface{}) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// If query is empty, return all records
	if len(query) == 0 {
		result := make([]map[string]interface{}, len(t.records))
		copy(result, t.records)
		return result
	}

	var results []map[string]interface{}
	for _, record := range t.records {
		if t.matches(record, query) {
			// Return a copy
			recordCopy := make(map[string]interface{})
			for k, v := range record {
				recordCopy[k] = v
			}
			results = append(results, recordCopy)
		}
	}

	// If exactly one match, return it directly (common case for unique queries)
	if len(results) == 1 {
		return results[0]
	}

	// If no matches, return nil
	if len(results) == 0 {
		return nil
	}

	// Multiple matches - return as array
	return results
}

// FindAll returns all records matching the query as an array (always returns array)
func (t *JSONTable) FindAll(query map[string]interface{}) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// If query is empty, return all records
	if len(query) == 0 {
		result := make([]map[string]interface{}, len(t.records))
		copy(result, t.records)
		return result
	}

	var results []map[string]interface{}
	for _, record := range t.records {
		if t.matches(record, query) {
			// Return a copy
			recordCopy := make(map[string]interface{})
			for k, v := range record {
				recordCopy[k] = v
			}
			results = append(results, recordCopy)
		}
	}

	return results
}

// FindOne returns the first record matching the query
func (t *JSONTable) FindOne(query map[string]interface{}) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, record := range t.records {
		if t.matches(record, query) {
			// Return a copy
			recordCopy := make(map[string]interface{})
			for k, v := range record {
				recordCopy[k] = v
			}
			return recordCopy
		}
	}

	return nil
}

// Create adds a new record and returns it with generated ID
func (t *JSONTable) Create(data map[string]interface{}) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create a copy with auto-generated ID
	record := make(map[string]interface{})
	for k, v := range data {
		record[k] = v
	}

	// Auto-generate ID if not provided
	if _, hasID := record["id"]; !hasID {
		record["id"] = t.nextID
		t.nextID++
	}

	t.records = append(t.records, record)
	t.save()

	return record
}

// Update modifies records matching the query
func (t *JSONTable) Update(query map[string]interface{}, updates map[string]interface{}) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	var updated []map[string]interface{}
	for i, record := range t.records {
		if t.matches(record, query) {
			// Apply updates
			for k, v := range updates {
				t.records[i][k] = v
			}
			updated = append(updated, t.records[i])
		}
	}

	if len(updated) > 0 {
		t.save()
	}

	if len(updated) == 1 {
		return updated[0]
	}
	return updated
}

// Delete removes records matching the query
func (t *JSONTable) Delete(query map[string]interface{}) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	var remaining []map[string]interface{}
	var deleted []map[string]interface{}

	for _, record := range t.records {
		if t.matches(record, query) {
			deleted = append(deleted, record)
		} else {
			remaining = append(remaining, record)
		}
	}

	if len(deleted) > 0 {
		t.records = remaining
		t.save()
	}

	return map[string]interface{}{
		"deleted": len(deleted),
	}
}

// Count returns the number of records matching the query
func (t *JSONTable) Count(query map[string]interface{}) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(query) == 0 {
		return len(t.records)
	}

	count := 0
	for _, record := range t.records {
		if t.matches(record, query) {
			count++
		}
	}
	return count
}

// All returns all records
func (t *JSONTable) All() interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]map[string]interface{}, len(t.records))
	copy(result, t.records)
	return result
}

// matches checks if a record matches a query
func (t *JSONTable) matches(record, query map[string]interface{}) bool {
	for key, queryValue := range query {
		recordValue, exists := record[key]
		if !exists {
			return false
		}

		// Handle numeric comparison (JSON numbers are float64)
		if qv, ok := queryValue.(int); ok {
			if rv, ok := recordValue.(float64); ok {
				if float64(qv) != rv {
					return false
				}
				continue
			}
		}

		// int64 comparison (Glyph uses int64 for integers)
		if qv, ok := queryValue.(int64); ok {
			if rv, ok := recordValue.(float64); ok {
				if float64(qv) != rv {
					return false
				}
				continue
			}
			if rv, ok := recordValue.(int64); ok {
				if qv != rv {
					return false
				}
				continue
			}
		}

		if qv, ok := queryValue.(float64); ok {
			if rv, ok := recordValue.(float64); ok {
				if qv != rv {
					return false
				}
				continue
			}
			if rv, ok := recordValue.(int); ok {
				if qv != float64(rv) {
					return false
				}
				continue
			}
		}

		// String comparison
		if qv, ok := queryValue.(string); ok {
			if rv, ok := recordValue.(string); ok {
				if qv != rv {
					return false
				}
				continue
			}
		}

		// Boolean comparison
		if qv, ok := queryValue.(bool); ok {
			if rv, ok := recordValue.(bool); ok {
				if qv != rv {
					return false
				}
				continue
			}
		}

		// Direct comparison for other types
		if queryValue != recordValue {
			return false
		}
	}
	return true
}
