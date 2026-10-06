# Workbook boolean filter authoring handoff

## Baseline and authority

This is one authorized frontend correction, not a new UI audit or normative amendment.
Work began on clean `main` at `2cbb8473eaae425e4b73421acc4129a10efb088b`, matching
local `origin/main`. The supplied review baseline was `a445a4ae6034ebc8a84fec064bef7a483b315c1f`;
the affected filter seam was unchanged between those commits. Unrelated work was
absent. The final checkout remains uncommitted for review.

Root `AGENTS.md`, applicable local source guides, generated-artifact policy,
frontend ownership/import inputs, and both module-author task guides were read.
Source belongs to `web.workbook`; browser verification belongs separately to
`module.workbook`. `web.architecture` verifies source ownership. Contracts are typed
projections, not substitutes for adopted owners.

| Adopted owner | Application in this slice |
| --- | --- |
| Core 01 §3.3.4.1, REQ-01-039–046 | Declared fields/operators; exact scalar, null and nonempty logical-set operand shapes; strict rejection and server-owned normalization/canonical metadata. |
| Core 03 REQ-03-220/223/224 and §14.9 | Native keyboard behavior, declared capabilities, explicit query application, canonical results and existing request fencing. |
| Core 03 REQ-03-286/299/100 | Data versus interaction state and scoped authorization lifetime remain with their current owners. No write authority added to read filtering. |
| Design §§8.3–8.5 and applicable §§12/14 | Local feedback, accepted chips, native controls, focus return and compact containment. |
| Domain | Vocabulary and owner navigation only. |

The requested digest README, START_HERE, LOCAL_AGENT_PROMPT, maps, query recipes,
rules and acceptance table, and prior enum/date handoffs informed human review.
Research and embedded upstream prompts supplied no additional authority or
permission. No contradiction was found and no specification changed.

## Boundary and implementation

The common decision is active contract type → authoring controls → admission →
typed query operand → reopening, within the existing `FilterDraft` and query model.
The remediation rubric selects this seam because both filter editors shared the
same observable failures but duplicated presentation and relied on incidental
operand types. A small pure boolean helper and shared native presentation hide that
real decision. Future declared boolean equality fields inherit it without another
field registry. Remaining risk is limited to query draft/presentation behavior;
state and authorization owners are retained.

- `models/workbookQuery.ts`: initializes and reopens from `readKind`, declared
  filterability and actual `filterOps`; routes boolean feedback/admission through
  one strict decision. Rejection preserves the query object and authored state.
- New `models/workbookBooleanFilterOperand.ts`: typed unset/scalar/set draft values,
  strict restoration, malformed-argument evidence, and mounted control keys. It
  has no import back into the query module. Duplicate members remain for server
  normalization; a both-value set remains a predicate.
- New `components/WorkbookBooleanFilterOperand.tsx`: explicit unchosen native
  scalar select and true/false checkboxes. False is a selected value, never absence.
  Null remains a separate mode. No custom boolean literals or query application.
- `WorkbookFiltersControl` and `WorkbookCandidateQueryControl` use that same
  contract-driven presentation and validation. Workbook closes only on admitted
  Apply. Candidate Add stages locally; Apply candidate query performs discovery.
  Selection acceptance and final creation remain separate.
- Boolean operand removal returns owned focus to matching mode. Native arrows and
  Space remain native; both checkboxes participate in Tab/Shift+Tab. External focus
  is not reclaimed. Invalid feedback is associated with the relevant controls.
- Tests cover model/saved/request round trips, both components, controllers,
  Assessment discovery, ordinary references and real server queries. Existing
  fixtures consuming the old draft property were migrated together.
- Source ownership and the model/component guides register the new authored files.
  Authored test families add selectors and three browser rows; Make generated only
  the browser batch manifest and execution-topology render index.

Retired `filterInputMode`, its hard-coded field allowlist, the old `booleanValue`
string property/parser, operand-derived boolean reopening and boolean string-set
construction. No legitimate remaining caller needs these paths. Generic string,
number, date, tag and enum paths retain their existing semantics; date validation
and enum custom/comma/case/null/operator coverage remain in the focused slices.

No new query owner, retained-authoring store, API/schema change, coercion, dependency,
storage migration or saved-view format change is required. Valid `{value:true}`,
`{value:false}`, `{values:[true]}`, `{values:[false]}`, `{values:[false,true]}` and
`{value:null}` retain their types and shapes. Malformed restored strings remain
invalid evidence until explicit replacement; they are not silently converted.

Rollback: revert this frontend slice and its authored tests/routing/ownership,
then regenerate derived artifacts through Make. No data rollback is required.

## Characterization and development results

All run roots below are under `.cartulary/test-results/` unless stated otherwise.
Final product and policy checks pass; no applicable acceptance blocker remains.

- Baseline six-unit frontend slice passed at `20261006T222703Z-p89056`; that old
  coverage did not disprove the reported failures.
- Four new characterization rows failed at `20261006T223225Z-p21138` before
  production edits: non-Timeline scalar initialization remained string; accepted
  sets/null lacked typed reopening; Assessment lacked boolean checkboxes; ordinary
  reference filtering lacked the boolean select. Other assertions remained intact.
- Seeded baseline `20261006T223240Z-p22534` reproduced HTTP 400 for Timeline
  Has Evidence set text `true` and Task Requests No Owner scalar text `true`.
  Apply closed the editor and left accepted rows stale with Retry/Revert. Capture
  bundle 23 was byte/SHA-256 verified and viewed directly at 1440×900. The terminal
  receipt reported cleanup complete; foreground exited 0.
- An earlier review startup `20261006T223115Z-p91899` rejected source drift during
  its build seal. It was stopped using its exact locator; cleanup complete was
  verified. Its foreground exited 2. Source was held stable for the next session.
- Intermediate frontend run `20261006T224139Z-p63735` passed the new characterizations
  but caught an overly broad focus fallback affecting a nonboolean field. The
  fallback was restricted to booleans; subsequent slices pass.
- Typecheck `20261006T224217Z-p64959` caught test-only React Testing Library `exact`
  options. Corrected to its supported exact string names.
- Catalog selectors initially lacked required ASCII sorting. Test slice
  `20261006T224411Z-p65896` and format `20261006T224628Z-p67496` failed before producing
  summaries with `artifact_error`. `make target-plan TARGET=format` identified the
  authored selector ordering. Generate `20261006T224639Z-p67705` also failed for
  this input; sorted selectors and subsequent Make generation resolved it.
- Lint `20261006T225053Z-p41479` rejected two test-only non-null assertions.
  An explicit declared-field fixture check replaced them; final lint passes.
- Browser run `20261006T224750Z-p79584` passed 9/10 selected scenarios. The ordinary
  reference test expected an incorrect empty-state phrase; the rendered state
  correctly retained the selected identity and reported no matching candidates.
  The assertion was corrected. No product workaround or golden update was used.

## Verification commands and evidence

Public discovery used `make help`, `make help-all`, and:

```sh
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=module.workbook
```

New IDs were produced by `make author-test-row-id` with these exact inputs:

| FAMILY_ID | CLAIM | SELECTOR_KEY |
| --- | --- | --- |
| module.workbook.browser_stateful | Boolean filters preserve typed saved and accepted operands across matching modes | workbook_boolean_filter_roundtrip |
| module.workbook.browser | Boolean support filters stage typed queries and retain selected identities across pages | assessment_boolean_support_filters |
| module.workbook.browser | Ordinary reference boolean filtering preserves the selected source until explicit acceptance | ordinary_boolean_reference_filters |

The final focused commands are:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=web.workbook ROWS=web.workbook.regression.grid_controls_component_a104000002,web.workbook.regression.grid_controls_model_a104000001,web.workbook.regression.workbookquery_suite_904073db6c,web.workbook.regression.useworkbookquerycontroller_suite_203a4e98cc,web.workbook.regression.assessment_discovery,web.workbook.regression.authoring_candidate_presentation
CARTULARY_HARNESS_CACHE_MODE=off make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.boolean_filters_preserve_typed_saved_and_accepte_221721fc9a,module.workbook.browser.boolean_support_filters_stage_typed_queries_and_7785af35c2,module.workbook.browser.ordinary_reference_boolean_filtering_preserves_t_b9997d12f7,module.workbook.browser.filter_editor_keyboard_focus,module.workbook.browser_stateful.query_continuation_canonical,module.workbook.browser_stateful.query_recovery_focus,module.workbook.browser.authoring_candidates_assessments,module.workbook.browser_stateful.enum_equality_choices_remain_explicit_and_preser_28bc32bc06,module.workbook.browser.assessment_timeline_support_enum_filtering_prese_9700791fe7,module.workbook.browser_stateful.date_filter_drafts_remain_locally_correctable_be_15ce5a6e7c
make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19
make format
make generate
make agent-finalize
make frontend-typecheck lint-biome frontend-import-boundary-check
make generate-drift generated-artifact-policy-check json-shape-check
make lint-markdown
git diff --check
```

| Verification | Result / run root |
| --- | --- |
| Six frontend rows | PASS 7/7 units, `20261006T225053Z-p41295`; earlier correction pass `20261006T224750Z-p79579`. |
| Source ownership | PASS 2/2 units, `20261006T225053Z-p41305`. |
| Format | PASS, final `20261006T225257Z-p94049`; earlier `20261006T225027Z-p35871`, `20261006T224657Z-p71246`, `20261006T224857Z-p25905`. Diff contained only this slice. |
| Generate | PASS, `20261006T224718Z-p76397`. |
| Agent finalize | PASS, `20261006T224918Z-p31219`, before broader terminal checks. |
| Ten browser rows | PASS 14/14 units, `20261006T225053Z-p41314`; all selected scenarios passed. |
| Final model recheck after lint-only test correction | PASS 2/2 units, `20261006T225309Z-p99396`. |
| Frontend typecheck | PASS, `20261006T225308Z-p99187`; preceding pass `20261006T225053Z-p41442`. |
| Biome lint | PASS, `20261006T225308Z-p99197`. |
| Frontend import boundaries | PASS, `20261006T225308Z-p99191`. |
| Generate drift | PASS 4/4 units, `20261006T225103Z-p48641`. |
| Generated-artifact policy | PASS 3/3 units, `20261006T225103Z-p48658`. |
| JSON shape check | PASS 3/3 units, `20261006T225103Z-p48669`. |
| Markdown / final diff | PASS `20261006T230215Z-p56545` and final confirmation `20261006T230327Z-p58990` (both `adhoc/lint-markdown/tool-run-summary.json`); `git diff --check` clean. |

`RESULTS_DIR` was unset: retained-run maintenance was skipped because no qualifying
successful full warm run was supplied. Broad backend/release/performance/visual
suites were not selected: no corresponding boundary or golden changed. Automated
product tests do not depend on Markdown. Documentation lint is maintenance only.

## Final rendered review

Fresh `make ui-review UI_MODE=seeded REVIEW_PROFILE=default` used isolated synthetic
fixtures and the seeded editor actor. Its exact locator was
`.cartulary/test-results/20261006T225319Z-p770/ui-review/session.json`.
Finite commands used `CARTULARY_OUTPUT_MODE=machine make ui-review-status`,
`make ui-browser`, `make ui-capture`, and `make ui-review-stop` with that exact
`UI_SESSION`; action/capture calls supplied task-owned private `UI_REQUEST` JSON.
No borrowed application data or raw browser automation was used. Source remained
fixed throughout the sealed review.

The bundle manifest for each capture supplied the original and observations paths.
Both components' byte counts and full SHA-256 digests were verified against that
manifest, and every original image was viewed before expiry:

| Bundle | Viewport | Direct observation |
| --- | --- | --- |
| 10 | 1440×900 | Unchosen Timeline scalar; associated local feedback and disabled Apply; accepted 66 rows unchanged. |
| 17 | 1440×900 | Has Evidence set true selected; native Space and Tab place visible focus on false; no query from selection. |
| 27 | 768×640 | Reopened accepted true set, changed to false; Apply, Cancel, accepted filter and Recovery remain reachable. |
| 35 | 768×640 | Reopened set → null → scalar yields an unchosen select, without false inference; accepted false filter and 64 rows retained. |
| 49 | 1440×900 | Task Requests No Owner scalar true selected with native ArrowDown; visible focus and Apply. |
| 75 | 1440×900 | Assessment set choice with native keyboard focus on Add filter, selected support retained, query Apply separate. |
| 80 | 768×640 | Staged boolean filter; focused Apply candidate query visible within existing inspector scroll ownership. |
| 84 | 768×640 | Explicit filtered result has two candidates; off-page selected support remains; selection Apply/Cancel and final creation are separate and reachable. |
| 87 | 768×640 | Escape cancels staged selection, returns visible focus to Choose support, and leaves the parent draft unchanged. |

Actions 18/28 produced successful Timeline queries (true set → two rows, false
set → 64 rows), with accepted reopening confirmed in observations 21/34. Task
Requests Apply at 50 produced HTTP 200 and its canonical accepted chip. Candidate
choice and Add filter at 68–77 issued no extra read; explicit query Apply at 81
produced HTTP 200 and preserved the selected support identity. Exact wire-type
and canonical-argument assertions belong to the automated real-route tests above;
manual network observations are supplementary.

Axe completed in all nine captures with zero violations and 14 accumulated
incomplete observations. These scoped observations do not certify accessibility;
frames and off-rendered content were not assessed. The terminal recorded zero
console errors and 13 accumulated failed-request observations. Observed network
records showed a status-null Task Requests navigation read followed immediately
by a successful replacement before boolean authoring; all reviewed boolean Apply
requests returned HTTP 200, without query recovery. Counts are observation totals,
not an assertion of 13 distinct failures. No unrelated UI audit was started.

`ui-review/terminal.json` reports status `ok`, state `closed`, and cleanup `complete`.
Its 519 bytes and receipt SHA-256 were checked; stop and the original foreground
both exited 0. Only this task's private request scratch directory was removed.
Private captures/observations expired with cleanup and were not copied into the
repository. Baseline and final exact-session cleanup are both verified.

## Digest acceptance

Only scoped obligations are assessed; N/A does not claim unchanged subsystem
matrices passed. Applicable obligations are supported by the evidence above.

| Row | Result | Evidence / scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Adopted owner map above; contracts guide projection, routing supplies verification only. No normative change. |
| A002 Scope | PASS | One typed boolean decision and shared presentation, both consumers migrated; allowlist/parser/string-set path retired. Future declared booleans reuse metadata; no generic framework. |
| A003 Repository state | PASS | Clean advanced baseline revalidated; local guides, source/import policy, generated roots and separate verification owners inspected. Final scope reviewed. |
| A004 Tokens | PASS | Current input, spacing and border tokens/native controls; no design literal registry or theme owner. |
| A005 Theme | N/A | Theme selection unchanged; reviewed current dark_graphite renderer. |
| A006 Density | N/A | No density, row/header or full-cell editor geometry change. |
| A007 Creation | N/A | Creation capabilities, minima and payloads unchanged; candidate tests prove filtering does not create records. |
| A008 Responsive | PASS | Both editors inspected at 1440×900 and 768×640; existing thresholds/clamp owners retained; focused actions remain reachable. |
| A009 Overflow | PASS | Existing popover/inspector scroll ownership retained; captures show safe navigation, status, query Apply and selection actions. |
| A010 Inspector | N/A | Feature dispatch and confirmation lifetimes unchanged; candidate rendering borrows existing Inspector. |
| A011 Continuity | PASS | Exact saved/request/accepted shapes, failed requested reopening, support/reference identity retention across page replacement/failure/cancel, existing query continuation/recovery rows. |
| A012 Transactions | N/A | No transaction dispatch, replay, identity or mutation queue change. |
| A013 Acknowledgement and recovery | N/A | No write acknowledgement/recovery change; query Retry/Revert remains covered under A016. |
| A014 Editing | PASS | Unset/empty/malformed drafts refused locally; correction and callback refusal retain authoring; explicit Apply and Escape/Cancel tested. |
| A015 Conflict | N/A | Cell mutation conflicts untouched. |
| A016 Query data and interaction | PASS | Closed-incident boolean filtering succeeds; failed replacement keeps accepted chips/rows; no record-write gate introduced. |
| A017 Refresh and authorization scope | PASS | Failed reads preserve authorized rows/selections; current query/discovery generations and scoped authorization lifetimes unchanged. Wider account/incident matrix outside this slice. |
| A018 Evidence | N/A | Has Evidence is a read predicate; evidence lifecycle/upload/preview behavior unchanged. |
| A019 Accessibility | PASS | Scoped native select/checkbox activation, complete Tab traversal, owned focus reconciliation, associated feedback and return verified by component/browser/manual review. Zero Axe violations; incomplete checks remain unproven. |
| A020 Components | PASS | Scalar/set/null/unset/reopened/candidate variants at both sizes; existing enum long-literal/text-spacing and date scenarios pass. No new compound-state owner. |
| A021 Virtualization | N/A | Grid Adapter and virtualized row geometry untouched; existing continuity checks selected without expanding into performance work. |
| A022 Visual fixtures | PASS | Fresh production renderer sealed captures were verified and directly viewed. No relevant golden changed; broad visual suites unnecessary. These are implementation-support observations only. |
| A023 Selectors | PASS | Existing semantic schema/field/record IDs, public test IDs and role/name selectors. No selector contract changed; package.ui suite unnecessary for this slice. |
| A024 Test authority | PASS | Tests and runtime consume typed contracts/JSON only; no Markdown-dependent executable evidence added. |
| A025 Generated artifacts | PASS | Authored family/ownership inputs precede Make generation; drift/policy/shape checks pass. Only expected derived routing changed. |
| A026 Authority and compatibility | PASS | Exact valid JSON shapes and server normalization retained. No API/storage change or migration; malformed restoration remains explicit. Rollback documented. |
| A027 Handoff | PASS | Owners, scope, compatibility, commands/run roots, failures, rendered observations, limitations and verified cleanup recorded. Final documentation checks recorded below. |

## Repository handoff state

Final branch is `main`; HEAD and local `origin/main` remain
`2cbb8473eaae425e4b73421acc4129a10efb088b`. The working tree contains only this
frontend slice, tests, ownership/routing projections, source-guide entries and
this handoff. No commit or pull request was created. No active review session or
request scratch remains. There is no required follow-up implementation or
verification blocker; the limitations above describe the bounded evidence.
