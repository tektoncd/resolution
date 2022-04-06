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

package pipeline_bundles_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knativetest "knative.dev/pkg/test"
)

// testNamespace is the namespace to construct and run this e2e test in.
const testNamespace = "tekton-resolution-pipeline-bundles-test"

// waitInterval is the duration between repeat attempts to check on the
// status of the test's kubernetes resources.
const waitInterval = time.Second

// waitTimeout is the total maximum time the test may spend waiting for
// successful creation or completion of resources deployed during this test.
// The timeout is high because the CI/CD cluster can be slow.
const waitTimeout = 2 * time.Minute

// testBundleRefEnvVar is the name of an environment variable that's used to
// pass the url of a bundle that can be pulled during the e2e test.
const testBundleRefEnvVar = "TEST_BUNDLE_REF"

// TestPipelineBundle executes a PipelineRun that relies on a Pipeline
// from a Bundle stored in a registry.
func TestPipelineBundle(t *testing.T) {
	bundleRef := os.Getenv(testBundleRefEnvVar)
	if bundleRef == "" {
		t.Fatalf("test requires a bundle be made available via environment variable %q", testBundleRefEnvVar)
	}

	ctx := context.Background()
	configPath := knativetest.Flags.Kubeconfig
	clusterName := knativetest.Flags.Cluster

	cfg, err := knativetest.BuildClientConfig(configPath, clusterName)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create kubeclient from config file at %s: %s", configPath, err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create dynamic client from config file at %s: %s", configPath, err)
	}

	tearDown := func() {
		err := kubeClient.CoreV1().Namespaces().Delete(ctx, testNamespace, metav1.DeleteOptions{})
		if err != nil {
			t.Errorf("error deleting test namespace %q: %v", testNamespace, err)
		}
	}

	knativetest.CleanupOnInterrupt(tearDown, t.Logf)
	defer tearDown()

	_, err = kubeClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
			Labels: map[string]string{
				"resolution.tekton.dev/test-e2e": "true",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace %s for tests: %s", testNamespace, err)
	}

	pipelineRunYAML, err := os.ReadFile("./pipelinerun.yaml")
	if err != nil {
		t.Fatalf("error reading pipelinerun yaml fixture: %v", err)
	}
	pipelineRunYAML = bytes.Replace(pipelineRunYAML, []byte("{{bundleRef}}"), []byte(bundleRef), 1)

	// Create PipelineRun using dynamic client to avoid importing
	// pipelines as dependency.
	pipelineRun := &unstructured.Unstructured{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(pipelineRunYAML, nil, pipelineRun)
	if err != nil {
		t.Fatalf("error parsing into unstructured pipelinerun: %v", err)
	}
	pipelineRunGVR := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1beta1",
		Resource: "pipelineruns",
	}
	if _, err := dynamicClient.Resource(pipelineRunGVR).Namespace(testNamespace).Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error creating pipelinerun object: %v", err)
	}

	pipelineRunName, has, err := unstructured.NestedString(pipelineRun.UnstructuredContent(), "metadata", "name")
	if err != nil {
		t.Fatalf("error reading pipelinerun name: %v", err)
	} else if !has {
		t.Fatalf("expected pipelinerun to have metadata.name but none was found")
	}

	err = wait.PollImmediate(waitInterval, waitTimeout, func() (bool, error) {
		pr, err := dynamicClient.Resource(pipelineRunGVR).Namespace(testNamespace).Get(ctx, pipelineRunName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting pipelinerun: %v", err)
		}
		conditions, err := getPipelineRunConditions(pr)
		if err != nil {
			t.Fatalf("error reading pipelinerun conditions: %v", err)
		}
		for _, condition := range conditions {
			if condition["type"] == "Succeeded" {
				switch condition["status"] {
				case "Unknown":
					return false, nil
				case "False":
					return false, fmt.Errorf("pipelinerun failed with reason %q and message %q", condition["reason"], condition["message"])
				case "True":
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("pipelinerun did not succeed: %v", err)
	}
}

// getPipelineRunConditions returns the status.conditions from an
// unstructured pipelinerun. If no conditions are found a nil map and
// nil error are returned.
func getPipelineRunConditions(pr *unstructured.Unstructured) ([]map[string]string, error) {
	conditions, has, err := unstructured.NestedSlice(pr.UnstructuredContent(), "status", "conditions")
	if !has {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("invalid conditions: %v", err)
	}
	ret := []map[string]string{}
	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("received condition with unexpected layout: %#v", cond)
		}
		conditionMap := map[string]string{}
		for condKey, condVal := range condition {
			stringVal, ok := condVal.(string)
			if !ok {
				return nil, fmt.Errorf("non-string value in condition %#v", cond)
			}
			conditionMap[condKey] = stringVal
		}
		ret = append(ret, conditionMap)
	}
	return ret, nil
}
