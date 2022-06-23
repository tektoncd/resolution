package bundle

import (
	"context"
	"testing"

	resolutioncommon "github.com/tektoncd/resolution/pkg/common"
)

func TestGetSelector(t *testing.T) {
	resolver := Resolver{}
	sel := resolver.GetSelector(context.Background())
	if typ, has := sel[resolutioncommon.LabelKeyResolverType]; !has {
		t.Fatalf("unexpected selector: %v", sel)
	} else if typ != LabelValueBundleResolverType {
		t.Fatalf("unexpected type: %q", typ)
	}
}

func TestValidateParams(t *testing.T) {
	resolver := Resolver{}

	paramsWithTask := map[string]string{
		ParamKind:           "task",
		ParamName:           "foo",
		ParamBundle:         "bar",
		ParamServiceAccount: "baz",
	}
	if err := resolver.ValidateParams(context.Background(), paramsWithTask); err != nil {
		t.Fatalf("unexpected error validating params: %v", err)
	}

	paramsWithPipeline := map[string]string{
		ParamKind:           "pipeline",
		ParamName:           "foo",
		ParamBundle:         "bar",
		ParamServiceAccount: "baz",
	}
	if err := resolver.ValidateParams(context.Background(), paramsWithPipeline); err != nil {
		t.Fatalf("unexpected error validating params: %v", err)
	}
}

func TestValidateParamsMissing(t *testing.T) {
	resolver := Resolver{}

	var err error

	paramsMissingBundle := map[string]string{
		ParamKind:           "pipeline",
		ParamName:           "foo",
		ParamServiceAccount: "baz",
	}
	err = resolver.ValidateParams(context.Background(), paramsMissingBundle)
	if err == nil {
		t.Fatalf("expected missing kind err")
	}

	paramsMissingName := map[string]string{
		ParamKind:           "pipeline",
		ParamBundle:         "bar",
		ParamServiceAccount: "baz",
	}
	err = resolver.ValidateParams(context.Background(), paramsMissingName)
	if err == nil {
		t.Fatalf("expected missing name err")
	}

}
