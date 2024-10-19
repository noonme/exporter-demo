package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	Port                      = ":8080"
	PrometheusPushgatewayUrl1 = "http://localhost:9091"
	PrometheusJob1            = "gin_test_prometheus_interface_timeout"
	PrometheusNameSpace1      = "stathe"
	EndpointsDataSubSystem1   = "endpoints"
)

/*
Histogram 直方图 数据分布情况
*/
var (
	pusher1                 *push.Pusher
	endpointLantencyMonitor = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: PrometheusNameSpace1,
			Subsystem: EndpointsDataSubSystem1,
			Name:      "lantency_statistic",
			Help:      "统计接口耗时数据",
			Buckets:   []float64{1, 5, 10, 20, 50, 100, 500, 1000, 5000, 10000},
		}, []string{EndpointsDataSubSystem1},
	)
)

func init() {
	pusher1 = push.New(PrometheusPushgatewayUrl1, PrometheusJob1)
	prometheus.MustRegister(
		endpointLantencyMonitor,
	)
	pusher1.Collector(endpointLantencyMonitor)
}

func HandleEndpointLantency() gin.HandlerFunc {
	return func(c *gin.Context) {
		endpoint := c.Request.URL.Path
		fmt.Println("endpoint", endpoint)
		start := time.Now()
		defer func(c *gin.Context) {
			lantency := time.Now().Sub(start)
			lantencyStr := fmt.Sprintf("%0.3d", lantency.Nanoseconds()/1e6)
			lantencyFloat64, err := strconv.ParseFloat(lantencyStr, 64)
			if err != nil {
				panic(err)
			}
			fmt.Println("lantencyFloat64", lantencyFloat64)
			endpointLantencyMonitor.With(prometheus.Labels{EndpointsDataSubSystem1: endpoint}).Observe(lantencyFloat64)
		}(c)
		c.Next()

	}
}

func main() {
	r := gin.New()
	go func() {
		for range time.Tick(15 * time.Second) {
			if err := pusher1.Add(); err != nil {
				log.Println(err)
			}
			log.Println("push")
		}

	}()

	go func() {
		// var req func(endponit string)
		req := func(endpoint string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}

			}()

			_, err := http.Get(fmt.Sprintf("http://localhost%s%s", Port, endpoint))
			if err != nil {
				panic(err)
			}
		}

		for {
			req("/hello")
		}
	}()

	r.Use(HandleEndpointLantency())
	var count int
	r.GET("/hello", func(c *gin.Context) {
		count++
		if count%100 == 0 {
			suddenSecond := rand.Intn(10)
			time.Sleep(time.Duration(suddenSecond) * time.Second)
			c.JSON(http.StatusOK, gin.H{
				"Hello": "World",
			})
			return
		}
		normalSecond := rand.Intn(100)
		time.Sleep(time.Duration(normalSecond) * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{
			"Hello": "World",
		})
	})
	r.Run(Port)
}
