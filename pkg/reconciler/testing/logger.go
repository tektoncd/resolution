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

package testing

import (
	"context"
	"testing"

	"github.com/tektoncd/resolution/pkg/apis/resolution/v1alpha1"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	filteredinformerfactory "knative.dev/pkg/client/injection/kube/informers/factory/filtered"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	logtesting "knative.dev/pkg/logging/testing"

	// Import for creating fake filtered factory in the test
	_ "knative.dev/pkg/client/injection/kube/informers/factory/filtered/fake"
)

// SetupFakeContext sets up the Context and the fake filtered informers for the tests.
func SetupFakeContext(t *testing.T) (context.Context, []controller.Informer) {
	ctx, _, informer := setupFakeContextWithLabelKey(t)
	return WithLogger(ctx, t), informer
}

// WithLogger returns the Logger
func WithLogger(ctx context.Context, t *testing.T) context.Context {
	return logging.WithLogger(ctx, TestLogger(t))
}

// TestLogger sets up the Logger
func TestLogger(t *testing.T) *zap.SugaredLogger {
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.AddCaller()))
	return logger.Sugar().Named(t.Name())
}

// setupFakeContextWithLabelKey sets up the Context and the fake informers for the tests
// The provided context includes the FilteredInformerFactory LabelKey.
func setupFakeContextWithLabelKey(t zaptest.TestingT) (context.Context, context.CancelFunc, []controller.Informer) {
	ctx, c := context.WithCancel(logtesting.TestContextWithLogger(t))
	ctx = controller.WithEventRecorder(ctx, record.NewFakeRecorder(1000))
	ctx = filteredinformerfactory.WithSelectors(ctx, v1alpha1.ManagedByLabelKey)
	ctx, is := injection.Fake.SetupInformers(ctx, &rest.Config{})
	return ctx, c, is
}
