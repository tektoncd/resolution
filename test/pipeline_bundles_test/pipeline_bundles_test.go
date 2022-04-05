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

	"github.com/tektoncd/resolution/test/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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

// TestPipelineBundle executes a PipelineRun that relies on a Pipeline
// from a Bundle stored in a registry.
func TestPipelineBundle(t *testing.T) {
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

	serviceIP, err := deployRegistry(ctx, kubeClient)
	if err != nil {
		t.Fatalf("error deploying registry: %v", err)
	}

	bundleRef := fmt.Sprintf("%s:5000/simple/pipeline:latest", serviceIP)

	pipelineYAML, err := os.ReadFile("./pipeline.yaml")
	if err != nil {
		t.Fatalf("unable to read pipeline yaml fixture: %v", err)
	}

	err = publishTestBundle(ctx, kubeClient, string(pipelineYAML), bundleRef)
	if err != nil {
		t.Fatalf("error publishing bundle: %v", err)
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

// publishTestBundle uses a pod running tkn to push a YAML file as a
// bundle to the registry identified by bundleRef.
func publishTestBundle(ctx context.Context, kubeClient kubernetes.Interface, bundleContent, bundleRef string) error {
	pod := &v1.Pod{}
	if err := helpers.ReadFileToObject("./pod-tkn-bundle-push.yaml", pod); err != nil {
		return fmt.Errorf("error reading pod to publish bundle: %w", err)
	}
	pod.ObjectMeta.Namespace = testNamespace
	pod.ObjectMeta.Annotations["pipeline_yaml"] = bundleContent
	pod.ObjectMeta.Annotations["bundle_ref"] = bundleRef
	_, err := kubeClient.CoreV1().Pods(testNamespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating pod to publish bundle: %w", err)
	}
	err = wait.PollImmediate(waitInterval, waitTimeout, func() (bool, error) {
		p, err := kubeClient.CoreV1().Pods(testNamespace).Get(ctx, pod.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch p.Status.Phase {
		case corev1.PodSucceeded:
			return true, nil
		case corev1.PodFailed:
			return false, fmt.Errorf("error publishing bundle. pod status: %#v", p.Status)
		default:
			return false, nil
		}
	})
	return err
}

// deployRegistry submits a registry deployment and service to
// kubernetes and then waits for them to come up and be ready. The
// clusterIP of the service is returned on success.
func deployRegistry(ctx context.Context, kubeClient kubernetes.Interface) (string, error) {
	deployment := &appsv1.Deployment{}
	if err := helpers.ReadFileToObject("./registry-deployment.yaml", deployment); err != nil {
		return "", fmt.Errorf("error reading registry deployment: %w", err)
	}

	service := &corev1.Service{}
	if err := helpers.ReadFileToObject("./registry-service.yaml", service); err != nil {
		return "", fmt.Errorf("error reading registry service: %w", err)
	}

	_, err := kubeClient.AppsV1().Deployments(testNamespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("error creating registry deployment: %w", err)
	}

	_, err = kubeClient.CoreV1().Services(testNamespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("error creating registry service: %w", err)
	}

	err = wait.PollImmediate(waitInterval, waitTimeout, func() (bool, error) {
		d, err := kubeClient.AppsV1().Deployments(testNamespace).Get(ctx, deployment.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		s, err := kubeClient.CoreV1().Services(testNamespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if d.Status.ReadyReplicas > 0 && s.Spec.ClusterIP != "" && s.Spec.ClusterIP != corev1.ClusterIPNone {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return "", fmt.Errorf("error waiting for deployment and service: %w", err)
	}

	s, err := kubeClient.CoreV1().Services(testNamespace).Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting service: %w", err)
	}

	return s.Spec.ClusterIP, nil
}
