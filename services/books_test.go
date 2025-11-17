package services

import (
	"context"
	"errors"
	"testing"

	"educabot.com/bookshop/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBooksProvider es un mock del BooksProvider para testing
type MockBooksProvider struct {
	mock.Mock
}

func (m *MockBooksProvider) GetBooks(ctx context.Context) ([]models.Book, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

// Tests para GetBooks
func TestGetBooks_Success(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	expectedBooks := []models.Book{
		{ID: 1, Name: "Book 1", Author: "Author 1", UnitsSold: 100, Price: 10},
	}
	mockProvider.On("GetBooks", mock.Anything).Return(expectedBooks, nil)

	service := NewBooksService(mockProvider)
	ctx := context.Background()

	// Act
	books, err := service.GetBooks(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedBooks, books)
	mockProvider.AssertExpectations(t)
}

func TestGetBooks_ProviderError(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	providerErr := errors.New("connection failed")
	mockProvider.On("GetBooks", mock.Anything).Return(nil, providerErr)

	service := NewBooksService(mockProvider)
	ctx := context.Background()

	// Act
	books, err := service.GetBooks(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.Contains(t, err.Error(), "provider error")
	mockProvider.AssertExpectations(t)
}

// Tests para CalculateMeanUnitsSold
func TestCalculateMeanUnitsSold_Success(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	books := []models.Book{
		{ID: 1, Name: "Book 1", Author: "Author 1", UnitsSold: 100, Price: 10},
		{ID: 2, Name: "Book 2", Author: "Author 2", UnitsSold: 200, Price: 20},
		{ID: 3, Name: "Book 3", Author: "Author 3", UnitsSold: 300, Price: 30},
	}

	service := NewBooksService(mockProvider)

	// Act
	result, err := service.CalculateMeanUnitsSold(books)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, uint(200), result) // (100 + 200 + 300) / 3 = 200
}

func TestCalculateMeanUnitsSold_NoBooksAvailable(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	service := NewBooksService(mockProvider)

	// Act
	result, err := service.CalculateMeanUnitsSold([]models.Book{})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrNoBooksAvailable, err)
	assert.Equal(t, uint(0), result)
}

// Tests para FindCheapestBook
func TestFindCheapestBook_Success(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	books := []models.Book{
		{ID: 1, Name: "Expensive Book", Author: "Author 1", UnitsSold: 100, Price: 50},
		{ID: 2, Name: "Cheap Book", Author: "Author 2", UnitsSold: 200, Price: 10},
		{ID: 3, Name: "Medium Book", Author: "Author 3", UnitsSold: 300, Price: 30},
	}

	service := NewBooksService(mockProvider)

	// Act
	result, err := service.FindCheapestBook(books)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Cheap Book", result)
}

func TestFindCheapestBook_NoBooksAvailable(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	service := NewBooksService(mockProvider)

	// Act
	result, err := service.FindCheapestBook([]models.Book{})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrNoBooksAvailable, err)
	assert.Equal(t, "", result)
}

// Tests para CountBooksByAuthor
func TestCountBooksByAuthor_Success(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	books := []models.Book{
		{ID: 1, Name: "Book 1", Author: "J.R.R. Tolkien", UnitsSold: 100, Price: 10},
		{ID: 2, Name: "Book 2", Author: "C.S. Lewis", UnitsSold: 200, Price: 20},
		{ID: 3, Name: "Book 3", Author: "J.R.R. Tolkien", UnitsSold: 300, Price: 30},
		{ID: 4, Name: "Book 4", Author: "J.R.R. Tolkien", UnitsSold: 400, Price: 40},
	}

	service := NewBooksService(mockProvider)

	// Act
	result := service.CountBooksByAuthor(books, "J.R.R. Tolkien")

	// Assert
	assert.Equal(t, uint(3), result)
}

func TestCountBooksByAuthor_NoMatchingBooks(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	books := []models.Book{
		{ID: 1, Name: "Book 1", Author: "Author 1", UnitsSold: 100, Price: 10},
		{ID: 2, Name: "Book 2", Author: "Author 2", UnitsSold: 200, Price: 20},
	}

	service := NewBooksService(mockProvider)

	// Act
	result := service.CountBooksByAuthor(books, "Non Existent Author")

	// Assert
	assert.Equal(t, uint(0), result)
}

func TestCountBooksByAuthor_EmptyAuthor(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	books := []models.Book{
		{ID: 1, Name: "Book 1", Author: "Author 1", UnitsSold: 100, Price: 10},
	}

	service := NewBooksService(mockProvider)

	// Act
	result := service.CountBooksByAuthor(books, "")

	// Assert
	assert.Equal(t, uint(0), result)
}

func TestCountBooksByAuthor_EmptyBooks(t *testing.T) {
	// Arrange
	mockProvider := new(MockBooksProvider)
	service := NewBooksService(mockProvider)

	// Act
	result := service.CountBooksByAuthor([]models.Book{}, "Some Author")

	// Assert
	assert.Equal(t, uint(0), result)
}
