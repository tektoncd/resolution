package framework

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
	logtesting "knative.dev/pkg/logging/testing"
)

// TestDataFromConfigMap checks that configmaps are correctly converted
// into a map[string]string
func TestDataFromConfigMap(t *testing.T) {
	for _, tc := range []struct {
		configMap *corev1.ConfigMap
		expected  map[string]string
	}{{
		configMap: nil,
		expected:  map[string]string{},
	}, {
		configMap: &corev1.ConfigMap{
			Data: nil,
		},
		expected: map[string]string{},
	}, {
		configMap: &corev1.ConfigMap{
			Data: map[string]string{},
		},
		expected: map[string]string{},
	}, {
		configMap: &corev1.ConfigMap{
			Data: map[string]string{
				"foo": "bar",
			},
		},
		expected: map[string]string{
			"foo": "bar",
		},
	}} {
		out, err := DataFromConfigMap(tc.configMap)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !mapsAreEqual(tc.expected, out) {
			t.Fatalf("expected %#v received %#v", tc.expected, out)
		}
	}
}

func TestGetResolverConfig(t *testing.T) {
	_ = &ConfigStore{
		resolverConfigName: "test",
		untyped: configmap.NewUntypedStore(
			"test-config",
			logtesting.TestLogger(t),
			configmap.Constructors{
				"test": DataFromConfigMap,
			},
		),
	}
}

func mapsAreEqual(m1, m2 map[string]string) bool {
	if m1 == nil || m2 == nil {
		return m1 == nil && m2 == nil
	}
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}
	return true
}
