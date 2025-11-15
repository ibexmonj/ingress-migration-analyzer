package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client with connection info
type Client struct {
	Clientset      *kubernetes.Clientset
	Config         *rest.Config
	ClusterVersion string
	Context        string
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath, contextName string) (*Client, error) {
	config, err := loadConfig(kubeconfigPath, contextName)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	client := &Client{
		Clientset: clientset,
		Config:    config,
		Context:   contextName,
	}

	// Verify connection and get cluster version
	if err := client.verifyConnection(); err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	return client, nil
}

// loadConfig loads the Kubernetes configuration from kubeconfig
func loadConfig(kubeconfigPath, contextName string) (*rest.Config, error) {
	// First try in-cluster config if no kubeconfig path is provided
	if kubeconfigPath == "" {
		if config, err := rest.InClusterConfig(); err == nil {
			return config, nil
		}
	}

	// Fall back to kubeconfig file
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
	}

	// Load the kubeconfig file
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)

	config, err := configLoader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client config: %w", err)
	}

	return config, nil
}

// getDefaultKubeconfigPath returns the default kubeconfig path
func getDefaultKubeconfigPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// verifyConnection tests the connection and retrieves cluster info
func (c *Client) verifyConnection() error {
	ctx := context.Background()

	// Get server version to verify connection
	serverVersion, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to get server version (check your connection and credentials): %w", err)
	}

	c.ClusterVersion = serverVersion.GitVersion

	// Test basic API access
	listOptions := metav1.ListOptions{Limit: 1}
	_, err = c.Clientset.CoreV1().Namespaces().List(ctx, listOptions)
	if err != nil {
		return fmt.Errorf("failed to list namespaces (check your RBAC permissions): %w", err)
	}

	return nil
}

// GetClusterInfo returns basic cluster information
func (c *Client) GetClusterInfo() (string, string, error) {
	return c.ClusterVersion, c.Context, nil
}

// ValidateConnection performs a basic connectivity check
func ValidateConnection(kubeconfigPath, contextName string) error {
	client, err := NewClient(kubeconfigPath, contextName)
	if err != nil {
		return err
	}

	version, context, err := client.GetClusterInfo()
	if err != nil {
		return err
	}

	fmt.Printf("✅ Connected to cluster (version: %s)\n", version)
	if context != "" {
		fmt.Printf("🎯 Using context: %s\n", context)
	}

	return nil
}

// ListAvailableContexts lists available contexts from kubeconfig
func ListAvailableContexts(kubeconfigPath string) ([]string, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	var contexts []string
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, nil
}

// GetCurrentContext returns the current context from kubeconfig
func GetCurrentContext(kubeconfigPath string) (string, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = getDefaultKubeconfigPath()
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config.CurrentContext, nil
}
