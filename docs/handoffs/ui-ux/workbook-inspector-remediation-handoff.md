# Workbook inspector remediation

## Scope and baseline

Implementation of the approved inspector remediation plan, including Notes source,
evidence, and related-note associations. Baseline: clean `main` at
`67c69a0abea3defe8f8a4477b3e6e49013f56dad`. Earlier completed inspector handoffs
are historical evidence only. The Cartulary UI/UX refactor skill and maintained
digest guide this work; adopted Core clauses and design direction own behavior.

## Phase ledger

| Phase | Status | Exit evidence |
| --- | --- | --- |
| P0 Close owners | Complete | Core 01 REQ-01-675 resolves Notes discovery versus exhaustive writable fields. Core 03 references retained lifetimes; design 7.3/12.4/12.5/12.7 own presentation; domain only navigates. |
| P1 Project and characterize | Complete | New API/discovery/design projections generated; additive compatibility entries registered; new admission, feedback identity, explicit panel-state, persistent-context and grouped-reason regressions catalogued. |
| P2 Complete Notes backend | Complete | Service and public-route checks pass; exact and concurrent replay, authorization, paging, rejection atomicity, History rollback, source-state round trip and projection rebuild verified. |
| P3 Repair feedback and panels | Complete | Focused feedback, explicit panels, draft/recovery, Task and Notes association tests pass; see current evidence below. |
| P4 Consolidate presentation | Complete | Persistent header, grouped typed reasons, token-backed controls; ordinary visual comparison passed at `20260920T233450Z-p667` (12/12), following full candidate review. |
| P5 Verify and hand off | Complete | Integration, visual, measurement, 704/704 fast units, final drift/type/boundary/lint gates and all A001–A027 acceptance rows pass. |

Dependencies follow phase order. A phase exit is recorded before dependent work.
Documentation or discovery success is not product verification.

P0 review: Notes management is a correction of declared but unimplemented
capabilities, expressed as an additive public resource. Explicit header/body,
panel-state vocabulary and typed reason grouping are newly adopted presentation
behavior. Field locality already existed; revision lifetime is clarified by
reference to the existing authoring contract. Core 02's canonical directions and
revision ownership and Core 04's authorization remain unchanged. No unresolved
owner contradiction blocks projection work.

## Ownership and compatibility

| Gap | Behavioral authority | Source responsibility | Verification ownership |
| --- | --- | --- | --- |
| G1 Field feedback | Core 03 ordinary authoring lifetime; design 10.8, 12.5 | Workbook draft/operation owners; stateless presentation | `web.workbook`, `web.design` |
| G2 Section completeness and Notes | Core 01 public resources/discovery; Core 02 relationships/history; Core 04 authorization | Artifacts semantics, Links storage/revisions, Workbook composition | `module.artifacts`, `module.links`, `module.workbook`, `web.workbook`, contract owners |
| G3 Header/body | design 7.3, 12.5; Core 03 focus | Shared inspector/layout | `web.workbook`, `web.design` |
| G4 Labels/reasons | Core 01 discovery; design 12.4 | Authored View Schemas and shared action presentation | `platform.viewschema`, `package.view_contracts`, `web.workbook` |
| G5 Controls | design 3.9, 12 | Shared presentation primitives, feature composition | `package.ui`, `web.architecture`, `web.design` |

Association storage remains canonical null-field Links. Existing linked-note
creation, link IDs, provenance, history and receipts remain supported. Additive
indexes need no backfill. Notes tags remain PATCH; pivots remain navigation.
New routes and bindings require coordinated backend/client/frontend delivery.
No fake PATCH adapter or old-label alias will be retained. Source-specific
mutation and recovery owners remain; no configurable workflow engine is added.

## Verification and handoff ledger

Focused backend product passes are recorded below. Each further check records its command,
result and retained run root where available. Final review includes current
digest acceptance rows, required scenarios from the approved plan, public API
compatibility, additive migration details, reviewed screenshots, removed
alternatives and retained responsibilities. Applicable blocked rows prevent
completion.

P1 evidence: `make generate` passed at
`.cartulary/test-results/20260920T213119Z-p72921`. Earlier generation failures
identified the new API compatibility entries, specialized action-key validation,
and catalog sorting; their authored inputs were corrected. Narrow `test-slice`
runs at `20260920T213514Z-p78836` and `20260920T213514Z-p78853` expose the intended
missing admission API, feedback/panel components and scrolling body. Those are
expected red characterization results, not a product pass. Initial runs could
not spawn `node`; adding the repository's `tmp/node-runtime/bin` to PATH resolved
that environment issue. Subsequent commands retain that PATH.

Capability mapping: Notes source/related-note controls belong in Relationships,
evidence controls in Evidence, and each reaches the new resource by explicit
kind. Tags stay with ordinary Details editing; source pivots stay with navigation;
History and contextual Workflow creation retain their existing owners. G1's
revision test supplements existing draft/recovery rows. G2's panel-state test
supplements feature-owned behavior tests. G3/G4 have focused regressions; G5 uses
existing presentation/import-boundary and visual checks instead of tests that
merely mirror CSS literals. Backend service scenarios and mounted-editor browser
scenarios are added alongside their implementing phase.

Presentation rollback must preserve committed associations and history. After
association writes exist, prefer frontend rollback or a forward fix; server
rollback requires evidence that projection/history readers understand the links.
Never delete relationships or rewrite history to simplify rollback.

P2 evidence: the Notes service row passed at
`.cartulary/test-results/20260920T220100Z-p15971`; the public route/History row
passed at `20260920T215940Z-p96948`. These cover existing contextual associations,
all three kinds, invalid/deleted/foreign counterparts, duplicate adds, atomic
removal rejection, incoming read-only traversal, version conflicts, original and
concurrent receipts, role/incident changes, bounded pagination, cursor binding,
CSRF/session checks, removal and historical restoration of the same link identity.
Links source-state export/import preserves canonical bytes and provenance;
rebuilding the Note projection preserves its returned row. Entities source
boundary checks passed at `20260920T220111Z-p24472`. Generated client operations
were added to the authored protocol selection and generation passed at
`20260920T215810Z-p93280`. Earlier service failures were fixture omissions (Evidence
source row and incident closure timestamp), corrected before these passes.
Full integration, browser, compatibility and acceptance gates remain P5 work.

P3 evidence: focused Workbook rows passed at `20260920T223417Z-p81267` (7/7),
Task lifecycle feedback passed at `20260920T223433Z-p82490`, and mounted shell,
entity and supersession composition passed at `20260920T222552Z-p72217`.
Notes regressions cover captured replay, accepted receipts before refresh,
read-only recovery, authority concealment, replacement retirement, stale paging,
incoming navigation, and no-op receipts. Field errors retain canonical identity
and authoring revision; Task errors retain their exact draft object. Explicit
panel contributions cover all four current inspector consumers, including Notes'
three real association operations. Related-note management is declared in
Relationships, matching its owner contribution. Query-window previews qualify
empty results and retain data through failed refresh. No missing implementation
is converted to a successful empty result. Fresh browser/visual acceptance is
still required by P4/P5. Typecheck passed at `20260920T223214Z-p76414`; a later
required-prop mismatch was corrected at the composition boundary before P4.
Exact association replay after incident closure passed the service row at
`20260920T222400Z-p51176`; fresh writes still reject closure.


## P4 presentation review

Mounted Notes, scalar, collection/reference, persistent-header, accessibility and
public-route rows passed together at `20260920T231000Z-p70228` (14/14 execution
units). The service row passed after source-boundary cleanup at
`20260920T230409Z-p67995`; the receipt lookup stays inside its admitted transaction,
including concurrent exact replay. Feedback/panel/reason/Notes regressions passed
at `20260920T231324Z-p46639` (6/6).

Manual offline advice query `keyboard focus local error recovery` returned five
results. R002 keyboard/visible focus guidance and R006/R008 owner-specific recovery
were adopted within existing owners; enhanced focus advice did not introduce a
new conformance profile. No generated design system or dependencies were added.

The first complete ordinary visual reconciliation, `20260920T225744Z-p1964`,
accounted for 252 captures and goldens, 29 registered fixtures, and zero missing,
orphan or ambiguous mappings. All functional assertions completed; differences
were expected presentation changes. Initial refresh `20260920T230417Z-p73351`
passed 12/12 units. All 37 changed candidates were inspected. Follow-up review
corrected explicit fixture focus for form-start captures and a generated actor-ID
mask in Decision recovery; subsequent settled-source comparison passed below.
An intervening visual run `20260920T230959Z-p70063` had a missing fixture import
(now corrected), so its five uncaptured goldens were retained, not deleted.

Accepted refresh triggers: adopted fixed header/body, authored concise labels,
grouped permission explanations, shared controls, and the explicit fixture focus
and generated-ID mask corrections. Viewports, zoom, crop scopes and renderer pins
are unchanged. Inspector scroll anchors now target the content body. Timeline
Evidence and coordination form-start captures explicitly focus their first control
before anchoring. Only `decision-supersession-accepted-linux.png` adds a screenshot
mask over the generated current-actor draft input; recovery text remains visible.
These images are implementation-support evidence, not published conformance.

### Golden accounting

Paths below are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
`active nonregistry` means the exact catalog/scenario capture is reconciled without
a registered fixture; it does not imply missing coverage or a deletion candidate.

| Golden | Authored owner row | Stable fixture |
| --- | --- | --- |
| `contextual-decision-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | active nonregistry |
| `contextual-decision-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | active nonregistry |
| `contextual-decision-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | active nonregistry |
| `contextual-task-request-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | active nonregistry |
| `contextual-task-request-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | active nonregistry |
| `contextual-task-request-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | active nonregistry |
| `coordination-comm-log-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `decision-supersession-accepted-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | active nonregistry |
| `decision-supersession-review-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | active nonregistry |
| `decision-supersession-review-narrow-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | active nonregistry |
| `entity-mention-chip-states-linux.png` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` |
| `indicator-lifecycle-authoring-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-observation-authoring-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | active nonregistry |
| `linked-note-authoring-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | active nonregistry |
| `linked-note-source-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | active nonregistry |
| `record-relationships-mention-chips-linux.png` | `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | active nonregistry |
| `timeline-related-evidence-authoring-linux.png` | `module.workbook.visual.timeline_related_evidence` | active nonregistry |
| `timeline-related-evidence-authoring-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | active nonregistry |
| `timeline-related-evidence-party-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | active nonregistry |
| `timeline-supersession-authoring-linux.png` | `module.workbook.visual.timeline_capture_actions` | active nonregistry |
| `timeline-supersession-review-linux.png` | `module.workbook.visual.timeline_capture_actions` | active nonregistry |
| `timeline-supersession-review-narrow-linux.png` | `module.workbook.visual.timeline_capture_actions` | active nonregistry |
| `workbook-inspector-compact-actions-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-destructive-confirmation-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.destructive_actions` |
| `workbook-inspector-history-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |


## Structural and retirement review

The common boundary is inspector presentation: data/access state, local feedback
association, fixed context, and action reason rendering. Source owners still decide
what the data means and which operations are admitted. This allows another declared
surface to contribute its own model without adding a generic workflow engine.
Feedback identity includes the captured authoring revision, so new editor kinds use
the same rejection lifetime. Typed reason identity survives future copy changes.

Removed alternatives: heading-only panel composition for supported consumers,
Notes' generic Party/inactive Evidence delegation, Notes association PATCH bindings,
ordinary field errors in the detached bottom channel, repeated per-action reason
paragraphs, generic inspector button/input styling, and outer-shell scroll anchors.
Labels have no compatibility aliases. Required History, Evidence, source mutation,
creation draft, retained receipt and recovery owners remain because they preserve
real data, transactional and authorization responsibilities.

Current consumers are covered as follows:

| Consumer | Details | Relationships | Evidence | History / Workflow |
| --- | --- | --- | --- | --- |
| Generic | Retained edit with local validation or declared read-only fields | Declared reference summaries and existing Party controls; Notes source and related-note resource | Existing Evidence owner; Notes evidence associations | Existing History and declared feature operation owners |
| Entity | Scalar/alias editing with captured feedback | Existing owned relationships and truthful query-window preview states | Projected evidence count and owner controls | Existing History, contextual creation and merge owners |
| Assessment | Explicit declared field values | Existing support references | Existing Evidence contribution | Existing History and Assessment owners |
| Timeline | Existing operational authoring | Owned mentions, tokens and observations | Existing attachment/file workflow | Existing History, capture and creation owners |

Migration 45 adds only incoming/outgoing null-field live-link lookup indexes.
No backfill, relationship rewrite, receipt deletion or persistence adapter is
introduced. Artifacts admits association semantics; Links stores and revises them;
Records provides envelope locks; source adapters own labels/projection updates;
Workbook composes transport. Cursor state stays server-owned and actor-bound.
The new backend and generated clients must ship before or with the consuming UI.

Final visual-update run `20260920T231530Z-p53916` passed 12/12 units and reconciled
all 252 active goldens with all 29 declared fixtures. Every changed candidate was
reviewed, including the corrected form focus and explicit identity mask. A normal
visual comparison subsequently passed at `20260920T233450Z-p667`. The real Notes screenshot from
`20260920T231531Z-p54041` was also reviewed and its spacing consolidated. Subsequent
receipt refresh coalescing passed at `20260920T231931Z-p31912`; older pending refresh
notices clear only when all association reads cover their accepted version and
receipt effects have been applied. Superseded read failures cannot downgrade a
completed refresh.

One accessibility/browser attempt, `20260920T232006Z-p38164`, was rejected by the
build receipt guard because golden promotion changed the source snapshot during
its build. No product conclusion is drawn from that run; it is rerun against the
settled source. `make agent-finalize` passed at `20260920T231949Z-p32658`.
Retained successful-run maintenance was skipped because `RESULTS_DIR` was unset.


P4 keyboard/navigation and source creation selection passed at
`20260920T232114Z-p92431` (15/15 units). Normal visual run
`20260920T232036Z-p56662` passed both browser groups (47 scenarios), including all
252 screenshot comparisons. Its target reconciliation failed solely because the
concurrent finalization snapshot restoration replaced the freshly promoted golden
manifest with the older hashes. The images were not changed by that restoration.
The isolated supported update passed at `20260920T232701Z-p35043`; its 252 hashes
were checked against the files. Sequential finalization passed at
`20260920T233421Z-p96217`, then the ordinary visual target passed 12/12 at
`20260920T233450Z-p667`. No failed target is reported as a complete visual pass.

The association read-version floor regression passed at
`20260920T232759Z-p67311`: successful list reads advance the known Note version,
and older later responses retain the newer data with a stale-failure explanation.
Public-route label coverage also now exercises Timeline's authored synopsis rather
than relying on a fallback identity when a field mapping is incorrect.

## P5 integration findings and disposition

Initial broad fast run `20260920T233608Z-p71755` completed 699/704 units. The
remaining failures exposed two migration-catalog hash characterizations, the
operator migration-evidence digest, an assertion comparing rendered reason text
with its typed model, and a scalar fixture using `change` instead of the editor's
native `input` event. Migration 45's catalog and evidence expectations were
updated; the presentation and input assertions now test their actual contracts.
Focused migration checks passed at `20260920T234619Z-p16738`; operator evidence
passed at `20260920T234705Z-p18721`; authorization rendering passed at
`20260920T234620Z-p17121`; save-state fixture passed at
`20260920T234750Z-p19799`.

The selected stateful run `20260920T233505Z-p42983` completed 16/19 units and
found two additional issues. Synthetic scalar paste now explicitly commits after
local authoring, matching the native editor contract. Queue-capacity refusal had
left a permanent deduplication result for an unchanged draft, preventing explicit
retry after capacity recovered. The Timeline driver now reports admission refusal
separately from settlement; only the unadmitted scalar/collection cache entry is
retired. Retained text, submitted failures and uncertain request identities remain
owned by their existing owners. Both refusal paths and submitted validation
retention pass the draft-registry row at `20260920T235051Z-p64983`. This is a
required continuity correction under Core 03 REQ-03-099, not a new retry policy.
The capacity browser row passed in `20260920T235115Z-p66274`. That run exposed
another stale settlement: a recovered request left the attached editor returning
its earlier rejection. Halted operation observers now remain until the same
retained unit resolves, updating the command's cached settlement without restoring
old focus or repeating the original promise's navigation. The recovery unit row
passed at `20260920T235938Z-p81918`, and both public-route browser scenarios passed
at `20260920T235938Z-p81885` (11/11 execution units). Discard, fresh-ID Retry and
same-field resolver handoff retain their separate semantics.

Measurement run `20260920T233450Z-p724` passed 23/25 execution units, with one
Timeline creation paint predicate over threshold while several suites ran
concurrently. Lower-contention rerun `20260920T234911Z-p21186` passed 25/25 units;
Timeline creation p95 was 74.6 ms against the unchanged 150 ms predicate. These
measurements remain implementation evidence, not a published claim. No performance
pass is inferred from the failed run.

The broad fast rerun passed 704/704 units at `20260920T235545Z-p17853`.
The final recovery observer correction has additional focused unit and production
browser passes above. No required scenario is deferred.

## Delivery and evidence map

All run identifiers below resolve under `.cartulary/test-results/`. Row selections
and their results are retained in each run's machine manifest and row artifacts;
this document does not define executable routing.

| Change area | Authored and source paths |
| --- | --- |
| Behavioral owners | `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, `docs/design.md`; vocabulary/navigation in `docs/domain.md` |
| API and discovery | `contracts/openapi-source/owners/module.workbook/openapi.json`, `contracts/openapi-source/owners/platform.viewschema/openapi.json`, `contracts/openapi-releases/2.0.0.change-set.json`, `contracts/protocol-ts/http-operations.v2.json`, `contracts/view-inspector/index.json`, authored `contracts/view-schemas/` |
| Shared presentation contract | `contracts/design/presentation.v1.json`, its schema/generator and generated UI facade |
| Notes server operation | Artifacts `note_association_admission.go`, `note_associations.go`, `note_association_mutation.go`; Links `association_lookup.go`; Workbook `note_associations.go`; `internal/app/workbookassembly/note_associations.go` |
| Shared inspector and consumers | `apps/web/src/workbook/inspector/`, generic/entity/assessment/Timeline compositions, shared `workbookFormStyles.ts` |
| Notes client lifetime | `apps/web/src/workbook/features/notes/WorkbookNoteAssociationOwner.ts`, its panel/recovery models and adapters; Workbook runtime and collaboration composition |
| Persistence | `db/migrations/00045_note_association_lookup.sql`, migration/source ownership projections, Links revision contribution |
| Verification | Authored owner catalogs, Notes API/service tests, inspector/Notes unit tests, `apps/web/e2e/note-associations.spec.ts`, reviewed visual fixtures; existing Timeline recovery scenarios strengthened |

| Verification | Result and run |
| --- | --- |
| `make service-backed-test-slice OWNER=module.artifacts ROWS=…` | PASS, 3/3 units, `20260920T230409Z-p67995` |
| Notes public route / History selection | PASS, 3/3 units, `20260920T232923Z-p71095` |
| Inspector field/panel/reason/Notes unit selection | PASS, 6/6 units, `20260920T231324Z-p46639`; Notes version-floor extension PASS, `20260920T232759Z-p67311` |
| Mounted Notes, scalar, collection/reference, header, accessibility and public route selection | PASS, 14/14 units, `20260920T231000Z-p70228` |
| Workbook keyboard, navigation, creation and reference accessibility selection | PASS, 15/15 units, `20260920T232114Z-p92431` |
| `module.revisions` selected stateful inspector row | PASS, 11/11 units, `20260920T233544Z-p3948` |
| `module.evidence` selected stateful/accessibility rows | PASS, 13/13 units, `20260920T233544Z-p3986` |
| `module.entities` selected stateful/accessibility rows | PASS, 13/13 units, `20260920T233544Z-p4053` |
| Workbook stateful continuity/authorization/history/Notes rows | PASS rows in `20260920T233505Z-p42983`; failed recovery rows resolved by the following two entries |
| Workbook capacity/replay status row | PASS row, `20260920T235115Z-p66274`; that aggregate's unrelated recovery-row failure is recorded above |
| Workbook public mutation/recovery browser row | PASS, 11/11 units, `20260920T235938Z-p81885` |
| Timeline draft/refusal and Workbook recovery unit rows | PASS, `20260920T235051Z-p64983`, `20260920T235938Z-p81918` |
| `make test-slice OWNER=package.ui` | PASS, 10/10 units, `20260920T233544Z-p4077` |
| `make browser-e2e-visual-update` followed by ordinary comparison | PASS, `20260920T232701Z-p35043` and `20260920T233450Z-p667`; 252 goldens/29 fixtures, all 37 changed images reviewed |
| `make browser-e2e-measurement` | PASS, 25/25 units, `20260920T234911Z-p21186` |
| `make test-fast` | PASS, 704/704 units, `20260920T235545Z-p17853` |
| `make agent-finalize` | PASS, `20260921T000217Z-p23631`; retained successful full-warm-run maintenance skipped because `RESULTS_DIR` was unset |
| `make generate-drift` / generated artifact policy | PASS, `20260921T000407Z-p28121` / `20260921T000407Z-p28123` |
| `make json-shape-check` / OpenAPI compatibility | PASS, `20260921T000407Z-p28125` / `20260921T000407Z-p28131` |
| `make migration-drift` | PASS, `20260921T000407Z-p28129` |
| `make frontend-typecheck` | PASS, `20260921T000407Z-p28179` |
| Frontend import / backend module boundaries | PASS, `20260921T000407Z-p28183` / `20260921T000407Z-p28185` |
| `make lint-biome` | PASS, `20260921T000407Z-p28189` |
| `make lint-markdown` | PASS, `20260921T000551Z-p37482` |

## Acceptance assessment

This assessment applies the digest to the implemented scope. Broader retained
features are regression evidence, not newly adopted behavior or conformance claims.

| Row | Status | Evidence and scope |
| --- | --- | --- |
| A001 Authority | PASS | Owner map above; Core 01 REQ-01-615/675 and §7.4.1A, Core 03 authoring and REQ-03-099, design §§7.3/10.8/12.4/12.5/12.7. Core 02 and Core 04 retain relationship/history and authorization authority. |
| A002 Scope | PASS | Shared presentation hides state/feedback/context/reason decisions; feature owners contribute semantics. Structural, correction and newly adopted behavior are distinguished in P0; retirement/extension rationale is recorded. |
| A003 Repository state | PASS | Clean baseline and current source guides, owner catalogs, manifests and import/generated policies inspected. Current task guides replace historical routing; Grid Adapter remains the sole vendor integration boundary. |
| A004 Tokens | PASS | New controls use authored component/spacing tokens; line clamp and state vocabulary are projected. No second style registry; import/type/visual checks pass. |
| A005 Theme | PASS | Existing graphite theme retained; current production visual fixtures reviewed. No theme/font/icon dependency introduced. |
| A006 Density | PASS | Grid density owner unchanged; independent inspector form spacing follows design §3.9. Grid/UI fast checks, measurement and base/compact visual cases pass. |
| A007 Creation | PASS | Existing atomic linked-note creation retained; association actions use declared discovery and real routes. Contextual creation, source replacement and authority regressions pass. |
| A008 Responsive | PASS | Persistent-header browser matrix covers base, 1024×720, 768×640, short height, zoom and text spacing. Existing clamp/resize/ARIA contracts remain under grid/layout unit and visual coverage. |
| A009 Overflow | PASS | One inspector body scrollport, outer separator, fixed context/Close; shell status and navigation geometry covered by header and stateful/visual scenarios. |
| A010 Inspector | PASS | Exact semantic routing and required disabled controls preserved; all four consumers provide explicit panel models. Dispatcher, review authorization and production recovery checks pass. |
| A011 Continuity | PASS | Raw drafts and captured identities survive navigation/detachment. New field revision, Notes replay/read floor, query continuation and repaired queue refusal/recovery scenarios pass. |
| A012 Transactions | PASS | Existing secure ID port supplies fresh identities; Notes captures immutable request bytes. Duplicate activation, lost response/exact replay, concurrent backend replay and public transaction recovery pass. |
| A013 Recovery | PASS | Accepted Notes receipts precede reads and read-only refresh; retained halted Timeline settlement updates after recovery. Fresh-ID Retry remains limited to its existing queue owner; discard/uncertain replay remain distinct. |
| A014 Editing | PASS | Mounted scalar, collection and reference rejection retains authoring; stale revision/attachment failures are fenced. Grid commit/Escape/input/paste and recovery scenarios pass. |
| A015 Conflict | PASS | Existing cell conflict and local/saved value owners retained; production same-field resolver handoff, fast conflict tests and conflict/rollback visuals pass. |
| A016 Data/access | PASS | Explicit loading/ready/refreshing/stale/unavailable data models compose independently with readable/concealed access. Missing declared contributions fail coverage; read-only content is not empty. |
| A017 Authorization | PASS | Notes owner suspension versus account/incident retirement, stale replies and current permissions tested. Server transaction/page authorization and browser authority/session continuation/revocation selections pass. |
| A018 Evidence | PASS | Notes attaches canonical Evidence associations without fetching bytes. Existing attachment/preview/download/blocked state owners and Evidence accessibility/stateful/visual matrix pass. |
| A019 Accessibility | PASS | Local invalid/describedBy, single rejection announcement, grouped reasons, full title name, keyboard close/navigation/focus and recovery checks pass under the adopted profile. |
| A020 Components | PASS | Shared primitives, mixed enabled/disabled groups, long labels, text spacing/zoom, narrow and compact cases reviewed and tested. |
| A021 Virtualization | PASS | Grid data/identity owner unchanged; no fake rows or whole-association synchronous metadata read added. Query continuity, Grid Adapter fast tests and measurement suite pass. |
| A022 Visual fixtures | PASS | Production renderer, unchanged renderer/viewport/crop/zoom pins, body anchors and the explicit generated-ID mask accounted for. All changed artifacts and populated Notes screenshot reviewed. |
| A023 Selectors | PASS | New scenarios use schema, record, field, feature and owner-state identities; `package.ui` selection passed 10/10. |
| A024 Test authority | PASS | Runtime, tests and generators consume typed authored inputs/generated facades. No dependency on Markdown, digest paths, prose hashes or line numbers was added. |
| A025 Generated artifacts | PASS | Authored-to-generated review and final settled-source generation, policy, JSON, API compatibility and migration drift gates pass. |
| A026 Compatibility | PASS | Additive route and indexes; supported canonical links/history/receipts and linked creation retained. No fabricated PATCH adapter, old-label alias, generic read endpoint or persistence setting. Coordinated delivery and rollback bounds are recorded. |
| A027 Handoff | PASS | Phase exits, owners, paths, evidence, failures and rollout are recorded; final static checks and documentation lint pass. All required acceptance rows are closed. |

## Release and rollback handoff

Ship migration 45 and additive server/generated contracts before or with the
frontend. Existing association data requires no backfill. Older clients retain
their existing omission behavior for unsupported bindings. The compatibility
check covers the authored OpenAPI release change set; it does not substitute for
the coordinated deployment order.

Presentation rollback may revert its source/tests/goldens as a unit. Keep committed
associations, link identities, provenance, revisions and receipts. After new writes,
prefer frontend rollback or forward repair; a server rollback must first establish
that the earlier projection/history owners understand the canonical links.
Migration Down removes only the two added indexes and must not rewrite user data.

No required capability remains deferred. No framework, persistent preference,
new icon package or generic inspector workflow engine was introduced. Deployment
and publication are outside this workspace implementation; no release or Core 05
conformance claim is made. The next action is review and coordinated release of
the complete change set.
