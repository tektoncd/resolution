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

package helpers

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// ReadFileToObject reads a file given a path relative to PWD and attempts
// to parse it into the given k8s runtime object.
func ReadFileToObject(filename string, obj runtime.Object) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %q: %w", filename, err)
	}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(content, nil, obj)
	if err != nil {
		return fmt.Errorf("error parsing %q: %w", filename, err)
	}
	return nil
}
