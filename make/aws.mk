PROJECT_NAME ?= $(or $(call tfvar,project_name),study-aws)
ENVIRONMENT ?= $(or $(call tfvar,environment),dev)
NAME_PREFIX ?= $(PROJECT_NAME)-$(ENVIRONMENT)
AWS_REGION ?= $(or $(call tfvar,aws_region),ap-northeast-1)
ECR_REPOSITORY_SHORT_NAME ?= $(or $(call tfvar,ecr_repository_name),todo-api)
ECR_REPOSITORY_NAME ?= $(NAME_PREFIX)-$(ECR_REPOSITORY_SHORT_NAME)
ECS_CLUSTER_NAME ?= $(NAME_PREFIX)-cluster
ECS_SERVICE_NAME ?= $(NAME_PREFIX)-api-service
RDS_INSTANCE_IDENTIFIER ?= $(NAME_PREFIX)-postgres

.PHONY: aws-ecr-empty
aws-ecr-empty:
	bash -c 'set -euo pipefail; \
		if [ -f .env.aws ]; then source .env.aws; fi; \
		image_ids="$$(aws ecr list-images \
			--repository-name "$(ECR_REPOSITORY_NAME)" \
			--region "$(AWS_REGION)" \
			--query "imageIds" \
			--output json)"; \
		if [ "$$image_ids" = "[]" ]; then \
			echo "ECR repository is already empty: $(ECR_REPOSITORY_NAME)"; \
			exit 0; \
		fi; \
		aws ecr batch-delete-image \
			--repository-name "$(ECR_REPOSITORY_NAME)" \
			--region "$(AWS_REGION)" \
			--image-ids "$$image_ids"'

.PHONY: aws-ecs-stop
aws-ecs-stop:
	bash -c 'set -euo pipefail; \
		if [ -f .env.aws ]; then source .env.aws; fi; \
		aws ecs update-service \
			--cluster "$(ECS_CLUSTER_NAME)" \
			--service "$(ECS_SERVICE_NAME)" \
			--desired-count 0 \
			--region "$(AWS_REGION)"'

.PHONY: aws-rds-stop
aws-rds-stop:
	bash -c 'set -euo pipefail; \
		if [ -f .env.aws ]; then source .env.aws; fi; \
		aws rds stop-db-instance \
			--db-instance-identifier "$(RDS_INSTANCE_IDENTIFIER)" \
			--region "$(AWS_REGION)"'

.PHONY: aws-cost-stop
aws-cost-stop: aws-ecs-stop aws-rds-stop
