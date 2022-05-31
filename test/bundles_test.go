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
	"time"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
)

// waitInterval is the duration between repeat attempts to check on the
// status of the test's resolution request.
const waitInterval = time.Second

// waitTimeout is the total maximum time the test may spend waiting for
// successful resolution of the test's bundle request.
const waitTimeout = 20 * time.Second

// TestBundlesSmoke creates a resolution request for a bundle and checks
// that it succeeds.
func TestBundlesSmoke(t *testing.T) {
	ctx := context.Background()

	requestYAML, err := os.ReadFile("./bundles_test/resolution-request.yaml")
	if err != nil {
		t.Fatalf("unable to read resolution request yaml fixture: %v", err)
	}

	req := &v1alpha1.ResolutionRequest{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(requestYAML, nil, req)
	if err != nil {
		t.Fatalf("error parsing resolution request yaml fixture: %v", err)
	}
	req.Name = helpers.ObjectNameForTest(t)

	c, ns := setup(ctx, t)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, ns) }, t.Logf)
	defer tearDown(ctx, t, c, ns)

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
