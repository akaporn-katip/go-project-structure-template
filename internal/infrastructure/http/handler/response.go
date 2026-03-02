package handler

import (
	"context"
	"fmt"
	nethttp "net/http"
	"strings"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/core/domainerrors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	TraceID    string      `json:"trace_id,omitempty"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// ===================================================================
// Response Writers
// ===================================================================
func RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		TraceID: getTraceID(ctx.Request.Context()),
	}

	ctx.JSON(statusCode, response)
}

// RespondSuccessWithMessage writes a successful JSON response with a message
func RespondSuccessWithMessage(ctx *gin.Context, statusCode int, message string, data interface{}) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
		TraceID: getTraceID(ctx.Request.Context()),
	}

	ctx.JSON(statusCode, response)
}

func ResponseCreated(ctx *gin.Context, data interface{}) {
	RespondSuccess(ctx, nethttp.StatusCreated, data)
}

func ResponseOK(ctx *gin.Context, data interface{}) {
	RespondSuccess(ctx, nethttp.StatusOK, data)
}

func ResponseNoContent(ctx *gin.Context, data interface{}) {
	ctx.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Writer.WriteHeader(nethttp.StatusNoContent)
}

func RespondAccepted(ctx *gin.Context, message string) {
	RespondSuccessWithMessage(ctx, nethttp.StatusAccepted, message, nil)
}

// ===================================================================
// Error Response Writers
// ===================================================================

// RespondError writes an error JSON response
func RespondError(ctx *gin.Context, statusCode int, code string, message string, details map[string]interface{}) {
	response := ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		TraceID: getTraceID(ctx.Request.Context()),
	}
	ctx.JSON(statusCode, response)
}

func RespondDomainError(ctx *gin.Context, err error) {
	if domainErr, isOk := domainerrors.As(err); isOk {
		response := ErrorResponse{
			Success: false,
			Error: ErrorDetail{
				Code:    string(domainErr.Code),
				Message: domainErr.Message,
				Details: domainErr.Details,
			},
			TraceID: getTraceID(ctx),
		}

		ctx.JSON(domainErr.StatusCode, response)
		return
	}

	RespondInternalError(ctx, "An unexpected error occurred", nil)
}

// RespondBadRequest writes a 400 Bad Request response
func RespondBadRequest(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusBadRequest, "BAD_REQUEST", message, details)
}

// RespondUnauthorized writes a 401 Unauthorized response
func RespondUnauthorized(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusUnauthorized, "UNAUTHORIZED", message, details)
}

// RespondForbidden writes a 403 Forbidden response
func RespondForbidden(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusForbidden, "FORBIDDEN", message, details)
}

// RespondNotFound writes a 404 Not Found response
func RespondNotFound(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusNotFound, "NOT_FOUND", message, details)
}

// RespondConflict writes a 409 Conflict response
func RespondConflict(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusConflict, "CONFLICT", message, details)
}

// RespondUnprocessableEntity writes a 422 Unprocessable Entity response
func RespondUnprocessableEntity(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message, details)
}

// RespondInternalError writes a 500 Internal Server Error response
func RespondInternalError(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusInternalServerError, "INTERNAL_ERROR", message, details)
}

// RespondServiceUnavailable writes a 503 Service Unavailable response
func RespondServiceUnavailable(ctx *gin.Context, message string, details map[string]interface{}) {
	RespondError(ctx, nethttp.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, details)
}

// ===================================================================
// Paginated Response Writers
// ===================================================================

// RespondPaginated writes a paginated JSON response
func RespondPaginated(ctx *gin.Context, data interface{}, page, pageSize, totalItems int) {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	response := PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		TraceID: getTraceID(ctx.Request.Context()),
	}
	ctx.JSON(nethttp.StatusOK, response)
}

// ===================================================================
// Validation Error Response
// ===================================================================

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// RespondValidationErrors writes a validation error response
func RespondValidationErrors(ctx *gin.Context, validationErrors []ValidationError) {
	details := map[string]interface{}{
		"validation_errors": validationErrors,
	}
	// RespondUnprocessableEntity(w, r, "Validation failed", details)

	RespondUnprocessableEntity(ctx, "Validation failed", details)
}

// ===================================================================
// Helper Functions
// ===================================================================

// getTraceID extracts the trace ID from context
func getTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func ParsePaginationParams(ctx *gin.Context) (page int, pageSize int) {
	// page = getQueryInt(r, "page", 1)
	page = getQueryInt(ctx, "page", 1)
	pageSize = getQueryInt(ctx, "page_size", 10)

	// Validate and apply limits
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// getQueryInt extracts an integer query parameter with a default value
func getQueryInt(ctx *gin.Context, key string, defaultValue int) int {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}

	return result
}

// getQueryString extracts a string query parameter with a default value
func getQueryString(ctx *gin.Context, key string, defaultValue string) string {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getQueryBool extracts a boolean query parameter with a default value
func getQueryBool(ctx *gin.Context, key string, defaultValue bool) bool {
	value := ctx.Query(key)
	if value == "" {
		return defaultValue
	}

	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}
