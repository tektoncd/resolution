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
	"testing"

	"k8s.io/client-go/dynamic"

	"github.com/tektoncd/resolution/pkg/client/clientset/versioned"
	"github.com/tektoncd/resolution/pkg/client/clientset/versioned/typed/resolution/v1alpha1"
	"k8s.io/client-go/kubernetes"
	knativetest "knative.dev/pkg/test"
)

// clients holds instances of interfaces for making requests to the Pipeline controllers.
type clients struct {
	KubeClient kubernetes.Interface

	ResolutionRequestClient v1alpha1.ResolutionRequestInterface
	DynamicClient           dynamic.Interface
}

// newClients instantiates and returns several clientsets required for making requests to the
// Pipeline cluster specified by the combination of clusterName and configPath. Clients can
// make requests within namespace.
func newClients(t *testing.T, configPath, clusterName, namespace string) *clients {
	t.Helper()
	var err error
	c := &clients{}

	cfg, err := knativetest.BuildClientConfig(configPath, clusterName)
	if err != nil {
		t.Fatalf("failed to create configuration obj from %s for cluster %s: %s", configPath, clusterName, err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create kubeclient from config file at %s: %s", configPath, err)
	}
	c.KubeClient = kubeClient

	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create resolution request clientset from config file at %s: %s", configPath, err)
	}
	c.ResolutionRequestClient = cs.ResolutionV1alpha1().ResolutionRequests(namespace)

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create dynamic client from config file at %s: %s", configPath, err)
	}
	c.DynamicClient = dynamicClient
	return c
}
