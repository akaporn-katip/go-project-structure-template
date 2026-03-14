package mongodb_test

import (
	"context"
	"testing"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCollectionWrapper_InsertOne(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	
	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		doc := bson.M{"foo": "bar"}
		res, err := wrapper.InsertOne(context.Background(), doc)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)

		expectedErr := mtest.CommandError{Code: 1234, Message: "insert error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		doc := bson.M{"foo": "bar"}
		res, err := wrapper.InsertOne(context.Background(), doc)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, expectedErr.Message, err.(mongo.CommandError).Message) // Compare message, mongo.CommandError equals are strict
	})
}

func TestCollectionWrapper_FindOne(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	
	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)

		now := time.Now()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: "123"},
			{Key: "foo", Value: "bar"},
			{Key: "created_at", Value: now},
		}))

		filter := bson.M{"_id": "123"}
		res := wrapper.FindOne(context.Background(), filter)

		assert.NotNil(t, res)
		assert.NoError(t, res.Err())

		var doc bson.M
		err := res.Decode(&doc)
		assert.NoError(t, err)
		assert.Equal(t, "bar", doc["foo"])
	})

	mt.Run("NotFound", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "foo.bar", mtest.FirstBatch))

		filter := bson.M{"_id": "not-found"}
		res := wrapper.FindOne(context.Background(), filter)

		assert.NotNil(t, res)
		var doc bson.M
		err := res.Decode(&doc)
		assert.Error(t, err)
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})

	mt.Run("InternalError", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)

		expectedErr := mtest.CommandError{Code: 1234, Message: "find error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{"_id": "error-key"}
		res := wrapper.FindOne(context.Background(), filter)
		
		assert.NotNil(t, res)
		assert.Error(t, res.Err())
		assert.Equal(t, expectedErr.Message, res.Err().(mongo.CommandError).Message)
	})
}
