# Workbook Timeline capture actions refactor

## Authority, baseline, and scope

This handoff records the authorized Timeline capture-state seam only. Core 01
REQ-01-083–088 and inspector §7.4 own routes and composition; Core 03
REQ-03-102–108, REQ-03-282, and REQ-03-292 own lifecycle, committed intent,
and continuity. Core 02 §13 and §15 own typed links, attribution, and rollback;
Core 04 current session, CSRF, membership, and role rules remain authoritative.
The UI/UX digest is advisory. Its localized read order was followed and mappings
were checked against current source. Design §§5, 7.3, 10, 12, and 14 guide local
feedback, density, and accessibility. No owner contradiction was found.

Implementation baseline: clean `main`, HEAD
`c1b0bd637b21ebb91e26a50777da9aa1de155151`, empty index. Root `AGENTS.md` is
the only applicable repository instruction file. Go 1.27.1, Node 24.15.0,
pnpm 10.33.0, ShellCheck 0.11.0, Docker Compose 5.1.3 were verified by
`make doctor`, PASS at `.cartulary/test-results/20260911T171417Z-p61104`.
Doctor's partial `/proc` inotify inspection warning did not fail readiness.

Permitted changes: Timeline frontend actions and necessary inspector, runtime,
transport, query, history/continuity integration; focused tests and browser
support; authored UI selectors and verification inputs; Make-generated
derivatives; reviewed affected visuals; this handoff. Existing public interfaces,
stored data, backend lifecycle classification, and routes remain unchanged.
No commit, reset, push, deploy, dependency change, migration, or unrelated seam
is authorized. User changes must be preserved.

Task, Decision, merge, history, save-status, and artifact-publication work are
regression baselines. The Task handoff records final `make check` PASS 835/835
units and 1223/1223 rows; no older waived-failure assumption is inherited.

## Tracker

| Row | Status | Exit criterion |
| --- | --- | --- |
| TC-01 | COMPLETE | Baseline, owner/gap ledger, and failing characterization recorded. |
| TC-02 | COMPLETE | Canonical actions, optional candidates, authored reason, frozen review verified. |
| TC-03 | COMPLETE | Sequencing, retained admission/transport, exact replay, and receipt checks verified. |
| TC-04 | COMPLETE | Reconciliation, accessibility, and real service/browser evidence verified. |
| TC-05 | COMPLETE | Final checks, scope audit, and evidence handoff complete. |

Only the current row may be IN_PROGRESS. Each exit records paths, commands,
results, compatibility, risks, and next action before advancing.

## Gap ledger

| Gap | Owner | Remediation and affected areas | Rationale and long-term benefit | Compatibility and unresolved risk | Binary completion |
| --- | --- | --- | --- | --- | --- |
| Required raw replacement ID | Core 01 REQ-01-086; Core 03 REQ-03-106 | Timeline selector and action adapter accept no replacement or an authorized paginated candidate. | Analysts can supersede obsolete facts without manufacturing replacement links. | Existing optional request member; stale candidates remain server-validated. | Omitted member and selected member both persist correctly; out-of-grid candidates selectable. |
| Fixed reason and uncaptured confirmation | Core 01 inspector; Core 03 REQ-03-106/292 | History authoring, reason_note_v1 validation, immutable review. | Preserve analyst intent and attribution through asynchronous work. | No new reason field or approval workflow; version changes invalidate review. | Authored reason persists; stale confirmation sends no mutation. |
| Separate row-menu mutation and missing inspector admission | Core 01 inspector registry; Core 03 REQ-03-104/106; Core 04 §2 | Canonical capability resolver and one Timeline controller. | All entry points share semantic routing and eligibility. | Inspect/History remain available to readers. | Exact canonical actions appear once, enforce roles/state, and shortcuts cannot bypass confirmation. |
| Silent version substitution after writes | Core 03 REQ-03-282/292 | Preparation reservation, existing save boundary, committed-row observation. | The acted-on version is the version the reviewer saw. | No autosave queue or lock-family change; drafts retained. | Changed version requires another review click or supersession review. |
| Silent rejection and discarded uncertainty | Core 01 REQ-01-085/087; Core 03 transaction recovery | Dedicated runtime action owner and immutable transport attempts. | Retain recovery independently of disposable presentation. | No persistence or automatic replay; current authorization still applies. | Lost/malformed responses retain exact replay; duplicate admission sends once. |
| Detached completion and refresh ambiguity | Core 01 REQ-01-088; Core 03 REQ-03-292 and query continuity | Receipt-first acknowledgement, version ledger, socket accounting, fenced reconciliation. | Avoid false failure, stale overwrite, and focus theft. | Existing history rollback; no filter changes or forced replacement selection. | Refresh-only retry sends no mutation; late receipts are monotonic; access loss conceals content. |

Advisory classifications: ADOPT keyboard parity, local recovery, explicit async
states, and stable identities; ADAPT confirmation to the declared History flow;
REJECT generic re-theme, workflow engine, or new behavior inferred from advice.

## TC-01 log

Read and revalidated all named Timeline entry points, view schema, backend action
admission/store, generated HTTP types, current runtime/observation/transaction
mechanics, committed-row ledger, inspector dispatcher, and routed tests.
`make task-guide ROLE=module-author OWNER=web.workbook`, `OWNER=module.timeline`,
`OWNER=module.revisions`, and `OWNER=module.workbook` all PASS. Sequencing tests
are currently owned by module.workbook, with web.workbook collaboration.

Planning baseline commands:

- `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_action_ports_73c9f1a40e`:
  PASS 2/2, `.cartulary/test-results/20260911T171810Z-p62970`.
- `make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8992eb8931`:
  PASS 2/2, `.cartulary/test-results/20260911T171810Z-p62976`.
- `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.store.a_material_edit_demotes_reviewed_enriched_legal_29bb8697de,module.timeline.store.supersede_with_replacement_rejects_illegal_targe_e4d81b8a06`:
  PASS 3/3, `.cartulary/test-results/20260911T171823Z-p63739`.

User decisions: a newer committed version during Mark reviewed requires another
click without a confirmation dialog; Supersede authoring and frozen confirmation
stay in History. Next: capture failing characterization before implementation.

TC-01 exit: added
`apps/web/src/workbook/timeline/timelineCaptureActions.characterization.test.ts`
and its authored web.workbook routing. `make generate` initially failed at
`.cartulary/test-results/20260911T172523Z-p82690`: the new row was not
ASCII-sorted. Corrected authored ordering; `make generate` PASS at
`.cartulary/test-results/20260911T172638Z-p86142`. Characterization command
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_capture_characterization`
failed as expected, 1/2 units at
`.cartulary/test-results/20260911T172655Z-p89316`: canonical actions absent and
request body substitutes a fixed reason and empty replacement. These failures
are related to the confirmed seam, not waived baseline failures. Compatibility
unchanged; admitted recovery and browser behavior remain unimplemented risks.
Next: TC-02 canonical composition and authoring, verified against focused tests.

## TC-02 log and exit

Added Timeline action policy, reason normalization, independent bounded candidate
reader and paginated hook, and History supersession authoring/review component
under `apps/web/src/workbook/timeline/actions` and `timeline/adapters`.
The canonical inspector resolver now admits exactly the two Timeline semantic
actions, retaining all Decision/other capability assertions. Added stable UI
selectors. Optional replacement omission and authored reason replace the old
wire substitution; complete retained dispatch wiring is the dependent TC-03 work.

`make generate` PASS `.cartulary/test-results/20260911T173307Z-p94736`.
`make format` PASS `.cartulary/test-results/20260911T173356Z-p99159`.
`make frontend-typecheck` PASS 2/2
`.cartulary/test-results/20260911T173422Z-p4153`.
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_capture_authoring,web.workbook.regression.timeline_capture_characterization,web.workbook.regression.decision_supersession_canonical_composition`
PASS 4/4 `.cartulary/test-results/20260911T173422Z-p4086`.

Earlier generation/format/typecheck attempts rejected unsorted authored test
titles (generation root `20260911T173254Z-p91482`); sorted those inputs. First
authoring/typecheck runs used unavailable DOM matchers (roots
`20260911T173331Z-p98243`, `20260911T173316Z-p97634`); replaced them with native
DOM property/focus assertions. Both failures were related to new test inputs
and are repaired without weakening assertions.

Compatibility: request/schema/backend unchanged. Component evidence covers reason,
no replacement/selection, confirmation invalidation, duplicate preparation,
read failure, and pagination. Remaining risk is real runtime admission and
completion wiring, addressed next in TC-03 and TC-04.

Architecture decision for TC-03: the common runtime cannot import Timeline
implementation under the existing import-boundary owner. Retain the concrete
Timeline action owner through a narrow injected lifecycle/status/version port;
construct it in Timeline/app composition. Keep the boundary intact.

## TC-03 log and exit

Implemented the retained Timeline owner and immutable capture/send port under
`timeline/actions`, `timeline/adapters`, and `timeline/ports`. The common runtime
retains its narrow injected port, includes action status outside autosave counts,
and observes socket versions before transaction suppression. App composition and
the Timeline fixture configure current authority, transport, reconciliation, and
persistent local recovery. Ordinary mutation commands no longer own actions or
replacement drafts. Both row shortcuts enter canonical History composition.

Preparation reserves synchronously, observes the captured earlier save boundary,
waits for committed row idle, and preserves unfinished drafts. Changed committed
versions require a fresh action. Supersession freezes its authored reason and
optional replacement and rechecks selected replacement visibility/version through
independent bounded query pages. Secure identity and serialized request are
captured only after preparation. Exact replay bypasses fresh capture-state
eligibility, retains its original route/body/version/identity, and distinguishes
uncertainty from definitive rejection. Receipts precede refresh, remain available
through panel closure and same-account recovery, and cannot lower version marks.

Verification:

- `make generate`: PASS `.cartulary/test-results/20260911T175818Z-p26900`.
- `make format`: PASS 2/2 `.cartulary/test-results/20260911T175844Z-p30044`.
- `make frontend-typecheck`: PASS 2/2
  `.cartulary/test-results/20260911T175857Z-p34610`.
- `make frontend-import-boundary-check`: PASS 2/2
  `.cartulary/test-results/20260911T175313Z-p18146`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_capture_recovery,web.workbook.regression.timeline_capture_transport,web.workbook.regression.timeline_capture_authoring,web.workbook.regression.timeline_capture_characterization`:
  PASS 5/5 `.cartulary/test-results/20260911T175857Z-p34485`.
- `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_action_ports_73c9f1a40e`:
  PASS 2/2 `.cartulary/test-results/20260911T175857Z-p34490`.
- `make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8992eb8931`:
  PASS 2/2 `.cartulary/test-results/20260911T175706Z-p25254`.

Related intermediate failures: typechecks `20260911T174908Z-p8837` and
`20260911T175714Z-p26013` exposed wiring/test type errors; formatting roots
`20260911T175052Z-p9941` and `20260911T175648Z-p20946` exposed missing hook
dependencies and a Promise condition in the new test. Lint roots
`20260911T175431Z-p19972` and `20260911T175713Z-p25832` identified those same
issues plus formatting and native assertion cleanup. All repaired; none waived.
The existing sequencing assertions remain, with authored reasons and renewed
review replacing obsolete fixed-reason/raw-ID assumptions.

Exit: focused preparation, race, exact replay, late settlement, session retirement,
operation-specific malformed receipt, and refresh-only recovery tests pass.
Public routes, backend production, contracts, and stored data are unchanged.
Remaining risk: real service/browser reconciliation, accessible focus and affected
visuals. Next: TC-04 service/browser evidence and continuity validation.

## TC-04 work log

Real service browser evidence now forwards the exact mutation, loses responses
before dispatch and after commit (including malformed success), explicitly replays,
and compares complete histories, receipts, serialized requests and replacement
link history-unit counts. Both supersession variants reverse through the existing
History change-set rollback UI. A reviewed replacement is discovered outside the
active rough filter; completion removes the target without changing the filter.
Keyboard recovery survives inspector closure and filtered-row exit.

`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.capture_action_exact_recovery`
PASS 11/11 `.cartulary/test-results/20260911T182235Z-p42411`.
`make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8992eb8931`
PASS 2/2 `.cartulary/test-results/20260911T182235Z-p42401` (five cases).
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_row_action_overlay_opens_from_pointer_a_d6ad456872,web.workbook.regression.timeline_inspector_lifecycle_close_71780eefa1`
PASS 3/3 `.cartulary/test-results/20260911T182418Z-p78706`.

The added selection race exposed a stale selected-row closure in Timeline's
accepted autosave projection. Read the current selection when applying an accepted
row. The filtered-exit race exposed unconditional fallback focus; preserve a live
focused control outside the grid. Both fixes remain within Timeline continuity.
Timeout accounting settles pending transaction entries while retaining immutable
action recovery; a validated late receipt cannot block unrelated editing.

Visual maintenance follows the current guide. Ordinary full run
`make browser-e2e-visual` failed 10/12 units at
`.cartulary/test-results/20260911T181237Z-p60678`, with screenshot differences only.
Reconciliation accounts for 217 intents, 213 active existing goldens, four newly
authored missing captures, zero orphans, zero ambiguous mappings, and all 26
registered fixtures. The four missing captures are the newly authorized Timeline
supersession frames, to be created through the public candidate-update target.
No existing missing or orphan golden is bypassed or removed.

Accepted refresh trigger: implement adopted canonical Timeline History actions
and retained local action feedback. Affected authored rows are
`module.workbook.visual.timeline_capture_actions`,
`module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`,
and `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`.
The grouped-grid fixture explicitly closes the shortcut-opened History to preserve
its existing framing. Existing masks, zoom, viewports, tolerances, and scroll
normalization remain unchanged. Visual inspection found new summary wrapping
clipped at 200% zoom; make its trigger and compact summary one line with an
ellipsis, retaining full status and recovery content. Golden update and two
ordinary validation runs remain pending; no visual acceptance is claimed yet.

TC-04 review follow-up: first `make browser-e2e-visual-update` passed 12/12
at `.cartulary/test-results/20260911T182556Z-p84306`, reconciling all 217
goldens. Review sheets in that root's `visual-review/` cover all 18 changed
existing inspector/view-bar images; the four new images were inspected directly.
The accepted-result image exposed a popup positioned partly above the viewport;
corrected it to a bounded fixed top-bar recovery region. Successful supersession
now closes its obsolete authoring form. This first promotion was not accepted as
final visual evidence; a fresh ordinary run and update must repair those frames.

The old Inspector toolbar control is open-only. Corrected new browser fixture
closure to use the existing Close inspector button and explicitly assert it stays
closed after replay. The grouped-grid fixture uses that same correct close
control, preserving its original 1440-pixel grid framing. This strengthens
closure evidence; the earlier replay run alone did not prove panel closure.

Expanded `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.capture_action_exact_recovery`
PASS 11/11 `.cartulary/test-results/20260911T183220Z-p74049`.
`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.timeline_capture_actions`
PASS 11/11 `.cartulary/test-results/20260911T183220Z-p74060`, including
recovery control bounds and visible keyboard focus at 1280, 768, and 390 pixels.
`make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_row_mutation_coordinator_1a7e2c9b44`
PASS 2/2 `.cartulary/test-results/20260911T182558Z-p85934`.
Candidate adapter coverage now checks independent 100-row cursor requests,
foreign incident and malformed rows, oversized pages, missing/repeated cursors,
authorization rejection and cancellation. First new adapter run failed because
its Response fixture omitted JSON Content-Type (`20260911T182855Z-p57680`);
fixed the fixture. `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_capture_transport`
PASS 2/2 `.cartulary/test-results/20260911T183009Z-p63998`.
`make generate` PASS `.cartulary/test-results/20260911T182827Z-p50211`.

Additional verification owner: `web.collaboration`; its task guide passes. The
existing duplicate-transaction test now asserts Timeline version observation
precedes own-transaction suppression, retaining all prior assertions.

Ordinary visual follow-up `.cartulary/test-results/20260911T183220Z-p74263`
failed 10/12 with four image comparisons only. Reconciliation has 217/217 active
goldens, zero missing, orphan or ambiguous mappings. Reviewed actuals confirm
correct popup placement and original grouped-grid framing. Two 58-pixel changes
were the duplicate-label rows swapping order because their default activity sort
keys were equal; assign distinct deterministic activity timestamps in the authored
fixture before navigation. This corrects fixture determinism without masking data,
changing tolerance, or modifying product sorting. The narrow confirmation image
remains stable. `make test-slice OWNER=web.collaboration ROWS=web.collaboration.regression.workbookcollaborationcoordinator_suite_514be174a8`
PASS 2/2 `.cartulary/test-results/20260911T183447Z-p74877`.

Late authorization rejection now applies lifecycle effects only to its dispatch
authority generation. A response from an expired session cannot conceal a newly
restored same-account session; receipts still remain recoverable across that
boundary. The retained-owner test covers this fence.

Final candidate update `make browser-e2e-visual-update` PASS 12/12
`.cartulary/test-results/20260911T183736Z-p85709`, with all 217 active goldens
and all 26 registered fixtures reconciled, no missing/orphan/ambiguous entries.
Reviewed the four new frames directly after promotion and the existing changed
frames through the retained review sheets; typography, focus, clipping, overflow,
and responsive states now follow the intended behavior. The grouped-grid golden
is byte-identical to HEAD after correcting fixture closure. No tolerances, browser
pins, masks, viewport/zoom definitions, or existing screenshot scopes changed.

Affected registered fixture identities: `visual.fixture.base_inspector`, `visual.fixture.destructive_actions`, `visual.fixture.inspector_compact_actions`, `visual.fixture.inspector_narrow_technical_details`, `visual.fixture.saved_view_query_controls_and_grouped_result`.
The new Timeline frames are active nonregistry captures mapped to their authored
row and exact capture intents. Changed golden filenames:

- `timeline-supersession-accepted-linux.png`
- `timeline-supersession-authoring-linux.png`
- `timeline-supersession-review-linux.png`
- `timeline-supersession-review-narrow-linux.png`
- `workbook-inspector-compact-actions-linux.png`
- `workbook-inspector-destructive-confirmation-linux.png`
- `workbook-inspector-history-linux.png`
- `workbook-inspector-narrow-technical-details-linux.png`
- `workbook-inspector-public-error-linux.png`
- `workbook-inspector-rollback-preview-linux.png`
- `workbook-query-saved-view-query-controls-linux.png`
- `workbook-view-bar-filter-editing-overflow-linux.png`
- `workbook-view-bar-long-columns-linux.png`
- `workbook-view-bar-maximum-pressure-base-linux.png`
- `workbook-view-bar-maximum-pressure-compact-linux.png`
- `workbook-view-bar-maximum-pressure-narrow-linux.png`
- `workbook-view-bar-ordered-maximum-sort-linux.png`
- `workbook-view-bar-saved-view-actions-linux.png`
- `workbook-view-bar-saved-view-clean-linux.png`
- `workbook-view-bar-saved-view-modified-linux.png`
- `workbook-view-bar-text-spacing-linux.png`
- `workbook-view-bar-zoom-200-linux.png`

Two fresh ordinary visual runs are now validating the promoted manifest.

Latest focused verification: typecheck PASS 2/2 `20260911T183739Z-p87454`;
lint-biome PASS 2/2 `20260911T183740Z-p91166`; import-boundary PASS 2/2
`20260911T183856Z-p51597`; retained-owner/transport slice PASS 3/3
`20260911T183633Z-p80417`; expanded accessibility PASS 11/11
`20260911T183737Z-p85840`. Each root is under `.cartulary/test-results/`.

`make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.capture_action_exact_recovery,module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89,module.timeline.store.a_material_edit_demotes_reviewed_enriched_legal_29bb8697de,module.timeline.store.supersede_with_replacement_rejects_illegal_targe_e4d81b8a06`
PASS 12/12 `.cartulary/test-results/20260911T184230Z-p58102`.
The browser report retains the `timeline-capture-exact-replay` JSON attachment
with original/replayed requests and receipts, complete history before/after,
and replacement link history-unit counts. It asserts the expected acting account.
The routed Timeline store test independently checks actual active supersedes-link
counts and history/revision/change-set counters before and after replay. Both
browser supersession variants also reverse through History rollback and clear
the replacement projection. Non-material preservation and material-edit demotion
remain server-owned and pass the existing store test unchanged.

First fresh ordinary `make browser-e2e-visual` PASS 12/12
`.cartulary/test-results/20260911T184229Z-p57902`. A concurrent second run
failed 10/12 at `.cartulary/test-results/20260911T184229Z-p57905` before a
screenshot in the existing entity-linking mention-dismissal fixture. Its trace
shows successful mention resolve/dismiss responses and a cancelled followed by
accepted Timeline query; the local dismissed chip was absent. No Timeline
capture-state action is admitted in that fixture. All completed comparisons,
including the four new frames, passed. Its one unobserved capture is retained;
no golden, assertion, timeout, or routing is relaxed. A standalone ordinary
repeat on unchanged source is required for the second successful visual run.
This is fresh observed evidence, not an inherited waived failure.

## TC-04 exit

Standalone ordinary `make browser-e2e-visual` PASS 12/12
`.cartulary/test-results/20260911T184826Z-p57226` on unchanged source. Together
with `20260911T184229Z-p57902`, this supplies two fresh successful ordinary
visual runs against the promoted manifest. Both reconcile all 217 active goldens.
The concurrent mention-fixture failure is documented above and passed unchanged
on this standalone run; it is an observed timing limitation, not a relaxed gate.

Exit criteria met: canonical role/state and authoring tests, preparation/version
and selection races, uncertain replay and late settlement, operation-specific
receipts, current authority/session fences, reconciliation, visible replacement,
filter exit, actual service replay counters/attribution, both History rollbacks,
keyboard/responsive evidence, and reviewed visuals pass. Changes are limited to
Timeline actions and their necessary common integration, focused tests/selectors,
authored routing and Make-generated outputs. Public wire/storage compatibility
is unchanged; no backend owner mismatch or scope expansion was necessary.

Next: TC-05 finalization and coordinated full verification. Retained-run
maintenance will be skipped with RESULTS_DIR unset: no exact-source successful
full warm check exists for this final implementation yet.

## TC-05 log and exit

`make explain-target TARGET=check DETAIL=summary` confirms the coordinated final
projection: 839 units and 1227 rows. This includes backend/frontend unit, service,
type, import, lint and policy verification; service browser, accessibility and
visual evidence are separately routed above. Finalize first, then validate
generated policy/drift and run the full check without prior-failure waivers.

`make agent-finalize` PASS 1/1
`.cartulary/test-results/20260911T185235Z-p92589`; its
`unit-artifacts/finalize-summary.json` records RESULTS_DIR null, retained-run
selection/performance/run checks skipped, zero updated files, and passing
current-source schema/catalog/generation checks. Retained-run maintenance was
skipped because RESULTS_DIR was unset, as required.

Pre-full verification: `make generate-drift` PASS 4/4
`20260911T185314Z-p96152`; `make generated-artifact-policy-check` PASS 3/3
`20260911T185314Z-p96169`; `make json-shape-check` PASS 3/3
`20260911T185314Z-p96177`; `make test-catalog-check` exits 0. All roots are
under `.cartulary/test-results/`. Finalize changed no source or generated inputs.

Final frontend prechecks: `make frontend-typecheck` PASS 2/2
`20260911T185314Z-p96443`; `make frontend-import-boundary-check` PASS 2/2
`20260911T185314Z-p96459`; `make lint-biome` PASS 2/2
`20260911T185314Z-p96483`.

### Changed source and verification paths

Goldens are listed in the visual refresh record above. The current non-PNG
change inventory is:

- `apps/web/e2e/support/workbook/timelineCaptureActions.ts`
- `apps/web/e2e/timeline-workbook.spec.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx`
- `apps/web/src/testing/timelineCaptureActionTestSupport.ts`
- `apps/web/src/testing/timelineWorkbookTestSupport.ts`
- `apps/web/src/workbook/WorkbookShell.actionSequencing.test.tsx`
- `apps/web/src/workbook/WorkbookShell.support.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts`
- `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts`
- `apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts`
- `apps/web/src/workbook/inspector/inspectorCapabilityResolver.test.ts`
- `apps/web/src/workbook/inspector/inspectorCapabilityResolver.ts`
- `apps/web/src/workbook/ports/WorkbookTimelineActionRuntimePort.ts`
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`
- `apps/web/src/workbook/timeline/actions/TimelineCandidatePort.ts`
- `apps/web/src/workbook/timeline/actions/TimelineCaptureRecovery.tsx`
- `apps/web/src/workbook/timeline/actions/TimelineSupersessionEditor.test.tsx`
- `apps/web/src/workbook/timeline/actions/TimelineSupersessionEditor.tsx`
- `apps/web/src/workbook/timeline/actions/WorkbookTimelineCaptureActionOwner.test.ts`
- `apps/web/src/workbook/timeline/actions/WorkbookTimelineCaptureActionOwner.ts`
- `apps/web/src/workbook/timeline/actions/reconcileTimelineCaptureReceipt.ts`
- `apps/web/src/workbook/timeline/actions/timelineCaptureActionModel.ts`
- `apps/web/src/workbook/timeline/actions/timelineCaptureOwnerFor.ts`
- `apps/web/src/workbook/timeline/actions/useTimelineCandidates.ts`
- `apps/web/src/workbook/timeline/actions/useTimelineCaptureActions.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineCandidateReader.ts`
- `apps/web/src/workbook/timeline/adapters/createTimelineRecordActionAdapter.ts`
- `apps/web/src/workbook/timeline/adapters/timelineCaptureProtocol.ts`
- `apps/web/src/workbook/timeline/adapters/timelineCaptureTransport.test.ts`
- `apps/web/src/workbook/timeline/components/TimelineRowActions.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx`
- `apps/web/src/workbook/timeline/composition/useTimelineMutationComposition.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineCommittedRows.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineMutationCommands.ts`
- `apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.ts`
- `apps/web/src/workbook/timeline/ports/TimelineRecordActionPort.ts`
- `apps/web/src/workbook/timeline/presentation/TimelineWorkbookInspectorRegion.tsx`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `apps/web/src/workbook/timeline/timelineCaptureActions.characterization.test.ts`
- `docs/handoffs/workbook-timeline-capture-actions-refactor-handoff.md`
- `packages/ui-contracts/src/index.ts`
- `packages/ui-contracts/src/workbookInteractionSelectors.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.timeline.json`
- `tools/test_families/module.workbook.json`
- `tools/test_families/web.workbook.json`

### Compatibility, rollback, and limits

Existing public routes, request/response contracts, stored formats, material-change
classification, Task/Decision lifecycle, history ownership, and artifact publication
are unchanged. The two Timeline actions use their existing generated operations
and authenticated CSRF transport. UI selectors are additive. The legacy raw-ID
selector export remains available for compatibility, but no raw-ID control or
independent row-menu mutation path remains. Existing pure adapter construction
in the foundation is compatible; only the retained Timeline controller admits
and dispatches capture actions.

Rollback consists of reverting this implementation's authored source, routing,
generated derivatives, reviewed goldens and handoff changes. Do not delete or
rewrite Timeline records, replacement links, history, revisions, persisted action
receipts, or analyst data. Already committed supersession is corrected through
the existing History rollback, including the coupled replacement link.

Recovery is deliberately limited to the existing incident/account browser-runtime
lifetime; there is no browser persistence or new server endpoint. A full browser
reload retires client recovery memory. Candidate and preparation/reconciliation
reads are bounded observations over existing cursor queries; unavailable or stale
materialization remains an explicit refresh/review requirement. Browser evidence
is service-backed implementation support, not a new conformance claim. One
concurrent visual run exposed an intermittent existing mention-projection fixture
failure; both successful ordinary visual runs, the trace and the unchanged
standalone repeat are recorded above. No assertion, timeout, tolerance, owner
requirement, or verification row was weakened.

`make browser-e2e-a11y` PASS 14/14
`.cartulary/test-results/20260911T185314Z-p96716`, including existing inspector,
Task/Decision, and the new Timeline keyboard/recovery coverage. All required
pre-full checks pass. Coordinated `make check` now runs on this source without
source edits or relaxed assertions. The index remains empty and branch/HEAD are
still `main` / `c1b0bd637b21ebb91e26a50777da9aa1de155151`.

First coordinated `make check` failed 838/839 units at
`.cartulary/test-results/20260911T185602Z-p37176`. The only failed row was
`web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`: 17
new TypeScript paths were absent from the authored source-ownership inventory.
This is related to the implementation; no failure is waived. The other 838 units
passed. Updated `tools/frontend_source_ownership.json` using existing physical
frontend ownership: the test helper belongs to `web.testing`, and the 16 workbook
port/Timeline files belong to `web.workbook`. Verification is owned by
`web.architecture`; its task guide passed. This mapping does not change Timeline
domain ownership or any runtime behavior. Regenerate, pass the narrow ownership
row, rerun finalization without RESULTS_DIR, and repeat the full check.

Ownership repair validation: `make generate` PASS
`.cartulary/test-results/20260911T190336Z-p86258`;
`make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`
PASS 2/2 `.cartulary/test-results/20260911T190406Z-p89400`.
The repair adds exactly 17 authored mappings; it changes no product source,
selectors, visual fixture, assertion, or public contract. Browser/a11y/visual
evidence remains applicable to the unchanged product implementation.

Second `make agent-finalize` PASS 1/1
`.cartulary/test-results/20260911T190431Z-p90080`, again with RESULTS_DIR unset,
retained-run maintenance skipped, and zero generated updates. The ownership
repair is now included in the coordinated repeated `make check`.

## Final completion record

Coordinated `make check` PASS 839/839 units and 1227/1227 rows at
`.cartulary/test-results/20260911T190514Z-p93683` (271583 ms).
`run-summary.json`, `target-summaries/check.json`, and owner-routed row evidence
retain the complete result. No failed or waived row remains in that run.
The source-ownership omission was corrected before this successful repeated
check. Product source is unchanged since the successful service/browser, full
accessibility, and two ordinary visual runs documented above.

Final `make lint-markdown` before the tracker completion edit PASS
`.cartulary/test-results/20260911T190622Z-p76865`; `git diff --check` passes.
The final pre-completion scope audit shows 77 changed/new files, including 22
reviewed goldens, empty index, `main`, and unchanged HEAD
`c1b0bd637b21ebb91e26a50777da9aa1de155151`. No backend production, contract,
SQL, dependency/lockfile, adopted spec/design, digest, or prior handoff changed.
No commits, resets, pushes, or deployments were performed.

TC-01 through TC-05 are complete. Compatibility and rollback are recorded above;
retained browser memory and the observed concurrent mention-fixture timing failure
are the genuine limitations. The required Markdown, diff, and scope/status
checks are rerun after this final tracker edit. Next action: none for this seam.
This completion does not authorize another seam.

Post-tracker audit: `make lint-markdown` PASS
`.cartulary/test-results/20260911T191103Z-p84428`; `git diff --check` PASS.
The final scope/status audit passes: all 77 changed/new paths match the handoff
inventory (55 non-PNG paths and 22 reviewed goldens), all five tracker rows are
COMPLETE, branch/HEAD are unchanged, the index is empty, and protected scopes
and all other Markdown files are unchanged. After recording these audit results,
Markdown lint, diff whitespace, and scope/status are checked once more; no
product or verification input changes follow the successful full check.
