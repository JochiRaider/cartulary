# Workbook query-recovery footer cleanup

## Scope and owners

This bounded interaction correction started from clean `main` at
`0815500f55886c5a64288cb84b68f319328fedd0`. Core 03 §14.9 owns bounded
browsing, accepted-result retention, Retry/Revert, and draft-safe departure;
REQ-03-224 owns latest applicable query acceptance. Design §§8.3 and 14 direct
compact loading, local recovery, keyboard focus, and accessible feedback. The
research document, domain vocabulary, UI/UX digest, and earlier query and Group
handoffs were used for navigation and advice, not as added requirements.

The source owner remains `WorkbookQueryBrowser` for destinations, request
lifetime, cursor replay, and bounded recovery. Its existing registry/controller
and `WorkbookQueryBrowsingControls` own the shared footer presentation on
Timeline and Notes. Current source, source guides, authored test catalogs, and
both `make task-guide ROLE=module-author OWNER=web.workbook` and
`OWNER=module.workbook` routes were checked. Verification is independently
routed through `web.workbook` unit rows and `module.workbook` browser rows. The
digest's localization snapshot and historical handoffs did not substitute for
current inspection or execution. The narrow offline digest focus query led to
adopting its stable-focus, local-feedback, and existing-owner advice; no theme,
framework, icon, or additional recovery store was adopted.

## Characterization and decision

Source confirmed that starting a read cleared `failure`, removing Retry and
Revert, while Retry compared its action to a pending **destination** that could
be continuation or replacement. The synchronous duplicate-activation guard
already existed. After the implementation, the new browser row was run against
the two original production files in a controlled baseline swap, then those
files were restored. The expected red run
`.cartulary/test-results/20260924T182008Z-p68981` failed at held Retry on
both Timeline and accepted-empty Notes: the focused Retry locator no longer
existed. The separate baseline component run
`.cartulary/test-results/20260924T182139Z-p2816` showed Revert focus on the
document body after activation. These are confirmed before-state defects; no
pre-edit physical-browser focus destination for Revert is claimed. The earlier
baseline slice could not start its units because of `service_start_error` at
`.cartulary/test-results/20260924T174923Z-p82082`; using the repo-local Node
runtime in `PATH` resolved that infrastructure issue.

The common decision is the distinction between the user's initiating footer
action and the read destination. `WorkbookQueryBrowser` now publishes an
internal `pendingAction` consumed at read start and cleared only with that
request's settlement or lifetime retirement. A bounded automatic read retry
keeps the originating action; a superseding read cannot inherit it. The footer
uses the action for Retry busy state and concise visible progress, while the
destination still controls request semantics. A restart destination shows
refresh progress for non-Retry actions. This is a behavior correction within
the existing owner, with no new query engine or public snapshot contract.

## Behavior and compatibility

- A keyboard-activated Retry remains connected and visibly focused through a
  held read. It is `aria-disabled` while pending, so repeated Enter cannot
  dispatch another read. Repeated failure makes the same control usable again.
  Success keeps the focused control unavailable until focus leaves, then
  removes it without calling `focus()` or sending focus to the grid.
- Revert restores the accepted query and keeps its focused control unavailable
  until the user leaves it. A read settling after focus moves does not reclaim
  focus. Focus retention is tied to the browser instance, surface, and lifetime
  epoch; detachment and authority retirement invalidate it.
- Pending Retry reads say “Retrying records…” for replacement, refresh, and
  continuation destinations. Stale failure copy is absent during loading. The
  footer's status is silent during loading/failure so the existing grid state
  plane owns those announcements; settled accepted counts use the footer's
  polite status. Accepted-empty results have the same recovery behavior.
- Existing accepted rows, requested-versus-accepted chips/grouping, scroll,
  retained authoring, cursor replay, canonicalization, bounded automatic
  recovery, draft-safe browsing, generation fences, and authorization rules
  remain with their owners. The new Timeline browser sequence observes no
  recovery-triggered record write. The adjacent history row verifies retained
  draft and scroll semantics; the Group row verifies accepted chips/grouping.

The removed mechanisms are failure-only recovery rendering and destination-based
busy comparison. The existing activation guard, footer focus styling, query
owner, and grid operational-state announcements remain. No HTTP route, schema,
stored data, migration, direct Grid Adapter import, or record write changed.
Rollback restores the two production files, focused tests, authored catalog
rows, generated catalog derivatives, and this handoff together. Analyst data
needs no migration. No latency improvement was measured or claimed.

## Verification

Paths and run roots below are relative to the repository root. The new
`module.workbook.browser_stateful.query_recovery_focus` row uses real keyboard
activation and held production responses. It asserts focus without locator
`focus()` or click repair between recovery steps, a computed visible focus ring,
duplicate suppression,
replacement request equality, continuation cursor/request equality, repeated
failure, success, focus departure, Revert, accepted-empty Notes, and no record
write. Component and browser-owner tests cover supersession, surface change,
authority invalidation, and action/destination separation. Authored catalog
inputs were changed before `make generate`; generated files were not hand-edited.

| Check | Result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260924T181105Z-p48359` |
| `make format` | PASS | `.cartulary/test-results/20260924T183323Z-p69235` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.query_browsing,web.workbook.regression.query_browsing_controls` | PASS 3/3 | `.cartulary/test-results/20260924T181641Z-p2557` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.query_recovery_focus` | PASS 11/11 | `.cartulary/test-results/20260924T183337Z-p73674` |
| Existing continuation, Group, authority, and Timeline rows | PASS 11/11 | `.cartulary/test-results/20260924T181134Z-p52024` |
| Existing draft/scroll history row | PASS 11/11 | `.cartulary/test-results/20260924T181758Z-p36004` |
| `make agent-finalize` | PASS 1/1 | `.cartulary/test-results/20260924T183447Z-p6883` |
| `make frontend-typecheck`, `make lint-biome`, `make frontend-import-boundary-check` | PASS 2/2 each | `.cartulary/test-results/20260924T183508Z-p10723`, `.cartulary/test-results/20260924T183527Z-p11256`, `.cartulary/test-results/20260924T182253Z-p8991` |
| `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | PASS 4/4, 3/3, 3/3 | `.cartulary/test-results/20260924T183153Z-p62810`, `.cartulary/test-results/20260924T182320Z-p13273`, `.cartulary/test-results/20260924T182324Z-p13703` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260924T182542Z-p15217` |

`RESULTS_DIR` was unset for `make agent-finalize`: no eligible successful full
warm check was supplied, so retained-run maintenance was skipped. Two
intermediate `make generate` runs rejected unsorted authored catalog entries
(`20260924T180130Z-p99502`, `20260924T181041Z-p45101`); sorting the entries
resolved them. An intermediate `make format` run caught an effect dependency
error (`20260924T181444Z-p86121`), and a unit slice caught an ambiguous test
status selector (`20260924T181602Z-p96540`); both were corrected. A concurrent
import-boundary check had `service_start_error`
(`20260924T182226Z-p7550`) and passed when rerun sequentially with the
repo-local Node runtime. These failures were either test/authoring corrections
or infrastructure, not remaining product failures. Whole browser, visual,
accessibility, and performance suites were not selected because the owner-routed
focus and adjacent regression rows covered this boundary. The browser evidence
is Chromium only; the draft-retention history row does not combine an open
editor with a Retry in the same scenario.

## Digest acceptance assessment

The digest is advisory and these dispositions apply only to this slice.

| Rows | Disposition and evidence |
| --- | --- |
| A001–A003 | PASS: adopted clauses, source/verification split, clean baseline, current source/catalog review, focused boundary, and explicit retirement/retention above. No structural abstraction was introduced. |
| A004–A005 | PASS: existing footer tokens and dark graphite styling; no new design literal, theme, or palette. |
| A011, A014 | PASS: focus/row continuity in the new row and retained raw draft/scroll in the adjacent history row; no write in recovery. |
| A016–A017 | PASS: distinct read destination/action state, accepted-empty and populated recovery, supersession and authority lifetime tests, and the existing authority browser row. |
| A019–A021 | PASS for applicable footer states: keyboard activation/focus, concise announcement priority, focused unavailable control, stable loaded-window behavior, and the production plus history rows. No whole-workbook accessibility or performance claim follows. |
| A023–A026 | PASS: semantic selectors, no executable Markdown dependency, authored routing followed by generated projections and drift/policy checks, and unchanged public compatibility with no migration. |
| A027 | PASS: this handoff records behavior, owners, results, failures, limits, retirement, and rollback. |
| A006–A010, A012–A013, A015, A018, A022 | N/A: density, creation, responsive layout, inspector, transactions/queue replay, conflict, Evidence, and visual-golden ownership were not changed. Query duplicate activation is covered under A019, not treated as a transaction claim. |

The slice exit is complete. A later visual or cross-browser audit can reuse the
new owner-routed row; it is not required to operate this correction.
