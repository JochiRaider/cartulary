# Workbook Sort focus continuity cleanup

## Boundary and authority

- Baseline: clean `main` at `b589252a0`. The selected workflow is the Timeline
  Sort editor, including return to the grid. The September 13 UI/UX digest is
  advisory navigation; current source includes the later Filters form-mode fix.
- Behavior: `docs/design.md` §§8.3–8.5 and 14.1 own the complete ordered Sort
  actions, menu keys, chip entry, Escape, outside dismissal and keyboard parity.
  Core 03 §14.1 (REQ-03-223/224) owns schema sort capability, canonical query
  state and latest applicable results. Core 03 §14.9 keeps requested overrides,
  accepted rows and canonical metadata distinct during pending/failing queries.
  The adopted Testing Harness NLSpec and authored catalog own verification
  routing, not product behavior.
- Source: workbook presentation owns Sort commands and the requested editor
  projection; `web.shared` owns registered menu navigation. Verification routes
  independently through `web.workbook`, `web.application`, `module.workbook`,
  `module.savedviews` and `platform.viewquery`. The repository source/import
  boundaries keep direct grid-vendor integration in Grid Adapter; this slice
  adds none. `docs/domain.md`, the prior view-bar and Filters handoffs, and
  digest maps were used for navigation and rechecked against current source.
- Advisory classification: adopt the digest's bounded keyboard parity and
  semantic focus concerns within the owners above. Its dated snapshot, bundled
  upstream examples and historical green results do not define requirements.

## Before and after

Before, Sort left registered-item reconciliation disabled. Add replaced its
focused control, Remove deleted its controls, and a completed move could disable
the focused earlier/later action. Pointer focus did not update the active menu
item. The pre-change component row confirmed Add left focus on `body` and pointer
focus left the chosen Remove item at `tabIndex=-1`; the pre-change production
Chromium row confirmed Add left its replacement direction control unfocused.
Removal, disabled-move and late-response focus loss were hypotheses until the
new post-change transition tests exercised them.

After each Sort edit, the open editor presents the requested sort list and
commands use that same list, including the eight-sort limit. Accepted chips,
trigger count and grid rows continue to describe the accepted query. A visible
status and `Requested sorts` legend identify an unapplied edit. Query submission,
server canonicalization, Retry/Revert, saved views, authorization and grid drafts
keep their existing owners.

Sort opts into the shared registered-menu reconciliation. It preserves a
surviving field/action key. When a key disappears, a newly added field or
disabled move goes to that field's direction; a removed row goes to the next
surviving old-order row's direction, then the previous row; the final row goes
to its Add control. A currently eligible Add or the helper's eligible fallback
handles the remaining edge; if no registered control remains, the menu closes
to its trigger. The helper validates the semantic choice against
currently connected and enabled registered items. Focus capture aligns pointer
and keyboard navigation with the actual focused item. Arrow/Home/End skip
disabled controls, and the menu keeps one roving Tab stop.

Reconciliation runs only while the Sort menu owns focus. Escape uses the
surviving invoking chip, then the Sort trigger; outside dismissal and surface
departure leave the user's newer destination alone. A delayed accepted result,
retained failed/refresh state or Revert cannot reopen the menu or restore its
obsolete focus. The shared helper's positional fallback, menu default and
Filters form mode remain in place. No timer, DOM-position identity, second focus
manager or new query state was introduced.

## Decision, compatibility and rollback

The common decision in the shared boundary is how a retiring registered item
chooses an eligible successor while the overlay still owns focus. A Sort-only
semantic callback expresses field/action identity; other consumers keep their
existing positional fallback. Leaving the gap requires users to reopen Sort or
reacquire focus after routine edits. This is a bounded interaction correction,
with no new sort operation or visual redesign. Another menu with semantic item
identities can opt into the callback; no speculative consumer was added.

The accepted query and requested editor remain distinct. Existing header
shortcuts, field capability checks, limit, ordering, request and continuation
identity, canonicalization, persistence, permission checks and draft lifetimes
are unchanged. No public route, schema, stored data, migration, compatibility
adapter or dependency changes. Rollback reverts the authored web source and
tests plus three catalog rows, then runs `make generate` to restore the two
derived topology/browser files. It needs no data rollback.

## Verification and limits

| Evidence | Result |
| --- | --- |
| Baseline component/model and shared helper slices | PASS at `.cartulary/test-results/20260924T001139Z-p82531` and `20260924T001154Z-p83132`. An earlier run stopped before tests because `node` was absent from `PATH` at `20260924T001112Z-p81486`; all later Make calls used the pinned `tmp/node-runtime/bin`. |
| New expected-red component row | Product FAIL at `.cartulary/test-results/20260924T004756Z-p97094`: Add detached focus; pointer-focused Remove was not roving. |
| New expected-red Chromium row | Product FAIL at `.cartulary/test-results/20260924T005037Z-p2794`: first keyboard Add did not focus direction. |
| Authored generation | `make generate` PASS at `.cartulary/test-results/20260924T010108Z-p81404` and after the last catalog title at `20260924T011118Z-p83576`. Only authored catalog rows were edited by hand; generated browser manifest/topology index came from Make. |
| Final focused units | `web.workbook` component/model PASS after the final helper change at `.cartulary/test-results/20260924T011305Z-p25019`; `web.application` helper PASS after the empty-menu correction at `20260924T011214Z-p88804`. Tests cover Add, direction, disabled reorder, middle/final removal, pointer-to-keyboard Arrow/Home, one roving item, chip return, requested/accepted delay, outside departure, retained failure/refresh, Revert and surface change. |
| Production Chromium | Final new Sort plus existing Filters and Timeline row-action menu rows PASS, 13/13 units at `.cartulary/test-results/20260924T011233Z-p89411`. The Sort row opens the editor for a continuous keyboard edit, adds, changes direction, reorders, removes, reaches/reopens limit capacity, enters from an accepted chip, removes its invoking entry and returns to Timeline grid. It asserts each transition and uses no focus-repair click or `locator.focus()`. A subsequent pointer-to-keyboard segment starts from an actually clicked direction control and Tabs out to Group, closing Sort without stealing focus. Saved-view menu accessibility PASS, 11/11 at `20260924T010428Z-p32951`; existing Timeline sort/filter/group query browser row PASS, 11/11 at `20260924T010528Z-p71774`. |
| Static and generated checks | Typecheck PASS after the final source change at `20260924T011401Z-p31396`; import boundary PASS at `20260924T011243Z-p4227`; Biome PASS at `20260924T011243Z-p4364`; catalog PASS; JSON shape PASS at `20260924T010443Z-p58148`; generated policy PASS at `20260924T011243Z-p3741`; generation drift PASS at `20260924T011243Z-p3770`. An interim Biome run found an unformatted Sort legend at `20260924T010400Z-p27283`; `make format` passed at `20260924T010421Z-p28557`, then Biome passed. |
| Interim failures | A new no-eligible helper test failed at `20260924T011134Z-p86676`: focus reached the trigger but the empty menu stayed open; the helper now closes it. A browser rerun stopped before Playwright at `20260924T010922Z-p44944` with `artifact_error` because the authored helper title was not ASCII-sorted; that catalog ordering was fixed and regenerated. An accompanying fixture cleanup error was harness fallout, with no product assertion. |
| Finalizer | `make agent-finalize` PASS before broader final verification at `.cartulary/test-results/20260924T010215Z-p87049` and after the final source and catalog edits at `20260924T011401Z-p31302`. `RESULTS_DIR` was unset because no eligible successful full warm-check run existed; retained-run maintenance was skipped. |
| Documentation | `make lint-markdown` PASS at `.cartulary/test-results/20260924T011401Z-p31431`. |

The production check is pinned Chromium evidence, not a cross-browser claim.
The delayed-result cases are controlled presentation-state tests; the existing
query browser row checks actual server query behavior. No visual golden or
measured performance improvement is claimed.

## Digest acceptance

| Rows | Assessment |
| --- | --- |
| A001–A003 | PASS. Exact owner/source/routing boundaries, clean baseline, dated digest drift, bounded common decision, extension path, retirement, risk and observable exit are recorded above. |
| A004, A011, A019–A020 | PASS. No design token or theme registry was added. Focus continuity, keyboard parity, disabled/replacement controls and recovery use the focused unit and production Chromium evidence above. |
| A016 | PASS for the touched query presentation boundary: requested Sort is explicitly unapplied, accepted chips/rows remain accepted, and existing server query behavior has its owner-routed browser regression. No separate interaction-permission model changed. |
| A023–A027 | PASS. Tests use semantic field/view identities and owner UI contracts; executable checks do not read Markdown; generation and policy checks pass; compatibility and rollback are explicit; this handoff records failures, limitations and skipped work. |
| A005–A010, A012–A015, A017–A018, A021–A022 | N/A to this interaction slice. It changes no theme, density, creation, shell layout, inspector, transaction, acknowledgement, editing, conflict, authorization, evidence, virtualization or visual fixture owner. |

No applicable acceptance row is blocked.
