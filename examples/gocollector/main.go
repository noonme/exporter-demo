package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for http requests.")

func main() {
	flag.Parse()

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(

			collectors.WithGoCollectorRuntimeMetrics(
				collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/sched/latencies:seconds")})),
	)
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	fmt.Println("Hello world from new Go Collector!")
	log.Fatal(http.ListenAndServe(*addr, nil))

}
/*
开启go_sched_latencies_seconds(Go调度延迟秒)指标展示
# HELP go_sched_latencies_seconds Distribution of the time goroutines have spent in the scheduler in a runnable state before actually running. Bucket counts increase monotonically. Sourced from /sched/latencies:seconds
# TYPE go_sched_latencies_seconds histogram
go_sched_latencies_seconds_bucket{le="6.399999999999999e-08"} 26
go_sched_latencies_seconds_bucket{le="6.399999999999999e-07"} 27
go_sched_latencies_seconds_bucket{le="7.167999999999999e-06"} 30
go_sched_latencies_seconds_bucket{le="8.191999999999999e-05"} 35
go_sched_latencies_seconds_bucket{le="0.0009175039999999999"} 38
go_sched_latencies_seconds_bucket{le="0.010485759999999998"} 38
go_sched_latencies_seconds_bucket{le="0.11744051199999998"} 38
go_sched_latencies_seconds_bucket{le="+Inf"} 38
go_sched_latencies_seconds_sum 0.00028358400000000003
go_sched_latencies_seconds_count 38

*/