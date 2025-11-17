package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse representa la estructura estándar de respuesta de error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RespondWithError envía una respuesta de error estandarizada al cliente
// Centraliza el manejo de errores HTTP para mantener consistencia en toda la API
func RespondWithError(ctx *gin.Context, statusCode int, errorType string, message string) {
	ctx.JSON(statusCode, ErrorResponse{
		Error:   errorType,
		Message: message,
	})
}

// RespondWithBadRequest es un helper para errores 400 Bad Request
func RespondWithBadRequest(ctx *gin.Context, message string) {
	RespondWithError(ctx, http.StatusBadRequest, "bad_request", message)
}

// RespondWithInternalError es un helper para errores 500 Internal Server Error
func RespondWithInternalError(ctx *gin.Context, message string) {
	RespondWithError(ctx, http.StatusInternalServerError, "internal_error", message)
}

// RespondWithServiceUnavailable es un helper para errores 503 Service Unavailable
func RespondWithServiceUnavailable(ctx *gin.Context, message string) {
	RespondWithError(ctx, http.StatusServiceUnavailable, "service_unavailable", message)
}
