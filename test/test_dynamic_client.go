package main

import (
	"context"
	"github.com/myoperator/poddeployer/pkg/k8sconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"strings"
)

func main() {

	//_, err := k8sconfig.DynamicClient.Resource(parseGVR("apps/v1/statefulsets")).Patch(context.TODO(), "mypoddeployer",
	//	types.JSONPatchType, jsonPatchBytes, metav1.PatchOptions{})

	_, err := k8sconfig.DynamicClient.Resource(parseGVR("apps/v1/deployments")).Namespace("default").Get(context.TODO(), "mypoddeployer-2", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

// parseGVR 解析并指定资源对象 "apps/v1/deployments" "core/v1/pods" "batch/v1/jobs"
func parseGVR(gvr string) schema.GroupVersionResource {
	var group, version, resource string
	gvList := strings.Split(gvr, "/")

	// 防止越界
	if len(gvList) < 2 {
		panic("gvr input error, please input like format apps/v1/deployments or core/v1/pods")
	}

	if len(gvList) < 3 {
		group = ""
		version = gvList[0]
		resource = gvList[1]
	} else {
		if gvList[0] == "core" {
			gvList[0] = ""
		}
		group, version, resource = gvList[0], gvList[1], gvList[2]
	}

	return schema.GroupVersionResource{
		Group: group, Version: version, Resource: resource,
	}
}
