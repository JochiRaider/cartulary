# Network Analysis import approval and recovery handoff

## Baseline and authority

Implementation baseline: clean `main` at
`cfea1353ecb766f8d2a447ddafee56bd32ab97a4`, revalidated on 2026-09-08.
No unrelated changes were present. No reset, commit, push, or deployment is authorized.

Read AGENTS.md and the digest's localized read order. Network Flow §§7–12,
16, 18 and 25 owns analytical identity, preview, mapping, selection, diagnostics
and claim boundaries. Core 01 §§3.3.9.1, 17.1–17.2 owns typed transport,
idempotency and jobs; Core 02 §7.2 owns source identity; Core 03 §11.2 owns
approval, persisted selection and apply lifecycle. Core 04 §§1–2 and
REQ-04-053 owns current authorization and inert input. Extensions §18,
EXT-REQ-201/212 owns cleanup and generation fencing. Design §§12/14 applies
within its declared scope. Domain vocabulary and NLSpec research are navigation
and supporting material, not additional product authority.

The Workbook Import Assistant and Network Analysis chrome handoffs are
implementation evidence only. Their completed polling and control work is reused.
The digest remains unchanged. The narrow offline feedback query ran with
`PYTHONDONTWRITEBYTECODE=1 python3 -B`: local recovery and visible focus are
ADOPT; asynchronous presentation and compact containment are ADAPT; unrelated
styling, animation prescriptions and advisory product policy are REJECT.

## Ordered tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| NI-01 Baseline, owner map and characterization | DONE | Verified gaps and focused reproductions |
| NI-02 Ownership and captured identities | DONE | Admission, preview, approval and transition tests pass |
| NI-03 Recovery, observation, cancellation and handoff | DONE | Replay, uncertainty, late-result and acknowledgement tests pass |
| NI-04 Presentation and browser evidence | DONE | Keyboard/recovery workflows and shared regressions pass |
| NI-05 Final validation and completed handoff | DONE | Required checks and final scope audit pass |

Only the current row is IN_PROGRESS. Complete its exit checks and append paths,
commands, results, decisions and next action before starting its successor.
An adopted-owner contradiction is BLOCKED: owner contradiction; do not advance.

## NI-01 evidence

Inspected the analytical controller/model/modal, workspace and table controller,
application/session and workbook lifetime bindings, shared import primitives,
authored Network Flow/Imports contracts, generated adapters, tests and routing.
Current `make help` and task guides for web.networkflow/module.networkflow pass.
Additional guides for module.imports/module.jobapi/web.application/web.workbook
and package.ui pass and will be used only for affected seams.

Confirmed gaps: write failures collapse to ready; discovery read failures collapse
to idle; submitted requests are not retained; durable approval and selection are
not independently recoverable; table loading swallows handoff failures; role loss
and unrelated table clearing reset drafts; the modal has no local recovery state.
The selected-profile model instead uses its first profile and the modal omits
filename/parser context and an explicit preview-slice qualifier.

Planning evidence on the exact clean baseline:

- Focused Network Flow model/role/returned-table slice: 3/4 execution units,
  two rows pass and returned-table row fails at
  `.cartulary/test-results/20260908T221306Z-p2930907`. The select fixture returns
  a ready unit without required approved mapping/fingerprint; typed rejection is
  correct. Repair the fixture while retaining returned-table assertions.
- Shared Imports lifecycle, typed transport and bounded observation: PASS 4/4,
  `.cartulary/test-results/20260908T222508Z-p2935765`.
- Selected Network Flow import facade/replay checks: PASS 1/1,
  `.cartulary/test-results/20260908T222508Z-p2935770`.
- `git diff --check`: PASS. No product files changed during planning.

Decisions: one current CSV workflow in application-session volatile memory;
dialog dismissal continues the current bounded observation window while the
workspace remains open; workspace departure pauses observation. Neither sends
server cancellation. No public API, stored-data migration, dependency or browser
storage is targeted. No owner contradiction was found.

The ready-unit selection fixture now includes its durable approval and
fingerprint. The unchanged ordinal/returned-table assertion passes alongside
the model and role rows: `make test-slice OWNER=web.networkflow ROWS=...`,
PASS 4/4 at `.cartulary/test-results/20260908T222937Z-p2938142`.
NI-01 exit: PASS. Next action: captured analytical operation identities and
deterministic preview/approval transitions.

## NI-02 evidence

Added NetworkFlowImportController and networkFlowImportState under the analytical
owner. They capture immutable write identity, source/discovery, candidate revision,
authority generation, applicable preview and acknowledged approval separately.
Added six deterministic ownership cases covering synchronous admission, inert
preview, stale success/failure, fingerprint mismatch, copy-only authority
transitions, immutable submitted drafts, unsupported profiles and row rejection.

The mapping model resolves selected-profile fields and requirements. A small
generated timestamp metadata export derives solely from the existing authored
Network Flow schemas and mapping registry; no public schema changed. Approval
column construction now accepts only its actual preview dependency.

- `make generate`: PASS, `.cartulary/test-results/20260908T223053Z-p2939127`.
- `make test-slice OWNER=web.networkflow ROWS=...analytical_import_ownership,...networkflowimportmodel_suite_12249943c5`:
  PASS 3/3, `.cartulary/test-results/20260908T223859Z-p2945558`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260908T223809Z-p2944952`.
  The earlier failed run `20260908T223737Z-p2943763` exposed an unnecessarily
  wide approval-helper parameter and unused future-stage imports; both corrected.

Structural changes are new state ownership and generated choice consumption.
Observable corrections are candidate invalidation, retained approval and local
required-field eligibility. UI integration remains deliberately in NI-04.
NI-02 exit: PASS. Next action: stage recovery, cancellation and typed handoff.

## NI-03 evidence

The analytical controller resumes selection from acknowledged approval, replays
uncertain writes with the captured request and ID, and assigns new IDs only after
definite rejection. Accepted jobs survive observation and source-read failures.
Bounded observation, workspace pause, cancellation admission/replay and terminal
races are independent of dialog presentation. Submitted drafts remain frozen.
Table handoff has a typed outcome; its table controller binding checks stable ID,
incident, source session/unit/hash and approved fingerprint before selection.
Publication success survives downstream failure, and delayed handoffs are fenced.

- `make test-slice OWNER=web.networkflow ROWS=...analytical_import_ownership,...analytical_import_recovery`:
  PASS 3/3, fourteen cases, `.cartulary/test-results/20260908T225010Z-p2949288`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260908T225011Z-p2949556`.
  Earlier failure `20260908T224638Z-p2947884` was the request-renewal union typing;
  corrected without changing shared request capture or transport.

Structural changes: focused operation state and typed analytical handoff port.
Observable corrections: recovery by acknowledged stage, honest paused/terminal
outcomes and correlated selection. Shared transport/observation are unchanged.
NI-03 exit: PASS. Next action: application lifetime, local presentation and browser
evidence, including the real returned-table binding.

## NI-04 evidence

The application session now owns the analytical controller. Workbook bindings
supply incident membership, lifecycle and both current claims. A retained surface
above workspace unmount provides copy-only draft review after closure. Dialog
Close/Escape dismiss presentation; workspace departure pauses observation and
fences previews and handoff. Neither requests cancellation. Confirmed closure is
observed through the shared collaboration session and reconciled with Core
incident identity, preserving the existing workspace fallback.

The modal presents stage, typed failure, job outcome and recovery locally. It
shows authorized filename/parser/profile context, ordinal-qualified source
headers, owner-returned safe samples, requirement tokens, policies, required
counts and explicit preview-slice limits. Field-associated diagnostics, status
announcements, token controls, focus containment and trigger restoration remain.
Table handoff returns selected/unavailable/failed/superseded and checks provenance;
confirmed import success survives list failure. Unrelated table clearing no longer
resets the import owner. New import intent requires an explicit new-workflow action.

One demonstrated shared-boundary exception: a queued Import request could dispatch
after the independent Network Flow claim was withdrawn. The focused reproduction
failed at `20260908T230611Z-p3012254`. ImportClient now accepts an optional current
consumer dispatch guard and checks it inside the extension queue before transport
and after receipt. The analytical binding supplies it; Workbook semantics and
shared polling remain unchanged. The regression passed at
`20260908T230835Z-p3013830` and in the final focused slice below.

Verification:

- Analytical ownership/recovery/dispatch: PASS 4/4 execution units,
  `.cartulary/test-results/20260908T232603Z-p3146374`. Seventeen controller cases
  plus queued dispatch coverage include fresh preview before retained replay,
  resource-specific denial, workspace preview fencing and delayed cancellation.
- Shared Imports bounded observation, lifecycle, typed transport and local
  recovery: PASS 5/5, `20260908T232616Z-p3170202`.
- Application analytical lifetime: PASS 2/2, `20260908T232849Z-p3254638`.
- Frontend typecheck: PASS 2/2, `20260908T232849Z-p3254721`; boundary check:
  PASS 2/2, `20260908T232849Z-p3254747`.
- Real Network Flow recovery and retained-draft/closure browser rows plus
  accessibility: PASS in `20260908T232604Z-p3146633`. Recovery injected discovery
  read failure, selection rejection, accepted apply with lost response, job read
  failure and table-list failure; upload/approval were not repeated and the exact
  apply body was replayed. Closure keeps a copy-only draft above workspace removal.
- That accessibility row checks 390x480, 768x360 and 1440x450 viewports, 200% zoom,
  text spacing, keyboard containment and focus return. Reviewed its attached
  `network-flow-mapping-text-spacing` PNG: content wraps inside a scrollable dialog.
- Workbook Import Assistant production workflow and keyboard recovery browser
  regressions: PASS 11/11, `20260908T232642Z-p3194665`.
- Existing ordinal mapping, alias collision and returned-table real browser rows:
  PASS 11/11, `20260908T230127Z-p2957885`. Full web.networkflow unit slice:
  PASS 42/42, `20260908T231108Z-p3068127` before the additional edge cases.
- `make format`: PASS, `20260908T232831Z-p3250197`.

Failed intermediate checks were resolved: closure initially did not refresh the
shell identity (`20260908T231012Z-p3018628`); hook dependency lint required explicit
lifetime scoping (`20260908T232426Z-p3135807`); accessibility screenshot cleanup
needed Node-compatible typing (`20260908T232616Z-p3170449`). A command-line VERBOSE
input was rejected by public typecheck preflight; subsequent ordinary typecheck
passed. No unrelated code was changed for these checks.

Ordinary visual evidence `20260908T231332Z-p3074429` exposed an unintended extra
workbook grid row from the retained host. Corrected with a layout-neutral host
and recovery affordance in the existing top bar. Repeated ordinary visual slice
`20260908T232604Z-p3146633` retains the original workbook geometry. Reviewed all
eight changed actual images: mapping dialog and the retained-import trigger/status
in accepted inspector, rejected diagnostics, graph contributors, saved graph,
narrow query, compact saved graphs and delete dialog. The filtered-empty capture
is unchanged. The overall slice is 13/15 because visual comparison and its summary
correctly reject old goldens. No golden was promoted from either run. Full visual
reconciliation, promotion and two ordinary passes are NI-05 work.

NI-04 exit: PASS. Next action: finalizer, owner-routed final checks, visual
maintenance and completed acceptance/scope audit.

## NI-05 verification and scope record

`make agent-finalize` passed at `20260908T233117Z-p3260462` before broader final
verification. `RESULTS_DIR` was unset: canonical retained-run, scheduler ordering
and warm-run timing maintenance were explicitly skipped because no qualifying
exact-source successful full warm check was supplied. The first finalizer attempt
(`20260908T233001Z-p3256185`) found stale generated topology after the final
recovery-case catalog edit; `make generate` passed at
`20260908T233051Z-p3257324`, then the finalizer passed. Generated outputs were
produced only through public Make targets.

Final owner checks completed so far:

- `make test-slice OWNER=web.networkflow`: PASS 42/42,
  `20260908T233210Z-p3268255`. The final explicit rejected-upload discard case
  also passes in the ownership slice, `20260908T233506Z-p3381533`.
- Network Flow import facade/preview and apply replay owner rows: PASS 1/1,
  `20260908T233210Z-p3268279`.
- Workbook shell assessment/surface regressions: PASS 3/3,
  `20260908T233226Z-p3290770`. Test fixture wiring changes do not change assertions.
- Protocol generated registry/Network Flow decoder rows: PASS 3/3,
  `20260908T233226Z-p3290693`; Network Flow selector contract: PASS 2/2,
  `20260908T233210Z-p3268304`.
- Network Flow real recovery, retained-draft/closure, duplicate-header and
  accessibility rows: PASS 13/13, `20260908T233226Z-p3290620`.
  Recovery also verifies an explicit new-import action does not upload again.
  Duplicate-header diagnostics remain contained at 390x480.
- `make frontend-typecheck`: PASS 2/2, `20260908T233505Z-p3381131`.
- `make lint-biome`: PASS 2/2, `20260908T233226Z-p3290942`;
  `make lint-scripts`: PASS 2/2, `20260908T233353Z-p3376256`.

Full ordinary visual reconciliation `20260908T233210Z-p3268519` accounts for all
210 active captures/goldens and 26 registered fixtures: zero orphans, missing
images, ambiguous mappings or unresolved fixtures. It exposed an additional
unintended lifecycle effect: the analytical binding refreshed incident identity
on closure when no analytical work existed, changing the background grid in three
lifecycle captures. Scoped the refresh to an active retained analytical workflow.
Those unrelated lifecycle goldens are not approved for promotion.

## Owner-to-change map

| Adopted owner | Implementation/projection paths | Behavior |
| --- | --- | --- |
| Network Flow §§5,7–12/18; NF-REQ-054a, 088b/088c, 098–103; Tables 7-B/7-C/11-A | `apps/web/src/networkFlow/NetworkFlowImportController.ts`, `networkFlowImportState.ts`, `networkFlowImportModel.ts`, `useNetworkFlowImportController.ts`, `NetworkAnalysisWorkspace.tsx`, `useNetworkFlowTableController.ts` | Captured analytical stages, inert applicable preview, durable approval comparison, ordinal identity, persisted selection, exact recovery, validated publication and stable-ID selection |
| Core 01 common jobs, upload/Import routes, typed errors and extension idempotency; Core 02 source identity; Core 03 §11.2 | Same analytical controller; existing `imports/importCoordinator.ts`, `imports/importRequests.ts`, `services/importJobContract.ts`; narrow `services/importClient.ts` guard | Shared immutable transport and bounded observation reused; request uncertainty, definite rejection, acknowledged jobs and cancellation remain separate |
| Core 04 authorization; Extensions §18 cleanup/generation | `app/useNetworkFlowImport.ts`, `app/App.tsx`, `workbook/hooks/useNetworkFlowImportBinding.ts`, `WorkbookShell.tsx`, `workbook/features/NetworkFlowFeature.tsx` | Session lifetime, both claims, current actor/incident/role/closure, hidden uncertainty, retirement and dispatch fencing |
| NF-REQ-099/103, Table 11-A; design §§12/14 | `networkFlow/NetworkFlowMappingModal.tsx`, `NetworkFlowImportSurface.tsx`, workbook top-bar/presentation bindings, `packages/ui-contracts/src/networkFlowSelectors.ts` | Local actionable feedback, copy-only recovery, explicit close/cancel, source context, safe samples, preview scope, focus and containment |
| Authored Network Flow mapping registry and schemas | `tools/protocol-ts/generate-protocol-types.mjs`, generated Network Flow mapping registry, protocol entrypoint and service adapter | Runtime timestamp choices/defaults generated from existing machine inputs; no schema, parser, normalization or fingerprint algorithm added to the UI |
| Authored verification routing | `tools/test_families/{web.networkflow,web.application,module.networkflow}.json`, frontend import boundaries, generated browser/topology artifacts, focused unit/browser tests | New ownership/recovery/dispatch and lifetime rows; existing assertions and shared-consumer coverage retained |

Structural changes are the subscribed session owner, focused analytical state
projection, React binding, typed handoff and retained presentation host. Observable
owner-backed corrections are stage recovery, approval/selection independence,
preview fencing, generated profile choices, local feedback, closure copy review,
dispatch authority, honest completion and exact cancellation/replay. The baseline
returned-table fixture now carries valid durable approval/fingerprint and matching
source provenance; the stable-ID selection assertion remains.

## Compatibility, limits and rollback

Actual impact: no public API, route, dependency, durable browser storage or stored
schema/data migration. Existing Import and Network Flow route contracts are used.
The generated timestamp export and optional ImportClient dispatch guard are
internal additive TypeScript seams. Workbook Import Assistant controller and
shared bounded observation/request capture are unchanged. No historical import
or reload recovery is introduced: one current CSV workflow is retained only in the
application session's volatile memory and retired at adopted authority boundaries.
Safe samples and fingerprints remain server-derived. A canceled job is never
presented as rollback of atomic publication.

No saved-graph lifecycle, general table rename/delete, algorithms, linking,
parsers, mapping kernels, backend scheduling or incident-bundle import behavior is
changed. Full backend/CI/release suites and deployment are outside this frontend
seam; owner-routed unit, service-backed browser, shared-consumer and boundary
checks supply the applicable evidence. No reset, commit, push or deployment ran.

For an authorized rollback, revert this coherent source set: analytical owner,
model/modal/workspace/table handoff; application/workbook lifetime and facade
bindings; optional transport guard; protocol generator/export; authored selectors,
boundaries, routing and their tests. Regenerate downstream artifacts through
`make generate`, restore the matching reviewed visual golden set through the
public maintenance workflow, then rerun affected owner slices, type/boundary and
ordinary visual checks. No database or server-job rollback is involved; already
published tables remain server state. Preserve unrelated working-tree changes.

## Binary acceptance

| Criterion | Result | Evidence |
| --- | --- | --- |
| One analytical session owner; no Workbook workflow coupling | PASS | Controller, application binding and architecture boundary checks |
| Immutable stage identities and synchronous admission | PASS | Ownership tests, exact request objects and queued-dispatch reproduction |
| Draft/profile/authority changes invalidate preview; late success and failure fenced | PASS | Ownership tests include edits, workspace departure and authority replacement |
| Preview is inert; duplicate headers retain ordinals and raw header text | PASS | Model/ownership cases and real duplicate-header import |
| Generated profiles, requirements, transforms, timestamps and policies | PASS | Authored registry/schema generator, model checks and protocol typecheck |
| Acknowledged approval compared byte-for-byte before selection/apply | PASS | Fingerprint mismatch, select/apply recovery and returned-table cases |
| Failed selection does not repeat acknowledged approval | PASS | Recovery unit and real browser request counts |
| Uncertain writes retain exact request/ID; changed intent is explicit | PASS | Recovery, queued dispatch, late receipt and new-import cases |
| Acknowledged discovery/apply jobs survive observation and resource failures | PASS | Source-read, observation expiry/resume and browser fault injection |
| Cancellation is admitted once, current-authorized and race-aware | PASS | Exact cancellation replay, cancel_requested and delayed terminal race cases |
| Import success remains distinct from table read and stable-ID selection | PASS | Typed handoff provenance checks and browser table-list recovery |
| Close, discard, workspace pause and server cancellation have distinct effects | PASS | Controller tests and keyboard/workspace departure browser workflows |
| Session uncertainty hides; replacement/access/claim loss retires; closure/write loss retains copyable drafts | PASS | Application lifetime/independent claims, controller authority and real closure rows |
| Resource-specific denial does not declare lost incident access | PASS | Typed denial test and local table-handoff failure projection |
| Local feedback, source context, safe samples, preview-only counts and blocking diagnostics | PASS | Modal inspection, rejected-preview admission and real partial import fixtures |
| Keyboard/focus, non-color status, live announcements, long context/diagnostics and constrained geometry | PASS | Accessibility row, duplicate-header narrow workflow and reviewed screenshots |
| Shared Imports and completed Workbook workflow regressions | PASS | Shared unit 5/5 and production/keyboard browser 11/11 |
| Intentional visual promotion followed by two ordinary scoped passes | PASS | Final visual maintenance below; full-suite instability explicitly limited |
| Finalizer, lint, final-byte and scope checks | PASS | Final closure record below |

The final ordinary reconciliation at `20260908T233652Z-p3387284` has all 210
active capture/golden pairs and 26 registered fixtures, with zero orphans,
missing/ambiguous images or unresolved fixtures. All 37 unrelated visual rows
pass; only the eight approved Network Flow screenshots differ. The closure guard
also passes its real retained-draft row (`20260908T233653Z-p3388718`) and typecheck
(`20260908T233654Z-p3393789`). Application lifetime and frontend boundaries pass
at `20260908T233916Z-p3487302` and `20260908T233915Z-p3486959` respectively.
Pre-completion Markdown lint passed at `20260908T233914Z-p3486009`; diff whitespace
and excluded-path scope checks pass.

## Visual refresh record

Accepted trigger: owner-backed local import progress/recovery, required source
context and explicit retained-import/completion feedback. No renderer, viewport,
zoom, mask, scroll normalization or screenshot-scope change was made to visual
capture code. The existing visual row is
`module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`,
scenario `scenario_252dd2b0ff55`, Chromium. Stable fixture IDs are
`visual.fixture.claimed_network_analysis_workspace_states`,
`visual.fixture.claimed_network_analysis_narrow_workspace` and
`visual.fixture.claimed_network_analysis_compact_workspace`.

The public `make browser-e2e-visual-update` passed 12/12 at
`20260908T234037Z-p3488641`. Reviewed every retained updated image under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

- `network-flow-analysis-mapping-dialog-linux.png`
- `network-flow-analysis-accepted-inspector-linux.png`
- `network-flow-analysis-rejected-diagnostics-linux.png`
- `network-flow-analysis-graph-contributors-linux.png`
- `network-flow-analysis-saved-graph-result-linux.png`
- `network-flow-analysis-narrow-query-controls-linux.png`
- `network-flow-analysis-compact-saved-graphs-linux.png`
- `network-flow-analysis-delete-dialog-linux.png`

Review: source/parser text wraps, ordinal rows remain bounded, focus and footer
controls remain visible/reachable, and surrounding workbook geometry is preserved.
Outside the mapping capture, changes are the retained-import trigger and explicit
import/table-selection status. Graph, query, inspector and delete dialog behavior
is unchanged. Seven promoted images exactly match the already reviewed ordinary
actuals; the inspector image was separately reviewed again.

The whole-directory update also rewrote 21 unrelated PNGs that had passed ordinary
comparison (mostly single-digit pixel differences). Rejected those incidental
refreshes and restored their original baseline bytes; no unrelated golden remains
changed. Regenerated the manifest with public `make generate`. The first ordinary
post-update run passed 12/12 (`20260908T234447Z-p3539255`) before that scope cleanup;
it is supporting evidence, not one of the two final-manifest acceptance runs.

Final authority audit correction: propagate typed ImportFailure through the shell
recovery callback. A server `incident_closed` response now enters copy-only closure
and triggers authoritative incident refresh. Handoff 401 pauses protected state;
403 requests authorization recovery without declaring incident access lost; table
404 preserves import success and remains local. Expanded the existing denial case
for these paths. Focused ownership/recovery/dispatch passes 4/4 at
`20260908T235153Z-p3652136`, typecheck 2/2 at `20260908T235153Z-p3652310`, boundaries
2/2 at `20260908T235153Z-p3652344`, and real recovery/closure 11/11 at
`20260908T235153Z-p3652197`.

An ordinary run at `20260908T234914Z-p3597126` is not acceptance evidence: the
unrelated Membership audit visual row timed out at 60 seconds and its account
preference cleanup could not run after context closure. Four subsequent rows
rendered with the retained density and failed comparison. Network Flow passed.
No goldens were changed for this run. Final fresh ordinary runs follow the final
authority correction and scoped manifest regeneration.

A further full ordinary run, `20260908T235436Z-p3706625`, passed the Network Flow
row and 36/37 unrelated rows. Only `lifecycle-compact.png` differed: inspection
shows a six-pixel drawer scroll-position shift in the existing lifecycle fixture.
The analytical controller is unavailable in that base-profile fixture; no import
work or analytical lifecycle refresh occurs. This is retained as full-suite visual
instability, not repaired or promoted within this seam. Final visual acceptance
uses two fresh owner-routed ordinary Network Flow visual runs against the final
manifest. They execute the same authored capture row, pinned renderer and ordinary
comparison mode through `make service-backed-test-slice`; they are not update runs.
Full-suite visual stability is a limitation, and no two-full-suite-pass claim is
made.

Final source finalizer: PASS 1/1, `20260908T235333Z-p3702630`, with retained-run
maintenance skipped because RESULTS_DIR is unset. Final Biome lint: PASS 2/2,
`20260908T235426Z-p3706038`. The scoped golden manifest was generated at
`20260908T234803Z-p3590469` and only the eight listed Network Flow PNGs remain
changed.

Final ordinary owner-routed visual pass 1: PASS 11/11,
`20260908T235912Z-p3758409`, using the unchanged authored Network Flow visual row.
The final scope audit covers 48 changed/new files: eight Network Flow PNGs, the
analytical seam and its application/workbook bindings, narrow shared dispatch and
projection updates, verification inputs/outputs and this handoff. Excluded owners,
the digest, public contracts, database code, dependencies and shared Workbook
controller remain unchanged. `git diff --check` passes.

Final ordinary owner-routed visual pass 2: PASS 11/11,
`20260909T000040Z-p3806044`. Both fresh passes compare the nine existing Network
Flow captures against the final scoped manifest, with no missing or ambiguous
matches and no unresolved registered fixture. The scoped reconciliation labels
201 unselected goldens as orphan; the earlier full reconciliation accounts for
all of them, and every one is retained. This is two ordinary scoped passes, not
an assertion that the unrelated full visual suite is stable.

Pre-completion `make lint-markdown` passed at `20260909T000041Z-p3806331`.
Whitespace and the final 48-file scope audit pass. NI-05 is DONE: owner-routed
acceptance, reviewed visual promotion, compatibility/limitations, rollback and
all tracker rows are complete. Final-byte verification repeats `make lint-markdown`,
`git diff --check` and the excluded-path/golden scope audit after this completion
record. No public API or stored-data migration, commit, push or deployment occurred.

Next action: none. Any work on the unrelated full-suite visual instability requires
separate scope authorization.
