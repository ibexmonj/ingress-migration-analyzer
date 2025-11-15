#!/bin/bash

# Test script for running the analyzer against the test cluster

set -e

CLUSTER_NAME="ingress-analyzer-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🧪 Testing Ingress-NGINX Migration Analyzer"

# Check if test cluster exists
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "❌ Test cluster '${CLUSTER_NAME}' not found. Please run:"
    echo "   ${SCRIPT_DIR}/setup-test-cluster.sh"
    exit 1
fi

# Set kubectl context
echo "🎯 Setting kubectl context to test cluster..."
kubectl config use-context "kind-${CLUSTER_NAME}"

# Build the analyzer if not exists
if [ ! -f "${REPO_ROOT}/bin/analyzer" ]; then
    echo "🔨 Building analyzer..."
    cd "${REPO_ROOT}"
    go build -o bin/analyzer cmd/analyzer/main.go
fi

# Show current cluster state
echo "📊 Current cluster state:"
echo "Cluster version: $(kubectl version --short --client=false | grep Server)"
echo "Ingress controller pods:"
kubectl get pods -n ingress-nginx | grep controller || echo "  No ingress-nginx pods found"
echo "Ingress resources:"
kubectl get ingress -A

echo ""
echo "🔍 Running migration analysis..."
echo "=========================================="

# Run the analyzer
"${REPO_ROOT}/bin/analyzer" scan

echo ""
echo "=========================================="
echo "✅ Analysis complete!"
echo ""

# Show the generated report
LATEST_REPORT=$(ls -t "${REPO_ROOT}/reports/"migration-report-*.md 2>/dev/null | head -n1)
if [ -n "$LATEST_REPORT" ]; then
    echo "📝 Latest report: $LATEST_REPORT"
    echo ""
    echo "📖 Report preview (first 50 lines):"
    echo "----------------------------------------"
    head -n 50 "$LATEST_REPORT"
    echo "----------------------------------------"
    echo ""
    echo "💡 View the full report with:"
    echo "   cat '$LATEST_REPORT'"
    echo "   code '$LATEST_REPORT'"
else
    echo "❌ No report found in reports/ directory"
fi