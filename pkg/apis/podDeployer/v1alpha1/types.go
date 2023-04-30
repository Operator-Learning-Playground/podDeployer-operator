package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Poddeployer
type Poddeployer struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PodDeployerSpec `json:"spec,omitempty"`
}

type PodDeployerSpec struct {
	DeploymentSpec appsv1.DeploymentSpec `json:"deployment_spec"`
	PriorityImages []PriorityImage       `json:"priority_images,omitempty"`
}

type PriorityImage struct {
	Image string `json:"image"`
	Value int    `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodReStarterList
type PoddeployerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Poddeployer `json:"items"`
}
