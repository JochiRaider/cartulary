# Seven frontend gap remediation

All seven original gaps are implemented and have fresh passing regression
evidence. The final integrated check passes 901/901 units, both visual validation
runs pass, and the complete measurement stage passes. The broader webserver stage
still has the separately enumerated failures below; full browser readiness is
not claimed.

## Authority, scope, and compatibility

Implementation started at `cd0be11dd321e48acbdcdbc507b6371099db48f9`.
The three policy failures and four browser failures are seven separate gaps.
Their reproduction on clean `d4cdca1f15aebe8771e469e158e679fa501f5837`
establishes provenance, not acceptability or a stale-test diagnosis.

Core 01 owns sparse saved-view PATCH. Core 03 owns surface entry, editing,
admission, acknowledgement, and detachment. Design owns recovery placement.
The Testing Harness NLSpec continues to own execution mechanics only. Domain
vocabulary and owner navigation needed no change. Neither executable verification
nor runtime behavior reads this document or other Markdown.

Explicit base-surface entry prioritizes visible authorized creation controls,
then committed navigation cells, then the grid root. Ordinary editor navigation
still waits for authoritative acceptance. Explicit surface detachment retains
admitted work and refused local drafts within the live workbook runtime. No
reload persistence, wire compatibility alias, dependency, or database migration
was introduced. Maintainers should review the Core 03 clarification and its typed
projections together; catalog membership alone cannot establish owner agreement.

## Gap ledger

### G1 — Canonical schema identities

Owner: View Contracts. Browser fixtures import current identities directly from
`@cartulary/view-contracts`. The five reported consumers and adjacent ordinary
creation identities were migrated; the redundant local Evidence alias was
removed. Deliberately invalid schema probes and negative assertions remain.
This is an import-only correction with unchanged request identities. It removes
fixture drift without a second facade or compatibility namespace.

Validation: the complete architecture policy passes in
`20260914T163752Z-p78632`; the broad frontend unit run also passed this row.
All six ordinary creation/security rows pass in `20260914T174526Z-p52363`,
including the fourteen-schema matrix, detached completion, uncertainty,
closure, and revocation. Accessibility and both visual stages also pass.

### G2 — Shared recovery bounds

Owner: Workbook layout and design. `WorkbookWorkAreaOverlay` owns portal hosting,
containing-block bounds, spacing, stacking, panel entry focus, and internal
scrolling. Coordination and Note features retain only their triggers, recovery
content, actions, Escape dismissal, attachment ownership, and focus return.
The complete layout scan exposed the same legacy viewport expression in Note
recovery after Coordination was repaired, so both consumers migrated atomically.

The new zoom test exposed a second cause: the application workbook route used
viewport block units even when root CSS zoom reduced the available containing
block. `AppRoot` and the workbook route now use their containing block for height;
authoring panels no longer inherit an oversized shell. No row-height arithmetic
or fixed toolbar subtraction moved into a feature. Visual placement intentionally
changes; mutation and recovery identities do not.

Validation: layout/selector policies pass in `20260914T164700Z-p33019`.
Coordination accessibility passes in `20260914T170229Z-p81183`, including normal
1280x720, short 390x320, and 1280x720 at 200% root zoom, visible controls, status
visibility, internal scrolling, and absence of document vertical overflow.
The expanded row also covers retained drafts and accepted refresh failure at all
three sizes, with every enabled control reachable; it passes in
`20260914T171858Z-p29850`. Visual evidence is recorded below.

### G3 — Owned selectors and mounted-row observations

Owners: UI Contracts, Grid Adapter, and browser measurement support. Reference
controls use scoped role/name queries. Shared selectors are prepared before
`page.evaluate` and passed as serializable descriptors. Browser observers contain
no imported closure, source evaluation, or driver callback during measurement.
Committed row version now accompanies typed stable identity on the mounted
semantic row in both production and support bindings. The hidden Timeline
version probes and their builder were removed with every workspace consumer.

Paint qualification reads identity, committed version, intended field, and
visibility from the same mounted row. Replacement of that DOM row resets the
consecutive-frame qualification. This reduces whole-document work and prevents
matching a new version probe to an old or unrelated rendered cell. Private
selectors migrate together; public APIs and persisted records do not change.

Validation: descriptor and DOM observer coverage passes in
`20260914T163752Z-p78632`; invalid paint qualification passes in
`20260914T164424Z-p28077`. Coverage includes draft/missing/stale versions, wrong
identity/field, hidden fields, and replaced rows. Fresh complete browser measurement passes 25/25 units in
`20260914T175220Z-p5285`, including real browser execution of the serialized
observer and all four Timeline predicates. Prior performance results were not
carried forward; final sampling and p95 evidence appears below.

### G4 — Timeline Tab contract

Owner: Core 03 REQ-03-218. The keyboard scenario now asserts the next visible
field (`timeline.analyst_text`) and its actual grid-cell focus, then deliberately
restores the date anchor for the inspector assertions. Grid navigation behavior
was preserved. The later Alt+H observation now captures the event before editor
propagation containment and observes its final cancellation in a microtask.
This keeps inspector shortcut coverage observable without changing dispatch.
The unit sentinel likewise awaits Escape's completed focus return before issuing
ArrowDown instead of sending both interactions in one synchronous stack.

Validation: the complete keyboard row passes in `20260914T164202Z-p87307`;
the corrected sentinel passes in `20260914T165724Z-p61782`. Existing wrap,
boundary exit, accepted movement, rejection, and non-mutation coverage remains.

### G5 — Completed, cancellable surface entry

Owners: Core 03, Workbook, and Grid Adapter. The causal trace established that
Indicator creation fields registered, but `draftFieldKeysRef` filtered them by
`contractWritable`, the permission to patch existing records. Those creation-only
fields were excluded. Workbook then acknowledged a committed-cell fallback.
This was a product eligibility defect in addition to the ambiguous boolean API.

`GridColumn.draftWritable` now represents creation authority separately.
`GridHandle.requestFocus` returns `focused`, `unavailable`, or `cancelled`.
The common production/support controller owns mounting, scrolling, registration,
actual DOM focus, and completion. Workbook owns ordering and request generation.
Registration/layout events drive progress; eligible unmounted targets remain
pending. Abort, supersession, unmount, authority changes, and deliberate user
navigation cancel requests. Hidden and disabled controls are unavailable.

The old boolean focus API and entry retry timers were removed. Continuity,
Timeline, Generic/Entity, Network Flow modal restoration, and support bindings
use the same mechanism, with completion and cancellation propagated where a
caller owns a longer lifetime. This is an atomic private workspace API change;
there are no compatibility aliases. It prevents false acknowledgement and late
focus theft as asynchronous and virtualized surfaces expand.

Validation: the Indicator browser row passes in `20260914T163226Z-p13043` and
again in `20260914T170229Z-p81183` without manual focus repair. Adapter evidence
includes `20260914T164422Z-p27406` and `20260914T170058Z-p72319`; entry ordering
and continuity units pass in `20260914T163754Z-p80040`. Diagnostic runs
`20260914T160702Z-p41403` and `20260914T160918Z-p73227` retain the causal sequence;
temporary console instrumentation was removed.

### G6 — Real overflow admission and retained local work

Owners: Core 03, Workbook mutation runtime, Timeline editor lifetime, and browser
support. `editTimelineSummary` now declares `accepted`, `queued`, or `rejected`
and uses a bounded predicate for that outcome. Admission telemetry never proves
authoritative completion. Queue counts come from shared status metadata; failure
diagnostics contain iteration, identity, anchor, mounted identities, queue counts,
and lifecycle state, not record contents or credentials.

The fixture seeds 65 distinct rows through public operations, disconnects,
submits each real editor, and explicitly switches to Notes and back after each
of the first 64 admissions. The 65th edit is refused without eviction. Reattaching
Timeline permits its source driver to replay; the test verifies 64 exact FIFO
successes, authoritative values, zero pending units, and no stolen overflow focus.
It then explicitly retries the refused draft and observes its sole success.

This exposed a real second defect: refused raw text lived in a mounted Timeline
registry and disappeared on unmount. `WorkbookLocalDraftStore` now holds only
source-neutral scalar values, scoped by surface in `WorkbookMutationRuntime`.
Timeline capture mappings, mounted input references, and listeners remain local
to its registry. Runtime retirement clears values. The common runtime never
imports Timeline implementation.
Surface changes neither clear them nor prolong their lifetime beyond the runtime.

The overflow notice supports Escape and a named close button with focus return;
the global status action reopens it. Closing presentation does not discard the
refused draft or clear Conflict. This prevents the notice from covering the
editor indefinitely while preserving global recovery access. Ordinary pending
editors still remain pending until authoritative acceptance.

Validation: `20260914T170229Z-p81183` passes the complete save-status row,
including all 65 editor attempts, exact capacity, retained refusal, FIFO replay,
server convergence, focus stability, dismissal/reopening, and explicit retry.
The store lifetime regression passes in `20260914T164701Z-p33439`. Earlier runs
`20260914T163710Z-p46953`, `20260914T164202Z-p87307`,
`20260914T164746Z-p34842`, and `20260914T165724Z-p61774` remain failed diagnostic
evidence; later passes do not relabel them.

### G7 — Sparse saved-view PATCH

Owner: Core 01. Browser support asserts unchanged layout omission and preservation
in the response and a subsequent fetch. A separate UI column-visibility change
must appear in PATCH, persist, and restore after reload/reselection. Unit coverage
proves structural equality despite reordered object keys. The existing serializer
and authoritative no-op/version-conflict behavior remain intact; client focus,
selection, and scroll remain outside portable layout. No wire/data migration is
needed. Sparse updates avoid coupling unrelated fields and accidental overwrites.

Validation: the complete browser command-helper row passes in
`20260914T163710Z-p46953`; structural layout units pass in
`20260914T163754Z-p80040`.

## Sequencing and integration evidence

Owner clarification preceded focus/admission changes. Canonical consumers and
saved-layout corrections were independent. Shared focus/metadata and overlay
boundaries preceded overflow fixture/lifetime work. New acceptance tests are
routed through authored owner families; generated topology is produced through
Make. Integration results, visual refresh disposition, and the final check are
recorded below.

All run IDs above are under `.cartulary/test-results/` in the implementation
checkout unless explicitly stated otherwise. Selected public commands retain
the exact seven regression rows from the approved plan. Failed run summaries and
browser traces are historical evidence, not disabled coverage or policy exemptions.

## Rollback and maintainer review

Roll back Core 03 entry/admission clarification, Grid Adapter interfaces and
bindings, Workbook/Timeline/Network Flow consumers, shared selector metadata,
support observers, tests, authored catalogs, and generated topology together.
Rolling back only the interface or only hidden metadata consumers is invalid.
Rollback of the recovery host includes Coordination, Notes, provider/host wiring,
application containing bounds, accessibility tests, and any accepted goldens.
Rollback of draft retention includes runtime retirement and registry bindings;
do not retain mounted DOM nodes in the runtime. G1 and G7 are independent
consumer/test corrections and can be reviewed separately.

Maintainers should review owner text against typed contracts, source ownership,
keyboard behavior, and retained evidence; this handoff does not claim that a
human approval already occurred. Accessibility, visual, and measurement results
remain scoped implementation-readiness evidence, not release publication claims.

## Integration run ledger

- `make frontend-unit`, `20260914T164926Z-p77029`: 616/617 units passed;
  the remaining sentinel sent ArrowDown before Escape focus completed. Its
  corrected focused row passed in `20260914T165724Z-p61782`.
- `make browser-e2e-a11y`, `20260914T170531Z-p66310`: passed 16/16 units,
  including all selected Workbook/Network Flow accessibility groups.
- `make browser-e2e-stateful`, `20260914T170531Z-p66102`: failed 32/40.
  Network Flow pagination timed out; later groups rejected a temporarily unsorted
  authored title list during development. No missing group is counted as passing.
  The corrected fresh run `20260914T171150Z-p62729` passed 40/40 units.
- `make browser-e2e-webserver-backed`, `20260914T165859Z-p31811`: failed.
  Retained reports contain history refresh after leaving Timeline, contextual
  Assessment-to-Decision receipt, and an ambiguous account-menu query; a later
  Timeline case lost API connectivity. Finalization also rejected a startup log
  that was not owner-only. The account query now uses the button role. These
  failures require distinct dispositions; no overall pass is claimed.
- First `make check`, `20260914T171130Z-p34138`: failed 883/901 units.
  The new source store violated the common-runtime/Timeline import boundary,
  and support-handle object replacement caused a render loop. Scalar retention
  was moved to the source-neutral store and focus registration now tracks the
  underlying controller and root instead of changing facade objects. Import
  boundaries pass in `20260914T171600Z-p98105`; entry and sentinel units pass
  in `20260914T171709Z-p24720`; draft retention passes in
  `20260914T171859Z-p30192`. The final integrated run below passes after these corrections.
- `make frontend-typecheck` passed in `20260914T171602Z-p99204`;
  `make lint-biome` passed in `20260914T170431Z-p30543`;
  `make generate` passed in `20260914T171955Z-p71941`;
  `make generate-drift` passed in `20260914T171037Z-p25439`;
  `make lint-markdown` passed in `20260914T171036Z-p24236`.
- `make agent-finalize` passed in `20260914T171035Z-p23896`. `RESULTS_DIR`
  was unset, so retained successful-full-warm-run maintenance was skipped.
  Failed summaries and diagnostic artifacts were not removed.

### Visual review and refresh basis

The full ordinary visual run `20260914T165738Z-p87500` reconciled all 252 active
captures and all 252 committed goldens, with zero orphans, missing goldens,
ambiguous mappings, or unresolved registered fixtures. All functional scenarios
completed; six rows had screenshot differences. Coordination and Note recovery
placement is intentionally changed and reviewed at desktop and narrow widths.
Panels remain within the work area with their controls and status visible.

A separate clean HEAD checkout at
`/tmp/cartulary-frontend-remediation-baseline-cd0` remains at
`cd0be11dd321e48acbdcdbc507b6371099db48f9`. Its first visual attempt
`20260914T170616Z-p35536` failed service readiness before browser assertions.
The fresh comparison `20260914T171149Z-p60871` reproduced all 15 older
Decision supersession, Indicator lifecycle, and contextual Task/Decision
screenshot differences with identical differing-pixel counts. Evidence affordance
comparison `20260914T171740Z-p36521` reproduced the same 4,178 differing pixels.
Several actual PNGs are byte-identical between implementation and clean HEAD;
the remaining scenes have the same comparison differences. Their functional
assertions passed and visual review found unchanged text, controls, focus,
wrapping, and interaction state with small existing text-position differences.
These are reviewed stale renderer comparisons, not changes attributed to G5.

The accepted refresh triggers are intentional shared recovery layout and stale
comparisons against functionally verified current behavior. Viewports, masks,
scroll normalization, renderer pins, density, and screenshot scopes are unchanged.
The application containing-block correction additionally preserves supported
zoom bounds. Golden updates use only `make browser-e2e-visual-update`; no PNG
or golden digest is hand-edited. The exact promoted files, owner rows, fixture
mapping, and two fresh ordinary validation runs are recorded below.

Promotion passed in `20260914T172241Z-p23960` (12/12 units; 47 browser
scenarios). It changed the following 27 goldens and their generated manifest.
The existing renderer, viewports, masks, and normalization stayed fixed; seven
additional zoom views now reflect the application containing-block correction.
All changed files reconcile to active captures.

| Golden filename | Exact owner row | Stable fixture |
| --- | --- | --- |
| `account-menu-long-label-zoom-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | Active nonregistry capture |
| `contextual-decision-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-decision-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-decision-recovery-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-decision-recovery-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-decision-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-task-request-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-task-request-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-task-request-recovery-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-task-request-recovery-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `contextual-task-request-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | Active nonregistry capture |
| `coordination-recovery-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `decision-supersession-accepted-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | Active nonregistry capture |
| `decision-supersession-review-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | Active nonregistry capture |
| `decision-supersession-review-narrow-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | Active nonregistry capture |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` |
| `indicator-lifecycle-authoring-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `lifecycle-review-zoom-linux.png` | `web.design.visual.lifecycle` | Active nonregistry capture |
| `linked-note-recovery-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | Active nonregistry capture |
| `linked-note-recovery-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | Active nonregistry capture |
| `membership-audit-inspected-zoom-linux.png` | `web.design.visual.membership_audit_browsing` | Active nonregistry capture |
| `membership-management-removal-zoom-linux.png` | `web.design.visual.membership_management_visual` | Active nonregistry capture |
| `metadata-review-zoom-linux.png` | `web.design.visual.metadata_editing` | Active nonregistry capture |
| `workbook-query-empty-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-view-bar-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | Active nonregistry capture |


### Final regression and integration evidence

- The four original browser rows pass together in
  `20260914T173625Z-p26170`: `make service-backed-test-slice
  OWNER=module.workbook ROWS=<the four exact rows below>`, 17/17 units.
  This includes the complete keyboard, System-view switcher, save-status,
  and saved-view command-helper scenarios after the final focus API migration.
- `make frontend-unit`, `20260914T172459Z-p41999`: passed 617/617 units.
  The earlier `20260914T171957Z-p72564` passed 616/617; its remaining App
  landing assertion expected the old oversized viewport height. That assertion
  now verifies the containing-block height.
- Expanded Coordination accessibility plus the full save-status row passed
  13/13 units in `20260914T171858Z-p29850`. The three recovery states are
  retained editable draft, uncertain submission, and accepted refresh failure.
- Final adapter focus/creation regression slice passed 4/4 units in
  `20260914T173346Z-p79843`. The full adapter owner attempt
  `20260914T173042Z-p54102` passed all 39 frontend units, but its browser build
  rejected changing inputs during formatting; six downstream units could not
  run. It is not reported as an owner-suite pass.
- `make check`, `20260914T173302Z-p57127`: failed 900/901 units. Its only
  failure was TS7030 in the final operational-state focus effect. The effect
  now returns cleanup on every active path; `make frontend-typecheck` passed
  in `20260914T173345Z-p79574` and `20260914T173456Z-p1150`.
- `make agent-finalize`, `20260914T174227Z-p320`: passed. `RESULTS_DIR` was
  unset because no eligible successful full warm check had been selected;
  retained-run maintenance was skipped, with failed evidence preserved.
- Both ordinary post-promotion `make browser-e2e-visual` runs passed 12/12
  units and 47 scenarios: `20260914T173157Z-p85703` and
  `20260914T174048Z-p64482`. Every changed golden
  was visually inspected, including all seven zoom states. The zoom panels use
  their own scrolling, and the status strip stays within the viewport. Older
  contextual/Indicator/Decision comparisons preserve their controls and focus.

The original regression identities are unchanged:

| Owner | Exact row |
| --- | --- |
| `harness.browser` | `harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131` |
| `web.architecture` | `web.architecture.boundary_support.workbook_layout_policy_suite_5f8fee9a2f` |
| `web.architecture` | `web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039` |
| `module.workbook` | `module.workbook.browser.required_keyboard_shortcuts_operate_on_the_live_3ff4a5c6a9` |
| `module.workbook` | `module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956` |
| `module.workbook` | `module.workbook.browser_stateful.workbook_save_status_preserves_authoritative_tra_0fa8b0b16c` |
| `module.workbook` | `module.workbook.browser_support.verify_browser_command_helpers_for_sort_filter_g_cfa68b33e4` |

### Broader browser failures that remain open

The fresh complete webserver stage `20260914T172458Z-p41830` failed with
47 passed, 4 failed, and 58 skipped units. It is not a full-stage pass.
Canonical `make explain-run` reported the product failure; retained individual
unit results and Playwright reports provide the detail below. Artifact
finalization additionally rejected a non-owner-only startup database stderr
log. The harness security check remains enabled and the failed run is retained.

- Assessment contextual Decision creation observed zero receipts where one was
  required (`contextual-create.spec.ts:271`). The exact row
  `module.assessments.browser.contextual_decision_refresh` also fails in clean
  HEAD comparison `20260914T172303Z-p45567` under the baseline checkout.
  This proves that it predates this implementation; its cause and acceptability
  remain open. No product behavior or receipt assertion was weakened.
- The broad collaboration attempt still used the earlier ambiguous account-menu
  query. Its corrected button-role query in focused run
  `20260914T173233Z-p22655` passed the three-editor FIFO step, then exposed
  zero successful replay PATCHes after real HTTP 401 and same-user login.
  The retained trace ends on the incident directory with an access-loss notice,
  despite the incident being listed. No Timeline driver is mounted there.
  The cause of that access-loss transition is unproven; it is not classified as
  a stale assertion or claimed to reproduce on clean baseline. The exact row is
  `module.collaboration.browser.within_one_browser_runtime_queued_unsent_writes_3efbd48054`.
- Membership audit support attempted to read `envelope.data.memberships` from
  a response without `data` (`support/incidentMembershipAudit.ts:67`). The next
  two cases timed out at API readiness; the following database reset also
  failed, leaving later stage groups skipped. These files were not modified.
  No successful retry is substituted for the failed stage or missing evidence.
- The earlier History refresh failure did not repeat in the fresh broad stage;
  its exact row also passed 11/11 units in clean HEAD run
  `20260914T172302Z-p43775`. The earlier failure remains separate evidence.

These broader failures limit integrated browser readiness. They are not policy
exemptions, skipped assertions, or reasons to preserve any of the seven original
defects. The handoff keeps their ownership and evidence separate from the seven
remediation dispositions.


## Changed-source review map

- Owner text and guidance: Core 03, design direction, the frontend implementation
  testing guide, the developer guide focus API reference, this ledger, and the
  historical inspector handoff correction.
  Core 01, `docs/domain.md`, and `docs/testing-harness-nlspec.md` were inspected
  and retained under their existing authority boundaries.
- Shared contracts: `packages/grid-adapter/src/core.ts`, production/support grid
  bindings, `semanticFocusRequest.ts` and its tests, and UI Contracts grid/status
  selectors. Removed boolean command and hidden-version consumers were migrated
  across the workspace in the same change.
- Workbook focus and continuity: `useWorkbookSemanticGridFocus`, startup/entry
  generation ownership, continuity port, Generic/Entity/Timeline adapters, and
  Network Flow modal restoration. Corresponding unit/support tests await actual
  completion and exercise cancellation and authority boundaries.
- Layout and recovery: `WorkbookWorkAreaOverlay`, shell/provider/work-area
  wiring, Coordination/Note triggers, overflow presentation/focus return,
  `AppRoot`/`App` containing bounds, and layout implementation guidance.
- Retained editing: `WorkbookLocalDraftStore`, `WorkbookMutationRuntime`,
  Timeline draft registry/foundation, and lifetime/retirement regressions.
- Browser support: explicit replay outcomes and detachment; mounted-row version
  consumers; measurement descriptors/observer; ordinary reference queries;
  saved-view reading, omission and persistence assertions; keyboard restoration;
  and expanded Coordination accessibility.
- Authored verification projections: `tools/test_families/harness.browser.json`,
  `module.timeline.json`, `package.grid_adapter.json`, `web.workbook.json`, and
  `tools/frontend_source_ownership.json`. Make regenerated
  `tools/execution_topology_render_index.json`.
- Visual inputs: the 27 reconciled PNGs listed above and the Make-produced
  `tools/frontend_visual_golden_manifest.json`. No renderer pin, mask, fixture
  viewport, threshold, or sample policy changed.

No dependency or lockfile changed. No database or stored-data migration is
required. Do not land an intermediate version that mixes the old focus interface
or hidden row-version probes with their migrated consumers.


### Final projection checks

- `make generate-drift`: passed 4/4, `20260914T174651Z-p10426`.
- `make generated-artifact-policy-check`: passed 3/3,
  `20260914T174651Z-p10442`.
- `make json-shape-check`: passed 3/3, `20260914T174651Z-p10475`.
- `make test-catalog-check`: exited successfully on 2026-09-14 after promotion.
- `make service-backed-test-slice OWNER=module.workbook` selecting the six
  `ordinary_create_*` rows: passed 13/13, `20260914T174526Z-p52363`.
- `git diff --check`: passed. Temporary causal console instrumentation,
  obsolete boolean focus API calls, the old optional commit-wait flag, and hidden
  Timeline row-version consumers were removed from the changed boundaries.


### Integrated check disposition

`make check` passed all **901/901 units** in
`20260914T174312Z-p4641`, after `make agent-finalize`. This includes the final
frontend typecheck, unit, import-boundary, Biome, catalog/architecture checks,
and selected backend checks. It is fresh integrated evidence; the two previous
failed checks remain recorded above. `make lint-markdown` separately passed in
`20260914T174638Z-p2384` and, after the final handoff and developer-guide
updates, `20260914T175738Z-p52928`. The complete browser measurement stage
ran after the other heavy jobs finished and retained the existing qualified
resource profile and samples.


### Fresh measurement evidence

`make browser-e2e-measurement` passed **25/25 units**, all eight browser rows,
in `20260914T175220Z-p5285`. It ran after the final check and other browser jobs
finished, under the existing `browser_measurement_quiet` resource profile.
Each Timeline predicate used one warm-up and 100 measured samples with no product
retry: 20,000 Timeline rows, 25 analyst sessions (24 background sessions),
4.8 background updates per second, presence enabled, and target rows excluded
from background mutation. All four Network Flow measurement rows also passed.

| Predicate | Observed p95 | Existing limit | Disposition |
| --- | --- | --- | --- |
| `perf.timeline_summary_selection_down.v1` | 30.5 ms | 100 ms | Passed |
| `perf.timeline_summary_focus_edit.v1` | 33.9 ms | 100 ms | Passed |
| `perf.typing_ack.v1` | 31.5 ms | 100 ms | Passed |
| `perf.timeline_blank_row_create.v1` | 85.9 ms | 150 ms | Passed |

Values are rounded to one decimal place from retained
`cartulary.frontend_measurement_observation.v2` attachments in each group's
`playwright-report.json`. The changed observer executes in the real browser;
creation qualification uses the same mounted row's committed version, identity,
visible field, and consecutive paint frames. Thresholds, fixture sizes, sampling,
and source-owned mutation admission were not relaxed.

## Final disposition and handoff

G1–G7 are closed at their specified boundaries with fresh policy, unit, browser,
accessibility, visual, and measurement evidence. The implementation preserves
Core 01 sparse PATCH and Core 03 authoritative editor acceptance while correcting
creation eligibility, completed focus, shared overlay geometry, retained refused
drafts, and observer ownership. No compatibility alias, policy exemption,
disabled assertion, database migration, or new reload-persistence guarantee was
introduced.

Broader webserver readiness remains limited by the failures enumerated above.
No full-stage pass or release-publication claim is made. Their failed reports,
the clean comparison checkout, and intermediate failures remain available.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset at
finalization; generation, formatting, drift, source policies, and verification
were completed through public Make targets. No commit or deployment was made.
Review and land the coupled source, owner text, authored routing, generated
projections, and visual evidence together using the rollback boundaries above.
