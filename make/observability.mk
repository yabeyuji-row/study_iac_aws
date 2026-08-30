OBSERVABILITY_COMPOSE = docker compose -f compose.yaml -f compose.observability.yaml
FAULT_DELAY ?= 2s

.PHONY: observability-up
observability-up: postgres-up migrate-up
	$(OBSERVABILITY_COMPOSE) up -d api prometheus grafana tempo otel-collector

.PHONY: observability-down
observability-down:
	$(OBSERVABILITY_COMPOSE) down

.PHONY: observability-logs
observability-logs:
	$(OBSERVABILITY_COMPOSE) logs -f api prometheus grafana tempo otel-collector

.PHONY: loadtest
loadtest:
	$(OBSERVABILITY_COMPOSE) --profile load run --rm k6

.PHONY: fault-5xx
fault-5xx:
	curl -fsS -o /dev/null -w 'status=%{http_code}\n' \
		-H 'X-Fault-Injection: enabled' http://localhost:8080/debug/fault/5xx || test $$? -eq 22

.PHONY: fault-delay
fault-delay:
	curl -fsS -H 'X-Fault-Injection: enabled' \
		"http://localhost:8080/debug/fault/delay?duration=$(FAULT_DELAY)"

.PHONY: fault-db-down
fault-db-down:
	docker compose stop postgres

.PHONY: fault-db-recover
fault-db-recover:
	$(MAKE) postgres-up
