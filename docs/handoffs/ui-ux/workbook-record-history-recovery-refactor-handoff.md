# Workbook record-history mutation recovery handoff

## Baseline and authority

- Branch: `main`; implementation baseline: `774a0845b38fa1323896b6a84a856f4847d3d4d4`.
- Initial dirty state: clean, reconfirmed at implementation entry.
- Root `AGENTS.md` applies. The localized digest read order was completed during
  planning and mappings were revalidated against this unchanged HEAD.
- Core 01 §3.3.4.2, REQ-01-071–082 and §3.3.5.0 own history, operation-specific
  authorization, exact selectors, receipts, concurrency and replay. Ordinary
  DELETE is outside the destructive-lock family (REQ-01-104).
- Core 02 §§14–15 own retention, source reconstruction and affected records;
  source-owner semantics remain behind current Revisions provider boundaries.
- Core 03 REQ-03-035/282/283, §§4.2–4.4 and §10 own committed versions,
  sequencing, continuity, save status, bounded autosave and reviewer history.
  REQ-03-287/288/301/302 govern closure and transaction recovery.
- Core 04 §§1–2 and REQ-04-127 own current authorization, CSRF, concealment and
  session lifecycle. Design §§10 and 14 own safe local feedback and accessibility.
  Domain and NLSpec research documents supply vocabulary and explanation only.
- No adopted-owner contradiction or required backend correction identified.

Permitted paths: history-related workbook source, necessary app/runtime and
transport composition, focused tests/browser support, authored selector and
verification inputs with Make-generated derivatives, reviewed affected visual
goldens, and this handoff. Preserve the digest and completed handoffs. No new
routes, persistence, dependencies, migrations, rollback algorithm, or workflow
engine. No commit, reset, push or deployment.

## Tracker

Only the current workstream is `IN_PROGRESS`. A blocked predecessor prevents
advancement. Every exit records changes, verification, compatibility, risks and
next action.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| HR-01 Baseline and characterization | DONE | Current paths and operation matrix inspected; acknowledgement-before-refresh characterization fails as expected. |
| HR-02 Admission and attempts | DONE | Shared owner and captured transport tests pass; frontend types pass. |
| HR-03 Surface integration | DONE | Shared controller, runtime retention, Timeline coordination and acknowledgement tests pass. |
| HR-04 Recovery and browser validation | DONE | Recovery, service/browser and keyboard checks pass; three promoted goldens reviewed. |
| HR-05 Final verification | DONE | Finalizer, focused/service/browser checks, two fresh visual passes and scope audit complete; unrelated baseline gate failures documented. |

## Operation and surface gap ledger

| Surface | Operation | Current admission and captured request | Outcome / acknowledgement gaps | Refresh and lifetime gaps |
| --- | --- | --- | --- | --- |
| Timeline | Delete | Broad not-viewer UI gate; no dispatch role recheck; queued committed version replaces confirmed version; duplicate submit transition ignored. | ID allocated before queued work, then discarded; transport failure called rejection; receipt reduced to record/version. | Local history state; selection/refresh callbacks can outlive origin; history read can displace operation. |
| Timeline | Tombstone restore | Same broad gate incorrectly includes editor; tombstone fallback exists. | No retained exact body or uncertain replay; deletion tuple discarded. | Restore may select stale origin after refresh; closure/navigation do not own durable recovery. |
| Timeline | Three rollback selectors | Advertised opaque selectors correctly preserved; queued base substitution and missing actual-dispatch gate. | Target, rollback change set and affected records discarded. | No explicit refresh recovery or multi-record receipt accounting. |
| Generic | Delete | Broad not-viewer gate; record/version captured; generated ID remains private to command invocation. | No exact replay; owner effects awaited before acknowledgement. | Disposable inspector owns attempt; refresh exception strands submitting. |
| Generic | Restore / three rollback selectors | Editor incorrectly admitted; advertised selectors correctly preserved. | Same missing receipt and uncertainty distinction. | Captured owner effects can select after retarget or panel closure. |
| Entity | Delete / restore / three rollback selectors | Shared controller has the same gaps; current entity delete cleanup is source-specific. | Same premature receipt reduction and acknowledgement ordering. | Entity refresh, deleted subject and selection remain narrow adapter responsibilities. |
| Assessment | Delete / restore / three rollback selectors | Shared controller also inherits unrelated creation admission; exact route roles required. | Same missing immutable identity and uncertain outcome. | Assessment cleanup and row refresh must stay owner-specific. |

All rows require actor/incident/lifecycle capture, current dispatch and recovery
authorization, exact serialized replay, bounded observation, safe local feedback,
and separate mutation settlement and refresh state. Preserve legal-action
discovery, stable selectors, confirmation invalidation, tombstone reads and
source effects already correct.

## Adopted implementation decisions

- One dedicated history owner in the existing registry-managed workbook lifetime;
  history actions never enter the 64-unit autosave FIFO.
- Retain attempts per record. Other records stay usable during uncertainty.
- A compact `History actions` entry opens recovery explicitly without changing
  surface or stealing focus. Protected presentation suspends on authorization loss.
- Confirmation distinguishes tombstone restore, historical row-field restore,
  single-entry reversal and change-set reversal.
- Acknowledgement precedes refresh; refresh recovery issues no mutation.
- Same-account reauthentication retains attempts in memory. Account/incident
  replacement retires them. No reload/tab-close durability is promised.

## Verification ledger

Planning baseline (unchanged source): selected workbook history/save-status slice
passed 5/5 execution units at
`.cartulary/test-results/20260910T105129Z-p55418`; selected Revisions history and
selector slice passed 3/3 at `.cartulary/test-results/20260910T105143Z-p56297`.
Implementation-entry task guides for `web.workbook` and `module.revisions` passed.
Implementation and final verification evidence follows in the workstream exits.

## Compatibility and rollback

Public contracts and source-owner semantics remain unchanged unless a demonstrated
adopted-owner mismatch requires a separately documented correction within this
seam. Rollback means reverting coordinated code/artifact changes; never deleting
analyst history, idempotency receipts, revisions or user data.

### HR-01 exit

Added the routed acknowledgement-before-refresh characterization under workbook
inspector tests. `make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.history_recovery_characterization` failed as expected
at `.cartulary/test-results/20260910T110217Z-p76687`: the successful mutation stays
submitting until projection refresh settles. This is related baseline defect
evidence, not a waived check. Authored routing was added; no product contract or
generated file changed. Next: implement durable admission and receipt ownership.

### HR-02 exit

Added a dedicated history owner, typed immutable attempts/receipts and shared
captured transport. Extracted existing bounded observation mechanics into the
service layer; the app export preserves existing consumers. Owner tests cover
five operations, exact replay, role/closure gates, changed review, late receipt,
account replacement, explicit re-key and acknowledgement/refresh separation.
Transport tests cover exact bytes, CSRF, receipt identity, malformed success,
network/server uncertainty and definitive rejection.

`make test-slice OWNER=web.workbook` selecting `history_operation_owner` and
`history_captured_transport` passed 3/3 units at
`.cartulary/test-results/20260910T111142Z-p81540`. `make frontend-typecheck` passed
at `.cartulary/test-results/20260910T111106Z-p80905` after correcting test types.
The first transport test run failed because its error fixture omitted required
`status`; corrected the fixture and made malformed rejection remain uncertain.
No public/backend interface changed. Surface integration remains next.

### HR-03 exit

Composed the shared owner into the registry-managed mutation runtime, forwarded
session/account/incident lifecycle, and projected its unsettled work separately
from FIFO units. All four history surfaces now use the shared controller.
Timeline retains explicit save sequencing through a cancellable coordination
adapter; its duplicate transport and port were removed. Record command ports
expose the shared captured-request adapter. Source-specific cleanup is separate
from projection refresh, and restored Timeline subjects survive the projection
reload interval. Presentation effects are fenced against detachment/retargeting.

Updated existing fixtures for current reviewer authorization, extra current-state
reads and complete receipts. Added authored routing for previously unrouted
inspector and Timeline sequencing evidence. Selected shared history, command-port,
foundation and sequencing checks passed 8/8 units at
`.cartulary/test-results/20260910T113956Z-p13624`. Timeline action-adapter checks
passed 2/2 at `.cartulary/test-results/20260910T113957Z-p13882`. Delete/restore
continuity passed at `.cartulary/test-results/20260910T113806Z-p7922`.
Frontend types passed at `.cartulary/test-results/20260910T113846Z-p8623`;
Make formatting passed at `.cartulary/test-results/20260910T113938Z-p9352`.
Earlier integration failures exposed stale ordered fixtures and the restored
subject gap; both were corrected. No public HTTP/backend contract changed.
Next: compact recovery, expanded lifecycle/effect coverage and real browser/service
validation. Broad lint, catalog/drift and finalization remain unexecuted/pending.

### HR-04 visual review before promotion

The ordinary full visual run at
`.cartulary/test-results/20260910T120938Z-p28789` completed all functional
assertions and found only three intended screenshot differences. Its
`browser-e2e-visual/frontend-visual-reconciliation.json` accounts for all 210
active captures/goldens, all 26 registered fixtures, zero orphans, zero missing
goldens and zero ambiguous mappings.

Accepted trigger: requested operation-specific labels and explicit workbook
recovery. Owner row:
`module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`;
fixture `visual.fixture.base_inspector`, design fixture `D-VFIX-002`, Chromium
scenario `scenario_cb54783764e3`. Affected filenames:
`workbook-inspector-history-linux.png`,
`workbook-inspector-rollback-preview-linux.png`,
`workbook-inspector-public-error-linux.png`. Reviewed actual/diff images show
readable distinct actions, the compact recovery entry and intact focus outlines.
The obsolete rollback paragraph mask was removed because it replaced the actual
confirmation with retired internal selector text. Current confirmation contains
no dynamic selector identity and needs no paragraph mask. Viewports (1280×720),
zoom (100%), density, scroll anchors, screenshot scope and tolerances are unchanged.
Classification: accepted intentional state / stale goldens. Promotion and two
fresh ordinary visual passes remain required.

### HR-04 exit

Added compact, keyboard-accessible recovery outside disposable inspectors,
current-state review, exact replay, explicit transaction-conflict replacement
confirmation and refresh-only completion recovery. All four surfaces retain
per-record attempts across inspector closure. Authorization/session changes hide
protected content and fence callbacks; admitted work is distinct from disposable
presentation. Removed obsolete Timeline request counters and kept sequencing
behind its explicit coordination adapter. Related rollback projection failures
retain acknowledgement; timed-out reconciliation cannot publish late local effects.

Deterministic coverage includes all four surfaces × delete/restore/three rollback
selectors, exact replay identity and mutation/change-set/revision counters, secure
ID failure, role/lifecycle gates, peer version/eligibility changes, duplicate
admission, late receipts, session replacement, refresh failures and callback
fencing. Selected history/runtime/save-status checks pass 30/30 execution units
at `.cartulary/test-results/20260910T121637Z-p49421`. Expanded transport checks
retain lifecycle tuples and all rollback receipt fields.

Real service/browser evidence:

- Eight new Revisions browser scenarios pass 11/11 units at
  `.cartulary/test-results/20260910T121544Z-p13758`. Full server-history equality
  before/after replay proves no duplicate mutation/change set/revision. These cover
  four-surface committed delete recovery and tombstone restore, pre-commit loss,
  malformed success, replay after later peer changes, and refresh-only recovery.
- Existing delete/restore, row-restore, history-entry and merge-change-set browser
  scenarios pass 11/11 at `.cartulary/test-results/20260910T115035Z-p24215`.
- Eight selected backend security, route, lock, rollback failure-injection and
  source reconstruction rows pass 4/4 execution units at
  `.cartulary/test-results/20260910T115637Z-p26870`.
- Existing keyboard/announcement/contrast inspector scenario passes 11/11 at
  `.cartulary/test-results/20260910T120922Z-p97537`. Explicit recovery summary focus
  also passes in the real Timeline recovery scenario at
  `.cartulary/test-results/20260910T122034Z-p63545`.
- Golden update passes 12/12 at
  `.cartulary/test-results/20260910T121439Z-p77684`; only the three reviewed PNGs
  and their Make-owned manifest changed. Every promoted image and four-surface
  recovery screenshot was inspected: readable feedback, intact controls and no
  clipping. Required two fresh ordinary visual runs are HR-05 final checks.
- Frontend types and import boundaries passed at
  `.cartulary/test-results/20260910T121302Z-p71978` and
  `.cartulary/test-results/20260910T121302Z-p71982`; subsequent types and Biome
  passed at `20260910T121636Z-p49006` and `20260910T121636Z-p49016`.

Intermediate browser failures identified opening history before selection
retarget, premature fixture closure before dispatch, and recovery-origin
admission tied to disposable presentation. Those paths were repaired and rerun.
No assertions or tolerances were weakened. Public routes, backend behavior and
FIFO semantics remain unchanged. Next: finalizer and final exact-source checks.

### HR-05 ownership and scope audit

Final source ownership:

- `web.workbook`: the dedicated history owner, immutable attempt/receipt types,
  recovery presentation and captured transport; existing inspector presentation,
  focus and controller; runtime/save-status composition; Timeline coordination;
  Generic, Entity and Assessment source-effect adapters.
- `web.services`: existing bounded observation mechanics extracted into
  `apps/web/src/services/asyncObservation.ts`. The existing app export in
  `apps/web/src/app/accountOperation.ts` preserves its consumers and deadline.
- `module.revisions`, `module.timeline` and `module.workbook`: unchanged adopted
  backend/source semantics, exercised through existing routed service evidence.
  New service/browser history scenarios are authored under `module.revisions`.
- Verification ownership remains in authored `tools/test_families/` and
  `tools/frontend_source_ownership.json`. Browser batches, topology render index
  and the three golden-manifest entries were updated only through Make.

Changed path inventory (all within the permitted seam):

- New `apps/web/src/workbook/history/`: `WorkbookRecordHistoryOwner.ts` and test,
  `workbookHistoryOperation.ts`, `WorkbookHistoryContext.ts`,
  `WorkbookHistoryRecovery.tsx` and test, `WorkbookHistoryLocalStatus.tsx`, and
  `historyOperationPresentation.ts`.
- New `apps/web/src/workbook/adapters/createWorkbookRecordHistoryAdapter.ts`
  and its test; new inspector
  `workbookHistoryRecovery.characterization.test.tsx`.
- Existing inspector controller, record-history panel and test, loaded
  presentation, focus, history presentation model and test, compatibility
  operation bridge, and source-owner effects.
- Existing `WorkbookShell.tsx` and history test; shell infrastructure and surface
  queries; mutation command-port factory and test, command-port types;
  `WorkbookMutationRuntime.ts` and status projector; the three shared-surface
  inspector composition files under `features/{generic,entities,assessments}`.
- Timeline history actions/state, committed-row/idle hooks, inspector selection,
  history model, inspector/surface/workbook composition and focused lifecycle,
  foundation and adapter tests. Removed only its duplicate history adapter and
  `TimelineHistoryPort.ts`. Updated the Timeline runtime testing fixture.
- New `apps/web/e2e/history-recovery.spec.ts`; existing
  `inspector-actions.spec.ts` and `workbook.visual.spec.ts`; the three PNGs named
  in HR-04. No viewport, tolerance or unrelated golden changed.
- Authored `tools/test_families/{web.workbook,module.revisions}.json` and
  `tools/frontend_source_ownership.json`; Make-owned
  `tools/browser_e2e_batch_manifest.json`,
  `tools/execution_topology_render_index.json` and
  `tools/frontend_visual_golden_manifest.json`; this handoff.

Final admission audit removed Assessment's unrelated creation-availability gate.
Final authorization audit preserves a generic access-unavailable alert while
hiding history, technical details, confirmations and recovery after access loss.
The stateful browser assertion verifies both safe feedback and concealment.
The public rollback race now verifies two distinct outcomes: a peer edit before
the final read prevents dispatch locally; an edit after that read reaches the
server and receives `row_version_conflict`, with the reviewed base unchanged.

### HR-05 final verification and limitations

`make agent-finalize` passed at
`.cartulary/test-results/20260910T124527Z-p37068`. `RESULTS_DIR` was unset:
retained-run maintenance was skipped because no qualifying successful full warm
check exists for this exact source. An earlier finalizer run exposed stale
generated routing; Make generation repaired it, and subsequent finalizers passed.

Final focused checks:

- Final combined `make service-backed-test-slice OWNER=module.revisions`,
  selecting all 15 authored browser/history rows (eight new recovery scenarios,
  five existing history scenarios and two inspector authorization/conflict
  scenarios), passed 17/17 execution units at
  `.cartulary/test-results/20260910T125006Z-p16941`. This supersedes the
  intermediate browser failures described below.
- `make test-slice OWNER=web.workbook` selecting the six new `history_*` rows
  passed 7/7 units at `.cartulary/test-results/20260910T124445Z-p3505`.
  The prior 30/30 history/runtime/save-status slice remains regression evidence.
- `make service-backed-test-slice OWNER=module.revisions` selecting eight new
  recovery scenarios and two existing inspector scenarios passed the eight new
  scenarios and the strengthened public rollback conflict scenario at
  `.cartulary/test-results/20260910T123828Z-p57372`. That run failed only the old
  stateful access-denied feedback expectation. After the safe-feedback fix, the
  isolated stateful row passed 11/11 units at
  `.cartulary/test-results/20260910T124445Z-p3506`.
- Existing history-browser coverage (five scenarios) passed in
  `.cartulary/test-results/20260910T123249Z-p9259`; other rows in that run exposed
  stale recovery-region selectors and old transaction-prefix expectations.
  Corrections use exact accessible-region matching and a strict shared
  operation-prefix/Web Crypto UUID assertion; they do not relax validation.
- Keyboard, announcement and contrast coverage passed again at
  `.cartulary/test-results/20260910T122629Z-p81292`. New browser recovery also
  asserts explicit summary focus and absence of focus theft.
- `make frontend-typecheck` and `make lint-biome` passed on final product source
  at `20260910T124508Z-p34198` and `20260910T124508Z-p34220`, respectively.
- `make frontend-import-boundary-check` passed 2/2 at
  `.cartulary/test-results/20260910T124551Z-p41037`; generated-artifact policy
  and JSON-shape checks passed 3/3 each at `20260910T124551Z-p40812` and
  `20260910T124551Z-p40820`. `make test-catalog-check` passed independently and
  as an executed finalizer substep.
- `make protocol-ts-browser-artifact-reachability` passed (exit 0) after the
  final visual builds. The separate dead-code gate limitation is recorded below.
- `make generate-drift` passed 4/4 at
  `.cartulary/test-results/20260910T124948Z-p12771`.
- Two fresh ordinary `make browser-e2e-visual` runs passed 12/12 each on final
  product source at `.cartulary/test-results/20260910T124551Z-p41120` and
  `.cartulary/test-results/20260910T124551Z-p41131`. Both account for all 210
  active captures/goldens and 26 fixtures, with no orphan, missing golden or
  ambiguous mapping. No second promotion was needed.
- `make lint-markdown` passed at
  `.cartulary/test-results/20260910T124741Z-p10627`; `git diff --check` passed.
  Repeat both after the final tracker edit together with scope/status inspection.

The full `make frontend-unit` run at
`.cartulary/test-results/20260910T122557Z-p22556` passed 529/533 units. All history
units passed. Remaining failures were investigated against clean baseline HEAD
in an isolated detached worktree; its retained summaries are copied under
`.cartulary/history-baseline-evidence/`. That temporary worktree was removed
without changing the working branch or user files.

- Baseline `20260910T123207Z-p7428` reproduces six existing Network Flow selector
  violations and 25 unaccounted source paths. All newly added/deleted history
  paths and the touched observation service are now accounted for. The current
  ownership check at `20260910T123432Z-p51154` retains only the other 24 baseline
  inventory omissions; unrelated source inventory was not expanded.
- Baseline `20260910T123226Z-p8373` reproduces the Network Flow production grid
  contract expectation failure (two expected fields versus five returned fields).
- A Network Flow focus test timed out under concurrent full-suite load; isolated
  baseline `20260910T123224Z-p8126` and current
  `.cartulary/test-results/20260910T123318Z-p38985` both passed.
- `make protocol-ts-dead-code-check` fails with the same four findings at baseline
  `20260910T123155Z-p7084` and current
  `.cartulary/test-results/20260910T122708Z-p39831`: existing `RecordHistoryItem`
  and `ViewCell` reexports, and `ViewCell` / `ViewInspectorRegistry` frontend
  entrypoint usage. These were already unused at HEAD; unrelated generated
  protocol contracts were not changed to silence the gate.
- An expanded browser run at `20260910T122629Z-p81408` failed during concurrent
  frontend artifact publication (`ENOENT`) before valid product evidence.
  Serialized browser reruns and the later visual runs use successful builds.

These baseline failures are reported limitations, not passing verification or
an adopted-owner contradiction. Completed Network Analysis work is not reopened.
Full release/check, retained-run maintenance and unrelated owner suites are not
claimed. No backend, contract, database, route, dependency, lockfile, digest or
completed tracker changed. Tests and generated evidence remain independent of
Markdown.

Compatibility: HTTP routes/schemas, opaque selectors, source reconstruction,
destructive locks and autosave FIFO semantics are unchanged. Attempts are
memory-only for the active workbook/incident runtime. Closing a panel or timing
out observation never implies server rollback. Explicit replay retains exact
bytes and identity; completed refresh recovery performs reads only. All local
state publication is fenced by current authority and presentation, with accepted
record versions applied monotonically. No unresolved owner decision remains.

Implementation rollback reverts the coordinated source, authored verification,
generated derivatives and reviewed golden changes. It must preserve analyst
history, route idempotency receipts, revisions and user data.

Final branch remains `main` at `774a0845b38fa1323896b6a84a856f4847d3d4d4`.
The dirty state consists only of the coordinated paths inventoried above;
baseline was clean. No commit, reset, push or deployment occurred. HR-01 through
HR-05 are complete with the explicitly documented baseline verification limits.
Next action: review the coordinated implementation and evidence; do not expand
into another refactor seam. After this final tracker update, rerun Markdown lint,
whitespace checks and scope/status inspection as required by the task.
