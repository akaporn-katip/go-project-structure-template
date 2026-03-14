package customerprofileapp

import (
	"context"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/unitofwork"
	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type FindByIDQuery struct {
	ID string
}

type FindByIdQueryHandler struct {
	uow    unitofwork.UnitOfWork
	tracer trace.Tracer
}

func NewFindByIdQueryHandler(uow unitofwork.UnitOfWork) *FindByIdQueryHandler {
	return &FindByIdQueryHandler{
		uow:    uow,
		tracer: otel.Tracer("github.com/akaporn-katip/go-project-structure-template/internal/application/customerprofile/find_by_id_query"),
	}
}

func (h *FindByIdQueryHandler) Handle(ctx context.Context, query FindByIDQuery) (*customerprofile.CustomerProfile, error) {
	ctx, span := h.tracer.Start(ctx, "FindByIdQueryHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	id, err := customerprofile.NewCustomerID(query.ID)
	if err != nil {
		return nil, customerprofile.NewFindByIDNotFoundError()
	}

	return h.uow.Repositories().CustomerProfileRepository().FindByID(ctx, *id)
}
