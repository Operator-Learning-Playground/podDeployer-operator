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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	poddeployerv1alpha1 "github.com/myoperator/poddeployer/pkg/apis/poddeployer/v1alpha1"
	versioned "github.com/myoperator/poddeployer/pkg/client/clientset/versioned"
	internalinterfaces "github.com/myoperator/poddeployer/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/myoperator/poddeployer/pkg/client/listers/poddeployer/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// PoddeployerInformer provides access to a shared informer and lister for
// Poddeployers.
type PoddeployerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.PoddeployerLister
}

type poddeployerInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewPoddeployerInformer constructs a new informer for Poddeployer type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPoddeployerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPoddeployerInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredPoddeployerInformer constructs a new informer for Poddeployer type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPoddeployerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApiV1alpha1().Poddeployers(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApiV1alpha1().Poddeployers(namespace).Watch(context.TODO(), options)
			},
		},
		&poddeployerv1alpha1.Poddeployer{},
		resyncPeriod,
		indexers,
	)
}

func (f *poddeployerInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPoddeployerInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *poddeployerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&poddeployerv1alpha1.Poddeployer{}, f.defaultInformer)
}

func (f *poddeployerInformer) Lister() v1alpha1.PoddeployerLister {
	return v1alpha1.NewPoddeployerLister(f.Informer().GetIndexer())
}