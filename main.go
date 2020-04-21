package main

import (
	"alibaba-cloud-metrics-adapter/pkg/utils"
	"flag"
	"os"
	"runtime"

	"k8s.io/component-base/logs"
	log "k8s.io/klog"

	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/provider"
)

func main() {

	logs.InitLogs()
	defer logs.FlushLogs()

	// golang 1.6 or before
	if len(os.Getenv("GOMAXPROvCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cl := flag.NewFlagSet("external-metrics", flag.ErrorHandling(0))
	//log.InitFlags(cl)
	cmd := &utils.AlibabaMetricAdapter{}
	cmd.Flags().AddGoFlagSet(cl)
	cmd.Flags().Parse(os.Args)
	if cmd.PrometheusServer!=""{
		cmd.MakePromClient()
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	setupAlibabaCloudProvider(cmd)
	if err := cmd.Run(stopCh); err != nil {
		log.Fatalf("Failed to run Alibaba Cloud metrics adapter: %v", err)
	}
}

func setupAlibabaCloudProvider(cmd *utils.AlibabaMetricAdapter) {

	mapper, err := cmd.RESTMapper()
	if err != nil {
		log.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	dynamicClient, err := cmd.DynamicClient()
	if err != nil {
		log.Fatalf("unable to construct dynamic k8s client: %v", err)
	}

	metricProvider, err := provider.NewAlibabaCloudProvider(mapper, dynamicClient)
	if err != nil {
		log.Fatal("Failed to setup Alibaba Cloud metrics provider:", err)
	}
	// TODO custom metrics will be supported later after multi custom adapter support.
	//cmd.WithCustomMetrics(metricProvider)
	cmd.WithExternalMetrics(metricProvider)
}

