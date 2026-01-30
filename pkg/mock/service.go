package mock

import (
	"fmt"
	"sync"
)

// Call records a single invocation of a mocked method.
type Call struct {
	Method string
	Args   []interface{}
	Result interface{}
}

// Service is a generic mock service for use with GlyphLang's dependency
// injection system. Methods can be stubbed with return values or dynamic
// functions, and all calls are tracked for assertions.
type Service struct {
	mu     sync.RWMutex
	stubs  map[string]interface{} // method -> return value or func
	calls  map[string][]Call      // method -> call history
	tables map[string]*TableMock  // sub-object mocks (e.g., db.users)
}

// NewService creates a new mock Service.
func NewService() *Service {
	return &Service{
		stubs:  make(map[string]interface{}),
		calls:  make(map[string][]Call),
		tables: make(map[string]*TableMock),
	}
}

// On stubs a method to return a static value.
func (s *Service) On(method string, returnValue interface{}) *Service {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stubs[method] = returnValue
	return s
}

// OnFunc stubs a method with a dynamic function.
// The function receives the arguments and returns a result.
func (s *Service) OnFunc(method string, fn func(args ...interface{}) interface{}) *Service {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stubs[method] = fn
	return s
}

// Call invokes a stubbed method, recording the call and returning the result.
// This method is exported so it can be called via reflection from the interpreter.
func (s *Service) CallMethod(method string, args ...interface{}) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stub, ok := s.stubs[method]
	if !ok {
		call := Call{Method: method, Args: args}
		s.calls[method] = append(s.calls[method], call)
		return nil, fmt.Errorf("mock: no stub defined for method %q", method)
	}

	var result interface{}
	if fn, ok := stub.(func(args ...interface{}) interface{}); ok {
		result = fn(args...)
	} else {
		result = stub
	}

	call := Call{Method: method, Args: args, Result: result}
	s.calls[method] = append(s.calls[method], call)
	return result, nil
}

// Table returns a sub-object mock (like db.users).
func (s *Service) Table(name string) *TableMock {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.tables[name]; ok {
		return t
	}
	t := &TableMock{
		service: s,
		name:    name,
		stubs:   make(map[string]interface{}),
	}
	s.tables[name] = t
	return t
}

// --- Assertion helpers ---

// Called returns true if the method was called at least once.
func (s *Service) Called(method string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.calls[method]) > 0
}

// CalledOnce returns true if the method was called exactly once.
func (s *Service) CalledOnce(method string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.calls[method]) == 1
}

// CalledTimes returns the number of times the method was called.
func (s *Service) CalledTimes(method string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.calls[method])
}

// Calls returns all recorded calls for a method.
func (s *Service) Calls(method string) []Call {
	s.mu.RLock()
	defer s.mu.RUnlock()
	calls := s.calls[method]
	result := make([]Call, len(calls))
	copy(result, calls)
	return result
}

// LastCall returns the most recent call for a method, or nil if none.
func (s *Service) LastCall(method string) *Call {
	s.mu.RLock()
	defer s.mu.RUnlock()
	calls := s.calls[method]
	if len(calls) == 0 {
		return nil
	}
	c := calls[len(calls)-1]
	return &c
}

// Reset clears all call history. Stubs remain.
func (s *Service) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = make(map[string][]Call)
	for _, t := range s.tables {
		t.Reset()
	}
}

// ResetAll clears both call history and stubs.
func (s *Service) ResetAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stubs = make(map[string]interface{})
	s.calls = make(map[string][]Call)
	s.tables = make(map[string]*TableMock)
}

// TableMock is a sub-object mock that records calls via its parent Service.
type TableMock struct {
	service *Service
	name    string
	mu      sync.RWMutex
	stubs   map[string]interface{}
	calls   []Call
}

// On stubs a method on the table mock.
func (t *TableMock) On(method string, returnValue interface{}) *TableMock {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stubs[method] = returnValue
	return t
}

// Get is a common method that returns a stubbed value by key.
func (t *TableMock) Get(id interface{}) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, Call{Method: "Get", Args: []interface{}{id}})
	if stub, ok := t.stubs["Get"]; ok {
		return stub
	}
	return nil
}

// All returns a stubbed list of all records.
func (t *TableMock) All() interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, Call{Method: "All"})
	if stub, ok := t.stubs["All"]; ok {
		return stub
	}
	return []interface{}{}
}

// Create records a create call and returns the stubbed result.
func (t *TableMock) Create(data map[string]interface{}) interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, Call{Method: "Create", Args: []interface{}{data}})
	if stub, ok := t.stubs["Create"]; ok {
		return stub
	}
	return data
}

// Delete records a delete call.
func (t *TableMock) Delete(id interface{}) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, Call{Method: "Delete", Args: []interface{}{id}})
	if stub, ok := t.stubs["Delete"]; ok {
		if b, ok := stub.(bool); ok {
			return b
		}
	}
	return true
}

// Called returns true if any method on this table was called.
func (t *TableMock) Called(method string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, c := range t.calls {
		if c.Method == method {
			return true
		}
	}
	return false
}

// CalledTimes returns how many times a method was called on this table.
func (t *TableMock) CalledTimes(method string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	count := 0
	for _, c := range t.calls {
		if c.Method == method {
			count++
		}
	}
	return count
}

// Reset clears call history on this table. Stubs remain.
func (t *TableMock) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = nil
}
