# 🧪 Testing Documentation

## Resumen de Cobertura

| Módulo | Tests | Cobertura | Estado |
|--------|-------|-----------|--------|
| **Providers** | 16 | 93.1% | ✅ |
| **Services** | 10 | 100.0% | ✅ |
| **Handlers** | 6 | 90.0% | ✅ |
| **Total** | 32 | ~95% | ✅ |

---

## 🎯 Estrategia de Testing

### 1. **Testing del Provider (providers/http_books_provider_test.go)**

El provider HTTP es la capa más crítica porque:
- Interactúa con servicios externos
- Maneja errores de red
- Procesa respuestas HTTP
- Implementa timeouts y cancelaciones

#### Casos de Éxito ✅
```go
TestHTTPBooksProvider_GetBooks_Success
- Verifica respuesta HTTP 200
- Valida parsing correcto del JSON
- Confirma headers HTTP apropiados
```

#### Casos de Error de Red 🌐
```go
TestHTTPBooksProvider_GetBooks_Timeout
- Cliente con timeout de 50ms
- Servidor que tarda 200ms
- Debe retornar ErrRequestExecution

TestHTTPBooksProvider_GetBooks_ConnectionError
- URL inválida (localhost:99999)
- Debe fallar con error de conexión
```

#### Casos de Cancelación ⏸️
```go
TestHTTPBooksProvider_GetBooks_ContextCancellation
- Context cancelado antes del request
- Debe detectar context.Canceled

TestHTTPBooksProvider_GetBooks_ContextTimeout
- Context con timeout de 50ms
- Servidor que tarda 200ms
- Debe detectar context.DeadlineExceeded
```

#### Casos de HTTP Errors 🚫
```go
TestHTTPBooksProvider_GetBooks_HTTPError_500
- Servidor retorna 500 Internal Server Error
- Debe retornar ErrUnexpectedStatusCode con status 500

TestHTTPBooksProvider_GetBooks_HTTPError_404
- Servidor retorna 404 Not Found
- Debe incluir información del error en el mensaje
```

#### Casos de JSON Inválido 📝
```go
TestHTTPBooksProvider_GetBooks_InvalidJSON
- JSON malformado (falta bracket)
- Debe retornar ErrParsingJSON

TestHTTPBooksProvider_GetBooks_InvalidJSONStructure
- JSON válido pero estructura incorrecta (objeto en vez de array)
- Debe fallar el unmarshal

TestHTTPBooksProvider_GetBooks_EmptyBody
- HTTP 200 pero body completamente vacío
- Debe retornar ErrEmptyResponse
```

#### Casos Edge 🎲
```go
TestHTTPBooksProvider_GetBooks_LargeResponse
- 1000 libros en la respuesta
- Verifica que el sistema escala

TestHTTPBooksProvider_GetBooks_SpecialCharactersInData
- UTF-8: 日本語 (japonés), émojis 📚
- Validar encoding correcto
```

#### Casos de Configuración ⚙️
```go
TestHTTPBooksProvider_GetBooks_InvalidURL
- Provider con URL vacía
- Debe retornar ErrInvalidURL inmediatamente

TestHTTPBooksProvider_NewHTTPBooksProvider_DefaultConfig
- Constructor sin parámetros
- Verifica valores por defecto (timeout 10s)

TestHTTPBooksProvider_NewHTTPBooksProviderWithConfig_NilClient
- Cliente HTTP nil
- Debe crear cliente por defecto
```

---

## 🛠️ Herramientas de Testing

### httptest.NewServer()
Simula un servidor HTTP real sin necesidad de levantar uno:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Simular respuesta del servidor
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(mockData)
}))
defer server.Close()

provider := NewHTTPBooksProviderWithConfig(server.URL, server.Client())
```

**Ventajas**:
- ✅ Tests rápidos (no hay red real)
- ✅ Predecibles (no dependen de API externa)
- ✅ Aislados (no afectan servicios reales)
- ✅ Controlables (simular cualquier escenario)

### Testify Assert
Framework de aserciones más expresivo que el estándar:

```go
assert.NoError(t, err)
assert.NotNil(t, books)
assert.Len(t, books, 2)
assert.Equal(t, "Book 1", books[0].Name)
assert.ErrorIs(t, err, ErrUnexpectedStatusCode)
assert.Contains(t, err.Error(), "500")
```

---

## 🎭 Errores Tipados

### Definición
```go
var (
    ErrInvalidURL           = errors.New("invalid provider URL")
    ErrRequestCreation      = errors.New("failed to create HTTP request")
    ErrRequestExecution     = errors.New("failed to execute HTTP request")
    ErrUnexpectedStatusCode = errors.New("unexpected HTTP status code")
    ErrReadingResponse      = errors.New("failed to read response body")
    ErrParsingJSON          = errors.New("failed to parse JSON response")
    ErrEmptyResponse        = errors.New("received empty response from server")
)
```

### Uso en Tests
```go
// Verificar tipo de error específico
assert.ErrorIs(t, err, ErrUnexpectedStatusCode)

// Verificar contexto adicional
assert.Contains(t, err.Error(), "500")
assert.Contains(t, err.Error(), "Internal Server Error")
```

### Wrapping de Errores
```go
// En el código de producción
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return nil, fmt.Errorf("%w: %d (body: %s)", ErrUnexpectedStatusCode, resp.StatusCode, string(body))
}

// En el test
assert.ErrorIs(t, err, ErrUnexpectedStatusCode) // ✅ Funciona gracias a %w
```

---

## 📊 Métricas de Testing

### Tiempo de Ejecución
```
Providers: 16 tests en ~1.1s
Services:  10 tests en ~0.5s
Handlers:  6 tests en ~0.8s
Total:     32 tests en ~2.4s
```

### Cobertura por Funcionalidad

**Provider GetBooks()**:
- ✅ Success path (1 test)
- ✅ HTTP errors (2 tests: 404, 500)
- ✅ JSON errors (3 tests: malformado, estructura, vacío)
- ✅ Network errors (2 tests: timeout, conexión)
- ✅ Context errors (2 tests: cancelación, timeout)
- ✅ Edge cases (3 tests: array vacío, datos grandes, caracteres especiales)
- ✅ Configuration (3 tests: URL vacía, constructor, nil client)

**Cobertura: 93.1%** (solo líneas de configuración inicial sin cubrir)

---

## 🚀 Ejecutar Tests

### Todos los tests
```bash
go test ./... -v
```

### Solo providers
```bash
go test ./providers -v
```

### Con cobertura detallada
```bash
go test ./providers -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Tests específicos
```bash
go test ./providers -run TestHTTPBooksProvider_GetBooks_Timeout -v
```

### Tests en paralelo
```bash
go test ./... -v -parallel 4
```

---

## ✨ Mejores Prácticas Aplicadas

### 1. **Patrón AAA (Arrange-Act-Assert)**
```go
func TestHTTPBooksProvider_GetBooks_Success(t *testing.T) {
    // Arrange: Setup del test
    server := httptest.NewServer(...)
    provider := NewHTTPBooksProviderWithConfig(...)
    
    // Act: Ejecutar la función bajo test
    books, err := provider.GetBooks(context.Background())
    
    // Assert: Verificar resultados
    assert.NoError(t, err)
    assert.Len(t, books, 2)
}
```

### 2. **Nombres Descriptivos**
```go
TestHTTPBooksProvider_GetBooks_HTTPError_500
// ↑ Componente     ↑ Método  ↑ Escenario
```

### 3. **Tests Independientes**
Cada test crea su propio servidor mock y provider, sin estado compartido.

### 4. **Tests Determinísticos**
No hay sleeps aleatorios ni dependencias de red real.

### 5. **Cleanup Apropiado**
```go
server := httptest.NewServer(...)
defer server.Close() // Siempre cerrar recursos
```

### 6. **Errores Informativos**
```go
assert.Equal(t, expectedValue, actualValue, "Expected X but got Y")
```

---

## 🔮 Próximas Mejoras

### Testing
1. **Benchmarks**: Medir performance con `go test -bench`
2. **Table-driven tests**: Reducir código repetitivo
3. **Integration tests**: Tests end-to-end con API real
4. **Fuzzing**: `go test -fuzz` para encontrar edge cases

### Tooling
5. **Coverage reports**: CI/CD con codecov.io
6. **Test containers**: Usar Docker para tests de integración
7. **Mocking avanzado**: Testify mocks para interfaces complejas

---

## 📚 Referencias

- [Testing in Go](https://go.dev/doc/tutorial/add-a-test)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
- [Testify](https://github.com/stretchr/testify)
- [Error Wrapping](https://go.dev/blog/go1.13-errors)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
