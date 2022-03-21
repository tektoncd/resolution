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

local failed=0

header "Deploying Tekton Resolution"
ko apply -f ./config

header "Deploying Git Resolver"
ko apply -f ./gitresolver/config

header "Deploying Resolver Template"
ko apply -f ./docs/resolver-template/config

wait_until_pods_running "tekton-remote-resolution" || fail_test "Tekton Resolution did not come up"

header "Running e2e tests"
# by default runs `go test -tags=e2e`
go_test_e2e -timeout=2m ./test/... || failed=1

(( failed )) && fail_test
success
