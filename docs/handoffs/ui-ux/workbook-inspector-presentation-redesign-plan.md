# Workbook inspector presentation redesign plan

## Status and scope

Planning baseline: 2026-09-21, branch `main`, commit
`a4ef612aea434696361e9b3d5683f72eecd0180b`. The checkout was clean before this
document was added. This plan incorporates the user's attachment titled
“Inspector redesign: concrete design closure” and supersedes the corresponding
open-ended recommendations in the preceding conversational audit.

**Delivery complete: 2026-09-21. All workstreams and applicable acceptance rows
are PASS/DONE; final validation and handoff appear in §14.**

This is the controlling implementation tracker. Implementation was authorized
on 2026-09-21 against the planning baseline above; this file was already staged.
The decisions below become requirements through amendments in their named owners,
not through this tracker. Historical planning evidence is not a current pass.
Automated checks and reviewed screenshots are the final validation gate. No
walkthrough, participant evaluation, or measured usability claim is required.

The scope is a bounded presentation redesign. Existing source owners retain
authoring, authorization, request identity, receipts, acknowledgement, and
recovery. The shared presentation gains explicit layout and contribution rules;
it does not gain another inspector architecture, draft store, mutation engine,
navigation registry, or global Recovery catalog.

The earlier completed inspector trackers remain historical evidence. Their
completion statements are not completion evidence for these new slices.

## 1. Current evidence and the intended improvement

The initial audit inspected shared presentation, Timeline and generic
composition, editing, navigation, History, tokens, and eight checked-in inspector
captures. It did not conduct a live participant study or recapture screenshots.
The attachment's author reviewed an uploaded specification snapshot; the
repository findings below were checked against the baseline above.

| Observed weakness | Proposed improvement | Current evidence |
| --- | --- | --- |
| Sparse Timeline metadata and repeated Edit buttons occupy the first Details viewport before activity content. | Explicit compact property rows for eight Timeline fields; preserve two narrative fields. | [Saved Details](../../apps/web/src/workbook/inspector/WorkbookInspectorSavedDetails.tsx), [Details baseline](../../apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-details-linux.png). |
| A tall header exposes only the current section and a chooser trigger. | Compact header and measured direct-navigation fit with deterministic chooser fallback. | [Shell](../../apps/web/src/workbook/inspector/presentation/WorkbookInspectorShell.tsx), [navigation](../../apps/web/src/workbook/layout/workbookInspectorNavigation.ts). |
| Ordinary editing and retained work present multiple competing controls. | Local action hierarchy, safe one-activation Resume, and record-scoped attention navigation. | [Details editing](../../apps/web/src/workbook/inspector/WorkbookInspectorDetails.tsx), [draft feedback](../../apps/web/src/workbook/inspector/WorkbookInspectorDraftFeedback.tsx). |
| Details-wide read-only explanation follows affected fields. | Place the reason before the first affected field. | `WorkbookInspectorDetails.tsx`. |
| Relationship identity and correction controls are visually fragmented. | Selected correction immediately after its item, with chooser-local read controls. | [Mentions](../../apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx), [correction](../../apps/web/src/workbook/timeline/components/TimelineMentionActionControls.tsx). |
| History detail mixes semantic changes and diagnostic references. | Readable typed Before/After content and subordinate technical disclosure. | [History presentation](../../apps/web/src/workbook/inspector/presentation/WorkbookHistoryPresentation.tsx). |
| Workflow repeats a button and outcome paragraph for each command. | Compact command rows and authoring immediately after the originating command. | [Actions](../../apps/web/src/workbook/inspector/presentation/WorkbookInspectorActions.tsx), [declared panels](../../apps/web/src/workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx). |

These were observed design weaknesses at the planning baseline. This effort
makes no measured claim about faster or more accurate analyst work. The design
retains the grid-first graphite workbook, its typography, and scarce semantic
accent. Density comes from hierarchy, alignment, and spacing, not smaller text.

## 2. Amendment ownership

| Existing owner | Proposed amendment or retained boundary |
| --- | --- |
| [Design](../design.md) §7.3 | Header composition, complete-title disclosure, persistent-region budget, Close, and one body scrollport. |
| Design §7.4 | Inspector-only reflow from 320px to the desktop minimum; preserve the surrounding workbook's separate support scope. |
| Design §12.7 | Direct/chooser fit, focus transitions, field-layout overrides, attention presentation, and specialized-section composition. |
| Design §§12.4, 14.1, 14.2 | Quiet field actions, target dimensions, state treatments, accessibility checks, and announcement ownership. |
| Design §3.1.1 | Explicitly include adopted inspector values and mappings in the executable presentation-projection scope. Its present enumeration does not include §12.7. |
| [Core 03](../spec/03_workbook_interaction_collaboration_and_workflows.md) §2.3A | Explicit Clear intent and owner-to-presentation attention boundary; reference existing authoring and operation lifetimes. |
| [Core 04](../spec/04_security_deployment_and_conformance.md) §9 near AC-453–AC-462 | Behavioral acceptance for navigation, continuity, access loss, and recovery. Visual-only acceptance remains Design-owned. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md) REQ-01-615 and §7.4 | Retain schema, panel, field, and feature identities. No new public discovery member is proposed. |
| Core 00 and Core 02 | Preserve authority, record families, persistence, authorization, and profiles. No substantive amendment is proposed. |
| Existing appendices and UI/UX/source guides | Informative rationale, annotated examples, component bindings, and verification procedure. No analyst ritual or executable authority is added. |

Core 01 remains the only inspector capability/configuration authority, including
the maximum five panels and 64 feature groups. This plan does not duplicate that
registry. Executable consumers use authored machine projections outside
documentation and their generated facades; human review checks owner fidelity.

## 3. Navigation and focus decisions

Use one labelled navigation region with ordinary `type="button"` controls and
`aria-current="location"` on exactly the current destination. Sections remain
locations within a mounted body. No tabs, tab panels, roving tab selection, or
tab-specific arrow-key behavior are introduced. The meaning of current location
is consistent with the dated [WAI-ARIA 1.2 definition](https://www.w3.org/TR/2023/REC-wai-aria-1.2-20230606/#aria-current).

| Condition | Selected presentation |
| --- | --- |
| No admitted destinations | No navigation. |
| Complete labels, targets, gaps, padding, and focus treatment fit one row | Every admitted destination directly, in configured order. Equality counts as fitting. |
| Complete row does not fit | Labelled Sections chooser containing those same destinations. |
| Font/label measurement unavailable | Chooser until fit is established. |
| Capacity changes while chooser is open | Preserve the usable open chooser; reevaluate after selection or dismissal. |

Reevaluate after inspector/viewport resizing, font availability, text-spacing
changes, and admitted-label changes. Do not abbreviate labels, shrink text, add
horizontal navigation scrolling, or wrap into multiple persistent rows. The five
standard labels must fit at 420px inspector width with default typography and
spacing. This replaces the earlier audit's underspecified fit suggestion.

Activation revalidates current subject and authority, closes an open chooser,
scrolls only the body, and focuses the heading or declared entry control. It
does not submit, discard, detach an ordinary editor, or initiate a resource read.
Unrequested History focuses Open history without invoking it.

| Event | Selected behavior |
| --- | --- |
| Explicit admitted opener destination | Enter that section. |
| Same subject reopened without a destination | Retain the existing section. |
| New subject without a destination | First admitted section. |
| Passive scrolling | Existing last-heading-at-or-above-usable-top rule; no focus movement or reads. |
| Scrollable body reaches the bottom | Final admitted section becomes current. |
| Body has no overflow | First section for passive indication; explicit activation still focuses its destination. |
| Current section removed | First remaining section becomes current. |
| Focused section removed/concealed | First surviving navigation control, else Close, else existing shell access-loss path. |
| Removed section did not contain focus | Do not move focus. |
| Focused direct button becomes chooser | Transfer focus to Sections without body scrolling. |
| Focused chooser trigger becomes direct navigation | Transfer focus to current destination button. |

Newer deliberate navigation, subject replacement, and authority replacement
invalidate pending focus work. Existing innermost Escape and pointer-dismissal
semantics remain. Late mounting must not steal focus.

## 4. Details layout and value semantics

The following override applies only to `cartulary.view.timeline.v2`. It imports
existing field identities and never establishes membership or editability.

| Exact field key | Saved-value layout |
| --- | --- |
| `timeline.date_entered_text` | Compact property row |
| `timeline.analyst_text` | Compact property row |
| `timeline.mitre_stage_text` | Compact property row |
| `timeline.device_object_text` | Compact property row |
| `timeline.ip_address_text` | Compact property row |
| `timeline.activity_utc_text` | Compact property row |
| `timeline.activity_local_text` | Compact property row |
| `timeline.raw_activity_text` | Full-width narrative |
| `timeline.activity_synopsis_text` | Full-width narrative |
| `timeline.data_source_text` | Compact property row |

Other fields retain existing semantic classification and readable stacked
fallback. Keep declared order. Reject duplicate overrides, nonexistent field
references, and unknown layout values in authored-contract validation. The
current Timeline projection contains all ten identities.

Compact rows align label, value, and quiet but discoverable Edit control. Long
values may increase row height. Below 320 CSS px inner content width, stack the
label above value/action. Both narrative and compact text exceeding six rendered
lines use explicit expansion; exactly six lines needs none. Expansion must
preserve editor identity. Empty narrative fields reserve no blank reading area.

Retain source text, whitespace, line endings, and inert rendering, including
Timeline's source-string date fields and existing 32,768-Unicode-scalar limit.
Do not infer canonical dates, entities, links, or Markdown from display content.

| Accepted observation | Presentation |
| --- | --- |
| `null` | Not set |
| Valid `""` | Empty text |
| Whitespace-only string | Whitespace only cue with exact source available for inspection/editing |
| Numeric zero / boolean false | `0` / `False` |
| Intentionally unrequested or unobserved | Not loaded, qualified by owning read state |
| Concealed | No protected label, value, count, disclosure, or revealing placeholder |
| Required response field missing | Owner-classified response failure; no fabricated empty/unloaded value |

Collection emptiness and zero counts describe their actual observed scope only.
Response validation remains in its existing owner, not the saved-value renderer.

Clear changes only local draft intent. Nullable Timeline Clear drafts explicit
`null`; deleting every character drafts `""`. Other fields use their declared
clear operation, or omit Clear when none exists. Specialized collections,
relationships, aliases, and workflows retain their own commands. Explain
Timeline Clear as “Sets the value to Not set when updated.” It never submits,
discards a draft, or cancels an operation.

Repository qualification: [Edit Control](../../apps/web/src/workbook/inspector/WorkbookInspectorEditControl.tsx)
already calls `edit.update(null)` for admitted scalar Clear. The draft stores
`string | null`, and [Timeline Details](../../apps/web/src/workbook/timeline/components/TimelineInspectorDetails.tsx)
submits `edit.value`. Preserve and characterize this distinction; do not replace
it under the assumption it is absent.

## 5. Editing and record-scoped attention

Keep at most one ordinary Details editor attached, with accepted value context
visible and the draft labelled Unsaved change. The accepted narrative may remain
collapsed. Group Update with Close editor; keep Clear beside the value and
Discard visually separate. Update or eligible Ctrl/Cmd+Enter submits. Composition
and nested controls take precedence; plain Enter retains native control behavior.
Close, Escape, blur, Tab, and section navigation do not submit or discard.
Grid autosave and specialized workflows retain separate contracts.

Unfinished work is a summary entry for the current authorized schema and record,
not a combined status enum. Detailed attention stays in the body.

| Owner condition | Local presentation and admitted next step |
| --- | --- |
| Unsubmitted draft; relevant baseline unchanged | Unsaved draft; Resume or Discard |
| Edited value/dependencies changed | Draft needs review; existing review |
| Validation rejection or known failure | Preserve draft and specific failure; owner correction/review |
| Dispatched, in progress | Operation-specific progress and existing allowed actions |
| Outcome uncertain | Update outcome unknown; existing operation recovery |
| Write acknowledged, refresh outstanding | Update saved; refresh needed; reads only |
| Original target currently ineligible | Explain unavailability; no retargeting; safe admitted inspection/discard only |
| No unresolved work | Omit summary entry |

Safe Resume revalidates through the existing owner and attaches in one activation.
Changed dependencies lead to review, never silent baseline acceptance. Conditions
can coexist across revisions/operations; an older operation and newer draft must
not be merged simply because their field is equal.

Extend existing presentation contributions with logical information, without
prescribing an unverified TypeScript interface:

| Contribution | Required logical information |
| --- | --- |
| Context | Owner-issued subject/authority reference plus schema and record |
| Section | Existing panel ID, authorized label, ordered regions, focus entry |
| Attention | Stable owner-issued work identity, category, safe label, original destination |
| Destination | Existing section and owner-issued region/field/action reference |
| Action | Existing owner-admitted command reference |
| Outcome | Existing attempt/outcome identity for announcement deduplication |

The shell may order, render, count, and navigate. It may not read draft stores,
compare dependencies, classify write success, or construct retries. Deduplicate
by owner work identity; order by admitted section, owner local order, then stable
identity. No global registration system is needed. Reopening and layout changes
must not duplicate owners, obligations, requests, receipts, or announcements.

## 6. Independent state and authority

Retain `initial_loading`, `ready`, `refreshing`, `stale_failure`, and `unavailable`
read states, independently from access, command eligibility, authoring, and
operation outcome. Unavailable retains `not_requested`, `load_failed`, and
`owner_blocked` causes. Initial reads show local progress; refresh and stale
failure retain authorized accepted observations; stale empty observations do
not claim current absence. A failed independent region leaves readable siblings
usable, including Notes sources/related notes and Evidence metadata/access.

Apply contribution decisions in order: omit undeclared/inapplicable; report a
missing declared implementation as coverage failure; conceal protected
contributions; retain required unavailable-action discoverability and reasons;
enable admitted commands subject to server authorization.

Place Details-wide read-only context before affected fields. Deduplicate typed
reason identities and parameters within the affected group, retaining
`aria-describedby`; equal wording is not identity. Do not introduce field ACLs.

| Authority change | Preserved owner behavior |
| --- | --- |
| Write permission lost, reading retained | Keep accepted values and owner-retained drafts; prohibit writes. |
| Account-session recovery | Conceal protected presentation while retaining required same-account work. |
| Confirmed incident membership loss | Invalidate incident presentation and leave that incident. |
| Account replacement | Clear former-account protected material and retained work at its boundary. |

Action denial, stale version, and read failure do not imply global access loss.
Concealment covers DOM/accessibility content, labels, descriptions, counts,
notices, previews, and pending presentation callbacks. Late retired responses
cannot restore access. Design §14.2 retains announcement ownership: passive stale
reads polite, immediate action failures on their prescribed assertive path, and
no duplicate success announcement from local and save-status presentation.

## 7. Geometry and accessibility

| Measure | Selected design decision |
| --- | --- |
| Base geometry | 420px default; 360px minimum; maximum `min(560px, 45vw)`; resize only in base mode; existing keyboard step/bounds and non-persistent lifetime. |
| Persistent region | At 1280×720, default typography/spacing, 420px inspector: at most 128 CSS px, including any compact attention entry. |
| Title | Maximum two visible lines; full accessible label and body disclosure. |
| Field action | At least 28×28 CSS px and sufficient width for complete label. |
| Property reflow | Stack below 320 CSS px inner content width. |
| Body | Existing body typography; safely wrapped raw values; one scrolling body and no text-caused inspector-wide horizontal scroll. |
| Enlarged text | Allow persistent content to grow; no fixed-height clipping to enforce the reference budget. |
| Focus | Fully reveal a fitting control; existing maximum-useful-exposure rule for oversized controls. |

128px and 28px are user-selected product choices, not research-derived values.

Amend Design §7.4 so an already-open inspector remains usable from 320px effective
viewport width to the desktop minimum as a work-area-bounded overlay. An existing
eligible selected record retains an accessible opener. Grid content remains
inert while overlaid. Do not apply the 360px base resize minimum to this overlay
or create document scrolling. Below 320px retain degraded safe navigation.
This adds inspector support, not a general mobile-workbook profile.

Current source already distinguishes adjacent and overlay styles in
[WorkbookSurfaceLayout](../../apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx).
Do not assume its base-only 360px clamp is itself an overlay defect. Verify the
opener, work-area geometry, and supported-width behavior together in slice 2.

Use the dated [WCAG 2.2 Recommendation](https://www.w3.org/TR/2024/REC-WCAG22-20241212/)
for applicable checks: ordinary-text contrast 4.5:1, relevant non-text state
contrast 3:1, 200% text enlargement, and 320px vertical-content reflow. Apply
line-height 1.5, paragraph spacing 2, letter spacing 0.12, and word spacing 0.16
font-size multiples together. The standard's 24×24 pointer minimum has exceptions;
the selected 28×28 field-action rule is a separate product decision. Inspector
checks alone do not establish whole-page conformance.

## 8. Specialized composition

| Section | Selected composition and preserved operation boundary |
| --- | --- |
| Relationships | Keep mention/link identity, resolution state, and authorized target together. Selected correction immediately follows its item within its collection. Filter, candidate read, paging, and retry stay inside the chooser without replacing selected identities. |
| Evidence | Separate accepted metadata, access controls, and attachment authoring. Navigation does not newly issue preview/download requests; preserve separately specified explicit opener behavior. Upload, record creation, linking, and refresh retain separate outcomes. A count cannot fabricate an item catalog. |
| History | Show semantic description, operation, supplied actor name or labelled identifier, and absolute UTC time with numeric offset. Before/After comes from typed historical units, not display-string or current-row comparisons. Unavailable historical data is not Not set. Scalar pairs may sit side by side only when both fit; narratives always stack. |
| Workflow | Each command once in declared order; active form immediately after its originating command in the same group. One primary affirmative action per local decision, preserving confirmation, prerequisites, and partial-success explanations. |

History technical references are subordinate except required attribution or safe
confirmation identifiers. Event reversal remains with its event; deletion/restore
remains in Record actions. Whole-change-set review names its wider scope. No
generic Cancel suggests that a dispatched operation or accepted write was undone.

## 9. Repository prerequisites and compatibility

| Prerequisite | Current evidence and remaining gate |
| --- | --- |
| Inspectable source | Baseline and named bindings inspected. Refresh branch, commit, dirty state, and owner contributions before implementation. |
| Projection inputs | [Presentation input](../../contracts/design/presentation.v2.json), [closed schema](../../tools/schemas/cartulary.design_presentation.v2.schema.json), [generator](../../tools/harness/generated-artifacts/design-presentation/design-presentation.mjs), [facade](../../packages/ui-contracts/src/designPresentation.ts), and [tokens](../../contracts/design/tokens.v1.json) inspected. |
| Projection compatibility | The inspector schema rejects extra members and the generator explicitly selects emitted fields. Adding input JSON alone cannot implement the design. Slice 1 must resolve the internal schema-version decision, validation, generation, and consumer migration together before editing projections. No public view discovery extension is needed. |
| Field cross-check | Validate override references against authored `contracts/view-schemas` inputs. Do not read the table in this document from tests or generators. |
| Sparse fixture | `browser.inspector-history workbook visual readiness` in [workbook.visual.spec.ts](../../apps/web/e2e/workbook.visual.spec.ts), starting near line 3502, creates the current sparse Timeline row. Its raw activity is `browser.inspector-history visual inspector details`; synopsis is `browser.inspector-history visual inspector target`. Capture remains `workbook-inspector-details`. |
| Fixture integrity | Keep fixture bytes, preceding field order, viewport, and renderer binding unchanged. Workstream 3 measured first complete raw-activity line visibility. Do not shorten data, skip fields, or substitute a demo renderer to pass. |
| Verification | Source boundary `web.workbook`; independent routes include `web.workbook`, `web.design`, `package.ui`, and browser evidence under `module.workbook`. Inspect current catalog rows before each slice. |
| Rollback | Revert presentation, adopted amendments, authored projections, and generated outputs coherently. No owner-held draft/request/receipt rewrite, database migration, or API migration is proposed. |

The existing contribution seam is retained because it already binds canonical
subjects, sections, independent region states, and source-owned commands. The
new common decisions are layout, measured navigation presentation, and attention
rendering. Another declared workflow extends those contributions without another
state owner. Retire replaced control/layout paths atomically with their callers;
direct and chooser modes share one descriptor sequence and one owner lifetime.

## 10. Delivery ledger and binary exits

Execute workstreams in order **1 → 2 → 3 → 4A → 4B → 4C → 4D → 5 → 6**.
Record each entry baseline and mark IN_PROGRESS before changing its implementation.
Complete its focused checks and update this tracker with the exit, commands,
artifacts, compatibility, and risks before beginning the next workstream.
An applicable failed or blocked check prevents DONE.

| Workstream | Gaps | Status | Entry / exit evidence |
| --- | --- | --- | --- |
| 1 | G1 owner adoption; G2 closed projection migration; G3 documentation coverage | DONE | Entry: `main`, `a4ef612aea434696361e9b3d5683f72eecd0180b`; exit evidence below. |
| 2 | G4 measured navigation; G5 geometry; G6 owner-contributed attention | DONE | Entry: same HEAD plus recorded workstream 1 changes; workstream 1 exit recorded before implementation. |
| 3 | G7 exact field layout; G8 value semantics; G9 Resume and read-only placement | DONE | Entry: same HEAD plus completed workstreams 1–2; workstream 2 exit below recorded before implementation. |
| 4A | G10 item-local Relationships | DONE | Entry: same HEAD plus completed workstreams 1–3; workstream 3 exit below recorded before implementation. |
| 4B | G11 independent Evidence stages and access | DONE | Entry: same HEAD plus completed workstreams 1–4A; workstream 4A exit below recorded before implementation. |
| 4C | G12 typed History values | DONE | Entry: same HEAD plus completed workstreams 1–4B; workstream 4B exit below recorded before implementation. |
| 4D | G13 command-local Workflow authoring | DONE | Entry: same HEAD plus completed workstreams 1–4C; workstream 4C exit below recorded before implementation. |
| 5 | G14 consolidation, retirement, all-schema coverage | DONE | Entry: same HEAD plus completed workstreams 1–4D; workstream 4D exit recorded before consolidation. |
| 6 | G15 current validation and final handoff | DONE | Entry: same HEAD plus completed workstreams 1–5; workstream 5 exit recorded before final handoff. |

| Slice | Entry and work | Required exit |
| --- | --- | --- |
| 1: Owner decisions, projections, representative designs | Refresh baseline; amend named owners; close schema/version decision; implement authored projection/schema/generator alignment and generated facade. Specify reading, editing, review-required, uncertain, read-only, and constrained designs. | Human owner/projection review, authored validation and generation checks, reviewed representative designs, and recorded compatibility/rollback. |
| 2: Shell/navigation | Slice 1 exit. Header budget, fit/fallback, focus transitions, passive position, attention presentation, and 320px opener/overlay. Attention content only from admitted owner contributions. | Direct/chooser and focus tests, no navigation reads/writes/detachment, reference geometry, constrained accessibility. |
| 3: Details/editing | Slice 2 exit. Exact overrides, empty/whitespace states, Clear, quiet controls, safe Resume, original-field cues, and read-only placement. | Field-semantic and draft/recovery tests; unchanged sparse fixture shows raw activity; keyboard and enlarged-text checks. |
| 4A: Relationships | Slice 3 exit. Item-local correction and chooser composition. | Selection identity, paging/filter/retry, retained authoring, and local outcome checks. |
| 4B: Evidence | Slice 4A exit. Metadata/access/authoring composition. | No navigation access request, distinct upload/create/link/refresh outcomes, read-only and failure accessibility. |
| 4C: History | Slice 4B exit. Typed diffs and event-local actions. | Complete typed units, unavailable-data meaning, attribution, reversal scope, read-only refresh recovery. |
| 4D: Workflow | Slice 4C exit. Command-local forms and outcome hierarchy. | Declared ordering, one command instance, source-context retention, partial success and confirmation preservation. |
| 5: Consolidation and retirement | All earlier exits recorded. Remove superseded presentation; cover every required and implemented optional schema; assess digest acceptance. | Applicable owner, accessibility, measurement, and reviewed visual evidence passes; no duplicate owners or legacy paths; final handoff records limitations and rollback. |
| 6: Validation and handoff completion | Slice 5 exit. Consolidate current automated evidence and reviewed screenshots; record final revision, changes, failures, limitations, and rollback. | All applicable acceptance rows PASS; justified N/A only; no required remediation unresolved. No walkthrough or participant evaluation. |

Behavioral and accessibility checks run within each slice. Slice 5 consolidates
their evidence rather than postponing those checks. Investigate failures before
broad reruns; use current public Make routing and Testing Harness mechanics.

| Acceptance area | Binary pass condition; current evidence is recorded in workstream exits | Final result |
| --- | --- | --- |
| Reading viewport | First complete rendered raw-activity line visible without scrolling in unchanged sparse fixture at 1280×720/420px; preceding fields present and ordered. | PASS |
| Navigation presentation | Five standard complete labels fit reference direct mode; pressure selects chooser without clipping, shrinking, scrolling, or wrapping navigation. | PASS |
| Navigation effects | One direct activation or chooser open+selection reaches declared focus target; no editor detachment, mutation, or extra lazy read. | PASS |
| Passive indication | Deterministic non-overflow and bottom cases; no focus movement or requests. | PASS |
| Field semantics | Exact overrides; source strings, whitespace, line endings, null, empty, zero, false, and unobserved states preserved where applicable. | PASS |
| Clear/submission | Clear changes only draft intent; null and empty submissions differ; navigation/close/blur/Tab/Escape do not submit. | PASS |
| Retained work | Safe Resume takes one activation; changed dependencies require review; older results cannot retire newer drafts or steal focus. | PASS |
| Operation recovery | Captured uncertain replay identity/bytes preserved; acknowledged-refresh recovery reads only. | PASS |
| Independent reads | Failure leaves readable siblings usable; stale empty does not assert current absence. | PASS |
| Authorization | Concealment removes protected content and derived cues before replacement; late retired responses stay retired; session/account/incident lifetimes remain distinct. | PASS |
| Geometry/accessibility | Reference budget and targets pass; base/narrow/compact/320px, 200% text, text spacing, long values, focus, and keyboard remain usable. | PASS |
| Specialized workflows | Item subject, Evidence stages, typed historical units, and source-owned workflow confirmations/recovery preserved. | PASS |
| Coverage/retirement | Every required and implemented optional schema exercised; no invented optional surface or parallel authoring/operation implementation. | PASS |

Verified task-guide routes at this unchanged session baseline are
`make task-guide ROLE=module-author OWNER=web.workbook`, `OWNER=web.design`, and
`OWNER=package.ui`. Select narrow `test-slice`/`service-backed-test-slice` rows from
those guides. The active inspector capture row is
`module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`;
the projection row includes
`package.ui.frontend_unit.design_presentation_projection_8e15b85b40`. They are
routing evidence, not requirements or proof of coverage completeness.

Use `make help`/`make help-all` and target investigation for current generation,
measurement, accessibility, and visual tasks. Follow the
[visual golden maintenance guide](../guides/cartulary_visual_golden_maintenance.md).
Run `make agent-finalize` before broader implementation verification. Record
RESULTS_DIR maintenance only when an actual qualifying successful run is used.

## Gap disposition and durable design rationale

These dispositions implement the authorized corrections; the named adopted
owners, not this table, govern behavior. Validation criteria are the binary
acceptance table and corresponding workstream exits above and below.

| Gap / areas | Remediation, rationale and long-term benefit | Compatibility / unresolved risk |
| --- | --- | --- |
| G1 — specification, acceptance, documentation | Adopt decisions once in Design, Core 03 and Core 04; keeps ownership clear and prevents competing requirements. | Replaces chooser-only design and adds 320px inspector support; preserve capability identities. Without closure, compliant-looking implementations can disagree. |
| G2 — contracts, generator, implementation, tests | Closed v2 projection and validated schema/field overrides; extensible mappings without silent dropped members or local constants. | Internal atomic cutover, stable facade, no persisted migration; v1 retired. Without it, schema/generator drift silently loses design values. |
| G3 — documentation, lint configuration | Exact tracker lint inclusion and corrected links/history; evidence remains auditable. | Documentation only. Otherwise excluded files appear falsely validated. |
| G4 — implementation, tests | One measured navigation sequence, synchronous guarded focus and passive position; labels and capacity can evolve without breakpoint guesses. | Location buttons, mounted content, no reads/writes on navigation; helpers migrated. Otherwise focus races, clipping and detachment remain. |
| G5 — specification, implementation, browser tests | Compact persistent chrome, one scrollport, measured reference budget and independent overlay bounds; preserves reading space and recovery reachability. | Base minimum remains 360px; 320px overlays; enlarged text may grow header. Otherwise controls or content become unreachable. |
| G6 — owner adapters, presentation, tests | Immutable owner-issued attention with work identity and authorization admission; future owners contribute without another work store. | Distinct newer drafts/older attempts retained; no field-equality deduplication. Otherwise obligations duplicate or protected work leaks. |
| G7 — projection, renderer, tests | Eight compact Timeline metadata overrides and two narratives, declared order and semantic fallback; exceptions stay centralized and testable. | No field/source-date/limit changes; old blanket layout retired. Otherwise sparse metadata keeps obscuring narrative reading. |
| G8 — implementation, tests, guidance | Exact typed saved states, measured expansion and explicit Clear/Update/Close/Discard hierarchy; presentation cannot coerce authoring intent. | Null remains distinct from empty and exact raw text. Otherwise false absence and unintended payload changes remain. |
| G9 — edit interface, callers, tests | Single owner-admitted Resume, explicit dependency review and leading typed read-only reasons; safety decisions have one owner. | Old two-step callers removed; specialized workflows and grid autosave retained. Otherwise stale authority, duplicate controls or draft retirement can occur. |
| G10 — collection composition, tests | Stable item-local correction with chooser-local read controls; growing collections retain explicit targets. | Source mention/link identity and retained authoring unchanged. Otherwise a command can appear to target the wrong item. |
| G11 — Evidence owner projections, presentation, tests | Preserve metadata/access/authoring separation and distinct stage outcomes; recovery remains precise after partial success. | Explicit same-origin access and existing upload/create/link commands retained. Otherwise counts imply access or accepted stages get replayed. |
| G12 — typed model, shared renderers, tests | Carry semantic historical values to stacked rendering and subordinate diagnostics; new value kinds need no display-string parsing. | Inspector and batch migrate together; source ordering/attribution/reversal untouched. Otherwise historical meaning and safe action scope become ambiguous. |
| G13 — shared contributions, source callers, tests | Feature-keyed command-local forms and outcomes, shared accessible descriptions; placement becomes one cohesive decision. | No-subject creation and owner cancellation retained; trailing forms retired. Otherwise duplicate dispatch or detached command context persists. |
| G14 — implementation, ownership inputs, guides, tests | Retire superseded paths and cover every implemented contract; one maintained renderer and contribution interface. | No invented optional features; retained boundaries have supported consumers. Otherwise fixes diverge across parallel paths. |
| G15 — current validation, documentation, handoff | Consolidate exact source state, automated results and reviewed images; reproducible completion rather than historical inference. | No walkthrough, deployment or usability study. Otherwise stale or incomplete evidence can conceal unresolved work. |

| Phase | Main risk and control | Dependency / exit |
| --- | --- | --- |
| 1 | Authority duplication or incomplete closed-schema migration; compare adopted owners with all generated values and reject invalid overrides. | Initial ledger; owner/projection/generation/lint checks pass. |
| 2 | Focus races, fit oscillation, owner migration; synchronous admission, complete-row measurements and request assertions. | Recorded 1 exit; navigation, geometry, attention and authorization pass. |
| 3 | Value coercion, duplicate submission, premature retirement; exact payload/raw-text and authoring lifetime checks. | Recorded 2 exit; sparse reading, keyboard, recovery and field semantics pass. |
| 4A | Remounting correction or treating partial candidates as absence; keyed collection identity and bounded read tests. | Recorded 3 exit; original-item, retained-work and screenshots pass. |
| 4B | Implicit access or collapsed partial success; explicit access and stage-owner projections. | Recorded 4A exit; access, stages, failure accessibility and screenshots pass. |
| 4C | Lost semantic units or display-driven reversal; closed typed rendering and source-owned actions. | Recorded 4B exit; complete units, scope, recovery and screenshots pass. |
| 4D | Duplicate dispatch or new workflow state; stable feature placement with existing commands/attachments. | Recorded 4C exit; order, adjacency, confirmation and partial-success checks pass. |
| 5 | Premature retirement or missed optional consumer; all-contract tests, source/import inventory and visual reconciliation. | Recorded 4D exit; retirement, applicable digest criteria and consolidated checks pass. |
| 6 | Stale evidence or unsupported completion; final source/manifest review and explicit failure/skip accounting. | Recorded 5 exit; all applicable exits PASS and handoff complete. |

## 11. Advisory rationale and remaining evidence

The user's R01–R09 research dispositions remain informative: R04/R05 support
local context and explicit interpretation; R01/R03 caution against concentrating
state; R08/R09 motivate rendering/editing and owner boundaries; R02 informs
optional coordination; R06/R07 inform the workbook's operational role. The
reports exist under `docs/research`; their full historical analyses were not
re-audited in this update. They establish neither current dependency behavior nor
the selected pixel budgets. No product check may depend on those reports.

Skill dispositions: ADOPT semantic controls, visible focus, local feedback,
independent state, and one design authority; ADAPT compact controls, typography,
and responsive advice to the explicit selected contracts; REJECT a second theme,
dashboard dominance, or generic advice redefining domain behavior. No new
framework, icon dependency, or font is proposed.

Current implementation evidence for owner adoption, projection compatibility,
representative designs, fit measurements, attention, 320px reachability, schema
coverage and unchanged-fixture reading appears in the workstream exits below. The user selected automated checks
and reviewed screenshots as sufficient final validation. No comparative
walkthrough is required and no measured usability gain is claimed. Core 05
separately governs any claim-bearing timed/fixture-sensitive publication.

## 12. Historical planning validation

This update inspected the attachment, unchanged Git baseline, named source and
projection inputs, fixture, verification catalogs, relevant owner paragraphs,
and the dated W3C publications linked above. At that planning stage, this document was the only intended
repository change. No product requirement was adopted by writing it.

Documentation validation: `make lint-markdown` PASS at
`.cartulary/test-results/20260921T181549Z-p15274`, with summary
`adhoc/lint-markdown/tool-run-summary.json`. Correction: that historical run did
not include this handoff path in the configured globs; workstream 1 adds it
explicitly. The previous record reported local link-target, trailing
whitespace, final-newline, and Git scope checks passed. Product suites,
generation, browser capture, golden updates, and agent-finalize are skipped for
this planning-only update. Retained-run maintenance is skipped because
RESULTS_DIR is unset.

## 13. Implementation decisions and evidence

Workstream 1 uses an atomic internal `cartulary.design_presentation.v2` cutover,
retaining the package facade import and removing active v1 inputs. There is no
public discovery, database, saved-view, or persisted-draft migration. Rollback
must revert adopted amendments, authored projections, generated output, and
presentation together; captured requests, receipts, and drafts are not rewritten.

Representative designs use the existing graphite tokens and typography:

| State | Composition |
| --- | --- |
| Reading | Two-line title and Close; measured navigation; one body with aligned property rows and full-width narratives. |
| Editing | Saved value remains visible; Unsaved change directly below its field; value-level Clear, Update beside Close editor, separate Discard. |
| Review required | Original-field cue and owner review below the saved value; no silent baseline acceptance or enabled unreviewed submission. |
| Uncertain operation | Owner-local outcome unknown and captured recovery; distinct newer draft; compact attention entry navigates without dispatch. |
| Read-only | Shared reason before affected fields with described controls; accepted content and permitted retained work remain readable. |
| Constrained | Chooser when complete labels do not fit; property stacking below 320px inner width; 320px work-area overlay; enlarged header may grow. |

Implementation verification and completed workstream exits follow. Evidence is
scoped to the exact checks and source states recorded; historical passes alone
do not establish final completion.

### Workstream 1 exit — 2026-09-21

Amended Design §§3.1.1, 7.3–7.4, 12.4, 12.7, 14.1–14.2; Core 03 §2.3A;
and Core 04 AC-456. Owner/projection comparison confirms the selected values,
exact ten Timeline layouts and authority boundaries agree. Domain vocabulary
requires no change. The representative state compositions above retain existing
tokens; production screenshot review belongs to the implementing slices.

Migrated the closed authored projection and schema to v2, schema attachment,
generator and generated facade together; active digest pointers now resolve to
v2. Added negative validation for duplicate layouts, nonexistent schema/field
references and unknown layout values. The facade import is unchanged. Added
this tracker to Markdown lint and corrected its guide link and historical claim.

| Command | Result / retained run root |
| --- | --- |
| `make generate` | PASS `.cartulary/test-results/20260921T190356Z-p39935` |
| `make json-shape-check` | PASS `.cartulary/test-results/20260921T190418Z-p45051` |
| `make lint-markdown` | PASS `.cartulary/test-results/20260921T190404Z-p42932` |
| `make test-slice OWNER=package.ui ROWS=package.ui.frontend_unit.design_presentation_projection_8e15b85b40` | PASS `.cartulary/test-results/20260921T190540Z-p47089` |
| `make generated-artifact-policy-check` | PASS `.cartulary/test-results/20260921T190540Z-p47055` |
| `make generate-drift` | PASS `.cartulary/test-results/20260921T190606Z-p48148` |
| `make harness-contract` | PASS `.cartulary/test-results/20260921T190725Z-p53439`, including authored-layout rejection cases |

Execution environment correction: the shell did not have `node` on PATH.
Make-owned graph workers require it, so subsequent commands prefix PATH with
the repository-pinned `tmp/node-runtime/bin`. The initial test/policy runs
failed before assertions with `service_start_error` at `20260921T190404Z-p42842`
and `20260921T190418Z-p45049`. A retired smoke target was unavailable, and an
internal `harness-contract-tests` invocation lacked the public suite runtime
(`20260921T190633Z-p52369`); the canonical public `harness-contract` passed.
These are invocation/environment failures, not projection failures.

Rollback is the coherent v2/spec/generated cutover described above. No open
workstream-1 acceptance failures remain. Broader product and screenshot checks
are not evidence for this contract-only slice; they remain required below.


### Workstream 2 exit — 2026-09-21

Implemented measured direct/chooser navigation over one admitted section list.
Exact fit is direct; absent/loading-font measurements fall back to the chooser.
An open chooser survives resizing. Current location, non-overflow and bottom
indication, subject replacement, removed sections, and navigation focus transfer
have focused coverage. Navigation is synchronous and fences stale section
contributions; no pending navigation effect can later steal focus.

The persistent title, visible Close, navigation, and optional attention target
fit the 128px reference budget. Full context is available in the body. Existing
base resize limits and work-area overlays were preserved: measured browser
checks confirm 420px reference width, 320px overlays, inert background, reachable
opener/Close, independent body scrolling, focus reveal, and no document or
inspector horizontal overflow, including 200% zoom and combined text spacing.

Sections now accept immutable owner attention contributions. The ordinary
adapter subscribes to the existing draft and explicit-patch owners for Timeline,
generic, and entity inspectors. It preserves original subject/field destinations,
revalidates owner snapshots on activation, and contributes no announcements or
requests. Captured authoring revisions deduplicate a submitted draft with its
operation; newer authoring remains separate even at the same field. The shell
only admits, orders, deduplicates, counts, renders, and navigates contributions.
Specialized Evidence contributions belong to workstream 4B.

| Command / selected rows | Result / retained run root |
| --- | --- |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_persistent_context,web.workbook.regression.inspector_draft_binding,web.workbook.regression.inspector_draft_lifetime` | PASS `20260921T192720Z-p90087`; final navigation/no-destination/passive-position additions PASS `20260921T193231Z-p39974` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention` | PASS `20260921T192022Z-p98512`; direct/chooser/320px transitions, unchanged draft, zero History reads |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_exact_replay,module.workbook.browser.inspector_persistent_header` | PASS `20260921T192720Z-p90119`; exact uncertain replay, older-operation/newer-draft counts, geometry |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_persistent_header` | Final geometry/28px attention target PASS `20260921T193105Z-p1927`; seven screenshot attachments reviewed |
| `make frontend-typecheck frontend-import-boundary-check lint-biome` | PASS `20260921T192720Z-p90262`, `20260921T192720Z-p90272`, `20260921T192720Z-p90292` |
| `make generate` | PASS `20260921T192853Z-p29683`; authored navigation test routing projected |
| `make format` | PASS `20260921T193213Z-p35372` |

Run roots are under `.cartulary/test-results/`. Screenshot review covered the
three navigation attachments in `20260921T192022Z-p98512` and every
`inspector-header-*` attachment in the final geometry run's Playwright report.
The final seven profiles are 1280×720, 1024×720, 768×640, 320×640, 1280×720 at
200% zoom, and combined spacing at 768×480 and 320×640. Titles, Close, navigation,
focus and work-area containment are readable. No golden has been promoted.

Resolved check failures: `20260921T191324Z-p60808` exposed DOM measurement mocks
on the wrong prototype; `20260921T191458Z-p63163` exposed a browser helper racing
responsive navigation replacement. The helper now awaits the available direct or
chooser path. `make format`/`make lint-biome` initially reported an intentional
DOM-geometry effect dependency (`20260921T192128Z-p48200`,
`20260921T192228Z-p68733`); the narrowly documented dependency suppression retains
label/subject invalidation. All affected checks subsequently passed.

Compatibility: chooser-only assumptions are replaced in touched browser callers;
public capabilities, mounted section content, owner lifetimes, and base resize
limits are unchanged. Rollback this shell/navigation/contribution cutover with
its adopted presentation rules; never rewrite retained work. No open workstream-2
acceptance failure remains. Section-specific composition and final golden
maintenance remain the explicit later slices, not deferred failures here.


### Workstream 3 exit — 2026-09-21

The shared saved-value renderer consumes the eight compact Timeline overrides
and two narrative overrides. Compact rows stack below 320px inner width. Source
strings remain inert and exact; null, empty, whitespace-only, zero, false and
unobserved values remain distinct. Six-line expansion keeps the same editor.
Clear remains an explicit null intent beside the value; Update and Close editor
stay together, with separate Discard. Details-wide typed read-only reasons precede
fields and describe their controls.

The edit owner now contributes one Resume command for unchanged dependencies or
one Review command for changed dependencies. Admission checks original editable
field membership, subject, authority generation and draft revision. Removed
Review-then-Resume callers; specialized authoring remains with its owners.
Captured CRLF authoring, null and empty request payloads are tested independently.
No database, capability, source-string, field-membership or persistence migration.
Rollback renderer and edit-interface callers together without rewriting drafts.

Focused verification (run roots under `.cartulary/test-results/`):

- Five Details/draft/retention unit rows PASS `20260921T194228Z-p54445`;
  final binding/lifetime/entity/saved-reading rows PASS `20260921T195418Z-p50639`.
  Final Timeline explicit-Details row PASS `20260921T195514Z-p57443`.
- Subject retention and exact replay and inspector-edit accessibility PASS
  `20260921T194439Z-p61344`. Expanded browser subject-retention/accessibility
  checks PASS `20260921T194848Z-p7960`: five/six/seven rendered lines, unchanged
  editor identity, raw text, 320px, keyboard recovery and 200% zoom.
- `make frontend-typecheck frontend-import-boundary-check lint-biome` PASS
  `20260921T195419Z-p50901`, `20260921T195419Z-p50905`,
  `20260921T195419Z-p50911`. `make generate` PASS (current slice routing).
- All eight screenshot attachments in `20260921T194848Z-p7960` reviewed:
  sparse reading at reference and 320px, three navigation widths, and three edit
  profiles. The unchanged sparse seed is shared with the visual fixture;
  bytes, order, density, viewport and production renderer are unchanged. The
  first complete raw-activity line is visibly inside the reference body.

Resolved failures: an accidental typed Evidence reason failed type checking
(`193824Z-p47251`) and was reverted; unsorted new catalog titles were sorted;
format/lint rejected a ReactNode truthiness test (`194159Z-p49830`,
`194228Z-p54519`) and the unnecessary wrapper was removed. The shared fixture was
moved from core workbook support to Timeline support after boundary rejection
(`194503Z-p91568`). New payload-test failures (`194848Z-p7938`,
`195418Z-p50639`) exposed a wrong accessible-name assumption and a test trying to
submit without authoring; the test now resumes actual retained raw authoring.
All affected checks pass. No open workstream-3 acceptance failure remains.


### Workstream 4A exit — 2026-09-21

Selected correction and its outcome now follow the original item inside its
actual collection. Active and session-dismissed items share one keyed sequence,
so moving between those states preserves correction controls. Filter, candidate
status, paging and retry are grouped within the target chooser. Removed group-tail
correction placement; source identities, commands and retained authoring remain
unchanged. Rollback only this composition with its tests; no data migration.

Focused verification PASS: candidate/creation review, operation recovery and
selection/undo ownership rows `20260921T195745Z-p66634`; collection inspection and
read-only/item-local node continuity `20260921T195838Z-p1316`; lifecycle browser
`20260921T195746Z-p67049` (dismiss and restore original identity without relinking).
Reviewed both browser attachments, `inspector-relationship-correction` and
`inspector-relationship-dismissed`. Format PASS `20260921T195837Z-p1149` and type
checking PASS `20260921T195837Z-p1129`. Initial collection/type failures
(`20260921T195755Z-p76709`, `20260921T195744Z-p66467`) were a missing test helper
import, corrected before rerunning. No open applicable failure remains.


### Workstream 4B exit — 2026-09-21

Retained the correct accepted-metadata/access/attachment separation and common
picker/drop/paste command. Existing Evidence, Timeline file, and related Evidence
owners now issue explicit attention projections for their unresolved stages.
Owner work identities, source record, stage category and outcome identity remain
separate from component state. The adapter only subscribes, admits the original
subject and navigates; no preview, download, upload or recovery dispatch occurs
on attention activation. File recovery has a focusable original destination.
Accepted creation/link/refresh outcomes remain separate. Count metadata still
confers neither an item catalog nor file access. There is no public or persisted
migration; rollback attention projections and bindings together.

Verification PASS (run roots under `.cartulary/test-results/`):

- Retained file/draft recovery `20260921T200316Z-p19033`; related Evidence
  exact-stage/partial-success/review/authority recovery `20260921T200557Z-p63032`.
- Timeline Evidence panel picker/drop/paste and read-only admission
  `20260921T200559Z-p63886`.
- File stage recovery, existing accepted receipt, preview admission and failure
  accessibility `20260921T200317Z-p19277`; related partial creation and role loss
  `20260921T200624Z-p68761`.
- Final type checking/import boundaries `20260921T200514Z-p61517` and
  `20260921T200514Z-p61521`; format/lint `20260921T200556Z-p62865` and
  `20260921T200556Z-p62855`.

Reviewed all five browser attachments: `inspector-evidence-uncertain`,
`file-recovery-local`, and `evidence-file-recovery-{1280,768,390}`. Source identity,
uncertainty, accepted-refresh failure, original recovery and narrow controls
remain readable. No golden promotion yet.

Resolved failures: lint (`200343Z-p49877`, `200514Z-p61527`) found a test-only
non-null assertion from workstream 4A; replaced it with an explicit fixture
check. A new attention assertion was initially inserted in a completed-success
fixture (`200459Z-p60869`); corrected it to assert no attention after completion,
and checked partial failure in the appropriate fixture. All applicable checks
now pass; no open workstream-4B failure remains. Command-local form composition
is the separately sequenced workstream 4D.


### Workstream 4C exit — 2026-09-21

History now carries the source-owned typed historical values to rendering.
Absent, null, empty text, whitespace, scalar values, empty collections and
collection members remain distinct; no display-string parsing or current-row
comparison is used. All before/after values stack. Optional side-by-side scalar
optimization is intentionally unnecessary; stacking avoids another measurement
path and satisfies the adopted fit restriction. One subordinate event technical
disclosure contains field keys, unit references and record references.

Inspector and batch History review already share this renderer and migrated
atomically. Server order, source attribution, UTC plus numeric offset, event-local
reversal and separate Record actions remain unchanged. Unavailable reads stay
owner-classified. No public/history-data migration; rollback the model and shared
renderer together.

Focused checks PASS: typed semantic detail/model, batch review and retained
History recovery `20260921T200914Z-p7340`; final semantic detail
`20260921T201533Z-p27157`. Type/import checks `20260921T200916Z-p7612` and
`20260921T200916Z-p7616`; final format/lint `20260921T201532Z-p27007` and
`20260921T201532Z-p26997`. Browser change-set reversal, tombstone restoration and
acknowledged-refresh recovery passed in `20260921T200943Z-p9037`; the failed
320px row was repaired and passed `20260921T201353Z-p89374`, then final single
technical-disclosure capture passed `20260921T201535Z-p28116`. Both final
`inspector-history-typed*` screenshots were reviewed.

The 320px check exposed a real shell defect: the hidden full-width navigation
measurement row contributed horizontal overflow (`200943Z-p9037`,
`201142Z-p54749`). A zero-height clipped measurement container now isolates it
without clipping navigation or changing fit measurements. Navigation tests PASS
`20260921T201352Z-p88568`; browser confirms zero overflow. Initial lint
`200916Z-p7622` required formatting and a documented immutable historical-list
ordinal-key exception (source collections can contain duplicates without item
IDs). All affected checks pass; no open workstream-4C failure remains.


### Workstream 4D exit — 2026-09-21

One shared feature-content contribution now places each active form directly
under its originating declared command. Timeline, generic, Entity and Assessment
callers migrated together. Removed trailing panel authoring, superseded Timeline
render callbacks, and repeated per-command outcome paragraphs. Standalone and
no-subject creation remain independent source-owned regions. Feedback carries
original subject and feature identity; retained forms never retarget silently.
Timeline ordinary attention is now forwarded through its presentation region.

Related Evidence uses one source-owned outcome renderer in local Workflow and
global Recovery. Local Resume, Discard and checkpoint commands retain original
attachment/operation identities and precise creation/link/refresh outcomes;
rendering creates no new recovery registration. Confirmation, review, captured
replay and source cancellation remain unchanged. There is no public or persisted
migration. Rollback feature contributions and all callers together.

Verification PASS (run roots under `.cartulary/test-results/`):

- Seven ordered-actions, admission, stale-source, capture and related-recovery
  unit rows `20260921T202428Z-p80040`, including all 17 declared schemas and 48
  capability groups, one command instance and adjacent feature content.
- Timeline seven create-related form/browser cases `20260921T202429Z-p80884`;
  Evidence partial creation and contextual discard `20260921T202520Z-p18592`;
  Assessment contextual refresh and follow-on `20260921T202521Z-p18880`;
  Entity contextual uncertain replay `20260921T202523Z-p19184`.
- Type checking/import boundaries `20260921T202427Z-p79856` and
  `20260921T202427Z-p79860`; final format/lint `20260921T202546Z-p8391` and
  `20260921T202546Z-p8368`.

Reviewed all seven `inspector-workflow-create_related.*` screenshots and the
`inspector-workflow-partial-evidence` screenshot from those browser runs. Forms
follow their actual commands; accepted Evidence creation and rejected original
link remain distinct and locally recoverable. Resolved failures: the old neutral
feedback expectation lacked the newly explicit original subject/destination
(`20260921T202204Z-p76721`); updated the characterization. A formatting failure
after adding subject guards (`20260921T202427Z-p79866`) passed after formatting.
No open workstream-4D acceptance failure remains.


### Workstream 5 consolidation decisions

The public capability registry and all 17 implemented schema contracts remain
unchanged. Required and implemented optional schemas use the same saved reader,
closed shell, typed region/access states, and ordered capability bindings;
unknown additive features remain omitted. No optional surface was invented.
All-schema saved-reading and command coverage run through `listViewContracts()`;
the latter verifies all 48 declared contextual groups exactly once.

| Retired path | Maintained replacement and extension boundary |
| --- | --- |
| Active presentation v1 input/schema | Atomic internal v2 with stable package facade; future layout overrides validate actual schema/field identity. |
| Permanent chooser navigation | One measured section sequence with direct/chooser presentation. |
| Blanket Timeline narrative classification | Projected exact field layouts with semantic fallback; no source-date coercion. |
| Ordinary Review-then-Resume | One source-admitted Resume or dependency review; owner-held drafts remain intact. |
| Group-tail relationship correction | Original keyed item-local correction, including dismissed items. |
| Flattened History value strings | Typed historical values shared by inspector and batch review. |
| Trailing panel authoring and Timeline supplement callbacks | `featureContent` keyed to admitted command identity; no-subject creation retains a real region. |
| Repeated per-command outcome paragraphs | One accessible description per outcome within the command group. |
| Duplicated related Evidence stage JSX | One source-owned outcome renderer reused locally and in Recovery. |

Retained boundaries have real consumers: grid autosave, collection editors,
History operation recovery, Evidence file access, no-subject creation, and global
Recovery keep their source-owned semantics. None is an inspector compatibility
adapter. Ordinary attention now reflects the edit owner's dependency-review
projection; default field-only owners use the draft owner's stale-field result.
Clear explicitly explains “Sets the value to Not set when updated.”

Source guides now describe the final navigation, Resume, layout, typed History,
attention and feature-content interfaces. The authored source ownership manifest
includes the five new source files and repairs existing unlisted Workbook files
and ordering discovered by the full architecture check. Digest pointer updates
and their manual integrity manifest are documentation maintenance only; product
checks do not consume them.

The first complete frontend-unit run (`20260921T203036Z-p28002`, 663/667 graph
units) exposed five failing assertions: multi-ID `aria-describedby`, ownership
inventory drift, selector policy, and a History mock without its required
representation generation. Fixed the assertions/fixtures and authored inventory;
no production validator or selector policy was weakened. Focused architecture,
History helper and shell reruns pass at `20260921T203930Z-p47108`,
`20260921T203930Z-p47147`, and `20260921T203930Z-p47200`.

Ordinary visual reconciliation was inspected before golden mutation:
`20260921T203036Z-p28078/browser-e2e-visual/frontend-visual-reconciliation.json`.
It accounts for 255 active captures and 255 committed goldens, 29 registered
fixtures, zero missing/orphan/ambiguous mappings, and the pinned renderer. Its
only reconciliation error is the unsuccessful comparison target; all failed
rows reached screenshot comparisons. Adopted inspector layout/navigation,
History, Relationships and command-local authoring changes justify refresh.
Viewport, zoom, density, masks, crop scope and scroll anchoring are unchanged.
The sparse Timeline seed bytes and production renderer are unchanged.


### Golden refresh review — workstream 5

`make browser-e2e-visual-update` PASS `20260921T203948Z-p90924` (12/12 units).
Reconciliation PASS: all 255 captures/goldens remain active, all 29 registered
fixtures resolve, and no orphan, missing or ambiguous mapping exists. Exactly
38 changed images were individually opened and reviewed after promotion. The
compact header and direct/chooser modes, sparse reading, item-local correction,
retained editor/draft controls, typed History and source-bound forms match the
adopted change. Focus remains visible; narrow controls and recovery remain within
the scrollport. Captures retain their existing explicit scroll anchors, so a
scrolled body's top/bottom may cut across adjacent content; header and Close
remain available. No font, fixture-byte, viewport, density, masking, crop or
scroll-normalization change was used to obtain a pass.

Promoted golden manifest SHA-256:
`319f859eeaa618f3cb9109b35bf0d372198125ee66b40fe2d9cea71d771d50f7`.
The following exact owner rows and fixture identities come from the reviewed
reconciliation, not filename inference. Filenames are relative to
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`.

| Semantic owner row | Registered fixtures (— means active nonregistry capture) | Changed, reviewed golden files |
| --- | --- | --- |
| `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` | `entity-mention-chip-states-linux.png` |
| `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | — | `record-relationships-mention-chips-linux.png` |
| `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` | `evidence-affordance-states-linux.png` |
| `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector`, `visual.fixture.inspector_compact_actions`, `visual.fixture.inspector_narrow_technical_details` | `workbook-inspector-attached-edit-linux.png`, `workbook-inspector-compact-actions-linux.png`, `workbook-inspector-details-linux.png`, `workbook-inspector-history-linux.png`, `workbook-inspector-narrow-technical-details-linux.png`, `workbook-inspector-public-error-linux.png`, `workbook-inspector-relationships-linux.png`, `workbook-inspector-retained-draft-linux.png`, `workbook-inspector-rollback-preview-linux.png` |
| `module.workbook.visual.contextual_task_decision_creation` | — | `contextual-decision-authoring-linux.png`, `contextual-decision-authoring-narrow-linux.png`, `contextual-decision-references-narrow-linux.png`, `contextual-task-request-authoring-linux.png`, `contextual-task-request-authoring-narrow-linux.png`, `contextual-task-request-references-narrow-linux.png` |
| `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` | `coordination-comm-log-authoring-linux.png`, `coordination-handoff-authoring-linux.png`, `coordination-lesson-authoring-linux.png`, `coordination-source-narrow-linux.png`, `coordination-status-review-authoring-linux.png` |
| `module.workbook.visual.decision_supersession_review_recovery` | — | `decision-supersession-review-linux.png`, `decision-supersession-review-narrow-linux.png` |
| `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` | `indicator-lifecycle-authoring-linux.png`, `indicator-lifecycle-authoring-narrow-linux.png` |
| `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` | `indicator-observation-authoring-linux.png`, `indicator-observation-authoring-narrow-linux.png` |
| `module.workbook.visual.note_create_authoring_recovery` | — | `linked-note-authoring-linux.png`, `linked-note-authoring-narrow-linux.png`, `linked-note-source-narrow-linux.png` |
| `module.workbook.visual.timeline_capture_actions` | — | `timeline-supersession-authoring-linux.png`, `timeline-supersession-review-linux.png`, `timeline-supersession-review-narrow-linux.png` |
| `module.workbook.visual.timeline_related_evidence` | — | `timeline-related-evidence-authoring-linux.png`, `timeline-related-evidence-authoring-narrow-linux.png`, `timeline-related-evidence-party-narrow-linux.png` |


Current consolidated checks also PASS: full `make frontend-unit`
`20260921T204003Z-p30509` (667/667 units); focused final draft attention/review and
binding `20260921T204433Z-p60583`; full `make browser-e2e-a11y`
`20260921T203931Z-p47767` (20/20 units); full `make browser-e2e-measurement`
`20260921T203930Z-p47617` (25/25 units); `make lint-scripts`
`20260921T204301Z-p30361`. These are implementation-support results, not a new
Core 05 performance or usability claim.

Final subject retention, persistent header and exact replay passed in
`20260921T204003Z-p30367`. Its Notes association recovery row exposed a test race:
read service recovery was enabled while the deliberately failed refresh was
still running. Waiting for the owner-admitted Retry refresh button before
restoring reads fixes the fixture without altering recovery behavior. That row
passes `20260921T204503Z-p70805` (11/11 units). Ordinary visual confirmation runs
remain required before the workstream-5 exit.


### Digest acceptance assessment — workstream 5

Evidence roots below refer to the current workstream exits. The complete
frontend unit run is `20260921T204003Z-p30509`; accessibility is
`20260921T203931Z-p47767`; measurement is `20260921T203930Z-p47617`.
Rows apply only to this inspector/presentation scope; no unimplemented optional
surface, release deployment or whole-page conformance claim is implied.

| Digest row | Status | Evidence and assessment |
| --- | --- | --- |
| A001 Authority | PASS | Amendment map, adopted Design/Core 03/Core 04 changes and v2 owner/projection comparison; runtime and routing remain downstream. |
| A002 Scope | PASS | Gap rationale and retirement table; navigation, saved rendering, typed historical values and command placement hide concrete shared decisions. All callers migrated. |
| A003 Repository state | PASS | Recorded main/HEAD and original staged tracker; source guides, source/import manifests and generated roots inspected; bounded digest pointer refresh explicitly qualified. |
| A004 Tokens | PASS | Existing typography, graphite, spacing and control tokens; adopted inspector dimensions/layouts projected through v2. No new theme/token registry. |
| A005 Theme | PASS | Existing dark_graphite fixtures retained; reviewed golden changes concern inspector composition only. |
| A006 Density | PASS | Shared density unchanged; all-schema saved reader and production geometry/measurement plus visual density fixtures retain their original profiles. |
| A007 Creation | PASS | All 48 declared contextual groups, seven Timeline workflows, Entity/Assessment/Evidence contextual browser cases; accessibility covers specialized source retention and permission loss. |
| A008 Responsive | PASS | Workstream-2 seven-profile geometry plus current persistent-header row; base/overlay distinction and inner-width fallback tests in full frontend suite. |
| A009 Overflow | PASS | One inspector scrollport, inert overlay, 320px reachability, independent grid scrolling; navigation measurement wrapper excludes hidden row width. |
| A010 Inspector | PASS | All 17 contracts/48 groups, exact semantic dispatch, permission reasons, omitted additive features, retained review and source/authority invalidation in unit/browser rows. |
| A011 Continuity | PASS | Subject retention, unchanged editor identity, raw authoring, field focus, removed sections, newer navigation and captured late results; current subject/replay browsers pass. |
| A012 Transactions | PASS | Existing secure identity policy and explicit patch tests; uncertain browser replay compares exact captured bodies and preserves newer authoring. No new request engine. |
| A013 Acknowledgement/recovery | PASS | Queue and explicit operation unit suites, History read-only refresh, Evidence separate stage recovery and Notes exact replay/refresh rerun. No accepted write is resent by navigation or refresh. |
| A014 Editing | PASS | Safe Resume/dependency review, explicit null versus empty payloads, exact CRLF, six-line expansion, source-string preservation and accessible recovery; grid autosave suites retained. |
| A015 Conflict | PASS | Existing local conflict/draft distinction and resolver tests; visual/a11y conflict fixtures remain unchanged; older receipts do not retire newer drafts. |
| A016 Independent state | PASS | Region/access matrix unit tests, Notes readable sibling behavior, Evidence metadata/access split, stale/unavailable and concealed producer tests in full frontend suite. |
| A017 Authority scope | PASS | Draft lifetime/account replacement, source owner concealment, same-account retention, incident revocation and late-response fencing in full frontend and specialized recovery tests. |
| A018 Evidence | PASS | File lifecycle/access browsers and full Evidence a11y/visual matrix; owner-supplied attention preserves distinct upload/create/link/refresh outcomes. |
| A019 Accessibility | PASS | Full public accessibility target; focus and keyboard recovery, contrast checks, 200% zoom, 320px inspector and combined text-spacing fixtures. Product 28px targets remain distinct from WCAG minimum. |
| A020 Components | PASS | Full frontend variants, required and implemented optional schemas, long-title/value expansion, supported geometry and read-only controls; state precedence is owner-admitted. |
| A021 Virtualization | PASS | Production Timeline and Network Flow measurement target plus Grid Adapter/producer continuity tests; virtualization and row ownership were not changed. |
| A022 Visual fixtures | PASS | All 38 promoted changes individually reviewed; both fresh ordinary 255-capture runs and reconciliations pass against the promoted manifest. Exact roots below. |
| A023 Selectors | PASS | Shared UI selector package tests and repaired architecture selector policy; tests select semantic field, schema, record and capability identities. |
| A024 Test authority | PASS | Source/dependency audit and full architecture suite; executable contracts and generators read authored machine inputs, never this tracker or Markdown. |
| A025 Generated artifacts | PASS | Authored v2/schema/attachment/generator changes; generated facade emitted through Make; final JSON/policy/drift checks pass. Visual manifest emitted only by public update target. |
| A026 Compatibility | PASS | No public discovery, API, database, saved-view or draft migration. Active v1 and superseded callbacks retired; all retained boundaries have supported consumers. Coherent rollback recorded. |
| A027 Handoff | PASS | All nine workstreams completed in order, each exit recorded before its dependent. Terminal finalizer, Markdown, source fingerprint, validation, compatibility, rollback and skipped-check evidence are complete below. |


### Workstream 5 exit — 2026-09-21

Consolidation is complete. Both fresh ordinary `make browser-e2e-visual` runs
PASS (12/12 units each): `20260921T204856Z-p18340` and
`20260921T205008Z-p53536`. Each reconciliation accounts for all 255 captures and
matches promoted manifest SHA-256
`319f859eeaa618f3cb9109b35bf0d372198125ee66b40fe2d9cea71d771d50f7`.
No further golden promotion or tolerance change occurred.

Final source inspection added wrapping to the exact-whitespace disclosure, so
escaped tabs/CRLF cannot create horizontal overflow. A separate production
Timeline row with 300 repeated tab/CRLF sequences passes exact-value and 320px
width assertions in `20260921T205223Z-p98340`; the sparse fixture is untouched.
This bounded style change is absent from the visual fixtures and changes no
reviewed golden; its changed state is covered by the focused browser run.

Final checks PASS: type checking `20260921T205246Z-p27229`, lint
`20260921T205246Z-p27251`, Markdown `20260921T205246Z-p27261`, import boundaries
`20260921T204954Z-p52843`, JSON shape `20260921T204954Z-p52785`, generated policy
`20260921T204954Z-p52783`, drift `20260921T204954Z-p52781`, and the public
`make test-catalog-check` (exit 0). Local links and `git diff --check` pass.
Full frontend, accessibility, measurement, reviewed visual refresh and specialized
browser evidence are recorded above. A001–A026 are PASS; A027 is the explicit
terminal handoff obligation in workstream 6, not a deferred product failure.

No production caller uses the retired paths. Ownership stays with existing
source modules; no unsupported compatibility adapter or optional feature was
introduced. Remaining limits are the declared inspector-only 320px scope,
existing session-bound draft lifetime, and absence of participant/performance
publication claims. None blocks the adopted acceptance. Rollback remains a
coherent presentation/spec/projection/fixture cutover without rewriting retained
work or receipts. No open workstream-5 failure remains.


## 14. Terminal handoff — workstream 6

Final source baseline is `main` at
`a4ef612aea434696361e9b3d5683f72eecd0180b`, plus this uncommitted implementation.
The original staged tracker remains staged; implementation changes remain in the
working tree. There are 127 changed paths, including 38 promoted PNGs and the
tracker. No dependency lockfile, public capability registry, API contract, SQL,
database migration or persisted draft format changed.

The 115 changed non-Markdown/non-`docs/` paths have a combined SHA-256 fingerprint
`1ce6e8e966d38dc708ca083b8f330b96103b22b16e65a9453cfb92be56688bb2`.
For reproducibility, sort the changed tracked and untracked paths, exclude
`docs/**` and `*.md`, then hash each `path + NUL + SHA256(bytes) + LF` record;
use `DELETED` in place of the digest for removed paths. This manual handoff
fingerprint is not an executable product or release requirement.

### Changed areas and inspected boundaries

- Adopted owners: `docs/design.md`, Core 03 §2.3A and Core 04 AC-456. Inspected
  `docs/domain.md` and Core 01 without changing their vocabulary/capability
  boundaries. Selected dimensions, field layouts, navigation, Clear, attention
  and accessibility now have governing owner clauses.
- Authored projection/schema/generator: `contracts/design/presentation.v2.json`,
  `tools/schemas/cartulary.design_presentation.v2.schema.json`, schema attachment,
  design generator and generated UI facade; active v1 inputs removed.
- Shared implementation: `apps/web/src/workbook/inspector/**` and
  `layout/workbookInspectorNavigation.ts`; compact shell/navigation, owner
  attention, saved reading/editing, typed History and feature-local authoring.
- Source adapters: Timeline and generic, Entity, Assessment and Evidence
  compositions. New Evidence attention/local-outcome modules stay under the
  Evidence owner; ordinary attention stays in inspector binding. Authorization,
  requests, authoring, receipts and recovery stay with their existing owners.
- Tests/maintenance: affected unit/browser fixtures, unchanged shared sparse
  Timeline seed, 38 reviewed goldens and generated manifest, authored frontend
  source inventory/routing, source guides, exact Markdown lint path and bounded
  digest pointer/integrity maintenance. All changed paths are reviewable in the
  working-tree diff; the golden table above lists every changed image.

### Validation closure

| Required evidence | Result and current retained evidence |
| --- | --- |
| Owner/projection fidelity | PASS — adopted amendments, representative designs, exact override validation, generation and harness checks in workstream 1. |
| Focused implementation slices | PASS — navigation/geometry/attention, Details/authoring, Relationships, Evidence, History and Workflow exits recorded before dependents. |
| All frontend units | PASS — `make frontend-unit`, `20260921T204003Z-p30509`, 667/667 execution units. Final attention assertions additionally pass `20260921T204433Z-p60583`. |
| Browser behavior/security/recovery | PASS — per-slice source-context, no-request navigation, exact replay, partial-success, read-only and concealment rows. Final Notes recovery `20260921T204503Z-p70805`; whitespace/320px and subject retention `20260921T205223Z-p98340`. |
| Accessibility | PASS — `make browser-e2e-a11y`, `20260921T203931Z-p47767`, 20/20 units, including keyboard, contrast and inspector recovery. |
| Measurement | PASS — `make browser-e2e-measurement`, `20260921T203930Z-p47617`, 25/25 units. No analyst usability or new publication claim. |
| Visual maintenance | PASS — ordinary reconciliation inspected; update `20260921T203948Z-p90924`; every changed PNG reviewed; ordinary passes `20260921T204856Z-p18340` and `20260921T205008Z-p53536` against the identical promoted manifest. |
| Type/import/lint | PASS — final type `20260921T205246Z-p27229`, import `20260921T204954Z-p52843`, Biome `20260921T205246Z-p27251`, scripts `20260921T204301Z-p30361`. |
| Authored/generated integrity | PASS — JSON `20260921T204954Z-p52785`, generated policy `20260921T204954Z-p52783`, drift `20260921T204954Z-p52781`, catalog exit 0; public generator/contract evidence recorded above. |
| Documentation | PASS — Markdown `20260921T205246Z-p27261`, exact tracker included; local links and diff whitespace reviewed. Terminal text is checked again before final exit. |

`make agent-finalize` ran before broad verification and passed at
`20260921T203003Z-p23665`; its `unit-artifacts/finalize-summary.json` reports
unchanged generated structure and no failures. RESULTS_DIR was unset because no
qualifying full warm `make check` run was supplied. Retained-run maintenance and
performance-evidence maintenance were therefore explicitly skipped, not inferred
from a narrower pass. The terminal finalizer result is recorded in the exit.

All observed applicable failures are recorded at their originating exits with
repairs and successful reruns. The initial ordinary visual comparison failure was
an expected, reviewed consequence of adopted redesign changes, not an unresolved
functional failure. No acceptance check was converted into deferred work.

### Compatibility, rollback, limits and skipped checks

The internal projection cutover is atomic and keeps its package facade import.
Public discovery, required/optional schema admission, field membership, source
string limits, stored rows, saved views and draft/request/receipt identities are
unchanged. There is no migration task. Deploy or roll back the owner amendments,
authored projections, generated outputs, source callers, tests and golden manifest
as one coherent change; never rewrite retained drafts or captured operations to
fit a presentation rollback.

The supported expansion is inspector-only 320px reflow. The base resize minimum
remains 360px; enlarged text may exceed the reference header budget. Session-held
work does not gain reload or cross-tab persistence. Unavailable historical data
remains owner-classified; semantic identity changes remain typed historical
values while diagnostic event/unit/field references are subordinate. Stacked
History values deliberately avoid a second scalar-fit measurement path.

Full backend `make check`, `ci` and release/deployment gates were not run: this
change does not modify backend/database behavior, and the broad frontend/browser,
contract/generator and architecture gates cover the changed owners. No release,
whole-page WCAG, Core 05 benchmark or measured usability certification is claimed.
No walkthrough, participant evaluation, deployment or usability study is required
by the user's selected final gate. No optional capability was invented to improve
coverage counts. No required remediation is assigned to a future workstream.


### Workstream 6 exit — 2026-09-21

Terminal `make agent-finalize` PASS `20260921T205621Z-p35087`; generated
structure is unchanged, failures are empty, and retained-run maintenance is
explicitly skipped because RESULTS_DIR is unset. Terminal `make lint-markdown`
PASS `20260921T205734Z-p39250`, including this controlling tracker. Final local
link resolution and `git diff --check` PASS. Recomputed non-documentation source
fingerprint matches the handoff value above; finalization changed no source.

All G1–G15 remediations and all nine workstreams are complete. Every applicable
binary acceptance area and digest row A001–A027 is PASS. No applicable blocked or
failed criterion remains, and no required implementation, validation, walkthrough
or handoff work remains. The reviewable working tree and this tracker are the
completed handoff; no deployment or publication is part of this effort.
