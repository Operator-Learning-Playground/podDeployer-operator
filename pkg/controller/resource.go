package controller

import (
	"context"
	"errors"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	podrestarterv1alpha1 "github.com/myoperator/poddeployer/pkg/apis/podDeployer/v1alpha1"
	"github.com/myoperator/poddeployer/pkg/k8sconfig"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
)

var (
	SetOtherContainers = make([]v1.Container, 0)
)

func (r *PodDeployerController) handleDeployment(ctx context.Context, podDeployer *podrestarterv1alpha1.Poddeployer, deployment appsv1.Deployment) error {

	deployment.Name = podDeployer.Name
	deployment.Namespace = podDeployer.Namespace

	mutateDeploymentRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &deployment, func() error {
		// TODO: 处理deployment image的数量
		// 1. 如果image数量为1或没有设置PriorityImages，没有必要做其他操作
		// 2. 如果有设置PriorityImages，需要先遍历PriorityImages的image name 与 container name
		// 只留最高优先集的image，把其他的都取出来，并且使用patch再加入，完成一个调协

		if len(podDeployer.Spec.DeploymentSpec.Template.Spec.Containers) == 1 || len(podDeployer.Spec.PriorityImages) == 0 {
			klog.Info("no need to go priority process...")
			mutateDeployment(podDeployer, &deployment)
		} else {
			klog.Info("need to go priority process...")
			priorityImages := calculatePriorityImages(podDeployer)
			setOtherContainers, err := handleDeploymentImageSort(priorityImages, podDeployer)
			klog.Info("setOtherContainers: ", setOtherContainers)
			SetOtherContainers = setOtherContainers
			if err != nil {
				klog.Error("err: ", err)
				return err
			}
			mutateDeployment(podDeployer, &deployment)

		}

		// 设置ownerReferences
		deployment.OwnerReferences = append(deployment.OwnerReferences, metav1.OwnerReference{
			APIVersion: podDeployer.APIVersion,
			Kind:       podDeployer.Kind,
			Name:       podDeployer.Name,
			UID:        podDeployer.UID,
		})

		return nil
	})
	if err != nil {

		return err
	}

	klog.Info("CreateOrUpdate", "Deployment", mutateDeploymentRes)
	return nil
}

func mutateDeployment(podDeployer *podrestarterv1alpha1.Poddeployer, deployment *appsv1.Deployment) {
	deployment.Spec = podDeployer.Spec.DeploymentSpec
	labels := map[string]string{
		"podDeployer": podDeployer.Name,
	}
	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}
	deployment.Spec.Template.Labels = map[string]string{}
	deployment.Spec.Selector = &selector
	deployment.Spec.Template.Labels["podDeployer"] = podDeployer.Name
}

// 为image排序
func calculatePriorityImages(podDeployer *podrestarterv1alpha1.Poddeployer) []podrestarterv1alpha1.PriorityImage {
	// 找出对应的
	imageList := podDeployer.Spec.PriorityImages
	// 返回第一个就是权重最高的容器镜像
	sort.SliceStable(imageList, func(i, j int) bool {
		return imageList[i].Value > imageList[j].Value
	})
	return imageList
}

// 替换image，并返回剩下排序后的images
func handleDeploymentImageSort(priorityImages []podrestarterv1alpha1.PriorityImage, podDeployer *podrestarterv1alpha1.Poddeployer) ([]v1.Container, error) {
	if len(priorityImages) != len(podDeployer.Spec.DeploymentSpec.Template.Spec.Containers) {
		return nil, errors.New("priorityImage len error")
	}
	firstContainer := make([]v1.Container, 0)
	setOtherContainers := make([]v1.Container, 0)

	for k, v := range priorityImages {
		for _, container := range podDeployer.Spec.DeploymentSpec.Template.Spec.Containers {
			// 第一个放进来
			if v.Image == container.Name && k == 0 {
				firstContainer = append(firstContainer, container)
			} else if v.Image == container.Name {
				setOtherContainers = append(setOtherContainers, container)
			}
		}
	}
	podDeployer.Spec.DeploymentSpec.Template.Spec.Containers = firstContainer
	return setOtherContainers, nil
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// patchDeployment 使用deployment patch的方式顺序执行pod
func patchDeployment(deploymentName, namespace string, container *v1.Container) {
	klog.Info("do deployment patch....")

	pa := make([]patchOperation, 0)
	p := patchOperation{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/template/spec/containers/-"),
		Value: container,
	}
	pa = append(pa, p)
	patchBytes, err := json.Marshal(&pa)
	if err != nil {
		klog.Error(err)
		return
	}

	jsonPatch, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		klog.Error("DecodePatch error: ", err)
		return
	}
	jsonPatchBytes, err := json.Marshal(jsonPatch)
	if err != nil {
		klog.Error("json Marshal error: ", err)
		return
	}
	klog.Info(string(jsonPatchBytes))
	_, err = k8sconfig.ClientSet.AppsV1().Deployments(namespace).
		Patch(context.TODO(), deploymentName, types.JSONPatchType,
			jsonPatchBytes, metav1.PatchOptions{})
	if err != nil {
		klog.Error("patch error: ", err)
		return
	}
}
