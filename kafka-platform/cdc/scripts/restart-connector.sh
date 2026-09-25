#!/usr/bin/env bash
# Gracefully restart a connector and its tasks.
set -euo pipefail

CONNECT_URL="${CONNECT_URL:-http://localhost:8083}"
CONNECTOR_NAME="${1:?Usage: $0 <connector-name>}"

echo "Restarting connector: ${CONNECTOR_NAME}"

curl -s -X PUT "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/pause" | jq
sleep 2
curl -s -X POST "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/restart?includeTasks=true&onlyFailed=false" | jq

echo "Restart initiated. Monitoring status..."
sleep 10
curl -s "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/status" | jq