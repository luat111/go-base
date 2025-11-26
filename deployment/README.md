# Kubernetes Deployment Guide

## Overview

This directory contains Helm charts and deployment scripts for deploying the go-base application to Kubernetes with a complete infrastructure stack.

## Architecture

The deployment includes:
- **Go Application**: Scalable microservice with HPA
- **PostgreSQL**: Primary database with StatefulSet
- **Pgpool**: Connection pooling and load balancing
- **Redis**: Caching layer
- **RabbitMQ**: Message queue with clustering
- **Ingress**: NGINX ingress controller with TLS

## Directory Structure

```
deployment/
├── helm/
│   └── go-base/
│       ├── Chart.yaml              # Helm chart metadata
│       ├── values.yaml             # Default values
│       └── templates/              # Kubernetes manifests
│           ├── deployment.yaml     # App deployment
│           ├── service.yaml        # App service
│           ├── ingress.yaml        # Ingress/Gateway
│           ├── hpa.yaml            # Horizontal Pod Autoscaler
│           ├── configmap.yaml      # Configuration
│           ├── secret.yaml         # Secrets
│           ├── postgresql.yaml     # PostgreSQL StatefulSet
│           ├── pgpool.yaml         # Pgpool deployment
│           ├── redis.yaml          # Redis StatefulSet
│           └── rabbitmq.yaml       # RabbitMQ StatefulSet
├── configs/
│   ├── values-production.yaml      # Production overrides
│   └── values-staging.yaml         # Staging overrides
└── scripts/
    ├── deploy.sh                   # Deployment script
    └── rollback.sh                 # Rollback script
```

## Prerequisites

1. **Kubernetes Cluster** (v1.24+)
   - AKS, EKS, GKE, or self-hosted
   - kubectl configured

2. **Helm** (v3.0+)
   ```bash
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
   ```

3. **NGINX Ingress Controller**
   ```bash
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm install nginx-ingress ingress-nginx/ingress-nginx \
     --namespace ingress-nginx --create-namespace
   ```

4. **Cert-Manager** (for TLS)
   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   ```

## Quick Start

### 1. Build and Push Docker Image

```bash
# Build image
docker build -t your-registry/go-base:v1.0.0 .

# Push to registry
docker push your-registry/go-base:v1.0.0
```

### 2. Update Configuration

Edit `deployment/configs/values-production.yaml`:

```yaml
app:
  image:
    repository: your-registry/go-base
    tag: "v1.0.0"

ingress:
  hosts:
    - host: api.your-domain.com

secrets:
  PG_PWD: "your-secure-password"
  CACHE_PWD: "your-redis-password"
  RMQ_PWD: "your-rabbitmq-password"
  JWT_SECRET: "your-jwt-secret"
```

### 3. Deploy to Kubernetes

```bash
# Make scripts executable
chmod +x deployment/scripts/*.sh

# Deploy to production
NAMESPACE=go-base-prod \
VALUES_FILE=../configs/values-production.yaml \
./deployment/scripts/deploy.sh

# Or deploy to staging
NAMESPACE=go-base-staging \
VALUES_FILE=../configs/values-staging.yaml \
./deployment/scripts/deploy.sh
```

## Deployment Commands

### Install/Upgrade

```bash
# Using script
./deployment/scripts/deploy.sh

# Or manually with Helm
helm upgrade --install go-base ./deployment/helm/go-base \
  --namespace go-base \
  --values ./deployment/configs/values-production.yaml \
  --create-namespace \
  --wait
```

### Check Status

```bash
# Helm release status
helm status go-base -n go-base

# Pod status
kubectl get pods -n go-base

# Service status
kubectl get svc -n go-base

# Ingress status
kubectl get ingress -n go-base
```

### View Logs

```bash
# Application logs
kubectl logs -f deployment/go-base -n go-base

# PostgreSQL logs
kubectl logs -f statefulset/go-base-postgresql -n go-base

# Redis logs
kubectl logs -f statefulset/go-base-redis -n go-base

# RabbitMQ logs
kubectl logs -f statefulset/go-base-rabbitmq -n go-base
```

### Rollback

```bash
# List revisions
helm history go-base -n go-base

# Rollback to previous version
./deployment/scripts/rollback.sh 1

# Or manually
helm rollback go-base 1 -n go-base
```

### Uninstall

```bash
helm uninstall go-base -n go-base
kubectl delete namespace go-base
```

## Configuration

### Environment-Specific Values

- **Production**: `deployment/configs/values-production.yaml`
- **Staging**: `deployment/configs/values-staging.yaml`
- **Development**: Use default `values.yaml`

### Key Configuration Options

#### Application

```yaml
app:
  replicaCount: 3                    # Number of replicas
  image:
    repository: your-registry/go-base
    tag: "v1.0.0"
  resources:
    limits:
      cpu: 1000m
      memory: 512Mi
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
```

#### PostgreSQL

```yaml
postgresql:
  enabled: true
  auth:
    username: postgres
    password: "changeme"
    database: go_base
  primary:
    persistence:
      enabled: true
      size: 10Gi
```

#### Redis

```yaml
redis:
  enabled: true
  auth:
    enabled: true
    password: "changeme"
  master:
    persistence:
      enabled: true
      size: 5Gi
```

#### Ingress

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: go-base-tls
      hosts:
        - api.example.com
```

## Secrets Management

### Option 1: Kubernetes Secrets (Basic)

Update secrets in values file (not recommended for production):

```yaml
secrets:
  PG_PWD: "password"
  JWT_SECRET: "secret"
```

### Option 2: External Secrets Operator (Recommended)

```bash
# Install External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets-system --create-namespace

# Create SecretStore (example with AWS Secrets Manager)
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets
  namespace: go-base
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
EOF
```

### Option 3: Sealed Secrets

```bash
# Install Sealed Secrets
helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
helm install sealed-secrets sealed-secrets/sealed-secrets -n kube-system

# Create sealed secret
kubeseal --format=yaml < secret.yaml > sealed-secret.yaml
```

## Monitoring

### Prometheus & Grafana

```bash
# Install kube-prometheus-stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
```

### Application Metrics

Add Prometheus annotations to deployment:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

## Scaling

### Manual Scaling

```bash
kubectl scale deployment go-base --replicas=5 -n go-base
```

### Horizontal Pod Autoscaler

HPA is enabled by default and scales based on CPU/Memory:

```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

## Troubleshooting

### Pod Not Starting

```bash
# Describe pod
kubectl describe pod <pod-name> -n go-base

# Check events
kubectl get events -n go-base --sort-by='.lastTimestamp'

# Check logs
kubectl logs <pod-name> -n go-base
```

### Database Connection Issues

```bash
# Test PostgreSQL connection
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -n go-base -- \
  psql -h go-base-pgpool -U postgres -d go_base

# Check Pgpool status
kubectl logs deployment/go-base-pgpool -n go-base
```

### Ingress Not Working

```bash
# Check ingress
kubectl describe ingress go-base -n go-base

# Check NGINX controller logs
kubectl logs -n ingress-nginx deployment/nginx-ingress-controller

# Test from inside cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -n go-base -- \
  curl http://go-base:8080/health
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy to Kubernetes

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build and push image
        run: |
          docker build -t ${{ secrets.REGISTRY }}/go-base:${{ github.sha }} .
          docker push ${{ secrets.REGISTRY }}/go-base:${{ github.sha }}
      
      - name: Deploy to Kubernetes
        run: |
          helm upgrade --install go-base ./deployment/helm/go-base \
            --namespace go-base \
            --set app.image.tag=${{ github.sha }} \
            --values ./deployment/configs/values-production.yaml
```

## Best Practices

1. **Use specific image tags** - Avoid `latest` in production
2. **Set resource limits** - Prevent resource exhaustion
3. **Enable health checks** - Ensure proper pod lifecycle
4. **Use secrets management** - Never commit secrets to git
5. **Enable monitoring** - Track application performance
6. **Regular backups** - Backup PostgreSQL and Redis data
7. **Test rollbacks** - Verify rollback procedures work
8. **Use namespaces** - Isolate environments
9. **Enable TLS** - Secure all external traffic
10. **Review security** - Run security scans regularly

## Support

For issues or questions, please refer to:
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- Project repository issues
