#!/bin/bash

# Setup script for testing the Ingress-NGINX Migration Analyzer
# This script creates a kind cluster with ingress-nginx and sample ingresses

set -e

CLUSTER_NAME="ingress-analyzer-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Setting up test cluster for Ingress-NGINX Migration Analyzer"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "🔍 Checking prerequisites..."
if ! command_exists kind; then
    echo "❌ kind is not installed. Please install kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

if ! command_exists kubectl; then
    echo "❌ kubectl is not installed. Please install kubectl"
    exit 1
fi

# Check if cluster already exists
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "⚠️  Cluster '${CLUSTER_NAME}' already exists. Deleting it..."
    kind delete cluster --name "${CLUSTER_NAME}"
fi

# Create kind cluster
echo "🏗️  Creating kind cluster: ${CLUSTER_NAME}"
kind create cluster --name "${CLUSTER_NAME}" --config "${SCRIPT_DIR}/kind-cluster.yaml"

# Set kubectl context
echo "🎯 Setting kubectl context..."
kubectl cluster-info --context "kind-${CLUSTER_NAME}"

# Wait for cluster to be ready
echo "⏳ Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=60s

# Install ingress-nginx controller with snippets enabled
echo "📦 Installing ingress-nginx controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Enable snippet annotations for testing (wait for deployment first)
echo "⚙️  Enabling snippet annotations for testing..."
kubectl patch configmap ingress-nginx-controller -n ingress-nginx --patch='{"data":{"allow-snippet-annotations":"true"}}' || true

# Wait for ingress-nginx to be ready
echo "⏳ Waiting for ingress-nginx controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

# Install sample applications and ingresses
echo "🚀 Installing sample applications..."

# Create namespaces
kubectl create namespace production --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -

# Install sample apps and ingresses
kubectl apply -f "${REPO_ROOT}/testdata/sample-apps.yaml"
kubectl apply -f "${REPO_ROOT}/testdata/sample-ingresses-live.yaml"

# Wait for deployments to be ready
echo "⏳ Waiting for sample applications to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/echo-app -n default
kubectl wait --for=condition=available --timeout=60s deployment/api-app -n production
kubectl wait --for=condition=available --timeout=60s deployment/admin-app -n production
kubectl wait --for=condition=available --timeout=60s deployment/legacy-app -n staging

echo "✅ Test cluster setup complete!"
echo ""
echo "📊 Cluster Summary:"
kubectl get nodes
echo ""
kubectl get pods -A | grep -E "(ingress-nginx|echo|api|admin|legacy)"
echo ""
kubectl get ingress -A
echo ""
echo "🧪 Ready to test the analyzer:"
echo "   ${REPO_ROOT}/bin/analyzer scan"
echo ""
echo "🌐 Test endpoints (after ingress is ready):"
echo "   curl -H 'Host: echo.local' http://localhost:8080/"
echo "   curl -H 'Host: api.example.com' http://localhost:8080/"
echo ""
echo "🧹 To cleanup:"
echo "   kind delete cluster --name ${CLUSTER_NAME}"