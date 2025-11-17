package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"educabot.com/bookshop/models"
)

var (
	// ErrInvalidURL indica que la URL del provider es inválida
	ErrInvalidURL = errors.New("invalid provider URL")
	// ErrRequestCreation indica que falló la creación del HTTP request
	ErrRequestCreation = errors.New("failed to create HTTP request")
	// ErrRequestExecution indica que falló la ejecución del HTTP request
	ErrRequestExecution = errors.New("failed to execute HTTP request")
	// ErrUnexpectedStatusCode indica que el servidor retornó un status code inesperado
	ErrUnexpectedStatusCode = errors.New("unexpected HTTP status code")
	// ErrReadingResponse indica que falló la lectura del body del response
	ErrReadingResponse = errors.New("failed to read response body")
	// ErrParsingJSON indica que falló el parseo del JSON
	ErrParsingJSON = errors.New("failed to parse JSON response")
	// ErrEmptyResponse indica que el response está vacío cuando no debería
	ErrEmptyResponse = errors.New("received empty response from server")
)

const (
	// DefaultTimeout es el tiempo máximo de espera para solicitudes HTTP
	DefaultTimeout = 10 * time.Second
	// DefaultBooksAPIURL es la URL por defecto del servicio de libros
	DefaultBooksAPIURL = "https://6781684b85151f714b0aa5db.mockapi.io/api/v1/books"
)

// HTTPBooksProvider obtiene libros desde un servicio HTTP externo
// Implementa BooksProvider siguiendo el principio de Inversión de Dependencias (DIP)
type HTTPBooksProvider struct {
	apiURL     string
	httpClient *http.Client
}

// NewHTTPBooksProvider crea un nuevo provider HTTP con configuración por defecto
// El timeout previene que requests bloqueados cuelguen la aplicación indefinidamente
func NewHTTPBooksProvider() *HTTPBooksProvider {
	return &HTTPBooksProvider{
		apiURL: DefaultBooksAPIURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// NewHTTPBooksProviderWithConfig permite personalizar la URL y el cliente HTTP
// Útil para testing o para usar diferentes endpoints
func NewHTTPBooksProviderWithConfig(apiURL string, client *http.Client) *HTTPBooksProvider {
	if client == nil {
		client = &http.Client{Timeout: DefaultTimeout}
	}
	return &HTTPBooksProvider{
		apiURL:     apiURL,
		httpClient: client,
	}
}

// GetBooks obtiene la lista de libros desde el servicio HTTP externo
// El contexto permite cancelación y timeouts a nivel de aplicación
// Retorna error para permitir manejo apropiado en capas superiores
func (p *HTTPBooksProvider) GetBooks(ctx context.Context) ([]models.Book, error) {
	// Validar URL
	if p.apiURL == "" {
		return nil, ErrInvalidURL
	}

	// Crear request con contexto para permitir cancelación
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestCreation, err)
	}

	// Establecer headers apropiados
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Ejecutar la solicitud HTTP
	resp, err := p.httpClient.Do(req)
	if err != nil {
		// Verificar si fue cancelación de contexto
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%w: %v", ErrRequestExecution, ctx.Err())
		}
		return nil, fmt.Errorf("%w: %v", ErrRequestExecution, err)
	}
	defer resp.Body.Close()

	// Verificar código de estado HTTP
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %d (body: %s)", ErrUnexpectedStatusCode, resp.StatusCode, string(body))
	}

	// Leer el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadingResponse, err)
	}

	// Validar body no vacío
	if len(body) == 0 {
		return nil, ErrEmptyResponse
	}

	// Deserializar JSON a slice de libros
	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("%w: %v (body: %s)", ErrParsingJSON, err, string(body))
	}

	return books, nil
}
