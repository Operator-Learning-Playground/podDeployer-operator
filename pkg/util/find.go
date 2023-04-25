package util

import (
	"context"
	"github.com/myoperator/poddeployer/pkg/k8sconfig"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// GetPodsByDeployment 根据传入Deployment获取当前"正在"使用的pod
func GetPodsByDeployment(depName, ns string) []v1.Pod {

	deployment, err := k8sconfig.ClientSet.AppsV1().Deployments(ns).Get(context.TODO(),
		depName, metav1.GetOptions{})
	if err != nil {
		klog.Error("create clientSet error: ", err)
		return nil
	}
	rsIdList := getRsIdsByDeployment(deployment, k8sconfig.ClientSet)
	podsList := make([]v1.Pod, 0)
	for _, rs := range rsIdList {
		pods := getPodsByReplicaSet(rs, k8sconfig.ClientSet, ns)
		podsList = append(podsList, pods...)
	}

	return podsList
}

// getPodsByReplicaSet 根据传入的ReplicaSet查询到需要的pod
func getPodsByReplicaSet(rs appv1.ReplicaSet, clientSet kubernetes.Interface, ns string) []v1.Pod {
	pods, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Error("list pod error: ", err)
		return nil
	}

	ret := make([]v1.Pod, 0)
	for _, p := range pods.Items {
		// 找到 pod OwnerReferences uid相同的
		if p.OwnerReferences != nil && len(p.OwnerReferences) == 1 {
			if p.OwnerReferences[0].UID == rs.UID {
				ret = append(ret, p)
			}
		}
	}
	return ret

}

// getRsIdsByDeployment 根据传入的dep，获取到相关连的rs列表(滚更后的ReplicaSet就没用了)
func getRsIdsByDeployment(dep *appv1.Deployment, clientSet kubernetes.Interface) []appv1.ReplicaSet {
	// 需要使用match labels过滤
	rsList, err := clientSet.AppsV1().ReplicaSets(dep.Namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(dep.Spec.Selector.MatchLabels).String(),
		})
	if err != nil {
		klog.Error("list ReplicaSets error: ", err)
		return nil
	}

	ret := make([]appv1.ReplicaSet, 0)
	for _, rs := range rsList.Items {
		ret = append(ret, rs)
	}
	return ret
}

