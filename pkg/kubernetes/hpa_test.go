package kubernetes

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"testing"
)

func TestBelongPrometheusSource(t *testing.T) {
	var HPA1, HPA2 autoscalingv2.HorizontalPodAutoscaler
	HPA1.Annotations = make(map[string]string)
	HPA1.Annotations[PROMETHEUS_SERVER] = "http://localhost:9090/metrics"
	HPA1.Annotations[PROMETHEUS_QUERY] = "up"
	HPA1.Annotations[PROMETHEUS_METRIC_NAME] = "test-metric"

	HPA2.Annotations = make(map[string]string)
	HPA2.Annotations[PROMETHEUS_SERVER] = "http://localhost:9090/metrics"
	HPA2.Annotations[PROMETHEUS_QUERY] = "up"

	tests := []struct {
		name string
		args autoscalingv2.HorizontalPodAutoscaler
		want bool
	}{
		{name: "hpa1",
			args: HPA1,
			want: true},
		{name: "hpa2",
			args: HPA2,
			want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BelongPrometheusSource(&tt.args); got != tt.want {
				t.Errorf("BelongPrometheusSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPrometheusMetricValue(t *testing.T) {
	var HPA1 autoscalingv2.HorizontalPodAutoscaler
	HPA1.Annotations = make(map[string]string)
	HPA1.Annotations[PROMETHEUS_SERVER] = "http://prom.cab2c279341ed4ddb9e88fbee1bc94e2e.cn-chengdu.alicontainer.com"
	HPA1.Annotations[PROMETHEUS_QUERY] = "http_requests_total"
	HPA1.Annotations[PROMETHEUS_METRIC_NAME] = "test-metric"

	value, err := GetPrometheusMetricValue(&HPA1)
	if err != nil {
		t.Error(err)
	}

	t.Log(value.Value)
}
