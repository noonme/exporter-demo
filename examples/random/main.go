package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	rpcDurations *prometheus.SummaryVec

	//prometheus.Histogram 类型的指标可以用于跟踪事件的规模，例如请求的持续时间或响应大小。Histogram 类型的指标会将观测数据放入有数值上界的桶中，并记录各桶中数据的个数以及所有数据的个数和数据数值总和。
	rpcDurationsHistogram prometheus.Histogram
}

func NewMetircs(reg prometheus.Registerer, normMean, normDomain float64) *metrics {

	m := &metrics{
		//创建一个摘要来跟踪虚拟的服务间RPC延迟
		//具有不同延迟分布的不同服务这些服务包括uniform/exponential/normal。
		//通过“serviceHe”标签进行区分。
		rpcDurations: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "rpc_durations_seconds",
				Help:       "RPC latency distributions.",
				Namespace:  "stathe",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"serviceHe"},
		),

		/*
			与上面相同，但现在是直方图，并且只用于正态分布。
			直方图既有传统的特征桶和稀疏桶，后者需要实验原生直方图（由普罗米修斯摄取）
			服务器v2.40与相应的特性标志启用)。
			常规储桶的目标是参数的正态分布，20桶以平均值为中心，每半西格玛宽。
			稀疏桶始终以0为中心，增长因子为从一个桶到下一个（最多）1.1。(确切的因素
			2^2^-3 = 1.0905077…)
		*/
		rpcDurationsHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:      "rpc_durations_histogram_seconds",
			Help:      "RPC latency distributions.",
			Namespace: "stathe",
			//常规桶 (start width count)
			Buckets: prometheus.LinearBuckets(normMean-5*normDomain, .5*normDomain, 20),
			//稀疏桶
			NativeHistogramBucketFactor: 1.1,
		}),
	}
	reg.MustRegister(m.rpcDurations)
	reg.MustRegister(m.rpcDurationsHistogram)
	return m

}

func main() {
	var (
		addr              = flag.String("listen-address", ":8080", "The address to listen on for http requests.")
		uniformDomain     = flag.Float64("uniform.domain", 0.0002, "The domain for the uniform  distributions.")
		normMean          = flag.Float64("noremal.mean", 0.00001, "The mean for the normal distribution.")
		normDomain        = flag.Float64("normal.domain", 0.0002, "The domain for the  normal distributions.")
		oscillationPeriod = flag.Duration("oscillation-period", 10*time.Minute, "The duration of the rate oscillation period.")
	)

	flag.Parse()

	reg := prometheus.NewRegistry()

	m := NewMetircs(reg, *normMean, *normDomain)
	/*
			“go_build_info”的收集器，其常量值为1，并带有三个标签“path”、“version”和“checksum”。它们的标签值分别包含主模块路径、版本和校验和。只有当二进制文件是用Go模块支持构建的，并且是从源存储库（而不是本地文件系统）检索的源代码构建的，这些标签才会有有意义的值。这通常是通过在GOPATH外部构建来完成的，指定主包的完整地址，例如。“GO111MODULE=on go run github.com/prometheus/client_golang/examples/random”。如果在没有Go模块支持的情况下构建，所有标签值将为“未知”。如果使用Go模块支持构建，但使用本地文件系统的源代码，则“path”将被适当设置，但“checksum”将为空，“version”将为“（devel）”。

		此收集器仅使用主模块的生成信息。请参阅https://github.com/povilasv/prommod获取模块依赖项的收集器示例。
	*/
	reg.MustRegister(collectors.NewBuildInfoCollector())

	start := time.Now()

	oscillationFactor := func() float64 {
		return 2 + math.Sin(math.Sin(2*math.Pi*float64(time.Since(start))/float64(*oscillationPeriod)))
	}

	go func() {
		for {
			v := rand.Float64() * *uniformDomain
			m.rpcDurations.WithLabelValues("uniform").Observe(v)
			time.Sleep(time.Duration(100*oscillationFactor()) * time.Millisecond)
		}
	}()

	go func() {
		for {
			v := (rand.NormFloat64() * *normDomain) + *normMean
			m.rpcDurations.WithLabelValues("normal").Observe(v)

			//通过类型断言将 prometheus.Histogram转换为 prometheus.ExemplarObserver 接口，这样就可以使用 ObserveWithExemplar 方法来记录带有示例标签的观测值
			m.rpcDurationsHistogram.(prometheus.ExemplarObserver).ObserveWithExemplar(
				v,

				prometheus.Labels{"dummyID": fmt.Sprint(rand.Intn(100000))},
			)
			time.Sleep(time.Duration(75*oscillationFactor()) * time.Millisecond)
		}
	}()

	go func() {
		for {
			v := rand.ExpFloat64() / 1e6
			m.rpcDurations.WithLabelValues("exponential").Observe(v)
			time.Sleep(time.Duration(50*oscillationFactor()) * time.Millisecond)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          reg,
		},
	))
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/*
# HELP go_build_info Build information about the main Go module.
# TYPE go_build_info gauge
go_build_info{checksum="h1:cxppBPuYhUnsO6yo/aoRol4L7q7UFfdm+bR9r+8l63Y=",path="github.com/prometheus/client_golang",version="v1.20.5"} 1


# HELP stathe_rpc_durations_histogram_seconds
# TYPE stathe_rpc_durations_histogram_seconds histogram
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0009"} 0
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0007999999999999999"} 0
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0006999999999999999"} 0
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0005999999999999998"} 0
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0004999999999999998"} 0
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0003999999999999998"} 1
stathe_rpc_durations_histogram_seconds_bucket{le="-0.0002999999999999998"} 7
stathe_rpc_durations_histogram_seconds_bucket{le="-0.00019999999999999982"} 22
stathe_rpc_durations_histogram_seconds_bucket{le="-9.999999999999982e-05"} 48
stathe_rpc_durations_histogram_seconds_bucket{le="1.8973538018496328e-19"} 81
stathe_rpc_durations_histogram_seconds_bucket{le="0.0001000000000000002"} 137
stathe_rpc_durations_histogram_seconds_bucket{le="0.0002000000000000002"} 191
stathe_rpc_durations_histogram_seconds_bucket{le="0.0003000000000000002"} 230
stathe_rpc_durations_histogram_seconds_bucket{le="0.0004000000000000002"} 251
stathe_rpc_durations_histogram_seconds_bucket{le="0.0005000000000000002"} 262
stathe_rpc_durations_histogram_seconds_bucket{le="0.0006000000000000003"} 264
stathe_rpc_durations_histogram_seconds_bucket{le="0.0007000000000000003"} 267
stathe_rpc_durations_histogram_seconds_bucket{le="0.0008000000000000004"} 267
stathe_rpc_durations_histogram_seconds_bucket{le="0.0009000000000000004"} 267
stathe_rpc_durations_histogram_seconds_bucket{le="0.0010000000000000005"} 267
stathe_rpc_durations_histogram_seconds_bucket{le="+Inf"} 267
stathe_rpc_durations_histogram_seconds_sum 0.023934494227377635
stathe_rpc_durations_histogram_seconds_count 267
# HELP stathe_rpc_duratioons_seconds RPC latency distributions.
# TYPE stathe_rpc_duratioons_seconds summary
stathe_rpc_duratioons_seconds{serviceHe="exponential",quantile="0.5"} 5.951343913478827e-07
stathe_rpc_duratioons_seconds{serviceHe="exponential",quantile="0.9"} 2.1645401862939506e-06
stathe_rpc_duratioons_seconds{serviceHe="exponential",quantile="0.99"} 4.490210675355974e-06
stathe_rpc_duratioons_seconds_sum{serviceHe="exponential"} 0.0003785796640659428
stathe_rpc_duratioons_seconds_count{serviceHe="exponential"} 401
stathe_rpc_duratioons_seconds{serviceHe="normal",quantile="0.5"} 9.370130704661714e-05
stathe_rpc_duratioons_seconds{serviceHe="normal",quantile="0.9"} 0.0003530841887625993
stathe_rpc_duratioons_seconds{serviceHe="normal",quantile="0.99"} 0.0006404658260782467
stathe_rpc_duratioons_seconds_sum{serviceHe="normal"} 0.023934494227377635
stathe_rpc_duratioons_seconds_count{serviceHe="normal"} 267
stathe_rpc_duratioons_seconds{serviceHe="uniform",quantile="0.5"} 9.953649874995836e-05
stathe_rpc_duratioons_seconds{serviceHe="uniform",quantile="0.9"} 0.00018394017636077115
stathe_rpc_duratioons_seconds{serviceHe="uniform",quantile="0.99"} 0.00019815213746912516
stathe_rpc_duratioons_seconds_sum{serviceHe="uniform"} 0.02048761379433671
stathe_rpc_duratioons_seconds_count{serviceHe="uniform"} 201

*/
