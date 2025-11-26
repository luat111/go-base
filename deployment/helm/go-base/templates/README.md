# Helm Templates Organization

The templates are organized into logical categories for better maintainability:

## Directory Structure

```
templates/
├── _helpers.tpl           # Template helpers and functions
├── app/                   # Application resources
│   ├── deployment.yaml    # Main application deployment
│   ├── service.yaml       # Application service
│   ├── hpa.yaml          # Horizontal Pod Autoscaler
│   ├── configmap.yaml    # Application configuration
│   ├── secret.yaml       # Application secrets
│   └── serviceaccount.yaml # Service account for RBAC
├── ingress/              # Gateway/Ingress resources
│   └── ingress.yaml      # NGINX Ingress with TLS
├── database/             # Database resources
│   ├── postgresql.yaml   # PostgreSQL StatefulSet
│   └── pgpool.yaml       # Pgpool connection pooler
└── dependencies/         # External dependencies
    ├── redis.yaml        # Redis cache
    └── rabbitmq.yaml     # RabbitMQ message queue
```

## Categories

### App (6 files)
Core application resources:
- **deployment.yaml** - Main Go application deployment with 3 replicas
- **service.yaml** - ClusterIP service exposing port 8080
- **hpa.yaml** - Auto-scaling based on CPU/memory (3-10 replicas)
- **configmap.yaml** - Non-sensitive configuration (endpoints, settings)
- **secret.yaml** - Sensitive data (passwords, keys, tokens)
- **serviceaccount.yaml** - Pod identity for RBAC

### Ingress (1 file)
External access:
- **ingress.yaml** - NGINX Ingress controller with TLS/SSL support

### Database (2 files)
PostgreSQL infrastructure:
- **postgresql.yaml** - PostgreSQL 16 StatefulSet with persistence
- **pgpool.yaml** - Connection pooling and load balancing

### Dependencies (2 files)
External services:
- **redis.yaml** - Redis 7 for caching with persistence
- **rabbitmq.yaml** - RabbitMQ 3 for messaging with clustering

## Benefits of This Structure

1. **Clear Separation** - Easy to find related resources
2. **Selective Deployment** - Can disable entire categories via values
3. **Better Maintenance** - Logical grouping reduces confusion
4. **Team Collaboration** - Different teams can own different directories
5. **Easier Debugging** - Quickly locate issues by category

## Disabling Components

You can disable entire categories in `values.yaml`:

```yaml
# Disable database components
postgresql:
  enabled: false
pgpool:
  enabled: false

# Disable dependencies
redis:
  enabled: false
rabbitmq:
  enabled: false

# Disable ingress
ingress:
  enabled: false
```

## Adding New Resources

When adding new templates, place them in the appropriate directory:

- **App resources** → `app/`
- **Gateway/routing** → `ingress/`
- **Database/storage** → `database/`
- **External services** → `dependencies/`

Example: Adding a monitoring sidecar would go in `app/monitoring-sidecar.yaml`
