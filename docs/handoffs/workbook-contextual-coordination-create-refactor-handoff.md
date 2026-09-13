# Contextual coordination creation and recovery handoff

## Baseline, authority, and scope

Execution baseline revalidated: clean `main` at
`3425d648c41feabfca7400a76220542ba1c32f92`. Only root `AGENTS.md` applies.
The localized digest read order, Core 01 creation/discovery/inspector/relationship
owners, Core 02 shared artifacts and history, Core 03 continuity/pending work and
coordination, Core 04 authorization, and domain/design boundaries were inspected
during planning. The digest and prior handoffs are advisory/evidence, not authority.
The user approved durable source associations and the implementation plan.

Scope is exactly the 12 contextual coordination actions below, necessary neutral
helpers, owner/projection corrections, implementation, and verification. Standalone
creation redesign, existing-record editing, acknowledgement/closure workflows,
general relationship management, reporting, new navigation, dependencies, browser
persistence, migrations/backfills, digest edits, commits, pushes, deployment, and
analyst-data modification are excluded. Test data belongs to isolated harnesses.
No executable artifact may depend on Markdown.

## Execution tracker

| Workstream | Dependency | State | Exit |
| --- | --- | --- | --- |
| CCA-01 source context and characterization | none | DONE | Every entry has an explicit adopted disposition and tested projection. |
| CCA-02 retained authoring and target references | CCA-01 | DONE | Four variants retain complete authoring with usable references. |
| CCA-03 atomic attempts and reconciliation | CCA-02 | DONE | Exact replay, atomic links, complete receipts and read-only recovery pass. |
| CCA-04 integrated behavior and security | CCA-03 | DONE | Matrix, race, security, browser, accessibility and visual evidence pass. |
| CCA-05 final validation and handoff | CCA-04 | DONE | Terminal checks cover completed handoff bytes. |

Only the current row may be IN_PROGRESS. Update paths, commands/results, risks,
disposition and next action before advancing. Never advance past a blocked dependency.

## Source-context and field contract

All features use `create_related.<target>`, `view_row_create`,
`view_row_create_route`, and the existing canonical target view-row endpoint.
The selected source seeds the declared create-only `coordination.source_record_id`
input, not a target semantic field. Source replacement admits exactly the source
surfaces in the corresponding target column below. Null explicitly clears source;
omission supplies none. A non-null source creates an atomic source-to-artifact
`references_artifact` link with null field key/confidence and manual provenance.

| Source | Target actions | Source field seed | Verification owners |
| --- | --- | --- | --- |
| Timeline | comm_log, handoff, status_review, lesson | none | web.workbook, module.workbook |
| Task Requests | comm_log, status_review, lesson | none | web.workbook, module.workbook |
| Decisions | comm_log, status_review | none | web.workbook, module.workbook |
| Communications Log | status_review | none | web.workbook, module.artifacts |
| Handoff | status_review | none | web.workbook, module.artifacts |
| Status Review | comm_log | none | web.workbook, module.artifacts |

Core 01 REQ-01-311, inspector registry §7.4.1A and REQ-01-503–506 govern
this approved correction. Core 03 REQ-03-256/258 govern explicit workbook-native
authoring. Unrelated semantic references remain independently editable; source
context never implies blocked/open work, follow-up, risk, or acknowledgement.

| Target | Minimum create signal | Defaults and reference contracts |
| --- | --- | --- |
| Communications Log | type, audience text, channel/meeting, summary | Server timestamp/ID; empty Decision, action Task and Party collections; null next report/privilege; audience text remains required. |
| Handoff | incoming member, current-state summary | Server timestamp/ID and outgoing actor; empty Task/Decision/risk collections; null next checks/acknowledged time. Risks use parent-scoped child identities, not record IDs. |
| Status Review | current-state summary | Server timestamp/ID and review actor; empty blocked Task/pending Evidence/open Decision collections; null risks/next report. |
| Lesson | summary | Server timestamp/ID and owner actor; open initial closure state; empty follow-up Task/Evidence collections. |

All declared optional fields remain available on create. Owner string/reference
normalization applies only during preparation; raw values and omission/null remain
distinct. Defaults, source and references never manufacture minimum signal.

## Gap ledger

| Gap | Remediation and affected areas | Rationale and benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 confirmed owner/interface gap | Adopt explicit create-only source contract; project inspector/discovery/OpenAPI and atomic Artifacts creation. | Durable context without false semantic associations. | Existing paths/fields/storage retained; additive input/receipt requires coordinated client/server contract generation. | All 12 actions seed chosen source; retained/replaced/cleared sources produce exactly the intended link or none. |
| G2 confirmed draft loss | Coordination owner separates draft, origin, source, attachment, readiness and attempts. | Navigation and version changes preserve work. | Memory-local account/incident lifetime; authority fencing requires evidence. | Raw values survive navigation; replacement requires discard; selection never retargets source. |
| G3 confirmed candidate gaps | Target-owned paged member/record/Party controls and structured risk editor. | Complete consistent authoring from every source. | Shared helpers only where necessary; page absence is not deletion. | Every reference field supports paging, retained selection, loading/empty/error/unavailable and local feedback. |
| G4 confirmed request/receipt weakness | Frozen typed attempts, synchronous dispatch guard, full association receipts. | Lost responses cannot create duplicate artifacts or wrong-source links. | Preserve ordinary/linked-Note replay identities and omitted-input hashes. | Same-frame activation sends once; uncertain replay is unchanged; malformed success stays uncertain. |
| G5 confirmed presentation-dependent acceptance | Commit receipt before effects; independent read-only refresh debt. | Late acceptance survives detachment and refresh failure. | Preserve newer rows and unrelated work; no invented projections/history. | Late receipt retained; recovery reads only; stale observations cannot overwrite newer state. |
| H1 race/security/presentation hypotheses | Integrated owner/service/browser/a11y/visual matrix. | Establish behavior under real authorization and concurrent work. | No broader security or workflow redesign. | Every required scenario passes or blocks its dependent row. |

## CCA-01 investigation and execution

Inspected shared/Timeline create hooks, reducer/form/adapter, target schemas,
reference policies, retained Task/Decision and Note owners/readers, mutation runtime,
Artifacts admission/transaction/idempotency, application receipt adapter, and routes.
All 12 baseline seeds are empty. The legacy reducer returns null on changed subject;
Timeline candidates contain only members; the legacy adapter creates a fresh ID per
call and retains only record/change-set/view IDs; Timeline completion requires the
original presentation to remain current. These are confirmed code observations;
unreproduced race consequences remain hypotheses.

Planning verification (same baseline):

- Requested task guides passed for web.workbook, module.artifacts and module.workbook;
  additional planned input owners platform.viewschema, package.view_contracts and
  package.protocol_ts also passed.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_create_related_workflow_stale_results_31f0cbe2d1,web.workbook.regression.inspector_related_record_model_20a3b73f6d,web.workbook.regression.contextual_task_decision_recovery,web.workbook.regression.note_create_recovery`:
  PASS, `.cartulary/test-results/20260913T125130Z-p48165`.
- `make test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.artifact_admission_and_replay_hashing_are_owner_local`:
  PASS, `.cartulary/test-results/20260913T125306Z-p50157`.
- Execution baseline `git diff --check`: PASS; clean checkout before this handoff.

CCA-01 exit: DONE. Adopted Core 01 REQ-01-674 and the Core 03/04 correction
(AC-569), the exact 12 input seeds, four nullable create descriptors, source receipt
union, exclusive field/input seed projection and owner-declared optional nullability.
Updated authored OpenAPI release accounting; historical release snapshots unchanged.
Contractgen and view adapters project both seed variants. Characterization demonstrates
legacy coordination draft loss on source version changes; the new owner will replace
that consumer. No unresolved source-context choice remains.

Changed areas: Core 01/03/04; view schemas and OpenAPI owner inputs; contractgen;
view-contract types/tests; viewschema discovery tests; necessary seed adapters and
existing regression tests. Generated roots changed only through Make.

- `make generate`: initial FAIL at `.cartulary/test-results/20260913T130816Z-p53918`
  because authored compatibility accounting lacked the approved changes; related,
  resolved. PASS at `.cartulary/test-results/20260913T130940Z-p55506`.
- `make format`: PASS, `.cartulary/test-results/20260913T131029Z-p58686`.
- `make test-slice OWNER=platform.viewschema`: PASS,
  `.cartulary/test-results/20260913T131038Z-p62978`.
- `make test-slice OWNER=package.view_contracts`: PASS,
  `.cartulary/test-results/20260913T131039Z-p63203`.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260913T131040Z-p63482`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_related_record_model_20a3b73f6d,web.workbook.regression.contextual_task_decision_recovery,web.workbook.regression.note_create_recovery`:
  PASS, `.cartulary/test-results/20260913T131050Z-p64765`.

Remaining implementation risk: discovery now advertises the adopted input while
runtime admission and retained authoring await CCA-02/03; this intermediate tree
must not be deployed. Next: CCA-02 retained coordination authoring and target paging.

## Compatibility, rollback, and acceptance

Ordinary requests omitting the new input retain their request hashes and stored
receipts. Linked Notes retain their route identity and receipt format. No new tables,
record types, view IDs, or backfill are required. Existing artifacts remain as
stored; prior empty context is never inferred or backfilled. Rollback must revert this seam's
owners, projections, generated output, implementation and tests consistently; it
must never delete committed analyst artifacts or links.

Digest review adopts keyboard/focus/non-color/local feedback and async recovery,
adapts dense paging/responsive authoring to existing controls, and rejects generic
advice as product authority or a parallel design system. The explicit A001–A027
dispositions and evidence are below. Retained-run maintenance was skipped with
RESULTS_DIR unset; no exact-source successful full warm check evidence is claimed.

## CCA-02 execution

IN_PROGRESS. Building one retained four-variant coordination owner with independent
source, raw fields, presentation attachment and target-owned paged references.

CCA-02 exit: DONE. The shared coordination owner, four explicit target variants,
attachment hook/context, complete form and neutral staged reference control now
cover all 12 action contracts in focused behavioral tests. Source replacement and
clear preserve raw values; required signal, Unicode/string limits, exact references,
structured risks, timestamps, defaults and omission/null are validated locally.
Authority/source changes withdraw readiness; paged review retains unavailable IDs.
Runtime routing and submission transport are connected together in CCA-03 so an
intermediate transport cannot use the obsolete submission fallback.

Additional confirmed gap G6: optional `comm_log.privilege_tag`,
`handoff.next_checks`, and `status_review.active_risks_summary` projected
`clearable=false` despite their adopted optional-string/null behavior. Corrected
these authored schema projections and regenerated output. Benefit: create controls
and server admission agree on explicit clear; existing identities and storage are
unchanged. Binary acceptance: omitted fields stay omitted, explicit null and
normalized empty text prepare as clear, and required text never clears. Focused
optional-field tests pass; service admission evidence remains CCA-03/04 work.

Paths: `features/coordination/coordinationCreateModel.ts`,
`WorkbookCoordinationCreateOwner.ts`, `CoordinationCreateForm.tsx`,
`CoordinationCreateContext.ts`, `useCoordinationCreateAttachment.ts`,
`coordinationCreateAuthoring.test.tsx`; neutral
`components/WorkbookAuthoringReferenceControl.tsx`; nullable create-input decoder;
authored frontend ownership/test routing and three optional-text view projections.
All frontend paths are under `apps/web/src/workbook/`.

- `make generate`: FAIL at `.cartulary/test-results/20260913T132448Z-p69275`
  because the added authored test row was not ASCII-sorted; corrected ordering.
  PASS at `.cartulary/test-results/20260913T132528Z-p72825`.
- Initial `make format` and `make frontend-typecheck` were blocked by that same
  routing-order check before execution. `make format`: PASS at
  `.cartulary/test-results/20260913T132549Z-p75772` and after fixes at
  `.cartulary/test-results/20260913T132710Z-p81485`.
- `make frontend-typecheck`: PASS, two units, at
  `.cartulary/test-results/20260913T132605Z-p80220`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.coordination_create_authoring`:
  initial FAIL at `.cartulary/test-results/20260913T132605Z-p80154`: two
  fixture assumptions plus form focus timing corrected. PASS, all nine behavioral
  cases, at `.cartulary/test-results/20260913T132725Z-p85799`.

Risks/disposition: no material owner ambiguity; account/incident lifetime conventions
retained. Atomic persistence, uncertain attempts, runtime presentation integration
and integrated security/browser evidence remain required. Next action: CCA-03,
Artifacts-owned atomic source validation/linking and immutable client recovery.

## CCA-03 atomic creation and recovery

CCA-03 exit: DONE. Artifacts admission owns nullable source input and immutable
presence-aware hashing. Fresh transaction validation locks the source envelope,
checks the route incident and exact target/source matrix, and reuses Links,
Revisions, projection and publication owners. The existing transaction includes the
source link and replay result. Concurrent requests roll back the losing transaction
and recover the identical committed receipt. Source context never changes route
incident or source scalar fields. Omitted-input hashes and linked-Note identity and
stored bytes remain compatible; linked ordinary-create results now retain their
association facts through the owner-validated storage union.

Client transport captures ordinary operation, API base, path parameters, exact body,
transaction identity, actor/incident and reviewed context. The retained owner guards
submission before preparation, validates full receipts and association equality,
retains late acceptance independently of attachment, and separates read-only refresh
debt. Runtime authority, pending-work, history and collaboration owners now include
coordination authoring. Timeline source coordination refuses unsaved/conflicted work
without initiating a save. Inspector and recovery presentations use this owner.

Consumer review found no remaining Timeline creation action needing the old local
fallback. Removed its dispatch/reconciliation implementation and unused
`timelineRelatedRecordWorkflow` model/test, and retired that test row. The generic
fallback still serves Assessment creation; its late-result tests now exercise that
remaining consumer. Task/Decision seed helpers, shared reducer characterization,
linked Notes and Timeline Evidence owners remain intact. Added module.timeline task
guide because its retired authored test routing changed.

Paths: Artifacts `coordination_source.go`, admission/create/idempotency files and
`coordination_source_integration_test.go`; application Artifacts stored-result
adapter/tests; web coordination operation/owner/recovery/transport and tests;
Workbook runtime, shell, infrastructure, inspector, collaboration and Timeline
composition; authored source/test routing. No database schema or backfill changed.

- `make task-guide ROLE=module-author OWNER=app.workbookassembly`: rejected unknown
  owner. Assembly idempotency tests are already routed to module.artifacts; used
  that existing owner. `make task-guide ROLE=module-author OWNER=module.timeline`:
  PASS.
- `make test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.artifact_admission_and_replay_hashing_are_owner_local,module.artifacts.source_mutations.application_idempotency_adapter_preserves_stored_payloads`:
  initial compile FAIL at `.cartulary/test-results/20260913T133142Z-p92107`
  (test used an error accessor as a field); corrected. PASS at
  `.cartulary/test-results/20260913T133600Z-p99741`.
- `make format`: PASS at `.cartulary/test-results/20260913T133132Z-p87787`;
  intermediate FAIL at `.cartulary/test-results/20260913T133545Z-p95229`
  (duplicate import introduced during integration); corrected. PASS at
  `.cartulary/test-results/20260913T133915Z-p4841`,
  `.cartulary/test-results/20260913T134130Z-p31190`, and
  `.cartulary/test-results/20260913T134414Z-p40575`.
- `make frontend-typecheck`: intermediate FAIL at
  `.cartulary/test-results/20260913T133600Z-p99806` (duplicate import and missing
  nullable-receipt guard); corrected. PASS at
  `.cartulary/test-results/20260913T134144Z-p35597`.
- `make generate`: PASS at `.cartulary/test-results/20260913T133838Z-p1716`,
  `.cartulary/test-results/20260913T134108Z-p28106`, and
  `.cartulary/test-results/20260913T134350Z-p37325` as authored routing landed.
- `make service-backed-test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.coordination_source_atomic_creation,module.artifacts.linked_notes.artifact_linked_note_creation_is_atomic_across_f_bce1ac123b`:
  PASS at `.cartulary/test-results/20260913T133924Z-p9135`. Covers all 12 pairs,
  omission/null, direction/provenance/field/confidence, history/projection, invalid
  sources, concurrent replay, injected rollback and replay after deletion/closure;
  linked-Note atomicity regression also passes.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.coordination_create_authoring,web.workbook.regression.coordination_create_recovery,web.workbook.regression.contextual_task_decision_recovery,web.workbook.regression.note_create_recovery`:
  PASS at `.cartulary/test-results/20260913T134144Z-p35534`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_create_related_workflow_stale_results_31f0cbe2d1,web.workbook.regression.timeline_create_related_workflow_stale_effects_4b8ce2aa71,web.workbook.regression.coordination_create_authoring,web.workbook.regression.coordination_create_recovery,web.workbook.regression.workbookz_mutation_runtime_surface_continuity_97808fe8ae`:
  PASS at `.cartulary/test-results/20260913T134455Z-p44953` after fallback removal.

Risks/disposition: focused owner/service recovery is proven; public route/browser,
full security transitions, accessibility and reviewed visual evidence remain
mandatory. No new owner decision required. Next action: CCA-04 integrated matrix
and presentation/security evidence, then final affected-owner checks.

## CCA-04 integrated evidence

Public-route and browser integration exposed confirmed gap G7: the Artifacts
transaction returned a linked receipt after commit, but Workbook's ordinary-row
success validator rejected its association keys. Added a distinct source-linked-row
outcome and selected it in the Artifacts create adapter. Ordinary-row and linked-Note
outcomes remain closed and unchanged. Added complete-route create/replay assertions
for all four targets alongside the existing coordination defaults integration row,
and mutation-outcome unit assertions. Benefit: successful atomic creation returns
its complete receipt instead of a misleading HTTP 500. Compatibility: additive
result variant, no endpoint/storage change. Binary acceptance: 201 create and 200
replay return identical complete data, each with correlated request metadata, and
one artifact; PASS.

Integrated review also tightened callback fencing, same-version deletion
observations, stale-reference admission, bounded reference-page traversal, and
late socket refresh dependencies inside the coordination owner. Presentation
callbacks retain draft/attachment/account-generation identity; scoped suspension
hides and detaches presentation while preserving memory-local work. Additional
known views arriving during refresh remain read-only reconciliation debt. Shared
candidate reading now reuses the existing preferred row-label helper rather than
showing lifecycle enums. Visual review corrected recovery-summary clipping at phone
width and spacing between retained-draft actions. These changes resolve G2/G3/G5
edges without broadening ownership or changing source autosave.

Paths: coordination owner/form/recovery, neutral reference control/authoring reader,
new `apps/web/e2e/contextual-coordination-create.spec.ts` and
`apps/web/e2e/support/workbook/coordinationCreate.ts`, workbook accessibility/visual
scenarios; Workbook mutation outcome, application create adapter and existing
coordination integration/catalog tests; authored browser batches, frontend visual
fixture registry and test routing. No analyst data is used by these fixtures.

Commands/results to date:

- Task guide `harness.browser`: PASS. `harness.visual` is not an active owner;
  visual routing belongs to module.workbook and harness.browser.
- `make generate`: initial routing-family FAIL at
  `.cartulary/test-results/20260913T135159Z-p48103`; corrected the authored browser
  family. PASS at `.cartulary/test-results/20260913T135253Z-p51338` and
  `.cartulary/test-results/20260913T141144Z-p36510`.
- `make format`: PASS at `.cartulary/test-results/20260913T135310Z-p54752`,
  `.cartulary/test-results/20260913T140331Z-p45543`, and
  `.cartulary/test-results/20260913T141213Z-p39609`.
- `make frontend-typecheck`: PASS at
  `.cartulary/test-results/20260913T135327Z-p59166` and
  `.cartulary/test-results/20260913T141240Z-p44551`.
- Five-row coordination browser selection: initial FAIL at
  `.cartulary/test-results/20260913T135327Z-p59100` (G7 and a test using the wrong
  system-view navigation control). After correction, four scenarios PASS at
  `.cartulary/test-results/20260913T140403Z-p50113`; matrix assertion then exposed
  a test treating `collection_value_v1` as a bare array. Corrected that assertion.
- Coordination accessibility/visual selection: initial FAIL at
  `.cartulary/test-results/20260913T135450Z-p90997` (G7 and new missing goldens).
  At `.cartulary/test-results/20260913T140403Z-p50141`, accessibility PASS;
  all seven ordinary visual captures reached comparison, requiring new goldens.
  Reviewed all seven embedded PNG attachments. Reconciliation also found two
  fixture-path spelling mismatches (underscore versus Playwright hyphen); corrected
  authored registry paths before generation. Full ordinary reconciliation and
  promotion remain pending.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.coordination_collection_fields_round_trip_typed_9c30c3f538,module.workbook.integration.parties_and_coordination_system_views_query_rout_e0808abf39`:
  PASS at `.cartulary/test-results/20260913T135529Z-p22616`.
- `make frontend-import-boundary-check`: PASS at
  `.cartulary/test-results/20260913T135529Z-p22761`.
- `make test-slice OWNER=platform.viewschema`: PASS at
  `.cartulary/test-results/20260913T135529Z-p22628`.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.integration.communications_log_handoff_status_review_and_les_94a6f7d0c2`:
  PASS at `.cartulary/test-results/20260913T140403Z-p50079`, including the new
  public-route receipt/replay regression.
- `make test-slice OWNER=module.workbook ROWS=module.workbook.unit.registry_derived_writable_conflict_capabilities_c0614db52a`:
  PASS at `.cartulary/test-results/20260913T140403Z-p50088`.
- Coordination authoring/recovery plus completed Task/Decision and Note recovery
  selection: PASS at `.cartulary/test-results/20260913T141240Z-p44383` (five units),
  including new stale-presentation and socket/authority ordering cases.

Risks/disposition: CCA-04 remains IN_PROGRESS pending the complete browser matrix,
regression boundaries and reviewed visual promotion with two ordinary passes.
No further product-owner ambiguity is present. Next action: finish those checks,
record their actual results, then advance to CCA-05.

CCA-04 follow-up evidence:

- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.coordination_create_matrix,module.workbook.browser.coordination_create_retained_source,module.workbook.browser.coordination_create_response_loss,module.workbook.browser.coordination_create_refresh_only,module.workbook.browser.coordination_create_revocation,module.workbook.browser.note_create_sources,module.workbook.browser.note_create_recovery,module.workbook.browser.note_create_source_edit,module.workbook.browser.note_create_revocation,module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4`:
  at `.cartulary/test-results/20260913T141240Z-p44438`, all five coordination and
  four linked-Note scenarios PASS. The older Timeline workflow fixture failed
  because it expected an immediately visible member select; adapted that fixture
  to the declared reference picker. Its first rerun at
  `.cartulary/test-results/20260913T142047Z-p27673` exposed a missing test import;
  corrected. Timeline workflow and expanded coordination accessibility PASS at
  `.cartulary/test-results/20260913T142338Z-p69355`, 13/13 units. The latter verifies
  keyboard access to System views and account/application navigation after recovery
  at 390×480 as well as fields, local feedback, visible focus and Escape restoration.
- Full `make browser-e2e-visual`: ordinary run
  `.cartulary/test-results/20260913T141240Z-p44719` reached 248 capture intents;
  all 241 existing goldens matched, zero orphans/ambiguous mappings. Seven new
  coordination goldens were missing. Two metadata rows also required adding the
  new fixture ID to the existing expected inventory; corrected. Reviewed new PNGs
  from that run and from `.cartulary/test-results/20260913T142047Z-p27673` after
  the narrow-header correction. Below the supported width, the existing system
  strip now retains its content width inside the shell-owned horizontal overflow;
  recovery, System views and account controls no longer overlap.
- `make service-backed-test-slice OWNER=web.design ROWS=web.design.visual.ensure_the_visual_fixture_matrix_includes_defaul_979c589967,web.design.visual.run_the_owned_stack_playwright_visual_suite_with_9b08711878`:
  `.cartulary/test-results/20260913T142338Z-p69407`: fixture matrix PASS; ordinary
  metadata existence check correctly remains FAIL until seven goldens are promoted.
  `make task-guide ROLE=module-author OWNER=web.design`: PASS, added because the
  authored expected visual inventory changed.
- An attempted generator-test row caused `make generate` to reject a Go selector
  under `tools/contractgen` at `.cartulary/test-results/20260913T141625Z-p15611`.
  The public catalog admits only internal/cmd and the specific testservices tool.
  Removed that attempted row and the newly added unroutable generator subtest;
  existing generator tests remain unchanged. Public view-schema/protocol projection,
  source-schema validation and all-12 seed tests provide this seam's executable
  coverage. No harness package-admission expansion was introduced.
  `make generate`: PASS at `.cartulary/test-results/20260913T141845Z-p19650`.
  Task guide `harness.generated_artifacts`: PASS.
- `make format`: PASS at `.cartulary/test-results/20260913T142031Z-p23279`,
  `.cartulary/test-results/20260913T142321Z-p64827`, and
  `.cartulary/test-results/20260913T142657Z-p73629`.
- Coordination authoring/recovery focused selection: PASS at
  `.cartulary/test-results/20260913T142047Z-p27675`, three units.
- `make service-backed-test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.coordination_source_atomic_creation`:
  `.cartulary/test-results/20260913T142338Z-p69342` was blocked before test execution
  by concurrent test-service image warm-stamp maintenance (`mv` source missing).
  This is harness infrastructure, unrelated to product assertions. Sequential rerun
  PASS at `.cartulary/test-results/20260913T142547Z-p56041`, including combined
  source association plus analyst-selected Task reference and atomic rollback for an
  unavailable target reference.
- `env -u RESULTS_DIR make agent-finalize`: PASS at
  `.cartulary/test-results/20260913T142445Z-p52202` before the broader visual
  verification. Retained-run maintenance skipped because RESULTS_DIR was unset;
  no exact-source successful full warm check evidence was claimed.
- Server validation failures now map only declared target fields or source input
  into field-local feedback; unknown field names stay out of the form. Corrected
  fresh submission retains raw values and obtains a new transaction identity.

Compatibility and rollback review: omitted-input ordinary requests and linked-Note
operation/hash/result formats remain compatible. Linked coordination receipts are
additive and new clients require matching live discovery before dispatch. No stored
artifact field, migration or backfill was added. Disable contextual entry points
through a coherent owner/projection/generated/UI change if an operational rollback
is needed, retaining the additive input/result decoder and replay compatibility for
already accepted requests. A wholesale old-code rollback cannot be assumed to read
new stored contextual receipts. Never remove committed artifacts, source links,
target references, revisions or history to roll back presentation code.

### Visual refresh review

The second full ordinary reconciliation at
`.cartulary/test-results/20260913T142706Z-p77994` accounted for 248 capture intents,
241 active existing goldens, zero orphans and zero ambiguous mappings. Failures were
seven new missing goldens plus six expected below-minimum header comparisons and
the metadata check requiring the new files. Inspected all six changed existing
captures against their prior images: differences are confined to the header's
content width/scrolling, with existing protected drawers and grid content intact.
Keyboard tests confirm System views and account/application navigation remain
reachable through that shell-owned overflow. The repair belongs to the existing
header layout owner; no coordination-specific navigation surface or breakpoint
registry was added.

Existing refresh paths, under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

- `workbook-view-bar-zoom-200-linux.png`
- `membership-audit-inspected-narrow-linux.png`
- `membership-management-removal-narrow-linux.png`
- `metadata-review-narrow-linux.png`
- `lifecycle-review-narrow-linux.png`
- `workbook-preferences-uncertain-narrow-linux.png`

New fixture `visual.fixture.contextual_coordination_creation` is implementation
support evidence, not a Core 05 claim. Reviewed all seven final new captures from
`.cartulary/test-results/20260913T143408Z-p18580` after explicitly normalizing the
underlying grid to top/left before establishing drawer focus/scroll anchors. That
ordinary run failed only because the seven new expected files did not yet exist.
New goldens in the same snapshot directory:

- `coordination-comm-log-authoring-linux.png`
- `coordination-handoff-authoring-linux.png`
- `coordination-status-review-authoring-linux.png`
- `coordination-lesson-authoring-linux.png`
- `coordination-source-narrow-linux.png`
- `coordination-recovery-narrow-linux.png`
- `coordination-recovery-linux.png`

All captures use the existing attested Playwright renderer, dark_graphite, compact
resolved density, 100% zoom and device scale factor 1. Four authoring views and full
recovery use 1280×720; source selection uses 768×640; narrow recovery uses 390×480.
Required fields, defaults, independent references/source, uncertainty feedback and
focus remain readable. Existing control spacing/tokens are reused. No tolerance,
renderer, font, or post-anchor adjustment changed. Public update and two fresh
ordinary validation results are recorded below once complete.

Public `make browser-e2e-visual-update`: PASS,
`.cartulary/test-results/20260913T143547Z-p50240`. Promotion changed exactly
the seven new and six reviewed existing PNGs above plus the golden manifest.
Inspected every promoted image again; no unrelated snapshot was removed or
changed. Two fresh full ordinary passes follow the final source cleanup.

Broader regression review and final cleanup:

- `make frontend-import-boundary-check`: PASS (2/2),
  `.cartulary/test-results/20260913T143636Z-p81560`.
- `make backend-module-boundary-check`: PASS (3/3),
  `.cartulary/test-results/20260913T143636Z-p81546`.
- `make frontend-typecheck`: initially FAIL at
  `.cartulary/test-results/20260913T143636Z-p81492` because the updated sentinel
  fixture tested a general view ID against a literal-array `includes` parameter.
  Use explicit equality with `some`; PASS on rerun at
  `.cartulary/test-results/20260913T143829Z-p13407`.
- `make test-slice OWNER=web.workbook`: 242/243 PASS at
  `.cartulary/test-results/20260913T143636Z-p81337`; the sole failing row expected
  five legacy Timeline fallback targets. Consumer review found that its policy
  map and workflow arguments no longer affect any behavior: all supported
  Timeline create presentations use their retained owners. Removed the obsolete
  policy/model tests and source, their unused hook/composition arguments, catalog
  row/source ownership, and stale human inventory entries. Preserved the shared
  related-record command adapter and port used by generic Assessment creation.
  The existing Timeline Evidence navigation/late-acceptance test remains routed.
- `make generate`: PASS, `.cartulary/test-results/20260913T144522Z-p15994`;
  `make format`: PASS, `.cartulary/test-results/20260913T144546Z-p19274`.

Risk/disposition: this cleanup removes dead fallback ownership rather than
relaxing the obsolete expectation. No target behavior or adopted contract changes.
Next action remains closing all workbook regression rows and two ordinary visual
passes before CCA-05.

Final cleanup verification:

- `make test-slice OWNER=web.workbook`: PASS (242/242),
  `.cartulary/test-results/20260913T144630Z-p23915`. This includes the retained
  Task/Decision, Note, Timeline Evidence hook, inspector and runtime boundaries.
- `make frontend-typecheck`: PASS (2/2),
  `.cartulary/test-results/20260913T144631Z-p25353`.
- `make lint-biome`: PASS (2/2),
  `.cartulary/test-results/20260913T144719Z-p69008`.
- `make generated-artifact-policy-check`: PASS (3/3),
  `.cartulary/test-results/20260913T144720Z-p69627`.
- Final cleanup `git diff --check`: PASS; digest, lockfiles and migrations untouched.

### Advisory acceptance dispositions

These are human review dispositions against the localized digest, not new product
authority or executable claims. Owner-backed behavior and routed evidence above
control the seam.

| Digest acceptance | Disposition and evidence |
| --- | --- |
| A001–A003 authority, scope, repository state | ACCEPT: approved Core 01/03/04 correction; actual clean baseline recorded; bounded source/test/generated diff reviewed. Core 02 shared artifacts/history and domain/design boundaries preserved. |
| A004 tokens | ADAPT: existing inspector controls and color/spacing/border tokens are reused. The recovery overlay follows existing memory-local recovery presentation geometry; bounded positioning/width literals do not define a second theme, density or breakpoint registry. |
| A005 theme; A006 density | ACCEPT for affected presentations: dark_graphite and shared compact density retained; ordinary visual matrix covers existing density variants. No density algorithm or preference semantics changed. |
| A007 creation | ACCEPT: live discovery, declared input/fields, exact minima and references verified by projection, owner, service and all-12 browser tests. Evidence/Indicator exceptions remain with their completed owners. |
| A008–A009 responsive/overflow | ACCEPT: existing shell chrome mode governs the necessary below-minimum header repair; keyboard, narrow/short and reviewed ordinary visuals preserve inspector/grid overflow and reachable application controls. |
| A010 inspector | ACCEPT: complete workbook unit selection, all-12 semantic dispatch and completed Timeline workflow browser boundary; disabled known actions and unknown-additive omission retained. |
| A011 continuity | ACCEPT: retained raw values/source identity survive selection, navigation, source version changes and presentation detachment; callback leases prevent obsolete attachment mutation. |
| A012 transactions | ACCEPT: shared secure transaction IDs, synchronous activation guard, immutable attempt and unchanged replay; corrected fresh requests receive new IDs. |
| A013 queue recovery | ADAPT: ordinary blocked-edit Retry/Discard behavior remains unchanged. Uncertain coordination submission uses the explicitly adopted exact-replay rule; accepted recovery reads only. |
| A014–A015 editing/conflict | ACCEPT as regression boundary: no unrelated source autosave; retained owners wait earlier committed work and refuse unresolved Timeline source work. Existing editing/conflict tests remain routed and pass. |
| A016–A017 async states/refresh | ACCEPT: candidate loading/empty/failure/unavailable, retained off-page selections, uncertain submission and accepted refresh debt are distinct. HTTP/socket ordering and stale reads cannot fabricate or overwrite projection rows. |
| A018 Evidence | ACCEPT as unchanged boundary: completed Timeline Evidence attachment/recovery and explicit creation browser tests pass; no two-stage Evidence workflow was copied. |
| A019–A020 accessibility/components | ACCEPT: four declared authoring variants, field-local errors, labels, keyboard completion, Escape and focus restoration; existing controls encode state through text and disabled attributes, with no new animation. |
| A021 virtualization | NOT AFFECTED: no grid adapter, row virtualization, measurement or result-size algorithm changed. Existing continuity tests and ordinary grid visual fixtures remain regression evidence; no new performance claim. |
| A022 visual fixtures | ACCEPT subject to the two ordinary passes recorded in the CCA-04 exit: registered seven captures, reviewed six necessary header refreshes, public promotion only, no tolerance or renderer changes. |
| A023 selectors | ACCEPT: new fixtures use canonical view IDs, record IDs, field keys, accessible control names and explicit owner state. UI package source is unchanged. |
| A024–A025 test authority/generated artifacts | ACCEPT: no executable dependency on Markdown; authored schemas/catalog/routing precede public generation; policy and drift checks govern derivatives. |
| A026 boundaries | ACCEPT: source input/association/receipt are the user-approved durable correction. No endpoint, stored field, migration, dependency or additional association write invented. |
| A027 handoff | ACCEPT: controlling tracker, actual commands/results, scope, compatibility, skips and rollback are recorded; final Markdown/diff confirmation follows completion. |

First fresh full ordinary `make browser-e2e-visual`: PASS (12/12 graph units),
`.cartulary/test-results/20260913T144629Z-p23784`, after promotion and final source
cleanup. No golden changed during ordinary verification.

Pre-terminal `make lint-markdown`: PASS,
`.cartulary/test-results/20260913T144918Z-p87172`; the completed handoff receives
another Markdown/diff check in CCA-05.

Additional affected public-contract verification:

- `make test-slice OWNER=package.protocol_ts`: PASS (8/8),
  `.cartulary/test-results/20260913T145152Z-p12613`, including generated HTTP
  bindings and the protected-browser-bundle boundary.
- `make openapi-compatibility-check`: PASS (4/4),
  `.cartulary/test-results/20260913T145153Z-p14987`. The approved source input
  and linked receipt are additive to existing create routes; release history is
  unchanged and the authored candidate change set records the correction.

Final-source integrated browser selection:

`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.coordination_create_matrix,module.workbook.browser.coordination_create_refresh_only,module.workbook.browser.coordination_create_response_loss,module.workbook.browser.coordination_create_retained_source,module.workbook.browser.coordination_create_revocation,module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4,module.workbook.accessibility.coordination_create_authoring_recovery`:
PASS (15/15 graph units), `.cartulary/test-results/20260913T145212Z-p20388`.
This closes the all-12 action matrix, retained replacement/clear, real post-commit
response loss/exact replay, read-only refresh recovery, scoped revocation/late
acceptance, all-four keyboard/field-local authoring and completed Timeline
Task/Decision/Note/Evidence workflows after removal of the obsolete fallback.

Second fresh full ordinary `make browser-e2e-visual`: PASS (12/12 graph units),
`.cartulary/test-results/20260913T145140Z-p89161`. Both ordinary passes followed
public promotion, inspected the same final implementation, and left the golden
manifest and images unchanged. The complete ordinary inventory has 248 registered
captures with no missing, orphaned or ambiguous mapping.

### CCA-04 exit

Full `make browser-e2e-a11y`: PASS (14/14 graph units),
`.cartulary/test-results/20260913T145436Z-p61329`. This broadens the focused
coordination keyboard checks to the existing drawers affected by shared header
overflow. Both required ordinary visual passes and the final-source functional
selection pass. G1–G5 and G7 have passing binary acceptance; H1 race/security
hypotheses are covered by the unit/service/browser cases described above.

Disposition: CCA-04 DONE. No unresolved source-context decision or confirmed
implementation defect remains. Risks are the documented memory-local lifetime
and additive stored-receipt rollback compatibility, not deferred work. Source,
selection, references, dispatch, acceptance and refresh have separate ownership;
no unrelated source editing is initiated. Next action: CCA-05 finalization,
affected terminal checks and final handoff-byte validation.

## CCA-05 final validation

Inputs and routing were registered as implementation landed: web.workbook source
and test ownership, module.artifacts atomic create integration, module.workbook
route/browser/accessibility/visual rows, browser batches and visual fixture registry.
Removed fallback test routes were removed with their sources. Generated derivatives
were refreshed through Make after authored projections. No executable artifact
depends on this handoff or any Markdown.

Terminal plan: run `env -u RESULTS_DIR make agent-finalize` first; broaden the
already-passing narrow owner tests to `make test-fast` because the shared create
response/seed projections cross source owners. Run `make lint`, affected boundaries
and generation/drift/policy checks, and refresh narrow atomic/replay service evidence.
Existing two fresh full visual passes and full accessibility evidence above cover
the final implementation; no further production edits are planned.

RESULTS_DIR remains unset: no exact-source successful full warm `check` run exists.
Retained-run maintenance is explicitly skipped. No performance, release or Core 05
claim is made. Final scope review and Markdown/diff checks must cover completed
handoff bytes before this workstream is reported complete.

`env -u RESULTS_DIR make agent-finalize`: PASS (1/1 graph unit),
`.cartulary/test-results/20260913T145801Z-p97457`, before broader terminal checks.
The finalizer refreshed/checked generated structure, JSON shape and catalog
coverage; retained-run maintenance was skipped because RESULTS_DIR was unset.

Terminal results:

- `make generate-drift`: PASS (4/4),
  `.cartulary/test-results/20260913T145837Z-p1525`.
- `make frontend-import-boundary-check`: PASS (2/2),
  `.cartulary/test-results/20260913T145837Z-p1875`.
- `make backend-module-boundary-check`: PASS (3/3),
  `.cartulary/test-results/20260913T145837Z-p1945`.

- `make lint`: PASS (11/11),
  `.cartulary/test-results/20260913T145837Z-p1997`, including affected backend,
  frontend typing, authored-source and script/shell checks.

- `make generated-artifact-policy-check`: PASS (3/3),
  `.cartulary/test-results/20260913T150008Z-p39372`.

### Final scope and compatibility review

Branch remains `main`; HEAD remains
`3425d648c41feabfca7400a76220542ba1c32f92`. Scope review found 110 dirty paths:
83 modified tracked files, four deleted obsolete fallback source/tests, and
23 new paths. Nothing is staged. Changes are confined to the coordination seam,
necessary neutral/shell helpers, adopted Core corrections, authored contracts and
routing, their Make-generated derivatives, reviewed visual fixtures and this
handoff/human source inventory. No digest, dependency/lockfile, migration, deployment
or analyst-data change is present; no commit, push or deployment was performed.

Accepted product decision: the user-approved durable source link replaces empty
context seeds. No additional owner decision was needed. Three existing nullable
optional-scalar projections were corrected to the unambiguous adopted owners.
Canonical view identities, routes and stored field membership remain unchanged.
Old ordinary omitted-input hashes and linked-Note stored replay formats remain
readable; source identity/null enters only new supplied-input comparisons. New
linked ordinary receipts require the additive replay reader during any operational
rollback. Disable contextual entry points coherently if rolling back presentation;
preserve committed artifacts, links, references, history and replay records.

Deliberate limits/skips: authoring remains memory-local to the established
account/incident runtime; browser persistence and migration/backfill are excluded.
Retained successful-run maintenance is skipped with RESULTS_DIR unset. Full
`check`, CI/release, deployment and performance/measurement suites are not claimed:
affected narrow service/browser selections plus the broader fast, lint, boundary,
full accessibility and two ordinary visual passes cover this seam. No virtualization
or performance algorithm changed. No executable check derives authority from prose.

Terminal follow-up:

- `make service-backed-test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.coordination_source_atomic_creation,module.artifacts.linked_notes.artifact_linked_note_creation_is_atomic_across_f_bce1ac123b,module.artifacts.collection_semantics.artifact_collection_mutations_preserve_typed_col_ab3546b03e,module.artifacts.collection_semantics.artifact_risk_references_use_canonical_parent_sc_b69b618649`:
  PASS (3/3 graph units), `.cartulary/test-results/20260913T150007Z-p39054`.
- `make test-fast`: 648/649 units PASS at
  `.cartulary/test-results/20260913T145837Z-p1952`; the sole failing assertion in
  `internal/modules/workbook/openapi_contract_test.go` still required the old
  ordinary-only response schema. Updated the existing routed contract test to
  verify the approved linked/unlinked response union, required complete row/change
  set/metadata, exact association type and four-target restriction, optional
  nullable source inputs, and exclusion from completed non-coordination requests.
  This closes stale test projection evidence; no production or visual source
  changed. CCA-05 remains IN_PROGRESS pending focused and broader reruns.

- `make format`: PASS (2/2),
  `.cartulary/test-results/20260913T150419Z-p86772`.
- `make test-slice OWNER=module.workbook ROWS=module.workbook.unit.openapi_record_mutation_contract_12a2ffaaf6`:
  PASS (1/1), `.cartulary/test-results/20260913T150426Z-p91063`, including the
  strengthened complete-receipt and optional-source contract assertions.

Second terminal `env -u RESULTS_DIR make agent-finalize`: PASS (1/1),
`.cartulary/test-results/20260913T150504Z-p91620`, after the test-only contract
correction and before broader reruns. Retained-run maintenance again skipped
with RESULTS_DIR unset. No generated drift or routing gap remains.

- Broader `make test-fast` rerun: PASS (649/649 graph units),
  `.cartulary/test-results/20260913T150541Z-p95514`. This covers the affected
  frontend unit owners, View Contracts, protocol bindings, view-schema projections,
  backend owner/assembly unit tests and completed seam regression boundaries.
- `make lint` rerun: PASS (11/11),
  `.cartulary/test-results/20260913T150541Z-p95530`.
- Final branch/HEAD/index and scope recheck: unchanged `main` at the baseline HEAD,
  110 dirty paths (83 modified, four deleted obsolete files, 23 new), zero staged.
  `git diff --check`: PASS.

### CCA-05 disposition

All required implementation, owner/projection, generated, routing, unit, service,
browser, accessibility and reviewed visual work is complete. Earlier failing
commands remain recorded with their corrections and successful reruns; no failing
product assertion is deferred.

Final content `make lint-markdown`: PASS,
`.cartulary/test-results/20260913T150719Z-p7379`; `git diff --check`: PASS.
The same commands confirm the completed tracker bytes once more after writing
this disposition; their final Make result is emitted in the terminal evidence.
No documentation or source edits follow successful confirmation.

Disposition: CCA-05 DONE. No blockers or deferred seam work remain. RESULTS_DIR
remains unset and retained-run maintenance is skipped. Compatibility and rollback
limits above apply. Work stops at this seam; no commit, push or deployment was
performed.
