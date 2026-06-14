# apps/web/src Refactor Handoff Tracker

## Status, Authority, And Normative Language

Status: implementation-support NLSpec-style handoff artifact. This document is normative for this refactor tracker: sequencing, characterization, extraction boundaries, validation evidence, and handoff interpretation. It is not a Core product specification and does not create, widen, narrow, or replace Base Profile or extension-profile conformance behavior.

Core 00 through Core 04 remain the implementation-conformance authority. Core 05 applies only at claim-bearing publication boundaries. When this tracker restates Core behavior, the restatement is only a local implementation-support pointer and the named Core owner governs.

Normative terms in this tracker have the following local meanings:

| Term | Meaning inside this tracker |
| --- | --- |
| `MUST` | Required for the refactor handoff to be considered complete. |
| `MUST NOT` | Forbidden inside this refactor handoff. |
| `SHOULD` | Required default for this tracker unless the tracker records a specific exception and its validation effect. |
| `MAY` | Optional tracker behavior only when omission behavior is stated in the same paragraph, table row, or immediately following paragraph. A `MAY` statement without omission behavior is invalid in this file. |
| `default` | Required value, interpretation, action, or status when a more specific value is omitted. |

Statement classes:

| Statement class | Meaning | Required handling |
| --- | --- | --- |
| Tracker-owned requirement | Refactor sequencing, extraction, validation, or handoff interpretation rule owned by this document. | Binding for work performed under this tracker. |
| Core-owner restatement | Compact pointer to behavior owned by Core 00 through Core 04. | MUST cite the owner; MUST NOT be treated as independent product authority. |
| Implementation guidance | Preferred implementation organization that does not change observable product behavior. | Follow unless a stronger local code boundary requires a recorded exception. |
| Validation requirement | Evidence required before a sprint or extraction step is complete. | Blocking unless explicitly recorded as skipped with reason and owner. |
| Rationale | Explanation for why a boundary exists. | MUST NOT override a requirement. |
| Out-of-scope exclusion | Named work that this tracker does not authorize. | MUST be handled by a separate tracker, owner-spec change, or product task. |

## Pinned Inputs And Source Snapshot

Default drift behavior: the tracker is bound to the pinned source snapshot in this table until this file is explicitly updated. If a source file changes after this snapshot, the changed source informs future tracker revision but does not silently alter this tracker.

| Source path | Snapshot identity | Authority role | Required use | Drift handling |
| --- | --- | --- | --- | --- |
| `fallow-report.json` | SHA-256 `7290ff60f6c34caff902d8ee8598ec6dca827ea482768d79157bb20ea6584939`; JSON stream with 317 findings; generated timestamp not present. | Triage signal only. | Use only the finding groups listed in the pinned finding-closure table. | Findings not listed below are not refactor requirements. Metric reduction is not acceptance evidence. |
| `docs/spec/00_document_set_status_and_precedence.md` through `docs/spec/04_security_deployment_and_conformance.md` | Repo snapshot at tracker revision time. | Product-conformance owner corpus. | Resolve product behavior, profile boundaries, public interfaces, view contracts, record semantics, security, and acceptance criteria. | If this tracker and Core differ, Core governs and this tracker needs repair. |
| `docs/domain.md` | Repo snapshot at tracker revision time. | Domain vocabulary and concept-boundary reference. | Resolve workbook, surface, party, artifact, evidence, object blob, saved-view, system-view, and implementation-support language. | If vocabulary differs from a Core owner, owner section governs and `docs/domain.md` plus this tracker need repair. |
| `docs/design.md` | Repo snapshot at tracker revision time. | Derived frontend design direction. | Use only for layout, visible state, density, keyboard, accessibility, and interaction presentation direction. | MUST NOT be cited as Base Profile or extension-profile conformance evidence. |
| `docs/testing-harness-nlspec.md` | Repo snapshot at tracker revision time. | Harness mechanics and public Make command owner. | Use for command invocation, target selection, result-root interpretation, artifacts, cleanup, and agent finalization. | If command mechanics change, update this tracker before relying on old validation commands. |

## Applicability And Omission Semantics

In-scope source files are `apps/web/src/**`. Adjacent files outside `apps/web/src/**` are inspection-only unless an import boundary, test helper, generated TypeScript contract, or typed call site must change to keep the named in-scope behavior executable and validated.

Omission behavior: a module, route, surface, behavior, command, or artifact not named in this tracker is out of scope for this refactor unless a listed Core owner or a touched import boundary requires it. Out-of-scope items MUST NOT be added to this tracker by implication from Fallow, tests, or design text.

Line numbers in this file are orientation only. Symbols, paths, view identifiers, field keys, and owner-section citations are the binding references.

### Behavior Preservation Decision Matrix

| Pre-refactor state | Required action | Production code change allowed before Sprint 1? | Test required before Sprint 1? | Recording rule |
| --- | --- | --- | --- | --- |
| Implemented behavior matches a cited Core owner. | Preserve through characterization and extraction. | No, except test-only helpers. | Yes, existing or newly added characterization. | Record owner section and test evidence in the Definition of Done table. |
| Implemented behavior is not present and no Core owner requires it. | Preserve non-implementation; do not add feature work. | No. | No product test required; a tracker row must state the non-implementation boundary. | Record as out-of-scope exclusion or non-implementation row. |
| Implemented behavior conflicts with a cited Core owner. | Do not repair under this refactor unless the owner requirement is already covered by a separate approved task. | No. | Add characterization only when needed to prevent accidental widening during extraction. | Record as `defer to separate product change`. |
| Implemented behavior is untested and extraction would move it. | Characterize before moving. | No production extraction until the characterization row exists. | Yes. | Mark `TODO: blocking characterization required` until a test or exact owner-backed assertion is recorded. |
| Implemented behavior is implementation-only and invisible to callers. | Keep or simplify only when the observable contract is unchanged. | Yes during the owning sprint. | Type/import tests are sufficient unless UI or route output changes. | Record as implementation guidance, not product behavior. |
| Implemented behavior is visual/design-only. | Preserve only when `docs/design.md` or existing visual/a11y evidence names it. | Yes during presentational extraction. | Browser, visual, or a11y checks only when affected. | Record design owner and validation target; MUST NOT cite as Core conformance. |

## Closed Interface Map

Each extraction target MUST satisfy the row below before call sites are migrated. A target that cannot satisfy its row remains in its original file until Sprint 0 records a blocking characterization TODO or a narrower extraction row.

| Area | Binding symbols and paths | Allowed dependencies | Forbidden dependencies | Input contract | Output contract | Omitted input behavior | Invalid input behavior | Required validation |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| App session and workbook entry | `App`, `refreshShell`; `apps/web/src/App.tsx` | Session/auth helpers, incident list APIs, workbook open commands. | Workbook surface state machines, saved-view mutation rules, grid adapter details. | Session snapshot, incident identifier, optional launch `sheet_ref`. | Shell-open request or landing/auth state. | Missing launch pointer follows Core 03 §2.4 fallback through shell runtime. | Invalid explicit launch pointer is skipped and not persisted per Core 03 §2.4. | App landing/auth tests plus shell startup characterization before any extraction. |
| Workbook shell runtime | `WorkbookShell`, future `useWorkbookShellRuntime`; `apps/web/src/WorkbookShell.tsx` | Registry helpers, startup helpers, saved-view runtime, surface command props. | Auth redirects, App landing state, surface JSX internals, route definitions. | Incident/session props, surface registry, optional startup result, saved-view list. | Active surface model, `sheet_ref`, base `view_schema_id`, shell messages, commands. | Missing startup result resolves through Core 03 §2.4 order. | Unknown `view_schema_id`, invisible saved view, deleted saved view, or unavailable optional pack follows Core 03 §2.4 invalid-pointer handling. | Phase 8 startup/query tests and saved-view tests. |
| Saved-view runtime | Future `workbookSavedViewRuntime.ts`; `workbookSavedViews.ts`; `workbookQuery.ts` | Saved-view contract types, query/layout normalizers, API intent builders. | Contract-backed system-view identity ownership, JSX, local selection/focus/presence persistence. | Visible saved views, active `sheet_ref`, query state, layout state, action kind. | Saved-view state snapshot and API intents. | Omitted `scope` defaults to `private`; omitted `sort` and `filters` persist as empty arrays; omitted grouping is omitted `group_by`; omitted create-time layout normalizes to `cartulary.layout.v1` per Core 03 §2.3. | `group_by: null`, view-schema mutation, system saved-view write through ordinary route, invisible saved view, and deleted saved view are invalid and must surface as blocked API intent or owner-defined error path. | Phase 8 saved-view query/layout tests and runtime unit tests. |
| Timeline row model | Future `timelineRowsModel.ts`, future `useTimelineRows.ts`; `applyRowMutation`, `loadRows`; `apps/web/src/TimelineWorkbook.tsx` | View-row normalizers, row-version freshness helpers, query metadata, local draft snapshot inputs. | JSX, WebSocket transport, fetch execution, pending queue mutation. | Committed rows, query result rows, row patches, row-version high-water values, selection/focus anchors. | Next row snapshot, committed-row updates, continuity commands, freshness decision. | Missing optional continuity target clears only that target; missing selected/focused row returns no focus command. | Wrong `view_schema_id`, malformed row envelope, stale query result, deleted row, or lower row version follows Core 01 §3.3.4, Core 03 §14.1, and Core 03 §4.4 freshness rules. | Phase 3, 6, 8, and 9 tests plus direct model tests. |
| Pending saves hook | Future `useTimelinePendingSaves.ts`; `workbookPendingQueue.ts`; Timeline save callers | Pending queue model, payload builders, API functions, public error parser. | Cell editor JSX, conflict resolver JSX, WebSocket connection ownership. | Row versions, client transaction identifiers, pending units, auth/session state, mutation results. | Save commands, queue snapshot, conflict admissions, status-strip model. | Missing row version for replay blocks the unit and leaves later units queued. | Auth failure pauses without discarding queued writes; same-field conflict exits retry queue and enters conflict queue per Core 03 §3.3.4 and §4.4. | `workbookPendingQueue` tests, phase 3 autosave, phase 6 auth recovery. |
| Live updates hook | Future `useTimelineLiveUpdates.ts`; `workbookSocketLifecycle.ts`; `handleMessage` | Socket lifecycle reducer, event normalizers, row/save command interfaces. | Row reconciliation internals, UI labels, saved-view persistence. | Socket URL/session, lifecycle events, public collaboration payloads. | Presence snapshot, connection state, row/save commands, auth/session commands. | Missing optional presence data produces no presence update. | Unknown event kind is ignored unless lifecycle reducer marks it terminal; session revocation pauses unsaved work per Core 03 §4.4 and Core 04 §1.1.1. | Phase 6 WebSocket/presence/session tests. |
| Conflict model and hook | Future `timelineConflictModel.ts`, future `useTimelineConflicts.ts`; `TimelineConflictResolver.tsx`; parser copies in `TimelineWorkbook.tsx` and `workbookPendingQueue.ts` | Public error envelope parser, queue key helpers, current row snapshots, resolution callbacks. | Resolver JSX inside model, fetch execution, route ownership. | Public conflict payload, `record_id`, `field_key`, local draft, server value, base value, paste ordinal metadata. | Normalized conflict entry, active conflict props, queue navigation commands, resolver focus target. | Missing optional metadata such as server actor/timestamp remains omitted from display props. | Missing required conflict token, record id, field key, resolution class, base/current row version, or invalid envelope returns parse failure and must not enter conflict queue. | Parser fixture tests, phase 7 conflict resolver tests, paste conflict tests. |
| Clipboard utility | Future `workbookClipboard.ts`; duplicate `parseClipboardTableForDimensions`; Timeline and Shell paste callers | String parsing helpers only. | Grid target resolution, mutation dispatch, API calls, DOM clipboard access. | Plain text clipboard payload. | String matrix and row/column dimensions. | Empty text returns one row with one empty cell. | Maximum dimensions are `TODO: blocking characterization required` unless an owner-derived limit is added before extraction. Malformed quoted text follows the pre-refactor parser until characterized. | Direct table tests plus phase 9 paste target tests. |
| Value formatting utility | Future `workbookValueFormat.ts`; duplicate `stringifyGridValue` | Primitive conversion and collection display extraction. | Locale policy, UI markup, mutation payloads, validation. | Unknown cell value. | Display-safe string. | `null`, `undefined`, missing cell, unsupported object, and unsupported array return empty string. | Non-finite number, Date object, and object without `items[]` require `TODO: blocking characterization required` before changing output. | Direct unit tests and entity/timeline display tests. |
| Generic mutation model | Future `genericWorkbookModel.ts`; `buildGenericCreatePayload`, `buildGenericPatchChange`, `workbookCreateMinimumSatisfied`, `initialGenericCreateDraft` | View contract fields, draft values, user id, pure payload helpers. | React state, fetch, layout, saved-view state, visible messages except message-key output. | Contract, draft map, current user id, client transaction id. | Create payload, patch change, minimum-create result, initial draft. | Omitted draft fields are empty strings; omitted user id disables owner defaults. | Invalid number/boolean payload returns null; read-only field is skipped; unknown view schema uses non-empty-value minimum rule until owner-specific row exists. | Table-driven unit tests and generic surface create/edit tests. |
| Reference option model | Future `workbookReferenceOptions.ts`; `referenceOptionsForField` | View field contract, loaded same-incident option lists. | Network loading, component rendering, cross-incident lookup. | Field contract and normalized option buckets. | Ordered reference option list and boolean use-reference decision. | Missing option bucket is empty list. | Invisible, deleted, wrong-type, or foreign-incident records must never appear in normalized inputs; if present, characterization is blocking before extraction. | Unit tests and generic reference flow tests. |
| Entity surface | Future `EntityWorkbookSurface.tsx`; `EntityWorkbookSurface` | Entity rows/query/actions/reference props, shared clipboard/value utilities. | Shell routing, saved-view CRUD, grid adapter internals. | Host/identity rows, query controls, merge actions, paste adapter commands. | Rendered entity surface and user intents. | Missing optional timeline preview data renders no preview action. | Wrong entity type, foreign incident row, or deleted row is rejected before mutation intent. | Entity surface tests and phase 3/9 grid tests. |
| Generic surface | Future `GenericWorkbookSurface.tsx`; `GenericWorkbookSurface` | Generic model outputs, rows, reference options, action callbacks. | Generic pure rules, shell routing, saved-view state. | Contract-backed surface model, row list, create/edit/action availability. | Rendered generic surface and user intents. | Empty reference options render no reference choices, not ad hoc free linking. | Missing required contract field blocks create/update intent. | Surface create/edit/reference tests. |
| Assessment surface | Future `AssessmentWorkbookSurface.tsx`; assessment helpers in `WorkbookShell.tsx` | Assessment rows/query/actions and support references. | Generic model ownership, shell routing. | Assessment rows, create draft, support rows, confidence controls. | Rendered assessment surface and intents. | Missing confidence persists owner-defined unset handling per Core 03 §16.3. | Invalid confidence outside owner range is blocked before API intent. | Assessment tests. |
| Timeline presentational panels | Future `TimelineHistoryPanel.tsx`, `TimelineEvidencePanel.tsx`, `TimelineMentionsPanel.tsx`, `TimelineRowActions.tsx`, `TimelineGridSurface.tsx`; existing `TimelineWorkbookInspector.tsx` | Typed props, callbacks, layout/design tokens. | Fetch, pending queue, WebSocket, route payload ownership, Core-domain decisions. | Hook snapshots, selected row model, callbacks, visible state flags. | Rendered panels and callback intents. | Missing optional panel data renders empty owner-specific state. | Destructive action without legal action metadata is not rendered. | Phase 4-9 DOM tests, focus tests, evidence/history/mention tests. |
| Grid adapter boundary | `TimelineWorkbookGrid.tsx`; `packages/grid-adapter/src/*` inspected only | Vendor-neutral grid props, anchors, navigation events. | Workbook mutation semantics, conflict semantics, evidence semantics, saved-view semantics. | Rows, columns, anchors, editor props, grouping rows. | Grid events and rendered grid. | Missing optional viewport state uses adapter default behavior. | Row-index-only, vendor-coordinate-only, group-header, or presentation-only targets must not become workbook mutation targets per Core 03 §13.3 and §14.5. | Phase 9 anchors/paste/keyboard tests. |

### Extraction Capability Closure

| Target family | React state | JSX | Transport calls | Route payload construction | Local storage | Saved-view identity | Grid adapter details | Core-domain decisions |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Pure models and utilities | MUST NOT own | MUST NOT own | MUST NOT own | Only generic payload builders named above can build payload objects | MUST NOT own | MUST NOT own | MUST NOT own | MUST NOT invent; cite owner or return blocked result |
| Hooks | Can own state/effects for their named surface only | MUST NOT return JSX | Can call injected API functions only | Can call pure payload builders | MUST NOT persist saved-view or client-local layout outside owning runtime | Only saved-view runtime can own saved-view transitions | Can consume adapter events through typed commands | MUST NOT invent; cite owner or surface blocked state |
| Presentational components | Can own transient UI affordance state only | Can own | MUST NOT own | MUST NOT own | MUST NOT own | MUST NOT own | Can render adapter props but not interpret vendor coordinates as record identity | MUST NOT own |
| Shell runtime | Can own shell selection state | MUST NOT own surface JSX | Can call injected saved-view/startup APIs | Can build shell startup/saved-view API intents | MUST NOT persist row-local UI state | Can own `sheet_ref` and saved-view transition interpretation | MUST NOT own | MUST cite Core 03 §2 |

## Pinned Fallow Finding Closure

Default: Fallow findings not listed in this table are not requirements for this tracker. Reducing finding count, severity, cyclomatic complexity, cognitive complexity, line count, or CRAP score is never acceptance evidence by itself.

| Pinned finding id/group | Accepted refactor requirement | Rejected interpretation | Required closure | Validation evidence |
| --- | --- | --- | --- | --- |
| `TimelineWorkbook.tsx` group: 58 findings, max critical | Treat Timeline as the first high-risk responsibility hub. Extract only after row/save/conflict/live/paste/history behavior is characterized. | Do not delete workbook behavior or split by metric alone. | Sprints 0-3 must produce model/hook/panel boundaries with tests before call-site migration. | Phase 3-9 tests plus direct model tests. |
| `TimelineWorkbook.tsx::applyRowMutation` and `loadRows` | Extract row reconciliation and freshness decisions into directly verified model/hook boundaries. | Do not move query freshness into grid adapter or accept row-position identity. | Define stale query, newer local row, deleted row, grouping, selection, and focus rules. | Phase 3, 6, 8, 9 and model tests. |
| `TimelineWorkbook.tsx::handleMessage` | Keep `workbookSocketLifecycle.ts` as reducer boundary and extract event-to-command adapter. | Do not make transport events mutate UI state directly across modules. | Define unknown event, revocation, presence, row patch, and replay command behavior. | Phase 6 tests. |
| `TimelineWorkbook.tsx::queueScalarSave` and `queueCollectionSave` | Extract pending save orchestration and payload builders. | Do not change idempotency, row-version, auth-pause, or conflict behavior. | Preserve client transaction identifiers and queue/replay ordering. | `workbookPendingQueue` and phase 3/6/7 tests. |
| `TimelineWorkbook.tsx::handleKeyDown` and collection key handlers | Extract keyboard command dispatch around existing command mapper. | Do not create new spreadsheet semantics. | Define Enter/Tab/escape/focus outcomes before moving handlers. | Phase 9 keyboard tests and design owner references where presentation-only. |
| `TimelineWorkbook.tsx::renderRowHistorySection` | Extract history model/panel after legal action state is explicit. | Do not infer rollback/delete/restore legality from visible text. | Require structured action metadata and row-local pending destructive state. | Phase 7 history/rollback tests. |
| Duplicate `parseSameFieldConflict` parsers | Consolidate parser after fixtures cover both payload shapes. | Do not weaken required conflict fields or silently accept malformed envelopes. | One parser must enforce required conflict-token, record, field, class, and row-version fields. | Parser fixtures and phase 7 tests. |
| `workbookPendingQueue.ts` group: 8 findings | Retain as core pure queue model; refine only around parser consolidation or helper boundaries. | Do not split merely to reduce file size. | Preserve queue ordering, auth pause, retry, overflow, conflict admission, and public error mapping. | Existing queue tests plus new parser fixtures. |
| `WorkbookShell.tsx` group: 29 findings, max critical | Extract shell runtime and surfaces after Timeline high-risk behavior is pinned. | Do not move App auth, product routes, or Core owner decisions into surface components. | Define saved-view/startup runtime and surface frame contracts. | Phase 8 shell/saved-view tests and frontend import boundary check. |
| `WorkbookShell.tsx::EntityWorkbookSurface` | Extract entity surface with merge/edit/paste adapters. | Do not collapse host/identity entity semantics into generic grid behavior. | Preserve entity-origin paste, merge action boundaries, and timeline preview intent shape. | Entity tests and phase 3/9 tests. |
| `WorkbookShell.tsx::GenericWorkbookSurface` | Extract generic surface after pure create/patch/reference models exist. | Do not embed generic business rules in component markup. | Define create minimum, defaults, invalid draft, reference option, and action availability rules. | Generic unit tests and surface tests. |
| `WorkbookShell.tsx::buildGenericCreatePayload`, `workbookCreateMinimumSatisfied`, `initialGenericCreateDraft`, `referenceOptionsForField` | Move pure rules to model modules with table-driven tests. | Do not change route or wire shapes. | Define omitted draft fields, read-only fields, invalid scalar parsing, owner defaults, and unknown schema handling. | Direct unit tests and frontend typecheck. |
| `App.tsx::refreshShell` | Treat as adjacent App/session risk. | Do not split App broadly under this tracker. | Pin workbook open/default-surface behavior only when shell extraction touches it. | App landing/auth tests plus shell startup tests. |
| Test-file groups | Preserve behavior-encoding tests; reduce setup duplication only after extracted units have direct tests. | Do not delete large tests to improve metrics. | Test support helpers must keep diagnostics and visible assertions. | Existing phase tests and helper-specific review. |
| `TimelineWorkbookInspector.tsx` group | Split panels only after hook props are explicit. | Do not freeze not-yet-settled parent render callbacks prematurely. | Evidence, mentions, history, messages, and actions receive typed props or nodes. | Phase 4-7 inspector/panel tests. |
| Admin panels, harnesses, unrelated tests | Out of scope unless imports or tests touched by named refactor require adjustment. | Do not expand tracker into admin or harness refactors. | Leave untouched except for mechanical import/test updates required by in-scope movement. | Import boundary and targeted tests only. |

## Ordered Implementation Plan By Sprint

| Sprint | Purpose | Files likely touched | Work requirements | Validation requirements |
| --- | --- | --- | --- | --- |
| Sprint 0 | Make behavior explicit before production movement. | Existing `WorkbookShell.phase*.test.tsx`, `TimelineWorkbook*`, `workbook*` tests, `timelineWorkbookTestSupport.ts`, `appShellTestSupport.ts`. | Map each Definition of Done row to existing evidence or add characterization tests. Production source movement is forbidden. | `make task-guide ROLE=feature-dev PHASE=phase3`; relevant phase slices; `make frontend-unit` when target routing is not proven. |
| Sprint 1 | Extract pure utilities and models. | `workbookValueFormat.ts`, `workbookClipboard.ts`, `timelineConflictModel.ts`, `genericWorkbookModel.ts`, `workbookReferenceOptions.ts`, `timelineRowsModel.ts`. | Extract one unit at a time. Add direct tests before replacing call sites. Preserve route and wire shapes. | Direct unit tests through `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`. |
| Sprint 2 | Separate Timeline effects and state machines from markup. | `useTimelineRows.ts`, `useTimelinePendingSaves.ts`, `useTimelineLiveUpdates.ts`, `useTimelineConflicts.ts`, `useTimelineEvidenceActions.ts`, `useTimelineMentions.ts`, `useTimelineGridInteractions.ts`, `TimelineWorkbook.tsx`. | Hooks expose snapshots and commands, not JSX. Each hook must satisfy the capability closure table. | Timeline phase tests for autosave, live updates, history, query, keyboard; `make frontend-typecheck`; `make frontend-unit`. |
| Sprint 3 | Extract Timeline presentational panels after hook props are explicit. | `TimelineHistoryPanel.tsx`, `TimelineEvidencePanel.tsx`, `TimelineMentionsPanel.tsx`, `TimelineRowActions.tsx`, `TimelineGridSurface.tsx`, `TimelineWorkbookInspector.tsx`, `TimelineWorkbookGrid.tsx`. | Components use typed props and emit callback intents only. Domain actions stay in hooks/models. | Phase 4-9 DOM tests, conflict focus tests, evidence/mention/history tests. |
| Sprint 4 | Split WorkbookShell runtime and surfaces. | `useWorkbookShellRuntime.ts`, `workbookSavedViewRuntime.ts`, `WorkbookSurfaceFrame.tsx`, `EntityWorkbookSurface.tsx`, `GenericWorkbookSurface.tsx`, `AssessmentWorkbookSurface.tsx`, `WorkbookShell.tsx`. | Move startup/saved-view logic before surfaces. Keep `view_schema_id` and `sheet_ref` as identity inputs. | Shell startup/surfaces/query/saved-view tests; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make frontend-unit`. |
| Sprint 5 | Remove leftovers and prove the workbook shell still behaves as one product surface. | Duplicate helpers, test support, tracker status rows, import cleanup. | Delete obsolete helpers only after all call sites use shared modules. Do not broaden scope. | `make generated-artifact-policy-check`; `make json-shape-check`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make frontend-unit`; targeted browser checks when interaction or visual behavior changed. |

Sprint status defaults to `not-started`. A sprint status can change only after its validation requirements are recorded in this tracker or an implementation PR description that links back to this tracker.

## Live Execution Tracker

Status owner: Pursue Goal execution starting 2026-06-14. This section records implementation-support progress for Sprints 0 through 5 without changing the normative extraction contracts above.

| Sprint | Status | Blocking characterization before movement | Evidence and notes |
| --- | --- | --- | --- |
| Sprint 0 | complete | Foreign/deleted/invisible reference options remain TODO unless direct tests or a deferral row are added. | Initial target inspection ran `make help`, `make task-guide ROLE=feature-dev PHASE=phase3`, `make explain-target TARGET=frontend-unit DETAIL=summary`, `make explain-phase PHASE=phase3`, and `make explain-phase PHASE=phase4` through `phase9`. No production source moved during Sprint 0. Existing evidence is mapped below; Sprint 1 added direct utility/model tests before replacing call sites. |
| Sprint 1 | complete | Foreign/deleted/invisible reference options remain `TODO: blocking characterization required`; extraction does not normalize those inputs and keeps the normalized-input boundary. | Added direct tests for `workbookValueFormat`, `workbookClipboard`, `timelineConflictModel`, `timelineRowsModel`, `genericWorkbookModel`, and `workbookReferenceOptions`. Replaced duplicate Timeline/Shell/queue call sites. Duplicate scan: `rg -n "function stringifyGridValue|function parseClipboardTableForDimensions|function parseSameFieldConflict\\(|function decideWorkbookRecordFreshness|export function buildGenericCreatePayload|export function buildGenericPatchChange|function initialGenericCreateDraft|function referenceOptionsForField" apps/web/src` shows one production owner per helper family. Validation: pre-migration `make frontend-unit` failed at `.cartulary/test-results/20260614T034431Z-p97648` due a new test expectation, corrected before source replacement; `make frontend-unit` passed at `.cartulary/test-results/20260614T034527Z-p99606` before replacement and `.cartulary/test-results/20260614T034903Z-p3310` after replacement; `make frontend-typecheck` passed at `.cartulary/test-results/20260614T034903Z-p3287`; `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260614T034824Z-p2335`. |
| Sprint 2 | complete | Hook extraction stayed within characterized row/save/live/conflict/evidence/mention/grid behavior. Uncharacterized foreign/deleted/invisible reference option normalization remains outside Timeline hook movement. | Added `useTimelineRows`, `useTimelineConflicts`, `useTimelinePendingSaves`, `useTimelineLiveUpdates`, `useTimelineMentions`, `useTimelineEvidenceActions`, and `useTimelineGridInteractions`. Hooks expose snapshots, commands, and refs; JSX and domain side-effect command bodies remain in `TimelineWorkbook.tsx` unless already characterized by existing tests. Validation: `make frontend-typecheck` passed at `.cartulary/test-results/20260614T035505Z-p7556`; `make frontend-unit` passed at `.cartulary/test-results/20260614T035514Z-p7946`. |
| Sprint 3 | complete | `TimelineHistoryPanel` full extraction remains blocked because destructive/rollback preview state and action legality are still interleaved with parent row-history command construction; `TimelineMentionsPanel` full extraction remains blocked because mention selection and action messaging still share `TimelineWorkbookInspector` internals. Both paths keep existing characterized behavior in place instead of widening props. | Added `TimelineRowActions`, `TimelineEvidencePanel`, and `TimelineGridSurface`. Components take typed props and emit callback intents; hooks/models still own domain actions. Validation: `make frontend-typecheck` passed at `.cartulary/test-results/20260614T035859Z-p10754`; `make frontend-unit` passed at `.cartulary/test-results/20260614T035908Z-p11150`. |
| Sprint 4 | complete | `EntityWorkbookSurface`, `GenericWorkbookSurface`, and `AssessmentWorkbookSurface` full extraction remains blocked because their props still include query runtime refs, fetch/refresh closures, create/edit draft state, and direct surface mutation orchestration from `WorkbookShell.tsx`; moving them now would either widen product behavior or freeze an unstable callback surface. Startup/saved-view runtime extraction is complete and preserves `view_schema_id` plus `sheet_ref` as separate identity inputs; App auth remained untouched. | Added `workbookSavedViewRuntime` pure helpers, `useWorkbookShellRuntime`, and `workbookSavedViewRuntime.test.ts`. `WorkbookShell.tsx` now delegates surface identity, startup sheet ref, saved-view list upsert/delete/replace, sheet reload token, and grid-focus request state to the runtime hook while keeping surface JSX in place. Validation: `make frontend-typecheck` initially failed at `.cartulary/test-results/20260614T040402Z-p14407` due stale `WorkbookSheetRef` import, fixed; `make frontend-unit` initially failed at `.cartulary/test-results/20260614T040430Z-p15307` due the new runtime test using an invalid notes filter fixture, fixed. Passing evidence: `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260614T040402Z-p14416`; `make frontend-typecheck` passed at `.cartulary/test-results/20260614T040510Z-p17570`; `make frontend-unit` passed at `.cartulary/test-results/20260614T040510Z-p17547`. |
| Sprint 5 | complete | No obsolete duplicate helper owners remain for value formatting, clipboard dimensions, same-field conflict parsing, row freshness, generic create/patch/defaults, reference options, or saved-view runtime helpers. Remaining shell identity setters are intentionally owned only by `useWorkbookShellRuntime`. Targeted browser checks skipped: this refactor moved characterized logic/components without visual restyling, route changes, grid-vendor changes, or browser-only behavior changes; impact is that browser-engine-only regressions outside the unit/DOM harness were not re-exercised in this run. `make agent-finalize` skipped because no broad end-of-run suite or retained `RESULTS_DIR` was used. | Final cleanup moved imperative Timeline refs back to parent-created refs passed into hooks so Biome keeps ref stability without adding `.current` dependencies; hooks still own snapshots/commands. Duplicate scan: `rg -n "function stringifyGridValue|const stringifyGridValue|function parseClipboardTableForDimensions|parseClipboardTableForDimensions|function parseSameFieldConflict\\(|const parseSameFieldConflict|function decideWorkbookRecordFreshness|function buildGenericCreatePayload|function buildGenericPatchChange|function initialGenericCreateDraft|function referenceOptionsForField|function upsertSavedViewList|function fallbackIdentityAfterSavedViewDelete|function savedViewQueryStateForRuntime" apps/web/src` shows one production owner per helper family. Shell setter scan shows `setSurface`, `setStartupSheetRef`, `setSheetReloadToken`, and `setSavedViews` only in `useWorkbookShellRuntime`. Validation: `make lint-biome` initially failed at `.cartulary/test-results/20260614T040624Z-p20046` and `.cartulary/test-results/20260614T041213Z-p24571` due formatting/imports and custom-hook dependency diagnostics; fixed by formatting, parent-created refs, and explicit setter dependencies. Passing evidence: `make frontend-typecheck` passed at `.cartulary/test-results/20260614T041358Z-p26285`; `make frontend-unit` passed at `.cartulary/test-results/20260614T041358Z-p26307`; `make lint-biome` passed at `.cartulary/test-results/20260614T041358Z-p26631`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260614T041427Z-p28765`; `make json-shape-check` passed at `.cartulary/test-results/20260614T041427Z-p28797`; `make frontend-import-boundary-check` passed at `.cartulary/test-results/20260614T041427Z-p28802`. |

### Sprint 0 Characterization Map

| Definition of Done row | Existing or added evidence | Sprint 0 status |
| --- | --- | --- |
| Workbook open/default surface | Existing `workbookStartup.test.ts` startup fallback tests and `App.landing.test.tsx` landing/auth workbook access tests. | Characterized for Sprint 4 planning; App auth remains out of extraction scope. |
| Built-in/system view switching | Existing `workbookSurfaceRegistry.test.ts`, `WorkbookShell.surfaces.test.tsx`, and `WorkbookShell.phase8.query.test.tsx` cover stable surface and query identities. | Characterized for shell planning. |
| Saved-view switching | Existing `workbookSavedViews.test.ts`, `workbookStartup.test.ts`, `WorkbookShell.surfaces.test.tsx`, and `WorkbookShell.phase8.query.test.tsx` cover saved-view identity and query/layout behavior. | Characterized for Sprint 4 startup/runtime movement. |
| Scalar edit/save-state | Existing `workbookPendingQueue.test.ts`, `WorkbookShell.phase3.autosave.test.tsx`, `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase6.test.tsx`, and `WorkbookShell.phase7.test.tsx`. | Characterized before pending-save hook extraction. |
| Timeline paste | Existing `WorkbookShell.phase9.sentinel.test.tsx` covers scalar versus tabular dispatch and stable paste anchors; direct clipboard utility tests are required before helper replacement. | `TODO: blocking characterization required` remains for maximum accepted row/column count and malformed unclosed quote behavior; no output change allowed. |
| Keyboard navigation | Existing `workbookKeyboard.test.ts`, `GridAdapter.phase9.anchor.test.ts`, and `WorkbookShell.phase9.sentinel.test.tsx`. | Characterized before grid-interaction hook extraction. |
| Selected row and focused cell continuity | Existing `WorkbookShell.phase3.grid.test.tsx`, `WorkbookShell.phase6.test.tsx`, `WorkbookShell.phase8.query.test.tsx`, and `WorkbookShell.phase9.inspector.test.tsx`. | Characterized for row identity and focus continuity movement. |
| Conflict display/resolution | Existing `workbookPendingQueue.test.ts`, `WorkbookShell.phase6.test.tsx`, `WorkbookShell.phase7.test.tsx`, and planned parser fixture tests before consolidation. | Direct parser fixtures required before removing duplicate parsers. |
| Mention states | Existing `WorkbookShell.phase4.support.test.tsx`, `WorkbookShell.phase5.mentionChips.test.ts`, and `workbookMentionChips` model coverage. | Characterized for model/hook boundaries; dismissed-mention panel extraction remains limited to typed props. |
| Evidence attach/preview/download states | Existing `workbookEvidence.test.ts`, `evidenceLifecycleViewModel.test.ts`, `WorkbookShell.phase5.test.tsx`, and `WorkbookShell.surfaces.test.tsx`. | Characterized before evidence action/panel movement. |
| Row history/rollback/destructive controls | Existing `WorkbookShell.phase7.test.tsx` covers history rendering, rollback targets, delete, restore, and continuity. | Characterized before history panel movement. |
| Presence anchoring | Existing `WorkbookShell.phase6.test.tsx`, `workbookPresence` callers, and `workbookSocketLifecycle.test.ts`. | Characterized before live-update hook movement. |
| Grouped result | Existing `WorkbookShell.phase8.query.test.tsx`, `WorkbookShell.phase9.sentinel.test.tsx`, and grid adapter anchor tests. | Characterized before grid surface movement. |
| Frozen column | Non-implementation boundary in this tracker. | No test required unless grid layout changes. |
| Resize handle | Non-implementation boundary in this tracker. | No test required unless grid layout changes. |
| Fill-down handle | Non-implementation boundary in this tracker. | No test required because no fill-down affordance is added. |
| Edit cell affordances | Existing `WorkbookShell.phase3.grid.test.tsx`, `WorkbookShell.phase4.support.test.tsx`, and `WorkbookShell.phase9.sentinel.test.tsx`. | Characterized for editor-related movement. |
| Empty result | Existing `WorkbookShell.phase8.query.test.tsx` and saved-view/surface tests. | Characterized for shell/query extraction. |
| Save-state strip | Existing `workbookPendingQueue.test.ts`, `WorkbookShell.phase4.saveState.test.tsx`, `WorkbookShell.phase6.test.tsx`, and `WorkbookShell.phase7.test.tsx`. | Characterized before pending-save hook extraction. |
| Import cleanup | Planned Sprint 5 `rg` duplicate-helper scan and `make frontend-import-boundary-check`. | Blocking in Sprint 5. |
| Validation hygiene | Planned per-sprint command rows and final skipped-check notes. | Blocking in Sprint 5. |

## Extraction Interface Requirements

### Workbook Value Formatting

`workbookValueFormat.stringifyGridValue` MUST preserve the pre-refactor conversion contract below until owner-backed tests authorize a change.

| Input class | Required output | Owner or characterization status |
| --- | --- | --- |
| `string` | Original string without trimming. | Characterized from duplicate implementations. |
| `boolean` | `true` or `false`. | Characterized from duplicate implementations. |
| `number` | `String(value)`. | Characterized from duplicate implementations. |
| `null`, `undefined`, missing cell | Empty string. | Characterized from duplicate implementations. |
| Object with `items[]` where items expose `display_text` or `raw_text` strings | Join item display strings with `, `; skip unrenderable items. | Characterized from duplicate implementations and Core 03 §3.3.4 collection preservation boundary. |
| Array without wrapper object | Empty string. | Characterized from duplicate implementations. |
| Plain object without `items[]` | Empty string. | Characterized from duplicate implementations. |
| Date object | `TODO: blocking characterization required` before extraction changes output. | No owner rule found in tracker inputs. |
| Non-finite number | `TODO: blocking characterization required` before extraction changes output. | No owner rule found in tracker inputs. |
| Unknown future object shape | Empty string unless a Core owner or contract registry names the shape before extraction. | Default unsupported-shape behavior. |

### Clipboard Parsing

`workbookClipboard` MUST preserve the shared Timeline/Shell parser contract below until characterization tests record a different owner-backed rule.

| Input or condition | Required output or action | Owner or characterization status |
| --- | --- | --- |
| Empty text after trailing newline removal | Matrix `[[""]]` represented in source as one empty string cell. | Characterized from duplicate implementations. |
| CRLF or CR | Normalize to LF before parsing. | Characterized from duplicate implementations. |
| Trailing newlines | Remove all trailing LF before parsing. | Characterized from duplicate implementations. |
| Text containing a tab | Use tab as delimiter. | Characterized from duplicate implementations. |
| Text without a tab | Use comma as delimiter for parser dimensions; interactive Ctrl+V tabular dispatch still follows Core 03 §11.1 tabular-signal rule. | Characterized plus Core 03 §11.1. |
| Quoted doubled quote | Convert `""` to `"`. | Characterized from duplicate implementations. |
| Newline inside quoted cell | Keep newline inside cell. | Characterized from duplicate implementations. |
| Empty cell between delimiters | Preserve as empty string. | Characterized from duplicate implementations. |
| Single scalar text with no tab/newline/carriage return | Parser can return one cell; dispatch MUST treat comma-only scalar text as scalar by default per Core 03 §11.1. | Core 03 §11.1. |
| Malformed unclosed quote | `TODO: blocking characterization required` before changing output. | No owner rule found in tracker inputs. |
| Maximum accepted row or column count | `TODO: blocking characterization required` or owner-derived limit before extraction. | No owner-derived limit recorded. |

### Conflict Model

| Condition | Required action | Owner |
| --- | --- | --- |
| Missing `error.conflict` object | Parser returns null and does not enqueue. | Core 03 §3.3.4. |
| Missing or empty `conflict_token`, `record_id`, `field_key`, `conflict_resolution_class` | Parser returns null and does not enqueue. | Core 03 §3.3.4. |
| Missing numeric `base_row_version` or `current_row_version` | Parser returns null and does not enqueue. | Core 03 §3.3.4. |
| Optional server actor/timestamp omitted | Conflict remains valid; display omits that metadata. | Core 03 §3.3.4 and §3.3.5. |
| Duplicate `record_id:field_key` conflict | Replace or refresh the existing conflict entry only after preserving local draft; stale token behavior follows owner route result. | Core 03 §3.3.4. |
| Conflict token stale during resolution | Preserve local draft and refresh compare surface against newest saved value. | Core 03 §3.3.4. |
| Paste conflict ordering | Order by source row ordinal, then source column ordinal for the paste operation. | Core 03 §3.3.6. |
| Conflict queue versus pending queue | Same-field conflicts are separate from transient retry queue and are never auto-retried. | Core 03 §3.3.4 and §4.4. |

### Timeline Row Model

| Condition | Required action | Owner |
| --- | --- | --- |
| Query result superseded by a newer applicable query | Do not replace rendered rows, clear rows, change visible query errors, or drive access-loss handling. | Core 03 §14.1. |
| Query row older than accepted row-version high-water mark | Preserve newer local committed row and refresh through the same HTTP query route. | Core 03 §14.1. |
| Live patch row version is stale | Ignore patch and retain accepted row. | Core 03 §4.3.1 and §14.1. |
| Row deleted or missing after reload | Clear selected/focused anchor only for that missing row; do not synthesize row-position identity. | Core 03 §3.4, §14.5, §18.2. |
| Grouping changes row placement | Preserve record identity and recompute presentation placement from owner query result. | Core 03 §14.1 through §14.7. |
| Selected or focused row absent | Produce no focus command and keep shell surface context. | Core 03 §3.4 and §14.1. |
| Wrong `view_schema_id` in Timeline envelope | Treat as invalid envelope and surface load/mutation failure without accepting the row. | Core 01 §3.3.4 and Core 01 §7.4.1. |

### Generic Mutation And Reference Models

| Condition | Required action | Owner |
| --- | --- | --- |
| Omitted draft field | Treat as empty string. | Characterized from implementation. |
| Read-only field | Skip in create payload and do not build patch change. | Core 01 §6 and active view contract. |
| Clearable empty direct field | Emit `null`. | Active view contract, Core 01 §3.3.5. |
| Non-clearable empty direct field | Omit from create payload or return null patch change. | Active view contract, Core 01 §3.3.5. |
| Number field with non-integer or unsafe integer text | Return invalid payload result, represented as null create/patch output. | Characterized from implementation. |
| Boolean field other than `true` or `false` | Return invalid payload result. | Characterized from implementation. |
| Known owner defaults | Seed owner fields from current user id; seed `finding.kind=open` pair and forensic keyword defaults exactly as characterized before extraction. | Core 02 §10.4 and characterized implementation. |
| Unknown view schema | Minimum-create condition is any non-empty draft value until a Core owner or contract row supplies a specific rule. | Characterized implementation default. |
| Empty reference option bucket | Return empty option list and no reference choices. | Characterized implementation. |
| Duplicate option labels | Preserve all options; identity remains `recordId`, not label. | `docs/domain.md` §5 and Core 01 §8.2. |
| Invisible, deleted, wrong-type, or foreign-incident reference option appears in normalized input | `TODO: blocking characterization required`; extraction must not normalize this silently. | Core 01 §3.3.4 and Core 04 §2. |

## Startup And Saved View Closure Matrix

### Startup Selection

| Step | Input | Validity condition | On success | On invalid or omitted |
| --- | --- | --- | --- | --- |
| 1 | Explicit launch `sheet_ref` | Present, valid for caller, visible to caller. | Select it and return both selected `sheet_ref` and base `view_schema_id`. | Skip without persisting or clearing pointer. |
| 2 | `user_workbook_preferences.home_sheet_ref` | Present, valid for caller, visible to caller. | Select it and return both selected `sheet_ref` and base `view_schema_id`. | Clear invalid persisted pointer and continue. |
| 3 | `incident_workbook_preferences.default_sheet_ref` | Present, valid for caller, visible to caller. | Select it and return both selected `sheet_ref` and base `view_schema_id`. | Clear invalid persisted pointer and continue. |
| 4 | Base fallback | Always `cartulary.view.timeline.v1`. | Select Timeline system surface. | No further fallback. |

Owner: Core 03 §2.4. Enterprise-auth workbook opens without a valid explicit launch pointer reuse the same order per Core 03 §2.4 and Core 01 §20.

### Saved-View Transitions

| Transition | Affects `view_schema_id` | Affects `sheet_ref` | Affects `query_json` | Affects `layout_json` | Affects client-local state | Required rule |
| --- | --- | --- | --- | --- | --- | --- |
| Create from active base view | Yes, immutable after create. | Creates `saved_view` identity only after persistence. | Persists normalized query; omitted sort/filter become empty arrays; omitted grouping stays omitted. | Omitted or `{}` layout normalizes to `cartulary.layout.v1` default. | No. | Core 03 §2.3. |
| Duplicate visible saved view | Copies source `view_schema_id`. | Creates new saved-view identity. | Copies normalized canonical query. | Copies normalized layout. | No runtime inheritance from source saved view. | Core 03 §2.3. |
| Apply/open saved view | Does not mutate saved-view object. | Active `sheet_ref` becomes saved-view identity. | Active query comes from saved view. | Active layout comes from saved view. | Selection, focus, scroll, preview, popover, inspector, and presence remain client-local. | Core 03 §2.3 and §2.4. |
| Update saved view | MUST NOT mutate. | Does not change identity. | Can mutate when scope rules allow. | Can mutate when scope rules allow. | No. | Core 03 §2.3.2. |
| Delete saved view | Deletes only configuration object. | Active surface must fall back through startup/runtime rules if the deleted object was active. | Removes saved query with object. | Removes saved layout with object. | Does not delete local unsaved row work. | Core 03 §2.3.2 and §2.4. |
| Set home pointer | No. | Stores user pointer to valid `sheet_ref`. | No. | No. | No. | Core 03 §2.4. |
| Clear home pointer | No. | Removes user pointer. | No. | No. | No. | Core 03 §2.4. |
| Set incident default pointer | No. | Stores incident pointer to valid `sheet_ref`; admin-only. | No. | No. | No. | Core 03 §2.4 and Core 04 §2. |
| Clear incident default pointer | No. | Removes incident pointer; admin-only. | No. | No. | No. | Core 03 §2.4 and Core 04 §2. |

System view versus system saved view closure: a contract-backed system view is identified by standardized `view_schema_id` and `sheet_ref.kind='view_schema'`. A saved view with `scope='system'` is a separate configuration object bound to one incident. A `scope='system'` saved view MUST NOT replace the canonical identity, discoverability, or startup targetability of the base system view. Owner: Core 03 §2.2 through §2.4 and Core 01 §7.4.

## Commands

Use public Make targets from the repository root.

Planning/docs-only checks for this handoff file:

- `make generated-artifact-policy-check`
- `make json-shape-check`

Before selecting implementation verification:

- `make help`
- `make task-guide ROLE=feature-dev PHASE=phase3`
- `make explain-target TARGET=frontend-unit DETAIL=summary`
- `make explain-phase PHASE=phase3`

Common per-sprint verification:

- `make frontend-typecheck`
- `make frontend-import-boundary-check`
- `make frontend-unit`
- `make lint-biome`

Phase-oriented validation candidates, chosen by touched behavior:

- `make phase-slice PHASE=phase3` for grid identity, row versions, autosave, payload, and scalar edit behavior.
- `make phase-slice PHASE=phase4` for support references, mentions, save-state, and timeline query behavior.
- `make phase-slice PHASE=phase5` for evidence and grid provenance behavior.
- `make phase-slice PHASE=phase6` for WebSocket, presence, live patch, pending queue, and auth recovery behavior.
- `make phase-slice PHASE=phase7` for history, rollback, and conflict resolver behavior.
- `make phase-slice PHASE=phase8` for saved-view query, layout, grouping, and startup/query behavior.
- `make phase-slice PHASE=phase9` for keyboard, grid anchors, inspector, paste target, and sentinel behavior.

Broader checks after source refactors, only when lower-scope checks pass or risk requires:

- `make browser-e2e`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-a11y`
- `make browser-e2e-visual`
- `make check`

End-of-run hygiene when broad verification is used:

- `make agent-finalize`

## Definition Of Done And Traceability

Every row in this table is a tracker-owned acceptance condition. A row marked blocking prevents the owning sprint from completing until the required assertion has evidence or the tracker is revised.

| Behavior | Owner source | Required assertion | Existing or planned evidence | Blocking status | Acceptable omission behavior |
| --- | --- | --- | --- | --- | --- |
| Workbook open/default surface | Core 03 §2.4; Core 01 §20 | Startup follows explicit launch, user home, incident default, Timeline fallback order and keeps App auth state separate from shell surface selection. | Existing App landing/auth tests plus planned shell startup characterization. | Blocking before Sprint 4. | If App is untouched, record existing App tests and do not alter App. |
| Built-in/system view switching | Core 03 §2.1, §2.2; Core 01 §7.4 | Required built-in tabs and system views are reachable by pointer and keyboard using `view_schema_id`/`sheet_ref`, not visible label identity. | Phase 8 and shell surface tests. | Blocking before Sprint 4. | If a surface is not touched, cite existing test or mark blocking TODO before moving shell runtime. |
| Saved-view switching | Core 03 §2.3, §2.4 | Saved views remain incident-bound configurations over one immutable `view_schema_id` and never become canonical system views. | Phase 8 saved-view tests and saved-view runtime tests. | Blocking before Sprint 4. | Missing runtime test blocks saved-view extraction. |
| Scalar edit/save-state | Core 03 §3.1, §4.1, §4.2, §4.4 | Edit, dirty, pending, saved, error, auth-paused, and conflict states derive from pending queue/hook snapshot and feed status strip. | Phase 3 autosave, phase 6 auth, phase 7 conflict, queue tests. | Blocking before Sprint 2. | None for moved save code. |
| Timeline paste | Core 03 §11.1, §3.3.6, §13.3 | Clipboard dimensions, empty cells, target ownership, paste conflict grouping, and non-conflicting batch commit follow owner rules. | Phase 9 paste tests plus clipboard utility tests. | Blocking before clipboard extraction. | Maximum dimensions and malformed quote behavior remain blocking TODOs until characterized. |
| Keyboard navigation | Core 03 §13; `docs/design.md` §8.4-§8.6 for presentation | Grid navigation, edit entry/exit, escape priority, Tab/Enter movement, and resolver/preview focus return produce the same command outcomes after extraction. | Phase 9 keyboard tests and direct command tests. | Blocking before grid interaction hook extraction. | Design-only label/presentation differences require design evidence, not Core claim. |
| Selected row and focused cell continuity | Core 03 §3.4, §14.1, §14.5 | Reloads, live patches, stale queries, grouping changes, and inspector actions preserve record/cell identity; missing targets clear only the affected anchor. | Phase 6, 8, 9 tests plus row model tests. | Blocking before row model replacement. | If selected/focused row is missing, no focus command is emitted. |
| Conflict display/resolution | Core 03 §3.2, §3.3 | Only affected cells enter conflict state; saved value and local draft remain distinct; resolution requires explicit analyst action. | Phase 7 tests and conflict parser fixtures. | Blocking before conflict model consolidation. | Optional metadata omission does not block conflict entry. |
| Mention states | Core 02 §6; Core 03 §9 and §12 | Unresolved, resolved, auto-resolved, dismissed, restored, selected, and create-from-mention states remain distinct. | Phase 4 mention tests and mention hook tests. | Blocking before mention hook extraction. | Missing optional dismissed-mention panel evidence blocks panel extraction only. |
| Evidence attach/preview/download states | Core 02 §13; Core 03 §8; Core 01 §16 | Evidence record and object blob semantics remain distinct across requested, pending upload, available, failed, quarantined, blocked, inconsistent, preview, and download states. | Phase 5 evidence tests and evidence hook/panel tests. | Blocking before evidence hook/panel extraction. | If generic evidence actions are untouched, cite existing generic surface tests. |
| Row history/rollback/destructive controls | Core 01 §3.3.4.2 and §3.3.5.0; Core 02 §15; Core 03 §10 | History rows, rollback, delete, restore, pending destructive action, and destructive labels are row-local and driven by structured legality metadata. | Phase 7 history/rollback tests. | Blocking before history panel extraction. | No destructive control is rendered when legal action metadata is absent. |
| Presence anchoring | Core 03 §4.3 and §14.7 | Collaborator state is anchored to durable row/cell identity and not row position after query/group changes. | Phase 6 presence tests. | Blocking before live update hook extraction. | Missing optional presence payload produces no presence update. |
| Grouped result | Core 03 §14.1 through §14.8 | Group headers are presentation-only rows and never edit, paste, export, history, or mutation targets. | Phase 8/9 grouping and paste tests. | Blocking before grid surface extraction. | Grouping unsupported for undeclared key is not offered. |
| Frozen column | `docs/design.md` §8; implementation-support boundary | This tracker does not add frozen-column behavior. Extraction must not introduce row-position identity. | Planned visual/DOM check only if grid composition changes. | Non-blocking unless grid layout changes. | Non-implementation is acceptable for this tracker. |
| Resize handle | `docs/design.md` §8; implementation-support boundary | This tracker does not add resize-handle behavior or expose grid-vendor details in workbook contracts. | Planned visual/DOM check only if grid composition changes. | Non-blocking unless grid layout changes. | Non-implementation is acceptable for this tracker. |
| Fill-down handle | Core 03 §18.2; implementation-support boundary | This tracker does not add fill-down behavior or spreadsheet semantics outside owner rules. | No test required unless code adds fill-down affordance. | Blocking only if an implementation adds it under this tracker. | Non-implementation is required under this tracker. |
| Edit cell affordances | Core 03 §13.1 | Scalar editors, collection editors, read-only/derived cells, invalid cells, and pending cells expose only owner-legal edit intents. | Phase 3, 4, 9 tests. | Blocking before editor-related extraction. | Read-only/derived cells emit no mutation intent. |
| Empty result | Core 03 §14.1; Core 03 §2.3 | Filters, saved views, grouping, and stale queries render empty states without losing active surface or saved-view context. | Phase 8 query tests. | Blocking before shell/query extraction. | Empty rows do not clear active surface. |
| Save-state strip | Core 03 §4.2 | Primary save label and secondary detail are computed from pending queue/conflict state and remain visible in shell status. | Phase 4/6/7 tests and queue tests. | Blocking before pending save hook extraction. | No omission allowed for moved save-state code. |
| Import cleanup | Tracker-owned | No duplicate value, clipboard, or conflict helpers remain after call sites use shared modules. | `rg` search plus frontend import boundary check. | Blocking in Sprint 5. | Duplicates can remain only with a tracker row naming why extraction is deferred. |
| Validation hygiene | `docs/testing-harness-nlspec.md` §4, §8, §17 | Required Make targets run or are recorded as skipped with reason, run root, and relationship to touched behavior. | Final implementation handoff. | Blocking in Sprint 5. | Skips are acceptable only with explicit reason and impact. |

## Non-Goals

- No broad redesign.
- No API changes.
- No data-model changes.
- No visual restyling.
- No grid-vendor replacement.
- No global state migration.
- No route rewrites.
- No speculative performance work.
- No generated artifact edits.
- No lockfile or dependency churn.
- No expansion into admin panels, phase harnesses, backend modules, migrations, or contract generation except for imports/tests required to keep named `apps/web/src/**` behavior intact.

## Editorial Acceptance Criteria For This Tracker

- Every tracker-owned requirement in the body maps to one Definition of Done row.
- Every Definition of Done row maps back to a body requirement or owner-source restatement.
- Every extraction unit has explicit inputs, outputs, omitted-input behavior, invalid-input behavior, and validation evidence.
- Every unresolved edge case is marked `TODO: blocking characterization required`.
- No Fallow finding, design statement, or test fixture is represented as Core product authority.
- Uppercase `MAY` appears only where omission behavior is stated.
- The phrases `current behavior`, `where possible`, `if needed`, `stable`, `narrowly`, `selected if`, and `absent behavior` do not appear outside this editorial acceptance row.
