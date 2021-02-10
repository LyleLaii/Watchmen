package event

import (
	"fmt"
	v1core "k8s.io/api/core/v1"
	kube_client "k8s.io/client-go/kubernetes"
	client_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	kube_restclient "k8s.io/client-go/rest"
	//"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	//v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	// Rate of refill for the event spam filter in client go
	// 1 per event key per 5 minutes.
	defaultQPS = 1. / 300.
	// Number of events allowed per event key before rate limiting is triggered
	// Has to greater than or equal to 1.
	defaultBurstSize = 1
	// Number of distinct event keys in the rate limiting cache.
	defaultLRUCache = 8192
)

//func CreateEventRecorder(kubeClient clientset.Interface) record.EventRecorder {
func CreateEventRecorder(componentName string, client kube_client.Interface) record.EventRecorder {
	//eventBroadcaster := record.NewBroadcasterWithCorrelatorOptions(getCorrelationOptions())
	eventBroadcaster := record.NewBroadcaster()
	klog.V(3).Infof("create event recorder")
	//if _, isfake := kubeClient.(*fake.Clientset); !isfake {
	//	actualSink := &v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events("")}
	//	// EventBroadcaster has a StartLogging() method but the throttling options from getCorrelationOptions() get applied only to
	//	// actual sinks, which makes it throttle the actual events, but not the corresponding log lines. This leads to massive spam
	//	// in the Cluster Autoscaler log which can eventually fill up a whole disk. As a workaround, event logging is added
	//	// as a wrapper to the actual sink.
	//	// TODO: Do this natively if https://github.com/kubernetes/kubernetes/issues/90168 gets implemented.
	//	sinkWithLogging := WrapEventSinkWithLogging(actualSink)
	//	eventBroadcaster.StartRecordingToSink(sinkWithLogging)
	//}
	eventBroadcaster.StartRecordingToSink(&client_v1.EventSinkImpl{Interface: client_v1.New(client.CoreV1().RESTClient()).Events("")})
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1core.EventSource{Component: componentName})
}

func getCorrelationOptions() record.CorrelatorOptions {
	return record.CorrelatorOptions{
		QPS:          defaultQPS,
		BurstSize:    defaultBurstSize,
		LRUCacheSize: defaultLRUCache,
	}
}

func CreateKubeClient() (kube_client.Interface) {
	var config *kube_restclient.Config
	var err error
	// Load config from Kubernetes well known location.
	config, err = kube_restclient.InClusterConfig()
	if err != nil {
		klog.V(3).Infof(fmt.Sprintf("error connecting to the client: %v", err))
	}
	config.ContentType = "application/vnd.kubernetes.protobuf"
	return kube_client.NewForConfigOrDie(config)
}