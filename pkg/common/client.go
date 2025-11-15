package common

import (
	"fmt"

	"ingress-migration-analyzer/pkg/discovery"
)

// ClientConfig holds configuration for creating analyzer clients
type ClientConfig struct {
	Kubeconfig  string
	ContextName string
	Namespace   string
	Output      string
	Format      string
}

// CreateAnalyzerClient creates a Kubernetes client with validation.
// It first validates the connection and context, then creates the client.
// This function consolidates shared client creation logic across commands.
func CreateAnalyzerClient(kubeconfig, contextName string) (*discovery.Client, error) {
	// Test Kubernetes connection first
	if err := discovery.ValidateConnection(kubeconfig, contextName); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// Create Kubernetes client
	client, err := discovery.NewClient(kubeconfig, contextName)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// ValidateCommonFlags validates common flags used by both scan and inventory commands
func ValidateCommonFlags(output, format string) error {
	// This function can be extended with common validation logic
	// Currently, individual commands handle their own validation
	// but this provides a place for shared validation in the future
	return nil
}