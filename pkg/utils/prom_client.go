package utils

import (
	"k8s.io/client-go/kubernetes"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/api"
	"net/http"
)
type PrometheusClient struct {
	promAPI promv1.API
	client  kubernetes.Interface
}
func (cmd *AlibabaMetricAdapter)MakePromClient(prometheusServer string)(*PrometheusClient,error)  {
	client, err := getKubernetesClient()
	if err!=nil{
	}

	cfg := api.Config{
		Address:      prometheusServer,
		RoundTripper: http.DefaultTransport,
	}

	promClient, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &PrometheusClient{
		client:  client,
		promAPI: promv1.NewAPI(promClient),
	}, nil
}
