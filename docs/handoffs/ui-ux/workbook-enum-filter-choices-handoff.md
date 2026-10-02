# Workbook enum filter choices

## Controlling record

Bounded implementation requested on 2026-10-01. Baseline: clean `main`,
`3833f7543d56ea2c12e82422c2da41ef7246da37`, synchronized with `origin/main`.
Date validation, retained Timeline bulk-tag authoring, and replay lifecycle
simplification remain outside this slice.

Behavior owners: [Core 01](../../spec/01_architecture_storage_and_view_contracts.md) §6 REQ-01-288 (ordered declared enum values),
§3.3.4.1 REQ-01-039–046 (field/operator eligibility, scalar/null/set shapes,
normalization, canonical metadata and rejection); [Core 03](../../spec/03_workbook_interaction_collaboration_and_workflows.md)
REQ-03-220/223/224/286/299/100/304 and §14.9 (native keyboard,
explicit query application, independent drafts and supporting-record identities,
request fencing, scoped authority); [design](../../design.md) §§8.3–8.5 and applicable §§12/14
control, responsive and accessibility direction. [Domain](../../domain.md) supplies vocabulary.
Digest guidance, research and historical handoffs provide navigation and evidence.
No adopted-owner contradiction or owner amendment is required.

Source owner is `web.workbook`; browser verification routes through
`module.workbook`, with the existing Assessment consumer under
`module.assessments`. Source policy verification belongs to `web.architecture`.
Executable inputs consume typed package facades; no executable Markdown dependency
is introduced.

## Workstream exits

| Workstream | Status | Evidence |
| --- | --- | --- |
| Orientation and characterization | Complete | Current source/contracts/owners inspected; fresh baseline frontend slice 6/6 units, zero cache hits at `.cartulary/test-results/20261001T235132Z-p64708/run-summary.json`. |
| Baseline rendered review | Complete | Seeded run `20261002T000225Z-p69157`, operations 7–34. Timeline text-only scalar/set authoring and candidate text-only staging reproduced. Stop receipt reports cleanup complete; foreground exited 0. |
| Shared implementation and regression | Complete | Both equality callers migrated; exact enum set draft storage and four contract-aware reopening callers implemented. Final five frontend rows passed 6/6 units at `20261002T002505Z-p2369`. |
| Production browser proof and rendered review | Complete | Seven Workbook browser rows and saved-query/text-spacing supplements pass. Sealed captures inspected at both viewports and 125% zoom, then spacing attachments imported and inspected. All three sessions cleaned up. |
| Terminal verification and acceptance | Complete | Final format/finalize, types, lint, boundaries, source policy, generation checks, Markdown lint and scope review pass. All applicable acceptance rows below pass. |
| Requested visual-golden follow-up | Complete | Initial comparison `20261002T005458Z-p19980` failed for prior date-validation changes. Authorized refresh `20261002T015412Z-p84397` passed; two fresh ordinary runs `20261002T020438Z-p31843` and `20261002T020438Z-p31845` passed 44/44 visual rows and all 255 comparisons against the final manifest. |
| Secondary cleanup diagnosis | Complete | Focused broker reproduction `20261002T015243Z-p81030` confirms synchronous quarantine's undefined result invokes destroy too. Harness repair is identified separately; no harness source change retained. |

## Characterization

Fresh Timeline `disable` Apply returned HTTP 200 with no matching rows.
`Disabled` Apply returned HTTP 200 with matching fixture rows; its applied
chip preserved case. The mixed set `disabled, disable, Disabled` returned
HTTP 200 and retained all distinct operands. Counts are observations of this
fixture, not a product assertion. Candidate enum edits and Add filter left
the accepted 66-row candidate page unchanged and selected-support draft at zero;
discovery belongs to Apply candidate query.

The first focused characterization run failed because its assertion guessed
locale canonical order incorrectly; the implementation retained all case variants.
Run: `.cartulary/test-results/20261002T001107Z-p10038/run-summary.json`.
The assertion now checks exact distinct membership and unchanged round-trip.
Final verification must supersede this development failure.

## Implementation boundary

The real common decision is projecting ordered declared enum choices and explicit
custom literals onto one equality FilterDraft. A shared pure model and native
select/checkbox presentation replace text-only defaults in both consumers.
Null remains a separate matching kind. Non-enum and non-equality authoring retain
their existing controls. Actual `filterOps` remain authoritative, including
enum prefix overrides; Capture State stays text.

Enum set reopening retains literal arrays, including comma-containing strings.
Scalar and singleton-set shapes remain distinct. Existing trimming, deduplication,
empty-set refusal and server canonicalization still apply at admission.
No membership rejection or case folding is added. Literal disclosure is local
presentation state; mounted controls participate in the Workbook focus registry.

Workbook Apply retains admission, close/focus return, then completion.
Candidate Add filter stages locally; Apply candidate query invokes existing
discovery. Supporting-record selection and final Assessment submission remain
separate explicit actions.

Four `filterDraftFromFilter` callers in WorkbookGridControls migrate together.
WorkbookCandidateQueryControl also serves ordinary reference pickers.
Missing/empty enum metadata retains text authoring. Shared facade imports, source
ownership registration and local source guides remain under Workbook ownership.

No data migration is required. Rollback consists of reverting this slice and
regenerating derived verification routing. Backend, route, schema, saved-view wire,
cursor, storage, theme and grid-vendor integration remain unchanged.

## Source and authority inspection

Inspected current Workbook Filters, candidate query control, GridControls, shared
control styles, workbookQuery, workbookGridQueryControls, useWorkbookQueryController,
AssessmentDiscovery, WorkbookQueryBrowser, relevant component/model/query/Assessment
source guides, public View/UI contracts, Timeline JSON contract, and server
viewquery/queryengine/Timeline compatibility. Source ownership and generated policy
were checked against authored JSON. The 19-field/11-view inventory describes current
metadata; the implementation derives choices without a surface allowlist.

The two new source files are registered in `tools/frontend_source_ownership.json`.
Component/model guides describe the shared boundary. Focused tests remain in the
existing component, query model and Assessment discovery files. Existing model and
query-controller suites are rerun without unrelated edits.

## Verification

Public task surface and both requested owner guides were rediscovered with
`make help`, `make help-all`, and `make task-guide ROLE=module-author OWNER=...`.
New browser IDs were derived using `make author-test-row-id` with the authored
family/claim/selector interface. `tools/test_families/web.workbook.json` routes
new unit titles through existing rows. `tools/test_families/module.workbook.json`
routes the two new Chromium scenarios through the existing default runtime and
matching stateful/functional profiles. `make generate` regenerated the browser
batch manifest and execution topology render index. No generated outputs were
hand-edited.

All slice executions below use `CARTULARY_HARNESS_CACHE_MODE=off`. Unit counts
include prerequisites; browser test counts come from current Playwright reports.
Run roots below are under `.cartulary/test-results/`; each graph result has a
`run-summary.json`, and browser groups carry `playwright-report.json`.

| Command / selection | Result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook ROWS=...` — the five requested rows (grid controls component/model, query model, query controller, Assessment discovery) | PASS, 6/6 units, zero cache hits | `20261002T002505Z-p2369` |
| `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | PASS, 2/2 units | `20261002T002505Z-p2350` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=...` — two new rows plus authoring_candidates_assessments, query_continuation_canonical, query_recovery_focus, date_filter_drafts_remain_locally_correctable_be_15ce5a6e7c, filter_editor_keyboard_focus | PASS, 14/14 units, 9/9 Chromium scenarios, zero cache hits | `20261002T002505Z-p2411` |
| `make service-backed-test-slice OWNER=module.assessments ROWS=module.assessments.browser.the_browser_workbook_drives_the_assessments_surf_4f277414a9` | PASS, 11/11 units | `20261002T002505Z-p2447` |
| `make format` — inspected authored diff | PASS, 2/2; no unrelated changes | `20261002T003818Z-p57666` |
| `make agent-finalize` before final checks | PASS, 1/1; generated artifacts unchanged | `20261002T003846Z-p62193` |
| `make frontend-typecheck` | PASS, 2/2 | `20261002T003947Z-p67278` |
| `make lint-biome` | PASS, 2/2 | `20261002T003947Z-p67295` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `20261002T003947Z-p67290` |
| `make generate-drift` | PASS, 4/4 | `20261002T002631Z-p80175` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `20261002T002632Z-p80378` |
| `make json-shape-check` | PASS, 3/3 | `20261002T002633Z-p80667` |

Supplemental final-source runs:

| Command / selection | Result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookquery_suite_904073db6c CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 2/2 units; custom saved-query arrays round-trip exactly | `20261002T003521Z-p19987` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=...` — the two new enum rows, including 768×640 text spacing and long literals | PASS, 13/13 units, 2/2 Chromium scenarios, zero cache hits | `20261002T003947Z-p67116` |
| `make lint-markdown` | PASS; repeated after final handoff edits | `20261002T004447Z-p7993` final handoff pass; earlier pass `20261002T003948Z-p67795` |
| `git diff --check` and working-tree/scope review | PASS; only listed slice files modified, no unrelated work | Local final review |

Exact new rows are:

- `module.workbook.browser_stateful.enum_equality_choices_remain_explicit_and_preser_28bc32bc06`
- `module.workbook.browser.assessment_timeline_support_enum_filtering_prese_9700791fe7`

The five frontend rows are:

- `web.workbook.regression.grid_controls_component_a104000002`
- `web.workbook.regression.grid_controls_model_a104000001`
- `web.workbook.regression.workbookquery_suite_904073db6c`
- `web.workbook.regression.useworkbookquerycontroller_suite_203a4e98cc`
- `web.workbook.regression.assessment_discovery`

The broader Workbook browser command selects the two new rows plus:

- `module.workbook.browser.authoring_candidates_assessments`
- `module.workbook.browser_stateful.query_continuation_canonical`
- `module.workbook.browser_stateful.query_recovery_focus`
- `module.workbook.browser_stateful.date_filter_drafts_remain_locally_correctable_be_15ce5a6e7c`
- `module.workbook.browser.filter_editor_keyboard_focus`

These exact IDs are supplied comma-separated through `ROWS`; owner IDs and
cache-mode settings are shown above.

`RESULTS_DIR` remained unset: retained-run maintenance, performance-baseline work
and retained full-run checks were skipped because no qualifying successful full
warm check was available. Broad backend, release and performance suites were not
selected: no backend, storage, route, schema or performance mechanism changed.
At the initial implementation handoff, no goldens had been refreshed. The
subsequent authorized refresh is recorded below. The interactive HTML review
report was unnecessary; images and relevant observations were inspected directly.

## Requested visual-golden comparison follow-up

On the user's subsequent request, ran `make agent-finalize` before the ordinary
comparison. It passed 1/1 units at `20261002T005429Z-p15905`; `RESULTS_DIR`
remained unset. Then ran
`make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off` against the unchanged
implementation source. Run `20261002T005458Z-p19980` failed: 10/12 execution
units passed, all 12 bypassed cache, and 43/44 visual catalog rows passed.
The primary classification is `product / test_assertion_failure`.

The failed row is
`module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`,
scenario `scenario_5974e42bb8fb`. Its one failed screenshot is
`workbook-view-bar-filter-editing-overflow-linux.png`, at 1440×900, compact
density and 100% zoom. The comparison reports 22,332 differing pixels (ratio
0.02). The scenario's other assertions and all other visual rows passed.
The reconciliation accounts for all 255 active capture intents and 255 committed
goldens, with zero orphans, missing goldens, ambiguous mappings or unresolved
registered fixtures. Its overall status remains failed because the visual
attempt failed; this is not a successful visual gate.

Expected, actual and diff images were imported by exact canonical capture ID
`visual.capture.9dece7e0dfcaa9014a74` through `cartulary-ui-review` artifacts mode,
run `20261002T010017Z-p56486`, bundle 1. All image bytes and SHA-256 values were
verified before native-resolution inspection. The actual popover is anchored to
the trigger's right edge and sits left of the expected popover; its always-mounted
validation status also adds spacing before Applied filters. Source history
attributes both changes to the already-landed date-validation commit
`ab658d413fb6e62d7719b2e40b53fbe7f5649016`: the `filterPopoverStyle` inline anchor
changed and the validation paragraph became permanently mounted. This golden was
last changed in `c2d04ecd9ec07dd97131057658ba51f621a07ba6`, before that commit.
The captured operand is Timeline Tags `contains_any`; the enum editor is not
rendered, and this slice leaves that operand and those styles unchanged.
No separate baseline visual execution was performed, so the historical attribution
is source/image analysis rather than a measured baseline comparison.

Durable diagnostics under `.cartulary/test-results/20261002T005458Z-p19980/`:

- `run-summary.json` and `target-summaries/browser-e2e-visual.json`.
- `browser-e2e-visual/frontend-visual-reconciliation.json`.
- `browser-e2e-visual/browser-groups/visual-workbook-visual/playwright-report.json`
  and that group's `playwright-output/` expected, actual and diff attachments.
- `_shared/fixture-leases/lease-000002.json` and `unit-events.ndjson` event 93:
  secondary harness `cleanup_error`, with the failed Workbook browser lease
  recorded as quarantined and cleanup failed. Both browser process-stop and
  runtime-root removal timing spans passed; shared-service lifecycle ended
  `cleanup_succeeded`. Its cause was not yet established at comparison handoff;
  the subsequent root-cause investigation below explains it. Ownership evidence
  was preserved; no runtime roots or leases were manually removed.

The artifact-review terminal receipt reports `closed`, status `ok`, cleanup
`complete`, with foreground exit 0. Caller-owned request scratch was removed.
No DOM or Axe channels were available for the imported images. The ordinary run
changed no committed golden bytes or golden manifest. A golden refresh, new
enum golden fixtures, a baseline rerun and harness cleanup repair were not part
of the requested comparison and were not performed.
Follow-up handoff `make lint-markdown` passed at `20261002T010819Z-p60966`;
`git diff --check` and scope review passed. Only this handoff was edited during
the follow-up; the implementation and committed goldens remain unchanged.

## Cleanup root cause and requested golden refresh

The user subsequently authorized refreshing the goldens and finding the secondary
cleanup error's root cause. The accepted refresh trigger is a stale golden relative
to already-validated date-filter behavior, as documented in the prior ordinary
comparison and native-resolution review above. Its affected owner row is
`module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`;
the exact active capture is `visual.capture.9dece7e0dfcaa9014a74`. It has no
registered stable fixture IDs, which the reconciliation explicitly permits for
this active nonregistry golden. Viewport, zoom, density, masks, scroll normalization,
screenshot scope, renderer and fonts remain unchanged.

The cleanup root cause is return-value-based fallback in
`tools/harness/scheduler/fixture-broker/index.mjs`:
`await (lease.allocation.quarantine?.() ?? lease.allocation.destroy?.())`.
The governing mechanics are Testing Harness NLSpec §2.1,
TH-HARNESS-REQ-804 (owned fixture leases) and TH-HARNESS-REQ-806 (failed-lane
quarantine and cleanup; overlapping ordinary visual work).
When a browser group fails, the scheduler releases its stack as unhealthy. The
browser provider assigns the same synchronous `close` callback to `quarantine`
and `destroy`; `close` returns `undefined`. The first call successfully stops
owned processes, removes the runtime and deletes the private session lease.
Because its return value is nullish, the expression calls `destroy` too. The
second call reaches `terminateBrowserStackLease` and `lstatSync` on that already
deleted lease, throwing before another stop command runs. The broker then records
`quarantined / failed / cleanup_error` after successful physical cleanup.
This explains the original run's successful teardown spans and service cleanup
alongside its failed fixture record. The original failure artifacts remain intact.

A temporary diagnostic in the existing fixture smoke assertion reproduced this
with the production `FixtureBroker`: an owned browser allocation exposed both
methods as one synchronous callback; the first call consumed its simulated lease,
the second threw for the missing lease, and the retained record became
`cleanup_error`. `make run-harness-smoke-fast` passed all three selected smoke
checks at `20261002T015243Z-p81030`, including explicit assertions of two callback
invocations, successful first cleanup and failed cleanup accounting. Its
`harness-smoke-fixture-broker-smoke/harness-smoke-fixture-broker-smoke/stdout.log`
contains the structural `ROOT_CAUSE_REPRO` confirmation. This is diagnostic
reproduction evidence, not evidence of a corrected harness. The diagnostic was
removed immediately afterward and `git diff --exit-code` verified the test file
was restored exactly. Existing smoke providers use async callbacks, whose Promise
return values prevent the fallback and explain the coverage gap.

Harness implementation remains unchanged. A targeted repair should select the
available `quarantine` or `destroy` method before invoking it, retaining its
receiver, and add regression coverage for synchronous void-returning callbacks,
destroy-only allocations and genuine cleanup exceptions. This request asked for
root-cause diagnosis; the repair is a separate, identified harness change.

Before the refresh, `make agent-finalize` passed 1/1 units at
`20261002T015101Z-p76147`, with `RESULTS_DIR` unset. The canonical transactional
`make browser-e2e-visual-update CARTULARY_HARNESS_CACHE_MODE=off` passed 12/12
units and 44/44 visual rows at `20261002T015412Z-p84397`, with all units bypassing
cache. Reconciliation passed for all 255 goldens and both browser fixture leases
ended `released / completed` with no cleanup failure. It promoted exactly one
changed PNG, `workbook-view-bar-filter-editing-overflow-linux.png`, plus its entry
in `tools/frontend_visual_golden_manifest.json`. The golden SHA-256 changed from
`22cd769b887bf4075ecd5e07ea4a9bda53f465c8c82cc256082342973601779a` to
`9c6ab774fd3d95456caea9efb22cf33c269c052a3f4dde5c0b721527dc375692`.
The final manifest SHA-256 is
`b49d67f9999b53358550811906f75a7e896c68db3adddb206780da3495bdd8c1`.

Promoted-image review used artifacts session `20261002T020234Z-p22493`, bundle 2.
The first canonical import attempt failed with `artifact_error / unsafe_artifact`
(exit 11): that importer accepts ordinary `browser-e2e-visual` target results,
whereas this was an update-target run. The exact promoted PNG was then imported
as an explicit image reference, with its `reference_only`, no DOM/Axe/trace
limitations retained. Bytes and SHA-256 were verified, and it is byte-identical
to the previously reviewed ordinary-run actual image. Native-resolution inspection
confirmed the expected inline anchor and validation spacing, with no additional
pixel changes. The review's terminal receipt reports `closed / ok`, cleanup
`complete`; foreground exited 0 and caller request scratch was removed.

After image review, `make agent-finalize` passed 1/1 units at
`20261002T020254Z-p22632`. `RESULTS_DIR` remained unset because no qualifying
successful full warm check was available. Terminal policy checks:

| Command | Result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 1/1 unit | `20261002T020303Z-p23245` |
| `make json-shape-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 3/3 units | `20261002T020438Z-p31442` |
| `make generated-artifact-policy-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 3/3 units | `20261002T020438Z-p31439` |
| `make generate-drift CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 4/4 units | `20261002T020438Z-p31455` |

The fixture-registry mapping scenario is included in all 44-row visual runs;
the registry itself is unchanged. Two fresh ordinary comparisons were started
against the same final manifest as `20261002T020438Z-p31843` and
`20261002T020438Z-p31845`, with `CARTULARY_HARNESS_CACHE_MODE=off`. They use
separate isolated stacks; no source, routing, manifest or golden edits occur
during these runs. Both passed 12/12 units, 44/44 visual rows and all 255 golden
comparisons, with all units bypassing cache. Both reconciliations passed with zero
missing, orphaned or ambiguous goldens. Each reconciliation recorded the same
final manifest SHA-256 shown above. All four browser fixture leases ended
`released / completed` with no cleanup failure. Each run's `run-summary.json`,
`target-summaries/browser-e2e-visual.json`,
`browser-e2e-visual/frontend-visual-reconciliation.json` and
`_shared/fixture-leases/` retain the terminal evidence.

One command-discovery attempt, `make explain-target
TARGET=harness-smoke-fixture-broker-matrix DETAIL=summary`, failed with
`config / usage_error`: that name is a smoke check, not a Make target. The
declared Make wrapper `run-harness-smoke-fast` was then used for the diagnostic.
This navigation failure does not affect the visual results. Follow-up handoff
Markdown lint passed at `20261002T020613Z-p6647` and was repeated after this
final update. Final `git diff --check` and scope review passed.

This follow-up changes only the one golden, its machine-generated manifest hash
and this controlling record. It introduces no harness repair, enum golden fixture,
renderer/profile change, data migration or expanded product audit. Broad backend,
release and performance suites remain unnecessary for this follow-up.

## Browser and rendered evidence

The Timeline browser journey uses an isolated closed incident: native keyboard
selection produces `{value:"disabled"}`, checkboxes produce the existing
`{values:[...]}` shape, custom/case/comma members remain in real-route requests,
and canonical metadata preserves the server's authoritative normalization.
A genuine aborted read retains accepted rows/chips; Revert and Retry recover
through the existing owner. Parties proves contract reuse. Case-variant scalar
`Disabled` remains accepted and usable without a client enum-membership check.

The Assessment browser journey stages choices with zero discovery requests until
Apply candidate query. Selected IDs survive paging, exclusion and a genuine failed
read/retry. Cancellation preserves the Assessment draft; support confirmation is
separate and no Assessment submission occurs. Unit tests additionally check
independent accepted/staged identities and saved/custom arrays.

Fresh seeded final review: `20261002T002421Z-p76884`. Captures 7, 18, 21, 24,
28, 33, 38, 59, 63, 66 and 69 were inspected after bundle byte/digest validation.
Representative enum evidence: 18 (1440×900 multiple/custom), 24 (768×640 reachable
Cancel/Apply after keyboard scrolling), 38 (125% zoom at 1440×900), 59 (Assessment
multiple/custom at 1440×900), and 63/66/69 (768×640 candidate application,
selection cancellation and full declared choices). Long inputs retain native
horizontal text scrolling; long declared tokens fit or wrap. Parent scroll regions
keep controls reachable without adding a second scrolling owner.

At 768×640, moving keyboard focus scrolls the enlarged popover to its actions.
Applying 125% zoom at 768×640 crosses the existing below-minimum chrome threshold;
that capture is diagnostic only. Supported zoom was reviewed at 1440×900.
Axe completed with zero violations in inspected captures, with 17 incomplete
observations across the final session. Incomplete checks remain unproven; captures
and scans are supporting evidence, not accessibility certification.

Both seeded terminal receipts report `cleanup: complete`, status `ok`; both
foreground processes exited 0. Final review recorded zero console errors and zero
failed requests. The baseline's console observations were not a product pass.
Private images, observations, credentials and scratch paths are not committed and
links expire on cleanup.

Spacing attachments from the successful final-source run `20261002T003947Z-p67116`
were imported into artifact-review session `20261002T003523Z-p20905`, captures 2/3.
Both originals were byte/digest checked and inspected. At 768×640 with the existing
supported spacing profile (line-height 1.5, letter spacing 0.12em, word spacing
0.16em, paragraph margin 2em), Workbook Apply/Cancel and candidate Apply remain
visible when focused; literal inputs stay contained and retain exact text.
Imported images supply no additional DOM/Axe channels. The artifact session also
reports cleanup complete and foreground exit 0. Only this task's private
request/image scratch directory was removed after inspection.

## Development failures and limitations

- Query characterization run `20261002T001107Z-p10038` failed a guessed locale-order
  expectation; exact case-distinct membership and unchanged round-trip now pass.
- Component run `20261002T001629Z-p15835` failed test assertions for accessible-name
  casing and an unscoped label shared by two editors. Semantic/scoped assertions
  corrected these; the five-row final run passes.
- `make frontend-typecheck` at `20261002T002119Z-p22872` caught test-only RTL `exact`
  options and an unused fixture ID. Corrected; final typecheck passes.
- `make lint-biome` at `20261002T002044Z-p21727` and `20261002T002246Z-p69309`, and
  `make format` at `20261002T002146Z-p25812` and `20261002T002231Z-p60683`, caught
  formatting/imports and positional literal-slot keys. Formatting was repaired;
  the narrow file suppression documents that positions retain input identity while
  text changes, without introducing a second operand/selection store.
- Browser run `20261002T002157Z-p30295` passed the new Assessment scenario but failed
  Timeline's final reuse navigation because the test passed a display label to a
  schema-ID helper. The corrected real-route final run passes both scenarios.
- Supplement typecheck `20261002T003521Z-p20132` caught `Node.remove()` in test
  cleanup; typed parent removal now passes. Supplement browser run
  `20261002T003521Z-p20006` passed Assessment but timed out on a chip moved into
  the existing narrow-screen Filters overflow. The test now edits through that
  reachable popover; both final spacing scenarios pass at `20261002T003947Z-p67116`.
- A review snapshot used an epoch before a pending authentication operation
  completed and was rejected with `session_mismatch`. The completed operation was
  observed, then a fresh snapshot proceeded; no mutation was blindly replayed.

No known product verification failure remains from these development runs.

## Digest acceptance

Assessments are limited to this slice's adopted obligations. N/A does not claim
an unchanged subsystem's broader acceptance matrix passed.

| Row | Result | Evidence / scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Exact owner map above; typed metadata guides choices without becoming query-validation authority. |
| A002 Scope | PASS | One shared choice/custom projection replaces both text-only enum defaults. Compatibility is required by current supported query operands. Future enum fields reuse active metadata; no speculative framework. |
| A003 Repository state | PASS | Clean synchronized baseline; current guides, imports, ownership and generated policy inspected; only bounded authored source/tests/routing/docs changed. |
| A004 Tokens | PASS | New spacing, borders and controls use current tokens/styles and native inputs; no theme or token registry. |
| A005 Theme | N/A | Theme selection and fixtures unchanged; reviewed existing dark_graphite renderer. |
| A006 Density | N/A | No row/header, density selection or full-cell editor geometry changed. |
| A007 Creation | N/A | Creation payloads/minima/capabilities unchanged. New support journey proves filtering never submits Assessment. |
| A008 Responsive | PASS | New controls inspected at 1440×900 and 768×640; 125% zoom within supported width. Existing thresholds and inspector clamp remain untouched. |
| A009 Overflow | PASS | Existing menu/inspector scroll ownership preserved; focused actions and safe chrome reachable in captures. |
| A010 Inspector | N/A | Feature dispatcher, route admission and confirmation lifetime unchanged. |
| A011 Continuity | PASS | Component/model/browser evidence for exact draft shapes, support identities across paging/filtering/failure/cancel and existing canonical query continuation. |
| A012 Transactions | N/A | No mutation dispatch, transaction identity or replay path changed. |
| A013 Acknowledgement and recovery | N/A | No write acknowledgement or mutation recovery changed; query Retry/Revert covered under A016. |
| A014 Editing | PASS | Choice edits remain drafts; literal slots retain focus; explicit admission, empty-set refusal, Escape/Cancel and reopened operands tested. |
| A015 Conflict | N/A | Cell mutation/conflict presentation untouched. |
| A016 Query data and interaction | PASS | Closed-incident filtering and failed-read accepted-row/chip retention exercised through real query owners; no mutation-state gating added. |
| A017 Refresh and authorization scope | PASS | Failed reads retain authorized rows/selection. Existing authorization cleanup remains with unchanged query/discovery owners; hook suites and diff reviewed, no new caching or retained labels. Wider account/incident matrix is outside this slice. |
| A018 Evidence | N/A | Evidence lifecycle/preview/upload behavior untouched. |
| A019 Accessibility | PASS | Native keyboard choice and checkbox activation, complete mounted Tab registry, focus return, associated validation and recovery tested. Manual focus review plus zero Axe violations supports the scoped result; incomplete scans are not certification. |
| A020 Components | PASS | Long tokens/custom operands, native input scrolling and supported zoom reviewed. Text-spacing geometry checks pass; both final spacing attachments inspected through the artifact harness. |
| A021 Virtualization | N/A | Grid adapter, virtualization and row geometry unchanged. |
| A022 Visual fixtures | PASS | Fresh enum rendering captures inspected. One stale Tags popover golden was intentionally refreshed and reviewed under the maintenance guide. Two fresh ordinary runs pass all 44 visual rows and 255 comparisons against the same manifest; no claim-publication boundary invoked. |
| A023 Selectors | PASS | Existing public test IDs, schema/field/record IDs and native role/name selectors; no positional surface allowlist. |
| A024 Test authority | PASS | New executable code/tests/catalog consume typed facades and JSON only. Docs guide human review and handoff; no executable Markdown read/stat/hash dependency added. |
| A025 Generated artifacts | PASS | Authored catalog/family inputs precede Make generation; drift, policy and shape checks pass. |
| A026 Authority and compatibility | PASS | No enum-membership reject/case fold, route/storage change or data migration. Exact custom shapes retained; rollback and retirement stated above. |
| A027 Handoff | PASS | This controlling record closes all exits, with exact owners/callers, current commands/artifacts, failures, limitations, cleanup, no migration and rollback. Markdown lint and diff/scope review pass. |

## Subsequent harness remediation

The false secondary cleanup failure observed during visual comparison is handled
in [Fixture cleanup lifecycle remediation](harness-fixture-cleanup-remediation.md).
That separate bounded record contains the lifecycle repair, current verification
and recovery evidence. The Workbook implementation, refreshed golden and
historical evidence above remain preserved.

