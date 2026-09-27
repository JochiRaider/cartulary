# Workbook candidate read keyboard continuity

## Baseline, authority and scope

The checkout was clean on `main` at
`bd282072ae0d9ab422dc88cc8d9810bc2d7776d6` before editing and remains
on that commit with only the files listed below modified. The root `AGENTS.md`
was the applicable repository procedure; no nested instruction changed this
slice. The UI/UX digest, its prompts, and historical handoffs supplied navigation
and review criteria, not authorization. There was no contradiction among the
adopted owners inspected.

Core 01 §§3.3.1 and 7.4/7.4.1A and its Parties/reference declarations own the
query, view and reference contracts. Core 03 §2.3A owns independent authoring,
attachment, read and operation lifetimes; §16.4 owns contextual authoring and
source preservation. REQ-03-299/100 own incident/account authorization and
concealment; REQ-03-304 owns Assessment Timeline support selection.
`docs/design.md` §§12.4 and 14 direct stable loading names, state communication
and visible keyboard focus. `docs/domain.md` supplies vocabulary only.

Current manifests assign source ownership to `web.workbook`. The focused unit
rows route through `web.workbook`; the service-backed Chromium consumers route
through `module.workbook`. The authored test-family manifests own the generated
execution topology render index. The source ownership, import boundary,
verification registry, and generated-artifact policy manifests were inspected.
No specification, API, schema, storage, dependency or migration changed.

The boundary is the shared `WorkbookCandidateBrowsing` presentation used by
authoring references and Assessment discovery. Discovery still owns admission,
captured failed destinations, canonical query/cursor handling, bounded pages,
cancellation, authorization and stale-response fencing. Picker staging and parent
drafts still own selection, Apply/Cancel, raw fields, labels and versions. The
separate surface-inventory read is outside this slice.

## Reproduction, correction and compatibility

On the baseline, seeded Chromium at 1440×900 and 1024×720 opened a saved Timeline
row, Create task request, then Choose Requester Party. Enter on Refresh returned
HTTP 200, after which focus moved to the page body. That seed had one Party; it
was evidence for Refresh only. The baseline discovery/presentation route passed
3/3 work units without pending-control focus assertions.

Controlled read gates now characterize held First, Refresh, Next, Previous and
Retry; repeat activation, repeated failure, success, empty result, exhausted
page and restart-required failure; outside focus/scroll, scope replacement and
authority loss. More than 100 synthetic Parties drive the Chromium paging case.
The new browser scenario covers both requested viewports, failed and successful
Retry, exhausted paging, off-page selection, parent draft retention and absence
of a Task Request write. Routed Assessment and Note scenarios exercise direct
shared-component use, source replacement and Escape closure during held reads.
Existing ordinary, related-Evidence and selected-reference removal scenarios pass.

The shared browsing component keeps the initiating button mounted, named and
focusable during its admitted read. `aria-busy`, `aria-disabled` and the described
status expose pending/unavailable state; activation guards and a synchronous
local latch supplement the discovery owner's admission check. An exhausted
focused control stays focusable but unavailable until focus leaves, then uses
native disabling. Retry remains through its pending read. After success or a
restart-required failure, it remains unavailable; after focus departure it
stays mounted and becomes natively disabled. This keeps the adjacent Apply
button in place during pointer activation. Repeated failure leaves Retry
available. No completion programmatically focuses or scrolls.

One attachment-local intent records the initiating element, action, exact
controller/scope and interaction ownership. Newer keyboard, pointer, focus,
input and scroll interaction retires focus ownership. Scope/authority replacement,
closure and unmount prevent stale focus claims. It stores no candidates,
selection, drafts, permissions, operations or requests. Existing accepted-page
selection and Apply/Cancel remain available during permitted reads and failures.
No compatibility adapter is needed; both direct consumers use the same shared
component. The removed behavior is native pending disabling of initiating
buttons and immediate conditional Retry removal. Real capability disabling and
authorization concealment remain.

Changed files:

| Path | Change |
| --- | --- |
| `apps/web/src/workbook/components/WorkbookCandidateBrowsing.tsx` | Shared read-control continuity and guarded activation. |
| `apps/web/src/workbook/components/WorkbookAuthoringReferencePicker.test.tsx` | Controlled promise focus, failure, exhaustion, empty, duplicate and retirement assertions. |
| `apps/web/src/workbook/features/assessments/assessmentDiscovery.test.tsx` | Direct shared support-picker held-Refresh assertion. |
| `apps/web/e2e/authoring-candidates.spec.ts` | Deterministic Chromium gates and consumer/focus assertions. |
| `apps/web/src/workbook/components/README.md` | Local source guide for the presentation boundary. |
| `tools/test_families/web.workbook.json` | Routed new unit titles. |
| `tools/test_families/module.workbook.json` | Routed new Party browser scenario/title. |
| `tools/execution_topology_render_index.json` | Regenerated projection of the authored test-family inputs. |
| `docs/handoffs/ui-ux/workbook-candidate-read-keyboard-continuity-handoff.md` | Durable scope, evidence and acceptance record. |

## Verification ledger

All repository commands ran from the root through Make. `make help`,
`make help-all`, and both `make task-guide ROLE=module-author OWNER=...`
commands passed and confirmed the narrow owner routes.

| Command or evidence | Result |
| --- | --- |
| Baseline `make test-slice` for authoring discovery/presentation | PASS, 3/3, `.cartulary/test-results/20260927T144305Z-p18329`. |
| Initial `make generate` | FAIL, `.cartulary/test-results/20260927T144859Z-p21814`: new authored titles were not ASCII sorted; corrected. |
| First `make format` | FAIL before formatting: new browser title lacked a paired scenario ID; corrected in `module.workbook.json`. |
| Preproduction presentation test route | Expected FAIL, `.cartulary/test-results/20260927T145101Z-p29370`: initiating pending button lacked `aria-busy` and Retry disappeared. An earlier red run at `.cartulary/test-results/20260927T145012Z-p28240` also exposed an unsupported Chai matcher in the new test; corrected. |
| `make generate` | PASS, `.cartulary/test-results/20260927T150153Z-p42517`; no generated root was hand-edited. |
| First three-row service-backed browser run | FAIL, `.cartulary/test-results/20260927T150206Z-p45394`: test assumptions about pointer/select focus and the existing Party Apply sequence; trace and snapshots inspected. |
| Party-only rerun | FAIL, `.cartulary/test-results/20260927T150549Z-p79403`: after successful Retry, pointer Apply became active but the picker stayed open. The trace showed Retry removal during focus departure shifted Apply before click completion; no JavaScript exception was recorded. This was a presentation regression in this slice, not presumed historical flakiness. |
| Pointer transition fix | Retry now stays mounted and natively disabled after focus departure. The original pointer Apply assertion was restored; both unit and Chromium routes pass. |
| Final `make format` | PASS, `.cartulary/test-results/20260927T152311Z-p89568`. |
| Final `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_discovery,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.assessment_discovery` | PASS, 4/4, `.cartulary/test-results/20260927T152321Z-p93999`. |
| Final `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.authoring_candidates_parties,module.workbook.browser.authoring_candidates_assessments,module.workbook.browser.authoring_candidates_note_source` | PASS, 11/11 work units and 5/5 Chromium scenarios, `.cartulary/test-results/20260927T152333Z-p94828`. |
| Closure-strengthened service-backed browser rerun | PASS, 11/11 work units and 5/5 Chromium scenarios, `.cartulary/test-results/20260927T153608Z-p63218`; Note Escape before a held response completed did not reclaim focus. |
| Final seeded UI review | PASS as supporting rendered evidence, `.cartulary/test-results/20260927T152529Z-p28500`; after Enter Refresh, observed focus remained on Refresh at 1440×900 and 1024×720. Both original captures were inspected. Session stopped with an OK terminal receipt. The one-Party seed does not prove paging/failure behavior. |
| Final `make agent-finalize` | PASS, `.cartulary/test-results/20260927T153735Z-p96236`; `RESULTS_DIR` unset. |
| Final frontend typecheck, Biome, import boundary | PASS, `.cartulary/test-results/20260927T153806Z-p1013`, `.cartulary/test-results/20260927T153806Z-p1161`, `.cartulary/test-results/20260927T153806Z-p1193`. |
| Final generated drift, artifact policy, JSON shape | PASS, `.cartulary/test-results/20260927T153806Z-p809`, `.cartulary/test-results/20260927T153806Z-p897`, `.cartulary/test-results/20260927T153806Z-p1114`. |
| Final `make lint-markdown` | PASS; covers the local guide and this handoff. The last run root is recorded in the task report. |
| `make test-catalog-check` | PASS; authored titles and scenario IDs route. |
| `git diff --check` | PASS; no whitespace errors. |

`RESULTS_DIR` was unset because there was no qualifying successful full warm
run, so retained-run maintenance was skipped. Broader product suites and visual
golden refreshes were skipped: this change affects the shared candidate read
control, the focused owner routes passed, and no visual styling, grid, backend,
schema or persistence boundary changed. The review artifacts are private and
their links expire when the seeded session stops; the product-test runs above
are the durable behavioral evidence. No commit or publication was performed.

## Digest acceptance assessment

The following applies the digest rows to this bounded slice. PASS concerns the
named changed boundary; it does not claim unrelated subsystem conformance.

| Row | Assessment and evidence |
| --- | --- |
| A001 | PASS: exact Core 01/Core 03/design owners and independent source/test ownership recorded above. |
| A002 | PASS: one existing shared browsing component holds the common decision; no workflow migration or parallel engine. |
| A003 | PASS: branch, HEAD, clean baseline, consumers, guides and authored/generated manifests checked. |
| A004 | PASS: no component-local color/spacing literal or new token registry. |
| A005 | PASS: final rendered review used the current dark graphite theme; no theme path changed. |
| A006 | N/A: density selection and Grid Adapter geometry were not changed. |
| A007 | PASS: contextual raw fields and source, ordinary/related-Evidence consumers, staged Apply/Cancel and no parent write covered by routed browser tests. |
| A008 | N/A: responsive threshold/clamp behavior was not changed; focus was separately reviewed at 1440×900 and 1024×720. |
| A009 | N/A: shell overflow/navigation ownership was not changed; reviewed focus was visible in the inspector. |
| A010 | N/A: inspector feature dispatch, review confirmation and detachment were not changed. |
| A011 | PASS: focus, off-page selection, raw parent authoring, source replacement and stale-response retirement asserted. |
| A012 | N/A: transaction IDs and uncertain write replay were not changed; duplicate candidate read activation was tested. |
| A013 | N/A: write acknowledgement and mutation recovery were not changed. |
| A014 | N/A: grid/inspector editing and validation were not changed; parent raw draft retention was checked. |
| A015 | N/A: conflict resolution was not changed. |
| A016 | PASS: pending/failure/empty/exhausted/restart presentation and read admission agree; selection and Apply remain separate. |
| A017 | PASS: accepted page remains during permitted pending/failure; authority concealment and replacement fence stale responses. |
| A018 | N/A: Evidence lifecycle, overlay and preview were not changed. |
| A019 | PASS: named controls, keyboard Enter/Space/Tab, unavailable semantics and visible focus checked in unit, Chromium and rendered review; no motion/color behavior changed. |
| A020 | PASS for the changed control state: pending, Retry, exhaustion and viewport presentation covered; broader density/zoom component variants are outside the slice. |
| A021 | N/A: Grid Adapter virtualization and authoritative row identity were not changed. |
| A022 | PASS as implementation-support evidence: final seeded captures at both requested viewports inspected; no golden changed or publication claim made. |
| A023 | PASS: semantic button, region, record and view identities used; new titles and scenario ID explicitly routed. |
| A024 | PASS: documentation remains outside executable product and test inputs. |
| A025 | PASS: authored selector changes generated the one index; generation, drift and policy checks passed. |
| A026 | PASS: no invented route, schema, write, permission or storage behavior; supported shared consumers preserved; no migration required. |
| A027 | PASS: this handoff includes owners, files, behavior, failures, commands, limitations, skipped checks and rollback. No controlling tracker update was required for this bounded implementation slice. |

Rollback consists of reverting the nine changed source, test, selector, guide,
handoff and generated-index files as one focused change, then running the public
generator and focused checks. There is no data conversion or compatibility
adapter to undo.
