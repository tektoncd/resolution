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

package main

import (
	"context"
	"time"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/tektoncd/resolution/bundleresolver/pkg/bundle"
	"github.com/tektoncd/resolution/pkg/common"
	"github.com/tektoncd/resolution/pkg/resolver/framework"
	"k8s.io/client-go/kubernetes"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection/sharedmain"
)

// TODO(sbwsg): This should be exposed as a configurable option for
// admins (e.g. via ConfigMap)
const timeoutDuration = time.Minute

func main() {
	sharedmain.Main("controller",
		framework.NewController(context.Background(), &resolver{}),
	)
}

type resolver struct {
	kubeClientSet kubernetes.Interface
}

// Initialize sets up any dependencies needed by the resolver. None atm.
func (r *resolver) Initialize(ctx context.Context) error {
	r.kubeClientSet = kubeclient.Get(ctx)
	return nil
}

// GetName returns a string name to refer to this resolver by.
func (r *resolver) GetName(context.Context) string {
	return "bundleresolver"
}

// GetSelector returns a map of labels to match requests to this resolver.
func (r *resolver) GetSelector(context.Context) map[string]string {
	return map[string]string{
		common.LabelKeyResolverType: "bundle",
	}
}

// ValidateParams ensures parameters from a request are as expected.
func (r *resolver) ValidateParams(ctx context.Context, params map[string]string) error {
	if _, err := bundle.OptionsFromParams(params); err != nil {
		return err
	}
	return nil
}

// Resolve uses the given params to resolve the requested file or resource.
func (r *resolver) Resolve(ctx context.Context, params map[string]string) (framework.ResolvedResource, error) {
	opts, err := bundle.OptionsFromParams(params)
	if err != nil {
		return nil, err
	}
	kc, err := k8schain.New(ctx, r.kubeClientSet, k8schain.Options{
		Namespace:          opts.Namespace,
		ServiceAccountName: opts.ServiceAccount,
	})
	ctx, cancelFn := context.WithTimeout(ctx, timeoutDuration)
	defer cancelFn()
	return bundle.GetEntry(ctx, kc, opts)
}

// resolvedResource wraps the data we want to return to Pipelines
type resolvedResource struct {
	data []byte
}

// Data returns the bytes of our hard-coded Pipeline
func (rr *resolvedResource) Data() []byte {
	return rr.data
}

// Annotations returns any metadata needed alongside the data. None atm.
func (*resolvedResource) Annotations() map[string]string {
	return nil
}
