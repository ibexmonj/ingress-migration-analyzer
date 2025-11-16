n the# Ingress-NGINX Migration Analyzer

> Analyze your ingress-nginx usage and plan your migration before the March 2026 EOL

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/badge/release-v0.1.0-green.svg)](https://github.com/user/ingress-migration-analyzer/releases)

## Overview

With the ingress-nginx project ending community support in **March 2026**, organizations need to assess their current usage and plan migration strategies. This tool scans Kubernetes clusters to identify ingress-nginx resources, classifies migration complexity, and generates actionable reports.

## Problem Statement

The ingress-nginx controller will reach end-of-life in March 2026, requiring all users to migrate to alternative solutions like Gateway API, other ingress controllers, or NGINX Inc's commercial offering. This migration's complexity depends heavily on which annotations and features you're currently using.

## Features

- 🔍 **Comprehensive Discovery**: Scan all or specific namespaces for ingress-nginx resources
- 📊 **Risk Classification**: Automatic complexity assessment with 3-tier system
- 📝 **Detailed Reports**: Generate markdown or JSON reports with migration guidance
- ⚡ **Fast Analysis**: Complete cluster scan in seconds
- 🎯 **Namespace Breakdown**: Per-namespace statistics and analysis
- 🔧 **Flexible Configuration**: Support for custom kubeconfig and contexts
- 📋 **Annotation Coverage**: 15+ known nginx annotations classified with source documentation
- 📦 **Comprehensive Inventory**: Detailed annotation usage analysis and migration planning

## Installation

### Binary Release
```bash
# Download latest release from GitHub
curl -L https://github.com/user/ingress-migration-analyzer/releases/latest/download/analyzer-linux-amd64 -o analyzer
chmod +x analyzer
sudo mv analyzer /usr/local/bin/
```

### Build from Source
```bash
git clone https://github.com/ibexmonj/ingress-migration-analyzer.git
cd ingress-migration-analyzer
make build
./bin/analyzer --version
```

### Go Install
```bash
go install github.com/ibexmonj/ingress-migration-analyzer/cmd/analyzer@latest
```

## Quick Start

```bash
# Basic scan of all namespaces
analyzer scan

# Scan specific namespace
analyzer scan --namespace production

# Generate comprehensive annotation inventory
analyzer inventory --format markdown --detailed

# Generate JSON report
analyzer scan --format json --output ./migration-reports/

# Use specific kubeconfig/context
analyzer scan --kubeconfig /path/to/kubeconfig --context production-cluster
```

## Migration Complexity Levels

The analyzer uses a **knowledge-based classification system** that maps each nginx annotation to Gateway API capabilities:

| Level | Icon | Description | Gateway API Mapping | Examples |
|-------|------|-------------|---------------------|----------|
| **AUTO** | ✅ | Direct Gateway API equivalents | Standard HTTPRoute filters | `rewrite-target`, `ssl-redirect`, `backend-protocol` |
| **MANUAL** | ⚠️ | No standard equivalent, but workarounds exist | Implementation-specific policies or service mesh | `proxy-body-size`, `auth-url`, timeouts |
| **HIGH_RISK** | ❌ | Custom NGINX configs with no Gateway API equivalent | Requires complete reimplementation | `server-snippet`, `configuration-snippet` |

### How Classification Works

The tool contains **expert-curated rules** based on:
1. **Gateway API Specification**: Standard HTTPRoute, Gateway, and policy features
2. **Implementation Analysis**: Support across popular Gateway implementations (Istio, Kong, Contour, etc.)
3. **Migration Experience**: Real-world migration patterns and common workarounds
4. **Community Input**: Feedback from the Kubernetes networking community

Each annotation includes:
- **Risk Level**: AUTO/MANUAL/HIGH_RISK classification
- **Migration Notes**: Specific guidance for that annotation
- **Source Documentation**: Links to official Gateway API and NGINX docs
- **Alternative Solutions**: Gateway API filters, service mesh options, or application-level changes

> **🔗 All migration recommendations are backed by source documentation** - Every annotation analysis includes links to official Gateway API specs and NGINX documentation to ensure credibility and provide engineers with authoritative references.

### Adding New Annotations

The classification rules are maintained in [`pkg/rules/annotations.go`](pkg/rules/annotations.go). To add support for new annotations:

```go
{
    Name:        "Custom Annotation",
    Pattern:     "nginx.ingress.kubernetes.io/custom-annotation",
    RiskLevel:   models.RiskManual,  // or RiskAuto/RiskHigh
    Description: "What this annotation does",
    MigrationNote: "How to migrate this to Gateway API or alternatives",
}
```

**Classification Guidelines:**
- **AUTO**: Direct 1:1 mapping to Gateway API standard features
- **MANUAL**: Requires Gateway implementation-specific policies or service mesh
- **HIGH_RISK**: Custom NGINX config with no Gateway API equivalent

## Sample Output

```
🔍 Starting ingress-nginx migration analysis...
📦 Scanning all namespaces

🔌 Testing Kubernetes connection...
✅ Connected to cluster (version: v1.28.2)

📊 Found 15 total Ingress resources
🎯 Found 8 ingress-nginx resources

📈 Analysis Summary:
   ✅ AUTO-MIGRATABLE: 3 (38%)
   ⚠️  MANUAL REVIEW: 3 (38%)
   ❌ HIGH RISK: 2 (25%)

✅ Report saved to: ./reports/migration-report-2025-11-15-143022.md
```

## Report Contents

Generated reports include:
- **Executive Summary** with migration complexity breakdown
- **High-Risk Resources** requiring immediate attention
- **Namespace Analysis** with per-namespace statistics  
- **Detailed Resource Analysis** with annotation-by-annotation guidance
- **Migration Recommendations** and next steps

## Supported Annotations

### Auto-Migratable (✅)
- `nginx.ingress.kubernetes.io/rewrite-target`
- `nginx.ingress.kubernetes.io/ssl-redirect`
- `nginx.ingress.kubernetes.io/force-ssl-redirect`
- `nginx.ingress.kubernetes.io/backend-protocol`
- `nginx.ingress.kubernetes.io/use-regex`

### Manual Review (⚠️)
- `nginx.ingress.kubernetes.io/proxy-body-size`
- `nginx.ingress.kubernetes.io/proxy-read-timeout`
- `nginx.ingress.kubernetes.io/proxy-send-timeout`
- `nginx.ingress.kubernetes.io/auth-url`
- `nginx.ingress.kubernetes.io/enable-cors`
- And more...

### High Risk (❌)
- `nginx.ingress.kubernetes.io/server-snippet`
- `nginx.ingress.kubernetes.io/configuration-snippet`
- `nginx.ingress.kubernetes.io/location-snippet`
- `nginx.ingress.kubernetes.io/stream-snippet`
- `nginx.ingress.kubernetes.io/http-snippet`

## Development

```bash
# Clone and setup
git clone https://github.com/ibexmonj/ingress-migration-analyzer.git
cd ingress-migration-analyzer
make dev-setup

# Run tests
make test

# Build
make build

# Lint and format
make lint
```

## Testing with Kind

For end-to-end testing, we provide a complete kind setup with ingress-nginx and sample ingresses:

```bash
# Setup test cluster with ingress-nginx and sample apps
./scripts/setup-test-cluster.sh

# Run analyzer against test cluster
./scripts/test-analyzer.sh

# Cleanup test cluster when done
kind delete cluster --name ingress-analyzer-test
```

The test setup includes:
- **Kubernetes 1.31** cluster with ingress-ready node
- **ingress-nginx controller** properly configured
- **5 sample ingresses** demonstrating different risk levels:
  - ✅ Simple rewrite rules (AUTO)
  - ⚠️ Auth and timeouts (MANUAL) 
  - ❌ Server snippets (HIGH_RISK)
  - 🔄 Mixed complexity scenarios
  - 📛 Deprecated annotation patterns

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] Core scanning and analysis engine
- [x] Markdown and JSON report generation
- [x] Risk-based annotation classification
- [ ] Gateway API specific migration suggestions
- [ ] Integration with CI/CD pipelines
- [ ] Migration cost estimation
- [ ] Support for additional ingress controllers
- [ ] Web dashboard interface

## Support

- 📋 [Issues](https://github.com/ibexmonj/ingress-migration-analyzer/issues) - Report bugs or request features
- 📖 [Documentation](./examples/) - Usage examples and guides
- 💬 [Discussions](https://github.com/ibexmonj/ingress-migration-analyzer/discussions) - Community support