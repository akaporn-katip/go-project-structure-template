package command

import (
	"context"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/persistence/postgres"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type CreateCustomerProfileCommand struct {
	Title       string
	Firstname   string
	Lastname    string
	Email       string
	DateOfBirth string
}

type CreateCustomerProfileHandler struct {
	uow    unitofwork.UnitOfWork
	tracer trace.Tracer
}

func NewCreateCustomerProfileHandler(uow unitofwork.UnitOfWork) *CreateCustomerProfileHandler {
	return &CreateCustomerProfileHandler{
		uow:    uow,
		tracer: otel.Tracer("api.katipwork.com/crm/internal/application/customer_profile/command/create_customer_profile"),
	}
}

func (c *CreateCustomerProfileHandler) Handle(ctx context.Context, cmd CreateCustomerProfileCommand) (*customerprofile.CustomerID, error) {
	_, span := c.tracer.Start(ctx, "CreateCustomerProfileHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	return postgres.WithTx(ctx, func(ctx context.Context) (*customerprofile.CustomerID, error) {
		repos := c.uow.GetCustomerProfileRepository()
		svc := customerprofile.NewService(repos)
		return do(ctx, repos, svc, cmd)
	}, c.uow)
}

func do(ctx context.Context, repos customerprofile.Repository, svc *customerprofile.Service, cmd CreateCustomerProfileCommand) (*customerprofile.CustomerID, error) {
	customer, err := customerprofile.CreateCustomer(customerprofile.CreateCustomerProfileProps{
		Title:       cmd.Title,
		Firstname:   cmd.Firstname,
		Lastname:    cmd.Lastname,
		DateOfBirth: cmd.DateOfBirth,
		Email:       cmd.Email,
	})

	if err != nil {
		return nil, err
	}

	if err = svc.CheckEmailAlreadyExists(ctx, customer.Email()); err != nil {
		return nil, err
	}

	if err = repos.Create(ctx, customer); err != nil {
		return nil, err
	}

	id := customer.ID()
	return &id, nil
}
