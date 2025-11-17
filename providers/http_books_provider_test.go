package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"educabot.com/bookshop/models"
	"github.com/stretchr/testify/assert"
)

func TestHTTPBooksProvider_GetBooks_Success(t *testing.T) {
	// Arrange: Mock server que retorna libros válidos
	mockBooks := []models.Book{
		{ID: 1, Name: "Book 1", Author: "Author 1", UnitsSold: 100, Price: 20},
		{ID: 2, Name: "Book 2", Author: "Author 2", UnitsSold: 200, Price: 15},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar método y headers
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockBooks)
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 2)
	assert.Equal(t, "Book 1", books[0].Name)
	assert.Equal(t, uint(100), books[0].UnitsSold)
	assert.Equal(t, uint(20), books[0].Price)
	assert.Equal(t, "Book 2", books[1].Name)
}

func TestHTTPBooksProvider_GetBooks_EmptyArrayResponse(t *testing.T) {
	// Arrange: Server que retorna array vacío
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{})
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 0)
}

func TestHTTPBooksProvider_GetBooks_InvalidURL(t *testing.T) {
	// Arrange: Provider con URL vacía
	provider := NewHTTPBooksProviderWithConfig("", nil)

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrInvalidURL)
}

func TestHTTPBooksProvider_GetBooks_HTTPError_500(t *testing.T) {
	// Arrange: Server que retorna 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrUnexpectedStatusCode)
	assert.Contains(t, err.Error(), "500")
}

func TestHTTPBooksProvider_GetBooks_HTTPError_404(t *testing.T) {
	// Arrange: Server que retorna 404 Not Found
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrUnexpectedStatusCode)
	assert.Contains(t, err.Error(), "404")
}

func TestHTTPBooksProvider_GetBooks_InvalidJSON(t *testing.T) {
	// Arrange: Server que retorna JSON malformado
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid": "json", "missing": bracket`))
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrParsingJSON)
	assert.Contains(t, err.Error(), "failed to parse JSON response")
}

func TestHTTPBooksProvider_GetBooks_InvalidJSONStructure(t *testing.T) {
	// Arrange: Server que retorna JSON válido pero con estructura incorrecta
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Retorna objeto en lugar de array
		w.Write([]byte(`{"message": "this is not an array"}`))
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrParsingJSON)
	assert.Contains(t, err.Error(), "failed to parse JSON response")
}

func TestHTTPBooksProvider_GetBooks_Timeout(t *testing.T) {
	// Arrange: Server que demora más que el timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Demora 200ms
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{})
	}))
	defer server.Close()

	// Cliente con timeout muy corto (50ms)
	client := &http.Client{Timeout: 50 * time.Millisecond}
	provider := NewHTTPBooksProviderWithConfig(server.URL, client)

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrRequestExecution)
	assert.Contains(t, err.Error(), "failed to execute HTTP request")
}

func TestHTTPBooksProvider_GetBooks_ContextCancellation(t *testing.T) {
	// Arrange: Server que demora
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{})
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Context que se cancela inmediatamente
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancelar antes de hacer la request

	// Act
	books, err := provider.GetBooks(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrRequestExecution)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestHTTPBooksProvider_GetBooks_ContextTimeout(t *testing.T) {
	// Arrange: Server lento
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{})
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Context con timeout muy corto
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Act
	books, err := provider.GetBooks(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.Contains(t, err.Error(), "failed to execute HTTP request")
}

func TestHTTPBooksProvider_GetBooks_ConnectionError(t *testing.T) {
	// Arrange: Provider con URL inválida que causa error de conexión
	provider := NewHTTPBooksProviderWithConfig("http://localhost:99999/invalid", nil)

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrRequestExecution)
}

func TestHTTPBooksProvider_GetBooks_EmptyBody(t *testing.T) {
	// Arrange: Server que retorna 200 pero sin body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No escribe nada en el body
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, books)
	assert.ErrorIs(t, err, ErrEmptyResponse)
}

func TestHTTPBooksProvider_NewHTTPBooksProvider_DefaultConfig(t *testing.T) {
	// Act
	provider := NewHTTPBooksProvider()

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, DefaultBooksAPIURL, provider.apiURL)
	assert.NotNil(t, provider.httpClient)
	assert.Equal(t, DefaultTimeout, provider.httpClient.Timeout)
}

func TestHTTPBooksProvider_NewHTTPBooksProviderWithConfig_NilClient(t *testing.T) {
	// Act: Pasar nil como cliente debe crear uno por defecto
	provider := NewHTTPBooksProviderWithConfig("http://test.com", nil)

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, "http://test.com", provider.apiURL)
	assert.NotNil(t, provider.httpClient)
	assert.Equal(t, DefaultTimeout, provider.httpClient.Timeout)
}

func TestHTTPBooksProvider_GetBooks_LargeResponse(t *testing.T) {
	// Arrange: Server que retorna muchos libros
	mockBooks := make([]models.Book, 1000)
	for i := 0; i < 1000; i++ {
		mockBooks[i] = models.Book{
			ID:        uint(i + 1),
			Name:      "Book " + string(rune(i)),
			Author:    "Author " + string(rune(i)),
			UnitsSold: uint(i * 100),
			Price:     uint(10 + i),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockBooks)
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 1000)
}

func TestHTTPBooksProvider_GetBooks_SpecialCharactersInData(t *testing.T) {
	// Arrange: Libros con caracteres especiales
	mockBooks := []models.Book{
		{ID: 1, Name: "Book with émojis 📚", Author: "Ñoño Sánchez", UnitsSold: 100, Price: 20},
		{ID: 2, Name: `Book with "quotes"`, Author: "O'Brien", UnitsSold: 200, Price: 15},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockBooks)
	}))
	defer server.Close()

	provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())

	// Act
	books, err := provider.GetBooks(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, books)
	assert.Len(t, books, 2)
	assert.Equal(t, "Book with émojis 📚", books[0].Name)
	assert.Equal(t, "Ñoño Sánchez", books[0].Author)
}
