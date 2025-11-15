package models

import (
	"time"
)

// RiskLevel defines the migration complexity level
type RiskLevel string

const (
	RiskAuto   RiskLevel = "AUTO"      // Auto-migratable
	RiskManual RiskLevel = "MANUAL"    // Requires manual review
	RiskHigh   RiskLevel = "HIGH_RISK" // Complex/dangerous
)

// IngressResource represents a discovered Ingress resource
type IngressResource struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	ClassName   string            `json:"className"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Hosts       []string          `json:"hosts"`
	Paths       []string          `json:"paths"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// ScanResult represents the results of cluster scanning
type ScanResult struct {
	ClusterVersion string            `json:"clusterVersion"`
	TotalIngresses int               `json:"totalIngresses"`
	NginxIngresses []IngressResource `json:"nginxIngresses"`
	ScanTime       time.Time         `json:"scanTime"`
}

// AnnotationRule defines how to classify a specific annotation
type AnnotationRule struct {
	Name          string    `json:"name"`
	Pattern       string    `json:"pattern"` // annotation key pattern
	RiskLevel     RiskLevel `json:"riskLevel"`
	Description   string    `json:"description"`
	MigrationNote string    `json:"migrationNote"` // What to do about it
	SourceURL     string    `json:"sourceUrl"`     // Documentation source
}

// IngressAnalysis represents the analysis result for a single Ingress
type IngressAnalysis struct {
	Resource           IngressResource  `json:"resource"`
	MatchedRules       []AnnotationRule `json:"matchedRules"`
	RiskLevel          RiskLevel        `json:"riskLevel"`
	UnknownAnnotations []string         `json:"unknownAnnotations"`
	Warnings           []string         `json:"warnings"`
}

// NamespaceSummary provides aggregated stats for a namespace
type NamespaceSummary struct {
	AutoCount     int `json:"autoCount"`
	ManualCount   int `json:"manualCount"`
	HighRiskCount int `json:"highRiskCount"`
}

// AnalysisSummary provides high-level analysis statistics
type AnalysisSummary struct {
	TotalIngresses int                         `json:"totalIngresses"`
	AutoCount      int                         `json:"autoCount"`
	ManualCount    int                         `json:"manualCount"`
	HighRiskCount  int                         `json:"highRiskCount"`
	ByNamespace    map[string]NamespaceSummary `json:"byNamespace"`
}

// ClusterAnalysis represents the complete analysis result
type ClusterAnalysis struct {
	ScanResult ScanResult        `json:"scanResult"`
	Analyses   []IngressAnalysis `json:"analyses"`
	Summary    AnalysisSummary   `json:"summary"`
	Inventory  interface{}       `json:"inventory,omitempty"`
}
