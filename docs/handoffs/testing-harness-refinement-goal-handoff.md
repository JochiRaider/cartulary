---
title: Testing Harness Refinement Goal Handoff
status: handoff
document_class: execution-handoff
handoff_type: codex-goal
created_at: 2026-06-08T23:11:30Z
source_runs:
  cold: 20260608T224321Z-p17291
  warm: 20260608T224533Z-p79693
source_diagnostic: current-thread harness timing table
---

# Testing Harness Refinement Goal Handoff

## 1. Purpose

This handoff converts the cold/warm `make check` diagnosis into an execution handoff for a future local coding goal to refine the Cartulary testing harness. The cold run is `.cartulary/test-results/20260608T224321Z-p17291`; the warm run is `.cartulary/test-results/20260608T224533Z-p79693`.

This handoff is not product conformance evidence, not benchmark evidence, not release evidence, and not a substitute for `docs/testing-harness-nlspec.md`. It is a goal-run contract for reducing avoidable local-check latency while preserving the harness authority boundary.

## 2. Document Contract and Normative Language

`document_class` is `execution-handoff`. This document is not a behavior NLSpec and MUST NOT redefine harness mechanics already owned by `docs/testing-harness-nlspec.md`.

Inside this handoff, uppercase **MUST**, **MUST NOT**, and **MAY** apply only to future goal execution and handoff-local mechanics. A handoff-owned **MAY** is valid only when the same paragraph or table row defines omission behavior. Omission behavior: when this handoff omits a harness behavior detail, the future goal runner MUST use the controlling owner document and MUST NOT invent a local rule.

Requirements imported from owner documents keep their original owner and meaning. If this handoff and an owner conflict, the owner governs and the future goal MUST stop with `owner_drift_conflict`.

## 3. Boundary-Completeness Rule

Rows in Sections 11 through 13 are required execution rows. A required execution row MUST NOT use an open delegation phrase unless the same row cites a closed owner rule, an imported default row from Section 5, or a decision table in this handoff that closes the behavior for the future goal runner.

| Controlled open-phrase scan seed | Allowed appearance |
| --- | --- |
| `accepted budget` | Only as this scan seed. Use `IDB-001` values instead. |
| `accepted balance ratio` | Only as this scan seed. Use `IDB-001` values instead. |
| `equivalent` | Only inside closed owner tokens such as `full_target_equivalent` or inside this scan seed. |
| `owner-approved` | Only as this scan seed. Use `owner_closure_required` when owner approval is missing. |
| `justified` | Only as this scan seed. Use closed reason-code fields instead. |
| `safe` | Only as this scan seed. Use the fixture-policy matrix and invalid-conversion rules instead. |
| `narrower` | Only as this scan seed. Name the exact target or exact owner rule instead. |
| `if available` | Only as this scan seed. Name the stop behavior for absence instead. |

When a future goal runner finds any other occurrence of a scan seed in this handoff, the runner MUST treat it as `owner_drift_conflict` unless the occurrence is in the same table row as the closed owner import or decision table that resolves it.

## 4. Authority and Ownership Matrix

| Contract family | Behavior owner | Handoff-owned consequence | Forbidden handoff action | Drift handling |
| --- | --- | --- | --- | --- |
| Harness mechanics | `docs/testing-harness-nlspec.md`, especially Sections 1, 4, 8, 10, 11, 12, 13, and 17 | Require future edits to preserve Make invocation, scheduling, fixture lifecycle, service ownership, artifact emission, cleanup, cache mechanics, and verification gates. | Redefine target semantics, scheduler accounting, cache-hit semantics, fixture policy, service lifecycle, cleanup, or summary schemas. | Mark the affected row `owner_drift_conflict` and stop. |
| Product behavior | Core 00 through Core 04 | Preserve existing product-conformance coverage while moving or decomposing harness work. | Change product semantics or weaken product-conformance evidence to reduce runtime. | Stop and require product-owner change before continuing. |
| Publication or benchmark behavior | Core 05 only | Prevent timing observations from becoming benchmark or publication claims. | Treat local run timings as claim-bearing performance evidence. | Stop on publication-boundary conflict. |
| Phase maps and target manifests | Owner specs plus `tools/phase*_test_map.json`, `tools/frontend_phase_maps/*.json`, `tools/task_surface_manifest.json`, and generated ledgers as downstream materialization | Require any default-check movement or slice reshaping to be reflected through owner inputs and generated drift targets. | Hand-edit generated schedules or move conformance rows out of default check by handoff decision. | Mark owner or repo mismatch `owner_drift_conflict`. |
| Cache mechanics | `docs/testing-harness-nlspec.md` Sections 1 and 8, including `cartulary.cache.readiness.v1` and `cartulary.cache.build_artifact.v1` | Allow local acceleration only when public summaries, failure classification, service lifecycle, cleanup, runtime reset, drift/security checks, and aggregate verdicts still execute or emit. | Mark correctness work as scheduler `reused` or skip required wrapper behavior because a cache hit exists. | Stop and repair cache modeling before relying on evidence. |
| Retained run evidence | Live retained run roots under `.cartulary/test-results` | Use cold/warm artifacts as run-analysis evidence only. | Treat the diagnostic table or retained timing artifacts as behavior owners. | Apply Section 7 retained-evidence decisions. |
| Handoff-local execution mechanics | This document | Own registry IDs, closure states, stop conditions, slice order, prompt text, and final reporting. | Create new harness behavior by handoff-local rule. | Revise this handoff before continuing. |

## 5. Imported Defaults and Bounds

| ID | Owner import | Closed value or rule | Omission, edge, and failure behavior | Handoff use |
| --- | --- | --- | --- | --- |
| `IDB-001` | Harness TH-HARNESS-REQ-057, Section 11.6, TH-HARNESS-AC-018 | Warm `check-service-backed` hard cap is `240000ms`; remediation target is `180000ms`; non-isolated backend/browser peer lane balance ratio is `1.25`; materiality floor is `5000ms`. | Omitted `SCHEDULER_WARM_CHECK_BUDGET_MS` and `SCHEDULER_WARM_CHECK_BALANCE_RATIO` use harness defaults. This handoff's remediation command supplies `SCHEDULER_WARM_CHECK_BUDGET_MS=180000` and `SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25`. A run that is not warm-ready cannot close warm scheduler health. | All timing acceptance and lane-balance rows. |
| `IDB-002` | Harness Sections 1 and 8; TH-HARNESS-AC-028 | Cache families are limited to `cartulary.cache.readiness.v1` and `cartulary.cache.build_artifact.v1` for this handoff's build/readiness work. | Missing, disabled, forced, invalid, corrupt, or missing-output cache states MUST execute the underlying command or fail as owner-defined `configuration_error` or `artifact_error`. They MUST NOT produce success by reuse. Public summaries, validation, cleanup, service lifecycle, runtime reset, drift/security checks, and aggregate verdicts still execute or emit. | `HRF-002`, `HRF-006`, `HRF-007`, `HRF-009`, `HRG-002`, `HRG-004`, `HRG-005`. |
| `IDB-003` | Harness Section 8; TH-HARNESS-AC-031 | `pressure-summary.json` is a required retained diagnostic artifact, not schema-owned conformance evidence in the current profile. `reused_accounting_counts` and `readiness_attribution_counts` MUST be present. | `{}` for either accounting field means no attribution was emitted by the current scheduler. The marker `cartulary.scheduler_pressure_summary.v1` MUST NOT be treated as a present schema attachment unless an owner revision adopts the schema and fixtures. | `HRF-009`, `HRG-006`. |
| `IDB-004` | Harness Section 4; TH-HARNESS-AC-027 | Default-check metadata fields are `default_check_required`, `default_check_kind`, `default_check_reason_code`, `primary_evidence_owner`, `duplicate_of`, `evidence_delta`, and `warm_local_cost_class`; non-obvious inclusion also requires `default_check_reason`. Projection rows require `check_projection` with `mode`, `schedule`, `stage`, `evidence`, `evidence_class`, `reason_code`, `full_target`, and `full_target_equivalent`. | `default_check_required=true` is valid only for implemented executable current-profile rows. Planned or blocked future rows use `future_default_check_candidate=true`. Projection mode MUST NOT also advertise direct `check` membership. | `HRF-001`, `HRF-003`, `HRF-008`, `HRG-001`, `HRG-007`, `HRG-008`. |
| `IDB-005` | Harness Section 8.2; TH-HARNESS-AC-019 | `agent-finalize RESULTS_DIR=<dir>` accepts only an existing retained full warm `make check` run root with passing full-check markers and required scheduler artifacts. | Missing, failed, incomplete, contaminated, service-backed-only, phase-slice, browser-only, or non-warm run roots fail before mutating refresh work. | `HR-9` final handoff closure and retained-run maintenance. |
| `IDB-006` | Harness Section 11; TH-HARNESS-AC-018 | Closed Postgres fixture reason codes are `committed_cross_connection_visibility`, `database_identity`, `process_lifecycle`, `schema_mutation`, `destructive_residue`, `shared_seeded_state`, `bounded_reset_surface`, and `migration_scratch`. | Explicit `template_clone`, explicit `group_clone`, explicit broad `package_reset`, and `migration_scratch` rows require both human reason and closed reason-code field. Fixture overuse remains non-final until source details and budgets reconcile. | `HRF-005`, `HRG-003`, `HRG-009`. |
| `IDB-007` | Harness Sections 4, 5, 8, and 10; TH-HARNESS-AC-037 | `build-operator` is a public build target and permitted `cartulary.cache.build_artifact.v1` profile. It accepts public `OPERATOR_BIN`; scheduler-selected operator Go work consumes only scheduler-owned `CARTULARY_OPERATOR_BIN` from the `operator` runtime-binary registry. | Default local `check` builds the operator only when selected runtime-binary work declares it. `OPERATOR_BIN` is not child-forwarded, and public command-line `CARTULARY_OPERATOR_BIN` is an internal override error. | `HRF-002`, `HRG-002`. |

## 6. Closure-State Model

This handoff uses one closure-state vocabulary across finding rows, boundary-gap rows, execution slices, stop conditions, acceptance criteria, and the Pursue Goal prompt.

| Closure state | Meaning | Final-allowed? | Required handling |
| --- | --- | ---: | --- |
| `owner_closed` | The controlling owner document defines the rule enough for implementation and verification. | Yes | Import the owner section and verify against live repo facts. |
| `repo_materialization_required` | The owner rule is closed enough, but repository edits or generated artifacts are still required. | No | Implement or regenerate the required repo surface and record evidence. |
| `repo_materialized` | The required repo surface has been inspected or created, exact validation passed, and evidence is cited. | Yes | Cite the path or retained artifact and close the linked row. |
| `owner_closure_required` | The controlling owner does not define a needed rule. | No | Stop, report the exact missing owner rule, and do not choose a provisional behavior. |
| `handoff_local_closed` | The item concerns only this handoff's goal-run mechanics. | Yes | Apply the local rule without turning it into harness behavior. |
| `owner_drift_conflict` | Live owners or repo-control artifacts conflict with this handoff. | No | Stop and revise the handoff or owner before implementation continues. |

Final harness-refinement completion is unavailable while any finding, decision, boundary-gap row, or execution slice has `repo_materialization_required`, `owner_closure_required`, or `owner_drift_conflict`.

### 6.1 Closure-State Lifecycle

| From state | Event | To state | Required evidence | Illegal transition |
| --- | --- | --- | --- | --- |
| `owner_closed` | Live repo fact shows implementation work is still needed. | `repo_materialization_required` | Owner section, current path or artifact, and linked finding/gap row. | None. |
| `owner_closed` | Live repo fact already satisfies the owner rule. | `repo_materialized` | Exact retained artifact or source path plus validation command result. | Claiming satisfaction without path or retained artifact evidence. |
| `repo_materialization_required` | Required source edits or owner-input edits are made, generated artifacts are regenerated by public Make targets when needed, exact validation command passes, and retained artifact/path is cited. | `repo_materialized` | Changed paths, generated drift target result, validation command, and retained artifact/path. | Moving to `repo_materialized` without exact validation or with hand-edited generated outputs. |
| Any non-final state | A required owner rule is missing. | `owner_closure_required` | `TODO(owner-closure): <specific missing owner rule>` and linked row IDs. | Choosing provisional harness behavior by handoff-local rule. |
| `owner_closure_required` | Owner document revision closes the missing rule and this handoff row is updated to import it. | `repo_materialization_required` | Revised owner section, imported default row if needed, and validation command. | Moving directly to `repo_materialized`. |
| Any state | Live owner or repo-control artifact conflicts with this handoff. | `owner_drift_conflict` | Conflicting owner path, conflicting row or rule, and affected row IDs. | Continuing implementation after conflict is known. |
| Any state | Item is only handoff-local and no repo behavior is required. | `handoff_local_closed` | Handoff-local rule and no linked harness behavior. | Using `handoff_local_closed` for harness mechanics. |

## 7. Retained Evidence Decision Table

| Retained evidence condition | Required handoff action | Closure consequence |
| --- | --- | --- |
| Named cold or warm run root exists and the required artifact exists. | Cite the artifact path and retain the observed fact in Section 9. | Linked rows may proceed according to their current closure state. |
| Named run root exists but a required artifact for a row is missing. | Add `TODO(repo-inspection): missing <artifact> for <row-id>` and name the affected run root. | Linked rows remain non-final. |
| Named run root is missing. | Add `TODO(repo-inspection): missing retained run root <path>` for each affected evidence row. | Linked rows remain non-final; newest-run fallback cannot close them. |
| A newer retained run exists. | Use only for human investigation unless the handoff is updated to name it and the run satisfies owner evidence identity rules. | No closure effect by itself. |
| Retained run is failed, partial, service-backed-only, phase-slice-only, browser-only, contaminated, or non-warm. | Cite only as diagnostic evidence. | It cannot satisfy warm scheduler maintenance, `agent-finalize RESULTS_DIR`, or final completion. |
| Artifact shape conflicts with the owner spec or required schema policy. | Mark affected rows `owner_drift_conflict` or `owner_closure_required` according to whether the owner conflicts or is missing. | Future goal stops before implementation continues. |

## 8. Source Limits and Evidence Inventory

Evidence stores MUST remain separate. The future goal runner MUST NOT silently merge diagnostic-table observations, retained run facts, owner specs, generated manifests, and live source inspection.

| Evidence store | Inspected source | Use in this handoff | Source limit |
| --- | --- | --- | --- |
| Current-thread diagnostic table | `source_diagnostic: current-thread harness timing table` | Required finding set `HRF-001..HRF-009`. | Not a behavior owner; re-check live artifacts before editing. |
| Retained cold run root | `.cartulary/test-results/20260608T224321Z-p17291` | Cold-start timings, cache misses, scheduler and fixture pressure. | Retained artifact evidence only; Section 7 governs absence or drift. |
| Retained warm run root | `.cartulary/test-results/20260608T224533Z-p79693` | Warm timings, cache hits, critical path, persistent bottlenecks. | Retained artifact evidence only; Section 7 governs absence or drift. |
| Repository specs | `docs/testing-harness-nlspec.md`, Core 00 through Core 05 | Authority boundaries, cache rules, default-check rules, fixture policy, warm scheduler health, acceptance criteria. | Owner text controls over this handoff. |
| Repository manifests | `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, `tools/scheduler_manifest.json`, `tools/phase10_test_map.json`, `docs/testing/phase10_coverage_ledger.md`, `tools/service_backed_make_target_duration_baselines.json` | Target structure, phase-map ownership, scheduler topology, default-check metadata. | Generated or mirrored artifacts must be updated through owner inputs and drift targets. |
| Generated artifacts | `tools/execution_topology_manifest.json`, generated ledgers and schedules where applicable | Downstream materialization to validate after owner-input edits. | Do not hand-edit generated outputs. |
| Live source files | `cmd/operator/operator_phase10_test.go`, `internal/testutil/pgtest/pgtest.go`, scheduler scripts, cache scripts, frontend/build scripts | Expected edit surfaces for future implementation. | Re-inspect before editing because source may have changed. |

Key inspected retained artifacts include `run-summary.json`, `check/target-summary.json`, `check/scheduler-summary.json`, `check/scheduler-events.jsonl`, `check/pressure-summary.json`, `check-service-backed/target-summary.json`, `backend-process/target-timing.json`, `backend-process/target-summary.json`, `_shared/backend-process-phase10-operator-inspection/runner.jsonl`, frontend/build target summaries, cache records, and `_shared/test-services/*/service-scope.json`.

No dedicated markdown or handoff validator was found in the inspected Make surface. The narrow validation for this docs-only handoff is therefore `make generated-artifact-policy-check` plus `make json-shape-check`.

## 9. Evidence Ledger

| Evidence ID | Path | Cold observation | Warm observation | Relevant findings |
| --- | --- | --- | --- | --- |
| `EVID-CHECK-TOTAL` | `*/run-summary.json`, `*/check/target-summary.json` | `check` pass, wall `120664ms`, executed `637924ms`, reused `0ms`, derived `7014ms`. | `check` pass, wall `107381ms`, executed `647299ms`, reused `0ms`, derived `7156ms`. | `HRF-001`, `HRF-004`, `HRF-009` |
| `EVID-CRITICAL-PATH` | `*/check/scheduler-summary.json`, `*/check/scheduler-events.jsonl` | Critical path includes `check-frontend-install 7443ms`, `testservices-build 3611ms`, service session `6030ms`, `backend-process 96788ms`. | Critical path includes `check-frontend-install 362ms`, `testservices-build 539ms`, service session `4714ms`, `backend-process 96157ms`. | `HRF-001`, `HRF-004`, `HRF-007`, `HRF-009` |
| `EVID-BACKEND-PROCESS` | `*/backend-process/target-timing.json` | `backend-process` wall `87409ms`; phase10 operator inspection `67336ms`. | `backend-process` wall `89389ms`; phase10 operator inspection `70038ms`. | `HRF-001` |
| `EVID-E-10-01-TESTS` | `*/_shared/backend-process-phase10-operator-inspection/runner.jsonl` | Four serial tests: `4.74s`, `6.94s`, `10.38s`, `43.92s`; package elapsed `66.025s`. | Four serial tests: `4.89s`, `7.03s`, `11.27s`, `45.27s`; package elapsed `68.495s`. | `HRF-001`, `HRF-003` |
| `EVID-HIDDEN-OPERATOR-BUILD` | `cmd/operator/operator_phase10_test.go` | Original gap evidence: the helper shelled out to `make --no-print-directory build-operator OPERATOR_BIN=<temp>` per test helper use. | Current closure expectation: canonical harness runs use injected `CARTULARY_OPERATOR_BIN`; the fallback build path is raw direct-Go support only. | `HRF-002` |
| `EVID-OBJECT-MIGRATION` | `cmd/operator/operator_phase10_test.go`, warm runner JSONL | Source starts two `s3test.StartOwnedWithLabels` harnesses and runs pass and mismatch migrations. | Single object-store migration test is `45.27s`. | `HRF-003` |
| `EVID-SERVICE-LANES` | `*/check-service-backed/target-summary.json`, `*/check/pressure-summary.json` | `check-service-backed` wall `102551ms`; backend-process child `87409ms`. | `check-service-backed` wall `100612ms`; backend-process child `89389ms`; other child walls `57907ms`, `38677ms`, `47110ms`, `45305ms`. | `HRF-004`, `HRF-008` |
| `EVID-FIXTURES` | `*/run-summary.json`, `_shared/test-services/*/service-scope.json` | Per-test template clone count `261`, duration `27089ms`; service-scope template clones `296`. | Per-test template clone count `261`, duration `25830ms`; service-scope template clones `296`. | `HRF-005` |
| `EVID-BUILD-CACHE` | Build summaries, `*/check/scheduler-events.jsonl`, `*/testservices-build/build-artifact-cache-testservices-build.json` | Scheduler durations: build-web `4375ms`, build-migrate `3892ms`, build-server `8926ms`; testservices-build miss `output_missing`. | Scheduler durations: build-web `2732ms`, build-migrate `1939ms`, build-server `6407ms`; individual build target summaries show only collation and `0ms`; testservices-build hit. | `HRF-006`, `HRF-009` |
| `EVID-FRONTEND-READINESS` | `*/frontend-install/target-summary.json`, `*/phase-map-check/readiness-cache-frontend-install.json` | `frontend-install` target wall `5370ms`; parent scheduler `check-frontend-install 7443ms`; readiness cache record hit in phase-map-check. | `frontend-install` target wall `0ms`; parent scheduler `check-frontend-install 362ms`; readiness cache record hit. | `HRF-007`, `HRF-009` |
| `EVID-BROWSER-SHARDS` | `*/check/pressure-summary.json`, `*/check/scheduler-summary.json`, `tools/execution_topology_manifest.json` | Browser webserver-backed aggregate `212023ms`; shard01 `40205ms`, shard06 `33635ms`, shard05 `33236ms`; stateful `38630ms`. | Browser webserver-backed aggregate `218088ms`; shard01 `41375ms`, shard05 `33941ms`, shard06 `33555ms`; stateful `40153ms`. | `HRF-008` |
| `EVID-PRESSURE-GAP` | `*/check/pressure-summary.json` | Empty `reused_accounting_counts` and `readiness_attribution_counts`; diagnostic-only pressure summary. | Same empty accounting fields; diagnostic-only pressure summary. | `HRF-009` |

## 10. Scenario Preservation Matrix

| Scenario | Phase owner | Required default-check test symbol | Required assertion surface | Required closure |
| --- | --- | --- | --- | --- |
| `SCN-001` | `E-10-01` | `TestPhase10_E_10_01_DeploymentLocalOperatorInspectLatestBackupMetadata` | Latest durable successful retained backup-set metadata, redacted durability diagnostics, deployment-local operator authorization rejection cases. | Scheduler-visible default-check evidence remains product-conformance evidence for `E-10-01`. |
| `SCN-002` | `E-10-01` | `TestPhase10_E_10_01_DeploymentLocalOperatorRestoreLatestBackup` | Latest durable retained backup restore, inactive deployment admin rejection, non-admin rejection, same-config target rejection. | Scheduler-visible default-check evidence remains product-conformance evidence for `E-10-01`. |
| `SCN-003` | `E-10-01` | `TestPhase10_E_10_01_DeploymentLocalOperatorRestoreVerifyDueRunner` | Due restore-verification cadence runner, unmarked verification target rejection, artifact-backed verification evidence. | Scheduler-visible default-check evidence remains product-conformance evidence for `E-10-01`. |
| `SCN-004-PASS` | `E-10-01` | `TestPhase10_E_10_01_ObjectStoreMigrationRunEmitsPassAndMismatchEvidence` pass scenario | Active `deployment_admin` object-store migration run emits `current_state=cutover_ready`, `cutover_ready=true`, pass ledger, pass validation, and copied or already-copied item state. | Split into its own scheduler-visible default-check scenario work. |
| `SCN-004-MISMATCH` | `E-10-01` | `TestPhase10_E_10_01_ObjectStoreMigrationRunEmitsPassAndMismatchEvidence` mismatch scenario | Target-side mismatch emits `current_state=failed`, `blocking_failure=true`, `cutover_ready` not true, fail ledger, fail validation, and target-mismatch evidence. | Split into its own scheduler-visible default-check scenario work. |

## 11. Fixture Policy Closure Matrix Contract

`HRF-005` cannot close from aggregate clone-count reduction alone. Before any fixture-policy conversion is accepted, the future goal runner MUST create or update a row inventory with the following columns for each affected row or symbol. Rows not present in this inventory MUST NOT be converted as part of `HRF-005`.

| Required column | Closed value set or source | Omission behavior |
| --- | --- | --- |
| `row_or_symbol` | Phase-map row ID plus test symbol when available. | Missing value blocks `HRF-005`. |
| `current_postgres_policy` | Current manifest or helper policy token. | Missing value blocks conversion. |
| `target_postgres_policy` | One of `transaction`, `package_reset`, `group_clone`, `template_clone`, or `migration_scratch`. | Missing value blocks conversion. |
| `fixture_budget_fields` | Owner fields `max_template_clones`, `max_group_clones`, `max_package_resets`, `max_transactions`, and `max_migration_scratch` as applicable. | Missing relevant budget blocks conversion. |
| `reason_code_field` | Closed code from `IDB-006` using the owner-required field for the selected policy. | Missing reason code blocks conversion. |
| `human_reason_field` | Row-owned human reason text required by the harness owner. | Missing reason blocks conversion. |
| `invalid_conversion_condition` | Exact condition that would make this conversion weaken fixture clarity or product evidence. | Missing condition blocks conversion. |
| `retained_evidence_path` | Fixture report, service-scope artifact, pressure summary, or phase-map path proving the conversion. | Missing retained evidence keeps `HRF-005` non-final. |

## 12. Harness Refinement Finding Registry

| Finding | Severity | Affected phase/target/test | Observed evidence | Owner import | Required correction | Expected repo surfaces | Acceptance evidence | Closure state | Blockers |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `HRF-001` | Critical | phase10 `E-10-01`; `backend-process`; `backend-process-phase10-operator-inspection` | `EVID-BACKEND-PROCESS`; `EVID-E-10-01-TESTS`. | Harness Sections 4, 10, 11, 17; `IDB-001`; `IDB-004`; phase10 `E-10-01`. | Decompose `E-10-01` into scheduler-visible scenario evidence for exactly `SCN-001`, `SCN-002`, `SCN-003`, `SCN-004-PASS`, and `SCN-004-MISMATCH`. This handoff forbids a default-check projection for `HRF-001`. | `tools/phase10_test_map.json`, scheduler owner inputs, phase-map generators, scheduler manifests, generated ledgers/schedules. | Retained warm full `check` evidence has no aggregate work unit representing all `E-10-01` scenarios; each Scenario Preservation Matrix row is selected by default check; `make scheduler-summary-timing-drift RESULTS_DIR=<successful-warm-check-run-root> TARGET=check SCHEDULER_WARM_CHECK_BUDGET_MS=180000 SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25` passes. | `repo_materialization_required` | Must preserve product-conformance ownership for `E-10-01`. |
| `HRF-002` | High | `cmd/operator` tests; hidden `buildOperatorBinary`; `make build-operator` | `EVID-HIDDEN-OPERATOR-BUILD`; live owner inspection confirms `build-operator` is public helper build and build-artifact cache family member. | Harness Sections 1, 4, 5, 8, 10, and 17; `IDB-002`; `IDB-007`; TH-HARNESS-AC-037. | Materialize the owner-defined runtime-binary interface: `build-operator OPERATOR_BIN=<path>` produces the binary; scheduler-selected operator Go work receives only `CARTULARY_OPERATOR_BIN`; Go test logs contain no nested `make build-operator` for canonical harness evidence. | `docs/testing-harness-nlspec.md`, `tools/execution_topology_manifest.json`, generated task/scheduler manifests, `tools/phase10_test_map.json`, `cmd/operator/operator_phase10_test.go`, scheduler/runtime scripts. | Passed: `make build-operator` (`.cartulary/test-results/20260609T003959Z-p86700/build-operator/build-artifact-cache-build-operator.json`); `make phase-slice PHASE=phase10` (`.cartulary/test-results/20260609T004201Z-p97822`); operator provenance at `.cartulary/test-results/20260609T004201Z-p97822/_shared/backend-process-phase10-operator-inspection/runtime-binaries.json`; negative `make check OPERATOR_BIN=/tmp/x` and `make check CARTULARY_OPERATOR_BIN=/tmp/x` exit `2` before child work; backend-process Go runner logs contain no nested `make build-operator`. | `repo_materialized` | Broader build-cache rows remain open for non-operator build targets. |
| `HRF-003` | High | phase10 object-store migration pass and mismatch work | `EVID-E-10-01-TESTS`; `EVID-OBJECT-MIGRATION`. | Harness Sections 10, 11, 12; `IDB-001`; `IDB-004`; phase10 `E-10-01`. | Split object-store migration evidence into separate scheduler-visible scenario work for `SCN-004-PASS` and `SCN-004-MISMATCH`. This handoff forbids a default-check projection for `HRF-003`. | `cmd/operator/operator_phase10_test.go`, `internal/testutil/s3test`, phase10 owner inputs, scheduler resource claims. | Retained default-check evidence has two separate scheduler-visible scenario results: pass/cutover-ready evidence for `SCN-004-PASS` and mismatch/blocking-failure evidence for `SCN-004-MISMATCH`; both remain under `primary_evidence_owner=E-10-01`. | `repo_materialization_required` | Must not drop active `deployment_admin`, mismatch, or artifact assertions. |
| `HRF-004` | High | `check-service-backed`; backend and browser lanes | `EVID-SERVICE-LANES`. | Harness Section 10; `IDB-001`; TH-HARNESS-AC-018. | Rebalance service-backed lanes after `HRF-001` decomposition so non-isolated backend/browser peer lanes remain within ratio `1.25` after the `5000ms` materiality floor. | `scripts/run-service-backed-schedule.mjs`, scheduler owner inputs, `tools/scheduler_manifest.json`, phase maps. | `make scheduler-summary-timing-drift RESULTS_DIR=<successful-warm-check-run-root> TARGET=check SCHEDULER_WARM_CHECK_BUDGET_MS=180000 SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25` passes. Failure keeps the row non-final and must name the skewed lane. | `repo_materialization_required` | `HRF-001` must be materialized first. |
| `HRF-005` | Medium | Postgres fixture policy; template-clone pressure | `EVID-FIXTURES`. | Harness Sections 10 and 11; `IDB-006`. | Populate the Section 11 fixture inventory for each affected row or symbol, then convert only rows whose inventory target policy and invalid-conversion condition are closed. Enforce structural fixture budgets and source-detail diagnostics. | Phase maps, `internal/testutil/pgtest/pgtest.go`, fixture budget checker, scheduler summaries, fixture reports. | Fixture budget validation fails unplanned clone, group, transaction, reset, migration-scratch, and budget overuse cases; retained evidence identifies source row/symbol, actual count, declared budget, policy, reason code, and retained evidence path for every converted row. | `repo_materialization_required` | Any affected row missing a Section 11 inventory entry blocks closure. |
| `HRF-006` | Medium | `build-server`, `build-web`, `build-migrate` lifecycle and cache accounting | `EVID-BUILD-CACHE`. | Harness Sections 1, 8, 10; `IDB-002`; `cartulary.cache.build_artifact.v1`. | Emit target-owned `cartulary.cache.build_artifact.v1` run-root cache artifacts for `build-server`, `build-web`, and `build-migrate`; preserve public target summaries on hit and miss; reconcile cache state, validation duration, target summary duration, and parent scheduler duration without marking correctness work as `reused`. | `scripts/cache-artifact.sh`, build scripts, `scripts/run-check-schedule.mjs`, target summary collation, schema validation. | Cold and immediate hot `make build-server build-web build-migrate` runs emit target-owned cache artifacts and target summaries for each build target; retained parent scheduler evidence contains no correctness `reused` accounting. | `repo_materialization_required` | Must preserve public target summary emission and artifact validation on cache hit. |
| `HRF-007` | Medium | `frontend-install`; frontend readiness cache | `EVID-FRONTEND-READINESS`. | Harness Sections 1, 4, 8, 13; `IDB-002`; `cartulary.cache.readiness.v1`. | Move frontend readiness cache attribution to the target that consumes it; target summary must report hit/miss state, validation duration, and installed-state validation while parent scheduler duration points to the same readiness source. | `scripts/frontend-install.sh`, `scripts/cache-artifact.sh`, task surface metadata, scheduler readiness units. | Cold and immediate hot `make frontend-install` runs emit target-owned `cartulary.cache.readiness.v1` artifacts, target summaries report cache state and validation duration, and default-check scheduler attribution references the same readiness unit. | `repo_materialization_required` | Must not skip installed-state validation or frontend summaries. |
| `HRF-008` | Medium | Browser shards; default browser support placement | `EVID-BROWSER-SHARDS`. | Harness Sections 4, 10; `IDB-001`; `IDB-004`; TH-HARNESS-AC-018. | Rebalance browser shard duration inputs and support/default-check metadata so selected default browser work uses owner-defined default-check metadata and non-isolated browser lanes remain within ratio `1.25` after the `5000ms` materiality floor. | Browser shard planner inputs, Playwright phase maps, task-surface metadata, scheduler manifests. | Warm retained `browser-e2e-webserver-backed` and full warm `check` evidence satisfy `IDB-001` lane balance; support rows expose `IDB-004` metadata; any failure names the non-isolated lane or metadata mismatch. | `repo_materialization_required` | Must preserve browser conformance rows selected by phase maps. |
| `HRF-009` | High | Scheduler, cache, readiness, pressure summaries | `EVID-PRESSURE-GAP`; cache-sensitive findings. | Harness Sections 8, 10, 17; `IDB-002`; `IDB-003`; TH-HARNESS-AC-018, AC-028, AC-031. | Keep `pressure-summary.json` diagnostic-only. Populate or preserve required diagnostic fields exactly: `reused_accounting_counts` and `readiness_attribution_counts` are present; `{}` means none emitted. Add cache/readiness attribution only through owner-owned summary fields or diagnostic fields, never through correctness `reused` accounting. | Scheduler summary reporters, pressure summary reporters, cache helper records, `agent-finalize` timing checks. | Retained check and check-service-backed pressure summaries contain the required diagnostic fields; scheduler summaries report no correctness work as `reused`; any schema promotion attempt is blocked until an owner revision adopts `cartulary.scheduler_pressure_summary.v1`. | `repo_materialization_required` | Pressure summary schema remains diagnostic-only unless an owner revision promotes it. |

## 13. Boundary Gap Registry

| Gap | Boundary | Linked findings | Owner import | Required closure payload | Acceptance evidence | Closure state |
| --- | --- | --- | --- | --- | --- | --- |
| `HRG-001` | Scheduler-native evidence decomposition | `HRF-001`, `HRF-003`, `HRF-004` | Harness Sections 4 and 10; `IDB-001`; `IDB-004`; phase10 map | Create scheduler-visible default-check scenario work for all rows in Section 10 without changing `primary_evidence_owner` or evidence class. | Retained scheduler summary names independent scenario work and every Section 10 row passes in default check. | `repo_materialization_required` |
| `HRG-002` | Operator build-artifact cache and test-consumption interface | `HRF-002`, `HRF-006` | Harness Sections 4, 5, 8, 10, and 17; `IDB-002`; `IDB-007`; TH-HARNESS-AC-037 | Materialized the runtime-binary registry entry `operator -> build-operator/OPERATOR_BIN -> CARTULARY_OPERATOR_BIN`, producer dependency, invalid-path checks, provenance artifact, and no-hidden-build test consumption. | Passed: `make harness-contract` (`.cartulary/test-results/20260609T004543Z-p10782`), `make phase-schedules` (`.cartulary/test-results/20260609T004151Z-p97620`), `make phase-slice PHASE=phase10` (`.cartulary/test-results/20260609T004201Z-p97822`), and negative public-input checks. | `repo_materialized` |
| `HRG-003` | Service fixture ownership and resource claims | `HRF-003`, `HRF-004`, `HRF-005` | Harness Sections 10 and 11; `IDB-006` | Each live Postgres/S3 row has explicit resource claims, fixture class, budget fields, reason code, human reason, and retained evidence path. | Pressure summary and service-scope artifacts identify claims and fixture classes for every converted row. | `repo_materialization_required` |
| `HRG-004` | Warm versus cold readiness attribution | `HRF-006`, `HRF-007`, `HRF-009` | Harness Sections 1, 8, 13; `IDB-002` | Separate cold provisioning, hot validation, cache state, and validation duration in target-owned summaries and scheduler-visible readiness units. | Cold/hot build/readiness runs emit target-owned cache artifacts and matching scheduler attribution. | `repo_materialization_required` |
| `HRG-005` | Cache-hit accounting without correctness reuse | `HRF-006`, `HRF-007`, `HRF-009` | Harness Section 1 cache limits; `IDB-002`; AC-028 | Cache hits accelerate only deterministic build/readiness outputs; correctness work remains actual scheduler work. | Scheduler `reused` accounting remains absent or zero for correctness work; cache counts appear only in cache/readiness fields. | `repo_materialization_required` |
| `HRG-006` | Lane-balance and critical-path accounting | `HRF-001`, `HRF-004`, `HRF-008`, `HRF-009` | Harness Section 10; `IDB-001`; AC-018 | Critical path and lane-balance reports include backend, browser, readiness, service-session setup, and service-session cleanup attribution. | Warm timing drift command from `IDB-001` passes. | `repo_materialization_required` |
| `HRG-007` | Phase-map and default-check metadata alignment | `HRF-001`, `HRF-008` | Harness Section 4; `IDB-004`; phase maps | Every default-check movement or split records closed default-check metadata and generated drift evidence. | Drift checks pass; no product-conformance row disappears from `check`; projection metadata appears only where owner permits projection. | `repo_materialization_required` |
| `HRG-008` | Browser shard duration balancing | `HRF-008` | Harness Section 10; `IDB-001`; browser topology | Duration inputs and shard planner distribute browser work so non-isolated lanes satisfy ratio `1.25` after `5000ms` materiality. | Warm browser evidence satisfies the imported balance rule or the row remains non-final with a named lane diagnostic. | `repo_materialization_required` |
| `HRG-009` | Fixture clone budget validation | `HRF-005` | Harness Section 11; `IDB-006` | Budget checker validates actual fixture counts against manifest budgets and reports row/symbol source details for overuse. | Fixture budget target fails synthetic overuse and retained warm evidence passes after fixture conversions. | `repo_materialization_required` |

## 14. Phase/Slice Execution Plan

| Slice | Owner refs | Expected edits | Validation command | Acceptance evidence | Stop conditions |
| --- | --- | --- | --- | --- | --- |
| `HR-0`: evidence normalization and owner inspection | Harness Sections 1, 4, 8, 10, 11, 17; Section 7 retained-evidence rules | Re-inspect live sources and named retained artifacts; record Section 7 TODOs for missing run facts. | `make explain-run RESULTS_DIR=.cartulary/test-results/20260608T224533Z-p79693` | Evidence ledger paths exist and no Section 7 missing-artifact TODO is required for active rows. | Missing run artifacts create TODOs and keep linked rows non-final. |
| `HR-1`: materialize operator binary injection owner rule | Harness Sections 4, 5, 8, 10, and 17; `IDB-007`; TH-HARNESS-AC-037 | Materialized: `build-operator` accepts `OPERATOR_BIN`; scheduler-selected operator Go work receives only `CARTULARY_OPERATOR_BIN`; canonical harness evidence no longer relies on hidden per-test rebuilds. | Passed: `make build-operator`; `make phase-slice PHASE=phase10`; negative public-input checks. | Build-operator cache record exists at `.cartulary/test-results/20260609T003959Z-p86700/build-operator/build-artifact-cache-build-operator.json`; operator aggregate provenance exists at `.cartulary/test-results/20260609T004201Z-p97822/_shared/backend-process-phase10-operator-inspection/runtime-binaries.json`; backend-process Go runner logs contain no child `make build-operator`. | Closed for operator injection; later slices still own scenario decomposition, fixture, and broader cache work. |
| `HR-2`: split phase10 operator evidence by scheduler-native scenario | Harness Sections 4 and 10; phase10 `E-10-01`; Section 10 matrix | Split owner inputs so the five Section 10 scenario rows are independently scheduler-visible default-check work. | `make phase-slice PHASE=phase10` | Scenario rows pass and retained scheduler evidence names each scenario row. | Product-conformance row would be weakened or projected. |
| `HR-3`: split object-store migration pass and mismatch evidence | Harness Sections 10, 11, 12; Section 10 matrix | Split pass and mismatch migration evidence into separate scheduler-visible scenario work under `E-10-01`. | `make phase-slice PHASE=phase10` | `SCN-004-PASS` and `SCN-004-MISMATCH` both pass and retain required artifact assertions. | Migration pass or mismatch coverage would become incomplete. |
| `HR-4`: reduce Postgres clone fixture cost and enforce clone budgets | Harness Section 11; Section 11 matrix | Populate fixture inventory, convert only inventoried rows, strengthen budget checker and retained reports. | `make service-backed-slice PHASE=phase10` | Retained fixture evidence and budget diagnostics include row/symbol, actual count, declared budget, policy, reason code, and evidence path. | Any affected row lacks Section 11 inventory or conversion would weaken product evidence. |
| `HR-5`: repair build/readiness cache and lifecycle accounting | Harness Sections 1, 8, 13; `IDB-002` | Align cache records, target summaries, parent scheduler durations, and readiness artifacts for build/readiness work. | `make build-server build-web build-migrate frontend-install` followed immediately by `make build-server build-web build-migrate frontend-install`; then `make json-shape-check` | Cold/hot cache artifacts and target summaries reconcile for each target. | Cache hit would skip wrapper summaries or required validation. |
| `HR-6`: add scheduler/pressure cache attribution without correctness reuse | Harness Sections 8, 10, 17; `IDB-003` | Keep pressure summary diagnostic-only and add attribution through owner-owned summary or diagnostic fields. | `make check-harness-smoke`; `make json-shape-check` | Required pressure diagnostic fields are present and correctness work is not scheduler `reused`. | Requires schema promotion not closed by owner. |
| `HR-7`: rebalance browser shards and default browser support placement | Harness Sections 4, 10; `IDB-001`; `IDB-004` | Update shard duration inputs/planner and default-check metadata for support rows. | `make browser-e2e-webserver-backed`; `make service-backed-slice PHASE=phase10` | Browser shard evidence and default-check metadata satisfy imported owner values. | Product-conformance browser row would be removed without owner metadata. |
| `HR-8`: validate phase maps, task surface, topology, and warm timing | Harness AC-018, AC-027, AC-028, AC-031 | Regenerate and validate derived manifests/ledgers after owner-input edits. | `make phase-ledger-drift phase-schedule-drift generated-artifact-policy-check json-shape-check`; `make scheduler-summary-timing-drift RESULTS_DIR=<successful-warm-check-run-root> TARGET=check SCHEDULER_WARM_CHECK_BUDGET_MS=180000 SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25` | Drift checks pass and retained warm timing command passes. | Generated artifact would require hand editing or no successful warm full `check` run exists. |
| `HR-9`: final handoff closure and Pursue Goal prompt verification | This handoff; `IDB-005` | Update registries to final closure states and verify prompt length. | `make agent-finalize RESULTS_DIR=<successful-warm-check-run-root>`; prompt length command from Section 17 | No blocking closure states remain; prompt stays under `4000` characters. | Any `repo_materialization_required`, `owner_closure_required`, or `owner_drift_conflict` remains. |

## 15. Stop Conditions

Stop the future goal and report a handoff update instead of continuing when any of these conditions is true:

1. An owner document conflicts with this handoff or with another owner.
2. A required repo fact, owner section, artifact, schema, target, or run file cannot be inspected and no specific TODO is recorded.
3. A proposed edit weakens default `make check` correctness evidence.
4. A cache hit would skip public wrapper summaries, failure classification, service lifecycle, cleanup, runtime reset, drift checks, security checks, or aggregate verdicts.
5. Correctness work would be classified as scheduler `reused`.
6. A product-conformance row would move out of default check without phase-map ownership and explicit metadata from `IDB-004`.
7. Timing evidence would be represented as product performance, benchmark, release, or Core 05 claim evidence.
8. Generated files would need hand editing.
9. `HRF-002` or `HRG-002` remains without runtime-binary producer dependency, provenance, invalid-path handling, or narrow child forwarding.
10. A Section 3 controlled open-phrase scan finds a non-closed occurrence.

## 16. Non-Goals

- Do not weaken default `make check` correctness evidence.
- Do not classify correctness work as scheduler `reused`.
- Do not use cache hits to skip public wrapper summaries, failure classification, service lifecycle, cleanup, runtime reset, drift checks, security checks, or aggregate success/failure computation.
- Do not move product-conformance rows out of default check without phase-map ownership and explicit default-check metadata.
- Do not treat timing or telemetry-like measurements as product performance claims.
- Do not implement harness refinements while revising this handoff unless the edit is necessary to create or validate the handoff document.
- Modify `docs/testing-harness-nlspec.md` only for owner closure that directly resolves a recorded gap; generated mirrors and retained artifacts remain evidence, not owners.
- Do not copy telemetry-specific requirements from the OpenTelemetry handoff; use only its execution-handoff structure.

## 17. Acceptance Criteria for This Document

- The handoff has one registry row for each diagnostic-table row: `HRF-001..HRF-009`.
- The handoff has one boundary-gap row for each `HRG-001..HRG-009`.
- Each registry and gap row has binary acceptance evidence.
- Every `MAY` in this document has same-paragraph or same-row omission behavior.
- Every repo fact is either cited by inspected path/artifact or must be replaced with `TODO(repo-inspection): <specific missing fact>` or `TODO(owner-closure): <specific missing owner rule>`.
- Any missing owner rule is represented as `owner_closure_required`.
- Final completion remains blocked while any non-final closure state remains.
- The document has a pasteable Pursue Goal prompt under `4000` characters.
- The prompt tells the future coding agent to execute this handoff, preserve authority boundaries, inspect live repo facts, implement slices in dependency order, run exact relevant Make targets, and stop on owner conflicts.
- The document does not copy telemetry-specific requirements except as structural inspiration.
- The document does not redefine Testing Harness NLSpec behavior.
- A controlled open-phrase scan over the Section 3 seed list reports only Section 3 scan seeds or closed owner-token occurrences.

- The prompt-length command returns a value below `4000`:

```sh
awk 'BEGIN{in_prompt=0; n=0} /^```text$/ && seen {in_prompt=1; next} /^```$/ && in_prompt {print n; exit} /## 18\. Pasteable Pursue Goal Prompt/ {seen=1} in_prompt {n += length($0) + 1}' docs/handoffs/testing-harness-refinement-goal-handoff.md
```

- The document passes the narrowest available docs or markdown validation target. No dedicated markdown/handoff validator was found during inspection, so use `make generated-artifact-policy-check` and `make json-shape-check`.

## 18. Pasteable Pursue Goal Prompt

```text
Execute docs/handoffs/testing-harness-refinement-goal-handoff.md as the active Cartulary testing-harness refinement handoff. docs/testing-harness-nlspec.md owns harness mechanics; Core 00-04 own product conformance; Core 05 owns claim-bearing benchmark/publication behavior. Retained run artifacts and the diagnostic table are evidence, not behavior owners. Before editing, inspect live repo facts plus the named cold/warm run roots. Use HRF-001..HRF-009 and HRG-001..HRG-009 as the required work set. Imported bounds: warm check-service-backed hard cap 240000ms, remediation target 180000ms, lane balance ratio 1.25, materiality floor 5000ms; remediation timing command is make scheduler-summary-timing-drift RESULTS_DIR=<successful-warm-check-run-root> TARGET=check SCHEDULER_WARM_CHECK_BUDGET_MS=180000 SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25. Decompose E-10-01 into scheduler-visible default-check scenario work for inspect metadata, restore latest, due restore-verify, object migration pass, and object migration mismatch; do not use default-check projection for HRF-001 or HRF-003. Treat build/readiness caches only under cartulary.cache.readiness.v1 and cartulary.cache.build_artifact.v1; cache hits must not skip summaries, validation, cleanup, lifecycle, runtime reset, drift/security checks, or aggregate verdicts, and correctness work must not become scheduler reused. Keep pressure-summary.json diagnostic-only: reused_accounting_counts and readiness_attribution_counts must be present, and {} means none emitted. HRF-002/HRG-002 import the owner-defined operator runtime-binary interface from docs/testing-harness-nlspec.md: build-operator accepts OPERATOR_BIN, scheduler-selected operator Go work consumes CARTULARY_OPERATOR_BIN, and retained evidence must include producer dependency and provenance without hidden nested builds. Implement HR-0..HR-9 in dependency order only when each row has owner closure. Run exact Make validations named in the slice table and record changed files, owner imports, retained artifacts, failures, blockers, closure-state changes, forbidden open-phrase scan, and prompt length. Final done is unavailable while any finding, gap, or slice has repo_materialization_required, owner_closure_required, or owner_drift_conflict. Stop on owner conflicts, missing owner closure, missing repo facts, generated-file hand edits, non-warm retained evidence for warm maintenance, or evidence gaps that would force guessing.
```

## 19. Handoff Update Template

```text
Testing harness refinement handoff status:
- Current slice:
- Changed paths:
- Owner refs inspected:
- Imported defaults checked:
- Retained evidence:
- Finding states:
- Boundary gap states:
- Owner-closure rows:
- Validation run:
- Forbidden open phrases scan:
- Prompt character count:
- Blockers:
- Next action:
```
