#!/usr/bin/env bash
# Exit non-zero if any connector or task is not RUNNING.
set -euo pipefail

CONNECT_URL="${CONNECT_URL:-http://localhost:8083}"
FAILED=0

for connector in $(curl -s "${CONNECT_URL}/connectors" | jq -r '.[]'); do
    STATUS=$(curl -s "${CONNECT_URL}/connectors/${connector}/status")
    STATE=$(echo "$STATUS" | jq -r '.connector.state')
    RUNNING_TASKS=$(echo "$STATUS" | jq '[.tasks[] | select(.state == "RUNNING")] | length')
    TOTAL_TASKS=$(echo "$STATUS" | jq '.tasks | length')

    echo "Connector: ${connector}"
    echo "  State: ${STATE}"
    echo "  Tasks: ${RUNNING_TASKS}/${TOTAL_TASKS} running"

    if [ "$STATE" != "RUNNING" ] || [ "$RUNNING_TASKS" != "$TOTAL_TASKS" ]; then
        echo "  WARNING: UNHEALTHY"
        FAILED=1
    fi
done

exit $FAILED