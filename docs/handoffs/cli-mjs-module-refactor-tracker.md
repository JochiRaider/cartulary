# cli.mjs Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `tools/harness/core/test-output/cli.mjs` |
| Target label | `cli.mjs` |
| Normalized target slug | `cli-mjs` |
| Output path | `docs/handoffs/cli-mjs-module-refactor-tracker.md` |
| Planning-only status | This tracker records inventory, boundaries, risks, and handoff planning only. |
| Allowed changes in this session | Create or update only this tracker file. |
| Non-goals | No production refactor, no test edits, no contract edits, no generated artifact edits, no package configuration edits, no migration edits, and no harness implementation edits. |
| Implementation authorization | Implementation requires a later authorized task. |
| Source hierarchy | 1. Adopted subsystem NLSpecs for their named subsystem. 2. Core 00 through Core 04 for implementation-conformance behavior. 3. Core 05 only for claim-bearing timed or fixture-sensitive publication. 4. Domain vocabulary and implementation-support guides. 5. Current repository code and tests. 6. Prior plans and handoffs as evidence only. |
| Owner contradiction status | No owner contradiction was found in the inspected sources. |
| Repository/framework mismatch | The planning framework module catalog does not make `cli.mjs` a durable module boundary. Live repository inspection shows it is a harness CLI implementation behind `test-output.mjs`, not workbook orchestration or product module ownership. |

Owner documents inspected:

| Document | Posture for this target |
| --- | --- |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Planning doctrine and template only. |
| `docs/testing-harness-nlspec.md` | Adopted owner for harness mechanics: command invocation, output modes, summaries, artifacts, schema validation, failure taxonomy, cleanup, and verification gates. |
| `docs/domain.md` | Vocabulary and boundary reference; confirms harness terms are implementation-support unless promoted by an owner document. |
| `docs/spec/00_document_set_status_and_precedence.md` | Authority, profile, conformance, and owner-section precedence. |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Product modular-monolith boundary and required logical product modules. |
| `docs/spec/02_domain_model_schema_and_history.md` | Record/source/history vocabulary and product source-state boundary. |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Workbook, saved-view, projection, collaboration, and interaction behavior boundaries. |
| `docs/spec/04_security_deployment_and_conformance.md` | Product security, auth, authorization, deployment, and acceptance criteria boundaries. |
| `tools/generated_artifact_policy.json` | Generated-root and generated-file hand-edit prohibition. |

Repository files inspected:

| Path | Inspection purpose |
| --- | --- |
| `tools/harness/core/test-output/cli.mjs` | Target implementation, command surface, responsibilities, imports, and artifact behavior. |
| `tools/harness/core/test-output.mjs` | Entry shim that imports `test-output/cli.mjs` for side-effect execution. |
| `tools/harness/core/test-output.sh` | Node launcher for the test-output helper. |
| `Makefile` | `TEST_OUTPUT_SCRIPT` binding and Make-owned invocation posture. |
| `tools/harness/core/test-output/context.mjs` | Shared schema IDs, result-root/run-ID resolution, coverage buckets, and timing buckets. |
| `tools/harness/core/harness-contract.mjs` | Harness schema validation, result-root, output-mode, config, and secure artifact helpers. |
| `tools/harness/core/tool-output.mjs` | Tool-run summary and public output helper ownership. |
| `docs/handoffs/test-harness-module-refactor-tracker.md` | Prior harness simplification evidence, especially SI-06 audit posture for `core/test-output/cli.mjs`. |
| `tools/schemas/*test*`, `tools/schemas/*summary*`, `tools/schemas/*frontend_row*`, `tools/schemas/*govuln*` | Schema attachment names relevant to emitted artifacts. |
| Harness tests under `tools/harness/{core,scheduler,planning,browser,frontend}/tests` | Current test surfaces that invoke or assert `test-output` behavior. |

Implementation requires a later authorized task. This tracker does not authorize moving code, deleting fallback readers, changing command names, changing output formats, or changing schema-owned artifact shapes.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `tools/harness/core/test-output/cli.mjs` | CLI dispatch and harness output/reporting implementation for phase summaries, target summaries, run summaries, lifecycle lines, timing spans, shared-execution records, failure normalization, test accounting, frontend row accounting references, security scan rollups, and Shell/Go/Vitest/Playwright runner summarization. | No ESM exports. Public surface is command-line behavior via `shell-phase`, `go-phase`, `go-manifest-phase`, `vitest-phase`, `vitest-manifest-phase`, `playwright-phase`, `playwright-manifest-phase`, `go-json-stream`, `target-summary`, `timing-span`, `shared-execution`, `run-summary`, `run-start`, `step-start`, and `target-start`. | `Makefile` `TEST_OUTPUT_SCRIPT`; `tools/harness/core/test-output.mjs`; `tools/harness/core/test-output.sh`; `run-phase-common.sh`; backend, frontend, and browser phase wrappers; browser batch wrappers; scheduler and planning CLIs; static-analysis wrappers; harness tests. | `artifact-discovery.mjs`; `failure-taxonomy.mjs`; `fixture-reporting.mjs`; `harness-contract.mjs`; `tool-output.mjs`; `test-output/context.mjs`; adapters for frontend evidence/accounting, phase manifests, Playwright reports/selection, summary topology, and target planning. | Direct and indirect coverage in `test-harness-contracts.mjs`, `test-run-make-sequence-fast.sh`, `test-cartulary-runner-service-backed-target.sh`, scheduler tests, phase-slice tests, browser Playwright wrapper tests, and frontend Vitest wrapper tests. | Emits or validates `cartulary.tool_run_summary.v3`, `cartulary.test_phase_summary.v3`, `cartulary.test_target_summary.v4`, `cartulary.test_run_summary.v6`, `cartulary.test_shared_execution_group.v1`, `cartulary.test_target_timing.v1`, `cartulary.frontend_row_accounting.v3`, `cartulary.vitest_failure_details.v1`, and `cartulary.govulncheck_findings.v1`. No generated files are in scope for hand edit. | Harness core output/reporting CLI, to be split behind narrower harness-owned command, phase-runner, target/run-summary, timing, accounting, and diagnostics boundaries. | High | The target is a single 7353-line executable module with no public JS exports. Sibling adapter files are dependencies, not in-scope target files. |

Every file in the target path is inventoried above. The target path is a single file, not a directory. Sibling files under `tools/harness/core/test-output/` are outbound dependencies or context files and are out of scope for implementation changes in this tracker task.

## 3. Module Boundary Diagnosis

Architectural finding: `cli.mjs` is not a legitimate permanent product module boundary. It is a harness-owned executable surface that currently concentrates multiple harness reporting concerns. It does not own workbook orchestration behavior for timeline, projections, revisions, collaboration, imports/tabular ingest, entities/indicators, evidence, links, saved views, view contracts, frontend shell state, or grid-adapter integration.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Command dispatch for `test-output.mjs` subcommands | `main()` switch in `cli.mjs` | Harness core test-output CLI facade | Keep | Direct inspection shows a command switch dispatching known helper commands. | Keep as a thin facade only; do not let command dispatch own domain logic. |
| Lifecycle status lines | `handleRunStart`, `handleStepStart`, `handleTargetStart` | Harness output lifecycle adapter | Split | CLI emits `[RUN]`, `[STEP]`, and `[TARGET]` lines. | Behavior is harness mechanics under the testing NLSpec. |
| Phase artifact generation | `writePhaseArtifacts` and phase handlers | Harness phase-summary/reporting module | Split | Target writes `meta.json`, `phase-summary.json`, tool-run summaries, and runner artifacts. | High coupling to runner parsing and failure taxonomy. |
| Shell failure classification | Shell helpers and `handleShellPhase` | Harness diagnostics/failure-classification boundary | Split | Shell path classifies config, infra, security, tool diagnostic, and generic failures. | Preserve failure class/reason and public exit mapping. |
| Go runner parsing and manifest evaluation | Go helper cluster and `handleGoPhase` | Backend harness runner-summary adapter plus phase-summary boundary | Split | Parses Go JSON events, maps manifests, and emits dossiers/inventory. | Do not change Go test selection semantics. |
| Vitest runner parsing and sidecar fallback | Vitest helper cluster and `handleVitestPhase` | Frontend harness runner-summary adapter plus phase-summary boundary | Split | Parses Vitest reports, sidecar failure details, diagnostic tags, and selection filters. | Preserve sidecar preference over `STACK_TRACE_ERROR` fallback. |
| Playwright runner parsing and timing | Playwright helper cluster and `handlePlaywrightPhase` | Browser harness runner-summary adapter plus phase-summary boundary | Split | Parses Playwright reports, selection reports, timing, attachments, and diagnostics. | Browser artifacts are diagnostic/readiness evidence, not product behavior. |
| Target summary collation | `handleTargetSummary` and target summary helpers | Harness target-summary module | Split | Reads phase summaries, child summaries, scheduler summaries, lifecycle spans, frontend accounting, security rollups, browser/service artifacts. | This is the densest mixed-responsibility area. |
| Run summary collation | `handleRunSummary` and run summary helpers | Harness run-summary module | Split | Reads target summaries, groups, helper units, shared execution, scheduler accounting, fixtures, and writes run/tool summaries. | Preserve aggregate output and artifact references. |
| Timing spans and shared execution | Timing/shared helpers | Harness timing/accounting module | Split | Writes timing-span JSON and shared-execution JSON; reads target/service spans. | Timing evidence may be Core 05-relevant only when publication predicates apply. |
| Frontend row accounting references | Target-summary path via frontend row accounting adapter | Frontend harness accounting adapter with core summary integration | Keep / defer | Current code imports adapter and writes `frontend-row-accounting.json` when scope applies. | Do not move frontend product or UI behavior into harness core. |
| Security scan rollups | Govulncheck and gosec helper paths | Harness security diagnostic adapter | Split / defer | Code validates govulncheck findings and maps security failures. | Security rollup is harness evidence, not product authorization behavior. |
| Product workbook behavior | Not found in target | Product modules named by Core 01/Core 03 | Defer | Direct inspection found harness artifact and runner behavior only. | No workbook orchestration move is supported by repository evidence. |

Diagnosis classification:

| Classification | Applies? | Evidence |
| --- | --- | --- |
| Legitimate thin application/service facade | Partially | Applies only to the command dispatch edge. |
| Accidental catch-all | Yes | Many independent harness concerns live in one 7353-line file. |
| View/projection orchestration layer | No | No workbook projection source behavior found. |
| Transport-adjacent adapter | Yes | It is a CLI/process adapter around public Make-owned harness behavior. |
| Persistence-adjacent adapter | No | It writes retained harness artifacts, not product storage. |
| Mutation coordinator | No for product; yes for harness artifacts | It creates and validates retained artifacts under result roots. |
| Frontend shell/controller surface | No | It references frontend test evidence only. |
| Grid-vendor integration layer | No | No direct grid vendor integration found. |
| Misplaced home for logic owned by other modules | Yes for harness subdomains | Runner parsing, summary collation, timing, security, and accounting should have clearer harness ownership. |
| Mixed-responsibility package | Yes | Multiple harness behaviors are concentrated in the CLI file. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Make-owned invocation posture | `docs/testing-harness-nlspec.md` and Make task surface | `Makefile` binds `TEST_OUTPUT_SCRIPT`; NLSpec requires canonical Make invocation. | Public wrapper and sequence tests. | Confirm CLI remains reachable through `test-output.mjs` and `test-output.sh`. | High | Direct raw commands are implementation details unless Make-owned. |
| CLI command names and argument shapes | `cli.mjs` implementation under harness NLSpec | `main()` command switch and usage errors. | Wrapper tests invoke target/run/phase commands. | Add command-specific fixture tests before extracting handlers. | High | Command strings are compatibility-sensitive for wrappers. |
| Output modes and machine JSON | Harness NLSpec Sections 7 and 8 | `tool-output.mjs`, `harness-contract.mjs`, target/run summary writers. | `test-run-make-sequence-fast.sh`, machine output acceptance tests in harness. | Characterize stdout/stderr for success/failure in `summary`, `quiet`, and `machine`. | High | Machine mode must emit exactly one JSON object where accepted. |
| `cartulary.tool_run_summary.v3` | Harness NLSpec Section 8 | `buildToolRunSummary` and `writeToolSummary` paths. | Sequence, frontend, browser, scheduler tests read tool summaries. | Preserve artifact refs, exit code, failure class/reason, and output mode. | High | Tool-run summaries are public retained artifacts. |
| `cartulary.test_phase_summary.v3` | Harness NLSpec Section 8 | `writePhaseArtifacts` validates phase summary. | Static assertion in `test-run-make-sequence-fast.sh`; runner wrapper tests. | Add per-runner fixtures before moving phase handlers. | High | Phase summaries feed target summaries. |
| `cartulary.test_target_summary.v4` | Harness NLSpec Section 8 | `handleTargetSummary` writes target summary. | Scheduler, sequence, frontend, browser, and phase-slice tests. | Characterize aggregate children, skipped children, scheduler failure override, frontend accounting, security rollup. | High | Target summaries are central to aggregate and investigation behavior. |
| `cartulary.test_run_summary.v6` | Harness NLSpec Section 8 | `handleRunSummary` writes run summary. | Sequence and scheduler tests. | Characterize groups, shared execution, helper units, missing targets, and failure public exit. | High | Aggregate public targets rely on this shape. |
| Timing and shared-execution artifacts | Harness NLSpec timing/artifact sections | `timing-span`, `target-timing.json`, and `shared-execution` helpers. | Scheduler and browser wrapper tests include timing assertions. | Preserve lifecycle bucket rollups and shared group failure classification. | Medium | Timing evidence can become publication-sensitive under Core 05 only when claims are active. |
| Test accounting and unmapped failures | Harness NLSpec Section 8 accounting text | `testAccountingUnmappedFailures` and classification manifest reader. | Harness, frontend, browser, and phase-slice tests. | Characterize successful target with unmapped own counts fails with `test_accounting_unmapped`. | High | Do not weaken evidence accounting. |
| Frontend row accounting | Harness NLSpec frontend readiness/accounting sections | `frontendRowAccountingForTarget` adapter and target summary integration. | Frontend and browser wrapper tests. | Preserve scope modes: active target, selected rows, disabled. | High | Harness evidence only, not product conformance unless owner allows. |
| Failure taxonomy and public exit codes | Harness NLSpec Section 9 | `failure-taxonomy.mjs` imports and summary writers. | Product/harness failure classification tests. | Preserve primary failure precedence and failure headline selection. | High | Public consumers read retained summaries for normalized reason codes. |
| Security scan rollup | Harness NLSpec schema/artifact/failure sections | Govulncheck findings validation and gosec/govulncheck classifiers. | Security target tests and summary rollup checks. | Characterize invalid findings artifact and blocking finding behavior. | Medium | Security evidence remains harness diagnostic/security target behavior. |
| Retained artifact paths and investigation commands | Harness NLSpec artifact identity and explain-run requirements | `artifactLine`, `terminalArtifactPath`, `[INVESTIGATE]` command output. | Sequence, scheduler, and tool-output tests. | Preserve relative artifact paths under run root. | High | Retained-run debugging depends on stable artifact references. |
| Product HTTP/WebSocket/workbook contracts | Core 01 through Core 04 | No route handlers, WebSocket handlers, workbook mutations, or product storage writes found in target. | Not applicable to target directly. | No product characterization required for tracker-only task. | Low direct risk | Future harness refactor must not claim product behavior. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `cli.mjs` concentrates unrelated harness reporting concerns. | Direct inspection found CLI dispatch, phase runners, target/run summaries, timing, accounting, security, frontend evidence, and failure diagnostics in one 7353-line file. | High review and regression risk. | `should_fix` | Harness core output/reporting with narrower internal modules. | Plan behavior-preserving extractions after characterization. |
| Command dispatch is the only durable facade-like responsibility. | `main()` switch maps subcommands to handlers. | Medium if command names drift. | `intentional/no_action` | Thin CLI facade. | Keep dispatch stable and move private complexity behind handlers. |
| Runner-specific parsing lives in the core CLI. | Go, Vitest, Playwright, and Shell parser/classifier clusters are in the target. | High because runner reports differ and feed summaries. | `should_fix` | Runner summary adapters under backend/frontend/browser harness ownership. | Split one runner family at a time with fixtures. |
| Target summary collation reads many artifact families directly. | `handleTargetSummary` reads phase summaries, child target summaries, scheduler summaries, timing spans, browser stack artifacts, service metadata, frontend accounting, and security findings. | High because target summaries are public artifacts. | `should_fix` | Harness target-summary module with typed readers. | Characterize target summary variants before extraction. |
| Run summary collation is mixed with target and tool summary helpers. | `handleRunSummary`, `runToolSummary`, group rollups, helper units, and shared execution logic are in the same file. | High for aggregate targets. | `should_fix` | Harness run-summary module. | Extract only after target-summary behavior is pinned. |
| Frontend row accounting is correctly behind an adapter but wired from core summary code. | Imports from `./adapters/frontend-row-accounting.mjs`; target summary writes `frontend-row-accounting.json`. | Medium if core starts owning frontend semantics. | `defer` | Frontend harness accounting adapter plus core summary integration point. | Keep adapter boundary; do not move frontend UI/product behavior into core. |
| Security diagnostic parsing is embedded in shell and target summary paths. | Govulncheck and gosec helpers classify shell failures and roll up findings artifacts. | Medium. | `should_fix` | Harness security diagnostic adapter. | Extract after preserving invalid-artifact and blocking-finding behavior. |
| Legacy retained-run readers and fallback behavior may be public diagnostic support. | Prior `test-harness-module-refactor-tracker.md` SI-06 says audit legacy readers and fallback paths before removal. | High if old retained runs become unreadable. | `must_fix` | Harness diagnostic compatibility policy. | Audit only first; remove nothing without owner evidence. |
| Direct generated artifact edits are prohibited. | `tools/generated_artifact_policy.json` lists generated roots and `tools/task_surface.generated.mk`. | High if generated outputs are hand-edited. | `intentional/no_action` | Generated-artifact owner inputs and Make generators. | Use Make-owned generation/drift only in later authorized work. |
| Product module responsibilities were not found in target. | No route handlers, SQL stores, workbook mutations, WebSocket hub, grid vendor imports, or generated protocol/view contract edits were found. | Low direct product risk, high risk of false claims. | `intentional/no_action` | Product modules remain outside this harness target. | Record as architecture finding; do not plan product moves from this file. |
| Test-only assumptions are part of the runtime helper surface. | CLI consumes phase-map/manifests, runner reports, frontend row accounting, and retained result roots. | Medium if phase identity becomes runtime product architecture. | `should_fix` | Harness evidence accounting modules. | Keep phase maps as evidence accounting only, not runtime product architecture. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Create tracker, record authority posture, normalize `cli-mjs`, and confirm no implementation edits. | `docs/handoffs/cli-mjs-module-refactor-tracker.md` | Tracker review; no broad validation required for tracker-only write. | Tracker Section 1 and session log present. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory current `cli.mjs` responsibilities, inbound callers, outbound dependencies, tests, and contracts. | `tools/harness/core/test-output/cli.mjs`; wrappers and tests for evidence only. | `wc -l`, `rg`, source inspection. | Section 2 populated from live repo. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map every observable harness contract to owner docs and current tests. | Harness NLSpec, schemas, CLI target/run/phase code. | `make help-all` command discovery; later `make harness-contract`. | Section 4 complete. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-06 | Identify missing tests before any code movement. | Existing harness tests; potential new fixture tests in later task. | Later targeted harness tests and `make check-harness-smoke`. | Required characterization rows identified. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05, WF-06 | Classify mixed responsibilities and decide what is keep, split, move, or defer. | `cli.mjs`, adapters, prior test-harness tracker. | Source inspection; later static/lint checks. | Sections 3 and 5 complete. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Design a stable thin CLI facade backed by narrower harness modules. | New or existing harness-internal modules in later task; no edits now. | Later `make lint-scripts`, `make harness-contract`. | Slices S-01 through S-04 ready for implementation authorization. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Define smallest behavior-preserving extraction order. | Tracker now; implementation files later. | Validation command attached to each slice. | Section 7 complete. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Preserve or add characterization tests and accounting checks around any movement. | Harness tests, schemas, phase maps if later owner-authorized. | `make check-harness-smoke`, `make harness-contract`, `make json-shape-check`. | Tests and accounting rows are explicit before implementation. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run the narrowest sufficient Make-owned validation after authorized implementation and update handoff. | Tracker plus touched implementation/test files in later task. | `make agent-finalize`; `make check` when broad verification is warranted. | Handoff log names commands, results, skipped checks, and blockers. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker only. | `docs/handoffs/cli-mjs-module-refactor-tracker.md` | Documentation only; risk is inaccurate inventory. | Preserve no implementation behavior. | Not run in this tracker-only session due the one-file write constraint. | Delete or correct tracker if facts are wrong. | Tracker exists with all required sections. |
| S-01 | S-00 | Add or confirm characterization fixtures for command dispatch, phase summary, target summary, run summary, output mode, and retained artifact behavior before moving code. | Existing harness tests under `tools/harness/*/tests`; no production files unless later authorized. | Test gaps can hide output or schema drift. | Preserve direct tests for `vitest-phase`, target/run summary schema validation, frontend row accounting, browser and scheduler summary paths. | `make check-harness-smoke`; `make harness-contract`; targeted shell tests as discovered. | Revert only characterization additions if they encode wrong current behavior. | Risky behavior has pre-move evidence or explicit `BLOCKED: missing characterization`. |
| S-02 | S-01 | Split phase-summary writing and Shell/Go/Vitest/Playwright runner parsing behind stable command handlers. | `tools/harness/core/test-output/cli.mjs`; likely new harness-internal phase/runner modules. | `phase-summary.json`, runner failure dossiers, inventory counts, manifest mismatch behavior, sidecar fallback. | Add or preserve runner fixtures for Go, Vitest sidecar fallback, Playwright timing, Shell diagnostic classification. | `make lint-scripts`; `make check-harness-smoke`; `make harness-contract`. | Restore moved helper cluster and import paths as one slice. | CLI commands produce byte-compatible schema-valid summaries for covered fixtures. |
| S-03 | S-02 | Split target summary, run summary, timing span, and shared-execution collation behind stable handlers. | `tools/harness/core/test-output/cli.mjs`; likely target/run/timing harness modules. | `target-summary.json`, `run-summary.json`, `tool-run-summary.json`, failure precedence, child aggregation, shared groups, fixture rollups, artifact refs. | Preserve scheduler, sequence, browser aggregate, frontend accounting, and missing/skipped child scenarios. | `make lint-scripts`; `make check-harness-smoke`; `make harness-contract`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`. | Revert extraction module and keep CLI monolith behavior. | Public target and run summaries remain schema-valid and compatible. |
| S-04 | S-03 | Audit legacy readers and fallback paths before removing anything. Removal requires later authorization and owner evidence. | `tools/harness/core/test-output/cli.mjs`; retained-run diagnostic readers; prior harness tracker. | Retained old runs, `explain-run`, historical schema diagnostics, frontend row accounting legacy artifacts. | Add retained-run fixtures before removing any fallback. | `make harness-contract`; `make check-harness-smoke`; retained-run diagnostic target if owner exposes one. | Keep fallback if owner evidence is ambiguous. | Each removed fallback has owner evidence, replacement coverage, and rollback note. |
| S-05 | S-04 | Final validation and handoff after authorized refactor slices. | Tracker plus all implementation/test files touched by later task. | Broad harness behavior and generated drift. | Preserve all characterization added earlier. | `make agent-finalize`; `make check`; add `RESULTS_DIR=<successful full warm check run root>` only when available. | Revert the last behavior-preserving slice that introduced failure. | Handoff log names validation results and remaining risks. |

Any slice that intentionally changes observable behavior, public target output, schema shape, command arguments, failure mapping, or retained artifact paths requires later authorization and an owner-document decision before implementation.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make lint-scripts` | Authored JavaScript orchestration and harness script lint/static checks. | yes for code movement; no for tracker-only write | Discovered from `make help-all`; not run in this tracker-only session. |
| integration | `make check-harness-smoke` | Narrow semantic smoke for harness command surface and scheduler/service-backed smoke behavior. | yes for implementation slices | Discovered from harness tracker and Make help; not run here. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser-backed harness behavior when Playwright/browser summary paths change. | no, conditional | Run only if browser wrapper or Playwright summary behavior changes. |
| generated drift | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make json-shape-check`; `make generated-artifact-policy-check` | Task surface, schemas, JSON shape, generated artifact policy. | yes if owner inputs, schemas, or task surface references change | Do not hand-edit generated outputs. |
| import-boundary/static | `make lint-scripts`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | Harness import/static posture and public target metadata. | yes for extraction slices | Backend/frontend product import-boundary targets are not primary unless later code crosses those boundaries. |
| full check | `make agent-finalize`; `make check` | End-of-run repository verification and finalizer maintenance. | no before implementation; yes before final review of broad refactor | If retained successful run evidence is used, pass `RESULTS_DIR=<successful full warm check run root>`. Otherwise record that retained-run maintenance was skipped because `RESULTS_DIR` was unset. |

Validation was not run in this tracker-only session because the task permitted only one repository write: this tracker file. The commands above were discovered, not executed as validation.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define `cli-mjs` target scope and non-goals. | WF-00 | DONE | none | Section 1 | Target path, label, output path, allowed write, and later-implementation requirement are explicit. |
| T-002 | Inventory current `cli.mjs` repository state. | WF-01 | DONE | T-001 | Section 2 | Target responsibility, callers, dependencies, tests, schemas, owner candidate, and risk are listed. |
| T-003 | Record module boundary diagnosis. | WF-04 | DONE | T-002 | Section 3 | `cli.mjs` classified as mixed harness CLI, not workbook/product module boundary. |
| T-004 | Map public contracts and behavior freeze points. | WF-02 | DONE | T-002 | Section 4 | Observable harness contracts have owners, evidence, tests, characterization gaps, and risk levels. |
| T-005 | Classify coupling and boundary findings. | WF-04 | DONE | T-002 | Section 5 | Findings use `must_fix`, `should_fix`, `defer`, or `intentional/no_action`. |
| T-006 | Plan characterization test gap closure. | WF-03 | TODO | T-004 | S-01 | Later task confirms or adds characterization before moving code. |
| T-007 | Plan behavior-preserving extraction sequence. | WF-05, WF-06 | DONE | T-003, T-004 | Section 7 | Slices S-00 through S-05 are ordered with validation and rollback notes. |
| T-008 | Plan harness/accounting validation. | WF-07 | DONE | T-004 | Section 8 | Make-owned validation commands are named or marked conditional. |
| T-009 | Defer implementation until authorized. | WF-08 | DEFERRED | T-006, T-007, T-008 | Section 10 | Next actor has tracker, risks, and validation plan without code changes. |
| T-010 | Audit legacy readers/fallback paths. | WF-04, WF-05 | TODO | T-006 | S-04 | Each fallback has owner evidence before any removal. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | Tracker created for `cli-mjs`; planning-only scope recorded. | Touched: `docs/handoffs/cli-mjs-module-refactor-tracker.md`. Inspected: framework, harness NLSpec, domain doc, Core 00-04, generated artifact policy. | `sed`; `git status --short --branch`; `date -u`. | Authority posture recorded; no owner contradiction found in inspected sources. | Implementation requires later authorized task. | Start S-01 characterization review before any code movement. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | No backend product module ownership found in target. Backend harness runner parsing for Go is present. | Touched: tracker. Inspected: `cli.mjs`, backend wrapper references from search results. | `rg`; `sed`. | Go runner parsing belongs to harness runner-summary planning, not product backend module behavior. | Need characterization before extracting Go parsing. | In S-01/S-02, preserve Go event parsing, manifest selection, package failure classification, and reproduce commands. |

### Frontend module boundary, if applicable

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | No frontend shell/controller or grid-vendor ownership found. Frontend row accounting and Vitest/Playwright evidence handling are harness concerns. | Touched: tracker. Inspected: `cli.mjs`, frontend/browser test references. | `rg`; `sed`; `find`. | Frontend behavior remains outside this target except harness evidence accounting and runner report parsing. | Need row-accounting and sidecar fallback fixtures before extraction. | Preserve frontend row accounting scope behavior and Vitest failure sidecar preference. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | Schema-owned harness artifacts mapped; no generated file edit planned. | Touched: tracker. Inspected: schema filenames, `tools/generated_artifact_policy.json`, harness NLSpec Section 8. | `ls`; `rg`; `cat`; `sed`. | Contracts to freeze include tool, phase, target, run, timing, shared, frontend row accounting, Vitest sidecar, and govulncheck findings artifacts. | Generated drift commands not run due tracker-only write constraint. | Run drift/schema checks only in later authorized implementation or validation task. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | Existing tests touch CLI directly and indirectly, but characterization is uneven across command families. | Touched: tracker. Inspected: harness test filenames and selected test snippets. | `rg`; `find`; `sed`; `make help`; `make help-all`. | Validation commands discovered; no validation success claimed. | Need targeted characterization before movement. | Use S-01 to confirm current coverage and add fixtures where gaps block safe extraction. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | Product authorization behavior is not in target. Harness security diagnostics for govulncheck/gosec are present. | Touched: tracker. Inspected: Core 04, `cli.mjs`, harness NLSpec security/failure sections. | `sed`; `rg`. | Security scan rollup is harness evidence, not product auth. | Need characterization before moving security diagnostic helpers. | Preserve invalid artifact and blocking finding classification in S-02/S-03. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T23:02:00Z | Codex tracker creation | Tracker is current enough for another agent to start characterization planning without rediscovery. | Touched: tracker only. | `test -f`; `wc -l`; `git status --short --branch`; `date -u`; plus discovery commands listed in Section 1. | Tracker created; no production refactor performed. | Later implementation authorization required. | Start S-01, then update this log before and after any implementation slice. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Later implementation authorization is required before editing harness code or tests. | This task permits only tracker creation/update. | User authorization for an implementation task. | Open; not blocking tracker completion. |
| RB-002 | Characterization coverage for every `cli.mjs` command family is not fully proven by current inspection. | Moving code without command-level fixtures risks changing public output, schema artifacts, or exit codes. | S-01 test inventory and targeted fixture review. | TODO. |
| RB-003 | Legacy retained-run readers and fallback paths must not be removed without owner evidence. | Retained-run diagnostics may be public harness support even when code appears obsolete. | Harness NLSpec evidence and retained-run fixture coverage. | TODO. |
| RB-004 | Generated task-surface or schema owner inputs may be affected by future extraction only if command metadata changes. | Generated files must not be hand-edited and task surface drift must remain controlled. | Later implementation diff review and Make-owned drift checks. | TODO if future slice changes owner inputs; otherwise not applicable. |

## 12. Binary Completion Criteria

| Criterion | Current tracker status |
| --- | --- |
| Every file in `tools/harness/core/test-output/cli.mjs` is inventoried or explicitly out of scope. | Passed: the target path is a single file and is inventoried in Section 2. |
| Every discovered public contract risk has an owner and test posture. | Passed for planning: Section 4 maps discovered risks to owners, current tests, and required characterization. |
| Every proposed workflow has dependencies and exit criteria. | Passed: Section 6 defines workflow dependencies and handoff checkpoints. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | Passed: Section 7 marks behavior changes as requiring later authorization. |
| Validation commands are discovered or marked `TODO` with a reason. | Passed: Section 8 lists Make-owned commands and notes they were not run in this tracker-only session. |
| Contradictions are marked `BLOCKED: owner contradiction`. | Passed: no owner contradiction was found in inspected sources. |
| Repository/framework mismatches are recorded as planning findings. | Passed: Section 1 and Section 3 record that `cli.mjs` is not a durable product module boundary. |
| Handoff sections are current enough for another agent to continue without rediscovery. | Passed for tracker scope: Section 10 records current session state, files, commands, blockers, and next action. |

