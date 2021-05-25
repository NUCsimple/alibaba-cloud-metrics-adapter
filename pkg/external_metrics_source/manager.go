package external_metrics_source

import (
	"fmt"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/external_metrics_source/ahas"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/external_metrics_source/cms"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/external_metrics_source/prom"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/external_metrics_source/slb"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/external_metrics_source/sls"
	p "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	log "k8s.io/klog"
	"k8s.io/metrics/pkg/apis/external_metrics"
	"time"
)

// RegisterMetricsSource add external metrics source
func (em *ExternalMetricsManager) RegisterMetricsSource() {
	em.register(sls.NewSLSMetricSource())
	em.register(slb.NewSLBMetricSource())
	em.register(cms.NewCMSMetricSource())
	em.register(ahas.NewAHASSentinelMetricSource())

	prometheusSource := prom.NewPrometheusSource(em.kubeClient)

	go func() {
		for {
			em.register(prometheusSource)
			time.Sleep(30 * time.Second)
		}
	}()
}

func NewExternalMetricsManager(client dynamic.Interface) *ExternalMetricsManager {
	return &ExternalMetricsManager{
		metricsSource: make(map[p.ExternalMetricInfo]MetricSource),
		kubeClient:    client,
	}
}

func (em *ExternalMetricsManager) register(m MetricSource) {
	em.AddMetricsSource(m)
}

type MetricSource interface {
	GetExternalMetricInfoList() []p.ExternalMetricInfo
	GetExternalMetric(info p.ExternalMetricInfo, namespace string, requirements labels.Requirements) ([]external_metrics.ExternalMetricValue, error)
}

type ExternalMetricsManager struct {
	kubeClient    dynamic.Interface
	metricsSource map[p.ExternalMetricInfo]MetricSource
}

func (em *ExternalMetricsManager) AddMetricsSource(m MetricSource) {
	metricInfoList := m.GetExternalMetricInfoList()
	for _, p := range metricInfoList {
		log.Infof("Register metric: %v to external metrics manager\n", p)
		em.metricsSource[p] = m
	}
}

func (em *ExternalMetricsManager) GetMetricsInfoList() []p.ExternalMetricInfo {
	metricsInfoList := make([]p.ExternalMetricInfo, 0)

	for metricInfo, _ := range em.metricsSource {
		metricsInfoList = append(metricsInfoList, metricInfo)
	}
	return metricsInfoList
}

func (em *ExternalMetricsManager) GetExternalMetrics(namespace string, requirements labels.Requirements, info p.ExternalMetricInfo) ([]external_metrics.ExternalMetricValue, error) {
	if source, ok := em.metricsSource[info]; ok {
		return source.GetExternalMetric(info, namespace, requirements)
	}

	return nil, fmt.Errorf("The specific metric %s is not found.\n", info.Metric)
}
