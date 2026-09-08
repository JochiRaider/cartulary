# Incident Lifecycle Controls Refactor Handoff

## Baseline, authority, and scope

Execution entry: main, HEAD 58d3b8c5e1b96b44b9aa938a2216277bd8905894,
clean tree. Root AGENTS.md is the only applicable instruction file. Stack:
React 19, TypeScript 6, Vite 8, pnpm 10.33, Vitest/Testing Library, Playwright,
Biome, and Go 1.27.1. Preserve unrelated work; no reset, commit, push or deployment.

Core 01 §3.3.5.3.2 / REQ-01-585–591 owns lifecycle, versions, current admin
admission, exact replay, resource invariants and closed-operation boundaries.
REQ-01-496 and the required reason binding own normalization and scalar limits;
§3.3.6 owns lifecycle errors. Core 03 §4.3.2 / REQ-03-287–288 owns persistent
closed mode and terminal rejected drafts; §13 owns keyboard boundaries.
Core 04 session requirements and REQ-04-017/023/026/028 own session lifetime,
current membership/role, retained reads and deployment-admin separation.
Design §§3, 7, 10, 12, 14–15 and domain vocabulary apply within their scopes.
No adopted-owner contradiction was identified in planning.

The digest was read in its localized order and its stale maps revalidated against
source. It is advisory and remains unchanged. Completed metadata and membership
handoffs are implementation evidence. The NLSpec research essay is not a product
owner. User chose inline review and retention of reason/recovery on demotion while
membership remains valid; mutations remain blocked until current admin authority
is revalidated. Session/access loss and reviewed incident departure retire work.

Allowed paths: lifecycle app/client/controller/presentation and tests; minimal
App/departure/shared resource/workbook integration; relevant incidents contract
and integration tests; semantic selectors, browser support/readiness tests and
reviewed visual goldens; authored verification/source ownership and their
Make-generated derivatives; demonstrated lifecycle projection corrections and
matching compatibility accounting; this handoff. No preferences/editor/audit
redesign, new policy/routes/fields, dependency or storage change, or digest edits.
The pre-existing metadata error-reason mismatch remains excluded.

Generated boundaries follow tools/generated_artifact_policy.json: internal/gen,
protocol-ts/view-contracts/ui-contracts generated roots, assembled OpenAPI,
generated topology/task files and visual golden manifest. Do not hand-edit.
Tests, generators and runtime evidence must not depend on Markdown.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| LC-01 — Baseline, owners, characterization | DONE | Eleven frontend failures and normalized-reason projection defect retained. |
| LC-02 — Typed client and operation owner | DONE | Twenty-five focused client/controller cases, backend projection and typecheck pass. |
| LC-03 — Presentation and integration | DONE | Retained owner, neutral publication, sequential departure and focused integration pass. |
| LC-04 — Browser, accessibility and visual evidence | DONE | Required journeys and failure transitions pass. |
| LC-05 — Final verification and completed handoff | DONE | All required evidence passed; completed handoff and scope audit retained. |

Only the current row is IN_PROGRESS. Record paths, commands, results, decisions,
risks and next action at each DONE/BLOCKED exit before advancing. Never cross a
blocked dependency.

## Ownership and planning evidence

| Concern | Existing or planned owner |
| --- | --- |
| Lifecycle draft/review/admission/replay | New App-retained lifecycle controller and typed client. |
| Session and current authorization | AppSessionController and existing authorization recovery port. |
| Latest incident resource | Neutral App acceptance/publication; feature controllers consume without owning each other. |
| Drawer/section/focus | useIncidentControlsDrawer, IncidentControlsDrawer and workbook presentation. |
| Route departure | App and reviewAppDeparture, memberships → metadata → lifecycle → deployment users. |
| Closed writes and retained drafts | Existing workbook interaction and mutation runtime owners. |

Inspected IncidentAdminPanel/tests, App, route/departure controllers, metadata and
membership controllers/bindings, workbook drawer/identity/resource publication,
shared shell/recovery contracts, generated HTTP bindings, authored incidents
OpenAPI, backend lifecycle tests, browser lifecycle/closed-workbook tests and
visual maintenance guidance. At baseline, the handler created a key per activation, had
no synchronous write lock, conflated 409 codes and cleared reason on completion.
These were source findings before the characterization recorded below.

Planning public task guides passed for module.incidents, web.application,
web.workbook, package.protocol_ts, package.ui and web.design; make help/help-all
were inspected. Baseline commands:

- make test-slice OWNER=module.incidents
  ROWS=module.incidents.frontend.lifecycle_product_surface,module.incidents.unit.incident_lifecycle_request_and_openapi_contract_9466dfc2b1:
  PASS 3/3 graph units, .cartulary/test-results/20260908T132938Z-p3130343.
- make service-backed-test-slice OWNER=module.incidents
  ROWS=module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20:
  PASS 3/3, .cartulary/test-results/20260908T132938Z-p3130347.
- git diff --check: PASS; execution entry remains clean.

These are existing baseline checks, not reproductions of the new race matrix.

## Execution evidence

LC-01 entry: routed unchanged-production characterization preceded production edits.

### LC-01 exit

Added incidentLifecycleCharacterization.test.tsx and its conditional App-style
mount surface, an authored module.incidents row/source ownership, and lifecycle
normalization/projection boundary tests. Production source was unchanged.
make generate initially failed at 20260908T134128Z-p3150974 because new selector
titles were not ASCII-sorted. Corrected authored ordering; make generate passed
at 20260908T134215Z-p3154302.

CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.lifecycle_characterization,module.incidents.unit.incident_lifecycle_request_and_openapi_contract_9466dfc2b1
intentionally failed with 1/3 graph units passing at 20260908T134245Z-p3157380. The frontend
row retains eleven failed assertions: duplicate dispatch (two writes), changed
transaction/payload on retry, conflated transaction/transition/version recovery,
newer reason erased, old rejection on another incident, acknowledgement withheld
behind follow-up authorization, old receipt replacing accepted version 3 with 2,
and drawer/section reason loss. Historical replay itself and committed transport
loss still require real service evidence; the old-receipt test establishes local
publication regression only. No exploit or unauthorized server mutation claimed.

Backend admission accepted 4096 normalized scalars represented as 8192 decomposed
raw scalars, extra surrounding whitespace, and supplementary scalars. The sole
new backend failure is the raw OpenAPI maxLength assertion. Correct this narrow
projection in LC-02 using the existing operation annotation and generation flow;
no public operation or normalization algorithm is missing.

All run roots above use .cartulary/test-results/. No blocked dependency.
Next: LC-02 typed lifecycle client, immutable operation owner and focused tests.

### LC-02 exit

Added lifecycle model/controller and generated-operation client, neutral resource
contract/semantic acceptance helpers, 21 ownership cases and four transport cases.
Admission precedes authorization; captured requests are frozen; replay is explicit
and exact; timeout keeps the independent transport lock; confirmation and current
observation are separate. Historical replay receipts are never broadcast as
current resources. Reason editing remains raw and normalization remains server-owned.

The demonstrated raw maxLength was removed from the authored lifecycle reason
schema; existing close/reopen operations now describe the required normalized
4096-scalar binding. Four exact compatibility fingerprints were reviewed and
recorded: reason description, raw maxLength removal and two operation annotations.
The compatibility tool mechanically labels constraint removal breaking; server
admission, routes, request members and stored data do not change.

make generate initially required compatibility accounting at 135219-p3160123.
After recording the four changes, generation at 135314-p3161679 found the new
rows inherited the wrong family ID; make format at that point also stopped on
that catalog error. Corrected family/evidence routing. Generation passed at
135341-p3165026 and 135359-p3168000. make format passed 2/2 at
135408-p3171024.

- make test-slice OWNER=web.application
  ROWS=web.application.regression.lifecycle_ownership,web.application.regression.lifecycle_transport:
  PASS 3/3 at 135419-p3175391 (25 cases).
- make test-slice OWNER=module.incidents
  ROWS=module.incidents.unit.incident_lifecycle_request_and_openapi_contract_9466dfc2b1:
  PASS 1/1 at 135419-p3175396, including the formerly failing projection boundary.
- make frontend-typecheck: PASS 2/2 at 135419-p3175516.

Suffixes resolve under .cartulary/test-results/20260908T. No blocked dependency.
Next: LC-03 App-retained inline presentation, neutral publication and departure.


### LC-03 exit

One App-retained lifecycle binding now owns inline review and presentation.
IncidentAdminPanel retains summary/preferences and contains no lifecycle mutation
handler. A neutral App resource boundary fences identity, session and version,
then ingests into lifecycle/metadata without echo; workbook reads publish through
that same boundary and accepted resources feed workbook identity. Lifecycle reads
do not reload preferences. Existing metadata admission/draft behavior is preserved;
its optional broadcast suppression is the only controller integration change.

The App departure order is memberships, metadata, lifecycle, deployment users.
All owners are checked again before navigation; session/access/incident retirement
clears lifecycle and neutral resource state. Drawer return and role changes
revalidate authority. Preflight demotion cancels admission; an already dispatched
receipt may settle while hidden within its valid lifetime.

Accepted closure uses existing workbook lifecycle invalidation. Inspection found
terminal queue pause lacked an active-state recovery path. A narrow queue/runtime
extension leaves retained writes blocked after reopening, exposes the existing
incident_closed discard recovery, and permits fresh work after explicit discard;
no retained unit is automatically replayed. Empty queues can admit fresh work.
make task-guide ROLE=module-author OWNER=module.workbook passed for this necessary
integration owner.

Evidence (run suffixes under .cartulary/test-results/20260908T):

- make generate: PASS 140836-p3180345 and 141445-p3192059.
- make format: PASS 2/2 at 140853-p3183563 and 141528-p3195346.
- First integration checks at 140954 exposed cleanup return typing and tests
  activating before inline review rendered; fixed. Ownership/transport passed
  3/3 at 140954-p3188106.
- make test-slice OWNER=module.incidents
  ROWS=module.incidents.frontend.lifecycle_characterization,module.incidents.frontend.lifecycle_product_surface:
  PASS 3/3 at 141149-p3190243 (eleven characterization cases plus composition).
- make test-slice OWNER=web.application
  ROWS=web.application.regression.lifecycle_ownership,web.application.regression.lifecycle_transport,web.application.regression.lifecycle_integration,web.application.regression.metadata_departure,web.application.regression.metadata_ownership,web.application.regression.metadata_lifetime:
  5/7 at 141550-p3199751. New integration fixtures used a noneditable title and
  an incident-specific envelope for session data; corrected. A broader metadata
  semantic-check experiment was removed to preserve its completed behavior.
- Rerun lifecycle_integration and metadata_ownership: PASS 3/3 at
  141704-p3202329. Four App integration cases cover StrictMode/session binding,
  demotion/return, hidden confirmation, account replacement, ignored abort, and
  neutral publication with retained metadata bases.
- make test-slice OWNER=module.workbook
  ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_8c99e88779,module.workbook.frontend_unit.verify_save_state_presentation_derives_one_prima_36510f67fb:
  PASS 3/3 at 141550-p3199762 (13 queue/save-state cases).
- make frontend-typecheck and make frontend-import-boundary-check: PASS 2/2 each
  at 141705-p3202617 and 141705-p3202621 after fixing new fixture typing.

No blocked dependency. Next: LC-04 real-service lifecycle replay, browser,
accessibility, visual and adjacent regression evidence.


### LC-04 progress

Added five routed browser journeys and lifecycle-specific accessibility/visual
rows. Mutations in these journeys reach the real service; interception drops or
delays their actual receipts or fails subsequent reads. No successful lifecycle
mutation is fabricated. The existing backend lifecycle integration now checks
another actor reopening before historical replay, current-role denial of receipt
recovery, exactly two durable audit/idempotency rows, and normalized reason bounds.

- make generate: PASS 142255-p3205408, 142724-p3279677, 142949-p3335061.
- make format: PASS 142320-p3208560, 142807-p3282936, 143026-p3338530.
- make service-backed-test-slice OWNER=module.incidents
  ROWS=module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20:
  PASS 3/3 at 142321-p3208703, including durable historical replay and 4096/4097.
- make service-backed-test-slice OWNER=module.incidents
  ROWS=module.incidents.browser.lifecycle_replay,module.incidents.browser.lifecycle_acknowledgement,module.incidents.browser.lifecycle_boundaries,module.incidents.browser.lifecycle_departure:
  PASS 11/11 at 142329-p3221350 (four real-service journeys).
- make service-backed-test-slice OWNER=web.design
  ROWS=web.design.accessibility.lifecycle: PASS 11/11 at 142819-p3287155.
  Keyboard Enter preserves multiline reasons; review/replay/refresh focus remains
  reachable at 1280x720, 1024x720, 768x640, 390x480, 200% zoom and text spacing.
- make frontend-typecheck: PASS 142331-p3221992 and 142821-p3287413.
- make lint-biome: 1/2 at 143043-p3343014; ten new test-only non-null assertions
  replaced with explicit fixture guards.
- Workbook browser row first run: 9/11 at 143041-p3342729. The source-text column
  was horizontally virtualized; the test now uses the existing grid target scroll
  helper before interacting. No lifecycle/write claim is inferred from that run.

Run suffixes use .cartulary/test-results/20260908T. Browser targets remain serialized.
Next: closed-workbook journey, ordinary visual reconciliation and image inspection,
then reviewed golden maintenance and adjacent browser regressions.

LC-04 continued evidence and decisions:

- The workbook row passed 11/11 at 143744-p3547223 after correcting the race
  fixture: dispatch a real source PATCH, hold transport, close on the service,
  then release the PATCH to receive incident_closed. An earlier attempt closed
  before admission and correctly observed no dispatch (143341-p3442899).
  The passing journey retains metadata work, permits real saved-view/preferences/
  membership writes while closed, and requires explicit discard followed by a
  fresh source edit after reopening.
- make frontend-typecheck and make lint-biome: PASS 2/2 each at
  143745-p3547498/3547508 and 144005-p3646004/3646014.
- The six lifecycle/metadata ownership, transport, integration, departure and
  lifetime rows listed above: PASS 7/7 at 143930-p3601138.
- make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e,web.workbook.regression.workbookz_mutation_runtime_common_coalescing_ba2a71e53d,web.workbook.regression.workbookz_mutation_runtime_surface_continuity_97808fe8ae:
  PASS 4/4 at 144112-p3651332.
- make test-slice OWNER=web.application
  ROWS=web.application.regression.incidentadminpanel_keeps_incident_summary_visibl_d7a6f59072,web.application.regression.incidentadminpanel_renders_null_and_view_schema_b2a93c36b5,web.application.regression.incidentadminpanel_renders_saved_view_workbook_p_8cc7f20d33:
  PASS 4/4 at 144113-p3651573.
- make test-slice OWNER=module.incidents
  ROWS=module.incidents.frontend.lifecycle_characterization,module.incidents.frontend.lifecycle_product_surface,module.incidents.unit.incident_lifecycle_request_and_openapi_contract_9466dfc2b1:
  PASS 4/4 at 144552-p3658508.

Ordinary visual lifecycle slices at 143209-p3391030 and 143514-p3495279 failed
only for eleven new missing goldens. The latter retained actual images as
Playwright attachments, decoded for human review under /tmp/cartulary-lifecycle-review.
All eleven were inspected: reason, review, narrow review, zoom, spacing, pending,
uncertain, confirmed-refresh-failure, closed, compact and comfortable. Text,
focus, scroll containment and recovery controls are legible at the declared
viewports. Reconciliation accounts for eleven capture intents with no ambiguous
mapping or unresolved registry identity. Unselected existing goldens are retained.

make browser-e2e-visual-update at 143929-p3599673 stopped before promotion.
A metadata fixture still expected initial loading even though workbook observation
now supplies an accepted resource; it correctly refreshes that resource on open.
The same fixture described closure at an already-used version. Corrected the
fixture to expect refresh and advance closure/reopen versions. No metadata editor
behavior was changed. The failed candidate also reported owner-only permissions
for a retained scratch PNG (artifact failure); no tracked golden was promoted.

Ordinary make service-backed-test-slice OWNER=web.design
ROWS=web.design.visual.metadata_editing at 144551-p3658269 completed the journey
and exposed four explained differences: metadata-loading (accepted fields during
refresh), metadata-closed (version 4), metadata-compact and metadata-comfortable
(version 5). Actual/loading and all four diffs were inspected. Capture viewports,
zoom, masks, scroll and scope remain unchanged. The catalog row owns these
nonregistry captures; no fixture registry expansion is required.

Final inspection tightened two necessary integrations: closed queue suspension
is distinct from authentication suspension, and each retained blocked unit stays
subject to explicit discard after reopening. Empty queues permit fresh work.
A rapid hide/return during authorization preflight cancels the unsent admission;
already dispatched outcomes keep their lifetime fence. Existing routed cases
cover these additions. The acknowledgement browser case now also fails follow-up
authorization after a real successful reopen. Pending outcome announcements sit
outside the busy review region.

make format at 144927-p3707091 caught an effect dependency declaration; the effect
now explicitly depends on the accepted identity object so unseen close/reopen
version advances are observed. make format passed 2/2 at 145651-p3763529.
make task-guide ROLE=module-author OWNER=module.auth passed before adjacent
administrative audit verification.

The public visual update passed 12/12 at 145412-p3712565. Its reconciliation
accounts for 204 active captures and 204 goldens, 26 registered fixtures, zero
orphans, missing goldens, ambiguous mappings or unresolved registered fixtures.
All eleven promoted lifecycle images and four intentionally changed metadata
images were inspected at their native viewport. Sixteen incidental rewrites of
existing goldens were restored byte-for-byte from the clean baseline; they are
not accepted changes. make generate passed at 150000-p3821026 to regenerate the
manifest for the bounded accepted set.

Accepted golden trigger: the new lifecycle workflow and the metadata fixture's
already-accepted-resource refresh / increasing lifecycle versions. Catalog rows:
web.design.visual.lifecycle (scenario_4c28f69d060d) and
web.design.visual.metadata_editing. Existing capture settings remain unchanged;
the lifecycle row adds explicit viewport/density/zoom/spacing captures using the
existing full-viewport harness, dynamic-value masks and renderer. These captures
are catalog-owned nonregistry scenarios; no stable registered fixture identity
is replaced. Changed filenames appear in the path inventory below.

- make test-slice OWNER=module.workbook with the queue/save-state rows above:
  PASS 3/3 at 145717-p3768060, including multiple retained units, empty queue,
  authentication separation and fresh action after explicit discard.
- make test-slice OWNER=web.application
  ROWS=web.application.regression.lifecycle_ownership,web.application.regression.lifecycle_transport,web.application.regression.lifecycle_integration:
  PASS 4/4 at 145717-p3768070, including unsent fresh/replay hide-return preflight.
- make frontend-typecheck and make lint-biome: PASS 2/2 each at
  145717-p3768233 and 145717-p3768286.
- make service-backed-test-slice OWNER=module.incidents
  ROWS=module.incidents.browser.incident_admins_manage_memberships_on_the_ordina_f26145a162,module.incidents.browser.lifecycle_acknowledgement,module.incidents.browser.lifecycle_boundaries,module.incidents.browser.lifecycle_departure,module.incidents.browser.lifecycle_replay,module.incidents.browser.lifecycle_workbook,module.incidents.browser.membership_audit_access_recovery,module.incidents.browser.membership_audit_read_recovery,module.incidents.browser.membership_audit_real_browsing,module.incidents.browser.membership_audit_response_inspection,module.incidents.browser.membership_management_departure,module.incidents.browser.membership_management_live_pages,module.incidents.browser.membership_management_mutation_recovery,module.incidents.browser.membership_management_self_actions,module.incidents.browser.metadata_authority,module.incidents.browser.metadata_departure,module.incidents.browser.metadata_live,module.incidents.browser.metadata_pending_departure,module.incidents.browser.metadata_recovery:
  PASS 19/19 graph units at 145815-p3770713. Nineteen Playwright cases passed,
  zero skipped, unexpected or flaky results across all five selected groups.

## Owner-to-change map

| Adopted owner | Change / implementation evidence |
| --- | --- |
| docs/spec/01_architecture_storage_and_view_contracts.md, REQ-01-585–591 | Single admitted immutable action; generated close/reopen client; explicit exact replay; independent receipt and latest resource; version-fenced App publication; real service audit/idempotency counts. |
| Core 01 REQ-01-496, required reason_note_v1 binding | Raw multiline editor; server normalization; normalized scalar boundary tests; narrow authored OpenAPI annotation and generated validator correction. |
| Core 01 lifecycle error registry | Typed validation/version/transition/transaction/auth failures and explicit recovery, with no automatic rebase or replacement. |
| docs/spec/03_workbook_interaction_collaboration_and_workflows.md, §4.3.2 / REQ-03-287–288 | Existing read-only interaction mode and persistent label; runtime closure invalidation; retained blocked edits and fresh action after reopening; allowed configuration and membership operations verified. |
| Core 03 keyboard ownership and docs/design.md | Native multiline reason, inline review, focused explicit controls, announcements outside busy review, inherited drawer/Escape ownership, density and viewport evidence. |
| docs/spec/04_security_deployment_and_conformance.md, applicable session/current-role/access rules | Existing App session/recovery port; incident/actor/lifetime fencing; current-admin checks; demotion retains memory-only local work; access/session replacement retires it. |
| docs/domain.md | Incident, lifecycle, Summary and preferences, Promoted fields, membership and read-only terminology; no archive/delete concept introduced. |
| Verification and harness owners | Authored catalog/source ownership and generated routing; real-service browser evidence; pinned-renderer visual maintenance; no Markdown runtime dependency. |

## Digest acceptance applicability

A001–A005, A012, A016–A017, A019, A022–A027 apply directly: adopted-owner map,
bounded seam, actual repository baseline, existing tokens/theme, Web Crypto,
explicit async/recovery states, accessible controls, reviewed visual captures,
semantic selectors, downstream generation and this handoff. A006 and A008–A009
apply to inherited density/responsive/overflow composition and were exercised at
the supported viewports plus narrow/short, zoom and text spacing. A013–A015 apply
only to the existing workbook recovery/keyboard/local-draft integration: rejected
source work is retained and not automatically replayed. A007, A010–A011, A018,
A020–A021 introduce no new create/inspector/evidence/component/virtualization
behavior; relevant existing owner regressions remain the baseline. No advisory
criterion authorizes a new product rule or a second design system.

## Compatibility, limitations and rollback

This is a frontend lifecycle workflow improvement. Public operations, routes,
request fields, authorization, lifecycle policy and stored data remain compatible;
no migration or dependency change is required. The reason projection now matches
existing normalized server admission; exact compatibility accounting explains the
mechanical OpenAPI constraint-removal classification. No alternate normalizer is
introduced. The prior metadata error-reason mismatch remains outside scope.

Recovery is memory-only. Explicit departure forgetting and session/access
retirement do not cancel or undo a server operation. A transport that ignores
abort keeps its dispatch lock until it actually settles; timeout alone does not
permit another dispatch. Acknowledgement survives failed subsequent observation.
Observed matching state is never claimed as proof of an uncertain action.

Rollback restores the coherent lifecycle feature, App/resource/departure/workbook
integration, authored projections/routing, generated derivatives, tests and
reviewed visual evidence together. It must not reverse legitimately committed
incident lifecycle state. No reset, commit, push or deployment was performed.

### LC-04 exit

The required lifecycle failure journeys, closed-workbook consequences, keyboard
and reviewed visual scenarios pass. Completed adjacent metadata, membership and
audit behavior is preserved by routed browser regressions.

- make service-backed-test-slice OWNER=module.auth
  ROWS=module.auth.browser.administrative_audit_read_recovery,module.auth.browser.administrative_audit_real_browsing:
  PASS 11/11 at 150027-p3824453.
- make test-slice OWNER=package.protocol_ts: PASS 7/7 at 150028-p3824691.
- make test-slice OWNER=package.ui: PASS 10/10 at 150029-p3825325.

No blocked dependency. Next: LC-05 agent-finalize, broader final verification,
two fresh ordinary visual passes against the bounded manifest, and final scope
audit. RESULTS_DIR will remain unset: no qualifying successful full warm check
from the exact final source exists.

## Changed path inventory

This exact inventory includes authored source, tests, projections, Make-generated
derivatives, accepted goldens and this handoff. No unrelated path is included.

- `apps/web/e2e/incident-administration.spec.ts`
- `apps/web/e2e/incident-lifecycle.spec.ts`
- `apps/web/e2e/incident-metadata-editing.spec.ts`
- `apps/web/e2e/support/incidentLifecycle.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-closed-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-comfortable-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-compact-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-confirmed-refresh-failure-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-pending-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-reason-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-review-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-review-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-review-spacing-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-review-zoom-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/lifecycle-uncertain-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-closed-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-comfortable-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-compact-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-loading-linux.png`
- `apps/web/src/app/App.tsx`
- `apps/web/src/app/IncidentAdminPanel.test.tsx`
- `apps/web/src/app/IncidentAdminPanel.tsx`
- `apps/web/src/app/IncidentLifecyclePanel.tsx`
- `apps/web/src/app/api/incidentLifecycleClient.test.ts`
- `apps/web/src/app/api/incidentLifecycleClient.ts`
- `apps/web/src/app/api/incidentMetadataContracts.ts`
- `apps/web/src/app/api/incidentResourceContracts.ts`
- `apps/web/src/app/appDepartureReview.test.ts`
- `apps/web/src/app/appDepartureReview.ts`
- `apps/web/src/app/incidentLifecycleCharacterization.test.tsx`
- `apps/web/src/app/incidentLifecycleController.test.ts`
- `apps/web/src/app/incidentLifecycleController.ts`
- `apps/web/src/app/incidentLifecycleIntegration.test.tsx`
- `apps/web/src/app/incidentLifecycleModel.ts`
- `apps/web/src/app/incidentLifecycleTestSurface.tsx`
- `apps/web/src/app/incidentMetadataController.ts`
- `apps/web/src/app/incidentResourceController.ts`
- `apps/web/src/app/useIncidentLifecycle.ts`
- `apps/web/src/app/useIncidentMetadata.ts`
- `apps/web/src/shared/incidentResource.ts`
- `apps/web/src/shared/workbookShellContracts.ts`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/adapters/createWorkbookIncidentAdapter.ts`
- `apps/web/src/workbook/hooks/useWorkbookIncidentIdentity.ts`
- `apps/web/src/workbook/ports/WorkbookIncidentPort.ts`
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`
- `apps/web/src/workbook/utils/workbookPendingQueue.test.ts`
- `apps/web/src/workbook/utils/workbookPendingQueue.ts`
- `contracts/openapi-releases/2.0.0.change-set.json`
- `contracts/openapi-source/owners/module.incidents/openapi.json`
- `contracts/openapi/cartulary.openapi.yaml`
- `docs/handoffs/incident-lifecycle-controls-refactor-handoff.md`
- `internal/gen/contractopenapi/artifacts_gen.go`
- `internal/gen/openapioperations/catalog_gen.go`
- `internal/modules/incidents/httpapi/lifecycle_contract_test.go`
- `internal/modules/incidents/lifecycle_integration_test.go`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `packages/protocol-ts/src/generated/core-http-validators.ts`
- `packages/ui-contracts/src/applicationSelectors.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.incidents.json`
- `tools/test_families/web.application.json`
- `tools/test_families/web.design.json`

## LC-05 final verification

make agent-finalize passed 1/1 at 150158-p3880013 before broader verification.
Its unit-artifacts/finalize-summary.json records zero generated updates, passing
JSON shape, catalog/tier coverage and generated-structure checks. RESULTS_DIR was
unset. Retained-run selection, canonical evidence, scheduler timing/order and
performance-evidence maintenance were skipped with results-dir-not-provided;
no retained successful full warm run is claimed.

- make frontend-typecheck: PASS 2/2 at 150231-p3883695.
- make frontend-import-boundary-check: PASS 2/2 at 150231-p3883737.
- make lint: PASS 11/11 at 150231-p3883844, including authored source and Markdown.
- make generate-drift: PASS 4/4 at 150419-p3973440.
- make generated-artifact-policy-check: PASS 3/3 at 150419-p3973407.
- make openapi-compatibility-check: PASS 4/4 at 150419-p3973480.
- make test-catalog-check: PASS (exit 0) in the same final verification batch.

Final inspection confirms no specification, digest, dependency, migration or
storage change, no direct grid-vendor import, and no executable Markdown
reference in the lifecycle feature/tests or new projection logic. The exact path
inventory accounts for all 68 changed/new paths. Full-repository check/CI/release
and unrelated backend suites are not claimed; validation follows the affected
owner slices plus the required broader frontend/design checks.

- make frontend-unit: PASS 487/487 at 150231-p3883683.

- First fresh ordinary make browser-e2e-visual: PASS 12/12 at
  150231-p3883794 against the bounded promoted manifest.

A final queue audit found that an intervening authorization pause or a late
retryable transport result could remove the existing closed-work discard anchor.
The same closure-description helper now runs when authentication recovers or a
settled unit returns to the queue. Two assertions in the existing routed queue
case cover this; no new recovery policy or UI is added. make format passed 2/2
at 150737-p3999795. Required affected checks and finalization are repeated below;
this follow-up does not change the reviewed golden manifest.

- Queue/save-state owner rows: PASS 3/3 at 150759-p4004076.
- Workbook shell/runtime owner rows: PASS 4/4 at 150800-p4004301.
- Repeated make agent-finalize: PASS 1/1 at 150802-p4004697; generated files
  unchanged, RESULTS_DIR still unset and retained-run maintenance still skipped.
- make frontend-typecheck and make lint-biome: PASS 2/2 each at
  150837-p4008911 and 150837-p4008928.
- make service-backed-test-slice OWNER=module.incidents
  ROWS=module.incidents.browser.lifecycle_workbook: PASS 11/11 at
  150837-p4008784 on the final queue source.

The first ordinary visual run retained 37 passing Playwright cases with zero
skips, unexpected or flaky outcomes. Its reconciliation again accounts for all
204 captures/goldens and 26 registered fixtures with no missing, orphan or
ambiguous entries. Branch and HEAD remain main / 58d3b8c5e1b96b44b9aa938a2216277bd8905894.

- Interim make lint-markdown: PASS at 151116-p4106342. It will be rerun after
  marking the completed LC-05 handoff DONE.

- make browser-e2e-a11y: PASS 12/12 at 151022-p4057835. All 32 Playwright cases
  passed, zero skipped, unexpected or flaky, including the lifecycle keyboard,
  focus, multiline Enter, uncertain replay and responsive/zoom/spacing journey.

## Required evidence coverage

| Requirement family | Evidence and scope of claim |
| --- | --- |
| Single admission, deadline, ignored abort | Routed lifecycle controller cases cover same-tick activation, preflight, actual transport lock after 30-second observation timeout and late settlement. No claim that abort cancels a server action. |
| Exact uncertain replay and divergent key | Real browser commits a close, drops its receipt, sends identical captured replay after another actor reopens, and observes active version 3. The backend integration checks exactly two durable lifecycle audit/idempotency records. Divergent-key reuse receives real client_txn_conflict. |
| Distinct rejection recovery | Typed client/controller and characterization cases distinguish validation, incident_version_conflict, illegal_transition and client_txn_conflict. Real service cases exercise each; rejected attempts never auto-rebase or manufacture a replacement action. |
| Acknowledgement versus observation | Real successful close/reopen receipts survive failed resource and authorization observations; the browser offers read recovery and no replay control. Historical receipt publication is separately fenced in controller and neutral-resource integration cases. |
| Identity, version and lifetime races | Client semantic validation rejects mismatched identity, invalid version/status/closed_at and wrong action receipts. Controller/App binding cases fence incident, account, role, session, draft revision and surface changes, including StrictMode and ignored abort. |
| Current authorization | Real lifecycle integration denies nonmember deployment-admin access and demoted original-actor replay, while a second current incident admin can reopen. Binding cases retain local work on demotion and retire it on session/access loss. |
| Concurrent local work and departure | Four App departure owners are exercised sequentially in unit tests, including final recheck and cancellation. The real browser retains membership, metadata and lifecycle work together, reviews each, stops on Escape and later completes explicit history departure. |
| Closed workbook | Real source PATCH is rejected with incident_closed after closure; persistent read-only mode remains distinct from save status. Saved views, both preferences and membership administration succeed while closed. Retained metadata and source drafts remain visible; reopening sends no retained PATCH, explicit discard and a fresh edit produce the next write. |
| Reason boundaries | Contract/integration and browser evidence covers required reason, whitespace, decomposed normalization, supplementary Unicode, normalized 4096 acceptance and 4097 rejection. Editing preserves raw text and uses no local scalar normalizer or raw maxlength approximation. |
| Adjacent behavior and design | Nineteen incident browser cases plus administrative audit, full frontend units, 32 accessibility cases, reviewed 15-image change set and ordinary visual reconciliation preserve the existing summary/preferences, metadata, membership and audit seams. |

### LC-05 exit

Second fresh ordinary make browser-e2e-visual: PASS 12/12 at
151252-p4108969. Both ordinary passes used the same bounded golden manifest and
pinned renderer. The final run again has 37 passing Playwright cases, no skips,
unexpected or flaky results, and complete 204-golden / 26-registered-fixture
reconciliation with no missing, orphan or ambiguous entries.

All workstreams are DONE. Required behavior, recovery, authorization, local-work,
closed-workbook, accessibility, visual, adjacent regression and generated-boundary
evidence passed. The final source scope is the 68 paths inventoried above. No
adopted-owner contradiction or required product-policy expansion arose.

After marking LC-05 DONE, make lint-markdown passed at 151648-p4158815;
git diff --check and the completed-handoff path/owner scope audit passed. The
audit confirms all five rows DONE, exact 68-path coverage, protected owners and
dependencies unchanged, and next action none. The same checks are rerun after
recording this result. Retained-run maintenance remains skipped for the explicitly
recorded RESULTS_DIR reason. No additional product work remains.

Next action: **none.**
