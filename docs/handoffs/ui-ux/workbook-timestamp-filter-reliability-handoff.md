# Reliable workbook timestamp filtering

## Scope and owners

This cohesive slice starts on clean `main` at
`324b0c7b540f1c79027f7936cc3951c8cb3e5f6d`. HEAD and branch were rechecked;
no unrelated changes were present. Root `AGENTS.md` applies; no narrower
instructions were found. The refactor digest, source guides, owner maps and both
requested skills supplied navigation and review procedure, not new authority.

Core 01 REQ039–046 owns timestamp query operands, normalization, null/set
semantics, permitted operators and contradictory ranges. Section 18A applies
only where a binding declares it; writable-field clearability does not restrict
query null equality. Core 03 REQ220, REQ223–224, section 14.9 and applicable
keyboard, candidate identity and authorization clauses govern submission and
continuity. Design section 8.3 and neighboring feedback/recovery guidance govern
local correction. Domain supplies vocabulary and owner navigation.

The Audit and Indicator parsing precedents were inspected. Their workflow and
canonical-output policies remain separate; only the narrow Workbook operand
decision is introduced here.

Current view contracts identify 11 timestamp filters in eight workbooks:
Task Requests due/completed, Parties updated, Evidence requested/received,
Findings closed, Indicators first/last observed, Notes updated, Decisions decided
and Assessments assessed. This is a discovered inventory, not a new registry.
Timeline has no timestamp filter and gains none. Boolean authoring is retained.

Production changes are confined to:

- `apps/web/src/workbook/models/workbookTimestampFilterOperand.ts`: pure explicit
  timezone/calendar parsing and nanosecond comparison key.
- `apps/web/src/workbook/models/workbookQuery.ts`: shared timestamp admission,
  before lossy set construction, with operand-specific feedback controls.
- `internal/platform/viewquery/query.go`: field-kind-aware range contradiction
  detection; normalized timestamps compare `time.Time` instants, dates retain
  canonical string ordering.

Source ownership is `web.workbook` and `platform.viewquery`. Verification routes
also include `module.workbook` and `module.savedviews`. The new source is listed
in the frontend ownership manifest and models guide. Test titles/rows are
registered in the four authored family inputs; IDs came from public
`author-test-row-id`. `make generate` alone changed browser-batch and topology
render-index projections. No executable dependency reads Markdown.

## Characterization and behavior

The initial new regressions failed on the reviewed source: `tomorrow` was locally
admitted; the whole-second to 100 ms range was rejected; its reverse was accepted
by both direct decoding and persisted normalization. The optional fractional
portion of RFC3339Nano makes lexical comparison unsuitable for instants.

The shared validator now rejects malformed/calendar-invalid/zone-less equality,
set members and range bounds. Every raw set member is checked, including empty
and non-string entries. Inputs retain raw editing text; whitespace is trimmed
only for frontend admission/serialization. Raw server whitespace remains invalid.
Feedback identifies the affected operand and explains `Z` or numeric offsets
with examples through the editors' existing accessible associations.

Authoring syntax is four-digit years, valid Gregorian dates, two-digit hours,
minutes and seconds, optional decimal fractions, and explicit uppercase `Z` or
`+/-HH:MM`. UTC arithmetic avoids browser timezone inference. Fractional digits
beyond nine remain accepted and comparison truncates them to nanoseconds, matching
existing Go normalization. Raw drafts and outgoing frontend text retain those
digits; authoritative canonical responses retain the supported nine digits.
Go's incidental comma fractions, single-digit hours and overflowing offset
components are characterized separately and are not new UI syntax. Server scalar
normalization is unchanged; the wire validator was not weakened.

One-sided ranges remain valid. Forward whole-second/fractional and nanosecond
ranges pass; reversed ranges fail. Offset-equivalent equal instants require both
bounds inclusive. Scalar sets keep existing canonical ordering; cursor semantics,
SQL, timestamp writes and date-only behavior are unchanged. Nanosecond query
normalization is distinct from PostgreSQL timestamp storage precision: populated
membership fixtures use representable millisecond values at both endpoints,
inside, before and after the interval.

`WorkbookFiltersControl`, `WorkbookCandidateQueryControl`, the main controller,
Timeline presentation, candidate discovery/reference picker and Assessment
discovery already share this validation owner. No editor parser or query-state
owner was added. Invalid Apply/Add cannot replace accepted filters, close the
editor or issue a query. Workbook Apply follows canonical request acceptance;
candidate Add stages, and explicit candidate-query Apply discovers. Existing
latest-request handling, selected identity/labels, authorization fences, native
keys, deliberate dismissal/focus return, accepted chips and Retry/Revert remain
with their existing owners. Audit and Indicator workflow parsers are unchanged.

This boundary retires timestamp draft permissiveness and lexical timestamp range
ordering. New timestamp fields declared by existing contracts naturally consume
the common decision. It introduces no general time framework or duplicate state.

## Persisted compatibility and correction

Valid queries retain their wire shape and canonical instant meaning. Tests cover
saved-query nanosecond/offset round trips, create/update, bundle preparation and
export. Older contradictory ranges may now fail execution or normalization.
Independent saved-view reads preserve the stored predicate. Failed query,
create/update, preparation and export validation do not delete or repair it.

The correction path is to open Filters, edit the unapplied range, apply a valid
range, then explicitly Update saved view. The database integration fixture writes
an older contradictory predicate directly, proves independent GET and workbook
startup preserve it and execution
returns `invalid_view_query` / `invalid_filter_operand`, then verifies explicit
PATCH correction. Browser coverage separately models the older resource read,
uses real query validation, and exercises Filters and Update saved view.
No automatic rewrite, predicate dropping, storage migration or data rollback is
required. Rollback reverts this frontend/server/test/catalog slice together and
regenerates derived outputs with Make.

## Verification ledger

All commands run from the repository root. Routing was discovered with
`make help`, `make help-all`, the four requested owner task guides, owner/target
explanations and plans. Run IDs below resolve beneath `.cartulary/test-results/`;
`run-summary.json` and unit artifacts contain results. UI-review captures are
private supporting observations, not product-test results.

| Command | Result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=platform.viewquery ROWS=platform.viewquery.unit.timestamp_query_operands_normalize_and_ranges_co_0b505c0d19` before fix | Expected FAIL | `20261006T232647Z-p18147` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookquery_suite_904073db6c` before fix | Expected FAIL | `20261006T232703Z-p19145` |
| `make test-slice OWNER=platform.viewquery ROWS=platform.viewquery.unit.timestamp_query_operands_normalize_and_ranges_co_0b505c0d19,platform.viewquery.unit.sort_and_filter_ceilings_canonical_normalization_4ca84c1e10` | PASS | `20261006T232906Z-p25740` |
| `make test-slice OWNER=platform.viewquery ROWS=platform.viewquery.unit.timestamp_query_operands_normalize_and_ranges_co_0b505c0d19` including date preservation | PASS | `20261006T234036Z-p64751` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookquery_suite_904073db6c` after final model test cleanup | PASS | `20261006T233825Z-p27182` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.grid_controls_component_a104000002` | PASS | `20261006T233411Z-p43470` |
| `make test-slice OWNER=module.savedviews ROWS=module.savedviews.unit.strict_saved_views_preparation_rejects_illegal_i_d1e9382c70` | PASS | `20261006T233423Z-p49693` |
| `make test-slice OWNER=module.workbook ROWS=module.workbook.unit.timestamp_instant_v1_accepts_rfc_3339_json_strin_ba11c9c407` | PASS | `20261006T233450Z-p79628` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.integration.timestamp_saved_queries_preserve_precision_and_r_993e5032bf,module.savedviews.integration.saved_view_export_emits_the_exact_canonical_elev_fb78b01a43` | PASS | `20261006T233517Z-p12334` |
| `make service-backed-test-slice OWNER=platform.viewquery ROWS=platform.viewquery.browser.timestamp_drafts_correct_locally_and_fractional_349e7ec0a3` | PASS | `20261006T233417Z-p45080` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.party_timestamp_candidate_filters_stage_explicit_45110487e8` | PASS | `20261006T233653Z-p87238` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.browser.timestamp_saved_queries_reload_and_reopen_withou_6a98a26dd7` including legacy correction and explicit Update | PASS, 11/11 | `20261006T235721Z-p27052` |
| `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.integration.timestamp_saved_queries_preserve_precision_and_r_993e5032bf` including startup preservation | PASS, 3/3 | `20261006T235538Z-p96262` |
| `make generate` | PASS | `20261006T233319Z-p35821` |
| `make format` after final source edits | PASS; diff inspected | `20261006T235711Z-p21754` |
| `make agent-finalize` before broader final verification | PASS | `20261006T235510Z-p77215` |
| `make frontend-typecheck` | PASS | `20261006T235814Z-p65535` |
| `make frontend-import-boundary-check` | PASS | `20261006T235402Z-p41040` |
| `make lint-biome` | PASS | `20261006T235837Z-p66722` |
| `make lint-markdown` | PASS; `adhoc/lint-markdown/tool-run-summary.json` | `20261006T235903Z-p67376` |
| `make generate-drift` | PASS | `20261006T234112Z-p8060` |
| `make generated-artifact-policy-check` | PASS | `20261006T234136Z-p35324` |

The registered browser rows use isolated synthetic incidents and the harness
worker-admin actor. The normal browser context is 1280×720; the populated range
row also changes to 768×640. The private rendered review below uses the separate
editor actor and explicitly selected 1440×900/768×640 viewports.

The combined frontend command selected these five rows:

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookquery_suite_904073db6c,web.workbook.regression.grid_controls_component_a104000002,web.workbook.regression.useworkbookquerycontroller_suite_203a4e98cc,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery
```

Final run `20261006T235336Z-p39148` passed all five rows (6/6 execution units),
including sparse-array rejection. Earlier run `20261006T232912Z-p26235` passed
model, controller, candidate and Assessment
rows; its grid row failed only because a new test expected the toolbar trigger
rather than the existing chip opener as the return-focus target. That assertion
was corrected and the grid row passed independently above. Native editing keys,
raw operands, attributed feedback, no invalid submission, staging and selected
identity continuity are covered. Assessment's Timeline view gains no new field.

Intermediate failures were resolved: finalizer `20261006T233219Z-p34367` and
`make json-shape-check` at `20261006T233300Z-p35155` found stale generated routing
before generation; frontend typecheck `20261006T233411Z-p43584` found test-only
locator options and union narrowing; Biome `20261006T233604Z-p59767` found three
test non-null assertions. None is counted as a pass. The earlier saved-query
browser run `20261006T233544Z-p31867` passed the valid round trip before adding
the explicit legacy correction path. Runs `20261006T234042Z-p67251` and
`20261006T235139Z-p90046` then timed out because the synthetic older-resource
fixture intercepted saved-view reads but omitted the selected resource embedded
in workbook startup. Trace inspection identified that fixture gap; startup and
ordinary resource reads now consistently carry the same older predicate.
Run `20261006T235447Z-p49477` reached correction but used the original backing
query as its update, producing an unchanged PATCH and an appropriately rejected
acknowledgement in the UI. The fixture now submits a distinct valid range and
asserts a real saved-view version increment; no acknowledgement policy changed.

Retained-run maintenance was skipped because RESULTS_DIR was unset; no qualifying
successful full warm check was used. Full release/backend, performance and visual
golden suites are outside this narrow risk surface. No goldens were refreshed.

## Rendered review and acceptance

Seeded default-profile session `20261006T234035Z-p64624` used the authorized
synthetic editor, dark_graphite/default density, Chromium, at 1440×900 and
768×640. Captures were resolved from their returned bundle manifests, checked
against byte lengths/SHA-256, and visually inspected before cleanup:

- Bundles 13/16: Notes `tomorrow`, local guidance and focused raw input at both
  sizes. No extra query during editing. Correction to explicit UTC returned 200.
- Bundle 73: Task Requests [whole second, +100 ms] returned 200 and retained the
  newly authored inside row at +50 ms, excluding the outside row at +101 ms and
  the existing null-due row. Product tests independently cover both endpoints
  and both outside directions with exact record IDs.
- Bundle 83: reversed bounds at 768×640 stayed editable; the accepted inside row
  and applied filter remained. No new query was issued.
- Bundle 106: a saved fractional query was fetched through the saved-view chooser
  after fresh navigation and reopened at 768×640 with its exact bounds intact.
- Bundles 136/147/150/157: the real Timeline contextual Task Request form's
  Requester Party picker selected the seeded Response coordination team. Its
  Parties Updated filter rejected `tomorrow`; Tab revealed the inline guidance
  and retained selection. Correction to a nanosecond UTC value and Add staged
  without discovery. Explicit Apply issued one query, returned no candidates,
  and retained the selected label; Apply references retained that identity.

These captures assess readable feedback, viewport placement and continuity;
they are not registered browser product tests, visual goldens or accessibility
certification. Axe completed: the empty Notes inline-create grid produced a
minor `empty-table-header` observation outside this filter slice (bundles 13/16).
Other captures had no completed violations; contrast and some expanded-control
ARIA checks were incomplete. No general accessibility conformance is claimed.
Genuine failed-read Retry/Revert and stale accepted rows were exercised separately
by the registered populated-range browser test, not by fault injection into the
private UI-review session. The final tiny sparse-array admission hardening is
model-tested; it does not change the string-input states captured here.

Two earlier preparation sessions were stopped cleanly before source changes.
The reviewed session's terminal receipt reports `closed`, `status=ok`,
`cleanup=complete`; its foreground process exited zero. Private observations and
images expired on stop and are not handoff links. Caller-owned request scratch
was removed. Invalid review-target requests were recovered with fresh snapshots;
none is product evidence.

| Acceptance rows | Assessment |
| --- | --- |
| A001–A003 | PASS: owner/change map, narrow common decision, actual clean baseline, manifests and final diff reviewed. |
| A004–A005 | PASS by diff: no tokens, theme or density registry changed. |
| A008–A009 | PASS for affected filter controls: desktop/constrained inspection, internal scrolling, reachable feedback and account/navigation chrome; unrelated responsive boundaries unchanged. |
| A011, A014, A016–A017 | PASS for this query seam: raw editing, explicit admission, accepted/requested state, candidate identities and read-failure recovery retained; existing controller/Assessment continuity rows passed. No authorization implementation changed. |
| A019–A020 | PASS for selected operand naming, error associations, native keys, focus restoration and filter states; scope is targeted interaction evidence, not global accessibility conformance. Axe limits are recorded above. |
| A022 | PASS: current-source seeded production renderer inspected with explicit viewports/theme/density; no automatic golden refresh. |
| A023–A026 | PASS: semantic view/field/record selectors, no Markdown runtime dependence, authored routing generated through Make, no new API/storage semantics or migration. |
| A006–A007, A010, A012–A013, A015, A018, A021 | N/A as changed-owner claims: density, creation policy, Inspector dispatch, mutation transactions/conflict, Evidence lifecycle and virtualization ownership are unchanged. Their existing consumers are retained. |
| A027 | PASS: focused implementation and compatibility checks pass; command artifacts, resolved failures, review limits, skipped checks and rollback are recorded. No required product check is blocked. |

Final `git diff --check` and complete working-tree review pass. The change is
uncommitted on the original branch. No remaining migration prerequisite or
owner contradiction was found. Previously stored contradictory ranges require
explicit correction as described above.
