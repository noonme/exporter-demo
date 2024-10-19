package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for http requests.")

func main() {

	flag.Parse()

	requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "http_request_duration_seconds",
		Help:        "A histogram of the HTTP request durations in seconds.",
		Namespace:   "stathe",
		Buckets:     prometheus.ExponentialBuckets(0.1, 1.5, 5),
		ConstLabels: prometheus.Labels{"serviceName": "my-service-name"},
	})

	//create a new registroy.
	reg := prometheus.NewRegistry()
	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": "my-service-name"}, reg).MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)
	reg.MustRegister(requestDurations)

	go func() {
		for {
			// Record fictional latency.
			now := time.Now()
			requestDurations.Observe(time.Since(now).Seconds())
			time.Sleep(600 * time.Millisecond)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/*
# HELP stathe_http_request_duration_seconds A histogram of the HTTP request durations in seconds.
# TYPE stathe_http_request_duration_seconds histogram
stathe_http_request_duration_seconds_bucket{serviceName="my-service-name",le="0.1"} 20
stathe_http_request_duration_seconds_bucket{serviceName="my-service-name",le="0.15000000000000002"} 20
stathe_http_request_duration_seconds_bucket{serviceName="my-service-name",le="0.22500000000000003"} 20
stathe_http_request_duration_seconds_bucket{serviceName="my-service-name",le="0.3375"} 20
stathe_http_request_duration_seconds_bucket{serviceName="my-service-name",le="0.5062500000000001"} 20
stathe_http_request_duration_seconds_bucket{serviceName="my-service-name",le="+Inf"} 20
stathe_http_request_duration_seconds_sum{serviceName="my-service-name"} 9.746e-06
stathe_http_request_duration_seconds_count{serviceName="my-service-name"} 20

*/
