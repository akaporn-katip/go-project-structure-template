package handler

import (
	"log/slog"

	customerprofileapp "github.com/akaporn-katip/go-project-structure-template/internal/application/customerprofile"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CustomerProfileHandler struct {
	createCustomerProfileHandler *customerprofileapp.CreateCustomerProfileHandler
	findByIdHandler              *customerprofileapp.FindByIdQueryHandler
	validator                    *validator.Validate
}

func NewCustomerProfileHandler(createCustomerProfileHandler *customerprofileapp.CreateCustomerProfileHandler, findByIdHandler *customerprofileapp.FindByIdQueryHandler, validator *validator.Validate) *CustomerProfileHandler {
	return &CustomerProfileHandler{
		createCustomerProfileHandler: createCustomerProfileHandler,
		findByIdHandler:              findByIdHandler,
		validator:                    validator,
	}
}

func (c *CustomerProfileHandler) Create(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()

	slog.InfoContext(reqCtx, "Creating customer profile",
		"method", ctx.Request.Method,
		"path", ctx.Request.URL.Path,
	)

	var req customerprofileapp.CreateCustomerProfileRequest

	if err := ctx.ShouldBind(&req); err != nil {
		slog.ErrorContext(reqCtx, "Failed to bind request",
			"error", err.Error(),
		)
		RespondInternalError(ctx, err.Error(), nil)
		return
	}

	command := customerprofileapp.CreateCustomerProfileCommand{
		Title:       req.Title,
		Firstname:   req.Firstname,
		Lastname:    req.Lastname,
		Email:       req.Email,
		DateOfBirth: req.DateOfBirth,
	}

	id, err := c.createCustomerProfileHandler.Handle(reqCtx, command)
	if err != nil {
		slog.ErrorContext(reqCtx, "Failed to create customer profile",
			"email", req.Email,
			"error", err.Error(),
		)
		RespondDomainError(ctx, err)
		return
	}

	slog.InfoContext(reqCtx, "Customer profile created successfully",
		"customer_id", id,
		"email", req.Email,
	)

	ResponseOK(ctx, customerprofileapp.ToCustomerIDResponse(id))
}

func (c *CustomerProfileHandler) FindByID(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	id := ctx.Param("id")
	command := customerprofileapp.FindByIDQuery{
		ID: id,
	}

	customer, err := c.findByIdHandler.Handle(reqCtx, command)
	if err != nil {
		slog.ErrorContext(reqCtx, "Failed to find customer by id",
			"error", err.Error(),
		)
		RespondDomainError(ctx, err)
		return
	}

	ResponseOK(ctx, customerprofileapp.ToCustomerResponse(customer))
}
