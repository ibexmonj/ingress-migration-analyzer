package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"ingress-migration-analyzer/internal/models"
	"ingress-migration-analyzer/pkg/analyze"
	"ingress-migration-analyzer/pkg/common"
	"ingress-migration-analyzer/pkg/report"
)

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Generate detailed annotation inventory and usage analysis",
	Long: `Generate a comprehensive inventory of all annotations used in your cluster.

This command provides detailed analysis beyond basic risk classification:
- Complete annotation usage statistics
- Value frequency analysis
- Cross-namespace usage patterns
- Unknown annotation identification
- Migration complexity heat map

This is particularly useful for:
- Understanding the full scope of nginx annotations in use
- Planning annotation-by-annotation migration strategies
- Identifying the most critical annotations to address first
- Creating comprehensive migration documentation`,
	RunE: runInventory,
}

func init() {
	// Add inventory flags
	inventoryCmd.Flags().BoolP("detailed", "d", false, "Include detailed value analysis")
	inventoryCmd.Flags().StringP("sort", "s", "usage", "Sort by: usage, risk, namespace, name")
	inventoryCmd.Flags().IntP("top", "t", 10, "Show top N most used annotations")
	inventoryCmd.Flags().StringVar(&output, "output", "./reports/", "Output directory for reports")
	inventoryCmd.Flags().StringVar(&format, "format", "json", "Output format (json recommended for inventory data)")
}

func runInventory(cmd *cobra.Command, args []string) error {
	detailed, _ := cmd.Flags().GetBool("detailed")
	sortBy, _ := cmd.Flags().GetString("sort")
	topN, _ := cmd.Flags().GetInt("top")

	fmt.Printf("📋 Generating annotation inventory...\n")
	fmt.Printf("📁 Output directory: %s\n", output)
	fmt.Printf("🔧 Detailed analysis: %v\n", detailed)
	fmt.Printf("📊 Sort by: %s\n", sortBy)
	fmt.Printf("🔝 Top N: %d\n", topN)

	if kubeconfig != "" {
		fmt.Printf("🔧 Kubeconfig: %s\n", kubeconfig)
	}
	if contextName != "" {
		fmt.Printf("🎯 Context: %s\n", contextName)
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

	// Build detailed inventory
	fmt.Println("\n📊 Building annotation inventory...")
	inventory := analyze.BuildAnnotationInventory(clusterAnalysis.Analyses)

	// Print console summary
	printInventorySummary(inventory, topN)

	// Add inventory to cluster analysis
	clusterAnalysis.Inventory = inventory

	// Generate detailed inventory report
	fmt.Println("\n📝 Generating inventory report...")
	var reportPath string

	switch format {
	case "markdown":
		generator := &InventoryMarkdownGenerator{
			Detailed:    detailed,
			SortBy:      sortBy,
			TopN:        topN,
			ContextName: contextName,
		}
		reportPath, err = generator.GenerateInventoryReport(inventory, clusterAnalysis, output)
	case "json":
		// JSON includes full inventory data automatically
		generator := report.NewJSONGenerator()
		reportPath, err = generator.GenerateReport(clusterAnalysis, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate inventory report: %w", err)
	}

	fmt.Printf("✅ Inventory analysis complete! Report saved to: %s\n", reportPath)

	return nil
}

func printInventorySummary(inventory *analyze.AnnotationInventory, topN int) {
	fmt.Println("\n📈 Annotation Inventory Summary:")
	fmt.Printf("   Total Unique Annotations: %d\n", inventory.Summary.TotalUniqueAnnotations)
	fmt.Printf("   Nginx Annotations: %d\n", inventory.Summary.NginxAnnotationsCount)
	fmt.Printf("   Unknown Nginx Annotations: %d\n", inventory.Summary.UnknownAnnotationsCount)

	if inventory.Summary.MostUsedAnnotation != "" {
		fmt.Printf("   Most Used: %s\n", inventory.Summary.MostUsedAnnotation)
	}

	// Show most critical annotations
	critical := inventory.GetMostCriticalAnnotations(topN)
	if len(critical) > 0 {
		fmt.Println("\n🚨 Most Critical Annotations (for migration):")
		for i, annotation := range critical {
			if i >= topN {
				break
			}
			fmt.Printf("   %d. %s (used %d times across %d namespaces)\n",
				i+1, annotation.Key, annotation.UsageCount, len(annotation.Namespaces))
		}
	}

	// Show annotations by risk
	byRisk := inventory.GetAnnotationsByRisk()
	if len(byRisk) > 0 {
		fmt.Println("\n📊 Annotations by Risk Level:")
		for riskLevel, annotations := range byRisk {
			fmt.Printf("   %s: %d annotations\n", riskLevel, len(annotations))
		}
	}
}

// InventoryMarkdownGenerator creates detailed inventory reports
type InventoryMarkdownGenerator struct {
	Detailed    bool
	SortBy      string
	TopN        int
	ContextName string
}

func (g *InventoryMarkdownGenerator) GenerateInventoryReport(inventory *analyze.AnnotationInventory, analysis *models.ClusterAnalysis, outputDir string) (string, error) {
	// Generate the inventory report content
	content := g.generateInventoryContent(inventory, analysis)

	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	filename := fmt.Sprintf("annotation-inventory-%s.md", timestamp)
	filePath := filepath.Join(outputDir, filename)

	// Write to file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write inventory report: %w", err)
	}

	return filePath, nil
}

// generateInventoryContent creates the comprehensive inventory markdown content
func (g *InventoryMarkdownGenerator) generateInventoryContent(inventory *analyze.AnnotationInventory, analysis *models.ClusterAnalysis) string {
	var content strings.Builder

	// Header
	g.writeInventoryHeader(&content, inventory, analysis)

	// Executive Summary
	g.writeInventorySummary(&content, inventory)

	// Critical Annotations (High Priority)
	g.writeCriticalAnnotations(&content, inventory)

	// Annotations by Risk Level
	g.writeAnnotationsByRisk(&content, inventory)

	// Unknown Annotations Analysis
	if len(inventory.UnknownAnnotations) > 0 {
		g.writeUnknownAnnotations(&content, inventory)
	}

	// Detailed Annotation Usage (if requested)
	if g.Detailed {
		g.writeDetailedUsage(&content, inventory)
	}

	// Migration Strategies
	g.writeMigrationStrategies(&content, inventory)

	// Footer
	g.writeInventoryFooter(&content)

	return content.String()
}

// writeInventoryHeader writes the report header
func (g *InventoryMarkdownGenerator) writeInventoryHeader(content *strings.Builder, inventory *analyze.AnnotationInventory, analysis *models.ClusterAnalysis) {
	content.WriteString("# NGINX Ingress Annotation Inventory Report\n\n")
	content.WriteString(fmt.Sprintf("**Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	if g.ContextName != "" {
		content.WriteString(fmt.Sprintf("**Cluster Context**: %s\n", g.ContextName))
	}
	content.WriteString(fmt.Sprintf("**Cluster Version**: %s\n", analysis.ScanResult.ClusterVersion))
	content.WriteString(fmt.Sprintf("**Total Ingress Resources Scanned**: %d\n", analysis.ScanResult.TotalIngresses))
	content.WriteString(fmt.Sprintf("**Total Unique Annotations Found**: %d\n", inventory.Summary.TotalUniqueAnnotations))
	content.WriteString(fmt.Sprintf("**NGINX Annotations**: %d\n", inventory.Summary.NginxAnnotationsCount))
	content.WriteString("\n---\n\n")
}

// writeInventorySummary writes the executive summary
func (g *InventoryMarkdownGenerator) writeInventorySummary(content *strings.Builder, inventory *analyze.AnnotationInventory) {
	content.WriteString("## Executive Summary\n\n")
	
	if inventory.Summary.TotalUniqueAnnotations == 0 {
		content.WriteString("🎉 **No annotations found!** Your ingress resources are using minimal configuration.\n\n")
		return
	}

	content.WriteString(fmt.Sprintf("This cluster uses **%d unique annotations** across all ingress resources.\n", inventory.Summary.TotalUniqueAnnotations))
	content.WriteString(fmt.Sprintf("Of these, **%d are NGINX-specific** annotations that will need migration attention.\n\n", inventory.Summary.NginxAnnotationsCount))
	
	content.WriteString("ℹ️  **Note**: System annotations (like `kubectl.kubernetes.io/*`) are automatically filtered out from this analysis as they are not relevant for migration planning.\n\n")

	if inventory.Summary.UnknownAnnotationsCount > 0 {
		content.WriteString(fmt.Sprintf("⚠️  **%d unknown NGINX annotations** were found - these require immediate investigation.\n\n", inventory.Summary.UnknownAnnotationsCount))
	}

	if inventory.Summary.MostUsedAnnotation != "" {
		if usage, exists := inventory.AllAnnotations[inventory.Summary.MostUsedAnnotation]; exists {
			content.WriteString(fmt.Sprintf("📈 **Most frequently used annotation**: `%s` (used %d times across %d namespaces)\n\n", 
				inventory.Summary.MostUsedAnnotation, usage.UsageCount, len(usage.Namespaces)))
		}
	}

	content.WriteString("---\n\n")
}

// writeCriticalAnnotations highlights the most critical annotations for migration
func (g *InventoryMarkdownGenerator) writeCriticalAnnotations(content *strings.Builder, inventory *analyze.AnnotationInventory) {
	critical := inventory.GetMostCriticalAnnotations(g.TopN)
	
	if len(critical) == 0 {
		content.WriteString("## Critical Annotations\n\n")
		content.WriteString("✅ **No critical annotations found!** This is excellent news for your migration.\n\n")
		content.WriteString("---\n\n")
		return
	}

	content.WriteString("## Critical Annotations (Migration Priority)\n\n")
	content.WriteString("These annotations require immediate attention for migration planning:\n\n")

	content.WriteString("| Rank | Annotation | Usage Count | Namespaces | Risk Level | Description |\n")
	content.WriteString("|------|------------|-------------|------------|------------|-------------|\n")

	for i, annotation := range critical {
		if i >= g.TopN {
			break
		}
		riskIcon := g.getRiskIcon(annotation.Risk)
		content.WriteString(fmt.Sprintf("| %d | `%s` | %d | %d | %s %s | %s |\n",
			i+1, annotation.Key, annotation.UsageCount, len(annotation.Namespaces), 
			riskIcon, annotation.Risk, annotation.Description))
	}

	content.WriteString("\n---\n\n")
}

// writeAnnotationsByRisk groups annotations by risk level
func (g *InventoryMarkdownGenerator) writeAnnotationsByRisk(content *strings.Builder, inventory *analyze.AnnotationInventory) {
	content.WriteString("## Annotations by Migration Risk Level\n\n")

	byRisk := inventory.GetAnnotationsByRisk()
	
	// Process each risk level
	riskLevels := []models.RiskLevel{models.RiskAuto, models.RiskManual, models.RiskHigh}
	
	for _, riskLevel := range riskLevels {
		annotations := byRisk[riskLevel]
		if len(annotations) == 0 {
			continue
		}

		icon := g.getRiskIcon(riskLevel)
		content.WriteString(fmt.Sprintf("### %s %s (%d annotations)\n\n", icon, riskLevel, len(annotations)))

		// Sort by usage for this view
		g.sortAnnotations(annotations, "usage")

		for i, annotation := range annotations {
			if i >= common.MaxAnnotationsPerRiskLevel {
				content.WriteString(fmt.Sprintf("   ... and %d more\n", len(annotations)-common.MaxAnnotationsPerRiskLevel))
				break
			}
			content.WriteString(fmt.Sprintf("- `%s` - used %d times", annotation.Key, annotation.UsageCount))
			if annotation.MigrationNote != "" {
				content.WriteString(fmt.Sprintf(" → %s", annotation.MigrationNote))
			}
			if annotation.SourceURL != "" {
				content.WriteString(fmt.Sprintf(" ([docs](%s))", annotation.SourceURL))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString("---\n\n")
}

// writeUnknownAnnotations details unknown annotations that need research
func (g *InventoryMarkdownGenerator) writeUnknownAnnotations(content *strings.Builder, inventory *analyze.AnnotationInventory) {
	content.WriteString("## Unknown NGINX Annotations\n\n")
	content.WriteString("These annotations are not in our knowledge base and require manual research:\n\n")

	// Convert to slice for sorting
	var unknownList []*analyze.AnnotationUsage
	for _, usage := range inventory.UnknownAnnotations {
		unknownList = append(unknownList, usage)
	}
	g.sortAnnotations(unknownList, g.SortBy)

	content.WriteString("| Annotation | Usage Count | Namespaces | Example Values |\n")
	content.WriteString("|------------|-------------|------------|----------------|\n")

	for _, annotation := range unknownList {
		examples := g.formatValueExamples(annotation.ValueExamples, 3)
		content.WriteString(fmt.Sprintf("| `%s` | %d | %s | %s |\n",
			annotation.Key, annotation.UsageCount, 
			strings.Join(annotation.Namespaces, ", "), examples))
	}

	content.WriteString("\n**Action Required**: Research these annotations in the [NGINX documentation](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/) and determine Gateway API equivalents.\n\n")
	content.WriteString("---\n\n")
}

// writeDetailedUsage provides comprehensive usage analysis
func (g *InventoryMarkdownGenerator) writeDetailedUsage(content *strings.Builder, inventory *analyze.AnnotationInventory) {
	content.WriteString("## Detailed Usage Analysis\n\n")

	// Convert nginx annotations to slice for sorting
	var allNginx []*analyze.AnnotationUsage
	for _, usage := range inventory.NginxAnnotations {
		allNginx = append(allNginx, usage)
	}
	g.sortAnnotations(allNginx, g.SortBy)

	content.WriteString("### All NGINX Annotations\n\n")
	content.WriteString("| Annotation | Usage | Namespaces | Unique Values | Risk | Migration Note |\n")
	content.WriteString("|------------|-------|------------|---------------|------|----------------|\n")

	for _, annotation := range allNginx {
		riskIcon := g.getRiskIcon(annotation.Risk)
		content.WriteString(fmt.Sprintf("| `%s` | %d | %s | %d | %s | %s |\n",
			annotation.Key, annotation.UsageCount, 
			strings.Join(annotation.Namespaces, ", "), len(annotation.UniqueValues),
			riskIcon, annotation.MigrationNote))
	}

	content.WriteString("\n")

	// Value frequency analysis for top annotations
	content.WriteString("### Value Frequency Analysis\n\n")
	content.WriteString("Most common values for frequently used annotations:\n\n")

	topUsed := g.getTopAnnotations(allNginx, 5)
	for _, annotation := range topUsed {
		if len(annotation.ValueExamples) <= 1 {
			continue // Skip if only one value
		}

		content.WriteString(fmt.Sprintf("#### `%s`\n\n", annotation.Key))
		
		// Sort values by frequency
		type valueFreq struct {
			value string
			count int
		}
		var freqs []valueFreq
		for value, count := range annotation.ValueExamples {
			freqs = append(freqs, valueFreq{value, count})
		}
		sort.Slice(freqs, func(i, j int) bool {
			return freqs[i].count > freqs[j].count
		})

		for i, vf := range freqs {
			if i >= common.MaxValueFrequencyExamples {
				break
			}
			content.WriteString(fmt.Sprintf("- `%s`: %d times\n", vf.value, vf.count))
		}
		content.WriteString("\n")
	}

	content.WriteString("---\n\n")
}

// writeMigrationStrategies provides strategic guidance
func (g *InventoryMarkdownGenerator) writeMigrationStrategies(content *strings.Builder, inventory *analyze.AnnotationInventory) {
	content.WriteString("## Migration Strategy Recommendations\n\n")

	byRisk := inventory.GetAnnotationsByRisk()
	
	content.WriteString("### Phase 1: Auto-Migratable (Low Risk)\n")
	if autoAnnotations := byRisk[models.RiskAuto]; len(autoAnnotations) > 0 {
		content.WriteString(fmt.Sprintf("**%d annotations** can be migrated automatically:\n\n", len(autoAnnotations)))
		for i, annotation := range autoAnnotations {
			if i >= 5 {
				content.WriteString(fmt.Sprintf("... and %d more\n", len(autoAnnotations)-5))
				break
			}
			content.WriteString(fmt.Sprintf("- `%s` → %s\n", annotation.Key, annotation.MigrationNote))
		}
	} else {
		content.WriteString("No auto-migratable annotations found.\n")
	}
	content.WriteString("\n")

	content.WriteString("### Phase 2: Manual Review Required (Medium Risk)\n")
	if manualAnnotations := byRisk[models.RiskManual]; len(manualAnnotations) > 0 {
		content.WriteString(fmt.Sprintf("**%d annotations** require manual review:\n\n", len(manualAnnotations)))
		for i, annotation := range manualAnnotations {
			if i >= 5 {
				content.WriteString(fmt.Sprintf("... and %d more\n", len(manualAnnotations)-5))
				break
			}
			content.WriteString(fmt.Sprintf("- `%s` → %s\n", annotation.Key, annotation.MigrationNote))
		}
	} else {
		content.WriteString("No manual review annotations found.\n")
	}
	content.WriteString("\n")

	content.WriteString("### Phase 3: High Risk (Complex Migration)\n")
	if highRiskAnnotations := byRisk[models.RiskHigh]; len(highRiskAnnotations) > 0 {
		content.WriteString(fmt.Sprintf("**%d annotations** require complex migration planning:\n\n", len(highRiskAnnotations)))
		for i, annotation := range highRiskAnnotations {
			if i >= 5 {
				content.WriteString(fmt.Sprintf("... and %d more\n", len(highRiskAnnotations)-5))
				break
			}
			content.WriteString(fmt.Sprintf("- `%s` → %s\n", annotation.Key, annotation.MigrationNote))
		}
	} else {
		content.WriteString("No high-risk annotations found.\n")
	}
	content.WriteString("\n")

	content.WriteString("### Namespace-Specific Considerations\n\n")
	namespaceMap := g.analyzeNamespaceComplexity(inventory)
	if len(namespaceMap) > 1 {
		content.WriteString("Migration complexity by namespace:\n\n")
		for namespace, complexity := range namespaceMap {
			content.WriteString(fmt.Sprintf("- **%s**: %s\n", namespace, complexity))
		}
	}

	content.WriteString("\n---\n\n")
}

// writeInventoryFooter adds footer information
func (g *InventoryMarkdownGenerator) writeInventoryFooter(content *strings.Builder) {
	content.WriteString("## Additional Resources\n\n")
	content.WriteString("- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)\n")
	content.WriteString("- [NGINX Ingress Annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/)\n")
	content.WriteString("- [Migration Planning Guide](https://kubernetes.github.io/ingress-nginx/migration/)\n\n")
	content.WriteString("---\n\n")
	content.WriteString("*Generated by [Ingress-NGINX Migration Analyzer](https://github.com/user/ingress-migration-analyzer)*\n")
}

// Helper methods

func (g *InventoryMarkdownGenerator) getRiskIcon(risk models.RiskLevel) string {
	switch risk {
	case models.RiskAuto:
		return "✅"
	case models.RiskManual:
		return "⚠️"
	case models.RiskHigh:
		return "❌"
	default:
		return "❓"
	}
}

func (g *InventoryMarkdownGenerator) sortAnnotations(annotations []*analyze.AnnotationUsage, sortBy string) {
	switch sortBy {
	case "usage":
		sort.Slice(annotations, func(i, j int) bool {
			return annotations[i].UsageCount > annotations[j].UsageCount
		})
	case "risk":
		sort.Slice(annotations, func(i, j int) bool {
			// Risk priority: HIGH > MANUAL > AUTO
			riskOrder := map[models.RiskLevel]int{
				models.RiskHigh:   3,
				models.RiskManual: 2,
				models.RiskAuto:   1,
			}
			return riskOrder[annotations[i].Risk] > riskOrder[annotations[j].Risk]
		})
	case "name":
		sort.Slice(annotations, func(i, j int) bool {
			return annotations[i].Key < annotations[j].Key
		})
	default:
		// Default to usage
		sort.Slice(annotations, func(i, j int) bool {
			return annotations[i].UsageCount > annotations[j].UsageCount
		})
	}
}

func (g *InventoryMarkdownGenerator) formatValueExamples(valueExamples map[string]int, limit int) string {
	if len(valueExamples) == 0 {
		return ""
	}

	// Sort by frequency
	type valueFreq struct {
		value string
		count int
	}
	var freqs []valueFreq
	for value, count := range valueExamples {
		freqs = append(freqs, valueFreq{value, count})
	}
	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].count > freqs[j].count
	})

	var examples []string
	for i, vf := range freqs {
		if i >= limit {
			break
		}
		examples = append(examples, fmt.Sprintf("`%s`(%d)", vf.value, vf.count))
	}

	result := strings.Join(examples, ", ")
	if len(freqs) > limit {
		result += "..."
	}
	return result
}

func (g *InventoryMarkdownGenerator) getTopAnnotations(annotations []*analyze.AnnotationUsage, limit int) []*analyze.AnnotationUsage {
	if len(annotations) <= limit {
		return annotations
	}
	return annotations[:limit]
}

func (g *InventoryMarkdownGenerator) analyzeNamespaceComplexity(inventory *analyze.AnnotationInventory) map[string]string {
	namespaceComplexity := make(map[string]string)
	
	// Analyze each namespace by looking at annotation usage
	namespaceAnnotations := make(map[string]map[models.RiskLevel]int)
	
	for _, usage := range inventory.NginxAnnotations {
		for _, namespace := range usage.Namespaces {
			if namespaceAnnotations[namespace] == nil {
				namespaceAnnotations[namespace] = make(map[models.RiskLevel]int)
			}
			namespaceAnnotations[namespace][usage.Risk] += usage.UsageCount
		}
	}
	
	// Classify complexity
	for namespace, riskCounts := range namespaceAnnotations {
		total := riskCounts[models.RiskAuto] + riskCounts[models.RiskManual] + riskCounts[models.RiskHigh]
		if total == 0 {
			continue
		}
		
		highRiskPercent := float64(riskCounts[models.RiskHigh]) / float64(total) * 100
		
		if highRiskPercent > 50 {
			namespaceComplexity[namespace] = "HIGH complexity (many high-risk annotations)"
		} else if highRiskPercent > 20 {
			namespaceComplexity[namespace] = "MEDIUM complexity (some high-risk annotations)"
		} else {
			namespaceComplexity[namespace] = "LOW complexity (mostly auto-migratable)"
		}
	}
	
	return namespaceComplexity
}