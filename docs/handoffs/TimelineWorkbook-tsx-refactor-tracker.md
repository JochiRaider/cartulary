# TimelineWorkbook-tsx Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx`
- **Target label:** `TimelineWorkbook-tsx`
- **Normalized identifier:** `timeline-workbook-tsx`
- **Output path:** `docs/handoffs/TimelineWorkbook-tsx-refactor-tracker.md`
- **Status:** Planning and documentation only.
- **Allowed change:** This tracker file only.
- **Non-goals:** No production refactor, tests, contracts, generated artifacts, dependency or package configuration, migrations, harness inputs, or runtime behavior changes.
- **Implementation authorization:** Any implementation described below requires a later, explicitly authorized task.

Normative terms in this tracker have the following meanings:

- **MUST** and **MUST NOT** identify mandatory refactor requirements.
- **SHOULD** and **SHOULD NOT** identify requirements that may be waived only
  when the handoff records contrary owner evidence and the reason for the waiver.
- **MAY** identifies intentional implementation freedom.
- Unspecified internal decomposition is an implementation choice only when two
  different choices cannot affect callers, tests, ownership, or interoperability.

`docs/research/nlspec-spec.md` supplies writing doctrine for this tracker, and
`temp/analysis-notes.md` supplies review evidence. Neither document owns product
behavior. Core 00 through Core 04 and adopted subsystem owners remain the
authorities for every preserved observable behavior.

The target exists and was inspected directly in full. It is a single 2,176-line
TypeScript React file, so the inventory in Section 2 has one in-scope row. The
safe identifier `timeline-workbook-tsx` is a planning label, not an assertion
that the component is or should become a permanent module boundary.

The source hierarchy used for this tracker was:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication; it is
   not applicable to this refactor plan.
4. Domain vocabulary, design direction, and implementation-support guides.
5. Current repository code and tests as current-state evidence.
6. The planning framework and prior handoffs as evidence and doctrine only.

### Implementation authority contract

This tracker does not itself authorize implementation. One later task MAY
authorize S-00 through S-08 together, but that task MUST provide every field in
the following record. An instruction such as "refactor TimelineWorkbook" is not
sufficient authority.

| Authority field | Required value | Current value |
| --- | --- | --- |
| Authorizing task or decision ID | A stable identifier that can be cited by every implementation handoff | `TODO: authorization not issued` |
| Authorized slices | An exact contiguous or enumerated subset of S-00 through S-08 | `TODO`; intended default is S-00 through S-08 |
| Starting source identity | Exact commit or immutable source snapshot | Planning baseline `b9f81d04`; implementation MUST re-read and record the live value |
| Allowed artifacts | The target; the Timeline-owned policy/controllers/hooks named by S-01 through S-06; the existing mention controller; focused tests; authored source-ownership, verification, and test-family inputs required by new paths/selectors; generated projections produced through public Make targets; this tracker and its implementation handoff | `TODO: must be adopted by the authorizing task` |
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
| TW-001 | Establish scope, authority, clean state, and safe target identifier | WF-00 | DONE | None | Section 1; clean `main` at inspected revision `b9f81d04` | Planning boundaries and source hierarchy are explicit. |
| TW-002 | Inventory the complete target and its callers/dependencies | WF-01 | DONE | TW-001 | Section 2 | The sole target file is fully inventoried. |
| TW-003 | Map observable contracts to owners and tests | WF-02 | DONE | TW-002 | Section 4 | Every discovered contract risk has an owner and test posture. |
| TW-004 | Define the five-seam characterization contract | WF-03 | DONE | TW-002 | Sections 4, 7, and 8 | Fill, keyboard/focus, presence, mention undo, and inspector routing have explicit current-behavior cases, owners, commands, and evidence fields. |
| TW-005 | Diagnose coupling and permanent-boundary posture | WF-04 | DONE | TW-002 | Sections 3 and 5 | Legitimate composition and residual controller ownership are separated. |
| TW-006 | Define facade and residual-owner plan | WF-05 | DONE | TW-003, TW-004, TW-005 | Section 3 | The facade remains stable and no new permanent module is proposed. |
| TW-007 | Sequence behavior-preserving slices | WF-06 | DONE | TW-006 | Section 7 | Every slice has dependencies, validation, rollback, and exit criteria. |
| TW-008 | Define same-slice source/test accounting and final reconciliation | WF-07 | DONE | TW-004, TW-007 | Sections 5, 7, and 8 | Every planned path has a default owner and each slice must land exact accounting before S-08. |
| TW-009 | Resolve the bounded implementation-authority gate | WF-08 | TODO | TW-007 | RB-001 | Task ID, slice range, artifact bounds, start revision, and S-00 prerequisite are recorded. |
| TW-010 | Resolve characterization and ownership preflight gates | WF-08 | TODO | TW-009 | S-00, RB-002, RB-003 | All five seam families pass on the pre-extraction implementation with exact owner/row/selector/run-root evidence. |
| TW-011 | Implement S-01 through S-07 independently | WF-08 | TODO | TW-010 | Future commits and test evidence | Each extraction satisfies its binary completion criterion without contract change. |
| TW-012 | Complete conditional accounting and final validation | WF-08 | TODO | TW-011 | S-08 and validation run root | Required focused-to-broad commands and ownership checks pass. |
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

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Protocol, view, UI, and selector contracts are frozen consumers; no codegen change is expected. | Inspected generated-artifact policy, package imports, HTTP-operation projection, and OpenAPI owner inputs. Touched: this tracker only. | Targeted `rg`, `sed`, `jq`; batched `make explain-target` attempt | No generated hand edit is planned. Explain-target output was truncated and is not validation evidence. | RB-003 if authored inputs later change | Run `make generate-drift` only when a later slice changes an authored generator input. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | Generation has an explicit default and applicability rule; generated projections remain downstream. | Inspected view-contract feature types, UI/grid contract types, generated policy, and live ownership inputs. Touched: this tracker only. | `sed`, `rg`, `jq` | `generation impact: none` is required unless an authored generator input changes; hand edits remain prohibited. | RB-003 | If an authored input changes, run the complete Make-owned generation/drift/policy sequence and record outputs. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 20:48:55 EDT | Codex planning session | Broad current coverage exists; direct seam coverage is required before risky movement. | Inspected WorkbookShell, focused Timeline, browser, owner catalog, and test-family evidence. Touched: this tracker only. | `make help`; `make help-all`; three `make task-guide` and three `make explain-test-owner` commands; source searches | Commands discovered and owner rows explained. No product test or validation suite was run. | RB-002, RB-003 | Run S-00 focused baseline and account for every new test/source path. |
| 2026-08-13 21:25:30 EDT | Codex NLSpec revision session | S-00 now has five mandatory seam families, exact owner rules, applicability rules, and an evidence-record schema. | Inspected `web.workbook`, `module.workbook`, and `module.timeline` family manifests and representative exact selectors. Touched: this tracker only. | `sed`, `rg`, `jq`; `make lint-markdown`; no product suite | Test ownership is determined by postcondition and must land with each selector. Markdown lint passed at `.cartulary/test-results/20260814T013042Z-p410027`; no product validation success is claimed. | RB-002, RB-003 | Add/confirm tests on the pre-extraction source, then retain exact row and run-root evidence. |

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

## 11. Open Questions and Blockers

These entries are implementation-readiness gates, not unresolved product-design
questions. Their closure rules are exhaustive; an implementer MUST NOT substitute
informal approval, a broad passing command, or inferred ownership.

| ID | Resolved decision and required closure evidence | Current status | Final status rule |
| --- | --- | --- | --- |
| RB-001 | Implementation is authorized only by a task adopting an exact S-00 through S-08 range and the Section 1 allowed/prohibited scope. The tracker MUST record the task ID, slice range, starting revision, artifact bounds, and S-00-before-movement rule. | `AWAITING_AUTHORIZATION` | Set `RESOLVED` only when every authority-record field is populated by a qualifying task. |
| RB-002 | S-00 covers fill, keyboard/focus, presence, mention undo, and inspector routing. Every new characterization test MUST pass against the pre-extraction implementation, and exact owners, stable rows, selectors, terminal results, applicable browser/service evidence, and retained roots MUST be recorded. | `CHARACTERIZATION_GATE` | Set `RESOLVED` only when all five seam families and every mandatory evidence-record field pass without encoding desired new behavior. |
| RB-003 | Every new path MUST have one live-catalog-confirmed owner; every exact test selector MUST have one active semantic row; accounting MUST land with the originating slice; generated outputs MUST come only from Make and be clean or explicitly inapplicable. | `OWNERSHIP_PREFLIGHT_GATE` | Set `RESOLVED` before S-01 only when planned paths/tests have a recorded disposition, then keep it resolved by satisfying the same-slice rule through S-08. Any later accounting defect reopens the gate. |

No owner contradiction was discovered during planning, so no
`BLOCKED: owner contradiction` entry exists. If implementation discovers one,
the affected slice MUST stop and the marker MUST be added here. The universal
feature-dispatch abstraction remains intentionally out of scope; the binding
default is a Timeline-owned first extraction.

## 12. Binary Completion Criteria

The following table is the Definition of Done for the future refactor. Every row
is binary. `SPECIFIED` means this tracker supplies the requirement and evidence
shape; it does not mean the future implementation has passed the criterion.

| ID | Normative requirement source | Acceptance criterion | Required evidence | Current disposition |
| --- | --- | --- | --- | --- |
| DOD-001 | Sections 1-2 | The sole target file remains fully inventoried, and `TimelineWorkbookProps`, `TimelineWorkbook`, all production callers, routes, schemas, events, selectors, identifiers, authorization outcomes, and package exports are unchanged. | Source diff, typecheck, import-boundary result, contract drift disposition, and caller inventory | SPECIFIED; implementation pending |
| DOD-002 | Sections 1 and 11, RB-001 | A qualifying task records the exact authority ID, slice range, source revision, allowed/prohibited artifacts, and S-00 prerequisite before any implementation write. | Completed authority record and handoff reference | SPECIFIED; `AWAITING_AUTHORIZATION` |
| DOD-003 | Section 4 and S-00 | Fill, keyboard/focus, presence, mention undo, and inspector routing each have isolated current-behavior evidence that passes before the corresponding production move. | Exact tests, stable rows, terminal results, applicable browser/service results, and retained roots | SPECIFIED; `CHARACTERIZATION_GATE` |
| DOD-004 | Section 3 | Every extracted responsibility satisfies its closed private input/output contract, default, unsupported result, and prohibited-coupling rules without a new public export. | Focused interface tests, typecheck, import-boundary result, and source review | SPECIFIED; implementation pending |
| DOD-005 | Section 5 and S-01 through S-08 | Every new source path has exactly one live owner, every exact selector has exactly one active semantic row, and accounting lands with the originating slice. | Source-owner catalog result, three applicable owner explanations, selector disposition, and no unknown/duplicate/cross-owner/zero-row result | SPECIFIED; `OWNERSHIP_PREFLIGHT_GATE` |
| DOD-006 | Sections 1, 4, and 5 | Server authorization, fail-closed unsupported-feature handling, stable semantic identities, transport/persistence/grid boundaries, and mention/entity separation remain unchanged. | Characterization rows, security/recovery evidence, import review, and public-contract comparison | SPECIFIED; implementation pending |
| DOD-007 | Section 7 | S-01 through S-07 are independently reversible and each satisfies its completion criterion before the next dependent slice begins. | Slice-scoped diffs, results, rollback note, ownership/accounting checkpoint, and handoff entry | SPECIFIED; implementation pending |
| DOD-008 | Section 8 | Generation is either explicitly `none` or all changed authored inputs are regenerated through Make with clean drift, artifact-policy, and JSON-shape results; no generated file is hand-edited. | Generation-impact record and applicable command artifacts | SPECIFIED; implementation pending |
| DOD-009 | Section 8 | All mandatory baseline, per-slice, applicable browser/service-backed, `make agent-finalize`, and `make check` commands have terminal results with no unresolved infrastructure failure, cancellation, or accidental zero-row selection. | Mandatory evidence record and retained run roots | SPECIFIED; implementation pending |
| DOD-010 | Sections 6, 9, and 10 | Workstream, tracker, and session records reflect the final live state, command results, changed files, generation disposition, resolved/reopened gates, failures, and next action without requiring rediscovery. | Current tracker and final implementation handoff | SPECIFIED; implementation pending |
| DOD-011 | Sections 1 and 11 | No owner contradiction remains unresolved. Any discovered contradiction stopped its slice and was resolved by the applicable owner before work resumed. | Owner-resolution reference or explicit `no contradiction found` record | SPECIFIED; implementation pending |
| DOD-012 | Entire tracker | No behavior-changing work was smuggled into the refactor. Any required new route, schema, error, state, default, permission, dependency, package export, feature tuple, or permanent cross-owner boundary was separately planned and authorized. | Final diff classification and owner approval for every authorized exception | SPECIFIED; implementation pending |

This tracker revision is complete when its Markdown validation passes and only
this tracker is changed by the documentation task. The refactor itself is not
complete until DOD-001 through DOD-012 all have passing implementation evidence.
This tracker does not authorize or perform that refactor.
