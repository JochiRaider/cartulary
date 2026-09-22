# Timeline pending-capture authoring remediation

Completed implementation and owner-selected verification. The working tree is the review artifact.

## Scope and authority

Started on clean `main` at `0ecfe278b70bc33c0478a10671ceddfbf2e7c899`.
No user edits existed. No commit, push or deployment is authorized.
The completed native Create activation remains in place.

Source owner: `web.workbook`; verification owners: `web.workbook` and
`module.timeline`, plus `module.workbook` for the narrow shared queue guard and
`module.evidence` for the existing screenshot/capture integration. Core 03 REQ-03-099/100/111/298/300 and Core 01
REQ-01-057/069/070 own capture, retention and idempotency; REQ-01-288 owns
active view-contract discovery/admission. Core 03
REQ-03-299/100 own security lifetimes. Design §§8 and 14 guide native editing
and accessibility. The dated digest is advisory navigation only.

## Characterization

Confirmed replacement defect: reveal Synopsis, hold first create, fill A,
replace with B without keydown, assert B before release or Create activation.
The connected textarea receives a React value write restoring A before its
target input listeners run. Instrumented failure:
`.cartulary/test-results/20260922T151152Z-p66386`; uninstrumented reproduction:
`.cartulary/test-results/20260922T151450Z-p1427`. The diagnostic accessor was
removed after characterization; no authored-text production logging was added.

Confirmed detached promotion defect: with real retained drafts, accept a create
while detached after newer B authoring. Alias remains `draft-1`, rather than the
accepted record. Failing owner test:
`.cartulary/test-results/20260922T151439Z-p883`.

Structural weaknesses: duplicate attachment-local capture maps, mount-local
allocation, synthetic stale-key reconstruction, value-equality acknowledgement
fallbacks and whole-row successor comparisons. Identity collisions and unrelated
field overwrite were hypotheses, not baseline reproduced defects. Regression
coverage now checks monotonic allocation, unknown-key refusal and field-owned
patches without treating structural evidence as a prior observed failure.

Baseline owner guides were refreshed for both owners. Existing blank-create and
committed typing acknowledgement measurements passed (16/16 graph units):
`.cartulary/test-results/20260922T150810Z-p29386`. Predicates and budgets unchanged.

A catalog update initially failed its ASCII title ordering check; corrected
before rerunning. An earlier selected recovery run passed existing cases but
excluded the new test; it is not evidence for detached promotion.

## Contract transitions

The Core 03 edits distinguish raw revision, queued intent, captured immutable
attempt and accepted row. See the transition table in §7. Creation omission does
not erase successor clear intent. Independent editor contexts remain separate.
Only admitted successors become patches; refusal retains raw authoring.

## Advice dispositions

- R008 ADOPT: retain owner-defined recovery and raw input; no generic autosave policy.
- R012 ADAPT: runtime-local lifecycle, mounted DOM and focus; no durable storage.
- R002 ADAPT: native controls within adopted accessibility scope, no imported AAA claim.
- R033 REJECT as authority: offline guidance cannot replace adopted Core owners.
- R034 REJECT: incidental selectors; use semantic production-grid selectors.

The narrow offline recovery query returned three results (error recovery, input
types and labels). R033 rejects importing restrictive input types into raw Timeline
authoring; existing semantic labels remain under R004. Product checks do not consume this document or other Markdown.

## Implemented boundary and workstream exits

| Workstream | Correction and exit evidence |
| --- | --- |
| 1 — Contract and characterization | Core 03 §7 now contains the transition table; REQ-03-099 ends create coalescing at dispatch, and REQ-03-298 covers recordless revisions. Both initial failing cases were run before retained-owner changes. The exact browser replacement remains in the test and passes. |
| 2 — Retained ownership | `WorkbookTimelineMutationOwner` retains `TimelineCaptureLifecycle`; allocation, aliases and draft-to-record movement outlive mounted controls. Detached acceptance and remount tests use the real draft store and driver. Unknown keys cannot synthesize draft rows. |
| 3 — Native authoring | `TimelineScalarEditor` subscribes to retained field values; independent local text state and focus-based prop copying are removed. Current committed rows supply fallback values, including after a patch while the grid retains its original editor row. Browser replacement, native paste, clear, continuous typing, selection and composition checks pass. |
| 4 — Successors and settlement | The existing driver records a create predecessor, patches only owned revisions, handles explicit clears, and halts orphan successors. The existing queue still owns capacity/FIFO/replay. Settlement accepts revision maps only; equal-text matching cannot clear someone else's work. Real-owner tests cover equal text, independent Inspector work, multiple fields, unrelated committed fields, overflow, rejection and discard. |
| 5 — Verification and handoff | Authored catalogs route the added browser and owner tests. Generated topology comes from Make tooling. Final selected runs, measurements, static checks and acceptance assessment are below. |

The first backward transition was a controlled-textarea write between native
`beforeinput` and the target `input` listener. Document capture-phase continuity
cancellation published React state before the scalar editor could read B; that
render restored A, which then became the input callback's value. The isolated
continuity correction passed in `20260922T151604Z-p34232`. Cancellation now aborts
focus work immediately in capture; document bubble retires only the old scheduling
token after authoring publication. A newer token admitted by that input survives.
There are no restoration timers or button-specific text repairs.

Retention is split by lifetime, not by competing text stores:

- `WorkbookLocalDraftStore` remains the only raw value/revision/baseline store.
  Its synchronous batch boundary prevents subscribers seeing half-settled
  acknowledgement/promotion state.
- `TimelineCaptureLifecycle` owns runtime draft allocation and aliases. Mounted
  registries borrow it and own DOM references, selection and composition.
  Aliases without draft, queued-operation or Evidence references retire on
  detachment; runtime retirement clears all capture metadata.
- The mutation owner promotes before invoking optional mounted presentation.
  Composition delays mounted row replacement until its final input; queue
  acceptance continues independently. Remount shows one fresh trailing draft
  and does not automatically activate an old editor.
- Source-owned successor metadata does not add another queue. A small guarded
  `prepareUnsentCreate` operation lets the source normalize only the authorized,
  never-dispatched FIFO head. It preserves the transaction ID and cannot change
  a dispatched or uncertain attempt. Existing patch-version preparation and
  authoring-baseline review remain in effect.

Removed paths: mount-local draft counter, attachment-local capture aliases,
`createDraftRowForKey`, scalar local-value synchronization, submitted-value
settlement arguments/fallbacks, whole-row successor scalar diffs, and the
row-wide `timelineDiscardedReconciliation` model and private callback. Discard
now retires only revisions owned by its operation. Orphan dependents remain
blocked for explicit recovery and cannot dispatch as a second create.

Retained paths: first qualifying input capture; scalar/collection command
admission; secure transaction generation; existing pending queue, retry and
conflict owners; exact uncertain replay; accepted-write/read-refresh separation;
Add row focus-only action; native Create activation; independent Evidence action;
row-menu lifetime and continuity cancellation for newer user intent.

## Selection rubric and compatibility

| Gap | Areas, rationale and long-term benefit | Compatibility, risk and validation |
| --- | --- | --- |
| Ambiguous recordless lifecycle | Specification, tests and handoff: make raw authoring, admitted work, captured attempt and accepted record distinct. Future field types reuse these transitions. | Clarifies existing routes without wire migration. Unresolved ambiguity would permit data loss despite individually passing tests. Exit: transition table plus exact red/green replacement evidence. |
| Attachment-owned identity/promotion | Implementation, private interfaces and source guides: retain one logical identity across detachment and acknowledgement. | No record-ID or selector changes. Incorrect aliases could cross drafts or retain protected data; real detached/remount and retirement tests validate this boundary. |
| Competing scalar values | Implementation and native browser tests: editor text projects retained authoring, while DOM selection/composition stays mounted. | Preserves capture and keyboard behavior. Removes stale prop copying rather than keeping a compatibility mode. Original replacement, paste/clear/Unicode, selection and composition tests are completion criteria. |
| Snapshot-based successors | Implementation and queue/driver tests: only owned field changes become dependent patches; omission cannot erase clear intent. | Same endpoints, payload contracts, queue capacity and FIFO. No general dependency graph. Tests require one create, exact replay, guarded patches, revision settlement and recovery without duplicate creation. |
| Incidental verification | Authored catalogs, generated routing and documentation: route behavior through real owners and the production grid. | No Markdown product inputs. Removal of obsolete model assertions is paired with real-owner discard tests. Existing performance predicates/budgets remain unchanged. |

No dependencies, endpoints, schemas, migrations, durable draft storage, cross-tab
recovery or other workbook producers were added. Shared changes are limited to
batching the existing local draft store and guarding pre-dispatch create
normalization in the existing queue. `packages/grid-adapter` remains the sole
vendor integration boundary; no direct grid-vendor import was added.

Authorization and data lifetime remain existing owner policy: closure/role loss
retain readable local work without granting writes; account-session suspension
conceals protected presentation while retaining same-account work; same-account
recovery resumes existing ownership. Incident retirement/account replacement
clear local stores, queued work and capture metadata before late receipts can
repopulate them. No production diagnostic contains authored text.

## Commands, results and artifacts

All repository commands ran from the repository root through public Make targets,
with command-local `PATH="/home/jochi/code/cartulary/tmp/node-runtime/bin:$PATH"`.
`make task-guide ROLE=module-author OWNER=web.workbook`, `OWNER=module.timeline`
and, for the changed shared queue, `OWNER=module.workbook` selected routing.

Artifact roots below are relative to `.cartulary/test-results/`. The applicable
`target-summaries/<target>.json`, `rows/<row-id>.json`, browser group
`playwright-report.json`, or `unit-logs/.../runner.json` contain exact selections
and outcomes. Graph work-unit counts are not test-case counts.

| Run root | Command/selection | Result and disposition |
| --- | --- | --- |
| `20260922T150810Z-p29386` | Service-backed Timeline blank-create and typing-ack measurements | PASS 16/16 graph units; pre-change production baseline. |
| `20260922T151152Z-p66386` | Service-backed `draft_create_continuity`, instrumented | Expected FAIL; old-value native event trace retained in test attachment. |
| `20260922T151450Z-p1427` | Same exact scenario without accessor diagnostics | Expected FAIL; excludes diagnostic interference. |
| `20260922T151439Z-p883` | Workbook `timeline_editor_recovery_settlement` | Expected FAIL; detached alias stayed on the draft key. |
| `20260922T151604Z-p34232` | Isolated event-order correction, exact replacement | PASS 11/11 graph units. |
| `20260922T152316Z-p71381` | Replacement, continuous typing and uncertain replay | FAIL during cutover: active editor fell back to its original row snapshot after revision settlement. Corrected to current committed-row fallback. |
| `20260922T152527Z-p8290` | Same three browser rows | PASS 13/13 graph units. |
| `20260922T152954Z-p46844` | New native-input and composition browser rows | PASS 11/11 graph units. |
| `20260922T153335Z-p86837` | Thirteen selected Timeline browser rows | PASS 15/15 graph units, including new detached capture and all prior activation/continuity/replay rows. |
| `20260922T153317Z-p83435` | Seven Timeline frontend rows | PASS 8/8 graph units. |
| `20260922T153319Z-p83662` | Eleven Workbook frontend rows | 11/12 graph units passed; continuity assertions exposed unretired scheduling state. Corrected with post-publication bubble cleanup. |
| `20260922T153738Z-p32293` | Real-owner recovery and deferred continuity | PASS 3/3 graph units, including overflow and multiple owned scalar fields. |
| `20260922T153320Z-p84071` | Shared queue owner row | PASS 2/2 graph units; includes normalization/authorization/replay fencing. |
| `20260922T153748Z-p32957` | `make agent-finalize` | FAIL at JSON shape: authored catalog was newer than generated topology. Regenerated first; no hand edits to generated output. |
| `20260922T153846Z-p34089` | `make generate` | PASS. |
| `20260922T153856Z-p37005` | `make agent-finalize` without `RESULTS_DIR` | PASS; retained full-run maintenance skipped. |

Intermediate corrections also included catalog ASCII title ordering, generating
new browser groups before execution, updating an old equality-based settlement
test to capture revisions, and giving the equal-valued gesture fixture a distinct
logical operation signature as production does. Typecheck exposed stale imports
and removed interface arguments during cutover; these were repaired.

Final verification of the completed production source:

| Run root | Command/selection | Result |
| --- | --- | --- |
| `20260922T153938Z-p40937` | Same two service-backed measurement rows as baseline | PASS 16/16 units, both unchanged budgets. |
| `20260922T154346Z-p77798` | `make format` | PASS. |
| `20260922T154351Z-p82141` | `make generate` | PASS. |
| `20260922T154401Z-p85053` | `make agent-finalize` | PASS before final broader checks; `RESULTS_DIR` unset. |
| `20260922T154505Z-p89207` | Timeline frontend slice | PASS 8/8 units; 33 selected tests. |
| `20260922T154505Z-p89243` | Workbook frontend slice | PASS 12/12 units; 42 selected tests, including real capture suspension/retirement and late receipts. |
| `20260922T154505Z-p89288` | Timeline service-backed browser slice | PASS 15/15 units; 13 scenarios, zero failures or retries. |
| `20260922T154505Z-p89328` | `make frontend-typecheck` | PASS. |
| `20260922T154505Z-p89343` | `make frontend-import-boundary-check` | PASS. |
| `20260922T154505Z-p89369` | `make lint-biome` | PASS. |
| `20260922T154814Z-p29748` | Existing Evidence-owned screenshot/capture integration | PASS 2/2 units; both capture orderings. |
| `20260922T154505Z-p89383` | `make lint-markdown` | PASS. |
| `20260922T155008Z-p30864` | Closing `make lint-markdown` | PASS after handoff and Core 03 cross-reference updates. |

An attempted duplicate catalog route for the two existing screenshot integration
tests was rejected before execution. Those cases already belong to
`module.evidence.frontend_unit.timeline_file_draft_recovery`; the duplicate edit
was removed and the current owner guide used. No routing ownership was changed.

Exact final narrow selections are recorded in each run's `run-manifest.json`
`declared_inputs.OWNER` and `declared_inputs.ROWS`; use them with `make test-slice`
or `make service-backed-test-slice`. This includes the shared queue row
`module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8c99e88779`,
which passed 14 selected tests in `20260922T153320Z-p84071`.


Exact Make invocations for the final selections (with the command-local Node
`PATH` above):

```sh
make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_scalar_editor_e5035bb033,module.timeline.frontend.timeline_editor_draft_registry_0de7a147c1,module.timeline.frontend.timeline_mutation_models_a110000001,module.timeline.frontend.timeline_row_mutation_coordinator_1a7e2c9b44,module.timeline.frontend.draft_create_activation,module.timeline.frontend.the_workbook_blank_row_action_submits_exactly_on_ad8dcf13e4,module.timeline.frontend.the_workbook_create_payload_builder_omits_zero_f_deaef46bf2
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_editor_recovery_settlement,web.workbook.regression.timeline_deferred_continuity,web.workbook.regression.timeline_row_menu_lifetime,web.workbook.regression.workbookshell__autosave_queues_a_follow_up_scala_3b8fa7ce27,web.workbook.regression.workbookshell__autosave_suppresses_duplicate_pen_293b7b0803,web.workbook.regression.workbookshell__grid_preserves_an_in_flight_draft_d12ffd720f,web.workbook.regression.workbookshell__grid_preserves_draft_row_edits_ac_cd490037ce,web.workbook.regression.grid_draft_lifetime,web.workbook.regression.workbookz_mutation_runtime_surface_continuity_97808fe8ae,web.workbook.regression.workbooktimelinemodel_builds_scalar_collection_e_13cb3751a8,web.workbook.regression.timeline_rows_owner_initialization_ccdda85db3
make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8c99e88779
make test-slice OWNER=module.evidence ROWS=module.evidence.frontend_unit.timeline_file_draft_recovery
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.pending_capture_detached,module.timeline.browser.pending_capture_composition,module.timeline.browser.pending_capture_raw,module.timeline.browser.draft_create_continuity,module.timeline.browser.draft_create_activation,module.timeline.browser.draft_create_cancellation,module.timeline.browser.draft_create_controls,module.timeline.browser.draft_create_rejection,module.timeline.browser.draft_create_uncertainty,module.timeline.browser.explicit_blank_create,module.timeline.browser.spreadsheet_creation,module.timeline.browser.spreadsheet_uncertain_create,module.timeline.browser.deferred_continuity
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6,module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13
```

## Performance and accessibility limits

Baseline measurements use fixture `cartulary.perf.large_grid.v1`, profile
`ac043_large_grid_snapshot_v1`, one warmup and 100 measured samples, p95:
blank-create 68.1 ms against 150 ms; typing acknowledgement 31.8 ms against
100 ms. Final measurement run `20260922T153938Z-p40937` passed 16/16 graph
units: blank-create p95 73.9 ms and typing acknowledgement p95 31.5 ms, with
the same fixture digest and sample counts. No predicates, budgets, fixture
traffic or measurement code changed. This work makes no speed-improvement claim.

Native clipboard insertion and keyboard paths run in the production Chromium grid.
Synthetic browser composition lifecycle events verify retained editing and DOM
identity, but do not establish OS IME compatibility. No manual OS IME or assistive
technology session was performed. DOM click-only activation establishes that event
path, not complete screen-reader compatibility. No new conformance claim is made.

`RESULTS_DIR` remains unset: narrow slices are not eligible full warm-check
evidence. Retained-run maintenance is skipped. No full repository test suite,
visual golden refresh, deployment or release-evidence publication is claimed.

## Acceptance assessment

All applicable rows pass within this slice; non-applicable rows have scope rationales.

| Row | Assessment | Evidence or scope rationale |
| --- | --- | --- |
| A001 | PASS | Exact Core owner-to-change map above; source and verification owners separated. |
| A002 | PASS | Selection rubric records correction, benefit, extension path, removal, risks and validation for every gap. |
| A003 | PASS | Clean initial branch/commit recorded; current authored owner catalogs, local guides, generated policy and grid boundary inspected. |
| A004 | PASS | No design literals or token registry introduced. |
| A005 | N/A | No theme controls or palette changes. |
| A006 | N/A | No density, sizing, padding or geometry changes; relevant capture/typing measurements retained. |
| A007 | PASS | Existing zero-field, incomplete fact, current-role/closed admission and independent Evidence/Create browser tests. |
| A008 | N/A | No responsive thresholds, viewport accessors or inspector clamp changes. |
| A009 | PASS | Existing production-grid continuation and safe-navigation focus checks; no overflow ownership change. |
| A010 | N/A | Inspector routing and confirmation dispatch unchanged; independent authoring context is checked under A014. |
| A011 | PASS | Original replacement, continuous selection, detached acceptance/remount and deferred-continuity tests. |
| A012 | PASS | Existing transaction generation retained; exact lost-response replay, duplicate activation and never-dispatched normalization fence tests. |
| A013 | PASS | Real-owner rejection/retry/discard, detached settlement, orphan prevention and captured replay evidence. |
| A014 | PASS | Native editing matrix, context/revision settlement, overflow retention and runtime lifetime tests. |
| A015 | PASS | Baseline guards and existing same-field queue/conflict behavior retained; recovery tests run through existing owner. |
| A016 | PASS | Current admission remains independent of raw authoring; production closed/viewer checks and retained-store lifetime fixtures. |
| A017 | PASS | Existing runtime replacement/concealment/retirement tests plus Timeline capture suspension/late-receipt fixture. |
| A018 | N/A | Evidence lifecycle/overlay/preview matrix unchanged; independent paperclip action remains covered. |
| A019 | PASS | Native keyboard, semantic focus, activation and recovery paths covered; manual IME/AT limits explicitly stated above. |
| A020 | PASS | Relevant editor variants cover raw Unicode, multiline/paste, clearing, read-only and composition; visual geometry unchanged. |
| A021 | PASS | Semantic production-grid identities, detached remount, one trailing draft and unchanged measurement predicates. |
| A022 | N/A | No visual styling/golden change; no new visual or publication claim. |
| A023 | PASS | Existing semantic UI-contract selectors; no component/class/vendor-coordinate selectors added. |
| A024 | PASS | Product code/tests/generators have no new Markdown dependency; documentation remains human review input. |
| A025 | PASS | Authored catalogs changed before Make generation; finalization validates generated routing. |
| A026 | PASS | Private interfaces migrated together; superseded paths removed; no wire/data migration. |
| A027 | PASS | Owner map, characterization, cutover/removal, exact artifact-backed results, limitations, retained-maintenance skip and rollback are recorded. |

## Changed files

Production changes are confined to Workbook's local draft store/queue and Timeline
capture, editor, command, presentation and composition owners. The new retained
identity model is `apps/web/src/workbook/timeline/models/TimelineCaptureLifecycle.ts`.
The removed row-wide discard model is documented above. Supporting files are the
existing scalar/collection, row, registry, model, queue and real-owner tests;
`apps/web/e2e/timeline-workbook.spec.ts`; Core 03; three Timeline source guides;
the two handoffs; authored catalogs for Timeline/Workbook; and generated browser
batch/topology index outputs. No native Create-button source was changed.

The final branch and commit remain `main` at
`0ecfe278b70bc33c0478a10671ceddfbf2e7c899`, with this slice uncommitted. Initial
user changes were absent; no unrelated changes were reverted. `git diff --check`
passes. Generated files changed only through Make generation/finalization.

## Rollback and handoff

Rollback is a coordinated revert of this source/specification/test/catalog slice,
followed by `make generate` and the same owner-selected checks. Keep the earlier
native Create activation correction. Existing accepted records and current wire
contracts remain valid; no data rollback or migration is necessary. Do not revert
unrelated user changes that arrive after this work.

The known replacement limitation in the earlier activation handoff is resolved by
this slice; its historical observations remain intact. The current review artifact
is the working-tree diff. No commit, push or deployment was performed.

