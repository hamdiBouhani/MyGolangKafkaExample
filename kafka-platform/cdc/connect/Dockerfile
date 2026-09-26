FROM confluentinc/cp-kafka-connect:7.5.0

USER root

# ── Debezium source connectors ─────────────────────────────
RUN confluent-hub install --no-prompt debezium/debezium-connector-postgresql:latest\
 && confluent-hub install --no-prompt debezium/debezium-connector-mysql:latest\
 && confluent-hub install --no-prompt debezium/debezium-connector-sqlserver:latest

# ── Confluent Avro converter (REQUIRED for io.confluent.connect.avro.AvroConverter) ──
RUN confluent-hub install --no-prompt confluentinc/kafka-connect-avro-converter:latest

# ── Confluent SMTs (REQUIRED for org.apache.kafka.connect.transforms.RegexRouter) ──
RUN confluent-hub install --no-prompt confluentinc/connect-transforms:latest

# ── Worker configuration ───────────────────────────────────
COPY connect-distributed.properties /etc/kafka/connect-distributed.properties

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -f http://localhost:8083/connector-plugins || exit 1

ENTRYPOINT ["/etc/confluent/docker/run"]