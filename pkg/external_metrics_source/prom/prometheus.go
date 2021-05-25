package prom

import (
	"fmt"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/kubernetes"
	p "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
	"k8s.io/metrics/pkg/apis/external_metrics"
	"time"
)

type externalMetric struct {
	labels map[string]string
	value  external_metrics.ExternalMetricValue
}

type prometheusSource struct {
	prometheusUrl string
	metricList    map[string]*externalMetric
}

func NewPrometheusSource(url string) *prometheusSource {
	ps := &prometheusSource{
		prometheusUrl: url,
		metricList:    make(map[string]*externalMetric)}

	go ps.MonitorHPAs()

	return ps
}

func (prom *prometheusSource) GetExternalMetricInfoList() []p.ExternalMetricInfo {
	metricInfoList := make([]p.ExternalMetricInfo, 0)

	for metric, _ := range prom.metricList {
		metricInfo := p.ExternalMetricInfo{
			Metric: metric,
		}
		metricInfoList = append(metricInfoList, metricInfo)
	}

	return metricInfoList
}

func (prom *prometheusSource) AddExternalMetric(metricName string, metric *externalMetric) {
	if _, ok := prom.metricList[metricName]; ok {
		klog.Warningf("metric %s has been registered as an external metric. ", metricName)
	}

	klog.Infof("metric %s is registered as an external metric. ", metricName)
	prom.metricList[metricName] = metric
}

func (prom *prometheusSource) DeleteExternalMetric(metricName string) {
	klog.Infof("metric %s is delete from external metric list. ", metricName)
	delete(prom.metricList, metricName)
}

func (prom *prometheusSource) GetExternalMetric(info p.ExternalMetricInfo, _ string, _ labels.Requirements) (values []external_metrics.ExternalMetricValue, err error) {
	for metric, _ := range prom.metricList {
		if metric == info.Metric {
			values = append(values, prom.metricList[metric].value)
			return values, nil
		}
	}

	return nil, fmt.Errorf("not found metric %s from metric list", info.Metric)
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
						if kubernetes.HasSpecificAnnotation(hpa) {
							metric := &externalMetric{}
							metricValue, err := kubernetes.GetPrometheusValue(hpa, prom.prometheusUrl)
							if err != nil {
								klog.Errorf("failed to get value from prometheus server: %v", err)
								continue
							}
							metric.value = metricValue
							prom.AddExternalMetric(metricValue.MetricName, metric)
						}
					case kubewatch.Deleted:
						if kubernetes.HasSpecificAnnotation(hpa) {
							metricValue, err := kubernetes.GetPrometheusValue(hpa, prom.prometheusUrl)
							if err != nil {
								klog.Errorf("get metric %s value from prometheus server err: %v", metricValue.MetricName, err)
								continue
							}
							prom.DeleteExternalMetric(metricValue.MetricName)
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
