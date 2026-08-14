# Postmortem Template

## Summary

- Incident title:
- Date:
- Owner:
- Severity:
- Status:

## Impact

- User impact:
- Affected endpoints:
- Affected data:
- Start time:
- Detection time:
- Mitigation time:
- Recovery time:

## Timeline

| Time | Event |
| --- | --- |
|  |  |

## Detection

- Triggered alarm:
- Dashboard or log evidence:
- Who noticed:

## Root Cause

- Direct cause:
- Contributing factors:
- Why existing tests, alarms, or process did not catch it earlier:

## Mitigation

- Immediate action:
- Rollback, restore, or config change:
- Residual risk:

## Recovery Validation

- `/healthz`:
- `/readyz`:
- TODO CRUD:
- ECS service stable:
- ALB target healthy:
- RDS metrics:

## Lessons

- What went well:
- What was confusing:
- What was missing:

## Action Items

| Action | Owner | Due Date | Status |
| --- | --- | --- | --- |
|  |  |  |  |

## Follow-up Links

- PR:
- Workflow run:
- Terraform plan:
- CloudWatch dashboard:
- Runbook:
