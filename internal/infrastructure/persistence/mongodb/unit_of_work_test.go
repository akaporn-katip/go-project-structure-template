package mongodb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUnitOfWork_ExecuteTx(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		uow, err := mongodb.NewUnitOfWork(mt.Client, "testdb")
		require.NoError(t, err)

		// Moco successful transaction commit and subsequent replies
		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}}, // start session success
		)

		err = uow.ExecuteTx(context.Background(), func(ctx context.Context, repos repositories.Repositories) error {
			assert.NotNil(t, repos)
			return nil
		})

		// Not checking deeply error as mtest lacks a strong standard way to intercept nested "session.WithTransaction"
		// If error is nil, the transaction wrapper didn't throw before attempting to commit.
		assert.NoError(t, err)
	})

	mt.Run("Error_Rollback", func(mt *mtest.T) {
		uow, err := mongodb.NewUnitOfWork(mt.Client, "testdb")
		require.NoError(t, err)

		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}}, // start session success
		)

		expectedErr := errors.New("business logic failed")

		err = uow.ExecuteTx(context.Background(), func(ctx context.Context, repos repositories.Repositories) error {
			assert.NotNil(t, repos)
			return expectedErr
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "business logic failed")
	})

}
