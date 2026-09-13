# Incident Metadata Editing Refactor Handoff

## Baseline, authority, and scope

Execution entry: main, HEAD e44bebd9e0088ff43f5316770d798c334845a498,
clean tree. Root AGENTS.md is the only applicable repository instruction file.
Stack: React 19, TypeScript 6, Vite 8, pnpm 10.33, Go 1.27.1. No reset,
commit, push, deployment, dependency change, or unrelated work is authorized.

Core 01 §3.3.5.3.1 / REQ-01-168–180 owns the incident resource, five mutable
fields, sparse/null semantics, current role, versions, no-op and acknowledgement.
REQ-01-590 owns closed-incident restrictions; REQ-01-491/491.1 owns bound text
normalization. Core 01 §3.3.6 owns public errors. Core 04 §1.1.1 and
REQ-04-017/023/028/029 own session lifetime and current incident authorization.
Core 03 §13 and §4.3.2 preserve keyboard ownership and closed-state consequences.
Design §§3, 7, 10, 12, 14–15 governs tokens, local feedback, accessibility and
visual evidence. Domain owns vocabulary. No adopted-owner contradiction found.

The digest was read in its localized order during planning. Its map at
2356949f was revalidated against current source; no package edits are authorized.
Membership management, membership audit, and account editing handoffs were read
as implementation evidence. The NLSpec research essay is not product authority.

Allowed paths: metadata frontend/client/controller/presentation and tests;
minimal App, route, workbook and shared integration; incidents OpenAPI and
protocol selection inputs plus required compatibility accounting; matching
Make-generated derivatives; relevant backend contract/behavior tests, semantic
selectors, authored source ownership and verification routing; browser support,
accessibility and reviewed visuals; this handoff. Summary, lifecycle, preferences,
membership/audit feature behavior, backend storage, dependencies, adopted specs
and the digest remain outside this seam.

Generated boundaries follow tools/generated_artifact_policy.json: internal/gen,
protocol-ts/view-contracts/ui-contracts generated roots, assembled OpenAPI,
generated task/topology files and the visual golden manifest. Never hand-edit.
Tests and generators must not depend on Markdown.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| ME-01 — Baseline, owners, characterization | DONE | Ten cases ran; nine failures and one pass retained. |
| ME-02 — Typed contract and operation ownership | DONE | Generation, 21 frontend cases, backend contract/intent tests and typecheck pass. |
| ME-03 — Presentation and lifetime/departure integration | DONE | One production owner and complete operational states. |
| ME-04 — Service-backed, accessibility and visual evidence | DONE | Required journeys, accessibility and two final ordinary visual passes retained. |
| ME-05 — Final verification and completed handoff | DONE | Final verification and scope/evidence audit retained; next action: none. |

Only the current row may be IN_PROGRESS. Record DONE or BLOCKED and its paths,
commands, results, decisions, risks and next action before advancing. Never cross
a blocked dependency.

## Discovery and decisions

App renders IncidentAdminPanel through WorkbookIncidentControlsPresentation;
useIncidentControlsDrawer owns section/close/Escape/return focus. App owns the
session and route; workbook authorization recovery observes current incident role.
The baseline metadata handler sent all five fields, replaced drafts on reads,
did not consume PATCH acknowledgement, and lacked synchronous write admission.
Baseline App route departure chose the first pending owner. These are source findings;
race consequences require characterization before being called reproduced defects.

Use one App-retained incident/session editor, exact raw field inputs, sparse
payloads, immutable attempts, separate acknowledgement and observation state,
30-second observation deadlines and no metadata replay receipt/transaction ID.
Use existing server normalization; no approximate frontend normalizer. User chose
inline current-versus-intended comparison, explicit Use this version, then Save.
Ordinary close/section changes retain memory-only work and revalidate on return.

Existing backend admission emits forbidden_field for immutable members although
that reason is absent from the adopted invalid_incident_patch registry. This
pre-existing server reason mismatch is outside scope; the client retains
only safe generic invalid-patch rejection and does not claim reason conformance.

## Planning verification

Current make help/help-all and task guides were inspected for module.incidents,
web.application, web.workbook, package.protocol_ts, package.ui and web.design.

- make test-slice OWNER=web.application
  ROWS=web.application.regression.incidentadminpanel_patches_required_incident_met_bd946d6caa:
  PASS 2/2, .cartulary/test-results/20260908T065510Z-p1822908.
- make service-backed-test-slice OWNER=module.incidents
  ROWS=module.incidents.integration.real_incident_patch_persists_only_patchable_meta_9fc8fc207f:
  PASS 3/3, .cartulary/test-results/20260908T065510Z-p1822913.
- git diff --check: PASS; execution entry revalidated clean.

These are existing baseline checks, not reproductions of the new race matrix.

## Execution evidence

The following records preserve each workstream exit and the evidence available
at that point. Later repairs and final results are recorded below.

### ME-01 exit

Added incidentMetadataCharacterization.test.tsx, its old-path test surface,
shared incident fixtures, authored module.incidents row and source ownership.
make generate initially failed at 20260908T070217Z-p1841857 because helper
calls were mistakenly included as selector titles. Corrected the authored row;
generation passed at 20260908T070310Z-p1845246.

CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.metadata_characterization ran against unchanged
production and intentionally failed 1/2 graph units at
.cartulary/test-results/20260908T070327Z-p1848277. The row stderr JSON retains
nine assertion failures: blanket payload, duplicate admission, newer input lost,
section-read draft loss, close draft loss, absent conflict review, cross-incident
rejection publication, incorrect closed explanation, and insufficient-role
values/explanation. The failed-follow-up-read case passes; this does not establish
that acknowledgement consumes the authoritative response. Session retirement,
ignored abort, timeout and late acknowledgement remain risks until controller
and App lifetime tests. No unauthorized server mutation or exploit is claimed.

No blocked dependency. Next: ME-02 typed projection, client and operation owner.

### ME-02 exit

Added incidentMetadataModel/controller, generated-operation metadata client,
17 ownership cases, four transport cases, and backend request projection and
normalization/field-intent tests. New request schema is closed, requires only a
positive base version, permits no-op, and carries nullable raw strings with
normalization descriptions. The operation-level x-cartulary-string-contracts
annotation records bound contracts and normalized scalar limits. No raw-input
maxLength rejects valid normalized input. TLP uses only canonical tokens/null.

make generate at 20260908T070515Z-p1849252 identified missing compatibility
accounting. Two intermediate generation runs (20260908T071025Z-p1851663 and
20260908T071402Z-p1854328) rejected schema-level extension keywords in strict
AJV. Moved annotations to the OpenAPI operation without changing the generator
or validation policy. The resulting compatibility report at
20260908T071506Z-p1856412 identified three exact additive entries; reviewed and
recorded them in the existing 2.0.0 change set. Final make generate PASS at
20260908T071626Z-p1858072.

- make test-slice OWNER=web.application
  ROWS=web.application.regression.metadata_ownership,web.application.regression.metadata_transport:
  PASS 3/3 at 20260908T071645Z-p1861290.
- make test-slice OWNER=module.incidents
  ROWS=module.incidents.unit.metadata_openapi,module.incidents.unit.metadata_field_intent:
  PASS 2/2 at 20260908T071645Z-p1861294.
- make frontend-typecheck: PASS 2/2 at 20260908T071645Z-p1861417.

The controller uses the existing bounded observation primitive, separately
tracks unresolved transport settlement, accepts a late current 200, and never
replays metadata. Raw text goes to existing backend admission. Publication and
read failures cannot revoke confirmation; field errors bind submitted revisions.
No blocked dependency. Next: ME-03 App-retained presentation and departure.

### ME-03 exit

App now retains useIncidentMetadata above drawer mounts. IncidentMetadataPanel
owns native fields, persistent feedback, inline review and deliberate discard.
The old panel retains summary/lifecycle/preferences only. Typed resource callbacks
publish monotonically into workbook identity and mounted summary without reading
preferences. App reviews membership, metadata, then deployment-user work in order;
history approval rechecks outstanding work before publication. Session retirement
and incident-access loss clear metadata before existing navigation.

Retired three IncidentAdminPanel cases and their catalog entries: blanket payload
and transient feedback were superseded; patch-triggered preference rereads are
unrelated work; role/access assertions now use the dedicated metadata owner.
Characterization scheduling now waits for actual dispatch before testing an old
response, and for initial observation before changing its mocked role. The new
preflight correctly prevents dispatch if the incident retires during admission.

- Generation PASS 20260908T073015Z-p1874429 and 20260908T073323Z-p1893767.
- Metadata ownership/transport/lifetime/departure slice PASS 5/5 at
  20260908T073333Z-p1897158 (27 frontend cases).
- Characterization PASS 2/2 at 20260908T073219Z-p1888352 (all ten cases).
- Frontend typecheck PASS 2/2 at 20260908T073335Z-p1897418.
- Import boundary PASS 2/2 at 20260908T073332Z-p1896816.
- Format PASS 2/2 at 20260908T073237Z-p1889392.

Intermediate failures: generation/format at 072824/072823 found an obsolete
unprefixed selector; typecheck at 072823 found unused retired imports and a
subscription destructor return; format at 073035 found timer dependency accounting;
characterization at 073044 found the dispatch/mock scheduling noted above;
import check at 073236 required generated types to enter through a protocol
adapter. These were repaired locally; no boundary policy was weakened.
No blocked dependency. Next: ME-04 service-backed, accessibility and visuals.

### ME-04 progress and repairs

Added service-backed metadata journeys, seam-specific fault fixtures, authored
browser rows/batches, native-keyboard/viewport checks and twelve visual capture
intents. Existing discovery/metadata browser feedback now remains persistent;
its hover-driven toast expiry assertion was intentionally replaced. The drawer
page helper lets metadata own loading/unavailable readiness, as membership does.

Narrow supporting slices passed: package.protocol_ts (2/2, 074120-p2049500),
package.ui (3/3, 074120-p2049506), web.workbook adapter/lifecycle policy (3/3,
074120-p2049522), web.application summary/preferences/audit lifetime (5/5,
074120-p2049536), membership management lifetime (2/2, 074630-p2116676).
All run roots use .cartulary/test-results/20260908T followed by the suffix shown.

First browser attempts exposed shared build/image-stamp collisions when separate
browser targets were launched concurrently (073922-p1907399 and
073923-p1907624). Browser builds are now serialized. Fixture-only typecheck
failure 073924-p1908496 needed nullable response typing; format failure
073912-p1903195 needed explicit nullable promise guards. Generation passed at
073852-p1899646, 074416-p2057578, 074617-p2113644, and 074927-p2129902.

The first functional browser run passed recovery/current authorization but found
two test assumptions: lifecycle closure requires its existing reason, and the
ordinary drawer itself has dialog semantics. Departure assertions now count
only departure dialogs, preserving the drawer. Four functional journeys then
passed (11/11 graph units, 074500-p2064941): all five real sparse fields and
clears/normalization/no-op, real competing version, persistent confirmation after
failed reads, uncertainty, current roles/closure/access loss, and sequential
membership/metadata departure with Stay/history. A fifth journey adds pending
transport across drawer return and deliberate departure into another incident.

Accessibility at 074051-p2004730 reproduced lost keyboard focus when inline review
removed its focused button. Explicit review now focuses Save. Drawer return can
observe authorization/resource while a write remains pending; its write lock
still prevents a second admission. The focused ownership/lifetime/departure slice
passed 4/4 at 074501-p2065158; typecheck passed 2/2 at 074502-p2066168.

Further characterization at 074902-p2129304 reproduced a local discard becoming
an unintended reversal when a pending acknowledgement arrived. A draft reset
revision now distinguishes explicit discard from later typing. The same row
passed 2/2 at 074928-p2130477. Field errors also require a submitted field, and
unresolved retired transport has a safe explanation for temporarily disabled Save.

The selected ordinary visual run (074051-p2004730) completed all twelve capture
intents; it failed because their new goldens do not exist. Reconciliation has
zero ambiguous mappings. Its 181 apparent orphans are unselected existing rows,
so this partial report cannot authorize removing any existing golden. At that point, a full
ordinary run was required before the public update target; no images had been
promoted.

make agent-finalize passed 1/1 at 074756-p2126031 before broader verification.
RESULTS_DIR was unset: no qualifying successful full warm run exists for this
source, so retained-run maintenance was skipped. Format passed at
074951-p2133536. Next: complete browser/accessibility and full visual review.

### Owner-to-change mapping

| Adopted owner | Implementation | Verification |
| --- | --- | --- |
| Core 01 §3.3.5.3.1, REQ-01-168–180 | Authored PATCH body/selection, narrow generated client, five raw sparse fields, authoritative response, no-op and version review. | Contract/admission tests; model/client/controller rows; live metadata journey. |
| Core 01 REQ-01-491/491.1 | Existing backend string contracts retained; raw nonempty input and explicit null; schema documents normalized limits. | Unicode, controls, normalized limits, null/empty and immutable-member tests. |
| Core 01 public error registry | Safe recognized identities and field/revision ownership; unknown outcomes remain uncertain. | Transport, delayed rejection and browser diagnostic exclusion. |
| Core 01 REQ-01-590 | Closed metadata stays visible to current members; writes disabled and eligibility reviewed after reopening. | Current-role/controller matrix and real lifecycle browser journey. |
| Core 04 current session/visibility/membership | App lifetime, current recovery port, protected-data clearing, role review, no deployment-admin bypass. | Strict Mode binding, actor/session/incident fencing and access-loss browser journeys. |
| Core 03 interaction/lifecycle and design §§7/10/12/14–15 | Native fields, local draft, persistent outcomes, inline comparison, Save focus, retained drawer work and sequential route review. | Keyboard/viewport/accessibility, departure, summary/preferences regressions and visual captures. |
| Domain vocabulary | Incident resource and promoted metadata terminology; no workbook creation discovery or invented severity/phase enum. | Scope and payload audit. |

Exact Core paths are docs/spec/01_architecture_storage_and_view_contracts.md,
docs/spec/03_workbook_interaction_collaboration_and_workflows.md, and
docs/spec/04_security_deployment_and_conformance.md. These remain unchanged. Typed schema
annotations and release accounting are downstream projections, not new policy.

### Digest acceptance applicability

| Rows | Disposition and evidence boundary |
| --- | --- |
| A001–A003 | PASS: adopted owner map, bounded seam, actual checkout and stale localization revalidation above. |
| A004–A005 | PASS: existing tokens/theme only; no palette, theme registry or component design system. |
| A006 | PASS: metadata density geometry uses current shared selection; computed compact/comfortable padding and typography are asserted. Full-cell/grid geometry remains unchanged and is adjacent regression evidence. |
| A007 | NOT APPLICABLE: incident metadata is not row creation; no discovery requirement imported. |
| A008–A009 | PASS: metadata uses existing drawer/shell sizing and internal scrolling. Narrow/short/zoom/text-spacing control reachability passes; no viewport algorithm replaced. |
| A010 | NOT APPLICABLE: inspector routing is unchanged. |
| A011 | PASS: existing workbook continuity remains an adjacent browser/visual regression boundary; only current incident identity is published. |
| A012–A013 | NOT APPLICABLE: PATCH has neither transaction ID nor workbook blocked-edit queue. No replay protocol copied. |
| A014–A017 | PASS: native editing, exact retention, explicit review, distinct acknowledgement/observation/access states; backend normalization remains authoritative. |
| A018 | NOT APPLICABLE: evidence lifecycle unchanged. |
| A019 | PASS: full accessibility at 082946-p2709388, including keyboard, focus, narrow/short viewport, zoom and text spacing. |
| A020 | PASS: metadata operational states have controller/browser coverage and twelve capture intents; no generic component variants introduced. |
| A021 | NOT APPLICABLE: row virtualization unchanged. |
| A022 | PASS: twelve individually reviewed additions; two fresh ordinary visual passes after the final promotion, 084432-p2966048 and 085037-p3015811. Existing fixtures unchanged. |
| A023–A024 | PASS: shared existing field selectors and semantic role/state; selector-policy correction passes; no executable Markdown dependency added. |
| A025 | PASS: final generation drift, generated policy, JSON shape and catalog checks after reviewed golden promotion; no hand-edited derivatives. |
| A026 | PASS: existing public behavior only; backend reason mismatch explicitly excluded. |
| A027 | PASS: ME-01–ME-05 complete, exact changed paths and retained results recorded, compatibility/limitations/rollback bounded. |

Full frontend-unit run 075012-p2139606 passed 482/483 units; the remaining
selector-policy assertion rejected two heading-based visual scroll anchors.
Anchors now use the metadata region. The focused policy row passed 2/2 at
075412-p2243294. No selector policy exception was added.

Checks at 075439 passed: lint-biome (p2244119, 2/2), import boundaries
(p2244113, 2/2), generated policy (p2244053, 3/3), JSON shape (p2244055, 3/3),
test-catalog-check (successful silent public target), and generate-drift
(p2244051, 4/4). Corrected focused accessibility passed 11/11 graph units at
075511-p2248487.

Full ordinary make browser-e2e-visual at 075011-p2137831 ran all existing
scenarios successfully. The new row completed its twelve functional capture
states and failed solely for missing new PNGs. Reconciliation: 193 intents,
181 existing active goldens, zero orphans/ambiguous/unresolved registry entries,
and twelve new missing goldens. This authorizes adding the new metadata evidence
through the requested public update workflow; no existing golden is to be removed.

### Visual refresh record

Accepted trigger: the owner-backed metadata editing workflow now has distinct
loading, exact draft, pending, conflict, uncertainty, confirmed-follow-up-failure,
closed/read-only and density presentations. No new design system or D-VFIX claim
is introduced. Authored row: web.design.visual.metadata_editing; scenario:
scenario_7462a6753b02; project chromium. The captures are valid active nonregistry
evidence. Existing registered fixtures remain adjacent regression evidence.

Public make browser-e2e-visual-update passed 12/12 graph units at
075631-p2295707: 36 browser scenarios passed; reconciliation accounts for all
193 active images without missing/orphan/ambiguous mappings. All twelve new
metadata images were individually inspected with the local image viewer.
Native labels, comparison values, persistent save/read distinction, read-only
values, current tokens/density and focus indicators are legible. Controls use the
existing drawer's internal scrolling; short/zoom captures intentionally show the
review area, with keyboard reachability covered separately.

The update also rewrote 18 unrelated existing PNGs. None had an accepted refresh
trigger; all had passed the preceding ordinary visual run. Their original HEAD
bytes were restored without resetting the checkout. Rejected update copies and
inspection sheets were retained at /tmp/cartulary-metadata-visual-review for this
session. Most changed bounds were tiny renderer variations; no unrelated snapshot
change remains. make generate then regenerated the golden manifest and passed at
080159-p2347364.

| Capture / new filename stem (all metadata-…-linux.png) | Viewport / state |
| --- | --- |
| loading, dirty, saving | 1280×720, default density, initial/top form context. |
| uncertain, conflict | 1280×720, default density, inline comparison in view. |
| review-narrow | 390×480, review action focused and reachable. |
| review-zoom | 1280×720, 200% browser zoom, review area. |
| review-spacing | 768×640, WCAG text-spacing override, review area. |
| confirmed-refresh-failure, closed | 768×640, default density, status or read-only resource. |
| compact, comfortable | 768×640, actual shared account density; computed typography/padding asserted. |

All captures remain full-viewport screenshots under the pinned renderer, fonts,
existing dynamic-identity masking, and existing preparation helpers. New metadata
scroll anchors are the semantic panel/review region. Existing viewport, zoom,
mask, scroll and crop contracts were not changed.

Full frontend-unit rerun passed 483/483 at 080239-p2352169. Backend contract,
normalization/intent and metadata characterization slices passed 4/4 at
075730-p2342879. Markdown lint passed at 075945-p2345542. At that point, two fresh ordinary
visual passes and the remaining final service/browser evidence were pending.

Ordinary visual verification passed 12/12 at 080237-p2350577 after promotion and
restoration of unrelated images. Full accessibility at 080824-p2446991 passed 30
scenarios and failed one older metadata assertion: it expected role=status on
the old paragraph and a raw authorization_denied element after dispatch. Updated
it to the single feature live region and current-role preflight explanation.
The adjacent auth browser test likewise now verifies no UI PATCH after demotion,
then separately verifies the unchanged public API 403 envelope with the typed
operation. Read-only local discard precedes deliberate reload.

Focus review also identified the removed-control case during role/read-access
changes. The metadata presentation now moves focus to its stable Refresh control
when a previously focused control is removed. The lifetime and ownership slice
passed 3/3 at 082318-p2508204. A test-only typecheck error at
082429-p2557160 used Playwright's exact option in Testing Library; removed that
unsupported option. No production type assertion or policy exception was used.

The combined incident service-backed slice passed 14/14 at 082319-p2508472:
all five new metadata browser journeys, existing directory/metadata and membership
browser journeys, and existing sparse/no-op/version backend integration. This
includes a pending PATCH through drawer return, deliberate forgetting, navigation
to another incident and exclusion of late publication. Ordinary visual
passes were then repeated after the focus correction.

### Changed paths

The following 60 paths form the reviewable change set. No unrelated PNG,
dependency, adopted owner, digest, backend production or storage change remains.

- apps/web/e2e/auth-and-incident-directory.spec.ts
- apps/web/e2e/incident-administration.spec.ts
- apps/web/e2e/incident-metadata-editing.spec.ts
- apps/web/e2e/pages/deploymentAdministration.ts
- apps/web/e2e/support/incidentMetadata.ts
- apps/web/e2e/workbook.a11y.spec.ts
- apps/web/e2e/workbook.visual.spec.ts
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-closed-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-comfortable-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-compact-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-confirmed-refresh-failure-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-conflict-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-dirty-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-loading-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-review-narrow-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-review-spacing-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-review-zoom-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-saving-linux.png
- apps/web/e2e/workbook.visual.spec.ts-snapshots/metadata-uncertain-linux.png
- apps/web/src/app/App.tsx
- apps/web/src/app/IncidentAdminPanel.test.tsx
- apps/web/src/app/IncidentAdminPanel.tsx
- apps/web/src/app/IncidentMetadataPanel.tsx
- apps/web/src/app/api/incidentMetadataClient.test.ts
- apps/web/src/app/api/incidentMetadataClient.ts
- apps/web/src/app/api/incidentMetadataContracts.ts
- apps/web/src/app/appDepartureReview.test.ts
- apps/web/src/app/appDepartureReview.ts
- apps/web/src/app/incidentMetadataCharacterization.test.tsx
- apps/web/src/app/incidentMetadataController.test.ts
- apps/web/src/app/incidentMetadataController.ts
- apps/web/src/app/incidentMetadataLifecycle.test.tsx
- apps/web/src/app/incidentMetadataModel.ts
- apps/web/src/app/incidentMetadataTestSurface.tsx
- apps/web/src/app/useAppRouteRuntime.ts
- apps/web/src/app/useIncidentMetadata.ts
- apps/web/src/shared/workbookShellContracts.ts
- apps/web/src/testing/incidentMetadataTestSupport.ts
- apps/web/src/workbook/WorkbookShell.tsx
- apps/web/src/workbook/components/WorkbookIncidentControlsPresentation.tsx
- apps/web/src/workbook/hooks/useWorkbookIncidentIdentity.ts
- contracts/openapi-releases/2.0.0.change-set.json
- contracts/openapi-source/owners/module.incidents/openapi.json
- contracts/openapi/cartulary.openapi.yaml
- contracts/protocol-ts/http-operations.v2.json
- docs/handoffs/incident-metadata-editing-refactor-handoff.md
- internal/gen/contractopenapi/artifacts_gen.go
- internal/gen/openapioperations/catalog_gen.go
- internal/modules/incidents/httpapi/metadata_contract_test.go
- internal/modules/incidents/metadata_intent_test.go
- packages/protocol-ts/src/generated/core-http-types.ts
- packages/protocol-ts/src/generated/core-http-validators.ts
- packages/protocol-ts/src/generated/http-operation-bindings.ts
- tools/browser_e2e_batch_manifest.json
- tools/execution_topology_render_index.json
- tools/frontend_source_ownership.json
- tools/frontend_visual_golden_manifest.json
- tools/test_families/module.incidents.json
- tools/test_families/web.application.json
- tools/test_families/web.design.json

### Compatibility, limitations, and rollback

Compatibility is frontend editing plus faithful typed projections of the existing
public PATCH contract. Public routes, fields, error status behavior, API/release
versions and stored data are unchanged. There is no migration or dependency change.
Severity and phase remain text. The existing backend forbidden_field reason
mismatch is retained outside this seam and receives a generic safe invalid-patch
rejection; this handoff does not claim complete server reason-registry conformance.

Uncertain writes intentionally have no inferred receipt. Matching current values
and newer versions are observations; explicit review permits a separate new Save.
An ignored abort can keep transport unsettled, so the single-write lock remains
until settlement. Local drafts/recovery are memory-only and deliberately forgotten
on reviewed departure, incident-access loss, or session retirement. No browser
reload or server cancellation guarantee is claimed.

A coherent rollback would restore the editor/App/workbook integration, authored
OpenAPI request/operation selection and release accounting, matching generated
bindings, tests/routing/selectors and visual additions together. Regenerate through
public Make targets and verify the restored source. The backend already supports
this request shape, so rollback requires no storage change and must not attempt to
reverse metadata values users have legitimately saved. No rollback, commit, push,
or deployment was performed during this task.

Late verification: frontend typecheck passed 2/2 at 082512-p2562316; the selector
policy row passed 2/2 at 082513-p2563108. The adjacent auth service-backed row
initially failed at 082511-p2562020 and 082700-p2610408 because its independent
API probe lacked the browser session cookie. The test now uses the established
authHeadersForStorageState helper without exposing credentials. It passed 11/11
at 082833-p2657697, proving preflight prevents UI dispatch while direct typed API
admission preserves the existing 403 authorization_denied envelope. Format passed
at 082938-p2705104; full make lint passed 11/11 at 082947-p2711380.

Full accessibility rerun passed 12/12 at 082946-p2709388 (31 browser scenarios).
Full frontend-unit passed 483/483 again at 083246-p2771923 after the final focus
change. Ordinary visual run 083245-p2770117 passed every other scenario and
identified one explained metadata difference: 392 pixels for the visible focus
outline on Refresh in metadata-closed-linux.png. The prior recovery button had
been removed by the read-only transition. Actual and diff images were inspected;
the change is the intended stable focus fallback, with no layout/content drift.
The second public update accepted this one metadata image; incidental rewrites
of other images retained their previously reviewed bytes.

The focus golden update passed 12/12 at 083854-p2865889. The updated
metadata-closed-linux.png was inspected again and matches the reviewed focus-only
diff. Thirteen unrelated existing images and one incidental metadata text-spacing
rewrite retained their pre-update bytes; only the closed-image focus change was
accepted in this update. No renderer, viewport, mask, density, crop, or assertion
tolerance changed. The golden manifest was regenerated through make generate.

make generate passed after the focus update at 084235-p2915143. The adjacent
membership-audit real-browsing and retained-membership departure service-backed
slice passed 13/13 at 084250-p2918203. Scope remains exactly 60 changed paths;
git diff --check passes. No additional production changes are planned.

The first fresh ordinary visual pass after the final focus update passed 12/12
at 084432-p2966048 (36 browser scenarios). Reconciliation accounts for 193
active images with zero missing, orphan, ambiguous, or unresolved mappings.

### ME-04 exit

The second fresh ordinary make browser-e2e-visual passed 12/12 at
085037-p3015811, after the final focus golden update. Together with
084432-p2966048, both runs passed all 36 browser scenarios and reconciled all
193 active images: zero missing, orphan, ambiguous, or unresolved mappings.
No source or golden changed between these runs.

Required live editing, conflict, uncertainty, current authorization, retained work,
sequential departure, pending transport, adjacent lifecycle/directory/membership/
audit, keyboard/accessibility and responsive/density journeys now have passing
retained evidence. Full frontend-unit (483/483), accessibility (31 browser
scenarios), and lint (11/11) passed after the final production change.
No blocked dependency. Next: ME-05 finalizer, final checks and handoff closure.

### ME-05 final verification

make agent-finalize passed 1/1 at 085350-p3065067 before the final verification
set. Its unit-artifacts/finalize-summary.json reports generated status unchanged,
zero updated files, and no rollback needed. RESULTS_DIR was unset because no
qualifying successful full warm check exists for the exact source. Retained-run
canonical evidence, scheduler timing, performance evidence and run maintenance
were explicitly skipped, not passed.

Final verification uses the existing owner routes plus full frontend unit,
accessibility, visual and lint targets. No full make check/ci/release-check or
unrelated backend/browser suite was requested or rerun; this bounded frontend
and typed-projection change makes no release or performance claim. Narrow backend
contract, normalization and live persistence tests cover the unchanged server
behavior needed by this seam.

| Final public target | Result / retained run suffix |
| --- | --- |
| make agent-finalize | PASS 1/1, 085350-p3065067; retained-run maintenance skipped as described above. |
| make frontend-typecheck | PASS 2/2, 085438-p3068425. |
| make frontend-import-boundary-check | PASS 2/2, 085438-p3068429. |
| make lint-biome | PASS 2/2, 085438-p3068435. |
| make generated-artifact-policy-check | PASS 3/3, 085438-p3068366. |
| make json-shape-check | PASS 3/3, 085438-p3068368. |
| make test-catalog-check | PASS, successful silent public target in the same final invocation. |
| make generate-drift | PASS 4/4, 085438-p3068364. |
| make lint-markdown after ME-05 DONE | PASS, 085615-p3073417; repeated after recording this result. |
| git diff --check and completed-handoff scope audit | PASS; five DONE rows and all 60 changed paths match, with exclusions intact. |
| make frontend-unit | PASS 483/483 graph units, 083246-p2771923. |
| make browser-e2e-a11y | PASS 12/12 graph units, 31 browser scenarios, 082946-p2709388. |
| make lint | PASS 11/11, 082947-p2711380. |
| make browser-e2e-visual, first final ordinary pass | PASS 12/12, 36 browser scenarios, 084432-p2966048. |
| make browser-e2e-visual, second final ordinary pass | PASS 12/12, 36 browser scenarios, 085037-p3015811. |

All suffixes resolve under .cartulary/test-results/20260908T. The retained
run-summary.json, target-summaries, unit logs, browser reports and visual
reconciliation provide execution evidence; they do not define product policy.
Finalizer and final checks changed no production source or goldens after the
passing full frontend, accessibility and ordinary visual runs.

### ME-05 exit and final scope audit

All five workstreams are DONE. The final checkout remains main at
e44bebd9e0088ff43f5316770d798c334845a498, with exactly the 60 authored/generated/
reviewed paths listed above; the initial tree was clean. The list was compared
against both git diff --name-only and git ls-files --others --exclude-standard,
with no missing or extraneous entries. Only this handoff is changed Markdown.
No existing PNG, dependency artifact, adopted spec, digest, backend production or
storage path changed. git diff --check passes. No reset, commit, push or deployment
was performed. The final Markdown, whitespace and scope checks are repeated after
marking this row DONE, against the completed handoff.

Next action: **none**.
