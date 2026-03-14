package migrate

import (
	"fmt"

	"github.com/akaporn-katip/go-project-structure-template/config"
	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

type SQLMigrator struct {
	migrator *migrate.Migrate
}

func (s *SQLMigrator) Up() error {
	return s.migrator.Up()
}

func (s *SQLMigrator) Down() error {
	return s.migrator.Down()
}

func NewSqlMigrator(cfg config.PostgresConfig) (*SQLMigrator, error) {
	mt, err := migrate.New("file://migrate/sql", cfg.DSN)

	if err != nil {
		return nil, err
	}

	return &SQLMigrator{
		migrator: mt,
	}, nil
}

type MongoMigrator struct {
	migrator *migrate.Migrate
}

func (s *MongoMigrator) Up() error {
	return s.migrator.Up()
}

func (s *MongoMigrator) Down() error {
	return s.migrator.Down()
}

func NewMongoMigrator(cfg config.MongoConfig) (*SQLMigrator, error) {
	mt, err := migrate.New("file://migrate/mongo", cfg.URI)

	if err != nil {
		return nil, err
	}

	return &SQLMigrator{
		migrator: mt,
	}, nil
}

func NewMigrator(cfg config.DatabaseConfig) (Migrator, error) {
	switch cfg.Type {
	case "postgres":
		return NewSqlMigrator(cfg.Postgres)
	case "mongodb":
		return NewMongoMigrator(cfg.MongoDB)
	default:
		return nil, fmt.Errorf("unsupported driver")
	}
}
