package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"educabot.com/bookshop/models"
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
	// Crear request con contexto para permitir cancelación
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Establecer headers apropiados
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Ejecutar la solicitud HTTP
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Verificar código de estado HTTP
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	// Leer el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Deserializar JSON a slice de libros
	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return books, nil
}
