package providers

import (
	"context"

	"educabot.com/bookshop/models"
)

type BooksProvider interface {
	GetBooks(ctx context.Context) []models.Book
}
