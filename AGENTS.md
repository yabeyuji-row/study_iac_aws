# AGENTS.md

## Project Overview

This repository is a phased learning project for building a TODO REST API with
Go, PostgreSQL, Docker, AWS, Terraform, GitHub Actions, observability, SRE,
security, backup and recovery, and cost management.

The target production-like architecture is:

```text
Client -> Application Load Balancer -> ECS Fargate -> Go REST API -> RDS PostgreSQL
```

## Technology

- Go REST API
- Echo v4
- PostgreSQL
- Docker and Docker Compose
- Amazon ECR, ECS Fargate, ALB, RDS, Secrets Manager, CloudWatch
- Terraform
- GitHub Actions with OIDC

## Phase Rule

Work must be done by phase. Stop at the end of each phase and report results.

- Phase 0: design only
- Phase 1: local TODO API
- Phase 2: local operational quality
- Phase 3: Terraform
- Phase 4: CI/CD
- Phase 5: SRE

## Commands

Commands will be added as phases introduce implementation files.

- Run: `make run`
- Test: `make test`
- Race test: `make test-race`
- Lint: `make lint`
- Vulnerability check: `make vuln`
- Build: `make build`
- Migrate: `make migrate-up` / `make migrate-down`

## Go Rules

- Use Echo v4 for HTTP routing and middleware.
- Keep service and repository layers independent from Echo.
- Use meaningful names except common names such as `ctx`, `err`, and `id`.
- Prefer named return values.
- Add context to returned errors.
- Propagate `context.Context`.
- Set HTTP and DB timeouts.
- Do not use global mutable state.
- Do not log secrets, passwords, or connection strings.
- Keep dependencies minimal.
- Keep business data in PostgreSQL, not process memory.

## Security Rules

- Do not commit AWS credentials, passwords, tokens, private keys, `.env`
  values, or real `terraform.tfvars` values.
- Use Secrets Manager for AWS DB credentials.
- Run containers as non-root.
- Validate inputs in the application and with database constraints.
- Do not expose RDS to the internet.

## AWS And Terraform Rules

- Do not run `terraform apply` without explicit user approval.
- Do not run `terraform destroy` without explicit user approval.
- Do not create billable AWS resources before reporting target resources and
  rough costs.
- Do not automatically deploy to production AWS.
- Keep bootstrap state and application infrastructure state separate.

## Prohibited Changes

Do not introduce Redis, ElastiCache, SSE, WebSocket, GraphQL, Kafka, Kinesis,
EventBridge, SNS, SQS, DynamoDB, Lambda, Kubernetes, EKS, CQRS, Event Sourcing,
unneeded caching, or in-memory persistent business state.

## Documentation Rules

Update design docs and ADRs when architecture, technology choices, deployment,
operations, security, or cost assumptions change.

## Files Not To Commit

- `.env`
- `.env.*` with real secrets
- `terraform.tfvars`
- `*.tfstate`
- `*.tfstate.*`
- AWS credential files
- Private keys and tokens
