# Service Mesh Setup for go-base

This document provides guidance on setting up and using the Linkerd service mesh for HTTP/2 load balancing in the go-base Kubernetes deployment.

## Overview

The go-base Helm chart includes optional Linkerd service mesh integration to enable:

- **HTTP/2 Load Balancing**: Intelligent Layer 7 load balancing with connection multiplexing
- **Latency-Aware Routing**: EWMA algorithm distributes traffic to fastest endpoints
- **Automatic Retries**: Configurable retry policies for failed requests
- **mTLS Encryption**: Mutual TLS for secure service-to-service communication
- **Observability**: Built-in metrics, distributed tracing, and traffic visualization

## Prerequisites

### 1. Install Linkerd CLI

```bash
curl -sL https://run.linkerd.io/install | sh
export PATH=$PATH:$HOME/.linkerd2/bin
```

### 2. Verify Kubernetes Cluster

```bash
linkerd check --pre
```

Requirements:
- Kubernetes 1.24 or later
- Cluster must support `ServiceAccount` token volume projection

### 3. Install Linkerd Control Plane

```bash
# Install CRDs
linkerd install --crds | kubectl apply -f -

# Install control plane
linkerd install | kubectl apply -f -

# Verify installation
linkerd check
```

### 4. Install Linkerd Viz (Optional - for observability)

```bash
linkerd viz install | kubectl apply -f -
linkerd viz check
```

## Configuration

### Enable Service Mesh

In `values.yaml`:

```yaml
serviceMesh:
  enabled: true
  provider: linkerd
```

### HTTP/2 Settings

```yaml
serviceMesh:
  linkerd:
    http2:
      enabled: true
      autoUpgrade: true
      maxConcurrentStreams: 100
```

### Retry Configuration

```yaml
serviceMesh:
  linkerd:
    retries:
      enabled: true
      maxRetries: 3
      retryableStatusCodes: [502, 503, 504]
      retryBudget: 0.2  # Max 20% additional load from retries
```

### Timeout Settings

```yaml
serviceMesh:
  linkerd:
    timeouts:
      request: "10s"
      idle: "60s"
      responseHeader: "10s"
```

### mTLS Configuration

```yaml
serviceMesh:
  linkerd:
    mtls:
      enabled: true
      mode: strict  # or permissive
```

**Modes:**
- `strict`: All connections must use mTLS (recommended for production)
- `permissive`: Allows both mTLS and plaintext connections

### Proxy Resources

```yaml
serviceMesh:
  linkerd:
    proxy:
      resources:
        cpu:
          request: "100m"
          limit: "500m"
        memory:
          request: "50Mi"
          limit: "128Mi"
```

## Deployment

### Install with Service Mesh

```bash
helm install go-base ./go-base \
  --set serviceMesh.enabled=true \
  --namespace production \
  --create-namespace
```

### Verify Injection

Check that pods have Linkerd proxy injected:

```bash
kubectl get pods -n production -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}'
```

Each pod should show 2 containers: the application container and `linkerd-proxy`.

### Check mTLS Status

```bash
linkerd edges deployment -n production
```

All connections should show as "secured".

## Observability

### Access Dashboard

```bash
linkerd viz dashboard
```

Opens the Linkerd dashboard in your browser showing:
- Service topology
- Success rates
- Latency percentiles (p50, p95, p99)
- Request rates

### View Live Traffic

```bash
# Watch traffic to a specific deployment
linkerd viz tap deployment/go-base -n production

# Filter by HTTP method
linkerd viz tap deployment/go-base -n production --method GET

# Filter by path
linkerd viz tap deployment/go-base -n production --path /api/v1
```

### Check Service Metrics

```bash
# Overall service stats
linkerd viz stat deployment -n production

# Detailed route metrics
linkerd viz routes deployment/go-base -n production
```

### View Service Profile

```bash
kubectl get serviceprofile -n production
kubectl describe serviceprofile go-base.production.svc.cluster.local -n production
```

## Performance Tuning

### Connection Pooling

HTTP/2 multiplexes multiple requests over a single connection. Adjust concurrent streams:

```yaml
serviceMesh:
  linkerd:
    http2:
      maxConcurrentStreams: 100  # Increase for high-throughput services
```

### Load Balancing Algorithm

Linkerd uses EWMA (Exponentially Weighted Moving Average) by default, which routes requests to the fastest endpoints based on recent latency:

```yaml
serviceMesh:
  linkerd:
    loadBalancing:
      algorithm: ewma
```

### Retry Budget

Prevent retry storms by limiting additional load:

```yaml
serviceMesh:
  linkerd:
    retries:
      retryBudget: 0.2  # Max 20% additional load
```

## Troubleshooting

### Pods Not Getting Injected

**Check namespace annotation:**
```bash
kubectl get namespace production -o yaml | grep linkerd.io/inject
```

**Manually inject:**
```bash
kubectl get deployment go-base -n production -o yaml | linkerd inject - | kubectl apply -f -
```

### HTTP/2 Not Working

**Verify protocol detection:**
```bash
kubectl logs -n production deployment/go-base -c linkerd-proxy | grep -i http2
```

**Check service appProtocol:**
```bash
kubectl get service go-base -n production -o yaml | grep appProtocol
```

Should show `appProtocol: HTTP2`.

### High Latency

**Check proxy overhead:**
```bash
linkerd viz stat deployment/go-base -n production --from deployment/other-service
```

**Adjust proxy resources:**
```yaml
serviceMesh:
  linkerd:
    proxy:
      resources:
        cpu:
          limit: "1000m"  # Increase if CPU-bound
```

### mTLS Failures

**Check certificate status:**
```bash
linkerd identity -n production
```

**Verify Server resources:**
```bash
kubectl get server -n production
kubectl describe server go-base-http -n production
```

### Retries Not Working

**Check ServiceProfile:**
```bash
kubectl get serviceprofile -n production
kubectl describe serviceprofile go-base.production.svc.cluster.local -n production
```

Ensure routes have `isRetryable: true`.

## Disabling Service Mesh

To disable the service mesh without uninstalling Linkerd:

```yaml
serviceMesh:
  enabled: false
```

Then upgrade the release:

```bash
helm upgrade go-base ./go-base -n production
```

Pods will be recreated without Linkerd proxy injection.

## Best Practices

1. **Start with permissive mTLS** in development, switch to strict in production
2. **Monitor retry rates** to avoid retry storms
3. **Set appropriate timeouts** based on your application's SLAs
4. **Use ServiceProfiles** for fine-grained route control
5. **Enable observability** (Linkerd Viz) for debugging
6. **Test thoroughly** before enabling in production
7. **Set resource limits** on proxies to prevent resource exhaustion

## Additional Resources

- [Linkerd Documentation](https://linkerd.io/docs/)
- [HTTP/2 Load Balancing Guide](https://linkerd.io/2/features/load-balancing/)
- [Automatic Retries](https://linkerd.io/2/features/retries-and-timeouts/)
- [mTLS Guide](https://linkerd.io/2/features/automatic-mtls/)
- [Performance Tuning](https://linkerd.io/2/reference/proxy-configuration/)
