package kubevirt

import (
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"kubevirt.io/client-go/kubevirt"
)

// KubevirtClient holds the Kubevirt clientset and configuration
type KubevirtClient struct {
	clientset *kubevirt.Clientset
	config    *rest.Config
}

// NewKubevirtClient creates and returns a new KubevirtClient, establishing the connection
func NewKubevirtClient(opts *Options) (*KubevirtClient, error) {
	// Convert options to REST config
	config, err := opts.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config: %w", err)
	}

	// Create the Kubevirt clientset
	clientset, err := kubevirt.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubevirt clientset: %w", err)
	}

	// Return the new KubevirtClient
	return &KubevirtClient{
		clientset: clientset,
		config:    config,
	}, nil
}

// GetClientset returns the Kubevirt clientset
func (c *KubevirtClient) GetClientset() *kubevirt.Clientset {
	return c.clientset
}

// GetConfig returns the Kubernetes REST config
func (c *KubevirtClient) GetConfig() *rest.Config {
	return c.config
}

// Ping checks if the Kubevirt API server is reachable
func (c *KubevirtClient) Ping(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to ping Kubevirt API server: %w", err)
	}
	return nil
}
