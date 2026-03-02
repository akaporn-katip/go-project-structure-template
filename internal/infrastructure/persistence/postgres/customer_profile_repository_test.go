package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/core/domainerrors"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
)

// MockDatabaseExecutor implements DatabaseExecutor for testing
type MockDatabaseExecutor struct {
	getContextResult error
	getContextCalls  int
	capturedQuery    string
	capturedArgs     []interface{}
	mockCustomer     *CustomerProfileModel
}

func (m *MockDatabaseExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m *MockDatabaseExecutor) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	m.getContextCalls++
	m.capturedQuery = query
	m.capturedArgs = args

	if m.getContextResult != nil {
		return m.getContextResult
	}

	if m.mockCustomer != nil {
		// Simulate populating the dest struct
		if customerDto, ok := dest.(*CustomerProfileModel); ok {
			*customerDto = *m.mockCustomer
		}
	}

	return nil
}

func (m *MockDatabaseExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestFindByEmail_Success(t *testing.T) {
	// Arrange
	expectedCustomer := &CustomerProfileModel{
		ID:          "123e4567-e89b-12d3-a456-426614174000",
		Title:       "Mr",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@example.com",
		DateOfBirth: "1990-01-01",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockDB := &MockDatabaseExecutor{
		mockCustomer: expectedCustomer,
	}

	repo := NewCustomerProfileRespository(mockDB)
	email, emailErr := customerprofile.NewEmail("john.doe@example.com")
	if emailErr != nil {
		t.Fatalf("failed to create email: %v", emailErr)
	}
	ctx := context.Background()

	// Act
	result, err := repo.FindByEmail(ctx, *email)

	// Assert
	if err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}

	if result == nil {
		t.Fatal("expected customer profile, but got nil")
	}

	// Verify database was called correctly
	if mockDB.getContextCalls != 1 {
		t.Errorf("expected GetContext to be called once, but was called %d times", mockDB.getContextCalls)
	}

	expectedQuery := "SELECT * FROM customer_profile cp WHERE cp.email = $1"
	if mockDB.capturedQuery != expectedQuery {
		t.Errorf("expected query '%s', but got '%s'", expectedQuery, mockDB.capturedQuery)
	}

	if len(mockDB.capturedArgs) != 1 {
		t.Errorf("expected 1 argument, but got %d", len(mockDB.capturedArgs))
	} else if mockDB.capturedArgs[0] != "john.doe@example.com" {
		t.Errorf("expected email argument 'john.doe@example.com', but got '%v'", mockDB.capturedArgs[0])
	}
}

func TestFindByEmail_NotFound(t *testing.T) {
	// Arrange
	mockDB := &MockDatabaseExecutor{
		getContextResult: errors.New("sql: no rows in result set"), // Simulate not found
	}

	repo := NewCustomerProfileRespository(mockDB)
	email, emailErr := customerprofile.NewEmail("notfound@example.com")
	if emailErr != nil {
		t.Fatalf("failed to create email: %v", emailErr)
	}
	ctx := context.Background()

	// Act
	result, err := repo.FindByEmail(ctx, *email)

	// Assert
	if result != nil {
		t.Errorf("expected nil result for not found customer, but got: %v", result)
	}

	if err == nil {
		t.Fatal("expected error for not found customer, but got nil")
	}

	// Verify it's the correct error type
	var domainErr *domainerrors.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected DomainError, but got: %T", err)
	}
	if domainErr.Code != domainerrors.ErrCodeNotFound {
		t.Errorf("expected NotFound error code, but got: %s", domainErr.Code)
	}
}

func TestFindByEmail_DatabaseError(t *testing.T) {
	// Arrange
	mockDB := &MockDatabaseExecutor{
		getContextResult: errors.New("database connection failed"), // Simulate database error
	}

	repo := NewCustomerProfileRespository(mockDB)
	email, emailErr := customerprofile.NewEmail("test@example.com")
	if emailErr != nil {
		t.Fatalf("failed to create email: %v", emailErr)
	}
	ctx := context.Background()

	// Act
	result, err := repo.FindByEmail(ctx, *email)

	// Assert
	if result != nil {
		t.Errorf("expected nil result for database error, but got: %v", result)
	}

	if err == nil {
		t.Fatal("expected error for database failure, but got nil")
	}
}

func TestFindByEmail_InvalidEmailFormat(t *testing.T) {
	// Test with invalid email (this should fail at email creation level)
	_, emailErr := customerprofile.NewEmail("invalid-email")
	// This test validates that invalid emails are caught before reaching the repository
	var domainErr *domainerrors.DomainError
	if !errors.As(emailErr, &domainErr) {
		t.Fatalf("expected DomainError, but got: %T", emailErr)
	}
	if domainErr.Code != domainerrors.ErrCodeInvalidInput {
		t.Errorf("expected InvalidInput error for invalid email, but got: %s", domainErr.Code)
	}
	// This test validates that invalid emails are caught before reaching the repository
	if domainErr.Code != domainerrors.ErrCodeInvalidInput {
		t.Errorf("expected InvalidInput error for invalid email, but got: %s", domainErr.Code)
	}
}

func TestFindByEmail_DomainReconstructionError(t *testing.T) {
	// Arrange - Create a customer with invalid data that will fail domain reconstruction
	invalidCustomer := &CustomerProfileModel{
		ID:          "invalid-uuid", // This should cause reconstruction to fail
		Title:       "Mr",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@example.com",
		DateOfBirth: "invalid-date", // This should cause reconstruction to fail
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockDB := &MockDatabaseExecutor{
		mockCustomer: invalidCustomer,
	}

	repo := NewCustomerProfileRespository(mockDB)
	email, emailErr := customerprofile.NewEmail("john.doe@example.com")
	if emailErr != nil {
		t.Fatalf("failed to create email: %v", emailErr)
	}
	ctx := context.Background()

	// Act
	result, err := repo.FindByEmail(ctx, *email)

	// Assert
	if result != nil {
		t.Errorf("expected nil result for domain reconstruction error, but got: %v", result)
	}

	if err == nil {
		t.Fatal("expected error for domain reconstruction failure, but got nil")
	}
}
