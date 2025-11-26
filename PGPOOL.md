# Pgpool-II Configuration

## What was added to database.yml

Added **pgpool** service for PostgreSQL connection pooling and load balancing.

## Features

- ✅ **Connection pooling** - Manages database connections efficiently
- ✅ **Load balancing** - Distributes queries across PostgreSQL instances
- ✅ **Health checks** - Monitors both PostgreSQL and Pgpool health
- ✅ **Automatic failover** - Ready for high availability setups

## Configuration

### Pgpool Service Details

- **Image**: `bitnami/pgpool:latest`
- **Port**: `5433` (configurable via `PGPOOL_PORT`)
- **Backend**: Connected to `postgres:5432`
- **Max connections**: 15 per child process
- **Load balancing**: Enabled

### Key Environment Variables

```yaml
PGPOOL_BACKEND_NODES=0:postgres:5432    # Backend PostgreSQL server
PGPOOL_MAX_POOL=15                       # Max connections per child
PGPOOL_ENABLE_LOAD_BALANCING=yes        # Enable load balancing
PGPOOL_CHILD_LIFE_TIME=300              # Child process lifetime (seconds)
```

## Required .env Variables

Add these to your `.env` file:

```bash
# Pgpool Configuration (optional, has defaults)
PGPOOL_PORT=5433
PGPOOL_ADMIN_USER=admin
PGPOOL_ADMIN_PWD=adminpassword
```

## Usage

### Start Services

```bash
docker-compose -f database.yml up -d
```

### Connect to PostgreSQL via Pgpool

Instead of connecting directly to PostgreSQL on port 5432, connect through Pgpool:

```bash
# Direct PostgreSQL connection (port 5432)
psql -h localhost -p 5432 -U your_user -d your_db

# Via Pgpool (port 5433) - recommended
psql -h localhost -p 5433 -U your_user -d your_db
```

### Update Your Application Connection String

Change your database connection from:
```
postgres://user:pass@localhost:5432/dbname
```

To:
```
postgres://user:pass@localhost:5433/dbname
```

## Benefits

1. **Connection Pooling** - Reduces connection overhead
2. **Better Performance** - Reuses existing connections
3. **Scalability** - Ready for read replicas and load balancing
4. **High Availability** - Supports automatic failover
5. **Query Caching** - Can cache SELECT query results

## Health Checks

Both services now have health checks:
- **PostgreSQL**: `pg_isready` command
- **Pgpool**: Built-in health check script

Pgpool waits for PostgreSQL to be healthy before starting.

## Monitoring

Check service health:
```bash
docker-compose -f database.yml ps
docker-compose -f database.yml logs pgpool
```

## Advanced Configuration

For production use, you can add:
- Multiple PostgreSQL backends for load balancing
- Streaming replication check
- Watchdog for automatic failover
- Connection limits and timeouts
- Query caching settings

See [Pgpool-II documentation](https://www.pgpool.net/docs/latest/en/html/) for advanced configurations.
