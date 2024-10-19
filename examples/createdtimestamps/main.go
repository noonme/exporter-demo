package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//一个简单的示例，说明如何以OpenMetrics格式公开已创建的时间戳。

func main() {
	requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:      "http_request_duration_seconds",
		Help:      "A historagram of the HTTP request durations in seconds",
		Namespace: "stathe",
		Buckets:   prometheus.ExponentialBuckets(0.1, 1.5, 5),
	})

	//创建非全局注册表。
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		requestDurations,
	)

	go func() {
		for {
			//记录虚构的延迟。
			now := time.Now()
			requestDurations.Observe(float64(time.Since(now).Seconds()))
			time.Sleep(600 * time.Millisecond)
		}
	}()

	// 创建httpServer暴露注册的指标
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
				// EnableOpenMetricsTextCreateSamples: true,
			}),
	)
	log.Fatalln(http.ListenAndServe(":8080", nil))

}
/*
# HELP stathe_http_request_duration_seconds A historagram of the HTTP request durations in seconds
# TYPE stathe_http_request_duration_seconds histogram
stathe_http_request_duration_seconds_bucket{le="0.1"} 5
stathe_http_request_duration_seconds_bucket{le="0.15000000000000002"} 5
stathe_http_request_duration_seconds_bucket{le="0.22500000000000003"} 5
stathe_http_request_duration_seconds_bucket{le="0.3375"} 5
stathe_http_request_duration_seconds_bucket{le="0.5062500000000001"} 5
stathe_http_request_duration_seconds_bucket{le="+Inf"} 5
stathe_http_request_duration_seconds_sum 2.397e-06
stathe_http_request_duration_seconds_count 5

*/