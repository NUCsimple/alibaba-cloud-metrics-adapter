package prom

import (
	"github.com/emirpasic/gods/sets/hashset"
	p "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/metrics/pkg/apis/external_metrics"
	"testing"
)

func TestPrometheusSource_AddExternalMetric(t *testing.T) {
	source := &prometheusSource{kubeClient: nil, metricList: hashset.New()}

	t.Run("Register external metric for the prometheus", func(t *testing.T) {
		testMetric := &externalMetric{
			info: p.ExternalMetricInfo{Metric: "test-metrics"},
		}

		source.AddExternalMetric(testMetric)
		if source.metricList.Contains(testMetric) {
			t.Log("Verify passed")
			return
		}
		t.Errorf("It shoule be include testMetric")
	})
}

func TestPrometheusSource_DeleteExternalMetric(t *testing.T) {
	source := &prometheusSource{kubeClient: nil, metricList: hashset.New()}

	t.Run("Delete external metric for the prometheus", func(t *testing.T) {
		testMetric := &externalMetric{
			info: p.ExternalMetricInfo{Metric: "test-metrics"},
		}

		source.AddExternalMetric(testMetric)
		source.DeleteExternalMetric(testMetric)
		if !source.metricList.Contains(testMetric) {
			t.Log("Verify passed")
			return
		}
		t.Errorf("It shoule be not include testMetric")
	})
}

func TestPrometheusSource_GetExternalMetricInfoList(t *testing.T) {
	source := &prometheusSource{kubeClient: nil, metricList: hashset.New()}

	t.Run("Get external metric list", func(t *testing.T) {
		testMetric := &externalMetric{
			info: p.ExternalMetricInfo{Metric: "test-metrics"},
		}
		want := p.ExternalMetricInfo{Metric: "test-metrics"}
		source.AddExternalMetric(testMetric)

		got := source.GetExternalMetricInfoList()[0]
		if got != want {
			t.Error("test metric has been registered external metric.")
		}
		t.Log(got)
	})
}

func TestPrometheusSource_GetExternalMetric(t *testing.T) {
	source := &prometheusSource{kubeClient: nil, metricList: hashset.New()}

	t.Run("Get external metric", func(t *testing.T) {
		labelRequirements := labels.Requirements{}

		testMetric := &externalMetric{
			info: p.ExternalMetricInfo{Metric: "test-metrics"},
			value: external_metrics.ExternalMetricValue{
				MetricName: "test-metrics",
				MetricLabels: map[string]string{
					"foo": "bar",
				},
				Value: *resource.NewQuantity(42, resource.DecimalSI),
			},
		}
		metricInfo := p.ExternalMetricInfo{Metric: "test-metrics"}

		source.AddExternalMetric(testMetric)
		value, err := source.GetExternalMetric(metricInfo, "default", labelRequirements)
		if err != nil {
			t.Error(err)
		}
		t.Log(value)
	})
}
