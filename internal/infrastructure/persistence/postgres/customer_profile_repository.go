package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CustomerProfileModel struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	Email       string    `db:"email"`
	DateOfBirth string    `db:"date_of_birth"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type CustomerProfileRepository struct {
	customerprofile.Repository
	db     DatabaseExecutor
	tracer trace.Tracer
}

func NewCustomerProfileRespository(db DatabaseExecutor) *CustomerProfileRepository {
	return &CustomerProfileRepository{
		db:     db,
		tracer: otel.Tracer("github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/postgres/customer_profile_repository"),
	}
}

func (cp *CustomerProfileRepository) Create(context context.Context, customerProfile *customerprofile.CustomerProfile) error {
	_, span := cp.tracer.Start(context, "CustomerProfileRepository.Create")
	defer span.End()
	_, err := cp.db.ExecContext(context, "INSERT INTO customer_profile (id, email, first_name, last_name, title, date_of_birth, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		customerProfile.ID().String(),
		customerProfile.Email().String(),
		customerProfile.Firstname(),
		customerProfile.Lastname(),
		customerProfile.Title(),
		customerProfile.DateOfBirth().ISOString(),
		customerProfile.CreatedAt(),
		customerProfile.UpdatedAt(),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cp *CustomerProfileRepository) Update(context context.Context, user *customerprofile.CustomerProfile) error {
	return nil
}

func (cp *CustomerProfileRepository) Delete(context context.Context, id customerprofile.CustomerID) error {
	return nil
}

func (cp *CustomerProfileRepository) FindByID(context context.Context, id customerprofile.CustomerID) (*customerprofile.CustomerProfile, error) {
	return &customerprofile.CustomerProfile{}, nil
}

func (cp *CustomerProfileRepository) FindByEmail(ctx context.Context, email customerprofile.Email) (*customerprofile.CustomerProfile, error) {
	_, span := cp.tracer.Start(ctx, "CustomerProfileRepository.FindByEmail")
	defer span.End()
	customer := CustomerProfileModel{}
	if err := cp.db.GetContext(ctx, &customer, "SELECT * FROM customer_profile cp WHERE cp.email = $1", email.String()); err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			span.SetStatus(codes.Ok, err.Error())
			return nil, customerprofile.NewFindByEmailNotFoundError(email.String())
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, domainerrors.NewInternalError("fail to get customer profile by email", err)
	}

	entity, err := customer.toDomain()

	if err != nil {
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return entity, nil
}

func (customer *CustomerProfileModel) toDomain() (*customerprofile.CustomerProfile, error) {
	entity, err := customerprofile.ReconstructCustomer(customer.ID, customerprofile.ReconstructCustomerProps{
		Title:       customer.Title,
		Firstname:   customer.FirstName,
		Lastname:    customer.LastName,
		DateOfBirth: customer.DateOfBirth,
		Email:       customer.Email,
	}, customer.CreatedAt, customer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return entity, nil

}
