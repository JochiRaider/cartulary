# Network Analysis saved-graph lifecycle remediation

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
