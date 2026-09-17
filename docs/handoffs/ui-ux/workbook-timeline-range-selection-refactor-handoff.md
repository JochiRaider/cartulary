# Timeline cell-range selection and gesture ownership

## Execution control

Implementation starts on clean `main` at
`369b35231a54c4970a3a4977073aace58f291b4d`, one commit ahead of
`origin/main`. Revalidated with `git status --short --branch` and
`git rev-parse HEAD`; no pre-existing modifications. Preserve existing commits.
No digest changes, commit, push, deployment, analyst-data changes, dependencies,
public APIs, persistent storage, or database/saved-layout migration are included.

| Workstream | Status | Dependency | Exit |
| --- | --- | --- | --- |
| TRS-01 Characterization and interaction contract | DONE | None | Core 03 §13.4/design amendments adopted; existing Adapter 50/50 and projection 10/10 units pass. |
| TRS-02 Shared selection and gesture ownership | DONE | TRS-01 DONE | Shared controller/pointer binding tests and type checks pass; browser geometry remains TRS-04. |
| TRS-03 Timeline integration and retirement | DONE | TRS-02 DONE | Production composition, editor admission, operations and retired paths verified. |
| TRS-04 Production-browser evidence | DONE | TRS-03 DONE | Applicable browser, spreadsheet, lifecycle and accessibility evidence passes. |
| TRS-05 Terminal validation and completed handoff | DONE | TRS-04 DONE | Terminal checks, gap ledger and acceptance assessment complete. |

Save only the current row as IN_PROGRESS before its work. Append actual evidence
and save DONE before starting its dependent. An applicable blocked dependency
prevents dependent implementation and completion. Historical next actions remain
dated evidence; the final status table controls current completion.

## Authority and bounded scope

Core 03 §§2.3A, 3–4, 11, 13–14 and REQ-03-218/297/298/300 own interaction,
retention, clipboard and mutation admission. Core 01 owns stable identities,
field capabilities, mutation contracts and saved layouts. Core 04 owns current
authorization, protected-content handling and formula-safe export. Design §§8/14
owns interaction presentation; domain owns vocabulary/navigation. The research
NLSpec essay and localized digest are advisory, not additional user requests.
The localized read order was followed during planning. Grid-edit/autosave,
query-continuation, clipboard-fidelity, paste/bulk-recovery, column-sizing and
recovery-navigation handoffs were inspected as historical evidence only.

Implementation scope: Grid Adapter semantic policies, production RDG binding,
interaction DOM and styles; Timeline composition/presentation and existing
editing/continuity/clipboard/fill bindings; narrow owner/projection changes,
source guides and owner-routed tests. Other Workbook and Network Analysis grids
are regression boundaries. No disjoint selection, row/column formatting, data
dragging, bulk-clear, formulas, query expansion or new touch profile.

## Adopted planning decisions

- Four CSS pixels on either axis is the stationary tolerance; crossing it once
  permanently classifies a drag. A stationary release edits immediately without
  a timer. Readable read-only cells can be selected.
- Shift-click/drag uses the valid range anchor, else valid active cell, else the
  destination. Shift changes during a gesture do not change that chosen anchor.
- Ranges cross expanded groups while excluding headers and collapsed records.
  Hidden fields and recordless drafts are not members. Grouped fill stays banned.
- Completed selection, tentative gesture, active cell, inspector subject and
  bulk checkbox selection have separate lifetimes. Completed range survives
  release. Range completion moves active focus to its endpoint and preserves
  inspector context. Ordinary clicks preserve established inspection behavior.
- Cancellation restores the previous still-valid completed selection. Escape
  cancels a tentative gesture without discarding an editor draft; otherwise the
  existing Escape hierarchy applies.
- Pending editor departure retains editor focus and shows a dashed tentative
  outline. Edge scrolling waits for acceptance. Only the latest still-valid
  gesture can publish after acceptance; rejection preserves the exact draft.
- Edge band 32 CSS pixels; linear maximum 720 CSS pixels/second; elapsed frame
  clamped to 32 ms. One scroll write/endpoint resolution per animation frame,
  existing grid scrollport and loaded membership only, no page requests.

## Characterization and gap ledger

Existing protections are present: Shift+Arrow produces GridCellRange; ranges
survive modifier release; retainGridCellRange compares exact ordered membership;
clipboard planners use semantic identities and owner-approved representations;
edit sessions deduplicate commits and preserve rejected raw drafts; Timeline
fill validates one scalar column, committed IDs/versions and ungrouped mode.

| Gap / classification | Remediation and affected areas | Rationale and durable benefit | Compatibility / unresolved risk | Binary criterion |
| --- | --- | --- | --- | --- |
| G1 New interaction | Adopt contiguous pointer gestures in Core 03/design; project constants; enable Timeline capability. | Natural pointer/keyboard parity with explicit scope. | Additive internal capability; threshold/zoom validated in production. | Drag and Shift-click select the same semantic rectangle as keyboard; stationary click edits once. |
| G2 Structural weakness | Consolidate capture, mouse-up/timer, vendor click and range-collapse decisions in Adapter. | One owner admits selection and editor departure. | Other grids retain supported clicks/keys; compatibility-click regressions pass. | No drag-triggered editor/fill; no duplicate click/commit; embedded controls keep ownership. |
| G3 Structural risk | Fence destinations by interaction, scope and retained membership; reuse editor/operation owners. | Late results cannot steal focus or substitute selection targets. | No draft/operation migration; delayed rejection/acceptance/supersession validated. | Pending selection is not accepted; rejection retains exact draft; obsolete completion cannot replace newer selection. |
| G4 New interaction | Adapter-private bounded pointer capture, hit testing and edge scrolling. | Virtualized selection without query coupling. | No new paging/touch profile; clipped geometry, capture and teardown validated. | Both axes traverse loaded cells with bounded frame work, no requests/document scrolling, and complete cleanup. |
| G5 Presentation extension | Distinguish tentative/completed range, active focus and inspector; concise announcements. | Selection remains understandable and accessible. | Existing density/state precedence retained; zoom/spacing reviewed. | Non-color cues and one completed-dimension announcement; inspector/bulk state unchanged by ranges. |
| G6 Regression boundary | Feed existing clipboard/fill planners and retain sizing/recovery ownership. | Shared selection changes no mutation authority. | No wire or persistence migration; native controls and operation recovery pass. | Pointer/keyboard copy/fill parity and spreadsheet/sizing/recovery regressions pass. |
| G7 Observed presentation defect | Preserve compiler-owned range classes during semantic state marking in rdgCompiler. | One state layer cannot erase another layer's range cues. | Shared range presentation restored; no new state precedence. | Production tentative dashed outline and completed hatching/border remain visible after registration. |
| G8 Observed contract leak | Project identity-only range anchors in semanticSelectionPolicy. | Editor mutation metadata never becomes selection identity; current operation versions stay source-owned. | Internal shape correction; no wire/data migration. | Editor-origin Shift extension equals the identity-only keyboard/pointer range. |
| G9 Observed integration defect | Use existing semantic focus requests for admitted endpoints even when RDG elides same-position selection. | Clicks reliably restore focus after external actions. | Necessary shared repair, including readable Indicator cells; no new focus owner. | Production binding and Indicator late-recovery browser refocus the already-active cell. |
| G10 Observed integration defect | Route native draft focus through the interaction owner and exclude draft aria-selected membership. | Editor entry retires range consistently without giving recordless drafts stable membership. | No change to first-qualifying-input creation or native text ownership. | Focusing the draft clears accepted range, exposes no selected draft cell and creates no record. |

## Evidence log

Planning preflight (not implementation acceptance):
`make help`, `make help-all`, and task guides for `package.grid_adapter`,
`web.workbook`, and `module.workbook` passed. The selected keyboard,
clipboard/fill and window-retention slice passed 4/4 graph units at
`.cartulary/test-results/20260917T161900Z-p3829129`. `git diff --check` passed.

TRS-01 implementation inspection confirms the same baseline. In
SemanticDataGrid, mouse-down capture installs a window mouse-up callback and
zero-delay timer; vendor mouse-down independently requests editor departure;
vendor click publishes range/inspection before checking pending departure;
vendor selected-cell callbacks also collapse ranges. RDG's default mouse-down
selects a cell and default double-click requests editing. These paths require
one admission owner, not another retained selection model.

## Acceptance, limitations and rollback

All applicable rows require PASS; N/A requires a scope/owner rationale. Browser
acceptance is established by the production runs recorded below, not fake-grid
callbacks, assigned range state or unit geometry. Terminal validation controls
final completion.

Rollback restores this slice's owner text, authored projections, generated
derivatives, implementation, tests and source guides together. Preserve accepted
analyst writes/history and unrelated work. Retained-run maintenance is skipped:
RESULTS_DIR was unset because this execution has no successful full warm-check run.

### TRS-01 exit

Core 03 §13.4 and REQ-03-300 now adopt the authorized gesture/lifetime contract.
Design §8.6 and §14.2 distinguish gesture completion from range retirement,
pending outlines, Escape and announcements. The existing design JSON/schema and
generator project all four bounds. No owner contradiction remains. No HTTP,
field, layout or authorization contract was changed.

- `make test-slice OWNER=package.grid_adapter`: PASS 50/50 units,
  `.cartulary/test-results/20260917T162441Z-p3831330`.
- `make generate`: PASS,
  `.cartulary/test-results/20260917T162615Z-p3869339`.
- `make test-slice OWNER=package.ui`: PASS 10/10 units,
  `.cartulary/test-results/20260917T162646Z-p3872436`.

The baseline suite characterizes single-click edits, acceptance/rejection,
native editor clipboard, semantic keyboard ranges, fill and retention. Browser
geometry/new gestures are not claimed. Next action: implement TRS-02's pure
interaction owner and DOM binding, with current membership and lifecycle tests.

### TRS-02 exit

Added the private gridInteractionController, gridInteractionDom and React binding.
Pointer and keyboard share extendSemanticCellRange; the optional semantic
capability is explicit. Gesture state captures loaded identities and coordinates
without owning drafts or mutation outcomes. Frame work, capture, native-control
exclusion, compatibility-click consumption and cleanup remain inside Adapter.
Current clipped geometry reuses column sizing's existing viewport calculation.

The first selected run was rejected before tests because authored catalog rows
were not ASCII-sorted; corrected the authored input, without weakening checks.
The final selected controller/DOM-binding run passed 3/3 graph units (11 cases),
`.cartulary/test-results/20260917T163556Z-p3878913`. Cases cover threshold/reversal,
pointer/keyboard anchor parity, pending rejection/acceptance/supersession, stable
membership, authority/scope invalidation, native controls, capture cleanup,
outside release and both-axis frame limits. Unit geometry is not browser proof.

`make frontend-typecheck` passed 2/2 units at
`.cartulary/test-results/20260917T163615Z-p3879665`.
TRS-01's `make lint-markdown` passed at
`.cartulary/test-results/20260917T162652Z-p3873930`.
Next action: TRS-03 production integration and competing-path retirement.

### TRS-03 exit

Production Timeline now enables contiguous selection using incident continuity,
sheet and accepted canonical query identity. Pending query inputs do not enter
the key. Existing GridCellRange feeds existing clipboard/fill planning, and
editor requestCommit remains the only departure write. Range completion changes
active focus without invoking row inspection or checkbox selection. No continuity
selection store was introduced. Updated Adapter, Timeline and continuity guides.

Retired the adapter mouse-down/window mouse-up timer, independent vendor click
activation, pending vendor range endpoint, and unconditional vendor range collapse.
Explicit focus/navigation replaces selection; focus on its endpoint preserves it.
Native editor, embedded actions, fill and column-sizing paths retain their owners.

Integration exposed two observed issues, resolved within G5/G6: semantic state
marking stripped the compiler's range classes, and an editor's versioned target
could leak mutation metadata into a Shift range anchor. State marking now preserves
range presentation; selection policy emits identity-only anchors, leaving current
versions to operation admission. Binary validation: tentative/completed classes
survive cell registration and editor-origin Shift ranges equal keyboard anchors.
Compatibility: no mutation/wire migration. No unresolved unit-level risk; browser
geometry and application lifetime remain TRS-04 acceptance.

The first integrated adapter run failed 1/52 units at
`.cartulary/test-results/20260917T164001Z-p3881474`; follow-ups diagnosed stale
mouse-down-only fixture assumptions and explicit focus anchor replacement.
Tests now await semantic focus and exercise completed clicks with fresh mounted
cells. No old mouse-down path was restored. New production RDG unit evidence
covers pending outline, exact rejection retention, deduplicated commit and latest
destination acceptance while inspection remains unchanged. The final focused
index/controller/binding run passed 4/4 units at
`.cartulary/test-results/20260917T165626Z-p4054530`.

- `make test-slice OWNER=web.workbook`: PASS 271/271,
  `.cartulary/test-results/20260917T164908Z-p3967318`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260917T165437Z-p4011288`.
- Formatting initially found five new lint errors and one warning; corrected
  dependencies, callback returns and nullable fixture access. `make format`
  passed at `.cartulary/test-results/20260917T165119Z-p4004487`.
- Attempted BIOME_CHECK_FLAGS override was correctly rejected as unsupported;
  no harness policy was changed. A truncated row ID was rejected before running;
  subsequent slices used exact authored IDs.

Next action: TRS-04 real pointer/keyboard browser evidence and scoped regressions.

### TRS-04 exit

Nine routed scenarios in `apps/web/e2e/timeline-range-selection.spec.ts` exercise
production Timeline with real mouse and keyboard input. Native cancellation is
injected only after starting a real drag; unexpected lost capture uses the
browser's releasePointerCapture. No test assigns production range state or
disables virtualization. The scrolling scenario seeds 115 isolated records,
traverses the loaded 100-record boundary and both virtualized axes, and observes
zero gesture-triggered query requests or document scrolling. Pending acceptance
suspends scrolling in controller evidence; release, cancellation and unmount stop
frame work. Pointer listeners are removed on release, while pending acceptance
retains only the lifecycle cancellation boundary it still needs.

The completed range survives compatible value/version refresh and failed refresh;
accepted sort replacement, deletion, hidden/reordered membership, collapse and
authority changes invalidate it. Grouped gestures follow visible committed order
across both buckets without admitting group headers or grouped fill. Readable
closed-incident cells select without writes. Bulk checkboxes and inspector subject
remain independent. Native editor text drag and clipboard retain ownership.

The browser run found and resolved integration gaps beyond the new gestures:

- One-cell ranges now receive the same membership and dimension feedback as
  other rectangles; active-cell focus remains a separate cue.
- RDG can elide selecting an already-active position after focus leaves it.
  The admitted destination now uses the existing semantic focus request owner.
  New production-binding regression and the affected Indicator browser pass.
- Entering the recordless native draft now clears completed selection through
  the interaction owner; draft cells never advertise accepted range membership.
  Focusing that input creates no record. Creation and keyboard regressions pass.
- Old browser fixtures used synthetic mouse-down and forced focus. All such
  committed-cell setup paths now share activateCommittedGridCell: a real click,
  cancellation of its fresh editor where applicable, and observed focus. The
  fixture change adds no product event path and weakens no focus assertion.

Earlier browser failures were diagnosed, not accepted as unrelated: clipped
midpoint targeting was replaced with current visible geometry; grouped expected
counts now reflect all visible committed members; row actions use their existing
menu; immediate rejection may cancel before any tentative outline can be observed.
A server-rejected queued edit retains its existing blocking recovery semantics;
independent acceptance/cancellation is tested before that rejection.

Fresh browser evidence (public service-backed owner slices; full row accounting
and Playwright reports reside below each run root):

| Coverage | Result | Run root under .cartulary/test-results/ |
| --- | --- | --- |
| All eight initial new range scenarios, including cancellation, grouping, zoom and authority | PASS 11/11 graph units | 20260917T171728Z-p267553 |
| Added live membership/deletion and strengthened modifier/native text/early acceptance cases | PASS 11/11 | 20260917T171934Z-p302933 |
| Final native-draft boundary, range departure, creation and spreadsheet keyboard | PASS 13/13 | 20260917T172946Z-p516637 |
| Existing pointer click, keyboard and fill scenarios | Three scenario PASS rows in the initial new-gesture run | 20260917T170013Z-p4088808 |
| Clipboard fidelity, paste/bulk conflicts, headers, replay, creation, rejection, refresh and virtualization | PASS 11/11 | 20260917T170924Z-p83702 |
| Column sizing gesture, surfaces, viewport and accessibility; Timeline continuation | Five scenario PASS rows; keyboard setup failures subsequently repaired | 20260917T170849Z-p24167 |
| Shared keyboard anchors and application shortcuts; service ingest/cursor admission | PASS rows; long-history setup subsequently repaired | 20260917T171114Z-p218415 |
| Long checkpoint history, reference popup, role/access loss and default Timeline visual | PASS 15/15 | 20260917T171954Z-p331954 |
| Network Analysis production interaction, accessibility and protected-state lifecycle | PASS 15/15 | 20260917T170925Z-p84690 |
| Timeline default, grouped and edit/save/conflict visual rows | PASS 11/11 | 20260917T171054Z-p161298 |
| Inspector refresh focus, representative paste, public-route recovery and default-closed inspector | PASS rows; Indicator refocus subsequently repaired | 20260917T172121Z-p375848 |
| Indicator late recovery after the admitted refocus repair | PASS 11/11 | 20260917T172400Z-p445958 |
| Paint-qualified typing, pointer focus and ArrowDown measurement rows | PASS 18/18 | 20260917T172121Z-p375867 |

The initial unbounded module.workbook owner selection included many unrelated
browser groups. It was interrupted in favor of explicit rows; its cancelled run
at `20260917T165455Z-p4012547` is not acceptance evidence. Interrupted raw trace
retention also failed its secret-capable-syntax scan. No trace content was copied
into this handoff. Applicable failures were revalidated through the focused runs
above, including the checkpoint-history fixture's remaining synthetic mouse-down.

Reviewed the production 100% and 200% zoom/text-spacing images attached to the
range accessibility scenario in `20260917T170547Z-p4176187`. Dense geometry,
readable text, hatching and outlines, distinct endpoint focus and horizontal
scrolling remain intact. The ordinary visual rows passed without golden changes;
no refresh or unrelated golden regeneration was performed. Screenshot review is
implementation evidence, not a Core 05 claim.

Final focused Adapter controller/binding/index slice: PASS 4/4 at
`20260917T172944Z-p516371`. Full Adapter slice passed 52/52 at
`20260917T170515Z-p4141999`; terminal validation will repeat it after finalization.
All applicable TRS-04 families have passing evidence. Touch/pen interaction,
unloaded/query-expanding selection, disjoint ranges and new bulk operations are
N/A by the explicit scope exclusions. Next action: finalizer, terminal owner
checks, documentation and generated-drift validation, then final scope review.


### TRS-05 terminal evidence

Refreshed `make help`, `make help-all` and owner task guides before terminal
selection. Routing starts at package.grid_adapter, web.workbook and
module.workbook, with module.timeline, package.ui, web.architecture and Network
Analysis owners added for changed/shared boundaries. Service-backed rows ran in
TRS-04 before finalization; these commands are owner-directed slices, not a claim
that every repository test or every Core acceptance claim was exercised.

`make agent-finalize` initially failed at JSON shape validation in
`20260917T173249Z-p553970`. Direct `make json-shape-check` at
`20260917T173605Z-p555266` identified stale generated topology metadata after the
added Adapter regression title. `make generate` passed at
`20260917T173627Z-p555814`; the rerun finalizer passed at
`20260917T173646Z-p558972`. Its `unit-artifacts/finalize-summary.json` records
schema, catalog/tier and generated-drift PASS, no unrelated mutations, and skipped
retained canonical/scheduler maintenance because RESULTS_DIR was unset.

Final review clarified design's Escape wording: a tentative gesture consumes the
key without applying editor cancellation. This matches adopted Core 03 and the
implemented/tested behavior. No owner contradiction remains. Source guidance now
also names native draft entry and the already-active endpoint focus obligation.

| Command / selection | Result | Run root under .cartulary/test-results/ |
| --- | --- | --- |
| make test-slice OWNER=package.grid_adapter | PASS 52/52 | 20260917T173713Z-p563029 |
| make test-slice OWNER=web.workbook | PASS 271/271 | 20260917T173713Z-p563088 |
| make test-slice OWNER=web.architecture | PASS 12/12 | 20260917T173713Z-p563105 |
| make test-slice OWNER=package.ui | PASS 10/10 | 20260917T173713Z-p563154 |
| make test-slice OWNER=module.timeline, all 24 authored Vitest rows | PASS 25/25 | 20260917T173823Z-p615757 |
| make test-slice OWNER=module.workbook, all 26 authored Vitest rows | PASS 27/27 | 20260917T173823Z-p615761 |
| make frontend-typecheck | PASS 2/2 | 20260917T173713Z-p563515 |
| make frontend-import-boundary-check | PASS 2/2 | 20260917T173713Z-p563581 |
| make lint-biome | PASS 2/2 | 20260917T173713Z-p563711 |
| make lint-scripts | PASS 2/2 | 20260917T173713Z-p563773 |
| make generated-artifact-policy-check | PASS 3/3 | 20260917T173713Z-p563215 |

Every run root contains its run manifest, run/target summaries and unit results;
selected ROWS and exact commands are retained there. Counts above are harness
graph units, including dependencies, not counts of browser scenarios or claims.
The frontend-only module selections were derived from the current authored
catalog's Vitest rows and passed to public `make test-slice` with explicit ROWS.


Final acceptance review strengthened the existing scrolling scenario with 405
isolated fixture records. Real gestures still stop at the initial loaded 100,
without query requests. Explicit Load more to 200 and 300 preserves the completed
range; the next load evicts its original membership and clears it. Earlier rows
restores the accepted first window before the unmount/capture check. This added
production eviction evidence passed with
`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_scrolling`,
11/11 units at `20260917T173949Z-p643233`. No routing, product code or golden changed
for this final evidence refinement.

### Final gap and acceptance assessment

| Applicable boundary | Status | Evidence / rationale |
| --- | --- | --- |
| G1 gesture matrix and pointer/keyboard semantic parity | PASS | Nine routed production range scenarios; identity-only anchor policy and threshold/reversal cases. |
| G2 single admission owner, immediate click, nested controls and compatibility click | PASS | Controller/DOM/index suite; pointer, keyboard, native editor, reference popup and shared consumer browser rows. |
| G3 commit gate, rejection, late acceptance and supersession | PASS | Production pending/rejection/query scenarios, active-editor integration tests and existing recovery/continuation rows. |
| G4 bounded scrolling, capture/lifecycle cleanup and retained membership | PASS | Real virtualized two-axis loaded-window scenario, explicit append/eviction refinement, controller bounds and cleanup tests. |
| G5 dimensions, non-color cues, visible focus, inspector/bulk independence | PASS | Range accessibility/grouped scenarios, reviewed 100%/200% images, ordinary visual validation and existing accessibility rows. |
| G6 clipboard/fill, sizing and spreadsheet behavior | PASS | Pointer/keyboard copy/fill equivalence; original clipboard/bulk recovery, resize/fit, direct typing, caret, Enter/Tab/Escape and creation rows. |
| G7 range classes survive state registration | PASS | Actual dashed preview and hatched/bordered completed cells in production screenshots/assertions. |
| G8 selection anchors contain identity only | PASS | Shared extension policy plus production editor-origin Shift range regression. |
| G9 already-active readable cell regains focus | PASS | Adapter production-binding regression and Indicator late-recovery browser rerun. |
| G10 draft entry retires selection without a record | PASS | Native draft focus unit binding and production range/creation/spreadsheet rerun. |
| Accepted query, grouping, hidden/reordered fields, deletion and authority changes | PASS | Presentation/query/live-membership/authority scenarios, failed refresh retention and explicit eviction evidence. |
| Other Workbook and Network Analysis consumers | PASS | Full Adapter/Workbook suites, Inspector/Indicator/public-route/keyboard/browser slices and Network Analysis interaction/accessibility/protection rows. |
| Architecture, source ownership, import boundaries, authored/generated projection fidelity | PASS | Architecture/import checks, UI projection suite, JSON/catalog/finalizer generated-drift and generated-artifact policy. |
| Touch/pen profile, disjoint selection, unloaded selection, query expansion, new bulk commands | N/A | Explicit user scope exclusions; capability adds no such authority or runtime. |
| HTTP/API, database, saved-layout, dependency and persistent-storage migration | N/A | No changes to those contracts or files; selection remains runtime-only. |
| Core 05 conformance/release claim publication | N/A | This handoff records implementation evidence; no claim-publication boundary was requested. |

No unresolved applicable blocker remains. G1–G10 risks have the binary passing
criteria above; bounds of verification remain the supported production browser,
zoom/text-spacing profiles and routed owner slices exercised here. Full repository
`make check`, CI/release checks and retained warm-run maintenance were not run:
the changed seam is covered by narrow owner, service-backed, accessibility,
visual and measurement evidence, and no full warm-check RESULTS_DIR is available.
Unrelated backend, API, dependency and deployment checks were not expanded.
No visual goldens were regenerated. No historical limitation is used to excuse
an applicable failure. Earlier failed/cancelled runs remain recorded above.

### Changed files

All paths below are repository-relative. This inventory includes new untracked
sources and the controlling artifact; no pre-existing dirty files existed.

- `apps/web/e2e/indicator-lifecycle.spec.ts`
- `apps/web/e2e/inspector-actions.spec.ts`
- `apps/web/e2e/keyboard.spec.ts`
- `apps/web/e2e/measurement/timeline-grid.spec.ts`
- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/e2e/support/workbook/decisionSupersession.ts`
- `apps/web/e2e/support/workbook/indicatorLifecycle.ts`
- `apps/web/e2e/support/workbook/rowMutations.ts`
- `apps/web/e2e/timeline-public-route.spec.ts`
- `apps/web/e2e/timeline-range-selection.spec.ts`
- `apps/web/e2e/workbook-inspector-edit.spec.ts`
- `apps/web/e2e/workbook-query-browsing.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/src/workbook/continuity/README.md`
- `apps/web/src/workbook/timeline/README.md`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookGrid.tsx`
- `apps/web/src/workbook/timeline/composition/useTimelineMutationComposition.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `contracts/design/presentation.v1.json`
- `docs/design.md`
- `docs/handoffs/ui-ux/workbook-timeline-range-selection-refactor-handoff.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `packages/grid-adapter/README.md`
- `packages/grid-adapter/src/SemanticDataGrid.tsx`
- `packages/grid-adapter/src/columnSizing.ts`
- `packages/grid-adapter/src/core.ts`
- `packages/grid-adapter/src/domInteraction.ts`
- `packages/grid-adapter/src/gridInteraction.test.ts`
- `packages/grid-adapter/src/gridInteractionController.ts`
- `packages/grid-adapter/src/gridInteractionDom.ts`
- `packages/grid-adapter/src/index.test.tsx`
- `packages/grid-adapter/src/rdgCompiler.tsx`
- `packages/grid-adapter/src/semanticKeyboardPolicy.ts`
- `packages/grid-adapter/src/semanticPresentation.ts`
- `packages/grid-adapter/src/semanticSelectionPolicy.ts`
- `packages/grid-adapter/src/styles.css`
- `packages/grid-adapter/src/useGridInteraction.ts`
- `packages/ui-contracts/src/design-presentation.test.ts`
- `packages/ui-contracts/src/generated/design-presentation.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/harness/generated-artifacts/design-presentation/design-presentation.mjs`
- `tools/schemas/cartulary.design_presentation.v1.schema.json`
- `tools/test_families/module.timeline.json`
- `tools/test_families/package.grid_adapter.json`


### Extension, compatibility and rollback

Another explicitly authorized grid supplies semantic rows, visible columns and
stable surface/record/field identities, then enables
`cellRangeSelection: { kind: "contiguous", scopeKey }`. The opaque scope key must
represent the accepted sheet/query/authority lifetime. The existing optional
`cellRange`/`onCellRangeChange` remains the sole completed-selection boundary;
there is no new registry or continuity store. Omission preserves keyboard range
support and ordinary clicks without enabling pointer ranges. Grid Adapter owns
coordinates, mounted-cell registration, capture and scrolling. The consumer
continues to supply its existing editor/commit and clipboard/fill contracts; it
imports no Timeline query, inspector, mutation or draft workflow semantics.

Retired paths: mouse-down/window mouse-up timer selection, independent vendor
click activation, pending vendor range endpoint, unconditional vendor range
collapse, and DOM-attribute parsing of Timeline record anchors. Accessibility
click activation uses the same admission controller. Browser fixture activation
now uses real completed clicks and observes focus, without synthetic mouse-down
or forced cell focus. Native actions, editor popups, sizing/reordering and fill
retain their explicit owners.

Rollback this slice's Core/design text, authored projection/schema/generator,
generated derivatives, Adapter/Timeline sources, tests/catalogs and source guides
together. Preserve existing commits, accepted writes, record history and analyst
data. No storage or API migration must be reversed. Finalizer mutations remained
within the already-authored scope; no unrelated changes required isolation.
The delivery is a local reviewable working-tree implementation. Next action:
review the diff and retained handoff evidence. No commit, push or deployment was
performed or requested as part of delivery.


### TRS-05 exit checklist

- PASS: Core 03/design dispositions reconciled before implementation; final Escape
  wording agrees with the tested gesture-only cancellation rule.
- PASS: G1–G10 and every applicable acceptance boundary have passing evidence;
  exclusions have explicit scope/owner rationale; no applicable BLOCKED row.
- PASS: production range, operation, spreadsheet, lifecycle, accessibility,
  visual and affected-consumer evidence recorded with actual run roots.
- PASS: finalizer completed before broader terminal checks; authored inputs and
  generated derivatives agree; no unrelated finalizer edits remain.
- PASS: final Adapter 52/52, Workbook 271/271, Timeline frontend 25/25,
  Workbook frontend 27/27, architecture 12/12 and UI projection 10/10 units.
- PASS: after the final browser eviction assertion, `make frontend-typecheck`
  passed at `20260917T174210Z-p675165`. `make lint-biome` found only formatting in
  that new assertion (`20260917T174210Z-p675179`); `make format` corrected it at
  `20260917T174246Z-p677713`, and normal `make lint-biome` passed at
  `20260917T174310Z-p682139`. No functional source changed during this repair.
- PASS: `make lint-markdown` completed at `20260917T174211Z-p675514`, with summary
  `adhoc/lint-markdown/tool-run-summary.json`. Import/source boundaries, scripts,
  JSON/catalog, generated policy and generated drift results are recorded above.
- PASS: final `git diff --check` and scope review. Branch remains `main`, HEAD
  `369b35231a54c4970a3a4977073aace58f291b4d`, one existing commit ahead of
  origin/main. Inventory is 40 modified tracked paths and six new paths. Digest,
  lockfiles, analyst data and existing commits are preserved.
- PASS: compatibility, extension adoption, retirement, rollback, limitations and
  skipped full-warm-run maintenance are explicit. Next action is local review.

TRS-05 has no remaining required implementation or validation work.
