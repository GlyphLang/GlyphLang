package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ORM provides a simple Object-Relational Mapping layer
type ORM struct {
	db    Database
	table string
}

// NewORM creates a new ORM instance for a table
func NewORM(db Database, table string) *ORM {
	return &ORM{
		db:    db,
		table: table,
	}
}

// QueryBuilder provides a fluent interface for building SQL queries
type QueryBuilder struct {
	orm        *ORM
	selectCols []string
	whereConds []WhereCondition
	orderBy    string
	limit      int
	offset     int
	offsetSet  bool
	joins      []Join
}

// WhereCondition represents a WHERE clause condition
type WhereCondition struct {
	Column   string
	Operator string
	Value    interface{}
}

// Join represents a JOIN clause
type Join struct {
	Type       string // INNER, LEFT, RIGHT
	Table      string
	Condition  string
	OnColumn   string
	WithColumn string
}

// NewQueryBuilder creates a new query builder
func (o *ORM) NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		orm:        o,
		selectCols: []string{"*"},
		whereConds: []WhereCondition{},
		joins:      []Join{},
	}
}

// Select specifies the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = columns
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(column string, operator string, value interface{}) *QueryBuilder {
	qb.whereConds = append(qb.whereConds, WhereCondition{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// WhereEq is a shorthand for Where with "=" operator
func (qb *QueryBuilder) WhereEq(column string, value interface{}) *QueryBuilder {
	return qb.Where(column, "=", value)
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	qb.orderBy = fmt.Sprintf("%s %s", column, direction)
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	qb.offsetSet = true
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(joinType string, table string, onColumn string, withColumn string) *QueryBuilder {
	qb.joins = append(qb.joins, Join{
		Type:       joinType,
		Table:      table,
		OnColumn:   onColumn,
		WithColumn: withColumn,
	})
	return qb
}

// InnerJoin adds an INNER JOIN clause
func (qb *QueryBuilder) InnerJoin(table string, onColumn string, withColumn string) *QueryBuilder {
	return qb.Join("INNER", table, onColumn, withColumn)
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table string, onColumn string, withColumn string) *QueryBuilder {
	return qb.Join("LEFT", table, onColumn, withColumn)
}

// Build constructs the SQL query and arguments
// Returns an error if any identifier is invalid
func (qb *QueryBuilder) Build() (string, []interface{}, error) {
	// Sanitize table name
	sanitizedTable, err := SanitizeIdentifier(qb.orm.table)
	if err != nil {
		return "", nil, fmt.Errorf("invalid table name: %w", err)
	}

	// Sanitize select columns
	sanitizedSelectCols := make([]string, len(qb.selectCols))
	for i, col := range qb.selectCols {
		if col == "*" {
			sanitizedSelectCols[i] = col
		} else {
			sanitized, err := SanitizeIdentifier(col)
			if err != nil {
				return "", nil, fmt.Errorf("invalid select column %q: %w", col, err)
			}
			sanitizedSelectCols[i] = sanitized
		}
	}

	// Build SELECT clause
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(sanitizedSelectCols, ", "), sanitizedTable)

	// Build JOIN clauses
	for _, join := range qb.joins {
		// Validate join type
		joinType := strings.ToUpper(join.Type)
		if joinType != "INNER" && joinType != "LEFT" && joinType != "RIGHT" && joinType != "FULL" {
			return "", nil, fmt.Errorf("invalid join type: %s", join.Type)
		}

		sanitizedJoinTable, err := SanitizeIdentifier(join.Table)
		if err != nil {
			return "", nil, fmt.Errorf("invalid join table %q: %w", join.Table, err)
		}
		sanitizedOnColumn, err := SanitizeIdentifier(join.OnColumn)
		if err != nil {
			return "", nil, fmt.Errorf("invalid join on column %q: %w", join.OnColumn, err)
		}
		sanitizedWithColumn, err := SanitizeIdentifier(join.WithColumn)
		if err != nil {
			return "", nil, fmt.Errorf("invalid join with column %q: %w", join.WithColumn, err)
		}

		query += fmt.Sprintf(" %s JOIN %s ON %s.%s = %s.%s",
			joinType, sanitizedJoinTable, sanitizedTable, sanitizedOnColumn, sanitizedJoinTable, sanitizedWithColumn)
	}

	// Build WHERE clause
	var args []interface{}
	if len(qb.whereConds) > 0 {
		query += " WHERE "
		for i, cond := range qb.whereConds {
			if i > 0 {
				query += " AND "
			}
			sanitizedColumn, err := SanitizeIdentifier(cond.Column)
			if err != nil {
				return "", nil, fmt.Errorf("invalid where column %q: %w", cond.Column, err)
			}
			// Validate operator (only allow safe operators)
			operator := strings.ToUpper(strings.TrimSpace(cond.Operator))
			validOperators := map[string]bool{
				"=": true, "!=": true, "<>": true, "<": true, ">": true,
				"<=": true, ">=": true, "LIKE": true, "ILIKE": true,
				"IN": true, "NOT IN": true, "IS": true, "IS NOT": true,
			}
			if !validOperators[operator] {
				return "", nil, fmt.Errorf("invalid operator: %s", cond.Operator)
			}
			query += fmt.Sprintf("%s %s $%d", sanitizedColumn, operator, i+1)
			args = append(args, cond.Value)
		}
	}

	// Build ORDER BY clause
	if qb.orderBy != "" {
		// Parse and sanitize orderBy (format: "column direction")
		parts := strings.Fields(qb.orderBy)
		if len(parts) >= 1 {
			sanitizedOrderCol, err := SanitizeIdentifier(parts[0])
			if err != nil {
				return "", nil, fmt.Errorf("invalid order by column: %w", err)
			}
			direction := "ASC"
			if len(parts) >= 2 {
				dir := strings.ToUpper(parts[1])
				if dir == "ASC" || dir == "DESC" {
					direction = dir
				} else {
					return "", nil, fmt.Errorf("invalid order direction: %s", parts[1])
				}
			}
			query += fmt.Sprintf(" ORDER BY %s %s", sanitizedOrderCol, direction)
		}
	}

	// Build LIMIT clause
	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	// Build OFFSET clause
	if qb.offsetSet {
		query += fmt.Sprintf(" OFFSET %d", qb.offset)
	}

	return query, args, nil
}

// Get executes the query and returns results
func (qb *QueryBuilder) Get(ctx context.Context) ([]map[string]interface{}, error) {
	query, args, err := qb.Build()
	if err != nil {
		return nil, err
	}
	return qb.orm.Query(ctx, query, args...)
}

// First executes the query and returns the first result
func (qb *QueryBuilder) First(ctx context.Context) (map[string]interface{}, error) {
	qb.Limit(1)
	results, err := qb.Get(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return results[0], nil
}

// FindByID finds a record by ID
func (o *ORM) FindByID(ctx context.Context, id interface{}) (map[string]interface{}, error) {
	return o.NewQueryBuilder().WhereEq("id", id).First(ctx)
}

// FindAll returns all records from the table
func (o *ORM) FindAll(ctx context.Context) ([]map[string]interface{}, error) {
	return o.NewQueryBuilder().Get(ctx)
}

// Query executes a raw SQL query
func (o *ORM) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := o.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows)
}

// Create inserts a new record
func (o *ORM) Create(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to insert")
	}

	// Sanitize table name
	sanitizedTable, err := SanitizeIdentifier(o.table)
	if err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	columns := make([]string, 0, len(data))
	sanitizedColumns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	i := 1
	for col, val := range data {
		// Sanitize column name
		sanitizedCol, err := SanitizeIdentifier(col)
		if err != nil {
			return nil, fmt.Errorf("invalid column name %q: %w", col, err)
		}
		columns = append(columns, col)
		sanitizedColumns = append(sanitizedColumns, sanitizedCol)
		values = append(values, val)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		sanitizedTable,
		strings.Join(sanitizedColumns, ", "),
		strings.Join(placeholders, ", "))

	row := o.db.QueryRow(ctx, query, values...)

	// Get column names for scanning
	result, err := scanRow(row, columns)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update updates a record by ID
func (o *ORM) Update(ctx context.Context, id interface{}, data map[string]interface{}) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to update")
	}

	// Sanitize table name
	sanitizedTable, err := SanitizeIdentifier(o.table)
	if err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)
	columns := make([]string, 0, len(data))

	i := 1
	for col, val := range data {
		// Sanitize column name
		sanitizedCol, err := SanitizeIdentifier(col)
		if err != nil {
			return nil, fmt.Errorf("invalid column name %q: %w", col, err)
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", sanitizedCol, i))
		values = append(values, val)
		columns = append(columns, col)
		i++
	}

	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE \"id\" = $%d RETURNING *",
		sanitizedTable,
		strings.Join(setClauses, ", "),
		i)

	row := o.db.QueryRow(ctx, query, values...)

	// Add id to columns for scanning
	columns = append(columns, "id")

	result, err := scanRow(row, columns)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete deletes a record by ID
func (o *ORM) Delete(ctx context.Context, id interface{}) error {
	// Sanitize table name
	sanitizedTable, err := SanitizeIdentifier(o.table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE \"id\" = $1", sanitizedTable)
	result, err := o.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Count counts records matching the WHERE conditions
func (o *ORM) Count(ctx context.Context, whereConds ...WhereCondition) (int64, error) {
	// Sanitize table name
	sanitizedTable, err := SanitizeIdentifier(o.table)
	if err != nil {
		return 0, fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", sanitizedTable)

	var args []interface{}
	if len(whereConds) > 0 {
		query += " WHERE "
		for i, cond := range whereConds {
			if i > 0 {
				query += " AND "
			}
			// Sanitize column name
			sanitizedColumn, err := SanitizeIdentifier(cond.Column)
			if err != nil {
				return 0, fmt.Errorf("invalid where column %q: %w", cond.Column, err)
			}
			// Validate operator (only allow safe operators)
			operator := strings.ToUpper(strings.TrimSpace(cond.Operator))
			validOperators := map[string]bool{
				"=": true, "!=": true, "<>": true, "<": true, ">": true,
				"<=": true, ">=": true, "LIKE": true, "ILIKE": true,
				"IN": true, "NOT IN": true, "IS": true, "IS NOT": true,
			}
			if !validOperators[operator] {
				return 0, fmt.Errorf("invalid operator: %s", cond.Operator)
			}
			query += fmt.Sprintf("%s %s $%d", sanitizedColumn, operator, i+1)
			args = append(args, cond.Value)
		}
	}

	var count int64
	err = o.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Exists checks if a record exists
func (o *ORM) Exists(ctx context.Context, whereConds ...WhereCondition) (bool, error) {
	count, err := o.Count(ctx, whereConds...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// StructToMap converts a struct to a map for database operations
func StructToMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get the field name from the struct tag or use the field name
		name := field.Tag.Get("db")
		if name == "" {
			name = strings.ToLower(field.Name)
		}

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		// Handle different types
		switch value.Kind() {
		case reflect.Ptr:
			if !value.IsNil() {
				result[name] = value.Elem().Interface()
			}
		default:
			result[name] = value.Interface()
		}
	}

	return result, nil
}

// MapToStruct converts a map to a struct
func MapToStruct(data map[string]interface{}, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get the field name from the struct tag or use the field name
		name := field.Tag.Get("db")
		if name == "" {
			name = strings.ToLower(field.Name)
		}

		// Get the value from the map
		if value, ok := data[name]; ok {
			setValue(fieldValue, value)
		}
	}

	return nil
}

// setValue sets a field value handling type conversion
func setValue(field reflect.Value, value interface{}) {
	if value == nil {
		return
	}

	val := reflect.ValueOf(value)

	// Handle type conversion
	if field.Type() == val.Type() {
		field.Set(val)
		return
	}

	// Handle common conversions
	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		switch v := value.(type) {
		case int64:
			field.SetInt(v)
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetInt(int64(v))
		}
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		}
	case reflect.Float64:
		if f, ok := value.(float64); ok {
			field.SetFloat(f)
		}
	}
}

// scanRows scans multiple rows into a slice of maps
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert byte arrays to strings
			if b, ok := val.([]byte); ok {
				val = string(b)
			}

			row[col] = val
		}

		results = append(results, row)
	}

	return results, rows.Err()
}

// scanRow scans a single row into a map
func scanRow(row *sql.Row, columns []string) (map[string]interface{}, error) {
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	if err := row.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]

		// Convert byte arrays to strings
		if b, ok := val.([]byte); ok {
			val = string(b)
		}

		result[col] = val
	}

	return result, nil
}

// Transaction executes a function within a transaction
func (o *ORM) Transaction(ctx context.Context, fn func(context.Context) error) error {
	// For PostgreSQL, we can cast to *PostgresDB
	if pgDB, ok := o.db.(*PostgresDB); ok {
		return pgDB.Transaction(ctx, func(tx *sql.Tx) error {
			// Create a new ORM instance with the transaction
			// This is a simplified version - in production, you'd want to properly wrap the tx
			return fn(ctx)
		})
	}

	return fmt.Errorf("transaction not supported for this database driver")
}

// Timestamp returns current timestamp
func Timestamp() int64 {
	return time.Now().Unix()
}
