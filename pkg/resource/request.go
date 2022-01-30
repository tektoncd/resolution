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

package resource

var _ Request = &BasicRequest{}

// BasicRequest holds the fields needed to submit a new resource request.
type BasicRequest struct {
	name      string
	namespace string
	params    map[string]string
}

func NewRequest(name, namespace string, params map[string]string) Request {
	return &BasicRequest{name, namespace, params}
}

var _ Request = &BasicRequest{}

func (req *BasicRequest) Name() string {
	return req.name
}

func (req *BasicRequest) Namespace() string {
	return req.namespace
}

func (req *BasicRequest) Params() map[string]string {
	return req.params
}
