#!/usr/bin/env bash

# Copyright 2022 The Tekton Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

source $(git rev-parse --show-toplevel)/vendor/github.com/tektoncd/plumbing/scripts/e2e-tests.sh

initialize $@

failed=0

header "Deploying Tekton Pipelines"
git clone https://github.com/tektoncd/pipeline
cd pipeline
ko apply -f ./config/100-namespace
ko apply -f ./config
cd -
rm -rf pipeline

header "Deploying Tekton Resolution"
ko apply -f ./config

header "Deploying Git Resolver"
ko apply -f ./gitresolver/config

header "Deploying Bundle Resolver"
ko apply -f ./bundleresolver/config

header "Deploying Hub Resolver"
ko apply -f ./hubresolver/config

header "Deploying Resolver Template"
ko apply -f ./docs/resolver-template/config

# update the feature-flags configmap in the tekton-pipelines namespace
# so that remote resolution is enabled
kubectl patch -n tekton-pipelines configmap feature-flags -p '{"data":{"enable-api-fields":"alpha"}}'

wait_until_pods_running "tekton-remote-resolution" || fail_test "Tekton Resolution did not come up"

# Deploy a test registry with an example bundle that can be used across
# multiple e2e tests.
header "Deploying OCI Registry and a Test Bundle"
kubectl create namespace tekton-resolution-example-registry
kubectl apply -n tekton-resolution-example-registry -f ./test/bundle_registry/registry-deployment.yaml
kubectl apply -n tekton-resolution-example-registry -f ./test/bundle_registry/registry-service.yaml

wait_until_pods_running "tekton-resolution-example-registry" || fail_test "Test registry did not come up"

TEST_REGISTRY_CLUSTER_IP=$(kubectl get service registry -n tekton-resolution-example-registry -o jsonpath="{$.spec.clusterIP}")
if [ "$TEST_REGISTRY_CLUSTER_IP" = "" ] ; then
  fail_test "Cant find test registry cluster ip"
fi
if [ "$TEST_REGISTRY_CLUSTER_IP" = "None" ] ; then
  fail_test "Test regsitry has cluster ip None"
fi

BUNDLE_PATH="simple/pipeline:latest"

kubectl port-forward -n tekton-resolution-example-registry service/registry 9090:5000 &
KUBECTL_PORT_FORWARD_PID=$!

# give the port-forward a moment to establish; without this tkn bundle
# push can emit strange errors like BLOB_UPLOAD_UNKNOWN
sleep 2

LOCAL_BUNDLE_REF="localhost:9090/${BUNDLE_PATH}"
export TEST_BUNDLE_REF="${TEST_REGISTRY_CLUSTER_IP}:5000/${BUNDLE_PATH}"

tkn bundle push -f ./test/bundle_registry/pipeline.yaml "$LOCAL_BUNDLE_REF"
if [ $? -ne 0 ] ; then
  kill -15 ${KUBECTL_PORT_FORWARD_PID}
  fail_test "Failed to push bundle to test registry"
fi

kill -15 ${KUBECTL_PORT_FORWARD_PID}


######### Run the Tests ########

header "Running e2e tests"
# by default runs `go test -tags=e2e`
go_test_e2e -timeout=3m ./test/... || failed=1

(( failed )) && fail_test
success
