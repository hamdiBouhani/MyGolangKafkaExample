#!/usr/bin/env bash
# Waits for Kafka Connect to be ready, deploys the configured connector,
# then verifies it reaches RUNNING state. Idempotent: safe to re-run.
set -euo pipefail

CONNECT_URL="${CONNECT_URL:-http://kafka-connect:8083}"
CONNECTOR_FILE="${CONNECTOR_FILE:-/connectors/postgres-cdc.json}"
WAIT_TIMEOUT="${WAIT_TIMEOUT:-180}"

# ── 1. Wait for Connect REST (worker ready, plugins loaded) ──
echo "▶ Waiting for Kafka Connect at ${CONNECT_URL}..."
elapsed=0
until curl -sf "${CONNECT_URL}/connector-plugins" >/dev/null 2>&1; do
    if [ "$elapsed" -ge "$WAIT_TIMEOUT" ]; then
        echo "✖ Timed out after ${WAIT_TIMEOUT}s waiting for Connect" >&2
        exit 1
    fi
    sleep 2
    elapsed=$((elapsed + 2))
done
echo "✔ Connect is ready"

# ── 2. Deploy the connector (create or update) ───────────────
CONNECTOR_NAME=$(jq -r '.name' "$CONNECTOR_FILE")
echo "▶ Deploying connector '${CONNECTOR_NAME}'..."

if curl -s -o /dev/null -w "%{http_code}" \
        "${CONNECT_URL}/connectors/${CONNECTOR_NAME}" | grep -q "200"; then
    echo "  → Connector exists, updating config"
    curl -s -X PUT "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/config" \
        -H "Content-Type: application/json" \
        -d "$(jq '.config' "$CONNECTOR_FILE")" >/dev/null
else
    echo "  → Creating new connector"
    curl -s -X POST "${CONNECT_URL}/connectors" \
        -H "Content-Type: application/json" \
        -d @"$CONNECTOR_FILE" >/dev/null
fi

# ── 3. Wait for RUNNING state ────────────────────────────────
echo "▶ Waiting for connector to reach RUNNING..."
elapsed=0
while true; do
    STATE=$(curl -s "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/status" \
            | jq -r '.connector.state // "UNKNOWN"')
    TASKS_RUNNING=$(curl -s "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/status" \
            | jq '[.tasks[]? | select(.state == "RUNNING")] | length')
    TASKS_TOTAL=$(curl -s "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/status" \
            | jq '.tasks | length')

    echo "  state=${STATE} tasks=${TASKS_RUNNING}/${TASKS_TOTAL}"

    if [ "$STATE" = "RUNNING" ] && [ "$TASKS_RUNNING" = "$TASKS_TOTAL" ]; then
        echo "✔ Connector is healthy"
        exit 0
    fi

    if [ "$elapsed" -ge "$WAIT_TIMEOUT" ]; then
        echo "✖ Connector did not become healthy within ${WAIT_TIMEOUT}s" >&2
        curl -s "${CONNECT_URL}/connectors/${CONNECTOR_NAME}/status" | jq >&2
        exit 1
    fi

    sleep 3
    elapsed=$((elapsed + 3))
done