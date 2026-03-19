// Package stdlib provides Database ORM support for FV 2.0
package stdlib

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents a database connection pool
type Database struct {
	DB                 *sql.DB
	ConnectionString   string
	MaxConnections     int
	Timeout            int
	LastInsertID       int64
	RowsAffected       int64
	QueryTimeout       int
}

// Row represents a database row result
type Row struct {
	Values map[string]interface{}
}

// Result represents a query result
type Result struct {
	Rows           []*Row
	LastInsertID   int64
	RowsAffected   int64
	Error          error
}

// Query represents a database query builder
type Query struct {
	db              *Database
	selectCols      []string
	fromTable       string
	whereClauses    []string
	orderByClause   string
	limitClause     string
	offsetClause    string
	joinClauses     []string
	groupByClause   string
	havingClause    string
	distinct        bool
	params          []interface{}
}

// Transaction represents a database transaction
type Transaction struct {
	tx *sql.Tx
	db *Database
}

// NewDatabase creates a new database connection
func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{
		DB:               db,
		ConnectionString: connectionString,
		MaxConnections:   10,
		Timeout:          5000,
		QueryTimeout:     30,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Exec executes a statement without returning rows
func (d *Database) Exec(query string, args ...interface{}) (*Result, error) {
	result, err := d.DB.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	lastID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return &Result{
		LastInsertID: lastID,
		RowsAffected: rowsAffected,
		Error:        nil,
	}, nil
}

// Query executes a query that returns rows
func (d *Database) Query(query string, args ...interface{}) (*Result, error) {
	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &Result{
		Rows: make([]*Row, 0),
	}

	columns, _ := rows.Columns()
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := &Row{
			Values: make(map[string]interface{}),
		}

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row.Values[col] = string(b)
			} else {
				row.Values[col] = val
			}
		}

		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) (*Row, error) {
	_ = d.DB.QueryRow(query, args...)

	// This is a simplified version - real implementation would need column info
	return &Row{
		Values: make(map[string]interface{}),
	}, nil
}

// Begin starts a transaction
func (d *Database) Begin() (*Transaction, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}

	return &Transaction{
		tx: tx,
		db: d,
	}, nil
}

// Commit commits a transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back a transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Exec executes a statement within a transaction
func (t *Transaction) Exec(query string, args ...interface{}) (*Result, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	lastID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return &Result{
		LastInsertID: lastID,
		RowsAffected: rowsAffected,
	}, nil
}

// Query executes a query within a transaction
func (t *Transaction) Query(query string, args ...interface{}) (*Result, error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &Result{
		Rows: make([]*Row, 0),
	}

	columns, _ := rows.Columns()
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := &Row{
			Values: make(map[string]interface{}),
		}

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row.Values[col] = string(b)
			} else {
				row.Values[col] = val
			}
		}

		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

// NewQuery creates a new query builder
func (d *Database) NewQuery() *Query {
	return &Query{
		db:          d,
		selectCols:  []string{},
		whereClauses: []string{},
		params:      []interface{}{},
	}
}

// Select specifies columns to select
func (q *Query) Select(cols ...string) *Query {
	q.selectCols = cols
	return q
}

// From specifies the table
func (q *Query) From(table string) *Query {
	q.fromTable = table
	return q
}

// Where adds a where clause
func (q *Query) Where(condition string, args ...interface{}) *Query {
	q.whereClauses = append(q.whereClauses, condition)
	q.params = append(q.params, args...)
	return q
}

// OrderBy specifies ordering
func (q *Query) OrderBy(column string, direction string) *Query {
	q.orderByClause = fmt.Sprintf("ORDER BY %s %s", column, direction)
	return q
}

// Limit specifies result limit
func (q *Query) Limit(limit int) *Query {
	q.limitClause = fmt.Sprintf("LIMIT %d", limit)
	return q
}

// Offset specifies result offset
func (q *Query) Offset(offset int) *Query {
	q.offsetClause = fmt.Sprintf("OFFSET %d", offset)
	return q
}

// Join adds a join clause
func (q *Query) Join(table string, condition string) *Query {
	q.joinClauses = append(q.joinClauses, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return q
}

// LeftJoin adds a left join clause
func (q *Query) LeftJoin(table string, condition string) *Query {
	q.joinClauses = append(q.joinClauses, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return q
}

// GroupBy specifies grouping
func (q *Query) GroupBy(column string) *Query {
	q.groupByClause = fmt.Sprintf("GROUP BY %s", column)
	return q
}

// Having specifies having clause
func (q *Query) Having(condition string) *Query {
	q.havingClause = fmt.Sprintf("HAVING %s", condition)
	return q
}

// Distinct adds distinct modifier
func (q *Query) Distinct() *Query {
	q.distinct = true
	return q
}

// Build builds the query string
func (q *Query) Build() string {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	if q.distinct {
		query.WriteString("DISTINCT ")
	}

	if len(q.selectCols) == 0 {
		query.WriteString("*")
	} else {
		query.WriteString(strings.Join(q.selectCols, ", "))
	}

	// FROM clause
	if q.fromTable != "" {
		query.WriteString(" FROM ")
		query.WriteString(q.fromTable)
	}

	// JOIN clauses
	for _, join := range q.joinClauses {
		query.WriteString(" ")
		query.WriteString(join)
	}

	// WHERE clauses
	if len(q.whereClauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(q.whereClauses, " AND "))
	}

	// GROUP BY clause
	if q.groupByClause != "" {
		query.WriteString(" ")
		query.WriteString(q.groupByClause)
	}

	// HAVING clause
	if q.havingClause != "" {
		query.WriteString(" ")
		query.WriteString(q.havingClause)
	}

	// ORDER BY clause
	if q.orderByClause != "" {
		query.WriteString(" ")
		query.WriteString(q.orderByClause)
	}

	// LIMIT clause
	if q.limitClause != "" {
		query.WriteString(" ")
		query.WriteString(q.limitClause)
	}

	// OFFSET clause
	if q.offsetClause != "" {
		query.WriteString(" ")
		query.WriteString(q.offsetClause)
	}

	return query.String()
}

// Execute executes the built query
func (q *Query) Execute() (*Result, error) {
	queryStr := q.Build()
	return q.db.Query(queryStr, q.params...)
}

// First returns the first row
func (q *Query) First() (*Row, error) {
	result, err := q.Execute()
	if err != nil {
		return nil, err
	}

	if len(result.Rows) == 0 {
		return nil, fmt.Errorf("no rows found")
	}

	return result.Rows[0], nil
}

// All returns all rows
func (q *Query) All() ([]*Row, error) {
	result, err := q.Execute()
	if err != nil {
		return nil, err
	}

	return result.Rows, nil
}

// Count counts rows
func (q *Query) Count() (int64, error) {
	countQuery := q.db.NewQuery()
	countQuery.selectCols = []string{"COUNT(*) as count"}
	countQuery.fromTable = q.fromTable
	countQuery.whereClauses = q.whereClauses
	countQuery.params = q.params

	result, err := countQuery.Execute()
	if err != nil {
		return 0, err
	}

	if len(result.Rows) == 0 {
		return 0, nil
	}

	count := result.Rows[0].Values["count"]
	if count == nil {
		return 0, nil
	}

	return count.(int64), nil
}

// InsertOne inserts a single row
func (d *Database) InsertOne(table string, data map[string]interface{}) (*Result, error) {
	var cols []string
	var vals []string
	var args []interface{}

	for k, v := range data {
		cols = append(cols, k)
		vals = append(vals, "?")
		args = append(args, v)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(cols, ", "), strings.Join(vals, ", "))

	return d.Exec(query, args...)
}

// UpdateOne updates a single row
func (d *Database) UpdateOne(table string, id int64, data map[string]interface{}) (*Result, error) {
	var sets []string
	var args []interface{}

	for k, v := range data {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		table, strings.Join(sets, ", "))

	return d.Exec(query, args...)
}

// DeleteOne deletes a single row
func (d *Database) DeleteOne(table string, id int64) (*Result, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	return d.Exec(query, id)
}

// CreateTable creates a table
func (d *Database) CreateTable(table string, schema map[string]string) (*Result, error) {
	var cols []string
	for col, typ := range schema {
		cols = append(cols, fmt.Sprintf("%s %s", col, typ))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
		table, strings.Join(cols, ", "))

	return d.Exec(query)
}

// DropTable drops a table
func (d *Database) DropTable(table string) (*Result, error) {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
	return d.Exec(query)
}

// Migrate runs migrations
func (d *Database) Migrate(migrations []string) error {
	for _, migration := range migrations {
		if _, err := d.Exec(migration); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves a value from a row
func (r *Row) Get(key string) interface{} {
	return r.Values[key]
}

// GetString retrieves a string value
func (r *Row) GetString(key string) string {
	val := r.Values[key]
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// GetInt retrieves an int value
func (r *Row) GetInt(key string) int {
	val := r.Values[key]
	if i, ok := val.(int64); ok {
		return int(i)
	}
	return 0
}

// GetInt64 retrieves an int64 value
func (r *Row) GetInt64(key string) int64 {
	val := r.Values[key]
	if i, ok := val.(int64); ok {
		return i
	}
	return 0
}

// GetBool retrieves a bool value
func (r *Row) GetBool(key string) bool {
	val := r.Values[key]
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}
