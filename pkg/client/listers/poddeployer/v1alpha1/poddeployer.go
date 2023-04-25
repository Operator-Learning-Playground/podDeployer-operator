/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/myoperator/poddeployer/pkg/apis/poddeployer/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PoddeployerLister helps list Poddeployers.
// All objects returned here must be treated as read-only.
type PoddeployerLister interface {
	// List lists all Poddeployers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Poddeployer, err error)
	// Poddeployers returns an object that can list and get Poddeployers.
	Poddeployers(namespace string) PoddeployerNamespaceLister
	PoddeployerListerExpansion
}

// poddeployerLister implements the PoddeployerLister interface.
type poddeployerLister struct {
	indexer cache.Indexer
}

// NewPoddeployerLister returns a new PoddeployerLister.
func NewPoddeployerLister(indexer cache.Indexer) PoddeployerLister {
	return &poddeployerLister{indexer: indexer}
}

// List lists all Poddeployers in the indexer.
func (s *poddeployerLister) List(selector labels.Selector) (ret []*v1alpha1.Poddeployer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Poddeployer))
	})
	return ret, err
}

// Poddeployers returns an object that can list and get Poddeployers.
func (s *poddeployerLister) Poddeployers(namespace string) PoddeployerNamespaceLister {
	return poddeployerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PoddeployerNamespaceLister helps list and get Poddeployers.
// All objects returned here must be treated as read-only.
type PoddeployerNamespaceLister interface {
	// List lists all Poddeployers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Poddeployer, err error)
	// Get retrieves the Poddeployer from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Poddeployer, error)
	PoddeployerNamespaceListerExpansion
}

// poddeployerNamespaceLister implements the PoddeployerNamespaceLister
// interface.
type poddeployerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Poddeployers in the indexer for a given namespace.
func (s poddeployerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Poddeployer, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Poddeployer))
	})
	return ret, err
}

// Get retrieves the Poddeployer from the indexer for a given namespace and name.
func (s poddeployerNamespaceLister) Get(name string) (*v1alpha1.Poddeployer, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("poddeployer"), name)
	}
	return obj.(*v1alpha1.Poddeployer), nil
}
