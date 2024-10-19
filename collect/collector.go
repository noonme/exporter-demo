package collect

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 定义一个指标来存储系统负载
var loadAvg = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name:      "system_load_average",
		Namespace: "stathe",
		Help:      "System 1m/5m/15m load average",
	},
	[]string{"time_linux"})

// 初始化指标
func init() {
	prometheus.MustRegister(loadAvg)
}

// HTTP 处理函数
func MetricsHandler(w http.ResponseWriter, r *http.Request) {

	loads, err := GetLoad()
	if err != nil {
		http.Error(w, "Failed to get load average", http.StatusInternalServerError)
		return
	}
	for i, load := range loads {
		switch i {
		case 0:
			loadAvg.With(prometheus.Labels{"time_linux": fmt.Sprint("1m")}).Set(load)

		case 1:
			loadAvg.With(prometheus.Labels{"time_linux": fmt.Sprint("5m")}).Set(load)

		case 2:
			loadAvg.With(prometheus.Labels{"time_linux": fmt.Sprint("15m")}).Set(load)
		}
	}
	h := promhttp.Handler()
	h.ServeHTTP(w, r)
}
