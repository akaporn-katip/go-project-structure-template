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

func TestCollectionWrapper_Aggregate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)

		ns := mt.Coll.Database().Name() + "." + mt.Coll.Name()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{{Key: "foo", Value: "bar"}}))

		pipeline := mongo.Pipeline{}
		res, err := wrapper.Aggregate(context.Background(), pipeline)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		// Close cursor to prevent session leaks
		defer res.Close(context.Background())
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "aggregate error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		pipeline := mongo.Pipeline{}
		res, err := wrapper.Aggregate(context.Background(), pipeline)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCollectionWrapper_BulkWrite(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 1},
			{Key: "nModified", Value: 0},
		})

		models := []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(bson.M{"foo": "bar"})}
		res, err := wrapper.BulkWrite(context.Background(), models)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "bulkwrite error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		models := []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(bson.M{"foo": "bar"})}
		res, err := wrapper.BulkWrite(context.Background(), models)
		assert.Error(t, err)
		assert.NotNil(t, res)
	})
}

func TestCollectionWrapper_CountDocuments(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(5)}}))

		filter := bson.M{}
		count, err := wrapper.CountDocuments(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "count error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{}
		count, err := wrapper.CountDocuments(context.Background(), filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestCollectionWrapper_DeleteMany(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: int32(2)}})

		filter := bson.M{}
		res, err := wrapper.DeleteMany(context.Background(), filter)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(2), res.DeletedCount)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "delete error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{}
		res, err := wrapper.DeleteMany(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCollectionWrapper_DeleteOne(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: int32(1)}})

		filter := bson.M{}
		res, err := wrapper.DeleteOne(context.Background(), filter)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.DeletedCount)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "delete error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{}
		res, err := wrapper.DeleteOne(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCollectionWrapper_Distinct(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "values", Value: bson.A{"a", "b"}}})

		filter := bson.M{}
		res, err := wrapper.Distinct(context.Background(), "field", filter)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.ElementsMatch(t, []interface{}{"a", "b"}, res)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "distinct error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{}
		res, err := wrapper.Distinct(context.Background(), "field", filter)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCollectionWrapper_Find(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		ns := mt.Coll.Database().Name() + "." + mt.Coll.Name()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{{Key: "foo", Value: "bar"}}))

		filter := bson.M{}
		res, err := wrapper.Find(context.Background(), filter)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		defer res.Close(context.Background())
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "find error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{}
		res, err := wrapper.Find(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCollectionWrapper_FindOneAndDelete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "value", Value: bson.D{{Key: "foo", Value: "bar"}}}})

		filter := bson.M{"_id": "123"}
		res := wrapper.FindOneAndDelete(context.Background(), filter)
		assert.NotNil(t, res)
		assert.NoError(t, res.Err())
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{"_id": "error-key"}
		res := wrapper.FindOneAndDelete(context.Background(), filter)
		assert.NotNil(t, res)
		assert.Error(t, res.Err())
	})
}

func TestCollectionWrapper_FindOneAndReplace(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "value", Value: bson.D{{Key: "foo", Value: "bar"}}}})

		filter := bson.M{"_id": "123"}
		res := wrapper.FindOneAndReplace(context.Background(), filter, bson.M{"foo": "baz"})
		assert.NotNil(t, res)
		assert.NoError(t, res.Err())
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{"_id": "error-key"}
		res := wrapper.FindOneAndReplace(context.Background(), filter, bson.M{"foo": "baz"})
		assert.NotNil(t, res)
		assert.Error(t, res.Err())
	})
}

func TestCollectionWrapper_FindOneAndUpdate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "value", Value: bson.D{{Key: "foo", Value: "bar"}}}})

		filter := bson.M{"_id": "123"}
		res := wrapper.FindOneAndUpdate(context.Background(), filter, bson.M{"$set": bson.M{"foo": "baz"}})
		assert.NotNil(t, res)
		assert.NoError(t, res.Err())
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		filter := bson.M{"_id": "error-key"}
		res := wrapper.FindOneAndUpdate(context.Background(), filter, bson.M{"$set": bson.M{"foo": "baz"}})
		assert.NotNil(t, res)
		assert.Error(t, res.Err())
	})
}

func TestCollectionWrapper_InsertMany(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Success", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		docs := []interface{}{bson.M{"foo": "bar"}}
		res, err := wrapper.InsertMany(context.Background(), docs)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	mt.Run("Error", func(mt *mtest.T) {
		wrapper := mongodb.NewCollectionWrapper(mt.Coll)
		expectedErr := mtest.CommandError{Code: 1234, Message: "insertmany error"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(expectedErr))

		docs := []interface{}{bson.M{"foo": "bar"}}
		res, err := wrapper.InsertMany(context.Background(), docs)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
