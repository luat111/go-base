#!/bin/bash
# Service Mesh Verification Script
# This script helps verify the Linkerd service mesh setup

set -e

echo "========================================="
echo "Service Mesh Verification Script"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Linkerd CLI is installed
echo "1. Checking Linkerd CLI..."
if command -v linkerd &> /dev/null; then
    echo -e "${GREEN}✓ Linkerd CLI is installed${NC}"
    linkerd version --client
else
    echo -e "${RED}✗ Linkerd CLI is not installed${NC}"
    echo "Install with: curl -sL https://run.linkerd.io/install | sh"
    exit 1
fi
echo ""

# Check Linkerd control plane
echo "2. Checking Linkerd control plane..."
if linkerd check &> /dev/null; then
    echo -e "${GREEN}✓ Linkerd control plane is running${NC}"
else
    echo -e "${YELLOW}⚠ Linkerd control plane not found${NC}"
    echo "Install with:"
    echo "  linkerd install --crds | kubectl apply -f -"
    echo "  linkerd install | kubectl apply -f -"
fi
echo ""

# Check if Helm is installed
echo "3. Checking Helm..."
if command -v helm &> /dev/null; then
    echo -e "${GREEN}✓ Helm is installed${NC}"
    helm version --short
else
    echo -e "${RED}✗ Helm is not installed${NC}"
    echo "Install from: https://helm.sh/docs/intro/install/"
    exit 1
fi
echo ""

# Validate Helm chart
echo "4. Validating Helm chart..."
cd "$(dirname "$0")/../helm/go-base"
if helm lint . ; then
    echo -e "${GREEN}✓ Helm chart validation passed${NC}"
else
    echo -e "${RED}✗ Helm chart validation failed${NC}"
    exit 1
fi
echo ""

# Template the chart with service mesh enabled
echo "5. Templating chart with service mesh enabled..."
TEMP_DIR=$(mktemp -d)
if helm template go-base-test . \
    --namespace test-mesh \
    --set serviceMesh.enabled=true \
    --output-dir "$TEMP_DIR" > /dev/null; then
    echo -e "${GREEN}✓ Chart templating successful${NC}"
    echo "Templates generated in: $TEMP_DIR"
    
    # Check for Linkerd annotations
    echo ""
    echo "6. Verifying Linkerd annotations..."
    if grep -r "linkerd.io/inject" "$TEMP_DIR" > /dev/null; then
        echo -e "${GREEN}✓ Linkerd injection annotations found${NC}"
    else
        echo -e "${RED}✗ Linkerd injection annotations not found${NC}"
    fi
    
    # Check for HTTP/2 configuration
    if grep -r "appProtocol: HTTP2" "$TEMP_DIR" > /dev/null; then
        echo -e "${GREEN}✓ HTTP/2 protocol configuration found${NC}"
    else
        echo -e "${YELLOW}⚠ HTTP/2 protocol configuration not found${NC}"
    fi
    
    # Check for ServiceProfile
    if grep -r "kind: ServiceProfile" "$TEMP_DIR" > /dev/null; then
        echo -e "${GREEN}✓ ServiceProfile resources found${NC}"
    else
        echo -e "${YELLOW}⚠ ServiceProfile resources not found${NC}"
    fi
    
    # Check for Server resources
    if grep -r "kind: Server" "$TEMP_DIR" > /dev/null; then
        echo -e "${GREEN}✓ Server resources found${NC}"
    else
        echo -e "${YELLOW}⚠ Server resources not found${NC}"
    fi
else
    echo -e "${RED}✗ Chart templating failed${NC}"
    exit 1
fi
echo ""

# Offer to deploy
echo "========================================="
echo "Verification Summary"
echo "========================================="
echo ""
echo "The Helm chart has been validated successfully!"
echo ""
echo "To deploy with service mesh enabled:"
echo ""
echo "  helm install go-base ./go-base \\"
echo "    --set serviceMesh.enabled=true \\"
echo "    --namespace production \\"
echo "    --create-namespace"
echo ""
echo "To verify after deployment:"
echo ""
echo "  # Check pod injection"
echo "  kubectl get pods -n production -o jsonpath='{range .items[*]}{.metadata.name}{\"\t\"}{.spec.containers[*].name}{\"\n\"}{end}'"
echo ""
echo "  # Check mTLS status"
echo "  linkerd edges deployment -n production"
echo ""
echo "  # View metrics"
echo "  linkerd viz stat deployment -n production"
echo ""
echo "  # Open dashboard"
echo "  linkerd viz dashboard"
echo ""

# Cleanup
rm -rf "$TEMP_DIR"
