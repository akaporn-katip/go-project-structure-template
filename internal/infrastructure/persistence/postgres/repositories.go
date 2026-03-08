package postgres

import (
	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"go.opentelemetry.io/otel/metric"
)

type PostgresRepositories struct {
	db    DatabaseExecutor
	meter metric.Meter
}

func NewPostgresRepositories(db DatabaseExecutor, meter metric.Meter) repositories.Repositories {
	return &PostgresRepositories{
		db:    db,
		meter: meter,
	}
}

func (r *PostgresRepositories) GetCustomerProfileRepository() customerprofile.Repository {
	wrapper := NewDatabaseWrapper(r.db, r.meter)
	return NewCustomerProfileRespository(wrapper)
}
