---
doc_id: THR-S7-HARNESS-NLSPEC
title: Testing Harness Natural Language Specification
status: draft
role: harness-nlspec
---

# Testing Harness Natural Language Specification

## Document role

This S7 draft specifies the repository testing harness contract recovered from
S0 through S6 evidence plus the 2026-05-09 maintainer decisions. It governs
harness mechanics only. Product behavior remains owned by Core 00 through Core
04 and `docs/domain.md` remains vocabulary reference for domain-facing text.

Final normative language must keep the evidence split from the S7 audit:
`selected_runtime_observed`, `source_observed`, `source_limit`, and
`maintainer_decision_required`.

## Scope

The harness contract covers Make entrypoints, scheduler behavior, service-backed
test setup, local-dev verification setup, retained artifacts, generated harness
inputs, cleanup bounds, fixture/golden/snapshot maintenance, machine-readable
schemas, and provider-neutral CI entrypoints.

The contract does not define product behavior, public product APIs, deployment
conformance, provider-specific CI behavior, visual snapshot refresh platform
bounds, or environment-variable precedence.

## Authority

| Rule | Evidence |
|---|---|
| Core 00 through Core 04 own product behavior. Harness documents may define validation mechanics only. | `AUTH-0001`, `MD-S7-0017` |
| `make` is the sole canonical harness command surface. | `AUTH-0002`, `PRES-0001`, `MD-S7-0001` |
| Direct package scripts are developer conveniences unless they re-enter Make-owned wrappers. | `AUTH-0003`, `PRES-0011`, `MD-S7-0001` |
| Generated task, schedule, Go, and TypeScript artifacts are downstream execution inputs. They do not own behavior. | `AUTH-0010`, `PRES-0002`, `MD-S7-0003` |
| Phase 7 and Phase 8 registry entries are planned future work, not active current coverage. | `AMB-0004`, `SL-0005`, `MD-S7-0002` |

## Command surface

The supported harness command surface is the Make public surface documented by
`make help`, `make help-all`, `make task-guide`, task-surface manifests, and the
generated Make include produced from upstream manifests.

`make ci` is the provider-neutral CI enforcement entrypoint. It composes the
canonical repo task surface and must not depend on provider-specific workflow
files that are absent from the repository checkout.

`make release-check` remains a release verification gate. Release readiness is
not required before S7 implementation proceeds, and any license/SBOM/build
readiness failure must be reported separately from the demoted stale extended
harness smoke.

## Local development verification

Local-dev Compose, `make db-up`, `make services-up`, `make db-reset`, and
`make dev` are part of the local verification contract. This contract describes
developer verification setup, not deployment conformance.

The local bootstrap server uses `configs/dev/config.toml` through
`CARTULARY_CONFIG_FILE`. Local-dev persistence and reset gaps remain documented
as local behavior and must not be generalized into production data guarantees.

## Scheduler model

Scheduler resource lanes and claims are logical scheduling constraints. They
bound scheduler work selection and ordering; they are not physical capacity
guarantees for host CPUs, Docker, Postgres, MinIO, browser processes, or network
ports.

Duration baselines are planning weights only. They can inform work ordering and
drift checks, but they do not prove throughput or service capacity.

## Environment and platform

The harness documents only source-observed environment variables, defaults, and
validation behavior. Cross-layer precedence remains unresolved and must stay
`TODO: precedence_unknown` until a future owner decision.

WSL/Linux is the primary observed environment. The harness should remain
portable across Linux environments where practical, but this draft does not
claim a complete support matrix for non-Linux hosts, missing tools, browsers, or
provider CI environments.

## Retained artifacts

Durable claims based on retained artifacts require explicit run identity:
`RESULTS_DIR`, `RUN_ID`, command or target, platform/tool profile, exit status,
and artifact paths. Ambient newest-artifact fallback is allowed only for human
investigation and cannot support final normative claims.

Stable machine JSON outputs are contract candidates only where source declares a
schema ID or where the acceptance matrix names the field contract. Tool-defined
reports, shell log contents, Playwright reports, traces, videos, screenshots,
and CI provider annotations remain tool-defined, diagnostic, or
`schema_unknown`.

## Reset route

The test runtime reset route is harness-owned. Its implementation must live
outside production application ownership, remain disabled by default, and be
registered only by explicit test-route wiring such as
`CARTULARY_ENABLE_TEST_ROUTES=1`.

The reset response schema remains `cartulary.test.runtime_reset.v1`. The route
is a test hook and must not be documented as a product API. Partial reset
failure semantics remain source-limited unless selected evidence is added.

## Cleanup and destructive safety

Cleanup claims use these tiers:

| Tier | Meaning |
|---|---|
| `observed_successful_cleanup` | Selected runtime evidence shows a specific cleanup path completed successfully. |
| `observed_cleanup_scheduling` | Source or selected evidence shows cleanup was scheduled, but completion is not proven. |
| `delayed_after_state_evidence` | A later observation saw the expected after-state, without proving synchronous cleanup completion. |
| `best_effort_cleanup` | Source attempts cleanup, but failures, interrupts, or detached completion are not guaranteed. |
| `source_limited_cleanup` | Source exists or routing is known, but selected runtime evidence or owner decision is missing. |

Stale janitors may delete DBs, buckets, or containers only when all deletion
proof gates pass: generated resource name, harness metadata or lease evidence,
conservative age or completed-summary check, and scope-limited resource type.

Parent-death cleanup, detached reaper completion as a hard guarantee, and active
DB connection cleanup remain source-limited. External Go caches under
`/tmp/cartulary-go-*` remain outside default `clean` and `distclean` scope.

## Fixtures, goldens, and snapshots

Fixture and golden updates require controlled owner-reviewed workflow: explicit
intent, targeted file list, source evidence or test reason, targeted
verification command, and review note explaining why the change is expected.

Visual snapshots are validation-only for now. `make browser-e2e-visual` may
validate committed baselines, but no snapshot update command, OS, browser, or
version is normatively blessed until an owner supplies exact bounds.

## CI and stale smoke

CI remains provider-neutral. Repository documentation must not invent `.github`
workflow behavior, provider annotations, upload paths, or dashboard semantics.

The stale `run-harness-smoke-extended` failure is demoted from blocking phase
advancement. `check-harness-smoke` and fast smoke remain the developer-gate smoke
surface if they pass. `run-harness-smoke-extended` may remain as an explicit
diagnostic target unless intentionally retired later.

## Acceptance hooks

Acceptance criteria live in
`docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`. Every
final `must` or `must not` in this NLSpec needs one acceptance criterion or a
source-limit/owner-decision blocker.

