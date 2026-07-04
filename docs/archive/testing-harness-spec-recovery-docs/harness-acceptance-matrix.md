---
doc_id: THR-S7-HARNESS-ACCEPTANCE-MATRIX
title: Testing Harness Acceptance Matrix
status: draft
role: harness-acceptance-matrix
---

# Testing Harness Acceptance Matrix

## Document role

This matrix binds S7 harness requirements to binary validation. Criteria marked
`owner_required` or `source_limited` block final normative claims, not S7
planning or source-limited drafting.

## Acceptance criteria

| Criterion ID | Requirement | Validation type | Validation command or evidence | Expected result | Evidence status |
|---|---|---|---|---|---|
| HAC-0001 | Make is the sole canonical harness command surface. | static inspection | `make help`, `make help-all`, `tools/task_surface_manifest.json`, `entrypoint-command-map.md` | Public harness entrypoints route through Make. Package scripts are not elevated to first-class harness contracts. | `maintainer_decision/source_observed` |
| HAC-0002 | Generated artifacts are downstream inputs only. | static inspection | `tools/generated_artifact_policy.json`; generated markers; `make generated-artifact-policy-check` | Generated Go/TS/task/schedule artifacts are fresh and not hand-edited. | `source_observed` |
| HAC-0003 | Phase 7/Phase 8 remain planned future work. | static inspection | `tools/phase_registry.json`; `source-limit-log.md` `SL-0005` | Missing phase7/phase8 manifests do not count as active coverage gaps. | `maintainer_decision/source_limit` |
| HAC-0004 | Reset route implementation is harness-owned and outside `internal/app`. | static inspection and Go tests | `rg "RegisterTestRuntimeResetRoute|testRuntimeResetSchemaID" internal/app internal/testutil cmd/server`; `go test ./internal/app ./internal/platform/bootstrap ./internal/testutil/testruntime ./cmd/server` | Reset implementation lives outside `internal/app`; startup bootstrap remains shared support; server wiring is minimal. | `source_observed` |
| HAC-0005 | Reset route is disabled by default. | integration test | `TestTestRuntimeResetRouteDisabledByDefault` | Default runtime returns 404 for `/api/v1/test/runtime/reset`. | `selected_runtime_observed` after test run |
| HAC-0006 | Test-route-enabled reset preserves schema and restores state. | integration test | `TestTestRuntimeResetRouteClearsStateAndRestoresBootstrap` | Response schema is `cartulary.test.runtime_reset.v1`; DB/object state is reset and bootstrap admin restored. | `selected_runtime_observed` after test run |
| HAC-0007 | Local-dev Compose and `make dev` are local verification behavior only. | static inspection | `harness-nlspec.md`; `environment-contract-observations.md`; `03-sprint-plan.md` | Docs classify local-dev behavior as verification setup, not deployment conformance. | `maintainer_decision/source_observed` |
| HAC-0008 | Scheduler lanes are logical constraints only. | static inspection | `harness-nlspec.md`; `harness-authority-map.md`; scheduler registry | Docs avoid host/service capacity guarantees. | `maintainer_decision/source_observed` |
| HAC-0009 | Stale janitors require generated name plus metadata or lease evidence and conservative age or completed summary. | unit/static test | Existing or added `tools/testservices` stale janitor tests | Janitors refuse unproven resources and accept only bounded generated fixtures. | `source_observed/selected_runtime_observed` after test run |
| HAC-0010 | Cleanup is best-effort unless selected evidence proves stronger behavior. | static inspection | `cleanup-signal-evidence-register.md`; `harness-nlspec.md`; `harness-acceptance-matrix.md` | Parent-death cleanup, active DB cleanup, and detached reaper completion remain source-limited. | `maintainer_decision/source_limit` |
| HAC-0011 | `/tmp/cartulary-go-*` is outside default cleanup scope. | static inspection | `Makefile`; `artifact-ownership-matrix.md`; `make -n clean`; `make -n distclean` | Default cleanup does not delete external Go cache paths. | `maintainer_decision/source_observed` |
| HAC-0012 | Environment docs list source-observed vars/defaults only and do not specify precedence. | static inspection | `environment-contract-observations.md`; `harness-nlspec.md`; `source-limit-log.md` `SL-0015` | Precedence remains `TODO: precedence_unknown` or source-limited. | `maintainer_decision/source_limit` |
| HAC-0013 | Platform docs preserve WSL/Linux as observed without claiming a full support matrix. | static inspection | `environment-contract-observations.md`; `harness-nlspec.md`; `ambiguity-register.md` `AMB-0028` | Non-Linux and missing-tool support remain source-limited. | `maintainer_decision/source_limit` |
| HAC-0014 | Durable retained-artifact claims require explicit run identity. | static inspection | `harness-nlspec.md`; `observable-interface-map.md`; `artifact-ownership-matrix.md` | Docs require `RESULTS_DIR`, `RUN_ID`, command, platform/tool profile, exit status, and paths. | `maintainer_decision/source_observed` |
| HAC-0015 | Fixture/golden refresh workflow is controlled and reviewable. | static inspection | `harness-nlspec.md`; `artifact-ownership-matrix.md`; `harness-review-packet.md` | Workflow names owner intent, file list, evidence/reason, verification command, and review note. | `maintainer_decision` |
| HAC-0016 | Visual snapshots remain validation-only. | static inspection | `harness-nlspec.md`; `structured-output-schema-notes.md`; snapshot rows | No snapshot update OS/browser/version/command is blessed. | `maintainer_decision/source_limit` |
| HAC-0017 | Unknown schemas remain unknown. | static inspection | `structured-output-schema-notes.md`; `harness-nlspec.md` | `partial`, `schema_unknown`, and `authority_unknown` rows are not promoted to stable contracts. | `source_observed/source_limit` |
| HAC-0018 | CI remains provider-neutral. | static inspection | `harness-nlspec.md`; `.github` absence; `scripts/ci/**` | Docs and code avoid provider workflow, annotation, and upload claims. | `maintainer_decision/source_limit` |
| HAC-0019 | Stale extended smoke is not blocking `ci` or `release-check`. | static inspection and dry run | `tools/execution_topology_manifest.json`; `tools/task_surface_manifest.json`; `scripts/test-run-make-sequence.sh`; `make -n ci`; `make -n release-check` | `run-harness-smoke-extended` is absent from blocking sequence steps and remains an explicit target. | `maintainer_decision/source_observed` |
| HAC-0020 | Final S7 `MUST` language is evidence-gated. | manual review | audit register, NLSpec, acceptance matrix | Every final normative claim maps to selected runtime evidence, source evidence, source limit, or owner decision. | `maintainer_decision_required` |

## Missing-test and future-gate list

| Gap ID | Gap | Required before final normative claim |
|---|---|---|
| HAC-GAP-0001 | Env precedence | Owner decision or override matrix. |
| HAC-GAP-0002 | Visual snapshot refresh platform/browser/version | Owner decision naming exact bounds and update command. |
| HAC-GAP-0003 | Parent-death cleanup | Controlled evidence or owner decision. |
| HAC-GAP-0004 | Active DB cleanup | Controlled evidence or owner decision. |
| HAC-GAP-0005 | CI provider annotations/uploads | Provider workflow source or owner decision. |
| HAC-GAP-0006 | Playwright report/trace/video/screenshot schemas | Tool-schema adoption decision or selected failure artifacts. |
| HAC-GAP-0007 | Release readiness beyond smoke demotion | Passing release evidence or explicit non-readiness classification. |

## Missing fixture/golden/snapshot review

This S7 review found no missing fixture, golden, or snapshot files inside the
source-bounded S7 package review. This is a reviewed absence statement for the
audited source set, not a claim that future harness work has complete fixture
coverage.

| Artifact class | S7 missing-item disposition | Evidence | Preserved limit |
|---|---|---|---|
| Fixtures | `none identified in the S7 source-bounded review` | `ART-0001`, `ART-0004`, `ART-0005` | Fixture update authority remains `AMB-0015`; committed fixtures change only through owner-reviewed source edits unless a later owner decision adds a supported refresh command. |
| Goldens | `none identified in the S7 source-bounded review` | `ART-0002`, `ART-0005` | Golden update authority remains `AMB-0015`; no generator or refresh command is adopted by this review. |
| Visual snapshots | `none identified as missing`; committed baselines exist | `ART-0003` | Visual snapshot refresh OS/browser/version/command remains `AMB-0022`, `AUTH-0014`, `PRES-0018`, `MD-S7-0013`, and `HAC-GAP-0002`. |
