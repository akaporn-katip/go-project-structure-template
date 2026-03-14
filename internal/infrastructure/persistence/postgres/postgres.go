package postgres

import (
	"fmt"

	"github.com/akaporn-katip/go-project-structure-template/config"
	"github.com/jmoiron/sqlx"
)

func NewPostgres(conf config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, err
}
