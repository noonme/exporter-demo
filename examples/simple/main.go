package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		//导出有关当前Go进程的指标 GCStats (base metrics) and runtime/metrics (both in MemStats style and new ones).
		collectors.NewGoCollector(),
		//该收集器导出进程指标的当前状态，包括CPU、内存和文件描述符使用情况以及进程启动时间。详细的行为由提供的ProcessCollectorOpts定义。ProcessCollectorOpts的零值为当前进程创建一个收集器，该收集器带有一个空的名称空间字符串，并且没有错误报告。
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			Namespace: "stathe",
		}),
	)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		//promhttp度量处理程序遇到的内部错误总数
		Registry: reg,
	}))

	log.Fatal(http.ListenAndServe(*addr, nil))

}
