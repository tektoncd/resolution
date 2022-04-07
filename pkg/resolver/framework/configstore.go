package framework

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
)

// resolverConfigKey is the contenxt key associated with configuration
// for one specific resolver, and is only used if that resolver
// implements the optional framework.ConfigWatcher interface.
var resolverConfigKey = struct{}{}

// DataFromConfigMap returns a copy of the contents of a configmap or an
// empty map if the configmap doesn't have any data.
func DataFromConfigMap(config *corev1.ConfigMap) (map[string]string, error) {
	resolverConfig := map[string]string{}
	if config == nil {
		return resolverConfig, nil
	}
	for key, value := range config.Data {
		resolverConfig[key] = value
	}
	return resolverConfig, nil
}

// ConfigStore wraps a knative untyped store and provides helper methods
// for working with a resolver's configuration data.
type ConfigStore struct {
	resolverConfigName string
	untyped            *configmap.UntypedStore
}

// GetResolverConfig returns a copy of the resolver's current
// configuration or an empty map if the stored config is nil or invalid.
func (store *ConfigStore) GetResolverConfig() map[string]string {
	resolverConfig := map[string]string{}
	untypedConf := store.untyped.UntypedLoad(store.resolverConfigName)
	if conf, ok := untypedConf.(map[string]string); ok {
		for key, val := range conf {
			resolverConfig[key] = val
		}
	}
	return resolverConfig
}

// ToContext returns a new context with the resolver's configuration
// data stored in it.
func (store *ConfigStore) ToContext(ctx context.Context) context.Context {
	conf := store.GetResolverConfig()
	return InjectResolverConfigToContext(ctx, conf)
}

// InjectResolverConfigToContext returns a new context with a
// map stored in it for a resolvers config.
func InjectResolverConfigToContext(ctx context.Context, conf map[string]string) context.Context {
	return context.WithValue(ctx, resolverConfigKey, conf)
}

// GetResolverConfigFromContext returns any resolver-specific
// configuration that has been stored or an empty map if none exists.
func GetResolverConfigFromContext(ctx context.Context) map[string]string {
	conf := map[string]string{}
	storedConfig := ctx.Value(resolverConfigKey)
	if resolverConfig, ok := storedConfig.(map[string]string); ok {
		conf = resolverConfig
	}
	return conf
}
