#!/usr/bin/env bash
# Idempotent connector deployment — safe to run multiple times.
set -euo pipefail

CONNECT_URL="${CONNECT_URL:-http://localhost:8083}"
CONNECTOR_FILE="${1:?Usage: $0 <connector.json>}"
CONNECTOR_NAME=$(jq -r '.name' "$CONNECTOR_FILE")

echo "Checking if connector '${CONNECTOR_NAME}' exists..."

if curl -s -o /dev/null -w "%{http_code}" "${CONNECT_URL}/connectors/${CONNECTOR_NAME}" | grep -q "200"; then
    echo "Connector exists. Updating configuration..."
    curl -s -X PUT "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/config" \
        -H "Content-Type: application/json" \
        -d "$(jq '.config' "$CONNECTOR_FILE")" | jq
else
    echo "Creating new connector..."
    curl -s -X POST "${CONNECT_URL}/connectors" \
        -H "Content-Type: application/json" \
        -d @"$CONNECTOR_FILE" | jq
fi

echo "Verifying status..."
sleep 5
curl -s "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/status" | jq