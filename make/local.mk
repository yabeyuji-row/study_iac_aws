MINISTACK_ENDPOINT ?= http://localhost:4566

.PHONY: local-up
local-up:
	$(MAKE) postgres-up
	$(MAKE) ministack-up
	$(MAKE) migrate-up
	$(MAKE) api-up

.PHONY: local-down
local-down:
	$(MAKE) postgres-down
	$(MAKE) ministack-down

.PHONY: postgres-up
postgres-up:
	docker compose up -d postgres
	@until docker compose exec -T postgres pg_isready -U todo -d todo_api; do \
		echo "waiting for postgres..."; \
		sleep 1; \
	done

.PHONY: postgres-down
postgres-down:
	docker compose stop postgres

.PHONY: compose-up
compose-up:
# 	docker compose up -d
	docker compose up

.PHONY: compose-down
compose-down:
	docker compose down

.PHONY: ministack-up
ministack-up:
	docker compose -f compose.ministack.yaml up -d

.PHONY: ministack-down
ministack-down:
	docker compose -f compose.ministack.yaml down

.PHONY: ministack-logs
ministack-logs:
	docker compose -f compose.ministack.yaml logs -f ministack

.PHONY: ministack-health
ministack-health:
	curl -fsS "$(MINISTACK_ENDPOINT)/_ministack/health"

.PHONY: ministack-list-secrets
ministack-list-secrets:
	AWS_ACCESS_KEY_ID="$(TERRAFORM_AWS_ACCESS_KEY_ID)" \
	AWS_SECRET_ACCESS_KEY="$(TERRAFORM_AWS_SECRET_ACCESS_KEY)" \
	AWS_DEFAULT_REGION="$(AWS_REGION)" \
	aws --endpoint-url="$(MINISTACK_ENDPOINT)" secretsmanager list-secrets

.PHONY: migrate-up
migrate-up:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

.PHONY: migrate-down
migrate-down:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down
