//go:build e2e

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
	"os"
	"testing"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	"github.com/tektoncd/resolution/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
)

// TestSmoke performs all the setup for an e2e test and exercises basic
// e2e functionality like namespace creation + teardown, resolution request
// creation and waiting for resolution request completion.
func TestSmoke(t *testing.T) {
	ctx := context.Background()

	c, ns := setup(ctx, t)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, ns) }, t.Logf)
	defer tearDown(ctx, t, c, ns)

	requestYAML, err := os.ReadFile("./smoke_test/resolution-request.yaml")
	if err != nil {
		t.Log(os.Getwd())
		t.Fatalf("unable to read resolution request yaml fixture: %v", err)
	}

	req := &v1alpha1.ResolutionRequest{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(requestYAML, nil, req)
	if err != nil {
		t.Fatalf("error parsing resolution request yaml fixture: %v", err)
	}
	req.Name = helpers.ObjectNameForTest(t)
	req.Namespace = ns

	_, err = c.ResolutionRequestClient.Create(ctx, req, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	err = wait.PollImmediate(waitInterval, waitTimeout, func() (bool, error) {
		latestResolutionRequest, err := c.ResolutionRequestClient.Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		resolvedData := latestResolutionRequest.Status.ResolutionRequestStatusFields.Data
		if resolvedData != "" {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		t.Fatalf("error waiting for completed resolution request: %v", err)
	}
}
