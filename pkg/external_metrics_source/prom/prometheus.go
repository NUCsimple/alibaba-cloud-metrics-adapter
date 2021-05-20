package prom

import (
	"fmt"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/kubernetes"
	"github.com/emirpasic/gods/sets/hashset"
	p "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	"k8s.io/metrics/pkg/apis/external_metrics"
	"time"
)

type externalMetric struct {
	info   p.ExternalMetricInfo
	labels map[string]string
	value  external_metrics.ExternalMetricValue
}

type prometheusSource struct {
	kubeClient dynamic.Interface
	metricList *hashset.Set
}

func NewPrometheusSource(client dynamic.Interface) *prometheusSource {
	ps := &prometheusSource{
		kubeClient: client,
		metricList: hashset.New()}

	go ps.MonitorHPAs()

	return ps
}

func (prom *prometheusSource) GetExternalMetricInfoList() []p.ExternalMetricInfo {
	metricInfoList := make([]p.ExternalMetricInfo, 0)

	for _, metric := range prom.metricList.Values() {
		//TODO log
		klog.Infof("prom metric is %v", metric)
		metricInfoList = append(metricInfoList, metric.(*externalMetric).info)
	}

	return metricInfoList
}

func (prom *prometheusSource) AddExternalMetric(metric *externalMetric) {
	klog.Infof("metric %s is registered as an external metric. ", metric.info)
	prom.metricList.Add(metric)
}

func (prom *prometheusSource) DeleteExternalMetric(metric *externalMetric) {
	klog.Infof("metric %s is deleted from external metric list. ", metric.info)
	prom.metricList.Remove(metric)
}

func (prom *prometheusSource) GetExternalMetric(info p.ExternalMetricInfo, _ string, _ labels.Requirements) (values []external_metrics.ExternalMetricValue, err error) {
	for _, metric := range prom.metricList.Values() {
		//TODO log
		klog.Infof("prom metric is %v", metric)

		if metric.(*externalMetric).info == info {
			//TODO log
			klog.Infof("metric %v is in the external metric list.", metric)
			values = append(values, metric.(*externalMetric).value)
			return values, nil
		}
	}
	return nil, fmt.Errorf("not found metric %s from external metric list", info.Metric)
}

func (prom *prometheusSource) MonitorHPAs() {
	client := kubernetes.NewKubernetesClient()

	for {
		watcher, err := client.AutoscalingV2beta2().HorizontalPodAutoscalers(metav1.NamespaceAll).Watch(metav1.ListOptions{
			Watch: true,
		})
		if err != nil {
			klog.Errorf("Failed to start watch for new HPAs: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()

	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				klog.Infof("HPA watch channel update. watchChanObject: %v", watchUpdate)
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if hpa, ok := watchUpdate.Object.(*autoscalingv2.HorizontalPodAutoscaler); ok {
					switch watchUpdate.Type {
					case kubewatch.Added, kubewatch.Modified:
						if kubernetes.BelongPrometheusSource(hpa) {
							metric := &externalMetric{}
							metricValue, err := kubernetes.GetPrometheusMetricValue(hpa)
							if err != nil {
								klog.Errorf("get metric %s value from prometheus server err: %v", metricValue.MetricName, err)
							}
							metric.info = p.ExternalMetricInfo{
								Metric: metricValue.MetricName,
							}
							metric.value = metricValue
							prom.AddExternalMetric(metric)
						}
					case kubewatch.Deleted:
						if kubernetes.BelongPrometheusSource(hpa) {
							metric := &externalMetric{}
							metricValue, err := kubernetes.GetPrometheusMetricValue(hpa)
							if err != nil {
								klog.Errorf("get metric %s value from prometheus server err: %v", metricValue.MetricName, err)
							}
							metric.info = p.ExternalMetricInfo{
								Metric: metricValue.MetricName,
							}
							metric.value = metricValue
							prom.DeleteExternalMetric(metric)
						}
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			}
		}
	}
}
