package v1alpha1

import "context"

// SetDefaults walks a ResourceRequest object and sets any default
// values that are required to be set before a reconciler sees it.
func (rr *ResourceRequest) SetDefaults(ctx context.Context) {
	if rr.TypeMeta.Kind == "" {
		rr.TypeMeta.Kind = "ResourceRequest"
	}
	if rr.TypeMeta.APIVersion == "" {
		rr.TypeMeta.APIVersion = "resolution.tekton.dev/v1alpha1"
	}
}
