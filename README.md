# CDC Cascade

A production-ready reference implementation for Change Data Capture (CDC) using Debezium, showcasing real-time data streaming from PostgreSQL to consumers via Kafka.

## What's This?

This repository demonstrates a complete CDC pipeline with a practical architecture that you can use as a blueprint for your own projects. It captures database changes in real-time and streams them to downstream consumers—perfect for microservices synchronization, audit logging, cache invalidation, or building event-driven systems.

**Use this when you need to:**

- Sync data between microservices without tight coupling
- Build real-time analytics pipelines
- Maintain materialized views or caches
- Implement audit logs or event sourcing
- Learn CDC patterns hands-on

## Architecture

View the [complete sequence flow diagram](/sequence-flows.mmd) showing all interaction patterns: cache miss, cache hit, CDC invalidation, and cache rebuild.

---

### Components

- **PostgreSQL**: Source of truth for all data
- **Debezium**: CDC engine capturing PostgreSQL WAL changes
- **Kafka**: Event streaming backbone for change events
- **Go Application** (runs two goroutines):
  - **Fiber REST API**: Handles client requests with cache-aside logic
  - **CDC Consumer**: Listens to Kafka and invalidates Redis on changes
- **Redis**: Cache layer with automatic CDC-driven invalidation

## Project Structure

```
.
├── config/                     # Database & cache configuration
├── controllers/                # HTTP request handlers
├── models/                     # Domain models & schemas
├── queue/                      # CDC consumer & event processing
│   ├── cdc.go                  # Event handlers & business logic
│   └── runner.go               # Consumer lifecycle management
├── scripts/                    # Setup & initialization
│   ├── debezium-setup.sh       # Connector registration
│   └── init.sql                # Database schema & seed data
├── docker-compose.yml          # Service orchestration
├── Dockerfile                  # Multi-stage build for API
├── main.go                     # Entry point (API + CDC consumer)
└── Makefile                    # Build, run, and management commands
```

## Configuration

### Environment Variables

```sh
# Postgres
DB_HOST=cdc-cascade-postgres
DB_USER=admin
DB_NAME=cdc-cascade-db
DB_PASS=adminpw
DB_PORT=5432

# Redis
REDIS_HOST=cdc-cascade-redis
REDIS_PORT=6379
REDIS_PASS=rootpw

# API
API_HOST=cdc-cascade-api
API_PORT=8080

# Kafka
KAFKA_HOST=cdc-cascade-kafka
KAFKA_CONSUMER_GROUP=cdc-cascade-kafka-consumers
KAFKA_CDC_TOPIC=cdc-cascade-postgres.public.sinners
KAFKA_BROKER_PORT=9092
KAFKA_CONTROLLER_PORT=9093

# Debezium
DEBEZIUM_HOST=cdc-cascade-debezium
DEBEZIUM_PORT=8083
```

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Make
- Go 1.25+ (for local development)
- 4GB RAM recommended

### Run the Stack

```bash
# Clone the repository
git clone https://github.com/tr1sm0s1n/cdc-cascade.git
cd cdc-cascade

# Start all services
make up

# Load .env if necessary
source .env

# View logs
docker logs -f $API_HOST
```

### Debezium Connector

To register connector:

```bash
make connect
```

### Verify It's Working

**1. Enter database:**

```bash
make enter
```

**1. Insert new data:**

```sql
INSERT INTO sinners (code, "name", class, libram, tendency)
VALUES
    (5, 'Augustus', 'S', 'War', 'Reticle');
```

**2. Check Kafka for CDC events:**

```bash
docker exec -it $KAFKA_HOST /opt/kafka/bin/kafka-console-consumer.sh \
  --topic $DB_HOST.public.sinners \
  --bootstrap-server localhost:$KAFKA_BROKER_PORT \
  --from-beginning
```

**3. Query the API:**

```bash
curl http://127.0.0.1:$API_PORT/api/v1/sinners/read/5
```

**4. Verify Redis cache:**

```bash
docker exec -it $REDIS_HOST redis-cli GET 5
```

## Learning Path

**For Beginners:**

1. Start the stack and observe logs
2. Make database changes, watch Kafka events
3. Trace how events flow to Redis
4. Experiment with the API endpoints

**For Advanced Users:**

1. Add custom CDC event handlers
2. Implement transformation logic
3. Add new tables to the CDC pipeline
4. Integrate with your own services

## Troubleshooting

**Debezium not capturing changes?**

- Verify PostgreSQL has `wal_level = logical`
- Check table has PRIMARY KEY
- Ensure table is in `table.include.list`

**Kafka consumer lag?**

- Check consumer group: `kafka-consumer-groups.sh --bootstrap-server localhost:$KAFKA_BROKER_PORT --describe --group $KAFKA_CONSUMER_GROUP`
- Increase consumer parallelism if needed

**Redis not updating?**

- Verify CDC consumer is running in logs
- Check Redis connectivity from API container

## Resources

- [Debezium Documentation](https://debezium.io/documentation/)
- [Kafka Connect Deep Dive](https://kafka.apache.org/documentation/#connect)
- [PostgreSQL Logical Replication](https://www.postgresql.org/docs/current/logical-replication.html)
- [Fiber Framework](https://docs.gofiber.io/)
- [Redis with Go](https://redis.io/docs/latest/integrate/go-redis/)

---

**Built with ❤️ for developers learning CDC patterns**

_Questions? Open an issue or start a discussion!_
