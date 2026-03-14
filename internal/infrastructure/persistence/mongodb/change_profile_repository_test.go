package mongodb_test

import (
	"context"
	"testing"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MockSingleResult mocks the mongodb.SingleResult interface
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)

	if args.Get(0) != nil {
		return args.Error(0)
	}

	// This is slightly tricky, we need to populate 'v' if it's meant to be successful
	if dest, ok := v.(*mongodb.CustomerProfileModel); ok && len(args) > 1 && args.Get(1) != nil {
		src := args.Get(1).(mongodb.CustomerProfileModel)
		*dest = src
	}

	return args.Error(0)
}

func (m *MockSingleResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockCollectionExecutor mocks the mongodb.CollectionExecutor interface
type MockCollectionExecutor struct {
	mock.Mock
}

func (m *MockCollectionExecutor) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollectionExecutor) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) mongodb.SingleResult {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(mongodb.SingleResult)
}

func (m *MockCollectionExecutor) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, pipeline, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollectionExecutor) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	args := m.Called(ctx, models, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.BulkWriteResult), args.Error(1)
}

func (m *MockCollectionExecutor) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCollectionExecutor) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollectionExecutor) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollectionExecutor) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	args := m.Called(ctx, fieldName, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockCollectionExecutor) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollectionExecutor) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) mongodb.SingleResult {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(mongodb.SingleResult)
}

func (m *MockCollectionExecutor) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) mongodb.SingleResult {
	args := m.Called(ctx, filter, replacement, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(mongodb.SingleResult)
}

func (m *MockCollectionExecutor) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) mongodb.SingleResult {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(mongodb.SingleResult)
}

func (m *MockCollectionExecutor) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

func TestCustomerProfileRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		customer, err := customerprofile.CreateCustomer(customerprofile.CreateCustomerProfileProps{
			Title:       "Mr",
			Firstname:   "Test",
			Lastname:    "User",
			DateOfBirth: "19900101",
			Email:       "test@example.com",
		})
		assert.NoError(t, err)

		expectedDoc := mongodb.CustomerProfileModel{
			ID:          customer.ID().String(),
			Email:       customer.Email().String(),
			FirstName:   customer.Firstname(),
			LastName:    customer.Lastname(),
			Title:       customer.Title(),
			DateOfBirth: customer.DateOfBirth().ISOString(),
			CreatedAt:   customer.CreatedAt(),
			UpdatedAt:   customer.UpdatedAt(),
		}

		mockColl.On("InsertOne", mock.Anything, expectedDoc, mock.Anything).Return(&mongo.InsertOneResult{}, nil)

		err = repo.Create(context.Background(), customer)
		assert.NoError(t, err)
		mockColl.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		customer, err := customerprofile.CreateCustomer(customerprofile.CreateCustomerProfileProps{
			Title:       "Mr",
			Firstname:   "Test",
			Lastname:    "User",
			DateOfBirth: "19900101",
			Email:       "test@example.com",
		})
		assert.NoError(t, err)

		expectedErr := mongo.CommandError{Code: 11000, Message: "duplicate key error"}
		mockColl.On("InsertOne", mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedErr)

		err = repo.Create(context.Background(), customer)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockColl.AssertExpectations(t)
	})
}

func TestCustomerProfileRepository_FindByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		mockResult := new(MockSingleResult)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		id := customerprofile.GenerateCustomerID()
		expectedEmail := "test@example.com"
		now := time.Now()

		mockModel := mongodb.CustomerProfileModel{
			ID:          id.String(),
			Title:       "Mr",
			FirstName:   "Test",
			LastName:    "User",
			Email:       expectedEmail,
			DateOfBirth: "19900101",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		filter := bson.M{"email": expectedEmail}
		mockColl.On("FindOne", mock.Anything, filter, mock.Anything).Return(mockResult)
		mockResult.On("Decode", mock.Anything).Return(nil, mockModel)

		email, _ := customerprofile.NewEmail(expectedEmail)
		customer, err := repo.FindByEmail(context.Background(), *email)

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, expectedEmail, customer.Email().String())
		mockColl.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		mockResult := new(MockSingleResult)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		expectedEmail := "notfound@example.com"
		filter := bson.M{"email": expectedEmail}

		mockColl.On("FindOne", mock.Anything, filter, mock.Anything).Return(mockResult)
		mockResult.On("Decode", mock.Anything).Return(mongo.ErrNoDocuments)

		email, _ := customerprofile.NewEmail(expectedEmail)
		customer, err := repo.FindByEmail(context.Background(), *email)

		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		domainErr := err.(*domainerrors.DomainError)
		assert.Equal(t, domainerrors.ErrCodeNotFound, domainErr.Code)
		mockColl.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		mockResult := new(MockSingleResult)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		expectedEmail := "error@example.com"
		filter := bson.M{"email": expectedEmail}
		expectedErr := mongo.CommandError{Code: 1234, Message: "some internal error"}

		mockColl.On("FindOne", mock.Anything, filter, mock.Anything).Return(mockResult)
		mockResult.On("Decode", mock.Anything).Return(expectedErr)

		email, _ := customerprofile.NewEmail(expectedEmail)
		customer, err := repo.FindByEmail(context.Background(), *email)

		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		domainErr := err.(*domainerrors.DomainError)
		assert.Equal(t, domainerrors.ErrCodeInternal, domainErr.Code)
		mockColl.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})
}

func TestCustomerProfileRepository_FindByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		mockResult := new(MockSingleResult)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		id := customerprofile.GenerateCustomerID()
		expectedEmail := "test@example.com"
		now := time.Now()

		mockModel := mongodb.CustomerProfileModel{
			ID:          id.String(),
			Title:       "Mr",
			FirstName:   "Test",
			LastName:    "User",
			Email:       expectedEmail,
			DateOfBirth: "19900101",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		filter := bson.M{"_id": id.String()}
		mockColl.On("FindOne", mock.Anything, filter, mock.Anything).Return(mockResult)
		mockResult.On("Decode", mock.Anything).Return(nil, mockModel)

		customer, err := repo.FindByID(context.Background(), id)

		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, id.String(), customer.ID().String())
		mockColl.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		mockResult := new(MockSingleResult)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		id := customerprofile.GenerateCustomerID()
		filter := bson.M{"_id": id.String()}

		mockColl.On("FindOne", mock.Anything, filter, mock.Anything).Return(mockResult)
		mockResult.On("Decode", mock.Anything).Return(mongo.ErrNoDocuments)

		customer, err := repo.FindByID(context.Background(), id)

		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		domainErr := err.(*domainerrors.DomainError)
		assert.Equal(t, domainerrors.ErrCodeNotFound, domainErr.Code)
		mockColl.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockColl := new(MockCollectionExecutor)
		mockResult := new(MockSingleResult)
		repo := mongodb.NewCustomerProfileRepository(mockColl)

		id := customerprofile.GenerateCustomerID()
		filter := bson.M{"_id": id.String()}
		expectedErr := mongo.CommandError{Code: 1234, Message: "some internal error"}

		mockColl.On("FindOne", mock.Anything, filter, mock.Anything).Return(mockResult)
		mockResult.On("Decode", mock.Anything).Return(expectedErr)

		customer, err := repo.FindByID(context.Background(), id)

		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.IsType(t, &domainerrors.DomainError{}, err)
		domainErr := err.(*domainerrors.DomainError)
		assert.Equal(t, domainerrors.ErrCodeInternal, domainErr.Code)
		mockColl.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})
}
