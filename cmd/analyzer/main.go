package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"ingress-migration-analyzer/pkg/analyze"
	"ingress-migration-analyzer/pkg/common"
	"ingress-migration-analyzer/pkg/report"
)

var (
	version = "0.1.0"
	kubeconfig string
	contextName string
	namespace string
	output string
	format string
)

var rootCmd = &cobra.Command{
	Use:   "analyzer",
	Short: "Ingress-NGINX Migration Analyzer",
	Long: `Analyze your ingress-nginx usage and plan your migration before the March 2026 EOL.

This tool scans Kubernetes clusters to identify ingress-nginx resources, 
classifies migration complexity, and generates actionable reports.`,
	Version: version,
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan cluster for ingress-nginx usage",
	Long: `Scan the Kubernetes cluster for ingress-nginx resources and generate
a migration complexity analysis report.

This command will:
- Connect to your Kubernetes cluster
- Discover all ingress-nginx resources
- Analyze annotation complexity
- Generate a detailed migration report`,
	RunE: runScan,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", getDefaultKubeconfig(), "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "Kubernetes context to use")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "Specific namespace to scan (default: all namespaces)")

	// Scan command flags
	scanCmd.Flags().StringVar(&output, "output", "./reports/", "Output directory for reports")
	scanCmd.Flags().StringVar(&format, "format", "markdown", "Output format (markdown|json)")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(inventoryCmd)
}

func getDefaultKubeconfig() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

func runScan(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔍 Starting ingress-nginx migration analysis...\n")
	fmt.Printf("📁 Output directory: %s\n", output)
	fmt.Printf("📄 Format: %s\n", format)
	
	if kubeconfig != "" {
		fmt.Printf("🔧 Kubeconfig: %s\n", kubeconfig)
	}
	if contextName != "" {
		fmt.Printf("🎯 Context: %s\n", contextName)
	}
	if namespace != "" {
		fmt.Printf("📦 Namespace: %s\n", namespace)
	} else {
		fmt.Printf("📦 Scanning all namespaces\n")
	}

	// Validate flags
	if err := validateFlags(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Create Kubernetes client with validation
	fmt.Println("\n🔌 Testing Kubernetes connection...")
	client, err := common.CreateAnalyzerClient(kubeconfig, contextName)
	if err != nil {
		return err
	}

	// Create analyzer and run analysis
	analyzer := analyze.NewAnalyzer(client, namespace)
	clusterAnalysis, err := analyzer.AnalyzeCluster(context.Background())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Generate report
	fmt.Println("\n📝 Generating report...")
	var reportPath string
	
	switch format {
	case "markdown":
		generator := report.NewMarkdownGenerator()
		generator.ContextName = contextName
		reportPath, err = generator.GenerateReport(clusterAnalysis, output)
	case "json":
		generator := report.NewJSONGenerator()
		reportPath, err = generator.GenerateReport(clusterAnalysis, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("✅ Analysis complete! Report saved to: %s\n", reportPath)
	
	if clusterAnalysis.Summary.HighRiskCount > 0 {
		fmt.Printf("\n⚠️  Warning: Found %d high-risk resources requiring careful migration planning\n", 
			clusterAnalysis.Summary.HighRiskCount)
	}
	
	return nil
}

func validateFlags() error {
	// Check if kubeconfig file exists
	if kubeconfig != "" {
		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			return fmt.Errorf("kubeconfig file not found: %s", kubeconfig)
		}
	}

	// Validate output format
	if format != "markdown" && format != "json" {
		return fmt.Errorf("invalid format '%s': must be 'markdown' or 'json'", format)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}