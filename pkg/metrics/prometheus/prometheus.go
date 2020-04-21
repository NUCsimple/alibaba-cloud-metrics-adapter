package prometheus

import (
	p "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

const (
	HTTP_REQUEST_SECOND = "http_request_second"
	MIN_PERIOD = 60
)

type PrometheusMetricSource struct{}

func (ps *PrometheusMetricSource) GetExternalMetricInfoList() []p.ExternalMetricInfo {
	metricInfoList := make([]p.ExternalMetricInfo, 0)
	var MetricArray = []string{}
	for _, metric := range MetricArray {
		metricInfoList = append(metricInfoList, p.ExternalMetricInfo{Metric: metric})
	}
	return metricInfoList
}

func (ps *PrometheusMetricSource) GetExternalMetric(info p.ExternalMetricInfo, namespace string, requirements labels.Requirements) (values []external_metrics.ExternalMetricValue, err error) {
	switch info.Metric {
	case HTTP_REQUEST_SECOND:
		ps.getPrometheusMetrics(namespace,"http_request_second",HTTP_REQUEST_SECOND,requirements)
	}
	return values, err
}

func NewPrometheusSource() *PrometheusMetricSource {
	return &PrometheusMetricSource{}
}
