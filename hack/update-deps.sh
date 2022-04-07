#!/usr/bin/env bash

# Copyright 2019 The Knative Authors
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

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)

source "$REPO_ROOT/vendor/knative.dev/hack/library.sh"

# This is a copy/paste of knative's library.sh function. We've added
# this to force the `go mod tidy` command to use the -compat=1.17 flag.
# Without this the call to go_update_deps doesn't seem to do anything
# but print an error.
#
# Update go deps.
# Parameters (parsed as flags):
#   "--upgrade", bool, do upgrade.
#   "--release <release-version>" used with upgrade. The release version to upgrade
#                         Knative components. ex: --release v0.18. Defaults to
#                         "main".
#   "--module-release <module-version>" used to define a different go module tag
#                         for a release. ex: --release v1.0 --module-release v0.27
# Additional dependencies can be included in the upgrade by providing them in a
# global env var: FLOATING_DEPS
# --upgrade will set GOPROXY to direct unless it is already set.
function go_update_deps() {
  cd "${REPO_ROOT_DIR}" || return 1

  export GO111MODULE=on
  export GOFLAGS=""
  export GONOSUMDB="${GONOSUMDB:-},knative.dev/*"
  export GONOPROXY="${GONOPROXY:-},knative.dev/*"

  echo "=== Update Deps for Golang"

  local UPGRADE=0
  local RELEASE="v9000.1" # release v9000 is so far in the future, it will always pick the default branch.
  local RELEASE_MODULE=""
  local DOMAIN="knative.dev"
  while [[ $# -ne 0 ]]; do
    parameter=$1
    case ${parameter} in
      --upgrade) UPGRADE=1 ;;
      --release) shift; RELEASE="$1" ;;
      --module-release) shift; RELEASE_MODULE="$1" ;;
      --domain) shift; DOMAIN="$1" ;;
      *) abort "unknown option ${parameter}" ;;
    esac
    shift
  done

  if [[ $UPGRADE == 1 ]]; then
    local buoyArgs=(--release ${RELEASE}  --domain ${DOMAIN})
    if [ -n "$RELEASE_MODULE" ]; then
      group "Upgrading for release ${RELEASE} to release module version ${RELEASE_MODULE}"
      buoyArgs+=(--module-release ${RELEASE_MODULE})
    else
      group "Upgrading to release ${RELEASE}"
    fi
    FLOATING_DEPS+=( $(run_go_tool knative.dev/test-infra/buoy buoy float ${REPO_ROOT_DIR}/go.mod "${buoyArgs[@]}") )
    if [[ ${#FLOATING_DEPS[@]} > 0 ]]; then
      echo "Floating deps to ${FLOATING_DEPS[@]}"
      go get -d ${FLOATING_DEPS[@]}
    else
      echo "Nothing to upgrade."
    fi
  fi

  group "Go mod tidy and vendor"

  # Prune modules.
  local orig_pipefail_opt=$(shopt -p -o pipefail)
  set -o pipefail
  go mod tidy -compat=1.17 2>&1 | grep -v "ignoring symlink" || true
  go mod vendor 2>&1 |  grep -v "ignoring symlink" || true
  eval "$orig_pipefail_opt"

  group "Removing unwanted vendor files"

  # Remove unwanted vendor files
  find vendor/ \( -name "OWNERS" \
    -o -name "OWNERS_ALIASES" \
    -o -name "BUILD" \
    -o -name "BUILD.bazel" \
    -o -name "*_test.go" \) -exec rm -f {} +

  export GOFLAGS=-mod=vendor

  group "Updating licenses"
  update_licenses third_party/VENDOR-LICENSE "./..."

  group "Removing broken symlinks"
  remove_broken_symlinks ./vendor
}

go_update_deps "$@"
