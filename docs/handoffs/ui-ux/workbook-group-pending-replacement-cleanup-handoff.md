# Workbook Group selector during query replacement

## Scope and ownership

This bounded correction was made from clean `main` at
`41931b2bd6f56302940a5f41091957feabc2e06d`. The user authorized the
implementation. Core 03 §14.1 REQ-03-224 governs latest applicable queries,
canonical metadata and saved-view clearing; §14.3 REQ-03-226/227 governs declared
group options and omitted `group_by`; §14.9 requires requested overrides,
canonical query and accepted windows to remain distinct, with accepted rows,
chips and grouping retained during pending or failed replacement. Design
§§8.3–8.5 and §14.1 govern the compact control, chip entry, keyboard behavior,
focus and non-color feedback. `docs/domain.md`, `docs/research/nlspec-spec.md`,
the UI/UX digest and prior handoffs supplied navigation or advice, not new
requirements.

The source boundary is `apps/web/src/workbook`: the existing per-surface
requested query controller, accepted query browser, shell and Timeline
presentation, and query controls. Its import/source boundaries were rechecked
against current source and the repository ownership inputs. Verification routes
independently through `web.workbook` component/query rows and
`module.workbook` stateful browser rows. Both current `make task-guide
ROLE=module-author` owner routes were inspected. No Grid Adapter or backend
ownership changed. The digest's September 13 localization refers to commit
`59fa79e`; this run uses `41931b2` and current source/catalog inputs, so its
snapshot paths and historical passes were not used as execution evidence.

## Observed defect and decision

Before the production edit, a Timeline Capture State result was accepted and a
replacement for Has Evidence was held open. A real native-select ArrowDown and
Enter sent `group_by=timeline.has_evidence`, but the focused controlled select
returned to `timeline.capture_state`; the Capture State chip and group header
correctly still described retained results. The same snapback occurred after a
failed replacement. The deliberately red characterization run is
`.cartulary/test-results/20260924T032159Z-p20327`; its Playwright report and
both traces are under `browser-e2e-stateful/browser-groups/stateful-default-workbook-query-browsing/`.
The `group-pending-baseline` and `group-failed-baseline` attachments in
`playwright-report.json` capture select value, focus, accepted chip/header and
request body. The assertion received Capture State where Has Evidence was
requested. An earlier red test run used ArrowDown without Enter and did not
commit a native option; that test interaction was corrected before declaring
the product defect.

The common decision is which query owns each visible part: requested grouping
owns the editable select and its title; the accepted presentation query owns
chips and rendered group rows. The shell and Timeline inline bindings now pass
`requestedGroupBy` from their existing requested owners. The shared controls
compare requested and accepted keys and show **Unapplied** beside Group only
while they differ. The select's accessible description names both the requested
choice and retained grouping, including None. This is a behavior correction in
projection and presentation; no shared query abstraction or third store was
added. Other declared-grouping surfaces use the same control binding.

The shell's incident/surface/sheet-reference/saved-view-version subject key now
reaches Timeline's inline controls. A deferred Group Escape focus return checks
that key before restoring focus. The requested query remains per surface;
saved-view selection applies its authored query through the existing controller.

## Behavior and compatibility

- During a held replacement, successive native Arrow-key choices remain in the
  focused selector while the accepted chip and grid groups stay with the prior
  authorized result. The latest applicable acceptance changes those chips and
  groups; an obsolete success or failure cannot undo it.
- On current failure, the selector keeps requested intent and the accepted
  result remains visible. Existing Retry reads the current request; Revert
  restores the accepted authored query and selector. Group: None is internal
  `null` and is omitted from public query and saved-view payloads.
- Native option order, select keyboard/pointer semantics, composition and
  group-chip activation/dismissal remain owned by their existing controls.
  The existing browser generation/abort fences, canonicalization, retained
  rows/drafts, filters, sorting, continuation and saved-view behavior remain.

The removed mechanism is the Group select and title's binding to the accepted
query. Accepted chips and rows deliberately retain that binding. The optional
requested prop retains supported direct component callers, which fall back to
the accepted query. There is no route, schema, stored-data, dependency or
migration change. Rollback restores the production bindings, authored test
catalog, generated derivatives, tests and this handoff together, without
touching analyst data or later unrelated work. Leaving the old binding would
continue to make rapid keyboard grouping unreliable and obscure unapplied
intent. No latency improvement was measured or claimed.

## Verification

All paths below are relative to the repository root. The new
`module.workbook.browser_stateful.group_requested_replacement` row covers
Timeline's held and failed queries, real consecutive Arrow-key selection without
focus repair or `selectOption`, adversarial late response release, accepted
chips/groups, Retry, Revert, None, chip entry/Escape, pointer-opened native
selection and Hosts. The existing canonical continuation row verifies omitted
`group_by` in public requests and saved views. Component and query-browser rows
cover projection transitions, current and superseded failure, Retry/Revert,
None, stale focus, and existing generation fences. New authored catalog titles
were projected with `make generate`; no generated output was hand-edited.

| Check | Result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260924T033239Z-p30668` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.grid_controls_component_a104000002,web.workbook.regression.query_browsing` | PASS 3/3 | `.cartulary/test-results/20260924T033256Z-p33788` |
| Final Group component slice after status-style review | PASS 2/2 | `.cartulary/test-results/20260924T034007Z-p14728` |
| `make frontend-typecheck` | PASS 2/2 | `.cartulary/test-results/20260924T033256Z-p33923` |
| `make frontend-import-boundary-check` | PASS 2/2 | `.cartulary/test-results/20260924T033256Z-p33961` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.group_requested_replacement,module.workbook.browser_stateful.query_continuation_canonical` | PASS 11/11 | `.cartulary/test-results/20260924T033404Z-p35888` |
| Final Group browser row, including chip focus return | PASS 11/11 | `.cartulary/test-results/20260924T033550Z-p69896` |
| `make format`, `make lint-biome`, `make json-shape-check` | PASS | `.cartulary/test-results/20260924T033128Z-p25724`, `.cartulary/test-results/20260924T033404Z-p36047`, `.cartulary/test-results/20260924T033403Z-p35810` |
| Final `make lint-biome` | PASS 2/2 | `.cartulary/test-results/20260924T034007Z-p14848` |
| `make agent-finalize` | PASS 1/1 | `.cartulary/test-results/20260924T033651Z-p3297` |
| `make generate-drift`, `make generated-artifact-policy-check` | PASS 4/4, 3/3 | `.cartulary/test-results/20260924T033739Z-p7304`, `.cartulary/test-results/20260924T033739Z-p7323` |
| Final `make lint-markdown` | PASS | `.cartulary/test-results/20260924T034104Z-p16726` |

`RESULTS_DIR` was unset for `make agent-finalize`: no eligible successful full
warm check run was supplied, so retained-run maintenance was skipped. Full
`make test-fast` and whole browser/a11y/visual suites were not selected; the
owner-routed slices cover the changed boundary. Chromium's native-select popup
and pointer-opened selection were exercised; option selection in the pointer
case uses Playwright `selectOption`, so that case does not prove a physical
mouse click on an operating-system option. No visual golden, latency
measurement or full WCAG conformance claim is made.

## Digest acceptance assessment

The digest is advisory. For this slice, A001–A003 pass on the authority/source/
verification mapping, clean baseline, current source inspection and bounded
projection decision above. A004–A005 pass because the cue uses existing ink
tokens and theme styling without a second palette. A011, A016–A017 and A019
pass for the applicable retained-query, keyboard, focus and non-color state
transitions in the focused component and production Chromium evidence. A023–A026
pass on semantic selectors, no Markdown-dependent executable checks, generated
artifact ownership and unchanged public compatibility. A027 passes with this
handoff and run artifacts. A006–A010, A012–A015, A018, A020–A022 are N/A
because density, creation, responsive ownership, inspector, transactions,
editing, conflict, Evidence, component variants, virtualization and visual
fixtures were not changed. The applicable claims are scoped to this cleanup
and do not certify the whole workbook or WCAG profile.
