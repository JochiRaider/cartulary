---
doc_id: THR-S5-FAILURE-CLASS-TAXONOMY
title: Testing Harness Recovery Failure Class Taxonomy
status: active
role: failure-class-taxonomy
---

# Testing Harness Recovery Failure Class Taxonomy

## Taxonomy

| Failure class | Meaning in S4 follow-up recovery | Retryability | Owner | Evidence status | Notes |
|---|---|---|---|---|---|
| `usage_error` | Invalid CLI args, missing required flags, or malformed resource override. | no | harness CLI owner | `observed` | Usually exits `2` in source-observed CLIs. |
| `configuration_error` | Missing binary, invalid env value, missing Playwright selector env, invalid result path, or invalid scratch root. | no until config changes | harness/tool owner | `observed/source_limit` | Missing-tool/platform cases route to S8 support profile. |
| `preflight_error` | Docker/testcontainers or suite preflight fails before service startup. | conditional | S6 runtime, S8 platform | `observed/source_limit` | Live Docker behavior was not run. |
| `service_start_error` | Postgres, MinIO, Compose, backend, or Vite start fails. | conditional | service harness owner | `observed/source_limit` | Static retry rules exist for some testcontainers paths only. |
| `service_readiness_timeout` | Ready loop exhausts before service is usable. | conditional | S6 runtime | `observed/source_limit` | Cannot close without controlled runtime evidence. |
| `fixture_error` | DB/bucket/template/reset/worker-admin fixture setup or cleanup fails. | conditional/unknown | harness fixture owner | `observed/source_limit` | Destructive janitor authority is S8. |
| `resource_conflict` | Port, scheduler lane, DB/bucket name, lock, or process conflict. | conditional | S6 hazard owner | `observed/source_limit` | Scheduler lanes do not prove concrete capacity. |
| `test_assertion_failure` | Product or harness test assertion fails after setup. | no, except rerun policy outside this pass | test owner | `observed/source_limit` | Must remain separate from harness operational failures. |
| `timeout` | Command, readiness, lock, cleanup, or test timeout. | conditional/unknown | S6 timing owner | `source_limit` | Timeout/interrupt cleanup was not executed. |
| `cancelled_or_interrupted` | INT/TERM, parent death, child shutdown, or abrupt runner abort. | unknown | S6 timing owner | `source_limit` | Parent-death behavior remains unobserved. |
| `cleanup_error` | Cleanup returns non-zero, leak check fails, reaper scheduling fails, or stale cleanup fails. | conditional | S6 runtime, S8 janitor authority | `observed/source_limit` | Completion strength must be classified by S6. |
| `unsupported_platform` | Required or assumed tool/platform behavior unavailable. | no until environment changes | S8 platform owner | `source_limit` | Covers Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, shell, localhost, Node/pnpm, browser runtime. |
| `authority_required` | Behavior cannot be made normative without owner decision. | n/a | maintainer/harness owner | `maintainer_decision_required` | Reset route, package scripts, local-dev services, env contracts, janitor bounds, and external caches. |

