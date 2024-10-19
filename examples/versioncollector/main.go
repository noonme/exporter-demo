package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()
	reg := prometheus.NewRegistry()
	//导出有关当前版本信息的度量。
	reg.MustRegister(version.NewCollector("example"))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		//promhttp度量处理程序遇到的内部错误总数
		Registry: reg,
	}))

	fmt.Println("hello world from new version collector!")

	log.Fatal(http.ListenAndServe(*addr, nil))
}
