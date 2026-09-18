# Timeline range-aware keyboard entry

## Execution control

Execution begins on clean `main` at
`eb043ab61c75804f5e2b1a15181c4ff9bfad3071`. Root `AGENTS.md` is the only
applicable instruction file. `git status --short --branch`, `git rev-parse HEAD`,
and staged/unstaged diff summaries revalidated the planning baseline before edits.
There are no pre-existing edits. No digest edits, commits, pushes, deployments,
analyst-data changes, dependencies, persistence, backend endpoints or migrations
are included.

| Workstream | Status | Dependency / exit |
| --- | --- | --- |
| RTE-01 Characterization and adopted interaction contract | DONE | Owner-backed transition matrix and every confirmed gap disposition; Markdown/whitespace pass. |
| RTE-02 One semantic traversal and selection owner | DONE | Opt-in semantic planner, shared ordered-membership primitive and focused/default-consumer checks pass. |
| RTE-03 Editing, settlement and lifecycle integration | DONE | RTE-02 DONE; production composition and recovery pass focused checks. |
| RTE-04 Production-browser and spreadsheet evidence | DONE | RTE-03 DONE; applicable browser, security, accessibility and measurement evidence passes. |
| RTE-05 Terminal validation and completed handoff | DONE | RTE-04 DONE; final source and completed handoff covered by required checks. |

Only the current workstream may be IN_PROGRESS. Save its actual exit evidence
and DONE before starting its successor. Applicable BLOCKED dependencies prevent
dependent work and completion. Historical passes are not current acceptance.

## Authority and inspected boundaries

Core 01 §§3.3.4–3.3.5, 7.4, 18/18A/18B owns stable identities, capabilities,
query contracts and authoritative mutation responses. Core 03 REQ-03-217/218/220/
281/298/299/300, §§2.3A, 3–4, 13.4–13.5 and 14.9 owns interaction, retained
authoring, conflict, query continuity and authority lifetimes. Core 04 §§1–2
owns authorization and protected-content handling. Design §§8.4–8.6, 12.5 and
14 supplies interaction presentation and accessibility; domain owns vocabulary
and navigation. The approved narrow Core 03/design amendments precede code.

The digest's localized read order and the completed range-selection, Find,
grid-edit/autosave, clipboard-fidelity, column-sizing, query-continuation and
correction-access handoffs were inspected during planning. They are advisory
and historical evidence. `docs/research/nlspec-spec.md` is research, not another
product request or authority.

Source placement is governed separately by `tools/frontend_source_ownership.json`
and `tools/frontend_import_boundaries.json`; verification routing by
`contracts/verification`, `tools/test_catalog_owner.json` and authored
`tools/test_families`. No executable input may depend on Markdown. Inspected
source includes Grid Adapter semantic keyboard/selection/presentation/focus,
interaction controller and DOM binding, editor sessions and RDG compiler;
Timeline grid/column/scalar-renderer bindings and keyboard/anchor controllers;
Workbook Find, native/reference editors, shell Escape and retained authoring.

## Adopted transition matrix

Selection geometry, active cell, editor, Find match, Inspector context and bulk
checkbox selection are independent. A valid completed multi-cell rectangle uses
the exact ordered eligible membership of the accepted presentation.

| Context / action | Result |
| --- | --- |
| No range or one cell; Enter/Tab and Shift variants | Existing Timeline vertical/horizontal navigation, authorized trailing draft and outer Tab shell exit. Navigation alone creates nothing. |
| Complete pointer/keyboard selection | Preserve original anchor/endpoints. Active cell is initially the endpoint, including reversed selections. |
| Multi-cell Tab / Shift+Tab | Forward/reverse row-major traversal with cyclic wrapping; endpoints remain unchanged. |
| Multi-cell Enter / Shift+Enter in navigation mode | Forward/reverse column-major traversal with cyclic wrapping; endpoints remain unchanged. |
| Printable input / unmodified F2 | Existing assigned application shortcuts, including Core 03 §13.4 Space evidence preview, keep priority. Eligible printable entry edits only the active authorized direct-value cell, retaining the valid rectangle. Printable input seeds its character; F2 preserves current/retained value and puts a supported caret at the end. |
| Stationary pointer click | Existing replacement selection, one single-click edit and explicit pointer inspection. A click inside the existing editor retains native caret/text selection. |
| Editor departure pending | Capture the semantic destination before settlement; retain original editor/draft/range. No destination is applied before acceptance. |
| Accepted departure | Apply only the latest still-valid destination once, retaining valid range geometry. |
| Rejected departure / conflict | Preserve exact raw draft, original eligible focusable editor and still-valid range. No destination or extra submission. |
| Editor cancellation / explicit Commit | Cancel returns to the same active cell and retains the valid range. Explicit Commit saves without inventing movement. |
| Grid-owned Escape with multi-cell range | Collapse to the active cell and consume; subsequent Escape follows Inspector/shell hierarchy. |
| Popup, tentative gesture, editor Escape | Higher-priority owner resolves exactly its own action before range collapse. Tentative cancellation does not discard editor work. |
| Multiline Shift+Enter | Native newline; no commit or traversal. Enter commits; Tab/Shift+Tab depart through acceptance. |
| Native select | Enter, option navigation and picker dismissal stay native. Tab/Shift+Tab or existing Commit controls settle the editor. |
| Reference popup, collection/nested controls, composition | Local keys retain ownership. No parent traversal or scalar activation; no added mutation capability. |
| Readable read-only cell | Traversable and copyable; no editor or mutation rights. |
| Shift+Arrow after internal traversal | Keep original anchor; step from current active cell and replace endpoint. Rectangle may shrink. User confirmed this policy. |
| Delete / Backspace | Existing active-cell clear/edit and range-retirement behavior; never a multi-cell clear. |
| Find opening/recompute; Inspector/layout focus borrowing | Preserve valid selection and source-owned draft. No implicit commit or Inspector retarget. |
| Explicit Find/replacement navigation | Acceptance-gated replacement with destination cell; preserve separate Inspector/bulk context. |
| Value/version-only refresh / outside-range append | Retain range only when exact ordered members remain unchanged. Never substitute membership. |
| Pending/failed query replacement | Keep selection under the still-accepted presentation. |
| Accepted query/surface/grouping change, reordered/hidden/collapsed/deleted/evicted members | Existing semantic invalidation clears range; obsolete movement cancels. Retained work remains source-owned. |
| Virtualization / offscreen destination | Use full semantic membership and existing cancellable reveal/focus; never fetch rows, expand groups or enter the draft. |
| Read-only transition / incident or session authority change | Cancel obsolete movement/selection under current owner rules. Preserve readable local work or conceal/retire protected work as required; late callbacks cannot resurrect presentation. |

For `A1 B1 / A2 B2`, Tab cycles `A1 → B1 → A2 → B2 → A1`; Enter cycles
`A1 → A2 → B1 → B2 → A1`; Shift reverses each cycle. Completion at B2
stays at B2 until the next action. After Tab moves B2 to A1, Shift+ArrowRight
keeps anchor A1 and sets endpoint B1. Other Grid Adapter consumers retain their
adopted outer navigation policies. Native select Enter ownership was explicitly
confirmed during planning.

## Characterization and gap ledger

| Gap / classification | Remediation and affected areas | Rationale / long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Authorized enhancement: keyboard navigation/edit retires ranges | Adopt Core 03/design; opt-in Adapter policy and Timeline binding; unit/browser tests. | Repeated entry with explicit capability and one range owner. | Other consumers default unchanged; no wire/data migration. Unresolved: selected rectangles cannot support repeated entry. | Correct four cycles/wraps, retained geometry/editing and accessible exit. |
| G2 Structural weakness: restoration equates active cell with endpoint | Separate selection disposition from focus in Adapter navigation and restoration. | Semantic geometry and focus remain independent, including virtualized cells. | Find continues replacement. Unresolved: traversal/restoration silently collapses selection. | Active member can differ from endpoint through traversal, rejection and focus borrowing. |
| G3 Structural weakness: accepted editor callback calculates movement after settlement | Capture decision before existing session gate; fence current membership/scope/authority/intent and reuse navigation owner. | Delayed results cannot select new members or override newer work. | Existing write/receipt/draft owners retained. Unresolved: stale focus or duplicate departure paths. | One request, latest valid captured destination, no obsolete focus after rejection/replacement. |
| G4 Owner reconciliation / implementation interception: multiline Shift+Enter | Narrow Core 03/design exception and Adapter/Timeline local-key guards; native select and IME evidence. | Predictable local text/control ownership. | Native select Enter stays native; other outer navigation unchanged. Unresolved: newline becomes save/traversal. | Newline/composition/local popup input never submits or traverses; Tab remains gated. |
| G5 Coupling: managed editor focus invokes Inspector selection | Retire passive managed-grid focus-to-inspection callback; preserve explicit pointer inspection. | Editing/focus do not become another Inspector owner. | Ordinary stationary pointer inspection stays supported. Unresolved: each edited range member retargets Inspector. | Range traversal/edit leaves Inspector context and bulk checkboxes unchanged. |
| G8 Confirmed event propagation gap | Bridge native cancellation for grid-owned Escape at the Adapter frame; test React parent propagation and production Inspector ladder. | RDG copies React events; copied propagation flags do not stop shell handlers. One Escape must resolve one owner. | Existing single-cell shell Escape preserved. Unresolved: range collapse also closes Inspector. | First Escape collapses only range; second follows Inspector hierarchy. |
| G7 Confirmed browser description collision | Adapter composes caller context via optional semantic `setAccessibleDescription`; Workbook retires direct DOM description writes. | One presentation owner prevents guidance overwrite and preserves loaded-window meaning. | Additive handle capability; no vendor details in caller. Unresolved: keyboard exit is undiscoverable. | Browser description includes loaded-window and cyclic/Escape guidance; default consumers keep context. |
| G6 Regression/evidence boundary | Preserve clipboard, fill, first-input creation, sizing, correction reveal, retained work and authority recovery; source guides and authored routing. | Existing supported obligations remain reviewable through production evidence. | No new bulk commands/framework/persistence. Unresolved: shared adapter regressions or inaccessible rejected drafts. | Applicable spreadsheet, shared-consumer, geometry, lifecycle and security rows pass. |

## Evidence log

Planning baseline only: public help/help-all and task guides for
`package.grid_adapter`, `web.workbook`, `module.timeline`, `module.workbook`
passed. Five selected Adapter rows (closed keyboard decisions, range interaction,
semantic navigation/focus and range retention) passed 6/6 graph units at
`.cartulary/test-results/20260918T001408Z-p2570816`. This does not complete an
implementation workstream. Implementation baseline inspection is recorded above.

## Acceptance and rollback

Applicable acceptance rows require PASS; N/A needs an owner/scope rationale;
an applicable BLOCKED row prevents completion. Current workstream evidence and
terminal acceptance are recorded below; baseline evidence is never substituted
for implementation acceptance.

Rollback restores owner amendments, authored inputs, implementation, tests,
generated routing and guides together; never undo accepted analyst writes.
RESULTS_DIR remains unset unless a qualifying successful full warm run matches
the final source. Retained-run maintenance must be reported as skipped when unset.

### RTE-01 exit

Core 03 REQ-03-217/218/220/300 and §13.4 and design §§8.4–8.6, 12.5/14.2
now adopt the authorized matrix. The edit-entry retirement and completed-range
Escape statements have narrow explicit exceptions. Multiline/select/composition
ownership is consistent across the two owners. No unrelated material owner
contradiction was identified. Existing Find, retained-authoring, mutation,
security, clipboard and browsing contracts remain controlling.

`make lint-markdown`: PASS at
`.cartulary/test-results/20260918T004212Z-p2577669`, summary
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check`: PASS.
Changes at this exit are the two owner documents and this handoff only.
Next action: implement RTE-02's neutral semantic traversal and focused evidence.

### RTE-02 exit

The existing `decideSpreadsheetNavigation` now owns both ordinary and opt-in
range order. Its explicit range result distinguishes retention from replacement.
`resolveSemanticCellRange` supplies ordered semantic membership to traversal,
visible-range expansion and existing retention; there is no cell cache, new
range store or coordinate identity. `cellRangeSelection.keyboardEntry='cycle'`
and editor activation `f2` are additive internal TypeScript contracts. Default
consumers remain unchanged. RTE-03 binds these decisions to production entry and
the existing settlement/focus owner.

- `make generate`: PASS, `20260918T004641Z-p2584544`; authored Adapter test row
  generated the topology render-index derivative.
- `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.range_keyboard_entry,package.grid_adapter.regression.core_verifies_c_closed_semantic_keyboard_decisio_23d2a0e429,package.grid_adapter.regression.window_range_retention,package.grid_adapter.regression.range_interaction`:
  PASS 5/5 graph units, final `20260918T004836Z-p2589839`.
- `make frontend-typecheck`: PASS 2/2, `20260918T004826Z-p2589403`.
  Initial `20260918T004705Z-p2587797` failed on two new test fixture types and
  the support binding's exhaustive decision switch. These change-related errors
  were corrected before the final pass; no product assertion was weakened.
- `make format`: PASS, `20260918T004631Z-p2580277`.

Run roots above are under `.cartulary/test-results/`. New tests cover every
member/wrap in both orders/directions, reversed geometry, one row/column,
read-only fields, presented membership, invalidation, ordinary draft/shell/default
policies, F2/typing/clear/exit and extension from the traversed active cell.
Next action: integrate RTE-03 editor departure and lifecycle without a second gate.


### RTE-03 exit

Production `SemanticDataGrid` and `rdgCompiler` deliver editor intent before
settlement to `semanticCellNavigation`, which captures and revalidates semantic
destinations, exact membership, accepted scope, authority and cancellation.
The existing editor session alone submits writes. Repeated keys supersede focus
intent without accumulating movement or duplicating submission. Imperative
activation, detachment, focus borrowing and new input cancel obsolete intentions.
Compatible append/value updates retain geometry; invalidation never reverses an
accepted write. The post-settlement movement calculation was retired.

Timeline enables the optional capability. F2, printable entry, cancellation,
select/multiline/IME ownership and polite grid-owned collapse are integrated.
Editor cells retain accessible selection. Managed Timeline focus no longer
requests Inspector selection; explicit pointer inspection remains. Supported test
bindings use the same semantic planner. Source guides describe the seam.

- `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.index_suite_d804ef9789,package.grid_adapter.regression.semantic_cell_navigation,package.grid_adapter.regression.range_keyboard_entry,package.grid_adapter.regression.nested_editor_interaction`:
  PASS 5/5, final `20260918T010223Z-p2613863`; production binding coverage includes
  F2/caret, rejection/raw draft, latest repeated departure, one submission,
  retained geometry/ARIA, Escape and native multiline/select/composition keys.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_keyboard_event_and_focus_ownership_97dc4e9a17,web.workbook.regression.grid_draft_lifetime`:
  PASS 3/3, `20260918T010108Z-p2607914`.
- `make frontend-typecheck`: PASS 2/2, `20260918T010133Z-p2608907`.
  Initial `20260918T010043Z-p2606654` failed on missing scope keys and an overly
  narrow range fixture type in new tests; fixed without changing assertions.
- `make generate`: PASS, `20260918T010029Z-p2603402`; authored new test titles
  routed through the generated render index. `make format`: PASS, final
  `20260918T010210Z-p2609570`. `git diff --check`: PASS.

All run roots use `.cartulary/test-results/`. No application fake establishes
browser focus or virtualization; those remain RTE-04 obligations. Next action:
production-browser scenarios and current regression/security/geometry evidence.


RTE-04 initial browser run `20260918T010508Z-p2623418` failed 9/11 graph units:
new cycle assertions passed until the accessible-description check revealed G7;
new settlement fixture assumed insertion order instead of accepted query order.
Native multiline/composition, Find restoration and single-cell clear passed.
G7 is remediated in the semantic handle and existing Workbook focus hook; fixture
values now come from accepted rows. No requirements or behavioral assertions were
weakened. Next action: rerun the three new production rows.


RTE-04 regression dispositions: `20260918T010834Z-p2720196` passed all nine
range rows and seven spreadsheet rows; the ordinary-key row still expected
multiline draft Shift+Enter traversal. It now asserts native newline and tests
reverse draft navigation through the existing single-line source editor.
`20260918T010806Z-p2683563` passed selected Find, sizing and six autosave rows;
the scalar-family row expected native select Enter to submit. It now verifies
native Enter/dismissal before gated Tab. Rerun of that row with shared-consumer
and reference rows passed 13/13 graph units at `20260918T011156Z-p2793832`.

New settlement test attempts were corrected to respect existing semantics:
server `invalid_request` halts that incident's queue, so rejection is tested after
successful departure; resubmitting an unchanged accepted value is a no-op, so the
rejected value must differ. Core 03 §13.4's assigned Space evidence shortcut retains
priority over eligible printable entry. Exact leading whitespace is entered via
F2 and native input, without weakening the raw-draft assertion. These are fixture
corrections, not new product exceptions. Failed roots:
`20260918T010753Z-p2660105`, `20260918T011135Z-p2764881`,
`20260918T011353Z-p2872953`. Biome initially flagged four non-null assertions in
new semantic fixtures at `20260918T011458Z-p2936957`; replaced with checked fixture
access. Required current passes remain pending.


RTE-04 full accessibility run `20260918T011245Z-p2830265` failed 17/20 graph
units. Select correction access exposed an overbroad native-arrow guard: the
existing Alt+ArrowDown correction shortcut must retain priority when the native
picker is closed. Core 03/design explicitly reconcile this before the guard is
reordered. Ordinary option arrows/Enter and open picker dismissal stay native.
Risk without remediation: inaccessible correction actions on narrow layouts.
Binary exit: correction-family and full accessibility checks pass. An unrelated
Coordination recovery focus failure is under investigation; no acceptance is
claimed for that row until current evidence passes or a blocker is recorded.


Current RTE-04 passes: the four range settlement/lifecycle/Inspector/accessibility
rows passed 11/11 graph units at `20260918T011714Z-p2948628`. They cover exact raw
rejection and same-field conflict, accepted writes with deleted destinations,
authority loss, latest repeated keys, newer external focus, unchanged Inspector
context, reordered columns and Escape/reentry under zoom/text spacing.
Query continuation (Timeline, live recovery, authority and session) plus Find
protected-context and source-editor borrowing passed 13/13 at
`20260918T011432Z-p2906836`. The selected shared-consumer/reference rows passed
13/13 at `20260918T011156Z-p2793832`. Typecheck passed 2/2 at
`20260918T011842Z-p3013307`. Remaining visual/measurement and accessibility reruns
must pass before RTE-04 can finish.


### RTE-04 command ledger

All roots below are under `.cartulary/test-results/`. Commands ran from the
repository root. Their `run-manifest.json`, `run-summary.json`, row receipts and
browser group reports preserve the exact routing and execution evidence.

| Run root | Exact command | Result |
| --- | --- | --- |
| `20260918T010508Z-p2623418` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_settlement,module.timeline.browser.range_entry_native` | FAIL 9/11 graph units |
| `20260918T010753Z-p2660105` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_settlement,module.timeline.browser.range_entry_native` | FAIL 9/11 graph units |
| `20260918T010806Z-p2683563` | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.find_edit_departure,module.workbook.browser.find_scope_lifecycle,module.workbook.browser.find_virtualized,module.workbook.browser.grid_autosave_editor_families,module.workbook.browser.grid_autosave_timestamp_retention,module.workbook.browser.grid_autosave_successive,module.workbook.browser.grid_autosave_refresh,module.workbook.browser.grid_autosave_role,module.workbook.browser.grid_autosave_session,module.workbook.browser.grid_autosave_availability,module.workbook.browser.column_sizing_gestures_saved` | FAIL 13/15 graph units |
| `20260918T010834Z-p2720196` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_accessibility,module.timeline.browser.range_authority,module.timeline.browser.range_cancellation,module.timeline.browser.range_departure,module.timeline.browser.range_gestures,module.timeline.browser.range_live_membership,module.timeline.browser.range_presentation,module.timeline.browser.range_rejection_query,module.timeline.browser.range_scrolling,module.timeline.browser.spreadsheet_keyboard,module.timeline.browser.spreadsheet_creation,module.timeline.browser.spreadsheet_pointer,module.timeline.browser.spreadsheet_rejection,module.timeline.browser.spreadsheet_refresh,module.timeline.browser.spreadsheet_virtualization,module.timeline.browser.spreadsheet_bulk,module.timeline.browser.clipboard_fidelity` | FAIL 11/13 graph units |
| `20260918T011135Z-p2764881` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_settlement,module.timeline.browser.spreadsheet_keyboard` | FAIL 11/13 graph units |
| `20260918T011156Z-p2793832` | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.grid_autosave_editor_families,module.workbook.browser.the_explicitly_visited_shared_grid_keyboard_anch_2b5da548a2,module.workbook.browser.reference_grid_commit,module.workbook.browser.reference_admission_lifecycle,module.workbook.browser.reference_session_recovery` | PASS 13/13 graph units |
| `20260918T011245Z-p2830265` | `make browser-e2e-a11y` | FAIL 17/20 graph units |
| `20260918T011353Z-p2872953` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_settlement,module.timeline.browser.range_entry_lifecycle,module.timeline.browser.range_entry_native` | FAIL 9/11 graph units |
| `20260918T011432Z-p2906836` | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.query_continuation_timeline,module.workbook.browser_stateful.query_continuation_live_recovery,module.workbook.browser_stateful.query_continuation_authority,module.workbook.browser_stateful.query_continuation_session,module.workbook.browser.find_authority,module.workbook.browser.find_source_editors` | PASS 13/13 graph units |
| `20260918T011714Z-p2948628` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_settlement,module.timeline.browser.range_entry_lifecycle,module.timeline.browser.range_presentation,module.timeline.browser.range_accessibility` | PASS 11/11 graph units |
| `20260918T011820Z-p2980300` | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.grid_autosave,module.workbook.accessibility.coordination_create_authoring_recovery` | PASS 13/13 graph units |
| `20260918T011902Z-p3016056` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810,module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206,module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf` | FAIL 6/11 graph units |
| `20260918T011926Z-p3040519` | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0,module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | PASS 11/11 graph units |
| `20260918T011927Z-p3040738` | `make service-backed-test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.visual.capture_test_only_grid_adapter_support_specimens_9c222633ba` | PASS 11/11 graph units |
| `20260918T012021Z-p3104541` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810,module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206,module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf` | PASS 11/11 graph units |
| `20260918T012022Z-p3104760` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_native,module.timeline.browser.range_scrolling` | PASS 11/11 graph units |


The correction/Coordination accessibility rerun passed 13/13 at
`20260918T011820Z-p2980300`; Coordination required no source change.
Timeline visuals passed 11/11 at `20260918T012021Z-p3104541`; Workbook visuals
passed 11/11 at `20260918T011926Z-p3040519`; Adapter support visuals passed 11/11
at `20260918T011927Z-p3040738`. Initial Timeline visual preparation
`20260918T011902Z-p3016056` rejected source changes during formatting; no visual
assertion ran. Rerun used settled source. No golden update or promotion occurred.

Reviewed the range zoom/text-spacing PNG attachments in the passing range
accessibility report, the narrow bottom-edge editor/correction-action PNGs in
the passing correction report, and the matching active-edit Timeline golden.
The range has hatching and an independent active outline; the narrow correction
toolbar remains anchored above the editor with visible focused actions. PNG
attachments are embedded in Playwright reports; decoded review copies are under
`.cartulary/rte-image-review/`. All selected visual rows and renderer-profile
attestations passed. Measurements and visual evidence remain implementation
support, not an expanded conformance claim.

The four isolated Timeline measurement rows passed 20/20 graph units at
`20260918T012145Z-p3167243`. No concurrent verification ran during their measured
execution. Current offscreen vertical/horizontal cycles and native/Find/clear
checks passed 11/11 at `20260918T012022Z-p3104760`; Adapter focused tests passed
5/5 at `20260918T012001Z-p3098929`; Biome passed 2/2 at
`20260918T011928Z-p3041384`.

`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6,module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13,module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3,module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95`: PASS, `20260918T012145Z-p3167243`.

`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810,module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206,module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf`: PASS, `20260918T012021Z-p3104541`.

`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_native,module.timeline.browser.range_scrolling`: PASS, `20260918T012022Z-p3104760`.

`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.grid_autosave,module.workbook.accessibility.coordination_create_authoring_recovery`: PASS, `20260918T011820Z-p2980300`.


RTE-04 two-stage Escape test `20260918T012627Z-p3212254` exposed G8: RDG's
copied cell event stopped native propagation but did not mark the original React
event stopped. The range collapsed and the Inspector also closed. Adapter now
bridges consumed native Escape at its frame before the shell bubble handler;
no second selection owner or Workbook DOM inspection is introduced. Current
browser and parent-propagation unit passes are required after this correction.


G8 correction passed Adapter parent-propagation and retained-entry tests 5/5 at
`20260918T012841Z-p3248314`, and typecheck 2/2 at
`20260918T012843Z-p3248980`. Browser rerun `20260918T012842Z-p3248557` passed
six range/native/lifecycle/virtualization cases and ordinary spreadsheet keys.
The Inspector case passed both Escape stages, then correctly rejected a pointer
fixture whose target became offscreen after Inspector reopening; the fixture now
reveals the target before the next gesture. Format/lint initially flagged the
new static event bridge and a mistyped test suppression (`20260918T012829Z-p3244032`,
`20260918T012911Z-p3278718`); justified boundary-only suppressions were corrected.
No unrelated source or formatting changes were retained. `make format` then
passed at `20260918T012935Z-p3281482`.


### RTE-04 exit

G8's complete Inspector ladder passed 11/11 at `20260918T013036Z-p3289154`:
`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_presentation`.
The final source passed `make lint-biome` 2/2 at
`20260918T013037Z-p3289447`. The four measurement rows were rerun after the
Escape bridge, isolated from other verification, and passed 20/20 at
`20260918T013157Z-p3321276` using the exact measurement command above.

Current aggregate evidence covers the adopted transition matrix, native editor
ownership, selected entry/settlement, semantic invalidation, authority lifetimes,
ordinary spreadsheet behavior, shared consumers, Find, correction geometry,
accessibility and the applicable visual/measurement rows. The final Escape
bridge changes event propagation only; it does not alter the already-passing
visual captures. No applicable BLOCKED dependency remains. G1–G8 are remediated
and have current passing evidence; the unrelated transient Coordination failure
passed its focused rerun without a source change. RTE-04 is DONE.

Next action: RTE-05 finalization, broad affected frontend/accessibility checks,
generation/ownership verification, final scope review and completed handoff.


### RTE-05 execution

RTE-04 exit was saved DONE before this workstream began. `RESULTS_DIR` is unset:
there is no successful full warm check run matching final source. Retained-run
maintenance is skipped for that reason. Selected fresh execution evidence is
not promoted to retained full-run evidence. Next action: `make agent-finalize`
before broader terminal verification.


### Final scope, compatibility and retired paths

Changed source is confined to the existing Adapter keyboard/presentation,
selection/focus, native-event and editor bindings; Timeline grid/scalar bindings;
and the existing Workbook semantic-focus/native-editor integration. Specifically,
`packages/grid-adapter/src/{SemanticDataGrid,rdgCompiler,test-support}.tsx`,
`core.ts`, `semanticKeyboardPolicy.ts`, `semanticCellNavigation.ts`,
`semanticPresentation.ts` and `domInteraction.ts` own the neutral mechanics.
`apps/web/src/workbook/timeline/components/{TimelineWorkbookGrid,TimelineScalarEditor}.tsx`,
`components/WorkbookGridEditorControl.tsx` and
`hooks/useWorkbookSemanticGridFocus.ts` own the consumer integration.

Owner amendments are limited to Core 03 and design. Source guides are
`packages/grid-adapter/README.md` and the Timeline components README. Tests are
`rangeKeyboard.test.ts`, Adapter `index.test.tsx` and
`semanticCellNavigation.test.ts`, Workbook `WorkbookShell.sentinel.test.tsx`,
and the three affected existing browser files
`timeline-range-selection.spec.ts`, `timeline-grid-entry.spec.ts` and
`workbook-grid-autosave.spec.ts`. Authored test routing changes are limited to
`tools/test_families/package.grid_adapter.json` and `module.timeline.json`;
`make generate` produced the browser batch manifest and topology render index.
Existing frontend source ownership covers all changed production files; no new
production file requires an ownership-manifest entry. Runtime/test/generator
inputs do not read or depend on Markdown.

The capability and activation union are additive internal TypeScript contracts;
Timeline alone opts into cyclic navigation. Other consumers keep their outer
navigation policies. Native multiline/select/composition exceptions apply at
the shared editor boundary, including the existing correction shortcut. No
transport contract, `GridCellRange` shape, dependency, backend route, database,
record cache, draft owner or persistent traversal session was added.

Retired redundant paths: the editor's post-settlement movement calculation;
the support binding's duplicate spreadsheet movement decisions; passive
Inspector selection on managed editor focus; and Workbook's direct DOM
accessible-description write. The existing semantic navigation owner now owns
destination lifetime, the existing editor session still owns submission, and
the Adapter composes accessible context and cyclic guidance.

Measurement observations in the final `20260918T013157Z-p3321276` reports used
100 measured samples per row. P95 was 35.7 ms for summary focus/edit, 31.6 ms for
typing acknowledgment, 33.1 ms for ArrowDown selection (each below 100 ms), and
104.1 ms for blank-row creation (below 150 ms). This is implementation evidence
for the existing fixture, not a broader performance claim.

Scope limits: existing Chromium/renderer-profile browser evidence establishes
DOM focus, native-control event ownership and virtualization. Synthetic
composition events establish that grid commands do not intercept composition;
this does not claim testing every operating-system IME candidate window.
Scalar families are exercised only on their adopted surfaces. No new Timeline
capabilities, bulk command, undo framework, formula support, unloaded selection
or disjoint selection is introduced. Backend unit-suite expansion and release
conformance publication are N/A: this seam changes frontend interaction only,
with real service-backed browser tests for existing writes and authorization.

Rollback must revert the owner amendments, implementation, guides, authored test
routing and generated derivatives as one coherent change, then regenerate and
rerun affected checks. Never roll back accepted analyst writes. The branch and
HEAD remain the planning baseline; all changes remain local and uncommitted.


### Final correction command ledger

These RTE-04 commands close G8 and its fixture/lint dispositions above.
All roots remain under `.cartulary/test-results/`.

| Run root | Exact command | Result |
| --- | --- | --- |
| `20260918T012627Z-p3212254` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_presentation` | FAIL 9/11 graph units |
| `20260918T012829Z-p3244032` | `make format` | FAIL 1/2 graph units |
| `20260918T012841Z-p3248314` | `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.index_suite_d804ef9789,package.grid_adapter.regression.range_keyboard_entry,package.grid_adapter.regression.semantic_cell_navigation,package.grid_adapter.regression.nested_editor_interaction` | PASS 5/5 graph units |
| `20260918T012842Z-p3248557` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_native,module.timeline.browser.range_entry_settlement,module.timeline.browser.range_entry_lifecycle,module.timeline.browser.range_presentation,module.timeline.browser.range_accessibility,module.timeline.browser.range_scrolling,module.timeline.browser.spreadsheet_keyboard` | FAIL 11/13 graph units |
| `20260918T012843Z-p3248980` | `make frontend-typecheck` | PASS 2/2 graph units |
| `20260918T012911Z-p3278718` | `make lint-biome` | FAIL 1/2 graph units |
| `20260918T012935Z-p3281482` | `make format` | PASS 2/2 graph units |
| `20260918T013036Z-p3289154` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_presentation` | PASS 11/11 graph units |
| `20260918T013037Z-p3289447` | `make lint-biome` | PASS 2/2 graph units |


RTE-05 full frontend-unit run `20260918T013731Z-p3367276` failed 646/647 graph
units. The sole failing row was
`web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_428c33b58f`:
the application-fake sentinel still expected multiline Shift+Enter to traverse.
Remediation updates `WorkbookShell.sentinel.test.tsx` to assert native ownership
and unchanged semantic anchor, then cancel the editor and exercise reverse
traversal from grid navigation. This is the same adopted exception already
validated through the production browser, not a runtime change or relaxed
assertion. Long-term benefit: support evidence agrees with the owner and native
text behavior. Compatibility: no added product change. Unresolved risk: a stale
test would demand a newline regression. Binary exit: this focused row and the
full frontend suite pass. Routing titles and ownership remain unchanged.

Terminal browser slices passed: Timeline 13/13 graph units at
`20260918T013900Z-p3426034`, Workbook/shared/reference/Find 15/15 at
`20260918T013900Z-p3426039`. Full `make browser-e2e-a11y` passed 20/20 at
`20260918T013731Z-p3367503`, resolving every earlier accessibility failure.
No product source changed after these passes. Next action: focused sentinel
rerun, finalizer and remaining terminal checks covering the test correction.

The first focused sentinel correction at `20260918T014345Z-p3548937` still
failed: the test sent reverse navigation before its existing asynchronous
Escape focus restoration completed. It now awaits the restored gridcell before
sending the next key. This changes only test synchronization; the production
browser already verifies the full editor/range/Inspector Escape sequence.
Exact focused command: `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_428c33b58f`.


### Terminal acceptance map

This map combines current semantic/unit, production binding, and real
service-backed browser evidence. Application fakes supply support evidence only;
the browser rows establish physical focus, virtualization and visible state.
Run roots and exact commands are in the ledgers above and below.

| Acceptance obligation | Status | Current evidence / disposition |
| --- | --- | --- |
| Adopted ownership and complete transition matrix | PASS | Core 03/design reconciled before implementation; RTE-01 exit, G1–G8 dispositions; no unrelated owner contradiction. |
| Four traversal directions and wraps, all range shapes/directions, stable geometry and independent active cell | PASS | `rangeKeyboard.test.ts`, production Adapter binding and Timeline `range_entry_cycles`; final Timeline slice `20260918T013900Z-p3426034`. |
| One captured destination and acceptance gate; rejection/conflict/raw draft, revision changes, repeated keys and blur cardinality | PASS | `semanticCellNavigation.test.ts`, production Adapter binding, Timeline settlement/lifecycle rows and Workbook autosave/reference rows. Final Timeline/Workbook slices passed. |
| Typing/F2, caret/selection, multiline/select/reference/IME local keys, readable read-only cells and single-cell clear | PASS | Timeline native/authority/cycles rows, Adapter production binding, Workbook scalar-family/reference rows; no added Timeline field capability. |
| Query continuity, exact membership, virtualization, hidden/reordered/grouped/evicted/deleted members and Find consequences | PASS | Timeline presentation/scroller/live/rejection rows, Workbook Find rows and query-continuation rows at `20260918T011432Z-p2906836`; semantic delayed-destination validation. |
| Read-only/revocation/session/account transitions and late-callback containment | PASS | Timeline held-write lifecycle/authority rows; Workbook autosave role/session/availability, query continuation authority/session, Find protected-context and reference session recovery; semantic scope/authority cancellation tests. |
| Ordinary navigation and trailing draft, first qualifying input, clipboard/fill, Inspector/bulk context, sizing and correction access | PASS | Final Timeline spreadsheet/clipboard slice; shared-consumer slice; sizing row and full accessibility correction rows. No new bulk operation. |
| Accessible cyclic exit, separate non-color cues, concise announcements, supported density/layout/text-spacing and narrow correction geometry | PASS | Final `browser-e2e-a11y` 20/20 at `20260918T013731Z-p3367503`, range zoom/text-spacing/Inspector rows, reviewed capture attachments. |
| Applicable production visual and measurement obligations | PASS | All selected Timeline/Workbook/Adapter visual rows passed unchanged; final isolated four measurement rows 20/20 at `20260918T013157Z-p3321276`. |
| Full frontend unit, type, style and ownership/import checks | PASS | Final full frontend suite 647/647 at `20260918T014535Z-p3554757`; final typecheck, Biome and import checks passed. |
| Authored routing, generated derivatives, JSON shape and no executable Markdown dependency | PASS | Final drift/policy/shape/import checks passed after the test correction; catalog and package reachability passed. No generated ownership change remained. |
| Completed handoff, Markdown, whitespace and final scope review | PASS | Completed handoff lint passed at `20260918T015008Z-p3626628`; whitespace and final 28-file scope review passed. |
| Backend endpoints, migrations, persistence, dependencies and new bulk commands | N/A | No changed input in these owner scopes; existing service-backed routes are exercised without expanding backend behavior. |
| Full release/conformance publication | N/A | This frontend seam makes no new Core 05 claim or release artifact. |
| Retained full-warm-run maintenance | N/A | Skipped: RESULTS_DIR unset; no qualifying successful full warm run matching source exists. Fresh selected evidence is retained at its own run roots. |

The latest `env -u RESULTS_DIR make agent-finalize` passed 1/1 at
`20260918T014507Z-p3550881` before the final broader checks. The preceding
finalizer also passed at `20260918T013700Z-p3362375`. No retained run was
maintained or represented as covering the changed source.


### RTE-05 terminal command ledger

All commands ran from the repository root. The multi-target shell invocations
were `make frontend-typecheck frontend-unit frontend-import-boundary-check lint-biome`
(the initial unit failure stopped later targets),
`make generate-drift generated-artifact-policy-check json-shape-check test-catalog-check`,
`make frontend-import-boundary-check lint-biome`, and, after the sentinel fix,
`make frontend-typecheck lint-biome generate-drift generated-artifact-policy-check json-shape-check frontend-import-boundary-check`.
The final full unit rerun uses `make frontend-unit`. `make frontend-fallow-static`
was selected from the Adapter task guide; it checks the retained manual package
surface. The tables map each target invocation to its own result root.

| Target invocation | Run root | Result |
| --- | --- | --- |
| `env -u RESULTS_DIR make agent-finalize` | `20260918T013700Z-p3362375` | PASS 1/1 graph units |
| `make frontend-typecheck` | `20260918T013731Z-p3367270` | PASS 2/2 graph units |
| `make frontend-unit` | `20260918T013731Z-p3367276` | FAIL 646/647 graph units |
| `make generate-drift` | `20260918T013731Z-p3367110` | PASS 4/4 graph units |
| `make generated-artifact-policy-check` | `20260918T013731Z-p3367116` | PASS 3/3 graph units |
| `make json-shape-check` | `20260918T013731Z-p3367122` | PASS 3/3 graph units |
| `make test-catalog-check` | `20260918T013754Z-p3400636` | PASS |
| `make frontend-fallow-static` | `20260918T013811Z-p3408472` | PASS 2/2 graph units |
| `make browser-e2e-a11y` | `20260918T013731Z-p3367503` | PASS 20/20 graph units |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_settlement,module.timeline.browser.range_entry_native,module.timeline.browser.range_entry_lifecycle,module.timeline.browser.range_accessibility,module.timeline.browser.range_authority,module.timeline.browser.range_cancellation,module.timeline.browser.range_departure,module.timeline.browser.range_gestures,module.timeline.browser.range_live_membership,module.timeline.browser.range_presentation,module.timeline.browser.range_rejection_query,module.timeline.browser.range_scrolling,module.timeline.browser.spreadsheet_keyboard,module.timeline.browser.spreadsheet_creation,module.timeline.browser.spreadsheet_pointer,module.timeline.browser.spreadsheet_rejection,module.timeline.browser.spreadsheet_refresh,module.timeline.browser.spreadsheet_virtualization,module.timeline.browser.spreadsheet_bulk,module.timeline.browser.clipboard_fidelity` | `20260918T013900Z-p3426034` | PASS 13/13 graph units |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.grid_autosave_editor_families,module.workbook.browser.the_explicitly_visited_shared_grid_keyboard_anch_2b5da548a2,module.workbook.browser.reference_grid_commit,module.workbook.browser.reference_admission_lifecycle,module.workbook.browser.reference_session_recovery,module.workbook.browser.find_edit_departure,module.workbook.browser.find_scope_lifecycle,module.workbook.browser.find_authority,module.workbook.browser.find_source_editors` | `20260918T013900Z-p3426039` | PASS 15/15 graph units |
| `make frontend-import-boundary-check` | `20260918T014306Z-p3547745` | PASS 2/2 graph units |
| `make lint-biome` | `20260918T014306Z-p3547751` | PASS 2/2 graph units |
| `make lint-markdown` | `20260918T014238Z-p3545559` | PASS |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_428c33b58f` | `20260918T014345Z-p3548937` | FAIL 1/2 graph units |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell__sentinel_grid_anchor_shell_suppor_428c33b58f` | `20260918T014445Z-p3550209` | PASS 2/2 graph units |
| `env -u RESULTS_DIR make agent-finalize` | `20260918T014507Z-p3550881` | PASS 1/1 graph units |
| `make frontend-typecheck` | `20260918T014535Z-p3554763` | PASS 2/2 graph units |
| `make lint-biome` | `20260918T014535Z-p3554783` | PASS 2/2 graph units |
| `make generate-drift` | `20260918T014535Z-p3554645` | PASS 4/4 graph units |
| `make generated-artifact-policy-check` | `20260918T014535Z-p3554650` | PASS 3/3 graph units |
| `make json-shape-check` | `20260918T014535Z-p3554654` | PASS 3/3 graph units |
| `make frontend-import-boundary-check` | `20260918T014535Z-p3554771` | PASS 2/2 graph units |
| `make frontend-unit` | `20260918T014535Z-p3554757` | PASS 647/647 graph units |

The final frontend pass resolves the initial stale sentinel expectation; no
production assertion was weakened and no runtime code changed during terminal
validation. Initial failing roots remain in this handoff with their dispositions.
The passing final source is covered by current functional/security/browser,
accessibility, visual, measurement, unit, type, style, ownership and generation
evidence. Earlier visual and measurement evidence remains applicable because the
only subsequent edit was the documented sentinel test correction and this handoff.

### RTE-05 final scope review

`git status --short --branch` and `git rev-parse HEAD` confirm `main` remains at
`eb043ab61c75804f5e2b1a15181c4ff9bfad3071`. There are 28 changed/new files,
all within the listed seam. `git diff --name-only -- docs/cartulary-ui-ux-refactor-digest pnpm-lock.yaml go.sum db internal cmd contracts`
is empty. `git diff --check` passed. No staged change, commit, push, deployment,
analyst-data modification, golden promotion or dependency change was performed.

No applicable acceptance obligation remains blocked. Full backend, all-product
browser/visual suites and release verification are not expanded beyond the
affected owner slices: no corresponding backend/release input changed. All
applicable four Timeline measurement rows and affected production visual rows
were executed. Retained-run maintenance alone is skipped because RESULTS_DIR
is unset; selected runs are not substituted for a full warm run.

Final next action: repeat documentation/whitespace checks covering this saved
terminal status, then stop after this seam. Changes remain reviewable and
uncommitted.


### RTE-05 exit

`make lint-markdown` passed on the completed handoff at
`.cartulary/test-results/20260918T015008Z-p3626628`, summary
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` passed.
Every applicable acceptance row is PASS; N/A rows have owner/scope rationales.
There is no applicable BLOCKED dependency or unresolved seam defect. RTE-05 is
saved DONE only after these results, and the terminal status receives a final
Markdown/whitespace recheck. RTE-01 through RTE-05 are complete.
