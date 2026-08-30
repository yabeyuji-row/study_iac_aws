DATABASE_URL ?= postgres://todo:todo_password@localhost:5432/todo_api?sslmode=disable
APP_URL ?= http://localhost:8080/
AIR_VERSION ?= v1.67.3
GO_BIN_DIR ?= $(shell go env GOBIN)
ifeq ($(GO_BIN_DIR),)
GO_BIN_DIR := $(shell go env GOPATH)/bin
endif
AIR ?= $(GO_BIN_DIR)/air
ACTIONLINT_VERSION ?= v1.7.7

.PHONY: run
run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

.PHONY: api-up
api-up: run

.PHONY: dev
dev:
	DATABASE_URL="$(DATABASE_URL)" HTTP_ADDR=":8080" "$(AIR)" -c .air.toml

.PHONY: install-air
install-air:
	go install github.com/air-verse/air@$(AIR_VERSION)

.PHONY: open
open:
	xdg-open "$(APP_URL)"

.PHONY: test
test:
	go test ./...

.PHONY: test-race
test-race:
	@if [ -z "$$(command -v gcc)" ]; then \
		echo "gcc is required for go test -race. Install a C compiler, then rerun make test-race."; \
		exit 1; \
	fi
	CGO_ENABLED=1 go test -race ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: vuln
vuln:
	govulncheck ./...

.PHONY: build
build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate

.PHONY: docker-build
docker-build:
	docker build -t todo-api:local .

.PHONY: workflow-lint
workflow-lint:
	go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)
