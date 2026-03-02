package customerprofile

import (
	"context"
	"errors"
	"testing"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/core/domainerrors"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	findByEmailResult *CustomerProfile
	findByEmailError  error
	findByEmailCalls  int
}

func (m *MockRepository) Create(context context.Context, user *CustomerProfile) error {
	return nil
}

func (m *MockRepository) Update(context context.Context, user *CustomerProfile) error {
	return nil
}

func (m *MockRepository) Delete(context context.Context, id CustomerID) error {
	return nil
}

func (m *MockRepository) FindByID(context context.Context, id CustomerID) (*CustomerProfile, error) {
	return nil, nil
}

func (m *MockRepository) FindByEmail(context context.Context, email Email) (*CustomerProfile, error) {
	m.findByEmailCalls++
	return m.findByEmailResult, m.findByEmailError
}

func TestCheckEmailAlreadyExists(t *testing.T) {
	tests := []struct {
		name              string
		email             string
		repoResult        *CustomerProfile
		repoError         error
		expectedError     bool
		expectedErrorType string
	}{
		{
			name:              "email does not exist",
			email:             "newuser@example.com",
			repoResult:        nil,
			repoError:         domainerrors.NewNotFoundError("find by email not found"),
			expectedError:     false,
			expectedErrorType: "",
		},
		{
			name:              "email already exists",
			email:             "existing@example.com",
			repoResult:        &CustomerProfile{}, // non-nil means exists
			repoError:         nil,
			expectedError:     true,
			expectedErrorType: "AlreadyExistsError",
		},
		{
			name:              "repository returns error",
			email:             "test@example.com",
			repoResult:        nil,
			repoError:         domainerrors.NewInternalError("database error", errors.New("fail connection to the database")),
			expectedError:     true,
			expectedErrorType: "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockRepository{
				findByEmailResult: tt.repoResult,
				findByEmailError:  tt.repoError,
			}
			service := NewService(mockRepo)

			// Create email
			email, err := NewEmail(tt.email)
			if err != nil {
				t.Fatalf("failed to create email: %v", err)
			}

			// Execute
			result := service.CheckEmailAlreadyExists(context.Background(), *email)

			// Assert
			if tt.expectedError {
				if result == nil {
					t.Errorf("expected error but got nil")
				} else {
					// You can add more specific assertions based on your DomainError implementation
					// For now, just verify the error is not nil
					if result.Error() == "" {
						t.Errorf("error message is empty")
					}
				}
			} else {
				if result != nil {
					t.Errorf("expected no error but got: %v", result.Error())
				}
			}

			// Verify repository was called
			if mockRepo.findByEmailCalls != 1 {
				t.Errorf("expected FindByEmail to be called once, but was called %d times", mockRepo.findByEmailCalls)
			}
		})
	}
}

func TestCheckEmailAlreadyExists_EmailNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		findByEmailResult: nil,
		findByEmailError:  nil,
	}
	service := NewService(mockRepo)

	email, _ := NewEmail("notfound@example.com")
	result := service.CheckEmailAlreadyExists(context.Background(), *email)

	if result != nil {
		t.Errorf("expected no error for non-existent email, but got: %v", result.Error())
	}
}

func TestCheckEmailAlreadyExists_EmailExists(t *testing.T) {
	// Create a customer to return from mock
	customer, _ := CreateCustomer(CreateCustomerProfileProps{
		Title:       "Mr",
		Firstname:   "John",
		Lastname:    "Doe",
		DateOfBirth: "1990-01-01",
		Email:       "existing@example.com",
	})

	mockRepo := &MockRepository{
		findByEmailResult: customer,
		findByEmailError:  nil,
	}
	service := NewService(mockRepo)

	email, _ := NewEmail("existing@example.com")
	result := service.CheckEmailAlreadyExists(context.Background(), *email)

	if result == nil {
		t.Errorf("expected AlreadyExistsError for existing email, but got nil")
	}
}

func TestCheckEmailAlreadyExists_RepositoryError(t *testing.T) {
	mockRepo := &MockRepository{
		findByEmailResult: nil,
		findByEmailError:  domainerrors.NewInternalError("database connection failed", errors.New("fail connection to the database")),
	}
	service := NewService(mockRepo)

	email, _ := NewEmail("test@example.com")
	result := service.CheckEmailAlreadyExists(context.Background(), *email)

	if result == nil {
		t.Errorf("expected error from repository, but got nil")
	}
}
