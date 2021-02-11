package watchmen

import (
	"k8s.io/klog"
	"strconv"
	"strings"
	"sync"
	"time"
	"watchmen/utils/duration"
)

const (
	TimeRangeDays     = "3d"
	TimeStep          = "5m"
	UsedLimitRatio    = 0.8
	RangeRatio        = 0.8
	PredictDuration   = "24h"
	PredictLimitRatio = 0.8
	MetricUpdateTime  = "5m"
)


const NodeMemUsed = "node_memory_MemTotal_bytes{instance=~\"%s.*\"} - node_memory_MemAvailable_bytes{instance=~\"%s.*\"}"
const NodesMemUsedPercent = "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes"
const NodeMemUsedPercent = "(node_memory_MemTotal_bytes{instance=~\"%s.*\"} - node_memory_MemAvailable_bytes{instance=~\"%s.*\"}) / node_memory_MemTotal_bytes{instance=~\"%s.*\"}"

const MemUsedPercent = "memusedper"

type MetricConfig struct {
	TimeRangeDays     duration.Duration `json:"time_range_days,omitempty"`
	TimeStep          duration.Duration `json:"time_step,omitempty"`
	UsedLimitRatio    float64           `json:"used_limit_ratio,omitempty"`
	RangeRatio        float64           `json:"range_ratio,omitempty"`
	PredictDuration   duration.Duration `json:"predict_duration,omitempty"`
	PredictLimitRatio float64           `json:"predict_limit_ratio,omitempty"`
	MetricUpdateTime  duration.Duration `json:"metric_update_time,omitempty"`
}

func genMetricConfig(m MetricConfig) *MetricConfig {
	zeroDuration := duration.Duration{Duration: time.Duration(0)}

	if m.TimeRangeDays == zeroDuration {
		m.TimeRangeDays, _ = duration.ParseDurationN(TimeRangeDays)
	}
	if m.TimeRangeDays == zeroDuration {
		m.TimeRangeDays, _ = duration.ParseDurationN(TimeRangeDays)
	}
	if m.TimeStep == zeroDuration {
		m.TimeStep, _ = duration.ParseDurationN(TimeStep)
	}
	if m.UsedLimitRatio == 0 {
		m.UsedLimitRatio = UsedLimitRatio
	}
	if m.RangeRatio == 0 {
		m.RangeRatio = RangeRatio
	}
	if m.MetricUpdateTime == zeroDuration {
		m.MetricUpdateTime, _ = duration.ParseDurationN(MetricUpdateTime)
	}
	if m.PredictDuration == zeroDuration {
		m.PredictDuration, _ = duration.ParseDurationN(PredictDuration)
	}

	if m.PredictLimitRatio == 0 {
		m.PredictLimitRatio = PredictLimitRatio
	}

	return &m
}

type PromData struct {
	date time.Time
	metricsDatas map[string]nodesMetrics
	m sync.RWMutex
}

type Metrics struct {
	labels map[string]string
	values [][]interface{}
}


type nodesMetrics []Metrics

func (n nodesMetrics) GetNodeMetrics(nodeIP string) [][]interface{} {
	for _, v := range n {
		ipInfo := v.labels["instance"]
		if strings.Split(ipInfo, ":")[0] == nodeIP {
			return v.values
		}
	}
	return nil
}

func (n nodesMetrics) GetNodeMetricsValues(nodeIP string) (data []float64) {
	vs := n.GetNodeMetrics(nodeIP)
	for _, v := range vs {
		vv, _ := strconv.ParseFloat(v[1].(string), 64)
		data = append(data, vv)
	}
	return
}

func (n nodesMetrics) GetNodeTwoDimeData(nodeIP string) (x []float64, y []float64) {
	vs := n.GetNodeMetrics(nodeIP)
	for _, v := range vs {
		x = append(x, v[0].(float64))
		vv, _ := strconv.ParseFloat(v[1].(string), 64)
		y = append(y, vv)
	}
	return
}

func (w *Watchmen) getNodesMemUsedPercent() (nodesData nodesMetrics) {
	data, err := w.promClient.FetchRangeData(NodesMemUsedPercent, w.metricConf.TimeRangeDays.TimeDuration(), w.metricConf.TimeStep.String())
	if err != nil {
		klog.Errorf("fetchPrometheusData get error: %s", err)
	}
	klog.V(5).Infof("getNodesMemUsedPercent data: %+v", data)
	result := data.Data.Result
	dr := result
	for _, r := range dr {
		nodesData = append(nodesData, Metrics{
			labels: r.Label,
			values: r.Metrics,
		})
	}
	return
}

func (w *Watchmen) nodeMemOverused(nodIP string) bool {
	d := w.promData.metricsDatas[MemUsedPercent].GetNodeMetricsValues(nodIP)
	return calculateCoverage(d, w.metricConf.UsedLimitRatio) > w.metricConf.RangeRatio
}

func (w *Watchmen) nodeOverusedMemLinearPredict(nodeIP string, duration float64) bool {
	x, y := w.promData.metricsDatas[MemUsedPercent].GetNodeTwoDimeData(nodeIP)
	klog.V(5).Infof("nodeOverusedMemLinearPredict: x data is %+v", x)
	klog.V(5).Infof("nodeOverusedMemLinearPredict: y data is %+v", y)
	now := float64(time.Now().Unix())
	klog.V(5).Infof("nodeOverusedMemLinearPredict: offset data is %+v", now)
	k, d := driftXLinearRegression(x, y, now)
	klog.V(5).Infof("nodeOverusedMemLinearPredict: slope %+v intercept %+v", k, d)
	return k*duration + d > w.metricConf.PredictLimitRatio
}