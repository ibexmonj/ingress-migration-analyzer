package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ingress-migration-analyzer/internal/models"
)

// JSONGenerator generates JSON reports
type JSONGenerator struct{}

// NewJSONGenerator creates a new JSON generator
func NewJSONGenerator() *JSONGenerator {
	return &JSONGenerator{}
}

// GenerateReport creates a JSON report
func (j *JSONGenerator) GenerateReport(analysis *models.ClusterAnalysis, outputDir string) (string, error) {
	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	filename := fmt.Sprintf("migration-report-%s.json", timestamp)
	filepath := filepath.Join(outputDir, filename)

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write JSON report: %w", err)
	}

	return filepath, nil
}