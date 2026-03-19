package stdlib

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDatabaseConnection tests creating a database connection
func TestDatabaseConnection(t *testing.T) {
	// Clean up test database
	os.Remove("test.db")

	db, err := NewDatabase("test.db")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "test.db", db.ConnectionString)
	assert.Equal(t, 10, db.MaxConnections)

	db.Close()
	os.Remove("test.db")
}

// TestCreateTable tests creating a table
func TestCreateTable(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	schema := map[string]string{
		"id":   "INTEGER PRIMARY KEY",
		"name": "TEXT NOT NULL",
		"age":  "INTEGER",
	}

	result, err := db.CreateTable("users", schema)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	db.Close()
	os.Remove("test.db")
}

// TestInsertOne tests inserting a single row
func TestInsertOne(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	db.CreateTable("users", map[string]string{
		"id":   "INTEGER PRIMARY KEY AUTOINCREMENT",
		"name": "TEXT NOT NULL",
		"email": "TEXT",
	})

	data := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	}

	result, err := db.InsertOne("users", data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.LastInsertID, int64(0))

	db.Close()
	os.Remove("test.db")
}

// TestQueryBuilder tests the query builder
func TestQueryBuilder(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Select("id", "name", "email").From("users").Where("age > ?", 18).OrderBy("name", "ASC").Limit(10)

	queryStr := query.Build()
	assert.Contains(t, queryStr, "SELECT id, name, email")
	assert.Contains(t, queryStr, "FROM users")
	assert.Contains(t, queryStr, "WHERE age > ?")
	assert.Contains(t, queryStr, "ORDER BY name ASC")
	assert.Contains(t, queryStr, "LIMIT 10")

	db.Close()
	os.Remove("test.db")
}

// TestQueryBuilderWithJoin tests query builder with JOIN
func TestQueryBuilderWithJoin(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Select("u.id", "u.name", "p.title").
		From("users u").
		Join("posts p", "u.id = p.user_id").
		Where("u.id = ?", 1)

	queryStr := query.Build()
	assert.Contains(t, queryStr, "JOIN posts p ON u.id = p.user_id")

	db.Close()
	os.Remove("test.db")
}

// TestQueryBuilderWithDistinct tests DISTINCT
func TestQueryBuilderWithDistinct(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Distinct().Select("city").From("users")

	queryStr := query.Build()
	assert.Contains(t, queryStr, "SELECT DISTINCT city")

	db.Close()
	os.Remove("test.db")
}

// TestQueryBuilderWithGroupBy tests GROUP BY
func TestQueryBuilderWithGroupBy(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Select("age", "COUNT(*) as count").
		From("users").
		GroupBy("age")

	queryStr := query.Build()
	assert.Contains(t, queryStr, "GROUP BY age")

	db.Close()
	os.Remove("test.db")
}

// TestUpdateOne tests updating a row
func TestUpdateOne(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	db.CreateTable("users", map[string]string{
		"id":   "INTEGER PRIMARY KEY",
		"name": "TEXT",
		"email": "TEXT",
	})

	data := map[string]interface{}{
		"name":  "Bob",
		"email": "bob@example.com",
	}

	result, err := db.UpdateOne("users", 1, data)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	db.Close()
	os.Remove("test.db")
}

// TestDeleteOne tests deleting a row
func TestDeleteOne(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	db.CreateTable("users", map[string]string{
		"id":   "INTEGER PRIMARY KEY",
		"name": "TEXT",
	})

	result, err := db.DeleteOne("users", 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	db.Close()
	os.Remove("test.db")
}

// TestTransaction tests database transactions
func TestTransaction(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	tx, err := db.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	err = tx.Commit()
	assert.NoError(t, err)

	db.Close()
	os.Remove("test.db")
}

// TestTransactionRollback tests transaction rollback
func TestTransactionRollback(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	tx, _ := db.Begin()
	err := tx.Rollback()
	assert.NoError(t, err)

	db.Close()
	os.Remove("test.db")
}

// TestRow tests row value retrieval
func TestRowGetters(t *testing.T) {
	row := &Row{
		Values: map[string]interface{}{
			"id":    int64(1),
			"name":  "Alice",
			"age":   int64(30),
			"active": true,
		},
	}

	assert.Equal(t, int64(1), row.GetInt64("id"))
	assert.Equal(t, "Alice", row.GetString("name"))
	assert.Equal(t, 30, row.GetInt("age"))
	assert.Equal(t, true, row.GetBool("active"))
}

// TestDropTable tests dropping a table
func TestDropTable(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	db.CreateTable("users", map[string]string{
		"id":   "INTEGER PRIMARY KEY",
		"name": "TEXT",
	})

	result, err := db.DropTable("users")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	db.Close()
	os.Remove("test.db")
}

// TestQueryWithOffset tests OFFSET
func TestQueryWithOffset(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Select("*").From("users").Limit(10).Offset(20)

	queryStr := query.Build()
	assert.Contains(t, queryStr, "LIMIT 10")
	assert.Contains(t, queryStr, "OFFSET 20")

	db.Close()
	os.Remove("test.db")
}

// TestQueryWithLeftJoin tests LEFT JOIN
func TestQueryWithLeftJoin(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Select("u.id", "u.name", "p.title").
		From("users u").
		LeftJoin("posts p", "u.id = p.user_id")

	queryStr := query.Build()
	assert.Contains(t, queryStr, "LEFT JOIN posts p")

	db.Close()
	os.Remove("test.db")
}

// TestMultipleWhere tests multiple WHERE clauses
func TestMultipleWhere(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	query := db.NewQuery()
	query.Select("*").
		From("users").
		Where("age > ?", 18).
		Where("status = ?", "active")

	queryStr := query.Build()
	assert.Contains(t, queryStr, "WHERE age > ? AND status = ?")

	db.Close()
	os.Remove("test.db")
}

// TestExec tests direct SQL execution
func TestExec(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	result, err := db.Exec("CREATE TABLE test (id INTEGER, name TEXT)")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	db.Close()
	os.Remove("test.db")
}

// TestMigrate tests running migrations
func TestMigrate(t *testing.T) {
	os.Remove("test.db")
	db, _ := NewDatabase("test.db")

	migrations := []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
		"CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT)",
	}

	err := db.Migrate(migrations)
	assert.NoError(t, err)

	db.Close()
	os.Remove("test.db")
}
