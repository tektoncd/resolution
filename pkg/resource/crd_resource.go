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

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"
	rrclient "github.com/tektoncd/resolution/pkg/client/clientset/versioned"
	rrlisters "github.com/tektoncd/resolution/pkg/client/listers/resolution/v1alpha1"
	resolutioncommon "github.com/tektoncd/resolution/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// CRDRequester implements the Requester interface using
// ResourceRequest CRDs.
type CRDRequester struct {
	clientset rrclient.Interface
	lister    rrlisters.ResourceRequestLister
}

// NewCRDRequester returns an implementation of Requester that uses
// ResourceRequest CRD objects to mediate between the caller who wants a
// resource (e.g. Tekton Pipelines) and the responder who can fetch
// it (e.g. the gitresolver)
func NewCRDRequester(clientset rrclient.Interface, lister rrlisters.ResourceRequestLister) *CRDRequester {
	return &CRDRequester{clientset, lister}
}

var _ Requester = &CRDRequester{}

// Submit constructs a ResourceRequest object and submits it to the
// kubernetes cluster, returning any errors experienced while doing so.
// If ResourceRequest is succeeded then it returns the resolved data.
func (r *CRDRequester) Submit(ctx context.Context, resolver ResolverName, req Request) (ResolvedResource, error) {
	rr, _ := r.lister.ResourceRequests(req.Namespace()).Get(req.Name())
	if rr == nil {
		if err := r.createResourceRequest(ctx, resolver, req); err != nil {
			return nil, err
		}
		return nil, resolutioncommon.ErrorRequestInProgress
	}

	if rr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown() {
		// TODO(sbwsg): This should be where an existing
		// resource is given an additional owner reference so
		// that it doesn't get deleted until the caller is done
		// with it. Use appendOwnerReference and then submit
		// update to ResourceRequest.
		return nil, resolutioncommon.ErrorRequestInProgress
	}

	if rr.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
		return crdIntoResource(rr), nil
	}

	message := rr.Status.GetCondition(apis.ConditionSucceeded).GetMessage()
	err := resolutioncommon.NewError(resolutioncommon.ReasonResolutionFailed, errors.New(message))
	return nil, err
}

func (r *CRDRequester) createResourceRequest(ctx context.Context, resolver ResolverName, req Request) error {
	rr := &v1alpha1.ResourceRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "resolution.tekton.dev/v1alpha1",
			Kind:       "ResourceRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name(),
			Namespace: req.Namespace(),
			Labels: map[string]string{
				resolutioncommon.LabelKeyResolverType: string(resolver),
			},
		},
		Spec: v1alpha1.ResourceRequestSpec{
			Parameters: req.Params(),
		},
	}
	appendOwnerReference(rr, req)
	_, err := r.clientset.ResolutionV1alpha1().ResourceRequests(rr.Namespace).Create(ctx, rr, metav1.CreateOptions{})
	return err
}

func appendOwnerReference(rr *v1alpha1.ResourceRequest, req Request) {
	if ownedReq, ok := req.(OwnedRequest); ok {
		newOwnerRef := ownedReq.OwnerRef()
		isOwner := false
		for _, ref := range rr.ObjectMeta.OwnerReferences {
			if ownerRefsAreEqual(ref, newOwnerRef) {
				isOwner = true
			}
		}
		if !isOwner {
			rr.ObjectMeta.OwnerReferences = append(rr.ObjectMeta.OwnerReferences, newOwnerRef)
		}
	}
}

func ownerRefsAreEqual(a, b metav1.OwnerReference) bool {
	return a.APIVersion == b.APIVersion && a.Kind == b.Kind && a.Name == b.Name && a.UID == b.UID && a.Controller == b.Controller
}

// readOnlyResourceRequest is an opaque wrapper around ResourceRequest
// that provides the methods needed to read data from it using the
// Resource interface without exposing the underlying API
// object.
type readOnlyResourceRequest struct {
	req *v1alpha1.ResourceRequest
}

var _ ResolvedResource = readOnlyResourceRequest{}

func crdIntoResource(rr *v1alpha1.ResourceRequest) readOnlyResourceRequest {
	return readOnlyResourceRequest{req: rr}
}

func (r readOnlyResourceRequest) Annotations() map[string]string {
	status := r.req.GetStatus()
	if status != nil && status.Annotations != nil {
		annotationsCopy := map[string]string{}
		for key, val := range status.Annotations {
			annotationsCopy[key] = val
		}
		return annotationsCopy
	}
	return nil
}

func (r readOnlyResourceRequest) Data() ([]byte, error) {
	encodedData := r.req.Status.ResourceRequestStatusFields.Data
	decodedBytes, err := base64.StdEncoding.Strict().DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("error decoding data from base64: %w", err)
	}
	return decodedBytes, nil
}
