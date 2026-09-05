# Incident-directory creation refactor handoff

## Baseline and scope

- Branch: `main`; HEAD: `3972662b1ca2f994eec4ae453cc3f231de4cb6e1`.
- Working tree clean at implementation start; no reset, commit, or push.
- One seam: directory creation through `POST /api/v1/incidents`.
- Allowed: creation model/controller, presentation, narrow typed client, App
  integration, related styles/selectors/tests/routing/visual evidence, this handoff.
- Excluded: backend, public contracts, authorization, routes, dependencies,
  persistence, directory query redesign, account/admin forms, adopted owners,
  digest sources, and completed workbook/grid-row/menu/inspector work.

## Tracker

| Workstream | Status | Exit evidence / next action |
| --- | --- | --- |
| IC-01 — Baseline, authority, characterization | DONE | Five focused characterization failures reproduce keyboard, duplicate, local-error, replay, and navigation gaps. |
| IC-02 — Creation model and typed operation lifecycle | DONE | Fourteen focused operation/client tests and frontend typecheck pass. |
| IC-03 — Same-surface form and application handoff | DONE | Inline form, nine App lifecycle cases, baseline metadata/bootstrap checks, typecheck/import boundaries, and real browser creation pass. |
| IC-04 — Browser, accessibility, and visual evidence | DONE | Four creation/browser rows, full a11y, reviewed eight goldens, and two final-source ordinary visual passes. |
| IC-05 — Final validation and completed handoff | DONE | A001–A027 dispositions, compatibility/atomic rollback, final Make checks, Markdown lint and 32-path scope audit complete. |

## Authority and inspected evidence

| Behavior | Owner |
| --- | --- |
| Root visibility and ordinary workbook selection | Core 01 §3.3.2.1A, REQ-01-580; Core 03 §2.4 |
| Active authenticated creator; cookie CSRF | Core 01 REQ-01-154–155; Core 04 REQ-04-006, 008–012, 015–016, 023 |
| Resource and closed create payload | Core 01 §3.3.5.3, REQ-01-152–159 |
| Normalization, errors, actor-scoped replay | Core 01 REQ-01-157, 160–162, 166–167; §3.3.6.2 |
| Atomic incident/admin/preferences bootstrap | Core 01 REQ-01-163–165 |
| Optional string semantics and canonical TLP | Core 01 §18, REQ-01-488, 491, 491.1; Core 02 §18, REQ-02-222/254 |
| Local feedback, keyboard/focus, dialog restrictions | Design §§5, 8.4–8.5, 10.8, 12.4–12.8, 14 |
| Responsive composition and visual evidence | Design §§7, 15; visual golden maintenance guide |

Read root AGENTS.md (no nested instructions), digest README and localized read
order including maps/rules/acceptance, domain vocabulary boundary, relevant Core
and design owners, and completed account-application-menu handoff as evidence.
No adopted-owner contradiction found. Digest guidance is advisory; executable
code and evidence do not consume Markdown.

Inspected IncidentLanding, App, landingAdminTypes/Styles, LandingAdminDisplay,
appShellClient, browserApi/httpTransport, generated createIncident binding and
request/response types, shared authorization/registered focus, App.landing tests,
IncidentDirectory browser helper, auth-and-incident-directory browser checks,
accessibility/visual sources, and authored owner catalogs.

Current stack revalidated: React 19, TypeScript 6, Vite 8, pnpm workspace,
Vitest/Testing Library, Playwright, existing tokens and Lucide. Generated roots
remain governed by tools/generated_artifact_policy.json.

## Interaction decisions

| State / action | Draft and operation | Focus and navigation |
| --- | --- | --- |
| Open / edit | Required key/title; optional existing disclosure | Explicit open focuses key; ordinary Tab order. |
| Local or definitive field rejection | Retain exact draft; reveal affected optional field | Local associated errors; no response-driven focus move. |
| Submit | Capture declared payload and Web Crypto ID once; synchronous dispatch lock | Native submit; textarea Enter remains editing. |
| Pending / uncertain | Read-only captured request; manual replay uses same ID/body | Account navigation stays reachable. |
| Close / Escape | Collapse, retain memory-only draft/attempt, expose resume status | Return to trigger; revoke automatic handoff. |
| Reopen | Resume existing state, not a new attempt | Focus relevant error/replay/open action; reopening does not restore automatic handoff. |
| Navigate away / Account settings | Retain same-session controller; revoke automatic handoff | Later responses do not replace chosen destination. |
| Confirmed create / handoff failure | Creation stays committed; retry opens only | Ordinary incident_id startup; no explicit sheet. |
| Session termination / account replacement | Clear account-scoped draft/attempt and invalidate callbacks | No stale focus or navigation. |

User selected collapse-and-retain and replay-before-editing. No persistent
recovery, background replay, global mutation framework, or workbook queue reuse.

## Verification routing and initial evidence

Resolved `make help`, `make help-all`, and `make task-guide ROLE=module-author`
for `module.incidents`, `web.application`, and `web.design`.

- Incident creation behavior: module.incidents frontend/browser rows; App and
  narrow integration regressions retain web.application routing.
- Presentation/accessibility/visual readiness: web.design; selectors: package.ui.
- Optional metadata baseline passed: `20260905T143008Z-p1009`.
- App creation/membership/incident_id baseline passed: `20260905T143008Z-p1005`.
- Service-backed directory creation/admin/workbook baseline passed:
  `20260905T143029Z-p2102`.

These are exact-source baseline checks, not evidence that remediation is done.
Previous menu handoff failures are historical evidence only; this seam's three
selected baseline checks pass at current HEAD.

## IC-01 activity

Initial inspection finds a non-form modal, seven App draft variables, component
open state, raw fetch/success cast, no synchronous dispatch guard, a new ID per
call, global creation errors, and unguarded response navigation. Characterization
tests will distinguish reproduced gaps from these source observations.

Characterization command: `make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.incident_creation_lifecycle_keyboard_replay_and_e99c1bedb1`.
Final baseline run: `20260905T150136Z-p55051`, expected product assertion failure.
Native Enter sent zero requests; same-tick activation sent two; required errors
had no accessible association; uncertain recovery had no replay action; delayed
success changed the route despite chosen Account settings. The last test now
waits for response parsing and checks the published URL, removing an initially
premature assertion hidden by lazy workbook rendering. No production
source changed before these failures were captured. Exact-title/sort catalog
authoring errors were corrected before the retained test runs.

IC-01 exit: owner/scope/state matrix, successful baseline, reproduced gaps, and
owner routes recorded. Next action: implement and test the bounded controller.

## IC-02 activity and exit

Added incidentCreationModel.ts and focused tests; added the narrow generated
createIncident operation wrapper to appShellClient. The controller captures one
immutable allowed request and ID before dispatch, owns a synchronous observation
guard and 30-second timeout, preserves uncertain replay, and separates confirmed
creation from membership/workbook handoff. Session and navigation generations
guard callbacks; no queue or persistent storage is used.

`make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.incident_creation_operation_payload_outcomes_and_2ecdf8c632`
passes all 14 cases at `20260905T151131Z-p59594`. `make frontend-typecheck`
passes at `20260905T151047Z-p59028`. Its earlier failure at
`20260905T150819Z-p57928` caught exact-optional-property typing and one unused
import; both were fixed. Generated output remains untouched. Next action:
wire the controller and same-surface form to App.

## IC-03 activity and exit

IncidentCreationForm and useIncidentCreation now replace the creation modal,
seven App draft variables, and global creation notices. Native Enter submission,
textarea editing, local associated errors, optional-error disclosure, explicit
focus return, retained close/reopen, and read-only replay use the controller.
App owns current-session replacement, membership refresh, and incident_id startup;
navigation and Account settings immediately invalidate automatic handoff.
Confirmed handoff observation is also bounded, without another create request.

The generated boundary correctly rejected incomplete historical test fixtures;
appShellTestSupport now supplies complete attribution/lifecycle resource fields.
Form errors were moved outside labels so error text does not change control names.
An import-boundary failure (`20260905T152323Z-p73648`) caught a direct protocol
import and a type cycle; model types now use publicHttpTypes and presentation
bindings live in the model, preserving the existing adapter boundary.

Passing evidence:

- Nine App lifecycle cases plus existing create/bootstrap row:
  `20260905T152202Z-p72206`.
- Existing optional metadata row: `20260905T152349Z-p74108`.
- Selector contract: `20260905T152354Z-p74642`.
- Typecheck: `20260905T152312Z-p73156`.
- Import boundaries: `20260905T152442Z-p75447`.
- Format: `20260905T152446Z-p75876`; only seam files were rewritten.
- Existing real-browser directory create/admin/workbook row:
  `20260905T152518Z-p80212`.

Next action: add deterministic browser recovery, accessibility, geometry, and
visual capture evidence; retain completed-seam goldens unless a reviewed change
is directly caused by this creation seam.

## IC-04 activity

Added three service-backed browser scenarios covering native keyboard creation,
real optional/key rejection, exact payload and server normalization, duplicate
activation, server commit followed by a lost client response, identical 200 replay,
confirmed creation followed by membership-refresh failure, opening-only retry,
and delayed success after Account settings navigation. All pass at
`20260905T153821Z-p37684` via the three new module.incidents browser rows.
The existing IncidentDirectory helper now opens/resumes the inline form and only
starts another draft explicitly after a confirmed result.

The focused web.design accessibility row passes at
`20260905T154530Z-p35302`. It checks keyboard/name/error associations, optional
error disclosure, restrained announcements, pending/recovery reachability at
1280×720, 1024×720, 768×640, and 640×480, 200% CSS zoom, text spacing, and safe
account navigation. Its first run (`20260905T154024Z-p83887`) failed at a test
focus operation after simulated zoom: focusing an already-active control did not
scroll it back into view. The scenario now uses actual Shift+Tab/Tab to re-enter
that control; all visibility and overflow assertions remain. Directory
aria-busy now belongs to the independent query region, so query refresh cannot
defer creation announcements.

Eight deterministic visual captures cover initial, required error, expanded
details, pending, short recovery, zoom, text spacing, and confirmed handoff
failure. Ordinary reconciliation and image review precede promotion. Routing
inputs are authored in module.incidents.json and web.design.json; `make generate`
passed at `20260905T153724Z-p34459`, updating only browser batch and topology
render projections. No stable digest or registry fixture IDs were invented.

`make agent-finalize` passed before broad visual verification at
`20260905T155150Z-p82048`; generated maintenance was unchanged. RESULTS_DIR was
unset because no qualifying exact-source full warm check exists. Canonical
retained-run evidence, scheduler ordering/timing, and performance maintenance
were therefore skipped, as recorded in unit-artifacts/finalize-summary.json.

First full ordinary visual run `20260905T155217Z-p85214`: all existing visual
tests passed and the new scenario reached all eight captures. It failed on the
eight expected missing initial goldens, plus reconciliation rejected the new
authored visual scenario ID because that artifact requires a hexadecimal ID.
Corrected only the new visual routing input and regenerated through Make
(`20260905T155608Z-p33293`). This was a fixture/routing defect, not a product or
unrelated baseline failure; no images were promoted by the failed run.

Ordinary run `20260905T155640Z-p36382` now produces the required v2
reconciliation: 110 capture intents, 102 active existing goldens, eight new
missing goldens, 26 resolved registered fixtures, zero orphans and zero ambiguous
mappings. All functional assertions and existing comparisons pass. The accepted
trigger is the authorized same-surface form and operation-state presentation;
the canonical updater adds eight initial goldens, without moving, deleting, or
re-accounting existing captures. The new row is
`web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde`,
scenario `scenario_46ea092e7f2b`, project chromium. It is an active nonregistry
capture with no invented D-VFIX identity. Existing renderer, fonts, masks,
registered fixture IDs, and screenshot scope remain unchanged.

Update attempt `20260905T155940Z-p83937` did not promote files. Its creation
scenario passed and produced all eight candidates; an unchanged menu fixture
timed out waiting for two filtered rows (31 remained), and the retained-artifact
scanner rejected a copied existing PNG's permissions. The menu fixture passed
both preceding ordinary runs; it remains unchanged. The permission failure is
also documented in the completed menu handoff. Restricted all 102 source PNGs to
0600 with SHA-256 equality verified, then reran the canonical updater; no harness
check was bypassed and no tracked file content changed for this maintenance.

Individually reviewed all eight creation candidate images from that run:

| Filename under workbook.visual.spec.ts-snapshots | Viewport / zoom | Review |
| --- | --- | --- |
| incident-create-initial-linux.png | 1280×720 / 100% | Compact inline form, key focus visible, account trigger reachable. |
| incident-create-required-errors-linux.png | 1280×720 / 100% | Associated errors beneath fields; single operation message. |
| incident-create-expanded-details-linux.png | 1024×720 / 100% | All optional fields and action fit; long inputs retain native text scrolling. |
| incident-create-pending-linux.png | 768×640 / 100% | Pending label, close semantics, and read-only fields remain visible. |
| incident-create-recovery-short-linux.png | 640×480 / 100% | Recovery text and focused retry fit above directory controls. |
| incident-create-recovery-zoom-linux.png | 1280×720 / 200% CSS zoom | Native focus scroll reveals recovery and retry; surrounding content remains scrollable. |
| incident-create-recovery-text-spacing-linux.png | 768×640 / 100% | Expanded spacing wraps feedback; retry remains contained and focused. |
| incident-create-confirmed-handoff-failure-linux.png | 1280×720 / 100% | Committed outcome is explicit with opening-only retry. |

Decision for each: accept, classified as accepted intentional state under Design
§§7, 12.4–12.8, 14–15. These are implementation/design evidence, not Core 05
claim-publication or release-readiness evidence. No unrelated golden change is
accepted. Two fresh ordinary passes remain required after promotion.

The second update (`20260905T160444Z-p32056`) passed private-artifact checks but
reproduced the same existing menu filter assertion. Its trace contains only
unfiltered directory requests: initial refresh completes after the fixture fills
search and before debounced search dispatch. The directory query/touched-state
implementation is unchanged by this seam. Added an explicit native Enter and
response barrier to that fixture's existing filter step, retaining its two-row
assertion, inputs, menu behavior, and all screenshot settings. This is a bounded
fixture synchronization repair needed by the shared visual gate, not a menu or
directory-query refactor. No failed update promoted tracked evidence.

Canonical refresh passes at `20260905T160935Z-p83842` (12/12 units). All eight
creation goldens were reviewed; six are byte-identical to the earlier individual
review and the two with small border-rasterization differences were inspected
again. The updater also changed four unrelated existing images (presence marker
ordering, entity chip edge, network inspector input edge, and pending replay
toolbar). Their old/new images were inspected and rejected from this seam;
restored their exact HEAD bytes. All 102 pre-existing PNGs now match HEAD by
byte comparison, including every menu golden. `make generate` regenerated the
manifest at `20260905T161402Z-p32475`. No renderer, dependency, threshold, or
registered fixture changed.

Ordinary visual pass `20260905T161437Z-p38568` succeeds (12/12 units, 110
captures). Final-source typecheck (`20260905T161507Z-p81324`), import boundaries
(`20260905T161507Z-p81341`), and Biome (`20260905T161507Z-p81367`) pass. Final
accessibility review then removed aria-hidden from the visible pending suffix
and added an accessible-name assertion: sighted and screen-reader users now
receive the same pending button label. No pixels changed, but two fresh ordinary
visual runs will still be retained after this final presentation-source edit.

After that edit, `make agent-finalize` passes again at
`20260905T161807Z-p91893` with RESULTS_DIR unset and the same documented
retained-run skips. Full `make browser-e2e-a11y` passes at
`20260905T161851Z-p95175` (12/12 units), including the new pending-name assertion.
Creation/controller/bootstrap frontend rows pass at `20260905T161912Z-p38572`;
the complete web.application owner slice passes at `20260905T161912Z-p38580`
(58/58 units), covering unchanged menu/session/route integration; selector
contracts pass at `20260905T161912Z-p38593`.

Two fresh ordinary visual passes on the final presentation source:
`20260905T162101Z-p51590` and `20260905T162445Z-p98917`, each 12/12 units.
Both reconcile all 110 active goldens and all 26 registered fixtures, with zero
missing, orphaned, ambiguous, or unresolved fixture mappings. Final service-backed
verification passes at `20260905T162914Z-p46655` (11/11 units): the three new
creation scenarios plus the unchanged ordinary create/admin/bootstrap scenario.

IC-04 exit: keyboard, actual public validation, committed-response loss and 200
replay, no duplicate recreation, chosen navigation, full accessibility, and
reviewed constrained visual evidence pass. No blocked owner dependency remains.
Next action: finish IC-05 acceptance, compatibility, rollback, and final audits.

## IC-05 architecture and changed paths

Before: modal ceremony, click-only dispatch, scattered draft/open state, global
creation errors, unchecked success, and fresh transaction IDs after ambiguous
failure. After: one native inline form renders one account-scoped controller;
the controller owns draft/disclosure, explicit outcome state, captured request,
synchronous dispatch lock, observation timeout, and lifetime guards. The typed
adapter accepts the generated public request/response boundary. App continues to
own session integration, directory queries/pagination, and ordinary incident_id
workbook startup. Confirmed creation is recorded before handoff; failed handoff
cannot re-enter create dispatch.

| Paths | Purpose |
| --- | --- |
| apps/web/src/app/incidentCreationModel.ts; incidentCreationModel.test.ts | Operation owner, payload/replay/lifetime tests. |
| apps/web/src/app/useIncidentCreation.ts; IncidentCreationForm.tsx | React subscription and native same-surface presentation/focus. |
| apps/web/src/app/App.tsx; IncidentLanding.tsx; landingAdminTypes.ts | Session/route handoff and removal of modal/scattered draft wiring. |
| apps/web/src/app/api/appShellClient.ts; publicHttpTypes.ts | Existing generated createIncident binding, exact request allowlist and validated response. |
| apps/web/src/app/App.landing.test.tsx; apps/web/src/testing/appShellTestSupport.ts | App lifecycle regressions and complete public incident fixtures. |
| packages/ui-contracts/src/applicationSelectors.ts; application-selectors.test.ts | Stable creation form/status selectors. |
| apps/web/e2e/auth-and-incident-directory.spec.ts; pages/incidentDirectory.ts; support/incidents/creation.ts | Real creation/replay/bootstrap and deterministic response/presentation helpers. |
| apps/web/e2e/workbook.a11y.spec.ts; workbook.visual.spec.ts | Keyboard/geometry/accessibility, eight visual states, existing filter fixture synchronization. |
| apps/web/e2e/workbook.visual.spec.ts-snapshots/incident-create-*.png | Eight individually reviewed new goldens listed above. |
| tools/test_families/module.incidents.json; web.design.json | Authored creation and design verification routing. |
| tools/browser_e2e_batch_manifest.json; execution_topology_render_index.json; frontend_visual_golden_manifest.json | Make-generated projections and reviewed golden manifest. |
| docs/handoffs/incident-directory-creation-refactor-handoff.md | This implementation/support record. |

No local user, membership, preference, saved view, initial sheet, server default,
capability discovery, or new public route is synthesized by this workflow. The
existing server transaction remains the sole creator/admin bootstrap owner.
Session identity uses the public actor/authentication/provider values plus local
invalidation generations; membership refresh within the same authentication
lifetime retains the draft. The 30-second limit bounds client observation and
does not assert server cancellation or rollback.

## Acceptance dispositions

These A001–A027 IDs are advisory review routing. Core/design owners above govern
their applicable scope; incident/workspace creation is distinct from completed
view-scoped grid-row creation.

| ID | Disposition and evidence |
| --- | --- |
| A001 | PASS: owner-to-change map above; no advisory source becomes authority. |
| A002 | PASS: one incident-directory creation seam; fixture synchronization and generated evidence are explicitly identified. |
| A003 | PASS: actual branch/HEAD, initially clean state, stack, owners, generated policy, initial files and routing inspected. No new direct grid imports. |
| A004 | PASS: existing landing styles and tokens; no second token, theme, or density registry. |
| A005 | PRESERVED: existing dark_graphite appearance; no new theme or palette. |
| A006 | PRESERVED BASELINE: shared workbook density/editor geometry untouched; full application/a11y/visual regressions pass. |
| A007 | NOT AN INCIDENT-CREATE REQUIREMENT: create_capable, inline_create and fields[].create_writable belong to view-scoped row creation. That completed baseline is unchanged; Core 01 REQ-01-154–159 instead governs this payload. |
| A008 | PASS IN SCOPE: four viewport sizes, short height, zoom, and text-spacing form reachability pass. Shell/inspector breakpoint and clamp owners remain unchanged. |
| A009 | PASS IN SCOPE: native directory scrolling, no fixed-height nested form scroller or horizontal document overflow; safe account navigation remains reachable. Workbook scroll ownership is unchanged. |
| A010 | PRESERVED BASELINE: inspector registry/routes/features untouched; existing visual and accessibility checks pass. |
| A011 | PASS IN SCOPE: explicit opening/return focus, no background error focus jump, and chosen-route preservation pass. Workbook continuity remains its existing owner's responsibility. |
| A012 | PASS: Web Crypto IDs, synchronous duplicate guard, immutable unresolved request, same-ID replay, explicit new attempt after definitive transaction rejection. |
| A013 | NOT AN INCIDENT-CREATE REQUIREMENT: workbook blocked-edit queue/new-ID retry/discard remains unchanged. This workflow uses actor-scoped incident replay and no workbook/global queue. |
| A014 | PASS IN SCOPE: native Enter, multiline description, ordinary Tab, Escape, draft retention, read-only uncertainty, explicit reopen. Grid click/paste/edit behavior remains unchanged. |
| A015 | PASS IN SCOPE: key conflict and recognized public field errors are beside their controls; operation feedback is local. Workbook cell-conflict presentation is unchanged. |
| A016 | PASS: editing/submitting/rejected/uncertain/created and handoff failure are distinct; independent directory refresh/loading is preserved. |
| A017 | PRESERVED: authorized directory rows and query/pagination ownership remain outside the controller; App query regression slice passes. Initial-refresh/debounce fixture limitation is recorded separately. |
| A018 | PRESERVED BASELINE: evidence lifecycle/overlay/preview combinations unchanged; full visual/a11y checks pass. |
| A019 | PASS: real keyboard assertions, names/error associations, visible focus, contrast, restrained live regions, reduced motion, zoom and text-spacing evidence; full a11y gate passes. |
| A020 | PASS IN SCOPE: deterministic form editing/error/pending/recovery/confirmed states use existing component styles. Existing component-matrix evidence remains unchanged. |
| A021 | PRESERVED BASELINE: no workbook virtualization changes or new performance claim. |
| A022 | PASS: 26 existing registered fixtures and all 110 capture intents reconcile; eight new nonregistry captures use exact authored row/scenario/project identities and reviewed settings. Two fresh ordinary passes succeed. |
| A023 | PASS: semantic native controls plus shared stable form/status selectors; package.ui selector tests pass. |
| A024 | PASS: executable code/tests/generation consume no Markdown/digest paths, contents, hashes, or structure. Markdown is human review/support only. |
| A025 | PASS: authored catalog inputs feed Make generation; generated protocol/backend/UI roots untouched. Drift, JSON shape, catalog, and generated-policy checks cover projections. |
| A026 | PASS: existing API and authorization boundaries; no route/schema/storage/lifecycle migration or new incident defaults. Client interaction states are local, not wire/stored contracts. |
| A027 | PASS: owners, architecture, changed paths, behavior, evidence, failures, limitations, compatibility and rollback are recorded here. No next seam is authorized or started. |

## Compatibility, rollback, and limitations

Compatibility: frontend interaction remediation over the existing typed API;
no stored-data migration and no backend, API/schema, authorization, route,
dependency, lockfile, persistent-storage, adopted-spec, or digest change.
Both 201 creation and 200 actor-scoped replay keep the existing atomic bootstrap
and ordinary workbook launch policy. The destination workbook owns its own
loading/error/focus behavior after route handoff.

Rollback must restore the controller/hook/form, App and IncidentLanding wiring,
typed adapter/type exports, selectors, test fixtures, authored routing inputs,
generated projections, eight new goldens and their manifest as one unit. Restore
the baseline interaction and evidence together; do not leave new replay tests
bound to the old dispatch behavior or a manifest pointing at removed images.
There is no data rollback or server migration. This handoff remains the audit
record of that compatibility boundary.

Drafts and recovery are memory-only: close/reopen and same-session application
navigation retain them; session termination, account/authentication replacement,
controller disposal, or reloading/closing the tab clears them. Observation abort
never proves the server cancelled creation. No background replay, unload promise,
or persistent recovery was added. A new attempt cannot silently replace an
unresolved one. Initial directory refresh can suppress a concurrently typed
debounced query under unchanged query code; the visual fixture explicitly
submits its filter, and no directory query redesign is included.

The previously completed menu handoff's broader browser failures were historical
context, not assumed current failures. This seam's selected baseline and final
creation tests pass. Actual failures encountered here are recorded above,
including characterization, fixture/type/import authoring fixes, initial golden
absence, unchanged filter synchronization, and private-artifact permissions.
No test expectation or screenshot threshold was weakened.

Full release/check, backend-wide suites, measurement gates, and owner evidence
publication audit were not selected for this frontend seam. No qualifying full
warm check/support evidence packet exists for the exact source, so RESULTS_DIR
remains unset; retained canonical-run, scheduler ordering/timing, and performance
maintenance are skipped by agent-finalize. These UI artifacts make no release,
Core 05, or new performance claim.

## Final commands and evidence

All repository verification ran through public Make targets from the root.
Focused creation selections used:

```bash
make test-slice OWNER=module.incidents \
  ROWS=module.incidents.frontend.incident_creation_lifecycle_keyboard_replay_and_e99c1bedb1,module.incidents.frontend.incident_creation_operation_payload_outcomes_and_2ecdf8c632,module.incidents.frontend.the_ordinary_landing_shell_creates_an_incident_r_b64ef5986b
make service-backed-test-slice OWNER=module.incidents \
  ROWS=module.incidents.browser.incident_creation_native_keyboard_payload_and_fi_b6b4631f5f,module.incidents.browser.incident_creation_uncertain_replay_and_committed_985931d51a,module.incidents.browser.incident_creation_preserves_chosen_navigation_af_24cb9bd4de,module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5
make service-backed-test-slice OWNER=web.design \
  ROWS=web.design.accessibility.incident_creation_accessible_keyboard_feedback_a_f2a19fbfdb
make test-slice OWNER=package.ui \
  ROWS=package.ui.frontend_unit.application_selector_contracts_675995f2ab
```

| Command / check | Result and retained run |
| --- | --- |
| Creation frontend slice above | PASS, 20260905T161912Z-p38572. |
| Creation browser slice above | PASS, 20260905T162914Z-p46655. |
| Focused accessibility slice above | PASS, 20260905T154530Z-p35302; superseded by full final-source gate below. |
| Selector slice above | PASS, 20260905T161912Z-p38593. |
| make test-slice OWNER=web.application | PASS, 20260905T161912Z-p38580, 58/58 units. |
| make browser-e2e-a11y | PASS, 20260905T161851Z-p95175, 12/12 units. |
| make browser-e2e-visual-update | PASS, 20260905T160935Z-p83842; reviewed/restored unrelated candidate changes as above. |
| make browser-e2e-visual, twice after final presentation edit | PASS, 20260905T162101Z-p51590 and 20260905T162445Z-p98917, each 12/12 units and 110 reconciled captures. |
| make generate | PASS, 20260905T161402Z-p32475 after golden scope restoration. |
| make agent-finalize before final static checks | PASS, 20260905T163043Z-p91096; generated unchanged; JSON shape/catalog/tier/drift pass; retained-run maintenance skipped with RESULTS_DIR unset. |
| make frontend-typecheck | PASS, 20260905T163130Z-p94574. |
| make frontend-import-boundary-check | PASS, 20260905T163130Z-p94768. |
| make lint-biome | PASS, 20260905T163130Z-p94710. |
| make generated-artifact-policy-check | PASS, 20260905T163130Z-p94244. |
| make lint-markdown | PASS, 20260905T163522Z-p97004; rerun after this final tracker edit. |
| git diff --check and explicit allowed-path/old-golden audit | PASS; rerun after this final tracker edit. |

Retained roots are under `.cartulary/test-results/<run-id>/`. Browser targets
retain per-row results, Playwright reports/traces, accessibility summaries, and
visual reconciliation/renderer attestations. Consult `make explain-run
RESULTS_DIR=<run-root>` for execution details. A queried, undeclared standalone
fixture-check target returned a usage error; registry validation instead uses
the authored target surface, JSON shape, and full visual reconciliation. No
unsupported tool command substituted for required Make evidence.

Final scope audit: exactly 32 changed/new allowed paths, comprising this seam's
source/tests, authored routing, three generated evidence projections, eight new
goldens and this handoff. All 102 pre-existing goldens are byte-identical to HEAD.
Branch remains main at `3972662b1ca2f994eec4ae453cc3f231de4cb6e1`; the dirty tree
contains only this authorized work. No excluded path changed. `git diff --check`
passes. No commit, push, deployment, or next seam occurred.

IC-05 exit: all required seam evidence and acceptance dispositions are complete;
compatibility and atomic rollback are documented, and no owner contradiction or
unresolved creation defect remains. Remaining limitations and intentionally
unselected broader checks are explicit above. The work is ready for user review;
no additional seam or publication action is authorized.
