package v1alpha1

import (
	resolutioncommon "github.com/tektoncd/resolution/pkg/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

// ResourceRequests only have apis.ConditionSucceeded for now.
var resourceRequestCondSet = apis.NewBatchConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (*ResourceRequest) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ResourceRequest")
}

// GetConditionSet implements KRShaped.
func (*ResourceRequest) GetConditionSet() apis.ConditionSet {
	return resourceRequestCondSet
}

// HasStarted returns whether a ResourceRequests Status is considered to
// be in-progress.
func (rr *ResourceRequest) HasStarted() bool {
	return rr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown()
}

// IsDone returns whether a ResourceRequests Status is considered to be
// in a completed state, independent of success/failure.
func (rr *ResourceRequest) IsDone() bool {
	finalStateIsUnknown := rr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown()
	return !finalStateIsUnknown
}

// InitializeConditions set ths initial values of the conditions.
func (s *ResourceRequestStatus) InitializeConditions() {
	resourceRequestCondSet.Manage(s).InitializeConditions()
}

// MarkFailed sets the Succeeded condition to False with an accompanying
// error message.
func (s *ResourceRequestStatus) MarkFailed(reason, message string) {
	resourceRequestCondSet.Manage(s).MarkFalse(apis.ConditionSucceeded, reason, message)
}

// MarkSucceeded sets the Succeeded condition to True.
func (s *ResourceRequestStatus) MarkSucceeded() {
	resourceRequestCondSet.Manage(s).MarkTrue(apis.ConditionSucceeded)
}

// MarkInProgress updates the Succeeded condition to Unknown with an
// accompanying message.
func (s *ResourceRequestStatus) MarkInProgress(message string) {
	resourceRequestCondSet.Manage(s).MarkUnknown(apis.ConditionSucceeded, resolutioncommon.ReasonResolutionInProgress, message)
}
