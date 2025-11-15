package analyze

import (
	"context"
	"fmt"
	"strings"

	"ingress-migration-analyzer/internal/models"
	"ingress-migration-analyzer/pkg/discovery"
	"ingress-migration-analyzer/pkg/rules"
)

// Analyzer performs the complete analysis of ingress-nginx resources
type Analyzer struct {
	scanner *discovery.Scanner
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(client *discovery.Client, namespace string) *Analyzer {
	scanner := discovery.NewScanner(client, namespace)
	return &Analyzer{
		scanner: scanner,
	}
}

// AnalyzeCluster performs complete cluster analysis
func (a *Analyzer) AnalyzeCluster(ctx context.Context) (*models.ClusterAnalysis, error) {
	fmt.Println("🔍 Starting cluster analysis...")

	// Scan cluster for ingress resources
	scanResult, err := a.scanner.ScanCluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("cluster scan failed: %w", err)
	}

	fmt.Printf("📊 Analyzing %d ingress-nginx resources...\n", len(scanResult.NginxIngresses))

	// Analyze each ingress resource
	var analyses []models.IngressAnalysis
	for _, resource := range scanResult.NginxIngresses {
		analysis := a.analyzeIngress(resource)
		analyses = append(analyses, analysis)
	}

	// Generate summary statistics
	summary := a.generateSummary(analyses)

	clusterAnalysis := &models.ClusterAnalysis{
		ScanResult: *scanResult,
		Analyses:   analyses,
		Summary:    summary,
	}

	a.printAnalysisSummary(summary)

	return clusterAnalysis, nil
}

// analyzeIngress analyzes a single Ingress resource
func (a *Analyzer) analyzeIngress(resource models.IngressResource) models.IngressAnalysis {
	// Match annotations against rules
	matchedRules := rules.MatchAnnotations(resource.Annotations)
	
	// Determine overall risk level
	riskLevel := rules.GetHighestRiskLevel(matchedRules)
	
	// Find unknown nginx annotations
	unknownAnnotations := rules.GetUnknownNginxAnnotations(resource.Annotations)
	
	// Generate warnings
	warnings := a.generateWarnings(resource, matchedRules)

	return models.IngressAnalysis{
		Resource:           resource,
		MatchedRules:       matchedRules,
		RiskLevel:          riskLevel,
		UnknownAnnotations: unknownAnnotations,
		Warnings:           warnings,
	}
}

// generateWarnings creates warnings for potential issues
func (a *Analyzer) generateWarnings(resource models.IngressResource, matchedRules []models.AnnotationRule) []string {
	var warnings []string

	// Warn about snippets
	for _, rule := range matchedRules {
		if strings.Contains(rule.Pattern, "snippet") {
			warnings = append(warnings, fmt.Sprintf("Contains %s: requires manual review and reimplementation", rule.Name))
		}
	}

	// Warn about unknown annotations
	unknown := rules.GetUnknownNginxAnnotations(resource.Annotations)
	if len(unknown) > 0 {
		warnings = append(warnings, fmt.Sprintf("Contains %d unknown nginx annotations", len(unknown)))
	}

	// Warn about deprecated class annotation
	if class, exists := resource.Annotations["kubernetes.io/ingress.class"]; exists && class == "nginx" {
		if resource.ClassName == "" {
			warnings = append(warnings, "Uses deprecated kubernetes.io/ingress.class annotation instead of spec.ingressClassName")
		}
	}

	return warnings
}

// generateSummary creates aggregate statistics
func (a *Analyzer) generateSummary(analyses []models.IngressAnalysis) models.AnalysisSummary {
	summary := models.AnalysisSummary{
		TotalIngresses: len(analyses),
		ByNamespace:    make(map[string]models.NamespaceSummary),
	}

	// Count by risk level and namespace
	for _, analysis := range analyses {
		// Global counts
		switch analysis.RiskLevel {
		case models.RiskAuto:
			summary.AutoCount++
		case models.RiskManual:
			summary.ManualCount++
		case models.RiskHigh:
			summary.HighRiskCount++
		}

		// Namespace counts
		ns := analysis.Resource.Namespace
		nsSummary := summary.ByNamespace[ns]
		switch analysis.RiskLevel {
		case models.RiskAuto:
			nsSummary.AutoCount++
		case models.RiskManual:
			nsSummary.ManualCount++
		case models.RiskHigh:
			nsSummary.HighRiskCount++
		}
		summary.ByNamespace[ns] = nsSummary
	}

	return summary
}

// printAnalysisSummary prints a summary of the analysis results
func (a *Analyzer) printAnalysisSummary(summary models.AnalysisSummary) {
	fmt.Println("\n📈 Analysis Summary:")
	fmt.Printf("   Total Resources: %d\n", summary.TotalIngresses)
	fmt.Printf("   ✅ AUTO-MIGRATABLE: %d (%.0f%%)\n", 
		summary.AutoCount, 
		float64(summary.AutoCount)/float64(summary.TotalIngresses)*100)
	fmt.Printf("   ⚠️  MANUAL REVIEW: %d (%.0f%%)\n", 
		summary.ManualCount,
		float64(summary.ManualCount)/float64(summary.TotalIngresses)*100)
	fmt.Printf("   ❌ HIGH RISK: %d (%.0f%%)\n", 
		summary.HighRiskCount,
		float64(summary.HighRiskCount)/float64(summary.TotalIngresses)*100)

	if len(summary.ByNamespace) > 1 {
		fmt.Println("\n📊 By Namespace:")
		for ns, nsSummary := range summary.ByNamespace {
			total := nsSummary.AutoCount + nsSummary.ManualCount + nsSummary.HighRiskCount
			fmt.Printf("   %s: AUTO=%d, MANUAL=%d, HIGH_RISK=%d (total=%d)\n",
				ns, nsSummary.AutoCount, nsSummary.ManualCount, nsSummary.HighRiskCount, total)
		}
	}
}

// GetRiskLevelIcon returns an icon for the risk level
func GetRiskLevelIcon(level models.RiskLevel) string {
	switch level {
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

// GetRiskLevelDescription returns a human-readable description
func GetRiskLevelDescription(level models.RiskLevel) string {
	switch level {
	case models.RiskAuto:
		return "Auto-migratable"
	case models.RiskManual:
		return "Manual review required"
	case models.RiskHigh:
		return "High risk - complex migration"
	default:
		return "Unknown risk level"
	}
}