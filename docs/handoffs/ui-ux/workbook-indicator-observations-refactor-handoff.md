# Workbook Indicator observations refactor

## Baseline, authority, and scope

Implementation entry: clean `main` at
`1477ba8b28e2afed45c91bd271b689c23915be0c`; tracked and untracked status empty.
Root AGENTS.md applies; its localized digest read order and current source were
inspected during planning and the unchanged baseline revalidated at execution.
Core 01 REQ-01-615–617/652/654, Core 02 REQ-02-075–080/260/264, Core 03 §9.1
and REQ-03-282/301/306, and Core 04 REQ-04-150 govern this seam. Current session,
authorization, transaction recovery, history, and continuity owners also apply.
Design supplies bounded presentation direction; domain supplies vocabulary.
The digest, NLSpec research, and completed handoffs are advisory evidence.
No owner contradiction identified in the inspected paths.

Allowed paths: observation-related source and narrow integration beneath
`apps/web/src/workbook`, focused support beneath `apps/web/src/testing`, browser
tests/support/affected visuals beneath `apps/web/e2e`, authored selectors beneath
`packages/ui-contracts/src`, focused Indicator service tests, authored ownership
and verification inputs beneath `tools` and `contracts/verification`, their
Make-generated derivatives, and this handoff. Backend production, public
routes/schema, adopted specifications, SQL/migrations, dependencies, digest
contents, and unrelated workflows are excluded. No commit, reset, push,
deployment, or analyst-data cleanup is authorized.

User decisions: unsaved source edits must save before explicit reselection;
candidate selection uses an independent type-filtered paged Indicator list.
Canonical create-from-observation and new Notes/Evidence entry points are out
of scope. This slice does not complete the broader REQ-03-135 feature.

## Ordered tracker

| Workstream | Dependency | Status | Binary exit |
| --- | --- | --- | --- |
| IOB-01 | None | DONE | Baseline/ledger complete; narrow baseline green; characterization demonstrates corrections |
| IOB-02 | IOB-01 | DONE | Source selection, isolated targets, legal transition UI, keyboard and omission tests pass |
| IOB-03 | IOB-02 | DONE | Admission, exact retained attempts, receipts, replay and authority/lifetime tests pass |
| IOB-04 | IOB-03 | DONE | Reconciliation, provenance, paging, real service/browser and accessibility evidence pass |
| IOB-05 | IOB-04 | DONE | Finalizer, full check, completed handoff and final byte/scope audits pass |

Only the current row may be IN_PROGRESS. Its log closes DONE or BLOCKED before
advancement. A blocked dependency prohibits subsequent workstreams.

## Owner/gap ledger

All entries preserve public interfaces and stored data. The risk column records
the risks identified before remediation. Every binary gap exit now has passing
execution evidence in IOB-02 through IOB-05; no unresolved in-scope gap remains.
The runtime-lifetime and browser-engine limits are stated below.

| Gap | Owner | Remediation and affected areas | Rationale and long-term benefit | Compatibility | Unresolved risk | Binary completion criterion |
| --- | --- | --- | --- | --- | --- | --- |
| Mandatory offsets without actual source | Core 02 078/264; Core 03 136/306 | Committed source selection and UTF-16/UTF-8 mapping in observation UI/model and Timeline composition | Bind capture to the occurrence the analyst reviewed | Existing span request only | Unicode, line endings, stale drafts | Exact spans and source-edit races pass; raw source unchanged |
| Editability-derived field inventory | Core 02 264; source-text capability | Active-contract text discovery over committed cells | One capability rule without label or editor drift | Existing Timeline entry point only | Read-only text and ambiguous view ownership | Capability parity and unavailable-value tests pass |
| Shared raw UUID target | Core 02 264; Core 04 150 | Per-operation paged target picker using existing Indicator query | Meaningful authorized selection without cross-observation leakage | No search API added | Stale target, visible-page dependence | Off-page candidates and isolated/stale targets pass |
| Disposable admission and regenerated identity | Core 01 652; Core 03 301 | Observation retained owner, immutable transport attempts and shell recovery | One recoverable logical operation independent of component lifetime | No persistence, FIFO admission, locks or automatic retry | Duplicate clicks, lost response, late settlement | Exact replay bytes/ID and runtime fencing tests pass |
| Incomplete receipt acceptance | Core 01 652; Core 02 264; Core 04 150 | Operation-specific observation/receipt validation in adapter | Separate known commit from uncertain transport | Same closed response | Malformed success or misleading version/tuple | Complete receipt matrix passes; malformed success remains uncertain |
| Visible-surface-only refresh | Core 02 260; Core 03 306 | Stable-ID multi-record reconciliation and neutral version integration | Keep source and previous/new Indicator projections monotonic | Same HTTP/socket channels | Failed reads, late or suppressed socket delivery | All affected IDs refreshed; refresh retry sends no mutation |
| Conflated collection/provenance states | Core 01 654; Core 02 076/260/264; design 7/12/14 | Observation paging state and provenance presentation | Preserve authorized data and select the correct recovery action | Opaque cursors and existing History rollback | Order, duplicate delivery, tombstones, focus | Paging/membership/rollback/keyboard/browser matrix passes |
| Native readonly control cannot select by keyboard | Core 01 654; design accessibility | Selection-only textarea with readonly accessibility semantics and rejected input events | Native keyboard ranges without exposing source editing | Source text and capture request unchanged | Browser input fallback | Keyboard selection, typing/deletion/Enter immutability and responsive recovery pass |
| Partial authority initialization clears authorized rows | Core 03 306; Core 04 150 | Neutral record bridge distinguishes initialization from later authority loss | Preserve successful queries while retaining immediate security fencing | No route or owner-policy change | Sequential owner notification timing | Initial query remains; either later authority loss clears; lifecycle browser regressions pass |
| Unsaved edit does not immediately update capture controls | Core 03 306; Core 01 654 | Row-scoped draft notifications observed only by the active observation source | Match visible admission state to the actual committed boundary | No save or flush behavior changed | Draft clearing and focus continuity | Browser requires save/reselection; unchanged stored capture source and target draft |

Advisory classification: ADOPT local feedback, keyboard parity, explicit states
and recovery; ADAPT compact presentation through current tokens; REJECT generic
confirmation ceremony, new product policy, and a parallel design system.

## IOB-01 execution log

Planning inspected the requested Inspector, Timeline/generic composition,
command/transport/runtime and committed-version paths, active view/OpenAPI
inputs, source-text port, Indicator admission/transitions/replay/collections,
projections/rollback, existing tests and owner routing. Backend production
already supplies source-bound capture, closed transitions and replay receipts.
Planning baseline: `make test-slice OWNER=module.indicators` with the two
observation frontend rows and source-span/transition unit rows passed 4/4 at
`.cartulary/test-results/20260911T214809Z-p16296`. The production child-route
`make service-backed-test-slice OWNER=module.indicators` passed 3/3 at
`.cartulary/test-results/20260911T214809Z-p16301`. Exact selectors are retained
in those run artifacts. Both required task guides were rerun at execution.

IOB-01 DONE: `make generate` PASS at
`.cartulary/test-results/20260911T215623Z-p40059`. Execution observation baseline
PASS 4/4 at `.cartulary/test-results/20260911T215645Z-p43602` with the same four
rows. `make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.indicator_observations_characterization` fails as
expected (1/2 units) at `.cartulary/test-results/20260911T215645Z-p43594`:
there is no accessible Saved source text control. No production movement
preceded this characterization. Baseline HEAD and user state remain unchanged.

## IOB-02 execution log

Source/target authoring and observation collection presentation are developed
against narrow read/draft interfaces first. The retained mutation owner and
production composition switch follow in IOB-03; no intermediate UI introduces
an alternative mutation path.

IOB-02 DONE: model and paging rows passed at
`.cartulary/test-results/20260911T220512Z-p53250`; the authoring row initially
failed because test cleanup was missing, then passed after correction at
`.cartulary/test-results/20260911T220543Z-p54658`. The expanded transition
isolation/restore authoring row passed at
`.cartulary/test-results/20260911T220948Z-p60485`. Tests cover exact byte spans,
browser line-ending mapping, surrogate boundaries, read-only text capability,
optional omission, independent paged targets, draft isolation, source-version
invalidation and closed transition presentation. Native text selection and
form controls retain keyboard semantics; real keyboard evidence follows in
IOB-04. Frontend typecheck passed at
`.cartulary/test-results/20260911T220512Z-p53326`, after correcting initial
type errors reported at `.cartulary/test-results/20260911T220320Z-p48297`.

## IOB-03 execution log

Connect authoring to an observation-specific retained owner and immutable
transport. Risks: preparation races, malformed receipts, duplicate delivery,
authority changes and late settlement. Exit requires routed owner/adapter
coverage and a green production characterization.

IOB-03 DONE: observation owner, transport, source-boundary and production
characterization rows passed 5/5 at
`.cartulary/test-results/20260911T222437Z-p82511`. The migrated Indicator
Inspector/public-contract rows passed 3/3 at
`.cartulary/test-results/20260911T222438Z-p82736`. Expanded observation runtime
retention/source-boundary and adjacent lifecycle owner/runtime checks passed
5/5 at `.cartulary/test-results/20260911T222552Z-p88393`.
The old component-owned `createIndicatorWorkflowPort` and its command/runtime
plumbing were removed; its public-contract assertions moved to
`adapters/observationContracts.test.ts`. Requests retain their original bytes,
base version and secure identity; receipt acceptance precedes refresh debt.
Source preparation uses the existing same-record idle/committed boundary and
does not flush drafts. Uncertainty survives Inspector and same-account shell
detachment; account/runtime retirement clears protected state.

Intermediate typecheck failures identified test typing/unused imports and were
corrected. Import-boundary failure at
`.cartulary/test-results/20260911T222552Z-p88476` identified direct protocol
imports and a controller-directory test placement; a narrow protocol adapter
and relocated test corrected both without changing boundary policy.
Final IOB-03 typecheck PASS 2/2 at
`.cartulary/test-results/20260911T222654Z-p93169`; import-boundary PASS 2/2 at
`.cartulary/test-results/20260911T222654Z-p93163`. Latest generation PASS at
`.cartulary/test-results/20260911T222644Z-p90036`.

## IOB-04 execution log

Reconcile all affected records independently of the visible page, integrate
monotonic HTTP/socket evidence, and prove service-backed recovery and ordinary
restore/History rollback. Risks: late reads, projection failure, page continuity,
focus changes and source/old/new Indicator membership.

Narrow reconciliation/socket checks passed 5/5 at
`.cartulary/test-results/20260911T223734Z-p5383`. Presentation, controller
continuity, authoring and source-boundary checks passed 5/5 at
`.cartulary/test-results/20260911T224948Z-p20529`; typecheck passed 2/2 at
`.cartulary/test-results/20260911T224948Z-p20713`.

Service/browser characterization found and corrected three implementation gaps:
explicit label associations were missing; the Timeline feature controller closed
the observation panel on its own committed source-version change; and collection
validation assumed timestamps were UTC-only although the public date-time
contract admits offsets. The controller now preserves the same live observation
subject while selection validates its version. Timestamp comparisons use the
exact instant, preserving submillisecond order and returned provenance strings.
The timestamp/transport rows passed 3/3 at
`.cartulary/test-results/20260911T225624Z-p10517`.

The complete Indicator service slice passed 8/8 at
`.cartulary/test-results/20260911T224948Z-p20579`. Its prior run at
`.cartulary/test-results/20260911T224218Z-p52435` failed only the existing
clean-install test's stale hardcoded migration head (40 versus baseline 41).
The test-only maintenance in `lifecycle_integrity_migration_test.go` now compares
against the canonical embedded catalog maximum; all integrity and rollback
assertions remain. No production or migration source changed.

The exact-recovery browser row passed at
`.cartulary/test-results/20260911T225609Z-p81320` (4.7 seconds). It proves real
commit followed by dropped response and exact replay for creation, reassignment,
dismiss and restore, unchanged service snapshots/history, source/old/new
projections, byte-identical raw source, illegal same-state rejection, intentional
identical-text observation identities, and History rollback. The combined run
remained failed (11/13 units) because native keyboard selection needed
investigation; IOB-04 remained open at that point. Earlier browser failures are retained at
`20260911T223742Z-p6341`, `20260911T224218Z-p52430`, and
`20260911T224948Z-p20554` beneath `.cartulary/test-results`.

Default Biome passed at `.cartulary/test-results/20260911T224254Z-p99811`;
format passed at `.cartulary/test-results/20260911T225558Z-p76938`. Earlier lint
and type failures were corrected without policy changes. A diagnostic Biome
flag override and an incorrectly named source-span test selector were rejected
by the command surface; neither produced verification evidence.

Additional IOB-04 evidence: the native keyboard probe isolated Chromium's
`readonly` textarea behavior (selection remains empty even with unprevented
Shift navigation). The source control now exposes `aria-readonly`, blocks
`beforeinput`, paste, cut and drop, and restores its immutable display on any
fallback change event. Its native caret and selection remain available. The
real accessibility row passed 11/11 units at
`.cartulary/test-results/20260911T230340Z-p13822`, including typing, deletion,
Enter, responsive widths, 200% zoom, names, focus and recovery. Diagnostic probes
were removed; failed keyboard runs remain at `20260911T225747Z-p50600`,
`20260911T225929Z-p82957`, `20260911T230041Z-p48394`, and
`20260911T230154Z-p80474` beneath `.cartulary/test-results`.

Adjacent browser tests at `.cartulary/test-results/20260911T225944Z-p11203`
exposed partial initialization in the neutral Indicator version bridge: one
owner's first authority notification cleared an already authorized query while
the other owner was initializing. A routed characterization failed at
`.cartulary/test-results/20260911T230535Z-p81211`. The bridge now publishes
completed initialization and immediately publishes subsequent authority loss.
The regression and authoring/collection/characterization rows passed 5/5 at
`.cartulary/test-results/20260911T230855Z-p16996`. Intermediate test-helper errors
at `20260911T230618Z-p82347` were corrected; their associated browser run could
not build. Latest typecheck passed at `20260911T230818Z-p9220`; generation passed
at `20260911T230817Z-p8603` beneath `.cartulary/test-results`.

The expanded paging/source-edit browser row passed 11/11 at
`.cartulary/test-results/20260911T231351Z-p29105`. It exercises 103 independent
Indicator candidates, 101 source observations, failed continuation retry with
identical cursors, retained loaded rows and explicit save/reselection. Earlier
runs at `20260911T230855Z-p17008` and `20260911T231126Z-p88235` exposed an invalid
test error envelope and the missing draft notification respectively. The
source-boundary/authoring rows passed 3/3 at `20260911T231350Z-p27765` beneath
`.cartulary/test-results` after adding row-scoped draft subscriptions. Observation
subscription is active only for its open source workflow; unrelated draft
editors do not acquire it.

The combined run at `.cartulary/test-results/20260911T230855Z-p17008` retained
passing evidence for exact observation recovery, all three lifecycle browser
rows, and all three observation/lifecycle/Timeline accessibility rows; only its
then-new paging row failed. Architecture passed 12/12 at
`.cartulary/test-results/20260911T231127Z-p89458`; the previous heading-name
selector-policy failure was corrected by scoped headings and explicit
accessible-name assertions. Import boundaries passed 2/2 at
`.cartulary/test-results/20260911T231000Z-p84969`. Collaboration passed 8/8 at
`.cartulary/test-results/20260911T231518Z-p97584`; shared UI passed 10/10 at
`.cartulary/test-results/20260911T231518Z-p97605`.

Final IOB-04 observation and save-status browser/accessibility evidence passed
15/15 at `.cartulary/test-results/20260911T231518Z-p97579` against the narrowed
active-workflow draft subscription. Timeline capture exact recovery passed
11/11 at `.cartulary/test-results/20260911T231537Z-p29681`. Timeline/generic
History paging, newer-row replay and failed-refresh recovery passed 13/13 at
`.cartulary/test-results/20260911T231617Z-p90857`. A Network Analysis regression
attempt at `.cartulary/test-results/20260911T231538Z-p30451` was rejected before
browser execution because formatting changed build inputs after its source
snapshot; it supplies no browser evidence. The stable-input rerun follows.

IOB-04 DONE: the stable-input Network Analysis indicator-link browser and
accessibility slice passed 13/13 at
`.cartulary/test-results/20260911T231803Z-p25660`. Latest frontend typecheck
passed 2/2 at `.cartulary/test-results/20260911T231804Z-p25921`. Markdown lint
passed at `.cartulary/test-results/20260911T231822Z-p54806`. All required
IOB-04 service/browser/accessibility exits now have successful evidence; the
first ordinary visual run also passed. No owner contradiction or required
out-of-scope production repair was found.

## Completed visual review record

Accepted trigger: the requested observation authoring interface and new source
selection fixture. Owner row:
`module.workbook.visual.indicator_observations_authoring`; fixture:
`visual.fixture.indicator_observations_authoring`. New golden filenames:
`indicator-observation-authoring-linux.png` and
`indicator-observation-authoring-narrow-linux.png` beneath
`apps/web/e2e/workbook.visual.spec.ts-snapshots`.
No existing viewport, zoom, masking, screenshot scope, renderer or normalization
policy changes. New captures use compact Timeline density at 1280×720 and
768×640, with the observation panel anchored at the Inspector scrollport start.

Ordinary full visual evidence and reconciliation:
`.cartulary/test-results/20260911T225624Z-p10775`. The 38 passing workbook scenarios
and separate passing Network Analysis visual preserved existing comparisons;
failures were the two missing new goldens and fixture-inventory assertions,
which now include the new fixture. Ordinary focused capture evidence:
`.cartulary/test-results/20260911T230400Z-p46784`; both retained review images
were inspected before update. Source, preview, labels, focus and responsive
scrolling were readable with no unexpected overflow or clipping. The remaining
controls are available by Inspector scrolling, as verified by accessibility.
`make browser-e2e-visual-update` passed 12/12 at
`.cartulary/test-results/20260911T230856Z-p17232`. Only the two new PNGs and the
Make-generated golden manifest changed. Both promoted images were inspected
again and matched the reviewed capture intent. The first ordinary visual run passed 12/12 at
`.cartulary/test-results/20260911T231434Z-p63903`, with successful full visual
reconciliation. The second ordinary visual run passed 12/12 at
`.cartulary/test-results/20260911T231846Z-p58999`, also with successful full
reconciliation. No existing golden was modified or removed.

## IOB-05 execution log

IOB-05 DONE: finalizer maintenance, the full unwaived check, both required
ordinary visual runs and the source-scope review passed. `RESULTS_DIR` was unset
because no successful exact-source full warm check qualified for retained-run
maintenance at entry. Broad regressions and routing/generated drift are covered
by the successful final check. Markdown, whitespace and scope checks are repeated
after marking this tracker DONE, against the final handoff bytes.

`env -u RESULTS_DIR make agent-finalize` passed 1/1 at
`.cartulary/test-results/20260911T231956Z-p94687`; its
`unit-artifacts/finalize-summary.json` records zero updated generated files,
no mutation rollback, and skipped retained-run selection, performance evidence
and retained run checks because `RESULTS_DIR` was not provided. Specifically,
`canonical_evidence_validation` and `scheduler_drift_validation` report
`results-dir-not-provided`. This is not
claimed as retained-run validation. The first subsequent `make check` completed 857/858 at
`.cartulary/test-results/20260911T232032Z-p98739`. Its only failure was
`harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131`:
`listTargetObservations` was exported despite being used only within its own
browser support module. It is now private; no assertion, routing or policy was
weakened. The focused policy check passed 2/2 at
`.cartulary/test-results/20260911T232836Z-p57125`. The second
`env -u RESULTS_DIR make agent-finalize` passed 1/1 at
`.cartulary/test-results/20260911T232906Z-p57784`, again with zero generated
updates and the same explicitly skipped retained-run maintenance. The subsequent
`env -u RESULTS_DIR make check` passed **858/858** units in 457147 ms at
`.cartulary/test-results/20260911T233042Z-p61552`, with no failed, skipped or
cancelled units. No assertions were weakened and no historical waiver was used.
Latest explicit import-boundary check passed 2/2 at
`.cartulary/test-results/20260911T232757Z-p56280`.

Final source review retained branch `main` and HEAD
`1477ba8b28e2afed45c91bd271b689c23915be0c`. The working tree contains 41 tracked
changes and 36 untracked files, all included in the inventory below and within
the allowed paths. The baseline had no user changes. `git diff --check` passed;
there are no scope violations, production backend changes, specification edits,
dependency changes or unrelated workflow additions.

After the tracker was marked DONE, `make lint-markdown` passed at
`.cartulary/test-results/20260911T233909Z-p6349`; `git diff --check` also passed.
The final status/scope audit found that this inventory exactly matches the
working tree: 41 tracked changes, 36 untracked files and zero scope violations.
Markdown lint, whitespace and scope auditing are repeated after this evidence
entry so completion is checked against the final handoff bytes.

## Changed-file inventory

The following working-tree inventory includes new files and deliberate legacy
removals. It is review evidence; no executable check reads this handoff.

```text
 M apps/web/e2e/workbook.a11y.spec.ts
 M apps/web/e2e/workbook.visual.spec.ts
 M apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx
 M apps/web/src/workbook/WorkbookShell.tsx
 M apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts
 M apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts
 M apps/web/src/workbook/features/generic/GenericWorkbookInspector.tsx
 M apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx
 M apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.test.tsx
 M apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.tsx
 M apps/web/src/workbook/features/indicators/indicatorLifecycle.characterization.test.tsx
 M apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts
 M apps/web/src/workbook/hooks/useWorkbookSurfaceQueries.ts
 D apps/web/src/workbook/mutations/createIndicatorWorkflowPort.test.ts
 D apps/web/src/workbook/mutations/createIndicatorWorkflowPort.ts
 M apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts
 M apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts
 M apps/web/src/workbook/query/useGenericSurfaceQuery.ts
 M apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts
 M apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx
 M apps/web/src/workbook/timeline/components/TimelineWorkbook.tsx
 M apps/web/src/workbook/timeline/composition/useTimelineInspectorWorkflowComposition.ts
 M apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts
 M apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.ts
 M apps/web/src/workbook/timeline/hooks/useTimelineInspectorFeatureController.ts
 M apps/web/src/workbook/timeline/models/timelineFieldRegistry.ts
 M apps/web/src/workbook/timeline/models/timelineModelBoundaries.test.ts
 M apps/web/src/workbook/timeline/models/timelineWorkbookSurfaceRuntime.ts
 M apps/web/src/workbook/timeline/presentation/TimelineWorkbookInspectorRegion.tsx
 M apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx
 M apps/web/src/workbook/timeline/useTimelineInspectorFeatureController.test.tsx
 M internal/modules/indicators/lifecycle_integrity_migration_test.go
 M packages/ui-contracts/src/index.ts
 M tools/browser_e2e_batch_manifest.json
 M tools/execution_topology_render_index.json
 M tools/frontend_source_ownership.json
 M tools/frontend_visual_fixture_registry.json
 M tools/frontend_visual_golden_manifest.json
 M tools/test_families/module.indicators.json
 M tools/test_families/module.workbook.json
 M tools/test_families/web.workbook.json
?? apps/web/e2e/indicator-observations.spec.ts
?? apps/web/e2e/support/workbook/indicatorObservations.ts
?? apps/web/e2e/workbook.visual.spec.ts-snapshots/indicator-observation-authoring-linux.png
?? apps/web/e2e/workbook.visual.spec.ts-snapshots/indicator-observation-authoring-narrow-linux.png
?? apps/web/src/testing/observationTestSupport.ts
?? apps/web/src/workbook/adapters/createObservationReader.ts
?? apps/web/src/workbook/adapters/createObservationTransport.test.ts
?? apps/web/src/workbook/adapters/createObservationTransport.ts
?? apps/web/src/workbook/adapters/observationContracts.test.ts
?? apps/web/src/workbook/adapters/observationProtocol.ts
?? apps/web/src/workbook/features/indicators/ObservationCaptureEditor.tsx
?? apps/web/src/workbook/features/indicators/ObservationCollection.test.ts
?? apps/web/src/workbook/features/indicators/ObservationCollection.ts
?? apps/web/src/workbook/features/indicators/ObservationContext.ts
?? apps/web/src/workbook/features/indicators/ObservationDetails.tsx
?? apps/web/src/workbook/features/indicators/ObservationDraftStore.ts
?? apps/web/src/workbook/features/indicators/ObservationOperationStatus.tsx
?? apps/web/src/workbook/features/indicators/ObservationPagingFeedback.tsx
?? apps/web/src/workbook/features/indicators/ObservationTargetPicker.tsx
?? apps/web/src/workbook/features/indicators/WorkbookObservationOwner.test.ts
?? apps/web/src/workbook/features/indicators/WorkbookObservationOwner.ts
?? apps/web/src/workbook/features/indicators/WorkbookObservationRecovery.tsx
?? apps/web/src/workbook/features/indicators/indicatorObservations.characterization.test.tsx
?? apps/web/src/workbook/features/indicators/observationAuthoring.test.tsx
?? apps/web/src/workbook/features/indicators/observationModel.test.ts
?? apps/web/src/workbook/features/indicators/observationModel.ts
?? apps/web/src/workbook/features/indicators/observationOperation.ts
?? apps/web/src/workbook/features/indicators/observationPresentation.test.tsx
?? apps/web/src/workbook/features/indicators/observationReconciliation.test.tsx
?? apps/web/src/workbook/features/indicators/observationStyles.ts
?? apps/web/src/workbook/features/indicators/reconcileObservationReceipt.ts
?? apps/web/src/workbook/features/indicators/useObservationTargetNames.ts
?? apps/web/src/workbook/query/WorkbookCommittedRecordPort.ts
?? apps/web/src/workbook/timeline/hooks/useTimelineObservationSource.ts
?? apps/web/src/workbook/timeline/useTimelineObservationSource.test.tsx
?? docs/handoffs/workbook-indicator-observations-refactor-handoff.md
```

## Compatibility, limitations and rollback

No public API, request/response schema, production backend, SQL, stored-data
format, dependency or migration changed. Existing child-resource clients remain
compatible. The sole Go change maintains an existing service test's clean-install
head assertion against its canonical embedded migration catalog.

This slice implements existing-Indicator capture, resolution/reassignment,
dismissal/restoration and provenance in the existing Timeline Relationships
workflow; the Indicator pivot remains read-only. Canonical creation from an
observation, Notes/Evidence entry points, entity mentions, Indicator lifecycle
changes and Network Analysis linking changes are deferred. This does not complete
the broader REQ-03-135 enrichment feature. Existing adjacent workflows have
regression evidence above.

Admitted attempts are retained only for the existing permitted workbook runtime
lifetime. A full browser/process reload has no new persistence mechanism; account
replacement and incident/runtime retirement discard protected client state.
Server observations, transaction receipts and history remain stored. Source
changes require explicit reselection, even when the text returns to an earlier
value. Ordinary observation Restore returns to unresolved; exact historical
rollback remains an existing History operation. Browser evidence uses the
repository's pinned Chromium renderer and does not claim coverage of other
browser engines.

To roll back the implementation, apply a reviewed reverse patch to the coordinated
file inventory, including restoring the removed legacy adapter/port and reversing
its composition changes. Regenerate derivatives through Make and verify the
restored source with the relevant owner checks. Recover outstanding client
attempts before replacing the running frontend where possible; no code rollback
may delete server observations, transaction receipts, history, revisions or
analyst data. No database rollback, data cleanup or API migration is part of this
procedure. No commit, reset, push or deployment was performed.

Next action: none unless separately authorized.
