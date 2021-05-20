package kubernetes

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/metrics/pkg/apis/external_metrics"
	"time"
)

const (
	PROMETHEUS_SERVER      = "prometheus.server"
	PROMETHEUS_QUERY       = "prometheus.query"
	PROMETHEUS_METRIC_NAME = "prometheus.metric.name"
)

func BelongPrometheusSource(hpa *autoscalingv2.HorizontalPodAutoscaler) bool {
	var (
		switch1 bool
		switch2 bool
		switch3 bool
	)
	for key, _ := range hpa.Annotations {
		if key == PROMETHEUS_SERVER {
			switch1 = true
		}
		if key == PROMETHEUS_QUERY {
			switch2 = true
		}
		if key == PROMETHEUS_METRIC_NAME {
			switch3 = true
		}
	}

	return switch1 && switch2 && switch3
}

func GetPrometheusMetricValue(hpa *autoscalingv2.HorizontalPodAutoscaler) (value external_metrics.ExternalMetricValue, err error) {
	var (
		prometheusServer string
		prometheusQuery  string
		metricName       string
	)
	for key, value := range hpa.Annotations {
		if key == PROMETHEUS_SERVER {
			prometheusServer = value
		}
		if key == PROMETHEUS_QUERY {
			prometheusQuery = value
		}
		if key == PROMETHEUS_METRIC_NAME {
			metricName = value
		}
	}

	promClient, err := api.NewClient(api.Config{
		Address: prometheusServer,
	})
	if err != nil {
		klog.Errorf("Error creating client: %v", err)
		return value, err
	}

	v1api := v1.NewAPI(promClient)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//TODO
	klog.Infof("query is %v", prometheusQuery)
	result, warnings, err := v1api.Query(ctx, prometheusQuery, time.Now())
	if err != nil {
		klog.Errorf("Error querying Prometheus: %v", err)
		return value, err
	}

	if len(warnings) > 0 {
		klog.Errorf("Warnings: %v", warnings)
		return value, errors.New("response include warning")
	}

	switch result.Type() {
	case model.ValVector:
		samples := result.(model.Vector)
		if len(samples) == 0 {
			return value, errors.New("")
		}
		sampleValue := samples[0].Value
		value = external_metrics.ExternalMetricValue{
			MetricName: metricName,
			Value:      *resource.NewQuantity(int64(sampleValue), resource.DecimalSI),
			Timestamp:  metav1.Now(),
		}
	case model.ValScalar:
		scalar := result.(*model.Scalar)
		sampleValue := scalar.Value
		value = external_metrics.ExternalMetricValue{
			MetricName: metricName,
			Value:      *resource.NewQuantity(int64(sampleValue), resource.DecimalSI),
			Timestamp:  metav1.Now(),
		}
	}

	return value, nil
}
