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

	"github.com/tektoncd/resolution/bundleresolver/pkg/bundle"
	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	"github.com/tektoncd/resolution/pkg/resolver/framework"
	filteredinformerfactory "knative.dev/pkg/client/injection/kube/informers/factory/filtered"
	"knative.dev/pkg/injection/sharedmain"
)

func main() {
	ctx := filteredinformerfactory.WithSelectors(context.Background(), v1alpha1.ManagedByLabelKey)
	sharedmain.MainWithContext(ctx, "controller",
		framework.NewController(ctx, &bundle.Resolver{}),
	)
}
