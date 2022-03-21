//go:build e2e
// +build e2e

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

package smoke_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	"github.com/tektoncd/resolution/pkg/client/clientset/versioned"
	"github.com/tektoncd/resolution/pkg/client/clientset/versioned/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knativetest "knative.dev/pkg/test"
)

// waitInterval is the duration between repeat attempts to check on the
// status of the test's resolution request.
const waitInterval = time.Second

// waitTimeout is the total maximum time the test may spend waiting for
// successful resolution of the test's resolution request.
const waitTimeout = 5 * time.Second

// TestSmoke performs all the setup for an e2e test and exercises basic
// e2e functionality like namespace creation + teardown, resolution request
// creation and waiting for resolution request completion.
func TestSmoke(t *testing.T) {
	ctx := context.Background()
	configPath := knativetest.Flags.Kubeconfig
	clusterName := knativetest.Flags.Cluster

	requestYAML, err := os.ReadFile("./resolution-request.yaml")
	if err != nil {
		t.Log(os.Getwd())
		t.Fatalf("unable to read resolution request yaml fixture: %v", err)
	}

	req := &v1alpha1.ResolutionRequest{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(requestYAML, nil, req)
	if err != nil {
		t.Fatalf("error parsing resolution request yaml fixture: %v", err)
	}

	cfg, err := knativetest.BuildClientConfig(configPath, clusterName)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create kubeclient from config file at %s: %s", configPath, err)
	}

	tearDown := func() {
		err := kubeClient.CoreV1().Namespaces().Delete(ctx, req.Namespace, metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("error deleting test namespace %q: %v", req.Namespace, err)
		}
	}

	knativetest.CleanupOnInterrupt(tearDown, t.Logf)
	defer tearDown()

	_, err = kubeClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Namespace,
			Labels: map[string]string{
				"resolution.tekton.dev/test-e2e": "true",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace %s for tests: %s", req.Namespace, err)
	}

	clientset, err := versioned.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("error getting resolution clientset: %v", err)
	}

	_, err = clientset.ResolutionV1alpha1().ResolutionRequests(req.Namespace).Create(ctx, req, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	err = wait.PollImmediate(waitInterval, waitTimeout, func() (bool, error) {
		latestResolutionRequest, err := clientset.ResolutionV1alpha1().ResolutionRequests(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
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
