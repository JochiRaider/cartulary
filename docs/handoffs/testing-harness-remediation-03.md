# Testing Harness Remediation 03 — Coordinated Execution Efficiency

This is a planning-only, evidence-backed gated roadmap. The document records
residual opportunities after remediation 01/02; it does not implement them.
Browser batching and narrow incident-test consolidation have concrete proposed
contracts. Cost calibration, further isolation relaxation and longer PostgreSQL
service lifetimes remain evidence-gated. No measured efficiency gain is claimed.

Document date: 2026-10-21, America/New_York. Artifact mode: **partial**.
The requested roadmap and gate definitions are provided; implementation discovery,
cross-owner independence proof and runtime characterization remain incomplete.
Implementation authorization: **not granted by this planning invocation**.

## 1. Scope and Source Posture

### 1.1 Target, output and authorized actions

The supplied repository audit target is `/home/jochi/code/cartulary`, with canonical
repository-relative identity `.` and target kind repository directory. The
inspected subject is the cross-directory test-execution and fixture-lifecycle
seam; this does not propose a new permanent module. The user-specified output
and label are `docs/handoffs/testing-harness-remediation-03.md` and
`testing-harness-remediation-03`. The permitted write location is this one file.

The repository root and every existing output-path ancestor were inspected as
directories without symlinks. The output was absent at admission. Its absence
and ancestor identities MUST be rechecked immediately before creation. A
concurrent file creation MUST stop replacement until its identity and contents
are reconciled; this task does not authorize overwriting another chat's work.

The [planning framework](cartulary_modular_refactor_planning_framework.md) and
the refactor-tracker skill's twelve-section format govern planning structure,
subject to repository instructions and the user's narrower write boundary.
Tests, builds, generation, formatting, finalization, service control, recovery,
runtime observations and performance experiments were not authorized here.
Only static reads, in-memory analysis and creation/revision of this document
are admitted. No additional script, report, scratch file or configuration edit
is part of this deliverable.

### 1.2 Authority and normative vocabulary

| Statement class | Meaning and authority | Required interpretation |
| --- | --- | --- |
| Adopted requirement | Named clause in the adopted/current [Testing Harness NLSpec](../testing-harness-nlspec.md), or a product primary owner allocated by [Core 00](../spec/00_document_set_status_and_precedence.md) | The source clause governs. This handoff is a subordinate reference, not a new owner. |
| Proposed requirement | R03-001 through R03-017 in Section 7 | MUST/MUST NOT describe conditions of a later authorized change. A proposal cannot override an adopted requirement before adoption. |
| Handoff procedure | Source qualification, entry gates, evidence and completion rules in this document | MUST/MUST NOT constrain this roadmap and its future acceptance records; they do not authorize execution. |
| Observed behavior | Directly inspected implementation or parsed metadata with an E-ID | Describes the inspected baseline; neither requirements completeness nor successful execution follows. |
| Rationale / hypothesis | Explanation of benefit or risk | Is not normative; an unmeasured opportunity cannot become a measured result by repetition. |
| Historical result | A named receipt in an earlier handoff | Is attributed to that document and source identity; it is not independently verified here or a current-source pass. |

Creating or approving this handoff MUST NOT amend an adopted NLSpec, grant
production implementation authority, establish product conformance, or activate
Core 05 publication. Tests, generators, runtime metadata, conformance and release
evidence MUST NOT read, stat, hash or otherwise depend on this document or other
Markdown. Static authoring inspection of documents is not executable evidence.

Core 00 owns allocation. Core 01 REQ-01-157, 160, 162–167 owns the incident-create
normalization, atomicity, replay and response behavior relevant to S-002.
Core 04 remains the security/conformance owner; Core 05 is publication-only.
No production behavior or presentation redesign is proposed. A newly discovered
conflict MUST be resolved through its primary owner, not preserved as an
accidental compatibility contract.

### 1.3 Baseline and preservation boundary

E-013 is the baseline for this writing session, not the historical executable
snapshot reported by remediation 02.

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

Remediation 02's opening executive summary and earlier S-008 table describe
blocked aggregate acceptance. A receipt appearing later in that file, now
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
the active continuation's later failures. During this writing session,
remediation 02 acquired additional active receipts. Its
[full-7 terminal receipt](testing-harness-remediation-02.md#d-003-full-7-terminal-receipt-and-next-repairs)
reports native 68130 / exit 2, 58 passed and four failed of 62 units, at source
`sha256:4dd8102743aac96c39c3138f7ddf4adb15da5928e1e5ada3d67eee4241452322`;
the document assigns subsequent repairs and validation to that continuation.
Those are historical reports from another chat, not observations of this audit.

The active [completion scope](testing-harness-remediation-02.md#current-completion-scope)
also says historical PostgreSQL causal attribution is dropped from that
continuation's acceptance. This changes its scope, not the factual status of
unproved causes. The original causal questions and closure criteria remain
recorded in the appendix; this handoff neither resolves them nor reactivates
them as mandatory work. W4's absent historical comparison likewise remains
unproved even where deterministic lifecycle regressions are reported passing.

E-012 records these distinct source/session postures. No retained run, process,
private proof or diagnostic bundle was opened here. This document MUST NOT
independently certify any reported result, turn the appendix into current-source
acceptance, edit the earlier summary, or impose a blanket historical blocker on
independently safe planning. Later implementation requires a fresh source and
acceptance recheck under RB-001.

## 2. Current-State Repository Inventory

### 2.1 Declared inspection universe

The bounded input universe contains the 124 distinct tracked files explicitly
listed below. Of these, 46 received direct inspection of relevant definitions,
scenarios or guidance; 78 received metadata-only accounting. Relevant reads
span this conversation and the final source recheck; content was compared against the writing-session snapshot. Concurrent changes
to the Makefile, pin registry, harness specification and remediation 02 were
reinspected and qualified under E-014. Direct inspection is not a
claim that every line or transitive effect was audited.

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

E-002 parses active rows from all 66 family JSON files. Counts are semantic
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

E-004 reconstructs the existing deterministic arithmetic in memory without
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

All harness clauses below refer to the adopted/current v3 NLSpec inspected in
E-001. Proposed changes are not current conformance violations.

| Contract / surface | Observed behavior and evidence | Adopted owner / required behavior | Disposition | Prerequisites, risk and validation |
| --- | --- | --- | --- | --- |
| C-001 Selection and attribution | Go/Vitest grouping removes reporting-owner partitions; browser work units already execute as harness.browser; rows retain their identities; E-003/E-004 | REQ-379/805/806; exact selection, attributable native outcomes, no widening | Preserve exact selection/accounting | Schema/consumer discovery and mixed-owner cases; S-001, V-001/V-002; RB-002 |
| C-002 Browser compatibility and state | Browser partition owner is required; test-owned rows can pack, isolated/measurement rows are singleton; E-004 | REQ-806 currently preserves owner and state partitions; AC-088 | Proposed owner amendment only for reporting-owner boundary | Adoption and cross-owner independence proof; S-001/S-004, V-002/V-003; RB-002/RB-003 |
| C-003 PostgreSQL/object ownership | Shared rollback cluster; unit-bound exclusive cluster; fresh clones and namespace cleanup; E-005 | REQ-804/811; AC-086/095; provisioning/retirement are within scope, cleanup uncertainty prevents reuse | Preserve; S-005 proposes a separate lifetime amendment | Complete connection/effect tracing and settlement evidence; V-003/V-007; RB-005 |
| C-004 Renderer/trust and measurement | Existing native pool; connection-owned browsers; sealed snapshots; E-006/E-010 | REQ-804/806 and populated-fixture REQ-812–815; immutable trust, quiet measurement and dedicated review | Preserve | No shared contexts, persistent Chromium daemon or measurement shortcut; V-003/V-005 |
| C-005 Failure, cancellation and evidence | Settled file outcomes and physical ownership outlive individual producer activity; E-006/E-008 | REQ-803–806; AC-085–088; primary/secondary failures and quarantine retained | Preserve | Failed/reset/interrupted work must not become success or silently retry; V-002/V-003/V-005 |
| C-006 Costs and resource admission | Generic per-row Go/Vitest estimates; browser title-scaled estimates; serialized rollback bucket; E-003/E-005 | REQ-803/805/806; claims cover actual lifetimes and bounded deterministic execution | Preserve defaults; evidence-gated calibration | Experimental limits and population fixed before candidate comparison; S-003, V-005; RB-004 |
| C-007 Product assurance | Full HTTP incident scenarios overlap in setup but have distinct assertions; E-009 | Core 01 REQ-01-157, 160, 162–167; Core 00 REQ-00-035 owner allocation | Preserve; add missing replay-header assertion while consolidating setup | Named survivor, unchanged route/status/envelope/auth/storage behavior, exact catalog update; S-002, V-004 |
| C-008 Documentation and projection boundary | Generated schemas/Make are downstream; lint does not select this file; prior receipt posture differs by section; E-001/E-011/E-012 | Harness REQ-001/003; root and tooling instructions | Preserve; documentation-only audit evidence | No runtime Markdown input, stale-pass claim, new compatibility layer or unauthorized linter change; V-006/V-008 |

HTTP/WebSocket, row/query/mutation, authorization, revisions, projections,
saved views, view schemas and UI selectors remain production-owner contracts.
S-001 changes execution placement, not those behaviors. S-002 changes where
HTTP assertions execute, not the product contract. Product API changes,
database migrations, frontend presentation and golden refresh are not applicable
to the proposed near-term edits. Any later discovered need for such a change
MUST become a separately owned/authorized task rather than an implicit part of
optimization.

## 5. Coupling and Boundary Findings

Benefits below are expected mechanisms, not measured rankings.

| Finding / classification | Direct evidence | Contract / owner | Required action or disposition | Slice / validation |
| --- | --- | --- | --- | --- |
| F-001 / should_fix | Owner participates in browser partition identity and validation; model yields 55 versus 17 cohorts; E-004 | C-001/C-002; harness.browser and catalog owner | Investigate/remove reporting-only partition after state proof and adoption. Expected medium–high benefit; medium complexity/reliability risk | S-001; V-001/V-002/V-003/V-005 |
| F-002 / should_fix | Standalone Location test starts an application fixture for one response assertion; replay scenarios overlap but differ; E-009 | C-007; module.incidents | Consolidate the named header assertion first. Small per-scenario benefit, low–medium complexity/risk; replay consolidation remains optional | S-002; V-004/V-005 |
| F-003 / defer | Go/Vitest price a catalog row independently of its selected test count; fixture borrowing can serialize; E-003/E-005 | C-006; harness.command_surface | Measure setup and execution separately before changing estimates/claims. Workload-dependent benefit, medium complexity/risk | S-003; V-005 |
| F-004 / defer | 113 transaction rows and 465 test-owned browser rows already use reuse paths; remaining eligibility is not proved; E-002/E-005 | C-002/C-003; affected semantic owners | No blanket conversion. Reopen only named candidates with complete effect proof. Potentially high local benefit and high reliability risk | S-004; V-003/V-007 |
| F-005 / defer | Exclusive cluster key includes unit ID and release retires it; E-005 | C-003; native fixture broker | Assess sequential exclusive reuse only if startup/template cost remains material. High potential benefit, high complexity/risk; existing contract forbids unrelated attachments | S-005; V-005/V-007 |
| F-006 / intentional/no_action | Grouped Go/Vitest, template clones, renderer pool, namespace reset and snapshots already exist; E-003/E-005/E-006/E-008/E-010 | C-001/C-003/C-004/C-005 | Preserve the foundations; no second broker, scheduler, resident test daemon or reimplementation justified | No new slice; preservation checks V-002/V-003 |
| F-007 / must_fix for this handoff's evidence accuracy | Blocked summary, successful appendix receipt and newer active failed receipts describe different source/session postures; E-012 | C-008; document author | Qualify chronology and unresolved historical causes, including the active continuation's acceptance-scope decision; do not edit remediation 02 here | WF-00/WF-06; V-008 |
| F-008 / must_fix for truthful validation reporting | Lint glob excludes remediation 03; Make readiness/cache/report writes exist; E-011 | C-008; documentation/tooling owner | Record noncoverage and skipped execution. Do not expand tooling scope or claim a lint pass | WF-06; V-006/V-008 |

No production module movement, vendor abstraction replacement or direct-storage
decoupling is justified by this audit. Finding classifications do not authorize
implementation; a must_fix item is confined to the acceptance named in its row.

## 6. Refactor Workstreams

This table is the sole authoritative workflow-to-workflow prerequisite graph.
The entry conditions in Section 7.5 and evidence gates in Section 11 also apply;
they do not create additional workflow-order edges. All work
after documentation preparation requires separately authorized execution.
"Terminal disposition" means accepted completion with evidence, or an explicit
DEFERRED/DROPPED decision with reason and owner; it never means silently waived
correctness.

| Workflow | Class | Authoritative prerequisites | Goal / likely components | Validation | Handoff checkpoint and binary exit |
| --- | --- | --- | --- | --- | --- |
| WF-00 Baseline and authority | root | None | Revalidate working tree, adopted clauses, counts, output identity, prior receipts and unresolved obligations | V-001/V-008 | Snapshot and bounded inventory recorded; discrepancies resolved or scoped blockers assigned |
| WF-01 Browser adoption and batching | chain | WF-00; RB-002 and RB-003 resolved; future authorization | S-001 across browser partition schema/generator, readers, graph and evidence consumers | V-002/V-003/V-005 | Adopted contract plus compatible-only mixed-owner grouping, exact outcomes and measured disposition |
| WF-02 Incident assertion consolidation | chain | WF-00; future authorization; V-004 pre-change characterization | S-002 in incident HTTP tests and authored catalog selectors | V-004/V-005 | Named survivor proves complete map with fewer fixture lifetimes, or a reasoned deferred/dropped disposition |
| WF-03 Cost/admission investigation | chain | WF-00; terminal dispositions for WF-01 and WF-02; RB-004 resolved | S-003 evaluates the resulting accepted workload and existing policy owners | V-005 | Calibration accepted against preregistered limits or explicit no-change result; no automatic tuning |
| WF-04 Isolation eligibility investigation | chain | WF-00; named candidate dossier and future authorization | S-004 in affected test owner and fixture declarations | V-003/V-007 | Each admitted candidate has proof and acceptance or retained stronger isolation |
| WF-05 Exclusive PostgreSQL lifetime investigation | chain | WF-00; terminal disposition for WF-03; RB-005 resolved; future authorization | S-005 within existing native broker only | V-003/V-005/V-007 | Adopted and validated bounded design, or explicit deferred/dropped proposal |
| WF-06 Document closure | chain | WF-00 | Record all slices, conditional inclusion rules, validation limits and present-session outcomes | V-006/V-008 | This document is complete as a gated roadmap; it does not require any future implementation workflow to run |
| WF-07 Future implementation closure | chain | WF-00; terminal disposition of each WF-01–WF-05 workflow explicitly activated by the later implementation authorization | Reconcile evidence, finalization and regression for accepted changes only | V-006/V-008 | Accepted changes pass required checks and physical settlement; deferred work remains explicitly deferred |

WF-04 and WF-05 are inactive by default. Omitting them leaves no prerequisite
unresolved for WF-01, WF-02 or document closure. WF-03 waits for recorded
dispositions, not mandatory implementation, of earlier recommendations. WF-07
uses the activation set recorded before implementation; it MUST NOT silently
activate an investigation or treat a failed required check as a deferral.
No parallel execution or additional agent delegation is requested by this graph.
If no optimization is activated, WF-06 may close with the evidence-backed
roadmap and all implementation items unexecuted.

## 7. Proposed Refactor Slice Plan

### 7.1 Proposed normative contracts and binary acceptance

The following are **proposed requirements**, subject to Section 1.2. They MUST
NOT be cited as adopted runtime behavior. Each requirement's acceptance column
is binary; an absent proof is unmet, not presumed true.

| ID | Proposed requirement | Binary acceptance / validation |
| --- | --- | --- |
| R03-001 | S-001 MUST remove reporting ownership alone from browser compatibility; it MUST preserve the exact remaining execution/state tuple in Section 7.2. Only test-owned rows may form multi-row cohorts. | Owner-only difference merges in the eligible fixture; changing any required tuple dimension separates or rejects the selection; isolated/measurement rows remain singleton. V-002 |
| R03-002 | Before a newly combined population executes, its owners MUST prove test-owned mutable data, actor/session isolation and compatible hooks/background effects. Unknown effects MUST retain existing separation until resolved. | The dossier covers every candidate row and shared helper, including listings/admin state, clocks/faults, cleanup and deferred work; alone/combined/reordered cases have no cross-test effect. Missing dossier blocks implementation. V-003 |
| R03-003 | Selection MUST precede packing. Row identities, selectors, reporting owners and verification attribution MUST remain exact. The work-unit execution owner MUST remain harness.browser. Ownership MUST be resolved from the catalog, not the first row. | Exact owner subsets execute no foreign rows; the selected semantic identities appear exactly once; missing/extra/duplicate outcomes fail acceptance. V-002 |
| R03-004 | The partition interface MUST remove owner_id and normalized ownerID rather than add an owner fallback/list registry. A versioned, coordinated producer/consumer cutover MUST reject predecessor input; no dual reader, alias or grouping toggle is permitted. | Complete consumer inventory is migrated together; generated-schema checks and malformed/predecessor-input cases pass; no old ownership lookup remains. V-002/V-006 |
| R03-005 | Packing MUST retain deterministic cost ordering, stable ties, indivisible selections, existing timeout headroom, one worker and max_failures=1. It MUST NOT extend deadlines or split catalog outcomes to force fit. | Same selected inputs yield the same identities/bindings; exact-fit and oversized cases preserve existing policy; a failing test leaves unexecuted outcomes explicit. V-002 |
| R03-006 | Grouping MUST preserve fresh contexts/actors, affinity/reset boundaries, failure quarantine, cancellation settlement and independently settled outcomes. Product work MUST NOT retry. | Product, teardown, reset, cancellation and owner-loss injections preserve primary/secondary failures, prevent contaminated reuse and retain exact owned-resource cleanup. V-002/V-003 |
| R03-007 | S-002 MUST move the first-create Location assertion to the named survivor in Section 7.3 before retiring the standalone symbol. It MUST also assert the adopted replay Location contract on the survivor's existing replay response. | First response is 201 with exact Location; replay is 200 with the same resource Location; existing payload and actor-scope assertions remain; no extra creation request or fixture is added. V-004 |
| R03-008 | Removing any further scenario MUST require an assertion-to-survivor map covering all distinct behaviors and evidence contexts. Catalog selector changes MUST be atomic with test changes and MUST NOT preserve obsolete aliases. Helper extraction alone MUST NOT count as an efficiency improvement. | Every removed assertion has a named, independently executable survivor in the same required assurance context; unsupported equivalence blocks removal. V-001/V-004 |
| R03-009 | Every efficiency comparison MUST preregister population, source/environment/cache identities, repetitions/order, primary metric, numerical benefit/regression limits and missing-data treatment before candidate execution. | A dated fixed comparison record exists before candidate results; post-hoc threshold selection or incomparable environments fails performance acceptance. V-005 |
| R03-010 | S-003 MUST separate startup, file setup, selected execution, reset, cleanup and publication costs and calibrate existing deterministic policy only when the measurements support change. | Every changed estimate/claim has measured provenance; within-noise benefit or unexplained resource/tail regression produces a no-change/deferred result. V-005 |
| R03-011 | Resource admission MUST account for physical resident services and active work, including cleanup headroom. Existing browser lanes, worker counts, capacities and deadlines MUST remain the defaults unless a separately reviewed measured proposal changes them. | No child work is uncharged; actual configuration matches retained identity; increased concurrency cannot be accepted solely because some queues shorten. V-003/V-005 |
| R03-012 | S-004 MUST preserve stronger isolation unless every consequential operation is confined to its proposed transaction/state boundary. Independent commits, connection escapes, workers or global effects MUST invalidate a rollback-only proof. | Candidate-alone, reordered and failure-settlement cases pass; deliberate escape/cleanup failure prevents readmission; no approved declaration without complete effect evidence. V-007 |
| R03-013 | S-005 MUST remain deferred until physical startup/template overhead, safe settlement and an adopted replacement for REQ-804's lifetime restriction are demonstrated. | Benefit dossier, adoption record and complete lifecycle design all exist; failure of any gate forbids implementation. V-005/V-007 |
| R03-014 | If S-005 is later admitted, only the existing run-owned broker may hold a bounded pool; unrelated borrowers MUST be sequential, use fresh databases and match physical/provisioning/template identity. No mutable database or cross-run daemon may be reused. | Concurrent unrelated borrow is denied; stale/foreign identity is rejected; fresh database state is demonstrated; no second manager exists. V-007 |
| R03-015 | Before any S-005 reassignment, all borrower connections, work, databases and namespaces MUST settle; uncertainty MUST quarantine the physical generation and retain its resource claims. Cluster mutation, migration, lifecycle/deadline, checkpoint and measurement cases MUST retain dedicated lifetimes. | Failure/cancellation/publication-loss cases cannot reassign the generation; excluded cases never use the pool; final physical termination releases ownership exactly once. V-003/V-007 |
| R03-016 | Existing renderer pooling, rollback fixtures, durable template cloning, immutable performance snapshots, runner grouping and incremental evidence MUST remain intact unless a separately justified owner change is adopted. | Preservation cases and exact evidence/cleanup contracts pass; no recreated infrastructure layer or weakened durability substitutes for optimization. V-002/V-003/V-006 |
| R03-017 | Static counts, model estimates, historical receipts and current measured results MUST remain distinct. Missing observations MUST remain unknown; a failed required assurance check MUST NOT be hidden by a performance gain. | Handoff/evidence review can trace every claim to its class and source; no estimated milliseconds or historical pass is reported as a current speedup. V-005/V-008 |

### 7.2 S-001 interface and compatibility definition

The unchanged compatibility dimensions are stage/target and evidence context,
project/configuration, runtime and immutable renderer/trust requirements,
resource profile, fixture capability/profile, PostgreSQL isolation,
service dependencies, preparation operations, browser session/affinity,
state scope, invocation/failure policy and timeout policy. The generator already
partitions within topology policy groups; this proposal MUST NOT merge distinct
policy groups merely because they reference the same service. Missing or invalid
required metadata MUST fail validation, not fall back to owner or filename.

| Interface | Current | Proposed change / invariant |
| --- | --- | --- |
| Authored test-family row | owner_id, exact selector, explicit state/fixture/resource declarations | Retain all row ownership and verification fields. No metadata-only declaration is proof of cross-owner independence. |
| Browser partition projection | cartulary.browser_e2e_batch_manifest.v14 requires owner_id | Adopt a successor schema removing partition owner_id; remaining fields preserve their meanings. The successor version MUST be allocated after baseline revalidation, before implementation, and recorded with all consumers. |
| Normalized browser partition | ownerID required by reader and cohort validation | Remove ownerID; resolve each selected row through the canonical catalog. Do not replace it with first-owner inference, optional-owner compatibility or a redundant owners array. |
| Executable cohort | Derived name/digest, exact selectedRowIDs/specs, policy and one worker | Recompute identities from revised compatibility and exact selected selectors; reporting ownership does not partition execution. Unselected rows never enter the cohort. |
| Work graph and results | owner_id=harness.browser; row_bindings identify selected catalog rows | Preserve execution owner and per-row ownership/verification attribution. No command-line or environment selection registry is added. |
| Public Make selection | Existing owner/row/suite selection | No new option or grouping mode. Direct and aggregate selection share the same compiler semantics. |
| Migration | Current schema/generator/readers and generated projections | Adopt owner clauses, migrate authored schema and consumers, regenerate through Make, then verify drift/schema/selection. Do not hand-edit generated outputs or translate old retained artifacts into new acceptance evidence. |

RB-002 includes final discovery of validator attachments, CLI selection,
reporters/normalizers, fixtures, generated artifacts and all partition readers.
The document does not pretend those consumers have been exhaustively audited.
The schema successor's allocation is an explicit adoption deliverable, not a
runtime compatibility fallback.

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

The mandatory first slice uses the existing survivor
`TestIncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance`.
It adds Location assertions to the first-create and replay responses already
produced by that test, then removes
`TestIncidentCreateReturnsStableLocationHeaderHTTPConformance`.
Both currently belong to the same multi-symbol catalog row, so this removes a
selected symbol and one test fixture lifecycle, not necessarily a work unit or
container. Existing response/request counts in the survivor remain unchanged.

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

- `module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972` retains its row ID, ownership, verification IDs and other four selected test symbols; remove only the retired Location symbol.
- `module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6` remains unchanged in the first slice.

The optional second step, replay-scenario consolidation, is DEFERRED. It MUST
produce a complete revised map and independently runnable survivor before
deleting either scenario. S-002's first step does not depend on that decision.

### 7.4 Deferred investigations and defaults

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
policy. No numerical benefit threshold is fabricated from the static model.
Until that record exists, implementation tuning remains blocked.

Measure real process/container start counts, template preparations, clone/drop
counts, resets, worker starts, teardown and publication work alongside wall
time, attributed CPU, observed memory and resource waits. Existing CPU readings
may be lower bounds and sampled RSS may omit processes or count shared pages
more than once. These limits MUST survive comparisons. A new instrumentation
subsystem is not a default prerequisite; missing data first receives a bounded
disposition through the existing evidence owner. The inspected REQ-835
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
quarantine MUST physically retire the service through the same owner. Dedicated
cluster exclusions and current REQ-804 continue to govern until adoption.

### 7.5 Slice entry and completion table

All implementation kinds below **require later authorization**.

| Slice / kind | Findings / contracts | Workflow dependency | Entry conditions and readiness | Completion / validation | Rollback or migration |
| --- | --- | --- | --- | --- | --- |
| S-001 proposed behavior/interface change | F-001; C-001/C-002/C-004/C-005 | WF-01 | **blocked**: RB-001/RB-002/RB-003/RB-004/RB-006; owner adoption, full consumer inventory, effect dossier, characterization and comparison record | R03-001–006/016/017 pass; V-002/V-003/V-005/V-006; fewer actual cohort/reset boundaries with acceptable resource/runtime results | Coordinated version cutover; revert source/schema/projections together if rejected; no dual mode or old-run reinterpretation |
| S-002 test restructuring | F-002; C-007 | WF-02 | **blocked** for implementation: RB-001/RB-006 and unexecuted V-004 baseline; named initial mapping is ready for review | R03-007/008/017 pass; one fewer fixture lifecycle, no extra survivor setup and all mapped assertions pass; optional replay step may remain DEFERRED | Restore removed test/catalog selection if coverage fails; do not retain obsolete selector aliases |
| S-003 measured policy investigation | F-003; C-006 | WF-03 | **blocked**: RB-004/RB-006; post-disposition workload and fixed experiment record | R03-009–011/017 pass or reasoned no-change result; V-005 | Revert rejected authored policy changes and regenerate normally; no permanent tuning compatibility mechanism |
| S-004 eligibility investigation | F-004; C-002/C-003 | WF-04 | **blocked / DEFERRED**: RB-003/RB-004; no new candidate approved | R03-012/016/017 plus V-007 for every admitted candidate, or retained stronger isolation | Restore original fixture contract if proof fails; no fallback auto-detection |
| S-005 lifetime investigation/change | F-005; C-003/C-004/C-005 | WF-05 | **blocked / DEFERRED**: RB-004/RB-005; benefit, adoption and lifecycle design unproved | R03-013–017 and V-003/V-005/V-007, or explicit defer/drop | No live mixed epochs; settle exact old ownership before switching; revert coordinated owner/projection/implementation if unsafe |

## 8. Validation Plan

### 8.1 Future canonical validation

Nothing in this section authorizes a command in the present writing task.
All future repository operations run from `/home/jochi/code/cartulary` through
public Make. Selection placeholders are explicit future discovery gates, not
runnable commands with invented row IDs. RB-006 MUST close by recording literal
selections before execution. Use fresh literal run IDs, capture native exit
separately from summaries, and retain exact source/environment identities.

| Validation / stage | Contracts and slices | Required cases / public selection | Required evidence and failure consequence |
| --- | --- | --- | --- |
| V-001 pre-change discovery | C-001/C-007; S-001/S-002 | `make help ALL=1`; `make task-guide OWNER=harness.browser`; `make task-guide OWNER=harness.command_surface`; `make task-guide OWNER=harness.test_catalog`; `make task-guide OWNER=harness.evidence_accounting`; `make task-guide OWNER=module.incidents` | Freeze exact row populations and validation routes. Unknown selection blocks dependent execution; do not invent a private runner. |
| V-002 pre/post grouping | C-001/C-002/C-005; S-001 | `make test-slice OWNER=<frozen-owner> ROWS=<frozen-rows>`; compiler/adapter cases for exact/partial/owner subsets, repeated titles, mixed owners, incompatible tuples, deterministic identity, predecessor schema rejection, oversized rows, failure and teardown | Exact native outcomes, bindings, digests and schema results. Missing/duplicate/widened execution or stale success blocks acceptance. |
| V-003 pre/post lifecycle | C-002–C-005; S-001/S-004/S-005 | Owner-discovered test-slice rows for contexts/actors, combined/reordered populations, concurrent borrowers, reset/cancellation/owner-loss, namespace safety, rollback/close/drop failure and quarantine | Logical and physical settlement plus row outcomes. Uncertain resource ownership remains charged and quarantined; no retry to obtain a pass. |
| V-004 incident characterization and survivor | C-007; S-002 | `make test-slice OWNER=module.incidents ROWS=module.incidents.support_integration.incident_create_bootstraps_creator_and_workbook_e012cac972,module.incidents.integration.real_create_with_trimmed_decomposed_input_normal_f1c6efb6b6`; use each row separately as well | Native Go JSON identifies all intended tests and assertions; complete before/after preservation map and fixture counts. Missing Location/replay/actor/durable checks blocks deletion. |
| V-005 matched efficiency comparison | C-006; all activated optimization slices | `make test-slice OWNER=<frozen-owner> ROWS=<frozen-rows> HARNESS_DIAGNOSTICS=basic`, or the frozen browser suite; `make explain-run RESULTS_DIR=<exact-run-root> DETAIL=performance JSON=1`; `make harness-observability-check RESULTS_DIR=<exact-run-root> CHECK=evidence` | Preregistered population/limits, exact runs, lifecycle costs, wall/CPU/memory/waits and omissions. Reader success is not execution or performance acceptance; unknown/within-noise benefit remains unaccepted. |
| V-006 future implementation final gate | All changed contracts | make agent-finalize before broader final verification; authored projection changes use make generate, make source-check CHECK=generated and make source-check CHECK=schema; required owner-routed aggregate gates and discovered explicit affected selections; make ci only when required by the affected validation closure | Source drift/schema, required regression, native exits and complete cleanup. Visual is explicit: make browser-e2e SUITE=visual when affected. Review contracts use make harness-contract SUITE=ui-review when affected. No golden update implied. |
| V-007 optional eligibility/lifetime gate | C-002/C-003; S-004/S-005 | Candidate-owned test-slice and harness lifecycle rows after exact discovery; permitted Go-only order check uses a recorded positive GO_SHUFFLE_SEED | Transaction escape, independent commits, mutable global state, cluster residue, stale generation, forbidden overlap, failed reassignment and physical cleanup cases. Failure retains prior isolation/lifetime. |
| V-008 document review / handoff | C-008; all slices | Static review of this document, links, requirement IDs, dependency graph, traceability, claims and source baseline; no repository test command | Every reference resolves, each proposal has an acceptance predicate, and all unexecuted or historical results are qualified. Incomplete evidence stays explicit. |

For V-005, use the same declared instrumentation/cache policy for compared
treatments; report cold and warm conditions separately. A fresh execution
comparison uses the supported `CARTULARY_HARNESS_CACHE_MODE=off` when required
by the experiment. Compare exact roots with the existing equivalent-comparison
reader only when it admits the source/selection identities; rejected comparison
compatibility MUST remain a recorded limitation, not be bypassed. Performance
sampling MUST honor measurement quiet periods. No live polling or companion
diagnostic work may be introduced into those intervals by this roadmap.

S-002's mandatory initial change stays within the existing multi-symbol row.
The row's required siblings remain selected; they are not removed merely to
obtain a lower count. Independently executing the survivor is a future
characterization requirement whose canonical exact-symbol selection route MUST
be settled under V-001/RB-006 before changing catalog structure. Do not use a
direct go test workaround or silently broaden the proposal.

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

### 8.2 Finding-to-completion traceability

Completion evidence below is a required future receipt, not evidence produced
in this writing session. The document-only rows identify their authoring receipt.

| Finding | Contract / proposed requirement | Slice / workflow | Validation | Required completion evidence |
| --- | --- | --- | --- | --- |
| F-001 | C-001/C-002; R03-001–006/016/017 | S-001 / WF-01 | V-001/V-002/V-003/V-005/V-006 | Owner adoption, exact schema/consumer roster, cross-owner effect dossier, selected-row outcome equality, settled failure cases and matched lifecycle/runtime results |
| F-002 | C-007; R03-007/008/017 | S-002 / WF-02 | V-001/V-004/V-005 | Assertion-to-survivor map, independently attributable survivor execution and one fewer observed fixture lifecycle; replay candidate remains separately accepted or deferred |
| F-003 | C-006; R03-009–011/017 | S-003 / WF-03 | V-005 | Preregistered numerical limits, phase/resource measurements and an accepted policy change or explicit no-change result |
| F-004 | C-002/C-003; R03-012/016/017 | S-004 / WF-04 | V-003/V-007 | Candidate-specific effect proof and alone/combined/failure outcomes, or a retained stronger boundary with reason |
| F-005 | C-003; R03-013–017 | S-005 / WF-05 | V-003/V-005/V-007 | Adopted lifetime contract, bounded pool/settlement design, measured benefit and exclusion/quarantine cases, or documented deferral |
| F-006 | C-001/C-003/C-004/C-005; R03-016 | Preservation within activated slices | V-002/V-003/V-006 | Unchanged valid evidence or affected preservation cases under current incremental-evidence rules; no new optimization claim |
| F-007 | C-008; evidence-classification procedure | WF-00/WF-06 | V-008 | E-012/E-014: attributed historical success, active receipt and acceptance-scope distinctions, and unresolved causes |
| F-008 | C-008; validation-reporting procedure | WF-06 | V-006/V-008 | E-011/E-014: lint omission, side effects and unperformed automated validation explicitly recorded |

### 8.3 Present-session command admission and skips

| Operation | Invocation-path evidence / effects | Present disposition |
| --- | --- | --- |
| rg, sed, nl, Git with optional locks disabled | Read-only source/status inspection; no private runner imported | Performed, E-001–E-013 |
| Python standard-library in-memory analysis | Reads explicit JSON/text, computes counts/hashes and models arithmetic; no repository module execution or output file | Performed; static evidence only |
| Public Make discovery and target-plan | No Make command needed for this bounded static task; future discovery is specified, not executed | Not run; RB-006 preserves required runtime selection discovery |
| make lint-markdown | Generated Make binding performs Node readiness, preflight and frontend-install prerequisites; runner writes cache stamp and retained artifacts; current config omits this file and public target declares no selection input | Not run; no lint coverage/pass claimed; no config/tooling repair authorized |
| Tests, browser sessions, builds, generation, formatting, service/recovery operations | Would execute product/harness work, mutate outputs or interfere with concurrent development | Not run, explicitly excluded |
| agent-finalize and retained-run maintenance | Implementation-only operation can rewrite generated artifacts | Not applicable to this documentation task; not run |
| Document structural inspection | Reading the new file and checking local references/IDs/tables/dependencies in memory | E-014; authoring inspection, not an automated lint/test suite |

Automated repository validation is **unperformed**. Static document inspection
MUST NOT be relabeled as a canonical lint pass. Adding this handoff to Markdown
lint selection requires a separately authorized documentation/tooling change;
the absent coverage does not authorize a private linter or broad config edit.

## 9. Top-Level Work Tracker

These are the only work-status values: TODO, IN_PROGRESS, BLOCKED, DONE,
DEFERRED and DROPPED. Readiness, artifact mode and performance acceptance are
separate fields.

| ID | Work item | Workflow | Status | Dependencies / evidence | Exit condition |
| --- | --- | --- | --- | --- | --- |
| T-001 | Static baseline, census, model and authority reconciliation | WF-00 | DONE | E-001–E-013 | Bounded source findings and model revalidated; no execution claims |
| T-002 | Browser consumer/adoption and state-proof dossier | WF-01 | BLOCKED | F-001; RB-001/RB-002/RB-003/RB-006 | Adopted interface plus complete effects/consumer/validation inventory |
| T-003 | Implement and validate eligible browser batching | WF-01 | BLOCKED | S-001; T-002; RB-004 | R03-001–006/016/017 accepted |
| T-004 | Incident header consolidation and characterization | WF-02 | BLOCKED | S-002; V-004; RB-001/RB-006 | Named survivor preserves first/replay Location and all existing assertions; removed fixture demonstrated |
| T-005 | Optional replay scenario consolidation | WF-02 | DEFERRED | S-002 optional step; R03-008 | Full distinct-assertion map and benefit, or retain both scenarios |
| T-006 | Fixed workload and cost/admission comparison | WF-03 | DEFERRED | S-003; RB-004; WF-01/WF-02 dispositions | Measured policy proposal or supported no-change result |
| T-007 | Additional isolation candidate review | WF-04 | DEFERRED | S-004; RB-003 | Named effect proof and validation; no blanket relaxation |
| T-008 | Exclusive PostgreSQL lifetime proposal | WF-05 | DEFERRED | S-005; RB-004/RB-005 | Benefit/adoption/settlement gates satisfied or explicit deferral |
| T-009 | New handoff and bounded preservation review | WF-06 | DONE | V-008; E-014 | All twelve sections, traceability and final bounded comparison recorded |
| T-010 | Future implementation acceptance | WF-07 | TODO | Later authorization and activation set | All activated work has accepted or explicit deferred/dropped disposition; required assurance failures remain open |

DONE on T-001 or T-009 MUST NOT be interpreted as implementation completion.
Historical W1/W2/W4 evidence and uncertainty remain recorded in remediation 02;
this tracker neither closes, reactivates nor duplicates those investigations.

## 10. Session Handoff Log

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

All E-001–E-011 observations refer to the working tree at E-013 and the source
reconciliation recorded in E-014. Earlier reads were reused only for unchanged
inputs; affected changed sources were reinspected. Locations below identify relevant definitions,
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

### 10.9 Final authoring receipt

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

## 11. Open Questions and Blockers

| ID | Blocker / why it matters | Owner and required evidence | Affected work / status |
| --- | --- | --- | --- |
| RB-001 | Concurrent working-tree baseline and later execution authority are not frozen by this document | Future implementation owner: revalidate source/dirty-state identity, coordinate with existing work, and obtain later scoped authorization | All implementation; BLOCKED until entry revalidation |
| RB-002 | Cross-owner browser behavior needs adoption and complete partition-consumer migration inventory | harness.browser with catalog/evidence owners: amendment to REQ-806 and affected acceptance clauses; successor schema allocation; exact readers/validators/reporters/generated attachments and characterization | S-001; BLOCKED |
| RB-003 | Complete cross-owner or newly relaxed state independence is unproved | Each affected semantic owner: connections, helpers, hooks, listings, actor/admin/session state, clocks/faults, background and cleanup dossier | S-001/S-004; BLOCKED; does not block S-002's mapping |
| RB-004 | No matched runtime measurements or preregistered numerical acceptance limits were collected | Harness performance owner: fixed experiment record, compatible exact runs, variability/missing-data policy and observed lifecycle/resource costs | Performance acceptance and S-003/S-005; BLOCKED; static model is insufficient |
| RB-005 | Current REQ-804 forbids unrelated exclusive allocations sharing a generation; complete replacement lifecycle is undesigned/unvalidated | Native broker owner: material-benefit proof, adopted ownership/residency semantics, generation/failure/cancellation/eviction/cleanup protocol and exclusions | S-005; DEFERRED |
| RB-006 | Future exact validation population and survivor-only canonical selection have not been discovered/executed | Implementer through public task-guide/target-plan: record literal row selections and any required authored selector restructuring before runtime work | Implementation readiness; BLOCKED; command families and initial incident row pair are known |
| RB-007 | Canonical Markdown lint currently omits this handoff and has incompatible side effects for this task | Documentation/tooling owner under separate authorization: admit exact document coverage and permitted canonical validation | Automated document-validation claim only; DEFERRED; does not authorize tool changes here |
| RB-008 | Broader semantic redundancy and transitive effects were not exhaustively audited | Future candidate owner: inspect exact source/owner contracts and complete preservation map | Optional S-002 replay consolidation and any newly proposed deletion; DEFERRED |

No unresolved primary-owner contradiction was established. Remediation 02's
summary/appendix/active-receipt distinctions are handled in Section 1.4.
Unproved historical causes MUST NOT be represented as repaired or newly proven
current failures; active acceptance-scope decisions remain owned by that
continuation.

## 12. Binary Completion Criteria

### 12.1 Independent outcomes

| Outcome | Current statement |
| --- | --- |
| Requested document delivery | **complete**: all five recommendations, required mappings and gates are recorded; static authoring receipt is E-014 |
| Planning completeness | **partial** at implementation-discovery depth: bounded source audit and complete roadmap/gate definitions, with RB-002/RB-003/RB-006/RB-008 explicitly pending |
| Per-slice readiness | S-001/S-002/S-003 blocked for implementation; S-004/S-005 blocked and deferred; see Section 7.5 |
| Implementation authorization | **not granted by this planning invocation**; authorization covers this handoff only |
| Measured performance acceptance | **unperformed**; model/counts do not establish actual execution savings |
| Canonical automated validation | **unperformed**; lint omission and side effects recorded |
| Session write-boundary verification | **bounded verification complete**: this task wrote only the new handoff; concurrent source/status changes are recorded in E-014; whole-workspace preservation is not certified |

### 12.2 Document acceptance checklist

| Criterion | Pass condition | Present evidence / disposition |
| --- | --- | --- |
| Target/output admission | Exact repository target and one nonsymlink output admitted; no replacement collision | E-013/E-014; MET: absent before creation, regular single-link output afterward |
| Accounted inspection scope | Every file in the 124-file bounded set has a named disposition; pending implementation discovery remains visible | Section 2; MET for bounded accounting, not exhaustive repository audit |
| Authority separation | Adopted clauses, proposals, observations, rationale and historical results cannot be mistaken for one another | Sections 1/4/7; static review under E-014 |
| Behavioral preservation | Named contract/state/failure boundaries and incident assurance survivors are explicit | C-001–C-008; R03-001–017; no execution proof claimed |
| Traceability | Every actionable finding reaches a contract, slice, validation and evidence source; no-action findings have reasons | Sections 3–8 and E-register |
| Dependencies | Workflow graph is acyclic; optional deferrals leave no dangling mandatory dependency | Section 6; static review under E-014 |
| Testable requirements | Every R03 requirement has a binary acceptance predicate and future validation route | Section 7.1; runtime predicates currently unexecuted |
| Defaults and interface boundaries | Existing runtime defaults remain explicit; proposed owner-field removal and clean schema cutover are unambiguous | Sections 7.2/7.4; successor allocation is a gated adoption deliverable |
| Historical/current evidence | Appendix success, active failed receipts, scope decisions and unresolved causes are separately attributed; no historical pass becomes current acceptance | E-012; Section 1.4 |
| Validation honesty | No tests/lint/generation/finalization reported as executed; omitted lint coverage disclosed | Section 8.3; E-011 |
| Preservation | Only this handoff was written; compared inputs/index/status unchanged except the new path, or mismatches reported | E-014; MET for this task's write scope, with concurrent changes and verification limits disclosed |
| Resumability | Next implementer can select a named workflow, revalidate source, and resolve its scoped gates without assuming unsupported readiness | Sections 6/9/11; implementation discovery remains explicitly bounded |

Future implementation completion additionally requires every activated
requirement's acceptance evidence, successful required regression and physical
settlement, and an explicit performance disposition. A documentation completion,
a reduced modeled cohort count, a passing reader, or a historical successful run
MUST NOT satisfy those implementation gates.
