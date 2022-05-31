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
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
)

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
	c, ns := setup(ctx, t)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, ns) }, t.Logf)
	defer tearDown(ctx, t, c, ns)

	pipelineRunYAML, err := os.ReadFile("./pipeline_bundles_test/pipelinerun.yaml")
	if err != nil {
		t.Fatalf("error reading pipelinerun yaml fixture: %v", err)
	}
	pipelineRunYAML = bytes.Replace(pipelineRunYAML, []byte("{{bundleRef}}"), []byte(bundleRef), 1)
	pipelineRunYAML = bytes.Replace(pipelineRunYAML, []byte("{{prName}}"), []byte(helpers.ObjectNameForTest(t)), 1)

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
	if _, err := c.DynamicClient.Resource(pipelineRunGVR).Namespace(ns).Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error creating pipelinerun object: %v", err)
	}

	pipelineRunName, has, err := unstructured.NestedString(pipelineRun.UnstructuredContent(), "metadata", "name")
	if err != nil {
		t.Fatalf("error reading pipelinerun name: %v", err)
	} else if !has {
		t.Fatalf("expected pipelinerun to have metadata.name but none was found")
	}

	err = wait.PollImmediate(waitInterval, waitTimeout, func() (bool, error) {
		pr, err := c.DynamicClient.Resource(pipelineRunGVR).Namespace(ns).Get(ctx, pipelineRunName, metav1.GetOptions{})
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
