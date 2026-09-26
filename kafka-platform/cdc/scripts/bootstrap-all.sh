#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

COMPOSE="docker compose"
docker compose version >/dev/null 2>&1 || COMPOSE="docker-compose"

echo "▶ Starting the full stack (healthchecks gate startup order)..."
$COMPOSE up -d

echo "▶ Deploying Debezium connector..."
CONNECT_URL="${CONNECT_URL:-http://localhost:8083}" \
    ./cdc/scripts/deploy-connector.sh cdc/connectors/postgres-cdc.json

echo "▶ Verifying connector health..."
CONNECT_URL="${CONNECT_URL:-http://localhost:8083}" \
    ./cdc/scripts/check-status.sh

echo
echo "Pipeline is live."
echo "   Kafka UI:           http://localhost:8080"
echo "   Schema Registry UI: http://localhost:8000"
echo "   Connect UI:         http://localhost:8001"
echo "   Producer metrics:   http://localhost:9090/metrics"
echo "   Connect REST:       http://localhost:8083/connectors"