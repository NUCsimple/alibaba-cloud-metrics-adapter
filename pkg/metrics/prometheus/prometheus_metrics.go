package prometheus

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

func (ps *PrometheusMetricSource)getPrometheusMetrics(namespace, metric, externalMetric string, requirements labels.Requirements) (values []external_metrics.ExternalMetricValue, err error)  {

}
