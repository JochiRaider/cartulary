# apps/web/src Workbook Refactor Handoff Tracker

This is a live implementation-support tracker for refactoring `apps/web/src/`.
It is not product-conformance evidence and does not change runtime behavior,
public route contracts, generated artifacts, lockfiles, or Core documents.

Governance sources:

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`
- `docs/domain.md`
- `docs/design.md`
- `docs/guides/cartulary-ui-ux-design-guide.md`

Use `temp/fallow-report.json` as a triage signal only. Reconcile each finding
against source behavior before acting. Existing behavior should be retained only
when it has continuing value under the Core documents and implementation
conformance corpus.

# Responsibility map

| Area | Symbols / approximate lines | Current responsibility | Main risk | Extraction seam |
| --- | --- | --- | --- | --- |
| Workbook shell composition | `WorkbookShell.tsx:3786-4863` | Assembles runtime state, shell controls, surface routing, active contract, incidents, role controls, saved-view selector, system-view switcher, and surface components. | Composition is tangled with data loading, URL/startup decisions, saved-view CRUD, and focus behavior. | Keep the shell as an assembler; extract startup/saved-view/query/session hooks and surface-specific containers. |
| Saved-view, startup, and query runtime | `WorkbookShell.tsx:3931-4483`, `4516-4573`; `workbookSavedViews.ts`, `workbookSavedViewRuntime.ts`, `workbookStartup.ts`, `workbookQuery.ts` | Tracks surface query state, saved-view state, home/default preferences, active saved-view selection, latest-query guards, and load effects. | Saved-view state boundaries can drift into client-local selection, focus, scroll, or popover state, which Core 03 excludes from saved views. | Extract `useWorkbookSurfaceQueries`, `useSavedViewActions`, and `useWorkbookStartupSelection` over existing pure saved-view/query modules. |
| Entity workbook surface | `EntityWorkbookSurface` at `WorkbookShell.tsx:918-1557`; `handleEntityPaste` around `1008`; `buildMergePlan` around `840-916` | Owns Hosts and Identities grids, inline edit, paste to clipboard route, selected record state, timeline preview, merge candidate/reason/message, and merge review UI. | Entity edit, paste, preview, and merge flows are co-located with grid markup and mutation state. | Extract entity row/merge pure model, `useEntityWorkbookSurface`, entity grid component, merge inspector, and paste action wrapper. |
| Assessment workbook surface | `AssessmentWorkbookSurface` at `WorkbookShell.tsx:1559-1999` | Owns assessment rows, create draft, support-row lookup from Timeline, role-gated create affordance, grid focus, and submit mutation. | Smaller than generic/entity surfaces but still mixes support-reference loading, draft state, create validation, and grid rendering. | Extract assessment create model/hook after generic/entity seams are stable; keep as lower-priority surface split. |
| Generic workbook surface | `GenericWorkbookSurface` at `WorkbookShell.tsx:2001-3247`; `refreshReferenceOptions` around `2097`; `evidenceActionsColumn` around `2359`; `renderCell` around `2367` | Owns required and optional non-Timeline surfaces, create/edit draft state, reference options, party links, evidence access, evidence attach/preview/download, task lifecycle, decision supersede, and generic grid columns. | A single generic component hides surface-specific mutation flows and broad reference loading behind one state/effect boundary. | Extract reference options hook/cache, generic mutation hook, evidence action controls, party/task/decision controls, and surface-specific panels. |
| Generic mutation and display helpers | `GenericMutationControl` at `3293-3473`; `genericCellLabel` around `3475`; `genericCreateMinimumMessage` around `3537`; `parseMutationError` around `3702`; width and cell-content helpers around `393`, `603` | Maps view-field contracts to labels, edit controls, minimum create messages, widths, display values, and user-facing errors. | Branch count is high, but much of it is contract-driven and not inherently bad; risk is untested mapping drift. | Move to pure model/presentation modules with branch coverage by read kind, write kind, enum, reference, and collection shape. |
| Timeline pure helpers and model shaping | `TimelineWorkbook.tsx:208-1280`; `workbookTimelineModel.ts`; `timelineRowsModel.ts`; `timelineConflictModel.ts`; `workbookClipboard.ts`; `workbookKeyboard.ts`; `workbookContinuity.ts`; `workbookPresence.ts` | Defines row/history/conflict payload shapes, cell readers, patch builders, collection actions, scalar and collection save intents, evidence payload shaping, paste payload shaping, WebSocket path helpers, and presence keys. | Pure logic is mixed with a large component file, making behavior hard to test or reuse even where adjacent model modules already exist. | Move row normalization, readers, patch intent builders, history normalization, and evidence payload helpers into pure modules. |
| Timeline runtime orchestration | `TimelineWorkbook` at `TimelineWorkbook.tsx:1281-5521`; existing hooks `useTimelineRows`, `useTimelineConflicts`, `useTimelineMentions`, `useTimelineEvidenceActions`, `useTimelineLiveUpdates`, `useTimelinePendingSaves`, `useTimelineGridInteractions`, `useTimelineWorkbookRuntime` | Owns rows, pending save queue, committed row versions, stale refresh, WebSocket lifecycle, sequence gaps, same-field conflicts, mention actions, history/rollback, evidence attach, paste, keyboard navigation, focus/scroll continuity, and presence updates. | Critical interaction behavior is spread through effects, refs, callbacks, and inline mutation flows inside one component. | Extract runtime hooks with typed inputs/outputs while preserving existing pure queue and socket lifecycle model boundaries. |
| Timeline rendering, inspector, conflict, and status | `TimelineWorkbook.tsx:5523-6510`; `TimelineGridSurface.tsx`; `TimelineWorkbookGrid.tsx`; `TimelineConflictResolver.tsx`; `TimelineWorkbookInspector.tsx`; `TimelineCellEditors.tsx`; `TimelineEvidencePanel.tsx`; `TimelineHistoryPanel.tsx`; `TimelineMentionsPanel.tsx`; `TimelineRowActions.tsx`; `WorkbookStatusStrip.tsx` | Builds grid columns, cell editors, row actions, conflict resolver, inspector sections, evidence/history/mention panels, auto-resolution notices, pending-save notices, and save-state strip. | Markup and view-model construction still carry behavior decisions such as conflict focus, paste navigation, evidence state, and inspector action wiring. | After hook boundaries stabilize, extract column/cell render builders, notices, presence markers, and inspector section containers. |
| Extracted workbook shell components | `ActiveSurfaceSavedViewSelector.tsx`, `SystemViewSwitcher.tsx`, `WorkbookGridControls.tsx`, `WorkbookShellSlots.tsx`, `IncidentAdminPanel.tsx` | Provides shell controls, saved-view selector, system-view switcher, grid control strip, and admin panel wiring. | Components can become implicit owners of saved-view or system-view semantics if behavior leaks into presentation. | Keep them presentational or narrowly stateful; route saved-view/system-view decisions through shell hooks. |
| Pure models and adapters | `workbookPendingQueue.ts`, `workbookSocketLifecycle.ts`, `workbookReferenceOptions.ts`, `workbookContractRows.ts`, `genericWorkbookModel.ts`, `workbookEvidence.ts`, `evidenceLifecycleViewModel.ts`, `workbookApi.ts` | Provides queue model, socket lifecycle reducer, reference option helpers, contract-row helpers, generic row model, evidence upload/view-model helpers, and API helpers. | New extractions could duplicate or bypass these boundaries. | Reuse these modules first; add narrowly named pure modules only when the existing module would become unclear. |
| Test support and regression suites | `WorkbookShell.phase*.test.tsx`, `WorkbookShell.surfaces.test.tsx`, `WorkbookShell.assessments.test.tsx`, `timelineWorkbookTestSupport.ts`, `appShellTestSupport.ts`, `fetchMockTestSupport.ts` | Covers workbook phases, surface behavior, Timeline interaction, app-shell support, and fetch mocking. | Refactors without pinned coverage can silently change Core 03 hot-path behavior. | Start each sprint by identifying or adding targeted tests before moving behavior. |

# Fallow reconciliation

Fallow is not an instruction list. Each finding below is reconciled against
current code behavior and grouped by root cause.

| Finding/group | Actual code interpretation | Classification | Reason | Action |
| --- | --- | --- | --- | --- |
| Oversized workbook surfaces: `EntityWorkbookSurface` critical at line 918, `GenericWorkbookSurface` critical at line 2001, `WorkbookShell` high at line 3786 | These symbols are genuinely large React orchestration units. They own data loading, mutation flows, selection/focus state, grid behavior, and markup. | Valid refactor target | The report aligns with source risk: excessive responsibility and hard-to-test state/effect boundaries. | Prioritize sprints 1 through 3. Pin behavior, then extract pure models, hooks, and presentational/surface containers. |
| Reference and saved-view flows: `refreshReferenceOptions` moderate at line 2097, `loadSavedViews` moderate at line 4434 | Reference loading spans many workbook surfaces; saved-view loading coordinates API data with startup/default and active query state. | Valid refactor target | Complexity reflects unclear data flow and broad effect scope, not just local branching. | Extract `useGenericReferenceOptions` and `useSavedViewActions`; preserve Core 03 saved-view boundaries and latest-query guards. |
| Contract-driven display/control branching: `GenericMutationControl`, `entityContractColumnWidth`, `genericCellLabel`, `genericCreateMinimumMessage`, `parseMutationError`, `entityCellContent`, `identifierLines` | Branching mostly maps stable contracts, closed vocabularies, and display states to controls, labels, widths, and messages. | Mostly acceptable contract branching | Switches and conditionals are expected here, but coverage and locality are weak. | Move to pure model/presentation helpers when touched; add focused tests rather than suppressing complexity by default. |
| Behavior mixed into markup: `evidenceActionsColumn`, nested `renderCell`, `handleEntityPaste` | Evidence controls, grid cell rendering, and entity paste are behavior-bearing callbacks embedded in component bodies. | Valid refactor target | These flows affect Core workbook hot paths and are difficult to test while embedded in markup. | Extract evidence action view models/components and paste action helpers after characterization tests. |
| Smaller but real surface weight: `AssessmentWorkbookSurface` line-count finding at line 1559 | The assessment surface is coherent but still owns support-row lookup, draft validation, create mutation, grid focus, and render. | Valid but lower priority | It is large enough to split, but less tangled than generic/entity surfaces. | Defer until generic/entity seams are stable; extract create hook and presentational panel if tests show value. |
| Line-count-only model helpers: `entityRowFromApi` and nearby row mapping | Entity row mapping is behavior-free but long because it normalizes a broad row shape. | Valid but lower priority | Extraction improves testability without changing behavior, but it is not the primary risk. | Move into `entityWorkbookModel.ts` during sprint 2 with row-mapping fixtures. |
| Manual non-Fallow target: `TimelineWorkbook.tsx` at 6815 lines | The file is larger than any reported symbol and owns Timeline rows, queueing, WebSocket updates, conflicts, mentions, evidence, history, paste, keyboard, presence, rendering, and status. | Valid refactor target | Fallow only reported `WorkbookShell.tsx`; manual inspection shows higher-value refactor seams in Timeline. | Treat as sprints 4 and 5 after Workbook shell behavior is pinned. Preserve existing extracted hooks/models. |

# Refactor objectives

Update the status cells in place while executing the refactor.

| Sprint | Refactor objectives | Implementation status | Validation status |
| --- | --- | --- | --- |
| 1. Characterize behavior and pin regression coverage | Record current Core 03 behavior, identify coverage for workbook open/default surface, views, saves, conflicts, paste, keyboard, grouping, evidence, mentions, history, rollback, and presence. | Not started | Not run |
| 2. Extract pure utilities and model helpers | Move behavior-free row normalization, cell readers, patch builders, labels, widths, messages, merge planning, and mutation error parsing out of large components. | Not started | Not run |
| 3. Decompose `WorkbookShell` plus entity/generic/assessment surfaces | Reduce shell and non-Timeline surfaces to composition over hooks and focused controls. | Not started | Not run |
| 4. Extract Timeline runtime hooks and mutation/live-update flows | Isolate Timeline pending queue orchestration, WebSocket/live updates, conflicts, history, mentions, evidence attach, paste, keyboard, focus, and presence. | Not started | Not run |
| 5. Extract Timeline presentation seams and finalize shell cleanup | Move Timeline render builders, notices, presence markers, inspector sections, and status wiring into focused components; leave `TimelineWorkbook.tsx` as an assembler. | Not started | Not run |

# Ordered implementation plan by sprint

| Sprint | Purpose | Files likely touched | Work requirements | Validation requirements |
| --- | --- | --- | --- | --- |
| 1 | Establish behavior baseline before movement. | `APPS_WEB_SRC_REFACTOR_HANDOFF.md`, `apps/web/src/*test*.tsx`, `timelineWorkbookTestSupport.ts`, `appShellTestSupport.ts`, `fetchMockTestSupport.ts` | Characterize behavior, identify or add focused tests, and mark gaps in this tracker. Do not move production logic. | Relevant existing `WorkbookShell.phase*.test.tsx` suites, `make frontend-unit`, `make frontend-typecheck` if tests change. |
| 2 | Create pure, testable utility/model boundaries. | New or existing pure modules under `apps/web/src/`, plus focused tests. | Extract only behavior-free helpers first. No React state, DOM, fetch, WebSocket, or grid markup in pure modules. | Unit tests for extracted modules, then `make frontend-unit` and `make frontend-typecheck`. |
| 3 | Split shell and non-Timeline surface orchestration. | `WorkbookShell.tsx`, shell hooks, entity/generic/assessment modules, presentational controls. | Extract hooks after state/effect boundaries are clear; extract presentational controls after props stabilize; keep Core saved-view and startup semantics intact. | `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase8`, `make phase-slice PHASE=phase9`, plus frontend checks. |
| 4 | Split Timeline runtime behavior from render tree. | `TimelineWorkbook.tsx`, Timeline hooks, pure Timeline modules, hook tests. | Extract runtime hooks around existing queue/socket/presence/conflict models. Hooks expose typed state and commands and must not own grid markup. | Timeline phase tests, `make phase-slice PHASE=phase3`, `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase8`, `make phase-slice PHASE=phase9`. |
| 5 | Split Timeline presentation and finish cleanup. | `TimelineWorkbook.tsx`, Timeline UI components, inspector/grid/status modules, visual or browser tests when layout changes. | Extract render builders and surface-specific UI after runtime props stabilize. Confirm no retained behavior conflicts with Core documents. | Frontend checks, targeted browser E2E, accessibility preflight when controls/focus change, and `make agent-finalize`. |

## Sprint 1 - Characterize behavior and pin regression coverage

Governing row: sprint 1 in the table above.

Execution order:

1. Characterize behavior against Core 03 and domain vocabulary before editing production code.
2. Add or identify tests for gaps that would make movement unsafe.
3. Do not extract pure utilities in this sprint unless a test helper requires it.
4. Do not extract hooks in this sprint.
5. Do not extract presentational components in this sprint.
6. Do not extract surface-specific components in this sprint.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Tracker status | `APPS_WEB_SRC_REFACTOR_HANDOFF.md` | Sprint status, behavior inventory, validation notes, risk decisions. | Product conformance claims or Core requirements. | Inputs: source inspection and test results. Outputs: updated status rows and checklist marks. |
| Workbook regression tests | `apps/web/src/WorkbookShell.phase*.test.tsx` | Existing and new assertions for workbook behavior. | Production behavior or generated fixtures. | Inputs: current UI behavior and mocked API responses. Outputs: pinned tests for moved code. |
| Timeline test support | `apps/web/src/timelineWorkbookTestSupport.ts` | Shared Timeline render, event, and mock setup. | Runtime-only behavior hidden from tests. | Inputs: existing test call sites. Outputs: stable setup helpers. |
| Shell test support | `apps/web/src/appShellTestSupport.ts`, `apps/web/src/fetchMockTestSupport.ts` | App-shell and fetch mocking utilities. | Saved-view or surface semantics. | Inputs: mocked route responses. Outputs: reusable fixtures for shell tests. |

## Sprint 2 - Extract pure utilities and model helpers

Governing row: sprint 2 in the table above.

Execution order:

1. Characterize each helper by current callers and test coverage before moving it.
2. Add or pin unit tests for row mapping, labels, widths, messages, merge plans, and patch shaping.
3. Extract pure utilities only after tests describe current behavior.
4. Do not extract hooks in this sprint except to remove no-op wrapper code created by a pure extraction.
5. Do not extract presentational components in this sprint.
6. Do not split full surfaces in this sprint.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Timeline row model helpers | `apps/web/src/timelineWorkbookRows.ts` or existing `timelineRowsModel.ts` | Timeline row normalization, cell readers, tag readers, row-version and patch-shape helpers. | Fetching, React state, DOM, WebSocket, grid renderers. | Inputs: API row envelopes and field keys. Outputs: normalized rows and patch intents. |
| Entity workbook model | `apps/web/src/entityWorkbookModel.ts` | Entity row mapping, entity cell display data, merge-plan shaping, entity column-width map if kept pure. | Entity paste, merge submission, timeline preview loading, React rendering. | Inputs: entity API rows and merge draft state. Outputs: normalized entity rows and merge view models. |
| Generic surface model | `apps/web/src/genericWorkbookSurfaceModel.ts` or existing `genericWorkbookModel.ts` | Generic labels, collection item labels, create minimum messages, reference option shaping, mutation error parsing. | Network loading, component draft state, evidence preview/download DOM behavior. | Inputs: view contracts, rows, server errors. Outputs: labels, messages, option models, parsed errors. |
| Pure tests | `apps/web/src/*Model.test.ts`, `apps/web/src/*Rows.test.ts` | Branch coverage for extracted pure behavior. | Browser behavior or component wiring. | Inputs: fixtures from current tests. Outputs: behavior-preserving unit coverage. |

## Sprint 3 - Decompose `WorkbookShell` plus entity, generic, and assessment surfaces

Governing row: sprint 3 in the table above.

Execution order:

1. Characterize shell startup, surface switching, saved-view switching, query reset, and latest-query behavior.
2. Add or pin shell/surface tests before moving stateful logic.
3. Reuse sprint 2 pure helpers; do not duplicate them.
4. Extract hooks for shell state, saved views, query loading, reference options, and mutation flows.
5. Extract presentational controls after hook props stabilize.
6. Extract entity, generic, and assessment surface containers last.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Surface query hook | `apps/web/src/useWorkbookSurfaceQueries.ts` | Query state registry, latest-query guards, surface row loaders, load errors. | Tab markup, saved-view selector UI, mutation controls. | Inputs: incident key, active contract, query drafts, API helper. Outputs: rows, loading flags, errors, reload commands. |
| Saved-view actions hook | `apps/web/src/useSavedViewActions.ts` | Saved-view load, create, update, delete, active selection, home/default preference commands. | Client-local focus, scroll, selected row, popovers, or inspector state. | Inputs: incident key, active `view_schema_id`, query/layout state. Outputs: saved-view state and commands. |
| Startup selection hook | `apps/web/src/useWorkbookStartupSelection.ts` | URL parameter, workbook home/default surface selection, startup fallback. | Saved-view query contents or surface rendering. | Inputs: route params, preferences, registry. Outputs: initial and current surface selection commands. |
| Entity surface container | `apps/web/src/EntityWorkbookSurface.tsx` | Hosts/Identities surface composition over entity hooks and grid components. | Shell tabs, saved-view ownership, generic surface behavior. | Inputs: surface props, contracts, rows, commands. Outputs: entity grid and merge UI. |
| Generic surface container | `apps/web/src/GenericWorkbookSurface.tsx` | Generic surface composition over mutation/reference/evidence/party/task/decision controls. | Timeline behavior, shell startup, saved-view persistence rules. | Inputs: contract, rows, role, reference options, commands. Outputs: generic grid and action panels. |
| Assessment surface container | `apps/web/src/AssessmentWorkbookSurface.tsx` | Assessment create flow, support reference selector, grid composition. | Generic mutation branching or shell routing. | Inputs: assessment rows, support rows, role, submit command. Outputs: assessment grid and create panel. |
| Surface controls | `apps/web/src/GenericMutationPanel.tsx`, `apps/web/src/EvidenceAccessActions.tsx`, party/task/decision controls | Focused UI for create/edit, evidence access, party links, task lifecycle, and decision supersede. | Route construction unrelated to the control or broad surface state. | Inputs: view models and callbacks. Outputs: accessible controls and events. |

## Sprint 4 - Extract Timeline runtime hooks and mutation/live-update flows

Governing row: sprint 4 in the table above.

Execution order:

1. Characterize Timeline runtime flows: pending queue, committed versions, conflict keys, live update sequencing, history, mentions, evidence, paste, keyboard, focus, and presence.
2. Add or pin tests for each flow before extracting it.
3. Extract any remaining pure helpers first.
4. Extract hooks over stable pure models and explicit command outputs.
5. Do not extract presentational components until hook props are stable.
6. Keep Timeline surface-specific rendering in `TimelineWorkbook.tsx` until sprint 5.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Committed row runtime | `apps/web/src/useTimelineCommittedRows.ts` | Row-version high-water marks, stale refresh decisions, committed-row lookup. | Rendering, WebSocket connection setup, mutation payload construction. | Inputs: current rows and accepted mutations. Outputs: version refs, freshness decisions, lookup commands. |
| Mutation queue hook | `apps/web/src/useTimelineMutationQueue.ts` | Pending replay orchestration around `WorkbookPendingQueueModel`, save-state inputs, replay admission. | Queue pure model internals or grid markup. | Inputs: mutation intents and API callbacks. Outputs: queue state, save-state data, enqueue/drain commands. |
| Socket connection hook | `apps/web/src/useTimelineSocketConnection.ts` | WebSocket connect/reconnect, lifecycle reducer effects, sequence-gap handling, presence message intake. | Row rendering or conflict UI. | Inputs: incident key, session identity, lifecycle callbacks. Outputs: connection state and event dispatches. |
| Viewport continuity hook | `apps/web/src/useTimelineViewportContinuity.ts` | Focus anchors, selected row preservation, scroll restore, input focus keys. | Mutation semantics or saved-view state. | Inputs: grid refs, row ids, focus requests. Outputs: restore commands and anchor state. |
| History actions hook | `apps/web/src/useTimelineHistoryActions.ts` | Row history load, rollback preview, delete/restore/rollback commands. | Inspector markup. | Inputs: selected row, API callbacks, row idle wait command. Outputs: history state and action callbacks. |
| Mention actions hook | `apps/web/src/useTimelineMentionActions.ts` | Mention dismiss/reopen/resolve/reject/convert flows and entity refresh triggers. | Mention chip presentation. | Inputs: mention state, selected row, API callbacks. Outputs: mention command callbacks and reload requests. |
| Evidence attachment hook | `apps/web/src/useTimelineEvidenceAttachment.ts` | Evidence upload slot creation, object upload, evidence link to Timeline row. | Evidence panel rendering or generic evidence surface actions. | Inputs: file, selected row, API and upload helpers. Outputs: attach state and commands. |
| Paste and keyboard hooks | `apps/web/src/useTimelinePaste.ts`, `apps/web/src/useTimelineKeyboardNavigation.ts` | Timeline paste target validation, conflict registration, scalar paste fallback, navigation shortcuts. | Grid cell markup or route-level saved-view behavior. | Inputs: current focus, rows, contracts, queue commands. Outputs: paste and keyboard handlers. |

## Sprint 5 - Extract Timeline presentation seams and finalize shell cleanup

Governing row: sprint 5 in the table above.

Execution order:

1. Characterize current visual, focus, and accessibility behavior before moving render code.
2. Add or pin component/browser tests for controls whose markup or focus behavior changes.
3. Do not extract new pure utilities unless a render builder still owns behavior-free mapping.
4. Do not add new runtime hooks unless sprint 4 left an explicit gap.
5. Extract presentational components and render builders.
6. Keep surface-specific components thin and fed by stable hook/view-model props.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Timeline grid columns | `apps/web/src/TimelineGridColumns.tsx` or existing grid modules | Column definitions, cell render bindings, row action wiring over provided callbacks. | Fetching, queue draining, WebSocket handling, row history mutation. | Inputs: rows, contracts, view models, callbacks. Outputs: grid column and cell render props. |
| Presence markers | `apps/web/src/TimelinePresenceMarkers.tsx` | Row and cell presence marker presentation. | Presence transport or edit locking. | Inputs: presence view models. Outputs: visible and accessible presence indicators. |
| Timeline notices | `apps/web/src/TimelineWorkbookNotices.tsx` | Auto-resolution notices, pending replay notices, queue conflict notices. | Queue mutation logic or conflict resolution state machine. | Inputs: notice view models and callbacks. Outputs: dismiss/review UI events. |
| Inspector sections | `apps/web/src/TimelineInspectorSections.tsx` | Details, evidence, mentions, history, rollback, destructive action section composition. | Runtime fetch/mutation orchestration. | Inputs: inspector view models and callbacks. Outputs: accessible inspector panels. |
| Timeline assembler | `apps/web/src/TimelineWorkbook.tsx` | Final composition of hooks, models, grid surface, inspector, resolver, and status strip. | Pure helper definitions, broad nested callbacks, or presentation-only subtrees. | Inputs: props and hook outputs. Outputs: assembled Timeline workbook surface. |

# Commands

Run repository commands from the repo root through Make targets.

Planning and context:

```sh
make help
make task-guide ROLE=feature-dev PHASE=phase3
make task-guide ROLE=feature-dev PHASE=phase4
make task-guide ROLE=feature-dev PHASE=phase8
make task-guide ROLE=feature-dev PHASE=phase9
```

Fast authoring checks:

```sh
make frontend-typecheck
make frontend-unit
make frontend-import-boundary-check
make lint-biome
```

Targeted behavior slices:

```sh
make phase-slice PHASE=phase3
make phase-slice PHASE=phase4
make phase-slice PHASE=phase8
make phase-slice PHASE=phase9
```

Service-backed and browser checks when moved flows require them:

```sh
make service-backed-slice PHASE=phase4
make service-backed-slice PHASE=phase8
make service-backed-slice PHASE=phase9
make browser-e2e-webserver-backed
make browser-e2e-stateful
```

Accessibility, visual, and end-of-run checks when controls or layout change:

```sh
make browser-e2e-a11y-preflight
make browser-e2e-visual
make agent-finalize
make test-fast
make check
```

# Workbook-shell regression checklist

- [ ] Workbook open/default surface.
- [ ] Built-in/system view switching.
- [ ] Saved-view switching.
- [ ] Scalar edit/save-state.
- [ ] Timeline paste.
- [ ] Keyboard navigation.
- [ ] Selected row/focused cell preservation.
- [ ] Conflict display/resolution.
- [ ] Mention states.
- [ ] Evidence attach/preview/download states.
- [ ] Row history/rollback/destructive controls.
- [ ] Presence anchoring.
- [ ] Grouped result.
- [ ] Frozen column.
- [ ] Resize handle.
- [ ] Fill-down handle.
- [ ] Edit cell.
- [ ] Empty result.
- [ ] Save-state strip.

# Assumptions and defaults

- This tracker lives at the repository root as
  `APPS_WEB_SRC_REFACTOR_HANDOFF.md`.
- This tracker is implementation-support documentation only.
- Core 00 through Core 04 and `docs/domain.md` govern terminology and behavior.
- Core 05 applies only if timed, fixture-sensitive, benchmark, or
  publication-claim evidence is introduced; this tracker does not introduce
  that evidence.
- `docs/design.md` and `docs/guides/cartulary-ui-ux-design-guide.md` guide UI
  direction but are not Base Profile or extension-profile conformance evidence
  by themselves.
- Any required Core, contract, generated artifact, route, migration, or lockfile
  change is out of scope for this tracker and should trigger a separate plan.
- No behavior should be removed merely because Fallow flagged complexity.
  Removal requires evidence that the behavior has no continuing Core value and
  is not needed by current tests or implementation-conformance coverage.
