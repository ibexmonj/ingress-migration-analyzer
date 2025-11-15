package common

// Report generation constants
const (
	// MaxAnnotationsPerRiskLevel limits the number of annotations displayed per risk level in markdown reports
	MaxAnnotationsPerRiskLevel = 10
	
	// MaxValueFrequencyExamples limits the number of value frequency examples shown in detailed reports
	MaxValueFrequencyExamples = 10
	
	// DefaultValueExampleLimit limits the number of value examples shown in summary tables
	DefaultValueExampleLimit = 3
	
	// MaxMigrationPhaseAnnotations limits the number of annotations shown per migration phase
	MaxMigrationPhaseAnnotations = 5
	
	// DefaultTopUsedAnnotations limits the number of top used annotations for analysis
	DefaultTopUsedAnnotations = 5
)

// CLI defaults
const (
	// DefaultReportsDir is the default directory for generated reports
	DefaultReportsDir = "./reports/"
	
	// DefaultKubeconfigPath is the standard kubeconfig location
	DefaultKubeconfigPath = ".kube/config"
)