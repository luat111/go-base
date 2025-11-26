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
REVISION="${1:-0}"

echo -e "${YELLOW}=== Helm Rollback ===${NC}"

if [ "$REVISION" -eq 0 ]; then
    echo -e "${RED}Error: Please specify a revision number${NC}"
    echo -e "Usage: ./rollback.sh <revision>"
    echo -e ""
    echo -e "Available revisions:"
    helm history $RELEASE_NAME -n $NAMESPACE
    exit 1
fi

echo -e "${YELLOW}Rolling back to revision $REVISION...${NC}"
helm rollback $RELEASE_NAME $REVISION -n $NAMESPACE --wait

echo -e "${GREEN}Rollback successful!${NC}"

# Show current status
kubectl get pods -n $NAMESPACE
