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

package framework

import (
	"context"
	"fmt"
	"strings"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	rrclient "github.com/tektoncd/resolution/pkg/client/injection/client"
	rrinformer "github.com/tektoncd/resolution/pkg/client/injection/informers/resolution/v1alpha1/resourcerequest"
	rrlister "github.com/tektoncd/resolution/pkg/client/listers/resolution/v1alpha1"
	"github.com/tektoncd/resolution/pkg/common"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// NewController returns a knative controller for a Tekton Resolver.
// This sets up a lot of the boilerplate that individual resolvers
// shouldn't need to be concerned with since it's common to all of them.
func NewController(ctx context.Context, resolver Resolver) func(context.Context, configmap.Watcher) *controller.Impl {
	if err := validateResolver(ctx, resolver); err != nil {
		panic(err.Error())
	}
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		logger := logging.FromContext(ctx)
		kubeclientset := kubeclient.Get(ctx)
		rrclientset := rrclient.Get(ctx)
		rrInformer := rrinformer.Get(ctx)

		if err := resolver.Initialize(ctx); err != nil {
			panic(err.Error())
		}

		r := &Reconciler{
			LeaderAwareFuncs:         leaderAwareFuncs(rrInformer.Lister()),
			kubeClientSet:            kubeclientset,
			resourceRequestLister:    rrInformer.Lister(),
			resourceRequestClientSet: rrclientset,
			resolver:                 resolver,
		}

		// TODO(sbwsg): Do better sanitize.
		resolverName := resolver.GetName(ctx)
		resolverName = strings.ReplaceAll(resolverName, "/", "")
		resolverName = strings.ReplaceAll(resolverName, " ", "")

		impl := controller.NewContext(ctx, r, controller.ControllerOptions{
			WorkQueueName: "TektonResolverFramework." + resolverName,
			Logger:        logger,
		})

		rrInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: filterResourceRequestsBySelector(resolver.GetSelector(ctx)),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: impl.Enqueue,
				UpdateFunc: func(oldObj, newObj interface{}) {
					impl.Enqueue(newObj)
				},
				// TODO(sbwsg): should we deliver delete events
				// to the resolver?
				// DeleteFunc: impl.Enqueue,
			},
		})

		return impl
	}
}

func filterResourceRequestsBySelector(selector map[string]string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		rr, ok := obj.(*v1alpha1.ResourceRequest)
		if !ok {
			return false
		}
		if len(rr.ObjectMeta.Labels) == 0 {
			return false
		}
		for key, val := range selector {
			lookup, has := rr.ObjectMeta.Labels[key]
			if !has {
				return false
			}
			if lookup != val {
				return false
			}
		}
		return true
	}
}

// TODO(sbwsg): I don't really understand the LeaderAwareness types beyond the
// fact that the controller crashes if they're missing. It looks
// like this is bucketing based on labels. Should we use the filter
// selector from above in the call to lister.List here?
func leaderAwareFuncs(lister rrlister.ResourceRequestLister) reconciler.LeaderAwareFuncs {
	return reconciler.LeaderAwareFuncs{
		PromoteFunc: func(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
			all, err := lister.List(labels.Everything())
			if err != nil {
				return err
			}
			for _, elt := range all {
				enq(bkt, types.NamespacedName{
					Namespace: elt.GetNamespace(),
					Name:      elt.GetName(),
				})
			}
			return nil
		},
	}
}

// ErrorMissingTypeSelector is returned when a resolver does not return
// a selector with a type label from its GetSelector method.
var ErrorMissingTypeSelector = fmt.Errorf("invalid resolver: minimum selector must include %q", common.LabelKeyResolverType)

func validateResolver(ctx context.Context, r Resolver) error {
	sel := r.GetSelector(ctx)
	if sel == nil {
		return ErrorMissingTypeSelector
	}
	if sel[common.LabelKeyResolverType] == "" {
		return ErrorMissingTypeSelector
	}
	return nil
}
