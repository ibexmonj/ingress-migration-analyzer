package rules

import (
	"testing"

	"ingress-migration-analyzer/internal/models"
)

func TestGetAnnotationRules(t *testing.T) {
	rules := GetAnnotationRules()
	
	if len(rules) == 0 {
		t.Fatal("Expected annotation rules, got none")
	}

	// Check that we have rules for each risk level
	var autoCount, manualCount, highRiskCount int
	for _, rule := range rules {
		switch rule.RiskLevel {
		case models.RiskAuto:
			autoCount++
		case models.RiskManual:
			manualCount++
		case models.RiskHigh:
			highRiskCount++
		}
	}

	if autoCount == 0 {
		t.Error("Expected at least one AUTO risk rule")
	}
	if manualCount == 0 {
		t.Error("Expected at least one MANUAL risk rule")
	}
	if highRiskCount == 0 {
		t.Error("Expected at least one HIGH_RISK rule")
	}
}

func TestMatchAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		wantCount   int
		wantRisk    models.RiskLevel
	}{
		{
			name: "auto-migratable annotations",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/api/$1",
				"nginx.ingress.kubernetes.io/ssl-redirect":   "true",
			},
			wantCount: 2,
			wantRisk:  models.RiskAuto,
		},
		{
			name: "mixed risk annotations",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target":    "/api/$1",
				"nginx.ingress.kubernetes.io/server-snippet":   "custom config",
				"nginx.ingress.kubernetes.io/proxy-body-size": "50m",
			},
			wantCount: 3,
			wantRisk:  models.RiskHigh, // Highest risk wins
		},
		{
			name: "high-risk snippets",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/server-snippet":        "server config",
				"nginx.ingress.kubernetes.io/configuration-snippet": "location config",
			},
			wantCount: 2,
			wantRisk:  models.RiskHigh,
		},
		{
			name:        "no nginx annotations",
			annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/issuer":      "letsencrypt",
			},
			wantCount: 0,
			wantRisk:  models.RiskAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := MatchAnnotations(tt.annotations)
			
			if len(matches) != tt.wantCount {
				t.Errorf("MatchAnnotations() returned %d matches, want %d", len(matches), tt.wantCount)
			}

			if len(matches) > 0 {
				riskLevel := GetHighestRiskLevel(matches)
				if riskLevel != tt.wantRisk {
					t.Errorf("GetHighestRiskLevel() = %v, want %v", riskLevel, tt.wantRisk)
				}
			}
		})
	}
}

func TestGetUnknownNginxAnnotations(t *testing.T) {
	annotations := map[string]string{
		// Known annotations
		"nginx.ingress.kubernetes.io/rewrite-target": "/api/$1",
		"nginx.ingress.kubernetes.io/ssl-redirect":   "true",
		
		// Unknown nginx annotations
		"nginx.ingress.kubernetes.io/custom-unknown":  "value",
		"nginx.ingress.kubernetes.io/another-unknown": "value",
		
		// Non-nginx annotations (should be ignored)
		"kubernetes.io/ingress.class": "nginx",
		"cert-manager.io/issuer":      "letsencrypt",
	}

	unknown := GetUnknownNginxAnnotations(annotations)
	
	expectedUnknown := []string{
		"nginx.ingress.kubernetes.io/custom-unknown",
		"nginx.ingress.kubernetes.io/another-unknown",
	}

	if len(unknown) != len(expectedUnknown) {
		t.Errorf("Expected %d unknown annotations, got %d", len(expectedUnknown), len(unknown))
	}

	// Convert to map for easier checking
	unknownMap := make(map[string]bool)
	for _, u := range unknown {
		unknownMap[u] = true
	}

	for _, expected := range expectedUnknown {
		if !unknownMap[expected] {
			t.Errorf("Expected unknown annotation %s not found", expected)
		}
	}
}

func TestGetRuleByPattern(t *testing.T) {
	// Test known pattern
	rule := GetRuleByPattern("nginx.ingress.kubernetes.io/rewrite-target")
	if rule == nil {
		t.Fatal("Expected to find rule for rewrite-target")
	}
	if rule.RiskLevel != models.RiskAuto {
		t.Errorf("Expected rewrite-target to be AUTO risk, got %v", rule.RiskLevel)
	}

	// Test unknown pattern
	rule = GetRuleByPattern("nginx.ingress.kubernetes.io/unknown-annotation")
	if rule != nil {
		t.Error("Expected nil for unknown annotation, got rule")
	}
}