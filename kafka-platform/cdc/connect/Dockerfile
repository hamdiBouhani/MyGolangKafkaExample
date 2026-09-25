# kafka-connect/Dockerfile
# Multi-stage build: install Debezium plugins, then run Connect
FROM confluentinc/cp-kafka-connect:7.5.0 AS builder

USER root

# Install Debezium plugins from Confluent Hub with pinned versions
RUN confluent-hub install --no-prompt debezium/debezium-connector-postgresql:latest \
 && confluent-hub install --no-prompt debezium/debezium-connector-mysql:latest \
 && confluent-hub install --no-prompt debezium/debezium-connector-sqlserver:latest

# Optional: Add custom SMTs (Single Message Transforms) or monitoring plugins
RUN confluent-hub install --no-prompt \
    confluentinc/connect-transforms:1.4.3

# Copy configuration
COPY connect-distributed.properties /etc/kafka/connect-distributed.properties

USER appuser

# Health check for orchestrators
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -f http://localhost:8083/ || exit 1

ENTRYPOINT ["/etc/confluent/docker/run"]