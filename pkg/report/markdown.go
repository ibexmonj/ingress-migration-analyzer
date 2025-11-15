package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ingress-migration-analyzer/internal/models"
	"ingress-migration-analyzer/pkg/analyze"
)

// MarkdownGenerator generates markdown reports
type MarkdownGenerator struct{
	ContextName string
}

// NewMarkdownGenerator creates a new markdown generator
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{}
}

// GenerateReport creates a comprehensive markdown report
func (m *MarkdownGenerator) GenerateReport(analysis *models.ClusterAnalysis, outputDir string) (string, error) {
	// Generate the report content
	content := m.generateReportContent(analysis)

	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	filename := fmt.Sprintf("migration-report-%s.md", timestamp)
	filepath := filepath.Join(outputDir, filename)

	// Write to file
	err := os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	return filepath, nil
}

// generateReportContent creates the full markdown content
func (m *MarkdownGenerator) generateReportContent(analysis *models.ClusterAnalysis) string {
	var content strings.Builder

	// Header
	m.writeHeader(&content, analysis)
	
	// Executive Summary
	m.writeExecutiveSummary(&content, analysis)
	
	// High-Risk Resources (if any)
	if analysis.Summary.HighRiskCount > 0 {
		m.writeHighRiskResources(&content, analysis)
	}

	// Namespace Analysis
	m.writeNamespaceAnalysis(&content, analysis)

	// Detailed Resource Analysis
	m.writeDetailedAnalysis(&content, analysis)

	// Migration Recommendations
	m.writeMigrationRecommendations(&content, analysis)

	return content.String()
}

// writeHeader writes the report header
func (m *MarkdownGenerator) writeHeader(content *strings.Builder, analysis *models.ClusterAnalysis) {
	content.WriteString("# Ingress-NGINX Migration Report\n\n")
	content.WriteString(fmt.Sprintf("**Generated**: %s\n", analysis.ScanResult.ScanTime.Format("2006-01-02 15:04:05")))
	if m.ContextName != "" {
		content.WriteString(fmt.Sprintf("**Cluster Context**: %s\n", m.ContextName))
	}
	content.WriteString(fmt.Sprintf("**Cluster Version**: %s\n", analysis.ScanResult.ClusterVersion))
	content.WriteString(fmt.Sprintf("**Total Ingress Resources**: %d\n", analysis.ScanResult.TotalIngresses))
	content.WriteString(fmt.Sprintf("**Ingress-NGINX Resources**: %d\n", len(analysis.ScanResult.NginxIngresses)))
	content.WriteString("\n---\n\n")
}

// writeExecutiveSummary writes the executive summary
func (m *MarkdownGenerator) writeExecutiveSummary(content *strings.Builder, analysis *models.ClusterAnalysis) {
	summary := analysis.Summary
	total := summary.TotalIngresses
	
	content.WriteString("## Executive Summary\n\n")
	
	if total == 0 {
		content.WriteString("🎉 **No ingress-nginx resources found!** Your cluster is already using other ingress solutions.\n\n")
		return
	}

	content.WriteString(fmt.Sprintf("- ✅ **AUTO-MIGRATABLE**: %d (%.0f%%)\n", 
		summary.AutoCount, float64(summary.AutoCount)/float64(total)*100))
	content.WriteString(fmt.Sprintf("- ⚠️  **MANUAL REVIEW**: %d (%.0f%%)\n", 
		summary.ManualCount, float64(summary.ManualCount)/float64(total)*100))
	content.WriteString(fmt.Sprintf("- ❌ **HIGH RISK**: %d (%.0f%%)\n", 
		summary.HighRiskCount, float64(summary.HighRiskCount)/float64(total)*100))

	content.WriteString("\n")
	m.writeMigrationComplexityExplanation(content)
	content.WriteString("\n---\n\n")
}

// writeMigrationComplexityExplanation explains the risk levels
func (m *MarkdownGenerator) writeMigrationComplexityExplanation(content *strings.Builder) {
	content.WriteString("### Migration Complexity Levels\n\n")
	content.WriteString("- **✅ AUTO-MIGRATABLE**: Simple annotations with direct Gateway API equivalents\n")
	content.WriteString("- **⚠️ MANUAL REVIEW**: Requires review but migration path exists\n")
	content.WriteString("- **❌ HIGH RISK**: Complex configurations requiring careful planning\n\n")
}

// writeHighRiskResources highlights high-risk resources
func (m *MarkdownGenerator) writeHighRiskResources(content *strings.Builder, analysis *models.ClusterAnalysis) {
	content.WriteString("## High-Risk Resources (Immediate Attention Required)\n\n")
	content.WriteString("These resources use complex annotations that require careful migration planning.\n\n")

	// Group by namespace
	byNamespace := make(map[string][]models.IngressAnalysis)
	for _, a := range analysis.Analyses {
		if a.RiskLevel == models.RiskHigh {
			ns := a.Resource.Namespace
			byNamespace[ns] = append(byNamespace[ns], a)
		}
	}

	// Sort namespaces
	var namespaces []string
	for ns := range byNamespace {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	for _, ns := range namespaces {
		content.WriteString(fmt.Sprintf("### Namespace: %s\n\n", ns))
		
		for _, a := range byNamespace[ns] {
			highRiskRules := m.getHighRiskRules(a.MatchedRules)
			ruleNames := make([]string, len(highRiskRules))
			for i, rule := range highRiskRules {
				ruleNames[i] = rule.Name
			}
			
			content.WriteString(fmt.Sprintf("- **%s** - Uses: %s\n", 
				a.Resource.Name, strings.Join(ruleNames, ", ")))
		}
		content.WriteString("\n")
	}

	content.WriteString("---\n\n")
}

// writeNamespaceAnalysis creates the namespace breakdown table
func (m *MarkdownGenerator) writeNamespaceAnalysis(content *strings.Builder, analysis *models.ClusterAnalysis) {
	if len(analysis.Summary.ByNamespace) <= 1 {
		return // Skip if only one namespace
	}

	content.WriteString("## Analysis by Namespace\n\n")
	content.WriteString("| Namespace | AUTO | MANUAL | HIGH RISK | Total |\n")
	content.WriteString("|-----------|------|--------|-----------|-------|\n")

	// Sort namespaces for consistent output
	var namespaces []string
	for ns := range analysis.Summary.ByNamespace {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	for _, ns := range namespaces {
		nsSummary := analysis.Summary.ByNamespace[ns]
		total := nsSummary.AutoCount + nsSummary.ManualCount + nsSummary.HighRiskCount
		content.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d |\n",
			ns, nsSummary.AutoCount, nsSummary.ManualCount, nsSummary.HighRiskCount, total))
	}

	content.WriteString("\n---\n\n")
}

// writeDetailedAnalysis writes detailed analysis for each resource
func (m *MarkdownGenerator) writeDetailedAnalysis(content *strings.Builder, analysis *models.ClusterAnalysis) {
	content.WriteString("## Detailed Resource Analysis\n\n")

	// Sort by namespace and name for consistent output
	analyses := make([]models.IngressAnalysis, len(analysis.Analyses))
	copy(analyses, analysis.Analyses)
	sort.Slice(analyses, func(i, j int) bool {
		if analyses[i].Resource.Namespace != analyses[j].Resource.Namespace {
			return analyses[i].Resource.Namespace < analyses[j].Resource.Namespace
		}
		return analyses[i].Resource.Name < analyses[j].Resource.Name
	})

	for _, a := range analyses {
		m.writeResourceDetails(content, a)
	}
}

// writeResourceDetails writes detailed analysis for a single resource
func (m *MarkdownGenerator) writeResourceDetails(content *strings.Builder, analysis models.IngressAnalysis) {
	resource := analysis.Resource
	icon := analyze.GetRiskLevelIcon(analysis.RiskLevel)
	
	content.WriteString(fmt.Sprintf("### %s %s/%s\n\n", icon, resource.Namespace, resource.Name))
	content.WriteString(fmt.Sprintf("- **Risk Level**: %s\n", analysis.RiskLevel))
	content.WriteString(fmt.Sprintf("- **Ingress Class**: %s\n", resource.ClassName))
	
	if len(resource.Hosts) > 0 {
		content.WriteString(fmt.Sprintf("- **Hosts**: %s\n", strings.Join(resource.Hosts, ", ")))
	}

	// Annotations analysis
	if len(analysis.MatchedRules) > 0 {
		content.WriteString("- **Annotations**:\n")
		
		// Group by risk level for better presentation
		autoRules := m.getRulesByRisk(analysis.MatchedRules, models.RiskAuto)
		manualRules := m.getRulesByRisk(analysis.MatchedRules, models.RiskManual)
		highRiskRules := m.getRulesByRisk(analysis.MatchedRules, models.RiskHigh)

		for _, rule := range autoRules {
			annotationValue := resource.Annotations[rule.Pattern]
			content.WriteString(fmt.Sprintf("  - ✅ %s: `%s` → %s", 
				rule.Name, annotationValue, rule.MigrationNote))
			if rule.SourceURL != "" {
				content.WriteString(fmt.Sprintf(" ([docs](%s))", rule.SourceURL))
			}
			content.WriteString("\n")
		}
		
		for _, rule := range manualRules {
			annotationValue := resource.Annotations[rule.Pattern]
			content.WriteString(fmt.Sprintf("  - ⚠️  %s: `%s` → %s", 
				rule.Name, annotationValue, rule.MigrationNote))
			if rule.SourceURL != "" {
				content.WriteString(fmt.Sprintf(" ([docs](%s))", rule.SourceURL))
			}
			content.WriteString("\n")
		}
		
		for _, rule := range highRiskRules {
			annotationValue := resource.Annotations[rule.Pattern]
			content.WriteString(fmt.Sprintf("  - ❌ %s: `%s` → %s", 
				rule.Name, annotationValue, rule.MigrationNote))
			if rule.SourceURL != "" {
				content.WriteString(fmt.Sprintf(" ([docs](%s))", rule.SourceURL))
			}
			content.WriteString("\n")
		}
	}

	// Unknown annotations
	if len(analysis.UnknownAnnotations) > 0 {
		content.WriteString("- **Unknown NGINX Annotations**:\n")
		for _, annotation := range analysis.UnknownAnnotations {
			value := resource.Annotations[annotation]
			content.WriteString(fmt.Sprintf("  - ❓ %s: `%s`\n", annotation, value))
		}
	}

	// Warnings
	if len(analysis.Warnings) > 0 {
		content.WriteString("- **Warnings**:\n")
		for _, warning := range analysis.Warnings {
			content.WriteString(fmt.Sprintf("  - ⚠️  %s\n", warning))
		}
	}

	// Migration notes for high-risk items
	if analysis.RiskLevel == models.RiskHigh {
		content.WriteString("\n**Migration Notes**:\n")
		highRiskRules := m.getHighRiskRules(analysis.MatchedRules)
		for _, rule := range highRiskRules {
			content.WriteString(fmt.Sprintf("- %s\n", rule.MigrationNote))
		}
	}

	content.WriteString("\n")
}

// writeMigrationRecommendations writes general migration recommendations
func (m *MarkdownGenerator) writeMigrationRecommendations(content *strings.Builder, analysis *models.ClusterAnalysis) {
	content.WriteString("## Migration Recommendations\n\n")

	if analysis.Summary.TotalIngresses == 0 {
		content.WriteString("No ingress-nginx resources found - no migration needed!\n")
		return
	}

	content.WriteString("### General Guidance\n\n")
	
	if analysis.Summary.AutoCount > 0 {
		content.WriteString(fmt.Sprintf("1. **Start with AUTO-MIGRATABLE resources** (%d resources)\n", analysis.Summary.AutoCount))
		content.WriteString("   - These have direct Gateway API equivalents\n")
		content.WriteString("   - Low risk for initial migration testing\n\n")
	}

	if analysis.Summary.ManualCount > 0 {
		content.WriteString(fmt.Sprintf("2. **Plan MANUAL REVIEW resources** (%d resources)\n", analysis.Summary.ManualCount))
		content.WriteString("   - Check your Gateway implementation's policy support\n")
		content.WriteString("   - Consider service mesh alternatives\n")
		content.WriteString("   - May require application-level changes\n\n")
	}

	if analysis.Summary.HighRiskCount > 0 {
		content.WriteString(fmt.Sprintf("3. **HIGH RISK resources require careful planning** (%d resources)\n", analysis.Summary.HighRiskCount))
		content.WriteString("   - Custom NGINX configurations have no direct equivalent\n")
		content.WriteString("   - Consider staying with NGINX Inc commercial controller\n")
		content.WriteString("   - Evaluate service mesh for complex routing needs\n\n")
	}

	content.WriteString("### Next Steps\n\n")
	content.WriteString("1. Review this report with your platform team\n")
	content.WriteString("2. Choose a Gateway API implementation (Istio, Kong, Contour, etc.)\n")
	content.WriteString("3. Set up a test cluster for migration validation\n")
	content.WriteString("4. Start with AUTO-MIGRATABLE resources for proof of concept\n")
	content.WriteString("5. Develop migration runbooks for MANUAL and HIGH RISK resources\n\n")

	content.WriteString("---\n\n")
	content.WriteString("*Generated by [Ingress-NGINX Migration Analyzer](https://github.com/user/ingress-migration-analyzer)*\n")
}

// Helper functions

func (m *MarkdownGenerator) getRulesByRisk(rules []models.AnnotationRule, riskLevel models.RiskLevel) []models.AnnotationRule {
	var filtered []models.AnnotationRule
	for _, rule := range rules {
		if rule.RiskLevel == riskLevel {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

func (m *MarkdownGenerator) getHighRiskRules(rules []models.AnnotationRule) []models.AnnotationRule {
	return m.getRulesByRisk(rules, models.RiskHigh)
}