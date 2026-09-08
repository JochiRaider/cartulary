# Incident Membership Management Refactor Handoff

## Baseline, authority, and scope

Execution entry: `main`, HEAD `f456fe2599b264f779d2874f305ca7c55d2976f2`,
clean tree. Root AGENTS.md is the only applicable instruction file. Stack:
React 19, TypeScript 6, Vite 8, pnpm 10.33, Go 1.27.1. No reset, commit, push,
deployment, dependency change, or unrelated work is authorized.

Core 01 §3.3.5.1 / REQ-01-127–137 owns membership identity, selectors,
operations, versions, and last-admin behavior; §§3.3.6–7 own errors and live
pagination; REQ-01-497 owns email normalization. REQ-01-590 explicitly allows
membership administration on closed incidents. Core 04 REQ-04-017, 023, and
030 owns current session/incident authorization and incident-only access loss.
Core 03 §13 preserves keyboard ownership. Design §§7, 10, 12, 14–15 governs
local feedback, density, accessibility, and visual evidence. Domain governs
vocabulary. The digest was read in its localized order during planning and is
advisory; account and incident-membership-audit handoffs are implementation
evidence. The research NLSpec essay is not an adopted product owner.

Allowed paths: membership frontend implementation/tests under apps/web/src/app
and its api directory; necessary App, route departure, workbook presentation,
and shared lifetime interfaces; membership OpenAPI/protocol selection inputs;
focused incidents backend contract/behavior tests; browser support, accessibility,
visual fixtures and reviewed goldens; UI selectors; authored source ownership and
verification routing; matching Make-generated derivatives; the narrow Core 01
REQ-01-136 clarification and this handoff. The digest, completed audit owner,
summary/promoted fields/lifecycle/preferences behavior, database schema,
dependencies, and stored data remain outside the refactor.

Generated roots are governed by tools/generated_artifact_policy.json, including
internal/gen and protocol-ts/view-contracts/ui-contracts generated roots. OpenAPI,
task topology, and golden manifests are generated outputs, never handwritten.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| MM-01 — Baseline, owners, characterization | DONE | Owner clarification and eight failed characterization assertions retained. |
| MM-02 — Typed client and workflow state | DONE | Generated operations, 22 client/controller cases, contract test and typecheck pass. |
| MM-03 — Presentation and integration | DONE | New owner integrated; characterization, session/drawer and retained-owner regressions pass. |
| MM-04 — Browser, accessibility, visual | DONE | Backend, ten browser journeys, keyboard/focus and two ordinary visual passes retained. |
| MM-05 — Final verification and handoff | DONE | Required evidence passes; completed handoff and final scope audit retained. |

During execution, only the current workstream was IN_PROGRESS. Each exit was
recorded before advancing; no blocked dependency was crossed. All are now DONE.

## Discovery and decisions

App renders IncidentAdminPanel through WorkbookIncidentControlsPresentation;
useIncidentControlsDrawer owns section transitions, Escape close and return
focus. Membership reads currently discard paging; accepted reads replace all
role drafts. Writes have no synchronous admission guard and mix acknowledgement
with follow-up reads. Timing and lifetime consequences require characterization.
AppSessionController owns session acceptance and recovery; workbook authorization
owns role presentation. The completed audit controller stays separate.

User-selected retention: one editor and recovery record across close, section
switch, and hidden document within one incident/session. Clear accepted rows and
cursors; return checks access and loads page one. Editor switches use Stay or
Discard. Route departure permits explicit forgetting of local recovery, with
clear notice that departure cannot cancel a server operation. No multi-incident
cache. Defaults: 100 members per page, 20 prior cursors, 30-second observation.

## Contract clarification and compatibility

REQ-01-136 explicitly requires base_membership_version and returns
membership_version_conflict when the existing membership version differs.
Rationale: removal must apply to the version the administrator reviewed; an
intervening role change requires renewed review. Missing membership remains 404,
successful deletion remains 204, and current-admin/last-admin rules stay intact.
No inspected adopted clause contradicts this clarification. Existing application
code already performs this check inside its per-incident transaction boundary;
that code is compatibility evidence, not the rationale for the requirement.

The change clarifies an existing public precondition, adds missing typed client
coverage, and improves the frontend workflow. No new route, permission,
dependency, session controller, or stored-data migration was introduced. Rollback
must restore the feature, matching authored/generated projections, tests, routing,
and evidence coherently, retaining agreement between the owner and implementation.

## Verification routing and planning evidence

Current task guides were read for module.incidents, web.application,
web.workbook, package.protocol_ts, package.ui, and web.design. Route/server
behavior belongs to module.incidents; workflow/App to web.application; drawer and
role propagation to web.workbook; generated binding and selector checks to their
packages; accessibility and visuals to web.design. Update authored test families
and source ownership before generating downstream topology.

Planning checks at the unchanged baseline:

- Membership frontend payload/decoder slice: PASS 3/3 graph units,
  `.cartulary/test-results/20260908T035522Z-p783533`.
- Service-backed concurrent last-admin/no-op PATCH slice: PASS 4/4 units,
  `.cartulary/test-results/20260908T035522Z-p783543`.
- git diff --check: PASS; execution entry is still clean.

These checks are baseline evidence, not reproductions of the new failure matrix.
No tests, generators, or runtime evidence may depend on Markdown.

## Execution evidence

Execution follows the sequential exits below.

### MM-01 exit

Clarified Core 01 REQ-01-136 without changing authorization, lifecycle, identity,
or idempotency. Added incidentMembershipManagementCharacterization.test.tsx and
its module.incidents row/source ownership. make generate passed at
20260908T040144Z-p802794.

CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.membership_management_characterization intentionally
failed at 20260908T040205Z-p805893 (1/2 graph units passed). The row stderr JSON
contains eight assertion failures against unchanged production: absent next-page
action; two writes from repeated create activation; newer email draft erased;
role draft replaced on section reload; old rejection published into another
incident; acknowledgement delayed behind follow-up reads; missing version-conflict
observation action; unnamed member role select. The conflict assertion demonstrates
missing recovery, not uncertain-transport behavior. Session retirement, ignored
abort, timeout/replay, and genuinely uncertain outcomes remain risks until focused
controller/integration tests. No unauthorized server write or exploit is claimed.

No blocked dependency. Next: MM-02 typed projections, client, workflow state.

### MM-02 exit

Added membership model/controller and generated-operation client, 16 workflow
cases and six transport cases, a strict request-projection backend test, and
shared membership fixtures. Authored PATCH/DELETE request schemas and operation
selection now generate their public types/bindings. The typed client additionally
found that listIncidentMemberships was selected but its authored query parameters
were missing; projected only existing REQ-01-129 limit/cursor_token. No new query
behavior or generator redesign.

make generate initially failed compatibility accounting at
20260908T040358Z-p806999: four new body/schema additions needed reviewed release
change-set entries. Added the exact module.incidents entries to
contracts/openapi-releases/2.0.0.change-set.json; the subsequent existing-list
parameter repair adds one more additive entry. Generation passed at
20260908T041013Z-p809704, 20260908T041349Z-p813634, and
20260908T041533Z-p822857. This is required projection compatibility accounting,
not a public policy or release-version change.

Initial focused frontend run 20260908T041420Z-p821078 passed controller tests but
exposed the missing list parameters. Initial typecheck
20260908T041420Z-p821221 found mock tuple/paging-union typing and an unused import,
all corrected. Backend request contract slice passed 1/1 at
20260908T041420Z-p821083. Final focused web.application ownership/transport slice
passed 3/3 at 20260908T041614Z-p830305; frontend-typecheck passed 2/2 at
20260908T041614Z-p830396. make format passed at 20260908T041559Z-p825977.

No blocked dependency. Next: MM-03 presentation, App lifetime integration, and
retirement of IncidentAdminPanel membership behavior.

### MM-03 exit

App now renders the membership feature directly and retains its sole controller
above drawer mounts. The hook reuses AppSessionController recovery and workbook
role propagation, including admitted self-demotion while closed. App route policy
asks Stay or Leave and forget recovery; session/incident-access retirement bypasses
that prompt. IncidentAdminPanel no longer owns membership state, reads, handlers,
or presentation. Summary, promoted fields, lifecycle, preferences and audit remain
with their existing owners.

Added labeled forms, one editor, identity/version removal review, local operation
receipts and distinct conflict/uncertainty/read recovery. Native registered dialogs
own draft/departure review; connected heading focus replaces detached removal
controls. Generated type references pass through the existing API adapter boundary.
Unknown error codes are sanitized. Authorization observation timeout suspends
protected presentation and privileged admission.

Migrated the eight characterization cases to the new presentation fixture: PASS
2/2 graph units at 20260908T043134Z-p842210. New presentation/transport/controller
slice passed 4/4 at 20260908T043134Z-p842200. Typecheck initially found extraction
cleanup and a subscription destructor returning boolean (p842330); corrected,
PASS 2/2 at 20260908T043443Z-p857125. Five StrictMode session/drawer lifetime cases
passed 2/2 at 20260908T043443Z-p857025. Import boundary initially rejected a direct
protocol type import (p857157); moved it through publicHttpTypes, PASS 2/2 at
20260908T043546Z-p863177.

Authored routing generation briefly failed because the new workbook row ID did
not match its copied family (20260908T043338Z-p847260); corrected both to
web.workbook.regression, generation PASS 20260908T043429Z-p853914. A format and
task-guide attempt during that invalid catalog failed for the same reason; format
passed at 20260908T043503Z-p858391. Final focused application slice including audit
and retained IncidentAdminPanel regressions PASS 13/13 at
20260908T043546Z-p863064; incidents characterization/lifecycle/access slice PASS
5/5 at 20260908T043546Z-p863079. The former combined membership mutation assertion
now belongs to the new presentation row; promoted-field access assertions remain
in the original panel row with an accurate title.

No blocked dependency. Next: MM-04 service-backed journeys, accessibility,
responsive operational-state fixtures and reviewed visual evidence.

### MM-04 execution notes

Added service-backed live browsing beyond 100 members, real add/change/remove
and audit filtering, pending departure, exact create replay, conflict review,
authorization unavailability, and real self-demotion/self-removal on closed
incidents. Backend tests now explicitly retain same-role create success and
closed-incident create/PATCH/DELETE, stale DELETE, 204, and missing membership.
Generated binding tests cover bodyless DELETE and exact list query names. The
new browser rows use authored semantic scenario identities and generated routing.

AppSessionController cancels superseded session observations. Follow-up access
and list observation therefore run sequentially; acknowledgement precedes both
and their outcomes remain separate. Failed access leaves the list unattempted;
access recovery can then fail the list without resending the write. Opening a
surface gates retained draft/details behind a fresh access check. Unknown or
malformed mutation error identities become uncertain, and repeated cursor
responses expose first-page recovery.

The real keyboard fixture reproduced two integration defects: editor focus
selected the heading before the role field, and nested native-dialog Escape
also reached the drawer. Fixed focus priority in the membership presentation;
the drawer now consumes only Escape belonging to its own dialog. A visual
computed-style assertion reproduced inline card padding overriding density;
removed that override. Disabled membership buttons use neutral existing tokens.

Command evidence (all roots below are under .cartulary/test-results):

| Command / selected evidence | Result and run root |
| --- | --- |
| make generate, new browser routing | Failed 20260908T044248Z-p867941: new scenario IDs violated the catalog identifier shape; fixed authored IDs. |
| make format during invalid catalog | Failed for the same catalog dependency; no product result. |
| make generate after catalog repair | PASS 20260908T044346Z-p871356. |
| make format | Failed 20260908T044406Z-p874477: nullable Promise condition in new fixture; fixed explicit null check. Unrelated existing template-literal advice was left alone. |
| make frontend-typecheck | Failed 20260908T044407Z-p874654: new fixture role union and Node removal types; corrected. |
| module.incidents browser slice | 11/13 units at 20260908T044409Z-p876361; original administration and new mutation recovery rows passed. Live flow reached audit then failed an unawaited/underspecified filter assertion; departure used the wrong directory heading. Both fixture assertions corrected. |
| module.incidents backend membership slice | PASS 6/6 at 20260908T044426Z-p911986, covering concurrency, no-op, versions, immediate authority, deployment-admin denial, live pagination, replay, domain failures and public envelopes. |
| web.application ownership/presentation slice | PASS 3/3 at 20260908T044426Z-p912034 after follow-up sequencing. |
| web.design accessibility slice | 9/11 units at 20260908T044426Z-p912018; reproduced editor-focus defect, corrected. |
| frontend-typecheck and protocol slice | Failed 20260908T044730Z-p998917 and p998804: new test treated generated query-name strings as objects; corrected test projection. |
| module.incidents browser retry | 8/11 at 20260908T044730Z-p998785: harness image warm-stamp temporary-file race between simultaneous public browser targets; no product assertion ran. Subsequent browser targets run sequentially. Harness mechanics unchanged. |
| web.design accessibility/visual slice | 9/13 at 20260908T044730Z-p998794: nested Escape and density defects; missing new goldens; visual finalization rejected the new scenario ID shape. Corrected implementation and authored visual identity; no golden promoted. |
| make format | PASS 20260908T044712Z-p994337 and 20260908T045033Z-p1094535. |
| frontend-typecheck | PASS 2/2 at 20260908T045057Z-p1099221. |
| package.protocol_ts selected operation row | PASS 2/2 at 20260908T045057Z-p1099155. |
| module.incidents new OpenAPI row | PASS 1/1 at 20260908T045057Z-p1099138. |
| web.design accessibility/visual slice | 11/13 at 20260908T045055Z-p1098790: accessibility passed; all ten visual state assertions reached, with only missing new snapshots; old invalid visual scenario identity still prevented reconciliation. |
| make generate | PASS 20260908T045219Z-p1149728 after assigning stable schema-compatible visual identities. |
| make format | PASS 20260908T045253Z-p1152911 and 20260908T045436Z-p1203953. |
| module.incidents browser slice | 9/11 at 20260908T045335Z-p1157230: pending departure and real closed-incident self-actions passed. Audit assertion required its existing target-kind filter; corrected fixture without changing audit behavior. |
| make agent-finalize | PASS 1/1 at 20260908T045522Z-p1209502 before broader visual work. RESULTS_DIR unset; qualifying retained-run maintenance skipped. |
| make format | PASS 20260908T045710Z-p1213003. |

Trace-frame review of the ten visual states was performed from the retained
ordinary visual trace at 20260908T045055Z-p1098790, accounting for renderer and
test clocks. Review accepted the token-based layout, local outcome records,
wrapping removal identity, and internal drawer scrolling; disabled action colors
were then made neutral. Exact PNG review and two fresh ordinary passes after the controlled update
are complete, as recorded below. No completed audit production code or existing golden was changed at this stage.

### Owner and digest applicability map

| Adopted owner | Change / retained behavior | Verification route |
| --- | --- | --- |
| Core 01 §3.3.5.1, REQ-01-127–137 | Exact selector/role/versioned operations, identity and last-admin rules; clarified DELETE concurrency. | module.incidents, package.protocol_ts, membership transport/controller rows. |
| Core 01 collection pagination and public errors | One 100-row accepted page, 20 prior cursor descriptors, join ordering, invalid-cursor recovery, safe typed errors. | Live service pagination, bounded controller and typed decoder rows. |
| Core 01 idempotency and email_address_v1 | Existing-account email normalization, fresh Web Crypto IDs, exact immutable uncertain replay; no PATCH/DELETE transaction IDs. | Create/replay/conformance backend tests and transport/browser recovery. |
| Core 01 REQ-01-590 | Closed incidents retain membership administration. | Extended HTTP conformance and real self-action browser row. |
| Core 04 REQ-04-017, 023, 030 | Current session/incident role admission, role loss versus incident loss, deployment-admin nonmember denial, session retirement. | Existing control-boundary/concurrency rows, real self-actions, StrictMode AppSessionController integration. |
| Core 03 keyboard and interaction ownership | Explicit editor save/cancel, draft/departure review, nested Escape, connected focus, workbook continuity. | web.workbook lifetime and web.design accessibility, operational browser rows. |
| docs/design.md and docs/domain.md | Existing tokens/density, readable local states, membership/account distinction and vocabulary. | Semantic selectors, computed layout/focus assertions and visual review. |

The localized digest remains advisory and unchanged. A001–A003 apply through the
baseline and owner/scope mapping; A004–A006 through existing theme/token/density
inputs; A008–A009 through existing responsive chrome and internally scrolling
drawer; A011–A012 through lifetime/focus continuity and immutable Web Crypto
attempts; A014–A017 through deliberate forms, local conflicts and distinct async
states; A019–A020 through keyboard/naming/focus and operational component states;
A022–A027 through exact visual identities, existing selector builders, generated
boundaries, compatibility and this handoff.

A007's workbook record-create capability rules, A010's inspector dispatch,
A013's workbook mutation queue, A018's evidence lifecycle, and A021's workbook
row virtualization do not apply to membership administration. No such behavior
or framework was introduced. Applicable portions of A006, A008, A011 and A014
are scoped to this drawer; completed workbook geometry, paste and inspector
owners are preserved, not rebuilt. A015's local-error principle applies at the
membership editor/operation, rather than inventing a workbook cell for this seam.

### Visual reconciliation decision

The complete ordinary make browser-e2e-visual run at
20260908T045920Z-p1266505 passed every existing visual row; only the new membership
row failed for ten missing goldens. Its retained frontend-visual-reconciliation
artifact accounts for 181 capture intents, 171 committed active goldens, no
orphans, no ambiguous mappings, and all 26 registered fixtures resolved. The ten
missing captures belong only to web.design.visual.membership_management_visual,
scenario_71b349a3a8ce, Chromium, with no registry fixture claimed.

Accepted trigger: the explicitly authorized new membership workflow and its
new deterministic state fixtures. The controlled update creates those missing
regression inputs; it does not delete or re-account any existing golden. Missing
new inputs block ordinary validation until generated and reviewed. No existing
capture identity, renderer pin, viewport, masking or crop contract was changed.
All new captures use full viewport scope and the existing font, UUID/incident
identity, scroll and screenshot preparation helpers.

| New golden basename (all end in -linux.png) | Viewport / state |
| --- | --- |
| membership-management-loading | 1280×720, default density, initial list read. |
| membership-management-role | 1280×720, default density, explicit role draft. |
| membership-management-pending | 1280×720, admitted immutable role write. |
| membership-management-uncertain | 1280×720, unresolved role outcome. |
| membership-management-confirmed-refresh-failure | 1280×720, receipt retained through failed follow-up read. |
| membership-management-removal-narrow | 390×480, removal review with wrapped identity. |
| membership-management-removal-zoom | 1280×720, 200% browser zoom, removal review. |
| membership-management-removal-spacing | 768×640, 1.5 line height, .12em letter and .16em word spacing, 2em paragraph spacing. |
| membership-management-compact | 768×640, compact density and role editor. |
| membership-management-comfortable | 768×640, comfortable density and role editor. |

All non-zoom captures use 100% zoom. The renderer remains the pinned
visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64 profile, locale en-US,
scale factor 1, and existing dark_graphite application tokens. Final PNG review
and two fresh ordinary passes are recorded below.


### MM-04 verification continuation

All roots below are under .cartulary/test-results. These are public Make target
results, not inferred success from snapshots or Markdown.

| Command / selected evidence | Result and run root |
| --- | --- |
| module.incidents live-pages / mutation-recovery browser slice | PASS 11/11, 20260908T045727Z-p1217452. Real 103-member workflow, add/change/remove and three resulting audit events; exact replay and access/read recovery. |
| web.application ownership / transport / presentation slice | PASS 4/4, 20260908T045727Z-p1217402. |
| web.workbook lifetime slice | PASS 2/2, 20260908T045727Z-p1217412. |
| make frontend-typecheck | PASS 2/2, 20260908T045727Z-p1217572. |
| make frontend-import-boundary-check | PASS 2/2, 20260908T045920Z-p1266399. |
| make frontend-unit | 479/481 units, 20260908T045920Z-p1266412. Two policy rows failed: unused browser support exports and heading-name readiness selector policy. |
| make lint-biome | PASS 2/2, 20260908T050113Z-p1356524. |
| make json-shape-check | PASS 3/3, 20260908T050113Z-p1356134. |
| make generated-artifact-policy-check | PASS 3/3, 20260908T050113Z-p1356132. |
| make lint-markdown | PASS, 20260908T050113Z-p1356631 (adhoc/lint-markdown/tool-run-summary.json). |
| make format | PASS, 20260908T050407Z-p1368500. |
| harness.browser architecture policy slice | PASS 2/2, 20260908T050436Z-p1372972 after export cleanup. |
| web.architecture selector policy slice | PASS 2/2, 20260908T050436Z-p1372958 after scoped semantic selector repair. |
| package.ui full owner slice | PASS 10/10, 20260908T050608Z-p1422683. Existing semantic selector contracts retained. |
| make browser-e2e-visual-update | PASS 12/12, 20260908T050437Z-p1373193; all visual rows passed. |
| make format | PASS 2/2, 20260908T051325Z-p1425224. |
| web.application ownership / transport / presentation slice | PASS 4/4, 20260908T051339Z-p1429489 after final recovery review. |
| make generate | PASS, 20260908T051441Z-p1430536, including golden manifest regeneration. |
| make agent-finalize | PASS 1/1, 20260908T051500Z-p1433754, before fresh broad verification. RESULTS_DIR unset; retained-run maintenance skipped. |

The unused-export policy failure included one new membership helper and two
existing audit support helpers. Removed only their unused export keywords in
support/incidentMembershipManagement.ts, support/incidentMembershipAudit.ts and
support/administrativeAudit.ts; internal uses and completed audit behavior stay
unchanged. Focus assertions now use scoped form/region headings and the existing
incident-directory shell selector, preserving the selector ownership policy.

Final recovery review added assertions that surface close clears an observed
recovery target and dismissing conflict/uncertainty clears pre-attempt rows and
requires a fresh list before another row editor can open. A surviving transport
settlement token gets generic local feedback even after explicit departure;
no old incident identity or multi-incident state is retained.

All ten promoted membership PNGs were individually inspected. The initial,
role, pending, uncertain and confirmed-read-failure records are legible and local;
removal identity wraps at 390×480; zoom, text spacing and both density overrides
retain controls through internal scrolling. Scrolled content beneath the fixed
drawer header is intentional, and keyboard reachability is asserted separately.
The confirmed-read-failure fixture deliberately submits a reviewed no-op after
an uncertain attempt, so its no-op receipt does not claim the earlier attempt
succeeded.

The full update also rewrote 13 existing PNGs without an accepted scope trigger.
Each was inspected; none was accepted as a membership change. Restored only these
tool-created side effects to their exact execution-entry bytes using git show
and regenerated the manifest through make generate: account-settings appearance
zoom; auth focused and invalid credentials; collaboration presence; entity
mention states; incident creation expanded details; incident import canceled,
queued, running, spacing and zoom; network-flow accepted inspector; timeline
pending replay. Every existing golden therefore remains byte-identical to HEAD.
The update retains exactly the ten new membership inputs listed above.


Final pagination review reproduced a retry-history defect in the new controller:
a failed Next from page two, followed by successful retry and Previous, returned
to page one. The added assertion failed at 20260908T051746Z-p1536663 (1/2 units),
expecting the page-two cursor but receiving the first-page position. Failed reads
now retain their bounded attempted cursor history separately from the accepted
page, and retry uses that history. This is a reproduced refactor defect, not a
claim about baseline server behavior. No additional page rows are retained.
make format passed at 20260908T051819Z-p1537455.


### Changed-path inventory and compatibility review

The final scope inventory contains 60 paths: 34 tracked modifications and 26 new
files, including ten new PNGs. Only Core 01 and this handoff are changed Markdown;
the digest, domain/design direction, research essay and completed handoffs remain
unchanged. Existing goldens remain byte-identical. The completed audit production
owner is untouched; its support helper loses only an unused export. The shared
drawer change is limited to nested-dialog Escape ownership. Source and verification
ownership are authored before generated routing; generated files come only from
public Make targets. No dependency, route, permission, session owner, SQL,
migration, persisted draft, or stored-data format was added.

- `apps/web/e2e/auth-and-incident-directory.spec.ts`
- `apps/web/e2e/incident-administration.spec.ts`
- `apps/web/e2e/incident-membership-management.spec.ts`
- `apps/web/e2e/pages/deploymentAdministration.ts`
- `apps/web/e2e/support/administrativeAudit.ts`
- `apps/web/e2e/support/incidentMembershipAudit.ts`
- `apps/web/e2e/support/incidentMembershipManagement.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-comfortable-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-compact-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-confirmed-refresh-failure-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-loading-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-pending-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-removal-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-removal-spacing-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-removal-zoom-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-role-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-management-uncertain-linux.png`
- `apps/web/src/app/App.tsx`
- `apps/web/src/app/IncidentAdminPanel.test.tsx`
- `apps/web/src/app/IncidentAdminPanel.tsx`
- `apps/web/src/app/IncidentMembershipManagementPanel.test.tsx`
- `apps/web/src/app/IncidentMembershipManagementPanel.tsx`
- `apps/web/src/app/api/incidentMembershipManagementClient.test.ts`
- `apps/web/src/app/api/incidentMembershipManagementClient.ts`
- `apps/web/src/app/api/publicHttpTypes.ts`
- `apps/web/src/app/incidentMembershipManagementCharacterization.test.tsx`
- `apps/web/src/app/incidentMembershipManagementController.test.ts`
- `apps/web/src/app/incidentMembershipManagementController.ts`
- `apps/web/src/app/incidentMembershipManagementLifecycle.test.tsx`
- `apps/web/src/app/incidentMembershipManagementModel.ts`
- `apps/web/src/app/incidentMembershipManagementTestSurface.tsx`
- `apps/web/src/app/useAppRouteRuntime.ts`
- `apps/web/src/app/useIncidentMembershipManagement.ts`
- `apps/web/src/testing/incidentMembershipManagementTestSupport.ts`
- `apps/web/src/workbook/components/IncidentControlsDrawer.tsx`
- `contracts/openapi-releases/2.0.0.change-set.json`
- `contracts/openapi-source/owners/module.incidents/openapi.json`
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/protocol-ts/http-operations.v2.json`
- `docs/handoffs/incident-membership-management-refactor-handoff.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `internal/gen/contractopenapi/artifacts_gen.go`
- `internal/gen/openapioperations/catalog_gen.go`
- `internal/modules/incidents/extra_integration_test.go`
- `internal/modules/incidents/http_conformance_test.go`
- `internal/modules/incidents/httpapi/membership_management_contract_test.go`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `packages/protocol-ts/src/generated/core-http-validators.ts`
- `packages/protocol-ts/src/generated/http-operation-bindings.ts`
- `packages/protocol-ts/src/index.test.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.incidents.json`
- `tools/test_families/web.application.json`
- `tools/test_families/web.design.json`
- `tools/test_families/web.workbook.json`


Compatibility: server admission and stored state are unchanged. The newly required
OpenAPI bodies and list query projection document preconditions already enforced
by the existing routes, with the authorized DELETE owner clarification. Consumers
of previously incomplete generated bindings gain PATCH/DELETE and supported
pagination; sending a stale or missing base version remains a server rejection.
Closed-incident administration, last-admin serialization, idempotent create and
no-op PATCH retain existing behavior.

Rollback must remove the feature and its App binding together, restore the old
membership owner only with matching tests, and reverse the authored projection,
operation-selection and compatibility entries with their Make-generated outputs.
Reconcile verification routing and the ten new captures through public targets.
Keep the owner and public precondition aligned; do not relax DELETE concurrency
as a frontend rollback. No stored-data rollback or migration is required.

Retention is deliberately memory-only: page size 100, cursor history 20, one editor
and one operation record. Closing retains recovery within the same incident and
session; explicit route departure forgets it and cannot cancel server execution.
There is no automatic full collection traversal, total, search, account discovery,
invite, bulk operation, or persistent recovery. Observation is bounded by 30
seconds; unresolved transport settlement continues to block write admission.
Browser/visual evidence uses the repository's pinned Chromium lane; it is not a
claim of exhaustive assistive-technology or cross-browser certification.


### Final-source verification evidence

| Public Make command / selection | Result and run root under .cartulary/test-results |
| --- | --- |
| make frontend-unit, before the retry-history correction | PASS 481/481, 20260908T051524Z-p1438625. Both earlier policy failures resolved. |
| make frontend-typecheck | PASS 2/2, 20260908T051525Z-p1443387 and 20260908T051951Z-p1548300 after the correction. |
| make frontend-import-boundary-check | PASS 2/2, 20260908T051634Z-p1505642. |
| make lint-biome | PASS 2/2, 20260908T051636Z-p1506243 and 20260908T051952Z-p1553610. |
| make generate-drift | PASS 4/4, 20260908T051633Z-p1504955. |
| make test-slice OWNER=web.application, membership ownership / transport / presentation | PASS 4/4, 20260908T051848Z-p1541778 including the reproduced retry-history assertion. |
| make test-slice OWNER=web.workbook, membership lifetime | PASS 2/2, 20260908T051849Z-p1542003. |
| make agent-finalize | PASS 1/1, 20260908T051907Z-p1543081. RESULTS_DIR unset; retained-run maintenance skipped. |
| make json-shape-check | PASS 3/3, 20260908T051952Z-p1553104. |
| make generated-artifact-policy-check | PASS 3/3, 20260908T051952Z-p1553119. |
| make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off | First fresh ordinary PASS 12/12, 20260908T051523Z-p1436946. |
| make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off | Second fresh ordinary PASS 12/12, 20260908T051950Z-p1546614, after retry-history correction. |
| make lint-markdown | PASS, 20260908T052159Z-p1598166, before completing MM-05. |
| git diff --check and manual scope review | PASS; 59 paths, no digest or existing PNG changes. |

Both ordinary visual runs used the same promoted golden manifest, without update
mode or cache reuse. All visual rows passed, with no skipped or flaky browser
cases. The later controller correction changes cursor retry state only; the second
run covers the settled implementation and all ten accepted membership captures.


The final ten-row module.incidents browser slice at
20260908T052314Z-p1599784 passed 15/17 graph units. All four membership-management
rows, all four completed audit-browser rows, and the retained incident-admin flow
passed. Only the old incident-selection row failed: it expected the retired
always-open role input. Updated that fixture to assert the saved role and open
Change role before its existing editor/access-loss assertions. This is selector
migration in auth-and-incident-directory.spec.ts, not a production access-loss
failure. A focused retry is required before closing MM-04.

The final-source make frontend-unit run passed 481/481 at
20260908T052316Z-p1600045, and make generate-drift passed 4/4 at
20260908T052317Z-p1600547. No frontend or generator failure remains.


The migrated incident-selection access-loss row passed 11/11 graph units at
20260908T052709Z-p1702882. It still proves incident membership loss returns to the
directory while the account remains signed in and another permitted incident can
open. make format passed at 20260908T052646Z-p1698566; the selector policy slice
passed 2/2 at 20260908T052710Z-p1703115. Together with the preceding nine passing
browser rows, the complete selected ten-journey regression set now has passing
evidence; no production behavior changed in this retry.


### MM-04 exit

All required journeys and failure transitions have passing owner-routed evidence.
The final web.design accessibility slice (membership management, completed audit,
and account-menu keyboard/focus rows) passed 11/11 graph units at
20260908T052852Z-p1750346, with all three browser cases passing and none skipped
or flaky. Final Biome passed 2/2 at 20260908T052853Z-p1750615.

Production paths, transport, state, presentation and App/drawer integration are
complete. Backend concurrency and closed-incident rules, 100+ live service members,
real mutations and audit visibility, duplicate/replay/conflict/uncertainty,
permission/lifetime/departure, and accessible operational states are covered by
the command evidence above. Existing administration, audit, account-menu and
IncidentAdminPanel regressions pass. No blocked dependency or unresolved product
failure remains. Next: MM-05 finalization and completed-handoff scope audit.

The selected final browser command used make service-backed-test-slice with
OWNER=module.incidents and these ROWS (the access-loss row was retried alone after
its fixture migration):

- module.incidents.browser.membership_management_live_pages
- module.incidents.browser.membership_management_mutation_recovery
- module.incidents.browser.membership_management_departure
- module.incidents.browser.membership_management_self_actions
- module.incidents.browser.incident_admins_manage_memberships_on_the_ordina_f26145a162
- module.incidents.browser.membership_audit_access_recovery
- module.incidents.browser.membership_audit_read_recovery
- module.incidents.browser.membership_audit_real_browsing
- module.incidents.browser.membership_audit_response_inspection
- module.incidents.browser.a_selected_incident_whose_membership_is_removed_21054ae899

The accessibility command used the same public target with OWNER=web.design and
ROWS=web.design.accessibility.membership_management_accessibility,
web.design.accessibility.membership_audit_browsing,
web.design.accessibility.account_menu_accessible_keyboard_and_focus_state_3c31f28451.
Focused frontend selections used the authored membership_management_ownership,
membership_management_transport and membership_management_presentation rows under
web.application.regression, and membership_management_lifetime under
web.workbook.regression. Run roots retain exact selections and per-row results.


### MM-05 exit

make agent-finalize passed 1/1 at 20260908T053050Z-p1797080 before the final
static verification. Its unit-artifacts/finalize-summary.json explicitly records
results_dir_status=skipped and results-dir-not-provided for retained run closure,
scheduler ordering and warm-check timing maintenance. RESULTS_DIR remained unset:
no qualifying successful full warm check from this exact source exists. These
maintenance checks were skipped, not represented as passing.

Final make frontend-typecheck passed 2/2 at 20260908T053132Z-p1800331 and make
frontend-import-boundary-check passed 2/2 at 20260908T053133Z-p1800729. The full
481/481 frontend run, focused backend/protocol/UI slices, final generation/drift,
JSON/generated-policy/Biome checks, ten browser journeys, three accessibility
rows, ten individually reviewed new goldens, and both fresh ordinary visual
passes are retained above. No required seam check remains skipped or failing.

Repo-wide check/ci/release, unrelated backend/browser suites, security audits and
additional renderer lanes were not run: verification followed the narrow owner
routes and broadened to the full frontend and visual suites for shared integration
risk. No claim of a full warm check or release readiness is made. Earlier failed
runs and their corrected fixture, contract, policy, state or infrastructure causes
remain recorded rather than erased.

Final manual scope review matches the 60-path inventory: 34 tracked modifications
and 26 new files, all within the authorized seam and its projections/evidence.
The branch remains main at f456fe2599b264f779d2874f305ca7c55d2976f2; changes remain
unstaged, with no commit, reset, push or deployment. The old membership production
path is retired; App renders the sole new feature owner. Summary, promoted fields,
lifecycle and preferences retain their owners, and completed audit production
code remains intact. No executable product verification was made dependent on
Markdown. git diff --check passed before completing this tracker.

All MM workstreams are DONE. Post-completion make lint-markdown passed at
20260908T053234Z-p1801511 (adhoc/lint-markdown/tool-run-summary.json).
Post-completion git diff --check and manual review against the completed 60-path
inventory passed. The same checks are repeated after recording this evidence;
no production or projection changes follow the passing final-source runs.

Next action: **none**.
