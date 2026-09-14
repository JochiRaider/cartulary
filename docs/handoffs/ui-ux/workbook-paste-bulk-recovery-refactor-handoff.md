# Workbook paste and bulk recovery handoff

## Control and scope

Baseline: clean `main`, HEAD `c85de391f3d795fafb6fe1e9dea534983dc99d78`,
revalidated before editing. No unrelated changes were present. Implementation
is limited to Timeline table paste, Hosts and Identities entity-origin paste,
Timeline `fill_down_v1`, and Timeline `multi_row_tag_assignment_v1`.
Scalar paste remains with its editor/creation owner. No new endpoints, command
kinds, dependencies, persistent browser storage, analyst-data operations,
commits, pushes, deployments, or digest edits are authorized by this handoff.

The user selected explicit uncertain Retry, retained waiting overlap, and
original semantic targets through presentation/navigation changes. Source and
verification guides are subordinate to adopted behavior. The localized digest
and NLSpec research are advisory material, not additional product requests.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| PBR-01 | DONE | Owner/action/outcome reconciliation complete; regression reproductions recorded below. |
| PBR-02 | DONE | Retained owner, queue boundary, and security lifetime tests pass. |
| PBR-03 | DONE | All five consumers migrated; complete receipts, grouped conflicts, exact retry, and read debt pass focused gates. |
| PBR-04 | DONE | Five-action service/browser, security, accessibility and interaction evidence passed; terminal owner checks follow. |
| PBR-05 | DONE | Affected gates and final bytes validated; baseline failures isolated; terminal evidence and rollback complete. |

## Authority and action matrix

Behavior: Core 01 §2, §3.3.5, REQ-01-672 and active Timeline v2 / Hosts /
Identities view contracts; Core 02 §§6,8,14–15; Core 03 §§3.3.6,4,11.1,13,18–19;
Core 04 §§1–2 and applicable acceptance contracts. `domain.md` owns vocabulary;
`design.md` owns bounded design direction. Source placement follows frontend
source/import manifests and local guides; verification follows the owner
catalog and authored test families. Executable consumers must not read Markdown.

| Action | Gesture / source mapping | Target and result semantics |
| --- | --- | --- |
| Timeline table paste | Grid Adapter semantic anchor; existing browser dimension decoder; shared server tabular ingest, exact schema-derived headers | Ordered committed-record/create targets; mention-origin rules; nonconflicting portion in one change set; source-ordered cell conflicts. |
| Hosts paste | Same clipboard boundary, source-owned Entity planner | Entity-origin create/upsert intent; server exact reuse/stubs; repeated source rows may address the same record. |
| Identities paste | Same clipboard boundary, source-owned Entity planner | Entity-origin create/upsert intent; server exact reuse/stubs; no implicit mention creation. |
| Timeline fill | Pointer and keyboard semantic fill intent | Captured scalar value, stable committed target IDs/versions; no grouped, draft, collection, read-only, or presentation targets. |
| Timeline tags | Current-page opt-in record selection and submitted label | Stable targets; `timeline.tags` collection review; retained same-frame duplicate guard. |

All five actions retain source-owned admission and current authorization. Common
ownership covers operation identity, ordering, transport uncertainty, receipt
retention, and read reconciliation. A future adopted batch command supplies a
source planner and receipt interpretation to that boundary; no new recovery
implementation or generic workflow endpoint follows.

| Outcome | Disposition |
| --- | --- |
| Known initial rejection | Preserve correction context; no accepted mutation implied. |
| Uncertain write / malformed success | Retain exact attempt; explicit Retry only, with current authority. |
| Full acceptance | Retain complete receipt before projection/refresh effects. |
| Partial conflicts | Preserve committed portion and every actionable conflict. |
| Conflicts only | Empty rows, omitted change set; resolve individual cells. |
| Failed read reconciliation | Keep acknowledgement; retry reads only. |

A denied replay cannot establish that an earlier uncertain dispatch committed
nothing. Batch attempts never inherit autosave's dispatch-time rebasing or its
special FIFO transaction-conflict rekey policy.

## Gap register

Every row's validation is binary. The register preserves characterization-time
risks; all G01–G08 are closed by the linked PBR-03/04 and terminal evidence.

| Gap | Observation and remediation / affected areas | Rationale and long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G01 | One-shot paste/fill/tag transports allocate IDs and discard requests; introduce retained capture/send ownership. | Exact replay prevents duplicate creates and preserves meaning. | Existing routes and request shapes retained; legacy receipts preserved. Response loss remains open. | Lost-response retry sends identical body and ID; one durable batch. |
| G02 | Controllers own completion and selection; move operation state above presentation. | Navigation and late results cannot erase recovery or steal focus. | Scalar editors and existing creation owners remain. Obsolete callback risk open. | Detach/reattach retains operation; newer input/selection unchanged. |
| G03 | Fill/tag accepted values contain counts; retain validated complete rows/conflicts/change-set presence. | Partial acceptance remains actionable and inspectable. | Source-specific reuse rules retained; malformed-success risk open. | Missing/invalid receipt becomes uncertain; valid partial/only conflicts remain recoverable. |
| G04 | Exact Timeline headers are removed only by server planning; reconcile browser dimensions against the same schema projection. | Predictable row count and fields without another ingest engine. | Exact, case-sensitive header rules only; scalar dispatch unchanged. | Header plus N source rows emits N targets and owner-defined mappings. |
| G05 | Entity decoder accepts absent required targets; reject omission. | Public admission matches adopted contract. | Invalid legacy request omission is intentionally rejected; historical accepted receipts not rewritten. | Missing targets rejects before any write. |
| G06 | Entity history is appended per source row; reproduce repeated exact reuse and reconcile per-record revisions. | One attributable batch with coherent row history. | Exact reuse remains server-owned; repeated-reuse revision behavior unverified. | Repeated entity source rows produce one affected-record revision and no duplicate entity. |
| G07 | Batch refresh callers omit authoritative acceptance; use retained generation-fenced read debt. | Refresh failure cannot erase committed work or trigger resubmission. | Existing monotonic projection and Collaboration owners retained. Lost-debt risk open. | Accepted then failed/superseded refresh retains debt; recovery sends no mutation. |
| G08 | Queued plans and reservations depend on presentation; capture semantics and coordinate existing save/merge owners. | Predictable overlap with unrelated-row availability. | Autosave FIFO/capacity preserved; batches are separate replay units. Ordering risk open. | Duplicate delivery sends once; deliberate repetition stays separate; overlap waits without retargeting. |

## Evidence and decisions

Planning inspected the root instructions, localized digest read order, applicable
Core clauses, domain/design boundaries, frontend/backend guides, clipboard
adapters/controllers/planners, fill/tag ports, runtime registry, conflict store,
refresh registry, shared ingest, source-owned batch implementations, and routing.

Planning-only evidence (not recovery completion): all five requested task guides
passed. Narrow paste adapter/model, Timeline plan, and Entity plan slices passed
at `.cartulary/test-results/20260914T001352Z-p10997`,
`.cartulary/test-results/20260914T001352Z-p11020`, and
`.cartulary/test-results/20260914T001352Z-p10994`. The implementation checkout
recheck remained clean at the baseline HEAD.

Existing protections: Timeline paste serializes with save work; bulk tagging
guards same-frame submission; server Timeline batches validate all targets before
mutating and persist full success/conflict receipts. A pending source label
`paste` denotes row-mutation provenance, not complete batch replay.

### PBR-01 exit

The action and outcome tables above freeze the five existing routes/intents.
No adopted-owner contradiction was found. G01/G02/G03/G07/G08 follow directly
from the inspected adapter/controller paths: a thrown fetch loses the attempt;
fill and tagging discard rows/conflicts; Entity selection follows awaited reads;
tag completion returns immediately after detachment; queued fill reads no fresh
authority; paste re-resolves the visible rectangle. Existing adapter and bulk
controller tests exercise these deterministic paths but encode the previous
dispositions. Their migration must replace those expectations, not preserve the
defects. The new owner tests will inject delayed/failed transport and reads at
these same boundaries for all outcome rows.

G04 reproduction: pass the exact projected Timeline header plus one data row
through `decodeWorkbookClipboardInput`; its two-row dimensions allocate two
targets, whereas `BuildClipboardPlan` removes the header and rejects target count.
G05 new executable regression failed at
`.cartulary/test-results/20260914T002050Z-p14170` (`missing_targets`: expected
admission failure). The decoder now rejects omission, preserving historical
request hashing. Its owner slice passed at
`.cartulary/test-results/20260914T002201Z-p32567`.

G06 new service reproduction posts two names with the same exact hostname in one
paste. The existing path returns HTTP 500 instead of a receipt, at
`.cartulary/test-results/20260914T002251Z-p33282`; this is a related product
failure, assigned to PBR-03. The initial run at
`.cartulary/test-results/20260914T002114Z-p14808` failed to compile the new
test's SQL helper call; the call was corrected before the product reproduction.
The passing pre-existing tests do not certify these new scenarios.

Paths changed in PBR-01: this handoff, Entity clipboard admission and its unit
test, and Workbook clipboard service test. Exit is characterization and binary
criteria, not a claim that the unrepaired recovery regressions pass. Next action:
implement retained ownership and tests under PBR-02.

### PBR-02 exit

Added `runtime/workbookBatchOperation.ts` and `WorkbookBatchOperationOwner.ts`
with immutable plans/attempts, delivery identity, explicit uncertain retry,
retained receipts, independent reconciliation, overlap waiting and late-result
lifetime fences. Runtime assembly and shell authority now own this lifecycle.
The autosave model exposes neutral dispatch admission and an explicit boundary
that prevents coalescing or duplicate suppression across a batch. Only receipts
from captured prerequisite autosaves may advance a waiting batch's versions;
background observations and replay do not rewrite an attempt.

Focused owner tests passed at
`.cartulary/test-results/20260914T003042Z-p53148`; frontend typecheck passed at
`.cartulary/test-results/20260914T003043Z-p53434`; pending-queue FIFO/boundary
slice passed at `.cartulary/test-results/20260914T003142Z-p54398`. Authored
frontend ownership and test-family routing include the new source/tests.
No persistent storage or new public command was added. The owner retains one
latest completed receipt per surface and every unresolved receipt; same-account
suspension hides presentation without retiring requests. PBR-03 must connect
production transport/consumers and close the remaining source result defects.

### PBR-03 exit

All five consumers now admit source-owned plans to the Workbook batch owner.
The paste, fill and tag ports no longer allocate transport identities or reduce
receipts to counts. The Timeline and Entity controllers retain scalar ownership
and submit batches synchronously with captured semantic targets; Timeline
prerequisite saves close the first-dispatch version boundary. Only accepted
preceding local actions advance undispatched versions. Dispatched bytes never
change. Native clipboard delivery is deduplicated in Grid Adapter; distinct
clipboard events remain distinct actions even with identical text.

Added `adapters/createWorkbookBatchTransport.ts` and
`components/WorkbookBatchRecovery.tsx`. Recovery uses current status/focus
entry points, retained original input, explicit uncertain retry, grouped
per-cell correction, and read-only refresh retry. Batch conflict metadata is
separate from compound-operation restrictions. Later waiting batches cannot
block the earlier batch's conflict correction. Batch completion and correction
do not clear newer editor drafts or restore an obsolete paste/fill anchor.

Retired controller-local paste group state, final-rectangle reconstruction,
count-only fill/tag results, per-adapter ID allocation and local asynchronous
batch completion/focus paths. Retained shared parsing, scalar owners, conflict
resolution, monotonic Timeline projection, Collaboration ledger, autosave FIFO,
and Entity merge coordination. A future adopted batch action supplies a source
planner and receipt adapter to the same lifecycle, without another endpoint.

G04 is corrected through the existing decoder plus schema-derived exact label
matching and a neutral ordered-field mapping in Grid Adapter. Non-header tables
remain positional; ordinary comma text remains scalar. G06's reproduced HTTP
500 was specifically a deferred Entity envelope/source integrity failure from
advancing one reused entity twice. Entity upserts now advance each affected
record once, preserve ordered source mutation entries, append one final
revision/publication, and return final rows at every repeated source position.
No historical receipts are rewritten. Entity replay supplies an empty conflict
array when historical receipts omitted it. Authored OpenAPI batch/conflict
schemas and the existing release change-set registry were corrected before
regenerating derivatives.

Source no-op disposition: Timeline omits unchanged existing-record targets.
Receipt validation therefore cannot require a returned row per record target;
it validates all supplied rows/conflicts, create coverage, ordering, versions,
and legal change-set presence. An empty no-op receipt is valid. Exact Entity
reuse may repeat identities. This is a compatibility boundary, not permission
to synthesize missing create results or accept malformed response members.

Focused evidence:

- Batch capture/transport/receipt and refresh debt: `make test-slice
  OWNER=web.workbook` with the authored batch retention, refresh debt and paste
  adapter rows, PASS at `.cartulary/test-results/20260914T011243Z-p62426`.
- Runtime lifecycle, coalescing, recovery and Entity merge admission rows: PASS
  at `.cartulary/test-results/20260914T010815Z-p53320` (8/8 work units).
- Expanded shared-ingest service row: PASS at
  `.cartulary/test-results/20260914T010315Z-p26884` (3/3), including existing and
  newly created repeated Entities, final revisions, ordered history and replay.
- Header planning and fill controller rows: PASS at
  `.cartulary/test-results/20260914T010442Z-p46521` and `…-p46512`.
- Grid semantic mapping row: PASS at
  `.cartulary/test-results/20260914T011152Z-p61056`.
- Pending FIFO/boundary row: PASS at
  `.cartulary/test-results/20260914T011137Z-p60517`.
- `make frontend-typecheck`: PASS at
  `.cartulary/test-results/20260914T011107Z-p59956`.
- `make frontend-import-boundary-check`: PASS at
  `.cartulary/test-results/20260914T011242Z-p62240`; repaired the initial direct
  runtime protocol import using the existing type-only adapter facade.
- `make generate`: PASS at `.cartulary/test-results/20260914T005633Z-p20424`;
  its first attempt correctly required four compatibility dispositions, now
  registered in the authored release change set.
- `make format`: PASS at `.cartulary/test-results/20260914T010934Z-p55268`.
  The initial formatting/lint failure identified an optional-Promise truthiness
  check and non-null assertions in new code; both were corrected. Baseline
  informational style diagnostics were preserved. An attempted unsupported
  `BIOME_CHECK_FLAGS` Make override was rejected before execution and not used.

Intermediate invalid selector/sort attempts and type errors were corrected
before the passing gates above. They are not completion evidence. Remaining
risk is integrated production interaction, security, visual/accessibility and
measurement coverage; that is the next PBR-04 workstream.

### PBR-04 exit

The expanded `clipboard_paste_integration_test.go` verifies missing, deleted,
foreign-incident and wrong-type Timeline targets; authorization before malformed,
wrong-view or excessive input for all five routes; and no mutation/change-set
side effects. Entity-origin record targets are unsupported by definition: their
create intents resolve exact reuse in the Entity owner. The service evidence
also verifies repeated existing/new Entity reuse, final versions, one revision
per affected record, ordered mutation history and original receipt replay after
intervening Entity changes. Timeline mixed create/record paste, mention origin,
partial conflicts and separately attributed per-cell correction remain covered.

Production browser additions in `timeline-grid-entry.spec.ts` cover lost responses
after committed creates, exact retry after navigation/intervening changes, failed
refresh and read-only retry, unrelated newer typing/focus, exact headers,
duplicate native event delivery versus deliberate repetition, grouped paste
conflicts, conflicts-only fill and collection-tag review. `keyboard.spec.ts`
covers retained Hosts and Identities exact reuse through navigation/replay.
The original seven non-bulk Timeline spreadsheet scenarios passed freshly;
the existing bulk scenario passed with keyboard/pointer parity and unsupported
range rejection. Fill now preserves the current range endpoint after completion;
the retired unconditional source-cell focus restoration must not be reinstated.

Sorted/filtered target mapping, draft overflow, source-ordered header mapping,
unsupported targets and same-frame tag admission have fresh owner-routed unit
evidence. Retained owner tests cover role loss, closure/reopen, authorization
uncertainty, same-account recovery, replacement/obsolete callbacks, rejected
replays after uncertainty, overlap admission and immutable attempts. Existing
runtime security/conflict/merge evidence from PBR-03 remains applicable; terminal
regression gates will recheck the affected shared owners.

Browser findings and corrections:

- The first production run (`20260914T011932Z-p68328`) exposed registration churn
  that invalidated successful reads. Timeline and Entity registrations now follow
  mounted presentation lifetime and dereference current callbacks. Separate
  registration, authorization and debt generations fence read completion.
- The retained-retry fixture now reveals virtualized columns after navigation.
  Batch recovery closes the conflict panel before opening, preventing overlapping
  controls. The revised run (`20260914T012917Z-p11600`) exposed obsolete fill focus
  expectations, a non-unique conflict-value locator and an assumption that tag
  history item count equals revision count. Assertions now check retained focus,
  semantic conflict regions, row versions and attributed change sets.
- The first accessibility build (`20260914T013450Z-p34277`) was invalidated by
  changing build inputs, before browser execution. The next run
  (`20260914T013602Z-p76549`) found wrapping above the compact toolbar. The trigger
  now has a short non-color state; the full message remains in the recovery panel
  and live announcement. The original-input control uses existing input styling.
- The initial expanded admission run (`20260914T013310Z-p80555`) incorrectly reused
  a helper that required a 404 error for a valid 403 authorization rejection.
  The corrected assertion separately checks admission code and absence of protected
  row/conflict/version members. No backend authorization change was necessary.
- Consumer audit removed `timelineBulkTagSubmissionIsCurrent` and unused settlement
  plan fields; only its obsolete test remained. Stable callback registration has
  a direct regression, and prepared versions cannot regress on older receipts.

All run roots below are under `.cartulary/test-results/`. The retained run manifests
and summaries contain exact catalog selectors and per-test accounting; use
`make explain-run RESULTS_DIR=<root>` for investigation.

| Command / selected owner evidence | Result / run root |
| --- | --- |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770` | PASS, 3/3, `20260914T013929Z-p52388` |
| `make service-backed-test-slice OWNER=module.timeline` with `spreadsheet_batch_replay`, `spreadsheet_batch_conflicts`, `spreadsheet_batch_headers`, `spreadsheet_batch_bulk_conflicts` browser rows | PASS, 11/11, `20260914T013930Z-p52680` |
| Timeline original spreadsheet bulk, grouped-conflict and fill/tag correction browser rows | PASS, 11/11, `20260914T013311Z-p80799` |
| Timeline original creation, keyboard, pointer, refresh, rejection, uncertain-create and virtualization browser rows | PASS, 11/11, `20260914T013631Z-p6962` |
| Workbook routine Host/Identity paste and retained recovery browser rows | PASS, 11/11, `20260914T012917Z-p11610` |
| Workbook Host/Identity recovery and batch accessibility rows | PASS, 13/13, `20260914T013928Z-p52163` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.batch_recovery` after final styling and authorization fence | PASS, 11/11, `20260914T014532Z-p86399` |
| Four owner-selected Timeline typing, creation, ArrowDown and pointer-focus measurement rows | PASS, 20/20, `20260914T014013Z-p32768` |
| `make test-slice OWNER=module.tabularingest` | PASS, 1/1, `20260914T013403Z-p31131` |
| Timeline sorted/filtered grid and clipboard planning rows | PASS, 3/3, `20260914T013632Z-p7179` |
| Grid Adapter semantic table/fill, filtered overflow and sorted target rows | PASS, 4/4, `20260914T013633Z-p7473` |
| Entity paste planner and mutation admission/hash rows | PASS, 3/3, `20260914T013634Z-p10669` |
| Workbook surface/sentinel scalar-comma and native multiline editor rows | PASS, 4/4, `20260914T013452Z-p35504` |
| Retained batch, receipt/transport, refresh debt and merge admission rows | PASS, 5/5, `20260914T013602Z-p76509` |
| Batch and refresh-debt rows after authorization fence | PASS, 3/3, `20260914T014452Z-p84683` |
| Timeline tag plan/controller after obsolete lifecycle retirement | PASS, 3/3, `20260914T014450Z-p84442` |
| Stable Timeline runtime binding row | PASS, 2/2, `20260914T014621Z-p21628` |
| `make frontend-typecheck` | PASS, 2/2, `20260914T014453Z-p85103` |

Accessibility attachments in the retained Playwright report include desktop,
768px and 390px captures and the recovery accessibility tree. Desktop and narrow
captures were visually inspected: controls fit, keyboard focus is visible, raw
input stays copyable and state does not depend on color. No golden was promoted.
Measurements preceded the final authorization-fence and styling corrections;
those corrections do not change the measured typing/selection/creation paths.

Exit disposition: G01–G08 meet their integrated binary criteria. Remaining work
is PBR-05: source-guide updates, finalization, affected shared-owner regression,
visual and drift gates, complete acceptance assessment and terminal-byte lint.
No adopted-owner contradiction or unresolved product decision blocks that work.

### PBR-05 terminal evidence

Source guides now describe capture/send contracts, retained conflict and refresh
ownership, stable surface registrations, and once-per-record Entity history.
Authored source registration is in `tools/frontend_source_ownership.json`;
exact test selectors are in the `web.workbook`, `module.timeline` and
`module.workbook` test-family inputs. Existing files retain their registered
owners. Generated protocol and harness derivatives came only from `make generate`.

| Correction | Adopted behavior / changed ownership |
| --- | --- |
| G01, G08 | Core 01 §3.3.5 Table A and replay rules; Core 03 §4, §13.3. Secure IDs, captured bytes, admission, overlap and recovery live in Workbook runtime; Grid Adapter supplies semantic delivery. |
| G02, G07 | Core 03 REQ-03-100/299 and §§18–19; Core 04 §§1–2. Existing incident/account lifetime and monotonic source projections govern retention, concealment and reads. |
| G03 | Core 01 §3.3.5 batch envelope; Core 03 §3.3.4, REQ-03-083–085. Full conflicts and attributed per-cell resolution remain in the existing conflict owner. |
| G04 | Core 01 §2 tabular-ingest contract and active view field registries; Core 03 REQ-03-147–150. Shared decoder/schema mapping reconciles exact headers; scalar comma text remains scalar. |
| G05 | Core 01 §3.3.5 Table A required `targets[]`; existing Entity admission handles omission. |
| G06 | Core 02 §6 entity-origin table, §8.2 exact-match precedence and §15 history; Core 03 REQ-03-151. Entity batch aggregation preserves ordered mutation history and one final record revision. |
| Limits/security | Core 01 §3.3.5 target-scope paragraph and REQ-01-672; Core 03 REQ-03-147; Core 04 authorization/concealment. Shared ingest and source-owned atomic validation remain in place. |

Terminal review made five substantive corrections:

- Batch conflicts now drive both semantic cell state and the existing local
  recovery buttons from the retained conflict store. The scalar draft queue is
  retained for scalar coordination. Sentinel unit, production grouped-conflict
  and stateful public-mutation scenarios verify the corrected projection.
- Completed-payload pruning now also runs after remount reads and conflict
  settlement. Every unresolved entry and the latest completed receipt per surface
  remain retained. The dedicated retention test verifies release only after both
  obligations settle.
- Historical Entity responses may omit the empty conflict member. Explicit
  `conflicts: null` is malformed and remains uncertain; the adapter regression
  distinguishes the two cases.
- Retirement now fences queue-drain admission before batch teardown emits state.
  The existing injected-scheduler disposal regression caught and verifies this
  correction. Refresh-registry consumers now observe debt settlement; the old
  no-notification assertion was replaced with the actual debt-state contract.

- The autosave prerequisite identity set is fixed at batch admission. Waiting
  for an explicit earlier save no longer reclassifies later edits as prerequisites.
  The retained-owner regression verifies preceding version advancement, later
  overlap blocking, unrelated-row availability and release after acceptance.

Test migration also corrected a tag fixture that returned stale versions from
its final refresh, and a scalar-conflict test that expected implicit focus
transfer. The latter now verifies editor focus remains until the existing
explicit recovery action is used. These changes preserve the adopted interaction
contract, not earlier incidental async behavior. Temporary test compilation
errors (an unsupported Testing Library `exact` option, a removed local variable,
and a DOM `Node.remove` typing mismatch) were corrected and typechecked.

Public Make evidence below supplements PBR-03/04. Counts are harness work units,
not independent behavioral claims. Each run retains `run-summary.json`,
`run-manifest.json`, `target-summaries/` and detailed row artifacts. Markdown lint
uses its tool summary under `adhoc/lint-markdown/`.

| Command / selection | Result / run root under `.cartulary/test-results/` |
| --- | --- |
| `make generate` after final selector registration | PASS, `20260914T020341Z-p14755` |
| `make agent-finalize` | PASS, final `20260914T022709Z-p11853`; `20260914T015033Z-p27378` preceded broad verification, and `20260914T020457Z-p56643` rechecked final selector registration. |
| `make test-slice OWNER=module.entities` | PASS, 50/50, `20260914T015135Z-p58369` |
| `make test-slice OWNER=module.revisions` | PASS, 31/31, `20260914T015135Z-p58478` |
| Entity ordinary upsert/exact-match service rows | PASS, 4/4, `20260914T015116Z-p31403` |
| Timeline clipboard parser/bulk source unit rows | PASS, 2/2, `20260914T015135Z-p58383` |
| Three Timeline production visual rows: default, edit/save/conflict, grouped | PASS, 11/11, `20260914T015116Z-p31429` |
| `make frontend-import-boundary-check` | PASS, 2/2, `20260914T021248Z-p40209` |
| `make generated-artifact-policy-check` | PASS, 3/3, `20260914T015341Z-p14727` |
| `make generate-drift` | PASS, 4/4, `20260914T021248Z-p39956` |
| Sorted/scalar/header/tag sentinel row after fixture correction | PASS in `20260914T020427Z-p50072`; grouped-conflict row subsequently PASS, 2/2, `20260914T021148Z-p74566` |
| Existing Collaboration scalar-conflict row after explicit-focus correction | PASS, 2/2, `20260914T020428Z-p50503` |
| Timeline production grouped batch conflicts with local marker assertion | PASS, 11/11, `20260914T021200Z-p75258` |
| Workbook grouped-paste, stateful Timeline public mutations, canonical Indicator creation/replay browser rows | PASS, 15/15, `20260914T021200Z-p75242` |
| `make test-slice OWNER=web.workbook` with batch retention, complete receipt, read debt and runtime-responsibility rows after terminal corrections | PASS, 5/5, final `20260914T022709Z-p11858` |
| Batch production accessibility, including 200% zoom, text spacing, reduced motion, desktop/narrow layouts and keyboard Retry/Close | PASS, 11/11, `20260914T021723Z-p20553` |
| `make frontend-typecheck` after terminal source/test corrections | PASS, 2/2, final `20260914T022709Z-p11994` |
| `make lint-biome` | PASS, 2/2, final `20260914T022709Z-p12075` |
| `make format` | PASS, 2/2, `20260914T022647Z-p7327` |
| Timeline production replay after fixed prerequisite admission | PASS, 11/11, `20260914T022538Z-p70167` |

`make agent-finalize` initially failed at `20260914T014749Z-p22594` because
renamed authored selectors had not yet been generated (`json-shape-check` at
`20260914T014933Z-p23385`). Regeneration and subsequent finalization passed.
`RESULTS_DIR` was unset throughout: retained-run maintenance was skipped because
there is no qualifying successful full warm check. No retained historical run
was substituted for this seam's fresh evidence.

### Broad failures and isolation

`make frontend-unit` at `20260914T015116Z-p31529` failed (603/610). The repeat at
`20260914T021248Z-p40186` failed (606/610): three pre-existing static-policy rows
and the runtime-responsibility row described above. All affected functional
frontend rows now pass, including the final targeted responsibility rerun.
The broad target remains reported FAIL; it is not represented as a green run.

The static failures reproduce at the exact baseline in the detached read-only
comparison checkout `/tmp/cartulary-pbr-baseline-c85de391`:

- `make test-slice OWNER=web.architecture` selecting layout and selector policy
  rows: FAIL at that checkout's `.cartulary/test-results/20260914T020456Z-p55727`.
  Findings are baseline geometry literals in `CoordinationCreateRecovery.tsx`
  and `EntityWorkbookSurface.tsx`, plus existing raw test selectors in ordinary
  creation, timing support and Sentinel tests.
- `make test-slice OWNER=harness.browser` selecting architecture policy: FAIL at
  that checkout's `.cartulary/test-results/20260914T020456Z-p55730` for five
  baseline literal view-schema bindings. The offending lines are unchanged by
  this seam. Source/import boundary verification independently passes.

`make test-slice OWNER=module.workbook` broadened to service/browser evidence and
failed at `20260914T015135Z-p58386` (75/87). Related grouped-conflict, sentinel
and stateful Timeline failures are repaired and pass in the table above. One
canonical Indicator snapshot assertion failed in that broad run; its baseline
comparison and focused current rerun both pass. No change was made to that
independent operation. Its intermittent broad-run failure remains recorded.

Five other browser failures reproduce at baseline under
`make service-backed-test-slice OWNER=module.workbook`, selected rows, in
`/tmp/cartulary-pbr-baseline-c85de391/.cartulary/test-results/20260914T020405Z-p17932`
(13/21 work units passed): Coordination option availability, the old keyboard
Tab expectation, System views creation-focus expectation, queue-overflow editor
fixture focus, and saved-view layout payload expectation. Detailed Playwright
reports and traces identify the same assertions as the broad current run.
These are isolated outside the changed behavior; this handoff does not claim
repository-wide readiness. The comparison checkout has no authored changes and
is retained solely with reproducible test evidence; fixture services did not
operate on analyst data.

### Acceptance assessment

The following PASS dispositions are scoped to the affected seam and its
regression boundaries. Unchanged product-wide behavior is not newly certified.
The three isolated static failures and five baseline browser failures are
reported above; no applicable seam criterion is waived or marked BLOCKED.

| Row | Disposition and evidence |
| --- | --- |
| A001 | PASS — adopted clauses mapped above; source placement and verification routing remain independent. |
| A002 | PASS — one shared batch lifecycle hides capture/uncertainty/reconciliation; five source planners retain meaning; retired/retained paths and future-command fit recorded. |
| A003 | PASS — clean baseline, current manifests/guides, generated roots, direct Grid Adapter boundary and final diff reviewed; no localization artifact was edited. |
| A004 | PASS — shared control, recovery-surface and input styles; no new token, theme or density registry. |
| A005 | PASS — existing dark_graphite fixtures and visually inspected batch recovery; no theme surface added. |
| A006 | PASS — unchanged shared density ownership; fresh compact production visual, Grid Adapter and Timeline measurement evidence. |
| A007 | PASS — mixed existing/create paste, Entity stub/reuse, original Timeline first-input/continued typing and full Entity owner evidence; ordinary/contextual creation regressions pass affected frontend rows. |
| A008 | PASS — shared RecoverySurface geometry, desktop/768px/390px and zoom/text-spacing recovery; responsive algorithms unchanged. |
| A009 | PASS — recovery scrolls within existing surface, Close/Escape and shell navigation remain reachable in narrow browser evidence. |
| A010 | PASS — existing inspector/resolver routing retained; conflict navigator and review are explicit; detached and authority-invalid callbacks cannot become current. No new inspector feature registry. |
| A011 | PASS — original targets, source text, selection and newer typing survive delay/navigation/read failure in production and retained-owner tests. |
| A012 | PASS — existing secure ID owner, immutable bytes, native-event duplicate guard, distinct repeated gestures and lost-response replay verified. |
| A013 | PASS — batch never rekeys automatically; acknowledged recovery sends reads only; autosave's separate queue policy remains covered by fresh frontend/runtime tests. |
| A014 | PASS — scalar comma/native multiline editor paths, direct typing, Enter/Tab/Escape, fill parity, local correction and memory-local lifetime verified. |
| A015 | PASS — retained conflicts mark affected cells, preserve local/saved values, and remain keyboard-reachable after navigator dismissal; no blanket overwrite. |
| A016 | PASS — data/authority projections remain separate; suspended/closed/read-only admission and refresh-debt tests pass without inferring authority from receipt data. |
| A017 | PASS — role loss, closure/reopen, session uncertainty/recovery, incident revocation, replacement and obsolete callback evidence; no new persistent browser storage. |
| A018 | N/A — no Evidence lifecycle, overlay or preview behavior changed; existing Evidence/related-creation frontend regression boundary remains covered. |
| A019 | PASS — named controls, polite status, visible non-color focus/state, explicit keyboard Retry, Escape/Close, reduced motion, zoom/text spacing and narrow layout evidence. |
| A020 | PASS — existing shared components, short toolbar state, copyable retained input, responsive scrolling and production text-spacing/zoom cases; no new component variant registry. |
| A021 | PASS — semantic IDs/fields and virtualized target mapping; original production virtualization scenario and owner-selected typing/selection/creation measurements pass. |
| A022 | PASS — current production visual fixture registry and three selected visual rows; batch screenshots manually inspected at desktop/narrow sizes; no golden promotion or claim-publication assertion. |
| A023 | PASS — authored tests use existing semantic selector builders, record/field/view identities and accessible roles; no new incidental test selectors. Baseline policy violations isolated above. |
| A024 | PASS — runtime, tests and generation depend on authored machine projections; Markdown changes are human guidance only. Terminal Markdown lint is documentation maintenance. |
| A025 | PASS — authored OpenAPI/release/ownership inputs precede generator output; generation drift and artifact policy pass; no hand-edited generated roots or lockfiles. |
| A026 | PASS — existing endpoints/commands and ingest owner retained; legacy receipts readable; omission rejection and repeated-Entity revision correction are intentional; rollback below. |
| A027 | PASS — sequential exits, gap dispositions, commands/artifacts, failures/limitations, compatibility and retirement decisions recorded; completed bytes are validated before terminal promotion. |

Digest classifications: apply R001–R015 with R012 adapted to existing runtime
lifetimes; R016–R025 remain owner-scoped adaptations (desktop density, legitimate
grid scrolling, local recovery and current confirmations). R035 informs narrow,
zoom and text-spacing recovery. Reject R026–R034's incompatible visual, workflow,
authority and incidental-selector prescriptions. No upstream dependency, generated
design system, touch profile or marketing layout is introduced.

### Compatibility, limitations and rollback

There is no database migration or new endpoint/command/dependency. Existing
clipboard/bulk idempotency keys and normalized historical request comparison stay
intact. Old Entity receipts lacking `conflicts` remain readable; explicit null
is invalid. Required-target omission now rejects intentionally. Repeated Entity
reuse now commits one revision per affected record with ordered mutation history.
Unchanged Timeline record targets may be absent from valid no-op receipts; complete
validation cannot infer a mutation from their absence. Every supplied member and
create target is still validated. This limitation follows the existing source
receipt contract, not a count-only approximation.

Recovery is memory-local, explicitly initiated after uncertainty, and cannot
survive a full browser process loss beyond existing session retention. Unresolved
operations are not evicted; a long offline session may retain substantial input.
Later overlap deliberately waits, while unrelated work follows existing owner
ordering. Receipt payloads can be released only after read/conflict obligations
settle. The latest completed receipt per surface remains available.

Skipped: full warm `make check`/release/build-all, exhaustive unrelated subsystem
browser suites, dependency/security vulnerability audits, golden promotion and
persistent-storage recovery. The seam changes no dependencies, deployment,
cryptography, SQL schema or persistent storage. Fresh selected backend/service,
frontend, production-browser, accessibility, visual and interaction measurements
cover the affected owners. The broader failures above remain reproducible evidence
and prevent a repository-wide clean-check claim. The four interaction measurements
precede final receipt validation, conflict-marker and retirement corrections;
those corrections do not alter the measured typing/creation/selection hot paths.

Rollback restores this coherent set together: authored OpenAPI and release
projections; generated protocol/OpenAPI and harness derivatives; Workbook runtime,
conflict/refresh owners and all five consumers; Entity batch/history correction;
Grid Adapter translation; authored source/test routing; tests and source guides.
Retain read compatibility for already committed receipts. Do not selectively
restore one-shot consumers while leaving the captured transport contract installed.
Preserve unrelated work and all committed analyst records. Never delete records
or reverse accepted batches to compensate for a UI rollback.

Terminal artifact procedure: prepare a temporary sibling containing the exact
completed handoff bytes, run public Markdown lint over it and this handoff, then
promote those already-validated bytes and remove the sibling. Final byte identity,
Git scope and diff checks are recorded as manual handoff evidence in
`.cartulary/pbr-final-validation/terminal-validation.json`. This documentation-only
record is not an executable product or conformance dependency. Next action after
terminal promotion: review the uncommitted change set; implementation stops at
this seam. No commit, push, deployment, digest edit or analyst-data action occurred.

PBR-05 exit: G01–G08 and every applicable acceptance row pass with the scoped
evidence above. Final `make lint-markdown` and `git diff --check` pass; their
terminal run reference, byte identity and final branch/scope review are retained
in `.cartulary/pbr-final-validation/terminal-validation.json`. The final tracker
was promoted only after these completed bytes passed Markdown validation. No
required seam work remains; the uncommitted result is ready for review.
