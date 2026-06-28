# Workbook Remediation Refactor Tracker

This tracker is the standalone planning and handoff artifact for workbook
remediation refactoring. It converts the historical remediation analysis from
`docs/handoffs/apps_web_src_workbook_refactor_tracker.md` section 16 into a
self-contained execution plan. The historical tracker is prior analysis only;
current repository state and owner documents govern this effort.

This slice is documentation and process remediation only. It does not authorize
runtime implementation, generated output edits, migrations, contracts, package
manifest changes, public routes, WebSocket behavior, storage behavior, browser
golden updates, or package export changes.

## 1. Authority and source map

### Authority order

| Source | Role for this effort |
| --- | --- |
| `docs/spec/00_document_set_status_and_precedence.md` | Current-profile document status, authority order, owner-section precedence, Base Profile boundary, Core 05 publication separation. |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Public route, storage, view-row, view-query, saved-view, WebSocket, workbook-surface, generated-contract, and modular-monolith owner rules. |
| `docs/spec/02_domain_model_schema_and_history.md` | Record model, source/projection separation, revision-history, recovery, and domain-state boundaries. |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Workbook shell, grid interaction, saved-view startup, pending queue, collaboration, inspector, conflict, history, restore, and workflow behavior. |
| `docs/spec/04_security_deployment_and_conformance.md` | Authorization, security, deployment, trust-boundary, and conformance verification owner rules. |
| `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` | Only claim-bearing timed, benchmark, fixture-sensitive, or publication evidence. It is not Base Profile implementation conformance. |
| `docs/domain.md` | Vocabulary and concept-boundary reference for workbook surfaces, view schemas, records, parties, artifacts, object blobs, saved views, system views, and entity mentions. |
| `docs/design.md` | Frontend design-direction constraints, token definitions, density, shell/grid/inspector layout, visual readiness, and accessibility presentation. It is not product-conformance evidence by itself. |
| `docs/testing-harness-nlspec.md` | Make-owned command invocation, target selection, scheduling, fixture lifecycle, retained artifact rules, generated-artifact gates, visual/a11y evidence class, and cleanup mechanics. |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Refactor workflow template and guardrails. Current code still overrides prior planning statements. |
| `docs/handoffs/apps_web_src_workbook_refactor_tracker.md` | Historical workbook analysis and completed slice notes. It is not proof of current repository state. |

If owner docs conflict, mark the affected slice `BLOCKED: owner contradiction`,
quote the conflicting owner sections, and stop before implementation. If live
repo state contradicts historical analysis, record the contradiction and adapt
the plan to live repo state.

### Current source inputs inspected

Live state was refreshed before creating this file:

| Input | Current observation |
| --- | --- |
| Branch and dirty tree | `git status --short --branch` returned `## main...origin/main`; no dirty files before this tracker was created. |
| Commit | `970c87dd29ee3481db55a911c8d84d7dd8651281`. |
| Target artifact | `docs/handoffs/workbook_remediation_refactor_tracker.md` was absent before this slice. |
| Workbook controller size | `apps/web/src/workbook/WorkbookShell.tsx` has 3072 lines after `I-01`; incident-identity and responsive-layout coordination moved behind app-owned model/hooks. |
| Timeline hot-path controller size | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` has 7094 lines after `I-04`; presence/session message construction, focus-anchor state synchronization, and pending replay HTTP dispatch moved behind behavior-owned seams. |
| Generated policy | `tools/generated_artifact_policy.json` was inspected. |
| Import boundaries | `tools/frontend_import_boundaries.json` was inspected. |
| Task surface | `tools/task_surface.generated.mk` was inspected for Make-owned targets. |
| Frontend phase maps | `tools/frontend_phase_maps/fe_p0_test_map.json` through `fe_p11_test_map.json` exist. |
| Frontend ledgers | `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md` through `fe_p11_coverage_ledger.md` exist. |
| E2E coverage inputs | `apps/web/e2e/*.spec.ts`, `*.test.ts`, helpers, visual snapshots, and a11y maps were inventoried by path. |
| Backend owners | `internal/modules/*/routes.go`, stores, contracts, queries, and migrations were scanned for workbook-relevant ownership. |

No owner contradiction was found during this planning slice.

### Generated roots and no-edit rules

Do not hand-edit these generated roots or generated files:

| Path | Owner input or generator class | Rule |
| --- | --- | --- |
| `internal/gen/**` | Go generated contracts and SQLC outputs | Refresh through Make-owned generation only. |
| `packages/protocol-ts/src/generated/**` | Generated TypeScript protocol contracts | Access through `@cartulary/protocol-ts`; do not import internals from app code. |
| `packages/ui-contracts/src/generated/**` | Generated design-token artifacts | Access through `@cartulary/ui-contracts`; do not import internals from app code. |
| `tools/task_surface.generated.mk` | Generated task-surface Make include | Update owner manifests/generators, then run Make-owned generator or drift target. |
| Generated phase ledgers and schedules | Phase manifests and frontend phase maps | Update owner inputs; do not hand-edit generated outputs. |
| `go.sum`, `pnpm-lock.yaml`, tool-managed installs | Dependency tooling | Do not hand-edit. |

Smallest safe response to stale generated artifacts: record the stale output,
name the owner input and Make target, run a drift target, and stop before
manual generated edits.

### Backend route and store owner map

| Behavior surface | Route owner | Store or logic owner | Notes |
| --- | --- | --- | --- |
| Workbook row query, generic create, patch, clipboard paste, bulk mutations, linked notes, conflict resolution | `internal/modules/workbook/routes.go` | `internal/modules/workbook/store.go`, `mutation_store.go`, `clipboard_paste_api.go`, `bulk_mutation_api.go`, plus `internal/platform/viewquery` | Owns public workbook row envelopes for generic workbook surfaces. |
| Timeline create, patch, clipboard paste substrate, review, supersede, time conversion profile | `internal/modules/workbook/routes.go` and `internal/modules/timeline/routes.go` | `internal/modules/timeline/store.go`, `clipboard_paste_store.go`, `state.go`, `auto_resolution.go` | Timeline remains the highest-risk hot path. |
| Collaboration WebSocket | `internal/modules/collaboration/routes.go` | `internal/modules/collaboration/*`, `internal/platform/ws/*` | Owns `GET /ws/v1/incidents/{incident_id}` lifecycle and session/presence stream behavior. |
| Saved views | `internal/modules/savedviews/routes.go` | `internal/modules/savedviews/store.go`, `scope.go`; query input `db/queries/savedviews_phase8.sql` | Owns incident saved-view list/create/update/delete behavior. |
| Evidence blobs, attach, preview, download handles | `internal/modules/evidence/routes.go` | `internal/modules/evidence/store.go`, `upload_token.go`, `blobref/*`; object-store platform boundary | Owns object blob lifecycle, attach-blob, preview/download handles, and handle redemption. |
| History, delete, restore, rollback | `internal/modules/revisions/routes.go` | `internal/modules/revisions/store.go`, `delete_restore_store.go`, `rollback_store.go` | Owns row history, restore, and rollback semantics. |
| View-schema discovery and inspector configuration payloads | `internal/modules/viewschemas/routes.go` | `internal/platform/viewschema/*`, `contracts/view-schemas/*` | Owns public view-schema discovery route shape. |
| Entity, identity, host, indicator create/patch, mentions, merge | `internal/modules/entities/routes.go` | `internal/modules/entities/*` | Related owner for generic surfaces and mention/merge workflows. |
| Assessments create | `internal/modules/assessments/routes.go` | `internal/modules/assessments/store.go` | Related owner for assessment workbook surface. |
| Incident startup and preferences | `internal/modules/incidents/routes.go`, `startup.go` | `internal/modules/incidents/store.go`, `db/queries/incidents_phase2.sql` | Owns incident directory, startup fallback, and workbook preference inputs. |

### Contract, migration, query, and generated-surface owners

| Surface | Authored owner inputs | Generated or derived surfaces |
| --- | --- | --- |
| OpenAPI | `contracts/openapi/cartulary.openapi.yaml` | Protocol facades and contract tests; update owner spec first for behavior changes. |
| WebSocket schema | `contracts/ws/index.schema.json` | Collaboration stream types and tests. |
| View schemas | `contracts/view-schemas/index.json`, `contracts/view-schemas/cartulary.view.*.json` | `packages/protocol-ts`, `packages/view-contracts`, `packages/ui-contracts` facade behavior. |
| Database migrations | `db/migrations/*.sql` | Database state and migration drift checks. |
| SQL queries | `db/queries/*.sql` | `internal/gen/sql/*.go` through SQLC generation. |
| Frontend package boundaries | `tools/frontend_import_boundaries.json` | `make frontend-import-boundary-check` enforcement. |
| Task surface and scheduler topology | `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, generated Make include | Make-owned target inventory and accounting. |

Behavior changes must be spec-first: owner spec, derived contract/migration or
generator input, drift/generation check, implementation, tests, handoff.

## 2. Current-state inventory

### Workbook frontend modules and controllers

| Area | Current files | State |
| --- | --- | --- |
| Shell coordinator | `apps/web/src/workbook/WorkbookShell.tsx` | Large shell coordinator still owns broad orchestration for surfaces, startup, saved views, queries, entity/assessment panels, incident controls, and some selectors. |
| Shell runtime seam | `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts` | Present; completed prior remediation remains in live repo. |
| Shell components | `components/ActiveSurfaceSavedViewSelector.tsx`, `SystemViewSwitcher.tsx`, `WorkbookGridControls.tsx`, `WorkbookSheetToolbar.tsx`, `WorkbookShellSlots.tsx`, `WorkbookStatusStrip.tsx`, `WorkbookSurfaceFrame.tsx` | App-owned presentation facades; preserve selectors, labels, layout, and package facades. |
| Generic surface facade | `components/GenericWorkbookSurface.tsx`, `GenericMutationControl.tsx` | Present; completed prior remediation remains in live repo. |
| Workbook models | `models/workbookQuery.ts`, `workbookSavedViews.ts`, `workbookSavedViewRuntime.ts`, `workbookStartup.ts`, `workbookSurfaceRegistry.ts`, `workbookContractRows.ts`, plus entity/assessment/generic models | App-owned state and behavior models; shared contract parsing stays behind `@cartulary/view-contracts` unless duplication or generated leakage proves otherwise. |

### Timeline hot-path modules

| Area | Current files | State |
| --- | --- | --- |
| Timeline controller | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Large hot-path controller still owns many render, mutation, collaboration, inspector, row, and continuity responsibilities. |
| Timeline grid components | `TimelineGridSurface.tsx`, `TimelineWorkbookGrid.tsx`, `TimelineCellEditors.tsx`, `TimelineRowActions.tsx` | Must keep grid vendor details behind `@cartulary/grid-adapter`. |
| Timeline inspector components | `TimelineWorkbookInspector.tsx`, `TimelineHistoryPanel.tsx`, `TimelineEvidencePanel.tsx`, `TimelineMentionsPanel.tsx`, `TimelineConflictResolver.tsx`, `TimelineWorkbookNotices.tsx`, `TimelinePresenceMarkers.tsx` | Preserve inspector default-closed behavior, focus, labels, and route semantics. |
| Timeline hooks | `useTimelinePendingSaves.ts`, `useTimelineInspectorSelection.ts`, `useTimelineHistoryState.ts`, `useTimelineLiveUpdates.ts`, `useTimelineGridInteractions.ts`, `useTimelineRows.ts`, `useTimelineCommittedRows.ts`, `useTimelineConflicts.ts`, `useTimelineEvidenceActions.ts`, `useTimelineMentions.ts`, `useTimelineWorkbookRuntime.ts` | Several completed seams remain present; live-update, grid-continuity, and mutation-submission seams remain active candidates. |
| Timeline models and services | `timelineConflictModel.ts`, `timelineRowsModel.ts`, `timelineViewportContinuityModel.ts`, `workbookTimelineModel.ts`, `workbookMentionChips.ts`, `workbookCollaborationMessages.ts`, `workbookSocketLifecycle.ts` | `workbookCollaborationMessages.ts` confirms phase-shaped workbook helper was renamed in production. |
| Timeline utilities | `utils/workbookPendingQueue.ts`, `workbookContinuity.ts`, `workbookGridFocus.tsx`, `workbookKeyboard.ts`, `workbookClipboard.ts`, `workbookPresence.ts`, `workbookValueFormat.ts`, `workbookStyles.ts` | Preserve current behavior; do not introduce direct vendor or generated-internal imports. |

### Shared packages and allowed facades

| Package | Current responsibility | Boundary rule |
| --- | --- | --- |
| `@cartulary/grid-adapter` | Direct `react-data-grid` integration, vendor CSS singleton, grid types, focus/selection primitives, test support. | Only `packages/grid-adapter` imports `react-data-grid`; app code must use the package facade. |
| `@cartulary/protocol-ts` | Protocol and generated contract facade. | Generated internals under `packages/protocol-ts/src/generated/**` are read-only and facade-only. |
| `@cartulary/ui-contracts` | Runtime-safe selectors, test-id builders, generated design-token facade. | Promote only shared runtime/test/browser selectors; private component-only IDs can remain local. |
| `@cartulary/view-contracts` | TypeScript adapters around generated view-schema contracts, row normalization, inspector configuration parsing. | Own generated-contract adaptation; do not move app workflow state into this package. |
| `@cartulary/test-utils` | Browser/helper choreography for tests. | Runtime app code must not import test helper surfaces. |

Current scan found direct `react-data-grid` text outside `packages/grid-adapter`
only in workbook tests asserting `gridAdapterVendor` is `"react-data-grid"`.

### Relevant tests and browser behavior families

| Behavior family | Unit or integration characterization | Browser/e2e families |
| --- | --- | --- |
| Shell/startup | `WorkbookShell.surfaces.test.tsx`, `workbookStartup.test.ts`, `workbookSavedViewRuntime.test.ts`, `workbookSurfaceRegistry.test.ts` | `frontend.phase4.public-route.spec.ts`, `phase8.workbook.spec.ts`, `browser-e2e-webserver-backed` when startup browser flow changes. |
| Grid/create | `WorkbookShell.phase3.grid.test.tsx`, `WorkbookShell.phase3.payload.test.tsx`, `workbookTimelineModel.test.ts`, `timelineViewportContinuityModel.test.ts` | `phase3.spec.ts`, `phase9.keyboard.spec.ts`, `phase9.sentinel.spec.ts`, stateful browser target when persistence, focus, paste, or scroll changes. |
| Generic/mentions | `genericWorkbookModel.test.ts`, `entityWorkbookModel.test.ts`, `WorkbookShell.surfaces.test.tsx`, `WorkbookShell.phase5.mentionChips.test.ts` | `phase4.workbook.generic.spec.ts`, `phase4.mentions.lifecycle.spec.ts`, `phase4.mentions.resolve.spec.ts`, webserver-backed when route flow changes. |
| Evidence | `evidenceLifecycleViewModel.test.ts`, `TimelineEvidencePanel.test.tsx`, `WorkbookShell.phase5.test.tsx` | `phase5.evidence.spec.ts`, `phase6.evidence-integration.spec.ts`, webserver-backed when evidence handle flow changes. |
| Collaboration/session | `workbookSocketLifecycle.test.ts`, `workbookPendingQueue.test.ts`, `WorkbookShell.phase6.test.tsx` | `phase6.collaboration.spec.ts`, `phase6.session-recovery.spec.ts`, stateful when socket/session or persistence interaction changes. |
| History/restore | `workbookInspectorModel.test.ts`, `WorkbookShell.phase7.test.tsx`, `WorkbookShell.phase9.inspector.test.tsx` | `phase7.history.spec.ts`, `phase10.restore.spec.ts`, stateful when route flow or continuity changes. |
| Saved views/query | `workbookQuery.test.ts`, `workbookSavedViews.test.ts`, `workbookSavedViewRuntime.test.ts`, `WorkbookShell.phase8.query.test.tsx` | `phase8.workbook.spec.ts`, `phase8.workbook.support.spec.ts`, webserver-backed when saved-view browser flow changes. |
| Inspector/keyboard | `workbookKeyboard.test.ts`, `workbookContinuity.test.ts`, `WorkbookShell.phase9.sentinel.test.tsx`, `WorkbookShell.phase9.inspector.test.tsx` | `phase9.keyboard.spec.ts`, `phase9.inspector-actions.spec.ts`, a11y targets when focus, keyboard, label, or accessible-name behavior changes. |
| Visual readiness | Component/model tests plus visual fixtures as support | `workbook.visual.spec.ts`, `browser-e2e-visual`; design/readiness evidence only unless Core 05 publication criteria are active. |
| Accessibility readiness | Keyboard/focus unit tests plus a11y rows as support | `workbook.a11y.spec.ts`, `workbook.a11y-preflight.spec.ts`; readiness evidence only unless Core 05 publication criteria are active. |

### Existing completed remediation still present

| Prior remediation | Live evidence |
| --- | --- |
| Shell runtime seam | `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts` exists and is imported by shell code. |
| Status strip split | `apps/web/src/workbook/components/WorkbookStatusStrip.tsx` exists and owns status strip composition. |
| Generic surface facade | `apps/web/src/workbook/components/GenericWorkbookSurface.tsx` exists. |
| Timeline pending-save seam | `apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts` exists. |
| Inspector/history state seams | `useTimelineInspectorSelection.ts` and `useTimelineHistoryState.ts` exist. |
| Shared selector helper | `timelineInspectorMessageTestId()` exists in `packages/ui-contracts/src/index.ts`. |
| Behavior-named collaboration helper | `apps/web/src/workbook/timeline/services/workbookCollaborationMessages.ts` exists; no workbook `workbookShellPhase4.ts` production helper was found. |

## 3. Remediation gap matrix

| Gap | Remediation | Area | Rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Backend route/store evidence | Require every slice to name route owner, store owner, auth role, storage/projection owner, and backend validation target before behavior changes. | Documentation, implementation, tests | Frontend refactors can accidentally encode stale route or store assumptions. | Keeps workbook UI aligned with backend owner truth as modules evolve. | Process-only unless the slice changes behavior; behavior changes may require backend tests or migrations. | Route, auth, or storage semantics drift silently. | Contract-impact checklist completed; affected backend/service-backed Make target named and run when required. |
| Generated protocol internals | Keep generated internals read-only and facade-only; app code uses `@cartulary/protocol-ts`, `@cartulary/ui-contracts`, and `@cartulary/view-contracts`. | Package boundary, documentation, tests | Generated outputs are unstable and downstream of owner inputs. | Regeneration and contract evolution remain safe. | Illegal imports migrate to facades if found. | Hand edits or direct imports create drift and brittle upgrades. | `make generated-artifact-policy-check`; `make frontend-import-boundary-check`; no generated-root diffs. |
| E2E coverage and behavior-target mapping | Select browser evidence by behavior family instead of filename or historical phase alone. | Harness, tests, documentation | Workbook browser coverage is distributed across phase, frontend, visual, a11y, and support specs. | Narrower and more reliable validation for each slice. | No product migration; future UI-impacting slices may run broader browser targets. | Browser-impacting changes may be under-validated. | Slice handoff names behavior family and mapped Make target; browser triggers are explicit. |
| Frontend phase maps and retained artifacts | Treat phase maps, ledgers, and retained artifacts as harness accounting. Runtime names stay behavior-based. Retained evidence must name exact current run roots. | Harness, documentation, tests | Phase identity is evidence metadata, not architecture. | Prevents stale-evidence claims and production phase coupling. | No runtime compatibility impact. | Runtime modules mirror phase history; handoffs overclaim stale artifacts. | Phase/harness changes run `make phase-map-check`, `make phase-ledger-drift`, or relevant harness target; handoff names exact roots. |
| Contracts/db/backend source evidence | Enforce spec-first behavior changes: owner spec, derived contract or migration input, drift/generation, implementation, validation, handoff. | Specification, contracts, implementation, tests | Workbook behavior is downstream of specs, contracts, migrations, and queries. | Public behavior, storage, and generated surfaces remain coherent. | Behavior changes may require migrations, contract updates, generated refresh, and backend/frontend updates. | Implementation-first changes diverge from public contracts or storage. | Owner spec diff precedes behavior implementation; drift/generation and affected tests pass. |
| `WorkbookShell.tsx` controller size/cohesion | Continue splitting only around durable behavior seams; shell should coordinate surfaces rather than own mutation/state workflows. | Implementation, tests | The shell still owns broad startup, surface, query, saved-view, entity, assessment, and incident-control responsibilities. | Smaller, reviewable modules and easier surface additions. | Internal TypeScript movement only unless owner specs authorize behavior change. | Shell remains fragile and hard to extend. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`; browser target only when route/startup/browser flow changes. |
| `TimelineWorkbook.tsx` hot-path controller size/cohesion | Split by high-value behavior seams: collaboration/live updates, grid continuity, mutation submission, and remaining inspector/panel routing. | Implementation, tests | Timeline is the highest-risk workbook path and still has a large controller. | Clearer ownership of pending, socket, grid, mutation, conflict, and inspector behavior. | Internal refactor unless selectors, routes, wire behavior, focus, or layout change. | Hot-path behavior remains hard to reason about and unsafe to extend. | Existing unit characterization plus mapped browser targets when persistence, route flow, focus/scroll, or browser behavior changes. |
| Phase-shaped runtime naming | Prohibit new phase-shaped production modules. Keep phase names only in tests, maps, ledgers, and harness accounting. | Implementation, documentation, tests | Runtime architecture should describe behavior, not historical implementation phases. | Prevents future phase coupling. | Import-only migration if phase-shaped runtime names reappear. | New production code may copy bad phase-shaped boundaries. | Source scan has no workbook phase-shaped production helper; future cleanup runs frontend gates and import-boundary check. |
| Literal production `data-testid` values | Classify selectors as shared contract selectors or private local IDs. Promote only cross-boundary shared selectors to `@cartulary/ui-contracts`; leave private IDs local. | Implementation, tests, package boundary | Centralizing every literal bloats the selector API; shared selectors need one stable owner. | Stable test/browser contracts with minimal public selector surface. | Selector strings remain stable unless an owner spec authorizes change. | Runtime/tests drift, or selector package becomes noisy and brittle. | Affected tests pass; package tests cover promoted selectors; import-boundary check passes. |
| View-contract access and package seam rules | Keep app workflow state in app models. Move only repeated generated-contract adaptation, parsing, or normalization into `@cartulary/view-contracts` when duplication or generated leakage is proven. | Implementation, package boundary, tests | Shared packages should own stable contract adaptation, not application workflows. | Cohesive package APIs and lower coupling to generated internals. | Future package API additions may require app import updates. | App duplicates parsing or shared package scope expands without discipline. | Duplication/leak trigger is recorded; package and app tests pass after any move. |
| Visual/a11y evidence overclaim risk | Keep visual and accessibility as design/readiness/support evidence unless Core 05 claim-publication criteria are explicitly active. | Specification, documentation, tests, harness | Readiness evidence and product conformance have different owners and failure semantics. | Honest evidence accounting and safer release claims. | No runtime migration. | Handoffs overclaim readiness rows as Base Profile proof. | Handoff labels evidence class; visual/a11y targets run only for layout, focus, keyboard, accessible-name, or readiness changes. |

## 4. Required contract-impact checklist

Every future workbook slice must copy and complete this checklist before
implementation. If all entries are `none`, the slice may remain frontend-only
and behavior-preserving.

| Field | Required value |
| --- | --- |
| Behavior family | One or more families from section 5. |
| Public route impact | Exact route, method, and owner module, or `none`. |
| WebSocket impact | Exact event family and owner, or `none`. |
| Storage/projection impact | Store, SQL query, migration, view-schema, or projection owner, or `none`. |
| Authorization/security impact | Core 04 owner section or backend auth primitive, or `none`. |
| Generated-surface impact | Owner input, generator, generated roots, and drift target, or `none`. |
| Backend validation | Make target required, or `not required: frontend-only behavior-preserving move`. |
| Browser validation | Make target from section 5, or `not required` with reason. |
| Visual/a11y classification | `not touched`, `readiness evidence`, or `Core 05 publication boundary active`. |
| Rollback scope | Smallest revertible file set and expected behavior after rollback. |

## 5. Behavior-to-validation evidence map

Use this map instead of choosing browser evidence by filename. When a slice
changes multiple families, run the union of required targets.

| Behavior family | Typical owned behavior | Default validation | Browser/readiness validation trigger |
| --- | --- | --- | --- |
| Shell/startup | Workbook route entry, incident identity, initial surface, saved-view startup, account density defaults. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`. | `make browser-e2e-webserver-backed` when public startup, route sequencing, or saved-view browser flow changes. |
| Grid/create | Timeline grid rows, draft row, create, paste, row identity, focus anchors, viewport continuity. | Default frontend gates. | `make browser-e2e-stateful` when persistence, focus/scroll, paste, or browser-only grid behavior changes. |
| Generic/mentions | Generic workbook surfaces, entity/indicator create, mention resolution, relationship chips, party links. | Default frontend gates. | `make browser-e2e-webserver-backed` when generic route flow or mention browser flow changes. |
| Evidence | Evidence attach, object blob upload, preview/download handle invocation, blocked access messages. | Default frontend gates. | `make browser-e2e-webserver-backed` when evidence handle or route flow changes. |
| Collaboration/session | Presence, WebSocket lifecycle, auth pause/recover, live row updates, resume/reset handling. | Default frontend gates. | `make browser-e2e-stateful` when socket/session or persistence interaction changes. |
| History/restore | Inspector history, row delete, restore, rollback, deleted-row subjects, rollback preview. | Default frontend gates. | `make browser-e2e-stateful` when row-history, restore, rollback, or browser continuity changes. |
| Saved views/query | Saved-view selection, query JSON, layout JSON, sort, filters, grouping, home/default pointers. | Default frontend gates. | `make browser-e2e-webserver-backed` when saved-view/query browser flow changes. |
| Inspector/keyboard | Default-closed inspector, row action routing, keyboard shortcuts, focus movement, overlay semantics. | Default frontend gates. | `make browser-e2e-a11y` or `make browser-e2e-a11y-preflight` when focus, keyboard, labels, or accessible names change. |
| Restore/native surfaces | Native surface read/write affordances, restore/import continuity, incident bundle or recovery entry into workbook. | Default frontend gates plus backend/service target if storage changes. | `make browser-e2e-stateful` when restore or native-surface browser flow changes. |
| Visual readiness | Layout, density, shell/grid/inspector visual state, visual fixture coverage. | Default frontend gates. | `make browser-e2e-visual`; do not refresh goldens without explicit authorization. |

Visual and accessibility targets remain design/readiness/support evidence unless
Core 05 claim-publication criteria are explicitly active.

## 6. Implementation slice plan

All slices must be independently revertible. Do not start dependent slices until
their prerequisites are validated or explicitly blocked.

| Slice | Status | Depends on | Remediation type | Likely files | Unchanged behavior | Characterization evidence | Validation targets | Rollback notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `R-00` standalone tracker | DONE | none | Documentation/process | `docs/handoffs/workbook_remediation_refactor_tracker.md` | Runtime, contracts, generated files, routes, WebSocket, storage, selectors, visual goldens, and package exports unchanged. | Existing source inspection and this tracker. | `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift`, `make frontend-import-boundary-check`, `make lint-markdown`, `make agent-finalize`. | Delete this tracker. |
| `D-01` authority and contract-source cleanup | DONE | `R-00` validated | Documentation/process | Owner docs only if a contradiction or missing current owner is proven; otherwise this tracker. | No runtime or generated behavior. | Owner-doc audit found no contradiction requiring owner-spec cleanup before runtime slices. | `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift`, `make lint-markdown`; no owner-specific drift target required because no owner docs changed. | Revert docs-only tracker updates. |
| `D-02` evidence and harness map cleanup | DONE | `R-00`; `D-01` if owner cleanup is needed | Documentation/harness process | `tools/frontend_phase_maps/*`, phase ledgers, task-surface owner manifests only if evidence-map drift is found. | Phase names remain harness-only; production names remain behavior-based. | Phase map and ledger evidence confirmed current; no retained roots claimed without exact run paths. | `make phase-map-check`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make generated-artifact-policy-check`, `make json-shape-check`. | Revert docs-only tracker updates; no generated output or harness owner input changed. |
| `I-01` `WorkbookShell.tsx` coordination continuation | DONE | `R-00`; any required `D-*` slice validated | Runtime implementation | `WorkbookShell.tsx`, `hooks/useWorkbookIncidentIdentity.ts`, `hooks/useWorkbookResponsiveLayout.ts`, `models/workbookIncidentIdentity.ts`. | Public routes, request bodies, saved-view IDs, query JSON, layout JSON, selector strings, density behavior, surface registry. | `WorkbookShell.surfaces.test.tsx`, `workbookStartup.test.ts`, `workbookSavedViewRuntime.test.ts`, `WorkbookShell.phase8.query.test.tsx`. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`; browser target not required because startup route sequencing, selectors, focus, and layout behavior did not change. | Revert new shell seam files and restore previous shell wiring. |
| `I-02` timeline collaboration/live-update seam | DONE | `I-01` validated or declared unrelated; contract checklist complete | Runtime implementation plus test-map owner input | `TimelineWorkbook.tsx`, `useTimelineLiveUpdates.ts`, `workbookCollaborationMessages.ts`, `workbookCollaborationMessages.test.ts`, `tools/frontend_phase_maps/fe_p7_test_map.json`, `tools/frontend_phase_registry.json`, generated FE-P7 ledger. | WebSocket URL, message families, resume/reset handling, self-origin filtering, conflict interaction, save-state behavior. | `workbookSocketLifecycle.test.ts`, `WorkbookShell.phase6.test.tsx`, `workbookPendingQueue.test.ts`, FE-P7 collaboration-message helper tests. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`, `make phase-ledgers`, `make phase-map-check`, `make phase-ledger-drift`; browser stateful not required because wire behavior and session browser flow did not change. | Restore local presence/session message construction in `TimelineWorkbook.tsx`; remove FE-P7 scenario-title additions, registry digest updates, generated ledger refresh, and new service test. |
| `I-03` timeline grid continuity seam | DONE | `I-02` validated or declared unrelated; contract checklist complete | Runtime implementation | `TimelineWorkbook.tsx`, `useTimelineGridInteractions.ts`. | Selector strings, row/field anchoring, scroll preservation, keyboard behavior, grid-adapter facade. | Existing continuity, keyboard, and workbook sentinel tests. | `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`; browser/a11y not required because focus semantics, keyboard handling, selectors, and layout did not change. | Revert focus-anchor command extraction in `useTimelineGridInteractions.ts` and restore local `TimelineWorkbook.tsx` focus-anchor state callbacks. |
| `I-04` timeline mutation submission seam | DONE | `I-02`; `I-03` if continuity is touched; contract checklist complete | Runtime implementation | `TimelineWorkbook.tsx`, `timelineMutationRequests.ts`. | HTTP methods, request envelopes, `client_txn_id`, `base_row_version`, conflict anchors, retry/halt behavior. | `workbookPendingQueue.test.ts`, `workbookTimelineModel.test.ts`, `WorkbookShell.phase3.payload.test.tsx`, `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase6.test.tsx`. | Default frontend gates; backend/service-backed target not required because route semantics did not change. | Revert `timelineMutationRequests.ts` and restore inline pending replay fetch/timing code in `TimelineWorkbook.tsx`. |
| `I-05` selector facade continuation | DONE | `R-00`; contract checklist complete | Runtime/package implementation | `packages/ui-contracts/src/index.ts`, affected workbook components/tests only. | Exact selector strings and browser helper behavior. | Affected runtime tests plus `packages/ui-contracts` tests. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`; browser/visual/a11y only if helper choreography, focus, layout, or labels change. | Remove added selector API and restore local literals. |
| `I-06` view-contract adapter cleanup | DEFERRED | Duplication or generated leak proven; contract checklist complete | Runtime/package implementation | `packages/view-contracts/src/index.ts`, app models consuming repeated generated-contract adaptation. | App workflow state remains in app; generated internals stay facade-only. | `packages/view-contracts` tests, app model tests using affected contracts. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`. | Revert package API addition and restore app-local adaptation. |
| `V-01` final validation and handoff | DONE | Last completed or blocked implementation slice | Validation/handoff | This tracker and affected slice handoff notes. | No unrecorded generated edits, no phase-shaped production dependency, no stale retained evidence claim. | Exact command roots and final `git status`. | `make agent-finalize`; broaden only if slice risk requires it. | Revert handoff-only updates if needed; implementation rollbacks come from each slice row. |

## 7. Workstreams and sequencing

| Workstream | Dependencies | Sequencing | Risks | Exit criteria |
| --- | --- | --- | --- | --- |
| Evidence and authority baseline | Current repository state and owner docs | Refresh branch, commit, dirty tree, owner docs, generated policy, import boundaries, backend owners, contracts, migrations, E2E map before each slice. | Relying on stale tracker claims or retained roots. | Every active gap has owner, evidence class, and validation target. |
| Spec/contract cleanup | Evidence baseline | Behavior changes start in owner specs, then contracts or migrations, then generation/drift, then implementation. | Implementing around missing or contradictory owner text. | Owner diff and drift/generation pass, or slice is `BLOCKED`. |
| Frontend structural remediation | Evidence baseline and contract checklist | One behavior seam per slice: shell coordination, collaboration, grid continuity, mutation submission, selectors, or view-contract adapter. | Large hot-path movement without characterization. | Slice is small, reviewable, independently revertible, and validated before the next slice. |
| Evidence/harness remediation | Evidence map and harness owner docs | Keep phase maps as accounting; add or adjust mappings only through owner inputs. | Overclaiming visual/a11y readiness or stale artifacts. | Handoff names exact run roots and skipped checks with reasons. |
| Validation/final handoff | Completed or blocked slice | Run `make agent-finalize`; record changed files, commands, roots, failures, skipped retained-run maintenance, generated status, rollback, and safe restart. | Leaving the next agent without restart context. | No generated-file policy violation, no out-of-scope diffs, and every slice is `DONE`, `TODO`, `DEFERRED`, or `BLOCKED` with reason. |

## 8. Validation plan

All repository commands must run from the repository root through public Make
targets. Direct `go`, `pnpm`, Vitest, Playwright, Biome, and raw script commands
are developer conveniences only unless a Make-owned wrapper invokes them.

| Target | When to run | Evidence class |
| --- | --- | --- |
| `make generated-artifact-policy-check` | Always for this tracker and any slice touching generated policy, package boundaries, generated roots, or generated-adjacent docs. | Harness/generated policy support. |
| `make json-shape-check` | Always for this tracker and any slice touching JSON manifests, contracts, maps, or schema-shaped inputs. | Harness/schema support. |
| `make generate-drift` | This tracker and any spec/contract/migration/query/generated-input slice; required before implementation when generated surfaces could drift. | Generated drift support. |
| `make frontend-unit` | Runtime TypeScript behavior or package implementation slices. Skipped for docs-only `R-00`. | Product or implementation support according to mapped rows. |
| `make frontend-typecheck` | Runtime TypeScript behavior or package implementation slices. Skipped for docs-only `R-00`. | Implementation correctness. |
| `make frontend-import-boundary-check` | Always for this tracker and all frontend/package slices. | Package-boundary support. |
| `make browser-e2e-webserver-backed` | Shell/startup, generic/mentions, evidence, or saved-view browser flow changes. | Browser product/support according to mapped rows. |
| `make browser-e2e-stateful` | Persistence, collaboration/session, grid/create, history/restore, restore/native surface, or full-state browser behavior changes. | Browser product/support according to mapped rows. |
| `make browser-e2e-measurement` | Only when timing/measurement evidence is intentionally in scope. | Measurement support; not default local proof. |
| `make browser-e2e-visual` | Visual/layout/density/readiness changes only. Do not update goldens without explicit authorization. | Design/readiness/support unless Core 05 publication applies. |
| `make browser-e2e-a11y` | Focus, keyboard, label, accessible-name, or accessibility readiness changes. | Accessibility readiness/support unless Core 05 publication applies. |
| `make browser-e2e-a11y-preflight` | Explicit preflight row or blocked future-row smoke when mapped. | Accessibility readiness/support. |
| `make lint-markdown` | Docs-only tracker updates and final handoff edits. | Documentation lint support. |
| `make agent-finalize` | End of every slice before final handoff. If using retained successful run evidence, pass `RESULTS_DIR`; otherwise record that retained-run maintenance was skipped because `RESULTS_DIR` was unset. | Harness/finalization support. |

For `R-00`, required targets are:

```bash
make generated-artifact-policy-check
make json-shape-check
make generate-drift
make frontend-import-boundary-check
make lint-markdown
make agent-finalize
```

Skip `make frontend-unit`, `make frontend-typecheck`, and browser targets for
`R-00` because this slice changes only this handoff artifact. If this slice
expands beyond docs/process, reclassify it and run the runtime targets before
handoff.

## 9. Execution ledger

| Slice | Result | Evidence | Validation |
| --- | --- | --- | --- |
| `D-01` authority and contract-source cleanup | DONE | Refreshed `main` at `520f24674e83343420f232ee5c6215d64e344bfe`; dirty tree only this tracker after status update. Audited Core 00 owner precedence, Core 01 public HTTP/WebSocket/view-schema owners, Core 03 workbook interaction/presence/pending/continuity owners, Core 04 authorization boundaries, `docs/design.md` visual/a11y evidence classification, and `docs/testing-harness-nlspec.md` retained-evidence posture. No owner contradiction or missing owner text blocked behavior-preserving runtime slices. | PASS: `make generated-artifact-policy-check` at `.cartulary/test-results/20260628T010719Z-p1556465`; `make json-shape-check` at `.cartulary/test-results/20260628T010723Z-p1556650`; `make generate-drift` at `.cartulary/test-results/20260628T010726Z-p1556985`; `make lint-markdown` at `.cartulary/test-results/20260628T010735Z-p1558008`. |
| `D-02` evidence and harness map cleanup | DONE | Inspected frontend phase maps, generated frontend phase coverage ledgers, task-surface and execution-topology manifests, and harness retained-artifact rules. No phase-map, ledger, schedule, or retained-evidence owner input needed remediation; phase identity remains harness accounting only. | PASS: `make phase-map-check` at `.cartulary/test-results/20260628T010819Z-p1558929`; `make phase-ledger-drift` at `.cartulary/test-results/20260628T010824Z-p1559204`; `make phase-schedule-drift` at `.cartulary/test-results/20260628T010829Z-p1559548`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260628T010839Z-p1559774`; `make json-shape-check` at `.cartulary/test-results/20260628T010839Z-p1559797`. |
| `I-01` `WorkbookShell.tsx` coordination continuation | DONE | Contract-impact checklist: behavior family `Shell/startup`; public route impact none beyond preserving existing `GET /api/v1/incidents/{incident_id}` call; WebSocket none; storage/projection none; authorization/security none; generated-surface none; backend validation not required because this is frontend-only behavior-preserving extraction; browser validation not required because startup route sequencing, selectors, focus, and layout behavior did not change; visual/a11y not touched; rollback is new hook/model files plus `WorkbookShell.tsx` import/wiring revert. Extracted incident identity normalization/loading into `models/workbookIncidentIdentity.ts` and `hooks/useWorkbookIncidentIdentity.ts`; extracted viewport responsive layout hook into `hooks/useWorkbookResponsiveLayout.ts`. | PASS: `make frontend-typecheck` at `.cartulary/test-results/20260628T011108Z-p1561123`; `make frontend-unit` at `.cartulary/test-results/20260628T011120Z-p1561527`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T011139Z-p1563229`. |
| `I-02` timeline collaboration/live-update seam | DONE | Contract-impact checklist: behavior family `Collaboration/session`; public route none; WebSocket event family preserved (`hello`, `resume`, `presence_update`, `hello_ack`, `resume_ack`, `presence_snapshot`, `presence_delta`, `record_changed`); storage/projection none; authorization/security unchanged; generated-surface impact limited to FE-P7 harness owner map/registry plus generated ledger refresh for new unit scenario titles; backend validation not required because this is frontend-only message-construction extraction; browser validation not required because socket URL, session ordering, resume/reset, self-origin filtering, conflict, and save-state behavior did not change; visual/a11y not touched. Moved presence input, presence_update, hello, and resume message construction into `workbookCollaborationMessages.ts`; shared `TimelinePresenceDraft` through the service; added FE-P7 service tests and refreshed frontend phase accounting. | PASS: `make frontend-typecheck` at `.cartulary/test-results/20260628T011610Z-p1565324`; initial `make frontend-unit` failed at `.cartulary/test-results/20260628T011623Z-p1565755` because four new passing tests were unmapped; after FE-P7 map/registry/ledger update, `make phase-ledgers` PASS at `.cartulary/test-results/20260628T011752Z-p1568113`, `make phase-ledger-drift` PASS at `.cartulary/test-results/20260628T011759Z-p1568548`, `make phase-map-check` PASS at `.cartulary/test-results/20260628T012001Z-p1570814`, `make frontend-unit` PASS at `.cartulary/test-results/20260628T012013Z-p1571159`, and `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T012034Z-p1572888`. |
| `I-03` timeline grid continuity seam | DONE | Contract-impact checklist: behavior family `Grid/create` and `Inspector/keyboard` only for focus-anchor state; public route none; WebSocket none; storage/projection none; authorization/security none; generated-surface none; backend validation not required because this is frontend-only behavior-preserving extraction; browser/a11y validation not required because selector strings, focus targets, keyboard handling, scroll preservation, and layout semantics did not change; visual/a11y not touched. Moved workbook focus-anchor ref/state synchronization and visible-column validation into `useTimelineGridInteractions.ts`, with a shell adapter preserving child component callback signatures. | FAIL then PASS: initial `make frontend-typecheck` failed at `.cartulary/test-results/20260628T012414Z-p1574700` because `TimelineScalarEditor` still expected a two-argument focus callback; after adding the surface-binding adapter, `make frontend-typecheck` PASS at `.cartulary/test-results/20260628T012546Z-p1575565`, `make frontend-unit` PASS at `.cartulary/test-results/20260628T012609Z-p1576027`, and `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260628T012632Z-p1577750`. |
| `I-04` timeline mutation submission seam | DONE | Contract-impact checklist: behavior family `Grid/create` and mutation submission; public route shape unchanged for existing timeline create/patch replay paths; WebSocket self-origin tracking unchanged; storage/projection none; authorization/security unchanged; generated-surface none; backend validation not required because no route/method/envelope semantics changed; browser validation not required because replay ordering, persistence, focus, and conflict browser behavior did not change; visual/a11y not touched. Extracted pending replay HTTP dispatch and timing callbacks into `timelineMutationRequests.ts`, preserving method, path, JSON body, `client_txn_id`, materialized payload, response parsing, retry, halt, and conflict settlement behavior. | PASS: `make frontend-typecheck` at `.cartulary/test-results/20260628T012921Z-p1579060`; `make frontend-unit` at `.cartulary/test-results/20260628T012943Z-p1579518`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T013003Z-p1581268`. |
| `I-05` selector facade continuation | DONE | Contract-impact checklist: behavior family `Inspector/keyboard` selector contract only; public route none; WebSocket none; storage/projection none; authorization/security none; generated-surface none because `packages/ui-contracts/src/index.ts` is authored package source, not a generated root; backend validation not required because this is frontend/package selector-facade cleanup; browser/visual/a11y validation not required because the exact `timeline-inspector` string, helper choreography, focus, layout, labels, and accessible names did not change. Promoted the shared inspector shell selector to `timelineInspectorTestId()` in `@cartulary/ui-contracts`, updated runtime, unit, e2e, visual, and a11y consumers to use the facade, and tightened selector-policy ownership so raw `timeline-inspector` literals are package-owned. | PASS: `make frontend-typecheck` at `.cartulary/test-results/20260628T013557Z-p1583894`; `make frontend-unit` at `.cartulary/test-results/20260628T013606Z-p1584303`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T013626Z-p1586023`. |
| `I-06` view-contract adapter cleanup | DEFERRED | Contract-impact checklist: behavior family `Shell/startup`, `Saved views/query`, and generated view-contract access only if duplication or generated leakage is proven; public route none; WebSocket none; storage/projection none; authorization/security none; generated-surface none; backend validation not required because no implementation trigger was found; browser validation not required because no workflow behavior changed; visual/a11y not touched. Scanned app/package sources for direct generated-internal imports and found none. Confirmed row normalization and contract lookup flow through `@cartulary/view-contracts`; remaining wrappers in `workbookContractRows.ts` and `workbookTimelineModel.ts` are app-specific API/timeline materialization, not duplicated generated-contract parsing. No package API move was warranted. | PASS: `make generated-artifact-policy-check` at `.cartulary/test-results/20260628T013759Z-p1587072`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T013800Z-p1587244`. |
| `V-01` final validation and handoff | DONE | Contract-impact checklist: final validation/handoff only; public route none; WebSocket none; storage/projection none; authorization/security none; generated-surface validation required because FE-P7 phase accounting and package selector facade changed; backend validation not required because no backend route, method, envelope, auth, storage, migration, query, or projection semantics changed; browser/visual/a11y validation skipped because final review found no behavior, focus, layout, label, accessible-name, or golden changes; rollback comes from the completed slice rows plus final tracker-only edits. Final sanity check reported no generated-file changes, no direct generated-internal app imports, and no whitespace errors from `git diff --check`. | PASS: `make generated-artifact-policy-check` at `.cartulary/test-results/20260628T013845Z-p1587762`; `make json-shape-check` at `.cartulary/test-results/20260628T013845Z-p1587767`; `make phase-map-check` at `.cartulary/test-results/20260628T013845Z-p1588427`; `make phase-ledger-drift` at `.cartulary/test-results/20260628T013845Z-p1587807`; `make phase-schedule-drift` at `.cartulary/test-results/20260628T013854Z-p1589084`; `make frontend-typecheck` at `.cartulary/test-results/20260628T013859Z-p1589264`; `make frontend-unit` at `.cartulary/test-results/20260628T013909Z-p1589678`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T013931Z-p1591454`; final `make lint-markdown` at `.cartulary/test-results/20260628T014133Z-p1594138`; `make agent-finalize` at `.cartulary/test-results/20260628T013944Z-p1592261`. |

## 10. Handoff

### Current handoff record

| Field | Value |
| --- | --- |
| Date/time | Final validation handoff updated 2026-06-27T21:39:44-04:00. |
| Branch/commit | `main` at `520f24674e83343420f232ee5c6215d64e344bfe`. |
| Pre-edit dirty tree | `R-00` began clean. Later slices intentionally dirtied workbook runtime, selector tests, FE-P7 harness accounting, and this tracker. |
| Post-edit dirty tree | Expected modified files under `apps/web/e2e`, `apps/web/src/testing`, `apps/web/src/workbook`, `docs/handoffs`, `docs/testing/frontend_phase_coverage_ledgers`, `packages/ui-contracts`, and `tools/frontend_phase_*`; expected untracked files are the new workbook identity/layout hooks, incident identity model, timeline mutation request service, and collaboration-message service test. |
| Target module or seam | Workbook remediation refactor completion: shell coordination, timeline collaboration/live-update, grid continuity, mutation submission, selector facade, view-contract trigger scan, validation/handoff. |
| Current slice | `V-01` final validation and handoff complete. |
| Completed slices | `R-00`, `D-01`, `D-02`, `I-01`, `I-02`, `I-03`, `I-04`, `I-05`, `V-01`; `I-06` is `DEFERRED` because no generated-contract duplication or leak trigger was proven. |
| Files changed | Runtime: `WorkbookShell.tsx`, timeline workbook hooks/services/components, new shell hooks/model, new timeline mutation/collaboration service files. Tests/e2e: affected workbook shell, timeline selector, visual/a11y, and FE-P7 collaboration tests. Package: `packages/ui-contracts/src/index.ts` and test. Harness/docs: FE-P7 phase map/registry/ledger and this tracker. |
| Decisions made | Preserve public routes, WebSocket event families, HTTP methods/envelopes, selector strings, focus/layout semantics, generated roots, and backend/storage behavior. Promote only the shared `timeline-inspector` selector; defer view-contract package changes because current app code already uses facades and the remaining wrappers are app-specific materialization. Keep visual/a11y as readiness/support evidence. |
| Commands run | Per-slice validation plus final `make generated-artifact-policy-check`, `make json-shape-check`, `make phase-map-check`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-markdown`, `make agent-finalize`, `git diff --check`, `git diff --stat`, `git diff --name-status`, and `git status --short --branch`. |
| Passing validation | Final PASS roots: generated policy `.cartulary/test-results/20260628T013845Z-p1587762`; JSON shape `.cartulary/test-results/20260628T013845Z-p1587767`; phase map `.cartulary/test-results/20260628T013845Z-p1588427`; phase ledger drift `.cartulary/test-results/20260628T013845Z-p1587807`; phase schedule drift `.cartulary/test-results/20260628T013854Z-p1589084`; typecheck `.cartulary/test-results/20260628T013859Z-p1589264`; unit `.cartulary/test-results/20260628T013909Z-p1589678`; import boundary `.cartulary/test-results/20260628T013931Z-p1591454`; final Markdown lint `.cartulary/test-results/20260628T014133Z-p1594138`; agent finalize `.cartulary/test-results/20260628T013944Z-p1592261`. |
| Failing validation | Historical in-slice failures only: I-02 first `make frontend-unit` failed for unmapped new FE-P7 scenario titles and passed after harness map/ledger updates; I-03 first `make frontend-typecheck` failed for a callback signature mismatch and passed after adding the surface-binding adapter. Final validation has no failures. |
| Skipped checks | Backend/service-backed targets skipped because no backend route, auth, storage, query, migration, or projection behavior changed. Browser, visual, and a11y targets skipped because behavior, helper choreography, focus, layout, labels, accessible names, and goldens did not change; visual/a11y remain readiness/support evidence. Retained-run maintenance was skipped by `make agent-finalize` because `RESULTS_DIR` was unset. |
| Blockers | None known. |
| Rollback notes | Use the slice rows above for scoped rollback: revert I-01 shell hook/model wiring, I-02 collaboration-message service/test and FE-P7 map/registry/ledger refresh, I-03 grid-interaction callback extraction, I-04 mutation-request service, I-05 selector facade/test/e2e updates, and this tracker handoff. No generated roots, migrations, lockfiles, package manifests, or backend code were hand-edited. |
| Next recommended slice | No remaining required workbook remediation slice. Future work should start with this tracker checklist, prove an owner/spec or duplication trigger, then choose behavior-family validation before implementation. |
| Safe restart command | See below. |

```bash
cd /home/jochi/code/cartulary && git status --short --branch && sed -n '1,260p' AGENTS.md && sed -n '1,260p' docs/handoffs/workbook_remediation_refactor_tracker.md
```
