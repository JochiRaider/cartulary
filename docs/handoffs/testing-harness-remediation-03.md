# Testing Harness Remediation 03 — Coordinated Execution Efficiency

This is the controlling execution tracker for the user-approved Remediation 03
plan. The implementation request supersedes the earlier planning-only action
boundary; historical receipts E-001–E-017 retain their original scope.

Document date: 2026-10-21, America/New_York. Execution status: **COMPLETE**.
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
| Proposed requirement | R03-001 through R03-018 in Section 7 | MUST/MUST NOT describe conditions of a later authorized change. A proposal cannot override an adopted requirement before adoption. |
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

E-018 is the current execution baseline: HEAD
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

No replacement scheduler, broker, resident test daemon, compatibility mode or
parallel agent execution is part of this effort. Additional conversions under
S-004 receive separate child workstream IDs and serial boundary receipts.

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
| Current execution | COMPLETE: all activated workstreams including WF-07a–d repairs and S-007 final handoff; E-071/E-072 final verification and E-073 completion lint |
| Implementation accepted | S-006 lint coverage, S-008/S-008a invocation accounting, and S-001 browser v15 under E-049; S-002 Location assertions and independent routing under E-055; fixture-removal candidate rejected under E-054 |
| Assurance acceptance | V-009 under E-030/E-038, S-001 under E-049, S-002 under E-055; E-070–E-072 current-source repairs, projections, accounting, support, visual, UI-review and 600/600 CI units pass with settled cleanup |
| Performance acceptance | E-049: S-001 support four-pair validation passes all five guardrails; mean elapsed improves 17.81%. Functional improvements 34.75%/36.10% qualify for user-authorized early exit without a full statistical guardrail claim |
| Canonical validation | E-070 agent-finalize precedes E-071 projections/explicit suites and E-072 final CI/audit; E-073 records completion lint and the final update command; retained-run maintenance skipped because RESULTS_DIR was unset |
| Preservation | Pre-existing tracker edits incorporated; no unrelated source edits made |

The checklist below records current implementation completion. E-015–E-017
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
