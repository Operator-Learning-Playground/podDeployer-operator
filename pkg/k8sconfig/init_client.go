package k8sconfig

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

// initClient 初始化 k8s-client
func initClient(config *rest.Config) kubernetes.Interface {
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// initDynamicClient 初始化 k8s-dynamic-client
func initDynamicClient(config *rest.Config) dynamic.Interface {
	c, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

var (
	ClientSet     kubernetes.Interface
	DynamicClient dynamic.Interface
)

func init() {
	ClientSet = initClient(K8sRestConfig())
	DynamicClient = initDynamicClient(K8sRestConfig())
}
