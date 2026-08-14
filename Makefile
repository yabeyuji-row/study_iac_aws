.PHONY: run dev install-air open test test-race lint vuln build docker-build \
	api-up local-up local-down postgres-up postgres-down compose-up \
	compose-down ministack-up ministack-down ministack-logs ministack-health \
	ministack-list-secrets \
	workflow-lint \
	migrate-up migrate-down terraform-fmt terraform-init terraform-validate \
	terraform-plan terraform-tflint-init terraform-tflint terraform-security \
	terraform-inframap-hcl terraform-inframap-state terraform-inframap-html \
	terraform-graph terraform-rover

DATABASE_URL ?= postgres://todo:todo_password@localhost:5432/todo_api?sslmode=disable
APP_URL ?= http://localhost:8080/
AIR_VERSION ?= v1.67.3
GO_BIN_DIR ?= $(shell go env GOBIN)
ifeq ($(GO_BIN_DIR),)
GO_BIN_DIR := $(shell go env GOPATH)/bin
endif
AIR ?= $(GO_BIN_DIR)/air
ACTIONLINT_VERSION ?= v1.7.7
MINISTACK_ENDPOINT ?= http://localhost:4566
AWS_REGION ?= ap-northeast-1
TFLINT ?= /home/yuji/.asdf/installs/golang/1.26.4/bin/tflint
TERRAFORM_AWS_ACCESS_KEY_ID ?= test
TERRAFORM_AWS_SECRET_ACCESS_KEY ?= test
TERRAFORM_DIR ?= infrastructure/app
TERRAFORM_TFVARS ?= $(TERRAFORM_DIR)/envs/dev.tfvars
ROVER_IMAGE ?= im2nguyen/rover:latest
ROVER_PORT ?= 9000
INFRAMAP ?= inframap
INFRAMAP_FORMAT ?= svg
INFRAMAP_WEEK ?= week10
INFRAMAP_OUTPUT_DIR ?= docs/inframap
ifeq ($(origin INFRAMAP_TIMESTAMP), undefined)
INFRAMAP_TIMESTAMP := $(shell date +%y%m%d-%H%M%S)
endif
INFRAMAP_OUTPUT_STEM ?= $(INFRAMAP_OUTPUT_DIR)/$(INFRAMAP_TIMESTAMP)-$(INFRAMAP_WEEK)
INFRAMAP_HCL_OUTPUT ?= $(INFRAMAP_OUTPUT_STEM).svg
INFRAMAP_STATE_OUTPUT ?= $(INFRAMAP_OUTPUT_STEM)-state.svg
INFRAMAP_DESCRIPTION_OUTPUT ?= $(INFRAMAP_OUTPUT_STEM).json
INFRAMAP_HTML_SVG_OUTPUT ?= $(INFRAMAP_OUTPUT_STEM).svg
INFRAMAP_HTML_OUTPUT ?= $(INFRAMAP_OUTPUT_STEM).html
# TERRAFORM_GRAPH_FORMAT ?= svg
TERRAFORM_GRAPH_FORMAT ?= png
TERRAFORM_GRAPH_OUTPUT ?= $(INFRAMAP_OUTPUT_STEM)-terraform-graph.$(TERRAFORM_GRAPH_FORMAT)

run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

api-up: run

dev:
	DATABASE_URL="$(DATABASE_URL)" HTTP_ADDR=":8080" "$(AIR)" -c .air.toml

install-air:
	go install github.com/air-verse/air@$(AIR_VERSION)

open:
	xdg-open "$(APP_URL)"

test:
	go test ./...

test-race:
	@if [ -z "$$(command -v gcc)" ]; then \
		echo "gcc is required for go test -race. Install a C compiler, then rerun make test-race."; \
		exit 1; \
	fi
	CGO_ENABLED=1 go test -race ./...

lint:
	go vet ./...

vuln:
	govulncheck ./...

build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate

docker-build:
	docker build -t todo-api:local .

workflow-lint:
	go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)

local-up:
	$(MAKE) postgres-up
	$(MAKE) ministack-up
	$(MAKE) migrate-up
	$(MAKE) api-up

local-down:
	$(MAKE) postgres-down
	$(MAKE) ministack-down

postgres-up:
	docker compose up -d postgres
	@until docker compose exec -T postgres pg_isready -U todo -d todo_api; do \
		echo "waiting for postgres..."; \
		sleep 1; \
	done

postgres-down:
	docker compose stop postgres

compose-up:
# 	docker compose up -d
	docker compose up

compose-down:
	docker compose down

ministack-up:
	docker compose -f compose.ministack.yaml up -d

ministack-down:
	docker compose -f compose.ministack.yaml down

ministack-logs:
	docker compose -f compose.ministack.yaml logs -f ministack

ministack-health:
	curl -fsS "$(MINISTACK_ENDPOINT)/_ministack/health"

ministack-list-secrets:
	AWS_ACCESS_KEY_ID="$(TERRAFORM_AWS_ACCESS_KEY_ID)" \
	AWS_SECRET_ACCESS_KEY="$(TERRAFORM_AWS_SECRET_ACCESS_KEY)" \
	AWS_DEFAULT_REGION="$(AWS_REGION)" \
	aws --endpoint-url="$(MINISTACK_ENDPOINT)" secretsmanager list-secrets

migrate-up:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-down:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down

terraform-fmt:
	terraform -chdir=infrastructure/app fmt -recursive

terraform-init:
	terraform -chdir=infrastructure/app init -backend=false

terraform-validate:
	terraform -chdir=infrastructure/app validate

terraform-plan:
	AWS_ACCESS_KEY_ID="$(TERRAFORM_AWS_ACCESS_KEY_ID)" \
	AWS_SECRET_ACCESS_KEY="$(TERRAFORM_AWS_SECRET_ACCESS_KEY)" \
	AWS_EC2_METADATA_DISABLED=true \
	terraform -chdir=infrastructure/app plan -input=false -refresh=false -var-file=envs/dev.tfvars

terraform-tflint-init:
	$(TFLINT) --chdir=infrastructure/app --init

terraform-tflint:
	$(TFLINT) --chdir=infrastructure/app

terraform-security:
	docker run --rm -v "$(CURDIR):/repo" bridgecrew/checkov -d /repo/infrastructure/app --framework terraform

terraform-graph:
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(TERRAFORM_GRAPH_OUTPUT))"
	bash -c 'set -o pipefail; output="$(TERRAFORM_GRAPH_OUTPUT)"; tmp="$${output}.tmp"; trap "rm -f \"$$tmp\"" EXIT; terraform -chdir="$(TERRAFORM_DIR)" graph | dot -T$(TERRAFORM_GRAPH_FORMAT) > "$$tmp" && test -s "$$tmp" && mv "$$tmp" "$$output"'

terraform-rover:
	@docker image inspect "$(ROVER_IMAGE)" >/dev/null 2>&1 || docker pull "$(ROVER_IMAGE)"
	bash -c 'set -euo pipefail; \
		workdir="$(TERRAFORM_DIR)/.rover-tmp"; \
		mkdir -p "$$workdir"; \
		trap "rm -rf \"$$workdir\"" EXIT; \
		AWS_ACCESS_KEY_ID="$(TERRAFORM_AWS_ACCESS_KEY_ID)" \
		AWS_SECRET_ACCESS_KEY="$(TERRAFORM_AWS_SECRET_ACCESS_KEY)" \
		AWS_EC2_METADATA_DISABLED=true \
		terraform -chdir="$(TERRAFORM_DIR)" plan -no-color -input=false -refresh=false -var-file=envs/dev.tfvars -out=.rover-tmp/rover.tfplan > "$$workdir/terraform-plan.log" || { cat "$$workdir/terraform-plan.log"; exit 1; }; \
		terraform -chdir="$(TERRAFORM_DIR)" show -json .rover-tmp/rover.tfplan > "$$workdir/rover-plan.json"; \
		echo "Rover URL: http://localhost:$(ROVER_PORT)"; \
		docker run --rm -p "$(ROVER_PORT):9000" \
			-v "$(CURDIR)/$$workdir:/plan:ro" \
			-v "$(CURDIR)/$(TERRAFORM_DIR):/src:ro" \
			"$(ROVER_IMAGE)" -planJSONPath /plan/rover-plan.json -workingDir /src -ipPort 0.0.0.0:9000'

terraform-inframap-hcl:
	@command -v "$(INFRAMAP)" >/dev/null || { echo "inframap is required. See https://github.com/cycloidio/inframap"; exit 1; }
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(INFRAMAP_HCL_OUTPUT))"
	bash -c 'set -o pipefail; output="$(INFRAMAP_HCL_OUTPUT)"; tmp="$${output}.tmp"; trap "rm -f \"$$tmp\"" EXIT; $(INFRAMAP) generate --hcl "$(TERRAFORM_DIR)" | dot -T$(INFRAMAP_FORMAT) > "$$tmp" && test -s "$$tmp" && mv "$$tmp" "$$output"'

terraform-inframap-state:
	@command -v "$(INFRAMAP)" >/dev/null || { echo "inframap is required. See https://github.com/cycloidio/inframap"; exit 1; }
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(INFRAMAP_STATE_OUTPUT))"
	bash -c 'set -o pipefail; output="$(INFRAMAP_STATE_OUTPUT)"; tmp="$${output}.tmp"; trap "rm -f \"$$tmp\"" EXIT; terraform -chdir="$(TERRAFORM_DIR)" state pull | $(INFRAMAP) generate --tfstate | dot -T$(INFRAMAP_FORMAT) > "$$tmp" && test -s "$$tmp" && mv "$$tmp" "$$output"'

terraform-inframap-html:
	@command -v "$(INFRAMAP)" >/dev/null || { echo "inframap is required. See https://github.com/cycloidio/inframap"; exit 1; }
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(INFRAMAP_HTML_OUTPUT))"
	bash -c 'set -o pipefail; svg="$(INFRAMAP_HTML_SVG_OUTPUT)"; desc="$(INFRAMAP_DESCRIPTION_OUTPUT)"; svg_tmp="$${svg}.tmp"; desc_tmp="$${desc}.tmp"; trap "rm -f \"$$svg_tmp\" \"$$desc_tmp\"" EXIT; $(INFRAMAP) generate --hcl --description-file "$$desc_tmp" "$(TERRAFORM_DIR)" >/dev/null && $(INFRAMAP) generate --hcl --raw "$(TERRAFORM_DIR)" | go run ./cmd/inframapdot | dot -Tsvg > "$$svg_tmp" && test -s "$$svg_tmp" && test -s "$$desc_tmp" && mv "$$svg_tmp" "$$svg" && mv "$$desc_tmp" "$$desc"'
	bash -c 'schema_tmp="$(INFRAMAP_HTML_OUTPUT).schema.tmp"; trap "rm -f \"$$schema_tmp\"" EXIT; terraform -chdir="$(TERRAFORM_DIR)" providers schema -json > "$$schema_tmp" && go run ./cmd/inframaphtml -svg "$(INFRAMAP_HTML_SVG_OUTPUT)" -description "$(INFRAMAP_DESCRIPTION_OUTPUT)" -vars "$(TERRAFORM_TFVARS)" -terraform-dir "$(TERRAFORM_DIR)" -schema "$$schema_tmp" -out "$(INFRAMAP_HTML_OUTPUT)" -title "Terraform InfraMap"'
