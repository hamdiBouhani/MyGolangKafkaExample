.PHONY: help build test run compose-up compose-down cdc-up cdc-deploy cdc-status cdc-restart bootstrap

help:
	@echo "Targets:"
	@echo "  build        - build the Go producer binary"
	@echo "  test         - run Go tests"
	@echo "  run          - run producer locally"
	@echo "  compose-up   - start all infrastructure"
	@echo "  compose-down - stop everything"
	@echo "  cdc-up       - start Postgres + Kafka Connect"
	@echo "  cdc-deploy   - deploy the Debezium Postgres connector"
	@echo "  cdc-status   - check connector health"
	@echo "  cdc-restart  - restart the connector"
	@echo "  bootstrap    - bring the whole pipeline up end-to-end"

build:
	go build -o bin/producer ./cmd/producer

test:
	go test ./...

run:
	go run ./cmd/producer

compose-up:
	docker compose up -d zookeeper kafka schema-registry

compose-down:
	docker compose down -v

cdc-up:
	docker compose up -d postgres kafka-connect
	@echo "Waiting for Connect REST API..."
	@until curl -sf http://localhost:8083/ >/dev/null; do sleep 2; done
	@echo "Connect is up."

cdc-deploy:
	./cdc/scripts/deploy-connector.sh cdc/connectors/postgres-cdc.json

cdc-status:
	./cdc/scripts/check-status.sh

cdc-restart:
	./cdc/scripts/restart-connector.sh postgres-cdc-prod

bootstrap:
	./scripts/bootstrap-all.sh