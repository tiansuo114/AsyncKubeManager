package kubevirt

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

// Options holds the configuration options for connecting to Kubevirt
type Options struct {
	KubeConfigPath string
	KubeContext    string
	InCluster      bool
}

// DefaultOption is a function that modifies the Options
type DefaultOption func(*Options)

// SetDefaultKubeConfigPath sets the default kubeconfig path
func SetDefaultKubeConfigPath(s string) DefaultOption {
	return func(o *Options) {
		o.KubeConfigPath = s
	}
}

// SetDefaultKubeContext sets the default Kubernetes context
func SetDefaultKubeContext(s string) DefaultOption {
	return func(o *Options) {
		o.KubeContext = s
	}
}

// SetDefaultKubeInCluster sets the default in-cluster configuration flag
func SetDefaultKubeInCluster(b bool) DefaultOption {
	return func(o *Options) {
		o.InCluster = b
	}
}

// NewKubeOptions creates a new Options with default values
func NewKubeOptions(opts ...DefaultOption) *Options {
	// Default kubeconfig path is ~/.kube/config
	defaultKubeConfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeConfig = filepath.Join(home, ".kube", "config")
	}

	o := &Options{
		KubeConfigPath: defaultKubeConfig,
		KubeContext:    "", // Default context is empty, which means use the current context
		InCluster:      false,
	}

	// Modify the configuration using DefaultOption
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// ToRESTConfig converts the Options to a REST config
func (o *Options) ToRESTConfig() (*rest.Config, error) {
	if o.InCluster {
		return rest.InClusterConfig()
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: o.KubeConfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: o.KubeContext},
	).ClientConfig()
}
