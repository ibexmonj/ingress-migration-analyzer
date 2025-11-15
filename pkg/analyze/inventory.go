package analyze

import (
	"sort"
	"strings"

	"ingress-migration-analyzer/internal/models"
	"ingress-migration-analyzer/pkg/rules"
)

// AnnotationUsage tracks how an annotation is used across the cluster
type AnnotationUsage struct {
	Key           string            `json:"key"`
	UniqueValues  []string          `json:"uniqueValues"`
	UsageCount    int               `json:"usageCount"`
	Namespaces    []string          `json:"namespaces"`
	ValueExamples map[string]int    `json:"valueExamples"` // value -> count
	Risk          models.RiskLevel  `json:"risk"`
	Description   string            `json:"description"`
	MigrationNote string            `json:"migrationNote"`
	SourceURL     string            `json:"sourceUrl"`
}

// AnnotationInventory provides comprehensive annotation analysis
type AnnotationInventory struct {
	AllAnnotations     map[string]*AnnotationUsage `json:"allAnnotations"`
	NginxAnnotations   map[string]*AnnotationUsage `json:"nginxAnnotations"`
	UnknownAnnotations map[string]*AnnotationUsage `json:"unknownAnnotations"`
	Summary           InventorySummary            `json:"summary"`
}

// InventorySummary provides high-level inventory statistics
type InventorySummary struct {
	TotalUniqueAnnotations    int `json:"totalUniqueAnnotations"`
	NginxAnnotationsCount     int `json:"nginxAnnotationsCount"`
	UnknownAnnotationsCount   int `json:"unknownAnnotationsCount"`
	MostUsedAnnotation       string `json:"mostUsedAnnotation"`
	MostComplexNamespace     string `json:"mostComplexNamespace"`
}

// BuildAnnotationInventory creates comprehensive annotation usage analysis
// by processing all ingress analyses and categorizing annotations by usage
// patterns, risk levels, and migration complexity. System annotations are
// automatically filtered out from the analysis.
func BuildAnnotationInventory(analyses []models.IngressAnalysis) *AnnotationInventory {
	inventory := &AnnotationInventory{
		AllAnnotations:     make(map[string]*AnnotationUsage),
		NginxAnnotations:   make(map[string]*AnnotationUsage),
		UnknownAnnotations: make(map[string]*AnnotationUsage),
	}

	// Process each ingress analysis
	for _, analysis := range analyses {
		for key, value := range analysis.Resource.Annotations {
			// Skip system annotations that are not relevant for migration
			if isSystemAnnotation(key) {
				continue
			}

			usage := getOrCreateUsage(inventory.AllAnnotations, key)
			updateUsage(usage, value, analysis.Resource.Namespace)

			// Categorize nginx annotations
			if strings.HasPrefix(key, "nginx.ingress.kubernetes.io/") {
				nginxUsage := getOrCreateUsage(inventory.NginxAnnotations, key)
				updateUsage(nginxUsage, value, analysis.Resource.Namespace)
				
				// Add risk and migration info
				if rule := rules.GetRuleByPattern(key); rule != nil {
					nginxUsage.Risk = rule.RiskLevel
					nginxUsage.Description = rule.Description
					nginxUsage.MigrationNote = rule.MigrationNote
					nginxUsage.SourceURL = rule.SourceURL
				} else {
					nginxUsage.Risk = models.RiskLevel("UNKNOWN")
					nginxUsage.Description = "Unknown nginx annotation - not in current knowledge base"
					nginxUsage.MigrationNote = "This annotation is not documented in our migration rules. Please research Gateway API equivalent or file an issue."
					nginxUsage.SourceURL = ""
				}
			}
		}

		// Track unknown nginx annotations specifically
		for _, unknown := range analysis.UnknownAnnotations {
			usage := getOrCreateUsage(inventory.UnknownAnnotations, unknown)
			value := analysis.Resource.Annotations[unknown]
			updateUsage(usage, value, analysis.Resource.Namespace)
		}
	}

	// Generate summary
	inventory.Summary = generateInventorySummary(inventory)

	return inventory
}

// getOrCreateUsage gets existing usage or creates new one
func getOrCreateUsage(usageMap map[string]*AnnotationUsage, key string) *AnnotationUsage {
	if usage, exists := usageMap[key]; exists {
		return usage
	}

	usage := &AnnotationUsage{
		Key:           key,
		UniqueValues:  []string{},
		UsageCount:    0,
		Namespaces:    []string{},
		ValueExamples: make(map[string]int),
	}
	usageMap[key] = usage
	return usage
}

// updateUsage updates usage statistics
func updateUsage(usage *AnnotationUsage, value, namespace string) {
	usage.UsageCount++

	// Track unique values
	found := false
	for _, v := range usage.UniqueValues {
		if v == value {
			found = true
			break
		}
	}
	if !found {
		usage.UniqueValues = append(usage.UniqueValues, value)
	}

	// Track unique namespaces
	found = false
	for _, ns := range usage.Namespaces {
		if ns == namespace {
			found = true
			break
		}
	}
	if !found {
		usage.Namespaces = append(usage.Namespaces, namespace)
	}

	// Track value frequency (limit to avoid bloat)
	usage.ValueExamples[value]++
}

// generateInventorySummary creates summary statistics
func generateInventorySummary(inventory *AnnotationInventory) InventorySummary {
	summary := InventorySummary{
		TotalUniqueAnnotations:  len(inventory.AllAnnotations),
		NginxAnnotationsCount:   len(inventory.NginxAnnotations),
		UnknownAnnotationsCount: len(inventory.UnknownAnnotations),
	}

	// Find most used annotation
	maxUsage := 0
	for key, usage := range inventory.AllAnnotations {
		if usage.UsageCount > maxUsage {
			maxUsage = usage.UsageCount
			summary.MostUsedAnnotation = key
		}
	}

	return summary
}

// GetAnnotationsByRisk returns annotations grouped by risk level,
// sorted by usage count within each risk level for prioritization.
func (inv *AnnotationInventory) GetAnnotationsByRisk() map[models.RiskLevel][]*AnnotationUsage {
	byRisk := make(map[models.RiskLevel][]*AnnotationUsage)
	
	for _, usage := range inv.NginxAnnotations {
		byRisk[usage.Risk] = append(byRisk[usage.Risk], usage)
	}

	// Sort each risk level by usage count
	for riskLevel := range byRisk {
		sort.Slice(byRisk[riskLevel], func(i, j int) bool {
			return byRisk[riskLevel][i].UsageCount > byRisk[riskLevel][j].UsageCount
		})
	}

	return byRisk
}

// GetMostCriticalAnnotations returns the most problematic annotations for migration,
// including high-risk and unknown annotations, sorted by usage frequency to prioritize
// the most impactful migration decisions.
func (inv *AnnotationInventory) GetMostCriticalAnnotations(limit int) []*AnnotationUsage {
	var critical []*AnnotationUsage

	// Collect high-risk annotations
	for _, usage := range inv.NginxAnnotations {
		if usage.Risk == models.RiskHigh {
			critical = append(critical, usage)
		}
	}

	// Add unknown annotations (also high risk)
	for _, usage := range inv.UnknownAnnotations {
		usage.Risk = models.RiskLevel("UNKNOWN")
		critical = append(critical, usage)
	}

	// Sort by usage count (most used = most critical)
	sort.Slice(critical, func(i, j int) bool {
		return critical[i].UsageCount > critical[j].UsageCount
	})

	if len(critical) > limit {
		return critical[:limit]
	}
	return critical
}

// isSystemAnnotation checks if an annotation is a system/management annotation
// that should be filtered out from migration analysis
func isSystemAnnotation(key string) bool {
	systemPrefixes := []string{
		"kubectl.kubernetes.io/",
		"deployment.kubernetes.io/",
		"control-plane.alpha.kubernetes.io/",
		"pv.kubernetes.io/",
		"volume.beta.kubernetes.io/",
		"alpha.kubernetes.io/",
		"beta.kubernetes.io/",
		"node.alpha.kubernetes.io/",
		"scheduler.alpha.kubernetes.io/",
		"autoscaling.alpha.kubernetes.io/",
		"controller.kubernetes.io/",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Also filter out common system annotations
	systemAnnotations := []string{
		"kubernetes.io/ingress.class",  // This is actually important for migration but handled separately
		"field.cattle.io/publicEndpoints",
		"meta.helm.sh/release-name",
		"meta.helm.sh/release-namespace",
	}

	for _, sysAnnotation := range systemAnnotations {
		if key == sysAnnotation {
			return true
		}
	}

	return false
}