Prometheus 官方和社区提供了非常多的exporter，涵盖数据库、中间件、OS、存储、硬件设备等，具体可查看[exporters](https://link.zhihu.com/?target=https%3A//github.com/prometheus/docs/blob/main/content/docs/instrumenting/exporters.md)、[exporterhub.io](https://link.zhihu.com/?target=https%3A//exporterhub.io/)，通过这些 exporter 基本可以覆盖80%的监控需求，依然有小部分需要通过自定义脚本或者定制、修改社区exporter实现。本文我们将学习如何通过go编写一个简单的expoter用于暴露OS的负载。

要实现的三个load指标如下：
![mark](http://images.opsblogs.cn/blog/20241016/lpM81FejAD8R.png?imageslim)

exporter的核心是http服务，对外暴露exporter本身运行时指标和监控信息。我们可以直接通过`net/http`暴力实现，更好的方式是使用Prometheus 官方提供的client library 来简化一部分工作。

client library官方支持语言：

- [Go](https://link.zhihu.com/?target=https%3A//github.com/prometheus/client_golang)
- [Java or Scala](https://link.zhihu.com/?target=https%3A//github.com/prometheus/client_java)
- [Python](https://link.zhihu.com/?target=https%3A//github.com/prometheus/client_python)
- [Ruby](https://link.zhihu.com/?target=https%3A//github.com/prometheus/client_ruby)
- [Rust](https://link.zhihu.com/?target=https%3A//github.com/prometheus/client_rust)

也有社区支持的其他语言库如C、C++、PHP等

### **获取数据源**



------



在使用client library暴露数据之前，我们得先找到数据源，以linux为例要获取系统负载我们可以读取/proc目录下的loadavg文件。涉及到各类操作系统指标的获取可以参考官方的[node-exporter](https://link.zhihu.com/?target=https%3A//github.com/prometheus/node_exporter)，这里我们给他写成load包，等会直接调用GetLoad()就能拿到数据了。
```go
package collect

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var procPath = "/proc/loadavg"

// / 读取系统负载
func GetLoad() (loads []float64, err error) {
	data, err := os.ReadFile(procPath)
	if err != nil {
		return nil, err
	}
	//uint8 --> float64 && 数据处理
	loads, err = parseLoad(string(data))
	if err != nil {
		return nil, err
	}
	return loads, nil
}

// Parse /proc loadavg and return 1m, 5m and 15m.
/*
为什么这里只有长度为3，因返回的只有三个可用，如果设置大于实际返回会导致指标丢失
cat /proc/loadavg
0.52 0.56 0.54 2/1370 226866
*/
func parseLoad(data string) (loads []float64, err error) {
	loads = make([]float64, 3)
	parts := strings.Fields(data)
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected content in %s", procPath)
	}
	for i, load := range parts[0:3] {
		loads[i], err = strconv.ParseFloat(load, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse load '%s': %w", load, err)
		}
	}
	return loads, nil
}

```
### **通过client_golang暴露指标**



------



开通我们提到exporter要暴露的指标包含两部分，一是本身的运行时信息，另一个监控的metrics。而运行时信息`client_golang`已经帮我们实现了，我们要做的是通过`client_golang`包将监控数据转换为metrics后再暴露出来。

一个最基础使用`client_golang`包示例如下：

```go
package main

import (
        "net/http"

        "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
        http.Handle("/metrics", promhttp.Handler())
        http.ListenAndServe(":2112", nil)
}
```

`promhttp.Handler()`封装了当前的go进程运行时态的 metrics，并按照metircs后接value的格式在前端输出。

当我们访问2112端口的metrics路径时得到如下数据：

```text
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 7
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.15.14"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
...
```
官方仓库[示例](https://github.com/prometheus/client_golang/tree/main/examples)
