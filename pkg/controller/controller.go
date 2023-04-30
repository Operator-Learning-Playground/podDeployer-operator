package controller

import (
	"context"
	"fmt"
	podrestarterv1alpha1 "github.com/myoperator/poddeployer/pkg/apis/podDeployer/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

type PodDeployerController struct {
	client.Client
}

func NewPodDeployerController() *PodDeployerController {
	return &PodDeployerController{}
}

// Reconcile 调协loop
func (r *PodDeployerController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	podDeployer := &podrestarterv1alpha1.Poddeployer{}
	var deployment appsv1.Deployment
	deployment.Name = podDeployer.Name
	deployment.Namespace = podDeployer.Namespace

	err := r.Get(ctx, req.NamespacedName, podDeployer)
	if err != nil {
		klog.Error("get err: ", err)
		if client.IgnoreNotFound(err) != nil {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{Requeue: true}, err
	}
	klog.Info(podDeployer)

	err = r.handleDeployment(ctx, podDeployer, deployment)
	if err != nil {
		klog.Error("handler deployment err: ", err)
		return reconcile.Result{}, nil
	}

	if len(podDeployer.Spec.DeploymentSpec.Template.Spec.Containers) != 1 || len(podDeployer.Spec.PriorityImages) != 0 {
		klog.Info("do patch image to deployment...")
		klog.Info(SetOtherContainers)
		// 执行patch操作
		for _, container := range SetOtherContainers {
			time.Sleep(time.Second * 15)
			patchDeployment(podDeployer.Name, podDeployer.Namespace, &container)
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
		if ref.Kind == podrestarterv1alpha1.PodDeployerApiVersion && ref.APIVersion == podrestarterv1alpha1.PodDeployerApiVersion {
			// 重新入列，这样删除pod后，就会进入调和loop，发现ownerReference还在，会立即创建出新的pod。
			fmt.Println("被删除的对象名称是", event.Object.GetName(), event.Object.GetObjectKind())
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
	}
}
