package git

import (
	"context"
	"testing"
	"time"

	resolutioncommon "github.com/tektoncd/resolution/pkg/common"
	"github.com/tektoncd/resolution/pkg/resolver/framework"
)

func TestGetSelector(t *testing.T) {
	resolver := Resolver{}
	sel := resolver.GetSelector(context.Background())
	if typ, has := sel[resolutioncommon.LabelKeyResolverType]; !has {
		t.Fatalf("unexpected selector: %v", sel)
	} else if typ != LabelValueGitResolverType {
		t.Fatalf("unexpected type: %q", typ)
	}
}

func TestValidateParams(t *testing.T) {
	resolver := Resolver{}

	paramsWithCommit := map[string]string{
		URLParam:    "foo",
		PathParam:   "bar",
		CommitParam: "baz",
	}
	if err := resolver.ValidateParams(context.Background(), paramsWithCommit); err != nil {
		t.Fatalf("unexpected error validating params: %v", err)
	}

	paramsWithBranch := map[string]string{
		URLParam:    "foo",
		PathParam:   "bar",
		BranchParam: "baz",
	}
	if err := resolver.ValidateParams(context.Background(), paramsWithBranch); err != nil {
		t.Fatalf("unexpected error validating params: %v", err)
	}
}

func TestValidateParamsMissing(t *testing.T) {
	resolver := Resolver{}

	var err error

	paramsMissingURL := map[string]string{
		PathParam:   "bar",
		CommitParam: "baz",
	}
	err = resolver.ValidateParams(context.Background(), paramsMissingURL)
	if err == nil {
		t.Fatalf("expected missing url err")
	}

	paramsMissingPath := map[string]string{
		URLParam:    "foo",
		BranchParam: "baz",
	}
	err = resolver.ValidateParams(context.Background(), paramsMissingPath)
	if err == nil {
		t.Fatalf("expected missing path err")
	}
}

func TestValidateParamsConflictingGitRef(t *testing.T) {
	resolver := Resolver{}
	params := map[string]string{
		URLParam:    "foo",
		PathParam:   "bar",
		CommitParam: "baz",
		BranchParam: "quux",
	}
	err := resolver.ValidateParams(context.Background(), params)
	if err == nil {
		t.Fatalf("expected err due to conflicting commit and branch params")
	}
}

func TestGetResolutionTimeoutDefault(t *testing.T) {
	resolver := Resolver{}
	defaultTimeout := 30 * time.Minute
	timeout := resolver.GetResolutionTimeout(context.Background(), defaultTimeout)
	if timeout != defaultTimeout {
		t.Fatalf("expected default timeout to be returned")
	}
}

func TestGetResolutionTimeoutCustom(t *testing.T) {
	resolver := Resolver{}
	defaultTimeout := 30 * time.Minute
	configTimeout := 5 * time.Second
	config := map[string]string{
		ConfigFieldTimeout: configTimeout.String(),
	}
	ctx := framework.InjectResolverConfigToContext(context.Background(), config)
	timeout := resolver.GetResolutionTimeout(ctx, defaultTimeout)
	if timeout != configTimeout {
		t.Fatalf("expected timeout from config to be returned")
	}
}
