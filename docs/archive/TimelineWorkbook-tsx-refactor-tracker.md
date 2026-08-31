# TimelineWorkbook-tsx Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Original target path:**
  `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`
- **Current planning scope:** `apps/web/src/workbook/timeline/**`, plus only the
  workbook-runtime and common-component consumers named by Section 13.
- **Target label:** `TimelineWorkbook-tsx`
- **Normalized identifier:** `timeline-workbook-tsx`
- **Output path:** `docs/handoffs/TimelineWorkbook-tsx-refactor-tracker.md`
- **Status:** S-00 through S-08 are complete historical evidence. PR-00 through
  PR-04 are complete, and PR-05 through PR-06 are authorized sequentially by
  the production-readiness remediation task recorded in Section 13.
- **Allowed change in the current task:** The bounded production, test,
  ownership, verification, implementation-support documentation, generated
  topology, and tracker artifacts recorded in Section 13.
- **Non-goals:** Public-contract, backend, dependency, deployment, migration,
  generated-root, transport, persistence, authorization, and unrelated runtime
  behavior changes.
- **Current task authorization:** User task
  `user-timeline-workbook-production-readiness-remediation-2026-08-14`
  authorizes PR-00 through PR-06 sequentially with the mandatory tracker
  checkpoints in Section 15.
- **Completed implementation authorization:** User task
  `user-timeline-workbook-remediation-2026-08-14` authorized the historical
  S-00 through S-08 iteration.

Normative terms in this tracker have the following meanings:

- **MUST** and **MUST NOT** identify mandatory refactor requirements.
- **SHOULD** and **SHOULD NOT** identify requirements that may be waived only
  when the handoff records contrary owner evidence and the reason for the waiver.
- **MAY** identifies intentional implementation freedom.
- Unspecified internal decomposition is an implementation choice only when two
  different choices cannot affect callers, tests, ownership, or interoperability.

`docs/research/nlspec-spec.md` supplies writing doctrine for this tracker, and
`temp/analysis-notes.md` supplied review evidence for the completed iteration.
Neither document owns product behavior. Core 00 through Core 04 and adopted
subsystem owners remain the authorities for every preserved observable
behavior. The research document is not changed by this planning task.

The completed S iteration began from a 2,176-line target; Sections 2 through 12
retain that iteration's inventory, decisions, commands, and evidence as an
auditable historical record. At planning baseline
`98082ac04c2a4e8a03df3a0982e30a7de12680f5`, the target is 1,619 lines and the
next material risks are package-wide duplication, cross-surface UI ownership,
and render-order-dependent runtime callback bridges. Section 13 is the
controlling plan for that next iteration. The safe identifier
`timeline-workbook-tsx` remains a planning label, not a permanent module
boundary.

The source hierarchy used for this tracker was:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication; it is
   not applicable to this refactor plan.
4. Domain vocabulary, design direction, and implementation-support guides.
5. Current repository code and tests as current-state evidence.
6. The planning framework and prior handoffs as evidence and doctrine only.

### Completed S-iteration implementation authority contract

This historical contract records the authority used by the completed S
iteration. It does not authorize PR-00 through PR-06.

| Authority field | Required value | Current value |
| --- | --- | --- |
| Authorizing task or decision ID | A stable identifier that can be cited by every implementation handoff | `user-timeline-workbook-remediation-2026-08-14` |
| Authorized slices | An exact contiguous or enumerated subset of S-00 through S-08 | S-00 through S-08, sequentially |
| Starting source identity | Exact commit or immutable source snapshot | `5d070c2a9970049825d49c78c0fa9b6f84b02fad` |
| Allowed artifacts | The target; the Timeline-owned policy/controllers/hooks named by S-01 through S-06; the existing mention controller; focused tests; authored source-ownership, verification, and test-family inputs required by new paths/selectors; generated projections produced through public Make targets; this tracker and its implementation handoff | Adopted exactly as stated; generated projections remain conditional and generated roots MUST NOT be hand-edited |
| Prohibited changes | Public routes, envelopes, schemas, errors, identifiers, selectors, events, authorization outcomes, dependencies, package exports, deployment configuration, Core 00-05 behavior, grid-vendor ownership, transport ownership, persistence/projection ownership, or a generalized cross-workbook dispatcher | Binding default; a task that needs any listed change MUST reopen planning |
| Characterization prerequisite | The applicable S-00 seam evidence MUST pass against the pre-extraction implementation before that responsibility moves | Binding default |

An authorized implementation MUST preserve `TimelineWorkbookProps`,
`TimelineWorkbook`, all current callers, HTTP and WebSocket contracts,
`view_schema_id`, `sheet_ref`, `record_id`, `row_version`, field keys, selectors,
feature tuples, conflict and pending-replay semantics, focus behavior, existing
package boundaries, and server-authoritative authorization.

The implementation MUST NOT hand-edit generated files, import a grid vendor into
`apps/web`, move HTTP/WebSocket construction, persistence, projection SQL, or
authorization into an extracted controller, introduce a universal workbook
feature dispatcher, or combine observable behavior change with an extraction.

### Defaults and stop conditions

- Public behavior change is not authorized. Omission of a behavior-change rule
  means preserve the adopted owner and current conforming behavior.
- Unknown or mismatched schema and feature identities MUST resolve to the
  existing unsupported/omitted posture. They MUST NOT fall back to visible-label
  routing or generic mutation dispatch.
- Missing evidence is not evidence of absence. If an applicability decision
  requires evidence that is not recorded, the affected extraction MUST NOT begin.
- Generation impact defaults to `none`. Generation MUST run only when an authored
  generator input changes.
- An owner contradiction, a required public-contract change, or an inability to
  characterize current behavior MUST stop the affected slice. The handoff MUST
  record `BLOCKED: owner contradiction` for a contradiction and reopen planning
  instead of selecting a local convention.

Owner and supporting documents inspected:

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/testing-harness-nlspec.md`, limited to harness mechanics
- `docs/domain.md`
- `docs/design.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/research/nlspec-spec.md`, for specification-writing doctrine only
- `temp/analysis-notes.md`, as review evidence only

Repository implementation and support files inspected directly include:

- The target; `apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx`;
  `apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx`; and
  `apps/web/src/testing/timelineWorkbookRenderTestSupport.tsx`.
- `apps/web/src/workbook/timeline/models/timelineWorkbookSurfaceRuntime.ts`;
  the Timeline clipboard, evidence, history, mention, record-action, and
  row-mutation-editor adapters; the workbook view-query adapter; the workbook
  operation executor; and the workbook mutation command ports and implementation.
- `TimelineCollaborationBoundary.tsx`, `useTimelineCollaborationBindings.ts`,
  `useTimelineRowsLoader.ts`, `useTimelinePendingReplayController.ts`, and
  `useTimelineRowMutationCoordinator.ts` in the Timeline surface area.
- Frontend source-ownership and import-boundary manifests and their policy tests,
  the workbook layout policy test, generated-artifact policy, test-family maps,
  and the relevant protocol and OpenAPI operation-owner projections.
- The Timeline-facing WorkbookShell grid, autosave, sequencing, collaboration,
  history, inspector, payload, query, save-state, and sentinel tests; focused
  Timeline adapter/controller tests; and the Timeline workbook, keyboard,
  inspector-actions, sentinel, and public-route browser scenarios.

No owner contradiction was discovered. The planning framework describes useful
frontend seams, but the live repository already implements many of them through
semantic adapters, hooks, coordinators, generated-contract facades, source
ownership, and import guards. That framework/repository mismatch changes the
plan: a later refactor should extract residual controller logic from the target,
not recreate existing seams or invent a new backend or frontend module.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Private Timeline surface renderer and controller composition root. Wires querying, projection refresh, scalar and collection mutations, pending replay, conflict coordination, live collaboration and presence, history and lifecycle actions, mentions and entity creation, evidence attachment, related-record and indicator workflows, bulk tag, clipboard paste, fill-down, inspector state, focus continuity, keyboard handling, layout, and load/error presentation. | `TimelineWorkbookProps`; `TimelineWorkbook`. `TimelineWorkbookContent` and helper values are private. | Production: `apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx`. Test-only: `apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx` through `timelineWorkbookRenderTestSupport.tsx`. Production import policy allows only the facade. | Correct semantic dependencies on `@cartulary/grid-adapter`, `@cartulary/ui-contracts`, and `@cartulary/view-contracts`; React; workbook runtime, adapters, mutation ports, Timeline hooks, models, policies, and child components. No raw HTTP, SQL/storage, generated-source, or grid-vendor dependency. | WorkbookShell grid, autosave, action-sequencing, collaboration, history, inspector, payload, Timeline query, save-state, and sentinel suites; focused Timeline adapter/controller tests; Timeline workbook, keyboard, inspector-actions, sentinel, and public-route browser scenarios. | Consumes view- and UI-contract package surfaces. Transport adapters below it consume generated protocol bindings. The target neither imports nor edits generated files. | `web.workbook`, specifically the Capture and Timeline frontend composition surface. Cross-cutting behavior remains owned by existing workbook controllers and package adapters. | HIGH | Legitimate private facade at its caller boundary, but its body is an oversized mixed-responsibility controller. The filename does not justify a permanent `timeline-workbook-tsx` module. |

No file under the supplied target path is omitted: the path names one file, and
that file is fully in scope for planning inventory.

## 3. Module Boundary Diagnosis

The target is a legitimate private application/view facade at the registered
surface boundary, but it is not thin. It is simultaneously a frontend
shell/controller surface, view/projection orchestration layer, mutation
coordinator composition point, and mixed-responsibility file. It is
transport-adjacent only through semantic ports. It is not a persistence adapter,
backend domain module, or grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Private registered Timeline renderer and runtime composition | `TimelineWorkbook` and `TimelineWorkbookContent` | `web.workbook` Capture and Timeline surface | keep | The facade is the only permitted production importer and selects the registered `timeline` renderer. | Preserve the two exported symbols and caller shape. |
| Query, high-water, reconciliation, projection refresh, and continuity composition | Target plus existing row loader and mutation coordinator hooks | Existing Timeline query/reconciliation hooks in `apps/web` | split | `useTimelineRowsLoader` and `useTimelineRowMutationCoordinator` already own the core rules; the target wires callbacks and state. | Do not move projection semantics into a presentation package. |
| Scalar, collection, bulk, paste, and lifecycle mutation orchestration | Target plus semantic mutation ports and adapters | Existing workbook mutation ports and Timeline controllers | split | The target dispatches semantic operations; adapters below it own operation translation. | Preserve row versions, pending signatures, conflict flow, and refresh behavior. |
| Fill-down validation and dispatch | Inline in target | Timeline bulk controller under `timeline/bulk` | move | Inline logic validates one writable field and stable target IDs/versions before invoking the bulk port. | Keep grid-vendor details behind `@cartulary/grid-adapter`. |
| Keyboard event ownership and selector-based focus | Inline in target | Timeline keyboard/focus controller under `timeline/hooks` | move | Inline handlers own shortcut, Enter/Tab/Escape, Space, and DOM focus behavior through UI-contract selectors. | Preserve prevent-default and propagation behavior exactly. |
| Presence filtering, derivation, and publication | Inline in target plus collaboration bindings | Timeline collaboration controller | split | Existing bindings own session/live updates; the target derives cell/row indicators and publishes active-cell state. | Do not move WebSocket transport into the component or new controller. |
| Related-record versus indicator feature routing | Inline maps and inspector callbacks | Timeline feature policy and inspector feature controller | move | Target selects view schemas, contracts, generic workflow state, and indicator workflow state. | Keep the first extraction Timeline-specific; do not generalize without later evidence. |
| Mention undo auto-resolution request construction | Inline in target with existing mention controller callbacks | Existing Timeline mention controller | move | The target constructs a semantic undo payload and dispatches through a controller ref. | Preserve selector and row-version behavior. |
| Layout measurement, load/error presentation, and shell rendering | Target | Timeline surface composition | keep | `apps/web` owns shell/controller layout, and the target uses `WorkbookSurfaceLayout`. | Local `ResizeObserver` measurement is acceptable unless a later reuse case appears. |
| Direct grid behavior | `@cartulary/grid-adapter` package, consumed by target | `packages/grid-adapter` | keep | Repository search found direct `react-data-grid` imports only in the grid adapter. | No vendor extraction is required from the target. |
| Saved-view ownership | Shell-provided selector and query reload token; target consumes them | Workbook shell and view-contract owners | keep | The target does not persist saved views directly. | Preserve reload and active view-schema behavior. |

### Private interface contracts

The following interfaces are private to `apps/web`. Implementations MAY choose
internal variable and helper names, but they MUST satisfy these input, output,
default, and coupling contracts. No extraction adds a package export.

| Extraction | Required inputs | Required output | Explicit default and prohibited coupling |
| --- | --- | --- | --- |
| Feature policy | `viewSchemaId: string` and one `InspectorFeatureGroup` from `@cartulary/view-contracts` | A closed result with exactly one kind: `indicator` carrying the existing `IndicatorInspectorHandler`; `create_related` carrying the exact feature group; or `unsupported` carrying no handler | Unknown or mismatched full tuples return `unsupported`. Indicator-specific resolution precedes generic create-related resolution. The policy MUST NOT use React state, DOM state, visible labels, transport, authorization, generated transport, or grid-vendor APIs. |
| Fill controller | Existing `GridFillIntent`; committed Timeline row snapshots; Timeline `ViewContract`; grouping state; semantic bulk-fill port; save, continuity, refresh, error, pending-transaction, and focus callbacks | A single `onFillCells(intent: GridFillIntent): void` command | Invalid input rejects before dispatch with `Fill was rejected because one or more targets are unavailable or stale.` The controller MUST NOT use raw HTTP, SQL, label-derived fields, or vendor event types outside the grid-adapter contract. |
| Keyboard controller | Existing semantic keyboard mapping, Timeline interaction state, row/focus anchors, and existing save, navigation, inspector, history, mention, and message callbacks | Scalar-editor, collection-editor, and work-area keyboard handlers with the current React event parameter types | An unmapped or unavailable command is a no-op and remains unconsumed. The controller MUST NOT construct unrelated mutation requests or own route, editor-internal, or grid-vendor behavior. |
| Presence controller | Already active-sheet-scoped `readonly PresenceRecord[]`; current `WorkbookPresenceDraft`; semantic presence publication callback | Current-presence snapshot and ref; row-presence lookup; editing-cell lookup; edit-mode publication command | Initial draft is `{ fieldKey: null, mode: "viewing", recordId: null }`. Upstream collaboration remains responsible for full `sheet_ref` scoping. The controller MUST NOT create sockets, own reconnect/replay, receive credentials, or infer identity from labels or positions. |
| Mention controller extension | `AutoResolutionNotice`, current rows, and the dependencies already accepted by `useTimelineMentionActions` | `handleUndoAutoResolutionNotice(notice: AutoResolutionNotice): void` alongside existing mention commands | No matching committed row or resolved item is a no-op. Notice-to-mention/request construction MUST leave the component and MUST preserve the existing semantic mention port, raw text, identity, row-version, success, and rejection behavior. |
| Inspector controller | Exact `InspectorFeatureGroup`, Timeline schema identity, current Indicator resolver, generic workflow callbacks, and inspector-message callback | Feature-support query, feature-action command, and selected Indicator-handler state | Unsupported tuples do not dispatch. Indicator-specific handlers take precedence. The controller MUST NOT route by label, guess a route, perform authorization, or fall back from an Indicator action to generic patch. |
| `TimelineWorkbookContent` | Runtime hooks/controllers and legitimate local presentation state | The existing rendered Timeline surface under the existing collaboration boundary | It MUST contain no residual mutation policy, route translation, transport lifecycle, grid-vendor logic, or feature-routing matrix. |

Architectural finding: preserve `TimelineWorkbook` as a private registered facade,
then make its body composition-oriented by extracting the residual controllers
to existing semantic owner areas. Do not create a module named after the file.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Workbook view query envelope, Timeline schema, row creation, and patching | Core 01 and the generated HTTP-operation owners; workbook query and mutation adapters | `queryWorkbookView`, `createViewRow`, and `patchRecord` operation use below the component | WorkbookShell Timeline query, grid, payload, autosave, and public-route scenarios | Preserve complete `view_row` cells, blank-create suppression, validation errors, refresh identity, and current row versions | HIGH | No HTTP route, request, response, or generated binding change is planned. |
| Clipboard paste | Core 03; Timeline clipboard adapter and workbook operation port | `pasteWorkbookClipboard` adapter path | WorkbookShell sentinel and browser keyboard/paste scenarios | Preserve scalar/table parsing, row targeting, conflict reporting, and focus recovery | HIGH | Clipboard paste is workbook behavior, not the file-imports extension. |
| Bulk fill-down and tag mutation | Core 03; workbook bulk mutation port | `applyWorkbookBulkMutation` and inline fill controller logic | Sentinel, keyboard browser scenarios, and bulk-tag controller tests | Add focused fill validation/dispatch tests before extraction | HIGH | Stable record IDs and row versions are frozen. |
| History, rollback, delete, restore, review, and supersede | Core 02 and Core 03; Timeline history and action adapters | Semantic history and lifecycle operation adapters | WorkbookShell history and action-sequencing suites; Timeline browser scenario | Preserve selectors, pending-save sequencing, preview invalidation, and continuity anchors | HIGH | Revision and change-set behavior remains below the component. |
| Mention resolution and entity creation | Core 02 and Core 03; Timeline mention adapter/controller | `createViewRow` plus `resolveEntityMention` | Inspector and mention-related WorkbookShell coverage | Characterize undo auto-resolution payload and stale-version handling at the controller seam | HIGH | Entity identity and mention binding remain distinct. |
| Evidence upload, Evidence row creation, and Timeline attachment | Core 02 through Core 04; Timeline evidence adapter and workbook Evidence port | Object upload, Evidence creation, then Timeline create/patch | Inspector-related Evidence workflows and owner test families | Preserve operation ordering, authorization failures, and attachment refresh | HIGH | Storage and Evidence authorization do not move into the component. |
| Related-record and indicator workflows | Core 02 and Core 03; Timeline feature handlers and view contracts | Inline schema map, generic workflow dispatch, and indicator workflow state | WorkbookShell inspector create-related scenarios | Add a pure routing matrix covering every supported target and reset path | MEDIUM | Preserve existing schema IDs and do not infer new workflows from visible tab names. |
| Incident WebSocket lifecycle and `record_changed` events | Core 01, Core 03, and collaboration boundary/hooks | `TimelineCollaborationBoundary` and `useTimelineCollaborationBindings` | WorkbookShell collaboration and public-route recovery scenarios | Preserve sparse patch application, invalidation, reconnect, and access-loss behavior | HIGH | No WebSocket path, authorization, or event shape change is planned. |
| Keyed presence | Core 03; Timeline collaboration controller and bindings | Presence keyed by sheet, record, and field | WorkbookShell collaboration presence cases | Add focused derivation/publication tests if the extraction seam is not already isolated | MEDIUM | Keep session transport below the controller. |
| Pending queue, autosave, replay, conflicts, and row-version high-water | Core 01 and Core 03; pending replay and mutation coordinator hooks | Existing hooks own admission, queue, replay, and conflict coordination | Autosave, collaboration, action-sequencing, history, and sentinel suites | Baseline existing duplicate-suppression, queue-capacity, replay, and stale-result cases | HIGH | Preserve exact user-visible save/conflict outcomes. |
| Saved-view reload, active view schema, and inspector lifecycle | Core 01 and Core 03; workbook shell and Timeline inspector controllers | Shell passes selector/query state; target retains Timeline schema and default-closed inspector | WorkbookShell inspector, save-state, and Timeline query suites | Preserve reload tokens, sheet switches, no-row state, and relationship-cell selection | HIGH | The target does not own saved-view persistence. |
| Grid adapter, stable row identity, keyboard/focus, selectors, and accessibility surface | Core 03, design direction, grid adapter, and UI contracts | Semantic grid callbacks and `@cartulary/ui-contracts` selectors | Grid, sentinel, keyboard, inspector-actions, layout policy, and accessibility-related browser rows | Add focused keyboard ownership cases around extracted handlers | HIGH | No direct vendor import or selector rename is allowed. |
| Generated protocol, view, and UI contract surfaces | Adopted owners projected through contract generators | Generated roots and artifact policy; package imports in and below target | Generated drift and package tests | None unless a later task intentionally changes an owner contract | LOW for this plan | Frozen consumers; never hand-edit generated files. |
| Harness/test evidence accounting | Adopted Testing Harness NLSpec; verification owner catalog and test families | `module.timeline`, `module.workbook`, and `web.workbook` owner explanations | Existing Vitest and Playwright rows | Account for any new source/test path through authored owner inputs | MEDIUM | Evidence routing does not define runtime architecture or requirements. |
| Authorization and incident access loss | Core 04 and server-side operation owners | UI editable/role gates plus operation and collaboration access-loss callbacks | Collaboration, Evidence, and public-route recovery scenarios | Preserve server-authoritative denials and client recovery behavior | HIGH | UI role gating is not the authorization authority. |

### S-00 characterization contract

Every test added by S-00 MUST pass against the pre-extraction implementation.
The test MUST characterize the applicable adopted owner and current conforming
behavior; it MUST NOT encode a desired cleanup behavior. If the current behavior
conflicts with an adopted owner, the test MUST NOT bless the conflict and the
slice MUST stop under the owner-contradiction rule in Section 1.

| Seam | Minimum isolated current-behavior cases required before extraction | Primary evidence owner |
| --- | --- | --- |
| Fill validation and dispatch | Valid single-field fill; pointer and `Ctrl/Cmd+D` target equivalence; leading committed-cell source value; explicit target `record_id` and current `base_row_version`; stable target order; grouped, draft, hidden, collection, non-grid-editable, stale, read-only, and presentation targets reject before dispatch; no partial mutation; failure or conflict preserves focus and draft posture; vendor double-click does not become an implicit fill. | `web.workbook`; add `module.workbook` where the semantic bulk-port result is the verified postcondition. |
| Keyboard and focus ownership | Grid versus editor, menu, popover, dialog, and inspector ownership; Enter, Shift+Enter, Tab, Shift+Tab, Escape, Space, and every registered shortcut; exact `preventDefault` and propagation behavior; unavailable actions remain unconsumed; edit commit/cancel order; inspector-open and inspector-close focus destinations; selector-based focus recovery after accepted and rejected mutations. | `web.workbook`; browser evidence MUST cover actual focus and event behavior when owner-qualified Vitest rows are insufficient. |
| Presence derivation and publication | Upstream full-`sheet_ref` scoping, including direct view-schema versus saved-view identity; row indicator by `record_id`; same-cell indicator by `record_id + field_key + editing mode`; active-cell publication; duplicate or reordered records do not change identity; ambient presence does not change save, conflict, or replay state; stale-session and authorization-loss cleanup; no label, row-number, or column-header inference. | `web.workbook`; `module.timeline` applies only when service-backed transport, session, replay, or publication behavior changes. |
| Mention undo auto-resolution | Exact current notice-to-mention conversion; source record, field, item, mention row version, raw text, target entity, and action semantics; absent row or resolved item is a no-op; stale version enters the existing refresh/conflict path; success and rejection handling remain unchanged; entity mention, host, identity, and canonical entity meanings remain distinct. | `web.workbook`. The exact internal object is frozen only for this extraction and does not become a public contract. |
| Inspector feature routing | Complete matrix for every Timeline-supported related-record target and Indicator-specific action; routing by the full stable feature tuple rather than label; Indicator-specific tuple resolution before generic families; unknown or mismatched tuple has no handler; row, row-version, surface, incident, authorization, and lifecycle changes clear stale workflow state; cancel, failure, and success retain the existing selection and reset behavior. | `web.workbook`; add `module.workbook` where a cross-module workflow postcondition is verified. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| The component concentrates residual controller branches and a large callback graph. | Direct inspection of all 2,176 lines and its 58 import sources. | Small edits can affect multiple observable workflows. | `must_fix` | `web.workbook` Timeline controllers | Extract one characterized semantic responsibility per slice while keeping the facade stable. |
| Fill-down policy, target validation, mutation dispatch, refresh, and focus recovery are inline. | Inline `GridFillIntent` handling builds stable ID/version targets and invokes the bulk port. | A move can change mutation or focus semantics. | `must_fix` | Timeline bulk controller | Characterize, then extract without changing messages, signatures, or sequencing. |
| Keyboard policy and direct selector-based focus actions are interleaved with composition. | Inline work-area and cell handlers use UI-contract selectors. | Event consumption or focus regression is user-visible. | `should_fix` | Timeline keyboard/focus controller | Freeze the event-ownership matrix before extraction. |
| Feature routing and schema constants are embedded in component assembly. | Related-record schema map, observation source fields, and indicator/generic dispatch are inline. | Adding a feature can expand the catch-all. | `should_fix` | Timeline feature policy and inspector controller | Extract a pure policy and keep it Timeline-specific initially. |
| Presence derivation is mixed with shell/controller assembly. | Target filters presence and produces row/cell indicators while hooks own transport updates. | Incorrect extraction could lose or mis-key presence. | `should_fix` | Timeline collaboration controller | Move only derivation/publication; retain transport and reconciliation owners. |
| Mention undo request construction remains outside the existing mention controller. | Target constructs and dispatches the auto-resolution payload through a ref. | Version or selector drift could break undo. | `should_fix` | Existing Timeline mention controller | Extend the semantic controller and add focused coverage. |
| Grid-vendor integration is already isolated. | No target or Timeline-area `react-data-grid` import; direct imports are in `packages/grid-adapter`. | Moving vendor details outward would violate the boundary. | `intentional/no_action` | `packages/grid-adapter` | Preserve current imports and guardrail. |
| Transport and generated protocol use are already below the component. | The target invokes semantic adapters/ports and contains no raw `fetch` or generated-source import. | Recreating a transport facade would duplicate existing ownership. | `intentional/no_action` | Workbook adapters and platform transport | Preserve semantic ports; do not add route logic to the target. |
| The view- and UI-contract imports are correct frontend dependencies. | Target consumes `@cartulary/view-contracts` and `@cartulary/ui-contracts`. | Bypassing them would risk schema and selector drift. | `intentional/no_action` | Contract packages | Keep generated implementation details behind package facades. |
| Local viewport measurement and presentation-state selection remain in the surface. | Target uses `ResizeObserver` and `WorkbookSurfaceLayout`; layout policy test accepts the surface structure. | Premature abstraction could add indirection without an owner. | `intentional/no_action` | Timeline surface composition | Retain unless a separate reuse requirement is evidenced. |
| Test fixtures import the private component directly. | Imports are under `apps/web/src/testing`; production import policy excludes test paths. | Treating test callers as production architecture would distort the boundary. | `intentional/no_action` | `web.testing` | Preserve test-only access and the production facade rule. |
| A universal cross-workbook feature dispatcher is not evidenced. | Current handlers are owner-specific and the target has only Timeline-specific dispatch evidence. | Premature generalization could erase bounded-context ownership. | `defer` | Workbook owner policy, if later justified | Keep the first extraction under Timeline and revisit only with owner evidence. |
| Generated artifacts must not be edited for this extraction. | Generated artifact policy declares protocol, view-contract, UI-contract, and topology outputs. | Hand edits would drift from adopted owners. | `intentional/no_action` | Contract and harness generators | Update authored inputs and run Make-owned generation only if a later slice actually requires it. |

No direct SQL/storage coupling, domain logic under platform packages,
authorization ownership inversion, duplicated saved-view persistence, or
test-only assumption leaking into production code was found in the target.

### Mandatory security, compatibility, and coupling invariants

- Client role gates MUST remain presentation controls. Every operation MUST
  continue through its server-authorized semantic port, and authorization loss
  MUST retain the existing protected-state cleanup and recovery behavior.
- Extracted controllers MUST NOT receive credentials, session tokens, object
  locators, raw Evidence bytes, transport secrets, or new egress capabilities.
- Identity MUST remain based on `view_schema_id`, complete `sheet_ref`,
  `record_id`, `row_version`, `field_key`, and the complete inspector feature
  tuple. Labels, row/column positions, React names, and DOM structure MUST NOT
  become persisted or dispatched identity.
- No extracted controller may import `react-data-grid`, create a WebSocket,
  construct raw HTTP, own projection or persistence logic, or import
  `TimelineWorkbook.tsx`. Dependency direction remains component to semantic
  controller/policy.
- The first feature-routing extraction MUST remain Timeline-specific. A shared
  dispatcher requires a separate owner contract and evidence from another
  independent owner.

### Deterministic source and test ownership

The live catalog assigns the existing Timeline surface and mention controller to
`web.workbook`. The following defaults are binding unless the implementation
preflight finds a changed, non-contradictory live owner fact.

| Planned path or responsibility | Required default owner | Ownership boundary |
| --- | --- | --- |
| `apps/web/src/workbook/timeline/models/timelineWorkbookFeaturePolicy.ts` | `web.workbook` | Timeline-specific pure feature/schema policy |
| `apps/web/src/workbook/timeline/bulk/useTimelineFillController.ts` | `web.workbook` | Client fill validation, semantic bulk dispatch, continuity, and focus recovery |
| `apps/web/src/workbook/timeline/hooks/useTimelineKeyboardController.ts` | `web.workbook` | Application keyboard ownership, focus, and semantic callback coordination |
| `apps/web/src/workbook/timeline/collaboration/useTimelinePresenceController.ts` | `web.workbook` | Active-sheet presence derivation/publication coordination, not transport |
| `apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts` | Preserve live `web.workbook` owner | Existing Timeline semantic mention controller extension |
| `apps/web/src/workbook/timeline/hooks/useTimelineInspectorFeatureController.ts` | `web.workbook` | Timeline-specific feature selection and handler dispatch |
| `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx` | Preserve live `web.workbook` owner | Existing private Timeline composition surface |
| Grid-vendor code | No change; retain `packages/grid-adapter` ownership | Vendor lifecycle and event translation remain behind the adapter |
| WebSocket transport and replay | No change; retain current collaboration ownership | Presence extraction does not absorb session or replay behavior |
| HTTP and generated-protocol adapters | No change; retain current adapter/contract owners | Controllers consume semantic ports only |

For every added test title or browser scenario, the implementer MUST apply this
algorithm in order:

1. Identify the independently reportable postcondition. Filename, new hook,
   delivery slice, runner, and maintainer do not determine ownership.
2. Reuse an existing semantic row when the new selector proves the same
   postcondition. Add a row only for a distinct postcondition that must be
   independently selected and reported.
3. Assign exactly one primary owner: frontend interaction/focus/routing belongs
   to `web.workbook`; cross-module mutation results belong to `module.workbook`
   only when that module owns the postcondition; service-backed Timeline
   transport/replay/publication belongs to `module.timeline` only when exercised.
4. Retain collaborators without transferring primary ownership.
5. Bind the exact full Vitest title or exact browser scenario ID. Prefixes,
   globs, and zero-row selectors are invalid.
6. Run owner explanation and prove that the selector is active, non-duplicate,
   non-cross-owner, and selects at least one row.

Rows MUST use durable semantic identities. They MUST NOT be named after S-00,
another delivery slice, a new filename, or "refactor regression."

Ownership accounting is part of each slice, not deferred cleanup:

```text
new source or test
  + authored source ownership
  + authored test-family or verification input when applicable
  + Make-owned generated refresh when applicable
  + focused validation
  = one independently reversible checkpoint
```

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Establish clean state, authority order, target existence, safe identifier, and planning-only limits. | Tracker only for this task | Read-only Git inspection and framework/owner review | Source posture and permitted-write boundary recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory the complete target, callers, dependencies, tests, and owner guardrails. | Target, facade, test fixture/support, source-ownership files | Repository search and direct file reads | Section 2 has no generic or omitted target row. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-05 | Map every affected observable behavior to its normative and implementation owner. | Core owners, adapters, hooks, contracts, operation-owner projections | Contract/test evidence review | Freeze map assigns an owner and test posture to each risk. |
| WF-03 | Characterization test-gap analysis | parallel | WF-01 | WF-05, WF-07 | Identify existing coverage and the smallest missing seam tests. | WorkbookShell, focused Timeline, and browser tests | `make explain-test-owner` and test source review | Fill, keyboard, presence, and feature-routing gaps are explicit. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Separate legitimate composition from residual owner logic and confirm transport/vendor boundaries. | Target, runtime, adapters, hooks, package/import policies | Direct imports and policy-test review | Findings are classified without inventing a new module. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Keep the private facade stable and assign residual logic to existing Timeline owners. | Timeline models, bulk, hooks, collaboration, mentions, inspector | Plan review against frozen contracts | Proposed owners and non-actions are decision complete. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Sequence independently reversible, behavior-preserving extractions. | Target and planned controller/policy files | Per-slice focused commands | Every slice has dependency, rollback, and binary exit criteria. |
| WF-07 | Same-slice test and ownership accounting | chain | WF-03, WF-06 | WF-08 | Require each new path or exact selector to land with its authored ownership, verification accounting, applicable generated refresh, and focused evidence. | Each implementation slice plus authored source-ownership and test-family inputs | Import-boundary checks, exact owner explanations, and conditional generation checks | Every slice is owned and selectable at its checkpoint; no accounting debt remains for S-08. |
| WF-08 | Final reconciliation, validation, and handoff | chain | WF-06 and WF-07 | None | Audit final source ownership, exact selectors, generation impact, and focused-to-broad evidence. | Implemented slice files, authored accounting, generated projections when applicable, and tracker | Section 8 command and evidence contract | Results, run roots, generation disposition, failures, and remaining risks are recorded. |

Planning for WF-00 through WF-07 is complete in this tracker. WF-08 and all
production implementation remain future work.

## 7. Proposed Refactor Slice Plan

All slices below preserve observable behavior. Any change to a route, envelope,
event, authorization result, schema, selector, saved-view behavior, mutation
semantics, or generated contract is out of scope and **requires later
authorization**.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | Recorded RB-001 authority | Establish passing pre-extraction evidence for fill, keyboard/focus, presence, mention undo, and inspector routing exactly as specified in Section 4. Determine and land exact selector accounting with every new test. | Existing Timeline and WorkbookShell tests; focused seam tests; authored test-family/verification inputs when selectors change | A characterization test could bless non-conforming or desired behavior, omit a seam, or select no row. | All five seam families in Section 4 MUST have isolated owner-qualified evidence before their production responsibility moves. | Required baseline commands in Section 8 plus conditionally required browser/service-backed commands | Revert only characterization/accounting changes that encode unsupported behavior; do not alter production to make a new expectation pass. | Every new test passes against the pre-extraction implementation; exact rows/selectors, source/catalog identity, results, and run roots are recorded; RB-002 and the preflight portion of RB-003 are resolved. |
| S-01 | S-00 feature-routing evidence | Extract the closed feature-policy result into `apps/web/src/workbook/timeline/models/timelineWorkbookFeaturePolicy.ts`; add its `web.workbook` ownership and exact test accounting in the same checkpoint. | Target, policy, policy test, authored ownership and test inputs | Full tuple, supported-target, precedence, and unsupported-result drift | Exhaustive routing matrix; existing inspector and related-record cases | Per-slice baseline commands; `module.workbook` only for a cross-module postcondition | Revert policy, target wiring, source ownership, selector accounting, and any generated projection as one checkpoint. | Target consumes the private pure policy; every characterized tuple maps identically; the new path and tests each have one owner/row. |
| S-02 | S-00 fill evidence | Move fill validation, stable target construction, semantic bulk dispatch, refresh, save state, continuity, and focus recovery into `apps/web/src/workbook/timeline/bulk/useTimelineFillController.ts`; land ownership/accounting with the file. | Target, fill controller/tests, authored ownership and test inputs | Record identity/version, grouping, pending signature, target order, error text, partial dispatch, conflict/save state, and focus | Complete Section 4 fill matrix; preserve sentinel and keyboard evidence | Per-slice baseline commands plus `make test-slice OWNER=module.workbook` where applicable | Revert controller, wiring, ownership, and selector changes together. | No inline fill policy remains; the existing rejection message and all accepted/rejected sequencing are unchanged; exact ownership/evidence is clean. |
| S-03 | S-00 keyboard evidence | Move scalar-editor, collection-editor, work-area, shortcut, and selector-based focus coordination into `apps/web/src/workbook/timeline/hooks/useTimelineKeyboardController.ts`; land ownership/accounting with the file. | Target, keyboard controller/tests, authored ownership and test inputs | Event consumption, edit commit/cancel, navigation, inspector state, focus destinations, and unavailable actions | Complete Section 4 keyboard matrix; preserve grid, sentinel, keyboard, inspector-actions, and accessibility evidence | Per-slice baseline commands plus browser E2E under Section 8 applicability | Revert controller, wiring, ownership, and selector changes together. | Target wires semantic handlers only; every characterized key path preserves event and focus results; exact ownership/evidence is clean. |
| S-04 | S-00 presence evidence | Move active-sheet presence derivation and publication coordination into `apps/web/src/workbook/timeline/collaboration/useTimelinePresenceController.ts`; land ownership/accounting with the file and leave transport bindings unchanged. | Target, presence controller/tests, existing collaboration bindings, authored ownership and test inputs | Sheet/row/cell identity, duplicate ordering, stale session state, authorization loss, and accidental transport/replay movement | Complete Section 4 presence matrix; preserve collaboration and public-route recovery evidence | Per-slice baseline commands, browser E2E, and conditional service-backed Timeline slice under Section 8 | Revert controller, wiring, ownership, and selector changes; do not change transport hooks as rollback. | Visible indicators and publications match baseline; transport/session/replay ownership is unchanged; exact ownership/evidence is clean. |
| S-05 | S-00 mention-undo evidence | Extend `apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts` with `handleUndoAutoResolutionNotice`; remove notice-to-mention/request construction from the component; land exact test accounting in the same checkpoint. | Target, existing mention controller/tests, authored test inputs when selectors change | Source record/field/item identity, mention version, raw text, target entity, no-op cases, and conflict/result handling | Complete Section 4 mention-undo matrix; preserve inspector mention evidence | Per-slice baseline commands | Revert controller extension, component delegation, and selector accounting together. | Component passes only `AutoResolutionNotice`; the controller owns conversion/dispatch; all characterized results and live `web.workbook` ownership remain unchanged. |
| S-06 | S-01 and S-00 inspector evidence | Move support checks, Indicator-first dispatch, generic workflow dispatch, selected Indicator state, and reset behavior into `apps/web/src/workbook/timeline/hooks/useTimelineInspectorFeatureController.ts`; land ownership/accounting with the file. | Target, inspector controller/tests, feature policy, existing handlers, authored ownership and test inputs | Full-tuple routing, prerequisites, unsupported omission, stale workflow state, selection/reset behavior, and schema mismatch | Complete Section 4 inspector matrix; preserve create-related and Indicator evidence | Per-slice baseline commands plus `module.workbook` where it owns the postcondition | Revert controller, wiring, ownership, and selector changes; retain S-01 only if it remains independently complete. | All supported actions route identically, unsupported actions omit, inactive state resets identically, and exact ownership/evidence is clean. |
| S-07 | S-02, S-03, S-04, S-05, S-06 | Reduce `TimelineWorkbookContent` to runtime/controller composition, collaboration-boundary wiring, legitimate local presentation state, and rendering. | Target and already owned controllers | Public prop/caller drift, callback ordering, layout, or residual policy | Preserve complete focused/browser evidence and production import policy | Per-slice baseline commands and applicable browser evidence | Revert composition-only cleanup without reverting independently passing controllers. | Public exports/callers and observable behavior are unchanged; Section 3 prohibited residual responsibilities are absent. |
| S-08 | S-07 and every same-slice accounting checkpoint | Reconcile final source ownership, exact selectors, verification rows, generation impact, retained evidence, and handoff. It MUST NOT be used to repair ownership debt that should have landed with S-01 through S-06. | Authored ownership/test inputs, generated projections only when applicable, tracker, retained results | Duplicate/missing owner, unknown/cross-owner/zero-row selector, generated drift, missing run roots, or unrecorded deviation | Ownership/harness policy evidence and the full Section 8 final gate | Owner explanations, conditional generation checks, `make agent-finalize`, and `make check` | Revert only final mechanical reconciliation; reopen the originating slice for semantic accounting defects. | Every final path and selector has exactly one correct owner/row; generation is clean or explicitly `none`; final evidence and handoff are complete. |

### Implementation checkpoint ledger

| Slice | Status | Dependency state | Exact files changed at checkpoint | Rollback posture | Next action |
| --- | --- | --- | --- | --- | --- |
| S-00 | DONE | RB-001, RB-002, and the preflight portion of RB-003 are resolved | `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert the authored mention-undo row and regenerate topology; no production source is involved | Begin S-01 only after this checkpoint passes Markdown validation. |
| S-01 | DONE | S-00 satisfied | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; new `apps/web/src/workbook/timeline/models/timelineWorkbookFeaturePolicy.ts` and colocated test; `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert policy, tests, ownership/accounting, generated topology, and target wiring together; S-00 remains independently valid | Begin S-02 after Markdown validation. |
| S-02 | DONE | S-01 satisfied | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; new `apps/web/src/workbook/timeline/bulk/useTimelineFillController.ts` and colocated test; `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert controller, tests, ownership/accounting, generated topology, and target wiring together; S-00 and S-01 remain independently valid | Begin S-03 after Markdown validation. |
| S-03 | DONE | S-02 satisfied | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; new `apps/web/src/workbook/timeline/hooks/useTimelineKeyboardController.ts` and `apps/web/src/workbook/timeline/useTimelineKeyboardController.test.tsx`; `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert controller, test, ownership/accounting, generated topology, and target wiring together; S-00 through S-02 remain independently valid | Begin S-04 after Markdown validation. |
| S-04 | DONE | S-03 satisfied | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`; new `apps/web/src/workbook/timeline/collaboration/useTimelinePresenceController.ts` and colocated test; `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert controller, semantic selection callback, test, ownership/accounting, generated topology, and target wiring together; do not touch collaboration transport | Begin S-05 after Markdown validation. |
| S-05 | DONE | S-04 satisfied | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts`; new `apps/web/src/workbook/timeline/useTimelineMentionActions.test.tsx`; `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert mention-controller extension, test/accounting, generated topology, and target delegation together; S-00 through S-04 remain independently valid | Begin S-06 after Markdown validation. |
| S-06 | DONE | S-05 and S-01 satisfied | `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`; `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`; new `apps/web/src/workbook/timeline/hooks/useTimelineInspectorFeatureController.ts` and `apps/web/src/workbook/timeline/useTimelineInspectorFeatureController.test.tsx`; `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker | Revert inspector controller, test/accounting, generated topology, lifecycle delegation, and target wiring together while retaining independently complete S-01 | Begin S-07 after Markdown validation. |
| S-07 | DONE | S-02 through S-06 satisfied | This tracker; no additional production edit was required after the originating slices removed their inline owners | The source audit is documentation-only; independently complete S-01 through S-06 remain intact | Begin S-08 after Markdown validation. |
| S-08 | DONE | S-07 satisfied | `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json`; this tracker; final reconciliation covers every source/test file named by S-01 through S-06 | Revert only final accounting/handoff edits and reopen an originating slice for semantic debt | No implementation work remains; use the final handoff and retained roots below. |

### S-00 implementation evidence

- **Authority and source:** task
  `user-timeline-workbook-remediation-2026-08-14`; source revision
  `5d070c2a9970049825d49c78c0fa9b6f84b02fad`; authorized range S-00
  through S-08; `domain vocabulary unchanged`; no Core or domain owner edit.
- **Catalog identity:** `contracts/verification/owners/web.workbook.json`
  SHA-256 `800d5d9cde70f8c737dd4292b1516bfe74e294915a2b476203f828abea85450f`;
  `tools/test_families/web.workbook.json` SHA-256
  `7c6934783e6635a72892730da9d23d274efd6952f717600fc8385a2957f4193f`;
  `tools/test_families/module.workbook.json` SHA-256
  `5e48c4eaa0566dce2a0656daee88ff7291e7b80f2186857bc1b08376609183df`;
  `tools/test_families/module.timeline.json` SHA-256
  `bc7c24ce2e6a07732130a3816c242701a27a3e6f8c56d7463821c9233d18a74b`.
- **Fill rows and selectors:**
  `module.timeline.browser_support.fill_down_helper_contract` selects
  `apps/web/e2e/timeline.support.spec.ts` title `exact-range fill-down emits
  the anchored bulk-mutation contract`;
  `module.timeline.browser_support.verify_full_keyboard_clipboard_contract_one_clic_ec36b90e7b`
  selects `apps/web/e2e/keyboard.spec.ts` title `Verify full
  keyboard/clipboard contract: one-click edit, copy, paste, exact-range
  fill-down, frozen columns, virtual scroll, group rows, focus restoration,
  and Esc priority ladder.` Existing grid-adapter tests retain rejection and
  double-click ownership evidence.
- **Keyboard/focus rows and selectors:**
  `module.workbook.frontend.keyboard_command_mapping_covers_arrow_keys_enter_936238eee1`
  selects the three exact titles in
  `apps/web/src/workbook/utils/workbookKeyboard.test.ts`; the five
  `web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_*`
  rows select the exact navigation, draft, Enter/Shift+Enter, unavailable
  shortcut, and focus-continuity titles in
  `apps/web/src/workbook/WorkbookShell.sentinel.test.tsx`; the full keyboard
  browser row above verifies live event and focus behavior.
- **Presence row and selector:**
  `web.workbook.regression.workbookshell_presence_indicators_render_from_ke_10dbb35d28`
  selects `apps/web/src/workbook/WorkbookShell.collaboration.test.tsx` title
  `workbook collaboration coverage presence indicators render from keyed
  socket state without changing save-state`. The case distinguishes direct
  view-schema and saved-view identities, self and other connections, exact
  row/cell/editing identity, removal, and ambient save-state isolation.
- **Mention-undo row and selector:** newly authored
  `web.workbook.regression.mention_undo_current_committed_identity_ddf69011e7`
  selects `apps/web/src/workbook/WorkbookShell.support.test.tsx` title
  `support TimelineWorkbookRuntimeFixture sends auto-resolution Undo with the
  current post-resolution row version`. It passed against the pre-extraction
  implementation and preserves current identity, version, focus, and scroll.
- **Inspector rows and selectors:**
  `module.workbook.frontend_unit.verify_active_view_schema_id_selects_inspector_c_fed994e037`
  selects the active-sheet reset and no-row titles, while
  `module.workbook.frontend_unit.verify_inspector_panels_and_feature_groups_rende_b947f0007c`
  selects the exact related Task Request and Evidence titles in
  `apps/web/src/workbook/WorkbookShell.inspector.test.tsx`. The selected-row
  fixture also asserts the Indicator route owner and omits an unsupported
  Timeline action. S-01 adds the isolated exhaustive tuple matrix before any
  target routing is replaced.
- **Terminal results:** `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T120806Z-p2530172`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T120806Z-p2530185`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T120814Z-p2531432`;
  the post-accounting `make test-slice OWNER=web.workbook` passed 121/121 at
  `.cartulary/test-results/20260814T121703Z-p2624650`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T120814Z-p2531439`; and
  `make browser-e2e-webserver-backed` passed 62/62 at
  `.cartulary/test-results/20260814T121036Z-p2568453`.
- **Accounting and generation:** all three owner explanations passed;
  `web.workbook` now reports 120 rows before execution expansion and the
  selected slice reports 121 work units. `make test-catalog-check` passed.
  The authored row made execution topology stale, so the initial
  `make json-shape-check` failed at
  `.cartulary/test-results/20260814T121703Z-p2624600`; this was repaired only
  through `make generate`, which passed at
  `.cartulary/test-results/20260814T121758Z-p2638646` and changed only the
  generated topology render index. `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T121839Z-p2641873`,
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T121839Z-p2641894`, and the terminal
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T121839Z-p2641882`. Product-contract and
  generated-root impact remains `none`; generated harness-topology impact is
  the one Make-produced render index.
- **Deviations and failures:** no owner contradiction, product failure,
  infrastructure failure, cancellation, cross-owner selector, duplicate row,
  or zero-row selection remains. `module.timeline` service-backed validation
  is inapplicable because S-00 changes neither transport nor publication.

### S-01 implementation evidence

- **Remediation:** `resolveTimelineWorkbookFeature` now resolves only the
  canonical Timeline schema and complete semantic feature binding. It returns
  the closed `indicator`, `create_related`, or `unsupported` result, evaluates
  Indicator handlers before generic creation, and requires a live target view
  contract. Target registration is derived from the immutable Timeline
  contract, so compatible future contract additions do not require another
  component-local schema list. Labels are deliberately excluded from identity.
- **Component disposition:** the eight-schema constant, target-contract map,
  permissive inline create-route predicate, and direct Indicator resolver call
  were removed. `TimelineWorkbookContent` consumes the policy and starts only
  the resolved canonical workflow. Public exports, callers, routes, selectors,
  payloads, and authorization outcomes are unchanged.
- **Source owner and selector:** both new model paths are explicitly owned by
  `web.workbook` in `tools/frontend_source_ownership.json`. Row
  `web.workbook.regression.timeline_feature_policy_closed_semantic_routing_bbd0ade7b2`
  selects the exact titles `timelineWorkbookFeaturePolicy resolves every
  canonical Timeline create-related and Indicator tuple` and
  `timelineWorkbookFeaturePolicy fails closed for unsupported schemas and
  altered semantic tuple members` in the colocated test. It covers all eight
  generic targets, the Timeline Indicator handler and precedence, every
  supported target registration, altered role/mutation/confirmation/panel/
  route/seed/disabled members, unknown features and targets, a foreign schema,
  and label independence.
- **Terminal results:** `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T122326Z-p2648287`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T122326Z-p2648301`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T122400Z-p2655854`;
  `make test-slice OWNER=web.workbook` passed 122/122 at
  `.cartulary/test-results/20260814T122400Z-p2655876`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T122400Z-p2655856`; and
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T122400Z-p2655822`.
  `make test-catalog-check` and both applicable owner explanations passed;
  `web.workbook` reports 121 active rows and no service-backed rows.
- **Generation and rollback:** `make generate` passed at
  `.cartulary/test-results/20260814T122346Z-p2652762`; only the generated
  harness topology render index changed. `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T122703Z-p2712786`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T122703Z-p2712807`. Product-contract and
  generated-root impact is `none`. The slice is independently reversible by
  restoring the old target wiring and removing its policy, test, owner row,
  ownership paths, and regenerated topology entry.
- **Deviations, failures, and next action:** none. No compatibility wrapper,
  dependency, feature flag, public export, backend path, or generalized
  dispatcher was added. Begin S-02 only after this tracker passes Markdown
  validation.

### S-02 implementation evidence

- **Remediation:** `planTimelineFill` is a pure, closed planner over the
  semantic `GridFillIntent`, current committed Timeline rows, immutable view
  contract, interaction mode, grouping state, and visible field identities. It
  emits one complete ordered `TimelineFillCommand` or the preserved rejection
  message. It rejects read-only, grouped, hidden, wrong-surface, extension,
  draft/missing, collection, non-grid-editable, stale, pending, mismatched-field,
  duplicate, and empty target shapes atomically. It preserves leading source
  text, exact `record_id`/`base_row_version` values, and target order.
- **Dispatch ownership:** `useTimelineFillController` is the only Timeline fill
  effect owner. It dispatches solely through `TimelineBulkMutationPort.fillDown`
  and preserves begin-continuity, begin-save, queued dispatch, client-transaction
  tracking/resolution, rejected-continuity clearing, conflict state, accepted
  refresh, source-focus restoration, and saved-state ordering. The component
  now supplies semantic ports and `onFillCells`; no inline fill policy remains.
- **Source owner and selector:** both new bulk paths are explicitly owned by
  `web.workbook`. Row
  `web.workbook.regression.timeline_fill_planning_and_semantic_dispatch_c446e90622`
  selects the exact three titles in the colocated test: `useTimelineFillController
  plans one ordered versioned command from a committed visible scalar source`,
  `useTimelineFillController rejects every malformed, hidden, grouped, stale,
  pending, or non-editable fill atomically`, and `useTimelineFillController
  dispatches only the semantic bulk port and preserves save, refresh, focus,
  and conflict sequencing`. `module.workbook` remains a collaborator for the
  semantic bulk-port postcondition; the browser rows recorded in S-00 retain
  pointer, keyboard, focus, and double-click evidence.
- **Compatibility and risk posture:** public props, exports, routes, payloads,
  selectors, and supported user flows are unchanged. Newly explicit rejection
  covers malformed internal duplicate/source-surface/read-only/hidden intents;
  no valid adapter-produced flow depends on them. Leaving those gaps open would
  permit duplicate or stale writes and ambiguous surface ownership, so the
  stricter fail-closed behavior is intentional and private.
- **Terminal results:** after correcting a test-fixture-only failure kind and
  callback return type found by the first typecheck, the terminal
  `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T123310Z-p2727578`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T123239Z-p2722968`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T123336Z-p2731126`;
  `make test-slice OWNER=web.workbook` passed 123/123 at
  `.cartulary/test-results/20260814T123336Z-p2731121`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T123336Z-p2731119`;
  `make browser-e2e-webserver-backed` passed 62/62 at
  `.cartulary/test-results/20260814T123620Z-p2787841`; and
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T123336Z-p2731065`.
- **Accounting, generation, and rollback:** `make test-catalog-check` and the
  `web.workbook` owner explanation passed with 122 active rows and no
  service-backed rows. `make generate` passed at
  `.cartulary/test-results/20260814T123322Z-p2727989`, changing only generated
  harness topology; `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T124029Z-p2841452`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T124029Z-p2841445`. Product-contract and
  generated-root impact is `none`. Rollback removes the controller/test and
  same-slice accounting, regenerates topology, and restores the inline callback
  without touching S-01.
- **Deviations and next action:** no product, owner, infrastructure, browser,
  duplicate-selector, cross-owner, or zero-row failure remains.
  `module.timeline` service-backed validation is inapplicable because transport,
  replay, and publication did not change. Begin S-03 only after Markdown passes.

### S-03 implementation evidence

- **Remediation:** `useTimelineKeyboardController` now returns the scalar-editor,
  collection-editor, and work-area handlers. It centralizes
  `mapWorkbookKeyboardCommand`, semantic anchors, editor commit-before-navigation,
  draft commit posture, adapter-owned Shift+Arrow ranges, Tab exit behavior,
  Escape focus restoration, Alt+H, Space, Ctrl/Cmd+K, selector-based inspector
  focus, and exact event consumption. Input, textarea, select, button, link,
  contenteditable, menu, dialog, listbox, and option owners retain their events.
- **Cleanup and compatibility:** the component-local common handler, three
  keyboard callbacks, selector focus callback, and now-unreachable editor
  history/evidence/quick-link branches were removed. Supported keyboard and
  focus behavior remains unchanged; the added dialog/listbox exclusions enforce
  the owner-required no-leak posture. Public selectors, props, exports, routes,
  saves, and focus anchors are unchanged.
- **Source owner and selector:** the controller and root-level Timeline test are
  explicitly owned by `web.workbook`. The test resides outside `timeline/hooks`
  because the repository's controller-isolation boundary correctly forbids a
  hook test from importing a sibling hook path. Row
  `web.workbook.regression.timeline_keyboard_event_and_focus_ownership_97dc4e9a17`
  selects the exact three titles covering scalar sequencing, collection
  sequencing/close, and work-area shortcut/event ownership. The S-00 mapping,
  sentinel, and browser rows retain Enter/Shift+Enter, Tab/Shift+Tab, Escape,
  Space, Ctrl/Cmd+K, Alt+H, focus recovery, and unavailable-action evidence.
- **Terminal results:** after renaming two lint-conflicting test variables,
  moving the test across the isolation boundary, and correcting two fixture
  types reported by the initial attempts, terminal `make frontend-typecheck`
  passed 2/2 at `.cartulary/test-results/20260814T124954Z-p2861467`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T124954Z-p2861454`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T125024Z-p2865432`;
  `make test-slice OWNER=web.workbook` passed 124/124 at
  `.cartulary/test-results/20260814T125024Z-p2865423`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T125519Z-p2935628`;
  `make browser-e2e-webserver-backed` passed 62/62 at
  `.cartulary/test-results/20260814T125107Z-p2880733`; and
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T125024Z-p2865342`.
- **Accounting, generation, and rollback:** `make test-catalog-check` and the
  owner explanation passed with 123 active `web.workbook` rows and no
  service-backed rows. `make generate` passed at
  `.cartulary/test-results/20260814T125009Z-p2862372`, changing only generated
  harness topology; `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T125519Z-p2935569`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T125519Z-p2935584`. Product-contract and
  generated-root impact is `none`. Rollback restores the component callbacks,
  removes the controller/test and same-slice accounting, and regenerates
  topology without touching S-01 or S-02.
- **Deviations and next action:** no product, accessibility, owner, browser,
  infrastructure, duplicate-selector, cross-owner, or zero-row failure remains.
  Begin S-04 only after Markdown validation.

### S-04 implementation evidence

- **Remediation:** `useTimelinePresenceController` now owns the initial/current
  Timeline presence draft and ref, exact `record_id` row lookup,
  `record_id + field_key + editing` cell lookup, viewing selection publication,
  edit-mode publication, and local surface/authorization lifecycle reset. Row
  selection consumes the semantic `publishViewingPresence` command instead of
  mutating presence state inside the inspector-selection hook.
- **Boundary posture:** the controller receives only
  `activeSheetPresenceRecords` from `useTimelineCollaborationBindings` and the
  semantic publisher. Full `sheet_ref` filtering, self filtering, keyed
  connection identity, ordering, coalescing, transport, replay, and
  authorization recovery remain unchanged in
  `WorkbookCollaborationCoordinator` and the existing binding. The controller
  imports no credentials, protocol envelopes, socket code, mutation runtime,
  or grid-vendor implementation.
- **Source owner and selector:** the two new collaboration paths are explicitly
  owned by `web.workbook`. Row
  `web.workbook.regression.timeline_presence_derivation_and_publication_eb6a4ede0e`
  selects the exact three colocated titles covering stable row/cell/editing
  identity under duplicate/reordered input, coherent viewing/editing
  transitions, edit-over-selection precedence, initial draft, and reset without
  publication. The S-00 collaboration row and browser suite retain direct
  view-schema versus saved-view scoping, self filtering, ambient save-state
  isolation, removal, and authorization recovery evidence.
- **Compatibility and risk posture:** wire messages, `sheet_ref`, connection
  keys, session timing, and transport behavior are unchanged. The reset key
  clears only local draft state and publishes nothing on surface or access-loss
  lifecycle change. Leaving the logic inline would keep two presence mutation
  owners and make stale indicators or cross-surface inference more likely.
- **Terminal results:** after adding an explicit reset-key read required by the
  exhaustive-dependency linter, terminal `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T130240Z-p2989706`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T130240Z-p2989718`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T130308Z-p2993734`;
  `make test-slice OWNER=web.workbook` passed 125/125 at
  `.cartulary/test-results/20260814T130308Z-p2993741`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T130805Z-p3064282`;
  `make browser-e2e-webserver-backed` passed 62/62 at
  `.cartulary/test-results/20260814T130353Z-p3009248`; and
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T130308Z-p2993692`.
- **Accounting, generation, and rollback:** `make test-catalog-check`, the
  `web.workbook` owner explanation with 124 active rows, and the
  `module.timeline` explanation with its unchanged 64 rows passed.
  `make generate` passed at
  `.cartulary/test-results/20260814T130253Z-p2990698`, changing only generated
  harness topology; `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T130805Z-p3064150`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T130805Z-p3064170`. Product-contract and
  generated-root impact is `none`. `module.timeline` service-backed validation
  is explicitly inapplicable because no boundary, binding, transport, replay,
  or service-backed publication code changed. Rollback restores the old inline
  state/lookups and selection callback without affecting S-01 through S-03.
- **Deviations and next action:** no product, owner, browser, infrastructure,
  duplicate-selector, cross-owner, or zero-row failure remains. Begin S-05 only
  after Markdown validation.

### S-05 implementation evidence

- **Remediation:** `useTimelineMentionActions` now exposes
  `handleUndoAutoResolutionNotice`. The command re-reads the current committed
  row and the exact relationship collection named by the notice, no-ops unless
  the current item is still `resolved_ref`, constructs the private
  `InspectorMention`, and submits `revert_to_unresolved` through the existing
  semantic action path. The component passes only `AutoResolutionNotice` and
  contains no mention-object fabrication.
- **Identity and sequencing:** the conversion preserves current raw text,
  source record/field/item, entity type, resolved target, mention row version,
  resolution method, provenance, confidence, and matched alias without treating
  the notice text or display label as identity. The queued action still waits
  for the committed record, re-reads the latest mention version, tracks the
  client transaction, applies the existing stale/rejection conflict path,
  requires the returned source-record version, refreshes, preserves viewport/
  inspector continuity, and settles Saved only after projection commit.
- **Source owner and selector:** the new root-level Timeline hook test is
  explicitly owned by `web.workbook`; the existing hook path retains its owner.
  Row
  `web.workbook.regression.timeline_mention_auto_resolution_undo_ownership_9e834f7b55`
  selects the exact three titles covering conversion/identity separation,
  invocation plus dispatch-time re-read and accepted continuity, and missing/
  unresolved no-op plus rejection conflict behavior. The S-00
  `mention_undo_current_committed_identity` row retains component focus, scroll,
  and current post-resolution version evidence.
- **Compatibility and risk posture:** route, action, payload, selectors, entity
  refresh, focus, scroll, and public interfaces are unchanged. The fabricated
  object remains private and is not a compatibility contract. Field-specific
  lookup now fails closed if a stale notice points at the other relationship
  collection, preventing host/identity meaning from drifting.
- **Terminal results:** `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T131557Z-p3115408`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T131557Z-p3115429`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T131625Z-p3119482`;
  `make test-slice OWNER=web.workbook` passed 126/126 at
  `.cartulary/test-results/20260814T131625Z-p3119479`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T131625Z-p3119500`; and
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T131625Z-p3119427`.
- **Accounting, generation, and rollback:** `make test-catalog-check` and the
  owner explanation passed with 125 active `web.workbook` rows and no
  service-backed rows. `make generate` passed at
  `.cartulary/test-results/20260814T131611Z-p3116389`, changing only generated
  harness topology; `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T131917Z-p3176648`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T131917Z-p3176627`. Product-contract and
  generated-root impact is `none`. Rollback restores component conversion and
  removes the hook command, test row/path, and regenerated topology without
  touching S-01 through S-04.
- **Deviations and next action:** no product, identity, owner, infrastructure,
  duplicate-selector, cross-owner, or zero-row failure remains. Browser rerun
  is unnecessary for this slice because the exact component focus/scroll test
  passed in `web.workbook` and no DOM handler changed. Begin S-06 only after
  Markdown validation.

### S-06 implementation evidence

- **Remediation:** `useTimelineInspectorFeatureController` now consumes the
  S-01 closed policy and the existing semantic Indicator and create-related
  workflow ports. It owns supported-action checks, Indicator-first dispatch,
  mutually exclusive selected Indicator state, unavailable and reset messages,
  explicit cancellation, and reset coordination. The component contains no
  feature-routing conditional or selected Indicator state.
- **Lifecycle and authorization posture:** the controller resets both workflow
  kinds when the selected subject or row version, canonical `sheet_ref`, shell/
  continuity lifecycle, inspector invalidation generation, incident role, or
  access-loss state changes. It uses `sheetRefKey`, so view-schema, saved-view,
  and extension-workspace identities are complete. It does not authorize an
  action, infer identity, touch transport, or bypass the server; existing
  contract disabled tokens and server outcomes remain authoritative.
- **Closed dispatch behavior:** canonical Indicator actions cancel generic
  creation before selecting their specialized handler. Canonical generic
  actions clear Indicator state and delegate the canonical feature group to the
  existing workflow. Unsupported or altered tuples clear both paths, show the
  existing unavailable message, and never fall back to generic creation. Close
  and Cancel now use the same controller cancellation command, preventing a
  stale specialized workflow from surviving inspector lifecycle changes.
- **Source owner and selectors:** both new paths are owned by `web.workbook`.
  Row
  `web.workbook.regression.timeline_inspector_feature_controller_a38c2d6f71`
  selects the exact three titles covering every canonical supported tuple and
  Indicator precedence, unsupported/no-generic-fallback plus cancellation and
  no-subject/disabled-boundary posture, and subject/version/surface/lifecycle/
  authorization resets. The S-00 component and browser inspector rows retain
  disabled-button, create-related success/failure, Indicator rendering, focus,
  and lifecycle evidence. Owner explanation reports 126 active
  `web.workbook` rows and no service-backed rows.
- **Compatibility and rollback:** public exports, callers, routes, feature
  tuples, payloads, selectors, result behavior, and authorization outcomes are
  unchanged. The controller accepts semantic callbacks and lifecycle
  identities only; it has no credentials, HTTP/WebSocket code, projection
  mutation, grid-vendor import, public export, flag, wrapper, or dependency.
  Rollback removes the controller/test/accounting and restores the former
  component state/callbacks plus lifecycle cancellation without touching S-01.
- **Terminal results:** `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T132610Z-p3187642`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T132537Z-p3186430`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T132737Z-p3192791`;
  `make test-slice OWNER=web.workbook` passed 127/127 at
  `.cartulary/test-results/20260814T132737Z-p3192798`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T132737Z-p3192810`;
  `make browser-e2e-webserver-backed` passed 62/62 at
  `.cartulary/test-results/20260814T133024Z-p3249733`; and
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T132711Z-p3191208`.
- **Accounting, generation, and failures:** `make test-catalog-check` and the
  applicable owner explanations passed without unknown, duplicate,
  cross-owner, or zero-row selection. `make generate` passed at
  `.cartulary/test-results/20260814T132656Z-p3188248`, changing only generated
  harness topology; `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T133433Z-p3303150`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T133433Z-p3303149`. Product-contract and
  generated-root impact is `none`. The initial typecheck failed at
  `.cartulary/test-results/20260814T132537Z-p3186429` because the first surface
  key expression omitted extension-workspace `SheetRef`; using canonical
  `sheetRefKey` resolved the defect. No failure remains. Begin S-07 only after
  Markdown validation.

### S-07 implementation evidence

- **Composition result:** the source audit confirms that the S-01 through S-06
  changes already completed the required cleanup, so no additional production
  edit was necessary in this slice. `TimelineWorkbookContent` now composes the
  runtime, semantic adapters/controllers, collaboration bindings, inspector and
  grid child surfaces, local load/layout presentation, and rendering. The
  registered `TimelineWorkbookProps` and `TimelineWorkbook` facade and its sole
  production caller remain unchanged; no module named after the file was added.
- **Residual-policy audit:** the target contains no feature-routing matrix,
  permissive route fallback, fill planner, keyboard command policy, mention
  object construction, presence state machine, raw HTTP/WebSocket lifecycle,
  generated-protocol translation, or grid-vendor import. Its grid types come
  only from `@cartulary/grid-adapter`, the project semantic boundary. The local
  `ResizeObserver` and grid-shell width state remain because Section 3
  explicitly identifies viewport measurement as legitimate presentation state
  with no evidenced reusable owner.
- **Size and cohesion:** the target is now 1,619 lines versus the 2,176-line
  starting inventory. Its cumulative diff is 112 additions and 669 deletions;
  the added lines are controller composition and semantic delegation. Direct
  `mutationCommands` references are port injection or client-transaction
  identity wiring, while mutation admission, pending replay, conflict handling,
  query reconciliation, and transport behavior remain in their existing hooks.
- **Source, compatibility, and rollback:** the source owner remains
  `web.workbook`, no new S-07 path or selector required accounting, and the
  production facade import policy still permits only
  `WorkbookSurfacesFacade.tsx`. Public props/exports, routes, envelopes,
  selectors, identities, focus behavior, accessibility, authorization, and
  package exports are unchanged. There is no S-07 product patch to roll back;
  independently complete controllers are deliberately retained.
- **Terminal results:** `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T133647Z-p3308763`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T133647Z-p3308668`;
  `make lint-biome` passed 2/2 at
  `.cartulary/test-results/20260814T133647Z-p3308693`;
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T133647Z-p3308445`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T133708Z-p3310592`;
  `make test-slice OWNER=web.workbook` passed 127/127 at
  `.cartulary/test-results/20260814T133708Z-p3310601`;
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T133708Z-p3310603`; and
  `make browser-e2e-webserver-backed` passed 62/62 at
  `.cartulary/test-results/20260814T133923Z-p3346020`.
- **Accounting, generation, and next action:** `make test-catalog-check`
  passed; no owner, selector, generated input, or generated output changed in
  S-07, so generation impact is `none` for this slice. Source searches and
  `git diff --check` found no prohibited residual or whitespace error. No
  deviation or failure remains. Begin S-08 only after Markdown validation.

### S-08 final accounting and implementation handoff

- **Authority and source identity:** implementation task
  `user-timeline-workbook-remediation-2026-08-14` completed the authorized
  S-00 through S-08 range. The worktree is based on immutable source revision
  `5d070c2a9970049825d49c78c0fa9b6f84b02fad`; no commit was created by this
  task. Core 01 through Core 04, the adopted subsystem owners, and
  `docs/domain.md` remained consistent. Domain vocabulary is unchanged and no
  Core or domain specification edit was required.
- **Final production files:**
  `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`,
  `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`, and
  `apps/web/src/workbook/timeline/hooks/useTimelineMentionActions.ts` changed;
  the new production paths are
  `apps/web/src/workbook/timeline/bulk/useTimelineFillController.ts`,
  `apps/web/src/workbook/timeline/collaboration/useTimelinePresenceController.ts`,
  `apps/web/src/workbook/timeline/hooks/useTimelineInspectorFeatureController.ts`,
  `apps/web/src/workbook/timeline/hooks/useTimelineKeyboardController.ts`, and
  `apps/web/src/workbook/timeline/models/timelineWorkbookFeaturePolicy.ts`.
- **Final test and accounting files:** the new tests are the colocated fill,
  presence, and feature-policy tests plus
  `apps/web/src/workbook/timeline/useTimelineInspectorFeatureController.test.tsx`,
  `apps/web/src/workbook/timeline/useTimelineKeyboardController.test.tsx`, and
  `apps/web/src/workbook/timeline/useTimelineMentionActions.test.tsx`.
  Authored accounting changed only `tools/frontend_source_ownership.json` and
  `tools/test_families/web.workbook.json`; Make refreshed generated
  `tools/execution_topology_render_index.json`; this tracker is the sole prose
  change. No generated product root, dependency lockfile, backend, SQL,
  configuration, public package, or adopted specification changed.
- **Ownership and selectors:** every new path has exactly one `web.workbook`
  entry. The exact active rows are
  `timeline_feature_policy_closed_semantic_routing_bbd0ade7b2`,
  `timeline_fill_planning_and_semantic_dispatch_c446e90622`,
  `timeline_keyboard_event_and_focus_ownership_97dc4e9a17`,
  `timeline_presence_derivation_and_publication_eb6a4ede0e`,
  `timeline_mention_auto_resolution_undo_ownership_9e834f7b55`, and
  `timeline_inspector_feature_controller_a38c2d6f71`, each under the
  `web.workbook.regression` prefix and each selecting the full test titles
  recorded in its originating slice. The S-00
  `mention_undo_current_committed_identity_ddf69011e7` row remains active.
  `make test-catalog-check` passed with no unknown, duplicate, cross-owner, or
  zero-row selection. Owner explanations report 126 `web.workbook` rows, 89
  `module.workbook` rows, and 64 `module.timeline` rows.
- **Focused final gate:** `make frontend-typecheck` passed 2/2 at
  `.cartulary/test-results/20260814T134444Z-p3401062`;
  `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260814T134444Z-p3401085`;
  `make json-shape-check` passed 3/3 at
  `.cartulary/test-results/20260814T134444Z-p3400845`;
  `make test-slice OWNER=web.architecture` passed 12/12 at
  `.cartulary/test-results/20260814T134506Z-p3403405`;
  `make test-slice OWNER=web.workbook` passed 127/127 at
  `.cartulary/test-results/20260814T134506Z-p3403411`; and
  `make test-slice OWNER=module.workbook` passed 100/100 at
  `.cartulary/test-results/20260814T134506Z-p3403425`.
- **Timeline and browser evidence:** the first final
  `make test-slice OWNER=module.timeline` run completed 79/81 units at
  `.cartulary/test-results/20260814T134506Z-p3403435`; its only semantic row
  failure was the paint-qualification timeout for
  `timeline_blank_row_creation_satisfies_the_paint_afddd2ce13`, while all
  functional Timeline rows passed. The exact failed measurement row then
  passed in isolation, 14/14 units, at
  `.cartulary/test-results/20260814T135149Z-p3476414`, resolving the
  load-sensitive contention without a product change. The standalone
  `make browser-e2e-webserver-backed` gate passed 62/62 at
  `.cartulary/test-results/20260814T135717Z-p3500698`. Service-backed
  `module.timeline` testing is inapplicable because transport, session, replay,
  and service-backed publication behavior did not change.
- **Generation and broad final gate:** final generation impact is `none` for
  product contracts and generated roots. The authored test-family change has
  only its expected Make-produced harness topology projection.
  `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260814T135717Z-p3500457`, and
  `make generated-artifact-policy-check` passed 3/3 at
  `.cartulary/test-results/20260814T135717Z-p3500469`.
  `make agent-finalize` passed 1/1 at
  `.cartulary/test-results/20260814T140124Z-p3557145`; retained-run maintenance
  was skipped because `RESULTS_DIR` was unset. `make check` passed 754/754 at
  `.cartulary/test-results/20260814T140149Z-p3560008`.
- **Compatibility, risk, and completion:** public interfaces, callers, routes,
  wire envelopes, semantic identities, selectors, authorization outcomes, and
  package exports are unchanged. No compatibility wrapper, feature flag,
  dependency, universal dispatcher, backend change, or migration was added.
  Every slice remains independently reversible as recorded in the ledger.
  `git diff --check` passes, and the final worktree has only the intentional
  authored files and one Make-generated harness topology file listed above.
  RB-001 through RB-003 and DOD-001 through DOD-012 are resolved. The completed
  record passed `make lint-markdown` at
  `.cartulary/test-results/20260814T141054Z-p3716091`; the post-result root
  insertion is revalidated by the final delivery root.

## 8. Validation Plan

These commands were discovered from the live Make task surface. A later
authorized implementation MUST execute the following sequence in order:

1. Record the complete RB-001 authority and live starting revision.
2. Re-read the source-owner and `web.workbook`, `module.workbook`, and
   `module.timeline` test-family catalogs.
3. Add or confirm all five S-00 seam families with exact same-slice accounting.
4. Run and retain the pre-extraction baseline.
5. Resolve RB-001, RB-002, and RB-003 before S-01 moves production behavior.
6. Execute S-01 through S-07 as independently reversible checkpoints.
7. Perform the S-08 reconciliation audit.
8. Run and retain the focused-to-broad final gate.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=web.workbook` | Focused frontend Workbook rows | yes | MUST pass before production movement and after every production slice. |
| integration | `make test-slice OWNER=module.workbook` | Cross-module Workbook postconditions | yes | MUST pass in the pre-extraction baseline and after slices that affect its postconditions. |
| service-backed | `make service-backed-test-slice OWNER=module.timeline` | Timeline collaboration transport, replay, or service-backed publication | conditional | Mandatory only if S-04 changes the collaboration boundary, bindings, transport lifecycle, replay, or service-backed publication path. A pure derivation extraction does not trigger it. |
| e2e/browser | `make browser-e2e-webserver-backed` | User-visible focus, keyboard, presence, inspector, Timeline, collaboration, and recovery behavior | conditional | Mandatory before S-03 or S-04 when owner-qualified Vitest rows do not fully demonstrate their behavior, after those applicable slices, and for final browser confidence. |
| generated drift | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated projections of changed authored inputs | conditional | Run only when an authored generator input changes. Otherwise record `generation impact: none`. |
| import-boundary/static | `make frontend-typecheck`; `make frontend-import-boundary-check` | Type safety and frontend source/import ownership | yes | MUST pass in the baseline and after every production slice. |
| owner accounting | `make explain-test-owner OWNER=web.workbook`; corresponding `module.workbook` and `module.timeline` commands when selected | Exact owner, stable row, selector, collaborator, and service-backed disposition | yes | MUST prove no unknown, duplicate, cross-owner, or accidental zero-row selection. |
| full check | `make agent-finalize` followed by `make check` | Harness maintenance and repository developer verification gate | no | Mandatory after S-08. Pass `RESULTS_DIR` for retained successful full warm-check evidence; otherwise explicitly record that retained-run maintenance was skipped. |
| tracker Markdown | `make lint-markdown` | Authored Markdown structure | no | Required for tracker revisions when documentation maintenance is authorized. It does not validate product behavior. |

Before adding S-00 tests, the implementation MUST run the three owner task
guides and owner explanations. Before any production extraction, it MUST run:

```bash
make frontend-typecheck
make frontend-import-boundary-check
make test-slice OWNER=web.workbook
make test-slice OWNER=module.workbook
```

After each production slice, it MUST rerun frontend type checking, frontend
import boundaries, and the `web.workbook` slice. It MUST add the applicable
`module.workbook`, `module.timeline`, browser, and generation commands from the
table rather than treating them as unconditional or silently optional.

### Mandatory evidence record

A command exit code alone is insufficient. S-00 and final handoffs MUST record:

| Evidence field | Required value |
| --- | --- |
| Source identity | Exact commit or immutable source snapshot used by the run |
| Catalog identity | Current verification owner, test-family, and execution-topology identity used by the run |
| Selected owners | `web.workbook` and every applicable collaborator |
| Selected semantic rows | Exact stable row IDs; a broad target name is insufficient |
| Test selectors | Exact file and full Vitest titles or exact browser scenario IDs |
| Terminal results | One terminal result for every selected row |
| Retained roots | Exact run root for every required evidence class |
| Infrastructure status | No unresolved `infrastructure_failed`, `cancelled`, or accidental zero-row selection |
| Characterization timing | Evidence that the test passed before the corresponding production extraction |
| Deviations | Every pre-existing failure or waived SHOULD requirement, separately identified and justified |
| Generation impact | Exact generated outputs and commands, or the explicit value `none` |

Read-only discovery commands run in this planning session were Git status,
branch, and revision inspection; targeted `rg`, `sed`, `find`, `wc`, and `jq`
inspection; `make help`; `make help-all`; `make task-guide ROLE=module-author`
for `module.timeline`, `module.workbook`, and `web.workbook`; and
`make explain-test-owner` for those three owners. The batched
`make explain-target ... DETAIL=summary` output was truncated, so it supplies no
claimed validation result. No product test, build, lint, code-generation, or
full-check result is claimed.

This NLSpec-style tracker revision ran `make lint-markdown` and passed with
retained run root
`.cartulary/test-results/20260814T013042Z-p410027`. This is documentation
validation only; it supplies no product-behavior or implementation evidence.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TW-001 | Establish scope, authority, clean state, and safe target identifier | WF-00 | DONE | None | Section 1; implementation base revision `5d070c2a9970049825d49c78c0fa9b6f84b02fad` | Planning boundaries and source hierarchy are explicit. |
| TW-002 | Inventory the complete target and its callers/dependencies | WF-01 | DONE | TW-001 | Section 2 | The sole target file is fully inventoried. |
| TW-003 | Map observable contracts to owners and tests | WF-02 | DONE | TW-002 | Section 4 | Every discovered contract risk has an owner and test posture. |
| TW-004 | Define the five-seam characterization contract | WF-03 | DONE | TW-002 | Sections 4, 7, and 8 | Fill, keyboard/focus, presence, mention undo, and inspector routing have explicit current-behavior cases, owners, commands, and evidence fields. |
| TW-005 | Diagnose coupling and permanent-boundary posture | WF-04 | DONE | TW-002 | Sections 3 and 5 | Legitimate composition and residual controller ownership are separated. |
| TW-006 | Define facade and residual-owner plan | WF-05 | DONE | TW-003, TW-004, TW-005 | Section 3 | The facade remains stable and no new permanent module is proposed. |
| TW-007 | Sequence behavior-preserving slices | WF-06 | DONE | TW-006 | Section 7 | Every slice has dependencies, validation, rollback, and exit criteria. |
| TW-008 | Define same-slice source/test accounting and final reconciliation | WF-07 | DONE | TW-004, TW-007 | Sections 5, 7, and 8 | Every planned path has a default owner and each slice must land exact accounting before S-08. |
| TW-009 | Resolve the bounded implementation-authority gate | WF-08 | DONE | TW-007 | RB-001; S-00 evidence | Task ID, slice range, artifact bounds, start revision, and S-00 prerequisite are recorded. |
| TW-010 | Resolve characterization and ownership preflight gates | WF-08 | DONE | TW-009 | S-00, RB-002, RB-003 | All five seam families pass on the pre-extraction implementation with exact owner/row/selector/run-root evidence. |
| TW-011 | Implement S-01 through S-07 independently | WF-08 | DONE | TW-010 | S-01 through S-07 evidence and checkpoint ledger | Each extraction satisfies its binary completion criterion without contract change. |
| TW-012 | Complete conditional accounting and final validation | WF-08 | DONE | TW-011 | S-08 handoff and retained validation roots | Required focused-to-broad commands and ownership checks pass. |
| TW-013 | Change public behavior, routes, schemas, events, selectors, or authorization outcomes | None | DEFERRED | Separate authorization and owner evidence | `requires later authorization` | A separately approved behavior-change plan names the owner and migration posture. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Authority order and planning-only boundary established; no contradiction found. | Inspected framework, Core 00-04, Testing Harness NLSpec, domain, design, and dev guide. Touched: this tracker only. | `sed`, `rg`, `find`; Git state/revision inspection | Target and applicable owners confirmed; Core 05 is inapplicable. | RB-001 | Under later authorization, start with S-00 rather than production movement. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Authority is expressed as a complete task-level record with defaults and stop conditions; the tracker remains non-authorizing. | Inspected this tracker, `docs/research/nlspec-spec.md`, `temp/analysis-notes.md`, and live owner evidence. Touched: this tracker only. | `sed`, `rg`, `wc`, `jq`; Git branch/status/revision inspection | Writing doctrine and review evidence were separated from product authority; no owner contradiction was found. | RB-001 | Record a qualifying implementation task before any S-00 or production write. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | No backend module, SQL/storage, raw route, or direct transport logic is owned by the target. | Inspected workbook operation/query/mutation adapters and relevant OpenAPI/protocol owner projections. Touched: this tracker only. | Targeted `rg`, `sed`, and `jq` | Semantic ports already isolate route and generated protocol behavior. | RB-002 | Freeze operation semantics in characterization evidence; make no backend edits for the extraction. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Backend, persistence, projection, transport, and authorization boundaries are explicit MUST NOT move invariants. | Reused prior backend inventory; inspected the current mention and collaboration semantic boundaries. Touched: this tracker only. | Targeted `rg` and direct source reads | No backend artifact is required by the planned extraction; generation impact defaults to `none`. | RB-002 | Stop and reopen planning if any slice requires a backend or public-contract change. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Private facade is legitimate; implementation body is an oversized mixed-responsibility controller. | Inspected target, facade, runtime fixture/support, runtime model, Timeline hooks/adapters, layout and ownership policies. Touched: this tracker only. | Full target read; targeted `rg`, `sed`, `wc` | Preserve exports/callers and extract residual controllers to existing Timeline owner areas. | RB-001, RB-002 | Implement S-01 through S-07 only after S-00 passes. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Private input/output contracts, defaults, prohibited couplings, and path owners are defined for every extraction. | Inspected target fill, keyboard, presence, mention-undo, and inspector-routing code plus supporting types. Touched: this tracker only. | `sed`, `rg`, `jq` | All planned paths default to live owner `web.workbook`; public exports and callers remain frozen. | RB-001, RB-002, RB-003 | Resolve the three gates, then execute one same-slice-owned extraction at a time. |
| 2026-08-14 10:10:09 EDT | Codex implementation session, S-01 through S-08 | Timeline feature, fill, keyboard, presence, mention-undo, and inspector policy now live in closed private owners; the facade is composition-only. | Touched exactly the final production and test paths enumerated in S-08 plus authored ownership/catalog inputs, generated harness topology, and this tracker. | Slice-scoped type/import, owner, unit, module, browser, source-review, and broad-check gates | Public facade and behavior remain stable; no residual inline owner or compatibility layer remains. | None | Use the completed handoff; no follow-on remediation is required. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Protocol, view, UI, and selector contracts are frozen consumers; no codegen change is expected. | Inspected generated-artifact policy, package imports, HTTP-operation projection, and OpenAPI owner inputs. Touched: this tracker only. | Targeted `rg`, `sed`, `jq`; batched `make explain-target` attempt | No generated hand edit is planned. Explain-target output was truncated and is not validation evidence. | RB-003 if authored inputs later change | Run `make generate-drift` only when a later slice changes an authored generator input. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Generation has an explicit default and applicability rule; generated projections remain downstream. | Inspected view-contract feature types, UI/grid contract types, generated policy, and live ownership inputs. Touched: this tracker only. | `sed`, `rg`, `jq` | `generation impact: none` is required unless an authored generator input changes; hand edits remain prohibited. | RB-003 | If an authored input changes, run the complete Make-owned generation/drift/policy sequence and record outputs. |
| 2026-08-14 10:10:09 EDT | Codex implementation session, S-08 | Product-contract and generated-root impact is `none`; only the expected harness topology projection follows the authored test-family update. | Authored source ownership and `web.workbook` family inputs; generated topology render index | `make generate`; drift, artifact-policy, JSON-shape, and final diff checks | Make-generated topology is current and all protected generated roots are untouched. | None | No generation or migration follow-up. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Broad current coverage exists; direct seam coverage is required before risky movement. | Inspected WorkbookShell, focused Timeline, browser, owner catalog, and test-family evidence. Touched: this tracker only. | `make help`; `make help-all`; three `make task-guide` and three `make explain-test-owner` commands; source searches | Commands discovered and owner rows explained. No product test or validation suite was run. | RB-002, RB-003 | Run S-00 focused baseline and account for every new test/source path. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | S-00 now has five mandatory seam families, exact owner rules, applicability rules, and an evidence-record schema. | Inspected `web.workbook`, `module.workbook`, and `module.timeline` family manifests and representative exact selectors. Touched: this tracker only. | `sed`, `rg`, `jq`; `make lint-markdown`; no product suite | Test ownership is determined by postcondition and must land with each selector. Markdown lint passed at `.cartulary/test-results/20260814T013042Z-p410027`; no product validation success is claimed. | RB-002, RB-003 | Add/confirm tests on the pre-extraction source, then retain exact row and run-root evidence. |
| 2026-08-14 08:20 EDT | Codex implementation session, S-00 | Pre-extraction authority, characterization, ownership, browser, and topology gates pass. | Touched authored `tools/test_families/web.workbook.json`, generated `tools/execution_topology_render_index.json`, and this tracker. | Static gates, three focused owners, browser E2E, owner explanations, `make generate`, drift/policy/shape checks | Five seam families are retained before production movement; exact evidence and the one repaired topology-staleness failure are recorded above. | None | Run Markdown validation, then begin S-01. |
| 2026-08-14 10:10:09 EDT | Codex implementation session, S-08 | Exact source ownership, selectors, collaborators, browser evidence, finalization, and repository-wide verification pass. | All final tests and accounting artifacts enumerated in S-08 | Final focused owner slices; owner explanations; browser E2E; `make agent-finalize`; `make check` | Final check passed 754/754; the sole load-sensitive Timeline measurement timeout passed on isolated rerun and is resolved. | None | Run the final tracker Markdown checkpoint and hand off. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Target has UI editable/role gates and access-loss recovery; server operations remain authoritative. | Inspected Core 04, target role gates, evidence adapter path, collaboration boundary/hooks, and recovery tests. Touched: this tracker only. | Targeted `rg` and direct file reads | No authorization check is proposed to move or change. | RB-002 | Preserve denial and incident-access-loss outcomes in characterization and browser evidence. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Security invariants now prohibit credentials, transport ownership, permissive fallback, and client authorization ownership in extracted controllers. | Reused Core 04 evidence and inspected current presence/mention interfaces. Touched: this tracker only. | Targeted `rg` and direct source reads | Server authority and fail-closed unsupported-feature behavior are mandatory preserved boundaries. | RB-002 | Characterize authorization-loss cleanup and stop any slice that requires an authorization change. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Planning is complete; implementation remains unauthorized and unvalidated. | Inspected all evidence summarized in Sections 1-5. Touched: this tracker only. | Read-only discovery only; final Git status inspection | Tracker provides slice dependencies, rollback, validation, and binary exits without a production patch. | RB-001, RB-002, RB-003 | Obtain authorization, recheck live state, run S-00, then execute one reversible slice at a time. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Specification gaps are closed; execution gates remain unresolved by design. | Inspected this tracker, NLSpec doctrine, analysis notes, target seams, types, source ownership, and test-family manifests. Touched: this tracker only. | Read-only analysis commands and `make lint-markdown` | The tracker now specifies authority, five-seam evidence, private interfaces, same-slice accounting, command applicability, and Definition of Done; documentation lint passed. | RB-001, RB-002, RB-003 | Resolve gates in order; do not move production behavior on missing or inferred evidence. |
| 2026-08-14 10:10:09 EDT | Codex implementation session, final handoff | S-00 through S-08 and DOD-001 through DOD-012 pass; RB-001 through RB-003 are resolved. | Final worktree, tracker, ownership/catalog inputs, and retained results | `git status`; `git diff --check`; focused-to-broad Make gates | No owner contradiction, unintended/generated edit, unresolved failure, migration, or compatibility debt remains. | None | Ready for review and commit; no deferred implementation work. |

## 11. Open Questions and Blockers

These entries are implementation-readiness gates, not unresolved product-design
questions. Their closure rules are exhaustive; an implementer MUST NOT substitute
informal approval, a broad passing command, or inferred ownership.

| ID | Resolved decision and required closure evidence | Current status | Final status rule |
| --- | --- | --- | --- |
| RB-001 | Implementation is authorized only by a task adopting an exact S-00 through S-08 range and the Section 1 allowed/prohibited scope. The tracker MUST record the task ID, slice range, starting revision, artifact bounds, and S-00-before-movement rule. | `RESOLVED` by `user-timeline-workbook-remediation-2026-08-14` | Set `RESOLVED` only when every authority-record field is populated by a qualifying task. |
| RB-002 | S-00 covers fill, keyboard/focus, presence, mention undo, and inspector routing. Every new characterization test MUST pass against the pre-extraction implementation, and exact owners, stable rows, selectors, terminal results, applicable browser/service evidence, and retained roots MUST be recorded. | `RESOLVED` by the S-00 evidence record | Set `RESOLVED` only when all five seam families and every mandatory evidence-record field pass without encoding desired new behavior. |
| RB-003 | Every new path MUST have one live-catalog-confirmed owner; every exact test selector MUST have one active semantic row; accounting MUST land with the originating slice; generated outputs MUST come only from Make and be clean or explicitly inapplicable. | `RESOLVED` through final S-08 reconciliation | Set `RESOLVED` before S-01 only when planned paths/tests have a recorded disposition, then keep it resolved by satisfying the same-slice rule through S-08. Any later accounting defect reopens the gate. |

No owner contradiction was discovered during planning, so no
`BLOCKED: owner contradiction` entry exists. If implementation discovers one,
the affected slice MUST stop and the marker MUST be added here. The universal
feature-dispatch abstraction remains intentionally out of scope; the binding
default is a Timeline-owned first extraction.

## 12. Binary Completion Criteria

The following table is the Definition of Done for the implemented refactor.
Every row is binary and has final evidence in the originating slice or S-08.

| ID | Normative requirement source | Acceptance criterion | Required evidence | Current disposition |
| --- | --- | --- | --- | --- |
| DOD-001 | Sections 1-2 | The sole target file remains fully inventoried, and `TimelineWorkbookProps`, `TimelineWorkbook`, all production callers, routes, schemas, events, selectors, identifiers, authorization outcomes, and package exports are unchanged. | Source diff, typecheck, import-boundary result, contract drift disposition, and caller inventory | PASS in S-07 and S-08 |
| DOD-002 | Sections 1 and 11, RB-001 | A qualifying task records the exact authority ID, slice range, source revision, allowed/prohibited artifacts, and S-00 prerequisite before any implementation write. | Completed authority record and handoff reference | PASS in S-00 |
| DOD-003 | Section 4 and S-00 | Fill, keyboard/focus, presence, mention undo, and inspector routing each have isolated current-behavior evidence that passes before the corresponding production move. | Exact tests, stable rows, terminal results, applicable browser/service results, and retained roots | PASS in S-00 |
| DOD-004 | Section 3 | Every extracted responsibility satisfies its closed private input/output contract, default, unsupported result, and prohibited-coupling rules without a new public export. | Focused interface tests, typecheck, import-boundary result, and source review | PASS in S-01 through S-07 |
| DOD-005 | Section 5 and S-01 through S-08 | Every new source path has exactly one live owner, every exact selector has exactly one active semantic row, and accounting lands with the originating slice. | Source-owner catalog result, three applicable owner explanations, selector disposition, and no unknown/duplicate/cross-owner/zero-row result | PASS in S-08 |
| DOD-006 | Sections 1, 4, and 5 | Server authorization, fail-closed unsupported-feature handling, stable semantic identities, transport/persistence/grid boundaries, and mention/entity separation remain unchanged. | Characterization rows, security/recovery evidence, import review, and public-contract comparison | PASS in S-01, S-04, S-05, S-06, and S-08 |
| DOD-007 | Section 7 | S-01 through S-07 are independently reversible and each satisfies its completion criterion before the next dependent slice begins. | Slice-scoped diffs, results, rollback note, ownership/accounting checkpoint, and handoff entry | PASS in the checkpoint ledger |
| DOD-008 | Section 8 | Generation is either explicitly `none` or all changed authored inputs are regenerated through Make with clean drift, artifact-policy, and JSON-shape results; no generated file is hand-edited. | Generation-impact record and applicable command artifacts | PASS in S-08 |
| DOD-009 | Section 8 | All mandatory baseline, per-slice, applicable browser/service-backed, `make agent-finalize`, and `make check` commands have terminal results with no unresolved infrastructure failure, cancellation, or accidental zero-row selection. | Mandatory evidence record and retained run roots | PASS in S-08 |
| DOD-010 | Sections 6, 9, and 10 | Workstream, tracker, and session records reflect the final live state, command results, changed files, generation disposition, resolved/reopened gates, failures, and next action without requiring rediscovery. | Current tracker and final implementation handoff | PASS in S-08 |
| DOD-011 | Sections 1 and 11 | No owner contradiction remains unresolved. Any discovered contradiction stopped its slice and was resolved by the applicable owner before work resumed. | Owner-resolution reference or explicit `no contradiction found` record | PASS; no contradiction found |
| DOD-012 | Entire tracker | No behavior-changing work was smuggled into the refactor. Any required new route, schema, error, state, default, permission, dependency, package export, feature tuple, or permanent cross-owner boundary was separately planned and authorized. | Final diff classification and owner approval for every authorized exception | PASS in S-08 |

The S iteration is complete: its final post-handoff `make lint-markdown` passed,
and DOD-001 through DOD-012 have passing implementation evidence. Sections 13
through 17 define a separate, not-yet-authorized production-readiness
iteration; they do not alter the completed S dispositions.

## 13. Controlling Production-Readiness Iteration

PR-00 through PR-06 are the controlling plan for the next iteration. Their goal
is to remove proven dead code, duplicated projections, cross-surface UI
ownership, and hidden render-order dependencies without preserving an internal
shape solely because it exists. The completed S evidence remains valid history;
it is not implementation authority for this iteration.

### Planning and future implementation authority

| Authority field | Required value | Current value |
| --- | --- | --- |
| Planning task ID | Stable identifier for this document update | `user-timeline-workbook-production-readiness-plan-2026-08-14` |
| Planning baseline | Exact source revision inspected while defining the iteration | `98082ac04c2a4e8a03df3a0982e30a7de12680f5` |
| Current authorized artifacts | Exact files this task may change | The allowed artifacts in the next row, adopted for PR-00 through PR-06 |
| Implementation authorization | A later user task adopting an exact PR-00 through PR-06 range | `GRANTED` by `user-timeline-workbook-production-readiness-remediation-2026-08-14` for PR-00 through PR-06 sequentially |
| Implementation starting identity | Live revision recorded before the first authorized production or test write | `98082ac04c2a4e8a03df3a0982e30a7de12680f5`; the only pre-existing worktree change was the staged planning update to this tracker |
| Future allowed artifacts | Timeline package source; the narrow workbook pending-runtime and common relationship-chip boundaries named by PR-02 and PR-04; direct production consumers needed to remove a reversed dependency; focused tests; `tools/frontend_import_boundaries.json`; authored source-ownership and test-family inputs; `apps/web/src/README.md`; Make-generated test topology when its authored input changes; this tracker | Adopted for the authorized PR-00 through PR-06 range |
| Future prohibited artifacts | Core and domain specifications; backend, protocol, persistence, projection, authorization, deployment, dependency, route, wire, selector, public-package, or unrelated workbook behavior; hand-edited generated roots | Binding default; a required exception stops the affected slice and reopens planning |
| Characterization prerequisite | Stable pre-change evidence for every behavior adjacent to a deletion or runtime-graph change | PASS in PR-00; retained roots are recorded below |

The future implementation MUST preserve `TimelineWorkbookProps`,
`TimelineWorkbook`, the sole supported production-facade import path, current
callers, routes, envelopes, selectors, `view_schema_id`, `sheet_ref`,
`record_id`, `row_version`, field keys, authorization outcomes, accessibility
semantics, and package exports. No compatibility wrapper, dual path, feature
flag, dependency, data migration, or generalized cross-workbook dispatcher is
planned.

### Deletion standard

A production path MAY be deleted only when PR-00 records all of the following:

- no adopted owner requires the behavior;
- no production caller or stable contract depends on it;
- it supplies no security, authorization, accessibility, recovery, or
  interoperability property;
- it has no demonstrated continuing user value; and
- its removal does not make an authorized future phase more brittle.

Test-only use does not create production ownership. A test that exists only to
exercise an unused production export MUST be removed or rewritten around the
surviving semantic owner. Missing evidence is not proof that behavior is dead.

The following are active semantics, not legacy-removal candidates:

- stale conflict-token refresh that preserves the current draft;
- fail-closed handling of unsupported or mismatched feature identities;
- authentication and authorization recovery and cleanup;
- saved-view versus base-surface `sheet_ref` and presence distinctions; and
- conflict, pending-replay, focus, keyboard, and accessibility outcomes owned by
  adopted specifications.

No literal obsolete Timeline protocol path was found at the planning baseline.
For this iteration, legacy burden means unused exports, duplicated semantic
projections, reversed package dependencies, placeholder callback bridges, and
forwarding layers without a continuing invariant.

## 14. PR Workstreams

The slices are sequential and independently reversible:

`PR-00 -> PR-01 -> PR-02 -> PR-03 -> PR-04 -> PR-05 -> PR-06`

### PR-00 - Baseline and deletion-proof gate

- **Areas:** Read-only production/reference inventory, characterization tests,
  ownership and verification planning, and tracker evidence.
- **Remediation:** Record the later implementation task, authorized slice range,
  live revision, artifact bounds, and prohibited changes. Inventory production
  references, test-only exports, duplicate helpers and labels, mutable callback
  bridges, cross-surface imports, lifecycle fallbacks, and orphan files.
  Characterize draft-row continuity, width measurement, row refresh and bounded
  stale retry, pending replay, conflict recovery, inspector reset, scalar and
  collection editing, and Entity-dependent Timeline previews. Add focused tests
  only where existing evidence cannot distinguish supported behavior from an
  implementation detail selected for deletion.
- **Rationale and benefit:** Deletion becomes evidence-based and cannot
  accidentally remove recovery, accessibility, or future-facing behavior.
- **Compatibility and migration:** Tests, authored verification metadata, and
  the tracker only. There is no product or data migration.
- **Risk if unresolved:** Active behavior can be mislabeled as legacy, while
  genuinely dead compatibility burden remains unmeasured.
- **Exit criteria:** Every PR-01 through PR-04 candidate has an owner, caller,
  deletion, preservation, or relocation disposition. Applicable pre-change
  tests pass with exact owner rows, titles, results, and retained roots. The
  implementation authority record is complete.

#### PR-00 implementation checkpoint

- **Status and dependency:** COMPLETE at source identity
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5`; PR-01 is unblocked.
- **Authority:**
  `user-timeline-workbook-production-readiness-remediation-2026-08-14`,
  PR-00 through PR-06 sequentially.
- **Files changed:** This tracker only. No production, test, owner, catalog, or
  generated artifact changed in PR-00.
- **Deletion-proof dispositions:** Duplicate draft allocation, the second
  width effect, test-only focus/draft helpers, the identity normalizer, copied
  Indicator field labels, the local editor-surface type, repeated identical
  styles, unused `useTimelineSaveStatePresentation` callback refs, callback
  mirrors, and forwarding-only imports are implementation details with no
  public caller or owner requirement. Conflict recovery, bounded stale retry,
  draft preservation, pending FIFO replay, inspector focus/reset semantics,
  relationship-chip accessibility, Entity Timeline previews, and width
  measurement remain active behavior and MUST be preserved.
- **Ownership and selectors:** Existing source paths have the single live owner
  `web.workbook`. Existing characterization rows remain singly owned by
  `web.workbook`, `web.architecture`, `module.workbook`, or `module.timeline` as
  recorded in their authored family manifests. No selector changed.
- **Commands:** `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T145146Z-p3737270`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T145146Z-p3737295`;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T145146Z-p3737188`;
  `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T145155Z-p3738254`;
  `make test-slice OWNER=web.workbook` PASS, 127/127 units, at
  `.cartulary/test-results/20260814T145155Z-p3738334`; and
  `make test-catalog-check` PASS.
- **Generation impact:** `none`; no authored generator input changed.
- **Deviation and rollback:** No deviation and no owner contradiction. Rollback
  is this checkpoint's tracker diff only.
- **Next action:** Implement PR-01, reconcile its source/test accounting, then
  update this tracker and pass Markdown lint before PR-02.

### PR-01 - Proven dead and duplicate code removal

- **Areas:** Timeline model, facade, editor/rendering support, focused tests,
  ownership/catalog inputs when selectors or paths change, and tracker.
- **Remediation:** Delete the second identical grid-width layout effect and
  forwarding lambdas that add no event or lifecycle behavior. Replace both
  private `ensureDraftRowWithFreshIndex` implementations and the test-only
  `ensureDraftRow` export with one production-used pure model command returning
  the rows plus an optional draft focus key. Delete the test-only
  `timelineFocusFieldForFieldKey` function and identity `normalizeValue`
  function. Derive Indicator observation source fields from the canonical
  Timeline bindings and immutable contract labels. Import the canonical
  `TimelineScalarEditorSurface` type in all components. Consolidate identical
  Timeline body, input, and button style primitives without changing tokens or
  rendered structure. Delete the unused callback-ref return surface from
  `useTimelineSaveStatePresentation`.
- **Rationale and benefit:** One owner remains for draft insertion, field
  identity, labels, editor surfaces, and reusable style primitives. Future
  contract additions cannot drift across copied lists.
- **Compatibility and migration:** Private TypeScript imports and tests change;
  public, persisted, wire, selector, and visual-token contracts do not. The old
  helper exports are deleted in the same slice with no alias.
- **Risk if unresolved:** Duplicate invariants and test-only APIs continue to
  create false compatibility obligations and divergent future behavior.
- **Exit criteria:** Removed symbols and hard-coded projections have no
  references; draft allocation happens at most once and preserves the expected
  focus key; contract-derived source fields retain order and labels; exactly one
  width-measurement effect remains; focused model and UI evidence passes.

#### PR-01 implementation checkpoint

- **Status and dependency:** COMPLETE on the authorized PR-00 baseline; PR-02
  is unblocked.
- **Files changed:** App mock cleanup in `App.auth.test.tsx`,
  `App.auth.support.test.tsx`, and `App.landing.test.tsx`; autosave
  characterization cleanup; Timeline cell, evidence, history, mentions, row
  action, workbook, inspector, notices, renderer, and shared-style components;
  Timeline row loader, save-state presentation, row model, Timeline model, and
  mutation coordinator source/tests; `tools/test_families/web.workbook.json`;
  Make-generated `tools/execution_topology_render_index.json`; and this tracker.
- **Substantive edits:** `ensureTimelineDraftRow` is the single production draft
  allocator and returns rows plus an optional focus key. The test-only draft and
  focus exports, identity normalizer, duplicate width effect, copied Indicator
  source labels, local editor-surface type, unused save-state callback refs, and
  byte-identical component style declarations were removed without aliases.
- **Deletion proof:** Removed exports had no production caller and supplied no
  security, recovery, accessibility, or interoperability property. One active
  width observer, contract-derived labels, draft continuity, autosave, and
  focus behavior remain.
- **Ownership and selectors:** All changed sources remain owned by
  `web.workbook`. New selector row
  `web.workbook.regression.timeline_rows_model_allocates_exactly_one_draft_8a23914036`
  owns the exact draft-allocation title. Existing titles retain their prior
  owners.
- **Commands:** `make format` PASS at
  `.cartulary/test-results/20260814T145943Z-p3757964`; `make generate` PASS at
  `.cartulary/test-results/20260814T145955Z-p3761488`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T150013Z-p3764857`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T150013Z-p3764883`;
  `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T150013Z-p3764666`;
  `make test-catalog-check` PASS;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T150029Z-p3766317`;
  `make test-slice OWNER=web.workbook` PASS, 128/128 units, at
  `.cartulary/test-results/20260814T150128Z-p3785024`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T150029Z-p3766237`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T150029Z-p3766263`; and
  `make browser-e2e-measurement` PASS, 27/27 units, at
  `.cartulary/test-results/20260814T150215Z-p3798980`.
- **Failure disposition:** The first post-catalog workbook slice exposed only a
  missing `vi` import in the new focused test. Adding the explicit import fixed
  the change-related test failure; the rerun passed all 128 units.
- **Generation impact:** Authored test-family input changed, so Make regenerated
  the execution-topology render index. Drift and protected-artifact policy pass.
- **Deviation and rollback:** No behavior or owner deviation. Rollback is the
  full PR-01 source, test, catalog, generated-index, and tracker unit.
- **Next action:** Implement PR-02's concrete runtime graph, then checkpoint it
  before PR-03.

### PR-02 - Explicit query, replay, and mutation runtime graph

- **Areas:** Timeline row loading, pending-save/replay controllers, mutation
  coordination and runtime bindings, workbook pending-runtime types, focused
  tests, ownership/catalog inputs, and tracker.
- **Remediation:** Make `useTimelineRowsLoader` own its bounded recursive retry
  and return `loadRows` directly. Remove the facade-owned no-op `loadRowsRef`.
  Move committed-record idle waiting into a focused hook that consumes the real
  load command and current pending/conflict state. Make the pending-replay
  controller register its real drainer with `WorkbookMutationRuntime`. After a
  refresh block settles, call `mutationRuntime.requestDrain()` rather than a
  scheduled-callback ref. Add a Timeline runtime-binding hook that registers
  surface refresh, accepted conflict resolution, focus restoration, and blocked
  edit discard only after all real callbacks exist. Remove
  `discardBlockedEditRef`, `schedulePendingReplayRuntimeRef`, and
  `schedulePendingReplayRef`. Make `useTimelinePendingSaves` expose one stable
  ref bundle and pass it directly instead of mirroring it through
  `pendingSavesRefsRef`.
- **Private interfaces:** `useTimelineRowsLoader` returns a semantic
  `loadRows(options)` command; the committed-record-idle hook returns
  `waitForCommittedRecordIdle(recordId, options)`; the runtime-binding hook
  receives concrete load, apply, discard, editor, and focus commands and owns
  registration cleanup. None is exported from a package.
- **Rationale and benefit:** The dependency graph becomes explicit and
  lifecycle-safe. Calling a command can no longer silently hit a placeholder
  because a later hook has not assigned its ref during render.
- **Compatibility and migration:** Private runtime interfaces only. Queue
  ordering, retry limits, row-version admission, refresh behavior, draft
  preservation, conflict outcomes, save-state presentation, and focus recovery
  remain unchanged.
- **Risk if unresolved:** Correctness continues to depend on hook order and
  render-time assignment, making teardown, concurrent rendering, and future
  replay expansion fragile.
- **Exit criteria:** The facade has no placeholder callback initialization or
  callback-assignment bridge. Runtime drainer and surface registrations use
  concrete commands and unregister on lifecycle change. Tests cover bounded
  stale retry, refresh-if-missing once, refresh unblock and drain, conflict
  resolution, blocked-edit discard, unmount cleanup, and replay ordering.

#### PR-02 implementation checkpoint

- **Status and dependency:** COMPLETE at immutable source revision
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5` plus the completed PR-00 and
  PR-01 rollback units. PR-03 is unblocked.
- **Authority:**
  `user-timeline-workbook-production-readiness-remediation-2026-08-14`,
  PR-00 through PR-06 sequentially.
- **Files added:**
  `hooks/useTimelineCommittedRecordIdle.ts`,
  `hooks/useTimelineMutationRuntimeBindings.ts`, and their two focused tests at
  the Timeline package root. The tests live outside `hooks/` to satisfy the
  controller-isolation import boundary.
- **Files changed:** `TimelineWorkbook.tsx`, the Timeline clipboard, mutation,
  pending-replay, pending-save, row-loader, save-state, and mutation-coordinator
  hooks/tests; `workbookPendingReplayRuntime.ts`;
  `tools/frontend_source_ownership.json`;
  `tools/test_families/web.workbook.json`; Make-generated
  `tools/execution_topology_render_index.json`; and this tracker.
- **Substantive edits:** The stable `WorkbookPendingSavesRefs` bundle is passed
  directly. The loader performs bounded recursive retry through its concrete
  named command. Committed-record idle waiting has one focused owner. Refresh
  settlement and completed replay units request the shared mutation runtime.
  The replay controller lifecycle-registers its concrete drainer, and the new
  runtime-binding hook registers concrete refresh, conflict-apply,
  focus-restoration, and blocked-edit-discard commands with effect cleanup.
- **Deletion proof:** `loadRowsRef`, `discardBlockedEditRef`, both scheduled
  replay bridges, `replayPendingQueueRef`, the outer pending-ref mirror, and
  `schedulePendingReplayRef` have zero live references. They were private
  render-order adapters with no owner, security, recovery, accessibility, or
  interoperability value. Retry backoff, FIFO ordering, refresh blocking,
  row-version admission, conflict recovery, draft retention, focus recovery,
  and save-state copy remain active.
- **Ownership and selectors:** Both new source/test pairs have the single live
  source owner `web.workbook`. Exact selector
  `useTimelineCommittedRecordIdle refreshes a missing committed version at most
  once` is row
  `web.workbook.regression.timeline_committed_idle_refreshes_once_19e6098c0e`;
  exact selector `useTimelineMutationRuntimeBindings registers concrete
  commands and cleans up on change and unmount` is row
  `web.workbook.regression.timeline_runtime_bindings_cleanup_f184aff191`.
  Existing coordinator evidence remains row
  `module.timeline.frontend.timeline_row_mutation_coordinator_1a7e2c9b44`.
- **Focused and broad results:** The two new `web.workbook` rows PASS at
  `.cartulary/test-results/20260814T152210Z-p3848562`; the exact
  `module.timeline` coordinator row PASS, 2/2 units, at
  `.cartulary/test-results/20260814T152814Z-p3901144`; `web.workbook` PASS,
  130/130 units, at
  `.cartulary/test-results/20260814T152603Z-p3876976`; and
  `web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T152647Z-p3890978`.
- **Static results:** `make format` PASS at
  `.cartulary/test-results/20260814T152716Z-p3896553`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T152720Z-p3900018`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T152341Z-p3860726`;
  `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T152655Z-p3892770`;
  `make test-catalog-check` PASS; and `git diff --check` PASS.
- **Failure disposition:** Intermediate typecheck failures identified stale
  coordinator-test parameters and incomplete mock typing. The first format
  attempt identified catalog ordering and callback dependencies. The first
  import-boundary run required relocating focused controller tests to the
  Timeline package root. The first broad workbook run exposed a real immediate
  replay-order regression at
  `.cartulary/test-results/20260814T152347Z-p3861132`; the registered drainer
  now executes ready work immediately while retaining the one-second retry
  backoff, and the exact failed row PASS at
  `.cartulary/test-results/20260814T152554Z-p3876458` before the broad rerun.
- **Generation impact:** Authored source ownership and test-family inputs
  changed. `make generate` PASS at
  `.cartulary/test-results/20260814T152332Z-p3857847`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T152658Z-p3893160`; and
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T152706Z-p3895995`. Only the declared
  execution-topology render index changed; no protected generated source was
  hand-edited.
- **Browser and service disposition:** No transport, protocol, persistence, or
  service behavior changed, so service-backed `module.timeline` and its
  webserver-backed replay rows are inapplicable under the Section 16 stop rule.
  The owner-routed frontend replay, stale-version, conflict, discard, refresh,
  and FIFO evidence passed.
- **Deviation and rollback:** Test relocation was required by the existing
  controller-isolation policy and did not change ownership or behavior. No
  owner contradiction or public-contract deviation occurred. Rollback is the
  full PR-02 source, tests, owner/catalog inputs, generated index, and tracker
  unit.
- **Next action:** Implement PR-03's row and inspector state ownership moves,
  then checkpoint it and pass Markdown lint before PR-04.

### PR-03 - Cohesive state ownership and facade cleanup

- **Areas:** Timeline row and inspector state hooks, facade composition,
  inspector lifecycle tests, ownership/catalog inputs, and tracker.
- **Remediation:** Make `useTimelineRows` own its initial draft, row ref, and
  draft counter, and expose semantic row state, `setRows`, and
  `nextDraftIndex`. Move `inspectorMessage` into the existing Timeline inspector
  state owner and delete `useTimelineEvidenceActions`. Retain
  `useTimelineMentions` as the cohesive owner of mention-specific local state.
  Add one semantic close command shared by explicit inspector and layout-close
  requests while leaving Escape-specific selection clearing and focus recovery
  with the keyboard/lifecycle owner. Remove obsolete imports, mirrors, and
  callback adapters exposed by PR-02.
- **Rationale and benefit:** State and lifecycle rules are colocated with their
  semantic owners instead of being represented by one-field wrappers or
  facade-owned refs.
- **Compatibility and migration:** Private hook shapes only; inspector messages,
  focus behavior, mention state, selectors, and callers remain stable.
- **Risk if unresolved:** The facade remains the implicit owner of initialization
  and close sequencing, and later inspector features can introduce divergent
  reset paths.
- **Exit criteria:** `TimelineWorkbookContent` contains controller/runtime
  composition, collaboration-boundary wiring, legitimate layout state, and
  rendering only. It owns no retry/replay policy or duplicated inspector-close
  sequence. Inspector lifecycle and row-state tests pass.

#### PR-03 implementation checkpoint

- **Status and dependency:** COMPLETE at immutable revision
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5` plus the completed PR-00 through
  PR-02 rollback units. PR-04 is unblocked.
- **Files added:** `useTimelineRows.test.tsx` and
  `useTimelineInspectorLifecycle.test.tsx` at the Timeline package root.
- **Files changed:** `TimelineWorkbook.tsx`, `useTimelineRows.ts`,
  `useTimelineInspectorSelection.ts`, `apps/web/src/README.md`, source-owner and
  `web.workbook` test-family inputs, Make-generated execution-topology render
  index, and this tracker.
- **File deleted:** `hooks/useTimelineEvidenceActions.ts`. Its only state was
  `inspectorMessage`; it did not own evidence attachment, preview, download,
  security, recovery, or accessibility behavior, and it had one facade caller.
  No forwarding alias remains.
- **Substantive edits:** `useTimelineRows` now owns the initial draft, stable row
  ref, and monotonic draft counter and returns `rows`, `rowsRef`, `setRows`, and
  `nextDraftIndex`. Inspector selection/state owns feedback state. Inspector
  lifecycle returns one `closeInspector` command used by the inspector close
  button and responsive layout close request. Escape retains its separate
  selection clearing and focus restoration sequence.
- **Ownership and selectors:** Deleted source ownership for the one-field hook;
  both new tests and all surviving sources remain singly owned by
  `web.workbook`. Exact selector `useTimelineRows owns the initial draft row ref
  and monotonic draft index` is row
  `web.workbook.regression.timeline_rows_owner_initialization_ccdda85db3`;
  exact selector `useTimelineInspectorLifecycle shares one close command across
  explicit and layout requests` is row
  `web.workbook.regression.timeline_inspector_lifecycle_close_71780eefa1`.
- **Results:** The two focused rows PASS, 3/3 units, at
  `.cartulary/test-results/20260814T153331Z-p3918166`;
  `web.workbook` PASS, 132/132 units, at
  `.cartulary/test-results/20260814T153356Z-p3921467`;
  `web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T153347Z-p3919945`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T153452Z-p3938843`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T153340Z-p3918909`;
  `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T153344Z-p3919534`;
  `make test-catalog-check` PASS; and `git diff --check` PASS.
- **Failure disposition:** The first format attempt at
  `.cartulary/test-results/20260814T153212Z-p3904047` found dependency arrays
  exposed by moving stable refs into `useTimelineRows`. The dependencies were
  expressed at their true owner and `make format` PASS at
  `.cartulary/test-results/20260814T153256Z-p3911323`. No product test failed.
- **Generation impact:** Authored source/test ownership changed. `make generate`
  PASS at `.cartulary/test-results/20260814T153322Z-p3915312`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T153442Z-p3935526`; and protected generated
  artifact policy PASS at
  `.cartulary/test-results/20260814T153450Z-p3938387`.
- **Browser and service disposition:** This private state-ownership slice did
  not cross transport or service behavior. The owner-routed workbook inspector,
  Escape/focus, lifecycle invalidation, row mutation, and rendering tests are
  the applicable evidence; service-backed `module.timeline` is inapplicable.
- **Deviation and rollback:** No owner, behavior, selector, public-contract, or
  accessibility deviation. `apps/web/src/README.md` now describes the live
  PR-02/PR-03 hook ownership and removes the misleading evidence-action entry.
  Rollback is the complete PR-03 source, tests, README, owner/catalog,
  generated-index, and tracker unit.
- **Next action:** Implement PR-04's workbook-owned relationship chip and
  focused editor/renderer boundaries, then checkpoint it before PR-05.

### PR-04 - Renderer and cross-surface boundary cleanup

- **Areas:** Timeline editors/renderers, a workbook-owned relationship-chip
  component, the Entity surface's direct import, focused component/integration
  tests, ownership/catalog inputs, and tracker.
- **Remediation:** Replace the Entity surface's import from
  `timeline/components/TimelineCellEditors` with a workbook-owned relationship
  chip accepting an explicit presentation model: label, state, optional detail,
  selected state, selection command, and stable selector identity. Timeline
  mention/entity interpretation remains in Timeline models. Split
  `TimelineCellEditors.tsx` into focused scalar-editor and draft-row-action
  components, then delete the mixed file. Split
  `TimelineWorkbookRenderers.tsx` into scalar rendering, collection rendering,
  and column assembly while preserving the existing private renderer facade.
  Delete an optional prop or branch only when PR-00 proves it has no supported
  caller; do not retain forwarding wrappers.
- **Rationale and benefit:** Common UI no longer depends on a Timeline-owned
  component, and editor, collection, and column responsibilities can evolve
  independently without a universal workbook renderer.
- **Compatibility and migration:** Private source moves and imports only.
  Relationship-chip accessibility names, state semantics, selectors, Timeline
  field order, grid/editor behavior, and Entity preview behavior remain stable.
- **Risk if unresolved:** Cross-surface imports continue to invert ownership,
  while large renderer/editor files accumulate unrelated future features.
- **Exit criteria:** No non-Timeline production surface imports a Timeline UI
  component. The mixed editor file is removed. Focused tests cover relationship
  chip states and selection, scalar controlled/uncontrolled editing, read-only
  and presence behavior, collection draft/overflow/conflict presentation, and
  column order/width. Entity-dependent Timeline previews still render.

#### PR-04 implementation checkpoint

- **Status and dependency:** COMPLETE at immutable revision
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5` plus the completed PR-00 through
  PR-03 rollback units. PR-05 is unblocked.
- **Files added:** Common `WorkbookRelationshipChip.tsx`, its focused test, and
  `workbookRelationshipChip.ts`; Timeline `TimelineDraftRowActions.tsx`,
  `TimelineScalarEditor.tsx` and its test,
  `TimelineWorkbookRendererTypes.ts`, `useTimelineScalarRenderers.tsx`,
  `useTimelineCollectionRenderer.tsx`, and
  `useTimelineColumnAssembly.tsx`.
- **Files changed:** `EntityWorkbookSurface.tsx`,
  `TimelineMentionsPanel.tsx`, `TimelineWorkbook.tsx`, the private
  `TimelineWorkbookRenderers.tsx` facade, `workbookMentionChips.ts`,
  `tools/frontend_import_boundaries.json`, source-owner and `web.workbook` plus
  `module.timeline` test-family inputs, `apps/web/src/README.md`, the
  Make-generated execution-topology render index, and this tracker.
- **File deleted:** `TimelineCellEditors.tsx`. All scalar-editor, draft-action,
  relationship-model, common-chip, style, and caller ownership moved in the
  same slice. The mixed file has zero references and no forwarding alias.
- **Substantive edits:** The common chip accepts only stable selector identity,
  label, state, optional accessible detail, selected state, and optional
  selection command. Timeline mention/entity interpretation maps into that
  model in `workbookMentionChips.ts`. Entity-dependent Timeline previews and
  Timeline surfaces render the common component. The unchanged private
  `useTimelineWorkbookRenderers` facade now composes focused scalar,
  collection, and column-assembly owners.
- **Boundary:** Rule
  `frontend-workbook-common-presentation-no-timeline-presentation` prevents
  `workbook/components/**` from importing Timeline presentation components.
  No non-Timeline production component imports `timeline/components/**`.
- **Ownership and selectors:** Every added path has source owner
  `web.workbook`; deleted `TimelineCellEditors.tsx` was removed from ownership.
  Exact selector `WorkbookRelationshipChip preserves state details selectors
  and optional selection` is row
  `web.workbook.regression.workbook_relationship_chip_presentation_f96d227ab5`.
  Exact selector `TimelineScalarEditor preserves controlled draft read-only
  presence and commit behavior` is row
  `module.timeline.frontend.timeline_scalar_editor_e5035bb033` with
  collaborator `web.workbook`.
- **Focused and broad results:** Common chip PASS at
  `.cartulary/test-results/20260814T160117Z-p4071528`; scalar editor PASS at
  `.cartulary/test-results/20260814T160041Z-p4066225`; seven applicable
  `module.timeline` scalar, draft, autosave, grid identity/order, and editor
  rows PASS, 8/8 units, at
  `.cartulary/test-results/20260814T154825Z-p3957211`;
  `web.workbook` PASS, 133/133 units, at
  `.cartulary/test-results/20260814T154835Z-p3958293`; and
  `web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T155028Z-p3982737`. The broad workbook run
  includes the Entity dependent-Timeline-preview selector.
- **Static results:** `make format` PASS at
  `.cartulary/test-results/20260814T155014Z-p3978843`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T160113Z-p4070843`;
  import-boundary PASS at
  `.cartulary/test-results/20260814T160114Z-p4071193`;
  JSON shape PASS at
  `.cartulary/test-results/20260814T160046Z-p4067006`;
  test catalog PASS; and `git diff --check` PASS.
- **Browser results:** Measurement PASS, 27/27 units, at
  `.cartulary/test-results/20260814T155040Z-p3984350`; accessibility PASS,
  14/14 units, at
  `.cartulary/test-results/20260814T155727Z-p4016471`; and visual PASS, 14/14
  units, at `.cartulary/test-results/20260814T155848Z-p4041076`. No golden or
  visual artifact changed.
- **Failure disposition:** The first typecheck found only unsupported DOM
  matcher typings in new tests. The pre-generation JSON-shape run at
  `.cartulary/test-results/20260814T154722Z-p3952545` correctly reported stale
  topology inputs. The first architecture run at
  `.cartulary/test-results/20260814T154920Z-p3972640` found a raw test-ID
  consumer literal in the scalar test; using the shared selector builder fixed
  it. No production behavior, accessibility, or visual regression failed.
- **Generation impact:** Both authored test-family inputs and source ownership
  changed. `make generate` PASS at
  `.cartulary/test-results/20260814T154743Z-p3953112`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T160049Z-p4067402`; and protected generated
  artifact policy PASS at
  `.cartulary/test-results/20260814T160057Z-p4070247`.
- **Deviation and rollback:** No public, selector, accessibility, token,
  rendering, owner, or behavior deviation. Relationship presentation moved to
  the common owner without moving Timeline semantics. Rollback is the complete
  PR-04 common/Timeline source, callers, tests, import policy, README,
  owner/catalog, generated-index, and tracker unit.
- **Next action:** Perform PR-05's full structural audit, remove any remaining
  obsolete path with its last caller, and checkpoint before final accounting.

### PR-05 - Production-readiness closure

- **Areas:** Full Timeline package source audit, facade import policy, focused
  regressions, tracker, and conditional ownership/catalog cleanup.
- **Remediation:** Audit for orphan files, production exports used only by
  tests, duplicate semantic maps, no-op callback initialization, direct
  HTTP/WebSocket or grid-vendor imports, reversed cross-owner imports, stale
  adapters, and obsolete forwarding files. Delete an obsolete file in the same
  slice as its final consumer migration. Confirm the Timeline facade remains the
  only supported production entry point and viewport measurement exists once.
- **Rationale and benefit:** The iteration closes structural debt rather than
  stopping after individual moves leave aliases or dual ownership behind.
- **Compatibility and migration:** No compatibility layer or migration is
  retained. Owner-required recovery and unsupported paths remain.
- **Risk if unresolved:** New files coexist with old paths, recreating the
  coupling and false API surface the iteration was intended to remove.
- **Exit criteria:** Source and import review finds no unexplained compatibility
  path, duplicate owner, hidden initialization dependency, orphan source, dead
  production export, transport leakage, vendor leakage, or unintended facade
  caller. Focused and browser regressions pass.

#### PR-05 implementation checkpoint

- **Status and dependency:** COMPLETE at immutable revision
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5` plus the completed PR-00 through
  PR-04 rollback units. PR-06 is unblocked.
- **Files changed:** `TimelineWorkbook.tsx`; Timeline fill, collaboration,
  evidence, history, mentions, row-action, row-loading, workbook-runtime,
  history-model, row-model, viewport-continuity, feature-policy,
  relationship-chip, record-freshness, Timeline-model, and mention-port source;
  `tools/frontend_source_ownership.json`; `apps/web/src/README.md`; and this
  tracker.
- **File deleted:** `TimelineGridSurface.tsx`. Its sole caller now renders
  `TimelineWorkbookGrid` directly, and its README and source-owner entries were
  removed in the same slice. No forwarding alias remains.
- **Deletion proof:** The deleted surface forwarded every input and ref without
  adding policy, lifecycle, accessibility, recovery, or interoperability
  behavior. Same-file-only exported types and the relationship-chip state
  helper were made private; the liveness audit found no caller outside their
  defining files and no public package export.
- **Audit results:** The Timeline package contains no orphan production file,
  test-only production export, duplicate draft/field/width owner, render-time
  placeholder callback bridge, stale deleted name, reversed common-to-Timeline
  presentation import, or direct transport/grid-vendor use outside the named
  adapter/grid boundaries. Timeline viewport measurement has one observer
  effect. `WorkbookSurfacesFacade.tsx` remains the sole production importer of
  `TimelineWorkbook`; test fixtures remain test-only callers.
- **Ownership and selectors:** Every surviving Timeline source remains singly
  owned by `web.workbook`; deleted `TimelineGridSurface.tsx` was removed from
  authored ownership. No test title or semantic row changed in PR-05.
- **Results:** `make format` PASS at
  `.cartulary/test-results/20260814T160543Z-p4080004`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T160614Z-p4089822`;
  import-boundary PASS at
  `.cartulary/test-results/20260814T160614Z-p4089830`; JSON shape PASS at
  `.cartulary/test-results/20260814T160614Z-p4089587`; test catalog PASS;
  `web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T160630Z-p4091166`; `web.workbook` PASS,
  133/133 units, at `.cartulary/test-results/20260814T160630Z-p4091172`; and
  `git diff --check` PASS. PR-04's retained measurement, accessibility, and
  visual browser evidence remains applicable because this slice removed only a
  behavior-free forwarding component and changed no rendered branch or token.
- **Generation impact:** `none` in PR-05. No authored test-family input changed,
  and no generated output was edited.
- **Deviation and rollback:** No public, behavior, selector, accessibility,
  owner, or package-export deviation. Rollback is the PR-05 direct-grid caller,
  private-export narrowing, README, source-owner, and tracker unit.
- **Next action:** Perform PR-06 final owner/selector accounting, generation and
  protected-root checks, finalization, broad verification, worktree review,
  Definition-of-Done closure, and implementation handoff.

### PR-06 - Final accounting and handoff

- **Areas:** Authored source ownership, test-family and verification accounting,
  conditional generated topology, final validation, tracker, and handoff.
- **Remediation:** Reconcile every added, moved, and deleted path in
  `tools/frontend_source_ownership.json`. Give every exact test title one active
  semantic row in the correct `tools/test_families/*.json` owner. Generate test
  topology through Make only when its authored input changed. Record the final
  file list, commands, retained roots, failures, deviations, generation impact,
  rollback posture, risks, and PR Definition of Done.
- **Rationale and benefit:** The production-ready structure is reproducible and
  remains selectable through the repository's owner-based verification system.
- **Compatibility and migration:** Verification metadata and documentation only
  except for a conditional Make-generated topology projection. Product
  contracts remain unchanged.
- **Risk if unresolved:** Deleted paths remain cataloged, new paths are unowned,
  tests select no row or multiple owners, or the handoff falsely reports a
  completed package.
- **Exit criteria:** Ownership, exact selectors, owner explanations, generation
  disposition, final diffs, retained evidence, and PR-DOD-001 through
  PR-DOD-010 are complete. Final Markdown validation passes.

#### PR-06 implementation checkpoint

- **Status and dependency:** COMPLETE at immutable revision
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5` plus the independently
  reversible PR-00 through PR-05 units. The production-readiness iteration has
  no remaining implementation workstream.
- **Files added:**
  `workbook/components/WorkbookRelationshipChip.tsx` and its test;
  `workbook/models/workbookRelationshipChip.ts`; Timeline
  `TimelineDraftRowActions.tsx`, `TimelineScalarEditor.tsx` and its test,
  `TimelineWorkbookRendererTypes.ts`, `useTimelineScalarRenderers.tsx`,
  `useTimelineCollectionRenderer.tsx`, `useTimelineColumnAssembly.tsx`,
  `useTimelineCommittedRecordIdle.ts`, `useTimelineMutationRuntimeBindings.ts`,
  and the four focused row, inspector-lifecycle, idle, and runtime-binding test
  files.
- **Files deleted:** `TimelineCellEditors.tsx`, `TimelineGridSurface.tsx`, and
  `useTimelineEvidenceActions.ts`. Their last callers, source ownership, source
  documentation, tests, and generated selector projection were migrated in the
  originating slices; none has a live reference or forwarding alias.
- **Files changed:** `apps/web/src/README.md`; three App test mocks and the
  workbook autosave test; `EntityWorkbookSurface.tsx`;
  `workbookPendingReplayRuntime.ts`; Timeline fill and collaboration owners;
  Timeline evidence, history, mentions, row-action, workbook, inspector,
  notices, renderer-facade, and style components; Timeline clipboard,
  inspector-selection, mutation-command, replay, pending-save, row,
  row-loader, save-state, and workbook-runtime hooks; Timeline history, row,
  viewport-continuity, feature-policy, relationship-chip, record-freshness,
  and workbook models plus focused model tests; the row-mutation coordinator
  and its test; `TimelineMentionPort.ts`; authored frontend import and source
  ownership policy; `module.timeline` and `web.workbook` test-family inputs;
  the Make-generated execution-topology render index; and this tracker.
- **Ownership and selectors:** The source-ownership policy accounts for every
  live TypeScript path exactly once. The final authored catalogs give each new
  exact title one active row: draft allocation and row ownership, committed
  idle refresh, runtime-binding cleanup, inspector lifecycle closure, common
  relationship-chip presentation, and Timeline scalar-editor behavior. Owner
  explanations report 132 `web.workbook`, 11 `web.architecture`, 65
  `module.timeline`, and 89 `module.workbook` semantic rows. Test-catalog and
  architecture policy checks report no unknown, duplicate, zero-row, or
  cross-owner selection.
- **Public and behavioral comparison:** `TimelineWorkbookProps` and
  `TimelineWorkbook` retain their declarations and production-facade caller;
  package entry points and generated protocol/UI contracts have no diff.
  Retained focused, workbook, architecture, measurement, accessibility, visual,
  and broad-check evidence preserves routes, wires, selectors, identities,
  authorization, conflict/replay, focus, Entity-preview, and accessible chip
  behavior. No visual golden changed.
- **Generation:** `make generate` PASS at
  `.cartulary/test-results/20260814T160830Z-p4109291`; generation drift PASS at
  `.cartulary/test-results/20260814T160839Z-p4112270`; generated-artifact policy
  PASS at `.cartulary/test-results/20260814T160839Z-p4112296`; and JSON shape
  PASS at `.cartulary/test-results/20260814T160839Z-p4112313`. The only changed
  generated projection is the Make-owned execution-topology render index.
  Protected `internal/gen`, protocol TypeScript, and UI-contract generated
  roots have zero changed paths.
- **Final validation:** All four applicable owner explanations PASS;
  `make test-catalog-check` PASS; `make agent-finalize` PASS at
  `.cartulary/test-results/20260814T160850Z-p4116223`; and `make check` PASS,
  761/761 units, at
  `.cartulary/test-results/20260814T160908Z-p4119063`; and final checkpoint
  Markdown lint PASS at
  `.cartulary/test-results/20260814T161836Z-p91552`. Final source review,
  `git diff --check`, authored JSON parsing, deleted-name search, placeholder
  callback search, facade-caller inventory, and protected-root inspection all
  PASS. Retained-run maintenance was skipped because `RESULTS_DIR` was unset;
  finalization itself completed normally before the broad check.
- **Browser and service disposition:** PR-04 retained measurement 27/27,
  accessibility 14/14, and visual 14/14 evidence with no golden change. The
  iteration did not cross service, transport, protocol, backend, or persistence
  behavior, so service-backed `module.timeline` and `module.workbook` runs are
  inapplicable under the controlling stop conditions.
- **Deviation, risks, and rollback:** No specification, public contract,
  dependency, selector, accessibility, authorization, behavior, or migration
  deviation occurred; Core 00 through Core 04 and `docs/domain.md` remain
  unchanged. No known production-readiness gap from this iteration remains.
  Rollback remains PR-slice atomic; the original staged tracker update remains
  preserved separately from the implementation worktree changes.
- **Next action:** Review and commit the complete PR-00 through PR-06 worktree as
  intentional rollback units. No compatibility cleanup or follow-on migration
  is required.

## 15. Sequencing and Mandatory Tracker Checkpoints

PR-00 through PR-06 MUST execute serially. After completing each slice and
before beginning the next, update this tracker with:

- the live source revision, authorizing task, completed status, and dependency
  state;
- exact files added, changed, moved, and deleted;
- deletion-proof dispositions for every removed behavior or symbol;
- source owner, semantic test row, and exact selector for each changed test or
  source path;
- commands, terminal results, failures, and retained result roots;
- generation impact, deviations, rollback posture, and next action; and
- updated PR risk and Definition-of-Done dispositions.

Run `make lint-markdown` after every checkpoint. The next slice MUST NOT begin
until the tracker reflects the live worktree and Markdown validation passes.
Semantic ownership or selector debt MUST be repaired in the originating slice;
PR-06 MUST NOT hide it as final mechanical accounting.

Each slice is rolled back as one unit with its source moves, callers, tests,
ownership entries, test-family entries, conditional generated topology, and
tracker checkpoint. A later failure does not require rolling back an earlier
slice that remains independently complete.

## 16. Validation and Stop Conditions

### Future implementation validation

Every authorized slice runs the applicable focused selectors plus:

- `make frontend-typecheck`
- `make frontend-import-boundary-check`
- `make test-catalog-check`
- `make json-shape-check`
- `make test-slice OWNER=web.architecture`
- `make test-slice OWNER=web.workbook`
- `make test-slice OWNER=module.workbook` when its independently owned
  postconditions are exercised
- relevant webserver-backed Timeline, keyboard, inspector, conflict, and
  measurement browser evidence
- `make lint-markdown` after its tracker checkpoint

`module.timeline` service-backed validation is not a default because this plan
does not change transport, replay protocol, persistence, or service behavior.
If a slice requires one of those changes, the slice stops and planning reopens.

PR-06 additionally runs all applicable `make explain-test-owner` commands,
conditional Make generation and drift checks, `make agent-finalize`,
`make check`, `git diff --check`, and final worktree/generated-root inspection.
A skipped conditional target requires an explicit inapplicability reason.

### Current documentation-step validation

This task changes this tracker only and runs:

- `make lint-markdown`
- `git diff --check`
- `git status --short`

Product tests, generation, source ownership, test-family metadata, and
production source are intentionally unchanged in this documentation step.

### Stop conditions

The affected PR slice MUST stop and reopen planning if implementation discovers:

- an adopted-owner contradiction or uncharacterizable required behavior;
- a required public, route, wire, selector, backend, protocol, persistence,
  projection, authorization, dependency, deployment, or package-export change;
- a need to preserve both old and new internal paths;
- a need for transport or service-backed behavior changes; or
- a deletion whose continuing product, recovery, security, or accessibility
  value cannot be disproved.

## 17. Production-Readiness Definition of Done and Handoff

| ID | Acceptance criterion | Required evidence | Current disposition |
| --- | --- | --- | --- |
| PR-DOD-001 | A later task records the exact authorized PR range, live starting revision, artifact bounds, and PR-00 prerequisite before implementation writes. | Completed Section 13 authority record | PASS in PR-00 |
| PR-DOD-002 | Every deletion has a complete owner/caller/value disposition and required behavior has pre-change evidence. | PR-00 inventory and selector map | PASS in PR-00 |
| PR-DOD-003 | Duplicate width measurement, draft insertion, identity normalization, editor-surface typing, contract labels, and style primitives have one surviving owner. | PR-01 diff and focused tests | PASS in PR-01 |
| PR-DOD-004 | The facade contains no placeholder load, replay, drain, or discard callback and no outer ref mirror. | PR-02 source review and runtime tests | PASS in PR-02 |
| PR-DOD-005 | Row initialization, inspector feedback, and inspector close sequencing have cohesive semantic owners with no obsolete one-field wrapper. | PR-03 source review and lifecycle tests | PASS in PR-03 |
| PR-DOD-006 | No non-Timeline surface imports a Timeline UI component; scalar, collection, column, and shared relationship-chip responsibilities have explicit boundaries. | PR-04 import review and component/integration tests | PASS in PR-04 |
| PR-DOD-007 | `TimelineWorkbookProps`, `TimelineWorkbook`, callers, selectors, routes, wires, identities, authorization, accessibility, conflict/replay semantics, and package exports are unchanged. | Public-contract comparison, import checks, focused and browser evidence | PASS in PR-06 |
| PR-DOD-008 | Every final source path and exact test selector has one correct live owner/row, and no deleted path remains in authored accounting. | Catalog checks, owner explanations, final path inventory | PASS in PR-06 |
| PR-DOD-009 | Generation impact is `none` or every changed authored input is regenerated through Make with clean drift and protected generated roots untouched. | Generation disposition and applicable command artifacts | PASS in PR-06 |
| PR-DOD-010 | Applicable focused, browser, finalization, broad-check, diff, worktree, and Markdown gates pass with failures and retained roots recorded. | PR-06 final handoff | PASS in PR-06 |

### Historical planning-session handoff

- **Planning source revision:**
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5`.
- **Files changed by this task:** This tracker only.
- **Specification disposition:** Core 00 through Core 04 and
  `docs/domain.md` remain unchanged; no owner contradiction was found.
- **Generation impact:** `none`.
- **Documentation validation:** `make lint-markdown` passed at retained root
  `.cartulary/test-results/20260814T143807Z-p3729334`;
  `git diff --check` passed; `git status --short` reported only this tracker as
  modified before the validation record was added.
- **Product validation:** Not run because this is a documentation-only task and
  repository policy prohibits product checks from depending on Markdown.
- **Historical blocker:** PR-00 through PR-06 originally lacked a later
  implementation task adopting their exact range and live source identity.
  Section 13 records the subsequent authorization that resolved this gate.
- **Historical next action:** The planned bounded implementation sequence was
  completed in the checkpoints above.

### Production-readiness implementation handoff

- **Authority and source:**
  `user-timeline-workbook-production-readiness-remediation-2026-08-14`, PR-00
  through PR-06, from immutable source
  `98082ac04c2a4e8a03df3a0982e30a7de12680f5`.
- **Outcome:** The Timeline workbook now has one draft allocator, one width
  measurement owner, explicit query/replay/mutation commands, cohesive row and
  inspector lifecycle owners, a common relationship presentation boundary,
  focused scalar/collection/column renderer owners, and no obsolete forwarding
  path or mutable placeholder bridge.
- **Specifications and migration:** No Core/domain specification changed, no
  owner contradiction was found, and no public, backend, protocol, persistence,
  dependency, package-export, data, or compatibility migration is required.
- **Verification:** PR-00 through PR-06 checkpoints contain exact focused and
  browser roots. Final generation, finalization, broad-check, worktree, diff,
  ownership, selector, and protected-root evidence is recorded in PR-06.
- **Worktree posture:** The pre-existing staged tracker update remains staged
  and preserved; implementation and checkpoint additions remain reviewable in
  the working tree. No commit, push, or external state change was requested or
  performed.
- **Residual risk:** No known gap remains within the authorized iteration.
  Future work should extend the explicit private commands and common
  presentation model directly rather than recreating callback mirrors,
  forwarding facades, or Timeline-to-common presentation back-edges.

## 18. Controlling Composition-Decomposition Iteration

DX-00 through DX-07 define the next planned Timeline Workbook refactoring
iteration. Sections 1 through 17 remain historical evidence for the completed S
and PR iterations. They do not authorize this new iteration and MUST NOT be
rewritten to describe future DX work as completed history.

The live planning baseline leaves `TimelineWorkbook.tsx` as a 1,568-line
composition root with 70 import declarations, 16 `useCallback` calls, 12
`useMemo` calls, eight `useRef` references, two effects, and two layout effects.
Its remaining size is predominantly orchestration: semantic adapter creation,
query and row state, mutation admission, pending replay, collaboration,
continuity, inspector workflows, editing interactions, presentation derivation,
and shell rendering still meet in one component.

This is an implementation-structure concern. Core 00 through Core 04,
`docs/domain.md`, and `docs/research/nlspec-spec.md` expose no behavioral,
vocabulary, or ownership contradiction requiring specification repair.
Components, hooks, composition folders, and package paths remain
implementation-support details. Domain vocabulary is unchanged.

### Planning and future implementation authority

| Authority field | Required value | Current value |
| --- | --- | --- |
| Planning task ID | Stable identifier for this documentation update | `user-timeline-workbook-composition-decomposition-plan-2026-08-14` |
| Planning baseline | Exact clean source revision inspected for the plan | `09604d144662fca71583ebb4eafdd71d6d671521` |
| Baseline worktree | Tracked and untracked state before this documentation write | Clean |
| Current-step authorization | Exact artifacts allowed to change now | The implementation artifacts in the next row, adopted for DX-00 through DX-07 |
| Implementation authorization | Later user task adopting an exact DX range and live source identity | `GRANTED` by `user-timeline-workbook-composition-remediation-2026-08-14` for DX-00 through DX-07 sequentially |
| Implementation starting identity | Immutable live revision and pre-existing worktree state before the first implementation write | `09604d144662fca71583ebb4eafdd71d6d671521`; the only pre-existing change was the staged Sections 18 through 22 planning update to this tracker |
| Authorized artifacts | Timeline composition, component, hook, model, and focused-test sources; `apps/web/src/README.md`; frontend import/source-ownership policy; authored test-family inputs; Make-generated test topology when an authored input changes; and this tracker | Adopted for DX-00 through DX-07 |
| Prohibited artifacts | Public API, selector, route, protocol, persistence, backend, dependency, package-export, Core/domain/NLSpec, unrelated workbook, compatibility-wrapper, and hand-edited generated changes | Binding default; any required exception stops the affected slice |
| Current public contract | `TimelineWorkbookProps`, `TimelineWorkbook`, `TimelineWorkbookSurfaceRuntime`, the private `useTimelineWorkbookRenderers` facade, the sole production-facade caller, and all observable behavior | Frozen for DX-00 through DX-07 |

The implementation task recorded above is the sole authority for DX-00 through
DX-07. The immutable starting identity and pre-existing staged tracker change
were recorded before any production, test, ownership, catalog, README, or
generated-artifact write.

### Current ownership and caller posture

- `WorkbookSurfacesFacade.tsx` remains the sole production importer of
  `TimelineWorkbook`; `TimelineWorkbookRuntimeFixture.tsx` is retained test
  composition.
- The Timeline package remains owned by `web.workbook`, with independent
  `module.timeline`, `module.workbook`, and `web.architecture` semantic evidence
  where the applicable postcondition belongs to those owners.
- Protocol translation remains in Timeline adapters. Composition and
  presentation MUST consume semantic ports and MUST NOT import generated
  protocol envelopes or browser services.
- The completed PR iteration already removed callback mirrors, forwarding
  surfaces, duplicated draft allocation, and reversed common-presentation
  imports. DX work MUST extend those boundaries, not recreate them under new
  names.

### Behavior and compatibility freeze

DX work MUST preserve routes, envelopes, `view_schema_id`, `sheet_ref`,
`record_id`, `row_version`, field keys, authorization outcomes, selector
identities, accessibility semantics, draft continuity, row-version admission,
stale retry bounds, FIFO pending replay, conflict recovery, focus restoration,
presence publication, inspector reset and Escape priority, field order and
widths, clipboard and fill behavior, bulk action behavior, Entity-dependent
Timeline previews, package exports, and persisted data.

No compatibility alias, dual composition path, React Context registry, service
locator, feature flag, dependency, generalized workbook abstraction, or data
migration is planned. A private symbol moves with all callers in its originating
slice; the superseded inline path is removed in that same rollback unit.

## 19. Target Architecture and Private Contracts

The final dependency direction is:

`TimelineWorkbook -> root composition -> presentation model -> stateless view`

Lower-level Timeline adapters, models, ports, mutation owners, collaboration
owners, editing owners, bulk owners, and leaf hooks remain below the new private
composition layer. Presentation remains above composition and receives only
grouped render-ready state and semantic commands.

### Composition owners

| Owner | Required responsibility | Required output boundary | Prohibited responsibility |
| --- | --- | --- | --- |
| `useTimelineSurfaceFoundation` | Construct semantic Timeline adapters and own query lifecycle, row, mention, pending-save, and editor-draft-registry foundations. | Grouped `snapshot`, `commands`, `ports`, and `refs` for base state only. | Grid DOM behavior, mutation/replay policy, inspector workflows, JSX, or transport envelopes. |
| `useTimelineGridEnvironment` | Own grid/shell refs, width observation, visible-column registration, focus anchors, grid continuity, viewport continuity, and the row-mutation editor adapter. | Measured width, semantic refs, continuity ports, and focus/anchor commands. | Query, replay, inspector feature, mutation outcome, or presentation policy. |
| `useTimelineInspectorStateComposition` | Own selected-row derivation, inspector feedback, open/close continuity capture, invalidation, and row-history state. | Inspector state snapshot plus semantic selection/open/history commands. | Network actions, feature execution, renderer construction, or shell JSX. |
| `useTimelineMutationComposition` | Compose row admission, loading, committed-idle waiting, collaboration, presence, pending replay, runtime registration, and scalar/collection/action mutation commands. | Nested query, save, conflict, replay, collaboration, and presence capabilities. | Inspector-specific presentation, grid JSX, or mutable callback bridges. |
| `useTimelineInspectorWorkflowComposition` | Compose related-record workflows, feature routing, member options, history actions, mention actions, evidence attachment, row menus, close, and Escape behavior. | Inspector workflow snapshot and semantic commands suitable for presentation. | Scalar/collection renderer construction or transport calls. |
| `useTimelineInteractionComposition` | Compose bulk tagging, keyboard commands, clipboard paste, fill, draft creation, scalar/collection commits, and accessible conflict/draft focus commands. | Grid/editor/bulk snapshots and semantic interaction commands. | Shell JSX, query admission, replay registration, or inspector state ownership. |
| `useTimelineWorkbookComposition` | Invoke the six composition owners above in the declared order and expose one grouped Timeline composition result. | Foundation, grid, inspector, mutation, workflow, and interaction groups without flattening them into a service locator. | Domain behavior, JSX, DOM queries, transport, duplicate state, or compatibility forwarding. |

The root invocation order is fixed:

1. surface foundation;
2. grid environment;
3. inspector state;
4. mutation/query/collaboration composition;
5. inspector workflows; and
6. editing and grid interactions.

This ordering reflects real data dependencies. It MUST NOT be implemented with
conditional hook calls, render-time registration, callback placeholders, or
cross-composer mutable refs.

### Presentation owners

| Owner | Required responsibility | Prohibited responsibility |
| --- | --- | --- |
| `useTimelineWorkbookPresentation` | Convert the root composition result into grouped grid, inspector, status, view-bar, overlay, and work-area models; invoke the retained renderer facade and inspector-section factories; derive visible columns, grid rows, load/empty state, and actionable conflict state. | Mutation execution policy, transport, lifecycle registration, raw query admission, or ownership of authoritative state. |
| `TimelineWorkbookView` | Render `WorkbookSurfaceLayout` from the grouped presentation model. | State, refs, effects, adapter construction, domain decisions, or direct composition-hook calls. |
| Timeline inspector region | Render the Timeline inspector and Indicator supplement from inspector presentation state. | Feature routing, history execution, or load policy. |
| Timeline view bar and bulk controls | Render inline/saved query controls and the selected-row bulk tag UI. | Query reduction, eligibility calculation, or mutation submission policy. |
| Timeline overlays | Render notices and the row context menu from explicit state and commands. | Pending queue, conflict, history, or row-action ownership. |

`TimelineWorkbook.tsx` retains the public prop type, keyed inner component, and
collaboration session boundary. Its inner component calls only
`useTimelineWorkbookComposition`, `useTimelineWorkbookPresentation`, and
`TimelineWorkbookView`. The final file owns no direct `useState`, `useRef`,
`useEffect`, `useLayoutEffect`, DOM-global access, adapter construction, leaf
controller call, mutation policy, feature policy, or inline shell region.

### Private contract rules

- Composition results use explicit grouped `snapshot`, `commands`, `ports`,
  and `refs` objects. They MUST NOT expose a flat bag of unrelated callbacks.
- Raw React setters remain private to their semantic owner unless an existing
  shell-owned query or layout contract requires the setter shape.
- Shared composition types live in a neutral Timeline model/port source and
  MUST NOT import hooks, components, or transport DTOs.
- The root composer is the only composition source allowed to import sibling
  composition sources. Other composers communicate through explicit inputs and
  neutral private contracts.
- Leaf hooks remain independently testable and MUST NOT import the composition
  layer. Composition coordinates leaf owners; it does not replace them.
- Presentation components MUST NOT invoke composition hooks directly or receive
  mutation-runtime, pending-runtime, transport, or collaboration-coordinator
  objects.
- Extraction is rejected when it merely moves the current body into one opaque
  mega-hook or creates an equally broad view-model file without distinct
  reasons to change.

### Planned import-boundary enforcement

Future implementation adds authored policy that enforces all of the following:

- Timeline adapters, bulk, collaboration, editing, hooks, models, mutations,
  and ports do not import `timeline/composition/**`.
- `timeline/composition/**` does not import Timeline components, protocol
  packages, or browser-service implementations.
- Composition sources do not import sibling composition sources except for
  `useTimelineWorkbookComposition`.
- Timeline components do not import composition hooks except for
  `TimelineWorkbook.tsx`.
- Existing common-workbook-to-Timeline presentation prohibitions and Timeline
  workflow transport prohibitions remain active.

### Baseline-to-target responsibility map

| Baseline region in `TimelineWorkbook.tsx` | Current responsibility | Target owner |
| --- | --- | --- |
| Lines 139-303 | Contract selection, adapter construction, layout/query lifecycle, rows, mentions, pending saves | Surface foundation |
| Lines 304-388 and 477-631 | Grid refs, measurement, focus anchors, viewport continuity, editor adapter, anchor commands | Grid environment |
| Lines 389-476 | Selection, inspector coordination, continuity capture, history state | Inspector state composition |
| Lines 526-670 and 725-917 | Row admission, loading, collaboration, replay, runtime binding, presence, mutation commands | Mutation composition |
| Lines 671-769 and 862-987 | Related workflows, feature routing, inspector lifecycle, row actions, history, mentions, evidence | Inspector workflow composition |
| Lines 770-838 and 988-1146 | Bulk tagging, keyboard, paste, fill, draft and editor commands, accessible focus actions | Interaction composition |
| Lines 1147-1318 | Renderers, visible columns, grid rows, inspector sections, load and conflict presentation | Workbook presentation owner |
| Lines 1320-1553 | Inspector, grid, status, view bar, notices, context menu, work-area shell | Stateless view and focused regions |
| Lines 1555-1568 | Public export and collaboration session boundary | Remains in `TimelineWorkbook.tsx` |

Line ranges describe the immutable planning baseline only. Future
implementation records exact symbol/file operations rather than treating these
line numbers as durable ownership.

## 20. DX Workstreams

The planned implementation sequence is:

`DX-00 -> DX-01 -> DX-02 -> DX-03 -> DX-04 -> DX-05 -> DX-06 -> DX-07`

Each slice is independently reversible and MUST complete its source, tests,
ownership, catalog, conditional generated topology, README, and tracker
checkpoint before the next begins.

### DX-00 - Authority and characterization gate

- **Areas:** Tracker authority, live source/caller inventory, responsibility
  map, import graph, and focused pre-move evidence.
- **Remediation:** Record the implementation task, exact authorized DX range,
  immutable live revision, artifact boundaries, source snapshot, and rollback
  posture. Map every inline responsibility to one target owner and every moved
  behavior to an existing or new exact selector. Characterize reload/access
  loss, width observation and cleanup, focus/viewport continuity, inspector
  lifecycle, replay/conflict recovery, collaboration/presence, keyboard,
  clipboard, fill, bulk actions, load/empty/error state, notices, context menu,
  and Entity-dependent Timeline previews.
- **Rationale and benefit:** Hook ordering and lifecycle effects are observable
  indirectly; evidence must distinguish required sequencing from incidental
  source layout before code moves.
- **Compatibility and migration:** Tests, verification metadata, and tracker
  evidence only. No product or data migration.
- **Risk if unresolved:** A structural move can silently change cleanup,
  authorization recovery, focus priority, or replay order while appearing
  behavior-neutral.
- **Exit:** Every responsibility has one current owner, target owner, caller,
  value disposition, selector, and rollback boundary. No owner contradiction or
  uncharacterizable required behavior remains.

#### DX-00 implementation checkpoint

- **Status and dependency:** COMPLETE at immutable source revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-01 is unblocked.
- **Authority:** User task
  `user-timeline-workbook-composition-remediation-2026-08-14` authorizes
  DX-00 through DX-07 sequentially. The only pre-existing worktree change was
  the staged Sections 18 through 22 planning update to this tracker.
- **Files changed:** This tracker only. No production, test, ownership,
  catalog, README, or generated artifact changed in DX-00.
- **Responsibility and caller map:** The Section 19 baseline-to-target map is
  complete for the 1,568-line source. `WorkbookSurfacesFacade.tsx` remains the
  sole production caller and `TimelineWorkbookRuntimeFixture.tsx` remains a
  test-only composition caller. Every moved responsibility retains continuing
  value and moves rather than being deleted; the rollback boundary is the
  originating DX slice. Core/domain behavior, routes, selectors, security, and
  persisted state remain out of scope.
- **Characterization map:** Query/runtime and access-loss behavior is owned by
  the existing `web.workbook` runtime, WorkbookShell query, collaboration, and
  public-route rows. Grid continuity is owned by the three
  `timelineviewportcontinuitymodel_*` rows and the five
  `workbookshell__sentinel_grid_anchor_shell_support_*` rows. Inspector
  lifecycle and routing are owned by
  `timeline_inspector_lifecycle_close_71780eefa1`,
  `timeline_inspector_feature_controller_a38c2d6f71`, and the existing
  `module.workbook` inspector rows. Replay/conflict and runtime cleanup are
  owned by `timeline_runtime_bindings_cleanup_f184aff191`, the WorkbookShell
  collaboration/action-sequencing rows, and the pending-queue model rows.
  Presence, fill, keyboard, mention undo, bulk tag, paste, autosave, and Entity
  preview behavior retain their existing exact rows recorded in the live
  `web.workbook`, `module.timeline`, and `module.workbook` manifests. PR-04's
  retained measurement, accessibility, and visual roots remain applicable
  because production source is unchanged.
- **Ownership and private contracts:** Every live Timeline source path has the
  single owner `web.workbook`. The six target composers, root composer,
  presentation hook, and stateless view use the Section 19 grouped private
  contracts. No raw transport, generated protocol, grid vendor, credential,
  or server-authorization responsibility is admitted.
- **Commands:** `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T170122Z-p111504`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T170122Z-p111528`;
  `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T170122Z-p111290`;
  `make test-catalog-check` PASS; `make test-slice OWNER=web.architecture`
  PASS, 12/12 units, at
  `.cartulary/test-results/20260814T170133Z-p112722`; and
  `make test-slice OWNER=web.workbook` PASS, 133/133 units, at
  `.cartulary/test-results/20260814T170133Z-p112724`.
- **Generation, deviation, and rollback:** Generation impact is `none`; no
  authored generator input changed. No owner contradiction, missing behavior,
  public-contract deviation, or migration was found. Rollback is this
  checkpoint's tracker diff only.
- **Next action:** Pass the tracker Markdown checkpoint, then implement DX-01's
  foundation, neutral contracts, root composer, and import policy as one
  rollback unit.

### DX-01 - Composition layer and surface foundation

- **Areas:** New private composition layer, neutral private contracts, adapter
  construction, base query/row/mention/pending/editor state, import policy,
  focused tests, README, ownership/catalog metadata, and tracker.
- **Remediation:** Add `useTimelineSurfaceFoundation` and the root
  `useTimelineWorkbookComposition`. Move semantic adapter construction,
  `useTimelineWorkbookRuntime`, rows, mentions, pending saves, and editor-draft
  registry ownership from the component. Add root-only sibling-composition and
  lower-layer-no-composition boundary rules. Delete the migrated inline path in
  the same slice.
- **Rationale and benefit:** Later composers receive stable semantic
  foundations without constructing transports, duplicating state, or importing
  presentation.
- **Compatibility and migration:** Private imports and types only; no shim or
  package export.
- **Risk if unresolved:** Later slices either duplicate foundation state or
  create a hidden service locator around the existing component body.
- **Exit:** Foundation state has one owner; lower layers cannot import
  composition; only the root composer can import sibling composers; no adapter,
  query lifecycle, row, mention, pending-save, or editor-registry construction
  remains inline.

#### DX-01 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-02 is unblocked after this
  checkpoint passes Markdown validation.
- **Production files:** Added
  `composition/useTimelineSurfaceFoundation.ts` and
  `composition/useTimelineWorkbookComposition.ts`. Updated
  `TimelineWorkbook.tsx`, `useTimelineRows.ts`, the neutral
  `timelineControllerPorts.ts` contract, and the existing row consumers in the
  loader, conflict, replay, collaboration, and mutation coordinator/command
  owners. Updated `apps/web/src/README.md` with the private composition owners.
- **Moved and deleted responsibility:** Adapter construction, guarded timing,
  query lifecycle, Timeline rows, mentions, pending saves, and editor-draft
  registry now have one foundation owner. Their former construction path was
  deleted from `TimelineWorkbook.tsx`. Raw `setRows` no longer crosses the row
  owner; callers receive stable semantic `replaceRows` and `updateRows`
  commands plus a stable read ref. No alias, forwarding facade, duplicate
  state, feature flag, or transport dependency was introduced.
- **Private-contract and dependency review:** The root returns the named
  `foundation` group. Its output is grouped as `snapshot`, `commands`, `ports`,
  and `refs`; it accepts the narrow query and mutation identities needed to
  construct the foundation. New import rules prohibit lower Timeline layers
  from importing composition, prohibit presentation/protocol/service imports
  from composition, restrict sibling-composer imports to the root, and
  restrict composition-hook callers under components to
  `TimelineWorkbook.tsx`.
- **Tests, ownership, and catalog:** Added the exact selector
  `useTimelineSurfaceFoundation owns stable adapter row query and pending
  foundations` as row
  `web.workbook.regression.timeline_surface_foundation_owns_stable_adapter_02f2dc6f6f`.
  It proves stable adapter, row/ref, pending-ref, semantic replacement, and
  query-state behavior. Updated the affected row-consumer tests, authored
  source ownership, `web.workbook` test-family input, and the Make-generated
  execution-topology index.
- **Commands:** `make format` PASS at
  `.cartulary/test-results/20260814T171132Z-p119268`; `make generate` PASS at
  `.cartulary/test-results/20260814T171327Z-p133273`; `make generate-drift`
  PASS at `.cartulary/test-results/20260814T171336Z-p136141`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T171344Z-p138998`; the new exact row PASS at
  `.cartulary/test-results/20260814T171218Z-p125863`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T171236Z-p129751`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T171246Z-p130234`;
  `make test-catalog-check` PASS; `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T171251Z-p130902`;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T171345Z-p139412`; and
  `make test-slice OWNER=web.workbook` PASS, 134/134 units, at
  `.cartulary/test-results/20260814T171350Z-p139964`.
- **Failures and correction:** The first post-extraction typecheck failed at
  `.cartulary/test-results/20260814T170707Z-p116133` because three test fixtures
  still supplied the removed raw setter; all were converted to the semantic
  row-store contract. The first architecture slice failed at
  `.cartulary/test-results/20260814T171254Z-p131325` because the two new source
  paths were not ASCII-sorted in their ownership entry; the authored entry was
  reordered and the full owner slice passed. Neither failure exposed a product
  behavior change.
- **Compatibility, risk, and rollback:** Public props, facade caller, runtime,
  selectors, routes, wires, authorization, accessibility, and persisted data
  are unchanged. No specification, backend, package-export, dependency, or data
  migration was needed. DX-01 is one rollback unit comprising the new
  composition sources, semantic row-store conversion, tests, policies,
  ownership/catalog inputs, generated projection, README, and this checkpoint.
- **Next action:** Pass `make lint-markdown`, then begin DX-02 by moving grid
  environment and inspector state lifecycle into their two explicit owners.

### DX-02 - Grid environment and inspector state

- **Areas:** Grid/DOM environment, width measurement, focus and viewport
  continuity, inspector selection/open/history state, import ownership, focused
  lifecycle tests, README, metadata, and tracker.
- **Remediation:** Add `useTimelineGridEnvironment` and
  `useTimelineInspectorStateComposition`. Move grid, shell, anchor, focus,
  continuity, and inspector refs; the width observer and fallback cleanup;
  visible-column registration; the row-mutation editor adapter; selected-row
  derivation; inspector continuity capture; invalidation; and row-history state.
- **Rationale and benefit:** DOM lifecycle and inspector lifecycle gain explicit
  owners and cannot be reassembled differently by future features.
- **Compatibility and migration:** Private ref and command migration. Preserve
  measurement rounding, resize/visual-viewport behavior, animation-frame
  follow-up, focus restoration, selected-row semantics, and reset keys.
- **Risk if unresolved:** Ref bundles and capture order remain implicit global
  state in the composition root, making concurrent rendering and future grid or
  inspector work brittle.
- **Exit:** `TimelineWorkbook.tsx` owns no grid/inspector ref, measurement
  effect, continuity token, selection state, inspector open-state adapter, or
  history state. Focused tests prove observer cleanup/fallback, width updates,
  column registration, continuity reset, selection, close, and focus restore.

#### DX-02 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-03 is unblocked after this
  checkpoint passes Markdown validation.
- **Production files:** Added `composition/useTimelineGridEnvironment.ts` and
  `composition/useTimelineInspectorStateComposition.ts`; updated the root
  composer, `TimelineWorkbook.tsx`, and the Timeline composition README table.
- **Moved and deleted responsibility:** The grid environment now owns the grid,
  shell, recovery, focus-anchor, visible-column, and viewport-token refs;
  rounded shell measurement; resize, visual-viewport, animation-frame, and
  observer cleanup; focus/anchor/viewport commands; and row-mutation editor
  adaptation. Inspector state composition now owns selected/draft-row
  derivation, feedback, open/close continuity capture, reset/invalidation, and
  row-history state. The corresponding refs, state, measurement effect,
  coordinator, continuity adapter, history hook, and editor-adapter
  construction were deleted from the component.
- **Private-contract review:** The root invokes foundation, grid, then inspector
  in dependency order. Grid consumes only reset identity, the editor registry,
  and the row read ref. Inspector consumes only semantic continuity, mention
  state, role, rows, reset identity, and the focus-anchor ref. DOM focus and
  measurement remain inside the grid owner and are exposed through semantic
  commands. The component temporarily performs only the planned
  presentation-side visible-column registration effect; the grid owns the sole
  registration command, and DX-06 moves that synchronization with
  presentation.
- **Tests, ownership, and catalog:** Added
  `useTimelineCompositionLifecycle.test.tsx` with exact rows
  `timeline_grid_environment_owns_rounded_measureme_98b38c69d0` and
  `timeline_inspector_state_preserves_continuity_an_39d7b2c152`. They prove
  width flooring, resize response, observer cleanup, selection/open state,
  continuity capture, close focus capture, and reset-key selection cleanup.
  Updated authored source ownership and `web.workbook` test-family input, then
  regenerated the execution-topology index.
- **Commands:** `make format` PASS at
  `.cartulary/test-results/20260814T172141Z-p159018`; `make generate` PASS at
  `.cartulary/test-results/20260814T172205Z-p163402`; both new exact rows PASS,
  3/3 units including prerequisite, at
  `.cartulary/test-results/20260814T172214Z-p166276`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T172147Z-p162555`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T172157Z-p163031`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T172225Z-p166956`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T172233Z-p169834`;
  `make test-catalog-check` PASS; `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T172235Z-p170543`;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T172238Z-p170972`;
  `make test-slice OWNER=web.workbook` PASS, 136/136 units, at
  `.cartulary/test-results/20260814T172244Z-p172460`; and
  `make browser-e2e-measurement` PASS, 27/27 units, at
  `.cartulary/test-results/20260814T172326Z-p187086`.
- **Failures and correction:** The first extraction typecheck failed at
  `.cartulary/test-results/20260814T171929Z-p156993` only for three obsolete
  component aliases left after ownership moved. Removing those aliases produced
  the passing typecheck above; no runtime or contract failure occurred.
- **Compatibility, risk, and rollback:** Width rounding, resize fallback,
  visual-viewport response, follow-up frame, continuity reset, selected-row
  semantics, inspector close, focus restoration, public contracts, selectors,
  and visual behavior are retained. No golden, specification, backend,
  protocol, dependency, or data migration changed. DX-02 is one rollback unit
  comprising the two composers, root/component wiring, direct tests,
  ownership/catalog inputs, generated projection, README, and this checkpoint.
- **Next action:** Pass `make lint-markdown`, then compose the explicit query,
  mutation, replay, collaboration, presence, and runtime-registration graph in
  DX-03.

### DX-03 - Query, mutation, replay, and collaboration graph

- **Areas:** Row mutation coordination, query loading, pending replay,
  collaboration/presence, runtime binding, semantic mutation commands, focused
  runtime tests, metadata, README, and tracker.
- **Remediation:** Add `useTimelineMutationComposition`. Move the row mutation
  coordinator, loader and reload effect, committed-record-idle command,
  collaboration binding, pending replay, runtime registration, presence
  controller, and scalar/collection/action mutation command composition. Return
  nested query, save, conflict, replay, collaboration, and presence capabilities
  rather than a flat callback bag.
- **Rationale and benefit:** The state-changing dependency graph becomes one
  lifecycle-owned unit while leaf controllers remain isolated and independently
  testable.
- **Compatibility and migration:** Private wiring only. Preserve query
  generation admission, bounded stale retry, draft retention, row-version high
  watermarks, FIFO replay, authorization recovery, presence reset, save-state
  copy, and conflict outcomes.
- **Risk if unresolved:** The public component continues to encode execution
  ordering across independently named hooks, so a future feature can reorder a
  correctness dependency accidentally.
- **Exit:** No query admission, refresh/retry, replay, conflict, save-state,
  collaboration, presence, socket-resolution, or authorization-recovery policy
  remains in the component. Mount/change/unmount tests prove concrete command
  registration and cleanup without callback placeholders.

#### DX-03 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-04 is unblocked after this
  checkpoint passes Markdown validation.
- **Production files:** Added
  `composition/useTimelineMutationComposition.ts`; updated the root composer,
  `TimelineWorkbook.tsx`, and the Timeline composition README table.
- **Moved and deleted responsibility:** Row-version admission, conflict/save
  projection, query loading and reload observation, committed-record idle
  waiting, collaboration admission/binding, pending replay, runtime command
  registration, presence reset/publication, authorization recovery, and
  scalar/collection/action mutation commands now form one explicit mutation
  composition graph. Every corresponding leaf-hook call, callback bridge,
  derived conflict-cell map, reload effect, and assembly block was deleted from
  the component.
- **Private-contract review:** The composer accepts explicit runtime ports plus
  narrow foundation row/pending/lifecycle capabilities, grid continuity/editor
  capabilities, and inspector selection commands. It does not receive the
  complete surface runtime or either preceding group. It returns nested
  collaboration, editor, identity, mutation, presence, query, replay, save,
  conflict, and registration capabilities. Leaf hooks remain independent;
  registration uses the concrete final commands in one render path, with no
  placeholder callbacks, mutable callback bridge, service locator, transport
  DTO, or presentation import.
- **Behavior and evidence disposition:** Existing exact rows continue to own
  query freshness, high-water admission, conflict recovery, pending FIFO,
  save-state copy, collaboration invalidation/teardown, presence reset, and
  runtime registration. Focused rows passed for the row coordinator at
  `.cartulary/test-results/20260814T173715Z-p230402`, collaboration at
  `.cartulary/test-results/20260814T173719Z-p230919`, presence and runtime
  cleanup at `.cartulary/test-results/20260814T173723Z-p231435`, and pending
  ordering/save presentation at
  `.cartulary/test-results/20260814T173727Z-p232037`.
- **Ownership and generation:** Added the composer to authored frontend source
  ownership and ran `make generate` PASS at
  `.cartulary/test-results/20260814T173706Z-p227528`. No test selector changed;
  the existing exact rows remained the correct semantic owners.
- **Commands:** `make format` PASS at
  `.cartulary/test-results/20260814T173652Z-p224015`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T173738Z-p232736`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T173748Z-p233221`;
  `make test-catalog-check` PASS; `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T173752Z-p233871`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T173755Z-p234278`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T173803Z-p237155`;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T173804Z-p237569`;
  `make test-slice OWNER=web.workbook` PASS, 136/136 units, at
  `.cartulary/test-results/20260814T173810Z-p239073`;
  `make test-slice OWNER=module.timeline` PASS, 82/82 units, at
  `.cartulary/test-results/20260814T173851Z-p253622`; and
  `make browser-e2e-stateful` PASS, 36/36 units, at
  `.cartulary/test-results/20260814T174517Z-p294275`.
- **Failures and correction:** The first extraction typecheck failed at
  `.cartulary/test-results/20260814T173525Z-p222193` for obsolete imports and
  aliases after the inline graph was deleted. Removing them produced the
  passing typecheck above; no behavior or contract assertion failed.
- **Compatibility, risk, and rollback:** Latest-query admission, stale retry,
  draft retention, FIFO replay, high-water row versions, conflict recovery,
  presence authorization reset, access-loss recovery, public contracts,
  selectors, routes, wires, and persisted data are unchanged. No golden,
  specification, backend, protocol, dependency, or data migration changed.
  DX-03 is one rollback unit comprising the mutation composer, root/component
  wiring, ownership/generation, README, and this checkpoint.
- **Next action:** Pass `make lint-markdown`, then move related-record,
  Indicator, history, mention, Evidence, row-menu, close, and Escape workflows
  into the cohesive DX-04 inspector workflow owner.

### DX-04 - Inspector workflows

- **Areas:** Related-record and Indicator workflows, reference options, history,
  mentions, evidence, row interactions, close/Escape behavior, inspector tests,
  README, metadata, and tracker.
- **Remediation:** Add `useTimelineInspectorWorkflowComposition`. Move
  create-related state/actions, fail-closed feature routing, incident-member
  options, history actions, mention actions, evidence attachment, row selection
  interactions and context menu, shared close behavior, and Escape priority.
  Expose presentation-neutral inspector workflow state and commands.
- **Rationale and benefit:** New inspector panels can extend one explicit
  workflow contract without importing mutation internals or adding another
  lifecycle path to the public component.
- **Compatibility and migration:** Private composition only. Preserve feature
  tuple validation, messages, selected mention/entity identity, evidence and
  history semantics, row-menu actions, inspector closure, and focus behavior.
- **Risk if unresolved:** Inspector features remain coupled through incidental
  hook order and acquire inconsistent reset, cancellation, or focus rules as
  future panels are added.
- **Exit:** Inspector behavior has one grouped state/workflow boundary; no
  presentation-to-mutation back-edge exists; feature, history, mention,
  evidence, close, Escape, context-menu, accessibility, and Entity-preview
  evidence passes.

#### DX-04 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-05 is unblocked after this
  checkpoint passes Markdown validation.
- **Production files:** Added
  `composition/useTimelineInspectorWorkflowComposition.ts`; updated the root
  composer, `TimelineWorkbook.tsx`, and the Timeline composition README table.
- **Moved and deleted responsibility:** Related-record workflow state,
  fail-closed feature routing, incident-member reference options, row
  selection/context-menu interactions, shared close lifecycle, row-history
  actions, mention resolution and undo, Evidence attachment, resolve-target
  feedback, and Escape priority now have one inspector workflow owner. All
  corresponding hook calls, feature lifecycle tuple construction, reference
  assembly, and callbacks were deleted from the component.
- **Private-contract review:** The composer consumes explicit semantic
  foundation ports/state, grid continuity/focus capabilities, inspector
  selection/lifecycle/history capabilities, mutation commands, and incident
  identities. It does not import presentation or receive the complete runtime,
  mutation runtime, collaboration coordinator, pending runtime, generated
  protocol, or transport DTO. Its output is grouped into workflow, feature,
  history, mention, Evidence, row-interaction, close, and resolve-target
  commands plus presentation-neutral snapshots.
- **Focused evidence:** Feature routing, lifecycle close/reset, mention
  auto-resolution/undo, and accessibility-name behavior passed together at
  `.cartulary/test-results/20260814T175432Z-p335330`; the exact
  `module.workbook` inspector-selection row passed at
  `.cartulary/test-results/20260814T175436Z-p336064`. Existing Evidence,
  history, row-menu, Entity-preview, and Indicator selectors remained unchanged
  and passed in the full owner/browser evidence below.
- **Ownership and generation:** Added the workflow composer to authored source
  ownership and ran `make generate` PASS at
  `.cartulary/test-results/20260814T175416Z-p332443`. No test selector changed;
  existing exact semantic rows remained authoritative.
- **Commands:** `make format` PASS at
  `.cartulary/test-results/20260814T175345Z-p328084`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T175448Z-p336659`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T175458Z-p337146`;
  `make test-catalog-check` PASS; `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T175503Z-p337780`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T175506Z-p338182`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T175514Z-p341028`;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T175515Z-p341448`;
  `make test-slice OWNER=web.workbook` PASS, 136/136 units, at
  `.cartulary/test-results/20260814T175521Z-p342930`; and
  `make browser-e2e-a11y` PASS, 14/14 units, at
  `.cartulary/test-results/20260814T175605Z-p357415`.
- **Failures and correction:** The first workflow extraction typecheck failed
  at `.cartulary/test-results/20260814T175217Z-p327062` for obsolete component
  aliases and an overly broad generic committed-record-idle return type. The
  input contract was narrowed to the Evidence/Timeline row result and obsolete
  aliases were removed. No behavior, security, accessibility, or contract
  assertion failed.
- **Compatibility, risk, and rollback:** Stable feature tuples, fail-closed
  unsupported actions, messages, selected identities, history and Evidence
  semantics, Entity previews, context-menu behavior, close/Escape priority,
  focus restoration, selectors, and accessible names are unchanged. No golden,
  specification, backend, protocol, dependency, or data migration changed.
  DX-04 is one rollback unit comprising the workflow composer, root/component
  wiring, ownership/generation, README, and this checkpoint.
- **Next action:** Pass `make lint-markdown`, then move bulk, keyboard,
  clipboard, fill, draft, commit, collection-input, and conflict/recovery focus
  behavior into the DX-05 interaction owner.

### DX-05 - Editing and grid interactions

- **Areas:** Bulk tag, keyboard, clipboard, fill, draft creation,
  scalar/collection commits, accessible focus actions, focused interaction
  tests, README, metadata, and tracker.
- **Remediation:** Add `useTimelineInteractionComposition`. Move bulk selection
  and submission, keyboard controller wiring, scalar and multi-cell paste,
  fill, blur and key commits, collection input tracking, explicit blank-draft
  creation, draft focus, and conflict/recovery focus commands. Consume only the
  semantic mutation, inspector, and grid-environment capabilities.
- **Rationale and benefit:** All user-originated grid/editor commands share one
  interaction boundary without owning mutation admission or presentation.
- **Compatibility and migration:** Private command migration. Preserve stable
  record/field targeting, quote-aware paste, fill rejection rules, current-value
  commits, read-only behavior, bulk eligibility, focus order, and selectors.
- **Risk if unresolved:** New editing modes must coordinate several unrelated
  inline callbacks and can bypass conflict, continuity, or authorization rules.
- **Exit:** No editor or grid command is assembled inline. Focused keyboard,
  paste, fill, bulk, draft, conflict-focus, and autosave evidence passes.

#### DX-05 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-06 is unblocked after this
  checkpoint passes Markdown validation.
- **Production files:** Added
  `composition/useTimelineInteractionComposition.ts`; updated the foundation,
  mutation and root composers, semantic row-mutation input, grid-anchor return
  contract, renderer facade, collection renderer, `TimelineWorkbook.tsx`, and
  the Timeline composition README table.
- **Moved and deleted responsibility:** Bulk tag eligibility/state/submission,
  keyboard ownership, scalar/tabular paste, fill, scalar blur/key commits,
  collection keyboard tracking, blank-draft creation, and conflict/recovery
  focus activation now have one interaction owner. Their hook calls and inline
  callbacks were deleted from the component. Mutation admission remains in
  DX-03; DOM focus execution remains in the grid environment.
- **Semantic collection-focus contract:** Collection focus state now lives
  beside the editor registry in the foundation and crosses boundaries only as
  `activateCollectionInput` and conditional `deactivateCollectionInput`
  commands. The mutation coordinator consumes the conditional semantic clear
  command; renderers no longer receive a raw React setter. This removes the
  final raw row/editor setter crossing without changing focus behavior.
- **Private-contract review:** The interaction composer consumes narrow
  foundation editor/pending capabilities, grid navigation/focus commands,
  inspector state commands, DX-03 mutation commands, DX-04 row/history
  commands, query state, interaction mode, and role. It returns grouped bulk,
  editor, grid, and conflict-focus models. It owns no query admission,
  collaboration, replay, inspector workflow, DOM query, JSX, transport, or
  presentation policy.
- **Focused evidence:** The expanded foundation row plus fill, keyboard, and
  quote-aware clipboard rows passed, 5/5 units including prerequisite, at
  `.cartulary/test-results/20260814T180819Z-p406314`; the exact
  `module.timeline` bulk-tag row passed at
  `.cartulary/test-results/20260814T180824Z-p407139`.
- **Ownership and generation:** Added the interaction composer to authored
  source ownership and ran `make generate` PASS at
  `.cartulary/test-results/20260814T180758Z-p403385`. No selector changed; the
  existing exact behavior rows remained authoritative.
- **Commands:** `make format` PASS at
  `.cartulary/test-results/20260814T180754Z-p399985`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T180836Z-p407724`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T180847Z-p408197`;
  `make test-catalog-check` PASS; `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T180851Z-p408831`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T180854Z-p409238`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T180902Z-p412094`;
  `make test-slice OWNER=web.architecture` PASS, 12/12 units, at
  `.cartulary/test-results/20260814T180903Z-p412508`;
  `make test-slice OWNER=web.workbook` PASS, 136/136 units, at
  `.cartulary/test-results/20260814T180909Z-p413996`; and
  `make browser-e2e-stateful` PASS, 36/36 units, at
  `.cartulary/test-results/20260814T180950Z-p428316`.
- **Failures and correction:** Intermediate gates exposed only extraction
  mechanics: a misplaced patch line caused the format failure at
  `.cartulary/test-results/20260814T180436Z-p386362`; typechecks at
  `.cartulary/test-results/20260814T180505Z-p393505`,
  `.cartulary/test-results/20260814T180607Z-p397723`, and
  `.cartulary/test-results/20260814T180632Z-p398288` identified an imprecise
  grid-commit result type, stale bulk identifier, nested keyboard result, test
  fixture rename, and an insufficiently precise focus-return contract. Each was
  corrected structurally before the passing gates. No product assertion
  failed.
- **Compatibility, risk, and rollback:** Stable record/field targeting,
  quote-aware paste, fill rejection, current-value commits, read-only behavior,
  bulk eligibility, event consumption, selectors, and focus ordering are
  unchanged. No golden, specification, backend, protocol, dependency, or data
  migration changed. DX-05 is one rollback unit comprising the interaction and
  semantic focus changes, root/component wiring, focused test update,
  ownership/generation, README, and this checkpoint.
- **Next action:** Pass `make lint-markdown`, then build the presentation model
  and stateless regions, move visible-column synchronization, and reduce the
  public root to its required DX-06 boundary.

### DX-06 - Presentation decomposition and slim public root

- **Areas:** Presentation model, grid/inspector/status/view-bar/overlay regions,
  `TimelineWorkbook.tsx`, visual and accessibility tests, README, import/source
  ownership, catalog metadata, conditional topology, and tracker.
- **Remediation:** Add `useTimelineWorkbookPresentation`, a stateless
  `TimelineWorkbookView`, and focused inspector-region, view-bar/bulk-controls,
  and overlay components. Move renderer and inspector-section invocation,
  visible columns, grid rows, row gutter, load/empty/error state, conflict status,
  notices, context menu, and `WorkbookSurfaceLayout` JSX. Delete superseded
  inline branches without forwarding files.
- **Rationale and benefit:** Presentation becomes render-only, future UI regions
  have narrow props, and the public source reveals the subsystem architecture at
  a glance.
- **Compatibility and migration:** Private component/import migration. Preserve
  rendered structure, field order, widths, test IDs, accessible names/live
  regions, visual tokens, grid/editor behavior, and Entity previews.
- **Risk if unresolved:** The large file survives as a presentation monolith or
  the same complexity moves into one unbounded view-model hook.
- **Exit:** `TimelineWorkbook.tsx` contains only the public contract,
  collaboration boundary, keyed inner component, root composition call,
  presentation call, and view render. It has no direct state/ref/effect,
  DOM-global, adapter, leaf-controller, mutation, feature-policy, or inline shell
  region. Measurement, accessibility, and visual suites pass with no golden
  change unless separately authorized.

#### DX-06 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`; DX-07 is unblocked after this
  checkpoint passes Markdown validation.
- **Production files:** Added
  `presentation/useTimelineWorkbookPresentation.tsx`,
  `presentation/TimelineWorkbookView.tsx`, and focused inspector, view-bar, and
  overlay region components. Updated the root composer, public component,
  Timeline README, import policy, and the existing workbook layout-policy
  fixture.
- **Moved and deleted responsibility:** Renderer and inspector-section
  invocation, visible-column derivation/synchronization, grid-row/gutter and
  load-state derivation, conflict/status projection, and inspector, view-bar,
  bulk, notice, and context-menu models moved into the presentation hook.
  `WorkbookSurfaceLayout` and every former inline region moved into stateless
  view components. All superseded JSX, callbacks, memo/effect calls, constants,
  and presentation imports were deleted from `TimelineWorkbook.tsx`; no
  forwarding component or alternate rendering path remains.
- **Public-root disposition:** `TimelineWorkbook.tsx` is 50 lines and retains
  only `TimelineWorkbookProps`, `TimelineWorkbook`, the keyed inner component,
  collaboration boundary, root-composition call, presentation call, narrow
  render-data projection, and view render. It contains no direct React state,
  ref, effect, memo, callback, DOM global, leaf hook, adapter construction,
  mutation policy, feature routing, or inline shell region.
- **Private-contract review:** The root composer returns all required named
  groups plus an explicit presentation projection. That projection contains
  only display snapshots, semantic UI commands, structural grid refs, and the
  editor-draft registry needed by renderer construction. The component creates
  exact entity and query-control objects rather than passing their wider
  runtime owners. Presentation therefore cannot receive pending-save runtime
  refs, adapters, mutation runtime, pending mutation port, collaboration
  coordinator, authorization-recovery callback, replay/registration ports,
  generated protocol, or browser service. Composition imports neither
  presentation nor components.
- **Cohesion review:** The presentation hook is a bounded derivation owner: it
  invokes the existing renderer and inspector-section facades, performs the
  sole visible-column layout synchronization, and returns named `grid`,
  `inspector`, `layout`, `overlays`, `status`, and `viewBar` models. Rendering is
  split among four stateless regions; mutation, query admission, collaboration,
  replay, authorization, and lifecycle policy remain in their composition
  owners. This is neither a service locator nor an opaque replacement hook.
- **Tests, ownership, and catalog:** Added
  `timelineCompositionArchitecture.test.ts` with exact rows
  `timeline_workbook_presentation_regions_remain_st_f91e14f779` and
  `timeline_workbook_public_root_remains_slim_and_i_72f70cc3fc`. They prove
  stateless regions, forbidden-capability exclusion, the sole visible-column
  synchronization, the slim root, and its sole production composition-hook
  call. Updated source ownership and the `web.workbook` authored family, then
  regenerated the execution-topology index. The existing layout-policy test
  now checks the stateless Timeline view as the concrete geometry owner.
- **Focused and static commands:** Both new rows PASS, 3/3 units including the
  prerequisite, at
  `.cartulary/test-results/20260814T183524Z-p582200`; `make format` PASS at
  `.cartulary/test-results/20260814T183507Z-p577949`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T183511Z-p581421`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T183521Z-p581859`; `make generate` PASS at
  `.cartulary/test-results/20260814T182244Z-p470996`; `make json-shape-check`
  PASS at `.cartulary/test-results/20260814T182308Z-p475004`;
  `make test-catalog-check` PASS; `make generate-drift` PASS at
  `.cartulary/test-results/20260814T182311Z-p475400`; and
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T182319Z-p478254`.
- **Owner and browser commands:** `make test-slice OWNER=web.architecture`
  PASS, 12/12 units, at
  `.cartulary/test-results/20260814T183549Z-p582948`; `make test-slice
  OWNER=web.workbook` PASS, 138/138 units, at
  `.cartulary/test-results/20260814T183554Z-p584435`;
  `make browser-e2e-measurement` PASS, 27/27 units, at
  `.cartulary/test-results/20260814T183637Z-p599401`;
  `make browser-e2e-a11y` PASS, 14/14 units, at
  `.cartulary/test-results/20260814T184318Z-p631643`; and
  `make browser-e2e-visual` PASS, 14/14 units with no golden change, at
  `.cartulary/test-results/20260814T184437Z-p656393`.
- **Failures and correction:** The first direct architecture run at
  `.cartulary/test-results/20260814T182132Z-p464974` correctly found that its
  caller scan also matched the root composer's declaration. Narrowing the scan
  to the invocation shape fixed that assertion. The next run at
  `.cartulary/test-results/20260814T182202Z-p466041` exposed an invalid unescaped
  brace in that scan's regular expression; escaping the literal braces produced
  the passing exact rows above. No product, component, browser, accessibility,
  measurement, or visual assertion failed.
- **Compatibility, generation, and rollback:** Public props/exports, the
  renderer facade, structure, field order/widths, selectors, accessibility
  semantics, live regions, visual tokens, grid behavior, Entity previews,
  routes, wires, authorization outcomes, and persisted data are unchanged. No
  golden, specification, backend, protocol, dependency, package export, or data
  migration changed. DX-06 is one rollback unit comprising the presentation
  files, root/composer wiring, direct and layout-policy tests, README,
  ownership/import/catalog inputs, generated topology, and this checkpoint.
- **Next action:** Pass `make lint-markdown`, then perform DX-07 structural
  closure, source/selector accounting, finalization, broad verification, and
  handoff.

### DX-07 - Structural closure, accounting, and handoff

- **Areas:** Full Timeline/direct-consumer audit, source ownership, test-family
  accounting, conditional generated topology, final validation, tracker, and
  handoff.
- **Remediation:** Audit for orphan files, oversized replacement owners, dead or
  test-only exports, raw setters, hidden hook-order dependencies, mutable
  callback wiring, duplicate semantic maps, transport/vendor leakage, reversed
  imports, stale README entries, forwarding files, and unintended facade
  callers. Reconcile every path and exact selector, then complete the DX
  Definition of Done.
- **Rationale and benefit:** Production readiness requires removal of the old
  composition path and proof that decomposition did not merely redistribute the
  same coupling.
- **Compatibility and migration:** Accounting and cleanup only. No
  compatibility layer or product migration.
- **Risk if unresolved:** The package gains more files without gaining clearer
  ownership, or verification/catalog drift makes the new structure difficult to
  extend safely.
- **Exit:** Every file has one reason to change and one source owner; every exact
  selector has one active semantic row; no obsolete path or unexplained
  dependency remains; generation, finalization, broad verification, diff,
  worktree, and protected-root inspections pass.

#### DX-07 implementation checkpoint

- **Status and dependency:** COMPLETE against starting revision
  `09604d144662fca71583ebb4eafdd71d6d671521`. DX-00 through DX-07 are complete;
  no further implementation slice is pending.
- **Closure cleanup:** Consolidated the duplicated context-menu position and
  committed-record-idle result contracts into
  `models/timelineControllerPorts.ts`, then changed both consumers to import the
  neutral types. Updated the stale Timeline root description in the source
  README. No compatibility alias, suppression, or forwarding export was added.
- **Canonical test identity:** Reallocated the two DX-06 architecture rows with
  `make author-test-row-id` before final handoff. Their final IDs are
  `timeline_workbook_presentation_regions_remain_st_f91e14f779` and
  `timeline_workbook_public_root_remains_slim_and_i_72f70cc3fc`; both exact
  selectors PASS, 3/3 units including the prerequisite, at
  `.cartulary/test-results/20260814T185214Z-p698637`. The earlier provisional
  IDs existed only in this uncommitted iteration and require no migration row.
- **Structural audit:** The final Timeline scope has no Fallow-reported unused
  file, export, or type and no duplication group involving composition or
  presentation. Every new source is reachable, substantive, and owner-listed.
  No raw row setter, React setter contract, mutable callback bridge, lower-layer
  composition/presentation import, protocol/vendor/service import, duplicate
  semantic contract, alternate inline leaf-hook path, forwarding file, or
  composition-to-presentation back-edge remains. The presentation owner is a
  bounded derivation hook with six named output models and four stateless view
  regions, not an opaque replacement monolith or service locator.
- **Caller and public-contract audit:** `WorkbookSurfacesFacade.tsx` remains the
  sole production importer/caller of `TimelineWorkbook`; the separate runtime
  fixture remains test-only. `TimelineWorkbook.tsx` is the sole production
  caller of the root composition hook, and the presentation hook is the sole
  production caller of the renderer facade. `TimelineWorkbookProps`,
  `TimelineWorkbook`, `TimelineWorkbookSurfaceRuntime`, the renderer facade,
  selectors, routes, wire shapes, authorization outcomes, accessibility
  semantics, package exports, and persisted data are unchanged.
- **Ownership and selector accounting:** `make explain-test-owner` resolves
  `web.architecture` to 11 static rows, `web.workbook` to 137 rows,
  `module.timeline` to 65 rows, and `module.workbook` to 89 rows, with the
  expected focused and service-backed routes. Source ownership covers every
  final path exactly once, and `make test-catalog-check` confirms every exact
  title has one active semantic row.
- **Generation and protected roots:** `make generate` PASS at
  `.cartulary/test-results/20260814T185142Z-p691710`;
  `make generate-drift` PASS at
  `.cartulary/test-results/20260814T185151Z-p694543`;
  `make generated-artifact-policy-check` PASS at
  `.cartulary/test-results/20260814T185159Z-p697437`; and
  `make json-shape-check` PASS at
  `.cartulary/test-results/20260814T185202Z-p698145`. The Make-generated
  execution-topology render index is the only generated projection changed;
  protected `internal/gen`, protocol-ts, and ui-contracts roots are untouched.
- **Focused and module validation:** `make format` PASS at
  `.cartulary/test-results/20260814T185055Z-p686560`;
  `make frontend-typecheck` PASS at
  `.cartulary/test-results/20260814T185058Z-p690033`;
  `make frontend-import-boundary-check` PASS at
  `.cartulary/test-results/20260814T185109Z-p690486`; and final
  `make frontend-fallow-static` PASS at
  `.cartulary/test-results/20260814T185117Z-p690937`. Full non-service owner
  slices PASS for `module.timeline`, 82/82 units, at
  `.cartulary/test-results/20260814T185236Z-p700178`, and `module.workbook`,
  100/100 units, at
  `.cartulary/test-results/20260814T185901Z-p740389`.
- **Finalization and broad verification:** `make agent-finalize` PASS, 1/1, at
  `.cartulary/test-results/20260814T190136Z-p781611`; retained-run maintenance
  was intentionally skipped because `RESULTS_DIR` was unset and no prior full
  warm-check root existed. The required subsequent `make check` PASS, 766/766,
  at `.cartulary/test-results/20260814T190155Z-p784467`. Final
  checkpoint `make lint-markdown` PASS at
  `.cartulary/test-results/20260814T191001Z-p956808`; `git diff --check` passes;
  tracked/untracked inspection shows only the authorized DX implementation,
  authored metadata/documentation, the Make-generated topology projection, and
  the controlling tracker. No visual golden or protected generated-root file
  changed.
- **Failures and correction:** The first DX-07 `make format` preflight rejected
  the newly canonical row IDs because they had not yet been moved to ASCII
  order. Reordering the authored family produced the passing format and catalog
  gates above. One read-only facade scan contained invalid shell quoting; a
  simplified rerun confirmed one production facade importer plus the test-only
  fixture. Neither failure changed product state or exposed a product defect.
- **Compatibility, residual risk, and rollback:** This closure slice changes
  only private type ownership, final test identity/accounting, README wording,
  and generated topology. No specification, backend, protocol, service,
  dependency, package export, selector, golden, authorization, persistence, or
  data migration changed. Service-backed owner slices were not run because no
  transport/service postcondition changed; the full non-service owners, all
  assigned browser suites, and broad check passed. No known remediation risk
  remains. DX-07 rolls back as one unit with its neutral-contract imports,
  canonical row IDs/topology, README, and this checkpoint; DX-00 through DX-06
  retain their independent rollback units.
- **Handoff:** The remediation is production-ready for review. Preserve the
  existing user-staged Sections 18 through 22 tracker change when committing;
  all implementation and checkpoint edits currently remain unstaged.

## 21. Validation, Checkpoints, and Stop Conditions

### Mandatory DX checkpoint protocol

After each authorized DX slice and before starting the next, update this
tracker with:

- implementation authority, immutable source identity, completed status, and
  dependency state;
- exact files added, changed, moved, and deleted;
- responsibility/caller/value disposition for every move or deletion;
- private input/output contract and prohibited dependency review;
- source owner, semantic row, and exact selector for each affected source/test;
- commands, terminal results, failures, and retained result roots;
- generation impact, deviations, residual risks, rollback posture, and next
  action; and
- current DX risk and Definition-of-Done dispositions.

Reconcile source ownership and test-family inputs in the originating slice. If
an authored test-family input changes, run `make generate` and the applicable
drift, protected-artifact, and JSON-shape checks; never hand-edit generated
outputs. Run `make lint-markdown` after every checkpoint. The next slice does
not begin until its predecessor is complete and Markdown-clean.

### Per-slice validation

Run the narrowest applicable focused selectors plus:

- `make frontend-typecheck`
- `make frontend-import-boundary-check`
- `make test-catalog-check`
- `make json-shape-check`
- `make test-slice OWNER=web.architecture`
- `make test-slice OWNER=web.workbook`
- exact `module.timeline` and `module.workbook` rows when their independent
  postconditions are exercised
- `make lint-markdown` after the tracker checkpoint

Additional required evidence by slice:

| Slice | Required behavioral evidence |
| --- | --- |
| DX-00 | Pre-move characterization for every responsibility in Section 19. |
| DX-01 | Foundation identity/reset behavior, architecture policy, and existing broad workbook behavior. |
| DX-02 | Width measurement/fallback/cleanup, focus and viewport continuity, inspector selection/reset/close, keyboard focus, and measurement browser rows. |
| DX-03 | Query freshness, pending replay, conflict recovery, collaboration invalidation, presence reset, authorization recovery, and public-route browser rows. |
| DX-04 | Inspector lifecycle, fail-closed feature routing, history, mention, Evidence, context menu, Entity-preview, and accessibility rows. |
| DX-05 | Keyboard, autosave, scalar/collection edit, clipboard, fill, bulk tag, conflict focus, and relevant grid/browser rows. |
| DX-06 | Grid/column/load-state, inspector, view bar, overlays, Entity preview, measurement, accessibility, and visual rows; any golden change is a regression unless separately authorized. |
| DX-07 | All applicable owner explanations, generation/drift policy, finalization, broad check, diff, worktree, generated-root, facade-caller, and orphan/dead-export audits. |

DX-07 additionally runs `make agent-finalize` before `make check`, followed by
`git diff --check` and final tracked/untracked plus protected-generated-root
inspection.

Service-backed `module.timeline` and `module.workbook` validation is not a
default because this iteration does not plan transport, service, protocol,
backend, or persistence behavior. Crossing any of those boundaries stops the
slice and reopens planning.

### Documentation-only planning-step validation

This planning step changes this tracker only and runs:

- `make lint-markdown`
- `git diff --check`
- `git status --short`

It does not run product tests, generation, source-ownership mutation, or
test-family mutation because repository verification and product behavior MUST
NOT depend on Markdown planning artifacts.

### Stop conditions

An authorized DX slice MUST stop and reopen planning if it discovers:

- an adopted-owner contradiction or required behavior that cannot be
  characterized;
- a required public API, selector, route, wire, backend, protocol, persistence,
  projection, authorization, deployment, dependency, or package-export change;
- a need for dual internal composition paths, a compatibility wrapper, or a
  feature flag;
- a need for transport or service-backed behavior changes;
- a composition contract that exposes unrelated state as a service locator;
- a presentation model that owns mutation or lifecycle policy;
- a lower-layer import of composition or a composition-to-presentation back-edge;
  or
- an extraction that only replaces `TimelineWorkbook.tsx` with an equally broad
  opaque hook or component.

## 22. Composition Definition of Done and Handoff

| ID | Acceptance criterion | Required evidence | Current disposition |
| --- | --- | --- | --- |
| DX-DOD-001 | A later implementation task records the exact authorized DX range, immutable live starting identity, artifact bounds, and DX-00 prerequisite before implementation writes. | Completed Section 18 authority record | PASS in DX-00 |
| DX-DOD-002 | Every moved responsibility has one current owner, target owner, caller/value disposition, rollback boundary, and pre-move behavior selector. | DX-00 inventory and characterization map | PASS in DX-00 |
| DX-DOD-003 | Foundation state and adapters have one composition owner without duplicate state, transport leakage, compatibility aliases, or lower-layer back-edges. | DX-01 source/import review and focused evidence | PASS in DX-01 |
| DX-DOD-004 | Grid environment and inspector state own all refs, measurement, continuity, selection, open/reset, and history lifecycle with cleanup evidence. | DX-02 focused lifecycle, measurement, focus, and inspector results | PASS in DX-02 |
| DX-DOD-005 | Query, mutation, replay, collaboration, presence, and runtime registration form one explicit command graph without placeholders or hidden render-order dependencies. | DX-03 runtime graph tests and source review | PASS in DX-03 |
| DX-DOD-006 | Inspector workflows consume semantic capabilities through one grouped boundary and preserve feature, history, mention, Evidence, row-menu, close, Escape, accessibility, and Entity-preview behavior. | DX-04 focused and browser evidence | PASS in DX-04 |
| DX-DOD-007 | Editing and grid interactions consume semantic capabilities without owning mutation admission or presentation and preserve keyboard, paste, fill, bulk, autosave, draft, and focus behavior. | DX-05 focused and browser evidence | PASS in DX-05 |
| DX-DOD-008 | `TimelineWorkbook.tsx` owns only the public/collaboration boundary and high-level composition/presentation/view calls; stateless presentation regions own existing JSX without policy. | DX-06 source review, import policy, component tests, and browser evidence | PASS in DX-06 |
| DX-DOD-009 | No extracted owner is an opaque replacement monolith, service locator, forwarding file, duplicate owner, dead export, reversed import, hidden lifecycle path, or unintended facade caller. | DX-07 structural audit | PASS in DX-07 |
| DX-DOD-010 | `TimelineWorkbookProps`, `TimelineWorkbook`, `TimelineWorkbookSurfaceRuntime`, renderer facade, public behavior, selectors, package exports, routes, wires, authorization, accessibility, and persisted data are unchanged. | Public-contract comparison and retained focused/browser evidence | PASS in DX-07 |
| DX-DOD-011 | Every final source path and exact selector has one correct owner/row, and all authored generator inputs have clean Make-generated projections and protected roots. | Catalog, owner explanations, generation/drift, and protected-root results | PASS in DX-07 |
| DX-DOD-012 | All applicable focused, browser, Markdown, finalization, broad-check, diff, worktree, and handoff gates have terminal evidence with failures and deviations recorded. | DX-07 final handoff | PASS in DX-07 |

### Composition-decomposition planning handoff

- **Planning task:**
  `user-timeline-workbook-composition-decomposition-plan-2026-08-14`.
- **Planning source revision:**
  `09604d144662fca71583ebb4eafdd71d6d671521` from a clean worktree.
- **Implementation range:** DX-00 through DX-07 completed serially with a
  tracker and Markdown checkpoint between every workstream. Sections 1 through
  17 remain historical evidence.
- **Specification disposition:** Core 00 through Core 04, `docs/domain.md`, and
  `docs/research/nlspec-spec.md` remain unchanged. No owner or vocabulary
  contradiction was found; domain vocabulary is unchanged.
- **Implementation status:** `COMPLETE`. All DX Definition-of-Done rows pass;
  no planned implementation work remains.
- **Generation and product validation:** Authored ownership, import policy, and
  test-family inputs are reconciled; generated topology is current; focused,
  owner, measurement, stateful, accessibility, visual, finalization, and broad
  verification evidence is recorded in the per-slice checkpoints above.
- **Documentation validation:** Every intermediate checkpoint passed
  `make lint-markdown`; final checkpoint lint passed at
  `.cartulary/test-results/20260814T191001Z-p956808`, and `git diff --check`
  passed.
- **Next action:** Review and commit the completed remediation as the recorded
  serial rollback units. No specification adoption, compatibility migration,
  backend rollout, protocol deployment, or data migration is required.
