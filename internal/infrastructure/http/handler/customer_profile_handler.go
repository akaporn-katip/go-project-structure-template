package handler

import (
	"log/slog"

	"github.com/akaporn-katip/go-project-structure-template/internal/application/customer_profile/command"
	"github.com/akaporn-katip/go-project-structure-template/internal/application/customer_profile/dto"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/observability"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CustomerProfileHandler struct {
	createCustomerProfileHandler *command.CreateCustomerProfileHandler
	validator                    *validator.Validate
}

func NewCustomerProfileHandler(createCustomerProfileHandler *command.CreateCustomerProfileHandler, validator *validator.Validate) *CustomerProfileHandler {
	return &CustomerProfileHandler{
		createCustomerProfileHandler: createCustomerProfileHandler,
		validator:                    validator,
	}
}

func (c *CustomerProfileHandler) Create(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	traceID := observability.GetTraceID(reqCtx)

	slog.InfoContext(reqCtx, "Creating customer profile",
		"trace_id", traceID,
		"method", ctx.Request.Method,
		"path", ctx.Request.URL.Path,
	)

	var req dto.CreateCustomerProfileRequest

	if err := ctx.ShouldBind(&req); err != nil {
		slog.ErrorContext(reqCtx, "Failed to bind request",
			"trace_id", traceID,
			"error", err.Error(),
		)
		RespondInternalError(ctx, err.Error(), nil)
		return
	}

	command := command.CreateCustomerProfileCommand{
		Title:       req.Title,
		Firstname:   req.Firstname,
		Lastname:    req.Lastname,
		Email:       req.Email,
		DateOfBirth: req.DateOfBirth,
	}

	id, err := c.createCustomerProfileHandler.Handle(reqCtx, command)
	if err != nil {
		slog.ErrorContext(reqCtx, "Failed to create customer profile",
			"trace_id", traceID,
			"email", req.Email,
			"error", err.Error(),
		)
		RespondDomainError(ctx, err)
		return
	}

	slog.InfoContext(reqCtx, "Customer profile created successfully",
		"trace_id", traceID,
		"customer_id", id,
		"email", req.Email,
	)

	ResponseOK(ctx, dto.ToCustomerIDResponse(id))
}
