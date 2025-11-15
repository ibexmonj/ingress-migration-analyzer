# Usage Examples

## Basic Usage

```bash
# Analyze all namespaces
analyzer scan

# Analyze specific namespace
analyzer scan --namespace production

# Generate JSON report
analyzer scan --format json

# Specify output directory
analyzer scan --output ./migration-reports/
```

## Using Different Kubeconfig

```bash
# Use different kubeconfig file
analyzer scan --kubeconfig /path/to/kubeconfig

# Use specific context
analyzer scan --context production-cluster
```

## Sample Output

After running the analyzer, you'll get output like:

```
🔍 Starting ingress-nginx migration analysis...
📁 Output directory: ./reports/
📄 Format: markdown
📦 Scanning all namespaces

🔌 Testing Kubernetes connection...
✅ Connected to cluster (version: v1.28.2)

🔍 Scanning cluster for Ingress resources...
📊 Found 15 total Ingress resources
🎯 Found 8 ingress-nginx resources

📊 Analyzing 8 ingress-nginx resources...

📈 Analysis Summary:
   Total Resources: 8
   ✅ AUTO-MIGRATABLE: 3 (38%)
   ⚠️  MANUAL REVIEW: 3 (38%)
   ❌ HIGH RISK: 2 (25%)

📊 By Namespace:
   default: AUTO=2, MANUAL=1, HIGH_RISK=0 (total=3)
   production: AUTO=1, MANUAL=2, HIGH_RISK=2 (total=5)

📝 Generating report...
✅ Analysis complete! Report saved to: ./reports/migration-report-2025-11-15-143022.md

⚠️  Warning: Found 2 high-risk resources requiring careful migration planning
```

## Report Contents

The generated report includes:

- **Executive Summary**: High-level migration complexity breakdown
- **High-Risk Resources**: Detailed breakdown of complex configurations
- **Namespace Analysis**: Per-namespace statistics
- **Detailed Resource Analysis**: Annotation-by-annotation analysis
- **Migration Recommendations**: Next steps and guidance

## Risk Levels Explained

- **✅ AUTO-MIGRATABLE**: Simple annotations with direct Gateway API equivalents
  - `rewrite-target`, `ssl-redirect`, `backend-protocol`
  - Low migration effort, can be automated

- **⚠️ MANUAL REVIEW**: Requires review but migration path exists
  - `proxy-body-size`, `auth-url`, `proxy-timeouts`
  - May need Gateway implementation policies or service mesh

- **❌ HIGH RISK**: Complex configurations requiring careful planning
  - `server-snippet`, `configuration-snippet`, `location-snippet`
  - Custom NGINX config with no direct Gateway API equivalent