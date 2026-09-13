# Network Analysis table lifecycle refactor

## Baseline and authority

Implementation baseline: clean `main` at
`0809e80628ada85a8497d0c21ab9ceff98cc4cf4`, revalidated 2026-09-09.
Root `AGENTS.md` is the only applicable agent instruction. The localized digest
read order was inspected during planning and its entry points revalidated.
Historical handoffs are completed regression evidence, not new authorization.

Owners: Network Flow Activity NLSpec §§5–8, 13, 17, 19 and 28; Core 00
precedence, Core 04 REQ-04-023/105A, Core 03 extension collaboration and
availability; Extensions EXT-REQ-201 cleanup; design §§13.1/14. Domain owns
vocabulary/navigation; the NLSpec research article is supporting guidance.
No adopted-owner contradiction was identified. Tests and generation consume
authored machine inputs only, never Markdown.

The approved current-only correction assumes no pre-refactor table mutation
receipts require compatibility. Correct the adopted digest and table response
discriminators together, with no legacy decoder, alternate hash, receipt
migration, receipt deletion, or new persistence. Frontend/backend delivery must
be coordinated; rollback cannot promise replay across the digest correction.
No commit, reset, push, or deployment is authorized.

## Tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| TL-01 Baseline, audit and characterization | DONE | Owner map, gap ledger and focused reproductions |
| TL-02 Admission, immutable attempts and recovery | DONE | Bounded contracts/backend repairs and operation tests |
| TL-03 Catalog, selection and scoped invalidation | DONE | Reconciliation and continuity race coverage |
| TL-04 Dialogs, accessibility and browser regressions | DONE | Local recovery, focus and real browser evidence |
| TL-05 Final validation and handoff | DONE | Required checks passed; final scope and compatibility recorded |

Only the current row may be IN_PROGRESS. Record paths, commands, results,
compatibility, risks and next action before completing each row. Never advance
past a blocked dependency.

## Gap ledger and implementation boundary

| Gap | Owner | Bounded repair |
| --- | --- | --- |
| Rename follows import capability; dispatch lacks current authority | NF §5; Core 04 REQ-04-105A | Exact operation roles and synchronous dispatch admission |
| Dialog retargets current props; conflicts reselect mutation target | NF §§7/8/17; design §14 | Captured target/version/name, explicit review, independent dialog identity |
| Transaction IDs allocated in transport; delete receipt discarded | NF §5 Table 5-B and §17 Table 17-B5 | Immutable attempt, exact recovery, validated receipt and independent refresh outcome |
| Raw request hashing; forbidden whitespace controls misclassified | NF-REQ-025/036a/066 | Adopted transcript, normalization before comparison, controls before trim |
| Table response discriminators differ from adopted spellings | NF §17 Tables 17-B3/B4/B5 | Repair authored projections, producers, decoders and fixtures; Make generation |
| Table admission accepts closed incidents | NF-REQ-022/069a; Core 04 REQ-04-105A | Open admission and transaction-bound current membership/role |
| Lists can regress versions or resurrect deletions | NF §§7/8/19 | Scope-local versions, terminal tombstones and local request fences |
| Any deletion broadly resets resources; graph reacts to metadata identity | NF-REQ-047b/066/125/170/170d/208 | Semantic source identity and affected-consumer invalidation |
| Feedback outside modal; UTF-16 length restriction | NF-REQ-066; design §14 | Owner normalization/scalar validation, local recovery and focus |

One workbook-lifetime table owner coordinates authorized catalog,
selection, captured presentation, one immutable admitted attempt and settlement.
Import, saved graphs and indicator linking retain their own controllers. Their
existing boundaries receive scoped table transitions. The digest and completed
DU tracker remain unchanged.

## Execution evidence

- Both `make task-guide ROLE=module-author OWNER=web.networkflow` and
  `make task-guide ROLE=module-author OWNER=module.networkflow` passed at
  planning and implementation entry.
- Planning table/selection baseline: 7/7 units passed at
  `.cartulary/test-results/20260909T222921Z-p5855`; this characterizes existing
  behavior and does not bless the identified mismatches.
- TL-01 started from the revalidated clean baseline; owner-backed failures were
  recorded before runtime changes, as detailed below.

## Final validation

Final evidence is recorded in TL-05 below. `RESULTS_DIR` remained unset; no
qualifying successful full warm check was used for retained-run maintenance.
Real browser, accessibility and visual verification were executed.

### TL-01 exit

Added backend digest/control-priority characterizations in
`internal/modules/networkflow/table_lifecycle_test.go`, catalog/selection races
in `apps/web/src/networkFlow/networkFlowController.test.ts`, and authored routing
under `tools/test_families`. `make generate` passed at
`.cartulary/test-results/20260909T225320Z-p12404`.

The new frontend row fails as expected (1/2 units) at
`.cartulary/test-results/20260909T225339Z-p15452`: catalog regression/resurrection
and prior-order fallback. Backend row fails as expected (0/1) at
`.cartulary/test-results/20260909T225339Z-p15455`: raw hash transcript and
forbidden-control precedence. These are intended owner mismatches, not harness
failures. Existing baseline evidence remains recorded above.

TL-02 starts with captured operation/transport and bounded backend repairs.
Catalog characterizations remain pending TL-03; no dependency is blocked.

### TL-02 exit

Added `NetworkFlowTableController` and `networkFlowTableOperation` with captured
intent, synchronous operation roles, immutable serialized attempts, bounded
observation, exact replay, receipt validation, and separate acknowledgement.
The client now preserves delete receipts and validates table response identity.
Backend `application.go`, `routes.go`, `names.go`, `api.go` and narrow store error
details implement normalized digests, current transaction admission, adopted
response discriminators, and normalization-first validation. Authored contracts,
protocol adapters and fixtures changed together; `make generate` passed at
`.cartulary/test-results/20260909T230936Z-p27539`.

- Captured-operation frontend row passed 2/2 units at
  `.cartulary/test-results/20260909T230955Z-p30566` (10 deterministic tests).
- Backend lifecycle unit row passed 1/1 at
  `.cartulary/test-results/20260909T230703Z-p25499`.
- Service-backed table admission/replay row passed 3/3 units at
  `.cartulary/test-results/20260909T231129Z-p49224`, including canonical replay,
  historical replay after deletion, exact roles, scalar boundaries, closure,
  and membership change after route admission while waiting on a transaction.
- The prior combined service run at
  `.cartulary/test-results/20260909T230955Z-p30564` passed the retained
  resource-change/replay/privacy row, but failed the new test's incorrect 422
  expectation. The adopted status is 400; corrected the test and completed the
  required display-name error details before the passing rerun.

Compatibility remains the approved coordinated current-only correction.
Presentation still uses the old hook until integration; final type and browser
checks follow the dependent workstreams. TL-03 now replaces catalog transitions
and connects semantic consumer boundaries. No blocked dependency.

### TL-03 exit

The table reducer now retains accepted per-table versions and terminal removal
knowledge inside the authorized workbook lifetime. Complete lists use prior
next/previous selection fallback. Request generations fence lists crossing
mutations and events; historical receipts acknowledge without replacing current
metadata. Missing resources purge their dependent draft/attempt. Session pause
hides knowledge; actor/incident replacement and complete read/claim loss purge it.

The table hook is a subscription adapter; the workbook owner hook binds current
authority and collaboration. Consumer subscriptions replace broad table-event
resets. Graph work uses semantic scope and resolved IDs, preserving metadata-only
continuity; affected deletion and all-active membership changes expose stale
state until recomputation. Indicator linking ignores table rename and unrelated
deletions, and has a narrow mutation-admission status boundary. Existing saved
graph resource-change handling remains its owner boundary.

New catalog, graph-consumer and indicator-continuity rows plus operation and
reducer rows passed 6/6 units at
`.cartulary/test-results/20260909T231932Z-p79147`. Earlier focused failures were
a test fixture accidentally using the original actor, and an all-active effect
starting work before the stale render; corrected the fixture and added a
synchronous stale fence. Generation passed at
`.cartulary/test-results/20260909T231909Z-p76183`.

No persistence or global catalog revision was introduced. TL-04 connects these
boundaries to workbook composition and extracts presentation, then verifies
complete workflow regressions and browser behavior. No blocked dependency.

### TL-04 exit

Workbook composition now owns the table controller. The workspace subscribes and
selects through explicit methods; extracted `NetworkFlowTableLifecycle.tsx`
provides captured rename/delete dialogs and volatile recovery. Current role gates,
normalized scalar feedback, local errors, explicit review, exact replay and modal
focus share the existing controls. No UTF-16 length cap remains. Focus restoration
respects a newer dialog and falls back to the selected table. Source removal
cancels resource-scoped pending work/drafts; an already acknowledged delete receipt
remains historical operation evidence. Session recovery pauses/hides volatile
state, while closure retains copyable rejected drafts. HTTP read loss reaches
existing workbook, saved-graph and indicator-link boundaries.

Frontend Network Flow passed 56/56 units at
`.cartulary/test-results/20260909T232742Z-p8604`. Type and import-boundary checks
passed; latest type/boundary/Biome evidence is
`.cartulary/test-results/20260909T233859Z-p90089`, `-p90093`, and `-p90099`.
Protocol projection tests passed 7/7 at
`.cartulary/test-results/20260909T233603Z-p76657`.

The real browser batch at `.cartulary/test-results/20260909T232842Z-p17914`
passed import recovery, indicator-link recovery, active deletion, saved-graph
deferred navigation and Network Analysis accessibility. Its failures identified
uncertain feedback obscured by peer metadata, the retained broad-reset assumption,
and the intentional delete-dialog comparison. Corrected feedback precedence.
The lifecycle regression now asserts preserved graph mode and stale state after
source deletion, then explicitly returns to Rows and verifies recovery; protected
payload/selector clearing assertions remain intact. The peer-change/exact replay
browser row passed at `.cartulary/test-results/20260909T233208Z-p66327`.
The lifecycle row passed 11/11 units at
`.cartulary/test-results/20260909T233445Z-p40935`, after correcting its assertion
to use the existing human-readable `graph stale` label.

Reviewed ordinary-run visual reconciliation (9 active captures, 210 accounted
PNGs, no missing golden or ambiguous mapping) and both old/current images.
`make browser-e2e-visual-update` passed 12/12 units at
`.cartulary/test-results/20260909T233227Z-p96562`. Only the Network Analysis delete
dialog PNG and its Make-owned manifest changed. Reviewed the promoted image:
scoped deletion wording, local feedback spacing and visible input focus; no
unexpected clipping, font, background or layout change. No assertion, fixture,
renderer or tolerance was weakened. Two ordinary post-promotion passes remain
required in TL-05 before completion.

`make agent-finalize` passed, latest at
`.cartulary/test-results/20260909T233858Z-p89856`. RESULTS_DIR was unset: retained
run selection, retained-run checks and performance-evidence maintenance were
explicitly skipped. No successful full warm check for this source is claimed.
An attempted guide lookup for nonexistent `protocol.networkflow` returned a usage
error; the actual `package.protocol_ts` and `web.workbook` task guides passed.

TL-05 now completes ordinary visual passes, final owner slices, projection/drift
checks, scope audit and the final handoff. No blocked dependency.

### TL-05 validation and scope audit

Final inspection distinguished an undispatched request from an unknown server
outcome in both settlement and feedback. The former can be reviewed and submitted
as a fresh attempt; the latter retains its original transaction and exact bytes.
Write rejection blocks further dispatch under the cached authority while allowing
a current authorized catalog read. A read rejection purges protected state through
existing workbook and consumer boundaries. Deterministic tests cover both paths.

The final service test holds an incident row lock and observes the PostgreSQL
blocking graph before releasing two normalized-equivalent store renames. Exactly
one allocation succeeds and the other reports `duplicate_display_name`. Initial
waiter instrumentation assumed direct blockers and visible query text; those
assumptions failed in this service harness. The corrected recursive lock graph
retains the simultaneous-waiter and exclusive-allocation assertions. Service runs
at `.cartulary/test-results/20260909T234017Z-p99907`,
`20260909T234255Z-p66857` and `20260909T234529Z-p26484` recorded these test
synchronization failures. The first also exposed a retained expectation that
whitespace containing a tab was empty: it now expects the adopted forbidden-control
error, with a separate actual whitespace-only case. The corrected table service
row passed 3/3 units at `.cartulary/test-results/20260909T234725Z-p81139`.

The latest `make agent-finalize` passed 1/1 at
`.cartulary/test-results/20260909T235323Z-p12088`, before final broader verification.
`RESULTS_DIR` was unset. Retained-run selection, retained-run checks and performance
evidence maintenance were explicitly skipped; no full warm or release claim is
made. No unrelated capacity, measurement, security or whole-repository release
suite was required for these lifecycle, projection and presentation changes.

Completed evidence (counts are harness units, not individual test cases):

- `make test-slice OWNER=package.protocol_ts`: 7/7 at
  `.cartulary/test-results/20260909T233603Z-p76657`.
- `make test-slice OWNER=web.workbook` with the three routed workbook surface and
  availability regression rows: 4/4 at
  `.cartulary/test-results/20260909T234134Z-p57069`.
- `make test-slice OWNER=module.networkflow` with table lifecycle, display-name
  and saved-receipt unit rows: 1/1 at
  `.cartulary/test-results/20260909T234337Z-p84375`.
- `make backend-module-boundary-check`: 3/3 at
  `.cartulary/test-results/20260909T234403Z-p14775`;
  `make protocol-ts-browser-artifact-reachability` also exited successfully.
- `make generated-artifact-policy-check` and `make json-shape-check`: 3/3 each
  at `.cartulary/test-results/20260909T234134Z-p57023` and `-p57027`.
- `make service-backed-test-slice OWNER=module.networkflow` with
  `module.networkflow.browser.table_lifecycle_continuity` and the claimed Network
  Analysis accessibility row: 13/13 at
  `.cartulary/test-results/20260909T234337Z-p84402`. The browser uses the real
  backend, loses observation after commit, navigates, receives a later peer
  rename, and replays the exact original request while retaining current metadata.
  Keyboard review, modal-local feedback, Unicode input, focus restoration and
  metadata-only graph/contributor continuity are asserted. Other completed import,
  saved-graph and indicator-link browser evidence is recorded in TL-04.
- Two fresh `make browser-e2e-visual` ordinary runs after promotion passed 12/12
  each at `.cartulary/test-results/20260909T234016Z-p98587` and
  `.cartulary/test-results/20260909T234528Z-p25138`. Every promoted image was
  reviewed; only the previously documented delete-dialog golden changed.

Final source verification also passed:

- `make test-slice OWNER=web.networkflow`: 56/56 at
  `.cartulary/test-results/20260909T235352Z-p15512`, including operation, catalog,
  graph continuity and completed import/saved-graph/indicator-link regression rows.
- `make service-backed-test-slice OWNER=module.networkflow` with the following six
  rows: table lifecycle admission/replay, resource-change replay/privacy, source
  mutation audit rollback, saved-graph lifecycle v2, bounded graph contributors,
  and active/soft-deleted store semantics. All six rows passed in 5/5 harness units
  at `.cartulary/test-results/20260909T235352Z-p15527`.
- `make frontend-typecheck`, `make frontend-import-boundary-check` and
  `make lint-biome`: 2/2 each at
  `.cartulary/test-results/20260909T235352Z-p15638`, `-p15650` and `-p15668`.
- `make generate-drift`: 4/4 at
  `.cartulary/test-results/20260909T235352Z-p15463`.
- `make format`: 2/2 at `.cartulary/test-results/20260909T235307Z-p7770`.
  `git diff --check` passed and final branch/HEAD remained unchanged.

Exact final service selection:

```sh
make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.resource_intent_failure_rolls_back_source_mutation_8a4c3d2e1f,module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232
```

`make lint-markdown` passed after the completed evidence update at
`.cartulary/test-results/20260909T235600Z-p43472` (summary:
`adhoc/lint-markdown/tool-run-summary.json`). `git diff --check` passed. Scope
inspection found 53 changed paths, exactly this one documentation file and one
reviewed PNG, no lockfile changes, and no paths outside the authorized roots.
TL-05 is DONE. The final Markdown lint, diff check and branch/HEAD/status inspection
are repeated after this final tracker update; no implementation work remains.

### Final architecture and compatibility

Catalog observations, selection, captured intent, admitted requests and receipts
are distinct state within `NetworkFlowTableController`. The old reducer is a pure
catalog transition used by that owner; the old hook is a subscription/presentation
adapter. There is one unresolved mutation slot and separate presentation identity.
Local read generations, accepted per-table versions and terminal tombstones obey
the API's complete-list ordering without inventing a global revision. Valid
selection survives rename and non-active deletion; active deletion uses prior
next/previous order and then empty. Import keeps its validated handoff boundary.

Rename preserves analytical identity and metadata-only continuity. Source removal
clears affected rows, diagnostics, cursors, selectors and contributors; affected
exploration stays in graph mode and requires recomputation. All-active membership
is tracked independently of display names. Saved declarations and indicator-link
state remain in their existing owners, receiving source-scoped transitions.
Current authorization governs visibility and dispatch, while historical receipts
remain operation evidence. Recovery is volatile workbook memory only.

Changed authored paths are in the Network Flow table owner, operation rules,
workspace/dialog/client/reducer and consumer adapters under
`apps/web/src/networkFlow`; workbook composition under `apps/web/src/workbook`;
Network Flow route/application/name/store error code and retained/new tests under
`internal/modules/networkflow`; the three authored `contracts/network-flow`
projections; protocol adapters/fixtures under `packages/protocol-ts`; source
ownership, test-family routing and JSON-shape checks under `tools`; and this
handoff. Browser changes are limited to `apps/web/e2e/network-flow.spec.ts` and the
single delete-dialog PNG. Make-owned generated protocol, Network Flow/import
contract artifacts, import target registry, browser batch/topology indexes and
visual manifest changed as projections of those authored inputs.

No migration, persistence, restore, hard deletion, deleted-table browsing or graph
algorithm/query redesign was introduced. Coordinated frontend/backend delivery
is required for corrected discriminators. Rollback cannot promise exact replay
between the old and corrected digest formats; no compatibility decoder or receipt
migration/deletion was added. No unresolved adopted-owner decision remains.
The digest package and completed DU tracker are unchanged. Branch and HEAD remain
`main` at `0809e80628ada85a8497d0c21ab9ceff98cc4cf4`; the baseline was clean and all
current modifications belong to this seam. No commit, reset, push or deployment
was performed.
