package customerprofileapp

import (
	"context"
	"errors"
	"testing"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"
)

// MockCustomerProfileRepository is a mock implementation of the Repository interface
type MockCustomerProfileRepository struct {
	findByEmailCalls  int
	findByEmailError  error
	findByEmailResult *customerprofile.CustomerProfile
	createCalls       int
	createError       error
}

func (m *MockCustomerProfileRepository) Create(ctx context.Context, user *customerprofile.CustomerProfile) error {
	m.createCalls++
	return m.createError
}

func (m *MockCustomerProfileRepository) Update(ctx context.Context, user *customerprofile.CustomerProfile) error {
	return nil
}

func (m *MockCustomerProfileRepository) Delete(ctx context.Context, id customerprofile.CustomerID) error {
	return nil
}

func (m *MockCustomerProfileRepository) FindByID(ctx context.Context, id customerprofile.CustomerID) (*customerprofile.CustomerProfile, error) {
	return nil, nil
}

func (m *MockCustomerProfileRepository) FindByEmail(ctx context.Context, email customerprofile.Email) (*customerprofile.CustomerProfile, error) {
	m.findByEmailCalls++
	if m.findByEmailError != nil {
		return nil, m.findByEmailError
	}
	return m.findByEmailResult, nil
}

func TestDo(t *testing.T) {
	tests := []struct {
		name                string
		cmd                 CreateCustomerProfileCommand
		mockFindByEmailErr  error
		mockFindByEmailRes  *customerprofile.CustomerProfile
		mockCreateErr       error
		expectError         bool
		expectErrorContains string
		validateResult      func(t *testing.T, id *customerprofile.CustomerID)
		validateMockCalls   func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service)
	}{
		{
			name: "successful customer creation",
			cmd: CreateCustomerProfileCommand{
				Title:       "Mr",
				Firstname:   "John",
				Lastname:    "Doe",
				Email:       "john.doe@example.com",
				DateOfBirth: "19900115",
			},
			mockFindByEmailErr: nil,
			mockFindByEmailRes: nil, // Email does not exist
			mockCreateErr:      nil,
			expectError:        false,
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id == nil {
					t.Error("expected CustomerID to be returned, got nil")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				if repo.findByEmailCalls != 1 {
					t.Errorf("expected FindByEmail to be called once, got %d calls", repo.findByEmailCalls)
				}
				if repo.createCalls != 1 {
					t.Errorf("expected Create to be called once, got %d calls", repo.createCalls)
				}
			},
		},
		{
			name: "invalid email format",
			cmd: CreateCustomerProfileCommand{
				Title:       "Mr",
				Firstname:   "John",
				Lastname:    "Doe",
				Email:       "invalid-email",
				DateOfBirth: "19900115",
			},
			mockFindByEmailErr:  nil,
			mockFindByEmailRes:  nil,
			mockCreateErr:       nil,
			expectError:         true,
			expectErrorContains: "invalid",
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id != nil {
					t.Error("expected nil CustomerID on error, got value")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				// Should not reach repository calls on validation error
				if repo.findByEmailCalls != 0 {
					t.Errorf("expected FindByEmail not to be called, got %d calls", repo.findByEmailCalls)
				}
				if repo.createCalls != 0 {
					t.Errorf("expected Create not to be called, got %d calls", repo.createCalls)
				}
			},
		},
		{
			name: "invalid date of birth format",
			cmd: CreateCustomerProfileCommand{
				Title:       "Mr",
				Firstname:   "John",
				Lastname:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: "invalid-date",
			},
			mockFindByEmailErr:  nil,
			mockFindByEmailRes:  nil,
			mockCreateErr:       nil,
			expectError:         true,
			expectErrorContains: "INVALID_INPUT",
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id != nil {
					t.Error("expected nil CustomerID on error, got value")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				// Should not reach repository calls on validation error
				if repo.findByEmailCalls != 0 {
					t.Errorf("expected FindByEmail not to be called, got %d calls", repo.findByEmailCalls)
				}
				if repo.createCalls != 0 {
					t.Errorf("expected Create not to be called, got %d calls", repo.createCalls)
				}
			},
		},
		{
			name: "email already exists",
			cmd: CreateCustomerProfileCommand{
				Title:       "Mr",
				Firstname:   "John",
				Lastname:    "Doe",
				Email:       "existing@example.com",
				DateOfBirth: "19900115",
			},
			mockFindByEmailErr:  domainerrors.NewAlreadyExistsError("email already exists"),
			mockFindByEmailRes:  nil,
			mockCreateErr:       nil,
			expectError:         true,
			expectErrorContains: "already exists",
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id != nil {
					t.Error("expected nil CustomerID on error, got value")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				if repo.findByEmailCalls != 1 {
					t.Errorf("expected FindByEmail to be called once, got %d calls", repo.findByEmailCalls)
				}
				// Should not reach Create if email check fails
				if repo.createCalls != 0 {
					t.Errorf("expected Create not to be called, got %d calls", repo.createCalls)
				}
			},
		},
		{
			name: "repository error during email check",
			cmd: CreateCustomerProfileCommand{
				Title:       "Mr",
				Firstname:   "John",
				Lastname:    "Doe",
				Email:       "john@example.com",
				DateOfBirth: "19900115",
			},
			mockFindByEmailErr:  errors.New("database connection error"),
			mockFindByEmailRes:  nil,
			mockCreateErr:       nil,
			expectError:         true,
			expectErrorContains: "database",
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id != nil {
					t.Error("expected nil CustomerID on error, got value")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				if repo.findByEmailCalls != 1 {
					t.Errorf("expected FindByEmail to be called once, got %d calls", repo.findByEmailCalls)
				}
				// Should not reach Create if email check fails
				if repo.createCalls != 0 {
					t.Errorf("expected Create not to be called, got %d calls", repo.createCalls)
				}
			},
		},
		{
			name: "repository error during create",
			cmd: CreateCustomerProfileCommand{
				Title:       "Mr",
				Firstname:   "Jane",
				Lastname:    "Smith",
				Email:       "jane@example.com",
				DateOfBirth: "19920520",
			},
			mockFindByEmailErr:  nil,
			mockFindByEmailRes:  nil,
			mockCreateErr:       errors.New("database write error"),
			expectError:         true,
			expectErrorContains: "database",
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id != nil {
					t.Error("expected nil CustomerID on error, got value")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				if repo.findByEmailCalls != 1 {
					t.Errorf("expected FindByEmail to be called once, got %d calls", repo.findByEmailCalls)
				}
				if repo.createCalls != 1 {
					t.Errorf("expected Create to be called once, got %d calls", repo.createCalls)
				}
			},
		},
		{
			name: "multiple customers created successfully in sequence",
			cmd: CreateCustomerProfileCommand{
				Title:       "Ms",
				Firstname:   "Alice",
				Lastname:    "Johnson",
				Email:       "alice@example.com",
				DateOfBirth: "19950310",
			},
			mockFindByEmailErr: nil,
			mockFindByEmailRes: nil,
			mockCreateErr:      nil,
			expectError:        false,
			validateResult: func(t *testing.T, id *customerprofile.CustomerID) {
				if id == nil {
					t.Error("expected CustomerID to be returned, got nil")
				}
			},
			validateMockCalls: func(t *testing.T, repo *MockCustomerProfileRepository, svc *customerprofile.Service) {
				if repo.findByEmailCalls != 1 {
					t.Errorf("expected FindByEmail to be called once, got %d calls", repo.findByEmailCalls)
				}
				if repo.createCalls != 1 {
					t.Errorf("expected Create to be called once, got %d calls", repo.createCalls)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockCustomerProfileRepository{
				findByEmailError:  tt.mockFindByEmailErr,
				findByEmailResult: tt.mockFindByEmailRes,
				createError:       tt.mockCreateErr,
			}
			svc := customerprofile.NewService(mockRepo)

			// Execute
			ctx := context.Background()
			id, err := do(ctx, mockRepo, svc, tt.cmd)

			// Assert error state
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			// Assert error message if specified
			if tt.expectError && tt.expectErrorContains != "" && err != nil {
				errMsg := err.Error()
				if !contains(errMsg, tt.expectErrorContains) {
					t.Errorf("expected error to contain '%s', got: %s", tt.expectErrorContains, errMsg)
				}
			}

			// Assert result
			if tt.validateResult != nil {
				tt.validateResult(t, id)
			}

			// Assert mock calls
			if tt.validateMockCalls != nil {
				tt.validateMockCalls(t, mockRepo, svc)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
