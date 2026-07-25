# Fallow Dead-Code Cleanup Tracker and Handoff

## 1. Purpose, Scope, and Authority

This document tracks the investigation and later resolution of the Fallow
dead-code findings captured in `temp/fallow-dead-code.json`. It is a planning,
accounting, and handoff artifact. It does not itself authorize product behavior
changes, generated-file edits, lockfile edits, dependency removal, bulk file
deletion, or suppression of findings.

| Item | Baseline posture |
| --- | --- |
| Repository | `github.com/JochiRaider/cartulary` |
| Branch and commit | `main` at `50c9ea2c065c3585f724465d126efbd5dff2a6b2` |
| Source report | `temp/fallow-dead-code.json` (ignored by Git) |
| Source fingerprint | SHA-256 `bbfb9b7354e1fea0fac4d91b813dfd483051e2ffad0475c62eeb9e061bb9da68` |
| Source timestamp | 2026-07-12 20:37:20 EDT |
| Fallow identity | Fallow `2.93.0`, dead-code schema `7`, analysis run `run_6327d4e5abc7b521` |
| Finding count | 1,656 |
| Output path | `docs/handoffs/dead-code-clean-up-tracker.md` |
| Current status | Dead-code cleanup remediation is complete through DC-010 final validation and handoff. Residual Fallow issues are limited to the four documented pending-queue class-member analyzer exceptions with current caller evidence and removal triggers. |

Source and authority posture:

1. Core 00 through Core 04 own current product behavior. Core 05 applies only
   to claim-bearing timed, benchmark, fixture-sensitive, or publication
   evidence and is not used by this tracker.
2. `docs/testing-harness-nlspec.md` owns Fallow and Make invocation, retained
   artifacts, failure normalization, and harness-evidence mechanics.
3. `docs/domain.md` owns domain vocabulary and concept boundaries. Fallow,
   static-analysis findings, phase accounting, task surfaces, and generated
   harness artifacts are implementation-support concepts, not domain behavior.
4. The current repository owns implementation truth.
5. `docs/handoffs/cartulary_modular_refactor_planning_framework.md` guides the
   tracker structure and rigor but is not evidence of current repository state.
6. The source report is diagnostic evidence, not proof that a file or symbol
   can be removed safely.

The requester supplied this convenience command:

```sh
pnpm exec fallow dead-code --format json > temp/example.json
```

It writes `temp/example.json`, not the frozen source path, and direct `pnpm`
execution is not canonical harness evidence. Canonical reruns use this public
repository command from the repository root:

```sh
make frontend-fallow-static
```

The retained `dead-code.json` emitted by that target becomes the source for a
new baseline only after its run root, fingerprint, version, schema, and counts
are recorded here.

## 2. Binding Cleanup Doctrine

Every disposition and implementation slice must satisfy these gates.

| Gate | Required decision rule | Rejection condition |
| --- | --- | --- |
| Structural quality | Prefer a cohesive ownership, dependency, or public-surface improvement over a tactical deletion or suppression. | The change hides the finding without improving the owning structure. |
| Future growth | Treat later phase, target, module, package, and workbook-surface growth as a design constraint. | Adding the next owner or phase would require more central imports, exceptions, or compatibility branches. |
| Legacy value | Retain legacy support only when a named current consumer or continuing operational need proves value. | Retention rests only on age, possibility, or undocumented compatibility. |
| Compatibility burden | Remove or simplify behavior whose compatibility cost exceeds its continuing value. | The change adds a dual path, alias, wrapper, or fallback without an exit plan. |
| Future-state value | Carry a feature or facade forward only when it materially improves the intended future state. | The feature survives only because removal requires work. |
| Expansion safety | Avoid changes that make later expansion more complex, brittle, or difficult to maintain. | The change creates phase-shaped runtime code, shallow helpers, peer-internal imports, or per-file exception growth. |
| Operability | Favor designs that make the subsystem easier to reason about, test, extend, and maintain. | Ownership, public contract, failure behavior, or validation becomes less explicit. |

Default rules:

- Never delete a file solely because Fallow reports `unused-file`.
- Resolve entry-point and dynamic-use coverage before acting on cascading
  file/export findings.
- Prefer removing an unnecessary `export` modifier over deleting a symbol that
  remains useful inside its owner.
- Prefer one canonical definition over duplicate exports or compatibility
  barrels.
- Do not add inline Fallow suppressions by default. A retained exception needs
  a named owner, concrete continuing-value rationale, regression evidence, and
  documented configuration-level handling.
- Preserve public Make target identity, command IDs, artifact paths, summary
  shapes, and failure mapping unless the harness owner is changed first.
- Update authored owners before generated task-surface, schedule, topology, or
  contract outputs. Never hand-edit generated roots or dependency lockfiles.
- Keep phase maps and ledgers as evidence accounting; do not introduce phase
  identity into production or reusable harness module architecture.

## 3. Baseline Reconciliation and Finding Identity

### 3.1 Count reconciliation

| Finding family | Source collection | ID prefix | Count |
| --- | --- | --- | ---: |
| Unused files | `unused_files` | `UF` | 246 |
| Unused exports | `unused_exports` | `UE` | 1,359 |
| Unused types | `unused_types` | `UT` | 20 |
| Unused development dependencies | `unused_dev_dependencies` | `UDD` | 1 |
| Unused class members | `unused_class_members` | `UCM` | 4 |
| Unresolved imports | `unresolved_imports` | `URI` | 1 |
| Duplicate exports | `duplicate_exports` | `DE` | 5 |
| Circular dependencies | `circular_dependencies` | `CD` | 20 |
| **Total** |  |  | **1,656** |

The report summary rolls the single `unused_dev_dependencies` entry into its
displayed unused-dependency total. The root `unused_dependencies` collection is
empty; `UDD-0001` is the only dependency finding.

### 3.2 Stable identity rule

A finding ID is its category prefix plus the one-based array position in the
frozen JSON, zero-padded to four digits. Examples: the first unused file is
`UF-0001`, and the twentieth cycle is `CD-0020`. IDs remain attached to this
baseline even if later edits move line numbers. A later Fallow report receives
a new baseline label and new source-position IDs; old findings are reconciled
as `DONE`, `DROPPED`, `FALSE_POSITIVE_CONFIG`, or explicitly carried forward.

Each live tracker row uses these fields:

| Field | Meaning |
| --- | --- |
| Finding or batch ID | Stable source ID, inclusive ID range, or named batch whose membership is defined here. |
| Status | `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, or `DROPPED`. |
| Disposition | `REMOVE`, `INTERNALIZE`, `CONSOLIDATE`, `RESTRUCTURE`, `RETAIN_JUSTIFIED`, `FALSE_POSITIVE_CONFIG`, or `BLOCKED`. |
| Affected owner | Module, package, harness subsystem, configuration owner, or asset owner responsible for the resolution. |
| Structural target | Intended ownership or dependency shape after resolution. |
| Dependencies | Prior tracker items or evidence required before action. |
| Validation | Narrow public Make target and any required retained artifact. |
| Before/after counts | Baseline count and verified replacement-report count. |
| Evidence | Inspected callers, tests, manifests, run roots, diffs, or owner decisions. |
| Continuing-value justification | Required only for retained legacy, dynamic, reflective, or externally consumed surfaces. |

### 3.3 Reliability diagnosis

The baseline has a material reachability problem:

- 242 of 246 unused-file findings are under `tools/**` (`UF-0005` through
  `UF-0246`).
- 1,332 of 1,359 unused-export findings are under `tools/**` (`UE-0028`
  through `UE-1359`).
- All 1,332 tool export findings occur inside files that Fallow already marks
  unused, so they are cascading findings rather than independent removal
  evidence.
- The tool files are invoked through Make recipes, authored task metadata,
  manifests, CLI dispatch, subprocesses, and tests that the current Fallow
  entry configuration does not fully model.
- The current `.fallowrc.json` names nonexistent
  `apps/web/src/testSetup.ts` and `apps/web/src/testSetup.dom.ts` entries while
  Vitest uses `apps/web/src/testing/testSetup.ts` and
  `apps/web/src/testing/testSetup.dom.ts`.
- `apps/web/index.html` loads `/assets/fonts/fonts.css` at runtime, while the
  report labels the stylesheet unused and the root-relative URL unresolved.
- `@cyclonedx/cdxgen` is invoked through `pnpm exec cdxgen` in release-evidence
  tooling rather than a static module import.

Therefore `UF-0005..UF-0246` and `UE-0028..UE-1359` are quarantined from
deletion until DC-002 and DC-003 complete. A corrected rerun may prove some
members genuinely unused, but it must do so from accurate reachability.

DC-002 was implemented on 2026-07-13. The corrected canonical Fallow run at
`.cartulary/test-results/20260713T171113Z-p616212` uses the retained effective
configuration at
`.cartulary/test-results/20260713T171113Z-p616212/frontend-fallow-static/fallow/resolved-fallowrc.json`.
Its `dead-code.json` SHA-256 is
`bbf66bb19c207cf2c1b3dfab6744085f2c4c54f8130a8112da963b67954cc0f6`.
The corrected counts are 38 unused files, 316 unused exports, 20 unused types,
zero unused development dependencies, four unused class members, zero unresolved
imports, 19 duplicate exports, and 20 circular dependencies. Known reachability
false positives for the Vitest setup files, `/assets/fonts/fonts.css`, the
public font stylesheet, and `@cyclonedx/cdxgen` are absent. Tool findings fell
from 242 to 37 unused files and from 1,332 to 289 unused exports; DC-003 must
map or close the surviving tool findings before deletion authority is granted.

DC-003 was reconciled on 2026-07-13 with a fresh canonical
`make frontend-fallow-static` run at
`.cartulary/test-results/20260713T172736Z-p727319`. Its `dead-code.json`
SHA-256 is `5ab8001f3c90133071f2a69190886d9ba10ddac18eb5bac2de6cbe04f7022f5a`.
The corrected counts are unchanged from the DC-002 corrected baseline: 38
unused files, 316 unused exports, 20 unused types, zero unused development
dependencies, four unused class members, zero unresolved imports, 19 duplicate
exports, and 20 circular dependencies. The changed fingerprint is artifact
identity only; the issue set and counts remained stable. DC-003 closes the
reachability-limited frozen source report as deletion authority and maps the
surviving tool findings to corrected-run survivor cleanup.

## 4. Counted Tool Reachability Batches

These batches preserve auditable coverage of the 1,574 quarantined tool
file/export findings without copying the 1.1 MB ignored report into this
document. Batch membership is the source IDs in the stated aggregate range
whose `path` begins with the listed prefix.

| Batch | Path area | Unused files | Unused exports | Status | Initial disposition |
| --- | --- | ---: | ---: | --- | --- |
| TR-01 | `tools/harness/phase-accounting/**` | 53 | 320 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-02 | `tools/harness/scheduler/**` | 29 | 254 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports; five duplicate-export rows remain for DC-006 |
| TR-03 | `tools/harness/generated-artifacts/**` | 28 | 124 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-04 | `tools/harness/backend/**` | 27 | 129 | DONE | Tool cleanup final survivors: 1 retained owner-facade file, 10 retained dynamic-test exports |
| TR-05 | `tools/harness/output/**` | 22 | 80 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-06 | `tools/harness/diagnostics/**` | 15 | 50 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-07 | `tools/harness/execution/**` | 13 | 78 | DONE | Tool cleanup final survivors: 0 unused files, 2 retained dynamic-test exports |
| TR-08 | `tools/harness/contract/**` | 12 | 167 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-09 | Other `tools/**` paths outside the named harness areas | 9 | 40 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports; OTel provenance refreshed by `make generate` |
| TR-10 | `tools/harness/static-analysis/**` | 9 | 12 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-11 | `tools/harness/browser/**` | 7 | 31 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-12 | `tools/harness/duration-accounting/**` | 6 | 14 | DONE | Tool cleanup final survivors: 1 retained owner-facade file, 0 unused exports |
| TR-13 | `tools/harness/finalization/**` | 4 | 10 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-14 | `tools/harness/readiness/**` | 2 | 0 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-15 | `tools/harness/test-support/**` | 2 | 0 | DONE | Tool cleanup final survivors: 0 unused files, 0 unused exports |
| TR-16 | `tools/harness/command-surface/**` | 1 | 10 | DONE | Corrected survivors mapped: 0 unused files, 0 unused exports |
| TR-17 | `tools/harness/runtime-binary-registry.mjs` | 1 | 13 | DONE | Corrected survivors mapped: 0 unused files, 4 unused exports |
| TR-18 | `tools/harness/smoke/**` | 1 | 0 | DONE | Corrected survivors mapped: 0 unused files, 0 unused exports |
| TR-19 | `tools/harness/tests/**` | 1 | 0 | DONE | Corrected survivors mapped: 0 unused files, 0 unused exports |
| **Total** | `UF-0005..UF-0246`; `UE-0028..UE-1359` | **242** | **1,332** |  |  |

The initial disposition was a quarantine posture, not a final claim that every
member was a false positive. DC-003 replaced each batch count with corrected
survivor counts. The tool survivor cleanup workstream then modeled durable
harness entrypoints, deleted no-caller private helpers, internalized unused
tool exports, refreshed generated owners through Make, and reduced the final
tool survivor set to two retained owner-facade files and 12 retained dynamic
test-support exports. Five tool duplicate-export findings remain intentionally
for DC-006 duplicate consolidation.

## 5. Explicit Actionable Finding Inventory

All rows remain `TODO` until their required discovery and validation are
recorded. Line numbers identify the frozen report and may drift.

### 5.1 Application files

| ID | Path | Current evidence | Disposition | Status | Owner and exit condition |
| --- | --- | --- | --- | --- | --- |
| UF-0001 | `apps/web/public/assets/fonts/fonts.css` | Loaded by `apps/web/index.html`; checked by frontend font-bundle evidence. | FALSE_POSITIVE_CONFIG | DONE | Web asset/config owner; corrected reachability plus `make frontend-unit` keep the runtime asset path intact. |
| UF-0002 | `apps/web/src/app/LandingAdminSurface.tsx` | Compatibility re-export documented in `apps/web/src/README.md`; no current code import found. | REMOVE | DONE | App shell; file and README responsibility row deleted after caller search found no continuing consumer. |
| UF-0003 | `apps/web/src/testing/testSetup.dom.ts` | Selected by the `browser-unit` Vitest project. | FALSE_POSITIVE_CONFIG | DONE | Frontend test setup; corrected reachability keeps Vitest browser setup behavior unchanged. |
| UF-0004 | `apps/web/src/testing/testSetup.ts` | Selected by both Vitest projects. | FALSE_POSITIVE_CONFIG | DONE | Frontend test setup; corrected reachability keeps Vitest project setup behavior unchanged. |

### 5.2 Live-file unused exports

For these findings, first search direct, type-only, test, dynamic, and package
consumers. Use `INTERNALIZE` when the symbol remains locally useful and
`REMOVE` only when the symbol itself has no continuing value.

| ID | Path and line | Export | Workstream owner | Status |
| --- | --- | --- | --- | --- |
| UE-0001 | `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx:78` | `primaryToolbarButtonStyle` | Workbook toolbar | DONE_INTERNALIZED |
| UE-0002 | `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx:124` | `WorkbookToolbarSearchLabel` | Workbook toolbar | DONE_REMOVED |
| UE-0003 | `apps/web/src/workbook/models/entityWorkbookModel.ts:125` | `readEntityStringCell` | Entity workbook model | DONE_INTERNALIZED |
| UE-0004 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:1` | `evidenceRecordLifecycleStates` | Evidence lifecycle view model | DONE_INTERNALIZED |
| UE-0005 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:13` | `objectBlobUploadStates` | Evidence lifecycle view model | DONE_INTERNALIZED |
| UE-0006 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:22` | `evidenceLifecycleViewStateKeys` | Evidence lifecycle view model | DONE_INTERNALIZED |
| UE-0007 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:36` | `evidenceCountDisplayStateKeys` | Evidence lifecycle view model | DONE_INTERNALIZED |
| UE-0008 | `apps/web/src/workbook/models/workbookContractRows.ts:39` | `materializeWorkbookViewRow` | Workbook contract rows | DONE_INTERNALIZED |
| UE-0009 | `apps/web/src/workbook/models/workbookInspectorModel.ts:154` | `clearRowBoundInspectorState` | Workbook inspector | DONE_INTERNALIZED |
| UE-0010 | `apps/web/src/workbook/models/workbookMutations.ts:16` | `submitViewRecordPatch` | Workbook mutations | DONE_INTERNALIZED |
| UE-0011 | `apps/web/src/workbook/models/workbookSurfaceRegistry.ts:222` | `getWorkbookSurfaceRegistryEntry` | Workbook surface registry | DONE_REMOVED |
| UE-0012 | `apps/web/src/workbook/models/workbookSurfaceRegistry.ts:228` | `isBuiltInWorkbookSurfaceId` | Workbook surface registry | DONE_REMOVED |
| UE-0013 | `apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts:56` | `timelineTabClientInstanceId` | Timeline pending saves | DONE_INTERNALIZED |
| UE-0014 | `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts:33` | `clearAppliedTimelineFilterDraft` | Timeline runtime | DONE_INTERNALIZED |
| UE-0015 | `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts:43` | `applyTimelineFilterDraftToQuery` | Timeline runtime | DONE_INTERNALIZED |
| UE-0016 | `apps/web/src/workbook/timeline/models/workbookMentionChips.ts:11` | `mentionChipStates` | Timeline mention chips | DONE_REMOVED |
| UE-0017 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:16` | `clipboardTextLooksTabular` re-export | Timeline model | DONE_REMOVED_FACADE |
| UE-0018 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:454` | `readTimelineStringCell` | Timeline model | DONE_INTERNALIZED |
| UE-0019 | `apps/web/src/workbook/utils/workbookGridFocus.tsx:37` | `formatWorkbookFocusAnchor` | Workbook grid focus | DONE_INTERNALIZED |
| UE-0020 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:465` | `mergePendingCollectionActionPayload` | Pending queue | DONE_INTERNALIZED |
| UE-0021 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:485` | `mergePendingReplayPayload` | Pending queue | DONE_INTERNALIZED |
| UE-0022 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:544` | `canonicalizePendingReplayChanges` | Pending queue | DONE_INTERNALIZED |
| UE-0023 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:561` | `buildPendingReplayMutationIdentity` | Pending queue | DONE_INTERNALIZED |
| UE-0024 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:969` | `shouldRetryPendingFailure` | Pending queue | DONE_INTERNALIZED |
| UE-0025 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:997` | `canCoalescePendingReplayUnits` | Pending queue | DONE_INTERNALIZED |
| UE-0026 | `apps/web/src/workbook/utils/workbookStyles.ts:62` | `statusStripMutedItemStyle` | Workbook styles | DONE_INTERNALIZED |
| UE-0027 | `apps/web/src/workbook/utils/workbookStyles.ts:78` | `statusStripSpacerStyle` | Workbook styles | DONE_REMOVED |

### 5.3 Unused exported types

| ID | Path and line | Type | Workstream owner | Status |
| --- | --- | --- | --- | --- |
| UT-0001 | `apps/web/src/app/AccountAdministrationPanels.tsx:65` | `AccountSecurityPanelHandle` | App administration | DONE_INTERNALIZED |
| UT-0002 | `apps/web/src/app/AccountAdministrationPanels.tsx:70` | `DeploymentUsersPanelHandle` | App administration | DONE_INTERNALIZED |
| UT-0003 | `apps/web/src/app/AccountAdministrationPanels.tsx:80` | `DeploymentUsersPanelCommandState` | App administration | DONE_INTERNALIZED |
| UT-0004 | `apps/web/src/app/ReferencePackAdminPanel.tsx:54` | `ReferencePackAdminPanelHandle` | Reference-pack administration | DONE_INTERNALIZED |
| UT-0005 | `apps/web/src/app/api/appShellClient.ts:111` | `EnterpriseAuthBeginResponse` | App-shell client | DONE_INTERNALIZED |
| UT-0006 | `apps/web/src/app/api/appShellClient.ts:686` | `Phase1Response` | App-shell client | DONE_REMOVED |
| UT-0007 | `apps/web/src/services/browserApi.ts:23` | `PublicErrorDetail` | Browser API | DONE_REMOVED_FACADE |
| UT-0008 | `apps/web/src/services/browserApi.ts:23` | `PublicErrorView` | Browser API | DONE_REMOVED_FACADE |
| UT-0009 | `apps/web/src/workbook/WorkbookShell.tsx:88` | `WorkbookIncidentIdentity` | Workbook shell | DONE_REMOVED_FACADE |
| UT-0010 | `apps/web/src/workbook/WorkbookShell.tsx:91` | `WorkbookAccountModel` | Workbook shell | DONE_REMOVED_FACADE |
| UT-0011 | `apps/web/src/workbook/WorkbookShell.tsx:92` | `WorkbookIncidentControlsMenuItem` | Workbook shell | DONE_REMOVED_FACADE |
| UT-0012 | `apps/web/src/workbook/WorkbookShell.tsx:94` | `WorkbookIncidentRole` | Workbook shell | DONE_REMOVED_FACADE |
| UT-0013 | `apps/web/src/workbook/WorkbookShell.tsx:95` | `WorkbookIncidentSnapshot` | Workbook shell | DONE_REMOVED_FACADE |
| UT-0014 | `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts:75` | `WorkbookShellActiveQueryControls` | Workbook shell runtime | DONE_INTERNALIZED |
| UT-0015 | `apps/web/src/workbook/timeline/components/TimelineGridSurface.tsx:12` | `TimelineGridSurfaceProps` | Timeline grid | DONE_INTERNALIZED |
| UT-0016 | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx:186` | `SaveState` | Timeline workbook | DONE_REMOVED |
| UT-0017 | `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts:38` | `TimelineRowContextMenuState` | Timeline inspector | DONE_INTERNALIZED |
| UT-0018 | `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts:21` | `TimelineWorkbookRuntimeSaveState` | Timeline runtime | DONE_INTERNALIZED |
| UT-0019 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:18` | `WorkbookRecordFreshnessDecision` | Timeline model | DONE_REMOVED_FACADE |
| UT-0020 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:19` | `WorkbookVersionedRecord` | Timeline model | DONE_REMOVED_FACADE |

### 5.4 Unused class members

All four findings concern `WorkbookPendingQueueModel`. Characterize pending
dispatch, auth recovery, conflict clearing, retry, and replay behavior before
removing methods; do not retain public methods merely as speculative extension
points.

| ID | Path and line | Member | Proposed disposition | Status |
| --- | --- | --- | --- | --- |
| UCM-0001 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1207` | `settleDispatched` | RETAIN_ANALYZER_EXCEPTION | DONE |
| UCM-0002 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1289` | `resumeAfterAuthRecovery` | RETAIN_ANALYZER_EXCEPTION | DONE |
| UCM-0003 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1294` | `pauseForAuthRecovery` | RETAIN_ANALYZER_EXCEPTION | DONE |
| UCM-0004 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1300` | `clearSameFieldConflict` | RETAIN_ANALYZER_EXCEPTION | DONE |

### 5.5 Duplicate exports

Each pair must acquire one semantic owner. Do not suppress a duplicate simply
because both definitions currently have local callers. The five frozen duplicate
exports were superseded by 19 corrected-run duplicate exports after DC-002
reachability repair.

| ID | Export | Locations | Proposed disposition | Status |
| --- | --- | --- | --- | --- |
| DE-0001 | `PagingMeta` | `apps/web/src/app/api/appShellClient.ts:71`; `apps/web/src/app/landingAdminTypes.ts:26` | CONSOLIDATE under the app-shell paging contract owner | DONE_CONSOLIDATED |
| DE-0002 | `TimelineRowContextMenuPosition` | `apps/web/src/workbook/timeline/components/TimelineRowActions.tsx:19`; `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts:33` | CONSOLIDATE under Timeline interaction state | DONE_INTERNALIZED_HOOK_TYPE |
| DE-0003 | `defaultTaskSurfaceManifestPath` | `tools/harness/execution/summary-topology.mjs:14`; `tools/harness/generated-artifacts/index.mjs:3` | CONSOLIDATE under task-surface model/path owner | DONE_CONSOLIDATED |
| DE-0004 | `deferred` | `apps/web/src/testing/fetchMockTestSupport.ts:175`; `apps/web/src/testing/timelineWorkbookTestSupport.ts:128` | CONSOLIDATE in generic test support | DONE_CONSOLIDATED |
| DE-0005 | `executionDependencyMetadata` | `tools/harness/execution/execution-dependencies.mjs:11`; `tools/harness/generated-artifacts/index.mjs:17` | CONSOLIDATE under execution-topology generated-artifact owner with no broad-barrel cycle | DONE_CONSOLIDATED |
| DE-0006 | `fetchJSON` | `apps/web/src/services/browserApi.ts:51`; `apps/web/src/services/workbookApi.ts:66` | Rename the workbook-specific wrapper and keep browser transport generic | DONE_RENAMED |
| DE-0007 | `formatResourceMap` | `tools/harness/scheduler/scheduler-reporting.mjs:79`; `tools/harness/scheduler/scheduler-resources.mjs:538` | CONSOLIDATE under scheduler resource formatting owner | DONE_CONSOLIDATED |
| DE-0008 | `mapServiceBackedClaimsToCheckClaims` | `tools/harness/execution/service-backed/schedule-expansion.mjs:106`; `tools/harness/execution/service-backed/schedule-resource-claims.mjs:7`; `tools/harness/scheduler/scheduler-resource-policy.mjs:52` | CONSOLIDATE under scheduler resource policy owner | DONE_CONSOLIDATED |
| DE-0009 | `normalizePositiveInteger` | `tools/harness/backend/go-duration-artifacts.mjs:8`; `tools/harness/backend/go-duration-baselines.mjs:21` | CONSOLIDATE under backend duration parsing owner | DONE_CONSOLIDATED |
| DE-0010 | `relToRepo` | `tools/harness/contract/artifact-discovery.mjs:26`; `tools/harness/contract/repo-paths.mjs:7` | CONSOLIDATE under contract repo-path owner | DONE_CONSOLIDATED |
| DE-0011 | `repoRoot` | `tools/harness/contract/artifact-discovery.mjs:11`; `tools/harness/contract/harness-contract.mjs:26`; `tools/harness/contract/test-output-context.mjs:10`; `tools/harness/diagnostics/target-plan-rows.mjs:26`; `tools/harness/diagnostics/task-execution-map.mjs:36`; `tools/harness/execution/summary-topology.mjs:13`; `tools/harness/generated-artifacts/index.mjs:10` | CONSOLIDATE shared result-root use without reintroducing cycles; keep genuinely local constants private | DONE_CONSOLIDATED |
| DE-0012 | `resolveResultsRoot` | `tools/harness/contract/artifact-discovery.mjs:34`; `tools/harness/contract/test-output-context.mjs:47` | CONSOLIDATE under contract repo-path/results-root owner | DONE_CONSOLIDATED |
| DE-0013 | `resourceClaimsObject` | `tools/harness/execution/service-backed/schedule-utils.mjs:7`; `tools/harness/scheduler/scheduler-resource-policy.mjs:46` | CONSOLIDATE under scheduler resource policy owner | DONE_CONSOLIDATED |
| DE-0014 | `resourceLimitSummary` | `tools/harness/scheduler/scheduler-reporting.mjs:285`; `tools/harness/scheduler/scheduler-resources.mjs:546` | CONSOLIDATE under scheduler resource formatting owner | DONE_CONSOLIDATED |
| DE-0015 | `resourceMapToObject` | `tools/harness/scheduler/scheduler-reporting.mjs:83`; `tools/harness/scheduler/scheduler-resources.mjs:534` | CONSOLIDATE under scheduler resource formatting owner | DONE_CONSOLIDATED |
| DE-0016 | `schedulerManifestSchemaID` | `tools/harness/execution/summary-topology.mjs:21`; `tools/harness/scheduler/scheduler-manifest.mjs:33` | CONSOLIDATE under scheduler manifest owner | DONE_CONSOLIDATED |
| DE-0017 | `serviceBackedGoExecutionDependencies` | `tools/harness/execution/execution-dependencies.mjs:17`; `tools/harness/generated-artifacts/index.mjs:23` | CONSOLIDATE under execution-topology generated-artifact owner with no broad-barrel cycle | DONE_CONSOLIDATED |
| DE-0018 | `serviceBackedSupportTargets` | `tools/harness/execution/execution-dependencies.mjs:19`; `tools/harness/generated-artifacts/index.mjs:24` | CONSOLIDATE under execution-topology generated-artifact owner with no broad-barrel cycle | DONE_CONSOLIDATED |
| DE-0019 | `timelineViewSchemaId` | `apps/web/src/testing/timelineWorkbookTestSupport.ts:28`; `apps/web/src/workbook/models/workbookSurfaceRegistry.ts:7` | CONSOLIDATE on the production stable identifier | DONE_CONSOLIDATED |

### 5.6 Dependency and unresolved-import findings

| ID | Finding | Current evidence | Proposed disposition | Exit condition |
| --- | --- | --- | --- | --- |
| UDD-0001 | `@cyclonedx/cdxgen` at `package.json:16` | Invoked by `tools/release-evidence/generate-sbom-license-evidence.mjs` through `pnpm exec cdxgen`; release evidence records its version. | FALSE_POSITIVE_CONFIG or RETAIN_JUSTIFIED | Model executable dependency use without hand-editing `pnpm-lock.yaml`; release/SBOM validation remains intact. |
| URI-0001 | `/assets/fonts/fonts.css` from `apps/web/index.html` | Valid Vite/public-root URL for UF-0001. | FALSE_POSITIVE_CONFIG | Accurate asset resolution or a documented config-level exception; `make frontend-unit` passes. |

## 6. Circular Dependency Inventory

The 20 reported cycles form a coupled harness architecture problem across
backend planning, phase accounting, diagnostics, execution, and generated
artifact rendering. Resolve the dependency directions, not the enumerated
cycles one at a time. Extract consumer-neutral data readers, selectors, and
render inputs into leaf modules; keep orchestration and CLI modules above those
leaves; avoid importing broad `index.mjs` barrels from reusable internals.

| ID | Length | Frozen cycle path | Status |
| --- | ---: | --- | --- |
| CD-0001 | 6 | `backend/backend-shard-plan.mjs` -> `backend/go-shard-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | DONE |
| CD-0002 | 5 | `backend/backend-shard-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | DONE |
| CD-0003 | 5 | `backend/backend-target-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | DONE |
| CD-0004 | 6 | `backend/backend-target-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | DONE |
| CD-0005 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/frontend-phase-manifest.mjs` -> `phase-accounting/frontend/phase-artifacts.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | DONE |
| CD-0006 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `execution/execution-dependencies.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | DONE |
| CD-0007 | 4 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | DONE |
| CD-0008 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | DONE |
| CD-0009 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/browser-work-units.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | DONE |
| CD-0010 | 5 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | DONE |
| CD-0011 | 5 | `diagnostics/task-execution-map.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | DONE |
| CD-0012 | 5 | `diagnostics/task-guidance.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` | DONE |
| CD-0013 | 5 | `diagnostics/task-guidance.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/browser-work-units.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` | DONE |
| CD-0014 | 4 | `diagnostics/task-guidance.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` | DONE |
| CD-0015 | 3 | `execution/execution-dependencies.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | DONE |
| CD-0016 | 5 | `execution/execution-dependencies.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` | DONE |
| CD-0017 | 4 | `execution/summary-topology.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/task-surface/index.mjs` -> `generated-artifacts/task-surface/validation.mjs` | DONE |
| CD-0018 | 5 | `execution/summary-topology.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/task-surface/index.mjs` -> `generated-artifacts/task-surface/validation.mjs` -> `generated-artifacts/task-surface/recipe-validation.mjs` | DONE |
| CD-0019 | 5 | `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/frontend-phase-manifest.mjs` -> `phase-accounting/frontend/phase-artifacts.mjs` | DONE |
| CD-0020 | 5 | `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/browser-work-units.mjs` | DONE |

All paths in this table are relative to `tools/harness/`. The cycle exit
condition is a corrected Fallow run with zero cycle findings, unchanged public
Make behavior, passing harness contract evidence, and no replacement cycle
hidden by a suppression or broad barrel.

## 7. Dependency-Ordered Work Tracker

| ID | Work item | Status | Depends on | Affected owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| DC-001 | Freeze source identity, counts, authority, and command posture. | DONE | none | Tracker | Sections 1-3 | Fingerprint and all 1,656 findings reconcile. |
| DC-002 | Repair Fallow reachability for Vitest setup, Make/manifest CLIs, browser assets, and executable dependencies. | DONE | DC-001 | `.fallowrc.json`, harness static analysis | Config diff; caller and manifest inventory; retained `frontend-fallow-static` run root `.cartulary/test-results/20260713T171113Z-p616212`; `make harness-contract` passed | Reachability represents durable owner patterns without per-file suppression growth. |
| DC-003 | Run canonical Fallow and reconcile the quarantined tool batches. | DONE | DC-002 | Fallow harness | Corrected Fallow fingerprint `5ab8001f3c90133071f2a69190886d9ba10ddac18eb5bac2de6cbe04f7022f5a`; retained run root `.cartulary/test-results/20260713T172736Z-p727319`; surviving tool counts 37 unused files and 289 unused exports | Every old ID/batch is closed or mapped to a corrected survivor; no count silently disappeared. |
| DC-004 | Resolve four application file findings and remove the obsolete compatibility barrel if no continuing consumer exists. | DONE | DC-003 | Web app, assets, frontend tests | Caller search found only tracker/README references; deleted `apps/web/src/app/LandingAdminSurface.tsx`; removed the README responsibility row; `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, and `make frontend-fallow-static` passed | Runtime/test-loaded files are modeled accurately; valueless barrel is removed. |
| DC-005 | Internalize or remove the 27 live exports and 20 exported types in cohesive owner slices. | DONE | DC-003 | App shell and workbook owners | Caller searches; export/type internalization and no-caller deletion; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make frontend-fallow-static` at `.cartulary/test-results/20260713T180901Z-p787033` | Public surface is smaller without product behavior drift or speculative compatibility exports. |
| DC-005A | Reconcile pending-queue class-member findings without removing active runtime API. | DONE | DC-005 | Pending queue | Caller evidence in runtime hooks and `workbookPendingQueue.test.ts`; `make frontend-unit`; `make frontend-fallow-static` at `.cartulary/test-results/20260713T181225Z-p790676` | UCM-0001 through UCM-0004 are retained with concrete caller evidence and no inline suppression. |
| DC-006 | Consolidate corrected duplicate exports under semantic owners. | DONE | DC-003 | App shell, Timeline, services, test support, harness owners | Caller migration; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make lint-scripts`; `make harness-contract`; `make frontend-fallow-static` at `.cartulary/test-results/20260713T182922Z-p811034` | Corrected Fallow reports zero duplicate exports. |
| DC-007 | Break the harness cycle family through leaf dependencies and narrow interfaces. | DONE | DC-001 | Harness backend, phase accounting, diagnostics, execution, generated artifacts | Narrowed the phase-accounting facade so it no longer re-exports phase-slice planning; switched reusable execution/diagnostics/frontend artifact modules from broad generated-artifact barrels to leaf owners; `make lint-scripts` passed; final `make harness-contract` passed; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T173735Z-p734881` with zero circular dependencies | CD-0001 through CD-0020 closed structurally and future phase growth remains declarative. |
| DC-007A | Remediate corrected tool survivor files and exports. | DONE | DC-007 | Harness, generated artifacts, OTel tooling | Reachability owner extension; no-caller helper deletion; export internalization; generated refresh; retained Fallow run `.cartulary/test-results/20260713T175435Z-p761422` | Tool unused files reduced from 37 to two justified owner facades, tool unused exports reduced from 290 to 12 justified dynamic test APIs, and tool duplicate exports are limited to DC-006 rows. |
| DC-008 | Close the font import and `cdxgen` dependency findings. | DONE | DC-002, DC-003 | Web assets and release evidence | Corrected Fallow run `.cartulary/test-results/20260713T172736Z-p727319` has zero unused development dependencies and zero unresolved imports; font stylesheet, Vitest setup files, and `@cyclonedx/cdxgen` are absent from corrected false-positive findings | Dynamic uses are represented accurately; no valuable dependency or asset is removed. |
| DC-009 | Review every retained exception and reject unjustified legacy or compatibility burden. | DONE | DC-004 through DC-008, DC-005A, DC-006 | All affected owners | Reachability-owner handling for retained harness entrypoints and dynamic shell-test exports; pending-queue caller evidence; final Fallow run `.cartulary/test-results/20260713T184446Z-p832672` | No inline suppression or undocumented indefinite compatibility path remains. |
| DC-010 | Complete final verification, accounting, and handoff. | DONE | DC-003 through DC-009 | Repository-wide | `make agent-finalize`; broad `make check`; retained-run finalizer; final Fallow run `.cartulary/test-results/20260713T185740Z-p1096336`; final tracker handoff | Findings reconcile, required checks pass, residual risk and next action are explicit. |

DC-004 through DC-006 may be split into independent owner-shaped patches after
DC-003. Do not combine analyzer configuration, harness cycle restructuring,
frontend behavior changes, and dependency removal in one review unless the
tracker records why they are inseparable.

## 8. Checkpoint and Validation Plan

Use only public Make targets from the repository root. Choose the narrowest
target that proves each change, then broaden according to risk.

| Checkpoint | Edit scope | Required validation | Required record |
| --- | --- | --- | --- |
| CP-01 | Fallow entry/reachability configuration | `make frontend-fallow-static`; `make harness-contract` if harness contract or task metadata changes; `make json-shape-check` for JSON owner inputs | New report fingerprint, retained run root, before/after counts, and per-batch reconciliation. |
| CP-02 | Harness cycle structure | `make lint-scripts`; `make harness-contract`; `make frontend-fallow-static` | Dependency-direction explanation, CD closure map, exact results. |
| CP-03 | Frontend exports, types, class members, and compatibility barrel | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; affected browser/phase slice only when observable behavior is at risk | Symbol disposition, characterization evidence, exact results. |
| CP-04 | Font asset classification | `make frontend-unit`; `make frontend-fallow-static` | Asset caller evidence and URI/UF closure. |
| CP-05 | Executable dependency classification | Relevant release/SBOM Make target discovered through `make task-guide` or `make explain-target`; `make frontend-fallow-static` | Process invocation evidence and UDD closure. |
| CP-06 | Authored/generated owner changes | Relevant generator; `make generated-artifact-policy-check`; `make generate-drift`; `make json-shape-check` | Owner-first diff and proof that generated files were not hand-edited. |
| CP-07 | Final handoff | `make agent-finalize`; then `make check` | Exact run roots, failures and resolutions, skipped checks with reasons. |

A Fallow count reduction is diagnostic evidence only. TypeScript, Biome,
frontend import boundaries, tests, security checks, generated-artifact drift,
and harness gates retain their separate authority. Fallow output must not be
represented as product conformance or Core 05 publication evidence.

Failure handling:

- Record the failing target and retained run root or summary artifact.
- State whether the failure is related to the current slice.
- Do not replace a failed result silently with a later passing run.
- If accurate reachability increases counts, record the increase as improved
  analysis coverage rather than a regression in cleanup.
- If a behavior-preserving cleanup exposes an owner-document mismatch, stop
  that slice and separate the behavior decision from dead-code remediation.

## 9. Workstream Notes and Exception Ledger

### 9.1 Evidence log

| Date | Work item | Source or command | Result | Impact and next action |
| --- | --- | --- | --- | --- |
| 2026-07-12 | DC-001 | Frozen JSON inspection, repository searches, `.fallowrc.json`, Vite config, task metadata, harness wrapper, domain and harness owner docs | Baseline and reachability diagnosis recorded; no implementation changed. | Begin DC-002 only under later authorization. |
| 2026-07-13 | DC-002 | Harness NLSpec, `.fallowrc.json`, `tools/fallow/reachability_owner.json`, Fallow reachability builder, Fallow wrapper, task-surface owner/manifest, JSON-shape validation, harness static-analysis tests | Owner-driven reachability implemented; corrected Fallow run recorded; Vitest setup, Vite public assets, task-surface CLIs, and `@cyclonedx/cdxgen` executable dependency are no longer false positive findings. | Begin DC-003 batch reconciliation using `.cartulary/test-results/20260713T171113Z-p616212/frontend-fallow-static/fallow/dead-code.json`. |
| 2026-07-13 | DC-003 | `make frontend-fallow-static`; corrected report inspection; duplicate-export inventory extraction; tool-batch survivor count reconciliation | Fresh corrected Fallow run passed at `.cartulary/test-results/20260713T172736Z-p727319` with SHA-256 `5ab8001f3c90133071f2a69190886d9ba10ddac18eb5bac2de6cbe04f7022f5a` and unchanged corrected counts. TR-01 through TR-19 are reconciled to corrected survivor counts; corrected duplicate-export inventory now has 19 rows. | Begin DC-007 harness cycle architecture before acting on survivor deletion/internalization. |
| 2026-07-13 | DC-007 | Import graph inspection; `make lint-scripts`; `make harness-contract`; `make frontend-fallow-static` | Phase-accounting `index.mjs` no longer re-exports phase-slice planning; execution, diagnostics, frontend artifact, and browser work-unit modules import generated-artifact leaves instead of broad barrels. An intermediate `make harness-contract` failed at `.cartulary/test-results/20260713T173538Z-p732436` because direct private phase-accounting imports violated the harness boundary; fixed by preserving the narrowed facade. Final Fallow run `.cartulary/test-results/20260713T173735Z-p734881` reports zero circular dependencies and SHA-256 `57ad9fc65b9a2ff6853ecdca686ad555a12f748982d7187d8a99cf7d905115a1`. | Begin tool survivor cleanup using the zero-cycle corrected report. |
| 2026-07-13 | DC-007A | Tool survivor report inspection; reachability-owner update; no-caller file deletion; mechanical export internalization; dynamic caller search; `make phase-schedules`; `make lint-scripts`; `make harness-contract`; `make run-harness-smoke-fast`; `make json-shape-check`; `make generated-artifact-policy-check`; `make generate`; `make generate-drift`; `make phase-schedule-drift`; `make otel-conformance`; `make frontend-fallow-static` | Durable harness entrypoints are modeled without inline suppressions. Four no-caller helper files plus the later exposed service-backed topology CLI were removed. Tool exports were internalized except shell-test dynamic APIs. Final Fallow run `.cartulary/test-results/20260713T175435Z-p761422` has SHA-256 `b9c3dd70bc34663512d979fd1464c24ccc2cd822e7005f31c8dbf72b2ffec721`, three unused files total, 39 unused exports total, 20 unused types, zero unused dependencies/imports, four unused class members, 10 duplicate exports, and zero cycles. | Begin DC-004 web/app file cleanup; handle five remaining tool duplicate exports in DC-006. |
| 2026-07-13 | DC-004 | Caller search for `LandingAdminSurface`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make frontend-fallow-static`; Fallow report extraction | `apps/web/src/app/LandingAdminSurface.tsx` had no consumer beyond the README responsibility row and tracker. The compatibility barrel and README row were removed. `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T180110Z-p774821` with SHA-256 `70786b20e4bea3e3f4093f844383baa4aa1e2a6d523ff0d9cb3d3d5cd8fa3faa`, two unused files, 39 unused exports, 20 unused types, zero unused dependencies/imports, four unused class members, 10 duplicate exports, and zero cycles. | Begin DC-005 frontend export/type/member cleanup; DC-006 duplicate consolidation still includes five frontend rows and five tool rows. |
| 2026-07-13 | DC-005 | Caller searches for 27 frontend exports and 20 exported types; focused export/type patch; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make format`; `make lint-biome`; `make frontend-fallow-static` | Frontend unused exports were internalized or removed; unused exported types were internalized, deleted, or removed from compatibility facades. `TimelineRowContextMenuPosition` duplicate was closed by internalizing the hook-local type. Final Fallow run `.cartulary/test-results/20260713T180901Z-p787033` has SHA-256 `13e027bc08cfe91bd5a4129a17ac033d87681384e859f0a80ca7a7a0df9b8a25`, two unused files, 12 unused exports, zero unused types, zero unused dependencies/imports, four unused class members, nine duplicate exports, and zero cycles. | Begin DC-005A pending-queue class-member reconciliation. |
| 2026-07-13 | DC-005A | Caller search for `settleDispatched`, `resumeAfterAuthRecovery`, `pauseForAuthRecovery`, and `clearSameFieldConflict`; Fallow config check; `make frontend-unit`; `make frontend-fallow-static` | All four methods have current runtime or unit-test callers: pending replay settlement, auth recovery pause/resume, and same-field conflict clearing. No class-member-specific owner config exists, and inline suppression is rejected. Fallow run `.cartulary/test-results/20260713T181225Z-p790676` has SHA-256 `38924db13c683fd1deaf52a9d9b94bf8984c9e07f1d2f1326b88007de3462a3f`, two unused files, 12 unused exports, zero unused types, zero unused dependencies/imports, four retained class-member findings, nine duplicate exports, and zero cycles. | Begin DC-006 duplicate-export consolidation. |
| 2026-07-13 | DC-006 | Duplicate-export caller migration; canonical owner import updates; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make lint-scripts`; `make harness-contract`; `make format`; `make frontend-fallow-static` | Duplicate frontend and harness exports were consolidated or renamed under semantic owners: app-shell paging, generic test `deferred`, production `timelineViewSchemaId`, workbook transport wrapper, scheduler resource formatting/policy, task-surface path, contract result-root/repo helpers, and execution dependency metadata. `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T182922Z-p811034` with SHA-256 `521717f5a73308570615270e4a210ceba350f32d3c82296d8d94cbc520ac0795`, two retained unused files, 12 retained unused exports, zero unused types, zero unused dependencies/imports, four retained class-member findings, zero duplicate exports, and zero cycles. | Begin DC-009 retained-exception audit; residual Fallow issues are only documented retained-exception candidates. |
| 2026-07-13 | DC-009 | Retained exception audit; reachability-owner schema/builder/test extension; dynamic-export owner records; facade entrypoint records; pending-queue caller review; `make run-harness-smoke-fast`; `make lint-scripts`; `make harness-contract`; `make frontend-unit`; `make phase-schedules`; `make json-shape-check`; `make phase-schedule-drift`; `make generated-artifact-policy-check`; `make frontend-fallow-static` | Two retained tool owner facades are now modeled as harness entrypoints, and 12 shell-embedded ESM test APIs are modeled through `harness_dynamic_exports` with named owners, evidence, and removal triggers. Pending-queue class members remain the only Fallow-reported exceptions because the current config has no safe class/member-scoped reachability rule; global `usedClassMembers` handling was rejected as too broad. Final Fallow run `.cartulary/test-results/20260713T184446Z-p832672` has SHA-256 `13b37af88c58781a27441befb3db610d81a6685a9d5256f1f1a5fd94c402d4d1`, zero unused files, zero unused exports, zero unused types, zero unused dependencies/imports, four retained class-member findings, zero duplicate exports, and zero cycles. | Begin DC-010 final validation and handoff. |
| 2026-07-13 | DC-010 | Final narrow validation; `make agent-finalize`; three broad `make check` retained runs until warm-readiness evidence passed; retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T185515Z-p1014134`; generated drift/schema checks; final `make frontend-fallow-static` | Final retained-run finalizer passed at `.cartulary/test-results/20260713T185653Z-p1094087` after warming `build-server-harness` and selecting the successful warm check run `.cartulary/test-results/20260713T185515Z-p1014134`. Final Fallow run `.cartulary/test-results/20260713T185740Z-p1096336` has SHA-256 `7983f643e3605913e89310328659267254136c535004292ee9425175f34a7512`, zero unused files, zero unused exports, zero unused types, zero unused dependencies/imports, four retained class-member findings, zero duplicate exports, and zero cycles. | Cleanup complete; no next remediation workstream remains. |

### 9.2 Retained-exception ledger

Add a row before assigning `RETAIN_JUSTIFIED`. An empty ledger is preferred.

| Finding ID | Owner | Continuing current value | Why structural removal is worse | Regression evidence | Config-level handling | Removal trigger | Review date |
| --- | --- | --- | --- | --- | --- | --- | --- |
| TR-04 retained file | Harness backend duration accounting | `tools/harness/backend/backend-duration-accounting.mjs` remains the named `backend_duration_accounting` owner facade in `tools/harness_helper_ownership.json` and the harness NLSpec helper ownership table. | Removing the facade would make helper ownership drift from the current harness boundary contract while backend duration ownership remains documented. | `make harness-contract` at `.cartulary/test-results/20260713T184352Z-p828407`; `make json-shape-check` at `.cartulary/test-results/20260713T184446Z-p832648`; final Fallow run `.cartulary/test-results/20260713T184446Z-p832672`. | Modeled in `tools/fallow/reachability_owner.json` under `harness_entrypoints.files`; no inline suppression. | Remove when `backend_duration_accounting` is removed from helper ownership/NLSpec or a narrower active owner replaces the facade. | 2026-07-13 |
| TR-12 retained file | Harness duration accounting | `tools/harness/duration-accounting/index.mjs` remains the named `scheduler_duration_accounting` owner facade in `tools/harness_helper_ownership.json` and the harness NLSpec helper ownership table. | Removing the facade would discard the current scheduler duration accounting boundary while duration-baseline ownership is still documented through that facade. | `make harness-contract` at `.cartulary/test-results/20260713T184352Z-p828407`; `make json-shape-check` at `.cartulary/test-results/20260713T184446Z-p832648`; final Fallow run `.cartulary/test-results/20260713T184446Z-p832672`. | Modeled in `tools/fallow/reachability_owner.json` under `harness_entrypoints.files`; no inline suppression. | Remove when `scheduler_duration_accounting` is removed from helper ownership/NLSpec or a narrower active owner replaces the facade. | 2026-07-13 |
| TR-04 retained exports | Backend target execution smoke tests | Seven `tools/harness/backend/backend-target-execution.mjs` facade exports and three leaf exports in `tools/harness/backend/target-execution/capture.mjs` are imported by shell-embedded ESM tests in `tools/harness/backend/tests/test-run-go-target.sh`. | Internalizing these exports breaks current regression tests for shared report capture, locking, reuse, aggregate report accounting, and execution-family assignment. | Exact caller search; `make run-harness-smoke-fast` at `.cartulary/test-results/20260713T184342Z-p827371`; `make harness-contract` at `.cartulary/test-results/20260713T184352Z-p828407`; final Fallow run `.cartulary/test-results/20260713T184446Z-p832672`. | Modeled in `tools/fallow/reachability_owner.json` under `harness_dynamic_exports`; no inline suppression. | Remove when the shell tests stop importing these helpers or move to static module tests modeled by Fallow. | 2026-07-13 |
| TR-07 retained exports | Service-backed schedule source validation | `validateServiceBackedScheduleManifestShape` is exported through `tools/harness/execution/service-backed/schedule-planning.mjs` and its leaf for the shell-embedded execution-topology smoke test. | Internalizing it breaks current generated-artifact regression checks that reject malformed service-backed source manifests and camelCase browser aliases. | Exact caller search in `tools/harness/generated-artifacts/tests/test-execution-topology.sh`; `make run-harness-smoke-fast` at `.cartulary/test-results/20260713T184342Z-p827371`; `make harness-contract` at `.cartulary/test-results/20260713T184352Z-p828407`; final Fallow run `.cartulary/test-results/20260713T184446Z-p832672`. | Modeled in `tools/fallow/reachability_owner.json` under `harness_dynamic_exports`; no inline suppression. | Remove when the generated-artifact smoke test moves to static imports or no longer needs this validation helper. | 2026-07-13 |
| UCM-0001 through UCM-0004 | Pending queue runtime model | `settleDispatched`, `resumeAfterAuthRecovery`, `pauseForAuthRecovery`, and `clearSameFieldConflict` are called by Timeline pending replay/live update/conflict hooks and by pending-queue unit tests. | Removing them breaks pending replay settlement, auth recovery, and same-field conflict resolution; replacing them with analyzer indirection would add code without improving ownership. | Caller search in `apps/web/src/workbook/timeline/**` and `apps/web/src/workbook/utils/workbookPendingQueue.test.ts`; `make frontend-unit` at `.cartulary/test-results/20260713T184416Z-p830006`; final Fallow run `.cartulary/test-results/20260713T184446Z-p832672`. | No inline suppression. Config-level `usedClassMembers` was rejected because it is global or heritage-scoped, not file/class-specific, and would hide future unrelated methods with the same names. | Remove or refactor only when runtime hooks no longer call the methods or Fallow gains owner-level class-member reachability modeling. | 2026-07-13 |

An exception is invalid when it cites only possible future use, historical use,
compatibility in the abstract, or the cost of cleanup. Invalid exceptions must
be removed or marked `BLOCKED` pending a named owner decision.

### 9.3 Blockers and risks

| ID | Risk or blocker | Blocking work | Resolution condition |
| --- | --- | --- | --- |
| DR-001 | Frozen Fallow entry coverage did not represent Make/manifest-driven tool execution. | Tool file/export cleanup | Resolved for current tool cleanup by DC-002, DC-003, and DC-007A: durable harness entrypoints are owner-modeled, no-caller helpers were deleted, and remaining tool findings are documented retained exceptions or DC-006 duplicates. |
| DR-002 | Removing pending-queue methods without characterization could change retry, conflict, or auth-recovery behavior. | UCM-0001 through UCM-0004 | Relevant frontend unit/behavior coverage is inspected and passing. |
| DR-003 | Breaking cycles through a new shared barrel could reproduce the same coupling under another path. | DC-007 | Resolved by DC-007: the phase-accounting facade is narrowed, generated-artifact leaf imports replace broad barrel imports in reusable modules, and Fallow reports zero cycles. |
| DR-004 | Compatibility barrels and duplicate types can obscure real ownership. | DE-0001 through DE-0019 | Resolved by DC-004, DC-005, and DC-006: the obsolete app compatibility file was removed, hook-local duplicate types were internalized, and corrected Fallow now reports zero duplicate exports. |
| DR-005 | Direct dependency and asset use is invisible to ordinary import analysis. | UDD-0001 and URI-0001 | Process and browser-asset entry semantics are modeled and tested. |

## 10. Session Handoff Template

Append one record before ending each remediation session.

| Field | Value |
| --- | --- |
| Date/time | TODO |
| Branch/commit | TODO |
| Authorized work items | TODO |
| Finding IDs or batches touched | TODO |
| Dispositions changed | TODO |
| Files inspected | TODO |
| Files changed | TODO |
| Commands run | TODO |
| Passing validation and run roots | TODO |
| Failing validation and run roots | TODO |
| Fallow before/after fingerprint and counts | TODO |
| Generated outputs | TODO |
| Decisions and continuing-value evidence | TODO |
| Open blockers and residual risks | TODO |
| Next work item | TODO |
| Safe restart command | TODO |

Do not claim a command passed unless it ran in the recorded session or an exact
retained artifact is named. Do not claim a caller, dynamic use, generated owner,
or behavior was preserved unless it was inspected. Use `TODO` for missing
evidence rather than guessing.

## 11. Binary Completion Criteria

The cleanup program is complete only when all of the following are true:

- [x] A corrected canonical Fallow baseline supersedes the frozen reachability-
  limited report and records its fingerprint and retained run root.
- [x] All 1,656 source findings are closed individually or through the defined
  auditable batches; totals reconcile without silent disappearance.
- [x] UF-0005 through UF-0246 and UE-0028 through UE-1359 were not used as
  deletion authority before DC-002 and DC-003 completed.
- [x] Every surviving genuine unused file, export, type, member, dependency,
  duplicate, unresolved import, and cycle has a completed disposition.
- [x] The seven binding cleanup gates in Section 2 are satisfied for every
  implementation slice.
- [x] No legacy facade, compatibility alias, dual path, or feature is retained
  without clear continuing value and an owner.
- [x] No inline suppression, per-file configuration sprawl, generated-file hand
  edit, lockfile hand edit, or phase-shaped runtime dependency was introduced.
- [x] Harness cycles are removed structurally and later phase/target growth
  remains declarative.
- [x] Public Make identities, command IDs, artifacts, summaries, and failure
  mapping remain stable unless their owner document was changed first.
- [x] Narrow validation, `make agent-finalize`, final broad verification, and
  final Fallow results are recorded with exact outcomes and skipped-check
  reasons.
- [x] The last handoff record lets another engineer resume without rediscovering
  source posture, decisions, counts, blockers, or commands.

## 12. Tracker-Creation Handoff

| Field | Value |
| --- | --- |
| Date | 2026-07-12 |
| Scope | Documentation-only creation of this tracker. |
| Files inspected | Frozen Fallow JSON; modular-refactor framework; domain vocabulary; testing-harness NLSpec; Fallow config/wrapper; task metadata; Vite setup; application compatibility barrel; font asset and HTML entry; release-evidence `cdxgen` use. |
| Files changed | `docs/handoffs/dead-code-clean-up-tracker.md` only. |
| Substantive result | All 1,656 findings reconcile; noisy tool findings are quarantined in counted batches; actionable findings and cycles are explicitly inventoried; dependency-ordered work and validation are defined. |
| Implementation performed | None. |
| Validation | `make lint-markdown` passed; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260713T005154Z-p69886`; `make json-shape-check` passed at `.cartulary/test-results/20260713T005154Z-p69879`; `git diff --check` passed. |
| Skipped checks | `make check` and `make agent-finalize` were not run because this was a documentation-only tracker change and no broader verification was needed. Retained-run maintenance was not requested. |
| Next authorized work | DC-002, analyzer reachability repair. |

## 13. DC-002 Implementation Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 13:17:13 EDT |
| Branch/commit | `main` at `12abcf036ea900e7cae75cbef9ef236fefdffc00` with uncommitted DC-002 changes |
| Authorized work items | DC-002 plus handoff needed to start DC-003 |
| Finding IDs or batches touched | UF-0001, UF-0003, UF-0004, UDD-0001, URI-0001, TR-01 through TR-19 |
| Dispositions changed | DC-002 `DONE`; DC-003 `TODO`; tool batches unblocked for corrected-run reconciliation |
| Files inspected | Harness NLSpec; `.fallowrc.json`; Vite/Vitest config and setup files; web public asset path; task-surface owner/manifest; Fallow wrapper and summary schema; release-evidence `cdxgen` caller; Fallow frozen and corrected reports |
| Files changed | `.fallowrc.json`; `docs/testing-harness-nlspec.md`; `docs/handoffs/dead-code-clean-up-tracker.md`; `tools/fallow/reachability_owner.json`; `tools/schemas/cartulary.fallow_reachability_owner.v1.schema.json`; `tools/harness/static-analysis/fallow-reachability.mjs`; `tools/harness/static-analysis/fallow-static-cli.mjs`; `tools/harness/static-analysis/tests/test-fallow-static.sh`; `tools/harness/generated-artifacts/check-json-shapes.mjs`; `tools/harness_schema_attachments.json`; `tools/task_surface_owner.json`; generated task-surface/topology/scheduler and duration-baseline artifacts refreshed by Make/finalizer |
| Commands run | `make phase-schedules`; `make json-shape-check`; `make generated-artifact-policy-check`; `make lint-scripts`; `make generate-drift`; `make phase-schedule-drift`; `make harness-contract`; `make lint-markdown`; `make frontend-unit`; `make frontend-fallow-static`; `make sbom`; `make agent-finalize`; `make check`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T171331Z-p635102`; `git diff --check` |
| Passing validation and run roots | Final `make json-shape-check` passed at `.cartulary/test-results/20260713T171612Z-p719261`; final `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260713T171612Z-p719276`; final `make generate-drift` passed at `.cartulary/test-results/20260713T171612Z-p719258`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260713T171054Z-p613781` and again inside finalizer `.cartulary/test-results/20260713T171520Z-p716915`; `make frontend-unit` passed at `.cartulary/test-results/20260713T171113Z-p616178`; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T171113Z-p616212`; `make sbom` passed at `.cartulary/test-results/20260713T171303Z-p633465`; `make check` passed at `.cartulary/test-results/20260713T171331Z-p635102`; final `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T171331Z-p635102` passed at `.cartulary/test-results/20260713T171520Z-p716915` |
| Failing validation and run roots | An intermediate `make frontend-fallow-static` run failed at `.cartulary/test-results/20260713T170635Z-p607006` due to Fallow summary artifact-ref schema shape; fixed by emitting summary-local Fallow artifact refs while preserving tool-run artifact refs |
| Fallow before/after fingerprint and counts | Frozen SHA-256 `bbfb9b7354e1fea0fac4d91b813dfd483051e2ffad0475c62eeb9e061bb9da68`: 246 unused files, 1,359 unused exports, 20 unused types, one unused dev dependency, four unused class members, one unresolved import, five duplicate exports, 20 cycles. Corrected SHA-256 `bbf66bb19c207cf2c1b3dfab6744085f2c4c54f8130a8112da963b67954cc0f6`: 38 unused files, 316 unused exports, 20 unused types, zero unused dev dependencies, four unused class members, zero unresolved imports, 19 duplicate exports, 20 cycles. |
| Generated outputs | `tools/task_surface_manifest.json`, `tools/execution_topology_render_index.json`, `tools/scheduler_manifest.json`, `tools/browser_e2e_duration_baselines.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, and `tools/service_backed_make_target_duration_baselines.json` were refreshed by Make/finalizer targets. |
| Decisions and continuing-value evidence | Reachability is modeled from durable owners, not per-file suppressions: Vitest setup from owner config, task-surface scripts from `tools/task_surface_owner.json`, Vite `/assets/**` public URLs from the declared public root, and `@cyclonedx/cdxgen` from release-evidence executable dependency ownership. `frontend-fallow-static` remains non-blocking, static-only, and Fallow Runtime remains disabled. |
| Open blockers and residual risks | DC-003 still needs to map or close surviving corrected-run tool findings; duplicate-export count increased under the corrected analyzer and must be handled in later cleanup; circular dependencies remain unchanged. |
| Next work item | DC-003, corrected-run reconciliation of TR-01 through TR-19 and surviving non-tool findings. |
| Safe restart command | `make frontend-fallow-static && make harness-contract` |

## 14. DC-003 Reconciliation Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 13:28:14 EDT |
| Branch/commit | `main` at `994711db` with DC-003 tracker changes |
| Authorized work items | DC-003 corrected baseline reconciliation |
| Finding IDs or batches touched | TR-01 through TR-19, UF-0001, UF-0003, UF-0004, UDD-0001, URI-0001, DE-0001 through DE-0019 |
| Dispositions changed | DC-003 `DONE`; DC-004, DC-005, and DC-006 unblocked to `TODO`; DC-008 `DONE` because corrected reachability closes the font/import/dependency false positives |
| Files inspected | `docs/handoffs/dead-code-clean-up-tracker.md`; corrected Fallow report at `.cartulary/test-results/20260713T172736Z-p727319/frontend-fallow-static/fallow/dead-code.json` |
| Files changed | `docs/handoffs/dead-code-clean-up-tracker.md` |
| Commands run | `make frontend-fallow-static`; `sha256sum .cartulary/test-results/20260713T172736Z-p727319/frontend-fallow-static/fallow/dead-code.json`; `jq` report summaries and duplicate-export extraction |
| Passing validation and run roots | `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T172736Z-p727319` |
| Failing validation and run roots | none |
| Fallow before/after fingerprint and counts | DC-002 corrected SHA-256 `bbf66bb19c207cf2c1b3dfab6744085f2c4c54f8130a8112da963b67954cc0f6`: 38 unused files, 316 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 19 duplicate exports, 20 cycles. DC-003 SHA-256 `5ab8001f3c90133071f2a69190886d9ba10ddac18eb5bac2de6cbe04f7022f5a`: same counts. |
| Generated outputs | none |
| Decisions and continuing-value evidence | The corrected Fallow issue set is stable across two canonical runs. Frozen `UF-0005..UF-0246` and `UE-0028..UE-1359` are no longer deletion authority; survivor cleanup must use the corrected report and owner evidence. |
| Open blockers and residual risks | Harness cycles remain; survivor tool findings must be classified after cycle architecture cleanup. |
| Next work item | DC-007, harness cycle architecture |
| Safe restart command | `make frontend-fallow-static && make harness-contract` |

## 15. DC-007 Harness Cycle Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 13:37:35 EDT |
| Branch/commit | `main` at `994711db` with DC-003/DC-007 changes |
| Authorized work items | DC-007 harness cycle architecture |
| Finding IDs or batches touched | CD-0001 through CD-0020 |
| Dispositions changed | CD-0001 through CD-0020 `DONE`; DC-007 `DONE` |
| Files inspected | Harness cycle report; phase-accounting facade; backend target planning; execution dependency, summary topology, task guidance, task execution map, frontend phase artifacts, browser work-unit, and generated-artifact renderer modules; harness import-boundary test output |
| Files changed | `tools/harness/phase-accounting/index.mjs`; `tools/harness/phase-accounting/phase-slice-planning/browser-work-units.mjs`; `tools/harness/execution/execution-dependencies.mjs`; `tools/harness/execution/summary-topology.mjs`; `tools/harness/phase-accounting/frontend/phase-artifacts.mjs`; `tools/harness/diagnostics/task-execution-map.mjs`; `tools/harness/diagnostics/task-guidance.mjs`; `tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs`; `docs/handoffs/dead-code-clean-up-tracker.md` |
| Commands run | `make lint-scripts`; `make harness-contract`; `make frontend-fallow-static`; `sha256sum .cartulary/test-results/20260713T173735Z-p734881/frontend-fallow-static/fallow/dead-code.json`; `jq` report summaries |
| Passing validation and run roots | `make lint-scripts` passed; final `make harness-contract` passed; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T173735Z-p734881` |
| Failing validation and run roots | Intermediate `make harness-contract` failed at `.cartulary/test-results/20260713T173538Z-p732436` because direct private phase-accounting leaf imports violated the harness import boundary; fixed by narrowing the facade and restoring non-owner imports through it |
| Fallow before/after fingerprint and counts | DC-003 SHA-256 `5ab8001f3c90133071f2a69190886d9ba10ddac18eb5bac2de6cbe04f7022f5a`: 38 unused files, 316 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 19 duplicate exports, 20 cycles. DC-007 SHA-256 `57ad9fc65b9a2ff6853ecdca686ad555a12f748982d7187d8a99cf7d905115a1`: 38 unused files, 324 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 19 duplicate exports, zero cycles. The unused-export increase is improved analysis coverage after facade narrowing. |
| Generated outputs | none |
| Decisions and continuing-value evidence | The general phase-accounting facade remains the approved boundary for manifest, registry, subsystem, and frontend phase contracts, but it no longer re-exports phase-slice planning. Phase-slice planning remains available through `tools/harness/phase-accounting/phase-slice-plan.mjs`, its classified facade. Reusable diagnostics/execution modules import generated-artifact leaf owners instead of `generated-artifacts/index.mjs`. |
| Open blockers and residual risks | Tool survivor findings increased to 324 total unused exports because facade narrowing exposed additional unused re-exports; survivor cleanup must classify the zero-cycle report. |
| Next work item | Tool survivor cleanup |
| Safe restart command | `make lint-scripts && make harness-contract && make frontend-fallow-static` |

## 16. Tool Survivor Cleanup Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 17:56:19 EDT |
| Branch/commit | `main` at `994711db` with DC-003/DC-007/tool-survivor changes |
| Authorized work items | Tool survivor cleanup after DC-007 |
| Finding IDs or batches touched | TR-01 through TR-19 corrected survivors; tool duplicate rows DE-0003, DE-0005, DE-0007 through DE-0018; OTel tool export survivors |
| Dispositions changed | DC-007A `DONE`; TR batches final-counted; retained exception ledger populated for two owner-facade files and 12 dynamic test-support exports |
| Files inspected | Corrected zero-cycle Fallow reports; `.fallowrc.json`; `tools/fallow/reachability_owner.json`; reachability owner schema and builder; task-surface owner; helper ownership manifest; Make surface; shell-embedded ESM tests for backend target execution and generated execution topology; generated drift and OTel contract outputs |
| Files changed | Reachability owner/schema/builder/test; generated phase render index; OTel source snapshot; many `tools/harness/**` modules had unused export modifiers or obsolete re-export specifiers removed; deleted no-caller files `tools/harness/contract/schema-ids.mjs`, `tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership-cli.mjs`, `tools/harness/phase-accounting/phase-run-verification.mjs`, `tools/harness/scheduler/check-schedule-manifest.mjs`, and `tools/harness/execution/service-backed/schedule-topology.mjs`; tracker updated |
| Commands run | `make json-shape-check`; `make lint-scripts`; `make phase-schedules`; `make frontend-fallow-static`; `make harness-contract`; `make phase-schedule-drift`; `make run-harness-smoke-fast`; `make generated-artifact-policy-check`; `make generate`; `make generate-drift`; `make otel-conformance`; `make help-all`; `make explain-target TARGET=run-harness-smoke-fast DETAIL=summary`; `make explain-target TARGET=run-harness-smoke-full DETAIL=summary`; `jq`, `rg`, `sha256sum`, and targeted report extraction commands |
| Passing validation and run roots | Final `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T175435Z-p761422`; `make run-harness-smoke-fast` passed at `.cartulary/test-results/20260713T175435Z-p761334`; `make json-shape-check` passed at `.cartulary/test-results/20260713T175435Z-p761278`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260713T175435Z-p761314`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260713T175536Z-p763818`; `make generate` passed at `.cartulary/test-results/20260713T175555Z-p765202`; `make generate-drift` passed at `.cartulary/test-results/20260713T175602Z-p766037`; `make otel-conformance` passed at `.cartulary/test-results/20260713T175619Z-p767215`; final `make lint-scripts` and `make harness-contract` passed |
| Failing validation and run roots | `make json-shape-check` failed at `.cartulary/test-results/20260713T174530Z-p740998` because prior DC-007 phase-schedule inputs were stale; fixed by `make phase-schedules`. `make json-shape-check` failed at `.cartulary/test-results/20260713T175412Z-p759623` and `make phase-schedule-drift` failed at `.cartulary/test-results/20260713T175412Z-p759690` after later barrel edits; fixed by `make phase-schedules`. `make generate-drift` failed at `.cartulary/test-results/20260713T175536Z-p763786` because OTel generator-source provenance drifted; fixed by `make generate` and a passing `make generate-drift`. `make harness-smoke-execution-topology` and `make harness-smoke-run-go-target` failed immediately because those are not declared public Make targets; the public aggregate `make run-harness-smoke-fast` was used instead. |
| Fallow before/after fingerprint and counts | DC-007 SHA-256 `57ad9fc65b9a2ff6853ecdca686ad555a12f748982d7187d8a99cf7d905115a1`: 38 unused files, 324 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 19 duplicate exports, zero cycles. Final tool-cleanup SHA-256 `b9c3dd70bc34663512d979fd1464c24ccc2cd822e7005f31c8dbf72b2ffec721`: three unused files, 39 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 10 duplicate exports, zero cycles. |
| Generated outputs | `tools/execution_topology_render_index.json` refreshed through `make phase-schedules`; `contracts/otel/otel_source_snapshot.v1.json` refreshed through `make generate` to update `semconv_generated_constants.generator_source_sha` for the changed OTel generator source. |
| Decisions and continuing-value evidence | Durable harness entrypoint files are modeled in the reachability owner instead of suppressed. No-caller private helpers were deleted. Reported tool exports were internalized unless caller search found current shell-embedded ESM tests; those dynamic APIs are retained with concrete owner evidence and removal triggers. `tools/harness/backend/backend-duration-accounting.mjs` and `tools/harness/duration-accounting/index.mjs` are retained only as documented helper ownership facades. |
| Open blockers and residual risks | Five tool duplicate-export rows remain for DC-006; web/app file cleanup remains; frontend exports/types/class-member work remains. Fallow still cannot model shell-embedded ESM imports, so the retained exception ledger must be reviewed in DC-009. |
| Next work item | DC-004, web file and asset findings |
| Safe restart command | `make frontend-fallow-static && make frontend-typecheck && make frontend-unit && make frontend-import-boundary-check` |

## 17. DC-004 Web/App File Cleanup Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 18:01:16 EDT |
| Branch/commit | `main` at `83bbe898` with DC-004 and prior cleanup changes |
| Authorized work items | DC-004 web/app file and asset findings |
| Finding IDs or batches touched | UF-0001 through UF-0004; URI-0001 retained from corrected reachability |
| Dispositions changed | DC-004 `DONE`; UF-0001, UF-0003, and UF-0004 `DONE/FALSE_POSITIVE_CONFIG`; UF-0002 `DONE/REMOVE`; DR-004 no longer blocks on `UF-0002` |
| Files inspected | `apps/web/src/app/LandingAdminSurface.tsx`; `apps/web/src/README.md`; repository caller search for `LandingAdminSurface`; corrected Fallow reports |
| Files changed | Deleted `apps/web/src/app/LandingAdminSurface.tsx`; removed its exhaustive responsibility-map row from `apps/web/src/README.md`; updated this tracker |
| Commands run | `rg -n "LandingAdminSurface"`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make frontend-fallow-static`; `sha256sum .cartulary/test-results/20260713T180110Z-p774821/frontend-fallow-static/fallow/dead-code.json`; `jq` report summaries |
| Passing validation and run roots | `make frontend-typecheck` passed; `make frontend-unit` passed at `.cartulary/test-results/20260713T180046Z-p772683`; `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260713T180105Z-p774442`; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T180110Z-p774821` |
| Failing validation and run roots | none |
| Fallow before/after fingerprint and counts | Tool-cleanup SHA-256 `b9c3dd70bc34663512d979fd1464c24ccc2cd822e7005f31c8dbf72b2ffec721`: three unused files, 39 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 10 duplicate exports, zero cycles. DC-004 SHA-256 `70786b20e4bea3e3f4093f844383baa4aa1e2a6d523ff0d9cb3d3d5cd8fa3faa`: two unused files, 39 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 10 duplicate exports, zero cycles. |
| Generated outputs | none |
| Decisions and continuing-value evidence | Font CSS and Vitest setup files remain because corrected reachability models current runtime/test owners. The landing/admin compatibility barrel had no current consumer and only preserved a dual import path, so it was deleted rather than carried forward. |
| Open blockers and residual risks | Two retained tool facade files, 39 frontend/tool unused exports, 20 unused types, four pending-queue class-member findings, and 10 duplicate-export rows remain for DC-005, DC-006, and DC-009. |
| Next work item | DC-005, frontend export/type/member cleanup |
| Safe restart command | `make frontend-typecheck && make frontend-unit && make frontend-import-boundary-check && make frontend-fallow-static` |

## 18. DC-005 Frontend Export/Type Cleanup Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 18:09:01 EDT |
| Branch/commit | `main` at `83bbe898` with DC-005 and prior cleanup changes |
| Authorized work items | DC-005 frontend export/type cleanup |
| Finding IDs or batches touched | UE-0001 through UE-0027; UT-0001 through UT-0020; DE-0002 as a natural side effect of internalizing the hook-local type |
| Dispositions changed | DC-005 `DONE`; UE and UT rows `DONE_INTERNALIZED`, `DONE_REMOVED`, or `DONE_REMOVED_FACADE`; DE-0002 `DONE_INTERNALIZED_HOOK_TYPE`; DC-005A created for pending-queue class-member reconciliation |
| Files inspected | Latest Fallow reports; owner files for app administration, app-shell client, browser/workbook services, workbook shell/runtime, toolbar, workbook models, timeline runtime/model components, styles, grid focus, and pending queue; caller searches for reported symbols |
| Files changed | Frontend owner modules under `apps/web/src/app`, `apps/web/src/services`, and `apps/web/src/workbook`; this tracker |
| Commands run | `jq` Fallow extraction; `rg` caller searches; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make format`; `make frontend-fallow-static`; `sha256sum .cartulary/test-results/20260713T180901Z-p787033/frontend-fallow-static/fallow/dead-code.json` |
| Passing validation and run roots | Final `make frontend-typecheck` passed; `make frontend-unit` passed at `.cartulary/test-results/20260713T180701Z-p780375`; `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260713T180738Z-p782188`; `make format` passed at `.cartulary/test-results/20260713T180750Z-p783019`; final `make lint-biome` passed; final `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T180901Z-p787033` |
| Failing validation and run roots | Intermediate `make frontend-typecheck` failed at `.cartulary/test-results/20260713T180623Z-p779360` because the initial cleanup left an unused imported type, an unused built-in-surface set, and an unused style constant; fixed by removing those leftovers. Intermediate `make lint-biome` failed at `.cartulary/test-results/20260713T180738Z-p782219` for formatting in three touched files; fixed by `make format`. |
| Fallow before/after fingerprint and counts | DC-004 SHA-256 `70786b20e4bea3e3f4093f844383baa4aa1e2a6d523ff0d9cb3d3d5cd8fa3faa`: two unused files, 39 unused exports, 20 unused types, zero unused development dependencies, four unused class members, zero unresolved imports, 10 duplicate exports, zero cycles. DC-005 SHA-256 `13e027bc08cfe91bd5a4129a17ac033d87681384e859f0a80ca7a7a0df9b8a25`: two unused files, 12 unused exports, zero unused types, zero unused development dependencies, four unused class members, zero unresolved imports, nine duplicate exports, zero cycles. |
| Generated outputs | none |
| Decisions and continuing-value evidence | Locally used helpers and types were internalized instead of deleted. No-caller helpers/types and compatibility-facade re-exports were removed where canonical owners remain. Runtime behavior is unchanged; no public package API is removed. |
| Open blockers and residual risks | Remaining unused exports are the 12 retained tool dynamic APIs from DC-007A. Pending-queue class-member findings remain for DC-005A. Nine duplicate-export rows remain for DC-006. Two retained tool facade files remain for DC-009 exception audit. |
| Next work item | DC-005A, pending-queue class-member reconciliation |
| Safe restart command | `make frontend-typecheck && make frontend-unit && make frontend-import-boundary-check && make lint-biome && make frontend-fallow-static` |

## 19. DC-005A Pending Queue Class-Member Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 18:12:31 EDT |
| Branch/commit | `main` at `83bbe898` with DC-005A and prior cleanup changes |
| Authorized work items | DC-005A pending-queue class-member reconciliation |
| Finding IDs or batches touched | UCM-0001 through UCM-0004 |
| Dispositions changed | DC-005A `DONE`; UCM-0001 through UCM-0004 `DONE/RETAIN_ANALYZER_EXCEPTION`; exception ledger row added |
| Files inspected | `.fallowrc.json`; Fallow reachability wrapper; `apps/web/src/workbook/utils/workbookPendingQueue.ts`; pending-queue tests; Timeline live update, pending replay, and conflict resolver hooks |
| Files changed | `docs/handoffs/dead-code-clean-up-tracker.md` only |
| Commands run | `rg` caller searches; Fallow config searches; `make frontend-unit`; `make frontend-fallow-static`; `sha256sum .cartulary/test-results/20260713T181225Z-p790676/frontend-fallow-static/fallow/dead-code.json`; `jq` report summaries |
| Passing validation and run roots | `make frontend-unit` passed at `.cartulary/test-results/20260713T181225Z-p790648`; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T181225Z-p790676` |
| Failing validation and run roots | none |
| Fallow before/after fingerprint and counts | DC-005 SHA-256 `13e027bc08cfe91bd5a4129a17ac033d87681384e859f0a80ca7a7a0df9b8a25`: two unused files, 12 unused exports, zero unused types, zero unused development dependencies, four unused class members, zero unresolved imports, nine duplicate exports, zero cycles. DC-005A SHA-256 `38924db13c683fd1deaf52a9d9b94bf8984c9e07f1d2f1326b88007de3462a3f`: same counts. |
| Generated outputs | none |
| Decisions and continuing-value evidence | The four methods are active pending-queue API with current runtime callers and unit coverage. Fallow currently offers no owner-level class-member reachability model in this repo; inline suppressions were not added. |
| Open blockers and residual risks | Nine duplicate-export rows remain for DC-006. Two retained tool facade files, 12 retained tool dynamic exports, and four retained pending-queue class-member findings remain for DC-009 exception audit. |
| Next work item | DC-006, duplicate-export consolidation |
| Safe restart command | `make frontend-unit && make frontend-fallow-static` |

## 20. DC-006 Duplicate Export Consolidation Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 18:29:22 EDT |
| Branch/commit | `main` at `83bbe898` with DC-006 and prior cleanup changes |
| Authorized work items | DC-006 duplicate-export consolidation |
| Finding IDs or batches touched | DE-0001 through DE-0019 |
| Dispositions changed | DC-006 `DONE`; DE-0001, DE-0003 through DE-0019 `DONE_CONSOLIDATED` or `DONE_RENAMED`; DE-0002 remained `DONE_INTERNALIZED_HOOK_TYPE`; DC-009 unblocked to `TODO` |
| Files inspected | Corrected Fallow reports; app-shell paging types; Timeline test support and workbook tests; browser and workbook API services; scheduler resource/policy modules; service-backed schedule planning modules; harness summary topology, generated-artifact, contract, and output helpers |
| Files changed | `apps/web/src/app/App.tsx`; `apps/web/src/app/App.test.tsx`; `apps/web/src/app/landingAdminTypes.ts`; `apps/web/src/services/workbookApi.ts`; `apps/web/src/services/workbookApi.test.ts`; affected workbook phase tests and test support under `apps/web/src/workbook` and `apps/web/src/testing`; service/workbook API callers under `apps/web/src/networkFlow` and workbook hooks/components; `tools/harness/execution/summary-topology.mjs`; `tools/harness/execution/service-backed/schedule-browser-planning.mjs`; `tools/harness/execution/service-backed/schedule-expansion.mjs`; `tools/harness/execution/service-backed/schedule-resource-claims.mjs`; scheduler resource/reporting callers; `tools/harness/contract/test-output-context.mjs`; affected test-output helpers; this tracker |
| Commands run | `rg` caller searches; `jq` Fallow extraction; `sha256sum .cartulary/test-results/20260713T182922Z-p811034/frontend-fallow-static/fallow/dead-code.json`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make lint-scripts`; `make harness-contract`; `make format`; `make frontend-fallow-static` |
| Passing validation and run roots | Final `make frontend-typecheck` passed; `make frontend-unit` passed at `.cartulary/test-results/20260713T182820Z-p803616`; `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260713T182820Z-p803677`; `make format` passed at `.cartulary/test-results/20260713T182856Z-p807005`; final `make lint-biome` passed; final `make lint-scripts` passed; final `make harness-contract` passed; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T182922Z-p811034` |
| Failing validation and run roots | Intermediate `make frontend-typecheck` failed at `.cartulary/test-results/20260713T182256Z-p800232` and `.cartulary/test-results/20260713T182722Z-p802414` due the initial mechanical Timeline support import migration removing required surface-id arguments or adding duplicate imports; fixed by importing the production `timelineViewSchemaId` owner and restoring helper call arguments. Intermediate `make lint-biome` failed at `.cartulary/test-results/20260713T182820Z-p803630` for import ordering/formatting; fixed by `make format`. Intermediate `make harness-contract` failed at `.cartulary/test-results/20260713T182820Z-p803703` because `schedule-expansion.mjs` still imported the service-backed claim mapper from the removed facade; fixed by importing the scheduler policy owner. |
| Fallow before/after fingerprint and counts | DC-005A SHA-256 `38924db13c683fd1deaf52a9d9b94bf8984c9e07f1d2f1326b88007de3462a3f`: two unused files, 12 unused exports, zero unused types, zero unused development dependencies, four unused class members, zero unresolved imports, nine duplicate exports, zero cycles. DC-006 SHA-256 `521717f5a73308570615270e4a210ceba350f32d3c82296d8d94cbc520ac0795`: two unused files, 12 unused exports, zero unused types, zero unused development dependencies, four unused class members, zero unresolved imports, zero duplicate exports, zero cycles. |
| Generated outputs | none |
| Decisions and continuing-value evidence | Duplicate names now have one semantic owner or an unambiguous renamed wrapper. App paging uses the app-shell API type; Timeline tests import the production surface id; generic test deferral lives in fetch mock support; workbook transport uses `fetchWorkbookJSON`; scheduler resource formatting and policy own their helper names; service-backed expansion imports policy leaves directly; task-surface and contract/result-root helpers use their leaf owners without broad barrel cycles. |
| Open blockers and residual risks | Remaining Fallow findings are retained-exception candidates only: two tool owner-facade files, 12 dynamic shell-test tool exports, and four pending-queue class-member analyzer findings. They must be audited in DC-009 before final validation. |
| Next work item | DC-009, retained-exception audit |
| Safe restart command | `make frontend-typecheck && make frontend-unit && make frontend-import-boundary-check && make lint-biome && make lint-scripts && make harness-contract && make frontend-fallow-static` |

## 21. DC-009 Retained Exception Audit Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 18:44:52 EDT |
| Branch/commit | `main` at `83bbe898` with DC-009 and prior cleanup changes |
| Authorized work items | DC-009 retained-exception audit |
| Finding IDs or batches touched | TR-04 retained file/exports, TR-07 retained exports, TR-12 retained file, UCM-0001 through UCM-0004 |
| Dispositions changed | DC-009 `DONE`; DC-010 unblocked to `TODO`; retained tool owner facades moved from reported Fallow findings to modeled harness entrypoints; retained dynamic tool exports moved from reported Fallow findings to modeled dynamic-export owner records; pending-queue methods retained as documented analyzer exceptions |
| Files inspected | DC-006 Fallow report; `tools/fallow/reachability_owner.json`; reachability owner schema and builder; static-analysis fixture; backend and generated-artifact shell-embedded ESM tests; scheduler reporting and same-run helper artifact code; pending-queue runtime caller evidence |
| Files changed | `tools/schemas/cartulary.fallow_reachability_owner.v1.schema.json`; `tools/harness/static-analysis/fallow-reachability.mjs`; `tools/fallow/reachability_owner.json`; `tools/harness/static-analysis/tests/test-fallow-static.sh`; `tools/harness/scheduler/scheduler-reporting.mjs`; `tools/harness/output/test-output/run-summary.mjs`; generated `tools/execution_topology_render_index.json`; this tracker |
| Commands run | `make lint-scripts`; `make harness-contract`; `make run-harness-smoke-fast`; `make frontend-unit`; `make phase-schedules`; `make json-shape-check`; `make phase-schedule-drift`; `make generated-artifact-policy-check`; `make frontend-fallow-static`; `make lint-markdown`; direct smoke repros for `tools/harness/scheduler/tests/test-check-scheduler.sh smoke` and `tools/harness/scheduler/tests/test-service-backed-scheduler.sh smoke`; `make explain-target`; `jq`, `rg`, `sha256sum`, and targeted stack-trace reproduction commands |
| Passing validation and run roots | `make run-harness-smoke-fast` passed at `.cartulary/test-results/20260713T184342Z-p827371`; `make lint-scripts` passed at `.cartulary/test-results/20260713T184352Z-p828332`; `make harness-contract` passed at `.cartulary/test-results/20260713T184352Z-p828407`; `make frontend-unit` passed at `.cartulary/test-results/20260713T184416Z-p830006`; `make phase-schedules` passed at `.cartulary/test-results/20260713T184436Z-p832421`; `make json-shape-check` passed at `.cartulary/test-results/20260713T184446Z-p832648`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260713T184446Z-p832652`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260713T184446Z-p832709`; `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T184446Z-p832672`; `make lint-markdown` passed at `.cartulary/test-results/20260713T184821Z-p836906` |
| Failing validation and run roots | `make json-shape-check` failed at `.cartulary/test-results/20260713T183623Z-p815927` because phase-schedule inputs were stale after earlier harness edits; fixed by `make phase-schedules`. `make run-harness-smoke-fast` failed at `.cartulary/test-results/20260713T183733Z-p821800` and `.cartulary/test-results/20260713T183847Z-p823429`; fixed by using the consolidated scheduler resource formatter in `scheduler-reporting.mjs` and by passing `repoRoot` to same-run helper artifact collation in `run-summary.mjs`. `make json-shape-check` failed at `.cartulary/test-results/20260713T184416Z-p830016` and `make phase-schedule-drift` failed at `.cartulary/test-results/20260713T184416Z-p830002` after the scheduler-reporting fix; fixed by `make phase-schedules`. |
| Fallow before/after fingerprint and counts | DC-006 SHA-256 `521717f5a73308570615270e4a210ceba350f32d3c82296d8d94cbc520ac0795`: two unused files, 12 unused exports, zero unused types, zero unused development dependencies, four unused class members, zero unresolved imports, zero duplicate exports, zero cycles. DC-009 SHA-256 `13b37af88c58781a27441befb3db610d81a6685a9d5256f1f1a5fd94c402d4d1`: zero unused files, zero unused exports, zero unused types, zero unused development dependencies, four retained class-member findings, zero unresolved imports, zero duplicate exports, zero cycles. Resolved Fallow config SHA-256 `034d5e29ef3804c3bf41ecead4b157cc73ccaac3aa6b716947bed547c21f575a`; reachability stats: 71 task-surface entrypoints, 16 harness entrypoints, four dynamic-export files, 12 dynamic exports, two Vitest setup files, one Vite public asset, and one executable tooling dependency. |
| Generated outputs | `tools/execution_topology_render_index.json` refreshed through `make phase-schedules` because `tools/harness/scheduler/scheduler-reporting.mjs` is a phase-schedule input. No generated roots or lockfiles were hand-edited. |
| Decisions and continuing-value evidence | Retained exceptions based only on legacy age, possible future use, or cleanup cost were rejected. Tool owner facades remain only because helper ownership still names those facades. Shell-test exports remain only because current smoke tests import them through embedded ESM that static import analysis does not see. Pending-queue methods remain because active Timeline hooks and unit tests call them; broad Fallow `usedClassMembers` handling was rejected because it would suppress unrelated future methods with the same names. |
| Open blockers and residual risks | The only Fallow dead-code findings left are UCM-0001 through UCM-0004, documented as pending-queue analyzer exceptions with current caller evidence and no inline suppression. DC-010 still must run finalization, broad verification, retained-run finalization, `git diff --check`, and the final tracker handoff. |
| Next work item | DC-010, final validation and handoff |
| Safe restart command | `make lint-scripts && make harness-contract && make run-harness-smoke-fast && make frontend-unit && make json-shape-check && make phase-schedule-drift && make generated-artifact-policy-check && make frontend-fallow-static` |

## 22. DC-010 Final Validation and Handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-13 18:57:40 EDT |
| Branch/commit | `main` at `83bbe898` with the complete dead-code cleanup remediation changes |
| Authorized work items | DC-010 final validation and handoff |
| Finding IDs or batches touched | All frozen findings and corrected-run findings through the controlling batches and rows; residual analyzer exceptions UCM-0001 through UCM-0004 remain documented |
| Dispositions changed | DC-010 `DONE`; all binary completion criteria checked; no unhandled Fallow dead-code, duplicate-export, cycle, dependency, unresolved-import, or unused-file/export/type finding remains |
| Files inspected | Finalizer summaries; successful `make check` run roots; failed warm-readiness finalizer artifacts; final Fallow reports; generated/duration artifact diffs; tracker completion criteria |
| Files changed | All implementation, test, harness, generated, and documentation files listed in the prior workstream handoffs; finalizer additionally refreshed `tools/browser_e2e_duration_baselines.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`, and `tools/service_backed_make_target_duration_baselines.json`; tracker updated with final handoff |
| Commands run | Final narrow `make lint-markdown`; final narrow `make frontend-fallow-static`; `make agent-finalize`; `make check` at `.cartulary/test-results/20260713T184931Z-p842240`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T184931Z-p842240`; `make check` at `.cartulary/test-results/20260713T185202Z-p926598`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T185202Z-p926598`; `make build-server-harness`; `make check` at `.cartulary/test-results/20260713T185515Z-p1014134`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T185515Z-p1014134`; post-finalizer `make frontend-fallow-static`; `make generated-artifact-policy-check`; `make json-shape-check`; `make phase-schedule-drift`; post-handoff `make lint-markdown`; `git diff --check`; `jq`, `sha256sum`, `git status`, and generated artifact inspection commands |
| Passing validation and run roots | Final narrow `make lint-markdown` passed at `.cartulary/test-results/20260713T184904Z-p838678`; final narrow `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T184904Z-p838700`; initial `make agent-finalize` passed at `.cartulary/test-results/20260713T184919Z-p840776`; broad `make check` passed at `.cartulary/test-results/20260713T184931Z-p842240`; broad `make check` passed at `.cartulary/test-results/20260713T185202Z-p926598`; `make build-server-harness` passed at `.cartulary/test-results/20260713T185502Z-p1007716`; final broad `make check` passed at `.cartulary/test-results/20260713T185515Z-p1014134` with 1,116 tests and zero failures; retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T185515Z-p1014134` passed at `.cartulary/test-results/20260713T185653Z-p1094087`; post-finalizer `make frontend-fallow-static` passed at `.cartulary/test-results/20260713T185740Z-p1096336`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260713T185740Z-p1096313`; `make json-shape-check` passed at `.cartulary/test-results/20260713T185740Z-p1096344`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260713T185740Z-p1096392`; post-handoff `make lint-markdown` passed at `.cartulary/test-results/20260713T190022Z-p1099647`; `git diff --check` passed |
| Failing validation and run roots | Retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T184931Z-p842240` failed at `.cartulary/test-results/20260713T185118Z-p925811` because `build-server-harness` readiness duration was 22,237 ms against a 15,000 ms warm threshold. Retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T185202Z-p926598` failed at `.cartulary/test-results/20260713T185349Z-p1007119` because `build-server-harness` readiness duration was 15,954 ms against the same threshold. No product or test failure occurred in either retained run; the failures were warm-run eligibility rejections. `make build-server-harness` warmed the artifact, and the final retained run recorded 13,440 ms within threshold. |
| Fallow before/after fingerprint and counts | DC-009 SHA-256 `13b37af88c58781a27441befb3db610d81a6685a9d5256f1f1a5fd94c402d4d1`: zero unused files, zero unused exports, zero unused types, zero unused development dependencies, four retained class-member findings, zero unresolved imports, zero duplicate exports, zero cycles. Final SHA-256 `7983f643e3605913e89310328659267254136c535004292ee9425175f34a7512`: same counts. Final resolved Fallow config SHA-256 `034d5e29ef3804c3bf41ecead4b157cc73ccaac3aa6b716947bed547c21f575a`; reachability stats: 71 task-surface entrypoints, 16 harness entrypoints, four dynamic-export files, 12 dynamic exports, two Vitest setup files, one Vite public asset, and one executable tooling dependency. |
| Generated outputs | DC-010 retained-run finalizer refreshed `tools/browser_e2e_duration_baselines.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`, and `tools/service_backed_make_target_duration_baselines.json`. Prior workstreams refreshed `tools/execution_topology_render_index.json` and `contracts/otel/otel_source_snapshot.v1.json`. All generated updates were produced through Make/finalizer targets; generated roots and lockfiles were not hand-edited. |
| Decisions and continuing-value evidence | Warm-run failures were resolved by obtaining a valid warm retained run instead of weakening scheduler readiness policy. The final residual Fallow issue set is intentionally limited to active pending-queue methods with caller evidence and removal triggers. Retained harness facades and dynamic test exports are represented through owner config instead of inline suppressions. |
| Open blockers and residual risks | No cleanup blocker remains. Residual risk is limited to the documented Fallow class-member reachability limitation for UCM-0001 through UCM-0004; deleting those methods would break current pending replay, auth recovery, and conflict-resolution behavior. |
| Next work item | None for this cleanup plan |
| Safe restart command | `make frontend-fallow-static && make check && make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T185515Z-p1014134 && git diff --check` |
