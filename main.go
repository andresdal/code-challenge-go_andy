package main

import (
	"fmt"

	"educabot.com/bookshop/handlers"
	"educabot.com/bookshop/providers"
	"educabot.com/bookshop/services"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.SetTrustedProxies(nil)

	// Inyección de dependencias siguiendo el patrón de arquitectura en capas:
	// Provider -> Service -> Handler
	// Esto sigue el principio de Inversión de Dependencias (DIP)

	// Capa de Provider: obtiene datos de fuentes externas (API HTTP)
	booksProvider := providers.NewHTTPBooksProvider()

	// Capa de Service: contiene la lógica de negocio
	booksService := services.NewBooksService(booksProvider)

	// Capa de Handler: maneja HTTP y orquesta servicios
	metricsHandler := handlers.NewGetMetrics(booksService)

	router.GET("/", metricsHandler.Handle())

	fmt.Println("Starting server on :3000")
	router.Run(":3000")
}
