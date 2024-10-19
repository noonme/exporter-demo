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
