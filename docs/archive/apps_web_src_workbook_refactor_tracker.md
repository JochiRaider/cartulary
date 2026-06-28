# apps/web/src/workbook Refactoring Tracker and Handoff Planner

## 1. Status and source limits

| Field | Value |
| --- | --- |
| Branch | `main` tracking `origin/main` |
| Commit | `6b36f1362917c069362a0dd7d3fc63c484a36eff` |
| Dirty tree at scan | Clean before this artifact was created |
| Scan timestamp | `2026-06-27T16:32:37-04:00` |
| Framework path | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Target directory | `apps/web/src/workbook` |
| Mode | Planning and handoff generation only |
| Artifact path | `docs/handoffs/apps_web_src_workbook_refactor_tracker.md` |
| Primary seam | `/apps/web` workbook shell, controllers, query, mutation, conflict, inspector, presence, and continuity state |
| Production files modified | No |

Inspected in-scope source files: all 108 files returned by `rg --files apps/web/src/workbook`.

Inspected adjacent files and owner inputs:

| Path | Reason inspected |
| --- | --- |
| `AGENTS.md` | Repository procedure, command surface, generated-file policy, verification handoff rules |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Required planning structure and workflow catalog |
| `docs/spec/00_document_set_status_and_precedence.md` | Authority, precedence, contract-owner matrix |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | HTTP, WebSocket, view rows, saved views, workbook startup, generated view-schema contracts |
| `docs/spec/02_domain_model_schema_and_history.md` | Source-state versus projection and record-model vocabulary by owner delegation |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Grid-first behavior, inspector, saved views, startup, collaboration, conflict, paste, continuity |
| `docs/spec/04_security_deployment_and_conformance.md` | Authorization and browser security owner boundary by reference |
| `docs/domain.md` | Domain vocabulary, view schema, saved view, system view, inspector, projection, artifact, party boundaries |
| `docs/design.md` | Design-direction constraints for workbook shell, density, grid, inspector, selectors, visual and accessibility posture |
| `docs/testing-harness-nlspec.md` | Make-owned target mechanics, frontend row accounting, generated-artifact rules, visual and a11y evidence class |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | Frontend package boundaries, evidence classes, row-accounting discipline |
| `docs/guides/cartulary_browser_design_readiness_workflow.md` | Design-readiness workflow and visual/a11y non-conformance boundary |
| `docs/guides/cartulary_visual_golden_maintenance.md` | Visual golden ownership, fixture IDs, update boundary |
| `docs/guides/cartulary-dev-guide.md` | Repo-local frontend package seams and generated artifact baseline |
| `tools/frontend_import_boundaries.json` | Current import-boundary guardrails |
| `tools/generated_artifact_policy.json` | Generated roots and hand-edit prohibition |
| `.markdownlint-cli2.jsonc` | Markdown lint scope and handoff-path source limit |
| `packages/grid-adapter/src/core.ts` | Public grid adapter types and vendor-semantic boundary |
| `packages/grid-adapter/src/index.tsx` | Runtime grid adapter facade and RDG stylesheet ownership |
| `packages/grid-adapter/src/test-support.tsx` | Test renderer boundary used by workbook tests |
| `packages/ui-contracts/src/index.ts` | Stable selector and test-id facade plus generated design-token facade |
| `packages/view-contracts/src/index.ts` | View schema, row, patch, and inspector contract adapter facade |
| `packages/test-utils/src/index.ts` | Browser helper choreography for grid, saved views, scroll, paste, anchors |
| `apps/web/src/testing/timelineWorkbookTestSupport.ts` | Workbook unit-test support, mocked WebSocket, selectors, row helpers |
| `apps/web/src/testing/fetchMockTestSupport.ts` | Fetch mock helpers used by workbook tests |
| `apps/web/e2e/workbook.visual.spec.ts` | Workbook visual readiness and current-profile visual scenarios by targeted scan |
| `apps/web/e2e/workbook.a11y.spec.ts` | Workbook accessibility readiness scenarios by targeted scan |
| `apps/web/e2e/workbook.a11y-preflight.spec.ts` | Accessibility preflight rows by targeted scan |
| `internal/modules/workbook/{routes.go,store.go,mutation_api.go,mutation_store.go,bulk_mutation_api.go,clipboard_paste_api.go,conflict_merge.go}` | Workbook-facing query, row create/patch, conflict, clipboard, bulk mutation, store, and backend tests by targeted scan |
| `internal/modules/savedviews/{routes.go,api.go,scope.go,store.go}` | Saved-view route, payload, scope, storage, and backend tests by targeted scan |
| `internal/modules/revisions/{routes.go,store.go,rollback_api.go,rollback_store.go,delete_restore_api.go,delete_restore_store.go}` | Row history, rollback, delete, restore route and storage owners by targeted scan |
| `internal/modules/evidence/{routes.go,api.go,store.go}` | Evidence blob, attach, preview-handle, download-handle, redeem route and storage owners by targeted scan |
| `internal/modules/timeline/{routes.go,api.go,store.go,state.go,hooks.go}` | Timeline review, test record-change socket, substrate, and legacy timeline contract owners by targeted scan |
| `internal/modules/viewschemas/routes.go`, `internal/platform/viewquery/query.go`, `internal/platform/viewschema/{registry.go,query.go,layout.go}`, `internal/platform/ws/ws.go` | View-schema discovery, query validation, registry, layout, and WebSocket platform owners by targeted scan |
| `packages/protocol-ts/src/generated/{index.ts,contracts.ts}` | Generated protocol bundle inspected read-only; generated by `tools/contractgen` |
| `packages/ui-contracts/src/generated/design-tokens.ts` | Generated design-token bundle inspected read-only; generated by `scripts/generate-design-tokens.mjs` |
| `contracts/view-schemas/*.json`, `contracts/ws/index.schema.json`, `contracts/openapi/cartulary.openapi.yaml` | Derived contract surfaces for view schemas, WebSocket payloads, and public OpenAPI routes by targeted scan |
| `db/migrations/00003_phase2_incidents.sql` through `db/migrations/00020_phase9_optional_artifact_surfaces.sql`, `db/queries/{incidents_phase2.sql,savedviews_phase8.sql,timeline_phase3.sql}` | Workbook-relevant storage and query inputs by targeted scan |
| `apps/web/e2e/*.spec.ts`, `apps/web/e2e/*.test.ts` | Complete e2e spec/test inventory counted as 41 files; workbook-relevant phase/spec families summarized in section 12 |
| `tools/frontend_phase_maps/fe_p*_test_map.json`, `tools/phase*_test_map.json` | Frontend phase maps and backend phase manifests inspected as harness accounting, not runtime structure |
| `.cartulary/test-results/20260627T214933Z-p1300061`, `.cartulary/test-results/20260627T214933Z-p1300076` | Current retained roots for artifact-level `generated-artifact-policy-check` and `json-shape-check` validation |

Adjacent evidence now inspected for remediation:

| Evidence surface | Inspection result | Scope decision |
| --- | --- | --- |
| Backend route handlers and stores | Workbook-facing route registrations and owner files are identified for query, saved views, mutation, history, rollback, evidence handles, view schemas, view query validation, and WebSocket collaboration. | Backend implementation remains out of scope for behavior-preserving frontend refactors; inspect deeper only when a slice changes public contract assumptions. |
| Generated protocol and UI internals | Generated files carry `Code generated ... DO NOT EDIT` headers; public facades remain `@cartulary/protocol-ts` and `@cartulary/ui-contracts`. | Generated files are read-only evidence. Any generated-surface change must start from owner inputs and drift checks, never hand edits. |
| Complete e2e coverage | All 41 `apps/web/e2e/*.spec.ts` and `*.test.ts` files were inventoried; workbook behavior is distributed across phase2 through phase10, frontend phase specs, visual/a11y specs, and support/helper tests. | Use Make-owned browser targets by behavior surface; do not infer product conformance from visual/a11y readiness rows. |
| Frontend phase maps and retained artifacts | FE-P0 through FE-P11 maps were inspected for row counts and target families; artifact validation retained roots are named. | Phase identity is harness accounting only and must not shape production runtime modules. |
| Contracts, migrations, and query inputs | 18 view-schema contracts, WebSocket/OpenAPI contracts, migrations `00003` through `00020`, and saved-view/timeline/incident query SQL were inventoried. | Behavior-affecting changes start in owner specs/contracts and may require generation or migration drift checks. |

Source limits:

- This tracker records source inspection and refactor planning only. It does not prove behavior preservation by itself.
- No validation command result is claimed unless listed in section 13 after it ran in this session.
- Visual and accessibility evidence are design-direction or implementation-support evidence unless a Core 05 publication boundary is separately active.
- Phase identity may remain in tests, phase maps, ledgers, and harness accounting, but must not become a production runtime dependency.

## 2. Scope and non-scope

| Category | Decision |
| --- | --- |
| Exact target | `apps/web/src/workbook` recursively |
| Allowed adjacent inspection | Imported package seams, app services, app testing helpers, workbook browser specs, owner docs, command manifests, generated-artifact policy |
| Primary owner | `/apps/web` workbook shell/controller/query/mutation/conflict/inspector/presence |
| Package seams in scope | `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/test-utils` when observed imports or tests require them |
| Generated roots | Not editable; owner inputs and drift checks only |
| Runtime phase identity | Forbidden as production dependency |
| Artifact-only allowed change | This Markdown handoff artifact |

Forbidden changes:

- Do not execute a refactor from this artifact alone.
- Do not rewrite components, hooks, tests, generated files, package manifests, lockfiles, configs, migrations, or contracts.
- Do not change public routes, HTTP or WebSocket wire behavior, generated surfaces, authorization behavior, workbook hot-path behavior, UI selectors, harness accounting, visual goldens, or accessibility artifacts.
- Do not hand-edit generated roots declared by `tools/generated_artifact_policy.json`.
- Do not make phase identity a runtime production dependency.

Stop conditions for future refactor sessions:

- Stop if a slice requires behavior changes not authorized by Core 00 through Core 04 or an adopted owner NLSpec.
- Stop if a slice requires generated artifacts but the owner input and generator path are not identified.
- Stop if a planned move would make `/apps/web` depend on `react-data-grid`, generated internals, package source internals, browser-test choreography, or phase identity.
- Stop if characterization for a high-risk movement is missing and no cheap pre-move test can be named.

## 3. Authority and contract map

| Source | Role | Sections or paths inspected | Governing behavior | Conflicts or TODOs |
| --- | --- | --- | --- | --- |
| `AGENTS.md` | Repo procedure | Entire file | Command use through Make, generated-file prohibition, verification reporting | No conflict found |
| Refactor framework | Planning structure | Full framework | Work tracker, workflow dependency map, frontend package seam doctrine | No conflict found |
| Core 00 | Top authority | Status, precedence, contract-owner matrix | Owner precedence, conformance, Core 01 and Core 03 owner families | No conflict found |
| Core 01 | Public contract owner | Architecture, routes, view row, query, saved view, startup, inspector config, evidence handle sections by targeted search; route owners under `internal/modules/{workbook,savedviews,revisions,evidence,timeline,viewschemas}` and platform owners under `internal/platform/{viewquery,viewschema,ws}` by targeted scan | `/api/v1/*`, `/ws/v1/*`, `view_row_v1`, `view_row_patch_v1`, saved views, workbook startup, view schemas, inspector metadata | No conflict found; backend behavior changes remain owner-doc-first work. |
| Core 02 | Domain model owner | Domain-facing references by targeted search; workbook-relevant backend modules, view-schema contracts, and migrations by targeted scan | Record, source-state, projection, evidence, entity, party, artifact, history boundaries | No conflict found; exact record-model edits remain out of this frontend refactor tracker. |
| Core 03 | Workbook interaction owner | Workbook surface, saved views, inspector, collaboration, conflict, paste, continuity sections | Grid-first workflow, inspector default closed, saved-view behavior, conflict resolver, focus and scroll continuity | No conflict found |
| Core 04 | Security owner | Referenced by owner maps; route-level authorization-adjacent workbook, saved-view, evidence-handle, rollback, and WebSocket owners by targeted scan | Authorization, session, egress and evidence access security | No conflict found; security behavior changes require owner-doc and backend validation. |
| `docs/domain.md` | Vocabulary owner | Status, thesis, distinctions, glossary, context map | Stable identifiers over visible labels, domain/source/projection distinctions | No conflict found |
| `docs/design.md` | Design-direction owner | Token, shell, grid, inspector, keyboard, visual and accessibility sections by targeted search | Dense workbook shell, default-closed inspector, density, visual state, selector discipline | Visual/a11y is not product conformance |
| Testing harness NLSpec | Harness owner | Status, public command surface, frontend row accounting, visual and a11y sections by targeted search | Make invocation, summaries, retained artifacts, frontend row accounting, generated-file policy | No conflict found |
| Frontend testing guide | Frontend evidence owner | Authority, package boundaries, test categories, evidence classes | Package seams, FE row accounting, support criteria | Historical guide status remains subordinate to live repo |
| Browser design workflow | Design-readiness support | Manual and Playwright workflow sections | Manual review and fixture classification only | Not conformance evidence |
| Visual golden guide | Visual maintenance support | Fixture map and update rules | `browser-e2e-visual`, fixture IDs, golden update discipline | Visual changes out of scope |
| Development guide | Implementation support | Frontend package boundary sections | Grid adapter, view contracts, ui contracts, test utils, apps web boundaries | No conflict found |
| Import boundaries manifest | Guardrail input | Full manifest | Direct RDG ban, generated-internal ban, workspace facade imports, runtime test-helper ban | No current production violation found in targeted scan |
| Generated artifact policy | Guardrail input | Full manifest | Generated roots and required markers | No generated workbook files found |

## 4. Top-level work tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target module and scope | scope | DONE | none | Codex | Sections 1 and 2 | One primary target seam and exclusions are explicit. |
| T-002 | Inspect current repo state | discovery | DONE | T-001 | Codex | Section 5 inventory | Relevant files, imports, tests, generated paths, and commands are listed. |
| T-003 | Map owner contracts | contracts | DONE | T-002 | Codex | Section 3 and 6 | Public behavior and owner docs are mapped. |
| T-004 | Freeze characterization evidence | tests | DONE | T-003 | Codex | Section 6 and 11 | Existing and missing characterization evidence are known. |
| T-005 | Plan boundary guardrails | architecture | DONE | T-003 | Codex | Sections 7, 8, 11, 12 | Import, generated, package, selector, and phase guardrails are defined. |
| T-006 | Plan behavior-preserving moves | implementation | DONE | T-004,T-005 | Codex | Section 10 | Small reviewable slice sequence is defined. |
| T-007 | Plan validation loop | validation | DONE | T-006 | Codex | Section 11 | Cheapest sufficient validation targets are named. |
| T-008 | Update docs/contracts if required | docs | DONE | T-003 | Codex | Sections 11 and 12 | Docs and generation are classified before codegen. |
| T-009 | Execute or hand off | handoff | DONE | T-006,T-007,T-008 | Codex | Section 13 | Next actor can continue without rediscovery. |

Status vocabulary for future updates: `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

### S-01 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `apps/web/src/workbook/WorkbookShell.tsx`; `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts`; this tracker |
| Decisions | Use the existing app-owned `useWorkbookShellRuntime` hook as the saved-view/query runtime seam; keep query DTOs, saved-view models, startup normalization, route calls, and selector strings unchanged; keep row loading, rendering, timeline/generic surface composition, and package facade imports in `WorkbookShell.tsx`; do not update the framework handoff because this tracker is the active S-01 handoff artifact |
| Commands and results | `make agent-finalize` PASS at `.cartulary/test-results/20260627T222432Z-p1334606`; `make frontend-unit` PASS at `.cartulary/test-results/20260627T222445Z-p1335834`; `make frontend-typecheck` FAIL at `.cartulary/test-results/20260627T222505Z-p1337765` for an unused `timelineContract` left in `WorkbookShell.tsx`; `make frontend-typecheck` PASS after cleanup at `.cartulary/test-results/20260627T222534Z-p1338736`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260627T222552Z-p1339498`; final `make frontend-unit` PASS at `.cartulary/test-results/20260627T222601Z-p1339981`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Browser E2E, visual, and a11y targets were not run because S-01 moved runtime ownership without intentional public route sequencing, layout, focus, or accessible-name changes; residual risk is limited to browser-only saved-view/startup sequencing beyond the existing unit characterization |
| Next workflow | Continue with S-02 shell slot/status facade after a fresh WF-00 state refresh, or run browser startup/saved-view evidence first if reviewers require browser-level confirmation |

### S-02 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `apps/web/src/workbook/WorkbookShell.tsx`; `apps/web/src/workbook/components/WorkbookStatusStrip.tsx`; this tracker |
| Decisions | Extracted the non-Timeline surface save-state/error/focus-anchor composition into `WorkbookSurfaceStatusStrip` inside the existing app-owned status-strip component module; kept `WorkbookSurfaceFrame`, `WorkbookShellSlotRegion`, shell-ready ID, slot IDs, status labels, queue/presence behavior, layout styles, selector strings, and route behavior unchanged |
| Commands and results | `make frontend-unit` PASS at `.cartulary/test-results/20260627T225528Z-p1371544`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260627T225528Z-p1371573`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Browser visual and a11y targets were not run because S-02 only moved status-strip composition and did not intentionally change layout, focus behavior, accessible names, or visual goldens |
| Rollback note | Revert the `WorkbookSurfaceStatusStrip` export/import and restore the three inline `statusStrip` fragments plus the local helper in `WorkbookShell.tsx` |
| Next workflow | Continue with S-03 generic surface controller split after a fresh WF-00 state refresh |

### S-03 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `apps/web/src/workbook/WorkbookShell.tsx`; `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`; this tracker |
| Decisions | Extracted the generic create/edit/evidence/party-link controller and rendering into the app-owned `GenericWorkbookSurface` component facade; `WorkbookShell.tsx` now only selects the active surface and passes rows, query state, refresh, saved-view selector, density, and incident/user inputs; kept generic create payloads, patch envelopes, evidence handle `{}` request bodies, attach-blob client transaction prefixes, party-link clear semantics, selector strings, UI copy, grid row IDs, status-strip wiring, route paths, generated artifacts, shared packages, and lockfiles unchanged |
| Commands and results | `make frontend-typecheck` FAIL at `.cartulary/test-results/20260627T230950Z-p1386418` for stale unused imports after the move; after import cleanup, `make frontend-typecheck` PASS at `.cartulary/test-results/20260627T231028Z-p1387110`; `make frontend-unit` PASS at `.cartulary/test-results/20260627T231051Z-p1387653`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260627T231113Z-p1389385`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Browser phase4 generic specs were not run because S-03 only moved app-owned component/controller code and did not intentionally change public route sequencing, persistence, browser navigation, evidence handle semantics, focus, layout, accessible names, or visual goldens; S-08 remains NA for this slice because existing unit characterization covered the moved behavior and no behavior gap was exposed |
| Rollback note | Revert `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`, remove its import/use from `WorkbookShell.tsx`, and restore the local `GenericWorkbookSurface` function and generic-only helper/style definitions inside `WorkbookShell.tsx` |
| Next workflow | Continue with S-04 timeline pending-save runtime seam after a fresh WF-00 state refresh |

### S-04 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts`; this tracker |
| Decisions | Expanded the app-owned `useTimelinePendingSaves` hook into the pending-save runtime seam: it now owns pending runtime ref construction, tab client-instance initialization, pending queue runtime factory, queue snapshot projection, refresh-block scopes, refresh-block predicates, and refresh-block begin/finish helpers; `TimelineWorkbook.tsx` still owns the replay algorithm, payload materialization, fetch calls, mutation settlement, socket effects, save-state labels, conflict resolver behavior, and row application; `WorkbookPendingQueueModel`, `pendingReplayCapacity`, queue coalescing, retry, halt, auth pause, overflow, save-state copy, conflict anchors, route paths, generated artifacts, shared packages, lockfiles, and selector strings were left unchanged |
| Commands and results | `make frontend-unit` PASS at `.cartulary/test-results/20260627T231926Z-p1394014`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260627T231948Z-p1395745`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260627T231948Z-p1395767`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Browser stateful/webserver-backed targets were not run because S-04 only moved app-owned pending runtime wiring and did not intentionally change public routes, persistence semantics, browser-flow sequencing, focus/scroll behavior, layout, accessible names, or visual goldens; S-08 remains NA for this slice because existing pending-queue/unit characterization covered the moved behavior and no behavior gap was exposed |
| Rollback note | Revert the new helper exports and internal ref construction in `useTimelinePendingSaves.ts`, restore the refs-input hook API, and restore the local pending runtime ref construction, refresh-block helpers, queue snapshot mapping, and client-instance fallback calls in `TimelineWorkbook.tsx` |
| Next workflow | Continue with S-05 inspector and history state seam after a fresh WF-00 state refresh |

### S-05 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`; `apps/web/src/workbook/timeline/hooks/useTimelineHistoryState.ts`; this tracker |
| Decisions | Moved inspector default-open state and selected-row workflow key into `useTimelineInspectorSelection`; moved row-history request sequencing, clear/cancel transitions, current history subject derivation, deleted-row subject handling, active live-history record selection, and current-subject matching into `useTimelineHistoryState`; kept `TimelineWorkbookInspector.tsx`, `TimelineHistoryPanel.tsx`, route paths, HTTP methods, request bodies, rollback target construction, server-advertised action visibility, no-row token copy, panel rendering, selector strings, generated artifacts, shared packages, and lockfiles unchanged |
| Commands and results | `make frontend-unit` PASS at `.cartulary/test-results/20260627T232721Z-p1400617`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260627T232743Z-p1402349`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260627T232743Z-p1402479`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Browser stateful/webserver-backed targets were not run because S-05 only moved app-owned state transitions and did not intentionally change public route sequencing, persistence, browser navigation, focus/scroll semantics, layout, accessible names, or visual goldens; S-08 remains NA for this slice because existing phase7/phase9/model characterization covered the moved behavior and no behavior gap was exposed |
| Rollback note | Restore local `isInspectorOpen`, selected-row workflow key, row-history request refs, current history subject derivation, and `clearRowHistory` logic in `TimelineWorkbook.tsx`; revert the added state/command exports in `useTimelineInspectorSelection.ts` and `useTimelineHistoryState.ts` |
| Next workflow | Continue with S-06 selector literal inventory and cleanup after a fresh WF-00 state refresh |

### S-06 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `packages/ui-contracts/src/index.ts`; `packages/ui-contracts/src/index.test.ts`; `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx`; `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`; this tracker |
| Decisions | Inventoried shared literal selectors and converted only the narrow `timeline-inspector-message` runtime/unit-test selector to `timelineInspectorMessageTestId()` in `@cartulary/ui-contracts`, preserving the exact string; left the broader `timeline-inspector` and conflict-resolver browser literals unchanged because converting them would touch many browser/a11y/visual/helper files and widen S-06 beyond a safe cleanup slice; did not touch generated design tokens, generated package internals, browser helper choreography, selector strings, or runtime behavior |
| Commands and results | `make frontend-unit` PASS at `.cartulary/test-results/20260627T233220Z-p1405529`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260627T233242Z-p1407260`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260627T233242Z-p1407322`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Remaining broad literal selectors such as `timeline-inspector` and conflict resolver IDs are intentionally left as inventoried future cleanup because they cross many browser specs and helpers; S-08 remains NA for this slice because the changed selector string is locked by ui-contracts/unit tests and no behavior gap was exposed |
| Rollback note | Remove `timelineInspectorMessageTestId()` and its package test assertion, restore `data-testid="timeline-inspector-message"` in `TimelineWorkbookInspector.tsx`, and restore the literal query in `WorkbookShell.surfaces.test.tsx` |
| Next workflow | Continue with S-07 phase-shaped production helper rename after a fresh WF-00 state refresh |

### S-07 implementation update, 2026-06-27

| Field | Value |
| --- | --- |
| Changed files | `apps/web/src/workbook/timeline/services/workbookCollaborationMessages.ts`; deleted `apps/web/src/workbook/timeline/services/workbookShellPhase4.ts`; updated imports in `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`, `TimelineMentionsPanel.tsx`, `TimelineWorkbookInspector.tsx`, `apps/web/src/workbook/timeline/services/workbookSocketLifecycle.ts`, `workbookSocketLifecycle.test.ts`, `apps/web/src/workbook/WorkbookShell.phase4.support.test.tsx`, `apps/web/src/testing/timelineWorkbookTestSupport.ts`; this tracker |
| Decisions | Renamed the phase-shaped production helper to `workbookCollaborationMessages.ts` because it owns `record_changed` collaboration message payload typing/guards and mention-resolution payload builders; did not rename exports, edit helper logic, alter WebSocket reducer behavior, alter mention payload shapes, change selector strings, touch generated artifacts, or update phase-shaped test names/maps because those remain harness evidence accounting |
| Commands and results | `git diff --no-index --exit-code <(git show HEAD:apps/web/src/workbook/timeline/services/workbookShellPhase4.ts) apps/web/src/workbook/timeline/services/workbookCollaborationMessages.ts` PASS; `make frontend-unit` PASS at `.cartulary/test-results/20260627T233800Z-p1410826`; `make frontend-typecheck` FAIL at `.cartulary/test-results/20260627T233819Z-p1412552` for one stale test-support import; after updating `apps/web/src/testing/timelineWorkbookTestSupport.ts`, `make frontend-typecheck` PASS at `.cartulary/test-results/20260627T233852Z-p1413318`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260627T233907Z-p1413785`; final `make frontend-unit` PASS at `.cartulary/test-results/20260627T233928Z-p1414349`; `git diff --check` PASS |
| Blockers | None |
| Unresolved risks | Browser stateful/webserver-backed, visual, and a11y targets were not run because S-07 was an import-only rename with byte-equivalent helper content and no public route, persistence, browser-flow, focus, layout, accessible-name, selector, or visual-golden change; S-08 is NA because no characterization gap was exposed |
| Rollback note | Rename `workbookCollaborationMessages.ts` back to `workbookShellPhase4.ts` and restore the seven import specifiers changed for this slice; no logic rollback is required because the helper content was verified byte-equivalent to the tracked old file |
| Next workflow | No implementation slice remains in S-02 through S-08; run end-of-session retained evidence maintenance and finalize the handoff |

## 5. Current-state inventory

| Path | Current responsibility | Exports | Imports | Target owner | Public/private/generated | External contracts touched | Risk | Current test evidence | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `apps/web/src/workbook/WorkbookShell.assessments.test.tsx` | Assessment workbook surface tests | none | Vitest, Testing Library, ui-contracts, fetch/test support, WorkbookShell | `/apps/web` tests | Test authored | Assessment query and create UI, selectors | Medium | Self test, phase4 map references | Covers assessment filters, support rows, record IDs. |
| `apps/web/src/workbook/WorkbookShell.phase3.autosave.test.tsx` | Timeline autosave behavior tests | none | Vitest, Testing Library, ui-contracts, test support, TimelineWorkbook | `/apps/web` tests | Test authored | Autosave, save labels, selectors | Medium | Self test, phase3 and carryover maps | Phase-shaped evidence only. |
| `apps/web/src/workbook/WorkbookShell.phase3.grid.test.tsx` | Timeline grid integration tests | none | grid-adapter, ui-contracts, view-contracts, test support, WorkbookShell | `/apps/web` tests | Test authored | Grid shell, RDG adapter facade, WebSocket mock, row cells | High | Self test, phase3 map | Asserts adapter vendor string in tests only. |
| `apps/web/src/workbook/WorkbookShell.phase3.payload.test.tsx` | Timeline payload tests | none | ui-contracts, Testing Library, Vitest, test support, WorkbookShell | `/apps/web` tests | Test authored | Create and mutation payloads, grid selectors | Medium | Self test, phase3 map | Uses grid-adapter test support mock. |
| `apps/web/src/workbook/WorkbookShell.phase4.actionSequencing.test.tsx` | Timeline action sequencing tests | none | ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | Same-record action sequencing, pending state | High | Self test | Characterizes ordering risk. |
| `apps/web/src/workbook/WorkbookShell.phase4.saveState.test.tsx` | Save-state status strip tests | none | ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | Status strip labels, pending queue | Medium | Self test, phase4 support | Good for pending queue movement. |
| `apps/web/src/workbook/WorkbookShell.phase4.support.test.tsx` | Support-only timeline helper and component tests | none | ui-contracts, Testing Library, Vitest, test support, mention/service helpers, TimelineWorkbook | `/apps/web` tests | Test authored support | Mention actions, focus continuity, mocked WebSocket, geometry | High | Self test, support-only phase maps | Must not become product conformance evidence. |
| `apps/web/src/workbook/WorkbookShell.phase4.timelineQuery.test.tsx` | Timeline query row identity integration | none | ui-contracts, view-contracts, Testing Library, Vitest, TimelineWorkbook | `/apps/web` tests | Test authored | `view_row_v1`, full cells, stable row identity | High | Self test | Strong characterization for query/normalization seam. |
| `apps/web/src/workbook/WorkbookShell.phase5.gridProvenance.test.tsx` | Hosts, identities, notes provenance grid tests | none | ui-contracts, view-contracts, Testing Library, Vitest, test support, WorkbookShell | `/apps/web` tests | Test authored | Contract-derived columns, provenance fields, selectors | High | Self test, phase5 map | Good guard for generic surface refactors. |
| `apps/web/src/workbook/WorkbookShell.phase5.mentionChips.test.ts` | Mention chip state model tests | none | Vitest, mention model | `/apps/web` tests | Test authored | Mention chip states | Low | Self test | Pure model characterization. |
| `apps/web/src/workbook/WorkbookShell.phase5.test.tsx` | Workbook evidence coverage tests | none | ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | Evidence counts and selectors | Medium | Self test, phase5 map | Evidence UI characterization. |
| `apps/web/src/workbook/WorkbookShell.phase6.test.tsx` | Collaboration, conflicts, replay tests | none | ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | WebSocket, presence, conflicts, pending replay | High | Self test, phase6 map | High-value before collaboration movement. |
| `apps/web/src/workbook/WorkbookShell.phase7.test.tsx` | History and rollback UI tests | none | ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | History route, rollback target, row version, selectors | High | Self test, phase7 map | Row-history characterization. |
| `apps/web/src/workbook/WorkbookShell.phase8.query.test.tsx` | Query controls tests | none | grid-adapter test support, ui-contracts, view-contracts, Testing Library, Vitest, WorkbookGridControls | `/apps/web` tests | Test authored | Sort, filters, grouping, saved-view query shape | Medium | Self test, phase8 map | Useful for query seam extraction. |
| `apps/web/src/workbook/WorkbookShell.phase9.inspector.test.tsx` | Inspector and row-local action tests | none | ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | Inspector panels, row-local actions, history | High | Self test | Strong guard for inspector refactors. |
| `apps/web/src/workbook/WorkbookShell.phase9.sentinel.test.tsx` | Keyboard, paste, anchor sentinel tests | none | grid-adapter, ui-contracts, Testing Library, Vitest, test support, TimelineWorkbook | `/apps/web` tests | Test authored | Clipboard paste, anchors, keyboard, grid semantics | High | Self test, phase9 map | Uses adapter types through package facade. |
| `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx` | Surface selection and generic mutation tests | none | ui-contracts, view-contracts, React, Vitest, fetch/test support, WorkbookShell | `/apps/web` tests | Test authored | Surface registry, saved views, generic mutations, private diagnostics | High | Self test, phase4 and later maps | Key characterization for `WorkbookShell.tsx`. |
| `apps/web/src/workbook/WorkbookShell.tsx` | Top-level workbook shell and generic/entity/assessment/evidence surfaces | `WorkbookShell`, types, re-exported timeline model and component types | grid-adapter, protocol-ts facade, ui-contracts, view-contracts, React, app services, workbook components, hooks, models, timeline components | `/apps/web` workbook shell/controller | Production authored public module | HTTP query/mutation, saved views, startup, selectors, density, evidence handles, inspector, generic surfaces | High | Many `WorkbookShell.*.test.tsx`, browser specs, visual/a11y specs | Largest shell file, broad responsibility, recommended split target. |
| `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx` | Saved-view selector and action menu UI | `ActiveSurfaceSavedViewSelector` | ui-contracts, lucide, React, saved view model, startup types | `/apps/web` workbook shell | Production authored private component | Saved-view selectors, home/default actions, query modified state | Medium | `WorkbookShell.surfaces.test.tsx`, browser phase8 | Candidate for saved-view seam characterization. |
| `apps/web/src/workbook/components/GenericMutationControl.tsx` | Generic create/edit input renderer | `GenericMutationControl` | view-contracts, React types, generic model, reference options | `/apps/web` mutation UI | Production authored private component | Field contracts, reference selectors, generic mutation payloads | Medium | `WorkbookShell.surfaces.test.tsx`, generic model tests | Keep contract-derived behavior stable. |
| `apps/web/src/workbook/components/SystemViewSwitcher.tsx` | System view menu UI | `SystemViewSwitcher` | ui-contracts, React, surface registry | `/apps/web` workbook shell | Production authored private component | Surface identity, system-view selectors | Medium | `WorkbookShell.surfaces.test.tsx`, phase2 browser | Avoid visible-label identity drift. |
| `apps/web/src/workbook/components/WorkbookGridControls.tsx` | Sort, filter, grouping controls | `WorkbookGridControls` | ui-contracts, view-contracts, lucide, React, workbookQuery | `/apps/web` query controller UI | Production authored private component | Query JSON, filter/sort/group selectors | Medium | `WorkbookShell.phase8.query.test.tsx`, browser phase8 | Good facade candidate after model lock. |
| `apps/web/src/workbook/components/WorkbookInspectorFeatureGroups.tsx` | Inspector panel feature group rendering | Types and `WorkbookInspectorPanelSection` | ui-contracts, view-contracts, React, inspector model | `/apps/web` inspector UI | Production authored private component | `inspector_config_v1`, feature-group selectors | Medium | `workbookInspectorModel.test.ts`, inspector tests | Preserve config-derived routing. |
| `apps/web/src/workbook/components/WorkbookSheetToolbar.tsx` | Shared sheet toolbar layout and styles | `WorkbookSheetToolbar`, toolbar styles, search label | ui-contracts, lucide, React | `/apps/web` workbook shell | Production authored private component | Toolbar selectors and add/filter/search affordances | Low | Surface and visual tests indirectly | Mostly presentational. |
| `apps/web/src/workbook/components/WorkbookShellSlots.tsx` | Shell slot region wrapper | `workbookShellId`, `WorkbookShellSlotRegion` | ui-contracts, React | `/apps/web` workbook shell | Production authored private component | Shell ready and slot test IDs | Low | Browser phase2, visual/a11y specs | Selector contract sensitive. |
| `apps/web/src/workbook/components/WorkbookStatusStrip.tsx` | Status strip, save state, presence summary | Type and `WorkbookStatusStrip` | ui-contracts, grid focus status, presence, styles | `/apps/web` presence/status | Production authored private component | Save-state selectors, presence display, focus anchor | Medium | save-state, presence, visual/a11y specs | Preserve role/status behavior. |
| `apps/web/src/workbook/components/WorkbookSurfaceFrame.tsx` | Surface layout frame and style constants | `WorkbookSurfaceFrame`, layout styles | React, workbook styles, shell slots | `/apps/web` workbook shell | Production authored private component | Work-area layout and shell slot selectors | Medium | browser layout, visual/a11y specs | Layout changes require visual/a11y evidence. |
| `apps/web/src/workbook/hooks/useAssessmentSupportRows.ts` | Fetch Timeline support rows for assessments | `useAssessmentSupportRows` | React, browserApi, workbookApi, surface registry, timeline model types | `/apps/web` query hook | Production authored private hook | View query route, timeline view schema | Medium | assessment tests | Duplicates query fetch pattern. |
| `apps/web/src/workbook/hooks/useEntityTimelinePreview.ts` | Fetch Timeline preview rows for entity inspector | `useEntityTimelinePreview` | React, browserApi, workbookApi, surface registry, timeline model | `/apps/web` inspector/query hook | Production authored private hook | View query route, entity preview filters | Medium | surface/entity tests | Query hook seam candidate. |
| `apps/web/src/workbook/hooks/useGenericReferenceOptions.ts` | Load reference rows for generic controls | `useGenericReferenceOptions` | view-contracts, React, browserApi, workbookApi, generic/query/reference/surface models, timeline model types | `/apps/web` query hook | Production authored private hook | View query route, reference options, view contracts | Medium | reference option and surface tests | Should stay app-owned, not package-owned. |
| `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts` | Active workbook identity and saved-view runtime state | `useWorkbookShellRuntime` | React, saved-view runtime model, saved view types, startup types | `/apps/web` shell controller | Production authored private hook | Surface selection, saved view identity | Medium | runtime model tests, surface tests | S-01 seam support. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.test.ts` | Assessment model tests | none | view-contracts, Vitest, timeline types, assessment model, surface registry | `/apps/web` tests | Test authored | Assessment fields and support labels | Low | Self test | Model characterization. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.ts` | Assessment draft and column helpers | Functions | view-contracts types, timeline types, generic enum helper | `/apps/web` assessment model | Production authored private model | Assessment view contract, enum values | Medium | assessment model and surface tests | Keep field keys contract-derived. |
| `apps/web/src/workbook/models/entityWorkbookModel.test.ts` | Entity model tests | none | view-contracts, Vitest, timeline types, entity model, surface registry | `/apps/web` tests | Test authored | Entity rows and merge plan | Low | Self test | Model characterization. |
| `apps/web/src/workbook/models/entityWorkbookModel.ts` | Host/identity row and merge plan model | Entity types and functions | view-contracts types, timeline row type, value formatter | `/apps/web` entity model | Production authored private model | Host/identity view cells, merge inputs | Medium | entity model and grid provenance tests | Avoid embedding dedupe semantics beyond UI plan. |
| `apps/web/src/workbook/models/evidenceLifecycleViewModel.test.ts` | Evidence lifecycle view-model tests | none | Vitest, evidence lifecycle model | `/apps/web` tests | Test authored | Evidence lifecycle display states | Medium | Self test, FE-P6 row | View model characterization. |
| `apps/web/src/workbook/models/evidenceLifecycleViewModel.ts` | Evidence state and count display view models | State constants, types, functions | none | `/apps/web` evidence UI model | Production authored private model | Evidence lifecycle and public error display | Medium | evidence model tests, evidence UI tests | Keep public error parsing stable. |
| `apps/web/src/workbook/models/genericWorkbookModel.test.ts` | Generic workbook model tests | none | ui-contracts, view-contracts, Vitest, generic model, surface registry | `/apps/web` tests | Test authored | Generic payloads, minimum create, labels | Medium | Self test | Good guard for generic surface split. |
| `apps/web/src/workbook/models/genericWorkbookModel.ts` | Generic payload, labels, references, create minimum model | Many generic model functions | ui-contracts type, view-contracts, timeline row type, value formatter, surface registry | `/apps/web` generic surface model | Production authored private model | View fields, create and patch payloads, party refs, public mutation errors | High | generic model and surface tests | Holds app-specific contract adaptation. |
| `apps/web/src/workbook/models/workbookContractRows.ts` | Generic view-row normalization and grid column builder | Row types and functions | grid-adapter types, ui-contracts selectors, view-contracts | `/apps/web` query/grid adapter boundary | Production authored private model | `view_row_v1`, grid columns, row/cell selectors | High | grid provenance and surface tests | Coordinate with view-contracts, avoid vendor leakage. |
| `apps/web/src/workbook/models/workbookDensity.test.ts` | Density resolution tests | none | Vitest, density model, surface registry | `/apps/web` tests | Test authored | Account density defaults | Low | Self test | Core 01 and Core 03 density guard. |
| `apps/web/src/workbook/models/workbookDensity.ts` | Effective workbook density resolver | `AccountDensityMode`, `resolveEffectiveWorkbookDensity` | grid-adapter type, protocol-ts facade type, surface registry | `/apps/web` shell model | Production authored private model | Density mode tokens | Low | density model tests, phase2 browser | Imports protocol facade only. |
| `apps/web/src/workbook/models/workbookInspectorModel.test.ts` | Inspector reducer tests | none | view-contracts, Vitest, inspector model | `/apps/web` tests | Test authored | Inspector default and retargeting | Medium | Self test | Important before inspector seam changes. |
| `apps/web/src/workbook/models/workbookInspectorModel.ts` | Inspector config selection and state reducer | Inspector types and functions | view-contracts types | `/apps/web` inspector model | Production authored private model | `inspector_config_v1`, panel IDs, no-row state | High | inspector model and inspector tests | Behavior freeze surface. |
| `apps/web/src/workbook/models/workbookQuery.test.ts` | Query model tests | none | view-contracts, Vitest, query model | `/apps/web` tests | Test authored | Query JSON, layout JSON, filter labels | Medium | Self test, phase8 tests | Strong for first slice. |
| `apps/web/src/workbook/models/workbookQuery.ts` | Query, saved-view query, layout serialization model | Query and layout types and functions | view-contracts | `/apps/web` query model | Production authored private model | View query request, saved-view `query_json`, `layout_json` | High | query model, phase8 unit and browser tests | S-01 seam dependency. |
| `apps/web/src/workbook/models/workbookReferenceOptions.test.ts` | Reference options tests | none | view-contracts, Vitest, reference model, surface registry | `/apps/web` tests | Test authored | Reference option field routing | Low | Self test | Generic surface support. |
| `apps/web/src/workbook/models/workbookReferenceOptions.ts` | Field-to-reference-option routing | Reference types and functions | view-contracts types, generic helper | `/apps/web` generic model | Production authored private model | Direct reference fields, party refs | Medium | reference option tests | Preserve field-key routing. |
| `apps/web/src/workbook/models/workbookResponsiveLayout.test.ts` | Responsive layout model tests | none | Vitest, responsive model | `/apps/web` tests | Test authored | Responsive chrome/block mode | Low | Self test, visual/a11y indirect | Design-direction support. |
| `apps/web/src/workbook/models/workbookResponsiveLayout.ts` | Chrome and block-size responsive model | Types and selectors | none | `/apps/web` shell model | Production authored private model | Responsive design direction | Medium | responsive model, visual/a11y specs | Layout-sensitive. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.test.ts` | Saved-view runtime tests | none | view-contracts, Vitest, runtime model, saved-view/startup types | `/apps/web` tests | Test authored | Selection identity, modified state | Medium | Self test | S-01 guard. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.ts` | Saved-view selection, deletion, modified-state helpers | Types and functions | view-contracts types, query model, saved-view/startup types, surface registry | `/apps/web` shell runtime model | Production authored private model | Saved-view identity, base-surface fallback, query/layout comparison | High | runtime and surface tests | S-01 core. |
| `apps/web/src/workbook/models/workbookSavedViews.test.ts` | Saved-view resource tests | none | view-contracts, Vitest, saved-view model | `/apps/web` tests | Test authored | Saved-view normalization and mutability | Medium | Self test | Contract freeze for saved views. |
| `apps/web/src/workbook/models/workbookSavedViews.ts` | Saved-view resource normalization and persistence helpers | Types and functions | view-contracts, query model, surface registry | `/apps/web` saved-view model | Production authored private model | Saved-view list, create/update payload normalization | High | saved-view tests and browser phase8 | First slice dependency. |
| `apps/web/src/workbook/models/workbookStartup.test.ts` | Startup model tests | none | Vitest, startup and surface registry | `/apps/web` tests | Test authored | Explicit launch, fallback, cleared pointers | Medium | Self test, browser phase8 | Startup contract freeze. |
| `apps/web/src/workbook/models/workbookStartup.ts` | Workbook startup selection normalization | Startup types and functions | surface registry | `/apps/web` shell startup model | Production authored private model | `sheet_ref`, startup source, URL query selectors | High | startup tests, browser phase8 | Keep URL and startup semantics unchanged. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.test.ts` | Surface registry tests | none | view-contracts, Vitest, surface registry | `/apps/web` tests | Test authored | Required and optional surface registry | Medium | Self test, browser phase2/9 | Public surface identity guard. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.ts` | Workbook surface constants and registry | Surface IDs, types, registry functions | ui-contracts types, view-contracts | `/apps/web` surface registry | Production authored private model | View schema IDs, system view grouping, optional surfaces | High | registry, surfaces, browser specs | Stable identities must not drift. |
| `apps/web/src/workbook/timeline/components/TimelineCellEditors.tsx` | Timeline scalar editors, relationship chips, draft create controls | Editor and chip components, styles | ui-contracts, lucide, React, styles, mention/timeline models | `/apps/web` timeline UI | Production authored private component | Cell editor selectors, relationship chips, draft creation | High | phase3, phase5, phase9 tests, visual/a11y | Hot-path UI. |
| `apps/web/src/workbook/timeline/components/TimelineConflictResolver.tsx` | Same-field conflict resolver UI | `TimelineConflictResolver` | ui-contracts, React, timeline model types | `/apps/web` conflict UI | Production authored private component | Conflict payload and resolution actions | High | phase6 tests, visual/a11y | Behavior freeze surface. |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.test.tsx` | Timeline evidence panel tests | none | ui-contracts, Testing Library, Vitest, timeline model types, panel | `/apps/web` tests | Test authored | Evidence panel selectors | Low | Self test | Component guard. |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.tsx` | Timeline inspector evidence count and upload panel | Types and `TimelineEvidencePanel` | ui-contracts, timeline row type | `/apps/web` timeline evidence UI | Production authored private component | Evidence attach selectors and row state | Medium | panel test, phase5/6 tests | Evidence hot path. |
| `apps/web/src/workbook/timeline/components/TimelineGridSurface.tsx` | Timeline grid viewport wrapper | Props and `TimelineGridSurface` | grid-adapter, React, query type, row type, TimelineWorkbookGrid | `/apps/web` timeline grid shell | Production authored private component | Grid viewport facade | Medium | grid tests, visual/a11y | Thin wrapper over adapter. |
| `apps/web/src/workbook/timeline/components/TimelineHistoryPanel.tsx` | Row history, rollback, delete/restore panel | History types and `TimelineHistoryPanel` | ui-contracts, React | `/apps/web` history UI | Production authored private component | History route data, rollback actions, destructive selectors | High | phase7/9 tests, visual/a11y | Preserve server-advertised action behavior. |
| `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx` | Mention resolution inspector UI | Types and `TimelineMentionsPanel` | ui-contracts, React, mention model, mention service types, cell editors | `/apps/web` mention UI | Production authored private component | Entity mention actions and selectors | High | phase4/5 tests, visual/a11y | Preserve raw mention semantics. |
| `apps/web/src/workbook/timeline/components/TimelinePresenceMarkers.tsx` | Presence cell and row-gutter markers | Presence marker components | ui-contracts, React, presence model | `/apps/web` presence UI | Production authored private component | Presence selectors and display initials | Medium | phase6 tests, visual/a11y | Presence display only. |
| `apps/web/src/workbook/timeline/components/TimelineRowActions.tsx` | Row action context menu | Types and `TimelineRowContextMenu` | ui-contracts, React, timeline surface ID, row type | `/apps/web` row-action UI | Production authored private component | Row action selectors, mark reviewed, history, supersede | High | phase7/9 tests | Keep actions row-bound. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Timeline workbook controller and UI composition | `TimelineWorkbook`, timeline types and helpers | grid-adapter, ui-contracts, view-contracts, React, ReactDOM, app services, shared components, hooks, models, utilities, timeline components/services | `/apps/web` timeline controller | Production authored public via WorkbookShell re-export | HTTP query/mutation, WebSocket, pending replay, conflicts, inspector, evidence, presence, focus/scroll continuity | High | Broad `WorkbookShell.phase*` tests, browser, visual, a11y | Largest hot-path file, high-risk split target after safer seams. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookGrid.tsx` | Timeline grid table rendering | `TimelineWorkbookGrid` | grid-adapter, ui-contracts, React, query type, surface ID, styles, row type | `/apps/web` timeline grid UI | Production authored private component | Grid columns, row/cell selectors, grouping | High | grid tests, visual/a11y | Adapter boundary sensitive. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx` | Timeline inspector shell | `TimelineWorkbookInspector` | ui-contracts, view-contracts, lucide, React, inspector feature groups, inspector model, surface ID, mention/timeline types, mentions panel | `/apps/web` timeline inspector UI | Production authored private component | Inspector config, no-row state, panel selectors | High | inspector tests, visual/a11y | Default-closed is controlled by caller state. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookNotices.tsx` | Pending queue and auto-resolution notices | `timelinePendingQueueMessage`, `TimelineWorkbookNotices` | ui-contracts, React, pending saves type, mention model | `/apps/web` notices/status | Production authored private component | Pending replay notice selectors, auto-resolution notices | Medium | pending queue tests, phase4/6 tests | Preserve save-state copy. |
| `apps/web/src/workbook/timeline/hooks/useTimelineCommittedRows.ts` | Committed row high-water mark and stale row guard | `useTimelineCommittedRows` | React, timeline model | `/apps/web` timeline query state | Production authored private hook | `record_id`, `row_version`, committed-version source | High | phase4/6 tests, timeline rows model tests | Critical for optimistic write safety. |
| `apps/web/src/workbook/timeline/hooks/useTimelineConflicts.ts` | Local same-field and paste conflict state | `useTimelineConflicts` | React, timeline model types | `/apps/web` conflict state | Production authored private hook | Conflict queue and paste conflict groups | High | phase6/9 tests, pending queue tests | Behavior-sensitive. |
| `apps/web/src/workbook/timeline/hooks/useTimelineEvidenceActions.ts` | Evidence action local state | `useTimelineEvidenceActions` | React | `/apps/web` evidence state | Production authored private hook | Evidence message and preview state | Medium | phase5/6 tests | Small hook. |
| `apps/web/src/workbook/timeline/hooks/useTimelineGridInteractions.ts` | Timeline selection, context menu, columns, focus state | Types and `useTimelineGridInteractions` | grid-adapter types, React, focus type, row type | `/apps/web` timeline grid state | Production authored private hook | Grid row/column state, focus anchor | High | phase9 tests, visual/a11y | Uses adapter public types. |
| `apps/web/src/workbook/timeline/hooks/useTimelineHistoryState.ts` | History panel state | `useTimelineHistoryState` | React, history panel types | `/apps/web` history state | Production authored private hook | History state and pending actions | Medium | phase7/9 tests | Candidate for history seam later. |
| `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts` | Inspector selected row and mention selection state | `useTimelineInspectorSelection` | React, mention model, row type | `/apps/web` inspector state | Production authored private hook | Selected row, mention, no-row behavior | High | inspector/history tests | Preserve retargeting. |
| `apps/web/src/workbook/timeline/hooks/useTimelineLiveUpdates.ts` | Live presence and socket lifecycle reducer bridge | Types and `useTimelineLiveUpdates` | React, presence model, socket lifecycle service | `/apps/web` collaboration state | Production authored private hook | WebSocket lifecycle, presence draft | High | phase6 tests, socket lifecycle tests | Keep socket effects in app seam. |
| `apps/web/src/workbook/timeline/hooks/useTimelineMentions.ts` | Mention and auto-resolution notice state | `useTimelineMentions` | React, mention model types | `/apps/web` mention state | Production authored private hook | Mention chips and notices | Medium | phase4/5 tests | Small state holder. |
| `apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts` | Pending replay runtime wrapper | Types and `useTimelinePendingSaves`, client instance ID | React, pending queue model | `/apps/web` pending replay state | Production authored private hook | Pending queue capacity, replay admission, save state | High | pending queue tests, phase4/6 tests | Good candidate after saved-view seam. |
| `apps/web/src/workbook/timeline/hooks/useTimelineRows.ts` | Row state and ref synchronization | `useTimelineRows` | React, row type | `/apps/web` timeline query state | Production authored private hook | Rendered rows and latest rows ref | Medium | broad timeline tests | Small hook. |
| `apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts` | Timeline query and save-state runtime helpers | Types and functions/hooks | view-contracts, React, query model, surface ID | `/apps/web` timeline controller | Production authored private hook/model | Timeline query state, filter draft, save state | High | phase8 query tests, timeline tests | Query seam overlap. |
| `apps/web/src/workbook/timeline/models/timelineConflictModel.test.ts` | Conflict parse tests | none | Vitest, conflict model | `/apps/web` tests | Test authored | Same-field conflict payload | Low | Self test | Guard conflict parsing. |
| `apps/web/src/workbook/timeline/models/timelineConflictModel.ts` | Same-field conflict parser | Types and parser functions | timeline model type | `/apps/web` conflict model | Production authored private model | Public same-field conflict payload | High | conflict tests, pending queue tests | Public error shape sensitive. |
| `apps/web/src/workbook/timeline/models/timelineRowsModel.test.ts` | Row freshness tests | none | Vitest, rows model | `/apps/web` tests | Test authored | High-water mark decisions | Low | Self test | Guard committed-row logic. |
| `apps/web/src/workbook/timeline/models/timelineRowsModel.ts` | Row freshness decision model | Types and function | none | `/apps/web` timeline query model | Production authored private model | `record_id`, `row_version` freshness | High | rows model and phase tests | Must not regress stale-row drop behavior. |
| `apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.test.ts` | Viewport continuity tests | none | Vitest, continuity model | `/apps/web` tests | Test authored | Entity refresh continuity barrier | Medium | Self test | Supports focus/scroll continuity. |
| `apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.ts` | Entity refresh continuity barrier model | Types and functions | none | `/apps/web` timeline continuity model | Production authored private model | Entity catalog, refresh expectations | Medium | continuity tests, phase4 support | Keep deterministic barrier. |
| `apps/web/src/workbook/timeline/models/workbookMentionChips.ts` | Mention chip, inspector mention, auto-resolution models | Types, constants, functions | none | `/apps/web` mention model | Production authored private model | Entity mention collection item shape | High | mention chip, phase4/5 tests, visual | Must preserve raw/resolved/dismissed distinctions. |
| `apps/web/src/workbook/timeline/models/workbookTimelineModel.test.ts` | Timeline model tests | none | Vitest, surface ID, pending queue type, timeline model | `/apps/web` tests | Test authored | Timeline rows, payloads, layout widths | High | Self test | Strong hot-path characterization. |
| `apps/web/src/workbook/timeline/models/workbookTimelineModel.ts` | Timeline row, payload, editor, patch, paste, create model | Many types and functions | view-contracts, surface ID, pending queue type, value formatter, mention/rows models, clipboard re-export | `/apps/web` timeline model | Production authored private model | Timeline view contract, `view_row_v1`, patch payloads, create payloads, collection actions | High | timeline model and broad phase tests | Large model but less risky than component controller. |
| `apps/web/src/workbook/timeline/services/workbookShellPhase4.ts` | Collaboration and mention payload helper service | Types and helper functions | mention relationship type | `/apps/web` collaboration/service helper | Production authored private service | `record_changed` payload, mention action payloads | High | phase4 support, socket lifecycle tests | Name is phase-shaped but production runtime import was observed. Rename only in behavior-preserving slice with characterization. |
| `apps/web/src/workbook/timeline/services/workbookSocketLifecycle.test.ts` | Socket lifecycle reducer tests | none | Vitest, record changed type, socket lifecycle | `/apps/web` tests | Test authored | WebSocket lifecycle effects | Medium | Self test | Reducer characterization. |
| `apps/web/src/workbook/timeline/services/workbookSocketLifecycle.ts` | WebSocket lifecycle reducer | Types and reducer functions | record changed type | `/apps/web` collaboration service | Production authored private service | `hello_ack`, `resume_ack`, `record_changed`, refresh/invalidate behavior | High | socket lifecycle and phase6 tests | Keep app-owned, not backend-owned. |
| `apps/web/src/workbook/utils/GridAdapter.phase9.anchor.test.ts` | Adapter-owned anchor unit tests from workbook side | none | grid-adapter, Vitest | `/apps/web` tests with grid seam | Test authored | Grid anchors, paste target resolution | Medium | Self test, phase9 map | Test-only phase name acceptable. |
| `apps/web/src/workbook/utils/workbookClipboard.test.ts` | Clipboard utility tests | none | Vitest, clipboard utility | `/apps/web` tests | Test authored | TSV/CSV scalar/tabular detection | Low | Self test | Paste characterization. |
| `apps/web/src/workbook/utils/workbookClipboard.ts` | Clipboard parsing and tabular detection | Types and functions | none | `/apps/web` workbook utility | Production authored private utility | Clipboard paste dispatch | Medium | clipboard, paste sentinel, browser tests | Base-profile hot path. |
| `apps/web/src/workbook/utils/workbookContinuity.test.ts` | Continuity utility tests | none | Vitest, continuity utility | `/apps/web` tests | Test authored | Viewport anchor and scroll restoration | Medium | Self test | Focus/scroll characterization. |
| `apps/web/src/workbook/utils/workbookContinuity.ts` | Viewport anchor and scroll restoration helpers | Types and functions | none | `/apps/web` continuity utility | Production authored private utility | Scroll continuity | High | continuity and phase4/9 tests | Preserve arithmetic. |
| `apps/web/src/workbook/utils/workbookGridFocus.tsx` | Grid focus state and focusable cell wrapper | Types, hook, components | grid-adapter, ui-contracts, React, keyboard/styles | `/apps/web` focus utility | Production authored private utility | Row-cell selectors, keyboard navigation, focus anchor status | High | keyboard, grid, visual/a11y tests | Adapter public type dependency only. |
| `apps/web/src/workbook/utils/workbookKeyboard.test.ts` | Keyboard command tests | none | Vitest, keyboard utility | `/apps/web` tests | Test authored | Keyboard command map | Low | Self test, phase9 map | Good guard for shortcut refactors. |
| `apps/web/src/workbook/utils/workbookKeyboard.ts` | Workbook keyboard command mapper | Types and function | grid-adapter type | `/apps/web` keyboard utility | Production authored private utility | Grid navigation keys, command availability | High | keyboard unit and browser phase9 | Keep key matrix aligned with design/Core. |
| `apps/web/src/workbook/utils/workbookPendingQueue.test.ts` | Pending queue and save-state tests | none | Vitest, pending queue model | `/apps/web` tests | Test authored | Pending replay, coalescing, save state, conflict anchors | High | Self test, phase4/7 rows | Critical characterization. |
| `apps/web/src/workbook/utils/workbookPendingQueue.ts` | Pending replay queue model and save-state derivation | Types, constants, queue class, helpers | shared public error, timeline conflict parser | `/apps/web` pending replay model | Production authored private model | Pending replay identity, public errors, same-field conflict, save-state labels | High | pending queue tests, phase4/6 tests | Central behavior-preservation surface. |
| `apps/web/src/workbook/utils/workbookPresence.ts` | Presence filtering, sorting, initials | Types and functions | startup sheet ref type | `/apps/web` presence model | Production authored private utility | Presence payload shape, sheet refs | Medium | phase6 tests, visual/a11y | Keep canonical ordering. |
| `apps/web/src/workbook/utils/workbookStyles.ts` | Shared inline style constants | Styles and style functions | React type | `/apps/web` UI styles | Production authored private utility | Status strip, focus, presence visual states | Medium | visual/a11y indirect | Design-token migration would need owner path. |
| `apps/web/src/workbook/utils/workbookValueFormat.test.ts` | Value formatting tests | none | Vitest, formatter | `/apps/web` tests | Test authored | Grid display value coercion | Low | Self test | Simple utility. |
| `apps/web/src/workbook/utils/workbookValueFormat.ts` | Grid display value stringifier | `stringifyGridValue` | none | `/apps/web` utility | Production authored private utility | Visible cell formatting | Medium | value format and model tests | Used broadly. |

No in-scope file was identified as generated by `tools/generated_artifact_policy.json`.

## 6. Contract freeze matrix

| Surface | Specific contract | Owner source | Current evidence | Characterization required | Drift risk |
| --- | --- | --- | --- | --- | --- |
| HTTP query | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, full `view_row_v1`, filters, sort, group, pagination body members | Core 01 §3.3.4, Core 03 §2 and §3 | `workbookQuery.test.ts`, `WorkbookShell.phase4.timelineQuery.test.tsx`, `WorkbookShell.phase8.query.test.tsx`, browser phase8 specs | Add only if moving query fetch orchestration out of current files | High |
| HTTP row create and patch | `view_schema_id`, `field_key`, `client_txn_id`, `base_row_version`, changed fields only | Core 01 route contracts, Core 03 §3 | `workbookTimelineModel.test.ts`, `genericWorkbookModel.test.ts`, `WorkbookShell.phase3.payload.test.tsx`, `WorkbookShell.surfaces.test.tsx` | Required before moving payload builders or mutation submission | High |
| Saved views | Saved-view list/create/update/duplicate/delete, active surface scope, `query_json`, `layout_json`, home/default pointers | Core 01 §3.3.4.2 and §3.3.5.2, Core 03 §2.3 and §2.4 | `workbookSavedViews.test.ts`, `workbookSavedViewRuntime.test.ts`, `workbookStartup.test.ts`, `WorkbookShell.surfaces.test.tsx`, browser phase8 specs | Existing evidence is adequate for first saved-view/query seam if run | High |
| Workbook startup | Explicit launch, persisted home/default fallback, cleared pointers, base surface fallback | Core 01 startup route, Core 03 §2.4 | `workbookStartup.test.ts`, browser phase8 startup specs | Required if startup URL or runtime state moves | High |
| WebSocket and live refresh | `/ws/v1/*`, ack lifecycle, record_changed patch/invalidate, presence, self-originated ignore | Core 01 collaboration payloads, Core 03 §4 | `workbookSocketLifecycle.test.ts`, `WorkbookShell.phase6.test.tsx`, `timelineWorkbookTestSupport.ts` | Required before socket lifecycle moves | High |
| Pending replay | Queue capacity, coalescing, retry, auth pause, conflict halt, save-state derivation | Core 03 §4 | `workbookPendingQueue.test.ts`, phase4/6 tests | Existing model tests strong; run before/after movement | High |
| Same-field conflict | Conflict keyed by field, explicit resolver, cell-local conflict state, no silent overwrite | Core 03 §3.2 and §3.3 | `timelineConflictModel.test.ts`, `workbookPendingQueue.test.ts`, `WorkbookShell.phase6.test.tsx`, visual FE-P7 | Required before conflict UI or queue movement | High |
| Workbook UI | Grid-first, low-friction row creation, inline edit, paste, no detached forms for hot path | Core 03 §1, §7, §11; design §4 and §7 | phase3/4/9 unit tests, browser phase9 keyboard and sentinel specs, visual specs | Browser validation required for UI-affecting moves | High |
| Focus, selection, scroll continuity | Preserve selected `record_id`, grid scroll, deterministic focus target after same-surface mutations | Core 01 §3.1, Core 03 §3.1, design §8 | `workbookContinuity.test.ts`, phase4 support tests, phase9 keyboard specs, test-utils anchor helpers | Required for grid/focus code movement | High |
| Inspector default closed | Inspector closed on workbook open, surface switch, saved-view switch, refresh; explicit open only | Core 03 §2.3 and §2.3A, design §7.3 | `workbookInspectorModel.test.ts`, `WorkbookShell.phase9.inspector.test.tsx`, visual/a11y specs | Required before inspector state movement | High |
| Inspector retargeting | No-row state, clear row-bound forms and previews on selected-row or row-version change | Core 03 §2.3A | `workbookInspectorModel.test.ts`, phase7/9 tests | Required before history/inspector movement | High |
| Generated contracts | Apps consume facades, not generated internals; generated files not hand-edited | Harness NLSpec, generated policy, import boundaries | Targeted scan found no production generated-subpath imports | Run `make generated-artifact-policy-check` and `make frontend-import-boundary-check` for code slices | Medium |
| UI selectors and test IDs | Runtime and tests share stable selector builders from `@cartulary/ui-contracts` | Frontend guide, ui-contracts package, design selector discipline | Broad imports from `@cartulary/ui-contracts`; some literal production IDs observed | Characterize literals before moving or replacing | Medium |
| Harness accounting | Make-owned targets, row accounting, frontend maps and retained artifacts | Testing harness NLSpec | `tools/frontend_phase_maps/fe_p*_test_map.json`, `tools/phase*_test_map.json`, and retained artifact roots for this artifact were inspected | Future code slices must run and name their own retained roots before claiming pass | Medium |
| Visual readiness | Visual goldens and fixtures remain design/support evidence, not product conformance | Visual golden guide, design, harness NLSpec | `workbook.visual.spec.ts`, fixture registry by targeted search | Required only for visual/layout changes | Medium |
| Accessibility readiness | Keyboard/focus/a11y readiness evidence remains support/design unless owner allows product conformance | Design, frontend guide, harness NLSpec | `workbook.a11y.spec.ts`, `workbook.a11y-preflight.spec.ts` by targeted search | Required only for interaction/a11y-affecting UI movement | Medium |

Behavior-preserving refactor candidates:

- Extract a saved-view/query runtime seam from `WorkbookShell.tsx` using existing models and tests.
- Extract small shell/status/presence UI facades without changing selector builders or status text.
- Move timeline pending-save runtime wiring around `WorkbookPendingQueueModel` only after running current unit coverage.
- Split inspector/history state wrappers after preserving default-closed and retargeting tests.

Behavior changes are out of scope unless explicitly authorized by owner docs and should start in owner docs, then contracts, generation, and implementation.

## 7. Boundary and coupling findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required fix or rationale |
| --- | --- | --- | --- | --- | --- |
| `WorkbookShell.tsx` carries shell, generic/entity/assessment/evidence surfaces, saved views, startup, HTTP mutations, and inspector UI | File exports `WorkbookShell`, includes `EntityWorkbookSurface`, `AssessmentWorkbookSurface`, `GenericWorkbookSurface`, and many HTTP helpers | High | should_fix | `/apps/web` workbook shell | Split by behavior-preserving seams after saved-view/query characterization. |
| `TimelineWorkbook.tsx` carries Timeline query, mutation, WebSocket, pending replay, conflicts, inspector, evidence, presence, focus, and layout | File has about 7k lines and imports most timeline hooks/models/components | High | should_fix | `/apps/web` timeline controller | Do not split first; stabilize shell/query seams before hot-path movement. |
| Production file name `workbookShellPhase4.ts` is phase-shaped | Runtime imports from `TimelineWorkbook.tsx` and tests | Medium | should_fix | `/apps/web` collaboration service | Rename only as a small isolated slice with no behavior change and import-boundary validation. |
| Production scan found no direct `react-data-grid` import outside `/packages/grid-adapter` | Targeted scan returned no production matches | Medium | intentional/no_action | `/packages/grid-adapter` | Preserve with `make frontend-import-boundary-check`; no current fix. |
| Production scan found no direct generated-subpath imports | Targeted scan returned no production generated-internal matches | Medium | intentional/no_action | `/packages/protocol-ts`, `/packages/ui-contracts` | Preserve facade imports only. |
| Apps/web uses grid adapter public coordinate and anchor types | `GridColumn`, `GridRow`, `GridCellAnchor`, `GridNavigationKey` imports from package facade | Medium | intentional/no_action | `/packages/grid-adapter` public API | Accept because package facade owns vendor translation; do not import vendor row/column semantics. |
| View contract internals are adapted in app models | `workbookQuery.ts`, `workbookContractRows.ts`, `workbookTimelineModel.ts`, generic/entity models read `ViewContract` fields | Medium | defer | `/apps/web` plus `/packages/view-contracts` | Current use is through facade. Move parsing to package only if duplication or leakage grows. |
| Stable selectors mostly flow through `@cartulary/ui-contracts`, but some literal production IDs remain | `WorkbookShell.tsx` and components contain literal `data-testid` values such as merge and assessment controls | Medium | should_fix | `/packages/ui-contracts` for shared selectors, `/apps/web` for private literals | Inventory literals before conversion; avoid churn without tests. |
| Browser choreography lives outside runtime packages | `packages/test-utils` and `apps/web/src/testing` own helper interactions | Medium | intentional/no_action | `/packages/test-utils`, `/apps/web` tests | Runtime must not import test support; guard with import-boundary check. |
| Phase-shaped tests and maps are broad | Many `WorkbookShell.phase*.test.*` and phase maps reference workbook files | Low | intentional/no_action | Harness/evidence accounting | Accept as evidence naming only; do not mirror phase identity in runtime. |
| Visual/a11y evidence could be overclaimed | Visual and a11y specs cover many workbook states | Medium | must_fix | Handoff and validation discipline | Keep evidence class explicit as design/support unless Core 05 boundary is active. |
| Backend route/store evidence now inventoried | Route, store, contract, query, and migration owners for workbook-facing backend behavior are identified, but backend implementation remains outside this frontend refactor tracker | Medium | defer | Backend module owners | Future slices inspect backend code deeper only when they change public contracts, route assumptions, authorization semantics, or generated surfaces. |

## 8. Workflow dependency map

| Workflow | Name | Class | Required previous workflows | Required subsequent workflows |
| --- | --- | --- | --- | --- |
| WF-00 | Session and source bootstrap | root | none | WF-01 |
| WF-01 | Current-state repository scan | chain | WF-00 | WF-02, WF-03 |
| WF-02 | Module ownership inventory | chain | WF-01 | WF-04, WF-05 |
| WF-03 | Public contract freeze | chain | WF-01 | WF-04, WF-05 |
| WF-04 | Refactor slice selection | chain | WF-02, WF-03 | WF-05, WF-06 |
| WF-05 | Characterization test plan | chain | WF-03, WF-04 | WF-09 |
| WF-06 | Boundary guardrail plan | chain | WF-02, WF-04 | WF-09 |
| WF-08 | Frontend package seam plan | parallel | WF-04, WF-05, WF-06 | WF-09 |
| WF-09 | Execution checkpoint plan | chain | WF-05, WF-08 | WF-10 |
| WF-10 | Validation and harness accounting plan | chain | WF-09 | WF-11 |
| WF-11 | Documentation and generated-artifact plan | parallel | WF-03, WF-09 | WF-12 |
| WF-12 | Cleanup and anti-drift plan | chain | WF-10, WF-11 | WF-13 |
| WF-13 | Handoff and next-slice bootstrap | chain | WF-12 | none |

WF-07 backend module facade plan is omitted because no backend imports or backend code are in this primary target scope.

## 9. Target-specific workflows

### WF-00 session/source bootstrap

Objective: establish exact local context for one workbook refactor session.

Dependencies: none.

Steps:

1. Record branch, commit, dirty tree, timestamp, target directory, artifact path, and user constraints.
2. Read `AGENTS.md` and the framework before edits.
3. Confirm the artifact is the only intended repo change.
4. Record source limits and uninspected adjacent files.

Outputs: section 1, section 2, current handoff record.

Acceptance criteria: one target seam is named, production edits are disallowed, and missing evidence uses an explicit open-evidence marker.

Validation: `git status --short --branch`, `git rev-parse HEAD`, `date -Iseconds`.

Handoff notes: rerun this workflow after any context compaction or branch change.

### WF-01 current-state repository scan

Objective: list current workbook files, imports, exports, tests, package seams, and generated-artifact boundaries.

Dependencies: WF-00.

Steps:

1. Run `rg --files apps/web/src/workbook`.
2. Inspect imports and exports for all in-scope files.
3. Inspect imported adjacent packages and test helpers only where workbook uses them.
4. Search for direct RDG imports, generated internals, phase runtime references, and test-helper runtime leaks.
5. Distinguish authored source from generated artifacts.

Outputs: section 5, source-limit notes, coupling evidence.

Acceptance criteria: all in-scope files are listed and generated status is explicit.

Validation: `rg` scans plus future `make frontend-import-boundary-check`.

Handoff notes: do not infer unseen backend behavior from frontend model names.

### WF-02 module ownership inventory

Objective: assign every file and behavior to a workbook or package seam.

Dependencies: WF-01.

Steps:

1. Map app files to shell, query, mutation, conflict, inspector, presence, timeline, model, utility, or test evidence responsibilities.
2. Map package imports to grid adapter, view contracts, UI contracts, and test utils.
3. Mark any unclear ownership with an explicit ownership-decision marker.

Outputs: section 5 inventory and section 7 findings.

Acceptance criteria: every production file has one owner or an explicit ownership-decision marker.

Validation: review imports against `tools/frontend_import_boundaries.json`.

Handoff notes: shared helpers must be justified by semantic ownership, not convenience.

### WF-03 public contract freeze

Objective: identify behavior that must not drift during refactor.

Dependencies: WF-01.

Steps:

1. Map HTTP, WebSocket, UI, selector, generated-artifact, harness, visual, and a11y surfaces.
2. Link each surface to owner docs and current test evidence.
3. Separate behavior-preserving candidates from out-of-scope behavior changes.

Outputs: section 6.

Acceptance criteria: every public surface that could drift is mapped.

Validation: run characterization targets from section 11 before code changes.

Handoff notes: behavior-affecting changes must start in owner docs.

### WF-04 refactor slice selection

Objective: select small independent refactor slices.

Dependencies: WF-02, WF-03.

Steps:

1. Pick one seam per slice.
2. Name files likely touched, behavior expected unchanged, characterization evidence, rollback point, and stop condition.
3. Prefer low-risk seams before hot-path Timeline movement.

Outputs: section 10.

Acceptance criteria: each slice is reviewable and has validation.

Validation: no code validation until a slice is implemented.

Handoff notes: recommended first slice is saved-view/query runtime seam extraction.

### WF-05 characterization test plan

Objective: preserve current behavior before moving code.

Dependencies: WF-03, WF-04.

Steps:

1. Reuse existing unit tests for models, reducers, selectors, and queue behavior.
2. Use browser specs only when public route or real browser interaction behavior is touched.
3. Use visual/a11y targets only for visual, layout, keyboard, focus, or accessibility readiness movement.
4. Add missing characterization only when a high-risk movement lacks a local guard.

Outputs: section 11.

Acceptance criteria: every high-risk slice has a pre-move evidence target or explicit open-evidence marker.

Validation: Make targets only when evidence is claimed.

Handoff notes: direct Vitest or Playwright commands are developer conveniences, not canonical evidence.

### WF-06 boundary guardrail plan

Objective: keep dependency direction clean during refactors.

Dependencies: WF-02, WF-04.

Steps:

1. Preserve `@cartulary/grid-adapter` as the only grid vendor boundary.
2. Preserve `@cartulary/view-contracts` as contract adapter boundary.
3. Preserve `@cartulary/ui-contracts` as stable selector and design-token facade.
4. Preserve `@cartulary/test-utils` and `apps/web/src/testing` as test-only helpers.
5. Keep phase identity out of production runtime.

Outputs: section 7 and section 11.

Acceptance criteria: each guardrail has a validation command or source-limit note.

Validation: `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, targeted production scans.

Handoff notes: new package exports must go through package names and `package.json` exports.

### WF-08 frontend package seam plan

Objective: keep package responsibilities intact while app code is split.

Dependencies: WF-04, WF-05, WF-06.

Steps:

1. Keep workbook shell and application state in `/apps/web`.
2. Keep vendor grid coordinate translation in `/packages/grid-adapter`.
3. Keep generated schema parsing in `/packages/view-contracts`.
4. Keep stable selector construction in `/packages/ui-contracts`.
5. Keep browser helper choreography in `/packages/test-utils`.

Outputs: package seam rows in sections 7, 10, and 12.

Acceptance criteria: `/apps/web` does not learn grid vendor coordinate semantics.

Validation: `make frontend-import-boundary-check`, targeted production scans.

Handoff notes: extracting app-internal facades is preferred before creating new shared packages.

### WF-09 execution checkpoint plan

Objective: convert a chosen slice into ordered edits.

Dependencies: WF-05, WF-08.

Steps:

1. Create one checkpoint per small movement.
2. Run narrow validation after each checkpoint.
3. Roll back at the file-level by reverting only the slice files if validation fails.
4. Stop before broad refactor creep.

Outputs: section 10 slice checkpoints.

Acceptance criteria: every checkpoint has expected unchanged behavior and validation.

Validation: slice-specific targets from section 11.

Handoff notes: do not combine saved-view, grid, inspector, and pending queue movement in one checkpoint.

### WF-10 validation and harness accounting plan

Objective: name the cheapest sufficient validation for behavior preservation.

Dependencies: WF-09.

Steps:

1. Choose Make-owned targets.
2. Distinguish product conformance from implementation support and design readiness.
3. Record retained artifacts only when available.
4. Run `make agent-finalize` before broad end-of-run verification when a broader implementation session completes.

Outputs: section 11 and section 13.

Acceptance criteria: validation commands and failure handling are explicit.

Validation: commands in section 11.

Handoff notes: visual and a11y targets do not replace product behavior tests.

### WF-11 documentation and generated-artifact plan

Objective: decide whether docs, contracts, or generated artifacts are needed.

Dependencies: WF-03, WF-09.

Steps:

1. For behavior-preserving refactors, update only implementation-support docs or this handoff when useful.
2. For behavior changes, update owner docs first, then derived contracts, then generated code, then implementation.
3. If generated artifacts are involved, name owner input and drift check.

Outputs: section 12 generated artifacts and docs notes.

Acceptance criteria: docs/contracts/generation classification is explicit.

Validation: `make generated-artifact-policy-check`, `make generate-drift`, `make json-shape-check` when relevant.

Handoff notes: do not edit generated roots by hand.

### WF-12 cleanup and anti-drift plan

Objective: avoid duplicate wrappers, stale imports, and evidence drift after a slice.

Dependencies: WF-10, WF-11.

Steps:

1. Remove old wrappers only after callers are moved.
2. Re-run import-boundary and typecheck targets.
3. Review diff for unrelated behavior changes.
4. Record intentional deferrals.

Outputs: section 12 risks and section 13 handoff.

Acceptance criteria: no orphaned old path remains unless deferred.

Validation: `make frontend-typecheck`, `make frontend-import-boundary-check`, `git diff --check`.

Handoff notes: never revert unrelated user changes.

### WF-13 handoff and next-slice bootstrap

Objective: allow another agent to resume safely.

Dependencies: WF-12.

Steps:

1. Update section 4 tracker statuses.
2. Record changed files, commands, results, decisions, open questions, and blockers.
3. Name the next recommended workflow and safe restart command.

Outputs: section 13.

Acceptance criteria: another agent can continue without rediscovery.

Validation: scan this artifact for open-evidence and blocker markers before handoff.

Handoff notes: start next session with WF-00 if branch, commit, or dirty tree changed.

## 10. Refactor slice plan

| Slice | Depends on | Change | Files likely touched | Behavior expected unchanged | Characterization evidence | Validation | Rollback note | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 tracker baseline | none | Maintain this tracker only | This artifact | No runtime behavior | Source scans in section 1 | `make generated-artifact-policy-check`, `make json-shape-check` | Revert this artifact only | DONE |
| S-01 saved-view/query runtime seam | S-00 | Extract saved-view selection and query-state orchestration from `WorkbookShell.tsx` into an app-owned runtime module or hook | `WorkbookShell.tsx`, `models/workbookQuery.ts`, `models/workbookSavedViewRuntime.ts`, `models/workbookSavedViews.ts`, `hooks/useWorkbookShellRuntime.ts`, tests if needed | Saved-view selector IDs, `query_json`, `layout_json`, startup/default behavior, active surface | `workbookQuery.test.ts`, `workbookSavedViewRuntime.test.ts`, `workbookSavedViews.test.ts`, `workbookStartup.test.ts`, `WorkbookShell.surfaces.test.tsx`, `WorkbookShell.phase8.query.test.tsx` | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check` | Stop after extracted seam compiles and tests pass; revert new module and import changes if behavior changes | DONE |
| S-02 shell slot/status facade | S-01 | Extract status strip and shell slot composition without changing selectors or layout | `WorkbookShell.tsx`, `components/WorkbookStatusStrip.tsx`, `components/WorkbookSurfaceFrame.tsx`, `components/WorkbookShellSlots.tsx` | Shell ready ID, slot IDs, status strip labels, presence summary | `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase6.test.tsx`, visual/a11y specs if layout changes | `make frontend-unit`, `make frontend-typecheck`; browser visual/a11y only if layout changes | Revert component wiring only | DONE |
| S-03 generic surface controller split | S-01 | Move generic create/edit/evidence surface controller logic behind an app-owned facade | `WorkbookShell.tsx`, `models/genericWorkbookModel.ts`, `models/workbookContractRows.ts`, `components/GenericMutationControl.tsx` | Generic mutation payloads, evidence handles, party links, selectors | `genericWorkbookModel.test.ts`, `WorkbookShell.surfaces.test.tsx`, `WorkbookShell.phase5.gridProvenance.test.tsx`, browser phase4 generic specs | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`; browser target if public route flow changes | Stop before moving evidence preview/download into shared package | DONE |
| S-04 timeline pending-save runtime seam | S-01 | Encapsulate pending replay hook wiring around existing `WorkbookPendingQueueModel` | `TimelineWorkbook.tsx`, `hooks/useTimelinePendingSaves.ts`, `utils/workbookPendingQueue.ts` | Queue capacity, coalescing, retry, halt, save-state copy, conflict anchors | `workbookPendingQueue.test.ts`, `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase6.test.tsx` | `make frontend-unit`, `make frontend-typecheck` | Revert hook wiring if any queue test or save state label drifts | DONE |
| S-05 inspector and history state seam | S-04 | Isolate inspector/history state transitions without changing panel rendering or route calls | `TimelineWorkbook.tsx`, `TimelineWorkbookInspector.tsx`, `TimelineHistoryPanel.tsx`, `hooks/useTimelineInspectorSelection.ts`, `hooks/useTimelineHistoryState.ts`, `models/workbookInspectorModel.ts` | Default-closed, no-row state, retargeting, stale confirmation invalidation, rollback action visibility | `workbookInspectorModel.test.ts`, `WorkbookShell.phase7.test.tsx`, `WorkbookShell.phase9.inspector.test.tsx` | `make frontend-unit`, `make frontend-typecheck`; browser stateful if route flow changes | Stop before altering visible inspector layout | DONE |
| S-06 selector literal inventory and cleanup | S-01 | Convert shared literal `data-testid` values to `@cartulary/ui-contracts` only where selectors cross runtime and tests | `WorkbookShell.tsx`, selected components, `packages/ui-contracts/src/index.ts`, tests | Existing selector strings and browser helpers | Existing tests that query affected literals | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check` | Do not touch generated design tokens; revert selector API if broad churn appears | DONE |
| S-07 phase-shaped production helper rename | S-04 | Rename `workbookShellPhase4.ts` to a behavior-named collaboration or mention service | `timeline/services/workbookShellPhase4.ts`, imports, tests | Payload shapes and WebSocket reducer behavior | `workbookSocketLifecycle.test.ts`, phase4 support tests, phase6 tests | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check` | Pure rename plus import changes; no logic edits | DONE |
| S-08 browser characterization gap fill | Any high-risk UI slice | Add minimal characterization only when existing tests miss a moved behavior | Tests only, no production code | Product behavior unchanged | Decide after slice diff | `make frontend-unit` or browser target matching the gap | Revert added test if it encodes implementation detail instead of behavior | NA: no active high-risk UI slice exposed an uncovered behavior gap |

Next recommended implementation slice: none; S-02 through S-07 are DONE and S-08 is NA because the conditional trigger did not occur.

## 11. Validation plan

| Target or command | Purpose | Required | Expected artifact | Failure handling | Evidence class |
| --- | --- | --- | --- | --- | --- |
| `make generated-artifact-policy-check` | Confirm generated roots and policy remain valid after artifact-only or generated-adjacent changes | Required for this artifact and generated-adjacent sessions | Retained tool summary | Report target, run root when available, and whether related | Implementation support |
| `make json-shape-check` | Validate repo JSON manifests after planning changes and before relying on manifests | Required for this artifact | Retained tool summary | Report manifest/schema failure and relation | Implementation support |
| `make lint-markdown` | Markdown lint for authored active docs | Optional here because `.markdownlint-cli2.jsonc` did not include `docs/handoffs/**` at scan time | Retained tool summary if run | If skipped, record scope reason; if failed, report existing vs related | Implementation support |
| `make frontend-unit` | Unit and integration-style frontend characterization for workbook models/components | Required for most code slices | Retained frontend unit summary | Stop slice and inspect failing tests | Product conformance or support depending row |
| `make frontend-typecheck` | Type safety for app/package splits | Required for code slices | Retained typecheck summary | Stop and fix type break in slice | Implementation support |
| `make frontend-import-boundary-check` | Enforce RDG, generated, test-helper, and package facade boundaries | Required for code slices | Retained boundary summary | Stop on new violation; do not bypass | Implementation support |
| `make browser-e2e-stateful` | Real browser and service-backed stateful behavior | Required when public browser route or persistence behavior changes | Retained browser summary | Report failure class and run root | Product conformance where owner-mapped |
| `make browser-e2e-webserver-backed` | Browser flow validation with webserver-backed stack | Required when startup, saved-view, or shell routing changes broadly | Retained browser summary | Report target and failure class | Product conformance where owner-mapped |
| `make browser-e2e-visual` | Visual regression and frontend visual readiness | Only for visual/layout changes | Visual summary and screenshots | Do not update goldens unless explicitly authorized | Design direction or implementation support |
| `make browser-e2e-a11y` | Accessibility readiness rows | Only for keyboard/focus/a11y changes | `frontend-accessibility-summary.json` | Treat as readiness evidence unless owner says otherwise | Design direction or implementation support |
| `make browser-e2e-a11y-preflight` | Blocked or mapped preflight smoke | Only when phase map routes a preflight row | Preflight summary | Do not claim completion for blocked rows | Implementation support |
| `make agent-finalize` | End-of-run retained evidence maintenance | Required before broader final handoff if a full implementation session ran | Agent finalize summary | If `RESULTS_DIR` omitted, report skipped retained-run maintenance | Implementation support |

Cheapest sufficient validation by slice:

- S-01: `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`.
- S-02: add visual or a11y only if layout, focus, or accessible names change.
- S-03: add browser target only if public route sequencing or evidence handle flow changes.
- S-04: `make frontend-unit` is the first gate because pending queue has strong model coverage.
- S-05: run browser target if inspector route calls or row-history flows change.
- S-06: add browser/visual/a11y targets only if selector cleanup changes browser helper choreography, layout, focus, accessible names, or visual readiness surfaces.
- S-07: run `make frontend-unit`, `make frontend-typecheck`, and `make frontend-import-boundary-check`; no browser target is required for an import-only rename with byte-equivalent helper content.
- S-08: mark NA unless an active high-risk UI slice exposes a behavior gap not already covered by existing characterization.

## 12. Workstream notes

### Scope and evidence

| Date | Note | Source or command | Impact |
| --- | --- | --- | --- |
| 2026-06-27 | Primary target is exactly `apps/web/src/workbook` with 108 in-scope files. | `rg --files apps/web/src/workbook` | Scope is explicit. |
| 2026-06-27 | No generated files detected under target by generated-marker search. | `rg "Code generated|DO NOT EDIT" apps/web/src/workbook` | Treat target as authored source. |
| 2026-06-27 | Backend, generated, e2e, phase-map, retained-artifact, contract, and migration evidence gaps were closed by targeted remediation inspection. | Source limit update | Future code slices still inspect deeper only when they cross behavior, contract, authorization, generated, or storage boundaries. |

### Adjacent evidence closure

| Gap closed | Evidence inspected | Remediation decision |
| --- | --- | --- |
| Backend route/store evidence | `internal/modules/workbook`, `savedviews`, `revisions`, `evidence`, `timeline`, `viewschemas`; `internal/platform/viewquery`, `viewschema`, `ws`; relevant backend tests by targeted file and route scans | Backend owners are known; backend implementation remains out of scope unless public contracts, route assumptions, authorization, or generated surfaces change. |
| Generated protocol internals | `packages/protocol-ts/src/generated/{index.ts,contracts.ts}`, `packages/ui-contracts/src/generated/design-tokens.ts` | Treat as read-only evidence; use package facades and generated-artifact policy checks. |
| E2E coverage outside targeted workbook specs | All 41 `apps/web/e2e/*.spec.ts` and `*.test.ts` files were inventoried | Browser evidence is phase-distributed; select Make targets by behavior, not by filename alone. |
| Frontend phase maps and retained artifacts | `tools/frontend_phase_maps/fe_p0_test_map.json` through `fe_p11_test_map.json`, `tools/phase*_test_map.json`, retained roots for artifact checks | Phase maps are harness accounting; retained roots only prove commands that ran in their session. |
| Contracts, DB, backend source | 18 `contracts/view-schemas/*.json`, `contracts/ws/index.schema.json`, `contracts/openapi/cartulary.openapi.yaml`, migrations `00003` through `00020`, `db/queries/{incidents_phase2.sql,savedviews_phase8.sql,timeline_phase3.sql}` | Behavior changes are spec/contract-first and may require generation, migration, and drift validation. |

### Contracts and docs

| Date | Owner section | Decision or conflict | Action |
| --- | --- | --- | --- |
| 2026-06-27 | Core 01, Core 03 | Workbook contracts and interaction behavior must not drift. | Freeze surfaces in section 6. |
| 2026-06-27 | Harness NLSpec | Visual and accessibility are readiness/support unless Core 05 publication boundary is active. | Keep evidence class explicit. |
| 2026-06-27 | Domain and design docs | Stable identifiers, not labels or vendor coordinates, own behavior. | Preserve selector and field-key discipline. |
| 2026-06-27 | Contracts and migrations | View-schema, WebSocket, OpenAPI, workbook migrations, and saved-view query inputs are downstream of owner specs. | Future behavior changes start in owner specs/contracts, not implementation. |

### Frontend packages

| Date | Package | Files | Current state | Next action |
| --- | --- | --- | --- | --- |
| 2026-06-27 | `/packages/grid-adapter` | `core.ts`, `index.tsx`, `test-support.tsx` | Owns grid facade, RDG stylesheet, vendor details, test renderer. | Preserve facade imports. |
| 2026-06-27 | `/packages/view-contracts` | `index.ts` | Owns generated view-schema parsing and row normalization facade. | Do not move app state into package. |
| 2026-06-27 | `/packages/ui-contracts` | `index.ts`, generated design-token facade | Owns stable selectors and design-token facade. | Convert literals only when shared. |
| 2026-06-27 | `/packages/test-utils` | `index.ts` | Owns browser helper choreography. | Keep out of runtime. |

### Tests and harness

| Date | Target | Result | Artifact | Follow-up |
| --- | --- | --- | --- | --- |
| 2026-06-27 | `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260627T214933Z-p1300061` | Completed after remediation update. |
| 2026-06-27 | `make json-shape-check` | PASS | `.cartulary/test-results/20260627T214933Z-p1300076` | Completed after remediation update. |
| 2026-06-27 | `make lint-markdown` | Skipped | n/a | `.markdownlint-cli2.jsonc` did not include `docs/handoffs/**` at scan time; config was not changed. |
| 2026-06-27 | E2E inventory | INSPECTED | 41 `apps/web/e2e/*.spec.ts` and `*.test.ts` files | Workbook coverage spans phase2 through phase10, frontend phase specs, visual/a11y specs, and support/helper tests. |
| 2026-06-27 | Frontend phase maps | INSPECTED | FE-P0 through FE-P11 maps | Use as harness accounting only; production runtime must not depend on phase identity. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T225528Z-p1371544` | S-02 shell/status facade validation. |
| 2026-06-27 | `make frontend-typecheck` | PASS | `.cartulary/test-results/20260627T225528Z-p1371573` | S-02 type-safety validation. |
| 2026-06-27 | `make frontend-typecheck` | FAIL | `.cartulary/test-results/20260627T230950Z-p1386418` | S-03 stale unused imports after extracting the generic facade; related to the slice and fixed before continuing. |
| 2026-06-27 | `make frontend-typecheck` | PASS | `.cartulary/test-results/20260627T231028Z-p1387110` | S-03 type-safety validation after import cleanup. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T231051Z-p1387653` | S-03 generic payload, evidence handle, party-link, and surface characterization validation. |
| 2026-06-27 | `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260627T231113Z-p1389385` | S-03 app-owned facade import-boundary validation. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T231926Z-p1394014` | S-04 pending queue model, save-state, conflict, retry, halt, and replay characterization validation. |
| 2026-06-27 | `make frontend-typecheck` | PASS | `.cartulary/test-results/20260627T231948Z-p1395745` | S-04 pending runtime seam type-safety validation. |
| 2026-06-27 | `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260627T231948Z-p1395767` | S-04 app-owned hook/helper import-boundary validation. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T232721Z-p1400617` | S-05 inspector/history default-closed, retargeting, stale confirmation invalidation, rollback, and history characterization validation. |
| 2026-06-27 | `make frontend-typecheck` | PASS | `.cartulary/test-results/20260627T232743Z-p1402349` | S-05 inspector/history hook seam type-safety validation. |
| 2026-06-27 | `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260627T232743Z-p1402479` | S-05 app-owned hook import-boundary validation. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T233220Z-p1405529` | S-06 selector facade and affected runtime/unit selector validation. |
| 2026-06-27 | `make frontend-typecheck` | PASS | `.cartulary/test-results/20260627T233242Z-p1407260` | S-06 ui-contracts selector facade type-safety validation. |
| 2026-06-27 | `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260627T233242Z-p1407322` | S-06 package facade import-boundary validation. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T233800Z-p1410826` | Initial S-07 import-only rename characterization before the stale test-support import was found by typecheck. |
| 2026-06-27 | `make frontend-typecheck` | FAIL | `.cartulary/test-results/20260627T233819Z-p1412552` | S-07 stale `apps/web/src/testing/timelineWorkbookTestSupport.ts` import after helper rename; related to the slice and fixed before continuing. |
| 2026-06-27 | `make frontend-typecheck` | PASS | `.cartulary/test-results/20260627T233852Z-p1413318` | S-07 type-safety validation after updating the test-support import. |
| 2026-06-27 | `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260627T233907Z-p1413785` | S-07 behavior-named helper import-boundary validation. |
| 2026-06-27 | `make frontend-unit` | PASS | `.cartulary/test-results/20260627T233928Z-p1414349` | Final S-07 characterization after all import updates. |
| 2026-06-27 | `make agent-finalize` | PASS | `.cartulary/test-results/20260627T234125Z-p1416711` | End-of-run retained evidence maintenance; generated files unchanged, `RESULTS_DIR` omitted so duration/run-check retained-run maintenance was skipped. |
| 2026-06-27 | `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260628T001850Z-p1521097` | Section 16 remediation operating-plan generated-artifact guardrail validation. |
| 2026-06-27 | `make json-shape-check` | PASS | `.cartulary/test-results/20260628T001850Z-p1521163` | Section 16 remediation operating-plan manifest/schema validation. |
| 2026-06-27 | `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260628T001850Z-p1521122` | Section 16 facade-only package-boundary policy validation. |
| 2026-06-27 | `make lint-markdown` | PASS | `.cartulary/test-results/20260628T002028Z-p1524885` | Section 16 Markdown validation after final handoff update. |
| 2026-06-27 | `make agent-finalize` | PASS | `.cartulary/test-results/20260628T002039Z-p1525496` | End-of-run retained evidence maintenance after implementing section 16; generated files unchanged, `RESULTS_DIR` omitted so duration/run-check retained-run maintenance was skipped. |

### Generated artifacts

| Date | Generator or target | Outputs | Drift status | Follow-up |
| --- | --- | --- | --- | --- |
| 2026-06-27 | `tools/generated_artifact_policy.json` | `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, `tools/task_surface.generated.mk` | Not touched | Policy check passed for this artifact. |
| 2026-06-27 | Frontend phase ledgers and maps | `tools/frontend_phase_maps/*.json`, generated ledgers | Not touched | Future behavior changes must update owner inputs first. |
| 2026-06-27 | Protocol TS generated bundle | `packages/protocol-ts/src/generated/index.ts`, `packages/protocol-ts/src/generated/contracts.ts` | Inspected read-only | Generated by `tools/contractgen`; use `@cartulary/protocol-ts` facade. |
| 2026-06-27 | UI contracts generated bundle | `packages/ui-contracts/src/generated/design-tokens.ts` | Inspected read-only | Generated by `scripts/generate-design-tokens.mjs`; use `@cartulary/ui-contracts` facade. |

### UX hot path

| Date | Surface | Current state | Risk | Next action |
| --- | --- | --- | --- | --- |
| 2026-06-27 | Timeline grid | Hot path owned by `TimelineWorkbook.tsx`, timeline models, grid adapter facade | High | Avoid first large split; protect with phase3/4/6/9 tests. |
| 2026-06-27 | Inspector | Default closed and retargeting owned by Core 03 and model tests | High | Move only after saved-view/query seam. |
| 2026-06-27 | Saved views and query controls | Strong model tests and browser phase8 coverage | Medium | S-01 completed with existing characterization. |

### Risks and blockers

| ID | Risk or blocker | Owner | Blocking workflow | Resolution condition |
| --- | --- | --- | --- | --- |
| R-001 | Large controller files make accidental behavior drift likely. | `/apps/web` | WF-04, WF-09 | Use one-seam slices and existing characterization. |
| R-002 | Literal production test IDs may encode private selectors. | `/apps/web`, `/packages/ui-contracts` | WF-06, WF-08 | Inventory before conversion; preserve strings. |
| R-003 | Backend route/store owners are inventoried, but backend implementation remains outside the frontend refactor slice. | Backend modules | WF-03 | Inspect backend code deeper only if a slice changes route assumptions, contracts, authorization, generated surfaces, or storage behavior. |
| R-004 | Retained harness artifacts are available only for artifact-level validation, not future code slices. | Harness | WF-10 | Future code slices must run and name their own retained roots. |
| R-005 | Phase-shaped runtime helper name could invite production phase coupling. | `/apps/web` | WF-06, WF-09 | Resolved by S-07 import-only rename to `workbookCollaborationMessages.ts`. |

## 13. Session handoff

### Current handoff record

| Field | Value |
| --- | --- |
| Date/time | `2026-06-27T20:19:05-04:00` after section 16 remediation operating-plan implementation and validation |
| Branch/commit | `main` at `95125db0da2ebb2eeeb3092ed0699ea682fe29a1`, one commit ahead of `origin/main` |
| Target seam | Workbook remediation operating plan and handoff process |
| Current workflow | Section 16 now implements the durable remediation process for future workbook slices; no runtime code, generated files, contracts, migrations, or package manifests changed |
| Completed workflows | WF-00 refresh, AGENTS/procedure refresh, current tracker/owner-doc snippets, remediation matrix replacement, required contract-impact checklist, behavior-to-validation evidence map, generated/package-boundary rules, spec-first behavior-change workflow, frontend continuation backlog, workstream sequencing, validation ledger update, finalizer, and handoff update |
| Changed files | `docs/handoffs/apps_web_src_workbook_refactor_tracker.md` only |
| Commands run | `git status --short --branch`; `git rev-parse --abbrev-ref HEAD`; `git rev-parse HEAD`; `date -Iseconds`; targeted reads of `AGENTS.md`, this tracker, the refactor framework, and testing-harness visual/readiness language; `make generated-artifact-policy-check`; `make json-shape-check`; `make frontend-import-boundary-check`; `make lint-markdown`; `make agent-finalize`; targeted diff/status inspection |
| Passing validation | `make generated-artifact-policy-check` at `.cartulary/test-results/20260628T001850Z-p1521097`; `make json-shape-check` at `.cartulary/test-results/20260628T001850Z-p1521163`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260628T001850Z-p1521122`; `make lint-markdown` at `.cartulary/test-results/20260628T002028Z-p1524885`; `make agent-finalize` at `.cartulary/test-results/20260628T002039Z-p1525496` |
| Failing validation | None for this section 16 remediation operating-plan implementation |
| Decisions made | Implemented the requested remediation plan as process and handoff controls: future workbook slices must complete the contract-impact checklist, select browser evidence by behavior family, follow spec-first behavior-change sequencing, keep generated internals behind facades, classify selectors before promotion, and keep visual/a11y evidence separate from product conformance unless Core 05 applies; `make agent-finalize` ran without `RESULTS_DIR`, so duration/run-check retained-run maintenance was skipped and generated files remained unchanged |
| Open questions | None for this documentation/process implementation; future runtime seams remain in the continuation backlog and must be implemented one slice at a time under the new checklist |
| Blockers | None identified |
| Next recommended workflow | Start the next runtime remediation slice from the section 16 continuation backlog only after a fresh WF-00 refresh and completed contract-impact checklist |
| Safe restart command | `git status --short --branch && make generated-artifact-policy-check && make json-shape-check && make frontend-import-boundary-check && make lint-markdown && make agent-finalize` |

### Reusable handoff record template

| Field | Value |
| --- | --- |
| Date/time | TODO |
| Branch/commit | TODO |
| Target seam | TODO |
| Current workflow | TODO |
| Completed workflows | TODO |
| Changed files | TODO |
| Commands run | TODO |
| Passing validation | TODO |
| Failing validation | TODO |
| Decisions made | TODO |
| Open questions | TODO |
| Blockers | TODO |
| Next recommended workflow | TODO |
| Safe restart command | TODO |

Handoff rules:

- Do not claim validation passed unless the exact command ran in this session or the retained artifact is named.
- Do not claim a file was preserved, compared, or verified unless it was inspected.
- Use explicit open-evidence markers for missing evidence.
- Record whether dirty worktree changes are intentional.
- Record generated files that need regeneration or drift checks.

## 14. Top-level checklist

- [x] Explicit scope: target seam is `apps/web/src/workbook`.
- [x] Source limits recorded.
- [x] Current repo inspection recorded.
- [x] Owner contracts mapped.
- [x] Behavior preservation surfaces listed.
- [x] Characterization evidence identified.
- [x] Boundary guardrails planned.
- [x] Reviewable first slice named.
- [x] Docs and generation needs classified.
- [x] Validation commands proposed.
- [x] Current handoff included.
- [x] No generated-file hand edit planned.
- [x] No phase-shaped runtime dependency planned.
- [x] Production files remain out of scope.
- [x] Visual and accessibility evidence not promoted to product conformance.

## 15. Binary acceptance criteria

| ID | Criterion | Status |
| --- | --- | --- |
| RF-AC-001 | Exactly one primary target directory or seam is named. | PASS: `apps/web/src/workbook` |
| RF-AC-002 | All inspected source files are listed. | PASS: section 5 lists all 108 in-scope files. |
| RF-AC-003 | Relevant adjacent evidence gaps are closed or explicitly scoped. | PASS: section 1 and section 12 replace source-limit TODOs with inspected evidence and future-slice scope decisions. |
| RF-AC-004 | Every public contract surface that could drift is mapped. | PASS: section 6 maps HTTP, WebSocket, UI, selectors, generated artifacts, harness, visual, a11y. |
| RF-AC-005 | Behavior-preserving refactors are separated from behavior changes. | PASS: section 6 and section 10. |
| RF-AC-006 | Characterization needs are explicit. | PASS: sections 6, 10, and 11. |
| RF-AC-007 | Checkpoints include validation. | PASS: section 10. |
| RF-AC-008 | Frontend package boundaries are preserved. | PASS: sections 7, 8, 11, and 12. |
| RF-AC-009 | `/apps/web` does not learn grid vendor coordinate semantics. | PASS: plan preserves `@cartulary/grid-adapter` facade; no current production direct RDG import found in targeted scan. |
| RF-AC-010 | Generated files are not hand-edited. | PASS: generated roots are marked no-edit. |
| RF-AC-011 | Phase identity is not a runtime dependency. | PASS: phase-shaped tests remain harness evidence; production phase helper name marked for isolated cleanup. |
| RF-AC-012 | Another agent can resume without rediscovery. | PASS: sections 1, 5, 10, 11, and 13 provide restart context. |
| RF-AC-013 | Public routes, wire contracts, authorization behavior, workbook hot path, UI selectors, and harness accounting are frozen. | PASS: section 6 contract freeze matrix. |
| RF-AC-014 | Product conformance evidence is separated from design/readiness/support evidence. | PASS: section 11 evidence classes. |

## 16. Remediation plan for specification and implementation gaps

This section implements the durable remediation operating plan for future workbook work. It is a process, documentation, and handoff change: no runtime behavior, generated files, package manifests, contracts, migrations, or public wire surfaces are changed by this section. Future implementation slices MUST use the checklists below before editing code.

### Remediation matrix

| Gap | Remediation | Area | Rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Backend route/store evidence missing | Convert the backend owner map into a required contract-impact checklist for workbook changes. Behavior-changing frontend slices must name the owning backend module, route/store tests, auth/storage implications, and whether backend validation is required. | Documentation, tests, implementation process | A static inventory is useful but insufficient; each future change needs a repeatable gate. | Keeps frontend refactors aligned with backend source of truth as phases expand. | No runtime migration. Process update only unless a slice changes behavior. | Frontend code can silently encode wrong route, auth, or storage assumptions. | Slice handoff includes the checklist; behavior-changing slices run the narrow backend or service-backed Make target named by the checklist. |
| Generated protocol internals evidence gap | Strengthen facade-only policy: app code may import generated-derived surfaces only through approved package facades; generated internals remain read-only. Keep import-boundary coverage for generated-internal imports. | Documentation, tests, package boundary | Generated internals are unstable outputs and should not become app dependencies. | Enables regeneration and contract evolution without app churn. | No compatibility impact unless illegal imports exist; those migrate to facades. | Hand edits or direct imports create drift and brittle upgrades. | `make generated-artifact-policy-check`; `make frontend-import-boundary-check`; no diffs under generated roots. |
| E2E coverage outside targeted workbook specs incomplete | Replace filename-based E2E selection with the behavior-to-target evidence map below for shell/startup, grid/create, generic/mentions, evidence, collaboration/session, history, saved views/query, inspector/keyboard, restore, visual, and a11y readiness. | Documentation, tests | Coverage is distributed by behavior and phase; future agents need target selection by behavior. | Narrower, more reliable validation for future slices. | No product migration; some future changes may require broader browser targets. | Browser-impacting changes may be under-validated. | Handoff names the behavior family and Make target; browser-impacting slices include the mapped target. |
| Frontend phase maps and retained artifacts incomplete | Treat phase maps and retained artifacts as harness accounting only. Runtime names must stay behavior-based, and every retained-artifact claim must name an exact current-session run root. | Documentation, tests, harness process | Phase identity is evidence metadata, not architecture. | Prevents production phase coupling and stale-evidence overclaiming. | No runtime compatibility impact. | Runtime modules may mirror phase history; handoffs may cite stale artifacts. | Phase/harness changes run `make phase-map-check`, `make phase-ledger-drift`, or the relevant harness target; handoffs list exact current run roots. |
| Contracts/db/backend source evidence gap | Establish spec-first rule for behavior changes: owner spec update, derived contract or migration input update, generation/drift check, backend/frontend implementation, then validation. | Specification, contracts, implementation, tests | Workbook behavior is downstream of specs, contracts, migrations, and query inputs. | Keeps public behavior, storage, and generated code coherent. | Behavior changes may require migration, contract generation, and client updates. | Implementation-first changes can diverge from public contracts or storage semantics. | Owner spec diff exists before implementation; generated/drift targets pass; affected backend/frontend tests pass. |
| Large `WorkbookShell.tsx` controller | Preserve the completed shell split around durable seams and continue extracting only when a new seam owns a real design decision. New shell work must keep `WorkbookShell.tsx` as coordinator, not mutation/state owner. | Implementation, tests | The shell should coordinate surfaces instead of owning every surface workflow. | Higher cohesion and easier future surface addition. | Internal TypeScript movement only unless owner specs require behavior change. | Shell becomes a fragile dependency magnet as phases grow. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`; browser target only for route/startup/browser-flow changes. |
| Large `TimelineWorkbook.tsx` hot-path controller | Continue splitting by behavior: pending saves, inspector/history, collaboration/live updates, grid continuity, and mutation submission. Each move must be characterized before extraction. | Implementation, tests | Timeline is the highest-risk hot path and needs behavior-owned seams. | Smaller modules and clearer ownership of pending, collaboration, inspector, and grid behavior. | Internal refactor unless selectors/routes/wire behavior change. | Hot-path behavior remains hard to reason about and unsafe to extend. | Existing unit characterization plus mapped browser targets when persistence, route flow, focus/scroll, or browser behavior changes. |
| Phase-shaped production helper name | Prohibit new phase-shaped production modules. Keep phase names only in tests, maps, ledgers, and harness accounting; behavior names are required for runtime modules. | Implementation, tests, documentation | Runtime architecture should describe behavior, not historical implementation phases. | Prevents future phase coupling. | Import-only migration when helper bodies are unchanged. | New production code may follow phase naming and harden bad boundaries. | Source scan shows no phase-shaped production helper imports; `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`. |
| Literal production `data-testid` values | Classify selectors as shared contract selectors or private local IDs. Promote only shared runtime/test/browser selectors to `@cartulary/ui-contracts`; keep private IDs local. | Implementation, tests, package seam | Centralizing every literal bloats the API; shared selectors still need one owner. | Stable test/browser contracts with minimal selector surface. | Selector strings remain stable unless owner docs explicitly authorize change. | Runtime and tests drift, or selector package becomes noisy and brittle. | Affected tests pass; `@cartulary/ui-contracts` tests cover promoted selectors; import-boundary check passes. |
| View contract field access in app models | Keep app-specific state in app models. Move only repeated generated-contract adaptation, parsing, or normalization into `@cartulary/view-contracts` when duplication or generated leakage proves the need. | Implementation, package seam, tests | Shared packages should own stable contract adaptation, not app workflow state. | Cohesive package APIs and lower coupling to generated internals. | Future package API additions may require app import updates. | Either app models duplicate parsing or shared package scope expands without discipline. | Duplication/leak trigger is documented; package and app tests pass after any move. |
| Visual/a11y evidence overclaim risk | Keep visual and accessibility as design/readiness evidence unless Core 05 claim-publication criteria are explicitly active. Never use readiness rows as Base Profile conformance proof. | Specification, documentation, tests | Readiness and product conformance have different owners and failure semantics. | Honest evidence accounting and safer release claims. | No runtime migration. | Future handoffs overclaim visual/a11y readiness as product proof. | Handoff labels evidence class; visual/a11y targets run only for layout, focus, accessible-name, keyboard, or readiness changes. |

### Required contract-impact checklist

Every future workbook slice MUST copy and complete this checklist before implementation. If all entries are `none`, the slice may remain frontend-only.

| Field | Required value |
| --- | --- |
| Behavior family | One of the behavior families in the evidence map below. |
| Public route impact | Exact route(s), method(s), and owner module, or `none`. |
| WebSocket impact | Exact event family and owner, or `none`. |
| Storage/projection impact | Store, query, migration, or view-schema owner, or `none`. |
| Authorization/security impact | Core 04 owner section or backend auth primitive, or `none`. |
| Generated-surface impact | Owner input, generator, and drift target, or `none`. |
| Backend validation | Make target(s) required, or `not required: frontend-only behavior-preserving move`. |
| Browser validation | Make target from the evidence map, or `not required` with reason. |
| Visual/a11y classification | `not touched`, `readiness evidence`, or `Core 05 publication boundary active`. |
| Rollback scope | Smallest revertable file set and expected behavior after rollback. |

### Behavior-to-validation evidence map

Use this map instead of choosing browser evidence by filename. When a slice changes multiple families, run the union of the required targets.

| Behavior family | Typical owned behavior | Default validation | Browser/ready validation trigger |
| --- | --- | --- | --- |
| Shell/startup | Workbook route entry, initial sheet selection, saved-view startup | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check` | `make browser-e2e-webserver-backed` when public startup, route sequencing, or saved-view browser flow changes |
| Grid/create | Timeline grid rows, create, paste, row identity, focus anchors | default frontend gates | `make browser-e2e-stateful` when persistence, focus/scroll, paste, or browser-only grid behavior changes |
| Generic/mentions | Generic surfaces, mention resolution, entity creation, relationship chips | default frontend gates | `make browser-e2e-webserver-backed` when public generic route flow or mention browser flow changes |
| Evidence | Evidence attach, preview, download handle invocation | default frontend gates | `make browser-e2e-webserver-backed` when evidence handle route flow changes |
| Collaboration/session | Presence, WebSocket lifecycle, auth pause/recover, live row updates | default frontend gates | `make browser-e2e-stateful` when socket/session or persistence interaction changes |
| History/restore | Inspector history, rollback, restore, deleted-row subjects | default frontend gates | `make browser-e2e-stateful` when row-history route flow or browser continuity changes |
| Saved views/query | Saved-view selection, query JSON, layout JSON, filters | default frontend gates | `make browser-e2e-webserver-backed` when saved-view/query browser flow changes |
| Inspector/keyboard | Default-closed inspector, keyboard shortcuts, focus movement | default frontend gates | `make browser-e2e-a11y` or `make browser-e2e-a11y-preflight` when focus, keyboard, labels, or accessible names change |
| Restore/native surfaces | Native surface read/write affordances and restore workflows | default frontend gates | `make browser-e2e-stateful` when restore or native-surface browser flow changes |
| Visual readiness | Layout, density, shell/grid/inspector visual state | default frontend gates | `make browser-e2e-visual`; do not refresh goldens without explicit authorization |

### Generated and package-boundary rules

- Production app code MUST use package facades: `@cartulary/protocol-ts`, `@cartulary/ui-contracts`, `@cartulary/view-contracts`, and `@cartulary/grid-adapter`.
- Production app code MUST NOT import from `packages/*/src/generated/**`, `internal/gen/**`, generated design tokens directly, browser test helpers, or `react-data-grid`.
- Move contract parsing or normalization to `@cartulary/view-contracts` only after two app seams duplicate the same generated-contract adaptation or a generated internal leak is observed.
- Promote a selector to `@cartulary/ui-contracts` only when runtime and tests/browser helpers share the selector as a stable contract. Private component-only IDs remain local.

### Spec-first behavior-change workflow

1. Identify the owner section in Core 00 through Core 04, adopted subsystem NLSpecs, or testing-harness NLSpec.
2. If the desired future behavior differs from existing owner text, update the owner spec before implementation.
3. Update derived contract, view-schema, OpenAPI, WebSocket, migration, query, or generator inputs when the owner spec requires it.
4. Run the narrow drift/generation target before implementation validation.
5. Implement the behavior behind the smallest module seam.
6. Validate with the checklist targets and record exact run roots.

If owner docs conflict or a generated artifact appears stale, mark the slice `BLOCKED` with evidence and do not implement around the conflict.

### Frontend structural continuation backlog

| Backlog item | Depends on | Allowed change | Required preservation | Required validation | Stop condition |
| --- | --- | --- | --- | --- | --- |
| Collaboration/live-update seam | Completed pending-save and socket characterization | Move socket lifecycle orchestration and live update application behind behavior-named timeline collaboration hooks/services | WebSocket URL, message families, resume/reset handling, self-origin filtering, conflict interaction, save-state behavior | default frontend gates; browser stateful only if socket/session browser behavior changes | Stop before changing event payloads, refresh rules, auth behavior, or pending-queue semantics |
| Grid continuity seam | Current grid interaction refs and focus-anchor behavior | Extract focus/scroll/keyboard continuity helpers that hide grid implementation details | selector strings, row/field anchoring, scroll preservation, keyboard behavior, grid-adapter facade | default frontend gates; a11y when keyboard/focus semantics change | Stop before changing visible layout, accessible names, or grid vendor boundary |
| Mutation submission seam | Existing pending queue and generic surface seams | Separate mutation request construction and save-queue sequencing from presentation | request bodies, client transaction IDs, conflict anchors, retry/halt behavior | default frontend gates; backend/service-backed target if route semantics change | Stop before changing HTTP methods, envelopes, authorization assumptions, or storage semantics |
| Selector facade continuation | S-06 inventory | Promote only newly shared cross-boundary selectors to `@cartulary/ui-contracts` | exact selector strings and browser helper behavior | package test, default frontend gates | Stop when conversion would touch broad visual/a11y/browser helper choreography without a selector-specific need |

### Remediation workstreams

| Workstream | Dependencies | Sequencing | Risks | Exit criteria |
| --- | --- | --- | --- | --- |
| Evidence and authority baseline | Current tracker and live repo inspection | Refresh branch, dirty tree, owner docs, generated policy, import boundaries, backend owners, contracts, migrations, and E2E map before each new slice | Overstating inspection depth or relying on stale retained roots | Every gap has a named owner, evidence class, and validation target. |
| Spec and contract cleanup | Evidence baseline | Future behavior changes start in owner specs, then contracts, generation, implementation, and tests | Starting implementation before owner contract change | Owner diffs and drift/generation targets exist for behavior changes, or the slice is blocked. |
| Frontend structural remediation | Evidence baseline and contract checklist | One behavior seam per slice: shell coordination, timeline collaboration, grid continuity, mutation submission, or selector facade | Large slices or hot-path movement without characterization | Each slice is small, reviewable, independently revertible, and validated before the next slice. |
| Evidence and harness remediation | Evidence map and contract checklist | Maintain behavior-to-browser target mapping; keep phase maps as harness accounting only | Overclaiming design/readiness or stale retained evidence | Handoff names exact run roots and skipped checks with reasons. |
| Final handoff | Completed or blocked slice | Run `make agent-finalize`; record changed files, commands, roots, failures, skipped retained-run maintenance, generated status, rollback, and safe restart command | Leaving the next agent without restart context | No generated-file policy violation, no out-of-scope diffs, and all gaps are `DONE`, `DEFERRED` with reason, or `BLOCKED` with evidence. |
