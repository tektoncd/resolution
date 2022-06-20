//go:build e2e

package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	knativetest "knative.dev/pkg/test"
)

// RunPipeline creates PipelineRun using dynamic
// client to avoid importing pipelines as dependency.
func RunPipeline(pipelineRunYAML []byte, t *testing.T, interval time.Duration, waitTimeout time.Duration) error {
	ctx := context.Background()
	c, ns := setup(ctx, t)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, ns) }, t.Logf)
	defer tearDown(ctx, t, c, ns)

	pipelineRun := &unstructured.Unstructured{}
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode(pipelineRunYAML, nil, pipelineRun)
	if err != nil {
		return fmt.Errorf("error parsing into unstructured pipelinerun: %v", err)
	}

	pipelineRunGVR := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1beta1",
		Resource: "pipelineruns",
	}
	if _, err := c.DynamicClient.Resource(pipelineRunGVR).Namespace(ns).Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("error creating pipelinerun object: %v", err)
	}

	pipelineRunName, has, err := unstructured.NestedString(pipelineRun.UnstructuredContent(), "metadata", "name")
	if err != nil {
		return fmt.Errorf("error reading pipelinerun name: %v", err)
	} else if !has {
		return fmt.Errorf("expected pipelinerun to have metadata.name but none was found")
	}

	err = wait.PollImmediate(interval, waitTimeout, func() (bool, error) {
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
		return err
	}
	return nil
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
