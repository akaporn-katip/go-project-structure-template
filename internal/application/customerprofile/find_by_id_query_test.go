package customerprofileapp

import (
	"context"
	"testing"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUnitOfWork is a mock implementation of unitofwork.UnitOfWork
type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Repositories() repositories.Repositories {
	args := m.Called()
	return args.Get(0).(repositories.Repositories)
}

func (m *MockUnitOfWork) ExecuteTx(ctx context.Context, fn func(context.Context, repositories.Repositories) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// MockRepositories is a mock implementation of repositories.Repositories
type MockRepositories struct {
	mock.Mock
}

func (m *MockRepositories) CustomerProfileRepository() customerprofile.Repository {
	args := m.Called()
	return args.Get(0).(customerprofile.Repository)
}

// MockCustomerProfileRepo is a mock implementation of customerprofile.Repository
type MockCustomerProfileRepo struct {
	mock.Mock
}

func (m *MockCustomerProfileRepo) Create(ctx context.Context, customer *customerprofile.CustomerProfile) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerProfileRepo) Update(ctx context.Context, customer *customerprofile.CustomerProfile) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerProfileRepo) Delete(ctx context.Context, id customerprofile.CustomerID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerProfileRepo) FindByID(ctx context.Context, id customerprofile.CustomerID) (*customerprofile.CustomerProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customerprofile.CustomerProfile), args.Error(1)
}

func (m *MockCustomerProfileRepo) FindByEmail(ctx context.Context, email customerprofile.Email) (*customerprofile.CustomerProfile, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customerprofile.CustomerProfile), args.Error(1)
}

func TestFindByIdQueryHandler_Handle(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUow := new(MockUnitOfWork)
		mockRepos := new(MockRepositories)
		mockRepo := new(MockCustomerProfileRepo)

		handler := NewFindByIdQueryHandler(mockUow)

		idStr := "550e8400-e29b-41d4-a716-446655440000"
		query := FindByIDQuery{ID: idStr}
		id, _ := customerprofile.NewCustomerID(idStr)

		now := time.Now()
		expectedCustomer, _ := customerprofile.ReconstructCustomer(id.String(), customerprofile.ReconstructCustomerProps{
			Title:       "Mr",
			Firstname:   "John",
			Lastname:    "Doe",
			Email:       "john@example.com",
			DateOfBirth: "19900101",
		}, now, now)

		mockUow.On("Repositories").Return(mockRepos)
		mockRepos.On("CustomerProfileRepository").Return(mockRepo)
		mockRepo.On("FindByID", mock.Anything, *id).Return(expectedCustomer, nil)

		result, err := handler.Handle(context.Background(), query)

		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, result)
		mockUow.AssertExpectations(t)
		mockRepos.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidID", func(t *testing.T) {
		mockUow := new(MockUnitOfWork)
		handler := NewFindByIdQueryHandler(mockUow)

		query := FindByIDQuery{ID: "invalid"} // Too short or invalid format

		result, err := handler.Handle(context.Background(), query)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		assert.Equal(t, domainerrors.ErrCodeNotFound, err.(*domainerrors.DomainError).Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUow := new(MockUnitOfWork)
		mockRepos := new(MockRepositories)
		mockRepo := new(MockCustomerProfileRepo)

		handler := NewFindByIdQueryHandler(mockUow)

		idStr := "550e8400-e29b-41d4-a716-446655440000"
		query := FindByIDQuery{ID: idStr}
		id, _ := customerprofile.NewCustomerID(idStr)

		mockUow.On("Repositories").Return(mockRepos)
		mockRepos.On("CustomerProfileRepository").Return(mockRepo)
		mockRepo.On("FindByID", mock.Anything, *id).Return(nil, customerprofile.NewFindByIDNotFoundError())

		result, err := handler.Handle(context.Background(), query)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		assert.Equal(t, domainerrors.ErrCodeNotFound, err.(*domainerrors.DomainError).Code)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockUow := new(MockUnitOfWork)
		mockRepos := new(MockRepositories)
		mockRepo := new(MockCustomerProfileRepo)

		handler := NewFindByIdQueryHandler(mockUow)

		idStr := "550e8400-e29b-41d4-a716-446655440000"
		query := FindByIDQuery{ID: idStr}
		id, _ := customerprofile.NewCustomerID(idStr)

		mockUow.On("Repositories").Return(mockRepos)
		mockRepos.On("CustomerProfileRepository").Return(mockRepo)
		mockRepo.On("FindByID", mock.Anything, *id).Return(nil, domainerrors.NewInternalError("db error", nil))

		result, err := handler.Handle(context.Background(), query)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		assert.Equal(t, domainerrors.ErrCodeInternal, err.(*domainerrors.DomainError).Code)
	})
}
