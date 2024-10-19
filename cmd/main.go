package main

import (
	"exporter-demo/collect"
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/metrics", collect.MetricsHandler)
	fmt.Println("Starting HTTP metrics server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting HTTP server: %v\n", err)
		os.Exit(1)
	}

}
