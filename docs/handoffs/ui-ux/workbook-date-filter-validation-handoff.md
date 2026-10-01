# Workbook date-filter validation and local correction

## Control and baseline

This is the controlling record for the bounded date-filter slice.
Implementation starts on clean `main` at
`c2d04ecd9ec07dd97131057658ba51f621a07ba6`, matching `origin/main`.
The readable-source and structural heading/gutter fixes remain outside this slice.

| Workstream | Status | Exit evidence |
| --- | --- | --- |
| Characterization | DONE | Fresh seeded reproduction plus four expected failing assertions in `20261001T195244Z-p16424`. |
| Shared admission and local feedback | DONE | Shared model decision and all Apply callers migrated; eight focused frontend rows passed in `20261001T200738Z-p29697`. |
| Integrated browser and rendered evidence | DONE | Five real-route scenarios pass in `20261001T202547Z-p6178`; eight final rendered captures and viewer filtering inspected; cleanup complete. |
| Final verification and handoff | DONE | Final frontend/browser checks, Markdown lint and diff/scope review pass; all applicable acceptance rows pass. |

## Authority

Core 01 REQ-01-039 through REQ-01-045 owns fields, operators, canonical dates,
set/null behavior and invalid-query rejection. Core 03 REQ-03-223/224/286/299,
section 14.9 and applicable keyboard/continuity clauses govern Workbook behavior.
Design sections 8.3, 8.4, 12 and 14 govern local feedback, accepted chips and focus.
Domain supplies vocabulary and navigation. Digest prompts, historical handoffs
and research advice are source material; no executable consumer reads Markdown.

Source ownership is `web.workbook`; verification is independently routed through
`web.workbook` regression and `module.workbook` browser families. View/UI contracts
are consumed through their public facades. Generated outputs follow their machine
policy and generators.

## Current evidence

Baseline selected frontend regressions passed at run
`20261001T191720Z-p75627` (6/6 execution units). That pass did not characterize
the date defect.

Fresh seeded default-profile review run `20261001T194738Z-p85030` reproduced:

- Timeline equality `2026-04-31` enabled Apply; applying closed the editor and
  returned HTTP 400 with stale-data feedback and 66 retained rows.
- Reopening the unapplied entry and applying `2026-04-18` returned HTTP 200.
- Lower `2026-04-19` / upper `2026-04-18` enabled Apply.
- The lower textbox was unnamed; the upper textbox was named.

The 1440x900 compact/dark-graphite captures `bundle-10` and `bundle-24` were
integrity-checked and visually inspected before cleanup. Axe completed with no
violations and incomplete `aria-valid-attr-value` / `color-contrast` observations;
it did not certify operand naming or accessibility. Private images and observation
paths are not retained. The session stopped successfully and the foreground
process exited with cleanup complete.

## Implementation boundary

`workbookQuery.ts` now owns the pure contract-aware validation/admission result.
It constructs accepted filter arguments once per decision and privately checks
calendar dates and range ordering. The controls model's old buildability-only
validator is retired. `applyFilterDraft` requires a ViewContract and returns the
original query object on refusal. UI feedback, general controller, separate
Timeline presentation, candidate staging and filter restoration use the same
function. Void Apply callbacks are replaced with synchronous admission results;
overlay close/focus return precedes editor completion only after valid admission.

The calendar helper uses four-digit syntax and Gregorian arithmetic, with no
Date rollover, local time interpretation or additional year restrictions.
Whitespace trimming, equality-set splitting/deduplication/order and null encoding
remain compatible. Invalid sets reject the whole draft. One-sided ranges pass;
empty, malformed, reversed and equal-exclusive ranges fail. Date handling follows
readKind, covering nine current date-filter fields across seven declared views:
Timeline, Communications Log, Handoff, Status Review, Lesson, Investigative
Queries and Forensic Keywords. This list documents discovered consumers and is
not a production allowlist. Timestamp operands keep existing permissiveness.

Feedback names date values and lower/upper operands, associates instance-local
status IDs with affected controls, and keeps raw text in the existing editor.
No new draft store, dependency, schema, route, calendar picker, grid-vendor
integration, serializer, saved-view format or cursor owner is introduced.
Canonical response adoption and stale-data/Retry/Revert remain in existing query
browser/state owners. Query capability remains independent of mutation permission.

This boundary hides the common calendar/range decision and prevents UI/controller
divergence. New contract-declared date fields automatically consume it; no future
capability is implemented speculatively. Non-date buildability remains deliberate.
Supported-field/operator admission is also enforced for direct callers; the old
controller test's unsupported Timeline full-text fixture was replaced with a
contract-supported tag filter while preserving its consecutive-edit assertion.
The relevant Notes full-text shell regression still passes.

This is frontend-only with no backend or persisted-data migration. Rollback is a
revert of this slice and its tests/routing. Timeline capture text and ordinary
cell editing are unchanged. The integrated and rendered workstream now passes; limitations are recorded below.

An intermediate sealed review (`20261001T201104Z-p82567`, bundles 7/9/14)
confirmed date naming/feedback and keyboard correction but exposed the existing
filter panel's clipping at 768x640. Tab to Apply scrolled the entire page sideways.
That session was stopped with cleanup complete before source edits. The filter
panel alone now aligns to the end of its trigger; repeat inspection passed.
The initial new browser run (`20261001T201021Z-p43849`) passed Communications Log;
Timeline timed out reaching an offscreen draft column after Inspector navigation.
Its fixture now uses the existing semantic grid scroll helper.

## Verification ledger

Commands run from the repository root through public Make targets. Routing was
rediscovered through help/help-all and both owner task guides. The new browser
row ID was derived through `author-test-row-id`, authored in
`tools/test_families/module.workbook.json`, then generated with `make generate`.
Its resolved file is `apps/web/e2e/workbook-query-browsing.spec.ts`, Chromium,
stateful, with separate Timeline and Communications Log titles/scenario IDs.
Only browser-batch and topology-index generated files changed; neither was edited
by hand. The frontend family extends the existing model/component/query/controller
row selectors.

| Command/selection | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make format` after final product edits | PASS; diff inspected, authored frontend only | `20261001T201615Z-p14192` |
| `make generate` | PASS; authored routing projected | `20261001T200953Z-p35686` |
| `make agent-finalize` before terminal verification | PASS | `20261001T201628Z-p18602` |
| `make test-slice OWNER=web.workbook ROWS=...` (eight selected rows below) | PASS, 9/9 execution units | `20261001T201713Z-p22942` |
| `make frontend-typecheck` | PASS, 2/2 | `20261001T201714Z-p23230` |
| `make lint-biome` | PASS, 2/2 | `20261001T201714Z-p23328` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `20261001T201714Z-p23308` |
| `make generate-drift` | PASS, 4/4 | `20261001T201713Z-p22903` |

The eight frontend row IDs were:

- `web.workbook.regression.grid_controls_model_a104000001`
- `web.workbook.regression.grid_controls_component_a104000002`
- `web.workbook.regression.workbookquery_suite_904073db6c`
- `web.workbook.regression.useworkbookquerycontroller_suite_203a4e98cc`
- `web.workbook.regression.use_workbook_query_controller_keeps_query_defaul_bfad01a7b4`
- `web.workbook.regression.workbook_query_controls_preserve_ordered_chip_ca_0fc544e3c5`
- `web.workbook.regression.workbookshell__query_normalizes_saved_view_query_22724faeb7`
- `web.workbook.regression.workbookshell__query_notes_full_text_controls_su_889a129c8b`

The model coverage discovers date fields from current contracts, exercises leap
centuries/year 0000, malformed/impossible/date-time operands, trimming, lists/null,
range combinations and schema refusal. Component coverage checks raw correction,
operand names and feedback IDs, native keys, Escape/Cancel, focus return,
completion order, callback refusal and candidate staging with instance-local IDs.
Controller coverage checks direct refusal and raw/request retention, latest-intent
consecutive editing, view-schema isolation and Workbook-instance isolation.

Development failures were related to this slice and resolved before terminal
verification: expected characterization reds (`195244Z-p16424`); controller's
unsupported fixture (`195824Z-p19812`); component comparison-name/native-Tab
fixture assumptions (`200442Z-p22355`, `200638Z-p23600`); test-only type options
(`200803Z-p31950`, `200855Z-p33449`); formatting/non-null assertions
(`200855Z-p33502`). Each abbreviated ID above has the date prefix `20261001T`.
The failed finalizer (`20261001T200706Z-p29094`) and json-shape checks
(`20261001T200802Z-p31696`, `20261001T200855Z-p33228`) identified stale generated
routing before `make generate`. Browser execution before generation was refused
with “missing generated browser groups,” producing no product pass. The initial
service-backed fixture failure is recorded above. Run-summary.json and per-unit
logs/reports explain those failures; none is treated as successful evidence.

Retained-run maintenance was skipped because RESULTS_DIR was unset: no successful
full warm check qualifies. Broad backend, release, performance and full visual
suites were skipped because this frontend date-query seam changes no corresponding
owner. No visual goldens were refreshed. Review captures and Axe remain supporting
observations, not product tests or accessibility certification.

Final production-renderer inspection used sealed default-profile run
`20261001T201714Z-p24097`. Integrity-checked, visually consumed captures were:
1440x900 impossible equality (bundle 7), date-time rejection (9), contradictory
range/longest feedback (13), corrected inclusive range with keyboard focus (22);
768x640 contradictory range (15), corrected range with Apply focus (20), viewer
malformed equality (32), and corrected viewer equality with Apply focus (36).
The panel, message, operand controls and actions stay inside the viewport after
alignment; Tab no longer scrolls the page sideways. Production snapshots name
both bound values and comparisons, mark both reversed bounds invalid and expose
one local status message. Editor and viewer valid keyboard Apply both returned
HTTP 200 and focus to Filters; rows remain usable. Escape returns to the trigger.
The viewer's read-only cells coexist with query controls and successful reads.
Axe completed with zero violations and incomplete aria-valid-attr-value and
color-contrast observations in each capture. These observations do not establish
screen-reader announcement cadence or broad accessibility certification.
All three review sessions stopped successfully with cleanup complete and their
foreground processes exited. Private artifacts expired and are not linked here.

The next service-backed run (`20261001T201713Z-p22973`) passed Communications Log,
canonical continuation and both recovery-focus scenarios. Timeline's draft
retention assertion exposed the fixture's assumption that rough capture stays
unsubmitted: first input creates immediately. The test now holds that independent
create request through invalid filter editing, then disposes it in a finally
block. Production rough capture is unchanged. Browser re-verification passed in the final run below.

Final fixture-only edits were formatted (`20261001T202452Z-p97219`) and followed
by `make agent-finalize` (`20261001T202509Z-p2000`, PASS). Final
`make frontend-typecheck` (`20261001T202548Z-p6266`) and `make lint-biome`
(`20261001T202548Z-p6308`) passed. Generated-artifact policy passed
(`20261001T202657Z-p39518`, 3/3 units).

The final service-backed command selected these three confirmed rows:

- `module.workbook.browser_stateful.date_filter_drafts_remain_locally_correctable_be_15ce5a6e7c`
- `module.workbook.browser_stateful.query_continuation_canonical`
- `module.workbook.browser_stateful.query_recovery_focus`

`make service-backed-test-slice OWNER=module.workbook ROWS=...` passed 11/11
execution units in `20261001T202547Z-p6178`. Its Playwright report records all five
scenarios passed: Timeline correction, Communications Log correction, canonical
continuation/saved-sort intent, Timeline recovery focus and accepted-empty Notes
recovery focus. The new scenarios observe requests when they begin, including
aborted/failed attempts. Invalid equality, reversed/equal-exclusive ranges and
invalid edits of failed replacements generate no draft-driven query; raw input,
accepted rows/chips, Inspector and grid anchor remain intact. Timeline also retains
selection and an independently pending authoring draft. Valid correction receives
HTTP 200. Genuine read failure retains accepted chips/rows and existing Retry;
correcting its unapplied filter restores successful browsing. Tests exercise
Escape to an invoking chip and 768x640 panel bounds. Fixture leases are released,
service cleanup succeeded, and lifecycle events end in cleanup_succeeded.

## Owner/change and acceptance assessment

| Owner clauses | Bounded result |
| --- | --- |
| Core 01 REQ-01-039/040/041 | Contract field/operator admission; unchanged operator shapes, logical equality sets and null encoding. |
| Core 01 REQ-01-042 | Date-valued query operands validate canonical real calendar dates after existing trimming; timestamp handling is retained outside this slice. |
| Core 01 REQ-01-043/044/045 | Reject local empty/malformed/contradictory date drafts while preserving independent server validation and invalid_view_query errors/details. |
| Core 03 REQ-03-223/224 and section 14.9 | In-sheet editing with latest requested intent, server-accepted chips, canonical continuation, existing failed-read recovery and independent authoring continuity. |
| Core 03 REQ-03-286/299 | Filter overlay stays within supported work area; query/read capability remains separate from mutation authority and existing scoped cleanup. |
| Design section 8.3 and applicable keyboard/local feedback sections | Invalid drafts excluded from submission; accepted rows/chips retained; named operands, associated feedback, local correction, Cancel/Escape and focus return. |

Reviewed navigation inputs include Workbook models/components/hooks/query/view-state
and Timeline presentation source guides; `tools/frontend_source_ownership.json`,
`tools/frontend_import_boundaries.json`, generated-artifact policy and current
verification catalog/families. Public ViewContract/UI facades remain the imports.
No direct grid-vendor import was added. The digest's September 13 localization
commit differs from this baseline; current machine inputs and source guides were
revalidated instead of treating its snapshot as current evidence. Existing stack
and dependency pins are retained. No adopted-owner contradiction was identified.

Applied digest dispositions: ADOPT R001/002/003/006/007/011/013/014/015 for
keyboard/focus, text feedback, state separation, shared tokens, semantic controls,
continuity and handoff; ADAPT R018/024/035 to the local filter panel's bounds and
long feedback. No upstream query, palette, framework or new behavior authority
was needed. No digest or Markdown consumer was introduced into executable code,
tests, runtime metadata or generators.

| Acceptance | Status | Evidence or scope rationale |
| --- | --- | --- |
| A001 | PASS | Owner/change map above; adopted behavior separated from source and verification owners. |
| A002 | PASS | One pure calendar/range admission boundary; old validator retired, callers migrated, scope and compatibility reviewed. |
| A003 | PASS | Baseline/current Git state, current source/import manifests, package facades, generated policy and guides inspected; localization drift recorded. |
| A004 | PASS | Existing tokens/styles retained; no design literal or registry added. |
| A005 | PASS | Compact dark_graphite seeded captures; no theme changes. |
| A006 | N/A | No density selection, row geometry or shared density owner changes; existing compact rendering observed. |
| A007 | N/A | No create capabilities or inputs change; held synthetic Timeline create only observes independent authoring continuity. |
| A008 | PASS | Filter overlay inspected at 1440x900 and 768x640, including keyboard focus; browser bounds assertion. Responsive selection and Inspector clamp/fallback algorithms unchanged. |
| A009 | PASS | Filter correction no longer induces document horizontal scrolling; grid, status and shell navigation remain visible in narrow captures. |
| A010 | PASS | Existing Inspector context retained during invalid edits in both real-route consumers; dispatcher/review/source lifetime mechanisms unchanged. |
| A011 | PASS | Real-route anchors, rows/chips, selection and independent raw draft continuity; controller latest-intent/schema/instance regressions. |
| A012 | N/A | No transaction identity, replay bytes or write admission changes. |
| A013 | N/A | No mutation acknowledgement/re-key/uncertain replay change; read recovery is covered under A016. |
| A014 | PASS | Exact local filter input retained, keyboard correction and Cancel/Escape/focus semantics tested; ordinary cell editing unchanged. |
| A015 | N/A | No cell conflict, saved-value resolution or toast behavior changes. |
| A016 | PASS | Invalid drafts excluded before query replacement; genuine failed reads/recovery remain; supported viewer filtering returns HTTP 200 independently of read-only mutation state. |
| A017 | PASS | Guards precede state clearing; accepted stale rows retained in browser scenarios. Scoped authorization cleanup paths untouched, with no new persistence or retention promise. |
| A018 | N/A | No evidence lifecycle, preview or overlay behavior changes. |
| A019 | PASS | Named operands, affected-control feedback associations, one stable status region, native keys, Tab/keyboard Apply, Escape and focus return verified; Axe limitations disclosed. |
| A020 | PASS | Empty, invalid, contradictory, corrected and long-feedback filter states remain operable in representative rendered inspection. Density/zoom/text-spacing algorithms unchanged; no broad text-layout claim. |
| A021 | PASS | Existing canonical continuation/recovery rows pass; stable row/anchor behavior during invalid edits verified. Virtualization/window owners unchanged, so performance suites are N/A. |
| A022 | PASS | Eight exact sealed production captures consumed with digest/byte verification; theme/density/viewports recorded. No golden refresh or publication claim. |
| A023 | PASS | Existing public view/record/field selectors and semantic role/names used; no selector registry changes require package.ui tests. |
| A024 | PASS | Test/dependency/diff review finds no Markdown reads, stats or hashes in any new executable consumer. |
| A025 | PASS | Authored families changed first; public generate/finalizer, drift and policy checks passed; generated outputs inspected. |
| A026 | PASS | No new route/schema/storage/permission/timestamp behavior; supported non-date regressions pass; frontend-only revert and no-migration result explicit. |
| A027 | PASS | Completed controlling record, final verification, rendered inspection, cleanup and scope review; no applicable blocker remains. |

The only presentation addition discovered during required inspection is end
alignment of the existing Filters panel. It prevents clipped correction controls
at the requested narrow viewport and changes no other overlay or shell algorithm.
All implementation work is reviewable in the working tree; no commit, publish,
merge or deployment was performed. No implementation work remains in this slice; the next action is review of the working-tree diff.

`make lint-markdown` passed in `20261001T203009Z-p40907`, covering this authored
handoff; its summary is `adhoc/lint-markdown/tool-run-summary.json`. Final
`git diff --check` passed. The final branch/HEAD/origin remain `main` at
`c2d04ecd9ec07dd97131057658ba51f621a07ba6`; the working tree contains only the
seventeen product/test/routing files listed by the scoped diff and this handoff.
No backend, contract, dependency/lockfile, persisted-data or unrelated heading/
source-context path changed. Caller-owned review request scratch was removed.
