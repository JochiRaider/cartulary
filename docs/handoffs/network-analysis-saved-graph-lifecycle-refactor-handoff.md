# Network Analysis saved-graph lifecycle remediation

## Current audit and gap closure

Authorized follow-up baseline: clean `main` at
`03af0570c492997d8466b45eb8f3298e80180600`, revalidated on 2026-09-09.
Root `AGENTS.md` remains the only applicable instruction file. The completed
implementation and all evidence below this follow-up are historical; they do
not establish completion of the new acceptance cases.

| Workstream | Status | Exit |
| --- | --- | --- |
| SG-01 — Reconciliation and characterization | DONE | Owner mappings, compatibility assessment and executable reproductions. |
| SG-02 — Transport and captured operations | DONE | Admission, replay, conflict, typed errors and route checks. |
| SG-03 — Observation and continuity | DONE | Declaration races, authority withdrawal, recovery and last-safe results. |
| SG-04 — UI and browser integration | DONE | Accessible recovery, real workflows and mounting limits. |
| SG-05 — Final validation | DONE | Fresh verification, acceptance and final scope audit. |

### Follow-up authority and compatibility

The localized digest was read in its declared order during planning and its
applicable paths revalidated. ADOPT semantic continuity, local recovery and
keyboard access; ADAPT confirmations and responsive presentation to existing
controls; REJECT redesigns and generic workflow machinery. The digest is
unchanged. `domain.md` supplies vocabulary; `design.md` supplies design direction;
the NLSpec research article supplies no product authority. The current visual
maintenance guide governs any golden maintenance.

| Correction | Adopted owner | Implementation boundary |
| --- | --- | --- |
| Replay uncertainty, immutable attempt context and dispatch fencing | NF-REQ-024–027, 170b, 206; Core 04 REQ-04-023 | Saved operation owner and narrow transport. |
| Historical receipt and declaration/read races | NF-REQ-170a–d, 208; Core 01 §3.3.9.1 | Saved controller and navigation. |
| Protected exposure and functional recovery | NF-REQ-022, 170d–e, 207–208 | Saved read outcomes, observation and presentation. |
| Error classification and query admission | NF-REQ-021, 173–176; Core idempotency/error contracts | Saved transport and saved HTTP routes only. |
| Accessible errors and distinct async states | NF-REQ-170e; design §§12.5, 12.8, 14 | Existing saved-graph panel and browser evidence. |

Current Network Flow major 6, declaration/receipt v4, all seven selected-binding
members, request versions, narrow job references and empty 204 retirement/replay
are retained. Private `SelectedResult` storage naming is a legitimate internal
projection, not a public alias. Publication can replace the binding without
incrementing declaration version; historical receipts are never current read
authority. Graph Projection v2 identity and Reporting lease ownership remain
unchanged. No owner contradiction was found in the reviewed seam.

No stored-data migration or additional public projection change is planned.
Rejecting undeclared saved-route query parameters tightens previously permissive
invalid-input handling under the existing owner contract; valid requests retain
their contract. No route, dependency, compatibility alias, browser persistence,
worker, scheduler or cutover change is authorized. Rollback is a coherent
follow-up code/test/routing/golden revert within major 6, not a rollback to the
pre-cutover release. No reset, commit, push or deployment is authorized.

Planning evidence: frontend operation/transport/continuity 4/4 in
`20260909T160716Z-p69331`; observer/local-recovery/lifecycle 4/4 in
`20260909T161108Z-p71520`; backend receipt integrity 1/1 in
`20260909T160842Z-p70278`. These passed baseline tests do not cover the newly
identified gaps. Both required owner task guides passed again at follow-up entry.

### Follow-up SG-01 exit

Executable characterizations added in `SavedGraphRecovery.test.ts` demonstrate
all nine controller boundaries: failed replay, historical publication regression,
selection retargeting, retired conflict review, concurrent declarations, dropped
invalidation, list denial, contributor denial and queued dispatch. All nine fail
at their intended assertions in `20260909T162151Z-p95760` (harness 1/2).
Transport classification and accessible name validation fail in
`20260909T162059Z-p77678` (1/4); that first run also exposed a test syntax typo,
corrected before the nine controller assertions were rerun.

`make service-backed-test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` fails at saved-query
admission in `20260909T162101Z-p77921` (2/3). Runtime remains unchanged at this
checkpoint. Authored routing was updated through `make generate`, passing in
`20260909T162019Z-p74641`. The unchanged result-retry callback also has a traced
no-op after navigation identity withdrawal; integration coverage will exercise
its declaration-read prerequisite when the recovery action is added.

SG-01 exit: owner mapping, compatibility assessment and executable failures
recorded before runtime changes. No structural movement is planned. Next action:
SG-02 transport validation, immutable attempt context and saved-query admission.

### Follow-up SG-02 exit

Corrected captured attempt revisions, replay uncertainty, historical receipt
admission, queued dispatch revalidation, exact error-code classification and
validated mutation rejection envelopes. Read status and active-target checks
were tightened without changing wire schemas. Saved HTTP query rejection now
runs after authentication and the exact per-action role/lifecycle gate. Only
`graph_view_routes.go` and its lifecycle fixture changed on the backend.

`make test-slice OWNER=web.networkflow` with the captured-operations,
receipt-transport and replay-admission rows passes 4/4 in
`20260909T163017Z-p24882`. This includes the real availability request queue.
The saved lifecycle service row passes 3/3 in `20260909T162849Z-p3585`, including
all eight saved-route query gates and role precedence. `make frontend-typecheck`
passes 2/2 in `20260909T162930Z-p21247`. Public routing generation passes in
`20260909T162955Z-p21925`; the remaining recovery cases retain their own row.
The earlier combined 3/4 run (`20260909T162839Z-p2799`) now fails only at the five
SG-03 read/authority assertions. No structural code movement occurred.

SG-02 exit passed. Next action: SG-03 per-declaration read generations,
central read-failure disposition and functional binding recovery.

### Follow-up SG-03 exit

The operation owner now fences independent declaration reads and conflict review,
retains invalidations through mutation admission, and reconciles historical
receipts without replacing current declarations. `savedGraphReadFailure.ts`
centralizes typed errors and scoped disposition across declarations, list,
results, contributors and Common Job observation. Missing jobs and transient
failures retain authorized bytes; explicit access loss clears exposure. A
collaboration removal additionally requires fresh authority before reads resume.
The navigation owner restarts interrupted initial reads and result recovery
reconciles a withdrawn binding before issuing its result request.

Six focused frontend rows pass 7/7 in `20260909T163604Z-p35532`, including the five
new binding recovery, interrupted-load, contributor transport/stale-binding,
typed job-error and late-receipt tests. The previous 6/7 run
`20260909T163341Z-p30638` caught the collaboration-removal recovery distinction;
its original tombstone/access test was preserved and the implementation corrected.
Typecheck passes 2/2 in `20260909T163342Z-p30915`; public generation passes in
`20260909T163535Z-p32569`. The immutable seven-member identity, observer bounds,
100-row contributor pages and 500/1,000 mounting limits remain intact.

SG-03 exit passed. Next action: SG-04 honest local states, accessible name errors
and controlled browser recovery/race evidence in the real application.

### Follow-up SG-04 exit

Saved surfaces now distinguish initial list/result failures, confirmed emptiness,
materialization failure and binding withdrawal. Recovery labels follow their
prerequisites. Name invalid state and local error associations retain byte-based
NFC validation and drafts. Controls, density, layout and mounting bounds are
preserved. Protocol envelope validation belongs in the existing service adapter;
feature files do not import protocol facades directly.

Focused UI/lifecycle/bounds/roles pass 5/5 in `20260909T163804Z-p41340`.
The new real-application deferred navigation test, existing saved lifecycle,
a11y (including narrow/short layouts, zoom, Escape and name descriptions), and
measurement rows pass in `20260909T164043Z-p46716` (combined run 14/16).
Its new replay/read-recovery row exposed lost creation selection when a newer
reconciliation superseded the first. Selection intent now survives superseding
reads and remains fenced by the original user-selection revision. That browser
row passes 11/11 in `20260909T164411Z-p95684`.

Current local recovery/receipt/declaration tests pass 4/4 in
`20260909T164829Z-p40537`; five operation/continuity rows pass 6/6 in
`20260909T164548Z-p33382`. Typecheck passes in `20260909T164654Z-p34682` after
correcting an unsupported test-only locator option. Import boundaries pass 2/2
in `20260909T164828Z-p40194` after the adapter correction (the earlier boundary
failure was `20260909T164655Z-p35116`). New browser routes are controlled test
responses over actual committed operations; no production route was added.

SG-04 exit passed. Next action: SG-05 finalizer, affected owner verification,
ordinary visual validation and final acceptance/scope audit. Retained-run
maintenance will be skipped with `RESULTS_DIR` unset: no qualifying exact-source
successful full warm check exists for this follow-up.

### Follow-up SG-05 validation and acceptance

`make agent-finalize` passed before broader verification in
`20260909T164912Z-p41423`. Its generated-structure, JSON shape and catalog coverage
actions passed. Canonical retained-evidence and scheduler maintenance were
explicitly skipped (`results-dir-not-provided`); `RESULTS_DIR` remained unset.
The final audit made the private transport dispatch guard mandatory at its type
boundary; current callers already supplied it. Its receipt tests and typecheck
were rerun on that final signature.

All run IDs below are under `.cartulary/test-results/`; their row results,
unit results and target summaries retain the executable evidence.

| Command | Result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=web.networkflow` | PASS 49/49 | `20260909T165000Z-p44974` |
| `make test-slice OWNER=module.networkflow` | PASS 35/35, all 81 owner rows | `20260909T165317Z-p62951` |
| `make service-backed-test-slice OWNER=module.networkflow` | PASS 29/29, all 34 service-backed rows | `20260909T165000Z-p45066` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.extension_availability_lifecycle_792bc604af` | PASS 2/2 | `20260909T165000Z-p45060` |
| `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.saved_graph_replay_admission,web.networkflow.regression.saved_graph_recovery_boundaries` | PASS 3/3 after test lint corrections | `20260909T165318Z-p64720` |
| `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.saved_graph_receipt_transport` | PASS 2/2, mandatory dispatch guard | `20260909T165705Z-p21683` |
| `make frontend-typecheck` | PASS 2/2, final signature | `20260909T165704Z-p21311` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260909T165219Z-p30730` |
| `make lint-biome` | PASS 2/2, final signature | `20260909T165744Z-p22455` |
| `make generate-drift` | PASS 4/4 | `20260909T165219Z-p30536` |
| `make go-gosec-targeted` | PASS 4/4 | `20260909T165032Z-p81072` |
| `make browser-e2e-visual` | PASS 12/12 | `20260909T165305Z-p35663` |
| `make lint-markdown` | PASS complete follow-up handoff before DONE marker | `20260909T165913Z-p23221` |

The first final lint run (`20260909T165000Z-p45268`) rejected three non-null
assertions in new tests; these were replaced with explicit checks or a failing
lookup. No lint rule was relaxed. A planning invocation of `target-plan` with
unsupported `OWNER` failed usage validation; `make explain-test-owner
OWNER=module.networkflow` supplied the routing inventory. Neither was a runtime
or service failure.

The owner runs include saved/temporal lifecycle and receipt integrity, exact
replay, roles, quota/version admission, leased-result cleanup, source deletion,
protected browser state, imports, retained import drafts, unsaved exploration,
a11y and both graph and contributor bounds. The complete visual reconciliation
is PASS: 210/210 active captures, zero missing goldens, ambiguous mappings or
orphans. The promoted manifest remains unchanged at SHA-256
`3ed785363470b97580ba6047197abaab9e875edf8d2ad1c8f47139c6ac23f168`.

Reviewed unchanged images:
`network-flow-analysis-saved-graph-result-linux.png` (1440×900) and
`network-flow-analysis-compact-saved-graphs-linux.png` (768×640), under the
committed visual snapshot directory. Controls, focus framing, density and layout
are preserved. No golden, capture preparation, mask, anchor, tolerance, renderer
or shared preference changed. Golden promotion and its two subsequent ordinary
passes were unnecessary because no golden changed. Existing owner visual passes
also succeeded in both full module runs.

| Follow-up binary acceptance | Status | Evidence |
| --- | --- | --- |
| Owner and wire reconciliation; major 6/v4/204 retained; query admission corrected after authorization | PASS | SG-01 mapping, saved route lifecycle and receipt suites. |
| Immutable attempts, synchronous admission, required dispatch guard, exact replay and honest uncertain recovery | PASS | Operation/replay/transport rows; real committed-create replay recovery. |
| Current reads and retirement tombstones fence conflict review, historical receipts and independent reconciliations | PASS | Recovery row and deferred real-application navigation. |
| Accepted writes remain acknowledged across subsequent list/job/result failures | PASS | Captured-operation, typed observation and last-safe-result tests. |
| Seven-member authorized continuity preserves results, contributors and bounded navigation across unchanged metadata | PASS | Semantic continuity, lifecycle, a11y and measurement rows. |
| Session/incident/claim loss, target absence and binding/source invalidation have scoped withdrawal and usable recovery | PASS | Recovery, protected-state, lifecycle and availability integration rows. |
| Initial/retained failure states, NFC/UTF-8 drafts, roles, duplicate names, keyboard access and responsive presentation | PASS | Local recovery/role/bounds rows; real lifecycle/recovery/a11y tests. |
| Generation, type/import boundaries, security, owner regression and ordinary visuals | PASS | Fresh SG-05 command table. |
| Final handoff bytes, whitespace and scope audit | PASS | Markdown run `20260909T165913Z-p23221`; `git diff --check`; explicit 22-path scope audit. |

Actual scope: 22 authorized paths. Runtime changes stay in the saved-graph
controller, navigation, error/observation helpers, transport/adapter, owner hook,
presentation and saved-route query gates. Supporting changes are focused tests,
authored owner routing and generator-produced browser/topology projections.
No structural workflow extraction or generic engine was introduced. The original
handoff body remains byte-for-byte intact below this follow-up.

Actual compatibility impact: unsupported saved-route query input now receives
`network_flow_invalid_request` (400) after authorization. Invalid response
payloads are treated as uncertain or invalid reads rather than trustworthy
acknowledgements. Valid major-6 requests and existing stored data retain their
schemas and lifecycle. No migration, new route, compatibility alias, browser
persistence, dependency or owner revision is needed. Roll back the follow-up
runtime, tests and authored routing together and regenerate their projections
within major 6; the pre-cutover baseline is not a compatible rollback.

Limitations and skipped checks: client observation remains bounded and cannot
prove cancellation or failure after timeout. Recovery uses existing current-read
and exact-replay routes. Full-repository `make check` and release/deployment gates
were not run; affected owner, service, browser, generation and security checks
cover this change. Retained-run maintenance was skipped as described above. No
reset, commit, push or deployment occurred.

SG-05 exit passed: all current tracker rows are DONE and all follow-up binary
acceptance criteria are PASS. The Markdown, whitespace and explicit 22-path
scope checks are repeated against these final handoff bytes after the DONE
marker, as required; their terminal results remain in the session evidence.
Next action: none.

## Execution contract and baseline

The approved implementation coordinates Network Flow contract major 6. Network
Flow owns declarations and workflow; Graph Projection owns immutable results;
Common Jobs owns execution; Reporting owns leases. No stored-data migration,
compatibility decoder, alias, browser persistence, route, dependency, commit,
push, or deployment is authorized by this work.

Baseline: clean `main` at `88319c90cff44d55b2fc003215a8294c28f0d660`, ahead of
`origin/main` by 17 commits. Revalidated before editing on 2026-09-09. Root
`AGENTS.md` applies. The approved plan explicitly resolves the previously
blocking Core 01 / Network Flow worker and reference-registry contradictions.

## Tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| SG-01 — Owners, contracts, characterization | DONE | Coherent owners, compatibility assessment, owner map, focused reproductions. |
| SG-02 — Transport and captured operations | DONE | Receipt, admission, replay, conflict, lifecycle tests pass. |
| SG-03 — Observation and result continuity | DONE | Race, safe-result, invalidation, observation and contributor tests pass. |
| SG-04 — UI and browser evidence | DONE | Accessible workflows and bounded rendering pass. |
| SG-05 — Cutover and final validation | DONE | Compatibility fixtures, required checks and scope audit pass. |

## Authority and advisory classification

Read the localized digest in its README read order. Its paths were checked
against the current repository; it remains advisory and unchanged. ADOPT local
recovery, semantic identity, keyboard access, and authorized-result continuity.
ADAPT responsive and confirmation guidance to existing Cartulary controls.
REJECT re-theming, generic workflow prescriptions, and compatibility aliases.
Completed chrome, import recovery, and visual reliability handoffs are
implementation evidence only.

| Gaps | Adopted owner and change boundary |
| --- | --- |
| G1, G3, G5 | Core 01 §3.3.9 and REQ-01-634; Extensions reference grammar and finalization; Network Flow NF-REQ-170c. |
| G2, G4 | Network Flow §5, §19.1, §21, §25, §27; authored Network Flow schemas/routes/errors; public generation. |
| G6, G7 | Network Flow NF-REQ-054a, NF-REQ-170b/e; Core authorization, CSRF, and route idempotency. |
| G8 | Network Flow §8.6, §16, NF-REQ-170b–d; incident admission and Common Job publication. |
| G9, G10, G11 | Core common-job reads; Network Flow stale-response, source lifecycle, binding and collaboration contracts. |
| G12, G13 | Network Flow NF-REQ-170e; design local errors/dialogs/focus; current visual maintenance guide; authored verification routing. |
| G14 | Extensions admission; Network Flow compatibility/state ownership; Auth receipt and Jobs proof ownership. |

Domain vocabulary and design direction remain owned by `docs/domain.md` and
`docs/design.md`. Graph Projection algorithms, Reporting leases, imports,
table management, scheduling and unsaved exploration are outside remediation
except focused regression evidence and demonstrated boundary corrections.

## Compatibility and rollback decision

Advance the public Network Flow major to 6 and declaration/embedding schemas
to v4. Advance changed request IDs only. Preserve semantic-query v2, Graph
Projection v2, deterministic identities, private storage columns, and compatible
durable-state version 4. Compatibility is conditional on read-only admission
preflight, not on the existence of valid declarations alone. Incompatible
retained receipts, jobs/proofs, names, or failure values block admission without
modification. Route-idempotency records have no expiry check; waiting for job
retention expiry does not establish upgrade compatibility.

A rejected cutover leaves prior-release state untouched. After major-6 writes,
prefer a forward fix; rollback needs a compatible release or separately
authorized restoration of a pre-cutover backup. Do not delete receipts, shorten
retention, rewrite history, or silently normalize stored data.

## Binary acceptance criteria

Every criterion passed its owner-routed executable checks. The workstream records
below identify commands, artifacts, corrections and compatibility limits.

| Criterion | Status | Evidence |
| --- | --- | --- |
| G1: Exact, unique worker/reference registrations and coherent owners | PASS | Extensions contract/coordinator tests; complete owner slice `20260909T153739Z-p63489`; generation drift. |
| G2: Complete v4 declaration through every public projection | PASS | Receipt, declaration-store, adapter and protocol tests; Network Flow owner `20260909T152625Z-p46280`, protocol `20260909T152710Z-p32875`. |
| G3: Integrity-checked declaration plus narrow job reference | PASS | `graph_view_receipts_test.go`, `savedGraphTransport.test.ts`; backend lifecycle and frontend owner slices. |
| G4: Exact requests, Unicode names, errors, no-op rename and empty 204 replay | PASS | Name/receipt unit tests, transport tests and saved lifecycle service tests, including `20260909T154536Z-p94448`. |
| G5: Immutable original receipts across lifecycle and job expiry | PASS | Saved lifecycle/replay, shared finalizer and cutover service fixtures; `20260909T154536Z-p94448`, `20260909T134146Z-p9520`. |
| G6: Synchronous admission and immutable action context | PASS | `SavedGraphController.test.ts`; frontend owner 47/47 in `20260909T153029Z-p78693`. |
| G7: Honest uncertainty, exact replay, conflict review and preserved drafts | PASS | Captured operation and transport tests, including late acknowledgement; local recovery browser evidence. |
| G8: Transactional roles, lifecycle, source and publication enforcement | PASS | Full Network Flow service slice `20260909T153111Z-p97662`; saved lifecycle/cutover rerun `20260909T154536Z-p94448`. |
| G9: Bounded, validated observation and explicit resume | PASS | `savedGraphObservation.test.ts` and controller observation-slot tests; frontend owner slice. |
| G10: Semantic result continuity and stale-list fencing | PASS | `SavedGraphResultNavigation.test.ts`, controller tombstone/race tests and browser rename continuity. |
| G11: Collaboration invalidation and bounded contributor fencing | PASS | Collaboration interpreter/outbox tests, deferred contributor/page tests and source-withdrawal tests. |
| G12: Local recovery, keyboard focus and accessible bounded viewer | PASS | Real lifecycle/accessibility/measurement 16/16 in `20260909T153925Z-p4165`. |
| G13: Complete regression, browser, accessibility and visual evidence | PASS | Affected owner suites and preserved import/unsaved workflows; two required ordinary visual passes plus final `20260909T153007Z-p19628`. |
| G14: Read-only compatibility admission and demonstrated rollback limits | PASS | Cutover byte-preservation fixtures `20260909T154536Z-p94448`; Extensions preflight ordering; Jobs retained-summary tests `20260909T154452Z-p76294`. |

## SG-01 evidence and checkpoint history

Baseline discovery passed for `web.networkflow` and `module.networkflow` using
`make task-guide ROLE=module-author OWNER=<owner>`. Additional confirmed owners:
`module.jobapi`, `platform.jobs`, `module.extensions`, `package.protocol_ts`,
`platform.openapi`, `web.workbook`, `package.ui`, `web.design`, and
`harness.browser`. Extension-store tests route through `module.extensions`.

The four existing saved-graph frontend rows passed (five total harness units)
in `.cartulary/test-results/20260909T123427Z-p12655`. This is baseline
characterization, not final evidence. `git diff --check` passed before edits.

Verified gaps include incomplete declaration serialization, retirement 200,
finalizer receipt replacement and replay reconstruction, scalar-based names,
React-only mutation admission, unbounded declaration polling, object-reference
result reloads, and contributor publication without graph fencing. The approved
G1–G14 plan records their required behavior; implementation must characterize
and validate each boundary independently of documentation.

At this checkpoint, implementation and final verification belonged to successor
workstreams. Their completed records below distinguish behavior corrections from
structural movement and identify the reviewed visuals. Retained-run maintenance
was deliberately skipped because qualifying full warm evidence was unavailable.

Checkpoint successor: contract-correct receipts, transactional admission and
captured frontend operation ownership in SG-02.

### SG-01 focused characterization

These reproducible triggers were traced end to end in the inspected baseline;
the existing lifecycle tests characterize the happy path but do not establish
the missing race and recovery guarantees. Later workstreams add executable
regressions for these triggers before claiming their binary criteria complete.

| Gap | Trigger and demonstrated baseline seam |
| --- | --- |
| G1 | A successful saved materialization emits a graph reference and uses a Network Flow worker forbidden by the old Core clauses; machine job facts were in `core01.profile-jobs.json`. |
| G2 | List/get/mutation/result serialization in `graph_view_routes.go` drops persisted query digest, desired source, creator and failure time; seven-member `selected_result` is the binding under the wrong public name. |
| G3 | Create/refresh returns top-level job ID/kind rather than a narrow reference, exposing job-family identity and omitting its canonical read route. |
| G4 | Retire returns 200 with a graph envelope; multibyte names are counted as scalars; unchanged rename increments version; retired access and quota reasons differ from the owner tables. |
| G5 | Finalize create, then rename and replay the creation: `finalizer.go` replaces the receipt; replay reconstructs from current declaration and retained handler payload instead of returning original data. |
| G6 | Activate twice in one tick or change selection while confirming: hook admission relies on asynchronously published React state and dispatch reads the current selection. |
| G7 | Lose create acknowledgement or fail the follow-up list: client functions generate transaction IDs internally and the hook collapses mutation and follow-up failures. |
| G8 | Close an incident before graph mutation/publication: shared route checks use `LifecycleAny` and saved publication lacks the required explicit incident-open check. |
| G9 | Leave a queued declaration selected: hook list polling repeats without an observation window, current Common Job validation, or explicit resume. |
| G10 | Rename or poll an unchanged declaration while paging: selected declaration object replacement retriggers result loading and clears navigation; loading hides the prior result. |
| G11 | Defer contributors for A then select B: graph switching does not advance the contributor generation; collaboration interpreter discards graph resource events and contributor UI exposes no continuation. |
| G12 | Close/reopen a failed create dialog: panel-local drafts disappear; pending-state errors and workspace banners do not provide attempt-specific recovery. |
| G13 | Existing decoder/lifecycle, role and mounting tests pass without uncertain-write, stale-list, graph-switch and authority-transition coverage. |
| G14 | Retain an old receipt after job expiry: Auth lookup has no expiry predicate, while the profile has no saved-graph compatibility preflight. |

The authored changes separate behavior corrections from ownership movement:
Core/Extensions receipt and registry clauses, NF major-6 declaration/requests,
retirement/name/error rules and read-only cutover are behavior corrections;
moving unchanged job/worker facts into the Network Flow fragment is ownership
movement. Core Import scheduling, Graph v2 algorithms, Recovery v4 and Reporting
lease contracts have no semantic changes. The OpenAPI Common Job reference is
an additive shared schema recorded in the existing authored 2.0.0 change set;
the separate Network Flow projection has the incompatible major-6 cutover.

Generation initially rejected an undeclared additive OpenAPI schema, then an
unsorted algorithm registration and an old Import owner-contract reference.
All were corrected in authored inputs. Shape validation exposed the v3-only
schema-ID pattern and failure-pair applicators incorrectly treated as new open
objects. The closure validator now carries closure only into same-instance
applicators; nested properties remain independently closed. Additional owner:
`harness.generated_artifacts` (`make task-guide` passed).

Evidence so far: `make generate` passed in
`.cartulary/test-results/20260909T130838Z-p38164`; the focused Network Flow
projection row passed in `.cartulary/test-results/20260909T130400Z-p29638`.
The Extensions projection row passed in
`.cartulary/test-results/20260909T130400Z-p29689`, with strengthened duplicate
ownership assertions awaiting its fresh rerun. `make lint-markdown` passed in
`.cartulary/test-results/20260909T130646Z-p35810`; final bytes still require the
prescribed final rerun. `git diff --check` passed during reconciliation.

SG-01 exit passed: `make json-shape-check` passed 3/3 units in
`.cartulary/test-results/20260909T130859Z-p41136`; the strengthened Extensions
projection row passed in `.cartulary/test-results/20260909T130859Z-p41177`.
Owner reconciliation, compatibility restrictions, exact no-change parity and
the focused trigger map are recorded above. Public generation ran through Make;
no generated output was hand-edited. At the SG-01 checkpoint, runtime still
required the SG-02 receipt/resource corrections before major-6 admission could
serve traffic.

## SG-02 execution

DONE. Work began at `graph_view_routes.go`, `graph_view_store.go`,
`graph_view_jobs.go`, the transaction-bound finalizer/receipt adapters and saved
graph browser transport/controller. Protect original receipts and transactional
authority before adding materialization observation or browser evidence.

Backend progress: declarations now serialize v4 without querying Jobs for copied
status; accepted receipts contain the narrow reference; retire/replay are
bodyless 204; names enforce NFC/control/whitespace/64-byte rules; unchanged
normalized rename preserves declaration/audit identity. Mutation comparison
frames the adopted route/path/body bytes. Saved mutations revalidate roles and
open-incident lifecycle transactionally, and publication rechecks lifecycle.
The finalizer now calls an owner receipt-reconciliation port. The Network Flow
adapter validates the original receipt without updating it, while Core adapters
retain terminal-receipt updates. Removed the unused retained-handler-payload
replay API and Network Flow reconstruction fallback.

`make test-slice OWNER=module.networkflow
ROWS=module.networkflow.unit.saved_graph_receipt_integrity` passed in
`.cartulary/test-results/20260909T132218Z-p57439`. An earlier compile failure
was a missing pgx import in the expanded incident admission port; corrected.
Service-backed saved lifecycle and temporal lifecycle rows passed in
`.cartulary/test-results/20260909T132705Z-p81674` (3/3 harness units). The
declaration-store row passed in the preceding mixed run
`.cartulary/test-results/20260909T132444Z-p63047`; its two route rows initially
failed because the runtime Import facade still named major 5. Corrected the
Network Flow binding and its test projection to major 6; Import behavior is
unchanged. Additional owner `module.imports` task-guide discovery passed.
The lifecycle test compares complete original receipt data after terminal
success, rename, retirement and logical job expiry, and tests no-op rename and
both empty retirement responses. Further failure/cancellation, role, conflict,
browser transport and captured-operation coverage remains required for SG-02.

Browser transport now accepts immutable operation attempts, validates exact
status/target/version/job-reference relationships, and treats malformed success
or lost transport as uncertain. The saved-graph operation owner synchronously
locks admission, captures confirmation context, retains drafts and exact replay
bytes, requires conflict review, fences obsolete reads, and keeps committed
receipts independent of follow-up reads. Its workbook lifetime binding uses the
application session lifetime and current incident/role/extension authority.
The panel and result adapter integration remain successor work.

Focused captured-operation tests passed in
`.cartulary/test-results/20260909T134144Z-p9293`. Receipt transport plus preserved
Import recovery passed in `.cartulary/test-results/20260909T134529Z-p36360`.
Transport tests caught an authored protocol entrypoint still using v3 validator
exports; updated it to the generated v4 validators. The protocol conformance
row passed in `.cartulary/test-results/20260909T134729Z-p73253` after correcting
its major literal. Shared finalizer success/failure/cancellation service tests
passed in `.cartulary/test-results/20260909T134146Z-p9520` (3/3 units). Receipt
integrity, failure/cancellation validation, and temporal/deadline unit coverage
passed in `.cartulary/test-results/20260909T134747Z-p75542`.

Generation passed in `.cartulary/test-results/20260909T134319Z-p27389` after
sorting the authored new test titles. Formatting passed in
`.cartulary/test-results/20260909T134744Z-p73876`; earlier new-source lint findings
were fixed without altering unrelated informational suggestions. The expanded
role/lifecycle service fixture initially addressed the incidents primary key as
`incident_id`; corrected the fixture to `id` and the reopened state to `active`.


SG-02 exit passed: expanded saved lifecycle service checks passed 3/3 units in
`.cartulary/test-results/20260909T134920Z-p96119`. The fixture closes and reopens
with the owner-required `closed_at` pairing. Each current member can read;
editor/admin and reviewer/admin mutation gates precede request decoding;
closed incidents reject reads and all four mutations. Pre-computation and
transactional publication both check incident lifecycle. `git diff --check`
passed. Added owner `web.application` task-guide discovery passed for threading
the accepted application session lifetime into the workbook.

## SG-03 execution

DONE. This workstream added Common Job observation through reusable timing/read
mechanics, semantic result and contributor fencing, collaboration invalidation,
and deterministic race/recovery evidence. Existing shared import observation
coverage passed after extracting the neutral bounded-read primitive. No server
scheduling or observation-driven cancellation is introduced.

SG-03 progress: the workbook operation owner now composes a serial bounded
Common Job observer and a separate result-navigation port. Result exposure uses
the full seven-member binding, preserves object identity and navigation through
metadata polling, and fences result/contributor reads across selection and
protected-scope changes. Contributor navigation retains one page and opaque
continuation state. Source invalidation withdraws exposure immediately; removed
source IDs and retired declaration IDs cannot be resurrected by stale reads.

Common Job validation and bounded-read/clock mechanics moved to neutral service
utilities. Import-family decoding remains in Imports; existing import recovery
assertions were preserved and pass. Deferred-response, metadata continuity,
initial versus last-safe failure, paging context, source invalidation, exact
replay, 30-second read expiry, 120-second observation expiry/resume and stale
job progression tests passed in
`.cartulary/test-results/20260909T140746Z-p81283` (5/5 units).

Network Flow now appends declaration lifecycle invalidations transactionally.
Service-backed lifecycle, declaration persistence/publication and immutable
result cleanup/lease races passed in
`.cartulary/test-results/20260909T140444Z-p51438` (5/5 units). These tests initially
exposed Collaboration's closed reason switch contradicting Core 01's explicit
future-additive invalidation rule; the validator now admits additive reasons
only for invalidation. Retirement uses the existing `remove/soft_deleted`
meaning for a retained inactive declaration. No new event family was added.
Core 01's historical Network Flow v1 resource-kind sentence now identifies the
current major-6 registered resources. Added owner `module.collaboration`
task-guide discovery passed; its additive invalidation/removal rejection tests
passed in `.cartulary/test-results/20260909T140635Z-p74084`.

The service-backed lifecycle now asserts seven exact graph outbox effects,
including one changed-name rename and one removal despite replay and no-op
rename; it passed in `.cartulary/test-results/20260909T140847Z-p83004`.
Collaboration interpretation passed in
`.cartulary/test-results/20260909T140759Z-p82237`. Shared Import lifecycle and typed
transport passed in `.cartulary/test-results/20260909T141011Z-p2427`; the first
extraction added a microtask before dispatch and the existing pause test caught
it. The neutral primitive now preserves synchronous dispatch admission while
bounding the awaited response. No import assertion was weakened.

Observation concurrency uses the effective incident graph-job limit from the
existing source-profile route, with conservative single-observation admission
when that read is unavailable. Workbook-lifetime collaboration subscription
keeps retirement/source invalidation effective while other sheets are active.
Result transport rejection that removes binding authority clears exposure;
ordinary transport failure retains the authorized last result. The latest
operation/observation/navigation slice passed in
`.cartulary/test-results/20260909T140944Z-p1505` before the final authority-recovery
check. Remaining SG-03 exit: fresh focused rerun after that check.


SG-03 exit passed: the final authority recovery and source-withdrawal checks
passed in `.cartulary/test-results/20260909T141154Z-p7832` (4/4 units).
The earlier five-unit service run and fresh outbox-count run establish lifecycle
publication and retained-result/lease behavior. No pending observation is treated
as server cancellation, failure, or materialization success. `make format`
passed in `.cartulary/test-results/20260909T141135Z-p3536`; `git diff --check`
passed. Next: SG-04 presentation and real-application browser evidence.

## SG-04 execution

DONE. This workstream integrated captured dialogs, local recovery, observation
status, semantic navigation state and contributor paging. Existing controls,
tokens, density, layout, mounting bounds, import and exploration assertions remain.

### SG-04 integration and visual review history

The workbook now owns saved operations and navigation across surface changes.
The hook subscribes; the panel presents captured dialogs, local conflict review,
exact replay, observation resume, result/contributor recovery and bounded pages.
Names use normalized UTF-8 byte validation. Contributor grouping and Escape
preserve focus. Confirmation focus restoration now keeps native controls in the
Tab sequence and falls back to Reload when retirement removes the trigger.

Integration exposed a distinction between Extensions request reservations and
accepted authority. `extensionAvailability.ts` now exposes an authority tag and
synchronous invalidation subscription without changing the request sequencing
contract. Saved graphs capture that authority tag, including workbook availability;
ordinary requests no longer invalidate their own completion. Its regression row
passes in `.cartulary/test-results/20260909T142818Z-p30651`.

The complete Network Flow frontend slice passed 47/47 units in
`.cartulary/test-results/20260909T143444Z-p80476`. The local recovery and lifecycle
rows passed 3/3 in `.cartulary/test-results/20260909T143121Z-p68051`. The real-server
lifecycle and mounted-result ceiling rows passed 14/14 in
`.cartulary/test-results/20260909T143123Z-p68306`. Accessibility passed 11/11 in
`.cartulary/test-results/20260909T143315Z-p14083`, including Unicode byte boundaries,
retained drafts, narrow/short dialogs, zoom, spacing, rename continuity, Escape,
and focus restoration. Subsequent small focus and announcement corrections
require fresh final evidence. Import-boundary checking passed in
`.cartulary/test-results/20260909T143255Z-p4087` after moving the shared read-state
type out of the controller/navigation import cycle.

Ordinary visual reconciliation in
`.cartulary/test-results/20260909T143316Z-p14381` accounted for all 210 captures and
210 active goldens, with zero missing, orphan, ambiguous or unresolved fixtures.
All 37 ordinary workbook scenarios passed. The Network Flow scenario completed
functional assertions and found two intentional screenshot differences:
`network-flow-analysis-saved-graph-result-linux.png` and
`network-flow-analysis-delete-dialog-linux.png`. The new local acknowledgement
persists across surface switches, so the saved result background remains stable
behind the existing table dialog. Actual, expected and diff images were inspected;
the dialog itself is unchanged. No viewport, scope, mask, anchor, renderer,
density, preference, or tolerance change is proposed.

Accepted refresh trigger: adopted saved-graph acknowledgement/result continuity.
Owner row: `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`.
The two captures are active nonregistry captures; no registered fixture ID is
claimed. Update only through `make browser-e2e-visual-update`, inspect all promoted
changes, then obtain two fresh ordinary passes. Checkpoint successor: complete this visual
workflow and record SG-04 exit before cutover implementation.

### SG-04 exit

DONE. The public visual update passed 12/12 units in
`.cartulary/test-results/20260909T143914Z-p94228`, promoting exactly the two PNGs
listed above and the complete golden manifest. Both promoted images were reviewed.
The created/refresh acknowledgement now states only the confirmed write; execution
status remains separate. Two fresh ordinary visual runs passed 12/12:
`.cartulary/test-results/20260909T144334Z-p96108` and
`.cartulary/test-results/20260909T145153Z-p10651`. An intervening ordinary run
(`20260909T144724Z-p69443`) passed Network Flow but failed an unrelated entity-mention
functional assertion before comparison. That fixture expected a dismissed mention
after resolving and dismissing it; it briefly observed the resolved chip instead.
No golden, tolerance, timing, fixture or product change was made for that failure.
The subsequent fresh ordinary run passed all scenarios.

The new retirement-focus assertion exposed a Reload control disabled by its
follow-up read. Reload now uses accessible busy/disabled state while retaining
native focus; activation is ignored during the pending read. The complete real
lifecycle, including retirement focus, passed 11/11 in
`.cartulary/test-results/20260909T144139Z-p62950`. Fresh accessibility passed in
`20260909T143915Z-p95588`; that combined run failed only the retirement-focus
assertion subsequently corrected above. Import recovery, retained drafts, existing
unsaved workflows and protected-scope browser regressions passed 13/13 in
`.cartulary/test-results/20260909T144424Z-p30274`.

Workbook integration rows passed 3/3 in `20260909T143732Z-p87185`; protocol rows
passed 7/7 in `20260909T143744Z-p88281`; selector contracts passed 2/2 in
`20260909T144711Z-p68510`. Frontend type checking passed in
`20260909T144726Z-p70945`. Final controller/observer/navigation rows passed 4/4 in
`20260909T145117Z-p5477`, including admission-slot reuse after paused observation.
Observation history now retains only current declaration and acknowledged-attempt
jobs, avoiding indefinite accumulation of superseded statuses. No goldens were
changed outside the two reviewed saved-graph backgrounds.

## SG-05 execution

The SG-05 entry scope was the declared read-only saved-graph cutover preflight
before Extensions migrations or readiness: complete declarations and retained
Auth receipts, Common Jobs admission state and Extensions proofs without rewriting
bytes, exact Extensions admission context/result validation and all-profile
preflight ordering. Additional owner guides: `module.auth` and `app.server`;
there is no active `platform.authn` test owner. The fresh, compatible, incompatible
and expired-job cases are completed below.


### SG-05 cutover implementation and narrow evidence

Behavior correction: startup now invokes the declared preflight before extension
state initialization or migration, and the declared post-migration validator
before publication. The all-profile phase reads state presence, metadata and
migration identities first. Its complete logical context contains a safe
configuration-key file handle, never an absolute path or resolved key material.
Owner errors, malformed results, panic and deadline expiry become the exact
shared startup finding and exit 2. No admission callback can commit a mutation.

The new Network Flow validator reads declarations, Auth route receipts, retained
Common Job metadata and Extensions proofs through their owners. It uses
repeatable-read, database-enforced read-only transactions and pages of 128.
It checks currently selected immutable bindings against Graph Projection's
metadata-only reader. Historical receipts are checked against original request,
actor, incident, target and job identities; they need no current declaration or
old result. Compacted job tombstones preserve proof validation without restoring
erased payloads. The serving-process exclusion and a coordinated, quiesced
cutover are required; mixed-major writers are unsupported.

A newly reproduced generator defect read absent locator fields from typed
admission facts and emitted null algorithm IDs. The generator now consumes the
explicit typed identities, requires packaged implementations, and coordinator
admission verifies exact descriptor/binding parity. This repairs a downstream
projection; typed fragments do not acquire owner locators or documentation
inputs. Common Job route identity was rechecked as six members, including scope.

Structural movement: the immutable-result reader now requires only read methods;
its existing locking read retains its lock behavior. Auth and Jobs expose narrow
retained-state readers. Core's manifest availability check runs before profile
state writes. No storage layout, migration, graph algorithm or lease policy
changed.

Initial narrow evidence:

- Saved-graph existing lifecycle: 3/3, `20260909T151228Z-p57729`.
- New read-only compatibility fixtures: 3/3, `20260909T152223Z-p58272`.
- Admission context/order/rejection: 3/3, `20260909T151614Z-p1845`.
- Coordinator and invocation units: 1/1, `20260909T152233Z-p67271`.
- Graph Projection envelope/storage regression: 3/3, `20260909T152237Z-p72602`.

Development failures were corrected: an initial validator omitted two Common Job
identity fields; an invalid-failure fixture initially violated the database's
broader token grammar; the new read-only result port initially required `Exec`;
and the startup diagnostic test omitted its JSON import. No retained installation
was modified. Compatibility fixtures compare declaration/result/lease/receipt/
job/proof/state/ledger bytes before and after both acceptance and rejection.
Final generation, finalization and broader verification followed this checkpoint.

Checkpoint successor: generation checks, finalization, full affected-owner
regressions and the final scope audit.


### SG-05 broader verification and scope corrections

`make agent-finalize` passed 1/1 in `20260909T152523Z-p42635` before broader
verification. `RESULTS_DIR` was unset: retained-run selection, retained-run
maintenance and performance-evidence maintenance were skipped. No exact-source
successful full warm `check` evidence was claimed.

Generation passed in `20260909T152444Z-p17719`; drift passed 4/4 in
`20260909T152454Z-p29802`; generated-artifact policy passed 3/3 in
`20260909T152503Z-p41667`; JSON shape passed 3/3 in
`20260909T152504Z-p42077`. Generation after the final authored routing/facade
inputs passed in `20260909T152932Z-p83665`.

Broader verification found and corrected stale major-5 asset expectations and
an exact Extensions routing inventory that needed the new admission rows. The
new operation owner also exposed an eager presentation-chunk load. Persistent
operations now have the workbook facade `NetworkFlowOperations.ts`, while
`NetworkFlowFeature.tsx` is only the lazy presentation entrypoint. The authored
frontend ownership/import rules explicitly admit those two narrow boundaries.
The existing browser assertion for lazy loading remains intact and passes.
Imports retained their controllers and surface behavior; only facade imports
moved. This is structural movement, separate from lifecycle behavior changes.

Backend boundary verification moved collaboration SQL assertions into the
existing Collaboration test helper, kept configuration projection independent
of the Extensions facade, and placed test job-definition composition in shared
application test support. No boundary suppression was added. Static analysis
removed the now-unused receipt helper. `make lint` then passed all 11 units in
`20260909T153459Z-p80389`.

| Verification command/scope | Result/artifact |
| --- | --- |
| `make test-slice OWNER=module.networkflow` | 35/35, `20260909T152625Z-p46280`; includes the owner's browser and bounded-rendering routing. |
| Full Network Flow Go service slice | 12/12, `20260909T153111Z-p97662`. |
| `make test-slice OWNER=web.networkflow` | 47/47, `20260909T153029Z-p78693`, after the facade correction. |
| `make test-slice OWNER=web.workbook` | 169/169, `20260909T153049Z-p86948`. |
| `make test-slice OWNER=package.protocol_ts` | 7/7, `20260909T152710Z-p32875`. |
| `make test-slice OWNER=package.ui` | 10/10, `20260909T152720Z-p36292`. |
| `make test-slice OWNER=platform.openapi` | 4/4, `20260909T152625Z-p46314`. |
| `make test-slice OWNER=module.jobapi` | 4/4, `20260909T152631Z-p56621`. |
| Common Job API Go service slice | 3/3, `20260909T152719Z-p35935`. |
| `make test-slice OWNER=platform.jobs` | 6/6, `20260909T152831Z-p506`. |
| Remaining Jobs service group retry | 3/3, `20260909T153521Z-p93115`; other service groups passed in `20260909T152918Z-p57117`. |
| `make test-slice OWNER=app.server` | 24/24, `20260909T152828Z-p98569`. |
| Configuration owner projection row | 1/1, `20260909T152822Z-p96215`. |
| `make frontend-typecheck` | 2/2, `20260909T153007Z-p19533`. |
| `make frontend-import-boundary-check` | 2/2, `20260909T153025Z-p78313`. |
| `make go-gosec-targeted` | 4/4, `20260909T152933Z-p86679`. |
| Additional ordinary visual pass | 12/12, `20260909T153007Z-p19628`, against the same promoted manifest. |

The two required fresh ordinary visual passes recorded in SG-04 remain valid;
the additional final pass includes the lazy-facade correction and no new golden
changes. Renderer, anchors, local preferences, tolerances and all 210 golden
entries remain governed by the current maintenance workflow.

Recorded infrastructure/orchestration failures: a Go service build overlapped
public regeneration (`20260909T152923Z-p62065`), one Jobs service group timed out
in setup (`20260909T152918Z-p57117`), and simultaneous service-image warm targets
raced on their shared stamp (`20260909T153521Z-p93125`). These are retried after
generation and with service startup serialized; no harness or tolerance change
was made. Failed development runs are retained as evidence, not counted as
successful verification.


### Final admission and browser evidence

The complete Extensions owner slice passed 24/24 in
`20260909T153739Z-p63489`, including the exact routing inventory and lazy-loading
browser workflow. The final cutover/lifecycle slice passed 3/3 in
`20260909T153655Z-p46194`. Current real-application lifecycle, accessibility and
500-vertex/1,000-edge measurement rows passed together, 16/16, in
`20260909T153925Z-p4165`.

Final drift passed 4/4 in `20260909T153856Z-p96073`; artifact policy passed 3/3
in `20260909T153905Z-p99334`; JSON shape passed 3/3 in
`20260909T153906Z-p99750`; the public test-catalog check exited 0.
Finalization passed again in `20260909T153911Z-p956`, with `RESULTS_DIR` still
unset and retained-run maintenance explicitly skipped. Backend module boundaries
passed 3/3 in `20260909T154148Z-p41069`; targeted security passed 4/4 in
`20260909T154151Z-p41504`. Markdown lint passed in
`20260909T154200Z-p67509`, and `git diff --check` passed.

A final retained-state audit closed another G14 case: the database enforces
summary presence but allows malformed summary JSON. The Common Jobs admission
reader now validates closed summary members, lifecycle timestamps, cancellation
codes and expired-summary erasure before Network Flow interprets owner facts.
This does not alter public reads, scheduling, finalization or retention. Focused
retained-summary tests cover malformed failure values, unsupported members,
wrong cancellation codes and invalid tombstones. Generation after this test-row
addition passed in `20260909T154442Z-p73440`.


### Compatibility, limitations and rollback audit

The implemented change is one coordinated Network Flow major-6 server/browser/
contract cutover. Public declaration and embedding resources are v4; changed
create/rename requests advance independently, while unchanged request IDs stay
unchanged. Core's job-reference schema is additive. Durable state remains v4;
Graph Projection v2, semantic-query v2, deterministic identities, Recovery
contracts and private database columns retain their existing meaning. There is
no stored-data migration, route addition, dependency or compatibility decoder.

A valid declaration alone does not establish upgrade eligibility: retained old
receipts can block startup forever under their existing replay obligations.
The preflight is subject to the existing configured validation deadline and
reports timeout separately; expiry of observation or job retention is never a
compatibility shortcut. Operators must quiesce the prior release before cutover.
A rejection changes no declaration, immutable result, lease, receipt, job proof,
state metadata or migration ledger. Such an installation remains on its prior
release until separate remediation is authorized. After major-6 writes, default
to a forward fix; use only a compatible rollback release or separately
authorized restoration of the pre-cutover backup.

The final source audit confirms an unchanged HTTP method/path set, unchanged
renderer and 210-entry golden set, exactly the two reviewed PNG changes, and no
edits to the advisory digest, database migrations, dependency manifests or
lockfiles. The only deleted authored files are the replaced v5 frontend-entrypoint
input and schema; v6 replaces them without an alias. New operation state is
memory-only. No reset, commit, push or deployment was performed. Baseline `main`
and HEAD `88319c90cff44d55b2fc003215a8294c28f0d660` remain unchanged.

Checks deliberately not run: repository-wide `test`, `check`, `ci`, and
`release-check`, which include unrelated owners beyond this change. Current
routing plans selected affected-owner, service-backed, frontend, browser,
accessibility, measurement, visual, generator, lint and security checks instead.
No production release, full-suite conformance or retained warm-run performance
claim is made. Retained-run maintenance was skipped with `RESULTS_DIR` unset.

### SG-05 completion checks

After the final retained-summary correction, the Jobs owner slice passed 6/6 in
`20260909T154452Z-p76294`; saved-graph cutover and lifecycle service checks passed
3/3 in `20260909T154536Z-p94448`. No production changes followed these runs.

| Command | Final result/artifact |
| --- | --- |
| `make generate-drift` | PASS, 4/4, `20260909T154652Z-p12442`. |
| `make generated-artifact-policy-check` | PASS, 3/3, `20260909T154701Z-p15701`. |
| `make json-shape-check` | PASS, 3/3, `20260909T154702Z-p16117`. |
| `make test-catalog-check` | PASS, exit 0. |
| `make agent-finalize` | PASS, 1/1, `20260909T154707Z-p16875`; `RESULTS_DIR` unset. |
| `make lint` | PASS, 11/11, `20260909T154722Z-p20101`. |
| `make go-gosec-targeted` | PASS, 4/4, `20260909T154743Z-p34715`. |
| `make lint-markdown` | PASS, `20260909T154753Z-p60667`. |
| `git diff --check` | PASS, exit 0. |
| Final source scope audit | PASS, 162 changed/new paths; unchanged route set, renderer, golden inventory, branch and HEAD; only two reviewed PNG changes. |

SG-05 is DONE. All 14 binary acceptance criteria passed. The completed acceptance
record passed Markdown lint in `20260909T155422Z-p63318`, `git diff --check` and
the 162-path scope audit before this tracker transition. The same three checks
are repeated against these final handoff bytes. No implementation or release
action remains within this task; release execution requires separate authorization.

Next action: none.
