# Web Workbook Modular-Refactor Tracker

## 1. Session, authority, and target

### 1.1 Session header

| Field | Value |
| --- | --- |
| Planning timestamp | 2026-07-31T09:12:19-04:00 |
| Branch | main |
| Baseline commit | c22ce208a176c595f74acdae5b46b3a947b6fb72 |
| Initial worktree | Clean |
| Execution authorization timestamp | 2026-07-31T09:44:27-04:00 |
| Execution-start worktree | This tracker untracked; no other change |
| Deliverable | docs/handoffs/web-workbook-modular-refactor-tracker.md |
| Exactly one primary target | Workbook application-composition boundary inside apps/web/src |
| Execution posture | Authorized remediation through P-00, F-01, S-01-S-14, and V-01 |
| Authorized repository change | Controlling tracker, authored protocol projections, generated protocol outputs through Make, Workbook implementation/tests, boundary/accounting inputs, and implementation-support README |
| Explicitly unauthorized | Normative-owner or domain-vocabulary changes without a demonstrated contradiction; server HTTP/WebSocket/database changes; unrelated UI redesign; hand-edited generated output or lockfiles |

This tracker applies the controlling instructions in
temp/planner-prompt.md through the workflow in
docs/handoffs/cartulary_modular_refactor_planning_framework.md. It is a
current-state analysis and the controlling execution ledger, not a claim that
a generic React folder pattern is desirable. Boundaries follow the state
lifetime, authority, and side-effect ownership demonstrated by the live
source. Section 19 supersedes planning-only constraints where they conflict
with the authorized remediation program.

Commit 80c7a5f0 completed an earlier workbook modular-refactor program. Commit
c22ce208 archived its tracker at
docs/archive/web-app-src-module-refactor-tracker.md. That archive is retained
implementation evidence for the shell, startup, query, collaboration,
continuity, surfaces, and layout boundaries it closed. It is not restored,
reopened, or treated as a current backlog. This tracker begins from the live
post-refactor source and records only residual composition coupling.

### 1.2 Authority posture

| Precedence | Source | Use in this tracker |
| --- | --- | --- |
| 1 | docs/spec/00_document_set_status_and_precedence.md | Authority order and owner navigation |
| 2 | docs/spec/01_architecture_storage_and_view_contracts.md | Architecture, storage, view-schema, and projection contracts |
| 3 | docs/spec/02_domain_model_schema_and_history.md | Stable record, version, history, and field identity |
| 4 | docs/spec/03_workbook_interaction_collaboration_and_workflows.md | Normative workbook, saved-view, grid, mutation, conflict, collaboration, inspector, evidence, density, and extension behavior |
| 5 | docs/spec/04_security_deployment_and_conformance.md | Authorization, revocation, safe public error, and conformance boundaries |
| 6 | Adopted subsystem NLSpecs, including docs/extension-subsystem-nlspec.md and docs/testing-harness-nlspec.md | Extension discovery/teardown and test-harness ownership |
| 7 | contracts/**, tools/test_catalog_owner.json, tools/test_families/**, and tools execution manifests | Typed projections, verification routing, and harness mechanics |
| 8 | docs/domain.md and docs/design.md | Vocabulary/owner navigation and design direction within their stated scopes |
| 9 | Live source, tests, manifests, and retained run artifacts | Current implementation and evidence |
| Supporting only | Archived trackers and docs/research/R04, R08, and R09 | Historical decisions and non-authoritative grid/responsive research |

No owner contradiction was found. If implementation later exposes one, the
slice stops before source movement and records BLOCKED with the conflicting
owner clauses. Tests and generated projections do not override adopted owners.

### 1.3 Goal, success criteria, scope, and exclusions

The goal is to make workbook functions, components, state, side effects,
adapters, and interaction logic independently modifiable at their logical
ownership boundaries. Success means a future implementer can execute one
behavior-preserving slice at a time without rediscovering the source tree,
public contracts, state lifetimes, validation routes, or rollback boundaries.

In scope:

- Workbook composition under apps/web/src/workbook.
- App entry and teardown only where they establish the Workbook lifetime.
- Services and package facades where the typed Workbook boundary requires a
  bounded implementation change.
- Authored protocol-ts projections, frontend import boundaries, source
  ownership, and test-family inputs required by an executing slice.
- Existing unit, owner, browser, visual, accessibility, and harness evidence.

Out of scope:

- UI or interaction redesign, CSS restyling, and DOM reshaping unrelated to a
  required safety, partial-completion, or invalid-contract correction.
- New custom sheets or treating saved views as sheets.
- Backend, HTTP, WebSocket, database, storage, revision, or authorization
  behavior changes. Frontend machine projection of existing HTTP owners is in
  scope.
- Direct grid-vendor work in application code.
- Hand edits to generated roots, lockfiles, generated topology, or schedules.
  Authored inputs may change and must be regenerated through Make.
- Broad App.tsx, Network Flow, Imports, extension, or package refactors without
  a newly demonstrated inseparable Workbook dependency.

### 1.4 Evidence levels

The inventory uses three explicit inspection levels:

- Full semantic read: the file was opened for its state, side effects, public
  interface, callers, and planned seam.
- Targeted semantic read: relevant declarations, imports, hooks, state
  transitions, and call sites were opened.
- Static inventory: path, owner, role, and imports were machine-accounted; the
  file is not proposed for movement in this tracker.

No static-only file is silently treated as understood implementation detail.
Every file proposed as a migration source has at least a targeted semantic
read. For S-01, the port declarations, adapter branches, controller state
transitions, surface call sites, tests, and boundary-policy shape were opened
at the exact planned seam.

## 2. Repository and ownership inventory

### 2.1 Exact reconciliation

tools/frontend_source_ownership.json and the live apps/web/src TypeScript tree
match exactly: 320 manifest paths, 320 live .ts/.tsx paths, no missing path, no
extra path, and no duplicate ownership entry.

| Owner | Production | Test/support | Total |
| --- | ---: | ---: | ---: |
| web.app | 23 | 11 | 34 |
| web.collaboration | 1 | 1 | 2 |
| web.entry | 1 | 0 | 1 |
| web.extensions | 3 | 1 | 4 |
| web.imports | 1 | 1 | 2 |
| web.network_flow | 24 | 13 | 37 |
| web.services | 9 | 7 | 16 |
| web.shared | 4 | 0 | 4 |
| web.testing | 0 | 14 | 14 |
| web.workbook | 142 | 64 | 206 |
| **Total** | **208** | **112** | **320** |

apps/web/src/README.md remains a narrative ownership guide and is not an
executable inventory source. The complete current file ledger is in Appendix
A.

### 2.2 Composition hotspots

| Path | Lines | Direct-import fan-out | Demonstrated responsibilities | Disposition |
| --- | ---: | ---: | --- | --- |
| apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx | 2,916 | 63 | Runtime wiring, authoritative rows, freshness, pending/replay, conflicts, collaboration, selection, drafts, bulk actions, inspector, evidence, continuity, grid refs, and rendering | Decompose only after typed I/O prerequisites; S-07 through S-14 |
| apps/web/src/workbook/components/EntityWorkbookSurface.tsx | 1,814 | 32 | Entity drafts, paste, create/patch parsing, merge workflow, inspector, timeline preview, continuity, and rendering | S-03 and S-04 |
| apps/web/src/workbook/components/GenericWorkbookSurface.tsx | 1,421 | 32 | Generic drafts, create/patch, party-link workflow, inspector, evidence/coordination bindings, continuity, and rendering | S-01 and S-02 |
| apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx | 900 | 28 | Assessment creation workflow, mutation state, inspector, grid interaction, and presentation | S-05 |
| apps/web/src/workbook/WorkbookShell.tsx | 1,057 | 41 | Application composition over established runtime, queries, commands, collaboration, layout, and surfaces facades | Retain; no broad shell refactor |
| apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx | 307 | 16 | Private concrete-surface renderer selection and owner input projection | Retain as the public rendering facade |
| apps/web/src/app/App.tsx | 1,614 | 17 | Application/auth/route composition and Workbook entry/teardown | Defer broad App work; retain Workbook lifetime evidence only |

File size is not itself a refactor reason. Each planned slice below is tied to
a demonstrated authority, state-lifetime, side-effect, or dependency problem.

### 2.3 Adjacent-area disposition

| Adjacent area | Material coupling found | Decision |
| --- | --- | --- |
| packages/grid-adapter | Application imports semantic grid types and components; react-data-grid remains package-private and the current boundary check passes | Evidence only; no package change |
| packages/view-contracts | Runtime-safe authored facade around generated view contracts is already the intended dependency | Retain; no new contract or generated edit |
| packages/ui-contracts | Stable selectors/test IDs are already consumed through the facade | Retain; no selector migration |
| packages/test-utils | Browser choreography only; no production dependency | Evidence only |
| apps/web/src/app/App.tsx | Establishes incident/auth lifetime but does not force the residual surface internals | Inspect entry/teardown; defer broad application refactor |
| apps/web/src/networkFlow | Consumed through NetworkFlowFeature and existing discovery gates; no Workbook back edge or cycle | Defer to Network Flow owner-specific work |
| apps/web/src/extensions | Owns discovery/availability; Workbook consumes the established projection | Preserve extension gate and teardown; no package-presence shortcut |
| apps/web/src/services | Raw transport is still consumed by some surface/action modules | Include only through typed ports in named slices |
| tools/frontend_import_boundaries.json | Machine enforcement for demonstrated imports | Include in future slices only when adding the corresponding rule |
| Backend and contracts | No inseparable server change is required | Explicitly excluded |

## 3. Import and dependency map

### 3.1 Intended direction

~~~text
app/auth/route composition
  -> WorkbookShell
    -> shell runtime + view state + saved views
    -> query owners
    -> semantic mutation command ports
    -> collaboration coordinator
    -> inspector/continuity/layout owners
    -> WorkbookSurfacesFacade
      -> owner surface controller
        -> pure models + narrow ports
        -> grid-adapter/view-contracts/ui-contracts facades
        -> presentation

services/httpTransport -> wire only
grid-adapter -> direct grid-vendor integration only
generated protocol -> protocol/view facade -> application adapters
~~~

Shared coordinators may depend on narrow semantic ports. They must not import
surface implementations or broad shell facades. Presentation may issue owner
commands and render owner snapshots; it must not interpret transport envelopes
or manufacture wire intent.

### 3.2 Measured current graph

The production-only static import graph contains zero cycles.

| Metric | Current evidence |
| --- | --- |
| Highest fan-out | TimelineWorkbook 63; WorkbookShell 41; Entity surface 32; Generic surface 32; Assessment surface 28; NetworkAnalysisWorkspace 19; App 17; WorkbookSurfacesFacade 16 |
| Highest fan-in | workbookSurfaceRegistry 45; services/browserApi 39; workbookTimelineModel 33; services/workbookApi 30; workbookQuery 24; WorkbookQueryRow 18; WorkbookMutationRuntime 14; extensionAvailability 13; workbookShellContracts 13 |
| Largest cross-directory edges | workbook -> workbook 537; workbook -> services 52; workbook -> shared 20; app -> services 17; networkFlow -> services 16 |
| Grid vendor | No direct react-data-grid import under application source; package boundary is intact |
| Generated protocol | Existing facade rules prevent arbitrary generated imports |
| Service back edges | services cannot import workbook; current check passes |
| Private renderers | Concrete surface imports are confined to WorkbookSurfacesFacade |

### 3.3 Residual coupling findings

| ID | Finding | Evidence | Classification | Planned response |
| --- | --- | --- | --- | --- |
| C-001 | Generic command ports return raw HTTP-shaped ok/status/payload results, so the generic controller and surface import readEnvelope and interpret wire failures | mutations/workbookMutationCommandPorts.ts, mutations/createWorkbookMutationCommandPorts.ts, hooks/useGenericSurfaceMutationController.ts, components/GenericWorkbookSurface.tsx | must_fix; first prerequisite | S-01 semantic accepted/rejected outcome |
| C-002 | Entity presentation performs response/error/conflict interpretation alongside drafts, paste, inspector, and merge workflow | components/EntityWorkbookSurface.tsx | must_fix after S-01 pattern | S-03 typed outcome, then S-04 merge owner |
| C-003 | Timeline action hooks call apiPath/fetchWorkbookJSON/readEnvelope/parseErrorMessage directly | useTimelineHistoryActions, useTimelineMentionActions, useTimelineClipboardPasteController, useTimelineMutationCommands, useTimelineRowsLoader, useTimelineCreateRelatedWorkflow, useTimelinePendingReplayController, and useTimelineEvidenceAttach | must_fix, high-risk | S-07 through S-09, one I/O seam at a time |
| C-004 | TimelineWorkbook composes live collaboration patching, active-surface registration, bulk actions, scalar drafts, mutation admission, freshness, DOM refs, inspector, and rendering | TimelineWorkbook.tsx hooks/state/effects/callbacks and 63 imports | should_fix after I/O stabilization | S-10 through S-14 |
| C-005 | Generic party creation and subsequent linking is a multi-command workflow owned inside the surface | GenericWorkbookSurface.tsx createPartyFromText and submitPartyLinkPatch | should_fix | S-02 typed workflow controller |
| C-006 | Entity merge confirmation, response handling, refresh, inspector retarget, and draft cleanup share presentation ownership | EntityWorkbookSurface.tsx | should_fix | S-04 merge workflow controller |
| C-007 | Assessment creation submission and presentation state share the concrete surface | AssessmentWorkbookSurface.tsx | should_fix | S-05 creation controller |
| C-008 | Existing startup, query-owner, reference-broker, transaction-ID, collaboration, inspector, continuity, surfaces, and layout boundaries are closed and passing | Archived tracker plus live modules and boundary rules | intentional/no_action | Preserve; do not re-plan completed slices |
| C-009 | Network Flow and extension workspaces already enter through facades and discovery/availability contracts | NetworkFlowFeature.tsx, extension availability projection, boundary manifest | intentional/no_action | Characterize visibility/teardown only |
| C-010 | There is no production import cycle and no direct grid-vendor leak | Static graph and import-boundary baseline | intentional/no_action | Keep existing policy; add only exact new rules |

## 4. State-ownership matrix

| State class | Current owner | Authoritative source | Lifetime | Persistence | Writers | Readers | Reset/invalidation triggers | Current problems | Target owner |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Authenticated session | AuthGateway/App and browser auth service | Core 04 auth/session clauses | Auth epoch | Server session; limited browser cookie state | Auth responses and revocation handling | App, Workbook admission, collaboration recovery | Logout, expiry, revocation, failed current-user probe | No residual Workbook ownership defect | App/auth owner; injected recovery port |
| Visible/current incident | App route runtime and WorkbookShell props | Core 03 startup; Core 04 authorization | Route/incident epoch | URL plus authoritative server record | Route navigation and incident selection | App and Workbook composition | Route change, deletion, access loss, logout | App is large but not inseparably coupled | App route owner |
| Active sheet_ref | Workbook startup/view-state owners | Core 03 stable surface identity | Workbook instance | URL/startup selection; not custom user schema | Startup admission and explicit surface switch | Shell, surfaces facade, collaboration | Incident/auth epoch, invalid availability, fallback | Closed by prior refactor | Existing Workbook view-state owner |
| Active view_schema_id | Workbook startup/query/view-state owners | Core 01/03 view-schema identity | Active surface | URL/saved-view selection | Startup, surface and saved-view apply | Query and rendering owners | Surface switch, fallback, availability change | Closed | Existing view-state owner |
| Selected saved view | useWorkbookSavedViewController and saved-view models | Core 03 saved-view clauses | Workbook instance/active schema | Authoritative saved-view configuration | Create/apply/update/duplicate/reset commands | View bar and query state | Incident/schema switch, deletion, reset, access loss | Must not absorb presentation state | Existing saved-view controller |
| Startup selection/reservation | useWorkbookStartupAdmission and useWorkbookStartupController | Core 03 startup chain | Admission request | None | Startup effects only | Shell runtime | New request, incident/auth epoch, availability change, unmount | Closed by prior slice | Existing startup facade |
| Query/sort/filter/group/pagination | Per-surface query owners and useWorkbookQueryState | Core 03 query and saved-view clauses | Workbook instance and active schema | Saved configuration only where contract permits | View controls, saved-view apply, continuation | Query loaders, grid, status | Surface/schema/incident/auth invalidation | Timeline loaders still own raw I/O | Existing query state; S-07 typed Timeline I/O |
| Authoritative rows | Entity/assessment/generic query hooks and Timeline committed rows | Core 02/03 projection contracts | Query generation | Server projection only | Successful queries, accepted mutation refresh, live patches | Grids, inspectors, actions | Query invalidation, access loss, incident/surface change | Timeline admission and I/O remain spread | Existing query owners; S-07/S-13 |
| Per-record version high-water marks | Timeline row mutation coordinator over the committed-row admission helper | Core 02 row_version; Core 03 stale-event rules | Workbook/incident coordinator instance | Memory only | Query admission, accepted fresh/replayed mutation, action outcome, conflict server state, and collaboration replay/live patch | Query/action/live stale filters, dispatch-time base-version lookup, and continuity sequencing | Incident/auth/access invalidation and coordinator disposal | Closed: one coordinator owns admission and keeps committed, pending, conflict, and visible-editor partitions distinct | S-13 mutation coordinator |
| Active grid cell/selection | Grid continuity and surface-local semantic state | Core 03 keyboard/selection continuity | Mounted surface | None | Grid semantic callbacks and continuity restoration | Editors, bulk actions, inspector | Surface/schema change, row removal, explicit clear | Timeline composition wires several adapters | Existing continuity port; S-12/S-13 consumers |
| Bulk selection | TimelineWorkbook | Core 03 bulk mutation clauses | Timeline surface | None | Selection interactions and query-row reconciliation | Bulk view-bar actions | Surface/query identity change, removed rows, completion | State/effect/submission share composition | S-11 bulk-tag controller |
| Draft row/create form | Concrete generic/entity/assessment/Timeline workflows | Core 03 low-friction creation | Surface/workflow | None | User input and workflow reset | Validation and mutation commands | Success, cancel, surface/incident/auth loss | Several concrete surfaces own orchestration and presentation | S-02, S-04, S-05; existing Timeline workflow after S-09 |
| Invalid scalar editor drafts | Timeline editor draft registry | Core 03 invalid-draft retention | Timeline schema generation and semantic row/field/surface identity | None | Editor validation and correction | Cell renderers, mutation/replay controllers, and continuity owner | Matching successful commit, cancel/conflict resolution, row/schema removal, or access invalidation | Closed: drafts and semantic input refs share one private lifetime owner with no vendor coordinates | S-12 editor draft registry |
| In-flight mutations | WorkbookMutationRuntime plus surface controllers | Core 03 save-state/mutation semantics | Logical action | None | begin/finish explicit mutation and action controllers | Status strip, teardown, conflict handling | Completion, rejection, invalidation, dispose | Generic/entity/assessment outcomes remain HTTP-shaped | Semantic ports S-01/S-03/S-05/S-06/S-09 |
| Local pending queue | Timeline pending-save/replay owners and workbookPendingQueue | Core 03 offline/pending queue | Incident/auth epoch | Approved local browser persistence | Transient failures and replay transitions | Status, replay controller, mutation admission | Successful replay, explicit removal, access/session invalidation | Replay controller also owns raw I/O | S-08 typed replay port |
| Same-field conflict queue | WorkbookMutationRuntime and conflict model | Core 03 same-field explicit resolution | Workbook instance | Memory only | Mutation rejection parser and conflict resolution | Resolver, status, inspector/continuity | Resolution, invalidation, access loss, dispose | Callers still parse transport payload | Normalized outcomes beginning S-01 |
| client_txn_id | SecureTransactionIdPort, command ports, pending entries | Core 03 mutation identity/replay clauses | One logical action across uncertain replay | Pending entry while unresolved | Command adapter creates; replay reuses | Server requests, conflict recovery, queue | Logical completion; explicit recovery creates new ID | Timeline direct action code still crosses boundary | Existing secure ID port plus S-08/S-09 |
| Inspector open/panel/subject | useWorkbookInspectorCoordinator per surface | Core 03 REQ-03-291/292 | Mounted surface | None | Semantic selection and inspector actions | Inspector panels and continuity | Close, subject removal, surface/incident/auth reset | Coordinator boundary is closed; surfaces still compose workflows | Existing coordinator; no redesign |
| Destructive confirmations/workflow forms | Concrete owner surfaces and feature bindings | Core 03 owner workflow clauses | Dialog/action | None | User action | Presentation and command controller | Confirm/cancel/success/invalidation | Merge/party/create workflows mix layers | S-02/S-04/S-05 |
| Presence | Collaboration coordinator projection and presence presenters | Core 03 presence semantics | Collaboration session | Ephemeral server stream | Decoded presence events | Cell markers, status | Disconnect, incident/sheet/auth change, revocation | Closed; Timeline only wires projection | Existing coordinator; S-10 thin binding |
| Collaboration replay high-water marks | WorkbookCollaborationCoordinator | Core 03 resume/replay/gap/stale rules | Collaboration session/incident | Resume protocol state | Connection/replay/live event reconciliation | Event admission and refresh decisions | Gap recovery, new incident/session, access loss, dispose | Closed owner; binding remains verbose | Existing coordinator; S-10 consumer binding |
| Background jobs | Imports/Network Flow/feature-specific owners | Relevant subsystem NLSpec and Core 03 extension clauses | Feature workspace/job | Server plus polling/session state | Feature controllers | Feature UI/status | Completion, failure, teardown, access loss | No inseparable Workbook defect | Adjacent owner; characterize extension teardown |
| Evidence preview/attachment | Evidence feature binding, evidence service, surface inspector | Core 03 evidence clauses | Selected subject/workflow | Server record/blob; preview ephemeral | Selection, upload, attach command | Inspector/panel | Subject/surface/incident/auth change, completion | Attach port is void/raw-error shaped | S-06A semantic evidence outcome |
| Density/presentation preferences | Layout facade/density model and account preference | Core 03 REQ-03-286/289 | Account and Workbook layout | Account preference where allowed | Account settings/layout controller | Shell and surfaces | Preference/auth change; responsive recomputation | Closed and measured | Existing layout facade |
| Notifications/status strip | Surface mutation/query state and Workbook status components | Core 03 saved/syncing/pending/conflict distinctions | Surface/action | None | Query/mutation/replay/conflict owners | Status strip and notices | Next action, success, dismissal, invalidation | Messages sometimes derived from raw payload | Outcome ports return presentation-safe meaning |

No row is assigned to a generic global store. Target ownership follows the
shortest authoritative lifetime that can perform every required reset.

## 5. Runtime-flow characterization

| Flow | Entry point | Components/modules traversed | State mutated | Network or storage side effects | Contract owner | Current coupling problem | Characterization evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1. Authenticated application entry | main/AppRoot -> AuthGateway | AppRoot, AuthGateway, App, route runtime | Session/auth epoch, route, current incident | Current-user/session requests | Core 04 | No residual Workbook defect; entry establishes cleanup boundary | App.auth tests and Workbook access-loss tests |
| 2. Incident workbook bootstrap | App renders WorkbookShell | startup admission/controller, shell runtime, view state, query owners | sheet_ref, schema, saved-view selection, query generation | Startup/view/saved-view reads | Core 03 startup | Closed by prior startup facade | Startup controller/admission tests and shell query tests |
| 3. Surface or saved-view selection | SystemViewSwitcher/ActiveSurfaceSavedViewSelector | saved-view controller, view state, surfaces facade | active sheet/schema, saved view, query state, inspector/continuity reset generation | Saved-view reads/writes when commanded | Core 03 saved views | Preserve; no custom sheets | Saved-view model/controller and shell surface tests |
| 4. Query and viewport rendering | Query state change or refresh | per-owner query hook, row model, surface, grid-adapter | load state, rows, continuation, selection admission | GET view rows/reference data | Core 01/03 | Timeline loader owns raw transport | Query tests, grid tests, surface tests |
| 5. Blank-row/low-friction creation | Draft row/create control | surface workflow, semantic command port, mutation runtime, refresh | draft, syncing/error state, rows after refresh | POST row/create endpoint | Core 03 grid-first creation | Generic/entity/assessment commands return raw envelopes | Payload/autosave/surface tests; S-01 adds outcome characterization |
| 6. Existing-row inline edit | Grid editor commit | grid adapter semantic intent, surface/controller, command/runtime | draft, in-flight, conflict or authoritative row | PATCH record | Core 02/03 | Response parsing lives in presentation for generic/entity and Timeline actions | Grid/payload/save-state/conflict tests |
| 7. Autosave dispatch | Valid edit commit | mutation runtime, timeline mutation commands/request model | syncing, transaction tracking, pending on transient failure | PATCH with stable client_txn_id/base version | Core 03 autosave | Timeline transport/action boundary remains broad | Workbook autosave/action-sequencing tests |
| 8. Transient failure and replay | Mutation transient rejection/online replay | pending saves, queue model, replay controller, mutation runtime | pending entry/state, replay attempt, authoritative refresh | Local queue storage and repeated HTTP mutation | Core 03 pending/replay | Replay controller owns both queue policy and raw I/O | Pending queue/replay/save-state tests and stateful browser evidence |
| 9. client_txn_conflict | Server conflict response | mutation action/replay, conflict recovery ID path | blocked transaction/recovery state | Explicit recovery request with new ID | Core 03 transaction conflict | Direct Timeline transport interpretation | Autosave/action sequencing and recovery tests |
| 10. Same-field conflict | PATCH rejection | outcome parser, mutation runtime conflict queue, resolver | conflict entry, focus target, resolution draft | Resolution mutation | Core 03 conflict resolver | Generic/entity surfaces parse raw payload before registering | Conflict model/resolver/shell surface tests |
| 11. WebSocket connect/resume/replay/gap | Collaboration session attach | collaboration transport, coordinator, projection refresh controller | connection state, resume/high-water, invalidation reason | WebSocket connect/resume; refresh on gap | Core 03 collaboration | Owner is closed; Timeline consumer binding remains in composition | Collaboration coordinator/session/shell tests and stateful browser evidence |
| 12. Live record_changed | Decoded collaboration event | coordinator, active surface port, query-row patch/freshness | authoritative rows and version high-water | None unless refresh required | Core 03 record_changed | Timeline registers patch/refresh callbacks inline | Collaboration and query-row patch tests |
| 13. Selection/focus/scroll continuity | Mutation/query/collaboration row replacement | semantic continuity port, grid adapter, surface anchor/controller | semantic anchor, focus snapshot, scroll restoration | None | Core 03 continuity | Port is closed; Timeline composes several ownership inputs | Sentinel/grid/continuity unit and browser/a11y evidence |
| 14. Inspector lifecycle | Semantic row/action selection | inspector coordinator, owner feature sections, continuity | open state, panel, subject, action form | Owner-specific action requests | Core 03 REQ-03-291/292 | Generic/entity/assessment workflow orchestration remains in surface | Inspector coordinator/shell inspector/browser evidence |
| 15. Evidence preview/attachment | Inspector evidence action | evidence binding, evidence service, command port, refresh | preview, attach progress/error, authoritative row | Blob create/upload/attach requests | Core 03 evidence | Evidence attach outcome is not semantic | Evidence service/model/shell/browser tests |
| 16. Session/revocation/access loss | Auth failure or typed invalidation | App recovery port, collaboration coordinator, lifecycle owners | protected rows, pending/conflicts, presence, inspector, query state | Auth probe/logout/navigation as owned | Core 04 and Core 03 cleanup | Closed routing; every new controller must implement same invalidation | App auth, query access-loss, collaboration tests |
| 17. Saved-view lifecycle | View bar commands | saved-view controller/model, query state, selector | saved configuration and active selection | CRUD saved-view requests | Core 03 saved views | No residual dependency defect; must remain separate from presentation state | Saved-view model/controller/shell tests |
| 18. Extension visibility/teardown | Availability projection and surface switch | extension discovery, surfaces facade, lazy feature, coordinator teardown | extension presence, active workspace, feature-local state | Discovery and feature-specific requests | Extension NLSpec and Core 03 | No package-presence shortcut; adjacent features stay deferred | Feature, extension, shell surface and browser tests |

## 6. Public-contract freeze

| Surface | Specific contract | Owner source | Current implementation path | Current evidence | Characterization required? | Refactor risk |
| --- | --- | --- | --- | --- | --- | --- |
| HTTP | Existing /api/v1 roots, methods, query parameters, request/response envelopes, CSRF behavior, public errors, and authorization | Core 01-04 and protocol contracts | services/httpTransport.ts, services/browserApi.ts, services/workbookApi.ts, mutation/query adapters | Payload, route-boundary, protocol, service, and shell tests | Yes for each action moved; byte-equivalent request and visible error | High |
| WebSocket | Existing path, event vocabulary, resume token, replay/gap, presence, record_changed, and stale-event semantics | Core 03 collaboration clauses | collaboration session plus WorkbookCollaborationCoordinator | Coordinator/message/shell/stateful evidence | Required before changing a consumer binding | High |
| Workbook UI | Grid-first create/edit, stable IDs, keyboard/paste, continuity, startup/fallback, saved views, inspector closed/retarget, pending/conflict distinction, density, read-only closure | Core 03 | WorkbookShell, surfaces facade, owner surfaces, grid adapter, layout/continuity/inspector owners | Workbook unit owners plus browser/measurement/a11y/visual evidence | Required per affected behavior | High |
| Client-visible storage/revision/projection | record_id, row_version, field_key, sheet_ref, view_schema_id, change_set_id, and client_txn_id meaning | Core 01-03 | Query rows, freshness, mutation ports, collaboration coordinator | Model, payload, query, conflict, collaboration tests | Yes when outcome shape moves | High |
| Generated artifacts | Generated protocol/UI roots are generator-owned and read-only | Adopted owners and generated-artifact policy | internal/gen and packages generated roots | generate-drift/generated policy retained evidence | No generated change planned | Medium if accidentally touched |
| Harness accounting | Tests route through authored owner catalogs/topology; Markdown is never executable input | Testing Harness NLSpec and tools owners | Make targets, test catalog, family inputs, execution topology | make task-guide/explain targets and retained summaries | Only if tests/files are added or moved; S-01 adds none | Medium |
| Accessibility/visual behavior | Existing semantic controls, focus, work-area geometry, density, and baselines remain unchanged | Core 03 plus design direction | layout/continuity/grid/surface components | Browser a11y, measurement, and visual retained evidence | Only if DOM/layout changes unexpectedly | High |

Zero-UX-delta covenant for every slice:

- Preserve route roots, WebSocket semantics, request fields, response meaning,
  stable IDs, selectors, copy, DOM semantics, and server authorization.
- Preserve grid-first row creation/editing, keyboard and paste behavior,
  selection/focus/scroll continuity, and default-closed inspector behavior.
- Preserve saved-view identity and query/layout configuration while excluding
  active cell, scroll, inspector, pending, conflict, and other presentation
  state from persistence.
- Preserve pending FIFO/replay and transaction-ID rules. Pending, same-field
  conflict, invalid draft, in-flight, and committed state remain distinct.
- Preserve collaboration replay, gap recovery, stale-event rejection, presence,
  evidence, closure/read-only, authorization-loss cleanup, density, and
  extension discovery/teardown.

Any desired behavior change becomes a separately authorized future task that
names its owner clause, rationale, migration impact, and new characterization.

## 7. Target architecture

### 7.1 Module catalog

| Target module | Public responsibility | Public facade | Inputs | Outputs/events | Private complexity hidden | State owned | Side effects owned | Allowed dependencies | Forbidden dependencies | Migration source files |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Existing shell/runtime | Compose owner snapshots and commands for one Workbook instance | WorkbookShell and useWorkbookShellRuntime | Auth/incident/API/extension inputs | Surface/layout/query/mutation owners | Startup and instance construction | Shell-scoped owner instances | Owner construction/disposal | Workbook facades and injected platform ports | Concrete surface internals, raw grid vendor | Retain |
| Existing surfaces facade | Select one registered surface renderer | WorkbookSurfacesFacade | Stable registration and owner inputs | Rendered active surface | Concrete imports/lazy feature selection | None | Lazy feature lifecycle only | Concrete owner surfaces privately | Shell back edge | Retain |
| Semantic mutation outcome adapter | Convert transport-shaped command results to owner-safe accepted/rejected outcomes | Owner-specific command port methods | Semantic command plus injected API/ID context | Accepted row/change-set identity or safe failure/conflict | Request envelope, response cast, error/conflict interpretation | No UI state | HTTP and transaction-ID creation | services, pure payload models, conflict parser | React presentation, shell facade | S-01/S-03/S-05/S-06/S-09 command adapters |
| Generic mutation controller | Coordinate generic mutation lifecycle and refresh using semantic outcomes | useGenericSurfaceMutationController | Generic port, mutation runtime, refresh ports | Saved/conflict/error state and accepted data | Explicit-mutation finish ordering and conflict registration | Generic mutation presentation state | Refresh/reference refresh | Runtime and semantic command port | services/workbookApi, raw payload | Existing generic controller and surface |
| Generic party-link workflow | Own create-party-then-link sequence | useGenericPartyLinkWorkflow | Selected row/pair, draft text, generic command port | Workflow status/error and completion | Normalization, party ID handoff, patch sequencing | Party-link form state | Two semantic commands | Generic model/controller ports | HTTP, grid vendor, broad shell | GenericWorkbookSurface |
| Entity merge workflow | Own confirmation, merge command, refresh, inspector retarget, and cleanup | useEntityMergeWorkflow | Selected survivor/loser, semantic entity port, refresh/inspector ports | Status/error/completion | Version admission and post-merge retarget | Merge form/confirmation state | Semantic merge and refresh | Entity model, narrow ports | HTTP, shell, grid vendor | EntityWorkbookSurface |
| Assessment creation workflow | Own assessment draft validation/submission/refresh lifecycle | useAssessmentCreateWorkflow | Draft, semantic assessment port, refresh | Status/error/completion | Payload admission and explicit mutation lifecycle | Assessment draft/submission state | Semantic create and refresh | Assessment model/runtime ports | Raw transport | AssessmentWorkbookSurface |
| Timeline query I/O | Execute and normalize Timeline row/history/mention reads | TimelineQueryPort | Query identity, continuation, abort generation | Typed rows/page/load failure | Routes, envelopes, abort/latest-request checks | Request generation only | GET requests | services and query models | React surface, pending queue | useTimelineRowsLoader and read-only action loaders |
| Timeline pending replay I/O | Submit one persisted logical action without owning queue policy | TimelinePendingReplayPort | Pending entry with stable client_txn_id | Semantic replay outcome | Wire request and response interpretation | None | Replay HTTP only | services, request models | Queue storage/policy, presentation | useTimelinePendingReplayController transport calls |
| Timeline action adapter | Execute non-query Timeline commands with semantic outcomes | TimelineActionPort | Stable record/field/version/action inputs | Accepted result, safe rejection, conflict/recovery event | Routes, payload/envelope/error parsing | No presentation state | Mutation HTTP and new/reused IDs per contract | services, mutation request models, secure ID port | React/DOM/grid internals | Direct-transport Timeline action hooks |
| Timeline collaboration binding | Adapt coordinator events to the active Timeline row/query owner | useTimelineCollaborationBindings | Coordinator, active-surface port, row admission and refresh ports | Applied patch/invalidation/presence snapshot | Subscription/registration lifecycle | Subscription handles only | Register/unregister and typed refresh | Coordinator and narrow row/query ports | Transport session internals, surface renderer | TimelineWorkbook collaboration effects |
| Timeline bulk-tag controller | Own bulk selection reconciliation and tag submission | useTimelineBulkTagController | Authoritative rows, semantic action port | Selection snapshot and action commands | Selection pruning, draft/submit lifecycle | Bulk selection/tag draft/status | Semantic bulk command | Timeline row model/action port | Raw HTTP, DOM refs | TimelineWorkbook bulk state/effects |
| Timeline editor-draft registry | Own invalid scalar drafts and editor input refs by semantic cell identity | useTimelineEditorDraftRegistry | record_id/field_key anchors and validation outcomes | Draft/ref commands and snapshot | Semantic keying, prune/reset rules | Invalid drafts and input refs | Focus restoration through continuity port | Continuity/grid semantic types | Vendor coordinates, HTTP | TimelineWorkbook scalar maps/refs |
| Timeline row-mutation coordinator | Admit mutations against authoritative versions and coordinate refresh/continuity | useTimelineRowMutationCoordinator | Rows/high-water, pending/conflict/runtime/action/continuity ports | Commit/pending/conflict outcome | Freshness admission and post-mutation ordering | In-flight logical action bookkeeping | Semantic command dispatch and refresh request | Pure freshness, runtime, narrow ports | Raw transport, presentation tree | TimelineWorkbook mutation callbacks |
| Existing layout/continuity/inspector/collaboration owners | Preserve already-closed design decisions | Current public facades/ports | Semantic owner inputs | Stable snapshots/commands | Vendor/layout/lifecycle details | Existing instance-scoped state | Existing owned side effects | Current allowed dependencies | New broad facades or peer internals | Retain; consumers become thinner |

### 7.2 Dependency rules

The migration direction is always presentation -> owner controller -> semantic
port -> adapter -> services. Accepted data and owner-safe failures travel back
up; raw Response, status/payload envelopes, route assembly, and generated wire
objects do not. Controllers communicate with collaboration, continuity,
inspector, query, and mutation runtime through narrow ports rather than a
WorkbookShell or surface facade.

## 8. Architectural guardrails

| # | Rule | Current violations/files | First slice? | Enforcement | Validation | Allowlist/removal |
| ---: | --- | --- | --- | --- | --- | --- |
| 1 | Direct grid-vendor imports remain in grid-adapter | None | Preserve | Existing package/vendor boundary | make frontend-import-boundary-check | None |
| 2 | Application identity uses semantic IDs, not positions/labels/vendor coordinates | No proven violation; Timeline composition remains sensitive | Preserve | Existing grid/continuity rules and stable assertion review | Boundary plus Workbook/grid tests | None |
| 3 | UI components do not parse wire contracts ad hoc | Generic/Entity/Assessment surfaces and several Timeline hooks interpret envelopes/errors | Yes: generic surface/controller | Exact no-services rule plus semantic outcomes | Boundary, typecheck, unit | Remove each recorded caller in its owning slice; no blanket allowlist |
| 4 | Generated files are never hand-edited | None | Preserve | Generated artifact policy | make generated-artifact-policy-check when applicable | None |
| 5 | Modules do not import peer internals | No current cycle; private surface/layout rules pass | Preserve | Existing exact-file/directory boundary rules | make frontend-import-boundary-check | None |
| 6 | Shared coordinators use narrow typed ports | Timeline composition passes broad local callbacks, but coordinator core is clean | No | Add consumer-port boundary in S-10 | Boundary and collaboration tests | None |
| 7 | Domain behavior stays out of transport/grid/presentation utilities | Raw interpretation and workflow sequencing remain in presentation | Yes | Semantic adapter/controller ownership rules | Boundary and owner tests | No compatibility wrapper |
| 8 | Browser controls are not authorization authority | None | Preserve | Auth/access-loss tests and review | App/Workbook owner tests | None |
| 9 | Sensitive state clears on session/capability/incident-access loss | Existing typed invalidation passes; new controllers do not yet exist | Preserve | Require invalidation/dispose port for every stateful extraction | Owner unit plus access-loss tests | None |
| 10 | Routes and WebSocket semantics remain unchanged | None | Preserve | Request characterization; no contract edits | Payload/route/collaboration tests | None |
| 11 | Extensions use discovery/owner gates, not component presence | None | Preserve | Existing feature/availability boundary | Boundary and extension/surface tests | None |
| 12 | Saved views remain configurations, not custom sheets | None | Preserve | Existing saved-view model and owner clauses | Saved-view unit/shell tests | None |
| 13 | Presentation state is not persisted without owner permission | None | Preserve | State matrix review and saved-view assertions | Saved-view tests | None |
| 14 | Pending, conflict, and committed state remain distinct | No type collapse; Timeline wiring is broad | Preserve | Typed outcome variants and existing runtime types | Save-state/pending/conflict tests | None |
| 15 | Stable client_txn_id survives uncertain replay | No proven violation; direct Timeline actions are risky | Preserve | Semantic replay/action ports retain ID policy | Action sequencing/payload/replay tests | None |
| 16 | Tests use runtime/machine contracts, never normative prose | None | Preserve | Existing harness and no Markdown execution | make test-slice OWNER=web.workbook | None |
| 17 | Historical phase identity is not a production/accounting boundary | None | Preserve | Slice IDs exist only in this tracker | Import/source ownership review | None |
| 18 | Layout does not depend on row count or prohibited viewport arithmetic | None; prior layout slice closed it | Preserve | Existing layout policy boundary and measurement evidence | Boundary; measurement only if layout touched | None |

## 9. Ordered refactor roadmap

| Slice ID | One primary seam | Current paths | Target paths | Problem demonstrated | Target facade/direction | Behavior frozen | Characterization prerequisite | Guardrail | Validation after move | Rollback point | Dependencies | Stop condition | Explicit exclusions |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| F-01 | Strict Workbook protocol foundation | Authored protocol-ts operation/schema projections and Workbook service boundary | Generated operation bindings/validators plus private typed Workbook operation executor | In-scope operations are incompletely projected and readEnvelope is an unchecked cast | Workbook adapters -> typed executor -> generated operation binding -> browser transport | Existing HTTP methods, paths, queries, requests, success envelopes, and safe errors | Valid and malformed success/error fixtures for every added operation | Projected success operations must have generated validators; generated roots remain Make-owned | package.protocol_ts and web.architecture slices, generation/drift/policy/JSON, boundary, typecheck, build-web | Revert authored projections, generator/facade changes, generated output, and tests as one unit | P-00 | Every in-scope operation is generated and malformed success fails closed | No server contract or normative-owner change; no fallback decoder |
| S-01 | Generic mutation outcomes | Generic command port/adapter, generic controller/surface | Same owner paths plus shared semantic outcome contract where required | Raw ok/status/payload and readEnvelope cross into presentation | Generic surface -> controller -> GenericMutationCommandPort -> typed adapter | Exact valid requests, IDs, refresh order, save state, errors, conflict registration, party-link result | Command adapter and surface success/rejection/conflict/invalid-contract cases | Exact generic no-services import rule | Boundary, typecheck, frontend unit, build-web | Revert the port/adapter/consumer/rule/test diff as one unit | F-01 | Generic surface/controller contain no workbookApi import and raw result type is absent from their graph | No UI/DOM redesign; no compatibility decoder or raw-payload escape |
| S-02 | Generic party-link workflow | GenericWorkbookSurface party-link state/callbacks | workbook/generic/useGenericPartyLinkWorkflow.ts or equivalently owned deep module | Two-command create-party/link sequence shares presentation state | Surface -> party-link workflow -> semantic generic port | Validation text, created-party identity, patch order, refresh, focus/inspector behavior | Existing party flows plus failure between commands | Workflow module cannot import services/grid vendor/shell | Boundary, typecheck, web.workbook slice, build | Revert new controller and surface wiring | S-01 | Surface renders snapshot/commands; one workflow owns sequence | No party UX or endpoint change |
| S-03 | Entity mutation outcomes | Entity command port/adapter and EntityWorkbookSurface | Entity semantic outcome in mutation port/adapter | Surface parses accepted/error/conflict envelopes | Entity surface/controller -> entity port -> adapter | Create/patch/paste/merge requests and visible outcomes | Entity success/error/conflict/paste characterization | Entity presentation no transport helper import | Boundary, typecheck, entity/workbook unit, build | One entity outcome diff | S-01 pattern | No raw result reaches Entity presentation | No merge workflow extraction yet |
| S-04 | Entity merge workflow | EntityWorkbookSurface merge form/confirmation/callbacks | workbook/entity/useEntityMergeWorkflow.ts | Merge lifecycle, refresh, inspector retarget, and UI share owner | Surface -> merge workflow -> entity port/inspector/refresh ports | Versions, confirmation, reason, survivor/loser result, retarget, continuity | Merge success/rejection/stale/access-loss | Workflow uses narrow ports | Boundary, typecheck, entity/workbook tests, build | Revert controller adoption | S-03 | One merge owner; surface is presentation | No merge semantics/UI redesign |
| S-05 | Assessment creation coordination | AssessmentWorkbookSurface create state/submission | workbook/assessment/useAssessmentCreateWorkflow.ts | Draft/submission/error/refresh mixed with rendering | Surface -> assessment workflow -> semantic port | Required fields, payload, visible errors, refresh, inspector continuity | Assessment create success/invalid/rejected/access-loss | Workflow no raw service import | Boundary, typecheck, assessment/workbook unit, build | Revert controller adoption | S-01 pattern | Surface consumes snapshot/commands only | No assessment UX/schema change |
| S-06A | Evidence capability outcomes | EvidenceMutationCommandPort void/raw exception path, direct handle calls, upload detail parsing, and evidence binding | Semantic EvidenceCapabilityPort and adapter | Callers cannot distinguish owner-safe rejection without transport/exception knowledge and raw storage responses influence retry/error text | Evidence binding -> semantic evidence capability -> service adapter | Blob/upload/attach order, opaque access handles, safe messaging, refresh, bounded attempts | Attach/create/preview/download/blocked/access-loss and raw-detail suppression | Evidence presentation no service/transport parsing; opaque href remains opaque | Boundary, typecheck, evidence/workbook unit, build, targeted stateful browser | One evidence-capability diff | F-01 and S-01 pattern | Binding consumes semantic outcomes and raw upload body reaches no state/log/copy | No server upload or handle contract change; no storage-detail compatibility |
| S-06B | Coordination mutation outcomes | Coordination command raw results and CoordinationWorkflowBindings | Semantic coordination outcomes | Feature binding receives transport-shaped results | Binding -> coordination port -> adapter | Task lifecycle/supersede payloads, IDs, errors, refresh | Coordination success/rejection/conflict | Binding no service/envelope import | Boundary, typecheck, tasks/decisions/workbook unit, build | One coordination-port diff | S-01 pattern | No raw result above adapter | No workflow redesign |
| S-07 | Timeline query I/O | useTimelineRowsLoader and read-only Timeline loaders calling services | timeline/query/TimelineQueryPort and adapter | Query policy and raw routes/envelopes share hooks | Runtime hook -> typed query port -> services | Query parameters, abort/latest admission, pages, stale retention, errors | Query, rapid-change, access-loss, stale-generation cases | Timeline query consumers no services import | Boundary, typecheck, Timeline/workbook unit, build | Migrate one declared query family atomically | S-01 | Selected query hook has only typed I/O | No query behavior/pagination change |
| S-08 | Timeline pending replay I/O | useTimelinePendingReplayController raw submission and transport-shaped pending units | timeline/replay/TimelinePendingReplayPort plus semantic pending-intent union | Queue/retry policy shares transport interpretation and stores path/method | Replay controller -> replay port -> shared action executor | Capacity 64, FIFO, coalescing limits, same stable/transaction IDs, dispatch-time committed versions, recovery-ID rule, pending visibility | Transient/retry/uncertain/conflict/auth/requery/closed-incident cases | Replay policy and units contain no raw transport coordinate | Boundary, typecheck, pending/replay owner tests, stateful browser | Revert semantic-unit and port adoption together | S-07 | Controller owns policy only and pending units contain semantic intent | No persistence addition or server protocol change |
| S-09 | Timeline action outcomes | Direct transport in history, mentions, clipboard, mutation, related, evidence hooks | Narrow history, mention, clipboard, related-record, evidence, bulk, and hot-path ports/adapters | Multiple actions manufacture/interpret wire intent | Action hook -> owner-specific semantic port -> shared operation executor | Every route/payload/ID/error/conflict/refresh behavior | Per-action payload and visible-outcome matrix | Action hooks no services/workbookApi import; no broad TimelineActionPort | Boundary, typecheck, Timeline/workbook owners, build; targeted browser by action | One action family per atomic commit under the slice | S-07/S-08 | All named action hooks are transport-neutral | No consolidation of unrelated UI state or compatibility facade |
| S-10 | Timeline collaboration binding | TimelineWorkbook event patch/refresh/active-surface effects | timeline/collaboration/useTimelineCollaborationBindings.ts | Composition owns subscription/registration details | Timeline composition -> binding -> coordinator and narrow row/query ports | record_changed admission, stale rejection, presence, gap refresh, teardown | Live/stale/gap/access-loss/surface-switch cases | Binding cannot import session transport or surface renderer | Boundary, collaboration/Timeline/workbook tests, stateful browser | Revert binding extraction | S-07/S-09 | TimelineWorkbook has one binding call | No WebSocket/retry change |
| S-11 | Bulk selection/tag coordination | TimelineWorkbook bulk state/effects/submission/view-bar mapping | timeline/bulk/useTimelineBulkTagController.ts | Selection lifetime and command workflow share root rendering | View bar/grid -> bulk controller -> action port | Stable record selection, pruning, tag draft, payload, save state | Selection/query-refresh/tag success/rejection/conflict | Bulk owner no raw transport/DOM refs | Boundary, typecheck, Timeline/workbook tests, browser keyboard if touched | Revert controller adoption | S-09 | Root receives snapshot/commands only | No bulk UX or query semantics |
| S-12 | Scalar editor draft/ref ownership | TimelineWorkbook invalid draft maps and input refs | timeline/editing/useTimelineEditorDraftRegistry.ts | Invalid drafts, semantic keys, refs, and focus correction are root-owned | Cell editors -> registry -> continuity port | Invalid text retention, correction, cancel, focus, row/schema pruning | Invalid/valid/cancel/removal/refresh continuity cases | Registry uses semantic IDs and no vendor coordinates | Boundary, typecheck, Timeline/grid/continuity unit, a11y if focus risk | Revert registry adoption | S-10 | No scalar draft/ref map in root | No editor UI/validation change |
| S-13 | Row mutation admission/freshness | TimelineWorkbook mutation callbacks and high-water coordination | timeline/mutations/useTimelineRowMutationCoordinator.ts | Freshness, pending, conflicts, refresh, continuity, and dispatch converge in root | Root -> coordinator -> freshness/runtime/action/query/continuity ports | Base versions, stale rejection, committed/pending distinction, refresh/focus order | Accepted/stale/transient/conflict/live-race matrix | Coordinator uses narrow semantic ports | Boundary, typecheck, Timeline/workbook unit, stateful/browser continuity | Revert coordinator as one unit | S-08-S-12 | Root delegates mutation admission to one owner | No freshness algorithm or save-state redesign |
| S-14 | Timeline composition closure | Residual TimelineWorkbook wiring/rendering | TimelineWorkbook plus private facade modules established above | Root remains broader than presentation after owner extraction | TimelineWorkbook composes snapshots/commands and renders only | Entire zero-UX covenant | All preceding owners green; composition test | Exact no-services and private-module direction rules | Boundary, typecheck, all affected owner tests, build, browser/a11y/measurement only where touched | Revert final wiring/guard diff | S-07-S-13 | Root has no raw I/O, queue policy, conflict parsing, collaboration subscription, or draft registry | No visual component reorganization |

## 10. First implementation slice: S-01

### 10.1 Interface decision

Only GenericMutationCommandPort changes in S-01. Other command-port methods
keep WorkbookMutationCommandResult until their own slices.

~~~ts
type GenericMutationAccepted = {
  readonly kind: "accepted";
  readonly viewSchemaId: string;
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
};

type GenericMutationRejected = {
  readonly kind: "rejected";
  readonly message: string;
  readonly sameFieldConflict: WorkbookSameFieldConflictPayload | null;
};

type GenericMutationOutcome =
  | GenericMutationAccepted
  | GenericMutationRejected;
~~~

createRecord, patchRecord, and createPartyFromText return
Promise<GenericMutationOutcome>. canCreateRecord remains unchanged.
createWorkbookMutationCommandPorts continues to build the identical requests
and uses the existing response cast/error/conflict parsers inside the adapter.
S-01 deliberately does not add stricter runtime decoding because changing
malformed-upstream behavior would exceed a behavior-preserving move.

The generic controller:

- Accepts GenericMutationOutcome rather than unknown payload.
- Registers sameFieldConflict through WorkbookMutationRuntime using the
  existing focus key, row label, surface label, and view schema.
- Preserves begin/finish explicit-mutation ordering.
- Preserves onRefresh before refreshReferenceOptions.
- Returns accepted semantic data to the surface.
- Presents the same message for validation, rejection, and refresh failure.

GenericWorkbookSurface uses accepted.row.record_id for party creation and no
longer imports services/workbookApi. No public route, request, response, DOM,
copy, CSS, selector, saved-view, or grid behavior changes.

### 10.2 Checkpoint plan

| Checkpoint | Edit scope | Files expected to change | Characterization required first | Validation | Expected diff | Rollback point | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CP-01 Characterize adapter/outcomes | Test-only assertions for exact requests and current visible success/rejection/conflict outcomes | mutations/createWorkbookMutationCommandPorts.test.ts; WorkbookShell.surfaces.test.tsx | Existing generic create/patch/party-link cases inspected before edits | make frontend-unit | Assertions only; no source change | Revert CP-01 only if it contradicts an owner, otherwise stop BLOCKED | Success, rejection, same-field conflict, refresh sequence, and party ID handoff are frozen |
| CP-02 Define/normalize outcome | Generic port types and adapter conversion | mutations/workbookMutationCommandPorts.ts; mutations/createWorkbookMutationCommandPorts.ts; adapter test | CP-01 green | Typecheck and frontend unit | One owner-specific discriminated union; identical fetch inputs | Revert CP-02 while retaining characterization | Generic methods return no ok/status/payload envelope |
| CP-03 Adopt consumers | Controller and concrete generic surface consume outcome | hooks/useGenericSurfaceMutationController.ts; components/GenericWorkbookSurface.tsx; surface tests | CP-02 green | Typecheck and frontend unit | Remove two readEnvelope imports and generic unknown completion API | Revert CP-03 with CP-02 | Existing user-visible behavior passes; no raw payload in generic presentation |
| CP-04 Guard | Add exact forbidden-import rule for those two consumers | tools/frontend_import_boundaries.json | CP-03 green | make frontend-import-boundary-check | One machine rule; no temporary allowlist | Revert rule with S-01 | Rule fails on a deliberate local probe and passes after probe removal |
| CP-05 Complete | Review atomic diff and build | Tracker/evidence only | CP-01-04 green | make build-web; git diff --check; make agent-finalize before any broader check | Independently reviewable source/test/policy diff | Revert all S-01 edits as one unit | App builds; only named files changed; no generated/manifest/normative delta |

No new source or test file is planned, so
tools/frontend_source_ownership.json and authored test-family accounting remain
unchanged. If implementation proves a new file is necessary, stop S-01 and
re-plan its accounting rather than silently expanding the slice.

## 11. Characterization plan

| Behavior | Owner contract | Existing unit evidence | Existing integration evidence | Browser/E2E evidence | Visual/a11y evidence | Gap | Required pre-move characterization | Stable assertion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Workbook startup | Core 03 startup chain | Startup model/controller/admission tests | Workbook query/surface tests | Webserver/stateful retained evidence | Not required for S-01 | None for current plan | Reuse | Explicit/home/default/Timeline fallback by stable sheet/schema IDs |
| Surface switching | Core 03 stable surface identity | Surface registry/registration tests | WorkbookShell.surfaces | Browser retained evidence | Existing a11y | None for S-01 | Reuse | Stable sheet_ref/view_schema_id and teardown |
| Saved-view lifecycle | Core 03 saved views | Saved-view models/controller | Shell query/surface tests | Browser retained evidence | Not applicable | None | Reuse | Create/apply/update/duplicate/reset without presentation-state persistence |
| Query dispatch | Core 03 query | Query models and per-surface query hooks | Shell query/timeline query | Browser retained evidence | Not applicable | Timeline port not yet characterized by one adapter contract | Add exact query request/admission cases in S-07 | Stable filters/sort/group/page and stale-request rejection |
| Generic row creation | Core 03 grid-first creation | Generic model and command adapter tests | WorkbookShell.surfaces | Browser retained evidence | No layout change | Adapter tests do not yet assert semantic outcome | S-01 accepted data plus identical request | row record_id/schema/change-set reach caller; refresh order unchanged |
| Generic inline edit | Core 02/03 mutation/version | Generic model, conflict/runtime tests | Shell surfaces/payload/save-state | Browser retained evidence | Existing focus evidence | Raw result is current seam | S-01 success/rejection/same-field conflict | Same base version/change list and visible state |
| Generic party creation/link | Core 03 party/reference workflow | Generic model and command tests | Shell surfaces | Browser retained evidence | Existing focus evidence | Handoff from create result to patch must be explicit | S-01 party ID handoff; expanded in S-02 | Created record_id feeds exactly one link patch |
| Entity create/edit/paste | Core 03 owner surfaces | Entity model and command tests | Shell surfaces/payload | Browser retained evidence | Existing a11y | Outcome boundary gap | Add in S-03 | Same routes/payloads/errors/conflicts |
| Entity merge | Core 02/03 merge workflow | Entity model/conflict tests | Shell entity/inspector evidence | Browser retained evidence | Existing focus evidence | Workflow-owner cases need isolation | Add in S-04 | Versions, survivor/loser, confirmation, retarget |
| Assessment create | Core 03 assessment workflow | Assessment model | Shell assessments/surfaces | Browser retained evidence | Existing a11y | Controller lifecycle gap | Add in S-05 | Draft validation, payload, refresh, error |
| Keyboard navigation | Core 03 grid behavior | Workbook keyboard/grid adapter tests | Shell grid/sentinel | Browser retained evidence | A11y retained | No S-01 DOM change | Reuse; rerun only if unexpected DOM diff | Semantic anchor/focus, not vendor coordinates |
| Paste | Core 03 paste | Clipboard/model tests | Shell payload/surfaces | Browser retained evidence | Not applicable | Entity/Timeline I/O seam gaps | Add at S-03/S-09 | Same parsed intent, target field, payload, result |
| Pending replay | Core 03 pending/replay | Pending queue/runtime tests | Shell autosave/save-state | Stateful browser retained | Not applicable | Transport-neutral replay adapter missing | Add in S-08 | FIFO, stable ID, no committed-state masquerade |
| Same-field conflict | Core 03 conflict | Conflict model/runtime tests | Shell surfaces/inspector | Browser retained evidence | A11y retained | S-01 needs normalized rejection coverage | S-01 generic conflict outcome/registration | Same token/field/values/focus and explicit resolution |
| client_txn conflict | Core 03 ID recovery | Mutation runtime/command tests | Action sequencing/autosave | Stateful retained | Not applicable | Timeline action boundary cases | Add S-08/S-09 | Reuse ID on uncertainty; new ID only explicit recovery |
| Collaboration replay/gap/stale | Core 03 collaboration | Coordinator/messages tests | Shell collaboration/query | Stateful retained | Not applicable | Consumer binding extraction cases | Add S-10 binding teardown/application | High-water and typed refresh reason |
| Selection/focus/scroll continuity | Core 03 continuity | Continuity/viewport/sentinel tests | Shell grid/inspector/surfaces | Browser retained | A11y retained | Draft registry and mutation coordinator need owner cases | Add S-12/S-13 | Stable record_id/field_key anchor and connected opener |
| Inspector open/retarget/close | Core 03 REQ-03-291/292 | Inspector coordinator/model | Shell inspector/surfaces | Browser retained | A11y retained | Workflow extractions must prove retarget | Add per S-02/S-04/S-05 | Default closed, semantic retarget, close restores focus |
| Incident closure/read-only | Core 03 incident closure | Lifecycle/model tests | Shell surfaces/actions | Stateful retained | Existing visual state | New controllers must honor invalidation | Add per stateful extraction | No protected mutation; local state cleared |
| Authorization loss | Core 04 | App auth and query access-loss tests | Shell collaboration/surfaces | Stateful retained | Not applicable | Every new state owner needs cleanup case | Add per extraction | Cleared state cannot repopulate from stale async work |
| Layout/density | Core 03 REQ-03-286/289 | Density/responsive/layout policy | Shell grid/surfaces | Measurement retained | Visual/a11y retained | None; no planned layout work | Reuse; rerun only if composition/DOM changes | Row-count-independent geometry and semantic density |
| Extension omission/teardown | Extension NLSpec/Core 03 | Feature and registry tests | Shell surfaces | Browser retained | Existing a11y | None | Reuse for S-14 only | Availability gate, lazy mount, teardown on switch/access loss |

Tests assert observable behavior and stable semantic identifiers. They do not
assert private hook names, future directory paths, vendor coordinates, or
transient DOM structure.

## 12. Validation and artifact plan

### 12.1 Planning baseline

| Validation target | Exact command | Scope | Cost | Checkpoint | Expected artifact/output | Baseline result | Failure classification |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Task routing | make task-guide ROLE=module-author OWNER=web.workbook | Owner-aligned command discovery | Low | Planning | Focused and broader targets | PASS; focused target is make test-slice OWNER=web.workbook, broader make test-fast | Harness guidance |
| Import boundaries | make frontend-import-boundary-check | Production import graph and policy | Low | Planning and every slice | Tool summary/run root | PASS; .cartulary/test-results/20260731T122452Z-p1811643 | Architecture/import failure |
| TypeScript | make frontend-typecheck | Web workspace types | Low | Planning and every slice | Exit status/tool output | PASS; exit 0 | Type failure |
| Frontend unit | make frontend-unit | Frontend unit suite | Medium | Planning and affected slices | Tool summary/run root | PASS; .cartulary/test-results/20260731T123102Z-p1815927 | Product/test/harness failure |
| Web build | make build-web | Production web build | Medium | Slice completion | Build artifact/summary | Not rerun in planning; retained PASS at .cartulary/test-results/20260731T080434Z-p1641468 | Build/type/environment failure |
| Markdown | make lint-markdown | Tracker syntax/style | Low | Tracker completion | Tool summary/run root | PASS; final content rerun at .cartulary/test-results/20260731T132137Z-p1831013 | Documentation lint |
| Diff hygiene | git diff --check and git diff --no-index --check /dev/null docs/handoffs/web-workbook-modular-refactor-tracker.md | Tracked and new tracker whitespace | Low | Tracker completion | Empty diagnostic output | PASS; no whitespace diagnostics; no-index exit 1 only because the new file differs from /dev/null | Diff hygiene |

### 12.2 Future slice ladder

Use the narrowest owner row first:

1. make frontend-import-boundary-check.
2. make frontend-typecheck.
3. make frontend-unit or make test-slice OWNER=web.workbook with affected rows.
4. Affected subsystem owner rows for entity, assessment, evidence, coordination,
   Timeline, collaboration, or package boundaries.
5. make build-web.
6. A targeted browser/stateful test only when the slice moves interaction,
   focus, replay, collaboration, teardown, or access-loss coordination.
7. Browser a11y, measurement, or visual checks only when rendered composition,
   focus semantics, or layout can change.
8. make agent-finalize before a broader end-of-run check; then make test-fast or
   a broader target only when risk warrants it.

Classify failures as product behavior, type/lint, import boundary,
generated-artifact drift, test harness, browser infrastructure, fixture, or
environment/dependency. Record the run root and whether the failure is related
before broad reruns.

Read-only command discovery also ran successfully: make help and
make explain-target TARGET=frontend-import-boundary-check DETAIL=summary,
make explain-target TARGET=frontend-typecheck DETAIL=summary,
make explain-target TARGET=frontend-unit DETAIL=summary, and
make explain-target TARGET=build-web DETAIL=summary.

A read-only final structure audit compared Appendix A with
tools/frontend_source_ownership.json and reported 320 manifest paths, 320
ledger paths, zero missing, zero extra, and zero duplicate rows. The same audit
counted 18 runtime flows, 18 guardrails, 10 framework acceptance rows, and 12
web-specific acceptance rows.

## 13. Documentation and generated-artifact classification

| Slice | Documentation classification | Generated/accounting classification |
| --- | --- | --- |
| S-01 | Implementation-support tracker update only | No new/moved file; no generated, source-ownership, family, or topology change |
| S-02-S-06 | Implementation-support README/tracker only if a new owner module is added | Update authored source ownership for a new file; update test-family input only for a new/moved test; generate projections through Make, never hand-edit |
| S-07-S-14 | Implementation-support ownership guide/tracker for new deep modules | Same authored-input rule; generated topology only if its authored input changes |
| Any behavior change | Blocked in this program until separately authorized | Owner update -> typed projection -> generator -> implementation -> conformance/drift |

No normative owner update is required for the behavior-preserving roadmap. A
path move alone never justifies normative churn.

## 14. Top-level tracker

| ID | Task | Category | Status | Depends on | Evidence | Exit criterion |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target module and scope | scope | DONE (planning) | none | Section 1 | One Workbook composition seam and exclusions are explicit |
| T-002 | Inspect current repo state | discovery | DONE (planning) | T-001 | Sections 2-3 and Appendix A | All 320 paths, imports, packages, tests, manifests, and commands are accounted |
| T-003 | Map owner contracts | contracts | DONE (planning) | T-002 | Sections 1.2, 4-6 | Public behavior and authority sources are mapped |
| T-004 | Freeze characterization evidence | tests | DONE (planning) | T-003 | Sections 5, 10, and 11 | Existing evidence and pre-move gaps are named |
| T-005 | Plan boundary guardrails | architecture | DONE (planning) | T-003 | Sections 3, 7, and 8 | All 18 required guardrails have disposition/enforcement |
| T-006 | Plan behavior-preserving moves | implementation | DONE (planning) | T-004,T-005 | Sections 9-10 | Ordered one-seam slices and rollback points are complete |
| T-007 | Plan validation loop | validation | DONE (planning) | T-006 | Section 12 | Exact Make targets, costs, artifacts, and classifications are named |
| T-008 | Update docs/contracts if required | docs | DONE (planning) | T-003 | Section 13 | Tracker-only/current and future classification are explicit |
| T-009 | Execute or hand off | handoff | DONE (handoff only) | T-006,T-007,T-008 | Sections 10 and 17 | S-01 can become a goal without repository-wide rediscovery |

Implementation status for S-01 through S-14 remains NOT STARTED. DONE above
means the planning workflow is complete, not that the refactor was executed.

## 15. Applicable workflow map

| Workflow | Status | Exact inputs | Work completed | Evidence | Blockers | Next action | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 Session/source bootstrap | DONE (planning) | Git status/branch/commit/date, AGENTS, prompt, framework | Established tracker-only authority and historical disposition | Section 1 | None | Preserve baseline in S-01 | Scope and authority explicit |
| WF-01 Current-state scan | DONE (planning) | 320 source paths, manifests, imports, tests, Make targets | Reconciled tree and graph | Sections 2-3, Appendix A | None | Re-run changed-slice searches only | Exact current inventory recorded |
| WF-02 Module ownership inventory | DONE (planning) | Source responsibilities, state, callers, side effects | Assigned current/target owners without generic store | Sections 2, 4, 7 | None | Apply one owner per slice | Every state has lifetime and reset owner |
| WF-03 Public contract freeze | DONE (planning) | Core owners, NLSpecs, contracts, tests | Froze HTTP/WS/UI/storage/generated/harness/a11y behavior | Sections 5-6 | None | Recheck affected clauses per slice | Drift surfaces known |
| WF-04 Slice selection | DONE (planning) | Findings, ownership, freeze | Ordered S-01 through S-14 | Section 9 | None | Start S-01 CP-01 | Each slice is bounded/reversible |
| WF-05 Characterization plan | DONE (planning) | Existing unit/browser/visual evidence | Mapped evidence and gaps | Section 11 | None | Add S-01 cases before source change | Risky moves gated |
| WF-06 Boundary guardrail plan | DONE (planning) | Import manifest and policy tests | Dispositioned 18 rules | Section 8 | None | Add exact S-01 rule at CP-04 | Every rule has command/removal posture |
| WF-08 Frontend package seam plan | DONE (planning) | App/package/workbook graph | Retained valid package facades; defined app-local deep modules | Sections 2.3 and 7 | None | No package extraction in S-01 | Dependency direction complete |
| WF-09 Execution checkpoint plan | DONE (planning) | S-01 interface/tests/guard | Defined CP-01 through CP-05 | Section 10 | None | Execute CP-01 when separately authorized | Validation and rollback at each risk |
| WF-10 Validation/harness plan | DONE (planning) | Make guidance and retained artifacts | Named ladder and accounting rules | Section 12 | None | Run narrow checks per checkpoint | No invented direct command |
| WF-11 Docs/generated plan | DONE (planning) | Generated policy, source ownership, owner order | Classified every slice | Section 13 | None | Change authored inputs only if files/tests move | No generated hand edit |
| WF-12 Cleanup/anti-drift plan | DONE (planning) | Guardrails and rollback units | Required obsolete-import, raw-service, diff, and generated audits | Sections 8-13 | None | Close each rule in owning slice | No compatibility debris/allowlist |
| WF-13 Handoff/bootstrap | DONE (planning) | Complete tracker | S-01 goal-ready handoff | Sections 10 and 17 | None | Begin with safe restart commands | No repo-wide rediscovery |

WF-07 is not applicable: this frontend composition plan demonstrates no
backend facade requirement.

## 16. Risks, blockers, and decisions

### 16.1 Risks

| ID | Risk | Evidence | Likelihood | Impact | Affected slices | Mitigation | Detection |
| --- | --- | --- | --- | --- | --- | --- | --- |
| R-001 | Strict envelope validation can reject an obsolete or malformed server response that was previously cast as valid | readEnvelope is currently a cast; operation coverage is incomplete | Medium | High | F-01 and every I/O slice | Project existing OpenAPI owners, fail closed with safe invalid-contract state, and ship no dual decoder | Valid/malformed per-operation fixtures and synchronized contract drift checks |
| R-002 | Controller extraction can change refresh/reference-refresh/finish ordering | Generic controller and owner surfaces sequence callbacks explicitly | Medium | High | S-01-S-06/S-13 | Characterize call order before movement | Spy/order assertions and save-state tests |
| R-003 | New owners can outlive incident/auth/surface state | Existing shell owners use typed invalidation/dispose | Medium | High | Every stateful extraction | Instance-create controllers; typed reset/dispose; no module singleton | Two-instance, incident-switch, access-loss tests |
| R-004 | Pending/conflict/local draft can be collapsed into committed state | Timeline root coordinates several state partitions | Medium | Critical | S-08/S-12/S-13 | Keep discriminated types and explicit transitions | Save-state/conflict/pending assertions |
| R-005 | Transaction ID can be regenerated during uncertain replay | Direct Timeline actions and replay remain broad | Medium | Critical | S-08/S-09 | Port accepts persisted ID; recovery ID has explicit command | Payload/action-sequencing tests |
| R-006 | Selection/focus/scroll can regress after owner extraction | Semantic continuity exists but root wiring is dense | Medium | High | S-10-S-14 | Keep semantic anchors and current callback order | Unit plus targeted browser/a11y |
| R-007 | Added files can escape source/test accounting | Ownership manifest is exact at 320/320 | Medium | Medium | S-02-S-14 | Update authored accounting in same slice; generate through Make | Source-ownership and JSON/generation checks |
| R-008 | Broad Timeline work becomes unreviewable | 2,916 lines, 63 imports, many state/effect families | High | High | S-07-S-14 | Enforce one seam and rollback per slice | Diff/path review and stop condition |
| R-009 | Archived completed work is accidentally reopened | Prior tracker is large and phase-shaped | Low | Medium | All | Treat archive as evidence only; current source controls | Tracker decision D-001 and diff review |
| R-010 | Structural work hides an unrelated UI/UX change | Ultimate objective is future UX improvement, while safety and partial-completion feedback are authorized corrections | Medium | High | All | Preserve normative interaction/DOM behavior and isolate authorized invalid-contract or partial-completion feedback | DOM/copy/visual diff and affected tests |

### 16.2 Blockers

| ID | Blocker | Authority or missing evidence | Blocking workflow | Resolution condition | Owner |
| --- | --- | --- | --- | --- | --- |
| B-001 | None at planning completion | Current owners are consistent and S-01 has adequate evidence routes | none | Reopen only on an owner contradiction or failed CP-01 prerequisite | Future slice implementer |

### 16.3 Decisions

| ID | Decision | Alternatives considered | Evidence | Consequences | Revisit trigger |
| --- | --- | --- | --- | --- | --- |
| D-001 | Create a new current tracker; do not restore the archived completed tracker | Reopen archive; append residual work to it | c22ce208 archive move and live post-refactor source | Historical closure remains intact; residual plan is legible | Only if repository owners designate another canonical tracker |
| D-002 | Establish strict generated protocol coverage before normalizing Generic outcomes | Start with Generic casts; start with Timeline; extract global store | Typed outcomes without runtime validation would conceal the existing unchecked boundary | F-01 becomes the prerequisite pattern; S-01 remains the first owner migration | F-01 discovers a normative owner contradiction |
| D-003 | Keep outcome types owner-specific during migration | Convert every command port in one broad change | Different owners consume different success/failure meaning | Avoids breaking unrelated ports; more staged repetition | Repeated owner types become identical after all slices |
| D-004 | Reject malformed success responses at the F-01 adapter boundary | Preserve unchecked casts; add a fallback decoder | Existing generated protocol machinery already validates projected operations, and malformed data has no continuing compatibility value | Malformed or obsolete server payloads produce a safe invalid-contract outcome; no compatibility decoder remains | An adopted owner explicitly requires forward-tolerant handling for a named operation |
| D-005 | Retain existing shell/surfaces/layout/collaboration/continuity/inspector facades | Re-plan them because components remain large | Prior closure evidence and current passing boundaries | Work targets residual coupling only | New concrete violation is demonstrated |
| D-006 | Keep packages and adjacent features out of the roadmap | Move grid/view/UI contracts or Network Flow with Workbook | No cycle/vendor leak/material reverse dependency | Smaller scope and owner-aligned changes | A future slice proves inseparable package coupling |
| D-007 | No generic global store | Centralize state to reduce props | State lifetimes and invalidation differ materially | More explicit owner ports; safer teardown | Owner contracts establish a shared authoritative lifetime |
| D-008 | Saved views remain configurations; no custom sheets | Treat user views as new surfaces | Core 03 owner behavior | Refactor cannot create a new product concept | Separately authorized owner change |

## 17. Session handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-07-31T09:12:19-04:00 |
| Branch/commit | main at c22ce208a176c595f74acdae5b46b3a947b6fb72 |
| Tracker | docs/handoffs/web-workbook-modular-refactor-tracker.md |
| Target seam | Workbook application composition under apps/web/src |
| Planning workflows complete | WF-00-WF-06 and WF-08-WF-13 |
| Implementation performed | P-00 execution-control and fresh-baseline work only at this checkpoint |
| Primary files semantically inspected | WorkbookShell.tsx, useWorkbookShellRuntime.ts, WorkbookSurfacesFacade.tsx, mutation port/adapter, generic/entity/assessment surfaces, generic mutation controller, TimelineWorkbook.tsx and its runtime/action/replay/query hooks, collaboration coordinator/binding/port, layout/continuity/inspector owners |
| Machine-accounted files | All 320 .ts/.tsx paths in Appendix A |
| Passing execution baselines | frontend-import-boundary-check `.cartulary/test-results/20260731T134300Z-p1842232`; frontend-typecheck exit 0; web.workbook 112/112 `.cartulary/test-results/20260731T134300Z-p1842079`; build-web `.cartulary/test-results/20260731T134411Z-p1847868` |
| First implementation slice | F-01 strict Workbook protocol foundation |
| Unresolved questions | None |
| Blockers | None |
| Dirty-worktree disposition | Execution began with only this untracked tracker; preserve it and account every later change in Section 19 |
| Safe restart | git status --short; git rev-parse HEAD; inspect Section 19 ledger; resume only its first non-DONE row |
| Next workflow | Close P-00 tracker validation, then execute F-01 |

The remediation program is authorized through V-01. Section 19 is the current
restart and execution authority; this planning-era S-01 restart constraint is
retained only as historical context.

## Appendix A. Complete apps/web/src TypeScript ownership ledger

Every row below is a live, uniquely owned path. Import domains are direct
static imports summarized to stable package or apps/web/src ownership areas;
the complete graph remains machine-derived rather than encoded as a runtime
dependency. “Static inventory” is an honest inspection boundary, not a claim
that the file is dead or irrelevant.

| Path | Manifest owner | Role | Inspection | Current responsibility | Direct dependency domains | Planned disposition |
| --- | --- | --- | --- | --- | --- | --- |
| apps/web/src/app/AccountAdministrationPanels.tsx | web.app | production | static inventory | Account Administration Panels application composition, route, auth, or administration module | @cartulary/ui-contracts, external, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/AccountSettingsPanels.tsx | web.app | production | static inventory | Account Settings Panels application composition, route, auth, or administration module | @cartulary/ui-contracts, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/App.auth.support.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for App.auth.support | @cartulary/ui-contracts, external, testing, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/App.auth.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for App.auth | @cartulary/ui-contracts, external, services, testing, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/App.landing.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for App.landing | @cartulary/ui-contracts, external, react, testing, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/App.timeline-invalidation.support.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for App.timeline invalidation.support | @cartulary/ui-contracts, external, testing, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/App.tsx | web.app | production | targeted semantic read | App application composition, route, auth, or administration module | @cartulary/ui-contracts, react, services, shared, app (+1) | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/AppRoot.tsx | web.app | production | static inventory | App Root application composition, route, auth, or administration module | @cartulary/ui-contracts, react, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/AuthGateway.tsx | web.app | production | static inventory | Auth Gateway application composition, route, auth, or administration module | @cartulary/ui-contracts, external, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/DeploymentAuditPanel.tsx | web.app | production | static inventory | Deployment Audit Panel application composition, route, auth, or administration module | external, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/IncidentAdminPanel.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for Incident Admin Panel | @cartulary/ui-contracts, external, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/IncidentAdminPanel.tsx | web.app | production | static inventory | Incident Admin Panel application composition, route, auth, or administration module | @cartulary/ui-contracts, @cartulary/view-contracts, react, services, shared (+1) | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/IncidentImportPanel.tsx | web.app | production | static inventory | Incident Import Panel application composition, route, auth, or administration module | external, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/IncidentLanding.tsx | web.app | production | static inventory | Incident Landing application composition, route, auth, or administration module | @cartulary/ui-contracts, external, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/LandingAdminDisplay.tsx | web.app | production | static inventory | Landing Admin Display application composition, route, auth, or administration module | services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/LandingAdminLayout.tsx | web.app | production | static inventory | Landing Admin Layout application composition, route, auth, or administration module | @cartulary/ui-contracts, external, react, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/ReferencePackAdminPanel.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for Reference Pack Admin Panel | @cartulary/ui-contracts, external, testing, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/ReferencePackAdminPanel.tsx | web.app | production | static inventory | Reference Pack Admin Panel application composition, route, auth, or administration module | @cartulary/ui-contracts, external, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/api/appShellClient.routeBoundary.test.ts | web.app | test/support | static inventory | Characterization, policy, or harness evidence for app Shell Client.route Boundary | external, services, testing, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/api/appShellClient.ts | web.app | production | static inventory | app Shell Client application composition, route, auth, or administration module | @cartulary/protocol-ts, services, shared | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/api/publicHttpTypes.ts | web.app | production | static inventory | public Http Types application composition, route, auth, or administration module | @cartulary/protocol-ts | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/debug/AuthenticationDebugHarness.tsx | web.app | production | static inventory | Authentication Debug Harness application composition, route, auth, or administration module | react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/debug/DebugHarnessShell.tsx | web.app | production | static inventory | Debug Harness Shell application composition, route, auth, or administration module | @cartulary/ui-contracts, workbook/features, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/debug/IncidentDirectoryDebugHarness.tsx | web.app | production | static inventory | Incident Directory Debug Harness application composition, route, auth, or administration module | @cartulary/ui-contracts, react, services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/fontBundle.test.ts | web.app | test/support | static inventory | Characterization, policy, or harness evidence for font Bundle | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/fontRoles.test.tsx | web.app | test/support | static inventory | Characterization, policy, or harness evidence for font Roles | @cartulary/ui-contracts, external, testing, workbook/models, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/landingAdminStyles.ts | web.app | production | static inventory | landing Admin Styles application composition, route, auth, or administration module | react | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/landingAdminTypes.ts | web.app | production | static inventory | landing Admin Types application composition, route, auth, or administration module | @cartulary/ui-contracts, react, services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/otelBoundary.test.ts | web.app | test/support | static inventory | Characterization, policy, or harness evidence for otel Boundary | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/referencePackAdminClient.ts | web.app | production | static inventory | reference Pack Admin Client application composition, route, auth, or administration module | services, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/referencePackAdminModel.ts | web.app | production | static inventory | reference Pack Admin Model application composition, route, auth, or administration module | none | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/routeState.test.ts | web.app | test/support | static inventory | Characterization, policy, or harness evidence for route State | external, app | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/app/routeState.ts | web.app | production | static inventory | route State application composition, route, auth, or administration module | none | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/app/useAppRouteRuntime.ts | web.app | production | targeted semantic read | App Route Runtime application composition, route, auth, or administration module | react, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/collaboration/IncidentCollaborationSession.test.tsx | web.collaboration | test/support | static inventory | Characterization, policy, or harness evidence for Incident Collaboration Session | external, react, collaboration | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/collaboration/IncidentCollaborationSession.tsx | web.collaboration | production | static inventory | Incident Collaboration Session collaboration protocol or lifecycle owner | @cartulary/protocol-ts, react, shared | Retain closed collaboration owner; S-10 narrows Timeline consumption only |
| apps/web/src/extensions/ExtensionAvailabilityContext.tsx | web.extensions | production | static inventory | Extension Availability Context extension discovery, projection, or workspace module | react, extensions | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/extensions/extensionAvailability.test.ts | web.extensions | test/support | static inventory | Characterization, policy, or harness evidence for extension Availability | external, extensions | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/extensions/extensionAvailability.ts | web.extensions | production | static inventory | extension Availability extension discovery, projection, or workspace module | services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/extensions/extensionWorkspaceIdentities.ts | web.extensions | production | static inventory | extension Workspace Identities extension discovery, projection, or workspace module | shared | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/imports/importCoordinator.test.ts | web.imports | test/support | static inventory | Characterization, policy, or harness evidence for import Coordinator | external, testing, imports | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/imports/importCoordinator.ts | web.imports | production | static inventory | import Coordinator Import Assistant owner module | extensions, services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/main.tsx | web.entry | production | static inventory | main application module | external, app | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for Network Analysis Workspace | @cartulary/protocol-ts, @cartulary/ui-contracts, external, react, extensions (+2) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx | web.network_flow | production | static inventory | Network Analysis Workspace Network Flow feature module | @cartulary/grid-adapter, @cartulary/ui-contracts, external, react, collaboration (+3) | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/NetworkFlowGridLoadFixture.tsx | web.network_flow | production | static inventory | Network Flow Grid Load Fixture Network Flow feature module | @cartulary/grid-adapter, @cartulary/ui-contracts, react, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/NetworkFlowMappingModal.tsx | web.network_flow | production | static inventory | Network Flow Mapping Modal Network Flow feature module | @cartulary/ui-contracts, react, imports, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/NetworkFlowQueryControls.tsx | web.network_flow | production | static inventory | Network Flow Query Controls Network Flow feature module | @cartulary/ui-contracts, react, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/NetworkFlowSemanticGrid.test.tsx | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for Network Flow Semantic Grid | @cartulary/ui-contracts, external, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/NetworkFlowSemanticGrid.tsx | web.network_flow | production | static inventory | Network Flow Semantic Grid Network Flow feature module | @cartulary/grid-adapter, @cartulary/ui-contracts, react, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowBoundaryPolicy.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Boundary Policy | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowClient.ts | web.network_flow | production | static inventory | network Flow Client Network Flow feature module | extensions, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowCollaborationInterpreter.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Collaboration Interpreter | external, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowCollaborationInterpreter.ts | web.network_flow | production | static inventory | network Flow Collaboration Interpreter Network Flow feature module | collaboration, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowController.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Controller | external, services, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowController.ts | web.network_flow | production | static inventory | network Flow Controller Network Flow feature module | services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowErrors.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Errors | external, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowErrors.ts | web.network_flow | production | static inventory | network Flow Errors Network Flow feature module | services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowImportModel.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Import Model | external, imports, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowImportModel.ts | web.network_flow | production | static inventory | network Flow Import Model Network Flow feature module | imports, services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowIndicatorLinkModel.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Indicator Link Model | @cartulary/grid-adapter, external, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowIndicatorLinkModel.ts | web.network_flow | production | static inventory | network Flow Indicator Link Model Network Flow feature module | @cartulary/grid-adapter, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowPresentation.test.tsx | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Presentation | external, services, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowPresentation.tsx | web.network_flow | production | static inventory | network Flow Presentation Network Flow feature module | @cartulary/grid-adapter, @cartulary/ui-contracts, react, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/networkFlowQueryModel.test.ts | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Query Model | external, services, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/networkFlowQueryModel.ts | web.network_flow | production | static inventory | network Flow Query Model Network Flow feature module | services | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowCollaborationController.ts | web.network_flow | production | static inventory | Network Flow Collaboration Controller Network Flow feature module | react, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowExtensionEvents.test.tsx | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for Network Flow Extension Events | external, collaboration, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/useNetworkFlowExtensionEvents.ts | web.network_flow | production | static inventory | Network Flow Extension Events Network Flow feature module | react, collaboration, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowGraphController.ts | web.network_flow | production | static inventory | Network Flow Graph Controller Network Flow feature module | react, extensions, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowGridLayout.test.tsx | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for Network Flow Grid Layout | external, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/useNetworkFlowGridLayout.ts | web.network_flow | production | static inventory | Network Flow Grid Layout Network Flow feature module | react, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowImportController.ts | web.network_flow | production | static inventory | Network Flow Import Controller Network Flow feature module | react, extensions, imports, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowIndicatorLinkController.ts | web.network_flow | production | static inventory | Network Flow Indicator Link Controller Network Flow feature module | react, extensions, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowModalFocus.ts | web.network_flow | production | static inventory | Network Flow Modal Focus Network Flow feature module | @cartulary/ui-contracts, react | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowPagedQuery.test.tsx | web.network_flow | test/support | static inventory | Characterization, policy, or harness evidence for Network Flow Paged Query | external, networkFlow | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/networkFlow/useNetworkFlowPagedQuery.ts | web.network_flow | production | static inventory | Network Flow Paged Query Network Flow feature module | react, services, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowRejectedRowsController.ts | web.network_flow | production | static inventory | Network Flow Rejected Rows Controller Network Flow feature module | react, extensions, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowRowsController.ts | web.network_flow | production | static inventory | Network Flow Rows Controller Network Flow feature module | react, extensions, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/networkFlow/useNetworkFlowTableController.ts | web.network_flow | production | static inventory | Network Flow Table Controller Network Flow feature module | react, extensions, networkFlow | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/services/browserApi.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for browser Api | external, services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/browserApi.ts | web.services | production | static inventory | browser Api browser service or owner-scoped adapter | @cartulary/protocol-ts, shared, services | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/clientTransactionId.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for client Transaction Id | external, services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/clientTransactionId.ts | web.services | production | static inventory | client Transaction Id browser service or owner-scoped adapter | none | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/clientTransactionIdPolicy.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for client Transaction Id Policy | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/extensionContractAdapter.ts | web.services | production | static inventory | extension Contract Adapter browser service or owner-scoped adapter | @cartulary/protocol-ts | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/httpTransport.ts | web.services | production | targeted semantic read | http Transport browser service or owner-scoped adapter | @cartulary/protocol-ts | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/importContractAdapter.ts | web.services | production | static inventory | import Contract Adapter browser service or owner-scoped adapter | @cartulary/protocol-ts | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/importTargetContractAdapter.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for import Target Contract Adapter | external, services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/importTargetContractAdapter.ts | web.services | production | static inventory | import Target Contract Adapter browser service or owner-scoped adapter | @cartulary/protocol-ts | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/networkFlowContractAdapter.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for network Flow Contract Adapter | external, services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/networkFlowContractAdapter.ts | web.services | production | static inventory | network Flow Contract Adapter browser service or owner-scoped adapter | @cartulary/protocol-ts | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/workbookApi.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for workbook Api | external, services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/workbookApi.ts | web.services | production | targeted semantic read | workbook Api browser service or owner-scoped adapter | services | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/services/workbookEvidence.test.ts | web.services | test/support | static inventory | Characterization, policy, or harness evidence for workbook Evidence | external, services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/services/workbookEvidence.ts | web.services | production | static inventory | workbook Evidence browser service or owner-scoped adapter | @cartulary/protocol-ts, shared, services | Retain service boundary; future semantic adapters may depend on it |
| apps/web/src/shared/authorizationRecovery.ts | web.shared | production | static inventory | authorization Recovery shared application utility | shared | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/shared/publicError.ts | web.shared | production | targeted semantic read | public Error shared application utility | none | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/shared/workbookSheetRef.ts | web.shared | production | static inventory | workbook Sheet Ref shared application utility | none | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/shared/workbookShellContracts.ts | web.shared | production | static inventory | workbook Shell Contracts shared application utility | @cartulary/ui-contracts | Adjacent owner evidence; defer from this Workbook tracker |
| apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for Timeline Workbook Runtime Fixture | @cartulary/grid-adapter, @cartulary/view-contracts, react, app, workbook/collaboration (+6) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/appShellTestSupport.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for app Shell Test Support | app, testing | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/extensionAvailabilityTestSupport.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for extension Availability Test Support | extensions | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/fetchMockTestSupport.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for fetch Mock Test Support | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/selectorContractPolicy.test.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for selector Contract Policy | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/sourceOwnershipPolicy.test.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for source Ownership Policy | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/testSetup.dom.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for test Setup.dom | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/testSetup.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for test Setup | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for timeline Workbook Render Test Support | external, react, testing | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/timelineWorkbookTestSupport.test.tsx | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for timeline Workbook Test Support | @cartulary/ui-contracts, external, react, workbook/models, testing | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/timelineWorkbookTestSupport.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for timeline Workbook Test Support | @cartulary/ui-contracts, @cartulary/view-contracts, external, workbook/collaboration, workbook/models (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/transportBoundaryPolicy.test.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for transport Boundary Policy | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/workbookInspectorTestSupport.test.tsx | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for workbook Inspector Test Support | @cartulary/ui-contracts, external, react, workbook/models, testing | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/testing/workbookInspectorTestSupport.ts | web.testing | test/support | static inventory | Characterization, policy, or harness evidence for workbook Inspector Test Support | @cartulary/ui-contracts, external, testing | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.actionSequencing.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.action Sequencing | @cartulary/ui-contracts, external, testing, workbook/models, @cartulary/grid-adapter | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.assessments.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.assessments | @cartulary/ui-contracts, external, app, services, testing (+3) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.autosave.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.autosave | @cartulary/ui-contracts, external, testing, workbook/models, workbook/timeline (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.collaboration.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.collaboration | @cartulary/ui-contracts, external, testing, workbook/models, workbook/mutations (+3) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.evidence.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.evidence | @cartulary/ui-contracts, external, testing, workbook/models, @cartulary/grid-adapter | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.grid.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.grid | @cartulary/ui-contracts, @cartulary/view-contracts, external, testing, workbook/models (+2) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.gridProvenance.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.grid Provenance | @cartulary/ui-contracts, @cartulary/view-contracts, external, app, services (+5) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.history.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.history | @cartulary/ui-contracts, external, testing, workbook/models, workbook/timeline (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.inspector.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.inspector | @cartulary/ui-contracts, external, testing, workbook/models, workbook/timeline (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.mentionChips.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.mention Chips | external, workbook/timeline | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.payload.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.payload | @cartulary/ui-contracts, external, testing, workbook/models, workbook/timeline (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.query.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.query | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, external, workbook/components (+2) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.saveState.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.save State | @cartulary/ui-contracts, external, testing, workbook/models, @cartulary/grid-adapter | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.sentinel.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.sentinel | @cartulary/grid-adapter, @cartulary/ui-contracts, external, react, testing (+2) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.support.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.support | @cartulary/ui-contracts, external, react, testing, workbook/collaboration (+3) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.surfaces.test.tsx | web.workbook | test/support | targeted semantic read | Characterization, policy, or harness evidence for Workbook Shell.surfaces | @cartulary/ui-contracts, @cartulary/view-contracts, external, react, app (+5) | Retain and extend for S-01 characterization |
| apps/web/src/workbook/WorkbookShell.timelineQuery.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Shell.timeline Query | @cartulary/ui-contracts, @cartulary/view-contracts, external, testing, workbook/models (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/WorkbookShell.tsx | web.workbook | production | targeted semantic read | Workbook Shell application module | @cartulary/ui-contracts, @cartulary/view-contracts, react, collaboration, extensions (+14) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Collaboration Coordinator | external, collaboration, shared, workbook/runtime, workbook/collaboration | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts | web.workbook | production | targeted semantic read | Workbook Collaboration Coordinator collaboration protocol or lifecycle owner | collaboration, shared, workbook/lifecycle, workbook/runtime, workbook/utils (+1) | Retain closed collaboration owner; S-10 narrows Timeline consumption only |
| apps/web/src/workbook/collaboration/useWorkbookCollaborationCoordinator.ts | web.workbook | production | targeted semantic read | Workbook Collaboration Coordinator collaboration protocol or lifecycle owner | react, collaboration, shared, workbook/collaboration | Retain closed collaboration owner; S-10 narrows Timeline consumption only |
| apps/web/src/workbook/collaboration/workbookCollaborationMessages.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Collaboration Messages | external, workbook/models, workbook/collaboration | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/collaboration/workbookCollaborationMessages.ts | web.workbook | production | static inventory | workbook Collaboration Messages collaboration protocol or lifecycle owner | shared, workbook/utils | Retain closed collaboration owner; S-10 narrows Timeline consumption only |
| apps/web/src/workbook/collaboration/workbookSurfacePort.ts | web.workbook | production | targeted semantic read | workbook Surface Port collaboration protocol or lifecycle owner | shared, workbook/lifecycle, workbook/collaboration | Retain closed collaboration owner; S-10 narrows Timeline consumption only |
| apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx | web.workbook | production | static inventory | Active Surface Saved View Selector React presentation and composition | @cartulary/ui-contracts, external, react, workbook/layout, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx | web.workbook | production | targeted semantic read | Assessment Workbook Surface React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, external, react (+12) | S-05 assessment-create workflow source |
| apps/web/src/workbook/components/EntityWorkbookSurface.tsx | web.workbook | production | targeted semantic read | Entity Workbook Surface React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, external, react (+14) | S-03 outcome adoption, then S-04 merge-workflow source |
| apps/web/src/workbook/components/GenericMutationControl.tsx | web.workbook | production | static inventory | Generic Mutation Control React presentation and composition | @cartulary/view-contracts, react, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/GenericWorkbookSurface.tsx | web.workbook | production | targeted semantic read | Generic Workbook Surface React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, external, react (+14) | S-01 outcome adoption, then S-02 party-link workflow source |
| apps/web/src/workbook/components/IncidentControlsDrawer.tsx | web.workbook | production | static inventory | Incident Controls Drawer React presentation and composition | @cartulary/ui-contracts, external, react, shared | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/SystemViewSwitcher.tsx | web.workbook | production | static inventory | System View Switcher React presentation and composition | @cartulary/ui-contracts, react, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookConflictResolver.tsx | web.workbook | production | static inventory | Workbook Conflict Resolver React presentation and composition | @cartulary/ui-contracts, react, workbook/runtime | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookGridControls.tsx | web.workbook | production | static inventory | Workbook Grid Controls React presentation and composition | @cartulary/ui-contracts, @cartulary/view-contracts, external, react, workbook/layout (+2) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookGridEditorControl.tsx | web.workbook | production | static inventory | Workbook Grid Editor Control React presentation and composition | @cartulary/grid-adapter, @cartulary/view-contracts, react, workbook/models, workbook/components | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookInspectorFeatureGroups.tsx | web.workbook | production | static inventory | Workbook Inspector Feature Groups React presentation and composition | @cartulary/ui-contracts, @cartulary/view-contracts, react, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookPresenceMarkers.tsx | web.workbook | production | static inventory | Workbook Presence Markers React presentation and composition | @cartulary/ui-contracts, react, workbook/utils | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookRecordCandidatePicker.tsx | web.workbook | production | static inventory | Workbook Record Candidate Picker React presentation and composition | react | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookShellSlots.tsx | web.workbook | production | static inventory | Workbook Shell Slots React presentation and composition | @cartulary/ui-contracts, react | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookStatusStrip.tsx | web.workbook | production | static inventory | Workbook Status Strip React presentation and composition | @cartulary/ui-contracts, react, workbook/continuity, workbook/utils | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/components/WorkbookViewBar.tsx | web.workbook | production | static inventory | Workbook View Bar React presentation and composition | @cartulary/ui-contracts, external, react | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/continuity/gridViewportContinuity.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for grid Viewport Continuity | external, workbook/continuity | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/continuity/gridViewportContinuity.ts | web.workbook | production | static inventory | grid Viewport Continuity semantic selection, focus, or viewport continuity | none | Retain closed continuity owner; consume through port in S-12/S-13 |
| apps/web/src/workbook/continuity/useWorkbookGridContinuity.tsx | web.workbook | production | targeted semantic read | Workbook Grid Continuity semantic selection, focus, or viewport continuity | @cartulary/grid-adapter, @cartulary/ui-contracts, react, workbook/utils, workbook/continuity | Retain closed continuity owner; consume through port in S-12/S-13 |
| apps/web/src/workbook/continuity/workbookContinuityPort.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Continuity Port | external, workbook/continuity | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/continuity/workbookContinuityPort.ts | web.workbook | production | targeted semantic read | workbook Continuity Port semantic selection, focus, or viewport continuity | none | Retain closed continuity owner; consume through port in S-12/S-13 |
| apps/web/src/workbook/evidence/evidenceAccessPresentation.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for evidence Access Presentation | external, workbook/models, workbook/evidence | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/evidence/evidenceAccessPresentation.ts | web.workbook | production | static inventory | evidence Access Presentation application module | workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/features/ImportAssistantFeature.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Import Assistant Feature | external, imports, testing, workbook/features | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/features/ImportAssistantFeature.tsx | web.workbook | production | static inventory | Import Assistant Feature feature facade or owner binding | @cartulary/ui-contracts, @cartulary/view-contracts, react, extensions, imports (+2) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/features/NetworkFlowFeature.tsx | web.workbook | production | static inventory | Network Flow Feature feature facade or owner binding | react, extensions, networkFlow | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/features/coordination/CoordinationWorkflowBindings.tsx | web.workbook | production | targeted semantic read | Coordination Workflow Bindings feature facade or owner binding | @cartulary/ui-contracts, @cartulary/view-contracts, react, workbook/hooks, workbook/models (+3) | S-06B semantic coordination-outcome consumer |
| apps/web/src/workbook/features/evidence/useEvidenceWorkbookBindings.tsx | web.workbook | production | targeted semantic read | Evidence Workbook Bindings feature facade or owner binding | @cartulary/ui-contracts, react, services, workbook/evidence, workbook/hooks (+6) | S-06A semantic evidence-outcome consumer |
| apps/web/src/workbook/hooks/useAssessmentSupportCandidates.ts | web.workbook | production | static inventory | Assessment Support Candidates state and side-effect hook | react, services, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useEntityTimelinePreview.ts | web.workbook | production | static inventory | Entity Timeline Preview state and side-effect hook | react, services, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useGenericSurfaceMutationController.ts | web.workbook | production | targeted semantic read | Generic Surface Mutation Controller state and side-effect hook | react, services, workbook/models, workbook/mutations, workbook/query (+1) | S-01 semantic-outcome consumer |
| apps/web/src/workbook/hooks/useIncidentControlsDrawer.ts | web.workbook | production | static inventory | Incident Controls Drawer state and side-effect hook | @cartulary/ui-contracts, react, shared | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useOwnerReferenceOptions.ts | web.workbook | production | static inventory | Owner Reference Options state and side-effect hook | @cartulary/view-contracts, react, services, workbook/models, workbook/query (+1) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useWorkbookIncidentIdentity.ts | web.workbook | production | static inventory | Workbook Incident Identity state and side-effect hook | react, services, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useWorkbookPendingGridFocus.ts | web.workbook | production | static inventory | Workbook Pending Grid Focus state and side-effect hook | @cartulary/ui-contracts, react | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useWorkbookProjectionRefreshController.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Projection Refresh Controller | external, workbook/hooks | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/hooks/useWorkbookProjectionRefreshController.ts | web.workbook | production | static inventory | Workbook Projection Refresh Controller state and side-effect hook | react | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useWorkbookQueryController.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Query Controller | external, workbook/hooks | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/hooks/useWorkbookQueryController.ts | web.workbook | production | targeted semantic read | Workbook Query Controller state and side-effect hook | @cartulary/ui-contracts, @cartulary/view-contracts, react, workbook/models, workbook/view-state | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/hooks/useWorkbookSavedViewController.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Saved View Controller | @cartulary/view-contracts, external, testing, workbook/layout, workbook/models (+1) | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/hooks/useWorkbookSavedViewController.ts | web.workbook | production | targeted semantic read | Workbook Saved View Controller state and side-effect hook | @cartulary/view-contracts, react, services, workbook/models | Retain completed Workbook boundary; no re-refactor |
| apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts | web.workbook | production | targeted semantic read | Workbook Shell Runtime state and side-effect hook | react, extensions, workbook/layout, workbook/models, workbook/startup (+1) | Retain completed Workbook boundary; no re-refactor |
| apps/web/src/workbook/hooks/useWorkbookStartupController.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Startup Controller | external, react, workbook/hooks | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/hooks/useWorkbookStartupController.ts | web.workbook | production | targeted semantic read | Workbook Startup Controller state and side-effect hook | react, services, workbook/models | Retain completed Workbook boundary; no re-refactor |
| apps/web/src/workbook/inspector/useWorkbookInspectorCoordinator.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Inspector Coordinator | @cartulary/view-contracts, external, workbook/models, workbook/inspector | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/inspector/useWorkbookInspectorCoordinator.ts | web.workbook | production | targeted semantic read | Workbook Inspector Coordinator inspector lifecycle coordination | @cartulary/view-contracts, react, workbook/models | Retain closed inspector owner; workflow slices use narrow commands |
| apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx | web.workbook | production | static inventory | Workbook Surface Layout layout, density, or responsive ownership | react, workbook/components, workbook/utils, workbook/layout | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/useWorkbookColumnLayoutController.ts | web.workbook | production | static inventory | Workbook Column Layout Controller layout, density, or responsive ownership | @cartulary/view-contracts, react, workbook/models, workbook/layout | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/useWorkbookLayoutFacade.ts | web.workbook | production | targeted semantic read | Workbook Layout Facade layout, density, or responsive ownership | @cartulary/grid-adapter, workbook/layout | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/useWorkbookResponsiveLayout.ts | web.workbook | production | static inventory | Workbook Responsive Layout layout, density, or responsive ownership | react, workbook/layout | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/workbookColumnLayout.ts | web.workbook | production | static inventory | workbook Column Layout layout, density, or responsive ownership | @cartulary/grid-adapter, @cartulary/view-contracts, workbook/models | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/workbookDensity.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Density | external, workbook/models, workbook/layout | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/layout/workbookDensity.ts | web.workbook | production | static inventory | workbook Density layout, density, or responsive ownership | @cartulary/grid-adapter, shared, workbook/models | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/workbookLayoutPolicy.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Layout Policy | external | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/layout/workbookResponsiveLayout.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Responsive Layout | external, workbook/layout | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/layout/workbookResponsiveLayout.ts | web.workbook | production | static inventory | workbook Responsive Layout layout, density, or responsive ownership | none | Retain closed layout owner; no planned change |
| apps/web/src/workbook/layout/workbookShellStyles.ts | web.workbook | production | static inventory | workbook Shell Styles layout, density, or responsive ownership | none | Retain closed layout owner; no planned change |
| apps/web/src/workbook/lifecycle/workbookInvalidation.ts | web.workbook | production | targeted semantic read | workbook Invalidation typed lifecycle and invalidation contract | shared | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/assessmentWorkbookModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for assessment Workbook Model | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/assessmentWorkbookModel.ts | web.workbook | production | static inventory | assessment Workbook Model pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/query, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/entityWorkbookModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for entity Workbook Model | @cartulary/view-contracts, external, workbook/query, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/entityWorkbookModel.ts | web.workbook | production | static inventory | entity Workbook Model pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/query, workbook/utils | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/evidenceLifecycleViewModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for evidence Lifecycle View Model | external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/evidenceLifecycleViewModel.ts | web.workbook | production | static inventory | evidence Lifecycle View Model pure model, projection, or transformation rules | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/genericWorkbookModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for generic Workbook Model | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/genericWorkbookModel.ts | web.workbook | production | static inventory | generic Workbook Model pure model, projection, or transformation rules | @cartulary/ui-contracts, @cartulary/view-contracts, workbook/query, workbook/utils, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookContractRows.ts | web.workbook | production | static inventory | workbook Contract Rows pure model, projection, or transformation rules | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookGridState.ts | web.workbook | production | static inventory | workbook Grid State pure model, projection, or transformation rules | @cartulary/grid-adapter, shared, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookIncidentIdentity.ts | web.workbook | production | static inventory | workbook Incident Identity pure model, projection, or transformation rules | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookInspectorModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Inspector Model | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookInspectorModel.ts | web.workbook | production | static inventory | workbook Inspector Model pure model, projection, or transformation rules | @cartulary/view-contracts | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookQuery.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Query | @cartulary/grid-adapter, @cartulary/view-contracts, external, workbook/layout, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookQuery.ts | web.workbook | production | static inventory | workbook Query pure model, projection, or transformation rules | @cartulary/view-contracts | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookReferenceOptions.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Reference Options | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookReferenceOptions.ts | web.workbook | production | static inventory | workbook Reference Options pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookSavedViewRuntime.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Saved View Runtime | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookSavedViewRuntime.ts | web.workbook | production | static inventory | workbook Saved View Runtime pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookSavedViews.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Saved Views | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookSavedViews.ts | web.workbook | production | static inventory | workbook Saved Views pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookStartup.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Startup | external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookStartup.ts | web.workbook | production | static inventory | workbook Startup pure model, projection, or transformation rules | shared, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookSurfaceQueryRuntime.ts | web.workbook | production | static inventory | workbook Surface Query Runtime pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/models | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookSurfaceRegistration.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Surface Registration | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookSurfaceRegistration.ts | web.workbook | production | static inventory | workbook Surface Registration pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/models/workbookSurfaceRegistry.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Surface Registry | @cartulary/view-contracts, external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/models/workbookSurfaceRegistry.ts | web.workbook | production | static inventory | workbook Surface Registry pure model, projection, or transformation rules | @cartulary/ui-contracts, @cartulary/view-contracts | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.test.ts | web.workbook | test/support | targeted semantic read | Characterization, policy, or harness evidence for create Workbook Mutation Command Ports | external, workbook/mutations | Retain and extend for S-01 characterization |
| apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts | web.workbook | production | targeted semantic read | create Workbook Mutation Command Ports semantic mutation port, adapter, or ID owner | @cartulary/view-contracts, services, workbook/models, workbook/timeline, workbook/mutations | S-01, then owner-specific S-03/S-05/S-06/S-09 extensions |
| apps/web/src/workbook/mutations/secureTransactionId.ts | web.workbook | production | static inventory | secure Transaction Id semantic mutation port, adapter, or ID owner | services | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/mutations/workbookConflictResolutionAdapter.ts | web.workbook | production | targeted semantic read | Generated-operation conflict-resolution adapter | workbook/mutations | S-09 validated semantic conflict-resolution callback |
| apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts | web.workbook | production | targeted semantic read | workbook Mutation Command Ports semantic mutation port, adapter, or ID owner | @cartulary/view-contracts, workbook/models, workbook/timeline | S-01, then owner-specific S-03/S-05/S-06/S-09 extensions |
| apps/web/src/workbook/policies/artifactSurfacePolicies.ts | web.workbook | production | static inventory | artifact Surface Policies surface ownership policy | workbook/models, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/policies/assessmentSurfacePolicies.ts | web.workbook | production | static inventory | assessment Surface Policies surface ownership policy | workbook/models, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/policies/captureTimelineSurfacePolicies.ts | web.workbook | production | static inventory | capture Timeline Surface Policies surface ownership policy | workbook/models, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/policies/coordinationSurfacePolicies.ts | web.workbook | production | static inventory | coordination Surface Policies surface ownership policy | workbook/models, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/policies/entitiesObservationsSurfacePolicies.ts | web.workbook | production | static inventory | entities Observations Surface Policies surface ownership policy | workbook/models, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/policies/evidenceSurfacePolicies.ts | web.workbook | production | static inventory | evidence Surface Policies surface ownership policy | workbook/models, workbook/policies | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/policies/workbookSurfaceOwnershipPolicy.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Surface Ownership Policy | external, workbook/models | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/policies/workbookSurfacePolicy.ts | web.workbook | production | static inventory | workbook Surface Policy surface ownership policy | @cartulary/view-contracts | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/query/WorkbookQueryRow.ts | web.workbook | production | static inventory | Workbook Query Row query owner, row contract, or patch admission | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/query/useAssessmentSurfaceQuery.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Assessment Surface Query | @cartulary/view-contracts, external, testing, workbook/models, workbook/query | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/query/useAssessmentSurfaceQuery.ts | web.workbook | production | targeted semantic read | Assessment Surface Query query owner, row contract, or patch admission | @cartulary/view-contracts, react, services, workbook/collaboration, workbook/lifecycle (+2) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/query/useEntitySurfaceQuery.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Entity Surface Query | @cartulary/view-contracts, external, testing, workbook/models, workbook/query | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/query/useEntitySurfaceQuery.ts | web.workbook | production | targeted semantic read | Entity Surface Query query owner, row contract, or patch admission | @cartulary/view-contracts, react, services, workbook/collaboration, workbook/lifecycle (+3) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/query/useGenericSurfaceQuery.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Generic Surface Query | @cartulary/view-contracts, external, testing, workbook/models, workbook/query | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/query/useGenericSurfaceQuery.ts | web.workbook | production | targeted semantic read | Generic Surface Query query owner, row contract, or patch admission | @cartulary/view-contracts, react, services, workbook/collaboration, workbook/lifecycle (+2) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/query/workbookQueryRowPatch.ts | web.workbook | production | static inventory | workbook Query Row Patch query owner, row contract, or patch admission | @cartulary/view-contracts, workbook/query | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/runtime/WorkbookMutationRuntime.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Mutation Runtime | external, workbook/runtime | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts | web.workbook | production | targeted semantic read | Workbook Mutation Runtime Workbook mutation or lifecycle runtime | @cartulary/grid-adapter, services, workbook/lifecycle, workbook/mutations, workbook/utils (+1) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/runtime/useWorkbookMutationRuntime.ts | web.workbook | production | static inventory | Workbook Mutation Runtime Workbook mutation or lifecycle runtime | react, workbook/runtime | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/runtime/workbookConflictModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Conflict Model | external, workbook/runtime | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/runtime/workbookConflictModel.ts | web.workbook | production | targeted semantic read | workbook Conflict Model Workbook mutation or lifecycle runtime | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/runtime/workbookLifecycleModel.ts | web.workbook | production | static inventory | workbook Lifecycle Model Workbook mutation or lifecycle runtime | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/runtime/workbookPendingReplayRuntime.ts | web.workbook | production | targeted semantic read | workbook Pending Replay Runtime Workbook mutation or lifecycle runtime | workbook/utils | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/services/referenceQueryBroker.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for reference Query Broker | external, workbook/services | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/services/referenceQueryBroker.ts | web.workbook | production | targeted semantic read | reference Query Broker browser service or owner-scoped adapter | @cartulary/view-contracts, services, workbook/lifecycle, workbook/models, workbook/query | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/startup/useWorkbookStartupAdmission.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Workbook Startup Admission | external, extensions, workbook/models, workbook/startup | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/startup/useWorkbookStartupAdmission.ts | web.workbook | production | targeted semantic read | Workbook Startup Admission Workbook startup admission owner | react, extensions, services, workbook/hooks, workbook/models | Retain completed Workbook boundary; no re-refactor |
| apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx | web.workbook | production | targeted semantic read | Workbook Surfaces Facade active-surface composition facade | @cartulary/view-contracts, react, shared, workbook/collaboration, workbook/components (+7) | Retain completed Workbook boundary; no re-refactor |
| apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts | web.workbook | test/support | targeted semantic read | Strict history, record-action, mention, clipboard, and Evidence attachment adapter characterization | external, testing, workbook/models, workbook/timeline | S-09 request/outcome/correlation/security evidence |
| apps/web/src/workbook/timeline/adapters/createTimelineClipboardPasteAdapter.ts | web.workbook | production | targeted semantic read | Timeline clipboard protocol adapter | @cartulary/protocol-ts, workbook/models, workbook/mutations, workbook/runtime, workbook/timeline | S-09 semantic clipboard boundary |
| apps/web/src/workbook/timeline/adapters/createTimelineEvidenceAttachmentAdapter.ts | web.workbook | production | targeted semantic read | Timeline Evidence attachment protocol adapter | @cartulary/protocol-ts, services, workbook/models, workbook/mutations, workbook/timeline | S-09 semantic attach boundary with stable row retry identity |
| apps/web/src/workbook/timeline/adapters/createTimelineHistoryAdapter.ts | web.workbook | production | targeted semantic read | Timeline history protocol adapter | @cartulary/protocol-ts, workbook/mutations, workbook/timeline | S-09 semantic history boundary |
| apps/web/src/workbook/timeline/adapters/createTimelineMentionAdapter.ts | web.workbook | production | targeted semantic read | Timeline mention protocol adapter | @cartulary/protocol-ts, workbook/collaboration, workbook/models, workbook/mutations, workbook/timeline | S-09 semantic mention boundary |
| apps/web/src/workbook/timeline/adapters/createTimelinePendingMutationAdapter.test.ts | web.workbook | test/support | targeted semantic read | Strict generated pending-mutation adapter characterization | external, testing, workbook/models, workbook/timeline, workbook/utils | S-08 request materialization, transport observation, and fail-closed evidence |
| apps/web/src/workbook/timeline/adapters/createTimelinePendingMutationAdapter.ts | web.workbook | production | targeted semantic read | Timeline pending-mutation protocol adapter | @cartulary/protocol-ts, workbook/models, workbook/mutations, workbook/timeline | S-08 shared fresh/replay execution boundary |
| apps/web/src/workbook/timeline/adapters/createTimelineRecordActionAdapter.ts | web.workbook | production | targeted semantic read | Timeline review/supersede protocol adapter | workbook/mutations, workbook/timeline | S-09 semantic record-action boundary |
| apps/web/src/workbook/timeline/adapters/createTimelineViewQueryAdapter.test.ts | web.workbook | test/support | targeted semantic read | Strict generated-query adapter characterization | @cartulary/view-contracts, external, testing, workbook/models, workbook/timeline | S-07 fail-closed query and abort evidence |
| apps/web/src/workbook/timeline/adapters/createTimelineViewQueryAdapter.ts | web.workbook | production | targeted semantic read | Timeline query protocol adapter | @cartulary/protocol-ts, @cartulary/view-contracts, workbook/models, workbook/mutations, workbook/timeline | S-07 semantic query adapter |
| apps/web/src/workbook/timeline/bulk/useTimelineBulkTagController.test.tsx | web.workbook | test/support | targeted semantic read | Timeline bulk-tag lifecycle and selection characterization | external, testing, workbook/models, workbook/mutations, workbook/timeline | S-11 stable-ID/pruning/outcome/access evidence |
| apps/web/src/workbook/timeline/bulk/useTimelineBulkTagController.ts | web.workbook | production | targeted semantic read | Timeline bulk-tag workflow controller | @cartulary/grid-adapter, react, workbook/mutations, workbook/timeline | S-11 single selection/draft/submission owner |
| apps/web/src/workbook/timeline/collaboration/useTimelineCollaborationBindings.test.tsx | web.workbook | test/support | targeted semantic read | Timeline collaboration binding lifecycle characterization | external, testing, workbook/collaboration, workbook/models, workbook/timeline | S-10 live/stale/gap/access/surface/teardown evidence |
| apps/web/src/workbook/timeline/collaboration/useTimelineCollaborationBindings.ts | web.workbook | production | targeted semantic read | Timeline active-surface collaboration binding | react, shared, workbook/collaboration, workbook/lifecycle, workbook/models (+1) | S-10 single coordinator-consumer binding |
| apps/web/src/workbook/timeline/components/TimelineCellEditors.tsx | web.workbook | production | targeted semantic read | Timeline Cell Editors React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, external, react, workbook/utils (+1) | S-12 editor-registry consumer; preserve rendering |
| apps/web/src/workbook/timeline/components/TimelineEvidencePanel.test.tsx | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Timeline Evidence Panel | @cartulary/ui-contracts, external, workbook/timeline | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/timeline/components/TimelineEvidencePanel.tsx | web.workbook | production | static inventory | Timeline Evidence Panel React presentation and composition | @cartulary/ui-contracts, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineGridSurface.tsx | web.workbook | production | static inventory | Timeline Grid Surface React presentation and composition | @cartulary/grid-adapter, react, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineHistoryPanel.tsx | web.workbook | production | static inventory | Timeline History Panel React presentation and composition | @cartulary/ui-contracts | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx | web.workbook | production | static inventory | Timeline Mentions Panel React presentation and composition | @cartulary/ui-contracts, react, workbook/collaboration, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineRowActions.tsx | web.workbook | production | static inventory | Timeline Row Actions React presentation and composition | @cartulary/ui-contracts, react, workbook/layout, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx | web.workbook | production | targeted semantic read | Timeline Workbook React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, react, external (+13) | S-07-S-14 staged composition source |
| apps/web/src/workbook/timeline/components/TimelineWorkbookGrid.tsx | web.workbook | production | static inventory | Timeline Workbook Grid React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, react, workbook/models (+2) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx | web.workbook | production | static inventory | Timeline Workbook Inspector React presentation and composition | @cartulary/ui-contracts, @cartulary/view-contracts, external, react, workbook/collaboration (+3) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineWorkbookInspectorSections.tsx | web.workbook | production | static inventory | Timeline Workbook Inspector Sections React presentation and composition | @cartulary/ui-contracts, react, workbook/components, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineWorkbookNotices.tsx | web.workbook | production | static inventory | Timeline Workbook Notices React presentation and composition | @cartulary/ui-contracts, react, workbook/runtime, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineWorkbookRenderers.tsx | web.workbook | production | static inventory | Timeline Workbook Renderers React presentation and composition | @cartulary/grid-adapter, @cartulary/ui-contracts, @cartulary/view-contracts, react, workbook/components (+3) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/components/TimelineWorkbookStyles.ts | web.workbook | production | static inventory | Timeline Workbook Styles React presentation and composition | react, workbook/layout | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.test.tsx | web.workbook | test/support | targeted semantic read | Timeline invalid-draft, input-reference, row-removal, and schema-invalidation characterization | @cartulary/view-contracts, external, testing, workbook/models, workbook/timeline | S-12 semantic lifetime and focus evidence |
| apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.ts | web.workbook | production | targeted semantic read | Timeline scalar draft and semantic editor-reference lifetime owner | react, workbook/timeline | S-12 single registry; no grid-vendor coordinates or root-owned maps |
| apps/web/src/workbook/timeline/hooks/useTimelineClipboardPasteController.ts | web.workbook | production | targeted semantic read | Timeline Clipboard Paste Controller state and side-effect hook | @cartulary/grid-adapter, react, services, workbook/models, workbook/runtime (+2) | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelineCommittedRows.ts | web.workbook | production | targeted semantic read | Timeline Committed Rows state and side-effect hook | react, workbook/timeline | S-10/S-13 narrow row/freshness port evidence |
| apps/web/src/workbook/timeline/hooks/useTimelineConflictProjectionAdapter.ts | web.workbook | production | targeted semantic read | Timeline Conflict Projection Adapter state and side-effect hook | react, workbook/runtime, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineConflicts.ts | web.workbook | production | static inventory | Timeline Conflicts state and side-effect hook | react, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts | web.workbook | production | targeted semantic read | Timeline Create Related Workflow state and side-effect hook | @cartulary/view-contracts, react, services, workbook/models, workbook/mutations (+2) | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelineEvidenceActions.ts | web.workbook | production | targeted semantic read | Timeline Evidence Actions state and side-effect hook | react | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelineEvidenceAttach.ts | web.workbook | production | targeted semantic read | Timeline Evidence Attach state and side-effect hook | react, services, workbook/models, workbook/timeline | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelineGridAnchorController.ts | web.workbook | production | static inventory | Timeline Grid Anchor Controller state and side-effect hook | @cartulary/grid-adapter, react, workbook/continuity, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineGridInteractions.ts | web.workbook | production | static inventory | Timeline Grid Interactions state and side-effect hook | @cartulary/grid-adapter, @cartulary/ui-contracts, react, workbook/continuity, workbook/models (+1) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineHistoryActions.ts | web.workbook | production | targeted semantic read | Timeline History Actions state and side-effect hook | react, services, workbook/timeline | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelineHistoryState.ts | web.workbook | production | static inventory | Timeline History State state and side-effect hook | react, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts | web.workbook | production | static inventory | Timeline Inspector Selection state and side-effect hook | @cartulary/ui-contracts, react, workbook/collaboration, workbook/continuity, workbook/models (+1) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts | web.workbook | production | targeted semantic read | Timeline Mention Actions state and side-effect hook | react, services, workbook/collaboration, workbook/models, workbook/timeline | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelineMentions.ts | web.workbook | production | static inventory | Timeline Mentions state and side-effect hook | react, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineMutationCommands.ts | web.workbook | production | targeted semantic read | Timeline Mutation Commands state and side-effect hook | @cartulary/grid-adapter, react, services, workbook/models, workbook/runtime (+2) | S-09 Timeline semantic-action-port source |
| apps/web/src/workbook/timeline/hooks/useTimelinePendingReplayController.ts | web.workbook | production | targeted semantic read | Timeline Pending Replay Controller state and side-effect hook | @cartulary/grid-adapter, react, workbook/mutations, workbook/runtime, workbook/timeline (+1) | S-08 transport-neutral replay policy owner |
| apps/web/src/workbook/timeline/hooks/useTimelinePendingSaves.ts | web.workbook | production | targeted semantic read | Timeline Pending Saves state and side-effect hook | react, workbook/runtime | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineRows.ts | web.workbook | production | static inventory | Timeline Rows state and side-effect hook | react, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineRowsLoader.ts | web.workbook | production | targeted semantic read | Timeline Rows Loader state and side-effect hook | react, workbook/models, workbook/runtime, workbook/timeline, workbook/utils | S-07 transport-neutral query lifecycle consumer |
| apps/web/src/workbook/timeline/hooks/useTimelineSaveStatePresentation.ts | web.workbook | production | static inventory | Timeline Save State Presentation state and side-effect hook | react, workbook/runtime, workbook/utils, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineViewportContinuityController.ts | web.workbook | production | static inventory | Timeline Viewport Continuity Controller state and side-effect hook | @cartulary/grid-adapter, @cartulary/ui-contracts, react, workbook/continuity, workbook/models (+1) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/hooks/useTimelineWorkbookRuntime.ts | web.workbook | production | targeted semantic read | Timeline Workbook Runtime state and side-effect hook | @cartulary/view-contracts, react, workbook/models, workbook/runtime | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/timelineControllerPorts.ts | web.workbook | production | targeted semantic read | timeline Controller Ports pure model, projection, or transformation rules | @cartulary/grid-adapter, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/timelineHistoryModel.ts | web.workbook | production | static inventory | timeline History Model pure model, projection, or transformation rules | workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/timelineRowsModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for timeline Rows Model | external, workbook/timeline | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/timeline/models/timelineRowsModel.ts | web.workbook | production | static inventory | timeline Rows Model pure model, projection, or transformation rules | @cartulary/grid-adapter, @cartulary/ui-contracts, react, workbook/models, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for timeline Viewport Continuity Model | external, workbook/timeline | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.ts | web.workbook | production | static inventory | timeline Viewport Continuity Model pure model, projection, or transformation rules | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/timelineWorkbookRuntime.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for timeline Workbook Runtime | external, workbook/runtime | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/timeline/models/timelineWorkbookSurfaceRuntime.ts | web.workbook | production | static inventory | timeline Workbook Surface Runtime pure model, projection, or transformation rules | react, workbook/collaboration, workbook/layout, workbook/models, workbook/mutations (+2) | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/workbookMentionChips.ts | web.workbook | production | static inventory | workbook Mention Chips pure model, projection, or transformation rules | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/models/workbookRecordFreshness.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Record Freshness | external, workbook/timeline | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/timeline/models/workbookRecordFreshness.ts | web.workbook | production | targeted semantic read | workbook Record Freshness pure model, projection, or transformation rules | none | S-10/S-13 narrow row/freshness port evidence |
| apps/web/src/workbook/timeline/models/workbookTimelineModel.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Timeline Model | external, workbook/models, workbook/utils, workbook/timeline | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/timeline/models/workbookTimelineModel.ts | web.workbook | production | targeted semantic read | workbook Timeline Model pure model, projection, or transformation rules | @cartulary/view-contracts, workbook/models, workbook/query, workbook/utils, workbook/timeline | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.test.tsx | web.workbook | test/support | targeted semantic read | Timeline accepted/stale/query/action/live/conflict admission and state-partition characterization | @cartulary/view-contracts, external, testing, workbook/models, workbook/runtime (+1) | S-13 high-water and race evidence |
| apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.ts | web.workbook | production | targeted semantic read | Timeline row-version, pending/conflict/runtime, and continuity coordinator | react, workbook/models, workbook/runtime, workbook/timeline, workbook/utils | S-13 single query/action/live mutation-admission owner with no transport or grid-vendor dependency |
| apps/web/src/workbook/timeline/ports/TimelineClipboardPastePort.ts | web.workbook | production | targeted semantic read | Semantic Timeline clipboard-paste port | workbook/mutations, workbook/timeline | S-09 narrow paste boundary |
| apps/web/src/workbook/timeline/ports/TimelineEvidenceAttachmentPort.ts | web.workbook | production | targeted semantic read | Semantic Timeline Evidence attachment port | workbook/mutations, workbook/timeline | S-09 narrow attach boundary |
| apps/web/src/workbook/timeline/ports/TimelineHistoryPort.ts | web.workbook | production | targeted semantic read | Semantic Timeline history port | workbook/mutations, workbook/timeline | S-09 narrow history boundary |
| apps/web/src/workbook/timeline/ports/TimelineMentionPort.ts | web.workbook | production | targeted semantic read | Semantic Timeline mention port | workbook/collaboration, workbook/mutations | S-09 narrow mention boundary |
| apps/web/src/workbook/timeline/ports/TimelinePendingMutationPort.ts | web.workbook | production | targeted semantic read | Semantic Timeline pending-mutation port | workbook/mutations, workbook/timeline, workbook/utils | S-08 narrow pending execution boundary; S-09 conflict-resolution normalization |
| apps/web/src/workbook/timeline/ports/TimelineRecordActionPort.ts | web.workbook | production | targeted semantic read | Semantic Timeline review/supersede port | workbook/mutations | S-09 narrow record-action boundary |
| apps/web/src/workbook/timeline/ports/TimelineViewQueryPort.ts | web.workbook | production | targeted semantic read | Semantic Timeline view-query port | workbook/models, workbook/mutations, workbook/timeline | S-07 narrow query boundary |
| apps/web/src/workbook/utils/GridAdapter.anchor.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for Grid Adapter.anchor | @cartulary/grid-adapter, external, react | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/utils/workbookClipboard.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Clipboard | external, workbook/utils | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/utils/workbookClipboard.ts | web.workbook | production | static inventory | workbook Clipboard Workbook interaction utility or state model | @cartulary/grid-adapter | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/utils/workbookKeyboard.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Keyboard | external, workbook/utils | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/utils/workbookKeyboard.ts | web.workbook | production | static inventory | workbook Keyboard Workbook interaction utility or state model | @cartulary/grid-adapter | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/utils/workbookPendingQueue.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Pending Queue | external, workbook/utils | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/utils/workbookPendingQueue.ts | web.workbook | production | targeted semantic read | workbook Pending Queue Workbook interaction utility or state model | shared, workbook/runtime | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/utils/workbookPresence.ts | web.workbook | production | static inventory | workbook Presence Workbook interaction utility or state model | shared | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/utils/workbookRowReconciliation.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Row Reconciliation | external, workbook/utils | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/utils/workbookRowReconciliation.ts | web.workbook | production | static inventory | workbook Row Reconciliation Workbook interaction utility or state model | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/utils/workbookStyles.ts | web.workbook | production | static inventory | workbook Styles Workbook interaction utility or state model | react | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/utils/workbookValueFormat.test.ts | web.workbook | test/support | static inventory | Characterization, policy, or harness evidence for workbook Value Format | external, workbook/utils | Retain as characterization/policy evidence; change only with its owning slice |
| apps/web/src/workbook/utils/workbookValueFormat.ts | web.workbook | production | static inventory | workbook Value Format Workbook interaction utility or state model | none | Retain current owner unless named by S-01-S-14 |
| apps/web/src/workbook/view-state/useWorkbookQueryState.ts | web.workbook | production | targeted semantic read | Workbook Query State Workbook view/query state owner | @cartulary/view-contracts, react, workbook/models | Retain current owner unless named by S-01-S-14 |

## 18. Binary acceptance

### 18.1 Framework criteria

| Criterion | Result | Evidence/explanation |
| --- | --- | --- |
| RF-AC-001 exactly one primary target | PASS | Section 1 names the Workbook application-composition boundary |
| RF-AC-002 all in-scope files listed and unseen files marked | PASS | Appendix A lists 320/320 with explicit inspection level |
| RF-AC-003 every public drift surface mapped | PASS | Sections 5-6 |
| RF-AC-004 behavior-preserving work separated from behavior changes | PASS | Zero-UX covenant and separately authorized change rule |
| RF-AC-005 characterization stated or existing evidence justified | PASS | Sections 10-11 |
| RF-AC-006 checkpoint sequence validates every risky move | PASS | S-01 CP-01-CP-05 and roadmap validation/rollback columns |
| RF-AC-007 module/package boundaries preserved | PASS | Sections 2.3, 3, 7, and 8 |
| RF-AC-008 generated files not hand-edited | PASS | Sections 1.3, 8, 12, and 13 |
| RF-AC-009 no phase identity in production | PASS | Guardrail 17; slice IDs exist only in this tracker |
| RF-AC-010 handoff avoids rediscovery | PASS | Sections 10, 17, and Appendix A |

### 18.2 Web-specific criteria

| Criterion | Result | Evidence/explanation |
| --- | --- | --- |
| WEB-RF-AC-001 Workbook composition is the primary target | PASS | Section 1 |
| WEB-RF-AC-002 adjacent areas require demonstrated material coupling | PASS | Section 2.3 |
| WEB-RF-AC-003 complete state-ownership matrix | PASS | Section 4 covers every discovered class requested by the prompt |
| WEB-RF-AC-004 bootstrap/query/mutation/replay/conflict/collaboration/inspector traced | PASS | All 18 flows in Section 5 |
| WEB-RF-AC-005 saved views preserved and no custom sheets | PASS | Sections 1.3, 4, 6, and D-008 |
| WEB-RF-AC-006 grid-vendor semantics confined | PASS | Graph evidence and Guardrails 1-2 |
| WEB-RF-AC-007 structural and UI behavior work separated | PASS | Scope and zero-UX covenant |
| WEB-RF-AC-008 first slice behavior-preserving and independently reviewable | PASS | Section 10 |
| WEB-RF-AC-009 risky moves have characterization/checkpoints | PASS | Sections 9-12 |
| WEB-RF-AC-010 no generated edit, auth substitution, visible-label identity, or phase boundary | PASS | Sections 6, 8, and 13 |
| WEB-RF-AC-011 planning changed no production/test/generated/manifest/normative file | PASS | Final diff is restricted to this tracker |
| WEB-RF-AC-012 next agent can author S-01 goal without repo-wide rediscovery | PASS | Interface, exact checkpoints/files, tests, commands, rollback, and restart are explicit |

## 19. Authorized remediation execution ledger

### 19.1 Gate and status policy

Execution order is strict and linear:

`P-00 -> F-01 -> S-01 -> S-02 -> S-03 -> S-04 -> S-05 -> S-06A -> S-06B -> S-07 -> S-08 -> S-09 -> S-10 -> S-11 -> S-12 -> S-13 -> S-14 -> V-01`.

After completing a workstream, its row MUST be updated with the exact paths,
substantive result, commands and run roots, generated/accounting changes,
failures and resolution, compatibility decision, residual risk, and rollback
boundary. `make lint-markdown` and `git diff --check` MUST pass after that
update and before the next workstream begins. A normative owner contradiction
sets the current row to `BLOCKED` and stops execution; implementation does not
invent replacement behavior.

### 19.2 Workstream ledger

| Workstream | Status | Depends on | Substantive result and paths | Validation evidence | Accounting/generated evidence | Compatibility, risk, and rollback |
| --- | --- | --- | --- | --- | --- | --- |
| P-00 | DONE | none | Preserved this untracked tracker; authorized the remediation program; added F-01, strict sequencing, execution ledger, terminal criteria, and fresh baseline evidence; superseded planning-only D-004 and incompatible zero-UX/no-generated constraints | PASS: task guide selected `make test-slice OWNER=web.workbook`; frontend boundary `.cartulary/test-results/20260731T134300Z-p1842232`; typecheck exit 0; web.workbook 112/112 `.cartulary/test-results/20260731T134300Z-p1842079`; build-web `.cartulary/test-results/20260731T134411Z-p1847868`; Markdown `.cartulary/test-results/20260731T134611Z-p1851558`; `git diff --check` exit 0 | No production, test, contract, manifest, generated, or lockfile change | Tracker began untracked and is intentionally preserved. Rollback is this P-00 documentation diff only. No residual P-00 risk |
| F-01 | DONE | P-00 | Added the existing-owner Workbook/entity/revision/Timeline operation IDs to `contracts/protocol-ts/http-operations.v1.json`; enhanced the generator with operation-keyed request/response maps, fail-closed missing-validator behavior, and validation-neutral OpenAPI discriminator support; generated the core HTTP types/bindings/validators; added the private typed `workbookOperationExecutor.ts`, a discriminated validated operation result in `services/browserApi.ts`, focused tests, and the exact protocol-import boundary exception | PASS: generate `.cartulary/test-results/20260731T135326Z-p1862770`; format `.cartulary/test-results/20260731T135528Z-p1870923`; package.protocol_ts 5/5 `.cartulary/test-results/20260731T135539Z-p1874358`; web.architecture 11/11 `.cartulary/test-results/20260731T135539Z-p1874389`; boundary `.cartulary/test-results/20260731T135539Z-p1874529`; typecheck and Biome exit 0; web.workbook 112/112 `.cartulary/test-results/20260731T135600Z-p1877612`; build-web `.cartulary/test-results/20260731T135600Z-p1878126`; generate-drift `.cartulary/test-results/20260731T135600Z-p1877519`; generated policy `.cartulary/test-results/20260731T135600Z-p1877544`; JSON `.cartulary/test-results/20260731T135600Z-p1877577`; `git diff --check` exit 0. Initial generate roots `.cartulary/test-results/20260731T134826Z-p1855422` and `.cartulary/test-results/20260731T134919Z-p1857811` exposed unsorted authored IDs and unsupported OpenAPI annotation metadata; initial type/boundary root `.cartulary/test-results/20260731T135356Z-p1868655`/`.cartulary/test-results/20260731T135356Z-p1868683` exposed result narrowing and the required exact adapter exception. All were repaired before closure | Authored source ownership and Workbook family input updated; generated `core-http-types.ts`, `http-operation-bindings.ts`, `protocol-validators.ts`, and execution topology refreshed through Make; 321 manifest paths equal 321 live paths | Malformed/obsolete success and malformed error responses now produce safe `invalid_contract`; no fallback decoder. Roll back the authored projection/generator, executor/browser adapter, tests/boundary/accounting inputs, and their generated outputs as one unit. No residual F-01 blocker |
| S-01 | DONE | F-01 | Added the transport-free `workbookOperationOutcome.ts` algebra and owner-specific Generic accepted DTO; migrated Generic create/linked-note/patch/party-create commands in `createWorkbookMutationCommandPorts.ts` to the generated executor; made `useGenericSurfaceMutationController.ts` and `GenericWorkbookSurface.tsx` consume semantic outcomes and normalized conflicts only; retained coordination raw-result handling inside its own binding until S-06B; strengthened generic create payload identity typing and contract-valid mutation fixtures | PASS: format `.cartulary/test-results/20260731T140739Z-p1920256`; boundary `.cartulary/test-results/20260731T140751Z-p1923716`; typecheck and Biome exit 0; web.workbook 112/112 `.cartulary/test-results/20260731T140622Z-p1915655`; build-web `.cartulary/test-results/20260731T140340Z-p1902530`; generate `.cartulary/test-results/20260731T140229Z-p1896904`; generate-drift `.cartulary/test-results/20260731T140340Z-p1901907`; generated policy `.cartulary/test-results/20260731T140340Z-p1901932`; JSON `.cartulary/test-results/20260731T140340Z-p1901915`; exact raw-import/result audit returned no match; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T140243Z-p1899409` exposed an under-typed existing create payload; initial owner run `.cartulary/test-results/20260731T140340Z-p1901993` exposed a non-UUID mutation fixture now corrected to the OpenAPI owner. Both were repaired before closure | Added one authored production path and one existing-family test title; generated execution topology refreshed through Make; 322 manifest paths equal 322 live paths | Valid request bodies, transaction IDs, refresh ordering, save state, conflicts, selectors, and focus behavior remain. Malformed/obsolete mutation envelopes now fail closed. Roll back the outcome contract, Generic port/adapter/controller/surface/test/boundary/accounting diff as one unit. No residual S-01 blocker |
| S-02 | DONE | S-01 | Added private `features/parties/useGenericPartyLinkWorkflow.ts`; moved pair selection, validation, create-then-link sequencing, lifecycle reset, clear/existing-link commands, and recoverable created-party identity out of `GenericWorkbookSurface.tsx`; added an explicit partial-completion status and retry command using stable UI-contract selectors; characterized that a failed link retains the created Party and a retry sends only the link command | PASS: generate `.cartulary/test-results/20260731T141234Z-p1937742`; final format `.cartulary/test-results/20260731T141349Z-p1942739`; boundary `.cartulary/test-results/20260731T141820Z-p1981236`; typecheck and Biome exit 0; module.parties 2/2 work units and 3 tests `.cartulary/test-results/20260731T141400Z-p1946129`; package.ui 6/6 `.cartulary/test-results/20260731T141702Z-p1971654`; web.workbook 112/112 `.cartulary/test-results/20260731T141702Z-p1971643`; build-web `.cartulary/test-results/20260731T141703Z-p1972099`; generate-drift `.cartulary/test-results/20260731T141820Z-p1980942`; generated policy `.cartulary/test-results/20260731T141820Z-p1980908`; JSON `.cartulary/test-results/20260731T141820Z-p1980946`. Initial format runs exposed an unused identity and unsorted authored family input; initial typecheck exposed missing closed-vocabulary selector tokens; initial web.workbook root `.cartulary/test-results/20260731T141400Z-p1946112` exposed a misspelled fixture field key; invalid owner spelling root `.cartulary/test-results/20260731T141400Z-p1946146` was corrected to `package.ui`. All were repaired before closure | Added one production path and one Workbook characterization title to authored ownership/family inputs; added two authored UI selector tokens; regenerated execution topology through Make; 323 manifest paths equal 323 live paths | Valid create/link request bodies, exact one-link-after-create ordering, transaction identity, refresh, conflict, focus, and inspector behavior remain. A successful create followed by failed link is now visible and retryable without duplicate Party creation or invented rollback. Roll back the private hook, surface/test/UI-selector/accounting diff together. No residual S-02 blocker |
| S-03 | DONE | S-01 | Added distinct entity create, patch, and paste accepted DTOs/outcomes in `workbookMutationCommandPorts.ts`; routed all three commands through generated operation bindings and strict response validation in `createWorkbookMutationCommandPorts.ts`; added request/result consistency checks and normalized paste rows; removed raw result, status, envelope, error, and conflict interpretation for ordinary entity mutations from `EntityWorkbookSurface.tsx`; retained only the characterized raw merge path for S-04 | PASS: generate `.cartulary/test-results/20260731T142546Z-p1995278`; final format `.cartulary/test-results/20260731T143231Z-p2040317`; module.entities 3/3 work units and 23 tests `.cartulary/test-results/20260731T142559Z-p1997757`; web.workbook 112/112 `.cartulary/test-results/20260731T143118Z-p2035770`; boundary `.cartulary/test-results/20260731T143242Z-p2044085`; typecheck and Biome exit 0; build-web `.cartulary/test-results/20260731T143242Z-p2044512`; generate-drift `.cartulary/test-results/20260731T143242Z-p2043732`; generated policy `.cartulary/test-results/20260731T143242Z-p2043786`; JSON `.cartulary/test-results/20260731T143242Z-p2043837`; ordinary-entity raw-result audit leaves only merge lines; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T142332Z-p1990236` exposed exact-optional format typing. Initial Workbook roots `.cartulary/test-results/20260731T142559Z-p1997769` and `.cartulary/test-results/20260731T142838Z-p2030694` exposed mutation fixtures with non-contract UUID identities. Typed request construction and owner-valid fixtures repaired both without weakening validation | Added one adapter characterization title to the authored Workbook family and regenerated execution topology through Make; no production path addition; 323 manifest paths equal 323 live paths | Valid entity requests, refresh, selection, alias focus, conflict registration, and exact-match paste behavior remain. Malformed or context-inconsistent create/patch/paste successes now fail closed. Merge compatibility remains deliberately unchanged until S-04. Roll back the entity DTO/adapter/surface/test-family/fixture diff as one unit. No residual S-03 blocker |
| S-04 | DONE | S-03 | Added the private lifecycle-guarded `features/entities/useEntityMergeController.ts` owner and focused hook tests; moved candidate/reason/plan state, confirmation, semantic rejection, allowlisted precondition presentation, accepted cleanup, refresh, preview, and survivor retarget out of `EntityWorkbookSurface.tsx`; added a distinct merge accepted DTO and generated-operation adapter normalization; made merge-precondition error detail projection operation-specific and bounded inside `workbookOperationExecutor.ts`; removed the final entity presentation import of Workbook services and raw envelopes | PASS: generate `.cartulary/test-results/20260731T144116Z-p2061853`; final format `.cartulary/test-results/20260731T144750Z-p2128918`; web.workbook 113/113 `.cartulary/test-results/20260731T144453Z-p2097627`; module.entities 3/3 work units, 23 tests, including browser merge `.cartulary/test-results/20260731T144610Z-p2102266`; boundary `.cartulary/test-results/20260731T144759Z-p2132668`; typecheck and Biome exit 0; build-web `.cartulary/test-results/20260731T144759Z-p2133091`; generate-drift `.cartulary/test-results/20260731T144759Z-p2132319`; generated policy `.cartulary/test-results/20260731T144759Z-p2132357`; JSON `.cartulary/test-results/20260731T144759Z-p2132410`; entity surface/controller transport audit returned no match; `git diff --check` exit 0. Initial owner roots `.cartulary/test-results/20260731T144148Z-p2066005` and `.cartulary/test-results/20260731T144148Z-p2066006` exposed one non-contract merge incident fixture and a lifecycle cleanup conflation that erased accepted continuity. Owner-valid UUID fixtures and separate plan/full-reset commands repaired both without relaxing validation | Added controller production/test paths and two characterized titles to authored ownership/family inputs; regenerated execution topology through Make; 325 manifest paths equal 325 live paths | Valid merge review copy, exact request, role gating, dependent preview, sanitized collision details, refresh, and survivor identity remain. Late completions after lifecycle/authorization change cannot clear drafts, refresh, or retarget. Unknown server details are discarded. Roll back the controller, semantic merge adapter/executor mapping, surface/test/accounting diff together. No residual S-04 blocker |
| S-05 | DONE | S-01 | Added the private `features/assessments/useAssessmentCreationController.ts` owner and focused lifecycle tests; moved assessment draft/mode, local validation, append-only submission, semantic failure, refresh, accepted reset, and lifecycle invalidation out of `AssessmentWorkbookSurface.tsx`; added a distinct assessment accepted DTO, adapter-side `canCreate`, generated create binding, strict response validation, and stronger payload identity typing; preserved the owner-visible `invalid_mutation_payload` copy while containing all other public error normalization | PASS: generate `.cartulary/test-results/20260731T145400Z-p2149703`; final format `.cartulary/test-results/20260731T145903Z-p2222161`; web.workbook 114/114 `.cartulary/test-results/20260731T145630Z-p2178413`; module.assessments 3/3 work units and 15 tests, including browser workflows `.cartulary/test-results/20260731T145823Z-p2202463`; boundary `.cartulary/test-results/20260731T145913Z-p2225917`; typecheck and Biome exit 0; build-web `.cartulary/test-results/20260731T145913Z-p2226370`; generate-drift `.cartulary/test-results/20260731T145913Z-p2225572`; generated policy `.cartulary/test-results/20260731T145913Z-p2225634`; JSON `.cartulary/test-results/20260731T145913Z-p2225682`; assessment surface/controller transport audit returned no match; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T145158Z-p2144701` exposed under-typed payload identity. Initial owner roots `.cartulary/test-results/20260731T145433Z-p2153818` and `.cartulary/test-results/20260731T145433Z-p2153821` exposed non-contract success UUID fixtures; follow-up owner root `.cartulary/test-results/20260731T145630Z-p2178411` exposed required invalid-payload copy normalization. All were repaired without adding a fallback decoder | Added controller production/test paths and two characterized titles to authored ownership/family inputs; regenerated execution topology through Make; 327 manifest paths equal 327 live paths | Accepted submission resets only owner-approved create fields while retaining the subject; rejected submission retains rationale and support selection; invalid drafts do not dispatch; late access/lifecycle completions cannot refresh or restore state. Valid routes, request bodies, follow-on behavior, keyboard selection, and copy remain. Roll back the controller, semantic assessment adapter/model typing, surface/tests/accounting diff together. No residual S-05 blocker |
| S-06A | DONE | F-01, S-01 | Replaced the void/exception Evidence attachment port with `EvidenceCapabilityPort` and distinct attach/access outcomes; moved object-slot creation, upload, row attachment, handle issuance, response correlation, and public-href admission into `createWorkbookMutationCommandPorts.ts`; made `useEvidenceWorkbookBindings.tsx` consume only semantic outcomes with reset/disposal generations; hardened `workbookEvidence.ts` so upload responses are never read, only network/503/504 failures are retryable, and attempts remain capped at three; removed the unused exception-based attach and direct-handle service APIs; retained only an explicit allowlist of owner-approved Evidence access reason codes; strengthened surface fixtures to return addressed record identity | PASS: final format `.cartulary/test-results/20260731T151829Z-p2342163`; module.evidence 44/44 `.cartulary/test-results/20260731T151833Z-p2345304`; web.workbook 114/114 `.cartulary/test-results/20260731T152026Z-p2373028`; boundary `.cartulary/test-results/20260731T152221Z-p2378287`; typecheck and Biome exit 0; build-web `.cartulary/test-results/20260731T152239Z-p2380058`; generate-drift `.cartulary/test-results/20260731T152242Z-p2383151`; generated policy `.cartulary/test-results/20260731T152254Z-p2387007`; JSON `.cartulary/test-results/20260731T152256Z-p2387351`; binding service/exception audit and upload-body audit returned no prohibited match; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T150249Z-p2238411` exposed generated unknown-valued upload headers and was repaired by bounded string-header projection. Initial evidence owner root `.cartulary/test-results/20260731T150606Z-p2249038` exposed loss of the required `unsupported_preview` copy and led to an explicit safe reason allowlist. Initial Workbook root `.cartulary/test-results/20260731T151226Z-p2318571` exposed owner copy and fixture record-correlation defects; focused root `.cartulary/test-results/20260731T151441Z-p2327076` then exposed the remaining fixed-record fixture. Both passed at `.cartulary/test-results/20260731T151521Z-p2330936` before the final full owner runs | Added the upload retry/body-discard test row to authored `module.evidence`; removed the obsolete dead-helper row from authored `web.application`; regenerated execution topology through Make; no source path addition; 327 manifest paths equal 327 live paths | Valid blob/upload/attach ordering, transaction IDs, refresh, opaque handles, bounded retry count, selectors, accessibility copy, and safe owner reason codes remain. Malformed or context-inconsistent responses fail closed; raw storage response bodies cannot reach logs, state, or copy; non-network/503/504 upload failures are terminal. The unused legacy attach/handle service APIs are intentionally removed with no production caller. Roll back the port/adapter/binding/service/test-family/topology diff as one unit. No residual S-06A blocker |
| S-06B | DONE | S-01 | Added distinct task-lifecycle and decision-supersede accepted DTOs/outcomes and a closed task-status type; migrated both adapter methods to generated `patchRecord`/`supersedeRecord` bindings with strict validation and request/response identity checks; added private lifecycle-guarded `useCoordinationWorkflowController.ts` for validation, submission, semantic rejection, accepted cleanup, refresh, and late-completion invalidation; reduced `CoordinationWorkflowBindings.tsx` to rendering and controller commands; added adapter and hook characterization | PASS: generate `.cartulary/test-results/20260731T152956Z-p2405737`; final format `.cartulary/test-results/20260731T153031Z-p2408747`; module.tasksdecisions 3/3 tests `.cartulary/test-results/20260731T153045Z-p2412479`; web.workbook 114/114 rows `.cartulary/test-results/20260731T153129Z-p2431636`; boundary `.cartulary/test-results/20260731T153250Z-p2436416`; typecheck and Biome exit 0; build-web `.cartulary/test-results/20260731T153308Z-p2438161`; generate-drift `.cartulary/test-results/20260731T153311Z-p2441262`; generated policy `.cartulary/test-results/20260731T153322Z-p2445120`; JSON `.cartulary/test-results/20260731T153324Z-p2445464`; binding/controller transport-result audit returned no prohibited match; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T152701Z-p2394045` exposed generated public alias names and the obsolete mutation-prop shape. Initial test-family validation exposed the tasks/decisions owner contract's missing Vitest evidence kind; the authored owner routing was expanded rather than misattributing the test. Follow-up typecheck `.cartulary/test-results/20260731T153004Z-p2407993` exposed readonly fixture inference. All were repaired before closure | Added controller production/test paths to authored source ownership; added a `module.tasksdecisions` frontend-unit family row and expanded its behavior verification to Vitest; added one adapter characterization title to the existing Workbook family; regenerated execution topology through Make; 329 manifest paths equal 329 live paths | Valid task/decision routes, payloads, transaction IDs, local validation copy, refresh, and accepted field cleanup remain. A valid response from the wrong supersede union branch or with mismatched record identity now fails closed. Late lifecycle/authorization completions cannot reject, refresh, or clear current state. Roll back the semantic port/adapter/controller/binding/tests/verification/accounting/topology diff together. No residual S-06B blocker |
| S-07 | DONE | F-01 | Added the narrow `TimelineViewQueryPort` and private `createTimelineViewQueryAdapter.ts`; derived route/method from the generated `queryWorkbookView` binding, validated the closed success envelope, correlated incident/schema identity, normalized rows below the port, contained aborts, and mapped transport/contract failures to safe semantic outcomes; injected the port from `TimelineWorkbook.tsx`; removed routes, services, envelope/error parsing, and wire-row decoding from `useTimelineRowsLoader.ts`; retained latest-generation, high-water freshness, pending-row reconciliation, draft/focus, loading, access-loss, and continuity policy; migrated routed query fixtures to owner-valid UUIDs and required query metadata | PASS: focused adapter 2 tests `.cartulary/test-results/20260731T154228Z-p2486348`; real webserver query row `.cartulary/test-results/20260731T153801Z-p2456139`; module.timeline 9/9 work units and 51 tests `.cartulary/test-results/20260731T154843Z-p2519957`; web.workbook 114/114 `.cartulary/test-results/20260731T155706Z-p2568401`; boundary `.cartulary/test-results/20260731T155940Z-p2574668`; web.architecture 11/11 `.cartulary/test-results/20260731T160022Z-p2576545`; final format `.cartulary/test-results/20260731T160046Z-p2578451`; typecheck `.cartulary/test-results/20260731T155655Z-p2567856`; Biome `.cartulary/test-results/20260731T160050Z-p2581638`; build-web `.cartulary/test-results/20260731T160053Z-p2582265`; generate-drift `.cartulary/test-results/20260731T160058Z-p2585969`; generated policy `.cartulary/test-results/20260731T160109Z-p2589833`; JSON `.cartulary/test-results/20260731T160111Z-p2590194`; loader transport audit returned no prohibited match. Initial focused root `.cartulary/test-results/20260731T154030Z-p2481490` exposed missing query metadata in the new fixture. Initial owner/Workbook roots `.cartulary/test-results/20260731T154245Z-p2486796` and `.cartulary/test-results/20260731T155019Z-p2546377` exposed obsolete placeholder identities in strict response fixtures. Initial boundary roots `.cartulary/test-results/20260731T155834Z-p2573212` and `.cartulary/test-results/20260731T155901Z-p2573894` exposed the checker’s suffix-glob semantics; the durable Timeline adapter directory was admitted exactly. Architecture root `.cartulary/test-results/20260731T155944Z-p2575152` exposed stale S-06A upload-header policy text, updated to the hardened bounded-header implementation. All were repaired without weakening the decoder | Added query adapter production/test and query-port paths to authored source ownership; added one `module.timeline` frontend family row covering both adapter behaviors; updated the authored protocol-adapter boundary; regenerated execution topology through Make; 332 manifest paths equal 332 live paths | Valid query route, POST body, sort/filter/pagination encoding, refresh ordering, freshness retry, selection, invalid draft text, focus, and continuity remain. Superseded/disposed requests now abort; malformed or context-inconsistent successes fail closed as `invalid_contract`. Roll back the port/adapter/injection/loader, contract-valid fixture migration, boundary/accounting/family/topology diff as one unit. No residual S-07 blocker |
| S-08 | DONE | S-07 | Added the narrow `TimelinePendingMutationPort` and private generated-operation adapter; made create/patch pending units retain semantic scope/intent plus stable unit/transaction identity without route, method, or cached request body; made the adapter materialize fresh and replay requests through one generated binding path and use the dispatch-time committed row version; normalized accepted rows and same-field conflicts below the port; kept FIFO/capacity/coalescing/auth pause/requery/re-key/discard policy in `useTimelinePendingReplayController.ts` while removing its transport, envelope, and error parsing; retained transport timing observation inside the adapter; updated the shared runtime to materialize its independent generic pending commands from semantic units; removed the obsolete Timeline replay dispatcher and cached-payload materializer | PASS: generate `.cartulary/test-results/20260731T161410Z-p2617631`; focused queue `.cartulary/test-results/20260731T162043Z-p2664885`; focused pending adapter `.cartulary/test-results/20260731T162523Z-p2720929`; measured blank-row create `.cartulary/test-results/20260731T162530Z-p2721306`; module.timeline 9/9 work units and 52 tests `.cartulary/test-results/20260731T162611Z-p2740167`; focused Workbook fixture repair 2/2 `.cartulary/test-results/20260731T162944Z-p2774568`; web.workbook 114/114 `.cartulary/test-results/20260731T162952Z-p2774928`; boundary `.cartulary/test-results/20260731T163112Z-p2779980`; web.architecture 11/11 `.cartulary/test-results/20260731T163132Z-p2785479`; final format `.cartulary/test-results/20260731T162940Z-p2771440`; typecheck `.cartulary/test-results/20260731T163112Z-p2779973`; Biome `.cartulary/test-results/20260731T163113Z-p2780029`; build-web `.cartulary/test-results/20260731T163113Z-p2780288`; generate-drift `.cartulary/test-results/20260731T163132Z-p2785341`; generated policy `.cartulary/test-results/20260731T163132Z-p2785359`; JSON `.cartulary/test-results/20260731T163132Z-p2785381`; pending-unit coordinate and controller transport audits returned no prohibited match. Initial format/type/focused roots exposed an obsolete adapter dependency, generated response alias, and stale queue expectations; initial module.timeline root `.cartulary/test-results/20260731T161530Z-p2625250` exposed a non-contract create fixture; the next owner root `.cartulary/test-results/20260731T162100Z-p2665789` exposed the dropped response-timing event; initial web.workbook root `.cartulary/test-results/20260731T162743Z-p2766195` exposed two non-contract created-record fixtures. Semantic typing, adapter-contained observation, and owner-valid UUID fixtures repaired all failures without weakening validation | Added pending adapter production/test and pending-port paths to authored source ownership; added one `module.timeline` frontend family row; renamed the semantic pending-queue and Timeline payload characterization titles in authored families; regenerated execution topology through Make; 335 manifest paths equal 335 live paths | Valid capacity 64, FIFO, contiguous coalescing, stable unit/transaction IDs, uncertain retry, auth pause/requery, same-field transfer, explicit re-key, discard, refresh, focus, and inspector continuity remain. Memory-local units require no persisted migration. Malformed/context-inconsistent results fail closed; compatibility with stored transport coordinates is intentionally removed. Roll back the semantic queue/runtime/controller, port/adapter, timing observer, fixture/accounting/family/topology diff together. No residual S-08 blocker |
| S-09 | DONE | S-07, S-08 | Added narrow history, record review/supersede, mention, clipboard, Evidence attachment, bulk, related-record, and mutation-identity ports/adapters; split `TimelineMutationCommandPorts` into `identity`, `bulk`, and `related`; moved history DTOs to `timelineHistoryModel.ts`; made every named Timeline hook consume semantic outcomes; made conflict resolution pass a generated-validated semantic mutation through `workbookConflictResolutionAdapter.ts` and Timeline normalization; removed the root’s raw envelope/error imports, the obsolete `createEvidenceWithInitialBlob` service workflow, and deleted duplicate serializer `timeline/services/timelineMutationRequests.ts`; documented adapters/ports in `apps/web/src/README.md`; added `createTimelineActionAdapters.test.ts` | PASS: generate `.cartulary/test-results/20260731T171733Z-p2889027`; final format `.cartulary/test-results/20260731T172422Z-p2968408`; typecheck `.cartulary/test-results/20260731T172522Z-p2972250`; boundary `.cartulary/test-results/20260731T171913Z-p2903540`; Biome `.cartulary/test-results/20260731T171920Z-p2904129`; build-web `.cartulary/test-results/20260731T171923Z-p2904758`; module.timeline 9/9 work units and 53 tests `.cartulary/test-results/20260731T172014Z-p2914583`; web.workbook 114/114 `.cartulary/test-results/20260731T171428Z-p2883183`, plus final generated-conflict adapter cases 3/3 `.cartulary/test-results/20260731T172533Z-p2972788`; history 13/13 `.cartulary/test-results/20260731T171411Z-p2882761`; new action adapters `.cartulary/test-results/20260731T170656Z-p2844304`; atomic Evidence retry `.cartulary/test-results/20260731T170704Z-p2844670`; repaired Evidence invalidation fixture `.cartulary/test-results/20260731T172426Z-p2971537`; web.architecture 11/11 `.cartulary/test-results/20260731T171955Z-p2913331`; generate-drift `.cartulary/test-results/20260731T171931Z-p2908560`; generated policy `.cartulary/test-results/20260731T171943Z-p2912412`; JSON `.cartulary/test-results/20260731T171945Z-p2912755`; named-hook/component transport, raw-result, legacy-service, and duplicate-serializer audits returned no prohibited match. Initial format/type roots `.cartulary/test-results/20260731T165419Z-p2801293` and `.cartulary/test-results/20260731T165502Z-p2807919` exposed incomplete port wiring and stale tests; generate roots `.cartulary/test-results/20260731T170359Z-p2825344` and `.cartulary/test-results/20260731T170439Z-p2827920` exposed authored ordering and overlapping selectors; focused adapter root `.cartulary/test-results/20260731T170614Z-p2840531` exposed a non-UUID mention fixture; Workbook root `.cartulary/test-results/20260731T170849Z-p2872390` exposed required history paging and conflict-resolution continuity; boundary roots `.cartulary/test-results/20260731T171633Z-p2888213` and `.cartulary/test-results/20260731T171833Z-p2899057` drove the generic conflict adapter behind the existing generated executor; module.evidence root `.cartulary/test-results/20260731T172151Z-p2941317` exposed one obsolete non-UUID query fixture, repaired by the focused passing root above. All changes preserve strict validation | Added 12 source/test paths and deleted one obsolete service path; updated authored source ownership and `module.timeline`/`module.evidence` test families; regenerated execution topology through Make; 346 manifest paths equal 346 live paths | Valid routes, bodies, stable transaction IDs, refresh ordering, FIFO replay, query/action/live freshness, focus, scroll, inspector, conflict, and copy remain. Malformed/context-inconsistent action results fail closed; upload bodies and unsafe details stay below adapters; no broad action facade, fallback decoder, re-export, allowlist, or duplicate serializer exists. Roll back the action ports/adapters, split command interfaces, semantic hook/root wiring, Evidence service change, tests/docs/accounting/topology, and service deletion as one unit. No residual S-09 blocker |
| S-10 | DONE | S-07, S-09 | Added `timeline/collaboration/useTimelineCollaborationBindings.ts` as the single Timeline consumer binding over the existing shell coordinator; moved client-transaction resolver and active-surface registration, sparse `record_changed` validation/application, version admission, reset/access invalidation, gap refresh, presence publication, authorization-recovery delegation, surface-switch cleanup, and teardown out of `TimelineWorkbook.tsx`; generalized the coordinator external-store hook to its minimal structural store contract; documented the boundary and added focused lifecycle/fail-closed tests | PASS: generate `.cartulary/test-results/20260731T173451Z-p2993260`; final format `.cartulary/test-results/20260731T173447Z-p2990154`; typecheck `.cartulary/test-results/20260731T173504Z-p2995674`; boundary `.cartulary/test-results/20260731T173258Z-p2984060`; Biome `.cartulary/test-results/20260731T173714Z-p3034148`; build-web `.cartulary/test-results/20260731T173533Z-p2997416`; web.collaboration 3/3 work units `.cartulary/test-results/20260731T173503Z-p2995603`; module.timeline 9/9 work units and 53 tests, including browser/visual rows, `.cartulary/test-results/20260731T173533Z-p2996980`; web.workbook 114/114 `.cartulary/test-results/20260731T173533Z-p2996985`; web.architecture 11/11 `.cartulary/test-results/20260731T173533Z-p2997005`; generate-drift `.cartulary/test-results/20260731T173714Z-p3033830`; generated policy `.cartulary/test-results/20260731T173714Z-p3033849`; JSON `.cartulary/test-results/20260731T173714Z-p3033871`; root collaboration-registration/sparse-patch audit returned no prohibited match; `git diff --check` exit 0. Initial format root `.cartulary/test-results/20260731T173227Z-p2977076` exposed an implicit patch type; initial typecheck `.cartulary/test-results/20260731T173258Z-p2984065` exposed an under-typed test fixture; concurrent preflight roots `.cartulary/test-results/20260731T173407Z-p2986242` and `.cartulary/test-results/20260731T173407Z-p2986276` exposed authored test-row ordering. Explicit patch typing, owner normalization, and ASCII-sorted authored accounting repaired all failures | Added collaboration binding production/test paths; updated authored source ownership and `web.collaboration` test family; regenerated execution topology through Make; 348 manifest paths equal 348 live paths | Stream sequencing, WebSocket/reconnect/reset policy, authorization recovery, presence debounce, transaction identity, high-water behavior, draft retention, focus, scroll, inspector, DOM, selectors, and copy remain unchanged. Malformed or target-inconsistent sparse patches request authoritative refresh; stale events are ignored semantically. Roll back the binding/test/docs/accounting changes, root adoption, and minimal store typing as one unit. No residual S-10 blocker |
| S-11 | DONE | S-09 | Added private `timeline/bulk/useTimelineBulkTagController.ts`; moved stable committed-record selection, current-page grid selection adaptation, pruning, tag draft/status, exact dispatch-time version targets, synchronous duplicate-submit exclusion, semantic rejection/conflict presentation, refresh, authorization/disposal generation guards, and accepted-state retention out of `TimelineWorkbook.tsx`; kept fill-down on its separate grid action path; added direct controller tests, bounded conflict-count copy, README ownership, and repaired the existing bulk-success fixture to the strict adopted batch envelope | PASS: generate `.cartulary/test-results/20260731T174540Z-p3054268`; final format `.cartulary/test-results/20260731T175305Z-p3108748`; typecheck `.cartulary/test-results/20260731T175453Z-p3139664`; boundary `.cartulary/test-results/20260731T175453Z-p3139689`; Biome `.cartulary/test-results/20260731T175453Z-p3139719`; build-web `.cartulary/test-results/20260731T175517Z-p3142147`; module.timeline 9/9 work units and 54 tests, including browser/visual rows, `.cartulary/test-results/20260731T175309Z-p3111875`; web.workbook 114/114 `.cartulary/test-results/20260731T175043Z-p3071471`; focused strict versioned bulk command `.cartulary/test-results/20260731T175453Z-p3139543`; web.architecture 11/11 `.cartulary/test-results/20260731T175043Z-p3071451`; generate-drift `.cartulary/test-results/20260731T175517Z-p3141629`; generated policy `.cartulary/test-results/20260731T175517Z-p3141634`; JSON `.cartulary/test-results/20260731T175517Z-p3141658`; root bulk-state/submission and controller transport audits returned no prohibited match; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T174428Z-p3046402` exposed an over-specified grid test identity. Existing owner roots `.cartulary/test-results/20260731T174600Z-p3056916`, `.cartulary/test-results/20260731T174715Z-p3062225`, and `.cartulary/test-results/20260731T174904Z-p3066283` exposed the stale pre-validation bulk-success fixture; the focused passing root above proves the repaired owner envelope. The first full Timeline root `.cartulary/test-results/20260731T175043Z-p3071524` had all product rows pass but failed harness accounting because a concurrent owner run occupied visual port 39048; the isolated full rerun passed | Added bulk controller production/test paths; updated authored source ownership and `module.timeline` test family; regenerated execution topology through Make; 350 manifest paths equal 350 live paths | Stable record IDs, current-page select-all, draft/group/pending exclusion, exact dispatch-time row versions, one batch request, selection/draft retention after success or rejection, authoritative refresh, keyboard/grid behavior, DOM, selectors, and ordinary success copy remain. Partial conflict results now disclose only bounded applied/conflict counts after refresh instead of claiming full success; raw conflict structures stay below the semantic port. Roll back the controller/test/docs/accounting/root adoption and strict fixture repair as one unit. No residual S-11 blocker |
| S-12 | DONE | S-10 | Added private `timeline/editing/useTimelineEditorDraftRegistry.ts`; moved scalar draft values, semantic input elements/test IDs, submitted-value cleanup, conflict/discard cleanup, row materialization, focus lookup, row pruning, access invalidation, and schema-generation invalidation out of `TimelineWorkbook.tsx`; made row query reconciliation, mutation commands, paste, conflict projection, viewport continuity, and renderer construction consume the registry; removed the root draft/ref maps and materialization/registration helpers; added focused invalid/valid/submitted/removal/schema/focus characterization and README ownership | PASS: generate `.cartulary/test-results/20260731T180617Z-p3155484`; final format `.cartulary/test-results/20260731T181707Z-p3266616`; focused registry row `.cartulary/test-results/20260731T181718Z-p3270082`; typecheck `.cartulary/test-results/20260731T181718Z-p3270180`; boundary `.cartulary/test-results/20260731T181718Z-p3270210`; Biome `.cartulary/test-results/20260731T181718Z-p3270262`; build-web `.cartulary/test-results/20260731T181718Z-p3270585`; module.timeline 9/9 work units and 55 tests, including browser/visual rows, `.cartulary/test-results/20260731T181335Z-p3233902`; package.grid_adapter 1/1 work unit and 27 tests `.cartulary/test-results/20260731T180931Z-p3197751`; web.workbook 114/114 `.cartulary/test-results/20260731T181511Z-p3261484`; web.architecture 11/11 `.cartulary/test-results/20260731T180931Z-p3197757`; generate-drift `.cartulary/test-results/20260731T181140Z-p3223875`; generated policy `.cartulary/test-results/20260731T181140Z-p3223880`; JSON `.cartulary/test-results/20260731T181140Z-p3223906`; root map/helper audit and registry vendor-coordinate/transport audit returned no prohibited match; `git diff --check` exit 0. Initial format root `.cartulary/test-results/20260731T180639Z-p3158200` reported an over-specified memo dependency; making schema identity an explicit memo input repaired it before closure | Added editor-registry production/test paths; updated authored source ownership and `module.timeline` test family; regenerated execution topology through Make; 352 manifest paths equal 352 live paths | Invalid scalar text survives authoritative replacement for a retained semantic row; grid and inspector drafts remain distinct; only an exactly submitted value is cleared; removed rows, access loss, and schema replacement invalidate owned state; focus resolution remains semantic with the existing DOM fallback. Request, validation, conflict, replay, focus/scroll, DOM, selector, keyboard, and copy behavior remain otherwise unchanged. Roll back the registry/test/docs/accounting changes and consumer/root adoption as one unit. No residual S-12 blocker |
| S-13 | DONE | S-08-S-12 | Added private `timeline/mutations/useTimelineRowMutationCoordinator.ts`; moved committed row/version high-water admission, query generations, action/live/conflict ordering, accepted-row projection, idle barriers, save-state publication, local conflict projection, runtime surface/drainer registration, socket transaction tracking, serialized explicit work, pending-discard reconciliation, and continuity sequencing out of `TimelineWorkbook.tsx`; made query and collaboration consume narrow admission ports; hardened `useTimelineCommittedRows.ts` so optimistic scalar values, collection drafts, and pending signatures cannot enter committed state; made conflict server versions/state pass through the same admission owner; reduced the root by roughly 600 lines and added accepted/stale/query-action/live/conflict partition tests | PASS: generate `.cartulary/test-results/20260731T182907Z-p3286488`; final format `.cartulary/test-results/20260731T183952Z-p3305929`; focused coordinator row `.cartulary/test-results/20260731T184011Z-p3309172`; module.timeline 9/9 work units and 56 tests, including browser/visual rows, `.cartulary/test-results/20260731T184136Z-p3314269`; web.workbook 114/114 `.cartulary/test-results/20260731T184019Z-p3309538`; web.collaboration 3/3 `.cartulary/test-results/20260731T184316Z-p3342313`; web.architecture 11/11 `.cartulary/test-results/20260731T184316Z-p3342368`; typecheck `.cartulary/test-results/20260731T184316Z-p3342448`; boundary `.cartulary/test-results/20260731T184316Z-p3342486`; Biome `.cartulary/test-results/20260731T184316Z-p3342539`; build-web `.cartulary/test-results/20260731T184316Z-p3342935`; generate-drift `.cartulary/test-results/20260731T184340Z-p3348916`; generated policy `.cartulary/test-results/20260731T184340Z-p3348932`; JSON `.cartulary/test-results/20260731T184340Z-p3348952`; root admission/queue/runtime-policy audit and coordinator transport/grid-vendor audit returned no prohibited match; `git diff --check` exit 0. Initial typecheck roots `.cartulary/test-results/20260731T182510Z-p3279761`, `.cartulary/test-results/20260731T182752Z-p3284946`, `.cartulary/test-results/20260731T183014Z-p3289567`, and `.cartulary/test-results/20260731T183058Z-p3290594` exposed integration imports, exact test-result typing, and nullable identity narrowing while the committed partition was hardened. Initial Biome root `.cartulary/test-results/20260731T183921Z-p3304272` exposed one formatting error and a test-only non-null assertion. All were corrected without weakening admission rules | Added coordinator production/test paths; updated authored source ownership and `module.timeline` test family; regenerated execution topology through Make; 354 manifest paths equal 354 live paths | Valid request bodies, stable transaction IDs, FIFO/re-key/discard behavior, conflict copy, authoritative refresh order, live-event admission, focus/scroll/inspector continuity, DOM, selectors, keyboard, and visible save state remain. An optimistic visible row can no longer masquerade as committed state, and conflict server state now advances the shared high-water mark; these are intentional correctness fixes. Roll back the coordinator/test/docs/accounting changes, root adoption, committed-projection hardening, and conflict-admission wiring as one unit. No residual S-13 blocker |
| S-14 | DONE | S-07-S-13 | Closed Timeline composition ownership: made `useTimelineMutationCommands.ts` own replacement drafts and expose structured snapshot/commands; added `TimelineRowMutationEditorPort` plus `createTimelineRowMutationEditorAdapter.ts` so the root supplies a semantic editor boundary instead of inline mutation translations; added `TimelineCollaborationBoundary.tsx` so session/coordinator attachment no longer belongs to `TimelineWorkbook.tsx`; retained the root as controller/adapter composition plus rendering; added exact Timeline workflow-to-transport, lower-layer-to-component, and adapter back-edge rules in `tools/frontend_import_boundaries.json`; documented private owners; completed strict optional standardized surface projection for Findings, Investigative Queries, and Forensic Keywords through the authored Workbook OpenAPI owner and generator; made retry-admitted pending edits close the active editor while the semantic unit continues replay; repaired stateful/a11y/visual fixtures to use contract-valid public error envelopes while preserving public copy and existing goldens | PASS: final generate `.cartulary/test-results/20260731T193211Z-p3732789`; format `.cartulary/test-results/20260731T193808Z-p3800523`; boundary `.cartulary/test-results/20260731T193812Z-p3803694`; typecheck `.cartulary/test-results/20260731T193816Z-p3804214`; Biome `.cartulary/test-results/20260731T193827Z-p3804806`; build-web `.cartulary/test-results/20260731T193830Z-p3805432`; module.timeline 9/9 work units and 56 tests `.cartulary/test-results/20260731T193337Z-p3745737`; web.workbook 114/114 `.cartulary/test-results/20260731T193510Z-p3773525`; web.architecture 11/11 `.cartulary/test-results/20260731T193625Z-p3778204`; web.collaboration 3/3 `.cartulary/test-results/20260731T193635Z-p3779451`; package.grid_adapter 27/27 `.cartulary/test-results/20260731T193643Z-p3779991`; protocol-ts 5/5 `.cartulary/test-results/20260731T190816Z-p3503255`; webserver-backed 2/2 `.cartulary/test-results/20260731T193900Z-p3814100`; stateful 2/2 `.cartulary/test-results/20260731T194351Z-p3844233`; a11y 2/2 `.cartulary/test-results/20260731T194606Z-p3868446`; measurement 2/2 `.cartulary/test-results/20260731T194727Z-p3889434`; visual 2/2 with no golden update `.cartulary/test-results/20260731T194820Z-p3909475`; generate-drift `.cartulary/test-results/20260731T193835Z-p3809234`; generated policy `.cartulary/test-results/20260731T193847Z-p3813114`; JSON `.cartulary/test-results/20260731T193849Z-p3813457`; root transport/subscription and semantic pending-coordinate audits returned no prohibited match; Timeline adapters have no React/DOM import; `git diff --check` exit 0. Initial format `.cartulary/test-results/20260731T185001Z-p3359803` exposed one stale dependency; generate `.cartulary/test-results/20260731T190719Z-p3495640` required the authored release fingerprint update; webserver roots `.cartulary/test-results/20260731T185256Z-p3428944` and `.cartulary/test-results/20260731T190002Z-p3463705` exposed retry-editor lifecycle and missing optional-surface projection; stateful `.cartulary/test-results/20260731T191335Z-p3540624` and a11y `.cartulary/test-results/20260731T191912Z-p3593266` exposed obsolete raw-code/malformed envelope fixtures; visual `.cartulary/test-results/20260731T192439Z-p3659787` confirmed preserving existing public copy was required; typecheck `.cartulary/test-results/20260731T193229Z-p3735637` exposed optional `apiBase` typing. All were structurally corrected without loosening validation or changing a golden | Added two production paths, updated authored source ownership, Workbook OpenAPI/change-set and protocol tests, regenerated OpenAPI/protocol bindings/validators and execution topology only through Make; 356 manifest paths equal 356 live paths | Existing routes, bodies, transaction identity, collaboration behavior, focus, scroll, inspector, DOM, selectors, keyboard behavior, accessibility semantics, and public visual copy remain. Optional adopted standardized Workbook surfaces now validate instead of failing closed; retry-admitted edits close while their memory-local semantic units remain queued; malformed legacy envelopes remain intentionally unsupported. No compatibility wrapper, fallback decoder, allowlist, or golden update was added. Roll back the two closure files, root/hook/port wiring, boundary/docs/accounting, authored protocol enum/change-set/test projection and generated outputs, retry settlement/test, and strict browser fixtures as one unit. No residual S-14 blocker |
| V-01 | DONE | F-01, S-01-S-14 | Completed terminal audits and the full final validation matrix; repaired strict-contract test fixtures without weakening production validation; published the final handoff in Section 19.4 | PASS evidence recorded in Section 19.4; tracker lint `.cartulary/test-results/20260731T204619Z-p229649`; `git diff --check` exit 0 | 356 manifest paths equal 356 live paths; generated drift and policy are clean; normative owners are unchanged | Valid behavior remains compatible; malformed legacy wire data remains intentionally unsupported; no residual blocker or compatibility exception |

### 19.3 Binary completion criteria

- Every ledger row is `DONE`; no unresolved `IN PROGRESS`, `NOT STARTED`, or
  `BLOCKED` row remains.
- Every in-scope Workbook operation has a generated binding and runtime success
  validator; malformed responses fail closed.
- No presentation/workflow owner interprets HTTP status, routes, raw envelopes,
  or storage details.
- Timeline pending units are semantic and fresh/replay execution shares one
  request path while preserving FIFO and transaction identity.
- TimelineWorkbook is composition and rendering, not the owner of transport,
  replay policy, collaboration subscription, draft registry, or mutation
  admission.
- No compatibility wrapper, fallback decoder, temporary allowlist, global
  store, duplicate serializer, generated hand edit, or undocumented owner
  exception remains.
- Source ownership, authored test families, and generated topology match the
  final tree exactly.
- Normative Core specifications and docs/domain.md remain unchanged unless an
  actual contradiction is recorded and resolved under owner authority.
- Focused owner evidence, final browser/accessibility/visual evidence,
  `make agent-finalize`, and the broad final target are recorded in V-01.

### 19.4 V-01 final validation and handoff

#### Planning and implementation summary

The authorized program executed in the required order from P-00 through S-14.
It added a generated, fail-closed Workbook protocol foundation before moving
transport interpretation behind semantic ports. Generic, Party, Entity,
Assessment, Evidence, coordination, and Timeline workflows now consume
owner-specific outcomes. Timeline query, replay, action, collaboration, bulk,
draft, mutation-admission, and continuity behavior is owned below the
composition root. `TimelineWorkbook.tsx` composes those owners and renders
their snapshots and commands.

No Core NLSpec, `docs/domain.md`, server route, WebSocket, database, saved-view,
or domain-vocabulary contract changed. C-008 through C-010 remain intentional
closed guardrails. The first V-01 search was broader than the remediation
scope and rediscovered preserved saved-view, startup, reference-option, and
non-Timeline query owners. The final absence audit is scoped to C-001 through
C-007 and all Timeline workflow owners, while the existing boundary, cycle,
browser, accessibility, and visual evidence revalidates C-008 through C-010.

#### Final tree and accounting

- The worktree contains 89 tracked-path changes and 38 new paths: 127 changed,
  added, or deleted path entries in total. The tracked diff contains 7,435
  insertions and 5,195 deletions. Exact production/test additions, moves,
  deletions, owner changes, and rollback units are recorded in the workstream
  rows above.
- V-01 itself changed only six existing Workbook characterization files plus
  this tracker: `WorkbookShell.actionSequencing.test.tsx`,
  `WorkbookShell.gridProvenance.test.tsx`,
  `WorkbookShell.inspector.test.tsx`, `WorkbookShell.saveState.test.tsx`,
  `WorkbookShell.support.test.tsx`, and
  `WorkbookShell.timelineQuery.test.tsx`.
- `tools/frontend_source_ownership.json` contains 356 source paths and the live
  `apps/web/src` TypeScript tree contains the same 356 paths, with no missing,
  extra, or duplicate projection.
- Authored protocol, OpenAPI, ownership, import-boundary, and test-family inputs
  were regenerated through Make. Generated protocol/OpenAPI/topology outputs
  are clean under drift and generated-artifact policy checks. No lockfile or
  generated file was hand-edited.
- `git diff` against baseline
  `c22ce208a176c595f74acdae5b46b3a947b6fb72` is empty for
  `docs/domain.md` and `docs/spec/**`.

#### Terminal audit results

| Audit | Result | Evidence |
| --- | --- | --- |
| Generated operation binding and validator coverage | PASS | `package.protocol_ts` 5/5 plus generator fail-closed coverage; generated drift and artifact policy pass |
| Raw transport above remediated adapters | PASS | No `readEnvelope`, `fetchWorkbookJSON`, `apiPath`, raw-result, protocol-envelope, or legacy service import in C-001-C-007 presentation/workflow owners or Timeline owners above adapters |
| Semantic replay | PASS | Pending replay/state/port files contain no path, method, URL, or transport coordinate; fresh and replay execution converge on the generated operation executor through the pending mutation adapter |
| Compatibility burden | PASS | No fallback/legacy decoder, compatibility facade, forwarding re-export, temporary allowlist, or duplicate Timeline wire serializer exists |
| Dependency direction | PASS | No service-to-Workbook back edge, Timeline adapter React/DOM dependency, root transport/queue-policy/subscription import, application `react-data-grid` import, or production cycle is present |
| Ownership/accounting | PASS | Source ownership parity is 356/356; authored test families and generated topology are drift-clean |
| Normative authority | PASS | Core specifications and `docs/domain.md` are unchanged; no owner contradiction was discovered |

#### Final verification evidence

| Verification | Result and run root |
| --- | --- |
| Final formatting and complete frontend unit suite | PASS: format `.cartulary/test-results/20260731T202159Z-p4138801`; frontend unit 768/768 `.cartulary/test-results/20260731T202207Z-p4142018` |
| Generated and shape checks | PASS: generate-drift `.cartulary/test-results/20260731T202246Z-p4144234`; generated policy `.cartulary/test-results/20260731T202246Z-p4144294`; JSON shape `.cartulary/test-results/20260731T202245Z-p4144130` |
| Frontend architecture and build | PASS: import boundary `.cartulary/test-results/20260731T202246Z-p4144752`; typecheck `.cartulary/test-results/20260731T202246Z-p4144771`; Biome `.cartulary/test-results/20260731T202246Z-p4144852`; build-web `.cartulary/test-results/20260731T202318Z-p4151284` |
| Protocol and architecture owners | PASS: `package.protocol_ts` 5/5 `.cartulary/test-results/20260731T195401Z-p3946769`; `web.architecture` 11/11 `.cartulary/test-results/20260731T195406Z-p3947100` |
| Workbook, Timeline, and Entity owners | PASS: `web.workbook` 114/114 `.cartulary/test-results/20260731T202318Z-p4150858`; `module.timeline` 56/56 `.cartulary/test-results/20260731T202546Z-p18341`; `module.entities` 23/23 `.cartulary/test-results/20260731T202318Z-p4150853` |
| Assessment, Evidence, and coordination owners | PASS: `module.assessments` 15 tests `.cartulary/test-results/20260731T195846Z-p4005764`; `module.evidence` 45 tests `.cartulary/test-results/20260731T195920Z-p4024498`; `module.tasksdecisions` 3 tests `.cartulary/test-results/20260731T200109Z-p4052469` |
| Collaboration and grid owners | PASS: collaboration 3 tests `.cartulary/test-results/20260731T200143Z-p4070595`; grid adapter 27 tests `.cartulary/test-results/20260731T200150Z-p4071137` |
| Agent finalization | PASS: `.cartulary/test-results/20260731T202720Z-p44814`; generated output unchanged; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| Broad fast suite | PASS: 1,002 tests, zero failed or missing, `.cartulary/test-results/20260731T202743Z-p51010` |
| Webserver-backed browser | PASS: 2/2 work units, `.cartulary/test-results/20260731T203036Z-p111248` |
| Stateful browser | PASS: 2/2 work units, `.cartulary/test-results/20260731T203545Z-p140874` |
| Accessibility browser | PASS: 2/2 work units, `.cartulary/test-results/20260731T203805Z-p164957` |
| Measurement browser | PASS: 2/2 work units, `.cartulary/test-results/20260731T203928Z-p185905` |
| Visual browser | PASS: 2/2 work units, existing goldens unchanged, `.cartulary/test-results/20260731T204025Z-p205962` |
| Final tracker and diff gate | PASS: Markdown `.cartulary/test-results/20260731T204619Z-p229649`; `git diff --check` exit 0; repeated after the binary `DONE` update |

#### Failures resolved during V-01

The initial `make test-fast` root
`.cartulary/test-results/20260731T200316Z-p4099535` and isolated reruns
`.cartulary/test-results/20260731T200506Z-p4109673`,
`.cartulary/test-results/20260731T200707Z-p4116749`, and
`.cartulary/test-results/20260731T200859Z-p4119467` exposed 25 existing raw
fixtures that no longer satisfied the generated validators. Contract-valid
UUIDs, mutation-only row shapes, required history paging, stable UI-contract
selector factories, and an asynchronous history-action wait repaired the
fixtures. The final frontend-unit and `test-fast` roots above prove the repair;
production validation was not loosened.

A diagnostic V-01 run launched three independent owner schedulers in parallel.
Its Timeline root `.cartulary/test-results/20260731T202318Z-p4150846` had all
non-support rows pass but one keyboard support assertion fail under competing
browser load. The canonical isolated Timeline rerun passed all 56 tests at
`.cartulary/test-results/20260731T202546Z-p18341`; the final browser targets
also ran sequentially and passed. No product change or waiver resulted from
the diagnostic failure.

#### Compatibility, residual risk, and rollback

Valid routes, request bodies, transaction identity, refresh order,
collaboration, selection, focus, scroll, inspector, keyboard, accessibility,
and required visual copy remain compatible. Malformed, obsolete, or
context-inconsistent wire responses now fail closed as `invalid_contract`.
Optional adopted standardized Workbook surfaces validate through the generated
protocol projection. Party partial completion, Entity survivor continuity,
Assessment draft retention, Evidence opaque-handle security, bounded upload
retry, semantic replay, and high-water admission remain explicitly tested.

There is no residual blocker or unowned compatibility exception. C-008 through
C-010 remain closed guardrails rather than deferred remediation. Operational
risk is limited to the normal need for coordinated client/server protocol
versions and the existing memory-local nature of pending Timeline replay.

Rollback is slice-granular in reverse order. V-01's fixture corrections can be
reverted independently only if strict response validation is also removed;
otherwise they are required contract evidence. Protocol owner inputs,
generator changes, generated outputs, operation executor, and semantic adapter
consumers form one atomic rollback boundary. Each feature/controller slice and
its ownership/test-family/topology updates form the narrower rollback boundary
recorded in its ledger row. No database, server, persisted queue, or user-data
migration is required.

## 20. Authorized legacy and dead-code cleanup iteration

### 20.1 Authority, objective, and scope

Section 19 is immutable execution history. This section authorizes the next
Workbook refactoring iteration and is the sole execution ledger for that
iteration. It does not reopen, reinterpret, or replace any completed Section 19
row.

The primary objective is to remove caller-proven dead Workbook code, reduce
unnecessary export surface, and delete the legacy `services/workbookApi.ts`
transport seam. Scope extends outside `apps/web/src/workbook/**` only to the
adjacent Evidence, Network Flow, Import, browser-service, protocol-projection,
reachability, ownership, and test-accounting changes required to remove that
seam cleanly.

Repository-wide Fallow findings without demonstrated coupling to that seam are
out of scope. Core NLSpecs, `docs/domain.md`, server behavior, database and
WebSocket contracts, persisted data, and domain vocabulary are also out of
scope. No finding in an out-of-scope area becomes an exit blocker for this
iteration merely because it appears in the same advisory report.

### 20.2 Planning baseline

| Fact | Recorded value |
| --- | --- |
| Baseline HEAD | `6c412506ad3adf2c9aa3fa2ccbe9afd058d94a45` |
| Baseline worktree | Clean before this tracker-only update |
| Baseline command | `make frontend-fallow-static` |
| Baseline result | PASS |
| Run root | `.cartulary/test-results/20260731T205305Z-p235622` |
| Dead-code report | `.cartulary/test-results/20260731T205305Z-p235622/frontend-fallow-static/fallow/dead-code.json` |
| Report SHA-256 | `60c3b026f9ce364bfe23da23d5f0993ce75c7dcbf92925cb9bcb50b82632a18a` |
| Fallow version and schema | Fallow 2.93.0; dead-code schema 7 |
| Repository findings | 127 total: 8 unused files, 64 unused exports, 35 unused types, 1 unused dependency, 15 unused class members, 1 duplicate export, and 3 circular dependencies |
| Workbook findings | 39 total: 12 unused exports, 13 unused types, and 14 unused class members; no unused Workbook file |

The 14 class-member diagnostics are report facts, not pre-approved retention.
The LC-P00 caller audit found direct or structurally typed callers for 12. It
found no caller for `WorkbookCollaborationCoordinator.reconnect` or
`WorkbookCollaborationCoordinator.presenceForRow`; those two are therefore
LC-01 removal candidates. This corrects the planning assumption that all 14
were analyzer-only false positives. Execution MUST NOT invent caller evidence
to preserve a declaration.

### 20.3 Binding cleanup decisions

1. A Fallow finding is a discovery input, not deletion authority. Every
   deletion requires a caller search, dynamic-entrypoint inspection when
   relevant, owner evidence, and focused validation.
2. Genuinely unused declarations are deleted. Locally useful declarations lose
   unnecessary `export` modifiers rather than being deleted or moved solely to
   satisfy the analyzer.
3. Live class members remain without inline Fallow suppressions. Retained
   exceptions record their callers, owner, continuing value, and removal
   trigger in Section 20.6.
4. `services/workbookApi.ts` is deleted after its final caller migrates. It is
   not renamed, wrapped, re-exported, or retained as a compatibility facade.
5. No fallback decoder, forwarding re-export, duplicate serializer, temporary
   allowlist, global store, or message-text authorization classifier is added.
6. The `ajv` runtime dependency remains. Generated protocol validators import
   its runtime helpers; the reachability model is corrected instead of deleting
   or broadly suppressing that dependency.
7. Generated roots are changed only through Make-owned generators. Source
   ownership and authored test-family inputs change in the same slice as any
   production or test path addition, move, or deletion.
8. Valid routes, request bodies, transaction identity, refresh order,
   lifecycle behavior, saved-view behavior, focus and scroll continuity,
   accessibility, and visible copy remain compatible. Malformed or obsolete
   responses continue to fail closed.
9. No Core NLSpec, `docs/domain.md`, server, database, WebSocket, or persisted
   data change is authorized. An actual owner contradiction sets the active
   row to `BLOCKED` and stops execution.
10. Fallow remains advisory. This iteration improves the accuracy and
    usefulness of its evidence but does not promote it to a repository gate.

### 20.4 Gap and risk register

| Gap | Remediation areas | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- |
| LC-G001: reachability evidence misclassifies runtime entrypoints and `ajv` | Reachability owner, Fallow configuration, harness tests | Live tools or runtime dependencies can be deleted while real dead code remains obscured by noise | Spawned entrypoints are reachable, `ajv` is not reported unused, no broad dependency ignore is added, and static-analysis checks pass |
| LC-G002: Workbook exposes dead or file-local symbols | Implementation, tests, source ownership | Public surface and refactoring cost continue to grow without consumers | Every one of the 12 export, 13 type, and 14 class-member findings is deleted, internalized, or retained with caller evidence |
| LC-G003: `workbookApi.ts` exposes unchecked envelopes and raw HTTP results | Protocol projection, adapters, implementation, tests, README | Invalid wire data and transport details can re-enter workflow state; future operations duplicate legacy patterns | Zero imports of the service; the file and obsolete test are deleted; semantic ports own all former Workbook consumers |
| LC-G004: query and pending-mutation requests have duplicate execution paths | Implementation, tests | Fresh/replayed behavior, response validation, aborts, and conflict handling can drift | Shared query and pending-mutation ports are the sole Workbook paths for their operations |
| LC-G005: accounting can retain deleted or moved paths | Ownership, test-family inputs, generated topology, tracker | Verification can silently omit the final tree or continue naming removed files | Live source paths exactly match ownership; generated topology is drift-clean; each ledger row contains exact accounting evidence |

Principal execution risks are accidental behavior change during transport
migration, loss of a dynamically invoked entrypoint, stale async results being
accepted after lifecycle invalidation, loss of pending transaction identity,
and deletion of a symbol that is consumed through a structural type. Focused
characterization precedes deletion wherever current evidence is not already
direct and unambiguous.

### 20.5 Interfaces and dependency rules

The iteration introduces or promotes these private Workbook boundaries:

- `WorkbookViewQueryPort`, returning a validated
  `WorkbookOperationOutcome<T>` or the distinct `aborted` result.
- Narrow startup, incident, membership, preference, and saved-view ports with
  owner-specific accepted DTOs.
- A shared `WorkbookPendingMutationPort` used by both Timeline replay and
  `WorkbookMutationRuntime`.
- The existing `WorkbookOperationOutcome<T>` failure algebra; no second
  catch-all result type is introduced.

The operation executor moves atomically from
`workbook/mutations/workbookOperationExecutor.ts` to
`workbook/adapters/workbookOperationExecutor.ts`. All callers migrate in the
same slice, and the former path is deleted without a forwarding export.

Dependency direction remains:

`surface -> controller -> semantic port -> adapter -> browser service/protocol`

Controllers and presentation do not import `workbookApi`, `httpTransport`, raw
protocol envelopes, or manually constructed projected routes. Adapters do not
import React, DOM renderers, or feature state. Query adapters validate incident
and view-schema identity before accepting rows. Stateful consumers retain
generation or abort guards for incident, schema, selection, authorization, and
disposal changes.

### 20.6 Finding disposition inventory

#### 20.6.1 Delete in LC-01

- `workbook/collaboration/workbookCollaborationMessages.ts`:
  `shouldIgnoreSelfOriginatedRecordChange`.
- `workbook/timeline/components/TimelineWorkbookStyles.ts`:
  `timelineNoticeOverlayStyle`, `panelStyle`, `eyebrowStyle`, and
  `headlineStyle`. Direct caller search confirmed that the three style exports
  were not file-local after all, so deletion is cleaner than the planned
  internalization.
- `workbook/features/NetworkFlowFeature.tsx`: the forwarding exports
  `networkAnalysisSheetRef`, `networkAnalysisWorkspaceKey`, and
  `networkFlowActivityProfileId`; canonical extension identities remain.
- `WorkbookCollaborationCoordinator.reconnect`; no coordinator caller was
  found. Internal calls to `session.reconnect()` do not consume this wrapper.
- `WorkbookCollaborationCoordinator.presenceForRow`; Timeline derives row
  presence from `activeSheetPresenceRecords` and no coordinator caller exists.

#### 20.6.2 Internalize in LC-01

- `WorkbookPresenceMarkers.tsx`: `WorkbookRowPresenceMarker`.
- `assessmentWorkbookModel.ts`: `isAssessmentSubjectType`.
- `workbookOperationExecutor.ts`: `workbookOperationIDs` during the atomic
  LC-F02 executor move. LC-01 accounts for the baseline finding as already
  disposed by its prerequisite rather than reopening the moved adapter.
- `workbookConflictModel.ts`: `isWorkbookCollectionValue` and the
  `WorkbookCollectionValue` type that became independently visible after the
  guard was internalized.
- `useWorkbookQueryController.ts`: `WorkbookActiveQueryControls` and
  `WorkbookQueryStateSetter`.
- `useWorkbookStartupController.ts`: `ApplyWorkbookIdentityOptions`.
- `workbookSurfaceRegistration.ts`: `WorkbookSurfaceOwnerId`,
  `WorkbookSurfacePolicy`, and `WorkbookSurfaceRenderer`.
- `workbookSurfaceRegistry.ts`: `WorkbookSurfaceKind` and
  `WorkbookSurfaceStatus`.
- `useTimelineBulkTagController.ts`: `TimelineBulkTagMessage`.
- `useTimelineEditorDraftRegistry.ts`: `TimelineInputIdentity` and
  `TimelineScalarEditorIdentity`.
- `useTimelineSaveStatePresentation.ts`: `TimelineSaveStateLabel`.
- `useWorkbookQueryState.ts`: `WorkbookQueryStateEntry`.

#### 20.6.3 Retain with caller evidence

| Owner and members | Caller evidence | Continuing value | Removal trigger |
| --- | --- | --- | --- |
| `WorkbookCollaborationCoordinator.getSnapshot`, `subscribe` | `useWorkbookCollaborationCoordinator.ts` and `useTimelineCollaborationBindings.ts` consume them through `Pick` projections and external-store subscription | Stable collaboration snapshot delivery to Workbook and Timeline | Remove only when both consumers move to a replacement snapshot subscription owner |
| `WorkbookCollaborationCoordinator.retain`, `setActiveSheet` | `useWorkbookCollaborationCoordinator.ts` retains coordinator lifetime and applies the active sheet | Owns subscription lifetime and sheet-scoped presence admission | Remove only when coordinator lifecycle or sheet admission moves atomically to another owner |
| `WorkbookCollaborationCoordinator.registerClientTxnResolver` | `useTimelineCollaborationBindings.ts` registers Timeline transaction resolution | Connects collaboration acknowledgements to pending mutation identity | Remove only when pending transaction resolution no longer uses collaboration events |
| `WorkbookPendingQueueModel.clearSameFieldConflict` | `WorkbookMutationRuntime.ts` and `workbookPendingQueue.test.ts` | Clears an owner-resolved conflict without collapsing queue partitions | Remove only with deletion of same-field conflict recovery |
| `WorkbookPendingQueueModel.discardHaltedUnit`, `retryHaltedWithNewClientTxnId` | `WorkbookMutationRuntime.ts`, `useTimelinePendingReplayController.ts`, and queue characterization tests | Implements explicit discard and transaction re-key recovery | Remove only when the matching user recovery command is removed under owner authority |
| `WorkbookPendingQueueModel.pauseForAuthRecovery`, `pauseForTerminalLifecycle`, `resumeAfterAuthRecovery` | `WorkbookMutationRuntime.ts`, `WorkbookCollaborationCoordinator.ts`, Workbook collaboration tests, and queue characterization tests | Preserves the normative auth pause/requery and terminal lifecycle behavior | Remove only when the pending queue lifecycle state machine is replaced with equivalent owner-approved behavior |
| `WorkbookPendingQueueModel.settleDispatched` | `WorkbookMutationRuntime.ts`, `useTimelinePendingReplayController.ts`, and extensive queue tests | Centralizes accepted, retryable, conflict, and terminal settlement | Remove only when dispatch settlement moves atomically to a replacement queue model |

No inline suppression is authorized for these retained members. LC-V01 reruns
caller searches so a member that becomes genuinely unused during this
iteration is deleted rather than carried as a stale exception.

### 20.7 Ordered workstream ledger

Execution order is strict and linear:

`LC-P00 -> LC-F01 -> LC-F02 -> LC-01 -> LC-02 -> LC-03 -> LC-04 -> LC-05 -> LC-06 -> LC-07 -> LC-V01`.

| Workstream | Status | Depends on | Substantive remediation and exit criteria | Required owner evidence | Risk and rollback boundary |
| --- | --- | --- | --- | --- | --- |
| LC-P00 - tracker authorization | DONE | none | Preserved Section 19; added Section 20 scope, baseline, binding decisions, gap and risk register, interfaces, complete finding disposition, ordered ledger, mandatory gate, validation, compatibility, rollback, and deferrals. Corrected the unsupported assumption that all 14 class-member findings were live. The execution amendment assigns `workbookOperationIDs` internalization to the atomic LC-F02 executor move and requires saved-view state changes only after accepted responses; no optimistic list mutation is introduced. Non-Workbook Fallow findings remain routed follow-up work rather than hidden exceptions or blockers for this bounded iteration. No production, test, contract, generated, manifest, lockfile, normative, or other documentation file changed | Baseline Fallow PASS at `.cartulary/test-results/20260731T205305Z-p235622`; original Markdown PASS at `.cartulary/test-results/20260731T211021Z-p242593`; execution-amendment Markdown PASS at `.cartulary/test-results/20260731T213144Z-p258280`; `git diff --check` exit 0 after both tracker versions | Documentation-only rollback is the Section 20 addition and its execution amendment. No product behavior changes |
| LC-F01 - reachability truth | DONE | LC-P00 | Added the six runtime-spawned browser/owner-slice CLIs to existing `harness_entrypoints.files`; added `tools/protocol-ts/generate-protocol-types.mjs` to the authored `generate-artifacts` backing scripts; removed only the generated protocol root from Fallow's global ignores while retaining generated-symbol overrides; regenerated `tools/task_surface_manifest.json` and `tools/execution_topology_render_index.json` through Make. Updated stale harness catalog-count assertions exposed by the new current-catalog validation. No schema, NLSpec, lockfile, broad ignore, suppression, or runtime behavior changed | PASS: generate `.cartulary/test-results/20260731T213256Z-p261854`; Fallow `.cartulary/test-results/20260731T213314Z-p264351` reports 119 issues instead of 127, one unused file instead of seven, zero unused dependencies, and no named entrypoint/`ajv` finding; harness contract `.cartulary/test-results/20260731T213512Z-p272763`; JSON `.cartulary/test-results/20260731T213406Z-p266497`; generate-drift `.cartulary/test-results/20260731T213406Z-p266504`; generated policy `.cartulary/test-results/20260731T213406Z-p266469`; `web.architecture` 11/11 `.cartulary/test-results/20260731T213558Z-p274379`; boundary `.cartulary/test-results/20260731T213558Z-p274461`; Markdown `.cartulary/test-results/20260731T213635Z-p276287`; direct report audit found no named entrypoint or `ajv`; `git diff --check` exit 0. Initial harness roots `.cartulary/test-results/20260731T213314Z-p264362` and `.cartulary/test-results/20260731T213406Z-p266741` exposed stale expected family, row, and selector totals; assertions were updated to the catalog's validated 60 owners, 216 families, 1006 rows, and 1871 selectors before the passing rerun | Runtime tools and generated validator dependencies remain live with no package migration. The remaining single unused file and other non-Workbook diagnostics are the explicit Section 20.11 deferrals. Roll back `.fallowrc.json`, reachability/task owners, harness count assertions, and both generated accounting files as one unit |
| LC-F02 - protocol completion | DONE | LC-F01 | Added the six existing OpenAPI operations to `contracts/protocol-ts/http-operations.v1.json` and generated closed request/response aliases, paths, methods, query parameters, validators, and core types. Added exact facade path/query/method plus malformed-response tests. Moved the executor to `workbook/adapters/workbookOperationExecutor.ts`, internalized `workbookOperationIDs`, updated all ten direct importers, the protocol-adapter boundary, and authored source ownership, and left no forwarding export or former path. Generated outputs changed only through Make | PASS: final generate `.cartulary/test-results/20260731T214324Z-p350770`; format `.cartulary/test-results/20260731T214022Z-p284548`; protocol-ts 5/5 `.cartulary/test-results/20260731T214058Z-p290262`; Workbook 114/114 `.cartulary/test-results/20260731T214121Z-p292408`; incidents 8/8 work units and 40 tests `.cartulary/test-results/20260731T214121Z-p292418`, service-backed 8/8 and 25 tests `.cartulary/test-results/20260731T214356Z-p362095`; saved views 17/17 `.cartulary/test-results/20260731T214121Z-p292415`, service-backed 13/13 `.cartulary/test-results/20260731T214356Z-p362089`; final architecture 11/11 `.cartulary/test-results/20260731T214338Z-p353209`; typecheck `.cartulary/test-results/20260731T214058Z-p290360`; boundary `.cartulary/test-results/20260731T214058Z-p290392`; Biome `.cartulary/test-results/20260731T214058Z-p290422`; build-web `.cartulary/test-results/20260731T214338Z-p353637`; generate-drift `.cartulary/test-results/20260731T214338Z-p353126`; generated policy `.cartulary/test-results/20260731T214338Z-p353148`; JSON `.cartulary/test-results/20260731T214356Z-p362000`; Fallow `.cartulary/test-results/20260731T214356Z-p362251` reports 118 issues and removes the operation-ID export finding; Markdown `.cartulary/test-results/20260731T214615Z-p411184`; former-path audit and `git diff --check` pass. Initial generate `.cartulary/test-results/20260731T213919Z-p280446` rejected unsorted operation IDs; the projection was sorted. Initial architecture `.cartulary/test-results/20260731T214121Z-p292402` rejected an out-of-order ownership path; the authored list was corrected and regenerated | Valid projected routes and existing mutation behavior remain compatible; malformed success payloads fail closed. No server, owner-specification, data, or external API migration exists. Roll back the authored operation projection, generated protocol files, executor move/imports, facade tests, boundary/ownership inputs, and generated topology atomically |
| LC-01 - Workbook symbol cleanup | DONE | LC-F02 | Accounted for all 39 baseline findings: LC-F02 internalized the operation-ID set; this slice deleted the unused collaboration helper, two coordinator methods, three Network Flow forwarding exports, the Timeline notice overlay and three uncalled style exports; internalized all remaining file-local functions/types and removed redundant surface type re-exports. Internalizing the conflict guard exposed its collection-value type as a cascade, so that type was internalized too. Retained exactly the 12 coordinator/queue members in Section 20.6.3 after fresh caller evidence. No suppression, facade, runtime path, ownership path, or generated artifact was added | PASS: final format `.cartulary/test-results/20260731T215024Z-p422177`; Workbook 114/114 `.cartulary/test-results/20260731T215107Z-p431351`; Timeline 9/9 work units and 56 tests `.cartulary/test-results/20260731T215107Z-p431359`; collaboration 3/3 `.cartulary/test-results/20260731T215107Z-p431370`; boundary `.cartulary/test-results/20260731T214944Z-p419750`; final typecheck `.cartulary/test-results/20260731T215037Z-p425660`; final Biome `.cartulary/test-results/20260731T215038Z-p425714`; build-web `.cartulary/test-results/20260731T215038Z-p425985`; Fallow `.cartulary/test-results/20260731T215038Z-p425715` reports 92 global issues and only the 12 caller-proven Workbook members; direct caller audit found 2-29 references per retained member; Markdown `.cartulary/test-results/20260731T215322Z-p464228`; `git diff --check` exit 0. Initial typecheck `.cartulary/test-results/20260731T214944Z-p419726` and Biome `.cartulary/test-results/20260731T214944Z-p419797` exposed the now-obsolete registry type imports; both were removed before passing reruns | Runtime behavior and canonical extension identities are unchanged. The retained members continue to implement collaboration snapshot/lifetime/transaction resolution and normative pending-queue recovery. Roll back declaration removals/internalizations and their import cleanup together; no generated or data rollback exists |
| LC-02 - query seam convergence | DONE | LC-01 | Added `workbook/query/WorkbookViewQueryPort.ts`, `workbook/query/workbookLatestRequest.ts`, and `workbook/adapters/createWorkbookViewQueryAdapter.ts`. Workbook composition now creates one incident-bound adapter and injects it into Generic, Entity, Assessment, support-candidate, Timeline-preview, reference-broker, and Timeline consumers. The adapter alone executes `queryWorkbookView`, validates the generated envelope plus exact incident/schema identity, and normalizes rows through the requested view contract. Timeline now materializes its richer model only after the shared boundary and contains Timeline-specific materialization failures. Deleted the Timeline-specific adapter and port without forwarding exports; removed latest-request helpers from `services/workbookApi.ts`; updated source ownership, the Timeline test family, protocol-adapter policy, generated topology, and strict query fixtures. Searches prove one production query operation site, one latest-request helper definition, no raw query route, and no old Timeline query symbol in production/accounting inputs | PASS: final format `.cartulary/test-results/20260731T222632Z-p578118`; shared adapter row `.cartulary/test-results/20260731T220627Z-p488402`; final Workbook owner `.cartulary/test-results/20260731T222728Z-p582195`; Timeline owner `.cartulary/test-results/20260731T221217Z-p503137` and service-backed `.cartulary/test-results/20260731T222948Z-p607636`; Entities owner `.cartulary/test-results/20260731T221217Z-p503147` and service-backed `.cartulary/test-results/20260731T223120Z-p632962`; Assessments owner `.cartulary/test-results/20260731T222852Z-p587062` and service-backed `.cartulary/test-results/20260731T223301Z-p658436`; architecture `.cartulary/test-results/20260731T222933Z-p606395`; generate `.cartulary/test-results/20260731T223350Z-p676814`; boundary `.cartulary/test-results/20260731T223403Z-p679107`; typecheck `.cartulary/test-results/20260731T223407Z-p679668`; Biome `.cartulary/test-results/20260731T223418Z-p680365`; build `.cartulary/test-results/20260731T223434Z-p681142`; Fallow `.cartulary/test-results/20260731T223437Z-p684438` reports 87 global issues, down from 92, with no new in-scope dead-code finding; generate-drift `.cartulary/test-results/20260731T223550Z-p685501`; generated policy `.cartulary/test-results/20260731T223601Z-p689388`; JSON `.cartulary/test-results/20260731T223603Z-p689761`; absence audits and `git diff --check` pass. Initial focused query roots `.cartulary/test-results/20260731T220627Z-p488358`, `.cartulary/test-results/20260731T220627Z-p488355`, and `.cartulary/test-results/20260731T220627Z-p488375` exposed legacy non-UUID rows and incomplete generated envelopes; strict fixtures and the shared error-envelope helper were corrected. Initial broad roots `.cartulary/test-results/20260731T221217Z-p503141` and `.cartulary/test-results/20260731T221217Z-p503157` exposed stale symbolic IDs and incomplete Status Review, Evidence, Task, and Party row shapes; fixtures were brought to current contracts before passing reruns. The concurrent broad run also caused an assessment browser-readiness build collision, so final owner and service-backed gates ran sequentially | Valid requests preserve cancellation, late-result rejection, refresh/stale/access-loss behavior, dual-query admission, reference deduplication, and Timeline freshness/continuity semantics. Malformed or cross-incident/schema responses now intentionally fail closed. No server, owner-specification, external API, persisted-data, or compatibility migration exists. Remaining 87 Fallow findings are the routed Section 20.11 follow-up set, not new suppressions. Roll back the port, adapter, latest-request owner, all consumer/composition changes, deleted Timeline seam, strict fixtures, authored accounting, generated outputs, and this row as one unit |
| LC-03 - startup and identity cleanup | DONE | LC-02 | Added semantic `WorkbookStartupPort`, `WorkbookPreferencePort`, `WorkbookIncidentPort`, and shared `WorkbookPortResult`; added generated-operation adapters plus the adapter-result normalizer under `workbook/adapters`. Workbook composition now creates the three incident-bound adapters once. Startup admission accepts semantic URL query fields and validated selection/availability, incident identity and member-reference hooks accept the incident port, and current/default preference commands accept only correlated server acknowledgements. Controllers no longer import the legacy service, build these routes, inspect HTTP envelopes/statuses, or classify access loss from message text. Startup validates incident and extension-availability identity; identity/membership validate incident/resource correlation; preferences validate incident and exact sheet-ref correlation. Abort, latest-request, selection-version, startup fallback, late-result, membership cleanup, extension admission, and accepted-response preference behavior remain explicit. Added direct route/query/method/CSRF/malformed/cross-context/authorization/abort tests, strict contract fixtures, authored source ownership, a catalogued Workbook row, and regenerated topology. Searches find no raw transport, manual route, message-substring access decision, or legacy import in the migrated controllers | PASS: final format `.cartulary/test-results/20260731T230353Z-p829596`; adapter boundary row `.cartulary/test-results/20260731T230512Z-p848420`; final Workbook owner 115/115 `.cartulary/test-results/20260731T225817Z-p765861`; incidents 8/8 work units and 40 tests `.cartulary/test-results/20260731T225623Z-p722771`, service-backed 8/8 and 25 tests `.cartulary/test-results/20260731T225721Z-p745116`; saved views 17 tests `.cartulary/test-results/20260731T225955Z-p772330`, service-backed 13 tests `.cartulary/test-results/20260731T230137Z-p800366`; architecture 11 tests `.cartulary/test-results/20260731T225941Z-p771052`; boundary `.cartulary/test-results/20260731T230319Z-p827965`; typecheck `.cartulary/test-results/20260731T230319Z-p827961`; Biome `.cartulary/test-results/20260731T230353Z-p829586`; build `.cartulary/test-results/20260731T230408Z-p833367`; Fallow `.cartulary/test-results/20260731T230408Z-p833305` reports the same 87 routed global issues and no new in-scope finding; final generate `.cartulary/test-results/20260731T230436Z-p841580`; generate-drift `.cartulary/test-results/20260731T230446Z-p843842`; generated policy `.cartulary/test-results/20260731T230446Z-p843844`; JSON `.cartulary/test-results/20260731T230446Z-p843846`; Markdown `.cartulary/test-results/20260731T230624Z-p849301`; absence audits and `git diff --check` pass. Initial typecheck `.cartulary/test-results/20260731T224613Z-p698757` exposed incomplete shell/test port injection. Initial Workbook root `.cartulary/test-results/20260731T225038Z-p708732` exposed mismatched incident identity and incomplete incident, membership, preference, and startup saved-view fixtures; the fixtures were corrected to the projected contract. A narrow `frontend-unit` attempt was rejected because that target does not declare `VITEST_FLAGS`; catalogued `ROWS` were used instead, and the initially unhashed new row ID was corrected to the required ten-hex suffix. Initial Biome `.cartulary/test-results/20260731T230319Z-p827971` rejected direct test cookie assignment; the test now spies on the cookie getter. Initial generate-drift `.cartulary/test-results/20260731T230408Z-p833239` detected that the row-ID correction postdated generation; generated outputs were refreshed before the passing drift run | Existing valid startup precedence, fallback, selection-version behavior, late-result rejection, extension gating, incident presentation, membership cleanup, and preference intent are preserved. Malformed, mismatched, stale, or unauthorized results now fail closed through semantic outcomes. No server, owner-specification, external API, persisted-data, WebSocket, or compatibility migration exists. Roll back the three ports/adapters, result helper, composition/controller changes, strict fixtures/tests, authored ownership/test row, generated outputs, and this row together; do not restore a raw compatibility facade |
| LC-04 - saved-view cleanup | DONE | LC-03 | Added semantic `WorkbookSavedViewPort` and generated-operation `createWorkbookSavedViewAdapter`; Workbook composition creates the incident-bound adapter once and the controller no longer imports the legacy service, constructs routes, or observes HTTP envelopes/statuses. The adapter validates incident, saved-view, and schema correlation, paging shape/progress, versions, delete acknowledgements, and system-view immutability. Listing explicitly accumulates pages with seen-cursor and seen-resource guards and publishes only a complete accepted result. Create, duplicate, patch, and delete mutate list/selection state only after accepted responses; context-version guards reject late mutation results after incident/port replacement. Selection, query/layout restoration, fallback identity, version behavior, duplicate-name handling, access-loss semantics, and non-optimistic state changes remain explicit. Removed obsolete raw saved-view envelope types; added direct adapter and controller paging/correlation/late-result tests, strict shell fixtures, authored ownership and a catalogued adapter row, then regenerated topology. Searches find no saved-view route, raw envelope, legacy import, or HTTP mechanic above the adapter | PASS: final format `.cartulary/test-results/20260731T232307Z-p943400`; final adapter row `.cartulary/test-results/20260731T232311Z-p946531`; controller/adapter rows `.cartulary/test-results/20260731T231551Z-p873025`; surface row `.cartulary/test-results/20260731T231637Z-p877332`; Workbook 116/116 `.cartulary/test-results/20260731T231717Z-p881267`; saved views 17 tests `.cartulary/test-results/20260731T232040Z-p915344`; service-backed saved views 13 tests `.cartulary/test-results/20260731T232320Z-p946906`; architecture 11 tests `.cartulary/test-results/20260731T232503Z-p975595`; generate `.cartulary/test-results/20260731T231459Z-p866803`; boundary `.cartulary/test-results/20260731T232513Z-p976832`; final typecheck `.cartulary/test-results/20260731T232610Z-p987467`; final Biome `.cartulary/test-results/20260731T232621Z-p988057`; build `.cartulary/test-results/20260731T232531Z-p978609`; Fallow `.cartulary/test-results/20260731T232535Z-p981936` reports the same 87 routed global issues with no new finding; generate-drift `.cartulary/test-results/20260731T232547Z-p982587`; generated policy `.cartulary/test-results/20260731T232558Z-p986499`; JSON `.cartulary/test-results/20260731T232600Z-p986842`; Markdown `.cartulary/test-results/20260731T232748Z-p989252`; absence audits and `git diff --check` pass. Initial typecheck `.cartulary/test-results/20260731T231032Z-p857003` exposed readonly semantic values crossing into mutable generated requests and a stale test option; the adapter now owns explicit request cloning and the test uses the incident-bound port. Initial focused row `.cartulary/test-results/20260731T231511Z-p869120` exposed an unstable test port identity, and initial surface root `.cartulary/test-results/20260731T231601Z-p873528` exposed a fixture without projected paging; both fixtures were corrected. Initial saved-view owner root `.cartulary/test-results/20260731T231843Z-p886312` retained a failed browser-startup diagnostic even though its automatic retry passed the saved-view browser assertion; the clean sequential rerun is the passing `.cartulary/test-results/20260731T232040Z-p915344` root | Valid saved-view selection, restoration, fallback, versions, duplicate-name behavior, and accepted-response updates are preserved. Malformed, cyclic, duplicate-resource, mismatched, immutable-system-view, and late-context results now intentionally fail closed. No server, owner-specification, external API, persisted-data, WebSocket, or compatibility migration exists. Roll back the port/adapter, composition/controller changes, removed envelopes, strict fixtures/tests, authored ownership/test row, generated outputs, and this row together; do not restore a raw compatibility facade |
| LC-05 - pending mutation convergence | DONE | LC-04 | Added semantic `WorkbookPendingMutationPort`, incident-bound `createWorkbookPendingMutationAdapter`, and shared semantic-failure-to-queue settlement mapping. Workbook composition creates one adapter and injects it into both `WorkbookMutationRuntime` and Timeline. The adapter alone builds generated `createViewRow`/`patchRecord` requests, replaces payload transaction/version fields with queue/dispatch identity, rejects cross-incident or unknown-schema units, validates schema/record/change-set correlation, and normalizes accepted rows through the registered view contract. The common runtime no longer imports the legacy service, serializes PATCH, parses status/envelopes, or carries `apiBase` in queued requests/meta. Timeline consumes the same port, retains richer row/high-water materialization above it, and owns resolved-conflict row normalization locally. Deleted the Timeline-only pending adapter/port without forwarding exports; removed `apiBase` from Generic/Entity autosave inputs; updated direct adapter/runtime tests, strict fixtures, source ownership, the test-family row, and generated topology. Refactored the new boundary into focused execution/correlation/normalization helpers after static health review. Searches prove one pending create/patch executor, zero old pending path/import, zero raw runtime serializer/envelope parser, and no queued `apiBase` | PASS: final format `.cartulary/test-results/20260731T235238Z-p1105147`; final shared adapter row `.cartulary/test-results/20260731T235253Z-p1108886`; stale-high-water plus adapter `.cartulary/test-results/20260731T235152Z-p1103769`; final Workbook 117/117 `.cartulary/test-results/20260731T235322Z-p1110040`; final Timeline 9/9 work units and 55 tests `.cartulary/test-results/20260731T235445Z-p1115051`; final service-backed Timeline 9/9 and 33 tests `.cartulary/test-results/20260731T235617Z-p1142757`; final architecture 11 tests `.cartulary/test-results/20260731T235847Z-p1178264`; final boundary `.cartulary/test-results/20260731T235751Z-p1168351`; final typecheck `.cartulary/test-results/20260731T235755Z-p1168872`; final Biome `.cartulary/test-results/20260731T235806Z-p1169457`; final build `.cartulary/test-results/20260731T235809Z-p1170090`; final Fallow `.cartulary/test-results/20260731T235258Z-p1109271` reports the same 87 routed findings with no new dead-code or pending-boundary health finding; final generate `.cartulary/test-results/20260731T234349Z-p1023037`; generate-drift `.cartulary/test-results/20260731T235812Z-p1173355`; generated policy `.cartulary/test-results/20260731T235824Z-p1177240`; JSON `.cartulary/test-results/20260731T235826Z-p1177583`; Markdown `.cartulary/test-results/20260731T235957Z-p1179853`; absence audits and `git diff --check` pass. Initial format `.cartulary/test-results/20260731T233721Z-p996448` found an implicit contract type; the next format preflight `.cartulary/test-results/20260731T233737Z-p999781` found the deleted selector in stale topology. Initial generate `.cartulary/test-results/20260731T233830Z-p1003206` found the new row out of ASCII order. Typecheck roots `.cartulary/test-results/20260731T233911Z-p1011076` and `.cartulary/test-results/20260731T234009Z-p1015155` found strict Timeline fixtures and the wrong generated response alias; typed contracts and fixtures were corrected. Mistyped narrow row IDs produced usage-only roots `.cartulary/test-results/20260731T234058Z-p1016952` and `.cartulary/test-results/20260731T234401Z-p1028479`; catalogued IDs were then used. Initial full Workbook `.cartulary/test-results/20260731T234135Z-p1017725` showed that rejecting a structurally valid stale replay before Timeline freshness admission broke high-water preservation; adapter validation was narrowed to identity/contract correlation and the controller continues to reject the stale application. Fallow `.cartulary/test-results/20260731T234929Z-p1093976` retained the global count but exposed new complexity recommendations; helper extraction removed them before the final report | FIFO/capacity, contiguous coalescing, stable unit/client transaction IDs, dispatch-time versions, accepted settlement, stale-result retention, row re-keying, same-field conflict routing, retry, discard, and authorization/lifecycle pause-resume remain unchanged. Malformed, cross-incident/schema/record, empty-change-set, and undispatchable responses now fail closed at the shared boundary. No server, owner-specification, external API, persisted-data, WebSocket, or compatibility migration exists. Roll back the shared port/adapter/settlement helper, both consumers and composition, removed Timeline seam, queue-input cleanup, strict tests/fixtures, authored accounting, generated outputs, and this row together; do not restore a duplicate serializer or forwarding facade |
| LC-06 - adjacent consumer migration | DONE | LC-05 | Evidence blob-slot creation now calls projected `createObjectBlobSlot` through `services/browserApi.fetchHTTPOperation`, validates the generated success envelope plus requested/accepted incident, byte-size, filename, and content-type correlation, and reaches the server-issued upload target only after acceptance. Its direct upload still omits credentials, copies only string target headers, supplies a content type when absent, retries only network/503/504 outcomes within the existing bound, never reads response bodies, and sanitizes public errors. Network Flow now uses owner-neutral `fetchJSON` through profile-and-route-scoped `runProfileRequest`; `networkFlowRouteFamily` is owned with the extension identities, and existing URL construction, decoding, abort signals, CSRF, and structured error mapping remain above/below that boundary as appropriate. Imports removed `parseErrorMessage` and now derives safe status/reason text from `extractError` plus `publicErrorView` while retaining Import profile/route admission. Added direct malformed/cross-incident Evidence, Network Flow transport/gating/decoder, and unsafe Import error tests; catalogued the new rows, added `networkFlowClient.test.ts` to source ownership, and regenerated topology. Searches find no adjacent production import or use of `workbookApi`, its unchecked envelope cast, or a manual blob-slot operation path | PASS: final format `.cartulary/test-results/20260801T001816Z-p1311425`; focused Evidence `.cartulary/test-results/20260801T000810Z-p1198755`; focused Import `.cartulary/test-results/20260801T000818Z-p1199161`; focused Network Flow `.cartulary/test-results/20260801T000858Z-p1200306`; Evidence 45 tests `.cartulary/test-results/20260801T000913Z-p1200693` and service-backed 34 tests `.cartulary/test-results/20260801T001114Z-p1228938`; Imports 13/13 work units and 28 tests `.cartulary/test-results/20260801T001306Z-p1255151`, service-backed 13/13 and 17 tests `.cartulary/test-results/20260801T001357Z-p1276241`; Network Flow 31/31 `.cartulary/test-results/20260801T001446Z-p1296923`; architecture 11 tests `.cartulary/test-results/20260801T001743Z-p1309958`; boundary `.cartulary/test-results/20260801T001520Z-p1298787`; typecheck `.cartulary/test-results/20260801T001529Z-p1299343`; Biome `.cartulary/test-results/20260801T001559Z-p1300110`; build `.cartulary/test-results/20260801T001610Z-p1300784`; Fallow `.cartulary/test-results/20260801T001617Z-p1304144` reports the same 87 routed findings with no new in-scope finding; generate `.cartulary/test-results/20260801T000758Z-p1196478`; generate-drift `.cartulary/test-results/20260801T001647Z-p1304920`; generated policy `.cartulary/test-results/20260801T001706Z-p1308872`; JSON `.cartulary/test-results/20260801T001713Z-p1309258`; Markdown `.cartulary/test-results/20260801T001927Z-p1314966`; production-import/manual-route searches and `git diff --check` pass. Initial format preflights `.cartulary/test-results/20260801T000717Z-p1186737` and `.cartulary/test-results/20260801T000729Z-p1190018` rejected unknown collaborator IDs; only catalogued collaborators were retained. Initial Network Flow row `.cartulary/test-results/20260801T000825Z-p1199528` exposed a test fixture missing the standard `{data, meta}` HTTP envelope; the fixture now exercises malformed resources behind that envelope and the generated decoder | Valid Evidence creation/upload, Network Flow decoding/cancellation/CSRF/error behavior, and safe Import error detail remain compatible. Cross-incident or request-mismatched blob slots and unclaimed Network Flow routes now intentionally fail closed; unsafe Import server text is no longer surfaced. No server, owner-specification, external API, persisted-data, WebSocket, or compatibility migration exists. The 87 non-Workbook findings remain routed Section 20.11 work. Each adjacent consumer plus its direct tests is an independent code rollback unit; source ownership, authored test rows, generated topology, and this tracker update roll back with the affected unit, without restoring a forwarding facade |
| LC-07 - legacy seam deletion | DONE | LC-06 | Zero production callers were proven, then `services/workbookApi.ts` and its obsolete test were deleted without a forwarding export. Existing `browserApi` tests retain JSON/CSRF behavior and now own the private-timing-header absence check; new `shared/publicError.test.ts` directly owns allowlisted detail and unsafe-message behavior; new `workbook/query/workbookLatestRequest.test.ts` owns request exclusivity and supersession. Removed the four obsolete `workbookApi` catalog rows, added the replacement owner rows, removed deleted paths and added replacement tests in source ownership, and regenerated topology. Removed the unused `csrfCookieName` re-export and internalized the transport constant. Removed the stale `workbookApi` alternative from the Workbook source-ownership policy; no executor-path or import-boundary allowance for the deleted seam remained. `apps/web/src/README.md` now documents the common Workbook adapter, semantic port, shared query/latest-request, startup, pending-mutation runtime, Evidence, Imports, and profile-route-scoped Network Flow owners and no longer lists the deleted Timeline query/pending seams. A strict Evidence fixture now echoes the requested blob-slot accepted contract. Searches find no legacy file, import, helper, forwarding export, README entry, test selector, ownership entry, or generated-topology path | PASS: final format `.cartulary/test-results/20260801T003141Z-p1344708`; replacement application rows `.cartulary/test-results/20260801T002633Z-p1329829`; replacement latest-request row `.cartulary/test-results/20260801T002642Z-p1330294`; full application 53/53 `.cartulary/test-results/20260801T002656Z-p1330686`; corrected surface row `.cartulary/test-results/20260801T002952Z-p1337926`; final Workbook 118/118 `.cartulary/test-results/20260801T003016Z-p1339532`; architecture 11 tests `.cartulary/test-results/20260801T003002Z-p1338299`; boundary `.cartulary/test-results/20260801T003153Z-p1347944`; typecheck `.cartulary/test-results/20260801T003202Z-p1348508`; Biome `.cartulary/test-results/20260801T003220Z-p1349217`; build `.cartulary/test-results/20260801T003229Z-p1349885`; Fallow `.cartulary/test-results/20260801T003243Z-p1353868` reports 84 routed findings, down from 87 because three legacy unused exports disappeared, with no replacement-test or in-scope finding; generate `.cartulary/test-results/20260801T002530Z-p1320553`; generate-drift `.cartulary/test-results/20260801T003318Z-p1354637`; generated policy `.cartulary/test-results/20260801T003336Z-p1358566`; JSON `.cartulary/test-results/20260801T003345Z-p1358952`; Markdown `.cartulary/test-results/20260801T003458Z-p1359961`; terminal legacy-path/helper/export/accounting searches and `git diff --check` pass. Initial full Workbook `.cartulary/test-results/20260801T002725Z-p1332194` exposed an old surface fixture whose otherwise valid blob-slot response described a different incident; product code correctly rejected it before upload. The fixture now derives incident, size, filename, and content type from the request, after which the focused and complete owner reruns passed | Valid app transport, public-error presentation, latest-query supersession, Workbook behavior, and Evidence upload behavior remain compatible. The unused browser-level CSRF constant export and raw legacy helpers are intentionally removed; there is no compatibility layer. Cross-incident blob slots remain intentionally rejected. No server, owner-specification, external API, persisted-data, WebSocket, or migration change exists. The remaining 84 findings are routed Section 20.11 work. Roll back the two deleted files, replacement tests, private export cleanup, README, strict fixture, source ownership, authored catalog, generated topology, and this tracker row as one unit; do not restore the legacy file without also restoring all former callers from prior slices |
| LC-V01 - validation and handoff | DONE | LC-07 | Completed terminal absence and retained-caller audits, all planned focused and service-backed owner routes, common frontend/static/generated gates, `test-fast`, and all six requested browser targets. Added `module.workbook` and its service-backed route after the broad unit gate exposed that semantic startup tests are owned separately from `web.workbook`. V01 changed only strict fixtures and assertions: startup URL parsing now expects the semantic query object, the grid-provenance fixture uses a complete `active` incident plus an incident-correlated query envelope with wire-only rows, and the Timeline browser test expects the shared query boundary's error copy. No production, contract, generated, accounting, golden, normative, server, data, or compatibility-facade change was required in V01 | PASS focused: architecture `.cartulary/test-results/20260801T003843Z-p1369551`; protocol-ts `.cartulary/test-results/20260801T003858Z-p1370782`; web Workbook `.cartulary/test-results/20260801T003906Z-p1371148`; incidents `.cartulary/test-results/20260801T004029Z-p1376271` and service-backed `.cartulary/test-results/20260801T004147Z-p1398126`; saved views `.cartulary/test-results/20260801T004241Z-p1419134` and service-backed `.cartulary/test-results/20260801T004419Z-p1447013`; Timeline `.cartulary/test-results/20260801T004559Z-p1474424` and service-backed `.cartulary/test-results/20260801T004734Z-p1501006`; entities `.cartulary/test-results/20260801T004906Z-p1526305` and service-backed `.cartulary/test-results/20260801T005037Z-p1552349`; assessments `.cartulary/test-results/20260801T005207Z-p1577792` and service-backed `.cartulary/test-results/20260801T005248Z-p1596646`; Evidence `.cartulary/test-results/20260801T005327Z-p1615125` and service-backed `.cartulary/test-results/20260801T005518Z-p1641990`; Imports `.cartulary/test-results/20260801T005712Z-p1668258` and service-backed `.cartulary/test-results/20260801T005802Z-p1689161`; Network Flow `.cartulary/test-results/20260801T005852Z-p1709550`; final module Workbook 88 tests `.cartulary/test-results/20260801T010702Z-p1783652` and service-backed 56 tests `.cartulary/test-results/20260801T011400Z-p1874980`. PASS finalize/common: `agent-finalize` `.cartulary/test-results/20260801T005922Z-p1711396` recorded `RESULTS_DIR=-` and retained-run maintenance skipped; final format `.cartulary/test-results/20260801T012010Z-p1917098`; unit `.cartulary/test-results/20260801T012019Z-p1920299`; boundary `.cartulary/test-results/20260801T012055Z-p1922466`; typecheck `.cartulary/test-results/20260801T012102Z-p1923021`; Biome `.cartulary/test-results/20260801T012116Z-p1923646`; build `.cartulary/test-results/20260801T012126Z-p1924330`; Fallow `.cartulary/test-results/20260801T012136Z-p1928300`; generate-drift `.cartulary/test-results/20260801T012224Z-p1929247`; generated policy `.cartulary/test-results/20260801T012236Z-p1933153`; JSON `.cartulary/test-results/20260801T012238Z-p1933535`; `test-fast` 1002 tests `.cartulary/test-results/20260801T012245Z-p1934168`; browser support `.cartulary/test-results/20260801T012544Z-p2009055`; webserver-backed `.cartulary/test-results/20260801T012624Z-p2012467`; stateful `.cartulary/test-results/20260801T013116Z-p2042117`; accessibility `.cartulary/test-results/20260801T013339Z-p2066329`; measurement `.cartulary/test-results/20260801T013504Z-p2087418`; visual `.cartulary/test-results/20260801T013603Z-p2107668` with no golden update. Final Fallow schema-7 report has 90 total diagnostics, down from 127: its only 12 Workbook findings are the Section 20.6.3 members with fresh production callers; the other 78 are the routed Section 20.11 owner follow-ups. Terminal searches and `git diff --check` pass. Initial unit `.cartulary/test-results/20260801T010021Z-p1722382` exposed two stale query-string assertions plus an incomplete/cross-incident grid fixture; the first rerun `.cartulary/test-results/20260801T011718Z-p1913863` narrowed the remaining issue to the forbidden row-local schema field and invalid incident status. Initial module Workbook `.cartulary/test-results/20260801T010248Z-p1725940` exposed stale Timeline-specific error copy; its corrected row passed at `.cartulary/test-results/20260801T010625Z-p1765064`. Initial module Workbook service-backed `.cartulary/test-results/20260801T011012Z-p1821286` had zero product-row failures but one terminal browser-startup artifact; the isolated row passed at `.cartulary/test-results/20260801T011326Z-p1857011` before the clean full rerun | All application-private valid behavior remains compatible; malformed/cross-context results intentionally fail closed. No server API, database, persisted data, WebSocket, external consumer, or migration change exists, and no legacy compatibility layer remains. All LC-G001 through LC-G005 risks are closed with executable evidence. Residual risk is limited to the 78 explicitly deferred, source-owner-routed non-Workbook diagnostics. Roll back in reverse slice order with each slice's implementation, tests, authored inputs, generated outputs, and tracker evidence together; V01's four test-only fixture/assertion files are its independent rollback boundary. The post-entry Markdown and diff gate is recorded in Section 20.13 |

### 20.8 Mandatory tracker gate

After every implementation workstream and before the next begins:

1. Complete focused implementation and validation.
2. Update its ledger row with exact substantive changes, paths added, moved or
   deleted, commands and run roots, failures and resolution, compatibility
   decision, residual risk, and rollback boundary.
3. Update source ownership and authored test-family inputs in the same slice
   when paths change; regenerate downstream topology rather than editing it.
4. Run `make frontend-fallow-static` and record both the global report and the
   scoped disposition delta. A global out-of-scope finding is not silently
   waived or converted into an in-scope blocker.
5. Run `make lint-markdown` and `git diff --check` after the tracker update.
6. Begin the next row only when the current row is `DONE`. An owner
   contradiction changes the row to `BLOCKED` and stops execution.

### 20.9 Validation plan

Every frontend implementation slice runs, at minimum:

- `make format`, followed by diff inspection.
- `make frontend-import-boundary-check`.
- `make frontend-typecheck`.
- `make lint-biome`.
- The narrow owner slice selected through
  `make task-guide ROLE=module-author OWNER=<owner-id>`.
- `make build-web`.
- `make frontend-fallow-static`.
- `git diff --check`.
- `make lint-markdown` after its tracker update.

Slices changing generated, ownership, or test-family inputs also run:

- `make generate`.
- `make generate-drift`.
- `make generated-artifact-policy-check`.
- `make json-shape-check`.

Owner routing is:

- LC-F01: harness/static-analysis ownership and `web.architecture`.
- LC-F02: `package.protocol_ts`, `web.architecture`, `web.workbook`,
  `module.incidents`, and `module.savedviews`.
- LC-01: `web.workbook`, `module.timeline`, and collaboration ownership.
- LC-02: `web.workbook`, `module.timeline`, `module.entities`,
  `module.assessments`, and `web.architecture`.
- LC-03: `web.workbook`, `module.incidents`, and `web.architecture`.
- LC-04: `web.workbook`, `module.savedviews`, and `web.architecture`.
- LC-05: `web.workbook`, `module.timeline`, and `web.architecture`.
- LC-06: `module.evidence`, `module.imports`, `web.networkflow`, and
  `web.architecture`.
- LC-07: all directly affected owners plus `web.workbook` and
  `web.architecture`.

LC-V01 additionally runs:

- Absence searches for `services/workbookApi`, `readEnvelope`, raw Workbook
  `{ok,status,payload}` handling, manual projected routes, duplicate pending
  PATCH serializers, message-string access-loss classification, forwarding
  compatibility exports, and stale path accounting.
- Direct caller searches for all retained Section 20.6.3 members.
- `make frontend-unit`.
- All focused owner slices listed above.
- `make agent-finalize`; `RESULTS_DIR` is supplied only for a qualifying
  successful full warm-check root, otherwise retained-run maintenance is
  recorded as skipped.
- `make test-fast`.
- `make browser-e2e-webserver-backed`.
- `make browser-e2e-stateful`.
- `make browser-e2e-a11y`.
- `make browser-e2e-measurement`.
- `make browser-e2e-visual`, without a golden update unless a separately
  authorized visual change is discovered.

### 20.10 Compatibility, migration, and rollback

All changed modules are application-private. Removing an export or file is
allowed only after repository-wide caller evidence proves it unused. No public
package API, server route, database record, stored saved view, pending queue,
or user data requires migration.

Valid server responses and existing Workbook interactions remain compatible.
Malformed or contract-incompatible responses continue to become safe semantic
failures. `ajv` remains installed and lockfiles do not change for LC-F01.
Runtime-spawned tools remain executable after their reachability correction.

Rollback is slice-granular in reverse order. Generated protocol inputs,
outputs, and executor relocation are atomic. Query, startup, saved-view, and
pending-mutation ports roll back with their adapters, consumers, and focused
tests. Adjacent owner changes roll back independently. The final legacy-service
deletion rolls back with its test relocation, README, source ownership, test
families, and generated topology; it MUST NOT be restored without also
restoring every former caller in the same rollback.

### 20.11 Deferred findings

The following remain explicitly outside this iteration:

- Harness duplicate exports and circular dependencies reported by the baseline
  Fallow run.
- `tools/harness/observability/retained-reference-migration-cli.mjs`, which has
  no discovered caller but requires a harness-owner audit before deletion.
- Non-Workbook unused exports and types that are unrelated to removing
  `workbookApi.ts`.
- Any broader Network Flow, Import, Evidence, app-shell, or protocol-package
  cleanup not required by LC-06.

These findings are not silently accepted as permanent. LC-V01 records their
final report locations and routes them to future owner-specific cleanup rather
than expanding Workbook authority.

### 20.12 Binary completion and handoff

The iteration is complete only when:

- Every row LC-P00 through LC-V01 is `DONE`.
- `services/workbookApi.ts` and its obsolete test are absent and have zero
  imports.
- No unchecked Workbook envelope cast, raw Workbook HTTP result, manual route
  for a projected operation, duplicate pending serializer, or message-string
  access-loss classifier remains.
- The scoped Fallow result contains no actionable unused Workbook file,
  export, type, or class member. Any retained analyzer diagnostic has current
  caller evidence in Section 20.6.3.
- The seven dynamically invoked tools are modeled as reachable and `ajv` is
  recognized as a live protocol-validator runtime dependency.
- Source ownership, authored test families, and generated topology exactly
  match the final tree.
- No compatibility facade, forwarding re-export, fallback decoder, temporary
  allowlist, generated hand edit, or undocumented owner exception exists.
- All focused, broad frontend, browser, accessibility, measurement, visual,
  Markdown, and diff checks are green and recorded with exact run roots.
- The final tracker handoff records files changed, substantive edits,
  generated/accounting changes, resolved failures, compatibility impact,
  deferred findings, residual risks, and slice-granular rollback guidance.

### 20.13 Final evidence summary

All Section 20 rows are `DONE`, and every row was updated before its successor
began. The preserved staged tracker state remains staged; no user change was
reset, unstaged, or overwritten. The final worktree contains 131 changed paths:
98 modified paths including this staged-plus-unstaged tracker, 9 deletions, and
24 additions.

The substantive result is one fail-closed Workbook architecture:

- Runtime-spawned tools, the protocol generator, generated imports, and `ajv`
  are represented by the existing harness and reachability owners.
- The six adopted HTTP operations are projected and generated, and generated
  protocol or transport details terminate at Workbook adapter boundaries.
- Startup, incident, preference, saved-view, shared query, and shared pending
  mutation behavior is expressed through private semantic ports composed once
  by the Workbook shell.
- The legacy service, Timeline-specific query and pending seams, duplicate raw
  serializers, stale forwarding exports, and unused symbols are deleted rather
  than retained behind compatibility facades.
- Evidence, Imports, and Network Flow use their durable owner-neutral or
  profile-scoped boundaries with strict response correlation and sanitized
  errors.
- Authored source ownership and test families, generated topology and task
  surface, import boundaries, and the web source map describe the final tree.

The generated/accounting changes are confined to the authored HTTP operation,
reachability, task-surface, source-ownership, import-boundary, and test-family
inputs plus Make-generated protocol, task-surface, and topology projections.
No generated root was hand-edited. Generation drift, artifact policy, JSON
shape, owner routing, and the 1,002-test fast gate are clean.

The final Fallow report is
`.cartulary/test-results/20260801T012136Z-p1928300/frontend-fallow-static/fallow/dead-code.json`
with SHA-256
`e7e84a24ea5ddf716ba0258c72599f2c0a06f91ef992158c909c3a78f1e19586`.
Its 90 schema-7 diagnostics are fully dispositioned: 12 are the current,
caller-proven Workbook members in Section 20.6.3, and 78 are non-Workbook
follow-up work. The latter comprise one harness entrypoint audit, 50 unused
exports, 22 unused types, one extension class member, one duplicate export,
and three dependency cycles; their source owners retain cleanup authority.
There is no unused dependency, unresolved import, boundary violation, stale
suppression, unused catalog entry, or unresolved catalog reference.

Compatibility impact is private TypeScript cutover only. Valid response,
selection, cancellation, stale-result, saved-view, queue, transaction,
conflict, authorization, lifecycle, upload, and public-error behavior remains
supported. Newly rejected malformed, uncorrelated, cyclic-page, or
cross-incident responses are intentional hardening. There is no server, API,
database, WebSocket, persisted-data, external-consumer, or visual-golden
migration and no legacy compatibility layer.

LC-G001 through LC-G005 are closed. The only residual risk is the explicitly
deferred non-Workbook static cleanup above; it is not hidden by a suppression
or exception. Rollback remains slice-granular in reverse order. Each slice's
production code, tests, authored accounting inputs, generated outputs, and
tracker entry move together, and rollback must not recreate the legacy service
or a forwarding facade independently.

The final post-entry tracker gate passes: `make lint-markdown` at
`.cartulary/test-results/20260801T014046Z-p2131282` and `git diff --check` with
exit code 0. The verification below reruns both commands after recording this
evidence so the checked content and the handed-off content are identical.
