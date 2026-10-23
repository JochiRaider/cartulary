# Testing Harness Remediation 03 — Coordinated Execution Efficiency

This is the controlling execution tracker for the user-approved Remediation 03
plan. The implementation request supersedes the earlier planning-only action
boundary; historical receipts E-001–E-017 retain their original scope.

Document date: 2026-10-21, America/New_York. Original execution through E-073: **COMPLETE**.
Previous continuation execution: **COMPLETE — E-166/E-167**. WF-09 diagnostics are accepted. E-149–E-157 complete Section 7.7 stability qualification, the common test repair and experiment 10 freeze. E-160/E-161 reject the lifecycle optimization: browser improves 1.57% / 1.33%; CI improves 1.55% / regresses 2.65% against the user-assumed 60-minute reference. Both workloads fail the first-two-observation 5% gate. WF-11 / S-011 is complete with coordinated restoration of the lifecycle owner, implementation and projections. Accepted diagnostics and common stability repairs remain. E-162 preserves a test-only latency assertion failure; E-163 completes corrective WF-07f. E-164 preserves the required-accounting UI-review fixture failure; E-165 completes corrective WF-07g, including explicit suites and independent evidence audits. E-166 records passing final CI, independent audit, settled cleanup and retained scans. WF-07e / S-007e final validation and handoff are complete under E-167, subject to its registered closing lint succeeding after this final update. Historical failures and the completed E-073 handoff retain their original meaning.
The continuation follows **WF-00e → WF-09 → WF-10 → WF-11 → WF-07e**.
Each boundary requires a tracker update and successful `make lint-markdown`
before the next workstream begins. Original receipts remain historical.
Completed workstreams: **WF-00, WF-06, WF-08, corrective WF-08a, WF-01–WF-05,
and WF-07 including corrective WF-07a–d**.
E-049 accepts S-001 browser v15 with separate assurance and performance
decisions under the E-048 four-pair/20% protocol amendment. E-053 adds the
user-required below-5% early rejection. E-054 rejects incident fixture removal;
E-055 accepts its required Location assertions and independent routing.
**WF-07 / S-007 validation and handoff complete, E-072/E-073.**
E-070 completes the template-preparation and native build-cache repairs.
E-059 retains dedicated PostgreSQL lifetimes; no pooling contract was adopted.
E-058 retains stronger isolation without admitting a conversion. E-057
retains current cost/admission policy with explicit calibration limits. Execute serially: **WF-00 → WF-06 → WF-08 → WF-01 → WF-02 → WF-03 → WF-04 →
WF-05 → WF-07**. Updating this tracker after each workstream and before starting
the next was mandatory. S-007 supplies the final validation and handoff boundary.

The user authorized conditional implementation of all five investigated slices.
S-001/S-002 require assurance and the applicable Section 8.2 performance gate.
When full performance validation is required, S-001/S-002 allow at most 5%
regression, S-003/S-004 require greater than 5% improvement, and S-005 requires
greater than 10%; all five guardrails allow at most 5% regression. E-048 adds
the user-authorized early exit for performance analysis only; E-053 adds rejection
when both first-pair improvements are strictly below 5%. Correctness,
cleanup, evidence integrity and S-007 final validation remain required.
Specification, documentation and lint repairs require correctness, not speedup.
Existing behavior is retained only when its contract and evidence justify it.

Reservation/retention continuation: **IN_PROGRESS — WF-16a repair passed; full CI remains unverified, E-199**.
The user authorized Section 7.8 implementation on 2026-10-23. Execution is
WF-00f → WF-12 → WF-13 → WF-14 → WF-15 → WF-16. Each boundary requires a tracker
receipt and passing `make lint-markdown`. Lifecycle correctness and optional
Go-budget performance acceptance remain separate. No speedup is established.

## 1. Scope and Source Posture

### 1.1 Target, output and authorized actions

The canonical target is repository `.` at `/home/jochi/code/cartulary`.
This file remains the sole controlling tracker. Preserve its twelve sections and
stable IDs; implementation changes may edit the adopted owner, authored machine
inputs, code and tests named by the approved slices. Generated artifacts are
refreshed only through Make. Product behavior and presentation are unchanged.

The creation/review sessions admitted only tracker writes. Those restrictions
are historical, not limits on this explicitly authorized implementation task.
The adopted NLSpec remains the harness behavior owner. Human projection-fidelity
review remains distinct from automated checks. No execution tool consumes this
tracker or other Markdown as behavior, configuration or acceptance evidence.

### 1.2 Authority and normative vocabulary

| Statement class | Meaning and authority | Required interpretation |
| --- | --- | --- |
| Adopted requirement | Named clause in the adopted/current [Testing Harness NLSpec](../testing-harness-nlspec.md), or a product primary owner allocated by [Core 00](../spec/00_document_set_status_and_precedence.md) | The source clause governs. This handoff is a subordinate reference, not a new owner. |
| Proposed requirement | R03-001 through R03-041 in Section 7 | MUST/MUST NOT describe conditions of a later authorized change. A proposal cannot override an adopted requirement before adoption. |
| Handoff procedure | Source qualification, entry gates, evidence and completion rules in this document | MUST/MUST NOT constrain this roadmap and its future acceptance records; they do not authorize execution. |
| Observed behavior | Directly inspected implementation or parsed metadata with an E-ID | Describes the inspected baseline; neither requirements completeness nor successful execution follows. |
| Rationale / hypothesis | Explanation of benefit or risk | Is not normative; an unmeasured opportunity cannot become a measured result by repetition. |
| Historical result | A named receipt in an earlier handoff | Is attributed to that document and source identity; it is not independently verified here or a current-source pass. |

This handoff does not itself amend an adopted NLSpec, establish product
conformance, or activate Core 05 publication. Execution authorization comes from
the user’s explicit implementation request; behavioral changes are adopted in
the NLSpec before their projections change. Tests, generators, runtime metadata, conformance and release
evidence MUST NOT read, stat, hash or otherwise depend on this document or other
Markdown. Static authoring inspection of documents is not executable evidence.

Core 00 owns allocation. Core 01 REQ-01-157, 160, 162–167 owns the incident-create
normalization, atomicity, replay and response behavior relevant to S-002.
Core 04 remains the security/conformance owner; Core 05 is publication-only.
No production behavior or presentation redesign is proposed. A newly discovered
conflict MUST be resolved through its primary owner, not preserved as an
accidental compatibility contract.

### 1.3 Baseline and preservation boundary

The continuation starts from a clean checkout at
`f42fc00b76ec4e296c2a379c09687e02ce669de7`, revalidated under E-074.
Historical `r03-s007-ci-03` motivates diagnosis; it is neither a current baseline
nor an acceptance sample. Accepted diagnostics and original separate resets
will define the new baseline. Candidate snapshots must differ only by the
registered lifecycle treatment and its owner/projection changes.

E-018 is the original execution baseline: HEAD
`21e9c4c15a67c69a7a97ce559653f5afadd964f4`, branch
`codex/specification-remediation-20261016`; only this tracker was already modified.
No staged or untracked files were present. Its admitted SHA-256 was
`00bc1cf368bbb9e6a91228089166e28d97ee0fdf7d7ff7e43b0b1c5ac761e9ee`;
logical-index SHA-256 was
`69852f4aff424555e80020ce6f89f551d407acec1cf1699761a70aa1b570d2b8`.
These are authoring records, not executable source digests. Existing tracker
edits are incorporated rather than reset. E-013–E-017 below are historical.

| Review baseline item | Starting identity | Closure / limit |
| --- | --- | --- |
| Repository / branch / HEAD | `/home/jochi/code/cartulary`; `codex/specification-remediation-20261016`; `21e9c4c15a67c69a7a97ce559653f5afadd964f4` | E-017; this commit alone does not certify a retained run |
| Git status / logical index | Clean; index SHA-256 `69852f4aff424555e80020ce6f89f551d407acec1cf1699761a70aa1b570d2b8` | No staged or untracked changes at review admission |
| Existing handoff | SHA-256 `abc4747081be6e8b6ceb6ffd01e094f971b355a2b8fb30bcce3d0e979ed457c6`; regular, one hard link, admitted target identity unchanged | Recheck before writing; this is an update, not a new file |
| Bounded preservation set | 124 referenced inputs plus this handoff; content and filesystem identities retained in memory | No complete original input snapshot survives from E-013; do not claim all old observations revalidated by hash |
| Exclusions | Ignored outputs, active processes and remaining repository contents | No runtime observation or whole-workspace content certification |

Historical creation baseline:

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Repository, branch and HEAD | `/home/jochi/code/cartulary`; `codex/specification-remediation-20261016`; `8bcece3e7478ec180e1925a12fb6bb9ba847db79` | See E-014 final receipt | HEAD alone does not describe the dirty working tree. |
| Logical index | SHA-256 of `git --no-optional-locks ls-files --stage -z`: `48980508ae5a424ee1cdbd7695aa565057982c138f50d656cc9f70dc5bd67a6d`; no staged changes reported | See E-014 | Optional index writes were disabled. |
| Pre-existing unstaged changes | `docs/handoffs/testing-harness-remediation-02.md`; `docs/testing-harness-nlspec.md`; `internal/platform/jobs/runner_supervision_test.go`; `internal/testutil/harnessfixtures/recovery_linux.go`; `internal/testutil/harnessfixtures/recovery_linux_test.go` | Content hashes, file identities and index compared under E-014 | These changes belong to the existing workspace; none is authorized for editing here. |
| Non-ignored untracked files | None at admission | Expected new path: this handoff only; E-014 records actual result | Status alone cannot prove unchanged contents of already-dirty files. |
| Bounded content snapshot | 124 input files: SHA-256, device/inode, mode, link count and size retained in session memory | Compared before creation and at closure; E-014 | The snapshot is not a new repository artifact or executable source fingerprint. |
| Output identity | Absent; existing ancestors are nonsymlink directories | New regular file; E-014 | No existing tracker was replaced. |
| Ignored outputs, remaining repository contents and active processes | Not traversed, hashed or operated | No command directed at them | Whole-workspace content preservation across concurrent external activity is not independently verified. |

The final write-boundary claim MUST remain bounded to the actions performed and
the compared input set. An unrelated concurrent change MUST be recorded without
attribution or automatic reversion.

### 1.4 Relationship to earlier handoffs

[Remediation 01](testing-harness-remediation-01.md) records earlier structural
work and exclusions. [Remediation 02](testing-harness-remediation-02.md) records
later runner grouping, incremental evidence, scenario simplification and
lifecycle repair. Neither is a standing behavior owner.

At E-015, remediation 02's opening executive summary and earlier S-008 table
still describe blocked aggregate acceptance. Its later active terminal receipt
now reports successful validation; the opening posture is stale for that later
receipt. Neither posture is an independently certified result of this review. A receipt appearing later in that file, now
explicitly inside its preserved Appendix A,
[reports completed-source acceptance](testing-harness-remediation-02.md#completed-source-baseline-and-aggregate-acceptance):
`remediation02-aggregate-ci`, native session 71887 / exit 0,
741/741 work units and 1,398 completed cleanup steps, bound to reported executable
snapshot `sha256:54b53ac3f86228bdc8967604ce0d103488e5c3ae4196f65b690bc44ae7ecec3a`.
Its
[unresolved causal obligations](testing-harness-remediation-02.md#unresolved-causal-obligations)
retain W1 first-drop, W2 readiness and W4 process-identity uncertainty.
Its
[explicit validation receipt](testing-harness-remediation-02.md#explicit-validation-phases-and-final-handoff)
reports additional successful selections.

The apparent summary/receipt conflict MUST be qualified by source and session,
not resolved by document position. The appendix's successful receipt establishes
only what its author reported for that historical snapshot. It does not supersede
the active continuation's later receipts. During the creation session,
remediation 02 acquired additional active receipts. Its
[full-7 terminal receipt](testing-harness-remediation-02.md#d-003-full-7-terminal-receipt-and-next-repairs)
reports native 68130 / exit 2, 58 passed and four failed of 62 units, at source
`sha256:4dd8102743aac96c39c3138f7ddf4adb15da5928e1e5ada3d67eee4241452322`;
the document assigns subsequent repairs and validation to that continuation.
That failed attempt remains historical evidence. The subsequent
[D-003 terminal validation and handoff](testing-harness-remediation-02.md#d-003-terminal-validation-and-handoff)
reports `d003-full-8`, native 40233 / exit 0, 62/62 parent units and 2,102
completed cleanup-ledger entries, bound by its
[renewed checkpoint](testing-harness-remediation-02.md#d-003-renewed-full-validation-checkpoint)
to source `sha256:69eae72ad683b2a21886644f61a014a839893fc3daf0be9c5f2c571784c792ca`.
It separately leaves two legacy pending-startup authorities unresolved. Those
ledger entries are not unique physical resources, and that successful execution
does not settle the legacy authorities. E-016 records only what these receipts
report; no retained artifacts were opened or independently certified here.

The active [completion scope](testing-harness-remediation-02.md#current-completion-scope)
also says historical PostgreSQL causal attribution is dropped from that
continuation's acceptance. This changes its scope, not the factual status of
unproved causes. The original causal questions and closure criteria remain
recorded in the appendix; this handoff neither resolves them nor reactivates
them as mandatory work. W4's absent historical comparison likewise remains
unproved even where deterministic lifecycle regressions are reported passing.

E-012 and E-016 record these distinct source/session postures. No retained run, process,
private proof or diagnostic bundle was opened here. This document MUST NOT
independently certify any reported result, turn the appendix into current-source
acceptance, edit the earlier summary, or impose a blanket historical blocker on
independently safe planning. Later implementation requires a fresh source and
acceptance recheck under RB-001.

## 2. Current-State Repository Inventory

### 2.1 Declared inspection universe

The creation audit's bounded input universe contains the 124 tracked files explicitly
listed below. Of these, 46 received direct inspection of relevant definitions,
scenarios or guidance; 78 received metadata-only accounting. Relevant reads
span this conversation and the final source recheck; content was compared against the writing-session snapshot. Concurrent changes
to the Makefile, pin registry, harness specification and remediation 02 were
reinspected and qualified under E-014. Direct inspection is not a
claim that every line or transitive effect was audited.

The current review revalidated the catalog census, packing model, incident
assertions, relevant grouping/adoption clauses, public selection inputs and lint
exclusion (E-016). Other original observations remain attributed to E-001–E-014;
this review is not a new complete implementation audit.

Metadata includes 66 authored family catalogs; generated browser/Make
projections are identified by their owner and used as data only. The three
concurrent implementation files receive preservation accounting only. Every
listed file was a regular file in the snapshot; external symlinks and nested
repositories were not traversed.

| Path / kind | Disposition and depth | Responsibility / surface | Inbound callers / discovery limits | Outbound dependencies | Test posture | Contracts / evidence |
| --- | --- | --- | --- | --- | --- | --- |
| guidance (12 files): [AGENTS.md](../../AGENTS.md); [tools/AGENTS.md](../../tools/AGENTS.md); [internal/AGENTS.md](../../internal/AGENTS.md); [.agents/skills/cartulary-test-ops/SKILL.md](../../.agents/skills/cartulary-test-ops/SKILL.md); [.agents/skills/cartulary-test-ops/references/command-recipes.md](../../.agents/skills/cartulary-test-ops/references/command-recipes.md); [docs/handoffs/cartulary_modular_refactor_planning_framework.md](../../docs/handoffs/cartulary_modular_refactor_planning_framework.md); [docs/spec/00_document_set_status_and_precedence.md](../../docs/spec/00_document_set_status_and_precedence.md); [docs/spec/01_architecture_storage_and_view_contracts.md](../../docs/spec/01_architecture_storage_and_view_contracts.md); [docs/domain.md](../../docs/domain.md); [docs/testing-harness-nlspec.md](../../docs/testing-harness-nlspec.md); [docs/handoffs/testing-harness-remediation-01.md](../../docs/handoffs/testing-harness-remediation-01.md); [docs/handoffs/testing-harness-remediation-02.md](../../docs/handoffs/testing-harness-remediation-02.md) | directly inspected, relevant sections | Authority, planning doctrine, prior receipts and permitted operations; documentation is not an executable input. | Product/harness owners and author; historical receipts not independently audited. | Owner clauses and source evidence | No checks executed | C-001–C-008; E-001/E-012 |
| planning (8 files): [tools/harness/scheduler/work-graph/compiler.mjs](../../tools/harness/scheduler/work-graph/compiler.mjs); [tools/harness/scheduler/work-graph/cohorts.mjs](../../tools/harness/scheduler/work-graph/cohorts.mjs); [tools/harness/backend/go-lpt-shards.mjs](../../tools/harness/backend/go-lpt-shards.mjs); [tools/harness/browser/browser-cohorts.mjs](../../tools/harness/browser/browser-cohorts.mjs); [tools/harness/scheduler/work-graph/browser.mjs](../../tools/harness/scheduler/work-graph/browser.mjs); [tools/harness/scheduler/work-graph/browser-functional-lanes.mjs](../../tools/harness/scheduler/work-graph/browser-functional-lanes.mjs); [tools/harness/generated-artifacts/execution-topology.mjs](../../tools/harness/generated-artifacts/execution-topology.mjs); [tools/harness/browser/browser-batch-manifest.mjs](../../tools/harness/browser/browser-batch-manifest.mjs) | directly inspected, relevant definitions | Catalog selection, compatibility, deterministic packing, resource claims and graph construction. | Task compiler → runner/broker; complete migration-consumer audit remains RB-002. | Catalogs, topology and invocation policy | Grouping/lifecycle test locations are a future discovery boundary | C-001/C-002/C-006; E-003/E-004 |
| lifecycles (12 files): [internal/testutil/harnessfixtures/services_linux.go](../../internal/testutil/harnessfixtures/services_linux.go); [internal/testutil/harnessfixtures/renderer_linux.go](../../internal/testutil/harnessfixtures/renderer_linux.go); [internal/testutil/pgtest/pgtest.go](../../internal/testutil/pgtest/pgtest.go); [internal/testutil/postgrescatalog/clone.go](../../internal/testutil/postgrescatalog/clone.go); [internal/testutil/appsupport/runtime.go](../../internal/testutil/appsupport/runtime.go); [internal/testutil/appsupport/store.go](../../internal/testutil/appsupport/store.go); [internal/testutil/s3test/s3test.go](../../internal/testutil/s3test/s3test.go); [internal/platform/harnessruntime/reset.go](../../internal/platform/harnessruntime/reset.go); [tools/harness/browser/renderer-quiescence.mjs](../../tools/harness/browser/renderer-quiescence.mjs); [tools/harness/performance-fixture/snapshot-builder-cli.mjs](../../tools/harness/performance-fixture/snapshot-builder-cli.mjs); [tools/harness/execution/vitest-invocation-cli.mjs](../../tools/harness/execution/vitest-invocation-cli.mjs); [tools/harness/execution/runners/vitest-worker-lifecycle.mjs](../../tools/harness/execution/runners/vitest-worker-lifecycle.mjs) | directly inspected, relevant definitions | Physical ownership, database/namespace isolation, renderer settlement, resets, snapshot reuse and Vitest worker settlement. | Native broker and runner consumers; no live service observation. | Published attachments, generation identity, cleanup and private protocols | Fixture/failure characterizations are future V-003/V-005 work | C-003–C-006; E-005–E-008/E-010 |
| scenarios (10 files): [apps/web/e2e/fixtures.ts](../../apps/web/e2e/fixtures.ts); [apps/web/e2e/global-setup.ts](../../apps/web/e2e/global-setup.ts); [apps/web/e2e/global-teardown.ts](../../apps/web/e2e/global-teardown.ts); [apps/web/playwright.shared.config.ts](../../apps/web/playwright.shared.config.ts); [apps/web/vite.config.ts](../../apps/web/vite.config.ts); [apps/web/e2e/incident-administration.spec.ts](../../apps/web/e2e/incident-administration.spec.ts); [apps/web/e2e/auth-and-incident-directory.spec.ts](../../apps/web/e2e/auth-and-incident-directory.spec.ts); [apps/web/e2e/support/incidents/fixtures.ts](../../apps/web/e2e/support/incidents/fixtures.ts); [internal/modules/incidents/integration_test.go](../../internal/modules/incidents/integration_test.go); [internal/modules/incidents/http_conformance_test.go](../../internal/modules/incidents/http_conformance_test.go) | directly inspected, named fixtures and scenarios | Browser actor lifecycle and runner configuration; incident creation, replay and supporting setup. | Selected test scenarios and fixture callers; full cross-owner effects remain RB-003. | Production HTTP/server composition, test contexts and storage | Named source assertions inspected; no execution | C-002/C-007; E-004/E-009 |
| metadata (9 files): [tools/harness_work_graph_owner.json](../../tools/harness_work_graph_owner.json); [tools/execution_topology_manifest.json](../../tools/execution_topology_manifest.json); [tools/browser_e2e_batch_manifest.json](../../tools/browser_e2e_batch_manifest.json); [tools/postgres_fixture_policy_registry.json](../../tools/postgres_fixture_policy_registry.json); [tools/schemas/cartulary.browser_e2e_batch_manifest.v14.schema.json](../../tools/schemas/cartulary.browser_e2e_batch_manifest.v14.schema.json); [tools/schemas/cartulary.harness_performance_explanation.v5.schema.json](../../tools/schemas/cartulary.harness_performance_explanation.v5.schema.json); [tools/task_surface_owner.json](../../tools/task_surface_owner.json); [tools/task_surface.generated.mk](../../tools/task_surface.generated.mk); [tools/toolchain_pins.json](../../tools/toolchain_pins.json) | metadata-only | Authored policy/schema/task inputs plus generated browser and Make projections. | Compiler, adapters and public Make bindings; consumers not exhaustively traced. | Adopted harness owner; generated outputs cannot redefine it | Presence is not validation | C-001/C-006/C-008; E-002/E-004/E-011 |
| doc_validation (4 files): [Makefile](../../Makefile); [.markdownlint-cli2.jsonc](../../.markdownlint-cli2.jsonc); [tools/harness/static-analysis/markdownlint.sh](../../tools/harness/static-analysis/markdownlint.sh); [tools/harness/static-analysis/markdownlint-runner.sh](../../tools/harness/static-analysis/markdownlint-runner.sh) | directly inspected, invocation/selection paths | Make prerequisite path, Markdown selection, non-fixing linter and cache/report writes. | Public lint-markdown; no target-local input exists to select this document. | Node/frontend readiness and retained/cache outputs | Canonical lint not executed and does not select this file | C-008; E-011 |
| catalogs (66 files): [tools/test_families/app.migrate.json](../../tools/test_families/app.migrate.json); [tools/test_families/app.operator.json](../../tools/test_families/app.operator.json); [tools/test_families/app.server.json](../../tools/test_families/app.server.json); [tools/test_families/harness.browser.json](../../tools/test_families/harness.browser.json); [tools/test_families/harness.command_surface.json](../../tools/test_families/harness.command_surface.json); [tools/test_families/harness.evidence_accounting.json](../../tools/test_families/harness.evidence_accounting.json); [tools/test_families/harness.generated_artifacts.json](../../tools/test_families/harness.generated_artifacts.json); [tools/test_families/harness.release.json](../../tools/test_families/harness.release.json); [tools/test_families/harness.test_catalog.json](../../tools/test_families/harness.test_catalog.json); [tools/test_families/module.artifacts.json](../../tools/test_families/module.artifacts.json); [tools/test_families/module.assessments.json](../../tools/test_families/module.assessments.json); [tools/test_families/module.auth.json](../../tools/test_families/module.auth.json); [tools/test_families/module.collaboration.json](../../tools/test_families/module.collaboration.json); [tools/test_families/module.crossownertransaction.json](../../tools/test_families/module.crossownertransaction.json); [tools/test_families/module.database_migrations.json](../../tools/test_families/module.database_migrations.json); [tools/test_families/module.entities.json](../../tools/test_families/module.entities.json); [tools/test_families/module.evidence.json](../../tools/test_families/module.evidence.json); [tools/test_families/module.extensions.json](../../tools/test_families/module.extensions.json); [tools/test_families/module.graphprojection.json](../../tools/test_families/module.graphprojection.json); [tools/test_families/module.imports.json](../../tools/test_families/module.imports.json); [tools/test_families/module.incidentbundles.json](../../tools/test_families/module.incidentbundles.json); [tools/test_families/module.incidents.json](../../tools/test_families/module.incidents.json); [tools/test_families/module.indicators.json](../../tools/test_families/module.indicators.json); [tools/test_families/module.jobapi.json](../../tools/test_families/module.jobapi.json); [tools/test_families/module.links.json](../../tools/test_families/module.links.json); [tools/test_families/module.networkflow.json](../../tools/test_families/module.networkflow.json); [tools/test_families/module.parties.json](../../tools/test_families/module.parties.json); [tools/test_families/module.projections.json](../../tools/test_families/module.projections.json); [tools/test_families/module.records.json](../../tools/test_families/module.records.json); [tools/test_families/module.recovery.json](../../tools/test_families/module.recovery.json); [tools/test_families/module.reference_data.json](../../tools/test_families/module.reference_data.json); [tools/test_families/module.reportcomposition.json](../../tools/test_families/module.reportcomposition.json); [tools/test_families/module.reporting.json](../../tools/test_families/module.reporting.json); [tools/test_families/module.revisions.json](../../tools/test_families/module.revisions.json); [tools/test_families/module.savedviews.json](../../tools/test_families/module.savedviews.json); [tools/test_families/module.stagedobjects.json](../../tools/test_families/module.stagedobjects.json); [tools/test_families/module.tabularingest.json](../../tools/test_families/module.tabularingest.json); [tools/test_families/module.tasksdecisions.json](../../tools/test_families/module.tasksdecisions.json); [tools/test_families/module.timeline.json](../../tools/test_families/module.timeline.json); [tools/test_families/module.workbook.json](../../tools/test_families/module.workbook.json); [tools/test_families/package.grid_adapter.json](../../tools/test_families/package.grid_adapter.json); [tools/test_families/package.protocol_ts.json](../../tools/test_families/package.protocol_ts.json); [tools/test_families/package.test_utils.json](../../tools/test_families/package.test_utils.json); [tools/test_families/package.ui.json](../../tools/test_families/package.ui.json); [tools/test_families/package.view_contracts.json](../../tools/test_families/package.view_contracts.json); [tools/test_families/platform.audit.json](../../tools/test_families/platform.audit.json); [tools/test_families/platform.bootstrap.json](../../tools/test_families/platform.bootstrap.json); [tools/test_families/platform.config.json](../../tools/test_families/platform.config.json); [tools/test_families/platform.cryptography.json](../../tools/test_families/platform.cryptography.json); [tools/test_families/platform.httpruntime.json](../../tools/test_families/platform.httpruntime.json); [tools/test_families/platform.jobs.json](../../tools/test_families/platform.jobs.json); [tools/test_families/platform.objectstore.json](../../tools/test_families/platform.objectstore.json); [tools/test_families/platform.openapi.json](../../tools/test_families/platform.openapi.json); [tools/test_families/platform.pagination.json](../../tools/test_families/platform.pagination.json); [tools/test_families/platform.postgres.json](../../tools/test_families/platform.postgres.json); [tools/test_families/platform.rootedfs.json](../../tools/test_families/platform.rootedfs.json); [tools/test_families/platform.telemetry.json](../../tools/test_families/platform.telemetry.json); [tools/test_families/platform.viewquery.json](../../tools/test_families/platform.viewquery.json); [tools/test_families/platform.viewschema.json](../../tools/test_families/platform.viewschema.json); [tools/test_families/platform.ws.json](../../tools/test_families/platform.ws.json); [tools/test_families/web.application.json](../../tools/test_families/web.application.json); [tools/test_families/web.architecture.json](../../tools/test_families/web.architecture.json); [tools/test_families/web.collaboration.json](../../tools/test_families/web.collaboration.json); [tools/test_families/web.design.json](../../tools/test_families/web.design.json); [tools/test_families/web.networkflow.json](../../tools/test_families/web.networkflow.json); [tools/test_families/web.workbook.json](../../tools/test_families/web.workbook.json) | metadata-only | 66 authored family files: selected symbols/titles, ownership, scopes, resources and fixture capabilities. | Catalog compiler; literal identity comparison does not resolve semantic equivalence. | Verification attribution and selector source | 2,367 active rows; no tests executed | C-001/C-007; E-002 |
| concurrent (3 files): [internal/platform/jobs/runner_supervision_test.go](../../internal/platform/jobs/runner_supervision_test.go); [internal/testutil/harnessfixtures/recovery_linux.go](../../internal/testutil/harnessfixtures/recovery_linux.go); [internal/testutil/harnessfixtures/recovery_linux_test.go](../../internal/testutil/harnessfixtures/recovery_linux_test.go) | metadata-only; preservation snapshot only | Pre-existing dirty implementation/test files. | Other chat; no behavioral conclusions drawn from these three files. | Not inspected for implementation planning | No tests executed | E-013; excluded from behavioral inventory |

Generated files in the metadata group include
`tools/browser_e2e_batch_manifest.json` and `tools/task_surface.generated.mk`.
The topology and work-graph owner JSON files are authored inputs; the
`generated-artifacts/execution-topology.mjs` implementation is authored code,
despite its directory name. Future edits MUST follow the generated-artifact
policy and public Make regeneration path, not edit generated projections.

The remaining repository, caches, ignored result roots, dependencies, binaries,
golden pixels, external services and unrelated domain implementations are
excluded from this bounded audit. Relevant but incomplete future discovery is
**pending**, not deemed safe by exclusion: browser partition consumers and
schema attachments (RB-002), complete effects of proposed mixed-owner rows
(RB-003), and runner/catalog validation selection (RB-006). The inventory does
not establish exhaustive semantic redundancy or specification completeness.

### 2.2 Revalidated catalog census

E-002 records the creation census; E-016 re-parsed all 66 family JSON files and
confirmed the same active-row counts at the review baseline. Counts are semantic
catalog rows, not assertions, work units, processes or physical service starts.

| Dimension | Observed count |
| --- | --- |
| Total active rows | 2,367 |
| Runner | Go 1,002; Vitest 751; Playwright 586; shell 28 |
| Fixture capability | None 1,365; browser stack 586; PostgreSQL dedicated 288; PostgreSQL transaction 113; PostgreSQL migration 11; object-store namespace 4 |
| PostgreSQL scope | Exclusive 890; shared 113; none 1,364 |
| Browser state scope | Test-owned 465; isolated 88; measurement 33 |
| Duplicate literal selector identities | Zero in the inspected comparison |

The literal comparison expands declared Go test names and declared title lists,
then keys by runner, package/file, project, stage and literal name/title. It does
not expand dynamic aliases or prove semantic non-overlap. A zero count MUST NOT
be used as evidence that assurance cannot be consolidated.

### 2.3 Static browser packing model

E-004 records, and E-016 recomputes, the deterministic arithmetic in memory without
importing repository modules, generating a work graph or launching a runner.
It uses the full test-owned population in five distinct base stages. It excludes
the aggregate `resettable` and `isolated` projections to avoid double counting,
all measurement rows, and every isolated row. Visual/support rows are included
as explicit populations; this is not a claim that ordinary CI selects them.

| Stage | Test-owned rows | Existing partitions | Partitions ignoring owner | Existing modeled cohorts | Proposed modeled cohorts |
| --- | ---: | ---: | ---: | ---: | ---: |
| webserver-backed | 325 | 20 | 2 | 24 | 9 |
| support | 17 | 4 | 1 | 4 | 1 |
| stateful | 46 | 8 | 2 | 8 | 2 |
| a11y | 43 | 8 | 2 | 9 | 3 |
| visual | 34 | 10 | 2 | 10 | 2 |
| Total | 465 | 50 | 9 | 55 | 17 |

Reproduction procedure:

1. Read the active family rows, browser batch projection, work-graph owner and
   execution topology. Retain the five base stages and test-owned rows only.
2. For the current model, keep each projected compatibility partition separate.
   For the alternative, compare all projected group fields except `owner_id`
   and the derived `name`, `selected_row_ids` and `specs` fields. Merge only
   otherwise identical groups within the same stage; union their selections.
   Omitting derived selection/name fields permits repacking, not broader
   compatibility or additional tests.
3. Give each indivisible row cost equal to its evidence-class estimate multiplied
   by its selected title count. Browser evidence costs 12,000 ms/title;
   accessibility and visual evidence cost 15,000 ms/title.
4. Sort by descending row cost, then ascending stable row ID. Place each row
   into the first existing cohort that fits; otherwise create a new cohort.
5. Charge 4,000 ms startup and 2,000 ms cleanup once per cohort. Use the applicable
   resource timeout, falling back to the existing 600,000 ms default, minus
   20,000 ms allocation-acquire and 5,000 ms allocation-release headroom.
   All modeled ordinary populations use that 600,000 ms envelope. An oversized
   indivisible item remains explicit; the model does not extend a deadline.
6. Sum cohort counts per stage. Selection-specific execution MUST repeat packing
   after selection rather than reuse these full-population counts.

Input SHA-256 values at revalidation:

| Input | SHA-256 |
| --- | --- |
| Browser batch projection | `1b4e9011a8b0c3a593c3a6f5ad428bfdc5083fa5970f99e3e5885a0f96d2e701` |
| Execution topology | `f4780efb58455c8233082e623123f66ba5bfc0d749f928464595d6842793c9eb` |
| Work-graph owner | `f14027b4ca2d940c277cc480ac0744bd00009c1c7e149bc13c873d6d8f436829` |

**Rationale / limit:** 55-to-17 is a static model of 38 fewer invocation
boundaries. It is neither a measured speedup nor 38 fewer containers. Functional
cohorts already use reusable lanes, and existing resident services may be
unchanged. Longer cohorts also change load balance and the number of selected
rows left unexecuted after failure. Equal metadata alone does not prove
cross-owner state independence; R03-002 remains an entry gate.

### Reservation continuation inventory (E-168)

The clean entry is `f7b917026e1174661a9bdf9118d18120bbd225b0`. The 7,502-file
tracked/nonignored universe was enumerated and preserved in
`.cartulary/remediation03/reservations/entry-source.json`. Executable source
matches the E-166 retained manifest; only the tracker differs. Historical
`r03-s007e-ci-01` remains motivation, not a fresh acceptance sample.

| Scope | Inspection / responsibility | Decision |
| --- | --- | --- |
| Native admission, fixture work/broker/renderer/residency/services and runtime lifecycle | Direct inspection of weighted claims, pressure retirement, consumer counts and release markers | Keep owners; correct their interfaces in S-012/S-013 |
| Go planner/compiler/row runner and authored topology/resource registry | Direct inspection of group-count CPU allocation, authored profiles and child arguments | S-014 profile-owned candidate; keep numeric non-CPU claims |
| Instrumentation owner/shapes, critical-path reader, task surface and native observations | Direct inspection of bounded diagnostic and generated projection seams | Extend existing diagnostic surface in S-012 |
| Native admission/fixture/runtime tests and catalog routing | Relevant assertions inspected; remaining cases are required characterization, not claimed passes | V-023–025 |
| Generated outputs and remaining repository files | Metadata-only; no whole-repository behavioral review | Regenerate projections through Make; no unrelated product change |
| Ignored retained evidence | Explicit E-166 exports and preservation manifest only | Never read private event staging; no historical result reclassification |

## 3. Module Boundary Diagnosis

The target is a coordinated orchestration/lifecycle seam, not an unstructured
collection requiring a replacement scheduler. Existing boundaries should remain
unless a particular requirement below establishes a change.

| Responsibility | Current location | Owner candidate | Disposition | Evidence / finding / contract |
| --- | --- | --- | --- | --- |
| Selection, compatibility and packing | JS catalog/topology/work-graph compiler | harness.command_surface and harness.test_catalog; harness.browser for browser partition behavior | Keep; remove the proposed reporting-only browser boundary after adoption | E-003/E-004; F-001; C-001/C-002 |
| Physical acquisition, residency and retirement | Native fixture broker | harness.browser | Keep; any later S-005 extension belongs here | E-005/E-006; F-005; C-003/C-004 |
| Test-state isolation | pgtest, s3test, browser actor/context fixtures | Harness owners with affected semantic test owners | Keep; candidate-specific changes only | E-005/E-007; F-004; C-002/C-003 |
| Repeated application scenario setup | Incident tests and appsupport | module.incidents; reusable composition remains in appsupport | Simplify assertions/scenarios, not production assembly | E-009; F-002; C-007 |
| Cost and admission policy | Work-graph owner and resource profiles | harness.command_surface | Keep; calibrate only against measured work | E-003/E-010; F-003; C-006 |
| Exact outcomes and teardown evidence | Runner adapters and native lifecycle accounting | harness.evidence_accounting | Keep; preserve row attribution during grouping | E-003/E-008; F-006; C-001/C-005 |
| Documentation history and validation posture | This handoff and prior receipts | Documentation author; no product authority | Keep source-qualified references; do not copy stale conclusions | E-011/E-012; F-007/F-008; C-008 |

### Resource lifecycle and isolation map

| Resource | Physical owner / borrower | Reuse identity and default | Mutable-state boundary | Settlement / uncertainty |
| --- | --- | --- | --- | --- |
| Shared PostgreSQL service/template | Native broker / approved rollback units | Published generation plus provisioning/template identities; shared scope only for approved consumers | Package database; fresh connection/transaction for each test | Full borrow serialized; rollback/close uncertainty poisons reuse |
| Exclusive PostgreSQL service | Broker / one non-browser unit or browser allocation | Unit/allocation-bound generation; measurement uses exact builder/profile/snapshot identity | Dedicated fresh database or lane reset contract | Complete fixture and physical retirement; unrelated allocations cannot attach under current REQ-804 |
| Dedicated test database | pgtest / testing.T | New template clone per independent test preparation | Real committed state; concurrent connections/workers remain meaningful | Test cleanup drops only its owned database; uncertain cleanup prevents unsafe reuse |
| Renderer container | Native renderer owner / one cohort borrower | Run, runtime/profile, immutable trust and admitted renderer identity; pooled ordinary, dedicated measurement/review | Clients and Chromium descendants are connection-owned | Successful return requires live server and absent clients, browser descendants and forwarding listeners; otherwise quarantine/retire |
| Browser context / actor | Test fixtures / one test | Fresh context and independently persisted actor/session; worker provisioner is control-plane only | Cookies, storage, permissions and observed actor state | Close contexts and revoke/verify owned sessions; never carry mutable user state to another test |
| Application stack | Broker / compatible browser lane | Exact stage/session/runtime/resource/fixture affinity | Cohort state; required boundary reset replaces runtime-local state and restores database/bootstrap state | Reset failure quarantines; no product retry or timeout extension |
| Object-store service/namespace | Published service owner / namespace borrower | Shared physical service; exact owned namespace or serialized package reset | Objects and bucket-wide operations remain scoped to the declared lease | Empty/reset and mutation readiness proof; namespace cleanup cannot be replaced by PostgreSQL termination |
| Populated performance snapshot | Exact run/profile/snapshot builder / clones | Sealed result within the run only | Fresh clones; measurement lifetime and quiet admission remain distinct | No cross-run snapshot discovery or mutable snapshot reuse |
| Vitest coordinator/file worker | Runner invocation / selected file | Compatible bounded coordinator cohort; one isolated file worker at a time | File environment, setup and teardown isolated | Acknowledge worker termination before proceeding; invocation failure can supersede provisional row success |

Rationale: infrastructure reuse and test-state reuse are different decisions.
Fresh database names cannot isolate cluster-wide cleanup effects; fresh record
IDs cannot isolate administrator listings, clocks, faults or shared sessions.
A pooled renderer does not mean a persistent shared Chromium process.

## 4. Public Contract and Behavior Freeze Map

Harness REQ/AC abbreviations below mean `TH-HARNESS-REQ-` and
`TH-HARNESS-AC-` in the adopted v3 NLSpec. E-016 rechecked the affected
REQ-805/806/835 and incremental-evidence rules; other baseline observations
remain attributed to E-001. Proposed changes are not current conformance violations.

| Contract / surface | Observed behavior and evidence | Adopted owner / required behavior | Disposition | Prerequisites, risk and validation |
| --- | --- | --- | --- | --- |
| C-001 Selection and attribution | Go/Vitest grouping removes reporting-owner partitions; browser work units already execute as harness.browser; rows retain their identities; E-003/E-004 | REQ-379/805/806; exact selection, attributable native outcomes, no widening | Preserve exact selection/accounting | Schema/consumer discovery and mixed-owner cases; S-001, V-001/V-002; RB-002 |
| C-002 Browser compatibility and state | Browser partition owner is required; test-owned rows can pack, isolated/measurement rows are singleton; E-004 | REQ-806 currently preserves owner and state partitions; AC-088 | Proposed owner amendment only for reporting-owner boundary | Adoption and cross-owner independence proof; S-001/S-004, V-002/V-003; RB-002/RB-003 |
| C-003 PostgreSQL/object ownership | Shared rollback cluster; unit-bound exclusive cluster; fresh clones and namespace cleanup; E-005 | REQ-804/811; AC-086/095; provisioning/retirement are within scope, cleanup uncertainty prevents reuse | Preserve; S-005 proposes a separate lifetime amendment | Complete connection/effect tracing and settlement evidence; V-003/V-007; RB-005 |
| C-004 Renderer/trust and measurement | Existing native pool; connection-owned browsers; sealed snapshots; E-006/E-010 | REQ-804/806 and populated-fixture REQ-812–815; immutable trust, quiet measurement and dedicated review | Preserve | No shared contexts, persistent Chromium daemon or measurement shortcut; V-003/V-005 |
| C-005 Failure, cancellation and evidence | Settled file outcomes and physical ownership outlive individual producer activity; E-006/E-008 | REQ-803–806; AC-085–088; primary/secondary failures and quarantine retained | Preserve | Failed/reset/interrupted work must not become success or silently retry; V-002/V-003/V-005 |
| C-006 Costs and resource admission | Generic per-row Go/Vitest estimates; browser title-scaled estimates; serialized rollback bucket; E-003/E-005/E-016 | REQ-803/805/806/835 and Section 16 REQ-675; lifecycle claims, diagnostic validity/completeness and immutable attempt history | Preserve defaults; evidence-gated calibration | Preregistered comparisons and separate assurance/performance decisions; Section 8.2; S-003, V-005; RB-004 |
| C-007 Product assurance | Full HTTP incident scenarios overlap in setup but have distinct assertions; E-009 | Core 01 REQ-01-157, 160, 162–167; Core 00 REQ-00-035 owner allocation | Preserve; add missing replay-header assertion while consolidating setup | Named survivor, unchanged route/status/envelope/auth/storage behavior, exact catalog update; S-002, V-004 |
| C-008 Documentation and projection boundary | Generated schemas/Make are downstream; lint does not select this file; prior receipt posture differs by section; E-001/E-011/E-012 | Harness REQ-001/003; root and tooling instructions | Preserve; documentation-only audit evidence | No runtime Markdown input, stale-pass claim, new compatibility layer or unauthorized linter change; V-006/V-008 |
| C-009 Invocation accounting | Existing diagnostics are partial; fresh cgroup and exact Docker placement probes succeeded, E-024/E-025 | REQ-856/AC-152 adopted for user-authorized S-008; REQ-847/849/851 retain their existing boundaries | Implement required accounting through public Make, native ownership and validated readers | Missing or malformed accounting blocks required acceptance; no changed-source equivalent admission or privileged tracing requirement; V-009 |
| C-010 Admission explanation | Host phase lacks decision-level causes | REQ-803/857; adopted amendment before S-012 | Separate graph, active, resident, prospective and measured quantities | R03-035; V-023 |
| C-011 Go execution budget | Selected group count overwrites profile CPU | REQ-379/805/803; candidate tested and restored | NO_CHANGE after performance rejection; selection coupling remains a future investigation | R03-038; S-014/V-025; E-194 |
| C-012 Retirement and consumer lifecycle | Quantity overlap and incomplete future-consumer counts | REQ-803/804; correction/amendment before S-013 | Exact releasable sets, weighted counterfactual, terminal consumer accounting | R03-036/037; V-024 |
| C-013 Performance and claim sizing | No per-dimension proof from idle utilization | REQ-353/803/856; Section 7.8 user acceptance | Preserve non-CPU quantities; isolate optional Go tuning | R03-039–041; V-026/027 |

HTTP/WebSocket, row/query/mutation, authorization, revisions, projections,
saved views, view schemas and UI selectors remain production-owner contracts.
S-001 changes execution placement, not those behaviors. S-002 changes where
HTTP assertions execute, not the product contract. Product API changes,
database migrations, frontend presentation and golden refresh are not applicable
to the proposed near-term edits. Any later discovered need for such a change
MUST become a separately owned/authorized task rather than an implicit part of
optimization.

## 5. Coupling and Boundary Findings

The direct-evidence column preserves the inspected baseline. Dispositions below
record the completed investigations and accepted changes; measured benefits remain
limited to their named source and workload receipts.

| Finding / classification | Direct evidence | Contract / owner | Required action or disposition | Slice / validation |
| --- | --- | --- | --- | --- |
| F-001 / should_fix | Owner participates in browser partition identity and validation; model yields 55 versus 17 cohorts; E-004 | C-001/C-002; harness.browser and catalog owner | ACCEPTED: owner-independent v15 with explicit compatibility and state proof, E-049; final integration E-072 | S-001; V-001/V-002/V-003/V-005 |
| F-002 / should_fix | Standalone Location test starts an application fixture for one response assertion; replay scenarios overlap but differ; E-009 | C-007; module.incidents | CORRECTNESS ACCEPTED: exact create/replay Location assertions and independent routing, E-055. Fixture removal REJECTED by E-054; original lifetimes retained; further replay consolidation DROPPED | S-002; V-004/V-005 |
| F-003 / defer | Go/Vitest price a catalog row independently of its selected test count; fixture borrowing can serialize; E-003/E-005 | C-006; harness.command_surface | NO_CHANGE, E-057: physical admission passes; selected-work estimate calibration and >5% benefit remain unproved; row-count limitation explicitly retained | S-003; V-005 |
| F-004 / defer | 113 transaction rows and 465 test-owned browser rows already use reuse paths; remaining eligibility is not proved; E-002/E-005 | C-002/C-003; affected semantic owners | NO_CHANGE, E-058: 895 declarations screened; 525 excluded, 370 unresolved, zero eligible conversions; retain stronger boundaries | S-004; V-003/V-005/V-007 |
| F-005 / defer | Exclusive cluster key includes unit ID and release retires it; E-005 | C-003; native fixture broker | NO_CHANGE, E-059: >10% opportunity unproved; dedicated lifetimes retained. E-070 template sealing is a correctness repair, not lifetime reuse | S-005; V-005/V-007 |
| F-006 / intentional/no_action | Grouped Go/Vitest, template clones, renderer pool, namespace reset and snapshots already exist; E-003/E-005/E-006/E-008/E-010 | C-001/C-003/C-004/C-005 | Preserve the foundations; no second broker, scheduler, resident test daemon or reimplementation justified | No new slice; preservation checks V-002/V-003 |
| F-007 / must_fix for this handoff's evidence accuracy | Opening blocked summary, historical appendix success, active full-7 failure and later full-8 success describe distinct source/session postures; E-012/E-016 | C-008; document author | RESOLVED: historical source/session receipts remain qualified; current implementation and final verification recorded separately, E-018–E-072 | WF-00/WF-06; V-008 |
| F-008 / must_fix for truthful validation reporting | Lint glob excludes remediation 03; Make readiness/cache/report writes exist; E-011 | C-008; documentation/tooling owner | RESOLVED: user-authorized canonical lint selection and structural repair, E-020; final completion lint E-073 | WF-06; V-006/V-008 |
| F-009 / must_fix for selected performance acceptance | Whole-invocation guardrails cannot be established by partial samples and CPU lower bounds, E-021/E-022 | C-009; harness.evidence_accounting with command-surface and fixture owners | ACCEPTED: S-008/S-008a charged resource accounting and fail-closed evidence, E-030/E-038, within declared Linux/cgroup/container/build scope | WF-08; V-009; RB-009 |
| F-010 / must_fix for diagnostic completeness | The existing performance view truncates paths at 20 units and lacks retirement causes; historical CI queueing cannot be completely attributed | REQ-079/835/841–844/857; harness.evidence_accounting and native fixture owner | ACCEPTED: complete paged canonical diagnostics and independent lifecycle coverage, E-081; required-accounting boundary repairs E-080/E-081 | S-009; V-010 |
| F-011 / conditional optimization | Separately scheduled resets can release a stack before the successor cohort and create reset-only generations | REQ-801/803/806; harness.browser and native runtime/fixture owners | REJECTED / RESTORED, E-160/E-161: assurance passed but both performance gates reject; dedicated reset scheduling restored | S-010/S-011; V-011/V-012 |
| F-012 / must_fix | Candidate environment protection discards inherited host admission for a nested executor with no local lease; nested cancellation assurance waits for unrelated host capacity | REQ-102/858; native process execution owner | ASSURANCE PASSED, E-112; local/inherited/absent binding tests and installed cancellation under 19-token and 4-token parents | WF-11c; V-013 |
| F-013 / must_fix | Mapped-port publication starts a redundant SQL-readiness window; failed-start graceful stop can consume the entire removal budget | Section 11.4; REQ-401/402; PostgreSQL startup owner | ASSURANCE PASSED, E-118; acquisition-owned caps, remaining attempt readiness and unpublished failed-start removal applied to both treatments | WF-11d; V-014 |
| F-014 / must_fix | Assessment invalid-timestamp browser scenario begins editing while setup records may still await collaboration publication; observed draft-review invalidation causes timeout | Core 03 REQ-03-304/305 and retained intent/continuity; existing fixture/test owner | ASSURANCE PASSED, E-123; exact fixture publication precedes authoring in both treatments; historical triggering message remains unavailable | WF-11e; V-015 |
| F-015 / must_fix | Deliberate nil-context rejection test triggers Staticcheck SA1012, failing full CI despite passing runtime assurance | Existing startup context rejection contract, NLSpec Section 11.4; static-analysis test owner | ASSURANCE PASSED, E-127; negative assertion retained and both-source Go lint passes | WF-11f; V-016 |
| F-016 / must_fix | Mention-recovery browser test assumes immediate successful presentation refresh after an accepted resolution, although the owner permits retained read-only recovery | Core 03 mention completion, REQ-03-283, acknowledged save-state and REQ-03-130/131 | ASSURANCE PASSED, E-130/E-131; controlled read recovery preserves all mutation assertions in both sources | WF-11g; V-017 |
| F-017 / must_fix | Space discovery test advances the shared page clock without proving linked metadata admission; failed metadata correctly disables its later Preview action | Core 03 REQ-03-248a and REQ-03-220; browser test owner | ASSURANCE PASSED, E-135/E-136; both-source exact/containing/owner checks; historical callback order remains unavailable | WF-11h; V-018 |
| F-018 / must_fix | Two owned forced drops exceed the 20-second operation cap while waiting for checkpoint file sync; later setup is rejected after quarantine | Harness NLSpec Section 11.4 and AC-007; PostgreSQL cleanup owner | IMPLEMENTED, E-137–E-141; shared bounded deadline and both-source focused assurance pass; current full-CI gate remains required | WF-11i; V-019 |
| F-019 / must_fix | PostgreSQL reaches internal readiness while its selected published host endpoint fails SQL connectivity and later refuses TCP connections | Harness NLSpec Section 11.4; PostgreSQL acquisition and endpoint ownership | OPEN for final integration; E-150 retains publication and E-152 qualifies 192/192 lifetimes; historical cause remains uncertain | WF-11j; V-020 |
| F-020 / must_fix for attribution | E-166 path has 1,650,220 ms host admission without detailed native deficits | C-010; native admission/evidence owners | Bounded decision observations, no second evaluator | S-012; V-023 |
| F-021 / should_fix | 192 ordinary CI Go units claim four CPUs although standard profile authors one | C-011; compiler/Go runner | Candidate rejected; prior budgets restored after two slower matched pairs. NO_CHANGE for this scope; future coupled admission design may revisit | S-014; V-025/026; E-194 |
| F-022 / must_fix | Retirement checks overlapping quantities rather than exact weighted relief | C-012; native admission/fixture owner | Counterfactual exact release set; preserve uncertain claims | S-013; V-024 |
| F-023 / must_fix | Static stack release marker and renderer profile counts omit terminal selection state | C-012; fixture/runtime owners | Exact compatible consumer registry, idempotent terminal notifications | S-013; V-024 |
| F-024 / intentional/no_action for numeric sizing | Low observed utilization does not prove safe non-CPU reductions | C-013; topology/capability owner | NO_CHANGE memory/process/port/service and host capacity values; register new evidence separately | S-012; V-023/026 |
| F-025 / must_fix | Second serial Task Requests handoff checks grid visibility without dock closure or destination focus before Work activation | Core 03 REQ-03-314/315; workbook browser test projection | REPAIRED; shared completion assertions, unchanged deliberate supersession coverage; focused assurance passed | S-016a; V-027a; E-198/E-199 |

No production module movement, vendor abstraction replacement or direct-storage
decoupling is justified by this audit. Finding classifications do not authorize
implementation; a must_fix item is confined to the acceptance named in its row.

## 6. Refactor Workstreams

This is the authoritative execution order. Each row requires a tracker receipt
before its successor starts. Investigation completion, implementation acceptance,
assurance acceptance and measured performance are separate dispositions. A
supported no-change outcome may close an investigation; a required correctness
failure remains open and cannot be relabeled deferred.

| Workflow / slice | Dependency | Required result | Binary exit |
| --- | --- | --- | --- |
| WF-00 baseline/specification/acceptance | None | Revalidate source, authority, consumer roster, exact selections and fixed experiment method | Current source and complete first experiment protocol recorded |
| WF-06 / S-006 canonical tracker validation | WF-00 | Include this tracker in public Markdown lint without another CLI | Coverage confirmed and `make lint-markdown` passes |
| WF-08 / S-008 measurement capability | WF-06; explicit user scope expansion after E-023 | Establish complete invocation resource accounting, adopt the owner contract and validate producer/readers through public Make | V-009 passes; full-scope pilot and secure validated export available; missing coverage remains a blocker |
| WF-01 / S-001 browser compatibility | WF-08 | Prove effects; adopt REQ-806/AC-088; migrate browser manifest v15 and consumers atomically | G-03 and structural G-04, or supported no-change with candidate disposition |
| WF-08a / S-008a nested accounting correction | Settled E-037 pilot discovery; entered before S-001 resumes | Inherit the proven outer accounting policy after environment clearing; preserve ordinary defaults | Independent-publication and nested cancellation receipts, exact membership/scans/settlement, then tracker/lint boundary |
| WF-02 / S-002 incident assertions | WF-01 disposition | Add create/replay Location assertions and exact-symbol routing; conditionally remove redundant fixture | E-055 correctness accepted; E-054 rejects removal under amended G-04 and restores all original fixtures |
| WF-03 / S-003 costs/admission | WF-02 disposition | Measure resulting workload; fix actual accounting defects; conditionally calibrate existing policy | G-03 and >5% benefit, or measured no-change; correctness defects stay open |
| WF-04 / S-004 isolation | WF-03 disposition | Screen stronger declarations; pilot highest-cost eligible rollback/browser candidates in separate child workstreams | Complete effect and runtime proof plus >5% benefit, or justified stronger isolation |
| WF-05 / S-005 PostgreSQL lifetime | WF-04 disposition | Measure remaining startup cost; conditionally adopt bounded sequential reuse | Full settlement/exclusion proof plus >10% benefit, or dedicated-lifetime decision |
| WF-07 / S-007 validation and handoff | WF-05 disposition | Finalize, verify projections/owners/CI and affected explicit-only suites; close evidence and tracker | Required checks pass, exact ownership settled, final tracker lint passes |

Continuation sequence (the preceding table remains the completed original effort):

| Workflow / slice | Dependency | Required result | Binary exit |
| --- | --- | --- | --- |
| WF-00e registration | E-073; new explicit authorization | Revalidate source; register requirements, validation and preservation boundary | Traceability complete; Markdown lint passes |
| WF-09 / S-009 critical-path diagnostics | WF-00e | Adopt diagnostic owner; paged views and terminal allocation lifecycle observations; diagnostics-equipped baseline pilots | V-010 passes; required coverage available; tracker and lint pass |
| WF-10 / S-010 preparation lifecycle | WF-09 | Adopt ordered preparation; successor holds one admitted borrow through reset and runner | V-011 passes; performance pending; tracker and lint pass |
| WF-11 / S-011 performance decision | WF-10 | Frozen browser and CI comparisons under Section 8.5 | Both decisions recorded; rejected lifecycle owner/projections/code restored; tracker and lint pass |
| WF-11b corrective assurance gate | E-099/E-100 failed baseline warm-up; serial prerequisite inside WF-11 | Repair socket-object fixtures; diagnose PostgreSQL deadline failures and preserve cleanup/accounting distinctions | Affected assurance passes, unresolved causes and settlement are explicit, source repairs applied to both treatments, new freeze and lint precede sampling |
| WF-11a corrective assurance gate | E-091/E-092 failed baseline warm-up; serial prerequisite inside WF-11 | Diagnose and repair nested installed-service startup; preserve the failed run and re-freeze both treatments if source changes | Exact failed route and affected containing assurance pass; evidence/settlement complete; tracker and lint pass before sampling resumes |
| WF-11c inherited-admission correction | E-109; active invocation settled and entry lint passed | Restore real inherited parent lease/CPU binding in the candidate while preserving local admission precedence and fixture protection | Native precedence/failure tests and installed nested cancellation pass; source receipt and lint before WF-11d |
| WF-11d PostgreSQL startup correction | WF-11c; Section 11.4 adoption before projections | One bounded attempt-owned readiness budget; forced termination only for unpublished failed-start allocations; diagnostic v5 | Deadline/probe/cancellation/cleanup and exact affected service checks pass; common correction in both treatments; new source freeze and lint before WF-11 resumes |
| WF-11e assessment fixture publication | E-120 settled baseline failure and entry lint; WF-11d complete | Observe committed fixture record publications before opening the assessment draft, using the existing socket monitor; preserve product review and all assertions | Affected row alone and containing selection pass; owner review and monitor assurance unchanged; identical correction in both sources; tracker/lint and new freeze before sampling |
| WF-11f negative-test static analysis | E-126 failed warm-up is settled and preserved; entry Markdown lint passes | Document the deliberate invalid argument at its exact test call; retain global SA1012 enforcement and identical B/C correction | Public Go lint and affected test pass in both sources; finalize/projections pass; tracker and closing lint precede fresh freeze |
| WF-11g mention refresh recovery | E-129 settled baseline browser failure and entry lint; WF-11f complete | Make post-acknowledgement refresh failure deterministic in both recovery branches; require explicit read-only recovery and unchanged mutation counts | Affected row alone and containing entity recovery rows pass in both sources; owner recovery contracts, types and projections pass; common correction and closing tracker/lint precede fresh freeze |
| WF-11h linked metadata precondition | E-134 settled failure and entry Markdown lint | Establish admitted metadata before advancing the Space-discovery clock; preserve access, deadline, late-result and explicit-preview assertions | Exact and containing browser selections plus metadata owner assurance pass in both sources; identical correction, source preservation, tracker and closing lint |
| WF-11i cleanup-budget investigation | WF-11h complete and boundary lint; E-134 retained checkpoint evidence | Evaluate one shared existing overall cleanup budget instead of an independent operation cutoff; adopt any behavioral amendment before code and projections | Parent deadline, admission, single drop, cancellation/quarantine and settlement assurance; affected integration checks in both sources; no unresolved required failure, new freeze before sampling |
| WF-11j published endpoint investigation | E-144 settled cancellation; entry tracker lint | Execute the Section 7.7 stability plan; distinguish cause, repository repair, environment qualification and unresolved historical evidence | E-147 planning receipt; child gates below precede a new performance freeze; no speculative retry or deadline increase |
| WF-11j-a transport characterization | WF-11j registration and planning lint | Adopt diagnostic additions if required; establish a bounded exact-allocation transport reproducer through existing public Make routing | Request and lifecycle identity proven; positive/negative transport matrix passes; every attempt retained; tracker and lint |
| WF-11j-b publication correction, conditional | WF-11j-a identifies a reproducible publication-path defect | Pilot the narrow daemon-publication alternative only if it addresses the reproducer and preserves the owner contract | Deterministic regression and strict publication/ownership/security matrix pass; common B/C correction or explicit rejected/no-change disposition; tracker and lint |
| WF-11j-c stability qualification | WF-11j-a and WF-11j-b dispositions | Run bounded lifecycle qualification and affected service assurance on the retained path; verify accounting and complete settlement | Section 7.7 qualification gate passes; recurrence blocks performance; tracker and lint |
| WF-11j-d experiment handoff | WF-11j-c | Finalize changed implementation, regenerate projections, freeze sources and amended experiment under Section 8.5 | No unresolved owned resources; exact source/treatment and environment bindings; closing tracker/lint before WF-11 resumes |
| WF-07f cleanup-delegation assurance | E-162 settled owner-check failure and entry lint | Replace elapsed-time inference with direct successful-child cleanup delegation, exact lease and no direct-close assertions; preserve runtime semantics | Focused lifecycle row, containing browser owner, finalization and projections pass; tracker/lint before final validation resumes |
| WF-07g renderer accounting fixture boundary | E-164 settled UI-review failure and entry lint | Complete renderer infrastructure dependency injection so fake Docker never invokes real placement proof; retain strict native defaults | Accounted/unaccounted/failure fixture matrix, exact row, containing UI-review and real visual checks pass; tracker/lint before final CI |
| WF-07e / S-007e final handoff | WF-11 | Finalize, projections/owners, explicit browser/support/visual/UI-review, final CI and evidence audits | All mandatory checks and settlement pass; final tracker lint passes |

No replacement scheduler, broker, resident test daemon, compatibility mode or
parallel agent execution is part of this effort. Additional conversions under
S-004 receive separate child workstream IDs and serial boundary receipts.

### Reservation continuation execution graph

| Workflow / slice | Dependency | Required result | Binary exit |
| --- | --- | --- | --- |
| WF-00f registration | E-167 historical completion; current user authorization | Section 7.8 traceability and preserved entry source | E-168 and registration lint |
| WF-12 / S-012 admission diagnostics | WF-00f | Adopt owner, implement bounded explanations, audit claims, focused pilots | V-023 and diagnostic baseline preserved |
| WF-13 / S-013 retention correctness | WF-12 | Exact release evaluation and complete consumer accounting | V-024, no unresolved ownership |
| WF-14 / S-014 Go budget candidate | WF-13 | Profile-owned CPU and migrated consumers | V-025; candidate pending performance |
| WF-15 / S-015 performance decisions | WF-14 | Test-fast benefit, browser guardrails, assumed CI elapsed comparison | V-026; accepted/rejected/inconclusive disposition and rollback |
| WF-16 / S-016 final handoff | WF-15 | Finalization, required checks, final CI/audit/settlement | V-027, final tracker receipt and closing lint |
| WF-16a / S-016a handoff test repair | E-198 user cancellation, settlement and entry lint | Repair incomplete serial handoff assertions under existing Core 03 owner | V-027a, E-199 focused assurance and closing lint; does not close incomplete WF-16 |

## 7. Proposed Refactor Slice Plan

### 7.1 Proposed normative contracts and binary acceptance

The following are **proposed requirements**, subject to Section 1.2. Supporting
tables in Sections 7.2–7.6 and 8.1–8.2 refine their predicates, not separate
unidentified runtime requirements. They MUST
NOT be cited as adopted runtime behavior. Each requirement's acceptance column
is binary; an absent proof is unmet, not presumed true.

| ID | Proposed requirement | Binary acceptance / validation |
| --- | --- | --- |
| R03-001 | S-001 MUST remove reporting ownership alone from browser compatibility and preserve Section 7.2's execution/state tuple. Only test-owned rows may form multi-row cohorts. | Owner-only differences yield one compatibility partition; a within-budget fixture packs together. Each incompatible dimension separates or rejects; isolated/measurement rows remain singleton. Merging partitions does not require every row to fit one cohort. V-002 |
| R03-002 | Before candidate product execution, affected owners MUST review a complete source-effect dossier and an explicit independence matrix. Runtime proof is required for acceptance, not as a prerequisite to performing that authorized characterization. Unknown source effects MUST retain stronger separation and block the affected cutover. | Every newly combinable row/helper has a reviewed state owner; the matrix covers alone, pairwise order and applicable concurrent-lane interactions. No dossier gap or missing matrix outcome is accepted. Synthetic harness fixtures may characterize failure paths before product execution. V-003 |
| R03-003 | Selection MUST precede packing. Row IDs, exact selectors, reporting owners and verification attribution MUST remain exact. Execution owner MUST remain harness.browser; reporting ownership MUST come from the catalog. | Section 7.2's membership oracle proves exact identity-set equality, one terminal disposition per selected case and correct row aggregation. Missing/duplicate/foreign results or false pass of an unexecuted case fail acceptance; diagnostic event repetition is not a duplicate terminal result. V-002 |
| R03-004 | Remove partition owner_id and normalized ownerID in one versioned producer/consumer cutover. Do not add an owner fallback, list registry, alias, dual reader or grouping toggle. | The migrated consumer roster and schema rejection cases pass. No partition-owner lookup remains; catalog row-owner lookups remain required. Old artifacts remain historical and are not translated into new acceptance evidence. V-002/V-006 |
| R03-005 | Packing MUST retain descending cost order, stable row-ID ties, indivisible selections, existing timeout headroom, one worker and max_failures=1. | Identical selected semantics produce identical cohorts/digests independent of input enumeration. Total estimated cost at the inclusive envelope fits; an individually oversized row remains explicit without extra selected cases or extended deadline. Actual timeouts remain failures. V-002 |
| R03-006 | Grouping MUST preserve fresh contexts/actors, affinity/reset boundaries, failure quarantine, cancellation settlement and independently settled outcomes. Product work MUST NOT retry. | Product, teardown, reset, cancellation and owner-loss injections preserve primary/secondary failures, prevent contaminated reuse and retain exact owned-resource cleanup. V-002/V-003 |
| R03-007 | Before retiring the standalone Location test, S-002 MUST add first-create and replay Location checks to the named survivor's existing HTTP responses. | Each Location equals the literal `/api/v1/incidents/` prefix plus `firstData.incident_id`; first-create is 201 and replay is 200. Payload equality, actor-scope and conflict assertions survive; no new request or fixture is added. Standalone canonical survivor execution and the aggregate preservation map pass. V-004 |
| R03-008 | Removing any further scenario MUST require an assertion-to-survivor map covering all distinct behaviors and evidence contexts. Catalog selector changes MUST be atomic with test changes and MUST NOT preserve obsolete aliases. Helper extraction alone MUST NOT count as an efficiency improvement. | Every removed assertion has a named, independently executable survivor in the same required assurance context; unsupported equivalence blocks removal. V-001/V-004 |
| R03-009 | Every efficiency comparison MUST satisfy the fixed protocol in Section 8.2 before measured baseline or candidate attempts. | All mandatory protocol fields and numerical limits are fixed; source identities bind each attempt before launch; all scheduled observations and failed attempts remain accounted. Explicit user amendments retain the original registration, their authority, completed observations and every omitted attempt. Repetition or selective replacement cannot supersede a same-source failure. Missing fields or unauthorized post-hoc selection fail G-04. V-005 |
| R03-010 | Separate startup, file setup, execution, reset, cleanup and publication costs. A changed estimate or admission policy MUST have measured provenance at its actual estimation granularity. | Apply Section 8.2: S-001/S-002 require observed structural simplification. Each workload rejects the candidate if both first-two-pair improvements are below 5%, qualifies for the user-authorized >=20% early exit, or passes full four-pair validation with all non-regression guardrails and the applicable S-003/S-004/S-005 benefit threshold. A static count alone cannot prove performance. Unknown required metrics are inconclusive; early exit does not establish omitted confidence bounds. V-005 |
| R03-011 | Resource admission MUST account for physical resident services and active work, including cleanup headroom. Existing browser lanes, worker counts, capacities and deadlines MUST remain the defaults unless a separately reviewed measured proposal changes them. | No child work is uncharged; actual configuration matches retained identity; increased concurrency cannot be accepted solely because some queues shorten. V-003/V-005 |
| R03-012 | S-004 MUST trace every consequential effect before choosing a weaker boundary. Independent commits, escaping connections, workers or process-global effects invalidate rollback-only eligibility unless confined and settled by the candidate's explicit contract. | The complete dossier and required alone/order/concurrency/failure matrix pass. Unknown effects retain the original declaration. No declaration changes on test-name or unique-ID inference alone. G-04 separately establishes benefit. V-003/V-005/V-007 |
| R03-013 | S-005 investigation is active in the approved sequence. Measure present lifetimes and design settlement; longer-lifetime implementation requires G-02, including adopted replacement of REQ-804's exclusive-lifetime restriction. | Current-lifetime baseline identifies material avoidable cost, the adopted design fixes all Section 7.4 pool limits and settlement rules, and characterization routes exist before implementation. Candidate safety/benefit results are G-03/G-04 outputs, not investigation entry conditions. V-005/V-007 |
| R03-014 | An admitted S-005 design MUST use the existing run-owned broker, immutable templates, fresh databases and bounded resident generations. Unrelated borrowers of one physical generation MUST be sequential; no cross-run reuse is allowed. | Pool-limit, expiration, quiet-admission, stale-borrower and simultaneous-borrow cases pass; only one reusable generation may exist; independent dedicated generations remain within unchanged capacity. A fresh logical borrow identity fences the previous borrower before reassignment. V-007 |
| R03-015 | Reassignment requires complete borrower settlement and a healthy generation. Uncertainty quarantines the generation and retains claims until exact physical settlement; it grants no new destructive authority. Migration, cluster mutation, lifecycle/deadline, checkpoint-subject and measurement cases retain dedicated lifetimes. | The lifecycle matrix covers failure, cancellation, owner loss, publication loss, orphan connections/work and excluded cases. No uncertain generation is reissued. Final/evicted/quarantined services retire only through existing ownership proof; failed retirement stays unresolved. V-003/V-007 |
| R03-016 | Existing renderer pooling, rollback fixtures, durable template cloning, immutable performance snapshots, runner grouping and incremental evidence MUST remain intact unless a separately justified owner change is adopted. | Preservation cases and exact evidence/cleanup contracts pass; no recreated infrastructure layer or weakened durability substitutes for optimization. V-002/V-003/V-006 |
| R03-017 | Static models, historical receipts, assurance and measured improvement MUST remain separate. Missing observations are unknown. Required failures MUST NOT be concealed by favorable performance. | G-03/G-04 have separate decisions and exact evidence. No accepted implementation without assurance and its applicable performance gate; an authorized early-exit decision is distinct from a full statistical pass. Deferred/dropped work records candidate source disposition and retains failures. Estimated costs, lifecycle counts or historical passes alone cannot establish efficiency. V-005/V-008 |
| R03-018 | S-008 MUST establish complete ownership and lifetime coverage before producing whole-invocation CPU or memory acceptance evidence. Existing partial diagnostics and REQ-847 receipts retain their original scope. | V-009 proves short-lived descendants, service startup/retirement, nested invocations, cancellation, terminal publication, exact resource scope and fail-closed unavailable/invalid outcomes. No retired REQ-851 experiment or changed-source equivalent comparison is reactivated. |
| R03-019 | S-009 SHALL extend explain-run with deterministic bounded critical-path pages, complete totals and explicit coverage; existing canonical readers remain the sole event interpreter. | V-010: all pages reconstruct complete target-local paths; filters do not alter path totals; invalid selectors and artifacts fail with their distinct classes. |
| R03-020 | The native fixture owner SHALL record each browser allocation's actual retirement trigger, timing, borrows and settlement independently of verbose truncation. | V-010: all closed reasons, missing triggers, repeated/conflicting records, limit exhaustion, failure and unsafe evidence cases. |
| R03-021 | S-010 SHALL execute ordered declared preparation under the successor unit's exact host and fixture borrow; remove standalone reset scheduler units atomically. | V-011: selection equivalence, identity sensitivity, bounds, ordering, separate deadlines and same-generation continuity under pressure. |
| R03-022 | Preparation failure SHALL prevent runner launch and preserve step outcomes, unexecuted cases, primary failures and physical settlement authority. | V-011: reset, publication, identity, cancellation, owner-loss and cleanup failure matrices; no false success. |
| R03-023 | S-011 SHALL apply Section 8.5 independently to browser and CI; rejection by either workload SHALL restore lifecycle owner/projections/code together while retaining accepted diagnostics. | V-012: frozen samples, audits, both decisions, explicit confidence/coverage limits and rollback outcome. |
| R03-024 | S-007e SHALL complete mandatory current-source verification and handoff independently of performance acceptance. | V-006/V-008/V-012: finalize, required checks, final CI, secret scans, settlement and final tracker lint. |
| R03-025 | Native execution SHALL preserve a real inherited parent host lease and CPU budget when no local lease is supplied; local admission SHALL take precedence and command/fixture values SHALL create neither authority nor inherited fixture borrows. | V-013: local/inherited/absent admission, command overrides, preparation/runner parity, fixture isolation and installed nested cancellation. |
| R03-026 | SQL readiness SHALL use the existing bounded attempt deadline; failed unpublished startup SHALL reserve its existing cleanup budget for exact termination/removal. Current diagnostic contracts SHALL migrate atomically after owner adoption. | V-014: late readiness within the cap, expiry, probe clamps, cancellation, invalid parents, terminal rejection, separate failures, ordinary shutdown preservation and v5 schema rejection. |
| R03-027 | The affected assessment scenario SHALL establish its fixture-publication precondition before authoring. It SHALL preserve all timestamp, no-write, append/history, selection and confidence assertions and the product review boundary. No sleep, retry, timeout increase or automatic draft resume SHALL replace the precondition. | V-015; existing Core 03 behavior remains authoritative; this is a test-fixture correction, not new product semantics. |
| R03-028 | The deliberate nil-context rejection test SHALL retain its invalid input and assertion. Any static-analysis exception SHALL apply only to that call and explain the required negative test. Global lint policy and runtime semantics SHALL remain unchanged. | V-016; existing startup context contract; no new behavioral requirement or schema. |
| R03-029 | The mention creation recovery scenarios SHALL distinguish accepted mutation receipts from presentation refresh acceptance. They SHALL deliberately fail the post-acknowledgement Timeline read, assert retained refresh-required state, explicitly recover by reads, and assert that creation/resolution counts and original receipt/version guarantees remain unchanged. No retry loop, sleep, deadline increase or assertion removal SHALL substitute for this controlled boundary. | V-017; existing Core 03 completion and read-recovery authority; test correction only. |
| R03-030 | The Space-discovery browser scenario SHALL establish current linked metadata admission for both selected records before advancing the page clock. It SHALL preserve the held preview request, discovery deadline, indeterminate result, explicit preview and late-result assertions. No forced click, automatic retry, sleep, deadline increase or product access relaxation SHALL substitute for the precondition. | V-018; existing Core 03 REQ-03-248a and REQ-03-220; test-only correction. |
| R03-031 | Cleanup investigation SHALL preserve the failed attempts, distinguish per-test failure from later service disposal, and evaluate phase budgets against the existing bounded ownership lifetime. Any amendment SHALL be adopted in the NLSpec before implementation. It SHALL NOT increase the overall cleanup allowance, change scheduler capacities/fairness/isolation, add retry, bypass checkpoints, release uncertain ownership, or convert failed cleanup into success. | V-019; Section 11.4 and AC-007 amendment adopted in E-137 before implementation; bounded-budget acceptance remains required. |
| R03-032 | Endpoint investigation SHALL distinguish container startup, internal listener readiness, published bindings, host transport and SQL readiness. A repair SHALL preserve exact ownership, bounded acquisition, failure attribution and physical settlement. It SHALL NOT treat internal readiness as successful client connectivity or mask unexplained failure with retry or increased deadlines. | V-020; any behavioral amendment precedes projections and implementation. |
| R03-033 | Successful-child cleanup assurance SHALL observe exact delegation and service-close behavior directly. Filesystem or journal elapsed time SHALL NOT substitute for proof of service termination. Runtime cleanup authority, deadlines and failure policy SHALL remain unchanged. | V-021; existing NLSpec Sections 11.5 and 13.5; test-only correction and atomic selector rename, no behavioral amendment. |
| R03-034 | Renderer acquisition SHALL use one explicit accounting dependency boundary with native proof resolution and exact-container verification as the production defaults. Fake Docker fixtures SHALL supply matching fake accounting dependencies. Required placement failure SHALL prevent browser use and retain exact cleanup; no public option, environment bypass or policy relaxation is permitted. | V-022; existing REQ-856 invocation-accounting ownership and renderer acquisition/cleanup contracts; internal dependency injection only. |
| R03-035 | Native admission SHALL emit bounded, redacted decision and phase observations from its own evaluator; public critical-path host-admission paging SHALL preserve existing bounds/defaults and explicit coverage. | C-010; S-012/V-023 |
| R03-036 | Pressure retirement SHALL identify exact released residents and strictly reduce a persistent deficit without increasing another under identical weighted accounting. Claims remain until exact cleanup. | C-012; S-013/V-024 |
| R03-037 | All compiled consumers SHALL receive exact compatibility registration and idempotent terminal accounting; zero-consumer idle allocations retire promptly, otherwise useful compatible residency may remain. | C-012; S-013/V-024 |
| R03-038 | Historical S-014 candidate: profile/admitted/PostgreSQL minimum CPU with exact qualification budgets and matching claims/GOMAXPROCS/parallelism. REJECTED under E-194; its owner amendments, implementation and projections are restored together. This is not retained runtime behavior. | C-011; S-014/V-025; E-194 |
| R03-039 | Memory, process, port, service and host capacity quantities SHALL remain unchanged; utilization alone does not establish safe calibration. | C-013; S-012/V-023 |
| R03-040 | Test-fast benefit and browser non-regression SHALL use matched sources; CI SHALL run once for final correctness; elapsed time against the assumed 3,600,000 ms reference is descriptive only, without formal CI performance acceptance. Section 7.8 gates apply. | C-013; S-015/V-026 |
| R03-041 | Finalization, explicit selections, final CI, evidence audits, physical settlement, secret scans and final tracker lint SHALL precede completion. | C-013; S-016/V-027 |

### 7.2 S-001 interface and compatibility definition

R03-001 removes only reporting ownership. The following equality dimensions
define the proposed compatibility boundary. A public entrypoint such as an
owner slice versus an aggregate is not an additional partition key: equal
resolved selected semantics MUST compile identically.

| Dimension | Required equality / explicit default boundary |
| --- | --- |
| Logical stage and policy group | Same resolved stage, topology target, policy-group semantics and evidence context. Distinct policy groups cannot merge solely because services match. |
| Selector execution environment | Same project and configuration; exact file/title/scenario selections remain payload, not a sharing permission. |
| Runtime / renderer / trust | Same runtime profile and resolved renderer, trust and immutable provisioning identities. Existing owner-defined defaults remain; no new fallback is introduced. |
| Resources and fixtures | Same resource profile, fixture capability, PostgreSQL isolation and fixture profile. An omitted optional fixture profile has the existing absence meaning; it is not a wildcard. |
| Dependencies / preparation | Same schema-normalized service dependencies and preparation operations, preserving any meaningful operation order. |
| State / affinity | Same browser state scope and declared session/affinity. Isolated and measurement partitions remain row-specific singleton boundaries even after removing the owner field. |
| Invocation policy | Same failure/isolation policy, worker policy, startup/cleanup cost policy and effective timeout/headroom inputs. |

Unknown, malformed or missing required metadata MUST fail validation. Optional
fields use only their adopted defaults. Neither owner names nor filenames
supply missing compatibility evidence.

| Interface | Current | Proposed change / invariant |
| --- | --- | --- |
| Authored test-family row | owner_id, exact selector, explicit state/fixture/resource declarations | Retain all row ownership and verification fields. No metadata-only declaration is proof of cross-owner independence. |
| Browser partition projection | cartulary.browser_e2e_batch_manifest.v14 requires owner_id | Adopt a successor schema removing partition owner_id; remaining fields preserve their meanings. The successor is v15 at E-018; the complete consumer roster is recorded in E-019 before implementation. |
| Normalized browser partition | ownerID required by reader and cohort validation | Remove ownerID; resolve each selected row through the canonical catalog. Do not replace it with first-owner inference, optional-owner compatibility or a redundant owners array. |
| Executable cohort | Derived name/digest, exact selectedRowIDs/specs, policy and one worker | Recompute identities from revised compatibility and exact selected selectors; reporting ownership does not partition execution. Unselected rows never enter the cohort. |
| Work graph and results | owner_id=harness.browser; row_bindings identify selected catalog rows | Preserve execution owner and per-row ownership/verification attribution. No command-line or environment selection registry is added. |
| Public Make selection | Existing owner/row/suite selection | No new option or grouping mode. Direct and aggregate selection share the same compiler semantics. |
| Migration | Current schema/generator/readers and generated projections | Adopt owner clauses, migrate authored schema and consumers, regenerate through Make, then verify drift/schema/selection. Do not hand-edit generated outputs or translate old retained artifacts into new acceptance evidence. |

RB-002 includes final discovery of validator attachments, CLI selection,
reporters/normalizers, fixtures, generated artifacts and all partition readers.
The document does not pretend those consumers have been exhaustively audited.
The schema successor's allocation is an explicit adoption deliverable, not a
runtime compatibility fallback. Its compatibility identity MUST exclude reporting
owner and the available-but-unselected row roster, including indirectly through
a derived partition name. Given the same selected semantics, adding an unselected
row or changing catalog enumeration order MUST NOT change cohort membership or
identity. Physical lane placement may differ when the selected closure differs;
direct and aggregate equality is required only for equal resolved selections.

R03-003 uses this membership oracle, fixed before execution:

1. Resolve the request into the exact catalog rows and their complete native
   case selectors using the existing canonical identity rules. A row selecting
   several cases remains indivisible; a request for some rows is not permission
   to run only some cases of an admitted row.
2. For a successful invocation, the native executed-case set MUST equal that
   expected set, with exactly one accepted terminal case result and one correctly
   aggregated terminal row result. Owners and verification IDs MUST match every
   catalog binding. A native case cannot satisfy two rows accidentally.
3. For failure/interruption, every selected case MUST have either its independently
   settled result or an explicit non-success unexecuted disposition. No selected
   case may disappear, and no foreign case may execute. Repeated diagnostic events
   are allowed; conflicting or duplicate terminal results are not.
4. Empty legal resolved selections emit no product cohort or fixture allocation;
   invalid public empty/unknown/duplicate selectors retain the adopted command's
   rejection behavior. Positive cases include exact, partial, owner-only and
   mixed-owner selections; negative cases include omitted, duplicate, foreign,
   repeated-title/different-file and widened outcomes.

The model in Section 2 is an investigation population, not a mandate to reach
17 cohorts at any cost. If an owner's effects require an additional physical
partition, that requirement MUST be expressed by a reviewed execution/state
contract. A transient owner-name exception list is not an acceptable design.

R03-006 preserves the following adopted REQ-806 settlement rules. V-002/V-003
MUST demonstrate each row after the cutover; none is optional because a cohort
contains several reporting owners.

| Boundary / event | Required ordering and outcome |
| --- | --- |
| Affinity construction | Compute the exact selected chain for `(stage, browser_session_group, runtime_profile_id, resource_profile_id, fixture_profile_id-or-mutable)`. Reset dependencies MUST NOT cross incompatible tuples. |
| First selected cohort | Readiness precedes execution; no leading reset. Partial and owner-specific selections MUST rebuild the selected chain before inserting resets. |
| Adjacent selected cohorts | A reset follows the selected predecessor and precedes its selected successor. Its identity binds both cohort IDs. No filename boundary or authored reset-placement override may insert an additional reset within the cohort. |
| Functional lane reset | Restore validated migrated-template domain/bootstrap state, excluding sequence values and physical layout; empty and verify the exact lane bucket's mutation path; clear session storage, route-idempotency, fault and clock state. Preserve database name, ports, runtime profile and secret-bearing child environment. Do not drop the live lane database or recreate its bucket. |
| Product failure | Stop the cohort, keep independently settled outcomes, classify every unexecuted selected outcome as non-success, and quarantine the allocation. The successor reset still follows the predecessor; it may obtain a fresh clean allocation with a new concrete session evidence identity. No product retry is permitted. |
| Teardown failure or interrupted worker | Publish completion only after applicable fixture/suite settlement. Invocation-wide teardown failure fails the cohort and quarantines its allocation without erasing independently settled test outcomes. Missing outcomes remain explicit non-success. |
| Reset failure | Fail the required target, quarantine the allocation and prohibit prior product retries. Independent later work may acquire a fresh clean stack; it may not reuse uncertain state. |
| Cancellation, owner loss or uncertain cleanup | Retain exact resource ownership and resident claims until physical settlement is proved. No new borrower may use an uncertain generation. Existing owner recovery semantics and primary/secondary failure attribution remain authoritative. |
| Measurement / visual exclusions | Immutable measurement clones remain reset-free and dedicated with quiet admission. Snapshot updates retain exclusive output-write ownership; ordinary renderer reuse does not relax these rules. |

### 7.3 S-002 assurance preservation map

The mandatory correctness repair uses the existing survivor
`TestIncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance`
and adds Location assertions to its existing first-create and replay responses.
The proposed deletion of
`TestIncidentCreateReturnsStableLocationHeaderHTTPConformance` was implemented
and measured, then rejected by the user's E-053 performance rule under E-054.
The standalone test is restored. Final correctness-only routing preserves the
original five-symbol union and all fixture lifetimes; no efficiency improvement
is claimed. Existing response/request counts in the survivor remain unchanged.

| Assurance | Current source | Required survivor / disposition | Owner |
| --- | --- | --- | --- |
| Real authenticated create, 201, common envelope and exact Location | Standalone Location test in http_conformance_test.go | Named survivor's firstCreate; assert Location against returned incident_id | Core 01 REQ-01-167 |
| Replay 200, originally created resource and exact Location | Named survivor already checks response equality; replay Location is not asserted in the inspected body | Same survivor's existing replay; add exact Location without an additional request | Core 01 REQ-01-164/166 |
| Trimmed key/title and normalized replay | Named HTTP survivor | Retain its existing inputs and equality assertions | Core 01 REQ-01-157/160/166 |
| Actor-scoped transaction key | Named survivor's second actor | Retain second authenticated actor and distinct resulting incident | Core 01 REQ-01-166 |
| Divergent transaction rejection | Both replay tests | Retain both initially; optional later consolidation must preserve error contract and durable-state checks | Core 01 REQ-01-166 |
| Unicode NFC equivalence and canonical duplicate-key details | TestIncidentCreateReplayAndDuplicateKeyConflictUseNormalizedState_Integration | Retain this integration scenario unchanged in the first slice | Core 01 REQ-01-157/162 |
| Replay/conflict durable side effects and absence of duplicate writes | Same integration scenario and snapshot assertions | Retain snapshots and mutation/idempotency checks; never replace with response-only checks | Core 01 REQ-01-163/164/166 |
| Browser creation, directory/reopen, preferences and shell behavior | Existing auth-and-incident-directory and incident-administration scenarios | Preserve distinct browser contexts and assertions; remediation 02 already simplified preference setup | Product owners; no browser scenario removed by S-002 |

Affected catalog identities:

- `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` selected five symbols at baseline, including both named HTTP tests. The rejected optimization removed only the standalone Location symbol. The correctness-only disposition restores that symbol: the final two-row union MUST equal the original five-symbol union, with unchanged owner, verification IDs, fixture, profile and assurance context.
- The current public `test-slice` input contract selects rows, not a symbol within a row (E-016). To satisfy survivor-only execution without adding a CLI escape hatch, the S-002 routing change moves the survivor into one new exact-symbol row and keeps the original row with four other symbols, including the restored standalone test. The rejected optimization had retained only three siblings. The owner MUST allocate the new canonical ID and update all affected routing atomically at G-02. No selector may remain duplicated across the rows, and no obsolete alias is retained. Compatible aggregate execution MUST still use ordinary Go grouping; the split does not mandate another process or fixture.
- `module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6` remains unchanged in the first slice. Its distinct durable/Unicode/duplicate-key assertions MUST NOT be removed to meet an efficiency target.

The optional second step, replay-scenario consolidation, is DROPPED from this
execution scope: the durable/Unicode/duplicate-key assurance is distinct. It MUST
produce a complete revised map and independently runnable survivor before
deleting either scenario. S-002's first step does not depend on that decision.

### 7.4 Conditional investigations and preserved limits

Observed granularity matters: Go/Vitest packing estimates a catalog row even
when that row selects several tests; browser packing scales its estimate by
selected title count. Neither scheme measures actual setup or execution.
Within a rollback package fixture, the borrow mutex spans the test's connection,
transaction and settlement, so nominally parallel borrowers can wait on a
serialized critical section. Increasing workers cannot be assumed to increase
that fixture's throughput. S-003 MUST distinguish admission queueing from this
fixture wait and preserve the serialization unless separately proved safe.

S-003 MUST use an approved comparison record satisfying R03-009. Its minimum
fields are exact selected rows/stages, baseline and candidate source identities,
toolchain/environment/storage identities, cache and instrumentation modes,
lane/capacity settings, run ordering and repetition count, primary cost,
numeric acceptance/regression thresholds, variability method and missing-data
policy. The user fixed the numerical thresholds and experiment method in Section 8.2.
Baseline evidence, not the static model, must justify policy tuning.

Measure real process/container start counts, template preparations, clone/drop
counts, resets, worker starts, teardown and publication work alongside wall
time, attributed CPU, observed memory and resource waits. Existing CPU readings
may be lower bounds and sampled RSS may omit processes or count shared pages
more than once. These limits MUST survive comparisons. The user-authorized S-008 adds complete charged-resource accounting through
the existing evidence owner; E-030 accepts that capability for the declared
profile. Existing partial samples retain their narrower meaning. The inspected REQ-835
clarification distinguishes validity of published records from completeness of
optional collection. A valid but partial diagnostic result does not establish
the measurements needed for R03-009/R03-010 performance acceptance.

Current defaults remain four browser lanes, maximum six, one worker per cohort,
one active isolated Vitest file worker at a time per cohort, zero product retries, existing
resource claims and declared timeout policies. This handoff does not introduce
new Go CPU budgets, concurrency flags, deadline overrides or cost constants.

For S-004, transaction proof includes all nested helpers and nontransactional
effects, not just the direct DB argument. Browser proof includes shared listings,
administrator state and background effects, not just unique row identifiers.
A failed proof retains the existing fixture declaration without a fallback path.

For S-005, "settlement between borrowers" means retirement of borrower-owned
databases, connections, work and namespaces while the physical service may remain
resident. It does **not** mean terminate the physical service between every
borrower, which would remove the proposed benefit. Final use, eviction or
quarantine MUST request physical retirement through the existing proven owner;
uncertain or failed retirement retains quarantine, authority and resident claims.
Uncertainty does not authorize killing an unproved resource. Dedicated cluster
exclusions and current REQ-804 govern until adoption.

R03-013–015 use the approved bounded design: at most one reusable exclusive
generation per run including active, idle, quarantined and retiring states;
ordinary non-browser borrowers are sequential. Retire immediately when no
compatible borrower is ready, a different identity is needed, quiet measurement
becomes eligible, or the run ends. No idle timer/cache or cross-run reuse exists.
The runtime default remains fixture-bound lifetime until explicit candidate
eligibility, owner adoption and acceptance are established. Limits MUST
fit unchanged aggregate capacities and MUST NOT starve eligible quiet measurement
work. Each sequential borrower receives a fresh fenced logical identity and fresh
database cloned from a sealed immutable template; reuse never reruns borrower
DDL against that template. The identity key covers engine, configuration,
extensions/roles, template/migration provenance and fixture profile as applicable.
All borrower handles, prepared work, background jobs, advisory locks and namespaces
must settle before reissue; a fresh database name alone is insufficient.

Checkpoint-subject tests remain dedicated. Ordinary cleanup may itself induce
cluster-wide checkpoints: S-005 MUST preserve that cleanup/durability behavior
and measure its effect rather than suppress it to obtain a faster result.

### 7.5 Slice entry and completion table

The user has authorized the following implementation kinds, subject to their
source-proof, adoption, assurance and performance entry gates. Current
readiness entries below identify evidence to obtain, not a new permission request.

| Slice / kind | Findings / contracts | Workflow dependency | Entry conditions and readiness | Completion / validation | Rollback or migration |
| --- | --- | --- | --- | --- | --- |
| S-001 proposed behavior/interface change | F-001; C-001/C-002/C-004/C-005 | WF-01 | WF-08/WF-08a complete, E-030/E-038. REQ-806/AC-088 adopted; source-effect dossier complete under E-032/E-039 before candidate product execution. E-049 accepts G-03/G-04; S-007 final integration validation passed under E-071/E-072 | R03-001–006/009–011/016/017; G-03 and G-04; V-002/V-003/V-005/V-006. At least one preregistered compatible population has fewer actual cohort/setup boundaries and passes the structural non-regression gate | Coordinated version cutover; settle active old ownership before source/schema/projection rollback; no mixed epochs, dual mode or old-run reinterpretation |
| S-002 test restructuring | F-002; C-007 | WF-02 | E-050 characterized both original rows; E-055 closes required assertion/routing correctness | R03-007–010/017; V-004 passes. E-054 rejects removal through G-04; restored final source has all five original symbols and no fixture reduction | Source and routing restored together, then mandatory assertions/exact row retained as correctness-only repair; further replay consolidation remains DROPPED |
| S-003 measured policy investigation | F-003; C-006 | WF-03 | Await WF-02 disposition; characterize actual selected work and resource claims before proposing a measured policy change | R03-009–011/017; G-03/G-04 for any policy change, or supported measured no-change result; V-005 | Revert rejected authored policy changes and regenerate normally; no permanent compatibility tuning layer |
| S-004 eligibility investigation | F-004; C-002/C-003 | WF-04 | Await WF-03 disposition; screen and rank candidates, then complete RB-003 effect proof for each proposed conversion before G-02 | R03-009–012/016/017; G-03/G-04 and V-003/V-005/V-007 for each candidate, or retained stronger isolation | Restore the stronger authored fixture boundary if proof fails; no fallback auto-detection |
| S-005 lifetime investigation/change | F-005; C-003/C-004/C-005 | WF-05 | Await WF-04 disposition; measure remaining startup cost under dedicated lifetimes, then resolve RB-005 only if the opportunity supports the bounded candidate | R03-009–011/013–017; G-03/G-04 and V-003/V-005/V-007, or supported dedicated-lifetime decision; source-level cost opportunity alone is insufficient | Settle old ownership before switching epochs; revert owner/projections/implementation together; no uncertain generation may be reused |

### 7.6 Stage gates and decision records

S-008 is a user-authorized prerequisite, not an optimization candidate. Its
capability design is developed under R03-018 and adopted in the NLSpec before
producer/reader changes. Prefer a fresh kernel accounting hierarchy containing
the measured Make invocation, all descendants and every owned service from
creation through retirement. Preserve the hierarchy until final CPU counters,
memory peak and empty membership have been read. Keep the measurement controller
outside that hierarchy. Do not change capacities, deadlines or service ownership.
Prove Docker membership in the same hierarchy rather than infer it from endpoint
names, operating-system strings or matching kernel versions.

The measured child Make command must include its initial startup, preparation,
execution, publication, scan and terminal exit. Report controller overhead and
scope exclusions explicitly. A fresh cgroup's charged memory peak is a defined
metric, not RSS or an estimate of all shared cache residency. The harness's own
ready/start and cleanup events supply queue delay and cleanup duration; Linux
run-queue latency is a different quantity. Required missing coverage fails closed.
Public Make selection, secure typed retention, exact source/run binding and
existing validated readers remain the integration boundaries. No ad hoc test
runner, automatic historical discovery or changed-source comparison waiver is
authorized by this measurement workstream.

eBPF is an evaluated alternative for missing lifecycle or scheduling detail,
not a prerequisite. It requires separately proved privileges, supported hooks,
bounded collection, loss accounting and observer overhead. CPU/memory aggregates
should use existing kernel counters when they provide complete ownership and
lifetime coverage. Probe results do not establish capability acceptance.

These gates separate investigation entry from implementation readiness and
acceptance. They implement the R03 requirements; they do not grant authorization.

| Gate | Entry / required evidence | Binary decision and owner |
| --- | --- | --- |
| G-01 Investigation | WF-00 revalidation; named candidate/population; scope for static discovery or separately authorized baseline execution | Semantic and harness owners admit the bounded investigation or record deferral. Missing candidate runtime results do not block discovery. |
| G-02 Implementation readiness | Reviewed source-effect/assurance map; exact validation routes; required owner adoption and schema/pool design; baseline characterization; fixed V-005 protocol. Candidate-specific source identity is bound before its first run. | Each affected semantic owner and harness contract owner records every applicable entry predicate met; implementer records separate scoped implementation authorization. A missing predicate is blocked. |
| G-03 Assurance acceptance | Implemented candidate; all selected functional, isolation, selection, failure, cleanup and projection cases required by V-002/V-003/V-004/V-006/V-007 | Zero required failures, missing cases, lost assertions or unresolved candidate ownership. Correct expected-failure fixtures must prove the specified failure disposition; they are not product failures to waive. |
| G-04 Applicable performance acceptance | G-03 met; independently validated first two pairs for each workload and the Section 8.2 decision rule | Both first-two-pair elapsed improvements <5%: reject the candidate. Either >=20%: remaining performance protocol is optional, with no omitted confidence-bound claim. Otherwise all four pairs must pass: S-001/S-002 <=5% guardrail regression; S-003/S-004 >5% benefit; S-005 >10%, with <=5% guardrails. Missing required observations are inconclusive. |

No gate may require its own output as an entry condition. Baseline measurements
can justify building a candidate; candidate measurements establish acceptance.
G-02 does not require G-03/G-04. Partial optional diagnostics can pass canonical
validity while leaving G-04 unmet. Deferral is a terminal workflow disposition,
not accepted implementation: the receipt MUST state whether candidate changes
were never made, reverted through the canonical path, or remain explicitly
unaccepted for later work. Required assurance failures remain visible.

### 7.7 WF-11j stability plan for published PostgreSQL connectivity

**Status and authority.** E-147 records the user's request to plan a stable
solution. E-148 records explicit implementation authorization. This section is the approved
execution contract; behavioral amendments still require prior NLSpec adoption. R03-032, F-019 and V-020 govern traceability.
The behavioral owners remain NLSpec Sections 11.4, 13.7 and 14, and the existing
fixture/accounting clauses. Any changed runtime or diagnostic contract SHALL be
adopted there before schema, policy, implementation or test changes. Executable
tooling SHALL NOT consume this tracker. No new product-domain concept is proposed.

**Observed boundary.** E-144 establishes internal PostgreSQL readiness and a
failed published client endpoint, with exact allocation identity and completed
cleanup. E-146 establishes successful unchanged focused and containing selections.
Neither proves the historical forwarding cause or removes its uncertainty.
A passing Docker API request or visible port mapping SHALL NOT establish client
reachability; TCP connection alone SHALL NOT establish authenticated SQL readiness.

**Recommended decision.** Qualify the transport boundary before more performance
runs. Preserve the required Windows/WSL2/Docker Desktop profile and the existing
fixture owner. First reproduce the failure in bounded fixture work, then adopt
only a demonstrated correction. The first candidate to evaluate is daemon-managed
dynamic publication of an exact approved port set, replacing explicit ephemeral
bindings only where evidence shows that path causes failure. It is a conditional
candidate, not an already proven fix or a global PublishAllPorts setting.

| Option | Disposition and rationale | Compatibility and unresolved risk |
| --- | --- | --- |
| Existing publication plus exact transport qualification | Required characterization control; exposes whether failure belongs to client routing, publication, service readiness or ownership | No source compatibility change; historical host failure remains open until appropriately qualified, and its exact cause may remain unavailable |
| Daemon-managed dynamic publication for the exact owned PostgreSQL port set | First conditional repair candidate; retains dynamic allocation and existing physical ownership without allocating fixed host ports or adding test retries | Requires effective-request proof and explicit port allowlist enforcement; cannot be accepted from an upstream anecdote or a happy-path pass |
| Larger deadlines, readiness retries or endpoint fallback | Rejected for this workstream; none corrects a broken forwarding path and each can conceal or prolong failure | Existing deadlines and no replacement after entering SQL readiness remain binding |
| Fixed host ports or a reserve-close-bind allocator | Rejected; creates collision, race, host-state and parallelism obligations | No host port range, sysctl reservation or process-global port broker is introduced |
| Direct container IPs, host networking, custom TCP tunnels or a containerized replacement runner | Not selected; adds routing/TLS identity, isolation, lifecycle or accounting responsibilities disproportionate to the evidence | No implicit fallback to another network or execution architecture |
| Native Docker Engine in WSL or a separate Linux host | Separate platform proposal only; current reference qualification explicitly requires Docker Desktop | Requires owner-profile adoption and fresh platform qualification; SHALL NOT be installed alongside the integrated daemon as an ad hoc repair |
| Host/WSL/Docker update or restart | Conditional operational remediation only when inventory and an applicable upstream fix support it | Requires a concrete impact/rollback plan and explicit approval before interrupting shared host services; changed host identities invalidate the experiment freeze |

**External evidence — non-normative.** Docker documents that Desktop forwarding
is mediated by its host backend and VM, rather than an ordinary local daemon
listener ([network architecture](https://docs.docker.com/desktop/features/networking/)).
The [upstream issue](https://github.com/docker/for-win/issues/14195) reports
intermittent WSL failures with explicit random port mappings and a different
observed outcome for publish-all; it does not establish a supported universal fix.
Docker also warns that integrated Desktop and a directly installed WSL Engine can
conflict ([WSL setup](https://docs.docker.com/desktop/features/wsl/)). These sources
motivate the pilot; repository requirements and measured results decide adoption.

#### Phase A — WF-11j-a: characterize the actual transport boundary

The implementation SHALL use the existing acquisition owner and public Make
selection surfaces. Diagnostic changes SHALL extend the current typed failure
observation, with an atomic schema migration where needed. No private executable
runner or replacement canonical-event reader SHALL be added.

Before execution, freeze the exact image, engine/Desktop/WSL versions, network
mode, source, requested and effective publication shape, selected rows, resource
profile and attempt roster. Capture only exact owned allocations. Private proof
MAY retain the endpoint required for the observation while that allocation is
owned; public records SHALL retain opaque identities, closed outcomes and explicit
coverage. Raw credentials, full host logs and unrelated process/network inventory
SHALL NOT be retained. Best-effort observations SHALL preserve the current bounded
diagnostic budget, primary failure and cleanup authority. Unknown or contradictory
identity SHALL produce unavailable evidence, not an inferred transport diagnosis.

| Observation | Interpretation and acceptance boundary |
| --- | --- |
| Image exposure, requested ports and effective Docker create request | Prove the intended publication operation and exact allowlist; a helper option alone is insufficient |
| Exact container state, current bindings and generation | Distinguish missing/replaced ownership from unchanged but unreachable publication |
| Internal listener / final-server marker | Service-side observation only; does not establish client authentication or routing |
| Selected-host resolution and TCP outcome | Client transport observation; preserve each actual failure stage without raw endpoint values |
| TLS and SQL handshake outcome | Ready binding requires the existing successful security/capability checks on the selected endpoint |
| Release/removal and terminal accounting | No new acquisition, reuse or performance sampling while owned resources remain unresolved |

The reproducer SHALL create only its own sockets/allocations, exercise controlled
publication pressure within existing capacities, and retain failed cases as
failures. It SHALL NOT exhaust unrelated host ports, mutate firewall/sysctl state,
restart Docker/WSL or infer ownership from PID/name resemblance. Negative fixtures
SHALL cover internal-ready/published-refused, TCP-connected/TLS-or-SQL-failed,
changed bindings, stale container identity, expired acquisition, unavailable
observation and cancellation. Actual allocation attempts SHALL be separately
identified; repeating a test after failure SHALL NOT erase the failed result.

Exit: the request transformation and failure boundary are observable, deterministic
negative tests pass, selected affected rows pass, and the runtime reproducer has
an explicit outcome. If the original failure cannot be reproduced, that absence
SHALL be recorded; a speculative transport replacement SHALL NOT be labeled a fix.

#### Phase B — WF-11j-b: conditional publication repair

Before changing pilot code or projections, adopt the bounded publication amendment
and any versioned interface changes in the NLSpec. Rejection SHALL restore that
owner amendment together with its projections and implementation.
For the proposed daemon-managed publication pilot, the internal fixture request
SHALL declare the exact service ports and permitted bind exposure. Omission SHALL
retain the adopted current policy until the candidate is accepted; this proposal
introduces no user-facing mode, compatibility switch or automatic host detection.
The pilot SHALL use one path for its admitted PostgreSQL allocations, not retry
one path and fall back to another. Renderer and other fixture publication contracts
SHALL remain separately owned.

The existing `testcontainersx` provider/client boundary is the appropriate place
to validate effective Docker requests. The inspected pinned Testcontainers v0.42.0
applies `mergePortBindings` after `HostConfigModifier`; merely clearing bindings
in that modifier does not select a different wire request. Any adapter SHALL use
the supported provider/client API, preserve installed-only image policy and
accounting hooks, and validate the final Docker request. No vendored dependency,
private provider access or second startup implementation SHALL be introduced.

A publish-all candidate SHALL reject image/config exposures outside the exact
allowlist before container start, verify the resulting binding set before ready
publication, and preserve the approved interface exposure. It SHALL NOT publish
additional image-declared ports. Unexpected bindings or stale identities SHALL
fail closed and retain the normal exact cleanup owner. Each borrower SHALL still
receive the original security, fixture, isolation and generation guarantees.

Acceptance requires the controlled publication reproducer to distinguish the
candidate from the failing control, plus deterministic tests of effective request
shape, extra image ports, IPv4/IPv6 bindings, provider errors, cancellation at
create/start/ready, installed-only preparation, accounting placement and failed
removal. If these conditions cannot be met cleanly, reject the candidate and retain
the diagnostic work. No speedup or stability claim is established at this stage.

#### Phase C — WF-11j-c: stability qualification

Before qualification, freeze three independent invocation rosters of 64 fresh
PostgreSQL allocation lifetimes each. Each roster SHALL contain 16 serial
lifetimes, 32 ordinary concurrent lifetimes within the existing admission/resource
limits, and 16 lifetimes with competing owned service allocation/release. This
work SHALL run outside quiet performance measurement. Every lifetime SHALL use
the normal acquisition/SQL-readiness/removal path and have its own outcome; failed
lifetimes SHALL NOT be replaced or retried. Synthetic failure-injection tests are
separate from this positive qualification roster.

All 192 positive lifetimes SHALL pass security/capability readiness and exact
settlement. There SHALL be zero unexpected acquisition failures, stale identities,
missing outcomes, unresolved resources or incomplete required diagnostics.
This finite qualification establishes observed reliability for the declared
conditions, not a zero future-failure probability. A failure blocks qualification
and returns to characterization with its evidence preserved.

Run the exact four affected module.extensions rows and the containing service
selection on the retained source. Where fixture code changes, apply the identical
correction to baseline and lifecycle candidate and validate both. Independently
validate each invocation, its secret scan and physical settlement. The mandatory
full CI gate remains required and SHALL NOT be replaced by this fixture workload.

If no publication defect is reproduced and no correction qualifies, the default
is to retain existing behavior, record the causal limitation, and require the
unchanged path to pass the same qualification. Repeated transport failures require
a concrete environment remediation proposal; they SHALL NOT trigger additional
full CI runs, weakened tests or an unbounded search for a passing attempt.

#### Phase D — WF-11j-d: reconciliation and experiment handoff

Reconcile the already adopted owner amendments and atomically migrated
diagnostic/request contracts with the retained implementation. Regenerate through
Make and run focused owner checks. Run
`make agent-finalize` before broader end-of-work verification if implementation
changed. Complete each child workstream with a tracker receipt and successful
`make lint-markdown` before its successor.

Freeze the retained source and common B/C treatment map only after qualification.
Resume Section 8.5: measured matched browser comparisons remain unchanged; CI uses
the assumed 3,600,000 ms reference, no baseline CI timing runs, descriptive resource
metrics and mandatory correctness/cleanup. Keep the lifecycle optimization separate
from any common transport correction. A rejected publication pilot SHALL restore
its owner amendments, implementation and projections together before freezing.

Handoff SHALL state the causal finding or unresolved historical uncertainty,
accepted/rejected/no-change transport decision, qualification roster/results,
exact source and environment identities, rollback outcome and physical settlement.
S-007e owns final integration assurance and overall completion. No transport
investigation, successful diagnostic reader or planning receipt closes that gate.

Execution disposition: E-149 accepts the diagnostic extension; E-150 retains
publication unchanged; E-152 passes all 192 qualification lifetimes; E-154/E-157
complete reconciliation and the final experiment freeze. E-160/E-161 reject the
lifecycle optimization. E-166/E-167 complete retained-source integration and
handoff. The historical transport failure remains causally unresolved; no
environment repair or zero-future-failure guarantee is claimed.


### 7.8 Reservation sizing and idle-stack retention continuation

This user-authorized continuation preserves all earlier decisions, including the
rejected fused reset/cohort optimization. NLSpec adoption precedes changed
projections. Runtime tooling MUST NOT consume Markdown. The scheduler ranking,
fairness, browser lanes/workers/retries, isolation eligibility, PostgreSQL pooling,
measurement quietness and public test commands remain unchanged.

#### 7.8.1 Admission observations and claim audit — S-012

`make explain-run RESULTS_DIR=<exact-run> DETAIL=critical-path
VIEW=host-admission JSON=1` extends the existing reader. TARGET defaults to the
invocation target, UNIT_ID omission selects the complete path, OFFSET defaults
to zero, LIMIT defaults to 20 and permits 1..100. Existing exact-run selection,
usage/artifact errors, 262,144-byte machine limit, bounded human output and
record-boundary pagination apply. Version changed schemas atomically; no legacy
translation reader or new executable diagnostic command.

The native evaluator SHALL expose opaque request/unit identity, queue position,
mode/clock, active claims, effective capacities, resident/prospective allocations,
current and persistent deficits, required ancestors, relevant holders, priority
and exclusive blockers, and retirement proposal/settlement. Host phases distinguish
logical-slot waiting, transaction, admission waiting, pressure retirement and
uncovered time. Decision snapshots are not continuous causal knowledge. Concurrent
blockers and overlapping phases MUST NOT be summed as exclusive causes.

Use an independently bounded section: 16,384 records, 16 MiB, 65,536 bytes per
record. Identical consecutive decisions coalesce. Omission, truncation and
unavailable coverage are explicit. Diagnostics do not change product outcomes
or cleanup authority; public records contain no lease capabilities, credentials
or recovery paths. No eBPF collector, automatic tuner or historical discovery.
Map every claim to its lifetime owner. Numeric non-CPU quantities are NO_CHANGE.

#### 7.8.2 Exact retirement and terminal consumers — S-013

Evaluate the pending request with current residents and again after removing
exact proposed residents, retaining required ancestors and replacement demand.
An eligible pressure retirement strictly reduces at least one persistent deficit
and increases none. Preserve non-additive volume capability semantics. Include
an exclusive PostgreSQL generation only when this cleanup settles it; exclude
shared services, measurement generations and unrelated allocations.

Consider idle renderers and application stacks together: no remaining compatible
consumer first, then useful pressure relief with stable logical-key ties. Slot
exhaustion remains renderer-specific. Revalidate idle/generation identity; retire
one candidate, settle, then reevaluate. Uncertainty retains claims and recovery.
Register exact compiled compatibility; terminal notifications cover passed,
failed, skipped, cached, cancelled and never-started work, are idempotent, and
reject unknown/contradictory facts. Zero-consumer idle allocations retire through
admitted cleanup. Others may remain when capacity permits. Remove static release
markers and incomplete counters; add no TTL, cross-run cache or owner exception.
Separate reset units and all reset/state/measurement boundaries remain mandatory.

#### 7.8.3 Go budget candidate — S-014

**Disposition: REJECTED / RESTORED, E-194.** The following describes the historical
candidate design, not an adopted requirement for the retained implementation.
Its correctness tests passed, but both initial matched performance pairs were
slower. The prior Go owner and projections were restored together.

Resolve ordinary CPU as min(authored profile CPU, admitted Go budget, applicable
PostgreSQL bound). Positive/infeasible validation precedes child execution.
Standard therefore normally uses one CPU; larger authored profiles retain their
budget. Explicit qualification contexts retain exact owner-required budgets or
fail feasibility; no silent clamp. CPU claims, GOMAXPROCS and test parallelism
agree; package-build parallelism respects the budget and existing package cap.
Unrelated selection splitting/order/count MUST NOT change a group's budget.
Preserve exact selectors, compatibility, isolation and Go JSON ownership. Remove
global execution worker/GOMAXPROCS fields; publication workers remain separate.
Version/migrate all affected consumers; budgets bind semantic/cache identity.

#### 7.8.4 Assurance, measurement and rollback — S-015

Characterization SHALL cover diagnostic pagination/bounds/malformed/private data;
transient/persistent/nested/multidimensional pressure and masked or replaced
savings; direct/owner/partial/aggregate/empty selections; terminal duplicates,
skips/caches/cancellation and last-consumer races; active borrowers, stale tokens,
owner loss, publication/cleanup failure, quarantine and exclusive progress;
Go parent/host/service bounds, explicit contexts, selector and identity invariance.

Discover exact rows through public task-guide for harness.browser,
harness.command_surface, harness.test_catalog and harness.evidence_accounting.
Before a new S-015 source freeze, each changed Go treatment SHALL pass
`make lint CHECK=go` after finalization, in addition to its affected assurance and
projection checks. This execution entry gate follows the E-184 late static-analysis
failure; it does not replace final CI or alter runtime semantics.
Measured baseline contains diagnostics and mandatory lifecycle repairs; candidate
adds only Go budgeting. Earlier CI is motivation, not an acceptance sample.
Freeze sources, selections, allowed differences, toolchain/storage/environment,
run order and coverage. Use CGO_ENABLED=0, CARTULARY_HARNESS_CACHE_MODE=off,
HARNESS_DIAGNOSTICS=basic, HARNESS_ACCOUNTING=required. Warm build/dependency caches;
exclude one warm-up per measured treatment. No intentional competing workload or
polling during quiet measurement.

| Workload | Acceptance |
| --- | --- |
| make test-fast | Four matched pairs maximum; first two both below 5% reject; either at least 20% permits early exit after both; otherwise 95% benefit lower bound exceeds 5% |
| make browser-e2e SUITE=functional | Four matched pairs; correctness/cleanup and all five 99% non-regression upper bounds at most 5%; no benefit minimum |
| make ci | One final correctness invocation against the retained source, after all cheaper gates. No baseline, warm-up or formal CI performance gate. Elapsed time against assumed 3,600,000 ms and other metrics are descriptive only |

The user-approved E-191 amendment supersedes earlier CI sampling requirements for
this continuation only. Complete focused correctness and projection checks first,
then test-fast warm-ups and benefit pairs. Run browser performance warm-ups and
four pairs only if the Go treatment survives the fast gate. Rejecting that gate
SHALL omit remaining optional performance sampling, restore the Go treatment
atomically and proceed to retained-source final assurance. Neither focused
repairs nor failed prior warm-ups authorize another CI warm-up. Final CI failure
SHALL remain open, be diagnosed and repaired through focused routes before any
necessary repeat; one planned invocation does not waive correctness.

Matched pairs alternate baseline-first odd and candidate-first even. Full analysis
uses 10,000 paired bootstrap resamples, seed 20261021 and existing decimal rules.
Measured guardrails are elapsed, charged CPU, kernel peak memory, p95 ready delay
and cleanup duration. Required correctness and cleanup never become optional.
Missing coverage is inconclusive. Every attempt is independently validated;
changed-source evidence is never COMPARISON=equivalent; failed/omitted attempts
remain recorded and repairs require a fresh freeze. Report attribution and claim
scope honestly. Reject optional Go tuning if a required performance gate fails;
restore owner/code/projections together. Mandatory lifecycle correctness repairs
and accepted diagnostics remain; do not reintroduce defects for historical timing.

#### 7.8.5 Final completion — S-016

Run make agent-finalize before broader final verification, then generated/schema/
catalog and owner checks, explicit browser support/visual and UI-review contracts,
one final make ci, independent evidence audits, final secret scans and physical
settlement. No golden refresh. RESULTS_DIR maintenance requires a compatible
successful full warm-check root; otherwise record skipped because RESULTS_DIR
was unset. Every receipt records source, traceability, commands/selections/roots,
native exits, cleanup, separate assurance/performance decisions, rollback and
next entry. Final tracker update plus passing Markdown lint closes this scope.

#### 7.8.6 Focused repair after user cancellation — S-016a

User instruction at E-198 stops the current CI and authorizes the identified
fixes. WF-16a SHALL follow the settled cancellation receipt and boundary lint.
Core 03 REQ-03-314/315 already own destination presentation and focus completion;
no behavioral amendment is needed for this test-projection correction. The test
SHALL await destination visibility, Work dock closure, and focus inside the
destination before issuing the next serial navigation. Both repeated transitions
SHALL use one local helper. Deliberate supersession tests SHALL remain separate
and SHALL NOT acquire that serial wait. Defaults, timeouts, selectors, assertions,
production behavior, capacity and retry policy SHALL remain unchanged.

V-027a SHALL run the failing 768-pixel row alone, then all six destination-handoff
rows plus the three other rows skipped in its CI cohort. Finalization SHALL
precede broader source type/lint/generated/catalog checks. Each run SHALL have
independent evidence validation, successful cleanup and retained-secret scanning.
The repair receipt SHALL record source identity, exact selections and native
exits, update this tracker, and pass Markdown lint. This repair does not close
S-016's incomplete full-CI assurance or establish CI performance acceptance.
No automatic CI restart is part of the focused repair after this cancellation.


## 8. Validation Plan

### 8.1 Future canonical validation

The implementation request authorizes the following validation scope.
All repository operations run from `/home/jochi/code/cartulary` through
public Make. Selection placeholders are explicit future discovery gates, not
runnable commands with invented row IDs. RB-006 MUST close by recording literal
selections before execution. Use fresh literal run IDs, capture native exit
separately from summaries, and retain exact source/environment identities.

| Validation / stage | Contracts and slices | Required cases / public selection | Required evidence and failure consequence |
| --- | --- | --- | --- |
| V-001 pre-change discovery | C-001/C-007; S-001/S-002 | `make help ALL=1`; `make task-guide OWNER=harness.browser`; `make task-guide OWNER=harness.command_surface`; `make task-guide OWNER=harness.test_catalog`; `make task-guide OWNER=harness.evidence_accounting`; `make task-guide OWNER=module.incidents` | Freeze exact row populations and validation routes. Unknown selection blocks dependent execution; do not invent a private runner. |
| V-002 pre/post grouping | C-001/C-002/C-005; S-001 | `make test-slice OWNER=<frozen-owner> ROWS=<frozen-rows>`; `make lint CHECK=types` for authored browser TypeScript; compiler/adapter cases for exact/partial/owner subsets, repeated titles, mixed owners, incompatible tuples, deterministic identity, predecessor schema rejection, oversized rows, failure and teardown | Exact native outcomes, bindings, digests and schema results. Missing/duplicate/widened execution or stale success blocks acceptance. |
| V-003 pre/post lifecycle | C-002–C-005; S-001/S-004/S-005 | Owner-discovered test-slice rows for contexts/actors, the independence matrix below, concurrent borrowers, reset/cancellation/owner-loss, namespace safety, rollback/close/drop failure and quarantine | Logical and physical settlement plus row outcomes. Uncertain resource ownership remains charged and quarantined; no retry to obtain a pass. |
| V-004 incident characterization and survivor | C-007; S-002 | Before change, run the two known incident rows listed in Section 7.3 through owner-routed test-slice. After the G-02 routing cutover, run the allocated exact-symbol survivor row alone, each retained row alone, and their combined selection; record literal ROWS values before execution | Source inspection proves the assertion map; native Go JSON proves test execution/outcomes, not individual assertion execution. Exact selector-union equality, one fewer fixture lifecycle, preserved response/actor/durable assertions and independent survivor execution are required. Missing proof blocks removal. |
| V-005 matched efficiency comparison | C-006; all activated optimization slices | `make test-slice OWNER=<frozen-owner> ROWS=<frozen-rows> HARNESS_DIAGNOSTICS=basic`, or the frozen browser suite; `make explain-run RESULTS_DIR=<exact-run-root> DETAIL=performance JSON=1`; `make harness-observability-check RESULTS_DIR=<exact-run-root> CHECK=evidence` | Preregistered population/limits, exact runs, lifecycle costs, wall/CPU/memory/waits and omissions. Reader success is not execution or performance acceptance; unknown/within-noise benefit remains unaccepted. |
| V-006 future implementation final gate | All changed contracts | make agent-finalize before broader final verification; authored projection changes use make generate, make source-check CHECK=generated and make source-check CHECK=schema; required owner-routed aggregate gates and discovered explicit affected selections; make ci only when required by the affected validation closure | Source drift/schema, required regression, native exits and complete cleanup. Visual is explicit: make browser-e2e SUITE=visual when affected. Review contracts use make harness-contract SUITE=ui-review when affected. No golden update implied. |
| V-007 optional eligibility/lifetime gate | C-002/C-003; S-004/S-005 | Candidate-owned test-slice and harness lifecycle rows after exact discovery; permitted Go-only order check uses a recorded positive GO_SHUFFLE_SEED | Transaction escape, independent commits, mutable global state, cluster residue, stale generation, forbidden overlap, failed reassignment and physical cleanup cases. Failure retains prior isolation/lifetime. |
| V-008 document review / handoff | C-008; all slices | Static review of this document, links, requirement IDs, dependency graph, traceability, claims and source baseline; no repository test command | Every reference resolves, each proposal has an acceptance predicate, and all unexecuted or historical results are qualified. Incomplete evidence stays explicit. |
| V-009 measurement capability | R03-018; S-008; RB-009 | Public Make capability/lifecycle tests and an exploratory measured support workload after owner adoption; independently validate retained output through existing readers | Account short-lived children and owned service startup/retirement; reject foreign/missing scopes; verify nested commands, cancellation, failed publication, counter integrity, empty membership and primary exits. Full guardrail export is available without truncated row lists or partial resource claims. |
| V-010 complete diagnostics | R03-019/020; S-009 | Discovered evidence-accounting and command-surface owner slices; schema/generated checks; exploratory browser and CI pilots after agent-finalize | Pagination beyond 20, target ties/filters/empty/byte limits, reason reconciliation, overlaps/uncovered time, every retirement reason, limits, malformed/unsafe/absent diagnostics and read-only behavior. Pilots are excluded from acceptance. |
| V-011 admitted preparation | R03-021/022; S-010 | Discovered browser, command-surface, catalog and evidence-accounting routes | Direct/owner/partial/aggregate/singleton/empty compiler selections; digest/bounds/kinds/deadlines; forced-pressure continuity; cancellation at transitions, owner loss, reset/product/publication/cleanup failures; fresh contexts, measurement quietness and visual ownership. |
| V-012 lifecycle experiment and closure | R03-023/024; S-011/S-007e | Section 8.5 browser and CI; independent audits; final explicit suites and CI | Both workload dispositions, retained source, current verification, final scans and no unresolved owned resources. |
| V-013 inherited-admission correction | R03-025; F-012 | Existing native runtime and exact harness.browser.integration.native_nested_cancellation routes | Proven parent binding survives preparation protection; local binding wins; absent binding cannot be forged; nested cancellation reaches its intended lifecycle transitions and settles without changing scheduling or test deadlines. |
| V-014 startup-budget correction | R03-026; F-013 | Existing platform.postgres fixture lifecycle, browser pgtest/service lifecycle, evidence-accounting helper and affected revisions store rows | Attempt budget, probes, terminal failures, cleanup authority and ordinary shutdown are explicit; current diagnostics reject retired/contradictory fields; affected real services and generated/schema/catalog checks pass. |
| V-015 fixture publication | R03-027; F-014 | Exact failing module.assessments browser row, containing assessment selection, retained-authoring owner and existing socket-monitor rows | Seeded IDs are observed through the accepted incident socket before draft authoring; focused and containing assertions pass; no product behavior or selected population changes. |
| V-016 negative-test static analysis | R03-028; F-015 | Public `make lint CHECK=go` and exact PostgreSQL startup-test catalog row in both sources; finalization/projection checks | SA1012 remains enabled globally, the intentional nil input remains tested, both treatments pass, and the common delta is identical. |
| V-017 mention refresh recovery | R03-029; F-016 | Exact creation-recovery browser row alone and containing module.entities recovery selection, mention owner/reconciliation/continuity unit rows, public types/generated/catalog checks | Both rejected and uncertain branches prove controlled read failure, retained acknowledgement, explicit successful read recovery, and no extra creation or resolution; both sources contain the identical correction. |
| V-018 linked metadata precondition | R03-030; F-017 | Exact Space-discovery browser row alone; containing linked-preview and metadata rows; metadata presentation/query and discovery owner tests; types/generated/catalog | Both metadata items are admitted before virtual time advances; all original assertions survive; identical B/C correction and passing audits/settlement |
| V-019 cleanup budget | R03-031; F-018 | Common cleanup unit/integration tests; affected incidentbundles and recovery rows; containing owner routes; generated/schema/catalog as changed | One bounded owner lifetime, shorter-parent clamp, bounded admission, one destructive operation, independent failure attribution and quarantine; affected service checks pass in both sources; no retry or durability weakening |
| V-020 endpoint acquisition | R03-032; F-019 | Existing pgtest/testcontainers and affected module.extensions owner routes; exact selections frozen before execution | Deterministic causal regression where possible, successful selected endpoint SQL handshake, cancellation and failed-start settlement; common B/C source fidelity; required full-CI assurance remains open |
| V-021 cleanup-delegation assurance | R03-033; WF-07f | Exact `harness.browser.unit.testservices_lifecycle_contract` row, containing browser owner, catalog/generated/schema and final CI | Successful child returns 0; exactly one reaper receives the exact two-service lease; neither direct close callback runs; cleanup remains cleaning until its owner settles; no elapsed-time acceptance threshold. |
| V-022 renderer accounting boundary | R03-034; WF-07g | Exact `harness.browser.boundary_support.ui_review_contract`, containing UI-review contracts, explicit visual execution under required accounting, and current projection/owner checks | Accounted/unaccounted fake matrices are independent of ambient host membership; parent is passed at create; verification occurs after start and before browser use; verification failure cleans the exact allocation; production defaults retain native proof. |
| V-023 admission observations | F-020/F-024; R03-035/039 | Native admission/runtime and retained-reader contract routes, followed by focused pilots | Complete redacted decision model, bounded paging and scoped claim audit |
| V-024 retention correctness | F-022/F-023; R03-036/037 | Native weighted admission, fixture and runtime lifecycle rows plus installed boundary tests | Exact useful eviction, complete terminal consumers, settlement and exclusive progress |
| V-025 Go budget | F-021; R03-038 | Compiler/Go runner/catalog and affected owner routes | Stable profile budget, selectors, nested bounds and migrated identities |
| V-026 performance | R03-040; Section 7.8.4 | test-fast, browser functional and CI with frozen sources | Applicable benefit/guardrail decisions, independent evidence and rollback |
| V-027 final handoff | R03-041 | agent-finalize, source checks, owner/explicit selections, final CI/audits/lint | Retained source fully verified, settled, scanned and resumable |
| V-027a serial handoff repair | F-025; Core 03 REQ-03-314/315; R03-041 | Exact failed row, nine-row containing selection, finalization/types/frontend/generated/catalog and tracker lint | Both serial transitions assert presentation completion; supersession unchanged; audits/cleanup/scans pass; full CI explicitly incomplete |

V-003's independence matrix MUST bind every newly combinable row/helper to a
reviewed state-effect class. Runtime representatives may cover several rows only
when their complete traced isolation, setup, cleanup and shared-state mechanisms
are identical; the mapping and justification are part of G-02. An unclassified
effect is a gap, not permission to sample it away. A full Cartesian product of
all rows is not a default validation requirement.

| Matrix case | Required observation / failure oracle |
| --- | --- |
| A alone and B alone | Each representative establishes its own expected assertions and owned state without predecessor setup. The S-002 survivor has its separate exact-symbol route. |
| A then B and B then A in a permitted cohort | Both retain the alone-case assertions; contexts/cookies and observed actors are distinct, and no predecessor fault, clock, session or unauthorized data becomes observable. Legitimately shared control-plane effects are explicitly bounded in the dossier. |
| Compatible work in separate active lanes | Owned namespaces and generation/borrow bindings remain disjoint where required; no conflicting writes overlap and admission claims cover resident plus active work. One worker per cohort is unchanged. |
| Failed/interrupted predecessor | Product failure, teardown failure, cancellation, owner loss and reset failure each preserve settled outcomes, classify unexecuted work, quarantine uncertain ownership and allow only proved-clean later allocation. |
| Representative equivalence | Compare the same behavioral assertions and relevant state predicates. Nondeterministic IDs/timestamps may be normalized only by a rule fixed in the dossier; never ignore a field because it differs after the candidate run. |

For V-005, use the same declared instrumentation/cache policy for compared
treatments; report cold and warm conditions separately. A fresh execution
comparison uses the supported `CARTULARY_HARNESS_CACHE_MODE=off` when required
by the experiment. Compare exact roots with the existing equivalent-comparison
reader only when it admits the source/selection identities; rejected comparison
compatibility MUST remain a recorded limitation, not be bypassed. Performance
sampling MUST honor measurement quiet periods. No live polling or companion
diagnostic work may be introduced into those intervals by this roadmap.

S-002 uses the explicit one-symbol row split in Section 7.3, with its final ID
allocated at G-02. The two existing row IDs are known, but the post-cutover
literal selection is not yet allocated. V-001/RB-006 MUST freeze that selection
and its routing closure before execution. Preserve all siblings, and verify
that compatible aggregate packing does not introduce another process solely
because the survivor gained independent routing. No direct go test or undeclared
symbol-filter variable is permitted.

Final implementation acceptance MUST include each required affected layer,
not only performance samples. Existing incremental evidence rules remain in
force: unchanged applicable evidence may carry forward only through the adopted
source/impact accounting, and no blanket full-suite rerun is added by this
roadmap. Import-boundary/static checks are selected through
owner guidance if code movement creates that risk; no new production boundary
move is proposed here. Product browser/visual/accessibility checks are preserved
when their fixtures or execution contract changes. Core 05 publication and
release qualification are not implied by ordinary harness acceptance.

Retained-run maintenance is not executed here. In a later implementation task,
`make agent-finalize` follows repository procedure; supplying RESULTS_DIR
requires a compatible successful full warm check run root. Otherwise maintenance
MUST be reported skipped because RESULTS_DIR was unset.

### 8.2 Experimental acceptance contract

R03-009/010/017 and G-04 use this user-approved protocol. It supersedes the
creation-session blanket speedup requirement. E-048 records the later user
amendments to four pairs and conditional performance-only early exit. E-053 adds
the user-required below-5% early rejection. No correctness, cleanup, evidence-integrity or final-validation gate is waived.

| Item | Fixed rule |
| --- | --- |
| S-001 focus / containing workload | `make browser-e2e SUITE=support` / `make browser-e2e SUITE=functional`; exact catalog sets recorded from current projection before launch |
| S-002 focus / containing workload | Original incident support row from Section 7.3, then its reviewed split union / `make test-slice OWNER=module.incidents SCOPE=services`; exact pre/post assurance treatment map |
| Measurement capability | After S-008 acceptance, both treatments use `CGO_ENABLED=0`, `HARNESS_ACCOUNTING=required HARNESS_DIAGNOSTICS=basic` and execution-result reuse off. Prepare the native harness serially with the same cgo-disabled profile before launching either treatment, matching the ordinary Go runner's existing profile. CPU is cumulative charged CPU in the fresh invocation cgroup; memory is its kernel peak charge, including fresh hierarchy overhead. Shared host daemons, unrelated work and controller setup/publication are outside that scope; no whole-host claim |
| Later investigations | Freeze exact candidate rows and a containing workload after preceding dispositions; no sampling before registration |
| Attempts | Four matched baseline/candidate pairs per declared workload when full performance validation is required; odd pairs baseline-first, even pairs candidate-first; one excluded warm-up per treatment. Completed pairs count toward four; early-exit omissions remain recorded |
| Performance-only first-two-pair decision | Assess each declared workload separately using its first two independently validated matched pairs, excluding warm-ups. Improvement is `1 - candidate elapsed / matched baseline elapsed`. If both improvements are strictly below 5%, reject the candidate; exactly 5% does not trigger rejection. If either is at least 20%, the remaining performance protocol is optional. Otherwise complete four pairs and full validation. Never replace either first pair with a later favorable result; correctness, cleanup and final checks remain mandatory |
| Preparation | Warm dependencies/builds; execution-result reuse off; identical `HARNESS_DIAGNOSTICS=basic`, default four browser lanes, unchanged capacities/workers/deadlines; no competing intentional workload |
| Identity | Freeze each executable source snapshot before its first sample; retain toolchain/system/storage, fixture/template/profile and all permitted treatment differences |
| Primary metric | Owning-command monotonic elapsed time including required preparation, cleanup/publication and terminal exit |
| Guardrails | Elapsed time, CPU consumption, observed peak memory, per-run p95 ready-to-start queue delay, cleanup duration; unconditional cleanup success |
| Estimator | Candidate/baseline ratio of paired arithmetic means; bootstrap whole matched pairs 10,000 times, seed 20261021, Python `random.Random` MT19937 with four `randrange(4)` draws per resample for full validation, reset stream per workload; all metrics share each resampled index set. Arithmetic uses decimal input values, precision 28, HALF_EVEN rounding; decisions precede display conversion to floating point. Nearest-rank quantiles: sorted element `ceil(p * 10000) - 1`; one-sided 99th-percentile guardrail upper bound and one-sided 5th-percentile improvement lower bound |
| Full-protocol numerical acceptance | When neither first-two-pair decision applies, S-001/S-002: each guardrail upper bound <=5% regression. S-003/S-004: improvement lower bound >5%. S-005: improvement lower bound >10%. Performance-driven candidates also satisfy all 5% guardrails. An early-exit decision does not claim these omitted confidence bounds passed |
| Zero denominator | A zero baseline accepts only an observed zero candidate for that metric; no invented relative percentage |
| Missingness | Missing required observations or incompatible measurement scope => inconclusive. Partial CPU/memory coverage cannot establish whole-run resource guarantees |
| Failure | Retain every attempt/omission. Unexpected assertion, timeout, cancellation or cleanup failure blocks that source candidate; relevant repair and newly frozen comparison required |
| Comparison method | Audit exact runs independently; calculate over validated exports and the explicit reviewed treatment map. Existing same-source `COMPARISON=equivalent` admission remains unchanged |

A preregistered protocol is not evidence that collection is sufficient. First
perform an explicitly exploratory baseline pilot to verify feasibility and
metric coverage; pilots cannot become acceptance samples. If baseline collection
cannot support the required scope, do not implement a candidate assuming a later
performance pass. Record the missing signal and its owning capability.

Count physical generations, logical borrows, database clones/drops and test
fixtures separately. Do not sum overlapping CPU/memory scopes or phase durations.
A numerical analysis here is internal remediation evidence, not an executable
Markdown input, canonical comparison-reader waiver or Core 05 claim.

### 8.3 Finding-to-completion traceability

Completion receipts below distinguish accepted implementation, supported no-change
investigation and historical authoring evidence. E-072 supplies final current-source
integration assurance; performance claims retain their original source identities.

| Finding | Contract / proposed requirement | Slice / workflow | Validation | Completion evidence / disposition |
| --- | --- | --- | --- | --- |
| F-001 | C-001/C-002; R03-001–006/009–011/016/017 | S-001 / WF-01 | V-001/V-002/V-003/V-005/V-006 | E-031/E-039 adoption, coordinated v15 and effect dossier; E-033/E-034 selection/state/failure checks; E-049 assurance and performance accepted; E-071/E-072 final explicit and CI checks |
| F-002 | C-007; R03-007–010/017 | S-002 / WF-02 | V-001/V-004/V-005 | E-055 assertion map and independently attributable execution pass; E-054 rejects fixture removal and restores original lifetimes; broader replay consolidation dropped |
| F-003 | C-006; R03-009–011/017 | S-003 / WF-03 | V-005 | E-056/E-057: measured NO_CHANGE; admission assertions pass, no calibrated policy accepted; row-count estimate limitation retained |
| F-004 | C-002/C-003; R03-009–012/016/017 | S-004 / WF-04 | V-003/V-005/V-007 | E-058: complete declaration screen and baseline isolation/poisoning checks; no eligible conversion; stronger boundaries retained with exclusions/missing proofs |
| F-005 | C-003; R03-009–011/013–017 | S-005 / WF-05 | V-003/V-005/V-007 | E-059: dedicated-lifetime NO_CHANGE; >10% opportunity not established; no pooling contract, implementation or toggle adopted |
| F-006 | C-001/C-003/C-004/C-005; R03-016 | Preservation within activated slices | V-002/V-003/V-006 | E-070 lifecycle and command-surface checks; E-071 explicit browser coverage and E-072 CI; existing scheduler/broker/renderer/evidence responsibilities retained |
| F-007 | C-008; evidence-classification procedure | WF-00/WF-06 | V-008 | E-012/E-016/E-017 historical attribution preserved; E-018 authority update and E-072 current-source final verification remain separate |
| F-008 | C-008; validation-reporting procedure | WF-06 | V-006/V-008 | E-020 canonical lint coverage and pass; E-073 completion lint and final update command |
| F-009 | C-009; R03-018; REQ-856/AC-152 | S-008 / WF-08; S-008a / WF-08a | V-009/V-006 | E-030 initial capability and E-038 nested-environment correction accepted before S-001 resumed; E-071 accounting row and E-072 CI pass; declared host/build scope only |
| F-010 | R03-019/020; REQ-079/835/841–844/857 | S-009 / WF-09 | V-010/V-006 | ACCEPTED, E-075–E-081; excluded pilots and current repair assurance remain source-bound |
| F-011 | R03-021–024; REQ-801/803/806 | S-010/S-011/S-007e | V-011/V-012/V-006 | REJECTED / RESTORED, E-160/E-161; both performance decisions complete; final integration passes E-166; closing receipt E-167 |
| F-012 | R03-025; REQ-102/858 | WF-11c | V-013/V-006 | Historical assurance E-112; candidate-only behavior restored with the rejected lifecycle under E-161; no accepted optimization |
| F-013 | R03-026; Section 11.4, REQ-401/402, AC-007 | WF-11d | V-014/V-006 | ASSURANCE PASSED, E-118; common owner/code/schema correction validated in both treatments; no performance acceptance |
| F-014 | R03-027; Core 03 retained intent/continuity and REQ-03-304/305 | WF-11e | V-015/V-006 | ASSURANCE PASSED, E-123; setup barrier and isolated/containing checks pass; historical trigger not inferred as observed |
| F-015 | R03-028; NLSpec Section 11.4 startup context boundary | WF-11f | V-016/V-006 | ASSURANCE PASSED, E-127; no acceptance sample admitted |
| F-016 | R03-029; Core 03 mention completion and acknowledged read recovery | WF-11g | V-017/V-006 | ASSURANCE PASSED, E-130/E-131; no runtime defect or performance regression inferred from the failed test |
| F-017 | R03-030; Core 03 REQ-03-248a/220 | WF-11h | V-018/V-006 | ASSURANCE PASSED, E-135/E-136; identical common correction retained; no performance sample |
| F-018 | R03-031; Section 11.4 and AC-007 | WF-11i | V-019/V-006 | IMPLEMENTED, E-137–E-141; shared deadline and both-source focused assurance pass; E-159 warm-up and E-166 retained-source final CI pass |
| F-019 | R03-032; Section 11.4 | WF-11j | V-020/V-006 | E-150 NO_CHANGE and E-152 stability qualification; historical causality remains uncertain; E-166 final integration passes |

### 8.4 Historical authoring-session command admission and skips

| Operation | Invocation-path evidence / effects | Present disposition |
| --- | --- | --- |
| rg, sed, nl, Git with optional locks disabled | Read-only source/status inspection; no private runner imported | Performed in creation/review as attributed under E-001–E-017 |
| Python standard-library in-memory analysis | Reads explicit JSON/text, computes counts/hashes and models arithmetic; no repository module execution or output file | Performed; static evidence only |
| Public Make discovery and target-plan | No Make command needed for this bounded static task; future discovery is specified, not executed | Not run; RB-006 preserves required runtime selection discovery |
| make lint-markdown | Generated Make binding performs Node readiness, preflight and frontend-install prerequisites; runner writes cache stamp and retained artifacts; current config omits this file and public target declares no selection input | Not run; no lint coverage/pass claimed; no config/tooling repair authorized |
| Tests, browser sessions, builds, generation, formatting, service/recovery operations | Would execute product/harness work, mutate outputs or interfere with concurrent development | Not run, explicitly excluded |
| agent-finalize and retained-run maintenance | Implementation-only operation can rewrite generated artifacts | Not applicable to this documentation task; not run |
| Document structural inspection | Reading the admitted handoff and checking local references/IDs/tables/dependencies in memory | E-014/E-017; authoring inspection, not an automated lint/test suite |

For E-001–E-017, automated repository validation was **unperformed**. Static document inspection
MUST NOT be relabeled as a canonical lint pass. The later implementation request authorized the narrow lint-selection repair;
E-020 records its completion. No private linter or broad configuration edit was added.

### 8.5 Continuation experiment and acceptance contract

E-145 amends this section by explicit user instruction. Historical freezes and
receipts retain their original meaning. Both `make browser-e2e SUITE=functional`
and `make ci` SHALL pass independently. Baseline source SHALL contain accepted
S-009 diagnostics and separate resets; candidate SHALL add only S-010 lifecycle
changes. Common correctness repairs SHALL remain identical. Before sampling,
freeze exact rows, source snapshots, graph differences, permitted treatment map,
toolchain/storage identities, diagnostics, run IDs and run order. Executed
workloads SHALL use `CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Dependency/build caches
SHALL be warm. No competing intentional workload or polling during quiet
measurement is permitted.

| Workload | Reference and samples | Warm-up and comparison limits |
| --- | --- | --- |
| Browser functional | Measured matched baseline/candidate pairs; four pairs maximum, baseline first on odd pairs and candidate first on even pairs | Exclude one warm-up per treatment. Section 8.2 paired analysis and all five guardrails remain binding. |
| CI | User-assumed constant elapsed reference of exactly 3,600,000 ms; four independently validated candidate samples maximum | No further CI baseline timing runs or baseline warm-up. Exclude one candidate warm-up. Improvement is conditional on the assumed reference, not a measured before/after result. |

| Decision point | Binding rule |
| --- | --- |
| First two comparisons | Browser uses the first two matched pairs; CI uses the first two candidate samples against the fixed assumed reference. If both elapsed improvements are strictly below 5%, reject that workload. If either is at least 20%, stop remaining performance sampling for that workload and report early acceptance without omitted confidence bounds. |
| Otherwise | Complete four comparisons for that workload. Browser preserves matched order; CI runs candidate-only samples. |
| Browser full protocol | Benefit lower one-sided 95% bound SHALL exceed 5%; all five one-sided 99% guardrail upper bounds SHALL be at most 5% regression. Use Section 8.2's 10,000 paired bootstrap resamples, seed 20261021, decimal arithmetic and nearest-rank method. |
| CI full protocol | Resample the four whole candidate observations 10,000 times using the same seed, arithmetic and nearest-rank method; divide each resampled mean by the fixed assumed 3,600,000 ms reference. Conditional benefit lower one-sided 95% bound SHALL exceed 5%. This is a candidate-only bootstrap conditional on the assumption; it SHALL NOT be labeled paired or represented as incorporating baseline uncertainty. |
| CI resource reporting | Charged CPU, kernel peak memory, p95 ready-to-start delay and cleanup duration SHALL be reported descriptively, with scope and coverage. No comparative guardrail pass or regression percentage SHALL be claimed for these metrics. |
| Correctness and cleanup | Always required. Independently audit every attempt. Failed/omitted attempts remain recorded; repaired source requires a new freeze. Changed-source runs SHALL NOT use COMPARISON=equivalent. |
| Cross-workload decision | Both workloads receive the first two comparisons unless correctness/cleanup prevents execution. If either rejects, stop remaining optional sampling and restore the lifecycle owner, projections and implementation together. Retain accepted diagnostics and common correctness repairs. |

Primary elapsed time SHALL use owning-command monotonic accounting. Browser
guardrails SHALL cover elapsed, charged CPU, kernel peak memory, p95 ready-to-start
delay and cleanup duration, with explicit scope exclusions. The analysis SHALL
report critical-path queue and host-admission observations; acquisition,
preparation, runner and settlement durations; generations, retirement causes and
reset-only generations; reset/runner identity continuity; changed queue-percentile
populations and corresponding cohort waits. Measured deltas require measured
comparable references; historical CI diagnostics MAY supply labeled context but
SHALL NOT be promoted to a current baseline. Diagnostic pilots do not count.

S-007e SHALL run agent-finalize before broader checks, then required owners,
generated/schema/catalog checks, explicit support/visual/UI-review selections,
and one final CI on retained source. No golden refresh is authorized. Retained-run
maintenance SHALL use RESULTS_DIR only with a compatible successful full warm
check root; otherwise record that maintenance was skipped because RESULTS_DIR
was unset. Current scans, physical settlement and final tracker lint are required.

## 9. Top-Level Work Tracker

Statuses: TODO, IN_PROGRESS, BLOCKED, DONE, DEFERRED, DROPPED. A DONE
investigation may have a no-change result; implementation acceptance is separate.

| ID | Work item | Workflow | Status | Evidence / exit |
| --- | --- | --- | --- | --- |
| T-001 | Current baseline, authority and fixed protocol | WF-00 | DONE | E-018; exact selection/consumer inventory before closure |
| T-002 | Browser source-effect and migration dossier | WF-01 | DONE | E-032/E-039; R03-001–006 source/consumer map complete. Runtime/performance acceptance remains T-003 |
| T-003 | Browser v15 implementation and acceptance | WF-01 | DONE | E-049 accepts assurance and performance separately; candidate restored and generated check passes; boundary lint precedes WF-02 |
| T-004 | Incident consolidation and independent routing | WF-02 | DONE | E-054 rejects fixture removal under the first-two-pair rule. E-055 accepts exact Location assertions and independent routing; all original tests remain, no fixture-reduction claim |
| T-005 | Further replay consolidation | WF-02 | DROPPED | Distinct Unicode/durable/duplicate-key assurance retained |
| T-006 | Cost/admission investigation | WF-03 | DONE / NO_CHANGE | E-057: physical admission passes; row-count estimate limitation confirmed, calibrated granularity and >5% benefit unproved; no tuning accepted |
| T-007 | Isolation eligibility investigation | WF-04 | DONE / NO_CHANGE | E-058: complete declaration screen; 525 excluded, 370 unresolved, zero admitted conversions; stronger boundaries retained and baseline isolation/poisoning checks pass |
| T-008 | PostgreSQL lifetime investigation | WF-05 | DONE / NO_CHANGE | E-059: eligible startup opportunity does not establish >10% end-to-end benefit; dedicated lifetimes retained, no owner/schema/broker cutover |
| T-009 | Canonical tracker validation | WF-06 | DONE | S-006 coverage and Make lint receipt |
| T-010 | Final validation and handoff | WF-07 | DONE | E-070 repairs/finalization, E-071 explicit/projection checks, E-072 600/600 CI and cleanup/audit, E-073 completion lint and final handoff |
| T-011 | Complete invocation measurement capability | WF-08 | DONE | E-030; R03-018/V-009 accepted for the declared profile; no optimization performance claim |
| T-012 | Cleared-environment nested accounting correction | WF-08a / S-008a | DONE | E-038: three-graph independent publication, ordinary mode, four-graph nested cancellation, native checks and public audits passed; boundary lint required before resumption |
| T-013 | Continuation registration | WF-00e | DONE | E-074; registration lint passed |
| T-014 | Complete critical-path diagnostics and baseline pilots | WF-09 / S-009 | DONE | V-010; diagnostic implementation accepted E-081; no optimization performance claim |
| T-015 | Preparation under one admitted borrow | WF-10 / S-010 | REJECTED / RESTORED | E-086 assurance passed; E-160/E-161 reject performance and restore owner, implementation and projections together |
| T-016 | Independent browser and CI performance decisions | WF-11 / S-011 | DONE / REJECTED | E-160/E-161; both workloads reject under the first-two <5% rule; six optional attempts omitted; lifecycle restored; V-012 / Section 8.5 |
| T-019 | Socket fixture correction and PostgreSQL deadline diagnosis | WF-11b | DONE | E-100–E-103; socket correction passes both sources; PostgreSQL mechanics unchanged; full-CI gate remains open |
| T-020 | Restore inherited host admission in nested native execution | WF-11c | DONE | E-111/E-112; precedence tests and real 19-token/4-token parent cancellations pass; no performance acceptance |
| T-021 | Consolidate startup budget and failed-start removal ownership | WF-11d | DONE | E-113/E-115 adoption; E-116–E-118 common B4/C4 assurance and preservation; fresh experiment required |
| T-022 | Establish assessment fixture-publication precondition | WF-11e | DONE | E-121–E-123 common B5/C5 correction, isolated/containing assurance, preservation and closing lint; new experiment required |
| T-023 | Reconcile deliberate invalid-input assurance with static analysis | WF-11f | DONE | E-126/E-127; single-site explanation, unchanged runtime/assertions, both-source lint and focused checks, finalization and closing lint |
| T-024 | Make mention-recovery read boundary deterministic | WF-11g | DONE | E-129–E-131; identical common correction, both-source browser assurance, owner/type/projection checks and closing lint |
| T-025 | Establish linked metadata admission before virtual deadline | WF-11h | DONE | E-135/E-136; both-source browser/owner assurance, projections and immutable preservation; closing lint required before WF-11i |
| T-026 | Resolve forced-drop cleanup failures within bounded ownership | WF-11i | DONE | E-137–E-141 owner adoption, schema v3, both-source focused assurance and preservation; closing lint gates renewed WF-11 sampling |
| T-027 | Resolve PostgreSQL published-endpoint acquisition failure | WF-11j | DONE / NO_CHANGE transport | E-149–E-154 complete Section 7.7: effective-request diagnostics, unchanged publication, 192 qualified lifetimes, both-source assurance and freeze; historical cause remains uncertain; final integration remains required |
| T-028 | Characterize exact publication and client transport | WF-11j-a | DONE | E-149; v6 observation and negative contracts pass; 64-lifetime pilot passes; historical failure not reproduced |
| T-029 | Qualify or reject daemon-publication candidate | WF-11j-b | DONE / NO_CHANGE | E-150; failing control was not reproduced; existing publication retained; no speculative candidate or compatibility mode |
| T-030 | Validate repeated fixture lifetime stability | WF-11j-c | DONE | E-152; 192/192 independent positive lifetimes; affected and containing selections pass in both treatments |
| T-031 | Reconcile transport disposition and freeze experiment | WF-11j-d | DONE | E-154; both sources finalized, owner/projection checks pass, common B/C fidelity and Section 8.5 experiment 09 frozen; boundary lint required before sampling |
| T-032 | Remove deprecated no-op SDK option from diagnostic test | WF-11k | DONE | E-157; identical test-only repair, both-source focused regression/Go lint/finalization/generated checks pass; experiment 10 frozen; boundary lint precedes warm-ups |
| T-018 | Corrective nested installed-service startup assurance | WF-11a | DONE | E-094–E-096; repaired B/C focused assurance passed; experiment 02 requires new warm-ups |
| T-033 | Make cleanup-delegation assurance causal | WF-07f | DONE | E-163; direct delegation, exact lease and no direct-close assertions; focused/containing/projection/lint checks pass; no runtime change |
| T-034 | Complete renderer accounting fixture boundary | WF-07g | DONE | E-165; R03-034/V-022; focused/containing/real visual and owner/projection checks pass; required accounting remains fail closed |
| T-017 | Continuation validation and handoff | WF-07e / S-007e | DONE | E-165/E-166 current owner/projection/explicit checks, 600/600 final CI, audits and settlement pass; E-167 final tracker closure; R03-024 |
| T-035 | Register reservation continuation | WF-00f | DONE | E-168; source/traceability and registration lint pass; closing lint gates WF-12 |
| T-036 | Explain native admission pressure | WF-12 / S-012 | DONE | R03-035/039; V-023 |
| T-037 | Correct retention and exact eviction | WF-13 / S-013 | DONE | R03-036/037; V-024; E-174 |
| T-038 | Implement profile-owned Go budget candidate | WF-14 / S-014 | REJECTED / RESTORED | R03-038; V-025; E-177 assurance; E-194 performance rejection and coordinated restoration |
| T-039 | Decide optional tuning performance | WF-15 / S-015 | DONE / REJECTED | R03-040; V-026; E-194; fourteen dependent attempts omitted; closing lint gates WF-16 |
| T-040 | Final validation and handoff | WF-16 / S-016 | IN_PROGRESS | R03-041; V-027; E-198 user-cancelled CI; focused repair WF-16a required; complete CI assurance remains unestablished |
| T-041 | Repair serial browser handoff assurance | WF-16a / S-016a | DONE | Core 03 REQ-03-314/315; R03-041; V-027a; E-199/E-200; completion lint passed and final closing lint registered |

## 10. Session Handoff Log

### 10.0 Implementation execution receipts

**E-018 — WF-00 admission (2026-10-21).** User explicitly requested execution of
the approved plan. Root/tooling/backend guidance and test-operations workflow
revalidated. HEAD and logical index match Section 1.3. Only the pre-existing
tracker edit was dirty. No product or machine-owner changes at admission.
Public command discovery succeeded. Current schema is browser batch v14; current
REQ-806 still partitions by owner and REQ-804 still prohibits unrelated exclusive
borrowers. No old receipt is promoted to current-source assurance. Historical
E-001–E-017 and the bounded 124-file audit remain below unchanged in authority.

**E-019 — WF-00 completed; WF-06 entered.** Public `make help ALL=1 JSON=1`
and the five `make task-guide OWNER=...` commands succeeded (native 85006/0 for
the owner chain and row-ID allocation). `make author-test-row-id` allocated
`module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5`
from family `module.incidents.support_integration`, claim `Authenticated incident
creation and normalized replay preserve exact Location headers`, selector key
`TestIncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance`.
This is allocation only; the row does not exist until S-002 edits routing.

The baseline browser manifest SHA-256 is
`1b4e9011a8b0c3a593c3a6f5ad428bfdc5083fa5970f99e3e5885a0f96d2e701`.
Support selects 18 rows (17 test-owned); functional stage inventory contains 382 (325 test-owned), before the public suite eligibility filter. E-033 corrects the public functional selection and includes the four new common assurance cases.
The SHA-256 of sorted row IDs joined by a newline without a terminal newline is
`dd5a91a31b4738e9dd22623e6bde57af710facaa8471a836c3f5c7320212924b`
for support and
`7fa83dd25a4192a6073cb1e27db0722d8712058f8ed03bc8dbf9a74d61beda3f`
for functional. These authoring selection records qualify the frozen workloads;
canonical execution binds its own machine-source identities.

S-001 consumer roster: browser batch schema/attachment; execution-topology
generator; browser manifest reader, cohort builder and catalog-group CLI;
scheduler browser adapter, browser work-graph compiler and functional-lane
planner; generated-shape validator and fixtures; browser lane-policy tests;
work-graph and execution evidence consumers. Generated schema validators,
foundation validators, manifest, task-surface and topology projections refresh
through `make generate`, never hand edits. Audit discovered no alternate
partition owner transport; row ownership remains mandatory. Specific dynamic
selection/failure behavior still requires S-001 characterization.

Focused validation routes include `harness.browser.unit.browser_lane_policy`,
`harness.browser.unit.native_fixtures`,
`harness.browser.integration.native_fixture_lifecycle`,
`harness.test_catalog.behavior.routing_contract`,
`harness.evidence_accounting.behavior.current_epoch_evidence`, and
`harness.evidence_accounting.behavior.observation_admission`, plus the exact
incident rows in Section 7.3. `source-check CHECK=generated|schema|catalog`
projects the owner changes. No product baseline ran during WF-00.
Metric feasibility is an exploratory pilot entry in WF-01; it cannot be presumed
from optional collector validity. WF-00 source/protocol work is complete.

**E-020 — WF-06/S-006 completed before WF-01.** Added exactly this tracker to
`.markdownlint-cli2.jsonc`. `CARTULARY_TEST_RUN_ID=r03-s006-lint-01 make lint-markdown`
passed; native session 73358, outer exit 0; run root
`.cartulary/test-results/r03-s006-lint-01`, summary
`adhoc/lint-markdown/tool-run-summary.json`. Target reports `adhoc`, status pass,
14,800 ms. No product, generated artifact or dependency change was needed.

**WF-01 exploratory pilot registration.** Baseline-only `browser-e2e SUITE=support`,
result parent `.cartulary/test-results`, fresh run ID `r03-s001-baseline-pilot-01`,
result reuse off, diagnostics basic, default lane/capacity policy. This pilot
checks current assurance and whether existing exported measurements can support
Section 8.2; it is excluded from all acceptance samples. No browser source or
owner amendment has been made. Source inspection found REQ-847 explicitly excludes
initial Make startup and terminal-receipt publication; the receipt cannot prove
total invocation cost. Current resource explanations declare observed-partial
coverage. These are potential capability blockers, not acceptable zero values.

**E-021 — WF-01 pilot and measurement-capability finding (in progress).** Native
session 81537 owns `r03-s001-baseline-pilot-01`. Its manifest binds source
`sha256:1ae10feef63bcec2c40f9e23b8836628ae39caf8115fa76bab4bbd7dbd19a348`
and graph
`sha256:afe3bbcd826886670fb3f0b33f6d1bfaad32c2e0cde62fea2480abfc45f02dfe`.
Public observation revision 35 reported 15/20 units passed, one running and four
pending, no observed failures; it is not a terminal receipt. Exact run root is
`.cartulary/test-results/r03-s001-baseline-pilot-01`.

Source evidence for RB-009: `resource-accumulator.mjs` counts CPU only over
jointly sampled intervals after attribution; `resource-presentation.mjs` names
`unit_cpu_lower_bounds`, bounds output to 20 units/scopes, and labels RSS as
`sweep_subset_shared_pages_may_repeat`. REQ-847 explicitly excludes initial
Make startup and receipt publication and prohibits treating its receipt as
total invocation cost. The existing same-source comparison admission is also
unchanged. None of these intentional bounds is a product failure or permission
to invent total cost. A user clarification has been requested; no candidate
cutover, spec amendment, acceptance experiment or later workstream has started.

**E-022 — WF-01 exploratory baseline terminal receipt.** Native 81537 ended with
outer exit 0. `r03-s001-baseline-pilot-01` reports pass, 20/20 work units, canonical
graph duration 152,749 ms. The independently validated terminal receipt reports
child exit 0, outcome passed, retained-secret scan passed, wrapper-observed
158,620 ms with the REQ-847 exclusions. `make harness-observability-check
RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/r03-s001-baseline-pilot-01
CHECK=all` passed (outer exit 0), `full_validity=1`, execution/resources collection
complete, zero collection omissions, envelope and terminal receipt available.
This establishes current baseline assurance/cleanup/privacy validity only.

`make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/r03-s001-baseline-pilot-01
DETAIL=performance JSON=1` also exited 0. Its resource coverage is nevertheless
`observed_partial`: 684 samples, zero collection omissions, 83 scopes omitted from
the bounded explanation. Complete collection is not complete physical coverage;
CPU lower bounds remain lower bounds, and observed RSS is not whole-invocation
peak memory. The successful pilot is excluded from acceptance samples.

RB-009 is confirmed by current execution and source contract. The requested
choice is an additional measurement-capability workstream versus explicitly
scoped diagnostic acceptance, with stopping at this blocker also offered. No
choice has been received at this checkpoint. WF-01 remains BLOCKED, not DONE or
an accepted no-change investigation. S-002 through S-007 have not begun, preserving
the approved serial workflow. No browser schema, product test, fixture declaration,
cost policy or service lifetime was changed. No 12-pair experiment was launched
with insufficient required measurements. Finalization/generation/final CI are
not run because no implementation cutover has begun and WF-07 has not been reached.
Retained-run maintenance is skipped: RESULTS_DIR is unset for agent-finalize and
there is no selected successful full warm check root.

Source-effect discovery remains bounded: the 465 test-owned rows span 72 browser
files; shared actor/context/session fixtures and the support incident scenarios
were directly read, but this does not establish the full R03-002 effect dossier.
A resumed implementer must complete it before cross-owner candidate execution.

**E-023 — Checkpoint validation and handoff.** `make explain-run` summary confirms
20 passed, zero failed/skipped/cancelled units and 39 cleanup steps with zero
failed or blocked. A first summary-reader attempt incorrectly combined
`DETAIL=summary JSON=1` and exited 2 with a usage error; the supported human
summary command then exited 0. This reader usage error is not a pilot failure.

`CARTULARY_TEST_RUN_ID=r03-checkpoint-lint-01 make lint-markdown` passed on the
checkpoint text, native 26117/0, summary
`.cartulary/test-results/r03-checkpoint-lint-01/adhoc/lint-markdown/tool-run-summary.json`.
A final lint confirmation after this receipt uses `r03-checkpoint-lint-02`.
Git status contains exactly `.markdownlint-cli2.jsonc` and this tracker; no source,
test, schema or generated output is modified. No execution session remains live.
Next action is resolution of RB-009, then resumption of WF-01; do not jump to
later workstreams or report S-007 complete. The pending clarification is a scope
choice caused by the plan’s measurement assumptions, not a new authorization
request for the already approved implementation.

**E-024 — Measurement scope decision and feasibility investigation.** The user
selected "Add a measurement-capability workstream". WF-08/S-008 is inserted after
completed WF-06 and before resuming WF-01. This resolves the scope decision, not
RB-009's missing capability. All original numerical guardrails remain binding.

Read-only environment inspection reports cgroup v2, delegated CPU/memory
controllers under the user manager, and Docker Desktop using cgroupfs v2 on the
WSL kernel. These facts alone do not prove shared accounting. A transient user
service feasibility probe created a fresh workload cgroup, ran a short-lived
child and read final usage after exit: child exit 0, CPU 17,674 microseconds,
memory peak 17,039,360 bytes, populated 0; probe exit 0. Its owned workload group
was removed and the transient service collected. This is an environment probe,
not repository validation or an acceptance sample. Docker membership remains
to be proved before implementation design is finalized.

The user's eBPF suggestion was evaluated: BTF is present, but effective process
capabilities are zero, unprivileged BPF is disabled (value 2), perf paranoia is 2,
and bpftool/bpftrace are absent. No privilege or host policy was changed. Kernel
cgroup accounting is the preferred first path; privileged tracing remains a
possible targeted alternative, with no current completeness claim.

**E-025 — S-008 implementation admission and first validation.** A second
environment probe placed one exact, newly created Docker container in the fresh
local accounting hierarchy. Create/start/wait/remove all exited 0; the container
group was visible locally, populated changed from 1 to 0, final CPU was 40,698
microseconds and memory peak 8,863,744 bytes. The empty parent was removed. A user
scope probe also proved inherited execution in a sole-member delegated scope and
CPU/memory controller enablement. These are feasibility observations, not product
or performance acceptance.

Adopted REQ-856/AC-152 before implementation. Authored changes now introduce the
required-accounting Make input, native cgroup controller and fenced container
placement, the closed invocation-accounting v1 artifact, explanation v6 and
whole-selection queue/cleanup aggregates. Existing terminal receipts, ordinary
optional diagnostics and same-source comparison admission retain their meanings.
The new native row was allocated through `make author-test-row-id`:
`harness.evidence_accounting.behavior.invocation_accounting_preserves_complete_kernel_74dff0b392`.

`CARTULARY_TEST_RUN_ID=r03-s008-format-01 make format` passed, native 61117/0,
graph duration 8,066 ms. `CARTULARY_TEST_RUN_ID=r03-s008-generate-01 make generate`
passed, native 7696/0, 10,440 ms. Schema registry changes regenerate the shared
validator digest and dependent validators; generated files were not hand-edited.

First focused attempt: `CARTULARY_TEST_RUN_ID=r03-s008-native-01
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.invocation_accounting_preserves_complete_kernel_74dff0b392,harness.evidence_accounting.behavior.native_observation`.
Native 75746 exited 2; graph source
`sha256:dc63e093b9113885b75b604f7148deebe13774a34298888e18d6582bbf014f19`,
two work units, one passed/one failed, duration 1,700 ms. Existing native
observation passed; `TestAccountingKernelLifecycle` rejected initial setup.
All 11 cleanup steps completed, with zero failed/blocked. Exact root:
`.cartulary/test-results/r03-s008-native-01`. Public summary/log readers exited 0.

Diagnosis found that creating an empty descendant cgroup itself can charge
kernel memory: an environment probe observed CPU 0, populated 0, OOM kills 0,
but peak memory 262,144 bytes. The implementation incorrectly demanded zero
initial memory. The repair retains this overhead in the peak, records the
initial peak separately and rejects decreasing peaks; it never subtracts peaks.
This failed source remains failed. Repaired-source validation is pending; no
S-008 acceptance or S-001 entry is claimed.

**E-026 — Repaired-source checks and first accounted command.** Format-02 passed,
native 9188/0, graph duration 8,023 ms; generate-02 passed, native 95065/0,
10,130 ms. The second focused selection added
`harness.evidence_accounting.behavior.current_epoch_evidence` to E-025's rows.
Run `r03-s008-native-02`, native 30879/2, source
`sha256:9f3eb11793c8e3d326f185b6d207636aa0f3a19fa25d913c8a9371f06a4465b0`,
passed three of four units. Native observation and the current-epoch reader
contracts passed. The new lifecycle fixture failed because its test directory
did not have the required 0700 private mode; the production constructor already
uses a private `MkdirTemp` directory. All 13 cleanup steps completed. The fixture
was corrected without weakening secure publication.

Run `r03-s008-native-03` selected only the new native accounting row with cache
off. Native 86531/0 passed, graph duration 704 ms. This establishes the tested
short-lived-child counters, live ownership fence and refusal to delete an
unretired service group; it is not the complete S-008 failure matrix.

First required-accounting command: `CARTULARY_TEST_RUN_ID=r03-s008-accounted-unit-01
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.native_observation
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Native 56426/0 passed;
graph duration 663 ms. The independently validated accounting artifact binds
source `sha256:ecfba78e5f04177acbd7ca768c5e60136bbaf12b019d1f302670423336976a82`
and graph `sha256:7d136d549dad2676d4320666cf39be2c8d333db910431538675371d7a6256745`.
Measured child duration 4,799.645002 ms; charged CPU 6,131,880 microseconds;
peak memory 302,280,704 bytes; initial hierarchy peak 262,144 bytes; terminal
membership empty; child exit/signal 0/0. `make harness-observability-check
RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/r03-s008-accounted-unit-01
CHECK=all` exited 0 with full validity. Performance JSON exited 0 and reported
the v6 accounting record. It exposed a projection omission: cleanup end was
reduced but not carried into the retained model. That source was repaired and
the projection receives a regression assertion; no missing value became zero.

An exploratory full support command is now registered as
`r03-s008-accounted-support-01`, native 70352, using cache off and
`make browser-e2e SUITE=support HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Its source is
`sha256:9d92c55c1edf684aaeb334de3cd786dba8d05a092080346c8c3923c48be0459f`;
graph `sha256:afe3bbcd826886670fb3f0b33f6d1bfaad32c2e0cde62fea2480abfc45f02dfe`.
This is a capability pilot, not an acceptance sample. Native 70352 exited 0;
all 20 units passed and all 39 cleanup steps completed. Graph duration was
157,282 ms. The accounting record reports child duration 161,845.737279 ms,
charged CPU 208,111,877 microseconds, peak charge 5,415,374,848 bytes and initial
hierarchy peak 262,144 bytes; terminal membership was empty, exit/signal 0/0.
All-selected aggregates report p95 queue delay 1 ms and cleanup union 5,120 ms.
`make harness-observability-check RESULTS_DIR=.cartulary/test-results/r03-s008-accounted-support-01 CHECK=all`
and performance explanation both exited 0; full validity was established. This
source precedes the nested-context repair below and is not a current-source pass.

**E-027 — Nested accounting context and focused validation.** Format-03
(native 60338/0, 8,010 ms) and generate-03 (native 93064/0, 10,160 ms)
completed. Review found that deliberate environment clearing in nested tests
could lose an ambient accounting variable. Context now comes from exact kernel
cgroup membership plus a private live-owner/inode binding; nested graph runtime
registration resolves the same outer invocation even after environment clearing.
No directory discovery or ownership-by-process-name fallback was added.

The delegated kernel probe is explicit-only, because ordinary CI must not require
user-systemd delegation. Counter-integrity and signal-settlement tests remain in
the fast row. Both rows are explicitly selected for capability acceptance.
`r03-s008-format-04` failed native-runtime compilation before formatting because
the new local name collided with the existing parent-admission variable; its
owning exit was lost across context recovery and remains unavailable. Format-05
(native 26755/2) exposed one remaining reference to that variable. Both failures
are retained; the corrected source passed format-06 (native 83208/0, 7,871 ms).
The mistaken status observation with `RUN_DIR` was rejected as a usage error;
the corrected public observer used `RESULTS_DIR` and did not rerun work.

`r03-s008-native-04` (native 94867/0) selected both new accounting rows plus
`native_observation` and `current_epoch_evidence`, with execution-result reuse
off. Source `sha256:ca76c93be49baa4f73c81324206cd6820dde18bf0e18185b7b61570aab823a1f`
passed in 11,709 ms with all 13 cleanup steps completed. This includes the real
delegated short-lived process, malformed counters, exact ownership fences,
signal propagation and retained-reader/aggregate contracts. Schema check
`r03-s008-schema-01` (native 66674/0) passed in 6,042 ms. The real nested service
cancellation pilot is registered as `r03-s008-accounted-nested-01`, native 4254;
its terminal outcome and the rest of V-009 remain open. S-008 is not complete.

**E-028 — Nested services, cancellation and closure hardening.** The E-027
nested pilot exited 0 (native 4254), with one passing outer unit, 48,309 ms
graph duration, all ten outer cleanup steps completed and a passing retained
scan. All three nested cancellations (summary, machine and startup) settled
their services and scans. One accounting record binds all four graphs: child
duration 52,921.889539 ms, CPU 51,697,582 microseconds, peak charge 1,817,718,784
bytes, empty membership and exit/signal 0/0. Public `CHECK=all` and performance
explanation exited 0. Generated-source check `r03-s008-generated-01`
(native 33809/0, 16,906 ms) and the exact command-surface rows `owner_boundaries`
and `public_registry_parity` in `r03-s008-command-surface-01`
(native 87086/0, 64,561 ms) passed.

The intentional external cancellation pilot used the same source and support
command as E-027, fresh ID `r03-s008-accounted-cancel-01`, native PTY 72257.
After public status showed fixture preparation, Ctrl-C was delivered through
that exact native handle. The outer session exited 1; the measured child Make
and receipt retain exit 130 and signal 2. The run is `cancelled`, with ten
passing units, one `harness/fixture_error` during interrupted acquisition and
nine cancelled units. All 21 cleanup steps completed; retained scan passed.
Accounting reports 80,750.705033 ms, CPU 100,729,074 microseconds, peak charge
3,931,529,216 bytes and empty terminal membership. Performance explanation
passed. `CHECK=all` exited 2 because its required-evidence closure deliberately
requires a passing run; this is not a cancellation pass or a product failure.
The appropriate `CHECK=diagnostics` exited 0 with accounting available and
partial execution observations, preserving unexecuted work. No reader gate was
weakened.

Further review made cleanup settlement an explicit producer/reader condition
and made failed required graph registration poison outer closure, even if a
nested caller catches its error. Native publication tests cover primary exit 7,
all-member validation before writing, publication failure, closed-record
overwrite rejection and failed-registration poisoning. Kernel tests reject stale
controller identity and replaced hierarchy identity; counter fixtures reject
OOM, populated groups and unreadable CPU before any hierarchy removal. These
do not claim actual host OOM injection. `r03-s008-native-05` (native 70821/0),
source `sha256:8a2228134f2a8a2c7743b95f0c22236663ab63d6a1c562416d5af768268fd8f0`,
passed all four focused units in 12,166 ms; all 13 cleanup steps and the scan
passed. A subsequent explicit probe adds real orphan-descendant settlement
through its fixture-owned pipe. Final current-source capability checks remain
pending; no optimization acceptance or S-008 completion is claimed here.

**E-029 — Completed pilot evidence and executable identity admission.**
Format-07/08/09 passed (native 94071/0, 7,956 ms; 13046/0, 8,201 ms;
60489/0, 8,302 ms). Generate-04 passed (native 9257/0, 10,570 ms).
On source `sha256:6fcd61b1ce0d0b1a35aef299bf364f34235b4bfaa335caa017e27cd5dd547b63`,
native-06 passed all four focused units (native 4698/0, 61,146 ms, thirteen
completed cleanup steps). This includes actual orphaned-descendant detection
after its launcher exited; accounting refused retirement, and the fixture's
exact pipe release settled the descendant without accounting taking over cleanup.
The support-02 pilot (native 20461/0) passed all twenty units in 198,774 ms and
all 39 cleanup steps. Its measured duration was 205,622.190113 ms, CPU
213,032,838 microseconds and peak charge 4,084,187,136 bytes. Nested-02 (native
54273/0) passed in 48,605 ms with ten completed outer cleanup steps; measured
duration 55,801.430082 ms, CPU 50,388,433 microseconds, peak charge 806,588,416
bytes. Both records report empty membership and exit/signal 0/0, and both public
`CHECK=all` validations passed. These overlapping capability pilots are not
performance samples. Markdown lint-01 passed (native 62452/0, 19,370 ms).

Final source-identity review added `harnessaccount` to the native build inputs
and requires identical controller/runtime executable content before graph work
and again at closure. REQ-856 was amended first. Format-10 (native 43571/0,
7,873 ms) and generate-05 (native 5040/0, 10,370 ms) passed. Native-07 passed
(native 60203/0, 12,185 ms). The concurrently launched support-03 and nested-03
pilots (native 78753/2 and 66760/2) encountered the intended stale-controller
rejection: their native build prerequisite replaced the executable after the
outer controller started. Public status reports only successful build tool
summaries; there is no product graph or complete accounting claim. These were
setup mistakes, not planned negative attempts, and remain retained failures.
Preparation must complete serially before any accounted invocation starts.

The deliberately invalid requests `r03-s008-reject-mode-01` (basic diagnostics
omitted) and `r03-s008-reject-cache-01` (cache mode `on`) both exited 2 at workspace
admission, before graph or fixture creation. The selected row was
`harness.evidence_accounting.behavior.native_observation`; both requested
`HARNESS_ACCOUNTING=required`, and the second also requested basic diagnostics.
No prerequisites were installed or privilege settings changed. Final serialized
preparation and capability validation remain pending at this receipt.

**E-030 — S-008 completion and WF-01 entry boundary.** The accepted executable
source is `sha256:65fa4b6f9231cb0638e0a60a26eb53b3f62f1b81a2e0c782de99e453c5a1b833`.
Serial preparation `r03-s008-prepare-01` (native 26910/0, 16,472 ms) passed the
generated check. Accounted-unit-02 (native 47164/0, graph 666 ms) passed and its
public `CHECK=all` passed: measured 4,920.176318 ms, CPU 6,446,074 microseconds,
peak charge 300,048,384 bytes, empty membership and exit/signal 0/0. Schema-02
(native 14941/0, 6,204 ms) passed.

Accounted-support-04 (native 21197/0) passed all twenty units in 159,319 ms;
all 39 cleanup steps and the retained scan passed, as did public `CHECK=all`.
Measured child duration was 164,052.967488 ms, CPU 206,002,704 microseconds,
peak charge 2,609,205,248 bytes and empty membership. The concurrent nested-04
attempt (native 64021/2) failed its assertion when its nested graph rejected a
different executable: ordinary Go-runner children explicitly disable cgo while
the outer native build used cgo. All ten outer cleanup steps completed. This
failed attempt and the missing complete-accounting record remain preserved;
the guard was not weakened. Two diagnostic `DETAIL=logs` requests were rejected
for missing/undeclared target selection; bounded published unit logs established
the failure. No private event staging was read.

The frozen measurement build profile now includes `CGO_ENABLED=0`, consistent
with the existing ordinary Go runner. Serial preparation with that environment,
`r03-s008-prepare-02` (native 35618/0, 16,043 ms), passed the generated check.
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s008-accounted-nested-05
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.native_nested_cancellation
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` passed (native 78516/0,
graph 40,974 ms). All three nested cancellation variants settled; all ten outer
cleanup steps and the retained scan passed. Public `CHECK=all` and performance
JSON both exited 0. One accounting record binds all four graphs: child duration
45,573.308911 ms, CPU 40,356,343 microseconds, peak charge 738,926,592 bytes,
initial hierarchy peak 262,144 bytes, empty membership and exit/signal 0/0.
The cryptographic build receipt's native binary SHA-256 remained
`b05736afa5e7ac182dd200ddedb6fe7b6e9374aa94c9a054fef9fd6e9d4cea16`
before and after nested execution. This is profile consistency evidence, not a
performance comparison or a claim that the earlier failed attempt passed.

Disposition: investigation complete; implementation and capability assurance
accepted; optimization performance not established and not required for S-008.
R03-018 maps to adopted REQ-856/AC-152, native `harnessaccount`/`harnessdiag`,
workspace admission, fixture/renderer placement and validated retained readers.
E-025–E-030 cover native short-lived and orphan descendants, container lifetimes,
nested commands, cancellation, unavailable/corrupt counters, stale owner and
executable identities, OOM-counter rejection, incomplete registration and
publication failure. OOM rejection uses controlled counter fixtures, not an
actual host OOM. Unaffected earlier failure-path receipts remain source-qualified;
current-source full support and nested-service pilots close integration.

The capability is bounded to a delegated Linux cgroup v2 hierarchy with locally
provable owned-container membership and consistent native build settings. It
does not include shared host-daemon costs, previously charged shared cache, or
controller work outside the measured child command. Required uncertainty fails;
default-off execution remains available. No eBPF privileges or installation,
new service pool, scheduler policy, product behavior or product-domain vocabulary
was introduced. Explanation v6 and accounting v1 migrate together; historical
evidence retains its source and meaning. Rollback would restore the owner,
authored projections and implementation together after invocation settlement.

No candidate-owned resource remains unresolved in these completed pilots.
S-007 finalization, broader checks and final CI remain pending. Retained-run
maintenance is skipped because `RESULTS_DIR` is unset: no compatible successful
full warm-check root has been established. WF-01 may start after this tracker
boundary passes `make lint-markdown`; its next work is the complete source-effect
dossier and fixed baseline selections, followed by REQ-806 adoption and v15
cutover. No browser batching acceptance is implied by this capability completion.
Boundary lint `r03-s008-boundary-lint-01` passed (native 96708/0, 9,350 ms).
This receipt was recorded before WF-01 investigation began.

**E-031 — S-001 candidate and source-effect investigation (in progress).**
REQ-806/AC-088 now adopt reporting-independent browser execution. The v15
candidate removes partition owner fields, shares one explicit compatibility
projection between generation and admission, and derives cohort identity from
that tuple and the exact selected title/scenario pairs. Runtime/fixture/stage,
isolation, preparation, affinity and invocation-policy distinctions remain.
The v14 schema is rejected; no dual reader, owner exception or toggle is added.
Catalog ownership and execution owner `harness.browser` remain separate.

Authored baseline inputs are retained under
`.cartulary/remediation03/s001-baseline/paths.json`; these are internal treatment
preparation records, not an alternative execution interface. Generated outputs
are produced through Make. The static row/helper inventory at
`.cartulary/remediation03/s001-state-inspection.json` covers 465 test-owned rows
and explicitly does not claim complete effect proof. Thirteen constant-title
cases and consequential helper review remain outstanding. No product candidate
execution or performance freeze has occurred.

All execution commands below use `CGO_ENABLED=0`, fresh explicit
`CARTULARY_TEST_RUN_ID` and `CARTULARY_HARNESS_CACHE_MODE=off`.
`make generate` attempt `r03-s001-generate-01` failed (native 53407/2,
11,354 ms) because the reader passed a profile object into string validation.
The projection was repaired; `r03-s001-generate-02` passed (77905/0,
11,430 ms). `make format`, `r03-s001-format-01`, passed (80191/0,
8,027 ms). The exact `test-slice OWNER=harness.browser
ROWS=harness.browser.unit.browser_lane_policy HARNESS_DIAGNOSTICS=basic`
selection passed as `r03-s001-lanes-01` (61755/0, 7,512 ms).
After adding current-manifest selection and rejection coverage,
`r03-s001-lanes-02` reports pass at source
`sha256:6dbcc6f6d93cd90036c243ed9cd76283180fa211fc07ba58b782910d5937d54a`;
its native session receipt was lost when tool output was truncated, so its outer
exit is unavailable and no complete command-success claim is made.

`r03-s001-graph-01`, `test-slice OWNER=harness.test_catalog
ROWS=harness.test_catalog.behavior.routing_contract HARNESS_DIAGNOSTICS=basic`,
reports product/test_assertion_failure at that same source (50,690 ms).
Thirty-eight of 39 contracts passed, including browser-work-graph. The failure
was a remaining visual contract assertion that required one reporting owner
per cohort. That assertion is being migrated to the adopted execution-owner
contract; exact visual membership and singleton-state assertions remain.
All eleven cleanup steps passed. Native outer exit is likewise unavailable
because that tool response was truncated. These failed/unobserved attempts are
retained and cannot become accepted performance samples.

Read-only diagnostic attempts with unsupported `FORMAT=json` and an internal
rather than public target were rejected (exit 2); the valid public
`explain-run RESULTS_DIR=.cartulary/test-results/r03-s001-graph-01
TARGET=test-slice DETAIL=logs` exited 0 and identified the assertion. That
reader exit is not a test exit. WF-01 remains active; G-03/G-04, subsequent
workstreams and S-007 remain open.

**E-032 — S-001 source-effect dossier and assurance preparation.** The
[per-row source evidence](testing-harness-remediation-03-browser-state-evidence.json)
records all 465 original test-owned rows, their exact case locations and named
helper references, plus four new independently selectable fixture-boundary
cases. This is documentation-only review evidence; no executable input reads it.
All thirteen constant-title cases now resolve. The transitive import roster is
an over-approximation, including helpers used by isolated cases in the same
file; imports alone never authorize a weaker scope.

The consequential shared boundaries were inspected: fresh test actor/context
and tracked-session revocation; membership-filtered incident SQL; exact
incident/record/view/member/object/job parameters; actor-bound account and
workbook preferences; page-owned routes, clocks and DOM; finally-settled
collaboration contexts and interception; per-page WeakMaps; the visual
investigation Map's verify/finally-clear hook; group-scoped evidence paths; and
immutable runtime/trust/profile inputs. No unproved global mutator was found in
the newly combinable selections. Existing global clock, public-error fault,
deployment-wide administration/listing, lifecycle and measurement rows retain
their stronger declarations. Source review permits characterization; it does
not establish runtime assurance or performance acceptance.

The independence matrix uses the four literal cases in `fixture-state.spec.ts`:
A alone, B alone, A→B, B→A, all four without an intervening cohort reset, and
separate simultaneous invocations with the same exact selected test IDs.
Each case requires an empty membership-visible directory and default own-user
preferences, creates its own incident/record/private saved view, changes its
own density and browser storage, and deliberately leaves those values for the
next fresh actor/context to resist. Existing product selections exercise the
consequential helper families; native failure fixtures cover quarantine,
reset/cancellation/owner loss and physical settlement. No product test retry or
weakened assertion is introduced.

The repaired graph selection `r03-s001-graph-02` passed (native 33410/0,
50,831 ms), before the new browser fixture cases were selected for execution.
Generation `r03-s001-generate-03` passed (75070/0, 10,310 ms); format-02
passed (16994/0, 8,055 ms). Exact architecture-policy selection
`r03-s001-architecture-01` passed (75784/0, 8,066 ms). Native rows
`harness.browser.unit.native_fixtures` and
`harness.browser.integration.native_nested_cancellation` passed together as
`r03-s001-native-01` (19398/0, 40,824 ms). Schema check
`r03-s001-schema-01` passed (99461/0, 6,699 ms). Commands retain the E-031
cgo-disabled/cache-off environment and basic diagnostics for test slices.
These focused passes are not S-001 completion; product characterization,
current projections, complete evidence-reader checks and the frozen matched
performance experiment remain outstanding.

**E-033 — S-001 exact selections and first product-boundary receipts.**
Public planning requires `INPUTS=<JSON path>`, not a forwarded `SUITE` variable.
The attempted direct-SUITE plans were rejected with usage_error; the first
redirected output was empty and unusable. The corrected
`make target-plan TARGET=browser-e2e
INPUTS=.cartulary/remediation03/s001-support-inputs.json JSON=1` and corresponding
functional input both exited 0 (native 86925/0 and 32643/0). The files contain
only the declared `SUITE` input. These planning exports use ordinary defaults;
measurement runs will explicitly select the frozen accounting/diagnostics/cache
settings from Section 8.2.

The exact support suite is unchanged: 18 rows/cases, 17 test-owned and one
isolated, row-ID SHA-256
`dd5a91a31b4738e9dd22623e6bde57af710facaa8471a836c3f5c7320212924b`.
The public functional suite now selects 383 rows and 436 cases: 327 test-owned
and 56 isolated, row-ID SHA-256
`22d63375b0e84f549957a468db323ce04c2aa58669e3a7e8bb842a0402a4c362`.
The earlier E-019 count described stage inventory before public eligibility,
not the exact executable suite. Exact selectors, owners and verification IDs
are retained in `.cartulary/remediation03/s001-support-selection.json` and
`s001-functional-selection.json`. The candidate plans contain two support
cohorts and 65 functional cohorts; these are plans, not observed setup counts.

The four new fixture-boundary cases are common assurance in both experiment
treatments. Their authored baseline catalog and file hashes are recorded in
`.cartulary/remediation03/s001-baseline/common-assurance.json`. This correction
precedes the performance freeze; no performance sample exists or is relabeled.

Current source is
`sha256:86f06909e6816f4b2f282a970df0bf56b197fa677925ba6944bce4ec869c05e1`.
`r03-s001-state-all-01` passed all four literal cases in one cohort (native
49128/0, 97,439 ms; 23 completed cleanup steps). A alone passed as
`r03-s001-state-a-alone-01` (94136/0, 92,015 ms); B alone passed as
`r03-s001-state-b-alone-01` (35365/0, 91,727 ms). Separate A→B and B→A selections
passed as `r03-s001-state-ab-01` (11514/0, 90,707 ms) and
`r03-s001-state-ba-01` (75512/0, 90,861 ms). Each uses exact E-032 row IDs,
`test-slice OWNER=harness.browser`, cgo disabled, result cache off and basic
diagnostics. Concurrent identical selections and the adapter failure matrix
are active at this checkpoint; the containing functional run will additionally
exercise multiple lanes within one run-owned broker.

Public `harness-observability-check CHECK=all` passed for state-all, both alone
runs, native-01, architecture-01, schema-01, generated-01 and graph-02. The
latter graph receipt binds earlier source
`sha256:ff75680e483c68912484d74777d41ac4ae6ced148a4376afee96450ce899d998`;
it is not silently reclassified as current source. The schema/generated checks
ran without optional diagnostics, which their validated reports state.
Dossier Markdown lint passed (35014/0, 15,800 ms); generated-01 passed
(96913/0, 16,634 ms). No optimization performance acceptance is claimed.

**E-034 — S-001 concurrent settlement and containing-workload pilot.**
The two identical A→B selections ran concurrently and passed:
`r03-s001-state-concurrent-1-01` (native 36403/0, 94,914 ms) and
`r03-s001-state-concurrent-2-01` (56121/0, 95,188 ms). Their independent run
identities preserve separate ownership; no source or capacity override changed.
Adapter/preparation/lane-policy selection `r03-s001-adapter-01` passed
(62235/0, 10,052 ms), including settled sibling preservation after interruption
and worker/global teardown failure. Public `CHECK=all` also passed for adapter-01
and both ordered selections. The two concurrent runs' audits remain next.

Register an exploratory candidate containing-workload run before launch:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s001-functional-pilot-01
CARTULARY_HARNESS_CACHE_MODE=off make browser-e2e SUITE=functional
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. It uses the E-033 public
383-row/436-case set, source `86f06909e6816f4b2f282a970df0bf56b197fa677925ba6944bce4ec869c05e1`,
four default lanes, one worker, existing deadlines/capacities and required
invocation counters. Prior fixture/adapter commands have settled before launch.
This run is assurance and measurement-feasibility evidence only; it is excluded
from all acceptance pairs and is not a substituted warm-up. No competing
intentional workload will be launched during it. Terminal/native/scan/cleanup
and reader results remain pending.

**E-035 — S-001 remaining admission correction and analysis precision.**
During the unchanged functional pilot, source inspection found that
`normalizeGroup` still locally defaults a missing runtime-profile ID even
though v15 requires that field. `validateBrowserBatchManifestShape` checks a
handwritten structural subset, not the complete v15 schema. A valid generated
manifest is unaffected, but malformed required metadata must not reach this
fallback. The candidate therefore remains unaccepted. After the active native
34099 pilot settles, make the reader enforce the canonical schema and add
missing-required-field/type rejection coverage; retain semantic checks such as
runtime/service agreement and sorted exact selectors. Do not change the active
run's source or reinterpret it as evidence for that later repair.

Before any performance freeze or sample, Section 8.2 now fixes the previously
implicit PRNG and percentile conventions: Python MT19937, one shared whole-pair
index matrix across all five metrics, nearest-rank bounds, reset per workload.
The internal `.cartulary/remediation03/performance-analysis.py` performs only
arithmetic over independently validated public performance exports and recorded
native/audit exits. It checks exact source/run identity, settled required
accounting, final scans, twelve pairs/order, and excluded warm-ups. It does not
read raw events, replace a canonical reader, relax changed-source comparison
admission, or establish treatment fidelity. Its numerical result is separate
from overall workstream acceptance. No real comparison has been analyzed.

**E-036 — S-001 pilot assertion diagnosis before source repair.** The live
pilot reported a product/test_assertion_failure in cohort
`4c02b087bc125a226729ee92bff00096eb1539e403ba6f3bb8ad5d49682b3837`.
Its published v11 group result identifies exactly one failed row,
`module.workbook.browser.find_accessible_geometry`; subsequent unexecuted rows
retain skipped_dependency outcomes. The referenced Playwright report identifies
`timeline-find.spec.ts:687`: the test expected one matching Recorded cell and
received two. The referenced trace's three Timeline query responses agree that
the test's own rows 0000 and 0002 both have Recorded value
`2026-10-21T15:22:15.626328Z`; row 0001 has
`2026-10-21T15:22:15.626429Z`. These are the three concurrently created fixture
rows, not foreign state. Core 03 §13.5 counts each matching committed cell;
timestamp uniqueness is not a requirement. The source uses a real clock and
does not enforce timestamp uniqueness.

After native 34099 settles, derive the expected count from the existing query's
committed Recorded values and retain the read-only focus assertion. Carry this
assurance correction into both experiment treatments before freezing sources;
add neither waits nor clock mutation. The prior failing attempt remains retained
and excluded from acceptance. A public `explain-run ... TARGET=browser-e2e
DETAIL=children` attempted before aggregate publication exited 2 because the run
summary was not yet available. Diagnosis instead followed the failed unit's
explicit public-plan output references and the published group's report/trace
references; no private event staging was read. At this checkpoint the owning
command, aggregate settlement, final scan and diagnostic audit remain pending.

**E-037 — S-001 pilot terminal boundary and S-008a corrective entry.** Native
34099 completed with outer Make exit 2. The graph failed with
product/test_assertion_failure after 1,143,639 ms. Its second failed row is
`harness.browser.boundary_support.frontend_artifact_lifetime`, which retains
isolation and invokes two independent public frontend builds. Its helper removes
all inherited `CARTULARY_*` variables. Native graph entry correctly recognized
the accounting hierarchy but rejected the absent cache-off setting before
registering either child graph. This is a required-accounting implementation
gap, separate from E-036's fixture timestamp assumption; no batching cause is
claimed. The source confirms that required kernel membership still tests the
cleared ambient cache/diagnostic settings instead of inheriting the proven
outer policy.

All 278 published cleanup steps completed; required accounting reports empty
membership and child exit 2, with final scan and terminal receipt present.
`harness-observability-check CHECK=diagnostics` and the public performance export
each exited 0. This validates diagnostic integrity, not passing assurance.
The export reports 1,156,250.031356 ms measured child time, 3,885,940,265 charged
CPU microseconds, 8,248,754,176 peak charged bytes, 1 ms p95 queue delay and
50,802 ms cleanup union. These are excluded pilot observations, not comparison
samples. Both E-034 concurrent runs also passed their outstanding `CHECK=all`
audits (reader exits 0).

Open separate corrective workstream **WF-08a / S-008a** now, before resuming
S-001 implementation. Its dependency is this settled pilot; its owner is
REQ-856/AC-152 and R03-018/V-009. Adopted clarification requires kernel/owner
proved descendants to inherit basic diagnostics and disabled execution-result
reuse despite environment clearing. Unaccounted commands retain their defaults.
Repair native graph entry, validate the existing independent-publication row
under required accounting and its ordinary mode, and repeat nested cancellation
coverage. Require three registered graphs for the publication row, exact scans,
native exits, completed cleanup and empty charged membership. An unproved scope
must still fail closed. No new command, selector, host setting or accounting
fallback is authorized or needed. Record this child workstream's completion and
lint before resuming S-001's reader and assertion repairs. Capability acceptance
for this newly exposed path is open; historical E-030 results are not erased.

**E-038 — S-008a completed; S-001 resumption boundary.** REQ-856/AC-152 now
explicitly owns policy inheritance for proved nested scope membership. Native
graph entry restores basic diagnostics, disabled result reuse and the exact
outer registration root only after validating kernel membership, live owner and
matching executable content. Unaccounted commands retain existing defaults.
The test-operations recipe describes that behavior. No schema, public selector,
capacity or deadline changed. This accepted correction is common to both later
S-001 comparison treatments; no compatibility mode or rollback is retained.

All following roots are under `.cartulary/test-results/` and bind source
`sha256:65853a520daa8d8e31a0ac99bf98b1e908c52c39a7fbbf113cd2698bdc497bac`.
Commands use `CGO_ENABLED=0`, their exact `CARTULARY_TEST_RUN_ID`, result cache
off and basic diagnostics. `make format` passed as `r03-s008a-format-01`
(native 90027/0, 8,435 ms). `make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.unit.native_runtime` passed as
`r03-s008a-native-01` (68854/0, 53,040 ms; ten completed cleanup steps).

`make test-slice OWNER=harness.browser
ROWS=harness.browser.boundary_support.frontend_artifact_lifetime
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` passed as
`r03-s008a-frontend-accounted-01` (75049/0, 89,835 ms; 23 completed cleanup
steps). Its accounting record binds exactly three registered graphs with all
three scans passed, child exit 0 and empty membership. The same exact row with
accounting off passed as `r03-s008a-frontend-ordinary-01` (82393/0, 87,111 ms;
23 completed cleanup steps). The unchanged helper still clears all inherited
`CARTULARY_*` variables and checks both independent Make exits.

`make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.native_nested_cancellation
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` passed as
`r03-s008a-nested-01` (48015/0, 40,066 ms; ten outer cleanup steps). Its four
registered graphs preserve all three deliberate child cancellations and the
successful enclosing assurance result; every scan passed and charged membership
is empty. Public `harness-observability-check CHECK=all` passed for all four
test invocations. Both accounted runs' public performance exports passed reader
validation; the exports are capability evidence, not optimization samples.
Assurance is accepted for S-008a; no performance improvement is claimed or
required for this correctness repair. E-037's failed pilot remains retained.
Record successful boundary Markdown lint before applying S-001's separate
v15 reader and timestamp assertion corrections. S-007 remains outstanding;
retained maintenance has no eligible successful full warm-check root.

Boundary lint `r03-s008a-boundary-lint-01` passed (native 8548/0, 9,550 ms).
This receipt was recorded before S-001 source work resumed.

**E-039 — S-001 admission/assertion repairs and current focused assurance.**
The v15 reader now invokes the canonical compiled schema before normalization,
and missing runtime-profile metadata has no default. Independent malformed
fixtures exercise missing required execution fields and incorrect scalar/array
types. Semantic compatibility and exact-selector checks remain in force.
The E-036 Timeline assertion now derives the exact matching Recorded-cell count
from its existing own-incident query. It retains read-only focus, collection,
geometry and native-control assertions without additional requests, waits or
fixture setup. Both performance treatments carry that assurance correction and
the accepted E-038 accounting repair; common-file hashes and the amended static
source dossier record this before the comparison freeze.

Current executable source is
`sha256:e08d6a7ea145e078779f5c48d9a0775fca54a913554de8ed2fc6a33e5fe3e385`.
All commands retain the cgo-disabled/cache-off profile; test slices use basic
diagnostics. `make generate` passed as `r03-s001-generate-04` (17363/0,
10,180 ms); `make format` passed as `r03-s001-format-03` (79916/0, 8,115 ms).
`test-slice OWNER=harness.browser` selecting exactly
`harness.browser.unit.browser_lane_policy`,
`harness.browser.boundary_support.playwright_adapter` and
`harness.browser.boundary_support.preparation_contract` passed as
`r03-s001-admission-03` (81090/0, 9,844 ms).
`source-check CHECK=schema` passed as `r03-s001-schema-02` (84681/0, 7,111 ms),
and `source-check CHECK=generated` passed as `r03-s001-generated-02`
(94548/0, 16,791 ms).

`test-slice OWNER=module.workbook
ROWS=module.workbook.browser.find_accessible_geometry` passed as
`r03-s001-timestamp-01` (68186/0, 83,707 ms). The exact routing-contract
selection from E-031 passed as `r03-s001-graph-03` (50365/0, 52,934 ms).
Every one of these five test/source-check invocations passed its public
`harness-observability-check CHECK=all` audit; schema/generated runs explicitly
lack optional diagnostics. The current support/functional public plans retain
the E-033 exact selections and graph digests; two and 65 candidate cohorts are
still planned, not a performance conclusion. Explicit visual validation is
active as `r03-s001-visual-01`, native 12191, using
`make browser-e2e SUITE=visual HARNESS_DIAGNOSTICS=basic`; no golden update.

Before any performance sample, numerical arithmetic was made explicitly decimal
(precision 28, HALF_EVEN) so an exact 5% improvement cannot pass a strict >5%
gate through binary rounding. Synthetic arithmetic checks cover inclusive 5%
guardrails, strict benefit thresholds, zero-baseline behavior, quantile indexes
and missing/invalid observations. Those checks validate internal calculation
only; they are not repository verification or run evidence. Authored candidate
inputs are snapshotted under `.cartulary/remediation03/s001-candidate/` before
any treatment switch. No source freeze or acceptance sample has occurred yet.

**E-040 — S-001 visual assurance and baseline preparation entry.** Explicit
`r03-s001-visual-01` completed successfully (native 12191/0, 627,924 ms), with
49 selected rows in 14 cohorts. Public `harness-observability-check CHECK=all`
passed. No golden or product presentation changed. All current candidate work
has settled before the next source transition. The archived authored v14
baseline will now be restored only across the recorded treatment paths and
regenerated through Make, retaining common E-036/E-038 corrections. Candidate
source `e08d6a7ea145e078779f5c48d9a0775fca54a913554de8ed2fc6a33e5fe3e385`
and its authored snapshot remain available for exact restoration. This is
experiment preparation, not a compatibility mode or accepted rollback. Baseline
source identity and independent schema/generated validation must be recorded
before the comparison freeze and excluded warm-ups.

**E-041 — S-001 exp01 comparison frozen before launch.** Baseline generation
passed as `r03-s001-baseline-generate-01` (83475/0, 10,120 ms). Serial native
preparation/generated validation passed as `r03-s001-baseline-generated-01`
(31070/0, 15,794 ms); schema validation passed as
`r03-s001-baseline-schema-01` (16814/0, 6,107 ms). Both public `CHECK=all`
audits passed. The baseline source is
`sha256:7c710b89542be7871d494522a63c8a04be1d9e42ebc99f506cb8bd958810948f`;
candidate source remains
`sha256:e08d6a7ea145e078779f5c48d9a0775fca54a913554de8ed2fc6a33e5fe3e385`.
Baseline native executable SHA-256 is
`afa20e4c13301ce2e3fd9f932612ec2582d171dc342f40ff1c46fc39c78afdbf`.
Both source manifests report toolchain digest
`54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and host capability digest
`008fe637b622cd8932b55cc1d8be5346bae02c09ab35d7518fc9f72b66eb6c21`.

Independent public plans have identical row metadata and common catalog/selectors:
18 support cases and 383 functional rows/436 cases. Baseline/candidate planned
cohort counts are 5/2 for support and 80/65 for functional. Baseline graph
digests are `afe3bbcd826886670fb3f0b33f6d1bfaad32c2e0cde62fea2480abfc45f02dfe`
and `0174590986cf12fbac2f324e35ff1ff8381d2f7536d3de717ea22d41040eb2d4`;
candidate graph digests remain those recorded with E-033 plans. These are still
planning counts, not observed speed or resource benefits.

The immutable internal registration is
`.cartulary/remediation03/s001-exp01-freeze.json`, SHA-256
`36f7d750ddef92ce8655a45b2e242e82e2245c7c7d79aa27baf7cf582ed7a3d1`.
It binds eleven authored differences, deterministic generated projections,
common corrections, exact selection/environment inventories and all 52 fresh
run IDs: four excluded warm-ups, then twelve matched pairs per workload.
Each treatment block runs support then functional. Warm-up order is baseline
then candidate; pair order alternates baseline-first/candidate-first. Sources
are restored only from the recorded authored snapshots and regenerated through
Make, with serial cgo-disabled native preparation outside measurement. No source
mutation or competing intentional workload occurs during measured commands;
only native-session observation is used until command exit. Public audits and
exports follow each run. The existing equivalent-comparison contract is unchanged.

Section 8.2 supplies the acceptance rule; the freeze explicitly preserves
charged-invocation CPU/memory scope and complete selected-graph queue/cleanup
scope. All earlier attempts, including the failed E-037 pilot, are excluded and
retained. `.cartulary/remediation03/s001-exp01-receipts.json` records actual
native exits, independent audit/export results and failures without modifying
the freeze. The active treatment is baseline. Start with excluded warm-up
`r03-s001-exp01-support-warm-b`, then `r03-s001-exp01-functional-warm-b`.
No acceptance sample or performance decision exists at this checkpoint.

**E-042 — S-001 exp01 stopped on its first excluded warm-up.** The exact E-041
command `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s001-exp01-support-warm-b
CARTULARY_HARNESS_CACHE_MODE=off make browser-e2e SUITE=support
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` ended native 15647/2,
graph failure after 214,049 ms. The root is
`.cartulary/test-results/r03-s001-exp01-support-warm-b`; the failing row is
`module.timeline.browser_support.reviewed_edit_demotion_preserves_query_filter_safety_0e2d9a8f10`.
The published Playwright report classifies a product test assertion failure:
the Beta summary click times out after its virtual cell is detached. The pending
PATCH waiter then observes page closure; that is a consequence, not the cause.
All 40 cleanup steps completed. Public `CHECK=diagnostics` and the performance
export succeeded; accounting reports empty physical settlement. Neither reader
exit converts the failed test into an assurance pass.

The referenced trace places the click at monotonic 52,533.838 ms while the
grouped query starts at 52,498.391 ms and completes 49.188 ms later. The test
awaits only request dispatch before scrolling; the later accepted result resets
the horizontal virtual viewport. Its terminal snapshot retains the expanded
reviewed group and Beta row but no mounted synopsis cell. The adjacent anchor
stability scenario already awaits the exact query response and the existing
`data-workbook-query-pending=false` observation before scrolling. Applying that
same boundary to this test preserves all edit, anchoring, demotion and filtered
empty-state assertions without extra requests, sleeps, retries or setup.

Exp01 remains immutable and unsuccessful: its receipt register preserves this
failed warm-up and 51 unlaunched entries; no acceptance sample was collected.
Baseline v14 is active. The synchronization correction is common to both
treatments, requires focused qualification, updated source-effect references
and a fresh exp02 freeze/source identities before any further comparison.

**E-043 — Common query correction qualified; exp02 frozen.** Formatting passed
as `r03-s001-query-format-01` (57860/0, 8,373 ms). Exact-row baseline qualification
used `make test-slice OWNER=module.timeline
ROWS=module.timeline.browser_support.reviewed_edit_demotion_preserves_query_filter_safety_0e2d9a8f10`
with the E-041 cgo/cache/diagnostic/accounting profile:
`r03-s001-query-baseline-01` passed (68847/0, 80,123 ms). The full support
selection then passed on baseline as `r03-s001-query-support-baseline-01`
(2466/0, 144,825 ms). Both `CHECK=all` audits and public performance exports
passed; physical settlement was empty. These are qualification receipts only.

After verifying the current baseline and archived candidate authored hashes,
the twelve recorded treatment paths were restored to candidate. Common source
corrections remained in place; generated outputs were rebuilt through
`make generate`, `r03-s001-query-candidate-generate-01` (37284/0, 9,980 ms).
`make source-check CHECK=generated` passed as
`r03-s001-query-candidate-generated-01` (47957/0, 15,583 ms), including its
`CHECK=all` audit. Candidate full support passed as
`r03-s001-query-support-candidate-01` (47646/0, 119,280 ms); its `CHECK=all`
audit/export passed and physical settlement was empty. All 18 cases remain
selected; the baseline/candidate graphs retain five/two cohorts. No performance
acceptance is inferred from these qualification timings.

The updated effect dossier binds `timeline.support.spec.ts` SHA-256
`763a8784eff203c0a9ea479b51b742438156b7f39096cea403230d07da19f52e`
and adjusts its source references. This is a test synchronization correction,
not an isolation relaxation or product-query change. The accepted-query
observation was already an existing implementation boundary; no new behavior
or runtime dependency on this document is introduced.

The new immutable registration is `.cartulary/remediation03/s001-exp02-freeze.json`,
SHA-256 `0aae9f26eb2cd8c927f3b9daefca2d7455c9477aa5570304207d6553220a52a3`.
Its baseline source is
`sha256:0f5954f9c17a12525e06a2ad4a7c0ba54f02d337e9c10732067de3917b6f8050`;
candidate source is
`sha256:51a33fca20be115ddc82652e6befefdfcd2ea65995e096c434602ff7c8470acb`.
Environment, toolchain/capability hashes, selections, treatment differences,
graph identities, statistical method and acceptance rules remain fixed as in
E-041. Fresh plans confirm unchanged candidate row metadata and graphs. The
new common-correction inventory is `s001-exp02-common-assurance.json`;
exp01 and all its referenced inventories remain unchanged.

Exp02 has 52 fresh run IDs: warm-ups run candidate support/functional, then
baseline support/functional, followed by twelve alternating matched pairs for
each workload, baseline-first on pair 1. This declared warm-up order starts
from the qualified active candidate and leaves baseline ready for pair 1.
Every owning exit, audit, export and failure belongs in
`.cartulary/remediation03/s001-exp02-receipts.json`. Only native-session
observation occurs during these runs. No exp02 run has launched at this receipt;
the next command is excluded candidate warm-up `r03-s001-exp02-support-warm-c`.
Pre-launch `make lint-markdown` passed as `r03-s001-exp02-freeze-lint-01`
(21673/0, 8,420 ms).

**E-044 — Exp02 candidate warm-ups passed; baseline transition.** Both commands
used E-043's frozen profile and selections. `r03-s001-exp02-support-warm-c`
passed (23871/0, graph 120,247 ms), with all 27 cleanup results completed.
`r03-s001-exp02-functional-warm-c` passed (73674/0, graph 1,138,335 ms),
with all 276 cleanup results completed. Each public `CHECK=all` audit and
performance export passed. The functional invocation registered three graphs;
all three retained-secret scans passed and final accounting settlement was
empty. This also establishes containing-workload assurance for the E-036
assertion and E-038 nested-accounting corrections.

The candidate warm-up exports report child-Make elapsed times of
124,833.726988 / 1,150,589.222544 ms, charged CPU of
166,795,765 / 4,008,060,162 microseconds, peak charged memory of
2,455,134,208 / 8,469,929,984 bytes, selected-graph p95 ready-to-start delay
of 0 / 1 ms and cleanup unions of 3,374 / 47,882 ms, respectively. These
excluded warm-ups are not acceptance samples or an improvement claim.
Exact export hashes and native/audit exits are retained in the exp02 receipt
register; the freeze is unchanged.

All candidate commands settled before the verified authored baseline restore.
Baseline regeneration passed as `r03-s001-exp02-baseline-generate-01`
(26221/0, 10,150 ms). Serial generated/native preparation is next, then
`r03-s001-exp02-support-warm-b` and `r03-s001-exp02-functional-warm-b`.
The active checkout is temporarily baseline v14; the validated candidate source
and all common corrections remain recoverable through the recorded treatment
snapshots. S-001 is still in progress; no paired sample has run.

**E-045 — Exp02 warm-up phase complete; paired sampling entry.** Baseline
generated/native preparation passed as `r03-s001-exp02-baseline-generated-01`
(52858/0, 15,807 ms), including `CHECK=all`; the restored source exactly
matches E-043. Baseline support warm-up `r03-s001-exp02-support-warm-b`
passed (13126/0, graph 146,390 ms), with 39 completed cleanup results.
Baseline functional warm-up `r03-s001-exp02-functional-warm-b` passed
(26172/0, graph 1,750,922 ms), with 336 completed cleanup results. Both
`CHECK=all` audits and performance exports passed. The functional invocation's
three registered graphs all passed their retained-secret scans; physical
settlement was empty.

Baseline support/functional accounting reports child-Make elapsed times of
151,001.933338 / 1,763,551.831717 ms, charged CPU of
194,571,511 / 4,565,945,313 microseconds, peak charged memory of
2,463,600,640 / 8,796,626,944 bytes, selected-graph p95 ready-to-start delay
of 1 / 1 ms and cleanup unions of 5,263 / 56,084 ms. The four warm-ups are
excluded from acceptance. Their manifests match the frozen source, graph,
selection, four-lane, no-result-cache, basic-diagnostics and required-accounting
inputs. Exact receipts and export hashes remain in the exp02 receipt register.

Both containing workloads now have passing warm-up assurance. This does not
close G-04: twelve matched pairs per workload remain required. The next two
commands are `r03-s001-exp02-support-p01-b` and
`r03-s001-exp02-functional-p01-b` on the current baseline, followed by a
verified candidate restore/regeneration and the corresponding candidate pair.
No acceptance sample has launched at this receipt; the frozen schedule and
analysis method remain unchanged.

**E-046 — Exp02 Pair 1 complete for support and functional workloads.** Every
command below used E-043's frozen literal run ID, `make browser-e2e
SUITE=<workload>`, cgo-disabled preparation, disabled execution-result reuse,
basic diagnostics and required accounting. Each native command exited 0, each
public `CHECK=all` audit/export exited 0, all retained-secret scans passed and
all physical accounting scopes settled empty.

| Run ID (under `.cartulary/test-results/`) | Native session / exit | Graph elapsed ms | Completed cleanup results |
| --- | --- | --- | --- |
| `r03-s001-exp02-support-p01-b` | 22717/0 | 146,696 | 39 |
| `r03-s001-exp02-functional-p01-b` | 66933/0 | 1,744,127 | 336 |
| `r03-s001-exp02-support-p01-c` | 38942/0 | 122,255 | 27 |
| `r03-s001-exp02-functional-p01-c` | 89662/0 | 1,133,483 | 276 |

Both functional observations include three registered graphs with passing
retained-secret scans. Their manifests match the frozen graph/source identities
and all declared execution inputs. The independent public exports and their
hashes are recorded in `.cartulary/remediation03/s001-exp02-receipts.json`.
The reduced cohort counts are observed in actual successful execution, but
one pair establishes neither a statistical bound nor G-04 acceptance.

Between baseline and candidate, the authored treatment hashes and common
corrections were checked before restoration. Candidate generation passed as
`r03-s001-exp02-candidate-generate-01` (43519/0, 10,130 ms); serial
generated/native preparation passed as `r03-s001-exp02-candidate-generated-01`
(74197/0, 15,878 ms), including `CHECK=all`. The resulting source exactly
matches E-043's candidate identity. These preparation commands occurred outside
the measured invocations.

The active source is candidate v15. Eleven matched pairs per workload remain.
The next commands are `r03-s001-exp02-support-p02-c` and
`r03-s001-exp02-functional-p02-c`, then a verified baseline restoration for
Pair 2's baseline observations. No source change, new experiment or performance
decision is introduced at this checkpoint.

**E-047 — Exp02 Pair 2 complete, candidate-first as preregistered.** All four
owning commands below exited 0. Their independent `CHECK=all` audits and public
performance exports passed. All cleanup results completed, all retained-secret
scans passed, and every invocation's physical settlement was empty. Both
functional observations registered three graphs. Source, graph, selection and
diagnostic/admission inputs match the E-043 freeze.

| Run ID (under `.cartulary/test-results/`) | Native session / exit | Graph elapsed ms | Completed cleanup results |
| --- | --- | --- | --- |
| `r03-s001-exp02-support-p02-c` | 47878/0 | 118,782 | 27 |
| `r03-s001-exp02-functional-p02-c` | 18625/0 | 1,128,697 | 276 |
| `r03-s001-exp02-support-p02-b` | 57393/0 | 144,182 | 39 |
| `r03-s001-exp02-functional-p02-b` | 95866/0 | 1,771,730 | 336 |

The candidate-to-baseline transition verified authored snapshot hashes and
common-correction identities before restoring the recorded paths. Baseline
generation passed as `r03-s001-exp02-baseline-generate-02` (33956/0,
10,010 ms); serial generated/native preparation passed as
`r03-s001-exp02-baseline-generated-02` (46442/0, 15,594 ms), including
`CHECK=all`. The baseline source identity matches the freeze exactly.

The receipt register now contains four excluded warm-ups and eight valid
measured runs, forming two complete pairs per workload. G-04 remains open
until all twelve pairs and the fixed numerical
analysis are complete. The active treatment is baseline. Next are
`r03-s001-exp02-support-p03-b` and `r03-s001-exp02-functional-p03-b`,
with no intervening source change required by the frozen alternating order.

**E-048 — User-authorized performance protocol amendment.** After E-047, the
user reduced the total matched baseline/candidate pairs per declared workload
from twelve to **four**, then required full performance validation only when
both of the first two runs show less than 20% improvement. The user clarified:
“Remaining performance protocol only; keep correctness and cleanup checks
required.” Each workload uses its first two matched pairs independently; either
improvement at or above 20% permits early exit. Warm-ups remain excluded. Full
validation retains the original 10,000 resamples, seed 20261021, one-sided 99%
guardrail bounds and one-sided 95% benefit bound, now with four whole-pair draws.
No runtime requirement, source treatment, comparison-reader admission or
correctness/finalization requirement changes.

Original exp02 registration remains byte-for-byte intact at
`.cartulary/remediation03/s001-exp02-freeze.json`, SHA-256
`0aae9f26eb2cd8c927f3b9daefca2d7455c9477aa5570304207d6553220a52a3`.
The explicit amendment is
`.cartulary/remediation03/s001-exp02-protocol-amendment-01.json`, SHA-256
`dd3471ca919464c1c903ac289f52588e764166ecef42333479a33cedd7c5bedf`.
It retains all thirteen completed receipts, omits thirty-two unlaunched runs
from pairs 5–12 by the four-pair amendment, and omits three further functional
runs by early exit. Historical twelve-pair statements in E-035/E-041–E-047 keep
their original meaning and are superseded for future acceptance by this receipt.

| Workload | Pair 1 elapsed improvement | Pair 2 elapsed improvement | Required remaining performance work |
| --- | --- | --- | --- |
| Support | 16.2632% | 17.0110% | Complete pairs 3 and 4 and full numerical validation |
| Functional | 34.7486% | 36.0983% | None under the authorized early-exit rule; no four-pair confidence-bound claim |

The improvements use the owning-command accounting durations in E-046/E-047,
not graph elapsed times. Both initial pairs retain independent passing native
exits, public audits, terminal scans and settled cleanup. All thirteen completed
run manifests were also checked against frozen toolchain, system, capability,
source, graph and declared-input identities; all match.

Support pair 3 baseline already passed as
`r03-s001-exp02-support-p03-b` (41754/0, graph 151,267 ms), with 39 completed
cleanup results, passing `CHECK=all` and performance export, passed scans and
empty physical settlement. Functional pair 3 baseline was already active as
`r03-s001-exp02-functional-p03-b`, native 3574. It continues as supplemental
assurance only; its timing is excluded from acceptance. Document and internal
arithmetic work after this amendment is consequently outside a scored sample.
Its terminal outcome and cleanup still must be recorded before source changes.

The old arithmetic script and its frozen hash remain unchanged. The amended
`.cartulary/remediation03/performance-analysis-v2.py` has SHA-256
`97a25d016748724d67ba38676aeccff67fc96b11a4bde114a7c88c0128cca653`.
Internal arithmetic checks passed for the inclusive 20% boundary, first-two-only
decision, incomplete/zero-baseline rejection, exact 5% guardrail acceptance,
over-5% rejection, strict greater-than-5% benefit and four-pair completeness.
These checks are analysis validation, not repository test assurance. Original
exports, source snapshots, failed attempts and immutable registrations remain.
Next: settle the active supplemental run, finish support pairs 3/4 in their
original alternating order, and record the S-001 assurance/performance decision.

Protocol-amendment Markdown lint passed through `make lint-markdown` as
`r03-protocol-amendment-lint-01` (native 93932/0, 10,840 ms); its root is
`.cartulary/test-results/r03-protocol-amendment-lint-01`. No product or generated
verification was substituted for this document check.

**E-049 — WF-01 / S-001 accepted; candidate restored.** F-001 and
RB-002/RB-003 are resolved for this slice. REQ-806/AC-088 own the reporting-owner
separation; R03-001–006 map to the shared compatibility projection, v15
manifest/reader/schema cutover and exact-selection cohort compiler. Catalog
ownership and verification attribution remain, while execution ownership stays
`harness.browser`. No owner exception, compatibility toggle, dual reader,
concurrency increase or isolation relaxation was introduced.

Assurance decision: **PASS for S-001**. E-032/E-039/E-042 trace the affected
rows/helpers and common assertion corrections. E-033/E-034 cover alone,
A→B/B→A, concurrent invocations and native failure/settlement cases; E-039
records strict schema/reader, exact selection, adapter, preparation, routing and
generated checks. E-040 supplies explicit visual assurance. The corrected
support and containing functional populations pass in E-044–E-049. All seventeen
completed exp02 invocations have independently audited, exact selected-row and
passed-row membership, with no duplicate or missing selected row; the retained
membership confirmation is
`.cartulary/remediation03/s001-exp02-membership-confirmation.json`.
Source-scoped earlier checks retain their recorded identities; S-007 current
integration validation is still required.

The final exp02 runs below use the E-043 command profile:
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off`, an exact fresh
`CARTULARY_TEST_RUN_ID`, and `make browser-e2e SUITE=<support|functional>
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` from the repository root.
Each run's independent `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` and `make explain-run RESULTS_DIR=<exact-root> DETAIL=performance
JSON=1` exited 0. Run roots are under `.cartulary/test-results/`.

| Run ID | Native session / exit | Graph elapsed ms | Completed cleanup results | Disposition |
| --- | --- | --- | --- | --- |
| `r03-s001-exp02-support-p03-b` | 41754/0 | 151,267 | 39 | Valid sample, first recorded E-048 |
| `r03-s001-exp02-functional-p03-b` | 3574/0 | 1,774,092 | 336 | Supplemental assurance only; E-048 excludes its timing |
| `r03-s001-exp02-support-p03-c` | 57925/0 | 119,734 | 27 | Valid sample |
| `r03-s001-exp02-support-p04-c` | 71788/0 | 121,564 | 27 | Valid sample |
| `r03-s001-exp02-support-p04-b` | 75988/0 | 147,621 | 39 | Valid sample |

All terminal secret scans passed, all cleanup results completed, and all
invocation accounting scopes settled empty. The supplemental functional run
registered three settled graphs. Its post-amendment companion authoring and
status observations are not included in acceptance timing. Source, graph,
selected inputs, toolchain/system and capability identities match the frozen
treatments. Original exp01 failures and all 35 authorized exp02 omissions remain
recorded; no failed attempt was replaced with a favorable sample.

Performance decision: **PASS under the user-amended applicable routes**.
Support completed four matched pairs. Ratios below use the owning-command
accounting scope; a ratio of 1 means no change. All five 99% upper bounds are
at most 1.05. Its elapsed mean improves 17.8069%; the optional descriptive 95%
lower benefit bound is 16.6338%. S-001 did not require a positive benefit gate.

| Support metric | Candidate / baseline means | One-sided 99% upper ratio | Guardrail |
| --- | --- | --- | --- |
| Owning-command elapsed | 0.821931 | 0.835523 | PASS |
| Charged CPU | 0.836061 | 0.849826 | PASS |
| Peak cgroup memory charge | 1.009134 | 1.019298 | PASS |
| Per-run p95 ready-to-start delay | 0.750000 | 1.000000 | PASS |
| Cleanup union duration | 0.667414 | 0.694388 | PASS |

Functional uses the first two pairs' 34.7486%/36.0983% elapsed improvements and
E-048's user-authorized early exit. Their observed CPU, peak memory and cleanup
ratios decrease; queue ratios equal 1. No omitted four-pair confidence bound
or whole-host resource guarantee is claimed. The public equivalent-comparison
reader's source admission remains unchanged.

Internal calculation inputs are
`.cartulary/remediation03/s001-exp02-analysis-registration.json`, SHA-256
`5ee7c099b94d9bced02589db14ba4f093dcea3daa00b2bbc1f5b71a8b18b2e48`;
results are `.cartulary/remediation03/s001-exp02-analysis.json`, SHA-256
`cd2c5f9341a743bfb9f3df39df13416fbd49bf545173d6353e62917d3fa429fb`.
The E-048 versioned arithmetic uses 10,000 whole-pair resamples and the frozen
seed and precision. This is internal remediation evidence, not Core 05
qualification. Audited published group results establish actual support
cohorts/resets of 5/4 → 2/1 and functional cohorts/resets of 80/76 → 65/61.
The original 55-to-17 model is neither asserted nor used for acceptance.

Each source transition first verified the active authored snapshot, target
snapshot and common-assurance hashes; generated outputs were regenerated
through Make. Generation/preparation receipts after E-048:

| Run ID | Native session / exit | Elapsed ms | Result |
| --- | --- | --- | --- |
| `r03-s001-exp02-candidate-generate-02` | 17330/0 | 10,500 | Generate passed |
| `r03-s001-exp02-candidate-generated-02` | 50499/0 | 16,288 | Generated check and public audit passed |
| `r03-s001-exp02-baseline-generate-03` | 32467/0 | 10,250 | Generate passed |
| `r03-s001-exp02-baseline-generated-03` | 78913/0 | 15,983 | Generated check and public audit passed |
| `r03-s001-accepted-generate-01` | 67211/0 | 10,270 | Accepted candidate regenerated |
| `r03-s001-accepted-generated-01` | 9283/0 | 15,923 | Generated check and public audit passed |

Final executable source is the accepted candidate,
`sha256:51a33fca20be115ddc82652e6befefdfcd2ea65995e096c434602ff7c8470acb`.
The v14 counterfactual is retired from the active tree; its snapshots and runs
remain historical comparison evidence. Current readers reject v14. Rollback
would require a settled, coordinated owner/schema/projection/source restoration;
no uncertain resources remain to drain at this boundary.

Long-term result: reporting-owner additions no longer create unrelated browser
compatibility partitions, while one explicit tuple retains execution/state
boundaries and exact reporting. Residual risk is later integration, covered by
mandatory S-007 checks rather than a slice-level claim of CI completion.
WF-02 may enter after this boundary's Markdown lint, with exact incident
pre-change characterization and an assertion/selector preservation map. No
incident source or routing has changed yet. Retained-run maintenance remains
skipped because `RESULTS_DIR` is unset: no eligible successful full warm-check
root has been established.

Boundary `make lint-markdown` passed as `r03-s001-boundary-lint-01`
(native 48084/0, 9,390 ms). This completion receipt was recorded before WF-02
investigation began.

**E-050 — WF-02 / S-002 characterization and candidate assurance.** After
E-049 boundary lint, the incident owner guide and exact public selections were
revalidated. Core 01 REQ-01-166/167 already require the original incident's exact
Location on replay/create; no behavioral specification or production change is
needed. REQ-01-157/160/162–164 retain normalization, duplicate-key and durable
side-effect assurance. Domain vocabulary and ownership remain unchanged.

Both existing affected rows passed before editing, at S-001's accepted source
`sha256:51a33fca20be115ddc82652e6befefdfcd2ea65995e096c434602ff7c8470acb`.
Commands use `CGO_ENABLED=0`, fresh explicit `CARTULARY_TEST_RUN_ID`,
`CARTULARY_HARNESS_CACHE_MODE=off`, and `make test-slice OWNER=module.incidents
ROWS=<exact-row> HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.

| Run ID under `.cartulary/test-results/` | Selection | Native / exit | Graph elapsed ms |
| --- | --- | --- | --- |
| `r03-s002-characterize-support-01` | Original five-symbol support row in Section 7.3 | 44667/0 | 22,122 |
| `r03-s002-characterize-durable-01` | Original Unicode/durable integration row in Section 7.3 | 82009/0 | 15,833 |

The existing public `make author-test-row-id` invocation with E-019's exact
family, claim and selector key returned the same survivor ID, ending
`authenticated_incident_creation_and_normalized_r_e2cd9c99b5` (exit 0).
The survivor now checks both existing 201-create and 200-replay responses against
`/api/v1/incidents/` plus the original returned incident ID. The standalone
Location test is removed. Its five-symbol row retains the other three surviving
symbols; the survivor has one exact-symbol row with the same owner,
verification, runtime/resource, fixture, isolation and assurance bindings.
All other eighty incident catalog rows remain identical.

The survivor's body is byte-identical after removing only the two new assertion
blocks: its four existing direct JSON requests, actor provisioning, normalization,
conflict checks, payload equality and one runtime/server setup are preserved.
The retired selector is absent from current routing. The separate durable
integration scenario is unchanged. The preservation record is
`.cartulary/remediation03/s002-assertion-preservation.json`; authored treatment
snapshots are under `s002-baseline/` and `s002-candidate/` in the same parent.

`make format` passed as `r03-s002-format-01` (51181/0, 7,939 ms), and
`make generate` passed as `r03-s002-generate-01` (28975/0, 10,000 ms).
Serial `make source-check CHECK=generated` passed as
`r03-s002-generated-01` (48017/0, 16,492 ms), with independent `CHECK=all` audit.
Candidate source is
`sha256:a06cf9eb03e26dbe53483bd3f9d1b39eeb721e936144c42d6a0eb35b71ff1532`.
Post-change exact selections at this source passed:

| Run ID | Selection | Native / exit | Graph elapsed ms |
| --- | --- | --- | --- |
| `r03-s002-survivor-01` | New survivor row alone | 77406/0 | 17,281 |
| `r03-s002-siblings-01` | Original row's remaining three symbols alone | 30863/0 | 18,374 |
| `r03-s002-durable-01` | Unchanged Unicode/durable row alone | 63078/0 | 21,267 |
| `r03-s002-combined-01` | All three exact rows above | 82461/0 | 22,451 |

Every pre/post test invocation passed independent public `CHECK=all` and
performance-export reads. Terminal scans passed and invocation settlement was
empty. These are characterization/assurance runs, never acceptance samples.
The focused workload's public plans preserve one Go invocation across the row
split. The declared containing services selection changes 49 → 50 reporting
rows (28 browser rows unchanged, 21 → 22 Go rows), while total work units remain
66 and Go invocations remain five. The native selector union changes only by
removal of the standalone Location test.

An inspection request using internal projection `TARGET=backend-integration-support`
was rejected with usage error/exit 2; the corrected public `TARGET=test-slice`
log explanation exited 0. This reader error is separate from the passing native
test commands. Catalog-owner and catalog-source checks are in progress; the
four-pair/conditional-early-exit comparison and exact fixture-lifetime evidence
remain required before S-002 acceptance. No later workstream has started.

**E-051 — S-002 comparison frozen before launch.** Catalog-owner selection
`harness.test_catalog.behavior.routing_contract` passed through `make test-slice
OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract
HARNESS_DIAGNOSTICS=basic` as `r03-s002-catalog-01` (2212/0, 50,676 ms).
`make source-check CHECK=catalog` passed as `r03-s002-catalog-source-01`
(35507/0, 1,700 ms). Both used `CGO_ENABLED=0`, disabled result reuse and fresh
run IDs; both independent public `CHECK=all` audits passed. Required catalog
assurance is complete; these checks are not performance samples.

The immutable registration is
`.cartulary/remediation03/s002-exp01-freeze.json`, SHA-256
`6e95fdbecd1ec255acae937552187d234e2e4ae0bf0e597983622c1f30c29352`.
It binds the E-050 baseline/candidate sources, the two authored-file snapshots,
permitted generated projections, exact native-selector treatment and public
plans, plus the current environment inventory
`s002-environment.json`, SHA-256
`7795bb9d7db45976978ab67611383828adaacae34b42faaeef8aa914a55e1262`.
Storage devices, toolchain/system/capability identities and authored environment
inputs match the revalidated baseline and candidate manifests.

The focused workload uses the original support row versus its two-row survivor
and sibling union. The containing workload remains `make test-slice
OWNER=module.incidents SCOPE=services`, including its browser population.
Both exact native-selector sets change only by retiring the standalone Location
symbol; no other native test is added or removed. Public focus graph digests:
baseline `sha256:eda2e432ba4e20e30355bf6eb6edf34ca356c2fdb826d09ec6caf95a0f3c84c8`,
candidate `sha256:42e339815188aa65cc95c53f6cdbc24dc59f83acb7237405556437f618d0c6d0`.
Containing graph digests:
baseline `sha256:bce467f46519e0657dc3804891cf930a17600e60ae8ae93e8eaf8a4e1b5fde6a`,
candidate `sha256:d995969665385e99f379b2af7463dbad887b2f7debf303f27959a4b12188a3ae`.

At most twenty invocations are registered: one excluded warm-up per treatment
and workload, then four matched pairs per workload in alternating B→C/C→B
order. Warm-ups run candidate focus/containing, then baseline focus/containing.
E-048's first-two-pair >=20% early-exit decision is now registered before any
sample; any resulting omission must be explicitly retained. Otherwise all four
pairs and the fixed 10,000-resample guardrail calculation remain required.
Both treatments use required invocation accounting, basic diagnostics, warm
build/dependency caches, disabled result reuse and unchanged capacities,
workers, lanes and deadlines. Only native-session observation occurs during a
measured owning invocation; independent audits/exports follow its terminal exit.
No S-002 performance acceptance is claimed at this registration boundary.
Registration Markdown lint passed as `r03-s002-freeze-lint-01`
(native 14623/0, 8,800 ms), before the first excluded warm-up launched.

**E-052 — S-002 excluded warm-ups complete; structural observation recorded.**
All four E-051 warm-ups passed their native commands, independent public
`CHECK=all` audits and performance exports. They use the exact registered
`test-slice` selections and required-accounting/basic-diagnostic/cache-off
profile. Every terminal secret scan passed, every cleanup result completed and
all invocation scopes settled empty. None is an acceptance sample.

| Run ID under `.cartulary/test-results/` | Native / exit | Graph elapsed ms | Completed cleanup results |
| --- | --- | --- | --- |
| `r03-s002-exp01-focus-warm-c` | 98275/0 | 18,784 | 14 |
| `r03-s002-exp01-containing-warm-c` | 86012/0 | 254,599 | 141 |
| `r03-s002-exp01-focus-warm-b` | 57903/0 | 19,650 | 14 |
| `r03-s002-exp01-containing-warm-b` | 66287/0 | 261,708 | 142 |

The exact focused native selectors executed five test fixture lifetimes in the
baseline and four in the candidate. Each selected top-level test has one
unconditional `appsupport.StartRuntime` and `StartDefaultServer` call; the
survivor's setup and request body remain identical apart from the assertions.
The unchanged helper creates and settles the test's application/database fixture.
Published, untruncated PostgreSQL checkpoint diagnostics corroborate five versus
four preparations, with one owned PostgreSQL service in each invocation. These
are different counts: removal of one test fixture does not claim removal of a
service container. The retained source/execution mapping is
`.cartulary/remediation03/s002-fixture-lifetime-evidence.json`.

The candidate-to-baseline transition verified both authored snapshots before
restoring the two baseline files. `make generate` passed as
`r03-s002-exp01-baseline-generate-01` (63361/0, 10,430 ms); serial
`make source-check CHECK=generated` passed as
`r03-s002-exp01-baseline-generated-01` (97247/0, 16,044 ms), including its
independent public audit. The restored source matches the E-051 baseline
exactly. Current active treatment is baseline; pair 1 runs baseline focus and
containing, then candidate focus and containing.

A receipt-authoring assertion initially expected an empty `ROWS` field for
`SCOPE=services`; the canonical manifest correctly omits that optional input.
The authoring expectation was corrected to the frozen selection's absence of
`ROWS`, then all declared inputs were checked exactly. No run was repeated and
no public evidence, reader or runtime behavior changed. Native and audit exits
remain independently recorded in `s002-exp01-receipts.json`.

**E-053 — User-required early rejection and S-002 first pair.** The user added:
“For performance-only analysis, reject the candidate if both of the first two
runs show less than a 5% improvement over the baseline.” Apply this per declared
workload to its first two independently audited matched pairs, excluding warm-ups.
Both improvements strictly below 5% reject the candidate; exactly 5% does not.
Either at least 20% still makes the remaining performance protocol optional.
Otherwise complete four pairs and the existing bootstrap gate. Correctness,
cleanup, evidence integrity and S-007 remain required. A later favorable pair
cannot replace either first pair or rescue an early rejection.

This explicit user amendment arrived during the first containing candidate
command and was recorded after its terminal exit and audit, before pair 2.
The original E-051 freeze remains unchanged. The additive record
`.cartulary/remediation03/s002-exp01-protocol-amendment-01.json` has SHA-256
`15c3bf8d3a9e1e690ba928a34a536f26b64e1d5430778f053e0fd0e00d75363c`.
It preserves both sources, selections, treatment map, run order and all eight
completed observations. Versioned internal arithmetic
`.cartulary/remediation03/performance-analysis-v3.py` has SHA-256
`d991bb541f5272b89891538df591bdc7c0da3c5a178716f5492f439938ab1c37`;
v2 and its historical bindings are unchanged. Arithmetic checks cover strictly
below/exactly 5%, inclusive 20%, the full-protocol route and first-two-only
rejection. These are internal numerical checks, not product verification.
S-001's accepted first-pair improvements (support 16.26%/17.01%; functional
34.75%/36.10%) do not trigger rejection; E-049's decision is unchanged.

All first-pair commands use E-051's exact selections and profile. Native exits,
public `CHECK=all` audits and performance exports passed; every cleanup result
completed, scans passed and accounting scopes settled empty.

| Run ID under `.cartulary/test-results/` | Native / exit | Graph elapsed ms | Owning-command elapsed ms | Cleanup results |
| --- | --- | --- | --- | --- |
| `r03-s002-exp01-focus-p01-b` | 96508/0 | 17,098 | 21,346.870722 | 14 |
| `r03-s002-exp01-containing-p01-b` | 15319/0 | 358,070 | 363,595.562753 | 142 |
| `r03-s002-exp01-focus-p01-c` | 94475/0 | 18,635 | 22,962.628949 | 14 |
| `r03-s002-exp01-containing-p01-c` | 31153/0 | 270,259 | 275,663.414994 | 142 |

Focused elapsed regressed 7.57%; containing elapsed improved 24.18%. The slower
containing baseline remains a valid retained sample; no cause or exclusion is
invented. Neither workload has two complete pairs yet, so neither first-pair
observation closes its decision. Snapshot-verified candidate restoration used
`r03-s002-exp01-candidate-generate-01` (8104/0, 10,230 ms), then
`r03-s002-exp01-candidate-generated-01` (78183/0, 16,074 ms) and its passing
public audit. Source matches E-051's candidate exactly. Active treatment is
candidate; pair 2 runs candidate focus/containing, then baseline focus/containing.

Amendment lint passed as `r03-s002-amendment-lint-01` (95630/0, 9,230 ms).
The preceding attempt reused `r03-protocol-amendment-lint-01` and was rejected
at admission with exit 2 because its root was nonempty; it did not run lint,
overwrite historical evidence or create a performance attempt.

**E-054 — S-002 optimization rejected; mandatory correctness retained.**
The second focused pair completed candidate-first. Candidate
`r03-s002-exp01-focus-p02-c` passed (78448/0, graph 18,780 ms; owning elapsed
23,121.636976 ms; fourteen cleanup results). Baseline
`r03-s002-exp01-focus-p02-b` passed (51487/0, graph 18,996 ms; owning elapsed
23,231.827706 ms; fourteen cleanup results). All audits/exports exited 0,
scans passed and scopes settled empty. The first two focused improvements are
**-7.569063625493389% and 0.4743093457581964%**. Both are strictly below 5%:
E-053 requires rejection of this optimization, without further sampling or a
claim about confidence bounds. Internal decision is retained in
`.cartulary/remediation03/s002-exp01-focus-decision.json`.

The containing candidate pair-2 member also passed as
`r03-s002-exp01-containing-p02-c` (58434/0, graph 369,194 ms; owning elapsed
375,506.058599 ms; 142 cleanup results), with audit/export 0, passed scans and
empty settlement. Its matching baseline and all later performance members are
omitted because the candidate already failed the required focused workload.
The receipt register preserves eleven completed runs and nine individually
identified omissions. The containing workload has only one complete pair;
its favorable first pair does not establish its own acceptance or override the
focused rejection. No same-source failure, omitted attempt or unfavorable
sample is discarded.

Baseline restoration was snapshot-verified, regenerated by
`r03-s002-exp01-baseline-generate-02` (60344/0, 10,160 ms), and qualified by
`r03-s002-exp01-baseline-generated-02` (10123/0, 15,639 ms; audit 0). The original
standalone Location test and its routing are restored. The optimization is not
accepted and no fixture reduction is claimed for the final source.

Mandatory Core 01 create/replay Location assurance remains an authorized
correctness repair. Its two exact assertions and independently selectable row
are retained without deleting any native test: the original row now has four
siblings including the standalone test, and the exact-symbol row has the
idempotency test. Their union equals the original five-symbol selection, with
unchanged owner, fixture, resource and verification metadata. No additional
request or setup is added to the survivor. This is a separate correctness-only
disposition, not a renamed performance candidate; current-source standalone,
combined, catalog and generated checks remain required before WF-02 closure.

**E-055 — WF-02 complete: correctness accepted, optimization rejected.**
Final executable source is
`sha256:91e281593e01e8cd49b501b1e9fdd2c7d324eb6a7cd34bfa308ad87cdfcf1f9f`.
Core 01 REQ-01-164/166/167 → C-007/R03-007/008 → the two exact Location
assertions in `internal/modules/incidents/http_conformance_test.go` and exact
row in `tools/test_families/module.incidents.json` → V-004's independent and
combined selections. Preservation record
`.cartulary/remediation03/s002-correctness-preservation.json` proves the full
HTTP file equals baseline plus only those assertion blocks, and the catalog
union retains all five original native symbols exactly once. All other rows
and owner/verification/fixture/resource metadata are unchanged. Compatible
support rows remain in ordinary shared Go invocation grouping.

All commands used `CGO_ENABLED=0` and fresh explicit IDs below. Test slices
used cache off and basic diagnostics. Format passed (8512/0, 7,781 ms), then
serial generation passed (13631/0, 10,060 ms), under respectively
`r03-s002-correctness-format-01` and `r03-s002-correctness-generate-01`.

| Run ID under `.cartulary/test-results/` | Public command / selection | Native / exit | Graph ms | Completed cleanup |
| --- | --- | --- | --- | --- |
| `r03-s002-correctness-generated-01` | `make source-check CHECK=generated` | 56915/0 | 15,653 | 14 |
| `r03-s002-correctness-catalog-01` | `make source-check CHECK=catalog` | 53225/0 | 1,716 | 11 |
| `r03-s002-correctness-survivor-01` | `make test-slice OWNER=module.incidents ROWS=module.incidents.support_integration.authenticated_incident_creation_and_normalized_r_e2cd9c99b5` | 13033/0 | 17,747 | 14 |
| `r03-s002-correctness-siblings-01` | `make test-slice OWNER=module.incidents ROWS=module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` | 92525/0 | 18,693 | 14 |
| `r03-s002-correctness-durable-01` | `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6` | 88570/0 | 15,474 | 14 |
| `r03-s002-correctness-combined-01` | Same owner, comma-separated union of the preceding three exact rows | 69755/0 | 22,665 | 17 |
| `r03-s002-correctness-routing-01` | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | 27430/0 | 52,864 | 11 |

Each listed graph also passed independent `make harness-observability-check
RESULTS_DIR=<exact root> CHECK=all`. Retained-secret scans passed, all cleanup
results completed and final resource-holder sets are empty. Native exits are
separate from audit exits. Exact inputs/source/cleanup receipts are retained in
`.cartulary/remediation03/s002-correctness-receipts.json`.

Assurance is accepted for this correctness-only source. The efficiency
investigation is complete with a supported rejection and rollback under E-054;
no speedup, fixture reduction or statistical guardrail pass is claimed. The
additional standalone fixture remains a measured, explicitly accepted cost of
this disposition. No production API/storage/authentication behavior changed.
S-007 still owns final integration and explicit browser checks. WF-03 may begin
only after this completion update and its boundary Markdown lint; its inputs
are accepted S-001 plus S-002's final correctness-only source, with capacities,
workers, deadlines and fixture lifetimes unchanged. Boundary lint passed as
`r03-s002-boundary-lint-01` (77875/0, 9,080 ms), recorded before WF-03 entry.

**E-056 — WF-03 entry and current-source characterization.** Entered after
E-055's tracker update and successful lint. Public task guides for
`harness.command_surface` and `harness.evidence_accounting` passed (native
65324/0 and 58145/0), retaining exact row selections under
`.cartulary/remediation03/s003-*-guide.json`.

REQ-803's resident/active distinction, drain entitlement and exact retirement
were traced through `harnessadmission/resident_linux.go`, the weighted admission
cases, `harnessfixtures/residency_linux.go` and service-owner settlement.
`make test-slice OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.observation_admission` passed as
`r03-s003-admission-01` (54349/0, 734 ms). The scheduler contract and scheduler
smoke rows passed together under `OWNER=harness.command_surface` as
`r03-s003-scheduler-01` (86561/0, 13,553 ms). Both used CGO disabled, cache off,
basic diagnostics, and passed independent `CHECK=all` audits. No admission
correctness failure is established by this bounded inspection and validation.

Go and Vitest still sum one authored evidence estimate per selected catalog row;
Go startup/cleanup are already charged once per invocation. Public current
focus planning (native 1908/0) confirms the five-symbol incident union now
estimates 15,250 ms versus baseline 10,250 ms with one invocation in each.
The native work is unchanged by the reporting split. This confirms F-003's
estimate-granularity limitation; it does not prove added physical work or a
measured tuning benefit. No policy or implementation has been changed.

Exploratory current-source probe registration is
`.cartulary/remediation03/s003-probe-registration.json`: exact run
`r03-s003-containing-probe-01`, `make test-slice OWNER=module.incidents
SCOPE=services HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`, with
CGO disabled and cache off. Source is E-055's accepted source. It collects
owning-command accounting and public phase diagnostics with unchanged settings,
using only native observation until terminal exit. It is not a baseline/candidate
acceptance pair and cannot establish improvement or confidence bounds. The
already audited current-source combined incident selection supplies a focused
phase observation, with its lack of whole-invocation accounting kept explicit.

**E-057 — WF-03 complete with supported no-change calibration decision.**
The E-056 current-source containing probe passed: native 22861/0, graph
363,456 ms, owning-command elapsed 368,977.998266 ms, charged CPU
783,005,776 microseconds and peak cgroup charge 6,041,956,352 bytes. All 142
cleanup results completed, final resource holders are empty and retained-secret
scans passed. Independent public `CHECK=all` and performance export exited 0;
accounting settlement is empty. Source remains E-055's accepted source.

Public canonical phase unions are fixture acquisition 84,163 ms, runner
309,255 ms, host admission 275,266 ms, fixture release 24,963 ms and evidence
write 2,135 ms. These overlap (reported phase overlap 332,917 ms), so they MUST
NOT be summed as elapsed time or treated as disjoint opportunity. The complete
selection reports p95 ready-to-start delay 1 ms, invocation-path queue 5 ms and
cleanup union 25,529 ms. Published diagnostic activities independently record
22 browser resets, 2,890–3,160 ms each, summed 65,770 ms; this sum is not an
elapsed union and overlaps runner/fixture phases. The focused current-source
combined selection reports acquisition 7,917 ms, runner 13,537 ms, admission
101 ms, release 1,226 ms and evidence write 109 ms. These observations are
exploratory, with no performance acceptance or favorable-attempt selection.

C-006/R03-009–011 → REQ-803/805 → compiler/cohort estimates, Go compatibility
planning, resident admission and pgtest rollback ownership → E-056's current
scheduler/admission tests and this independently audited probe. Existing Go
startup/cleanup are already charged once. The five-symbol reporting split
still changes its estimate from 10,250 to 15,250 ms without changing the single
Go invocation. Vitest similarly accumulates per-row estimates. The desired
reporting-split invariance criterion is **not met by the existing estimator**;
this investigation does not label that limitation fixed. No adopted physical
accounting requirement failed the inspected tests. Rollback's mutex remains
held through transaction settlement and cannot safely be bypassed for packing.

**Disposition: NO_CHANGE to estimates and admission policy.** Distinct process
startup versus package setup/native execution, per-file setup and rollback mutex
waits are not all separately available in these retained observations. Whole-run
accounting does not manufacture that missing attribution. Reinterpreting the
old per-row constants as per-symbol constants or inventing file/setup weights
would lack the required provenance and could distort ten-minute packing.
No calibrated candidate or >5% benefit is established; no policy change is
accepted and no four-pair experiment is claimed. The conditional investigation
ends with this specific missing calibration proof, rather than a false numerical
pass. The estimate-granularity limitation remains a documented optimization
risk; physical resource ownership, capacities, workers, deadlines and rollback
serialization remain unchanged. Any future candidate needs a fresh freeze and
complete selected-work measurements before acceptance.

Evidence and scope limits are retained in
`.cartulary/remediation03/s003-investigation.json`; public CI planning passed
(native 73664/0), showing eleven current Vitest cohorts whose estimates range
6,250–384,250 ms, within the existing 600,000 ms envelope. This is planning
evidence, not measured cost. No executable, schema or authored policy changed
in WF-03, so no migration or rollback was needed. WF-04 may enter after this
tracker update and boundary lint; candidate-specific isolation proofs remain
mandatory and cannot borrow a pass from this investigation.
Boundary lint passed as `r03-s003-boundary-lint-01` (42714/0, 8,910 ms),
recorded before WF-04 entry.

**E-058 — WF-04 complete with stronger isolation retained.** Entered only
after E-057 boundary lint. The current browser task guide passed (10075/0).
The declaration screen covers all 895 rows declaring exclusive PostgreSQL or
stronger browser state, with exact IDs/selectors and per-row dispositions in
[the isolation evidence appendix](testing-harness-remediation-03-isolation-evidence.json),
SHA-256 `baebba74030c578b5ad43c34fd1b10b6e8a916c445f8a76189c71be5ba6d3f93`.
This is authoring evidence subordinate to this tracker, never runtime policy.

525 declarations are excluded: 469 already test-owned browser rows whose
PostgreSQL allocation remains exclusive, 33 measurement rows, eleven migration
rows, five self-owned/lifecycle/release declarations, and seven inspected
incident rows requiring independent committed connections. The remaining
370 declarations (88 isolated browser and 282 dedicated Go rows) are unresolved,
not proved eligible. Their stronger declarations remain. The screen is complete
as an inventory; it is explicitly not a complete transitive effect proof for
all 370 unresolved rows.

The current-source E-057 probe supplies measured boundary exposures. Browser
candidates are ordered by descending observed preceding reset duration, then
stable row ID. Highest is `module.incidents.browser.lifecycle_acknowledgement`
at 3,160 ms, followed by the workbook-home/default-preference row at 3,130 ms
and membership-audit access recovery at 3,120 ms. These are measured reset costs,
not demonstrated avoidable savings. Unknown costs remain null. Native test
elapsed time includes assertions/setup and is not substituted for per-row
avoidable PostgreSQL lifecycle cost, which is unavailable.

The highest browser candidate was traced through `incident-lifecycle.spec.ts`,
`support/incidentLifecycle.ts`, `support/incidents/fixtures.ts`,
`pages/incidentDirectory.ts`, actor/context/session fixtures and
`auditBrowserBarrier`. Its normal path has a fresh incident and actor, local
route interception/fault flags, and explicit gate release. An assertion before
release can leave the delayed request promise unresolved; candidate-specific
failure settlement and shared-order/concurrency proof are absent. Retaining the
isolated boundary preserves quarantine/settlement authority while that proof
is missing. This is an eligibility uncertainty, not a newly observed runtime
failure or a claim that the current isolated test is incorrect.

The inspected PostgreSQL candidates fail transaction confinement: the mutation
admission test opens three pgx connections and exercises independent locks;
cross-admin deletion uses a pool and concurrent committed application mutations;
HTTP/create/replay and route-inventory fixtures use `newRouteFixture` or
`StartDefaultServer`, an independently admitted pool, a separate SQL handle,
real HTTP commits and durable reads. Replacing those with a single rollback
transaction would change the assurance subject. Background jobs, global
listings/admin state, clock/fault/session effects and nested cleanup remain
explicit proof dimensions for the untraced candidates; none is silently waived.

C-002/C-003, R03-012 and REQ-804/806 → current fixture declarations and
pgtest/broker settlement → V-003/V-007 baseline checks: with CGO disabled,
cache off and basic diagnostics, `make test-slice OWNER=harness.browser
ROWS=harness.browser.unit.fixture_broker_smoke,harness.browser.unit.native_fixtures,harness.browser.integration.postgres_transaction_isolation`
passed as `r03-s004-boundaries-01` (13303/0, 7,794 ms, seven units and eighteen
completed cleanup results). The exact row
`harness.browser.unit.pgtest_helper_contract` passed as
`r03-s004-rollback-settlement-01` (47820/0, 2,061 ms, one unit and ten completed
cleanup results), including rollback/close failure poisoning and bounded
cleanup after cancellation. Both public `CHECK=all` audits passed; scans passed
and final resource-holder sets are empty. These baseline checks do not certify
an unimplemented relaxation.

**Disposition: NO_CHANGE.** No candidate satisfies the complete eligibility
proof, so no conversion pilot or child implementation workstream is admitted.
There is no converted source, performance experiment, >5% benefit claim or
migration/rollback action. The stronger boundaries retain some setup cost but
avoid unproved reuse and accidental assurance changes. This completes the
conditional investigation, not the outstanding proofs for potential future
conversions. WF-05 may enter after this tracker update and boundary lint, using
the unchanged E-055 executable source and its current E-057 workload evidence.
Boundary lint passed as `r03-s004-boundary-lint-01` (15569/0, 9,350 ms),
recorded before WF-05 entry.

**E-059 — WF-05 complete: dedicated PostgreSQL lifetimes retained.** Entered
after E-058 boundary lint, with executable source unchanged from E-055. The
current-source E-057 probe is therefore the resulting workload after all earlier
dispositions. Public `make explain-run RESULTS_DIR=.cartulary/test-results/r03-s003-containing-probe-01
DETAIL=fixtures` exited 0 and reports all five Go leases released. Their held
durations are 2,778 ms for the shared rollback unit and 4,879/4,859/45,914/34,179 ms
for the four exclusive units. Lease duration is not service startup duration.

Published `cartulary.test_services.scope.v2` records identify the four exact
non-browser exclusive generations, each with one successful start and completed
physical settlement:

| Go unit prefix / subject | Startup ms | Per-test clones | Proposed lifetime eligibility |
| --- | --- | --- | --- |
| `go:12a7fe25f0e43199` mutation admission | 13,146 | 1 | Excluded: lifecycle locking/deadline assertion with independent connections and explicit 100 ms / 2 s deadlines |
| `go:758713088a770679` cross-admin mutation | 13,140 | 1 | Plausible ordinary borrower; complete future borrow/effect contract unproved |
| `go:52a3ddf1d8e07e7a` support HTTP inventory | 7,636 | 45 | Plausible ordinary borrower; committed per-test database isolation must remain |
| `go:9325be585dd8cbe6` integration HTTP inventory | 7,735 | 29 | Plausible ordinary borrower; committed per-test database isolation must remain |

The three nonexcluded starts sum to 28,511 ms, 7.727% of the observed
368,977.998266 ms owning-command duration, even before retaining a required first
start or accounting for overlapping starts. This is a deliberately generous
startup-cost exposure, **not** a predicted speedup or a complete upper bound:
separate template create/migrate entries have zero reported duration, which does
not establish zero template cost; sequential borrower scheduling can change the
critical path. Per-test clones/drops and durable checkpoint behavior cannot be
removed by service reuse. Matching template names alone cannot establish exact
engine/configuration/roles/extensions/provenance compatibility or settlement.

C-003/R03-013–015 → REQ-804 and its current per-fixture exclusive generation →
`harnessfixtures/services_linux.go` (`clusterKey`, `releaseCluster`, exact
`SuiteOwner.Close`), resident admission/recovery and pgtest fresh database
ownership → E-056 admission checks, E-058 broker/failure/rollback checks and
E-057's fully audited current-source workload. Browser, measurement, migration,
cluster/role mutation, checkpoint-subject, lifecycle/deadline and self-owned
service tests remain outside the proposed ordinary reuse class. The existing
shared rollback unit already has its own correct scope and is not a new reuse
opportunity. All carried validation names the same unchanged source; no new
candidate overlap, stale-token or quarantine pass is claimed.

**Disposition: NO_CHANGE / dedicated lifetime.** The investigation does not
establish a material remaining end-to-end opportunity exceeding the required
10% gate. The added eligibility, fenced-borrow, sequential scheduling, retirement
and versioned lease responsibilities are not justified by this evidence. The
conditional design is not activated: REQ-804 remains unchanged, no reusable
exclusive generation or pooling toggle is implemented, and no catalog/graph/lease
version is cut over. No candidate source needs rollback. This is a supported
bounded-workload decision, not proof that reuse can never help another workload.
A future proposal needs separately observed template cost, exact eligible ready
borrowers, a new freeze and all safety/performance gates before adoption.

Exact scope paths, hashes, startup/fixture observations and limitations are
retained in `.cartulary/remediation03/s005-lifetime-investigation.json`.
The measured setup cost remains a known tradeoff; current physical settlement,
quiet admission and checkpoint/durability semantics remain intact. WF-07 is
next after this completion update and boundary lint. It must run agent-finalize
before broader verification, then final source/schema/catalog/owner checks,
CI and explicit browser support/visual/UI-review coverage. No final completion
or publication qualification is claimed here.
Boundary lint passed as `r03-s005-boundary-lint-01` (47503/0, 15,490 ms),
recorded before WF-07 entry.

**E-060 — WF-07 entry and finalization.** Entered only after E-059 boundary
lint passed. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-finalize-01
make agent-finalize` published a pass, unchanged generated outputs, source
`sha256:91e281593e01e8cd49b501b1e9fdd2c7d324eb6a7cd34bfa308ad87cdfcf1f9f`
and passing retained-secret scan. Context recovery lost its native session
handle; outer exit is UNKNOWN and is not inferred from the summary. Public
`make test-run-status RESULTS_DIR=.cartulary/test-results/r03-s007-finalize-01
JSON=1` independently returned exit 0 and a terminal pass with
`outer_exit_not_observed`. This attempt remains retained, not erased.

After that terminal diagnosis, a fresh finalizer obtained a complete receipt:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-finalize-02 make agent-finalize`,
native session **87656**, outer **0**, graph **27762 ms**. Generated outputs
were unchanged. Both exact roots are under `.cartulary/test-results/` with the
literal run IDs above. `RESULTS_DIR` was unset: retained-run maintenance was
skipped because no compatible successful full warm-check root was selected.
This is not a claimed maintenance pass.

Next: generated/schema/catalog validation, affected owner and explicit accounting
checks, one final CI invocation, explicit browser support/visual and UI-review
contracts, then independent evidence/settlement/secret-scan checks and final
tracker lint. Final completion remains open until those receipts exist.

**E-061 — WF-07a boundary-projection correction, IN_PROGRESS.** Final
owner checks exposed one mandatory correctness failure on source `91e2815…`:
`r03-s007-command-01`, native **22562/2**, graph **82727 ms**, 9/10 units
passed. `harness.command_surface.behavior.owner_boundaries` reports
`forbidden_private_browser_import`: the generator imports the new pure
`browser-compatibility.mjs` projection without a declared semantic facade.
All 19 cleanup outcomes completed; retained-secret scan passed; independent
`CHECK=diagnostics` returned 0, which does not convert the failed assertion
into passing assurance. The browser-owner run is allowed to drain before any
executable projection changes.

Adopt the `browser_compatibility` semantic facade in the NLSpec closed helper
registry before its machine projection. Add its exact existing pure module
path with no private-consumer exception. This preserves the single shared
compatibility tuple and avoids importing service/renderer implementation into
the generator. Reconcile the three stale normative browser-v14/performance-v5
references with the already adopted v15/v6 cutovers. No behavior or schema shape
changes; earlier source-bound performance observations retain their own identity.
Validation: exact owner-boundary row, generation, finalization, final source
checks and CI. No performance rerun is required for this owner-registry/spec
correction. This child repair must close and update this tracker before broader
S-007 validation resumes.

Selection/diagnostic mistakes are retained separately: inline JSON supplied to
`target-plan INPUTS` returned 2 (requires a JSON file), then the explicit
`.cartulary/remediation03/s007-ui-review-inputs.json` selection returned 0 and
selected nine rows. An internal projection name supplied as `explain-run TARGET`
returned 2; the public `TARGET=test-slice` returned 0 and exposed the bounded
assertion. Neither unsuccessful reader dispatched or retried a test.

**E-062 — WF-07a complete.** The
NLSpec facade adoption precedes the authored `tools/harness_helper_ownership.json`
projection. The new `browser_compatibility` entry exposes only the existing pure
module; `allowed_consumers` is empty, so no private-import exemption was added.
No JavaScript, compatibility tuple, selected-row identity, runtime behavior,
scheduler capacity, deadline, schema shape or resource lifetime changed.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007a-boundary-01
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.behavior.owner_boundaries HARNESS_DIAGNOSTICS=basic`
passed, native **30810/0**, **27544 ms**, independent `CHECK=all` **0**.
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007a-generate-01 make generate`
passed, native **71672/0**, **10590 ms**. The subsequent
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007a-finalize-01 make agent-finalize`
passed, native **6199/0**, **26325 ms**. Maintenance remains skipped because
`RESULTS_DIR` was unset. Exact roots use these IDs under `.cartulary/test-results/`.

The distinct earlier browser-owner failure remains open: native **82438/2**,
`r03-s007-browser-owner-01`, **335216 ms**, 49/50 units passed; all 83 graph cleanup
outcomes completed and retained-secret scan passed. Snapshot clone 1's first drop
exceeded its unchanged operation budget while PostgreSQL reported
`CheckpointDone` / `DataFileSync`, zero blockers and requested/done 1/0.
`CHECK=diagnostics` returned 0, not assurance. This is not attributed to the facade
correction or treated as resolved by successful physical retirement. WF-07b will
investigate this mandatory current failure after the boundary update/lint.

WF-07a boundary lint `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s007a-boundary-lint-01 make lint-markdown` passed,
native **17652/0**, **15400 ms**, before WF-07b entry. The resulting executable
source is `sha256:e00113a91fe3878dc42bf7a7b5fd1a5caeae5365340a0e1d7e403e30968be59f`.

**E-063 — WF-07b snapshot cleanup diagnosis, IN_PROGRESS.** The independently
validated predecessor diagnostic binds service ordinal 6 to generation
`7699235663522619432` and scope
`_shared/test-services/a4660bd18fb545aa9ea7c766/service-scope.json`. Its first
checkpoint synced **6421 files in 21091 ms**, write phase **143 ms**, longest
single sync **99 ms**. These observations establish a checkpoint sync overrun;
they do not identify the responsible host/storage contention or prove a new
code defect. The existing drop operation cap is **20 s**; it remains unchanged,
as do WAL-backed clone creation, durability, capacities and lane counts.

Before another execution, freeze one isolated characterization:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007b-snapshot-isolated-01
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.performance_fixture_snapshot_lifecycle
HARNESS_DIAGNOSTICS=basic`, on source `e00113a…`, existing toolchain/system/storage
profile. Do not run other repository work or diagnostic readers during this
characterization; observe its native session only. Inspect its existing retained
checkpoint and cleanup projections afterward. A pass alone will not erase the
failed containing run or establish its cause. Broader S-007 checks remain pending.

WF-07b isolated characterization passed, native **75033/0**, graph **16080 ms**,
independent `CHECK=all` **0**, all **14** cleanup outcomes completed and retained
scan passed. Its first checkpoint synced **6587 files in 2438 ms**, write phase
**96 ms**, longest sync **5 ms**. This contrast establishes variability, not a
host-contention cause or implementation repair. No failure was removed, no
capacity/deadline/durability setting changed, and no optimization is claimed.
A current-source containing validation remains required. The next registered
containing validation is the already required final `make ci`, cache off/basic,
source `e00113a…`, unchanged host/toolchain/storage. It will run without other
repository execution; use native observation only through quiet measurement
intervals. Preserve any failure and diagnose it before further action. Final
S-007 acceptance cannot be inferred from the isolated pass.

**E-064 — WF-07b current assertion validation complete; causal limit retained.**
The registered containing validation `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s007-ci-01 CARTULARY_HARNESS_CACHE_MODE=off
make ci HARNESS_DIAGNOSTICS=basic` completed with native **78464/2**,
**3680717 ms**, **599/600** units passed, no skipped/cancelled units. The exact
source is `sha256:e00113a91fe3878dc42bf7a7b5fd1a5caeae5365340a0e1d7e403e30968be59f`,
graph `sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
The earlier target-plan used default diagnostics; its graph digest is not
substituted for this actual basic-diagnostics graph. Both select 2279 rows and
600 units. All **1160** graph cleanup outcomes completed; retained-secret scan
passed. Independent `CHECK=diagnostics` returned **0**, `full_validity=0`:
execution/resource observations were truncated (4773 resource omissions), and
checkpoint publication was unavailable. No whole-run performance conclusion is
made from this CI invocation. Nine bounded public status observations returned
active or terminal states; none reported a protected measurement pause. Read-only
source inspection followed the type-check failure; no competing test workload
or source mutation was introduced during CI.

The snapshot lifecycle row passed in this containing run: native row exit **0**,
**22870 ms** test duration, **27181 ms** row wall duration. Together with the
isolated pass, this completes current execution of that required assertion on
source `e00113a…`; it does **not** establish a causal repair of the predecessor
21.091-second checkpoint sync overrun. Retain that variability as an unresolved
storage/environment risk, with its failed source/root unchanged. No deadline,
capacity, durability or fixture contract was relaxed and no retry loop added.

CI itself **FAILED** frontend type checking. Its published log reports TS2305
for three non-exported internal schema names imported by
`apps/web/e2e/fixture-state.spec.ts`, plus the resulting implicit-any error.
The failed unit is `operation:harness.frontend_typecheck`, exit **2**, class
`config/configuration_error`. This mandatory correctness failure remains open;
S-007 is not complete. Next is WF-07c, a type-only browser test repair using the
existing public operation request/response types. First record this boundary and
lint; then change the annotations, run the public type check and affected browser
selection, finalize the repaired source, and complete final CI/explicit coverage.

WF-07b boundary lint passed before the next repair:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007b-boundary-lint-01
make lint-markdown`, native **54985/0**, **8930 ms**.

**E-065 — WF-07c type-only fixture repair, IN_PROGRESS.** Use public
`HTTPOperationRequest` / `HTTPOperationResponse` annotations keyed by the already
adopted operation IDs. Do not widen the protocol entrypoint, import private
generated schemas, weaken typing or change requests/assertions. The repair
belongs to tests and validation documentation; its emitted browser behavior is
unchanged. Add `make lint CHECK=types` to the browser-source validation route
before future browser comparisons; Playwright execution alone did not detect
these invalid type imports. Validate the compiler, four exact fixture-state
rows, generated finalization and remaining final suites. The predecessor CI
failure remains retained and does not become a pass after this repair.

**E-066 — WF-07c implementation and focused validation complete.** The only
executable-source change from `e00113a…` is type-only imports/annotations in
`apps/web/e2e/fixture-state.spec.ts`: public operation request/response types
replace unavailable internal schema exports. No request, assertion, selector,
setup, cleanup or emitted browser behavior changed. No protocol export was added.
The current source is
`sha256:69cd6bc2cad6168b274ac905ac80190a19641aa201d79309205294b8faecd245`.

All commands below use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<listed-id>`;
the two test/check commands additionally use `CARTULARY_HARNESS_CACHE_MODE=off`
and `HARNESS_DIAGNOSTICS=basic`. Roots are `.cartulary/test-results/<listed-id>`.

| Run ID | Public command / exact selection | Native session / exit | Result |
| --- | --- | --- | --- |
| `r03-s007c-format-01` | `make format` | 95763 / 0 | 7801 ms |
| `r03-s007c-types-01` | `make lint CHECK=types` | 78129 / 0 | 22530 ms; independent `CHECK=all` 0 |
| `r03-s007c-fixture-state-01` | `make test-slice OWNER=harness.browser` with the four exact fixture-state rows below | 79792 / 0 | 83183 ms; independent `CHECK=all` 0 |
| `r03-s007c-finalize-01` | `make agent-finalize` | 6239 / 0 | 25511 ms; generated outputs unchanged; maintenance skipped because `RESULTS_DIR` unset |

The exact comma-joined `ROWS` selection is:
`harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_a_577b9aa596`,
`harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_b_e9fdc93f9d`,
`harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_a_a194a3225b`,
`harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_b_ab488bf040`.
Current TypeScript assurance passes; the failed predecessor CI is preserved.
Next entry requires this tracker update and successful boundary lint, followed
by current-source projection checks, explicit support/visual/UI-review and
accounting coverage, and final CI. No performance remeasurement is claimed or
required for the type-only correction.

WF-07c boundary lint passed before broader S-007 work resumed:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007c-boundary-lint-01
make lint-markdown`, native **38110/0**, **9460 ms**. The type, fixture-state
and finalizer roots have respectively 11, 23 and 10 completed cleanup outcomes,
no failed/blocked cleanup, and passing retained-secret scans.

**E-067 — S-007 current-source explicit coverage and CI registration.**
Source remains
`sha256:69cd6bc2cad6168b274ac905ac80190a19641aa201d79309205294b8faecd245`.
The following checks ran after E-066 finalization, with
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<listed-id>
CARTULARY_HARNESS_CACHE_MODE=off` and `HARNESS_DIAGNOSTICS=basic`.
Every listed completed run has native exit 0 and independent
`make harness-observability-check RESULTS_DIR=.cartulary/test-results/<listed-id>
CHECK=all` exit 0. These are correctness checks, not performance samples.

| Run ID | Public command / selection | Native session / exit | Duration |
| --- | --- | --- | --- |
| `r03-s007-generated-03` | `make source-check CHECK=generated` | 6379 / 0 | 16805 ms |
| `r03-s007-schema-03` | `make source-check CHECK=schema` | 32801 / 0 | 6808 ms |
| `r03-s007-catalog-03` | `make source-check CHECK=catalog` | 62894 / 0 | 1839 ms |
| `r03-s007-accounting-02` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.invocation_accounting_preserves_complete_kernel_74dff0b392` | 16075 / 0 | 15107 ms |
| `r03-s007-support-01` | `make browser-e2e SUITE=support` | 69628 / 0 | 118438 ms |
| `r03-s007-visual-01` | `make browser-e2e SUITE=visual` | 21360 / 0 | 635580 ms |

Visual coverage is explicit: all 37 work units passed, 74 cleanup outcomes
completed, no failed/blocked cleanup, retained-secret audit passed. UI-review
contracts use the separately resolved nine-row selection from
`make target-plan TARGET=harness-contract
INPUTS=.cartulary/remediation03/s007-ui-review-inputs.json JSON=1`, whose input
file is exactly `{"SUITE":"ui-review"}`. The completed invocation is
`r03-s007-ui-review-01`, native **67519/0**, with the same environment
above and `make harness-contract SUITE=ui-review HARNESS_DIAGNOSTICS=basic`.
All ten work units passed in **119873 ms**; all 19 cleanup outcomes completed,
retained-secret audit passed, and independent `CHECK=all` exited 0. The public
summary confirms the source above and graph
`sha256:c24c132594b289f325e5787d2e45b928f397da97c9648a61d2f61834830a5d31`.

Final CI registration: after that invocation settles and passes its evidence
audit, run `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-ci-02
CARTULARY_HARNESS_CACHE_MODE=normal make ci HARNESS_DIAGNOSTICS=basic` from the
repository root, with the source above and default capacities, lanes and workers.
Its root will be `.cartulary/test-results/r03-s007-ci-02`. Normal is the public
default cache policy; report any reuse separately from fresh execution. This is
final correctness validation, not a performance comparison. Preserve E-064's
failed CI, keep protected measurement intervals quiet, and perform no competing
workload or source mutation during the run. No final CI success is yet claimed.

**E-068 — WF-07d required cleanup repair, IN_PROGRESS.** The user explicitly
requested “end the run and make the repairs.” CI `r03-s007-ci-02` was interrupted
through native session **57778** and exited **130**, after **297054 ms**. Its
terminal result is cancelled: 47 passed, three failed and 550 cancelled work
units. All 116 published cleanup outcomes completed; `CHECK=diagnostics` exited
0 with `full_validity=0`, partial execution diagnostics, complete sampled resource
diagnostics and one checkpoint service unavailable because of measurement quiet.
Normal cache reported zero hits, eight misses and 42 bypasses. This run is not a
pass or a performance sample. The first failed units are the Timeline and Jobs
Go batches; a measurement browser unit also failed before native launch.

Both Go failures began at their first owned database deletion, before subsequent
fixture preparation was correctly rejected by quarantine. Drop admission took
6/5 ms; the 20-second operation expired while samples showed `CheckpointDone`
and checkpointer `DataFileSync`, with zero blockers. Typed checkpoint completion
records bind Jobs generation `7699260582585307176` to service ordinal 6:
1397 files, 20065 ms sync, 33 ms write; Timeline generation
`7699260582296317992` to ordinal 7: 1398 files, 20424 ms sync, 54 ms write.
These observations establish slow checkpoint synchronization, not a Windows-host
storage cause. Exact public logs, performance export and native receipts are
retained under `.cartulary/remediation03/s007-ci-02-*`.

The adopted owner requires one forced drop, unchanged cleanup budgets, durable
WAL-backed clones and quarantine on uncertainty. PostgreSQL's DROP implementation
forgets the dropped database's pending sync requests before requesting the
cluster checkpoint. Both current template producers seal the migrated template
but publish without settling their initialization checkpoint work. WF-07d moves
that work to the unpublished physical owner's template-sealing boundary: one
checkpoint after sealing and before admitting borrowers, shared by native and
broker producers. It adds no per-clone checkpoint, retry, capacity increase,
timeout increase, disabled durability or service lifetime reuse. The corresponding
REQ-804 amendment must precede implementation. Provisioning identity must change
so earlier templates cannot satisfy the amended preparation contract.

Validation must prove checkpoint completion before template publication, failure
and cancellation withholding publication, no checkpoint added to clone creation,
unchanged durability/cleanup/isolation assertions, and the affected Jobs,
Timeline and snapshot selections. Current-source finalization and final CI remain
required after the repair; earlier failures are retained. This is a correctness
repair, with startup-cost impact to be recorded, not accepted performance tuning.

**E-069 — WF-07d template and build-identity validation.** Formatting passed as
`r03-s007d-format-01`, native **52756/0**, **8116 ms**. The first expanded template
check (`r03-s007d-template-01`, **37644/2**, **22724 ms**) exposed an incorrect new
assertion: `ownsDatabase` deliberately retains historical exact ownership for
idempotent cleanup. The assertion now tests physical absence and withheld
publication rather than deletion of ownership history. Its 12 cleanup outcomes
completed; `CHECK=diagnostics` exited 0. The repaired selection
`r03-s007d-template-02`, native **66187/0**, **20097 ms**, passed all assertions,
12 cleanup outcomes and independent `CHECK=all`. Source was
`sha256:a304393d9281196ac88579e1905053cfdf6b7e96afbc8f27a64cac49b7e26368`.
Both runs used `make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.postgres_template_preparation
HARNESS_DIAGNOSTICS=basic` with CGO disabled and execution reuse off.

The first affected Jobs selection (`r03-s007d-jobs-01`, **71515/2**, **8930 ms**)
failed before native test launch with `artifact/artifact_error`. Its 12 cleanup
outcomes completed and diagnostic audit exited 0. Inspection found that
`HARNESS_RUNTIME_BUILD_INPUTS` omitted `pgtest` and its other transitive sources,
while testservices rebuilt against the changed provisioning digest. The native
broker therefore retained its earlier compiled capability identity. The repair
uses the existing conservative executable roots `tools/harnessruntime internal
db/migrations`, matching the test-service build policy; the normal cache path
must now rebuild and validate the new identity. No force-rebuild workaround or
weakened capability comparison is introduced. The NLSpec cache clause was
clarified before this Make projection changed.

Selection-reader mistakes are retained separately: a public explanation with
internal `TARGET=browser-e2e-measurement` exited 2; combined `ROWS` without
OWNER/SUITE and then with SUITE each exited 2. No test dispatched. Public exact
row selections require one OWNER, so affected owner selections are executed
separately; final CI supplies their actual combined context. These errors confer
no assurance or cleanup obligation.

The checkpoint placement rationale is supported by PostgreSQL's
[DROP implementation](https://github.com/postgres/postgres/blob/REL_18_STABLE/src/backend/commands/dbcommands.c),
which forgets the target's pending sync requests before requesting a checkpoint.
The repair does not claim a measured physical host cause or eliminate all future
storage stalls. Initialization sync is now explicitly owned by unpublished
provisioning; existing durability, clone publication and cleanup limits remain.
WF-07d remains IN_PROGRESS until affected checks and finalization complete.

**E-070 — WF-07d repair and focused validation complete.** Current source is
`sha256:70194640bc6471900fe506f37dd3dbe91cbb52eb0ace3085aaba9a4f93bbbfc3`.
The native and broker template paths now share `pgtest.SealTemplateDatabase`:
seal once, finish the initialization checkpoint, then publish. The common
provisioning digest includes those SQL steps. Native preparation uses the same
existing two-minute bound as broker preparation; shorter caller deadlines still
clamp it. A seal failure joins its exact cleanup error instead of discarding it.
The incomplete native build-cache input list was replaced with complete
executable internal-source and migration roots. Earlier template identities
are rejected; no compatibility reader or force-rebuild requirement was added.

All checks below use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<id>` from the repository
root. Test commands additionally use `CARTULARY_HARNESS_CACHE_MODE=off` and
`HARNESS_DIAGNOSTICS=basic`; roots are `.cartulary/test-results/<id>`. Each test
has native exit 0 and independent `CHECK=all` exit 0, with passing cleanup and
retained-secret scan. Earlier failed attempts in E-068/E-069 remain retained.

| Run ID | Exact public selection | Native session / exit | Duration |
| --- | --- | --- | --- |
| `r03-s007d-jobs-02` | `make test-slice OWNER=platform.jobs ROWS=<J>` | 72658 / 0 | 13257 ms |
| `r03-s007d-timeline-01` | `make test-slice OWNER=module.timeline ROWS=<T>` | 69407 / 0 | 48741 ms |
| `r03-s007d-extensions-01` | `make test-slice OWNER=module.extensions ROWS=module.extensions.integration.jobs_runner_dispatch_and_recovery_preserve_durab_f126232e74` | 10044 / 0 | 13251 ms |
| `r03-s007d-lifecycle-01` | `make test-slice OWNER=harness.browser ROWS=<L>` | 42542 / 0 | 23420 ms |
| `r03-s007d-command-01` | `make test-slice OWNER=harness.command_surface` | 83286 / 0 | 69876 ms |
| `r03-s007d-finalize-01` | `make agent-finalize` | 53268 / 0 | 26412 ms |

Selections J, T and L are the comma-joined exact IDs below; no other rows are
implied. Their public manifests retain the resolved native selectors.

- J: `platform.jobs.integration.claim_recovery_and_publication`,
  `platform.jobs.integration.dedicated_database_assertions_for_expiry_and_com_87bcee82e8`,
  `platform.jobs.integration.dedicated_database_assertions_for_lifecycle_and_604c8d7b1f`,
  `platform.jobs.integration.runner_failure_security`,
  `platform.jobs.integration.transition_progress_contract`.
- T: `module.timeline.integration.create_patch_mark_reviewed_supersede_query_and_w_90276912d7`,
  `module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6`,
  `module.timeline.integration.real_http_create_patch_review_supersede_replay_m_37e5c78ca1`,
  `module.timeline.integration.rough_and_uncertain_timeline_capture_preserves_n_a868f4aa14`,
  `module.timeline.integration.same_field_timeline_patch_conflicts_are_transpor_808ccc2585`,
  `module.timeline.integration.the_canonical_incident_web_socket_route_proves_h_aef38ee6e5`,
  `module.timeline.integration.the_query_route_reads_projection_rows_instead_of_2edde6eb1c`,
  `module.timeline.integration.timeline_create_and_patch_routes_persist_manual_ac8c9abb5b`,
  `module.timeline.integration.timeline_patch_and_query_routes_auto_resolve_onl_8ab2448829`.
- L: `harness.browser.integration.postgres_template_preparation`,
  `harness.browser.integration.performance_fixture_snapshot_lifecycle`,
  `harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277`,
  `harness.browser.integration.postgre_sql_cleanup_rejects_invalid_input_before_3f68344cda`,
  `harness.browser.integration.postgres_single_drop`,
  `harness.browser.integration.service_free_contract_assertions_for_postgres_cl_83b8534255`,
  `harness.browser.unit.testservices_lifecycle_contract`.

The template test proves checkpoint completion through the initialization WAL
position before publication, withheld templates on failure/cancellation, settled
unpublished databases, no added clone checkpoint, durable WAL progress, restricted
clone-role operation and isolation. Existing forced-drop, deadline-clamping,
failed-settlement and service lifecycle tests pass. Jobs' initialization checkpoint
synchronized 1397 files in 2789 ms before borrowing; its first borrower drop took
45 ms within the unchanged 19999 ms effective budget, and the next checkpoint
synchronized nine files in seven ms. This supports placement of initialization
sync at its owner boundary; it is not a matched performance result or a universal
future storage guarantee. The earlier snapshot overrun's physical cause remains
unproved; its current lifecycle assertions pass.

Finalization reports generated outputs unchanged, zero updated files and ten
completed cleanup outcomes. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset; no compatible successful full warm-check root is available.
A diagnostic `DETAIL=logs` call omitted TARGET and failed usage before the corrected
`TARGET=agent-finalize` read succeeded; no additional execution occurred.
After boundary lint, S-007 resumes with current-source projection checks, explicit
accounting/support/visual/UI-review coverage and final CI. None is inferred from
the earlier `69cd6bc…` source's passes. No performance acceptance is claimed for
this required correctness repair; the S-001 comparisons retain their actual sources.

WF-07d boundary lint passed before final validation resumed: `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s007d-boundary-lint-01 make lint-markdown`, native
**41400/0**, **10070 ms**.

**E-071 — S-007 validation resumed after the repair.** Source remains
`sha256:70194640bc6471900fe506f37dd3dbe91cbb52eb0ace3085aaba9a4f93bbbfc3`.
Following E-070 finalization and boundary lint, all seven completed checks below
passed with native exit 0 and independent `CHECK=all` exit 0. Environment remains
CGO disabled, execution reuse off and basic diagnostics; exact roots are
`.cartulary/test-results/<id>`.

| Run ID | Public command / selection | Native session / exit | Duration |
| --- | --- | --- | --- |
| `r03-s007-generated-04` | `make source-check CHECK=generated` | 77922 / 0 | 28218 ms |
| `r03-s007-schema-04` | `make source-check CHECK=schema` | 35130 / 0 | 7192 ms |
| `r03-s007-catalog-05` | `make source-check CHECK=catalog` | 70439 / 0 | 1726 ms |
| `r03-s007-accounting-03` | E-067 exact accounting-row command | 46658 / 0 | 754 ms |
| `r03-s007-support-02` | `make browser-e2e SUITE=support` | 13047 / 0 | 120473 ms |
| `r03-s007-visual-02` | `make browser-e2e SUITE=visual` | 76668 / 0 | 611412 ms |
| `r03-s007-ui-review-02` | `make harness-contract SUITE=ui-review` | 21287 / 0 | 108534 ms |

An attempted `make test-catalog-check` under intended ID
`r03-s007-catalog-04` exited 2 before graph dispatch because that legacy internal
name is not a public target. Inspection of the original catalog-03 manifest
confirmed `source-check`; E-067's command transcription was corrected. The fresh
catalog-05 invocation above uses the public command and supplies current evidence.

Visual completed all 37 units with 74 completed cleanup records; UI-review
completed all 10 units (nine selected rows) with 19 completed cleanup records.
Both retained-secret scans passed. Each used `CGO_ENABLED=0`, its table run ID,
`CARTULARY_HARNESS_CACHE_MODE=off` and `HARNESS_DIAGNOSTICS=basic`; no reused
execution results or updated goldens. Independent public `CHECK=all` audits
passed with complete execution and sampled-resource diagnostics. These are
correctness receipts, not whole-invocation performance measurements.

With these exact runs settled and audited, the registered final CI is
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-ci-03
CARTULARY_HARNESS_CACHE_MODE=normal make ci HARNESS_DIAGNOSTICS=basic`, root
`.cartulary/test-results/r03-s007-ci-03`, with the source above and unchanged
default capacities, lanes and workers. Normal cache use must be reported
separately from fresh execution. No competing workload or source mutation is
permitted during CI; protected measurement intervals retain their quiet boundary.
This is correctness validation and makes no new performance comparison claim.
E-064's failed CI and E-068's user-cancelled CI remain historical failures.
Final CI subsequently passed under E-072; E-073 records completion lint and the
final tracker update.

**E-072 — S-007 final current-source validation and handoff.** The E-071 registered
command `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-ci-03
CARTULARY_HARNESS_CACHE_MODE=normal make ci HARNESS_DIAGNOSTICS=basic`
completed in native session **90788**, outer exit **0**, reported duration
**3683519 ms**. Exact root:
`.cartulary/test-results/r03-s007-ci-03`. Source remains
`sha256:70194640bc6471900fe506f37dd3dbe91cbb52eb0ace3085aaba9a4f93bbbfc3`,
graph `sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
No source mutation or competing test workload ran during CI. At the published
measurement pause, structured status polling stopped; only the native session
was observed until exit.

All **600/600 work units passed**, with zero failed, skipped or cancelled units.
Normal cache policy produced **zero hits, 211 misses and 389 bypasses**; no
execution-result reuse contributed to this pass. All **1154 cleanup records
completed**, with no failed or unresolved cleanup outcome. The retained-secret
scan passed across 9614 files. The independent public command
`make harness-observability-check
RESULTS_DIR=.cartulary/test-results/r03-s007-ci-03 CHECK=all`
completed in native session **84237**, exit **0**, `full_validity=1`.

The exact formerly failing combined Timeline and Jobs/Extensions units
`go:30687414977513c8:91228ad65aad` and
`go:7fb1be85f8b62e82:45f64ad680be` passed with native unit exit 0 and no missing
outputs. The previously unlaunched measurement unit
`browser_group:measurement:cohort-0e1822ccecc9fde9a55ea660f38a201915e626efe5e13a25f911382f2fa22840`
also passed with exit 0. The snapshot lifecycle row
`harness.browser.integration.performance_fixture_snapshot_lifecycle` passed
with exit 0. These results establish current assertions and integration; they
do not retroactively diagnose the historical storage/measurement failures.

The audit explicitly reports truncated execution diagnostics (omission count
unavailable) and truncated resource diagnostics (2877 omissions); invocation
accounting and aggregate PostgreSQL checkpoint observations are unavailable.
The envelope and terminal receipt are available. These limits do not invalidate
the passing correctness/cleanup evidence, but prohibit a whole-run CPU, memory,
checkpoint or performance comparison claim for this CI. E-049's separately
accepted performance results retain their actual sources, workload scope and
omitted-confidence-bound qualifications.

E-071's current-source generated/schema/catalog/accounting/support runs also
have passing retained-secret scans and respectively 14/12/11/10/27 completed
cleanup records. Visual and UI-review add their separately audited 74/19 completed
records; neither explicit-only selection is inferred from CI. E-070 finalization
preceded this broader verification and left generated artifacts unchanged.
Retained-run maintenance was **skipped because RESULTS_DIR was unset**: no
compatible successful full warm `make check` root was available. This CI root
is not substituted for that prerequisite. No golden was refreshed.

Handoff disposition: S-001, S-006, S-008/S-008a and S-002's assertion/routing
correction are accepted; S-002's fixture removal is rejected and rolled back;
further replay consolidation is dropped. S-003/S-004/S-005 are completed
NO_CHANGE investigations, not implemented optimization. WF-07a–d corrections
are complete and current CI passes. No candidate-owned resource or active
native execution session remains unresolved from these final runs. The
pre-existing tracker edits and historical failure/cancellation receipts remain.
HEAD/branch are unchanged from E-018; no files are staged. Authored changes and
Make-generated projections remain in the working tree for review.

Migration/handoff boundaries: current browser readers reject v14; current
provisioning identities include template sealing/checkpoint semantics. Any later
rollback must drain owned invocations and restore the adopted owner, authored
inputs, generated projections and implementation together. No dual reader,
owner fallback, tuning override, isolation conversion or PostgreSQL lifetime
pool was retained. Remaining risks are explicitly bounded: incomplete future
isolation effect proofs, row-count estimate granularity, unproved lifetime-reuse
benefit, host/build-specific accounting capability, and historical physical
causes that current success cannot resolve. Future optimization requires a new
source-bound proposal and acceptance experiment. This effort has no pending
implementation slice; E-073 supplies the completion lint receipt and final update.

**E-073 — S-007 completion receipt.** The completed verification/handoff update
passed `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-final-tracker-lint-01
make lint-markdown`: native session **84868**, exit **0**, **9400 ms**. Its
retained summary is
`.cartulary/test-results/r03-s007-final-tracker-lint-01/adhoc/lint-markdown/tool-run-summary.json`.
The tracker now marks T-010/S-007 and the overall effort complete, reconciles
findings/traceability/independent outcomes, and preserves all historical receipts.
The final receipt/status update is followed by the closing command
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007-final-tracker-lint-02
make lint-markdown`; its exact retained root is
`.cartulary/test-results/r03-s007-final-tracker-lint-02`.
No executable source or generated projection changed after E-070. No further
workstream is activated; future optimization proposals inherit the explicit
limitations and migration boundaries in E-072 rather than an implementation claim.

**E-074 — WF-00e continuation registration.** The explicit implementation request
activates only the serial continuation in Section 6. Clean Git status and HEAD
`f42fc00b76ec4e296c2a379c09687e02ce669de7` were revalidated. No live invocation
was adopted. E-073 remains the completed original handoff. F-010/F-011,
R03-019–024, V-010–012 and T-013–017 register the new scope. Source preservation
excludes only authorized tracker, owner, tooling, schemas, native fixture/runtime
and test projections; no product behavior, ranking/fairness, capacity, lanes,
workers/retries, isolation eligibility or PostgreSQL pooling changes are allowed.

Requirements remain proposals until adopted in the NLSpec before projections.
WF-09 owns diagnostic completeness; WF-10 owns preparation behavior; WF-11 owns
performance decisions; WF-07e owns final assurance and handoff. Principal public
routes are test-slice for harness.browser, harness.command_surface,
harness.test_catalog and harness.evidence_accounting, plus generated/schema
checks. Exact row populations and treatment identities must be frozen before
sampling. New diagnostics use the approved 4,096 allocation / 4 MiB / 65,536-byte
record bounds and 262,144-byte response bound. Public defaults and failure
classes are adopted under WF-09 before implementation. Current performance and
implementation acceptance are UNESTABLISHED. No candidate resource exists.

Boundary command: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-wf00e-tracker-lint-01
make lint-markdown`; native session **59132**, exit **0**, retained summary
`.cartulary/test-results/r03-wf00e-tracker-lint-01/adhoc/lint-markdown/tool-run-summary.json`,
16,410 ms. WF-00e is DONE; this receipt update is checked by
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-wf00e-tracker-lint-02 make lint-markdown`
before WF-09 begins. No product validation or performance experiment is
claimed by registration.


**E-075 — WF-09 diagnostic contract and initial verification (IN_PROGRESS).**
REQ-857/AC-153 and the REQ-844 execution-index amendment were adopted before
implementation. R03-019 maps to `observability/critical-path.mjs`, the existing
canonical event reducer and public explain-run input owner. R03-020 maps to the
native fixture observation owner, independently bounded allocation collector,
execution-index v3 reader and instrumentation-policy v4. New critical-path v1
and allocation-lifecycle v1 schemas project those clauses. No reset lifecycle,
scheduler policy, capacity or product change has been made.

All commands below ran from the repository root with CGO_ENABLED=0. Focused
checks additionally set CARTULARY_HARNESS_CACHE_MODE=off. Each suffix is the
literal CARTULARY_TEST_RUN_ID suffix after `r03-s009-`; roots are
`.cartulary/test-results/r03-s009-<suffix>`. Native exits below are owning Make
exits, distinct from diagnostic-reader exits.

| Suffix | Exact Make selection | Native session / exit | Result and scope |
| --- | --- | --- | --- |
| generate-01 | make generate | 1142 / 0 | Initial coordinated projections; 11,210 ms |
| evidence-01 | make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence | 87585 / 0 | Initial reader/pagination checks, 11,660 ms; intermediate source |
| native-01 | make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime | 32975 / 0 | Initial runtime assurance, 54,662 ms; source subsequently extended, not final closure |
| format-01 | make format | 95523 / 0 | Authored formatting, 8,804 ms |
| finalize-01 | make agent-finalize | 77089 / 2 | Failed schema precheck: generated topology stale after command-input source edit. Ten cleanup steps settled. Receipt retained; no semantic or cleanup failure inferred |
| generate-02 | make generate | 93229 / 0 | Reconciled projections, 10,740 ms |
| finalize-02 | make agent-finalize | 7909 / 0 | Passed, 26,468 ms; retained-run maintenance skipped because RESULTS_DIR was unset |
| schema-01 | make source-check CHECK=schema | 24679 / 0 | Passed, 6,791 ms |
| generated-01 | make source-check CHECK=generated | 68087 / 0 | Passed, 17,053 ms |
| evidence-02 | make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence | 70199 / 0 | Passed, 81,204 ms; paged path/filters/byte bounds and lifecycle reader fault cases |
| broker-01 | make test-slice OWNER=harness.browser ROWS=harness.browser.unit.fixture_broker_smoke | 49079 / 0 | Passed, 72,129 ms; closed retirement reasons and failed settlement |
| command-01 | make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity,harness.command_surface.unit.native_runtime | 6730 / 0 | Passed, 133,797 ms; public selector and native observation coverage |

The final focused round binds source
`sha256:f7d6789451662dec729112577c0d29db0d14a1a1f0a1f62413aa1bc0d2a50bd2`.
Public task-guide discovery passed for evidence_accounting (25272),
command_surface (88777) and browser (47497), all exit 0. A public exact-run
critical-path query with LIMIT=1 returned complete totals and next_offset=1.
Public CHECK=all on command-01 passed independently after native exit.
These are assurance/development receipts, not acceptance performance samples.
WF-09 remains open pending final review, evidence audits and excluded browser/CI
coverage pilots. No optimization or performance acceptance is claimed.


**E-076 — WF-09 final diagnostic refinement and excluded pilot registration.**
The actual retirement cause is now frozen before physical settlement, so a later
cancellation cannot rewrite a capacity-pressure decision. A regression test
covers that transition. Independent lifecycle metadata bounds, canonical stream
identity across wait-page passes and phase time outside unit lifetime are
explicit. Schema and generated inputs remain coordinated; no lifecycle
optimization is active.

The following commands use CGO_ENABLED=0 and fresh literal run IDs prefixed
r03-s009-. Root and suffix conventions are identical to E-075.

| Suffix | Make selection | Native session / exit | Result |
| --- | --- | --- | --- |
| generate-03 | make generate | 36968 / 0 | Passed, 10,700 ms |
| format-02 | make format | 45269 / 0 | Passed, 7,923 ms |
| finalize-03 | make agent-finalize | 16654 / 0 | Passed, 26,648 ms; maintenance skipped because RESULTS_DIR was unset |
| evidence-03 | make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence | 94620 / 0 | Passed, 12,092 ms; cache off |
| broker-02 | make test-slice OWNER=harness.browser ROWS=harness.browser.unit.fixture_broker_smoke | 83073 / 0 | Passed, 2,909 ms; cache off |

The next two runs are explicitly excluded diagnostic/measurement coverage pilots,
not acceptance samples or substitutes for the later treatment warm-ups:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s009-baseline-browser-pilot-01
CARTULARY_HARNESS_CACHE_MODE=off make browser-e2e SUITE=functional
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`, then the same environment
and options with run ID `r03-s009-baseline-ci-pilot-01` and `make ci`.
They SHALL execute serially, with no source edits or competing intentional
workload. Required coverage includes complete charged accounting, canonical
critical-path pages, independently complete allocation lifecycles, native exits,
secret scans and physical settlement. A failed pilot must be diagnosed and
retained; no acceptance source is frozen until WF-09 closes. WF-10 has not begun.


**E-077 — WF-09 browser pilot coverage and queue closure.** Registered browser
pilot `r03-s009-baseline-browser-pilot-01` completed under native session **54962**,
exit **0**; canonical graph duration **1,161,884 ms**, all **137/137** units passed.
Source is `sha256:36bb82ad0459d4f97c8573fafd8753877c8f5a92bd588d9dde5bef0a64169887`;
graph is `sha256:cf92b0a6c12752fe0769de3c8cebf259693395f60c2db17ec6d4f966f33876d0`.
Public `make harness-observability-check RESULTS_DIR=.cartulary/test-results/r03-s009-baseline-browser-pilot-01 CHECK=all`
passed, exit 0, full_validity=1, complete execution/resources with zero omissions,
invocation accounting available. All **276** cleanup records completed, accounting
settlement is **empty**, and the final secret scan passed across **2,296** files.

The public critical-path view returned all **80** path units with LIMIT=100 and
next_offset=null: queue **19 ms**, unit lifetime **1,161,224 ms**. Every unit's
reason partition closes exactly. The path allocation view returned one generation
with **75** logical borrows and a settled terminal-borrower release. Independently
validated execution-index v3 contains **four** complete browser allocations,
**126** borrows, four terminal-borrower releases, **zero** reset-only generations,
and zero lifecycle omissions. Path-local counts and whole-run counts are distinct.

Required accounting measured owning-command elapsed **1,175,171.249048 ms**,
charged CPU **4,008,593,954 us**, and kernel peak charge **12,579,811,328 bytes**.
Selected-unit p95 ready-to-start delay is **1 ms**, cleanup union **48,547 ms**.
These establish metric availability for this excluded pilot, not performance
acceptance. Public exports are retained in `.cartulary/remediation03/continuation/`
as `browser-pilot-01-performance.json`, `browser-pilot-01-units.json` and
`browser-pilot-01-allocations.json`.

Final review made queue-partition rejection explicit in REQ-857 and the view's
model, with a contradictory-partition test. This is a reader/test-only refinement;
no producer, graph, scheduling or fixture behavior changed. It does not relabel
the pilot as current-source execution. `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s009-finalize-04 make agent-finalize` passed under native
**53016**, exit **0**, **27,777 ms**; RESULTS_DIR remained unset. Focused evidence
run `r03-s009-evidence-04` precedes the registered CI coverage pilot. WF-09 remains
IN_PROGRESS; WF-10 has not started. WF-00e closing lint `r03-wf00e-tracker-lint-02`
was native **34505**, exit **0**, **9,180 ms** before WF-09 began.


**E-078 — WF-09 CI pilot admission (IN_PROGRESS).** Queue reconciliation now
validates the complete path before view/filter selection, so units, waits and
allocations reject the same contradictory facts. The evidence row passed in
`r03-s009-evidence-04` (native **31993**, exit **0**, **12,216 ms**) and
`r03-s009-evidence-05` (native **44634**, exit **0**, **12,232 ms**); each used the
E-075 exact row, CGO_ENABLED=0 and cache off. Both independent public CHECK=all
audits passed. `r03-s009-finalize-05`, exact command `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s009-finalize-05 make agent-finalize`, passed native
**91473**, exit **0**, **27,389 ms**, before the final evidence row and CI pilot.
Maintenance was skipped because RESULTS_DIR was unset.

The E-076 registered CI pilot is active under native session **58875**, root
`.cartulary/test-results/r03-s009-baseline-ci-pilot-01`. Its **600** units use source
`sha256:39935f230e200cd8f9059b2f0afc4c98f5a2f6098bcecfa6455a23eb4733ac4b`,
graph `sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
The graph equals the historical motivation graph; the source does not. Native
exit, terminal scan, coverage audit and physical settlement are pending. No
live invocation is drained or interrupted for a schema cutover. WF-10 remains
unstarted until this pilot and the WF-09 tracker/lint boundary complete.

**E-079 — WF-09 CI pilot failure, coverage and required repairs.** E-078's native
session **58875** exited **2** after canonical **3,594,475 ms**: **595/600** units
passed, five failed, none skipped/cancelled. The exact command/source/root remain
E-076/E-078. All **1,154** cleanup records completed; the retained secret scan
passed **9,314** files. `make explain-run RESULTS_DIR=<E-078 root> DETAIL=summary`
exited **0**. An invalid `DETAIL=summary JSON=1` inspection exited **2** without
mutation. `make harness-observability-check RESULTS_DIR=<E-078 root> CHECK=all`
(native **24942**) exited **2**, normalized artifact **11**, because full evidence
closure requires a passing invocation. The independently scoped `CHECK=diagnostics`
(native **73411**) exited **0**, full_validity **0**; this is not a CI pass.

Public performance and critical-path JSON exports exited **0** (native **96586**,
**48038**, **91451**) and are retained under
`.cartulary/remediation03/continuation/ci-pilot-01-{performance,units,allocations}.json`.
The complete 80-unit path has **655,636 ms** graph queue and **2,937,521 ms** unit
lifetime. Its allocation view contains 25 associated generations. Whole-run
lifecycle coverage is complete: **76** generations, **22** reset-only generations,
**63** capacity-pressure retirements and **13** terminal-borrower retirements;
all 76 settled, zero lifecycle omissions. Verbose execution/resources were
truncated, explicitly; this did not truncate lifecycle records. Required accounting
is available: owning-command **3,669,694.172115 ms**, charged CPU **11,469,915,619 us**,
kernel memory peak **14,202,494,976 bytes**, empty settlement, child exit **2**.
Selected-unit p95 queue is **3,148,162 ms** and cleanup union **157,503 ms**. These
are excluded exploratory observations, not acceptance samples or a speedup claim.

Failed rows are `harness.command_surface.unit.native_runtime` and browser
integration rows `native_graph_observations`, `native_selected_service_lifecycle`,
`postgres_template_preparation`, `postgres_owned_barrier_observation`. Three
embedded graph probes hit the required-accounting native executable-identity
check; two self-owned PostgreSQL paths lack the required reaper policy. Repair
the probes through the prepared graph executable and project the accounted
descendant cleanup policy, preserving binary identity checks and exact fixture
cleanup. REQ-856 clarifies these boundaries before implementation. REQ-857 also
requires independent allocation-begin registration so a missing terminal
observation cannot falsely report complete coverage after verbose truncation.
Validate these repairs with the failed exact rows under required accounting,
preserving this failed attempt. WF-09 remains IN_PROGRESS; WF-10 has not begun.

**E-080 — WF-09 accounting-boundary repair verification (IN_PROGRESS).** The
independent allocation registration/omission repair and accounted descendant
reaper policy are implemented. Embedded integration probes now launch the
prepared native graph executable with inherited worktree admission and graceful
cancellation; they retain the executable-content check and shared receipt owner.
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<id> make format` passed for
`r03-s009-format-03` (native **79549**, **7,807 ms**) and `format-04` (native
**6402**, **7,904 ms**). The same command pattern with `make agent-finalize`
passed `r03-s009-finalize-06` (native **46751**, **26,399 ms**) and `finalize-07`
(native **40083**, **26,495 ms**); all native exits **0**, maintenance skipped
because RESULTS_DIR was unset.

The exact focused command is `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<id>
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.unit.native_runtime HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required`. Attempt `r03-s009-accounted-native-01` (native
**49632**) exited **2** before graph creation after a native edit invalidated the
prepared executable; its root contains only the rebuild receipt. After finalizing,
`r03-s009-accounted-native-02` (native **27708**) passed, exit **0**, **54,100 ms**,
source `sha256:6a4f44ded2a21604fa8e1bb34daddd8b6785d177dd3e31947003c38a02907fe0`.
Public CHECK=all passed with full_validity **1**. Required accounting covered
four graph memberships and empty settlement; all 10 cleanup records completed
and the retained scan passed 77 files.

`r03-s009-accounted-boundaries-01` used the same environment/diagnostic/accounting
settings with OWNER=harness.browser and exact ROWS
`harness.browser.integration.native_graph_observations,harness.browser.integration.native_selected_service_lifecycle,harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_owned_barrier_observation`.
Native **84828** exited **2**, canonical **27,295 ms**: four of six units passed,
two failed. Both self-owned PostgreSQL rows passed. Nested graph probes collided
on run ID `postgres` across different parent directories; the shared receipt
registry correctly rejected ambiguous membership and withheld complete accounting.
The outer cleanup (15 records), completed pure child (10) and completed PostgreSQL
child (18) all settled and scanned successfully. Rejected children stopped at
graph registration, before fixture acquisition. No failed evidence is replaced.
REQ-856 now explicitly requires invocation-unique nested run IDs; the probes use
distinct `native-observation-*` and `native-selection-*` identities. Fresh focused
verification is required before WF-09 closure. WF-10 remains unstarted.

**E-081 — WF-09 assurance, coverage and boundary decision.** Final retained
diagnostic source is
`sha256:6626f44a5d83eae47199214abf8c26bf4a1fff7d90be8875fda4c94c2937ecd0`
on HEAD `f42fc00b76ec4e296c2a379c09687e02ce669de7`. No S-010 lifecycle change is
present. The following current-source checks passed, native exit **0**:

| Exact command, with CGO_ENABLED=0 and CARTULARY_TEST_RUN_ID set to run ID | Run ID suffix after r03-s009- | Native session | Canonical ms |
| --- | --- | --- | --- |
| make agent-finalize | finalize-08 | 49459 | 26,412 |
| CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_graph_observations,harness.browser.integration.native_selected_service_lifecycle,harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_owned_barrier_observation HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required | accounted-boundaries-02 | 56492 | 103,173 |
| CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity,harness.command_surface.unit.native_runtime | command-02 | 25497 | 118,440 |
| make source-check CHECK=schema | schema-02 | 5359 | 40,839 |
| make source-check CHECK=generated | generated-02 | 9112 | 51,744 |
| make source-check CHECK=catalog | catalog-01 | 63654 | 1,603 |

Every root is `.cartulary/test-results/<run ID>`. Public CHECK=all audits passed
for accounted-boundaries-02 and command-02. The former covers seven graph
memberships, complete scoped diagnostics, required accounting with empty
settlement, all 15 outer cleanup records completed and a passing 307-file scan.
The latter covers the ordinary diagnostic-off path and independent evidence
reader probes, all 12 outer cleanup records completed and a passing 230-file scan.
Finalization maintenance remained skipped because RESULTS_DIR was unset.

| Requirement / authority | Implementation | Validation and disposition |
| --- | --- | --- |
| R03-019; REQ-857/835, AC-153 | Existing explain-run surface; critical-path model, canonical wait callback/digest, v1 response schema | E-075–E-078 and current command-02 nested evidence row: complete paging, filters/ties, bounded records, phase unions, queue reconciliation and read-only rejection pass |
| R03-020; REQ-857/844 | Native fixture retirement observations, independent allocation registration/quota, index v3 and policy v4 | Broker/native fault matrices, current native tests and E-077/E-079 real pilots establish lifecycle coverage; missing terminal observation remains explicit |
| REQ-856/847, AC-152 | Prepared executable probes, unique nested identities, outer receipt ownership and accounted self-owned-container policy | E-080 required native row; current four boundary rows and ordinary native row pass; identity fences and exact cleanup remain enforced |
| Projection fidelity | Authored owners and schemas; Make-generated validators/task/topology projections | Current schema/generated/catalog checks pass; no executable Markdown dependency or legacy translator introduced |

Assurance is accepted for S-009; performance acceptance is **not applicable** to
this diagnostic slice. E-077 and E-079 are excluded pilots establishing metric
availability. E-079 remains a failed historical attempt; its five failed rows
have the explicit repair receipts above, and it is not a current-source CI pass.
The mandatory baseline warm-up and final CI still must pass later in the approved
sequence. Verbose diagnostic truncation remains visible; required kernel
accounting and the independent lifecycle section supply their declared coverage.

Migration: diagnostic schemas/readers/producers cut over together; no live run
was switched or translated. The accepted source will be preserved before S-010;
S-010 rollback must retain these diagnostic/accounting correctness repairs.
No unresolved fixture ownership was reported by the completed repair checks.
Next entry requires this tracker update and successful `make lint-markdown`;
`r03-s009-tracker-lint-01` passed (native **86276**, exit **0**, **15,500 ms**).
This final status update is checked by `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s009-tracker-lint-02 make lint-markdown` before S-010
begins. S-010 entry is authorized after that closing check succeeds.

**E-082 — WF-10 entry, source preservation and owner adoption.** WF-09 closing
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s009-tracker-lint-02 make lint-markdown`
passed (native **92587/0**, **9,620 ms**). The accepted diagnostic baseline is
`sha256:6626f44a5d83eae47199214abf8c26bf4a1fff7d90be8875fda4c94c2937ecd0`
on unchanged HEAD `f42fc00b76ec4e296c2a379c09687e02ce669de7`.
Before lifecycle edits, native **86664/0** preserved all 7,499 tracked/nonignored
source files with per-file hashes and modes in the internal
`.cartulary/remediation03/continuation/diagnostic-baseline-source.json` manifest
and `diagnostic-baseline-source.tar.gz` archive (51,977,038 bytes;
SHA-256 `f82256188b44a699dcca1169dafab2949113d07d8694c44295c5a2fa2b39cfbc`).
These backups are not executable harness inputs. Rollback SHALL restore only
lifecycle-owned changes and SHALL preserve tracker history and accepted WF-09
repairs. No live invocation was switched.

Before changing projections or implementation, this receipt adopts NLSpec
REQ-858/AC-154 and reconciles REQ-801/806/452, runtime ownership, reset deadlines
and evidence paths. R03-021/022 map to ordered preparation in the compiler,
native execution/fixture owners, canonical events/results and browser finalizers;
V-011 owns selection, identity, deadline, failure and continuity assurance.
The atomic contract versions and closed step/result interfaces are normative in
the NLSpec. Implementation and assurance remain IN_PROGRESS; performance is
PENDING and requires both V-012 workloads. No change to ranking, fairness,
capacities, lanes, workers, retries, isolation or pooling is authorized. Next:
implement the adopted projections and lifecycle, run focused assurance, then
record the WF-10 disposition and successful boundary lint before WF-11.


**E-083 — WF-10 implementation and exploratory assurance receipts.** The compiler
now transfers each selected boundary reset into the successor's ordered
preparation contract, including its cost, deadline and evidence. The native owner
retains one host admission and fixture borrow through preparation and runner;
fixture validation fences the receipt to the active borrow and next generation.
Canonical phase declarations, complete step outcomes, independently published
step completion receipts and browser finalizers project REQ-858. Completion
receipts preserve finished preparation after owner loss without inventing unit
completion. The owner amendment preceded these projections. No performance
acceptance is claimed; S-010 remains IN_PROGRESS.

All commands below used `CGO_ENABLED=0`, the indicated fresh
`CARTULARY_TEST_RUN_ID=r03-s010-<suffix>` and repository-root Make. Test slices
also used `CARTULARY_HARNESS_CACHE_MODE=off`. Roots are
`.cartulary/test-results/r03-s010-<suffix>`; native session/outer exit are distinct
from the reported graph status.

| Suffix | Command or exact selection | Native / exit | Outcome and repair |
| --- | --- | --- | --- |
| generate-01 / format-01 | `make generate` / `make format` | 5318/0; 56606/0 | Passed; 11,540 / 8,017 ms |
| native-01 | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime,harness.command_surface.unit.scheduler_smoke` | 1078/2 | Pregraph build rejected duplicate event field; repaired |
| native-02 | Same native selection | 82612/2 | 28,649 ms; stale cache/result/reset fixtures rejected; repaired. Exploratory source overlapped edits; not acceptance evidence |
| format-02 / generate-02 | `make format` / `make generate` | 71729/0; 41022/0 | Passed; 8,047 / 10,920 ms |
| native-03 | Same native selection | 18811/2 | 56,078 ms; new test incorrectly assumed exclusive work always followed the cohort. Replaced with order-independent assurance and a deterministic waiter becoming ready during preparation |
| browser-contract-01 | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.playwright_adapter,harness.browser.boundary_support.preparation_contract,harness.browser.unit.browser_lane_policy` | 73618/2 | 8,882 ms; two synthetic group fixtures lacked the new borrow binding; repaired |
| generate-03 / format-03 / generate-04 | Corresponding Make target | 10587/0; 72551/0; 82205/0 | Passed; 10,970 / 7,913 / 11,020 ms |
| native-04 | Same native selection | 76726/0 | 57,489 ms; independent `make harness-observability-check RESULTS_DIR=.cartulary/test-results/r03-s010-native-04 CHECK=all` passed, full validity 1 |
| generate-05 | `make generate` | 44492/0 | Passed; 10,900 ms |
| browser-contract-02 | Browser selection above plus `harness.browser.unit.native_fixtures` | 8128/2 | 12,253 ms; five renderer-pressure synthetic allocations lacked a private lease generation. Updated fixtures; production validation retained |
| format-04 / format-05 / generate-06 / format-06 | Corresponding Make target | 61139/0; 29806/0; 22427/0; 35186/0 | Passed; 7,999 / 7,975 / 10,890 / 7,994 ms |
| native-05 | Same native selection | 29338/0 | 57,287 ms; durable receipt and cancellation-boundary assurance passed. Independent `CHECK=all` passed, full validity 1 |
| browser-contract-03 | Browser selection including native fixtures | 66011/0 | 12,448 ms; all selected rows passed. Independent `CHECK=all` passed, full validity 1 |
| routing-01 | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | 35076/2 | 41,741 ms; test import preceded its shebang. Repaired |
| routing-02 | Same routing selection | 79685/2 | 45,884 ms; old assertions still treated preparation as separately claimed scheduler work. Replaced with assertions that only the cohort owns claims and policy |

Native-04 source was
`sha256:d51c844270674796f259e501d577b619cc60eeffa425efa91099056de9285fb3`;
native-05 source was
`sha256:0d0258de3364afde6144acf55ca0d26c2e3527694f8d47c38202b3d0163f2b38`.
Both used graph
`sha256:197de09b6027ed8a15e8822dedbdb1f895a974a172ee0df28e361027aa9d830c`.
Later canonical negative tests remain subject to their own current-source check.
Native-02/03/04/05 each settled all 13 cleanup records; secret scans passed over
43/85/85/85 files respectively. Browser-contract-01/02 settled all 13/14 records
and passed secret scans over 29/31 files. The unsupported diagnostic query
`make explain-run RESULTS_DIR=.cartulary/test-results/r03-s010-native-03
DETAIL=failures` returned usage exit 2; the supported children view then succeeded.
**E-084 — WF-10 compiler and evidence closure, finalization preparation.**
The E-083 routing command passed as `r03-s010-routing-03` (native **89203/0**,
**51,232 ms**); its independent `CHECK=all` passed with full validity 1. The
source was `sha256:43ad1e337ff502d45eb413801f3de0c5cf5c6ed3b00fc5142cbf795a64695e04`.
Static review then corrected explicit UTF-8 decoding at the new browser borrow
reader before live execution.

`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=r03-s010-evidence-01 make test-slice
OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.current_epoch_evidence` failed (native
**86977/2**, **12,159 ms**): the newly authored canonical fixture placed wait start
later than eligibility. Attempt `evidence-02` failed (**64059/2**, **12,240 ms**)
because its phase IDs did not follow the existing `phase:<integer>` grammar.
Both fixture defects were repaired without relaxing canonical validation.
`evidence-03` passed (**17846/0**, **12,272 ms**, 108 tests) on source
`sha256:041bd791a4c84b8a9f674bfb4d3e3e6205a1266396e5dda7ab66aa8293d8472b`;
independent `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/r03-s010-evidence-03 CHECK=all` passed,
full validity 1. All routing/evidence attempts settled 11 cleanup records and
passed scans over 16 files. Browser-contract-03 settled 14 and scanned 34 files.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s010-finalize-01 make agent-finalize`
failed (**74333/2**, **4,273 ms**) before mutation: shape validation detected stale
input hashes for the two changed browser adapters. Its typed
`unit-artifacts/finalize-summary.json` records no rollback needed and skipped
retained-run maintenance because `RESULTS_DIR` was unset. The corrective
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s010-generate-07 make generate` passed
(**14164/0**, **10,760 ms**). Current finalization and live continuity remain
required; these are correctness receipts, not performance samples.

Routing-01/02 each settled all 11 cleanup records. Diagnostic usage attempts
with an internal TARGET for logs, omitted TARGET for logs, and OWNER directly
on target-plan returned exit 2 without running tests; public help established
`TARGET=test-slice` for logs and a JSON `INPUTS` file for planning. The corrected
continuity plan contains exactly two selected cohorts and no reset scheduler unit.
All failed attempts remain retained. No golden, capacity, scheduler policy or
isolation declaration changed. Next: compiler/evidence assurance, a real
multi-cohort continuity run, then the S-010 boundary decision and Markdown lint.


**E-085 — WF-10 current-source structural assurance and live-run entry.**
The owner now explicitly excludes path-like step identities, permits zero
estimated cost, and distinguishes native runner-phase entry from physical child
launch. Schemas share the graph's step-ID definition. The native executor retains
cancellation at the runner-entry boundary without creating contradictory phase
and unit-result facts. These refinements preceded their projections and tests.

All commands used `CGO_ENABLED=0` and fresh `r03-s010-<suffix>` IDs. Test slices
used cache mode `off`; no timing below is an optimization acceptance sample.

| Suffix | Public command / selection | Native / exit | Result |
| --- | --- | --- | --- |
| finalize-02 | `make agent-finalize` | 14712/0 | 26,797 ms; generated unchanged; retained maintenance skipped, `RESULTS_DIR` unset |
| native-06 | E-083 exact command-surface native/runtime selection | 99980/0 | 62,150 ms; 13 cleanup records completed; secret scan passed, 82 files |
| schema-01 | `make source-check CHECK=schema` | 90114/0 | 6,310 ms; 12 cleanup records completed; scan passed, 19 files |
| generated-01 | `make source-check CHECK=generated` | 56872/0 | 61,333 ms; 14 cleanup records completed; scan passed, 29 files |
| format-07 / format-08 / generate-08 | Corresponding Make target | 13535/0; 14670/0; 37604/0 | Passed; 7,907 / 8,030 / 10,900 ms |
| finalize-03 | `make agent-finalize` | 56113/0 | 26,894 ms; current projections passed; retained maintenance skipped, `RESULTS_DIR` unset; 10 cleanup records completed and scan passed over 15 files |
| routing-04 | E-084 exact test-catalog routing selection | 80986/0 | 51,119 ms; compiler boundary, identity, zero-cost and unsafe-ID tests passed; independent `CHECK=all` passed |
| native-07 | E-083 exact command-surface native/runtime selection | 72118/0 | 108,164 ms; current preparation/cancellation/failure and scheduler tests passed; independent `CHECK=all` passed |
| lint-scripts-01 | `make lint CHECK=scripts` | 47050/0 | 30,762 ms; passed |

Finalize-02 through generated-01 used source
`sha256:fd84420b17e4d5107498f268835844b436830edc27f2be25414d23f2b1a36eb4`.
Finalize-03 and the current focused checks use
`sha256:8adc73fda425f33d7fba4aed1fda6d88a42f0788be430111ebc1366965dc27dc`.
Native and routing checks overlapped as correctness work; their durations SHALL
NOT be compared as performance evidence.

The live, excluded continuity check is
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=r03-s010-continuity-01 HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required make test-slice OWNER=module.networkflow
ROWS=module.networkflow.browser_stateful.pagination_recovery,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf`.
It owns native session **3334** and exact root
`.cartulary/test-results/r03-s010-continuity-01`. The public target plan, using
`.cartulary/remediation03/continuation/continuity-inputs.json`, selects exactly
two cohorts and graph
`sha256:9c18bf8b443cde98df8496e8bfb3be097d91717c72befdd8f5b27fabc76d7e1c`.
Terminal outcome, physical settlement and reset/runner continuity remain pending.


**E-086 — WF-10 assurance complete; performance acceptance pending.** The retained
candidate source is
`sha256:8adc73fda425f33d7fba4aed1fda6d88a42f0788be430111ebc1366965dc27dc`
on HEAD `f42fc00b76ec4e296c2a379c09687e02ce669de7`. REQ-858/AC-154 were adopted
before their projections. This receipt closes implementation assurance only;
S-011 SHALL independently accept both workloads before the lifecycle optimization
can be retained as accepted functionality.

| Owner requirement | Implementation projection | Validation receipt |
| --- | --- | --- |
| REQ-801/803/806/858: selected predecessor boundary and one cost charge | `work-graph/browser.mjs`, graph model and graph v9 | E-084/E-085 routing-03/04: direct, owner, partial, aggregate, singleton, empty and incompatible selections; semantic sensitivity, 16-step bound, zero-cost and unsafe-ID rejection |
| REQ-852/858: one admitted borrow, independent deadlines and complete outcomes | Native lifecycle/preparation/process owners, runtime plan v7 and unit event/result v5/v2 | E-083/E-085 native-05/06/07: order, cancellation boundaries, owner-loss cancellation, receipt/publication failures, product/cleanup failure, protected bindings, forced exclusive waiter and independent timeouts |
| REQ-806/858: exact generation continuity and unexecuted-row accounting | Native fixture preparation validator, browser receipt/group/finalizer projections | E-083 browser-contract-03; native failure continuation; continuity-01 actual reset/runner binding and public evidence audit |
| REQ-835/857/858: canonical interpretation and durable completed preparation | Canonical reader, per-step completion receipt and retained-selection validation | E-084 evidence-03; native durable-receipt tests; current native-07 includes current-epoch evidence assurance; independent audits of evidence-03, native-07 and live continuity-01 passed |
| Current contract closure | Generated schema attachments, validators, runtime policy and task/topology projections | E-085 finalize-03, schema/generated checks; finalization retained no stale generated changes; catalog-01 below |

The E-085 live continuity command completed in **124,845 ms**, native **3334/0**,
with all selected rows passed. `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/r03-s010-continuity-01 CHECK=all` exited **0**,
full validity **1**. Canonical, allocation, execution and resource diagnostic
coverage is complete with zero disclosed omissions; invocation accounting and
outer terminal receipt are available. Every one of **25** cleanup records
completed; the retained-secret scan passed over **126** files.

Exactly one browser-stack generation served two logical borrows. The successor's
`reset-2ce93809f70df63312b2fdb5` preparation passed in **3,100 ms**. Its reset
receipt and browser group result both bind `lease-000002`, allocation
`allocation:974fd5fc-8216-45c7-9112-4f1b564f54c4`, the same concrete browser
session, and backend generation **2** (previous generation **1**). The first
cohort had no leading reset. The allocation retired for
`terminal_borrower_release`, triggered by the successor, with settlement
`settled`. There was no reset-only generation. Public `DETAIL=critical-path`
exports (units and allocations, `LIMIT=100`, JSON, both exit 0) retain a complete
7-unit path, 1 ms graph queue and 124,361 ms unit lifetime in
`.cartulary/remediation03/continuation/continuity-01-{units,allocations}.json`.
These observations prove this execution's continuity, not a speedup.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s010-lint-go-01 make lint CHECK=go`
passed (**12310/0**, **12,021 ms**). `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s010-catalog-01 make source-check CHECK=catalog` passed
(**78313/0**, **3,344 ms**). Native-07, routing-04 and lint-scripts-01 settled all
13/11/11 cleanup records and passed secret scans over 82/16/15 files.
No candidate-owned resource remains unresolved in the completed runs. No live
invocation was migrated. No golden, scheduling policy, capacities, isolation or
pooling changed. The preserved E-081 diagnostic baseline remains the rollback
boundary. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

WF-11 entry requires the boundary command `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s010-tracker-lint-01 make lint-markdown` to pass.
The command passed (**56882/0**, **15,950 ms**). This closing receipt text is
checked by `r03-s010-tracker-lint-02` before the next workstream starts.
After that receipt, freeze both sources, exact browser/CI selections and graph
treatment map, storage/toolchain/accounting scope, warm-ups, run order and all
sample IDs before sampling. A workload rejection SHALL restore the lifecycle
owner, projections and implementation together while retaining WF-09 diagnostics.


**E-087 — WF-11 entry and experiment 01 freeze.** WF-10 closing Markdown lint
`r03-s010-tracker-lint-02` passed (**93923/0**, **9,660 ms**), after E-086's tracker
update and before this workstream. No live invocation remained during source
cutover. The candidate snapshot contains 7,507 source files in
`.cartulary/remediation03/continuation/lifecycle-candidate-source.tar.gz`
(48,065,866 bytes; SHA-256
`64f18f3ed0090ccc12b9bb8674dbadc4edc9aea5a7f07cb63991000e5a1d7178`),
with per-file hashes/modes in the adjacent JSON manifest. E-082 preserves the
baseline. The lifecycle treatment map contains 95 authored and 264 generated
changed paths. This controlling tracker is preserved across every switch.
Frozen owner documents and projections travel with their corresponding source;
no live invocation is translated to another contract version.

| Frozen treatment | Executable source identity | Meaning |
| --- | --- | --- |
| Baseline B | `sha256:6626f44a5d83eae47199214abf8c26bf4a1fff7d90be8875fda4c94c2937ecd0` | Accepted WF-09 diagnostics and original separately scheduled resets |
| Candidate C | `sha256:8adc73fda425f33d7fba4aed1fda6d88a42f0788be430111ebc1366965dc27dc` | Same diagnostics plus WF-10 preparation lifecycle and necessary contract/assurance projections |

Public target-plan exports, with exact JSON inputs, freeze these populations:

| Workload | Command | Rows / cohorts | B → C work units | Removed reset units |
| --- | --- | --- | --- | --- |
| browser | `make browser-e2e SUITE=functional` | 383 / 65 | 137 → 76 | 61 |
| ci | `make ci` | 2,279 / 96 | 600 → 517 | 83 |

Browser graph B/C identities are
`sha256:cf92b0a6c12752fe0769de3c8cebf259693395f60c2db17ec6d4f966f33876d0`
and `sha256:30890f4b065d42f7753b88b77e968489010abdde3d7d98fb5f9c1146dc8bb4d5`.
CI identities are
`sha256:5ec94b87768a0e55509ad63f69d1f9e0411ada58f620ab22742130900375c869`
and `sha256:b4fa0ade1409ed954b6f33211b5fd4c47d8640b49669a6aad2af94096ee035ae`.
The full row populations, compiled plans and explicit reset-to-successor maps are
in `.cartulary/remediation03/continuation/{baseline,candidate}-{browser,ci}-plan.json`
and `performance-graph-treatment.json`. Their sorted row-population SHA-256
identities are `6fdc6c0849ef28d983d03adf15116013cdcaef93b91fb09687e91655f3556196`
(browser) and `33fa243ebdb3aaef32b86a2231cf09b4ad4332e3a74f7d1c4155886b91233eac`
(CI). Static analysis confirms identical rows, cohort IDs, retained-unit ownership,
claims, fixture capabilities, services, cache policies and top-level required
outputs. Each removed reset maps to exactly one successor; its estimate transfers
once and the complete graph estimate is unchanged. Semantic digests, dependency
edges, preparation evidence and versioned projections differ as declared.

Both treatments SHALL use `CGO_ENABLED=0`,
`CARTULARY_HARNESS_CACHE_MODE=off`, `HARNESS_DIAGNOSTICS=basic`, and
`HARNESS_ACCOUNTING=required`. Dependency/build caches SHALL remain warm.
`performance-environment.json` freezes Linux
`6.6.114.1-microsoft-standard-WSL2`, x86_64, the same repository/storage device
(4096-byte filesystem blocks), toolchain digest
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`,
system digest
`sha256:008fe637b622cd8932b55cc1d8be5346bae02c09ab35d7518fc9f72b66eb6c21`
and unchanged execution topology. Charged CPU and kernel memory cover the proven
delegated Linux invocation and owned containers; controller, unrelated/shared
host work and Windows-host totals remain excluded. No competing intentional
workload or structured polling during quiet measurement is admitted.

`performance-experiment.json` freezes twenty possible attempts before sampling.
Run IDs are `r03-s011-exp01-<browser|ci>-<warm|p01|p02|p03|p04>-<b|c>`.
Order SHALL be B browser warm-up, B CI warm-up, C browser warm-up, C CI warm-up;
then each pair's browser members followed by its CI members. Odd pairs are B→C;
even pairs C→B. Exactly one warm-up per treatment/workload is excluded. Both
workloads SHALL receive their first two pairs unless correctness/cleanup prevents
execution, including when the other workload has already rejected. Optional
later attempts SHALL be explicitly marked omitted with reasons.

Section 8.5 remains binding: both first-two improvements strictly below 5% reject;
either at least 20% permits early acceptance after those two pairs, with no omitted
confidence claim; otherwise four pairs require a >5% one-sided 95% benefit lower
bound and ≤5% one-sided 99% upper bounds for all five guardrails. Numerical
analysis SHALL use decimal precision 28, ROUND_HALF_EVEN, Python MT19937 Random,
10,000 paired resamples, seed 20261021 reset per workload, common sample indices
across metrics, and nearest-rank `ceil(p*N)-1` quantiles. Whole runs are samples.
Primary elapsed is `invocation_accounting.metrics.duration_ms`, the established
owning-command measurement; the terminal receipt independently corroborates outer
completion. Remaining guardrails are
charged CPU, kernel peak memory, complete-population p95 ready-to-start delay and
cleanup union duration. Cleanup success is unconditional.

Every attempt SHALL be independently validated through public retained readers
while its own source/contract version is installed. Exports SHALL be preserved
before switching source. Changed-source observations SHALL NOT use
`COMPARISON=equivalent`; internal numerical analysis applies the explicit treatment
map. Complete critical-path pages SHALL report queue versus host admission,
acquisition/preparation/runner/settlement intervals, lifecycle causes, reset-only
generations and continuity. The changed queue-percentile populations SHALL be
reported explicitly, with supplementary matched-cohort waits and their observed
coverage rather than treating removed reset rows as unchanged samples.
Missing required measurements make acceptance inconclusive. Failed attempts remain
recorded; repaired executable source requires a newly frozen experiment.

The initial B restoration verified all frozen files except this preserved tracker,
then `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-exp01-b-generate-01
make generate` passed (**77299/0**, **10,670 ms**). Regeneration was byte-identical
to the frozen baseline. `source-switches.jsonl` records each immutable restoration;
no generated bytes are manually edited. Sampling has not started at this receipt.


**E-088 — Experiment preparation failure preserved; no sample admitted.** E-087
freeze lint `r03-s011-freeze-lint-01` passed (**39709/0**, **15,980 ms**).
The first excluded B browser warm-up attempt,
`r03-s011-exp01-browser-warm-b`, returned native **74288/2** before graph execution.
The exact-root public explanation exited 0 and reported only a passing
`harness-runtime-build` preparation (1,700 ms); no graph manifest, row result or
complete invocation-accounting receipt exists. The owning command reported
`config/configuration_error` and incomplete accounting. It is not a passing
warm-up, acceptance sample, or resource guarantee.

Diagnosis: the outer accounting controller started from the previously installed
candidate executable, and the inner Make prerequisite rebuilt the baseline
runtime. Existing `VerifyRuntimeExecutable` intentionally rejects a controller
and graph running different native executable bytes. This is the documented
measurement-preparation boundary, not a reason to weaken accounting. No graph or
fixture acquisition began. The accounting owner's ordinary scope teardown ran;
no complete resource claim is inferred from the missing receipt.

Source remains the frozen B identity. Before every source switch's accounted
work, the procedure SHALL regenerate and execute an unaccounted public
`CGO_ENABLED=0 make source-check CHECK=catalog` to prepare the installed native
controller before its outer wrapper begins. This is dependency/build preparation,
not an acceptance sample or additional workload during measurement. The failed
attempt remains in `performance-experiment.json`; the replacement completed-warm-up
slot is explicitly frozen as `r03-s011-exp01-browser-warm-b-02`. The original
pair IDs/order, source treatments, thresholds and one excluded completed warm-up
per treatment/workload remain unchanged. No executable source was repaired and
no same-source product failure is erased.

The baseline preparation command is `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s011-exp01-b-prepare-01 make source-check CHECK=catalog`,
native **47731**. Its successful terminal receipt is required before the
replacement warm-up begins. Other planned attempts remain unexecuted.


**E-089 — Baseline browser warm-up complete, excluded.** E-088's public native
preparation passed (**47731/0**, **1,610 ms**). The replacement warm-up
`r03-s011-exp01-browser-warm-b-02` passed (**40081/0**, graph **1,155,240 ms**)
on the frozen B source and graph. Independent public `CHECK=all` and four public
performance/critical-path JSON exports each exited **0**. Diagnostic execution,
resource and allocation coverage are complete with no disclosed omissions;
required accounting reports settlement `empty`. All **276** cleanup records
completed; the retained-secret scan passed over **2,296** files. The first
pregraph failure remains separately retained under E-088.

The established owning-command measurement is
`invocation_accounting.metrics.duration_ms`: **1,169,442.308664 ms**. Charged CPU
is **3,991,196,494 µs**, kernel peak memory **13,311,836,160 bytes**, p95 graph
ready-to-start **1 ms**, and cleanup union **48,425 ms**. The independent terminal
receipt's quantized duration is corroboration, not a replacement for that
established primary metric. This clarification precedes every acceptance sample.
The complete critical path has 80 units, 15 ms graph queue and 1,154,656 ms unit
lifetime. All four stack allocations retired on terminal borrower release; no
allocation was reset-only. Complete typed lifecycle observations were copied from
the publicly validated execution index for internal numerical analysis; no raw
canonical-event parser was introduced. All exports use the run-ID prefix under
`.cartulary/remediation03/continuation/`; the ledger retains exact metrics.

No acceptance pair is complete. The next frozen command is
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=r03-s011-exp01-ci-warm-b HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required make ci`, also an excluded warm-up.


**E-090 — Runtime graph binding during excluded CI warm-up.** The B CI warm-up
is running in native session **98330**. Its published manifest binds source B,
the frozen system/toolchain identities, 600 units, and host Go worker budget 19.
Its runtime graph is
`sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
E-087's CI graph is the static public target-plan identity. The static compiler
uses its default Go budget of 4; runtime preparation captures host capability
and binds the available budget before compiling. Comparing every projected unit
field, with the documented projection field names mapped, identifies exactly two
differences: `go:38ac079360dd6fb6:ba2f66716e6a` and
`go:5f7c20b93d80a0f3:3fa1ec86e540` each claim 9 CPU tokens at runtime instead of
4 in the static plan, with corresponding semantic digests. All unit IDs and
other projected fields match. This is an existing capability binding, not a
scheduler or capacity change introduced by either treatment.

Before acceptance sampling, both runtime graphs SHALL be frozen from their
excluded warm-ups, alongside the unchanged static treatment map and captured
worker budget. The two runtime graphs SHALL preserve identical populations and
the declared reset/preparation treatment differences. Subsequent samples SHALL
match their treatment's runtime graph and the frozen system identity. The
candidate runtime binding remains unobserved at this receipt. No acceptance
sample has begun. Final CI outcome, diagnostics, scan and settlement remain
pending; live progress is not a completion claim.


**E-091 — Baseline CI warm-up failure observed; correctness entry gate open.**
The still-running `r03-s011-exp01-ci-warm-b` (native **98330**) published a failed
unit `go:836628bb3a2bb63e:e4a7a5ec5858`, selecting exactly
`harness.browser.integration.native_selected_service_lifecycle`. Its unit result
records exit **10**, `product/test_assertion_failure`, no missing declared outputs.
The row records the same classification. Bounded published unit-log inspection
shows all four `TestNativeSelectedServiceLifecycle` subcases (`pure`, `postgres`,
`object-store`, `mixed`) failed because their installed nested graph command
returned exit **2**, `config/configuration_error`, before publishing a child graph.
The public `make explain-run RESULTS_DIR=<exact root> TARGET=ci DETAIL=logs`
reader exited **0** with no target logs at this live publication stage. Direct
inspection used the owning publisher's exact unit-log location and was bounded
to 80 lines / 8 KiB per stdout/stderr artifact; no event reducer was introduced.

The underlying startup cause remains unresolved. This is a required assurance
failure, not an optimization disposition. The failed warm-up SHALL remain
excluded and preserved. No acceptance sample SHALL begin until the cause is
resolved and the required fresh assurance passes. If executable source requires
repair, the experiment SHALL be frozen again with repaired baseline and candidate
sources; this failure SHALL NOT be erased or relabeled as a pass. Independent
work and finalizers continue under the original invocation. Native command exit,
complete failure census, diagnostics audit, terminal scan and settlement remain
pending. No executable source or live contract version has changed.


**E-092 — Failed CI warm-up settled; corrective WF-11a registered.** E-091's
native **98330** exited **2**. The graph reports **599/600** units passed, one
failed, none skipped/cancelled, canonical duration **3,751,876 ms**. Public
`make explain-run RESULTS_DIR=<exact root> DETAIL=children` exited **0** and
confirmed the failure is confined to the selected service-lifecycle row. All
**1,151** cleanup records completed. The retained-secret scan passed over
**9,445** files. Required invocation accounting reports settlement **empty**,
child exit **2**, owning-command duration **3,831,613.589397 ms**, charged CPU
**11,617,914,732 µs**, and kernel peak memory **15,317,274,624 bytes**. Complete
selected-unit p95 queue is **3,312,149 ms**, cleanup union **155,131 ms**.
These are failed, excluded warm-up observations; none establishes acceptance.

Public `make harness-observability-check RESULTS_DIR=<exact root>
CHECK=diagnostics` passed (**22193/0**, full_validity **0**). Verbose execution
and sampled-resource sections are explicitly truncated (1,528 resource omissions);
invocation accounting and terminal receipts are available. Separate allocation
lifecycle coverage is complete: **73** generations, **60** capacity-pressure and
**13** terminal-borrower retirements, **27** reset-only generations, zero lifecycle
omissions, all settled. Public performance and root critical-path exports exited
**0** (**11575**, **21824**, **83701**, **65661**). Three continuation wait pages
(**1149**, **1570**, **43371**, each exit **0**) reconstruct all **345** ordered wait
records; final `next_offset=null`. The browser target-local unit page
(**34375/0**) and root path both contain 80 units, **577,386 ms** graph queue and
**3,173,182 ms** unit lifetime. Exports remain under the E-089 internal directory
with the exact run-ID prefix; the canonical reader owns event interpretation.

WF-11a is a required corrective assurance workstream within the serial WF-11
entry gate. It SHALL diagnose the installed nested-service startup failure,
repair the responsible owner/projection/implementation or test composition, and
validate the exact failing route plus the affected containing boundary. It SHALL
preserve executable-identity verification, ordinary deadlines, resource policy,
and all failed evidence. If source changes, it SHALL apply the common repair to
both frozen treatments, preserve experiment 01, and freeze a new experiment
before sampling. No candidate performance disposition is inferred from this
baseline failure. Existing diagnostic and lifecycle receipts retain their actual
source scope. No candidate-owned resource remains unresolved in this attempt.

The observed runtime-build receipt under
`nested-cancellation/summary/harness-runtime-build` is later than the failed row;
it is a hypothesis for shared executable mutation, not proof of causation. The
startup failure's precise cause remains open. Next: successful tracker lint,
public native preparation, and a narrow reproduction of
`harness.browser.integration.native_selected_service_lifecycle` under required
accounting, followed by evidence-backed repair. No source mutation or rerun has
occurred since the failed invocation ended at this receipt.


**E-093 — Corrective characterization and bounded startup attribution.** The
WF-11a entry lint passed (**50797/0**, **15,130 ms**) before corrective work.
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-corrective-prepare-01
make source-check CHECK=catalog` passed (**16869/0**, **1,603 ms**). All frozen
baseline source fingerprints matched except this tracker. The installed runtime
SHA-256 before and after preparation was unchanged:
`77a59d1159eae198c3ebb5c514f526e36b58b3fd2d2b35a9a8c982a907714ea5`.

The exact failed row, run alone on unchanged B source with the E-087 environment,
passed: `make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.native_selected_service_lifecycle`, run
`r03-s011-corrective-selected-01`, native **75974/0**, **40,809 ms**. Public
CHECK=all passed with full_validity **1**, complete scoped diagnostics, all
**10** outer cleanup records completed, a passing **141**-file scan and empty
accounting settlement. This narrows the containing-context investigation; it
does not erase E-091/E-092 or prove flakiness or a repair.

REQ-856 now adopts closed startup `failure_stage` attribution before its native
projection changes. The dispatcher exposes only the listed stage enum, preserving
class, reason, exit and cleanup. Unknown/unattributed stages become `unavailable`;
private cause text is never printed. `failure.go`, `run_linux.go` and the native
composition entry project this clause. Existing native failure/compilation tests
cover wrapped errors, cancellation/deadline identity, unknown values and rejection
before publication. No graph/event/result schema, scheduling policy or product
behavior changes. The E-087 archives remain intact; the working source is now
an investigative diagnostic refinement of B, not either frozen treatment.

With `CGO_ENABLED=0` and explicit run ID, `make format`
`r03-s011-entry-stage-format-01` passed (**82049/0**, **7,867 ms**), followed by
`make agent-finalize` `r03-s011-entry-stage-finalize-01` (**71766/0**,
**26,380 ms**). Maintenance was skipped because RESULTS_DIR was unset.
The accounted exact native row `make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.unit.native_runtime`, run
`r03-s011-entry-stage-native-01`, passed (**15851/0**, **54,480 ms**) on source
`sha256:0bdbf9970176095beece44c5eed3658ce05ae8bcdaccc5e6520c3ac73bc319cd`.
Public CHECK=all passed, all **10** cleanup records completed, scan passed
**77** files, accounting empty. No source edits occurred during those invocations.

The next containing-boundary investigation is running in native **32871**, run
`r03-s011-corrective-boundary-01`, using the E-087 environment and
`make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_selected_service_lifecycle`.
It tests the preceding physical-fixture routes with the failed service-selection
route. Its outcome and cleanup remain pending. WF-11a remains IN_PROGRESS and
acceptance sampling remains unstarted.


**E-094 — Shared toolchain-root mutation reproduced and repair adopted.**
The E-093 three-row reproduction ended **32871/2**, **129,461 ms**, source
`sha256:0bdbf9970176095beece44c5eed3658ce05ae8bcdaccc5e6520c3ac73bc319cd`,
graph `sha256:5fc6244f1d2a6fb73243caec323199af9bf5051f6f0facf803a22a3dfea4aaea`.
Two units passed; installed-service selection failed after its pure and PostgreSQL
subcases passed. Object-store and mixed subcases rejected at the newly attributed
`accounting_executable_identity` stage. All 12 outer cleanup records completed;
the 165-file retained-secret scan passed. CHECK=diagnostics passed with
full_validity 0 and complete execution/resource observations, but invocation
accounting was unavailable/incomplete. This attempt cannot supply resource
acceptance or complete accounting settlement evidence.

Preserved executable copies under the E-089 internal directory establish the
mutation: fixture preparation produced SHA-256
`a63b7324ab1da92e47c7ea0325fd28b47769a795662dd9ed78e3ceb65e7a4376`
(35,732,154 bytes), while public preparation restored
`3528a07f3a308e59449d0216110fcf447d81bd27664485fddb647d67adcdfc19`
(35,743,970 bytes). Build-info settings and dependencies match, but embedded
standard-library source roots differ: `$HOME/go/pkg/mod/golang.org/toolchain@…`
versus the declared `$HOME/.cache/cartulary/go/mod/golang.org/toolchain@…`.
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-corrective-prepare-02
make source-check CHECK=catalog` passed **88413/0**, **1,665 ms**.

The code trace identifies the missing projection: public admission strips native
cache aliases, while the graph Go row runner supplied the static build profile
without restoring the three resolved machine-state paths. Automatic toolchain
selection could therefore use Go's unrelated default module cache; nested fixture
preparation inherited that toolchain root. Section 14/AC-092 now explicitly
requires graph runner path projection, executable-owned root resolution, and
same-profile nested native executable identity preservation. REQ-856's content
identity check remains mandatory. The runtime-only projection uses the existing
machine-path resolver; static selector discovery remains host independent.
Public Make rejects caller `GOROOT` overrides and strips inherited values;
review preparation strips the inherited root as well. Runner and preparation
contracts cover the polluted environment and invalid path boundaries.

The first `make agent-finalize`, run `r03-s011-toolchain-finalize-01`, failed
**38660/2**, graph **4,253 ms**, before mutation: schema-shape validation detected
stale topology input fingerprints and required `make generate`. All ten outer
cleanup records completed. The failed receipt remains retained. Public generation
is the required next action, followed by finalization and focused assurance;
this is not an accepted repair yet. Original experiment 01 archives remain intact.


**E-095 — Repaired diagnostic baseline assurance and preservation.**
Public generation `r03-s011-toolchain-generate-01` passed **51602/0**,
**11,190 ms**; `make agent-finalize` `r03-s011-toolchain-finalize-02` then passed
**23555/0**, **27,026 ms**. Both used `CGO_ENABLED=0`; maintenance was skipped
because RESULTS_DIR was unset. The common repaired B source is
`sha256:00db8f4ee50678ea3dd9a0c3ccd80b5f930009359b0ff667d0bbfb80012d45c4`.
The following exact selections used the E-087 four environment settings and
`make test-slice`; all were independently audited with CHECK=all (exit 0,
full_validity 1, complete scoped execution/resource observations), all cleanup
records completed, every scan passed, and required accounting settled empty.

| Run ID | Owner / exact rows | Native exit / graph duration | Cleanup / scan files |
| --- | --- | --- | --- |
| `r03-s011-toolchain-boundary-02` | `harness.browser`; `harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_selected_service_lifecycle` | 72668/0; 154,263 ms | 12 / 243 |
| `r03-s011-toolchain-contracts-01` | `harness.command_surface`; `harness.command_surface.behavior.public_registry_parity` | 96140/0; 65,241 ms | 11 / 173 |
| `r03-s011-toolchain-ui-01` | `harness.browser`; `harness.browser.boundary_support.ui_review_lifecycle` | 36622/0; 38,190 ms | 10 / 17 |

The boundary run's prepared native SHA-256 remained
`3528a07f3a308e59449d0216110fcf447d81bd27664485fddb647d67adcdfc19`
after fixture preparation. Its formerly failing nested service probes passed;
all five graph memberships share the proven accounting owner. This supplies
focused causal repair evidence; it does not erase failed attempts or replace a
new full CI warm-up. Public performance exports (exit 0) are retained with each
run-ID prefix under the E-089 internal directory.

`diagnostic-baseline-source-exp02.{tar.gz,json}` preserves 7,499 files,
48,045,248 archive bytes, archive SHA-256
`d6d5329f2494b0d34f9d739db4d2c4cdae84be33930ba4bbd2cd03f3da2fb458`.
The original B/C archives remain untouched. A verified source restoration to
original C preserved this tracker; a clean three-way merge applied the same 15
authored repair files, behavioral owner first. Two generated projections were
excluded from manual transfer and are refreshed through Make. The internal
`repair-overlay.json` records the original/repaired B/C fingerprints.
Candidate generation `r03-s011-candidate-repair-generate-01` passed **92074/0**,
**10,700 ms**. Candidate finalization and focused native/boundary assurance remain
pending; WF-11a is still IN_PROGRESS, and experiment 02 has not admitted samples.


**E-096 — WF-11a complete; both treatments repaired without weakening identity.**
Candidate `make agent-finalize` `r03-s011-candidate-repair-finalize-01` passed
**80750/0**, **26,694 ms**; maintenance was skipped because RESULTS_DIR was unset.
Unaccounted `CGO_ENABLED=0 make source-check CHECK=catalog`, run
`r03-s011-candidate-repair-prepare-01`, passed **52278/0**, **1,666 ms**.
On repaired C source
`sha256:502eda8efafa9fb2002b96be3755b83ce5b543b59dd6b7808b27d4960c8effea`,
the E-087 environment and exact public test-slice selections passed:

| Run ID | Owner / exact rows | Native exit / graph duration | Cleanup / scan files |
| --- | --- | --- | --- |
| `r03-s011-candidate-repair-native-01` | `harness.command_surface`; `harness.command_surface.unit.native_runtime` | 88311/0; 58,082 ms | 10 / 77 |
| `r03-s011-candidate-repair-boundary-01` | `harness.browser`; `harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_selected_service_lifecycle` | 74960/0; 126,857 ms | 12 / 243 |

Both CHECK=all audits passed (exit 0, full_validity 1), all cleanup completed,
scans passed, and accounting settled empty. Native tests include ordered
preparation outcomes and startup failure-stage classification. The containing
boundary reproduces the formerly failing physical-preparation/service-selection
interaction on both repaired treatments. `CGO_ENABLED=0 make help
GOROOT=/foreign/toolchain` (run `r03-s011-go-root-cli-01`) rejected undeclared input
with native exit **2** as required; the corresponding inherited environment
`CGO_ENABLED=0 GOROOT=/foreign/toolchain make help` (run
`r03-s011-go-root-env-01`) passed with exit **0**. No product execution was
expected from these command-admission probes.

Traceability: Section 14/AC-092 → `goExecutionEnvironment`, shared path resolver,
public environment admission and review recipes → E-095 runner/review contracts,
B/C containing reproduction and public override probes. REQ-856 → closed native
startup stage projection → E-093/E-096 native tests. Required content identity,
resource policy, lifecycle semantics and deadlines remain unchanged by this
common repair. Assurance is ACCEPTED for this corrective scope; performance is
NOT ESTABLISHED. The new full warm-ups remain mandatory containing assurance.

The original experiment and failed attempts remain preserved. Candidate snapshot
`lifecycle-candidate-source-exp02.{tar.gz,json}` records the exact repaired source;
its archive identity is included in the next frozen experiment receipt. Public
candidate target-plan exports (**77621/0**, **67460/0**) preserve the E-087 graph
identities, exact populations and reset treatment map. No acceptance samples have
run. WF-11a is DONE; WF-11 resumes only after this receipt passes public Markdown
lint. Next: restore repaired B, regenerate and prepare native execution, freeze
experiment 02 sources/runtime-graph validation rules, then perform new excluded
warm-ups. No unresolved candidate-owned resource remains from corrective checks.


**E-097 — Experiment 02 frozen; no samples admitted.** Corrective closure lint
`r03-s011-corrective-close-lint-01` passed **81206/0**, **15,260 ms**, before WF-11
resumed. Repaired C snapshot contains 7,507 files, 48,074,090 archive bytes,
SHA-256 `9ca5b8bf50189fe84ebc678065b8e9ee128757271d34cd61a00e15ede17a4a11`.
The source restoration C2→B2 verified all source fingerprints except the preserved
tracker. Public generation `r03-s011-exp02-b-generate-01` passed **93470/0**,
**11,020 ms**, byte-identical to frozen B2. Unaccounted native preparation
`r03-s011-exp02-b-prepare-01` passed **33111/0**, **1,641 ms**. Both use
`CGO_ENABLED=0` and the E-087 public commands. Repaired B2/C2 source identities
are respectively:

- `sha256:00db8f4ee50678ea3dd9a0c3ccd80b5f930009359b0ff667d0bbfb80012d45c4`
- `sha256:502eda8efafa9fb2002b96be3755b83ce5b543b59dd6b7808b27d4960c8effea`

The E-087 selection, treatment, environment and numerical protocol apply unchanged.
Fresh public B2 target-plan exports (**60901/0**, **29939/0**) and C2 exports in
E-096 are byte-for-byte equal to their original treatment plans: browser 383 rows,
137→76 units and 61 transferred resets; CI 2,279 rows, 600→517 units and 83
transferred resets. The common toolchain/root repair and bounded startup stage
attribution occur in both treatments and are not lifecycle treatment differences.
No speedup from those common repairs is attributed to S-010.

The old ledger, environment and treatment map are retained as `*-exp01.json`.
The current internal `performance-experiment.json` and immutable
`performance-experiment-exp02-freeze.json` register twenty possible attempts
before execution. Run IDs are
`r03-s011-exp02-<browser|ci>-<warm|p01|p02|p03|p04>-<b|c>`.
Order remains B browser warm-up, B CI warm-up, C browser warm-up, C CI warm-up;
then each pair's browser members followed by CI members, B→C on odd pairs and
C→B on even pairs. The four settings, warm caches, no competing intentional work,
quiet-measurement observation boundary, independent audits, failed-attempt
preservation and greater-than-5% full-protocol threshold remain mandatory.

Runtime graph bindings SHALL be verified from all four successful excluded warm-ups
before any acceptance sample. E-090's runtime Go worker budget is 19 versus the
static-plan default 4; only its two declared CPU-claim/semantic-digest differences
are permitted. Every acceptance run SHALL match its treatment's validated warm-up
source, selection, environment and runtime graph. Any additional difference or
required evidence failure SHALL stop admission pending diagnosis. The first-two
pair rules remain per workload; both workloads SHALL receive those pairs unless
correctness or cleanup prevents execution. All source/contract readers SHALL
validate and export a run before switching treatments. This freeze receipt SHALL
pass Markdown lint before the first new warm-up starts.
Freeze lint `r03-s011-exp02-freeze-lint-01` passed **91884/0**, **8,720 ms**.
The status reconciliation is checked by closing lint
`r03-s011-exp02-status-lint-01` before execution.


**E-098 — Experiment 02 baseline browser warm-up passed and audited.**
Status reconciliation lint `r03-s011-exp02-status-lint-01` passed **60752/0**,
**9,250 ms**, before execution. The exact E-097 first command was
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=r03-s011-exp02-browser-warm-b
make browser-e2e SUITE=functional HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Native **52553/0** completed all **137/137** work units, graph duration
**1,150,911 ms**, frozen B2 source and expected browser graph
`sha256:cf92b0a6c12752fe0769de3c8cebf259693395f60c2db17ec6d4f966f33876d0`.
The root is `.cartulary/test-results/r03-s011-exp02-browser-warm-b`.

Independent public CHECK=all passed, exit **0**, full_validity **1**. Execution,
resource sampling and separate allocation lifecycle sections are complete with
zero omissions; terminal and invocation-accounting receipts are available. All
**276** cleanup records completed, accounting settled **empty**, and the final
**2,296**-file secret scan passed. Owning-command elapsed is
**1,164,259.327581 ms**, charged CPU **3,995,360,833 µs**, kernel peak memory
**8,590,774,272 bytes**, selected-population p95 queue **1 ms**, cleanup union
**46,535 ms**. These are excluded warm-up observations, not acceptance samples.

Public performance and all three critical-path exports passed with exit **0**.
The complete 80-unit path has **13 ms** graph queue and **1,150,295 ms** lifetime;
the unit/wait pages each return 80 records and terminate with `next_offset=null`.
Four stack generations retired for `terminal_borrower_release`, all settled,
zero reset-only generations. The validated full lifecycle section is preserved
separately from the path-only allocation page. All exports use the exact run-ID
prefix under the E-089 internal directory. The experiment ledger records this
completed excluded warm-up and its baseline-browser runtime graph binding.
Next is the frozen baseline CI warm-up. No source changes or performance decision
are authorized by this successful warm-up alone; sampling remains unstarted.


**E-099 — Baseline CI warm-up active with required cleanup failures.**
`r03-s011-exp02-ci-warm-b` runs the frozen `make ci` command with all four E-097
settings in native **90641**. Its published source is B2 and graph is the expected
runtime CI identity
`sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`,
600 units. The command has not exited; no final cleanup/accounting disposition is
claimed. Initial published failures are:

| Unit / affected row | Typed observation |
| --- | --- |
| `go:1aa28ab6cd1a86c2:8f1f46438669`; `module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413` | First cleanup failure in `TestIncidentBundlesEvidenceBlobStagingCleanup_Integration/export_missing_blob`; one forced drop, 5 ms admission, 20,001 ms operation, 19,999 ms effective operation budget; generation `7699393126689120296` |
| `go:21743f28df257b2c:b12e48365b78`; `module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085` | First cleanup failure in `TestRestoreReadinessAndCoherentStoreOrder_Unit`; one forced drop, 5 ms admission, 20,000 ms operation, 19,999 ms effective operation budget; generation `7699393325161295912` |

Both typed row diagnostics show `CheckpointDone`, checkpointer `DataFileSync`,
zero blockers and no completed checkpoint during the sampled failure interval.
Later fixture setup correctly rejects the poisoned PostgreSQL capability with
zero attempts. This is distinct from E-094's executable-identity failure. The
observations prove bounded database cleanup failed while waiting for checkpoint
synchronization; they do not identify a physical Windows-host cause or establish
that the shared Go-path repair caused it. Published unit/row results and bounded
referenced logs are retained under the exact run root. Earlier template sealing
and initialization-checkpoint repair E-070 remains present.

The invocation continues under its existing failure and cleanup policy. No source,
capacity, deadline, retry, isolation or pooling change has been made in response.
Performance sampling remains unstarted and blocked by this required warm-up
failure. Terminal checkpoint observations, physical settlement and root-cause
analysis are required before choosing repair or further execution. No performance
candidate disposition or completed corrective workstream is claimed here.

A third published failure subsequently appeared: unit
`go:f0aa161d0716c72f:331ab287cb5f`, exact row
`platform.postgres.integration.fresh_fixture_admission`. It failed before product
launch with `infra/service_readiness_timeout`. Its referenced service scope
`_shared/test-services/48262af3b2ad425b9e8096ac/service-scope.json` reports
`postgres acquisition sql_connectivity: deadline_expired`, one 26,030 ms startup
attempt, no retry, and `cleanup_failed` / `physical_settlement=unresolved` at
`2026-10-22T08:01:32.232416041Z`. Preflight had passed. This is a distinct startup
failure, not proof of the database-drop cause. The live invocation retains
recovery authority; settlement MUST be rechecked at terminal completion. No
external service action or competing reproduction has been launched.

A fourth failure is independently identified: unit
`go:e962f3a66796e4ee:6e029b7540dc`, row
`platform.rootedfs.unit.secure_absolute_file_reads`, `product/test_assertion_failure`.
`TestSecureFileRejectsUnsafeObjectsAndRedactsPaths_Unit` fails in fixture creation
at `read_linux_test.go:68`: binding a Unix socket beneath the canonical Go temporary
root plus the test's long temporary-directory prefix exceeds the Unix socket
pathname boundary. The unsafe-object read assertions are not reached. The file
already imports `unix` and needs only a socket inode, not a listening service.
After the live invocation settles, the proposed test-only repair is to create
that inode directly and retain every rejection/redaction assertion, including
coverage of long valid filesystem paths. Other socket-shaped test fixtures will
be inspected for the same assumption; real service sockets retain their protocol
requirements. This proposed correction does not resolve or relabel the three
PostgreSQL failures. No executable source has changed during the invocation.


### 10.1 Scope and authority

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| 2026-10-20–21 authoring; E-013 | User authorized only the new handoff; root/tooling/backend guidance and primary owners read | Static reads and Git without optional locks; E-001/E-013 | No harness implementation authority; revalidate baseline before any later slice |

### 10.2 Backend boundary

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| Same session | Broker, pgtest/appsupport, reset and incident assertion sources inspected; none edited | E-005/E-007/E-009 | Shared infrastructure is distinct from state isolation; exclusive extension remains RB-005 |

### 10.3 Frontend boundary

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| Same session | Browser grouping, fixtures/configuration and Vitest execution inspected; none edited | E-003/E-004/E-006/E-008 | Cross-owner model is not proof of independence; RB-002/RB-003 precede S-001 |

### 10.4 Contract and codegen

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| Same session | Current browser schema v14, owner JSON and generated projections inspected as downstream data | E-001/E-004/E-011 | No schema or generator changed; future owner adoption precedes coordinated cutover |

### 10.5 Tests and harness

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| Same session | Catalogs parsed; source assertions reviewed; model recomputed in memory | E-002/E-004/E-009 | No test, benchmark, live observation or retained-run reader executed; future validation gates remain unmet |

### 10.6 Security and authorization

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| Same session | Context/session, immutable trust, physical ownership and uncertainty boundaries reviewed | E-005/E-006/E-007 | No credential, security setting or production authorization changed; future tests preserve real authentication and exact cleanup authority |

### 10.7 Open risks and next session

| Session / baseline | State and inspected/touched files | Commands / evidence | Reconciliation and next action |
| --- | --- | --- | --- |
| Same session | This document only is created; E-014 contains final review | E-011/E-012/E-013/E-014 | Prior receipt chronology qualified; current lint excludes this file; resolve only the blockers of the next explicitly authorized slice |

### 10.8 Evidence register

E-001–E-014 describe the historical creation baseline and reconciliation.
E-015–E-017 identify this review's baseline, reinspection and corrections.
Historical creation receipts retain their own reuse/reinspection limits; this
review re-read the affected sources listed in E-016. Locations below identify relevant definitions,
not exhaustive source-body inspection.

| Evidence | Source / inspection | Observation / outcome | Applicability and limits |
| --- | --- | --- | --- |
| E-001 | Root/tooling/backend instructions, framework, Core 00 owner allocation, Core 01 incident-create clauses, Testing Harness REQ-001/003/379/803–806/811–815 and AC-085–088/095 | Governing authority, isolation, grouping and preservation boundaries | Human projection-fidelity review remains separate; no document consumed by runtime |
| E-002 | 66 tools/test_families JSON files; standard-library parsing and literal identity grouping | Census in Section 2.2; no literal duplicates | Rows are not test/process counts; no dynamic/semantic equivalence proof |
| E-003 | compiler.mjs row grouping/estimation; go-lpt-shards.mjs; cohorts.mjs | Existing grouped runners; per-row estimates; deterministic packing | Costs are authored estimates, not measurements |
| E-004 | execution-topology.mjs partition builder; browser-batch-manifest.mjs; browser-cohorts.mjs; browser graph/lane compiler; metadata hashes in Section 2.3 | Owner partition and full-population 55-to-17 static comparison | No canonical plan generated; missing cross-owner effect proof remains RB-003 |
| E-005 | services_linux.go clusterKey/releaseCluster; pgtest BeginRollbackDBT/PrepareIsolatedDatabaseT; clone.go; appsupport; s3test reset | Unit-scoped exclusive lifetimes; package rollback serialization; fresh WAL-backed clones; namespace boundaries | No physical startup/drop time measured; current exclusivity is intentional owner behavior |
| E-006 | renderer_linux.go reserve/borrow/release; renderer-quiescence.mjs; browser fixtures/config | Existing renderer pool; fresh contexts/actors and absent connection-owned browsers before reuse | Physical renderer reuse is not persistent shared Chromium |
| E-007 | internal/platform/harnessruntime/reset.go; appsupport runtime and object namespace reset | Reset clears/rebuilds/verifies state; application test assembly repeats per fixture | No claim that reset is the dominant cost |
| E-008 | vitest-invocation-cli.mjs and vitest-worker-lifecycle.mjs; vite.config.ts | One coordinator with isolated file workers; explicit settlement; node/jsdom projects separate | No new daemon or worker-reuse redesign justified |
| E-009 | Incident integration and HTTP conformance test bodies, module.incidents catalog; named browser creation/preferences cases | Header-only fixture and overlapping/distinct replay assertions; survivor map in Section 7.3 | No test executed; current browser API setup simplification already exists |
| E-010 | Work-graph estimates, topology resources, snapshot-builder-cli.mjs and performance explanation schema | Existing snapshot reuse and diagnostic surface; CPU/RSS coverage limitations | No historical artifact opened and no timings inferred from estimates |
| E-011 | Makefile, task surface owner/generated binding, Markdown config and lint scripts; test-ops recipes | Public validation routes; lint omission, readiness prerequisites and output writes | No Make/lint invocation performed; exact future row discovery remains RB-006 |
| E-012 | Remediation 01/02 source text; appendix receipt, active full-7 receipt and completion-scope headings linked in Section 1.4; reinspection after concurrent edit | Appendix reports successful historical execution with unresolved causes; active continuation reports a different failing source and drops historical causal attribution from its acceptance scope | Reported source/run receipts only; retained roots were not opened; document position is not execution chronology |
| E-013 | git --no-optional-locks status --porcelain=v1 --untracked-files=all; branch/revision; ls-files --stage -z; explicit 124-file hash/identity snapshot and path lstat | Baseline in Section 1.3; initial output absent and five dirty files preserved as existing work | Snapshot lives only in session memory; not a full-workspace or executable-evidence digest |
| E-014 | Pre-write and post-write source/output comparisons; static document review | Twelve sections and 17 unique proposed requirement IDs reviewed; local links and table structure inspected; concurrent changes reconciled below | Authoring inspection only; no canonical automated validation or whole-workspace preservation certification |
| E-015 | Critical-review snapshot: existing handoff, 124 linked inputs, Git index/status and path identities | HEAD `21e9c4c15a67c69a7a97ce559653f5afadd964f4`; clean admission; identities in Section 1.3 | Bounded preservation baseline for this update; does not recreate E-013's lost in-memory snapshot |
| E-016 | Static reinspection of catalogs, browser generator/cohort packing, incident HTTP tests and row selection, relevant NLSpec clauses, lint path and remediation 02 terminal receipt | Counts and five-stage 55-to-17 static model reproduced; existing multi-symbol selection and later full-8 receipt confirmed as source text | No repository modules imported; no tests, retained results or live services opened; earlier broader audit remains source-qualified |
| E-017 | Critical-review corrections and closure comparison | Section 10.10 | Static document acceptance only; implementation readiness and measured efficiency remain unmet |

### 10.9 Historical creation-session authoring receipt

The single admitted handoff was created after confirming continued output
absence and unchanged nonsymlink ancestor identities. The resulting output is a
regular file with one hard link. This session's only filesystem mutation was
creation/revision of that file through the patch tool.

Before creation, the 124-input comparison found 121 unchanged files and three
concurrent content changes. Those files were reread; the reconciled 124-input
set matched at the immediate pre-write and first post-write comparisons.
The closure comparison then found another remediation 02 edit and a harness
specification edit. Both were reinspected. Relative to admission, the closure
snapshot therefore has 120 unchanged inputs and four changed inputs, listed
below. All four retained their device/inode, mode and link count. The logical
index, HEAD, branch and output ancestors matched admission throughout the
comparisons. These facts do not attribute another writer's changes.

This receipt is bounded to that closure snapshot. Ongoing concurrent work may
change the workspace after inspection; no durable freeze or current execution
acceptance is asserted.

| Concurrently changed input | Admission SHA-256 | Reconciled SHA-256 / effect on this handoff |
| --- | --- | --- |
| `Makefile` | `702d18654b43f13ea495158fb55fe22502b00c0dc9330c7cf09da34036065f3c` | `6b781067a4913e8efc6a0d3377c17d3ba60dcf2330a808da85cc73dd0897131e`; Go patch selection changed; Markdown invocation-path conclusion is unchanged |
| `tools/toolchain_pins.json` | `f0968e3ace0b7282a384ecb609ccc16db0e9e49c646cf0b45ba457ed4669ae6c` | `54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`; current toolchain identity differs from the historical appendix receipt; no current runtime acceptance inferred |
| `docs/handoffs/testing-harness-remediation-02.md` | `74a6557755ddb01ff81834f76da5ccba7987d23d2c271934e282bd822d2a326a` | `e535c13ce06af6da05d3ec6a4daa2d8fd13fe54e5bf8d3150334feadc834026a`; active full-7 and subsequent D-012/D-013 repair receipts reconciled with the historical appendix; no aggregate success inferred from focused repair receipts |
| `docs/testing-harness-nlspec.md` | `6fc785c967f06e8d5f96b734ba525fbe07e531b55116b8d62cf57571b8c65ace` | `7fe631f7f960c261bce31d238f9301d71e89b5e88e3ce0c98c03eb9102222ab1`; REQ-835 clarification separates optional diagnostic completeness from canonical validity; no change to this roadmap's unknown-data/performance gate |

Nine additional paths became dirty relative to the initial status:
`Makefile`, `go.mod`, `internal/platform/cryptography/execution.go`,
`tools/go-analysis/go.mod`,
`tools/harness/tests/test-work-graph-cutover.mjs`,
`tools/toolchain_pins.json`,
`tools/harness/observability/observability-check-cli.mjs`,
`tools/harness/observability/retained-checks.mjs` and
`tools/harness/observability/tests/test-instrumentation.mjs`.
The closure status snapshot therefore has fourteen modified tracked paths,
including the original five, plus this one untracked handoff. The seven
additional dirty paths outside the 124-input universe received status accounting
only; their initial contents were not snapshotted and are not certified
unchanged. No concurrent edit was reverted or incorporated as an implementation
change by this task.

Static authoring review covered all twelve required sections, 17 unique proposed
requirement IDs, local link targets and named receipt anchors, table consistency,
acyclic workflow ordering, contract/slice/validation traceability, normative
wording, default boundaries and evidence qualifications. The catalog census and
packing inputs remained unchanged across the source reconciliation. Document
review is complete as a gated roadmap; implementation discovery remains partial.

No test, build, linter, generation, finalization, performance experiment, service
operation or retained-run reader was executed. Automated repository validation
remains unperformed. Ignored outputs, active processes and the remaining
repository were not inspected, so no whole-workspace immutability or concurrent
runtime-health claim is made.

### 10.10 Critical review — 2026-10-21

This review updates only the admitted handoff. E-015 records its existing target
identity, clean starting checkout, index and 125-file preservation snapshot.
The creation receipt in Section 10.9 is retained as history, not this session's
status. The review re-read affected source and authority rather than treating
the earlier handoff as proof of current behavior.

| Review correction | Gap / contradiction resolved | Contract and acceptance consequence |
| --- | --- | --- |
| RV-001 Gate ordering | Investigations required adoption/proof they were meant to produce; RB-004 conflated protocol with candidate results | G-01–G-04 and WF-01–WF-05 separate discovery, readiness, assurance and benefit; no workflow requires its own output at entry |
| RV-002 Efficiency proof | Lifecycle reduction and “acceptable” runtime could pass without a numerical, comparable whole-workload result | Section 8.2 fixes protocol fields, uncertainty/guardrail rules, missing-data disposition, immutable failed attempts and phase/resource attribution |
| RV-003 Independent selection | S-002 froze a multi-symbol row while requiring survivor-only execution; native test results were treated as assertion traces | Section 7.3 specifies an atomic exact-symbol row split; V-004 separates source assertion mapping from native execution evidence |
| RV-004 Browser oracle | Owner-only partition equivalence was conflated with single-cohort packing; selected identity could inherit unselected roster/owner through names | R03-001–005 and Section 7.2 define exact membership, aggregation, enumeration invariance, empty/failure outcomes and clean identity cutover |
| RV-005 Lifetime safety | Pool bounds were unspecified and quarantine language implied unconditional physical retirement | R03-013–015 require finite adoption-time bounds, fresh fenced borrows, unchanged durability and retirement only through proven ownership |
| RV-006 Evidence freshness | Creation-session baseline and active failed receipt appeared current after later completed-source execution | E-015/E-016 and Section 1.4 attribute full-8 success, full-7 failure and unresolved legacy ownership separately; no current run certification |

The scope/authority update is E-015; backend and frontend boundary review is
E-016; contract/codegen, test/harness and security consequences are RV-001–005.
Only document prose/tables changed. No product code, catalog, schema, spec or
public command contract was edited. Open-risk ownership remains in Section 11.

Closure receipt: the pre-write comparison matched all 125 recorded files and
the admitted output identity. After editing, all 124 non-output inputs retained
their content hashes and filesystem identities; HEAD, branch, logical index and
ancestor identities were unchanged. Git status showed only this handoff modified,
with no staged or untracked changes. The output remained a regular single-link
file at the admitted identity. This is a bounded preservation result, not a scan
of ignored files or active processes.

Static reread covered the twelve required sections, 17 R03 requirements, four
stage gates, workflow dependencies, 136 local links/receipt anchors, table
structure and all document-local references. The S-008 mention is explicitly
remediation 02's historical slice, not an undefined slice of this handoff.
The catalog census and static packing model were independently recomputed.
No tests, builds, Make targets, linter, generation, finalization, service control,
retained-run reader or performance experiment ran. Automated repository
validation is unperformed; Markdown lint still excludes this handoff and its
Make path has non-document write effects.

**E-100 — Experiment 02 CI warm-up terminal failure; WF-11b registered.**
E-099 completed in native **90641/2**, graph **3,618,182 ms**, with **596 passed /
4 failed / 600 units**. Source B2 and the E-099 graph are unchanged. Original
failures remain in the exact root
`.cartulary/test-results/r03-s011-exp02-ci-warm-b`. There are 1,185 completed
cleanup records and one failed `host_release` (sequence 514) for
`go:f0aa161d0716c72f:331ab287cb5f`. Final provider/service/runtime closes completed.
The referenced service scope now records physical settlement completed and cleanup
succeeded at `2026-10-22T08:36:39.763639918Z`; this later settlement does not erase
the earlier cleanup failure. The final retained-secret scan passed 10,242 files.

Read-only `make explain-run RESULTS_DIR=<exact-root>` succeeded. Independent
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=events`
(**61930/0**) and `CHECK=timing` (**83233/0**) passed with `full_validity=0`.
`CHECK=diagnostics` (**37756/2**) failed artifact admission; performance and
critical-path exports (**19395/2**, **32432/2**) also failed because the required
invocation-accounting receipt is absent. This is the existing fail-closed behavior:
`harnessdiag.publishInvocationAccounting` rejects any non-completed cleanup record.
No complete resource claim or acceptance sample is available. Empty failed export
files MUST NOT be used as validated measurements.

Scoped terminal PostgreSQL observations (not a full diagnostics audit) show the
incident-bundle generation's checkpoint synchronizing 2,518 files in 25,490 ms,
and recovery's generation synchronizing 4,050 files in 27,782 ms. Both exceed the
20-second forced-drop allowance and explain the immediate checkpoint wait. They
do not establish the underlying storage cause. The separate readiness failure
recorded 60 connection resets, no internal listener, only `initdb_started`, and no
first SQL connection within its 15-second SQL phase. Container OOM was false.
No deadline, durability, retry, capacity, isolation or pooling change is authorized
by these observations.

WF-11b is a serial correctness prerequisite within WF-11. Its first correction is
test-only: replace listening Unix sockets used solely as unsafe filesystem objects
with socket inodes, and exercise valid long filesystem paths. Owner
`platform.rootedfs` routes `make test-slice OWNER=platform.rootedfs` (task-guide
**57691/0**). Preserve every security rejection and redaction assertion. This
requires no product-behavior amendment. The second investigation traces PostgreSQL
clone/checkpoint work and startup/settlement ownership before selecting a repair;
a later isolated pass alone cannot close the original CI failure. Any behavioral
repair MUST adopt its NLSpec owner first, preserve the stated scope constraints,
and be applied equally to baseline and candidate. Exit requires affected focused
assurance, explicit physical-settlement evidence, a common repaired-source freeze,
and tracker lint before sampling. Failed experiment 02 and all original sources
remain immutable historical evidence; no lifecycle acceptance or S-007e completion
is claimed. `RESULTS_DIR` remains unset for retained-run maintenance.

**E-101 — Socket-object correction and focused verification.**
WF-11b entry lint `r03-s011-corrective2-entry-lint-01` passed
(**97062/0**, **9,500 ms**) before source edits. Both
`internal/platform/securefile/read_linux_test.go` and
`internal/platform/rootedfs/root_linux_test.go` now create socket inodes with
`unix.Mknod` instead of binding listeners used only as filesystem objects.
Deliberately long valid directory names exercise the actual filesystem boundary;
all symlink, FIFO, socket, containment and redaction checks remain. Real control
socket listeners are unchanged. No behavioral owner or generated contract changed.

`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=r03-s011-corrective2-rootedfs-01 make test-slice
OWNER=platform.rootedfs HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`
passed (**76357/0**, **4,894 ms**, 2/2 units), source
`sha256:2dd3d784ea08c6c7b2286275dda4580b8cbe2a1e476b73a6a13aa033677c5a34`.
Independent `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`
passed, `full_validity=1`; all 11 cleanup records completed and required accounting
settled empty. This closes the socket fixture defect for the focused baseline
source, not the PostgreSQL or full-CI acceptance gate.

Three selection-preflight mistakes are preserved: `r03-s011-corrective2-postgres-01`
used ROWS without OWNER/SUITE; `make target-plan TARGET=test-slice ... ROWS=...`
used an undeclared selector; `r03-s011-corrective2-postgres-02` combined SUITE and
ROWS without OWNER. All exited 2 before graph execution. Exact-row selection
requires OWNER; the affected rows therefore run as serial owner-specific slices,
not a fabricated combined command. Owner guides for `module.incidentbundles`,
`module.recovery` and `platform.postgres` passed (**47228/0**). These preflight
attempts establish no metric or cleanup result and are not acceptance samples.

**E-102 — Affected PostgreSQL assurance; unchanged lifecycle contract.**
Three serial exact-row commands used `CGO_ENABLED=0`, execution-result reuse off,
`HARNESS_DIAGNOSTICS=basic`, `HARNESS_ACCOUNTING=required`, and fresh run IDs:

| Run suffix under `.cartulary/test-results/` | Owner and exact row | Native / graph duration | Assurance and settlement |
| --- | --- | --- | --- |
| `r03-s011-corrective2-bundles-01` | `module.incidentbundles`; `module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413` | 2205/0; 25,808 ms | 3/3 units; 14 completed cleanup records; complete audit; accounting empty |
| `r03-s011-corrective2-recovery-01` | `module.recovery`; `module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085` | 17802/0; 31,535 ms | 3/3 units; 14 completed cleanup records; complete audit; accounting empty |
| `r03-s011-corrective2-admission-01` | `platform.postgres`; `platform.postgres.integration.fresh_fixture_admission` | 4191/0; 10,485 ms | 3/3 units; 14 completed cleanup records; complete audit; accounting empty |

Each command was `make test-slice OWNER=<table-owner> ROWS=<table-row>` with the
settings above. Each independent `make harness-observability-check
RESULTS_DIR=<exact-root> CHECK=all` exited 0 with `full_validity=1`, complete
execution/resource diagnostics, available required accounting and observed
PostgreSQL checkpoints. Sources match E-101. These focused passes establish their
selected assertions and settlement; they do not erase E-099 or establish that the
containing CI workload is reliable under load. The underlying storage/startup
latency cause remains unproved.

The implementation trace supports **NO_CHANGE to PostgreSQL mechanics**:
REQ-804's WAL-backed clones preserve durability without whole-cluster per-clone
checkpoints; Section 11's one bounded forced drop rejects uncertain settlement;
the fixture poison prevents subsequent reuse; terminal `Production.close` drains
retained partial acquisitions through the shared cleanup evaluator. E-099's final
service settlement is therefore compatible with its earlier failed release. The
accounting owner correctly refuses a complete receipt after that failure. No
reader relaxation, deadline increase, checkpoint injection, retry or concurrency
change is made. The required full-CI gate remains open for the newly frozen source;
this disposition does not close that gate by deferral.

Baseline `make agent-finalize`, run `r03-s011-corrective2-finalize-b-01`, passed
(**73296/0**, **27,048 ms**); retained-run maintenance was skipped because
`RESULTS_DIR` was unset. Both source snapshots MUST include the identical two-file
socket-fixture correction before a new experiment starts. Candidate preparation
and source-preservation receipts follow; WF-11b is not yet closed here.

**E-103 — WF-11b correction complete in both treatments.**
Candidate regeneration `r03-s011-corrective2-generate-c-01` passed
(**38556/0**, **10,810 ms**). `make agent-finalize`, run
`r03-s011-corrective2-finalize-c-01`, passed (**83441/0**, **26,752 ms**), with
retained-run maintenance skipped because `RESULTS_DIR` was unset. Unaccounted
`make source-check CHECK=catalog`, run `r03-s011-corrective2-prepare-c-01`,
passed (**47175/0**, **1,693 ms**) before required-accounting execution.
The E-101 full `platform.rootedfs` selection passed on candidate source in
`r03-s011-corrective2-rootedfs-c-01` (**47646/0**, **4,403 ms**), with independent
`CHECK=all` exit 0, `full_validity=1`, complete diagnostics and empty accounting.

Immutable internal preservation snapshots:

| Treatment | Source digest | Archive SHA-256 |
| --- | --- | --- |
| B3 diagnostics + original separate resets + common repairs | `sha256:2dd3d784ea08c6c7b2286275dda4580b8cbe2a1e476b73a6a13aa033677c5a34` | `a1c1786d99c45c64507f3c6e141cb3ac498a6f3b46ae9e8d41ac63d6c2571f77` |
| C3 identical diagnostics/common repairs + S-010 lifecycle | `sha256:cacf38201c64a9850a990637232cbf01c3eda5005ce3618909245d8a24c39945` | `bfd0e51a7b797df2bb57646c06dee387b8237ede9a626ffab823baee4ced3368` |

The archives are `diagnostic-baseline-source-exp03.tar.gz` and
`lifecycle-candidate-source-exp03.tar.gz` under
`.cartulary/remediation03/continuation/`, with adjacent per-file manifests.
Fingerprint comparison against B2/C2 finds exactly the same two authored test-file
changes, excluding the controlling tracker. HEAD remains `f42fc00b7`; the tracker
is preserved across restoration. Original and experiment 02 archives are intact.
Candidate public browser/CI plan exports (**91047/0**, **95259/0**) are exactly
equal to C2, so this repair does not alter selected populations or scheduling.

Assurance decision: socket-object correction accepted on both sources; PostgreSQL
mechanics NO_CHANGE under E-102. Performance decision: none, zero matched samples.
The failed CI warm-up remains failed and its containing assurance gate remains
open. WF-11b completes only this correction/diagnosis slice. WF-11 resumes after
this receipt passes Markdown lint, with a new exact freeze and four newly declared
excluded warm-ups. Every required check and the final CI gate still applies.

**E-104 — Experiment 03 frozen after the required correction.**
WF-11b closing Markdown lint passed in `r03-s011-corrective2-close-lint-01`
(**73652/0**, **15,780 ms**) before WF-11 resumed. Baseline restoration from E-103's
immutable archive, public `make generate` and unaccounted
`make source-check CHECK=catalog` passed in native **41010/0**, respective runs
`r03-s011-exp03-b-generate-01` (**10,910 ms**) and
`r03-s011-exp03-b-prepare-01` (**1,658 ms**). Per-file source checks match B3,
excluding the intentionally preserved tracker. Public B3 plan exports
(**6571/0**, **7901/0**) match B2 exactly, as C3 matched C2 in E-103.

E-097's workload populations, original reset-to-preparation treatment map, four-pair
maximum, odd/even order, thresholds, bootstrap definition, metric scopes, matching
diagnostics and strict cleanup/assurance requirements apply unchanged. Sources are
B3/C3 from E-103; kernel, toolchain, repository/storage identities and machine Go
paths are revalidated unchanged. Neither PostgreSQL behavior nor selected graph
populations changed in the common repair. The frozen ledger is
`.cartulary/remediation03/continuation/performance-experiment-exp03-freeze.json`;
its mutable observation ledger remains `performance-experiment.json`.
Experiment 02 is archived as `performance-experiment-exp02.json`, including every
failed attempt and omitted unstarted sample. No prior run is relabeled current.

New excluded warm-ups are declared in order:
`r03-s011-exp03-browser-warm-b`, `r03-s011-exp03-ci-warm-b`,
`r03-s011-exp03-browser-warm-c`, `r03-s011-exp03-ci-warm-c`. Each invokes its E-097
public command with `CGO_ENABLED=0`, `CARTULARY_HARNESS_CACHE_MODE=off`, and explicit
Make selectors `HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
All four MUST pass independent evidence audits, physical settlement and source/
runtime-graph binding before acceptance sampling begins. Planned matched pairs
retain the same order with the `r03-s011-exp03` prefix. No comparison uses
`COMPARISON=equivalent` across sources. Source repair requires another freeze.

Assurance: focused corrective checks pass; containing full-CI gate open.
Performance: zero acceptance samples, no benefit or guardrail conclusion.
Required failure remains a stop condition. Freeze Markdown lint `r03-s011-exp03-freeze-lint-01` passed
(**37284/0**, **9,570 ms**). The first declared excluded browser warm-up is active
in native **89053**; source is fixed until terminal settlement.

**E-105 — Experiment 03 baseline browser warm-up passed.**
The E-104 command `r03-s011-exp03-browser-warm-b` completed in native **89053/0**,
137/137 units, graph **1,160,834 ms**. Published source is B3; browser runtime graph
matches the frozen identity. Independent `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/r03-s011-exp03-browser-warm-b CHECK=all`
passed with `full_validity=1`, complete execution/resource diagnostics and zero
omissions. All 276 cleanup records completed, required accounting settled empty,
and the final retained-secret scan passed.

All four public exports passed: `make explain-run RESULTS_DIR=<exact-root>
DETAIL=performance JSON=1`, and `DETAIL=critical-path VIEW=<units|waits|allocations>
OFFSET=0 LIMIT=100 JSON=1`. Unit and wait pages each contain all 80 records with
`next_offset=null`; the path allocation page contains one record. Complete path
totals are 21 ms graph queue and 1,160,297 ms lifetime. Full validated lifecycle
coverage records four stack generations, each retired for terminal borrower
release and settled; zero reset-only generations.

Excluded warm-up metrics are owning-command elapsed **1,173,437.902156 ms**, charged
CPU **3,981,379,935 microseconds**, kernel peak memory **12,035,534,848 bytes**,
p95 ready-to-start **1 ms**, and cleanup union **47,166 ms**. These are scoped
observations, not comparative benefits or guardrail acceptance. Exact exports and
lifecycle records are retained under the matching run-ID prefix in
`.cartulary/remediation03/continuation/`. Source remained fixed during execution.
The baseline CI warm-up is active in native **60168**, using the E-104 command
and settings. Matched samples remain unstarted; source is fixed until terminal
settlement.

**E-106 — Experiment 03 baseline CI warm-up passed and independently audited.**
`r03-s011-exp03-ci-warm-b` completed in native **60168/0**: **600/600 units**,
graph **3,546,359 ms**, source B3 and expected runtime graph
`sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
Post-run fingerprints match the immutable B3 snapshot except the intentionally
preserved tracker. No source change or competing verification ran during CI.
Structured observation stopped on its explicit quiet-measurement notification;
only native-session observation continued until exit.

Independent `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`
passed (**36897/0**, `full_validity=1`). All **1,164** cleanup records completed;
retained-secret scanning passed **9,806 files**; required accounting settled empty.
Optional verbose execution and sampled-resource diagnostics are truncated
(resource omissions **1,295**), and PostgreSQL checkpoint coverage is unavailable.
These limitations remain explicit. Canonical critical-path coverage and the
separately bounded allocation-lifecycle section are complete with zero omissions;
required kernel accounting is available. Sampled resource coverage does not
supply or replace the kernel CPU/memory guardrails.

Public performance and critical-path exports all passed. Both root CI and exact
target `browser-e2e-webserver-backed` produce an 80-unit path, **657,373 ms** graph
queue and **2,887,500 ms** unit lifetime. Unit and allocation views are complete
on their first 100-record page; the wait view contains **345** segments, retrieved
at offsets **0, 100, 200, 300** with page counts **100, 100, 100, 45** and terminal
`next_offset=null`. Root exports used sessions **84604/0**, **91264/0**,
**16632/0**, **69408/0**; browser-target exports **63198/0**, **12236/0**,
**5414/0**; remaining wait pages **15113/0**, **49369/0**, **73899/0**.
The complete validated lifecycle section records **86** stack generations:
**73 capacity-pressure** retirements and **13 terminal-borrower-release**
retirements, all settled. **32** generations have only standalone reset borrowers.
This corroborates the diagnostic motivation without establishing candidate benefit.

Excluded warm-up metrics: owning-command elapsed **3,631,644.687354 ms**, charged
CPU **11,693,896,365 microseconds**, kernel peak memory **11,403,444,224 bytes**,
p95 ready-to-start **3,109,931 ms**, cleanup union **149,995 ms**. The p95 population
contains all 600 units; it MUST NOT be presented as a browser-only percentile.
Supplementary cohort comparisons use matching cohort identities in complete
selected-target path exports, report their coverage explicitly, and do not imply
a percentile over unobserved non-path cohorts. Phases and simultaneous blockers
remain nonexclusive observations.

Assurance decision: the repaired baseline's containing warm-up passes; E-099's
older failure and unproved underlying latency cause remain preserved. Performance
decision: none; both baseline runs are excluded warm-ups, and no matched sample
exists. Baseline receipt lint `r03-s011-exp03-baseline-receipt-lint-01` passed
(**29377/0**, **8,980 ms**). Exact C3 restoration, public regeneration and
unaccounted controller preparation passed in native **64045/0**: runs
`r03-s011-exp03-c-generate-01` (**10,710 ms**) and
`r03-s011-exp03-c-prepare-01` (**1,644 ms**). Fingerprints match C3 except the
preserved tracker. C3 browser warm-up `r03-s011-exp03-browser-warm-c` is active in native
**90738** using E-104 settings. Candidate CI follows only after its complete audit. No final
S-007e or lifecycle acceptance is claimed. Maintenance remains skipped because
`RESULTS_DIR` was unset.

**E-107 — Candidate browser warm-up and same-borrow continuity passed.**
`r03-s011-exp03-browser-warm-c` completed in native **90738/0**, **76/76 units**,
graph **1,141,535 ms**, source C3 and frozen graph
`sha256:30890f4b065d42f7753b88b77e968489010abdde3d7d98fb5f9c1146dc8bb4d5`.
Post-run fingerprints match the immutable C3 snapshot except the tracker.
Independent `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`
passed with `full_validity=1`, complete execution/resource diagnostics and zero
omissions. All **154** cleanup records completed; the final retained-secret scan
passed **2,235 files**; required accounting settled empty.

All four public performance/critical-path exports exited 0. The complete candidate
path has **43 units**, **9 ms** graph queue and **1,140,994 ms** lifetime; unit/wait
pages each contain all 43 records and allocation view one record, each exhausted
with `next_offset=null`. Full lifecycle coverage records four generations, each
settled by terminal borrower release, with zero reset-only generations.

Post-audit correspondence over declared typed artifact references confirms all
**61** preparation steps passed, every owning unit records `runner_started=true`,
and every reset/cohort pair agrees on unit, fixture lease, allocation, browser
session and resulting backend generation. Each reset increments its prior
backend generation exactly once. The independent arithmetic correspondence report
`r03-s011-exp03-browser-warm-c-continuity.json` records every unit/step and both
artifact references under `.cartulary/remediation03/continuation/`; it neither
parses canonical events nor replaces the existing evidence validator. Original
receipts remain under the exact run root.

Excluded warm-up metrics: owning-command elapsed **1,153,646.275917 ms**, charged
CPU **3,973,884,438 microseconds**, kernel peak memory **8,939,888,640 bytes**,
p95 ready-to-start **1 ms**, cleanup union **47,622 ms**. These warm-ups do not
establish comparative acceptance or satisfy either first-two-pair decision rule.
Candidate CI warm-up `r03-s011-exp03-ci-warm-c` is active in native **99501** on
unchanged C3 with E-104 settings. No matched samples have started, and final
S-007e validation remains pending.

**E-108 — Candidate CI warm-up readiness failure; terminal settlement pending.**
Active native **99501** has published one failed unit,
`go:74dd976f6fd54aa7:ca4cfe5a862e`, source C3 and runtime graph
`sha256:851e640604871435b8ad905285055395671fb93be82112295370ca9ddc5b911d`.
The graph has 517 units as frozen. The affected unit owns eight `module.revisions`
store rows, has an empty preparation array, and records `runner_started=false`.
All eight rows retain `infra/service_readiness_timeout`; no product assertion was
launched. The exact service reference is
`_shared/test-services/53cbbf5cb0224ff19582adae/service-scope.json` under
`.cartulary/test-results/r03-s011-exp03-ci-warm-c`.

The typed diagnostic records one PostgreSQL acquisition attempt: container start
1,196 ms, mapped endpoint resolution 12 ms, then a 15,001 ms expired SQL-readiness
phase. There were 61 probes (60 connection resets and one probe timeout), no first
SQL connection, no internal listener at failure inspection, and no container OOM.
Scoped Docker-wall startup markers place initialization completion at 14,856 ms,
temporary bootstrap-server readiness at 15,043 ms, and bootstrap shutdown at
15,467 ms after container start. No final-server ready marker was observed.
These log observations are descriptive and are not silently aligned with the
canonical monotonic clock. The initial cleanup receipt is failed/unresolved at
`2026-10-22T10:53:15.080363431Z`; final settlement MUST be checked after native exit.
Independent work continues under the existing policy; no cancellation, retry or
source mutation has been introduced in response.

The existing implementation faithfully projects Section 11.4's nested limits:
35 seconds per startup attempt, but a separate 15-second SQL-readiness window
begins immediately after mapped-port resolution. A mapped port is available before
PostgreSQL bootstrap finishes. This failure identifies a specification/design
question about competing startup-budget ownership, not evidence that browser
preparation executed in the failed unit. Investigation will evaluate removing
the redundant inner readiness window in favor of the existing bounded parent
attempt, while preserving the outer attempt cap, five-second probe cap, polling,
terminal capability rejection, no replacement after readiness begins, and fresh
bounded failed-start cleanup. This is a proposal, not an adopted owner amendment.
No deadline or implementation has changed. Any accepted repair requires its own
serial corrective workstream, owner adoption, contract/projection validation,
application to both treatments and a new frozen experiment. Existing failed and
omitted attempts remain retained. Performance sampling is blocked; no matched
sample or lifecycle acceptance is claimed.

**E-109 — Second candidate CI failure; two serial corrective workstreams registered.**
Native **99501** remains active; the last published status has 308 passed, two
failed, seven running and 200 pending units at 3,079,362 ms. This is an observation,
not a terminal outcome. Source remains the frozen C3; no executable source or
adopted NLSpec has changed during the run. Drafts are confined to ignored internal
authoring storage and are neither adopted contracts nor execution inputs.

The second failed unit is `go:9c49060badc1c62b:b910b821328b`, exact row
`harness.browser.integration.native_nested_cancellation`. Its product runner
started; its three-minute observation deadline failed with `nested service did
not dispatch`. The first nested `summary` command was cancelled by that test.
Published child cleanup has nine completed records. The public command
`make explain-run RESULTS_DIR=.cartulary/test-results/r03-s011-exp03-ci-warm-c/nested-cancellation/summary
DETAIL=critical-path VIEW=units LIMIT=100 JSON=1` exited **0** and reconstructs one
path unit, `operation:harness.testservices_build`: 174,620 ms lifetime, 174,616 ms
host admission, zero graph queue, no runner launch and 4 ms uncovered lifetime.
This diagnostic read is not a passing assurance audit. The exported response is
retained as internal `r03-s011-exp03-ci-warm-c-nested-cancellation-units-000.json`.

Source comparison proves that S-010's protected-environment loop deletes
`CARTULARY_HOST_ADMISSION_LEASE` inherited through the native executor's base
context when `executionRequest.Host` is absent. The installed nested-cancellation
fixture intentionally executes beneath its enclosing lifecycle without a new
local host lease. Its existing CPU budget survives while its paired parent lease
is lost. The child therefore requests unrelated host capacity before the intended
cancellation boundary. The baseline has no such deletion. F-012/R03-025/V-013 and
T-020 register **WF-11c**, a candidate-only corrective slice: preserve the original
parent lease/budget pair, let a locally admitted binding take precedence, reject
fixture/command replacement, and continue to exclude parent fixture identities
from child fixture ownership. The three-minute test deadline, scheduler policy,
capacities and fairness SHALL remain unchanged.

F-013/R03-026/V-014 and T-021 register **WF-11d**, after WF-11c. E-108's startup
proposal remains unadopted until that slice. Inspection of pinned testcontainers
v0.42.0 additionally confirms a default ten-second graceful stop before removal,
while failed PostgreSQL startup supplies ten seconds for all physical cleanup.
That is a concrete competing-budget defect; it does not prove the exclusive
cause of this run's unresolved initial cleanup. The proposed correction uses the
existing bounded startup-attempt deadline for readiness and zero graceful-stop
allowance only for an exact unpublished failed-start allocation with no borrower.
It preserves the 35-second attempt cap, five-second probe cap, polling, retry
exclusions, ten-second failed-start cleanup cap, ordinary published-harness
shutdown and all durability boundaries. Diagnostic v5 replaces the retired SQL
window field with existing phase-budget attribution; no translation fallback is
proposed. This correction SHALL be common to both performance treatments.

Each corrective slice SHALL adopt its owner clarification before code/projections,
run its mapped assurance, update this tracker, and pass Markdown lint before the
next slice. Current native completion and exact physical settlement are entry
conditions. Both failures and all omitted attempts SHALL remain recorded. A new
source freeze and excluded warm-ups are required before matched samples. The
internal arithmetic helper now rejects duplicate pair treatments, source mismatch,
nonfinite metrics and missing successful exports; thresholds, pair count and
bootstrap arithmetic are unchanged. No performance acceptance is claimed.

**E-110 — Experiment 03 failed warm-up settled; WF-11c entry.**
`r03-s011-exp03-ci-warm-c` ended in native **99501/2** after **3,893,201 ms**
graph duration, with **515 passed / 2 failed / 517 units**. The owning terminal
receipt reports **3,968,360 ms** outer duration and failed outcome. That receipt
is descriptive wrapper timing, not a replacement for missing required invocation
accounting. The two failures are exactly E-108/E-109; no matched acceptance sample
was admitted. Fingerprint comparison against the immutable C3 archive found zero
source differences except this controlling tracker.

Final cleanup contains **967 completed and one failed** record. The failed record
is sequence 67, `host_release` for E-108's revisions unit, with `cleanup_error`
and `host_release/unresolved`; it remains preserved. The exact failed service's
final scope now records `physical_settlement=completed`, `status=succeeded` at
`2026-10-22T11:54:48.323453048Z`. Final provider/service/runtime closes completed.
All 58 browser allocation records report settled, with complete lifecycle coverage
and zero omissions. These are retained producer observations, not a successful
full diagnostic audit. No separate recovery or ordinary cleanup retry was invoked.
The terminal secret scan passed over **9,173 files**.

Public `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/r03-s011-exp03-ci-warm-c` with `CHECK=events`
(**52953/0**) and `CHECK=timing` (**6320/0**) passed with `full_validity=0`.
`CHECK=diagnostics` (**63989/2**) failed artifact validation. A public critical-path
export (`DETAIL=critical-path VIEW=units LIMIT=100 JSON=1`, **51021/2**) also rejects
missing `diagnostics/invocation-accounting.json`. The invocation-accounting owner
requires every cleanup record to be completed; later physical settlement does not
erase the earlier failed host release. The required receipt was therefore withheld.
Whole-invocation charged resources and kernel settlement coverage are unavailable;
no complete resource guarantee or qualifying timing comparison is claimed.

The final experiment ledger is preserved as internal
`performance-experiment-exp03.json`; its original freeze remains unchanged.
All planned samples are explicitly omitted after required failure and impending
source repair. Earlier successful warm-ups remain source-bound and excluded.
WF-11c now begins with public entry lint
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-corrective3-entry-lint-01
make lint-markdown`. Its candidate-only owner clarification is adopted only after
that boundary passes. WF-11d then applies the common startup correction; each
slice has independent assurance, tracker and lint closure. S-011 performance and
S-007e final validation remain open.

**E-111 — WF-11c owner adoption and candidate repair.**
Entry Markdown lint passed in `r03-s011-corrective3-entry-lint-01`
(**69447/0**, **8,920 ms**) before implementation. REQ-858 now explicitly maps
local, inherited and absent host admission to their permitted lease/CPU bindings,
with REQ-102 retaining validation and worker-budget authority. Local admission
wins; without a new local lease the original inherited parent pair survives;
fixture/command environments cannot manufacture either binding. Parent fixture
identities remain excluded from child fixture ownership. No schedule, capacity,
fairness, retry, timeout or isolation policy changes.

The owner amendment precedes the process-executor and native assurance changes.
This candidate-only repair maps F-012/R03-025 to `process_linux.go`, the protected
fixture-binding list and the existing `TestRuntimeHostAdmission`/
`TestRuntimeProcessLifecycle` symbols. V-013 requires their native-runtime row
and the exact installed nested-cancellation row before closure. Source identity,
commands, retained outcomes and settlement will follow. PostgreSQL drafts remain
unapplied; WF-11d has not started. No lifecycle performance acceptance is claimed.

**E-112 — WF-11c complete; inherited admission preserved without policy changes.**
The final corrective source is
`sha256:8416b847d0f96505edca171526f3211c35fbdf4c26e5d6065b7bd312743367c9`.
Compared with C3 it changes only the NLSpec, the protected fixture-binding list,
native process execution, its unit tests and the installed nested-cancellation
assertions. Native execution now preserves an inherited parent admission pair
when no local lease exists; a real local admission wins; fixture and command
values cannot create or replace admission. Parent fixture identities remain
excluded. REQ-858's precedence table and REQ-102 map to these changes and V-013.
No scheduler, capacity, deadline, retry, worker default, isolation or pooling
policy changed.

All commands below ran from the repository root. Accounted tests used
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off`, fresh `CARTULARY_TEST_RUN_ID`, and
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Run roots are
`.cartulary/test-results/<run-id>`.

| Run ID | Public command / exact selection | Native / outer exit | Graph or tool duration |
| --- | --- | --- | ---: |
| `r03-s011-corrective3-format-01` | `make format` | 41923 / 0 | 7,911 ms |
| `r03-s011-corrective3-prepare-01` | `make source-check CHECK=catalog`, unaccounted | 46682 / 0 | 1,646 ms |
| `r03-s011-corrective3-native-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` | 7494 / 0 | 54,681 ms |
| `r03-s011-corrective3-nested-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_nested_cancellation` | 40936 / 0 | 42,768 ms |
| `r03-s011-corrective3-finalize-01` | `make agent-finalize` | 39766 / 0 | 26,569 ms |
| `r03-s011-corrective3-nested-02` | Same exact nested-cancellation row, with explicit inherited-budget assertion | 28651 / 0 | 47,492 ms |
| `r03-s011-corrective3-finalize-02` | `make agent-finalize`, before containing verification | 52289 / 0 | 27,028 ms |
| `r03-s011-corrective3-containing-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_nested_cancellation,harness.browser.unit.native_fixtures,harness.browser.unit.httptestx_helper_contract,harness.browser.unit.testcontainersx_helper_contract` | 54370 / 0 | 47,809 ms |

Task guides for `harness.command_surface` and `harness.browser` passed
**18766/0** and **84891/0**. The native and first nested checks bind intermediate
source `sha256:817877bfcdf5f98734649cf4029624506bfeeaa1ab8fed32e8718cc9377c936a`;
the final two checks add only the installed test's explicit parent-presence and
retained-budget assertions and bind the final source above. Full public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` passed
independently for all four accounted tests, with `full_validity=1`, complete
execution/resource diagnostics and complete required accounting. Both finalize
runs skipped retained-run maintenance because `RESULTS_DIR` was unset; no
compatible successful full warm-check root was available.

The native row passed 1/1 units with ten completed cleanup records and a passing
77-file secret scan. The first exact nested row passed 1/1, ten completed cleanup
records and a passing 124-file scan. The final exact nested row passed 1/1, ten
completed cleanup records and a passing 127-file scan. The containing selection
passed 4/4, thirteen completed cleanup records and a passing 134-file scan.
All four required accounting receipts report `settlement=empty`.

The focused row legitimately receives a 19-token enclosing reservation; its child
manifests retain 19. The ordinary four-row containing selection gives the enclosing
unit four tokens; all three child modes (`summary`, `machine`, `summary-startup`)
retain four. Thus the proof does not rely on an unconstrained focused pass or an
artificial capacity override. Each mode reaches its intended cancellation boundary,
records cancelled terminal state, passes its secret scan and has only completed
cleanup (14, 14 and 12 records respectively). These checks preserve cancellation
and startup settlement assurance; their durations are not performance samples.

The candidate-only correction is preserved as internal
`lifecycle-candidate-source-corrective3.tar.gz` / `.json`, 7,507 files, archive SHA-256
`5d7b74defee13d5499a5f7029b6fb946528d43ecfa2914b5bba907536240d2c3`, validated by the
containing run. This intermediate snapshot is called C3a and permits exact rollback
or later application of the common startup correction. WF-11c is DONE for assurance;
lifecycle performance remains pending. WF-11d begins only after the closing public
lint `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-corrective3-close-lint-01
make lint-markdown` passes. The failed C3 run and its incomplete accounting remain
unchanged.

**E-113 — WF-11d owner adoption; common startup correction.**
WF-11c closing Markdown lint passed in `r03-s011-corrective3-close-lint-01`
(**83191/0**, **9,560 ms**). Exact guarded restoration from C3a to B3 preserved the
tracker and restored the baseline's original separate-reset owner/projections/code.
C3a remains immutable. Section 11.4 was adopted on this baseline before changing
its implementation or projections; the corresponding Section 11.5 retry table
and AC-007 acceptance route were reconciled.

F-013/R03-026/V-014 now require one attempt-owned SQL-readiness budget within the
unchanged 35-second attempt and 180-second outer startup caps. Endpoint resolution
retains its 15-second cap; polling remains 250 ms and each SQL/security probe stays
bounded by five seconds and the earlier parent deadline. Missing or unbounded
readiness callers fail before probing; already cancelled/expired callers preserve
that outcome. Authentication/capability rejection remains terminal and no
replacement startup is permitted after SQL readiness begins.

For an exact failed-start allocation before ready publication and with no admitted
borrower, physical termination/removal receives the existing fresh ten-second
cleanup context with zero graceful-stop allowance. Published harness shutdown,
borrower/database durability, retries, capacities, admission and recovery authority
are unchanged. Uncertain cleanup retains primary/secondary failure distinctions
and exact ownership. Diagnostic v5 removes `sql_window_ms`, declares phase-budget
semantics in a mapping table, and rejects a SQL phase claiming an independent cap.
Current readers reject retired versions without a translation path.

The implementation maps to `pgtest` readiness/startup cleanup, its existing test
symbols, the typed `suiteservices` diagnostic and schema attachments. Generated
projections will be regenerated through Make. Existing catalog routes cover
platform PostgreSQL fixture lifecycle, browser pgtest and service-lifecycle
helpers, evidence-accounting helpers and the exact affected revisions store rows.
The correction SHALL be applied identically to B3 and C3a before experiment 04 is
frozen. This is a correctness/specification correction, not performance acceptance
or a claim that the host's startup latency cause has been proved.

**E-114 — WF-11d baseline projection and first focused attempt.**
B3 owner adoption preceded authored startup/diagnostic changes. Public generation
`r03-s011-corrective4-generate-b-01` passed **52848/0**, **10,870 ms**; format
`r03-s011-corrective4-format-b-01` passed **54252/0**, **7,844 ms**; unaccounted
`source-check CHECK=catalog` preparation `r03-s011-corrective4-prepare-b-01`
passed **29484/0**, **1,750 ms**. Task guides passed for `platform.postgres`
(**36745/0**), `harness.evidence_accounting` (**22678/0**) and `module.revisions`
(**23003/0**), confirming the existing exact-row routes.

The first accounted command, `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s011-corrective4-postgres-b-01
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=platform.postgres
ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle,platform.postgres.integration.fresh_fixture_admission
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`, exited **4835/2** after
**11,550 ms**. Source was
`sha256:8283cbd1a868787656a99d58f93a635750fb5875782bb2d14862e9f45fb14502`.
Two setup units passed and both selected Go units failed compilation before test
execution: one old assertion still referenced the removed `SQLWindowMS` field.
The harness retains `infra/preflight_error`; the compiler diagnostic identifies
the precise stale projection. This does not establish a product assertion failure.
The assertion is corrected to require the v5 schema identity while retaining
poll/probe checks; separate new tests reject the retired field and contradictory
phase budgets. No runtime deadline or acceptance rule changed in response.

All fifteen cleanup records completed, required accounting reports empty
settlement, and the secret scan passed over 35 files. Public
`make harness-observability-check
RESULTS_DIR=.cartulary/test-results/r03-s011-corrective4-postgres-b-01
CHECK=diagnostics` passed **0**, `full_validity=0`, with complete diagnostics and
required accounting. This failed development attempt remains retained; the next
focused attempt uses a fresh run ID and source. WF-11d remains IN_PROGRESS.

**E-115 — Startup owner default discovered and clarified before implementation.**
The second exact baseline PostgreSQL selection,
`r03-s011-corrective4-postgres-b-02`, exited **87631/2** after **11,501 ms**.
Compilation succeeded and the real fresh-fixture row passed. The fixture-lifecycle
unit failed two new/affected assertions: diagnostic SQL budget had no parent, and
capability validation could not launch a probe without a bounded parent. Source
inspection identifies the gap: direct `StartOwned` callers supplied empty options,
while only suite composition passed the 35-second `AttemptTimeout`. Removing the
inner SQL window correctly exposed that missing ownership boundary. This is not
resolved by giving individual tests convenient deadlines.

Section 11.4 now explicitly requires every owned acquisition entry point to impose
the existing 35-second attempt and 180-second outer caps, with an earlier caller
context still authoritative. The attempt cap belongs to PostgreSQL acquisition;
the private suite-supplied attempt-timeout option is removed. Nil acquisition
contexts fail before preflight. This owner clarification precedes its implementation
and suite-composition projection. Endpoint, probe, polling, retry and cleanup caps
remain unchanged. The existing suite-startup test will verify forwarding of the
caller context; PostgreSQL owner tests verify the cap. No new public option or
legacy fallback is introduced.

The failed attempt remains at its exact root. Public `harness-observability-check
CHECK=diagnostics` passed **0**, `full_validity=0`, with complete diagnostic and
required-accounting coverage. Its cleanup/scan/source receipt is retained below
with subsequent focused outcomes. WF-11d remains IN_PROGRESS; this attempt is not
performance evidence.

**E-116 — WF-11d baseline owner-cap and projection assurance.**
E-115's failed attempt `r03-s011-corrective4-postgres-b-02` retains source
`sha256:74bfce87286215d0410890bb5850f2e6b5c099cbeae70d1c300854fdf0fea847`,
15 completed cleanup records, a passing 39-file secret scan and empty accounting
settlement. Its failure remains recorded; later passes do not replace it.

PostgreSQL acquisition now supplies both existing caps for every owned entry:
180 seconds overall and 35 seconds per attempt, each bounded by the caller.
Suite composition forwards the caller context and no longer supplies a private
PostgreSQL attempt-timeout option. Tests verify nil rejection before preflight,
earlier deadlines, three bounded attempts, unchanged backoff, and an outer
deadline that is not renewed across retries. Existing readiness/probe, terminal
security rejection and cleanup tests exercise the remaining-attempt contract.

Public baseline `format` (`r03-s011-corrective4-format-b-02`, **27362/0**,
**8,047 ms**) and unaccounted `source-check CHECK=catalog`
(`r03-s011-corrective4-prepare-b-02`, **45854/0**, **1,625 ms**) passed before
accounted focused execution. The following commands use `CGO_ENABLED=0`,
`CARTULARY_HARNESS_CACHE_MODE=off`, fresh listed `CARTULARY_TEST_RUN_ID`, and
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Roots are under
`.cartulary/test-results/`.

| Run ID | Public selection | Native session / exit | Graph duration | Result |
| --- | --- | --- | ---: | --- |
| `r03-s011-corrective4-postgres-b-03` | `make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle,platform.postgres.integration.fresh_fixture_admission` | **47635/0** | 11,468 ms | Four units pass; 15 completed cleanup records; 36-file scan passes |
| `r03-s011-corrective4-browser-b-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract` | **34867/0** | 3,564 ms | Both helper selections pass; 11 completed cleanup records; 21-file scan passes |
| `r03-s011-corrective4-evidence-b-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.suiteservices_helper_contract` | **64654/0** | 1,367 ms | Typed diagnostic consumer passes; ten completed cleanup records; 17-file scan passes |

Each run's public `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed **0**, `full_validity=1`, complete diagnostic coverage and
available required accounting. The validated executable source is
`sha256:dd2b8081667dc5e14ce240d2c3b2d18eed94e0a3cdbe8212b12928f795dcd1a7`.
These are development assurance receipts, not acceptance samples. Exact affected
revisions checks and baseline finalization/projections follow, then the same
common owner/code/schema correction is applied to preserved C3a. WF-11d remains
IN_PROGRESS; no treatment freeze or performance conclusion is claimed.

**E-117 — WF-11d baseline validation and immutable source; candidate adoption.**
The exact eight E-108 revisions rows passed under `make test-slice
OWNER=module.revisions ROWS=<the eight comma-separated E-108 IDs>` with the E-116
accounted environment: `r03-s011-corrective4-revisions-b-01`, **42680/0**,
**32,763 ms**, 14 completed cleanup records, passing 42-file scan and empty
accounting settlement. Public `CHECK=all` passed **0**, `full_validity=1`.
No readiness failure or unexecuted selected row remains in this focused check.

Baseline `make agent-finalize` passed at
`r03-s011-corrective4-finalize-b-01` (**59695/0**, **27,353 ms**), before broader
projection validation. Retained-run maintenance was skipped because `RESULTS_DIR`
was unset; no compatible successful full warm-check root was available.
Public `make source-check CHECK=generated`, `CHECK=schema`, and `CHECK=catalog`
passed respectively at `r03-s011-corrective4-generated-b-01` (**40865/0**,
**16,694 ms**), `r03-s011-corrective4-schema-b-01` (**93529/0**, **6,229 ms**),
and `r03-s011-corrective4-catalog-b-01` (**13338/0**, **1,627 ms**).
All bind the E-116 source digest unchanged after finalization.

Baseline B4 is preserved as
`.cartulary/remediation03/continuation/diagnostic-baseline-source-exp04.tar.gz`
and its file manifest. Archive digest:
`e28b00067605a1facc829d6b8ef7fc8bada27d30e4c5a048eab950ed21d6d44b`;
7,499 files, 48,072,603 bytes; HEAD remains `f42fc00b76ec4e296c2a379c09687e02ce669de7`.
This preservation receipt is not an acceptance sample.

Guarded restoration to C3a preserved the controlling tracker. Candidate Section
11.4 and the Section 11.5 retry mapping were adopted before applying the common
repair. The old Section 11.4 and all nine common authored source/schema inputs
were checked against the B3 preservation manifest before replacement; unexpected
divergence would stop the operation. The final common overlay includes suite
composition and its context-forwarding assurance. Current schema attachments move
to v5 and retire v4 atomically. No generated file was hand-edited; public candidate
generation and focused assurance follow. C3a's preparation lifecycle and E-112
host-binding correction remain the only intended behavioral treatment difference.
WF-11d remains IN_PROGRESS until candidate assurance, preservation and closing lint.

**E-118 — WF-11d complete; common startup correction validated in B4/C4.**
Candidate public generation passed at `r03-s011-corrective4-generate-c-01`
(**68294/0**, **11,000 ms**), followed by unaccounted
`make source-check CHECK=catalog` at `r03-s011-corrective4-prepare-c-01`
(**47259/0**, **1,622 ms**), both with `CGO_ENABLED=0`.
The following fresh executions use the E-116 environment and accounting selectors.

| Run ID | Public selection | Native session / exit | Graph duration | Settlement and scan |
| --- | --- | --- | ---: | --- |
| `r03-s011-corrective4-browser-c-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.native_nested_cancellation,harness.browser.unit.native_fixtures,harness.browser.unit.httptestx_helper_contract,harness.browser.unit.testcontainersx_helper_contract` | **43164/0** | 45,360 ms | Six units pass; 15 completed cleanup records; 145-file scan passes; accounting empty |
| `r03-s011-corrective4-postgres-c-01` | `make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle,platform.postgres.integration.fresh_fixture_admission` | **72129/0** | 9,459 ms | Four units pass; 15 completed cleanup records; 36-file scan passes; accounting empty |

Both exact roots passed public `make harness-observability-check
RESULTS_DIR=<exact-root> CHECK=all`, **0**, `full_validity=1`, complete diagnostic
coverage and available required accounting. Nested cancellation remains successful
with both the inherited-admission correction and common PostgreSQL startup change.
No candidate-owned resource remains unresolved in these corrective invocations.

`make agent-finalize` passed at `r03-s011-corrective4-finalize-c-01`
(**24631/0**, **27,153 ms**), before broader candidate projection checks.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
`make source-check CHECK=generated`, `CHECK=schema`, and `CHECK=catalog` then
passed at `r03-s011-corrective4-generated-c-01` (**20108/0**, **16,042 ms**),
`r03-s011-corrective4-schema-c-01` (**4296/0**, **6,289 ms**), and
`r03-s011-corrective4-catalog-c-01` (**4881/0**, **1,659 ms**).

All candidate checks bind executable source
`sha256:f67da8a4a3e3340b4470a045499846064289a85bc042e039293a744b35f59fb0`.
C4 is preserved in
`.cartulary/remediation03/continuation/lifecycle-candidate-source-exp04.tar.gz`
and its manifest: archive digest
`411e5a00e5ec22fe92802771adcf86229fc8e6ef4e710afb985b13144f7bce41`,
7,507 files, 48,101,729 bytes, unchanged HEAD. B4 is E-117's preserved source.
Nine common authored implementation/test/schema inputs are byte-identical in B4/C4.
The treatment-difference set remains 360 paths, excluding the controlling tracker,
with only the common generated PostgreSQL validator's v4-to-v5 filename migration.
Its generated bundle still contains the existing lifecycle schema differences.
The comparison receipt is `exp04-treatment-difference.json` under the preservation
directory; it does not interpret canonical events.

Public candidate browser/CI `make target-plan TARGET=<target>
INPUTS=<performance-browser-inputs.json or performance-ci-inputs.json> JSON=1`
exports passed **68612/0** and **69098/0**; both exactly equal their C3 plans.
Kernel, system/toolchain digests, filesystem identity and repository location
remain unchanged. These observations do not replace new excluded warm-ups.

Traceability: F-013 → R03-026/Section 11.4/REQ-401/402/AC-007 → `pgtest`
acquisition/readiness/cleanup, suite composition and diagnostic v5 → V-014's
focused routes, real fixture admission, exact affected revisions rows, nested
cancellation and projection checks. Assurance is PASSED; performance is NOT
ESTABLISHED. The diagnostic contract cutover is complete in both preserved
sources with no dual reader; previous live invocations were terminal before
mutation. Common corrections are retained regardless of the lifecycle experiment's
disposition. Earlier failed attempts and cleanup failures remain historical evidence.

WF-11d is DONE. Closing public `make lint-markdown` uses
`r03-s011-corrective4-close-lint-01`; WF-11 may resume only after it passes.
Next entry requires a fresh B4/C4 experiment freeze, baseline plan revalidation,
four declared excluded warm-ups, independent evidence audits and runtime-graph
binding before any matched sample. Full CI and S-007e remain open.

**E-119 — Experiment 04 frozen after common startup correction.**
WF-11d closing Markdown lint passed at `r03-s011-corrective4-close-lint-01`
(**94943/0**, **16,790 ms**) before WF-11 resumed. Guarded C4→B4 restoration
preserved the tracker. Public baseline `make generate` passed at
`r03-s011-exp04-b-generate-01` (**84632/0**, **10,750 ms**); unaccounted
`make source-check CHECK=catalog` passed at `r03-s011-exp04-b-prepare-01`
(**22271/0**, **1,636 ms**), with `CGO_ENABLED=0`.

B4 and C4 are the immutable E-117/E-118 sources. Both contain the accepted
critical-path diagnostics and identical common PostgreSQL correctness repair.
B4 retains standalone reset units. C4 contains the preparation lifecycle and its
required inherited-host binding preservation (E-112). No scheduler ranking,
fairness, capacities, lanes, workers, retries, eligibility or pooling changes are
permitted. Changed-source comparison SHALL NOT use `COMPARISON=equivalent`.

Public B4 browser/CI plan exports passed **29258/0** and **56449/0**, using the
same `target-plan` inputs as E-118. B4 plans exactly equal B3; C4 plans exactly
equal C3. Frozen populations remain 383 browser rows / 65 cohorts and 2,279 CI
rows / 96 cohorts. Browser units change 137→76; CI units change 600→517 through
61 and 83 reset transfers respectively. Static plan digests and exact populations
are retained in `baseline4-*-plan.json` and `candidate4-*-plan.json` under
`.cartulary/remediation03/continuation/`. Runtime CI graph binding MUST still be
validated after the warm-ups: the existing machine Go-worker resolution may
change the same two CPU claims from four to nine and their semantic identities
(E-090). Prior run graphs do not establish current binding.

The immutable protocol is `performance-experiment-exp04-freeze.json`; mutable
observations use `performance-experiment.json`. Previous experiment 03 and its
freeze remain preserved, including failed/omitted attempts. The source treatment
map is `exp04-treatment-difference.json`; environment identity is retained in
`performance-environment-exp04.json`. Kernel, system/toolchain, machine Go paths,
repository and filesystem identities are unchanged. Build/dependency caches stay
warm; execution-result reuse is off. No competing intentional workload is added.
Observation SHALL stop structured polling during known quiet measurement.

New excluded warm-ups are declared in order:
`r03-s011-exp04-browser-warm-b`, `r03-s011-exp04-ci-warm-b`,
`r03-s011-exp04-browser-warm-c`, `r03-s011-exp04-ci-warm-c`.
Each uses its workload's public `make browser-e2e SUITE=functional` or `make ci`,
`CGO_ENABLED=0`, `CARTULARY_HARNESS_CACHE_MODE=off`, fresh explicit
`CARTULARY_TEST_RUN_ID`, and Make selectors
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Every warm-up MUST pass independent evidence validation, final secret scan,
physical settlement and source/graph binding before acceptance sampling begins.

The four-pair maximum, alternating baseline-first/candidate-first order,
first-two-pair <5% rejection and >=20% early acceptance, >5% full-protocol benefit
threshold, five <=5% guardrails, 10,000 bootstrap resamples, seed `20261021`,
decimal arithmetic and confidence levels remain unchanged (Section 8.5).
Both workloads receive their first two pairs unless correctness or cleanup fails.
A workload rejection stops optional sampling and restores the lifecycle owner,
projections and implementation together while retaining accepted diagnostics and
common corrections. No confidence bound is claimed for omitted performance work.
All failed or omitted attempts remain explicit; source repair requires another
freeze. Supplementary critical-path, allocation retirement/reset-only generation,
identity continuity and cohort-wait observations retain their declared scopes.

Assurance: corrective focused checks and projections pass; full-CI gate open.
Performance: zero acceptance samples, no benefit or guardrail conclusion.
Freeze Markdown lint `r03-s011-exp04-freeze-lint-01` passed
(**11902/0**, **8,990 ms**) before the first excluded warm-up began. Final S-007e validation remains required for the retained source.

**E-120 — Experiment 04 failed baseline warm-up; WF-11e registered.**
`r03-s011-exp04-browser-warm-b` completed in **47447/2**, graph duration
**1,111,994 ms**, against frozen B4 and its exact browser graph. The summary
contains 135 passed and two failed units: the assertion cohort
`browser_group:webserver-backed:cohort-8fadb014babd4274e2b8616a59de21d4c011190889717f7fe69d6b55f4016848`
and the propagated `browser_target_summary:browser-e2e-webserver-backed` failure.
Both retain native exit **10**, `product/test_assertion_failure`.

The one root assertion failure is row
`module.assessments.browser.the_browser_workbook_assessment_workflow_appends_aefa86174b`,
Playwright `sentinel.spec.ts` test **assessment workflow keeps invalid timestamp
drafts local**. It timed out at 60 seconds while filling the rationale field
(line 1182). The retained error context shows that the inspector remained open,
with the existing draft preserved behind **Resume assessment draft**; the field
had detached. Thirteen further rows in that cohort are `skipped_dependency`, not
independent product failures. The exact group result, Playwright report, trace and
error-context references remain under that cohort's retained `browser-groups/`
directory. Historical instructions embedded in Playwright's error-context Markdown
are diagnostic data, not execution authority.

Source inspection establishes a missing setup precondition: the scenario commits
two subject rows and a support row, then immediately opens and edits a draft.
Ordinary write completion does not establish collaboration publication. The
retained-authoring owner invalidates review for relevant observed record changes,
preserving the draft. The browser capture does not retain the exact triggering
socket message, so a particular delayed publication is a supported explanation,
not a directly observed historical cause. Existing workbook-query browser fixtures
already establish publication barriers using `installIncidentSocketMonitor` before
starting their assertions. The correction uses that existing protocol boundary;
it SHALL NOT weaken the product's review/authority behavior, add sleeps, raise the
60-second allowance, automatically resume a draft, or retry the test.

Public `make harness-observability-check RESULTS_DIR=<exact-root>` with
`CHECK=events`, `CHECK=timing` and `CHECK=diagnostics` each passed **0**,
`full_validity=0`; diagnostic coverage is complete and accounting available.
All **277** cleanup records completed; the final **2,319-file** secret scan passed;
required accounting reports `settlement=empty`. Owning-command duration is
**1,124,307.399472 ms**, charged CPU **3,842,631,909 us**, kernel peak memory
**13,461,626,880 bytes**; these describe a failed excluded warm-up and establish
no performance result. Public performance export passed **0** and is retained as
`r03-s011-exp04-browser-warm-b-performance.json`. No resource remains unresolved
from this invocation. Source was unchanged until its native command exited.

Experiment 04 is preserved in `performance-experiment-exp04.json` and its immutable
freeze. The other three warm-ups and all sixteen acceptance attempts are explicitly
omitted because correctness failed; none is silently rerun or counted. There are
still zero matched acceptance samples. F-014/R03-027/V-015/T-022 register WF-11e
as the next serial corrective slice. It changes only the affected browser test's
fixture preparation in both treatments; Core 03 remains authoritative and no new
product behavior or NLSpec amendment is required. Public owner guides for
`module.assessments` and `web.workbook` passed **97522/0** and **91203/0**.

WF-11e SHALL first require public entry `make lint-markdown` at
`r03-s011-corrective5-entry-lint-01`, then establish an accepted socket before
fixture writes and observe each exact seeded record's publication before authoring.
Validation SHALL run the affected row alone, containing assessment rows, and the
existing retained-authoring/socket-monitor contract routes. Source correction,
finalization, projection checks and immutable B5/C5 preservation SHALL precede a
new experiment freeze. Both performance decisions and S-007e remain open; this
correctness failure is not deferred or relabeled an optimization rejection.

**E-121 — WF-11e baseline fixture correction and focused assurance.**
Entry Markdown lint passed at `r03-s011-corrective5-entry-lint-01`
(**41823/0**, **9,270 ms**) before editing. The sole executable-source change
from B4 is `apps/web/e2e/sentinel.spec.ts`: the existing incident socket monitor
is installed before navigation; its accepted-socket observation precedes the
three ordinary fixture creates; each exact returned `record_id`/`row_version`
publication is observed before draft authoring. Existing event buffering handles
publication arriving before the wait is registered. No sleep, application retry,
extra create, timeout increase, automatic draft resume, selector change or product
behavior change was introduced. The original navigation moved earlier; no extra
page navigation was added. All original assertions remain.

Public `make format` passed at `r03-s011-corrective5-format-b-01`
(**5700/0**, **8,056 ms**); unaccounted `make source-check CHECK=catalog` passed
at `r03-s011-corrective5-prepare-b-01` (**76216/0**, **1,691 ms**).
Focused commands below use `CGO_ENABLED=0`, `CARTULARY_HARNESS_CACHE_MODE=off`,
fresh listed `CARTULARY_TEST_RUN_ID`, and explicit Make selectors
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Roots are under
`.cartulary/test-results/`.

| Run ID | Public `make test-slice` selection | Native session / exit | Graph duration | Cleanup / final scan |
| --- | --- | --- | ---: | --- |
| `r03-s011-corrective5-assessment-b-01` | `OWNER=module.assessments ROWS=module.assessments.browser.the_browser_workbook_assessment_workflow_appends_aefa86174b` | **90096/0** | 86,533 ms | 23 completed; 97-file scan passes |
| `r03-s011-corrective5-assessments-b-01` | `OWNER=module.assessments ROWS=module.assessments.browser.append_recovery,module.assessments.browser.contextual_decision_refresh,module.assessments.browser.the_browser_appends_a_subject_only_assessment_fo_2a82260393,module.assessments.browser.the_browser_workbook_assessment_workflow_appends_aefa86174b,module.assessments.browser.the_browser_workbook_drives_the_assessments_surf_4f277414a9` | **78911/0** | 104,519 ms | 23 completed; 101-file scan passes |
| `r03-s011-corrective5-authoring-b-01` | `OWNER=web.workbook ROWS=web.workbook.regression.assessment_append_recovery,web.workbook.regression.workbookz_assessment_creation_controller_2f0f2487ec` | **85317/0** | 3,279 ms | 11 completed; 27-file scan passes |
| `r03-s011-corrective5-socket-b-01` | `OWNER=harness.browser ROWS=harness.browser.boundary_support.incidentsocket_suite_6db375c4ec` | **83064/0** | 1,381 ms | 11 completed; 26-file scan passes |

Every run passed independent public `make harness-observability-check
RESULTS_DIR=<exact-root> CHECK=all`, **0**, `full_validity=1`, complete diagnostic
coverage, available required accounting and empty settlement. Both browser runs
passed all twelve scheduler units, including finalization; owner and monitor
slices each passed two units. Validated executable source:
`sha256:73c5d4ab4d97f0454de97dfd399dbca604819d93c3f1dfa2ee40cddbd3c63d5a`.
The historical timeout remains recorded and its precise socket trigger remains
unavailable. These current passes establish corrected setup and retained
assertions, not a reconstructed historical event or performance acceptance.

Baseline `make agent-finalize` passed at `r03-s011-corrective5-finalize-b-01`
(**89578/0**, **27,138 ms**), before broader type/projection checks. Retained-run
maintenance was skipped because `RESULTS_DIR` was unset. The current public
`make help ALL=1` identifies `make lint CHECK=types` for type checking; no retired
private frontend command is used. WF-11e remains IN_PROGRESS pending projection
checks, both-source preservation, candidate assurance and closing Markdown lint.

**E-122 — WF-11e baseline preservation and candidate isolated pass.**
After E-121 finalization, public baseline `make lint CHECK=types` passed at
`r03-s011-corrective5-types-b-01` (**1764/0**, **20,542 ms**).
`make source-check CHECK=generated` and `CHECK=catalog` passed at
`r03-s011-corrective5-generated-b-01` (**86005/0**, **16,457 ms**) and
`r03-s011-corrective5-catalog-b-01` (**28369/0**, **1,697 ms**).
No schema contract changed in WF-11e; WF-11d's schema receipts remain explicit,
and S-007e still requires current retained-source schema validation.

B5 preserves E-121's source digest and unchanged HEAD in
`.cartulary/remediation03/continuation/diagnostic-baseline-source-exp05.tar.gz`
with its file manifest: archive digest
`83ed391ddf0f84f6aaf50dadc6796852e23ec330b2e595b6432b30040eb871a6`,
7,499 files, 48,076,435 bytes. Guarded restoration to C4 preserved the tracker.
The candidate's pre-edit `sentinel.spec.ts` matched B4 exactly; the single repaired
file was copied byte-for-byte from preserved B5 after verifying its archive member
digest. No product, owner, schema or unrelated source was replaced.

Candidate public `make generate` passed at `r03-s011-corrective5-generate-c-01`
(**30813/0**, **10,980 ms**), followed by unaccounted
`make source-check CHECK=catalog` at `r03-s011-corrective5-prepare-c-01`
(**43994/0**, **1,669 ms**). The exact affected browser row from E-121 passed
on the candidate with identical accounted settings at
`r03-s011-corrective5-assessment-c-01` (**74159/0**, **85,394 ms**).
Public `CHECK=all` passed **0**, `full_validity=1`, complete diagnostics and
available required accounting. Candidate containing assurance and finalization
follow; WF-11e remains IN_PROGRESS. No experiment 05 freeze or performance sample
is claimed yet.

**E-123 — WF-11e complete; fixture correction retained in B5/C5.**
The candidate containing command uses E-121's exact five `module.assessments`
browser rows and identical accounted settings. It passed at
`r03-s011-corrective5-assessments-c-01` (**28751/0**, **109,982 ms**).
Public `CHECK=all` passed **0**, `full_validity=1`, complete diagnostics and
available required accounting. Both candidate browser checks passed all twelve
scheduler units and completed all **23** cleanup records. The isolated check's
final scan passed **97** files; the containing check's scan passed **101** files.
Both required accounting receipts report `settlement=empty`.

Candidate `make agent-finalize` passed at `r03-s011-corrective5-finalize-c-01`
(**14419/0**, **27,691 ms**) before broader validation. Retained-run maintenance
was skipped because `RESULTS_DIR` was unset. Public `make lint CHECK=types`
passed at `r03-s011-corrective5-types-c-01` (**82299/0**, **20,734 ms**).
`make source-check CHECK=generated` and `CHECK=catalog` passed at
`r03-s011-corrective5-generated-c-01` (**85081/0**, **17,576 ms**) and
`r03-s011-corrective5-catalog-c-01` (**27905/0**, **1,710 ms**).
Candidate public browser and CI target-plan exports passed **28020/0** and
**39278/0** using the E-118 input files; both exactly equal their C4 plans.

All candidate corrective checks bind executable source
`sha256:fd1d49190a4d248a3d4c68e46e4c5045b14655c1765e3165684cf935572c8d39`.
C5 is preserved in
`.cartulary/remediation03/continuation/lifecycle-candidate-source-exp05.tar.gz`
and its manifest: archive digest
`947adeb8a726959df5c13b5e9a6acbe2c91fc666f2fdc8b5413def18575f4a09`,
7,507 files, 48,106,275 bytes; HEAD unchanged. B5 is preserved under E-122.
Comparing each treatment against its experiment-04 snapshot shows exactly one
changed executable-source file, `apps/web/e2e/sentinel.spec.ts`; that file is
byte-identical in B5/C5. The B/C treatment-difference set remains the same 360
paths excluding the controlling tracker, recorded in `exp05-treatment-difference.json`.
No policy, capacity, scheduling, deadline, retry, product or selected-population
change entered this repair.

Traceability: F-014 → R03-027 and existing Core 03 retained-intent/review authority
→ the assessment scenario's fixture setup using the existing incident socket
monitor → V-015 isolated/containing browser checks, retained-authoring and monitor
contracts, type checking and generated/catalog validation. Assurance is PASSED;
performance is NOT ESTABLISHED. The precise historical invalidation message
remains unavailable; no assertion of reconstructed cause replaces that limit.
The failed experiment and its omitted attempts remain preserved. No corrective
invocation has unresolved candidate-owned resources. There is no schema migration
or compatibility path in this test-only correction; rejection of the lifecycle
candidate retains this identical common correction.

WF-11e is DONE. Closing public `make lint-markdown` uses
`r03-s011-corrective5-close-lint-01`; WF-11 resumes only after it passes. Next entry
requires baseline plan revalidation, a fresh B5/C5 freeze, four excluded warm-ups,
independent audits and actual runtime-graph binding. No matched acceptance sample
has run. Both workload performance decisions and final S-007e remain required.

**E-124 — Experiment 05 frozen after fixture-publication correction.**
WF-11e closing Markdown lint passed at `r03-s011-corrective5-close-lint-01`
(**77578/0**, **9,680 ms**) before WF-11 resumed. Guarded C5→B5 restoration
preserved the tracker. Public baseline `make generate` passed at
`r03-s011-exp05-b-generate-01` (**43715/0**, **11,260 ms**); unaccounted
`make source-check CHECK=catalog` passed at `r03-s011-exp05-b-prepare-01`
(**64645/0**, **1,702 ms**), with `CGO_ENABLED=0`.
Public baseline browser/CI target-plan exports passed **40791/0** and **8882/0**.
Both B5 and C5 plan exports exactly equal their respective experiment-04 plans.

B5/C5 are E-122/E-123's immutable sources. They share all accepted diagnostics,
common startup/toolchain/socket-object corrections, and the exact assessment
fixture-publication correction. The only permitted treatment remains the
standalone-reset → admitted preparation lifecycle plus its E-112 inherited-host
binding preservation. Populations remain 383 browser rows / 65 cohorts and
2,279 CI rows / 96 cohorts; units remain B137/C76 and B600/C517. No selection,
capacity, ranking, fairness, lane, worker, retry, isolation or pooling treatment
is admitted. Kernel, system/toolchain, storage/repository and machine Go-cache
identities revalidate unchanged. Runtime graph binding remains a warm-up obligation;
static-plan equality does not replace it.

The immutable ledger is `performance-experiment-exp05-freeze.json`; observations
use `performance-experiment.json`, with `performance-environment-exp05.json` and
`exp05-treatment-difference.json` retained alongside all prior failed experiments.
The entire Section 8.5/E-119 protocol remains unchanged: four matched pairs maximum,
odd baseline-first/even candidate-first, first-two-pair <5% rejection, >=20% early
acceptance without omitted bounds, otherwise >5% benefit lower 95% bound and all
five <=5% regression upper 99% bounds, 10,000 paired bootstrap resamples with seed
`20261021` and the frozen decimal arithmetic. Every attempt requires independent
correctness, evidence, cleanup and final-scan validation. Changed-source comparison
never uses `COMPARISON=equivalent`; failed/omitted attempts remain explicit.
A repair requires another source freeze. Both workloads receive their first two
pairs unless correctness or cleanup prevents execution; rejection restores the
lifecycle owner/projections/code together and retains the common corrections.

New excluded warm-ups are declared in order:
`r03-s011-exp05-browser-warm-b`, `r03-s011-exp05-ci-warm-b`,
`r03-s011-exp05-browser-warm-c`, `r03-s011-exp05-ci-warm-c`.
Commands remain `make browser-e2e SUITE=functional` and `make ci`, with
`CGO_ENABLED=0`, `CARTULARY_HARNESS_CACHE_MODE=off`, fresh explicit
`CARTULARY_TEST_RUN_ID`, and Make selectors
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. All four warm-ups must
pass their independent audits and actual source/runtime-graph binding before
sampling. Dependency/build caches remain warm; execution-result reuse remains
disabled. No competing intentional workload is added, and structured observation
stops during known quiet measurement. Supplementary path phases, queue populations,
allocation retirement causes, reset-only generations and reset/runner identity
continuity retain E-119's explicit measurement scopes.

Assurance: focused repairs and projections pass; full-CI gate open. Performance:
zero matched samples; no measured benefit or guardrail conclusion. Freeze public
Markdown lint `r03-s011-exp05-freeze-lint-01` passed (**49315/0**, **9,430 ms**)
before the first excluded warm-up began. Final S-007e remains required for whichever lifecycle disposition is retained.

**E-125 — Experiment 05 baseline browser warm-up passed; CI warm-up active.**
The frozen E-124 browser command `r03-s011-exp05-browser-warm-b` passed in native
**16002/0**, graph duration **1,251,216 ms**, all **137** scheduler units passed.
Its source, graph, system/toolchain and 19-worker machine resolution match the
frozen B5 browser plan. The formerly failing cohort passed all **32** rows,
including the repaired assessment scenario; that published cohort duration is
**320,451 ms**, not an independent performance sample.

Independent public `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed **0**, `full_validity=1`, complete diagnostic coverage and
available required accounting. All **276** cleanup records completed; the final
**2,296-file** secret scan passed; accounting reports `settlement=empty`.
Public `make explain-run` performance and critical-path `VIEW=units`, `waits`,
`allocations` exports each exited **0**. The three path views at `OFFSET=0
LIMIT=100` returned their complete populations (80 units, 80 wait segments, one
path-associated allocation), with `next_offset=null`, complete coverage and zero
allocation omissions. Full invocation lifecycle records independently contain
four allocations, all settled with actual `terminal_borrower_release` reasons;
none is reset-only. All exports and the validated lifecycle section are preserved
with the run-ID prefix under `.cartulary/remediation03/continuation/`.

| Excluded warm-up metric | Observation |
| --- | ---: |
| Owning-command elapsed | 1,263,908.600535 ms |
| Charged invocation CPU | 4,087,450,746 us |
| Kernel peak memory charge | 8,342,966,272 bytes |
| p95 ready-to-start delay | 1 ms |
| Cleanup interval union | 47,339 ms |

The complete 80-unit path has graph queue **19 ms** and unit lifetime
**1,250,676 ms**. Summed recorded path phase durations are host admission
**106,358 ms**, acquisition **26,058 ms**, runner **1,089,362 ms**, release
**26,232 ms**, publication **1,806 ms**, cache lookup/publication **4/1 ms**;
uncovered unit lifetime is **859 ms**. These phase observations are not additive
exclusive causes; graph queue and host admission remain distinct. Claims are not
utilization. Resource metrics retain the delegated Linux scope and exclusions.
This is an excluded warm-up with no benefit or guardrail acceptance conclusion.

The next frozen command, `r03-s011-exp05-ci-warm-b`, is active in native **40977**:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-exp05-ci-warm-b
CARTULARY_HARNESS_CACHE_MODE=off make ci HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required`. Source remains B5. Candidate warm-ups follow only
after the baseline CI's terminal outcome, evidence audit and physical settlement.
All four warm-ups still precede matched sampling. S-007e remains open.

**E-126 — Experiment 05 failed CI warm-up; WF-11f registered.**
The frozen baseline CI command in E-125 exited **40977/2**, graph duration
**3,664,837 ms**, **599 passed / 1 failed / 600 total** units. Source is B5
`sha256:73c5d4ab4d97f0454de97dfd399dbca604819d93c3f1dfa2ee40cddbd3c63d5a`;
runtime graph is `sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
Exact 2,279-row/600-unit membership matches the frozen baseline plan. The only
static-to-runtime differences are the previously permitted two Go CPU claims
4→9 and their semantic digests; `exp05-baseline-ci-runtime-differences.json`
preserves the comparison. There is no new scheduling or capacity treatment.

The sole failed unit is `operation:harness.lint_go_staticcheck`, native unit exit
**1**, canonical `product/test_assertion_failure`. Its referenced stdout reports
SA1012 at `internal/testutil/pgtest/pgtest_test.go:257`, the deliberately nil
`StartOwned` input asserting rejection before preflight/provider execution.
All other units passed, including product/browser and corrected startup tests.
The initial bounded inspection used that unit's stdout/stderr only. This is a
static-analysis failure introduced by the common negative test, not evidence of
a new runtime startup or browser failure. It remains a mandatory failed CI run.

Independent public `make harness-observability-check RESULTS_DIR=<exact-root>`
with `CHECK=events`, `timing`, and `diagnostics` each exited **0**,
`full_validity=0`. Optional verbose execution/resources are truncated (resource
omissions **581**); required invocation accounting is available. The separate
allocation-lifecycle section is complete with zero omissions: **87** allocations,
all settled, **74 capacity_pressure / 13 terminal_borrower_release**, including
**30 reset-only** generations. No reason is inferred from timing. All **1,165**
cleanup records completed; accounting reports `settlement=empty`; the final
**9,819-file** secret scan passed. No resource remains unresolved by this attempt.

Public performance and critical-path unit exports each exited **0** and are
retained under the run-ID prefix. The complete 80-unit path has graph queue
**690,162 ms** and unit lifetime **2,973,154 ms**; the complete page has
`next_offset=null`. Required charged metrics are owning-command elapsed
**3,750,879.763446 ms**, CPU **12,070,930,574 us**, kernel peak
**12,536,094,720 bytes**, p95 ready-to-start **3,227,158 ms**, and cleanup union
**156,500 ms**. These are descriptive failed-warm-up observations, not acceptance
samples or additive phase causes. Resource scope remains the delegated Linux
invocation and proven containers, with previously declared exclusions.

`performance-experiment-exp05.json` preserves both attempted warm-ups and marks
the remaining two warm-ups and all sixteen planned samples omitted because the
baseline correctness failure requires a repaired source and a fresh experiment.
The immutable freeze is unchanged. No matched acceptance sample has run.

F-015/R03-028/V-016/T-023 register **WF-11f**. Its only executable-source change
SHALL document a single-call SA1012 exception for this intentional invalid-input
test, identically in both sources. It SHALL retain the test, its assertion,
global analyzer enforcement, runtime contracts and all treatment differences.
No NLSpec behavioral amendment is required. Entry public `make lint-markdown`
uses `r03-s011-corrective6-entry-lint-01` and MUST pass before implementation.
Public `make lint CHECK=go` SHALL pass before the next freeze, alongside the
exact affected owner row and finalization/projection checks. After both-source
assurance and closing tracker/lint, WF-11 requires fresh B6/C6 identities and all
four excluded warm-ups. Lifecycle performance and S-007e remain open.

**E-127 — WF-11f complete; negative-test lint correction preserved in B6/C6.**
Entry Markdown lint `r03-s011-corrective6-entry-lint-01` passed **51136/0**,
**9,150 ms**, before the source edit. The adopted Section 11.4 nil-context
rejection remains unchanged. `internal/testutil/pgtest/pgtest_test.go` now has
one `//lint:ignore SA1012` comment immediately before `StartOwned(nil)`, explaining
the deliberate public nil-context guard test before acquisition. The nil input,
assertion, preflight/provider exclusion and global analyzer policy are unchanged.
The repository already uses this explicit negative-test convention in migration
contract assurance. No runtime, schema, catalog or product behavior changed.

All commands below ran from the repository root with `CGO_ENABLED=0`. Accounted
lint/test commands additionally used `CARTULARY_HARNESS_CACHE_MODE=off`,
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Roots use
`.cartulary/test-results/r03-s011-corrective6-<suffix>` and fresh explicit IDs.

| Suffix | Public command | Native session / exit | Graph/tool duration ms |
| --- | --- | --- | ---: |
| `format-b-01` | `make format` | 27550 / 0 | 8,076 |
| `prepare-b-01` | `make source-check CHECK=catalog` | 76356 / 0 | 1,642 |
| `go-lint-b-01` | `make lint CHECK=go` | 17147 / 0 | 8,132 |
| `postgres-b-01` | `make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle` | 75115 / 0 | 1,213 |
| `finalize-b-01` | `make agent-finalize` | 79424 / 0 | 27,197 |
| `generated-b-01` | `make source-check CHECK=generated` | 23488 / 0 | 16,534 |
| `catalog-b-01` | `make source-check CHECK=catalog` | 91045 / 0 | 1,714 |
| `generate-c-01` | `make generate` | 5425 / 0 | 11,000 |
| `prepare-c-01` | `make source-check CHECK=catalog` | 32576 / 0 | 1,687 |
| `go-lint-c-01` | `make lint CHECK=go` | 55922 / 0 | 11,955 |
| `postgres-c-01` | Same exact PostgreSQL row command above | 40631 / 0 | 1,261 |
| `finalize-c-01` | `make agent-finalize` | 26886 / 0 | 27,074 |
| `generated-c-01` | `make source-check CHECK=generated` | 5426 / 0 | 16,389 |
| `catalog-c-01` | `make source-check CHECK=catalog` | 57597 / 0 | 1,653 |

Each of the four accounted lint/test runs independently passed public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, exit **0**,
`full_validity=1`. Lint has six passing units and 15 completed cleanup records
per treatment; the exact test row has one passing unit and ten completed cleanup
records per treatment. Every accounting scope is `empty`. Final retained-secret
scans pass (36 files per lint run, 17 per test run). Candidate lint's optional
resource sampling is partial with 51 omissions; required invocation accounting
is available. This focused assurance makes no performance or total sampled-resource
claim. Finalization preceded broader generated/catalog checks. Retained-run
maintenance was skipped because `RESULTS_DIR` was unset; no compatible successful
full warm-check root is available. No golden changed.

B6 is `sha256:57547d07a9631951141df095cb7add363c5428c3f2de82a308774bfb4cb32c1e`,
archive `f6093559efb17a2646a60021bb6011d78dcbde3fd4f82022b6754e1431797123`
(7,499 files / 48,078,991 bytes), bound to baseline catalog validation.
C6 is `sha256:2b06c4b28df6d1a342a65378bd05c5b1db2b02b047177073b31844d236b45971`,
archive `dd90817a828f6343eb0ed325954d1e942ada7c5d9474dfbc76c3dc5aca68be13`
(7,507 files / 48,111,759 bytes), bound to candidate catalog validation.
Both immutable archives remain under `.cartulary/remediation03/continuation/`.
HEAD remains `f42fc00b76ec4e296c2a379c09687e02ce669de7`. Guarded restoration
preserved tracker edits. The common test file is byte-identical; each source's
only delta from its experiment-05 counterpart is this comment. Their unchanged
360-path treatment difference remains the lifecycle change and its already
adopted inherited-host binding correction.

Traceability is F-015 → R03-028 / Section 11.4 → exact negative-test call → V-016
public Go lint and affected owner row → E-127 audits. Assurance passes; performance
is unestablished. There is no schema migration or runtime rollback for this common
annotation; a rejected lifecycle candidate retains it. All previous failed and
omitted attempts remain preserved. No resource is unresolved by this correction.

WF-11f is DONE. Closing public `make lint-markdown` uses
`r03-s011-corrective6-close-lint-01` and MUST pass before WF-11 resumes. The next
entry requires revalidated plans/environment, a fresh B6/C6 freeze, four excluded
warm-ups, and independent source/graph/evidence/settlement audits before any matched
sample. Both workload performance decisions and S-007e remain open.

**E-128 — Experiment 06 frozen after negative-test lint correction.**
WF-11f closing Markdown lint passed at `r03-s011-corrective6-close-lint-01`,
**37761/0**, **9,840 ms**, before WF-11 resumed. Candidate public target-plan
exports exited **8795/0** (browser) and **84571/0** (CI). After guarded C6→B6
restoration, public `make generate` passed at `r03-s011-exp06-b-generate-01`,
**33018/0**, **10,820 ms**; unaccounted `make source-check CHECK=catalog` passed
at `r03-s011-exp06-b-prepare-01`, **59146/0**, **1,606 ms**. All used
`CGO_ENABLED=0`. Baseline public target-plan exports exited **27893/0** and
**1132/0**. All four plans exactly equal their experiment-05 counterparts.

E-127's immutable B6/C6 source identities and archives are the new treatments.
The common nil-context annotation is byte-identical and is each source's only
change from experiment 05. Their 360 differing paths are unchanged. Exact
populations remain browser 383 rows / 65 cohorts / B137→C76 units (61 reset
transfers), and CI 2,279 rows / 96 cohorts / B600→C517 units (83 reset transfers).
No selection, ranking, fairness, capacity, lane, worker, retry, isolation or
pooling change is permitted. Kernel, system/toolchain, storage and machine
Go-cache identities revalidate unchanged. Actual runtime-graph binding remains
a warm-up obligation; only the already specified two Go claims and corresponding
semantic digests may differ from static worker defaults.

`performance-experiment-exp06-freeze.json`, `performance-environment-exp06.json`,
and `exp06-treatment-difference.json` preserve the freeze; mutable observations
use `performance-experiment.json`. All prior experiment freezes, failed attempts
and omissions remain preserved. No historical run is promoted to this source.

The frozen excluded warm-ups are, in order,
`r03-s011-exp06-browser-warm-b`, `r03-s011-exp06-ci-warm-b`,
`r03-s011-exp06-browser-warm-c`, `r03-s011-exp06-ci-warm-c`.
Commands remain `make browser-e2e SUITE=functional` and `make ci` with
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off`, explicit fresh
`CARTULARY_TEST_RUN_ID`, and `HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Dependencies/build caches remain warm and execution-result reuse is disabled.
All four excluded warm-ups MUST pass independent correctness/evidence/cleanup/
scan audits and exact source/runtime-graph binding before sampling. No competing
intentional workload or structured polling during known quiet measurement is
introduced.

Section 8.5's protocol remains unchanged: at most four matched pairs per
workload, odd baseline-first/even candidate-first; both first-two benefits <5%
reject that workload; either >=20% permits early acceptance without omitted
confidence bounds; otherwise the four-pair 95% benefit lower bound must exceed
5% and all five 99% guardrail upper bounds must be <=5%, with 10,000 paired
bootstrap resamples, seed `20261021` and the frozen decimal arithmetic. Correctness
and cleanup remain mandatory for every attempt. Both workloads receive their
first two pairs unless correctness/cleanup prevents them. Rejection restores
lifecycle owner/projections/implementation together and retains common corrections.
Changed-source comparison never uses `COMPARISON=equivalent`. Repairs require
another freeze. Supplementary phase/queue/allocation/reset continuity analysis
retains its declared coverage and population limits.

Assurance: focused correction and projection checks pass; full-CI acceptance is
open. Performance: zero matched samples, no benefit/guardrail conclusion.
Freeze public `make lint-markdown` at `r03-s011-exp06-freeze-lint-01` passed
**28325/0**, **9,480 ms**, before the first warm-up began. Baseline browser warm-up
`r03-s011-exp06-browser-warm-b` is active in native **49166** under the frozen
command and inputs. Both workload decisions and S-007e remain required.

**E-129 — Experiment 06 failed browser warm-up; WF-11g registered.**
`r03-s011-exp06-browser-warm-b` exited **49166/2**, graph duration
**1,018,731 ms**, **135 passed / 2 failed / 137 total** units. B6 source and
`sha256:cf92b0a6c12752fe0769de3c8cebf259693395f60c2db17ec6d4f966f33876d0`
graph match the frozen browser selection. The failed cohort
`cohort-3bef2df9d96eb5a5cbc0f07a6b8d6143026dfed9f7ea0aa0983083e3fb380771`
and propagated browser target finalizer both retain the failure; the cohort's
nine rows are **3 passed / 1 failed / 5 dependency-skipped**, native exit **10**,
canonical `product/test_assertion_failure`. Independent work and finalizers
were allowed to drain without cancellation or source edits.

The root scenario is `Mention creation survives rejected resolution and sheet
navigation without another creation`, row
`module.entities.browser.timeline_mention_creation_recovery`.
`apps/web/e2e/mentions.recovery.spec.ts:245` expected `Mention action completed.`
but observed the accepted action with `refresh is still required` and a
`WorkbookSurfaceRefreshError` message. The assertion expired after 5 seconds;
the scenario lasted 8,474 ms. Its retained group result and Playwright report
are under `browser-e2e-webserver-backed/browser-groups/<cohort>/`.
The referenced trace's selected network records show a successful final resolve,
Timeline reads with 200 responses, then a `net::ERR_ABORTED` Timeline read and a
later 200 Timeline read. No credentials, request bodies or raw trace were exported.
This supports overlapping/superseded read activity; it does not reconstruct the
exact application callback or socket-message ordering.

Core 03's `complete_mutating_action` contract and REQ-03-283 retain accepted
mention receipts after terminal refresh failure and permit only read recovery.
The acknowledged save-state rule likewise separates saved mutation from failed
follow-up reads. Source inspection confirms the mention owner retains its
receipt and reports `refresh=required` when reconciliation cannot establish current
accepted state. Existing mention owner/continuity tests cover this boundary.
The browser assertion assumed automatic presentation success without admitting
that specified recovery state. No product requirement or query acceptance guard
will be weakened to satisfy it. REQ-03-130/131's preserved raw mention and exactly
one entity remain binding.

Independent public `make harness-observability-check RESULTS_DIR=<exact-root>`
with `CHECK=events`, `timing`, and `diagnostics` each passed **0**,
`full_validity=0`; execution/resource coverage is complete, required accounting
available and PostgreSQL checkpoints observed. All **277** cleanup records
completed; final **2,319-file** secret scan passed; accounting is `empty`.
The separate complete allocation section has five settled generations, four
`terminal_borrower_release` and one actual `unhealthy_borrower`, with zero
omissions and zero reset-only generations. Public performance export passed **0**.
Owning elapsed **1,029,543.668425 ms**, CPU **3,756,898,962 us**, and kernel peak
**9,857,998,848 bytes** are descriptive failed-warm-up observations only.
No required resource remains unresolved.

`performance-experiment-exp06.json` preserves the failed first warm-up and marks
all three remaining warm-ups and sixteen planned matched samples omitted due to
the correctness failure. The immutable freeze remains unchanged. There are still
zero matched acceptance samples; neither lifecycle performance decision is closed.

F-016/R03-029/V-017/T-024 register **WF-11g**. After entry public Markdown lint
`r03-s011-corrective7-entry-lint-01` passes, the shared test fixture SHALL make
post-acknowledgement Timeline refresh failure deliberate in both rejected and
uncertain scenarios, assert that failure was exercised, then invoke the existing
explicit read-recovery affordance and require completed state. It SHALL retain
all raw values, navigation, exact-replay, version, receipt, single-entity and
mutation-count assertions. It SHALL use no polling retry loop, sleep, added
product deadline or automatic repeated mutation. This is a test correction under
existing authority, not a new runtime behavior or NLSpec amendment.
Focused and containing browser selections in both sources, existing mention-owner
assurance, type/projection checks, source preservation and closing tracker/lint
are required before a new experiment. S-007e remains open.

**E-130 — WF-11g baseline correction and preservation.**
Entry Markdown lint `r03-s011-corrective7-entry-lint-01` passed **6934/0**,
**9,520 ms**, before implementation. The only source change is in
`apps/web/e2e/mentions.recovery.spec.ts`: both recovery branches now deliberately
abort Timeline query reads after an acknowledged resolution, assert that at least
one read was blocked and that the accepted action retains refresh debt, then
activate the existing read-only recovery control and require completed state.
The rejected branch now explicitly follows that same contract. The final shared
assertion requires exactly two resolution requests and one entity creation.
Existing uncertain exact-body replay, receipt equality, rejected version advance,
raw form values, sheet navigation, responsive recovery bounds, focus recovery,
canonical entity count and null hostname assertions remain. There is no new
mutation retry, sleep, timeout, runtime change, selector change or schema migration.
The prior full run remains failed; this correction does not reconstruct its exact
callback ordering or claim a product defect.

Public commands ran with `CGO_ENABLED=0`, fresh IDs under
`.cartulary/test-results/r03-s011-corrective7-<suffix>`. Browser and owner test
commands also used `CARTULARY_HARNESS_CACHE_MODE=off`,
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.

| Suffix | Public command / exact selection | Native session / exit | Duration ms |
| --- | --- | --- | ---: |
| `format-b-01` | `make format` | 80396 / 0 | 8,273 |
| `prepare-b-01` | `make source-check CHECK=catalog` | 59699 / 0 | 1,671 |
| `mention-b-01` | `make test-slice OWNER=module.entities ROWS=module.entities.browser.timeline_mention_creation_recovery` | 48498 / 0 | 88,385 |
| `containing-b-01` | `make test-slice OWNER=module.entities ROWS=module.entities.browser.entity_merge_exact_recovery,module.entities.browser.timeline_mention_candidate_recovery,module.entities.browser.timeline_mention_creation_recovery` | 33858 / 0 | 109,665 |
| `owner-b-01` | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_mention_auto_resolution_undo_ownership_9e834f7b55,web.workbook.regression.timeline_mention_operation_recovery,web.workbook.regression.timeline_mention_reconciliation` | 46989 / 0 | 3,261 |
| `finalize-b-01` | `make agent-finalize` | 41621 / 0 | 26,951 |
| `types-b-01` | `make lint CHECK=types` | 82302 / 0 | 20,562 |
| `generated-b-01` | `make source-check CHECK=generated` | 91775 / 0 | 16,462 |
| `catalog-b-01` | `make source-check CHECK=catalog` | 75723 / 0 | 1,741 |

Each test independently passed public `make harness-observability-check
RESULTS_DIR=<exact-root> CHECK=all`, exit **0**, `full_validity=1`, complete
execution/resource diagnostics and required accounting. Isolated/containing runs
have 12 passing units and 23 completed cleanup records each; owner assurance has
three passing units and 12 completed cleanup records. All accounting scopes are
`empty`. Final secret scans passed with **97 / 99 / 37** files respectively.
Finalization preceded broader types/generated/catalog checks. Retained-run
maintenance was skipped because `RESULTS_DIR` was unset.

B7 is preserved in `diagnostic-baseline-source-exp07` with source
`sha256:78edf7613dc86e49912cc68751a80ce49a496ab7940d0cd4110c438ed7345e58`,
archive `c097ee8d8b6829f215feb027f8354b03ef73b602f855b6e638c32362b70be988`
(7,499 files / 48,084,651 bytes), bound to the successful catalog run. Its only
non-tracker difference from B6 is the authored recovery test. Guarded transfer to
the candidate and identical correction follow. WF-11g remains IN_PROGRESS until
both-source assurance/preservation and closing tracker/lint; no new performance
freeze or acceptance sample is yet admitted.

**E-131 — WF-11g complete; both-source read-recovery assurance passes.**
Guarded B7→C6 restoration preserved the controlling tracker. The exact E-130
recovery-test correction was applied to the candidate without other source edits.
Commands used `CGO_ENABLED=0` and fresh roots under
`.cartulary/test-results/r03-s011-corrective7-<suffix>`. Browser selections also
used cache `off`, diagnostics `basic`, and accounting `required`.

| Suffix | Public command / selection | Native session / exit | Duration ms |
| --- | --- | --- | ---: |
| `generate-c-01` | `make generate` | 37061 / 0 | 10,940 |
| `prepare-c-01` | `make source-check CHECK=catalog` | 56642 / 0 | 1,677 |
| `mention-c-01` | E-130 exact `module.entities.browser.timeline_mention_creation_recovery` owner row | 33511 / 0 | 93,425 |
| `containing-c-01` | E-130 exact three-row module.entities recovery selection | 51783 / 0 | 107,235 |
| `finalize-c-01` | `make agent-finalize` | 70462 / 0 | 27,406 |
| `types-c-01` | `make lint CHECK=types` | 80229 / 0 | 20,854 |
| `generated-c-01` | `make source-check CHECK=generated` | 43501 / 0 | 16,427 |
| `catalog-c-01` | `make source-check CHECK=catalog` | 6580 / 0 | 1,651 |

Both browser selections passed independent public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, exit **0**,
`full_validity=1`, complete execution/resource coverage, available required
accounting and observed PostgreSQL checkpoints. Each has 12 passing units and
23 completed cleanup records; both accounting scopes are `empty`. Final secret
scans passed with 97 and 99 files respectively. E-130 supplies the unchanged
mention owner, reconciliation and continuity assurance. Both treatments preserve
all original assertions and add explicit proof that recovery issues no extra
creation or resolution. Finalization preceded broader types/generated/catalog
validation. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
No golden refresh occurred.

C7 is preserved in `lifecycle-candidate-source-exp07`, source
`sha256:ef7158431082077b0bf72ebb55cb396882c16bf686641cb92ebfc9823b24bb21`,
archive `050f864dd828d196fd173dcbd106dfd847ed01e12847c460a05f76a59469b708`
(7,507 files / 48,120,290 bytes), bound to the successful candidate catalog run.
B7 remains E-130's immutable baseline. HEAD remains
`f42fc00b76ec4e296c2a379c09687e02ce669de7`. Manifest comparison proves each source's
only non-tracker delta from its experiment-06 counterpart is the same authored
recovery-test file; that file is byte-identical across treatments. Their 360-path
lifecycle treatment difference is unchanged. No runtime requirement, compatibility
fallback, product behavior, selection or scheduler policy changed in WF-11g.

Traceability is F-016 → R03-029 / existing Core 03 completion and read-recovery
contracts → controlled test read failure and explicit recovery → V-017 isolated,
containing, owner, types and projection checks → E-130/E-131 audits. Assurance
passes. Performance is unestablished. This test correction remains common if the
lifecycle change is rejected; it has no separate migration or runtime rollback.
The failed source and all omitted attempts remain preserved. No corrective
resource remains unresolved.

WF-11g is DONE. Closing public `make lint-markdown` uses
`r03-s011-corrective7-close-lint-01` and MUST pass before WF-11 resumes. Next entry
requires revalidated exact plans/environment, a new B7/C7 freeze and four excluded
warm-ups with independent source/graph/evidence/cleanup/scan validation. No matched
sample has run. Both workload decisions and final S-007e remain required.

**E-132 — Experiment 07 frozen after read-recovery test correction.**
WF-11g closing Markdown lint passed `r03-s011-corrective7-close-lint-01`,
**77871/0**, **8,810 ms**, before WF-11 resumed. Candidate browser/CI public
plan exports passed **70244/0** and **62244/0**. After guarded C7→B7 restoration,
public `make generate` passed `r03-s011-exp07-b-generate-01`, **27994/0**,
**10,950 ms**; unaccounted `make source-check CHECK=catalog` passed
`r03-s011-exp07-b-prepare-01`, **34262/0**, **1,647 ms**, all with `CGO_ENABLED=0`.
Baseline browser/CI public plan exports passed **9713/0** and **94211/0**.
All four plans exactly equal the corresponding experiment-06 plans.

E-130's B7 and E-131's C7 immutable source identities and archives are frozen.
The common recovery-test file is byte-identical; each source has only that
non-tracker delta from experiment 06. The 360-path lifecycle treatment difference
is unchanged. Browser membership remains 383 rows / 65 cohorts / B137→C76 units,
with 61 reset transfers. CI remains 2,279 rows / 96 cohorts / B600→C517 units,
with 83 transfers. No selection, scheduling, fairness, capacity, lane, worker,
retry, isolation or pooling change is admitted. System/toolchain, kernel,
repository/storage and machine Go-cache identities revalidate unchanged.
Actual runtime-graph binding remains required; only the previously declared
two Go CPU claims 4→9 and their semantic digests may differ from static defaults.

The immutable files are `performance-experiment-exp07-freeze.json`,
`performance-environment-exp07.json` and `exp07-treatment-difference.json`;
observations use `performance-experiment.json`. All earlier freezes, failed
attempts and omissions remain preserved. New excluded warm-ups are, in order,
`r03-s011-exp07-browser-warm-b`, `r03-s011-exp07-ci-warm-b`,
`r03-s011-exp07-browser-warm-c`, `r03-s011-exp07-ci-warm-c`.
Commands remain `make browser-e2e SUITE=functional` and `make ci`, with
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off`, fresh explicit
`CARTULARY_TEST_RUN_ID`, and `HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Dependencies/build caches remain warm; execution-result reuse stays disabled.
All four warm-ups MUST pass independent evidence, source/graph, cleanup and final
scan audits before matched sampling. No competing intentional workload is added;
structured polling stops during known quiet measurement.

Section 8.5/E-128's protocol is unchanged: four pairs maximum per workload,
odd baseline-first/even candidate-first; first-two benefits both <5% reject;
either >=20% permits early acceptance without omitted confidence claims;
otherwise the 95% benefit lower bound must exceed 5% and all five 99% guardrail
upper bounds must be <=5% using 10,000 paired bootstrap resamples, seed `20261021`
and frozen decimal arithmetic. Correctness/cleanup remains unconditional.
Both workloads receive their first two pairs unless correctness/cleanup prevents
execution. A rejected lifecycle restores its owner/projections/code together and
retains common corrections. Changed-source runs never use `COMPARISON=equivalent`.
Repairs require a new freeze; partial or failed attempts cannot disappear.
Supplementary queue-population, phase, allocation, retirement and identity
continuity reporting keeps its declared coverage and scope limits.

Focused assurance passes; full-CI acceptance is open. Zero matched acceptance
samples have run, and no lifecycle benefit or guardrail conclusion is established.
Freeze public `make lint-markdown` at `r03-s011-exp07-freeze-lint-01` passed
**1486/0**, **9,750 ms**, before the first warm-up began. Baseline browser warm-up
`r03-s011-exp07-browser-warm-b` is active in native **64115** under the frozen
command and inputs. Both workload decisions and S-007e remain required.

**E-133 — Experiment 07 baseline browser warm-up passed.**
The E-132 frozen command `r03-s011-exp07-browser-warm-b` passed **64115/0**,
graph duration **1,151,495 ms**, all **137** units passed. Source B7, graph
`sha256:cf92b0a6c12752fe0769de3c8cebf259693395f60c2db17ec6d4f966f33876d0`,
system/toolchain and 19-worker resolution match the freeze. Exact source
preservation revalidates before the next workload, excluding only the tracker.
Both previously repaired browser scenarios executed in this complete selection.

Independent public `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed **0**, `full_validity=1`, complete execution/resource coverage,
available required accounting and observed PostgreSQL checkpoints. All **276**
cleanup records completed, accounting is `empty`, and the final **2,296-file**
secret scan passed. Public performance and critical-path units/waits/allocations
exports each passed **0**. At `OFFSET=0 LIMIT=100`, complete populations are
80 units, 80 wait segments and one path-associated allocation; all have
`next_offset=null`, complete canonical/allocation coverage and zero omissions.
The full lifecycle section contains four settled allocations with actual
`terminal_borrower_release` reasons, zero omissions and no reset-only generation.
Exports are retained under the run-ID prefix in the continuation evidence folder.

| Excluded warm-up metric | Observation |
| --- | ---: |
| Owning-command elapsed | 1,164,127.179813 ms |
| Charged invocation CPU | 3,990,362,645 us |
| Kernel peak memory charge | 8,217,047,040 bytes |
| p95 ready-to-start delay | 1 ms |
| Cleanup interval union | 49,265 ms |

The complete path has graph queue **21 ms** and unit lifetime **1,150,954 ms**.
Summed recorded path phases are host admission **2,360 ms**, acquisition
**26,273 ms**, runner **1,092,911 ms**, release **26,665 ms**, publication
**1,864 ms**, and cache lookup/publication **7/2 ms**; uncovered lifetime is
**879 ms**. These observations are not exclusive additive wall-time causes.
Claims are not utilization; resource metrics retain the delegated Linux scope
and exclusions. This is an excluded warm-up, not performance acceptance or a
comparison with earlier sources.

Baseline CI `r03-s011-exp07-ci-warm-b` is active in native **38724**, using the
exact E-132 command, environment and selectors. Candidate warm-ups remain next after baseline
CI's terminal result, independent audits and physical settlement. All four
warm-ups precede matched sampling. Both workload decisions and S-007e remain open.

**E-134 — Experiment 07 failed CI warm-up settled; corrective gates registered.**
The frozen `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-exp07-ci-warm-b
CARTULARY_HARNESS_CACHE_MODE=off make ci HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required` returned **38724/2**, graph duration **3,633,691 ms**,
with **596 passed / 4 failed / 600** units. The exact root is
`.cartulary/test-results/r03-s011-exp07-ci-warm-b`. B7 source and E-132 system,
toolchain and selection identities match; runtime graph
`sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`
contains the frozen 2,279 rows and 600 units with only the previously permitted
Go CPU-resolution differences. `exp07-baseline-ci-runtime-differences.json`
retains that mapping. No source changed during the invocation.

Three root failures account for the four failed units:

| Root failure | Retained observation | Disposition |
| --- | --- | --- |
| `go:1aa28ab6cd1a86c2:8f1f46438669` | Native 3, `harness/fixture_error`; owned forced drop exceeded 20 seconds, `IPC/CheckpointDone`, checkpointer `IO/DataFileSync`, zero SQL blockers; subsequent setup rejected after quarantine | F-018; two incidentbundles rows infrastructure-failed, nine passed |
| `go:21743f28df257b2c:b12e48365b78` | Native 3, same forced-drop deadline/wait classification in a separate dedicated generation | F-018; three recovery rows infrastructure-failed, one passed |
| `module.evidence.browser.space_discovery_deadline_dcc019ff11` | Native 10, `product/test_assertion_failure`; test timed out at 60,207 ms while clicking disabled Preview; browser target finalizer propagates this failure | F-017; cohort has nine passed, one failed and 37 dependency-skipped rows |

The browser trace records a successful safe preview probe and metadata HTTP 200
before the 30,000-ms virtual-clock advance. It does not prove metadata admission
into the UI before that advance. The retained UI reports unavailable metadata
with an explicit retry; Preview is disabled, as REQ-03-248a requires. The test
asserts the discovery warning but lacks a metadata-admission precondition.
The final `page.unroute` error follows test timeout/context closure and is not
its initiating cause. Exact JavaScript callback order remains unavailable.
No product access defect or automatic-retry requirement is inferred.

The affected PostgreSQL generations are `7699526363903758376` and
`7699526520657629224`. Their typed checkpoint records are complete and include
sync phases of **25,608 ms / 3,779 files** and **21,512 ms / 4,621 files**.
These are recorded service-generation checkpoints, not proof that each belongs
exclusively to one drop or an attribution of host storage latency. Both failed
drops report **5 ms** admission and **20,000 ms** operation against a **19,999 ms**
effective operation budget. Existing admission/operation caps are 15/20 seconds;
no budget amendment is adopted here. Storage had 757 GiB available on the
repository filesystem after settlement; capacity exhaustion was not observed.

Public `make harness-observability-check RESULTS_DIR=<exact-root>` passed
`CHECK=events` **23306/0**, `CHECK=timing` **26519/0**, and
`CHECK=diagnostics` **26943/0**, each with `full_validity=0`.
Verbose execution/resources are truncated (665 resource omissions). PostgreSQL
checkpoint coverage is unavailable for four quiet-measurement services; affected
failure generations are observed. Required invocation accounting is available.
All **1,163** final cleanup records completed, accounting settlement is `empty`,
and the final **9,807-file** secret scan passed. This final disposal does not
reclassify failed eager per-test cleanup. All **85** browser allocations settled;
complete lifecycle coverage has zero omissions, **71 capacity-pressure**,
**13 terminal-borrower-release**, and **1 unhealthy-borrower** retirements;
**29** generations served only separately scheduled resets.

Public performance export **58395/0** and critical-path units **71327/0** pass.
Both root CI and explicit internal browser target exports retain the complete
80-unit path, 343 wait segments and 20 path-associated allocations. Wait pages
use offsets 0/100/200/300, checked for exhaustive deterministic continuation;
all final pages have `next_offset=null`. Exports remain under the exact run-ID
prefix in `.cartulary/remediation03/continuation/`. Path totals are graph queue
**612,734 ms** and unit lifetime **3,019,384 ms**, independent of pagination.
Owning elapsed is **3,716,791.222555 ms**, charged CPU **11,605,099,315 us**, kernel
peak charge **9,420,677,120 bytes**, p95 ready-to-start **3,116,448 ms**, and cleanup
union **156,474 ms**. These failed/excluded observations establish no benefit or
guardrail acceptance. Scope remains the declared Linux accounting boundary;
overlapping phases and blockers are not additive causal wall-time components.

`performance-experiment-exp07.json` preserves the failed attempt and marks both
unstarted candidate warm-ups and all 16 matched-sample attempts omitted.
The immutable freeze is unchanged. B7/C7 archives remain intact; no candidate
resource is unresolved. No matched performance sample has run. Retained-run
maintenance remains skipped because `RESULTS_DIR` is unset.

F-017/R03-030/V-018/T-025 register **WF-11h**, a common browser-test precondition
correction under existing Core 03 authority. It SHALL first pass public entry
Markdown lint at `r03-s011-corrective8-entry-lint-01`; then prove admitted metadata
before advancing virtual time, retain all original assertions, validate exact
and containing selections plus owner contracts in both sources, and record its
source/assurance receipt and closing lint before any next slice.
F-018/R03-031/V-019/T-026 register **WF-11i**, after that boundary, to resolve the
cleanup-budget failure without retries, durability weakening, scheduler changes
or an increased overall cleanup lifetime. Any behavioral amendment MUST precede
its projections. Both corrective gates and a fresh frozen comparison are required
before WF-11 resumes. S-007e and both workload decisions remain open.

**E-135 — WF-11h baseline correction and preservation.**
Entry Markdown lint `r03-s011-corrective8-entry-lint-01` passed **72240/0**,
**9,540 ms**, before executable edits. Existing Core 03 REQ-03-248a remains
unchanged. `apps/web/e2e/evidence.spec.ts` now requires both linked items to expose
admitted `data-evidence-metadata-state="ready"` before advancing virtual time.
This separates metadata admission from the deliberately stalled preview discovery.
The held request, 30,000-ms advance, indeterminate warning, visible list, explicit
preview, fresh preview frame, close, late settlement and no-late-preview assertions
remain. No product code, selector, retry, timeout or isolation declaration changed.

Commands ran with `CGO_ENABLED=0` and fresh roots
`.cartulary/test-results/r03-s011-corrective8-<suffix>`. Test selections also used
`CARTULARY_HARNESS_CACHE_MODE=off HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required`. Public owner guides were rediscovered.

| Suffix | Public command / exact selection | Native session / exit | Duration ms |
| --- | --- | --- | ---: |
| `format-b-01` | `make format` | 93205 / 0 | 8,238 |
| `prepare-b-01` | `make source-check CHECK=catalog` | 44608 / 0 | 1,648 |
| `discovery-b-01` | `make test-slice OWNER=module.evidence ROWS=module.evidence.browser.space_discovery_deadline_dcc019ff11` | 4557 / 0 | 86,008 |
| `containing-b-01` | `make test-slice OWNER=module.evidence ROWS=module.evidence.browser.linked_metadata_window,module.evidence.browser.linked_preview_recovery_6065aad8d8,module.evidence.browser.space_discovery_deadline_dcc019ff11,module.evidence.browser.timeline_linked_text_review_3f18dc6414` | 14610 / 0 | 112,186 |
| `metadata-b-01` | `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend.metadata_window_lifetime,module.evidence.frontend.metadata_window_presentation` | 72415 / 0 | 3,467 |
| `discovery-owner-b-01` | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_linked_evidence_preview_resolution_975344b0d0` | 99564 / 0 | 2,786 |
| `finalize-b-01` | `make agent-finalize` | 53528 / 0 | 27,265 |
| `types-b-01` | `make lint CHECK=types` | 58238 / 0 | 20,494 |
| `generated-b-01` | `make source-check CHECK=generated` | 76668 / 0 | 16,597 |
| `catalog-b-01` | `make source-check CHECK=catalog` | 40891 / 0 | 1,658 |

All four test selections passed independent public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, exit **0**,
`full_validity=1`, complete execution/resource coverage and available required
accounting. The browser selections each have 12 passed units and 23 completed
cleanup records; metadata/discovery owner checks each have two passed units and
11 completed cleanup records. All accounting scopes are `empty`. Final secret
scans passed with **97 / 100 / 28 / 26** files respectively. PostgreSQL checkpoint
coverage is observed for the browser tests and not observed for the service-free
owner checks. Finalization preceded broader types/generated/catalog checks.
Maintenance was skipped because `RESULTS_DIR` was unset; no golden refresh occurred.

B8a is preserved in `diagnostic-baseline-source-corrective8`, source
`sha256:65be45de79037e7a022fb397595e7cd3372f9f9099c7d86ee8ce9d00c40ce21f`,
archive `7145dcc1b1f713d9c5cd9dc3bdb16b8f2c903514223ffe8855af983ca0b2c96a`
(7,499 files / 48,096,113 bytes), bound to the successful catalog run. Its only
non-tracker delta from B7 is the seven-line authored browser-test correction.
WF-11h remains IN_PROGRESS until identical candidate assurance, preservation and
closing tracker/lint. WF-11i has not started; cleanup deadlines remain unchanged.
No new performance freeze or sample is admitted.

**E-136 — WF-11h complete; metadata-admission correction preserved in both sources.**
Guarded B8a→C7 restoration preserved the tracker and all immutable histories.
The exact E-135 seven-line correction was applied to the candidate. Commands used
`CGO_ENABLED=0` and roots `.cartulary/test-results/r03-s011-corrective8-<suffix>`;
test selections also used cache `off`, diagnostics `basic`, accounting `required`.
Exact selections are the E-135 commands with the corresponding candidate run ID.

| Suffix | Public command / selection | Native session / exit | Duration ms |
| --- | --- | --- | ---: |
| `generate-c-01` | `make generate` | 48874 / 0 | 11,130 |
| `prepare-c-01` | `make source-check CHECK=catalog` | 36679 / 0 | 1,655 |
| `discovery-c-01` | E-135 exact Space-discovery browser row | 9059 / 0 | 85,988 |
| `containing-c-01` | E-135 exact four-row linked metadata/preview selection | 14661 / 0 | 109,082 |
| `metadata-c-01` | E-135 exact two metadata owner rows | 88961 / 0 | 3,432 |
| `discovery-owner-c-01` | E-135 exact workbook discovery owner row | 14733 / 0 | 2,553 |
| `finalize-c-01` | `make agent-finalize` | 74226 / 0 | 27,655 |
| `types-c-01` | `make lint CHECK=types` | 51394 / 0 | 21,043 |
| `generated-c-01` | `make source-check CHECK=generated` | 92760 / 0 | 17,050 |
| `catalog-c-01` | `make source-check CHECK=catalog` | 91896 / 0 | 1,709 |

All four candidate test selections passed independent public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, exit **0**,
`full_validity=1`, complete execution/resource diagnostics and available required
accounting. Browser runs have 12 passing units and 23 completed cleanup records
each; metadata/discovery owner runs have two passing units and 11 completed
cleanup records each. Every accounting scope is `empty`. Final secret scans pass
with **97 / 100 / 28 / 26** files. Browser checkpoint coverage is observed;
service-free owner checks correctly report not observed. Finalization preceded
broader checks. Maintenance was skipped because `RESULTS_DIR` was unset.

C8a is preserved in `lifecycle-candidate-source-corrective8`, source
`sha256:1aba066a8da43c2bca569e4358b3e00c2898faba243b6a12ede582d557804ccb`,
archive `52793a85f63a1e0104beca542d099b3e7da7827390dd120d18e78b2b785d755d`
(7,507 files / 48,125,808 bytes), bound to the passing catalog run. E-135 preserves
B8a. Each source differs from its B7/C7 predecessor only in the same authored test,
excluding the tracker. The common file is byte-identical; the 360-path lifecycle
treatment difference is unchanged. HEAD remains `f42fc00b76ec4e296c2a379c09687e02ce669de7`.
No product behavior, owner requirement, schema, catalog selection, deadline,
scheduler policy, capacity, worker, retry, isolation or pooling setting changed.

Traceability: F-017 → R03-030 / existing Core 03 REQ-03-248a/220 → explicit UI
metadata-admission assertion → V-018 exact/containing/owner/type/projection checks
→ E-135/E-136 audits. Assurance passes. The historical callback order is not
claimed reconstructed, and no performance conclusion follows from these checks.
The correction remains common if lifecycle optimization is rejected; no runtime
migration or compatibility path is introduced. All earlier failures and omissions
remain retained, and no corrective resource is unresolved.

WF-11h is DONE for assurance. Closing public `make lint-markdown` uses
`r03-s011-corrective8-close-lint-01` and MUST pass before WF-11i begins. The next
slice investigates F-018 under the unchanged 15/20-second cleanup caps and MUST
adopt any accepted amendment before implementation. No new frozen comparison or
matched sample is admitted. Both workload decisions and S-007e remain open.

**E-137 — WF-11i owner adoption: one bounded cleanup deadline.**
WF-11h closing Markdown lint `r03-s011-corrective8-close-lint-01` passed
**97315/0**, **9,880 ms**, before this workstream began. C8a is the preserved
starting checkout; B8a remains intact. Public guides for `platform.postgres`,
`harness.browser`, `module.incidentbundles` and `module.recovery` were consulted.
The actual common cleanup tests belong to `harness.browser`, not the platform
service owner. No invocation remains active.

Adoption precedes projections: NLSpec Section 11.4 / TH-HARNESS-AC-007 now defines
one cleanup deadline, established before target admission as
`min(parent deadline if present, owner start + 35 seconds)`. Target admission
retains its 15-second cap; the one forced drop receives the remaining owner time;
lock release retains its 2-second cap and remains clamped to the owner deadline.
The independent 20-second operation timer is retired. Parent cancellation,
expired deadlines, exact ownership, same-session target exclusion, single-drop
execution, normal checkpoints, quarantine and independently retained cleanup
failures remain mandatory. No overall allowance, capacity, fairness, worker,
retry, isolation or pooling increase is authorized.

Rationale: E-134 records fast admission (5 ms) followed by an operation cutoff at
20 seconds, while affected service generations record checkpoint sync phases
longer than 20 seconds. One finite owner budget allows unused admission time to
serve the operation. It does not claim that the exact historical drop would have
succeeded, resolve host storage latency, or make those failed runs passing.
The earlier 15+20 phase decomposition is not a compatibility requirement.
Settlement still may fail at an exhausted deadline; that is recorded uncertainty,
not permission to return a lease or release capacity.

The shared PostgreSQL cleanup-observation component SHALL migrate v2→v3 with
operation-budget maximum 35,000 ms. Its browser-fixture, test-failure and service
scope references, attachment registry and generated validators SHALL migrate in
one cutover. Admission remains bounded at 15,000 ms; unadmitted operation values
remain zero. No dual reader or runtime toggle is admitted. Existing live
invocations are drained; historical archives retain their original epoch.

V-019 SHALL validate the two service-free `harness.browser` cleanup rows, then
real single-drop, target-coordination and barrier-observation rows; the affected
incidentbundles and recovery rows and their original containing Go selections;
public generated/schema/catalog and applicable owner contracts. Tests SHALL prove
absent/longer/shorter/expired/cancelled parents, unchanged deadline through phase
transition, remaining-budget diagnostics, no statement after cancellation, one
statement on all outcomes, independent failure causes and schema bounds. Native
exits, audits, scans and physical settlement remain required in both sources.
The correction SHALL be identical across baseline and candidate; finalization
and schema regeneration use Make. No new experiment may begin before both-source
assurance, preservation, completion receipt and closing lint.

Entry public `make lint-markdown` uses `r03-s011-corrective9-entry-lint-01` and
MUST pass before code/schema edits. F-018 → R03-031 / Section 11.4 and AC-007 →
shared cleanup context and schema v3 → V-019 → completion evidence remains the
traceability route. Assurance and performance remain open. S-007e still requires
current full CI and explicit selections; maintenance remains skipped because
`RESULTS_DIR` is unset.

### E-138 — WF-11i first unit attempt: test-interface defect

The first candidate implementation attempt is preserved in
`.cartulary/remediation03/continuation/lifecycle-candidate-source-corrective9-draft1.{json,tar.gz}`.
Its executable source is `sha256:a4c223927d147e82a31a7bbe0c7c2c9209662ec6b8e66bc531c2b466a2be379e`;
this failed source MUST NOT be reported as passing. Public preparation receipts:
`make format` (`r03-s011-corrective9-format-c-01`, native 38358/0, 8,019 ms),
`make generate` (`r03-s011-corrective9-generate-c-01`, 86734/0, 11,190 ms),
and `make source-check CHECK=catalog` (`r03-s011-corrective9-prepare-c-01`,
78292/0, 1,675 ms). Entry lint `r03-s011-corrective9-entry-lint-01`
completed 76616/0 in 9,730 ms before implementation.

Command: `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.service_free_contract_assertions_for_postgres_cl_83b8534255,harness.browser.integration.postgre_sql_cleanup_rejects_invalid_input_before_3f68344cda HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`,
with `CGO_ENABLED=0`, cache off and run ID `r03-s011-corrective9-unit-c-01`.
Root `.cartulary/test-results/r03-s011-corrective9-unit-c-01`; native 4734/2;
graph `sha256:b6557065ae6017bfaaea12dd83dc1831edd29a8f22ce48453c149a52ec58ad13`,
1,506 ms. The service-free row passed; the invalid-input row failed in three
new schema-parity subcases. The test incorrectly called a shared component as
a public native schema root. Native component validation and the deadline,
virtual-time, cancellation and one-statement assertions passed. The repair SHALL
exercise the component through the actual Go diagnostic envelope and test the
JavaScript consumer's budget bounds and retired component registration. It SHALL
NOT expand the native root registry to satisfy the mistaken test.

Independent public diagnostics/events/timing checks exited zero with
`full_validity=0`; execution/resource coverage was complete, PostgreSQL observation
not applicable, accounting available and empty, all 10 cleanup results completed,
and the terminal retained-secret scan passed across 18 files. This is a settled
failed assurance attempt, not a performance sample. The repaired source requires
a fresh run. WF-11i remains active; E-137 acceptance and common-source requirements
are unchanged.

### E-139 — WF-11i consumer fixture correction

`make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`
with the declared cgo-disabled, cache-off profile failed at
`.cartulary/test-results/r03-s011-corrective9-consumer-c-01`: native 26328/2,
63,393 ms; graph `sha256:7c6f8f65ce4eed568b19d6cfaa4475c71bdf1c2f325d47fcf719cca73ac55696`.
Its source `sha256:2ab7ad6f45da42a053d4e004dfa7aa4d609856babe618e21a53a20b99034b51f`
is preserved as `lifecycle-candidate-source-corrective9-draft2` (archive SHA-256
`544acc7e843528c9761aa68e4374444e95fcb48d3010f1d26e5a9058cdd9f396`).

The new JavaScript positive fixture declared rejected admission and zero
operation duration/budget but retained sampled operation waits. The current
schema correctly rejected that contradiction. The fixture SHALL use explicit
`not_observed` with null snapshots; a separate negative case SHALL preserve
rejection of sampled waits without admission. No production reader relaxation
is authorized or made. Other command-surface cases passed. Public diagnostic
validation exited zero with `full_validity=0`, complete execution/resources,
accounting available and empty; all 11 cleanup results completed and the retained
secret scan passed. This failed attempt remains excluded from performance.

Before this consumer failure, the corrected service-free rows, live cleanup,
affected incidentbundles/recovery rows and their original containing selections
passed; the complete native receipts follow in the WF-11i completion record.
The outcome/deadline tests were additionally moved into virtual time to avoid
host-scheduling-dependent timing assertions; their fresh unit check passed.
WF-11i remains active pending the repaired consumer check and both-source closure.

### E-140 — WF-11i candidate assurance and preservation

The candidate implements E-137 with one cleanup context shared by admission,
one forced drop and bounded settlement. Native observation bounds and schema v3
migrate together. Virtual-clock tests prove unused admission time remains usable
and an expired owner cannot be replenished. The public Go diagnostic envelope
and JavaScript consumer cover budget bounds and contradictory observations.
No production schema-root expansion, retry or cleanup success substitution was made.

All test commands below use `CGO_ENABLED=0`, `CARTULARY_HARNESS_CACHE_MODE=off`,
`HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`; each run ID is its
listed suffix prefixed by `r03-s011-corrective9-`, under `.cartulary/test-results/`.
Execution uses `make test-slice OWNER=<owner> ROWS=<selection>`.
The exact selected inputs and source/graph identities are additionally retained
in `.cartulary/remediation03/continuation/corrective9-assurance-c.json`.

| Selection | Owner | Exact rows |
| --- | --- | --- |
| `unit-c-02` | `harness.browser` | `harness.browser.integration.postgre_sql_cleanup_rejects_invalid_input_before_3f68344cda,harness.browser.integration.service_free_contract_assertions_for_postgres_cl_83b8534255` |
| `live-c-01` | `harness.browser` | `harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277,harness.browser.integration.postgres_owned_barrier_observation,harness.browser.integration.postgres_single_drop` |
| `incidents-c-01` | `module.incidentbundles` | `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8,module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413` |
| `incidents-containing-c-01` | `module.incidentbundles` | `module.incidentbundles.integration.incident_bundle_descriptor_reads_reject_singleto_ea18aa4fe5,module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8,module.incidentbundles.integration.incident_bundle_export_jobs_use_the_incident_sco_d94a8de941,module.incidentbundles.integration.incident_bundle_export_uses_the_common_job_resou_a1e51b82d7,module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755,module.incidentbundles.integration.incident_bundle_import_upload_envelope_failures_bf5449e3d4,module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413,module.incidentbundles.integration.incident_bundle_retired_version_is_rejected_with_1ec7e171fe,module.incidentbundles.integration.layout_legacy_conversion_replay,module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2` |
| `recovery-c-01` | `module.recovery` | `module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085,module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a,module.recovery.integration.selecting_a_retained_backup_set_missing_a_requir_d6af667fba` |
| `recovery-containing-c-01` | `module.recovery` | `module.recovery.integration.reset_preserves_crypto_format_identity,module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085,module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a,module.recovery.integration.selecting_a_retained_backup_set_missing_a_requir_d6af667fba` |
| `consumer-c-02` | `harness.command_surface` | `harness.command_surface.behavior.public_registry_parity` |

| Run suffix | Native session/exit | Graph ms | Passed units | Completed cleanup | Scan files |
| --- | --- | --- | --- | --- | --- |
| `unit-c-02` | 23975/0 | 2,131 | 1/1 | 10 | 21 |
| `unit-c-03` | 56087/0 | 1,344 | 1/1 | 10 | 18 |
| `live-c-01` | 85787/0 | 15,244 | 4/4 | 15 | 36 |
| `incidents-c-01` | 60473/0 | 42,629 | 3/3 | 14 | 36 |
| `incidents-containing-c-01` | 30855/0 | 60,384 | 3/3 | 14 | 45 |
| `recovery-c-01` | 3445/0 | 51,465 | 3/3 | 14 | 33 |
| `recovery-containing-c-01` | 88703/0 | 34,062 | 3/3 | 14 | 34 |
| `consumer-c-02` | 72785/0 | 65,182 | 2/2 | 11 | 173 |

All eight passing attempts have public `CHECK=all` exit zero and
`full_validity=1`, complete execution/resource diagnostics, empty invocation
accounting scopes and passing retained-secret scans. PostgreSQL checkpoint
observations are present for live/service selections and not applicable to the
service-free/consumer selections. E-138 and E-139 retain both failed attempts.

Source progression is explicit: unit-c-02, live and incident selections use
`sha256:c68305fb6262861220b1bbc30255b78e5fc7768f2efd557bd778cd1de75d4602`;
unit-c-03 and recovery use
`sha256:2ab7ad6f45da42a053d4e004dfa7aa4d609856babe618e21a53a20b99034b51f`
after virtualizing outcome assertions; consumer-c-02 uses
`sha256:d3c429416e3de8233e7d54fd22786b9acb5729f1aa22ce0d0fc9581099cf09aa`
after correcting the JavaScript fixture. These last two revisions change tests
only; the production correction and generated contracts are identical. Earlier
passes retain their actual source identity and are not relabeled as current runs.

Public format receipts: `format-c-02` 45776/0, 8,315 ms and `format-c-03`
61413/0, 7,973 ms. `make agent-finalize` (`finalize-c-01`) completed 64389/0
in 27,390 ms before broader verification. `make source-check CHECK=generated`
(`generated-c-01`) completed 85227/0 in 16,405 ms; `CHECK=schema`
(`schema-c-01`) 58782/0 in 6,386 ms; `make lint CHECK=go` (`go-c-01`)
10644/0 in 14,937 ms; `CHECK=catalog` (`catalog-c-01`) 44713/0 in 1,642 ms.
All use the same run prefix and cgo-disabled profile. Retained-run maintenance
was skipped because `RESULTS_DIR` was unset.

The immutable `lifecycle-candidate-source-exp08` archive has SHA-256 `1eb054d2cc9b612a59d823b604e3a540aa40958135f628d904a21d145545c361`
and binds `sha256:d3c429416e3de8233e7d54fd22786b9acb5729f1aa22ce0d0fc9581099cf09aa` through `catalog-c-01`.
Baseline application SHALL first adopt the same owner clauses, then apply only
identical common authored changes and regenerate through Make. Candidate lifecycle
owner clauses and projections MUST NOT enter that baseline. WF-11i remains active
pending baseline assurance and closing tracker/lint; performance and final CI remain open.

### E-141 — WF-11i complete: shared cleanup deadline, both sources

WF-11i is DONE for the adopted bounded cleanup correction and focused assurance.
F-018 → R03-031 / Section 11.4 and AC-007 → shared cleanup context, native
observation validation and schema v3 → V-019 → E-138–E-141 is the closure route.
Current full-CI assurance remains a mandatory next-stage gate; historical failed
drops retain their failures and are not claimed to have succeeded under this fix.

The baseline first adopted the same owner amendment, then received the identical
common authored patch. Standalone browser reset behavior remains in the baseline.
Public `make generate` (`r03-s011-corrective9-generate-b-01`) completed 75706/0
in 10,860 ms; cgo-disabled `make source-check CHECK=catalog` preparation
(`prepare-b-01`) completed 24769/0 in 1,625 ms before accounted tests.

All baseline test commands use the E-140 exact owner/row selections, command and
environment, with `-b-01` suffixes below. Run roots are
`.cartulary/test-results/r03-s011-corrective9-<suffix>`.
Exact source/graph identities and inputs are preserved in
`.cartulary/remediation03/continuation/corrective9-assurance-b.json`.

| Run suffix | Native session/exit | Graph ms | Passed units | Completed cleanup | Scan files |
| --- | --- | --- | --- | --- | --- |
| `unit-b-01` | 89247/0 | 1,369 | 1/1 | 10 | 18 |
| `live-b-01` | 55434/0 | 14,660 | 4/4 | 15 | 36 |
| `incidents-b-01` | 79564/0 | 51,619 | 3/3 | 14 | 36 |
| `incidents-containing-b-01` | 29595/0 | 63,285 | 3/3 | 14 | 45 |
| `recovery-b-01` | 83430/0 | 33,685 | 3/3 | 14 | 33 |
| `recovery-containing-b-01` | 45100/0 | 39,478 | 3/3 | 14 | 34 |
| `consumer-b-01` | 10805/0 | 66,567 | 2/2 | 11 | 173 |

All seven runs passed independent public `CHECK=all` with `full_validity=1`,
complete execution/resources, empty accounting scopes and passing terminal scans.
Live/service selections observed PostgreSQL checkpoints; service-free/consumer
selections correctly report not observed. All declared cleanup completed.

`make agent-finalize` (`finalize-b-01`) completed 55546/0 in 27,397 ms before
broader verification. `make source-check CHECK=generated` (`generated-b-01`)
completed 11252/0 in 16,495 ms; `CHECK=schema` (`schema-b-01`) 21849/0 in
6,372 ms; `make lint CHECK=go` (`go-b-01`) 55783/0 in 11,818 ms; and
`CHECK=catalog` (`catalog-b-01`) 16851/0 in 1,672 ms. These use the same run
prefix and `CGO_ENABLED=0`. Maintenance was skipped because `RESULTS_DIR` was unset.

| Preserved treatment | Executable source | Archive SHA-256 |
| --- | --- | --- |
| Baseline, `diagnostic-baseline-source-exp08` | `sha256:d0fb65ed1ff660dd14e83d3c337b8ba5621d9242cf1e00c1f7a4af9950866aee` | `495929f2a30b066ef8309154d639622fcd025552570fdae77b003806dbd52422` |
| Candidate, `lifecycle-candidate-source-exp08` | `sha256:d3c429416e3de8233e7d54fd22786b9acb5729f1aa22ce0d0fc9581099cf09aa` | `1eb054d2cc9b612a59d823b604e3a540aa40958135f628d904a21d145545c361` |

`exp08-treatment-difference.json` confirms that the existing 360-path treatment
difference remains, with the generated cleanup-validator path migrated v2→v3.
The shared validator body differs only by registry-local generated identifiers;
all invariant common changes are identical. The six common authored files and
owner replacement blocks are preserved in `repair9-common/manifest.json`.
Generated projections were produced by Make; no candidate lifecycle files entered
the baseline. Both sources retain the E-136 browser metadata-precondition repair.

Assurance decision: accept the bounded cleanup correction. Performance decision:
not evaluated by this corrective workstream; zero matched samples have run.
Migration: current cleanup references use v3 atomically, without a dual reader or
compatibility toggle; no live invocation crossed the cutover. Rollback of the
lifecycle candidate SHALL retain this common correctness correction. Failed
cleanup still quarantines uncertain ownership, even if later service disposal
succeeds. Host checkpoint latency remains an external source of possible bounded
cleanup failure; this change does not claim to eliminate it.

Closing public `make lint-markdown` uses `r03-s011-corrective9-close-lint-01`
and MUST pass before resuming WF-11. The next entry requires freshly frozen
baseline/candidate plans, exact populations, environment and treatment identities;
all four excluded warm-ups and independent evidence/settlement audits MUST pass
before matched sampling. The four-pair, first-two-pair and rollback rules remain
unchanged. S-007e final validation and handoff remain open.

### E-142 — WF-11 experiment 08 frozen before warm-ups

WF-11i closing Markdown lint `r03-s011-corrective9-close-lint-01` completed
44799/0 in 16,460 ms before this continuation. Experiment 08 is frozen at
`2026-10-22T18:32:37.983947+00:00` in
`.cartulary/remediation03/continuation/performance-experiment-exp08-freeze.json`
(SHA-256 `467c360bd8db47de228ea4e42b349deebbc4b9399f403c9a20a6aaf1888b0b04`).
The mutable attempt ledger is `performance-experiment.json`; all earlier frozen
experiments and their failed/omitted attempts remain preserved. E-141 identifies
both immutable source archives. No performance sample has yet run.

| Workload | Exact public execution | Selected rows | Baseline / candidate units | Reset operations transferred |
| --- | --- | --- | --- | --- |
| Browser | `make browser-e2e SUITE=functional` | 383 | 137 / 76 | 61 |
| CI | `make ci` | 2,279 | 600 / 517 | 83 |

Both treatments SHALL retain `CGO_ENABLED=0`,
`CARTULARY_HARNESS_CACHE_MODE=off`, `HARNESS_DIAGNOSTICS=basic` and
`HARNESS_ACCOUNTING=required`, warm dependency/build caches and exact fresh run
IDs from the frozen ledger. The unchanged selected rows, target plans, input
files, environment and treatment-difference proof are hash-bound by the freeze.
No ranking, fairness, capacity, lane, worker, retry, eligibility or pooling
change is admitted. Cleanup schema v3 and the E-136 browser precondition are
common corrections, not candidate-only treatments.

Public plans use `make target-plan TARGET=<browser-e2e|ci> JSON=1` with
`INPUTS=.cartulary/remediation03/continuation/performance-browser-inputs.json`
or `performance-ci-inputs.json`, respectively. Baseline exports completed
2889/0 and 78821/0; candidate exports 69648/0 and 81164/0. Two initial inline-JSON
input attempts failed preflight with exit 2 and launched no product work;
`exp08-plan-preflight.json` and the empty-output receipts preserve those attempts.
All four successful plan models equal their experiment-07 models; source and
warm-up evidence nevertheless MUST be established anew.

Candidate restoration was followed by public `make generate`
(`r03-s011-exp08-plan-generate-c`, 60354/0, 10,870 ms) and
`make source-check CHECK=catalog` (`plan-prepare-c`, 81335/0, 1,721 ms).
Baseline restoration then used `prewarm-generate-b` (1432/0, 10,910 ms) and
`prewarm-prepare-b` (58057/0, 1,667 ms), under the same experiment prefix and
cgo-disabled preparation profile. These are preparation, not samples.

Warm-up order SHALL be browser baseline, CI baseline, browser candidate, CI
candidate, each excluded once. All four MUST pass independent evidence audits,
source/runtime-graph binding, retained-secret scans and physical settlement
before matched sampling. Runtime Go worker budget 19 retains the previously
frozen two-unit static CPU-claim expansion (4→9) and corresponding digests;
no other runtime/static graph difference is permitted. The two unit IDs remain
`go:38ac079360dd6fb6:ba2f66716e6a` and `go:5f7c20b93d80a0f3:3fa1ec86e540`.

Each workload SHALL receive the first two matched pairs unless correctness or
cleanup failure blocks execution. Both improvements strictly below 5% reject;
either at least 20% permits early acceptance without omitted confidence claims;
otherwise four pairs are required. Odd pairs are baseline-first and even pairs
candidate-first. Full analysis retains 10,000 paired bootstrap resamples, seed
20261021, decimal precision 28/HALF_EVEN, one-sided 95% benefit lower bound >5%,
and five one-sided 99% guardrail upper bounds ≤5%. Owning elapsed, charged CPU,
kernel peak memory, p95 ready-to-start delay and cleanup duration retain their
specified scopes. Queue-population changes SHALL be accompanied by corresponding
cohort-wait analysis. Changed-source runs SHALL be validated separately, never
as equivalent comparisons. Any workload rejection restores lifecycle owner,
implementation and projections together while preserving accepted diagnostics
and common correctness repairs.

Entry tracker lint uses `r03-s011-exp08-entry-lint-01` before the first warm-up.
Assurance during warm-up, performance acceptance and S-007e final completion
remain separate and open. Maintenance remains skipped because `RESULTS_DIR`
is unset.

### E-143 — Experiment 08 baseline browser warm-up passed, excluded

Entry Markdown lint `r03-s011-exp08-entry-lint-01` completed 85084/0 in
9,600 ms. The frozen baseline command
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-exp08-browser-warm-b CARTULARY_HARNESS_CACHE_MODE=off make browser-e2e SUITE=functional HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`
completed native 75031/0 at `.cartulary/test-results/r03-s011-exp08-browser-warm-b`.
All 137 units passed; graph duration was 1,158,526 ms. Source and graph identities
match the baseline freeze. Public `CHECK=all` passed with `full_validity=1`, complete
execution/resources and observed PostgreSQL checkpoints. All seven public
performance/critical-path exports exited zero; invocation-root and explicit
`browser-e2e-webserver-backed` units/waits/allocations pages are complete with
`next_offset=null`. Three bounded progress snapshots during this excluded warm-up
reported ordinary execution; they are retained and do not supply acceptance samples.

Owning elapsed was 1171092.402861 ms, charged CPU
3984507423 µs, kernel peak memory 12570312704
bytes, ready-to-start p95 1 ms, and cleanup union
47154 ms. These are excluded warm-up observations with
E-142 accounting scope, not a measured lifecycle benefit. The complete 80-unit
path reports 15 ms graph queueing and 1,158,009 ms unit lifetime. Phase-duration
sums (not exclusive wall-clock attribution) and uncovered lifetime are retained
in the attempt ledger and complete public exports.

All 276 cleanup results completed, invocation accounting settled empty, and the
retained-secret scan passed across 2,296 files. Lifecycle coverage is complete:
four allocations, all retired for `terminal_borrower_release`, all settled,
zero omissions and zero reset-only generations. The baseline browser runtime
graph is bound; the baseline CI and both candidate warm-ups remain required.
Zero matched samples have run. Next is `r03-s011-exp08-ci-warm-b` on the same
frozen baseline. This receipt does not establish CI assurance or S-007e completion.

### E-144 — Experiment 08 cancelled after endpoint failure; WF-11j registered

The user instructed early termination and repair. SIGINT was sent through the
verified native session 12677. The exact command was
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-exp08-ci-warm-b CARTULARY_HARNESS_CACHE_MODE=off make ci HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Its outer exit was 130; retained root is
`.cartulary/test-results/r03-s011-exp08-ci-warm-b`. Baseline source remains
`sha256:d0fb65ed1ff660dd14e83d3c337b8ba5621d9242cf1e00c1f7a4af9950866aee`;
runtime graph is
`sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818`.
The terminal graph reports cancelled, 95 passed / 1 failed / 504 cancelled of
600 units, 1,179,008 ms. No passing CI or acceptance sample is claimed.

The pre-cancellation failure is unit `go:997f964b43d2975d:73fc5074240a`,
infra / service_readiness_timeout. Its four module.extensions rows did not
execute. Published scope
`_shared/test-services/8be08a78482f49b393f8cc58/service-scope.json` reports one
failed 35-second startup attempt and completed failed-start physical settlement.
Typed readiness v5 shows internal final-server readiness at 2,975 ms, an internal
listener present at failure, unchanged published bindings and exact container
identity, but no successful SQL connection. The 77 probes include three timeouts,
one closed connection and 73 refused connections; the independent final
published TCP dial also refused. These observations distinguish internal service
readiness from published client transport. They do not establish the transport
failure's root cause or justify a larger deadline.

Public diagnostics validation (native 87119/0) and performance export (66379/0)
completed. Diagnostic execution/resources are complete; full_validity=0 remains
appropriate for this cancelled run. All 229 cleanup results completed; invocation
accounting settled empty. The retained-secret scan passed across 5,358 files.
Owning elapsed was 1208829.420362 ms, charged CPU 3405833141 µs and kernel peak
7714185216 bytes; these interrupted observations are excluded from acceptance.
The immutable experiment-08 closed ledger preserves the failed attempt and all
unstarted attempts as omitted. No owned-resource uncertainty remains in the
reported terminal evidence. WF-11j / F-019 / R03-032 / V-020 / T-027 is registered;
Entry Markdown lint `r03-s011-corrective10-entry-lint-01` passed
(native 98702/0, 10,080 ms) before implementation. No runtime amendment or
speculative repair has yet been adopted.

### E-145 — User amendment: assumed CI elapsed baseline

The user instructed: skip CI baseline timing and assume 60 minutes. The user
then explicitly selected elapsed-only CI performance acceptance against that
assumption, descriptive resource reporting, and mandatory correctness/cleanup;
browser acceptance remains unchanged. Section 8.5 is the current binding
protocol. The reference is exactly 3,600,000 ms; it is not a measured run and
carries no source, graph, CPU, memory, queue or cleanup observations. Existing
failed and cancelled baselines retain their original status. No additional CI
baseline timing run or baseline warm-up SHALL be launched for this experiment.
A fresh freeze after corrective assurance SHALL use this protocol. This amendment
changes internal remediation acceptance only, not runtime NLSpec behavior.

### E-146 — WF-11j endpoint investigation and baseline assurance

E-144 public event and timing audits also passed (native 33830/0 and 65356/0),
without full_validity. No additional cancelled-run failure is inferred from the
504 cancelled units. Both baseline reproduction commands used CGO_ENABLED=0,
cache off, basic diagnostics and required accounting. Native 31117/0 ran
`make test-slice OWNER=module.extensions ROWS=module.extensions.integration.admission_preflight_v6,module.extensions.integration.managed_service_assertions_for_state_admission_m_55b90ea265,module.extensions.integration.network_flow_state_v4,module.extensions.integration.state_concurrent_admission`
at `r03-s011-corrective10-repro-b-01`: 3/3 units passed in 15,863 ms, 14 completed
cleanup records and 34 scanned files. Native 25102/0 ran
`make test-slice OWNER=module.extensions SCOPE=services` at
`r03-s011-corrective10-containing-b-01`: 18/18 units passed in 129,600 ms, 41
completed cleanup records and 162 scanned files. Each independent public
CHECK=all audit passed with full_validity=1, complete execution/resources and
observed checkpoints. Both scans passed. These runs retain exact baseline-08
source identity; the first run's graph is
`sha256:bb69b62d4dff2763c3ef2fd46623132b1d41911cff7eb70a0cc3672057ad253b`;
the containing graph is
`sha256:fadc8f5ad336e76c9b08eb01e7921d371a785988aef27870b8cee85765cf7730`.

Read-only source inspection found the fixture owner selecting Docker's current
published endpoint, preserving its exact container identity and failing closed
when SQL connectivity never succeeds. No cached-inspection substitution or
wrong-owner removal was established. Docker Desktop VM logs retained seven WSL
relay accept-stall observations within the failed startup interval. The redacted
context and file digest are retained in `corrective10-host-context.json`.
The Windows backend log covering the exact endpoint failure had rotated out.
The relay observations have no exact endpoint binding and SHALL NOT be treated
as the proven cause. The upstream [Docker report on intermittent WSL published ports](https://github.com/docker/for-win/issues/14195)
is external corroborating context only, not a reproduction of this run.

**Investigation decision:** no source repair is justified by the current evidence.
Increasing timeouts, retrying failed SQL-readiness generations, selecting a
fallback endpoint, changing network modes or adding Docker-specific publication
workarounds would weaken attribution or add unsupported compatibility behavior.
No such change was made. The historical acquisition failure remains unresolved
at the host transport layer; passing focused runs do not erase it or establish
full CI assurance. Candidate source validation and mandatory current full-CI
assurance remain open. This is not a claim that the underlying fault was fixed.

After baseline settlement, the exact source switch restored C8 from its immutable
archive while preserving this tracker. Public generation
`r03-s011-corrective10-generate-c-01` passed (native 9567/0, 11,500 ms). Runtime
preparation and matching candidate assurance precede the workstream disposition.

### E-147 — Requested stability plan; performance work awaits qualification

The user asked to plan around the issue and find a stable solution. Section 7.7
adds serial WF-11j-a–d with explicit transport, publication, stability and handoff
gates. It preserves the required Docker Desktop reference profile in NLSpec
Sections 13.7/14 and does not adopt a native-Linux substitution. The proposed
publication candidate is conditional on a controlled reproducer and exact
security/ownership proof; the alternative is a qualified unchanged path or a
concrete environment-remediation decision. No historical failure is declared fixed.

The current checkout is exact lifecycle candidate C8 at executable source
`sha256:d3c429416e3de8233e7d54fd22786b9acb5729f1aa22ce0d0fc9581099cf09aa`;
HEAD remains `f42fc00b76ec4e296c2a379c09687e02ce669de7`. Public catalog preparation
`r03-s011-corrective10-prepare-c-01` passed native 96674/0, 1,773 ms. Candidate
reproduction, stability qualification and all acceptance samples are unstarted.
No test invocation remains active. The user-requested plan edits only this
controlling tracker; no new runtime behavior, schema or product change was made.
Existing source archives and E-073 remain preserved. Retained-run maintenance was
not invoked; RESULTS_DIR remains unset because no compatible successful full warm
check root exists. Planning lint uses
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-stability-plan-lint-01 make lint-markdown`;
the first lint passed native 16157/0 in 17,950 ms. The final receipt and
owner-before-projection clarification are verified with
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s011-stability-plan-lint-02 make lint-markdown`.
Overall continuation and mandatory current CI assurance remain open.

### E-148 — WF-11j-a entered; publication-observation owner adopted

The user explicitly authorized implementation of Section 7.7. Source preservation
verification matches every C8 archive path except this intentionally maintained
tracker; HEAD remains f42fc00b76ec4e296c2a379c09687e02ce669de7. No invocation was active.
The test-operations skill and root/tooling/backend instruction chain were reloaded;
public command and harness.browser owner discovery are retained in the continuation
directory. E-147 final planning lint passed native 85305/0 in 8,920 ms.

Before projection or implementation changes, NLSpec Section 11.4 adopts readiness
v6's bounded, non-mutating final Docker create-request publication observation.
This closes an actual diagnostic gap: provider-normalized bindings were previously
known only from source inspection and later container inspection. Existing failure
observations, security checks, allocation ownership, deadlines and retry rules are
retained. R03-032 → Section 11.4 → testcontainersx/pgtest observation → V-020 is the
traceability route. Diagnostic cutover is common to both treatments; old artifacts
remain historical. Entry lint precedes implementation; no publication behavior
change or transport repair has been accepted.

### E-149 — WF-11j-a characterization completed

**Source and authority.** HEAD remains `f42fc00b76ec4e296c2a379c09687e02ce669de7`;
current executable source is
`sha256:1bd52588e175c06311e9c4bf34036697c211ba90b9003aed6f9d9dac8d023c8b`.
E-148 entry lint passed native **42848/0**, 16,820 ms. R03-032 / F-019 → NLSpec
Section 11.4 → `testcontainersx.GenericContainerObserved`, the pgtest publication
observer and explicit qualification test → V-020 is the implementation trace.
The owner amendment preceded implementation. No scheduler, publication, deadline,
retry, capacity, isolation or pooling behavior changed.

Readiness diagnostic **v6** records the final provider-normalized Docker request
through the supported client API. It bounds and redacts requested/effective ports,
binding scope/kind and publish-all state, with explicit unavailable coverage.
Schemas, attachments and generated Go/JavaScript projections migrated together;
v5 remains historical with no translation reader. Existing installed-only image
policy, accounting hooks and partial-container cleanup ownership are preserved.
The SDK error-classification dependency was already pinned; its existing
`github.com/containerd/errdefs v1.0.0` requirement is now direct, without a version
or resolved-module change.

The independently selectable qualification row was allocated through
`make author-test-row-id FAMILY_ID=harness.browser.integration
CLAIM='Published PostgreSQL transport qualification across fresh owned lifetimes'
SELECTOR_KEY='go:./internal/testutil/pgtest:TestPublishedPostgresTransportQualification'`:
`harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9`.
It is explicit-only, fixture-none, exclusive-process, with the existing
`postgres_catalog_isolated` resource profile. The compiled selection reserves
four PostgreSQL and CPU lanes. Each roster contains 16 serial lifetimes, eight
waves of four concurrent lifetimes, and eight competing pairs. In a competing
pair A reaches authenticated readiness, B is created, then A retires while B
starts. Every lifetime retains its own existing suite-service journal and summary;
native log truncation cannot silently remove qualification evidence.

**Environment and freeze.** Docker Desktop 4.75.0 (227598), Engine 29.5.2, WSL
2.7.3.0, kernel 6.6.114.1-microsoft-standard-WSL2; default provider bridge network.
The exact pinned PostgreSQL image exposes only `5432/tcp`. The frozen control
request has that effective port, one unspecified-address ephemeral binding and
`publish_all_ports=false`. Inventory, rosters and exclusions are retained in
`stability-a-characterization-freeze*.json`; `stability-qualification-plan.json`
records the public target plan. These files are internal evidence, not executable
harness inputs or alternative behavioral owners.

All run IDs below are under `.cartulary/test-results/`. Test commands use
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off`, `HARNESS_DIAGNOSTICS=basic` and
`HARNESS_ACCOUNTING=required`. Public discovery and all native exits are retained.

| Run / exact selection | Native receipt | Result |
| --- | --- | --- |
| `r03-stability-a-unit-03`; `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.boundary_support.installed_only_images` | 43100/0; 2,456 ms | Negative transport matrix, request bounds/redaction, provider delegation, installed-only policy and actual SDK 404/409/500 classification pass |
| `r03-stability-a-live-03`; `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9` | 17276/0; 100,976 ms | 64 unique allocations, 64 authenticated readiness passes, 64 exact removal confirmations; complete individual records |
| `r03-stability-a-affected-c-01`; `make test-slice OWNER=module.extensions ROWS=module.extensions.integration.admission_preflight_v6,module.extensions.integration.managed_service_assertions_for_state_admission_m_55b90ea265,module.extensions.integration.network_flow_state_v4,module.extensions.integration.state_concurrent_admission` | 98798/0; 16,065 ms | Exact four previously affected rows pass |
| `r03-stability-a-schema-01`; `make source-check CHECK=schema` | 97221/0; 6,891 ms | Schema projections pass |
| `r03-stability-a-generated-01`; `make source-check CHECK=generated` | 19446/0; 16,986 ms | Generated projections pass |

All three passing test invocations passed independent
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, exit zero,
`full_validity=1`, complete execution/resource coverage and invocation accounting.
Their cleanup records are respectively **11/12/14 completed**, with zero failed
or blocked records. Retained secret scans passed over **21/217/34 files**.
The pilot is excluded from the three qualification rosters and all performance
acceptance. No original published-endpoint failure was reproduced.

**Preserved unsuccessful attempts.** `r03-stability-a-unit-01` stopped before test
execution at `accounting_executable_identity` (19895/2); canonical catalog
preparation repaired the stale executable binding. Unit-02 then passed (69897/0,
2,491 ms), before the final qualification helper changes. Format-02 and generate-02
failed because the new authored catalog row was not ASCII-sorted; the authored
input was corrected, generate-03 passed (25870/0, 11,280 ms), and subsequent
format/catalog checks passed. These are implementation-entry failures, not timing
samples or evidence of a product regression.

Pilot `r03-stability-a-live-01` failed (9480/2, normalized 10, 102,615 ms) because
its success assertion required failure-only reinspection fields, its absence
check used an inadequate SDK classification, and bounded logs could not retain
the complete roster. Pilot `r03-stability-a-live-02` failed (81259/2, normalized
10, 96,229 ms): all 64 readiness records were retained, but the SDK exposes its
404 category through `errors.Is`, not the marker interface used by the test.
Fresh exact-ID inspection confirmed all 64 allocations absent; the original
failed/unproven receipts remain unchanged in `stability-a-live-02-analysis.json`.
The corrected test uses the SDK's typed classification, a live inspection client,
initial success observations and durable native journals. Both failed pilots
passed diagnostic/event/timing structural audits with `full_validity=0` and
completed outer cleanup; neither is reinterpreted as a passing qualification.

**Disposition.** Diagnostic assurance ACCEPTED; runtime characterization COMPLETE;
transport replacement NOT ADOPTED; performance NOT ESTABLISHED. The historical
host-forwarding cause remains uncertain. WF-11j-b must record its conditional
no-change or candidate decision after this receipt and boundary Markdown lint.
WF-11j-c must still complete all three independent qualification rosters and both
source treatments' affected/containing assurance. S-007e remains open.

### E-150 — WF-11j-b conditional publication decision: NO_CHANGE

WF-11j-a boundary lint passed **26436/0**, 16,880 ms, run
`r03-stability-a-boundary-lint-01`, before this workstream began. Source remains
`sha256:1bd52588e175c06311e9c4bf34036697c211ba90b9003aed6f9d9dac8d023c8b`.

**Decision and rationale.** Retain the current explicit ephemeral publication
path. The E-149 controlled pilot establishes the final wire-request shape and
successful service/client/cleanup boundaries, but does not reproduce the original
publication failure. The Section 7.7 Phase B acceptance condition—a failing
control distinguished from a passing replacement—therefore is not met. No
publish-all pilot, runtime toggle, host detection, fallback endpoint, retry or
longer deadline is introduced. There is no candidate transport implementation or
owner amendment to roll back. Accepted v6 diagnostics remain in both treatments.

R03-032 → NLSpec Section 11.4 and reference profile Sections 13.7/14 → existing
PostgreSQL owner with accepted observation → E-149 / V-020 is the retained route.
Assurance of observation and existing boundaries is accepted; publication-change
assurance is NOT APPLICABLE; performance is NOT ESTABLISHED. The historical
forwarding cause remains unresolved and is not asserted to be repaired. Stability
qualification must pass on the unchanged path before performance resumes; a
recurrence returns to characterization and blocks performance, as Section 7.7
requires. No shared host service has been changed or restarted.

This documentation-only decision requires boundary `make lint-markdown`; it
requires no additional product run. After that pass, WF-11j-c enters with three
frozen 64-lifetime rosters, exact affected rows, containing service assurance,
both source treatments and complete settlement/evidence checks still required.

### E-151 — WF-11j-c qualification entered

WF-11j-b boundary lint passed **94191/0**, 9,440 ms, run
`r03-stability-b-boundary-lint-01`, before qualification began. The immutable entry
roster is `stability-c-qualification-freeze.json` in the continuation directory.
It binds source `sha256:1bd52588e175c06311e9c4bf34036697c211ba90b9003aed6f9d9dac8d023c8b`,
the E-149 environment/request/resource profile, and independent run IDs
`r03-stability-c-qualify-c-01`, `-02`, `-03`, each with the exact 64-lifetime
population. No failed lifetime may be replaced. All qualification is excluded
from performance samples. A failure blocks the gate; the E-150 retained
publication path is not presumed stable from its pilot alone.

### E-152 — WF-11j-c stability qualification completed

**Qualification.** The E-151 freeze completed without replacement attempts or
roster changes. All **192** fresh lifetimes used distinct allocations, passed
normal authenticated/security/capability readiness, retained complete required
request and endpoint observations, and proved exact removal. There were zero
unexpected acquisition failures, stale identities, missing outcomes or unresolved
allocations. Each invocation exercised 16 serial lifetimes, 32 concurrent
lifetimes within four admitted lanes, and 16 lifetimes in competing pairs.
This establishes observed reliability for this declared finite workload; it does
not identify the historical forwarding cause or guarantee absence of future host
failures. All qualification and service assurance remain excluded from timing
acceptance samples.

R03-032 → NLSpec Section 11.4 and reference profile Sections 13.7/14 → unchanged
publication with accepted v6 observation and the explicit qualification row →
V-020 / E-151 / this receipt is the traceability route. No environment mutation,
publication replacement, scheduler or capacity change occurred.

| Independent qualification run | Native receipt | Result |
| --- | --- | --- |
| `r03-stability-c-qualify-c-01` | 68163/0; 99,740 ms | 64/64 ready and exactly settled |
| `r03-stability-c-qualify-c-02` | 5377/0; 99,643 ms | 64/64 ready and exactly settled |
| `r03-stability-c-qualify-c-03` | 22892/0; 97,196 ms | 64/64 ready and exactly settled |

Each used E-149's exact qualification row and public `make test-slice` command,
`CGO_ENABLED=0`, execution cache off, basic diagnostics and required accounting.
Each passed `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`
with exit zero and `full_validity=1`, complete execution/resource coverage and
accounting. Each has **12 completed outer cleanup records**, zero failed/blocked
cleanup, a passing **217-file** retained secret scan, and 64 complete native
suite-service scope/journal pairs. `stability-c-qualification-results.json`
retains the roster results and verifies 192 distinct allocation identities.

**Treatment reconciliation and service assurance.** The 17 authored changes from
E-149 were preserved as `stability-common-authored.patch` and applied to the
immutable B8 source only after exact restoration and successful patch admission.
Generated projections were regenerated independently through Make. All common
transport/qualification implementation paths and diagnostic schemas are byte
identical in both treatments; composite owner/attachment/catalog files preserve
the intended reset-lifecycle differences. `stability-c-treatment-fidelity.json`
records the common-path proof. No changed-source run is admitted as equivalent.

| Treatment | Executable source | Preserved archive SHA-256 |
| --- | --- | --- |
| Baseline, standalone resets | `sha256:5650c8f8ac0f2306779f9d1c732ddaa3ec925731cac8c9947d865f4ad5e8e312` | `d88e97ba442409136ba47d2be83f2027acf95f605eaf5cf133a6574d7f6d58ac` |
| Candidate, admitted preparation | `sha256:1bd52588e175c06311e9c4bf34036697c211ba90b9003aed6f9d9dac8d023c8b` | `4bf61a5fc38925d531b20385170c4689e1e45c1f2bc511732ff63fbc918f2aa5` |

Archives are `diagnostic-baseline-source-stability` (7,502 files) and
`lifecycle-candidate-source-stability` (7,510 files) in the continuation directory.
They preserve source for finalization and rollback; WF-11j-d still owns the final
performance freeze. The tracker remains preserved across source switches.

| Service assurance run / public selection | Native receipt | Cleanup and scan |
| --- | --- | --- |
| `r03-stability-c-affected-c-01`; E-149's exact four module.extensions rows | 66416/0; 14,872 ms | 14 completed; 34-file scan pass |
| `r03-stability-c-containing-c-01`; `make test-slice OWNER=module.extensions SCOPE=services` | 39172/0; 107,035 ms | 41 completed; 162-file scan pass |
| `r03-stability-c-unit-b-01`; E-149's exact two diagnostic/helper rows | 25945/0; 2,451 ms | 11 completed; 21-file scan pass |
| `r03-stability-c-affected-b-01`; E-149's exact four module.extensions rows | 8375/0; 22,138 ms | 14 completed; 34-file scan pass |
| `r03-stability-c-containing-b-01`; `make test-slice OWNER=module.extensions SCOPE=services` | 44304/0; 90,744 ms | 41 completed; 162-file scan pass |

All five passed independent public `CHECK=all`, exit zero, `full_validity=1`,
complete diagnostic coverage, accounting and physical settlement. Baseline
regeneration passed `r03-stability-c-generate-b-01` (36337/0, 10,760 ms), followed
by canonical catalog preparation `r03-stability-c-prepare-b-01` (94885/0,
1,665 ms). No baseline CI timing invocation was run.

**Decision and next entry.** Stability assurance ACCEPTED for the declared
conditions; transport disposition remains NO_CHANGE; lifecycle performance remains
PENDING. Failed pilots and the historical interrupted CI remain preserved with
their original outcomes. Boundary Markdown lint must pass before WF-11j-d starts.
That workstream must run finalization, reconcile projections and owner checks,
freeze the exact experiment and treatment map, and pass its tracker boundary.
Section 8.5 then governs matched browser timing and the assumed 3,600,000 ms CI
elapsed reference. Final S-007e validation and handoff remain mandatory.

### E-153 — WF-11j-d finalization and performance freeze entered

WF-11j-c boundary lint passed **27457/0**, 9,860 ms, run
`r03-stability-c-boundary-lint-01`, before this workstream began. Both treatment
snapshots and their source identities are preserved under E-152. The current
checkout is the qualified standalone-reset baseline with the identical common
diagnostic correction. Finalization and focused projection/owner verification
precede the experiment freeze. No performance samples have run.

`RESULTS_DIR` is unset: no compatible successful full warm-check root exists.
Retained-run maintenance must therefore be reported skipped for that reason.
CI uses the fixed assumed 3,600,000 ms elapsed reference, descriptive resource
metrics and mandatory correctness/cleanup under Section 8.5. No CI baseline
warm-up or timing run is admitted. Browser remains a measured matched comparison.

### E-154 — WF-11j-d completed; experiment 09 frozen

**Assurance — PASS. Performance — PENDING. Transport — NO_CHANGE.** Both
qualified treatments passed finalization before broader owner and projection
checks. The effective-request diagnostic v6 and explicit lifetime qualification
are common to both. The original published-endpoint failure was not reproduced;
no causal repair or zero-future-failure claim is made. E-152's 192 qualified
lifetimes and both-source affected/containing checks satisfy Section 7.7's bounded
stability gate. S-007e still owns final integration and completion.

All commands below ran from the repository root with `CGO_ENABLED=0`. Owner
checks additionally used `CARTULARY_HARNESS_CACHE_MODE=off`,
`HARNESS_DIAGNOSTICS=basic`, and `HARNESS_ACCOUNTING=required`.
Run roots are `.cartulary/test-results/r03-stability-d-<label>-<treatment>-01`,
where treatment is `b` or `c`. Native session/exit and graph milliseconds follow.

| Label / exact public command | Baseline session/exit; ms | Candidate session/exit; ms |
| --- | --- | --- |
| `finalize`: `make agent-finalize` | 54073/0; 27982 | 13000/0; 27677 |
| `catalog`: `make source-check CHECK=catalog` | 15800/0; 1680 | 10660/0; 1688 |
| `catalog-owner`: `make test-slice OWNER=harness.test_catalog` | 7262/0; 56162 | 26821/0; 56457 |
| `command-owner`: `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity` | 11889/0; 70304 | 86931/0; 71877 |
| `evidence-owner`: `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence` | 49847/0; 13004 | 4805/0; 13319 |
| `schema`: `make source-check CHECK=schema` | 44043/0; 6939 | 36302/0; 7014 |
| `generated`: `make source-check CHECK=generated` | 56914/0; 17254 | 73845/0; 17631 |

Each of the six owner invocations passed the independent public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` audit,
exit 0, `full_validity=1`, complete required execution/resource diagnostics and
available invocation accounting. Cleanup and final retained-secret scans passed.
The native completion receipts for candidate commands whose initial output was
compacted were recovered from this task's command-execution records; no test was
rerun to replace a missing transcript. No candidate-owned resource is unresolved.
`RESULTS_DIR` was unset for finalization: retained-run maintenance was skipped
because no compatible successful full warm-check root exists.

| Treatment | Executable source digest | Immutable archive SHA-256 |
| --- | --- | --- |
| Baseline | `5650c8f8ac0f2306779f9d1c732ddaa3ec925731cac8c9947d865f4ad5e8e312` | `3a9d0d58fea6f77ec5942c78f1e07bbf4450b038aba4cd1df1a598dab3b6f973` |
| Candidate | `1bd52588e175c06311e9c4bf34036697c211ba90b9003aed6f9d9dac8d023c8b` | `50fa83eb6c16be0d00d38e1721f014d87d960a48d1cb765badd6c2598dd998db` |

The preservation artifacts are `diagnostic-baseline-source-exp09.*` and
`lifecycle-candidate-source-exp09.*` under
`.cartulary/remediation03/continuation`. `exp09-treatment-difference.json` records
360 lifecycle treatment paths; the only path-set change from experiment 08 is
the common generated readiness validator's v5-to-v6 migration.
`stability-c-treatment-fidelity.json` establishes the identical common corrections.
F-019 → R03-032 / Section 7.7 → bounded publication observer and qualification
row → V-020 / E-149–E-154 is the stability trace. R03-019–026 and V-010–013 retain
their diagnostics/lifecycle/experiment ownership; this work does not alter
scheduler policy, capacity, retries, isolation or pooling.

**Frozen experiment — normative for this attempt.**
`performance-experiment-exp09-freeze.json` freezes source archives, the exact
sorted row populations, public static plans, environment/storage identities,
permitted treatment differences, run roster and analysis-file hashes.
`performance-experiment.json` is its mutable receipt ledger; the freeze is
immutable. Public plans were exported with
`make target-plan TARGET=<browser-e2e|ci> INPUTS=<exp09-workload-inputs.json> JSON=1`.
Both treatments select the same 383 browser rows and 2,279 CI rows. Browser
scheduled units change 137→76; CI units change 600→517. Removed units are exactly
standalone browser resets. These are structural observations, not speedup evidence.
The baseline CI plan is counterfactual selection evidence, not a measured run.

The exact run prefix is `r03-s011-exp09-`. Execution order is:

1. `browser-warm-b`, `browser-warm-c`, `ci-warm-c` (excluded warm-ups).
2. `browser-p01-b`, `browser-p01-c`, `browser-p02-c`, `browser-p02-b`.
3. `ci-p01-c`, `ci-p02-c`.
4. Only if required by Section 8.5, `browser-p03-b`, `browser-p03-c`,
   `browser-p04-c`, `browser-p04-b`, then `ci-p03-c`, `ci-p04-c`.

All timing commands use native PTYs, warm dependency/build caches, cgo disabled,
execution-result reuse off, basic diagnostics and required accounting. There is
no CI baseline invocation or baseline warm-up. All three warm-ups must pass
independent audits and bind runtime graph identities before acceptance sampling.
Source switches must preserve exact snapshots and prepare the canonical native
controller through Make before timing starts. No intentional competing workload
or status polling is admitted during quiet measurement.

Browser uses measured matched pairs. CI uses the fixed assumed 3,600,000 ms
reference; resource/queue/cleanup metrics remain descriptive, with no comparative
guardrail or baseline-uncertainty claim. Section 8.5's first-two strict-below-5%
rejection and inclusive-20% early acceptance apply separately. Otherwise browser
completes four matched pairs and CI four candidate observations. Full CI analysis
uses candidate-only whole-run resampling conditional on the fixed assumption;
it is not paired. Both use the frozen 10,000 resamples, seed 20261021 and decimal
arithmetic. Correctness and successful cleanup remain unconditional.
Both workloads receive their first two observations unless correctness/cleanup
prevents them; rejection stops optional sampling and restores lifecycle owner,
implementation and projections together, retaining accepted common diagnostics.

Next entry: successful `make lint-markdown` for this boundary, then the first
excluded browser baseline warm-up. No performance acceptance sample has run.

### E-155 — Experiment 09 warm-ups entered

WF-11j-d boundary Markdown lint passed **63077/0**, 9,690 ms, run
`r03-stability-d-boundary-lint-01`, before this workstream resumed. E-154 freezes
the approved source/selection/environment/protocol. Three excluded warm-ups
precede acceptance sampling. The historical endpoint cause remains uncertain;
a repeated transport failure blocks sampling and requires a concrete disposition,
not another unbounded CI retry. CI has no measured baseline invocation.

### E-156 — Experiment 09 closed; WF-11k staticcheck repair entered

The three declared warm-ups ran in frozen order. Browser baseline native
**61916/0**, graph 1,159,644 ms, owning-command 1,172,372.902877 ms, and candidate
**21480/0**, graph 1,151,666 ms, owning-command 1,163,977.375995 ms, both passed.
Exact roots are `.cartulary/test-results/r03-s011-exp09-browser-warm-b` and
`r03-s011-exp09-browser-warm-c`. Each passed `harness-observability-check CHECK=all`,
exit 0 / `full_validity=1`, complete canonical/allocation diagnostics, empty
accounting scope and final secret scan. Both used four stack generations and zero
reset-only generations. Candidate's 61 preparation receipts exactly matched the
subsequent runner lease, allocation, session and backend generation. These runs
were declared excluded and establish no performance acceptance.

`r03-s011-exp09-ci-warm-c`, native **33067/2**, normalized exit 10, failed after
3,549,550 ms graph execution; owning-command duration was 3,621,141.577502 ms.
516 of 517 units passed. The sole failed unit was
`operation:harness.lint_go_staticcheck`: SA1019 at
`internal/testutil/pgtest/publication_diagnostic_test.go:35`, where
`client.WithAPIVersionNegotiation()` is deprecated because negotiation is already
the SDK default and the option is a no-op. The failure was introduced by the new
SDK error-classification regression test. No transport acquisition failure was
reported. All 966 cleanup steps completed, invocation accounting settled empty,
and the retained-secret scan passed across 9,147 files. This failed CI remains a
failure; passing units do not establish the mandatory full-CI gate.

Public diagnostics/events/timing audits each returned 0 with `full_validity=0`.
Canonical timing and all 54 terminal allocation records were complete and settled.
Optional verbose execution diagnostics and sampled resources exhausted their
existing limits; 2,473 resource samples were omitted. Required owning-command
elapsed, charged CPU, kernel peak memory, canonical queue/cleanup metrics and
allocation lifecycle observations remained available. Optional sample truncation
is disclosed and does not authorize increasing limits or claiming sampled
whole-host coverage. Complete paged critical-path exports and accounting retain
their independent scopes. An initial `explain-run TARGET=ci JSON=1` without a
supported detail returned usage exit 2; the corrected human reader and supported
performance/critical-path readers succeeded. This reader usage error is separate
from the test outcome.

`performance-experiment-exp09-closed.json` preserves the failed warm-up and marks
all twelve unexecuted acceptance attempts omitted. The immutable experiment 09
freeze and both source archives remain unchanged. Zero acceptance samples ran.
A repaired source requires a fresh experiment; no failed attempt will be relabeled
or replaced. The CI reference remains assumed 3,600,000 ms under Section 8.5.

**Corrective WF-11k — authorized implementation scope.** Remove the deprecated
no-op argument in both source treatments, retaining the SDK's default behavior
and all 404/409/500 classification assertions. No behavioral owner amendment,
deadline change, retry, suppression or publication-policy change is required.
R03-032 / V-020 → SDK error-classification regression → focused pgtest helper
assurance and `make lint CHECK=go` supply traceability. Finalization must precede
broader end-of-work verification; regenerate through Make, verify affected
projections, preserve matching B/C treatment differences, freeze a new experiment,
and pass this tracker's boundary lint before resuming warm-ups. The 192-lifetime
qualification is unchanged by a test-client no-op removal; it is retained with
its original source identity rather than relabeled as a fresh qualification.
S-007e final integration and handoff remain open.

### E-157 — WF-11k completed; experiment 10 frozen

**Assurance — PASS. Performance — PENDING.** The common diagnostic regression
test now constructs the SDK client with `client.WithHost(server.URL)` alone.
The pinned SDK source confirms the removed negotiation option returns nil without
changing configuration; default negotiation and all 404/409/500 assertions remain.
Imports are grouped normally. There is no fixture runtime, publication, deadline,
retry, admission, contract or requirement change and no lint suppression.
`wf11k-common-repair.json` preserves the identical before/after file hashes.

All commands ran from the repository root with `CGO_ENABLED=0`. Each treatment
ran finalization before broader checks. Unit checks used cache off, basic
diagnostics and required accounting; Go lint used cache off.
Exact roots are `.cartulary/test-results/r03-s011-k-<label>-<b|c>-01`.

| Label / exact public command | Baseline native session/exit; graph ms | Candidate native session/exit; graph ms |
| --- | --- | --- |
| `finalize`: `make agent-finalize` | 59415/0; 27321 | 92574/0; 27622 |
| `catalog`: `make source-check CHECK=catalog` | 12487/0; 1716 | 41374/0; 1709 |
| `unit`: `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract` | 85756/0; 21091 | 24466/0; 16981 |
| `lint`: `make lint CHECK=go` | 79339/0; 12425 | 46347/0; 8295 |
| `generated`: `make source-check CHECK=generated` | 97501/0; 20437 | 24260/0; 16401 |

Both unit invocations passed independent `harness-observability-check CHECK=all`,
exit 0 / `full_validity=1`, with complete required diagnostics/accounting.
All ten invocations completed cleanup and passed retained-secret scans. No owned
resource remains unresolved. `RESULTS_DIR` was unset for finalization; retained-run
maintenance was skipped because no compatible successful full warm-check root
exists. R03-032 / V-020 → SDK classification regression → helper and Go-lint checks
is the corrective trace. The 192-lifetime qualification remains source-bound
historical assurance for the unchanged fixture runtime; it is not relabeled.

| Treatment | Executable source digest | Immutable archive SHA-256 |
| --- | --- | --- |
| Baseline | `b1b95fd3e781b9c89227db99ee7e620ab45022d976265388dda8e2543a51bc12` | `27bedc8e4644a1fc8a089e80d4f7905204a6f39c35d158881dc68193b92b3937` |
| Candidate | `e3a14f28ed62e110eaee3b8ca22815fa414ab3982ae96a620f27f5fb2edf8a00` | `0bf5c5d34eadb581a8f602271f6a08aa8465b5805a13f0dce11e544b0681b368` |

Preservation uses `diagnostic-baseline-source-exp10.*` and
`lifecycle-candidate-source-exp10.*` under `.cartulary/remediation03/continuation`.
Relative to each experiment 09 source, the only executable-source path changed
is `internal/testutil/pgtest/publication_diagnostic_test.go`.
`exp10-treatment-difference.json` retains the same 360 lifecycle treatment paths.
Fresh public target plans preserve the exact 383 browser / 2,279 CI row sets and
137→76 / 600→517 scheduled-unit populations. Docker, WSL, toolchain, storage,
resource profile and acceptance policy remain unchanged.

`performance-experiment-exp10-freeze.json` is the immutable fresh freeze;
`performance-experiment.json` is the receipt ledger. E-154's exact roster/order
applies with prefix **`r03-s011-exp10-`**. One fresh excluded browser warm-up per
treatment and one CI candidate warm-up precede sampling. There is no CI baseline
run or baseline warm-up. CI retains the fixed assumed 3,600,000 ms reference,
descriptive resource metrics and mandatory correctness/cleanup. Browser retains
measured pairs. Both first-two rules and the maximum four observations/pairs,
bootstrap arithmetic and rollback requirement remain unchanged under Section 8.5.
Failed experiment 09 and its omitted acceptance attempts remain preserved.

**Descriptive diagnostic observation — not acceptance.** The failed CI's complete
critical path contains 724,940 ms graph queueing and 1,538,229 ms host admission.
These are distinct timing categories. Fifty-four allocations settled: 41 retired
for capacity pressure and 13 after terminal borrower release. All 61 webserver
reset/runner bindings matched. Canonical phase sums may overlap and are not
exclusive wall-clock causes. The complete wait export required four pages per
view (365 records, no omissions). Optional activity/sample truncation did not
remove the required canonical, allocation or invocation-accounting observations.
No scheduling or capacity change is admitted by these findings.

Next entry: successful `make lint-markdown` for this workstream boundary, then
experiment 10's excluded baseline browser warm-up. S-011 acceptance and S-007e
final validation/handoff remain open; zero acceptance samples have run.

### E-158 — Experiment 10 warm-ups entered

WF-11k boundary Markdown lint passed **71703/0**, 10,450 ms, root
`.cartulary/test-results/r03-s011-k-boundary-lint-01`, before performance work
resumed. E-157 supplies the fresh immutable sources and protocol. The exact
baseline snapshot was revalidated before its first warm-up. The three excluded
warm-ups precede acceptance sampling; no CI baseline invocation is admitted.

### E-159 — Experiment 10 warm-up completion and sampling entry

**Assurance: PASS. Performance: not yet decided.** E-157's immutable sources,
E-158's boundary lint and Section 8.5's protocol remain unchanged. The exact
commands below used `CGO_ENABLED=0`, disabled execution-result reuse, basic
diagnostics, required accounting and native PTYs. All roots are under
`.cartulary/test-results/`; all three runs are excluded warm-ups.

| Run ID | Command | Native session / exit | Owning elapsed ms | Passed / total units |
| --- | --- | --- | --- | --- |
| `r03-s011-exp10-browser-warm-b` | `make browser-e2e SUITE=functional` | 33804 / 0 | 1,171,759.995699 | 137 / 137 |
| `r03-s011-exp10-browser-warm-c` | `make browser-e2e SUITE=functional` | 28152 / 0 | 1,267,860.358337 | 76 / 76 |
| `r03-s011-exp10-ci-warm-c` | `make ci` | 40386 / 0 | 3,791,240.400591 | 517 / 517 |

Each exact root passed `make harness-observability-check CHECK=all` with
`full_validity=1`. Public `make explain-run` performance and complete root /
webserver critical-path unit, wait and allocation exports passed. The CI wait
views each contain 313 records across offsets 0, 100, 200 and 300; both final
pages contain 13 records and `next_offset=null`. Thirteen CI exports exited 0;
the final page readers were native 63446/0 and 47757/0. CI audit was 93162/0.
The receipt ledger retains all browser reader exits and exact export paths.

All frozen row populations match the audited runtime plans: 383 browser rows
and 2,279 CI rows. Browser graphs match the static freeze exactly. CI has only
the predeclared E-090 two Go CPU claims changing 4 to 9, with corresponding
semantic digests; no other shared public unit field differs.
`exp10-runtime-plan-binding.json` records the exact comparison. The validated
runtime graph digests are frozen in `performance-experiment.json` before samples.

Browser baseline / candidate cleanup completed 276 / 154 obligations; CI
completed 956. All invocation accounting scopes settled empty and retained
secret scans passed (2,296 / 2,235 / 9,005 files). Candidate browser and CI each
have 61 matching reset/runner lease, allocation, session and backend-generation
bindings. No candidate-owned resource remains unresolved. CI optional execution
activity is truncated and 793 resource samples were omitted; canonical timing,
allocation lifecycle and invocation accounting are complete. Optional sampling
coverage is not promoted to whole-host utilization coverage.

**Descriptive observations, not acceptance samples.** The CI critical path has
43 units, 646,598 ms graph queueing and 3,070,582 ms unit lifetime. Phase-duration
sums include 1,834,851 ms host admission, 114,929 ms acquisition, 125,556 ms
preparation, 966,012 ms runner and 26,228 ms release. Phases may overlap and these
are not exclusive wall-clock causes. All 44 browser-stack generations settled;
31 retired for capacity pressure and 13 after terminal borrower release. None
was reset-only. Charged CPU was 11,545,199,419 us; kernel peak memory was
9,388,920,832 bytes; ready-to-start p95 was 3,417,463 ms; cleanup union was
159,905 ms. CPU, memory, queue and cleanup remain descriptive for CI under the
assumed 3,600,000 ms reference. No resource-policy change is authorized by this
observation, and instantaneous container CPU does not measure reserved claims.

R03-019–024 / V-010–012 → canonical diagnostics, preparation continuity and
independent workload validation remain the trace. Migration remains coordinated;
there is no new executable source change or rollback. Next: tracker Markdown
lint, then `r03-s011-exp10-browser-p01-b` under the frozen order. Both first-two
workload decisions and S-007e remain mandatory and open.

### E-160 — Browser performance rejection; CI observations still required

**Assurance: PASS. Browser performance: REJECTED.** The first two matched
experiment 10 pairs used E-157's frozen sources and Section 8.5 settings, with
baseline first for pair one and candidate first for pair two. Each command was
`make browser-e2e SUITE=functional`, with `CGO_ENABLED=0`, cache off, basic
diagnostics, required accounting and a native PTY. Exact roots use the run IDs
below under `.cartulary/test-results/`.

| Run ID | Native session / exit | Owning elapsed ms | Cleanup completed | Retained scan files |
| --- | --- | --- | --- | --- |
| `r03-s011-exp10-browser-p01-b` | 83531 / 0 | 1,163,402.224141 | 276 | 2,296 |
| `r03-s011-exp10-browser-p01-c` | 56458 / 0 | 1,145,086.220052 | 154 | 2,235 |
| `r03-s011-exp10-browser-p02-c` | 96318 / 0 | 1,163,319.307308 | 154 | 2,235 |
| `r03-s011-exp10-browser-p02-b` | 62310 / 0 | 1,178,957.724865 | 276 | 2,296 |

All four exact roots passed `make harness-observability-check CHECK=all`, exit 0,
`full_validity=1`. Each had seven successful public `make explain-run` exports:
performance, plus complete root and webserver unit/wait/allocation views. All
required and optional diagnostic coverage was complete. Baselines passed all
137 units; candidates passed all 76. All cleanup obligations completed,
accounting scopes settled empty and retained scans passed. Each candidate had
61 matching preparation/runner identities. No resource remains unresolved.

Decimal arithmetic over independently validated owning-command observations
shows improvements of **1.574348381749368976772163170%** and
**1.326461265503877393774253820%**. Both are strictly below 5%, so Section 8.5
requires rejection. No bootstrap confidence bounds or full guardrail pass are
claimed. `performance-decision.json` records the decision; the original source
snapshots and every attempt remain preserved.

All four runs used four stack generations, each retired after terminal borrower
release; none was reset-only. Scheduled-unit populations changed 137→76 and
critical-path populations 80→43. Whole-selection ready-to-start p95 was 1 ms in
all four runs; that equality does not erase the changed percentile population.
`exp10-browser-corresponding-cohort-waits.json` separately compares exact matching
cohorts on both reconstructed webserver paths, with scope and unmatched sets
explicit. Canonical phase sums and resource claims do not establish exclusive
wall-clock causes or utilization.

The 360-path lifecycle treatment remains temporarily present only for the two
mandatory CI observations. Both workloads SHALL receive their first two
observations unless correctness or cleanup prevents execution. Optional browser
pairs three/four and CI observations three/four are now omitted because the
browser gate rejected the optimization. After the two CI observations, the
lifecycle owner, implementation and projections SHALL be restored together;
accepted diagnostics and common stability repairs SHALL remain. No CI baseline
run is admitted; CI uses the user-assumed 3,600,000 ms reference and descriptive
resource observations. R03-023 / V-012 owns this decision and restoration.

Next: successful tracker Markdown lint, candidate-source restoration and public
Make preparation, then `r03-s011-exp10-ci-p01-c` and `r03-s011-exp10-ci-p02-c`.
S-011 is not complete until those dispositions and restoration are recorded;
S-007e final verification remains required.

### E-161 — WF-11 / S-011 closed: both workloads rejected; lifecycle restored

**Assurance: PASS for both measured treatments. Performance: REJECTED for browser
and CI. Restoration: COMPLETE.** E-160 records the two measured browser pairs.
The two mandatory CI candidate observations use the fixed user-assumed
3,600,000 ms elapsed reference under E-145 / Section 8.5. No CI baseline timing
or baseline warm-up was run. Both commands were `make ci` with `CGO_ENABLED=0`,
cache off, basic diagnostics, required accounting and native PTYs.

| Run ID under `.cartulary/test-results/` | Native session / exit | Owning elapsed ms | Conditional improvement | Cleanup completed | Scan files |
| --- | --- | --- | --- | --- | --- |
| `r03-s011-exp10-ci-p01-c` | 49372 / 0 | 3,544,358.279026 | 1.545603360388888888888888890% | 960 | 9,018 |
| `r03-s011-exp10-ci-p02-c` | 33855 / 0 | 3,695,341.092058 | -2.648363668277777777777777800% | 960 | 9,066 |

Both exact roots passed all 517 units and `make harness-observability-check
CHECK=all`, `full_validity=1` (native 71945/0 and 47410/0). Each run has thirteen
successful public `make explain-run` exports: performance plus complete root /
webserver critical-path unit, allocation and four-page wait views. Wait totals
are 379 and 308 records per view; final pages contain 79 and 8 records, with
`next_offset=null`. All required canonical, allocation and invocation-accounting
coverage is complete. Each run has 61 matching reset/runner lease, allocation,
session and generation bindings. Cleanup completed, accounting scopes settled
empty and retained-secret scans passed; no owned resource remains unresolved.
Optional execution diagnostics are truncated; optional resource samples omit
1,683 / 1,046 observations and PostgreSQL checkpoint observations are unavailable.
These limitations remain explicit and do not establish a whole-host utilization
or checkpoint-cause claim.

Both CI improvements are strictly below 5%, so CI independently rejects the
candidate. Browser independently rejects at 1.574348381749368976772163170% and
1.326461265503877393774253820%. No bootstrap confidence bounds are claimed;
CI resource metrics are descriptive without a measured baseline. All six optional
attempts (browser pairs three/four, CI observations three/four) are recorded as
omitted after the browser rejection. The frozen analysis helper consumed only
independently validated exports, with decimal arithmetic and no equivalent-run
admission across changed sources. `performance-decision.json` and
`performance-experiment-exp10-closed.json` preserve the exact decisions and scope.

| Descriptive CI metric | Observation 1 | Observation 2 |
| --- | --- | --- |
| Charged CPU, us | 11,092,128,203 | 11,324,320,708 |
| Kernel peak memory, bytes | 10,976,186,368 | 8,805,273,600 |
| Ready-to-start p95, ms | 3,198,826 | 3,244,574 |
| Cleanup union, ms | 158,762 | 152,879 |
| Critical-path graph queue, ms | 706,615 | 650,270 |
| Critical-path host admission phase sum, ms | 1,536,081 | 1,710,449 |
| Acquisition phase sum, ms | 129,087 | 156,930 |
| Preparation phase sum, ms | 125,165 | 124,748 |
| Runner phase sum, ms | 946,096 | 949,537 |
| Release phase sum, ms | 26,308 | 26,187 |

Each CI path contains 43 units. Each run has 48 settled browser-stack generations:
35 retire for capacity pressure and 13 after terminal borrower release; none is
reset-only. Phase sums may overlap and SHALL NOT be added as exclusive wall-clock
causes. Reservations SHALL NOT be described as measured utilization. CI has no
measured corresponding-cohort baseline; that comparison is unavailable. Browser's
supplementary comparison matches 38 cohorts on both webserver paths: graph queue
totals change 8→7 ms / 10→7 ms and host admission totals 1,267→1,712 ms /
1,409→1,341 ms. This is explicitly path-scoped. Whole-selection queue percentiles
have different populations after reset-unit removal; their shared 1 ms browser
p95 does not establish identical work populations.

**Coordinated restoration.** After both CI invocations and all readers exited,
the exact immutable baseline 10 snapshot restored all 360 treatment paths while
preserving this tracker. The NLSpec owner, graph/runtime/event/result/browser
contracts, compiler, executor, finalizers and tests were restored together.
No dual reader or compatibility switch was introduced. `make generate` passed
native 76144/0, 10,980 ms, root `r03-s011-exp10-restore-generate-01`;
`make source-check CHECK=catalog` passed 64071/0, 1,578 ms, root
`r03-s011-exp10-restore-catalog-01`. Post-generation preservation checks matched
all 7,501 non-tracker snapshot paths. Retained executable source is
`sha256:b1b95fd3e781b9c89227db99ee7e620ab45022d976265388dda8e2543a51bc12`.
`exp10-restoration.json` records the exact boundary. S-009 diagnostics and common
corrective/stability work remain; candidate-only preparation work is not shipped.

R03-023 / V-012 → frozen experiment → both first-two decisions → exact source
restoration is complete. The observed admission delay and capacity-pressure
retirement justify a future investigation of reservation ownership and lifetime;
they do not authorize changing capacities, fairness, ranking, workers or pooling
in this effort. Section 7.7's historical transport cause remains uncertain; its
192-lifetime qualification and current CI passes support the retained path without
claiming a reproduced transport repair.

Next entry: successful `make lint-markdown` for this workstream, then WF-07e /
S-007e: `make agent-finalize`, required projections/owner checks, explicit support,
visual and UI-review checks, one final `make ci` on the retained source, audits,
settlement, final scans and a closing tracker/lint handoff. That final CI is
validation, not a prohibited CI performance baseline or an acceptance sample.
`RESULTS_DIR` remains unset unless a compatible successful full warm-check root
is established; the passing CI warm-ups do not supply that prerequisite.

### E-162 — Final validation entry and corrective WF-07f registration

WF-11 boundary lint passed native **76620/0**, 10,370 ms, root
`r03-s011-exp10-boundary-lint-01`, before WF-07e began. Retained source remains
E-161's exact baseline 10. `make agent-finalize` passed **19672/0**, 26,209 ms,
root `r03-s007e-finalize-01`; `RESULTS_DIR` was unset, so retained-run maintenance
was skipped. Subsequent generated/schema/catalog checks passed **31001/0**,
**89152/0**, **52501/0**, respectively, roots `r03-s007e-generated-01`,
`r03-s007e-schema-01`, `r03-s007e-catalog-01`. These are source-bound intermediate
passes, not final completion. Public task guides selected the four principal
owners and explicit visual/UI-review routes; public help also confirms support.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007e-browser-owner-01
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.browser
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` failed native **21182/2**,
normalized product/test_assertion_failure **10**, in 233,630 ms. Forty-nine of
50 units passed. The sole failing unit is `go:f78ef08a0c0e9bfc:e8bb5347f545`,
row `harness.browser.unit.testservices_lifecycle_contract`. Its sibling retired
command assertion passed. Browser execution, integration and frontend targets
passed. All 83 cleanup obligations completed, the accounting scope settled empty
and the retained-secret scan passed across 899 files. The exact run passed the
public diagnostics audit, exit 0, **full_validity=0**; this does not establish
passing assurance.

The native failing assertion is
`TestRunReturnsBeforeSlowServiceTerminationAfterChildSuccess` at
`tools/testservices/main_test.go:323`: elapsed **853.322385 ms** exceeded a
**500 ms** whole-wrapper threshold. The fixture's service-close stubs each sleep
one second; wrapper time also includes actual journal and filesystem work. The
elapsed comparison does not directly prove whether service termination occurred.
Source inspection shows the successful-child path transfers its exact lease to
the reaper; the current test already observes direct close calls but fails on
timing before reaching that assertion. No runtime cleanup defect or particular
host-storage cause is established by this observation.

Corrective **WF-07f / T-033**, R03-033 / V-021, SHALL replace that timing inference
with direct assertions: successful child status, exactly one reaper delegation,
exact two-service lease, no direct service-close calls and retained cleaning state.
The test SHALL remove synthetic slow-close sleeps and the wall-clock threshold.
Its renamed exact selector SHALL be updated atomically in the authored catalog,
then projections regenerated through Make. Existing runtime deadlines, cleanup
ownership, production behavior and performance decisions SHALL remain unchanged.
This is test assurance repair under existing NLSpec Sections 11.5 / 13.5, not a
new runtime requirement. Entry lint `r03-s007f-entry-lint-01` failed (73098/2)
on a misplaced tracker table row; the row was moved into its proper validation
table before the corrected lint. The failed run SHALL remain preserved. Focused lifecycle
assurance, the containing browser owner, finalization/projections and closing
tracker lint SHALL pass before S-007e broader verification resumes.

Two read-only selection errors remain distinct: `target-plan` with undeclared
`OWNER` exited 2 before execution; corrected `task-guide DETAIL=coverage` passed.
An internal `TARGET=backend-unit DETAIL=logs` query exited 2; corrected public
`TARGET=test-slice` inspection exposed the bounded native Go failure. Neither
error dispatched tests or confers assurance. No rerun has replaced the failure.

### E-163 — WF-07f complete: direct cleanup-delegation assurance

Corrective entry Markdown lint passed **95847/0**, 9,490 ms, root
`r03-s007f-entry-lint-02`, after E-162's documented tracker-table correction.
The successful-child test is now
`TestRunDelegatesServiceTerminationAfterChildSuccess`. It asserts return status 0,
exactly one reaper call, no PostgreSQL or object-store close callback, both exact
service identities in the delegated lease, nonempty lease identity, unchanged
suite identity and pending cleanup, plus lifecycle state `cleaning` until the
cleanup owner settles. The 500 ms wall-clock inference and synthetic one-second
close sleeps are removed. The catalog retains the same row, owner, fixture,
resource and verification bindings; its exact native symbol is renamed without
an obsolete alias or duplicate selector. Runtime behavior and deadlines are
unchanged; existing NLSpec Sections 11.2, 11.5 and 13.5 remain authoritative.

The initial `make agent-finalize` attempt `r03-s007f-finalize-01` failed
**18044/2**, normalized child-target failure, because the authored catalog changed
and the generated topology input digest was stale. Its summary required
`make generate`; generated refresh was skipped after that prerequisite failure.
Its ten cleanup obligations and retained scan passed. A catalog-only check
`r03-s007f-catalog-01` passed **68658/0**, but does not substitute for successful
finalization. Public `make generate` then passed **83524/0**, 11,000 ms, root
`r03-s007f-generate-01`. This tool invocation has its own tool summary; no graph
cleanup or accounting receipt is fabricated for it.

The following checks use `CGO_ENABLED=0`, fresh explicit IDs under
`.cartulary/test-results/`, and disabled execution-result reuse for validation.
Graph checks other than finalization use basic diagnostics; the two test slices
also require invocation accounting.

| Run ID | Public command / selection | Native session / exit | Duration ms |
| --- | --- | --- | --- |
| `r03-s007f-finalize-02` | `make agent-finalize` | 28829 / 0 | 25,916 |
| `r03-s007f-catalog-02` | `make source-check CHECK=catalog` | 22793 / 0 | 1,646 |
| `r03-s007f-lifecycle-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract` | 16430 / 0 | 3,448 |
| `r03-s007f-browser-owner-01` | `make test-slice OWNER=harness.browser` | 66194 / 0 | 253,030 |
| `r03-s007f-generated-01` | `make source-check CHECK=generated` | 65756 / 0 | 16,977 |
| `r03-s007f-schema-01` | `make source-check CHECK=schema` | 98430 / 0 | 6,957 |
| `r03-s007f-go-lint-01` | `make lint CHECK=go` | 47278 / 0 | 8,386 |

All seven exact roots passed `make harness-observability-check CHECK=all`, exit 0,
`full_validity=1`. Finalization's optional diagnostics were off; its pass makes
no instrumentation claim. Focused lifecycle passed 1/1 units; containing browser
owner passed 50/50, including the formerly failing row. Their 10 / 83 cleanup
obligations completed, accounting scopes settled empty, and retained scans passed
across 17 / 896 files. All other graph checks also completed cleanup and passed
retained scans. `RESULTS_DIR` remained unset; retained-run maintenance was skipped.
E-162's failure remains a failure and is not erased by the repaired-source pass.

Current executable source:
`sha256:e5d75a501c01a2edf5f970065dec3fd32e39b11fec37d07cb527b8753df54695`.
Relative to E-161's restored source, only `tools/testservices/main_test.go`,
authored `tools/test_families/harness.browser.json`, and generated
`tools/execution_topology_render_index.json` differ. The generated path was
refreshed through Make. `s007f-authored-change.json`,
`s007f-restored-source-difference.json` and `s007f-validation-ledger.json` retain
the source and receipt map. No product, fixture runtime, scheduler, capacity,
retry, timeout or lifecycle optimization change is included. Experiment 10's
rejected sources and measurements retain their historical identities.

R03-033 / V-021 → successful-child exact delegation assertion → exact row and
containing owner → finalization, projections and Go lint is complete. WF-07f
assurance is PASS; no performance claim applies to this test-only correction.
After closing tracker Markdown lint, S-007e resumes with the remaining three
principal owner checks, explicit browser support / visual / UI-review, one final
CI on this retained source, independent audits, scans, settlement and handoff.
The successful current-source finalization and projection checks above satisfy
that ordering; no executable edits are planned before final CI.

### E-164 — Explicit UI-review failure; WF-07g accounting fixture correction

WF-07f boundary Markdown lint passed **68836/0**, 9,830 ms, root
`r03-s007f-boundary-lint-01`. On E-163's retained source, the command-surface,
test-catalog and evidence-accounting owner checks passed **14439/0**,
**63408/0**, **67286/0** in 70,424 / 50,538 / 12,758 ms. Exact roots are
`r03-s007e-command-owner-01`, `r03-s007e-catalog-owner-01` and
`r03-s007e-evidence-owner-01`. Explicit support passed **55640/0**, 121,637 ms,
root `r03-s007e-support-01`; visual passed **92558/0**, 637,797 ms, root
`r03-s007e-visual-01`, with unchanged goldens. All five passed independent public
`CHECK=all` audits, with full validity, settled accounting and successful cleanup.
Their source-bound passes are preserved; they do not close the following failure.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007e-ui-review-01
CARTULARY_HARNESS_CACHE_MODE=off make harness-contract SUITE=ui-review
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` failed **47040/2**,
normalized product/test_assertion_failure **10**, in 122,609 ms. Nine of ten units
passed. The sole failing row is
`harness.browser.boundary_support.ui_review_contract`. All 19 cleanup obligations
completed, accounting settled empty, and the retained scan passed across 61 files.
The exact root passed the diagnostics audit, exit 0, **full_validity=0**. No final
CI was launched after this failure.

The native failure is `invocation accounting placement unavailable` from
`verifyInvocationContainer`, reached by `test-renderer-identity.mjs` through
`startVisualRendererLease`. That fixture injects a fake Docker executor returning
`owned-id`, but the lease function resolves real kernel accounting membership and
asks the native owner to verify that fake container. Required accounting correctly
rejects it. The other UI-review assertions and real visual execution do not
establish that this partially mocked acquisition boundary is valid.

Corrective **WF-07g / T-034**, R03-034 / V-022, SHALL make accounting dependencies
explicit alongside the existing Docker, trust, readiness and browser dependencies.
Production defaults SHALL remain the native accounting parent resolver and exact
container verifier. The sole production caller constructs its options directly;
no public input, environment flag, schema field or accounting bypass SHALL be
added. The fake Docker fixture SHALL inject matching deterministic accounting
operations and test both accounted and unaccounted cases. It SHALL assert exact
parent placement at create, verification after start and before browser use,
verification rejection, and exact cleanup after rejection. Real required-accounting
visual execution SHALL validate the retained defaults. Runtime admission,
capacity, deadlines, cleanup ownership and measurement policy SHALL not change.

This is an internal infrastructure dependency correction under existing invocation
accounting and renderer ownership requirements; no new product behavior or
performance treatment is admitted. Entry tracker lint precedes implementation.
Focused UI-review contract assurance, containing UI-review, real visual validation,
finalization/projections and closing tracker lint precede final-CI resumption.
All failed evidence and the rejected lifecycle experiment remain preserved.

### E-165 — WF-07g completed; final CI entry boundary

**Assurance — PASS. Performance — NOT APPLICABLE.** Entry Markdown lint
`r03-s007g-entry-lint-01` passed **50308/0**, 9,880 ms, before implementation.
R03-034 / existing REQ-856 → the private renderer accounting dependency boundary
in `visual-renderer-lease.mjs` → the 58-case acquisition matrix in
`test-renderer-identity.mjs` → V-022 and this receipt provide traceability.
The production default remains native parent resolution and exact container
verification. The sole production caller supplies no accounting override.
No public input, environment bypass, legacy path, deadline or admission change
was introduced. Both accounted and unaccounted fake fixtures are deterministic;
verification occurs after start and before browser use. Verification rejection
prevents browser use and removes the exact owned container. Existing cleanup
failure preservation remains tested. Real visual execution validates the native
defaults with required accounting and unchanged goldens.

Initial finalization `r03-s007g-finalize-01` failed **5995/2**, 4,150 ms, because
the generated topology index still contained the old renderer source digest.
Its typed summary required `make generate`; all ten cleanup records completed,
and its 11-file secret scan passed. Public `CHECK=diagnostics` passed with
**full_validity=0**. No product test ran before repairing this projection.
`make generate`, root `r03-s007g-generate-01`, passed **34206/0**, 10,900 ms.
This generation receipt is a tool summary; no graph cleanup or scan is invented.
A logs query without `TARGET` returned usage exit 2; the corrected public
`TARGET=agent-finalize DETAIL=logs` query succeeded. These attempts are preserved.

The retained source is
`sha256:1df9bd0269720e06e83d70b2d29fc627c290db5f26ac06544483515990255367`
at unchanged HEAD `f42fc00b76ec4e296c2a379c09687e02ce669de7` with authorized
working-tree changes. Every command below used `CGO_ENABLED=0`; test invocations
also used execution cache off, basic diagnostics and required accounting.
Source checks used basic diagnostics. Finalization used its default diagnostics.
Roots are `.cartulary/test-results/r03-s007g-<label>`.

| Label | Exact public command | Native session/exit | Graph ms | Passed units | Completed cleanup / scanned files |
| --- | --- | --- | --- | --- | --- |
| `finalize-02` | `make agent-finalize` | 45139/0 | 26192 | 1 | 10 / 12 |
| `catalog-01` | `make source-check CHECK=catalog` | 25411/0 | 1631 | 2 | 11 / 20 |
| `contract-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract` | 65156/0 | 2803 | 2 | 11 / 21 |
| `ui-review-01` | `make harness-contract SUITE=ui-review` | 69490/0 | 109897 | 10 | 19 / 61 |
| `visual-01` | `make browser-e2e SUITE=visual` | 77252/0 | 624003 | 37 | 74 / 545 |
| `generated-01` | `make source-check CHECK=generated` | 71340/0 | 16589 | 5 | 14 / 34 |
| `schema-01` | `make source-check CHECK=schema` | 48588/0 | 6166 | 3 | 12 / 24 |
| `browser-owner-01` | `make test-slice OWNER=harness.browser` | 83882/0 | 236234 | 50 | 83 / 896 |
| `command-owner-01` | `make test-slice OWNER=harness.command_surface` | 20425/0 | 69342 | 10 | 19 / 270 |
| `catalog-owner-01` | `make test-slice OWNER=harness.test_catalog` | 48015/0 | 51581 | 2 | 11 / 21 |
| `evidence-owner-01` | `make test-slice OWNER=harness.evidence_accounting` | 33973/0 | 12952 | 12 | 21 / 63 |
| `support-01` | `make browser-e2e SUITE=support` | 21893/0 | 122998 | 14 | 27 / 142 |

All twelve passing roots independently passed public
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, exit 0,
**full_validity=1**. All cleanup records completed, every final secret scan passed,
and every required-accounting invocation settled **empty**. Basic execution and
resource diagnostics are complete without omissions. Finalization diagnostics
are explicitly unavailable because they were off; source checks did not request
invocation accounting. These limits do not become resource guarantees.
`s007g-validation-ledger.json` and `s007g-public-check-receipts.json` retain exact
receipts under `.cartulary/remediation03/continuation`.

The E-164 failure is resolved by a complete fixture dependency boundary; it is
not reclassified as a passing run. E-160/E-161's lifecycle rejection and coordinated
restoration remain final. No candidate-only lifecycle behavior is retained.
`RESULTS_DIR` was unset for finalization; retained-run maintenance was skipped
because no compatible successful full warm-check root exists.

**Next entry.** Run `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=r03-s007g-boundary-lint-01 make lint-markdown` after this
tracker update. Only its successful outer exit admits S-007e's final
`make ci` with cache off, basic diagnostics and required accounting, using fresh
run ID `r03-s007e-ci-01`. That CI is retained-source correctness/integration
validation, not a CI performance baseline or renewed lifecycle experiment.
S-007e still requires terminal evidence audit, physical settlement, final scan,
resumable handoff and final tracker Markdown lint.

### E-166 — S-007e final integration and evidence acceptance

WF-07g closing tracker lint `r03-s007g-boundary-lint-01` passed **91687/0**,
9,270 ms, before final CI began. E-165's current-source finalization, projection,
owner and explicit-suite assurance is retained without further executable edits.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007e-ci-01
CARTULARY_HARNESS_CACHE_MODE=off make ci HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required` completed **36651/0**. Root
`.cartulary/test-results/r03-s007e-ci-01` reports **600/600 units passed**, zero
failed/cancelled/skipped units, and graph duration **3,767,878 ms**. This was
retained-source correctness/integration validation. It is not a CI performance
baseline, matched acceptance sample or replacement for E-160/E-161's decisions.

| Identity / terminal gate | Observed result |
| --- | --- |
| Executable source | `sha256:1df9bd0269720e06e83d70b2d29fc627c290db5f26ac06544483515990255367` |
| Runtime graph | `sha256:a87855b3f83ca00204467fdcfbbed031f4c0987a3e2b4b3245b983faa06ad818` |
| Independent evidence audit | `make harness-observability-check RESULTS_DIR=.cartulary/test-results/r03-s007e-ci-01 CHECK=all`; **79714/0**, **full_validity=1** |
| Physical cleanup | **1,163 completed** records; zero failed, blocked or unresolved records |
| Invocation accounting | Child exit 0; scope settled **empty**; no signal |
| Final retained-secret scan | **PASS**, 9,790 files; nested terminal scan outcomes also passed |
| Required diagnostics | Canonical critical path and allocation lifecycle **complete**, zero allocation omissions; invocation accounting available |
| Optional diagnostic limits | Execution activity truncated; omitted execution count unavailable; resource sampling truncated with **757** omissions; PostgreSQL checkpoints unavailable |
| Maintenance | Skipped because `RESULTS_DIR` was unset; no compatible successful full warm-check root exists |

Native observation identified a quiet interval after 52 completed units; structured
polling then stopped until the owning command returned. No competing intentional
workload, source mutation, capacity adjustment or deadline change was introduced.
An assumed performance reference was not treated as a command timeout.

**Descriptive timing and resource observation — not a new performance decision.**
The complete root critical path contains 80 units, 773,320 ms graph queue time
and 2,993,127 ms unit lifetime. Graph queue reasons reconcile exactly:
659,622 ms resources, 113,255 ms exclusive barrier, 443 ms rank, zero scheduler
stop. Host-admission phase durations total 1,650,220 ms; runner 1,073,364 ms;
acquisition 237,722 ms; release 27,004 ms. Uncovered unit lifetime is 2,495 ms.
These categories retain their scopes; phase sums and simultaneous blockers MUST
NOT be generalized into additive exclusive whole-run causes.

The path's allocation view contains **25** complete, settled generations:
**24 capacity-pressure** retirements and **one terminal-borrower release**.
Twelve generations have only reset-unit borrows. These are path-associated
allocations, not a whole-run allocation count. This observation motivates future
reservation/residency investigation; it does not authorize a scheduler, capacity,
isolation or pooling change or revive the rejected lifecycle candidate.

The accounting scope reports 3,850,147.979938 ms elapsed, 11,769,754,609 charged
CPU microseconds and 9,437,315,072 bytes kernel peak memory. This averages about
3.06 charged cores within that scope; shared host daemons, unrelated host work
and controller setup/publication remain excluded. Ready-to-start p95 is
3,180,216 ms and cleanup union is 161,211 ms. Partial samples are not promoted to
complete per-process or whole-host utilization. No percentages against the
assumed CI reference or new confidence bounds are calculated for this run.

Public `make explain-run` exports `DETAIL=performance JSON=1`, and
`DETAIL=critical-path VIEW=units|allocations LIMIT=100 JSON=1`, all passed with
native receipts **7155/0**, **12505/0**, **58314/0**. Unit and allocation views
return 80/80 and 25/25 records, each with `next_offset=null`.
`s007e-ci-performance.json`, `s007e-ci-critical-units.json`,
`s007e-ci-critical-allocations.json` and `s007e-ci-descriptive-analysis.json`
retain those observations under `.cartulary/remediation03/continuation`.
The internal arithmetic consumes validated public exports, not canonical events.

**Preservation and migration.** `retained-source-final.tar.gz` preserves 7,502
source files with archive SHA-256
`31bf07dab1b958fc330985255ee04c14a8c427a44364a6239f24139cab629089`.
Its companion manifest binds the current executable source and unchanged HEAD
`f42fc00b76ec4e296c2a379c09687e02ce669de7`. All 7,501 non-tracker paths still match
that archive after CI. Relative to experiment 10's restored diagnostic baseline,
only E-163's cleanup-delegation test/catalog change, E-165's renderer dependency
boundary/test change and their shared generated topology index differ. The
360-path rejected lifecycle treatment remains restored. No golden refresh,
compatibility toggle, scheduler ranking/capacity change or service pool remains.
All earlier failures, cancellations, omitted attempts and source snapshots retain
their original meanings.

**Completion trace.** R03-019/020 → accepted diagnostics and current owner checks;
R03-021–023 → E-160/E-161 rejection and coordinated restoration;
R03-032 → E-149–E-154 diagnostics, NO_CHANGE transport and 192/192 qualification;
R03-033/034 → E-163/E-165 causal verification repairs;
R03-024 → E-165 finalization/explicit assurance and this final CI/audit/settlement
receipt. V-006/V-008/V-010–012/V-020–022 remain separately attributable.
The sole remaining closure step is the final tracker update and Markdown lint.

### E-167 — Continuation completion and resumable handoff

**WF-07e / S-007e — COMPLETE.** E-166's final integration receipt and all earlier
required current-source checks passed. Completion Markdown lint
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007e-completion-lint-01
make lint-markdown` passed **69058/0**, 9,630 ms. This final tracker update records
the completed work and registers one closing lint over the final text:
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s007e-final-lint-01 make lint-markdown`.
Completion requires that command's successful outer exit; failure reopens document
closure. No executable change or additional product run is part of this step.

| Outcome | Final disposition and evidence |
| --- | --- |
| Diagnostics | ACCEPTED, E-081/E-149; complete paged canonical critical path and separately bounded allocation lifecycle observations; E-165/E-166 current validation |
| Stability plan, Section 7.7 | COMPLETE, E-149–E-157; publication transport NO_CHANGE; 192/192 lifetimes qualified; original failure not causally reproduced |
| Reset/cohort lifecycle | REJECTED and RESTORED, E-160/E-161; browser improvements 1.5743% / 1.3265%; CI 1.5456% / -2.6484% against the assumed 60-minute reference; both fail the first-two <5% rule |
| Performance scope | Four-pair maximum retained; six optional attempts omitted after mandatory rejection decisions; no omitted bootstrap bounds or CI resource-regression guarantee claimed |
| Final corrective assurance | E-163 direct cleanup-delegation test; E-165 complete renderer dependency fixture; native accounting defaults and cleanup authority preserved |
| Current verification | E-165 finalization, generated/schema/catalog, four owner routes, support, visual and UI-review all pass; E-166 final CI 600/600 and independent audit pass |
| Settlement and scans | All required invocations settled; final CI 1,163 cleanup records completed, accounting empty, 9,790-file scan and all 15 terminal graph scan memberships passed |
| Maintenance | SKIPPED because `RESULTS_DIR` was unset; no compatible successful full warm-check root exists |
| Source and migration | E-166 immutable final archive and source identity; rejected owner/code/projections restored together; no live invocation remains at cutover or handoff |
| Remaining uncertainty | Historical endpoint and host-latency causes are not proven; optional diagnostic truncation is disclosed; claims are not utilization; finite qualification is not a universal reliability guarantee |

**Resumption boundary.** There is no pending implementation, performance sample,
required correctness repair or candidate-owned resource. New work SHALL start
from E-166's retained source or explicitly revalidate a newer checkout. A future
reservation/admission or allocation-residency experiment requires its own owner
proposal, exact workload/treatment freeze, safety proof and acceptance decision;
this handoff does not authorize capacity, ranking, isolation, worker or pooling
changes. Historical artifacts SHALL retain their actual source/schema identity.
No old run becomes a current-source pass, and no restored candidate is accepted
without a new explicitly authorized experiment.

The sole controlling artifact remains this twelve-section tracker. Original
completion E-073 and all intervening receipts remain intact. The adopted NLSpec
owns behavior; schemas, generated projections and tests remain downstream.
Executable tooling does not consume this Markdown. Changes remain in the
working tree at the recorded HEAD; no commit, external publication or golden
refresh is claimed. `s007e-validation-ledger.json` retains the final native exits,
audit outcome, diagnostic coverage and source-preservation references.

### E-168 — Reservation continuation registration

2026-10-23: user-authorized implementation begins at clean HEAD
`f7b917026e1174661a9bdf9118d18120bbd225b0`. Entry source manifest preserves 7,502
tracked/nonignored files under `.cartulary/remediation03/reservations/entry-source.json`.
No pre-existing changes. `make help ALL=1` exits 0 and confirms public discovery,
reader, validation and execution surfaces. Applicable root/tools/backend
instructions and test-operations skill reloaded. Section 7.8 and F-020–024,
C-010–013, R03-035–041, S-012–016, V-023–027 and T-035–040 register the approved
serial execution. Historical E-167 remains complete; this scope is separate.
No implementation, test pass or performance benefit is yet claimed.
Registration lint `r03-wf00f-register-lint-01` passed native **72407/0**,
17,810 ms. This receipt closes WF-00f assurance, with no runtime/performance
claim and no source migration. The four owner routes were discovered through
public Make. Closing lint `r03-wf00f-close-lint-01` MUST pass after this update
before WF-12 owner adoption and implementation begin.

### E-169 — WF-12 owner adoption

WF-00f closing lint passed native **58478/0**, 9,900 ms, root
`r03-wf00f-close-lint-01`. All four public task-guide commands exit 0.
Before projection/implementation changes, REQ-803/857 and AC-153 adopt the
host-admission observation boundary, defaults, independently bounded section,
coalesced decision observations and disjoint phase aggregates. Source claims and
admission behavior remain unchanged in this slice. F-020/F-024 → C-010/C-013 →
R03-035/039 → S-012/V-023. Numeric non-CPU sizing is NO_CHANGE, not calibration.

### E-170 — WF-12 implementation and focused assurance (open)

Native admission captures decision facts inside its existing evaluator and emits
observations after releasing its mutex. Graph queue reservations remain canonical
scheduler events. `VIEW=host-admission` consumes a separately bounded section,
with opaque references, explicit coverage, coalesced snapshots, disjoint phase
durations and record-boundary pagination. Index v4, policy v5, critical-path v2,
and admission-record v1 replace current readers atomically; runtime-plan v6's
unchanged clock shape references index v4. No historical translation is added.

| Resource | Authored claim / capacity owner | Lifetime and decision |
| --- | --- | --- |
| CPU / I/O | Topology resource profiles; resource registry host/CPU-derived capacities | Active admission and graph reservations only. Residents carry neither; CPU budget treatment waits for S-014. I/O unchanged. |
| Memory / processes | Topology active and physical profiles; host/cgroup/process-limit-bounded registry capacities | Active claim plus weighted resident claims through exact cleanup. Values unchanged; low utilization supplies no sizing proof. |
| Ports / browser stacks | Topology physical profiles; port-lane-derived capacities | Physical residency through settlement, including idle stacks. Values unchanged. |
| PostgreSQL / object store | Topology service/physical profiles; bounded registry lanes | Active requirements and exact resident service generations; borrowed/persistent owners retain their own claims. Values unchanged. |
| Service stack / volume | Runtime profile and writable-root capability | Service residency; volume remains non-additive capability. Values unchanged. |
| Host admission / ancestors | Native weighted admission owner | Effective capacities retain governing ancestor envelopes; private ownership cannot be inferred from public diagnostic names. Algorithm unchanged in S-012. |

Canonical command receipts so far (all roots under `.cartulary/test-results`):

| Command / selection | Root / native handle | Exit / outcome |
| --- | --- | --- |
| `make generate`, first attempt | No run emitted | 2; stale runtime-plan reference to index v3; corrected before retry |
| `make generate` | `20261023T122627Z-p79564`, 82106 | 0; 11,930 ms |
| `make format RUN_ID=r03-s012-format-01` | No run emitted | 2; format does not declare RUN_ID; no execution |
| `CARTULARY_TEST_RUN_ID=r03-s012-format-02 make format` | `r03-s012-format-02`, 44064 | 0; graph 8,042 ms |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission,harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic` | `r03-s012-admission-evidence-01`, 35627 | 0; graph 13,613 ms |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime HARNESS_DIAGNOSTICS=basic` | `r03-s012-native-runtime-01`, 6128 | 0; graph 54,731 ms |
| `make explain-run RESULTS_DIR=.cartulary/test-results/r03-s012-admission-evidence-01 DETAIL=critical-path VIEW=host-admission JSON=1` | Exact preceding run | 0; complete admission coverage, zero omissions, six path records |

Execution receipts use `CGO_ENABLED=0`, fresh `CARTULARY_TEST_RUN_ID`, and
`CARTULARY_HARNESS_CACHE_MODE=off` for test slices. Passing focused tests do not
close S-012: subsequent bounded-failure/rounding refinements, finalize, focused
physical pilot, generated checks, evidence audits and closing lint remain due.
No performance benefit is claimed. No active test process remains from these
receipts. Final candidate resource settlement still requires retained audits.

### E-171 — WF-12 completion and reservation-sizing disposition

**S-012 implemented; focused assurance passed; no performance benefit claimed.**
REQ-803/844/857 and AC-153 own the behavior. Native admission captures weighted
decisions and transaction outcomes; fixture owners report actual retirement
sets/triggers/settlement; runtime admission owns bounded observations; the
canonical reader supplies host intervals and the retained reader validates the
independent section. Critical-path v2 projects the same validated model for
human/JSON pagination. V-023 covers these routes and the exact claim audit in
E-170. F-020/F-024 and R03-035/039 are satisfied for this workstream; F-021–023
remain assigned to their subsequent slices.

Post-refinement commands all use repository-root public Make, `CGO_ENABLED=0`,
fresh environment `CARTULARY_TEST_RUN_ID`, and execution-cache off for tests.
`HARNESS_DIAGNOSTICS=basic` applies to the three test selections below; only the
service pilot additionally uses `HARNESS_ACCOUNTING=required`.

| Exact public command / selection | Run ID / native session | Exit; graph/tool duration |
| --- | --- | --- |
| `make generate` | `r03-s012-generate-02` / 74829 | 0; 11,240 ms |
| `make format` | `r03-s012-format-03` / 25033 | 0; 7,934 ms |
| `make agent-finalize` | `r03-s012-finalize-01` / 56982 | 0; 28,215 ms |
| `make source-check CHECK=generated` | `r03-s012-generated-01` / 86841 | 0; 16,938 ms |
| `make source-check CHECK=schema` | `r03-s012-schema-01` / 17851 | 0; 6,589 ms |
| `make source-check CHECK=catalog` | `r03-s012-catalog-01` / 8275 | 0; 1,780 ms |
| E-170 exact two-row evidence/admission selection | `r03-s012-admission-evidence-02` / 8107; `r03-s012-admission-evidence-03` / 84003 | 0 / 0; 13,491 / 29,130 ms |
| E-170 exact native-runtime selection | `r03-s012-native-runtime-02` / 96118 | 0; 55,369 ms |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_graph_observations` | `r03-s012-service-pilot-01` / 34263 | 0; 105,543 ms |
| `make format` | `r03-s012-format-04` / 10621 | 0; 7,815 ms |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` | `r03-s012-admission-facts-01` / 86416 | 0; 2,313 ms |

The retained executable source for finalize, generated/schema/catalog checks,
final evidence/runtime assurance and the physical pilot is
`sha256:e799f7b221a696eea0f5a98157309c3d33a240359bb978079ba72e85b8700702`.
The last facts-only run adds diagnostic assertions to existing weighted-tree
cases for ancestor envelopes, persistent memory deficits, non-additive volume
and priority-versus-capacity attribution. It changes no executable behavior.
Its exact manifest and `.cartulary/remediation03/reservations/s012-source.json`
preserve the post-assurance source; `s012-source.tar.gz` preserves authored and
generated files together. These internal authoring snapshots are not runtime
inputs and introduce no Markdown dependency.

`make harness-observability-check RESULTS_DIR=<exact-root>` exits 0 for evidence
03, native-runtime 02, service-pilot 01 and both nested pilot roots
`native-boundary/native-observation-postgres` and
`native-boundary/native-observation-https`. The pilot roots contain complete
host-admission coverage with respectively 3, 21 and 36 records and zero omissions;
the HTTPS path view exposes 18 records with paging. The nested runs independently
validate PostgreSQL checkpoint observations and required invocation accounting.
Outer/pilot PostgreSQL/pilot HTTPS cleanup closes 10/18/23 obligations, all
`completed`; secret scans pass 163/47/97 files respectively. Evidence/runtime
checks likewise settle all recorded cleanup and pass scans (26/74 files).
`make explain-run RESULTS_DIR=<exact-root> DETAIL=critical-path
VIEW=host-admission JSON=1 LIMIT=1` passes for evidence 03 and the nested HTTPS
pilot, preserves complete-path totals and returns `next_offset=1`.

**Disposition and boundaries.** Memory, process, port, PostgreSQL, object-store,
service and host capacity quantities remain **NO_CHANGE**. The audit establishes
ownership/lifetime, not that current values are optimal. Historical low container
utilization does not prove excess reservations. No eBPF collector, tuning,
capacity relaxation, scheduling change or lifetime correction ships in S-012.
Pilots are excluded from acceptance. Early prototype runs remain source-bound
historical evidence; current readers do not translate earlier draft shapes.
Migration is atomic and no live invocation spans the source cutover. No
candidate-owned resource from this slice is unresolved. Agent-finalize retained
maintenance was skipped because `RESULTS_DIR` was unset; no compatible successful
full warm-check root exists.

WF-13 may begin only after this receipt and `r03-s012-close-lint-01` pass. Its
first action SHALL adopt exact retirement evaluation and complete consumer
accounting in REQ-803/804; optional Go budgeting remains blocked by S-013.

### E-172 — WF-13 entry and owner adoption

S-012 closing Markdown lint passed, native **63059/0**, 16,360 ms, root
`r03-s012-close-lint-01`. Its preserved post-assurance source contains 7,506 files,
archive 55,086,449 bytes, and executable source digest
`sha256:1346f609a8a1b26cf7f40735de928a4212c5716132bacca6eeecedd4554754a6`.
S-013 now adopts REQ-803/804 and AC-085/086 before changing projections or code:
exact counterfactual retirement, unified stable idle selection, graph-complete
idempotent terminal accounting, exact compatibility keys and joined fixture-owned
cleanup dispatch. The graph event owner must not wait on physical cleanup. This
is a required lifecycle correction, independent of S-014 performance acceptance.
F-021/F-022 → C-011/C-012 → R03-036/037 → S-013/V-024. S-012's unchanged admission
and real-service pilot provide characterization; existing weighted-tree, renderer,
broker and runtime rows remain the narrow validation owners. No implementation
or performance acceptance is yet claimed for S-013.

Pre-change characterization `r03-s013-characterization-01` passes native
85075/0 (2,485 ms) after format 14016/0 (8,027 ms). The exact selection is
`make test-slice OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.observation_admission
HARNESS_DIAGNOSTICS=basic`, with CGO disabled and execution-cache off. A newly
added weighted-tree case demonstrates the defect: an 800 MiB ancestor envelope
masks a 200 MiB nested resident, so retiring that resident cannot reduce the
pending request's 100 MiB deficit; the old quantity-only query nevertheless
returns true. S-013 SHALL change that assertion to require false under exact
counterfactual evaluation, while preserving the rest of the weighted accounting.



### E-173 — S-013 counterfactual assurance and consumer implementation

The native retirement request now names exact immutable resident records. The
same weighted evaluator runs before/after removal, reintroduces prospective
replacements and retains required ancestors. Its decision joins independently
bounded admission observations; execution index v5, host record v2 and critical
path v3 migrate together through `make generate`. No quantity fallback remains.

| Command / selection | Run ID / native handle | Outcome |
| --- | --- | --- |
| `make generate` | `r03-s013-generate-01` / 21313 | 0; 12,300 ms |
| `make format` | `r03-s013-format-01` / 56433 | 0; 8,075 ms |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission,harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic` | `r03-s013-exact-retirement-01` / 16410 | 2; 14,865 ms; stale quantity-only test request rejected |
| `make format` | `r03-s013-format-02` / 46382 | 0; 8,052 ms |
| Same evidence selection after exact-record fixture repair | `r03-s013-exact-retirement-02` / native handle unavailable | Terminal summary pass; 15,348 ms; outer exit observation lost during tool-output truncation |
| `make format` | `r03-s013-format-03` / 33509 | 0; 7,989 ms |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures,harness.browser.unit.fixture_broker_smoke HARNESS_DIAGNOSTICS=basic` | `r03-s013-fixtures-01` / 43695 | 2; 17,706 ms; dedicated provisional-slot pressure loop |
| `make format` | `r03-s013-format-04` / 34535 | 2 before graph creation; incomplete Go field edit failed native bootstrap |
| `make format` after field repair | `r03-s013-format-05` / 75232 | 0; 7,988 ms |
| `make format` after runtime terminal matrix | `r03-s013-format-06` / 37543 | 0; 8,000 ms |

All execution commands use `CGO_ENABLED=0`; test selections also disable execution
cache. Roots use `.cartulary/test-results/<run-id>`. Failed attempts remain intact.
Exact-retirement-02 source is
`sha256:2148b0532cc13e50c089a9f63e50adca680374a9567b3de8994bdfdae831304c`;
`make harness-observability-check RESULTS_DIR=<that-root>` independently passed
with full validity and complete diagnostics. It is focused assurance, never a
performance sample or proof of its unobserved outer command exit.

The new consumer registry registers selected graph units, centralizes stack and
renderer compatibility, records idempotent terminal dispositions and dispatches
cleanup through one joined fixture worker. Compiler cleanup markers are removed;
separate reset units remain. The first fixture run caught a provisional renderer
slot repeatedly winning pressure selection without releasing physical capacity.
The repair excludes unallocated slots from capacity-pressure candidates while
preserving bounded provisional reassignment. Its rerun and additional terminal
matrices are pending. Fixture-01 cleanup reported 13 completed steps, none failed
or blocked. S-013 remains IN_PROGRESS; no performance claim or successor entry.

S-013 follow-up assurance (same public selections as above): fixtures-02
**53884/0**, 69,075 ms; exact-retirement-03 **93268/0**, 79,722 ms;
runtime-01 **77338/2**, 69,079 ms. Fixture/evidence readers passed independent
`CHECK=all`; failed runtime diagnostics passed `CHECK=diagnostics`. These
correctness runs shared the host and their durations are not comparative samples.
The runtime failure was the synthetic continuation fixture's missing consumer
registration: its adapter now uses the registered graph owner, preserving the
required two generations across the injected failure. Format-07 **66607/0**,
8,050 ms. The new terminal matrix itself passed, including failure before launch,
publication failure, cancelled/never-started units, cached and skipped units.

Two attempted combined commands, native-assurance-01 (comma-separated OWNER)
and native-assurance-02 (ROWS without OWNER/SUITE), failed usage validation with
native exit 2 before run creation. They executed no tests. Required reruns use
the documented single-owner routes; no selector or public interface changed.

After repair, fixtures-03 **24530/0** (115,611 ms), runtime-02 **78265/0**
(116,931 ms), and catalog-01 **79812/0** (51,540 ms) pass. The catalog route is
`make test-slice OWNER=harness.test_catalog
ROWS=harness.test_catalog.behavior.routing_contract HARNESS_DIAGNOSTICS=basic`.
The other selections match their prior owner routes. Format-08 **87281/0**,
7,935 ms, adds explicit failed asynchronous settlement/claim retention, actual
terminal-trigger attribution and non-additive volume retirement assertions.
Those new assertions require their own pass before S-013 closes.

Traceability correction: E-172's final finding label refers to **F-022/F-023**
(retirement and consumer lifecycle). F-021 belongs to the still-pending Go-budget
candidate. This corrects the label without changing the preserved receipt or IDs.


### E-174 — WF-13 completion and mandatory retention baseline

F-022/F-023 → REQ-803/804, AC-085/086 → exact native retirement evaluation,
fixture-owned terminal registry/cleanup dispatcher and graph terminal transitions
→ V-024. Mandatory correctness is **PASS**; implementation is **ACCEPTED**.
Performance is **NOT ESTABLISHED**. S-014 optional Go budgeting remains unchanged
and may begin only after this receipt's closing Markdown lint passes.

The registry freezes every selected graph unit before admission. Stack reuse and
consumer registration now use one exact compatibility identity; renderer counts
bind profile, lifetime and diagnostics within immutable run/trust ownership.
Terminal accounting covers cache, dependency skip, pre-launch failure, cancellation
and publication failure. Identical duplicates are idempotent; unknown, conflicting
or active-borrow notifications fail. A bounded, coalesced fixture worker performs
admitted cleanup and joins before inventory reconciliation. It cannot start a
second sweep after shutdown begins. Shared and measurement services wait for
all consumers and dependent allocations; external persistent owners remain separate.

Renderer/stack pressure selection is joint, unused-first and stable by logical
key. A reusable pressure candidate must improve exact persistent deficits;
provisional slots release no physical capacity and are excluded from pressure
selection. The pinned stack retirement decision rechecks zero-consumer state so
a concurrent last-terminal notification produces the actual mandatory retirement
reason and trigger instead of an unsupported capacity-relief claim. Failed cleanup
retains native resident claims and prevents reuse. Static release markers and
profile-only counters are removed. Separate reset units and boundaries remain.

| Public command / selection | Run ID / native handle | Exit; duration |
| --- | --- | --- |
| E-173 fixture selection with failed-dispatch and trigger tests | `r03-s013-fixtures-04` / 90827 | 0; 12,066 ms |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_graph_observations,harness.browser.integration.native_nested_cancellation,harness.browser.integration.native_selected_service_lifecycle HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` | `r03-s013-boundary-01` / 88970 | 0; 166,254 ms |
| `make generate` | `r03-s013-generate-02` / 34057 | 0; 11,270 ms |
| `make format` | `r03-s013-format-09` / 85875 | 0; 8,048 ms |
| E-173 fixture selection after shutdown and stable-key assertions | `r03-s013-fixtures-05` / 75794 | 0; 20,105 ms |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission HARNESS_DIAGNOSTICS=basic` | `r03-s013-admission-final` / 58298 | 0; 2,761 ms |
| `make source-check CHECK=generated HARNESS_DIAGNOSTICS=basic` | `r03-s013-generated-01` / 13053 | 0; 35,344 ms |
| `make source-check CHECK=schema` | `r03-s013-schema-02` / 39762 | 0; 6,668 ms |
| `make source-check CHECK=catalog` | `r03-s013-catalog-02` / 55059 | 0; 1,819 ms |
| `make format` | `r03-s013-format-10` / 75833 | 0; 7,932 ms |
| E-173 fixture selection with deterministic terminal/pressure race | `r03-s013-fixtures-06` / 59827 | 0; 20,384 ms |

All execution uses `CGO_ENABLED=0` and test execution-cache off. The earlier
`make schema-check` attempt returned usage exit 2 (no such target, no run root);
the canonical replacement above passed. Exact roots are
`.cartulary/test-results/<run-id>`. Final fixture source is `sha256:496b69d59ce04a644f34a3d23cf3ce80ece293b78dceab4531c11f8eedf337fa`.

Independent `make harness-observability-check RESULTS_DIR=<root>` passed for
fixtures-04/05, runtime-02, catalog-01 and boundary-01. Both nested boundary roots,
`native-boundary/native-observation-postgres` and `native-observation-https`, also
passed independently with invocation accounting and PostgreSQL checkpoints.
The boundary outer invocation has 14 completed cleanup records, zero failed or
blocked, and a passing 504-file retained-secret scan. Fixtures-04 has 13 completed
and a passing 33-file scan; runtime-02 has 13 completed/90 scanned; catalog-01
has 11 completed/24 scanned. The final fixture run has 13 completed, zero failed
or blocked and a passing 33-file scan. Injected failure tests preserve their
expected failures inside the assurance outcome; no candidate-owned physical
resource remains unresolved.

The diagnostics and mandatory lifecycle corrections form the measured baseline
for S-014/S-015. Preserve this source under
`.cartulary/remediation03/reservations/s013-source.{json,tar.gz}` after closing
lint. This archive is an authoring preservation artifact, never runtime input.
Rollback of optional Go tuning SHALL retain these mandatory fixes. Non-CPU
quantities, scheduler policy, browser workers/lanes, PostgreSQL lifetimes and
isolation eligibility remain unchanged. Maintenance remains skipped because
`RESULTS_DIR` is unset: no compatible full warm-check root has been established.


### E-175 — WF-14 entry and Go budget owner adoption

S-013 closing lint **20987/0**, 17,650 ms, passed at
`r03-s013-close-lint-01`. Its preservation manifest/archive contain 7,509 files
and 48,204,551 compressed bytes, with executable source
`sha256:496b69d59ce04a644f34a3d23cf3ce80ece293b78dceab4531c11f8eedf337fa`.
Final fixtures-06 independently passed `harness-observability-check CHECK=all`.
No live invocation remains at cutover.

Before implementation, REQ-379/805 adopt per-group profile-owned CPU budgets,
explicit qualification feasibility, matching CPU/GOMAXPROCS/Go parallelism and
independent publication workers. REQ-354's older equality between PostgreSQL and
CPU claims is reconciled: PostgreSQL reservations remain separately bounded by
authored quantity, captured PostgreSQL capacity and admitted Go entitlement;
lowering an ordinary CPU profile does not lower that PostgreSQL reservation.
Current authored non-CPU quantities and host capacities remain untouched. A
before/after compiled-population check SHALL prove non-CPU claims remain equal
for the frozen test-fast/browser/CI workloads. Explicit contexts retain their
exact owner budget and reject infeasibility rather than clamp.

Characterization is the preserved S-013 source plus the passing catalog graph
contract (E-173/E-174), whose old `go_lpt_budget` case explicitly establishes
selection-count-derived global worker/GOMAXPROCS fields. The current planner
computes `min(group_count, min(8, floor(available/4)))` workers and assigns each
child `floor(available/workers)` CPU tokens; profile CPU is overwritten. This
is the sole optional performance treatment. S-014 implementation and S-015
performance acceptance remain pending; mandatory S-013 fixes cannot be rolled
back to recover historical timing. F-021 → REQ-379/805/354/803 → S-014/V-025.

### E-176 — S-014 candidate implementation and initial assurance

The v3 Go plan removes global child worker/GOMAXPROCS fields and binds each
resolved CPU budget into group identity. One resolver owns profile/host/parent/
PostgreSQL bounds and explicit-context feasibility. The compiler projects its
claims and matching environment; the runner validates declared binding and emits
explicit `-parallel` plus bounded package-build `-p`. Publication workers are
unchanged. No non-CPU authored quantity or public command changed.

`make generate`, root `r03-s014-generate-01`, **17322/0**, 11,270 ms;
`make format`, `r03-s014-format-01`, **74807/0**, 8,122 ms. Both owner routes use
CGO disabled, cache off and basic diagnostics. Catalog graph assurance
`make test-slice OWNER=harness.test_catalog
ROWS=harness.test_catalog.behavior.routing_contract` passed at
`r03-s014-catalog-01`, **51300/0**, 57,187 ms. It covers direct/owner/partial/
aggregate/empty selections, profile bounds, exact-context rejection, unchanged
membership, ordering invariance and semantic identity changes.

Runner/public-surface assurance `make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.behavior.public_registry_parity` failed at
`r03-s014-runner-01`, **28395/2**, 66,480 ms. The sole failure was an old literal
argv expectation that omitted the newly required `-parallel=2`; 46 of 47 Node
tests passed. Cleanup completed all 11 steps. The expected argv is corrected;
the failed attempt remains retained. Actual scheduled Go-child checks now also
compare native GOMAXPROCS and the live Go test parallel flag with the declared
CPU budget, under existing scheduler/runtime assurance rows. Further validation
and performance acceptance remain pending.

### E-177 — WF-14 assurance completion; performance remains pending

F-021 → REQ-354/379/803/805 → Go claim resolver, compiler, plan v3 and
runner → V-025 catalog, runner and actual-child budget checks. The candidate
is implemented and focused correctness passes; acceptance remains conditional
on S-015. Original S-013 behavior is preserved for coordinated optional rollback.
No mandatory lifecycle correction is subject to that rollback.

Commands below run from the repository root with `CGO_ENABLED=0` and unique
`CARTULARY_TEST_RUN_ID`; focused tests also disable execution cache and enable
basic diagnostics. Roots are `.cartulary/test-results/<run-id>`.

| Command / selection | Run ID / native handle | Outcome |
| --- | --- | --- |
| `make format` | `r03-s014-format-02` / 24990 | 0; 8,007 ms |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity` | `r03-s014-runner-02` / 85159 | 0; 68,129 ms |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime,harness.command_surface.unit.scheduler_smoke` | `r03-s014-runtime-01` / 49971 | 0; 61,520 ms |
| `make agent-finalize` | `r03-s014-finalize-01` / 61778 | Terminal fail; outer exit unavailable after native session loss; graph 4,170 ms |
| `make generate` | `r03-s014-generate-02` / 28747 | 0; 11,340 ms |
| `make agent-finalize` | `r03-s014-finalize-02` / 23661 | 0; 26,360 ms; no generated rewrites |
| `make source-check CHECK=generated HARNESS_DIAGNOSTICS=basic` | `r03-s014-generated-01` / 23825 | 0; 17,402 ms |
| `make source-check CHECK=schema HARNESS_DIAGNOSTICS=basic` | `r03-s014-schema-01` / 15643 | 0; 6,677 ms |
| `make source-check CHECK=catalog HARNESS_DIAGNOSTICS=basic` | `r03-s014-source-catalog-01` / 58684 | 0; 1,731 ms |

Finalization-01 rejected a stale topology digest after the final planner edit.
Its ten cleanup steps completed; regeneration repaired the projection before
finalization-02. A source-catalog invocation accidentally reused the existing
`r03-s014-catalog-01` ID and was rejected with usage exit 2 before execution;
the original evidence was preserved. The fresh source-catalog root above passes.
Independent `make harness-observability-check RESULTS_DIR=<root>` passed for
catalog-01, runner-02 and runtime-01 with full validity. Failed runner-01 remains
retained under E-176. No unresolved candidate-owned resource is reported.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The final generated source is
`sha256:07f571ee9eec012940dbd98fa4ac0accbe4efe1d16e597afbbe9ee98bec75543`.
Its predecessor focused runtime source was
`sha256:98a2717d1947d2fe0455a1bd26b8e93114586508f07a4c4a73b52d5337c2737c`;
the difference is the regenerated topology digest, not changed runtime behavior.
S-015 entry requires this boundary's Markdown lint, immutable candidate
preservation, exact workload/non-CPU-claim comparison and a new frozen protocol.
No measured benefit or guardrail result is claimed.

### E-178 — WF-15 frozen comparison entry

WF-14 closing `make lint-markdown` passed at `r03-s014-close-lint-01`,
**90772/0**, 17,160 ms, before WF-15 began. The candidate preservation contains
7,509 files / 48,209,427 compressed bytes; archive SHA-256
`e040718aa00a6fd85f71eb056166b9485aa0de06d5f4cc93a2ece3ee87df5428`.
Both immutable sources reside under `.cartulary/remediation03/reservations/`:
S-013 baseline and S-014 candidate. Restoration verifies every preserved source
hash, adopts the destination NLSpec before its projections, and preserves this
tracker. The 17 differing paths contain only the Go budget treatment and its
necessary owner/contracts/projections/tests. No live invocation crosses cutover.

| Frozen workload | Rows / units | Changed CPU claims | Non-CPU reservations |
| --- | --- | --- | --- |
| `make test-fast` | 1,112 / 158 | 150 units: 4 → 1 | Identical per corresponding unit |
| `make browser-e2e SUITE=functional` | 383 / 137 | None; identical graph | Identical per corresponding unit |
| `make ci` | 2,279 / 600 | 201 units: 4 → 1 | Identical per corresponding unit |

All three baseline/candidate projections were exported through `make target-plan
TARGET=<target> INPUTS=<exp01-workload-inputs.json> DETAIL=graph JSON=1`, exit 0.
Membership, fixture capability, service dependencies, owner, estimate and cache
policy are equal. Baseline source restoration then passed `make generate`,
`r03-s015-prepare-b-01`, **63631/0**, 11,010 ms, and `make source-check
CHECK=catalog`, `r03-s015-source-catalog-b-01`, **35105/0**, 1,639 ms.

`performance-experiment-exp01-freeze.json` binds both executable sources
(E-175/E-177), exact populations/graphs, allowed differences, environment,
toolchain/storage identities, analysis hash and the ordered 25-attempt roster.
Captured host and admitted Go capacity are both 19 CPU tokens. Settings are
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off HARNESS_DIAGNOSTICS=basic
HARNESS_ACCOUNTING=required`. Native PTYs own timing commands. Execution-result
reuse is disabled; dependency/build caches stay warm. Public generation/catalog
preparation follows each drained source switch before required accounting.

One warm-up per measured treatment is excluded; CI has one candidate warm-up
and no measured baseline. Baseline fast/browser warm-ups precede candidate
fast/browser/CI warm-ups. Each subsequent odd pair is baseline-first and each
even pair candidate-first; within a treatment fast precedes browser, with CI
following candidate browser. Optional fast/CI pairs 3–4 follow Section 7.8.4;
browser always requires four matched pairs. CI compares elapsed only against
the assumed 3,600,000 ms. Decimal precision 28, half-even rounding, Python
MT19937 seed 20261021, 10,000 whole-run resamples and nearest-rank quantiles are
frozen. Measured guardrails use one-sided 99% upper bounds; full benefit uses a
one-sided 95% lower bound. Zero baseline admits only a zero candidate value.

Metrics cover the delegated Linux invocation and proven owned containers;
controller, unrelated/shared host work and Windows-host totals are excluded.
Every attempt requires independent evidence validation, successful cleanup and
secret scan, separately captured outer exit and source/graph binding. No
changed-source equivalent-comparison admission is used. Warm-ups and runtime
graph confirmation remain pending; no performance conclusion is established.

### E-179 — WF-15a common browser-readiness repair; experiment 01 closed

Freeze lint passed **83756/0**, 10,320 ms. Excluded baseline fast warm-up
`r03-s015-exp01-fast-warm-b` passed **86968/0**, 158/158 units, graph 216,822 ms,
owning-command 224,956.675213 ms, CPU 717,641,172 µs, kernel peak 4,642,394,112
bytes, p95 queue 90,894 ms and cleanup union 62 ms. All 167 cleanup steps and
the 2,007-file secret scan passed. Independent observability audit **4545/0**
reported full validity and complete diagnostics. It remains an excluded warm-up.

Excluded baseline browser warm-up `r03-s015-exp01-browser-warm-b` ended
**11034/2**, normalized product exit 10, graph 1,038,831 ms. Of 137 units, 135
passed; one cohort and its aggregate finalizer failed. The exact failing row
is `module.tasksdecisions.browser.task_lifecycle_patch_recovery`; two subsequent
same-file cases were unexecuted and remained explicitly accounted for.
All 277 cleanup steps completed. Diagnostic-only audit passed, full validity 0.

The retained Playwright trace establishes the test's missing precondition:
reopening the owner picker started a membership read at trace time 95,064.687 ms
that took 39.553 ms. Space removal began at 95,081.560 ms and completed before
the candidate list appeared at 95,106.627 ms. The prior snapshot showed
“Loading references”. The focus helper therefore selected its existing labelled
group fallback; a later background read correctly did not replay that intention.
The test assumed the loaded-candidate branch without awaiting its readiness.

WF-15a SHALL repair this shared assurance precondition in both treatments by
awaiting the enabled candidate control before removal. Existing exact focus,
keyboard, cancellation, payload and durable-state assertions SHALL remain.
No product focus behavior, deadline, retry, selector population or golden changes
are authorized by this repair. Existing Core 03 keyboard/retained-intent and
design §8.5 behavior remain authoritative. Validate the exact row alone, then regenerate/finalize and preserve both repaired
sources before a fresh freeze. The excluded functional warm-ups SHALL also
validate the original containing cohort before acceptance samples begin; the
public owner slice cannot select its mixed-owner population independently. V-026 correctness entry remains open meanwhile.

Experiment 01 is closed without acceptance samples. Every remaining planned
attempt is recorded as omitted after the correctness failure in
`performance-experiment-exp01-closed.json`; the successful fast warm-up cannot
be reused by a repaired-source experiment. S-013 lifecycle correctness and
S-014 focused assurance remain recorded at their actual sources. Performance
acceptance is inconclusive until the new experiment completes.

### E-180 — WF-15a completion and experiment 02 freeze

The shared repair adds `await expect(ownerCandidates).toBeEnabled()` before
the existing Space-removal interaction. It admits the intended loaded-list
branch without weakening its focus assertion, adding a sleep or changing
product behavior. All five existing task-lifecycle scenarios remain selected.

| Public command / exact selection | Run ID / native handle | Outcome |
| --- | --- | --- |
| `make format` | `r03-s015-readiness-format-01` / 28801 | 0; 8,234 ms |
| `make generate` baseline | `r03-s015-readiness-generate-b-01` / 29379 | 0; 11,200 ms |
| `make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.task_lifecycle_patch_recovery HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` | `r03-s015-readiness-alone-b-01` / 19236 | 0; 96,179 ms; 12/12 units |
| `make agent-finalize` baseline | `r03-s015-readiness-finalize-b-01` / 47592 | 0; 26,836 ms |
| `make generate` candidate | `r03-s015-readiness-generate-c-01` / 24366 | 0; 11,220 ms |
| `make agent-finalize` candidate | `r03-s015-readiness-finalize-c-01` / 43873 | 0; 27,059 ms |

All commands use `CGO_ENABLED=0`; the browser slice also disables result reuse.
Independent observability audit passed with full validity; all 23 cleanup steps
and the 97-file secret scan passed. The failed E-179 warm-up's accounting also
settled empty and its 2,319-file secret scan passed. The failed assertion remains
a correctness failure, not a cleanup failure or an accepted sample. An attempted
`target-plan TARGET=test-slice` without required owner/suite returned usage exit
2 without running tests; no unsupported mixed-owner selector was introduced.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Both repaired sources are preserved as `s013-source-exp02` and
`s014-source-exp02` under the reservation evidence directory. Baseline executable
identity is `sha256:f1efa9e051e901ad9841bdb5a776c7b74128c7dd8c00d47d987f4e399bb86aa8`;
candidate is `sha256:10fd1d1b5b7adee3a5712442f6f3fc76ad85d6a6a3177bf841015b39d14332e6`.
Their archives contain 7,509 files each, with SHA-256
`20d0650e619664eb9e7243f469442ad94cb741a9256222527fe70d7b65471451` and
`edc5cbcdee1bb017621e59a2c38035922f2edbacaad13ef500936590beb3f55a`.
The sentinel source is byte-identical across treatments. Public graph exports
and the new `performance-experiment-exp02-freeze.json` confirm E-178's exact
populations, non-CPU parity and the same 17 permitted differing paths.

WF-15a correctness repair and source migration are complete. Experiment 02
retains E-178's protocol, run ordering, limits and environment; all 25 attempt
IDs use `r03-s015-exp02-*`. No experiment-01 observation is reused. The excluded
functional warm-ups must pass the original containing cohort before sampling.
Boundary Markdown lint gates resumption. No performance benefit is established.

### E-181 — Experiment 02 baseline warm-ups validated

WF-15a closing lint passed at `r03-s015-readiness-close-lint-01`, **29174/0**,
10,020 ms. Baseline restoration passed `make generate` at
`r03-s015-exp02-prepare-b-01`, **49639/0**, 11,090 ms, and `make source-check
CHECK=catalog` at `r03-s015-exp02-source-catalog-b-01`, **3522/0**, 1,625 ms.

The following excluded warm-ups use E-178/E-180 settings and source B2.
Independent full observability audits and public performance/critical-path
exports passed for both. All accounting scopes settled empty; no diagnostic
omission or unresolved cleanup was reported.

| Run ID / command | Native exit | Units | Owning elapsed ms | CPU µs | Kernel peak bytes | p95 queue / cleanup union ms |
| --- | --- | --- | --- | --- | --- | --- |
| `r03-s015-exp02-fast-warm-b` / `make test-fast` | 52705/0 | 158/158 | 222,800.251439 | 573,893,851 | 4,034,801,664 | 77,623 / 62 |
| `r03-s015-exp02-browser-warm-b` / `make browser-e2e SUITE=functional` | 92648/0 | 137/137 | 1,163,296.527143 | 3,988,741,135 | 8,156,327,936 | 1 / 47,846 |

Fast cleanup completed 167 steps; browser cleanup completed 276. Browser's
2,296-file secret scan passed. The original nine-row cohort, including the
repaired task-lifecycle row and both previously unexecuted rows, passed fully.
Its 80-unit critical path reports graph queue 25 ms, host admission 2,409 ms,
fixture acquisition 27,356 ms, runner 1,090,942 ms and release 25,196 ms.
Phase sums are descriptive and can overlap. Four observed stack allocations
all retired for terminal borrower release; no reset-only generation occurred.
These are current workload observations, not a controlled comparison with
historical CI or a lifecycle speedup claim. Both runs remain excluded warm-ups.

Runtime graph identities match the frozen public projections. Candidate
fast/browser/CI warm-ups and every acceptance sample remain pending.

### E-182 — Experiment 02 candidate focused warm-ups validated

Candidate restoration passed `make generate`, `r03-s015-exp02-prepare-c-01`,
**82900/0**, 11,150 ms, and `make source-check CHECK=catalog`,
`r03-s015-exp02-source-catalog-c-01`, **51856/0**, 1,640 ms. The following
excluded warm-ups use the frozen C2 identity and required accounting settings.

| Run ID / command | Native exit | Units | Owning elapsed ms | CPU µs | Kernel peak bytes | p95 queue / cleanup union ms |
| --- | --- | --- | --- | --- | --- | --- |
| `r03-s015-exp02-fast-warm-c` / `make test-fast` | 48792/0 | 158/158 | 227,869.721680 | 653,489,737 | 5,014,781,952 | 20,498 / 61 |
| `r03-s015-exp02-browser-warm-c` / `make browser-e2e SUITE=functional` | 44368/0 | 137/137 | 1,156,567.039805 | 4,015,310,877 | 8,102,240,256 | 1 / 47,562 |

Both independent full observability audits passed; all public exports succeeded.
Runtime graphs match their frozen projections. Fast/browser cleanup completed
167/276 steps; secret scans passed for 2,007/2,296 files; accounting settled
empty. Complete allocation/admission coverage reports no omissions. Browser
again observed four stack generations, all terminal-borrower retirement, zero
reset-only generations, and a passing original containing cohort. Its path host
admission is 2,279 ms and graph queue is 17 ms. These observations are descriptive
and excluded from acceptance. Candidate CI warm-up and all acceptance samples
remain pending; no benefit or non-regression decision is established.

### E-183 — Candidate CI warm-up resource observation

At the user's request, the excluded `r03-s015-exp02-ci-warm-c` warm-up was
observed while native session **7809** remained running. This is observation
activity, not acceptance sampling or authorization to change capacities. Source
C2 and graph `sha256:26d07527f3291358b6364e2704091a0510266d43437117a8493ebea13612f3b4`
remain frozen. No executable-source, service, lane or scheduling change was made.

Public `make test-run-status RESULTS_DIR=.cartulary/test-results/r03-s015-exp02-ci-warm-c
JSON=1` returned exit 0. Revision 327, graph elapsed 972,706 ms, reported
121 passed, 15 running and 464 pending units, with no failure observed. The
complete running list contained two executing and thirteen waiting units.
Mapping those exact IDs to the frozen public target-plan export gives:

| Graph reservations | Executing units | Waiting units | Combined / graph capacity |
| --- | --- | --- | --- |
| CPU | 4 | 15 | 19 / 19 |
| PostgreSQL | 4 | 12 | 16 / 16 |
| Object store | 2 | 6 | 8 / 8 |

These are graph reservations, not granted host claims or utilization. The live
reader truncates the pending-wait list and does not yet establish the native
admission cause. The retained decision view must distinguish actual deficits,
queue-priority/exclusive blockers and logical-slot waits before attribution.

Read-only `vmstat 1 6`, `free -m`, `/proc/pressure/{cpu,memory,io}` and
`docker stats --no-stream --format '{{json .}}'` snapshots completed successfully.
The five interval samples, excluding vmstat's since-boot first line, showed
85–89% Linux CPU idle, 0–1% I/O wait, no swap-out, approximately 22.3 GiB available
memory, and zero ten-second CPU/memory pressure. These host observations have
different scope from required invocation accounting and do not establish
whole-run guardrails. The supplied Windows screenshot was read through its
WSL path; its host graphs are likewise not invocation attribution.

Evidence is retained under `.cartulary/remediation03/reservations/` as
`ci-warm-observation-01.json`, `ci-warm-observer-activity.json`,
`ci-warm-vmstat-01.txt` and `ci-warm-containers-01.jsonl`. The initial interactive
resource snapshot precedes the saved snapshots; their intervals must not be
represented as simultaneous or continuous monitoring. Observation remains
bounded and SHALL stop during any declared quiet measurement interval.

The supported finding is idle hardware alongside fully occupied graph
reservations. A future coordination/fairness investigation is plausible, not an
adopted change: prove whether fitting independent work is obstructed by pending
host requests, and preserve starvation bounds, exclusive measurement fences,
physical ownership and cleanup entitlement in any proposal. Current S-015
continues unchanged; non-CPU quantities retain their NO_CHANGE disposition.
No current CI correctness, cleanup or performance conclusion is available yet.

A second bounded observation, revision 415 at graph elapsed 1,247,348 ms,
reported 147 passed, 17 running and 436 pending, with no observed failure. Its
running list omitted one unit, so no complete reservation total is inferred.
Seven visible Go groups were executing, three preparing, one finalizing and five
waiting. Subsequent five-second Linux samples showed 57–68% idle and 8–20% I/O
wait with no swap traffic. This is evidence of phase-dependent contention, not
proof that every delay has the same cause. Device-specific latency was not
collected (`iostat` is unavailable; no prerequisite was installed). Saved public
status and resource snapshots use the corresponding `-02` filenames. Terminal
fixture and admission evidence is still required to attribute the I/O activity.

Revision 589 at graph elapsed 1,512,372 ms reported 324 passed, 17 running
and 259 pending units, with no failure observed. Its running list omitted one
unit; the visible list again contained many host-admission waiters. This does
not establish a constant run-wide utilization level. The native command remains
running; no timing sample or terminal pass is inferred from live progress.
REQ-353's existing 25% CPU safety margin explains the 19-token capacity, but
cannot alone explain the earlier 85–89% Linux CPU-idle intervals.

At revision 845, graph elapsed 2,127,091 ms, CI reported 442 passed, one failed,
seven running and 150 pending units. All seven running IDs were visible: one
browser runner and six host-admission waiters. Two waiters selected
`harness.browser.integration.native_nested_cancellation` and
`harness.browser.integration.native_graph_observations`; each retains its authored
four-CPU/twenty-process reservation. These larger profiles are unchanged by the
ordinary Go-budget treatment. A subsequent five-second sample showed 88–92%
CPU idle and 0–2% I/O wait; the corresponding saved resource files use `-03`.
This identifies candidates for reservation-lifetime analysis, not permission to
lower their peak entitlements or change current capacities.

### E-184 — Experiment 02 CI warm-up correctness failure; repair pending drain

Public live revision 752 at graph elapsed 1,801,663 ms reported a failed
`operation:harness.lint_go_staticcheck`, 429 passed units and nine running units.
The native CI session **7809** remained active; its outer exit and final cleanup
are not yet known. The published unit result reports exit **1**,
`product/test_assertion_failure`. Its bounded published stdout contains one
Staticcheck diagnostic: `internal/testutil/harnessfixtures/consumers_linux_test.go:37:3`,
**SA4006**, an assigned `key` value that is never read. Stderr is empty.
This is a test-source defect introduced in S-013, common to both treatments,
not evidence of a product failure or a Go-budget performance result.

WF-15b SHALL remove the redundant assignment in both source treatments after the
current invocation drains. It SHALL preserve the exercised consumer-key behavior,
run the canonical static-analysis route and affected fixture assurance, finalize
and regenerate projections as required, and freeze a new experiment before
sampling. No deadline, isolation, capacity, fairness, golden or product behavior
change is required by this diagnostic. Experiment 02 cannot provide accepted CI
performance evidence; its failed attempt and unused scheduled attempts SHALL
remain recorded. The native owner SHALL retain cleanup authority while draining.
No executable-source edit, cancellation, retry or experiment restart occurred in this receipt.

At revision 1182, elapsed 3,093,567 ms, 506 units had passed and the same single
Staticcheck failure remained. Five browser units were preparing, executing or
waiting; 88 units remained pending. Optional process sampling reported stale
coverage (last sample age 207,307 ms, 2,852 omitted observations). That sample
MUST NOT be presented as current host or invocation utilization. Required
cgroup accounting and terminal canonical timing have separate coverage contracts
and still await validation.
An independent subsequent `vmstat` sample (`-04`, native **56530/0**) showed
72–77% CPU idle and 1–2% I/O wait; it is not a replacement for the omitted
invocation samples. Revision 1282 at elapsed 3,321,666 ms reported 534 passed,
one failed, three running and 62 pending units. All three running units were
executing browser cohorts; no host-admission waiter appeared in that complete
running list. This later phase has a different parallelism constraint from the
earlier fully reserved graph with waiting units. Structured observations `-05` through `-09` are
preserved alongside the earlier bounded snapshots.

### E-185 — CI warm-up closed; scoped utilization diagnosis

**Correctness: FAIL. Cleanup/evidence integrity: PASS. Performance: excluded.**
Native CI **7809/2** ended with normalized exit 10 at frozen C2 source. All
600 units settled: 599 passed and only `operation:harness.lint_go_staticcheck`
failed. Graph duration was **3,683,426 ms**; owning-command duration was
**3,757,370.700701 ms** (62m37s). All **1,160 cleanup steps** completed, required
invocation accounting settled empty, and the retained-secret scan passed across
**9,738 files**. No candidate-owned resource remains unresolved from this run.

Public `make explain-run RESULTS_DIR=.cartulary/test-results/r03-s015-exp02-ci-warm-c`
passed. Independent `make harness-observability-check` checks `diagnostics`,
`events` and `timing` passed, native **73156/0**, **30670/0**, **17989/0**, each
with `full_validity=0`; they do not turn the failed test run into a pass.
Performance and critical-path exports passed. Changed-source comparisons were
not submitted as equivalent. Experiment 02 is closed in
`performance-experiment-exp02-closed.json`; all twenty unexecuted acceptance
attempts are marked omitted. Zero acceptance samples ran. An initial internal
receipt-collation attempt selected a nonexistent JSON field and failed before
writing; the corrected collation used the retained run manifest.

| Observation | Exact value and scope |
| --- | --- |
| Charged CPU | 12,531,059,629 µs; about 3.335 mean cores over owning-command duration; invocation cgroup and proven containers only |
| Kernel peak memory | 14,584,365,056 bytes, about 13.58 GiB; charged peak, not simultaneous Windows-host usage |
| Graph ready-to-start p95 | 1,581,411 ms across the declared unit population |
| Cleanup union | 163,790 ms; all required cleanup succeeded |
| Complete critical path | 80 units; 634,413 ms graph queue plus 3,046,171 ms unit lifetime |
| Path host admission | 1,665,758 ms (27m46s), distinct from graph queue |
| Path runner / acquisition / release phase sums | 1,078,163 / 272,932 / 25,093 ms; these are canonical phase observations |
| Path queue reasons | Resources 553,634 ms; exclusive barrier 80,463 ms; rank 316 ms |
| Whole-run application stacks | 82 generations; 69 capacity-pressure and 13 terminal-borrower retirements; 31 reset-only generations |
| Path-associated stacks | 30 generations; 29 capacity-pressure retirements; 15 reset-only generations |

The phase view reports zero phase overlap on this path and seven milliseconds
outside unit lifetime from cache-lookup placement; phase sums are not additional
wall time to add to queue-plus-lifetime. Other targets share units and are not
additive. The earlier E-166 counts have different source/workload conditions and
MUST NOT be treated as a controlled before/after comparison. This warm-up does
not establish either speedup or regression acceptance.

**Coverage limitation.** Canonical timing and all allocation lifecycle records
are complete. Optional execution/resource diagnostics are truncated. The separate
host-admission section retained 3,317 records, exhausted its 16 MiB byte budget,
and disclosed **1,725,059 omitted observations**. Its last retained observation
ends at graph elapsed 752,318.990611 ms, approximately 12m32s. Consequently,
1,507,923.953157 ms of path host-admission time lacks the detailed phase breakdown;
uncovered time is explicit. The complete path's host-admission total remains
valid, but its later causes cannot be reconstructed from the missing samples.

The retained native decisions contain 86,929 denied observations, including
40,143 with current process deficits, 15,054 with current object-store deficits,
186 with current PostgreSQL deficits, and zero with current CPU deficits.
75,836 show priority blockers; 41,235 have no current capacity deficit but still
have priority blockers. These categories overlap, represent recorded decision
counts rather than durations, and cover only the retained early portion.
They support investigation of process reservations and queue priority, not a
whole-run causal percentage or permission to reduce resource claims.

**Recommendations, not adopted behavioral amendments.** First improve bounded
admission retention so compact per-unit phase/denial summaries survive detailed
record saturation without exceeding existing limits. Then audit process-slot
claims against exact nested/fixture lifetimes and distinguish logical reservation
pressure from measured processes. If complete evidence confirms avoidable
head-of-line blocking, design coordinated graph/host admission with explicit
starvation bounds, exclusive fences and cleanup entitlement before changing
fairness. Separately investigate remaining browser-chain imbalance. The present
continuation's non-CPU sizing, fairness and lane NO_CHANGE boundaries remain.
Current mandatory retirement corrections remain; observed turnover does not
justify restoring an incorrect counterfactual or terminal-consumer model.

The authored un-nested footprint illustrates this cross-resource constraint:
PostgreSQL, application stack and renderer declare 8 + 4 + 2 resident process
slots; an active browser unit declares four additional process slots and two CPU
tokens. Four such lanes plus one shared two-process object store account for
74 nominal process slots against the captured capacity of 76, with only eight
browser CPU tokens. This is an illustrative profile sum, not a second admission
algorithm or a reconstruction of a particular live state; native weighted
accounting, shared identities and actual resident sets remain authoritative.
A future process-capacity/footprint investigation MUST validate real peak nested
work and cleanup headroom before proposing numeric changes. Current values
remain unchanged under F-024.

WF-15b has removed only the overwritten test assignment after the invocation
settled. Candidate `make agent-finalize`, `r03-s015-lintrepair-finalize-c-01`,
passed **84330/0**, 27,595 ms. Candidate `make lint CHECK=go`,
`r03-s015-lintrepair-lint-c-01`, passed **3633/0**, 8,310 ms. Affected fixture
assurance, baseline repair, source preservation and fresh freeze remain pending.
`RESULTS_DIR` was unset for finalization; retained-run maintenance was skipped
because no compatible successful full warm-check root exists.

### E-186 — WF-15b complete; experiment 03 frozen

**Corrective assurance: PASS. Performance: PENDING.** The common fixture test now
computes its compatibility key once, after restoring the intended service list.
All mutation-alias, service-order, compatibility, terminal-disposition and cleanup
assertions remain. No runtime/owner semantics, capacity, profile, fairness,
deadline, selector or golden changed. F-021/F-022 → R03-036/037 → consumer-key
regression → V-024 fixture assurance and Go lint supplies the corrective trace.

Both treatments ran `make agent-finalize` before broader verification with
`CGO_ENABLED=0`. Lint and fixture checks disabled execution-result reuse; fixture
checks used basic diagnostics and required accounting. Exact roots are
`.cartulary/test-results/r03-s015-lintrepair-<label>-<b|c>-01`.

| Label / public command | Baseline native/exit; graph ms | Candidate native/exit; graph ms |
| --- | --- | --- |
| `finalize` / `make agent-finalize` | 72191/0; 26,401 | 84330/0; 27,595 |
| `lint` / `make lint CHECK=go` | 8902/0; 8,087 | 3633/0; 8,310 |
| `fixtures` / `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required` | 40039/0; 14,589 | 64120/0; 14,695 |
| `generated` / `make source-check CHECK=generated` | 47870/0; 17,064 | 28444/0; 17,531 |
| `schema` / `make source-check CHECK=schema` | 45819/0; 6,713 | 29403/0; 6,732 |
| `catalog` / `make source-check CHECK=catalog` | 54297/0; 1,770 | 90328/0; 1,796 |

Baseline restoration also ran `make generate`,
`r03-s015-lintrepair-generate-b-01`, **33209/0**, 11,220 ms. Both fixture invocations
passed independent full observability audits, exit 0 / `full_validity=1`, with
complete required coverage and empty accounting scopes. Every listed graph
completed cleanup and passed its retained-secret scan. No resource remains
unresolved. `RESULTS_DIR` was unset; maintenance was skipped because no compatible
successful full warm-check root exists.

| Treatment | Executable source digest | Immutable archive SHA-256 |
| --- | --- | --- |
| B3 | `e79809d7980b80ef65d4bbb3b4b898577ff4a2889c776a1c9b970558e18684a4` | `0d14dd26b2d0f77aaf80e9e80637ec9900df504dffe31d318e596e56bfcedf37` |
| C3 | `92fda6e5b496b2bfa44e41d5034503c00b5e2d4721c2b8812b31ff1134010c19` | `b5cdcb36edede142afb836c9353a5f0034ed0d24bfc2fd906720a7921350b515` |

Archives `s013-source-exp03.*` / `s014-source-exp03.*` each preserve 7,509 files
at unchanged HEAD `f7b917026e1174661a9bdf9118d18120bbd225b0`. Excluding this tracker,
the sole change from each experiment 02 treatment is the common
`internal/testutil/harnessfixtures/consumers_linux_test.go` repair. Both treatments
retain identical repaired consumer tests and the E-180 browser-readiness fix.
Their difference remains the same seventeen Go-budget owner/code/test/projection
paths. Fresh public target plans preserve exact rows, non-CPU claims and graph
identities: fast 1,112 rows / 158 units, browser 383 / 137, CI 2,279 / 600.

`performance-experiment-exp03-freeze.json` freezes the same protocol, settings,
source treatment map, environment and twenty-five-attempt roster as E-178/E-180,
with prefix **r03-s015-exp03**. Five excluded warm-ups precede acceptance sampling;
CI still has no measured baseline and uses assumed 3,600,000 ms. Four browser
pairs and the existing first-two/full-protocol benefit rules remain unchanged.
Every failed/omitted earlier attempt remains preserved. New timings SHALL NOT
reuse the failed CI warm-up or reinterpret historical evidence.

E-185's bounded-detail limitation remains explicit: optional host-admission and
process observations can exhaust their budgets; they cannot establish complete
causal/resource attribution. Required canonical timing, allocation settlement,
invocation accounting, cleanup and scans retain their independent acceptance
requirements. No detail-limit increase or unproved whole-run attribution is
included in this test-only repair.

WF-15b is complete; its source migration is preserved and reversible. The next
entry condition is this tracker's successful boundary `make lint-markdown`, then
the first excluded baseline fast warm-up. S-015 performance decisions and S-016
final validation/handoff remain open.

### E-187 — Experiment 03 warm-ups entered

WF-15b boundary Markdown lint passed at `r03-s015-lintrepair-close-lint-01`,
**44353/0**, 9,920 ms. Unaccounted controller preparation with `CGO_ENABLED=0`,
`make source-check CHECK=catalog`, `r03-s015-exp03-prepare-b-01`, passed
**81623/0**, 1,657 ms, before required-accounting measurements resumed.

The first excluded warm-up, `r03-s015-exp03-fast-warm-b`, used frozen B3 and
`CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off`, public `make test-fast
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Native **36472/0** passed
158/158 units in 211,806 ms graph execution. Owning elapsed was
219,974.211495 ms; charged CPU 569,903,861 µs; peak memory 4,158,492,672 bytes;
p95 queue 79,026 ms; cleanup union 62 ms. All 167 cleanup steps completed,
accounting settled empty and the secret scan passed. Independent full audit
**80864/0** returned `full_validity=1`; public performance and critical-path
exports passed with complete canonical, allocation and host-admission coverage.
The runtime graph matches the frozen baseline fast projection. This is an
excluded warm-up, not a performance acceptance sample. Remaining warm-ups and
all acceptance attempts remain pending.

### E-188 — Experiment 03 baseline warm-ups complete

The excluded B3 browser warm-up `r03-s015-exp03-browser-warm-b` passed native
**2085/0**, 137/137 units, graph duration 1,147,116 ms. Owning elapsed was
1,159,666.996709 ms; charged CPU 3,973,787,516 µs; kernel peak memory
8,124,256,256 bytes; p95 graph-ready delay 1 ms; cleanup union 48,131 ms.
All 276 cleanup steps completed; invocation accounting settled empty and the
retained-secret scan passed. Independent public full audit exited 0 with
`full_validity=1`; public performance and complete critical-path unit exports
exited 0. Canonical, allocation and host-admission coverage are complete.
Four browser-stack generations settled. No performance acceptance claim follows.

One read-only status checkpoint exited 0 during this excluded warm-up:

```sh
make test-run-status RESULTS_DIR=.cartulary/test-results/r03-s015-exp03-browser-warm-b JSON=1
```

Its preserved `exp03-browser-warm-b-status-01.json` observed 129 passed,
one executing and seven dependency-waiting units, without an observed failure.
This observer activity is recorded in the experiment ledger and is not an
acceptance-sample observation. Both B3 warm-ups and their audits are complete;
the next entry is drained switch to frozen C3, public generation and unaccounted
catalog preparation, then the frozen candidate warm-up sequence. Acceptance
samples and the optional Go-budget decision remain pending.

### E-189 — Experiment 03 candidate fast warm-up passed

After the drained B3 → C3 switch, `CGO_ENABLED=0 make generate` passed at
`r03-s015-exp03-generate-c-01`, **41900/0**, 11,100 ms; unaccounted
`CGO_ENABLED=0 make source-check CHECK=catalog` passed at
`r03-s015-exp03-prepare-c-01`, **42514/0**, 1,571 ms. Tracker changes remain
outside executable identity; C3 is unchanged from E-186's frozen source.

The excluded `r03-s015-exp03-fast-warm-c` used the frozen environment and public
`make test-fast HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`.
Native **3774/0** passed 158/158 units, graph duration 216,383 ms. Owning elapsed
was 224,969.991507 ms; charged CPU 636,506,706 µs; kernel peak memory
4,861,554,688 bytes; p95 graph-ready delay 19,843 ms; cleanup union 67 ms.
All 167 cleanup steps completed, invocation accounting settled empty and the
secret scan passed. Independent full audit **67749/0** returned
`full_validity=1`; public performance **90395/0** and complete critical-path
units **58370/0** exports passed. Canonical, allocation and host-admission
coverage are complete. No allocation generation was required by this workload.

The candidate browser warm-up `r03-s015-exp03-browser-warm-c` is active under
native handle **84755**; its result is pending. CI candidate warm-up and all
acceptance attempts remain pending. Warm-up differences are descriptive only:
the lower graph-ready percentile is not evidence of end-to-end benefit, and no
performance acceptance decision is established.

### E-190 — Experiment 03 focused warm-ups complete; CI warm-up entry

The excluded C3 browser warm-up `r03-s015-exp03-browser-warm-c` passed native
**84755/0**, 137/137 units, graph duration 1,142,353 ms. It used the frozen
source/environment and public `make browser-e2e SUITE=functional
HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. Owning elapsed was
1,154,965.367339 ms; charged CPU 3,973,233,238 µs; kernel peak memory
8,288,608,256 bytes; p95 graph-ready delay 1 ms; cleanup union 48,323 ms.
All 276 cleanup steps completed; invocation accounting settled empty and the
secret scan passed. Independent public full audit and both public performance /
complete critical-path unit exports exited 0; audit returned `full_validity=1`.
Canonical, allocation and host-admission coverage are complete; all four
browser-stack generations settled. No speedup or non-regression conclusion is
established by excluded warm-ups.

Both treatments' fast and browser warm-ups are now independently validated.
The next frozen attempt is candidate-only excluded CI warm-up
`r03-s015-exp03-ci-warm-c` under the same four environment settings and public
`make ci HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required`. No CI baseline
execution is authorized or planned; the acceptance reference remains the user's
assumed 3,600,000 ms. All acceptance samples remain pending. Known bounded-detail
limitations and the complete canonical/accounting requirements remain as E-186.

**In-progress observation — not an acceptance sample.** Native CI handle **9013**
remains active. A public status checkpoint at graph elapsed 488,438 ms, revision
183, reported 85 passed, 503 pending and 12 running units, with zero observed
failures and no omitted running records. Of those 12 admitted units, two were
executing and ten waiting. Summing their declared claims from the identical
frozen graph gives CPU 19, PostgreSQL 16 and object-store 8, exhausting those
graph capacities. Waiting units held CPU 15, PostgreSQL 12 and object-store 6;
executing units held CPU 4, PostgreSQL 4 and object-store 2. These are graph
reservations, not native host grants or measured consumption. This instantaneous
association cannot assign exclusive causal durations to blockers.

After active publication was confirmed, one bounded `vmstat 1 6` sample
(**88392/0**, first since-boot row excluded) observed 81–87% Linux CPU idle,
0–1% I/O wait and no interval swap-in/out. A separate `free -m` snapshot reported
23,469 MiB available; CPU/memory pressure avg10 was zero. `docker stats
--no-stream --format '{{json .}}'` (**23014/0**) showed the two active browser
containers at about 100.54% and 118.00% CPU; this is the Docker percentage scale,
not whole-host saturation. Host and container samples are adjacent, not
synchronized or invocation-attributed. They do not include Windows host totals.

All commands were read-only and exited 0. Evidence is retained in
`exp03-ci-warm-status-01.json`, `exp03-ci-warm-vmstat-01.txt`,
`exp03-ci-warm-containers-01.jsonl`, `exp03-ci-warm-memory-01.txt`,
`exp03-ci-warm-pressure-01.txt` and `exp03-ci-warm-observer-activity.json` under
the continuation evidence directory. Observation overhead is excluded with this
warm-up. This reinforces E-185's future admission-coordination investigation;
it does not authorize fairness/capacity changes or establish a numeric sizing
reduction. No service control or executable-source mutation occurred.

A second excluded warm-up checkpoint at 1,686,477 ms, revision 655, reported
331 passed, 250 pending and 19 running units with no observed failure. The
bounded running list omitted three units; the reported entries included one
executing Go unit, two preparing Go units and thirteen waiting units. No complete
activity distribution or reservation total is claimed for this snapshot.
`vmstat 1 6` (**98966/0**, first row excluded) observed 47–75% CPU idle,
1–5% I/O wait and no interval swap-in/out. Available memory was 23,902 MiB;
CPU pressure avg10 0.35%, memory pressure avg10 zero. Docker stats **65682/0**
included one unavailable row, which is not interpreted as zero utilization.
The `exp03-ci-warm-*-02` files and observer ledger preserve these adjacent,
scoped observations; they are not a whole-run average or acceptance sample.

A third excluded checkpoint at 2,839,771 ms, revision 1089, reported 483 passed,
111 pending and six running units with no observed failure. The complete running
list contained one executing accessibility cohort, one preparing stateful reset,
and four waiting units (two webserver cohorts, a webserver reset and the scheduler
contract row). The optional live resource sampler disclosed 2,118 omissions and
two discovery truncations; its partial RSS is not a whole-run memory guarantee.
The subsequent bounded `vmstat 1 6` (**51600/0**, first row excluded) observed
90–92% CPU idle, zero interval I/O wait and no interval swap-in/out. Available
memory was 26,009 MiB; CPU/memory pressure avg10 was zero. Docker stats
**89071/0** and the `exp03-ci-warm-*-03` files preserve the separate host/daemon
scope. This late-phase snapshot supports investigating browser critical-path and
admission delays; it does not establish their continuous duration or a speedup.

### E-191 — Experiment 03 failed and settled; WF-15c and reduced CI policy

`r03-s015-exp03-ci-warm-c` completed native **9013/2**, normalized product exit
**10**, graph **3,730,897 ms**, owning-command elapsed **3,804,872.133992 ms**.
C3 remains `sha256:92fda6e5b496b2bfa44e41d5034503c00b5e2d4721c2b8812b31ff1134010c19`.
Of 600 units, 598 passed; one accessibility cohort failed and its target summary
reported that failure. All **1,164 cleanup steps completed**, accounting settled
empty and the secret scan passed **9,800 files**. There are no unresolved owned
resources. Diagnostic/events/timing audits passed **27121/0**, **69615/0** and
**76623/0**, with `full_validity=0` as required for failed assurance. Performance
and complete critical-path units exports passed **68332/0** and **86148/0**.
These reader successes do not make the invocation pass.

The failed exact row is
`web.design.accessibility.membership_management_accessibility`; public
`make task-guide OWNER=web.design JSON=1` confirmed its route. The retained
Playwright report classifies `route.fulfill: Fetch response has been disposed`
as `product/test_assertion_failure`. The referenced trace shows the last focus
assertion complete at 6466.205 ms, context teardown starting at 6466.999 ms,
upstream fetch completing at 6470.092 ms and fulfillment failing at 6473.856 ms.
Pinned Playwright 1.63.0 disposes the context request client at the start of
context close, before its channel-close log. This supports a fixture teardown
ordering defect, not a failed UI assertion or a demonstrated 60-second timeout.
A secondary teardown timeout message remains retained without that causal claim.
The sanitized call/timing/error excerpt is
`exp03-ci-route-failure-trace-excerpt.json` in the continuation evidence directory.
No request bodies, cookies or browser storage were exported.

WF-15c SHALL adopt REQ-806/AC-088 teardown ordering, then repair the shared
owned-context fixture to unregister and drain page/context route handlers before
request-client disposal. It SHALL preserve cleanup failures and existing
deadlines, fresh contexts and session revocation order. It SHALL NOT suppress
route errors, add sleeps/retries or widen timeouts. Deterministic support tests,
the exact failed accessibility row, relevant shared-fixture assurance and
projection/type checks SHALL precede a new source freeze. The repair is common
to baseline and candidate; optional Go tuning is unchanged. Exit requires those
focused checks, both source snapshots, traceability and Markdown lint.

The user explicitly selected **one final CI run**, with timing descriptive
against the assumed **3,600,000 ms**, and **no formal CI performance acceptance**.
Section 7.8.4 now orders cheap correctness checks, test-fast benefit screening,
conditional browser non-regression, then final retained-source CI. No additional
CI warm-up is planned. The mandatory correctness and cleanup gates remain.
Historical CI experiments and E-167 completion retain their original meaning.
Experiment 03's failed attempt and twenty unexecuted acceptance attempts are
preserved in `performance-experiment-exp03-closed.json`; zero acceptance samples
were collected. Source repairs require a new experiment and cannot erase this
failure. WF-15 and WF-16 remain open; performance is unestablished.

### E-192 — WF-15c implementation and focused repair evidence

REQ-806/AC-088 was amended before implementation. The shared fixture now calls
`closeOwnedBrowserContext` in teardown. The helper begins all page/context HTTP
route drains with `behavior: wait`, waits for settlement, then closes the context;
it preserves single failures or aggregates multiple drain/close failures. Four
routed deterministic tests cover pending fetched responses, concurrent drain
initiation, synchronous drain failure, close failure and error preservation.
The existing support row owns those exact titles. No test timeout, renderer
worker, resource capacity, retry, reset boundary or product contract changed.

Authoring initially appended unsorted catalog titles: format-01 failed plan
admission and task-guide exited 2. Sorting the exact selector list repaired that
configuration failure before execution. `r03-s015-route-repair-format-02`
passed **97785/0**, 8,053 ms; generate-01 passed **89254/0**, 11,930 ms.
The public browser task-guide then passed **62929/0**.

All roots below are `.cartulary/test-results/r03-s015-route-repair-<label>`.
Commands used `CGO_ENABLED=0`; focused execution disabled result reuse and used
basic diagnostics. Source was
`sha256:829d5e67807a907ff6bb290e21f357e0a18a9b469bcb7e376126dc85eb2c98ac`.

| Label | Public command | Native/exit; result |
| --- | --- | --- |
| support-01 | make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.cleanup_suite_9036a615b1 HARNESS_DIAGNOSTICS=basic | 7115/0; 2/2 units; 1,527 ms; 11 cleanup steps completed; scan pass, 26 files |
| a11y-01 | make test-slice OWNER=web.design ROWS=web.design.accessibility.membership_management_accessibility HARNESS_DIAGNOSTICS=basic | 2175/0; 12/12 units; 100,782 ms; 23 cleanup steps completed; scan pass, 97 files |
| finalize-c-01 | make agent-finalize | 40366/0; 27,636 ms; maintenance skipped because RESULTS_DIR was unset |
| types-c-01 | make lint CHECK=types | 23351/2; 20,778 ms; TS2532 in new support test at mock.results[0]; all 11 cleanup steps settled |

Both focused invocations passed independent full observability audits, exit 0.
The types failure concerns unchecked optional mock-array access, not the runtime
repair; optional access now follows the already asserted exact invocation count.
The failed check remains retained. Finalization, focused support and types checks
will be repeated on that corrected source before preservation. No additional CI
or performance attempt has run, and WF-15c remains open.

### E-193 — WF-15c assurance complete; experiment 04 frozen

**Implementation and focused assurance: PASS. Performance: PENDING.** REQ-806/
AC-088 → owned-context cleanup helper/shared fixture → exact support row and
membership accessibility row supplies the repair trace. The support row retains
its original eight selectors and adds four lifecycle assertions. E-192's failed
sort/type attempts remain retained; no CI rerun was used to repair them.

All commands below use `CGO_ENABLED=0`; verification disables execution-result
reuse. Roots are `.cartulary/test-results/r03-s015-route-repair-<label>`.

| Label / public command | Native/exit; graph ms |
| --- | --- |
| finalize-c-02 / make agent-finalize | 57716/0; 26,706 |
| types-c-02 / make lint CHECK=types | 93519/0; 21,940 |
| frontend-c-01 / make lint CHECK=frontend | 77824/0; 2,262 |
| go-c-01 / make lint CHECK=go | 25953/0; 9,045 |
| support-02 / make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.cleanup_suite_9036a615b1 HARNESS_DIAGNOSTICS=basic | 60910/0; 1,831; independent full audit exit 0 |
| generated-c-01 / make source-check CHECK=generated | 32148/0; 18,411 |
| schema-c-01 / make source-check CHECK=schema | 48563/0; 7,224 |
| catalog-c-01 / make source-check CHECK=catalog | 67564/0; 2,285 |
| generate-b-01 / make generate | 41869/0; 11,250 |
| finalize-b-01 / make agent-finalize | 87614/0; 27,868 |
| go-b-01 / make lint CHECK=go | 98201/0; 9,715 |
| generated-b-01 / make source-check CHECK=generated | 56784/0; 19,013 |
| schema-b-01 / make source-check CHECK=schema | 36986/0; 8,365 |
| catalog-b-01 / make source-check CHECK=catalog | 56096/0; 2,054 |

Every listed graph passed its final secret scan and completed all cleanup steps;
`route-repair-validation-receipts.json` retains counts and exact source identities.
No owned resource remains unresolved. Maintenance was skipped because RESULTS_DIR
was unset; no compatible successful full warm-check root exists. Candidate type,
frontend and support assurance applies to byte-identical common browser sources
in B4. E-192's real accessibility pass uses identical fixture/cleanup production
code; only the support test's nullable mock access changed afterward. It remains
an earlier-source focused receipt, not a fabricated B4/C4 full-run pass.

| Treatment | Executable source digest | Archive SHA-256 |
| --- | --- | --- |
| B4, mandatory diagnostics/retention and common repairs | `6b58080ae0d6f13b75aa1fc9b10cf17a28643739af299a91caf4bef3770369a7` | `ba191f629b6747ee41b482d19c3c27eb5155373171068ee010ecc6f87f567502` |
| C4, same plus optional Go budgets | `02d796280f47c02ed0734f009457f76414483734dddc045baf5ffe0ac8dc239e` | `6ad031db4bc40a8518e9729280a48e1ccdd875b6717e29c5008b5c9791a0323d` |

Both archives contain 7,509 files at unchanged HEAD `f7b917026`. Their only
non-tracker differences remain the approved seventeen Go-treatment paths.
The NLSpec was restored before authored Go projections; generated outputs were
regenerated through Make. No compatibility fallback or unrelated edit was added.
The immutable freeze is `performance-experiment-exp04-freeze.json`.
Public target-plan graph exports for fast/browser/CI passed on both sources:
C4 **70369/0**, **75361/0**, **80313/0**; B4 **96995/0**, **58305/0**, **30264/0**.
These are read-only plans, not CI executions. Row populations and all non-CPU
claims match. Host/toolchain/storage/Docker identities were revalidated in
`exp04-environment-revalidation.json`; authored non-CPU capacities remain unchanged.

The new freeze has twenty conditional fast/browser attempts and **zero CI timing
attempts**. First run one excluded fast warm-up per treatment, then matched fast
pairs (B→C on odd pairs, C→B on even pairs). Apply the first-two-pair rejection/
early-acceptance rules; complete four only when required. Browser warm-ups and
four non-regression pairs occur only if fast passes. Rejection omits dependent
performance attempts and restores optional Go owner/code/projections together.
WF-16 then performs retained-source final assurance and **one final CI**; its
elapsed comparison with assumed 60 minutes is descriptive, without a confidence
bound or formal CI performance acceptance. Earlier freezes are immutable.

Required accounting/basic diagnostics/cache-off/CGO-disabled settings, independent
run audits, all five measured guardrails, matched run ordering and decimal
bootstrap rules are unchanged. `analyze-exp04.py` preserves measured-workload
arithmetic and removes CI acceptance analysis. No intentional competing workload
or polling is allowed during acceptance measurements. WF-15c closing Markdown lint `r03-s015-route-repair-lint-01` passed
**99940/0**, 18,640 ms. This completion receipt is checked by
`r03-s015-route-repair-lint-02` before the next workstream begins. WF-15c is
complete; optional Go tuning remains unaccepted pending the frozen fast gate.

### E-194 — WF-15 Go performance rejection and coordinated restoration

**S-014 tuning: REJECTED. S-015 decision: COMPLETE. Final assurance: PENDING.**
Experiment 04's two matched fast pairs show elapsed improvements of
**-2.5023392755%** and **-3.8813706578%**. Both are strictly below 5%, triggering
the agreed early rejection. No four-pair confidence bounds or guardrail acceptance
are claimed. Baseline B4 is now the retained implementation. Its behavioral
owner, compiler/runner, contracts and generated projections were restored
together; there is no dual reader or tuning toggle. Mandatory diagnostic and
retention corrections, and the common browser teardown repair, remain.

All six executed attempts used CGO_ENABLED=0, cache off, basic diagnostics and
required accounting. Roots are `.cartulary/test-results/<run-id>`. Each native
exit and independent full audit was 0; every run passed **158/158 units**, all
**167 cleanup steps**, its final secret scan and empty accounting settlement.
Canonical, host-admission and allocation coverage are complete. No allocation
generation was needed. Public performance/critical-path exports all exited 0.

| Run ID suffix after r03-s015-exp04- | Command | Native/exit | Owning elapsed ms | Charged CPU µs | Kernel peak bytes | p95 queue / cleanup union ms |
| --- | --- | --- | --- | --- | --- | --- |
| fast-warm-b | make test-fast | 78927/0 | 225688.405975 | 577550525 | 5040623616 | 79062 / 67 |
| fast-warm-c | make test-fast | 83323/0 | 223597.494320 | 619934232 | 5038977024 | 27137 / 46 |
| fast-p01-b | make test-fast | 93431/0 | 224340.061199 | 578020659 | 4196003840 | 78225 / 64 |
| fast-p01-c | make test-fast | 43032/0 | 229953.810661 | 634400228 | 5116846080 | 22699 / 63 |
| fast-p02-c | make test-fast | 24587/0 | 228062.297381 | 640622145 | 5056049152 | 25635 / 69 |
| fast-p02-b | make test-fast | 59554/0 | 219541.093785 | 569938143 | 4320157696 | 75047 / 47 |

Warm-ups are excluded; pair one is B→C and pair two C→B. The audited public
exports and raw arithmetic are retained in `performance-decision.json`;
`performance-disposition-exp04.json` records the final conditional-workload
outcomes. Browser non-regression is **NOT ESTABLISHED / OMITTED** because its fast
prerequisite failed. Four remaining fast attempts and ten browser attempts
(including two browser warm-ups) are recorded as omitted, fourteen in total.
CI formal performance acceptance is **NOT REQUESTED**, per E-191; one final
correctness invocation remains required in WF-16, with descriptive timing only.
The immutable freeze and closed ledger remain separate; failed experiments 01–03
are unchanged. No changed-source run used equivalent comparison admission.

**Diagnostic observation, not another acceptance metric.** Every fast sample's
critical path ends in the same frontend Vitest cohort. Baseline path queue is
0/1 ms; candidate queue is 8,911/3,159 ms, mainly resource waiting for that cohort
(8,900/3,136 ms). Cohort lifetimes remain approximately 211–216 seconds. Path host
admission is only 6/6 ms baseline and 13/40 ms candidate. Thus the lower overall
ready-delay percentile did not shorten the long frontend path; it also concealed
increased waiting on that path. This supports future critical-path-aware admission
and frontend cohort-balance investigation, not a capacity increase or a claim
that Go execution alone caused every elapsed difference. CI's distinct host-wait
and browser-chain opportunity remains E-185/E-190; no CI speedup is established.

Between drained source switches, public CGO-disabled generation/catalog preparation
passed. Roots use `r03-s015-exp04-<label>`:
generate-c-01 **68179/0**, 11,030 ms; prepare-c-01 **13574/0**, 1,689 ms;
generate-b-01 **83180/0**, 10,940 ms; prepare-b-01 **47686/0**, 1,634 ms;
generate-c-02 **63230/0**, 11,300 ms; prepare-c-02 **58284/0**, 1,709 ms;
generate-b-02 **79151/0**, 11,260 ms; prepare-b-02 **8225/0**, 1,681 ms.
WF-15c final receipt lint-02 passed **57374/0**, 10,300 ms before these runs.
No intentional competing workload or resource polling occurred during samples.
The bounded route-gate source inspection occurred only after a sample drained;
`route-gate-audit.json` states its limited scope and found no additional justified
source repair. It does not replace final browser assurance.

F-021 → R03-038/040 → frozen seventeen-path treatment → V-025 correctness and
V-026 matched measurements supplies the decision trace. Retained B4 source is
`sha256:6b58080ae0d6f13b75aa1fc9b10cf17a28643739af299a91caf4bef3770369a7`.
`exp04-retained-source-verification.json` confirms all 7,508 non-tracker snapshot
files match; candidate C4 remains archived for historical investigation. There
are no unresolved candidate-owned resources. Risks remaining are unchanged
selection-dependent Go budgets, incomplete long-CI detailed diagnostic coverage
and unoptimized graph/host admission; these are not represented as fixed tuning.
WF-16 entry requires this tracker update and successful boundary Markdown lint.

### E-195 — WF-16 retained-source validation entry

WF-15 boundary lint `r03-s015-decision-lint-01` passed **60593/0**, 10,170 ms,
after E-194's rejection/restoration update and before this workstream. Retained
source remains B4; no further Go tuning or optional performance sampling is
admitted. All four principal public task-guides were refreshed successfully:
browser **92369/0**, command surface **25646/0**, catalog **60270/0** and evidence
accounting **92097/0**. `final-owner-selections.json` records the exact rows.

WF-16 SHALL run CGO-disabled `make agent-finalize` first, with RESULTS_DIR unset
because no compatible successful full warm-check root exists. Broader retained-
source validation then consists of generated/schema/catalog checks; native
fixture lifecycle/observations; runtime/scheduler/public-surface contracts;
catalog routing; and current-epoch evidence/admission contracts through their
owner routes. It SHALL also run `make browser-e2e SUITE=support`,
`make browser-e2e SUITE=visual`, and `make harness-contract SUITE=ui-review`.
No golden refresh or capacity change is authorized.

Only after those prerequisites pass SHALL the single planned `make ci` start,
with CGO_ENABLED=0, cache off, basic diagnostics and required accounting. Every
execution SHALL settle, pass its retained-secret scan and receive applicable
independent evidence validation. CI elapsed time is descriptive against the
assumed 60 minutes; no CI baseline, warm-up or performance acceptance experiment
will run. Physical settlement, a resumable handoff and closing tracker lint
remain mandatory. This entry receipt makes no final-assurance pass claim.

### E-196 — WF-16 pre-CI assurance passed; single CI entry

All checks below ran on retained B4 source
`sha256:6b58080ae0d6f13b75aa1fc9b10cf17a28643739af299a91caf4bef3770369a7`.
Finalization preserved every non-tracker snapshot file. All runs used CGO_ENABLED=0;
verification disabled execution-result reuse. Owner and explicit suite runs also
used basic diagnostics and required accounting. Roots are
`.cartulary/test-results/r03-s016-<label>-01`.

| Label / public route | Native/exit | Graph ms | Passed units | Completed cleanup | Secret-scan files |
| --- | --- | --- | --- | --- | --- |
| finalize / make agent-finalize | 47653/0 | 27113 | 1 | 10 | 12 |
| generated / make source-check CHECK=generated | 62971/0 | 17306 | 5 | 14 | 29 |
| schema / make source-check CHECK=schema | 79506/0 | 6693 | 3 | 12 | 19 |
| catalog / make source-check CHECK=catalog | 30180/0 | 1764 | 2 | 11 | 15 |
| command / make test-slice OWNER=harness.command_surface | 33685/0 | 177550 | 5 | 14 | 247 |
| catalog-owner / make test-slice OWNER=harness.test_catalog | 85520/0 | 54201 | 2 | 11 | 21 |
| evidence / make test-slice OWNER=harness.evidence_accounting | 95044/0 | 122323 | 3 | 12 | 26 |
| browser-owner / make test-slice OWNER=harness.browser | 23653/0 | 137383 | 3 | 12 | 264 |
| support / make browser-e2e SUITE=support | 66698/0 | 124478 | 14 | 27 | 142 |
| visual / make browser-e2e SUITE=visual | 7687/0 | 621283 | 37 | 74 | 545 |
| ui-review / make harness-contract SUITE=ui-review | 11085/0 | 110495 | 10 | 19 | 61 |

Owner commands supplied these exact comma-separated ROWS values; the table above
does not claim unrestricted owner coverage:

| Owner | Exact selected rows |
| --- | --- |
| harness.command_surface | harness.command_surface.unit.native_runtime, harness.command_surface.unit.scheduler_smoke, harness.command_surface.behavior.public_registry_parity |
| harness.test_catalog | harness.test_catalog.behavior.routing_contract |
| harness.evidence_accounting | harness.evidence_accounting.behavior.current_epoch_evidence, harness.evidence_accounting.behavior.observation_admission |
| harness.browser | harness.browser.unit.native_fixtures, harness.browser.integration.native_fixture_lifecycle, harness.browser.integration.native_graph_observations |

All seven owner/explicit runs passed independent full observability audits and
public performance exports, exit 0, with accounting settlement **empty** and child
exit 0. Every listed graph completed cleanup and passed its secret scan. Exact
identities/counts are in `s016-pre-ci-validation-receipts.json`; no owned resource
remains unresolved. No golden, capacity, isolation or retry changed. Maintenance
was skipped because RESULTS_DIR was unset, not replaced with a CI or partial root.

**Assurance: pre-CI PASS. Implementation: mandatory corrections retained; Go tuning
restored. Performance: E-194 rejection unchanged.** This covers R03-035–041 through
their admitted diagnostic, retirement, consumer, schema/routing and explicit
browser owners. It does not yet establish final containing CI assurance.

The next and sole planned CI command is:

```sh
CGO_ENABLED=0 CARTULARY_HARNESS_CACHE_MODE=off \
CARTULARY_TEST_RUN_ID=r03-s016-ci-01 \
make ci HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required
```

`r03-s016-pre-ci-lint-01` checks this tracker before CI entry. There is no CI
baseline or warm-up. Elapsed comparison against assumed 3,600,000 ms remains
descriptive. Bounded resource observations may support the user's utilization
question outside known quiet measurement intervals; their scope and timing SHALL
be recorded and SHALL NOT be presented as continuous invocation utilization.
The final CI native exit, audit, scan, physical settlement, handoff and closing
Markdown lint remain open. No rerun is authorized by observation alone.

### E-197 — Single final CI active; identity bound

Pre-CI tracker lint `r03-s016-pre-ci-lint-01` passed **47286/0**, 9,970 ms.
The E-196 exact command started at **2026-10-23 20:02:14 UTC** under native
handle **64360**, root `.cartulary/test-results/r03-s016-ci-01`. Its published
source matches retained B4 exactly:
`sha256:6b58080ae0d6f13b75aa1fc9b10cf17a28643739af299a91caf4bef3770369a7`.
The actual runtime graph is
`sha256:839598de81121cba3d6e8e24bac7478ca9df65b414f102d7b122ce6255e22e08`.
This actual runtime identity is not relabeled as the earlier static target-plan
graph or an equivalent performance comparison; CI is final correctness with
descriptive timing only.

The initial exact-root public status observation exited 0 and reported active
execution: **28 passed, 554 pending, 18 running, zero observed failures**, out of
600 units. It does not establish final outcome or physical settlement. The native
session remains authoritative for the outer command exit. The observation and
its scope are retained in `s016-ci-status-00.json` and
`s016-ci-observer-activity.json`. No CI baseline, warm-up or second CI was launched.
Final outcome, independent audit, retained scan, settlement, handoff and closing
tracker lint remain pending.

At graph elapsed **609,475 ms**, a bounded public status checkpoint reported
**67 passed, 522 pending and 11 running units**, no observed failures and no
omitted running entries. Two webserver-backed cohorts were executing; nine
admitted units were waiting. Adjacent host samples were read-only:
`vmstat 1 6` **83582/0**, excluding its since-boot first row, recorded **80–85% CPU
idle**, **0–2% I/O wait** and zero interval swap-in/out. `free -m` reported
**23,626 MiB available**. CPU/memory PSI avg10 was zero. Docker stats **38826/0**
showed two browser containers at **115.93%/131.85% CPU**, where 100% is one core.
The daemon-wide/container and Linux host views are adjacent, not synchronized or
invocation-attributed; they do not include Windows host totals. These observations
support admission/dependency investigation, not numeric memory/process cuts or a
whole-run utilization percentage.

Read-only target-plan probes **50254/0** and **48458/0** retained the static graph,
even with CGO_ENABLED=0 and the exact published declared inputs. Source inspection
identified different compiler contexts: `buildTargetPlan` uses constructor
Go/PostgreSQL lane defaults **4/4**; runtime preparation captures the published
**19/16** capacities. No live reservation totals were inferred from the
nonmatching static plan. Public runtime evidence remains authoritative, and no
changed-source or changed-context equivalent comparison is claimed. The scope,
commands and native exits are in `s016-ci-observer-activity.json`; the
`s016-ci-*-01` samples and `s016-ci-graph-reservations-01.json` retain the limits.
No source or capacity mutation occurred.

At graph elapsed **1,719,659 ms**, the second bounded checkpoint reported
**230 passed, 360 pending, nine running and one failed unit**, with active
publication in draining phase. The observed failure is classified
**product / test_assertion_failure**, in webserver-backed cohort
`4c02b087bc125a226729ee92bff00096eb1539e403ba6f3bb8ad5d49682b3837`.
This is not a terminal result; the native invocation continues independent work
and settlement. Final correctness is not established and the exact failure
requires terminal diagnosis before any repair or repeated verification.

Adjacent `vmstat 1 6` **61037/0** recorded **72–86% CPU idle**, **1–5% I/O wait**
and no interval swapping; `free -m` reported **24,757 MiB available**. Docker
stats **50207/0** was a later daemon-wide snapshot and is not synchronized with
the listed running units. CPU/memory/I/O PSI avg10 some values were
**0.31/0.01/3.17%**. Scope remains bounded Linux-host context; these samples do
not prove reservation overstatement or identify the browser assertion's cause.
The `s016-ci-*-02` files and observer ledger preserve the exact observations.

A bounded read of this already completed cohort's public unit/group results and
referenced Playwright report identifies row
`module.workbook.browser.destination_handoff_768_8f85fa718d`: **40 passed, one
failed and six dependency-skipped rows**. Its `Work` heading focus assertion
failed because the heading was absent; this was not a fixture-startup or cleanup
classification. The referenced trace shows the second Task Requests grid-visibility
assertion completing at 586,630.177 ms, followed by Work key activation starting
at 586,632.235 ms. The test's first such navigation also waits for dock closure
and destination focus; the second omits both. Core 03 REQ-03-314 distinguishes
accepted/mounted destinations from completed presentation. This establishes a
test sequencing gap, not direct instrumentation of the internal focus race.
`s016-ci-destination-handoff-diagnosis.json` preserves the bounded evidence and
proposed common serial-handoff assertion helper. No executable source was changed,
no test was rerun, and whole-invocation terminal diagnosis remains required.

### E-198 — User cancellation settled; focused handoff repair registered

The user explicitly instructed: **end the CI test run early and apply the fixes**.
At **2026-10-23 20:38:30 UTC**, Ctrl-C was delivered once to verified native
handle **64360**. It exited **1** at **20:39:33 UTC**; the owning child reported
exit **130 / signal 2**. CI status is **cancelled**, with **255 passed, one failed
and 344 cancelled units** of 600. Its primary classification remains
**product / test_assertion_failure** from the E-197 workbook row. No failure is
erased by cancellation. Graph time was **2,179,526 ms**; owning-command accounting
time was **2,239,176.837118 ms**. Interrupted timing is not a completed CI timing
or an improvement against the assumed baseline.

All **617 cleanup steps completed**, accounting settlement is **empty**, and
the final retained-secret scan passed **6,461 files**. Independent public
`make harness-observability-check RESULTS_DIR=.cartulary/test-results/r03-s016-ci-01
CHECK=diagnostics` passed **8377/0**, explicitly **full_validity=0**. The public
performance reader passed **53811/0**. Diagnostic resource coverage remains
partial with 552 omissions; this is not whole-run sampled utilization assurance.
The exact-root cancellation status reader exited 0. These outputs are retained
as `s016-ci-cancel-audit.json`, `s016-ci-performance.json` and
`s016-ci-status-cancel.json`. No candidate-owned resource remains unresolved.

WF-16a/S-016a now addresses the E-197 serial handoff test gap. The authoritative
Core 03 REQ-03-314/315 already distinguish visible/accepted destinations from
completed presentation; implementation SHALL consolidate the two Task Requests
transitions in `apps/web/e2e/workbook-destination-handoff.spec.ts` around those
existing completion assertions. This changes tests and tracker documentation,
not product requirements or runtime behavior. No timeout, retry, sleep, golden,
resource-profile or selector change is admitted. The causal limit remains:
the trace proves the incomplete test synchronization, not the internal focus
event ordering.

`make task-guide OWNER=module.workbook JSON=1` passed **64160/0** and confirms
the exact routes retained in `s016-repair-workbook-guide.json`. V-027a selects
the failed 768 row alone, then six destination-handoff rows (1440, 1024, 768,
hidden, supersession and viewer variants as enumerated by the guide) and the
three other dependency-skipped rows (startup, reference queries, filter focus).
The exact nine-row combined selection SHALL be checked before execution. Existing earlier source
and experiment receipts remain historical, including the rejected Go treatment.

Assurance: **CI INCOMPLETE / CANCELLED**. Performance: **no CI claim**.
Implementation: mandatory diagnostics/retention and B4 common fixes retained;
Go tuning remains restored. This child repair does not close WF-16, and no
automatic full CI restart is part of the requested focused repair. Boundary
`r03-s016a-entry-lint-01` SHALL pass before the source edit.

### E-199 — WF-16a focused repair passed; resumable validation handoff

Entry lint `r03-s016a-entry-lint-01` passed **37705/0**, **9,010 ms**, before
editing executable source. `openTaskRequestsFromWork` now owns the repeated
serial navigation in `apps/web/e2e/workbook-destination-handoff.spec.ts`. Both
transitions require the destination grid to be visible, Work to be detached,
and focus to be inside that grid before the next action. This removes duplicated
partial synchronization and preserves every prior assertion. The deliberate
supersession scenario remains independent and unchanged. No product code,
requirement, test title, routing selector, timeout, retry or golden changed.

Traceability is **F-025 → Core 03 REQ-03-314/315 → local serial-handoff helper →
V-027a**. The existing owner requires presentation completion independently of
accepted/mounted state; this is a test projection correction and requires no
new behavioral amendment. The retained trace does not directly instrument the
internal focus race. A passing focused run does not erase the interrupted CI
failure or establish that every possible timing defect is repaired.

The repaired executable source is
`sha256:a81b80320b89cd193ab69312e22da1849114c433ca03ae64452424b3f321ef68`
at unchanged HEAD `f7b917026e1174661a9bdf9118d18120bbd225b0`. Comparing all
7,509 tracked/nonignored files against preserved B4 finds only the test file and
this tracker changed. `s016a-source-delta.json` records the comparison, and
`s016a-repair.patch` preserves the test delta against the B4 file. The accepted
mandatory harness corrections and restored Go budgets remain byte-identical.
No schema migration or runtime rollback is involved; earlier treatment archives
and failed attempts are preserved.

All roots below are `.cartulary/test-results/r03-s016a-<label>`. All commands use
CGO_ENABLED=0; checks disable execution-result reuse. Both browser selections
use basic diagnostics and required accounting.

| Label / public command | Native/exit | Graph ms | Passed units | Completed cleanup | Scan files |
| --- | --- | --- | --- | --- | --- |
| handoff-alone-01 / make test-slice OWNER=module.workbook | 11547/0 | 92677 | 12 | 23 | 97 |
| handoff-combined-01 / make test-slice OWNER=module.workbook | 59747/0 | 100989 | 12 | 23 | 105 |
| finalize-01 / make agent-finalize | 58429/0 | 26691 | 1 | 10 | 12 |
| generated-01 / make source-check CHECK=generated | 98257/0 | 17442 | 5 | 14 | 29 |
| catalog-01 / make source-check CHECK=catalog | 78171/0 | 1742 | 2 | 11 | 15 |
| frontend-02 / make lint CHECK=frontend | 11896/0 | 2068 | 2 | 11 | 18 |
| types-02 / make lint CHECK=types | 46402/0 | 21215 | 2 | 11 | 15 |

The alone command's ROWS is
`module.workbook.browser.destination_handoff_768_8f85fa718d`. The combined
command uses the following nine exact rows as one comma-separated ROWS argument;
all nine terminal row results passed:

```text
module.workbook.browser.browser_startup_honors_explicit_sheet_home_defau_ee3b02ee01
module.workbook.browser.destination_handoff_1024_963c4de116
module.workbook.browser.destination_handoff_1440_920aaf64d5
module.workbook.browser.destination_handoff_768_8f85fa718d
module.workbook.browser.destination_handoff_hidden_a9564ecda1
module.workbook.browser.destination_handoff_supersession_62c776e360
module.workbook.browser.destination_handoff_viewer_64fc8d19e2
module.workbook.browser.existing_task_reference_queries_stage_typed_time_f92c8c967a
module.workbook.browser.filter_editor_keyboard_focus
```

`make harness-observability-check RESULTS_DIR=<each browser root>` passed exit
0 with **full_validity=1** for both runs. Their public performance exports also
exited 0 and establish empty accounting settlement and child exit 0, not a
performance acceptance claim. Every listed run completed all cleanup and passed
its final retained-secret scan. `s016a-validation-receipts.json` records exact
source/graph identities and counts; `s016a-selections.json` freezes the rows.
No candidate-owned resources remain unresolved.

Two mistaken `make source-check CHECK=types` / `CHECK=frontend` invocations
were rejected before execution with **config / usage_error, exit 2**; their
intended `r03-s016a-types-01` / `frontend-01` roots were not created. Public
`make help ALL=1` confirmed those selectors belong to `make lint`; the corrected
fresh runs above passed. These command-selection errors are recorded separately
from product failures and are not successful checks.

**Implementation: repair accepted. Focused assurance: PASS. Performance: no
new claim; E-194 Go rejection unchanged. Full CI: INCOMPLETE after user
cancellation.** Agent-finalize preceded the broader source checks and preserved
the repaired source. Retained-run maintenance was skipped because RESULTS_DIR
was unset; no compatible successful full warm-check root exists.

Closing `r03-s016a-completion-lint-01` SHALL validate this completion receipt.
WF-16a then closes; WF-16 remains open for complete containing CI assurance and
its final evidence/settlement/handoff boundary. There is no active invocation
and no automatic CI restart in this repair. When final validation resumes,
revalidate the retained source, use a fresh explicit run ID with the unchanged
E-196 CI environment, and retain the failed/cancelled root. Do not relabel B4
CI or the interrupted duration as a repaired-source pass or an elapsed benefit.

### E-200 — WF-16a completion lint and final repair handoff

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s016a-completion-lint-01 make
lint-markdown` passed **48628/0**, **9,870 ms**, after E-199's workstream update.
WF-16a/V-027a is complete: source repair, focused assertions, independent audits,
finalization, types/frontend/generated/catalog checks, cleanup and secret scans
passed. The final authored receipt is checked by
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=r03-s016a-closing-lint-01 make lint-markdown`;
that closing command MUST pass before delivery.

WF-16 and the reservation continuation remain **IN_PROGRESS**, because the user
cancelled the only final CI attempt. No additional CI was launched, no complete
CI timing or performance acceptance is claimed, and no resource remains owned
by the stopped invocation or focused repair. Next work is the outstanding
containing-validation boundary described in E-199, not another tuning experiment.
The current repair is a test-only delta; prior accepted harness behavior, rejected
Go tuning and all historical receipts remain intact. Maintenance was skipped
because RESULTS_DIR was unset.

## 11. Open Questions and Blockers

| ID | Blocker / why it matters | Owner and required evidence | Affected work / status |
| --- | --- | --- | --- |
| RB-001 | Execution baseline and authorization | RESOLVED by explicit implementation request and E-018/E-019; executable pilot identity recorded under E-021 | Later candidate sources still bind their own exact identities |
| RB-002 | Cross-owner owner adoption and partition-consumer migration | RESOLVED under E-031/E-039/E-049: adopted REQ-806/AC-088, manifest v15, shared tuple and coordinated consumer cutover | S-001 G-03/G-04 accepted; E-071/E-072 final explicit/integration validation passed |
| RB-003 | Source-effect and runtime independence proof | E-032/E-039 completes the S-001 source dossier; E-033/E-034 records alone/order/concurrency and failure fixtures. E-036/E-037 pilot failures are diagnosed and focused repairs pass | E-049 accepts S-001; E-058 retains stronger isolation with no eligible conversion; E-071/E-072 final assurance passed |
| RB-004 | Protocol and measurement capability fixed | Section 8.2/E-019/E-048/E-053 fix the amended protocol; S-008/E-030 supplies charged-resource capability | E-049 accepts S-001 under its applicable performance routes; later slices require their own comparisons |
| RB-005 | Conditional replacement of REQ-804 exclusive lifetime | E-059 completes investigation with dedicated-lifetime NO_CHANGE; material >10% opportunity not established | No replacement owner/contract or reusable generation admitted; a future proposal needs fresh proof and adoption |
| RB-006 | Validation routes and exact-symbol row allocation | RESOLVED by E-019/E-050/E-055: exact-symbol row allocated, survivor/siblings/durable/combined selections and catalog routing pass | Correctness-only final union preserves all five original symbols after optimization rejection |
| RB-007 | Canonical Markdown lint coverage | RESOLVED by S-006/E-020: exact inclusion and successful public lint | E-073 completion lint and final update command |
| RB-008 | Broader semantic redundancy and transitive effects were not exhaustively audited | Future candidate owner: inspect exact source/owner contracts and complete preservation map | Further S-002 replay consolidation DROPPED from this effort; any future deletion requires a new proposal |
| RB-009 | Required measurement capability | RESOLVED by user-authorized S-008 and E-030; R03-018/V-009 accepted for the declared Linux/cgroup/container/build profile | WF-08 DONE. Later experiments must freeze inputs and collect their own samples; earlier pilots remain excluded |
| RB-010 | Complete diagnostic and baseline measurement coverage | S-009; V-010 pilots and accounting audits | RESOLVED, E-077/E-079/E-081; current baseline warm-up remains required before acceptance sampling |
| RB-011 | Lifecycle correctness and independent workload acceptance | S-010/S-011; V-011/V-012 | RESOLVED by E-160/E-161 rejection and coordinated restoration; no lifecycle optimization retained; E-166/E-167 final validation and handoff complete |

No unresolved primary-owner contradiction was established. Remediation 02's
summary/appendix/active-receipt distinctions are handled in Section 1.4.
Unproved historical causes MUST NOT be represented as repaired or newly proven
current failures; active acceptance-scope decisions remain owned by that
continuation.

## 12. Binary Completion Criteria

### 12.1 Independent outcomes

| Outcome | Current statement |
| --- | --- |
| Execution authorization | Granted by explicit implementation request; scope and conditional gates above |
| Planning contract | Approved; E-048 sets four pairs and permits first-two-pair >=20% early exit; E-053 requires rejection when both first-two-pair improvements are <5%; serial sequence unchanged |
| Original execution through E-073 | COMPLETE: all activated workstreams including WF-07a–d repairs and S-007 final handoff; E-071/E-072 final verification and E-073 completion lint |
| Implementation accepted | S-006 lint coverage, S-008/S-008a invocation accounting, and S-001 browser v15 under E-049; S-002 Location assertions and independent routing under E-055; fixture-removal candidate rejected under E-054 |
| Assurance acceptance | V-009 under E-030/E-038, S-001 under E-049, S-002 under E-055; E-070–E-072 current-source repairs, projections, accounting, support, visual, UI-review and 600/600 CI units pass with settled cleanup |
| Performance acceptance | E-049: S-001 support four-pair validation passes all five guardrails; mean elapsed improves 17.81%. Functional improvements 34.75%/36.10% qualify for user-authorized early exit without a full statistical guardrail claim |
| Canonical validation | E-070 agent-finalize precedes E-071 projections/explicit suites and E-072 final CI/audit; E-073 records completion lint and the final update command; retained-run maintenance skipped because RESULTS_DIR was unset |
| Preservation | Pre-existing tracker edits incorporated; no unrelated source edits made |

Continuation status: COMPLETE — E-166/E-167. S-009 diagnostics remain accepted.
S-010 lifecycle assurance passed, but experiment 10 rejects its performance and
restores its owner, implementation and projections together. Both workload
decisions are complete. Section 7.7 retains publication unchanged with improved
diagnostics and 192/192 qualified lifetimes. Common corrections remain; candidate-only
preparation corrections remain historical with their rejected treatment.
E-165 validates the retained source and explicit suites; E-166 supplies 600/600
final CI, independent audit, physical settlement and final scan acceptance.
E-167 records completion lint and the mandatory closing lint after the final
tracker update. All failed and omitted attempts remain preserved. Historical host
checkpoint latency and browser callback/socket order are not claimed fully
reconstructed; the original endpoint failure remains causally uncertain.
Original completion below remains separate from this continuation's acceptance.

The checklist below records original implementation completion. E-015–E-017
remain historical authoring receipts and do not establish runtime acceptance.

### 12.2 Implementation completion checklist

| Criterion | Pass condition | Present evidence / disposition |
| --- | --- | --- |
| Controlling artifact | Preserve this tracker, its twelve sections, stable IDs and historical receipts | MET; pre-existing edits retained; no competing plan |
| Authority and projections | Adopt behavioral owner before implementation; refresh generated outputs through Make | MET; E-030/E-031/E-038/E-061/E-070 adoption; E-070 finalization and E-071 generated/schema/catalog passes |
| Browser compatibility | Exact selection, state/failure/settlement proof and applicable performance acceptance | MET; E-033/E-034/E-039/E-049; final explicit/CI checks E-071/E-072 |
| Incident assurance | Exact create/replay Location and independent routing; reject unaccepted simplification | MET; E-054 rejects removal; E-055 correctness accepted and all original fixtures retained |
| Conditional investigations | Accepted implementation or supported NO_CHANGE, with unresolved proof visible | MET; E-057/E-058/E-059; no tuning, isolation conversion or lifetime pool claimed |
| Measurement capability | Declared coverage, fenced ownership, fail-closed evidence and nested cancellation | MET within declared profile; E-030/E-038; E-071 accounting and E-072 CI pass |
| Current final verification | Finalize before broader checks; CI plus explicit-only support/visual/UI-review checks pass | MET; E-070–E-072, native exits and independent audits recorded |
| Physical settlement and retained evidence | No unresolved owned resource; terminal scans pass; historical failures remain | MET for final runs, E-070–E-072; earlier failures/cancellation and source identities preserved |
| Performance honesty | Apply four-pair and first-two-pair rules without promoting partial diagnostics | MET; E-048/E-049/E-053/E-054; no whole-run performance claim for final CI |
| Maintenance prerequisite | Supply RESULTS_DIR only for compatible successful full warm check root | SKIPPED because RESULTS_DIR unset; E-070/E-072 explain unavailable prerequisite |
| Final tracker validation | Final tracker update followed by successful public Markdown lint | E-073 completion lint passes; its final receipt text is checked by the registered closing lint invocation |
| Resumability | Accepted/no-change decisions, current source, exact roots, migration boundaries and remaining risks recorded | MET; E-070–E-072 and Sections 9/11/12 |

A slice may be DONE as an accepted implementation only after G-03 and its
applicable G-04 rule pass for its final source and workload. Structural acceptance
does not claim measured improvement. S-006 documentation validation uses lint
and coverage, not G-04. An investigation may instead close with
an explicit no-change, DEFERRED or DROPPED disposition; that is not an accepted
optimization and MUST record the disposition of any candidate source changes.
Future implementation closure also requires the affected regression/projection
evidence and physical settlement for every activated slice. A documentation completion,
a reduced modeled cohort count, a passing reader, or a historical successful run
MUST NOT satisfy those implementation gates.

### 12.3 Continuation completion checklist

| Criterion | Completion evidence |
| --- | --- |
| Behavioral authority and projections | Accepted diagnostic/stability amendments precede their projections; rejected lifecycle amendments restored; E-165 generated/schema/catalog checks pass |
| Every activated workstream disposition | Section 9 tasks closed; lifecycle rejected, transport unchanged, diagnostics/common corrections accepted; E-160/E-161/E-167 |
| Required final assurance | E-165 explicit and owner checks; E-166 600/600 CI with independent full-validity audit |
| Cleanup and retained integrity | E-166 zero unresolved cleanup, empty accounting, all terminal scans passed; optional diagnostic limits explicit |
| Performance honesty | E-160/E-161 first-two rejection; measured browser pairs and assumed CI reference remain distinct; no post-hoc baseline substitution |
| Resumable source and handoff | E-166 final archive/digest and five-path correction map; E-167 future entry boundary and unresolved historical uncertainty |
| Final tracker validation | E-167 completion lint passed; registered closing lint MUST pass after this final update |
| Maintenance prerequisite | Skipped because RESULTS_DIR was unset; no successful full warm-check root was substituted |

### Reservation continuation completion boundary

Status: IN_PROGRESS, E-199. WF-00f and S-012/013 are complete. S-014 Go tuning
is rejected and restored after S-015 measured rejection; common repairs remain.
The sole final CI attempt was user-cancelled after one assertion failure.
WF-16a has repaired the narrow test projection and passed V-027a; complete S-016
CI assurance remains unestablished. No test invocation is active. Mandatory
correctness, optional Go tuning and measured performance have separate decisions.
Completion requires Section 7.8.5 and V-027, followed by closing Markdown lint.
