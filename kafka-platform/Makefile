# ═════════════════════════════════════════════════════════════
# Kafka Platform — Makefile
# ═════════════════════════════════════════════════════════════

SHELL := /bin/bash
.DEFAULT_GOAL := help

# ── Compose detection (v2 preferred, v1 fallback) ────────────
COMPOSE := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

# ── Paths ────────────────────────────────────────────────────
CONNECTOR_FILE   ?= cdc/connectors/postgres-cdc.json
CONNECTOR_NAME   ?= postgres-cdc-prod
CONNECT_URL      ?= http://localhost:8083
SCHEMA_REGISTRY  ?= http://localhost:8081
KAFKA_TOPIC      ?= cdc.user_events

# ── Go ───────────────────────────────────────────────────────
BINARY      := bin/producer
PKG         := ./cmd/producer

# ── Colors ───────────────────────────────────────────────────
CYAN  := \033[36m
GREEN := \033[32m
YELLOW:= \033[33m
RESET := \033[0m

# ═════════════════════════════════════════════════════════════
# Help
# ═════════════════════════════════════════════════════════════
.PHONY: help
help: ## Show this help
	@echo ""
	@echo "$(CYAN)Kafka Platform — Available Targets$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-22s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ═════════════════════════════════════════════════════════════
# Go — build / test / run
# ═════════════════════════════════════════════════════════════
.PHONY: build
build: ## Build the Go producer binary
	@mkdir -p bin
	go build -o $(BINARY) $(PKG)
	@echo "$(GREEN)✔ Built $(BINARY)$(RESET)"

.PHONY: test
test: ## Run Go unit tests
	go test ./... -count=1

.PHONY: test-race
test-race: ## Run Go tests with race detector
	go test -race ./... -count=1

.PHONY: test-cover
test-cover: ## Run Go tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

.PHONY: run
run: ## Run producer locally (uses .env)
	go run $(PKG)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out

# ═════════════════════════════════════════════════════════════
# Docker Compose — full stack
# ═════════════════════════════════════════════════════════════
.PHONY: up
up: ## Start the full stack (healthchecks gate startup)
	$(COMPOSE) up -d
	@echo "$(GREEN)✔ Stack is up$(RESET)"

.PHONY: up-build
up-build: ## Rebuild images and start the full stack
	$(COMPOSE) up -d --build
	@echo "$(GREEN)✔ Stack is up (rebuilt)$(RESET)"

.PHONY: down
down: ## Stop the stack (keep volumes)
	$(COMPOSE) down

.PHONY: down-v
down-v: ## Stop the stack and wipe volumes (fresh start)
	$(COMPOSE) down -v

.PHONY: restart
restart: down up ## Restart the stack

.PHONY: ps
ps: ## Show running services
	$(COMPOSE) ps

.PHONY: logs
logs: ## Tail logs from all services
	$(COMPOSE) logs -f --tail=100

.PHONY: logs-connect
logs-connect: ## Tail Kafka Connect logs
	$(COMPOSE) logs -f --tail=200 kafka-connect

.PHONY: logs-producer
logs-producer: ## Tail producer logs
	$(COMPOSE) logs -f --tail=200 producer

.PHONY: logs-deployer
logs-deployer: ## Show the one-shot CDC deployer output
	$(COMPOSE) logs cdc-deployer

# ═════════════════════════════════════════════════════════════
# Infrastructure — start subsets
# ═════════════════════════════════════════════════════════════
.PHONY: infra
infra: ## Start zookeeper, kafka, schema-registry
	$(COMPOSE) up -d zookeeper kafka schema-registry

.PHONY: cdc-up
cdc-up: ## Start postgres + kafka-connect
	$(COMPOSE) up -d postgres kafka-connect
	@echo "Waiting for Connect REST..."
	@until curl -sf $(CONNECT_URL)/connector-plugins >/dev/null 2>&1; do sleep 2; done
	@echo "$(GREEN)✔ Connect is ready$(RESET)"

# ═════════════════════════════════════════════════════════════
# CDC — connector lifecycle
# ═════════════════════════════════════════════════════════════
.PHONY: cdc-deploy
cdc-deploy: ## Deploy the Debezium connector (idempotent)
	CONNECT_URL=$(CONNECT_URL) ./cdc/scripts/deploy-connector.sh $(CONNECTOR_FILE)

.PHONY: cdc-status
cdc-status: ## Show connector + task health
	@curl -s $(CONNECT_URL)/connectors/$(CONNECTOR_NAME)/status

.PHONY: cdc-restart
cdc-restart: ## Gracefully restart the connector
	CONNECT_URL=$(CONNECT_URL) ./cdc/scripts/restart-connector.sh $(CONNECTOR_NAME)

.PHONY: cdc-pause
cdc-pause: ## Pause the connector
	@curl -s -X PUT $(CONNECT_URL)/connectors/$(CONNECTOR_NAME)/pause | jq

.PHONY: cdc-resume
cdc-resume: ## Resume the connector
	@curl -s -X PUT $(CONNECT_URL)/connectors/$(CONNECTOR_NAME)/resume | jq

.PHONY: cdc-delete
cdc-delete: ## Delete the connector
	@curl -s -X DELETE $(CONNECT_URL)/connectors/$(CONNECTOR_NAME) | jq

.PHONY: cdc-plugins
cdc-plugins: ## List installed Connect plugins
	@curl -s $(CONNECT_URL)/connector-plugins | jq -r '.[].class' | grep -i debezium

.PHONY: cdc-logs
cdc-logs: logs-connect ## Alias for logs-connect

# ═════════════════════════════════════════════════════════════
# Kafka — inspect topics, consume
# ═════════════════════════════════════════════════════════════
.PHONY: topics
topics: ## List all Kafka topics
	@docker exec -it kafka kafka-topics \
		--bootstrap-server kafka:29092 --list

.PHONY: topic-create
topic-create: ## Create a topic (TOPIC=name PARTITIONS=3)
	@docker exec -it kafka kafka-topics \
		--bootstrap-server kafka:29092 \
		--create --if-not-exists \
		--topic $(or $(TOPIC),user-events) \
		--partitions $(or $(PARTITIONS),3) \
		--replication-factor 1

.PHONY: topic-describe
topic-describe: ## Describe a topic (TOPIC=name)
	@docker exec -it kafka kafka-topics \
		--bootstrap-server kafka:29092 \
		--describe --topic $(or $(TOPIC),cdc.user_events)

.PHONY: consume
consume: ## Consume Avro CDC topic (TOPIC=cdc.user_events)
	@echo "$(CYAN)Consuming $(KAFKA_TOPIC)... (Ctrl+C to stop)$(RESET)"
	@docker exec -it kafka-connect kafka-avro-console-consumer \
		--bootstrap-server kafka:29092 \
		--topic $(KAFKA_TOPIC) \
		--property schema.registry.url=$(SCHEMA_REGISTRY) \
		--from-beginning \
		--consumer-property log4j.logger.io.confluent=WARN

.PHONY: consume-json
consume-json: ## Consume any topic as raw JSON (TOPIC=user-events)
	@echo "$(CYAN)Consuming $(or $(TOPIC),user-events) as JSON... (Ctrl+C to stop)$(RESET)"
	@docker exec -it kafka kafka-console-consumer \
		--bootstrap-server kafka:29092 \
		--topic $(or $(TOPIC),user-events) \
		--from-beginning

.PHONY: consume-head
consume-head: ## Consume first N messages (N=5)
	@docker exec -it kafka-connect kafka-avro-console-consumer \
		--bootstrap-server kafka:29092 \
		--topic $(KAFKA_TOPIC) \
		--property schema.registry.url=$(SCHEMA_REGISTRY) \
		--from-beginning \
		--max-messages $(or $(N),5)

# ═════════════════════════════════════════════════════════════
# Schema Registry
# ═════════════════════════════════════════════════════════════
.PHONY: schemas
schemas: ## List all registered schema subjects
	@curl -s $(SCHEMA_REGISTRY)/subjects | jq

.PHONY: schema-get
schema-get: ## Show latest schema (SUBJECT=cdc.user_events-value)
	@curl -s $(SCHEMA_REGISTRY)/subjects/$(or $(SUBJECT),user-events-value)/versions/latest | jq

# ═════════════════════════════════════════════════════════════
# Database — inspect / seed
# ═════════════════════════════════════════════════════════════
.PHONY: psql
psql: ## Open a psql session
	@docker exec -it postgres psql -U appuser -d production

.PHONY: db-insert
db-insert: ## Insert a test row into user_events (triggers CDC)
	@docker exec -it postgres psql -U appuser -d production \
		-c "INSERT INTO user_events (user_id, event_type) VALUES ('user-$$(date +%s)', 'LOGIN') RETURNING id, user_id, event_type;"

.PHONY: db-tables
db-tables: ## List tables in the production DB
	@docker exec -it postgres psql -U appuser -d production -c "\dt"

.PHONY: db-slots
db-slots: ## Show replication slots
	@docker exec -it postgres psql -U appuser -d production \
		-c "SELECT slot_name, plugin, active, restart_lsn FROM pg_replication_slots;"

# ═════════════════════════════════════════════════════════════
# End-to-end verification
# ═════════════════════════════════════════════════════════════
.PHONY: verify
verify: ## End-to-end pipeline check
	@echo "$(CYAN)▶ Checking services...$(RESET)"
	@$(COMPOSE) ps
	@echo ""
	@echo "$(CYAN)▶ Kafka topics:$(RESET)"
	@docker exec kafka kafka-topics --bootstrap-server kafka:29092 --list || true
	@echo ""
	@echo "$(CYAN)▶ Connector status:$(RESET)"
	@curl -s $(CONNECT_URL)/connectors/$(CONNECTOR_NAME)/status | jq '.connector.state, .tasks[].state' || true
	@echo ""
	@echo "$(CYAN)▶ Registered schemas:$(RESET)"
	@curl -s $(SCHEMA_REGISTRY)/subjects | jq || true
	@echo ""
	@echo "$(CYAN)▶ Replication slots:$(RESET)"
	@docker exec postgres psql -U appuser -d production \
		-c "SELECT slot_name, active FROM pg_replication_slots;" || true

.PHONY: e2e
e2e: ## Full end-to-end test (insert → consume)
	@echo "$(CYAN)▶ Inserting test row...$(RESET)"
	@$(MAKE) db-insert
	@echo ""
	@echo "$(CYAN)▶ Waiting 5s for CDC to propagate...$(RESET)"
	@sleep 5
	@echo ""
	@echo "$(CYAN)▶ Reading from $(KAFKA_TOPIC)...$(RESET)"
	@$(MAKE) consume-head N=3

# ═════════════════════════════════════════════════════════════
# Bootstrap — one-shot full pipeline
# ═════════════════════════════════════════════════════════════
.PHONY: bootstrap
bootstrap: ## Bring up the entire stack + deploy connector + verify
	@./scripts/bootstrap-all.sh

# ═════════════════════════════════════════════════════════════
# UIs — open in browser (macOS/Linux; Windows uses `start`)
# ═════════════════════════════════════════════════════════════
.PHONY: ui
ui: ## Print all UI URLs
	@echo "$(CYAN)UIs:$(RESET)"
	@echo "  Kafka UI:           http://localhost:8080"
	@echo "  Schema Registry UI: http://localhost:8000"
	@echo "  Kafka Connect UI:   http://localhost:8001"
	@echo "  Connect REST API:   http://localhost:8083/connectors"
	@echo "  Producer metrics:   http://localhost:9090/metrics"
	@echo "  Producer health:    http://localhost:9090/healthz"

# ═════════════════════════════════════════════════════════════
# Phases (placeholders for future work)
# ═════════════════════════════════════════════════════════════
.PHONY: phase3-up
phase3-up: ## (Phase 3) Start ksqlDB — placeholder
	@echo "$(YELLOW)Phase 3 (ksqlDB) not yet implemented$(RESET)"

.PHONY: phase4-up
phase4-up: ## (Phase 4) Start ClickHouse — placeholder
	@echo "$(YELLOW)Phase 4 (ClickHouse) not yet implemented$(RESET)"