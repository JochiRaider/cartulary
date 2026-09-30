# Human-readable Timeline bulk-selection labels

Date: 2026-09-30. Status: implementation, focused regression verification and
scoped rendered review pass. Actual assistive-technology usability was not
tested. No commit, push or deployment was performed.

## Baseline, authority and selection decision

The checkout started clean on `main` at
`d11d02158582ee7ea482f91e1fba785da67abcc9`, two commits ahead of `origin/main`.
HEAD and branch remain unchanged. The dismissed-mention readability, mention
focus reveal, Columns focus continuity and built-in surface keyboard work in the
existing stack are preserved. The root AGENTS.md, current Make help/help-all,
owner task guides, local source guides, source/import manifests and generated
artifact policy were inspected. No owner contradiction was found.

Core 00 owns precedence. Core 03 REQ-03-297 owns Timeline's opt-in record
selection, independent active-cell/Inspector/bulk state, loaded-window scope,
membership, pruning and captured-target continuity. Section 13.3,
REQ-03-221–222 owns explicit multi-row tag mutation batches. REQ-03-298,
REQ-03-219 and REQ-03-283 govern editing and semantic continuity;
REQ-03-299/100 governs scoped authority and retained operation lifetimes.
Design §6.2 places technical identifiers in secondary metadata; §8.1 preserves
stable behavioral identity, §8 supplies selection direction, and §14 supplies
the accessibility contract. Domain supplies vocabulary and owner navigation.
These maintained owners are unchanged.

The user's approved plan authorizes this bounded presentation correction.
Research, the prior bulk-tag handoff, the digest and bundled prompts were source
material, without becoming authorization or executable requirements. Both
`cartulary-ui-ux-refactor` and `cartulary-ui-review` were applied. The digest
README, START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP, QUERY_RECIPES,
rules and acceptance were read. No separate controlling tracker exists for this
slice; this handoff records the completed workstream exits.

The selection-rubric finding was a confirmed naming defect: the production and
test renderers identified an ordinary record solely by its UUID even when a
readable committed synopsis existed. The correction belongs in presentation,
tests, authored verification routing and source guides. Timeline's existing
row-model layer owns the record context; Grid Adapter owns consuming a neutral
selection presentation. The small required accessor hides the shared decision
of how to label a selection control without importing Timeline field knowledge
into the adapter. Another authorized consumer could supply its own presentation
through the same port; no other surface is enabled now.

This boundary keeps record identity and selection semantics independent of
display text, avoids duplicate naming implementations and removes copy-dependent
test targeting. Leaving the gap would continue making record choice depend on
technical IDs. The existing selection policy and bulk mutation planner remain
valuable owners and are retained. No title framework, cache, selection store,
request, persistence, endpoint, authorization or normative-specification change
was introduced.

## Behavior and compatibility

Before: `Select record <UUID>`.

After: `Select Timeline record: Initial triage — 2026-04-10T10:00:00Z`,
with accessible description `Record ID: <UUID>`.

The producer reads authorized `committedValues`, never local values or retained
editor text. Accepted updates can refresh the presentation; unsubmitted,
pending and rejected drafts do not become its authority. Whitespace is collapsed
and trimmed only in presentation strings. Stored and draft values are untouched.
UTC text takes precedence; otherwise local text has the suffix `(local time)`.
Time text is preserved without conversion.

| Committed content | Name context after `Select Timeline record: ` |
| --- | --- |
| Synopsis and UTC | `Initial triage — 2026-04-10T10:00:00Z` |
| Synopsis and local time only | `Local activity — 2026-04-10 08:00 (local time)` |
| Time only | `No synopsis — 2026-04-10T10:00:00Z` |
| Synopsis only | `Synopsis only` |
| Absent or blank synopsis and time | `No synopsis or activity time` |
| Multiline synopsis | `First line second line — 2026-04-10T10:00:00Z` |
| Synopsis longer than 120 Unicode code points | First 119 code points plus `…`, followed by available time |

Every description includes the stable record ID. For a shortened synopsis, the
full normalized synopsis/time context precedes that secondary ID. Duplicate
committed contexts retain identical readable names and distinct descriptions;
neither row position nor current-window duplicate counts influences naming.
The synopsis bound is code points, as approved, rather than UTF-16 units or a
grapheme bound. No extra live announcement accompanies label refresh.

`GridCoreRecordBulkSelection<Row>` now requires
`getRecordSelectionPresentation(row)` returning `{ label, description? }`.
All current callers and both bindings were migrated in one change. This is an
internal compile-time interface change with no legacy UUID naming fallback.
Timeline remains the sole production consumer. There is no data migration.

Native checked/mixed state, click/Space activation, focus and stable test IDs are
retained. Selection mechanics, row versions, mutation identity, complete-set
admission and exact uncertain replay have no production edits. Selection still
excludes group/draft rows and ineligible creation pins, prunes evicted or
unauthorized members, and does not expand on page append. Active cell, range and
Inspector state remain independent. Raw tag text/caret, explicit submission,
recovery, read-only and closed-incident behavior retain their existing owners.

## Changed and inspected paths

Paths are relative to `/home/jochi/code/cartulary`.

- `packages/grid-adapter/src/core.ts`: required neutral presentation accessor.
- `packages/grid-adapter/src/SemanticDataGrid.tsx`: forwards the accessor through
  the production binding.
- `packages/grid-adapter/src/rdgCompiler.tsx`: eligible selection checkboxes
  consume `aria-label` and optional `aria-description` from their current row.
- `packages/grid-adapter/src/test-support.tsx`: equivalent test binding.
- `apps/web/src/workbook/timeline/models/timelineRowsModel.ts`: pure committed
  presentation helper beside stable row projection.
- `apps/web/src/workbook/timeline/bulk/useTimelineBulkTagController.ts`: supplies
  that helper through the existing selection capability.
- `packages/grid-adapter/src/index.test.tsx`: both-binding presentation,
  identity/version update, reorder, same DOM node, focus, checked/mixed state and
  explicit deselection coverage; existing callers/locators migrated.
- `apps/web/src/workbook/timeline/timelineBulkTagInteraction.test.tsx`:
  normalization, missing/local/long/duplicate content, source preservation,
  retained drafts, accepted committed refresh, selected identity and zero-send
  coverage alongside existing controller guards.
- `apps/web/src/workbook/WorkbookShell.sentinel.test.tsx`: stable row-scoped
  checkbox lookup while retaining its semantic bulk-target assertions.
- `apps/web/e2e/timeline-bulk-tag.spec.ts`: independent accessible-name and
  description assertions, accessibility-tree attachments, both viewports,
  keyboard selection, zero incidental writes, Inspector coexistence, accepted
  label/version update without replacing the focused control, actual sort order,
  refresh and explicit assignment to the intended ID/version. Existing pending
  save and rejected edit scenarios now assert committed names too.
- `apps/web/e2e/timeline-grid-entry.spec.ts`,
  `timeline-range-selection.spec.ts`, `timeline-deferred-continuity.spec.ts`,
  `batch-history-review.spec.ts` and `timeline-find.spec.ts`: migrated affected
  locators, including negative creation-pin assertions, while preserving
  unrelated semantic assertions.
- `apps/web/e2e/timeline-auto-resolution-feedback.spec.ts`: migrated locator and
  an accessibility-tree assertion that protected selection names disappear after
  authority removes the controls.
- `apps/web/e2e/workbook-column-sizing.spec.ts`: stable row/checkbox targeting
  replaces the accessible-copy CSS lookup in frozen-gutter paint coverage.
- `tools/test_families/module.timeline.json` and
  `tools/test_families/package.grid_adapter.json`: authored scenario/title routing.
- `tools/browser_e2e_batch_manifest.json` and
  `tools/execution_topology_render_index.json`: generated projections refreshed
  through Make; the new scenario joins the existing bulk-tag batch.
- `packages/grid-adapter/README.md`, Timeline `models/README.md` and
  `bulk/README.md`: narrow source-guide updates.
- This handoff.

Inspected and unchanged: `semanticSelectionPolicy.ts`,
`TimelineWorkbookGrid.tsx`, `timelineRowModel.ts`, `timelineBulkTagPlan.ts`,
`gridSelectors.ts`, the shell accessibility scenario, source/import ownership
inputs, governing documents and historical bulk-tag handoff. Direct vendor-grid
imports remain inside Grid Adapter.

Retired together: UUID-only naming in both renderer bindings and every current
`Select record` locator in the affected TS/TSX sources. Tests now locate the
stable row, then its checkbox, and assert presentation independently. No
compatibility alias or parallel selection store remains. Executable consumers
do not read, stat or hash Markdown; prose changes are human navigation only.

## Regression and verification evidence

All commands ran from the repository root through public Make targets.
Counts below are harness graph units, not counts of individual assertions.
`make help`, `make help-all` and `make task-guide ROLE=module-author` with
`OWNER=package.grid_adapter`, `module.timeline` and `module.workbook` succeeded.
Behavioral evidence used `CARTULARY_HARNESS_CACHE_MODE=off`.

The new scenario is `scenario_timeline_bulk_selection_names`, routed as
`module.timeline.browser.timeline_bulk_selection_exposes_committed_readab_0793a3e62b`.
Before production edits, its focused service-backed slice failed at
`.cartulary/test-results/20260930T160016Z-p16119` (9/11 graph units): a
stable-row-located checkbox expected the readable name but received
`Select record <UUID>`. The accessibility-tree attachment records the original
gap. This expected red establishes that selection/count passes alone did not
test meaningful naming.

Final unit slices used `make test-slice OWNER=<owner> ROWS=<row>`:

| Owner and row | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `package.grid_adapter` / `package.grid_adapter.regression.index_suite_d804ef9789` | PASS 2/2 | `20260930T163158Z-p44304` |
| `module.timeline` / `module.timeline.frontend.timeline_bulk_tag_controller_8a15df23b4` | PASS 2/2 | `20260930T161132Z-p74363` |
| `module.workbook` / `module.workbook.frontend.sorted_and_filtered_paste_translation_uses_stabl_d4950d24ba` | PASS 2/2 | `20260930T161132Z-p74374` |

Browser slices used `make service-backed-test-slice OWNER=<owner> ROWS=<list>`.
The exact row sets were:

1. `OWNER=module.timeline`: the new naming row plus
   `module.timeline.browser.bulk_tag_characterization`, `bulk_tag_delivery_entry`,
   `bulk_tag_layouts`, `bulk_tag_membership`, `bulk_tag_prerequisites`,
   `bulk_tag_rejection` and `bulk_tag_replay` (the shortened names share the
   `module.timeline.browser.` prefix). PASS 11/11 at
   `20260930T161132Z-p74399`. After strengthening actual reorder evidence, the
   naming row was rerun fresh: PASS 11/11 at `20260930T164200Z-p62222`.
2. `OWNER=module.timeline`: `module.timeline.browser.row_action_menu`,
   `spreadsheet_batch_bulk_conflicts`, `spreadsheet_pointer`, `range_presentation`,
   `clear_rectangle`, `deferred_continuity`, `batch_history_review_tag` and
   `auto_resolution_feedback_characterization` (same prefix). PASS 16/16 at
   `20260930T161244Z-p9135`.
3. `OWNER=module.workbook`: `module.workbook.browser.find_literal`,
   `module.workbook.browser.find_query_creation`,
   `module.workbook.browser.frozen_region_paint_01360d5734` and
   `module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159`.
   PASS 14/14 at `20260930T161245Z-p9362`.

These are selected authored rows, expanded by the repository's batch routing;
no unrestricted browser suite was requested or inferred. They cover selection
membership through filtering/grouping/virtualization/page append, Inspector and
range independence, read-only/closed/access-loss, prerequisites and rejected
drafts, retained tag/caret, captured targets, conflicts and exact uncertain replay.

| Public command | Final result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make format` | PASS 2/2; diff inspected | `20260930T164247Z-p95124` |
| `make agent-finalize` | PASS 1/1 | `20260930T164336Z-p811` |
| `make frontend-typecheck` | PASS 2/2 | `20260930T164424Z-p5531` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260930T161642Z-p24895` |
| `make lint-biome` | PASS 2/2 | `20260930T164424Z-p5561` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260930T161642Z-p24670` |
| `make json-shape-check` | PASS 3/3 | `20260930T161648Z-p26599` |
| `make generate-drift` | PASS 4/4 | `20260930T161654Z-p27129` |
| `make lint-markdown` | PASS | `20260930T165443Z-p12580` |

Authored routing preceded generation through
`CARTULARY_GENERATE_DRIFT_REFRESH=1 make generate-drift`, PASS 4/4 at
`20260930T155947Z-p11867` and `20260930T160844Z-p60590`.
`RESULTS_DIR` was unset: retained-run maintenance was skipped with
`results-dir-not-provided`, because no eligible successful full warm check exists.
Markdown lint's summary is `adhoc/lint-markdown/tool-run-summary.json` in its
recorded run root. `git diff --check` passes; final source and documentation
scope were reviewed.

Intermediate failures were repaired within scope:

- Initial routing preflight rejected unsorted row IDs; authored rows were sorted.
  `make agent-finalize` at `20260930T155845Z-p10713` and
  `make json-shape-check` at `20260930T155924Z-p11316` exposed the resulting stale
  projection. Make regeneration resolved it; final drift/policy/shape checks pass.
- `make frontend-typecheck` at `20260930T161642Z-p24869` failed solely on two
  unsupported Testing Library `exact` properties in the new adapter assertions.
  Those test-only properties were removed, then the adapter slice, formatter,
  finalizer, typecheck and Biome were rerun successfully.
- Manual review revealed that the first naming-test header clicks did not sort
  the UTC text field. The scenario now uses the owner sort menu's Activity Sort
  Time and asserts that both stable IDs are present in the resulting order. The
  fresh successful rerun above supersedes that narrower initial sort evidence.

## Rendered review and limitations

`make ui-review UI_MODE=seeded REVIEW_PROFILE=default` used isolated synthetic
data, the editor actor and a sealed production frontend. Run root:
`.cartulary/test-results/20260930T161408Z-p90950`; session:
`uireview-0a581045712b7afa52f2d39a4fce03d5`.

Seven captures were consumed before shutdown. Original image and observation
byte lengths/SHA-256 were verified against each returned bundle before viewing.
The production accessibility tree exposes synopsis/time names at 1440×900 and
1024×720. Space selected the intended record with native checked state and
`1 record selected.` feedback. A distinct saved row's UTC cell was activated,
its editor canceled without authoring, and the Inspector opened for that other
row while the first stayed checked. Captures 22 and 24 show this coexistence at
both viewports. Existing geometry and chrome remain unchanged by naming.

Activity Sort Time was then added and changed to descending through its menu;
the accepted order ran from later to earlier records. Refresh retained the count.
Explicit scrolling revealed the original selected record at the end of the
66-record window, still checked with the same readable context (capture 36).
Automated production-browser coverage additionally proves the accepted synopsis
update preserves the same focused DOM control and targets the accepted version.

Early manual targeting did not establish Inspector coexistence: clicking a row
hit its checkbox, and opening an empty Inspector supplied no subject. Focus
attempts in that empty-Inspector state returned `target_unavailable` in
operations 13/15; fresh observations were used, the empty Inspector was closed,
and explicit checkbox focus/Space then succeeded. An empty-name grid scroll
request was rejected as `invalid_request`; a fresh observed grid element
reference supplied the successful scroll. These attempts do not count as
positive focus/Inspector evidence. The completed states above supersede them;
no focus or harness implementation change was made.

Axe completed on captures. It reported `empty-table-header` (minor) and
`page-has-heading-one` (moderate), with incomplete `color-contrast` analysis;
the narrow Inspector capture had only the missing-heading violation. The terminal
aggregates 11 violations and 7 incomplete results across seven scans, not unique
defects. These are outside this naming correction and no CSS/header/heading
change is present. No whole-page Axe or conformance pass is claimed. There were
zero recorded console errors and failed requests.

The exact `make ui-review-stop UI_SESSION=<this run's ui-review/session.json>`
succeeded. `ui-review/terminal.json` records `closed`, `status=ok`, exit 0 and
`cleanup=complete`; the foreground process exited successfully. The 37
caller-owned request files and their directory were removed. Private screenshots,
raw observations and runtime paths are deliberately absent from this durable
handoff and their links expired at cleanup.

No actual screen reader or participant study was available/performed. Browser
names/descriptions and tree assertions establish the rendered semantic result,
not assistive-technology usability. No visual golden was refreshed. Broad visual,
measurement, backend, release and unrestricted browser suites were skipped:
attributes and the neutral presentation port changed, with no layout, token,
endpoint, data or persistence change requiring those suites. The focused frozen
paint row and inspected production captures provide bounded geometry evidence.
Historical visual differences were not treated as current results.

## Digest acceptance assessment

These are scope-specific human assessments, not a completeness or publication
claim. No acceptance TSV or governing owner text was changed.

| ID | Outcome | Evidence or scope rationale |
| --- | --- | --- |
| A001 | PASS | Core/design mapping above; source ownership and verification routing remain distinct. |
| A002 | PASS | Confirmed defect, small producer/adapter seam, benefit, extension, retirement and rollback recorded. |
| A003 | PASS | Clean baseline, existing stack, source guides/manifests and final diff inspected; HEAD preserved. |
| A004 | PASS | Diff adds no design literals, token registry or styling. |
| A005 | PASS | Existing dark_graphite review; no theme change. |
| A006 | PASS | Existing bulk-layout density scenarios and frozen geometry pass; no density implementation changed. |
| A007 | PASS | Membership and Find creation-pin exclusions pass; draft/group controls remain excluded. |
| A008 | PASS | Both requested viewports and existing bulk-layout/Inspector scenarios pass; responsive owners unchanged. |
| A009 | PASS | Shell accessibility and rendered shell browsing/Inspector reachability pass. |
| A010 | PASS | Distinct Inspector subject retained alongside selection; routed row-action/range scenarios pass. |
| A011 | PASS | Same control/focus through accepted label update; reorder/refresh, deferred and range continuity pass. |
| A012 | PASS | Existing delivery/captured-target/exact replay rows pass; mutation planner unchanged. |
| A013 | PASS | Rejection/replay recovery rows pass; existing operation and acknowledgement owners retained. |
| A014 | PASS | Committed/draft separation, raw tag/caret and prerequisite/rejection scenarios pass. |
| A015 | PASS | Existing batch-conflict and rejected-edit coverage pass; names use committed values. |
| A016 | PASS | Existing membership, read-only/closed and complete-set admission rows pass; no permission inferred by names. |
| A017 | PASS | Authority-loss tree excludes protected names; existing scoped revocation/retention scenarios pass. |
| A018 | N/A | Evidence lifecycle/overlay/preview behavior is untouched by a selection-name projection. |
| A019 | PASS | Scoped name/description, native keyboard/state, focus and owner shell accessibility evidence pass; Axe/AT limits stated. |
| A020 | PASS | Blank/duplicate/multiline/long Unicode names and existing bulk layout/authoring controls pass. |
| A021 | PASS | Membership virtualization guards and manual selected-record reveal after reorder/refresh pass. |
| A022 | PASS | Production captures inspected at both requested viewports; frozen paint row passes; no golden change needed. |
| A023 | PASS | Stable view/record row selectors replace copy-dependent lookup; names are asserted independently. |
| A024 | PASS | Executable diff/dependency review finds no Markdown consumption; documentation maintenance stays separate. |
| A025 | PASS | Authored family changes precede Make-generated batch/render-index projections; policy/shape/drift pass. |
| A026 | PASS | All internal consumers/bindings migrated; no data migration, compatibility fallback or invented capability. |
| A027 | PASS | Completed evidence, failure dispositions, limits, skipped checks and rollback recorded in this handoff. |

## Delivery and rollback

The baseline/authority, expected-red characterization, implementation/caller
migration, routing generation, focused verification and rendered review exits are
complete in dependency order. The final diff is limited to the presentation port,
Timeline projection, corresponding tests/routing and source documentation. No
selection policy, captured mutation plan, normative specification, dependency,
CSS or golden changes are included.

Rollback reverts this task's adapter interface and both bindings, Timeline helper
and capability, tests, authored routing/generated projections and source guides
as one coherent change. It does not revert accepted record data or the preceding
UI fixes. The next action is normal code review; there is no required implementation
or focused-verification blocker. A real assistive-technology session is useful
additional evidence, with the stated current limitation preserved.
