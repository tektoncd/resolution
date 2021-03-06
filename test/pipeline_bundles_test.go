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
	"os"
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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

	pipelineRunYAML, err := os.ReadFile("./pipeline_bundles_test/pipelinerun.yaml")
	if err != nil {
		t.Fatalf("error reading pipelinerun yaml fixture: %v", err)
	}
	pipelineRunYAML = bytes.Replace(pipelineRunYAML, []byte("{{bundleRef}}"), []byte(bundleRef), 1)
	pipelineRunYAML = bytes.Replace(pipelineRunYAML, []byte("{{prName}}"), []byte(helpers.ObjectNameForTest(t)), 1)

	err = RunPipeline(pipelineRunYAML, t, waitInterval, waitTimeout)

	if err != nil {
		t.Fatalf("pipelinerun did not succeed: %v", err)
	}
}
