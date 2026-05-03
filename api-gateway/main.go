package main

import (
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	monolith, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to parse Monolith URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(monolith)

	r.Any("/*proxyPath", func(c *gin.Context) {
		c.Request.Host = monolith.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	log.Println("API Gateway running on port 3000")
	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
