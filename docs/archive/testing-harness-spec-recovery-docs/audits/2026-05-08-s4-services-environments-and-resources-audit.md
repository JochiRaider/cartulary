---
doc_id: THR-S4-AUDIT-2026-05-08
title: S4 Services Environments and Resources Audit
status: complete
role: recovery-audit
---

# S4 Services, Environments, and Resources Audit

## Audit verdict

`pass_with_followups`

Sprint 4 is complete enough to support Sprint 5 lifecycle, observable
interface, and failure recovery. The required S4 outputs exist, S4 is marked
complete with blocker `none`, service and resource rows are table-complete,
environment assumptions are tied to entrypoints, and unsupported runtime,
platform, timeout, cleanup, and precedence behavior is preserved in source-limit
and ambiguity rows instead of being promoted to runtime proof.

No S5 work was started. No harness implementation, service startup behavior,
environment handling, cleanup script, generated artifact, fixture, lockfile,
runtime service, or S5 output was modified by this audit.

## Audit scope

| Scope item | Result | Evidence status | Evidence |
|---|---|---|---|
| Audit write surface | Audit artifact created under `docs/testing-harness-spec-recovery-docs/audits/`. | `observed` | This file. |
| S4 sprint status | S4 is marked `complete` with blocker `none`; S5 remains `not_started`. | `observed` | `03-sprint-plan.md` S4 and S5 sections. |
| Required S4 outputs | Required S4 output files exist. | `runtime_observed` | Required-file presence check for S4 output paths. |
| S4 handoff | Handoff exists and records static-only discovery, outputs, source areas, and S5/S6/S8 constraints. | `observed` | `handoffs/2026-05-08-s4-services-environments-resources.md`. |
| Implementation edits | No implementation, service, cleanup, fixture, generated, lockfile, runtime, or S5 paths were modified by this audit. | `runtime_observed` | `git status --short --branch`; this file is the only audit-created path. |
| Audit command boundary | Only non-mutating inspection commands were run. | `runtime_observed` | Commands listed below. |

## Evidence reviewed

| Evidence area | Audit check | Result | Evidence status | Notes |
|---|---|---|---|---|
| Recovery controls | Checked recovery process, sprint plan, S4 handoff, S3 audit, source-limit log, and ambiguity register. | pass | `observed` | S4 follows the recovery-doc-only boundary and evidence-label model. |
| Output presence | Checked all S4 outputs required by the sprint plan. | pass | `runtime_observed` | `service-lifecycle-map.md`, `environment-contract-observations.md`, `resource-allocation-register.md`, hazard/register updates, and S4 handoff are present. |
| Service lifecycle rows | Counted `SVC-0001` through `SVC-0015` and checked table field counts. | pass | `observed/runtime_observed` | Every service row has 19 logical columns: provision, configure, start, readiness, use, reset, stop, cleanup, failure behavior, scope, evidence, and evidence status are represented. |
| Environment rows | Counted `ENV-0001` through `ENV-0024` and checked table field counts. | pass | `observed/runtime_observed` | Env rows tie defaults, accepted values, effects, precedence observations, evidence, and entrypoints together. |
| Resource rows | Counted `RES-0001` through `RES-0026` and checked table field counts. | pass | `observed/runtime_observed` | Scheduler lanes and concrete resources include allocation, collision, release, reuse, and parallel-safety fields. |
| S4 hazards | Counted `HAZ-S4-0001` through `HAZ-S4-0009` and checked table field counts. | pass | `observed/runtime_observed` | Hazards cite S4 service/resource rows and S3 source hazards where needed. |
| Ambiguities and source limits | Checked `AMB-0023` through `AMB-0028` and `SL-0012` through `SL-0015`. | pass | `observed` | Missing live readiness, reaper completion, platform behavior, env precedence, stale janitor bounds, and timeout/interrupt behavior are preserved. |
| Service and environment dependency identification | Compared S4 rows with S2 entrypoints, S3 external-state rows, and S4 inspected source areas. | pass | `observed/source_limit` | Covered testservices, Postgres, MinIO, Docker/testcontainers, Compose, browser backend/frontend, Playwright state, reset route, dev stack, scheduler lanes, caches, scratch paths, ports, process groups, DBs, buckets, and browser profiles. |
| Lifecycle phase coverage | Checked provision, configure, start, ready-check, use, reset, stop, cleanup, and failure columns across sampled service rows. | pass | `observed/source_limit` | Reset-route route-specific readiness and timeout are explicitly unknown instead of inferred. |
| Readiness, retry, and wait loops | Checked source-observed readiness timeouts, poll loops, retry attempts, and failure behavior for service-backed, browser, Playwright, and dev surfaces. | pass | `observed/source_limit` | Runtime readiness remains unexecuted under `SL-0012`; S4 does not claim live timing proof. |
| Resource allocation and isolation | Checked scheduler lanes, DB/bucket naming, locks, ports, process groups, Playwright state, runtime roots, containers, cache roots, reset state, and scratch dirs. | pass | `observed/source_limit` | Logical scheduler lanes are explicitly not concrete host or service capacity guarantees. |
| Platform, network, and secrets | Checked Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, localhost, Node/pnpm, browser runtime, and local credentials. | pass | `observed/source_limit` | Unsupported platform behavior is carried as `TODO: platform_unknown`; local test credentials are not treated as production secret policy. |
| S5 boundary | Checked S5 sprint status and searched for expected S5 output artifacts. | pass | `runtime_observed` | No S5 output path was found beyond the S4 service lifecycle map. |

## Commands run

All commands below were non-mutating inspection commands. The audit did not run
service-backed targets, browser E2E targets, Docker/testcontainers, Docker
Compose, reset routes, cleanup paths, formatters, generators, broad
verification gates, or S5 work.

| Command | Result | Evidence status | Notes |
|---|---|---|---|
| `git status --short --branch` | Exit 0; showed existing S4 recovery-doc changes and no implementation changes from this audit. | `runtime_observed` | Worktree already contained uncommitted S4 docs before audit creation. |
| `git rev-parse HEAD` | Exit 0; printed `947e6254d6fbc5154ad4691e0485f6e22e3153e1`. | `runtime_observed` | Matches S4 metadata. |
| `date -Is` | Exit 0; printed `2026-05-08T22:04:53-04:00`. | `runtime_observed` | Audit timestamp evidence. |
| `uname -a` | Exit 0; printed the WSL2 Linux platform recorded in S4 metadata. | `runtime_observed` | Platform evidence. |
| Required-file presence loop for S4 outputs | Exit 0 for all required S4 paths. | `runtime_observed` | Confirmed S4 output presence. |
| `sed` reads of `03-sprint-plan.md` and S4 handoff | Exit 0. | `observed` | Checked S4 status, blockers, issues, outputs, and S5 handoff constraints. |
| `rg -o ... | sort -u` for `SVC-*`, `ENV-*`, `RES-*`, `HAZ-S4-*`, `AMB-0023..0028`, and `SL-0012..0015` | Exit 0. | `runtime_observed` | Confirmed expected ID ranges. |
| `awk -F'|' ...` over S4 tables | Exit 0; all sampled S4 rows had consistent field counts. | `runtime_observed` | Checked table shape for service, env, resource, and hazard rows. |
| Targeted `rg` over S4 docs for readiness, cleanup, timeout, interrupt, source limits, platform, and precedence markers | Exit 0. | `observed/source_limit` | Verified unsupported claims are explicitly marked. |
| Targeted `rg` over `tools/testservices`, `pgtest`, `s3test`, and `testcontainersx` | Exit 0. | `observed/source_limit` | Checked service startup, readiness timeouts, retries, stale janitor, and reaper evidence. |
| Targeted `rg` over browser/reset/process lifecycle scripts | Exit 0. | `observed/source_limit` | Checked port allocation, `ss`, wait loops, process groups, reset route enablement, and cleanup paths. |
| Targeted `rg` over Compose/dev services and dev stack files | Exit 0. | `observed/source_limit` | Checked fixed ports, health checks, wait timeouts, and `db-reset` object-store boundary. |
| Targeted `rg` over scheduler registry/manifests and scheduler resource scripts | Exit 0. | `observed` | Checked logical resource lanes, env overrides, resource claims, validation, and forwarding. |
| Targeted `rg` over Playwright config and E2E harness files | Exit 0. | `observed/source_limit` | Checked worker defaults, global setup/teardown, locks, manifests, and cleanup markers. |
| Targeted `rg` over `harness-scratch.sh` | Exit 0. | `observed` | Checked temp-path validation and outside-repo scratch rule. |
| `find docs/testing-harness-spec-recovery-docs ...` for S5-like outputs | Exit 0; found only `service-lifecycle-map.md`. | `runtime_observed` | Confirmed no S5 output was started. |

`make agent-finalize` was not run because this audit's non-scope prohibited
mutating maintenance commands and generated-artifact refresh paths.

## Findings

| Finding ID | Finding | Severity | Evidence status | Evidence reference | Disposition |
|---|---|---|---|---|---|
| AUD-S4-0001 | Required S4 outputs are present, S4 is marked complete, and S5 remains not started. | none | `observed/runtime_observed` | `03-sprint-plan.md`; S4 output file checks; S4 handoff. | Pass. |
| AUD-S4-0002 | `service-lifecycle-map.md` is row-complete for discovered S4 services and environments. | none | `observed/source_limit` | `SVC-0001` through `SVC-0015`; table field-count check. | Pass. |
| AUD-S4-0003 | Services include provision, configure, start, readiness, timeout, use, reset, stop, cleanup, failure behavior, scope, evidence, and source-limit notes where needed. | none | `observed/source_limit` | Sampled `SVC-0001`, `SVC-0002`, `SVC-0005`, `SVC-0009`, `SVC-0010`, `SVC-0011`, `SVC-0013`, `SVC-0014`, `SVC-0015`. | Pass. |
| AUD-S4-0004 | Readiness checks, retry behavior, wait loops, and timeouts are mapped without claiming live runtime proof. | none | `observed/source_limit` | `SVC-0001` through `SVC-0015`; `SL-0012`; `AMB-0023`. | Pass. |
| AUD-S4-0005 | Reset behavior is represented for DB fixtures, object namespaces, browser reset boundaries, standalone browser state, and the app test runtime reset route. | none | `observed/source_limit` | `SVC-0004`, `SVC-0006`, `SVC-0011`, `SVC-0012`, `SVC-0014`; `RES-0023`, `RES-0026`; `HAZ-S4-0007`. | Pass. |
| AUD-S4-0006 | Cleanup paths are documented while timeout, interrupt, detached reaper completion, active DB connections, stale janitors, and port release remain source-limited. | none | `observed/source_limit` | `SVC-0001`, `SVC-0006`, `SVC-0009`, `SVC-0011`, `SVC-0013`, `SVC-0015`; `SL-0014`; `AMB-0024`, `AMB-0027`; `HAZ-S4-0003`, `HAZ-S4-0004`. | Pass. |
| AUD-S4-0007 | `environment-contract-observations.md` ties env defaults and assumptions to entrypoints and preserves unresolved precedence. | none | `observed/source_limit` | `ENV-0001` through `ENV-0024`; `AMB-0012`; `AMB-0026`; `SL-0015`. | Pass. |
| AUD-S4-0008 | Network and secret assumptions are represented without broadening local test credentials into production secret policy. | none | `observed/source_limit` | `ENV-0010`, `ENV-0012`, `ENV-0014`, `ENV-0020`, `ENV-0024`; environment observations carried forward. | Pass. |
| AUD-S4-0009 | `resource-allocation-register.md` covers scheduler lanes and concrete resources with allocation, collision, release, reuse, and parallel-safety statements. | none | `observed/source_limit` | `RES-0001` through `RES-0026`; table field-count check. | Pass. |
| AUD-S4-0010 | Logical scheduler lanes are explicitly separated from concrete service and host capacity guarantees. | none | `observed/source_limit` | `RES-0001` through `RES-0010`; allocation gaps; `HAZ-S4-0001`. | Pass. |
| AUD-S4-0011 | Port, process, browser profile, DB, bucket, container, cache, temp-path, and reset-state allocation rules are represented. | none | `observed/source_limit` | `RES-0012` through `RES-0026`; `SVC-0004`, `SVC-0006`, `SVC-0009`, `SVC-0013`, `SVC-0014`. | Pass. |
| AUD-S4-0012 | Unsupported platform behavior is recorded rather than normalized into a support matrix. | none | `observed/source_limit` | `ENV-0024`; `RES-0017`, `RES-0018`, `RES-0019`; `AMB-0025`, `AMB-0028`; `SL-0013`. | Pass. |
| AUD-S4-0013 | S4 hazards are present and cover shared-resource conflicts, cleanup uncertainty, reset-route authority, platform assumptions, and package-script bypass. | none | `observed/source_limit` | `HAZ-S4-0001` through `HAZ-S4-0009`. | Pass. |
| AUD-S4-0014 | S4 converted high-risk S3 external-state hazards into concrete service and resource rows where applicable. | none | `observed/source_limit` | S3 `HAZ-S3-0007` through `HAZ-S3-0014`; S4 `SVC-*`, `RES-*`, and `HAZ-S4-*` rows. | Pass. |
| AUD-S4-0015 | Direct package-script behavior remains separate from Make-owned result-root and scheduler guarantees. | none | `observed` | `ENV-0004`, `ENV-0016`, `ENV-0022`; `HAZ-S4-0009`; `AMB-0020`. | Pass. |
| AUD-S4-0016 | S4 did not create S5 observable-interface, terminal-state, or failure-taxonomy artifacts. | none | `runtime_observed` | S5 status in `03-sprint-plan.md`; S5-output `find` check. | Pass. |
| AUD-S4-0017 | The audit itself stayed inside recovery documentation and did not modify harness behavior. | none | `runtime_observed` | `git status --short --branch`; this file. | Pass. |

## Blocking issues

No blocking S4 issues were found.

S5 may proceed from S4 outputs, especially:

- `service-lifecycle-map.md` rows `SVC-0001` through `SVC-0015`
- `environment-contract-observations.md` rows `ENV-0001` through `ENV-0024`
- `resource-allocation-register.md` rows `RES-0001` through `RES-0026`
- `shared-state-hazard-list.md` rows `HAZ-S4-0001` through `HAZ-S4-0009`
- source-limit rows `SL-0012` through `SL-0015`
- ambiguity rows `AMB-0023` through `AMB-0028`

S5 must preserve the S4 boundary: S4 service, env, and resource row IDs may
anchor interface and failure recovery, but they are not final stdout/stderr,
exit-code, structured-output, terminal-state, retryability, failure-bundle,
or runtime timing contracts.

## Follow-up issues

| Follow-up ID | Issue | Target sprint | Why non-blocking for S5 |
|---|---|---|---|
| AUD-S4-FU-0001 | Recover live service readiness for Docker/testcontainers, service-backed targets, browser E2E, Docker Compose, reset route calls, and actual host-specific wait behavior. | S6 or controlled runtime pass | `SL-0012` and `AMB-0023` preserve the gap; S5 can still map static interfaces with source-limit labels. |
| AUD-S4-FU-0002 | Classify cleanup guarantees for timeout, interrupt, parent death, detached reaper completion, active DB connections, stale fixture janitor execution, and port release. | S6 | `SL-0014`, `AMB-0024`, and `AMB-0027` preserve cleanup uncertainty. |
| AUD-S4-FU-0003 | Decide authority for the app test runtime reset route and its visibility/security boundary. | S8 or explicit owner decision | `AMB-0006`, `SVC-0014`, `RES-0026`, and `HAZ-S4-0007` preserve the boundary. |
| AUD-S4-FU-0004 | Decide whether direct package scripts are first-class harness contracts or developer conveniences. | S8 | `AMB-0020` and `HAZ-S4-0009` preserve Make-versus-package-script differences. |
| AUD-S4-FU-0005 | Define public environment-variable contracts and precedence across Make, schedulers, shell wrappers, package scripts, Go helpers, config files, and Playwright. | S8, with S5 source-limited interface notes | `AMB-0012`, `AMB-0026`, and `SL-0015` preserve the precedence gap. |
| AUD-S4-FU-0006 | Define supported platform and required-tool profile for Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, shell, localhost networking, Node/pnpm, and browser runtime. | S8/S6 | `ENV-0024`, `AMB-0025`, `AMB-0028`, and `SL-0013` preserve platform uncertainty. |
| AUD-S4-FU-0007 | Classify scheduler lanes versus concrete Docker, Postgres, MinIO, browser, host process, and port capacity. | S6 | `HAZ-S4-0001` and resource allocation gap notes prevent overclaiming capacity. |
| AUD-S4-FU-0008 | Decide local-dev service lifecycle authority, including Compose persistence, fixed ports, volumes, and object-store reset boundary. | S8/S6 | `SVC-0008`, `SVC-0012`, `SVC-0015`, `RES-0019`, and `HAZ-S4-0008` keep local-dev behavior separate. |
| AUD-S4-FU-0009 | Decide whether external Go cache roots are part of the harness cleanup contract. | S8 | `RES-0025`, S3 `HAZ-S3-0015`, and `AMB-0021` preserve the low-severity authority gap. |

## Source-limit and ambiguity updates

No mandatory register updates were found during the audit. Existing S4 rows
cover the unsupported assumptions surfaced by this audit:

- live readiness and runtime behavior: `SL-0012`, `AMB-0023`
- unsupported platform and missing-tool behavior: `SL-0013`, `AMB-0025`, `AMB-0028`
- timeout, interrupt, detached reaper, active-connection cleanup, and stale janitor behavior: `SL-0014`, `AMB-0024`, `AMB-0027`
- environment override precedence: `SL-0015`, `AMB-0012`, `AMB-0026`
- reset-route authority: `AMB-0006`, `HAZ-S4-0007`
- package-script authority: `AMB-0020`, `HAZ-S4-0009`

## Validation checklist

| Check | Result | Notes |
|---|---|---|
| S4 sprint status and handoff checked. | pass | S4 complete, blocker `none`; S5 not started. |
| Required S4 outputs present. | pass | All required S4 output paths exist. |
| `SVC-*` rows sampled and completeness-checked. | pass | `SVC-0001` through `SVC-0015`; consistent field count. |
| `ENV-*` rows sampled and precedence gaps checked. | pass | `ENV-0001` through `ENV-0024`; unresolved precedence recorded. |
| `RES-*` rows sampled and collision/release rules checked. | pass | `RES-0001` through `RES-0026`; allocation gaps carried forward. |
| `HAZ-S4-*`, `AMB-0023..0028`, and `SL-0012..0015` verified. | pass | All expected IDs present. |
| Source citations checked against repo files. | pass | Targeted source searches validated representative S4 claims. |
| Unsupported platform, network, and secret assumptions checked. | pass | Platform unknowns and local test credentials are explicitly represented. |
| Missing readiness, cleanup, shared-resource, and ambiguity gaps recorded. | pass | No new mandatory register row found. |
| Blocking versus follow-up issues classified. | pass | No S5 blockers; follow-ups listed above. |
| No prohibited mutation or S5 work performed. | pass | Audit added this recovery-doc file only. |
| Final verdict states whether S5 may proceed. | pass | Verdict is `pass_with_followups`; S5 may proceed with S4 source limits preserved. |

## Implementation-change audit

| Check | Result |
|---|---|
| Harness implementation files modified | `no` |
| Service startup behavior modified | `no` |
| Environment handling modified | `no` |
| Cleanup scripts modified | `no` |
| Test logic modified | `no` |
| Fixture contents modified | `no` |
| Golden or snapshot contents modified | `no` |
| Generated code modified | `no` |
| Generated manifests modified | `no` |
| Duration baselines modified | `no` |
| Lockfiles modified | `no` |
| Runtime services started | `no` |
| Reset routes called | `no` |
| Docker/testcontainers or Docker Compose run | `no` |
| S5 output files created | `no` |
| Only recovery docs changed by this audit | `yes` |

## Final audit note

S4 is ready for S5 with follow-ups. S5 should consume S4 service, environment,
and resource IDs as anchors, but it must preserve S4's source limits and avoid
turning static lifecycle evidence into final observable-interface, failure,
timing, retry, cleanup, or terminal-state contracts.
