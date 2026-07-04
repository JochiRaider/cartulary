---
doc_id: THR-S1-AUDIT-2026-05-08
title: S1 Inventory and Boundary Audit
status: complete
role: recovery-audit
---

# S1 Inventory and Boundary Audit

## Audit verdict

`pass_with_followups`

Sprint 1 is ready to serve as the basis for Sprint 2 entrypoint recovery. The
required S1 artifacts exist, the inventory is broad enough for the next sprint,
the boundary labels are explicit, high-impact gaps are recorded as source
limits or ambiguities, and the current working-tree changes remain limited to
recovery documentation.

The follow-ups are not S1 blockers: CI workflow provenance is unavailable
because `.github/**` is absent, runtime-only behavior remains intentionally
unobserved, artifact schemas and cleanup/service lifecycle rules are deferred to
later sprints, and several owner decisions remain open.

## Audit scope

| Scope item | Result | Evidence status | Evidence |
|---|---|---|---|
| Audit write surface | Audit artifact created under `docs/testing-harness-spec-recovery-docs/audits/`. | `observed` | This file. |
| S1 sprint status | S1 is marked `complete` with blocker `none`. | `observed` | `03-sprint-plan.md` S1 section. |
| Required S1 outputs | All required S1 output files exist. | `observed` | `harness-inventory.md`, `uninvoked-surface-list.md`, `embedded-harness-logic-list.md`, `ambiguity-register.md`, `source-limit-log.md`, S1 handoff. |
| Implementation edits | No implementation, test, fixture, CI, cleanup, generated, or lockfile paths were modified by this audit. | `runtime_observed` | `git status --short --branch` before and after audit discovery commands. |
| Product and harness behavior | Not changed and not re-specified. | `observed` | Audit non-scope and recovery charter rules. |
| Sprint 2 work | Not started. | `observed` | Audit reviewed S1 readiness only; no entrypoint command map was created. |

## Evidence reviewed

| Evidence area | Audit check | Result | Evidence status | Notes |
|---|---|---|---|---|
| Recovery controls | Checked charter, process rules, register templates, and prior audit format. | pass | `observed` | S1 artifacts follow the evidence-label and source-limit model from `01-recovery-process.md` and `04-registers-and-checklists.md`. |
| Inventory table shape | Checked `HI-0001` through `HI-0050` for role, status, owner hypothesis, boundary, evidence, and evidence status. | pass | `observed` | Rows cover entrypoints, orchestration, services, fixtures, generated artifacts, temp/log paths, cleanup, policy, adapters, derived views, and embedded logic. |
| Task and command surface | Compared S1 against Make help, exhaustive help, task guide, task-surface report, and backend target plan. | pass | `runtime_observed` | `task-surface-report --all` reported 77 public, 17 check-internal, and 28 helper-only targets, matching S1. |
| Phase manifests | Checked active phase explain output for phase0 through phase6. | pass | `runtime_observed` | Active phase manifests and ledgers exist for phase0 through phase6. Planned phase7/phase8 files are absent and recorded as source-limited. |
| Generated artifacts | Compared S1 against `tools/generated_artifact_policy.json`. | pass | `observed` | Policy names `internal/gen`, `packages/protocol-ts/src/generated`, and `tools/task_surface.generated.mk`; S1 records them without treating generated outputs as behavior owners. |
| Fixtures and snapshots | Cross-checked fixture, golden, and Playwright snapshot paths. | pass | `observed` | S1 inventories committed fixtures/goldens/snapshots and defers authority/update rules to S3. |
| Runtime artifacts and ignores | Checked `.gitignore`, Make cleanup paths, and ignored result roots. | pass | `observed/runtime_observed` | `.cartulary/test-results` is ignored via `test-results/`; Make names result, report, coverage, and Playwright cleanup paths. Artifact schema remains deferred. |
| Services and shared resources | Checked scheduler resources, service wrappers, testutil helpers, Playwright stack scripts, and service configs. | pass | `observed` | S1 covers Postgres, MinIO, browser stack, process, reset lanes, ports, runtime roots, DBs, buckets, and retained artifacts at inventory level. |
| Uninvoked surfaces | Checked `US-0001` through `US-0006`. | pass | `observed/source_limit` | S1 separately records absent CI, one apparently uninvoked script, placeholder/library-only surfaces, planned missing phases, and retained runtime artifacts. |
| Embedded harness logic | Checked `EHL-0001` through `EHL-0018`. | pass | `observed/runtime_observed` | S1 separates representative harness mechanics embedded in ordinary tests from product assertions. |
| Ambiguities and source limits | Checked `AMB-0001` through `AMB-0010` and `SL-0001` through `SL-0005`. | pass | `observed/source_limit` | Open rows have impact, owner, workaround, decision prompt, and evidence. |

## Commands run

All commands below were read-only discovery or inspection commands. They did
not run tests, start services, regenerate artifacts, format files, clean paths,
or begin Sprint 2 work.

| Command | Result | Evidence status | Notes |
|---|---|---|---|
| `git status --short --branch` | Exit 0; listed only S1 recovery-doc changes before this audit file was added. | `runtime_observed` | Re-run after discovery commands showed the same non-audit scope. |
| `git ls-files \| wc -l` | Exit 0; printed `746`. | `runtime_observed` | Matches S1 repository metadata. |
| `rg --files docs/testing-harness-spec-recovery-docs docs/spec docs/guides \| sort` | Exit 0. | `runtime_observed` | Confirmed recovery, spec, and guide documents exist. |
| `rg --files scripts tools internal/testutil apps/web/e2e apps/web/src packages cmd internal \| sort` | Exit 0. | `runtime_observed` | Cross-checked major script, tool, backend, frontend, and test-support surfaces. |
| `make help` | Exit 0. | `runtime_observed` | Confirmed compact public surface. |
| `make help-all` | Exit 0. | `runtime_observed` | Confirmed public task tiers including local dev, fast verification, full gates, investigation, maintenance, and release. |
| `make task-guide` | Exit 0. | `runtime_observed` | Confirmed role guidance and latest retained artifact references. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | Exit 0; check passed. | `runtime_observed` | Confirmed public/check-internal/helper-only target counts and logical harness checks. |
| `make target-plan` | Exit 0. | `runtime_observed` | Confirmed backend target families and service-backed classification. |
| `make target-plan-json` | Exit 0. | `runtime_observed` | Confirmed target-plan JSON row shape; output was not treated as runtime test evidence. |
| `make explain-phase PHASE=phase0` through `phase6` | Exit 0 for each phase. | `runtime_observed` | Confirmed active phase manifests, ledgers, dependencies, and service requirements. |
| `git check-ignore -v .cartulary/test-results test-results playwright-report coverage tmp \|\| true` | Exit 0. | `runtime_observed` | Matched `.cartulary/test-results`, `coverage`, and `tmp`; `.gitignore` and Makefile still name `test-results/` and `playwright-report/` policy surfaces. |
| Targeted `rg`, `jq`, and `test -e` checks | Exit 0. | `observed/runtime_observed` | Confirmed hidden dependency terms, generated roots, scheduler resources, `.github/**` absence, and planned phase7/phase8 absence. |

## Findings

| Finding ID | Finding | Severity | Evidence status | Evidence reference | Disposition |
|---|---|---|---|---|---|
| AUD-S1-0001 | S1 required artifacts are present and linked by the sprint plan or handoff. | none | `observed` | `03-sprint-plan.md`; S1 output files. | Pass. |
| AUD-S1-0002 | The inventory is complete enough for S2: it covers Make/task surface, package scripts, runner configs, scheduler scripts, phase manifests, generated policy, fixtures, services, runtime artifacts, cleanup, CI scripts, and embedded harness logic. | none | `observed/runtime_observed` | `harness-inventory.md` `HI-0004` through `HI-0050`; `task-surface-report --all`; file inventories. | Pass. |
| AUD-S1-0003 | Boundary classifications are explicit and cautious. Product behavior, ordinary product assertions, generated outputs, and build outputs are not promoted to harness normative behavior. | none | `observed` | `harness-inventory.md` boundary vocabulary and boundary answers; `embedded-harness-logic-list.md`. | Pass. |
| AUD-S1-0004 | High-impact absences and uncertain ownership are documented rather than guessed. | none | `observed/source_limit` | `SL-0001` through `SL-0005`; `AMB-0001` through `AMB-0010`. | Pass with follow-ups. |
| AUD-S1-0005 | S1 correctly identifies `.github/**` as absent while retaining `scripts/ci/**` and `make ci` as provider-neutral CI evidence for S2. | low | `source_limit/observed` | `HI-0003`, `HI-0050`, `US-0001`, `AMB-0001`, `SL-0001`. | Follow-up in S2; not a blocker to starting S2. |
| AUD-S1-0006 | Runtime-only artifacts, service lifecycle, cleanup idempotency, failure bundles, and report schemas remain unobserved but are explicitly deferred to the intended later sprints. | low | `observed/source_limit` | `SL-0002` through `SL-0004`; `AMB-0007`, `AMB-0008`, `AMB-0010`; S1 handoff. | Follow-up in S3/S4/S5/S6. |
| AUD-S1-0007 | The current changed-file scope remains recovery-doc-only. | none | `runtime_observed` | `git status --short --branch`. | Pass. |

## Blocking issues

No blocking S1 issues were found.

S2 may begin from:

- `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`
- `tools/task_surface_manifest.json`
- `tools/task_surface.generated.mk`
- `harness-inventory.md`
- `uninvoked-surface-list.md`
- `embedded-harness-logic-list.md`
- `ambiguity-register.md`
- `source-limit-log.md`

## Follow-up issues

| Follow-up ID | Issue | Target sprint | Why non-blocking for S2 |
|---|---|---|---|
| AUD-S1-FU-0001 | Decide whether CI is external, absent, or represented by `scripts/ci/**` plus `make ci`. | S2 | S1 records `.github/**` absence as `SL-0001` and `AMB-0001`; S2 can still trace local/provider-neutral entrypoints. |
| AUD-S1-FU-0002 | Recover artifact authority, freshness, machine-consumed schemas, and cleanup for `.cartulary/test-results/**`, reports, traces, screenshots, coverage, and Playwright output roots. | S3/S5 | S1 inventories the roots and source-limits runtime-only details. |
| AUD-S1-FU-0003 | Recover service lifecycle and resource allocation across Make, schedulers, `tools/testservices`, `pgtest`, `s3test`, Playwright stack scripts, Docker, ports, DBs, buckets, and process slots. | S4/S6 | S1 inventories the surfaces and shared resources without guessing runtime ordering. |
| AUD-S1-FU-0004 | Decide whether `internal/app/test_runtime_reset.go` belongs in the recovered harness spec and under what authority/visibility boundary. | S3/S4/S8 | S1 records it as `owner_decision_required` and does not normalize it into harness contract. |
| AUD-S1-FU-0005 | Decide how planned phase7/phase8 registry entries should be treated before later NLSpec or coverage work. | S7 or owner decision | S1 records absent files as planned/source-limited, not active missing coverage. |
| AUD-S1-FU-0006 | Split generated-artifact execution impact from generated-output authority. | S3/authority pass | S1 inventories generated committed outputs and points to `tools/generated_artifact_policy.json`. |

## Validation checklist

| Check | Result | Notes |
|---|---|---|
| Confirm S1 status is `complete` in `03-sprint-plan.md`. | pass | S1 status and exit checklist are complete. |
| Confirm required S1 artifacts exist. | pass | All expected files are present. |
| Confirm S1 artifacts use evidence labels from recovery process docs. | pass | Rows use `observed`, `runtime_observed`, `source_limit`, and combined observed/source-limit statuses. |
| Cross-check inventory against Make/task-surface/package/script/test-runner surfaces. | pass | Task-surface and file inventory checks support S1 coverage. |
| Cross-check fixtures, snapshots, goldens, generated roots, temp paths, reports, logs, and cleanup surfaces. | pass | Covered in `HI-0024` through `HI-0049`; authority/lifecycle intentionally deferred. |
| Cross-check service and shared-resource surfaces. | pass | Scheduler resources and service wrappers are inventoried. |
| Confirm uninvoked surfaces are listed separately. | pass | `US-0001` through `US-0006`. |
| Confirm embedded harness logic is listed separately. | pass | `EHL-0001` through `EHL-0018`. |
| Confirm `.github/**` absence is logged as source limit and ambiguity. | pass | `SL-0001`, `AMB-0001`, `HI-0003`, `US-0001`. |
| Confirm planned phase7/phase8 absence is logged as planned/source-limited. | pass | `SL-0005`, `AMB-0004`, `US-0005`, `HI-0019`. |
| Confirm retained runtime artifacts are not treated as authoritative current evidence. | pass | `SL-0004`, `AMB-0002`, `AMB-0010`, `HI-0044`. |
| Confirm unsupported assumptions are either removed, downgraded, or recorded as ambiguity/source-limit. | pass | No unsupported blocker found; open uncertainty is registered. |
| Confirm no implementation, fixture, generated, CI, service, or cleanup files were modified. | pass | Git status remained recovery-doc-only before this audit file was added. |
| Assign audit verdict. | pass | Verdict is `pass_with_followups`. |
| If not blocked, state S2 starting evidence. | pass | Listed above. |

## Implementation-change audit

| Check | Result |
|---|---|
| Harness implementation files modified | `no` |
| Test logic modified | `no` |
| CI behavior modified | `no` |
| Fixture contents modified | `no` |
| Cleanup scripts modified | `no` |
| Services modified | `no` |
| Generated code modified | `no` |
| Lockfiles modified | `no` |
| Sprint 2 work started | `no` |
| Only recovery docs changed | `yes` |

## Final audit note

S1 is complete enough to support the next recovery step. Begin S2 by tracing
entrypoints from the task surface outward, while preserving S1's documented
limits: CI workflow files are absent, live runtime behavior was not exercised,
artifact schemas and cleanup/service lifecycle are later-sprint work, and
product assertions must remain separate from harness mechanics.
