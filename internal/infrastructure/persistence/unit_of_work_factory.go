package persistence

import (
	"fmt"

	"github.com/akaporn-katip/go-project-structure-template/config"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/mongodb"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/postgres"
)

func NewUnitOfWork(conf config.DatabaseConfig) (unitofwork.UnitOfWork, error) {
	switch conf.Type {
	case "postgres":
		return newUnitOfWorkPostgres(conf.Postgres)
	case "mongodb":
		return newUnitOfWorkMongodb(conf.MongoDB)
	default:
		return nil, fmt.Errorf("not supported")
	}
}

func newUnitOfWorkMongodb(conf config.MongoConfig) (unitofwork.UnitOfWork, error) {
	client, err := mongodb.NewMongoClient(conf.URI)
	if err != nil {
		return nil, err
	}
	return mongodb.NewUnitOfWork(client, conf.DBName)
}

func newUnitOfWorkPostgres(conf config.PostgresConfig) (unitofwork.UnitOfWork, error) {
	db, err := postgres.NewPostgres(conf)
	if err != nil {
		return nil, err
	}
	return postgres.NewUnitOfWork(db)
}
