package postgres

import (
	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
)

type PostgresRepositories struct {
	db DatabaseExecutor
}

func NewPostgresRepositories(db DatabaseExecutor) repositories.Repositories {
	return &PostgresRepositories{
		db: db,
	}
}

func (r *PostgresRepositories) CustomerProfileRepository() customerprofile.Repository {
	wrapper := NewDatabaseWrapper(r.db)
	return NewCustomerProfileRespository(wrapper)
}
