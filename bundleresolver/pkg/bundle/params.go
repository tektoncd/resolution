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

package bundle

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
)

// ParamServiceAccount is the parameter defining what service
// account name to use for bundle requests.
const ParamServiceAccount = "serviceAccount"

// ParamBundle is the parameter defining what the bundle image url is.
const ParamBundle = "bundle"

// ParamName is the parameter defining what the layer name in the bundle
// image is.
const ParamName = "name"

// ParamKind is the parameter defining what the layer kind in the bundle
// image is.
const ParamKind = "kind"

// OptionsFromParams parses the params from a resolution request and
// converts them into options to pass as part of a bundle request.
func OptionsFromParams(params map[string]string) (RequestOptions, error) {
	opts := RequestOptions{}

	sa, ok := params[ParamServiceAccount]
	if !ok {
		return opts, fmt.Errorf("parameter %q required", ParamServiceAccount)
	}

	bundle, ok := params[ParamBundle]
	if !ok {
		return opts, fmt.Errorf("parameter %q required", ParamBundle)
	}
	if _, err := name.ParseReference(bundle); err != nil {
		return opts, fmt.Errorf("invalid bundle reference: %w", err)
	}

	name, ok := params[ParamName]
	if !ok {
		return opts, fmt.Errorf("parameter %q required", ParamName)
	}

	kind, ok := params[ParamKind]
	if !ok {
		return opts, fmt.Errorf("paramater %q required", ParamKind)
	}

	opts.ServiceAccount = sa
	opts.Bundle = bundle
	opts.EntryName = name
	opts.Kind = kind

	return opts, nil
}
