package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	Port2                     = ":8080"
	PrometheusPushgatewayUrl2 = "http://localhost:9091"
	PrometheusJob2            = "gin_test_prometheus_interface_error"
	PrometheusNameSpace2      = "stathe"
	EndpointsDataSubSystem2   = "endpoints"
	ErrorCodeDataSubsystem    = "code"
)

/*
graph
PrometheusNameSpace +  EndpointsDataSubSystem  + Name
Gauge可增减的仪表盘
*/
var (
	pusher2                   *push.Pusher
	endpointsErrorcodeMonitor = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: PrometheusNameSpace2,
			Subsystem: EndpointsDataSubSystem2,
			Name:      "errcode_statistic",
			Help:      "统计接口错误码信息数据",
		}, []string{EndpointsDataSubSystem2, ErrorCodeDataSubsystem},
	)
)

type RespCode struct {
	Code int
	Msg  string
}

func NewRespCode(code int, msg string) RespCode {
	return RespCode{
		code,
		msg,
	}
}

var (
	SUCCESS        = NewRespCode(1000, "Success")
	ERROR_MYSQL    = NewRespCode(2000, "Mysql发送错误")
	ERROR_REDIS    = NewRespCode(2001, "Redis发送错误")
	ERROR_INTERNAL = NewRespCode(2002, "Internal发送错误")
)

type DataResp struct {
	Code int
	Msg  string
	Data any
}

func NewDataResp(respCode RespCode, data any) DataResp {
	return DataResp{
		respCode.Code,
		respCode.Msg,
		data,
	}
}

func init() {
	pusher2 = push.New(PrometheusPushgatewayUrl2, PrometheusJob2)
	prometheus.MustRegister(
		endpointsErrorcodeMonitor,
	)
	pusher2.Collector(endpointsErrorcodeMonitor)
}

type Model struct {
	gin.ResponseWriter
	respBody *bytes.Buffer
}

func newModel(c *gin.Context) *Model {
	return &Model{
		c.Writer,
		bytes.NewBuffer([]byte{}),
	}
}

func (s Model) Write(b []byte) (int, error) {
	s.respBody.Write(b)
	return s.ResponseWriter.Write(b)
}

func HandleEndpointErrorcode() gin.HandlerFunc {
	return func(c *gin.Context) {
		endpoint := c.Request.URL.Path
		model := newModel(c)
		c.Writer = model
		defer func(c *gin.Context) {
			var resp DataResp

			fmt.Println(model.respBody.String())
			if err := json.Unmarshal(model.respBody.Bytes(), &resp); err != nil {
				log.Printf("json unmarsh respBody failed %+v", resp)
				//panic(err)
			}
			endpointsErrorcodeMonitor.With(prometheus.Labels{EndpointsDataSubSystem2: endpoint, ErrorCodeDataSubsystem: resp.Msg}).Inc()
		}(c)
		c.Next()
	}
}

func main() {
	r := gin.New()

	//15 second push metrics
	go func() {
		for range time.Tick(15 * time.Second) {
			if err := pusher2.Add(); err != nil {
				log.Println(err)
			}
			log.Println("interface time push ")
		}

	}()

	//模拟请求
	go func() {
		// var req func(endpoint string)

		req := func(endpoint string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}
			}()

			_, err := http.Get(fmt.Sprintf("http://localhost%s%s", Port2, endpoint))
			fmt.Printf("http://localhost%s%s\n", Port2, endpoint)
			if err != nil {
				panic(err)
			}
		}

		//设置访问周期
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

		//for {
		//	req("/hello")
		//	time.Sleep(1 * time.Second)
		//
		//}

	}()

	r.Use(HandleEndpointErrorcode())

	//服务应答
	var counter int
	r.GET("/hello", func(c *gin.Context) {
		counter++
		switch {
		case counter%10 == 0:
			c.JSON(http.StatusOK, NewDataResp(ERROR_MYSQL, "123"))
		case counter%2 == 1:
			c.JSON(http.StatusOK, NewDataResp(ERROR_INTERNAL, "456"))

		case counter%3 == 1:
			c.JSON(http.StatusOK, NewDataResp(ERROR_REDIS, "789"))

		default:
			c.JSON(http.StatusOK, NewDataResp(SUCCESS, "100"))

		}

	})
	r.GET("/world", func(c *gin.Context) {
		counter++
		switch {
		case counter%10 == 0:
			c.JSON(http.StatusOK, NewDataResp(ERROR_MYSQL, "123"))

		case counter%2 == 1:
			c.JSON(http.StatusOK, NewDataResp(ERROR_INTERNAL, "456"))

		case counter%3 == 1:
			c.JSON(http.StatusOK, NewDataResp(ERROR_REDIS, "789"))

		default:
			c.JSON(http.StatusOK, NewDataResp(SUCCESS, "100"))

		}

	})
	r.Run(Port2)
}
