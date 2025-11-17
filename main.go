package main

import (
	"fmt"

	"educabot.com/bookshop/handlers"
	"educabot.com/bookshop/repositories/mockImpls"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.SetTrustedProxies(nil)

	metricsHandler := handlers.NewGetMetrics(mockImpls.NewMockBooksProvider())
	router.GET("/", metricsHandler.Handle())
	router.Run(":3000")
	fmt.Println("Starting server on :3000")
}
