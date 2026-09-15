# Workbook query continuation and recovery

## Control and baseline

This is the controlling WQC-01 through WQC-05 record. Only the current row may
be IN_PROGRESS. Save its actual exit evidence and DONE status before beginning
its successor. An applicable blocked dependency prevents advancement.

Execution begins on clean `main` at
`129d1d8ba8b57cce08242cc8be3b42bc6accab78`, matching the planning baseline.
`git status --short --branch`, `git rev-parse HEAD`, `git diff --stat`, and
AGENTS.md discovery confirmed no pre-existing changes. The digest's localized
read order was followed during planning. Digest and research instructions are
advisory, not additional product authority.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| WQC-01 Characterization and browsing contract | DONE | Approved owner decisions, complete consumer/lifetime matrix and expected-red adapter reproduction recorded. |
| WQC-02 Complete results and coherent ownership | DONE | Complete correlated envelopes, canonical/authored separation, independent Entity browsing, reference capabilities and focused verification pass. |
| WQC-03 Continuation and spreadsheet integration | DONE | Bounded continuation, recovery, live coalescing, retained sources and grid integration pass focused checks. |
| WQC-04 Integrated evidence | DONE | All seventeen surfaces, long traversal, canonical queries, live/security recovery, spreadsheet regressions, accessibility and Timeline measurements pass. |
| WQC-05 Final validation | DONE | Required terminal checks, two fresh visual validations, all changed-image reviews, completed-handoff formatting and final scope review pass. Stop at this seam. |

## Authority and approved decisions

Behavior: Core 01 §§3.3.4–3.3.4.1, §3.3.7, REQ-01-240–242/554–559;
Core 03 REQ-03-099/100/218/224/275/297–300 and §§13–14; Core 04 current
authorization and concealment. Domain vocabulary and design direction apply
within their declared boundaries. Source placement belongs to frontend source
and import manifests; verification routing belongs independently to the owner
catalog and authored test families. No executable artifact depends on Markdown.

The user approved three 100-row response pages, twenty earlier request
checkpoints, explicit continuation, loaded-window selection, a trailing Timeline
draft, retained editor detachment during explicit browsing, re-fetching the last
position on sheet return, bounded live window reconciliation, and retaining
accepted labels when a replacement query fails. Core query cursors remain opaque
live-authorized continuations; Network Analysis cursor or expiry rules do not apply.

### Surface and lifetime matrix

| Consumer | Query and normalization owner | Browsing / retention / recovery disposition |
| --- | --- | --- |
| Timeline | Timeline loader, projection, committed ledger and creation owners | Shared bounded window; Timeline projects drafts and one creation pin; original-source obligations stay independent of query membership. |
| Hosts | Entity conversion/indexing owner | Independent Host window, request/error lifetime and checkpoint; references do not depend on loaded-page completeness. |
| Identities | Entity conversion/indexing owner | Independent Identity window with the same neutral mechanics; no paired-load failure coupling. |
| Assessments | Assessment committed-record and append-only authoring owners | Shared window; accepted source versions and append-only semantics remain with Assessment. |
| Generic registered surfaces | Contract normalization and each source owner | Schema-keyed window; Notes, Evidence, Indicator, coordination and optional artifact semantics remain local. |
| Reference broker and Entity Timeline preview | Existing reference/preview owners | Migrate complete port envelopes; keep auxiliary reads independently scoped, bounded and explicitly partial. |
| Candidate discovery, History, saved-view listing, Network Analysis | Existing independent owners | No browsing-policy migration; shared contract fixtures change only where required. |

All main surfaces use keyboard-accessible Load more, Earlier rows and Refresh.
The fourth page evicts the oldest. Earlier rows re-fetches one preceding request
into a replacement window and retires old forward descendants. A sheet departure
releases query rows and retains a bounded request checkpoint and semantic anchor.
Refresh and invalid-cursor recovery restart without a cursor. Explicit browsing
can detach a draft without submitting it; cell navigation retains acceptance gates.
Select-all covers loaded committed query members only, never unloaded matches,
drafts or an out-of-query creation pin. Eviction prunes selection; captured writes
and receipts are unaffected. Authority invalidation clears query rows and cursors;
retained work follows its existing account/incident owner.

## Gap register

Each confirmed gap includes remediation, affected areas, rationale and long-term
benefit, compatibility, unresolved risk and binary validation. Inspection is not
browser execution evidence.

| Gap | Observation and remediation / affected areas | Rationale, benefit and compatibility | Risk / binary validation |
| --- | --- | --- | --- |
| G01 | The query port/adapter discard canonical and paging metadata. Extend the semantic request/result and migrate consumers; implementation, tests and guides. | One correlated read boundary; no new HTTP route or data migration. | Partial admission / full validated rows and metadata publish together. |
| G02 | Authored OpenAPI makes paging optional and a contract test requires that omission. Require paging and generate projections; projection, generated types, tests. | Matches Core 01's existing mandatory response; malformed old fixtures must change. | Fixture drift / missing paging rejects and real responses remain valid. |
| G03 | Query controls and grouping use requested state while old rows survive failure. Separate requested, effective and accepted state; implementation and tests. | Honest labels, stable saved-view comparison, no persistence migration. | Feedback loops / failure retains accepted labels; normalization causes no extra read or false dirty state. |
| G04 | Grouping injects a request sort and can turn eight overrides into nine. Remove injection; implementation and owner-directed test correction. | Clearing sort means schema defaults; grouping remains presentation-owned. | Group order / omitted sort and eight-entry grouped requests pass route and UI checks. |
| G05 | Hosts and Identities share one parallel load and error state. Separate browsing lifetimes and reference indexing; implementation, tests and guides. | One surface cannot block another; no public wire change. | Reference labels / later results and failures remain independently scoped. |
| G06 | Non-Timeline retention uses accepted row count. Track accepted result presence explicitly; implementation/tests. | Empty success survives refresh failure as known accepted state. | Empty-state regression / empty-then-failed-refresh remains distinct from initial failure. |
| G07 | Passive rows enter full-row committed caches without eviction. Separate query observation retention from retained work; owner stores, query integration and scale tests. | Long browsing stays bounded without discarding drafts or receipts. | Version regression / passive traversal plateaus while retained source versions survive eviction. |
| G08 | Main loaders never send a continuation cursor. Add shared bounded browsing and recovery; Core 03/design clarification, implementation, tests and guides. | All live matching records become reachable without fetch-all or offsets. | Live overlap/focus / later records, return, retries and keyboard continuity pass. |
| G10 | Real-route continuation after a Timeline edit reused the older retained page. Observe committed loaded rows before staging continuation; Timeline loader and 405-row route test. | Keeps accepted versions and continuation coherent; no wire or stored-data change. | Older retained copies / editing a later row then loading two further pages preserves version 2 and permits Earlier rows. |
| G11 | Real Hosts/Identities browsing exposed an authorization/reference-loader feedback loop and an unstable off-window inspector projection. Separate initial authority reads from sheet reads, use only the selected reader, memoize retained Entity conversion, and remove passive inspector cache insertion; refresh controller, Entity surface/inspector, unit and real-route tests. | Independent lifetimes and bounded request/cache volume; existing authority and retained-work contracts remain. | Late reads or render recursion / both 405-row traversals complete with bounded requests, retained inspector identity and no page errors. |
| G12 | Sheet navigation changed the global query invalidator identity, recreated the collaboration coordinator and retired incoming browsing state. Keep a stable workbook invalidation binding with current surface callbacks; surface-query integration and long traversal test. | One incident coordinator lifetime with independent sheet lifetimes; no compatibility migration. | Navigation cancellation / a 2,505-row traversal returns to the terminal anchor and exhausts exactly twenty earlier checkpoints. |
| G13 | The real route returns `invalid_cursor_token`, expressly permitted by Core 04 AC-375, but the Core 01 reason table and typed error projection omitted it. Complete that reason table/projection and the frontend pagination recovery allowlist; Core 01, errors contract/generated outputs, browser and unit tests. | Restores the existing REQ-03-275 recovery behavior for malformed/tampered Core cursors; no backend or cursor-format change. | Misclassified failure / real HTTP 400 invalid cursor produces one cursor-free restart; all four pagination reasons pass bounded recovery tests. |
| G14 | Group headers followed first encountered query rows or an injected grouping sort, and some surfaces read cells instead of canonical grouping scalars. Add an optional Grid Adapter bucket comparator; Workbook supplies ascending/null-last policy and Timeline supplies REQ-03-230 overrides, using `group_values`; grid, all surface projections, owner tests and real grouped query evidence. | Correct presentation order without altering authored or server row sorts; additive internal grid capability, no wire migration. | Group boundaries / each bucket appears once in the specified order, within-bucket row order is unchanged, and eight-sort grouped requests remain valid. |

The initial hypotheses were late authority effects, off-window source actions,
range/selection drift, creation-pin accumulation, virtualized focus/scrolling and
shared loading performance. WQC-03/04 evidence below resolves them within the
explicit acceptance matrix. No adopted-owner contradiction remains.

## Evidence log

WQC-01 exit: Core 03 §14.9 and the selection clause now adopt the approved
browsing policy; design §8.3 carries its compact presentation. The adapter
characterization asserts preservation of opaque cursor bytes and canonical
metadata. `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_query_port_validation_6d44f87e20`
failed as expected (1/2 units) at
`.cartulary/test-results/20260915T140222Z-p35397`: the accepted value lacks
`paging` and `canonicalQuery`; malformed/cross-context/abort coverage passed.
The failure is directly related to G01, not an environmental blocker. Together
with the passing baseline hook/model and real-route evidence below, this
completes characterization and owner disposition. WQC-02 must make that
regression pass before its exit. No unresolved owner contradiction remains.

Planning baseline evidence, all PASS:

- `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_query_port_validation_6d44f87e20,module.timeline.frontend.timeline_load_machine_a109000001`:
  3/3 units, `.cartulary/test-results/20260915T134150Z-p10103`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.use_generic_surface_query_owner_fab9b06dde,web.workbook.regression.entity_merge_query_concealment,web.workbook.regression.assessment_reconciliation,web.workbook.regression.workbookquery_suite_904073db6c`:
  5/5 units, `.cartulary/test-results/20260915T134447Z-p11669`.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.cursor_pagination_re_derives_live_authorization_2d2231384b`:
  3/3 units, `.cartulary/test-results/20260915T134609Z-p14363`.

Task guides were inspected for module.workbook, web.workbook, module.timeline,
module.savedviews, package.grid_adapter, platform.viewquery, module.entities,
module.assessments, package.protocol_ts and web.architecture. A lookup of
`package.protocol` failed because it is not an active owner; discovery corrected
the identifier to `package.protocol_ts`, whose guide passed.

## Compatibility, limitations and rollback

No new endpoint, search operator, dependency, persistent browser storage,
snapshot pagination, deep offset, analyst-data operation, digest edit, commit,
push or deployment is included. The response projection correction enforces an
existing owner obligation; internal TypeScript consumers migrate together.

Rollback restores owner text, authored projections, generated derivatives,
implementation and tests together while preserving analyst records and history.
The final acceptance and rollback instructions below supersede historical
next-action entries in the evidence log. RESULTS_DIR remained unset;
agent-finalize recorded retained-run maintenance as skipped.

### Additional confirmed gap G09 — query pagination error projection

- Remediation: add Core 01's `invalid_view_query` reason registry to the typed
  error projection and regenerate protocol outputs. The public-error decoder
  currently discards every query pagination reason, preventing Core 03 recovery.
- Affected areas: Core 01 §3.3.6.2 (unchanged owner), `contracts/errors/index.json`,
  generated protocol error registry, semantic adapter recovery tests, this handoff.
- Rationale and benefit: preserve validated owner-defined code/reason pairs for
  local recovery without trusting arbitrary error details or duplicating tokens.
- Compatibility: additive reason typing; no route or runtime payload change.
- Risk: unrelated reason registries must remain unchanged.
- Binary validation: adapter preserves `cursor_query_mismatch`; invalid-cursor
  browser recovery makes exactly one automatic cursor-free restart; unknown
  reason values remain untrusted.

### WQC-02 investigation evidence

- Adapter slice PASS, `.cartulary/test-results/20260915T142227Z-p50986`.
- Browser and authored-query rows PASS in
  `.cartulary/test-results/20260915T142227Z-p50987`; one generic fixture failure
  exposed an incomplete typed mock, since corrected.
- Metadata and generic rows PASS,
  `.cartulary/test-results/20260915T142541Z-p60785`.
- Assessment row PASS, `.cartulary/test-results/20260915T142541Z-p60815`.
- Entity row initially failed its old shared-error notification assumption;
  active-sheet production reads now have independent lifetimes. Updated fixture
  expectation and added an explicit active-sheet isolation test.
- `make generate` initially rejected the additional required paging member in
  compatibility review. Added the exact owner-backed projection correction to
  the current authored release change set; generation then PASS at
  `.cartulary/test-results/20260915T142247Z-p52589`.
- `make format` PASS at `.cartulary/test-results/20260915T142505Z-p56239`.
- Initial typecheck failures identified incomplete semantic mocks and G09.
  These are implementation work, not environmental blockers.

### WQC-02 exit

- All main consumers stage complete normalized results through
  `WorkbookQueryBrowser`; Timeline, Entity, Assessment and generic projection
  remain with their existing owners. Canonical presentation and saved-view
  serialization retain authored sort overrides and omitted inactive grouping.
- Hosts and Identities read independently. Timeline obtains bounded Entity
  reference observations through the existing reference broker; those
  observations are independent of both sheets' filters and are not a registry.
  Reference broker and Entity Timeline preview explicitly request 100 rows and
  continue to belong to their existing auxiliary scopes.
- `make test-slice OWNER=web.workbook` with query browsing, query metadata,
  generic query and workbookQuery rows: PASS 5/5,
  `.cartulary/test-results/20260915T143117Z-p71842`.
- Entity query slice including active-sheet isolation: PASS 2/2,
  `.cartulary/test-results/20260915T143117Z-p71857`.
- Timeline adapter slice including opaque cursor and reason-code preservation:
  PASS 2/2, `.cartulary/test-results/20260915T143117Z-p71877`.
- Assessment query slice: PASS 2/2,
  `.cartulary/test-results/20260915T142541Z-p60815`.
- Viewquery normalization/OpenAPI requirement slice: PASS 1/1,
  `.cartulary/test-results/20260915T143406Z-p79193`.
- Reference broker slice: PASS 2/2,
  `.cartulary/test-results/20260915T143428Z-p80022`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260915T143406Z-p79279`.
- `make generate`: PASS after G09 projection,
  `.cartulary/test-results/20260915T142748Z-p63403`.
- `make format`: PASS, `.cartulary/test-results/20260915T143429Z-p80339`.
- Compatibility: complete metadata is required at the semantic boundary; test
  doubles were migrated. Core route payloads remain unchanged. The release
  change set records the correction from optional to required paging.
- Next action: WQC-03 explicit controls, authority-wide cleanup, source and
  draft continuity, query observation retention, and grid integration.

### WQC-03 exit

- Added compact work-area controls, synchronous duplicate rejection, atomic
  three-page reconciliation, twenty request-only checkpoints and anchor-based
  sheet return. Failed replacements retain accepted query presentation and offer
  Retry/Revert. Failed cursor-free restarts cannot reuse retired cursors.
- Added shared freshness/cursor recovery bounds, producing-request correlation,
  permitted overlap admission and version floors. Placement changes route back
  through Core queries, including synthetic full-text predicates.
- Grid Adapter now supports explicit editor detachment and semantic anchor
  capture. Range membership survives append but clears after eviction or changed
  contiguity. Selection prunes to loaded members and excludes Timeline's pin.
  Controls retain focus during browsing. Accessible descriptions identify the
  loaded window rather than global row ordinals.
- Timeline's foundation now owns its committed-record capability, shared by the
  inspector and mutation coordinator. Evicted passive observations are released;
  retained drafts, one inspector source and accepted mutation versions survive.
  Source coordination checks off-window retained drafts without scanning pages.
- Removed `query/entityLiveEventPatchPlanner.ts`, its obsolete paired-sheet test
  and its source/test routing. Independent Entity loaders own live patches.
- `make test-slice OWNER=web.workbook` with `query_browsing`,
  `query_browsing_controls`, `inspector_window_retention` and generic-loader rows:
  PASS 5/5, `.cartulary/test-results/20260915T150835Z-p27473`.
  Expanded recovery controls: PASS 3/3,
  `.cartulary/test-results/20260915T151325Z-p43980`.
- Timeline adapter/load-machine/ledger/coordinator slice: PASS 5/5,
  `.cartulary/test-results/20260915T150835Z-p27461`. Added evicted-source
  coordination and retained-work ledger tests: PASS 4/4,
  `.cartulary/test-results/20260915T151439Z-p49242`.
- Grid range and complete adapter slice: PASS 3/3,
  `.cartulary/test-results/20260915T150835Z-p27483`. Added explicit editor-detach
  regression: PASS 2/2, `.cartulary/test-results/20260915T151209Z-p38134`.
- Timeline/Entity inspector and shell regression slice: PASS 6/6,
  `.cartulary/test-results/20260915T150914Z-p29980`. Entity loader PASS 2/2,
  `.cartulary/test-results/20260915T150915Z-p30215`; Assessment loader PASS 2/2,
  `.cartulary/test-results/20260915T150916Z-p30604`.
- `make format`: PASS, `.cartulary/test-results/20260915T151438Z-p49040`.
  `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260915T151440Z-p50611`.
- Intermediate type/format failures identified missing dependencies, obsolete
  inspector parameters and test fixtures; corrected before the passing runs.
  A test-family prefix error prevented execution and was corrected. A diagnostic
  override attempt (`make lint-biome BIOME_CHECK_FLAGS=...`) was rejected by the
  harness input contract; normal public targets were used afterwards.
- Remaining evidence work: real API/browser traversal, live authority changes,
  large-result layouts, scrolling and full spreadsheet regressions in WQC-04.
  No source-data migration. Rollback must restore Core 03/design clarification,
  OpenAPI/error projections and generated derivatives with the implementation,
  routing and tests; accepted analyst records and history remain untouched.

### WQC-04 investigation evidence

The first real-route run (`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.query_continuation_notes,module.workbook.browser_stateful.query_continuation_timeline`) failed at `.cartulary/test-results/20260915T152209Z-p72031`: the test located an editor beneath a display node removed during editing. The row-scoped editor locator now follows the production editor. The next five-surface run at `.cartulary/test-results/20260915T152401Z-p5692` passed Notes and Assessments and confirmed G10/G11. The rerun at `.cartulary/test-results/20260915T153413Z-p48142` passed Timeline and exposed remaining Entity reader/inspector lifetime coupling. These are implementation failures, not environmental blockers.

`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.useworkbookprojectionrefreshcontroller_suite_5bad71bbc5,web.workbook.regression.query_browsing` passes 3/3 units at `.cartulary/test-results/20260915T153413Z-p48128`, including the reproduced authority-reader feedback dependency.

The 16-scenario route run at `.cartulary/test-results/20260915T153944Z-p21874` passed thirteen scenarios, including both Entity sheets and canonical eight-sort/saved-view persistence. The other two surface failures captured a cursor before a legitimate startup authorization reread; assertions now correlate the actual continuation cursor to its exact preceding producing response. The long return failure confirmed G12. The six-scenario run at `.cartulary/test-results/20260915T154556Z-p66774` passes Decisions, Findings, the 2,505-row history/draft traversal and live role/membership concealment. Its remaining failures confirmed G13 and a test assumption that closed incidents accept a collaboration socket; the latter now observes expiry by an explicit query, consistent with incident-closure ownership.

`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.query_continuation_live_recovery,module.workbook.browser_stateful.query_continuation_session` passes 11/11 units at `.cartulary/test-results/20260915T155009Z-p30222`. Real-route evidence now includes duplicate activation, exact failed-read retry, malformed metadata rejection, invalid-cursor restart, unapplied query Revert, live insert/delete/restore, older query versus newer committed data, closed-readable continuation, expiry, and account replacement.

Selected Grid autosave, batch recovery and mutation-lifecycle accessibility rows pass 13/13 at `.cartulary/test-results/20260915T154950Z-p71720`. The nine-row committed editing regression run passed eight and exposed a nondeterministic Timeline fixture: empty activity timestamps tie on opaque record IDs, so the original row can be last. Its acceptance-navigation case now explicitly selects the authored synopsis sort before editing; the full assertion remains. The corrected case plus grouped canonical continuation pass 13/13 at `.cartulary/test-results/20260915T155458Z-p78487`. The live cursor authorization integration row also passed in `.cartulary/test-results/20260915T154750Z-p99933`.

The full-text/prefix browser row passed at `.cartulary/test-results/20260915T154749Z-p99709`. That run exposed the old Timeline group-order expectation; REQ-03-230 requires rough before reviewed. The owner-correct comparator and updated query/group assertions pass 11/11 at `.cartulary/test-results/20260915T155457Z-p78217`. Group comparator/row-order and browser recovery focused units pass at `.cartulary/test-results/20260915T155523Z-p36388` (4/4) and `.cartulary/test-results/20260915T155524Z-p36638` (3/3).

### WQC-04 exit evidence

- All seventeen registered surfaces have real-route later-result inspection and
  permitted editing evidence. Timeline/Hosts/Identities/Assessments/Notes use 405
  records; the other twelve use 105. The registry assertion fails if a surface is
  added without coverage. Passing scenarios are recorded in the runs above.
- Canonical full-text, set-like tags, prefix values, eight authored sorts with
  grouping, omitted cleared sort/grouping and saved-view persistence pass in
  `.cartulary/test-results/20260915T160113Z-p85821` (11/11 units). The same run
  passes the 2,505-record traversal, twenty-checkpoint exhaustion, retained
  off-window draft, sheet-anchor return, narrow layout and 200% zoom checks.
  The first tag fixture used an unsupported scalar array; it now uses the
  existing `collection_actions_v1` create contract. No production assertion changed.
- Timeline loaded selection remains 100 after appending the second page;
  continuation never expands its captured selection. The 405-row Timeline case
  passes at `.cartulary/test-results/20260915T155855Z-p40949`.
- All twelve Timeline entry/paste/fill/creation/acceptance/virtualization rows pass
  in `.cartulary/test-results/20260915T155633Z-p76548` (11/11 execution units).
- All four AC-043 Timeline measurement rows pass, including 25-session background
  traffic and one warm-up plus 100 samples per interaction:
  `.cartulary/test-results/20260915T155634Z-p76766` (20/20 units). Measurement
  observations remain implementation evidence; no new conformance claim is made.
- Saved-view query replay through reload passes at
  `.cartulary/test-results/20260915T155525Z-p37023` (11/11). Its request observer
  now waits for the requested sort body instead of an unrelated live reread.
- The final pre-terminal typecheck passes at
  `.cartulary/test-results/20260915T160023Z-p79093` (2/2). Earlier typecheck failures
  were new-test unused imports, unsupported `findLast`, missing scalar export and
  duplicate test-support type export; all were corrected without config changes.

Every confirmed gap has passing boundary evidence. No owner contradiction or
blocked dependency remains. WQC-05 next: finalize maintenance, terminal owner/type/
boundary/browser/visual/drift checks, formatting and final scope review.

### WQC-05 terminal investigation

`make agent-finalize` passed 1/1 at
`.cartulary/test-results/20260915T160311Z-p18020`, before broader terminal
verification. `RESULTS_DIR` was unset; retained-run maintenance was skipped.

The first full `make frontend-unit` passed 614/628 execution units at
`.cartulary/test-results/20260915T160347Z-p22417`. Failures exposed old request
fixtures and two integration defects. Complete test transports now correlate
canonical metadata to the captured request before resolving deferred responses;
raw malformed-response tests retain independent transports. Expectations require
`limit: 100`, omit grouping-induced sorts, retain off-window inspector identity,
and account for independent Host/Identity failures. The full rerun reached
627/628 at `.cartulary/test-results/20260915T161943Z-p14405`; its remaining
collection-draft regression was corrected and passed its focused row at
`.cartulary/test-results/20260915T162414Z-p94308`.

| Gap | Remediation and affected areas | Rationale, benefit and compatibility | Risk and binary validation |
| --- | --- | --- | --- |
| G15 | A socket can publish a collection's accepted version before HTTP settlement; inspector collection revisions were not captured, and duplicate clearing through two registries lost ownership of mounted input cleanup. Capture the shared collection revision from either editor context, clear through the mounted owner once, and clear only matching submitted input text; Timeline draft registry, retained mutation owner, unit and real-route mention visual fixture. | Accepted text disappears while newer typing survives; scalar context isolation and detached settlement remain unchanged. No data or wire migration. | Newer draft loss / accepted collection input clears, a newer input stays intact, and real auto-resolution converges at the accepted version. |
| G16 | Generic Core fallback focus overrode an extension's existing page policy; deferred empty-query acceptance dropped its action focus anchor when a disabled button left focus on the document body. Restrict Core-record fallback and retain the action anchor through asynchronous refreshing until accepted replacement or a newer interaction; Grid Adapter and delayed-action/Network Analysis regressions. | Preserves independent extension ownership and predictable Clear filters recovery. Additive internal behavior; no stored-state migration. | Focus theft / Core empty recovery focuses the grid root and the existing Network Analysis replacement focus row passes. |

The first `make test-slice OWNER=module.workbook` terminal run passed 83/97 at
`.cartulary/test-results/20260915T160414Z-p53999`; it also exercised all 22 new
real-route browsing cases successfully. Remaining failures were correlated unit
fixtures, intentional visual differences, obsolete query-body/chip expectations,
reference reads on unchanged authorization, and focus/retained-work regressions.
The targeted browser rerun passed 20/24 at
`.cartulary/test-results/20260915T161942Z-p14150`: closure, preference operations,
canonical optional-surface chips, Timeline error presentation and Coordination
accessibility passed. The remaining support/save-status rows then passed 15/15
at `.cartulary/test-results/20260915T162514Z-p50507`.

Reference refresh now uses the latest broker without restarting on a successful
same-authority recheck. Existing authority invalidation still clears reference
observations and aborts pending reads. Its focused test asserts stable reader
identity and dispatch through the replacement broker. Query failures appear once
in the grid's owned error presentation; browsing controls supply compact
Retry/Revert context. Closed ordinary drafts remain copyable in the existing
retained-work area and return with their exact text after reopening.

Core 03 REQ-03-089 gives unresolved conflicts precedence until settlement. With
inactive Timeline rows released, resolving a conflict from Notes no longer runs
an inactive query that prolongs an intermediate Syncing paint. The save-status
fixture now asserts the exact five applicable announcements and zero inactive
Timeline reads, while retaining the original FIFO, conflict-scope and focus
assertions. No delay or fabricated transition was added.

Initial terminal import-boundary failure moved source coordination coverage
from Timeline hooks to composition and moved its narrow inspector port to the
Timeline model boundary. `make frontend-import-boundary-check` subsequently
passed at `.cartulary/test-results/20260915T161538Z-p6461` and
`.cartulary/test-results/20260915T162800Z-p99428`. Formatting/type failures were
local test imports, mock signatures, optional-value narrowing and lint rules;
configurations and tolerances were unchanged.

One visual attempt and two protocol attempts failed with artifact exit 11 while
formatting changed inputs after their run snapshots:
`.cartulary/test-results/20260915T162416Z-p94919`,
`.cartulary/test-results/20260915T162902Z-p47461`, and
`.cartulary/test-results/20260915T162904Z-p48146`.
The harness rejected those attempts before accepting their build evidence;
subsequent runs use settled source inputs.

Terminal investigation is closed. The final acceptance matrix and visual
refresh record below contain the passing evidence.

### Terminal commands and outcomes

| Command | Outcome | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make agent-finalize` | PASS 1/1; RESULTS_DIR unset | `20260915T160311Z-p18020` |
| `make frontend-unit` | PASS 628/628 | `20260915T162753Z-p92625` |
| `make frontend-typecheck` | PASS 2/2 | `20260915T164452Z-p24641` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260915T162800Z-p99428` |
| `make lint-biome` | PASS 2/2 | `20260915T163009Z-p77116` |
| `make frontend-fallow-static` | PASS 2/2 | `20260915T162515Z-p50797` |
| `make generate` | PASS; owner inputs generated | `20260915T161446Z-p97394` |
| `make generate-drift` | PASS 4/4 | `20260915T162755Z-p93120` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260915T162756Z-p93632` |
| `make json-shape-check` | PASS 3/3 | `20260915T162757Z-p94649` |
| `make test-slice OWNER=package.protocol_ts` | PASS 8/8 | `20260915T163007Z-p76007` |
| `make protocol-ts-browser-artifact-reachability` | PASS 3/3 | `20260915T163008Z-p76758` |
| `make format` | PASS 2/2 | `20260915T162902Z-p47047` |
| `make lint-markdown` | PASS, completed handoff | `20260915T165406Z-p1792` |
| `git diff --check` | PASS, completed handoff and final scope | No retained run artifact |

The final Viewquery unit selection was:

```sh
make test-slice OWNER=platform.viewquery ROWS=platform.viewquery.unit.list_query_duplicate_precedence_and_normalization_c36fa08271,platform.viewquery.unit.sort_and_filter_ceilings_canonical_normalization_4ca84c1e10,platform.viewquery.unit.view_query_filters_sort_and_group_by_accept_stab_0032c0a062
```

PASS 2/2 execution units at `20260915T162940Z-p66896`. The final support and
save-status selection was:

```sh
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.verify_browser_command_helpers_for_sort_filter_g_8f8dc4efe2,module.workbook.browser_support.verify_browser_command_helpers_for_sort_filter_g_cfa68b33e4,module.workbook.browser_stateful.workbook_save_status_preserves_authoritative_tra_0fa8b0b16c
```

PASS 15/15 at `20260915T162514Z-p50507`. Full commands, exact selected row IDs,
per-row results, logs and artifact references are retained in each run's
`run-manifest.json`, `target-summaries/`, and `rows/` outputs. Earlier failed
attempts are diagnostic history, not substituted for passing evidence.


The ordinary visual run at `20260915T162754Z-p92850` still found startup
reference cancellation and Clear filters focus during a refreshing intermediate
state. Both were fixed before golden mutation. The two exact visual rows then
completed all functional assertions at `20260915T164148Z-p21087` (Entities) and
`20260915T164149Z-p21511` (saved views); their remaining failures were screenshot
comparisons only. The Grid Adapter slice passed 2/2 at
`20260915T164205Z-p77208`; Entity query ownership passed 2/2 at
`20260915T164254Z-p88772`. Typecheck attempts at `20260915T164147Z-p20924` and
`20260915T164254Z-p88897` rejected a new test mock's empty-tuple return annotation;
the fixture now uses the same array type as its replacement broker.

| Gap | Remediation and affected areas | Rationale, benefit and compatibility | Risk and binary validation |
| --- | --- | --- | --- |
| G17 | Startup authority acceptance can replace the reference broker before its first read settles. Reissue only an unfinished reference obligation through the new broker, fence obsolete publication and keep accepted observations free of eager rereads; Entity query hook, existing owner test, query guide and Entity visual fixture. | Resolved labels survive startup reconciliation while request volume remains bounded. No wire or persisted-state migration. | Revalidation loop / replacement before first acceptance causes exactly one new reference read; accepted-set broker replacement causes none; resolved chips display the authorized target label. |

### Final implementation and compatibility

- Workbook query ownership now consists of `WorkbookViewQueryPort`, its semantic
  adapter/metadata validator, `WorkbookQueryBrowser`, the workbook-scoped registry
  and compact browsing controls. Authored overrides, canonical query metadata and
  accepted producing requests are separate. All primary loaders use this seam.
- Timeline retains projection, creation-pin and committed-version ownership;
  Generic, Entity and Assessment owners retain normalization, source semantics
  and independent inspector/operation capabilities. The reference broker and
  Entity Timeline preview consume the complete port result but keep their
  existing bounded auxiliary scope.
- Grid Adapter adds editor detachment, semantic anchor capture, valid-range
  retention and group bucket comparison. Workbook integration supplies loaded
  selection/count descriptions, approved grouping and focus fallback.
- Removed paths: `query/entityLiveEventPatchPlanner.ts` and its paired-query test,
  plus their authored routing. The four first-page-only loader implementations
  have been replaced in place. History, candidate discovery, saved-view resource
  listing and Network Analysis retain their own browsing owners.
- Authored OpenAPI now requires the already-mandatory paging response.
  `invalid_view_query` reason projections include the existing cursor failures.
  Make regenerated OpenAPI, Go/protocol derivatives and test topology. Core 03
  and design document the approved policy; Core 01's error table matches Core 04.
- Internal TypeScript callers and fixtures migrate together. There is no database
  migration or wire-route change. Older malformed test responses without paging
  intentionally fail validation. Saved views retain authored sort overrides;
  canonical defaults and record-ID tie breakers are never persisted as overrides.

### Resource and contract limits

The live query chain is not a snapshot. Each accepted window retains at most
three 100-row pages; an atomic reconciliation can temporarily stage three new
pages beside the accepted window. The browser retains twenty earlier producing
requests, not their evicted row payloads. Inactive sheets release query rows.
Forward traversal has no total-record cap and performs no prefetch. Earlier rows
can return only as far as the retained checkpoints; Refresh returns to the first
request. Counts and row indices refer to the loaded window, without a total.

Independent drafts, pending operations, receipts and the inspector can retain
original-source snapshots/version floors under their existing lifetimes. These
are retained work, not a passive query cache. Reference observations and Entity
Timeline previews remain bounded partial results with owner-defined fallback
labels. No page scan establishes deletion, authorization loss or mutation source
identity. Readable stale content does not renew authorization.

### Final acceptance matrix

| Boundary | Disposition and evidence |
| --- | --- |
| Adopted browsing policy and consumer inventory | PASS — WQC-01 owner clarification and surface/lifetime matrix. |
| Complete response, canonical/authored state, cursor bytes | PASS — metadata/adapter/browser owner tests, full frontend suite and canonical real-route case. |
| All registered surfaces, later inspection/editing and return | PASS — all 22 new real-route cases in `20260915T160414Z-p53999`; five 405-row surfaces and twelve 105-row surfaces. Other failures in that run are separately resolved above. |
| Bounds, overlap, empty/short pages, eviction and checkpoint exhaustion | PASS — neutral browser tests plus 405/2,505-record routes; no prefetch and bounded DOM/retained payload assertions. |
| Admission, duplicate activation, supersession and bounded recovery | PASS — adapter/browser owner tests and live-recovery real-route case. |
| Live insert/delete/restore, placement and committed-version floors | PASS — live-recovery, later-edit continuation, retained-source/ledger and grouping rows. |
| Authorization and concealment | PASS — live-cursor backend row, role/membership route case and closed/session/account route case. |
| Retained drafts, source actions, receipts, creation pin and bulk targets | PASS — retained-source/inspector/ledger unit rows, long traversal and all twelve spreadsheet regressions. |
| Canonical filters, cleared sort/grouping and saved views | PASS — canonical route case, query support rows and saved-view persistence/reload row. |
| Keyboard, grouping, scrolling, focus and accessibility | PASS — Grid Adapter owner suite, spreadsheet routes, accessibility rows, narrow/zoom traversal and focused empty-query visual functional assertions. |
| Timeline entry/navigation measurements | PASS — all four AC-043 rows, 100 samples each with 25-session background traffic. |
| Visual goldens and refreshed-manifest validation | PASS — all 143 changed images reviewed; update and two fresh ordinary visual runs pass against the promoted manifest. |
| Types, import boundaries, source/test ownership and generated drift | PASS — terminal commands table; final narrow reruns below include terminal fixes. |
| Markdown and whitespace | PASS — completed handoff passes `make lint-markdown` and `git diff --check`; final status update is validated with the same commands. |
| Snapshot/deep-offset/backend query rewrite, new endpoints/operators | N/A — explicitly excluded; existing real route already implements adopted continuation. Only demonstrated projection mismatches changed. |
| Release, deployment, Core 05 claim publication and analyst-data migration | N/A — outside the authorized seam. Isolated test fixtures are the only created records. |
| Retained full-warm-run maintenance | SKIPPED — RESULTS_DIR unset as requested; agent-finalize passed without retained-run maintenance. |

### Rollback

Revert this seam's owner text (Core 01/Core 03/design), authored OpenAPI/error
projections and release correction, Workbook/Grid Adapter implementation,
source/test ownership, local guides, browser fixtures and visual goldens as one
change. Restore generated derivatives through `make generate`, including the
matching error/OpenAPI/protocol outputs and test topology. Restore the visual
manifest with its matching goldens. Re-run the affected owner, frontend,
service/browser, generation/drift, artifact-policy and formatting checks against
that consistent source state. Do not revert database records, accepted mutations,
receipts or analyst history; no data rollback is part of this seam.


### Visual refresh record

Accepted triggers: approved Core 03 §14.9 browsing controls/window presentation,
REQ-03-230 group order without injected query sorts, and current-owner draft
presentation after incident closure. The ordinary closed-draft fixture retains
copyable text in its owned read-only area and no longer presents an editable
trailing draft while closed. All other reviewed pixel differences are the compact
browsing row, the resulting grid crop height/position, and group bucket order.
Inspector and overlay geometry remain owned by their existing work areas.

`make browser-e2e-visual-update` passed 12/12 at
`.cartulary/test-results/20260915T164327Z-p90522`. Reconciliation v3 accounts for
252 capture intents, 252 active goldens and all 29 registered fixtures, with zero
orphans, missing goldens, ambiguous mappings or unresolved registered fixtures.
The ordinary pre-refresh reconciliation at `20260915T162754Z-p92850` had zero
missing/ambiguous mappings; its eight uncaptured goldens were retained and are
fully accounted for by the successful update. Functional failures were corrected
and re-executed before mutation; no missing fixture was deleted.

All 143 changed images were reviewed in before/current contact sheets under
`.cartulary/wqc-final-visual-review/`, with full-size inspection of the default
Timeline and lifecycle-overlay captures where needed. Review passed: compact
controls, visible focus, retained drafts, density, narrow layouts, 200% zoom,
loading/error states, inspector scrolling and recovery overlays remain legible
and coherent. The 143 reviewed promoted-image hashes match the golden manifest.
The other 109 goldens retain their existing bytes. No viewport, browser zoom,
mask, scroll normalization, tolerance, snapshot selector or screenshot scope was
changed. The pinned renderer and vendored fonts remain unchanged. These are
implementation/design regression artifacts, not new Core 05 claims.

The following exact owner/fixture mapping comes from the successful run's
reconciliation against the authored catalog. Filenames are relative to
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. “Unregistered capture” means a
valid active catalog capture without a fixture-registry claim, not an orphan.

| Semantic owner row ID | Stable fixture ID | Changed golden filenames |
| --- | --- | --- |
| `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.presence_overflow` | `collaboration-presence-markers-linux.png` |
| `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.same_field_conflict` | `collaboration-conflict-resolver-linux.png` |
| `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec` | Unregistered capture | `collaboration-grid-presence-markers-linux.png` |
| `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | Unregistered capture | `collaboration-grid-conflict-resolver-compact-linux.png`, `collaboration-grid-conflict-resolver-linux.png`, `collaboration-grid-conflict-resolver-narrow-linux.png` |
| `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | Unregistered capture | `collaboration-grid-blocked-conflict-linux.png` |
| `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` | `entity-mention-chip-states-linux.png` |
| `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | Unregistered capture | `record-relationships-mention-chips-linux.png` |
| `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | Unregistered capture | `evidence-timeline-evidence-count-linux.png` |
| `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` | `evidence-affordance-states-linux.png` |
| `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | Unregistered capture | `evidence-grid-blocked-preview-linux.png`, `evidence-grid-timeline-evidence-badge-linux.png` |
| `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc` | Unregistered capture | `record-relationships-evidence-access-linux.png` |
| `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | Unregistered capture | `evidence-grid-available-evidence-linux.png`, `evidence-grid-requested-evidence-linux.png` |
| `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | Unregistered capture | `workbook-view-bar-filter-editing-overflow-linux.png`, `workbook-view-bar-long-columns-linux.png`, `workbook-view-bar-maximum-pressure-base-linux.png`, `workbook-view-bar-maximum-pressure-compact-linux.png`, `workbook-view-bar-maximum-pressure-narrow-linux.png`, `workbook-view-bar-ordered-maximum-sort-linux.png`, `workbook-view-bar-saved-view-actions-linux.png`, `workbook-view-bar-saved-view-clean-linux.png`, `workbook-view-bar-saved-view-modified-linux.png`, `workbook-view-bar-text-spacing-linux.png`, `workbook-view-bar-zoom-200-linux.png` |
| `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` | `workbook-query-empty-closed-read-only-linux.png`, `workbook-query-empty-compact-linux.png`, `workbook-query-empty-density-comfortable-linux.png`, `workbook-query-empty-density-compact-linux.png`, `workbook-query-empty-narrow-linux.png`, `workbook-query-empty-successful-query-linux.png`, `workbook-query-empty-text-spacing-linux.png`, `workbook-query-empty-zoom-200-linux.png`, `workbook-query-filtered-empty-linux.png` |
| `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.saved_view_query_controls_and_grouped_result` | `workbook-query-saved-view-query-controls-linux.png` |
| `module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810` | Unregistered capture | `timeline-grid-grouped-grid-linux.png` |
| `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206` | Unregistered capture | `timeline-grid-timeline-default-linux.png` |
| `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf` | `visual.fixture.edit_cell` | `timeline-grid-active-edit-cell-linux.png` |
| `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.compact_desktop_workbook_shell` | `incident-directory-compact-desktop-workbook-shell-linux.png` |
| `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.default_timeline_workbook_shell` | `incident-directory-default-timeline-workbook-shell-linux.png` |
| `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.narrow_desktop_workbook_shell` | `incident-directory-narrow-desktop-workbook-shell-linux.png` |
| `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` | `workbook-inspector-history-linux.png`, `workbook-inspector-public-error-linux.png`, `workbook-inspector-relationships-linux.png`, `workbook-inspector-rollback-preview-linux.png` |
| `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` | `workbook-inspector-narrow-technical-details-linux.png` |
| `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | Unregistered capture | `timeline-mutation-empty-timeline-query-linux.png`, `timeline-mutation-pending-replay-status-linux.png`, `timeline-mutation-transaction-recovery-panel-compact-linux.png`, `timeline-mutation-transaction-recovery-panel-linux.png`, `timeline-mutation-transaction-recovery-panel-narrow-linux.png` |
| `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | `visual.fixture.edit_cell` | `timeline-mutation-active-edit-cell-linux.png` |
| `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc` | `visual.fixture.task_requests_or_decisions` | `record-relationships-task-requests-linux.png` |
| `module.workbook.visual.contextual_task_decision_creation` | Unregistered capture | `contextual-decision-authoring-linux.png`, `contextual-decision-recovery-linux.png`, `contextual-task-request-authoring-linux.png`, `contextual-task-request-recovery-linux.png` |
| `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` | `coordination-comm-log-authoring-linux.png`, `coordination-handoff-authoring-linux.png`, `coordination-lesson-authoring-linux.png`, `coordination-recovery-linux.png`, `coordination-recovery-narrow-linux.png`, `coordination-status-review-authoring-linux.png` |
| `module.workbook.visual.decision_supersession_review_recovery` | Unregistered capture | `decision-supersession-accepted-linux.png`, `decision-supersession-review-linux.png` |
| `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` | `indicator-lifecycle-authoring-linux.png` |
| `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` | `indicator-observation-authoring-linux.png` |
| `module.workbook.visual.note_create_authoring_recovery` | Unregistered capture | `linked-note-authoring-linux.png`, `linked-note-recovery-linux.png`, `linked-note-recovery-narrow-linux.png` |
| `module.workbook.visual.ordinary_create_authoring_recovery` | Unregistered capture | `ordinary-closed-retained-narrow-linux.png`, `ordinary-recovery-1280-linux.png`, `ordinary-recovery-390-linux.png`, `ordinary-reference-authoring-linux.png` |
| `module.workbook.visual.preferences` | Unregistered capture | `workbook-preferences-comfortable-linux.png`, `workbook-preferences-compact-linux.png`, `workbook-preferences-confirmed-stale-linux.png`, `workbook-preferences-uncertain-linux.png`, `workbook-preferences-uncertain-narrow-linux.png`, `workbook-preferences-unset-linux.png` |
| `module.workbook.visual.timeline_capture_actions` | Unregistered capture | `timeline-supersession-accepted-linux.png`, `timeline-supersession-authoring-linux.png`, `timeline-supersession-review-linux.png` |
| `module.workbook.visual.timeline_related_evidence` | Unregistered capture | `timeline-related-evidence-authoring-linux.png`, `timeline-related-evidence-partial-linux.png`, `timeline-related-evidence-partial-narrow-linux.png` |
| `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | Unregistered capture | `account-menu-controls-compact-linux.png`, `account-menu-controls-narrow-linux.png`, `account-menu-controls-short-linux.png`, `account-menu-long-label-text-spacing-linux.png`, `account-menu-long-label-zoom-linux.png`, `account-menu-workbook-root-linux.png` |
| `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | Unregistered capture | `design-grid-closed-read-only-rows-linux.png` |
| `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.delayed_initial_loading` | `design-delayed-initial-loading-linux.png`, `design-immediate-initial-loading-linux.png` |
| `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.error_presentation_loci` | `design-grid-background-refresh-linux.png`, `design-grid-stale-refresh-linux.png`, `design-grid-unavailable-initial-load-linux.png` |
| `web.design.visual.lifecycle` | Unregistered capture | `lifecycle-closed-linux.png`, `lifecycle-comfortable-linux.png`, `lifecycle-compact-linux.png`, `lifecycle-confirmed-refresh-failure-linux.png`, `lifecycle-pending-linux.png`, `lifecycle-reason-linux.png`, `lifecycle-review-linux.png`, `lifecycle-review-narrow-linux.png`, `lifecycle-review-spacing-linux.png`, `lifecycle-review-zoom-linux.png`, `lifecycle-uncertain-linux.png` |
| `web.design.visual.membership_audit_browsing` | Unregistered capture | `membership-audit-comfortable-linux.png`, `membership-audit-compact-linux.png`, `membership-audit-cursor-recovery-linux.png`, `membership-audit-empty-linux.png`, `membership-audit-inspected-linux.png`, `membership-audit-inspected-narrow-linux.png`, `membership-audit-inspected-spacing-linux.png`, `membership-audit-inspected-zoom-linux.png`, `membership-audit-loading-linux.png`, `membership-audit-stale-linux.png` |
| `web.design.visual.membership_management_visual` | Unregistered capture | `membership-management-comfortable-linux.png`, `membership-management-compact-linux.png`, `membership-management-confirmed-refresh-failure-linux.png`, `membership-management-loading-linux.png`, `membership-management-pending-linux.png`, `membership-management-removal-narrow-linux.png`, `membership-management-removal-spacing-linux.png`, `membership-management-removal-zoom-linux.png`, `membership-management-role-linux.png`, `membership-management-uncertain-linux.png` |
| `web.design.visual.metadata_editing` | Unregistered capture | `metadata-closed-linux.png`, `metadata-comfortable-linux.png`, `metadata-compact-linux.png`, `metadata-confirmed-refresh-failure-linux.png`, `metadata-conflict-linux.png`, `metadata-dirty-linux.png`, `metadata-loading-linux.png`, `metadata-review-narrow-linux.png`, `metadata-review-spacing-linux.png`, `metadata-review-zoom-linux.png`, `metadata-saving-linux.png`, `metadata-uncertain-linux.png` |


Post-refresh checks pass: `make json-shape-check` 3/3 at
`20260915T164925Z-p84217`, `make generated-artifact-policy-check` 3/3 at
`20260915T164925Z-p84225`, and `make generate-drift` 4/4 at
`20260915T164925Z-p84233`. The `web.design` task guide was inspected after the
shared workbook visual changes; its browser routes remain catalog-owned.
Final typecheck passes 2/2 at `20260915T164452Z-p24641`; the Entity query slice
passes 2/2 at `20260915T164452Z-p24576`. The delayed-action Grid Adapter slice
passed 2/2 at `20260915T164205Z-p77208`, and Biome passed 2/2 at
`20260915T164254Z-p88944`. These narrow checks cover the terminal fixes after the
628/628 full frontend run.


Two fresh ordinary `make browser-e2e-visual` runs pass against the promoted
manifest, 12/12 each: `20260915T164853Z-p26888` and
`20260915T164853Z-p26895`. Both reconciliation v3 artifacts report all 252
captures/goldens active with zero missing, ambiguous, orphan or unresolved
fixture mappings. Network Analysis's existing visual row passes unchanged.
No assertion, tolerance, viewport or mask was weakened to obtain these results.

### Final scope and next action

Final scope review confirms `main` and HEAD
`129d1d8ba8b57cce08242cc8be3b42bc6accab78` are unchanged. There were no
pre-existing edits to preserve. The working tree contains 253 tracked changes
(including 143 goldens) and 14 new source/test/guide/handoff files. Changes stay
within Workbook/Grid Adapter integration, demonstrated contract projections,
owner text, source/test routing, generated derivatives and regression evidence.
The digest, lockfiles, dependencies, database inputs and binary composition roots
are unchanged. No commit, push or deployment occurred.

The current Timeline catalog has four AC-043 measurement rows; all four passed.
Scrolling and virtualization are additionally exercised by the spreadsheet,
long-traversal, inspector and visual rows. No separate Timeline scroll measurement
row is declared by the current catalog. Full release/backend rewrites and
production publication checks are outside this seam; no conformance claim or
retained full-warm maintenance is asserted.

`git diff --check` passes after the final scope review. The completed handoff
passes `make lint-markdown` at `20260915T165406Z-p1792`; its summary is
`adhoc/lint-markdown/tool-run-summary.json`. WQC-05 is DONE. No applicable
acceptance row is blocked or pending. Next action: review the uncommitted
working tree; implementation stops at this seam.
