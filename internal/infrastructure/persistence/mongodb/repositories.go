package mongodb

import (
	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepositories struct {
	db *mongo.Database
}

func NewMongoRepositories(db *mongo.Database) repositories.Repositories {
	return &MongoRepositories{
		db: db,
	}
}

func (r *MongoRepositories) GetCustomerProfileRepository() customerprofile.Repository {
	coll := r.db.Collection("customer_profile")
	wrapper := NewCollectionWrapper(coll)
	return NewCustomerProfileRepository(wrapper)
}
