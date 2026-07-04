---
doc_id: THR-S7-CLEANUP-SIGNAL-EVIDENCE-REGISTER-2026-05-09
title: S7 Cleanup and Signal Evidence Register
status: active
role: cleanup-signal-evidence-register
---

# S7 Cleanup and Signal Evidence Register

## Document role

This register records cleanup, signal, process, service, and port-release
evidence gathered during the authorized S7 evidence pass on 2026-05-09. It is
paired with `runtime-evidence-register.md` and uses the same selected-run rule:
claims require explicit `CARTULARY_TEST_RESULTS_DIR` and
`CARTULARY_TEST_RUN_ID`.

This register does not turn source-observed cleanup paths into guaranteed
cleanup `MUST` language. It distinguishes observed completion, observed
scheduling, process/container after-state, source-limited behavior, and
maintainer-required authority.

## Common evidence locations

- Runtime root: `/home/askahn/code/cartulary/.cartulary/test-results/s7-recovery`
- Service leases and scopes: `<artifact-root>/_shared/test-services/<suite-id>/`
- Browser stack metadata: `<artifact-root>/<target>/owned-stack/stack.env` and
  `stack.json`
- Evidence wrapper logs: `<artifact-root>/_evidence/stdout.log`,
  `_evidence/stderr.log`, `_evidence/result.env`, and
  `_evidence/platform-profile.env`

## Scenario matrix

| Scenario | Command / setup | Required evidence captured | Expected exit handling | Validation |
|---|---|---|---|---|
| Normal exit | `29-normal-testservices`: `make test-service-images`; `tmp/toolbin/cartulary-test-services run -- bash -lc "exit 0"` | `service-scope.json`, `service-lease.json`, `service-reaper.log`, `goose.log`, Docker managed-container post-check, platform profile. | Exit `0`. | `service-scope.json` cleanup status is `succeeded`, child exit status is `0`, and delayed `34-reaper-postcheck` found no managed containers. Reaper log exists but is empty, so no separate reaper-log completion claim is made. |
| Abrupt child exit | `30-child-exit`: `tmp/toolbin/cartulary-test-services run -- bash -lc "exit 42"` | Child status, service scope, lease, reaper log, Docker managed-container post-check. | Expected nonzero; selected exit was `42`. | Parent propagated child status. `service-scope.json` cleanup status is `failed` with child exit status `42`; delayed Docker post-check found no managed containers. Treat as resource-release evidence, not successful cleanup-summary evidence. |
| Timeout | `31-timeout`: `timeout --signal=TERM --kill-after=20s 30s tmp/toolbin/cartulary-test-services run -- bash -lc "sleep 300"` | Timeout status, service scope, lease, reaper log, Docker managed-container state, process state. | Selected exit `124`. | Immediate post-command output still showed managed containers and a `terminate-suite` process; delayed `34-reaper-postcheck` found no managed containers and no matching child process. Cleanup completion is not promoted because service scope says `failed` and reaper logs are empty. |
| Interrupt | `32-interrupt`: `timeout --signal=INT --kill-after=20s 30s tmp/toolbin/cartulary-test-services run -- bash -lc "sleep 300"` | INT timeout status, service scope, lease, reaper log, Docker managed-container state, process state. | Selected exit `124`. | Same as timeout: immediate post-command state showed managed containers and termination process; delayed post-check found no managed containers and no matching child process. Guaranteed interrupt cleanup remains source-limited. |
| Parent death | Not executed as a live SIGKILL-parent scenario. | No selected runtime evidence. | Expected 137 if later executed. | Preserved as `source_limit`. A live parent-death claim still requires a controlled parent/child PID harness, stale janitor/reaper follow-up, and maintainer approval for any cleanup authority beyond owned test resources. |
| Active DB connection cleanup | `33-cleanup-unit-evidence`: focused `go test ./tools/testservices -run ... TestCleanupOwnedServicesFailsFastOnBrowserFixtureLeak ...` | Go test output for leak-detection cleanup path. | Exit `0`. | Unit-level runtime evidence supports the existence of a fail-fast leak detection path. Live active-DB cleanup behavior remains source-limited until a selected live DB fixture holds an active connection and records DB after-state. |
| Stale janitor behavior | `33-cleanup-unit-evidence`: `TestStaleWebE2EJanitorBoundsAndFiltersGeneratedFixtures` and `TestPreviousSuiteContainerCleanupEligibilityUsesCompletedSummaryOrAge`. | Go test output for generated-name/eligibility logic. | Exit `0`. | Supports bounded logic only. Destructive janitor authority remains `AUTH-0007` maintainer-required and cannot be expanded by unit evidence alone. |
| Detached reaper completion | Live service runs `29` through `32` plus delayed `34-reaper-postcheck`. | `service-reaper.log`, leases, Docker managed-container state, timestamps. | Command-specific exits: `0`, `42`, `124`, `124`; post-check exit `0`. | Delayed Docker state had no managed containers. Reaper logs were present but empty, so completion is supported only by Docker after-state, not by reaper-log proof. Do not promote detached reaper completion as guaranteed. |
| Process-group teardown | `23b-dev-int-writable` and `28b-browser-start-fail`. | Timeout/INT status, process listing, `ss` output, browser failure diagnostics. | Dev interrupt selected exit `124`; browser corrected failure selected exit `2`. | `23b` returned `124` after `make dev` reported Error 130 internally, with no listeners on ports 8080 or 5173. `28b` failed before browser stack startup; it validates build-prerequisite failure handling, not process-group teardown after backend start. |
| Shell traps | `23b-dev-int-writable` for dev stack INT; browser failure did not reach stack startup in `28b`. | Trap/signal output, server/frontend logs under `tmp/dev-stack`, `ss` checks, process listing. | Signal-derived status accepted. | Dev INT path observed no fixed-port listeners after timeout. Scratch/runtime-dir cleanup is not fully claimed because this run used local dev logs and disposable root overrides, not a separate shell scratch self-test. |
| Port release | `15-19-browser-port-release`, `23b-dev-int-writable`, `28b-browser-start-fail`. | Browser stack env files, recorded backend/frontend ports, `ss -ltnp` post-checks, dev fixed-port checks. | Command-specific. | Browser selected ports from runs `15` through `19` were absent from post-run `ss` output. Dev ports 8080/5173 were absent after INT run. If `ss` is unavailable on another host, port-release claims remain platform-limited. |

## Selected cleanup run details

| Evidence ID | Run ID | Exit | Cleanup summary | Leases/events/logs | DB/bucket/container/process state |
|---|---|---|---|---|---|
| `CLEAN-0001` | `s7-20260509T133255Z-29-normal-testservices` | 0 | `cleanup.status=succeeded`; `child_exit_status=0`. | Lease/scope/reaper/goose files under `_shared/test-services/f8d8ac6fc829f5599af188c3`. | Immediate Docker check still listed two managed containers briefly; delayed `34` check found none. |
| `CLEAN-0002` | `s7-20260509T133258Z-30-child-exit` | 42 | `cleanup.status=failed`; `child_exit_status=42`. | Lease/scope/reaper/goose files under `_shared/test-services/4a6ce3ef267e8ff6d740456c`. | Parent propagated status; delayed Docker check found no managed containers. |
| `CLEAN-0003` | `s7-20260509T133301Z-31-timeout` | 124 | `cleanup.status=failed`; `child_exit_status=-1`. | Lease/scope/reaper/goose files under `_shared/test-services/3ff03e0f8ce0b072fa13442c`. | Immediate output showed containers and `terminate-suite`; delayed check found no managed containers or child process. |
| `CLEAN-0004` | `s7-20260509T133330Z-32-interrupt` | 124 | `cleanup.status=failed`; `child_exit_status=-1`. | Lease/scope/reaper/goose files under `_shared/test-services/40ebc28fbbcce6ff939121d2`. | Immediate output showed containers and `terminate-suite`; delayed check found no managed containers or child process. |
| `CLEAN-0005` | `s7-20260509T133438Z-34-reaper-postcheck` | 0 | Post-check only. | Read service scopes and reaper logs for runs `29` through `32`. | `docker ps --filter label=cartulary.test-services.managed=true` printed no managed containers; process scan found no matching testservices/sleep child. |
| `CLEAN-0006` | `s7-20260509T132623Z-15-19-browser-port-release` | 0 | Browser port post-check only. | Uses selected browser stack env files from runs `15` through `19`. | `ss` found no listeners on recorded browser ports; managed testservices container check was empty. |
| `CLEAN-0007` | `s7-20260509T133010Z-23b-dev-int-writable` | 124 | Dev-stack INT with writable root overrides. | Backend/frontend logs in `tmp/dev-stack`; evidence stdout has `ss`/process output. | No listeners on fixed dev ports 8080/5173 after timeout; process scan showed only the evidence wrapper itself. |

## Preserved limits

- Parent-death cleanup remains `source_limit`; no live SIGKILL-parent scenario was
  executed.
- Active DB connection cleanup remains live-runtime `source_limit`; only focused
  unit evidence was run.
- Stale janitor deletion authority remains `AUTH-0007`
  `maintainer_decision_required`.
- Detached reaper completion is not promoted as a final guarantee because
  reaper logs were empty; delayed Docker after-state is recorded separately.
- Browser startup failure after backend/frontend process start remains
  source-limited. The corrected browser failure selected run failed at
  `build-server` before owned-stack startup.
- Port-release evidence is current-host and `ss`-dependent.
