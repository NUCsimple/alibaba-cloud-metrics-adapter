package utils

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getKubernetesClient()(*kubernetes.Clientset,error) {
	clientConfig, err := rest.InClusterConfig()
	if err!=nil{
		return nil,fmt.Errorf("unable to construct lister client config to initialize provider: %v", err)
	}
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil,fmt.Errorf("failed to initialize new client: %v", err)
	}
	return client,nil
}
