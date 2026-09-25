# 1. Start infrastructure
docker compose up -d zookeeper kafka schema-registry

# 2. Build & run the producer (10 messages then exit)
docker compose up --build producer

# Or run locally against Docker infra:
go run ./cmd/producer

# 3. Observe metrics
curl http://localhost:9090/metrics
curl http://localhost:9090/healthz

# 4. Run tests
go test ./...


# Kafka Platform

A complete streaming platform: Go Avro producer + Debezium CDC + (future) ksqlDB + ClickHouse.

## Architecture
