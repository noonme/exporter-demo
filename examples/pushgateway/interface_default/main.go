package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"log"
)

func main() {
	r := gin.Default()
	r.GET("/metrics", func(c *gin.Context) {
		handler := promhttp.Handler()
		handler.ServeHTTP(c.Writer, c.Request)
	})
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
