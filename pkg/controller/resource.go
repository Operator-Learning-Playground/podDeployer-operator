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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
)

var (
	SetOtherContainers = make([]v1.Container, 0)
)

// handleDeployment 处理顺序的主要逻辑：使用依赖第一个容器创建 deployment，
// 再按照顺序 patch 其他容器
func (r *PodDeployerController) handleDeployment(ctx context.Context, podDeployer *podrestarterv1alpha1.Poddeployer) error {
	var deployment appsv1.Deployment
	deployment.Name = podDeployer.Name
	deployment.Namespace = podDeployer.Namespace
	// 创建 deployment 对象
	mutateDeploymentRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &deployment, func() error {
		// TODO: 处理 image 的数量
		// 1. 如果 image 数量为 1 或没有设置 PriorityImages ，没有必要做其他操作
		// 2. 如果有设置 PriorityImages，需要先遍历 PriorityImages的 image name 与 container name
		// 只留最高优先集的 image，把其他的都取出来，并且使用 patch 再加入，完成一个调协
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

		// 设置 ownerReferences
		c := true
		deployment.OwnerReferences = append(deployment.OwnerReferences, metav1.OwnerReference{
			APIVersion: podDeployer.APIVersion,
			Kind:       podDeployer.Kind,
			Name:       podDeployer.Name,
			UID:        podDeployer.UID,
			Controller: &c,
		})
		return nil
	})
	if err != nil {
		return err
	}

	klog.Info("CreateOrUpdate", "Deployment", mutateDeploymentRes)
	return nil
}

func (r *PodDeployerController) handleStatefulSet(ctx context.Context, podDeployer *podrestarterv1alpha1.Poddeployer) error {
	var statefulSet appsv1.StatefulSet
	statefulSet.Name = podDeployer.Name
	statefulSet.Namespace = podDeployer.Namespace
	// 创建 statefulSet 对象
	mutateStatefulSetRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &statefulSet, func() error {
		// TODO: 处理 image 的数量
		// 1. 如果 image 数量为 1 或没有设置 PriorityImages ，没有必要做其他操作
		// 2. 如果有设置 PriorityImages，需要先遍历 PriorityImages的 image name 与 container name
		// 只留最高优先集的 image，把其他的都取出来，并且使用 patch 再加入，完成一个调协
		if len(podDeployer.Spec.StatefulSetSpec.Template.Spec.Containers) == 1 || len(podDeployer.Spec.PriorityImages) == 0 {
			klog.Info("no need to go priority process...")
			mutateStatefulSet(podDeployer, &statefulSet)
		} else {
			klog.Info("need to go priority process...")
			priorityImages := calculatePriorityImages(podDeployer)
			setOtherContainers, err := handleStatefulSetImageSort(priorityImages, podDeployer)
			klog.Info("setOtherContainers: ", setOtherContainers)
			SetOtherContainers = setOtherContainers
			if err != nil {
				klog.Error("err: ", err)
				return err
			}
			mutateStatefulSet(podDeployer, &statefulSet)
		}

		// 设置 ownerReferences
		c := true
		statefulSet.OwnerReferences = append(statefulSet.OwnerReferences, metav1.OwnerReference{
			APIVersion: podDeployer.APIVersion,
			Kind:       podDeployer.Kind,
			Name:       podDeployer.Name,
			UID:        podDeployer.UID,
			Controller: &c,
		})
		return nil
	})
	if err != nil {
		return err
	}

	klog.Info("CreateOrUpdate", "StatefulSet", mutateStatefulSetRes)
	return nil
}

func (r *PodDeployerController) handleDaemonSet(ctx context.Context, podDeployer *podrestarterv1alpha1.Poddeployer) error {
	var daemonSet appsv1.DaemonSet
	daemonSet.Name = podDeployer.Name
	daemonSet.Namespace = podDeployer.Namespace
	// 创建 deployment 对象
	mutateDaemonSetRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &daemonSet, func() error {
		// TODO: 处理 image 的数量
		// 1. 如果 image 数量为 1 或没有设置 PriorityImages ，没有必要做其他操作
		// 2. 如果有设置 PriorityImages，需要先遍历 PriorityImages的 image name 与 container name
		// 只留最高优先集的 image，把其他的都取出来，并且使用 patch 再加入，完成一个调协
		if len(podDeployer.Spec.DaemonSetSpec.Template.Spec.Containers) == 1 || len(podDeployer.Spec.PriorityImages) == 0 {
			klog.Info("no need to go priority process...")
			mutateDaemonSet(podDeployer, &daemonSet)
		} else {
			klog.Info("need to go priority process...")
			priorityImages := calculatePriorityImages(podDeployer)
			setOtherContainers, err := handleDaemonSetImageSort(priorityImages, podDeployer)
			klog.Info("setOtherContainers: ", setOtherContainers)
			SetOtherContainers = setOtherContainers
			if err != nil {
				klog.Error("err: ", err)
				return err
			}
			mutateDaemonSet(podDeployer, &daemonSet)
		}

		// 设置 ownerReferences
		c := true
		daemonSet.OwnerReferences = append(daemonSet.OwnerReferences, metav1.OwnerReference{
			APIVersion: podDeployer.APIVersion,
			Kind:       podDeployer.Kind,
			Name:       podDeployer.Name,
			UID:        podDeployer.UID,
			Controller: &c,
		})
		return nil
	})
	if err != nil {
		return err
	}
	klog.Info("CreateOrUpdate", "DaemonSet", mutateDaemonSetRes)
	return nil
}

// mutateStatefulSet 设置 labels 标签，目前无其他作用
func mutateStatefulSet(podDeployer *podrestarterv1alpha1.Poddeployer, statefulSet *appsv1.StatefulSet) {
	statefulSet.Spec = podDeployer.Spec.StatefulSetSpec
	labels := map[string]string{
		"podDeployer": podDeployer.Name,
	}
	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}
	statefulSet.Spec.Template.Labels = map[string]string{}
	statefulSet.Spec.Selector = &selector
	statefulSet.Spec.Template.Labels["podDeployer"] = podDeployer.Name
}

// mutateStatefulSet 设置 labels 标签，目前无其他作用
func mutateDaemonSet(podDeployer *podrestarterv1alpha1.Poddeployer, daemonSet *appsv1.DaemonSet) {
	daemonSet.Spec = podDeployer.Spec.DaemonSetSpec
	labels := map[string]string{
		"podDeployer": podDeployer.Name,
	}
	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}
	daemonSet.Spec.Template.Labels = map[string]string{}
	daemonSet.Spec.Selector = &selector
	daemonSet.Spec.Template.Labels["podDeployer"] = podDeployer.Name
}

// mutateDeployment 设置 labels 标签，目前无其他作用
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

// calculatePriorityImages 为 image 排序
func calculatePriorityImages(podDeployer *podrestarterv1alpha1.Poddeployer) []podrestarterv1alpha1.PriorityImage {
	// 找出对应的
	imageList := podDeployer.Spec.PriorityImages
	// 返回第一个就是权重最高的容器镜像
	sort.SliceStable(imageList, func(i, j int) bool {
		return imageList[i].Value > imageList[j].Value
	})
	return imageList
}

// handleDeploymentImageSort 替换 image，并返回剩下排序后的 images
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

// handleStatefulSetImageSort 替换 image，并返回剩下排序后的 images
func handleStatefulSetImageSort(priorityImages []podrestarterv1alpha1.PriorityImage, podDeployer *podrestarterv1alpha1.Poddeployer) ([]v1.Container, error) {
	if len(priorityImages) != len(podDeployer.Spec.StatefulSetSpec.Template.Spec.Containers) {
		return nil, errors.New("priorityImage len error")
	}
	firstContainer := make([]v1.Container, 0)
	setOtherContainers := make([]v1.Container, 0)

	for k, v := range priorityImages {
		for _, container := range podDeployer.Spec.StatefulSetSpec.Template.Spec.Containers {
			// 第一个放进来
			if v.Image == container.Name && k == 0 {
				firstContainer = append(firstContainer, container)
			} else if v.Image == container.Name {
				setOtherContainers = append(setOtherContainers, container)
			}
		}
	}
	podDeployer.Spec.StatefulSetSpec.Template.Spec.Containers = firstContainer
	return setOtherContainers, nil
}

// handleDeploymentImageSort 替换 image，并返回剩下排序后的 images
func handleDaemonSetImageSort(priorityImages []podrestarterv1alpha1.PriorityImage, podDeployer *podrestarterv1alpha1.Poddeployer) ([]v1.Container, error) {
	if len(priorityImages) != len(podDeployer.Spec.DaemonSetSpec.Template.Spec.Containers) {
		return nil, errors.New("priorityImage len error")
	}
	firstContainer := make([]v1.Container, 0)
	setOtherContainers := make([]v1.Container, 0)

	for k, v := range priorityImages {
		for _, container := range podDeployer.Spec.DaemonSetSpec.Template.Spec.Containers {
			// 第一个放进来
			if v.Image == container.Name && k == 0 {
				firstContainer = append(firstContainer, container)
			} else if v.Image == container.Name {
				setOtherContainers = append(setOtherContainers, container)
			}
		}
	}
	podDeployer.Spec.DaemonSetSpec.Template.Spec.Containers = firstContainer
	return setOtherContainers, nil
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// patchResource 使用 dynamic client patch
func patchResource(deploymentName, namespace string, container *v1.Container, gvr schema.GroupVersionResource) error {
	klog.Infof("do resource patch [%v/%v]", namespace, deploymentName)

	pa := make([]patchOperation, 0)
	p := patchOperation{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/template/spec/containers/-"),
		Value: container,
	}
	pa = append(pa, p)
	patchBytes, err := json.Marshal(&pa)
	if err != nil {
		klog.Error("patch marshal error: ", err)
		return err
	}

	jsonPatch, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		klog.Error("DecodePatch error: ", err)
		return err
	}

	jsonPatchBytes, err := json.Marshal(jsonPatch)
	if err != nil {
		klog.Error("json Marshal error: ", err)
		return err
	}
	klog.Infof("patch operation: %v", string(jsonPatchBytes))
	_, err = k8sconfig.DynamicClient.Resource(gvr).Namespace(namespace).Patch(context.TODO(), deploymentName,
		types.JSONPatchType, jsonPatchBytes, metav1.PatchOptions{})

	if err != nil {
		klog.Error("patch error: ", err)
		return err
	}
	return nil
}
