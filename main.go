package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/utils/prom"
	"github.com/kubernetes/client-go/k8s.io/klog"
	"io/ioutil"
	"net/url"
	"os"
	"runtime"
	"time"
	"fmt"
	"net/http"
	"k8s.io/component-base/logs"
	"k8s.io/client-go/transport"
	log "k8s.io/klog"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	mprom "github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/utils/prom"

	"github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/provider"
	basecmd "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd"
	adaptercfg "github.com/AliyunContainerService/alibaba-cloud-metrics-adapter/pkg/config"
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
	cmd := &AlibabaCloudAdapter{}
	cmd.Flags().AddGoFlagSet(cl)
	cmd.Flags().Parse(os.Args)

	stopCh := make(chan struct{})
	defer close(stopCh)

	setupAlibabaCloudProvider(cmd)
	if err := cmd.Run(stopCh); err != nil {
		log.Fatalf("Failed to run Alibaba Cloud metrics adapter: %v", err)
	}
}

func setupAlibabaCloudProvider(cmd *AlibabaCloudAdapter) {

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

type AlibabaCloudAdapter struct{
	basecmd.AdapterBase

	// PrometheusURL is the URL describing how to connect to Prometheus.  Query parameters configure connection options.
	PrometheusURL string
	// PrometheusAuthInCluster enables using the auth details from the in-cluster kubeconfig to connect to Prometheus
	PrometheusAuthInCluster bool
	// PrometheusAuthConf is the kubeconfig file that contains auth details used to connect to Prometheus
	PrometheusAuthConf string
	// PrometheusCAFile points to the file containing the ca-root for connecting with Prometheus
	PrometheusCAFile string
	// PrometheusTokenFile points to the file that contains the bearer token when connecting with Prometheus
	PrometheusTokenFile string
	// AdapterConfigFile points to the file containing the metrics discovery configuration.
	AdapterConfigFile string
	// MetricsRelistInterval is the interval at which to relist the set of available metrics
	MetricsRelistInterval time.Duration
	// MetricsMaxAge is the period to query available metrics for
	MetricsMaxAge time.Duration

	metricsConfig *adaptercfg.MetricsDiscoveryConfig
}

func (cmd *AlibabaCloudAdapter) addFlags() {
	cmd.Flags().StringVar(&cmd.PrometheusURL, "prometheus-url", cmd.PrometheusURL,
		"URL for connecting to Prometheus.")
	cmd.Flags().BoolVar(&cmd.PrometheusAuthInCluster, "prometheus-auth-incluster", cmd.PrometheusAuthInCluster,
		"use auth details from the in-cluster kubeconfig when connecting to prometheus.")
	cmd.Flags().StringVar(&cmd.PrometheusAuthConf, "prometheus-auth-config", cmd.PrometheusAuthConf,
		"kubeconfig file used to configure auth when connecting to Prometheus.")
	cmd.Flags().StringVar(&cmd.PrometheusCAFile, "prometheus-ca-file", cmd.PrometheusCAFile,
		"Optional CA file to use when connecting with Prometheus")
	cmd.Flags().StringVar(&cmd.PrometheusTokenFile, "prometheus-token-file", cmd.PrometheusTokenFile,
		"Optional file containing the bearer token to use when connecting with Prometheus")
	cmd.Flags().StringVar(&cmd.AdapterConfigFile, "config", cmd.AdapterConfigFile,
		"Configuration file containing details of how to transform between Prometheus metrics "+
			"and custom metrics API resources")
	cmd.Flags().DurationVar(&cmd.MetricsRelistInterval, "metrics-relist-interval", cmd.MetricsRelistInterval, ""+
		"interval at which to re-list the set of all available metrics from Prometheus")
	cmd.Flags().DurationVar(&cmd.MetricsMaxAge, "metrics-max-age", cmd.MetricsMaxAge, ""+
		"period for which to query the set of available metrics from Prometheus")
}
//
func (cmd *AlibabaCloudAdapter)makePromClient() (prom.Client, error){
	baseURL, err := url.Parse(cmd.PrometheusURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Prometheus URL %q: %v", baseURL, err)
	}

	var httpClient *http.Client

	if cmd.PrometheusCAFile != "" {
		prometheusCAClient, err := makePrometheusCAClient(cmd.PrometheusCAFile)
		if err != nil {
			return nil, err
		}
		httpClient = prometheusCAClient
		klog.Info("successfully loaded ca from file")
	} else {
		kubeconfigHTTPClient, err := makeKubeconfigHTTPClient(cmd.PrometheusAuthInCluster, cmd.PrometheusAuthConf)
		if err != nil {
			return nil, err
		}
		httpClient = kubeconfigHTTPClient
		klog.Info("successfully using in-cluster auth")
	}

	if cmd.PrometheusTokenFile != "" {
		data, err := ioutil.ReadFile(cmd.PrometheusTokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read prometheus-token-file: %v", err)
		}
		httpClient.Transport = transport.NewBearerAuthRoundTripper(string(data), httpClient.Transport)
	}

	genericPromClient := prom.NewGenericAPIClient(httpClient, baseURL)
	instrumentedGenericPromClient := mprom.InstrumentGenericAPIClient(genericPromClient, baseURL.String())
	return prom.NewClientForAPI(instrumentedGenericPromClient), nil
}

func makeKubeconfigHTTPClient(inClusterAuth bool, kubeConfigPath string) (*http.Client, error)  {
	// make sure we're not trying to use two different sources of auth
	if inClusterAuth && kubeConfigPath != "" {
		return nil, fmt.Errorf("may not use both in-cluster auth and an explicit kubeconfig at the same time")
	}

	// return the default client if we're using no auth
	if !inClusterAuth && kubeConfigPath == "" {
		return http.DefaultClient, nil
	}

	var authConf *rest.Config
	if kubeConfigPath != "" {
		var err error
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}
		loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
		authConf, err = loader.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to construct  auth configuration from %q for connecting to Prometheus: %v", kubeConfigPath, err)
		}
	} else {
		var err error
		authConf, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to construct in-cluster auth configuration for connecting to Prometheus: %v", err)
		}
	}
	tr, err := rest.TransportFor(authConf)
	if err != nil {
		return nil, fmt.Errorf("unable to construct client transport for connecting to Prometheus: %v", err)
	}
	return &http.Client{Transport: tr}, nil
}

func makePrometheusCAClient(caFilename string) (*http.Client, error) {
	data, err := ioutil.ReadFile(caFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read prometheus-ca-file: %v", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("no certs found in prometheus-ca-file")
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}, nil
}

func (cmd *AlibabaCloudAdapter)loadConfig() error  {
	// load metrics discovery configuration
	if cmd.AdapterConfigFile == "" {
		return fmt.Errorf("no metrics discovery configuration file specified (make sure to use --config)")
	}
	metricsConfig, err := adaptercfg.FromFile(cmd.AdapterConfigFile)
	if err != nil {
		return fmt.Errorf("unable to load metrics discovery configuration: %v", err)
	}

	cmd.metricsConfig = metricsConfig

	return nil
}

