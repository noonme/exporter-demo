package main

import (
	"exporter-demo/examples/middleware/httpmiddleware"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(
		// collectors.WithGoCollectorRuntimeMetrics(
		// 	collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/sched/latencies:seconds")})
		),

		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	http.Handle(
		"/metrics",
		httpmiddleware.New(registry, nil).WarpHandler("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
