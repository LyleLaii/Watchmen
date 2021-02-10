package watchmen

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"time"
	pc "watchmen/pkg/promclient"
	"watchmen/utils/event"
)

// Plug Name
const Name = "Watchmen"

type Args struct {
	PromConfig  pc.PromConfig `json:"prometheus_config,omitempty"`
	MetricConf  MetricConfig `json:"metric_config,omitempty"`
}

// TODO: Put all data/config in this watchmen struct, is it a good way?  Find a better way
type Watchmen struct {
	handle framework.FrameworkHandle
	promClient *pc.PromClient
	eventRecorder record.EventRecorder
	metricConf *MetricConfig
	promData *PromData
}

var _ framework.PreFilterPlugin = &Watchmen{}
var _ framework.FilterPlugin = &Watchmen{}

func (w *Watchmen) Name() string {
	return Name
}

func (w *Watchmen) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	//klog.V(3).Infof("prefilter pod: %v", pod.Name)
	w.eventRecorder.Event(pod, v1.EventTypeNormal, "PreFilter", "Prepare Prometheus Data")
	now := time.Now()
	klog.V(5).Info(fmt.Printf("Previous Date is %s", w.promData.date.Format(time.RFC3339)))
	if now.Sub(w.promData.date).Seconds() > w.metricConf.MetricUpdateTime {
		klog.V(3).Infof("Now fetch prometheus data cause have not acquired or is outdated")
		w.promData.m.Lock()
		defer w.promData.m.Unlock()
		w.promData.date = now
		w.promData.metricsDatas[MemUsedPercent] = w.getNodesMemUsedPercent()
	} else {
		klog.V(3).Infof("No need to fetch data")
	}

	return nil
}

// PreFilterExtensions returns a PreFilterExtensions interface if the plugin implements one.
func (w *Watchmen) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (w *Watchmen) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *schedulernodeinfo.NodeInfo) *framework.Status {
	//klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, nodeInfo.Node().Name)
	nodeAddress := nodeInfo.Node().Status.Addresses
	var nodeIP string
	for _, v := range nodeAddress{
		if v.Type == "InternalIP" {
			nodeIP = v.Address
		}
	}
	w.promData.m.RLock()
	defer w.promData.m.RUnlock()
	if w.nodeMemOverused(nodeIP) {
		//w.eventRecorder.Event(pod, v1.EventTypeNormal, "Filter", "Unschedulable")
		return framework.NewStatus(framework.Unschedulable, "ResourceLack")
	}
	if w.nodeOverusedMemLinearPredict(nodeIP, w.metricConf.PredictDuration) {
		return framework.NewStatus(framework.Unschedulable, "PredictResourceLack")
	}
	// w.eventRecorder.Event(pod, v1.EventTypeNormal, "Filter", "Success")
	return framework.NewStatus(framework.Success, "Filter success")
}

//type PluginFactory = func(configuration *runtime.Unknown, f FrameworkHandle) (Plugin, error)
func New(configuration *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(configuration, args); err != nil {
		klog.Errorf("decode plugin config get error: %s", err)
		return nil, err
	}
	klog.V(3).Infof("get plugin config args: %+v", args)
	prom, err := pc.New(args.PromConfig)
	if err != nil {
		klog.Errorf("create plugin  get error: %s", err)
		return &Watchmen{}, err
	}
	return &Watchmen{
		handle: f,
		promClient: prom,
		eventRecorder: event.CreateEventRecorder(Name, event.CreateKubeClient()),
		metricConf: genMetricConfig(args.MetricConf),
		promData: &PromData{
			date:         time.Time{},
			metricsDatas: make(map[string]nodesMetrics),
		},
	}, nil
}
