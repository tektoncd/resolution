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

package main

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	resolutioncommon "github.com/tektoncd/resolution/pkg/common"
	ttesting "github.com/tektoncd/resolution/pkg/reconciler/testing"
	frtesting "github.com/tektoncd/resolution/pkg/resolver/framework/testing"
	"github.com/tektoncd/resolution/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "knative.dev/pkg/system/testing"
)

func TestResolver(t *testing.T) {
	ctx, _ := ttesting.SetupFakeContext(t)

	r := &resolver{}

	request := &v1alpha1.ResolutionRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "resolution.tekton.dev/v1alpha1",
			Kind:       "ResolutionRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "rr",
			Namespace:         "foo",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				resolutioncommon.LabelKeyResolverType: "demo",
			},
		},
		Spec: v1alpha1.ResolutionRequestSpec{},
	}
	d := test.Data{
		ResolutionRequests: []*v1alpha1.ResolutionRequest{request},
	}

	expectedStatus := &v1alpha1.ResolutionRequestStatus{
		ResolutionRequestStatusFields: v1alpha1.ResolutionRequestStatusFields{
			Data: base64.StdEncoding.Strict().EncodeToString([]byte(pipeline)),
		},
	}

	// If you want to test scenarios where an error should occur, pass a non-nil error to RunResolverReconcileTest
	var expectedErr error

	frtesting.RunResolverReconcileTest(ctx, t, d, r, request, expectedStatus, expectedErr)
}
