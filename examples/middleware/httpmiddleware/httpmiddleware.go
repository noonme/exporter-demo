package httpmiddleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Middleware interface {
	WarpHandler(HandlerName string, handler http.Handler) http.HandlerFunc
}

type middleware struct {
	buckets  []float64
	registry prometheus.Registerer
}

func (m *middleware) WarpHandler(handlerName string, handler http.Handler) http.HandlerFunc {
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"handler": handlerName}, m.registry)

	requestTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name:      "http_requests_total",
			Namespace: "stathe",
			Help:      "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)

	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "http_request_duration_seconds",
			Help:      "Tracks the latencies for HTTP requests.",
			Namespace: "stathe",
			Buckets:   m.buckets,
		},
		[]string{"method", "code"},
	)

	requestSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name:      "http_request_szie_bytes",
			Namespace: "stathe",
			Help:      "Tracks the size of HTTP requests.",
		}, []string{"method", "code"},
	)

	responseSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name:      "http_response_size_bytes",
			Namespace: "stathe",
			Help:      "Tracks the size of HTTP responses.",
		}, []string{"method", "code"},
	)

	base := promhttp.InstrumentHandlerCounter(
		requestTotal,
		promhttp.InstrumentHandlerDuration(
			requestDuration,
			promhttp.InstrumentHandlerRequestSize(
				requestSize,
				promhttp.InstrumentHandlerResponseSize(
					responseSize,
					handler,
				),
			),
		),
	)
	return base.ServeHTTP

}

func New(registry prometheus.Registerer, buckets []float64) Middleware {
	if buckets == nil {
		buckets = prometheus.ExponentialBuckets(0.1, 1.5, 5)
	}
	return &middleware{
		buckets:  buckets,
		registry: registry,
	}
}

/*
注册不同类型的指标
# HELP stathe_http_request_duration_seconds Tracks the latencies for HTTP requests.
# TYPE stathe_http_request_duration_seconds histogram
stathe_http_request_duration_seconds_bucket{code="200",handler="/metrics",method="get",le="0.1"} 9
stathe_http_request_duration_seconds_bucket{code="200",handler="/metrics",method="get",le="0.15000000000000002"} 9
stathe_http_request_duration_seconds_bucket{code="200",handler="/metrics",method="get",le="0.22500000000000003"} 9
stathe_http_request_duration_seconds_bucket{code="200",handler="/metrics",method="get",le="0.3375"} 9
stathe_http_request_duration_seconds_bucket{code="200",handler="/metrics",method="get",le="0.5062500000000001"} 9
stathe_http_request_duration_seconds_bucket{code="200",handler="/metrics",method="get",le="+Inf"} 9
stathe_http_request_duration_seconds_sum{code="200",handler="/metrics",method="get"} 0.042010321999999996
stathe_http_request_duration_seconds_count{code="200",handler="/metrics",method="get"} 9
# HELP stathe_http_request_szie_bytes Tracks the size of HTTP requests.
# TYPE stathe_http_request_szie_bytes summary
stathe_http_request_szie_bytes_sum{code="200",handler="/metrics",method="get"} 1371
stathe_http_request_szie_bytes_count{code="200",handler="/metrics",method="get"} 9
# HELP stathe_http_requests_total Tracks the number of HTTP requests.
# TYPE stathe_http_requests_total counter
stathe_http_requests_total{code="200",handler="/metrics",method="get"} 9
# HELP stathe_http_response_size_bytes Tracks the size of HTTP responses.
# TYPE stathe_http_response_size_bytes summary
stathe_http_response_size_bytes_sum{code="200",handler="/metrics",method="get"} 56033
stathe_http_response_size_bytes_count{code="200",handler="/metrics",method="get"} 9
*/
