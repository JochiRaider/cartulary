---
doc_id: THR-S7-RUNTIME-EVIDENCE-REGISTER-2026-05-09
title: S7 Testing-Harness Runtime Evidence Register
status: active
role: runtime-evidence-register
---

# S7 Testing-Harness Runtime Evidence Register

## Document role

This register records the authorized S7 testing-harness recovery evidence pass
performed on 2026-05-09. It is evidence only: it does not rewrite harness code,
refresh snapshots, refresh baselines, hand-edit generated files, or promote any
source-limited behavior into final S7 `MUST` language.

Runtime-sensitive claims may close only through a selected runtime run recorded
here, or through an explicit maintainer decision. All other claims remain
`source_limit` or `maintainer_decision_required`.

## Evidence selection rule

Retained artifacts are strong evidence only when selected by explicit result
directory and run ID. Ambient newest-run lookup is allowed only for human
investigation if the tool prints the selected run being inspected; it is not
sufficient for normative claims.

For this pass:

- `RUN_ROOT`: `/home/askahn/code/cartulary/.cartulary/test-results/s7-recovery`
- artifact root pattern: `${RUN_ROOT}/<slug>/<CARTULARY_TEST_RUN_ID>`
- each selected run has `_evidence/result.env`,
  `_evidence/platform-profile.env`, `_evidence/stdout.log`, and
  `_evidence/stderr.log`;
- each `result.env` records command, UTC start/end, exit status, git HEAD, git
  status, result directory, run ID, and artifact root;
- each `platform-profile.env` records host/tool profile and `make doctor`
  output.

## Authorization and decisions

| Decision ID | Decision | Owner/source | Date | Effect |
|---|---|---|---|---|
| `S7-DEC-0001` | Authorized selected runtime evidence collection for the listed commands and scenarios. | Maintainer request in the implementation instruction. | 2026-05-09 | Runtime commands in this register may be used as selected evidence, subject to each validation result. |
| `S7-DEC-0002` | Retained artifacts are strong evidence only when selected by explicit result dir/run ID. Newest-run fallback is human-investigation only. | Maintainer request in the implementation instruction. | 2026-05-09 | Closes the artifact-provenance rule for this evidence pass; unresolved artifacts without selected identity remain source-limited. |
| `S7-DEC-0003` | S7 may draft source-limited text, but final `MUST` language needs selected evidence or owner decision. | Maintainer request in the implementation instruction. | 2026-05-09 | Preserves `source_limit` and `maintainer_decision_required` routing for unresolved rows. |
| `S7-DEC-0004` | Generated artifacts are downstream execution inputs only. Upstream specs, manifests, SQL, Make/task definitions, and adopted NLSpec own behavior. | Maintainer request in the implementation instruction. | 2026-05-09 | Drift checks prove freshness; generated files do not define behavior. |

## Platform profile

Current selected-host profile from run `s7-20260509T130820Z-01-profile`:

| Field | Value |
|---|---|
| Git HEAD | `6c1feb5a262e337431e0ec07f7f72ddaafc5ef19` |
| Git status | `## main...origin/main [ahead 1]` |
| Host | `Linux DeskRip 6.6.114.1-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Mon Dec 1 20:46:23 UTC 2025 x86_64 GNU/Linux` |
| Go | `go version go1.26.3 linux/amd64` |
| Node | `v24.15.0` |
| pnpm | `10.33.0` |
| Docker | `Docker version 28.5.1, build e180ab8` |
| Docker Compose | `Docker Compose version v2.40.3-desktop.1` |
| Playwright | `Version 1.59.1` |
| Required tools observed | `bash`, `curl`, `ss`, `setsid`, `realpath`, `timeout` |
| Doctor | `make doctor` exited `0` in selected run `s7-20260509T130820Z-01-profile`. |

This proves current-host readiness only. It does not define a portable platform
matrix without an additional platform-owner decision.

## Selected run index

Every row below used `CARTULARY_OUTPUT_MODE=summary` and a unique
`CARTULARY_TEST_RESULTS_DIR`/`CARTULARY_TEST_RUN_ID` pair. The absolute artifact
root for each row is `${RUN_ROOT}/<slug>/<run-id>`.

| Slug | Run ID | Command surface | UTC start | UTC end | Exit | Validation / disposition |
|---|---|---|---|---|---|---|
| `00-discovery` | `s7-20260509T130840Z-00-discovery` | `make help-all`; task-surface, target, and phase explainers | 2026-05-09T13:08:41Z | 2026-05-09T13:08:47Z | 0 | Static discovery succeeded. |
| `01-profile` | `s7-20260509T130820Z-01-profile` | `make doctor` | 2026-05-09T13:08:21Z | 2026-05-09T13:08:21Z | 0 | Current-host tool readiness observed. |
| `02-test-fast` | `s7-20260509T130906Z-02-test-fast` | `make test-fast` | 2026-05-09T13:09:06Z | 2026-05-09T13:09:42Z | 0 | Fast-loop readiness observed for current host. |
| `03-test` | `s7-20260509T131001Z-03-test` | `make test` | 2026-05-09T13:10:01Z | 2026-05-09T13:11:59Z | 0 | Full-corpus readiness observed; reader validation run `35` succeeded. |
| `04-check` | `s7-20260509T131217Z-04-check` | `make check` | 2026-05-09T13:12:18Z | 2026-05-09T13:13:37Z | 0 | Developer gate readiness observed; reader validation run `35` succeeded. |
| `04-check-scheduler-drift` | `s7-20260509T131359Z-04-check-scheduler-drift` | scheduler event/order and summary timing drift against `04-check` | 2026-05-09T13:14:00Z | 2026-05-09T13:14:00Z | 0 | Scheduler event/order and timing readers passed. |
| `05-ci` | `s7-20260509T131418Z-05-ci` | `make ci` | 2026-05-09T13:14:19Z | 2026-05-09T13:17:02Z | 2 | Current CI failed in harness-smoke/phase-map accounting. No CI readiness claim. |
| `06-release-check` | `s7-20260509T131736Z-06-release-check` | `make release-check` | 2026-05-09T13:17:36Z | 2026-05-09T13:20:13Z | 2 | Current release-check failed with same harness-smoke/phase-map accounting failure. No release readiness claim. |
| `07-service-images` | `s7-20260509T132044Z-07-service-images` | `make test-service-images` | 2026-05-09T13:20:44Z | 2026-05-09T13:20:44Z | 0 | Current-host Docker/testcontainers image readiness observed. |
| `08-backend-store` | `s7-20260509T132044Z-08-backend-store` | `make backend-store` | 2026-05-09T13:20:45Z | 2026-05-09T13:20:53Z | 0 | Service-backed store readiness observed; fixture reader in run `35` succeeded. |
| `09-backend-integration` | `s7-20260509T132053Z-09-backend-integration` | `make backend-integration` | 2026-05-09T13:20:53Z | 2026-05-09T13:21:07Z | 0 | Service-backed integration readiness observed. |
| `10-backend-process` | `s7-20260509T132107Z-10-backend-process` | `make backend-process` | 2026-05-09T13:21:08Z | 2026-05-09T13:21:14Z | 0 | Service-backed process readiness observed. |
| `11-backend-integration-support` | `s7-20260509T132115Z-11-backend-integration-support` | `make backend-integration-support` | 2026-05-09T13:21:15Z | 2026-05-09T13:21:31Z | 0 | Support harness evidence only; not authoritative product phase completion. |
| `12-svc-phase4` | `s7-20260509T132131Z-12-svc-phase4` | `make service-backed-slice PHASE=phase4` | 2026-05-09T13:21:32Z | 2026-05-09T13:21:55Z | 0 | Phase4 service-backed selected run passed. |
| `13-svc-phase5` | `s7-20260509T132155Z-13-svc-phase5` | `make service-backed-slice PHASE=phase5` | 2026-05-09T13:21:56Z | 2026-05-09T13:22:31Z | 0 | Phase5 service-backed selected run passed. |
| `14-svc-phase6` | `s7-20260509T132231Z-14-svc-phase6` | `make service-backed-slice PHASE=phase6` | 2026-05-09T13:22:31Z | 2026-05-09T13:23:02Z | 0 | Phase6 service-backed selected run passed. |
| `12-14-svc-phase-explain` | `s7-20260509T132318Z-12-14-svc-phase-explain` | explicit `explain-run ... DETAIL=children` for phase runs | 2026-05-09T13:23:18Z | 2026-05-09T13:23:18Z | 0 | Child reader resolved selected phase children. |
| `15-browser-webserver` | `s7-20260509T132338Z-15-browser-webserver` | `make browser-e2e-webserver-backed` | 2026-05-09T13:23:39Z | 2026-05-09T13:24:14Z | 0 | Browser shared-stack readiness observed; reader validation run `35` succeeded. |
| `16-browser-stateful` | `s7-20260509T132414Z-16-browser-stateful` | `make browser-e2e-stateful` | 2026-05-09T13:24:14Z | 2026-05-09T13:24:33Z | 0 | Browser stateful run passed. |
| `17-browser-resettable` | `s7-20260509T132433Z-17-browser-resettable` | `make browser-e2e-resettable` | 2026-05-09T13:24:34Z | 2026-05-09T13:25:08Z | 0 | Harness reset flow passed; public reset API remains `AUTH-0004`. |
| `18-browser-measurement` | `s7-20260509T132508Z-18-browser-measurement` | `make browser-e2e-measurement` | 2026-05-09T13:25:09Z | 2026-05-09T13:25:19Z | 0 | Measurement validation passed. |
| `19-browser-visual` | `s7-20260509T132519Z-19-browser-visual` | `make browser-e2e-visual` | 2026-05-09T13:25:20Z | 2026-05-09T13:25:48Z | 0 | Visual validation passed. No snapshot update command was run. |
| `15-19-browser-port-release` | `s7-20260509T132623Z-15-19-browser-port-release` | `ss` and Docker post-run checks for browser runs | 2026-05-09T13:26:23Z | 2026-05-09T13:26:23Z | 0 | Recorded browser ports were released where `ss` is available. |
| `20-services-up` | `s7-20260509T132644Z-20-services-up` | `make services-up` plus Compose and `ss` checks | 2026-05-09T13:26:45Z | 2026-05-09T13:26:45Z | 0 | Local Compose startup observed; contract status remains `AUTH-0005`. |
| `21-db-up` | `s7-20260509T132645Z-21-db-up` | `make db-up` plus Compose checks | 2026-05-09T13:26:46Z | 2026-05-09T13:26:46Z | 0 | Local DB startup observed; local-dev contract remains owner-required. |
| `22-db-reset` | `s7-20260509T132646Z-22-db-reset` | `make db-reset` plus Compose checks | 2026-05-09T13:26:47Z | 2026-05-09T13:26:48Z | 0 | Local DB reset observed; destructive/local-dev status remains owner-required. |
| `23-dev-int` | `s7-20260509T132648Z-23-dev-int` | `timeout --signal=INT --kill-after=20s 60s make dev` | 2026-05-09T13:26:49Z | 2026-05-09T13:26:50Z | 2 | Failed before interrupt due unwritable `/var/lib/cartulary/backups`; config failure evidence only. |
| `23b-dev-int-writable` | `s7-20260509T133010Z-23b-dev-int-writable` | same dev interrupt with disposable writable root overrides | 2026-05-09T13:30:10Z | 2026-05-09T13:31:08Z | 124 | Signal-derived dev interruption observed; ports 8080/5173 released. |
| `24-baseline-drift` | `s7-20260509T133108Z-24-baseline-drift` | baseline coverage/drift readers against selected runs | 2026-05-09T13:31:08Z | 2026-05-09T13:31:09Z | 0 | All selected drift readers passed; no baseline refresh run. |
| `25-usage-fail` | `s7-20260509T133109Z-25-usage-fail` | `make service-backed-slice` without `PHASE` | 2026-05-09T13:31:09Z | 2026-05-09T13:31:10Z | 2 | Usage failure shape observed. |
| `26-scheduler-config-fail` | `s7-20260509T133110Z-26-scheduler-config-fail` | `CHECK_HOST_CPU_JOBS=0 make check` | 2026-05-09T13:31:10Z | 2026-05-09T13:31:11Z | 2 | Config failure shape observed before child work. |
| `27-docker-preflight-fail` | `s7-20260509T133111Z-27-docker-preflight-fail` | missing Docker socket with `make backend-store` | 2026-05-09T13:31:11Z | 2026-05-09T13:31:12Z | 2 | Preflight failure evidence observed; no managed containers remained. |
| `28-browser-start-fail` | `s7-20260509T133112Z-28-browser-start-fail` | attempted `CARTULARY_SERVER_BIN=/bin/false make browser-e2e-webserver-backed` | 2026-05-09T13:31:12Z | 2026-05-09T13:31:47Z | 0 | Investigation-only: env override was overwritten and target passed. |
| `28b-browser-start-fail` | `s7-20260509T133254Z-28b-browser-start-fail` | corrected `SERVER_BIN=/bin/false make browser-e2e-webserver-backed` | 2026-05-09T13:32:54Z | 2026-05-09T13:32:55Z | 2 | Build-prerequisite failure observed before browser stack startup. |
| `29-normal-testservices` | `s7-20260509T133255Z-29-normal-testservices` | `make test-service-images`; `cartulary-test-services run -- exit 0` | 2026-05-09T13:32:55Z | 2026-05-09T13:32:58Z | 0 | Normal testservices path observed. |
| `30-child-exit` | `s7-20260509T133258Z-30-child-exit` | `cartulary-test-services run -- exit 42` | 2026-05-09T13:32:58Z | 2026-05-09T13:33:01Z | 42 | Parent propagated child exit; cleanup details in cleanup register. |
| `31-timeout` | `s7-20260509T133301Z-31-timeout` | TERM timeout around `cartulary-test-services run -- sleep 300` | 2026-05-09T13:33:01Z | 2026-05-09T13:33:30Z | 124 | Timeout path observed; cleanup details in cleanup register. |
| `32-interrupt` | `s7-20260509T133330Z-32-interrupt` | INT timeout around `cartulary-test-services run -- sleep 300` | 2026-05-09T13:33:30Z | 2026-05-09T13:34:00Z | 124 | Interrupt path observed; cleanup details in cleanup register. |
| `33-cleanup-unit-evidence` | `s7-20260509T133400Z-33-cleanup-unit-evidence` | focused cleanup/janitor/reaper `go test ./tools/testservices -run ...` | 2026-05-09T13:34:00Z | 2026-05-09T13:34:01Z | 0 | Unit-level cleanup/janitor/leak/reaper checks passed. |
| `34-reaper-postcheck` | `s7-20260509T133438Z-34-reaper-postcheck` | delayed Docker/reaper/scope post-check | 2026-05-09T13:34:38Z | 2026-05-09T13:34:47Z | 0 | No managed testservices containers or signal-child processes remained. |
| `35-reader-validation` | `s7-20260509T133546Z-35-reader-validation` | explicit `explain-run` and `fixture-report` readers | 2026-05-09T13:35:46Z | 2026-05-09T13:35:47Z | 0 | Reader validation succeeded for selected success/failure runs. |

## Closure matrix

| Priority | Gap / source limit | Closure mechanism | Toolchain command or decision needed | Evidence to capture | Output file/register to update | Blocks final `MUST` language? | Validation |
|---|---|---|---|---|---|---|---|
| P0 | Evidence identity and source-limit preservation: `SL-0004`, `SL-0009`, `SL-0010`, `RTR-0016`, `PRES-0017`, `AUTH-0013` | `S7-DEC-0002` plus selected reader run | Decision plus `35-reader-validation` | Explicit `RESULTS_DIR`, `RUN_ID`, artifact roots, commands, timestamps, platform profiles, exit codes, reader output. | This register; `source-limit-log.md`; `harness-authority-map.md`. | Yes. | Closed for this pass by selected-run rule. Any artifact lacking selected identity remains source-limited. |
| P0 | S7 source-limit carry-forward | `S7-DEC-0003` | Decision-only; no final NLSpec promotion in this step. | Decision ID/date/source and affected rows. | This register; `s7-s6-audit-gap-follow-up.md`; `source-limit-log.md`. | Yes. | Closed as a carry-forward rule only; unresolved rows remain labeled. |
| P1 | Tool/platform readiness: `SL-0013`, `AUTH-0008`, `PRES-0015` | Runtime evidence | `01-profile`: `make doctor` | Doctor output and platform profile. | This register. | Yes. | Exit 0 closes current-host readiness only; platform matrix remains maintainer-owned. |
| P1 | Broad fast verification: `SL-0002`, `SL-0006` | Runtime evidence | `02-test-fast`: `make test-fast` | Run summaries, child artifacts, logs. | This register. | Yes. | Exit 0 supports current-host fast-loop readiness. |
| P1 | Full corpus verification: `SL-0002`, `SL-0006`, `SL-0012` | Runtime evidence | `03-test`: `make test`; reader in `35-reader-validation` | Full run summary, service-backed and browser artifacts. | This register. | Yes. | Exit 0 and reader output support current-host full-corpus readiness. |
| P1 | Developer gate and scheduler behavior: `SL-0002`, `SL-0006`, `RTR-0001` | Runtime evidence | `04-check`; `04-check-scheduler-drift` | Scheduler summaries/events, child summaries, drift readers. | This register. | Yes. | Exit 0 plus scheduler event/order and summary timing drift pass. |
| P1 | Provider-neutral CI: `SL-0001`, `AUTH-0015`, `PRES-0019` | Runtime evidence plus decision needed | `05-ci`: `make ci` | CI summary and failure diagnostics. | This register; `failure-mode-register.md`; `harness-authority-map.md`. | Yes. | Exit 2. Current failure is harness smoke/phase-map accounting. Provider workflow behavior remains source-limited while `.github/**` is absent. |
| P2 | Release gate and release artifacts: `FAIL-0026` | Runtime evidence or maintainer decision | `06-release-check`: `make release-check` | Release/check summaries and failure diagnostics. | This register; `failure-mode-register.md`. | Yes. | Exit 2 with same harness smoke/phase-map accounting failure as CI; no release readiness claim. |
| P1 | Docker/testcontainers image readiness: `SL-0012`, `RTR-0004` | Runtime evidence | `07-service-images`: `make test-service-images` | Image warm output and platform profile. | This register. | Yes. | Exit 0 supports current-host Docker image readiness only. |
| P1 | Service-backed backend DB/object/process behavior: `SL-0012`, `RTR-0004..RTR-0010` | Runtime evidence | `08`, `09`, `10` | Service events, leases, fixture summaries, child summaries, cleanup artifacts. | This register; cleanup register. | Yes. | Exit 0 for all selected runs; no managed containers remained after post-check. |
| P1 | Support-only integration surface | Runtime evidence | `11-backend-integration-support` | Support coverage and service-backed artifacts. | This register. | Yes for support harness claims only. | Exit 0; not authoritative product phase completion. |
| P1 | Phase-selected service-backed coverage | Runtime evidence | `12`, `13`, `14`, and `12-14-svc-phase-explain` | Phase summaries, service artifacts, child reader output. | This register. | Yes. | Exit 0 for each phase; child reader succeeded. |
| P1 | Browser shared-stack readiness: `SL-0012`, `RTR-0011..RTR-0013` | Runtime evidence | `15-browser-webserver`; `15-19-browser-port-release` | Stack env/json, logs, Playwright reports, `ss` port checks. | This register; cleanup register. | Yes. | Exit 0; recorded backend/frontend ports released where `ss` is available. |
| P1 | Browser stateful/reset-adjacent behavior | Runtime evidence plus `AUTH-0004` remains | `16-browser-stateful`; `17-browser-resettable` | Reset boundary files, status/state markers. | This register. | Yes. | Exit 0 proves harness reset flow only. Product reset API semantics remain maintainer-required. |
| P2 | Browser measurement and visual validation: `PRES-0018`, `AUTH-0014` | Runtime evidence plus decision | `18-browser-measurement`; `19-browser-visual` | Timing and visual validation artifacts. | This register; `harness-authority-map.md`. | Yes. | Exit 0 for validation. No snapshot update command was run. |
| P1 | Docker Compose/local-dev behavior: `AUTH-0005`, `PRES-0012`, `RTR-0019` | Runtime evidence plus maintainer decision | `20`, `21`, `22`, `23`, `23b` | Compose state, fixed-port checks, DB reset effects, dev interrupt logs. | This register; cleanup register. | Yes if adopted into harness contract. | Compose/db commands exited 0. Initial `make dev` failed on unwritable `/var/lib` roots; corrected interrupt returned 124 with ports released. |
| P1 | Baseline and scheduler artifact checks: `RTR-0003`, `TMR-0027` | Runtime evidence against selected runs | `24-baseline-drift`; `04-check-scheduler-drift` | Drift outputs tied to selected runs; no refresh command. | This register. | Yes. | All drift/coverage readers exited 0 against explicit selected runs. |
| P2 | Controlled usage/config failure schemas: `SL-0007`, `FAIL-0001..FAIL-0004` | Runtime evidence | `25-usage-fail`; `26-scheduler-config-fail` | Expected stderr and exit codes. | This register; `failure-mode-register.md`. | Yes. | Usage exits 2; invalid scheduler config exits 2 before child work. |
| P1 | Controlled service preflight failure: `FAIL-0005`, `SL-0013` | Runtime evidence | `27-docker-preflight-fail` | Missing Docker socket diagnostics and post-run Docker check. | This register; `failure-mode-register.md`. | Yes. | Exit 2 at Make surface. Valid preflight failure evidence; no managed containers remained. |
| P1 | Controlled browser startup failure: `FAIL-0009`, `FAIL-0021`, `RTR-0012` | Runtime evidence | `28b-browser-start-fail` | Build/startup failure diagnostics and `ss`/process post-check. | This register; cleanup register; `failure-mode-register.md`. | Yes. | Corrected command exits 2 before stack start. Browser process-start cleanup remains source-limited. |

## Failure and disposition notes

- `make ci` and `make release-check` failed at `run-harness-smoke-extended`
  because `harness-smoke-print-target-plan` reported phase5 authoritative-looking
  Go tests missing from the generated smoke phase map. No phase maps were edited.
- S7 maintainer decision `MD-S7-0016` demotes that stale extended smoke failure
  from blocking phase advancement. The selected failure remains evidence that
  the diagnostic target is stale, not evidence of product readiness or
  provider-specific CI behavior.
- `23-dev-int` demonstrated current local config failure for `make dev`:
  `/var/lib/cartulary/backups` was not writable. `23b-dev-int-writable` used
  disposable root overrides to reach the interrupt path and returned `124`.
- `28-browser-start-fail` is investigation-only because
  `CARTULARY_SERVER_BIN=/bin/false` was overwritten by the Make target export
  and the target passed. `28b-browser-start-fail` is the selected corrected
  failure, but it fails at `build-server` before owned-stack process startup.
- Baseline refresh targets, snapshot update commands, `make generate`,
  `make format`, `make clean`, `make distclean`, and `make agent-finalize` were
  not run during this pass.

## Execution order and safety outcome

Executed order:

1. Non-mutating discovery/profile commands: `00-discovery`, `01-profile`.
2. Broad verification: `02-test-fast`, `03-test`, `04-check`, `05-ci`,
   `06-release-check`.
3. Docker/testcontainers and service-backed evidence: `07` through `14`.
4. Browser evidence and port checks: `15` through `19`,
   `15-19-browser-port-release`.
5. Compose/local-dev evidence: `20` through `23b`.
6. Drift readers and controlled failures: `24` through `28b`.
7. Cleanup/signal scenarios: `29` through `34`.
8. Selected artifact readers: `35-reader-validation`.

Commands that were mutating or potentially destructive and did run:

- `make services-up`, `make db-up`, and `make db-reset`;
- service-backed, browser, Docker/testcontainers, and resettable browser runs;
- controlled signal/failure scenarios using `timeout`, missing `DOCKER_HOST`,
  and invalid scheduler/Make inputs.

Commands that were locally safe planning/discovery and did run:

- `make help-all`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`,
  `make target-plan-json`, `make explain-target`, `make explain-phase`,
  `make explain-run`, `make fixture-report`, and `make doctor`.

Commands deliberately not run:

- `make bootstrap`, `make frontend-install`, `make playwright-install`;
- `make generate`, `make format`, `make phase-ledgers`,
  `make phase-schedules`, `make agent-finalize`;
- all `*-duration-baselines` refresh targets;
- snapshot update commands, `make clean`, and `make distclean`.

Closure criteria used:

- `closed`: selected run has explicit result dir/run ID, timestamps, command,
  exit status, platform profile, expected artifacts, and validation output that
  satisfies the row for the current host.
- `preserved_source_limit`: evidence is absent, incomplete, host-specific for a
  broader claim, lacks selected identity, or shows only source/unit behavior for
  a live-runtime claim.
- `maintainer_decision_required`: behavior is public-contract-like,
  destructive, authority-sensitive, platform-wide, provider-specific, or outside
  runtime proof alone.

## Readiness checklist

- [x] Every runtime evidence claim in this pass cites a selected run ID and
      result directory.
- [x] Artifact claims use explicit selected runs; ambient newest-run evidence is
      not used for normative claims.
- [x] Runtime successes are limited to the observed WSL/Linux host profile.
- [x] CI and release readiness are not claimed because selected runs failed.
- [x] Cleanup claims distinguish command exit, cleanup status, delayed Docker
      state, empty reaper logs, and source-limited parent-death behavior.
- [x] Generated artifacts are treated as downstream execution inputs only.
- [x] No final S7 `MUST` language was written in this evidence pass.
