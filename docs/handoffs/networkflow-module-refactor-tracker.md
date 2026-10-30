# networkflow Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

### Current implementation iteration — IMP-07

**DONE:** S29–S31 and S15-R4 close F34/V28 OAuth scope-description classification
and F35/V27 independent preview-retention validation. These obligations and the
corresponding S15-R3 acceptance portions were reopened and are now qualified by
the affected-gap portfolio in §25. Earlier passes remain evidence of their
recorded coverage; the broader S15-R3 integration/browser portfolio is historical.

### Historical implementation iteration — IMP-06

**DONE:** S25–S28 and S15-R3 implement and validate the maintainer's byte-only
CSV, complete security accounting, and absent-field choices. F32/V27 and F33/V28 pass the frozen portfolio in §24;
S15-R3 supersedes the affected S15-R2 acceptance claims within that explicit
scope. Earlier passing evidence remains unchanged and supports only its actual
coverage. No broader current Network Flow owner qualification is inferred.

### Historical implementation iteration — IMP-05

**DONE:** S21–S24 and S15-R2 implement and validate approval integrity, explicit
route semantics and legacy retirement. F29–F31 and the affected specification,
mapping-approval and final-acceptance claims are reconciled in §23. Prior-format
compatibility obligations are DROPPED under the adopted fresh-environment
decision. Earlier implementation and retained runs remain historical evidence
of their actual coverage.

The selected boundary requires a fresh database initialized through the canonical
migration chain. Earlier pre-production import plans are unsupported; their
registry translation and exclusive fixtures/harness support are removed.
Every public operation declares mutation semantics explicitly. Network Flow
major 8/state 5 and the coordinated API candidate remain unchanged. No database
reset, commit or deployment is authorized. Each workstream must pass its required
checks and record a checkpoint before its successor begins. See §23.

### Historical implementation iteration — IMP-04

**Historical completion claim, affected exits reopened by IMP-05:** the four-gap remediation authorized on 2026-10-29 completed
S16–S20 and S15-R. F25–F28 and the reopened S09, S12-I/V19, S13, S14 and S15
acceptance obligations are closed by §22's compatible focused evidence and
operational handoff. Earlier passing and failing runs remain historical evidence
of their actual coverage. S10, S11, S12-M and S12 retain their completed
implementation posture with affected final regressions passed.

The controlling sequence is **S16 -> S17 -> S18 -> S19 -> S20 -> S15-R** in §22.
Existing uncommitted changes and S00–S15 history were preserved. The user adopted
target-neutral CSV admission: source custody precedes target-specific
interpretation, and target selection remains after upload. Explicit owner
adoption preceded dependent implementation. Transaction-ownership corrections
implement the adopted Imports commit rule. Major 8/state 5 remains the release
boundary; commit and deployment are separate unperformed actions.

### Historical implementation iteration — IMP-03

**Historical completion claim, superseded by IMP-04:** S09–S15 were recorded DONE,
including S12-M/S12-I, with F18–F24 and RB-004/RB-005 closed. §22 reopens the
affected acceptance claims without erasing §21's owner adoption, source identity,
focused validation results, or operational handoff. Deployment remains separate.

On 2026-10-29 the user requested implementation of the revised remediation plan. This authorizes specification preparation, characterization, implementation, authored migrations, generated projections and focused validation. The first execution change is this tracker checkpoint. Existing DOC-02 changes and S00–S08 history are preserved. The implementation baseline is `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`; this tracker is the sole pre-existing modified file, with no staged or non-ignored untracked files.

The current sequence is **S09 -> S10 -> S11 -> S12-M -> S12 -> S12-I -> S13 -> S14 -> S15**. Each workstream requires completed validation and a tracker checkpoint before its successor begins. §21 controls this execution and supersedes DOC-02's tracker-only authorization, no-migration assumption, unchanged internal-source-port assumption and target-independent source-row admission. Human owner adoption remains a distinct S09 gate before dependent implementation; deployment is not authorized. Tests and generated/runtime evidence must not consume Markdown.

### Historical planning iteration — DOC-02

This is the sole controlling plan and execution record for Network Flow. The current request authorizes a **document update only**: record the next iteration for legacy removal and bounded-resource hardening. S09–S15 are future implementation work; this invocation authorizes no production, test, contract, generated, migration, deployment, deletion, or conversion changes. The user explicitly included the necessary Imports source-reader dependency and excluded a full CI rerun.

DOC-02 baseline: clean `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, revalidated before editing. The current target has 132 Go files: the original 129-file inventory plus three files added by the completed remediation. Current inspection coverage, owner decisions, F18–F23, S09–S15, validation obligations and handoff are in §§16–20. That targeted inspection is not a new exhaustive module or conformance audit.

The current plan is `S09 -> S10 -> S11 -> S12 -> S13 -> S14 -> S15`. Each future implementation slice is a separate workstream. Its validation and tracker checkpoint must finish before its successor begins. Planning completion is separate from implementation readiness and authorization; the current readiness and completion records are in §§18 and 20.

Major 8 / document 8.0.0 / durable state 5 / minimum admitted state 5 remains the coordinated release boundary. Human owner adoption and deployment remain separate gates. Older data remains untouched; original CSV sources and mapping decisions are required for re-import. No new public major, durable state, parser-profile ID, compatibility alias, or migration is proposed merely to remove dead code or correct implementation against an unchanged owner definition.

### Completed remediation history — S00–S08

The remainder of this section records the completed remediation; §§2–12 retain DOC-01 planning history, and §§13–15 retain its implementation evidence and handoff. Their original baselines, statuses and authorization statements are historical, not current instructions or acceptance evidence for S09–S15. Current planning decisions supersede earlier preservation restrictions only where explicitly identified below, including F18's targeted retirement of the inert contracts retained under F10.

On 2026-10-29 the user explicitly authorized the previous complete remediation plan, including owner amendments, tests, implementation, generated projections, a diagnostic storage migration, and a coordinated fresh-state cutover. That work completed as S00–S08. Deployment, destructive resets, automatic deletion, and historical data conversion were not authorized.

### Completed remediation execution

Baseline revalidated: `main` at `a182673ae3b460de9c0d078d3315f17e18c4f36a`; tracked files clean; this tracker is the sole pre-existing untracked file. The original 129-file inventory and DOC-01 evidence remain historical baseline evidence, not current test results.

The user approved implementation of major 8 / document 8.0.0 / durable state 5 / minimum state 5. Owner amendments remain **proposed** pending human owner review and explicit coordinated adoption; implementation authorization, passing tests and deployment adoption are separate gates. Prior state is left untouched and fails admission. Original sources are required for re-import into fresh state.

Execute exactly `S00 -> S03 -> S01 -> S02 -> S04 -> S05 -> S06 -> S07 -> S08`. Update this tracker after each workstream's validation and before beginning its successor. Tests, generators, runtime metadata, conformance and release evidence must never consume Markdown. Source requirements and projection fidelity are reviewed by humans.

| Slice / workstream | Status | Depends on | Goal / binary exit |
| --- | --- | --- | --- |
| S00 Specification closure and tracker revision | DONE | baseline | Coordinated owners and complete finding/validation traceability |
| S03 Canonical tracker validation | DONE | S00 | Module-tracker glob and passing Make Markdown lint |
| S01 Contract characterization | DONE | S03 | Dedicated registered preview baseline and safe Imports failure tests pass |
| S02 Transport-independent semantic boundary | DONE | S01 | Shared strict decoder and transitive import-root guard pass |
| S04 Diagnostic semantics and storage identity | DONE | S02 | Correct identities/messages/privacy and table-scoped uniqueness pass |
| S05 Complete previews and accurate retention | DONE | S04 | Complete bounded preview; exact ordered apply retention and rejected-row counts pass |
| S06 Strict preview-result admission | DONE | S05 | Closed typed results, nested values and mapping consistency pass |
| S07 Coordinated major/state/client cutover | DONE | S06 | Major 8/state 5 publication, fresh admission and unchanged old-state rejection pass |
| S08 Validation and handoff completion | DONE | S07 | Finalized source, user-revised focused portfolio and complete handoff |

The original S01/S02 IDs are retained. Active statuses in this table and §§14–15 supersede the explicitly historical planning statuses; S08 reconciled every current finding and blocker without erasing DOC-01 evidence. F03/RB-002 is resolved by retaining owner-local harness controls with ordinary-build isolation and runtime scan inclusion. F04-F07 and F09 remain supported boundaries; F08's zero-generated-diff expectation is superseded by S07; F10 does not authorize cosmetic historical renaming.

| New finding | Required remediation / areas | Rationale, benefit and unresolved risk | Compatibility / validation / slice |
| --- | --- | --- | --- |
| F13 Incomplete preview-result admission | Exact required presence, closed nested types, mapping and diagnostic equality; implementation/tests/owner clarification | One trustworthy owner boundary; otherwise malformed results and inconsistent approvals can escape | Valid schemas unchanged; malformed results reject safely; S06 |
| F14 Diagnostic construction drift | Owner ID bytes, closed reasons, messages, attribution and redaction; implementation/tests/fixtures/docs | Stable provenance and privacy; otherwise collisions, misleading values or disclosure | Correct erroneous emitted values under fresh-state cutover; fixed vectors and preview/apply parity; S04 |
| F15 Preview/retention accounting conflated | Separate preview completeness and apply retention; count rejected records independently; implementation/tests/spec | Predictable bounded validation; otherwise low limits disable preview and exact-limit flags/counts lie | Correct only nonconforming behavior; zero/exact/overflow and 50/51 cases; S05 |
| F16 Global diagnostic uniqueness | New authored migration to table+diagnostic identity; implementation/SQL/tests | Same source can appear in distinct tables; otherwise legitimate imports fail | Physical schema change, no data rewriting; cross-table success/same-table rejection; S04 |
| F17 Retained diagnostic admission absent | State-5 owner validator and coordinated current-only publication; spec/contracts/code/tests/docs | Prevent incompatible facts from being served | Major 8/state 5; old state rejected unchanged and fresh/restored valid state admitted; S07 |

F01/F02 -> S02; F11/G01/RB-001 -> S01; F12/RB-003 -> S03. Each workstream retains exact commands, exits and artifacts in the execution log below. No historical success is promoted into current evidence.

### Remediation execution log

| Session / slice | Changes / decisions | Commands / results / evidence | Next gate |
| --- | --- | --- | --- |
| 2026-10-29 / S00 start | Baseline and authorization revalidated; controlling tracker updated first | Read-only Git and applicable instructions; no product verification yet | Owner closure, then S03 |
| 2026-10-29 / S00 completion | Prepared user-approved major 8/state 5 amendments across NF, Extensions and Core; explicit diagnostic contracts and current-only policy | Agent source review; human owner adoption not established; product execution not claimed; tracker lint selection pending S03 | S03 |
| 2026-10-29 / S03 completion | Module-tracker glob covers both trackers; RB-003 resolved | `CARTULARY_TEST_RUN_ID=nf-s03-markdown-20261029-01 make lint-markdown` exit 0; `.cartulary/test-results/nf-s03-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json` | S01 |
| 2026-10-29 / S01 completion | Dedicated preview admission row registered; G01/RB-001 closed for decoder movement | Make generate/format passed; `test-slice` NF preview/error/boundary rows passed at `.cartulary/test-results/nf-s01-baseline-20261029-01`; Imports safe owner translation unit/integration rows passed at `.cartulary/test-results/nf-s01-imports-20261029-01`; outer exits 0 | S02 |
| 2026-10-29 / S02 completion | F01/F02 fixed; strictjson used directly and import roots guarded | Focused NF rows exit 0 at `.cartulary/test-results/nf-s02-focused-20261029-01`; backend boundary lint exit 0 at `.cartulary/test-results/nf-s02-boundary-20261029-01` | S04 |
| 2026-10-29 / S04 validation repair | New duplicate-rejection test poisoned its borrowed rollback transaction; isolate expected SQL failure in a nested transaction | `nf-s04-focused-20261029-01` failed product assertion (outer 2); unit identity passed, schema check passed at `nf-s04-schema-20261029-01`; migration-drift not reached by command chain | Rerun affected store row and migration drift |


### Historical planning baseline

The following baseline scope table and §§2–12 preserve DOC-01 exactly as planning history. Their restrictions, readiness states, skipped checks and completion claims are superseded by the active remediation table and execution evidence. They do not describe this implementation's current authorization or acceptance.

| Scope item | Recorded value |
| --- | --- |
| Supplied and canonical target | `internal/modules/networkflow`, resolving to `/home/jochi/code/cartulary/internal/modules/networkflow` |
| Derived label | `networkflow`: last target-path component, normalized to lowercase kebab case; matches `[a-z0-9]+(?:-[a-z0-9]+)*` |
| Output / sole allowed write | `docs/handoffs/networkflow-module-refactor-tracker.md` |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, read first; doctrine/template, not current-state proof |
| Baseline | Branch `main`; commit `a182673ae3b460de9c0d078d3315f17e18c4f36a`; initially clean index/worktree; no nonignored untracked files; tracker absent |
| Admission | Target and output parent are canonical repository directories, with no symlink traversal; no target symlinks, nested AGENTS, ignored or nonignored untracked target files found |
| File universe | 129 tracked Go files: 115 in the package root (67 production/support, 48 tests), 14 in `harnesscontrol` (8 support, 6 tests); 75 non-test and 54 test files total |
| Allowed changes | This tracker only, including its planning decisions and session history |
| Non-goals | Code movement, patches, test execution/addition, generated edits, dependency changes, migrations, runtime changes, harness changes, frontend redesign, release or conformance claims |
| Planning outcome | Complete for the declared target inventory and bounded structural recommendation; broader behavioral conformance and exhaustive caller coverage are not asserted |
| Slice readiness | S01 ready for a later authorized characterization task after baseline revalidation; S02 blocked by RB-001 |
| Implementation authorization | Not granted by this planning invocation; “implement this plan” means create this tracker |

Authority is applied by named scope, in this order: adopted subsystem NLSpecs; Core 00–04 implementation-conformance requirements; Core 05 only for claim-bearing timed/fixture-sensitive publication; domain vocabulary and implementation-support guidance; current code/tests as implementation evidence; prior plans/handoffs/framework as evidence only. Typed contracts and test rows are downstream projections. A matching test does not prove specification completeness. Unresolved applicable primary-owner conflict must be recorded as **BLOCKED: owner contradiction**; none was established in the inspected clauses. A supporting-file mismatch is not automatically such a contradiction.

| Evidence ID | Owner/support source inspected | Authority and inspection scope |
| --- | --- | --- |
| E01-NF | `docs/network-flow-activity-nlspec.md` | Adopted/current 7.0.0, public major 7, durable state 4, minimum migratable 3. Scope NF-REQ-001–006; imports/rows/query/graph/link/security/routes/saved graphs; §§27–29, especially NF-REQ-210–212. Existing facade/application boundary is required, not a new redesign proposal. |
| E01-GP | `docs/graph_projection_nlspec.md` | Adopted/current 2.2.0; scope/native V2 boundary and result/lease/restore ownership. NF imports §§8–9, while its own §29 governs flow-source contributions. |
| E01-EXT | `docs/extension-subsystem-nlspec.md` | Adopted/current 0.12.0, contract major 2; status/scope and discovery/claim coordination. Historical revision explanations do not override the current NF major-7 owner. |
| E01-RPT | `docs/reporting-subsystem-nlspec.md` | Adopted/current 1.4.0; status and scope separation. NF §29 defines the source-side exact graph lease/read/release obligations used here. |
| E01-HAR | `docs/testing-harness-nlspec.md` | Adopted/current; relevant Network Flow control families and §12.2.9. Registry tests alone cannot prove real product rollback/audit behavior. |
| E01-C00 | `docs/spec/00_document_set_status_and_precedence.md` | Precedence, adoption and scope allocation; generic graph and analytical import owner boundaries. |
| E01-C01 | `docs/spec/01_architecture_storage_and_view_contracts.md` | Required modules, current-record ownership, HTTP/discovery/view/import/restore interfaces; source-owner interfaces remain narrow. |
| E01-C02 | `docs/spec/02_domain_model_schema_and_history.md` | Core record/indicator terminology and indicator participant REQ-02-074C; analytical rows cannot become Core observation source records. |
| E01-C03 | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Workbook/import/collaboration and save-state sections; durable resource invalidation belongs to Collaboration. |
| E01-C04 | `docs/spec/04_security_deployment_and_conformance.md` | Route admission, including REQ-04-105A, and transaction/audit/privacy obligations including REQ-04-135–138. |
| E01-SUP | `docs/domain.md`; `docs/research/nlspec-spec.md`; supplied framework | Vocabulary, source posture and planning structure only. Core 05 applicability was assessed; no timed/publication claim is made and no capacity certification was run. |
| E01-PROC | `AGENTS.md`, `internal/AGENTS.md`, `tools/AGENTS.md`; refactor-tracker skill and its `references/tracker-format.md`; cartulary-test-ops skill and command recipes | Repository procedure, boundary and harness mechanics; planning-only scope takes precedence over implementation examples in these sources. |

Repository files inspected are accounted individually by E02-001–E02-129 in §2. The external inspection ledger E03–E09 is in §10. “Inspected” means the exact file was opened: imports/declarations plus relevant bodies or selected test assertions. It does not mean every line was audited. The proposed edit seam, its decoder implementation, existing semantic guard, and related Imports call sites received direct body-level inspection. No inventory file is excluded or left metadata-only.

The next discovery boundary is intentionally narrow: revalidate the baseline and S01/S02 files and their Imports callers. A future move of state, worker, security, or harness responsibilities must deepen that family's characterization first; this tracker does not authorize such movement.

## 2. Current-State Repository Inventory

All paths in the first column are inside the canonical target. D followed by a number records the inspected file's line count, not a claim that every line was behaviorally audited. Inbound “reference candidates” and “source references” are bounded lexical cross-references between directly opened files, not a resolved Go call graph or coverage report. Private receiver methods with exported spellings are not automatically supported public APIs. Contract IDs resolve in §4. No target file is generated; generated dependencies are recorded as consumed surfaces.

Tests listed here were inspected, **not executed**. A TODO for an exact direct test mapping is an entry gate if that file is later selected for change; the contract-level posture in §4 supports this bounded plan without claiming exhaustive coverage.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| [api.go](/home/jochi/code/cartulary/internal/modules/networkflow/api.go) | Profile IDs, owner validation errors and effective resource limits | ProfileID, RouteContributionID, WorkspaceContributionID, EffectiveLimits | same-package reference candidates: `configuration.go`, `extension_state.go`, `graph_restore_source.go` | `github.com/google/uuid`; 1 standard-library imports | source references: `application_composition_test.go`, `graph_capacity_test.go`, `graph_query_application_test.go`; +6 other source references | C01, C03, C16; no generated file authored here | Network Flow | medium | E02-001; D194 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [application.go](/home/jochi/code/cartulary/internal/modules/networkflow/application.go) | Serialized table rename/soft-delete commands, admission/replay and atomic effects | private tableApplication, tableMutationOutcome | same-package reference candidates: `table_http.go` | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/incidents/admission`, `internal/platform/authn`; 2 standard-library imports | indirect tests in C04, C08, C09; TODO: exact direct call-to-test mapping before editing this file | C04, C08, C09; no generated file authored here | Network Flow | high | E02-002; D82 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [application_composition.go](/home/jochi/code/cartulary/internal/modules/networkflow/application_composition.go) | One retained private application bundle from the module dependency set | private activeApplications; prepareActiveApplications | same-package reference candidates: `graph_result_cleanup_dispatcher.go`, `module.go`, `routes.go` | `internal/modules/graphprojection/postgresresult`, `internal/modules/imports`, `internal/modules/incidents/admission`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 2 standard-library imports | indirect tests in C01; TODO: exact direct call-to-test mapping before editing this file | C01; no generated file authored here | Network Flow | high | E02-003; D59 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [application_composition_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/application_composition_test.go) | Construction probes for coherent dependencies, inactive/active contributions and retained applications | test-only `RegisterHandler`, `constructionModule`; +1 other source references | query_authoring_test.go via production surface | `internal/modules/crossownertransaction`, `internal/platform/httpapi`, `internal/platform/jobs`, `internal/platform/postgres`; 3 standard-library imports | helper/bridge; query_authoring_test.go via production surface | C01; no generated file authored here | Network Flow owner tests | medium | E02-004; D139 (declarations/imports and selected assertions); not executed |
| [binding_store.go](/home/jochi/code/cartulary/internal/modules/networkflow/binding_store.go) | Owner binding identity lookup/insert and audit in caller transaction | private store binding methods | same-package reference candidates: `indicator_link.go`, `indicator_link_receipts.go`, `transaction_participants.go` | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/indicators`; 9 standard-library imports | indirect tests in C07, C09; TODO: exact direct call-to-test mapping before editing this file | C07, C09; no generated file authored here | Network Flow binding source; Indicators owns target | high | E02-005; D275 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [collaboration_producer.go](/home/jochi/code/cartulary/internal/modules/networkflow/collaboration_producer.go) | Append table/graph resource invalidation intents within owner transaction | ResourceIntent; ResourceIntentAppender | store.go; saved_graph_application.go; app/server/collaboration_intents.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 4 standard-library imports | store_test.go; routes_integration_test.go | C08, C09; no generated file authored here | Network Flow producer / Collaboration persistence | medium | E02-006; D84 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [configuration.go](/home/jochi/code/cartulary/internal/modules/networkflow/configuration.go) | Closed configuration overlay, limits, key-ring findings and normalization | Configuration, ResourceLimitOverrides, ConfigurationFinding; normalization/clone/overlay functions | app/configassembly/{configuration,deployment,extension_admission}.go | 7 standard-library imports | source references: `configuration_test.go`, `graph_capacity_test.go` | C01, C16; no generated file authored here | Network Flow configuration contribution | medium | E02-007; D458 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [configuration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/configuration_test.go) | Configuration overlays, typed findings and effective limit assertions | test-only `TestNetworkFlowConfigurationContribution_Unit`, `TestNetworkFlowEffectiveResourceLimits_Unit` | Go test entrypoints | `github.com/google/uuid`; 4 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C01, C16; no generated file authored here | Network Flow owner tests | medium | E02-008; D239 (declarations/imports and selected assertions); not executed |
| [csv_parser.go](/home/jochi/code/cartulary/internal/modules/networkflow/csv_parser.go) | Bounded headered CSV parsing, field normalization, accepted rows and rejected diagnostics | private parser and row validators | same-package reference candidates: `graph.go`, `import_facade.go`, `mapping.go`; +3 other source references | 13 standard-library imports | source references: `network_flow_contract_test.go`, `network_flow_unit_test.go`, `store_test.go` | C03, C05; no generated file authored here | Network Flow analytical import target | medium | E02-009; D577 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [digest.go](/home/jochi/code/cartulary/internal/modules/networkflow/digest.go) | Canonical JSON, source/row/mapping digests and deterministic flow IDs | private identity/canonicalization helpers | same-package reference candidates: `csv_parser.go`, `extension_state.go`, `graph.go`; +11 other source references | `github.com/google/uuid`; 11 standard-library imports | source references: `graph_capacity_test.go`, `graph_streaming_test.go`, `graph_temporal_test.go`; +4 other source references | C03, C05, C06; no generated file authored here | Network Flow | medium | E02-010; D300 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [entropy.go](/home/jochi/code/cartulary/internal/modules/networkflow/entropy.go) | Instance-owned entropy reader for table IDs | private entropyReader | same-package reference candidates: `store.go` | 3 standard-library imports | indirect tests in C03, C14; TODO: exact direct call-to-test mapping before editing this file | C03, C14; no generated file authored here | Network Flow; harness injects test entropy | medium | E02-011; D26 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [extension_state.go](/home/jochi/code/cartulary/internal/modules/networkflow/extension_state.go) | Owner durable-family counts and validation | ExtensionStateReader; ExtensionStateFamilyCounters; ValidateExtensionState | saved_graph_admission.go; app/configassembly/extension_admission.go | `internal/platform/extensionstore`; 4 standard-library imports | source references: `extension_state_v3_integration_test.go`, `saved_graph_admission_integration_test.go` | C12; no generated file authored here | Network Flow extension contribution | medium | E02-012; D139 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [extension_state_v3_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/extension_state_v3_integration_test.go) | Reject obsolete saved semantic bytes without rewriting retained state | test-only `TestNetworkFlowExtensionStateV4RejectsV1WithoutRewritingBytes_Integration` | Go test entrypoint | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/networkflow`, `internal/platform/extensionstore`, `internal/platform/postgres`, `internal/testutil/revisionsupport`; 5 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C12; no generated file authored here | Network Flow owner tests | medium | E02-013; D149 (declarations/imports and selected assertions); not executed |
| [graph.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph.go) | Flow graph semantic admission, aggregates, projection input and contributor selectors | private graph request/composition types and helpers | same-package reference candidates: `extension_state.go`, `graph_query_application.go`, `graph_query_presentation.go`; +11 other source references | `github.com/google/uuid`; 8 standard-library imports | source references: `graph_capacity_test.go`, `graph_query_application_test.go`, `graph_streaming_test.go`; +6 other source references | C06, C10; no generated file authored here | Network Flow flow-specific semantics | medium | E02-014; D1278 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_capacity_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_capacity_test.go) | Capacity workload and host/runtime evidence capture; separate claim-bearing gate | test-only `TestNetworkFlowGraphCapacityCertification_Integration` | Go test entrypoint; Core 05 only if publishing timed/fixture claims | `github.com/google/uuid`; 11 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C06, C15; no generated file authored here | Network Flow owner tests | medium | E02-015; D412 (declarations/imports and selected assertions); not executed |
| [graph_cleanup_deadline_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_cleanup_deadline_test.go) | Stalled DB probe verifies cleanup deadline and rollback | test-only `BeginTx`, `QueryRow`; +3 other source references | graph cleanup integration assertions | `internal/platform/postgres`, `github.com/jackc/pgx/v5`; 4 standard-library imports | helper/bridge; graph cleanup integration assertions | C11; no generated file authored here | Network Flow owner tests | medium | E02-016; D64 (declarations/imports and selected assertions); not executed |
| [graph_materialization_payload.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_materialization_payload.go) | Strict shared materialization payload admission | private decodeGraphViewMaterializationPayload | same-package reference candidates: `graph_restore_source.go`, `graph_view_jobs.go`, `saved_graph_admission.go` | `internal/platform/strictjson`, `github.com/google/uuid`; 4 standard-library imports | source references: `graph_materialization_payload_test.go` | C10, C12; no generated file authored here | Network Flow | medium | E02-017; D36 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_materialization_payload_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_materialization_payload_test.go) | Malformed payload corpus and startup/restore admission probes | test-only `assertMaterializationPayloadAdmission`, `ObserveExecution`; +9 other source references | saved_graph_admission_integration_test.go / worker assertions | `internal/platform/authn`, `internal/platform/jobs`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 7 standard-library imports | helper/bridge; saved_graph_admission_integration_test.go / worker assertions | C10, C12; no generated file authored here | Network Flow owner tests | medium | E02-018; D139 (declarations/imports and selected assertions); not executed |
| [graph_materialization_timeout_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_materialization_timeout_test.go) | Deadline becomes terminal before publication; finalizer probes | test-only `TestGraphViewMaterializationDeadlineIsTerminalBeforePublication_Unit` | Go test entrypoint | `github.com/google/uuid`, `internal/platform/jobs`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C10; no generated file authored here | Network Flow owner tests | medium | E02-019; D75 (declarations/imports and selected assertions); not executed |
| [graph_projection_adapter.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_projection_adapter.go) | Single native Graph Projection V2 invocation and safe provider errors | private graphProjectionPort/adapter | same-package reference candidates: `graph_restore_source.go`, `graph_source_composer.go`, `module.go` | `internal/modules/graphprojection`; 11 standard-library imports | source references: `graph_capacity_test.go`, `graph_materialization_payload_test.go`, `graph_materialization_timeout_test.go`; +5 other source references | C06; no generated file authored here | Network Flow adapter / Graph Projection engine | medium | E02-020; D95 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_query_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_query_application.go) | Graph read admission, composition, contributors and completion audit | private graphQueryApplication; graphQueryComposer | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `saved_graph_read_application.go` | `internal/modules/incidents/admission`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 4 standard-library imports | source references: `saved_graph_read_application_test.go` | C02, C06, C09; no generated file authored here | Network Flow | high | E02-021; D179 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_query_application_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_query_application_test.go) | Graph success/failure/audit/session completion with DB and writer probes | test-only `TestGraphQueryCompletion_Unit` | Go test entrypoint | `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpauth`, `internal/platform/postgres`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 7 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C06, C09; no generated file authored here | Network Flow owner tests | medium | E02-022; D176 (declarations/imports and selected assertions); not executed |
| [graph_query_presentation.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_query_presentation.go) | Graph and saved-graph contributor response resources | private response helpers | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go` | 0 standard-library imports | indirect tests in C06, C10; TODO: exact direct call-to-test mapping before editing this file | C06, C10; no generated file authored here | Network Flow response projection | medium | E02-023; D23 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_response_v2.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_response_v2.go) | Exact V2 result metadata, selector identity and response conversion | private graph result conversion | same-package reference candidates: `graph_routes.go`, `graph_source_composer.go`, `graph_view_routes.go`; +1 other source references | `internal/modules/graphprojection`; 2 standard-library imports | source references: `graph_streaming_test.go`, `graph_temporal_test.go` | C06, C10; no generated file authored here | Network Flow response projection | medium | E02-024; D201 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_restore_source.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_restore_source.go) | Restore source registration and bounded restored-job reconciliation | NewGraphRestoreSourceRegistration; ReconcileGraphRestoreJobsTx | app/recoveryassembly/graphprojection_restore.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/restore`, `internal/platform/jobs`, `internal/platform/postgres`; 3 standard-library imports | source references: `application_composition_test.go`, `graph_materialization_payload_test.go`, `graph_view_authority_integration_test.go`; +1 other source references | C12; no generated file authored here | Network Flow source contribution / Recovery orchestration | high | E02-025; D143 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_result_cleanup.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup.go) | Bounded result candidate sweep, lease cleanup and selected-result safety | private graphResultCleanupService | same-package reference candidates: `graph_result_cleanup_dispatcher.go` | `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/platform/postgres`; 4 standard-library imports | source references: `graph_cleanup_deadline_test.go`, `graph_result_cleanup_dispatcher_test.go`, `graph_result_cleanup_integration_test.go`; +1 other source references | C11; no generated file authored here | Network Flow selection / Graph Projection result storage | high | E02-026; D178 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_result_cleanup_dispatcher.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_dispatcher.go) | Periodic instance-owned cleanup scheduling and lifecycle | GraphResultCleanupDispatcher; NewGraphResultCleanupDispatcher; Start/Close | private package use; no distinct direct caller established in lexical map | `internal/modules/graphprojection`; 4 standard-library imports | source references: `application_composition_test.go`, `graph_result_cleanup_dispatcher_test.go`, `graph_telemetry_test.go` | C11; no generated file authored here | Network Flow | high | E02-027; D193 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_result_cleanup_dispatcher_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_dispatcher_test.go) | Scheduler lifecycle, nonoverlap and fatal-loss assertions | test-only `TestGraphResultCleanupDispatcherLifecycleSchedulingAndFatalLoss_Unit` | Go test entrypoint | `internal/modules/graphprojection`; 5 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C11; no generated file authored here | Network Flow owner tests | medium | E02-028; D231 (declarations/imports and selected assertions); not executed |
| [graph_result_cleanup_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_integration_test.go) | Selected/leased results protected; owner scoping, sweep bounds and commit failure | test-only `TestNetworkFlowGraphResultCleanupIsOwnerScopedSelectedSafeAndBounded_Integration` | Go test entrypoint | `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/platform/postgres`, `internal/testutil/revisionsupport`; 5 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C11; no generated file authored here | Network Flow owner tests | medium | E02-029; D206 (declarations/imports and selected assertions); not executed |
| [graph_result_cleanup_race_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_race_integration_test.go) | Publication and lease races against candidate deletion | test-only `TestNetworkFlowGraphCleanupPublicationAndLeaseRaces_Integration` | Go test entrypoint | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`, `internal/modules/auth/testsupport/storetest`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/testutil/appsupport`, `internal/testutil/networkflowsupport`, `internal/testutil/pgtest`; 3 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C10, C11; no generated file authored here | Network Flow owner tests | medium | E02-030; D319 (declarations/imports and selected assertions); not executed |
| [graph_result_cleanup_test_bridge_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_test_bridge_test.go) | Test-only constructors/bounds for private cleanup service | test-only `NewGraphResultCleanupService`, `SetGraphResultCleanupTestBounds` | cleanup integration tests | `internal/platform/postgres`; 1 standard-library imports | helper/bridge; cleanup integration tests | C11; no generated file authored here | Network Flow owner tests | medium | E02-031; D19 (declarations/imports and selected assertions); not executed |
| [graph_routes.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_routes.go) | HTTP graph/query contributor extraction and response framing | private routeService handlers | private package use; no distinct direct caller established in lexical map | `internal/platform/httpapi`; 1 standard-library imports | source references: `graph_query_application_test.go` | C02, C06; no generated file authored here | Network Flow transport adapter | medium | E02-032; D70 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_source_composer.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_source_composer.go) | Shared route/worker/restore flow source traversal and projection composition | private graphSourceComposer; graphSourceReader | same-package reference candidates: `graph_query_application.go`, `graph_restore_source.go`, `graph_view_jobs.go`; +3 other source references | `github.com/google/uuid`; 4 standard-library imports | source references: `graph_query_application_test.go`, `graph_telemetry_test.go`, `graph_temporal_test.go`; +1 other source references | C06, C10, C12; no generated file authored here | Network Flow source owner | medium | E02-033; D317 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_streaming_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_streaming_test.go) | Aggregate retained-state bounds, limit outcomes, native V2 identity and SQL error propagation | test-only `TestStreamingDefaultGraphV2Metadata_Unit`, `TestStreamingGraphRetainedStateFollowsAggregateCardinality_Unit`; +3 other source references | Go test entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/platform/postgres`; 7 standard-library imports | 5 named Test entrypoint(s) in this file; assertions inspected | C06; no generated file authored here | Network Flow owner tests | medium | E02-034; D223 (declarations/imports and selected assertions); not executed |
| [graph_telemetry.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_telemetry.go) | Safe phase/result/cleanup observations and failure vocabulary | GraphTelemetryObserver; GraphPhaseTelemetryObservation; GraphResultTelemetryObservation; GraphCleanupTelemetryObservation | app/server/network_flow_telemetry.go; source composer/cleanup | `internal/modules/graphprojection`; 3 standard-library imports | source references: `graph_telemetry_test.go` | C16; no generated file authored here | Network Flow contribution / telemetry adapter | medium | E02-035; D205 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_telemetry_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_telemetry_test.go) | Observer panic containment, closed outcomes and semantic observation boundaries | test-only `TestNetworkFlowGraphTelemetryBoundaryContainsObserverFailure_Unit`, `TestNetworkFlowGraphTelemetryOutcomeVocabulary_Unit`; +1 other source references | Go test entrypoints | `internal/modules/graphprojection`; 4 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C16; no generated file authored here | Network Flow owner tests | medium | E02-036; D194 (declarations/imports and selected assertions); not executed |
| [graph_temporal.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_temporal.go) | Time buckets, conservation, canonical timestamp/selector and digest semantics | private temporal composition helpers | same-package reference candidates: `graph.go`, `graph_response_v2.go`, `graph_source_composer.go`; +3 other source references | `github.com/google/uuid`; 4 standard-library imports | source references: `graph_capacity_test.go`, `graph_temporal_test.go` | C06; no generated file authored here | Network Flow | medium | E02-037; D229 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_temporal_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_temporal_test.go) | Bucket arithmetic/property seeds, conservation, projection identity and digest fixtures | test-only `TestTimeBucketArithmeticFixturesAndLimit_Unit`, `TestTimeBucketAdmissionAndCanonicalTimestampErrors_Unit`; +4 other source references | Go test entrypoints and FuzzTimeBucketArithmetic | 5 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C06; no generated file authored here | Network Flow owner tests | medium | E02-038; D364 (declarations/imports and selected assertions); not executed |
| [graph_view_authority_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_authority_integration_test.go) | Submitter role revalidation and notification-after-commit probes | test-only `TestSavedGraphMaterializationRevalidatesSubmitterRole_Integration` | Go test entrypoint / saved-graph integration | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/platform/authn`, `internal/platform/jobs`, `internal/platform/postgres`, `internal/testutil/appsupport`, `internal/testutil/collaborationsupport`, `internal/testutil/httptestx`; 7 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C09, C10; no generated file authored here | Network Flow owner tests | medium | E02-039; D269 (declarations/imports and selected assertions); not executed |
| [graph_view_authority_test_bridge_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_authority_test_bridge_test.go) | Test-only worker handler and saved-create command construction | test-only `NewGraphViewAuthorityTestHandler`, `NewSavedGraphCreateCommandForTest` | graph_view_authority_integration_test.go | `internal/modules/incidents/admission`, `internal/platform/authn`, `github.com/google/uuid`, `internal/platform/jobs`; 3 standard-library imports | helper/bridge; graph_view_authority_integration_test.go | C09, C10; no generated file authored here | Network Flow owner tests | medium | E02-040; D35 (declarations/imports and selected assertions); not executed |
| [graph_view_jobs.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_jobs.go) | Materialization execution, authority recheck, finalization and selected bindings | GraphViewMaterializationJobKind; GraphViewJobTransactions/Manager/Runner/Mutation/SuccessFinalization/FailureFinalization/Finalizer | Module.RegisterGraphViewWorker; app/extensionassembly/network_flow_jobs.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/incidents/admission`, `internal/platform/jobs`; 4 standard-library imports | source references: `graph_materialization_payload_test.go`, `graph_materialization_timeout_test.go`, `graph_view_receipts_test.go` | C09, C10; no generated file authored here | Network Flow worker / Jobs lifecycle | high | E02-041; D360 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_receipts.go) | Final idempotency reconciliation and saved declaration receipt integrity | GraphViewReceiptReconciler | same-package reference candidates: `graph_view_routes.go`, `indicator_link.go`, `routes.go`; +5 other source references | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/jobs`; 7 standard-library imports | source references: `graph_view_receipts_test.go` | C10, C12; no generated file authored here | Network Flow receipt contribution | high | E02-042; D195 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_receipts_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_receipts_test.go) | Name-byte constraints and historical receipt integrity/comparison | test-only `TestGraphViewNameByteContract_Unit`, `TestGraphViewReceiptIntegrityAndComparison_Unit` | Go test entrypoints | `github.com/google/uuid`, `internal/platform/jobs`; 6 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C10; no generated file authored here | Network Flow owner tests | medium | E02-043; D122 (declarations/imports and selected assertions); not executed |
| [graph_view_routes.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_routes.go) | Saved-graph HTTP commands, reads, refresh and contributors | private saved-graph route handlers/response projection | same-package reference candidates: `reporting_graph_source.go`, `routes.go`, `saved_graph_application.go`; +2 other source references | `internal/modules/incidents/admission`, `github.com/google/uuid`, `internal/modules/graphprojection`, `internal/platform/authn`, `internal/platform/httpapi`; 4 standard-library imports | source references: `graph_view_receipts_test.go` | C02, C10; no generated file authored here | Network Flow transport adapter | medium | E02-044; D339 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_store.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_store.go) | Saved declarations, generations, selected binding publication and invalidation | private store declaration methods | same-package reference candidates: `application.go`, `extension_state.go`, `graph_restore_source.go`; +7 other source references | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgtype`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`; 11 standard-library imports | source references: `extension_state_v3_integration_test.go`, `graph_result_cleanup_integration_test.go`, `graph_result_cleanup_race_integration_test.go`; +6 other source references | C10, C11; no generated file authored here | Network Flow persistence adapter | high | E02-045; D623 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_store_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_store_test.go) | Declaration persistence and generation-checked publication with rename | test-only `TestNetworkFlowGraphViewDeclarationPersistence_Integration`, `TestNetworkFlowGraphViewPublicationRejectsStaleGenerationAndAllowsRename_Integration` | Go test entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/testutil/revisionsupport`; 5 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C10; no generated file authored here | Network Flow owner tests | medium | E02-046; D227 (declarations/imports and selected assertions); not executed |
| [harness_audit_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_audit_integration_test.go) | Real committed audit consumers and exact count checks | test-only `assertNetworkFlowCommittedAuditConsumers` | network flow integration wrappers | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow/harnesscontrol`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/postgres`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `internal/testutil/networkflowsupport`, `internal/testutil/pgtest`, `github.com/google/uuid`; 8 standard-library imports | helper/bridge; network flow integration wrappers | C09, C14; no generated file authored here | Network Flow owner tests | medium | E02-047; D226 (declarations/imports and selected assertions); not executed |
| [harness_authorization_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_authorization_integration_test.go) | Real authority-transition consumers and resource/cursor rechecks | test-only `assertNetworkFlowAuthorizationConsumers` | network flow integration wrappers | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `internal/testutil/networkflowsupport`, `github.com/google/uuid`; 4 standard-library imports | helper/bridge; network flow integration wrappers | C09, C14; no generated file authored here | Network Flow owner tests | medium | E02-048; D159 (declarations/imports and selected assertions); not executed |
| [harness_randomness_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_randomness_integration_test.go) | Table entropy collision/retry transaction consumers | test-only `assertNetworkFlowTableEntropyTransactions` | network flow integration wrappers | `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/platform/httpapi`, `internal/testutil/httptestx`, `internal/testutil/revisionsupport`, `github.com/google/uuid`, `github.com/jackc/pgx/v5/pgconn`; 7 standard-library imports | helper/bridge; network flow integration wrappers | C03, C14; no generated file authored here | Network Flow owner tests | medium | E02-049; D97 (declarations/imports and selected assertions); not executed |
| [harness_worker_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_worker_integration_test.go) | Real worker and transaction fault consumers | test-only `assertNetworkFlowWorkerFaultConsumers` | network flow integration wrappers | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/platform/authn`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `github.com/google/uuid`; 5 standard-library imports | helper/bridge; network flow integration wrappers | C10, C14; no generated file authored here | Network Flow owner tests | medium | E02-050; D99 (declarations/imports and selected assertions); not executed |
| [harnesscontrol/control_json.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/control_json.go) | Bounded strict control JSON object decoding | private decodeControlRequest | same-package reference candidates: `harnesscontrol/network_flow_audit_assertion.go`, `harnesscontrol/network_flow_auth_transition.go`, `harnesscontrol/network_flow_fault.go`; +1 other source references | `internal/platform/strictjson`; 5 standard-library imports | indirect tests in C14; TODO: exact direct call-to-test mapping before editing this file | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-051; D42 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/control_json_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/control_json_test.go) | Malformed/duplicate/trailing/required-field control request matrix | test-only `assertStrictNetworkFlowControlRequests` | control route test wrappers | 4 standard-library imports | helper/bridge; control route test wrappers | C14; no generated file authored here | Testing Harness support tests | medium | E02-052; D33 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/controls.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/controls.go) | Instance registries and guarded harness contribution assembly | Controls; NewControls; Clear; Contribution | app/server/server_profile_harness.go (cartulary_harness build) | `internal/platform/harnessruntime`, `internal/platform/httpapi`; 0 standard-library imports | source references: `harness_audit_integration_test.go`, `harness_authorization_integration_test.go`, `harness_randomness_integration_test.go`; +2 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-053; D56 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/fault_consumers.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/fault_consumers.go) | Instance decorators for transaction/import/worker fault consumption | Controls.WrapDatabase/WrapImportFacade/ConfigureNetworkFlow | harness Controls composition; server NetworkFlowComposition | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/imports`, `internal/modules/networkflow`, `internal/platform/jobs`, `internal/platform/postgres`; 3 standard-library imports | source references: `application_composition_test.go`, `graph_materialization_payload_test.go`, `graph_materialization_timeout_test.go`; +1 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-054; D205 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_audit_assertion.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion.go) | Bound exact committed audit-count assertion registry and admission route | NetworkFlowAuditAssertionRegistry/Assertion/Scope; constructor; RegisterNetworkFlowAuditAssertionRoutes; fixture/consumption methods | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/network_flow_auth_transition.go`, `harnesscontrol/network_flow_fault.go`; +1 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 6 standard-library imports | source references: `harness_audit_integration_test.go`, `harness_authorization_integration_test.go`, `harnesscontrol/control_json_test.go`; +4 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-055; D398 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_audit_assertion_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion_integration_test.go) | Disabled/default, harness authorization, count arming, duplicates and clearing | test-only `TestNetworkFlowAuditAssertionRouteDisabledByDefault`, `TestNetworkFlowAuditAssertionRouteRequiresHarnessAuthorization`; +4 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 3 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-056; D218 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/network_flow_auth_transition.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_auth_transition.go) | Bound authority-transition registry and admission route | NetworkFlowAuthTransitionRegistry/NetworkFlowAuthTransition; constructor; route/consume/bind methods and closed tokens | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/network_flow_audit_assertion.go`, `harnesscontrol/network_flow_fault.go`; +1 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 6 standard-library imports | source references: `harness_audit_integration_test.go`, `harness_authorization_integration_test.go`, `harnesscontrol/control_json_test.go`; +4 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-057; D342 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_auth_transition_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_auth_transition_integration_test.go) | Disabled/default, admission, bound transition keys, invalid requests and clearing | test-only `TestNetworkFlowAuthTransitionRouteDisabledByDefault`, `TestNetworkFlowAuthTransitionRouteRequiresHarnessAuthorization`; +4 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 3 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-058; D197 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/network_flow_fault.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_fault.go) | One-shot correlation/boundary fault arming and guarded route | NetworkFlowFaultRegistry/Fault; constructor; RegisterNetworkFlowFaultRoutes; ConsumeNetworkFlowFault/For; boundary/effect tokens | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/fault_consumers.go`, `harnesscontrol/network_flow_audit_assertion.go`; +2 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 6 standard-library imports | source references: `harness_worker_integration_test.go`, `harnesscontrol/control_json_test.go`, `harnesscontrol/network_flow_audit_assertion_integration_test.go`; +3 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-059; D297 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_fault_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_fault_integration_test.go) | Disabled/default, admission, one-shot effects, duplicate arming and invalid requests | test-only `TestNetworkFlowFaultRouteDisabledByDefault`, `TestNetworkFlowFaultRouteRequiresHarnessAuthorization`; +4 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 3 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-060; D188 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/network_flow_randomness.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_randomness.go) | Table-ID UUID collision stream registry, strict arming and exhaustion | NetworkFlowRandomnessRegistry/NetworkFlowRandomnessState; constructor/route/consume methods; UUID tokens; exhaustion errors | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/network_flow_audit_assertion.go`, `harnesscontrol/network_flow_auth_transition.go`; +2 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 5 standard-library imports | source references: `harness_randomness_integration_test.go`, `harnesscontrol/control_json_test.go`, `harnesscontrol/network_flow_audit_assertion_integration_test.go`; +3 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-061; D289 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_randomness_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_randomness_integration_test.go) | Disabled/default, admission, deterministic collisions, retired token rejection and clearing | test-only `TestNetworkFlowRandomnessRouteDisabledByDefault`, `TestNetworkFlowRandomnessRouteRequiresHarnessAuthorization`; +6 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 5 standard-library imports | 8 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-062; D209 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/randomness_consumers.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/randomness_consumers.go) | Entropy adapter for an instance's armed randomness stream | Controls.TableIDEntropy and private adapter | harness Controls.ConfigureNetworkFlow | 3 standard-library imports | source references: `harness_randomness_integration_test.go` | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-063; D33 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/test_helpers_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/test_helpers_test.go) | HTTP test requests/body and assertion helpers for guarded control tests | test-only `testRuntimeEnabledEnv`, `testHTTPDependencies`; +7 other source references | harnesscontrol route tests | `internal/platform/httpapi`, `internal/testutil/httpapiextensions`; 5 standard-library imports | helper/bridge; harnesscontrol route tests | C14; no generated file authored here | Testing Harness support tests | medium | E02-064; D102 (declarations/imports and selected assertions); harness-only composition, F03 |
| [import_facade.go](/home/jochi/code/cartulary/internal/modules/networkflow/import_facade.go) | Imports owner binding, preview/result validation and transactional analytical apply | private importFacade implements imports.ExtensionImportFacade; Binding/PrepareImportUnitMapping/ValidateImportUnitMappingResult/ApplyImportUnitTx | application_composition.go; modules/imports/{mapping,application_operations}.go through ExtensionImportFacade; harness decorator | `github.com/jackc/pgx/v5`, `internal/modules/imports`, `internal/platform/httpapi`, `internal/platform/jobs`; 7 standard-library imports | network_flow_contract_test.go (binding/error assertions); network_flow_behavior_test.go (import helper wrappers); preview helper in network_flow_unit_test.go. Gap G01 | C03; no generated file authored here | Network Flow analytical target | high | E02-065; D439 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [import_owner_errors.go](/home/jochi/code/cartulary/internal/modules/networkflow/import_owner_errors.go) | Closed owner error union, validation and Imports translation | private facade TranslateImportUnitError/ValidateImportUnitError | same-package reference candidates: `import_facade.go` | `internal/modules/imports`; 2 standard-library imports | network_flow_contract_test.go: TestImportOwnerErrorTranslationUsesClosedRegisteredUnion | C03; no generated file authored here | Network Flow owner errors / Imports framing | medium | E02-066; D208 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link.go) | Typed explicit linking commands, selectors and owner participant coordination | private indicator-link application/outcomes | same-package reference candidates: `indicator_link_admission.go`, `indicator_link_http.go`, `indicator_link_receipts.go`; +1 other source references | `github.com/google/uuid`, `internal/modules/crossownertransaction`, `internal/modules/incidents/admission`, `internal/modules/indicators`, `internal/platform/authn`; 7 standard-library imports | source references: `indicator_link_boundary_test.go`, `network_flow_contract_test.go` | C07, C09; no generated file authored here | Network Flow linking / Indicators target | medium | E02-067; D470 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_admission.go) | Source row/target admission and canonical confirmation | private indicator-link admission | same-package reference candidates: `binding_store.go`, `indicator_link.go`, `indicator_link_graph_admission.go`; +4 other source references | `github.com/google/uuid`; 6 standard-library imports | source references: `indicator_link_boundary_test.go`, `indicator_link_fixture_test.go` | C07; no generated file authored here | Network Flow source semantics | high | E02-068; D300 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_boundary_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_boundary_test.go) | Public facade/participant boundary assertion for indicator linking | test-only `TestIndicatorLinkPublicBoundary` | Go test entrypoint | 6 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C07; no generated file authored here | Network Flow owner tests | medium | E02-069; D123 (declarations/imports and selected assertions); not executed |
| [indicator_link_fixture_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_fixture_test.go) | Test-only link store construction and request/result fixtures | test-only `decodeIndicatorLinkFixture` | network_flow_contract_test.go and route tests | `internal/platform/httpapi`; 1 standard-library imports | helper/bridge; network_flow_contract_test.go and route tests | C07; no generated file authored here | Network Flow owner tests | medium | E02-070; D16 (declarations/imports and selected assertions); not executed |
| [indicator_link_graph_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_graph_admission.go) | Graph selector/source admission for indicator linking | private graph-linked source validation | same-package reference candidates: `indicator_link_admission.go` | 1 standard-library imports | source references: `indicator_link_boundary_test.go` | C06, C07; no generated file authored here | Network Flow | high | E02-071; D102 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_http.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_http.go) | HTTP indicator-link decoding, command invocation and translation | private route handler/HTTP adapter | private package use; no distinct direct caller established in lexical map | `internal/platform/authn`, `internal/platform/httpapi`, `github.com/google/uuid`; 2 standard-library imports | indirect tests in C02, C07; TODO: exact direct call-to-test mapping before editing this file | C02, C07; no generated file authored here | Network Flow transport adapter | medium | E02-072; D49 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_receipts.go) | Indicator-link idempotent outcome replay and durable receipts | private indicator receipt adapter | same-package reference candidates: `indicator_link_http.go` | `github.com/jackc/pgx/v5`, `internal/modules/indicators`, `internal/platform/authn`, `internal/platform/extensionstore`, `github.com/google/uuid`; 6 standard-library imports | indirect tests in C07, C09; TODO: exact direct call-to-test mapping before editing this file | C07, C09; no generated file authored here | Network Flow | high | E02-073; D191 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [integration_surface_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/integration_surface_test.go) | Test-only aliases/constructors expose private module internals to integration package | test-only `parseKeyRingsWithDefaultRegistry`, `lifecycleStates`; +2 other source references | external-package networkflow integration tests | `internal/platform/authn`, `internal/platform/secretpurpose`, `internal/testutil/fixtures`; 2 standard-library imports | helper/bridge; external-package networkflow integration tests | C01, C15; no generated file authored here | Network Flow owner tests | medium | E02-074; D106 (declarations/imports and selected assertions); not executed |
| [keyring.go](/home/jochi/code/cartulary/internal/modules/networkflow/keyring.go) | Purpose-separated cursor/safe-digest keys, manifest validation and rotation | KeyRings; KeyRingsOverrideKey; KeyRingManifestMaximumSize; ParseKeyRingsWithRegistry | configuration.go; app/configassembly/deployment.go | `internal/platform/secretpurpose`; 12 standard-library imports | source references: `integration_surface_test.go`, `keyring_test.go`, `network_flow_contract_test.go` | C05, C16; no generated file authored here | Network Flow security contribution | high | E02-075; D403 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [keyring_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/keyring_test.go) | Key purpose separation, cursor rotation, inactive-epoch purge and invalid/null manifests | test-only `TestNetworkFlowKeyRingsAndCursorRotation`, `TestNetworkFlowSafeDigestRingPurgesInactiveEpoch`; +1 other source references | Go test entrypoints | `internal/platform/authn`; 6 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C05, C16; no generated file authored here | Network Flow owner tests | medium | E02-076; D218 (declarations/imports and selected assertions); not executed |
| [mapping.go](/home/jochi/code/cartulary/internal/modules/networkflow/mapping.go) | Closed source/mapping profiles and approved-mapping materialization | private mapping types/decoders and registry projections | same-package reference candidates: `csv_parser.go`, `import_facade.go`, `query_filter.go`; +1 other source references | `internal/gen/networkflowmapping`, `internal/platform/strictjson`; 4 standard-library imports | source references: `network_flow_contract_test.go`, `network_flow_unit_test.go`, `timestamp_test.go` | `gen/networkflowmapping`; C03 | Network Flow | medium | E02-077; D539 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [module.go](/home/jochi/code/cartulary/internal/modules/networkflow/module.go) | Coherent public module construction, active contribution admission and registration | Module; ModuleDependencies; ImportSourcePort; NewModule; RegisterRoutes/ImportOwner/RegisterGraphViewWorker/InstallCrossOwnerCoordinator/TransactionCapabilities | app/server/runtime_assembly.go; app/server/extension_admission.go; app/extensionassembly/cross_owner_transaction.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/gen/networkflowroutes`, `internal/modules/crossownertransaction`, `internal/modules/imports`, `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/postgres`; 6 standard-library imports | application_composition_test.go via production_surface_test.go / query_authoring_test.go | `gen/networkflowroutes`; C01, C02 | Network Flow facade | medium | E02-078; D249 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [names.go](/home/jochi/code/cartulary/internal/modules/networkflow/names.go) | Display-name derivation, Unicode normalization and collision naming | private naming functions | same-package reference candidates: `csv_parser.go`, `graph_view_store.go`, `import_facade.go`; +5 other source references | `internal/gen/networkflowunicode`; 3 standard-library imports | source references: `graph_view_receipts_test.go`, `network_flow_contract_test.go`, `network_flow_unit_test.go`; +2 other source references | `gen/networkflowunicode`; C03, C04 | Network Flow | medium | E02-079; D179 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [network_flow_behavior_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/network_flow_behavior_test.go) | Integration wrappers execute contract helpers for envelopes, imports, cursor invalidation, authorization, redaction and graph/link behavior | test-only `TestNetworkFlow_CoreEnvelopeAdmissionEvidence_Integration`, `TestNetworkFlow_ReservedSourceProfileEvidence_Integration`; +16 other source references | Go test entrypoints | 1 standard-library imports | 18 named Test entrypoint(s) in this file; assertions inspected | C02, C03, C04, C05, C06, C07, C09, C16; no generated file authored here | Network Flow owner tests | medium | E02-080; D78 (declarations/imports and selected assertions); not executed |
| [network_flow_contract_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/network_flow_contract_test.go) | Reusable behavioral assertion bodies plus closed Imports error translation and binding metadata tests | test-only `TestImportOwnerErrorTranslationUsesClosedRegisteredUnion`, `TestNetworkFlowImportFacadePublishesCompleteBinding` | network_flow_behavior_test.go / own entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/imports`, `internal/modules/incidents/admission`, `internal/platform/httpapi`; 10 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C02, C03, C04, C05, C06, C07, C09, C16; no generated file authored here | Network Flow owner tests | medium | E02-081; D720 (declarations/imports and selected assertions); not executed |
| [network_flow_unit_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/network_flow_unit_test.go) | Parser/mapping/resource/identity fixtures and reusable unit assertion helpers | test-only `TestNetworkFlow_CiscoSNARequiredFields_Unit`, `TestNetworkFlow_CSVParserEdgeOutcomes_Unit`; +12 other source references | Go test entrypoints / contract helpers | `github.com/google/uuid`, `internal/modules/graphprojection`; 14 standard-library imports | 14 named Test entrypoint(s) in this file; assertions inspected | C03, C05, C06, C15; no generated file authored here | Network Flow owner tests | medium | E02-082; D813 (declarations/imports and selected assertions); not executed |
| [operation_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/operation_admission.go) | Operation-policy-based current incident authorization | private operationAdmission; admitOperation | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `indicator_link_http.go`; +1 other source references | `internal/modules/incidents/admission`, `internal/platform/httpapi`, `internal/platform/httpauth`, `github.com/google/uuid`; 3 standard-library imports | indirect tests in C02, C09; TODO: exact direct call-to-test mapping before editing this file | C02, C09; no generated file authored here | Network Flow operations / Incidents admission authority | high | E02-083; D72 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [operation_admission_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/operation_admission_integration_test.go) | Real major-7 route admission precedence matrix | test-only `assertMajor7AdmissionMatrix` | table_lifecycle_integration_test.go and route integration wrappers | `internal/gen/networkflowroutes`, `internal/platform/authn`, `internal/platform/postgres`, `internal/testutil/httptestx`, `github.com/google/uuid`; 4 standard-library imports | helper/bridge; table_lifecycle_integration_test.go and route integration wrappers | `gen/networkflowroutes`; C02, C09 | Network Flow owner tests | medium | E02-084; D78 (declarations/imports and selected assertions); not executed |
| [portability_state.go](/home/jochi/code/cartulary/internal/modules/networkflow/portability_state.go) | Retained owner-state presence for incident portability | PortabilityStateQuery; PortabilityStateBinding; NewPortabilityStateBinding | app/extensionassembly/incident_portability.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 3 standard-library imports | source references: `configuration_test.go` | C12; no generated file authored here | Network Flow source contribution / Incident Portability | medium | E02-085; D56 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [ports.go](/home/jochi/code/cartulary/internal/modules/networkflow/ports.go) | Consumer-owned incident lock, audit and indicator participant ports | IncidentLockPort; AdministrativeAuditPort; IndicatorParticipationPort | module.go; store.go; app/server assembly | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/indicators`; 1 standard-library imports | indirect tests in C01, C07, C09; TODO: exact direct call-to-test mapping before editing this file | C01, C07, C09; no generated file authored here | Network Flow narrow ports | medium | E02-086; D24 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [production_surface_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/production_surface_test.go) | AST positive public API and private implementation surface guard | test-only `assertProductionSurface` | query_authoring_test.go | 6 standard-library imports | helper/bridge; query_authoring_test.go | C01; no generated file authored here | Network Flow owner tests | medium | E02-087; D154 (declarations/imports and selected assertions); not executed |
| [query.go](/home/jochi/code/cartulary/internal/modules/networkflow/query.go) | Typed accepted/rejected query decoding, sort/filter/cursor bindings and normalization | private query request/cursor/scope types | same-package reference candidates: `graph.go`, `graph_query_application.go`, `graph_source_composer.go`; +13 other source references | `internal/platform/strictjson`; 13 standard-library imports | source references: `indicator_link_fixture_test.go`, `network_flow_contract_test.go`, `network_flow_unit_test.go`; +2 other source references | C05; no generated file authored here | Network Flow | medium | E02-088; D922 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [query_authoring_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_authoring_test.go) | Direct read authority, public/semantic guards and filter normalization/error detail cases | test-only `TestQueryAuthoringAdmission_Unit` | Go test entrypoint | `internal/modules/incidents/admission`, `github.com/google/uuid`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C01, C05, C09; no generated file authored here | Network Flow owner tests | medium | E02-089; D78 (declarations/imports and selected assertions); not executed |
| [query_filter.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_filter.go) | Closed filter grammar, scalar/range normalization and diagnostic tokens | private filter admission helpers | same-package reference candidates: `query.go` | `internal/gen/contractnetworkflow`; 7 standard-library imports | source references: `query_authoring_test.go` | `gen/contractnetworkflow`; C05 | Network Flow | medium | E02-090; D310 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [query_sql.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_sql.go) | Parameterized source-owned row/diagnostic SQL with keyset pagination | private store QueryRowsPage/QueryRejectedDiagnosticsPage and SQL helpers | same-package reference candidates: `store.go`, `table_query_application.go` | `github.com/google/uuid`; 5 standard-library imports | source references: `network_flow_contract_test.go`, `query_sql_test_bridge_test.go` | C05; no generated file authored here | Network Flow persistence adapter | medium | E02-091; D409 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [query_sql_test_bridge_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_sql_test_bridge_test.go) | Test-only live SQL keyset boundary assertions | test-only `AssertLiveSQLKeysetBoundaries` | network_flow_contract_test.go | `github.com/google/uuid`; 4 standard-library imports | helper/bridge; network_flow_contract_test.go | C05; no generated file authored here | Network Flow owner tests | medium | E02-092; D50 (declarations/imports and selected assertions); not executed |
| [recovery_bindings.go](/home/jochi/code/cartulary/internal/modules/networkflow/recovery_bindings.go) | Owner authoritative-table inventory contribution | RecoveryPostgresTables | app/extensionassembly/recovery.go | 1 standard-library imports | indirect tests in C12; TODO: exact direct call-to-test mapping before editing this file | C12; no generated file authored here | Network Flow / Recovery registration | medium | E02-093; D24 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [recovery_state.go](/home/jochi/code/cartulary/internal/modules/networkflow/recovery_state.go) | Owner recovery family contribution | RecoveryStateContribution | app/recoveryassembly/state_catalog.go | `internal/platform/recoverystate`; 0 standard-library imports | indirect tests in C12; TODO: exact direct call-to-test mapping before editing this file | C12; no generated file authored here | Network Flow / Recovery registration | medium | E02-094; D13 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [reporting_graph_source.go](/home/jochi/code/cartulary/internal/modules/networkflow/reporting_graph_source.go) | Exact selected graph leasing, renewal/read, label conversion and release | ReportingGraphSource; Module.ReportingGraphSource; SourceOwnerID/ValidateAndLeaseResultTx/ReadAndRenewLeasedResultTx/ReleaseJobLeasesTx/ReleaseJobLeases | Module construction; app/server runtime composition; Reporting source interface | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/reporting/graphsourcecontract`, `internal/platform/postgres`; 6 standard-library imports | source references: `application_composition_test.go`, `graph_result_cleanup_race_integration_test.go`, `reporting_graph_source_integration_test.go` | C11, C12; no generated file authored here | Network Flow source / Reporting consumer / Graph Projection leases | medium | E02-095; D304 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [reporting_graph_source_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/reporting_graph_source_integration_test.go) | Exact lease/read/release lifecycle, scope and result stability | test-only `TestReportingGraphSourceValidatesLeasesReadsAndReleasesExactResult_Integration` | Go test entrypoint | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/modules/reporting/graphsourcecontract`, `internal/testutil/revisionsupport`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C11, C12; no generated file authored here | Network Flow owner tests | medium | E02-096; D229 (declarations/imports and selected assertions); not executed |
| [request_decoding.go](/home/jochi/code/cartulary/internal/modules/networkflow/request_decoding.go) | Transport-independent typed table and saved-graph request admission | private decode...RequestValue functions | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `routes.go` | 1 standard-library imports | source references: `network_flow_contract_test.go` | C02, C04, C06, C10; no generated file authored here | Network Flow semantics | medium | E02-097; D303 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [resources.go](/home/jochi/code/cartulary/internal/modules/networkflow/resources.go) | Owner table/row/diagnostic/binding/profile resource projection | private resource functions | same-package reference candidates: `graph.go`, `graph_query_presentation.go`, `graph_response_v2.go`; +8 other source references | 2 standard-library imports | source references: `configuration_test.go`, `graph_streaming_test.go`, `graph_temporal_test.go`; +2 other source references | C02, C03, C05, C07; no generated file authored here | Network Flow | medium | E02-098; D218 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [routes.go](/home/jochi/code/cartulary/internal/modules/networkflow/routes.go) | Table/profile HTTP adapters, session framing and safe response errors | private routeService; no public service constructor | same-package reference candidates: `application.go`, `graph_routes.go`, `graph_view_receipts.go`; +9 other source references | `github.com/google/uuid`, `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/httpauth`; 6 standard-library imports | source references: `network_flow_contract_test.go`, `table_lifecycle_test.go` | C02, C04, C05, C09; no generated file authored here | Network Flow transport adapter | medium | E02-099; D373 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [routes_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/routes_integration_test.go) | Claimed/unclaimed routes, query pagination, graph/indicator links, saved lifecycle and recovery composition | test-only `TestNetworkFlowPaginationRecovery_Integration`, `TestNetworkFlowRoutesRemainUnclaimedByDefault`; +6 other source references | Go test entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/app/recoveryassembly`, `internal/app/server`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/collaboration/protocol`, `internal/modules/graphprojection`, `internal/modules/graphprojection/restore`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/modules/recovery/restorecontract`, `internal/modules/reporting`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/jobs`, `internal/platform/postgres`, `internal/platform/renderprocess`, `internal/testutil/appsupport`, `internal/testutil/authcookietest`, `internal/testutil/collaborationsupport`, `internal/testutil/collaborationsupport/incidentwstest`, `internal/testutil/fixtures`, `internal/testutil/httptestx`, `internal/testutil/pgtest`, `internal/testutil/reportingfixture`; 12 standard-library imports | 8 named Test entrypoint(s) in this file; assertions inspected | C01, C02, C04, C05, C06, C07, C08, C10, C12; no generated file authored here | Network Flow owner tests | medium | E02-100; D2232 (declarations/imports and selected assertions); not executed |
| [saved_graph_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_admission.go) | Readiness/cutover validation of declarations, jobs, receipts and selected proofs | SavedGraphCutoverAlgorithmID; ValidateSavedGraphAdmission; ValidateRetainedExtensionState | app/configassembly/extension_admission.go; app/server/extension_admission.go | `internal/modules/graphprojection/postgresresult`, `internal/platform/authn`, `internal/platform/extensionstore`, `internal/platform/jobs`, `github.com/google/uuid`; 7 standard-library imports | source references: `extension_state_v3_integration_test.go`, `graph_materialization_payload_test.go`, `routes_integration_test.go`; +1 other source references | C10, C12; no generated file authored here | Network Flow retained-state admission | high | E02-101; D355 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [saved_graph_admission_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_admission_integration_test.go) | Retained state cutover admission and unchanged invalid bytes | test-only `TestSavedGraphCutoverAdmission_Integration` | Go test entrypoint | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/platform/authn`, `internal/platform/extensionstore`, `internal/platform/jobs`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`; 8 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C10, C12; no generated file authored here | Network Flow owner tests | medium | E02-102; D192 (declarations/imports and selected assertions); not executed |
| [saved_graph_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_application.go) | Typed create/refresh/rename/retire commands, capacity, jobs, audit and after-commit notification | private savedGraphApplication/outcomes | same-package reference candidates: `graph_view_routes.go` | `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/jobs`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 5 standard-library imports | source references: `graph_view_authority_test_bridge_test.go` | C09, C10; no generated file authored here | Network Flow | high | E02-103; D359 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [saved_graph_read_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_read_application.go) | Saved declaration/result/contributor reads and unavailable-result classification | private savedGraphReadApplication; reader ports | same-package reference candidates: `graph_query_application.go`, `graph_routes.go`, `graph_view_routes.go` | `internal/modules/graphprojection`, `internal/platform/jobs`, `github.com/google/uuid`; 4 standard-library imports | source references: `saved_graph_read_application_test.go` | C09, C10, C11; no generated file authored here | Network Flow | high | E02-104; D252 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [saved_graph_read_application_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_read_application_test.go) | Unavailable result classification distinguishes missing Jobs history from operational failure | test-only `TestSavedGraphReadFailures_Unit` | Go test entrypoint | `internal/platform/jobs`, `github.com/google/uuid`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C09, C10; no generated file authored here | Network Flow owner tests | medium | E02-105; D101 (declarations/imports and selected assertions); not executed |
| [saved_graph_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_receipts.go) | Saved command receipt adapter and immutable outcome serialization | private savedGraphReceiptAdapter/failure | same-package reference candidates: `graph_view_routes.go` | `internal/platform/authn`, `github.com/jackc/pgx/v5`; 5 standard-library imports | indirect tests in C10; TODO: exact direct call-to-test mapping before editing this file | C10; no generated file authored here | Network Flow | high | E02-106; D90 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [security.go](/home/jochi/code/cartulary/internal/modules/networkflow/security.go) | Authenticated opaque cursor encoding/binding/expiry and retired-key purge | private cursor codec/protector/payload | same-package reference candidates: `module.go` | `internal/platform/cryptography`; 9 standard-library imports | source references: `keyring_test.go`, `network_flow_contract_test.go` | C05, C09, C16; no generated file authored here | Network Flow cursor semantics | high | E02-107; D233 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_boundary_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_boundary_test.go) | Semantic import allowlist/root guard and safe failure assertions | test-only `assertSemanticBoundary` | query_authoring_test.go | 8 standard-library imports | helper/bridge; query_authoring_test.go | C01, C02, C03; no generated file authored here | Network Flow owner tests | medium | E02-108; D87 (declarations/imports and selected assertions); not executed |
| [semantic_failure.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_failure.go) | Closed private failures, safe typed details and wrapped causes | private semanticFailure/failureKind/failureDetails | same-package reference candidates: `graph.go`, `graph_query_application.go`, `graph_response_v2.go`; +11 other source references | `internal/modules/incidents/admission`; 2 standard-library imports | semantic_boundary_test.go via query_authoring_test.go | C02, C09; no generated file authored here | Network Flow | medium | E02-109; D106 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_http_adapter.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_http_adapter.go) | HTTP translation for semantic request admission and graph-stale failure | private networkFlowAPIError and HTTP decoder wrappers | same-package reference candidates: `graph_view_routes.go`, `operation_admission.go`, `routes.go` | `internal/platform/httpapi`; 1 standard-library imports | indirect tests in C02; TODO: exact direct call-to-test mapping before editing this file | C02; no generated file authored here | Network Flow transport adapter | medium | E02-110; D29 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_http_errors.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_http_errors.go) | Closed semantic failure to registered HTTP error mapping | private semanticHTTPError | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `indicator_link_http.go`; +3 other source references | `internal/platform/httpapi`; 2 standard-library imports | source references: `indicator_link_boundary_test.go`, `indicator_link_fixture_test.go`, `network_flow_unit_test.go`; +2 other source references | C02; no generated file authored here | Network Flow transport adapter | medium | E02-111; D138 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_reference_boundary_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_reference_boundary_test.go) | Transitive AST semantic reference guard, including aliases and reachable declarations | test-only `semanticReferenceViolations`, `assertSemanticReferences` | semantic_boundary_test.go | 9 standard-library imports | helper/bridge; semantic_boundary_test.go | C01, C02, C03; no generated file authored here | Network Flow owner tests | medium | E02-112; D174 (declarations/imports and selected assertions); not executed |
| [store.go](/home/jochi/code/cartulary/internal/modules/networkflow/store.go) | Owner table/row/diagnostic SQL, immutable inserts, lifecycle, limits and atomic audit/intents | private store and records; no exported store constructor | same-package reference candidates: `application.go`, `binding_store.go`, `graph_query_application.go`; +9 other source references | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgtype`, `internal/platform/postgres`; 8 standard-library imports | source references: `graph_streaming_test.go`, `network_flow_contract_test.go`, `routes_integration_test.go`; +2 other source references | C03, C04, C05, C08, C09; no generated file authored here | Network Flow persistence adapter | high | E02-113; D1125 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [store_fixture_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/store_fixture_test.go) | Test-only transaction wrappers around private store operations | test-only `CreateTable`, `RenameTable`; +2 other source references | store/lifecycle/route/worker integration tests | `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 1 standard-library imports | helper/bridge; store/lifecycle/route/worker integration tests | C03, C04; no generated file authored here | Network Flow owner tests | medium | E02-114; D51 (declarations/imports and selected assertions); not executed |
| [store_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/store_test.go) | Real persistence, immutable rows, active/retained counts, naming and intent rollback assertions | test-only `TestNetworkFlowResourceIntentFailureRollsBackSourceMutation_Integration`, `TestNetworkFlow_TableLifecycleVocabularyHasNoRenamedState_Unit`; +1 other source references | Go test entrypoints / helper assertions | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/app/indicatorassembly`, `internal/app/projectionassembly`, `internal/modules/auth/testsupport/storetest`, `internal/modules/incidents`, `internal/modules/incidents/admission`, `internal/modules/incidents/testsupport/storetest`, `internal/modules/indicators`, `internal/modules/networkflow`, `internal/modules/records`, `internal/modules/reference_data/testsupport`, `internal/modules/revisions`, `internal/platform/authn`, `internal/platform/postgres`, `internal/testutil/collaborationsupport`, `internal/testutil/networkflowsupport`, `internal/testutil/revisionsupport`; 6 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C03, C04, C08, C09; no generated file authored here | Network Flow owner tests | medium | E02-115; D584 (declarations/imports and selected assertions); not executed |
| [table_http.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_http.go) | Table typed outcome to HTTP and route command adapter | private tableOutcomeHTTP and commit route wrappers | same-package reference candidates: `routes.go` | `internal/modules/incidents/admission`, `internal/platform/httpapi`, `github.com/google/uuid`; 2 standard-library imports | indirect tests in C02, C04; TODO: exact direct call-to-test mapping before editing this file | C02, C04; no generated file authored here | Network Flow transport adapter | medium | E02-116; D29 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [table_lifecycle_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_lifecycle_integration_test.go) | Current admission before replay, serialized lifecycle and historical response invariants | test-only `TestTableLifecycleAdmissionAndReplay_Integration` | Go test entrypoint | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/platform/authn`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 8 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C04, C09; no generated file authored here | Network Flow owner tests | medium | E02-117; D191 (declarations/imports and selected assertions); not executed |
| [table_lifecycle_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_lifecycle_test.go) | Name control precedence and mutation digest assertions | test-only `TestTableLifecycleNameControlPrecedence_Unit`, `TestTableLifecycleMutationDigest_Unit` | Go test entrypoints | 4 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C04; no generated file authored here | Network Flow owner tests | medium | E02-118; D36 (declarations/imports and selected assertions); not executed |
| [table_query_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_query_application.go) | Read admission, active table scope, rows/diagnostics and cursor continuation | private tableQueryApplication and typed outcomes | same-package reference candidates: `graph_query_application.go`, `graph_source_composer.go`, `graph_view_jobs.go`; +5 other source references | `internal/modules/incidents/admission`, `github.com/google/uuid`; 5 standard-library imports | source references: `network_flow_contract_test.go`, `query_authoring_test.go` | C05, C09; no generated file authored here | Network Flow | high | E02-119; D425 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [table_query_presentation.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_query_presentation.go) | Accepted row and rejected diagnostic response projection | private acceptedRowsResource/diagnosticsResource | same-package reference candidates: `routes.go` | 0 standard-library imports | indirect tests in C05; TODO: exact direct call-to-test mapping before editing this file | C05; no generated file authored here | Network Flow | medium | E02-120; D32 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [table_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_receipts.go) | Table command replay/save and stable response serialization | private tableReceiptAdapter/storedMutationReceipt | same-package reference candidates: `table_http.go` | `internal/platform/authn`, `github.com/jackc/pgx/v5`; 3 standard-library imports | indirect tests in C04; TODO: exact direct call-to-test mapping before editing this file | C04; no generated file authored here | Network Flow | high | E02-121; D43 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [textcanon_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/textcanon_test.go) | Pinned Unicode canonicalization assertions | test-only `TestUnicode17TextCanonicalization` | Go test entrypoint | `internal/gen/networkflowunicode`; 1 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | `gen/networkflowunicode`; C03, C04 | Network Flow owner tests | medium | E02-122; D38 (declarations/imports and selected assertions); not executed |
| [timestamp.go](/home/jochi/code/cartulary/internal/modules/networkflow/timestamp.go) | Exact timestamp variants, epoch/RFC3339/uptime parsing and DST ambiguity checks | private timestamp profile/error/decoder helpers | same-package reference candidates: `csv_parser.go`, `mapping.go`, `query_filter.go` | 6 standard-library imports | source references: `network_flow_unit_test.go`, `timestamp_test.go` | C03; no generated file authored here | Network Flow | medium | E02-123; D314 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [timestamp_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/timestamp_test.go) | Closed timestamp variants, precise grammar, timezone transitions and uptime ordinals | test-only `TestTimestampProfileExactGrammarPrecisionAndZoneTransitions`, `TestTimestampProfileClosedJSONVariantsRejectNullMissingAndCrossVariantMembers`; +1 other source references | Go test entrypoints | 4 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C03; no generated file authored here | Network Flow owner tests | medium | E02-124; D122 (declarations/imports and selected assertions); not executed |
| [timezone.go](/home/jochi/code/cartulary/internal/modules/networkflow/timezone.go) | Pinned timezone lookup | private loadPinnedTimezone | same-package reference candidates: `timestamp.go` | `internal/gen/networkflowtz`; 6 standard-library imports | indirect tests in C03; TODO: exact direct call-to-test mapping before editing this file | `gen/networkflowtz`; C03 | Network Flow using generated timezone data | medium | E02-125; D51 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [transaction.go](/home/jochi/code/cartulary/internal/modules/networkflow/transaction.go) | Owner-local transaction begin/rollback/commit lifecycle | private withinTransaction | same-package reference candidates: `application.go`, `graph_query_application.go`, `graph_view_jobs.go`; +1 other source references | `github.com/jackc/pgx/v5`, `internal/platform/postgres`; 2 standard-library imports | transaction_test.go | C04, C09, C10; no generated file authored here | Network Flow persistence adapter | high | E02-126; D25 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [transaction_participants.go](/home/jochi/code/cartulary/internal/modules/networkflow/transaction_participants.go) | Import apply and explicit indicator-link participants in cross-owner transaction | ImportApplyParticipantID; IndicatorLinkParticipantID; private transactionCapability/participant | same-package reference candidates: `import_facade.go`, `indicator_link.go` | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/crossownertransaction`, `internal/modules/imports`, `internal/modules/incidents/admission`, `internal/modules/indicators`, `internal/platform/authn`; 6 standard-library imports | indirect tests in C03, C07, C09; TODO: exact direct call-to-test mapping before editing this file | C03, C07, C09; no generated file authored here | Network Flow participant / crossownertransaction coordinator | high | E02-127; D299 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [transaction_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/transaction_test.go) | Begin/commit/rollback lifecycle probe cases | test-only `TestTransactionLifecycle` | Go test entrypoint | `github.com/jackc/pgx/v5`, `internal/platform/postgres`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C09; no generated file authored here | Network Flow owner tests | medium | E02-128; D123 (declarations/imports and selected assertions); not executed |
| [v3_contract_projection_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/v3_contract_projection_test.go) | Current major-7 generated graph contract projection assertions despite historical filename | test-only `TestNetworkFlowV7GraphContractProjection_Unit` | Go test entrypoint | `internal/gen/contractgraphprojection`, `internal/gen/contractnetworkflow`, `internal/gen/networkflowroutes`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | `gen/contractgraphprojection`, `gen/contractnetworkflow`, `gen/networkflowroutes`; C13 | Network Flow owner tests | medium | E02-129; D223 (declarations/imports and selected assertions); not executed |

## 3. Module Boundary Diagnosis

The target is a **legitimate source-owner module with a small application/service facade**, private query/command orchestration, transport adapters and persistence adapters. It is a mutation coordinator and a flow-specific graph orchestration layer. Its breadth follows the adopted Network Flow scope; the directory's existence alone did not establish that conclusion. It is not a frontend controller or grid-vendor layer. No evidence supports replacing the existing facade or moving flow semantics wholesale into generic projections, Imports, Indicators, or Core saved views.

The main exception is a narrow transport dependency in import preview validation (F01). The harness subpackage adds a distinct testing responsibility, but its runtime isolation and security-scan accounting make relocation a deferred architectural question (F03), not an implementation decision.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Coherent facade and retained applications | `module.go`, `application_composition.go`, private application files | Network Flow | keep | E01-NF §29.2; E02; E03 | Active bundle validation, one DB/clock/port set, single-assignment coordinator and error-returning ImportOwner already exist. No new facade required. |
| Strict preview-result object admission through HTTP helper | `import_facade.go:validatePreviewDiagnosticShape` | NF owns semantic admission; platform/strictjson owns syntax decoding | split | E01-NF §29; E04; F01 | Remove only HTTP-helper dependency in a later slice; do not move owner validation into generic Imports or change rejection behavior. |
| CSV, mappings, immutable rows and rejected diagnostics | `csv_parser.go`, `mapping.go`, `digest.go`, `timestamp.go`, `store.go` | Network Flow analytical import target | keep | NF-REQ-001/002; Imports facade call sites | Imports owns upload/session/unit workflow; NF owns analytical target semantics. |
| Table/read command coordination and local SQL | `application.go`, `table_query_application.go`, `query_sql.go`, `store.go`, `transaction.go` | Network Flow, behind existing facade/private adapters | keep | NF-REQ-210; internal/AGENTS; E02 | SQL/pgx presence is not itself a boundary defect. Preserve transaction ownership and source-owned relations. |
| Flow graph semantics and source composition | `graph*.go`, `graph_source_composer.go` | Network Flow; generic projection engine remains Graph Projection | keep | E01-NF §§14,26,29; native V2 adapter; E05 | Shared route/worker/restore composer is already present. No generic projection engine migration into NF. |
| Saved graph declarations, lifecycle and refresh | `saved_graph_*.go`, `graph_view_*.go` | Network Flow; Jobs/Graph own their substrates | keep | E01-NF §§19.1,28,29; E02/E05 | These are not Core saved views. Versions, generations and selected bindings have distinct meanings. |
| Exact result leases, cleanup and restore contributions | cleanup, reporting and restore files | NF source contributions; Graph Projection persistence/leases; Recovery/Reporting orchestration | keep | E01-NF §29; E03/E05 | Preserve exact result after declaration refresh/retirement and all lease scopes. |
| Canonical indicator creation/reuse plus source binding | indicator link, binding store and transaction participant files | Indicators owns canonical target; NF owns source selector/binding | keep | E01-C02 REQ-02-074C/075; E02/E04 | No analytical-row-to-Core-record or observation conversion. |
| Durable resource invalidation production | `collaboration_producer.go`; server translator | NF semantic event producer; Collaboration persistence/delivery | keep | E01-C03; E03/E06 | Owner transaction appends intent; generic WebSocket sequence/auth/resume stays outside NF. |
| HTTP authentication/framing/error translation | route files, table/indicator HTTP adapters, semantic HTTP adapters | NF transport adapters using platform HTTP and Incidents admission | keep | E01-NF NF-REQ-210; E01-C04; E02 | Reusable semantic code must remain independent. Do not relocate authorization authority into decoders. |
| Fault/randomness/authority/audit controls | `harnesscontrol/*`; harness-tagged server assembly; `internal/testutil/networkflowsupport` | Testing Harness support; exact package destination unresolved | defer | E01-HAR §12.2.9; E07; F03 | Candidate shared test support requires preserving runtime security scanning and guarded owned-process consumers. |
| Browser shell/controller and grid integration | `apps/web/src/networkFlow`; protocol and UI adapter consumers | Existing frontend feature owners, shell, Collaboration session and grid adapter | keep | E08; E01-NF NF-REQ-004 | No frontend code lives in the target. No frontend redesign slice is supported by this backend finding. |
| Timeline, evidence, links, Core records/revisions/view schemas | No owned implementation found in this target's inspected source surfaces | Existing Core owners | defer | NF-REQ-002; Core 01/02; target imports and route inventory | Not applicable to the proposed seam. Absence of a direct import is not proof about all transitive behavior. No move or API is proposed. |

## 4. Public Contract and Behavior Freeze Map

The freeze preserves owner-aligned behavior. It does not promote a latent defect into a requirement. If characterization exposes an owner mismatch, record it separately; any behavior correction **requires later authorization**, including the owner clause, compatibility implications and new acceptance evidence.

All “existing tests” below mean source/selected assertions inspected. None was executed in this planning task. C03/G01 is the concrete missing characterization gate for S02. Other rows require preserving their named tests if later slices expand into those responsibilities.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| C01 — Public Go facade, active/inactive construction and contribution acquisition | Network Flow; Core Extensions publishes claims | E01-NF NF-REQ-210–212; module/composition and application assembly | `query_authoring_test.go` invokes `production_surface_test.go`; `application_composition_test.go` probes retained applications; routes test unclaimed default | Preserve construction/no-IO/API guard. Deepen only if dependencies or registrations change. | high for redesign; low for S02 | No supported public symbol/type change planned. |
| C02 — 19 HTTP operations; request/success/error envelopes, omission/defaults, admission precedence, safe error detail and session handling | NF route semantics; Core HTTP/auth/Incidents authority | E01-NF §§17,21,29; E01-C04 REQ-04-105A; E02 route/admission/error files; E09 route registry | `routes_integration_test.go`; `operation_admission_integration_test.go:assertMajor7AdmissionMatrix`; `network_flow_contract_test.go:AssertJSONAdmissionAndErrorDetails`; graph completion unit probes | Preserve current method/path/status/envelope and multi-invalid-field precedence assertions. No new route tests required solely for replacing a forwarding decoder. | high | Full route inventory below; invalid preview result still becomes Imports `owner_preview_validation_failed`. |
| C03 — Import binding/preview/apply, exact mapping fingerprint and diagnostics, immutable imported rows and replay | Imports workflow and transaction framework; NF analytical target | NF §§7–12,18,29; E04; binding/schema artifacts in E09 | `TestNetworkFlowImportFacadePublishesCompleteBinding` checks metadata; `TestImportOwnerErrorTranslationUsesClosedRegisteredUnion`; CSV/diagnostic vectors; behavior wrappers for source change/replay | **G01/S01:** direct diagnostic raw-object rejection matrix plus public facade validation of a valid prepared result, count/mapping/fingerprint errors and preserved closed error strings. Preserve existing import integration helpers. | high; direct S02 contract | No direct targeted `ValidateImportUnitMappingResult`/diagnostic-shape test found by symbol search in NF/Imports. This is a coverage finding, not proof of total absence of indirect tests. |
| C04 — Table rename/soft-delete admission before replay, immutable receipts, no-op/version behavior, retained-name/limit semantics | Network Flow | `application.go`, `table_receipts.go`, `store.go`; NF §§28–29.1 | `TestTableLifecycleAdmissionAndReplay_Integration`; name/digest unit tests; store lifecycle tests | Preserve replay after lifecycle change, conflicting transaction bodies, version errors, audit/intent counts and rollback when any participant fails. | high | Flow rows are immutable; table metadata mutation is not Core record patch/revision behavior. |
| C05 — Accepted/rejected row queries, filter/sort/keyset semantics, opaque cursors and revocation | Network Flow; Incidents current authority | NF §§12–13; query/filter/SQL/security/read app files | `TestQueryAuthoringAdmission_Unit`; contract helper keyset/cursor/crypto assertions; `TestNetworkFlowPaginationRecovery_Integration`; cursor invalidation wrapper | Preserve equal sort values/null ordering, active-table snapshots, expired/tampered cursor rejection and current authority/lifecycle rechecks. | high | Core entity/view row and test-util mutation/save-state contracts are not alternate owners for NF immutable rows. |
| C06 — Native Graph V2 identity, temporal buckets, limits, selectors, contributors and query completion | NF flow source semantics; Graph Projection generic engine | NF §§14,26–29; E05; generated graph semantics | Streaming/default identity tests; temporal arithmetic/conservation/digest tests; `TestGraphQueryCompletion_Unit`; route graph contributors tests | Preserve deterministic IDs/digests/counter strings, bounded aggregate state, safe cancellation/provider failure and one complete outcome/audit. | high | No projection refresh or graph schema redesign. |
| C07 — Explicit indicator link admission, canonical target visibility, duplicate/replay and atomic binding | NF source binding; Indicators canonical identity/participant; cross-owner transaction coordinator | NF §15; Core 02 REQ-02-074C/075; ports/participants/link stores | `TestIndicatorLinkPublicBoundary`; `AssertIndicatorLinkContractBoundary`; graph/link routes and existing-indicator behavior wrapper | Preserve current target visibility on replay, source selector validation, dedupe, one transaction and empty observation refs where no Core source record exists. | high | Not a generic links module responsibility; no automatic observation creation. |
| C08 — Durable `extension_resource_changed` intent and UI invalidation | NF producer; Collaboration durable stream/WebSocket lifecycle; frontend interpreter | `collaboration_producer.go`; server translator; E08 interpreter; Core 03 | `TestNetworkFlowResourceIntentFailureRollsBackSourceMutation_Integration`; route resource-change helpers; frontend extension-event tests located | Preserve exact payload/profile/resource/change/reason, stable intent identity, transaction rollback and no extra replay event. Before changing wire delivery, inspect Collaboration route/session tests beyond this plan. | high | NF registers no WebSocket route. Do not invent a NF socket path or migrate sequence/auth/resume into it. |
| C09 — Authorization, atomic audit, commit ordering and transaction capabilities | Core Incidents/auth/audit/coordinator; NF operation policy and source changes | E01-C04; operation admission, ports, transaction helpers/participants | Admission matrix, `TestTransactionLifecycle`, deployment-admin-no-bypass and redaction helpers; real auth/audit harness consumers | Preserve current admission before replay, fresh transaction checks, actor/incident attribution, no extra no-op/replay audit and failure rollback. | high | Privileged deployment role is not an incident-access bypass. Fixture SQL is confined to test support, not production authority. |
| C10 — Incident-shared saved graph declaration lifecycle, refresh/jobs, generation/version split, selected binding and retained admission | Network Flow; Jobs/finalizer substrate | NF §§19.1,27–29; saved graph commands/receipts/reads/store/jobs | Saved lifecycle/time-bucket routes; publication-stale-generation test; submitter-role revalidation; materialization timeout/payload; `TestSavedGraphReadFailures_Unit`; cutover admission tests | Preserve admission/replay order, postcommit notification, stale publication rejection, selected/desired snapshot identity, strict payload and safe missing-history classification. | high | Separate from Core saved views and view-schema persistence; no compatibility decoder or migration proposed. |
| C11 — Exact results, owner-scoped leases and bounded cleanup safety | Graph Projection storage/leases; NF declaration selection and source contribution | NF §29; cleanup files; reporting source; E05 | Owner-scoped selected-safe cleanup, publication/lease race tests, dispatcher tests and `TestReportingGraphSourceValidatesLeasesReadsAndReleasesExactResult_Integration` | Preserve lookup→renew→exact read, expiry/races/rollback, lease owner `snapshot_reporting`, purpose `render`, resource job ID; scoped release and valid empty release. | high | Refresh/retire cannot redirect a previously leased exact result. |
| C12 — Recovery/portability and authoritative extension state | NF source facts; Recovery/Incident Portability/Extensions orchestration; Graph restore | NF §§27–29; recovery files; E03/E05 | Obsolete semantic bytes rejected without rewrite; saved-state admission; routes restore composition; payload/startup helpers; reporting lease tests | Before movement, prove real restore source use, sorted pagination at 256 boundaries, entire borrowed-transaction rollback, defaults/no observer and selected+desired snapshot equality. Existing helpers are not a claim of exhaustive coverage. | high | Five source tables: graph views, indicator bindings, rejected diagnostics, rows, tables. Incompatible retained bytes remain unchanged. |
| C13 — Generated protocol, mapping, route, presentation and frontend entrypoint surfaces | Authored NF contracts projected by repository generators | E09 index/policy/topology; generated roots; current major-7 projection test | `TestNetworkFlowV7GraphContractProjection_Unit`; protocol compile surface; frontend boundary policy | Preserve authored contract bytes and generated outputs for S01/S02. Run canonical drift check only in later implementation. | high | No hand edits to generated files; old filenames such as `v3_contract_projection_test.go` do not define runtime major. |
| C14 — Four guarded test control surfaces and real consumers | Testing Harness; owner-local support and harness-tagged server | E01-HAR §12.2.9; E07; harnesscontrol files | Four route suites verify default disabled/auth/arming/rejection; root harness integration helpers exercise real operations | Before any relocation, preserve strict controls, consumption/correlation, owned-process crash, current DB authority, committed audit counts, security scan inclusion and default absence. | high | Randomness schema observed as v3; other control schemas observed as v2. Do not infer versions from family labels. |
| C15 — Test rows, owner slices, evidence routing and runtime/support scan accounting | Testing Harness/tooling accounting | E07/E09; 99 module.networkflow rows and 65 web.networkflow rows | Existing tests linked by manifest selectors; source assertions inspected, no run artifacts created | Keep existing selectors/rows for subtest additions; if new top-level selectors are needed, inspect authored owner inputs and regenerate only in an authorized implementation task. | medium | Phase/family maps are evidence accounting, not runtime boundaries. Capacity test presence does not authorize Core 05 claims. |
| C16 — Configuration, pinned canonicalization, key purposes, safe telemetry and raw-value privacy | NF source configuration/security/telemetry; platform secret/telemetry contracts | configuration/keyring/digest/timestamp/timezone/telemetry files; E01-NF §§16,20,29 | Configuration/effective-limit, key-ring rotation/purpose, Unicode/timestamp and telemetry containment tests; redaction helper | Preserve key separation, limit discovery, stable canonical bytes, non-disclosure of raw causes/secrets and observer failure isolation. | high | No config knobs, key rotation, fixture or generated timezone changes proposed. |
| C17 — Frontend selectors, save-state/controller lifetime and grid adapter consumption | Frontend NF feature; workbook shell/Collaboration session/grid adapter; NF browser contract | E08; NF NF-REQ-004 and browser workflows | Frontend boundary policy assertions inspected; controller/extension-event/selector tests located in owner map, not all assertions audited | No UI movement required. If later scope expands, inspect affected controller/event and browser assertions first and reselect web owner verification. | medium | No direct grid vendor code found in the Go target. NF presentation contracts are not Core view schemas. |

C02 route inventory: prepend `/api/v1/incidents/{incident_id}/network-flow` to each suffix.

| Method | Route suffix | Current handler family |
| --- | --- | --- |
| GET | `/source-profiles` | source profile read |
| GET | `/tables` | table list |
| GET | `/tables/{network_flow_table_id}` | table get |
| PATCH | `/tables/{network_flow_table_id}` | table rename |
| DELETE | `/tables/{network_flow_table_id}` | table soft-delete |
| POST | `/tables/{network_flow_table_id}/query` | accepted rows |
| POST | `/tables/{network_flow_table_id}/rejected-rows/query` | diagnostics |
| POST | `/rows/query` | cross-table accepted rows |
| POST | `/graphs/query` | graph query |
| POST | `/graphs/contributors/query` | graph contributors |
| GET | `/graph-views` | saved graph list |
| POST | `/graph-views` | saved graph create |
| GET | `/graph-views/{graph_view_id}` | saved graph get |
| PATCH | `/graph-views/{graph_view_id}` | saved graph rename |
| DELETE | `/graph-views/{graph_view_id}` | saved graph retire |
| POST | `/graph-views/{graph_view_id}/refresh` | saved graph refresh |
| GET | `/graph-views/{graph_view_id}/result` | selected result |
| POST | `/graph-views/{graph_view_id}/contributors/query` | saved result contributors |
| POST | `/indicator-links` | explicit indicator link |

Imports also calls the facade from its own mapping approval and `/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview` workflow (E04/E09); this is not a twentieth NF route. C14 consists of POST routes under `/api/v1/test/runtime/`: `network-flow-faults`, `network-flow-randomness`, `network-flow-auth-transitions`, and `network-flow-audit-assertions`. They are available only through admitted harness composition.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| F01 — Preview diagnostic validation reaches transport package solely for strict object decoding | `import_facade.go:validatePreviewDiagnosticShape` calls `httpapi.DecodeStrictJSONObject`; `platform/httpapi/json_object.go` simply delegates to `strictjson.DecodeObject`; NF §29 forbids transport helpers in reusable semantics | medium dependency leak; high if validation behavior accidentally changes | must_fix | NF semantic validation + platform/strictjson | Scoped to S02 acceptance: replace dependency after G01 characterization, preserving decoder and all local result/error semantics. |
| F02 — Semantic boundary roots omit Imports facade/error adapter | `semantic_boundary_test.go` explicit roots and application glob exclude `import_facade.go` and `import_owner_errors.go`; transitive AST checker is already present | recurrence of F01 | should_fix | NF boundary tests | Extend existing semantic root guard in S02; inspect reachable functions; do not relax the guard to allow HTTP. |
| F03 — Harness controls are a distinct owner-local support responsibility; relocation changes security accounting | `test_support_inventory.json` declares owner_local/runtime_scan included; harness-tagged server import; backend allowlist; gosec test requires harnesscontrol not excluded, while internal/testutil is excluded | high if moved mechanically | defer | Testing Harness, with NF-specific consumers | RB-002: design preserved scan/build/consumer coverage before choosing destination. No move slice is approved. |
| F04 — Large package contains legitimate source state, adapters and orchestration behind an existing facade | E01-NF NF-REQ-001/002/210/211; module/application/ports; production surface guard | high for speculative redesign | intentional/no_action | Network Flow | Keep module boundary. Framework facade-extraction examples are not proof that this repository still needs extraction. |
| F05 — SQL, pgx and transaction helper live within owner-local private adapters | store/query SQL/transaction and participant bodies; internal instructions permit owner adapters | high if transaction/lock ownership changes | intentional/no_action | Network Flow persistence and caller transaction ports | Do not move SQL to generic platform/domain helpers merely to reduce imports. No shared-helper extraction has semantic justification here. |
| F06 — Graph, Reporting and restore integrations are narrow source-owner contributions | native V2 adapter, one source composer, reporting exact leases, restore registration; E03/E05 | high identity/lease/publication risk | intentional/no_action | NF source; Graph/Reporting/Recovery respective owners | Preserve contribution contracts and separate defaults/observer behavior by consumer. |
| F07 — Mutation effects cross owners explicitly through ports and app translators | NF ResourceIntentAppender, Indicators participation and coordinator, Jobs finalizer; source mutation rollback test | high atomicity/auth risk | intentional/no_action | Source owner + Collaboration/Indicators/Jobs | Preserve transaction boundaries, audit/intent identity and after-commit notification. No hidden Core revision/change-set rewrite discovered in this seam. |
| F08 — Generated artifacts are projections, not edit targets or authority | generated-artifact policy/topology; contracts/network-flow index; protocol entrypoint | high drift/hand-edit risk | intentional/no_action | Contract authors/generators | S01/S02 expect zero contract/generated changes. Reassess if any arise; use authored inputs and Make only under later authorization. |
| F09 — No duplicated Core view-schema or grid-vendor implementation established in target | target imports/source surfaces, NF scope and frontend contract adapter | medium risk of conflating ownership | intentional/no_action | NF analytical contracts; existing view-schema/grid owners | Keep Core saved views/records and NF saved graphs/immutable rows distinct. No frontend/grid slice. |
| F10 — Historical filenames/family labels differ from current runtime version | `v3_contract_projection_test.go` asserts V7; extension-state “v3” file asserts V4 rejection behavior; saved cutover ID remains v6 | medium if names drive design | intentional/no_action | NF compatibility owner | Record repository/framework mismatch: current major7/state4/V2 source contracts supersede historical plan naming. Do not rename for cosmetic consistency. |
| F11 — Preview validator's exact rejection matrix is not directly characterized in located tests | binding metadata and owner-error tests inspected; targeted symbol search; Imports callers depend on validation outcome | medium structural-change uncertainty | must_fix | NF tests | G01/RB-001 gates S02; S01 adds meaningful characterization, not tests mirroring implementation structure. |
| F12 — Required Markdown check has write side effects and default globs omit this tracker | Make prerequisite/run-step chain; markdownlint scripts; `.markdownlint-cli2.jsonc` enumerates selected handoffs but not this file | documentation verification limitation | defer | Documentation tooling | V08/RB-003: do not execute incompatible command or edit tooling in this task; record skipped required check and supported-selection/redirection discovery need. |

No adopted primary-owner contradiction was found in the bounded review. Generic “move test helpers to test-util” guidance versus explicit current owner-local harness/security policy is a planning tension recorded in F03, not proof of conflicting product owners. No claim is made that all owner documents or runtime behavior were exhaustively reconciled.

## 6. Refactor Workstreams

These are planning workstreams. The DONE entries in §9 mean their documentary deliverables exist, not that a refactor or verification ran. Successors below are the inverse dependency edges.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source/bootstrap and tracker admission | root | none | WF-01 | Establish scope, baseline and owner posture | framework, AGENTS, owner documents, tracker | Read-only identity/Git/source checks | Canonical target/output admitted; only tracker write authorized |
| WF-01 | File-level inventory | chain | WF-00 | WF-02 | Account for all 129 source/test files and direct interfaces | E02 target files; E03–E08 relevant callers | Enumeration reconciled to §2; depth explicit | 129 unique rows, no silent exclusions |
| WF-02 | Contract-owner map | chain | WF-01 | WF-03, WF-04 | Separate required scope from observed implementation | E01 owners; routes/contracts and consumers | C01–C17 owner/test posture reviewed | Every discovered risk assigned owner and applicability |
| WF-03 | Characterization gap analysis | parallel | WF-02 | WF-05, WF-07 | Identify concrete pre-move evidence gaps | import facade, contract tests, semantic guard; family maps | G01 plus preserved tests; V01/V02 discovery | S01 entry/exit and RB-001 explicit; no fake execution |
| WF-04 | Boundary/coupling diagnosis | parallel | WF-02 | WF-05 | Judge transport, SQL, contributor and harness placement | E02/E03/E04/E05/E07 | F01–F12 traced to source | Narrow seam selected; harness relocation deferred |
| WF-05 | Ownership/facade recommendation | chain | WF-03, WF-04 | WF-06 | Keep supported facade; choose only justified structural change | module, composition, import facade, guard | NF §29; freeze map | No facade redesign, API change or generic-module migration assumed |
| WF-06 | Small-slice sequencing | chain | WF-05 | WF-07, WF-08 | Separate characterization from dependency cleanup | S01/S02 files | Entry gates, commands, rollback and binary exits | Each slice behavior-preserving and readiness stated |
| WF-07 | Harness/accounting plan | chain | WF-06, WF-03 | WF-08 | Retain selector coverage and security scan posture | family manifests, test support inventory, boundary policy | E07/E09; V01–V07 | Existing rows reusable; no harness edits now; F03 deferred |
| WF-08 | Documentation audit and final handoff | chain | WF-06, WF-07 | none | Reconcile rows/dependencies/traceability and preserve baseline | tracker only | §12 document audit; final Git/content/index comparison; V08 skip recorded | Planning artifact current, limits explicit and only tracker changed |

## 7. Proposed Refactor Slice Plan

These are future actions, **not implementation patches**. The current authorized output is the tracker. S01 is ready only as a later test-characterization task after source revalidation. S02 is **blocked: missing characterization** (RB-001), even though the wrapper's forwarding body makes behavior preservation plausible. No state, transport route, API/type, generated contract, migration, frontend, or harness relocation slice is proposed.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S01 | WF-08 handoff; later implementation authorization and baseline revalidation | Characterize existing preview-result admission before changing its dependency; add subcases to the existing facade contract test so its selector remains valid | `internal/modules/networkflow/network_flow_contract_test.go`; inspect `import_facade.go`, mapping/diagnostic resource helpers; no production edit required | C03; G01; Imports mapping preview/approval error translation | Raw diagnostic object: valid complete 15-member shape including legitimate nullable values; malformed/non-object/null/duplicate nested or top-level/trailing input; extra, missing and same-count wrong members. Public facade: valid prepared result roundtrip; schema/count/fingerprint/source-column failures. Preserve binding/error union and CSV preview assertions. | V01 facade/preview rows; V01 error row; V05 boundary baseline as appropriate to later task | Revert only newly added subcases; preserve baseline test names/fixtures and source behavior | Existing implementation passes meaningful accepted/rejected cases and exact error/result assertions; retained command/run evidence identified; no new production behavior or contract bytes |
| S02 | S01 passing evidence; RB-001 closed; later implementation authorization | Replace the import facade's forwarding HTTP decoder dependency with `platform/strictjson.DecodeObject`; add `import_facade.go` and `import_owner_errors.go` to existing semantic guard roots without weakening checks | `internal/modules/networkflow/import_facade.go`, `semantic_boundary_test.go`; inspect `semantic_reference_boundary_test.go` reachability and `query_authoring_test.go`; shared decoder and HTTP wrapper are read-only dependencies | C01/C02/C03/C13; identical strict syntax rejection and existing local validation/error semantics required | All S01 cases plus existing semantic import/reference guard, binding/error union, CSV preview and relevant import integration consumers | V01 selected rows; V02 service slice for Imports path if scope/risk guide requires; V04 generated drift and V05 backend boundary; reselect with task-guide in authorized implementation | Revert the dependency/root-list change together; no data rollback, contract version or migration expected | No HTTP import/reference from the admitted facade roots; same results/errors for characterization corpus; required selected checks pass; zero unexpected generated/contract/API changes |

S01 should test the validator's current shape responsibility, not strengthen it to validate every diagnostic member's value. The helper currently checks strict object syntax and the exact member set; changing that contract is outside S02. The three local diagnostic error forms remain “not an object”, “shape mismatch”, and “missing &lt;member&gt;” with their existing full prefixes.

A behavior discrepancy discovered during characterization must become a separate **requires later authorization** task; it must not be silently fixed in S01/S02. F03 has no ready implementation slice: destination, scan coverage and actual consumers must first be resolved by the Testing Harness owner.

## 8. Validation Plan

Discovery commands were executed successfully; product validation commands below were **not run**. “Required before implementation?” means an entry baseline before the corresponding production change, not authorization to execute checks during this documentation task. Re-run owner guidance when implementation is authorized because selectors and scopes may change.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit — V01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.query_authoring_admission` | Existing facade binding, closed error translation, preview and semantic/API guard selectors | yes, before S02 | Future baseline and post-change evidence. S01 subtests must remain reachable through these selectors. Command formed using discovered public ROWS selector and exact authored row IDs; not executed. |
| integration — V02 | `make test-slice OWNER=module.networkflow SCOPE=services` | Service-backed owner rows, including actual source mutation/import/lifecycle consumers | yes when entering affected behavior; baseline before production change | Discovered from task-guide. Registry-only support tests cannot substitute. Inspect Imports-owned row coverage before widening scope; do not infer from family names. |
| e2e/browser — V03 | `make test-slice OWNER=web.networkflow` | Frontend owner rows selected by its manifest | no for isolated decoder dependency change; yes if browser contracts expand | Discovered from task-guide. `make browser-e2e SUITE=visual` is an explicit-only candidate, not a default gate for S02. No visual run or UI claim here. |
| generated drift — V04 | `make source-check CHECK=generated` | Repository generated-artifact drift | no before S01; required after S02 in authorized implementation | Expected zero authored/generated contract changes. This command is not presumed read-only; use implementation workflow prerequisites. |
| import-boundary/static — V05 | `make lint CHECK=backend-boundary` | Backend package/import policy | yes baseline before S02; repeat after changes | Semantic file/reference guard additionally runs through V01. Generic boundary lint alone will not establish the preview contract. |
| frontend static — V06 | `make lint CHECK=frontend-boundary` | Frontend package boundary policy | no for S01/S02 | Only if later task expands into browser adapters/controllers. |
| full check — V07 | `make check`; `make ci` only when required by later scope/owner guide | Broader repository gates | no before narrow change | Discoverable broader targets, not promised execution. Later implementation follows repository `make agent-finalize` procedure before broader end-of-run verification. Neither is run for documentation. |
| documentation — V08 | `make lint-markdown` | Canonical documentation check | no for production; repository-required for this documentation change | **Skipped**: inspected path writes cache/retained artifacts and has readiness/dependency prerequisites; complete side-effect redirection was not established. Default configured globs also omit this tracker. TODO: discover an admitted Make-owned selection/redirection path before claiming file-specific lint coverage. |
| discovery — V09 | `make help ALL=1`; `make task-guide OWNER=module.networkflow`; `make task-guide OWNER=web.networkflow` | Public command/owner routing discovery | completed planning prerequisite | All three returned exit 0 in prior exploration; no product test/verification result follows from that. |

V08 skip is required by the sole-write boundary and the [refactor-tracker skill](/mnt/c/Users/jochi/.codex/skills/refactor-tracker/SKILL.md): “Do not run commands that may write outside the tracker”. The inspected path includes `tools/task_surface.generated.mk`, readiness/preflight and frontend prerequisites, `tools/harness/execution/run-step.sh`, `tools/harness/static-analysis/markdownlint.sh` and its runner. Cache and result directory overrides alone do not establish that remaining prerequisites are safe. No direct linter invocation, run-then-restore, generated edit or tooling-config edit was used as a workaround.

The existing Markdown config is `fix: false`, but that does not prevent wrapper/cache writes. Its handoff globs list selected files rather than all `docs/handoffs/*.md`; this tracker is absent from that list. A future generic successful invocation alone would therefore not establish that this file was linted. No command is invented to bypass that selection gap.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T01 | Admit canonical target, safe label/output and baseline | WF-00 | DONE | none | §1; baseline record E10 | Target exists, output absent/admitted, scope fixed |
| T02 | Account for 75 non-test and 54 test Go files | WF-01 | DONE | T01 | E02-001–129 / §2 | Exactly 129 unique rows and explicit inspection depth |
| T03 | Map major7/state4 owners and 19 HTTP operations | WF-02 | DONE | T02 | E01; C01–C17; route table | Owner and test posture for each discovered contract risk |
| T04 | Identify preview validator characterization gap | WF-03 | DONE | T03 | G01/F11; S01 | Direct missing-case matrix and located indirect tests distinguished |
| T05 | Diagnose existing facade, SQL, graph and harness boundaries | WF-04 | DONE | T03 | F01–F12; §3 | Supported keep/split/defer findings, no speculative redesign |
| T06 | Select narrow decoder seam and preserve existing owner architecture | WF-05 | DONE | T04, T05 | F01/F02; no-change recommendations | No public API or owner-state redesign included |
| T07 | Sequence characterization then seam cleanup | WF-06 | DONE | T06 | S01/S02 | Entry gates, rollback, commands and binary exits documented |
| T08 | Map selector reuse and harness security-scan constraint | WF-07 | DONE | T04, T07 | E07/E09; V01; RB-002 | Test accounting separated from runtime design |
| T09 | Add/run meaningful preview-result characterization | WF-03 | TODO | T07; later authorization | S01/G01 | S01 passes on pre-refactor source with retained evidence |
| T10 | Remove facade transport decoder dependency and guard it | WF-06 | BLOCKED | T09; later authorization | S02/RB-001 | Same C03 behavior, admitted semantic roots, selected checks pass |
| T11 | Decide harness-control relocation with security coverage | WF-07 | DEFERRED | separate scoped authorization; RB-002 | F03 | Destination and actual build/scan/consumer coverage justified before any move |
| T12 | Establish admitted file-specific Markdown lint path | WF-08 | DEFERRED | RB-003; documentation tooling scope | V08/F12 | Canonical check covers tracker without unauthorized writes |
| T13 | Reconcile document and final baseline preservation | WF-08 | DONE | T07, T08 | §12; E10; CMD-05/06 | Read-back reconciled tables/IDs; content/index comparison confirms only tracker changed |

## 10. Session Handoff Log

This is a newly created tracker. No prior on-disk handoff history existed. P00 below retains the useful prior read-only planning result; DOC-01 is the newly appended documentation session. Times use UTC where known; no exact timestamp is invented for the prior turn. “Touched” always means this tracker only. Named source files were read, not edited.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29, exact time not retained | Codex / P00 | Read-only plan; tracker absent | Framework first, E01 owner/support sources, AGENTS/skills; none touched | Git/inventory/source reads; V09 | Main baseline recorded; no writes or refactor | Full per-file inventory still pending then | DOC-01 completes file accounting |
| 2026-10-29 | Codex / DOC-01 | Canonical target/output and authority revalidated | E01; all E02; tracker created | Read-only identity/Git checks and source inspection | Major7 facade remains owner-required; no primary-owner contradiction established | None blocking this bounded plan | Revalidate baseline and governing clauses before later implementation |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Existing module/application architecture retained; narrow decoder seam selected | All E02; E03–E06; touched tracker only | `rg`, exact source reads, read-only inventory extraction | 129 file rows; F01/F02 supported; SQL/source contributions intentional | RB-001 gates S02 | Authorize S01 separately, characterize before changing dependency |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Consumer freeze mapped; no frontend refactor recommended | E08 client, protocol entrypoint, table owner, Collaboration interpreter and boundary tests; touched tracker only | Frontend reference searches and selected exact reads; P00 web task-guide | Shell/session/controller ownership remains outside backend target; no Core view/grid migration inferred | No frontend slice; deeper UI behavior audit would need scoped inspection | Preserve C08/C13/C17; use web owner guidance only if scope expands |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Route/schema/binding/native graph surfaces frozen | E09 contracts/policies; E02 current projection test; touched tracker only | Read JSON indexes, generated-root policy/topology and route handlers | 19 NF HTTP operations; no proposed API/type/schema/generated changes | No generated drift check run | Expect zero contract/generated diff in S01/S02; never hand-edit generated paths |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29, exact time not retained | Codex / P00 | Public verification routes discovered | Make/task surface, family manifests, test-ops guidance; none touched | V09 commands | Exit 0 discovery only; no product verification | No passing product evidence | Carry commands forward as candidates |
| 2026-10-29 | Codex / DOC-01 | Test assertions and support/security accounting inspected; G01 defined | 54 E02 test files; E07/E09; touched tracker only | Source/selector searches and exact assertion reads | Existing guard reusable; no direct preview-validation case found; harness move deferred | RB-001; RB-002; V08 skipped | S01 uses stable existing selector; preserve real-consumer evidence distinct from registry tests |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Current authority, replay, audit and guarded control boundaries mapped | E01-C04; operation admission/transactions/ports/security/keyring; E07 server/fixture/scan sources; touched tracker only | Exact source reads and import/reference searches | Production code uses injected ports; test fixture mutation lives in test support; ordinary server build excludes harness profile | Security scan preservation unresolved only for possible relocation | Keep C09/C14/C16 fixed; do not alter controls or authorization in S02 |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Tracker prepared for final editorial/preservation audit | Tracker plus baseline metadata; touched tracker only | Read-only document reconciliation and Git/content/index comparison pending final entry | Planning conclusions bounded; implementation not authorized | RB-001 and deferred RB-002/RB-003 | Complete E10, then hand off S01 as the first later authorized task |
| 2026-10-29 13:34:58 UTC | Codex / DOC-01 final audit | Tracker created and read back; planning complete for the declared scope | Tracker; repository content/index identities; touched tracker only | CMD-05/06 read-only audit | Twelve sections; 129 exact file rows; nine acyclic workflows with reciprocal edges; local paths resolve; baseline outside tracker unchanged | RB-001 gates S02; RB-002/RB-003 deferred; Markdown lint skipped | Revalidate baseline, then pursue S01 only under later implementation authorization |

### Inspected-file ledger outside the target

Grouped paths below are repository-relative; braces enumerate named files, not inferred wildcard coverage. The entire target's individual ledger is §2. Search-only discoveries do not imply body inspection.

| Evidence ID | Exact files or bounded surfaces inspected | Depth and relevance |
| --- | --- | --- |
| E03 | `internal/app/server/{runtime_assembly,extension_admission,network_flow_telemetry,collaboration_intents,server_profile_harness}.go`; `internal/app/configassembly/{configuration,deployment,extension_admission}.go`; `internal/app/extensionassembly/{cross_owner_transaction,incident_portability,recovery,network_flow_jobs}.go`; `internal/app/recoveryassembly/{graphprojection_restore,state_catalog}.go` | Relevant construction/contribution bodies and import references; harness profile, Collaboration translator and restore participant directly read. Not a full server audit. |
| E04 | `internal/modules/imports/{extension_facade,mapping,application_operations}.go`; `internal/platform/httpapi/json_object.go`; `internal/platform/strictjson/object.go` | Closed facade surface, real preview/approval validation calls, exact forwarding wrapper and shared strict decoder bodies. Indicators boundary traced through NF ports/participants and Core owner rather than auditing all Indicators internals. |
| E05 | NF graph/projection/store/reporting/restore files in §2; `internal/app/extensionassembly/network_flow_jobs.go`; `internal/app/recoveryassembly/graphprojection_restore.go` | Native projection invocation, Graph result-reader construction, lease read/release and restore borrowed transaction paths. Downstream Graph implementation not exhaustively audited. |
| E06 | NF Collaboration producer and store/commands; `internal/app/server/collaboration_intents.go` | Source-owned intent projection into Collaboration appender; no separate NF WebSocket server found. |
| E07 | `internal/testutil/networkflowsupport/{authorization,audit,intents,graph_source}.go`; `tools/test_support_inventory.json`; `tools/backend_module_boundaries.json`; `tools/harness/static-analysis/tests/test-go-gosec-targeted.mjs`; harness-profile server file in E03 | Fixture mutation/committed audit consumers, build admission and explicit runtime/support scan posture. Registry tests cannot substitute for real product assertions. |
| E08 | `apps/web/src/networkFlow/{networkFlowClient.ts,networkFlowCollaborationInterpreter.ts,useNetworkFlowTableOwner.ts,networkFlowBoundaryPolicy.node.test.ts}`; `packages/protocol-ts/src/entrypoints/network-flow.ts` | Exact selected bodies/imports and boundary assertions. Frontend test/selector files also located via family maps and targeted `rg --files`; those location-only findings are not coverage assertions. |
| E09 | `contracts/network-flow/{index.json,routes.v1.json,schemas.v3.json,import-facade-binding.v1.json,frontend-entrypoints.v7.json}`; `tools/{generated_artifact_policy.json,execution_topology_manifest.json,task_surface_owner.json,task_surface.generated.mk,task_surface.runtime.generated.mk}`; `tools/test_families/{module.networkflow,web.networkflow}.json`; `Makefile`; `.markdownlint-cli2.jsonc`; `tools/harness/static-analysis/{markdownlint,markdownlint-runner}.sh`; `tools/harness/execution/run-step.sh` | Contract versions/routes/registered inputs, protected generated roots and execution outputs, owner selectors, public targets and check side effects. Other contract registries referenced by the index were inspected at registry/dependency level; generated bodies are downstream projections, not semantic authority. |
| E10 | Git HEAD/index and tracked/nonignored worktree content/file identities; canonical output identity; completed tracker | Baseline and final preservation audit, described below. No product runtime reads or hashes Markdown as part of this work. This is documentation inspection only. |

Generated dependencies identified in source/policy include `internal/gen/contractnetworkflow`, `networkflowroutes`, `networkflowmapping`, `networkflowtz`, `networkflowunicode`, and `packages/protocol-ts/src/generated/network-flow-*`. Their generated posture was inspected; no generated file edit or regeneration is planned.

### Command and preservation ledger

| ID | Commands or inspection method actually used | Result and limits |
| --- | --- | --- |
| CMD-01 | `git --no-optional-locks status --short`, branch/HEAD/index inspection, `git --no-optional-locks ls-files`, diff inspection | Initial repository clean on stated main commit; optional index writes disabled. Target count and no extra target entries established. |
| CMD-02 | `rg --files`, `rg -n`/`rg -l`, `cat`, `sed -n`, bounded `head`/`tail`; read-only Python source/JSON/path inspection | Opened exact sources, extracted imports/declarations/test entrypoints and inspected selected bodies/assertions. Searches were extended along relevant dependencies and caller/contract/test paths only. Search hits were not used as edit sources. |
| CMD-03 | V09's three Make discovery commands | Each exited 0 during P00; establishes available commands/owner routing, not product validation. |
| CMD-04 | Read-only Python content/file-identity and index hashing; no comparison files written | Baseline covers 7,622 tracked files and zero nonignored untracked files, excluding admitted tracker; baseline fingerprints below. Target ignored/untracked scan empty. No mutating validation command run. |
| CMD-05 | Read-back/editorial table, ID, file-universe and workflow dependency audit | Completed: twelve ordered sections, 129 unique inventory rows exactly matching live Go files, consistent table widths, nine acyclic workflows with reciprocal predecessor/successor edges, valid work statuses, existing local file links and ledger paths, no trailing whitespace. This is not canonical Markdown lint or product verification. |
| CMD-06 | Final content/file-identity/index/HEAD comparison and Git diff/status | Completed: tracked and nonignored-untracked digests excluding tracker, index hash/inode and HEAD exactly match E10 baseline. Git reports only the new tracker; tracked diff is empty. Tracker is a regular file with one link. |
| CMD-07 | `make lint-markdown`, tests, formatting, generators, `agent-finalize`, source checks and full checks | **Not executed.** V08 is the skipped required documentation check for the recorded side-effect/selection reasons. Product checks are outside this documentation task. No test run root or result artifact exists for this session. |

E10 baseline: tracked content/identity digest `d0d83ba5294f1f12b5433bc52ba871364b0c541dc83eb4e9eb81e7ff6b2b5f06`; untracked digest `4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945`; index SHA-256 `7b50e63df03e37bfbb74005623a8673e0b2c81b7ffb40e10a88e6bf637e50f16`, inode `110884`. The read-only comparison includes path, mode, inode, link count, content size/hash for tracked/nonignored files and index bytes, excluding only this tracker. Relevant ignored command-output locations were considered through invocation-path inspection; no artifact-producing command was executed. This is not a claim to have hashed every unrelated ignored cache in the repository.

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | S02 lacks executed direct preview-result/diagnostic-shape characterization | Forwarding-wrapper equivalence does not prove complete caller behavior or test coverage | Later authorized S01; retained passing V01 baseline and preserved Imports contract behavior | BLOCKED — applies to S02 only; planning can finish with this explicit gate |
| RB-002 | Can harness controls move to shared test support while preserving ordinary-build isolation, runtime security scan inclusion and real consumers? | Current internal/testutil exclusion would change scan coverage under a naive relocation | Testing Harness owner decision plus updated scan/build/consumer evidence in a separate authorized scope | DEFERRED — no relocation proposed; does not block S01/S02 |
| RB-003 | What admitted Make-owned invocation both covers this tracker and avoids writes outside its allowed output? | Required docs check has cache/results/prerequisite writes; default globs omit this handoff | Tooling-supported selection and complete safe redirection, or a later task with broader write authorization | DEFERRED — required lint skipped and not claimed successful |

Unknown direct per-file test mappings in §2 are gates only if those additional files are selected for implementation; none expands S01/S02. No unresolved owner contradiction blocks the chosen architectural diagnosis. If later source review establishes one, record the exact clauses as **BLOCKED: owner contradiction** and block dependent decisions without choosing a side.

## 12. Binary Completion Criteria

Completion here is documentary, not implementation or conformance certification.

| Criterion | Status | Evidence / binary test |
| --- | --- | --- |
| Every target file inventoried or explicitly excluded | DONE | 129 unique E02 rows; 75 non-test + 54 test; no exclusions or unreadable entries; direct inspection depth declared |
| Every discovered public contract risk has owner and test posture | DONE | C01–C17, exact 19-route inventory; direct/indirect tests and future characterization distinguished |
| Every workflow has dependencies and exit criteria | DONE | WF-00–08 predecessors/successors/checkpoints; no intended cycles or dangling workflow references |
| Every implementation slice preserves behavior or is explicitly marked as later-authorized behavior correction | DONE | S01/S02 structural scope; any discovered correction requires later authorization and separate task |
| Validation commands discovered or TODO with reason | DONE | V01–V09; V08 selection/redirection gap explicit; no fabricated product result |
| Owner contradictions marked if found | DONE | None established in inspected owner scopes; required blocker wording retained |
| Repository/framework mismatches recorded | DONE | F03/F04/F10/F12: current facade, guarded support accounting, current versions versus historical names, docs validation coverage |
| Handoff sufficient for next agent without repeating broad discovery | DONE | Per-file ledger, owner/contract/finding links, S01 entry and S02 gate, external evidence/command ledger |
| Tracker read back and inventory/dependency/traceability audit completed | DONE | CMD-05: exact live file-set reconciliation, table/ID/path checks and acyclic reciprocal workflow graph |
| Content, file identity and index preserved outside tracker | DONE | CMD-06: E10 baseline digests/index/HEAD unchanged; only new tracker appears in Git status |
| Required documentation check result accurately reported | DONE | V08 skipped with exact limitation; no claim that lint passed or covered the file |

Final outcome: planning completeness is **complete** for this bounded target; S01 readiness is **ready** for a later authorized task after revalidation; S02 readiness is **blocked** by RB-001; implementation authorization is **not granted by this planning invocation**. Harness relocation remains deliberately deferred. Only this tracker was created; no production refactor was performed. V08 remains skipped, and no product validation success is claimed.

## 13. Remediation execution evidence

S04 repair evidence: `nf-s04-store-20261029-02` failed (Make exit 2, runner exit 3); `run-summary.json` and the retained Go stderr identify a test compilation error (`DB.Begin` instead of the owned `BeginTx` port). Corrected the test transaction setup; migration drift was not reached by that invocation.

S04 repair evidence: `nf-s04-store-20261029-03` failed (Make exit 2, runner exit 10) because the new two-table fixture reused an unrelated accepted-row primary key. Its retained `run-summary.json` and Go stderr identify `network_flow_rows_pkey`. Corrected the accepted-row fixture identity while keeping the diagnostic identity identical.

S04 completion (2026-10-29): F14/F16 implemented. Parsing now supplies structural facts; semantic validation stamps the owner NUL-delimited identity using source digest and mapping fingerprint. Diagnostics use complete message tokens, nullable row attribution, and numeric-only samples. Owner Table 21 now closes the previously unspecified mapped-interface failure with `network_flow_invalid_text`; request/resource schema identifiers remain unchanged. Migration 00030 changes only diagnostic uniqueness to `(network_flow_table_id, diagnostic_id)`. Core 01-661 and migration allocation now permit an owner-declared increment beyond the unchanged 29-file baseline. No historical migration was edited. Historical illustrative fixture diagnostics are not used as semantic acceptance evidence; the independent fixed vector and live producer tests are.

S04 validation: `nf-s04-diagnostics-20261029-02` focused preview/identity/privacy (exit 0); `nf-s04-store-20261029-04` storage lifecycle including repeated diagnostic identity across two tables and same-table SQLSTATE 23505 (exit 0); `nf-s04-migration-20261029-01` migration drift (exit 0); `nf-s04-schema-20261029-01` schema checks (exit 0); `nf-s04-generate-20261029-02` generation (exit 0); `nf-s04-markdown-20261029-01` Markdown lint (exit 0). Each run root is `.cartulary/test-results/<run-id>`; graph runs retain `run-summary.json`, tool runs retain their printed tool summary. Corrected test-setup failures are recorded above. F15 retention/counts remain explicitly pending S05; strict semantic admission remains pending S06/S07.

S05 validation repair: `nf-s05-focused-20261029-01` failed (Make exit 2, runner exit 10; retained `run-summary.json` and unit stderr). New retention cases and storage passed; the existing all-rejected error-translation test exposed its validator still requiring `int` after the explicit count became `int64`. Updated that owner-local validator; no wire type changes.

S05 completion: F15 fixed with an explicit row-validation result and separate preview/apply retention policies. Per-row sorting precedes retention; rejected rows and bounded all-rejected samples are independent of retained diagnostic multiplicity. Storage receives the actual rejected-row count. Tests cover limits 0/2/3/4 for three diagnostics, mixed structural/field ordering, preview/apply ID parity, multiple errors per row, all-rejected sample accounting, and 50/51 input records. `nf-s05-focused-20261029-02` focused preview/error tests passed; storage passed in `nf-s05-focused-20261029-01`; format and `nf-s05-markdown-20261029-01` passed. All successful owning exits are 0; earlier translation failure and repair are recorded above. Next S06 tightens the admission boundary without weakening these producer guarantees.

S06 validation repair: `nf-s06-focused-20261029-01` failed compilation (Make exit 2, runner exit 3; `run-summary.json` and unit stderr) after moving validation left two unused imports in the facade. Removed those imports before rerunning.

S06 validation repair: `nf-s06-focused-20261029-02` (Make exit 2, runner exit 10) passed the result corruption corpus but the dependency guard rejected the newly introduced reflection import. Replaced reflection with existing canonical JSON comparison and explicitly admitted standard-library `slices`; transport prohibitions and transitive traversal remain unchanged.

S06 completion: F13 fixed. The private strict decoder validates closed outer/nested objects, required presence and nullability, descriptors and sample digests/counts, mapping variants/default materialization, canonical outer/inner mapping agreement, fingerprints, bounded complete diagnostic counts, attribution, identity, messages, reasons and owner order. It shares the S04 diagnostic validator with retained-state admission planned in S07. The semantic guard includes both new private validation files. `nf-s06-focused-20261029-03` preview corruption corpus and transitive boundary passed; `nf-s06-imports-20261029-01` safe owner translation unit/integration consumers passed; `nf-s06-boundary-20261029-01` backend boundary and `nf-s06-markdown-20261029-01` Markdown lint passed (all exits 0). Both Imports publication and approval continue invoking the owner admission method before returning/storing owner output; no Network Flow semantics were added to generic Imports. Prior failures and repairs remain recorded. Next S07 publishes this coordinated breaking correction and validates retained diagnostics.

S07 generation repair: `nf-s07-generate-20261029-01` failed (Make exit 2, generator exit 1; `generate/tool-run-summary.json`, `generate/generate-contracts/stderr.log`) on the authored initialization digest. The Extensions canonical digest includes a terminal newline; corrected the authored digest and retained the old 3→4 definition as an inert historical ledger record, with no executable migration.

S07 validation scheduling repair: `nf-s07-format-20261029-01` exited 2 at `runtime_plan/configuration_error` before a run manifest or formatting execution; only `harness-runtime-build/tool-run-summary.json` was published (its build passed). Authored selectors had changed since generation. Regenerate the routing projections before invoking graph-backed checks; generation/schema checks chained after format were not reached.

S07 routing repair: `nf-s07-generate-20261029-03` failed (Make exit 2, generator exit 1; `generate/tool-run-summary.json`, `generate/generate-task-surface-and-execution-topology/stderr.log`) because renamed selector names were not ASCII-sorted. Sorted the authored selector array. This also explains the preceding format planning failure; format/schema had not run. The browser adapter now derives its sole supported major from the generated descriptor, eliminating its duplicate hard-coded major.

S07 routing repair: `nf-s07-generate-20261029-04` failed (Make exit 2, generator exit 1; `generate/tool-run-summary.json`, topology step stderr) because the browser test title changed to major 8 before its authored selector. Updated the selector and active graph projection test name together; historical row IDs remain stable.

S07 validation repairs: `nf-s07-schema-20261029-01` (Make exit 2, runner exit 10) identified unregistered retained historical schemas; registered both historical and current schemas without adding runtime major aliases. `nf-s07-state-nf-20261029-01` (Make exit 2, runner exit 10) passed preview/projection/binding cases but the new restore corruption fixture hit the correct immutable-update trigger; changed fixture setup to transactional delete/reinsert representing restore, followed by rollback. `nf-s07-state-ext-20261029-01` (Make exit 2, runner exit 10) passed state tests but an adoption-matrix assertion still expected major 7; corrected that projection expectation. Each failure is retained under its named `.cartulary/test-results` root with `run-summary.json` and unit logs. `nf-s07-client-20261029-01` passed current/unsupported client-major and transport tests (exit 0).

S07 schema repair: `nf-s07-schema-20261029-02` failed (Make exit 2, runner exit 10; `run-summary.json`, JSON-shape unit stderr) on the current contract index schema's stale major-7 constant. Updated that authored constant. Publication review also found approved display-name overrides were not normalized before strict result admission; materialization now uses the existing owner normalizer, with a valid-input roundtrip regression. The corruption corpus was moved outside the retention-limit loop to avoid redundant execution.

Operational cutover procedure (deployment remains pending): retain the previous environment, database, object storage, keys and matching server/browser release as a coherent rollback unit. Obtain original CSV sources and mapping decisions before scheduling re-import; retained tables and diagnostics cannot reconstruct source files. Provision a separate normal empty state-5 environment, configure the intended claims and keys, apply the authored database migrations through `make db-migrate` against that new environment, and let normal admission initialize metadata. Publish the server and browser assets from the same major-8 build, verify discovery/availability and readiness, then re-import originals through Imports and verify counts/diagnostics. Admission of old metadata or old authoritative state is expected to fail without conversion or deletion. Roll back by returning to the untouched previous environment; never downgrade a database after state-5 writes, remove metadata, or delete data to force admission. No deployment or database reset is performed by this task.

S07 schema repair: `nf-s07-schema-20261029-03` failed (Make exit 2, runner exit 10; `run-summary.json`, JSON-shape unit stderr) on hard-coded major-7 semantic assertions in the existing shape checker. Updated all four active Network Flow checks to major 8. Retained-state review found Recovery's declarative post-restore algorithm was not executed: the catalog now requires a bound owner callback, invoked once per profile/algorithm inside the authoritative restore transaction. Network Flow checks state-5 metadata and diagnostic semantics before commit; empty owner state may still initialize normally. Removed the obsolete state-1 pristine restore allowance. These are F17 implementation obligations, not a general Recovery redesign.

S07 schema repair: `nf-s07-schema-20261029-04` failed (Make exit 2, runner exit 10; JSON-shape unit stderr) because the semantic error inventory omitted the newly specified optional-interface diagnostic. Added that row to the existing exact inventory. Owner-guide discovery for `app.extensionassembly` returned unknown owner (exit 2); its verification routes belong to the already selected `app.server` owner. `module.recovery` guidance passed.

S07 validation repairs: `nf-s07-schema-20261029-05` failed (Make exit 2, runner exit 10) because the edited shape checker required topology regeneration; rerun generation before schema verification. `nf-s07-restore-20261029-01` failed compilation (Make exit 2, runner exit 3) because a same-package Recovery test imported Network Flow through Imports back to Recovery. The atomic-restore test now exercises the callback boundary with a transaction-reading validator; Network Flow's real validator is exercised in its owner integration test. `nf-s07-server-20261029-01` publication/discovery and `nf-s07-state-nf-20261029-03` real owner state/restore validation passed. Final producer review corrected an overly strict empty-policy admission check: a null policy for a non-nullable current-profile scalar must reject that row, not panic or reject an otherwise well-formed mapping. Added the owner-required roundtrip regression.

S07 completion: F08/F10/F17 now use major 8, document 8.0.0, state 5/minimum 5, frontend entrypoint v8 and operation policy v2. Resource/request schemas, Graph semantic query v2, Jobs payload v1 and diagnostic algorithm v1 remain unchanged. Historical contract files and migration ledger facts remain inert history; active server/browser publication admits only major 8 and no legacy state migration. Retained diagnostics are validated through the shared owner boundary at startup and through a bound callback before Recovery's authoritative transaction commits. No prior data was converted or deleted.

S07 child evidence: `nf-s07-state-ext-20261029-01` state-1/2/3/4 rejection, initialization and state-runtime cases passed, but the overall invocation failed on its separate stale registry assertion. That assertion passed in `nf-s07-registry-20261029-02`. S07 successful invocations (owning exits 0): `nf-s07-repaired-20261029-01` validates real current/restored diagnostics, mapping roundtrips and the actual Recovery catalog binding. `nf-s07-restore-20261029-02` validates callback execution/rollback in the authoritative restore transaction and the catalog contract. `nf-s07-server-20261029-01` covers publication/discovery; `nf-s07-client-20261029-01` covers supported and unsupported majors. `nf-s07-generate-20261029-09`, `nf-s07-schema-20261029-06`, `nf-s07-generated-20261029-01`, `nf-s07-boundary-20261029-01` and `nf-s07-markdown-20261029-01` passed. Earlier failures remain explicit diagnostic history, not final acceptance evidence.

S08 entry (initial plan reading; migration coverage corrected below): focused remediation gates passed. Reviewed the canonical CI plan (`make target-plan TARGET=ci DETAIL=summary` and `DETAIL=graph`; exits 0; read-only plan saved at `.cartulary/nf-ci-target-plan.txt`). CI selects 424 work units including backend/frontend verification, build/ordinary-versus-harness policy, runtime security scans, schemas, generated policy and owner boundaries. Its graph has no standalone `migration-drift` operation; S04 migration drift remains slice evidence and the final portfolio will run that missing obligation after source freeze. No extra visual or benchmark qualification is selected. Run `make agent-finalize`, review mutations, then freeze product source and run `make ci`; S08 remains incomplete until final acceptance and handoff reconciliation pass.

S08 source review: `nf-s08-finalize-20261029-01` passed (exit 0). Review found only intended generated changes. Added a second independent diagnostic identity vector for null column/field attribution and cleaned import grouping and the migration allocation variable name before the final freeze; these review changes require finalization again.

S08 review correction: the earlier S00 wording incorrectly treated implementation authorization as completed human owner adoption. New amendments are now explicitly proposed in Network Flow, Extensions and affected Core declarations. Unchanged owner requirements retain their adopted baseline authority. Human review of the coordinated amendment and projection fidelity is an explicit release handoff gate, not a test result or a deployment action performed here.

S08 first final portfolio: `CARTULARY_TEST_RUN_ID=nf-s08-finalize-20261029-02 make agent-finalize` exited 0. Frozen product source was `sha256:1a3244d9a3946782afa70bbd1a40b92b3021bf6ce8616695865af61925ad0bee`. `CARTULARY_TEST_RUN_ID=nf-s08-migration-20261029-01 make migration-drift` exited 0 (8/8 units, 19 cleanup steps passed). `CARTULARY_TEST_RUN_ID=nf-s08-ci-20261029-01 make ci` exited 2 (owning graph exit 10, `product/test_assertion_failure`; 420/424 units passed, four failed, none skipped/cancelled; 726 cleanup steps passed). Evidence is in the named `.cartulary/test-results/<run-id>/run-summary.json`, `target-summaries/` and `cleanup-results.json`. CI is not accepted.

S08 failure diagnosis before repair: CI rows `module.networkflow.unit.query_authoring_admission` and `module.extensions.integration.admission_preflight_v6` each exited 10. Retained unit logs `go-0a1eb56aa3639a55-219fa29a18aa/stderr.log` and `go-8121b7778a618816-c071127b976b/stderr.log` identify the missing reviewed export entry for the new real restore validator and an obsolete expectation of one executable migration. The `harness.lint_go_format` projection identifies three unformatted test files; `harness.lint_go_staticcheck` identifies the pre-existing unused `requireWritableStringNormalization` helper in Incidents request tests (confirmed in baseline HEAD). All four failures require repair; the unused helper can be removed without changing product behavior. Failed observer usage attempts (`DETAIL=failures`, logs without a target, and an internal target name) are diagnosis-command errors, not additional product failures; supported summary/children and retained structured results established the findings.

S08 portfolio correction: the earlier plan inspection incorrectly concluded CI omitted migration drift. The actual retained CI projection includes passing `migration-drift` and its eight units. The separate passing drift run is redundant historical evidence; do not repeat it in the repaired final portfolio. CI also passed schema/generated checks, browser suites, security scans, builds, store tests and Recovery tests. These are earlier-source results and do not close final acceptance after repair.

S08 source-review obligations: admit Table 10-A2's zero-record shape (while retaining the parser's separate header-only error); replace active legacy diagnostic test placeholders with valid owner diagnostics and exercise the same identity across incidents; run Make formatting. Documentation review also reconciled remaining active major-7/state-4 adoption and acceptance declarations. Re-run affected focused rows, finalize again, freeze source and execute CI before handoff completion.

S08 repair scheduling: `nf-s08-format-20261029-01` passed (exit 0). The attempted `test-slice ROWS=...` under `nf-s08-repair-20261029-01` exited 2 at public input admission because exactly one of OWNER/SUITE is required; no product row ran. Split the mixed-owner repair selection into exact Network Flow and Extensions owner invocations. The earlier `agent-finalize` successes did not imply Go formatting had run.

S08 user-directed portfolio revision: after the first CI run, the user explicitly instructed, “Don't re-rrun a full CI, validate with focused tests.” This supersedes the planned repaired full-CI invocation. Final acceptance will use focused affected behavior, static checks, generated/schema checks and documentation lint on finalized source. The failed earlier CI remains broad diagnostic evidence, not a final-source pass or complete-owner audit. No second full CI is authorized or required for this handoff.

S08 focused repair diagnosis: `nf-s08-repair-nf-20261029-01` exited 2 (runner 3, `infra/preflight_error`; 18 cleanup steps passed) because new fixture code referenced private owner symbols from the external `networkflow_test` package. Its unit stderr identifies the inaccessible diagnostic helpers and request type. `nf-s08-lint-go-20261029-01` exited 2 (lint units exit 1) for the same compile defect in vet/staticcheck; Go formatting passed. Use an independently constructed valid test diagnostic, the existing exported retained-state validator, and a test-build request alias rather than adding production exports. `nf-s08-repair-ext-20261029-01` and `nf-s08-markdown-20261029-01` passed with owning exit 0. Each exact root is `.cartulary/test-results/<run-id>`.

## 14. Completed remediation finding dispositions and phase gates

This section supersedes historical findings' scheduling and completion states without erasing their evidence. The user-revised focused final portfolio and documentation checkpoint passed; S08 is complete. The scope remains bounded remediation, not a complete Network Flow conformance audit.

| Finding | Current disposition and areas | Owner / rationale and durable benefit | Compatibility and risk if unresolved | Required validation / evidence |
| --- | --- | --- | --- | --- |
| F01 | Fixed in S02: use `strictjson.DecodeObject` directly; implementation | NF §29, NF-REQ-210; semantic code depends on the existing shared primitive, preserving one-way ownership | No independent wire/storage change; transport coupling otherwise spreads into reusable semantics | Dedicated preview characterization and transitive semantic-boundary tests; S02 evidence, S08 portfolio |
| F02 | Fixed in S02/S06: guard facade, error adapter and both new private validation files; tests | NF §29; guard actual public contribution entrypoints and local reachability rather than an incomplete root sample | No compatibility change; incomplete roots otherwise conceal architectural regression | Existing direct/alias/local-reference guard and boundary lint; S02/S06 evidence, S08 portfolio |
| F03 / RB-002 | Resolved, no relocation: keep harness controls owner-local; architecture/documentation | Testing Harness §12.2.9 and E07/E09 build/scan evidence; guarded runtime controls need owner context and runtime security scanning | Avoids relocation coupling and accidental scan exclusion; unreviewed movement could weaken security evidence | Ordinary-build exclusion, runtime scan inclusion and real rollback/audit/auth consumers passed in CI-01; registry presence alone is insufficient |
| F04 | Resolved, retain module/application boundary; architecture/documentation | NF-REQ-001/002/210/211; the source owner already has a coherent facade and retained application bundle | No compatibility change; speculative splitting would fragment ownership and lifecycle | Construction/source tests passed in CI-01; the repaired public-surface guard passed in the final Network Flow selection |
| F05 | Resolved, retain private SQL/transaction adapters; architecture/documentation | NF persistence and caller transaction ports, backend procedure; cohesive storage and lock ownership | No compatibility change; moving owner SQL into generic helpers adds coupling and obscures atomicity | Store lifecycle, transaction rollback and boundary checks in CI-01 |
| F06 | Resolved, retain distinct Graph/Reporting/Recovery owners; narrow F17 restore callback added in S07 | NF §§14/29, Graph V2, Reporting exact lease contracts, Recovery authoritative restore; owners exchange narrow capabilities | Unchanged Graph semantic query v2; unchecked restore otherwise publishes invalid owner state | Projection/lease contracts plus actual owner callback and restore commit/rollback tests; S07 evidence, S08 portfolio |
| F07 | Resolved, retain explicit mutation ports and application translators | NF §29.1 and Core transaction/audit/intent ownership; one commit boundary and owner-specific effects | No ownership transfer; implicit cross-owner effects would threaten rollback and authorization | Replay, rollback, audit and real consumer tests in CI-01 |
| F08 | Superseded: coordinated authored and generated updates in S01/S04/S07 | Core 00 precedence and generated-artifact policy; projections follow reviewed owner decisions | Zero-generated-diff restriction no longer applies; hand edits or partial publication would drift | Make generation, schema/generated checks and finalized CI; no historical migration edits |
| F09 | Resolved, retain Core saved views/records versus NF saved graphs/immutable rows | NF scope and Core view-schema/grid ownership; no duplicate frontend subsystem introduced | Current client major changes only; conflating resources would burden later extension | Browser adapter/publication and native Graph contracts in CI-01; no visual redesign obligation |
| F10 | Resolved in S07: update active contracts/selectors, retain historical filenames and stable row IDs | NF-REQ-212 / EXT-REQ-170; version truth resides in active declarations, not cosmetic filenames | Active major 8/state 5 only; inert v7/v1 contract files and old migration ledger facts remain history | Discovery/client/registry/schema/generated checks; historical names are not active aliases |
| F11 / G01 / RB-001 | Fixed in S01: dedicated preview admission selector; tests/routing | NF Table 10-A2 and Core Imports publication/approval boundary; meaningful public behavior independent of binding metadata assertions | Test-only baseline change; preserving malformed acceptance would obstruct corrections | S01 valid/framing/safe-translation baseline; S04–S07 correction corpus; S08 portfolio |
| F12 / RB-003 | Fixed in S03: module-tracker glob; tooling/docs | Canonical Markdown maintenance; future module trackers receive ordinary lint coverage | No product change; omitted controlling artifacts otherwise escape lint | `make lint-markdown`, selecting both module trackers; S03 and final handoff check |
| F13 | Fixed in S06: private typed strict result admission; implementation/tests/spec clarification | NF Table 10-A2, Tables 10-A/10-B and §12.4; one closed boundary verifies descriptors, canonical mappings, fingerprints and diagnostics | Malformed internal output now rejects through `owner_preview_validation_failed`; unchecked output could disclose unsafe data or approve inconsistent mappings | Required/unknown/duplicate/null/type and semantic corruption corpus; valid producer roundtrips; both Imports consumers; S06/S07 evidence and S08 portfolio |
| F14 | Fixed in S04: source/mapping-derived identities and shared diagnostic construction/validation; implementation/tests/contracts/spec | NF §12.4, Tables 12-E/12-F and 21-A/A1; stable provenance, closed message/reason relationships and safe scalar treatment | Erroneous old values change under fresh-state cutover; unresolved identities collide or misattribute sensitive data | Independent scalar and null-attribution vectors, preview/apply parity, sensitive/row-level samples and retained-state corruption tests; S08 portfolio |
| F15 | Fixed in S05: explicit validation result and separate preview/apply policies; implementation/tests/spec | NF Table 10-A2, §12.4, NF-AC-063; complete bounded preview, ordered retention and independent rejected-record accounting | Corrected low/exact-limit behavior; conflation otherwise disables preview or misreports rejection/truncation | Limits 0/N/N+1, mixed structural/field failures, multiple diagnostics per row, all-rejected sample and 50/51-record cases; S08 portfolio |
| F16 | Fixed in S04: migration 00030 adds table-scoped diagnostic key; SQL/implementation/tests/spec | NF §12.4/NF-REQ-212 and Core REQ-01-661 increment; physical uniqueness matches table-independent semantic identity | New physical schema; no ID algorithm fork or old data rewrite; otherwise identical valid source imports collide across tables | Cross-table duplicate identity accepted, same-table duplicate rejected, incident isolation/pagination/atomic rollback and migration drift; S08 portfolio plus drift |
| F17 | Fixed in S07: current-only state-5 admission and transactional restore validation; spec/contracts/code/tests/docs | NF-REQ-212, §29.2 restore obligation, Extensions state ownership and EXT-REQ-170; corrected runtime never silently serves incompatible retained facts | Breaking major-8/state-5 cutover; old state remains untouched and blocks admission; originals required for re-import | Fresh initialization, unchanged old-state rejection, malformed state-5 rejection, valid restore and atomic invalid-restore rollback; S07 evidence and S08 portfolio |

| Phase / workstream | Dependency | Principal risk and control | Exit / handoff |
| --- | --- | --- | --- |
| S00 | Revalidated baseline | Mistaking defects for requirements or implying human adoption; consistent proposed amendments and explicit review gate | Owner decisions and all finding dispositions recorded; proposed status retained for adoption review |
| S03 | S00 | Unbounded historical lint expansion; constrain selection to module-tracker glob | Canonical lint covers both trackers and passes; RB-003 closed |
| S01 | S03 | Tests copying implementation assumptions; use public facade invariants and independent valid/framing cases | Passing pre-change characterization retained; G01/RB-001 closed |
| S02 | S01 | Decoder movement changes admission/error translation; retain characterization and transitive guard | Shared primitive, safe public failure translation and boundary checks pass |
| S04 | S02 | Identity/privacy drift or global collisions; one typed validator, independent byte vectors, additive migration | Identity/privacy and storage uniqueness checks pass |
| S05 | S04 | Parser/semantic ordering loses early errors or preview grows unbounded; merge in source-row order and retain only under explicit mode | Exact retention/count/order and bounded-preview cases pass |
| S06 | S05 | Strict boundary rejects supported producer output; paired producer roundtrips and independently corrupted results | Closed result/mapping/diagnostic corpus and both Imports consumers pass |
| S07 | S06 | Partial major publication or invalid restored facts; generated current-only registry and owner validator inside restore transaction | State/client/publication/schema/generated checks pass; operational cutover documented |
| S08 | S07 | Treating historical, cached or partial evidence as complete; inspect exact final roots, source identity, owning exits and cleanup | All required portfolio checks pass and tracker is reconciled; human adoption/deployment remain explicit handoff gates |

No phase may skip its exit and begin a dependent phase. S08 repairs reopen only the affected obligations and require renewed finalization/source qualification before completion. Historical evidence remains useful for diagnosis but cannot replace final compatible acceptance.

## 15. Completed remediation validation and operational handoff

### Source, environment and acceptance scope

Baseline and current HEAD are `a182673ae3b460de9c0d078d3315f17e18c4f36a` on `main`; implementation changes remain in the working tree. The pre-existing untracked tracker and its original inventory, IDs and historical evidence were preserved. Final executable-source identity is `sha256:3c532611ae42c1f142d42688f6d60984bbb3704e795109904004923d84fc4122`, recorded identically by all final focused manifests. Documentation is outside executable-source identity and was edited only for the human handoff after the product freeze.

Execution environment: Linux `6.6.114.1-microsoft-standard-WSL2`, x86_64, GNU/Linux; repository `/home/jochi/code/cartulary`; execution date 2026-10-29. The pin registry specifies Go 1.27.3, Node 24.15.0 and pnpm 10.33.0; CI-01's toolchain check passed. Final manifests record toolchain digest `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f` and system digest `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

The user changed final acceptance to focused validation after CI-01. No second CI ran. CI-01 remains a failed earlier-source invocation: its four failures are individually repaired and requalified by the final focused checks. Passing earlier-source publication, browser, build, scan, Imports, Recovery and migration projections are identified as such; they are not a full-source current acceptance audit. No full Network Flow conformance, benchmark qualification, deployment adoption or complete-owner audit is claimed.

### Final commands and evidence

Every command below ran from the repository root. All completed product commands have owning exit 0. Each run root is `.cartulary/test-results/<run-id>`; graph results use `run-summary.json`, `run-manifest.json`, `target-summaries/` and `cleanup-results.json`.

| Run ID | Exact Make selection | Result / purpose |
| --- | --- | --- |
| `nf-s08-format-20261029-02` | `make format` | PASS; includes authored untracked Go files |
| `nf-s08-finalize-20261029-03` | `make agent-finalize` | PASS; generated refresh unchanged, zero rewritten files; `unit-artifacts/finalize-summary.json` |
| `nf-s08-final-nf-20261029-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.query_authoring_admission,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.pagination_recovery,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.extension_state_v4_v1_rejection` | PASS, 6/6 work units; preview/diagnostics/retention/strict admission, dependency and export guards, storage identity and incident isolation, pagination/replay/privacy, current-state and real restore binding |
| `nf-s08-final-ext-20261029-01` | `make test-slice OWNER=module.extensions ROWS=module.extensions.integration.admission_preflight_v6,module.extensions.integration.network_flow_state_v4` | PASS, 3/3 work units; no executable prior-state migration, fresh initialization and unchanged old-state rejection |
| `nf-s08-lint-go-20261029-02` | `make lint CHECK=go` | PASS, 6/6 work units; formatting, vet and staticcheck, including the removed unused baseline helper |
| `nf-s08-schema-20261029-01` | `make source-check CHECK=schema` | PASS, 3/3 work units; authored schema/contract consistency |
| `nf-s08-generated-20261029-01` | `make source-check CHECK=generated` | PASS, 5/5 work units; generated consistency after finalization |
| `nf-s08-handoff-markdown-20261029-01` | `make lint-markdown` | PASS, owning exit 0; summary under `adhoc/lint-markdown/tool-run-summary.json` |

Each invocation set `CARTULARY_TEST_RUN_ID` to the exact table value. These work-unit counts are not test counts. Final Network Flow and Extensions invocations used no execution-result cache hits. Their cleanup checks passed (21 and 14 steps respectively); Go lint/schema/generated cleanup passed (15/12/14 steps), with none failed or blocked. No selected final product work was skipped or cancelled. Earlier slice evidence and exact failure roots remain in §13.

`agent-finalize` did not select retained-run maintenance because no compatible successful full warm-check root was supplied. Its optional retained-evidence and scheduler checks therefore report `results-dir-not-provided`; no redundant warm check was run to obtain one. Full CI was not repeated at the user's explicit direction. No extra visual or benchmark qualification was selected. No required check under the revised focused portfolio is waived. The final documentation checkpoint passed.

### Delivered changes and compatibility boundary

The semantic import entrypoints now use the shared strict decoder under the transitive owner guard. New private preview and diagnostic validators enforce required presence, closed nested objects, canonical mapping equality, deterministic identities, safe messages/samples and complete preview accounting. Preview diagnostics are bounded by 50 records; apply retention uses the first configured N in owner order, with actual omission and independent rejected-record counts. Active integration fixtures now use independently constructed owner-valid diagnostic bytes, including the same identity in two tables of one incident and a third table in another incident; duplicate identities inside a table still fail.

Authored migration `db/migrations/00030_network_flow_diagnostic_identity.sql` changes diagnostic uniqueness to `(network_flow_table_id, diagnostic_id)`. Versions 1–29 remain unchanged. The allocation/catalog/schema projections were regenerated through Make. This is a physical key change, not a diagnostic-ID format fork or historical data conversion. Migration drift passed both its recorded standalone invocation and CI-01; the SQL and allocation inputs were unchanged by final repairs.

Publication is coordinated at Network Flow document/public major 8, durable state 5 and minimum admitted/migratable state 5. Current frontend entrypoint is v8 and operation policy v2. Existing request/resource schema IDs, diagnostic algorithm v1, Graph semantic query v2 and Jobs payload v1 remain. Historical filenames and ledger facts remain inert history. The active prior-state migration route and dual-major compatibility are absent. Startup and authoritative restore validate diagnostic semantics; restore rejection rolls back before publication.

Harness controls remain in their owner-local package, preserving ordinary-build isolation and runtime scanning. Graph, Reporting, Recovery, Indicators and Collaboration retain their respective responsibilities. The sole unrelated source cleanup removes a pre-existing unused Incidents test helper that blocked required Go static analysis; no Incidents behavior changed. `docs/domain.md` required no edit because vocabulary and owner navigation are unchanged.

### Adoption and rollout instructions

Human owners must review the proposed Network Flow 8.0.0 / Extensions 0.13.0 / coordinated Core amendments, including the diagnostic reason registry and migration increment, and explicitly adopt them before release. Reviewers must also verify their fidelity to authored projections. Tests cannot establish those judgments. Deployment adoption is separate from implementation completion.

Older Network Flow data was **neither converted nor deleted**. Incompatible retained metadata or authoritative state blocks admission without rewriting it. Original CSV source files and mapping decisions are prerequisites for re-import; retained tables and diagnostics do not reconstruct those sources.

Retain the previous environment's database, object store, keys and matching server/browser assets as one rollback unit. Provision a separate empty environment, apply the authored migrations through `make db-migrate` there, configure claims/keys, and let normal admission initialize state 5. Publish the matching major-8 server and browser assets together; verify discovery, supported client behavior and readiness before re-importing original sources. Verify imported row counts, retained diagnostics and cross-incident isolation before operational acceptance.

Rollback returns traffic to the untouched previous environment and matching release. Do not downgrade a database after state-5 writes, remove metadata, convert old rows, or delete data to force admission. This effort performs no deployment or destructive reset.

### Completion checklist

| Criterion | Status / evidence |
| --- | --- |
| S00/S03/S01/S02/S04/S05/S06/S07 completed in dependency order with tracker checkpoints | PASS; §1 and §13 |
| F01–F17 and G01/RB-001/RB-002/RB-003 explicitly fixed, superseded or resolved without unnecessary relocation | PASS; §14 |
| Required focused final behavior and static/generated/schema checks pass on finalized source | PASS; exact roots and source identity above |
| Earlier CI failures diagnosed, repaired and focused checks passed; no false full-CI claim | PASS; §13 and final portfolio |
| Historical inventory/evidence retained; compatibility, original-source requirement and rollback documented | PASS; §§2–13 and this handoff |
| Human owner adoption and deployment separated from implementation | PASS; amendments proposed, human adoption/release pending |
| Final controlling-artifact Markdown checkpoint | PASS; `nf-s08-handoff-markdown-20261029-01`, owning exit 0 |

S08 completion (2026-10-29): implementation, validation and handoff are complete under the user-revised focused scope. G01/RB-001 is closed, RB-002 is resolved by retaining owner-local controls, and RB-003 is closed by canonical glob coverage. No remediation blocker remains. Human owner adoption and operational deployment remain explicit future gates; older Network Flow data was neither converted nor deleted. The completion-status edit is followed by one final documentation-only lint invocation, `CARTULARY_TEST_RUN_ID=nf-s08-completion-markdown-20261029-01 make lint-markdown`; no product checks or finalization are repeated for that Markdown-only checkpoint.

## 16. DOC-02 baseline, inventory and owner decisions

### Current source and permitted changes

Artifact mode: **normal**, for the bounded cleanup and hardening plan requested by the user. Repository root is `/home/jochi/code/cartulary`; canonical target remains `internal/modules/networkflow`, with related Imports source-reader, contract publication and verification dependencies explicitly included. The sole permitted tracked output is `docs/handoffs/networkflow-module-refactor-tracker.md`. The planning framework remains `docs/handoffs/cartulary_modular_refactor_planning_framework.md`.

Revalidation found clean `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with no staged, unstaged or non-ignored untracked changes. This is the same baseline as the immediately preceding read-only planning turn. S00–S08 changes are now committed; §15's working-tree description is the historical S08 posture, not the DOC-02 posture. Prior test results remain historical results, not checks of this iteration's future implementation.

The tracker and its parent directories are not symlinks. The tracker is a regular file with one hard link; initial identity was device `2096`, inode `110366`. Its pre-edit SHA-256 was `e72ca49d3645047236d4a209af5c61a0492228c91bcf34f443cd261b16e565f9`; the Git index SHA-256 was `2e73fc3ff1e2a15ffe8c9b3e37443b524125bef093ea439e7827f9aec4e871d8`. These are manual document-preservation observations, never inputs to product checks, generators, conformance or release evidence.

The user authorized canonical Markdown lint and its ordinary ignored cache/result artifacts. That explicit instruction supersedes DOC-01's narrower command-write restriction and its historical V08 skip. No other managed verification, generator, formatter, dependency change or service operation is part of this documentation update.

### Inventory delta and inspection coverage

The current universe is all tracked and non-ignored untracked files under the literal target: **132 Go files, 77 non-test and 55 test files**. Every one of the original 129 paths in §2 still exists. The following three paths complete the current inventory; no original file was removed. No target path is generated. Generated projections outside the target remain explicitly accounted dependencies.

| Evidence ID / added path | Current responsibility and owner | Current inspection disposition | Finding / validation |
| --- | --- | --- | --- |
| E02-130 / `internal/modules/networkflow/diagnostic_validation.go` | Private Network Flow diagnostic identity, message, attribution, reason and limit validation | Directly inspected relevant validator bodies; retained as the shared semantic boundary | F23; C21; V13/V14 |
| E02-131 / `internal/modules/networkflow/import_preview_test.go` | Network Flow preview construction/admission and diagnostic regression assertions | Directly inspected relevant assertions, including the forwarding-helper caller; no test executed | F19/F23; C18/C21; V10/V14 |
| E02-132 / `internal/modules/networkflow/import_preview_validation.go` | Private typed preview-result decoder and admission | Directly inspected decoder, wrapper and nested admission bodies | F19/F23; C18/C21; V10/V14 |

Current direct-inspection set, including the preceding planning turn on the identical baseline: `application_composition.go`, `csv_parser.go`, `diagnostic_validation.go`, `import_facade.go`, `import_preview_test.go`, `import_preview_validation.go`, `indicator_link.go`, `mapping.go`, `network_flow_contract_test.go`, `store.go`, and `transaction_participants.go`. These 11 paths received relevant-body or selected-assertion inspection, not an assertion that every line was audited. Every other target path in the §2 inventory is **metadata-only for DOC-02**: 121 files enumerated and accounted, with historical DOC-01 inspection retained separately. No file is silently excluded or claimed unreadable. This partition, plus the three delta rows, gives every current target file a disposition without rewriting historical evidence.

The next discovery boundary is the named files and real consumers for each workstream. Consumer searches are lexical evidence, not proof that every dynamic or external consumer is absent. S10 must establish consumer closure before removing artifacts; S12 must characterize source-port consumers; S13/S14 must establish focused baseline evidence before replacing the parser/apply path. Other module families are outside this bounded iteration's conformance claim.

### Current evidence ledger

| Evidence ID | Exact inspected sources or observation | Finding supported / limits |
| --- | --- | --- |
| E11 | Read-only Git baseline, canonical-path/file-identity observations, tracker fingerprint, and current inventory comparison against §2 | Current clean baseline and 129 + 3 inventory; no product validation result |
| E12 | The 11 directly inspected Network Flow files listed above; `rg` caller/reference searches | Wrapper used by a test, `io.ReadAll`, materialized `Records`/`Rows`, late accepted-row limit, and `getAcceptedRowForLink` calling full-table `ListRows`; F19/F21/F22/F23 |
| E13 | `internal/modules/imports/source_streams.go`, `extension_facade.go`, `apply_coordination.go`; relevant `unit_outcomes.go` references; `internal/platform/postgres/db.go` | Source port loads `source_bytes` into memory; capability lookup also uses that loader; Imports owns the borrowed transaction, terminal outcome, cancellation and commit barrier; F20/F21 |
| E14 | `contracts/network-flow/index.json`, `frontend-entrypoints.v7.json`, `operation-policy.v1.json`; current frontend entrypoint reference in `contracts/protocol-ts/frontend-entrypoints.v2.json`; matching schemas and `tools/harness_schema_attachments.json`; generated artifact references in `internal/gen/contractnetworkflow/artifacts_gen.go` | Active major 8 uses operation-policy v2/entrypoint v8, while obsolete v7/v1 artifacts remain attached and packaged; F18. Generated output was read, never edited. |
| E15 | `internal/modules/extensions/catalogs.go` and `state.go`, historical ledger declarations in `contracts/extensions/fragments/network_flow_activity.profile.json`; existing semantic guard and test-family references | Ledger definitions authenticate committed rows without making migrations executable; retain them. Test/guard presence supplies routing evidence, not a fresh pass. |
| E16 | Current NF header and §§8.2/9.2/10/12.4/20/29; Core 01 analytical binding/unit contract; Core 04 database privilege rules; `docs/domain.md`, `docs/design.md`, `docs/research/nlspec-spec.md`; root/backend/tooling instructions and planning framework | Owner requirements, proposed-amendment status, vocabulary and supporting design doctrine. Supporting documents do not override behavioral owners. No presentation redesign or domain vocabulary change is needed. |
| E17 | Prior planning turn: `make help`, `make task-guide OWNER=module.networkflow`, `make task-guide OWNER=module.imports`, all exit 0 on the current baseline; `.markdownlint-cli2.jsonc`, generated Make entry and Markdown runner inspection | Public focused selectors and current canonical tracker glob. Discovery success is not test success; only Markdown lint is executed by this document-update step. |
| E18 | DOC-02 command, preservation and completion log in §20 | Documentation validation and final change-boundary evidence; no implementation acceptance |

### Contract decisions and adoption posture

The NF 8.0.0 header remains **proposed**, with unchanged requirements retaining adopted 7.0.0 baseline authority. DOC-02 records recommendations and does not adopt or edit any owner document. Owner-specific requirements govern behavior; machine contracts and tests remain projections. Human review establishes fidelity. No unresolved contradiction between different primary owners was established; the raw-cell diagnostic extension below is an explicit clarification/adoption gate, not a claim that code or a test can resolve owner policy.

| Contract ID / owner | Required or proposed decision | Disposition and dependent work |
| --- | --- | --- |
| C18 / NF §29, NF-REQ-210–212; Core 00 precedence and active contract publication | Keep private semantic validation in Network Flow; retain the active major-8 contract set and remove only demonstrably obsolete packaging obligations | Structural cleanup; F18/F19, S10. F10's historical decision to leave inert v7/v1 files is superseded only for those files and their exclusive projections. |
| C19 / Core 01 analytical binding Table 17.2-F and unit commit contract; NF §10 source capability | Imports owns opaque source access and generic orchestration; Network Flow receives bytes and source identity without Imports SQL or storage details | Preserve owner boundary and public capability/wire shape; F20, S12. Source integrity must be established before a successful preview/apply result. |
| C20 / NF §9.2 Table 9-B, NF-REQ-073/074 and §10 Table 10-A2 | Preview interprets the first `min(50, record_count)` complete data records. It may drain raw bytes for identity verification but must not report CSV/encoding errors confined to later records | Correct current whole-source normalization behavior in S13. Raw source read or digest failure is an integrity failure, not a later-record CSV diagnostic. |
| C21 / NF §12.4, NF-REQ-111–114, Tables 12-E/12-F; §20 and §21 | Enforce raw-cell limits for mapped, unmapped and ignored cells; preserve closed error tokens, exact scalar counts, attribution, hashes, redaction and ordered retention | Existing limit requirement plus proposed diagnostic-production clarification below; F23, S09/S13. RB-004 gates the new clarification. |
| C22 / NF §8.2, NF-REQ-057–062b; Core Imports atomic unit and cancellation contract | Validate before allocating a domain table. Reopen the source for bounded insertion into the existing borrowed transaction; source/summary mismatch, cancellation before commit or insertion failure rolls back all effects | F21, S14. Keep exactly one successful table and existing outcome/replay/commit ownership; no `creating`/`failed` table state. |
| C23 / NF indicator-link ownership and §29; NF-REQ-058; Core authorization/disclosure and transaction ownership | Read accepted rows by incident/table/row identity, retain supplied-reference checks and revalidate selected sources at commit | Structural retrieval change; F22, S11. Authorization and error precedence are characterization obligations, not permission to disclose hidden state. |
| C24 / Core 04 database role/privilege contract; NF §8.2 | Runtime has no DDL or temporary-table privilege. Do not add temporary tables, filesystem spooling, a durable staging format, or broader database grants | S12/S14 must work within current source/storage capabilities. No schema migration or state-version bump is planned. |

The proposed C21 clarification is: field-count mismatch emits its single row-level diagnostic first and skips target validation; otherwise emit one `cell_scalar_limit_exceeded` diagnostic per oversized source column, skip target validation for those cells, and validate remaining targets normally. Attribute a target only where the approved mapping supplies one; preserve exact decoded-scalar hashes and null safe samples. Run cross-field checks only when their operands validated successfully. Final diagnostic order remains NF-REQ-113. The conservative preview bound becomes `50 * (max_columns_per_csv + mapped_target_count + 1)`, with all terms bounded by the owner contracts. This replaces the insufficient target-field-only bound for the new raw-cell diagnostics; it neither changes diagnostic-ID bytes nor invents a parallel resource schema. Count and hash oversized scalars incrementally without retaining their full text. Adoption of this clarification is required before relying on it in S13.

Keep historical migration-ledger definitions because they authenticate retained ledger rows; keep owner-local harness controls with ordinary-build isolation/runtime scan inclusion; keep live Graph iterators, current request/resource schemas and stable test IDs. Historical numeric suffixes alone justify no removal or renaming. Graph, Reporting, Recovery, Indicators and Collaboration retain their distinct responsibilities. No broader compatibility decoder, migration route, generic validation framework or speculative module split is proposed.

## 17. DOC-02 finding dispositions

All F18–F23 statuses are **TODO**. Each is `must_fix` for its named workstream's exit; that classification does not authorize unrelated cleanup. No finding below is marked fixed by this document update. F01–F17 remain resolved as recorded in §14; F18 explicitly supersedes only F10's retention of obsolete packaged contracts.

| Finding / evidence / owner clauses | Remediation and affected areas | Rationale and expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation / workstream |
| --- | --- | --- | --- | --- | --- |
| F18 — Superseded contracts remain packaged; E14; C18 | Remove inactive `frontend-entrypoints.v7.json` and `operation-policy.v1.json`, their exclusive schema attachments/schemas, and generated copies through authored inputs and Make. Contracts, tooling, generated outputs and tests. | One active publication set reduces misleading declarations, validator/package weight and unnecessary compatibility maintenance. | Active major-8/public resource contracts remain unchanged. Preserve history in Git and this tracker; do not remove historical migrations or live ledger facts. | Old declarations continue to look supported and impose validation/packaging obligations despite having no active dispatch role. | V10/V16; S10. Consumer closure, active resolution, schema/generated checks, and absence of obsolete packaged copies. |
| F19 — Test-only forwarding helper in production; E12/E02-131/132; C18 | Remove `validatePreviewDiagnosticShape`; retarget its assertions to the real private decoder and public preview admission. Implementation and tests. | Eliminates a redundant production surface and keeps tests coupled to meaningful owner behavior. | No wire, storage or error-translation change. Do not remove the real strict decoder or weaken framing/null/member coverage. | A convenience test continues to sustain unnecessary production code and obscures the actual validation boundary. | V10/V14; S10. Valid/invalid shape cases and public safe failure translation remain covered without the helper. |
| F20 — Source-stream access materializes whole blobs; E13; C19/C24 | Read bounded byte chunks behind Imports' existing port; split metadata-only capability/precondition reads from content retrieval. Implementation, port documentation and tests. | Bounds application buffering at the actual upstream boundary and removes repeated blob allocation for metadata checks. | Public source capability/wire shape and existing storage remain; private reader integrity/closure obligations become explicit. No data conversion, new storage backend or privilege change. | Network Flow streaming still inherits a whole-file allocation, increased concurrent memory use and repeated unnecessary hashing/loading. | V12/V15; S12. Digest/size integrity, cancellation, close/error cleanup, isolated snapshots and bounded read sizes. |
| F21 — Apply retains complete parsed and accepted collections; E12/E13; C22/C24 | Replace full `Records`/accepted `Rows` retention with a summary-producing validation pass and a reopened-source insertion pass. Network Flow implementation, private store interface and integration tests. | Application memory depends on configured fields/diagnostics and bounded insertion, rather than total source/accepted row count. Retains one understandable atomic commit owner. | Preserve row/diagnostic/provenance identities, accepted order, counts, audit, replay and cancellation semantics. Extra I/O is an explicit tradeoff; no durable format or schema change. | Large imports retain several representations, enforce limits late and create avoidable memory pressure during transactions. | V13/V15; S14. Two-pass source/summary agreement, bounded buffering, complete rollback on late failure and correct committed-success recovery. |
| F22 — One-row linking reads a whole table; E12; C23 | Add private incident/table/row lookup and remove production `ListRows` after migrating its test-only consumers. Implementation and tests. | Narrows the store surface and makes lookup work proportional to selected rows instead of the enclosing table. | Same active-table, authorization, disclosure, supplied-reference and commit-time checks; no route or resource shape change. Live graph streaming remains. | A small authorized request can allocate and scan a large table; repeated bounded selectors multiply this cost. | V11; S11. Positive/missing/wrong-scope/deleted/stale-reference cases and deletion race retain their contract, with no full-table enumeration. |
| F23 — Parser/limits drift from owner behavior; E12/E16; C20/C21 | Use one private incremental CSV decoder for preview/apply; correct dialect, preview-tail handling, raw-cell limits, count timing and limit details. Owner clarification, implementation, tests and documentation. | Predictable bounded parsing and deterministic provenance/diagnostics; limits become effective before full collections are retained. | Correct erroneous behavior under existing owner definitions; proposed C21 clarification needs review/adoption. Existing state-5 tables are not rewritten; corrected interpretation requires original-source re-import. | Preview can fail on forbidden later-record errors; blank/CR-delimited input can be interpreted incorrectly; oversized cells and late limits defeat bounded processing. | V13/V14/V15; S09/S13. Independent dialect/provenance vectors, split-byte cases, exact bounds, privacy, preview publication and apply parity. |

## 18. DOC-02 workstreams, sequencing and readiness

Implementation sequence: **S09 -> S10 -> S11 -> S12 -> S13 -> S14 -> S15**. The graph is acyclic and each workstream has one immediate predecessor. All workstream statuses are **TODO**. None is implementation-complete because its plan is written. Later implementation authorization and baseline revalidation are common entry conditions; readiness below describes technical prerequisites separately from that authorization.

| Workstream | Depends on / finding | Current readiness | Principal risk | Binary exit and successor |
| --- | --- | --- | --- | --- |
| S09 Contract closure and characterization | DOC-02 handoff; F18–F23 | ready to start under later authorization | Freezing defects as intended behavior or treating a proposal as adopted | Owner decisions/classification, registered characterization and explicit review gate recorded; V10–V15 entry evidence scoped; checkpoint before S10 |
| S10 Retire obsolete surfaces | S09; F18/F19 | blocked by S09 and missing characterization/consumer closure, RB-005 | An overlooked consumer depends on a removed artifact | V10/V16 pass; obsolete copies absent and active contracts resolve; checkpoint before S11 |
| S11 Narrow accepted-row retrieval | S10; F22 | blocked by predecessor and V11 baseline, RB-005 | Disclosure or missing-row precedence changes | Exact lookup with retained transaction checks and passing V11; checkpoint before S12 |
| S12 Bound Imports source reading | S11; F20 | blocked by predecessor and V12 source-port characterization, RB-005 | Connection leaks, integrity translation drift or inefficient database chunking | V12 and affected Imports consumers pass; application bound and database cost distinguished; checkpoint before S13 |
| S13 Incremental CSV and bounded preview | S12; F23 | blocked by predecessor, RB-004 and V13/V14 characterization | Parser/provenance changes or invalid preview bounds | C21 clarification gate satisfied; V13/V14 pass, V15 apply regression subset passes; checkpoint before S14 |
| S14 Streaming atomic apply | S13; F21 | blocked by predecessor and V15 entry evidence, RB-005 | Extra I/O/transaction duration or partial effects after late failure | V15 passes with bounded storage insertion, source/summary agreement and atomic outcomes; checkpoint before S15 |
| S15 Focused validation and handoff | S14; F18–F23 | blocked by all preceding exits | Reusing historical checks as final acceptance | Finalized source, missing focused obligations completed, findings/evidence/compatibility reconciled; no remediation blocker remains |

### S09 — Contract closure and characterization

Re-run applicable owner guidance and revalidate the exact named sources before editing. Record C18–C24 dispositions in the behavioral owners only where clarification is necessary; keep amendments proposed until the owner-review/adoption gate is satisfied. In particular, resolve RB-004's raw-cell diagnostic production and preview bound without changing the existing ID algorithm or resource shape. Keep Core Imports' source/transaction responsibility and Core 04 privilege restrictions explicit.

Register focused, clearly named behavioral selectors through authored verification inputs and Make generation. Capture passing entry evidence for retained valid preview/result admission, source integrity/error translation, row lookup disclosure and atomic import/replay/cancellation behavior. Add correction-specific cases with their correcting slices; do not encode erroneous permissiveness, whole-file allocation, or the wrong CSV dialect as permanent expectations. Test public contracts and meaningful resource invariants, not arbitrary private error text.

Exit requires consistent owner decisions, an explicit proposed-versus-adopted distinction, registered characterization obligations, and a recorded adoption gate for dependent changes. S09 may record an outstanding adoption gate while completing independent characterization; S13 remains blocked until RB-004 closes. Failure evidence and the S09 checkpoint precede S10.

### S10 — Retire obsolete surfaces

Establish references for the two obsolete contract files, their matching schemas, schema attachments, generators, generated packaging and active Go/browser consumers. Remove the authored obsolete artifacts and exclusive attachments/schemas; regenerate, never hand-edit generated copies. Keep entrypoint v8, operation-policy v2, current request/resource IDs, stable test rows, historical SQL and ledger facts. Do not create aliases or dual-major dispatch.

Remove only the forwarding `validatePreviewDiagnosticShape` production wrapper; move its assertions to `decodePreviewDiagnostic` or the public facade as appropriate. Preserve strict framing, forbidden-null/member checks, valid roundtrips and safe Imports failure translation. V10/V16 cover consumer closure, compilation, publication, schema/generated consistency and boundary behavior. Any discovered live consumer must be addressed within the active owner contract before deletion; absence from one lexical search is insufficient.

Exit is a smaller active artifact/production surface with no unresolved consumer and passing focused checks. Rollback is the coordinated authored-input and generated-output source change; no data rollback is required. Record the checkpoint before S11.

### S11 — Narrow accepted-row retrieval

Replace full-table retrieval with a private exact lookup constrained by incident ID, table ID and row ID. Preserve authorization before state disclosure, active-table admission, supplied source-row-number/mapping-fingerprint checks, bounded selector limits and commit-time source revalidation. Keep the existing indicator-link application and transaction owners.

Migrate test callers that used bulk retrieval for convenience to scoped assertions or test-only helpers, then remove the now-unused production `ListRows`. Do not replace live graph streaming paths or move SQL into a generic module. V11 covers success, wrong incident/table, absent row, soft-deleted table, stale supplied reference and deletion races, and demonstrates that unrelated rows are not retrieved. Exit and checkpoint precede S12; rollback is source-only and preserves stored data.

### S12 — Bound Imports source reading

Implement a context-aware reader over bounded database byte chunks from one consistent query snapshot, inside Imports. Metadata and byte chunks must refer to the same admitted source identity; separate capability/precondition lookup must not select or hash the full blob. Keep the existing public capability and wire contracts, SQL ownership and storage format. Use fixed internal chunk bounds rather than a new public tuning surface.

Verify exact byte count and digest before successful consumption completes. EOF, read error, cancellation and explicit close must release query/connection resources; close must be safe after EOF and must not falsely certify an incompletely consumed source. Preview will drain raw bytes before success; failed parsing may close early without treating that source as verified. Distinguish safe source-integrity failures from CSV semantics through existing owner error boundaries without exposing raw bytes or causes.

V12 covers fragmented/empty reads, size/digest mismatch, early close, repeated close, cancellation, concurrent snapshots and cleanup. Include focused real Imports consumers so moving integrity checking into consumption cannot silently weaken admission. Demonstrate bounded application buffering and inspect database-side chunk cost separately; a bounded Go reader is not evidence of bounded PostgreSQL memory or acceptable repeated blob decompression. No broad benchmark certification is claimed. Exit/checkpoint precedes S13. No storage migration or privilege expansion is permitted by this design.

### S13 — Incremental CSV and complete bounded previews

Use one Network Flow-private decoder for both modes. Implement Table 9-B rather than assuming the standard CSV reader has the same dialect: CR/LF/CRLF record delimiters, blank logical records, quoted line breaks, opening/closing/escaped quotes, BOM placement and incremental UTF-8 validation. Preserve exact decoded field content, logical record numbering, header descriptors and source/provenance digests. Handle boundaries inside UTF-8 sequences, BOMs, quotes and delimiters.

Enforce column/header/raw-cell/row limits at owner-defined stages and use the exact closed error details. For oversized fields, retain bounded content only, continue counting/hash computation as required for the exact scalar diagnostic, and reject the row without persisting the oversized value. A malformed unterminated record while seeking the next logical record retains the Table 9-B failure precedence. Stop count processing once one excess item establishes the count limit; do not scan for an unnecessary total.

After RB-004 closes, implement C21's diagnostic production, attribution and revised bound through the existing shared validator. Preview keeps only bounded records, descriptors, samples and diagnostics; it stops semantic parsing after the 50th complete data record. Drain the remainder only as raw bytes for source identity. Later malformed CSV, misplaced BOM or invalid UTF-8 must not become preview diagnostics, while read/identity failures must still prevent success. Preview remains side-effect free and always reports complete in-scope diagnostics with `diagnostics_truncated=false`.

V13/V14 use independent expected dialect/digest vectors, chunk boundaries, malformed input, 50/51 cases, limits at/below/above the bound, oversized-cell privacy and valid public-facade roundtrips. Preserve the existing apply path's contracts while it consumes the shared decoder until S14 removes its remaining materialization; do not claim end-to-end apply memory completion at S13. Checkpoint before S14.

### S14 — Streaming atomic apply

Replace whole-file `Records`/accepted `Rows` collections and `preparedImportApply`'s retained parser payload with a private validation summary. First pass parses and validates the source, retaining accepted/rejected counts, descriptor/source/mapping identities, the configured diagnostic prefix, actual omission state and the bounded all-rejected diagnostic sample. No domain table is allocated until this pass succeeds and admits at least one row. Accepted-row limits are enforced during validation, not after a full collection has been built.

Second pass reopens the same capability and streams accepted rows in source order through the private storage adapter into Imports' existing borrowed transaction. Keep insertion buffering bounded independently of total rows; do not collect input merely to sort rows already produced in owner order. Recompute source identity, mapping/descriptor agreement and summary agreement before returning success. The same approved mapping and pure validation path must govern both passes. A changed source, read/insert failure, summary mismatch or cancellation before commit must roll back table, rows, diagnostics, audits and outcome publication together.

Retain Imports' final commit/cancellation barrier and replay ownership. A proven committed success wins over later cancellation; no Network Flow facade may independently commit the borrowed transaction. Preserve count/diagnostic limits, IDs, provenance and audit behavior. Use no temporary tables, filesystem spool, durable staging format or broader database privileges. Additional reading/validation and time spent in the transaction are explicit tradeoffs for bounded application memory and simpler lifecycle ownership.

V15 covers zero/exact/overflow limits, partial/all-rejected input, multiple diagnostics per row, source changes between passes, source read errors, insertion failure, cancellation before commit, committed-result recovery and exact replay. Resource checks must demonstrate that retained row collections do not grow with source length, separately from the bounded diagnostic allowance and database costs. Exit/checkpoint precedes S15. Reverting source does not justify modifying already committed state or changing the release's admission boundary.

### S15 — Focused validation and handoff completion

Audit actual obligation coverage rather than rerunning entire suites by name. After focused slice checks, run repository-required `make agent-finalize`, review any generated changes, freeze source and execute only the remaining necessary focused selections. Reassess affected obligations after any repair. **Do not run full CI.** Add no unrelated visual, benchmark, release or full-owner qualification.

Reconcile F18–F23, RB-004/RB-005, all slice statuses, exact commands/owning exits/run roots, environment, finalized source identity, generated changes and current compatibility notes. Every deletion needs consumer evidence; bounded application-resource claims need direct evidence; database-side cost and validation limits must be stated. Historical evidence supports diagnosis but cannot substitute for final compatible acceptance.

Exit requires all preceding slices complete, required focused validation passing, no remediation blocker remaining, and a handoff that separates implementation completion from human adoption and operational deployment. Preserve the §15 operational original-source and untouched-environment rollback instructions, with the current release boundary restated. No destructive reset, deployment or data conversion is an S15 task.

## 19. DOC-02 validation obligations and evidence routing

Only V18 is executed for this document update. V10–V17 are future implementation obligations, with **no current pass claimed**. Run public Make targets from the repository root. Re-run owner guidance when entering an affected slice; select exact rows through `make target-plan` and `make test-slice` instead of running complete owners by default. New cases must be registered through authored verification inputs and Make generation. Tests, generators, runtime metadata, conformance and release evidence must never read, stat or hash Markdown.

| Validation ID | Required cases / evidence | Selection and owning workstreams |
| --- | --- | --- |
| V10 Active surfaces and dead-helper removal | Active major-8 resolution, obsolete artifact consumer closure and generated removal; real strict decoder and safe preview failure translation still exercised | S09/S10; NF preview/admission and binding rows, affected contract publication rows; source/schema/generated checks under V16 |
| V11 Exact accepted-row lookup | Positive row, wrong incident/table, missing row, soft-deleted table, stale supplied reference, deletion race; no unrelated row enumeration | S09/S11; focused NF indicator-link, storage and transaction consumers |
| V12 Imports source-port lifecycle | Bounded chunking, metadata-only lookup, digest and byte-size mismatch, fragmented/empty input, close/EOF/error/cancellation release, concurrent source snapshot isolation and existing consumer error translation | S09/S12; focused Imports unit/service tests plus analytical-owner consumers; distinguish application memory from database costs |
| V13 Incremental parser, limits and provenance | Independent CR/LF/CRLF, blank/quoted/escaped-record and digest vectors; UTF-8/BOM/delimiter splits; malformed framing; header/column/raw-cell/row/accepted-row bounds and exact error details | S09/S13/S14; focused NF parser/owner-error tests, with correction cases added alongside implementation |
| V14 Complete bounded preview and diagnostic admission | 50/51 records, later-record errors suppressed, raw integrity failures still rejected, complete diagnostics including raw-cell failures, correct null attribution/hash/redaction, 0/N/N+1 apply retention parity, both Imports preview publication and approval consumers | S09/S10/S13; dedicated NF preview admission and semantic guard rows plus focused Imports consumers |
| V15 Atomic streaming apply | Valid/partial/all-rejected imports, counts independent of diagnostic multiplicity, two-pass identity/summary agreement, source mutation/read failure/insertion failure/cancellation rollback, committed-success recovery, replay and bounded row buffering | S09/S12/S14; focused NF and Imports service-backed atomic-unit/outcome tests; no temporary-table/DDL privilege requirement |
| V16 Generated/publication/boundary consistency | Authored removals regenerate cleanly, active Go/browser resolution remains, no compatibility aliases, semantic source entrypoints retain dependency direction | S10 and later affected slices: `make generate`, `make source-check CHECK=schema`, `make source-check CHECK=generated`, `make lint CHECK=backend-boundary`; generated files never hand-edited |
| V17 Final focused acceptance | Finalization changes reviewed, source frozen, missing focused obligations selected, exact failure/repair evidence and cleanup outcomes reconciled | S15: `make agent-finalize`, then only required focused checks identified by current plans; `make ci` excluded by user instruction |
| V18 Documentation checkpoint | Tracker selected by canonical module-tracker glob; structural Markdown lint passes; only admitted tracked file changed | DOC-02 and each documentation checkpoint: `make lint-markdown`; current command/root/result recorded in §20 |

Existing selector anchors inspected on this baseline include `module.networkflow.unit.import_preview_admission_f0dc5f2d8b`, `module.networkflow.unit.query_authoring_admission`, `module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99`, `module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`, `module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`, `module.imports.integration.uncommitted_unit_cancellation`, and `module.imports.integration.unit_commit_crash_recovery_108cd054a6`. They are candidate evidence routes, not proof that every new case is already covered. The authored family inputs are `tools/test_families/module.networkflow.json` and `tools/test_families/module.imports.json`; no row is invented or changed during DOC-02.

For future failures record the exact target/selection, owning exit, run root or summary artifact, and relationship to the change **before** repair or rerun. A failed exit leaves its slice incomplete and blocks successors. Broaden only for an identified missing obligation. No standalone migration check is needed merely for this planned source-only iteration; if an unexpected required schema change emerges, record and resolve the scope/owner decision before proceeding instead of silently adding a migration. Human owner review remains independent of all automated results.

## 20. DOC-02 blockers, completion and planning handoff

### Gates and authorization

| Gate | Status | Dependent work / closure evidence |
| --- | --- | --- |
| RB-004 Raw-cell diagnostic-production clarification and revised preview bound | BLOCKED — proposed owner decision awaits review/adoption | S09 documents and presents C21; S13 and dependent S14/S15 cannot rely on the new rule until the owner gate is satisfied. Existing unchanged owner requirements remain authoritative. |
| RB-005 Missing executed characterization for the new structural/streaming seams | BLOCKED — tests were inspected, not run for this iteration | S09 establishes and records baseline obligations; each dependent S10–S14 slice consumes its exact compatible evidence. Code similarity and prior S08 passes do not close this gate. |
| Major-8/state-5 owner adoption and coordinated deployment | DEFERRED — human/operational release gates | Carry forward §15's proposed-amendment review and coordinated server/browser publication. Implementation completion alone cannot close these gates. |
| Future implementation authorization | Not granted by DOC-02 | This invocation updates the tracker only. S09–S15 require a later implementation task; the authorization column is not a work-status claim. |

Planning completeness is **complete for the declared bounded iteration**: V18 passed and the preservation audit confirms the tracker is the sole changed tracked file. S09 is ready to start under later authorization and baseline revalidation; subsequent readiness is blocked by the named predecessors, characterization and adoption gates. These known future prerequisites do not imply that implementation has started. There is no claim of exhaustive module conformance, production deployment readiness, full-owner acceptance or a new full-CI pass.

### Planning session and command log

| Session / action | Evidence and result | Next action |
| --- | --- | --- |
| DOC-02 read-only planning turn, 2026-10-29 | Current baseline inspected; user selected cleanup plus hardening and included Imports reader changes. `make help` and both `make task-guide OWNER=module.networkflow` / `OWNER=module.imports` returned 0. No files changed or product tests ran in that turn. | Authorized tracker-only update |
| DOC-02 update admission, 2026-10-29 | E11: baseline unchanged, clean `main`, safe regular tracker identity; original inventory comparison yields 129 retained paths plus E02-130–132. Root/backend/tooling guidance, skills and relevant owners consulted. | Write current scope, findings and workstreams while retaining historical evidence |
| DOC-02 tracker checkpoint | §§1/14/15 distinguish completed history; §§16–20 record current inventory, F18–F23, C18–C24, S09–S15, V10–V18 and gates. All implementation statuses remain TODO. | V18 and final preservation audit |
| DOC-02 V18 | PASS, owning exit 0: `CARTULARY_TEST_RUN_ID=nf-doc02-plan-markdown-20261029-01 make lint-markdown` from repository root. Emitted root `.cartulary/test-results/nf-doc02-plan-markdown-20261029-01`; summary `adhoc/lint-markdown/tool-run-summary.json` reports pass, zero failures. The module-tracker glob selects this artifact. | Record result and perform final post-log documentation check |
| DOC-02 preservation audit | `git --no-optional-locks diff --check` returned 0. Diff/status show only this tracker changed; no staged changes or non-ignored untracked files. Index SHA-256 still equals E11's starting value. Historical inventory and evidence retained; no production refactor performed. | Later authorized S09 after baseline revalidation |

The completion-log edit is followed by the same documentation-only target with fresh ID: `CARTULARY_TEST_RUN_ID=nf-doc02-plan-markdown-20261029-02 make lint-markdown`. Its exact evidence location is `.cartulary/test-results/nf-doc02-plan-markdown-20261029-02/adhoc/lint-markdown/tool-run-summary.json`; the owning terminal result is reported with the handoff. This final confirmation adds no product verification or finalization.

No product tests, formatter, generator, `agent-finalize`, full CI, visual suite, benchmark or deployment is run for this document-only update. These are intentionally outside the current execution scope, not skipped implementation acceptance checks. Ordinary ignored lint cache and retained-run artifacts are permitted. No source, tests, contracts, migrations, generated files or dependencies may be changed by DOC-02.

### Binary document-completion criteria

| Criterion | Status / evidence |
| --- | --- |
| Current clean baseline and 132-file universe recorded without presenting targeted review as exhaustive | DONE; E11, §16 |
| Original inventory, F01–F17, S00–S08, failures and operational handoff preserved as history | DONE; §§1–15 retained with explicit historical labels |
| F18–F23 include owners, remediation areas, rationale/benefit, compatibility, unresolved risk and validation | DONE; §§16–17 |
| S09–S15 form an acyclic sequence with separate risks, readiness, exits and tracker checkpoints | DONE; §18 |
| Proposed amendments, human adoption, technical readiness and implementation authorization separated | DONE; C21/RB-004 and this section |
| Focused validation and no-full-CI instruction carried forward; no Markdown-dependent product verification introduced | DONE; §19 |
| Canonical Markdown lint and final only-tracker preservation audit | DONE; V18/E18, owning lint exit 0 and unchanged index/only-tracker diff |

The later implementation handoff must explicitly state that older Network Flow data was **neither converted nor deleted**, that original CSV sources and mapping decisions are required for re-import, and that matching major-8 server/browser assets with state 5 are one coordinated release boundary. Correcting source interpretation does not rewrite existing state-5 tables. Rollback uses the untouched previous environment with its matching database, object store, keys and assets; do not downgrade a database after state-5 writes, remove metadata or delete data to force admission. Human adoption and deployment remain separate from engineering completion.

## 21. IMP-03 controlling implementation plan and execution

### Decisions and finding delta

The user approved the revised plan for implementation, including Imports storage migrations and analytical source authority based on original bytes and approved owner mapping. Explicit human owner adoption precedes dependent code. Keep major 8/state 5, focused acceptance and S15 completion; no full CI, deployment, destructive reset or historical Network Flow conversion is part of this effort.

F18–F23 retain §17's remediation, rationale, risks and validation, with these amendments:

- F20: migrate `import_source_streams.source_bytes` to PostgreSQL `EXTERNAL` storage and physically reserialize existing values without changing logical bytes. Use private fixed 64 KiB reads, metadata-only capability queries, owned consistent snapshots for ordinary reads and transaction-aware borrowed reads for apply. This replaces the no-migration/no-internal-port-change assumptions in C19/C24. Runtime receives no additional privilege.
- F21: resource acceptance includes the real Imports orchestration path, not merely the Network Flow facade.
- F23/C21: obtain adoption for raw-cell diagnostic production and the complete preview bound `50 * (max_columns_per_csv + mapped_target_count + 1)` before dependent correction.
- F24 (new, must_fix): Imports' `getApplyUnits`, `insertApplyUnitPlansTx` and `lockApplyUnitTx` fetch/retain generic `source_rows_json` before analytical application. Use ordered scheduling metadata, target-specific admission and lazy generic-row loading. Analytical content authority is its original source stream/digest and approved owner mapping/fingerprint plus target binding. Migrate analytical plan `source_rows_sha256` to null; retain the required digest for generic targets. Areas: specification, implementation, migrations, tests and documentation. Benefit: eliminate hidden full-source allocations and duplicate interpretation. Unresolved risk: F20/F21 cannot establish bounded analytical apply, and analytical content remains coupled to generic parsing. Compatibility: no public response change; migrate internal plan metadata with workers quiesced, preserve outcomes/replay identities, and prohibit old binaries on the migrated schema. V19 requires no analytical generic-row reads, source/mapping failure coverage, generic regression, migrated-plan behavior and cancellation/recovery.

### Workstreams and binary exits

| Workstream | Status | Depends on | Implementation / principal risk | Binary exit |
| --- | --- | --- | --- | --- |
| S09 Specification closure and characterization | DONE | Baseline and this checkpoint | Candidate explicitly adopted by the user on 2026-10-29; focused characterization passed. | Explicit owner adoption; baseline evidence and later obligations mapped; validated tracker checkpoint. |
| S10 Retire obsolete surfaces | DONE | S09 | Remove inactive v7/v1 publication, exclusive schemas/attachments and diagnostic forwarding wrapper after consumer closure; regenerate. Risk: overlooked live consumer. | V10/V16 pass; active publication intact; all deletions evidenced. |
| S11 Narrow accepted-row retrieval | DONE | S10 | Exact incident/table/row lookup and scoped tests; retain commit revalidation. Risk: disclosure or deletion-race drift. | V11 passes; production bulk retrieval removed; graph iteration intact. |
| S12-M Source storage migration | DONE | S11 | Imports-owned EXTERNAL-storage migration with byte-preserving rewrite and regenerated ownership/projections. Risk: locks, disk/WAL use and surviving compression. | Empty/populated migration, byte/digest equality, uncompressed storage, region copying, Recovery compatibility and migration drift pass. |
| S12 Bounded source readers | DONE | S12-M | Fixed 64 KiB reads; metadata lookup; owned snapshot and `OpenSourceStreamTx`; digest/size validation and complete resource lifecycle. Risk: leaked connections, lock order, false EOF or changed error translation. | V12 passes, including small-pool/concurrent/cancellation cases and database cost evidence. |
| S12-I Analytical admission and scheduling | DONE | S12 | Metadata-only scheduling; separate analytical/generic admission and unit reads; nullable analytical plan row digest migration. Risk: weaker admission or replay/mixed-version failure. | V19 passes; analytical SQL does not fetch/hash generic rows; generic and migration regressions pass. |
| S13 Incremental CSV and complete preview | DONE | S12-I; adopted C21 | Shared owner-dialect decoder, bounded fields, exact scalar/hash accounting, full in-scope diagnostics and raw-only tail drain. Risk: boundaries, provenance, precedence and privacy. | V13/V14 plus affected apply regression pass; no end-to-end apply claim yet. |
| S14 Streaming atomic apply | DONE | S13 | Summary validation followed by protected-source, one-row-at-a-time insertion in borrowed transaction. Risk: longer transactions and partial late-failure effects. | V15 passes on real Imports path; source/mapping/descriptor/summary agreement and atomic outcomes demonstrated. |
| S15 Focused validation and handoff completion | DONE | S14 | Finalize, review generation, freeze source and fill focused acceptance gaps. Risk: stale evidence presented as current acceptance. | All findings/workstreams reconciled, required checks pass, no engineering blocker, complete operational handoff. |

S09's diagnostic rule: field-count mismatch emits one row-level diagnostic and skips targets; otherwise emit one raw-cell-limit diagnostic per oversized column (mapped, unmapped or ignored), skip those cells' target validation and validate remaining targets normally. Attribute targets only through approved mappings. Cross-field checks require valid operands. Preserve diagnostic IDs, exact decoded scalar hashes, null unsafe samples and owner ordering.

S12 ordinary readers own a read-only consistent snapshot; borrowed readers hold source protection but never commit/roll back the caller transaction. Each bounded query finishes before returning data, allowing writes on the same connection. Successful EOF verifies size/digest. Error, cancellation, EOF and close release owned resources; early close is not integrity success. Preserve existing safe error boundaries.

S13 implements CR/LF/CRLF, blank records, literal/quoted/escaped/closing quotes, quoted line breaks, incremental UTF-8/BOM and logical numbering. Excess scalar text is discarded while counting/hashing continues. Stop semantic preview parsing after record 50; drain only raw bytes for integrity. Later CSV/encoding defects are outside preview diagnostics; integrity/read failures still prevent success.

S14 retains only counts, identities, descriptors/mapping agreement, configured diagnostic prefix and bounded all-rejected sample in validation. Enforce accepted limits while validating; allocate no table before validation succeeds with an accepted row. Reopen protected source and insert in canonical source order without collecting/sorting all rows or querying within COPY callbacks. Verify both-pass identity/summary agreement before success. All precommit failures roll back table/rows/diagnostics/audit/outcome; Imports owns commit, cancellation barrier and committed-success recovery.

### Validation, evidence and handoff requirements

V10–V18 remain defined in §19; V12 additionally includes physical storage migration and transaction-aware readers; V19 is the F24 admission/scheduling/migration obligation above. Discover exact current selections via `make task-guide` and `make target-plan`, then run `make test-slice` with fresh run IDs. Register tests in authored inputs and regenerate via Make. Required affected checks include schema/generated consistency, backend boundaries, Go checks and `make migration-drift`. Run `make lint-markdown` at checkpoints. Run `make agent-finalize`, review generated changes and freeze source before final focused acceptance. No full CI by default.

Every checkpoint records source identity, exact commands/selections, owning exits, run roots, cleanup, failures/repairs and skipped checks/reasons. Failure leaves its workstream incomplete. Separate application-memory evidence from database cost observations. No product evidence consumes Markdown.

Operational handoff requires quiesced import workers and a recoverable database snapshot before new migrations; publish matching server/browser assets and prohibit mixed workers around the changed plan schema. Existing Network Flow tables are neither reinterpreted nor deleted. Physical source migration preserves bytes; plan migration removes obsolete analytical metadata. Corrected interpretation requires original CSV and mappings for re-import. Upload/discovery and Recovery-wide memory are outside this preview/apply resource claim. Rollback uses the previous matching environment snapshot, never downgrade/delete-to-admit after new writes. Deployment remains separate from S15 engineering completion.

### Execution log

| Checkpoint | Changes and evidence | Result / next gate |
| --- | --- | --- |
| IMP-03 initial tracker update, 2026-10-29 | Revalidated `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`; preserved pre-existing tracker changes and all history. Added F24, S12-M/S12-I, revised source/storage decisions and explicit adoption gate. | S09 IN_PROGRESS; no product changes or validation passes claimed. |
| S09 owner candidate | Prepared Network Flow §9.2/§12.4 and Core 01 REQ-01-620b amendments; Core 00 §1.3 names the complete coordinated adoption candidate and companions. Added fragmented-reader, exact-digest and source-read-error characterization under existing registered `TestImportPreviewAdmission`. | Reviewable candidate prepared; human adoption still pending. No dependent implementation started. |
| S09 discovery | `make help ALL=1`, `make task-guide OWNER=module.networkflow`, `make task-guide OWNER=module.imports`, and `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s09-unit-selection.json DETAIL=rows` / `INPUTS=/tmp/cartulary-nf-s09-imports-selection.json`, all exit 0. | Existing semantic rows cover baseline obligations; new correction cases accompany their slices. |
| S09 format and documentation | `CARTULARY_TEST_RUN_ID=nf-s09-format-20261029-01 make format` and `CARTULARY_TEST_RUN_ID=nf-s09-markdown-20261029-01 make lint-markdown`, owning exits 0. Roots `.cartulary/test-results/nf-s09-format-20261029-01` and `.cartulary/test-results/nf-s09-markdown-20261029-01`; Markdown summary `adhoc/lint-markdown/tool-run-summary.json`. | PASS; subsequent status/document edits receive a final checkpoint lint. |
| S09 Network Flow baseline | `CARTULARY_TEST_RUN_ID=nf-s09-baseline-nf-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99`; owning exit 0, root `.cartulary/test-results/nf-s09-baseline-nf-20261029-01`. | PASS; preview admission/privacy/integrity, semantic boundaries and current CSV baseline. |
| S09 Imports baseline | `CARTULARY_TEST_RUN_ID=nf-s09-baseline-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6`; owning exit 0, root `.cartulary/test-results/nf-s09-baseline-imports-20261029-01`. | PASS; real preview/approval, authorization, unit atomicity, safe translation, cancellation and committed-result recovery. |
| S09 lookup baseline | `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s09-lookup-selection.json DETAIL=rows`, exit 0; `CARTULARY_TEST_RUN_ID=nf-s09-baseline-lookup-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37`, owning exit 0, root `.cartulary/test-results/nf-s09-baseline-lookup-20261029-01`. | PASS; table lifecycle/limits, indicator admission, graph contributors and bounded-source regressions provide S11 entry evidence. |
| S09 evidence review | `make explain-run RESULTS_DIR=.cartulary/test-results/nf-s09-baseline-nf-20261029-01`, corresponding `baseline-imports` and `baseline-lookup` roots, all exit 0. NF and Imports source identity `sha256:7c4fc60cf4e93e53779c3214bd00d426295e932c8eac8214206f25eb6ed77ded`; NF 1/1 and Imports 5/5 work units, no failed/skipped/canceled units; their cleanup has 10/20 steps with zero failed or blocked. | Baseline passes are characterization, not implementation acceptance. Exact lookup run identity/cleanup remains in its named artifacts. |
| S09 adoption review gate | Prepared candidate linked to the user with an explicit adoption request after baseline Network Flow/Imports passes. `CARTULARY_TEST_RUN_ID=nf-s09-review-markdown-20261029-01 make lint-markdown` passed with owning exit 0; summary `.cartulary/test-results/nf-s09-review-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json`. | BLOCKED pending human adoption under the user's S09 gate. S10 and successors remain TODO. No production behavior, migration or generated artifact changed. |

S09 resumption: obtain an explicit adoption answer for Core 00 §1.3's prepared
candidate; if adopted, reconcile candidate/companion status headers and record
the human decision before completing S09 and beginning S10. If changes are
requested, revise the concrete candidate and its affected characterization.
Do not infer adoption from elapsed time, passing tests, or the implementation
request itself. No repeated product baseline run is required without a relevant
source change. The final checkpoint runs
`CARTULARY_TEST_RUN_ID=nf-s09-checkpoint-markdown-20261029-01 make lint-markdown`;
its summary is
`.cartulary/test-results/nf-s09-checkpoint-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json`.
Generation, migrations, `agent-finalize` and final acceptance are pending their
authorized later slices, not claimed complete or waived. Full CI is excluded by
the focused plan. All pre-existing tracker text remains preserved as history.


### S09 completion and adoption checkpoint

The user explicitly answered **“Adopt the prepared candidate and continue
implementation”** to the S09 adoption request on 2026-10-29. Core 00 §1.3 and
Network Flow/Extensions/Core companion headers now record adoption. RB-004 is
closed; RB-005 entry characterization is satisfied by the three named baseline
roots (same product source identity, owning exits 0). Lookup baseline has 5/5
work units and 18 cleanup steps, all passing. The historical pending-review
entries above are superseded by this explicit human decision, not erased.
S09 is DONE. The S09 checkpoint Markdown command passed with exit 0; the
adoption status update is checked by
`CARTULARY_TEST_RUN_ID=nf-s09-adoption-markdown-20261029-01 make lint-markdown`
before S10 starts. Deployment remains unauthorized and separate.


### S10 execution

S09 adoption Markdown checkpoint passed (owning exit 0). Consumer closure:
active contract index selects operation-policy v2; protocol entrypoint selects
v8; removed v7/v1 references were confined to their authored declarations,
exclusive schema attachments and generated packaging. Contract generation
walks the authored Network Flow artifact directory; JSON shape checking uses
those schema attachments. Historical ledger/state inputs are retained.
Removed the two obsolete declarations, their exclusive schemas/attachments and
the test-only diagnostic forwarding wrapper. Tests now call the real decoder;
public preview admission and safe Imports translation remain selected evidence.


S10 exit: V10/V16 PASS. All commands used explicit `CARTULARY_TEST_RUN_ID`:
`nf-s10-generate-20261029-01` / `make generate`;
`nf-s10-nf-20261029-01` / `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700`;
`nf-s10-translation-20261029-01` / `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`;
`nf-s10-schema-20261029-01` / `make source-check CHECK=schema`;
`nf-s10-generated-20261029-01` / `make source-check CHECK=generated`;
`nf-s10-boundary-20261029-01` / `make lint CHECK=backend-boundary`.
Each owning exit is 0; roots are `.cartulary/test-results/<run-id>`.
Generation removed obsolete packaged copies/validators and renumbered generated
AJV identifiers downstream of the deleted schemas; this is expected generated
churn. No authored or generated active reference remains to the removed schema
IDs or helper. F18/F19 are fixed. Source rollback is coordinated authored and
generated changes, with no data rollback. Checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s10-checkpoint-markdown-20261029-01 make lint-markdown`.


### S11 execution

S10 checkpoint Markdown passed (exit 0). Replaced `ListRows` with an exact
incident/table/row lookup constrained to active tables; missing rows retain
selector errors and table deletion is rechecked on a missing read. Supplied
references and commit-time source validation remain with the existing link
application/transaction owner. Migrated all three test-only bulk consumers to
known row identities. Graph streaming remains unchanged. Scoped missing,
wrong-incident and wrong-table cases extend the existing registered store row.


V11 passed with `CARTULARY_TEST_RUN_ID=nf-s11-lookup-20261029-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37`, owning exit 0. Root: `.cartulary/test-results/nf-s11-lookup-20261029-02`; all 5 work units passed, cleanup 18 steps without failures. The first invocation (`nf-s11-lookup-20261029-01`, same command/selection) exited 2 with harness `infra/preflight_error`: a new test misspelled local variable `racing` as `tracing`, preventing compilation. Corrected that typo before the fresh successful invocation; no failed assertion was hidden. Both roots retain their source identities and diagnostics. Formatting runs `nf-s11-format-20261029-01` and `nf-s11-format-20261029-02` passed `make format`.

The database guard test rejects enumerating queries and simulates deletion between table admission and row retrieval. Production `ListRows` references are absent. F22 is fixed; graph iteration, supplied-reference admission and commit-time revalidation remain covered. No remaining S11 engineering risk blocks S12-M. Source rollback restores the prior implementation without data conversion. Checkpoint: `CARTULARY_TEST_RUN_ID=nf-s11-checkpoint-markdown-20261029-01 make lint-markdown`.


### S12-M execution

S11 checkpoint passed (owning exit 0). The physical migration retains the existing source table, grants, references, digest and Recovery identity. Only compressed existing bytea values require reserialization; EXTERNAL is the policy for future writes. Operational application requires quiesced workers, a recoverable snapshot, and disk/WAL headroom; no deployment is performed here.


S12-M exit: PASS. Migration 31 is Imports-owned, sets EXTERNAL storage and reserializes only compressed legacy values by bytea concatenation. Existing source references, bytes, digests, grants and table identity are preserved. The populated migration test begins at version 30 with genuinely compressed binary content, checks exact bytes/digest/size after migration, and checks an operator-region-style INSERT SELECT copy. Empty bootstrap and repeated migration also pass. Recovery database coverage and actual operator-region replay remain valid. Migration locks and disk/WAL headroom remain rollout considerations, not unverified engineering behavior.

Commands (all owning exits 0; roots `.cartulary/test-results/<run-id>`):

- `nf-s12m-generate-20261029-01`: `make generate`; reviewed migration ownership/history projections.
- `nf-s12m-format-20261029-01`: `make format`.
- `nf-s12m-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`.
- `nf-s12m-drift-20261029-01`: `make migration-drift`.
- `nf-s12m-generated-20261029-01`: `make source-check CHECK=generated`.
- `nf-s12m-region-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`.
- `nf-s12m-recovery-20261029-01`: `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3`.

Selections were discovered with `make task-guide OWNER=module.database_migrations`, `make task-guide OWNER=module.imports`, and `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s12m-selection.json DETAIL=rows` (also `-region.json` and `-recovery.json`); all exited 0. Exact source identities, work-unit and cleanup results are retained in each run manifest/summary. No required check failed; no broad CI or deployment was run. Checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12m-checkpoint-markdown-20261029-01 make lint-markdown`.


### S12 execution

S12-M checkpoint passed (exit 0). Separating metadata lookup from bounded content access; ordinary readers own read-only consistent snapshots and apply readers borrow the existing protected transaction. Upload/discovery and Recovery inventory memory remain outside this claim.


S12 exit: PASS. Metadata queries no longer select source bytes. Content readers issue completed substring queries of at most 64 KiB, validate declared length and SHA-256 before EOF, and release owned read-only repeatable-read snapshots on EOF/error/close/cancellation. Borrowed readers take a source SHARE lock, use the supplied transaction connection and never end it; Network Flow apply now uses that path. XLSX region discovery explicitly consumes the reader into its required complete archive and remains outside the bounded analytical claim. Application buffer bounds and PostgreSQL buffer observations are separate evidence.

V12 extends the registered upload/replay integration selector with empty/exact/boundary/multiple-chunk sizes, size corruption, digest mismatch, injected read errors, repeated/early close, idle cancellation, consistent snapshot mutation, single-connection borrowed execution and mutation exclusion after reader close. Database cost observations compare a 64 KiB slice from 512 KiB and 4 MiB uncompressed sources, checked in the selected test. Successful test log lines are not retained by this harness, so no numeric observations are claimed from those logs. A final page-access assertion requires at most 64 shared pages for each fixed slice; the 4 MiB whole value requires over 500 pages. This is a database access bound, not a timing benchmark. No public error envelope or capability identity changed.

Commands (roots `.cartulary/test-results/<run-id>`):

- `nf-s12-format-20261029-01`, `-02`, `-03`: `make format`, each exit 0.
- `nf-s12-boundary-20261029-01`: `make lint CHECK=backend-boundary`, exit 0.
- `nf-s12-readers-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`, exit 2, harness/tool diagnostic failure while compiling a remaining XLSX-region caller of the removed helper; selected product cases did not execute. Repaired that caller to consume the bounded source reader.
- `nf-s12-readers-20261029-02`: same selection plus `module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`, exit 0; 4/4 work units, cleanup 17/17.
- `nf-s12-readers-20261029-03`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation`, exit 0 after adding fault/cost cases and preserving cleanup errors.

Source and graph identities are retained in each manifest. Long snapshots consume one connection each; cancellation and deterministic cleanup bound their lifecycle. Source decoding and generic-row orchestration remain S13/S14 and S12-I obligations, respectively; no end-to-end bounded apply claim is made yet. Checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12-checkpoint-markdown-20261029-01 make lint-markdown`.


S12 checkpoint follow-up, before S12-I: successful-test logs were not retained, so strengthened database-cost evidence to an explicit page-access assertion and reopened S12 for its focused verification.


S12 database-cost follow-up passed: `CARTULARY_TEST_RUN_ID=nf-s12-cost-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2`, owning exit 0, root `.cartulary/test-results/nf-s12-cost-20261029-01`. Both EXTERNAL source sizes satisfied the fixed-slice bound of at most 64 shared pages. S12 is DONE. Renewed checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12-checkpoint-markdown-20261029-02 make lint-markdown`.


### S12-I execution

S12 renewed checkpoint passed (exit 0). Removing job-wide row payloads and splitting generic versus analytical admission/execution. Migration 32 will require the generic row digest only for workbook plans and remove obsolete analytical digests; mixed old/new workers are unsupported.


S12-I exit: PASS; F24 is fixed. Job scheduling loads only ordered unit identifiers. Admission reads mapping/source metadata and hashes generic discovered rows only in the generic branch; analytical admission protects and checks the original source identity. Locked analytical execution never fetches or decodes generic rows, and verifies source capability, mapping fingerprint/digest and admitted target binding. Generic execution retains discovered-row digest admission. Migration 32 clears analytical row digests, enforces target-dependent nullability and preserves every other plan field. Downgrade with analytical plans fails instead of synthesizing obsolete metadata; snapshot rollback remains required.

V19 evidence includes the real Network Flow mapping/admission/apply/replay path with poisoned generic JSON and database SELECT access to that column denied to runtime, generic source/mapping mutation regressions, analytical source/mapping mutation after admission, cancellation and committed-result recovery. Populated migration tests preserve source bytes and all plan metadata except the obsolete analytical digest, retain generic digests, and reject either invalid nullability shape.

Commands (all owning exits 0; roots `.cartulary/test-results/<run-id>`):

- `nf-s12i-generate-20261029-01`: `make generate`.
- `nf-s12i-format-20261029-01`, `-02`: `make format`.
- `nf-s12i-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.uncommitted_unit_cancellation`.
- `nf-s12i-current-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.apply_transaction_current_state_bc8ad0ac32`, after extending the selector with analytical source/mapping mutations.
- `nf-s12i-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`.
- `nf-s12i-drift-20261029-01`: `make migration-drift`.
- `nf-s12i-generated-20261029-01`: `make source-check CHECK=generated`.
- `nf-s12i-schema-20261029-01`: `make source-check CHECK=schema`.

Selection: `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s12i-selection.json DETAIL=rows`, exit 0. Manifests retain exact source/graph identities; all selected work units and cleanup passed. No engineering failure remains for this slice. Operational risk is mixed workers against the changed schema: quiesce workers and take a recoverable snapshot before migration. S13 follows the checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12i-checkpoint-markdown-20261029-01 make lint-markdown`.


### S13 execution

S12-I checkpoint passed (exit 0). Implementing the adopted owner dialect incrementally, with bounded per-cell retained text, exact scalar/hash accounting, field-count precedence, complete raw-cell diagnostics and raw-only preview tail verification. Apply collection removal remains S14.


S13 exit: PASS; F23 is fixed. Replaced whole-source loading and standard-library CSV interpretation with the adopted incremental decoder. It retains bounded cell text, hashes decoded values while discarding excess text, preserves exact scalar counts, counts blank/logical records correctly and accepts CR/LF/CRLF without changing quoted line breaks. Preview stops semantic decoding after record 50 and drains raw bytes for integrity. Field-count mismatch has sole row-level precedence; otherwise mapped/unmapped/ignored oversized cells produce exact, redacted diagnostics, skip dependent target validation and preserve other target checks. Preview admission accepts the adopted column-aware bound and rejects nonconforming limit attribution/privacy. Apply still retains complete record/accepted collections until S14.

V13/V14 cover independently authored dialect cells at byte splits, UTF-8/BOM/framing failures, source and row digest vectors, 50/51 boundaries with malformed/invalid-encoding tails, raw read/hash failures, header/column/row limits, exact scalar counts/hashes after discarded text, mapped/unmapped/ignored diagnostics, field-count precedence and a complete 1,500-diagnostic preview. Existing diagnostics-retention and apply/error regressions pass.

Commands (all owning exits 0; roots `.cartulary/test-results/<run-id>`):

- `nf-s13-format-20261029-01`, `-02`, `-03`: `make format`.
- `nf-s13-parser-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.query_authoring_admission`.
- `nf-s13-parser-20261029-02`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a`, after adding full-preview bound and valid ignored-column cases.
- `nf-s13-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`.

All selected work units and cleanup passed; source identities remain in the manifests. Existing tables are unchanged. Corrected interpretation applies to new original-source imports and does not reinterpret stored tables. No remaining S13 blocker; checkpoint before S14: `CARTULARY_TEST_RUN_ID=nf-s13-checkpoint-markdown-20261029-01 make lint-markdown`.


### S14 execution

S13 checkpoint passed (exit 0). Replacing complete apply collections with bounded summary validation followed by row-at-a-time insertion through the borrowed source transaction. No table is allocated until validation succeeds with accepted rows; Imports retains rollback/commit/cancellation ownership.


S14 exit: PASS; F21 is fixed. First-pass validation retains counts, source and
mapping identities, bounded descriptors, the configured diagnostic prefix and
at most 50 all-rejected diagnostics. Accepted-row limits are enforced before
allocating a table. The second pass inserts one accepted row at a time in source
order on the borrowed transaction, then compares source, descriptor, count and
full validation-summary digests. Production apply no longer collects parsed or
accepted rows; materialized apply helpers exist only in test fixtures. Imports
passes the complete locked source capability and retains final commit,
cancellation and committed-result recovery ownership.

V15 covers increasing 10/1,000/10,000-row inputs with fixed retained diagnostic
bounds and 64 KiB reads, retention 0/N/N+1, bounded all-rejected sampling,
accepted limits, multiple diagnostics per rejected row, second-pass read/digest/
insertion/cancellation failures and summary disagreement. The real Imports path
still runs with generic-row SELECT access denied. A database trigger proves
failure on the second insert rolls back the first row and table; a later
outcome-write failure rolls back table, rows, diagnostics and audit effects.
This establishes bounded application structures and query widths, not a measured
heap/RSS ceiling. Database slice-page assertions remain separate S12 evidence.

Commands (roots `.cartulary/test-results/<run-id>`; explicit
`CARTULARY_TEST_RUN_ID` for each):

- `nf-s14-format-20261029-01` through `-05`: `make format`, each exit 0.
- `nf-s14-generate-20261029-01`: `make generate`, exit 0; registered the new
  streaming validation row through the authored Network Flow family.
- `nf-s14-initial-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`,
  owning exit 2, product/test_assertion_failure, 2/3 work units passed, cleanup
  14/14. Full-capability validation exposed Imports' previously fabricated
  partial capability; repaired the facade to pass its already-validated locked
  capability. Source digest `sha256:9adbba202eac4a8895a232cb5dac47f16f5ad0a7794d1cdc58e18c3b98996aae`.
- `nf-s14-streaming-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232`,
  exit 0, 4/4 work units, cleanup 15/15.
- `nf-s14-imports-20261029-02`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6`,
  exit 0, 5/5 work units, cleanup 20/20. These two repaired runs share source
  digest `sha256:cd27701e28ea6cd81766884aa9b38896be8e11d3f93ac3d86e2e391e67c64155`.
- `nf-s14-atomic-20261029-03`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`,
  exit 0, 3/3 work units, cleanup 14/14 after extending late-failure assertions;
  source digest `sha256:46a34642d19edb452c056d509625f9e0354175759ac19705edba63b15cdb3b6e`.
- `nf-s14-go-20261029-01`: `make lint CHECK=go`, exit 0.
- `nf-s14-boundary-20261029-01`: `make lint CHECK=backend-boundary`, exit 0.
- `nf-s14-regression-20261029-02`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.import_preview_admission_f0dc5f2d8b`,
  exit 0, 1/1 work unit, cleanup 10/10. Final review replaced zero-time sentinel
  checks with explicit operand validation, preserving cross-field validation
  for valid year-0001 timestamps. Source digest
  `sha256:a0feefbfcdd484f8b48b31e25dd53d9b2ffaeaee4118d5423feaf9be1c57aee7`.

`make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s14-selection.json DETAIL=rows`
and `make explain-run` for the five named test roots exited 0. Source commit is
`b6d073974802489edfeb892d39f48b128b9d00c4`; exact graph/toolchain identities remain
in each manifest. No cleanup failures, skipped or cancelled units remain in the
passing runs. Longer transactions and two source passes are intentional
operational tradeoffs; pool lifecycle and cancellation are covered. S15 will
finalize and qualify the final source. Checkpoint before S15:
`CARTULARY_TEST_RUN_ID=nf-s14-checkpoint-markdown-20261029-01 make lint-markdown`.


### S15 execution

S14 checkpoint Markdown passed (owning exit 0). Finalizing generated artifacts,
reviewing the complete change and then freezing product source for the focused
acceptance portfolio. Prior slice evidence remains historical development
qualification; final source identity and selected acceptance roots will be
recorded below. Full CI and deployment remain excluded. No compatible full
warm-check root exists, so retained-run maintenance is omitted from
`agent-finalize`; no redundant warm check is launched to manufacture one.


S15 source freeze: `nf-s15-format-20261029-01` / `make format` and
`nf-s15-finalize-20261029-01` / `make agent-finalize` passed, owning exits 0.
Finalizer summary `unit-artifacts/finalize-summary.json` reports zero updated
files, successful schema/catalog/tier validation and generated refresh. Reviewed
migration projections, active major-8 Go/browser publication, removed obsolete
schemas and expected generated AJV identifier renumbering. `git diff --check`
passes. Product source is frozen at commit
`b6d073974802489edfeb892d39f48b128b9d00c4`, dirty source digest
`sha256:e7ae7590b06382933642898e7e906bcc0c893683b2e15d87ae20f58d956188fb`.
Toolchain digest is
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
environment/capability digest is
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Only this non-executable tracker changes after freeze unless a documented repair
requires a new finalization and evidence decision.


### S15 final acceptance evidence

Current `make task-guide OWNER=module.networkflow`, `OWNER=module.imports`,
`OWNER=module.database_migrations`, `OWNER=module.recovery` and
`OWNER=web.networkflow` all exited 0. Exact selections were planned with
`make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s15-<selection>.json DETAIL=rows`
for selections `nf`, `imports`, `migration`, `recovery` and `web`, each exit 0.
The commands below preserve the selected rows directly so the temporary planning
files are not required for reproduction. Every invocation uses its named
`CARTULARY_TEST_RUN_ID`, repository-root Make and default result parent
`.cartulary/test-results`. Every owning command exited 0.

- `nf-s15-nf-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8`; PASS, 6/6 work units and 21/21 cleanup steps.
- `nf-s15-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`; PASS, 5/5 work units and 20/20 cleanup steps.
- `nf-s15-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`; PASS, 7/7 work units and 18/18 cleanup steps.
- `nf-s15-recovery-20261029-01`: `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3`; PASS, 3/3 work units and 14/14 cleanup steps.
- `nf-s15-web-20261029-01`: `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.networkflowclient_requires_a_claimed_profile_and_65c861c0d2,web.networkflow.regression.saved_graph_major5_decoder_and_lifecycle`; PASS, 2/2 work units and 11/11 cleanup steps.

| Run ID | Exact Make command | Passed work units | Cleanup steps |
| --- | --- | --- | --- |
| `nf-s15-drift-20261029-01` | `make migration-drift` | 8/8 | 19/19 |
| `nf-s15-schema-20261029-01` | `make source-check CHECK=schema` | 3/3 | 12/12 |
| `nf-s15-generated-20261029-01` | `make source-check CHECK=generated` | 5/5 | 14/14 |
| `nf-s15-go-20261029-01` | `make lint CHECK=go` | 6/6 | 15/15 |
| `nf-s15-types-20261029-01` | `make lint CHECK=types` | 2/2 | 11/11 |
| `nf-s15-boundary-20261029-01` | `make lint CHECK=backend-boundary` | 4/4 | 13/13 |

`make explain-run RESULTS_DIR=.cartulary/test-results/<run-id>` exited 0 for
all eleven acceptance roots. All report the exact frozen source, toolchain and
environment identities above, with zero failed, skipped or cancelled work units
and zero failed or blocked cleanup steps. Per-command graph digests, selected
rows, timestamps and fixture profiles are retained in each `run-manifest.json`;
`run-summary.json` and `cleanup-results.json` preserve the results. Work-unit
counts are not test counts. No S15 repair or source drift occurred. These are
focused obligation passes, not complete-owner or full-CI acceptance claims.

### Finding and gate reconciliation

| Finding / gate | Final decision and evidence |
| --- | --- |
| F18 Obsolete contracts | CLOSED, S10/V10/V16: authored and packaged v7/v1 artifacts and exclusive schemas removed after consumer closure; current v8/v2 Go/browser resolution, schema, generation and TypeScript acceptance pass. Historical migration ledger remains intact. |
| F19 Forwarding helper | CLOSED, S10/V10: production helper removed; direct strict decoder, public preview admission and safe Imports translation pass. |
| F20 Whole-source reads | CLOSED, S12-M/S12/V12: byte-preserving EXTERNAL migration, metadata-only access, 64 KiB integrity-checked source readers, owned/borrowed lifecycle, cancellation, source protection and bounded database page access pass. |
| F21 Materialized apply | CLOSED, S14/V15: bounded first-pass summaries and second-pass source-order insertion; accepted limits, identity agreement, late-failure rollback and real Imports cancellation/recovery pass. |
| F22 Whole-table row lookup | CLOSED, S11/V11: scoped incident/table/row retrieval, isolation, active-table and deletion-race checks pass; graph streaming retained. |
| F23 Divergent CSV/limits | CLOSED, S13/S14/V13/V14: shared incremental dialect, exact decoded hashes/counts, raw-cell privacy/precedence, complete bounded previews, explicit cross-field operand validity and 50/51 semantic boundary pass. |
| F24 Generic analytical rows | CLOSED, S12-I/V19: metadata-only scheduling and analytical admission/apply; real PostgreSQL generic-column access denial, generic regressions, nullable-plan migration and identity/replay checks pass. |
| RB-004 Owner adoption | CLOSED by the user's explicit S09 adoption; coordinated owner headers record the human decision. Tests do not supply specification authority. |
| RB-005 Characterization and implementation evidence | CLOSED by S09 characterization plus the implementing slices and compatible frozen-source acceptance above. |
| V17 Final focused acceptance | PASS: finalizer reviewed, product source frozen, exact required portfolio passed without repairs or cleanup failures. |
| V18 Tracker/handoff | Completion text is checked by the final Markdown commands recorded below. |

S09, S10, S11, S12-M, S12, S12-I, S13 and S14 are DONE. Their individual
checkpoint results precede their successors. No engineering remediation blocker
remains; S15 closes after the final documentation checkpoint. The historical
S00–S08 and DOC-02 evidence, including prior failures and obsolete authorization
postures, remain preserved and explicitly superseded by this execution record.

### Operational handoff and acceptance limits

1. Before applying migrations 31/32, quiesce import workers and prevent new apply
   scheduling. Take and verify a recoverable database/environment snapshot.
   Estimate source-storage expansion, temporary rewrite space, WAL volume and
   lock duration from the deployment's source inventory; reserve headroom and
   a maintenance window. Migration 31 changes storage policy and rewrites
   compressed values, preserving source references, logical bytes and digests.
2. Apply the authored migrations with the repository's normal migration owner.
   Migration 32 removes only obsolete analytical plan row digests and enforces
   target-specific source authority; mapping snapshots, source identities,
   outcomes and replay identifiers survive. Prohibit mixed old/new workers and
   old binaries against the migrated admission-plan schema. Both empty and
   populated migration paths and migration drift passed engineering validation.
3. Publish matching server/browser assets at the existing **major-8/state-5**
   boundary, then perform the separately authorized operational acceptance before
   resuming workers. Internal storage changes do not create another public major
   or Network Flow state version. Existing Network Flow tables are neither
   reinterpreted nor deleted. To obtain corrected CSV interpretation, re-import
   original source bytes with the approved mapping decisions.
4. Observe import failures, cancellation/connection cleanup, transaction duration
   and database storage/WAL use after deployment. Apply deliberately performs two
   source passes and row-at-a-time inserts; larger inputs can hold the transaction
   longer. Resource evidence establishes fixed query widths and retained state
   bounded by configured cell/column/diagnostic limits, including analytical
   orchestration without generic rows. It does not claim a measured heap/RSS
   maximum, timing improvement, bounded upload/discovery, or Recovery-wide memory.
5. If rollback is needed, return to the previous matching application/environment
   and database snapshot, including its matching assets, object store and keys.
   Do not downgrade a database after new writes, synthesize analytical generic
   metadata, or delete metadata/data to force admission. Migration 32 intentionally
   cannot reverse populated analytical plans without the matching snapshot.

No deployment, destructive reset or broad historical conversion occurred during
these slices. Operational deployment is a separate, unperformed action; this
handoff is its prerequisite, not its authorization. Full CI, visual suites,
release benchmarks and complete-owner audits were not selected because the
approved plan requires focused acceptance and no presentation behavior changed.
No required focused check is waived. Finalizer retained-run maintenance was
intentionally skipped because no compatible full warm-check root was supplied.

Final documentation checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s15-checkpoint-markdown-20261029-01 make lint-markdown`.
Its root is `.cartulary/test-results/nf-s15-checkpoint-markdown-20261029-01`,
with summary `adhoc/lint-markdown/tool-run-summary.json`.


S15 exit: **DONE**. The final documentation checkpoint passed with owning exit 0
and zero failures. V18 and the operational handoff are complete; every preceding
workstream and finding is reconciled. No required engineering check remains.
All edits are left uncommitted for review, preserving the original tracker
changes and completed history. This completion-status edit receives the final
confirmation `CARTULARY_TEST_RUN_ID=nf-s15-completion-markdown-20261029-01 make lint-markdown`;
its retained summary is
`.cartulary/test-results/nf-s15-completion-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json`.
No product source changed after finalization or acceptance.

## 22. IMP-04 controlling remediation and execution

### Findings, decisions, and acceptance corrections

The user authorized implementation of this plan on 2026-10-29. The working tree
already contains IMP-03 source, specification, generated, test, and migration
changes, including untracked migrations 31/32 and bounded-reader/parser files.
They are preserved. The implementation baseline remains `main` at
`b6d073974802489edfeb892d39f48b128b9d00c4`. No IMP-04 check has run at this initial
checkpoint. The review was read-only and its retained S15 passes do not prove
the missing obligations.

| Finding | Remediation and areas | Rationale / long-term benefit | Compatibility and unresolved risk | Required validation |
| --- | --- | --- | --- | --- |
| F25 Generic discovery blocks owner-valid analytical input | Admit CSV bytes and a source-only unit; defer Workbook projection until generic preview/approval; add owner source inspection and owner-derived analytical descriptors. Specification, implementation, storage migration, contracts, browser, tests, docs. | Imports owns custody/lifecycle; each target owns interpretation. New targets no longer inherit Workbook dialect/limits. | Nullable tabular metadata and coordinated API/browser publication; preserve source, approval, outcome and replay identities. Without correction valid sources cannot reach the owner. | V20 public upload/inspection/preview/approval/apply, literal quotes, 50/51 boundary, default 100001 rows, CSV byte-only limits, generic failure isolation, authorization/integrity/cancellation/replay. |
| F26 Unit failures admitted as row diagnostics | Limit row resource diagnostics to cell_scalar_limit_exceeded at shared preview/retained-state admission. Implementation, tests, affected projections, docs. | One semantic boundary preserves failure scope and privacy. | Tighten already-invalid payloads without altering valid envelopes or rewriting data. Otherwise forged consistent unit failures can appear as row rejection. | V21 negative admission for header/column/row/accepted-row limits; positive mapped/unmapped/ignored cell diagnostics and safe unit translation. |
| F27 Contradictory transaction ownership | Reconcile Network Flow §10.1 with §9.2 and Core: Imports owns transaction lifecycle/final commit; owner writes within borrowed transaction. Specification/docs. | Single commit authority makes atomicity and recovery enforceable. | No intended runtime/storage change. Otherwise future owners can split commits or publish early. | V22 human normative consistency review plus borrowed-transaction/atomicity regression; no executable Markdown dependency. |
| F28 Migrated execution and replay unproved | Valid version-30 generic/analytical plans, migrate same databases through current head, attach normal runtime, execute/cancel/recover/replay. Tests, harness, routing, docs; repair exposed defects. | Proves workflow survival instead of only SQL shape preservation. | Preserve durable identities; explicit reviewed registry translation for equivalent apply bindings only. Otherwise pending plans can strand or duplicate effects after upgrade. | V23 queued/interrupted execution, precommit cancellation, before/after-upgrade committed outcomes, postcommit publication crash, replay and mismatch rejection for both target families. |

CSV upload performs envelope/auth/byte classification/byte ceiling/hash/source
admission only. Discovery creates one csv_file unit without interpreting CSV.
Its rectangle, header/data row references and inferred counts form an all-null
or materialized group; unknown values are never fabricated. Generic preview is
read-only; generic approval atomically materializes its projection and mapping.
Generic apply retains row-digest protection. Imports row/column/cell/archive
limits remain XLSX-specific. Existing generic projections remain usable.

The registered analytical source-inspection route accepts target_kind and
extension_profile_id, derives authorized source identity, and dispatches to its
owner. Network Flow parses only its header using the owner decoder and drains
remaining raw bytes for integrity, returning canonical descriptors/digest with
no mapping requirement or data-validation claim. Mapping preview/approval and
the browser use owner-derived descriptors, never generic rows/columns.

Add an Imports migration for absent tabular projections and removal of obsolete
analytical plan rectangles. Preserve original bytes, generic projections,
approvals, jobs, outcomes and replay identifiers. Registry changes require an
explicit old-to-new migration only for reviewed apply-equivalent bindings;
unknown/mismatched bindings fail closed, with no runtime legacy reader.

### Workstreams and authoritative dependencies

| Workstream | Status | Depends on | Scope / principal risk | Binary exit |
| --- | --- | --- | --- | --- |
| S16 Reopen acceptance and close specifications | DONE | Initial IMP-04 tracker checkpoint | Correct F27; prepare coordinated source-admission amendments and map obligations. Risk: treating proposals as adopted or erasing evidence. | F27 reconciled, new amendments explicitly adopted, decisions/validation mapped, Markdown checkpoint passes. |
| S17 Source admission and target-owned interpretation | DONE | S16 | Nullable unit/projection migration, byte-only CSV admission, lazy generic projection, owner inspection and mapping admission, authored contracts/generation. Risk: weakened admission, projection races, registry drift. | Focused backend, migration and generated checks pass; analytical content remains independent of generic projection. |
| S18 Browser and public workflow integration | DONE | S17 | Owner inspection for mapping, honest unavailable metadata, stable source identity, public pipeline/browser tests. Risk: generic fallback or stale descriptors. | V20 passes including real 100001-row default-limit apply; browser no longer requests generic preview for Network Flow. |
| S19 Diagnostic admission | DONE | S18 | Shared row-level reason restriction and adversarial tests. Risk: rejecting valid cell diagnostics or changing safe unit translation. | V21 passes at preview and retained-state boundaries. |
| S20 Migrated workflow execution | DONE | S19 | Scratch-runtime harness capability, valid legacy fixtures, execution/cancellation/recovery/replay and exposed repairs. Risk: invalid fixtures, binding drift, duplicated effects or cleanup leaks. | V23 and reopened V19 pass for both targets; migration drift and affected harness checks pass. |
| S15-R Final validation and handoff | DONE | S20 and all reopened obligations | Finalize/review/freeze, compatible focused portfolio, tracker reconciliation and operations handoff. Risk: stale evidence or unclosed obligations. | All findings/exits reconciled, required checks pass, handoff complete, no engineering blocker. |

Every successor waits for its predecessor's validation and tracker checkpoint.
Record exact commands/selections, source identity, owning exits, run roots,
cleanup, failures/repairs, skipped checks/reasons, remaining risks and exit
decision. A failed required check leaves the workstream incomplete.

### Validation and handoff obligations

V20 covers literal quotes, malformed framing/UTF-8 within versus beyond preview
record 50, full-apply rollback, a real 100001-row valid CSV under default limits,
CSV independence from generic row/column/cell limits, owner limits, byte-overflow
no-side-effects, generic-failure isolation, XLSX/generic regressions, source and
mapping integrity, authorization, cancellation and idempotency. Register the
large fixture with a declared resource budget; scaled examples do not replace it.

V21 uses otherwise-valid forged diagnostics for all four unit-only limit reasons
and positive privacy/attribution/hash cases. V22 is a human owner-text audit plus
existing executable transaction tests, never a Markdown-based product check.
V23 uses genuine legacy mappings and frozen old registry/binding identities at
schema 30, migrates the same database through 31/32/new head, and executes real
workers/recovery. Cover outcomes committed before upgrade and after migration,
precommit cancellation, publication interruption, repeated replay, source/mapping/
generic-row/binding mismatches, and analytical generic-column access denial.
Retain physical byte/compression/region-copy/Recovery migration coverage.

Discover current guidance with make task-guide and exact selections with
make target-plan; register authored verification inputs and regenerate via Make.
Run focused test-slice, schema/generated, Go/TypeScript, backend/frontend boundary,
migration-drift, affected browser and harness checks. Run agent-finalize, review
changes, freeze source, then final acceptance; recompute compatibility after
repairs. Run lint-markdown at tracker checkpoints. Full CI is excluded by default.
Tests, generators, runtime metadata and release evidence never consume Markdown.

Operational handoff requires quiesced workers, a verified recoverable snapshot,
migration preflight and matching server/browser publication; mixed workers are
unsupported. Existing Network Flow tables remain untouched. Corrected results
require original-source re-import. Upload buffering, generic materialization and
Recovery-wide memory are outside the bounded analytical claim. Rollback restores
the matching application/database environment snapshot, never downgrade or
metadata deletion after writes. Commit/deployment are separate unperformed actions.

### Execution log

| Checkpoint | Changes / evidence | Exit / next action |
| --- | --- | --- |
| IMP-04 initial tracker edit | Preserved IMP-03 working-tree changes and history; reopened affected completion claims; registered F25–F28, V20–V23 and S16–S20/S15-R. | S16 IN_PROGRESS; prepare concrete owner amendments before adoption request. |
| S16 owner candidate preparation | Core 00 §1.4 records proposed adoption scope; Core 01 §17.2 and Core 03 §11 define byte-only CSV source admission, nullable metadata, lazy generic projection and target-owned inspection/approval; Network Flow §10.1 defines the closed header-only inspection operation and corrects final commit ownership; Extensions and Core 04 are coordinated. Existing §1.3 adoption remains intact. | F27 text corrected; source-admission candidate awaits explicit human adoption. No dependent implementation begun. |
| S16 command discovery and initial document check | `make task-guide OWNER=module.imports` and `make task-guide OWNER=module.networkflow` exited 0. `CARTULARY_TEST_RUN_ID=nf-s16-candidate-markdown-20261029-01 make lint-markdown` exited 0; root `.cartulary/test-results/nf-s16-candidate-markdown-20261029-01`, summary `adhoc/lint-markdown/tool-run-summary.json`, zero failures. | A final candidate checkpoint follows the small owner-scope clarification and this log entry. Product tests, generation and finalization are intentionally deferred until the implementation workstreams. |
| S16 coordinated review checkpoint | `CARTULARY_TEST_RUN_ID=nf-s16-review-markdown-20261029-01 make lint-markdown` exited 0, zero failures; root `.cartulary/test-results/nf-s16-review-markdown-20261029-01`, summary `adhoc/lint-markdown/tool-run-summary.json`. Human-oriented source review reconciled §10.1 with §9.2 and Core transaction ownership; analytical versus generic descriptor authority is explicit. Source inspection samples are explicitly empty because no data records are interpreted. | Candidate ready for explicit owner adoption; S16 remains IN_PROGRESS and S17 has not begun. |

S16 migration baseline: the existing generated import-target registry digest is
`6ed462c191ff86e499efefc153e222f2bc0b7ee7098c6e018443bfcdacb5937b`.
This is an observed executable identity for later legacy fixtures/migration
review, not a document-derived runtime input. Capture legacy machine fixtures
from the authored/generated contract baseline, never from this Markdown record.
No product source, test, migration or generated output has been changed by
IMP-04/S16. Required product checks are deferred to their dependent slices;
agent-finalize is deferred to S15-R. The final prepared-document checkpoint is
`CARTULARY_TEST_RUN_ID=nf-s16-adoption-candidate-markdown-20261029-01 make lint-markdown`.

That prepared-candidate check passed with owning exit 0 and zero failures; root
`.cartulary/test-results/nf-s16-adoption-candidate-markdown-20261029-01`, summary
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check -- docs` also
exited 0. The explicit owner-adoption question has been presented for Core 00
§1.4; no answer has yet been recorded. S16 cannot exit and S17 cannot begin
until that required decision arrives. The user already adopted the distinct
bounded-import candidate in §1.3; that decision remains in force. This evidence
and gate-status log receives
`CARTULARY_TEST_RUN_ID=nf-s16-pending-adoption-markdown-20261029-01 make lint-markdown`.


### S16 adoption and exit

On 2026-10-29 the user explicitly answered "Adopt the prepared candidate and
continue implementation" to the Core 00 §1.4 adoption request, including the
nullable source-unit metadata, owner source inspection, lazy Workbook
interpretation, coordinated browser/API changes and migration rules. Core 00,
Network Flow, Extensions and Core 03 now record the adopted posture. This closes
the owner gate; no further adoption request is required for these decisions.
The pending-adoption Markdown checkpoint above passed with owning exit 0 and
zero failures. F27 is reconciled; transaction regressions remain required in
final acceptance. S16 final adoption checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s16-adopted-markdown-20261029-01 make lint-markdown`.
No product code or generated input changed during S16.


S16 exit: **DONE**. The adopted-document checkpoint passed with owning exit 0,
zero failures, root `.cartulary/test-results/nf-s16-adopted-markdown-20261029-01`,
summary `adhoc/lint-markdown/tool-run-summary.json`. Explicit owner adoption,
F27 correction, interface/migration decisions and validation mappings are
complete. Product checks remain allocated to implementation/final acceptance;
no passing product evidence is claimed for S16. Before S17, this status checkpoint
uses `CARTULARY_TEST_RUN_ID=nf-s16-exit-markdown-20261029-01 make lint-markdown`.


S16 status checkpoint passed (owning exit 0, zero failures) at
`.cartulary/test-results/nf-s16-exit-markdown-20261029-01`. S17 begins with
authored contracts, original-source admission and nullable generic projections.

### S17 implementation and focused qualification (in progress)

Source identity remains HEAD `b6d073974802489edfeb892d39f48b128b9d00c4`
plus the preserved IMP-03 changes and current IMP-04 working tree. CSV admission
now creates a source-only unit; generic preview derives its projection without
persisting it and generic approval persists it atomically. Analytical inspection
and mapping use the owner decoder and source capability. Migration 33 introduces
the nullable projection group, removes analytical plan rectangles, and translates
only the known old registry and reviewed selector/binding pairs. Machine contract
comparison found no changes to mapping/create/apply/error/commit contracts;
unknown identities remain subject to normal fail-closed admission. Browser work
and migrated execution proof remain assigned to S18 and S20.

Execution evidence so far (all roots under `.cartulary/test-results/`):

| Exact command | Run ID / owning exit | Result / repair |
| --- | --- | --- |
| `make generate` with explicit `CARTULARY_TEST_RUN_ID` | `nf-s17-generate-20261029-01`, exit 2; tool exit 1 | Failed: release ledger covered 439/449 changes. Added the ten adopted nullable-metadata/inspection changes to the authored 2.0.0 change set. |
| Same command | `nf-s17-generate-20261029-02`, exit 2; tool exit 1 | Failed: duplicated source-column schema ID in inspection helper. Removed the helper's duplicate public identity. |
| Same command | `nf-s17-generate-20261029-03`, exit 2; tool exit 1 | Failed: new public inspection schemas missing from owner index. Registered both authored schema IDs. |
| Same command | `nf-s17-generate-20261029-04`, exit 0 | Passed, zero failures. |
| Same command | `nf-s17-generate-20261029-05`, exit 2; tool exit 1 | Failed: newly registered test rows were not ASCII sorted. Sorted authored row registries. |
| Same command | `nf-s17-generate-20261029-06`, exit 0 | Passed, zero failures, including updated case routing. |
| `make target-plan TARGET=test-slice DETAIL=rows OWNER=module.networkflow` | exit 2 | Selection usage error: OWNER belongs in INPUTS. No product execution. |
| `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s17-plan.json DETAIL=rows` | exit 0 | Corrected selection: Network Flow preview admission row only. |
| `make format` with explicit run ID | `nf-s17-format-20261029-01`, exit 0 | Passed. |
| Same command | `nf-s17-format-20261029-02`, exit 2 | Configuration failure from unsorted authored rows above; no product result. |

Generator summaries are `<run>/generate/tool-run-summary.json`; failures retain
their original harness/unknown_failure classification and concrete causes above.
Formatting summaries are graph-owned. These runs allocate no product services.
No product pass or S17 exit is claimed yet; focused backend, migration, schema
and generated validation remain required before S18 begins.

S17 follow-up evidence and repairs:

- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s17-imports-selection.json DETAIL=rows`,
  `/tmp/nf-s17-nf-selection.json`, and `/tmp/nf-s17-migration-selection.json`
  each exited 0. Selections are the seven Imports rows listed below, the two
  Network Flow rows, and migration schema bootstrap respectively.
- `CARTULARY_TEST_RUN_ID=nf-s17-format-20261029-03 make format` passed, exit 0.
- `CARTULARY_TEST_RUN_ID=nf-s17-networkflow-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.import_preview_admission_f0dc5f2d8b`
  passed, exit 0. New inspection vectors cover malformed/invalid-encoding data
  tails, source digest mismatch, invalid headers and strict response admission.
- `CARTULARY_TEST_RUN_ID=nf-s17-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.unit.behavior_contract,module.imports.unit.module_composition_contract,module.imports.unit.target_neutral_csv_source_admission,module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`
  passed, exit 0. This covers source-only metadata, read-only generic preview,
  lazy generic approval/apply, analytical inspection/approval/apply under denied
  generic-row access, read-role rejection, replay and XLSX operator regions.
- `CARTULARY_TEST_RUN_ID=nf-s17-migrations-20261029-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`
  and `CARTULARY_TEST_RUN_ID=nf-s17-migration-drift-20261029-01 make migration-drift`
  passed, owning exits 0. Preserved bytes/compression/region copies and plan
  metadata tests now include analytical rectangle removal and grouped-null
  projection constraints. These are storage proofs, not S20 execution proof.
- The four runs above used source digest
  `sha256:6a7065863dd3314e0d7deace15f486c550fcb878d4b20c8f1d75b8563c3ceace`.
  Their cleanup records show respectively 10, 18, 18 and 19 completed steps,
  no failures. Roots are `.cartulary/test-results/<run-id>`; graph acceptance and
  cleanup are in `run-summary.json` and `cleanup-results.json`.
- `make source-check CHECK=schema`, run `nf-s17-schema-20261029-01`, failed
  with owning exit 2 / graph product test_assertion_failure: the authored route
  shape validator lacked source inspection. Updated its exact keys/operation
  model. Run `nf-s17-schema-20261029-02` then failed on stale topology metadata;
  `make source-check CHECK=generated`, run `nf-s17-generated-20261029-01`,
  independently detected that same stale generated index. All cleanup completed.
- `CARTULARY_TEST_RUN_ID=nf-s17-generate-20261029-07 make generate` passed,
  exit 0. Subsequent schema run `nf-s17-schema-20261029-03` and generated run
  `nf-s17-generated-20261029-02` passed, exits 0, cleanup 12/14 completed steps.
  Both used source digest
  `sha256:6b1e8bff6c4e58b4894b229ff5569c78074ec276b19529cf3ba00fb0358c84c0`.
- `make lint CHECK=backend-boundary`, run `nf-s17-boundary-20261029-01`,
  exited 2 at runtime planning while tooling regeneration was in progress;
  no product execution or acceptance is claimed. After regeneration, run
  `nf-s17-boundary-20261029-02` passed, exit 0, cleanup 13 completed steps.
- `make lint CHECK=go`, run `nf-s17-go-20261029-01`, failed, exit 2:
  staticcheck S1016 required the direct private selector conversion. Repaired
  it; run `nf-s17-go-20261029-02` passed, exit 0, cleanup 15 completed steps.
  The repaired boundary/Go runs used source digest
  `sha256:8e9795fd214cce6562820a6ba00bb20994bdf8d9b534c7082194a4240ffd9f41`.
- Two diagnosis-only invocations were malformed: explain-run DETAIL=failures
  is unsupported, and DETAIL=logs requires TARGET; both exited 2 and executed
  no tests. Corrected bounded diagnosis used TARGET plus DETAIL=summary/logs.
- `CARTULARY_TEST_RUN_ID=nf-s17-markdown-progress-20261029-01 make lint-markdown`
  passed, exit 0. `git diff --check` passed.

Source capability lookup additionally rejects streams attached to a different
session/unit. The affected seven-row Imports selection is being requalified in
`nf-s17-imports-20261029-02` with the same exact command/ROWS as above. Historical
checks retain their actual scope; final-source compatibility is reassessed in
S15-R. Full CI, browser acceptance and migration execution/recovery remain
intentionally deferred to their assigned slices. S17 exit is still pending the
requalification and final Markdown checkpoint.


### S17 exit checkpoint

The repeated seven-row Imports selection passed with owning exit 0 at
`.cartulary/test-results/nf-s17-imports-20261029-02`, source digest
`sha256:8e9795fd214cce6562820a6ba00bb20994bdf8d9b534c7082194a4240ffd9f41`,
all 18 cleanup steps completed. S17 implementation, focused backend checks,
source/schema regeneration, Go/boundary validation and migration checks are
complete. **S17 exit: DONE**, subject to the Markdown checkpoint below before
S18 starts. F25 end-to-end closure remains open for S18; F28 remains open for
S20. Matching browser publication, known-registry translation execution,
large-input public scenarios and final evidence compatibility remain explicit
later obligations. No commit or deployment occurred.

Checkpoint command:
`CARTULARY_TEST_RUN_ID=nf-s17-exit-markdown-20261029-01 make lint-markdown`.

S17 Markdown exit checkpoint passed, owning exit 0 and zero failures. S18 now
begins; S19/S20/S15-R have not begun.

### S18 implementation and qualification (in progress)

Network Flow's browser controller and mapping model now consume the registered
owner source-inspection result. Approval echoes owner descriptors with the
owner's fixed row references. Generic Workbook preview accepts initially absent
metadata and supplies its own projection references on approval. Unit source
identity excludes derived rectangles. The Import Assistant displays unavailable
tabular metadata explicitly. Browser fixtures use source-only units, and the
workspace workflow asserts inspection is called and generic preview is not.

Two public backend rows are registered: target-neutral source admission exercises
literal quotes, early/late malformed framing and encoding, generic-preview
failure followed by analytical success, generic row/column/cell settings,
owner column limits, byte overflow, atomic apply failure and request replay.
The separate default-limit row imports exactly 100001 records under default
owner/import limits. Its declared `backend_capacity_isolated` profile reserves
1024 MiB, one CPU/IO/process, plus required service minima; its test bounds apply
observation at two minutes. This profile is an admission reservation, not a
measurement of application or database resource use.

- `make task-guide OWNER=web.networkflow` passed, exit 0.
- `CARTULARY_TEST_RUN_ID=nf-s18-generate-20261029-01 make generate` passed,
  exit 0; summary `generate/tool-run-summary.json` at the explicit run root.
- `CARTULARY_TEST_RUN_ID=nf-s18-types-20261029-01 make lint CHECK=types`
  failed, exit 2. Causes were a duplicated validator argument and test fixtures
  still shaped as generic previews/non-null units. Repaired the authored
  entrypoint and fixtures; `nf-s18-types-20261029-02` passed, exit 0.
- `CARTULARY_TEST_RUN_ID=nf-s18-format-20261029-01 make format` failed,
  exit 2, because the inspection local lacked an explicit type. Added the
  owner-decoder return type; `nf-s18-format-20261029-02` passed, exit 0.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s18-web.networkflow-selection.json DETAIL=rows`,
  `/tmp/nf-s18-module.imports-selection.json`, and `/tmp/nf-s18-public-selection.json`
  all passed, exit 0. The first selects four Network Flow browser/controller
  rows; the second selects seven Imports transport/lifecycle/presentation rows;
  the third selects the two new public backend rows.

These exact slices are running at `nf-s18-web-networkflow-20261029-01`,
`nf-s18-web-imports-20261029-01`, and `nf-s18-public-20261029-01` under
`.cartulary/test-results/`. Their eventual owning exits, source identities,
cleanup, failures and repairs must be reconciled before S18 exits. S19 has not
begun. No acceptance claim is made for pending runs.

S18 public and browser-unit results:

- `CARTULARY_TEST_RUN_ID=nf-s18-public-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.target_neutral_public_source_admission,module.imports.integration.network_flow_default_limit_100001_rows`
  passed, owning exit 0, all 17 cleanup steps completed. The default-limit case
  successfully committed exactly 100001 rows through public upload, inspection,
  mapping preview, approval, selection and apply. Late framing/encoding failures
  permitted preview but atomically failed apply; repeated requests retained the
  same job. Generic preview failure did not block owner interpretation.
- `CARTULARY_TEST_RUN_ID=nf-s18-web-networkflow-20261029-01 make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a`
  passed, exit 0, all 12 cleanup steps completed.
- `CARTULARY_TEST_RUN_ID=nf-s18-web-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle,module.imports.frontend.operator_region_assistant_039d399324,module.imports.frontend.operator_region_coordinator_eb1902fb6e,module.imports.frontend.presentation_recovery,module.imports.frontend.sanitize_import_transport_errors_through_the_sha_1cdfb222a8,module.imports.frontend.skip_reselect_mapping_retention_522ba6591e,module.imports.frontend.typed_transport`
  failed, owning exit 2 / product test_assertion_failure, with all 12 cleanup
  steps completed. Only lifecycle failed: its overlap fixture used two CSV
  units. Corrected the fixture to two overlapping XLSX units and kept the
  source-change/terminal-session assertions. The other six selected rows passed.
- These three runs used source digest
  `sha256:95ebb5cec161068c3e04e63898b31a3e87f58f3ab9aefb9a499e8bf137933759`.
  Detailed failure artifact:
  `nf-s18-web-imports-20261029-01/unit-logs/row-module.imports.frontend.lifecycle/vitest-failure-details.json`.
- `CARTULARY_TEST_RUN_ID=nf-s18-lifecycle-20261029-02 make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle`
  passed, owning exit 0. The focused repair rerun does not rewrite the earlier
  failure or constitute complete-owner acceptance.
- `CARTULARY_TEST_RUN_ID=nf-s18-format-20261029-03 make format`,
  `CARTULARY_TEST_RUN_ID=nf-s18-boundary-20261029-01 make lint CHECK=frontend-boundary`,
  and `CARTULARY_TEST_RUN_ID=nf-s18-protocol-20261029-01 make lint CHECK=protocol`
  passed with owning exits 0. Type/format run roots retain their manifest source
  digests; no cross-source final acceptance is inferred.
- Both exact browser plans passed, exit 0:
  `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s18-e2e-nf-selection.json DETAIL=rows`
  and `/tmp/nf-s18-e2e-imports-selection.json`. The Network Flow plan selects
  import, recovery and retained-draft scenarios; the Imports plan selects the
  claimed production workflow. Recovery's deliberate source-read interruption
  now targets source-inspection, preserving the original recovery obligation.

Real-browser runs `nf-s18-e2e-networkflow-20261029-01` and
`nf-s18-e2e-imports-20261029-01` remain pending. S18 is incomplete until their
outcomes/cleanup and the final tracker checkpoint are recorded. All run roots
above are under `.cartulary/test-results`; graph summaries and cleanup records
remain the owning evidence. No full CI, commit or deployment has occurred.


### S18 exit checkpoint

Both real-browser slices passed with owning exit 0:

- `CARTULARY_TEST_RUN_ID=nf-s18-e2e-networkflow-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b,module.networkflow.browser.import_recovery,module.networkflow.browser.import_retained_draft`.
- `CARTULARY_TEST_RUN_ID=nf-s18-e2e-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee`.

Both used source digest
`sha256:300b8b0e64c917fa8668b76a6a4bd1e17c491bb0f6ab05da8156219df01705cb`.
Their exact roots retain run summaries, browser row results and cleanup records.
The repaired lifecycle and frontend boundary/protocol checks also passed on that
source. F25's required public reachability and browser operation are now proved,
including large-input default limits; S15-R must still requalify the final-source
portfolio. F26 and F28 remain open, allocated to S19 and S20. Existing historical
S09/S12-I/S13/S14/S15 claims remain reopened until their remaining obligations
are reconciled. No commit, deployment, broad conversion or reset occurred.

**S18 exit: DONE**, subject to final Markdown checkpoint before S19 begins:
`CARTULARY_TEST_RUN_ID=nf-s18-exit-markdown-20261029-01 make lint-markdown`.

S18 progress and exit Markdown checks passed, owning exits 0. Both browser
runs completed all 23 cleanup steps; repaired lifecycle/boundary/protocol runs
completed all 11/11/13 steps respectively, without failures. S19 now begins.

### S19 — shared row-diagnostic boundary (complete)

F26/V21 is implemented at the shared semantic admission boundary used by preview
and retained-state/Recovery admission. Resource row diagnostics permit only
`cell_scalar_limit_exceeded`, with positive configured limit, actual greater than
limit, exact limit name, attributed column and both hashes, and no raw sample.
The authored rejected-row schema carries the same conditional restrictions;
unit-level error reasons remain available in the unit error contract.

The existing registered preview admission row now forges all four prohibited
reasons with coherent identities/messages/limit fields and proves rejection by
both semantic validation and public owner-result admission. The existing
registered state-5 admission row repeats these forgeries inside restored-state
transactions and proves rejection and rollback. Existing mapped/unmapped/ignored
cell tests and Imports safe-error translation remain green. No stored diagnostics
are rewritten to accommodate invalid content.

Schema regeneration changed the registry digest to
`b7bd50637cb33fb417d0e490f52d8add3203c45bee130a6d35cd4c449beda859`.
Migration 33's explicit known-binding translation now targets that identity.
The change adds source inspection and rejects previously forbidden row payloads;
valid mapping/apply contracts, outcomes and commit ownership remain equivalent.
S20 must prove that translation with frozen legacy identities.

Exact commands and retained roots (all owning exits 0):

- `make task-guide OWNER=module.networkflow`; `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s19-selection.json DETAIL=rows`; `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s19-imports-selection.json DETAIL=rows` selected the rows below.
- Initial `HARNESS_RUN_ID=nf-s19-generate-20261029-01 make generate` passed; that variable is not the run-ID input, so the actual root is `.cartulary/test-results/20261029T221829Z-p13567`. No root was renamed or reused.
- `CARTULARY_TEST_RUN_ID=nf-s19-format-20261029-01 make format`.
- `CARTULARY_TEST_RUN_ID=nf-s19-admission-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.integration.extension_state_v4_v1_rejection`.
- `CARTULARY_TEST_RUN_ID=nf-s19-schema-20261029-01 make source-check CHECK=schema`.
- `CARTULARY_TEST_RUN_ID=nf-s19-generate-20261029-02 make generate` regenerated after the explicit migration digest update.
- `CARTULARY_TEST_RUN_ID=nf-s19-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338`.
- `CARTULARY_TEST_RUN_ID=nf-s19-generated-20261029-01 make source-check CHECK=generated`.
- `CARTULARY_TEST_RUN_ID=nf-s19-drift-20261029-01 make migration-drift`.

Named run roots are `.cartulary/test-results/<CARTULARY_TEST_RUN_ID>`; graph
summaries/manifests and cleanup results are retained there. Admission/schema
source identity was `sha256:8ae0a1181f07c2fccafb5f5f730b71b8f94c1c8ef5ed8c9d63521f38a88a20a9`;
post-migration-digest source identity was
`sha256:6f1e96c35bcb8a04f8d8e3f3060aad5a024f9d3aee7fe61f9288903abf60743f`.
Admission/schema/generated/format cleanup completed all 15/12/14/11 steps.
`nf-s19-imports-20261029-01` cleanup completed 15/15 steps.
`nf-s19-drift-20261029-01` cleanup completed 19/19 steps.

No selected validation failed. No full CI, deployment, commit, or stored-data
conversion occurred. Final compatibility checks and F28 remain allocated to
S20/S15-R. **S19 exit: DONE**, pending the documentation checkpoint below; S20
starts only after its successful owning exit.

`CARTULARY_TEST_RUN_ID=nf-s19-exit-markdown-20261029-01 make lint-markdown`.

S19 documentation checkpoint passed with owning exit 0 and zero failures. S20
now begins; the migration fixture will retain its own database/role/cleanup
authority and admit runtime attachment only after current-head readiness.

### S20 — migrated workflow execution (in progress)

The migration-scratch capability now retains its issued fixture/lease and can
lend that same database to normal application composition only after canonical
head readiness and application crypto-format checks through the runtime role.
It does not initialize missing identity, accept arbitrary handles, or transfer
cleanup ownership. REQ-810/811 were reviewed manually; executable inputs do not
read the specification. Unit checks reject unissued capabilities; service-backed
cases exercise pre-head and missing-crypto rejection and normal startup.

Authored JSON fixtures freeze schema-30 source bytes, mappings, fingerprints,
parser identities and registry `6ed462c191ff86e499efefc153e222f2bc0b7ee7098c6e018443bfcdacb5937b`.
The new registered `module.imports.integration.migrated_workflows` row creates
real jobs/plans/idempotency state at version 30, applies 31/32/33 in the same
database, and runs normal workers. Both targets cover queued/interrupted jobs,
precommit cancellation, explicit before/after-commit faults, pre-upgrade committed
outcomes, two restarts/replays, source/mapping/row-digest/binding failures and
analytical execution without generic-row read permission. A nontransactional
PostgreSQL sequence witnesses owner write attempts independently of row counts.
The row uses exclusive `postgres_migration` isolation and existing standard
resources; it is functional migration evidence, not a resource measurement.

Evidence so far (all roots `.cartulary/test-results/<run-id>`):

- `CARTULARY_TEST_RUN_ID=nf-s20-generate-20261029-01 make generate` failed (owning exit 2): new collaborator IDs were not sorted. `nf-s20-generate-20261029-02` then failed (exit 2) on an unknown collaborator ID. The authored row was corrected to sorted registered owners; `nf-s20-generate-20261029-03` passed, exit 0. Task-guide/target-plan attempts made before that repair failed at the same configuration boundary; no product execution occurred in those attempts.
- `CARTULARY_TEST_RUN_ID=nf-s20-format-20261029-01 make format` failed (exit 2) at that configuration boundary. Format runs `nf-s20-format-20261029-02`, `-03`, `-04`, `-05` passed (exit 0).
- `make task-guide OWNER=module.imports`, `make task-guide OWNER=module.database_migrations`, `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s20-selection.json DETAIL=rows`, and `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s20-harness-selection.json DETAIL=rows` subsequently passed (exit 0).
- `CARTULARY_TEST_RUN_ID=nf-s20-workflows-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_workflows` failed (owning exit 2): fixture membership history had to precede the membership and use `last_version`. Source `sha256:151eaa96fb731c923a00e2e060427efc1d89ab8979144ce30294ce16b868af50`; cleanup 14/14 completed.
- The same exact selection with run ID `nf-s20-workflows-20261029-02` failed (exit 2): the analytical fixture lacked required RFC3339 null/policy members; successful-path jobs also did not finish. Source `sha256:bd7024a832a51eec82a4b03f75a92b3f3ce7d7d80128e305b161e405367b1f4f`; cleanup 14/14 completed.
- `nf-s20-workflows-20261029-03` failed (exit 2): the bounded diagnostic wait showed handler retries, disproving the initial timing hypothesis. The fixtures lacked the frozen Jobs extension ownership/idempotency identity needed for final success publication. Those fields were added; no product timeout or retry policy changed. Source `sha256:a27ffedecb84d873942c5e2baf662632cf5c8337d845c8220fae3e36f81c5648`; cleanup 14/14 completed.
- Failure diagnosis used `make explain-run RESULTS_DIR=.cartulary/test-results/<exact-run-id> TARGET=test-slice DETAIL=summary` and the bounded referenced Go logs. Earlier failures remain retained.
- `CARTULARY_TEST_RUN_ID=nf-s20-harness-20261029-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.test_harness_targeted_operation_validation` passed (exit 0); additional nil/canceled/missing-lease assertions were subsequently added and require revalidation.

The expanded workflow run `nf-s20-workflows-20261029-04` is pending. Committed
fixture response fidelity and capability negatives were further strengthened;
passing earlier runs will not be represented as final-source acceptance. S20
remains IN_PROGRESS; S15-R has not begun.

S20 qualification and repairs:

- `nf-s20-workflows-20261029-04` failed, owning exit 2: the committed generic fixture omitted its record envelope, and analytical cancellation could not settle. Source `sha256:46832623708390a074a0c82de74955559f0d3a0d8342c7608efaaf1c990d9c92`; cleanup 14/14. The envelope was added without changing product admission.
- Cancellation was split into the exact registered row `module.imports.integration.migrated_cancellation`; the workflow row retains 24 other cases across both families. `CARTULARY_TEST_RUN_ID=nf-s20-cancellation-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_cancellation` failed, exit 2, source `sha256:105daa67dab418d2b0daab0915fc9641d07621dcfd9179a2bbd64dee97229f92`; cleanup 14/14. The generic case passed; analytical cancellation remained blocked.
- The same cancellation selection at `nf-s20-cancellation-20261029-02` failed, exit 2, source `sha256:324028b22dced463bc8569e694eff79689d6bc38a1c32601443d92c0c5ab3040`; cleanup 14/14. A five-second request deadline and PostgreSQL lock evidence identified the actual dependency: the owner's incident `FOR UPDATE` lock blocked cancellation's foreign-key reference while owner insertion waited for the test gate. This was a product cancellation defect, not an insufficient runner deadline.
- Incidents now locks immutable incident identity `FOR NO KEY UPDATE`. Owner writes, lifecycle changes and publication still serialize; independent key-share references can proceed so cancellation can interrupt the borrowed transaction. A focused incident test verifies both allowed references and conflicting writes. The owner still never commits or rolls back the borrowed transaction.
- `nf-s20-workflows-20261029-05` passed the 24-case workflow row, exit 0, on source `sha256:324028b22dced463bc8569e694eff79689d6bc38a1c32601443d92c0c5ab3040`; cleanup 14/14. After the lock repair, `nf-s20-cancellation-20261029-03` passed both live cancellation cases, exit 0; cleanup 14/14.
- `CARTULARY_TEST_RUN_ID=nf-s20-incidents-20261029-01 make test-slice OWNER=module.incidents ROWS=module.incidents.integration.participant_reference_lock,module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20` passed, exit 0, cleanup 14/14.
- `CARTULARY_TEST_RUN_ID=nf-s20-bundles-20261029-01 make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2` passed, exit 0, cleanup 14/14. These two runs and cancellation-03 used source `sha256:a5473f6c97ebde1116476cce587371dd81e928f7eb8e3bd7b84c7715edcd31f7`.
- `CARTULARY_TEST_RUN_ID=nf-s20-go-20261029-01 make lint CHECK=go` failed, exit 2, on that source: Staticcheck SA1012 flagged the deliberately nil context in the capability rejection test. Added a local explanation/suppression for that negative test; no production behavior changed. Cleanup 15/15.
- Generate runs `nf-s20-generate-20261029-04`, `-05`, `-06` and format runs `nf-s20-format-20261029-06`, `-07`, `-08`, `-09` passed with owning exits 0. Exact plans `/tmp/nf-s20-complete-selection.json`, `/tmp/nf-s20-cancellation-selection.json`, `/tmp/nf-s20-incidents-selection.json`, and `/tmp/nf-s20-bundles-selection.json` passed through `make target-plan TARGET=test-slice INPUTS=<file> DETAIL=rows`. `nf-s20-progress-markdown-20261029-01` passed, exit 0.

The strengthened write witness observes the first owner write (generic record or
Network Flow table); cancellation blocks the child insertion to prove parent
rollback. Assertions include parent/child rows, generic journal/change sets,
one outcome/commit identity, preserved outcome bytes, idempotency identity and
two normal-runtime restarts/replays. Already-committed outcomes cause zero owner
write attempts. Unknown bindings/registries and source/mapping/row changes fail
closed. Runtime attachment rejects old schema and absent crypto identity without
repairing either; all fixture and runtime cleanup remains harness-owned.

On source `sha256:072a3ff4185ae9a445c319556c799ff4b9913138da682cb5f94002ec27873cb6`,
the following commands passed with owning exits 0. Each run uses
`CARTULARY_TEST_RUN_ID=<run-id>` and retains its manifest, summaries and cleanup
under `.cartulary/test-results/<run-id>`:

| Run ID | Exact Make command | Completed cleanup |
| --- | --- | --- |
| `nf-s20-final-workflows-20261029-01` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_workflows,module.imports.integration.migrated_cancellation` | 14/14 |
| `nf-s20-harness-20261029-02` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.test_harness_targeted_operation_validation` | 10/10 |
| `nf-s20-go-20261029-02` | `make lint CHECK=go` | 15/15 |
| `nf-s20-drift-20261029-01` | `make migration-drift` | 19/19 |
| `nf-s20-schema-20261029-01` | `make source-check CHECK=schema` | 12/12 |
| `nf-s20-generated-20261029-01` | `make source-check CHECK=generated` | 14/14 |
| `nf-s20-boundary-20261029-01` | `make lint CHECK=backend-boundary` | 13/13 |

A final fixture review aligned the committed generic journal's mutation reference
with its already-correct response reference. This changes test data only;
`nf-s20-final-workflows-20261029-02` repeats the exact two-row selection on that
source. Its owning result is required before S20 exits. No full CI, commit,
deployment, or production data conversion has occurred.

The S20 normative consistency audit also found related stale wording in
NF-REQ-057/062a beyond the original §10.1 sentence. Corrected those passages to
allow validation-summary/streamed insertion without mandatory staging and to
label Imports' cancellation barrier, atomic outcome commit and rollback
explicitly. This implements the already-adopted division of responsibility;
there is no new owner decision or runtime contract. V22's review now includes
§8, §9.2, §10.1, Core source access and the Extensions restatement. Product
checks do not consume these Markdown amendments.

### S20 exit checkpoint

`CARTULARY_TEST_RUN_ID=nf-s20-final-workflows-20261029-02 make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_workflows,module.imports.integration.migrated_cancellation`
passed with owning exit 0 on source
`sha256:2e16f7f5b4698b1539a69774f96f067953ade5c050b54b81b5c8b6b6590b44e6`.
All 14 cleanup steps completed. The 26 scenarios close V23 and the outstanding
S12-I/V19 migrated-execution obligation. F28 is resolved; V22 also includes the
corrected §8 ownership sequence and NF-BLOCK-011's three-operation restatement.
No engineering blocker remains for S20. Operational snapshot, quiescence,
coordinated publication and final-source acceptance remain S15-R obligations.

**S20 exit: DONE**, subject to the documentation checkpoint before S15-R:
`CARTULARY_TEST_RUN_ID=nf-s20-exit-markdown-20261029-01 make lint-markdown`.

The S20 Markdown checkpoint passed with owning exit 0 and zero failures.

### S15-R finalization and source freeze

Finalizing through `CARTULARY_TEST_RUN_ID=nf-s15r-finalize-20261029-01 make agent-finalize`.
There is no compatible successful full warm-check root, so retained-run
maintenance is skipped; no redundant warm check is launched to obtain one.
After review, the focused final portfolio will cover current public admission,
owner parsing/diagnostics/atomicity, migrated execution and cancellation,
generic/XLSX regression, source readers/storage/Recovery, browser integration,
incident lock participants, and affected static/generated/harness obligations.
Earlier source runs remain slice evidence, not final-source acceptance.

Finalization passed, owning exit 0, with zero generated files changed. Reviewed
the finalizer summary and active generated operation/schema/migration projections;
`git diff --check` passed. Product source is frozen at dirty digest
`sha256:2e16f7f5b4698b1539a69774f96f067953ade5c050b54b81b5c8b6b6590b44e6`,
HEAD `b6d073974802489edfeb892d39f48b128b9d00c4`. Toolchain digest remains
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
environment/capability digest remains
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Only non-executable documentation may change without a documented repair.
The exact S20 two-row run `nf-s20-final-workflows-20261029-02` has this same
source/toolchain/environment identity and current graph/selection. It supplies
migrated workflow/cancellation evidence without another redundant execution;
other S20 checks used the earlier fixture identity and are requalified below.

Current task guides for module.networkflow, module.imports,
module.database_migrations, module.recovery, web.networkflow, module.incidents
and module.incidentbundles passed with owning exits 0. Exact current plans use
`make target-plan TARGET=test-slice INPUTS=/tmp/nf-s15r-<selection>-selection.json DETAIL=rows`
for nf, imports, migration, recovery, web, web-imports, incidents, bundles,
browser-nf and browser-imports. The completion record below preserves their
selected rows independently of temporary files.

### IMP-04 operational handoff

Deployment is a separate action. Before applying migrations 31–33, operators
must stop admission and quiesce all import workers, verify a recoverable matching
database/environment snapshot, and preflight canonical migration history,
application crypto identity, registry identities and available disk/WAL capacity.
Migration 31 rewrites compressed source values; schedule its locks and storage
growth within the maintenance window. Migrations 32/33 change admission-plan
and source-unit metadata. Do not run old workers against the migrated shape.

Publish matching server/browser assets at the existing major-8/state-5 boundary.
Confirm current schema head and active target registration before resuming
workers. Verify a source-only CSV through inspection, mapping and apply, a
generic Workbook import, and pending-job recovery using the normal application
workflow. Retain the snapshot until operational verification is complete.

Migration 33 translates only predecessor registry
`6ed462c191ff86e499efefc153e222f2bc0b7ee7098c6e018443bfcdacb5937b`
to current registry
`b7bd50637cb33fb417d0e490f52d8add3203c45bee130a6d35cd4c449beda859`
for the explicitly listed selector/binding pairs with equivalent valid apply
contracts. Unknown identities or mismatched bindings remain untransformed and
fail closed. Preflight must identify those pending plans for investigation;
never edit their metadata merely to force execution. The migrated workflow
tests prove both target families, including failures, recovery and replay.

Existing Network Flow tables are neither reinterpreted nor deleted. Logical
source bytes, source/mapping identities, valid approvals, committed outcomes and
replay identities are preserved. Existing generic projections remain usable;
new CSV units honestly expose absent tabular metadata until generic approval.
Analytical plans lose obsolete generic-row digest and rectangle metadata.
Corrected CSV results require original-source re-import with mapping decisions;
there is no automatic historical conversion or compatibility parser.

The bounded-resource claim applies to analytical preview/apply and its Imports
orchestration. Upload buffering, generic Workbook materialization and Recovery-wide
memory are outside it. Two-pass apply increases source reads and transaction
duration; each ordinary reader occupies a snapshot connection until EOF, close,
error or cancellation. Fixed-slice database page assertions are distinct from
application retained-state bounds and are not a performance benchmark.

Rollback restores the previous matching application/database/object-store/key
environment snapshot and assets. Do not downgrade a database after new writes,
mix old/new workers, delete metadata, or reinterpret existing tables to regain
admission. Engineering work here performs no deployment, destructive reset,
commit or production data conversion. Operational rollout must be recorded
separately after this handoff.


### S15-R final focused acceptance

All ten exact target plans passed, owning exits 0. The following fresh
invocations passed with owning exits 0 on the frozen source/toolchain/environment
identities above. Each command is prefixed with
`CARTULARY_TEST_RUN_ID=<run-id>`; every root is
`.cartulary/test-results/<run-id>`. Selected rows are recorded explicitly so the
temporary planning files are not execution dependencies.

- `nf-s15r-nf-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8,module.networkflow.unit.source_inspection_header_only`; PASS, 6/6 work units, 21/21 cleanup steps.
- `nf-s15r-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912,module.imports.integration.xlsx_import_discovery_accepts_a_bounded_open_xml_d7c25ef6c8,module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.target_neutral_public_source_admission,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.unit.target_neutral_csv_source_admission,module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338`; PASS, 7/7 work units, 24/24 cleanup steps.
- `nf-s15r-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap,module.database_migrations.unit.test_harness_targeted_operation_validation`; PASS, 8/8 work units, 19/19 cleanup steps.
- `nf-s15r-recovery-20261029-01`: `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3`; PASS, 3/3 work units, 14/14 cleanup steps.
- `nf-s15r-web-20261029-01`: `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowclient_requires_a_claimed_profile_and_65c861c0d2,web.networkflow.regression.saved_graph_major5_decoder_and_lifecycle`; PASS, 3/3 work units, 12/12 cleanup steps.
- `nf-s15r-web-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle,module.imports.frontend.operator_region_assistant_039d399324,module.imports.frontend.operator_region_coordinator_eb1902fb6e,module.imports.frontend.presentation_recovery,module.imports.frontend.sanitize_import_transport_errors_through_the_sha_1cdfb222a8,module.imports.frontend.skip_reselect_mapping_retention_522ba6591e,module.imports.frontend.typed_transport`; PASS, 3/3 work units, 12/12 cleanup steps.
- `nf-s15r-incidents-20261029-01`: `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.participant_reference_lock,module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20`; PASS, 3/3 work units, 14/14 cleanup steps.
- `nf-s15r-bundles-20261029-01`: `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2`; PASS, 3/3 work units, 14/14 cleanup steps.
- `nf-s15r-browser-nf-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b,module.networkflow.browser.import_recovery,module.networkflow.browser.import_retained_draft`; PASS, 12/12 work units, 23/23 cleanup steps.
- `nf-s15r-browser-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee`; PASS, 12/12 work units, 23/23 cleanup steps.

| Run ID | Exact Make command | Passed work units | Completed cleanup |
| --- | --- | --- | --- |
| `nf-s15r-go-20261029-01` | `make lint CHECK=go` | 6/6 | 15/15 |
| `nf-s15r-types-20261029-01` | `make lint CHECK=types` | 2/2 | 11/11 |
| `nf-s15r-backend-boundary-20261029-01` | `make lint CHECK=backend-boundary` | 4/4 | 13/13 |
| `nf-s15r-frontend-boundary-20261029-01` | `make lint CHECK=frontend-boundary` | 2/2 | 11/11 |
| `nf-s15r-protocol-20261029-01` | `make lint CHECK=protocol` | 4/4 | 13/13 |
| `nf-s15r-schema-20261029-01` | `make source-check CHECK=schema` | 3/3 | 12/12 |
| `nf-s15r-generated-20261029-01` | `make source-check CHECK=generated` | 5/5 | 14/14 |
| `nf-s15r-drift-20261029-01` | `make migration-drift` | 8/8 | 19/19 |

`make explain-run RESULTS_DIR=.cartulary/test-results/<run-id>` exited 0 for
all eighteen new acceptance roots and the compatible S20 final-workflows-02
root. Manifests agree on source, toolchain and environment identities; their
per-command graph/selection identities remain explicit. All selected units
passed, none failed/skipped/cancelled, and all cleanup completed without failures
or blocked steps. Work-unit counts are not test counts. Across these slices and
the compatible S20 root, 50 registered verification rows cover the selected
obligations. This is focused acceptance, not a full-owner, full-CI or release
performance claim. No S15-R product-source repair or drift occurred.

### IMP-04 finding and reopened-exit reconciliation

| Finding / gate | Final disposition and evidence |
| --- | --- |
| F25 / V20 | CLOSED: target-neutral source custody, nullable metadata, lazy generic interpretation, registered owner inspection and browser integration. Public admission proves literal quotes, early/late framing/encoding, 50/51 scope, default 100001-row apply, CSV byte-only admission, generic failure isolation and existing safety/replay. Backend and real-browser final rows pass. |
| F26 / V21 | CLOSED: shared semantic/schema row-diagnostic boundary accepts only the cell limit reason. Coherent header/column/row/accepted-row forgeries fail preview and retained-state admission; valid cell and unit-error translation regressions pass. |
| F27 / V22 | CLOSED: Core, Extensions, Network Flow §8/§9.2/§10.1 and NF-BLOCK-011 consistently assign transaction lifecycle, final barrier/commit and recovery to Imports. Optional preparation is not mandatory durable staging. Human document review is supported by borrowed-reader, atomicity and cancellation tests; executable checks never read Markdown. |
| F28 / V23 / outstanding V19 | CLOSED: frozen valid schema-30 generic/analytical state migrates through 31/32/33, normal workers execute/cancel/recover/replay, pre-upgrade outcomes remain unchanged, duplicate owner writes are excluded, and mismatches fail closed. The compatible S20 final root covers all 26 scenarios; final migration/harness/static checks pass. |
| Reopened S09 | CLOSED: adopted source-admission amendments and transaction-ownership cleanup are reconciled. The user's explicit adoption remains the authority; passing checks are not a substitute for it. |
| Reopened S12-I/V19 | CLOSED: analytical orchestration avoids generic content; migrated workflow evidence now proves the previously missing execution/recovery obligation. |
| Reopened S13 | CLOSED: owner parsing/preview and row-diagnostic boundaries pass independent vectors and public-path checks. |
| Reopened S14 | CLOSED: public reachability, bounded validation/insertion, source protection, atomic rollback and Imports commit/cancellation are qualified on compatible source. |
| S10/S11/S12-M/S12 retained implementation | CONFIRMED by active-publication, scoped-row/graph, byte-preserving storage/bootstrap, bounded-reader and Recovery/operator-region regressions. Prior historical results retain their original scope. |
| S15 / V17–V18 / S15-R | CLOSED: finalization, compatible focused acceptance, documentation checkpoint and operational handoff complete. |

No required engineering check is outstanding. Full CI was intentionally not run under the user's focused-validation
instruction. Pixel-golden suites, independent performance/release qualification,
retained warm-check maintenance, commit and deployment are not selected and are
not claimed. All earlier failed runs and their repairs remain preserved above.

Final documentation checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s15r-exit-markdown-20261029-01 make lint-markdown`.

That checkpoint passed with owning exit 0 and zero failures. A final
`git diff --check` found only an extra blank line at this document's EOF;
the completion edit removed it. This documentation-only repair does not affect
the frozen product source or acceptance compatibility.

**S15-R exit: DONE.** All workstreams/findings and reopened acceptance obligations
are reconciled; no engineering blocker remains. Migration, compatibility,
cancellation/recovery, rollback and operational handoff are complete. The work
remains uncommitted and undeployed. Final record check:
`CARTULARY_TEST_RUN_ID=nf-s15r-exit-markdown-20261029-02 make lint-markdown`.

Final record check passed with owning exit 0 and zero failures; its retained
summary is `nf-s15r-exit-markdown-20261029-02/adhoc/lint-markdown/tool-run-summary.json`.
The final whitespace check also passed. S15-R remains DONE.

## 23. IMP-05 — Approval integrity, explicit route semantics and legacy retirement

### Authorization, baseline and supersession

The user explicitly requested implementation of the complete S21–S24/S15-R2
plan after selecting fresh-environment admission and whole-catalog explicit
mutation semantics. This tracker update is the first implementation change.
Baseline: `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with IMP-03/IMP-04
tracked and untracked changes preserved. No product checks have run in IMP-05.
The current instruction supersedes old-plan preservation/translation and
schema-30 execution obligations, not current-version recovery or replay.

| Finding | Remediation / areas | Rationale, benefit and unresolved risk | Compatibility / validation |
| --- | --- | --- | --- |
| F29 / V24 | Protect and compare source identity in one private transaction operation; implementation and integration tests | Prevent stale approval; retain one transaction owner and explicit protection lifetime | No wire/storage change; deterministic generic/analytical contention, rollback, cancellation and replay |
| F30 / V25 | Complete normative inventory; explicit boolean mutation metadata for all operations; contracts, tooling, runtime, tests and generated outputs | Remove method inference and conflicting security metadata; missing declarations fail closed | Same public request/response shapes; complete catalog, parity, CSRF, assembly and compatibility checks |
| F31 / V26 | Retire prior-format plan translation, fixtures and exclusive harness support; replace useful scenarios with current public admission | Remove unnecessary compatibility burden and invalid test state without losing atomicity/recovery coverage | Fresh database required; bootstrap, current previews/workflows, storage, Recovery and migration drift |

| Workstream | Status | Dependency | Risk / exit |
| --- | --- | --- | --- |
| S21 Authority and acceptance reconciliation | DONE | User decisions | Owner documents and tracker agree; Markdown checkpoint passed |
| S22 Protected approval | DONE | S21 checkpoint | V24 and focused approval regressions passed |
| S23 Explicit mutation semantics | DONE | S22 checkpoint | V25 and generated/compatibility checks passed |
| S24 Legacy retirement/current workflows | DONE | S23 checkpoint | Current workflow matrix, bootstrap, drift and harness qualification passed |
| S15-R2 Validation and handoff | DONE | All preceding checkpoints | Finalization, compatible acceptance and fresh-environment handoff complete |

Affected S09/S16 specification, S17 mapping-admission, S20 compatibility,
S12-I/V19 migrated-execution and S15/S15-R completion declarations are reopened.
Legacy execution obligations will be marked DROPPED, superseded by the selected
fresh-environment policy; their passing historical evidence is preserved.
F25–F28 implementation coverage unrelated to these findings remains retained.

### Execution and validation contract

Use public Make task-guide/target-plan/test-slice workflows; register authored
verification inputs and regenerate through Make. Required checks include
applicable schema/generated, Go, boundary, compatibility, migration drift,
browser/client and harness checks. Run agent-finalize, review changes, freeze
source, then execute compatible final acceptance. Run lint-markdown at every
checkpoint. Full CI and visual-golden updates are not selected by default.
Record exact commands/selections, source identity, owning exits, run roots,
cleanup, failures/repairs, skipped checks and exit decisions. A required failure
leaves its workstream incomplete. Executable evidence never consumes Markdown.

The final handoff requires a fresh canonical-head database and matching
server/browser publication. Prior pending imports and pre-production database
snapshots are not upgrade inputs. Current-format restart/cancellation/recovery/
replay remain supported. Resubmit original sources when needed. Rollback restores
the previous matching application/database environment; it does not downgrade
the new database. Provisioning/reset, commit and deployment remain separate.

### S21 exit checkpoint

Core 00 §1.5 records the user's explicit fresh-environment/whole-catalog decisions
and supersedes only predecessor compatibility obligations. Core 01's twelve-route
Imports inventory now includes source inspection and all six read-only operations;
it requires explicit boolean mutation semantics for the complete public catalog.
Extensions now agrees with fresh-environment admission and current recovery.

Commands: `make task-guide OWNER=module.imports` (exit 0, diagnostic only);
`CARTULARY_TEST_RUN_ID=nf-s21-markdown-20261029-01 make lint-markdown` (owning exit
0, pass, zero failures, run root `.cartulary/test-results/nf-s21-markdown-20261029-01`,
summary `adhoc/lint-markdown/tool-run-summary.json`). No product checks or fixture
resources were required for this documentation slice. Source is baseline HEAD
with preserved dirty implementation plus the specified owner/tracker edits.
**S21 exit: DONE.** F29–F31 remain open for their implementation slices. Final
checkpoint command: `CARTULARY_TEST_RUN_ID=nf-s21-checkpoint-20261029-01 make lint-markdown`.

### S22 exit checkpoint

`protectMappingSourceTx` now locks the selected source before validating the
returned capability and ownership; ordinary metadata lookups remain non-locking.
Approval retains session/unit/source protection through projection, mapping and
idempotency publication. The registered mapping-source-protection integration
row exercises eight generic/analytical contention, cancellation and lifetime
cases using observable database lock waits. No production test hook was added.

Commands (all owning exit 0):

- `CARTULARY_TEST_RUN_ID=nf-s22-generate-20261029-01 make generate`.
- `CARTULARY_TEST_RUN_ID=nf-s22-format-20261029-01 make format`.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s22-selection.json DETAIL=rows JSON=1`.
- `CARTULARY_TEST_RUN_ID=nf-s22-approval-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.mapping_source_protection,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6`.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s22-regression-selection.json DETAIL=summary`.
- `CARTULARY_TEST_RUN_ID=nf-s22-regressions-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.upload_admission_replay_complete_b8684935f2`.

Both test roots under `.cartulary/test-results/` passed 4/4 work units, with
17 cleanup steps each and zero failed/blocked cleanup. Exact `make explain-run
RESULTS_DIR=<root> DETAIL=summary` diagnostics also exited 0. Compatible product
source: `sha256:3bf1e18fd2fa4baa097a25766bd9a15563dfedc7d23ef154a8734d767e318c3d`.
No test failure or repair was needed. Full CI remains unselected; final static
and compatibility obligations remain S23/S15-R2 work. F29/V24 is resolved.
**S22 exit: DONE**, subject to the checkpoint command below before S23 begins.
`CARTULARY_TEST_RUN_ID=nf-s22-checkpoint-20261029-01 make lint-markdown`.

### S23 exit checkpoint

All 130 public operations now declare `x-cartulary-state-changing` (47 read-only,
83 state-changing). Assembly and generation reject missing/non-boolean values;
there is no method fallback. Imports resolves method/classification from generated
operations and rejects missing/duplicate/foreign/unbound entries before serving.
Parity checks now compare the generated flag directly. Compatibility treats the
extension as security-sensitive and the authored 2.0.0 change set is reconciled.

Human annotation review includes the six read-only POSTs (Imports preview and
inspection, Workbook query and locate, composition validation, Evidence metadata),
plus auth-completion GETs and download-handle consumption. Core REQ-01-328a's
read-only Evidence metadata retains its independent cookie-CSRF requirement;
auth callbacks/handle redemption retain their protocol protections. No security
policy was inferred from mutation metadata outside Imports. The narrow operation
catalog test package was added to the closed harness roots, schema, registry and
normative harness inventory; arbitrary tool roots remain excluded.

Generation failures and repairs are preserved: `nf-s23-generate-20261029-01`
failed on stale change-set fingerprints; 02 failed on unsorted test selectors;
03 failed on the unregistered operation-catalog test root; 04 failed on the
remaining runner-registry mirror. Each owning Make exit was 2, underlying step
exit 1. Exact-root `make explain-run` plus bounded referenced logs identified each
cause. Authored inputs were repaired; `nf-s23-generate-20261029-05` passed (exit 0).
No product-test failure occurred. `nf-s23-format-20261029-01` also passed.

`make task-guide OWNER=platform.openapi` and target plans for test-slice inputs
`/tmp/nf-s23-{openapi,imports,workbook,composition}.json` and harness-contract input
`/tmp/nf-s23-harness.json` all exited 0. The following exact invocations use
`CARTULARY_TEST_RUN_ID=<run-id>`; roots are `.cartulary/test-results/<run-id>`.

| Run ID | Make command | Result / cleanup |
| --- | --- | --- |
| nf-s23-openapi-20261029-01 | `make test-slice OWNER=platform.openapi` | PASS 5/5; 14 cleanup steps |
| nf-s23-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.unit.behavior_contract,module.imports.integration.target_neutral_public_source_admission,module.imports.integration.upload_admission_replay_complete_b8684935f2` | PASS 5/5; 18 cleanup steps |
| nf-s23-workbook-20261029-01 | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.locator_public_contract,module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` | PASS 3/3; 14 cleanup steps |
| nf-s23-composition-20261029-01 | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.public_lifecycle_authority` | PASS 3/3; 14 cleanup steps |
| nf-s23-schema-20261029-01 | `make source-check CHECK=schema` | PASS 3/3; 12 cleanup steps |
| nf-s23-generated-20261029-01 | `make source-check CHECK=generated` | PASS 5/5; 14 cleanup steps |
| nf-s23-go-20261029-01 | `make lint CHECK=go` | PASS 6/6; 15 cleanup steps |
| nf-s23-boundary-20261029-01 | `make lint CHECK=backend-boundary` | PASS 4/4; 13 cleanup steps |
| nf-s23-harness-20261029-01 | `make harness-contract SUITE=contract` | PASS; cleanup checked in exact-root summary |

All owning exits were 0; exact-root summary diagnostics exited 0 and cleanup had
zero failed/blocked steps. Compatible source:
`sha256:86ab1eceab050d310c4383f13b50a1abe6fa3eb224c20d24969b2733b06a8402`.
Full CI remains unselected; final browser/client checks remain S15-R2 obligations.
F30/V25 is resolved. **S23 exit: DONE**, subject to the documentation checkpoint
before S24: `CARTULARY_TEST_RUN_ID=nf-s23-checkpoint-20261029-01 make lint-markdown`.

### S24 implementation and qualification

Migrations 31–33 retain canonical numbering and structural constraints but no
longer reserialize predecessor values, clear old analytical plan metadata or
translate old registry identities. Core 01 REQ-01-620b now states the adopted
fresh-environment boundary directly. The frozen schema-30 fixtures and exclusive
migration-scratch/runtime attachment capability are removed, including their
private fields and tests. General migration isolation and canonical bootstrap
remain. Current-format binding mismatches continue to fail closed.

The replacement workflow matrix obtains source, mapping, selection and apply
admission through public routes on a fresh current database. It covers both
target families across queued/interrupted execution, cancellation, source/mapping/
binding/registry mismatches, generic-row digest failure, analytical execution
without generic-row read privileges, precommit and postcommit faults, restart,
and repeated replay. Nontransactional write witnesses detect attempted duplicate
owner effects. Generic preview asserts exact column/cell members before mapping,
after projection materialization, and after successful restarts. Bootstrap verifies
nullable projection invariants, EXTERNAL storage, byte/digest equality and source
copy behavior; current workflows verify target-dependent plan constraints.

Legacy F28/V23 and the migrated-execution part of V19, S20 and S12-I are
**DROPPED—superseded by the fresh-environment decision**. The compressed
predecessor rewrite and upgrade preservation claims in S12-M/S12-I are likewise
dropped. Their historical runs remain unchanged and do not qualify this build.
Current recovery/replay, bounded readers, byte integrity and canonical migration
isolation remain required, mapped to V26 and final acceptance below.

Commands so far (roots under `.cartulary/test-results/`, exact run ID supplied
through `CARTULARY_TEST_RUN_ID`):

- `nf-s24-generate-20261029-01`: `make generate`, owning exit 0.
- `nf-s24-format-20261029-01`: `make format`, owning exit 0.
- `make task-guide OWNER=module.database_migrations`, exit 0.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s24-current.json DETAIL=summary`
  and `/tmp/nf-s24-migration.json`, both exit 0.
- `nf-s24-bootstrap-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap,module.database_migrations.unit.test_harness_targeted_operation_validation`, PASS 8/8 work units, 19 cleanup steps.
- `nf-s24-drift-20261029-01`: `make migration-drift`, PASS 8/8, 19 cleanup steps.
- `nf-s24-go-20261029-01`: `make lint CHECK=go`, PASS 6/6, 15 cleanup steps.
- `nf-s24-current-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.current_workflows,module.imports.integration.current_cancellation`, FAIL 2/3 work units; Make exit 2, runner exit 10, product/test_assertion_failure; 14 cleanup steps all completed. Both binding-change cases expected target unavailability instead of the existing identity-mismatch `source_changed` result. The runtime failed closed correctly; the test expectation was repaired. Cancellation passed. No deadlock or runtime repair occurred.

The first four qualification roots use source
`sha256:11be81376fce1b8b4ba8131aadeac54138ee63d0cbda87d38a0dc693c0b9c5d3`.
Successful checks had owning exit 0 and no failed/blocked cleanup. Exact-root
`make explain-run RESULTS_DIR=<root> DETAIL=summary` diagnosed each result with
exit 0. The repaired workflow rerun is `nf-s24-current-20261029-02`, using the
same exact command and source
`sha256:12915ad34d9d20eda9ffc679e76da01dcd0c2f14718b91806d1c67791497e0fc`.
The repaired run passed all 3 work units with owning exit 0, 14 completed cleanup
steps and zero failed/blocked cleanup. Both registered rows passed, covering all
26 scenarios. The retirement scan found no schema-30 workflow, translation or
runtime-attachment references in affected active code, migrations and routing;
`git diff --check` passed. No required S24 check remains failed. F31/V26 is
resolved. **S24 exit: DONE**, subject to the following checkpoint before S15-R2:
`CARTULARY_TEST_RUN_ID=nf-s24-checkpoint-20261029-01 make lint-markdown`.
Full CI is unselected; final static/browser and compatible regression evidence
remain S15-R2 work.

### S15-R2 finalization, frozen acceptance and handoff

S21, S22, S23 and S24 Markdown checkpoint commands all passed with owning exit 0
and zero failures. Each checkpoint completed before its successor began.

`CARTULARY_TEST_RUN_ID=nf-s15r2-finalize-20261029-01 make agent-finalize`
passed with owning exit 0, 1/1 work units and 10 completed cleanup steps.
Generated changes were reviewed against the pre-finalization diff/status;
finalization introduced no additional changes. Source was then frozen:

- HEAD: `b6d073974802489edfeb892d39f48b128b9d00c4`, dirty working tree preserved.
- Source: `sha256:12915ad34d9d20eda9ffc679e76da01dcd0c2f14718b91806d1c67791497e0fc`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

The S24 current-workflow rerun has exactly these identities and remains compatible
acceptance for its 26 scenarios. No product repair followed it. Final source/diff
review found no drift. Subsequent tracker edits are documentation only.

`make task-guide OWNER=module.networkflow` and `OWNER=module.recovery` exited 0.
Every final selection below was resolved successfully with
`make target-plan TARGET=<target> INPUTS=/tmp/nf-s15r2-<selection>.json DETAIL=summary`,
using the exact inputs reproduced in its command. Selection names are the run-ID
portion between `nf-s15r2-` and `-20261029-01`. Planning is not test evidence.

Each command below was executed with `CARTULARY_TEST_RUN_ID=<run-id>`.
Every owning exit was 0. Roots are `.cartulary/test-results/<run-id>`;
all selected work units passed, with zero failed/skipped/cancelled units and
zero failed/blocked cleanup steps. Work-unit counts are not test counts.

| Run ID | Exact Make command and selection | Work units / completed cleanup |
| --- | --- | --- |
| nf-s15r2-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.unit.behavior_contract,module.imports.integration.mapping_source_protection,module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912,module.imports.integration.xlsx_import_discovery_accepts_a_bounded_open_xml_d7c25ef6c8,module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.target_neutral_public_source_admission,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.unit.target_neutral_csv_source_admission,module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338` | 7/7; 24 steps |
| nf-s15r2-nf-20261029-01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8,module.networkflow.unit.source_inspection_header_only` | 1/1; 10 steps |
| nf-s15r2-openapi-20261029-01 | `make test-slice OWNER=platform.openapi` | 5/5; 11 steps |
| nf-s15r2-migration-20261029-01 | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap,module.database_migrations.unit.test_harness_targeted_operation_validation` | 8/8; 19 steps |
| nf-s15r2-recovery-20261029-01 | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3` | 3/3; 14 steps |
| nf-s15r2-browser-nf-20261029-01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b` | 12/12; 23 steps |
| nf-s15r2-browser-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee` | 12/12; 23 steps |
| nf-s15r2-harness-20261029-01 | `make harness-contract SUITE=contract` | 13/13; 22 steps |
| nf-s15r2-workbook-20261029-01 | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.locator_public_contract,module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` | 3/3; 14 steps |
| nf-s15r2-composition-20261029-01 | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.public_lifecycle_authority` | 3/3; 14 steps |
| nf-s15r2-web-nf-20261029-01 | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 2/2; 11 steps |
| nf-s15r2-web-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle,module.imports.frontend.typed_transport` | 3/3; 12 steps |
| nf-s15r2-go-20261029-01 | `make lint CHECK=go` | 6/6; 15 steps |
| nf-s15r2-types-20261029-01 | `make lint CHECK=types` | 2/2; 11 steps |
| nf-s15r2-backend-boundary-20261029-01 | `make lint CHECK=backend-boundary` | 4/4; 13 steps |
| nf-s15r2-frontend-boundary-20261029-01 | `make lint CHECK=frontend-boundary` | 2/2; 11 steps |
| nf-s15r2-protocol-20261029-01 | `make lint CHECK=protocol` | 4/4; 13 steps |
| nf-s15r2-schema-20261029-01 | `make source-check CHECK=schema` | 3/3; 12 steps |
| nf-s15r2-generated-20261029-01 | `make source-check CHECK=generated` | 5/5; 14 steps |
| nf-s15r2-drift-20261029-01 | `make migration-drift` | 8/8; 19 steps |

For each exact root, `make explain-run RESULTS_DIR=<root> DETAIL=summary`
exited 0. Run manifests agree on the frozen identities; command-specific graph,
row, fixture and case identities remain in their original manifests. This is
focused acceptance of the stated obligations, not a full-CI or whole-repository
conformance claim. Compatible cache reuse, where admitted by the harness, is not
claimed as a fresh performance sample.

### IMP-05 finding and acceptance reconciliation

| Finding / reopened claim | Disposition |
| --- | --- |
| F29 / V24 / S17 source approval | CLOSED. Protected capability comparison and complete association checks precede durable approval. Generic/analytical race, unchanged-content contention, cancellation, lock lifetime and replay cases pass, including no projection/mapping/selection/idempotency effects after rejection. |
| F30 / V25 / S09/S16 specification closure | CLOSED. Core inventories twelve Imports operations with six read-only and six mutations. All 130 public operations declare boolean semantics; assembly/generation fail closed, generated flags feed Imports, independent CSRF requirements remain explicit, and security-sensitive compatibility accounting passes. |
| F31 / V26 / S20 compatibility | CLOSED under the adopted fresh-environment boundary. Old translation, frozen workflows and exclusive runtime attachment are absent. Fresh bootstrap, source storage/copy, current public preview and both target families' execution/cancellation/recovery/replay pass. |
| F28 / V23 / legacy part of V19/S12-I/S20 | DROPPED—superseded by the fresh-environment decision. Historical migrated-workflow runs remain historical; no prior-format upgrade guarantee is carried forward. |
| Current V19/S12-I and S14 guarantees | CONFIRMED. Analytical execution succeeds without generic-row SELECT authority; current binding/source/mapping/row-digest changes fail closed; atomic rollback and committed-result recovery/replay preserve one set of owner effects. |
| F25–F27 / S13 and affected S14 regressions | CONFIRMED. Source admission, strict diagnostic admission, borrowed transaction ownership, 50/51 preview scope and full apply, default 100001-row public import, generic/XLSX paths and both browser workflows pass. |
| S15/S15-R/S15-R2 final acceptance | CLOSED by compatible focused evidence and this operational handoff, subject to the final Markdown record check below. |

### Fresh-environment operational handoff

The revised application requires a **fresh database initialized to canonical
head through the normal migration chain**. Earlier pre-production databases,
pending import plans and snapshots are not upgrade inputs for this build.
Original sources must be resubmitted through current upload, mapping approval
and apply when prior development results are needed. Existing development data
is neither reset nor automatically deleted by these changes.

Publish matching server/browser assets together at Network Flow major 8/state 5
and the existing coordinated API candidate. Do not mix old/new workers or point
this application at a predecessor database. Canonical bootstrap and migrations
31–33 establish source storage and current plan/projection constraints; they do
not perform historical plan conversion. Current-format restart, cancellation,
Recovery and replay remain supported. Recovery after a committed unit does not
repeat owner writes.

Upload buffering, generic Workbook materialization and Recovery-wide memory are
outside the bounded analytical preview/apply claim. CSV interpretation changes
require original-source re-import when different results are desired. No existing
Network Flow table is reinterpreted by this implementation.

Rollback restores the previous matching application/database environment snapshot;
it never downgrades the new database after writes or deletes metadata to force
admission. Provisioning, database resets, commit and deployment are separate
unperformed operational actions.

All selected engineering checks passed. Full CI, visual-golden updates,
independent performance/release qualification and retained warm-check maintenance
were not selected; no compatible full warm-check root was supplied to finalization.
No required check was skipped, and no engineering blocker remains. Historical
failures and their repairs remain recorded. Final record command:
`CARTULARY_TEST_RUN_ID=nf-s15r2-checkpoint-20261029-01 make lint-markdown`.

The final record check passed with owning exit 0 and zero failures; its retained
summary is `nf-s15r2-checkpoint-20261029-01/adhoc/lint-markdown/tool-run-summary.json`.
A whitespace review found and removed one extra blank line at this document's
EOF. Final documentation-only confirmation:
`CARTULARY_TEST_RUN_ID=nf-s15r2-checkpoint-20261029-02 make lint-markdown`.
**S15-R2 exit: DONE.** All implementation, validation, tracker reconciliation
and operational handoff obligations are complete. Commit and deployment remain
unperformed.

## 24. IMP-06 resource safety and security accounting

Baseline: `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with the existing
IMP-05 dirty tracked and untracked work preserved. Full file fingerprints,
prior working-tree patch and index patch are retained outside the repository at
`/tmp/cartulary-nf-remediation-h8v4nkoe`. This invocation implements the user-approved
plan; the earlier planning-only restriction is superseded. No commit, deployment,
database reset or historical data conversion is included.

| Finding | Areas / remediation | Benefit and unresolved risk | Compatibility / validation |
| --- | --- | --- | --- |
| F32 / P1 | Core owners, shared lexer, Imports preview/materialization/apply, client contracts and tests: incremental preview, sparse actual cells, cancellation | Prevent rectangular memory amplification and fabricated source values; unresolved small inputs can exhaust memory | Missing trailing CSV fields are absent; coordinated fresh-environment API/client candidate; V27 |
| F33 / P2 | Common API owner, semantic compatibility traversal, end-to-end tests and release review records | Security accounting follows semantic surfaces even under structural changes; unresolved declarations bypass review | Additional exact security fingerprints in existing 2.0.0 candidate; immutable release baseline retained; V28 |

| Workstream | Status | Depends on | Risk / binary exit |
| --- | --- | --- | --- |
| S25 Owner and tracker closure | DONE | authorized plan | Adopt concrete coordinated owners; Markdown checkpoint passes |
| S26 Characterization and registration | DONE | S25 | Small regressions expose both defects without oversized baseline allocation; registered Make coverage |
| S27 CSV architecture and consumers | DONE | S26 | V27: sparse geometry, cancellation, source protection, mapping, replay and browser coverage pass |
| S28 Security accounting | DONE | S27 | V28: Compare matrix, exact review records and production gate pass |
| S15-R3 Frozen acceptance and handoff | DONE | S27, S28 | Finalize, freeze source, compatible required portfolio, successful cleanup and final record |

V27 requires the 10,000-column / 10,000 one-field-row case, first-50 retention,
late width/framing, explicit empty versus absent, incremental cancellation,
public source/mapping/replay and client alignment. The isolated builder allocation
budget is 64 MiB per operation, excluding fixture/harness setup; it is not a runtime
admission limit. V28 requires whole-path, operation and property add/remove/change,
false and empty-array presence, root/operation security, schemes, replacement,
escaping, deterministic deduplication, descriptive/unrelated-property negatives,
and version/change-set rejection. Tests and release records never consume Markdown.

The read-only planning observations of `nf-s15r2-imports-20261029-01`,
`nf-s24-current-20261029-02`, `nf-s15r2-openapi-20261029-01`, and
`nf-s15r2-browser-imports-20261029-01` report historical passes. Their source is
`sha256:12915ad34d9d20eda9ffc679e76da01dcd0c2f14718b91806d1c67791497e0fc`;
those observations did not audit current source or observe owning exits.
They do not cover the new findings. Earlier outer exits and cleanup remain
recorded in §23, without alteration.

### IMP-06 execution record

| Workstream / event | Change or evidence | Result / next gate |
| --- | --- | --- |
| S25 start | User policy selections and implementation request recorded in Core 00 §1.6; Core 01 separates inspection/materialization and adopts absent CSV fields; common API owns nested security accounting; Core 04 limits are explicit about XLSX | Owner changes prepared under the approved concrete plan; Markdown checkpoint next |
| S25 checkpoint | `CARTULARY_TEST_RUN_ID=nf-s25-markdown-20261029-01 make lint-markdown` | PASS, owning exit 0; summary `.cartulary/test-results/nf-s25-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json` |
| S26 regressions | `make test-slice OWNER=module.imports ROWS=module.imports.unit.target_neutral_csv_source_admission` at `nf-s26-imports-baseline-20261029-01`; `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.release_compatibility` at `nf-s26-openapi-baseline-20261029-01` | Both expected product/test_assertion_failure, owning exits 2; 10 cleanup steps each, none failed/blocked. Small ragged rows expose padding; Compare exposes missing nested declarations. Exact-root summary diagnosis exits 0. The oversized fixture was not executed against vulnerable code. |
| S26 checkpoint | `CARTULARY_TEST_RUN_ID=nf-s26-markdown-20261029-01 make lint-markdown` | PASS, owning exit 0 |
| S27 initial focused checks | `nf-s27-unit-20261029-01` and `nf-s27-tabular-20261029-01` | Both product assertion failures, owning exits 2; all 10 cleanup steps passed each. Ragged resource test passed (preview 4,759,088 allocated bytes; approval 12,709,448). Corrected a non-clearable test field and restored exact clipboard lexical error text; no limit was relaxed. |
| S27 browser diagnosis | `nf-s27-browser-20261029-01`, Imports browser/lifecycle/typed-transport selection | Product assertion failure, owning exit 2; 12/14 work units passed, all 25 cleanup steps completed. Existing transport validation still required rectangular CSV rows, hiding the mapping UI. Corrected CSV-only sparse admission and added ordinal/width/empty-row negatives; XLSX remains rectangular. |
| S27 checkpoint | Unit/behavior `nf-s27-unit-20261029-02`; shared kernel `nf-s27-tabular-20261029-02`; seven integration rows `nf-s27-integration-20261029-01`; browser/transport `nf-s27-browser-20261029-02`; Markdown `nf-s27-markdown-20261029-01` | PASS, all owning exits 0. Public sparse preview→approval→apply and restart/replay passed; 17 integration cleanup steps completed. Browser confirms absent fields render without padding. Focused evidence precedes final frozen acceptance. |
| S28 accounting review | `nf-s28-generate-20261029-01` and explicit report `nf-s28-accounting-20261029-01` | Expected gate failure (owning exits 2): previous change set covered 559 of 600 records. Reviewed 40 newly visible security records for 20 existing operation additions against their owner inputs and prior structural rationales, plus the sparse-preview description. Preserved all 559 previous approvals, immutable 1.0.0 release and 2.0.0 candidate; added exact owner/value-specific records. All 13 compatibility-run cleanup steps completed. |
| S28 checkpoint | `nf-s28-generate-20261029-02`, `nf-s28-openapi-20261029-01` | PASS, owning exits 0. Canonical generation and complete OpenAPI owner selection cover nested Compare matrices, false-positive exclusions, exact review/version negatives, assembly, runtime parity, and production release gate. |

The final frozen portfolio is gap-scoped: complete Tabular Ingest and OpenAPI
owner selections; Imports resource/behavior/nullability unit rows, public source
admission, mapping/source protection, current-workflow recovery/cancellation,
CSV/XLSX/region/reimport controls, the public 100,001-row Network Flow admission
case, Imports browser/transport/lifecycle, and Network Flow preview/source
inspection plus its discovery browser workflow. This is not a new complete
Network Flow owner qualification or a full CI/performance/visual-baseline claim.
Final source/generated/schema/catalog, Go/type/protocol/boundary lint and Markdown
checks complete the portfolio. Any source repair invalidates the frozen cohort;
its affected acceptance must be re-established before closure.

### IMP-06 frozen validation and handoff

Frozen executable source:
`sha256:78e53e3695ccbdca74cfd3bde529c372319af6373e565cce90b1320f8a0b335a`.
Every final product/source-check/lint/compatibility run below reports this same
identity. Native owning exits were observed as 0, independently of summary
status. Final main product selections have no missing, skipped or failed selected
rows; all published cleanup steps completed. OpenAPI reused five compatible
cached work units; its complete-owner audit validates retained selection,
provenance, current source and cleanup. Tabular Ingest's complete-owner audit also
passes. Imports and Network Flow are deliberately exact-row selections, not
complete-owner audit claims.

All run roots below are under `.cartulary/test-results/`. The retained
`run-manifest.json` contains each exact command input/row selection;
`run-summary.json`, `target-summaries/`, `rows/`, and `cleanup-results.json` retain
outcomes and cleanup. No Markdown is consumed as executable evidence.

| Command / scope | Run ID | Observed result |
| --- | --- | --- |
| `make agent-finalize` | `nf-s15r3-finalize-20261029-01` | PASS; generation/finalization completed before frozen acceptance |
| `make test-slice OWNER=module.imports ROWS=…` (17 rows: the five unit, nine integration and three client/browser rows in its manifest) | `nf-s15r3-imports-20261029-01` | PASS; 17/17 rows, 19/19 work units; 36 cleanup steps completed |
| `make test-slice OWNER=module.networkflow ROWS=…` (three preview/inspection unit rows and production discovery browser row) | `nf-s15r3-networkflow-20261029-01` | PASS; 4/4 rows; 24 cleanup steps completed |
| `make test-slice OWNER=module.tabularingest` | `nf-s15r3-tabular-20261029-01` | PASS; both complete-owner rows; 10 cleanup steps completed |
| `make test-slice OWNER=platform.openapi` | `nf-s15r3-openapi-20261029-01` | PASS; all six owner rows; five compatible cache hits; nine cleanup steps completed |
| `make test-evidence-audit OWNER=module.tabularingest EVIDENCE_ROOTS_FILE=…` | `nf-s15r3-tabular-audit-20261029-01` | PASS complete-owner current-source audit; manifest `/tmp/cartulary-nf-remediation-h8v4nkoe/tabular-evidence.json` |
| `make test-evidence-audit OWNER=platform.openapi EVIDENCE_ROOTS_FILE=…` | `nf-s15r3-openapi-audit-20261029-01` | PASS complete-owner current-source audit; manifest `/tmp/cartulary-nf-remediation-h8v4nkoe/openapi-evidence.json` |
| `make source-check CHECK=generated` | `nf-s15r3-generated-20261029-01` | PASS |
| `make source-check CHECK=schema` | `nf-s15r3-schema-20261029-01` | PASS |
| `make source-check CHECK=catalog` | `nf-s15r3-catalog-20261029-01` | PASS |
| `make lint CHECK=go` | `nf-s15r3-go-20261029-01` | PASS |
| `make lint CHECK=types` | `nf-s15r3-types-20261029-01` | PASS |
| `make lint CHECK=protocol` | `nf-s15r3-protocol-20261029-01` | PASS |
| `make lint CHECK=backend-boundary` | `nf-s15r3-backend-boundary-20261029-01` | PASS |
| `make lint CHECK=frontend-boundary` | `nf-s15r3-frontend-boundary-20261029-01` | PASS |
| `make openapi-compatibility-check` | `nf-s15r3-compatibility-20261029-01` | PASS; exact 600-record review set, immutable baseline 1.0.0, candidate 2.0.0 |
| `make lint-markdown` | `nf-s15r3-markdown-20261029-01` | PASS, owning exit 0; final handoff documentation checkpoint |

F32/V27 closes the reopened generic resource portion of F25/V20: preview retains
only header/column metadata and the first 50 actual rows, scans the remainder for
geometry/framing/integrity, and materializes approval separately. Missing trailing
fields never become blank cells or unknown-value warnings; explicit empty fields
retain their mapped empty-value policy. The browser consumes sparse rows, while
XLSX remains rectangular. The 10,000-column/10,000-row adversarial regression
passes the separate 64 MiB builder budgets. Focused measured totals were about
4.8 MB preview and 12.7 MB materialization; these are fixture allocation checks,
not a production RSS bound or performance qualification. Admission remains
byte-only for CSV, including the passing public 100,001-row analytical case.

F33/V28 closes the reopened F30/V25/S23 accounting portion: structural records
remain, while added/removed/replaced paths and operations expose exact nested
mutation/security declarations. Root requirements and security schemes receive
the same treatment. Structural matching excludes unrelated schema properties and
examples, preserves descriptive-only changes, and supports reusable path items,
webhooks and callbacks. Compare tests establish absence versus false/empty,
escaping, deterministic unique records, version policy and exact review rejection.
The 40 newly visible security entries are accounting for already adopted
operations, not new route behavior.

Compatibility handoff: ship the coordinated server/client candidate together.
Missing trailing CSV fields are intentionally absent; there is no rectangular
compatibility mode. The established fresh-database migration boundary, Network
Flow major 8/state 5 and API candidate 2.0.0 remain in force. Old pre-production
imports require explicit reimport; no historical projection converter or extra
major is introduced. No database reset, commit, deployment or release publication
was performed. Original dirty work and index were preserved; the six paths absent
at entry remain absent. The task delta is 32 files, including authored owners,
tests, documentation and canonical generated projections.

Full CI, complete Network Flow owner qualification, visual-golden refresh,
performance publication and deployment are outside this gap-scoped acceptance.
Historical passes and failed intermediate runs remain available with their
original classifications. All required product and documentation checks passed;
none were skipped. The acceptance claims reopened by IMP-06 are closed within
the frozen gap-scoped portfolio above.

## 25. IMP-07 descriptive security changes and preview retention

Baseline: `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with existing dirty
tracked/untracked work and index preserved. Entry fingerprints, source copies,
and prior/index patches are retained at `/tmp/cartulary-imp07-kmxt5qho`.
The maintainer authorized implementation after selecting affected-gap final
qualification. Core 01 §17.1 and Core 04 acceptance text clarify the already
adopted descriptive-change and bounded-preview rules; they do not introduce a new
runtime profile, OAuth authentication capability, CSV admission limit, migration,
API version or state converter. This implementation request supersedes the earlier
planning-only boundary.

| Finding / obligation | Remediation / areas | Rationale, benefit and unresolved risk | Compatibility / exit |
| --- | --- | --- | --- |
| F34 / V28 | Clarify owner; make security classification change-aware; correct and expand Compare tests; reconcile tracker | Distinguish scope identity from its description without weakening security additions/removals; otherwise prose edits require major releases and a wrong test expectation becomes policy | No public interface change; description-only fixture changes become non-behavioral. Production has no OAuth2 scheme; the existing 600-record report/change set must remain unchanged |
| F35 / V27 | Clarify acceptance; add narrow 10,000/100,000-row preview tests with a 1 MiB per-call allocation bound and a full-materialization negative control; register tests and document evidence | Separate bounded-retention protection from the existing 64 MiB dense-padding guard; otherwise full sparse materialization followed by truncation would pass | No planned parser/runtime change. Exact first-50 results, full-tail validation and discriminating allocation bounds must pass |

| Phase | Status | Depends on | Principal risk / binary exit |
| --- | --- | --- | --- |
| S29 Owner and tracker clarification | DONE | authorized plan | Clarified owners; Markdown checkpoint passed; earlier evidence preserved |
| S30 Characterization | DONE | S29 | Corrected OAuth expectations failed on baseline; preview passed and full-materialization control exceeded its bound; selectors regenerated |
| S31 Classifier correction and focused checks | DONE | S30 | Exact structural/type/change-kind guards implemented; both focused selections passed |
| S15-R4 Frozen qualification and handoff | DONE | S31 | Finalized and frozen; affected-gap portfolio, complete OpenAPI audit, exact production report equality, cleanup and documentation checkpoint passed |

V28 covers all four OAuth flow kinds; ordinary, reserved-looking and escaped
scheme/scope names; empty descriptions; scope add/remove/rename; malformed value
replacements; flow/requirement/scheme changes; structural records and exact
version/change-set enforcement. V27 retains the 10,000-wide ragged test and adds
1 MiB narrow-preview budgets at 10,000 and 100,000 data rows. A separate 10,000-row
full-materialization control must exceed 1 MiB. Fixture construction/assertions
are outside sequential TotalAlloc measurements; no timing/RSS gate, GC suppression
or runtime quota is introduced. Tail width, malformed quote, terminal I/O/integrity
failure and cancellation must still prevent a truncated scan from passing.

The final portfolio is complete `platform.openapi`, the exact Imports
`module.imports.unit.target_neutral_csv_source_admission` row, compatibility,
generated/schema/catalog checks, Go/backend-boundary lint, Markdown, and the
complete OpenAPI owner evidence audit. The broader S15-R3 passes remain historical
at `sha256:78e53e3695ccbdca74cfd3bde529c372319af6373e565cce90b1320f8a0b335a`.
No browser/integration rerun, complete Imports/Network Flow owner qualification,
performance publication, commit, database reset or deployment is part of IMP-07.
Tests, generators and release evidence must not consume Markdown.

### IMP-07 execution record

| Event | Command / evidence | Result / next gate |
| --- | --- | --- |
| Entry | Existing source and owner guidance inspected; immutable compatibility baseline copied; `make task-guide OWNER=platform.openapi` and `OWNER=module.imports` | Read-only discovery passed; owner/tracker clarification prepared |
| Owner checkpoint | `make lint-markdown`, `nf-imp07-owners-markdown-20261030-01` | Passed, outer exit 0 |
| Characterization preparation | `make generate`, `nf-imp07-characterization-generate-20261030-01` | Passed, outer exit 0; generated topology render index updated through Make |
| Expected baseline failure | `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.release_compatibility`, `nf-imp07-characterization-openapi-20261030-01` | Product assertion failure, outer exit 2; existing scope descriptions incorrectly classified security-sensitive. One failed work unit; 10 cleanup steps succeeded. Failure retained, not acceptance evidence |
| Retention characterization | `make test-slice OWNER=module.imports ROWS=module.imports.unit.target_neutral_csv_source_admission`, `nf-imp07-characterization-imports-20261030-01` | Passed, outer exit 0; narrow previews stayed within 1 MiB, full 10,000-row materialization exceeded 1 MiB, complete-tail checks passed; 10 cleanup steps succeeded |
| Correction and formatting | Change-aware classifier and structural/type guards; `make format`, `nf-imp07-format-20261030-01` | Passed, outer exit 0; no CSV runtime change or release approval refresh |
| Focused checks | Same two exact row selections, `nf-imp07-focused-openapi-20261030-01` and `nf-imp07-focused-imports-20261030-01` | Both passed, outer exits 0; ready for finalization |

Run IDs above resolve beneath `.cartulary/test-results/`. Characterization source
digest: `sha256:2a196e58529817829c684212c2c66c12a15f562aed87dc40211130d2987f61bc`.
`make explain-run ... DETAIL=summary` confirmed both characterization outcomes;
`DETAIL=logs TARGET=test-slice` confirmed the expected comparator assertions.
An initial observer-only `DETAIL=logs LIMIT=8` request was rejected because LIMIT
belongs to critical-path queries; removing LIMIT succeeded. Neither observer
request reran product tests or changed the retained results.

### S15-R4 frozen affected-gap portfolio

`make agent-finalize` passed at `nf-imp07-finalize-20261030-01`, outer exit 0,
with 10 successful cleanup steps. Source was then frozen at
`sha256:e9d96b0779b713b92a9ea8a3ad18789584ce621c254a8056b18d3990268330d8`;
toolchain `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`
match throughout. No compatible successful full warm-check root was available;
retained warm-check maintenance was skipped, without launching a redundant check.

All runs below passed with outer exit 0 and zero failed/blocked cleanup steps.
Their roots are `.cartulary/test-results/nf-imp07-final-<suffix>-20261030-01`.
The work-unit counts are execution units, not test counts.

| Command / selection | Suffix | Passed work units | Successful cleanup steps |
| --- | --- | --- | --- |
| `make test-slice OWNER=platform.openapi` (all six rows) | `openapi` | 5 | 13 |
| `make test-slice OWNER=module.imports ROWS=module.imports.unit.target_neutral_csv_source_admission` | `imports` | 1 | 9 |
| `make openapi-compatibility-check` | `compatibility` | 4 | 13 |
| `make source-check CHECK=generated` | `generated` | 5 | 14 |
| `make source-check CHECK=schema` | `schema` | 3 | 12 |
| `make source-check CHECK=catalog` | `catalog` | 2 | 11 |
| `make lint CHECK=go` | `go` | 6 | 15 |
| `make lint CHECK=backend-boundary` | `backend-boundary` | 4 | 13 |
| `make test-evidence-audit OWNER=platform.openapi EVIDENCE_ROOTS_FILE=/tmp/cartulary-imp07-kmxt5qho/openapi-evidence.json` | `openapi-audit` | 1 | 10 |

Canonical exact-root `make explain-run ... DETAIL=summary` observations confirmed
these outcomes, source identities and cleanup. Both characterization rows and
the corrected focused comparator row executed fresh. The final OpenAPI selection
executed four work units and accepted one valid cached comparator unit; the final
Imports selection accepted its valid cached unit. The audit accepted all six
OpenAPI rows at the frozen source, with no rejected or unused input roots. Imports
is an exact-row qualification, not a complete-owner audit.

The production compatibility report remains exactly equal to the entry report:
600 records, identical classifications, pointers, hashes and fingerprints. The
2.0.0 change set and immutable released baseline remain unchanged; no blanket
reapproval, version increment or migration was needed. Preview runtime source is
unchanged. The source review found only the intended 11 changed/new files relative
to entry, including the Make-generated topology render index. Existing unrelated
tracked/untracked contents and the index remain preserved; executable fingerprints
still match the post-finalize snapshot in `/tmp/cartulary-imp07-kmxt5qho`.

V28's follow-up implementation accepts only string-to-string `change` records at
exact OAuth scope locations in the four defined flows. Tests cover scope identity
changes, reserved/escaped names, empty text, invalid types, unknown flows, nested
non-description values, endpoints, whole schemes, patch-versus-major version
policy and exact review failures. Prior nested operation/security requirements,
uniqueness, determinism and false-positive tests still pass. V27's new independent
1 MiB guard passes for 10,000 and 100,000 rows, and the 10,000-row full-materialization
control exceeds the bound. Full-tail width, malformed quote, terminal integrity
failure and cancellation checks pass alongside the original wide-ragged test.
The bound is fixture-specific allocation evidence, not a universal memory guarantee.

**S15-R4 exit: DONE.** F34/V28 and F35/V27 are closed within this handoff scope.
S15-R4 supersedes only the affected descriptive-classification and retention
acceptance claims. Earlier S15-R3 integration/browser and owner evidence remains
historical at its recorded source. Broader browser/integration/Network Flow/Tabular
qualification was not repeated because IMP-07 changes the comparator and resource
tests, not those runtime surfaces. No CI-wide/current complete Imports or Network
Flow owner claim, publication, commit, database reset or deployment is made.

Final documentation checkpoint: `CARTULARY_TEST_RUN_ID=nf-imp07-final-markdown-20261030-01 make lint-markdown`
passed, outer exit 0. The status-only closure above is followed by the same
command with run ID `nf-imp07-final-markdown-20261030-02`; its terminal result is
retained in that root's `adhoc/lint-markdown/tool-run-summary.json`.
