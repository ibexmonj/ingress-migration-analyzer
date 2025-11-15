# Testing Scripts

This directory contains scripts for testing the Ingress-NGINX Migration Analyzer with a complete kind cluster setup.

## Quick Start

```bash
# Setup complete test environment
./scripts/setup-test-cluster.sh

# Run analysis
./scripts/test-analyzer.sh

# Cleanup when done
kind delete cluster --name ingress-analyzer-test
```

## Test Environment

### Cluster Configuration
- **Kubernetes**: v1.31.0 (via kind)
- **ingress-nginx**: Latest from official deployment
- **Namespaces**: `default`, `production`, `staging`
- **Port Mappings**: 8080 → 80, 8443 → 443

### Sample Applications
Four simple nginx-based web applications:
- **echo-app** (default): Basic hello world
- **api-app** (production): API service
- **admin-app** (production): Admin dashboard  
- **legacy-app** (staging): Legacy application

### Sample Ingresses
Demonstrates all risk levels and migration scenarios:

#### ✅ AUTO-MIGRATABLE (1 ingress)
- **simple-echo**: Basic rewrites and SSL redirects
  - `nginx.ingress.kubernetes.io/rewrite-target`
  - `nginx.ingress.kubernetes.io/ssl-redirect`
  - `nginx.ingress.kubernetes.io/force-ssl-redirect`

#### ⚠️ MANUAL REVIEW (3 ingresses)
- **api-with-auth**: Authentication and timeouts
  - `nginx.ingress.kubernetes.io/auth-url`
  - `nginx.ingress.kubernetes.io/proxy-body-size`
  - `nginx.ingress.kubernetes.io/proxy-read-timeout`
  
- **complex-routing**: Mixed annotations
  - `nginx.ingress.kubernetes.io/use-regex`
  - `nginx.ingress.kubernetes.io/rate-limit-rps`
  - `nginx.ingress.kubernetes.io/enable-cors`

- **legacy-with-deprecated-annotation**: Deprecated patterns
  - `kubernetes.io/ingress.class: nginx` (instead of ingressClassName)
  - Triggers deprecation warnings

#### ❌ HIGH RISK (optional)
- **high-risk-admin**: Server snippets (may be blocked by security policies)
  - `nginx.ingress.kubernetes.io/server-snippet`
  - `nginx.ingress.kubernetes.io/configuration-snippet`

#### Ignored Resources
- **other-ingress-controller**: Non-nginx ingress for comparison

## Expected Analysis Results

When you run the analyzer, you should see:
- **5 total** Ingress resources discovered
- **4 ingress-nginx** resources identified  
- **1 AUTO-MIGRATABLE** (25%)
- **3 MANUAL REVIEW** (75%)
- **0 HIGH RISK** (0% - snippets blocked by default)

## Testing Endpoints

```bash
# Test SSL redirect (should return 308)
curl -H 'Host: echo.local' http://localhost:8080/

# Test HTTPS endpoint (should return HTML)
curl -k -H 'Host: echo.local' https://localhost:8443/

# Test API endpoint
curl -k -H 'Host: api.example.com' https://localhost:8443/api
```

## Generated Reports

The analyzer generates comprehensive markdown reports with:
- Executive summary with percentages
- Namespace-level breakdown
- Detailed per-resource analysis
- Migration recommendations
- Risk-specific guidance

Example output location: `reports/migration-report-YYYY-MM-DD-HHMMSS.md`

## Development Workflow

```bash
# Make changes to analyzer code
vim pkg/analyze/analyzer.go

# Rebuild and test
go build -o bin/analyzer cmd/analyzer/main.go
./scripts/test-analyzer.sh

# View latest report
cat $(ls -t reports/migration-report-*.md | head -n1)
```