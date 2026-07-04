---
doc_id: THR-S7-HARNESS-REVIEW-PACKET
title: Testing Harness S7 Review Packet
status: draft
role: review-packet
---

# Testing Harness S7 Review Packet

## Review purpose

This packet summarizes the S7 implementation state for maintainer review. It
does not close source limits by inference and does not claim release readiness
unless verification proves it.

## Implemented decisions

| Topic | Disposition | Evidence |
|---|---|---|
| Make command authority | Documented as sole canonical harness command surface. | `harness-nlspec.md`, `harness-acceptance-matrix.md`, `MD-S7-0001` |
| Generated artifacts | Documented as downstream execution inputs only. | `harness-nlspec.md`, `artifact-ownership-matrix.md`, `MD-S7-0003` |
| Reset route | Moved out of `internal/app` implementation ownership into harness test support, with bootstrap support extracted to `internal/platform/bootstrap`. | `internal/platform/bootstrap/**`, `internal/testutil/testruntime/**`, `cmd/server/main.go` |
| Local dev | Documented as normative local verification behavior, not deployment conformance. | `harness-nlspec.md`, `MD-S7-0005` |
| Scheduler lanes | Documented as logical scheduling constraints only. | `harness-nlspec.md`, `harness-authority-map.md`, `MD-S7-0006` |
| Stale janitor safety | Documented proof gates for destructive deletion. | `harness-nlspec.md`, `harness-acceptance-matrix.md`, `MD-S7-0007` |
| Cleanup strength | Documented best-effort default and preserved parent-death/active-DB source limits. | `harness-nlspec.md`, `cleanup-signal-evidence-register.md`, `MD-S7-0008` |
| External Go caches | Documented outside default cleanup. | `harness-nlspec.md`, `artifact-ownership-matrix.md`, `MD-S7-0009` |
| Env/defaults | Documented source-observed only; precedence remains unresolved. | `environment-contract-observations.md`, `harness-nlspec.md`, `MD-S7-0010` |
| WSL/Linux | Documented as primary observed environment without full platform matrix. | `harness-nlspec.md`, `MD-S7-0011` |
| Retained artifacts | Documented explicit identity requirements for durable claims. | `harness-nlspec.md`, `artifact-ownership-matrix.md`, `MD-S7-0012` |
| Fixture/golden/snapshot workflow | Documented controlled refresh workflow and validation-only visual snapshots. | `harness-nlspec.md`, `harness-acceptance-matrix.md`, `MD-S7-0013` |
| Unknown schemas | Preserved `partial`, `schema_unknown`, and `authority_unknown` blockers. | `structured-output-schema-notes.md`, `harness-acceptance-matrix.md`, `MD-S7-0014` |
| CI and stale smoke | Demoted `run-harness-smoke-extended` from blocking `ci` and `release-check`; kept provider-neutral CI. | `tools/execution_topology_manifest.json`, `tools/task_surface_manifest.json`, `scripts/test-run-make-sequence.sh`, `MD-S7-0015`, `MD-S7-0016` |
| Phase 6 browser login burst | Documented as approved harness-only browser test stress margin for eviction coverage; Core 04's session cap remains unchanged. | `apps/web/e2e/phase6.session-recovery.spec.ts`, commit `6f7ec1c`, `MD-S7-0018` |

## Deferred source limits and owner-required items

| Item | Status | Blocking scope |
|---|---|---|
| Env-var precedence | `source_limit/owner_required` | Blocks final precedence claims only. |
| Visual snapshot refresh OS/browser/version/command | `source_limit/owner_required` | Blocks snapshot update workflow claims. |
| Parent-death cleanup | `source_limit` | Blocks guaranteed abrupt-exit cleanup claims. |
| Active DB cleanup | `source_limit` | Blocks guaranteed live-connection cleanup claims. |
| Detached reaper completion | `source_limit` | Blocks hard cleanup completion claims. |
| CI provider annotations/uploads | `source_limit` | Blocks provider-specific CI claims. |
| Playwright report/trace/video/screenshot internals | `schema_unknown` | Blocks stable schema claims for tool-owned reports. |
| Release readiness beyond smoke demotion | `not_claimed` | Blocks release-ready claim unless `make release-check` passes or failures are classified. |

## Verification results

| Check | Result | Notes |
|---|---|---|
| `go test ./internal/app ./internal/platform/bootstrap ./internal/testutil/testruntime ./cmd/server` | pass | Reset route move and shared bootstrap support compile and pass package tests. |
| `go test ./tools/testservices -run 'TestStaleWebE2EJanitorBoundsAndFiltersGeneratedFixtures|TestPreviousSuiteContainerCleanupEligibilityUsesCompletedSummaryOrAge'` | pass | Stale janitor proof-gate coverage remains green. |
| `bash scripts/test-run-make-sequence.sh` | pass | CI/release sequence assertions now reject blocking extended smoke. |
| `make agent-finalize` | pass | First attempt failed at `json-shape-check` until the Phase 0 map was updated for moved bootstrap tests. Final run passed and skipped duration baseline refresh because `RESULTS_DIR` was unset. |
| `make generated-artifact-policy-check` | pass | Generated artifact policy boundaries remain valid. |
| `make json-shape-check` | pass | Phase map paths and JSON shapes are valid after moving bootstrap tests. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | pass | Public/check/helper task surface remains valid; `run-harness-smoke-extended` is helper-only diagnostic work. |
| `make test-fast` | pass | Fast and service-backed fast slices pass. |
| `make ci` | pass | Provider-neutral CI now runs `check` without blocking on stale extended smoke. |
| `make release-check` | pass | Release gate passes with `check`, release evidence, and build work; no extended smoke blocker. |
| `bash scripts/test-print-target-plan.sh` | fail, non-blocking | Still fails with the known stale Phase 5 smoke phase-map accounting issue. This is the demoted diagnostic problem tracked by `MD-S7-0016`. |

## Review questions not reopened

The maintainer decisions recorded in
`maintainer-decision-summary-2026-05-09.md` are treated as settled S7 inputs.
Review should focus on implementation consistency, source-limit preservation,
and whether any acceptance criterion needs stronger selected evidence before
final normative wording.

## Final handoff

S7 package status after this follow-up: the missing final handoff and explicit
fixture/golden/snapshot review gaps are documented in recovery docs. The
historical Phase 6 browser test burst change is documented and authorized by
`MD-S7-0018` as harness-only test logic.

Recovery-doc paths changed by this follow-up:

- `docs/testing-harness-spec-recovery-docs/harness-review-packet.md`
- `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`
- `docs/testing-harness-spec-recovery-docs/maintainer-decision-summary-2026-05-09.md`

No implementation, fixture, golden, snapshot, generated artifact, CI, cleanup,
lockfile, package-manager, runtime-service, or harness behavior path is changed
by this follow-up.

Existing verification summary only: this handoff relies on the verification
results already recorded in this review packet and the S7 audit. No new runtime
evidence was collected for the follow-up.

Preserved source limits and owner-required items:

| Item | Preserved status |
|---|---|
| Environment-variable precedence | `source_limit/owner_required`; do not specify precedence without owner decision or override matrix. |
| Visual snapshot refresh OS/browser/version/command | `source_limit/owner_required`; visual snapshots remain validation-only under `MD-S7-0013`, `AUTH-0014`, and `PRES-0018`. |
| Parent-death cleanup | `source_limit`; no guaranteed abrupt-exit cleanup claim. |
| Active DB cleanup | `source_limit`; no guaranteed live-connection cleanup claim. |
| Detached reaper completion | `source_limit`; no hard cleanup completion guarantee. |
| CI provider behavior | `source_limit`; provider-specific workflow, annotation, upload, and dashboard behavior remain out of scope while `.github/**` is absent. |
| Playwright report/trace/video/screenshot internals | `schema_unknown`; tool-owned report internals are not promoted to stable harness schemas. |
| Release readiness beyond recorded evidence | `not_claimed`; release readiness remains separate from stale-smoke demotion. |

Remaining open owner decisions:

| Decision needed | Current routing | Next action |
|---|---|---|
| Visual snapshot refresh bounds. | `source_limit/owner_required` through `AMB-0022`, `AUTH-0014`, `PRES-0018`, and `HAC-GAP-0002`. | Browser/harness owner should name exact OS, browser, version, and update command before any snapshot refresh workflow becomes normative. |
| Environment precedence. | `source_limit/owner_required` through `SL-0015` and `HAC-GAP-0001`. | Harness config owner should provide an override matrix or owner decision before final precedence wording. |

Historical implementation-change audit:

| Change area | Accounted status | Evidence and authority | Follow-up disposition |
|---|---|---|---|
| Reset-route movement and bootstrap extraction | Accounted and authorized. | `internal/platform/bootstrap/**`, `internal/testutil/testruntime/**`, `cmd/server/main.go`, `MD-S7-0004`. | Keep the reset route harness-owned, disabled by default, test-wired only, and not documented as a product API. |
| Stale extended smoke demotion | Accounted and authorized. | `tools/execution_topology_manifest.json`, generated task-surface artifacts, `scripts/test-run-make-sequence.sh`, `MD-S7-0015`, `MD-S7-0016`. | Keep `run-harness-smoke-extended` out of blocking `ci` and `release-check` paths while retaining provider-neutral CI bounds. |
| Generated task, schedule, and code artifacts | Accounted as downstream inputs. | Generated-artifact policy, generated markers, `MD-S7-0003`. | Do not treat generated outputs as behavior owners or hand-edit generated paths. |
| Phase 6 browser concurrency eviction login burst | Accounted and authorized as harness-only test logic. Commit `6f7ec1c` changed `apps/web/e2e/phase6.session-recovery.spec.ts` from literal `5` to `concurrencyEvictionLoginBurst = 10`. | Core 04 still owns the human-user concurrent-session cap at `5`; AC-159 still describes the `6th` concurrent login eviction behavior; `MD-S7-0018` approves the `10`-login burst only as browser-test stress margin. | No implementation remediation is needed for S7. Do not cite the `10`-login burst as product behavior or as a change to session-cap semantics. |

Exact next actions:

1. Treat the Phase 6 browser burst audit blocker as closed by `MD-S7-0018`.
2. Keep visual snapshot refresh bounds and environment precedence routed to
   owner-required future decisions.
3. Do not collect new runtime evidence unless a later follow-up explicitly
   authorizes it.
