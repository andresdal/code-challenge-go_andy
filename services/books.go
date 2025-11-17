package services

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"educabot.com/bookshop/models"
	"educabot.com/bookshop/providers"
)

var (
	// ErrNoBooksAvailable indica que no hay libros disponibles para procesar
	ErrNoBooksAvailable = errors.New("no books available")
	// ErrProviderFailed indica que el provider falló al obtener los libros
	ErrProviderFailed = errors.New("failed to retrieve books from provider")
)

// BooksService define los métodos de negocio relacionados con libros
// Siguiendo el principio de Responsabilidad Única (SRP), este servicio
// se enfoca únicamente en la lógica de negocio, sin preocuparse por HTTP o presentación
type BooksService interface {
	GetBooks(ctx context.Context) ([]models.Book, error)
	CalculateMeanUnitsSold(books []models.Book) (uint, error)
	FindCheapestBook(books []models.Book) (string, error)
	CountBooksByAuthor(books []models.Book, author string) uint
}

// booksService es la implementación concreta de BooksService
type booksService struct {
	booksProvider providers.BooksProvider
}

// NewBooksService crea una nueva instancia del servicio de libros
// Usa inyección de dependencias para recibir el provider, facilitando testing y flexibilidad
func NewBooksService(provider providers.BooksProvider) BooksService {
	return &booksService{
		booksProvider: provider,
	}
}

// GetBooks obtiene los libros desde el provider
// Este método se expone para que los handlers puedan obtener los libros una sola vez
func (s *booksService) GetBooks(ctx context.Context) ([]models.Book, error) {
	books, err := s.booksProvider.GetBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}
	return books, nil
}

// CalculateMeanUnitsSold calcula el promedio de unidades vendidas de todos los libros
// Retorna error si no hay libros disponibles para evitar división por cero
func (s *booksService) CalculateMeanUnitsSold(books []models.Book) (uint, error) {
	if len(books) == 0 {
		return 0, ErrNoBooksAvailable
	}

	var sum uint
	for _, book := range books {
		sum += book.UnitsSold
	}

	return sum / uint(len(books)), nil
} // FindCheapestBook encuentra el libro con el precio más bajo
// Retorna el nombre del libro más barato o error si no hay libros
func (s *booksService) FindCheapestBook(books []models.Book) (string, error) {
	if len(books) == 0 {
		return "", ErrNoBooksAvailable
	}

	cheapest := slices.MinFunc(books, func(a, b models.Book) int {
		return int(a.Price - b.Price)
	})

	return cheapest.Name, nil
} // CountBooksByAuthor cuenta cuántos libros fueron escritos por un autor específico
// La búsqueda es case-sensitive para mantener consistencia
// No retorna error ya que 0 es un resultado válido si no hay coincidencias
func (s *booksService) CountBooksByAuthor(books []models.Book, author string) uint {
	var count uint
	for _, book := range books {
		if book.Author == author {
			count++
		}
	}

	return count
}
