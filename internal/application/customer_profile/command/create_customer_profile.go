package command

import (
	"context"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
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
		tracer: otel.Tracer("github.com/akaporn-katip/go-project-structure-template/internal/application/customer_profile/command/create_customer_profile"),
	}
}

func (c *CreateCustomerProfileHandler) Handle(ctx context.Context, cmd CreateCustomerProfileCommand) (*customerprofile.CustomerID, error) {
	_, span := c.tracer.Start(ctx, "CreateCustomerProfileHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	return unitofwork.ExecuteTx(ctx, c.uow, func(ctx context.Context, repos repositories.Repositories) (*customerprofile.CustomerID, error) {
		repo := repos.GetCustomerProfileRepository()
		svc := customerprofile.NewService(repo)
		return do(ctx, repo, svc, cmd)
	})
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
