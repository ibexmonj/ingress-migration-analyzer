# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-15

### Added
- Initial release of Ingress-NGINX Migration Analyzer
- Complete cluster scanning for ingress-nginx resources
- Risk-based annotation classification system with 15+ known annotations
- Support for both `spec.ingressClassName` and legacy `kubernetes.io/ingress.class` detection
- Markdown report generation with detailed analysis
- JSON report generation for programmatic consumption
- CLI with cobra framework supporting:
  - Namespace-specific scanning
  - Custom kubeconfig and context selection
  - Configurable output directory
  - Format selection (markdown/json)
- Three-tier risk classification:
  - AUTO: Direct Gateway API equivalents (rewrite-target, ssl-redirect, etc.)
  - MANUAL: Requires review but migration path exists (timeouts, auth, etc.)  
  - HIGH_RISK: Complex custom configurations (server-snippet, etc.)
- Comprehensive migration recommendations and next steps
- Unknown annotation detection for nginx.ingress.kubernetes.io/* annotations
- Multi-namespace analysis with per-namespace breakdowns

### Features
- **Discovery**: Scan all or specific namespaces for ingress resources
- **Classification**: Automatic risk assessment based on annotation complexity
- **Reporting**: Human-readable markdown and machine-readable JSON reports
- **Analysis**: Detailed per-resource annotation analysis with migration notes
- **Warnings**: Detection of deprecated patterns and potential issues

### Documentation
- Complete README with project overview and quick start
- Usage examples and sample commands
- Test fixtures with various annotation patterns
- Makefile with common development tasks