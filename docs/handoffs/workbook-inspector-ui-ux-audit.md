# Workbook inspector specification and implementation remediation plan

## 1. Scope and source posture

**Controlling artifact:** this file. **Planning status: DONE**. Product
implementation status: **DONE**, including mandatory WS-10 validation and handoff. The user approved
implementation of WS-00 through WS-10, the reading-first design defaults, and a
coordinated server/browser History cutover without a legacy response adapter.
Execute slices serially; record each passing exit and update this tracker before
beginning the next slice. Planning evidence does not establish product readiness.

Rebuild the inspector around reading a selected record, examining its connections,
and taking a deliberate action. Use a bounded record header, persistent section
navigation, aligned saved-value rows, and local editing/action disclosure. Repair
History's missing semantic detail through its source owners as well as its UI.
Retire competing styling and presentation paths as their callers migrate.

The user requested durable remediation, not preservation of incidental behavior.
Retain a capability because it fulfills an adopted obligation or has a concrete
future benefit. Existing appearance, wording, helper structure and passing tests
are not compatibility requirements. The audit's original presentation-only
assumption is superseded for History completeness (G10).

### Authority and limits

- Adopted subsystem owners and Core 00–04 govern their named behavior. Core 05
  applies only to claim-bearing publication; this plan makes no such claim.
- [Design](../design.md) owns presentation direction; [Domain](../domain.md)
  owns vocabulary and owner navigation. Neither becomes a competing route,
  permission, lifecycle or storage authority.
- [NLSpec research](../research/nlspec-spec.md) supplies evaluation methods:
  independent-implementer equivalence, explicit defaults, complete mappings,
  omission/error behavior and binary completion criteria. Its agent-directed
  appendices are source material, not instructions to execute this plan or a
  newly adopted Cartulary specification. Its generic definition of a design
  document does not override Core 00's adoption of Cartulary's design owner.
- This audit, the [planning framework](cartulary_modular_refactor_planning_framework.md)
  and the UI/UX digest are planning/support material. Requirements proposed here
  enter their actual owner before implementation. Tests and generated projections
  never read, stat or hash Markdown to obtain product behavior.
- A genuine conflict between adopted owners is `BLOCKED: owner contradiction`;
  stop only the affected dependency chain. No contradiction was established for
  the proposed presentation changes. Missing History detail is an implementation
  and projection deficiency, not permission to weaken REQ-03-139.

Exact Core navigation: [status and precedence](../spec/00_document_set_status_and_precedence.md),
[view and History contracts](../spec/01_architecture_storage_and_view_contracts.md),
[source model and history](../spec/02_domain_model_schema_and_history.md),
[workbook interaction](../spec/03_workbook_interaction_collaboration_and_workflows.md),
and [security](../spec/04_security_deployment_and_conformance.md).

Baseline: `main` at `67f3d98ac01b42a1da4409c0a6630d08d24e1761`.
At this planning task's start, this audit was already staged as an added file;
there were no other changes. The original audit used seven retained screenshots,
source inspection and scenario definitions, without a live browser or user study.
This pass rechecked source and owners; it does not turn retained evidence into a
fresh execution result. Only this tracker may be changed during planning.

Target label: `workbook-inspector`. The target is the set of presentation and
History seams inventoried below, not every file under Workbook. No new theme,
inspector density preference, workflow engine, generic retry engine, persistent
draft storage, grid vendor integration, or richer Evidence enumeration is planned.

### Planning execution checkpoints

Update this tracker after completing each checkpoint and before starting the
next. These planning statuses are independent of the implementation tracker.

| Checkpoint | Status | Evidence / exit |
| --- | --- | --- |
| P-01 — Evidence and owner review | DONE | Current presentation, Core 00–04 seams, design/domain/research guidance, History materialization/OpenAPI, source boundaries and verification routing inspected; History completeness escalated into G10. Recorded before P-02 began. |
| P-02 — Remediation and workstream definition | DONE | G00–G11 cover all nine audit findings plus specification, History data and validation gaps; eleven implementation workstreams have owners, migrations, dependencies, risks and exits. Recorded before P-03 began. |
| P-03 — Validation and handoff completion | DONE | Planning review, four owner task guides, whitespace checks and Markdown lint passed at unchanged baseline; lint run `20260921T033650Z-p1857`. The user then approved implementation. |

## 2. Current-state repository inventory

Paths below are repository-relative. Each row accounts for the named seam;
paths grouped in a row share its responsibility and risk. Source ownership is
`web.workbook` unless specified. Verification ownership is separately listed in
§8. Inspected code establishes structure, not a passing runtime baseline.

| Path | Responsibility / exported surface | Inbound callers | Outbound dependencies | Tests / contracts | Target owner and risk |
| --- | --- | --- | --- | --- | --- |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorShell.tsx` | `WorkbookInspectorShell`, `WorkbookInspectorPanelSection`; record header, body, section headings | Timeline, Generic, Entity, Assessment compositions | UI/view contracts, layout, subject and presentation types | `WorkbookInspectorPresentation.test.tsx`; design §§7.3,12.7 | Shared inspector presentation; high focus/concealment risk |
| `apps/web/src/workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx` | Ordered admitted panels and contextual actions | Feature inspector compositions | Capability resolver, panel models, section renderer | Presentation tests; Core 01 REQ-01-615–617 | Inspector composition; high coverage risk |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorPanelContent.tsx` | Explicit data/access regions, owner delivery slots | All declared panel producers | Typed design states, owner-supplied content/notices | `WorkbookInspectorPanelContent.test.tsx`; `contracts/design/presentation.v1.json` | Keep shared state rendering; high lifetime/security risk |
| `apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx` | Adjacent/overlay composition, resize and return focus | Workbook surfaces | Existing layout metrics and overlay host | Browser geometry and inspector fixtures; layout tokens | Layout owner; integrate navigation without moving its geometry policy |
| `apps/web/src/workbook/inspector/WorkbookInspectorDetails.tsx`, `WorkbookInspectorEditControl.tsx`, `WorkbookInspectorDraftFeedback.tsx` | Accepted field rows, attached controls, draft/review feedback | Timeline Details, Generic/Entity editing | View contracts, draft port, reference and mutation controls | `TimelineInspectorDetails.test.tsx`, draft tests, inspector edit browser rows | Shared Details presentation; high authoring risk |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorActions.tsx` and `apps/web/src/workbook/components/workbookFormStyles.ts` | Buttons, action groups, shared form roles | Inspector and specialized forms | Authored UI tokens, typed disabled reasons | Presentation tests and `package.ui` tests | Shared presentation; medium cross-consumer visual risk |
| `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx`, `TimelineMentionActionControls.tsx` | Source-local mention groups, selected controls, target filtering | Timeline inspector | Existing mention operation owner, candidate picker, semantic focus registry | Timeline relationship/mention regressions | Timeline presentation; high identity/recovery risk |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.tsx` | Count, file chooser, drop/paste and retained recovery | Timeline inspector | Evidence attachment context and file owner | `TimelineEvidencePanel.test.tsx`; Evidence owner contracts | Timeline/Evidence presentation; high security and duplicate-write risk |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookStyles.ts` | Shared Timeline style exports used by inspector and other callers | Timeline components | Design variables plus local literals | Density/visual evidence | Migrate inspector callers; do not delete exports still used by grid or other surfaces |
| `apps/web/src/workbook/inspector/workbookHistoryPresentationModel.ts`, `WorkbookRecordHistoryPresentation.tsx`, `presentation/WorkbookHistoryPresentation.tsx` | Event mapping, event/reversal/delete/restore presentation | Inspector History composition | History action permissions, opaque selectors, safe technical fields | `workbookHistoryPresentationModel.test.ts`, History recovery tests | History-to-presentation adapter and stateless rendering; high review-scope risk |
| `apps/web/src/workbook/history/workbookHistoryItem.ts` | History normalization, selector construction, committed-content equality | History browsing and action owners | Protocol history model | History browsing/recovery tests | Keep History semantics here; high rollout/cache-epoch risk |
| `apps/web/src/workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `features/entities/EntityWorkbookInspector.tsx`, `features/assessments/AssessmentWorkbookInspector.tsx` | Surface-specific region and editor composition | Respective feature surfaces | Shared inspector, feature-owned commands | Generic/Entity/Assessment tests and browser flows | Feature composition; medium presentation, high source-context risk |
| `apps/web/src/workbook/inspector/InspectorCreateRelatedWorkflow.tsx` | Dispatch into Note, coordination, contextual and related-Evidence forms | Inspector feature compositions | Existing creation contexts, capability metadata | Related-creation tests and specialized browser scenarios | Keep dispatch in composition; never move it into stateless presentation |
| `apps/web/src/workbook/features/indicators/IndicatorLifecycleWorkflow.tsx`, `indicatorLifecycleStyles.ts` | Lifecycle controls, retained draft binding, local styling | Indicator feature contribution | Lifecycle owner, paging, feedback and support picker | Lifecycle workflow/reconciliation tests | Share visuals only; retain feature semantics |
| `internal/modules/revisions/history_materializer.go`, `history_service.go`, `history_model.go` | Materialize and serialize row-centric History | Revisions History HTTP adapter | Retained mutations/revisions, attribution, rollback availability | `history_test.go`, `history_components_test.go`; Revisions OpenAPI | Revisions coordination; high disclosure/compatibility risk |
| `internal/modules/revisions/provider_contributions.go`, `target_semantics_catalog.go`; `internal/app/revisionassembly/` | Source-owned provider catalog and application composition | Revisions assembly | Source-owner providers, generated target requirements | Catalog and source-boundary tests | Revisions owns catalog validation; source modules own semantic history units |
| `contracts/openapi-source/owners/module.revisions/openapi.json`; `packages/protocol-ts/src/generated/core-http-types.ts` | Authored History transport projection and generated browser types | OpenAPI generation and History adapters | Owner requirements | `RecordHistoryDiffSummary.units` is currently open objects | Edit authored source only; generated types inspected read-only |
| `contracts/design/tokens.v1.json`, `presentation.v1.json`; `contracts/view-schemas/cartulary.view.timeline.v2.json` | Tokens, presentation states, immutable schema fields/panels | UI/view contract facades and workbook consumers | Adopted design/Core owners | UI contract/generation checks | `package.ui` / view projections; avoid new field or panel registries |
| `apps/web/e2e/workbook.visual.spec.ts`, `workbook-inspector-edit.spec.ts`; `tools/frontend_visual_fixture_registry.json` | Production-renderer visual and behavioral evidence | Browser harness | Semantic fixtures, selectors, declared capture parameters | Public browser targets | Test owners; missing first-viewport Details capture is G11 |
| `tools/frontend_source_ownership.json`, `frontend_import_boundaries.json`, `generated_artifact_policy.json`, `test_families/{web.workbook,module.workbook,module.revisions,package.ui}.json`, `task_surface_owner.json` | Source/import ownership, generation boundaries and independent routing | Repo harness and generators | Authored machine inputs | Public investigation/generation targets | Harness owners; do not infer requirements from routing |

The inspector, presentation and History README guides were read as navigation.
The package manifests confirm the existing React, UI/view/protocol package and
Lucide boundaries; no dependency or framework change is justified.

**Explicit exclusions:** remaining inspector controllers/stores, feature operation
owners, source-provider implementations, adapters, grid-adapter internals,
Evidence storage/access internals, unrelated modules and tests are regression
neighbors, not exhaustively audited targets. Their guides/imports/catalog rows
were used for navigation where stated. No file move is proposed without reading
its implementation and consumers in its workstream. WS-06 must enumerate every
participating source provider before changing its interface. Existing seven-image
review and its limitations remain historical; no images were recaptured here.

## 3. Module boundary diagnosis

The inspector is a presentation/composition seam, not a new domain module. A
shared interface is justified when it hides the decision about presenting and
reaching an owner-supplied subject, section, field, action or state. It must not
absorb the decision about permission, request admission, domain validation or
operation lifetime.

| Responsibility | Current location | Owner candidate | Disposition | Evidence and design consequence |
| --- | --- | --- | --- | --- |
| Record frame, typography, accessible section navigation | Shared shell and layout | Inspector presentation, with geometry retained in layout | keep / extend | Shell already isolates header and one scroll body; derive one admitted section descriptor sequence for both navigation and body. |
| Panel access and contribution admission | Declared list and explicit region model | Inspector composition plus feature producers | keep | Concealed unions have no payload; reuse that decision, never scan React children or register a second permission map. |
| Ordinary field presentation and command placement | Details plus injected editor children | Shared Details presentation | split | Give editor content and owner-supplied actions explicit slots; group Update/Close without introspecting children or owning drafts. |
| Styling decisions repeated across forms | Timeline, Indicator, shared styles | Existing shared form/presentation boundary | move | Migrate actual callers and remove dead local alternatives; no generic form engine. |
| Source-specific authoring, retries and receipts | Feature/runtime owners | Existing source owners | keep | Their different validation, review and security lifetimes are material behavior, not duplicate visual machinery. |
| History semantic interpretation | Coarse Revisions materializer and frontend string presentation | Source-owned pure history contributions; Revisions coordinates; frontend renders typed facts | split | Core 02 REQ-02-218 prohibits central source-type logic; the current provider catalog is the extension seam. |
| Rich Evidence enumeration | Count plus separate owner access contributions | Evidence source/read owner | defer | Count does not supply item metadata or grant access. No new enumeration feature is needed to close this presentation plan. |
| Evidence accounting | Test catalog/fixture inputs | Current verification/harness owners | keep | Add missing scenario coverage through authored routing; no documentation-driven runtime tests. |

A new surface should supply subject, admitted ordered regions, fields and command
presentation through existing composition. It should need neither a new shell
copy nor a surface-name switch in shared rendering. A new history source supplies
its own semantic projector through a checked catalog; it does not add storage
knowledge to the frontend or a central Revisions switch.

| Coupling / boundary finding | Classification | Owner / required action |
| --- | --- | --- |
| Repeated local styling prevents complete role propagation | must_fix | G06; consolidate presentation decisions and migrate real callers. |
| Shared presentation already forbids feature/controller/mutation imports | intentional/no_action | Retain the existing import guard; navigation adds local presentation state only. |
| History units are untyped and the renderer loses semantic detail | must_fix | G10/G04; source-owned typed projection, Revisions coordination and a pure display adapter. |
| A universal form or retry engine would merge different owner lifetimes | intentional/no_action | Retain separate feature operation owners; share visual roles, not domain policy. |
| Rich Evidence enumeration lacks an established data contribution in the audited count panel | defer | G09/O-01; do not create a new read dependency to beautify a count. |
| Framework snapshots omit some current generated boundaries | must_fix | Planning correction: current generated policy also protects view-contract generated files and aggregate OpenAPI; use it rather than prose inventories. |

## 4. Public contract and behavior disposition map

This is a justified contract boundary, not a freeze of incidental UI behavior.
Navigation, layout, disclosure, action copy and hierarchy deliberately change.
History projection is corrected. Other public contracts stay stable because the
identified gaps do not justify changing them.

| Contract | Governing owner | Current evidence / tests | Required characterization | Risk / disposition |
| --- | --- | --- | --- | --- |
| Schema identity, fields, panels, feature routes and order | Core 01 §7.4, REQ-01-615–617; Core 03 §2.3A | View contracts, capability resolver and presentation tests | Every emitted schema renders each declared readable panel/action once; concealed and absent differ; saved-view changes cannot override schema config | Keep canonical identities/order. Do not add a public view schema version for visual layout alone. |
| Ordinary authoring and explicit submission | Core 03 §2.3A; design §§12.5,12.7 | Details/editor/draft components; inspector edit browser scenarios | Update and Ctrl/Cmd+Enter only; IME/native picker precedence; detach/resume/discard and captured revision | Preserve useful raw authoring and exact intent; replace misleading completion copy. |
| Mutation HTTP, idempotency and recovery | Core 01 §3.3.5; Core 03 §§3–4 | Feature operation owners and History recovery tests | Duplicate activation, identical uncertain request bytes, late receipt, accepted-write/failed-refresh | No new route or retry policy; UI disclosure never owns attempts. |
| History read and reversal | Core 01 §3.3.4.2; Core 02 §15; Core 03 §§10.1–10.4, REQ-03-139/140 | Materializer, OpenAPI source, presentation mapper, selector/equality code | Required semantic unit families, actor/time, retained entries, opaque selectors, ordering, paging, all legal reversal scopes | Correct missing structured detail; explicitly classify wire changes and cache migration in WS-06. |
| Evidence preview, attachment and access | Core 03 §8; Core 04 §2.0A; design §11 | Timeline Evidence panel, separate Evidence owner contributions | Lifecycle/availability/access combinations, file/drop/paste parity, pending linking, replay and read recovery | Keep independent rights; a visible count never authorizes metadata or bytes. |
| Authorization, stale reads and session boundaries | Core 03 REQ-03-299/100; Core 04 §§1–2; design §§12.7–12.8 | Explicit region models and current owner lifetimes | Immediate concealment of contents/counts/destinations, rejected late reads, incident loss versus account replacement | No new hidden-content cache or generic clear-all policy. |
| Storage, revisions, immutable audit and source projection | Core 02 §15 and source owners | Revisions provider catalog and retained snapshots | Derive detail from retained source facts; do not rewrite committed history, mutate projections to invent history, or expose raw snapshots | No storage migration assumed; prove coverage of historical snapshots before claiming none is needed. |
| WebSocket, saved views and preference persistence | Core collaboration/view owners | Existing inspector integration only | Refresh/retarget and selection continuity | No protocol, saved-view JSON or persistent inspector preference change. |
| Grid boundary, semantic selectors and generated contracts | Repo boundaries; design §§8,14–15 | Source/import rules and UI/view facades | Semantic focus/selector continuity; direct grid imports remain in adapter; authored generation | Change tests tied to retired copy/layout; retain identities with actual consumers. |
| Verification accounting | Adopted harness owner; machine routing under `tools/` | Current task guides and catalogs | Active rows run once, new rows are owned, visual captures reconcile | Catalog coverage is not proof of specification completeness. |

## 5. Gap-by-gap remediation decisions

Classification uses `must_fix`, `should_fix`, `defer`, or `intentional/no_action`.
Priority is delivery priority, not an incident severity. Each gap names its change
areas, rationale, long-term value, migration impact, unresolved risk and binary
validation. Proposed UX benefits remain hypotheses until observed in task review.

### G00 — Proposed observable behavior is not yet a complete owner contract

**Finding:** specification completeness gap in this proposal; `must_fix`, P1.
Current owners already specify field order, explicit editing and state delivery;
they do not adopt this proposed navigation/disclosure design. The audit's broad
slice descriptions are not sufficient defaults or acceptance contracts.

**Remediation / areas:** specification, typed projections where needed, tests and
documentation; WS-00. Update design §§7.3,8.4–8.5,12.4–12.7,14–15 and matching §18
criteria with the decision table in §7. Update Core 01's History transport detail
and Core 03's presentation references only where G10 needs owner clarification.
Use domain terminology without copying requirements into `domain.md`; change that
file only if a vocabulary/navigation reference actually changes. Leave the NLSpec
research artifact unchanged.

**Rationale / benefit:** one observable rule per owner prevents independent
implementations from inventing incompatible focus, disclosure or missing-data
behavior. Keep component factoring intentionally unspecified.

**Compatibility / migration:** specification adoption is a deliberate UX change;
reconcile cross-references and machine projections together. No competing
inspector specification or versioned UI mode. Existing obligations are retained
for their value, not copied wholesale into the new clauses.

**Risk if unresolved:** code/tests become accidental requirements, and later
surfaces repeat incompatible navigation and completion semantics.

**Validation:** each G01–G11 exit traces to an owner or explicit design hypothesis;
all defaults, state transitions, omission cases and error paths in §7 have one
owner; no contradictory duplicate text or prose-dependent executable check.

### G01 — Long inspectors have no persistent section navigation

**Finding:** structural weakness; `should_fix`, P1; usability benefit unmeasured.
The shell renders a continuous stack and the declared list has no shared navigator.

**Remediation / areas:** design specification, implementation, tests, documentation;
WS-02 after WS-00/01. Add persistent Sections disclosure and current-section text.
Use the same admitted semantic descriptors as the body, preserving declared order.
Target the section heading or existing entry control without forcing lazy reads.
Keep sections mounted and one internal scrollport.

**Rationale / benefit:** navigation cost no longer grows with preceding content;
one identity/admission decision serves every schema and prevents concealed
sections from leaking through a separate navigation list.

**Compatibility / migration:** new local navigation replaces scroll-only access;
existing contextual openers and focus registry converge on it. No URL, saved-view,
persistence or grid-vendor change. Remove redundant navigation adapters, not
owner-required shortcuts.

**Risk if unresolved:** growing content becomes harder to find; independent
feature navigation increases coupling. Excess task time is a hypothesis.

**Validation:** every readable declared destination is reachable with at most two
control activations (open, choose; keyboard traversal is not counted as activation).
Passive scrolling never moves focus or requests data. Subject/access changes
remove stale destinations immediately; menu Escape precedes editor/inspector
Escape; supported widths retain visible Close, focus and status strip.

### G02 — Saved Details values have equal visual weight and long travel

**Finding:** structural weakness; `should_fix`, P1. Details stacks every declared
field; Timeline declares 24 fields.

**Remediation / areas:** design, implementation, tests; WS-03. Use aligned
label/value/action rows for short values, full-width narrative rows, responsive
stacking and explicit full-value disclosure. Classify through existing semantic
field metadata or a minimal authored presentation mapping, never display labels.
Give collection summaries a semantic route to existing management regions.

**Rationale / benefit:** field meaning and reading density become consistent
across surfaces without creating a second field registry or turning collections
into scalar mutations. New fields use deterministic defaults.

**Compatibility / migration:** retain declared fields/order, read-only content,
safe rendering and zero/false/empty/null/unloaded distinctions because they carry
meaning. Retire per-surface duplicate field-row layouts. Update layout assertions
and goldens; no field or stored-data migration.

**Risk if unresolved:** adding fields increases scanning effort and repeated
per-surface layout work; aggressive ad hoc hiding could remove required facts.

**Validation:** all readable declared fields occur once and in order; long labels,
URLs and prose remain fully accessible; unrecognized field presentation defaults
to a readable stacked row. Collections navigate only to admitted destinations;
read-only, zero, false, empty and missing states have separate assertions.

### G03 — Workflow and History actions compete with reading

**Finding:** structural weakness; `should_fix`, P1. Workflow wraps equal-weight
buttons; History repeats reversal buttons and leads with row deletion.

**Remediation / areas:** design, implementation, tests, documentation; WS-07/08.
Use an ordered Workflow action list with authored labels and explicit outcomes
(open authoring, review, or navigate). Put event reversal choices inside the
relevant event disclosure; place row delete/restore in a labelled Record actions
group after the event-reading region. Keep one primary affirmative action per
active authoring/review region, with destructive language on destructive actions.

**Rationale / benefit:** reading and action selection have a stable hierarchy;
new declared actions reuse presentation while capability owners remain explicit.

**Compatibility / migration:** migrate current action bindings and tests together;
preserve declared order, identity, legal disabled discovery and each typed shared
reason. No arbitrary More menu, duplicate action registry or global Save control.

**Risk if unresolved:** reviewers must scan repetitive controls and can mistake
row-level and event-level scope; wrong-action frequency is a hypothesis to test.

**Validation:** every action dispatches its canonical route exactly once, only
when currently admitted. Each reversal shows its real scope, including changes
to other records. Destructive confirmation starts at a safe control and invalidates
on owner-defined changes. Disabled reasons are deduplicated by identity/parameters.

### G04 — History reads like raw implementation output

**Finding:** confirmed rendering limitation plus structural weakness; `must_fix`
for required data display, `should_fix` for visual hierarchy, P1. The renderer uses
full ISO text; the mapper forwards summary/operation and no structured unit detail.
Its caller supplies no readable actor label.

**Remediation / areas:** design, implementation, tests; WS-07, dependent on G10.
Render concise authorized descriptions from typed semantic facts, absolute time
with explicit timezone, and visible attribution. Use an authorized actor label
when supplied; otherwise show the attributed actor ID as such, not an invented
person name. Disclosure exposes complete required semantic units, exact original
timestamp and technical identifiers. Never parse prose to recover semantics or
infer action eligibility.

**Rationale / benefit:** reviewers can tell what happened without decoding storage
names; new unit families gain an explicit mapping instead of string heuristics.

**Compatibility / migration:** change display copy/time formatting and associated
snapshots. Keep exact timestamp and opaque identities available. An unavailable
friendly label does not remove mandatory actor attribution. Historical missing
semantics remain a G10 completion failure, not a permanent cosmetic fallback.

**Risk if unresolved:** incomplete audit understanding and brittle text-derived
behavior as new source families arrive.

**Validation:** field, link, mention, tag, Evidence association and capture-state
changes render truthfully; absent optional names/details remain explicit. Timezone,
invalid-input handling and exact-time access are deterministic. Event expansion
preserves paging, order, opaque targets, requested-change highlighting and focus.

### G05 — Editing presents ambiguous completion choices

**Finding:** structural weakness; `should_fix`, P1; actual user confusion is a
hypothesis. Update and Done editing are produced in different contributors.

**Remediation / areas:** design, implementation, tests, documentation; WS-03.
Compose one Unsaved change region with accepted value still visible. Group Update
and Close editor; explain that closing retains unfinished work. Keep Clear at the
value control and Discard draft distinct. Expose retained-work status and eligible
Resume/Discard on the original field even when no editor is attached. Derive cues
from the canonical draft store, not another retention cache.

**Rationale / benefit:** command placement makes saving versus detachment explicit
while preserving source-specific validation and exact authoring identity.

**Compatibility / migration:** replace misleading Done editing copy and migrate
its tests without a label alias. Do not change explicit submission keys, null/clear
intent, account/incident scope or grid editing. No durable persistence is added.

**Risk if unresolved:** analysts may think detached work was saved or abandon
retained work; duplicate UI draft state would create correctness bugs.

**Validation:** Update and Ctrl/Cmd+Enter submit only the captured revision;
Close, Escape, blur, Tab and section navigation never submit/discard. Late success
clears only captured work; newer drafts survive. Retarget, stale field dependency,
read-only and reauthorization cases expose correct Resume/review/Discard behavior.
Global Saved cannot imply an unfinished local draft was submitted.

### G06 — Inspector controls do not consistently consume design roles

**Finding:** confirmed token-use defects; `must_fix`, P1. Title/section styles omit
full typography roles; buttons inherit font and use general spacing/radius;
Timeline/Indicator inputs have competing padding definitions.

**Remediation / areas:** implementation and tests; specification/projections only
for missing variants that prove necessary; documentation; WS-01. Apply complete title,
section, body, metadata, button and technical roles and existing component states.
Use shared field/group/action/feedback presentation. Adopt a compact variant only
if existing tokens cannot satisfy an observable requirement after layout changes.

**Rationale / benefit:** one styling decision keeps density, accessible focus and
future token changes coherent. Remove accidental native fieldset outlines while
retaining legends and group semantics.

**Compatibility / migration:** visual change across actual consumers; inventory
imports before editing shared styles. Move inspector callers first and delete
unused variants in the same slice. Keep grid-specific geometry separate. No new
theme, font, icon dependency or package is justified.

**Risk if unresolved:** token updates stop propagating reliably, and each future
workflow introduces another styling dialect or inaccessible compact control.

**Validation:** computed roles and component states match authored projections;
all supported densities, disabled/pending/destructive states, long labels, zoom
and focus remain usable. Migrated callers have no obsolete style path. Generated
facades are rebuilt from authored inputs if those inputs change.

### G07 — Selected relationships consume most of the viewport

**Finding:** structural weakness; `should_fix`, P2. One mention expands status,
target, resolution, filter, picker and operation controls into a tall form.

**Remediation / areas:** design, implementation, tests; WS-04. Use compact item
summaries with raw mention, resolution state and authorized target. Expand selected
correction controls inside the original source group; keep filter/read status,
retry and paging within the target chooser presentation. Preserve unfinished or
uncertain-operation cues when its controls are collapsed.

**Rationale / benefit:** source context remains legible with less repeated form
chrome; the existing mention owner remains the single source of operation truth.

**Compatibility / migration:** presentation-only cutover. Keep loaded-target scope,
selected identities across pages, raw versus canonical links, and the exact
Dismissed in this session qualification. No automatic enumeration or entity deletion.

**Risk if unresolved:** larger relationship collections become difficult to inspect;
local UI shortcuts can accidentally imply complete search or erase pending work.

**Validation:** selected controls remain beside the correct field/item after paging
and retargeting; failed/partial reads do not imply no targets or unresolve a known
mention. Resolve, correct, dismiss, restore and create-then-link preserve their
existing permissions, reviews, exact replay, receipts and local outcomes.

### G08 — Specialized authoring lacks a consistent visual and mode hierarchy

**Finding:** structural weakness; `should_fix`, P2. Assessment always supplies a
creation title even with a saved subject; specialized forms use different group
and control treatments.

**Remediation / areas:** design, implementation, tests, documentation; WS-09.
Use the saved subject for the record header and put the specialized operation title
in its attached region. No-subject creation gets an explicit creation header and
admitted context. Reuse field/group/action/feedback presentation while leaving
submission, review and lifecycle with each feature.

**Rationale / benefit:** users can distinguish record identity, operation and saved
outcome. New forms reuse a consistent presentation without inheriting another
feature's validation, authoring lifetime or submission semantics.

**Compatibility / migration:** migrate Assessment, Indicator, Note, coordination,
Task/Decision and related-Evidence callers by family. Keep append-only Assessment
and Indicator lifecycle rules because their domain meaning remains valuable.
Retire duplicate spacing, fieldset and action-row treatments as callers migrate;
no universal form engine or panel-global Save is introduced.

**Risk if unresolved:** creation can be mistaken for editing, and every new workflow
adds another visual and interaction dialect that must be maintained separately.

**Validation:** saved-record and no-subject headers differ correctly. Every declared
form retains explicit source identity, required minima, review and recovery.
Source replacement/clearing, duplicate submit, detached late outcomes and
permission loss are tested for each feature family, not inferred from shared CSS.

### G09 — Evidence presentation emphasizes mechanics over accepted information

**Finding:** structural weakness and bounded design opportunity; `should_fix`, P2.
The Timeline contribution presents an attachment count and file chooser, with
recovery above them; other Evidence-owned contributions supply access behavior.

**Remediation / areas:** design, implementation, tests, documentation; WS-05.
Separate authorized accepted information, independently admitted preview/download
controls, and attachment progress/recovery. Use a compact Attach file entry with a
named drop/paste region and keyboard alternative. Render item rows only from
already-authorized owner data; otherwise describe the observed count/scope honestly.

**Rationale / benefit:** inspecting evidence and attaching a file become distinct
activities. Independent access and upload/link outcomes stay visible, so the
presentation can accommodate future owner-provided metadata without owning reads
or inferring capabilities from counts.

**Compatibility / migration:** chooser, drop and paste retain current file-type,
size, lifecycle, authorization and upload contracts. No richer enumeration API or
storage migration is required. New Evidence item browsing is DEFERRED outside this
plan unless existing authorized data already supports its presentation.

**Risk if unresolved:** upload, Evidence creation and source linking can be mistaken
for one completed transaction; a richer-looking UI could falsely imply available
metadata, complete totals or permission to preview/download.

**Validation:** counts, preview/download rights, unavailable/quarantined/deleted
states, uncertain upload, saved-but-unlinked and saved-but-unrefreshed outcomes are
distinct. Chooser/drop/paste use the same admitted owner command. Recovery remains
source-local, preserves exact uncertain attempts, and reads after acknowledged
writes. Preview stays in the workbook and no protected metadata or bytes leak.

### G10 — History's transport and source projection cannot express required detail

**Finding:** confirmed end-to-end gap; `must_fix`, P1. Core 03 REQ-03-139 requires
unit-level change detail. The authored OpenAPI allows arbitrary unit objects;
`history_materializer.go` emits target metadata and before/after-presence flags,
not field or relationship deltas. The frontend then discards units entirely.

**Remediation / areas:** specification clarification, authored protocol projections,
backend and frontend adapters, source-owner contributions, tests and documentation;
WS-00/06/07. Define a closed, versioned semantic unit contract with explicit variants
for the required families and safe value/availability rules. Source owners derive
units from retained authoritative before/after facts through pure contributions;
Revisions validates/catalogs/composes them. No central source-table switch, browser
snapshot decoder, untyped map spreading, or projection-derived history fallback.

Define stable unit identity/order, public field or relation identities, operation,
value omission/null/redaction/unavailable meanings, actor provenance and unknown
version handling in the owner. Keep presentation text distinct from selector
identity. Include imported attribution in the producer/schema/consumer parity
review: `history_model.go` can emit `source_actor_id`, while the inspected closed
OpenAPI item does not declare it. Resolve against its adopted import owner rather
than silently accepting arbitrary additional properties.

**Rationale / benefit:** semantic History becomes reconstructable, testable and
extensible without coupling source storage formats to every client. Narrow typed
facts reduce accidental disclosure compared with exposing whole snapshots.

**Compatibility / migration:** use the approved coordinated server/browser
cutover under the pending OpenAPI 2.0.0 boundary, keeping `/api/v1` and the
immutable released baseline. No legacy response adapter is supported. Preserve
logical item/entry identities and canonical retained audit data. Core 02
REQ-02-265 prohibits schema-less readers, member-presence inference and backfill;
unsupported old databases remain subject to its operator-controlled reset policy.
This implementation does not reset a database. Missing detail in supported
canonical history blocks closure; never invent values or substitute current rows.

History content equality currently treats `diff_summary` as immutable. Changing
its representation during a retained browsing session needs a read-cache generation
cutover/reload policy that leaves captured mutation attempts and receipts intact.
Test upgrade and rollback; do not use a cache reset to bypass receipt recovery.
No storage migration is planned until retained-data inventory proves one necessary.

**Risk if unresolved:** cosmetic History can still violate required review behavior;
new sources multiply ambiguous payloads and unsafe ad hoc decoding.

**Validation:** required unit families are produced from representative current and
retained snapshots; source catalogs reject missing/duplicate/invalid contributions;
public schema, generated types and renderer agree; hidden data and raw storage
identifiers/bytes are absent. Paging/order/selectors and rollback semantics remain
stable. Compatibility report and retained-session upgrade/rollback evidence are
recorded, with every supported consumer accounted for.

The WS-00 contract review must close at least this semantic mapping; wire member
names and source-private decoding stay with their owners:

| Change family | Minimum truthful public detail | Semantic owner |
| --- | --- | --- |
| Field change | Public field identity, operation, authorized before/after value states; absence distinct from null, false, zero and empty text | Source record owner using retained snapshots |
| Link change | Public relationship identity/type, add/remove/replace and authorized endpoint references | Links or the source owner of the declared relationship |
| Mention change | Stable public mention reference, raw mention when visible, prior/result resolution state and authorized target facts | Entities mention owner |
| Tag change | Add/remove and authorized tag identity/value | Existing tag/source owner |
| Evidence association | Attach/detach and authorized Evidence/source references; no object bytes, upload tokens or preview handles | Evidence association owner |
| Capture-state transition | Prior/result capture state and admitted operation | Timeline owner |
| Record/merge and other source-owned events | Public record operation, merge participants when authorized, and source-defined semantic deltas for Assessment/Indicator and other supported histories | Respective source owner; no reinterpretation as arbitrary scalar patches |

Each unit needs deterministic identity/order within the stable logical item and
an explicit representation version. An omitted value is not automatically null;
unavailable or withheld detail uses a typed state only when the owner permits
disclosing its existence. Whole-item concealment exposes no such hint. Keep
presentation detail separate from the opaque rollback target already supplied by
History. This is a display projection of immutable source facts, not a second
event store or event-sourcing redesign.

### G11 — Validation evidence does not yet prove redesign completeness

**Finding:** confirmed coverage gap and untested usability hypotheses; `must_fix`,
P1 for delivery completion. The base visual scenario asserts Details content but
scrolls to Relationships before its first inspector screenshot. The original audit
sampled specialized workflows, not all seventeen configurations or accessibility states.

**Remediation / areas:** tests, fixture/catalog projections, documentation and
manual task review; WS-10 plus relevant tests in each earlier workstream. Add
first-viewport saved Details, attached edit and retained-draft cases; exercise all
schema configurations and the state/geometry matrix in §8. Compare representative
reading, editing, relationship, History and creation tasks.

**Rationale / benefit:** evidence tests the intended user outcomes and dangerous
transitions rather than preserving implementation-shaped snapshots. Future sources
inherit a finite coverage contract.

**Compatibility / migration:** update authored fixture declarations/catalog rows
and generated evidence only through public targets. Retire superseded captures only
after reconciliation proves zero consumers; keep useful existing scenarios.

**Risk if unresolved:** a visually improved happy path can ship with hidden focus,
authorization, disclosure or retained-work regressions; sampled screenshots may be
misreported as complete acceptance.

**Validation:** every applicable acceptance row has fresh evidence, every exclusion
has a scope/owner rationale, and no required row is blocked. No golden update hides
a functional failure. Task review records outcomes and remaining hypotheses without
claiming measured improvement that was not observed.

## 6. Workstreams, dependencies and phase gates

Each WS is one separately reviewable implementation slice. Workstreams are
serial in the default delivery order below, even where minimum technical
dependencies would allow overlap. Before starting the next workstream, update §9
with changed paths, adopted decisions, commands/run roots, risks, retirement,
compatibility and its observed exit. `DONE` means the slice was executed and its
exit passed; writing its plan never qualifies. Allowed work statuses are `TODO`,
`IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

| WS / phase | Class / technical prerequisites | Goal / gaps | Next workstream | Exit and checkpoint |
| --- | --- | --- | --- | --- |
| WS-00 / phase A | root / none | Owner cleanup and complete decisions; G00, G10 | WS-01 | Adopted text resolves every §7 decision, owner references and acceptance mapping; record any blocked owner issue before dependent work. |
| WS-01 / phase B | chain / WS-00 | Typed presentation projections and shared visual foundation; G06 | WS-02 | Authored roles/projections and migrated common controls agree; focused tests and token review pass; obsolete variants removed. |
| WS-02 / phase B | chain / WS-01 | Shared frame and section navigation; G01 | WS-03 | One admitted descriptor sequence drives navigation/body, focus and concealment checks pass in adjacent/overlay layouts. |
| WS-03 / phase C | chain / WS-02 | Saved Details and ordinary edit hierarchy; G02, G05 | WS-04 | Read/edit/close/resume/discard and revision-bound outcomes work across Timeline, Generic and Entity consumers. |
| WS-04 / phase C | chain / WS-02 | Relationship summary/correction disclosure; G07 | WS-05 | All mention transitions and bounded target recovery remain correct; hidden controls leave actionable unresolved-status cues. |
| WS-05 / phase C | chain / WS-02 | Evidence information and attachment hierarchy; G09 | WS-06 | Accepted data, independently authorized access and attachment/link/recovery states remain distinguishable. |
| WS-06 / phase D | chain / WS-00 | Semantic History contract, source contributions and compatibility cutover; G10 | WS-07 | Required units complete across source families, retained data and schema/generation/security/compatibility checks; no unresolved migration. |
| WS-07 / phase D | chain / WS-02, WS-06 | Readable History and deliberate corrective actions; G03, G04 | WS-08 | Mandatory attribution/time/units displayed, exact scopes and opaque targets preserved, History recovery and focus pass. |
| WS-08 / phase D | chain / WS-02 | Workflow action catalog presentation; G03 | WS-09 | Every declared action is discoverable once, ordered and correctly described/admitted; no hidden compatibility alias. |
| WS-09 / phase D | chain / WS-03, WS-05, WS-07, WS-08 | Specialized form and mode consistency; G08 | WS-10 | All feature families migrated to shared visuals with source identity, independent submission and recovery intact. |
| WS-10 / phase E | chain / WS-00 through WS-09 | Validation and handoff completion; G11 and all gaps | none | §12 product criteria met, acceptance/evidence recorded, retirement complete, rollout/rollback and next maintainer documented. |

| Phase | Main risk | Exit criteria | Stop/rollback rule |
| --- | --- | --- | --- |
| A — Specification cleanup | Design proposal becomes accidental authority or weakens required history | Complete owner decision tables and traceable binary criteria; no owner contradiction in dependent scope | Keep proposals in tracker until adopted; repair contradiction before dependent implementation. |
| B — Shared foundation | New navigation duplicates admission or changes mounts/focus | WS-01/02 exits and shared-consumer coverage pass | Revert coherent presentation/projection slice; do not keep two shell implementations behind an indefinite flag. |
| C — Reading and local authoring | Disclosure loses identity, masks draft or conflates Evidence rights | WS-03/04/05 exits and local failure/permission/recovery matrix pass | Revert affected visual composition; retained operation owners and receipts remain authoritative. |
| D — History and workflow | Payload disclosure, incompatible clients, incorrect rollback scope or form mode | WS-06–09 exits, explicit compatibility classification and source-family coverage pass | Roll back frontend/server/projection cutover coherently; preserve immutable audit and in-flight recovery. |
| E — Validation and handoff | Baseline failure or golden update is mistaken for completion | Every applicable acceptance row is evidenced; all mandatory WS are DONE | Do not mark effort complete on partial runs; record failing target/run root, cause and remaining work. |

## 7. Proposed slice implementation contracts

### Decisions to adopt in WS-00

The user approved these defaults. WS-00 adopts them in design §12.7 and Core 01
REQ-01-052A, Core 02 REQ-02-218/265 and Core 03 REQ-03-139. The explicit
implementation decisions above supersede historical conditional compatibility
recommendations. The History representation is `cartulary.history_diff.v1`,
canonical snapshots only, no partial-detail success and no legacy adapter.

| Concern / owner | Complete recommended rule |
| --- | --- |
| Navigation destinations / design §12.7 with Core 03 §2.3A | Derive from the active config and the same subject/access admission used by rendering. Include readable unrequested/failed sections; omit undeclared, concealed and unavailable-by-subject sections. Missing required implementation is an error, not omission. No subject: only explicitly admitted creation sections. Zero destinations: no Sections control, retain truthful no-row/creation status. |
| Active section / design §§7.3,12.7 | An explicit opener destination wins if admitted. Reopening the same subject preserves active panel under Core 03; a new subject/schema without explicit destination starts at first admitted panel. If active section disappears, use first remaining admitted panel. Passive scroll selects the last heading at/above the usable body top, or first visible heading when none precedes it; at body end select last admitted section. No automatic focus or fetch. |
| Section selection / design §§8.4–8.5,12.6 | Use a labelled disclosure containing ordinary semantic navigation controls, not tab roles. Enter/Space activates; Tab visits destinations. Selection closes disclosure, scrolls only inspector body, and focuses the heading or existing owner entry control. Unrequested History targets Open history without issuing a read. Escape closes the topmost popup before editor/inspector handling. Escape dismissal restores trigger focus; pointer outside dismissal preserves destination focus. |
| Transient lifetime / design §12.7 | Navigation does not unmount contributions; same-subject section selection retains an attached ordinary editor without submitting it. Retarget clears old-subject presentation and menu state. Scope disclosure selection to schema/record/field or item identity; presentation invalidation must not retire raw authoring or captured operations. On concealment remove content, counts, labels and navigation/focus destinations together. Restore focus only when the concealed region owned it, to a safe surviving control. |
| Geometry / design §§7.3,14 | Retain 420px default, 360px minimum, `min(560px, 45vw)` maximum and existing overlay bands/clamps. One body scrollport and bounded two-line title with explicit full-title access. Sections/Close remain visible. No persistent inspector preference. |
| Field classification / design §12.7 | Short scalar values use aligned rows; declared multiline/narrative values use full-width rows; unknown presentation stacks safely. Never omit/reorder fields for layout. Prefer wrapping; for lengthy narrative use a six-rendered-line preview with explicit Show full value / Show less controls. Project the line limit through the authored design contract if adopted. No truncation of editable control, retained-error or required recovery content. |
| Editor actions / design §§12.4–12.7 | Unsaved change region keeps accepted value visible. Owner supplies Update, Close editor, value-level Clear and distinct Discard bindings; Close explains retention. Exactly one ordinary editor attaches. Detached cues come from draft owner. Specialized workflows keep their own submit/review/detach policy. |
| Relationship disclosure / design §12.7 | Summaries always show raw mention and semantic resolution state; authorized target only when known. One selected correction region per current source collection binding; read/paging controls are local to picker. Collapsing hides controls, not pending/uncertain/retained-status or owner-required recovery entry. |
| History / Core 01 §3.3.4.2, Core 03 §10, design §§12.7,14 | One row per stable logical item; server order/paging. Display actor label if authorized, otherwise attributed ID, plus operation/description and absolute UTC time with numeric offset and unambiguous date. Exact committed timestamp and required typed units are accessible in event detail. No relative-only time, inferred diff or text-derived action target. |
| History wire/retained data / Core 01–02 | Core 01 REQ-01-052A owns closed units and explicit absent/null/present values. Validate canonical snapshots under REQ-02-265; reject unsupported payloads locally. No partial-detail success, schema-less translation or legacy adapter. |
| Workflow and creation / design §§12.4,12.7 | Show declared actions in order with outcome description. Saved subject remains the record title; local attached operation title identifies append/create/transition. Creation without a subject has explicit mode and source context. Primary treatment belongs to the active local affirmative decision; no panel-global Save. |
| Evidence / design §§11,12.7 | Separate accepted metadata, available access commands and attachment work. Counts never become item lists. File chooser, drop and file paste share owner admission/validation, prevent accidental duplicate activation, and have keyboard parity. Lifecycle, availability and access remain independent. |
| Unknown/additive facts / respective owner | Reuse existing capability omission rules; don't interpret unknown feature labels. History unknown unit/version handling is explicitly versioned and safe: no guessed value or action, preserve readable known authorized facts only when the adopted decoder allows it. Unsupported mandatory payload versions fail the read locally. |

### Slice delivery details

File groups refer to §2; this table is the concrete implementation handoff, not
permission to make those changes in the planning task.

| Slice | Likely authored files / boundary | Characterization and changes | Verification / completion | Rollback |
| --- | --- | --- | --- | --- |
| WS-00 | Design owner and targeted Core 01/03 History clauses; domain navigation only if needed | Resolve §7 decisions; map current obligations versus new design behavior; specify unit semantics and retained-data policy without mirroring helpers | Human owner review; Markdown lint; trace all gaps to exits; no machine product check reads prose | Revert proposed owner changes before implementation, or coordinate with dependent slices after adoption |
| WS-01 | Shared actions/form styles, shell typography, affected Timeline/Indicator style callers; authored design contracts if needed | Use complete roles; consolidate actual variants; retain semantic legends; characterize computed state/density rather than literal style object shape | Focused `package.ui`/`web.workbook`, frontend type/import checks; generation/drift if inputs changed; all common callers migrated | Revert authored and generated projections plus callers together |
| WS-02 | Shell, declared list, feature shell bindings, existing semantic focus integration; layout only as needed | Introduce one admitted section descriptor seam; keep body mounts; test menu/scroll/retarget/access/overlay transitions | Focused presentation tests plus browser focus/geometry; all G01 exits | Revert navigation and its bindings, keep operation owners unchanged |
| WS-03 | Details, editor, feedback, Timeline/Generic/Entity adapters | Explicit editor content/action slots; aligned fields/full value; owner-derived retained cues; remove duplicate Update placement | Draft/revision and Details rows; inspector edit browser/a11y; G02/G05 exits | Revert presentation/adapters as one unit, preserving draft identity/data |
| WS-04 | Mention panels/actions and candidate presentation integration | Compact summaries; selected correction disclosure; focus restore/reveal; qualified bounded-target status; no second operation store | Mention transition/reconciliation suites and browser relationship scenarios; G07 exit | Revert presentation with same owner attempts and semantic item IDs |
| WS-05 | Timeline Evidence and affected existing Evidence presentations | Distinct accepted/read/access/attachment regions; chooser/drop/paste admission; no new metadata enumeration | Evidence panel and attachment/recovery tests; public browser/a11y/visual cases; G09 exit | Revert visual grouping; never clear pending upload/link state as cleanup |
| WS-06 | Revisions OpenAPI owner, typed unit projection, History model/materializer/service, source provider contributions, revisionassembly and browser History adapter | Enumerate source families; add pure source-owned semantic projectors; close schema; preserve selectors/order; review imported attribution parity; test retained-data and session-generation cutover | Revisions/source-owner slices; protocol generation/drift; schema and OpenAPI compatibility; service-backed history/security cases; G10 exit | Coordinated server/client/schema revert; avoid data rewrite; no legacy adapter; deploy and roll back the bundled server/browser together |
| WS-07 | History event mapper/renderer/actions and focus binding | Visible actor/time, complete typed unit disclosure, event-local rollback, Record actions group; remove redundant operation labels | History mapping/browsing/recovery and browser reversal/a11y; G03 History/G04 exits | Revert UI/protocol consumer coherently with WS-06 if required |
| WS-08 | Contextual actions renderer and authored action presentation metadata where needed | Ordered action rows with result descriptions, existing reason association and canonical command binding | Capability/order/reason and browser action dispatch checks; G03 Workflow exit | Revert action layout and labels with callers/tests, no aliases |
| WS-09 | Assessment/Indicator, Note, coordination, Task/Decision, related-Evidence and fallback forms | Migrate family by family; record consumed shared styles; distinguish selected record from operation/creation; delete dead local styles | Each changed feature's authoring/recovery tests and representative browser flow; §8 schema audit complete | Revert each completed family coherently; no mixed duplicate form authority |
| WS-10 | Tests, authored routing/fixture registry, reconciled goldens, affected source guides and this tracker | Add missing evidence, reconcile owner acceptance, inspect every changed golden, document migration/retirement and residual limitations | All §8 gates and §12 product exits; final tracker/handoff update is the last required slice | Withhold cutover on blockers; record exact successful deployment-compatible rollback bundle |

A style export shared with non-inspector callers is retained only until those
callers are intentionally migrated or it remains a justified grid-specific role.
Do not delete it merely to satisfy a textual cleanup count. Other scope additions
require a separately identified gap/owner/exit here before work starts.

## 8. Validation plan

### Discovered public command routes (planning record)

This subsection records the pre-implementation selection, not current execution
status. WS-00–09 checkpoints and the WS-10 final evidence below supersede its
"Not run" entries.

Task guides for `web.workbook`, `module.workbook`, `package.ui` and
`module.revisions` ran successfully in this planning pass. `explain-test-owner`
confirmed 288 `web.workbook` rows and nine `package.ui` rows at this baseline.
These are routing observations, not tests or permanent inventory requirements.

| Layer | Public command | Scope / when required | Planning execution |
| --- | --- | --- | --- |
| Routing | `make task-guide ROLE=module-author OWNER=<active-owner>`; `make explain-test-owner OWNER=<active-owner>` | Refresh at slice entry; use authored catalog to select changed behavior | Four task guides and two owner explanations passed |
| Narrow frontend | `make test-slice OWNER=web.workbook ROWS=<selected-active-rows>` | Each presentation slice; add new behavior rows to authored catalog before relying on them | Not run; implementation pending |
| UI projections/selectors | `make test-slice OWNER=package.ui` | WS-01 and any later UI projection/selector change | Not run |
| Revisions | `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions ROWS=<selected-active-rows>` | WS-06/07, plus each changed source owner's current guide | Not run |
| Browser | `make service-backed-test-slice OWNER=module.workbook ROWS=<selected-active-rows>` | Changed stateful/webserver-backed scenarios per current routing | Not run |
| Static | `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome` | Authored frontend changes and shared boundary migration | Not run |
| Generation | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Only when corresponding authored contracts/catalogs change; include all generated roots declared by policy | Not run |
| HTTP compatibility | `make openapi-compatibility-check` | WS-06 against immutable release history, before choosing cutover | Not run |
| Finalization | `make agent-finalize` | Before broader implementation verification; pass RESULTS_DIR only for an actual qualifying successful full warm run | Skipped for documentation-only planning; it includes mutating generated-structure refresh |
| Accessibility / visuals | `make browser-e2e-a11y`; `make browser-e2e-visual` | Shared UI changes and WS-10 | Not run |
| Golden refresh | `make browser-e2e-visual-update`, then ordinary visual validation | Only after ordinary-run reconciliation and functional correctness; inspect every changed image | Not run |
| Broader end-to-end | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make test-fast` | WS-10 only where changed owner breadth or unresolved risk requires beyond focused rows | Not run; no automatic full `make check` requirement for planning |
| Documentation | `make lint-markdown`; `git diff --check` | This tracker and later documentation changes | Pending P-03 |

Current relevant row examples, revalidate before execution:

- `web.workbook.regression.inspector_persistent_context`,
  `inspector_explicit_panel_states`, `inspector_grouped_permission_reasons`,
  `inspector_draft_binding`, `inspector_draft_lifetime`,
  `inspector_field_feedback_identity`, `inspector_timeline_explicit_details`,
  `history_inspector_recovery`, `history_recovery_characterization` all use the
  `web.workbook.regression.` prefix (apply it to each abbreviated name).
- `module.workbook.browser.inspector_subject_retention`,
  `module.workbook.browser.inspector_detached_refresh`,
  `module.workbook.browser.inspector_exact_replay` and
  `module.workbook.accessibility.inspector_edit_recovery`.
- History browser routing also belongs to `module.revisions`, including
  `module.revisions.browser.verify_history_and_rollback_preview_action_use_p_4201ff80a8`.
  Do not assume every inspector browser row belongs to `module.workbook`.

These existing rows do not yet assert every proposed behavior. Add targeted tests
for navigation, field disclosure, actor/unit presentation, source projectors and
specialized modes; avoid tests that merely mirror helper decomposition.

### Coverage and acceptance matrix

| Axis | Required cases / outcome | Evidence responsibility |
| --- | --- | --- |
| Schema coverage | Timeline; Hosts; Identities; Evidence; Notes; Indicators; Assessments; Task Requests; Decisions; Parties; Communications Log; Handoff; Status Review; Lesson; Findings; Investigative Queries; Forensic Keywords | Composition/capability table covers all 17 configurations. Optional surfaces are tested when implemented, otherwise documented N/A under Core 01; no silent omission. |
| Surface families | Direct behavioral review of Timeline, Entity, Generic, Assessment and Indicator compositions; each specialized authoring family has its own lifetime/recovery check | WS-03–09; tests can share fixtures, not substitute one owner's semantics for another's |
| Read state | Unrequested, initial loading, explicit empty/populated, refreshing with retained data, stale failure, unavailable; independently failing sibling regions | No invented absence, automatic read-all or disabled unrelated actions |
| Authority | Viewer/editor/reviewer/admin where relevant; open/closed incident; concealed data; incident-access loss, same-account recovery and account replacement | Contents, counts, navigation, feedback and late-result admission agree with current authority |
| Authoring/operation | Pristine, dirty, invalid, detached, review required, uncertain, accepted but unrefreshed, saved but unlinked; duplicate invocation and late completion | Captured identity/revision/request bytes stay with existing owners; recovery does not resend accepted writes |
| Geometry | 1280×720, 1024×720, 768×640, below-minimum containment; 360/420/maximum panel widths where admitted; vertical resize; all supported densities; 200% zoom; long text and text-spacing override; reduced motion | No document overflow, obscured focus or lost Close/status-strip access; compare actual owner fixture parameters |
| Keyboard/accessibility | Open/choose/dismiss sections, editor submit/detach, picker nesting, event disclosure, review/confirm/cancel, return to original field/grid; screen-reader names/announcements | Preserve semantic focus identity; one announcement per admitted transition/outcome; non-color state and supported contrast |
| History semantics | Each required unit family, current and retained snapshots, imported attribution, unknown/malformed version, concealed target/value, all advertised rollback scopes, pagination and eligibility change | Source-owner and Revisions integration tests plus production-renderer browser tests |
| Visual evidence | First-viewport saved Details, editing, retained draft, selected relationship, Evidence/recovery, History units/reversal, Workflow/specialized creation, narrow/read-only/error | Follow visual maintenance guide; original base-inspector screenshot did not prove first-viewport Details |
| Boundary/security | No presentation imports of request/mutation owners, no direct grid vendor imports, no raw history snapshots/object bytes/handles in generic diff, no prose-consuming executable tools | Source/import check, payload fixtures and human owner-to-projection review |

Assess every digest row A001–A027 in the implementation handoff, with actual
`PASS` evidence, `N/A` rationale, or `BLOCKED` reason. These are acceptance
outcomes, not work-item statuses. A004–A011, A014, A016–A020, A022–A026 are central;
A012–A013 and A015 apply to exposed mutation/recovery paths. A021 requires evidence
of grid continuity for shared layout changes; performance measurement is needed
only if virtualization/geometry behavior changes or a regression is suspected.
A001–A003 and A027 cover authority, repository state and handoff. Nothing is
silently dropped because this began as a visual audit.

For comparative task review, use representative records and the same five tasks:
locate Evidence; edit a scalar, close and resume; resolve/correct a mention;
inspect a History change and reversal scope; create a related task. Record task
completion, incorrect actions, scroll travel and Saved/Update/Close interpretation
for baseline and candidate. Required task paths must be completable without an
unintended mutation or inaccessible control. Record qualitative results honestly;
no invented time-saving threshold or conformance claim. Lack of participants must
be recorded as a remaining usability hypothesis, not an automated usability pass.

## 9. Top-level implementation tracker

Implementation is authorized. Update the row and
append its evidence to §10 after completing each workstream, before the next starts.
If the exit fails, keep it `IN_PROGRESS` or `BLOCKED` with a concrete reason.

| ID | Work item | Status | Technical dependencies | Evidence to attach | Binary exit |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Specification cleanup and decision closure | DONE | none | Design §12.7 and D-AC-082–086; Core 01 REQ-01-052A, Core 02 REQ-02-218, Core 03 REQ-03-139; lint and whitespace pass | Approved defaults and canonical-only History adopted without contradiction |
| WS-01 | Shared roles and authored projections | DONE | WS-00 | Shared roles and generated narrative limit; focused/UI/type/import/drift checks passed | G06 common-control migration complete; integrated visual evidence follows in WS-10 |
| WS-02 | Record frame and section navigation | DONE | WS-01 | Unit, browser, type/import/lint evidence in §10 | G01 navigation/focus exit passed; integrated matrix in WS-10 |
| WS-03 | Details and ordinary edit hierarchy | DONE | WS-02 | Saved reading, retained draft, browser and accessibility evidence in §10 | G02/G05 behavioral exits passed; integrated visual matrix remains WS-10 |
| WS-04 | Relationship summaries and correction | DONE | WS-02 | Source-local unit/browser evidence in §10 | G07 behavioral exits passed |
| WS-05 | Evidence reading and attachment | DONE | WS-02 | Access/lifecycle and attachment/link/recovery evidence | G09 complete |
| WS-06 | History semantic data contract and cutover | DONE | WS-00 | Source-family inventory, unit fixtures, compatibility and retained-data report | G10 complete |
| WS-07 | History reading and corrective actions | DONE | WS-02, WS-06 | Unit/actor/time and reversal/focus evidence | G04 and G03 History portion complete |
| WS-08 | Workflow action hierarchy | DONE | WS-02 | Declared action/order/reason/dispatch evidence | G03 Workflow portion complete |
| WS-09 | Specialized mode and form migration | DONE | WS-03, WS-05, WS-07, WS-08 | Per-family migration and source/receipt evidence | G08 complete; dead alternatives removed |
| WS-10 | Validation and handoff completion | DONE | WS-00–09 | Final 708/708 fast, all required browser scenarios, reviewed 255-capture reconciliation and two ordinary 12/12 visual runs; full evidence in validation handoff | G11 and all §12 exits complete; A001–A027 PASS; paired bundles and limitations recorded |
| O-01 | New rich Evidence enumeration/read capability | DEFERRED | Separate adopted owner scope | None; not required by this plan | Do not silently add to WS-05 or count as missing presentation work |

## 10. Session handoff log

### Implementation execution checkpoints

| Slice | Status / evidence | Changed boundary / next action |
| --- | --- | --- |
| WS-10 exit | DONE; 2026-09-21 07:54 UTC. Final fast 708/708 `20260921T073129Z-p97071`; final public visual update 12/12 `20260921T074355Z-p30667`, all changed images reviewed, then two fresh ordinary visual runs 12/12 each `20260921T074928Z-p69256/-p69271`. Final Markdown `20260921T075527Z-p42495`, whitespace, static/contract/generation checks and paired builds pass. Exact full/focused browser results and resolved failures are retained in the validation handoff. | G11 and overall effort complete; A001–A027 PASS with explicit optional-feature N/A rationale. Fixed canonical null/empty preservation, saved-text copy and access/presentation scope separation during integrated validation. All 358 changed paths inventoried, 151 images reviewed, 255 captures/mappings reconciled with no orphan/missing/ambiguous entries. Removed superseded presentations; retained exports have named consumers. Verified baseline/candidate 73-file bundles, coordinated rollout/rollback, skipped retained-run maintenance (RESULTS_DIR unset), and unavailable participant evidence recorded. No deployment, reset, backfill, data rewrite, legacy adapter or forced reload. This is the mandatory final slice; no later workstream remains. |
| WS-10 entry | IN_PROGRESS; WS-09 DONE checkpoint saved before entry | Complete integrated owner checks, browser/accessibility/visual reconciliation and review, two ordinary visual confirmations, acceptance A001–A027, retirement/cutover/rollback and final handoff. Run agent-finalize before broader final verification; retained-run maintenance is skipped unless a qualifying full warm check RESULTS_DIR is supplied. |
| WS-09 exit | DONE; 23 authoring/recovery selections passed 24/24 units (`20260921T055045Z-p28162`), final Assessment/subject/Task/Generic checks passed 5/5 (`20260921T055218Z-p69021`). Eight specialized accessibility rows passed 11/11 (`20260921T055157Z-p36472`). Assessment append-recovery/follow-on browser rows passed 13/13 (`20260921T055302Z-p75081`); ordinary creation matrix/continuity passed 11/11 (`20260921T055429Z-p9250`). Final format/type/lint/import/drift passed `20260921T055428Z-p9113/-p9093/-p9103/-p9097/-p9035`; Markdown `20260921T055509Z-p44468`; whitespace passed. Exact selectors are recorded in the retained run manifests and family review below. | G08 complete. Changed shared form roles, Shell/Details, Assessment/Generic compositions, Note/coordination/Task/Decision/Evidence/Indicator forms, related-row renderer, source guides/ownership and affected tests. Saved headers win over operation titles; creation modes and target labels are explicit; Assessment saved reading and append controls share the adopted presentation. Removed two Indicator style modules and migrated thirteen consumers; removed local form style alternatives. All seventeen configurations accounted for below. Existing source/draft/receipt owners and supported generic row-command adapter retained; no storage/protocol migration. Initial lint (`20260921T055116Z-p31387`) found the unused CSSProperties import after local-style retirement; removed and reran successfully. No check waived. Save before mandatory WS-10 entry. |
| WS-09 entry | IN_PROGRESS; WS-08 DONE checkpoint saved before this entry | Migrate specialized form presentation and creation modes while retaining each source owner, submission, detachment and operation recovery. Inventory all seventeen configurations and remove superseded local treatments. |
| WS-08 exit | DONE; `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_ordered_actions,web.workbook.regression.inspector_grouped_permission_reasons,web.workbook.regression.contextual_task_decision_discovery,web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.timeline_inspector_feature_controller_a38c2d6f71` passed 6/6 (`20260921T054436Z-p79987`). Stateful default/closed/denied/surface-switch browser row `module.workbook.browser_stateful.verify_default_closed_inspector_state_no_row_sta_897d604d9e` passed 11/11 (`20260921T054457Z-p81569`). Format/type/lint/import passed `20260921T054426Z-p75155/-p75135/-p75145/-p75139`; generation `20260921T053918Z-p63699`, drift/shape `20260921T054510Z-p3670/-p3682`, Markdown `20260921T054555Z-p20037`, whitespace passed. | G03 Workflow presentation complete for bound actions. Changed shared ContextualActions, Actions/model, presentation contract/schema/generator/facade, presentation tests/catalog, README and inventory below. Canonical order and single dispatch replace wrapping/duplicate inputs; authored outcomes share accessible descriptions with deduplicated reasons. Existing command owners and unsupported-feature omission remain explicit. No stored-data/protocol migration; no wording alias. Initial type/reason checks failed on obsolete tests expecting no description on enabled actions (`20260921T053918Z-p63759`, `20260921T054321Z-p72851`, `20260921T054348Z-p73844`); fixed to validate outcome-plus-reason semantics and reran. One invalid combined target invocation applied OWNER to typecheck; corrected target routing. No acceptance waived. Save before WS-09 entry. |
| WS-08 entry | IN_PROGRESS; WS-07 DONE checkpoint saved before this entry | Replace equal-weight Workflow wrapping with ordered action rows using feature identity, authored labels and explicit outcome metadata. Preserve capability admission, dispatch and reason deduplication. |
| WS-07 exit | DONE; final pending-read/retained-recovery/batch-review units passed 5/5 (`20260921T053140Z-p16493`); public stale/error/reversal and stateful authorization browser rerun passed 13/13 (`20260921T053139Z-p16273`). Four-family paging and legal scopes plus accessibility evidence below passed. Generation `20260921T053117Z-p8731`, format `-p8811`, typecheck `-p8791`, Markdown `20260921T053223Z-p48494`, lint `-p48490`, import boundary `-p48484`, drift and whitespace passed. | G04 and History portion of G03 complete. Each event carries all admitted semantic detail and exact timestamps, attribution is visible, reversal review is event-local and explicitly scoped, record actions trail reading. Closing an event cancels only its unsubmitted review, including late reads; submitted requests/receipts retain their owner. Changed selectors, visual setups and README follow the same public disclosure flow. No protocol/data migration beyond WS-06; integrated visual review remains WS-10. Save before WS-08 entry. |
| WS-07 validation | Semantic unit/actor/value tests and existing reading/recovery selections passed except old hidden-control focus assumptions (`20260921T052425Z-p11841`, 7/8 units). After opening event disclosure in those fixtures, recovery passed 2/2 (`20260921T052601Z-p22104`); final presentation/model/recovery selection 4/4 (`20260921T052806Z-p34533`). All four family overflow/paging/retry/refresh/reversal browser scenarios and four legal-action scenarios passed in `20260921T052358Z-p81064`. Both inspector accessibility scenarios passed 11/11 (`20260921T052922Z-p73034`). | Changed History presentation model/types, event renderer, loaded action/confirmation composition, semantic focus handling, source guidance and public browser fixtures. Visible source attribution, UTC numeric-offset display, exact timestamp and complete units live under Event details; scoped reversal explanations and local confirmations; trailing Record actions. No display label grants an action. Final pending-read cancellation and public failure/authorization reruns follow before exit. |
| WS-07 resolved checks | Generation first rejected a new test title routed to the wrong file (`20260921T052219Z-p66220`); corrected and gave semantic detail its own authored row. Typecheck identified the generated nonempty-unit tuple requirement; fixture now constructs an explicit first unit. Two fixture focus failures clicked concealed actions, and browser inspector-actions fixtures omitted Event details or assumed a different owner error sentence; updated the user flow and asserted the current review requirement while keeping no-write/public-error checks. | `20260921T052358Z-p81064` and `20260921T052804Z-p34284` retain failing browser traces. Inspected disclosure cancellation and fixed pending-read review cancellation/clearing without aborting submitted owner operations; a focused late-response regression now covers this transition. Removed technical-only actor, raw-operation secondary paragraph, global reversal confirmation and leading destructive block. Shared event rendering remains used by record History and batch/merge review. |
| WS-07 entry | IN_PROGRESS; WS-06 DONE checkpoint saved before this entry | Render complete semantic History with visible attribution/time, event-local reversal and trailing Record actions. Preserve action metadata, confirmation safety, selection, paging and operation recovery. |
| WS-06 exit | DONE; final generation `20260921T051327Z-p15586`, typecheck `20260921T051432Z-p50356`, lint `20260921T051522Z-p58283`, imports `20260921T051108Z-p91432`, drift `-p91358`, generated policy `-p91360`, JSON shape `-p91362`, OpenAPI compatibility `-p91368`, Markdown `20260921T051400Z-p19103` and whitespace passed. Behavioral results and resolved failures are below. | G10 source/transport exit complete. Retired arbitrary units and coarse materializers, with no legacy response adapter. Source README and inventory below document canonical-only imports, coordinated 2.0.0 cutover/rollback, and disposable generation restarts preserving operation recovery. No stored-data change. Save this checkpoint before WS-07 entry. |
| WS-06 validation | Source/catalog/materialization/portability unit selection passed 6/6 (`20260921T051155Z-p46164`); frontend History browsing, lookup, transport, retained operations and batch review passed 8/8 (`20260921T050349Z-p3191`). History envelope and retained restart passed in `20260921T050012Z-p58487`; Indicator reversal rerun passed 3/3 (`20260921T050311Z-p82466`); link/tag integration 3/3 (`20260921T050458Z-p35453`); imported attribution/retained reversal 3/3 (`20260921T051259Z-p66812`). Browser entry reversal passed in `20260921T051106Z-p90931`; full keyboard browsing/recovery/reversal rerun passed 11/11 (`20260921T051359Z-p18836`). | Pure source projectors in nine modules, required catalog fields/compiler, Revisions historycontract/materializer/service/HTTP, authored OpenAPI, generated protocol, browser validation/generation boundaries, canonical fixtures and source README changed. All supported source families have complete deterministic units. Missing/malformed facts fail locally; values preserve absent/null/false/zero/empty. No raw snapshot or storage mutation ID is exposed. Imported actor fallback remains declared; no directory lookup. |
| WS-06 resolved failures | Early catalog fixtures lacked required schema/projector fields; corrected fixtures. Generation initially stopped on unreviewed compatibility fingerprints; reviewed eleven findings in pending 2.0.0 and regenerated. Old tests searched retired target IDs or expected assertive read alerts; migrated to semantic references and the existing polite notice owner. The paging scenario exposed absent end-of-history feedback; added it. Imported fixture had partial canonical Host fields and then an incorrect origin default; corrected both to source facts. | Initial failing roots include `20260921T050425Z-p4666`, `20260921T051106Z-p90931`, `20260921T050542Z-p63822` and `20260921T051107Z-p91156`; successful focused reruns above close them. A parallel service-image warm-stamp collision (`20260921T051300Z-p67161`, harness failure before browser launch) was resolved by sequential rerun, without changing harness policy. A mistyped unit row was rejected before execution and corrected to the authored catalog ID. Type/lint fixture errors and a missing callback dependency were corrected; final format passed `20260921T051503Z-p53691`. |
| WS-06 entry | IN_PROGRESS; WS-05 exit saved before entry | Implement source-owned semantic projectors, closed public History contract and coordinated representation cutover. Canonical retained snapshots and conformant imports only; no historical-shape inference, automatic reset or legacy response adapter. |
| WS-05 exit | DONE; module.evidence guide refreshed; web.workbook access slice passed 4/4 (`20260921T043242Z-p25928`), final sibling-failure check 2/2 (`20260921T043452Z-p76052`); Timeline Evidence 2/2 (`20260921T043243Z-p26722`); lifecycle/upload retained 6/6 (`20260921T043313Z-p31737`). Browser/accessibility/stateful run `20260921T043314Z-p32150` passed five selected rows but failed original-source review because its drop targeted the retired container. Updated that fixture to the named attachment region; focused rerun passed 11/11 (`20260921T043746Z-p77584`). Typecheck `20260921T043312Z-p31564`, imports `20260921T043402Z-p64662`, lint `20260921T043403Z-p65087`, format `20260921T043441Z-p71409`, Markdown `20260921T043806Z-p7501` and whitespace passed. | Changed EvidenceAttachmentEntry, EvidenceAccessActions, useEvidenceWorkbookBindings, EvidenceFileRecovery, TimelineEvidencePanel, source ownership, focused unit/browser fixtures and source guidance. Three independent regions distinguish saved information, authorized access and attachment recovery. Picker/drop/paste share admission and source-owned commands; acknowledged writes remain retained. Removed the old outer-container drop/paste path and combined access/attachment component. Compact row access remains a named consumer. Count scope stays explicit; no new enumeration, schema or stored-data migration. Failure artifacts and successful browser output remain under the named run roots. Next WS-06. |
| WS-05 entry | IN_PROGRESS; WS-04 exit saved before entry | Separate accepted Evidence information, independently authorized access and source-local attachment recovery. Keep all upload entry paths on the existing owner admission. |
| WS-04 exit | DONE; refreshed module.entities/module.timeline guides; web.workbook mention candidate/receipt/recovery/reconciliation/auto-resolution slices passed 6/6 units (`20260921T042428Z-p73919`); module.timeline collection inspection passed 2/2 (`20260921T042457Z-p5697`); module.entities dismissal/restoration and creation-recovery browser slice passed 13/13 (`20260921T042428Z-p73950`). Typecheck `20260921T042428Z-p74128`, format `20260921T042404Z-p69087`, lint `20260921T042520Z-p8789`, import `20260921T042522Z-p8984`, whitespace passed. | Changed TimelineMentionsPanel, TimelineMentionActionControls, candidate/collection tests, mentions.recovery browser scenario and presentation README. Compact raw-text/state/target summaries; complete provenance under Mention details; selected correction disclosure remains in original group with mounted inputs. Creation and mention receipts/replay/refresh stay visible outside closed controls. Empty partial candidate pages explicitly retain the continuation meaning. Existing source identity, permissions, canonical-link distinction and Dismissed in this session retained. No protocol/data migration. Expanded shared relationship chips remain used by EntityRelationships' matching Timeline preview; they are not an abandoned alternative. Browser results retained under the successful run's browser groups. Next WS-05. |
| WS-04 entry | IN_PROGRESS; WS-03 exit saved before entry | Compact mention summaries and source-local correction disclosure; preserve bounded candidate discovery and visible operation recovery while controls are collapsed. |
| WS-03 exit | DONE; owner guide refreshed; unit slice `20260921T041810Z-p13491` passed 7/7 units: saved reading, Entity editing, Timeline explicit Details, ordinary recovery, draft binding and lifetime. Browser/accessibility slice `20260921T041811Z-p13712` passed 13/13 units, including subject retention/narrative expansion, detached refresh, exact replay and inspector edit accessibility. Final typecheck `20260921T042045Z-p58296`, focused navigation `20260921T042043Z-p57294`, format `20260921T042042Z-p57026`, lint `20260921T041842Z-p48383`, import `20260921T041843Z-p48574`, generation `20260921T041727Z-p5291`, drift `20260921T041812Z-p14062`, Markdown `20260921T041943Z-p54588` and whitespace passed. | Reworked WorkbookInspectorDetails, ordinary draft read projections/hook, TimelineInspectorDetails and section bindings, Generic presentation, Entity composition, shared section focus descriptor, unit/browser fixtures, catalog and generated routing; updated presentation README. All 17 configurations preserve declared field order and value distinctions. Source string/read contracts classify rows; no second field registry. Explicit editor slots group Update/Close and retain owner feedback. Original-field cues select the correct collection action before Resume; captured receipts remain independent. Removed Done editing and Entity's unused control-stack style. Source owners supply collection destinations; semantic descriptors now carry their focus resolver explicitly. No stored-data or draft-lifetime migration. Next WS-04. |
| WS-03 resolved checks | Initial type/lint exposed an unused Entity style and an effect dependency; corrected both. Initial generation/type/format stopped because the new authored test row was not ASCII-sorted (`20260921T041650Z-p1508` generation); sorted the owner catalog and regenerated. | No generated output was edited by hand and no acceptance was waived. Browser artifacts are the successful run's `browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-inspector-edit/` and `browser-e2e-a11y/browser-groups/a11y-workbook-inspector-edit/` reports. |
| WS-03 entry | IN_PROGRESS; WS-02 exit saved before entry | Implement saved scalar/narrative rows and explicit editor slots; expose original-field retained work through the existing draft owner. Preserve submission, clear/null, revision captures and operation recovery. |
| WS-02 exit | DONE; `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_persistent_context,web.workbook.regression.inspector_explicit_panel_states` passed (3/3, `20260921T040716Z-p77268`); `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention` passed (11/11, `20260921T040648Z-p42149`); typecheck `20260921T040649Z-p42561`, import boundary `20260921T040716Z-p77362`, lint `20260921T040432Z-p3564`, format `20260921T040706Z-p68723`, whitespace passed | Changed declared panel admission, shared shell, new layout navigation hook/context and SurfaceLayout provider; migrated Timeline, Entity, Generic and Assessment callers; marked History Open focus destination; updated unit/browser tests, source ownership and presentation README. One sequence supplies mounted body/navigation. One bounded section selection survives close; no source/request state added. Removed standalone list rendering fallback. Existing semantic openers and operation lifetimes retained. Browser verifies adjacent/overlay widths 1440/760, zero navigation History reads, nested Escape and retained editing. Next WS-03. |
| WS-02 resolved failures | Initial format/lint rejected nonsemantic disclosure markup; unit fixture used unavailable matchers; typecheck rejected a testing-library option. Corrected and reran. Browser runs `20260921T040259Z-p64137` and `20260921T040523Z-p4281` failed grid focus during retarget; trace identified panel remount causing editor autofocus. | Kept section keys stable across subject changes and restricted concealment focus recovery to lost focus within the same subject. Added mount-continuity assertion. Successful ordinary browser rerun above verifies recovery; no test assertion waived. Artifacts are under each run root's `browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-inspector-edit/`. Invalid diagnostic flag invocation was rejected by harness configuration; ordinary lint provided diagnostics. |
| WS-02 entry | IN_PROGRESS; WS-01 exit recorded first | Derive one admitted descriptor sequence in composition and consume it in shell navigation/body. |
| WS-01 entry | IN_PROGRESS; WS-00 exit recorded before entry | Shared role/caller inventory, authored narrative line-limit projection and focused owner verification. |
| WS-01 exit | DONE; generation `20260921T034856Z-p13199`, format `20260921T034932Z-p16970`, typecheck `20260921T034951Z-p22064`, UI slice `20260921T035119Z-p25506` (10/10 units), Workbook slice `-p25543` (4/4), import `-p25779`, drift `-p25482` passed | Shared complete typography/button/input roles; shell focus/disabled presentation; Timeline/Indicator style declarations now consume common roles. Existing style export names retain current callers without separate values. Authored narrative limit generated; no dependency/theme change. Final browser/density/golden review remains WS-10. |
| WS-01 verification setup | `make generate`, `make format`, `make frontend-typecheck` and `make doctor` passed. Initial slices/import/drift runs failed before child launch with `infra/service_start_error` (runs `20260921T034951Z-p21785`, `-p21832`, `-p22136`, `-p21882`). `make explain-run` confirmed classification. | The shell had no `node` on PATH; Make's pinned runtime exists at `tmp/node-runtime/bin`. Rerun public Make targets with that runtime prepended to PATH; no harness policy change or global installation. |
| WS-00 entry | IN_PROGRESS; baseline `67f3d98ac01b42a1da4409c0a6630d08d24e1761`; audit already staged, no other changes | Adopt approved design/History decisions. Canonical retained snapshots only: Core 02 REQ-02-265 forbids schema-less history decoding or backfill. No database reset is authorized. |
| WS-00 exit | DONE; `make lint-markdown` passed, final owner-text run `.cartulary/test-results/20260921T034447Z-p9117`, summary `adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` passed | Changed design, Core 01/02/03 and this tracker. Adopted closed History values/units, attribution, coordinated cutover and canonical-only admission; added explicit navigation, editing, Evidence, History and Workflow criteria. No product check claimed. Next WS-01. |

### WS-06 source and compatibility inventory

| Source owner | Required canonical contribution | Detail boundary |
| --- | --- | --- |
| Timeline | Timeline record | Public fields, generated-value flags, capture/review/supersession facts |
| Entities | Host and Identity; mention, alias, preserved identifier | Public fields, merge target, mention provenance/state/resolution, semantic identifiers |
| Evidence | Evidence record | Public metadata and lifecycle/upload state; excludes object bytes, object storage identifiers and access handles |
| Assessments | Assessment record | Subject, assessor, assessed time, confidence, state and rationale |
| Artifacts | Artifact record, all eight configured variants | Existing source catalog fields and readonly assigned identifiers/lifecycle facts |
| Tasks/Decisions | Task and Decision records | Existing source catalog public fields |
| Parties | Party record | Public party fields |
| Indicators | Indicator record, observation and interval | Identity values, observation resolution, interval rationale/support/confidence/times |
| Links | Record link and tag | Public endpoints/type/provenance and tag identity; attached Evidence has its own semantic kind |

The application catalog admits ten record snapshot types and fourteen mutation
target kinds from these nine owners. Twenty-four authored fixtures exercise
seventeen record configurations and seven non-row families; row target aliases
reuse their admitted record projector. Missing or duplicate contributions fail
catalog construction. Deterministic units expose semantic references and typed
absent/null/present values, never raw target IDs or snapshots. The browser uses
the generated closed response validator and validates duplicate identities.

The public owner source is
`contracts/openapi-source/owners/module.revisions/openapi.json`. Generated
protocol types/validators and assembled OpenAPI are downstream. The pending
2.0.0 change set records eleven reviewed findings; the released 1.0.0 baseline
is unchanged. `data.representation_generation` also exists on empty pages.
Continuation and action-proof reads compare this marker before immutable item
content and require a fresh first page across generations. Captured mutation
bytes, receipts, drafts, selectors, item references and `/api/v1` remain intact.

Deploy the bundled server/browser together and retain the coherent prior bundle
for rollback. No response adapter, SQL migration, automatic reset or historical
shape repair exists. Supported data is canonical retained history and conformant
imports. Unsupported legacy databases remain at the existing explicit operator
reset boundary. Imported attribution uses the portability owner's source actor,
without an account-directory dependency. The final visual display is WS-07.

### WS-10 integrated validation progress

The [validation handoff](workbook-inspector-validation-handoff.md) records exact
commands, run roots, failed-run resolution, the complete changed-file inventory,
151 individually mapped image reviews, digest acceptance and paired rollout/rollback.
This tracker remains the controlling status and decision artifact.

WS-10 is **DONE**. All substantive implementation corrections, required behavioral
checks, reviewed visual refresh, two ordinary confirmations and final documentation
checks passed. All applicable A001–A027 rows are PASS in the validation handoff.

- `make agent-finalize` passed `20260921T073043Z-p91957` before broader final
  verification. Retained-run maintenance was **SKIPPED** because `RESULTS_DIR`
  was unset; no qualifying full warm check was supplied.
- Final-source `make test-fast` passed **708/708** at
  `20260921T073129Z-p97071`. Integrated checks exposed and corrected exact saved-text
  copy, canonical null/empty/absent preservation and the separation of presentation
  reset from access scope. Regression checks and browser exercises pass.
- All nine failed scenarios from the initial full webserver-backed run
  `20260921T061637Z-p18330` passed their documented focused reruns. Final stateful
  run `20260921T073129Z-p96952` passed 39/42 units; its two transient failures passed
  unchanged in isolation at `20260921T073603Z-p78561` and
  `20260921T074157Z-p46106`. Final accessibility run `20260921T073453Z-p32834`
  passed 18/20 units; its Coordination focus scenario passed unchanged at
  `20260921T074150Z-p32164`. Earlier full stateful 42/42 and accessibility 20/20
  runs remain recorded separately. Failed full runs are never relabeled as passes.
- Revisions canonical-fixture failures were corrected without relaxing production
  admission; the three failed rows passed 4/4 at `20260921T062140Z-p68836` and
  all remaining full-run rows passed. No decoder, reset or backfill was introduced.
- Final generation, drift, generated policy, JSON shape, compatibility, type,
  import-boundary and lint checks pass; exact runs are in the handoff ledger.
  Server/web candidate builds pass at `20260921T073129Z-p97267/-p97300`.
- Baseline and final candidate archives each contain 73 verified server/web files;
  their manifests and SHA-256 digests are in the handoff. No external deployment,
  SQL migration, history rewrite or forced browser reload was performed.
- Visual reconciliation accounts for 255 captures/goldens/active mappings, zero
  orphans/missing/ambiguous mappings and 29 resolved registered fixtures. All 151
  changed images were reviewed; later changed captures were reviewed again in full.
  The Evidence anchor and Network Flow read-readiness fixture corrections remove
  timing-dependent placement/state. Final public refresh `20260921T074355Z-p30667`
  passed 12/12, followed by two fresh ordinary runs `20260921T074928Z-p69256` and
  `20260921T074928Z-p69271`, each 12/12. Final Markdown
  `20260921T075527Z-p42495` and staged/unstaged whitespace checks passed.
- All seventeen configurations and specialized authoring families have named
  source/verification coverage. Optional unbound features retain the explicit
  Core 01 `omit_feature` N/A rationale. Five-task comparison records scripted
  completion and source differences; participant completion/error/scroll/interpretation
  metrics were unavailable and remain usability hypotheses.

Run roots are relative to `.cartulary/test-results/`; manifests, row receipts,
logs and browser reports remain available. No mandatory acceptance is waived.

### WS-09 family and retirement review

| Schema configurations | Saved presentation / authoring boundary | Validation retained |
| --- | --- | --- |
| Timeline | Reading-first Details; source-local Note, Task/Decision, coordination and related-Evidence forms | WS-03/05 plus specialized authoring, exact replay and accepted-write recovery |
| Hosts, Identities | Shared saved Details and Entity relationships; Note and Task/Decision forms | WS-03/04 plus contextual creation source-retention matrix |
| Evidence | Shared Generic saved Details; distinct access/attachment; Note and Task/Decision forms | WS-05 plus contextual and related-Evidence owner tests |
| Notes | Shared Generic saved Details and explicitly named new-record draft; Note source control and local Note form | Note source replacement/clearing, detachment, revocation and recovery |
| Indicators | Shared Generic saved Details; separate canonical proposal, observation capture and lifecycle authoring | Each Indicator family's authoring/reconciliation tests and keyboard/accessibility flow |
| Assessments | Append-only saved values use shared read-only Details; selected record remains the header; append title is local | Stable selection, subject-only follow-on, close/resume/discard and append receipt recovery |
| Task Requests, Decisions | Shared Generic saved Details; dedicated Task lifecycle and Decision supersession; contextual creation | Guarded source/lifecycle review, independent submission and retained receipt tests |
| Parties | Shared Generic saved Details and named new-record draft | Ordinary schema matrix, required minima and continuity |
| Communications Log, Handoff, Status Review, Lesson | Shared Generic saved Details; source-owned coordination forms and contextual entries | Coordination source replacement/clearing, raw values, permission loss and recovery |
| Findings, Investigative Queries, Forensic Keywords | Shared Generic saved Details and named new-record draft; contextual entries where declared | Ordinary seventeen-schema registry and source-owned contextual tests |

Every configuration is accounted for. Optional unimplemented discovery surfaces
retain the explicit WS-08 owner-based N/A inventory. Assessment immutable history
is not ordinary field editing; its read-only facade supplies no mutation commands.

Removed `features/indicators/observationStyles.ts` and
`features/indicators/indicatorLifecycleStyles.ts` after migrating all thirteen
consumers to shared form roles. Removed local input/group/action styles in Note,
coordination, Decision, related-Evidence, Assessment and the generic related-row
presentation. Source-specific textarea resize, select appearance and source text
preservation remain meaningful variants of those roles. Shared group, heading,
field, action and preserved-text exports have named consumers in these families.

The related-row renderer and seed builder remain for their existing checked
row-creation command adapter; contextual source owners still intercept their
specialized families. This effort does not add a form engine or merge their
validation/recovery lifetimes. `Close draft` in an attached Assessment now invokes
the existing detach command next to its explicit append action; discard remains
separate. Generic creation labels now name the target instead of "Commit draft row".

### WS-08 action inventory and scope

The shared dispatcher admits 48 contextual command instances across all seventeen
configurations: 37 row-creation entries, four linked-Note entries, four Indicator
entries, two Timeline capture actions and one Decision supersession action.
They render once in canonical order, with authored labels and outcome descriptions.
Indicator mutation entry points open authoring; read entry points open review.
The other source-owned History, Evidence, mention, merge, reference and lifecycle
controls continue to use their existing commands and admission. `panel_read`
features describe content contributions, not additional buttons.

Core 01 REQ-01-615 explicitly specifies `unsupported_feature_behavior=omit_feature`.
The following optional discovery features have no current bound implementation;
they are **N/A for presentation migration**, not claimed implemented or removed
from the protocol. This scope decision does not waive supported field editing,
History or source-owned controls. No placeholder action or invented query is added.

| Configurations | Optional Workflow discovery not bound by this client | Existing supported destination |
| --- | --- | --- |
| Hosts, Identities | Timeline/Evidence/Assessment surface pivots | Entity Relationships preview and ordinary sheet navigation |
| Evidence, Notes | Linked/source-record surface pivots; Evidence Timeline pivot | Authorized association content and ordinary sheet navigation |
| Parties | Four usage pivots and reference-clear summary | Saved Party Details; reference controls belong to referencing records |
| Assessments | Subject/prior-assessment summary shortcuts | Saved subject Details and explicit History read |
| Communications Log, Handoff, Status Review | Named party/next-report/acknowledgement shortcuts and cross-record review summaries | Declared readable fields and admitted ordinary editors; related creation |
| Decisions, Findings, Lesson | Named owner/state/affected-record shortcuts | Admitted Details editors; Decision supersession retains its dedicated owner |
| Task Requests | Separate decision link/clear shortcuts | Admitted decision-reference editor; dedicated Task lifecycle control |
| Investigative Queries, Forensic Keywords | Named source/result/finding/Timeline reference shortcuts | Admitted collection editors in Details |
| Timeline, Indicators | No additional unsupported Workflow shortcuts | All declared Workflow creation entries use their bound owners |

This is an explicit migration inventory under the adopted omission boundary,
not a claim that the repository implements every feature in its discovery vocabulary.
Future implementations must supply a checked source binding and exact seeded query
or command, then join the same presentation hierarchy. Labels are never dispatch keys.

### Scope and authority

| Session | State / files | Commands or review | Result / next action |
| --- | --- | --- | --- |
| Original audit | Source/screenshot audit at the baseline above | Original evidence below | Historical findings retained; not fresh verification |
| Remediation planning P-01 | Only this tracker changed; four user documents, AGENTS, framework and skill guidance read | Branch/HEAD/status and owner/source inspection | DONE; updated checkpoint before beginning P-02 |
| Remediation planning P-02 | G00–G11 and WS-00–10 defined in this tracker | Owner-to-gap and compatibility review | DONE; each original finding separately accounted for, History unit mapping and migration gates recorded before P-03 |
| Remediation planning P-03 | Documentation validation and final handoff | Completeness, owner review, whitespace and public Markdown lint | DONE; fresh lint `20260921T033650Z-p1857`, four task guides and whitespace passed before implementation authorization |

### Backend boundary

| Session | Inspected / proposed boundary | Result / risk / next action |
| --- | --- | --- |
| P-01 | Revisions History materializer/service/model, provider contributions/catalog, revisionassembly navigation; Core 01 History and Core 02 §15 | Confirmed coarse units; WS-06 must inventory source providers, retained formats and imported actor parity before changing interfaces |

### Frontend boundary

| Session | Inspected / proposed boundary | Result / risk / next action |
| --- | --- | --- |
| P-01 | §2 shell/Details/History/Timeline/feature seams and source guides | Keep shared presentation stateless; navigation uses existing admission and focus; retain source-owned authoring/operations |
| P-02 | Shared-form migration and explicit editor action slots | No new form/retry engine; inventory remaining feature callers at each slice entry |

### Contract and code generation

| Session | Inspected / proposed boundary | Result / next action |
| --- | --- | --- |
| P-01 | Design/view contracts, Revisions OpenAPI source, generated History type, generated-artifact policy | Edit authored owners; generated policy additionally declares `packages/view-contracts/src/generated` and aggregate OpenAPI as generated; do not hand-edit them |
| P-02 | Public History unit correction | API classification and retained browsing generation transition required; no assertion of zero migration until WS-06 evidence |

### Tests and harness

| Session | Commands / evidence | Result / artifacts / next action |
| --- | --- | --- |
| P-01 | Four `make task-guide ROLE=module-author OWNER=...` calls: web.workbook, module.workbook, package.ui, module.revisions | All passed; routing only |
| P-01 | `make explain-test-owner OWNER=web.workbook` and `OWNER=package.ui` | Both passed; current catalog families/rows inspected |
| P-01/P-02 | Manual source searches, reads and `make help` | Located exact source paths and public command surface; confirmed 24 Timeline fields; no product verification target executed |
| P-03 | Documentation lint, whitespace and local-link review | PASS: Markdown `20260921T033650Z-p1857`; staged/unstaged whitespace and local references reviewed. Planning evidence only. |

### Security and authorization

| Session | Evidence / retained obligations | Result / next action |
| --- | --- | --- |
| P-01/P-02 | Core 04 authorization/Evidence clauses, explicit concealed region unions, History unit data boundary | No confirmed new authorization defect claimed. Require no hidden destinations/counts, no raw snapshot disclosure, source-authorized actor/value presentation and current request-time rights. |

### Open risks and next session

| Session | State | Next action |
| --- | --- | --- |
| Planning handoff | Production code, adopted docs, tests, contracts and generated artifacts unchanged | Later implementation task starts WS-00; inspect current HEAD/dirty state and this tracker, close owner decisions, then execute slices with tracker updates |

### Historical audit execution and advisory dispositions

The following record predates this remediation pass. Its commands, timings and
screenshot reviews remain historical evidence only.

`make task-guide ROLE=module-author OWNER=web.workbook` completed successfully
and selected `make test-slice OWNER=web.workbook` as the focused entry point.
`make task-guide ROLE=module-author OWNER=module.workbook` also completed and
identified the focused service-backed and browser verification routes.
Repository/source inspection and seven local image inspections completed;
all 28 document source/image links resolved locally.

One cache-suppressed offline UX query, `progressive disclosure hierarchy dense
navigation`, returned five results with no fallback. Applied dispositions:

- **ADOPT** semantic heading hierarchy and unobscured focus through R002/R013.
- **ADAPT** sticky navigation to the inspector's internal scrollport through
  R014/R018; generic page-padding examples do not govern workbook geometry.
- **ADAPT** typography hierarchy through R011/R019 using Cartulary's existing
  roles, rather than an upstream scale.
- **REJECT for this scope** breadcrumb navigation and hero-title balancing:
  neither addresses the inspector's record/section structure. No generated
  palette, framework, font, icon dependency, or design-system output was used.

Only this audit document was authored. Product tests, browser captures,
accessibility scans, generated-artifact changes, and full suites were skipped
because this task is planning only. `agent-finalize` was not run: no broader
product verification was performed, and its generated-structure refresh would
exceed this documentation-only scope. Retained-run maintenance was skipped
because `RESULTS_DIR` was unset. Documentation lint results are recorded below
for this audit. No product-readiness or accessibility-pass claim is made.

`make lint-markdown` passed (84.4 seconds). Run root:
`.cartulary/test-results/20260921T031046Z-p81814`; summary:
`adhoc/lint-markdown/tool-run-summary.json`. The new document was also inspected
with a no-index whitespace check; no whitespace diagnostics were emitted.
The documentation run reported no failing target. Final working-tree review
showed only this new Markdown document.

## 11. Open decisions, blockers and risk register

There is no blocker to completing this plan. Implementation dependencies below
must be resolved in their named workstream; they do not justify inventing payload
facts or silently dropping required behavior.

| ID | Decision / risk | Why it matters | Needed authority / evidence | Status and resolution gate |
| --- | --- | --- | --- | --- |
| RB-001 | Adopt proposed navigation/disclosure/default copy and layout rules | These are intentional new UX behaviors, not current defects | Design owner and §7 complete decision table | DONE; user-approved defaults adopted in design §12.7, D-AC-082–086 |
| RB-002 | Close History unit semantics and safe visibility | Current payload lacks required detail; arbitrary units are not a durable public contract | Core 01/02/03, Revisions/source owner review and payload security fixtures | DONE; adopted Core 01/02/03 clauses and WS-06 closed contract, source projectors, safe disclosure and canonical validation implemented and tested |
| RB-003 | Supported HTTP consumers and released-contract classification | Cannot assume a tightened schema or new unit shape is compatible | Immutable OpenAPI release comparison; actual supported consumer inventory | DONE; eleven reviewed History fingerprints in pending 2.0.0 change set; immutable 1.0.0 baseline preserved; bundled client/server cutover |
| RB-004 | Retained History reconstruction and imported attribution parity | Older snapshots or undeclared actor fields may defeat a superficial UI fix | Source-provider/retained-data inventory and imported-attribution owner clauses | DONE; ten canonical record snapshot types / fourteen targets require source projectors; imported attribution and retained rollback verified; incomplete canonical fixtures corrected without a runtime decoder |
| RB-005 | History read equality across deployment | Changed diff payload can look like altered immutable committed content | Browsing generation and compatibility/rollback tests | DONE; opaque representation generation checked before item equality, fresh reads preserve operation owners, and WS-07 renders every admitted semantic unit with scoped actions |
| RB-006 | Full feature-family coverage beyond the original sample | Generic appearance does not prove specialized semantics | §8 17-schema and authoring-family matrix | DONE; WS-09 seventeen-schema/family inventory, source-local recovery tests and WS-10 integrated evidence |
| RB-007 | Baseline services/browser renderer and current test health | Historical passes/failures are not current evidence | Fresh focused runs; explain-target/run artifacts if failure | DONE; fresh owner/browser checks and resolved fixture/source failures are recorded in the validation handoff; no historical result substituted for current verification |
| RB-008 | Claimed reduction in effort/confusion | Design judgment is not a measured result | Comparative task review with disclosed participants/method | DONE with limitation; same five tasks compared through source/scripted behavior and reviewed captures; no participants, timed completion, error-rate or scroll-distance measurements; residual hypothesis recorded |
| RB-009 | Rich Evidence browsing without existing authorized item data | Count cannot manufacture an item catalog or access | Separate Evidence read capability proposal | DEFERRED outside required workstreams |

## 12. Binary completion criteria

### Planning completion

- P-01, P-02 and P-03 are DONE with updates recorded in sequence.
- All nine original audit findings have G01–G09 coverage; specification closure,
  History data completeness and validation gaps are covered by G00/G10/G11.
- Each gap has remediation, change areas, rationale, long-term benefit,
  compatibility/migration impact, unresolved risk and testable exit.
- Every named target seam is inventoried; uninspected neighbors are explicitly
  excluded or assigned a workstream discovery gate. Owner, implementation and
  verification identities are distinguished.
- Each implementation slice is its own workstream with dependencies, risks,
  validation, retirement and rollback; WS-10 is the mandatory final slice.
- Fresh documentation validation and final diff scope are recorded. Only this
  tracker changed; its pre-existing staged content remains staged without index
  manipulation. No product PASS or implementation completion is claimed.

### Overall remediation completion after implementation

- WS-00 through WS-10 are DONE, with each exit and tracker update recorded before
  the next workstream began. No mandatory gap is silently DEFERRED or DROPPED.
- Adopted owner text is coherent, projections are generated from authored machine
  inputs, and code/tests/docs agree without any product dependency on Markdown.
- G01–G10 behavioral exits pass across the schema/state/authority matrix. History
  displays required source-owned detail; a cosmetic fallback cannot close G10.
- All migrated callers use the intended shared presentation. Redundant style,
  label and layout alternatives are removed; each retained adapter has a named
  consumer and removal gate. No indefinite dual inspector implementation remains.
- Applicable A001–A027 outcomes are PASS with evidence; N/A is justified by scope
  and owner, and no applicable BLOCKED row remains. Every required test failure
  has been fixed or resolved with evidence rather than waived as historical.
- Golden updates follow ordinary-run reconciliation, authorized update, image
  review and ordinary validation. First-viewport Details/edit/retained-work views
  are covered. Product readiness, visual support and usability observations are
  reported separately.
- History compatibility/migration and rollback are proven for supported clients
  and retained data. Logical audit identities and captured operation recovery
  survive the supported cutover. No unauthorized snapshot/actor/Evidence data is
  newly exposed.
- Final handoff states adopted decisions, changed paths, substantive edits,
  exact commands/results/run roots, failures and resolutions, skipped checks with
  reasons, compatibility/retirement, rollout/rollback, residual hypotheses and
  next maintainer action. RESULTS_DIR maintenance is explicitly recorded as used
  or skipped. Only then is the overall effort complete.
