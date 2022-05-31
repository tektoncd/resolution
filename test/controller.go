/*
Copyright 2022 The Tekton Authors

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

package test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	fakeresolutionclientset "github.com/tektoncd/resolution/pkg/client/clientset/versioned/fake"
	resolutioninformersv1alpha1 "github.com/tektoncd/resolution/pkg/client/informers/externalversions/resolution/v1alpha1"
	fakeresolutionrequestclient "github.com/tektoncd/resolution/pkg/client/injection/client/fake"
	fakeresolutionrequestinformer "github.com/tektoncd/resolution/pkg/client/injection/informers/resolution/v1alpha1/resolutionrequest/fake"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	coreinformers "k8s.io/client-go/informers/core/v1"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	fakeconfigmapinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/configmap/fake"
	fakelimitrangeinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/limitrange/fake"
	fakeserviceaccountinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount/fake"
	"knative.dev/pkg/controller"
)

// Data represents the desired state of the system (i.e. existing resources) to seed controllers
// with.
type Data struct {
	ResolutionRequests []*v1alpha1.ResolutionRequest
	Namespaces         []*corev1.Namespace
	ConfigMaps         []*corev1.ConfigMap
	ServiceAccounts    []*corev1.ServiceAccount
	LimitRange         []*corev1.LimitRange
}

// Clients holds references to clients which are useful for reconciler tests.
type Clients struct {
	ResolutionRequests *fakeresolutionclientset.Clientset
	Kube               *fakekubeclientset.Clientset
}

// Informers holds references to informers which are useful for reconciler tests.
type Informers struct {
	ConfigMap         coreinformers.ConfigMapInformer
	ServiceAccount    coreinformers.ServiceAccountInformer
	LimitRange        coreinformers.LimitRangeInformer
	ResolutionRequest resolutioninformersv1alpha1.ResolutionRequestInformer
}

// Assets holds references to the controller, logs, clients, and informers.
type Assets struct {
	Logger     *zap.SugaredLogger
	Controller *controller.Impl
	Clients    Clients
	Informers  Informers
	Recorder   *record.FakeRecorder
	Ctx        context.Context
}

// AddToInformer returns a function to add ktesting.Actions to the cache store
func AddToInformer(t *testing.T, store cache.Store) func(ktesting.Action) (bool, runtime.Object, error) {
	return func(action ktesting.Action) (bool, runtime.Object, error) {
		switch a := action.(type) {
		case ktesting.CreateActionImpl:
			if err := store.Add(a.GetObject()); err != nil {
				t.Fatal(err)
			}

		case ktesting.UpdateActionImpl:
			objMeta, err := meta.Accessor(a.GetObject())
			if err != nil {
				return true, nil, err
			}

			// Look up the old copy of this resource and perform the optimistic concurrency check.
			old, exists, err := store.GetByKey(objMeta.GetNamespace() + "/" + objMeta.GetName())
			if err != nil {
				return true, nil, err
			} else if !exists {
				// Let the client return the error.
				return false, nil, nil
			}
			oldMeta, err := meta.Accessor(old)
			if err != nil {
				return true, nil, err
			}
			// If the resource version is mismatched, then fail with a conflict.
			if oldMeta.GetResourceVersion() != objMeta.GetResourceVersion() {
				return true, nil, apierrs.NewConflict(
					a.Resource.GroupResource(), objMeta.GetName(),
					fmt.Errorf("resourceVersion mismatch, got: %v, wanted: %v",
						objMeta.GetResourceVersion(), oldMeta.GetResourceVersion()))
			}

			// Update the store with the new object when it's fine.
			if err := store.Update(a.GetObject()); err != nil {
				t.Fatal(err)
			}
		}
		return false, nil, nil
	}
}

// SeedTestData returns Clients and Informers populated with the
// given Data.
// nolint: revive
func SeedTestData(t *testing.T, ctx context.Context, d Data) (Clients, Informers) {
	c := Clients{
		Kube:               fakekubeclient.Get(ctx),
		ResolutionRequests: fakeresolutionrequestclient.Get(ctx),
	}
	// Every time a resource is modified, change the metadata.resourceVersion.
	PrependResourceVersionReactor(&c.ResolutionRequests.Fake)

	i := Informers{
		ConfigMap:         fakeconfigmapinformer.Get(ctx),
		ServiceAccount:    fakeserviceaccountinformer.Get(ctx),
		LimitRange:        fakelimitrangeinformer.Get(ctx),
		ResolutionRequest: fakeresolutionrequestinformer.Get(ctx),
	}

	// Attach reactors that add resource mutations to the appropriate
	// informer index, and simulate optimistic concurrency failures when
	// the resource version is mismatched.
	c.ResolutionRequests.PrependReactor("*", "resolutionrequests", AddToInformer(t, i.ResolutionRequest.Informer().GetIndexer()))
	for _, pr := range d.ResolutionRequests {
		pr := pr.DeepCopy() // Avoid assumptions that the informer's copy is modified.
		if _, err := c.ResolutionRequests.ResolutionV1alpha1().ResolutionRequests(pr.Namespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range d.Namespaces {
		n := n.DeepCopy() // Avoid assumptions that the informer's copy is modified.
		if _, err := c.Kube.CoreV1().Namespaces().Create(ctx, n, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	c.Kube.PrependReactor("*", "configmaps", AddToInformer(t, i.ConfigMap.Informer().GetIndexer()))
	for _, cm := range d.ConfigMaps {
		cm := cm.DeepCopy() // Avoid assumptions that the informer's copy is modified.
		if _, err := c.Kube.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, cm, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	c.Kube.PrependReactor("*", "serviceaccounts", AddToInformer(t, i.ServiceAccount.Informer().GetIndexer()))
	for _, sa := range d.ServiceAccounts {
		sa := sa.DeepCopy() // Avoid assumptions that the informer's copy is modified.
		if _, err := c.Kube.CoreV1().ServiceAccounts(sa.Namespace).Create(ctx, sa, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	c.ResolutionRequests.PrependReactor("*", "resolutionrequests", AddToInformer(t, i.ResolutionRequest.Informer().GetIndexer()))
	c.ResolutionRequests.ClearActions()
	c.Kube.ClearActions()
	c.ResolutionRequests.ClearActions()
	return c, i
}

// ResourceVersionReactor is an implementation of Reactor for our tests
type ResourceVersionReactor struct {
	count int64
}

// Handles returns whether our test reactor can handle a given ktesting.Action
func (r *ResourceVersionReactor) Handles(action ktesting.Action) bool {
	body := func(o runtime.Object) bool {
		objMeta, err := meta.Accessor(o)
		if err != nil {
			return false
		}
		val := atomic.AddInt64(&r.count, 1)
		objMeta.SetResourceVersion(fmt.Sprintf("%05d", val))
		return false
	}

	switch o := action.(type) {
	case ktesting.CreateActionImpl:
		return body(o.GetObject())
	case ktesting.UpdateActionImpl:
		return body(o.GetObject())
	default:
		return false
	}
}

// React is noop-function
func (r *ResourceVersionReactor) React(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
	return false, nil, nil
}

var _ ktesting.Reactor = (*ResourceVersionReactor)(nil)

// PrependResourceVersionReactor will instrument a client-go testing Fake
// with a reactor that simulates resourceVersion changes on mutations.
// This does not work with patches.
func PrependResourceVersionReactor(f *ktesting.Fake) {
	f.ReactionChain = append([]ktesting.Reactor{&ResourceVersionReactor{}}, f.ReactionChain...)
}
