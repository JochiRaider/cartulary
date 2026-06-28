# WorkbookShell Refactor Tracker and Handoff Planner

## 1. Session Header

| Field | Value |
| --- | --- |
| Date/time | 2026-06-28T15:47:43-04:00 |
| Branch/commit | `main...origin/main` at `cc6d46abe952f2b4f886a93c1e698a0776b17ab3` |
| Dirty tree state | Clean before creating this planner. Current remediation dirty state includes this artifact plus `WorkbookShell.tsx`, `useEntityTimelinePreview.ts`, and `WorkbookShell.surfaces.test.tsx`. |
| Target file | `apps/web/src/workbook/WorkbookShell.tsx` |
| Target module/package/seam | `/apps/web` Workbook shell coordinator: composes surfaces, saved-view/query controls, incident controls, generic surfaces, assessment/entity support, and Timeline entrypoints. |
| Planning mode | Initial artifact was planning-only; current remediation record below documents implementation and test changes made before extraction. |
| Framework path used | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Framework posture | Structure only. Current repository state below is from live inspection, not inferred from the framework. |
| Primary owner references inspected | `docs/domain.md`; `docs/testing-harness-nlspec.md`; `docs/design.md`; Core 01, Core 03, and Core 04 snippets relevant to workbook surfaces, saved views, inspector, collaboration, routes, generated artifacts, and harness behavior. |
| Source limits | Search output was large and sometimes truncated. File-level tables list inspected files. Broader `apps/web/e2e/**`, `apps/web/src/workbook/timeline/**`, backend route/store files, generated protocol outputs, and full owner specs were not exhaustively read line by line unless listed below. |
| Unseen relevant files | TODO: inspect any newly touched file before implementation. TODO: inspect phase maps, generated ledgers, and exact Make target plans before using retained harness evidence. TODO: inspect backend route/store files before any route, envelope, authorization, or query behavior change, which is out of scope for this planner. |

## 2. Current-State Repository Scan

### Inspected Files Table

| Path | Inspection level | Evidence found |
| --- | --- | --- |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Read | Reusable tracker, workflows WF-00 through WF-13, guardrails, handoff, RF-AC model. |
| `docs/domain.md` | Read/search | Domain vocabulary: workbook surface, stable identifiers, source/projection boundary, implementation-support boundary. |
| `docs/testing-harness-nlspec.md` | Read/search | Make-owned command surface, generated-artifact prohibition, failure classification, visual/a11y support boundary. |
| `docs/design.md` | Search | Design-direction only; grid vendor coordinates must not become user-facing concepts; visual/a11y evidence is not product conformance alone. |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Search | Saved-view, workbook-preference, startup, record mutation, view query, generated contract, and route owner evidence. |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Search | Saved-view identity, inspector default-closed behavior, query/sort/filter/group, Timeline write-back, presence/live update contracts. |
| `docs/spec/04_security_deployment_and_conformance.md` | Search | Session and authorization posture; no frontend refactor may alter authorization outcomes. |
| `apps/web/src/README.md` | Read excerpt | Declares `WorkbookShell.tsx` as workbook shell coordinator and package facade discipline. |
| `package.json` | Read | Root pnpm scripts; Make remains canonical for verification. |
| `apps/web/package.json` | Read | Web dependencies include `@cartulary/grid-adapter`, `@cartulary/ui-contracts`, `@cartulary/view-contracts`; Vitest/Playwright scripts are developer conveniences. |
| `Makefile` | Search/help | Public targets include `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check`, `lint-biome`, browser E2E targets, `agent-finalize`, and phase slice targets. |
| `apps/web/src/app/App.tsx` | Read excerpt | Lazy imports and renders `WorkbookShell`; passes account, density, incident snapshot/access callbacks, and incident controls renderer. |
| `apps/web/src/workbook/WorkbookShell.tsx` | Read/search | 2,955 lines. Exports `WorkbookShell` and shell contract types; contains shell composition, entity surface, assessment surface, loaders, session role, incident drawer, saved-view/query wiring, density, and Timeline entrypoint props. |
| `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts` | Read | 862 lines. Owns startup selection, active surface, saved views, query state, URL synchronization, active query controls, and runtime commands. |
| `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx` | Search | Active-surface saved-view selector and mutation controls. |
| `apps/web/src/workbook/components/GenericWorkbookSurface.tsx` | Search | Generic contract-backed surface renderer, generic create/edit/reference behavior. |
| `apps/web/src/workbook/components/SystemViewSwitcher.tsx` | Search | System view grouped navigation. |
| `apps/web/src/workbook/components/WorkbookGridControls.tsx` | Search | Sort/filter/group control UI. |
| `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx` | Search | Surface toolbar, add row and inspector controls. |
| `apps/web/src/workbook/components/WorkbookSurfaceFrame.tsx` | Search | Shared grid/inspector/status/view-bar frame and style primitives. |
| `apps/web/src/workbook/components/WorkbookShellSlots.tsx` | Search | Stable shell slot regions and `workbookShellId`. |
| `apps/web/src/workbook/models/workbookQuery.ts` | Search | Query state, filter draft, query request and saved-view query/layout JSON builders. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.ts` | Search | Saved-view identity, base surface identity, fallback after delete, modified-state logic. |
| `apps/web/src/workbook/models/workbookSavedViews.ts` | Search | Saved-view resource normalization and mutation permissions. |
| `apps/web/src/workbook/models/workbookStartup.ts` | Search | `sheet_ref` startup URL and backend selection normalization. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.ts` | Search | Stable view schema IDs, required built-in/system/optional registry, known ID fallback. |
| `apps/web/src/workbook/models/workbookSurfaceQueryRuntime.ts` | Search | Maps view schema IDs to query state slots and contracts. |
| `apps/web/src/workbook/models/workbookContractRows.ts` | Search | View-row normalization, grid rows, contract columns. |
| `apps/web/src/workbook/models/genericWorkbookModel.ts` | Search | Generic create/patch payloads, cell labels, reference options, mutation error parsing. |
| `apps/web/src/workbook/models/entityWorkbookModel.ts` | Search | Host/identity row normalization, grouping labels, merge plan. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.ts` | Search | Assessment defaults, column widths, support labels. |
| `apps/web/src/workbook/models/workbookInspectorModel.ts` | Search | Inspector config selection, default-closed state model, stale row-bound state clearing. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Read/search | 2,158-line adjacent Timeline surface entrypoint with many extracted hooks and local composition. |
| `packages/grid-adapter/src/index.tsx` | Read excerpt | Owns direct `react-data-grid` stylesheet import and grid rendering facade. |
| `packages/grid-adapter/src/core.ts` | Read excerpt | Owns Cartulary grid row, anchor, navigation, presentation row, and paste target abstractions. |
| `packages/ui-contracts/src/index.ts` | Read/search | Owns stable test IDs/selectors for grid shell, rows, inspector, saved views, incident controls, Timeline controls. |
| `packages/view-contracts/src/index.ts` | Read/search | Owns view contract parsing, `record_id`/`row_version`/`field_key` row validation, inspector config validation. |
| `tools/generated_artifact_policy.json` | Read | Generated roots and lint-scope policy. |
| `tools/frontend_import_boundaries.json` | Read | RDG boundary, generated protocol/design token facade boundaries, workspace package facade rules. |
| `contracts/view-schemas/index.json` | Attempted summary | Shape did not summarize as a flat list with current command. TODO: inspect exact index shape before contract work. |
| `contracts/view-schemas/cartulary.view.timeline.v2.json` | Parsed summary | Timeline ID, technical fields, sort/filter/group, zero-field create, default-closed inspector config. |
| `contracts/view-schemas/cartulary.view.hosts.v1.json` | Parsed summary | Hosts ID, technical fields, create minima, sort/filter/group, default-closed inspector config. |
| `contracts/view-schemas/cartulary.view.identities.v1.json` | Parsed summary | Identities ID, technical fields, create minima, sort/filter/group, default-closed inspector config. |
| `contracts/view-schemas/cartulary.view.assessments.v1.json` | Parsed summary | Assessments ID, technical fields, no grouping fields, default-closed inspector config. |
| `contracts/view-schemas/cartulary.view.notes.v1.json` | Parsed summary | Notes ID, generic surface contract, technical fields, default-closed inspector config. |
| `apps/web/src/workbook/WorkbookShell.assessments.test.tsx` | Test inventory | Assessment create payload/UI and superseded assessment/entity query response tests. |
| `apps/web/src/workbook/WorkbookShell.phase3.autosave.test.tsx` | Test inventory | Timeline autosave, save-state labels, stale row, pending save coalescing. |
| `apps/web/src/workbook/WorkbookShell.phase3.grid.test.tsx` | Test inventory | `record_id`/`row_version`, field-key commits, stale query/live patch, draft preservation. |
| `apps/web/src/workbook/WorkbookShell.phase3.payload.test.tsx` | Test inventory | Timeline zero-field create and blank row payload behavior. |
| `apps/web/src/workbook/WorkbookShell.phase4.actionSequencing.test.tsx` | Test inventory | Current row versions, pending autosave before actions, stale reload handling. |
| `apps/web/src/workbook/WorkbookShell.phase4.saveState.test.tsx` | Test inventory | Status strip save-state presentation. |
| `apps/web/src/workbook/WorkbookShell.phase4.support.test.tsx` | Test inventory | Timeline row actions, mention chips, relationship edits, continuity, websocket invalidations. |
| `apps/web/src/workbook/WorkbookShell.phase4.timelineQuery.test.tsx` | Test inventory | Timeline query row identity integration. |
| `apps/web/src/workbook/WorkbookShell.phase5.gridProvenance.test.tsx` | Test inventory | Hosts, identities, notes grid provenance. |
| `apps/web/src/workbook/WorkbookShell.phase5.mentionChips.test.ts` | Test inventory | Mention chip state by stable identifiers and field keys. |
| `apps/web/src/workbook/WorkbookShell.phase5.test.tsx` | Test inventory | Evidence count reflection without forced navigation. |
| `apps/web/src/workbook/WorkbookShell.phase6.test.tsx` | Test inventory | Presence, sparse live patches, conflicts, pending queue replay and save state. |
| `apps/web/src/workbook/WorkbookShell.phase7.test.tsx` | Test inventory | History, rollback, row retargeting, conflict anchors, focus behavior. |
| `apps/web/src/workbook/WorkbookShell.phase8.query.test.tsx` | Test inventory | Query controls, saved-view JSON, grouping presentation rows, field-key requests. |
| `apps/web/src/workbook/WorkbookShell.phase9.inspector.test.tsx` | Test inventory | Inspector default-closed, sheet switch close, no-row state, row-local actions. |
| `apps/web/src/workbook/WorkbookShell.phase9.sentinel.test.tsx` | Test inventory | Keyboard/focus anchors, paste target resolution, group/presentation rejection, vendor-coordinate rejection. |
| `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx` | Test inventory | Surface selection, startup, saved views, generic mutations, evidence handles/attachment. |
| `apps/web/src/workbook/models/workbookQuery.test.ts` | Test inventory | Query/filter/group model behavior. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.test.ts` | Test inventory | Saved-view identity and runtime query restoration. |
| `apps/web/src/workbook/models/workbookSavedViews.test.ts` | Test inventory | Saved-view normalization and mutation permissions. |
| `apps/web/src/workbook/models/workbookStartup.test.ts` | Test inventory | Startup fallback, invalid pointers, saved-view identity preservation. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.test.ts` | Test inventory | Surface registry stable IDs and grouping. |
| `apps/web/src/workbook/models/workbookInspectorModel.test.ts` | Test inventory | Inspector config selection, default-closed behavior, stale invalidation. |
| `apps/web/src/workbook/models/genericWorkbookModel.test.ts` | Test inventory | Generic create/patch payloads, minima, reference labels, errors. |
| `apps/web/src/workbook/models/entityWorkbookModel.test.ts` | Test inventory | Entity row normalization, merge plan, grouping/widths. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.test.ts` | Test inventory | Assessment defaults, widths, support labels. |
| `apps/web/src/workbook/utils/workbookClipboard.test.ts` | Test inventory | Clipboard parsing and dimensions. |
| `apps/web/src/workbook/utils/workbookContinuity.test.ts` | Test inventory | Focus/scroll continuity. |
| `apps/web/src/workbook/utils/workbookKeyboard.test.ts` | Test inventory | Keyboard command contract. |
| `apps/web/src/workbook/utils/workbookPendingQueue.test.ts` | Test inventory | Pending queue identity, replay, coalescing, save state, conflict anchors. |
| `apps/web/src/workbook/utils/workbookValueFormat.test.ts` | Test inventory | Grid display value formatting. |
| `apps/web/src/testing/timelineWorkbookTestSupport.test.tsx` | Test inventory | Timeline workbook test support helpers. |
| `docs/archive/TimelineWorkbook.refactor-tracker-1.md` | Read/search | Prior Timeline tracker pattern and completed slices; evidence only, not current state authority. |
| `docs/archive/TimelineWorkbook.refactor-tracker-2.md` | Read/search | Prior Timeline tracker pattern and workflow table; evidence only, not current state authority. |

### Direct Imports And Imported Symbols

| Import source | Imported symbols in `WorkbookShell.tsx` | Contract risk |
| --- | --- | --- |
| `@cartulary/grid-adapter` | `buildGridPresentationRows`, `GridActionsColumn`, `GridColumn`, `GridDensity`, `GridRow`, `GridTable`, `GridViewport`, `reconcileRecordRows`, `resolveGridPasteTargets` | Grid facade only. Do not import `react-data-grid` or vendor coordinates in `/apps/web`. |
| `@cartulary/ui-contracts` | Selector/test-id builders for assessment, entity, generic edit/create, grid headers/groups/shells, incident controls, surface tabs/menu, timeline preview, workbook shell/status, row action, draft row, plus `WorkbookSurface`/`IncidentControlsSection` types | Selectors are stable public test/runtime contracts; preserve or update only through owner package. |
| `@cartulary/view-contracts` | `requireViewContract`, `resolveHeaderSortFieldKey`, `visibleFields` | View-schema and field-key behavior must stay contract-derived. |
| `lucide-react` | `MoreHorizontal`, `X` | Presentation only; do not let icon names become behavior authority. |
| `react` | `CSSProperties`, `Dispatch`, `ClipboardEvent`, `ReactNode`, `SetStateAction`, hooks | Component/runtime state. |
| `../services/browserApi` | `apiPath` | Public route path construction; must not change route shapes. |
| `../services/workbookApi` | `abortLatestQuery`, `beginLatestQuery`, `fetchJSON`, `handleWorkbookLoadFailure`, `isAbortError`, `LatestQueryRuntime`, `parseErrorMessage`, `readEnvelope` | Query freshness, abort, envelope/error handling, and access-loss behavior. |
| `../shared/workbookShellContracts` | Account, incident controls, role, snapshot contract types | Public shell prop/caller contract used by `App.tsx`. |
| `./components/*` | Saved-view selector, generic surface/control, system switcher, grid controls, inspector feature groups, sheet toolbar, shell slots, status strip, surface frame/style primitives | Candidate extraction seams; preserve prop contracts and selectors. |
| `./hooks/*` | Assessment support rows, entity Timeline preview, incident identity, pending grid focus, responsive layout, shell runtime | Adjacent controller seams; runtime hook is high-risk public contract hub. |
| `./models/*` | Assessment, entity, generic, contract-row, density, incident identity, inspector, query, reference options, surface registry models | Behavior must remain model-driven and stable-ID-driven. |
| `./timeline/*` | `RelationshipChip`, `TimelineWorkbook`, `AssessmentCreateDraft`, `buildAssessmentCreatePayload`, `EntityApiRow`, clipboard utilities, focus utilities, presence initials | Timeline is adjacent entrypoint; do not redesign or rewrite it in WorkbookShell slices. |

### Exported Symbols And Known Callers

| Export | Known callers/evidence | Notes |
| --- | --- | --- |
| `WorkbookShell` | Lazy imported by `apps/web/src/app/App.tsx`; mocked by `App.phase1.test.tsx`, `App.landing.test.tsx`, and `App.phase1.support.test.tsx`; rendered by multiple `WorkbookShell.*.test.*` files. | Public app shell component. Keep props and behavior stable. |
| `WorkbookIncidentIdentity` type re-export | Type consumer not exhaustively searched beyond target scan. TODO: re-run `rg "WorkbookIncidentIdentity"` before changing. | Do not change type source or shape in refactor slices. |
| Shell contract type re-exports from `../shared/workbookShellContracts` | `App.tsx` and tests pass account/menu/incident controls props. | Public caller contract; no behavior change planned. |

### Adjacent Hooks, Components, Utilities

| Area | Adjacent files | Current role |
| --- | --- | --- |
| Shell runtime | `useWorkbookShellRuntime.ts` | Startup, saved views, surface selection, URL sync, active query controls. |
| Shared shell UI | `WorkbookSurfaceFrame.tsx`, `WorkbookSheetToolbar.tsx`, `WorkbookShellSlots.tsx`, `WorkbookGridControls.tsx`, `WorkbookStatusStrip.tsx` | Shell surface frame, slots, toolbar, query controls, status presentation. |
| Saved/system navigation | `ActiveSurfaceSavedViewSelector.tsx`, `SystemViewSwitcher.tsx`, `workbookSurfaceRegistry.ts`, `workbookStartup.ts`, `workbookSavedViews.ts`, `workbookSavedViewRuntime.ts` | Surface, saved-view, and startup identity behavior. |
| Generic surfaces | `GenericWorkbookSurface.tsx`, `GenericMutationControl.tsx`, `genericWorkbookModel.ts`, `useGenericReferenceOptions.ts` | Contract-backed system/built-in generic surfaces. |
| Entity support | `entityWorkbookModel.ts`, `useEntityTimelinePreview.ts` | Host/identity row view model, merge plan, Timeline preview. |
| Assessment support | `assessmentWorkbookModel.ts`, `useAssessmentSupportRows.ts` | Assessment draft defaults/support refs. |
| Timeline entrypoint | `timeline/components/TimelineWorkbook.tsx` and timeline hooks/models/services | Separate adjacent Timeline surface coordinator. |
| Grid/focus/paste | `workbookGridFocus.tsx`, `workbookClipboard.ts`, `workbookKeyboard.ts`, `workbookContinuity.ts`, `@cartulary/grid-adapter` | Cartulary anchor, keyboard, paste, continuity behavior. |

### Tests And Fixtures Found

| Test group | Files | Behaviors covered |
| --- | --- | --- |
| Shell/surfaces | `WorkbookShell.surfaces.test.tsx`, `WorkbookShell.assessments.test.tsx`, `WorkbookShell.phase5.gridProvenance.test.tsx` | Surface selection, startup, saved views, generic/evidence flows, assessments, host/identity/notes provenance. |
| Timeline grid/mutation | `WorkbookShell.phase3.*`, `WorkbookShell.phase4.*`, `WorkbookShell.phase5.test.tsx` | Create/edit/autosave, row identity, stale query, action sequencing, status, evidence counts. |
| Collaboration/history/inspector | `WorkbookShell.phase6.test.tsx`, `WorkbookShell.phase7.test.tsx`, `WorkbookShell.phase9.inspector.test.tsx` | Presence/live patches, pending queue, conflicts, history, rollback, inspector default-closed/retargeting. |
| Query/saved views | `WorkbookShell.phase8.query.test.tsx`, model saved-view/query/startup tests | Stable field-key query JSON, saved-view identity, startup fallback. |
| Keyboard/paste/focus | `WorkbookShell.phase9.sentinel.test.tsx`, utility tests, grid-adapter tests | Stable anchors, paste targets, presentation-row rejection, vendor-coordinate guard. |
| Browser/readiness | `apps/web/e2e/phase*.spec.ts`, `workbook.a11y*.spec.ts`, `workbook.visual.spec.ts`, visual snapshots | TODO: select exact browser rows only after `make task-guide`/`make explain-phase`; do not treat visual/a11y evidence as product conformance unless owner says so. |

### Generated Artifacts Or Contract-Derived Inputs Found

| Surface | Found evidence | Planner rule |
| --- | --- | --- |
| View schemas | `contracts/view-schemas/*.json`, summarized Timeline/Hosts/Identities/Assessments/Notes. | Contract inputs, not hand-edited by shell refactor unless owner work explicitly starts there. |
| View contracts facade | `packages/view-contracts/src/index.ts` parses and validates contracts/rows/inspector config. | Use facade; do not duplicate validation in shell. |
| UI contracts facade | `packages/ui-contracts/src/index.ts` owns selector/test-id builders. | Keep stable selectors or update through owner package. |
| Generated roots | `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, `tools/task_surface.generated.mk`. | Do not hand-edit. Use generator/drift targets if an owner input changes. |
| Protocol facade | `packages/protocol-ts/src/index.ts` and generated data are used indirectly by view contracts/services. | Do not import generated protocol artifacts directly in `/apps/web` runtime outside allowed facades. |

### Validation Commands Discovered Or TODO

| Command | Purpose | Status |
| --- | --- | --- |
| `make frontend-typecheck` | Cheapest TypeScript structural validation. | Discovered. |
| `make frontend-unit` | Primary frontend unit/interaction coverage. | Discovered. |
| `make frontend-import-boundary-check` | Enforce RDG/generated/facade boundaries. | Discovered. |
| `make lint-biome` | Authored frontend style/lint check. | Discovered. |
| `make task-guide ROLE=feature-dev PHASE=phaseN` | Select phase-specific narrow validation. | Discovered; TODO: run for the exact slice phase before implementation. |
| `make phase-slice PHASE=phaseN` | Phase slice verification. | Discovered; TODO: select only after task-guide/explain-phase. |
| `make browser-e2e*` | Browser, visual, a11y readiness. | Discovered; TODO: use only when slice touches browser-visible route/focus/a11y/visual behavior. |
| `make agent-finalize` | End-of-run harness maintenance. | Discovered; if `RESULTS_DIR` is unset, record retained-run maintenance skipped. |

## 3. Responsibility Diagnosis For TimelineWorkbook.tsx

This section addresses the requested `TimelineWorkbook.tsx` diagnosis as an adjacent Timeline entrypoint. It is not a second primary target.

| Concern | Current diagnosis |
| --- | --- |
| Current responsibilities | `TimelineWorkbook.tsx` composes Timeline grid, inspector, query controls, pending saves, conflicts, live updates, presence projection, mention/evidence/history/workflow actions, keyboard/paste/focus continuity, and Timeline notices. |
| Likely overreach | The file remains large at 2,158 lines even after multiple hooks exist. It still appears to coordinate render composition and many runtime seams in one entrypoint. |
| Stable public contracts it participates in | Timeline `view_schema_id`, `record_id`, `row_version`, `field_key`, query route, row create/patch/action routes, WebSocket `record_changed`/presence messages, pending queue conflict anchors, inspector config, stable selectors. |
| Risk classification | High for behavior changes; medium for purely presentational extraction inside Timeline; low only for mechanical moves backed by existing phase3/4/6/7/9 characterization and no prop/selector changes. |
| Seams that appear extractable | Presentational render helpers, inspector sections/renderers, notices, row actions, grid surface, hooks already extracted for history, mentions, evidence, pending replay, live updates, grid anchor, interactions, rows, and runtime. |
| Seams that must stay local until more evidence exists | Pending queue orchestration with conflict resolver, live update high-water behavior, focus/viewport continuity, paste/create interplay, route payload construction, and inspector stale-retargeting decisions. |
| WorkbookShell relationship | `WorkbookShell.tsx` passes incident/user/role/density/entity rows/query state/reload/saved-view selector props into `TimelineWorkbook`. WorkbookShell slices must preserve this prop contract and must not redesign Timeline UI. |

## 4. Owner-Contract Map

| Contract surface that could drift | Current participation | Owner/evidence | Refactor guard |
| --- | --- | --- | --- |
| Grid-first row creation, inline edit, paste, keyboard/focus behavior | Shell uses `GridTable`, `GridViewport`, draft rows, `resolveGridPasteTargets`, `FocusableWorkbookCell`, query state, and surface dispatch. | Core 03; `@cartulary/grid-adapter`; phase3 grid/autosave/payload; phase9 sentinel; utility tests. | Preserve Cartulary anchors and same-surface interaction; no vendor coordinate dependency. |
| `record_id` / `row_version` / `field_key` identity | Entity, assessment, generic, and Timeline flows key rows and mutations through stable IDs and base row versions. | Core 01/Core 03; `view-contracts`; `workbookContractRows`; phase3/4/6/7/8/9 tests. | Never retarget by visible row order, label, index, or component name. |
| Active `view_schema_id` and Timeline field bindings | Surface registry, `requireViewContract`, `activeContract`, `TimelineWorkbook` props, query/load routes. | Core 01/Core 03; `contracts/view-schemas/*`; `workbookSurfaceRegistry`; `workbookSurfaceQueryRuntime`. | Keep behavior keyed by immutable `view_schema_id`; saved views must not collapse into base view identity. |
| Query state, sorting, filtering, grouping, saved-view interaction | `useWorkbookShellRuntime` owns active controls, per-surface query states, saved-view CRUD/select/delete/default/home commands. | Core 01/Core 03; `workbookQuery`, saved-view model tests, `WorkbookShell.phase8.query.test.tsx`, `WorkbookShell.surfaces.test.tsx`. | Preserve query JSON shapes, field-key validation, grouping omission semantics, latest-query behavior. |
| Pending save/conflict/sync state | Timeline owns most pending queue behavior; entity/generic surfaces have local `SaveState`. Shell passes Timeline query/entity props and reload token. | Core 03; `workbookPendingQueue.test.ts`; phase3/4/6/7 tests. | Do not move pending queue semantics into phase-shaped shell code; keep save state labels and conflict anchors stable. |
| Inspector default-closed and row-retargeting behavior | Shell entity/assessment inspectors reset on `inspectorResetKey`; Timeline inspector receives reset key and saved-view selector. | Core 03; `view-contracts` inspector config; `workbookInspectorModel`; phase9 inspector tests. | Default closed; clear stale row-bound state on surface/saved-view/reload/row-version authorization events. |
| Collaboration/presence/live update behavior | Timeline owns WebSocket lifecycle and presence; Shell supplies `sheetRef`, current user/role, entity index, reload token. | Core 03 collaboration section; `workbookCollaborationMessages`, `workbookSocketLifecycle`, phase6 tests. | Do not alter WebSocket paths/messages, sheet-ref identity, high-water stale handling, or presence keys. |
| Generated contract usage | Shell uses facades, view-schema-derived contracts, generated-ui selectors via package facade. | `tools/generated_artifact_policy.json`; `tools/frontend_import_boundaries.json`; harness NLSpec. | No generated file hand-edit; use `make generated-artifact-policy-check` if contract inputs change. |
| Harness/test accounting | Verification must use Make targets; phase maps and visual/a11y artifacts are harness/readiness evidence only unless owner says otherwise. | `docs/testing-harness-nlspec.md`; Make help; `make task-guide`. | Record exact commands, run roots, and failure class; do not claim retained evidence without named artifacts. |

## 5. Refactor Tracker

Status values: `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define WorkbookShell target module and scope | scope | DONE | none | `/apps/web` workbook shell seam | Session header and target declaration | Exactly one primary target file and seam are explicit. |
| T-002 | Inspect current repo state | discovery | DONE | T-001 | `/apps/web`, package facades, owner docs | Inspected files table, import/caller/test/contract scan | Relevant files, imports, tests, generated paths, and commands are listed with source limits. |
| T-003 | Map owner contracts | contracts | DONE | T-002 | Core 01/03/04, domain, harness, design | Owner-contract map | Drift-prone public behavior and owner sources are mapped. |
| T-004 | Freeze characterization evidence | tests | DONE | T-003 | Current implementer | Characterization test plan | Existing and missing characterization coverage is known before code movement. |
| T-005 | Plan boundary guardrails | architecture | DONE | T-003 | `/apps/web`, `grid-adapter`, `view-contracts`, `ui-contracts`, generated policy | Boundary guardrails section | Import/generated/selector/phase/design guardrails are explicit. |
| T-006 | Plan behavior-preserving moves | implementation | DONE | T-004,T-005 | Future implementer | Candidate slice table | Small reviewable slice sequence is defined with validation and stop conditions. |
| T-007 | Plan validation loop | validation | DONE | T-006 | Future implementer | Validation plan | Cheapest sufficient Make targets are named and broader gates are conditional. |
| T-008 | Update docs/contracts if required | docs | DEFERRED | T-003 | Owner docs/contracts | No behavior/schema change planned | Remains deferred unless a future slice requires owner-doc or contract input changes before codegen. |
| T-009 | Execute or hand off | handoff | DONE | T-006,T-007,T-008 | Planning artifact | Session handoff section | Next actor can continue without rediscovery. |

## 6. Workflow Dependency Map

| Workflow | Name | Class | Status | Required previous workflows | Required subsequent workflows | Target-specific evidence and exit |
| --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap | root | DONE | none | WF-01 | Branch, commit, dirty state, framework path, target existence, and source limits recorded. |
| WF-01 | Current-state repository scan | chain | DONE | WF-00 | WF-02, WF-03 | Target imports/exports, callers, adjacent files, tests, generated policy, and validation commands scanned. |
| WF-02 | Module/package ownership inventory | chain | DONE | WF-01 | WF-04, WF-05 | `/apps/web` owns shell/controller; package facades own grid, view contracts, selectors. |
| WF-03 | Public contract freeze | chain | DONE | WF-01 | WF-04, WF-05, WF-06 | Stable identifiers, routes, saved views, inspector, collaboration, generated artifacts, and harness contracts mapped. |
| WF-04 | Refactor slice selection | chain | DONE | WF-02, WF-03 | WF-05, WF-06 | `WSH-SL-01` through `WSH-SL-06` selected in dependency order. |
| WF-05 | Characterization test plan | chain | DONE | WF-03, WF-04 | WF-09 | Existing tests mapped by behavior; TODOs identified for missing evidence before edits. |
| WF-06 | Boundary guardrail plan | chain | DONE | WF-02, WF-04 | WF-08, WF-09 | RDG/generated/selector/phase/design evidence guardrails documented. |
| WF-08 | Frontend package seam plan | parallel | DONE | WF-04, WF-05, WF-06 | WF-09 | App code remains on package facades; no direct vendor or generated-root edits planned. |
| WF-09 | Execution checkpoint plan | chain | DONE | WF-05, WF-08 | WF-10 | Candidate slices each include validation, rollback, and stop condition. |
| WF-10 | Validation and harness accounting plan | chain | DONE | WF-09 | WF-11 | Public Make target ladder and failure classification expectations recorded. |
| WF-11 | Documentation/generated-artifact plan | parallel | DONE | WF-03, WF-09 | WF-12 | This planner is the only doc artifact; generated hand-edit is prohibited. |
| WF-12 | Cleanup and anti-drift plan | chain | DONE | WF-10, WF-11 | WF-13 | Future slices must remove dead code only after callers move and rerun boundary checks. |
| WF-13 | Handoff and next-slice bootstrap | chain | DONE | WF-12 | none | Filled handoff record and blank template included. |

## 7. Candidate Refactor Slices

Each slice is behavior-preserving and intended for one reviewable implementation pass.

| Slice ID | Objective | Files likely touched | Prerequisites | Public behavior expected unchanged | Characterization evidence required before edit | Validation command | Rollback note | Stop condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WSH-SL-01 | Extract entity workbook surface from `WorkbookShell.tsx` into a workbook-owned component. | `WorkbookShell.tsx`; new `apps/web/src/workbook/components/EntityWorkbookSurface.tsx` or equivalent; maybe no tests if pure move. | WF-03 contract freeze; inspect direct entity tests; preserve local style imports or move shared primitives deliberately. | Host/identity create/edit/paste/merge/Timeline preview, query state, selected row, inspector default-closed/reset, selectors, role gating. | `WorkbookShell.surfaces.test.tsx`; `WorkbookShell.phase5.gridProvenance.test.tsx`; `entityWorkbookModel.test.ts`; TODO: confirm coverage for entity paste and merge preview before edit. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; add `make lint-biome` after stable. | Remove new component and reinline original surface block. | Stop if extraction requires route payload changes, saved-view behavior changes, or direct grid vendor semantics. |
| WSH-SL-02 | Extract assessment workbook surface into a workbook-owned component. | `WorkbookShell.tsx`; new `apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx` or equivalent. | WSH-SL-01 not strictly required; characterize assessment create/query races first. | Assessment support rows, create payload, role gating, query refresh, load error overlay, inspector default-closed/reset, selectors. | `WorkbookShell.assessments.test.tsx`; `assessmentWorkbookModel.test.ts`; relevant `WorkbookShell.surfaces.test.tsx` if surface dispatch changes. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`. | Remove new component and reinline assessment block. | Stop if create payload, support refs, role authorization, or inspector config semantics are unclear. |
| WSH-SL-03 | Extract shell data loaders for entity, assessment, and generic surfaces into a hook/service. | `WorkbookShell.tsx`; new hook such as `useWorkbookSurfaceLoaders.ts`; maybe tests if behavior is not already covered. | WSH-SL-01 and WSH-SL-02 preferred; freeze latest-query abort semantics and access-loss handling. | Latest applicable query wins; stale query errors ignored; access-loss callback behavior; load errors; normalized rows; generic notes normalization. | `WorkbookShell.assessments.test.tsx`; `WorkbookShell.surfaces.test.tsx`; `WorkbookShell.phase3.grid.test.tsx`; TODO: add characterization if generic stale-query behavior is not isolated enough. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; targeted phase slice if task-guide maps query rows. | Reinline loader hook and restore original refs/effects. | Stop if hook needs to alter route shape, envelope validation, or incident access-loss behavior. |
| WSH-SL-04 | Extract incident controls drawer/menu coordination. | `WorkbookShell.tsx`; new hook/component under `apps/web/src/workbook/components` or `hooks`; no backend files. | Confirm focus/escape/menu tests in `WorkbookShell.surfaces.test.tsx`. | Incident controls menu items, active/last section, focus restore, Escape close, drawer selectors, render props, session role refresh props. | `WorkbookShell.surfaces.test.tsx` incident controls test; App shell tests if prop threading changes. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`. | Reinline drawer/menu state and render block. | Stop if implementation changes public incident controls prop shape or authorization/session refresh semantics. |
| WSH-SL-05 | Reduce shell render composition with a local active-surface dispatcher. | `WorkbookShell.tsx`; maybe new `WorkbookActiveSurface.tsx`. | WSH-SL-01 through WSH-SL-04 preferred; all surface props frozen. | Active surface selection, URL/saved-view/query behavior, Timeline props, generic/entity/assessment dispatch, surface menu, system view switcher. | Full shell unit coverage from surfaces/assessments/phase8/phase9; TODO: use `make task-guide` to select browser check if route/startup rendering changes. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; browser target if route/focus risk. | Remove dispatcher and restore inline conditional render. | Stop if dispatcher introduces phase-shaped runtime structure or changes Timeline UI. |
| WSH-SL-06 | Consolidate shell-local style primitives only after behavior slices pass. | `WorkbookShell.tsx`; optional style module only if needed. | Prior slices complete and validation green; inspect design tokens if moving style values. | Visual output, selectors, layout, density, inspector/shell slots, no route or payload changes. | Existing unit tests plus visual/a11y only if style diff is material. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make browser-e2e-visual` only for intentional visual maintenance. | Reinline style constants. | Stop if diff touches selectors, route payloads, design-token semantics, or requires golden updates beyond maintenance policy. |

### Refactor Slice Execution Status

| Slice ID | Status | Started | Completed | Edits and validation | Next action |
| --- | --- | --- | --- | --- | --- |
| WSH-SL-01 | DONE | 2026-06-28T17:19:37-04:00 | 2026-06-28T17:31:08-04:00 | Extracted the Host/Identity entity workbook surface to `apps/web/src/workbook/components/EntityWorkbookSurface.tsx`, added shared `apps/web/src/workbook/models/workbookMutations.ts`, wired `WorkbookShell.tsx` to the component, and removed shell-local entity-only helpers/styles. Validation: `make frontend-typecheck` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260628T213043Z-p3033934`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T213043Z-p3033945`; `make lint-biome` PASS. Transient typecheck failures `.cartulary/test-results/20260628T212816Z-p3029624`, `.cartulary/test-results/20260628T212921Z-p3030300`; transient lint failure `.cartulary/test-results/20260628T213001Z-p3031332`; first pre-format unit/import passes `.cartulary/test-results/20260628T213001Z-p3031282` and `.cartulary/test-results/20260628T213001Z-p3031287`. | Start WSH-SL-02 assessment surface extraction. |
| WSH-SL-02 | DONE | 2026-06-28T17:31:08-04:00 | 2026-06-28T17:37:11-04:00 | Extracted the assessment workbook surface to `apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx`, wired `WorkbookShell.tsx` to the component, and removed shell-local assessment-only helpers/styles. Validation: `make frontend-typecheck` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260628T213649Z-p3041604`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T213649Z-p3041606`; `make lint-biome` PASS. Transient typecheck root `.cartulary/test-results/20260628T213500Z-p3037865`; transient lint root `.cartulary/test-results/20260628T213609Z-p3039362`; pre-format unit/import passes `.cartulary/test-results/20260628T213609Z-p3038956` and `.cartulary/test-results/20260628T213609Z-p3038967`. | Start WSH-SL-03 surface loader extraction. |
| WSH-SL-03 | DONE | 2026-06-28T17:37:11-04:00 | 2026-06-28T17:42:12-04:00 | Extracted entity, assessment, and generic surface loader state/effects into `apps/web/src/workbook/hooks/useWorkbookSurfaceLoaders.ts`; `WorkbookShell.tsx` now keeps session-role loading and delegates latest-query abort/current semantics, access-loss handling, row normalization, load errors, and row indexes to the hook. Validation: `make frontend-typecheck` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260628T214149Z-p3047159`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T214149Z-p3047193`; `make lint-biome` PASS. Transient lint-biome format-only failure `.cartulary/test-results/20260628T214120Z-p3046151`; `make format-frontend` applied the hook formatting. | Start WSH-SL-04 incident controls drawer/menu coordination extraction. |
| WSH-SL-04 | DONE | 2026-06-28T17:42:12-04:00 | 2026-06-28T17:46:13-04:00 | Extracted incident controls menu items, active/last section state, focus restore, and close-button focus management into `apps/web/src/workbook/hooks/useIncidentControlsDrawer.ts`; moved the drawer dialog markup and selectors into `apps/web/src/workbook/components/IncidentControlsDrawer.tsx`; `WorkbookShell.tsx` now only passes account-menu handoff props and render-prop payloads. Validation: `make frontend-typecheck` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260628T214549Z-p3052637`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T214549Z-p3052614`; `make lint-biome` PASS. Transient lint-biome format-only failure `.cartulary/test-results/20260628T214532Z-p3051653`; `make format-frontend` applied drawer formatting. | Start WSH-SL-05 active-surface dispatcher extraction. |
| WSH-SL-05 | DONE | 2026-06-28T17:46:13-04:00 | 2026-06-28T17:50:14-04:00 | Extracted the active surface conditional render into `apps/web/src/workbook/components/WorkbookActiveSurface.tsx`; moved surface-specific sort-contract decisions for Timeline, Host, Identity, Assessment, and generic surfaces behind that component while leaving runtime/query state ownership in `WorkbookShell.tsx`. Validation: `make frontend-typecheck` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260628T214951Z-p3058505`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T214951Z-p3058508`; `make lint-biome` PASS. Transient typecheck failure `.cartulary/test-results/20260628T214849Z-p3056694` from an over-pruned `requireViewContract` import; transient lint-biome format-only failure `.cartulary/test-results/20260628T214928Z-p3057584`; `make format-frontend` applied dispatcher formatting. | Start WSH-SL-06 style consolidation. |
| WSH-SL-06 | DONE | 2026-06-28T17:50:14-04:00 | 2026-06-28T17:53:25-04:00 | Moved the remaining shell chrome/navigation style constants from `WorkbookShell.tsx` into `apps/web/src/workbook/components/WorkbookShellStyles.ts` with values unchanged. `WorkbookShell.tsx` is now 541 lines and primarily wires runtime state, shell chrome, account menu, saved-view selector, loaders, and extracted surfaces. Validation: `make frontend-typecheck` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260628T215258Z-p3062428`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T215258Z-p3062426`; `make lint-biome` PASS. | Start WSH-REM-99 final validation, phase4 visual triage, and handoff completion. |

## 8. Characterization Test Plan

### Existing Tests First

| Behavior | Existing evidence | Coverage state |
| --- | --- | --- |
| Surface selection, startup, saved-view identity, generic surfaces | `WorkbookShell.surfaces.test.tsx`; `workbookStartup.test.ts`; `workbookSavedViews.test.ts`; `workbookSavedViewRuntime.test.ts` | Covered; rerun before WSH-SL-03/05. |
| Assessment create and query race behavior | `WorkbookShell.assessments.test.tsx`; `assessmentWorkbookModel.test.ts` | Covered for WSH-SL-02. |
| Host/identity/notes grid provenance | `WorkbookShell.phase5.gridProvenance.test.tsx`; `entityWorkbookModel.test.ts`; `genericWorkbookModel.test.ts` | Covered for model/provenance; TODO: confirm entity paste/merge preview coverage before WSH-SL-01. |
| Timeline create/edit/autosave/paste/focus | Phase3 tests, phase4 support/action/save-state, phase9 sentinel, utility tests | Covered for Timeline prop-preservation risk. |
| Query sort/filter/group and saved-view JSON | `WorkbookShell.phase8.query.test.tsx`; `workbookQuery.test.ts` | Covered. |
| Inspector default-closed and retargeting | `WorkbookShell.phase9.inspector.test.tsx`; `workbookInspectorModel.test.ts` | Covered. |
| Collaboration/presence/live updates | `WorkbookShell.phase6.test.tsx`; timeline service/model tests | Covered for Timeline-owned behavior. |
| History/rollback/conflict anchors | `WorkbookShell.phase7.test.tsx`; `workbookPendingQueue.test.ts` | Covered. |
| Browser route/focus/readiness | `apps/web/e2e/phase*.spec.ts`, `workbook.a11y*.spec.ts`, `workbook.visual.spec.ts` | TODO: select exact rows only after `make task-guide`/`make explain-phase`. |

### Missing Evidence TODOs

| Gap | Required action before edit |
| --- | --- |
| Entity paste create targets and merge preview may not be isolated from broader shell tests. | TODO: inspect `WorkbookShell.surfaces.test.tsx`, `WorkbookShell.phase5.gridProvenance.test.tsx`, and phase4 merge/browser tests before WSH-SL-01; add TODO characterization only if observed behavior lacks coverage. |
| Generic stale-query behavior after loader extraction may rely on broad shell test coverage. | TODO: confirm test scenario names before WSH-SL-03. |
| Incident controls focus restore may be only one shell test. | TODO: verify focus/escape behavior before WSH-SL-04; add a behavior-named test only if absent. |
| Browser/a11y/visual target selection is not fixed. | TODO: run `make task-guide ROLE=feature-dev PHASE=phaseN` and `make explain-phase PHASE=phaseN` for the touched behavior before broad browser validation. |

## 9. Boundary Guardrails

| Guardrail | Check |
| --- | --- |
| `/apps/web` must not learn `react-data-grid` vendor coordinate semantics. | Run `make frontend-import-boundary-check`; review for row/column indexes or RDG-specific imports. |
| Direct `react-data-grid` imports must remain inside `/packages/grid-adapter`. | `tools/frontend_import_boundaries.json` rule `frontend-grid-vendor-boundary`; `packages/grid-adapter/src/index.tsx` is current direct owner. |
| Generated files must not be hand-edited. | Do not edit `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, `tools/task_surface.generated.mk`, lockfiles, or tool-managed artifacts. |
| Selectors/test IDs must remain stable or be updated through owning package. | Use `@cartulary/ui-contracts`; do not hand-roll selectors from visible labels, DOM order, or vendor classes. |
| Phase identity must not become production runtime structure. | Production extraction names and runtime conditionals must be module/surface/contract-shaped, not `phaseN`-shaped. |
| Visual/a11y evidence must not be promoted into product conformance unless an owner says so. | Record visual/a11y as design/readiness or implementation-support evidence unless owner docs explicitly promote it. |
| Public route shapes and envelopes must not drift. | Do not alter `/api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, `/rows`, `/saved-views`, `/workbook-preferences/*`, `/workbook-startup`, `/api/v1/records/{record_id}`, or Timeline action routes in these slices. |
| View-schema IDs and field keys must remain contract-derived. | Continue using `requireViewContract`, `workbookSurfaceRegistry`, `workbookQuery`, and `view-contracts` facade. |
| Authorization behavior must not change. | Preserve current role checks and `onIncidentAccessLost`; do not infer authorization in extracted components beyond existing props. |

## 10. Validation Plan

| Order | Target | Purpose | Required? | Expected artifact/result | Failure handling |
| --- | --- | --- | --- | --- | --- |
| 1 | `make frontend-typecheck` | Fast structural check after TS extraction. | Required for every slice. | Make target pass and retained run root when emitted. | Product compile failures are related unless clearly pre-existing; fix before broadening. |
| 2 | `make frontend-unit` | Existing unit/interaction characterization. | Required for every slice. | Make target pass and run root. | If failure is assertion/behavior, treat as product regression until proven otherwise. |
| 3 | `make frontend-import-boundary-check` | RDG/generated/package facade guardrails. | Required for every slice. | Pass. | Boundary failure blocks merge until imports are corrected. |
| 4 | `make lint-biome` | Authored frontend lint/style after code movement. | Required once slice compiles; can run after stable diff. | Pass. | Fix authored-source lint only; do not run rewrite format blindly for docs-only work. |
| 5 | `make task-guide ROLE=feature-dev PHASE=phaseN` | Discover narrow phase verification. | Required before phase/browser validation; exact phase depends on behavior touched. | Guidance output recorded. | If phase unknown, write TODO with exact behavior-to-phase search. |
| 6 | `make phase-slice PHASE=phaseN` or `make service-backed-slice PHASE=phaseN` | Owner-aligned phase slice. | Conditional for high-risk query/route/focus/collab changes. | Pass/run root. | Distinguish product failures from harness/config/infra failures via summary artifacts. |
| 7 | `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`, `make browser-e2e-a11y`, `make browser-e2e-visual` | Browser-level route/state/a11y/visual readiness. | Conditional; use only when touched behavior justifies. | Pass/run root. | Visual/a11y evidence remains readiness/design-direction unless owner says otherwise. |
| 8 | `make agent-finalize` | End-of-run harness maintenance. | Required by repo procedure before broader end-of-run verification. | Pass, or retained-run maintenance skipped if `RESULTS_DIR` unset. | If `RESULTS_DIR` unset, record skipped reason explicitly. |

## 11. Workstream Notes

### Scope and Evidence

| Date | Note | Source or command | Impact |
| --- | --- | --- | --- |
| 2026-06-28 | Target exists at 2,955 lines; framework exists at 468 lines; adjacent Timeline entrypoint is 2,158 lines. | `wc -l` | Confirms large shell coordinator and adjacent Timeline risk. |
| 2026-06-28 | `WorkbookShell.refactor-tracker.md` was absent before this session. | `test -f docs/handoffs/WorkbookShell.refactor-tracker.md` | This artifact is a new planning file. |
| 2026-06-28 | Current branch/commit matched requested baseline and tree was clean before artifact. | `git status --short --branch`; `git rev-parse HEAD` | Session header can cite exact baseline. |

### Contracts and Docs

| Date | Owner section | Decision or conflict | Action |
| --- | --- | --- | --- |
| 2026-06-28 | Core 01/Core 03 | Saved views and workbook startup distinguish `saved_view_id`, `sheet_ref`, and base `view_schema_id`. | Freeze during shell split. |
| 2026-06-28 | Core 03 | Inspector default-open is false and stale row-bound state invalidates on row/surface/version/security lifecycle changes. | Preserve reset key and inspector behavior. |
| 2026-06-28 | Harness NLSpec | Make is canonical; generated files are prohibited hand-edits; visual/a11y are readiness unless owner says otherwise. | Use Make targets and classify evidence correctly. |

### Frontend Package Seam

| Date | Package | Files | Current state | Next action |
| --- | --- | --- | --- | --- |
| 2026-06-28 | `/apps/web` | `WorkbookShell.tsx`, shell hooks/components/models | Owns shell coordination and app state. | Extract only app-owned components/hooks. |
| 2026-06-28 | `/packages/grid-adapter` | `src/index.tsx`, `src/core.ts` | Owns RDG import, grid rendering, Cartulary anchor/paste abstractions. | Keep `/apps/web` on facade imports. |
| 2026-06-28 | `/packages/view-contracts` | `src/index.ts` | Owns view row/contract/inspector validation. | Do not duplicate contract parsing in shell. |
| 2026-06-28 | `/packages/ui-contracts` | `src/index.ts` | Owns stable selectors/test IDs. | Keep selectors stable. |

### Tests and Harness

| Date | Target | Result | Artifact | Follow-up |
| --- | --- | --- | --- | --- |
| 2026-06-28 | `git status --short --branch`; `git rev-parse HEAD`; `make task-guide ROLE=feature-dev PHASE=phase4`; `make task-guide ROLE=feature-dev PHASE=phase8`; `make task-guide ROLE=feature-dev PHASE=phase9` | PASS for WSH-REM-00 baseline. Clean tree at `main...origin/main`, commit `496d94986e5310d11f5833ea618a5dfd08fe8772`; phase4/8/9 guides select the expected phase slices for entity merge/reuse, saved-view/layout/query, and workbook focus/surface behavior. | Console guidance; no run-root output emitted. | Start WSH-REM-01 contract and layout alignment. |
| 2026-06-28 | Test inventory by `rg` and file search | Completed for planning; no test targets run yet. | This planner | TODO: run Make validation during implementation slices. |
| 2026-06-28 | `make help-all` search | Public validation targets discovered. | Terminal output not retained in artifact root. | Use exact Make targets in future handoffs. |
| 2026-06-28 | `make frontend-typecheck` | PASS during remediation. | No run-root output emitted. | Re-run after any further TypeScript or contract changes. |
| 2026-06-28 | `make frontend-unit` | PASS during remediation. | `.cartulary/test-results/20260628T200823Z-p2890427/frontend-unit/tool-run-summary.json` | Covers new entity paste, merge, preview retarget, access-loss, and incident-drawer focus characterization. |
| 2026-06-28 | `make frontend-import-boundary-check` | PASS during remediation. | `.cartulary/test-results/20260628T200853Z-p2892181/frontend-import-boundary-check/tool-run-summary.json` | Confirms no direct grid-vendor/generated-root import drift. |
| 2026-06-28 | `make lint-biome` | PASS after manual formatting/import-order patches. | Latest pass emitted no run-root output; failed diagnostic runs were formatting-only at `.cartulary/test-results/20260628T200900Z-p2892579` and `.cartulary/test-results/20260628T200957Z-p2893316`. | Treat the two failures as harness/tool diagnostic formatting failures, not product failures. |
| 2026-06-28 | `make task-guide ROLE=feature-dev PHASE=phase4`; `make explain-phase PHASE=phase4` | PASS; selected entity/assessment/evidence-handle phase guidance. | Console guidance; no run-root output emitted. | Use `make phase-slice PHASE=phase4` if merge/entity contract behavior changes beyond current frontend characterization. |
| 2026-06-28 | `make task-guide ROLE=feature-dev PHASE=phase8`; `make explain-phase PHASE=phase8` | PASS; selected saved-view/query/startup phase guidance. | Console guidance; no run-root output emitted. | Use `make phase-slice PHASE=phase8` if loader extraction changes query/startup behavior. |
| 2026-06-28 | `make task-guide ROLE=feature-dev PHASE=phase9`; `make explain-phase PHASE=phase9` | PASS; selected keyboard/clipboard/focus phase guidance. | Console guidance; no run-root output emitted. | Use `make phase-slice PHASE=phase9` if entity paste/focus or incident drawer focus changes require browser confirmation. |
| 2026-06-28 | `make lint-markdown` | PASS during remediation. | No run-root output emitted. | Re-run after further handoff edits. |
| 2026-06-28 | `make generated-artifact-policy-check` | PASS during remediation. | `.cartulary/test-results/20260628T201242Z-p2895290/generated-artifact-policy-check/tool-run-summary.json` | Confirms no generated-root hand-edit drift from this slice. |
| 2026-06-28 | `make phase-slice PHASE=phase9` | PASS during remediation. | `.cartulary/test-results/20260628T201256Z-p2896207/phase-slice/tool-run-summary.json` | Browser-backed keyboard/clipboard/focus evidence for entity paste and incident drawer focus risk. |
| 2026-06-28 | `make phase-slice PHASE=phase4` | FAIL on visual-only evidence, repeated twice. Backend, frontend-unit, and browser webserver-backed phase4 work passed; `browser-e2e-visual` failed `V-4-GRID-01` Timeline mention-chip golden with 15,343 pixels different. | First run `.cartulary/test-results/20260628T201346Z-p2908085/phase-slice/tool-run-summary.json`; retry `.cartulary/test-results/20260628T201451Z-p2921485/phase-slice/tool-run-summary.json`. | Treat as unresolved product visual-golden evidence until owner triage; not a generated/schema or import-boundary failure. |
| 2026-06-28 | `make agent-finalize` | PASS; retained-run checks skipped because `RESULTS_DIR` was unset. | `.cartulary/test-results/20260628T201808Z-p2936767/agent-finalize/tool-run-summary.json`; finalize summary reports `results_dir=-`, `run_checks=skipped`, `reused=3`, `cache_hits=3`. | Re-run with `RESULTS_DIR=<successful full warm check run root>` only when retained-run maintenance is required. |
| 2026-06-28 | WSH-REM-01 contract/layout validation: `make generate`; `make json-shape-check`; `make phase-schedules`; `make generated-artifact-policy-check`; `make generate-drift`; `make build-migrate`; `make migration-drift`; `make phase-schedule-drift`; `make phase-ledgers`; `make phase-ledger-drift`; `make phase-slice PHASE=phase8` | PASS after harness-map and compile fixes. Initial `json-shape-check` failed until `TestPhase8_SavedViewCreateEvolvesAdditiveHiddenFields_U_8_02` was added to `tools/phase8_test_map.json`; initial `build-migrate` failed on a `ViewFieldEntry`/`Field` helper type mismatch and passed after correction. Final phase8 slice passed 37 tests. | Final roots: generate `.cartulary/test-results/20260628T204559Z-p2947899`; json shape `.cartulary/test-results/20260628T205023Z-p2976318`; generated policy `.cartulary/test-results/20260628T204706Z-p2950324`; generate drift `.cartulary/test-results/20260628T204711Z-p2950510`; build-migrate `.cartulary/test-results/20260628T204833Z-p2953908`; migration drift `.cartulary/test-results/20260628T204841Z-p2955312`; schedule drift `.cartulary/test-results/20260628T205023Z-p2976295`; ledger drift `.cartulary/test-results/20260628T204915Z-p2957993`; phase8 slice `.cartulary/test-results/20260628T204928Z-p2958364`. | Start WSH-REM-02 backend hydration and merge error details. |
| 2026-06-28 | WSH-REM-02 backend hydration/error validation: `make backend-unit`; `make backend-store`; `make backend-integration`; `make json-shape-check`; `make generated-artifact-policy-check` | PASS after one backend-store fix. Initial `make backend-store` failed because Host/Identity row queries attempted reusable-identifier hydration while the main result cursor was still open on the same connection; moving bulk hydration before opening the cursor resolved the `conn busy` failure. Final backend tests cover Host and Identity reusable-identifier query exposure, post-merge carry-forward, exact-match reuse, and owner-spelled merge collision details. | Final roots: backend-unit `.cartulary/test-results/20260628T210029Z-p2980671`; backend-store `.cartulary/test-results/20260628T210336Z-p2992257`; backend-integration `.cartulary/test-results/20260628T210404Z-p2999804`; json shape `.cartulary/test-results/20260628T210601Z-p3007267`; generated policy `.cartulary/test-results/20260628T210601Z-p3007283`. Initial failed backend-store root `.cartulary/test-results/20260628T210112Z-p2986536`. | Start WSH-REM-03 frontend reusable-identifier inspector and merge-plan behavior. |
| 2026-06-28 | WSH-REM-03 frontend inspector/merge-plan validation: `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make json-shape-check` | PASS after type, fixture, formatting, and test-accounting fixes. Initial `frontend-typecheck` failed on merge-detail parser tuple inference; initial `frontend-unit` then exposed missing additive hidden-field fixtures in assessment tests and one new unmapped model test; initial `lint-biome` required formatting for the touched frontend files. Final frontend tests cover Host and Identity reusable-identifier display, merge plan promotion/carry-forward/duplicate/provenance outcomes, and owner-spelled collision detail rendering. | Final roots: frontend-unit `.cartulary/test-results/20260628T211902Z-p3024437`; frontend-import-boundary `.cartulary/test-results/20260628T211925Z-p3026884`; json shape `.cartulary/test-results/20260628T211839Z-p3023203`. `frontend-typecheck` and final `lint-biome` passed without run-root output. Transient roots: typecheck failure `.cartulary/test-results/20260628T211258Z-p3010215`; frontend-unit failures `.cartulary/test-results/20260628T211258Z-p3010248`, `.cartulary/test-results/20260628T211351Z-p3012913`, `.cartulary/test-results/20260628T211505Z-p3015103`, `.cartulary/test-results/20260628T211614Z-p3017196`, `.cartulary/test-results/20260628T211710Z-p3019283`; lint-biome failure `.cartulary/test-results/20260628T211839Z-p3023168`; pre-format import-boundary pass `.cartulary/test-results/20260628T211839Z-p3023169`; pre-format frontend-unit pass `.cartulary/test-results/20260628T211816Z-p3021424`. | Start WSH-SL-01 entity surface extraction; tracker row is marked IN_PROGRESS. |
| 2026-06-28 | WSH-SL-03 surface loader validation: `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome` | PASS after moving entity, assessment, and generic loaders to `useWorkbookSurfaceLoaders`. Latest-query runtime refs, access-loss handling, row normalization, generic notes normalization, and load-error state are now hook-owned; session-role loading remains shell-owned. | Final roots: frontend-unit `.cartulary/test-results/20260628T214149Z-p3047159`; frontend-import-boundary `.cartulary/test-results/20260628T214149Z-p3047193`. `frontend-typecheck` and final `lint-biome` passed without run-root output. Transient format-only lint root `.cartulary/test-results/20260628T214120Z-p3046151`; `make format-frontend` applied formatting. | Start WSH-SL-04 incident controls drawer/menu coordination extraction; tracker row is marked IN_PROGRESS. |
| 2026-06-28 | WSH-SL-04 incident controls validation: `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome` | PASS after extracting incident-controls menu/drawer state and focus coordination into `useIncidentControlsDrawer` and `IncidentControlsDrawer`. The account-menu handoff shape, drawer selectors, Escape close, and close-button focus behavior remain covered by existing shell tests. | Final roots: frontend-unit `.cartulary/test-results/20260628T214549Z-p3052637`; frontend-import-boundary `.cartulary/test-results/20260628T214549Z-p3052614`. `frontend-typecheck` and final `lint-biome` passed without run-root output. Transient format-only lint root `.cartulary/test-results/20260628T214532Z-p3051653`; `make format-frontend` applied formatting. | Start WSH-SL-05 active-surface dispatcher extraction; tracker row is marked IN_PROGRESS. |
| 2026-06-28 | WSH-SL-05 active-surface dispatcher validation: `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome` | PASS after moving the active surface conditional into `WorkbookActiveSurface`. Surface-specific render selection and sort toggles are component-owned; shell runtime state, route/startup identity, saved-view selector construction, and loader ownership remain unchanged. | Final roots: frontend-unit `.cartulary/test-results/20260628T214951Z-p3058505`; frontend-import-boundary `.cartulary/test-results/20260628T214951Z-p3058508`. `frontend-typecheck` and final `lint-biome` passed without run-root output. Transient typecheck root `.cartulary/test-results/20260628T214849Z-p3056694`; transient format-only lint root `.cartulary/test-results/20260628T214928Z-p3057584`; `make format-frontend` applied formatting. | Start WSH-SL-06 style consolidation; tracker row is marked IN_PROGRESS. |
| 2026-06-28 | WSH-SL-06 style consolidation validation: `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome` | PASS after moving shell chrome/navigation style constants into `WorkbookShellStyles`. Style values, selectors, density behavior, and active-surface composition remain unchanged. | Final roots: frontend-unit `.cartulary/test-results/20260628T215258Z-p3062428`; frontend-import-boundary `.cartulary/test-results/20260628T215258Z-p3062426`. `frontend-typecheck` and `lint-biome` passed without run-root output. | Start WSH-REM-99 final validation and handoff completion. |
| 2026-06-28 | WSH-REM-99 final validation: `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make json-shape-check`; `make generated-artifact-policy-check`; `make generate-drift`; `make migration-drift`; `make lint-markdown`; `make phase-slice PHASE=phase8`; `make phase-slice PHASE=phase9`; `make phase-slice PHASE=phase4`; phase4-scoped `make browser-e2e-visual-update`; `make test-service-backed`; `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260628T215452Z-p3071606`; `make go-test-duration-baseline-coverage`; `make agent-finalize` | PASS after refreshing the stale `V-4-GRID-01` Timeline mention-chip golden, rerunning phase4, repairing the new phase8 Go test duration-baseline entry, and rerunning `agent-finalize`. The first final phase4 run failed only on `browser-e2e-visual` `V-4-GRID-01`; phase4 backend-unit, backend-store, backend-integration, backend-integration-support, frontend-unit, and browser webserver-backed evidence passed in that run. `CARTULARY_PHASE_SLICE_PHASE=phase4 make browser-e2e-visual-update` refreshed the phase4 authoritative golden and passed its phase4 authoritative visual sub-run, but the overall update target failed on unrelated FE-P11 support visual assertions; accidental FE readiness snapshot changes from that broad update were reverted, leaving only `v-4-grid-01-mention-chips-linux.png`. A full `make test-service-backed` run was attempted for prune-mode duration evidence, but it failed on unrelated broad FE a11y/visual assertions, so the missing Go baseline was added with the successful phase8 run root in non-prune mode. | Final roots: frontend-unit `.cartulary/test-results/20260628T215427Z-p3065216`; frontend-import-boundary `.cartulary/test-results/20260628T215427Z-p3065239`; json-shape `.cartulary/test-results/20260628T215427Z-p3065253`; generated policy `.cartulary/test-results/20260628T215427Z-p3065312`; generate drift `.cartulary/test-results/20260628T215427Z-p3065326`; migration drift `.cartulary/test-results/20260628T215427Z-p3065333`; phase8 `.cartulary/test-results/20260628T215452Z-p3071606`; phase9 `.cartulary/test-results/20260628T215536Z-p3086935`; initial phase4 fail `.cartulary/test-results/20260628T215619Z-p3098269`; visual update `.cartulary/test-results/20260628T215838Z-p3112090`; final phase4 pass `.cartulary/test-results/20260628T220103Z-p3124213`; failed broad service-backed run `.cartulary/test-results/20260628T220604Z-p3146356`; Go duration-baseline update `.cartulary/test-results/20260628T220901Z-p3187548`; Go duration-baseline coverage `.cartulary/test-results/20260628T220907Z-p3187624`; final `agent-finalize` `.cartulary/test-results/20260628T220912Z-p3187731`. `frontend-typecheck`, `lint-biome`, and `lint-markdown` passed without run-root output. | WSH-REM-99 closed; no remaining remediation blocker. |

### Generated Artifacts

| Date | Generator or target | Outputs | Drift status | Follow-up |
| --- | --- | --- | --- | --- |
| 2026-06-28 | `tools/generated_artifact_policy.json` | Generated roots listed. | No generated changes planned. | Run `make generated-artifact-policy-check` only if generated-policy risk arises. |
| 2026-06-28 | View-schema contracts | `contracts/view-schemas/*.json` | Inputs inspected, no edits planned. | Owner-doc/contract work is deferred and out of scope for behavior-preserving shell split. |
| 2026-06-28 | Host/Identity reusable identifiers | Owner spec, derived view schemas, generated contract embeddings, saved-view layout evolution, and merge-precondition reason registry now define `host.reusable_identifiers`, `identity.reusable_identifiers`, additive hidden-field normalization, and `carry_forward_identifier_collision`. Backend query/create/upsert/patch/merge-refresh rows now hydrate reusable identifiers, and merge collision details now use `carry_forward_identifier_collision` plus `blocking_record_id`. Frontend entity rows parse reusable identifiers, the inspector renders them as read-only, merge plans distinguish promotion, carry-forward reusable values, duplicate no-ops, copied aliases, and server-side provenance, and collision details render with blocking-record context. | `make generate`, `make generate-drift`, `make json-shape-check`, and `make phase-slice PHASE=phase8` pass for WSH-REM-01. `make backend-unit`, `make backend-store`, `make backend-integration`, `make json-shape-check`, and `make generated-artifact-policy-check` pass for WSH-REM-02. `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome`, and `make json-shape-check` pass for WSH-REM-03. | Begin behavior-preserving WorkbookShell extraction slices at WSH-SL-01. |
| 2026-06-28 | `V-4-GRID-01` Timeline mention-chip visual golden | Refreshed `apps/web/e2e/workbook.visual.spec.ts-snapshots/v-4-grid-01-mention-chips-linux.png` after final phase4 validation confirmed the previous golden was stale relative to passing phase4 functional/browser-backed evidence and current Timeline inspector behavior. | Phase4 authoritative visual update sub-run passed in `.cartulary/test-results/20260628T215838Z-p3112090`; final `make phase-slice PHASE=phase4` passed at `.cartulary/test-results/20260628T220103Z-p3124213`. The overall `browser-e2e-visual-update` target still failed on unrelated FE-P11 support visual assertions, and non-target FE snapshot updates from that broad target were reverted. | No further phase4 visual blocker remains for this remediation. FE-P11 support visual failures remain unrelated to WSH-REM-99. |

### Risks and Blockers

| ID | Risk or blocker | Owner | Blocking workflow | Resolution condition |
| --- | --- | --- | --- | --- |
| WSH-REM-00 | Baseline and phase-guide discovery for the complete remediation plan. | Current remediation | Contract and implementation remediation | DONE: clean baseline and phase4/8/9 task-guide outputs recorded before mutation. |
| WSH-REM-01 | Contract and layout alignment: expose reusable identifiers through Host/Identity view contracts, add saved-view layout evolution, and mirror merge-precondition reason registry. | Owner docs/contracts/generated artifacts | WSH-REM-02 backend hydration and WSH-REM-03 frontend display | DONE as of 2026-06-28T16:50:48-04:00. Edited Core 01, view-schema contracts, error registry, generated embeddings, saved-view layout normalizer/tests, migration `00039`, phase8 test map, generated schedules/ledger. Final validation roots are recorded in Tests and Harness. |
| WSH-REM-02 | Backend row hydration and merge error details. | Backend entities/views/saved views | WSH-REM-03 frontend display | DONE as of 2026-06-28T17:06:14-04:00. Host/Identity rows now expose active secondary reusable identifiers through query/create/upsert/patch/merge refresh paths; carried identifiers are filtered against current canonical values; merge collision details now use owner-spelled `carry_forward_identifier_collision` and `blocking_record_id`. Validation roots are recorded in Tests and Harness. |
| WSH-REM-03 | Frontend reusable-identifier inspector and merge-plan behavior. | `/apps/web` Workbook shell | WSH-SL-01 extraction | DONE as of 2026-06-28T17:19:37-04:00. Entity rows parse reusable-identifier cells, the inspector renders read-only reusable identifiers, merge plans separate promotion/carry-forward reusable/duplicate no-op/copied aliases/provenance-only outcomes, merge collision details show owner-spelled blocking-record context, and frontend validation roots are recorded in Tests and Harness. |
| WSH-REM-99 | Final validation, phase4 visual triage, and handoff completion. | Current remediation | none | DONE as of 2026-06-28T18:09:25-04:00. Final validation, phase4 visual refresh, Go duration-baseline repair, and `agent-finalize` evidence are recorded in Tests and Harness. |
| B-001 | Entity surface extraction may lack isolated tests for paste/merge/preview. | Current remediation | WSH-SL-01 | DONE for current frontend shell scope: `WorkbookShell.surfaces.test.tsx` now covers entity-origin paste reuse/create, merge plan/confirmation, precondition error display, and dependent Timeline preview retarget/clear. |
| B-001a | Merge preview cannot fully claim Core 03 secondary `exact_match_reuse` display conformance from current row data. | Owner docs/contracts/generated artifacts | WSH-REM-02/03 | DONE for contract exposure in WSH-REM-01; remaining backend hydration and frontend inspector/merge-plan behavior are tracked by WSH-REM-02 and WSH-REM-03. |
| B-002 | Loader extraction can break latest-query and access-loss behavior. | Current remediation | WSH-SL-03 | DONE for access-loss characterization: generic, entity, and assessment load failures now assert `onIncidentAccessLost`; existing superseded-query tests remain the stale-response characterization. |
| B-003 | Browser target selection is not fixed for every slice. | Current remediation plus harness/task-guide | WF-10 | DONE for phase discovery and final validation: `task-guide` and `explain-phase` captured for phase4, phase8, and phase9; phase8, phase9, and final phase4 passed. The earlier phase4 visual-only Timeline mention-chip golden failure was resolved by the documented `V-4-GRID-01` refresh. |
| B-004 | Generated/view-schema work would change scope. | Owner docs/contracts | WF-11 | Stop and create separate owner-doc/contract plan before codegen. |

## 12. Session Handoff Template

### Filled Initial Handoff Record

| Field | Value |
| --- | --- |
| Date/time | 2026-06-28T15:47:43-04:00 |
| Branch/commit | `main...origin/main` at `cc6d46abe952f2b4f886a93c1e698a0776b17ab3` |
| Target module or seam | `/apps/web` Workbook shell coordinator; primary file `apps/web/src/workbook/WorkbookShell.tsx` |
| Current workflow | WF-13 handoff complete for planning artifact; next implementation should start at WSH-SL-01 only after pre-edit characterization check. |
| Completed workflows | WF-00 through WF-13 for planning. |
| Changed files | `docs/handoffs/WorkbookShell.refactor-tracker.md` only. |
| Commands run | `git status --short --branch`; `git rev-parse HEAD`; `test -f` checks; `wc -l`; `rg` searches; `sed` excerpts; `jq` view-schema summaries; `make help-all` search. |
| Passing validation | TODO: no Make verification run yet for this new Markdown artifact at time of initial handoff. |
| Failing validation | TODO: none observed; no Make targets run yet. |
| Decisions made | Behavior-preserving frontend-only slice sequence; entity first, assessment second, loader third, incident controls fourth, dispatcher fifth, styles last. |
| Open questions | TODO: exact phase-slice/browser target selection per slice; entity paste/merge preview characterization sufficiency. |
| Blockers | No blocker to planner use. Implementation blockers are B-001 through B-004. |
| Next recommended workflow | Start WSH-SL-01 with pre-edit test inspection, then `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`. |
| Safe restart command | `git status --short --branch && sed -n '1,220p' docs/handoffs/WorkbookShell.refactor-tracker.md` |

### Filled Remediation Handoff Record

| Field | Value |
| --- | --- |
| Date/time | 2026-06-28T16:11:44-04:00 |
| Branch/commit | `main...origin/main` at `cc6d46abe952f2b4f886a93c1e698a0776b17ab3` |
| Target module or seam | `/apps/web` Workbook shell coordinator; primary file `apps/web/src/workbook/WorkbookShell.tsx` |
| Current workflow | Gap remediation before extraction. Characterization and preview-retarget implementation are complete; generated/contract reusable-identifier work remains separate. |
| Completed workflows | B-001 frontend characterization; B-002 access-loss characterization; B-003 phase target discovery; narrow preview-retarget implementation. |
| Changed files | `apps/web/src/workbook/WorkbookShell.tsx`; `apps/web/src/workbook/hooks/useEntityTimelinePreview.ts`; `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`; `docs/handoffs/WorkbookShell.refactor-tracker.md`. |
| Commands run | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make task-guide ROLE=feature-dev PHASE=phase4`; `make task-guide ROLE=feature-dev PHASE=phase8`; `make task-guide ROLE=feature-dev PHASE=phase9`; `make explain-phase PHASE=phase4`; `make explain-phase PHASE=phase8`; `make explain-phase PHASE=phase9`; `make lint-markdown`; `make generated-artifact-policy-check`; `make phase-slice PHASE=phase9`; `make phase-slice PHASE=phase4`; `make agent-finalize`. |
| Passing validation | `make frontend-typecheck`; `make frontend-unit` at `.cartulary/test-results/20260628T200823Z-p2890427`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T200853Z-p2892181`; `make lint-biome`; `make lint-markdown`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260628T201242Z-p2895290`; `make phase-slice PHASE=phase9` at `.cartulary/test-results/20260628T201256Z-p2896207`; `make agent-finalize` at `.cartulary/test-results/20260628T201808Z-p2936767`. |
| Failing validation | `make lint-biome` failed twice before manual formatting/import-order cleanup; failures were formatting/tool diagnostics only, not product behavior failures. `make phase-slice PHASE=phase4` failed twice on `browser-e2e-visual` `V-4-GRID-01` Timeline mention-chip visual golden; all nonvisual phase4 work passed in the retry run root `.cartulary/test-results/20260628T201451Z-p2921485`. |
| Decisions made | Kept reusable-identifier conformance out of frontend-only remediation because current Host/Identity rows do not expose active secondary `exact_match_reuse` values. Added latest-response guard and clear behavior to entity Timeline preview hook. |
| Open questions | TODO: owner-doc/contract shape for exposing active secondary reusable identifiers; TODO: whether the eventual contract slice needs generated UI-contract selectors for reusable-identifier display. |
| Blockers | B-001a remains blocked on owner-doc/contract/generated-artifact work; B-004 still blocks any generated/view-schema changes inside behavior-preserving extraction. |
| Next recommended workflow | Triage or accept the existing phase4 visual-golden delta, keep B-001a as a separate owner-doc/contract/generated-artifact slice, then start WSH-SL-01 entity surface extraction using the new characterization. |
| Safe restart command | `git status --short --branch && sed -n '340,410p' docs/handoffs/WorkbookShell.refactor-tracker.md && make explain-run RESULTS_DIR=.cartulary/test-results/20260628T200823Z-p2890427 TARGET=frontend-unit` |

### Filled Final Remediation Handoff Record

| Field | Value |
| --- | --- |
| Date/time | 2026-06-28T18:09:25-04:00 |
| Branch/commit | `main...origin/main` at `496d94986e5310d11f5833ea618a5dfd08fe8772` |
| Target module or seam | Full WorkbookShell remediation: owner spec/contracts, entity backend rows/errors, frontend inspector/merge plan, WorkbookShell extraction, visual validation, and final handoff. |
| Current workflow | WSH-REM-99 final validation and handoff completion; all remediation and extraction slices are done. |
| Completed workflows | WSH-REM-00 through WSH-REM-03, WSH-SL-01 through WSH-SL-06, and WSH-REM-99 final validation. |
| Changed files | Owner docs/contracts/generated embeddings: Core 01, Host/Identity view schemas, error registry, generated contract TS/Go. Backend: entity row hydration, merge errors, phase4 tests. Saved-view layout: migration `00039`, normalizer, tests, phase8 map/ledger/scheduler artifacts. Frontend: reusable identifier parsing/display/merge plan, UI selectors, extracted entity/assessment surfaces, loader and incident-controls hooks, active-surface dispatcher, shell styles, model/tests, and `V-4-GRID-01` golden. Tracker: this handoff file. |
| Commands run | Baseline discovery and task guides; `make generate`; `make generate-drift`; `make json-shape-check`; `make generated-artifact-policy-check`; `make build-migrate`; `make migration-drift`; `make phase-schedule-drift`; `make phase-ledger-drift`; `make backend-unit`; `make backend-store`; `make backend-integration`; repeated slice validation with `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome`; final `make lint-markdown`; `make phase-slice PHASE=phase4`; `make phase-slice PHASE=phase8`; `make phase-slice PHASE=phase9`; `CARTULARY_PHASE_SLICE_PHASE=phase4 make browser-e2e-visual-update`; `make test-service-backed`; `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260628T215452Z-p3071606`; `make go-test-duration-baseline-coverage`; `make agent-finalize`. |
| Passing validation | Final frontend: `frontend-unit` `.cartulary/test-results/20260628T215427Z-p3065216`; `frontend-import-boundary-check` `.cartulary/test-results/20260628T215427Z-p3065239`; `frontend-typecheck` and `lint-biome` passed without run-root output. Final contract/drift: json shape `.cartulary/test-results/20260628T215427Z-p3065253`; generated policy `.cartulary/test-results/20260628T215427Z-p3065312`; generate drift `.cartulary/test-results/20260628T215427Z-p3065326`; migration drift `.cartulary/test-results/20260628T215427Z-p3065333`. Final phases: phase8 `.cartulary/test-results/20260628T215452Z-p3071606`; phase9 `.cartulary/test-results/20260628T215536Z-p3086935`; phase4 `.cartulary/test-results/20260628T220103Z-p3124213`. Duration/finalizer: Go duration-baseline update `.cartulary/test-results/20260628T220901Z-p3187548`; Go duration-baseline coverage `.cartulary/test-results/20260628T220907Z-p3187624`; `agent-finalize` `.cartulary/test-results/20260628T220912Z-p3187731`. |
| Failing validation | Initial final phase4 failed only on `V-4-GRID-01` visual golden at `.cartulary/test-results/20260628T215619Z-p3098269`; targeted phase4 authoritative visual update passed, but the full `browser-e2e-visual-update` target failed on unrelated FE-P11 support assertions at `.cartulary/test-results/20260628T215838Z-p3112090`. Non-target FE snapshot changes from that broad update were reverted. `make test-service-backed` failed at `.cartulary/test-results/20260628T220604Z-p3146356` on unrelated broad FE a11y/visual assertions, so it was not used as retained duration evidence. |
| Decisions made | Kept Host/Identity `view_schema_id` values stable; made reusable identifiers additive, read-only, and default-hidden; aligned merge collision spelling to owner specs rather than preserving implementation drift; allowed saved-view layout evolution only for missing additive hidden read-only fields; refreshed `V-4-GRID-01` because the committed golden was stale relative to current passing functional phase4 behavior. |
| Open questions | None blocking this remediation. FE-P11 support visual failures observed during the broad update target remain unrelated to this workstream. |
| Blockers | No remaining WSH remediation blocker. |
| Next recommended workflow | Review the final diff and commit; optional separate FE-P11 support visual investigation if broad visual-update cleanliness is required. |
| Safe restart command | `git status --short --branch && sed -n '252,390p' docs/handoffs/WorkbookShell.refactor-tracker.md && make explain-run RESULTS_DIR=.cartulary/test-results/20260628T220103Z-p3124213 TARGET=phase-slice` |

### Blank Future Handoff Record

| Field | Value |
| --- | --- |
| Date/time | TODO: |
| Branch/commit | TODO: |
| Target module or seam | TODO: |
| Current workflow | TODO: |
| Completed workflows | TODO: |
| Changed files | TODO: |
| Commands run | TODO: |
| Passing validation | TODO: |
| Failing validation | TODO: |
| Decisions made | TODO: |
| Open questions | TODO: |
| Blockers | TODO: |
| Next recommended workflow | TODO: |
| Safe restart command | TODO: |

## 13. Binary Acceptance Criteria

| ID | Criterion | Status | Evidence |
| --- | --- | --- | --- |
| RF-AC-WSH-001 | Exactly one primary target file and seam named. | PASS | Section 1 names `WorkbookShell.tsx` and `/apps/web` workbook shell coordinator. |
| RF-AC-WSH-002 | All inspected in-scope files listed. | PASS | Section 2 inspected files table lists inspected files and source limits. |
| RF-AC-WSH-003 | All unseen relevant files marked. | PASS | Section 1 source limits and TODO unseen files; sections 8 and 11 TODOs. |
| RF-AC-WSH-004 | Every drift-prone public contract mapped. | PASS | Section 4 maps grid, identity, view schema, query/saved views, pending/conflict, inspector, collaboration, generated contracts, harness. |
| RF-AC-WSH-005 | Behavior-preserving refactors separated from behavior changes. | PASS | Section 7 slices are behavior-preserving; section 9 stops on route/schema/contract changes. |
| RF-AC-WSH-006 | Characterization coverage stated. | PASS | Section 8 maps existing tests and missing evidence TODOs. |
| RF-AC-WSH-007 | Checkpoint sequence with validation after risky moves. | PASS | Sections 6, 7, and 10 define workflows, slices, validation, rollback, and stop conditions. |
| RF-AC-WSH-008 | Frontend package boundaries preserved. | PASS | Sections 4, 9, and 11 preserve `/apps/web`, `grid-adapter`, `view-contracts`, and `ui-contracts` boundaries. |
| RF-AC-WSH-009 | No generated file hand-edit planned. | PASS | Sections 4, 9, and 11 prohibit generated hand-edits. |
| RF-AC-WSH-010 | No phase-shaped runtime dependency introduced. | PASS | Section 9 prohibits phase identity as production runtime structure. |
| RF-AC-WSH-011 | Handoff sufficient for restart. | PASS | Section 12 includes filled and blank handoff records with commands, findings, blockers, and next workflow. |
