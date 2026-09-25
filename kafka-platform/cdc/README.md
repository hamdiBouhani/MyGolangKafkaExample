# CDC Pipeline — Kafka Connect + Debezium

Captures row-level changes from PostgreSQL/MySQL and publishes them to Kafka.

## Layout

- `connect/`    — custom Connect image + worker config
- `connectors/` — connector JSON definitions
- `scripts/`    — deploy / status / restart helpers
- `init/`       — source DB bootstrap SQL

## Quick start

```bash
# From repo root
make cdc-up
make cdc-deploy
make cdc-status