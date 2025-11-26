#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-go-base}"
RELEASE_NAME="${RELEASE_NAME:-go-base}"
CHART_PATH="./deployment/helm/go-base"
VALUES_FILE="${VALUES_FILE:-values.yaml}"

echo -e "${GREEN}=== Go-Base Helm Deployment ===${NC}"

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo -e "${RED}Error: Helm is not installed${NC}"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    exit 1
fi

# Create namespace if it doesn't exist
echo -e "${YELLOW}Creating namespace: $NAMESPACE${NC}"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Add labels to namespace
kubectl label namespace $NAMESPACE environment=${ENVIRONMENT:-production} --overwrite

# Install or upgrade the release
echo -e "${YELLOW}Deploying Helm chart...${NC}"
helm upgrade --install $RELEASE_NAME $CHART_PATH \
  --namespace $NAMESPACE \
  --values $CHART_PATH/$VALUES_FILE \
  --create-namespace \
  --wait \
  --timeout 10m

echo -e "${GREEN}Deployment successful!${NC}"

# Show deployment status
echo -e "${YELLOW}Checking deployment status...${NC}"
kubectl get pods -n $NAMESPACE
kubectl get svc -n $NAMESPACE
kubectl get ingress -n $NAMESPACE

echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo -e "Release: ${GREEN}$RELEASE_NAME${NC}"
echo -e "Namespace: ${GREEN}$NAMESPACE${NC}"
echo -e ""
echo -e "To check logs:"
echo -e "  kubectl logs -f deployment/$RELEASE_NAME -n $NAMESPACE"
echo -e ""
echo -e "To check status:"
echo -e "  helm status $RELEASE_NAME -n $NAMESPACE"
