package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// setupInMemoryDB creates a SQLite in-memory database for testing
func setupInMemoryDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to in-memory database: %v", err)
	}

	// Create test schema
	schema := `
	CREATE TABLE customer_profile (
		id TEXT PRIMARY KEY,
		title TEXT,
		first_name TEXT,
		last_name TEXT,
		email TEXT UNIQUE,
		date_of_birth TEXT,
		created_at DATETIME,
		updated_at DATETIME
	);`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// getMeterForTesting returns a no-op meter for testing
func getMeterForTesting(t *testing.T) metric.Meter {
	meterProvider := noop.NewMeterProvider()
	return meterProvider.Meter("test")
}

// TestUnitOfWork_Begin tests transaction begin
func TestUnitOfWork_Begin(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Test successful begin
	err = uow.Begin(context.Background())
	if err != nil {
		t.Errorf("Begin() failed: %v", err)
	}

	if uow.tx == nil {
		t.Errorf("Begin() should set tx, but got nil")
	}

	// Test double begin should fail
	err = uow.Begin(context.Background())
	if err == nil {
		t.Errorf("Begin() with existing transaction should fail")
	}

	if err.Error() != "transaction already exists" {
		t.Errorf("expected error 'transaction already exists', got: %v", err)
	}

	// Cleanup
	uow.Rollback(context.Background())
}

// TestUnitOfWork_Commit tests transaction commit
func TestUnitOfWork_Commit(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Test commit without begin should fail
	err = uow.Commit(context.Background())
	if err == nil {
		t.Errorf("Commit() without transaction should fail")
	}

	if err.Error() != "no active transaction" {
		t.Errorf("expected error 'no active transaction', got: %v", err)
	}

	// Test successful commit
	err = uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() failed: %v", err)
	}

	err = uow.Commit(context.Background())
	if err != nil {
		t.Errorf("Commit() failed: %v", err)
	}

	if uow.tx != nil {
		t.Errorf("Commit() should set tx to nil, but got: %v", uow.tx)
	}
}

// TestUnitOfWork_Rollback tests transaction rollback
func TestUnitOfWork_Rollback(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Test rollback without transaction should not error
	err = uow.Rollback(context.Background())
	if err != nil {
		t.Errorf("Rollback() without transaction should not error, got: %v", err)
	}

	// Test successful rollback
	err = uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() failed: %v", err)
	}

	err = uow.Rollback(context.Background())
	if err != nil {
		t.Errorf("Rollback() failed: %v", err)
	}

	if uow.tx != nil {
		t.Errorf("Rollback() should set tx to nil, but got: %v", uow.tx)
	}
}

// TestUnitOfWork_WithTx_Success tests WithTx with successful operation
func TestUnitOfWork_WithTx_Success(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	executed := false
	err = uow.WithTx(context.Background(), func(ctx context.Context) error {
		executed = true

		// Insert data in transaction
		_, err := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"test-id", "test@example.com", "John", "Doe", "Mr", "1990-01-01", time.Now(), time.Now())
		return err
	})

	if err != nil {
		t.Errorf("WithTx() failed: %v", err)
	}

	if !executed {
		t.Errorf("WithTx() function should be executed")
	}

	if uow.tx != nil {
		t.Errorf("WithTx() should set tx to nil after commit, but got: %v", uow.tx)
	}

	// Verify data was committed
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "test@example.com")
	if err != nil {
		t.Errorf("failed to query inserted data: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

// TestUnitOfWork_WithTx_Rollback tests WithTx with rollback on error
func TestUnitOfWork_WithTx_Rollback(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	err = uow.WithTx(context.Background(), func(ctx context.Context) error {
		// Insert data that should be rolled back
		_, execErr := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"rollback-id", "rollback@example.com", "Jane", "Doe", "Ms", "1985-05-15", time.Now(), time.Now())
		if execErr != nil {
			return execErr
		}

		return context.DeadlineExceeded // Simulate error
	})

	if err == nil {
		t.Errorf("WithTx() should return error")
	}

	if uow.tx != nil {
		t.Errorf("WithTx() should set tx to nil after rollback, but got: %v", uow.tx)
	}

	// Verify data was rolled back
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "rollback@example.com")
	if err != nil {
		t.Errorf("failed to query data: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 rows (rolled back), got %d", count)
	}
}

// TestUnitOfWork_WithTx_Panic tests WithTx with panic recovery
func TestUnitOfWork_WithTx_Panic(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Use defer to catch the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("WithTx() should re-raise panic")
		}
		// Panic should be re-raised, so we expect to catch it here
	}()

	uow.WithTx(context.Background(), func(ctx context.Context) error {
		// Insert data
		uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"panic-id", "panic@example.com", "Panic", "User", "Dr", "1970-12-31", time.Now(), time.Now())

		panic("test panic")
	})
}

// TestUnitOfWork_WithTx_Panic_Rollback tests that panic triggers rollback
func TestUnitOfWork_WithTx_Panic_Rollback(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Recover from panic
	defer func() {
		recover() // Suppress panic
	}()

	uow.WithTx(context.Background(), func(ctx context.Context) error {
		// Insert data
		uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"panic-id", "panic@example.com", "Panic", "User", "Dr", "1970-12-31", time.Now(), time.Now())

		panic("test panic")
	})

	// Verify data was rolled back
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "panic@example.com")
	if err != nil {
		t.Errorf("failed to query data: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 rows (rolled back after panic), got %d", count)
	}
}

// TestUnitOfWork_MultipleOperations tests multiple operations in single transaction
func TestUnitOfWork_MultipleOperations(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	err = uow.WithTx(context.Background(), func(ctx context.Context) error {
		// Multiple insert operations
		queries := []struct {
			id    string
			email string
			name  string
		}{
			{"id1", "user1@test.com", "User One"},
			{"id2", "user2@test.com", "User Two"},
			{"id3", "user3@test.com", "User Three"},
		}

		for _, q := range queries {
			_, err := uow.tx.ExecContext(ctx,
				"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				q.id, q.email, q.name, "Test", "Mr", "1990-01-01", time.Now(), time.Now())
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		t.Errorf("WithTx() failed: %v", err)
	}

	// Verify all operations were committed
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile")
	if err != nil {
		t.Errorf("failed to query count: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

// TestUnitOfWork_ConstraintViolation tests rollback on constraint violation
func TestUnitOfWork_ConstraintViolation(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	err = uow.WithTx(context.Background(), func(ctx context.Context) error {
		// First insert succeeds
		_, err := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"id1", "unique@test.com", "John", "Doe", "Mr", "1990-01-01", time.Now(), time.Now())
		if err != nil {
			return err
		}

		// Second insert fails due to unique constraint
		_, err = uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"id2", "unique@test.com", "Jane", "Doe", "Ms", "1985-05-15", time.Now(), time.Now())

		return err
	})

	if err == nil {
		t.Errorf("WithTx() should return error for constraint violation")
	}

	// Verify nothing was committed (rollback occurred)
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile")
	if err != nil {
		t.Errorf("failed to query count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 rows (transaction rolled back), got %d", count)
	}
}

// TestUnitOfWork_BeginAndRollback tests explicit begin and rollback
func TestUnitOfWork_BeginAndRollback(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Begin transaction
	err = uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() failed: %v", err)
	}

	// Insert data
	_, err = uow.tx.ExecContext(context.Background(),
		"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-id", "test@example.com", "John", "Doe", "Mr", "1990-01-01", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Rollback transaction
	err = uow.Rollback(context.Background())
	if err != nil {
		t.Errorf("Rollback() failed: %v", err)
	}

	// Verify data was rolled back
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile")
	if err != nil {
		t.Errorf("failed to query count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 rows (rolled back), got %d", count)
	}
}

// TestUnitOfWork_BeginAndCommit tests explicit begin and commit
func TestUnitOfWork_BeginAndCommit(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Begin transaction
	err = uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() failed: %v", err)
	}

	// Insert data
	_, err = uow.tx.ExecContext(context.Background(),
		"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-id", "test@example.com", "John", "Doe", "Mr", "1990-01-01", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Commit transaction
	err = uow.Commit(context.Background())
	if err != nil {
		t.Errorf("Commit() failed: %v", err)
	}

	// Verify data was committed
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile")
	if err != nil {
		t.Errorf("failed to query count: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 row (committed), got %d", count)
	}
}

// TestWithTx_Success tests generic WithTx with successful string return
func TestWithTx_Success(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	result, err := WithTx[string](context.Background(), func(ctx context.Context) (*string, error) {
		// Insert data in transaction
		_, execErr := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"test-id", "test@example.com", "John", "Doe", "Mr", "1990-01-01", time.Now(), time.Now())
		if execErr != nil {
			return nil, execErr
		}

		success := "success"
		return &success, nil
	}, uow)

	if err != nil {
		t.Errorf("WithTx() failed: %v", err)
	}

	if result == nil || *result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}

	if uow.tx != nil {
		t.Errorf("WithTx() should set tx to nil after commit, but got: %v", uow.tx)
	}

	// Verify data was committed
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "test@example.com")
	if err != nil {
		t.Errorf("failed to query inserted data: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

// TestWithTx_IntReturn tests generic WithTx with int return type
func TestWithTx_IntReturn(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	count, err := WithTx[int](context.Background(), func(ctx context.Context) (*int, error) {
		// Insert multiple rows
		queries := []struct {
			id    string
			email string
			name  string
		}{
			{"id1", "user1@test.com", "User One"},
			{"id2", "user2@test.com", "User Two"},
			{"id3", "user3@test.com", "User Three"},
		}

		for _, q := range queries {
			_, execErr := uow.tx.ExecContext(ctx,
				"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				q.id, q.email, q.name, "Test", "Mr", "1990-01-01", time.Now(), time.Now())
			if execErr != nil {
				return nil, execErr
			}
		}

		insertedCount := len(queries)
		return &insertedCount, nil
	}, uow)

	if err != nil {
		t.Errorf("WithTx() failed: %v", err)
	}

	if count == nil || *count != 3 {
		t.Errorf("expected result 3, got %v", count)
	}

	// Verify all rows were committed
	var dbCount int
	err = db.Get(&dbCount, "SELECT COUNT(*) FROM customer_profile")
	if err != nil {
		t.Errorf("failed to query count: %v", err)
	}

	if dbCount != 3 {
		t.Errorf("expected 3 rows in database, got %d", dbCount)
	}
}

// TestWithTx_StructReturn tests generic WithTx with custom struct return
func TestWithTx_StructReturn(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	type Customer struct {
		ID    string
		Email string
		Name  string
	}

	result, err := WithTx[Customer](context.Background(), func(ctx context.Context) (*Customer, error) {
		customer := &Customer{
			ID:    "cust-123",
			Email: "customer@test.com",
			Name:  "John Doe",
		}

		_, execErr := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			customer.ID, customer.Email, "John", "Doe", "Mr", "1990-01-01", time.Now(), time.Now())
		if execErr != nil {
			return nil, execErr
		}

		return customer, nil
	}, uow)

	if err != nil {
		t.Errorf("WithTx() failed: %v", err)
	}

	if result == nil || result.ID != "cust-123" || result.Email != "customer@test.com" {
		t.Errorf("expected Customer struct, got %v", result)
	}

	// Verify data was committed
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE id = ?", "cust-123")
	if err != nil {
		t.Errorf("failed to query inserted data: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

// TestWithTx_Rollback tests generic WithTx with rollback on error
func TestWithTx_Rollback(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	result, err := WithTx[string](context.Background(), func(ctx context.Context) (*string, error) {
		// Insert data that should be rolled back
		_, execErr := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"rollback-id", "rollback@example.com", "Jane", "Doe", "Ms", "1985-05-15", time.Now(), time.Now())
		if execErr != nil {
			return nil, execErr
		}

		return nil, context.DeadlineExceeded // Simulate error
	}, uow)

	if err == nil {
		t.Errorf("WithTx() should return error")
	}

	if result != nil {
		t.Errorf("WithTx() should return nil result on error, got %v", result)
	}

	if uow.tx != nil {
		t.Errorf("WithTx() should set tx to nil after rollback, but got: %v", uow.tx)
	}

	// Verify data was rolled back
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "rollback@example.com")
	if err != nil {
		t.Errorf("failed to query data: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 rows (rolled back), got %d", count)
	}
}

// TestWithTx_Panic tests generic WithTx with panic recovery
func TestWithTx_Panic(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Use defer to catch the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("WithTx() should re-raise panic")
		}
	}()

	WithTx[string](context.Background(), func(ctx context.Context) (*string, error) {
		// Insert data
		uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"panic-id", "panic@example.com", "Panic", "User", "Dr", "1970-12-31", time.Now(), time.Now())

		panic("test panic")
	}, uow)
}

// TestWithTx_Panic_Rollback tests that panic triggers rollback
func TestWithTx_Panic_Rollback(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Recover from panic
	defer func() {
		recover() // Suppress panic
	}()

	WithTx[string](context.Background(), func(ctx context.Context) (*string, error) {
		// Insert data
		uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"panic-id", "panic@example.com", "Panic", "User", "Dr", "1970-12-31", time.Now(), time.Now())

		panic("test panic")
	}, uow)

	// Verify data was rolled back
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "panic@example.com")
	if err != nil {
		t.Errorf("failed to query data: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 rows (rolled back after panic), got %d", count)
	}
}

// TestWithTx_NilReturn tests generic WithTx with nil return
func TestWithTx_NilReturn(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	result, err := WithTx[string](context.Background(), func(ctx context.Context) (*string, error) {
		// Insert data
		_, execErr := uow.tx.ExecContext(ctx,
			"INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			"nil-id", "nil@example.com", "Nil", "User", "Mr", "1990-01-01", time.Now(), time.Now())
		if execErr != nil {
			return nil, execErr
		}

		return nil, nil // Return nil but no error
	}, uow)

	if err != nil {
		t.Errorf("WithTx() failed: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}

	// Verify data was still committed
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM customer_profile WHERE email = ?", "nil@example.com")
	if err != nil {
		t.Errorf("failed to query data: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 row committed, got %d", count)
	}
}

// TestWithTx_BeginError tests generic WithTx with begin error
func TestWithTx_BeginError(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	meter := getMeterForTesting(t)
	uow, err := NewUnitOfWork(db, meter)
	if err != nil {
		t.Fatalf("NewUnitOfWork() failed: %v", err)
	}

	// Set up first transaction to cause begin to fail
	err = uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin() failed: %v", err)
	}

	// Try WithTx with existing transaction should fail
	result, err := WithTx[string](context.Background(), func(ctx context.Context) (*string, error) {
		return nil, nil
	}, uow)

	if err == nil {
		t.Errorf("WithTx() should return error when begin fails")
	}

	if result != nil {
		t.Errorf("expected nil result on begin error, got %v", result)
	}

	// Cleanup
	uow.Rollback(context.Background())
}
