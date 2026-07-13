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
| Current status | DC-002 implemented; corrected Fallow baseline recorded; DC-003 unblocked for batch reconciliation. |

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

## 4. Counted Tool Reachability Batches

These batches preserve auditable coverage of the 1,574 quarantined tool
file/export findings without copying the 1.1 MB ignored report into this
document. Batch membership is the source IDs in the stated aggregate range
whose `path` begins with the listed prefix.

| Batch | Path area | Unused files | Unused exports | Status | Initial disposition |
| --- | --- | ---: | ---: | --- | --- |
| TR-01 | `tools/harness/phase-accounting/**` | 53 | 320 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-02 | `tools/harness/scheduler/**` | 29 | 254 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-03 | `tools/harness/generated-artifacts/**` | 28 | 124 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-04 | `tools/harness/backend/**` | 27 | 129 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-05 | `tools/harness/output/**` | 22 | 80 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-06 | `tools/harness/diagnostics/**` | 15 | 50 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-07 | `tools/harness/execution/**` | 13 | 78 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-08 | `tools/harness/contract/**` | 12 | 167 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-09 | Other `tools/**` paths outside the named harness areas | 9 | 40 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-10 | `tools/harness/static-analysis/**` | 9 | 12 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-11 | `tools/harness/browser/**` | 7 | 31 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-12 | `tools/harness/duration-accounting/**` | 6 | 14 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-13 | `tools/harness/finalization/**` | 4 | 10 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-14 | `tools/harness/readiness/**` | 2 | 0 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-15 | `tools/harness/test-support/**` | 2 | 0 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-16 | `tools/harness/command-surface/**` | 1 | 10 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-17 | `tools/harness/runtime-binary-registry.mjs` | 1 | 13 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-18 | `tools/harness/smoke/**` | 1 | 0 | TODO | Corrected-run reconciliation required in DC-003 |
| TR-19 | `tools/harness/tests/**` | 1 | 0 | TODO | Corrected-run reconciliation required in DC-003 |
| **Total** | `UF-0005..UF-0246`; `UE-0028..UE-1359` | **242** | **1,332** |  |  |

The initial disposition is a quarantine posture, not a final claim that every
member is a false positive. DC-003 must replace each batch count with corrected
results and identify any surviving finding individually before remediation.

## 5. Explicit Actionable Finding Inventory

All rows remain `TODO` until their required discovery and validation are
recorded. Line numbers identify the frozen report and may drift.

### 5.1 Application files

| ID | Path | Current evidence | Proposed disposition | Owner and exit condition |
| --- | --- | --- | --- | --- |
| UF-0001 | `apps/web/public/assets/fonts/fonts.css` | Loaded by `apps/web/index.html`; checked by frontend font-bundle evidence. | FALSE_POSITIVE_CONFIG | Web asset/config owner; accurate static-asset reachability with `make frontend-unit` passing. |
| UF-0002 | `apps/web/src/app/LandingAdminSurface.tsx` | Compatibility re-export documented in `apps/web/src/README.md`; no current code import found. | REMOVE | App shell; remove the barrel and its documentation unless a concrete continuing consumer is found. |
| UF-0003 | `apps/web/src/testing/testSetup.dom.ts` | Selected by the `browser-unit` Vitest project. | FALSE_POSITIVE_CONFIG | Frontend test setup; correct the Fallow entry and keep Vitest behavior unchanged. |
| UF-0004 | `apps/web/src/testing/testSetup.ts` | Selected by both Vitest projects. | FALSE_POSITIVE_CONFIG | Frontend test setup; correct the Fallow entry and keep Vitest behavior unchanged. |

### 5.2 Live-file unused exports

For these findings, first search direct, type-only, test, dynamic, and package
consumers. Use `INTERNALIZE` when the symbol remains locally useful and
`REMOVE` only when the symbol itself has no continuing value.

| ID | Path and line | Export | Workstream owner | Status |
| --- | --- | --- | --- | --- |
| UE-0001 | `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx:78` | `primaryToolbarButtonStyle` | Workbook toolbar | TODO |
| UE-0002 | `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx:124` | `WorkbookToolbarSearchLabel` | Workbook toolbar | TODO |
| UE-0003 | `apps/web/src/workbook/models/entityWorkbookModel.ts:125` | `readEntityStringCell` | Entity workbook model | TODO |
| UE-0004 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:1` | `evidenceRecordLifecycleStates` | Evidence lifecycle view model | TODO |
| UE-0005 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:13` | `objectBlobUploadStates` | Evidence lifecycle view model | TODO |
| UE-0006 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:22` | `evidenceLifecycleViewStateKeys` | Evidence lifecycle view model | TODO |
| UE-0007 | `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts:36` | `evidenceCountDisplayStateKeys` | Evidence lifecycle view model | TODO |
| UE-0008 | `apps/web/src/workbook/models/workbookContractRows.ts:39` | `materializeWorkbookViewRow` | Workbook contract rows | TODO |
| UE-0009 | `apps/web/src/workbook/models/workbookInspectorModel.ts:154` | `clearRowBoundInspectorState` | Workbook inspector | TODO |
| UE-0010 | `apps/web/src/workbook/models/workbookMutations.ts:16` | `submitViewRecordPatch` | Workbook mutations | TODO |
| UE-0011 | `apps/web/src/workbook/models/workbookSurfaceRegistry.ts:222` | `getWorkbookSurfaceRegistryEntry` | Workbook surface registry | TODO |
| UE-0012 | `apps/web/src/workbook/models/workbookSurfaceRegistry.ts:228` | `isBuiltInWorkbookSurfaceId` | Workbook surface registry | TODO |
| UE-0013 | `apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts:56` | `timelineTabClientInstanceId` | Timeline pending saves | TODO |
| UE-0014 | `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts:33` | `clearAppliedTimelineFilterDraft` | Timeline runtime | TODO |
| UE-0015 | `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts:43` | `applyTimelineFilterDraftToQuery` | Timeline runtime | TODO |
| UE-0016 | `apps/web/src/workbook/timeline/models/workbookMentionChips.ts:11` | `mentionChipStates` | Timeline mention chips | TODO |
| UE-0017 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:16` | `clipboardTextLooksTabular` re-export | Timeline model | TODO |
| UE-0018 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:454` | `readTimelineStringCell` | Timeline model | TODO |
| UE-0019 | `apps/web/src/workbook/utils/workbookGridFocus.tsx:37` | `formatWorkbookFocusAnchor` | Workbook grid focus | TODO |
| UE-0020 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:465` | `mergePendingCollectionActionPayload` | Pending queue | TODO |
| UE-0021 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:485` | `mergePendingReplayPayload` | Pending queue | TODO |
| UE-0022 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:544` | `canonicalizePendingReplayChanges` | Pending queue | TODO |
| UE-0023 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:561` | `buildPendingReplayMutationIdentity` | Pending queue | TODO |
| UE-0024 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:969` | `shouldRetryPendingFailure` | Pending queue | TODO |
| UE-0025 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:997` | `canCoalescePendingReplayUnits` | Pending queue | TODO |
| UE-0026 | `apps/web/src/workbook/utils/workbookStyles.ts:62` | `statusStripMutedItemStyle` | Workbook styles | TODO |
| UE-0027 | `apps/web/src/workbook/utils/workbookStyles.ts:78` | `statusStripSpacerStyle` | Workbook styles | TODO |

### 5.3 Unused exported types

| ID | Path and line | Type | Workstream owner | Status |
| --- | --- | --- | --- | --- |
| UT-0001 | `apps/web/src/app/AccountAdministrationPanels.tsx:65` | `AccountSecurityPanelHandle` | App administration | TODO |
| UT-0002 | `apps/web/src/app/AccountAdministrationPanels.tsx:70` | `DeploymentUsersPanelHandle` | App administration | TODO |
| UT-0003 | `apps/web/src/app/AccountAdministrationPanels.tsx:80` | `DeploymentUsersPanelCommandState` | App administration | TODO |
| UT-0004 | `apps/web/src/app/ReferencePackAdminPanel.tsx:54` | `ReferencePackAdminPanelHandle` | Reference-pack administration | TODO |
| UT-0005 | `apps/web/src/app/api/appShellClient.ts:111` | `EnterpriseAuthBeginResponse` | App-shell client | TODO |
| UT-0006 | `apps/web/src/app/api/appShellClient.ts:686` | `Phase1Response` | App-shell client | TODO |
| UT-0007 | `apps/web/src/services/browserApi.ts:23` | `PublicErrorDetail` | Browser API | TODO |
| UT-0008 | `apps/web/src/services/browserApi.ts:23` | `PublicErrorView` | Browser API | TODO |
| UT-0009 | `apps/web/src/workbook/WorkbookShell.tsx:88` | `WorkbookIncidentIdentity` | Workbook shell | TODO |
| UT-0010 | `apps/web/src/workbook/WorkbookShell.tsx:91` | `WorkbookAccountModel` | Workbook shell | TODO |
| UT-0011 | `apps/web/src/workbook/WorkbookShell.tsx:92` | `WorkbookIncidentControlsMenuItem` | Workbook shell | TODO |
| UT-0012 | `apps/web/src/workbook/WorkbookShell.tsx:94` | `WorkbookIncidentRole` | Workbook shell | TODO |
| UT-0013 | `apps/web/src/workbook/WorkbookShell.tsx:95` | `WorkbookIncidentSnapshot` | Workbook shell | TODO |
| UT-0014 | `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts:75` | `WorkbookShellActiveQueryControls` | Workbook shell runtime | TODO |
| UT-0015 | `apps/web/src/workbook/timeline/components/TimelineGridSurface.tsx:12` | `TimelineGridSurfaceProps` | Timeline grid | TODO |
| UT-0016 | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx:186` | `SaveState` | Timeline workbook | TODO |
| UT-0017 | `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts:38` | `TimelineRowContextMenuState` | Timeline inspector | TODO |
| UT-0018 | `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts:21` | `TimelineWorkbookRuntimeSaveState` | Timeline runtime | TODO |
| UT-0019 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:18` | `WorkbookRecordFreshnessDecision` | Timeline model | TODO |
| UT-0020 | `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts:19` | `WorkbookVersionedRecord` | Timeline model | TODO |

### 5.4 Unused class members

All four findings concern `WorkbookPendingQueueModel`. Characterize pending
dispatch, auth recovery, conflict clearing, retry, and replay behavior before
removing methods; do not retain public methods merely as speculative extension
points.

| ID | Path and line | Member | Proposed disposition | Status |
| --- | --- | --- | --- | --- |
| UCM-0001 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1207` | `settleDispatched` | REMOVE after characterization | TODO |
| UCM-0002 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1289` | `resumeAfterAuthRecovery` | REMOVE after characterization | TODO |
| UCM-0003 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1294` | `pauseForAuthRecovery` | REMOVE after characterization | TODO |
| UCM-0004 | `apps/web/src/workbook/utils/workbookPendingQueue.ts:1300` | `clearSameFieldConflict` | REMOVE after characterization | TODO |

### 5.5 Duplicate exports

Each pair must acquire one semantic owner. Do not suppress a duplicate simply
because both definitions currently have local callers.

| ID | Export | Locations | Proposed disposition | Status |
| --- | --- | --- | --- | --- |
| DE-0001 | `PagingMeta` | `apps/web/src/app/api/appShellClient.ts:71`; `apps/web/src/app/landingAdminTypes.ts:26` | CONSOLIDATE under the app-shell paging contract owner | TODO |
| DE-0002 | `TimelineRowContextMenuPosition` | `apps/web/src/workbook/timeline/components/TimelineRowActions.tsx:19`; `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts:33` | CONSOLIDATE under Timeline interaction state | TODO |
| DE-0003 | `deferred` | `apps/web/src/testing/fetchMockTestSupport.ts:175`; `apps/web/src/testing/timelineWorkbookTestSupport.ts:128` | CONSOLIDATE in generic test support or keep private per fixture if semantics differ | TODO |
| DE-0004 | `fetchJSON` | `apps/web/src/services/browserApi.ts:51`; `apps/web/src/services/workbookApi.ts:66` | CONSOLIDATE behind one browser transport helper if error semantics match | TODO |
| DE-0005 | `timelineViewSchemaId` | `apps/web/src/testing/timelineWorkbookTestSupport.ts:28`; `apps/web/src/workbook/models/workbookSurfaceRegistry.ts:7` | CONSOLIDATE on the production stable identifier | TODO |

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
| CD-0001 | 6 | `backend/backend-shard-plan.mjs` -> `backend/go-shard-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | TODO |
| CD-0002 | 5 | `backend/backend-shard-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | TODO |
| CD-0003 | 5 | `backend/backend-target-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | TODO |
| CD-0004 | 6 | `backend/backend-target-plan.mjs` -> `backend/target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | TODO |
| CD-0005 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/frontend-phase-manifest.mjs` -> `phase-accounting/frontend/phase-artifacts.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | TODO |
| CD-0006 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `execution/execution-dependencies.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | TODO |
| CD-0007 | 4 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` | TODO |
| CD-0008 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | TODO |
| CD-0009 | 6 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/browser-work-units.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | TODO |
| CD-0010 | 5 | `backend/backend-target-plan.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | TODO |
| CD-0011 | 5 | `diagnostics/task-execution-map.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` -> `diagnostics/task-guidance.mjs` | TODO |
| CD-0012 | 5 | `diagnostics/task-guidance.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/backend-work-units.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` | TODO |
| CD-0013 | 5 | `diagnostics/task-guidance.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/browser-work-units.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` | TODO |
| CD-0014 | 4 | `diagnostics/task-guidance.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/row-selection.mjs` | TODO |
| CD-0015 | 3 | `execution/execution-dependencies.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` | TODO |
| CD-0016 | 5 | `execution/execution-dependencies.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` | TODO |
| CD-0017 | 4 | `execution/summary-topology.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/task-surface/index.mjs` -> `generated-artifacts/task-surface/validation.mjs` | TODO |
| CD-0018 | 5 | `execution/summary-topology.mjs` -> `generated-artifacts/index.mjs` -> `generated-artifacts/task-surface/index.mjs` -> `generated-artifacts/task-surface/validation.mjs` -> `generated-artifacts/task-surface/recipe-validation.mjs` | TODO |
| CD-0019 | 5 | `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/frontend-phase-manifest.mjs` -> `phase-accounting/frontend/phase-artifacts.mjs` | TODO |
| CD-0020 | 5 | `generated-artifacts/index.mjs` -> `generated-artifacts/render-service-backed-schedule-manifest.mjs` -> `phase-accounting/index.mjs` -> `phase-accounting/phase-slice-plan.mjs` -> `phase-accounting/phase-slice-planning/browser-work-units.mjs` | TODO |

All paths in this table are relative to `tools/harness/`. The cycle exit
condition is a corrected Fallow run with zero cycle findings, unchanged public
Make behavior, passing harness contract evidence, and no replacement cycle
hidden by a suppression or broad barrel.

## 7. Dependency-Ordered Work Tracker

| ID | Work item | Status | Depends on | Affected owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| DC-001 | Freeze source identity, counts, authority, and command posture. | DONE | none | Tracker | Sections 1-3 | Fingerprint and all 1,656 findings reconcile. |
| DC-002 | Repair Fallow reachability for Vitest setup, Make/manifest CLIs, browser assets, and executable dependencies. | DONE | DC-001 | `.fallowrc.json`, harness static analysis | Config diff; caller and manifest inventory; retained `frontend-fallow-static` run root `.cartulary/test-results/20260713T171113Z-p616212`; `make harness-contract` passed | Reachability represents durable owner patterns without per-file suppression growth. |
| DC-003 | Run canonical Fallow and reconcile the quarantined tool batches. | TODO | DC-002 | Fallow harness | Corrected Fallow fingerprint `bbf66bb19c207cf2c1b3dfab6744085f2c4c54f8130a8112da963b67954cc0f6`; surviving tool counts 37 unused files and 289 unused exports | Every old ID/batch is closed or mapped to a new finding; no count silently disappears. |
| DC-004 | Resolve four application file findings and remove the obsolete compatibility barrel if no continuing consumer exists. | BLOCKED | DC-003 | Web app, assets, frontend tests | Import searches, Vite config, frontend unit evidence | Runtime/test-loaded files are modeled accurately; valueless barrel is removed. |
| DC-005 | Internalize or remove the 27 live exports, 20 types, and four class members in cohesive owner slices. | BLOCKED | DC-003 | App shell and workbook owners | Symbol searches, characterization tests, frontend checks | Public surface is smaller without product behavior drift or speculative compatibility exports. |
| DC-006 | Consolidate five duplicate exports under semantic owners. | BLOCKED | DC-003 | App shell, Timeline, services, test support | Caller migration and owner-local tests | One definition or deliberately private per-owner definition remains for each name. |
| DC-007 | Break the harness cycle family through leaf dependencies and narrow interfaces. | TODO | DC-001 | Harness backend, phase accounting, diagnostics, execution, generated artifacts | Dependency diagram, script lint, harness contract, Fallow rerun | CD-0001 through CD-0020 close structurally and future phase growth is declarative. |
| DC-008 | Close the font import and `cdxgen` dependency findings. | BLOCKED | DC-002, DC-003 | Web assets and release evidence | Fallow report, frontend unit, release/SBOM evidence as applicable | Dynamic uses are represented accurately; no valuable dependency or asset is removed. |
| DC-009 | Review every retained exception and reject unjustified legacy or compatibility burden. | BLOCKED | DC-004 through DC-008 | All affected owners | Exception ledger with owner, value, evidence, and removal trigger | No inline suppression or undocumented indefinite compatibility path remains. |
| DC-010 | Complete final verification, accounting, and handoff. | BLOCKED | DC-003 through DC-009 | Repository-wide | Finalizer, broad check, retained run roots, final Fallow fingerprint | Findings reconcile, required checks pass, residual risk and next action are explicit. |

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

### 9.2 Retained-exception ledger

Add a row before assigning `RETAIN_JUSTIFIED`. An empty ledger is preferred.

| Finding ID | Owner | Continuing current value | Why structural removal is worse | Regression evidence | Config-level handling | Removal trigger | Review date |
| --- | --- | --- | --- | --- | --- | --- | --- |
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

An exception is invalid when it cites only possible future use, historical use,
compatibility in the abstract, or the cost of cleanup. Invalid exceptions must
be removed or marked `BLOCKED` pending a named owner decision.

### 9.3 Blockers and risks

| ID | Risk or blocker | Blocking work | Resolution condition |
| --- | --- | --- | --- |
| DR-001 | Frozen Fallow entry coverage did not represent Make/manifest-driven tool execution. | Tool file/export cleanup | DC-002 is complete; DC-003 must map the surviving corrected-run tool findings before cleanup. |
| DR-002 | Removing pending-queue methods without characterization could change retry, conflict, or auth-recovery behavior. | UCM-0001 through UCM-0004 | Relevant frontend unit/behavior coverage is inspected and passing. |
| DR-003 | Breaking cycles through a new shared barrel could reproduce the same coupling under another path. | DC-007 | Leaf dependency direction is documented and the corrected graph has no cycle. |
| DR-004 | Compatibility barrels and duplicate types can obscure real ownership. | UF-0002 and DE-0001 through DE-0005 | One semantic owner is selected; callers migrate without dual paths. |
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

- [ ] A corrected canonical Fallow baseline supersedes the frozen reachability-
  limited report and records its fingerprint and retained run root.
- [ ] All 1,656 source findings are closed individually or through the defined
  auditable batches; totals reconcile without silent disappearance.
- [ ] UF-0005 through UF-0246 and UE-0028 through UE-1359 were not used as
  deletion authority before DC-002 and DC-003 completed.
- [ ] Every surviving genuine unused file, export, type, member, dependency,
  duplicate, unresolved import, and cycle has a completed disposition.
- [ ] The seven binding cleanup gates in Section 2 are satisfied for every
  implementation slice.
- [ ] No legacy facade, compatibility alias, dual path, or feature is retained
  without clear continuing value and an owner.
- [ ] No inline suppression, per-file configuration sprawl, generated-file hand
  edit, lockfile hand edit, or phase-shaped runtime dependency was introduced.
- [ ] Harness cycles are removed structurally and later phase/target growth
  remains declarative.
- [ ] Public Make identities, command IDs, artifacts, summaries, and failure
  mapping remain stable unless their owner document was changed first.
- [ ] Narrow validation, `make agent-finalize`, final broad verification, and
  final Fallow results are recorded with exact outcomes and skipped-check
  reasons.
- [ ] The last handoff record lets another engineer resume without rediscovering
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
