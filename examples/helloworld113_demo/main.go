package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	requsetQps := prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "request_qps",
		Namespace: "stathe",
		Help:      "The a request qps",
	},
	)

	go func() {
		for {
			//Inc将计数器加1。
			requsetQps.Inc()
			time.Sleep(time.Second)
		}
	}()

	prometheus.MustRegister(requsetQps)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
