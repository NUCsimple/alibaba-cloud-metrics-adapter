package utils

import (
	basecmd "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd"
)

type AlibabaMetricAdapter struct {
	basecmd.AdapterBase
	PrometheusServer string
}

func (cmd *AlibabaMetricAdapter) addFlags() {
	cmd.Flags().StringVar(&cmd.PrometheusServer, "prometheus-server", cmd.PrometheusServer, ""+
		"url of prometheus server to query")
}
