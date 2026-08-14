# TODO API Learning Project

This repository is a phased learning project for building a production-like TODO
REST API with Go, Echo v4, PostgreSQL, Docker, AWS, Terraform, GitHub Actions,
observability, SRE, security, backup and recovery, and cost management.

Current phase: Phase 5, SRE and operations.

## Architecture

```text
Client
  |
Application Load Balancer
  |
Amazon ECS Fargate
  |
Go REST API
  |
Amazon RDS for PostgreSQL
```

## Local Startup

Start PostgreSQL, MiniStack, run migrations, and then start the API:

```bash
make local-up
```

The API runs in the foreground. Stop it with `Ctrl+C`, then stop PostgreSQL and
MiniStack:

```bash
make local-down
```

Start only PostgreSQL:

```bash
make postgres-up
```

Start only MiniStack:

```bash
make ministack-up
```

Start only the API:

```bash
make api-up
```

`make api-up` expects PostgreSQL to be running and migrations to be applied.

Alternatively, start PostgreSQL and the API with Docker Compose:

```bash
make compose-up
```

Run migrations:

```bash
make migrate-up
```

Run the API with Air hot reload:

```bash
make dev
```

The API listens on `http://localhost:8080` by default. If the API is already
running before migrations are applied, apply migrations before sending TODO
requests.

If Air is not installed locally:

```bash
make install-air
```

Open the local TODO UI in your browser:

```bash
make open
```

## API

Implemented endpoints:

```text
POST   /v1/todos
GET    /v1/todos
GET    /v1/todos/{todo_id}
PUT    /v1/todos/{todo_id}
DELETE /v1/todos/{todo_id}
GET    /healthz
GET    /readyz
GET    /version
```

The API implements the `/v1/todos` CRUD endpoints and operational endpoints for
health, readiness, and build version checks.

The detailed API design is in [docs/design/api-design.md](docs/design/api-design.md).

## Curl Examples

Create a TODO:

```bash
curl -sS -X POST http://localhost:8080/v1/todos \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Terraformを学ぶ",
    "description": "VPCとECSをTerraformで構築する",
    "status": "pending",
    "due_date": "2026-08-31T23:59:59+09:00"
  }'
```

List TODOs:

```bash
curl -sS 'http://localhost:8080/v1/todos?limit=20&sort=created_at_desc'
```

Get a TODO:

```bash
curl -sS http://localhost:8080/v1/todos/{todo_id}
```

Update a TODO:

```bash
curl -sS -X PUT http://localhost:8080/v1/todos/{todo_id} \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Terraformを学ぶ",
    "description": "ECSまで完了させる",
    "status": "in_progress",
    "due_date": "2026-08-31T23:59:59+09:00",
    "version": 1
  }'
```

Delete a TODO:

```bash
curl -i -X DELETE http://localhost:8080/v1/todos/{todo_id}
```

## Tests And Lint

```bash
make test
make test-race
make lint
make vuln
make build
```

Repository tests use real PostgreSQL when `TEST_DATABASE_URL` is set:

```bash
TEST_DATABASE_URL='postgres://todo:todo_password@localhost:5432/todo_api?sslmode=disable' \
  go test ./internal/todo -run TestPostgresRepositoryCRUD
```

Install `govulncheck` first if it is not available locally:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Migrations

SQL-file-based migrations are explicit commands and do not run unconditionally
on API startup.

```bash
make migrate-up
make migrate-down
```

## Docker

Build the API image:

```bash
make docker-build
```

Run local PostgreSQL and API:

```bash
make compose-up
```

Run each service separately:

```bash
make postgres-up
make ministack-up
make api-up
```

Run PostgreSQL, MiniStack, migrations, and the API together:

```bash
make local-up
```

The Compose API service uses Air and reloads when Go templates, Go files, HTML,
CSS, or JavaScript files change.

Stop local services:

```bash
make compose-down
```

## MiniStack

MiniStack provides a local AWS-compatible endpoint for learning and partial
integration checks.

```bash
make ministack-up
make ministack-health
make ministack-down
```

See [MiniStack setup](docs/ministack-setup.md) for the Terraform provider
override example and the boundary between MiniStack checks and real AWS checks.

## Terraform

Application Terraform lives under `infrastructure/app/`. The local validation
flow uses `envs/dev.tfvars` and dummy AWS credentials for MiniStack-compatible
planning.

```bash
cp infrastructure/app/envs/dev.tfvars.example infrastructure/app/envs/dev.tfvars
make terraform-init
make terraform-fmt
make terraform-validate
make terraform-tflint
make terraform-security
make terraform-plan
```

See [Terraform file overview](docs/terraform-file-overview.md) and
[MiniStack Terraform app checks](docs/ministack-terraform-app.md).

Do not run `terraform apply` or `terraform destroy` without explicit approval.

## AWS Deployment

GitHub Actions use OIDC, not static access keys.
Image push can run from `main`, `v*` tags, or manual dispatch.
ECS service update is gated by manual dispatch with `deploy = true` and the
GitHub `dev` environment approval gate.

AWS operations entry points:

- [RDS backup and restore](docs/operations/rds-backup-restore.md)
- [Rollback procedures](docs/operations/rollback.md)
- [Terraform drift handling](docs/operations/terraform-drift.md)
- [AWS cost review checklist](docs/operations/cost-review-checklist.md)
- [AWS cleanup checklist](docs/operations/cleanup-checklist.md)
- [Postmortem template](docs/templates/postmortem-template.md)

## CI/CD

GitHub Actions workflows live under `.github/workflows/`.
Pull request checks include workflow lint, Go tests, race tests, linting,
vulnerability checks, Docker build, Terraform fmt/validate, TFLint, Checkov,
and a Terraform plan artifact.

Deploy workflow image push can run from `main`, `v*` tags, or manual dispatch.
ECS service update is gated by manual dispatch with `deploy = true` and the
GitHub `dev` environment approval gate.

## Monitoring And SRE

Initial observability uses JSON request logs, CloudWatch Logs metric filters,
CloudWatch alarms, and a CloudWatch dashboard. See
[Observability and SLO design](docs/design/observability-slo.md) and the
[runbooks](docs/runbooks/).

For final recovery drills and review notes, see
[Week 12 final review](docs/operations/week12-final-review.md).

## Cost Notice

Do not create AWS resources before reviewing the target resources and rough
costs. ALB, NAT Gateway, RDS, VPC endpoints, CloudWatch Logs, Secrets Manager,
and data transfer can create ongoing charges.

## Design Documents

- [System overview](docs/design/system-overview.md)
- [API design](docs/design/api-design.md)
- [Database design](docs/design/database-design.md)
- [AWS architecture](docs/design/aws-architecture.md)
- [Learning roadmap](docs/learning-roadmap.md)
- [ADRs](docs/adr/)
