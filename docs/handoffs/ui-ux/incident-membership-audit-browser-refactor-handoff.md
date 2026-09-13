# Incident Membership Audit Browser Refactor Handoff

## Baseline, authority, and scope

Implementation entry: `main`, HEAD `610e8d15e54b4658dfdf3fd4d09c441162ffa874`,
clean tree. Root AGENTS.md is the only applicable instruction file. No reset,
commit, push, deployment, or unrelated seam is authorized.

Core 01 §3.3.5.1A / REQ-01-603–607 owns incident scope, exact filters, resource
shape, additive read vocabulary, ordering and pagination. Core 01 §3.3.6–7 and
§18A own public errors, live cursor traversal and explicit instants. Core 04
§1.1.1, REQ-04-017, 028–030 and 123 own session and incident authorization.
Core 03 §13 preserves keyboard ownership; design §§7, 12, 14–15 govern local
presentation and evidence. Domain governs vocabulary. The digest's localized
read order and deployment audit handoff were read as advisory/implementation
evidence; neither authorizes incident policy. The NLSpec research document is
not an adopted behavior owner. No inspected adopted-owner contradiction exists.

Allowed paths: incident audit implementation/tests under apps/web/src/app and
its api directory; pure audit helpers under apps/web/src/shared; minimal App,
workbook drawer/presentation/lifetime contracts; related testing fixtures and
apps/web/e2e journeys; authored UI selectors; tools source ownership, test
families and browser/visual routing; Make-generated downstream artifacts;
reviewed incident audit goldens; this handoff. No membership mutation redesign,
other Incident Controls redesign, backend/writer/storage/retention/auth change,
route, dependency, lockfile, generic framework, or digest edit is included.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| MA-01 — Baseline, authority, characterization | DONE | Five defects reproduced; remaining lifetime risks are explicit. |
| MA-02 — Query, client, read/access lifecycle | DONE | Focused model/controller/client checks pass. |
| MA-03 — Presentation and Incident Controls integration | DONE | One App controller; presentation and lifetime tests pass. |
| MA-04 — Browser, accessibility, visual regressions | DONE | Real paging, keyboard/recovery, ten reviewed images and two fresh visual passes. |
| MA-05 — Final validation and completed handoff | DONE | All required evidence passes; terminal documentation and scope audit retained separately. |

Only the current row is IN_PROGRESS; a blocked dependency stops progression.

## Discovery and decisions

App renders IncidentAdminPanel through WorkbookIncidentControlsPresentation.
The drawer unmounts on close; useIncidentControlsDrawer owns Escape return focus
and section selection. The old audit loader constructs requests from editable
filter refs, accumulates pages, lacks its own generation fence and formats values
with String. These are source findings until characterization reproduces them.
AppSessionController owns session acceptance/retirement and its recoveryPort;
useWorkbookAuthorizationState owns workbook role presentation. Generated
listIncidentMembershipAuditEvents already exposes all six filters and uses
administrative_audit_future_vocabulary response validation. Authored inputs are
module.incidents OpenAPI, platform.audit resource schemas, and protocol-ts
http-operations.v2.json. No projection repair is currently indicated.

User-selected retention: draft and applied filters only across ordinary close or
section exit within one incident/session. Clear rows, details, cursors and reads;
return observes access and reads page one. A rejected-cursor marker requires
explicit Reload first page even after return. Authority loss clears everything.
Use one page of 100 events, 20 prior cursor descriptors, one expanded event,
30-second total read observation and explicit retry. No persistent preferences,
multi-incident cache, totals, offsets, snapshot promises or accumulation shim.

Planning ran make help/help-all and task-guide for module.incidents,
web.application, web.workbook, web.design, package.ui, package.protocol_ts and
harness.browser. Existing baseline evidence at the unchanged HEAD:

- make test-slice OWNER=module.incidents ROWS=module.incidents.frontend.membership_audit_product_surface:
  PASS 2/2 graph units, `.cartulary/test-results/20260908T013020Z-p4088024`.
- make service-backed-test-slice OWNER=module.incidents ROWS=module.incidents.integration.membership_audit_route_is_scope_safe_and_keyset_8e31c7f24a:
  PASS 3/3, `.cartulary/test-results/20260908T013020Z-p4088034`.
- git diff --check: PASS. These existing checks are not race reproductions.

## Execution evidence

MA-01 implementation entry revalidated the same clean branch and HEAD. The
frontend remains React 19 / TypeScript 6 / Vite 8 / pnpm 10.33. Verification
routes are authored under module.incidents (route/product), web.application
(controller/presentation/App), web.workbook (drawer/role integration), web.design
(accessibility/visual), and package.ui/protocol_ts (selectors/bindings).

## Compatibility and rollback

Frontend read workflow only. No runtime API or stored-data migration. Rollback
restores this seam's frontend, routing, tests and reviewed images together;
unlimited accumulation requires no compatibility shim.

## Completion

All five workstreams are complete. Required product, browser, accessibility,
visual and static checks pass. Retained-run maintenance was skipped for the
explicit RESULTS_DIR reason recorded in MA-05. Terminal completed-byte review
results are retained in .cartulary/incident-membership-audit-final-review.json.

## MA-01 exit

Registered six characterization cases through module.incidents and frontend source
ownership. make generate initially failed because new selector titles were not
ASCII-sorted (20260908T014033Z-p4108562); corrected the authored catalog, then
make generate passed at 20260908T014124Z-p4111866.

CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidents
ROWS=module.incidents.frontend.membership_audit_characterization intentionally
failed 1/2 graph units at 20260908T014153Z-p4115020. Its row stderr JSON records
five product assertion failures against unchanged production: initial empty
message during loading; draft filters in continuation; duplicate Apply admission;
late initial success replacing newer results; JSON string/null ambiguity.
The cross-section visible error assertion passed: it did not reproduce that risk.
The duplicate-page assertion follows the duplicate-Apply assertion and was not
reached. Close, incident, role, session and ignored-abort boundaries remain risks
until the focused lifecycle tests exercise them. No demonstrated security exploit
is claimed. The scope and adopted owners are verified; no blocked dependency.
Next: MA-02 model, typed client, shared pure semantics and controller tests.

## MA-02 exit

Added incidentMembershipAuditModel/controller and the narrow typed client, with
18 focused query/controller/client cases. Shared auditReadValues owns extracted
pure filter/instant normalization and JSON formatting; deployment query and panel
behavior stay distinct. Tests exercise one-page/20-cursor bounds, 25-page traversal,
query separation, coalescing, old success/errors, ignored cancellation, access-only
cursor recovery, observation failure, scope/actor/session/role changes and JSON
validation. No generated protocol or authored public contract changed.

make generate passed at 20260908T015234Z-p4117921. Initial focused run
20260908T015300Z-p4121909 passed 4/6 units including query and deployment
regressions; the two failing rows exposed test-fixture errors (missing JSON
Content-Type and recovery returning the prior actor for a new actor). Corrected
fixtures. The controller/client rerun passed 3/3 at 20260908T015342Z-p4123104.
Typecheck first found an omitted shared AuditField import, then passed 2/2 at
20260908T015342Z-p4123182. Import boundaries passed 2/2 at
20260908T015246Z-p4121161. No skipped focused case or blocked dependency.
Next: MA-03 presentation and existing lifetime/recovery integration.

## MA-03 exit

App retains one incident controller above drawer mounts; existing session
acceptance synchronously observes membership changes. Narrow section-transition
notifications suspend reads before unmount, and recovered roles flow to existing
workbook authorization presentation. The new native form/paged event inspection
replaces the audit loader, state and rendering in IncidentAdminPanel. Membership
mutations and other sections retain their original owners. Shared value semantics
keep deployment authorization/lifecycle distinct.

Moved the existing audit placement test without dropping its routed coverage;
added invalid-form, unavailable/stale/retry and focus-continuity tests. All six
characterization cases now pass, including duplicate page admission. Four tests
exercise actual AppSessionController and drawer hooks under Strict Mode: ordinary
filter retention, visible role loss/regrant, hidden incident loss, and retirement.

make generate passed at 20260908T020041Z-p4130025 and
20260908T020508Z-p4135733. Initial format/typecheck errors were local missing
imports, accessible-region semantics, and unsubscribe return types, corrected
before exit. Concurrent format/generation caused an artifact ENOENT while SQL
generation replaced a file; sequential make format passed at
20260908T020827Z-p4139614. Generation/format mutations will run sequentially.

Characterization plus placement passed 3/3 at 20260908T020136Z-p4134065. Expanded
presentation tests first exposed test-fixture mistakes (reused consumed Response
and omitted test imports) at 20260908T020849Z-p4144024, corrected; product surface
passed 2/2 at 20260908T020938Z-p4145841. Real lifetime integration passed 2/2 at
20260908T020849Z-p4144035. Typecheck passed 2/2 at
20260908T020938Z-p4145911; import boundaries passed 2/2 at
20260908T020053Z-p4133193. No blocked dependency.
Next: MA-04 public service-backed pagination, deterministic browser recovery,
keyboard/responsive and reviewed visual evidence.

## MA-04 evidence in progress

Added a public-fixture 105-role-change browser journey with exact actor/action/
target filters, real 100+5 paging, accepted-query continuation despite draft edits,
nanosecond inclusive/exclusive route boundaries, inspection, and non-admin denial.
Deterministic HTTP fixtures exercise overlapping query/page reads, cursor recovery
across close/reopen, drawer/section/document suspension, transient observation,
visible role downgrade, hidden incident (including deployment-admin status) and
session loss. Browser fixtures use public membership PATCH directly because that
mutation has no generated operation binding; the production audit read remains
on the existing generated binding. No mutation projection expansion is needed.

The first service-backed selection at 20260908T021639Z-p4159347 failed 12/14:
all three new browser scenarios and the existing service route slice passed;
the adjacent administration journey failed before workbook readiness while
concurrent Vite builds replaced its asset files. Build stderr lists missing chunk
files. Subsequent browser targets run sequentially. The focused accessibility
slice passed 11/11 at 20260908T021640Z-p4159567, covering native form errors,
keyboard inspection, names/focus, reachable controls at 1280x720, 1024x720,
768x640, 640x480 and 390x480, 200% zoom, text spacing and drawer focus return.

Browser fixture typecheck first exposed unsupported mutation binding usage and
literal-union/Node fixture typing errors at 20260908T021612Z-p4155831; corrected.
Typecheck passed 2/2 at 20260908T021731Z-p59650. Six focused incident and adjacent
deployment frontend rows passed 7/7 at 20260908T022003Z-p108431.

The first ordinary visual run 20260908T021730Z-p58097 failed 10/12: all existing
33 visual scenarios passed, and the new scenario completed all ten intended
captures with missing-new-golden failures only. Final reconciliation rejected
human-readable new scenario IDs because its schema requires 12 hexadecimal
characters. Corrected the narrow authored IDs; generation passed at
20260908T022159Z-p114580. No golden mutation preceded reconciliation review.
Additional browser coverage now exercises wrong incident scope, malformed paging,
unknown additive action/target vocabulary, long inert content, all JSON types and
redaction. Final visual capture review and ordinary-run reconciliation follow.

Sequential service rerun 20260908T022213Z-p117612 passed the three established
browser journeys, the adjacent Incident Controls journey and service route slice.
The added response-inspection scenario correctly rejected its unsorted field-path
fixture, causing the overall 12/14 result. Corrected fixture ordering; that browser
row passed 11/11 at 20260908T022712Z-p225236. This confirms generated validation
remains active. Expanded late-access-observation matrix passed 2/2 at
20260908T022525Z-p224089; workbook surfaces passed 2/2 at
20260908T022407Z-p211541. An earlier guessed workbook selector was rejected before
execution; the actual catalog row above was used. Import boundaries passed 2/2
at 20260908T022307Z-p164485 and typecheck 2/2 at 20260908T022406Z-p211032.

The corrected ordinary visual run 20260908T022348Z-p168062 completed all ten
new capture intents and all 33 existing visual scenarios. Its retained
browser-e2e-visual/frontend-visual-reconciliation.json accounts for 171 intents,
161 baseline goldens, all 26 registered fixtures, no orphans/ambiguous mappings,
and exactly ten missing new incident-audit goldens. The 10/12 result is confined
to those new snapshots and the corresponding target summary. Existing images
must remain byte-identical. Candidate generation is limited to the explicitly
authorized new audit states; no existing golden may conceal a failure.

## Visual refresh record

Accepted trigger: the authorized incident audit read workflow and its accessible
bounded presentation. Exact owner row web.design.visual.membership_audit_browsing,
scenario scenario_d4214d0e8b76, project chromium. Nonregistry capture identities:
membership-audit-loading, inspected, stale, empty, cursor-recovery,
inspected-narrow, inspected-zoom, inspected-spacing, compact, comfortable.
Golden filenames append -linux.png in apps/web/e2e/workbook.visual.spec.ts-snapshots.
No new D-VFIX claim, crop, renderer, font, mask or theme is introduced. Existing
full-viewport capture and dynamic identity/time normalization remain; exact raw
timestamp precision is independently asserted by unit and service/browser checks.

Normal captures: 1280x720, 100% zoom. Narrow: 390x480. Zoom: 1280x720 at 200%.
Text spacing and explicit compact/comfortable account densities: 768x640, 100%.
Inspected states align Published field changes inside the existing drawer scroll
owner. Computed density font size and row padding are compared with existing
exported tokens. All capture state assertions completed before candidate creation.
Candidate review and the two fresh ordinary validation runs remain pending.

## Final ownership and acceptance map

| Adopted owner / existing boundary | Change and evidence |
| --- | --- |
| Core 01 §3.3.5.1A, REQ-01-603–607 | Incident model/client use six exact route filters, inclusive lower/exclusive upper instants, server keyset order and paging; generated operation and validation remain authoritative projections. Real 105-event route/UI traversal and current-vocabulary parity tests. |
| Core 01 §3.3.6–7, timestamp_instant_v1 | Public identity-based errors, captured query/cursor context, explicit rejected-cursor recovery, nanosecond normalization with explicit offset. No inferred timezone, total, snapshot or offset. |
| Core 04 REQ-04-123 and session/incident-access requirements | Incident-admin admission only; existing session owner accepts authority, existing workbook recovery observes access. Generation/lifetime/surface fences, bounded deadlines and mandatory clearing; role loss stays in workbook, hidden incident and session loss delegate existing recovery. |
| Core 03 keyboard ownership; design §§7, 12, 14–15 | Native form, stable page/detail controls, labeled faithful Before/After JSON, inert event text, explicit redaction, restrained announcements, narrow/short/zoom/spacing and density evidence. Drawer retains Escape/return focus and scrolling. |
| Domain vocabulary | Incident-scoped membership audit remains distinct from deployment administrative audit. No new search, mutation, retention, export or authorization vocabulary. |
| Application / workbook source ownership | App retains one controller. IncidentAdminPanel no longer owns audit reads, query state or rendering. Minimal drawer transition and recovered-role ports preserve surrounding owners. |
| Authored verification routing / generated policy | New source ownership, family rows, stable detail selector and browser batches; existing product row coverage moved intact. Make regenerates topology and visual manifest. Tests and generators never depend on Markdown. |

The common auditReadValues module contains only query primitives, exact instant
normalization and JSON formatting. Deployment arrays, query/controller lifetimes,
authorization and materialization remain deployment-owned. No mode-flag framework
or duplicate session controller was introduced. Reused pure styles and design
variables preserve the current theme and density selection.

| Digest acceptance | Disposition within this seam |
| --- | --- |
| A001–A003 | Applied: adopted owner map, actual baseline/callers/generated boundaries and one incident audit seam recorded. Advisory maps revalidated against current source. |
| A004–A006 | Applied locally: existing tokens/theme and workbook density; no new palette, preference or registry. Grid geometry outside this read panel is unchanged. |
| A007 | Not applicable: no product create entry point or mutation payload change. Public test-fixture writes only establish browser evidence. |
| A008–A009 | Applied locally: existing responsive drawer and scrolling preserved; controls reachable through supported and below-minimum viewport/zoom/spacing tests. No shell threshold change. |
| A010 | Not applicable: workbook inspector dispatcher is unchanged; this panel's event detail is local read-only inspection. |
| A011 | Applied: stable identities, focus continuity on refresh/paging, explicit drawer focus return. Ordinary close retains only filters by the user's selected policy. |
| A012–A015 | Not applicable: transaction IDs, edit queues, cell editing and conflict behavior remain existing owners; adjacent regression checks retained. |
| A016–A017 | Applied: explicit initial/paging/refresh/empty/filtered-empty/stale/invalid/unavailable/denied states and authorized previous-query context. Confirmed authority loss removes protected data. |
| A018 | Not applicable: evidence lifecycle and preview are unchanged. |
| A019 | Applied: native keyboard controls, names, focus, announcements, inert values, redaction, readable layouts and owner-routed accessibility. |
| A020 | Applied to local states and event inspection; unrelated component variants are unchanged. |
| A021 | Not applicable to workbook virtualization: the audit is a bounded native list of at most 100 events. No grid import, unbounded accumulation or automatic full-history load. |
| A022 | Applied to ten exact nonregistry capture identities; existing D-VFIX fixtures retain their identities and golden bytes. No new registry claim. |
| A023 | Applied: event/detail IDs use authored stable audit-event selectors; controls use semantic roles/names and existing surface selectors. |
| A024–A026 | Applied: no Markdown dependency in executable artifacts, no hand-edited generated roots, no API/storage/backend/auth-policy invention. |
| A027 | This handoff records all workstreams, evidence, compatibility/rollback and final scope audit. |

## Deliberate limits

The bounded history holds twenty prior descriptors, one page of at most 100
published events, and one expanded event ID. Older backward links are explicitly
unavailable after eviction; First page starts a fresh live traversal. Requests
have a 30-second deadline including access observation. Retry repeats the captured
intent; no timer polls for authority or audit updates. Authority changes are
observed through existing accepted-session updates and bounded activation/read
checks. Temporary failure is not confirmed revocation.

Ordinary close, section exit and document hiding retain only draft/applied filters
inside the same incident/session, plus a rejected-cursor recovery marker when
needed. Return rechecks access and reads page one, except rejected cursors require
explicit Reload first page. No rows, detail, cursor history, error or pending read
survives suspension. Incident/actor/session change or confirmed authority loss
clears all feature data. These are frontend memory limits, not server retention,
security vulnerability claims or snapshot guarantees.

Candidate update 20260908T022841Z-p276140 was not promoted. The audit visual row
passed all ten captures, including computed density assertions, but the unrelated
existing dismissed-mention fixture failed (same failure previously recorded by the
completed deployment audit handoff). Final artifact admission also reported an
owner-only scratch-file error. Tracked snapshot/manifest bytes remained unchanged.
The incomplete run's apparent orphan belongs to the interrupted mention capture;
it is not eligible for deletion or refresh to conceal that assertion.

All ten audit candidate images were individually inspected. Labels, quoted JSON,
redaction and wrapped long paths remain readable; narrow/zoom/spacing states use
owned scrolling and preserve keyboard reachability. Review found that inherited
inline button colors visually overrode disabled-page CSS. Corrected this local
style with the existing disabled ink/surface tokens and explicit precedence,
matching the established deployment panel convention. No behavioral or unrelated
visual expansion. A fresh candidate run follows, then image review and two fresh
ordinary final-source passes.

The fresh update passed 12/12 at 20260908T023333Z-p327978. Reconciliation passed
with 171 active intents/goldens, 26 registered fixtures, zero missing, orphaned or
ambiguous mappings. All ten promoted incident audit images were individually
reviewed again: disabled navigation is visibly distinct, loading never claims
empty results, stale/recovery context is explicit, field labels and all JSON
forms remain readable, redaction is explicit, and narrow/zoom/spacing/density
captures preserve wrap and owned scrolling. Review outcome: accept intentional
incident-audit states. No remaining local layout defect was found.

Update mode rewrote twelve unrelated existing images; each was restored exactly
from the recorded HEAD. make generate passed at 20260908T023739Z-p382961 to
regenerate the visual manifest from the accepted ten additions and unchanged
baseline images. No generated manifest was hand-edited. The completed backward
history traversal test passed 2/2 at 20260908T023555Z-p381588. Source and accepted
golden bytes are now frozen for the two fresh ordinary visual passes.

## Changed path inventory

The final source/golden inventory is bounded to the following authored, generated
and reviewed paths (the handoff itself is documentation evidence):

- `apps/web/e2e/incident-membership-audit.spec.ts`
- `apps/web/e2e/pages/deploymentAdministration.ts`
- `apps/web/e2e/support/incidentMembershipAudit.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-comfortable-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-compact-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-cursor-recovery-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-empty-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-inspected-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-inspected-narrow-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-inspected-spacing-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-inspected-zoom-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-loading-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/membership-audit-stale-linux.png`
- `apps/web/src/app/App.tsx`
- `apps/web/src/app/DeploymentAuditPanel.tsx`
- `apps/web/src/app/IncidentAdminPanel.test.tsx`
- `apps/web/src/app/IncidentAdminPanel.tsx`
- `apps/web/src/app/IncidentMembershipAuditPanel.test.tsx`
- `apps/web/src/app/IncidentMembershipAuditPanel.tsx`
- `apps/web/src/app/administrativeAuditModel.ts`
- `apps/web/src/app/api/incidentMembershipAuditClient.test.ts`
- `apps/web/src/app/api/incidentMembershipAuditClient.ts`
- `apps/web/src/app/incidentMembershipAuditCharacterization.test.tsx`
- `apps/web/src/app/incidentMembershipAuditController.test.ts`
- `apps/web/src/app/incidentMembershipAuditController.ts`
- `apps/web/src/app/incidentMembershipAuditLifecycle.test.tsx`
- `apps/web/src/app/incidentMembershipAuditModel.test.ts`
- `apps/web/src/app/incidentMembershipAuditModel.ts`
- `apps/web/src/app/useIncidentMembershipAudit.ts`
- `apps/web/src/shared/auditReadValues.ts`
- `apps/web/src/shared/workbookShellContracts.ts`
- `apps/web/src/testing/incidentMembershipAuditTestSupport.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/components/WorkbookIncidentControlsPresentation.tsx`
- `apps/web/src/workbook/hooks/useIncidentControlsDrawer.ts`
- `docs/handoffs/incident-membership-audit-browser-refactor-handoff.md`
- `packages/ui-contracts/src/applicationSelectors.ts`
- `packages/ui-contracts/src/index.ts`
- `packages/ui-contracts/src/workbook-interaction-selectors.test.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.incidents.json`
- `tools/test_families/web.application.json`
- `tools/test_families/web.design.json`

First fresh ordinary final-source visual run passed 12/12 at
20260908T023802Z-p386106. Its reconciliation passes for all 171 active captures/
goldens and 26 registered fixtures, with zero missing, orphaned or ambiguous
entries. This includes the ten new audit images and unchanged deployment/adjacent
visual fixtures. Second fresh ordinary pass follows before MA-04 closes.

## MA-04 exit

Second fresh ordinary final-source visual run passed 12/12 at
20260908T024105Z-p435648. Together with 20260908T023802Z-p386106, both reconcile
171 active captures/goldens and all 26 registered fixtures with zero missing,
orphaned or ambiguous entries. All ten new incident images were inspected; all
161 baseline golden bytes remain unchanged. Required public 105-event paging,
exact boundaries, deterministic races/recovery, malformed/additive inspection,
keyboard, responsive and density cases have passing evidence above. Adjacent
Incident Controls and deployment visual fixtures pass. No blocked dependency or
unresolved local product defect remains. Next: MA-05 finalization before broader
verification, final compatibility/scope review and completed handoff bytes.

## MA-05 finalization and verification

`env -u RESULTS_DIR make agent-finalize` passed 1/1 at
20260908T024410Z-p484447, before broader final checks. Its
unit-artifacts/finalize-summary.json reports zero generated updates and passing
JSON/schema, catalog, tier coverage and generated-structure drift checks.
RESULTS_DIR was unset because no qualifying exact-source successful full warm
check run exists. Retained-run selection, canonical retained-evidence validation,
scheduler event/timing validation and performance-evidence maintenance were
therefore skipped with results-dir-not-provided. No older-run override was used.

Final owner slices use CARTULARY_HARNESS_CACHE_MODE=off. Counts below are graph
units, not inferred assertion counts. Run IDs resolve under
.cartulary/test-results; per-row/browser reports and unit logs remain there.

| Final public Make command | Result | Run ID |
| --- | --- | --- |
| make test-slice OWNER=web.application | PASS 85/85; all application rows, including deployment audit/session/Account/directory/import/reference-pack regressions | 20260908T024501Z-p487874 |
| make test-slice OWNER=web.workbook | PASS 164/164; all workbook rows, including surfaces, authorization, layout and adjacent editing owners | 20260908T024501Z-p487853 |
| make service-backed-test-slice OWNER=module.incidents with the exact six rows below | PASS 14/14; four audit browser scenarios, adjacent membership administration and route scope/filter/keyset slice | 20260908T024501Z-p487870 |
| make frontend-typecheck | PASS 2/2 | 20260908T024501Z-p488049 |
| make frontend-import-boundary-check | PASS 2/2 | 20260908T024520Z-p506684 |
| make lint-biome | PASS 2/2 | 20260908T024522Z-p507685 |
| make test-slice OWNER=package.ui | PASS 10/10 | 20260908T024534Z-p517076 |
| make test-slice OWNER=web.architecture | PASS 12/12 | 20260908T024534Z-p516942 |
| make generated-artifact-policy-check | PASS 3/3 | 20260908T024534Z-p516733 |
| make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.generated_http_operation_bindings | PASS 2/2 | 20260908T024654Z-p609829 |
| make test-slice OWNER=platform.audit ROWS=platform.audit.unit.contract_and_safe_projection | PASS 1/1 | 20260908T024656Z-p610547 |
| make test-slice OWNER=module.auth ROWS=module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2 | PASS 1/1 | 20260908T024657Z-p611260 |

Exact final incident service-backed command (and initial/repeated selections above
use these row IDs, omitting response_inspection until that row was added):

```sh
CARTULARY_HARNESS_CACHE_MODE=off make service-backed-test-slice OWNER=module.incidents ROWS=module.incidents.browser.membership_audit_real_browsing,module.incidents.browser.membership_audit_read_recovery,module.incidents.browser.membership_audit_access_recovery,module.incidents.browser.membership_audit_response_inspection,module.incidents.browser.incident_admins_manage_memberships_on_the_ordina_f26145a162,module.incidents.integration.membership_audit_route_is_scope_safe_and_keyset_8e31c7f24a
```

The full module.incidents test slice passed 34/34 at 20260908T024534Z-p517028,
including existing authored route/schema/frontend/support coverage. Final review
strengthened the existing controller matrix so old success/error arrives after
role loss *and regrant*, and added an explicit late session error after disposal.
No production, browser fixture, API or golden bytes changed. The role-regrant
case passed 2/2 at 20260908T024954Z-p661473; the combined controller row passed
2/2 at 20260908T025121Z-p666614.

The adjacent deployment browser selection initially failed 9/11 at
20260908T024758Z-p617357 with infra/service_readiness_timeout before any browser
scenario executed. Its run-summary.json and browser group unit result classify
the failure; no product assertion or golden change is implicated. A fresh
isolated retry follows. Full accessibility and final static checks remain pending.

Deployment audit's fresh browser retry passed 11/11 at
20260908T025142Z-p667287 with the exact command:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make service-backed-test-slice OWNER=module.auth ROWS=module.auth.browser.administrative_audit_read_recovery,module.auth.browser.administrative_audit_real_browsing
```

The failed predecessor's service-scope artifact identifies isolated object-store
startup readiness (transport-unreachable/operation-timeout), before browser startup.
No production or harness fix was needed. Final typecheck after the test-only
strengthening passed 2/2 at 20260908T025212Z-p711544. The two ordinary visual
passes remain after the last production/fixture/golden change; subsequent changes
only strengthened controller assertions and this evidence record.

Full `make browser-e2e-a11y` passed 12/12 at 20260908T025317Z-p714720:
all 29 scenarios passed, including membership audit, deployment audit and adjacent
workbook/application controls. Final `make lint-biome` passed 2/2 at
20260908T025318Z-p716348 after the test strengthening. Standalone generated drift
and completed-byte documentation/scope checks follow without any product edit.

## Skipped checks and compatibility closure

No full `make check`, release, CI, vulnerability/dependency scan, unrelated backend
integration matrix, or workbook performance/measurement suite was run: this seam
changes only the bounded frontend read workflow and its verification. All affected
owner unit slices, required public service/browser journeys, full accessibility,
full visual suites, typecheck, import boundaries and applicable frontend lint ran.
The full conformance/owner evidence publication audit was not selected: there is no
new Core 05 claim, and no qualifying full warm check root exists. Retained-run
maintenance exclusions are precisely recorded in MA-05 finalization above.

Compatibility remains frontend-only: no runtime API, route, stored-data migration,
backend authorization, writer or retention change. The old accumulation behavior
requires no compatibility shim. Rollback restores the incident feature and minimal
App/workbook ports, shared pure helper extraction, moved tests, authored selectors/
routing, generated topology/visual manifest and ten new goldens together. The 161
existing goldens stay at baseline. No reset, commit, push or deployment occurred.

Final audit covers the 48-path inventory, exact branch/HEAD, generated boundaries,
removed old audit loader, unchanged excluded owners/digest/dependencies, unchanged
baseline images, and the completed handoff bytes. The terminal documentation and
scope review record is retained at .cartulary/incident-membership-audit-final-review.json;
it is human implementation evidence, not a product-test or generator dependency.

## MA-05 exit

Standalone `make generate-drift` passed 4/4 at 20260908T025651Z-p764150;
`make json-shape-check` passed 3/3 at 20260908T025835Z-p767690. Final branch is
main at 610e8d15e54b4658dfdf3fd4d09c441162ffa874. The dirty tree contains exactly
the 48 authorized paths listed above. No unrelated edits, digest/dependency or
backend changes, stored-data/API migration, commit, push, reset or deployment.
All baseline PNG bytes remain unchanged.

One feature owns incident audit query/read state and its typed boundary; no audit
transport remains in presentation or IncidentAdminPanel. Stable controls retain
focus; removing a focused event inspection or completed Retry control returns
focus to the existing Refresh control when it would otherwise fall to the body.
Drawer Escape and return focus remain workbook-owned. This local recovery choice
also covers the transient Retry control and is exercised by the accessibility
journey. No unrelated focus owner changed.

All required workstreams have passing evidence and no blocked dependency remains.
After this DONE entry, run `make lint-markdown`, `git diff --check`, and the final
48-path/baseline-image/owner-boundary scope audit against these completed bytes.
Their exact command result, Markdown run root and completed-handoff SHA-256 are
recorded in .cartulary/incident-membership-audit-final-review.json so recording
the terminal results does not change the reviewed Markdown bytes. This review
record is outside product tests/generators and does not confer conformance status.

Next action: none.
