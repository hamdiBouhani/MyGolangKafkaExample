#!/usr/bin/env bash
set -euo pipefail

echo "Starting infrastructure..."
docker compose up -d zookeeper kafka schema-registry

echo "Waiting for Schema Registry..."
until curl -sf http://localhost:8081/subjects >/dev/null; do sleep 2; done

echo "Starting producer, Connect, and source DB..."
docker compose up -d postgres kafka-connect producer

echo "Waiting for Connect REST..."
until curl -sf http://localhost:8083/ >/dev/null; do sleep 2; done

echo "Deploying Debezium connector..."
./cdc/scripts/deploy-connector.sh cdc/connectors/postgres-cdc.json

echo "Verifying..."
./cdc/scripts/check-status.sh

echo "Pipeline is live."
echo "   Producer metrics:   http://localhost:9090/metrics"
echo "   Connect REST:       http://localhost:8083/connectors"
echo "   Kafka UI:           http://localhost:8080"
echo "   Schema Registry UI: http://localhost:8000"
echo "   Connect UI:         http://localhost:8001"