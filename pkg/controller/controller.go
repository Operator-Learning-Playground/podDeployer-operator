package controller

import (
	"context"
	podrestarterv1alpha1 "github.com/myoperator/poddeployer/pkg/apis/podDeployer/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

type PodDeployerController struct {
	client.Client
	DynamicClient dynamic.Interface
}

func NewPodDeployerController(dc dynamic.Interface) *PodDeployerController {
	return &PodDeployerController{DynamicClient: dc}
}

// Reconcile 调协loop
func (r *PodDeployerController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	podDeployer := &podrestarterv1alpha1.Poddeployer{}
	err := r.Get(ctx, req.NamespacedName, podDeployer)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Error("get err: ", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}
	klog.Info(podDeployer)

	// 非填，默认 "apps/v1/deployments"
	if podDeployer.Spec.Type == "" {
		podDeployer.Spec.Type = "apps/v1/deployments"
	}

	if podDeployer.Spec.Type == "apps/v1/deployments" {
		err = r.handleDeployment(ctx, podDeployer)
		if err != nil {
			klog.Error("handler deployment err: ", err)
			return reconcile.Result{}, err
		}

		if len(podDeployer.Spec.DeploymentSpec.Template.Spec.Containers) != 1 || len(podDeployer.Spec.PriorityImages) != 0 {
			klog.Info("do patch image to deployment...")
			klog.Info(SetOtherContainers)
			// 执行patch操作
			for _, container := range SetOtherContainers {
				time.Sleep(time.Second * 15)
				err = patchResource(podDeployer.Name, podDeployer.Namespace, &container, parseGVR(podDeployer.Spec.Type))
				if err != nil {
					klog.Error("patch deployment err: ", err)
					return reconcile.Result{}, err
				}
			}
		}
	} else if podDeployer.Spec.Type == "apps/v1/statefulsets" {
		err = r.handleStatefulSet(ctx, podDeployer)
		if err != nil {
			klog.Error("handler statefulset err: ", err)
			return reconcile.Result{}, err
		}

		if len(podDeployer.Spec.StatefulSetSpec.Template.Spec.Containers) != 1 || len(podDeployer.Spec.PriorityImages) != 0 {
			klog.Info("do patch image to statefulset...")
			klog.Info(SetOtherContainers)
			// 执行patch操作
			for _, container := range SetOtherContainers {
				time.Sleep(time.Second * 15)
				err = patchResource(podDeployer.Name, podDeployer.Namespace, &container, parseGVR(podDeployer.Spec.Type))
				if err != nil {
					klog.Error("patch statefulset err: ", err)
					return reconcile.Result{}, err
				}
			}
		}
	} else {
		err = r.handleDaemonSet(ctx, podDeployer)
		if err != nil {
			klog.Error("handler daemonSet err: ", err)
			return reconcile.Result{}, err
		}

		if len(podDeployer.Spec.DaemonSetSpec.Template.Spec.Containers) != 1 || len(podDeployer.Spec.PriorityImages) != 0 {
			klog.Info("do patch image to daemonSet...")
			klog.Info(SetOtherContainers)
			// 执行patch操作
			for _, container := range SetOtherContainers {
				time.Sleep(time.Second * 15)
				err = patchResource(podDeployer.Name, podDeployer.Namespace, &container, parseGVR(podDeployer.Spec.Type))
				if err != nil {
					klog.Error("patch daemonSet err: ", err)
					return reconcile.Result{}, err
				}
			}
		}
	}
	return reconcile.Result{}, nil
}

// InjectClient 使用controller-runtime 需要注入的client
func (r *PodDeployerController) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

func (r *PodDeployerController) DeploymentDeleteHandler(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == podrestarterv1alpha1.PodDeployerKind && ref.APIVersion == podrestarterv1alpha1.PodDeployerApiVersion {
			// 重新入列
			klog.Infof("deleted deployment object [%v] name [%v]]\n", event.Object.GetName(), event.Object.GetObjectKind().GroupVersionKind().Kind)
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
	}
}

// FIXME: 目前没效果，因为 pod OwnerReferences 没有設置 PodDeployer
func (r *PodDeployerController) PodDeleteHandler(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == podrestarterv1alpha1.PodDeployerKind && ref.APIVersion == podrestarterv1alpha1.PodDeployerApiVersion {
			// 重新入列
			klog.Infof("deleted pod object [%v] name [%v]]\n", event.Object.GetName(), event.Object.GetObjectKind().GroupVersionKind().Kind)
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
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
