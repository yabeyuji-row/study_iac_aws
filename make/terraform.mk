TFLINT ?= $(HOME)/.asdf/installs/golang/1.26.4/bin/tflint
TERRAFORM_AWS_ACCESS_KEY_ID ?= test
TERRAFORM_AWS_SECRET_ACCESS_KEY ?= test
TERRAFORM_DIR ?= infrastructure/app
TERRAFORM_VAR_FILE ?= envs/dev.tfvars
TERRAFORM_TFVARS ?= $(TERRAFORM_DIR)/$(TERRAFORM_VAR_FILE)
tfvar = $(strip $(shell awk -F= -v key="$(1)" '$$1 ~ "^[[:space:]]*" key "[[:space:]]*$$" { value=$$2; sub(/#.*/, "", value); sub(/^[[:space:]]+/, "", value); sub(/[[:space:]]+$$/, "", value); sub(/^"/, "", value); sub(/"$$/, "", value); print value; exit }' "$(TERRAFORM_TFVARS)" 2>/dev/null))
TERRAFORM_DESTROY_PLAN ?= destroy.tfplan
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

.PHONY: terraform-fmt
terraform-fmt:
	terraform -chdir=infrastructure/app fmt -recursive

.PHONY: terraform-init
terraform-init:
	terraform -chdir=infrastructure/app init -backend=false

.PHONY: terraform-validate
terraform-validate:
	terraform -chdir=infrastructure/app validate

.PHONY: terraform-plan
terraform-plan:
	AWS_ACCESS_KEY_ID="$(TERRAFORM_AWS_ACCESS_KEY_ID)" \
	AWS_SECRET_ACCESS_KEY="$(TERRAFORM_AWS_SECRET_ACCESS_KEY)" \
	AWS_EC2_METADATA_DISABLED=true \
	terraform -chdir="$(TERRAFORM_DIR)" plan -input=false -refresh=false -var-file="$(TERRAFORM_VAR_FILE)"

.PHONY: terraform-destroy-plan
terraform-destroy-plan:
	bash -c 'set -euo pipefail; \
		if [ -f .env.aws ]; then source .env.aws; fi; \
		terraform -chdir="$(TERRAFORM_DIR)" plan -destroy -var-file="$(TERRAFORM_VAR_FILE)" -out="$(TERRAFORM_DESTROY_PLAN)"'

.PHONY: terraform-destroy-apply
terraform-destroy-apply:
	bash -c 'set -euo pipefail; \
		if [ -f .env.aws ]; then source .env.aws; fi; \
		terraform -chdir="$(TERRAFORM_DIR)" apply "$(TERRAFORM_DESTROY_PLAN)"'

.PHONY: terraform-tflint-init
terraform-tflint-init:
	$(TFLINT) --chdir=infrastructure/app --init

.PHONY: terraform-tflint
terraform-tflint:
	$(TFLINT) --chdir=infrastructure/app

.PHONY: terraform-security
terraform-security:
	docker run --rm -v "$(CURDIR):/repo" bridgecrew/checkov -d /repo/infrastructure/app --framework terraform

.PHONY: terraform-graph
terraform-graph:
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(TERRAFORM_GRAPH_OUTPUT))"
	bash -c 'set -o pipefail; output="$(TERRAFORM_GRAPH_OUTPUT)"; tmp="$${output}.tmp"; trap "rm -f \"$$tmp\"" EXIT; terraform -chdir="$(TERRAFORM_DIR)" graph | dot -T$(TERRAFORM_GRAPH_FORMAT) > "$$tmp" && test -s "$$tmp" && mv "$$tmp" "$$output"'

.PHONY: terraform-rover
terraform-rover:
	@docker image inspect "$(ROVER_IMAGE)" >/dev/null 2>&1 || docker pull "$(ROVER_IMAGE)"
	bash -c 'set -euo pipefail; \
		workdir="$(TERRAFORM_DIR)/.rover-tmp"; \
		mkdir -p "$$workdir"; \
		trap "rm -rf \"$$workdir\"" EXIT; \
		AWS_ACCESS_KEY_ID="$(TERRAFORM_AWS_ACCESS_KEY_ID)" \
		AWS_SECRET_ACCESS_KEY="$(TERRAFORM_AWS_SECRET_ACCESS_KEY)" \
		AWS_EC2_METADATA_DISABLED=true \
		terraform -chdir="$(TERRAFORM_DIR)" plan -no-color -input=false -refresh=false -var-file="$(TERRAFORM_VAR_FILE)" -out=.rover-tmp/rover.tfplan > "$$workdir/terraform-plan.log" || { cat "$$workdir/terraform-plan.log"; exit 1; }; \
		terraform -chdir="$(TERRAFORM_DIR)" show -json .rover-tmp/rover.tfplan > "$$workdir/rover-plan.json"; \
		echo "Rover URL: http://localhost:$(ROVER_PORT)"; \
		docker run --rm -p "$(ROVER_PORT):9000" \
			-v "$(CURDIR)/$$workdir:/plan:ro" \
			-v "$(CURDIR)/$(TERRAFORM_DIR):/src:ro" \
			"$(ROVER_IMAGE)" -planJSONPath /plan/rover-plan.json -workingDir /src -ipPort 0.0.0.0:9000'

.PHONY: terraform-inframap-hcl
terraform-inframap-hcl:
	@command -v "$(INFRAMAP)" >/dev/null || { echo "inframap is required. See https://github.com/cycloidio/inframap"; exit 1; }
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(INFRAMAP_HCL_OUTPUT))"
	bash -c 'set -o pipefail; output="$(INFRAMAP_HCL_OUTPUT)"; tmp="$${output}.tmp"; trap "rm -f \"$$tmp\"" EXIT; $(INFRAMAP) generate --hcl "$(TERRAFORM_DIR)" | dot -T$(INFRAMAP_FORMAT) > "$$tmp" && test -s "$$tmp" && mv "$$tmp" "$$output"'

.PHONY: terraform-inframap-state
terraform-inframap-state:
	@command -v "$(INFRAMAP)" >/dev/null || { echo "inframap is required. See https://github.com/cycloidio/inframap"; exit 1; }
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(INFRAMAP_STATE_OUTPUT))"
	bash -c 'set -o pipefail; output="$(INFRAMAP_STATE_OUTPUT)"; tmp="$${output}.tmp"; trap "rm -f \"$$tmp\"" EXIT; terraform -chdir="$(TERRAFORM_DIR)" state pull | $(INFRAMAP) generate --tfstate | dot -T$(INFRAMAP_FORMAT) > "$$tmp" && test -s "$$tmp" && mv "$$tmp" "$$output"'

.PHONY: terraform-inframap-html
terraform-inframap-html:
	@command -v "$(INFRAMAP)" >/dev/null || { echo "inframap is required. See https://github.com/cycloidio/inframap"; exit 1; }
	@command -v dot >/dev/null || { echo "Graphviz dot is required."; exit 1; }
	@mkdir -p "$(dir $(INFRAMAP_HTML_OUTPUT))"
	bash -c 'set -o pipefail; svg="$(INFRAMAP_HTML_SVG_OUTPUT)"; desc="$(INFRAMAP_DESCRIPTION_OUTPUT)"; svg_tmp="$${svg}.tmp"; desc_tmp="$${desc}.tmp"; trap "rm -f \"$$svg_tmp\" \"$$desc_tmp\"" EXIT; $(INFRAMAP) generate --hcl --description-file "$$desc_tmp" "$(TERRAFORM_DIR)" >/dev/null && $(INFRAMAP) generate --hcl --raw "$(TERRAFORM_DIR)" | go run ./cmd/inframapdot | dot -Tsvg > "$$svg_tmp" && test -s "$$svg_tmp" && test -s "$$desc_tmp" && mv "$$svg_tmp" "$$svg" && mv "$$desc_tmp" "$$desc"'
	bash -c 'schema_tmp="$(INFRAMAP_HTML_OUTPUT).schema.tmp"; trap "rm -f \"$$schema_tmp\"" EXIT; terraform -chdir="$(TERRAFORM_DIR)" providers schema -json > "$$schema_tmp" && go run ./cmd/inframaphtml -svg "$(INFRAMAP_HTML_SVG_OUTPUT)" -description "$(INFRAMAP_DESCRIPTION_OUTPUT)" -vars "$(TERRAFORM_TFVARS)" -terraform-dir "$(TERRAFORM_DIR)" -schema "$$schema_tmp" -out "$(INFRAMAP_HTML_OUTPUT)" -title "Terraform InfraMap"'
