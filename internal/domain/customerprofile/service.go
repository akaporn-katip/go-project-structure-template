package customerprofile

import (
	"context"
	"fmt"

	"github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	repo   Repository
	tracer trace.Tracer
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		tracer: otel.Tracer("github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile/service"),
	}
}

func (s *Service) CheckEmailAlreadyExists(context context.Context, email Email) error {
	_, span := s.tracer.Start(context, "Service.CheckEmailAlreadyExists",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()
	exists, err := s.repo.FindByEmail(context, email)

	if err != nil {
		domainError, ok := domainerrors.As(err)
		if ok && domainError.Code == "NOT_FOUND" {
			span.SetStatus(codes.Ok, err.Error())
			return nil
		} else {
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	if exists != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("EmailAlreadyExists : %v", email.String()))
		return NewEmailAlreadyExistsError(email.String())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
