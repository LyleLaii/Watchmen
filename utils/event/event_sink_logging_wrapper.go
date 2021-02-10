package event

import (
	clientv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

//
//import (
//clientv1 "k8s.io/api/core/v1"
//"k8s.io/client-go/tools/record"
//klog "k8s.io/klog/v2"
//)

type eventSinkLoggingWrapper struct {
	actualSink record.EventSink
}

// Create wraps EventSink's Create().
func (s eventSinkLoggingWrapper) Create(event *clientv1.Event) (*clientv1.Event, error) {
	logEvent(event)
	return s.actualSink.Create(event)
}

// Update wraps EventSink's Update().
func (s eventSinkLoggingWrapper) Update(event *clientv1.Event) (*clientv1.Event, error) {
	logEvent(event)
	return s.actualSink.Update(event)
}

// Patch wraps EventSink's Patch().
func (s eventSinkLoggingWrapper) Patch(oldEvent *clientv1.Event, data []byte) (*clientv1.Event, error) {
	logEvent(oldEvent)
	return s.actualSink.Patch(oldEvent, data)
}

func logEvent(e *clientv1.Event) {
	klog.V(4).Infof("Event(%#v): type: '%v' reason: '%v' %v", e.InvolvedObject, e.Type, e.Reason, e.Message)
}

// WrapEventSinkWithLogging adds logging each event via klog to an existing event sink.
func WrapEventSinkWithLogging(sink record.EventSink) record.EventSink {
	return eventSinkLoggingWrapper{actualSink: sink}
}
