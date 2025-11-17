package handlers

import (
	"net/http"

	"educabot.com/bookshop/services"
	"educabot.com/bookshop/utils"
	"github.com/gin-gonic/gin"
)

// GetMetricsRequest define los parámetros de query aceptados por el endpoint
type GetMetricsRequest struct {
	Author string `form:"author"`
}

// GetMetrics es el handler HTTP para el endpoint de métricas de libros
// Siguiendo el principio de Responsabilidad Única (SRP), solo maneja HTTP:
// - Validación de entrada
// - Orquestación de servicios
// - Serialización de respuesta
type GetMetrics struct {
	booksService services.BooksService
}

// NewGetMetrics crea un nuevo handler inyectando el servicio de libros
// Usa inyección de dependencias para facilitar testing y seguir el principio DIP
func NewGetMetrics(booksService services.BooksService) GetMetrics {
	return GetMetrics{
		booksService: booksService,
	}
}

// Handle retorna el handler function para Gin
// Optimizado para obtener libros una sola vez y calcular todas las métricas
func (h GetMetrics) Handle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Validar y parsear los parámetros de query
		var query GetMetricsRequest
		if err := ctx.ShouldBindQuery(&query); err != nil {
			utils.RespondWithBadRequest(ctx, "Invalid query parameters: "+err.Error())
			return
		}

		// Obtener libros una sola vez (optimización: evita 3 llamadas HTTP)
		books, err := h.booksService.GetBooks(ctx.Request.Context())
		if err != nil {
			utils.RespondWithInternalError(ctx, "Failed to retrieve books: "+err.Error())
			return
		}

		// Calcular métricas usando los libros ya obtenidos
		meanUnitsSold, err := h.booksService.CalculateMeanUnitsSold(books)
		if err != nil {
			utils.RespondWithInternalError(ctx, "Failed to calculate mean units sold: "+err.Error())
			return
		}

		cheapestBook, err := h.booksService.FindCheapestBook(books)
		if err != nil {
			utils.RespondWithInternalError(ctx, "Failed to find cheapest book: "+err.Error())
			return
		}

		booksWrittenByAuthor := h.booksService.CountBooksByAuthor(books, query.Author)

		// Responder con los datos en formato JSON
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"mean_units_sold":         meanUnitsSold,
			"cheapest_book":           cheapestBook,
			"books_written_by_author": booksWrittenByAuthor,
		})
	}
}
