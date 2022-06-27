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

package hub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/tektoncd/resolution/pkg/common"
	"github.com/tektoncd/resolution/pkg/resolver/framework"
)

// LabelValueHubResolverType is the value to use for the
// resolution.tekton.dev/type label on resource requests
const LabelValueHubResolverType string = "hub"

const defaultCatalog string = "Tekton"

// Resolver implements a framework.Resolver that can fetch files from OCI bundles.
type Resolver struct {
	// HubURL is the URL for hub resolver
	HubURL string
}

// Initialize sets up any dependencies needed by the resolver. None atm.
func (r *Resolver) Initialize(context.Context) error {
	return nil
}

// GetName returns a string name to refer to this resolver by.
func (r *Resolver) GetName(context.Context) string {
	return "Hub"
}

// GetSelector returns a map of labels to match requests to this resolver.
func (r *Resolver) GetSelector(context.Context) map[string]string {
	return map[string]string{
		common.LabelKeyResolverType: LabelValueHubResolverType,
	}
}

// ValidateParams ensures parameters from a request are as expected.
func (r *Resolver) ValidateParams(ctx context.Context, params map[string]string) error {
	if kind, ok := params[ParamKind]; !ok {
		return errors.New("must include kind param")
	} else if kind != "task" && kind != "pipeline" {
		return errors.New("kind param must be task or pipeline")
	}
	if _, ok := params[ParamName]; !ok {
		return errors.New("must include name param")
	}
	if _, ok := params[ParamVersion]; !ok {
		return errors.New("must include version param")
	}

	return nil
}

type dataResponse struct {
	YAML string `json:"yaml"`
}

type hubResponse struct {
	Data dataResponse `json:"data"`
}

// Resolve uses the given params to resolve the requested file or resource.
func (r *Resolver) Resolve(ctx context.Context, params map[string]string) (framework.ResolvedResource, error) {
	if _, ok := params[ParamCatalog]; !ok {
		params[ParamCatalog] = defaultCatalog
	}

	url := fmt.Sprintf(r.HubURL, params[ParamCatalog], params[ParamKind], params[ParamName], params[ParamVersion])
	// #nosec G107 -- URL cannot be constant in this case.
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error requesting resource from hub: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	hr := hubResponse{}
	err = json.Unmarshal(body, &hr)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json response: %w", err)
	}
	return &ResolvedHubResource{
		Content: []byte(hr.Data.YAML),
	}, nil
}

// ResolvedHubResource wraps the data we want to return to Pipelines
type ResolvedHubResource struct {
	Content []byte
}

var _ framework.ResolvedResource = &ResolvedHubResource{}

// Data returns the bytes of our hard-coded Pipeline
func (rr *ResolvedHubResource) Data() []byte {
	return rr.Content
}

// Annotations returns any metadata needed alongside the data. None atm.
func (*ResolvedHubResource) Annotations() map[string]string {
	return nil
}
