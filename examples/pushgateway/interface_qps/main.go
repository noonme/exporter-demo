package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	Port                     = ":8080"
	PrometheusPushGatewayUrl = "http://localhost:9091"
	PrometheusJob            = "gin_test_prometheus_qps"
	PrometheusNameSpace      = "stathe"
	EndpointsDataSubSystem   = "endpoints"
)

/*
prometheus data type Counter：只增不减的计数器
graph
PrometheusNameSpace +  EndpointsDataSubSystem  + Name
*/
var (
	pusher              *push.Pusher
	endpointsQPSMonitor = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: PrometheusNameSpace,
			Subsystem: EndpointsDataSubSystem,
			Name:      "QPS_statistic",
			Help:      "统计QPS数据",
		}, []string{EndpointsDataSubSystem},
	)
)

func init() {
	pusher = push.New(PrometheusPushGatewayUrl, PrometheusJob)
	prometheus.MustRegister(
		endpointsQPSMonitor,
	)
	pusher.Collector(endpointsQPSMonitor)
}

// HandleEndpointQps gin server requests message
func HandleEndpointQps() gin.HandlerFunc {
	return func(c *gin.Context) {
		endpoint := c.Request.URL.Path
		fmt.Println("endpoint", endpoint)
		endpointsQPSMonitor.With(prometheus.Labels{EndpointsDataSubSystem: endpoint}).Inc()
		c.Next()
	}
}

func main() {

	r := gin.New()

	//每十五秒上报一次数据
	go func() {
		// 通过for编写一个死循环，每15s执行一次
		for range time.Tick(15 * time.Second) {
			if err := pusher.Add(); err != nil {
				log.Println(err)
			}
			log.Println("push PushGatewayServer endpoints count message is OK")
		}
	}()

	go func() {
		// var req func(endpoint string)
		req := func(endpoint string) {
			defer func() {
				//管理panic异常并处理，没有panic,r返回nil；反之打印panic信息，并阻止panic进一步向外传播
				if r := recover(); r != nil {
					log.Println("panic message", r)
				}
			}()

			//通过http.Get模拟客户端访问
			//fmt.Printf("client get url is http://0.0.0.0%s%s", Port, endpoint)
			_, err := http.Get(fmt.Sprintf("http://0.0.0.0%s%s", Port, endpoint))
			if err != nil {
				panic(err)
			}
		}
		//模拟访问: 设置访问周期
		twoSecondTicker := time.NewTicker(time.Second * 2)
		halfSecondTicker := time.NewTicker(time.Second / 2)
		for {
			select {
			case <-halfSecondTicker.C:
				req("/world")
			case <-twoSecondTicker.C:
				req("/hello")
			}

		}
	}()
	//服务端的自身的展示信息
	r.Use(HandleEndpointQps())

	// 客户端访问指定接口时,服务段展示的信息 200/json信息
	r.GET("hello", func(c *gin.Context) {
		//client 请求成功返回{hello: World} json字符
		c.JSON(http.StatusOK, gin.H{
			"hello": "World",
		})
	})
	r.GET("world", func(c *gin.Context) {
		//client 请求成功返回{hello: World} json字符
		c.JSON(http.StatusOK, gin.H{
			"world": "World",
		})
	})
	//
	r.Run(Port)
}

/*
# HELP stathe_endpoints_QPS_statistic 统计QPS数据
# TYPE stathe_endpoints_QPS_statistic counter
stathe_endpoints_QPS_statistic{endpoints="/hello",instance="",job="gin_test_prometheus_qps"} 14
stathe_endpoints_QPS_statistic{endpoints="/metrics",instance="",job="gin_test_prometheus_qps"} 1
stathe_endpoints_QPS_statistic{endpoints="/world",instance="",job="gin_test_prometheus_qps"} 59

*/
