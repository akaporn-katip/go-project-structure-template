package http

import (
	"net/http"

	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/http/handler"
	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	CustomerProfileHandler *handler.CustomerProfileHandler
}

func NewRouter(
	handlers Handlers,
	tracingMiddleware *middleware.TracingMiddleware,
	metricsMiddleware *middleware.MetricsMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())

	if tracingMiddleware != nil {
		r.Use(tracingMiddleware.Handle())
	}

	if metricsMiddleware != nil {
		r.Use(metricsMiddleware.Handle())
	}

	if loggingMiddleware != nil {
		r.Use(loggingMiddleware.Handle())
	}

	api := r.Group("/crm-api/v1")
	api.GET("/health", healthCheck)
	api.POST("/customer-profile", handlers.CustomerProfileHandler.Create)
	api.GET("/customer-profile/:id", handlers.CustomerProfileHandler.FindByID)
	return r
}

func healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
