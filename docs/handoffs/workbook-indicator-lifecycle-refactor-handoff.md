# Workbook Indicator lifecycle refactor

## Baseline, authority, and scope

Execution baseline: clean `main` at
`71dee41c8557550d831fc1092b21f2ad43a7216c`; tracked and untracked status empty.
Root AGENTS.md and the digest localized read order were read during planning
and revalidated at execution. No nested instructions apply. The digest,
research, and completed handoffs are advisory evidence, not new authorization.

Owners: Core 01 REQ-01-615–617/652/654, Core 02 REQ-02-079/080/260/263,
Core 03 REQ-03-035/283/301/306 and transaction/continuity owners, Core 04
REQ-04-150 and current session/membership/CSRF/incident authorization, and
design §§5.3, 7.3, 10, 12, 14. No owner contradiction identified.

Adopted owner paths are
`docs/spec/01_architecture_storage_and_view_contracts.md`,
`docs/spec/02_domain_model_schema_and_history.md`,
`docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, and
`docs/spec/04_security_deployment_and_conformance.md`.
Vocabulary/navigation and design direction remain in `docs/domain.md` and
`docs/design.md`. Read `docs/research/nlspec-spec.md` and
`docs/cartulary-ui-ux-refactor-digest/README.md` as supporting material in their
stated boundaries. The completed Timeline capture/actions, Task lifecycle,
Decision supersession, entity merge recovery, record-history browsing/recovery,
save-status, and Network Analysis indicator-link handoffs under `docs/handoffs/`
were implementation/regression evidence only.

Permitted: lifecycle frontend, narrow Inspector/runtime/transport/query and
typed-constraint integration, focused tests/browser support, authored selectors,
source ownership and verification mappings, Make-generated derivatives, reviewed
affected visuals, and this handoff. Public routes/schemas, backend production,
migrations, dependencies, digest contents, and unrelated workflows are excluded.
No commit, reset, push, deployment, or analyst-data cleanup is authorized.
UTC date/time fields are the user-selected authoring convention.

Planning evidence: focused baseline PASS 4/4 at
`.cartulary/test-results/20260911T192557Z-p93437`; production child-route service
baseline PASS 3/3 at `.cartulary/test-results/20260911T192557Z-p93439`.
Both `make task-guide ROLE=module-author OWNER=web.workbook` and
`make task-guide ROLE=module-author OWNER=module.indicators` passed again at
execution. Existing tests mostly cover observation and server behavior.

## Gap ledger

| Gap | Owner | Remediation / affected areas | Rationale / long-term benefit | Compatibility | Unresolved risk | Binary completion |
| --- | --- | --- | --- | --- | --- | --- |
| G1 History omits lifecycle content | Core 01 615–617; Core 03 306 | Generic Inspector canonical History composition | Make declared actions usable through one semantic entry | Same tuples and panel | Duplicate or wildcard dispatch | Exact read/manage render once; mismatches omitted |
| G2 Shared disposable lifecycle/observation state | Core 02 079/080; Core 03 continuity | Separate lifecycle presentation and retained owner | Isolate drafts and async operations by subject | Observation behavior retained | Cross-subject or account leakage | Close/retarget/reopen retains only the matching draft |
| G3 Missing details and support authoring | Core 02 079/263 | Detail list and paged reference reader/picker | Understand assessments and choose records without UUID knowledge | Existing query and support_refs | Incomplete candidates or hidden data | All details visible; outside-page references selectable |
| G4 Noncanonical times and incomplete local validation | Core 01 652; Core 02 263 | UTC parser and typed constraints | Prevent invalid admission without destroying typed work | Same nullable wire fields | Calendar normalization or precision loss | Invalid dates reject; canonical dates, equal bounds and 0/100 pass |
| G5 New IDs per submission and lost uncertainty | Core 03 301; Core 01 652 | Synchronous reservation, immutable attempt, runtime recovery | One logical append with auditable recovery | No FIFO, locks, persistence or automatic retry | Late receipt or security change | Duplicate sends once; replay route/body/ID/order exact |
| G6 Receipt dropped and refresh owns completion | Core 03 035/306 | Complete receipt checks, version accounting, reconciliation | Preserve accepted writes independently of reads | Existing affected-record response | Stale overwrite or false rejection | Receipt accepted first; refresh-only retry sends no append |
| G7 Prepending and conflated pagination states | Core 01 654; design 10/14 | Ordered generation-fenced collection model | Reliable browsing and appropriate recovery | Same opaque cursor contract | Page races and duplicate identities | Backdated/equal-time order, page retry and obsolete-read tests pass |

Advisory classification: ADOPT local feedback, keyboard parity, explicit states
and recovery; ADAPT dense presentation through existing tokens; REJECT generic
workflow, re-theme, new API and destructive-confirmation ceremony.

## Ordered tracker

| Workstream | Dependency | Status | Binary exit |
| --- | --- | --- | --- |
| ILC-01 | None | DONE | Baseline/ledger recorded; failing characterization; observation baseline green |
| ILC-02 | ILC-01 | DONE | History, details, authoring, references, paging and isolated drafts pass |
| ILC-03 | ILC-02 | DONE | Admission, immutable recovery, receipt and authority/lifetime tests pass |
| ILC-04 | ILC-03 | DONE | Reconciliation, real service/browser, accessibility and adjacent regressions pass |
| ILC-05 | ILC-04 | DONE | Finalizer, full check, final Markdown/diff/scope audits and handoff pass |

Only the current row is IN_PROGRESS. Each log closes DONE or BLOCKED before
advancement; blocked dependencies prohibit later workstreams.

## ILC-01 execution log

Inspected the requested frontend composition/handler/command paths, mutation
runtime and retained Timeline/Decision/history patterns, query/reference broker,
OpenAPI/view inputs, Indicator admission/replay/persistence/projection/rollback,
focused tests, browser inventory, and verification routing. The backend already
implements append-only intervals, canonical timestamps, ordered collections,
exact replay and tombstone-aware rollback. The reference broker drops paging;
use a lifecycle-specific paged reader over the existing workbook query route.
Remaining risks are enumerated in the ledger. Characterization is next.

ILC-01 DONE: routed History composition characterization fails as expected at
`.cartulary/test-results/20260911T193953Z-p19087` (1/2 units), because lifecycle
content is absent. Existing observation slices PASS 3/3 at
`.cartulary/test-results/20260911T193954Z-p19312`. `make generate` PASS at
`.cartulary/test-results/20260911T193935Z-p15912`. No production edits in ILC-01.

## ILC-02 execution log

Started with separate lifecycle contracts, draft/paging state, History UI and
supporting-record reads. Mutation dispatch remained inactive until ILC-03.

ILC-02 DONE: lifecycle code is separate from observation capture/resolution;
History now renders canonical lifecycle content. Added UTC/calendar and nullable
validation, constraints generated from unchanged OpenAPI, per-subject draft
storage, generation-fenced paged collections and supporting-record selection.
The shared reference broker and public contracts remain unchanged.

Focused characterization/model/paging/presentation PASS 4/4 at
`.cartulary/test-results/20260911T195145Z-p43382`.
`make frontend-typecheck` PASS 2/2 at
`.cartulary/test-results/20260911T195134Z-p38881`.
`make format` PASS 2/2 at `.cartulary/test-results/20260911T195133Z-p38681`.
Earlier format/lint failures were new hook dependency, label association, and
non-null assertion diagnostics; earlier typecheck failed two untyped test mocks.
All were corrected without weakening checks. Generation PASS at
`.cartulary/test-results/20260911T195016Z-p33540`; `git diff --check` PASS.
Runtime-backed admission is intentionally the dependent ILC-03 work.

## ILC-03 execution log

Started with retained owner, immutable transport, current authorization,
synchronous reservation, complete receipts, and exact replay.

ILC-03 DONE: synchronous admission, immutable attempts, bounded observation with
late settlement, security-aware exact replay, complete receipt validation,
retained recovery and draft/version review are integrated into the workbook
runtime. Append stays outside PATCH, FIFO, and destructive locks. Account and
incident retirement clear protected runtime state; shell detachment retains it.

Focused owner/transport/runtime PASS 4/4 at
`.cartulary/test-results/20260911T201425Z-p77765`; typecheck PASS 2/2 at
`.cartulary/test-results/20260911T201425Z-p77840`. Format PASS 2/2 at
`.cartulary/test-results/20260911T201414Z-p73363`; generation PASS at
`.cartulary/test-results/20260911T201311Z-p64364`. Earlier failures were new test
fixture reference metadata and a save-status assertion using the projector input
instead of its public snapshot. Corrected without production-policy changes.
Receipt-first projection reconciliation is the dependent ILC-04 work.

## ILC-04 execution log

Started with stable-identity reconciliation, collection/focus fences, browser
fault injection, accessibility, and adjacent regressions.

ILC-04 intermediate evidence: focused lifecycle PASS 9/9 at
`.cartulary/test-results/20260911T203426Z-p69501`; import boundary PASS 2/2 at
`.cartulary/test-results/20260911T203425Z-p69065`; source ownership PASS 2/2 at
`.cartulary/test-results/20260911T203515Z-p71711`; adjacent Decision, Timeline,
merge and save-status units PASS 15/15 at
`.cartulary/test-results/20260911T203549Z-p97692`.

Production child-route service PASS 3/3 at
`.cartulary/test-results/20260911T203236Z-p67269`; existing Revisions Indicator
child rollback service PASS 3/3 at
`.cartulary/test-results/20260911T203236Z-p67280`, including versioned tombstones
and cleared lifecycle projection. These tests leave backend production unchanged.

Real browser evidence: recovery/rollback passed in
`.cartulary/test-results/20260911T202830Z-p83938` (the combined run failed its
new visual fixture's missing import). Ordered paging/support selection and
accessibility passed in `.cartulary/test-results/20260911T202442Z-p39727`
(the combined run exposed an initial authorization/query-row ordering race).
Late malformed delivery, subject draft/focus isolation and closed-incident exact
replay PASS 11/11 at `.cartulary/test-results/20260911T203537Z-p72612`.

Browser findings repaired in scope: normalize RFC3339 offset reads to UTC while
keeping canonical UTC authoring; register an already accepted selected Records
row after same-account authority becomes available; open existing History before
rollback; expect the existing closure-driven Inspector dismissal. The original
receipt and before/after collection/history equality are attached to the recovery
scenario. No backend policy or route adjustment was needed.

Visual review: ordinary captures at
`.cartulary/test-results/20260911T203028Z-p27001` exposed browser-default white
inputs. Existing workbook input tokens corrected them; inspected desktop and
narrow ordinary trace captures at
`.cartulary/test-results/20260911T203347Z-p38127` preserve focus and drawer
scrolling. Both runs failed only because the two new goldens were absent after
the fixture import correction. Intentional trigger: new lifecycle authoring in
History. Stable fixture `visual.fixture.indicator_lifecycle_authoring`; row
`module.workbook.visual.indicator_lifecycle_authoring`; new filenames
`indicator-lifecycle-authoring-linux.png` and
`indicator-lifecycle-authoring-narrow-linux.png`. Existing masks, renderer,
viewport conventions and screenshot scope are unchanged; explicit profiles use
1280×720 and 768×640, default density, dark_graphite, 100% zoom. Promotion and two
fresh ordinary passes remain pending.

Additional ILC-04 evidence: observation frontend/vocabulary PASS 4/4 at
`.cartulary/test-results/20260911T204124Z-p19753`; Network Analysis indicator-link
browser/accessibility PASS 13/13 at
`.cartulary/test-results/20260911T203837Z-p45693`; Task/Decision exact-recovery
browser PASS 11/11 at `.cartulary/test-results/20260911T203837Z-p45696`.

The first full visual update at `.cartulary/test-results/20260911T203559Z-p3900`
failed only two fixture-inventory assertions because the new fixture ID had not
been added to the authored expected inventory. No golden or manifest was
promoted. Updated that inventory, corrected native date-picker icon contrast with
the existing dark-control convention, and inspected fresh ordinary desktop and
narrow trace captures at `.cartulary/test-results/20260911T204123Z-p18340`.
Its accessibility row passed; visual comparison failed only for absent new
goldens. Later tuple-omission test edits briefly failed parsing/type checks;
corrected the fixture syntax and used a valid mismatched typed route owner.
Typecheck PASS 2/2 at `.cartulary/test-results/20260911T204302Z-p61701` and
canonical tuple/authoring PASS 3/3 at
`.cartulary/test-results/20260911T204303Z-p61862`.

Final lifecycle unit selection PASS 9/9 at
`.cartulary/test-results/20260911T204529Z-p97592`; all three lifecycle browser
scenarios plus accessibility PASS 13/13 at
`.cartulary/test-results/20260911T204528Z-p97353`. Biome PASS 2/2 at
`.cartulary/test-results/20260911T204710Z-p34363`.

The second full visual update at
`.cartulary/test-results/20260911T204325Z-p63034` passed every browser group
but failed final reconciliation because the new authored scenario IDs lacked
the required hexadecimal form. No golden was promoted. Corrected the five
new scenario IDs in the authored catalog and regenerated through `make generate`
at `.cartulary/test-results/20260911T205028Z-p36852`.
Ordinary reconciliation v3 at
`.cartulary/test-results/20260911T205049Z-p44455/browser-e2e-visual/frontend-visual-reconciliation.json`
now validates the capture identities and profiles with zero ambiguous mappings;
comparison fails for the two intentional new, absent goldens. Existing goldens
are outside this narrow selection and are not deletion candidates.
The two failed update candidate directories contained only these two additions,
with no changes to existing PNG bytes.

Final paging review added semantic equality for repeated identities when JSON
object keys arrive in another order; array order remains significant. Focused
authoring/paging PASS 2/2 at `.cartulary/test-results/20260911T205051Z-p44727`;
format PASS 2/2 at `.cartulary/test-results/20260911T205029Z-p37509`.

Full visual promotion PASS 12/12 at
`.cartulary/test-results/20260911T205406Z-p77887`. Reconciliation v3 accounts
for 219 active captures and 219 goldens, all 27 registered fixtures, zero
orphans, zero missing goldens, and zero ambiguous mappings. Inspected both
promoted images: visible native date controls, explicit UTC labels, default
density, yellow keyboard focus, correct wrapping and drawer scrolling at desktop
and narrow widths. Only the two intended lifecycle PNGs and the Make-generated
golden manifest changed; all earlier golden bytes remain unchanged.

First post-promotion ordinary full visual run failed 10/12 at
`.cartulary/test-results/20260911T205856Z-p15470`. The lifecycle visual row
passed. The existing entity-mention chip-state row
`module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`
failed before capture: dismiss returned HTTP 200 with `resolution_status=dismissed`
and source/mention version 3, but the expected dismissed chip disappeared from
the mounted Inspector. No screenshot mismatch or golden mutation occurred.
Investigating through its routed row without changing unrelated behavior.
The same entity-mention row passed in isolation, unchanged, 11/11 at
`.cartulary/test-results/20260911T210347Z-p53611`. No waiver, assertion change,
timeout increase, or unrelated repair was applied. Two successful full ordinary
visual passes are still required.
The unchanged full ordinary rerun PASS 12/12 at
`.cartulary/test-results/20260911T210504Z-p85287`, including the entity-mention
row and all lifecycle captures. This is the first required successful ordinary
pass against the promoted manifest.

Additional owner guide `package.protocol_ts` passed; its focused
`make test-slice OWNER=package.protocol_ts` PASS 8/8 at
`.cartulary/test-results/20260911T210218Z-p50895`.

Second unchanged full ordinary visual PASS 12/12 at
`.cartulary/test-results/20260911T210903Z-p21681`. Both required successful
ordinary runs reconcile all 219 active goldens and 27 registered fixtures with
zero missing or ambiguous mappings. The earlier failed ordinary attempt remains
reported; no historical waiver was inherited.

ILC-04 DONE: receipt-first reconciliation, monotonic HTTP/socket accounting,
late-completion and identity fences, real-service lost-response exact replay,
read-only refresh recovery, History rollback, ordered/support paging,
accessibility, adjacent regressions, reviewed promotion and two fresh full
ordinary visual passes meet the binary exit. No outstanding ILC-04 blocker.

## ILC-05 execution log

Started after ILC-04 DONE. Run `make agent-finalize` with `RESULTS_DIR` unset:
there is no qualifying exact-source successful full warm check evidence.
Then run `make check`, complete the delivered-path/compatibility/rollback audit,
and revalidate final handoff bytes with Markdown lint and `git diff --check`.
Any failing required gate blocks completion; no scope expansion is presumed.

`env -u RESULTS_DIR make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260911T211308Z-p57251`; the detailed
`unit-artifacts/finalize-summary.json` reports zero updated files. JSON shape,
catalog/tier coverage and generated structure/drift passed. Retained-run canonical
evidence, scheduler event/timing checks and performance evidence maintenance were
SKIPPED because `RESULTS_DIR` was not provided. No historical waiver or retained
warm-run claim was used. Full `make check` started afterward.

The first full check at `.cartulary/test-results/20260911T211328Z-p60990`
identified a new accessibility-test selector-policy violation:
`web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039`.
The recovery focus assertion used a heading-name selector. Remediation is to
scope through the shared recovery selector and independently assert the heading's
accessible name and focus; the rule will not be weakened. Await the complete run
before repairing and repeating required verification.
The run finished FAIL 846/847 with only this violation. Corrected the focus
assertion and added the new lifecycle selector family to the shared-builder
policy, preserving enforcement. No runtime source or golden changed.
Focused selector policy PASS 2/2 at
`.cartulary/test-results/20260911T212138Z-p20228`; corrected real-service
accessibility PASS 11/11 at `.cartulary/test-results/20260911T212138Z-p20234`.
Format PASS 2/2 at `.cartulary/test-results/20260911T212126Z-p15648`;
`git diff --check` PASS. Repeat finalization before the next full check.
Repeated `env -u RESULTS_DIR make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260911T212254Z-p52116`, again with zero generated
updates and the same explicitly skipped retained-run maintenance. The second
full `make check` follows this finalizer.

Second full `make check` PASS 847/847 at
`.cartulary/test-results/20260911T212327Z-p56014` (428,982 ms), with zero failed,
skipped or cancelled units. This includes the corrected selector policy, focused
type/import/source boundaries, generated/contract checks, full frontend units,
service integration and harness checks. No failed check was weakened and no
historical waiver was inherited. Full browser/visual/accessibility evidence is
recorded separately above; the full check is not presented as browser evidence.

ILC-05 DONE: finalizers, full check, generated/source boundaries, reviewed visual
evidence, completed path inventory, compatibility/rollback documentation and
scope audit pass. Completed-handoff Markdown lint PASS at
`.cartulary/test-results/20260911T213142Z-p2149`; `git diff --check` and the
complete tracked/untracked audit PASS. The mandated post-DONE Markdown/diff/status
recheck runs against this final handoff text; its retained root accompanies the
completion report. All workstream dependencies are closed. Next action: none.

## Delivered changes and compatibility

The History panel now composes the exact lifecycle read/manage actions once.
Lifecycle presentation, UTC draft validation, page state and retained mutation
ownership are separate from observations. Constraints come from unchanged typed
OpenAPI inputs through the existing Protocol-TS generator. Supporting records
use paginated, authorized same-incident workbook queries and reference labels.

The retained owner reserves each admission synchronously, captures an immutable
request with its original secure transaction identity, rechecks fresh admission,
and keeps uncertainty independent of disposable Inspector components. Exact
replay uses the original route and bytes with current transport credentials.
Accepted receipts advance version accounting before reads; every affected
identity is reconciled, and refresh recovery sends reads only. Collection reads
rebuild the loaded extent in server order and fence obsolete pages and subjects.

Public routes, JSON schemas, wire tokens, append-only storage, server-owned
attribution and lifecycle summary are unchanged. Equal bounds, overlaps, null
optional fields, empty support and confidence zero/100 remain legal. Existing
History rollback is the only reversal mechanism. No API or stored-data migration
is required. Observation capture/resolution and adjacent workflows retain their
existing implementations, with regression evidence above.

Runtime limits: drafts and attempts survive closing/retargeting the Inspector and
same-runtime shell detachment. Account/incident replacement, runtime disposal or
page reload retire that in-memory state. No browser persistence, automatic
mutation retry or recovery endpoint was added. Supporting-record discovery may
require further pages; errors retain an explicit incomplete/stale indication.
The earlier intermittent entity-mention visual failure passed unchanged in its
focused row and both later full ordinary runs; no new limitation was accepted
to bypass it.

## Delivered path inventory

The following are the actual changed/new paths. Generated derivatives and PNGs
were produced through Make; authored tests and selectors are routed through the
authored owner/catalog inputs. No executable check reads this handoff.

```text
apps/web/e2e/indicator-lifecycle.spec.ts
apps/web/e2e/support/workbook/indicatorLifecycle.ts
apps/web/e2e/workbook.a11y.spec.ts
apps/web/e2e/workbook.visual.spec.ts
apps/web/e2e/workbook.visual.spec.ts-snapshots/indicator-lifecycle-authoring-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/indicator-lifecycle-authoring-narrow-linux.png
apps/web/src/testing/indicatorLifecycleTestSupport.ts
apps/web/src/testing/selectorContractPolicy.test.ts
apps/web/src/workbook/WorkbookShell.tsx
apps/web/src/workbook/adapters/createIndicatorLifecycleAdapter.test.ts
apps/web/src/workbook/adapters/createIndicatorLifecycleAdapter.ts
apps/web/src/workbook/adapters/createIndicatorLifecycleReader.ts
apps/web/src/workbook/adapters/indicatorLifecycleProtocol.ts
apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts
apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts
apps/web/src/workbook/features/generic/GenericWorkbookInspector.tsx
apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx
apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.test.tsx
apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.tsx
apps/web/src/workbook/features/indicators/IndicatorLifecycleContext.ts
apps/web/src/workbook/features/indicators/IndicatorLifecycleDraftStore.ts
apps/web/src/workbook/features/indicators/IndicatorLifecycleOperationStatus.tsx
apps/web/src/workbook/features/indicators/IndicatorLifecycleSupportPicker.tsx
apps/web/src/workbook/features/indicators/IndicatorLifecycleWorkflow.test.tsx
apps/web/src/workbook/features/indicators/IndicatorLifecycleWorkflow.tsx
apps/web/src/workbook/features/indicators/WorkbookIndicatorLifecycleOwner.test.ts
apps/web/src/workbook/features/indicators/WorkbookIndicatorLifecycleOwner.ts
apps/web/src/workbook/features/indicators/WorkbookIndicatorLifecycleRecovery.tsx
apps/web/src/workbook/features/indicators/indicatorLifecycle.characterization.test.tsx
apps/web/src/workbook/features/indicators/indicatorLifecycleModel.test.ts
apps/web/src/workbook/features/indicators/indicatorLifecycleModel.ts
apps/web/src/workbook/features/indicators/indicatorLifecycleOperation.ts
apps/web/src/workbook/features/indicators/indicatorLifecyclePaging.ts
apps/web/src/workbook/features/indicators/indicatorLifecycleReconciliation.test.tsx
apps/web/src/workbook/features/indicators/indicatorLifecycleRuntime.test.ts
apps/web/src/workbook/features/indicators/indicatorLifecycleStyles.ts
apps/web/src/workbook/features/indicators/reconcileIndicatorLifecycleReceipt.ts
apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts
apps/web/src/workbook/hooks/useWorkbookSurfaceQueries.ts
apps/web/src/workbook/models/genericWorkbookModel.ts
apps/web/src/workbook/mutations/createIndicatorWorkflowPort.ts
apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts
apps/web/src/workbook/query/useGenericSurfaceQuery.ts
apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts
docs/handoffs/workbook-indicator-lifecycle-refactor-handoff.md
packages/protocol-ts/src/entrypoints/http.ts
packages/protocol-ts/src/generated/core-indicator-registry.ts
packages/ui-contracts/src/index.ts
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_source_ownership.json
tools/frontend_visual_fixture_registry.json
tools/frontend_visual_golden_manifest.json
tools/protocol-ts/generate-protocol-types.mjs
tools/test_families/module.workbook.json
tools/test_families/web.workbook.json
```

## Code rollback

A separately authorized code rollback should reverse only the listed feature and
narrow integration changes, restore the corresponding authored ownership/catalog
inputs, and regenerate derivatives through Make. Restore the matching reviewed
visual sources/manifest together. Preserve any subsequent user edits. Do not run
database rollback, delete committed intervals, remove transaction receipts or
history, or discard analyst data. Resolve retained uncertain attempts through
the existing exact-replay path before intentionally retiring their runtime.
No rollback or deployment was performed during this implementation.

## Final scope and exit audit

G1–G7 binary exits are satisfied by the routed characterization, model/paging,
presentation, retained-owner, transport, runtime, reconciliation and socket-order
tests, plus the real-service/browser evidence above. The ledger's risk column
records the risks addressed by those exits; the permitted runtime lifetime and
explicit paging/recovery limits are described under compatibility.

Final branch remains `main`; HEAD remains
`71dee41c8557550d831fc1092b21f2ad43a7216c`. The delivery contains 28 modified
tracked files and 28 new untracked files, all listed in the path inventory. No
staged changes or pre-existing user edits were altered. The final scope audit
found no changes to public contracts, backend production, migrations,
dependencies, lockfiles, adopted specifications, digest contents, or unrelated
workflow implementations. Both new goldens and the generated manifest are the
only visual artifact changes. No commit, reset, push or deployment was performed.

The final Markdown/diff/status closeout follows the completed full check and
the DONE tracker update. No additional product work or external action is
authorized or required by this delivery. Next action: none.
