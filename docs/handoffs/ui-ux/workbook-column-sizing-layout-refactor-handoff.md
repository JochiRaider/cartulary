# Workbook column sizing and saved-layout fidelity

## Execution control

Baseline revalidated before editing: clean `main`, HEAD
`031b3b0b175266f9ddf87f55af36048df5da8748`. Root AGENTS.md is the only
applicable agent instruction file. There is no pre-existing work to overwrite.
The user approved the complete CSL plan. Only the current workstream is
IN_PROGRESS; its exit evidence and DONE status must be saved before its successor
begins. An unresolved adopted-owner contradiction blocks dependent work.

| Workstream | Status | Dependency / exit |
| --- | --- | --- |
| CSL-01 Characterization and sizing contract | DONE | Owner map, scope, measurement bounds and binary criteria recorded below. |
| CSL-02 Sizing and measurement ownership | DONE | Focused sizing, obsolete-result and type checks pass; evidence below. |
| CSL-03 Controls, saved layouts and continuity | DONE | Shared consumers, controls, saved-view and retained-work checks pass; evidence below. |
| CSL-04 Production interaction evidence | DONE | Production sizing, persistence, continuity, a11y, measurement and two fresh visual passes recorded below. |
| CSL-05 Terminal validation and handoff | DONE | Terminal validation, completed acceptance/checklist and final scope review pass. |

## Authority, inspected paths and bounded scope

Behavior: Core 01 REQ-01-143; Core 03 REQ-03-295, REQ-03-298/299,
REQ-03-099/100 and REQ-03-218/300; Core 04 authorization; design §§7–8 and 14.
Domain owns vocabulary/navigation only. The localized digest read order and the
view-bar, grid-edit/autosave, query-continuation, clipboard, saved-view authoring
and discovery, recovery-navigation and authoring-candidate handoffs were inspected
during planning. Those documents and research/nlspec-spec.md are advisory or
implementation evidence, not additional behavioral authority or fresh passes.

Source inspection covers Workbook layout/controller/facade, query/layout codecs,
Columns/Grid controls, saved-view controller/adapter/startup, continuity and retained
editor ownership, Timeline column assembly/defaults, Entity/Assessment/Generic
bindings, Grid Adapter core/compiler/production binding, installed RDG 7.0.0-beta.59,
and Network Analysis layout consumers. Source placement is separately projected by
tools/frontend_source_ownership.json and tools/frontend_import_boundaries.json.
Verification routing comes from contracts/verification and tools/test_families.

Authorized changes are this shared frontend seam, necessary owner clarification,
authored machine projections, generated derivatives, tests and source guides.
No backend, dependency, database/storage migration, browser persistence, unrelated
chrome, row-height redesign, freeze panes, formula/query capability or analyst-data
change is planned. No digest edit, commit, push or deployment is authorized.

## Planning preflight (not implementation acceptance)

Public Make discovery and task guides were read for web.workbook,
package.grid_adapter, module.workbook, module.savedviews, web.architecture,
web.networkflow and package.ui. Fresh cache-disabled narrow runs passed:

| Evidence | Result | Run root |
| --- | --- | --- |
| Workbook layout/query, dirty comparison, Timeline defaults, controls and saved-resource validation | PASS, 6/6 harness units | .cartulary/test-results/20260917T025327Z-p2512491 |
| Production Grid Adapter component suite | PASS, 2/2 harness units | .cartulary/test-results/20260917T025327Z-p2512490 |
| Selector-policy rows | PASS, 5/5 harness units | .cartulary/test-results/20260917T025422Z-p2514266 |

Initial cache-enabled runs reused valid evidence; the runs above disabled cache.
The historical selector-policy failure does not reproduce. Component tests do not
establish browser geometry. Production-browser and terminal evidence remain due.

## Acceptance and rollback

Applicable acceptance rows require PASS; N/A requires scope/owner rationale.
Applicable BLOCKED rows prohibit completion. Tests use isolated harness fixtures.
Rollback must restore this change's owners, authored inputs, generated outputs,
implementation, tests and any justified goldens as one reviewed set, preserving
later user work and analyst data. No persisted-layout migration is planned.

## CSL-01 characterization and owner decisions

Core 03 REQ-03-295 now states the approved focus-borrowing, sparse default,
reset, bounded fit and obsolete-result rules. Design §8.3 describes their
presentation, numeric validation, keyboard step, rounding, focus and feedback.
Core 01's wire grammar and range are unchanged. No owner contradiction remains.

| Entry / consumer | Existing owner and observed behavior | Adopted sizing result |
| --- | --- | --- |
| Header drag and Ctrl/Cmd+Arrow | Installed RDG ResizeHandle/HeaderCell; already supported | Preserve gestures; normalize CSS coordinates and enforce declared bounds through Adapter. |
| Double-click | RDG max-content measurement; includes rendered vendor content | Same bounded semantic Fit as Columns; no competing auto-fit. |
| Default/explicit width | Workbook field default builders, sparse layout controller and codec | Keep adaptive Timeline defaults; explicit equal-to-default stays explicit; restore deletes one override. |
| Visibility/order/Reset columns | Shared Workbook layout/controller and Columns | Full semantic permutation persists; schema defaults reset columns only. |
| Saved selection/startup | Saved-resource observer, startup admission and layout application | Apply canonical widths, invalidate obsolete fit even for equal layout. |
| Create/update/duplicate/dirty/Reset | Existing saved-view operation owner and codecs | Capture working layout, duplicate selected saved layout, compare sparse state, restore saved/default configuration. |
| Timeline/Entity/Assessment/Generic | Source column builders and shared adapter | One bound sizing capability, source-owned defaults and committed-content eligibility. |
| Network Analysis | Extension-local layout and declared minimum widths | Neutral adapter capability; no Workbook layout schema or persistence semantics. |
| Empty/grouped/read-only/all hidden | Existing query/interaction and layout owners | Header-only fit when no eligible visible cells; hidden/off-screen fit unavailable; readable local customization remains available. |

Measurement intersects actual viewport geometry, excluding overscan, drafts,
local overlays and structural/group rows. Production DOM copies retain density,
fonts, padding, border and summary indicators without a parallel React renderer.
No fetch, scroll, collection expansion or virtualization change is permitted.
Missing geometry/unready fonts leave state unchanged. Clamp long content and
report the maximum. Fit requests bind field, mount, configuration, layout and
presentation generations; obsolete work cannot apply or focus.

| Gap / classification | Remediation / affected areas | Rationale and long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Confirmed: vendor minimum 50 versus portable 40, no adapter maximum | Project bounds, compile min/max and unify gestures; Workbook and Adapter | Rendered geometry agrees with saved configuration | Same layout/API; zoom remains a browser hypothesis | 40 and 4096 render and persist exactly; gestures cannot escape range. |
| G2 Confirmed: vendor fractional callbacks rejected by integer setter; codec truncates separately | Normalize measurements once; share validation and remove codec truncation | No silently lost gesture or competing normalization | Malformed saved resources remain rejected | Fractions normalize; explicit fractional input rejects without unrelated changes. |
| G3 Product improvement: no explicit width/fit/default controls | Accessible per-column sizing panel in Columns | Discoverable keyboard parity and precise restoration | UI/internal types only; narrow/zoom focus needs browser evidence | Every action works by keyboard, including all-hidden layouts. |
| G4 Confirmed structural weakness: vendor max-content scope includes drafts and lacks request lifetime | Adapter-private bounded DOM measurement and Workbook request fencing | One measurement owner; no stale width overwrite | No new storage; typography/collection copies need geometry evidence | Only eligible visible committed content measures; cancellation/replacement cannot apply. |
| G5 Continuity hypothesis: layout controls blur editors and hidden fields can detach | Clarified owner exception; existing draft/session retention and semantic binding | Preserve work without a second recovery owner | Ordinary commit/navigation unchanged | Sizing emits no write; exact drafts and semantic selection survive or remain recoverable. |
| G6 Evidence gap: resize visual specimen is synthetic HTML | Add production-grid browser/visual scenarios | Evidence exercises the shipped composition | Isolated fixtures only; no broad golden churn | Geometry, lifecycle and reviewed production captures pass. |

Acceptance: each gap's binary criterion plus saved-view lifecycle, current
authorization, spreadsheet navigation, clipboard/fill and extension regressions.
Principal risks are DOM intrinsic measurement, editor blur, zoom coordinates,
late results and source/runtime type coupling. No backend/schema migration is
needed. Source evidence and prior preflight are not new browser passes.

CSL-01 exit: owner mapping, all sizing paths, compatible persistence semantics,
measurement bounds and binary criteria complete. Next action: CSL-02.

## CSL-02 implementation and exit evidence

Added a sole WorkbookColumnLayoutController working-layout store with synchronous
command/configuration cancellation, sparse default deletion and observational
binding descriptors. The React hook adapts that owner. Workbook validation uses
one typed presentation projection; the saved codec no longer truncates fractions.

Grid Adapter exposes neutral sizing intents, measurement capability, maximum
bounds and committed-content eligibility. It measures inert production DOM
copies, excludes structural/draft/editor presentations, checks viewport and font
availability, and cancels obsolete or timed-out work. Pointer/header keys use
CSS-coordinate normalization. RDG virtualization/rendering is retained. The old
resize callback remains temporarily for current consumers and will be retired
in CSL-03, not kept as an indefinite compatibility path.

| Command | Result / evidence |
| --- | --- |
| make generate | PASS; .cartulary/test-results/20260917T031123Z-p2529654 |
| make frontend-typecheck | PASS 2/2; .cartulary/test-results/20260917T031124Z-p2530277 |
| make test-slice OWNER=web.workbook ROWS=web.workbook.regression.column_sizing_contract | PASS 2/2; .cartulary/test-results/20260917T031147Z-p2533185 |
| make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.column_sizing_contract | PASS 2/2; .cartulary/test-results/20260917T031148Z-p2533428 |

Intermediate catalog validation rejected unsorted new test titles; corrected the
authored rows and regenerated. Initial generation failure 20260917T031100Z-p2522163
and typecheck preflight share that cause. `make format` at
20260917T031114Z-p2525384 formatted authored sources but reported four remaining
lint errors with diagnostics truncated by existing informational findings;
further lint investigation is due. No unrelated files changed.

CSL-02 exit: focused portable-width, default, adapter-intent and cancellation
tests pass; typecheck passes. Geometry, consumer integration and browser evidence
remain due. Next action: CSL-03 consumer/control bindings.

## CSL-03 implementation and exit evidence

All four Workbook surface families register their existing GridHandle capability
and declared defaults through useWorkbookColumnSizingBinding. The facade and view
bar share the existing layout owner. Timeline retains adaptive defaults; ordinary
surfaces keep their declared defaults. Explicit committed-content eligibility
excludes local overlays and retained Timeline collection/scalar drafts.

Columns now uses a labelled non-modal dialog with native visibility checkboxes,
move buttons and per-column Width forms. Validation preserves invalid text;
Escape/Cancel returns to the invoking Width button, closing returns to Columns,
and outside dismissal retains destination focus. Apply, Fit, Restore and Reset
have bounded descriptions and concise feedback. All-hidden layouts remain usable.
Local controls use the existing editor external-action boundary; collection blur
now retains exact drafts when borrowing focus. No new draft/recovery store exists.

Network Analysis consumes the neutral intents and measurement port with its own
minimum widths and session layout. Removed the legacy Grid Adapter width callback,
vendor auto-fit/resizing path and the Timeline fixture's duplicate width store.
Browser helpers now use native checkbox/button semantics. Source guides and the
authored source ownership manifest describe the extension boundary.

| Command / slice | Result / run root |
| --- | --- |
| make frontend-typecheck | PASS 2/2; .cartulary/test-results/20260917T033055Z-p2609508 |
| make frontend-import-boundary-check | PASS 2/2; .cartulary/test-results/20260917T032950Z-p2607942 |
| Workbook sizing and Columns control rows | PASS 3/3; .cartulary/test-results/20260917T032950Z-p2607834 |
| Full package.grid_adapter unit slice | PASS 50/50; .cartulary/test-results/20260917T032528Z-p2551586 |
| Saved-view hook, committed autosave and spreadsheet anchor rows | PASS 4/4; .cartulary/test-results/20260917T032638Z-p2591437 |
| Network Analysis session layout row | PASS 2/2; .cartulary/test-results/20260917T032637Z-p2591218 |
| Timeline collection retained-draft row | PASS 2/2; .cartulary/test-results/20260917T032840Z-p2602029 |
| make format | PASS 2/2; .cartulary/test-results/20260917T032934Z-p2603214 |

Integration type failures were missing migrated fixture props/imports, a nullable
binding guard and an explicit async return type; repaired and revalidated. The
first updated sizing run (032528Z-p2551579) asserted before error-handling promise
completion; tests now await the semantic command completion. Lint diagnostics
identified a false-positive focused-test rule on a method named fit and a forEach
return; renamed fitVisible and used a void callback. A direct wrapper invocation
was diagnostic only after the public target rejected an unsupported diagnostic
flag; passing Make-owned format/type/import checks above are the evidence.

CSL-03 exit: shared consumers compile, scoped controls and retained-work tests pass,
and unchanged saved-view capture/reset ownership passes. Production geometry,
real saved persistence and browser continuity remain CSL-04 obligations. Next:
production-browser sizing and regression evidence.

## CSL-04 production investigation and exit evidence

Production browser evidence confirmed additional gaps; these are implementation
findings, not new behavioral owners:

| Gap | Remediation / affected areas | Rationale / long-term benefit | Compatibility / remaining risk | Binary validation |
| --- | --- | --- | --- | --- |
| G7 Confirmed: supplying RDG a sparse width map without its change callback seeds a second private width cache on remount | Remove the Adapter columnWidths prop and all callers; compiled semantic GridColumn.width is authoritative | Saved/startup/reset and grouped remount geometry follow the working layout | Internal package callers migrated; cartulary.layout.v1 unchanged | Saved 4096 startup, selected duplication, reset and grouped Fit render the commanded width. |
| G8 Confirmed source gap: a wide grid can extend past a clipped parent; grouping can replace its scroll root | Intersect the window and clipping ancestors; observe the current mounted root, ancestors, font and visibility events | Fit scope matches the visible presentation and cancels across root replacement | Adapter private only; no virtualization change | Off-screen fields cannot fit; off-screen/overscan cells do not affect width; grouped/header-only fitting passes. |
| G9 Confirmed browser regression: replaced empty-state callbacks abort pending semantic root focus | Keep cancellation with the existing operational focus lifetime, ending on departure/new action/user navigation | Incidental sizing capability publication cannot cancel query-action focus recovery | Existing focus owner only, no new recovery state | Production Clear filters returns focus to the empty grid; callback replacement test retains the request and unmount cancels it. |

The initial production suite exposed G7 rather than a saved API failure. The
first viewport fixture omitted view_schema_id on patch and later collapsed only
one of two capture-state groups; repaired the fixture. Authority assertions were
initially ambiguous between toolbar and empty-state Add row; now use its shared
selector. A test-only stylesheet handle used Node.remove; changed to parent
removeChild after current typecheck diagnostics. These failures were investigated,
not waived. Relevant failed roots: 033539Z-p2618443, 034044Z-p2689885,
034206Z-p2752533, 034828Z-p2790018, 035228Z-p2865670,
040313Z-p2983787 and typecheck 040300Z-p2983315 / 040501Z-p3088241
(all under .cartulary/test-results/20260917T...).

Fresh full visual run 035344Z-p2933570 reproduced G9 and reported one intended
image difference. After the focus fix, narrow visual 040314Z-p2984053 completed
functional assertions and reported only workbook-view-bar-long-columns.png.
Reviewed its actual production image: native checkbox focus, full field labels,
consistent themed movement/Width buttons, bounded scrolling and unchanged grid
framing. Accepted refresh trigger: approved Columns interaction contract/design
§8.3. Pinned renderer and golden manifest reconciliation passed; no renderer,
mask, tolerance, fixture geometry or normalization change is authorized.

| Current evidence | Result / run root under .cartulary/test-results/ |
| --- | --- |
| Production spreadsheet keyboard, pointer, rejection, creation, refresh, virtualization and bulk rows | PASS 11/11; 20260917T034044Z-p2689892 |
| Production autosave/recovery availability, invalid timestamp, role/read-only, session, refresh and exact replay rows | PASS 11/11; 20260917T034828Z-p2790035 |
| New sizing gesture/saved, viewport and accessibility rows | PASS within mixed run 20260917T035228Z-p2865670; all-row rerun pending |
| Saved-view stateful round trips | PASS within mixed visual/stateful run 20260917T035228Z-p2865690; visual failure independently recorded |
| Network Analysis claimed discovery/import and bounded production virtualization | PASS 14/14; 20260917T040343Z-p3041802 |
| Adapter sizing plus full production component row, including empty-state focus | PASS 3/3; 20260917T040502Z-p3088396 |
| Current frontend typecheck | PASS 2/2; 20260917T040625Z-p3124704 |

Public diagnostic flag attempts were rejected as unsupported Make inputs; they
changed nothing. Raw pinned typecheck/Biome wrappers were used only to reveal
suppressed diagnostics, followed by Make-owned checks. Golden refresh and two
ordinary passes, collection-fit evidence and terminal validation remain due.

Collection/viewport rerun PASS 11/11 at 20260917T041207Z-p3270957. The
intermediate 040626Z-p3125080 and 040949Z-p3199884 failures addressed the wrong
collection presentation selector (scalar, then inspector); the production grid
selector now exercises committed summary/overflow content and unchanged versions.
The other three new rows, including all-surface/read-only and a11y, passed at
040626Z-p3125080. Current typecheck PASS 2/2 at 041219Z-p3297998.

Golden update PASS 12/12 at 20260917T040547Z-p3089763. Reconciliation: 252
active/committed captures, zero missing/orphan/ambiguous/unresolved entries, zero
errors. Only workbook-view-bar-long-columns-linux.png and its machine-owned hash
changed (new SHA-256 b93cb4447222f20f7c0de16c25389e06281808de526ee28fed2160a4b31b532c).
Reviewed that committed image and actual sizing-panel/a11y captures at normal,
1024-wide, 200%-zoom and text-spacing presentations. Buttons wrap, labels remain
readable, keyboard focus is visible and no controls require horizontal scrolling.
Two fresh ordinary visual passes remain in progress.

Timeline measurement/clipboard run 20260917T040637Z-p3142278: clipboard PASS;
100-sample p95 typing 42.1/100 ms, pointer focus 36.3/100 ms and ArrowDown
33.2/100 ms PASS. Blank-row creation measured 157.8/150 ms (FAIL); it is not waived.
The retained measurement summary/observation under browser-e2e-measurement names
predicate perf.timeline_blank_row_create.v1. This invocation overlapped separate
visual/build work; a quiet, isolated repeat is required before attributing cause.

The refreshed image belongs to authored row
module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc,
scenario_5974e42bb8fb, capture visual.capture.1f0c5e239a5b50226642, Chromium.
It is an active nonregistry capture (no stable fixture ID), reconciled exactly;
viewport 1440×900, compact density, 100% zoom, masks and screenshot scope did not
change. First ordinary post-refresh visual PASS 12/12 at
20260917T041140Z-p3241981. Existing Grid Adapter accessibility PASS 11/11 at
20260917T041417Z-p3312616.

Isolated blank-row creation repeat PASS 14/14 at 20260917T041611Z-p3344532:
100 samples, p50 93.0 ms, p95 114.6 ms against the unchanged 150 ms threshold.
No other validation job ran during this repeat. The earlier threshold failure is
retained as competing-load evidence; no implementation, sample count, admission
rule or threshold was changed to obtain this pass. Additional production checks
now explicitly replace a saved view during held measurement and distinguish Reset
columns from saved-view Reset; final row validation is pending.

Final combined sizing browser/accessibility slice PASS 13/13 at
20260917T042212Z-p3455156, with cache disabled. All four authored rows pass:
module.workbook.browser.column_sizing_gestures_saved,
module.workbook.browser.column_sizing_surfaces,
module.workbook.browser.column_sizing_viewport and
module.workbook.accessibility.column_sizing. This includes held measurement during
saved-view replacement, preserved destination focus, selected-configuration
duplication, startup/update/default removal, both reset scopes, 40/4096 geometry,
invalid text, CSS zoom, active draft/caret, visible collection summaries, capped
long tokens, virtualization, grouping, cancellation and readable closed incidents.

The 042038Z-p3410758 test failure was an off-screen virtualized header after saved
Reset preserved horizontal scroll. The corrected assertion uses the owned grid
scroll helper before inspecting its geometry; product scroll/focus behavior was
not changed. Browser evidence is Chromium/pinned Linux support evidence, not a
new release or universal assistive-technology claim. No unresolved product defect
is being classified as unrelated.

Second ordinary post-refresh visual PASS 12/12 at
20260917T041945Z-p3377291. Both fresh ordinary passes ran with cache disabled,
used unchanged pinned renderer/tolerances and reconciled the complete visual
inventory. No other golden was updated.

CSL-04 exit: PASS. Lifecycle, saved persistence, geometry, native controls,
keyboard/editor continuity, representative authorization, accessibility, extension
and production-browser evidence passes. The timing failure was rerun in isolation
without weakening its criterion. Next action: begin CSL-05 terminal owner slices
and finalizer; RESULTS_DIR remains unset because no full warm-check run qualifies.

## CSL-05 terminal validation

Current public commands were rediscovered through make help, make help-all and
module-author task guides for web.workbook, package.grid_adapter,
module.workbook, module.savedviews, web.architecture and package.ui. The guides
remain routing evidence, not behavioral authority. Begin full narrow owner unit
slices, then agent-finalize before broader terminal checks. Principal remaining
risk: generated/source ownership drift or finalizer maintenance outside this seam.
Preserve a pre-finalizer changed-path inventory; isolate unrelated changes.

Terminal full owner slices (cache disabled): web.workbook PASS 271/271 at
20260917T042455Z-p3489886; package.grid_adapter PASS 50/50 at
20260917T042455Z-p3489905. The current owner selector includes its routed
browser/static evidence as well as units; module.workbook remains running.
No failure is waived. The pre-finalizer tracked/untracked inventory is retained
in /tmp/csl-before-finalize-{tracked,untracked}.txt for scope comparison.

## Acceptance assessment

| Criterion / owner | Result | Evidence and practical boundary |
| --- | --- | --- |
| G1/G2: semantic 40..4096 bounds; integer input versus measured rounding (Core 01 REQ-01-143) | PASS | Shared validator/compiled constraints; production 40/4096, fractional drag, CSS zoom; finite/nonfinite and nearby-invalid unit cases. |
| G3: discoverable native controls, validation and focus return (design §§8/14) | PASS | New keyboard/a11y row, component tests, all-hidden and long-label/zoom captures. No numeric form remains inside menu semantics. |
| G4/G8: bounded committed viewport Fit (Core 03 REQ-03-295) | PASS | Production row/column virtualization, off-screen field rejection, collection summary/overflow, empty/collapsed header-only, draft exclusion, cap feedback; font/geometry unit rejection. |
| G4: cancellation and completed-width stability (REQ-03-295) | PASS | Owner tests fence field/configuration/layout/binding/context; production held frames cover dismissal, newer width, saved replacement and authority transition. No fetch or implicit refit path. |
| G7: saved configuration fidelity (REQ-03-026/295) | PASS | Service-backed create/update/duplicate/selection/startup/reset and sparse deletion; real browser geometry after replacement and grouped remount. No vendor width map remains. |
| Reset distinctions and equal-to-default explicit widths (REQ-03-295) | PASS | Single-default deletion preserves other state; Reset columns restores schema configuration and selected identity; saved Reset reapplies selected configuration. Explicit-default unit case stays overridden. |
| Schema field validation/evolution (Core 01 REQ-01-143) | PASS | Existing portable-field permutation, unknown/structural field exclusion and strict saved-resource tests pass in web.workbook. No new schema-evolution branch or silent saved-resource repair. |
| G5: editor borrowing, exact retention and recovery (REQ-03-218/298/300) | PASS | Production caret/raw draft survives pointer/control sizing with zero writes; existing valid/invalid/rejected/hidden/query/departure recovery rows and collection external-action unit pass. Scalar retention ownership is reused; collection blur needed the explicit borrowing marker. |
| Spreadsheet interaction (REQ-03-218/300) | PASS | Existing production keyboard, pointer, acceptance/rejection, creation/navigation, range/clipboard/fill, virtualization and refresh rows; AC-043 typing/focus/selection/creation checks pass. |
| Readability and current authority (Core 03/04) | PASS | Closed-incident local sizing remains usable and cancels pending work; role/session/availability recovery browser rows preserve/conceal owner presentation. Local widths do not require record or saved-view write permission. |
| Shared extension compatibility | PASS | Network Analysis retains its own session layout/minima; claimed discovery/import and actual production virtualization measurement pass. No Workbook serialization enters Adapter. |
| G6/G9: production rendering, accessible focus and reviewed visuals | PASS | Existing Adapter a11y, operational action focus unit/browser regression, new sizing a11y, reviewed real-grid captures, exactly one refreshed golden and two ordinary 12/12 visual passes. |
| Executable authority boundary | PASS | Typed authored inputs precede generation; no runtime/test/generator reads Markdown. Source placement and test routing remain separate from adopted owners. |
| Terminal checks and scope review | PASS | Finalizer, affected owner slices, types, lint, boundaries, JSON/generated checks and reviewed diff pass; exact results below. |
| Storage/dependency/migration and excluded features | N/A | No database, dependency, persisted wire/storage/browser-persistence change; row-height/freeze/formula/multi-column/query/chrome redesign are outside this seam. |
| Full warm-check retained maintenance / release claim | N/A | No qualifying successful full warm check exists; RESULTS_DIR intentionally unset. This is dirty-tree implementation evidence, not a release claim. |

No adopted-owner contradiction remains. G1–G4 and G7–G9 remediations retain the
existing saved format/APIs; G3 is an intentional product improvement. G5's scalar
continuity hypothesis is resolved by production evidence and reuse of the existing
retention boundary; its collection blur path was repaired. G6 is an evidence gap
closed by production composition tests, not by promoting the synthetic specimen.

## Terminal results and final scope

All run roots below are under .cartulary/test-results/. Harness unit counts are
not assertion counts. Exact selected ROWS and execution identities are retained
in each run-manifest.json; summaries and browser/measurement artifacts are under
the same root. Follow-up owner selection used every active non-Playwright row in
the named authored family manifest; it did not invent new verification ownership.

| Command / selection | Result | Run root |
| --- | --- | --- |
| make test-slice OWNER=module.workbook CARTULARY_HARNESS_CACHE_MODE=off | Initial FAIL 101/103: one support scenario and its aggregate; all other routed evidence passed | 20260917T042455Z-p3489922 |
| make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_support.verify_browser_command_helpers_for_sort_filter_g_cfa68b33e4 CARTULARY_HARNESS_CACHE_MODE=off | PASS 11/11 after native-checkbox assertion repair | 20260917T043337Z-p3628137 |
| env -u RESULTS_DIR make agent-finalize | PASS 1/1; no extra changed paths or unrelated finalizer changes | 20260917T043257Z-p3624092 |
| make test-slice OWNER=module.savedviews (active non-Playwright rows, cache off) | PASS 10/10 including service-backed Go rows | 20260917T043337Z-p3628227 |
| make test-slice OWNER=package.ui (active non-Playwright rows, cache off) | PASS 10/10 | 20260917T043423Z-p3678935 |
| make test-slice OWNER=web.architecture (active non-Playwright rows, cache off) | PASS 12/12; current selector and source/package policies | 20260917T043430Z-p3680478 |
| make frontend-typecheck | PASS 2/2 | 20260917T043337Z-p3628242 |
| make lint-biome | PASS 2/2 | 20260917T043337Z-p3628313 |
| make frontend-import-boundary-check | PASS 2/2 | 20260917T043337Z-p3628283 |
| make lint-markdown | PASS | 20260917T043337Z-p3628351 (adhoc/lint-markdown/tool-run-summary.json) |
| make json-shape-check | PASS 3/3 | 20260917T043337Z-p3627996 |
| make generated-artifact-policy-check | PASS 3/3 | 20260917T043337Z-p3627987 |
| make generate-drift | PASS 4/4 | 20260917T043337Z-p3627977 |
| git diff --check and final changed-path review | PASS | Reviewed in the repository worktree; only the listed seam paths changed. |

The broader module sweep found one remaining obsolete aria-checked expectation
in workbook.support.spec.ts after its role had already migrated to checkbox.
Changed it to not.toBeChecked(), preserving the original hidden-column assertion.
The exact failed support row then passed. This is a test projection repair for
native checkbox semantics, not a waived product failure. No other failure from
the full module sweep remains unresolved.

agent-finalize's unit-artifacts/finalize-summary.json reports pass and
results_dir=null. Canonical retained-evidence and scheduler/timing maintenance
were skipped with results-dir-not-provided. There was no qualifying full warm
check; no stale run was supplied. The finalizer introduced no unrelated edits.
Repository-wide frontend-unit, test-fast/check, CI/release and unrelated backend
security/build targets were not run: applicable owner slices, production browser,
service-backed persistence, full visual passes and affected static/type/drift
checks cover this bounded frontend seam. No release/conformance claim is made.

### Substantive changes and retired paths

WorkbookColumnLayoutController is the only working-layout store. Its mounted
bindings provide defaults and observational capabilities, not a second width map.
The adapter owns CSS-coordinate gestures and inert production-DOM measurements.
Controls and header intents reach the same semantic commands. Timeline preserves
adaptive omitted widths; explicit widths remain sparse, fixed and saveable.

Retired paths: onColumnWidthChange; the separate Adapter columnWidths input and
RDG private width-cache seeding; RDG's competing resize/auto-fit callbacks;
fraction truncation in saved-layout decoding; menu semantics for editable Columns
forms; the Timeline test fixture's duplicate width state. Useful RDG rendering,
virtualization, sort/reorder, navigation and structural columns remain in place.
The operational empty-state focus repair stays in its existing owner.

Another Workbook surface supplies semantic GridColumn definitions/defaults and
committed-presentation eligibility, applies applyWorkbookLayoutToColumns, calls
useWorkbookColumnSizingBinding with its existing GridHandle and layout commands,
and forwards onColumnSizingIntent. Columns consumes commands.sizing. No vendor
imports, DOM coordinates, new width store or persistence implementation are needed.

### Compatibility, limitations and rollback

cartulary.layout.v1 and saved-view APIs are unchanged. No dependency, database,
storage, browser-persistence or data migration is needed. Neutral internal Adapter
callers were migrated together; extensions keep their own limits/layout ownership.
A future internal caller uses semantic intents/capabilities instead of the retired
width callback/map props. Structural state never enters saved layout.

Fit deliberately excludes unavailable, off-screen and uncommitted presentations.
Timeline conservatively excludes pending local row presentations until its owner
reconciles them. Collections measure their displayed summaries and indicators,
not undisclosed members. Later content/font/query/viewport changes require another
explicit Fit. Tests and captures use isolated harness data and the supported
Chromium/Linux renderer; manual screen-reader/other-engine release certification
is not claimed. No applicable acceptance row is BLOCKED.

Rollback restores the owner clarifications, authored sizing projection/schema and
generator, generated derivatives, runtime/control/binding changes, test/catalog
updates and the single golden/hash together. Revert only these reviewed changes;
preserve later user work. Saved resources need no rollback migration and analyst
data must remain untouched. HEAD and branch remain the original main baseline;
this task does not create a rollback commit because commits were prohibited.

### Reviewed changed-file inventory

- `apps/web/e2e/support/workbook/collections.ts`
- `apps/web/e2e/support/workbook/ordinaryCreate.ts`
- `apps/web/e2e/workbook-column-sizing.spec.ts`
- `apps/web/e2e/workbook-grid-autosave.spec.ts`
- `apps/web/e2e/workbook.support.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-view-bar-long-columns-linux.png`
- `apps/web/src/networkFlow/NetworkFlowSemanticGrid.tsx`
- `apps/web/src/networkFlow/useNetworkFlowGridLayout.test.tsx`
- `apps/web/src/networkFlow/useNetworkFlowGridLayout.ts`
- `apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx`
- `apps/web/src/workbook/WorkbookShell.query.test.tsx`
- `apps/web/src/workbook/WorkbookShell.support.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx`
- `apps/web/src/workbook/components/EntityWorkbookSurface.tsx`
- `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- `apps/web/src/workbook/components/WorkbookColumnsControl.tsx`
- `apps/web/src/workbook/components/WorkbookGridControls.test.tsx`
- `apps/web/src/workbook/components/WorkbookGridControls.tsx`
- `apps/web/src/workbook/components/WorkbookShellViewBarControls.tsx`
- `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts`
- `apps/web/src/workbook/layout/README.md`
- `apps/web/src/workbook/layout/WorkbookColumnLayoutController.test.ts`
- `apps/web/src/workbook/layout/WorkbookColumnLayoutController.ts`
- `apps/web/src/workbook/layout/useWorkbookColumnLayoutController.ts`
- `apps/web/src/workbook/layout/useWorkbookColumnSizingBinding.ts`
- `apps/web/src/workbook/layout/useWorkbookLayoutFacade.ts`
- `apps/web/src/workbook/layout/workbookColumnLayout.ts`
- `apps/web/src/workbook/models/README.md`
- `apps/web/src/workbook/models/workbookColumnSizing.ts`
- `apps/web/src/workbook/models/workbookQuery.ts`
- `apps/web/src/workbook/timeline/components/TimelineCollectionCell.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineCollectionCell.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookGrid.tsx`
- `apps/web/src/workbook/timeline/components/useTimelineColumnAssembly.tsx`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `contracts/design/presentation.v1.json`
- `docs/design.md`
- `docs/handoffs/ui-ux/workbook-column-sizing-layout-refactor-handoff.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `packages/grid-adapter/README.md`
- `packages/grid-adapter/src/ColumnResizeHandle.tsx`
- `packages/grid-adapter/src/GridOperationalStatePlane.tsx`
- `packages/grid-adapter/src/SemanticDataGrid.tsx`
- `packages/grid-adapter/src/columnSizing.test.tsx`
- `packages/grid-adapter/src/columnSizing.ts`
- `packages/grid-adapter/src/core.ts`
- `packages/grid-adapter/src/index.test.tsx`
- `packages/grid-adapter/src/index.tsx`
- `packages/grid-adapter/src/rdgCompiler.tsx`
- `packages/grid-adapter/src/styles.css`
- `packages/grid-adapter/src/useGridColumnSizing.ts`
- `packages/ui-contracts/src/generated/design-presentation.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/harness/generated-artifacts/design-presentation/design-presentation.mjs`
- `tools/schemas/cartulary.design_presentation.v1.schema.json`
- `tools/test_families/module.workbook.json`
- `tools/test_families/package.grid_adapter.json`
- `tools/test_families/web.workbook.json`

### Completed handoff checklist

- Applicable behavioral/geometry/persistence/continuity/accessibility rows PASS;
  explicit N/A rows have scope/owner reasons. No owner contradiction or applicable
  BLOCKED row remains.
- Confirmed gaps, hypotheses and intentional improvements are distinguished;
  remediation, rationale, compatibility, risks and binary criteria are recorded.
- Source ownership, authored projections and verification routing remain distinct;
  no executable dependency on Markdown was introduced.
- Meaningful model/component, production-browser, service-backed saved-view,
  spreadsheet, authority/recovery, extension, measurement and visual evidence is
  retained with run roots. Every changed screenshot was reviewed.
- Exactly one justified golden/hash changed; two fresh ordinary visual passes
  passed. Virtualization, thresholds, tolerances and functional criteria were not
  weakened.
- Finalizer ran before broader terminal checks; retained-run maintenance was
  explicitly skipped. Generated inputs preceded generators; no generated output
  or lockfile was hand-edited.
- Final scope is the listed seam. The digest, dependencies, backend/storage,
  analyst data and pre-existing work were not modified. No commit, push or deploy.
- Rollback and the shared extension path are documented. Next action: reviewer
  inspection of the worktree and this evidence-backed handoff; no implementation
  dependency remains.

CSL-05 exit: DONE. Final completed-handoff Markdown lint PASS at
.cartulary/test-results/20260917T043719Z-p3687103, summary
adhoc/lint-markdown/tool-run-summary.json. Final git diff --check PASS; 62 reviewed
changed/new files, unchanged main HEAD 031b3b0b175266f9ddf87f55af36048df5da8748.
No digest, lockfile, dependency, backend or storage path changed. Every applicable
acceptance row and the completed-handoff checklist pass; no required implementation
or validation remains. Next action: review the uncommitted implementation.

## 2026-09-23 width-editor coherence follow-up

This bounded follow-up starts from clean `main` at `669cb26b3`. Core 03
REQ-03-295 owns the working layout, sparse widths, one-time Fit, cancellation
and editor-focus borrowing; Core 01 REQ-01-143 owns the `40..4096` limits;
design §8.3 owns the panel's input, feedback and dismissal behavior. Source
ownership remains `web.workbook`; verification routes through `web.workbook`
unit rows and `module.workbook` production-browser rows. The digest and this
earlier handoff were navigation and historical evidence, not product authority.
The narrow advisory query used the verified React stack and UX domain. R001,
R002 and R006 support keyboard operation, focus and local validation; R012 is
adapted to the existing controller and local input lifetime.

### Characterization and change

The new component characterization first ran against the old implementation:
Fit applied `520` px while the still-open input remained `240`. The focused
row failed exactly at that assertion, 13/14 tests passing, at
`20260924T015115Z-p59293`. Earlier attempts at `20260924T014647Z-p55011`
and `20260924T014712Z-p55820` stopped before assertions with
`service_start_error`: the shell PATH lacked the repository's pinned Node
runtime. Subsequent public Make runs placed `tmp/node-runtime/bin` on PATH.

The controller remains the sole working-layout store. Its panel commands now
return typed completed, unavailable or cancelled outcomes, with effective width
on completion. Apply, Fit and Restore publish completion through the owner's
notice; the component's separate Apply notice path is removed. Grid Adapter
header intents retain their existing width path, so drag movement does not
generate panel completion announcements. The binding still supplies defaults
and measurement capability, never a second width map.

The panel retains only unapplied text, a manual-draft marker and an edit
revision. Untouched text follows current width. A successful Fit replaces text
that preceded activation; text typed while it was pending remains exact, with
focus and caret, even though Fit applies to layout. Invalid Apply cancels a
pending Fit and leaves layout unchanged. A successful Apply canonicalizes the
input to the applied integer; Restore removes only the chosen sparse override
and displays its effective default. Cancel and Escape discard unapplied text
and pending measurement, not a completed sizing action. Existing owner fences
for newer actions, binding changes and context replacement remain in place.

### Verification and limits

| Public Make selection | Result and run root |
| --- | --- |
| Focused `web.workbook` sizing controller and component rows, cache off | PASS 3/3 at `20260924T021044Z-p35207`; includes Fit → Apply, refinement, newer invalid text/caret, passive sync, Restore and cancellation. |
| `module.workbook.browser.column_sizing_gestures_saved`, cache off | PASS 11/11 at `20260924T020207Z-p11146`; one open Timeline panel covers keyboard Fit, input/header geometry, Apply, manual refinement, Restore, focus and zero record writes. |
| `module.workbook` sizing surfaces, viewport and accessibility rows, cache off | PASS 13/13 at `20260924T020313Z-p44073`; includes viewport cancellation, cross-surface geometry and narrow/zoom keyboard checks. |
| `make generate` then `env -u RESULTS_DIR make agent-finalize` | PASS at `20260924T020523Z-p80918` and `20260924T020543Z-p84054`. Generator changed only the expected topology input digest/hash index. Finalizer made no additional changes; retained-run maintenance was skipped because no qualifying full warm-check `RESULTS_DIR` was supplied. |
| `make frontend-typecheck`, `make lint-biome`, `make frontend-import-boundary-check` | PASS at `20260924T020624Z-p90918`, `20260924T020906Z-p33353` and `20260924T015949Z-p84187`. |
| `make json-shape-check`, `make generated-artifact-policy-check`, `make generate-drift` | PASS at `20260924T020624Z-p90733`, `20260924T020624Z-p90809` and `20260924T020624Z-p90841`. |
| `make lint-markdown` and final `git diff --check` | PASS at `20260924T021209Z-p36597` and in the reviewed worktree. |
| Full `make test-slice OWNER=web.workbook`, cache off | FAIL 296/297 rows at `20260924T020613Z-p87925`: the unchanged upload-recovery test in `WorkbookShell.surfaces.test.tsx` did not find `File recovery: screenshot.txt`. Its exact row failed again in isolation at `20260924T020906Z-p33226`. No sizing action occurs in that test; this follow-up did not change upload or attachment code. The failure is retained as an unresolved broader-check limitation, not a sizing pass or a baseline-proven pre-existing defect. |

The first `agent-finalize` attempt (`20260924T020444Z-p79429`) and direct
`json-shape-check` (`20260924T020507Z-p80083`) failed because the authored
test-family manifest changed before topology regeneration. Running `make
generate` repaired that projection. The first browser attempt
(`20260924T015942Z-p74895`) failed only at a new test assumption that the Fit
button retained focus after its pending disabled state; the assertion was
replaced with an enabled-state check while keyboard activation, input focus
and Escape focus return remain tested. No product focus rule was relaxed.

No storage, saved-view wire format, dependency, backend, record mutation,
Grid Adapter interface, golden or design token changed. Fit still measures
bounded committed content on screen once; it does not refit passively. Browser
evidence is the supported Chromium harness, not universal assistive-technology
certification. No measured performance improvement is claimed. Full CI,
release and unrelated backend suites were not run because this is a bounded
frontend interaction correction.

### Digest acceptance assessment

| Rows | Assessment and evidence |
| --- | --- |
| A001–A004, A023–A026 | PASS: exact Core/design owners and current source/test routing were checked; the one controller action boundary removes a duplicate notice path; no token or Markdown product dependency was introduced; stable semantic selectors and the public generator/policy checks pass. |
| A011, A014, A019–A021 | PASS: controller/component and production rows cover raw input, caret/focus, invalid validation, keyboard operation, cancellation, zoom, cross-surface use and virtualized viewport measurement. |
| A027 | PASS for this sizing follow-up: owner map, change, retirement, verification failures and limits, compatibility and rollback are recorded here. The unrelated broad upload row remains explicitly failed. |
| A005–A010, A012–A013, A015–A018, A022 | N/A: theme, density, creation, shell/inspector/overflow, transaction and query/evidence semantics, and visual fixtures are outside this input/controller correction; no visual artifact changed. |

Rollback reverts this follow-up's width control, controller/hook, tests,
authored `web.workbook` test-family selector, generated topology render index
and this handoff addition together. Saved resources and analyst data require no
migration or rollback. The remaining action outside this sizing slice is to
triage the reproducible upload-recovery test failure under its own owner.
