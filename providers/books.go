package providers

import (
	"context"

	"educabot.com/bookshop/models"
)

// BooksProvider define la interfaz para obtener libros desde diferentes fuentes
// Retorna error para permitir manejo apropiado en capas superiores
type BooksProvider interface {
	GetBooks(ctx context.Context) ([]models.Book, error)
}
