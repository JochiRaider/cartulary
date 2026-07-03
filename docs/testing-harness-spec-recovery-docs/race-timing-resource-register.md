---
doc_id: THR-S6-RACE-TIMING-RESOURCE
title: Testing Harness Recovery Race Timing Resource Register
status: active
role: race-timing-resource-register
---

# Testing Harness Recovery Race Timing Resource Register

## Document role

This S6 artifact classifies the testing harness's concurrency, race, timing,
cleanup, sharding, resource-allocation, and authority-sensitive hazards. It is
evidence for later NLSpec drafting, not a behavior owner.

S6 did not rewrite harness code, tests, fixtures, cleanup scripts, retry loops,
timeouts, sharding, resource allocation, generated artifacts, or lockfiles. S6
also did not execute service-backed targets, browser E2E targets, Docker,
Docker Compose, reset routes, cleanup paths, formatters, generators, baseline
refreshes, or broad verification gates.

## Session metadata

| Field | Value |
|---|---|
| Repository root | `/home/askahn/code/cartulary` |
| Branch state at S6 refresh | `## main...origin/main` |
| HEAD revision | `900d6d858982e8db00b0b15d950d708b43e9e7c9` |
| S6 refresh timestamp | `2026-05-08T23:38:53-04:00` |
| Runtime platform observed | `Linux DeskRip 6.6.114.1-microsoft-standard-WSL2 x86_64 GNU/Linux` |
| Recovery write boundary | `docs/testing-harness-spec-recovery-docs/**` only |

## Inspected source groups

| Group | Sources |
|---|---|
| Command and schedule surfaces | `Makefile`, `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`, `tools/check_schedule_manifest.json`, `tools/service_backed_schedule_manifest.json`, `tools/browser_e2e_batch_manifest.json`, `tools/scheduler_resource_registry.json` |
| Scheduler and sharding code | `scripts/run-check-schedule.mjs`, `scripts/run-service-backed-schedule.mjs`, `scripts/lib/scheduler*.mjs`, `scripts/lib/scheduler/**`, `tools/harness/backend/go-shard-plan.mjs`, `scripts/lib/browser-shard-plan.mjs`, `tools/harness/backend/go-target-runner.mjs` |
| Service setup and cleanup | `tools/testservices/main.go`, `internal/testutil/testcontainersx/testcontainersx.go`, `internal/testutil/pgtest/pgtest.go`, `internal/testutil/s3test/s3test.go`, `internal/testutil/suiteservices/**` |
| Browser and local-dev runtime | `scripts/start-web-e2e.sh`, `scripts/reset-web-e2e-stack.sh`, `tools/harness/readiness/process-lifecycle.sh`, `scripts/dev-services.sh`, `scripts/dev-stack.sh`, `docker-compose.dev.yml`, `apps/web/playwright*.config.ts`, `apps/web/e2e/**` |
| Test reset and authority-sensitive hook | `internal/testutil/testruntime/reset.go`, `internal/testutil/testruntime/reset_integration_test.go` |
| Failure and timing examples | `tools/testservices/*_test.go`, `internal/testutil/testcontainersx/*_test.go`, `internal/testutil/pgtest/*_test.go`, `internal/testutil/s3test/*_test.go`, `scripts/test-*scheduler*.sh`, `scripts/test-web-e2e-lifecycle.sh`, `scripts/test-browser-shard-plan.sh`, `scripts/test-run-frontend-unit.sh` |

## Classification rules used

| Class | S6 meaning |
|---|---|
| `confirmed_failure` | Runtime or retained-artifact evidence shows the failure. No new S6 rows use this class without selected runtime evidence. |
| `plausible_latent_failure` | Static source shows a race window, shared resource, partial mitigation, or cleanup uncertainty. |
| `accepted_nondeterminism` | Behavior is source-observed, bounded, and expected to remain nondeterministic unless an owner changes policy. |
| `authority_required` | Behavior is destructive, public-contract-like, product-spec adjacent, or requires maintainer judgment. |
| `source_limited` | Static evidence exists, but live runtime, platform, cleanup, or timing behavior is not proved. |

## Resource coverage

Every shared mutable resource from S3/S4 is covered below. Allocation details
remain in `resource-allocation-register.md`; this table records hazard coverage.

| Resource rows | Resource group | Hazard rows | Timing rows | Coverage status |
|---|---|---|---|---|
| `RES-0001` through `RES-0010` | scheduler host/service/process/browser lanes | `RTR-0001`, `RTR-0002`, `RTR-0015` | `TMR-0020`, `TMR-0021` | Scheduler-accounting only; concrete capacity remains source-limited. |
| `RES-0011`, `ART-0013` through `ART-0017`, `ART-0029` | retained result roots, run summaries, service events, fixture reports | `RTR-0016` | `TMR-0027` | Freshness and selected-run identity remain source-limited. |
| `ART-0008`, `ART-0010`, `ART-0011`, `ART-0009` | generated Make/schedule/ledger/baseline artifacts | `RTR-0002`, `RTR-0003` | `TMR-0021`, `TMR-0027` | Execution-driving downstream artifacts, not behavior owners. |
| `RES-0012`, `RES-0013`, `CLN-0011`, `CLN-0012` | service leases, detached reaper, Docker containers, stale cleanup | `RTR-0004`, `RTR-0007`, `RTR-0009` | `TMR-0001`, `TMR-0009`, `TMR-0010`, `TMR-0028` | Reaper scheduling is not cleanup completion evidence. |
| `RES-0014`, `RES-0015`, `ART-0025`, `CLN-0005` through `CLN-0008` | Postgres templates, DB names, package/group fixtures, transactions | `RTR-0005`, `RTR-0010`, `RTR-0018` | `TMR-0003`, `TMR-0006`, `TMR-0017`, `TMR-0022` | Name isolation and locks mitigate collisions; active connection cleanup is source-limited. |
| `RES-0016`, `ART-0026`, `CLN-0009`, `CLN-0010` | object-store buckets, package buckets, prefixes, object namespaces | `RTR-0006`, `RTR-0010` | `TMR-0005`, `TMR-0023` | Bucket isolation depends on naming, mutexes, and successful cleanup. |
| `RES-0017`, `RES-0018`, `RES-0019`, `CLN-0014`, `CLN-0015` | browser, dev-stack, and Compose ports | `RTR-0011`, `RTR-0019` | `TMR-0008`, `TMR-0010`, `TMR-0013`, `TMR-0014`, `TMR-0015` | Dynamic browser ports race before bind; fixed dev/Compose ports are not parallel-safe. |
| `RES-0020`, `CLN-0015`, `CLN-0016` | shell process groups, monitors, runtime roots | `RTR-0008`, `RTR-0012` | `TMR-0009`, `TMR-0029` | Traps and process groups are cleanup paths, not timeout/interrupt guarantees. |
| `RES-0021`, `ART-0027` | browser runtime root, stack env/json, logs | `RTR-0012`, `RTR-0016` | `TMR-0008`, `TMR-0029` | Runtime root may be retained intentionally under target artifacts. |
| `RES-0022`, `ART-0028`, `CLN-0017`, `CLN-0018` | Playwright state dir, lock, worker admin manifest, cleanup markers, sessions | `RTR-0013` | `TMR-0011`, `TMR-0012`, `TMR-0024`, `TMR-0025` | Shared state is lock-mitigated; abrupt-exit behavior is source-limited. |
| `RES-0023`, `RES-0026`, `CLN-0019`, `HAZ-S3-0014` | reset boundary files and app test runtime reset state | `RTR-0014` | `TMR-0016`, `TMR-0030` | Destructive reset behavior is authority-required. |
| `RES-0024` | shell harness scratch dirs | `RTR-0020` | `TMR-0031` | Scratch allocation is validated, but caller trap behavior remains source-limited. |
| `RES-0025`, `ART-0024`, `HAZ-S3-0015` | external Go caches | `RTR-0017` | `TMR-0026` | Proposed default is tool-managed outside cleanup contract pending owner decision. |
| `ART-0030`, `HAZ-S3-0013`, `HAZ-S4-0009` | direct package-script artifacts and tool-native outputs | `RTR-0015` | `TMR-0020`, `TMR-0032` | Package scripts bypass Make result-root, scheduler, and cleanup guarantees unless adopted. |
| `ART-0003`, `AMB-0022` | visual snapshots and browser platform baselines | `RTR-0021` | `TMR-0012`, `TMR-0025` | Snapshot update/platform authority is open. |

## Hazard register

| Hazard ID | Surface | Trigger | Shared resource | Timing assumption | Observable failure | Current mitigation | Spec gap | Severity | Evidence | Evidence status | Classification | Proposed disposition | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| RTR-0001 | Scheduler lanes versus concrete capacity | check/service-backed schedules run concurrent work under logical resource claims | `host_cpu`, `host_io`, `postgres`, `object_store`, `process`, `browser_stack`, `browser_stage_*`, `postgres_reset` | Lane release on work-unit completion equals resource availability | Service contention, timeouts, port conflicts, DB/bucket pressure, misleading capacity assumptions | scheduler limits, `suite_service_stack=1`, stage lanes, reset lane, fixture naming | Must state lanes are scheduling constraints, not host/Docker/Postgres/object-store capacity guarantees | high | `HAZ-S4-0001`; `RES-0001` through `RES-0010`; scheduler manifests | `observed/source_limit` | `plausible_latent_failure` | `preserve_with_clarification` | Do not infer concrete capacity from default limits. |
| RTR-0002 | Generated schedules and manifests driving concurrency | topology or phase ownership changes without refreshed derived manifests | generated Make, check/service/browser schedules, browser batch manifest | Committed generated artifacts match owner manifests | stale work graph, stale resource claims, wrong sharding, drift failures | `phase-schedule-drift`, generated artifact policy, `agent-finalize` | Need authority text: generated outputs may drive execution but do not own behavior | high | `ART-0008`, `ART-0010`, `HAZ-S3-0003`, `HAZ-S3-0004` | `observed` | `plausible_latent_failure` | `preserve_with_clarification` | Keep codegen drift separate from migration/runtime drift. |
| RTR-0003 | Duration baseline contamination and stale weights | retained timing contains retries or service contamination, or shard set changes | committed duration baseline JSON, retained run artifacts | Selected run timing is clean and representative | baseline drift failure, skewed scheduler planning, misleading timing claims | coverage/drift checks, contamination diagnostics, explicit `RESULTS_DIR` refresh path | Successful-run provenance and contamination exclusions must be explicit | medium | `ART-0011`, `HAZ-S3-0005`, duration drift helpers | `observed/source_limit` | `plausible_latent_failure` | `preserve_with_clarification` | No baseline refresh was run. |
| RTR-0004 | Docker/testcontainers live readiness | service-backed work needs Docker, suite Postgres, or suite object-store | Docker daemon, testcontainers, managed containers | Static preflight/start/readiness loops match live host behavior | preflight failure, start retry exhaustion, readiness timeout, partial startup | Docker ping, port wait, client ping/ListBuckets, retry wrappers | Live host behavior requires authorized runtime evidence | high | `SVC-0002`, `SVC-0005`, `SVC-0007`, `FAIL-0005`, `FAIL-0006`, `SL-0012` | `observed/source_limit` | `source_limited` | keep source-limited | S8 owns supported platform/tool profile. |
| RTR-0005 | Postgres template, clone, package DB, and transaction state | concurrent Go shards/tests clone, reset, drop, or reuse databases | template DB, cloned DBs, package/group DBs, transaction handles | name/hash/counter isolation and mutexes are enough for parallel fixture use | create collision, stale template, reset leakage, rollback/cleanup failure, active connection blocks drop | generated DB names, package/group mutexes, transaction rollback, force drop, leak checks | Need explicit collision/release and active-connection guarantee status | high | `SVC-0003`, `SVC-0004`, `RES-0014`, `RES-0015`, `HAZ-S3-0007`, `HAZ-S3-0008` | `observed/source_limit` | `plausible_latent_failure` | preserve with source limits | Ordinary DB operation timeouts remain unknown in S4. |
| RTR-0006 | object-store bucket and prefix reuse | concurrent tests or package fixtures reuse bucket names or prefixes | object-store buckets, package buckets, object prefixes | bucket naming and prefix cleanup prevent cross-test object leakage | bucket create conflict, stale object leakage, failed prefix cleanup, partial bucket cleanup | generated bucket names, package mutex, prefix cleanup, per-test cleanup | Need object cleanup timeout/failure contract | medium | `SVC-0006`, `RES-0016`, `HAZ-S3-0009`, `ART-0026` | `observed/source_limit` | `plausible_latent_failure` | preserve with source limits | Prefix cleanup uses caller context. |
| RTR-0007 | Detached service reaper completion | suite cleanup schedules detached `terminate-suite` | service lease, reaper process, Docker containers | Scheduling detached reaper is enough cleanup evidence | containers remain, cleanup result unknown to parent, stale service state | direct cleanup fallback if reaper scheduling fails, labels, lease schema | Need best-effort versus guaranteed classification | high | `HAZ-S4-0003`, `RES-0012`, `CLN-0011`, `AMB-0024` | `observed/source_limit` | `source_limited` | source-limited best effort pending evidence | Scheduling event is not completion evidence. |
| RTR-0008 | Timeout, interrupt, and parent-death cleanup | INT/TERM, child timeout, CI cancellation, parent death, abrupt abort | process groups, suite services, DB/bucket fixtures, browser stack, result artifacts | shell traps, contexts, and Go `t.Cleanup` run for all abrupt exits | orphaned processes, unreleased ports, leaked containers, partial summaries | traps, process monitors, `stop_process_group`, contexts, `t.Cleanup` | Need controlled signal/timeout evidence or owner decision | high | `SL-0014`, `CLN-0011` through `CLN-0019`, `FAIL-0022`, `PCS-0013` | `observed/source_limit` | `source_limited` | do not call cleanup guaranteed | Applies across browser, testservices, Go, and Playwright paths. |
| RTR-0009 | Stale janitor destructive bounds | stale DB/bucket/container metadata found at suite startup/teardown | generated DBs, buckets, containers, metadata, cleanup summaries | generated-name and metadata checks prove ownership | unintended deletion, skipped stale cleanup, cleanup failure blocks startup | generated-name guards, labels, summaries, bounded workers, leak checks | Maintainer must decide sufficient proof before deletion | high | `HAZ-S4-0002`, `HAZ-S4-0004`, `AMB-0027`, `RES-0013`, `RES-0014`, `RES-0016` | `observed/source_limit` | `authority_required` | route to S8; keep source-limited | Do not broaden janitor authority. |
| RTR-0010 | DB and bucket naming collision limits | concurrent suites/processes allocate DBs or buckets | Postgres DB names, object-store bucket names | suite/process hashes and counters avoid practical collisions | create failure, stale resource confusion, cleanup ambiguity | hash/counter/suffix names, 63-char bounds, mutexes for package reuse | Collision handling should be stated as failure, not impossible condition | medium | `RES-0014`, `RES-0016`, `SVC-0004`, `SVC-0006` | `observed/source_limit` | `plausible_latent_failure` | preserve observed collision behavior only | No runtime collision test was run. |
| RTR-0011 | Browser dynamic port allocation race | browser stack picks dynamic ports, or caller supplies configured ports | backend/frontend ports | OS port-0 candidate stays free until backend/Vite binds | bind conflict, wrong listener, readiness timeout, release diagnostic | configured-port checks, dynamic candidates, Vite `--strictPort`, scheduler lanes | Need `ss` dependency or best-effort platform contract | high | `HAZ-S4-0005`, `RES-0017`, `AMB-0025` | `observed/source_limit` | `plausible_latent_failure` | preserve with platform decision | `ss` absence disables collision/release checks in source. |
| RTR-0012 | Browser process group and runtime-root cleanup | backend/frontend/child starts then startup, child, or cleanup path fails | process groups, monitors, runtime root, stack logs/env/json | trap and process helper cleanup finish on all paths | orphaned process, retained runtime root, stale stack metadata | `setsid`, monitor, TERM/KILL sequence, runtime-root retention rules | Cleanup strength under timeout/interrupt is source-limited | high | `SVC-0009`, `SVC-0010`, `RES-0020`, `RES-0021`, `CLN-0015`, `CLN-0016` | `observed/source_limit` | `source_limited` | source-limited until runtime evidence | Target artifact dirs intentionally retain logs/metadata. |
| RTR-0013 | Playwright shared state lock and manifest staleness | external-server/shared state, parallel workers, interrupted run | state dir, global setup lock, worker admin manifest, cleanup markers, sessions | lock/manifest state matches current run and worker count | lock timeout, missing admin, stale session, skipped cleanup marker | `mkdir` lock deadline, manifest validation, reset state cleanup, worker teardown | Need stale-state and abrupt-exit contract | medium | `HAZ-S4-0006`, `SVC-0013`, `RES-0022`, `CLN-0017`, `CLN-0018` | `observed/source_limit` | `plausible_latent_failure` | preserve with clarification | Playwright interruption behavior unobserved. |
| RTR-0014 | Reset route targets wrong backend or partially resets state | reset script points at wrong origin, or test route is enabled on unintended backend | app DB, object store bucket, bootstrap admin, reset response files | reset route only exists in safe test-enabled backend | destructive app state reset, partial mutation, state dir not cleared | browser-owned backend sets `CARTULARY_ENABLE_TEST_ROUTES=1`, reset response validation | Authority/security boundary required before any normative reset contract | high | `HAZ-S4-0007`, `SVC-0014`, `RES-0023`, `RES-0026`, `AMB-0006`, `MSC-0001` | `observed/source_limit` | `authority_required` | route to S8 | Do not expose as product API. |
| RTR-0015 | Direct package scripts bypass Make policy | user runs root/app pnpm scripts directly | tool-native outputs, worker defaults, browser env, cleanup roots | package tool defaults are acceptable harness behavior | missing result root/run ID, scheduler limits skipped, weaker cleanup/report guarantees | some browser scripts re-enter wrappers; Make remains canonical by default | Maintainer must decide first-class contract status | medium | `HAZ-S4-0009`, `ART-0030`, `EP-0016`, `FAIL-0018`, `PCS-0008` | `observed` | `authority_required` | route to S8 | Direct package outputs remain separate from Make outputs. |
| RTR-0016 | Retained artifact freshness and selected-run identity | investigation/drift/fixture tools read retained results | `.cartulary/test-results/**`, summaries, logs, fixture reports | selected artifacts are current, complete, and relevant | stale newest-run evidence, failed schema reads, misleading diagnostics | run IDs, artifact discovery, explicit `RESULTS_DIR` for some tools | Need stable selection rule for normative claims | medium | `HAZ-S3-0001`, `ART-0013` through `ART-0017`, `FAIL-0027`, `PCS-0014`, `SL-0004`, `SL-0009` | `observed/source_limit` | `source_limited` | preserve explicit source limit | Newest-run fallback is human-investigation only until owner adopts more. |
| RTR-0017 | External Go cache staleness | Go cache persists outside repo cleanup | `/tmp/cartulary-go-build`, `/tmp/cartulary-go-mod` | Go tooling owns cache consistency | stale/corrupt cache affects tools or performance | Go cache semantics only | Need cleanup-scope decision | low | `RES-0025`, `ART-0024`, `AMB-0021` | `observed/source_limit` | `authority_required` | likely `exclude_from_contract` pending S8 | Do not expand cleanup behavior. |
| RTR-0018 | Go shared report locks and shard capture | multiple shards/targets capture shared Go reports | shared report dirs, `capture.lock`, warm dependency lock | directory locks serialize capture and release before timeout | lock timeout, stale lock cleanup, command mismatch, missing accounting | mkdir lock, stale pid handling, metadata checks, timeout failure | Need lock timeout and stale-lock behavior in timing rules | medium | `HAZ-S3-0006`, `tools/harness/backend/go-target-runner.mjs` | `observed/source_limit` | `plausible_latent_failure` | preserve with clarification | Lock polling uses 100ms sleep in source. |
| RTR-0019 | Local-dev Compose fixed ports and persistent volumes | local dev or standalone browser mode starts/reuses Compose services | ports `5432` and `8333`, Compose volumes, default bucket | persistent local state is acceptable for harness contract | port conflict, stale DB/object state, object-store not reset by `db-reset` | Compose healthchecks, wait scripts, standalone DB drop/create | Contract boundary for local-dev persistence required | medium | `HAZ-S4-0008`, `SVC-0008`, `SVC-0012`, `SVC-0015`, `RES-0018`, `RES-0019` | `observed/source_limit` | `authority_required` | route to S8 | Keep separate from verification gates. |
| RTR-0020 | Shell harness scratch directories | shell harness self-tests allocate scratch paths | `CARTULARY_HARNESS_SCRATCH_ROOT`, temp dirs | trap cleanup always removes registered scratch dirs | stale scratch dirs, unsafe cleanup path refusal, partial self-test artifacts | basename validation, outside-repo guard, cleanup traps | Signal behavior remains caller-script-specific | low | `RES-0024`, `CLN-0004`, `tools/harness/test-support/harness-scratch.sh` | `observed/source_limit` | `plausible_latent_failure` | preserve with clarification | Scratch roots are not product state. |
| RTR-0021 | Visual snapshot platform authority | visual tests compare committed PNG baselines | Playwright snapshots, browser/runtime platform | Linux snapshot baseline is intended across supported environments | false visual failures, unsupported refresh mutation, platform drift | committed snapshots, visual target isolation, no update command found | Need platform/browser/update authority question | medium | `ART-0003`, `AMB-0022`, `MSC-0007` | `observed/source_limit` | `authority_required` | route to S8/browser owner | Validation-only targets must not refresh snapshots. |

## Failure and partial-state linkage

| Linked failure/state rows | S6 hazard coverage | Notes |
|---|---|---|
| `FAIL-0005`, `PCS-0001` | `RTR-0004`, `RTR-0008` | Docker/testcontainers preflight remains source-limited and platform-sensitive. |
| `FAIL-0006`, `PCS-0002` | `RTR-0004`, `RTR-0007`, `RTR-0008` | Startup retry exhaustion and cleanup completion were not runtime-observed. |
| `FAIL-0007`, `FAIL-0015`, `PCS-0011` | `RTR-0005`, `RTR-0006`, `RTR-0009` | Fixture setup, stale janitors, and destructive cleanup require source limits or owner decision. |
| `FAIL-0008`, `PCS-0004`, `PCS-0009` | `RTR-0001`, `RTR-0010`, `RTR-0011`, `RTR-0018` | Resource conflicts may come from logical lanes, concrete capacity, locks, ports, DBs, or buckets. |
| `FAIL-0009`, `FAIL-0021`, `PCS-0005` | `RTR-0011`, `RTR-0012` | Browser readiness timeouts leave process/port cleanup questions. |
| `FAIL-0010`, `PCS-0007`, `PCS-0015` | `RTR-0014` | Reset can fail after response/status artifacts are written; authority remains open. |
| `FAIL-0011`, `PCS-0003`, `PCS-0006` | `RTR-0007`, `RTR-0008`, `RTR-0012` | Product assertion failures still require harness cleanup when services/browser stacks are active. |
| `FAIL-0012`, `PCS-0017` | `RTR-0008`, `RTR-0012`, `RTR-0013` | Timeouts are harness-operational unless a specific product test owns the deadline. |
| `FAIL-0013`, `FAIL-0023`, `FAIL-0026`, `PCS-0016`, `PCS-0018`, `PCS-0019` | `RTR-0002`, `RTR-0003`, `RTR-0016`, `RTR-0018` | Generated/stale/missing artifacts and schema gaps remain harness operational or source-limited. |
| `FAIL-0014`, `PCS-0010`, `PCS-0013` | `RTR-0007`, `RTR-0008`, `RTR-0009`, `RTR-0012` | Cleanup failure must distinguish attempted cleanup, scheduled cleanup, and verified cleanup. |
| `FAIL-0016` | `RTR-0017`, `RTR-0019`, `RTR-0020` | Cleanup target boundaries exclude external/local state unless owner expands contract. |
| `FAIL-0018`, `PCS-0008` | `RTR-0015` | Direct package scripts remain authority-required. |
| `FAIL-0019` | `RTR-0015`, `RTR-0020` | Public env contract and precedence remain authority-required. |
| `FAIL-0020` | `RTR-0004`, `RTR-0011`, `RTR-0019` | Unsupported platform behavior is source-limited except where source explicitly validates or rejects a tool. |
| `FAIL-0022` | `RTR-0007`, `RTR-0008`, `RTR-0012`, `RTR-0013` | Cancellation, interrupt, and parent-death cleanup are not guaranteed by static traps alone. |
| `FAIL-0027`, `PCS-0014` | `RTR-0016` | Retained artifacts need selected run identity for durable claims. |
| `FAIL-0028` | `RTR-0016` plus `MSC-0010` | Provider CI annotations remain unavailable while `.github/**` is absent. |

## Checklist support

| Field | Value |
|---|---|
| Status | `complete` for documentation-only S6 hazard classification. |
| Blockers | Runtime service readiness, browser runtime, Docker/Compose behavior, reset route execution, timeout/interrupt cleanup, detached reaper completion, active DB connection cleanup, stale janitor execution, unsupported platform matrix, env override matrix, CI provider workflows, and snapshot update authority remain source-limited or owner-required. |
| Findings | Scheduler lanes are accounting constraints; generated schedules and duration baselines can affect execution; cleanup scheduling is not cleanup completion; reset route and stale janitors are authority-sensitive; package scripts bypass Make policy; every S3/S4 shared resource now maps to an S6 hazard row. |
| Handoff notes | S7 can consume `RTR-*`, `CONC-*`, and `TMR-*` rows as NLSpec inputs while preserving all `source_limit` and `authority_required` classifications. |
| No-change audit | Documentation-only changes under `docs/testing-harness-spec-recovery-docs/**`; no harness implementation, test, fixture, cleanup, generated, lockfile, timing, retry, sharding, or allocation behavior was changed. |
