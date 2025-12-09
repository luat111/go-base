# Search Engine Guide

## Overview

The go-base Helm chart supports two search engine options: **Manticore Search** and **Elasticsearch**. You can deploy either one and easily switch between them for performance comparison.

## Quick Comparison

| Feature | Manticore Search | Elasticsearch |
|---------|-----------------|---------------|
| **Performance** | 2-16x faster | Slower |
| **Language** | C++ | Java |
| **Memory Usage** | Lower (~1-2GB) | Higher (~4GB+) |
| **SQL Support** | ✅ Native | ❌ Limited |
| **JSON API** | ✅ Yes | ✅ Yes |
| **Ecosystem** | Growing | Mature |
| **Best For** | High-speed search, logs | Complex analytics, ML |

## Configuration

### Using Manticore Search (Default)

```yaml
# values.yaml
search:
  provider: manticore
  manticore:
    enabled: true
    replicaCount: 1
    persistence:
      size: 10Gi
```

### Using Elasticsearch

```yaml
# values.yaml
search:
  provider: elasticsearch
  manticore:
    enabled: false  # Disable Manticore
  elasticsearch:
    enabled: true
    replicaCount: 1
    persistence:
      size: 10Gi
```

## Deployment

### Deploy with Manticore

```bash
helm install go-base ./go-base \
  --set search.provider=manticore \
  --namespace production
```

### Deploy with Elasticsearch

```bash
helm install go-base ./go-base \
  --set search.provider=elasticsearch \
  --set search.manticore.enabled=false \
  --set search.elasticsearch.enabled=true \
  --namespace production
```

## Switching Between Engines

### From Manticore to Elasticsearch

```bash
helm upgrade go-base ./go-base \
  --set search.provider=elasticsearch \
  --set search.manticore.enabled=false \
  --set search.elasticsearch.enabled=true \
  --namespace production
```

### From Elasticsearch to Manticore

```bash
helm upgrade go-base ./go-base \
  --set search.provider=manticore \
  --set search.manticore.enabled=true \
  --set search.elasticsearch.enabled=false \
  --namespace production
```

> [!WARNING]
> Switching engines will NOT migrate data automatically. You'll need to re-index your data.

## Application Integration

The application receives these environment variables:

```bash
SEARCH_PROVIDER=manticore  # or elasticsearch
SEARCH_HOST=go-base-manticore  # or go-base-elasticsearch
SEARCH_PORT=9308  # HTTP port (9200 for Elasticsearch)
SEARCH_URL=http://go-base-manticore:9308
SEARCH_SQL_PORT=9306  # Only for Manticore
```

### Example Usage in Go

```go
import (
    "fmt"
    "os"
)

func main() {
    provider := os.Getenv("SEARCH_PROVIDER")
    searchURL := os.Getenv("SEARCH_URL")
    
    fmt.Printf("Using %s at %s\n", provider, searchURL)
}
```

## Manticore Search

### Ports

- **9306**: MySQL protocol (SQL queries)
- **9308**: HTTP/JSON API
- **9312**: Binary protocol

### Querying via SQL

```bash
# Connect via MySQL client
kubectl exec -it deployment/go-base-manticore -n production -- \
  mysql -h127.0.0.1 -P9306

# Example queries
SHOW TABLES;
CREATE TABLE products (title text, price float);
INSERT INTO products VALUES (1, 'Product 1', 19.99);
SELECT * FROM products WHERE MATCH('Product');
```

### Querying via HTTP

```bash
# Search query
curl -X POST http://go-base-manticore:9308/search \
  -H "Content-Type: application/json" \
  -d '{
    "index": "products",
    "query": {
      "match": { "title": "Product" }
    }
  }'

# Insert document
curl -X POST http://go-base-manticore:9308/insert \
  -H "Content-Type: application/json" \
  -d '{
    "index": "products",
    "doc": {
      "title": "New Product",
      "price": 29.99
    }
  }'
```

### Configuration

Adjust in `values.yaml`:

```yaml
search:
  manticore:
    config:
      maxMatches: 1000  # Max results
      rtMemLimit: "128M"  # Memory for RT indexes
      threads: 4  # Number of threads
```

## Elasticsearch

### Ports

- **9200**: REST API (HTTP)
- **9300**: Transport (node-to-node)

### Querying via REST API

```bash
# Cluster health
curl http://go-base-elasticsearch:9200/_cluster/health

# Create index
curl -X PUT http://go-base-elasticsearch:9200/products

# Index document
curl -X POST http://go-base-elasticsearch:9200/products/_doc \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Product 1",
    "price": 19.99
  }'

# Search
curl -X GET http://go-base-elasticsearch:9200/products/_search \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "match": {
        "title": "Product"
      }
    }
  }'
```

### Configuration

Adjust in `values.yaml`:

```yaml
search:
  elasticsearch:
    config:
      clusterName: go-base-search
      javaOpts: "-Xms2g -Xmx2g"  # Heap size
      xpackSecurityEnabled: false  # Enable for production
```

## Performance Tuning

### Manticore Search

**Increase threads for better query performance:**
```yaml
search:
  manticore:
    config:
      threads: 8  # Use more CPU cores
```

**Increase memory for larger indexes:**
```yaml
search:
  manticore:
    resources:
      limits:
        memory: 4Gi
```

### Elasticsearch

**Increase heap size for better performance:**
```yaml
search:
  elasticsearch:
    config:
      javaOpts: "-Xms4g -Xmx4g"
    resources:
      limits:
        memory: 8Gi  # Heap + overhead
```

## Troubleshooting

### Manticore Not Starting

**Check logs:**
```bash
kubectl logs deployment/go-base-manticore -n production
```

**Common issues:**
- Insufficient memory
- Port conflicts
- Configuration syntax errors

### Elasticsearch Not Starting

**Check logs:**
```bash
kubectl logs statefulset/go-base-elasticsearch -n production
```

**Common issues:**
- `vm.max_map_count` too low (handled by init container)
- Insufficient heap memory
- Disk space issues

### Connection Issues

**Test connectivity:**
```bash
# Manticore
kubectl exec -it deployment/go-base -n production -- \
  curl http://go-base-manticore:9308/sql -d "query=SHOW TABLES"

# Elasticsearch
kubectl exec -it deployment/go-base -n production -- \
  curl http://go-base-elasticsearch:9200/_cluster/health
```

## Best Practices

1. **Start with Manticore** for better performance
2. **Use Elasticsearch** if you need advanced analytics or ML features
3. **Enable persistence** in production
4. **Set resource limits** to prevent OOM
5. **Monitor resource usage** with `kubectl top pods`
6. **Backup indices** regularly
7. **Test queries** before deploying to production

## Benchmarking

To compare performance:

1. Deploy with Manticore
2. Index sample data
3. Run test queries and measure latency
4. Switch to Elasticsearch
5. Index same data
6. Run same queries and compare

Example benchmark script:

```bash
#!/bin/bash
# Benchmark search engines

echo "Testing Manticore..."
time for i in {1..100}; do
  curl -s http://go-base-manticore:9308/search \
    -d '{"index":"test","query":{"match_all":{}}}' > /dev/null
done

echo "Testing Elasticsearch..."
time for i in {1..100}; do
  curl -s http://go-base-elasticsearch:9200/test/_search \
    -d '{"query":{"match_all":{}}}' > /dev/null
done
```

## Additional Resources

- [Manticore Search Documentation](https://manual.manticoresearch.com/)
- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Manticore vs Elasticsearch Benchmarks](https://manticoresearch.com/blog/manticore-vs-elasticsearch/)
