package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"educabot.com/bookshop/models"
	"educabot.com/bookshop/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBooksService es un mock del servicio de libros para testing
type MockBooksService struct {
	mock.Mock
}

func (m *MockBooksService) GetBooks(ctx context.Context) ([]models.Book, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBooksService) CalculateMeanUnitsSold(books []models.Book) (uint, error) {
	args := m.Called(books)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockBooksService) FindCheapestBook(books []models.Book) (string, error) {
	args := m.Called(books)
	return args.String(0), args.Error(1)
}

func (m *MockBooksService) CountBooksByAuthor(books []models.Book, author string) uint {
	args := m.Called(books, author)
	return args.Get(0).(uint)
}

func TestGetMetrics_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange
	mockService := new(MockBooksService)
	books := []models.Book{
		{ID: 1, Name: "The Go Programming Language", Author: "Alan Donovan", UnitsSold: 5000, Price: 40},
		{ID: 2, Name: "Clean Code", Author: "Robert C. Martin", UnitsSold: 15000, Price: 50},
		{ID: 3, Name: "The Pragmatic Programmer", Author: "Andrew Hunt", UnitsSold: 13000, Price: 45},
	}

	mockService.On("GetBooks", mock.Anything).Return(books, nil)
	mockService.On("CalculateMeanUnitsSold", books).Return(uint(11000), nil)
	mockService.On("FindCheapestBook", books).Return("The Go Programming Language", nil)
	mockService.On("CountBooksByAuthor", books, "Alan Donovan").Return(uint(1))

	handler := NewGetMetrics(mockService)
	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/?author=Alan+Donovan", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, 11000, int(resBody["mean_units_sold"].(float64)))
	assert.Equal(t, "The Go Programming Language", resBody["cheapest_book"])
	assert.Equal(t, 1, int(resBody["books_written_by_author"].(float64)))
	mockService.AssertExpectations(t)
}

func TestGetMetrics_WithoutAuthorParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange
	mockService := new(MockBooksService)
	books := []models.Book{
		{ID: 1, Name: "Some Book", Author: "Some Author", UnitsSold: 50000, Price: 20},
	}

	mockService.On("GetBooks", mock.Anything).Return(books, nil)
	mockService.On("CalculateMeanUnitsSold", books).Return(uint(50000), nil)
	mockService.On("FindCheapestBook", books).Return("Some Book", nil)
	mockService.On("CountBooksByAuthor", books, "").Return(uint(0))

	handler := NewGetMetrics(mockService)
	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, 50000, int(resBody["mean_units_sold"].(float64)))
	assert.Equal(t, "Some Book", resBody["cheapest_book"])
	assert.Equal(t, 0, int(resBody["books_written_by_author"].(float64)))
	mockService.AssertExpectations(t)
}

func TestGetMetrics_ServiceError_GetBooks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange
	mockService := new(MockBooksService)
	mockService.On("GetBooks", mock.Anything).Return(nil, errors.New("provider failed"))

	handler := NewGetMetrics(mockService)
	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/?author=Test", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, resBody["error"], "internal_error")
	assert.Contains(t, resBody["message"], "Failed to retrieve books")
	mockService.AssertExpectations(t)
}

func TestGetMetrics_ServiceError_MeanUnitsSold(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange
	mockService := new(MockBooksService)
	books := []models.Book{}

	mockService.On("GetBooks", mock.Anything).Return(books, nil)
	mockService.On("CalculateMeanUnitsSold", books).Return(uint(0), services.ErrNoBooksAvailable)

	handler := NewGetMetrics(mockService)
	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/?author=Test", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, resBody["error"], "internal_error")
	assert.Contains(t, resBody["message"], "Failed to calculate mean units sold")
	mockService.AssertExpectations(t)
}

func TestGetMetrics_ServiceError_CheapestBook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange
	mockService := new(MockBooksService)
	books := []models.Book{}

	mockService.On("GetBooks", mock.Anything).Return(books, nil)
	mockService.On("CalculateMeanUnitsSold", books).Return(uint(100), nil)
	mockService.On("FindCheapestBook", books).Return("", services.ErrNoBooksAvailable)

	handler := NewGetMetrics(mockService)
	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/?author=Test", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, resBody["error"], "internal_error")
	assert.Contains(t, resBody["message"], "Failed to find cheapest book")
	mockService.AssertExpectations(t)
}

func TestGetMetrics_ServiceError_CountBooksByAuthor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange - Este test ya no aplica porque CountBooksByAuthor no retorna error
	// Lo convertimos en test de validación de query params
	t.Skip("CountBooksByAuthor no retorna errores")

	mockService := new(MockBooksService)
	books := []models.Book{{ID: 1, Name: "Book", Author: "Author", UnitsSold: 100, Price: 10}}

	mockService.On("GetBooks", mock.Anything).Return(books, nil)
	mockService.On("CalculateMeanUnitsSold", books).Return(uint(100), nil)
	mockService.On("FindCheapestBook", books).Return("Book Name", nil)
	mockService.On("CountBooksByAuthor", books, "Test").Return(uint(0))

	handler := NewGetMetrics(mockService)
	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/?author=Test", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, resBody["error"], "internal_error")
	assert.Contains(t, resBody["message"], "Failed to count books by author")
	mockService.AssertExpectations(t)
}

// MockBooksProvider para test de integración con datos reales
type TestBooksProvider struct{}

func (p *TestBooksProvider) GetBooks(ctx context.Context) ([]models.Book, error) {
	return []models.Book{
		{ID: 1, Name: "The Go Programming Language", Author: "Alan Donovan", UnitsSold: 5000, Price: 40},
		{ID: 2, Name: "Clean Code", Author: "Robert C. Martin", UnitsSold: 15000, Price: 50},
		{ID: 3, Name: "The Pragmatic Programmer", Author: "Andrew Hunt", UnitsSold: 13000, Price: 45},
	}, nil
}

func TestGetMetrics_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange - Usar implementación real del servicio con un provider de prueba
	provider := &TestBooksProvider{}
	service := services.NewBooksService(provider)
	handler := NewGetMetrics(service)

	r := gin.Default()
	r.GET("/", handler.Handle())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/?author=Alan+Donovan", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	// Assert
	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, 11000, int(resBody["mean_units_sold"].(float64)))        // (5000 + 15000 + 13000) / 3
	assert.Equal(t, "The Go Programming Language", resBody["cheapest_book"]) // Precio 40
	assert.Equal(t, 1, int(resBody["books_written_by_author"].(float64)))
}
