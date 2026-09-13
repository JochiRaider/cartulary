# Incident and session revocation remediation

## Baseline and authority

Execution baseline: clean `main`,
`bc7664afdf49b14870c327b20732209169a23586`. AGENTS.md and the adopted
Testing Harness NLSpec were read. Core 01 §3.3.6.2 and REQ-01-277, Core 03
REQ-03-299 and REQ-03-100, and Core 04 REQ-04-017 and REQ-04-156 govern
revocation scope, protected presentation, retained work and Evidence access.
Harness TH-HARNESS-REQ-650/653 govern routing and browser readiness, not product
authorization. No executable evidence reads Markdown.

Confirmed decision: an incident revocation immediately returns to the directory;
account authorization refresh is independent, and incident reentry is explicit.
The existing wire event and four reasons remain. No internal compatibility alias,
new persistence, dependency, migration, commit, push or deployment is included.

Planning evidence at this exact HEAD:

- `make test-slice OWNER=web.collaboration`: PASS, 8/8 units,
  `.cartulary/test-results/20260913T013601Z-p44366`.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser_stateful.verify_evidence_attach_preview_download_blocked_b2c37d6c7b`:
  FAIL, `.cartulary/test-results/20260913T013601Z-p44369`. One scenario fails
  at `evidence-integration.spec.ts:524`, helper line 778: revoked member's
  Evidence grid is absent. Its group and aggregate account for two failing
  graph units. Both hidden-404 checks already pass. The previous baseline
  control is retained at `.cartulary/tre-baseline-control/20260913T005032Z-p47519`.

## Execution tracker

| Workstream | Dependency | State | Exit |
| --- | --- | --- | --- |
| IR-01 authority and projection | none | DONE | Owners, schema, generated types and producer characterization agree. |
| IR-02 transport ownership | IR-01 | DONE | Scoped events and terminal-order tests pass. |
| IR-03 application integration | IR-02 | DONE | Directory navigation, independent session refresh and subscribers pass. |
| IR-04 browser and routing | IR-03 | DONE | Split real-service cases, accessibility, visuals and routing pass. |
| IR-05 final validation and handoff | IR-04 | DONE | Finalization, affected checks and finished-handoff review pass. |

Only one row may be IN_PROGRESS. Record paths, commands, results, risks,
disposition and next action before advancing. A blocked dependency stops its
successors.

## Gap ledger

| Gap / classification | Adopted authority | Remediation and areas | Rationale / benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- | --- |
| G1 confirmed owner wording ambiguity | Core 01 REQ-01-277; Core 03 REQ-03-299/100; Core 04 REQ-04-017 | Clarify Core 01/03 and Core 04 acceptance criteria. | One scope model prevents conflicting implementations. | Existing owner intent; ambiguous wording otherwise encourages global logout. | Four reasons, uncertainty, immediate concealment/navigation and retention are unambiguous. |
| G2 confirmed schema/event loss | Core 01 §3.3.6.2 canonical reason registry | Project the reason enum; normalize scoped frontend events centrally; Make-generate derivatives. | Exhaustive types replace repeated string interpretation. | Valid wire bytes unchanged; invalid reasons no longer decode as valid terminals. | Four valid reasons pass, invalid reasons never imply session loss, producers remain canonical. |
| G3 confirmed structural weakness; race consequences require characterization | Core 03 REQ-03-299 and REQ-03-100 | One terminal result per socket generation, identity fencing and uncertain recovery. | Deterministic connection ownership prevents timing-dependent navigation. | No new wire events/codes; stale callbacks and duplicate effects remain risks until tested. | Duplicate, close-only, invalid, foreign, replaced and disposed cases pass. |
| G4 confirmed scope integration defect | Core 03 REQ-03-299/100; existing account/incident owner lifetimes | Immediate incident invalidation/directory navigation, independent session refresh, all subscribers migrated. | Shared lifecycle handling prevents feature-specific patches. | Existing retention boundaries; no account-wide retirement on incident loss. | Other incident/account access survives; protected data stays hidden; stale refresh cannot reopen. |
| G5 confirmed test-design gap | Core 04 REQ-04-156/Table 2-C; TH-HARNESS-REQ-650/653 | Split local access and incident revocation; member is primary page; install readiness observers first. | Assertions and diagnostics describe the same actor and scope. | Harness artifacts remain diagnostic-only; no reporter/schema change. | Both cases pass independently and exact 404s, directory/session/isolation assertions pass. |
| G6 confirmed coverage gap | TH-HARNESS-REQ-650/653; adopted verification ownership | Reason/race/lifetime/service tests, authored routing, generated derivatives and this handoff. | Green owner selections meaningfully cover scope isolation. | Historical runs remain historical; skipped checks are explicit. | Every gap has final passing evidence and rollback disposition. |

## IR-01 authority and projection

Inspected the socket producer (`RevokeIncidentAccess`, `writeSessionRevoked`),
existing backend membership-isolation evidence, generated decoder inputs,
session planner/provider, workbook coordinator, application session/directory
callbacks, Network Flow subscribers and the Evidence browser scenario.
The backend already emits `incident_access_revoked` only for the subscribed
incident. No backend behavior prerequisite has been identified.

IR-01 exit: DONE. Core 01/03 wording and Core 04 AC-484 now distinguish
incident, session and uncertain authorization. `contracts/ws/index.schema.json`
projects the four canonical reasons. Make generated the Go contract artifact
and TypeScript types/validators; no generated file was hand-edited.
The protocol family test now rejects its formerly accepted noncanonical
`membership_removed` fixture and covers all reasons/additive members.

- `make generate`: PASS, `20260913T014457Z-p81296` under
  `.cartulary/test-results/`. An earlier owner-only generation also passed;
  the schema edit was corrected before this final generation.
- `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.boundary_support.index_decodes_exact_network_flow_contracts_witho_8fee657f09`:
  PASS, `20260913T014534Z-p84559`.
- `make test-slice OWNER=module.collaboration ROWS=module.collaboration.support_unit.semantic_hub_and_payload_ownership_d1e2f3a4b5`:
  PASS, `20260913T014534Z-p84589`.
- `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.support_integration.incident_socket_revocation_sources_00f908473e`:
  PASS, `20260913T014534Z-p84581`; canonical producers and other-incident
  isolation remain correct. No backend prerequisite or implementation change.

Residual risk: frontend still loses scope and terminal ordering. Next: IR-02
normalized events and connection-owned terminal outcomes with focused tests.

## IR-02 transport ownership

IR-02 exit: DONE. `incidentCollaborationSessionPlan.ts` emits a typed scoped
revocation with its canonical reason and incident. The connection provider owns
one monotonic terminal result per WebSocket, fences foreign/obsolete messages,
and treats malformed terminals and policy-close-only outcomes as uncertain.
Layout-lifetime cleanup prevents callbacks from surviving account replacement.
No unscoped internal revocation alias is retained.

- `make format`: PASS, `20260913T014932Z-p4537`.
- `make generate`: PASS, `20260913T015002Z-p8897` after authored test routing.
- `make test-slice OWNER=web.collaboration ROWS=web.collaboration.regression.incident_collaboration_session_plan_lr07,web.collaboration.regression.incidentcollaborationsession_suite_6b9c4a2c72`:
  PASS, `20260913T015017Z-p12023`. Tests cover all four reasons, additive data,
  invalid/missing reasons, duplicate terminals, message/close ordering, close-only
  recovery with a fresh hello, foreign incidents, replaced sockets, account
  replacement and disposal.

Residual risk: dependent consumers still require migration to the scoped event;
whole-frontend typing is intentionally checked after IR-03 integration. Next:
immediate directory navigation and independent account authorization refresh.

## IR-03 application integration

IR-03 exit: DONE. The workbook coordinator applies incident invalidation before
calling application navigation, cancels obsolete authority reads and rejects
terminal repeats. `App.tsx` fences incident callbacks by account lifetime,
returns immediately to the directory and starts the existing session owner's
independent refresh after navigation. Transient failures preserve the session;
only authoritative session failure enters authentication. Network Flow table,
graph, indicator-link and extension bindings share one scope translation;
import binding pauses incident-local work without account-wide retirement.
Existing target-change notifications and owner retention policies remain.

- `make format`: PASS, `20260913T015743Z-p15043`.
- `make generate`: PASS, `20260913T015758Z-p19309` after authored routing.
- `make test-slice OWNER=web.collaboration`: PASS, 8/8,
  `20260913T015817Z-p22607`; includes immediate scoped invalidation, duplicates,
  foreign incidents and rejection of late successful recovery.
- Application access-loss row `web.application.regression.app_landing_returns_to_the_landing_screen_when_w_2ee7e4975a`:
  PASS, `20260913T015817Z-p22616`; delayed successful, transient-failed and
  authoritative-session-failed refreshes, no reopening, explicit other incident.
- Network Flow interpreter row: PASS, `20260913T015817Z-p22633`.
- `make frontend-typecheck`: initial FAIL, `20260913T015817Z-p22793` exposed
  a missing test presence mode, omitted lifecycle union case and removed
  indicator target-change callback during migration. All corrected; PASS,
  `20260913T015902Z-p25438`.
- Application session lifecycle row: PASS, `20260913T015917Z-p26044`;
  late reads, account replacement and authoritative session failure.
- Workbook runtime coalescing/continuity rows: PASS,
  `20260913T015917Z-p26063`; same-account suspension and incident/account retirement.
- Network Flow import ownership/recovery and captured graph rows: PASS,
  `20260913T015917Z-p26068`; existing owner retention boundaries.

All run IDs above are under `.cartulary/test-results/`. No old unscoped internal
event consumers remain; the wire event and existing owner-local session failure
reason retain their names. Residual risk: integrated browser scope and focus
must still be proved. Next: IR-04 split primary-member browser evidence and
update semantic routing.

## IR-04 browser and routing

IR-04 exit: DONE. `evidence-integration.spec.ts` separates local preview/download
behavior from incident revocation. The new primary-member scenario observes socket
acceptance before revocation, checks both exact hidden-404 envelopes, immediate
protected-content removal, directory focus and live notice, preserved member
identity, no reconnection to the revoked incident and explicit keyboard entry to
another authorized incident. Administrator actions use the request fixture.
`tools/test_families/module.evidence.json` owns its dedicated semantic row;
Make generated batch/topology derivatives. Existing frontend paths retain their
authored owners; modified unit titles were added to their existing owner rows.

`internal/modules/collaboration/integration_test.go` additionally establishes
another user's same-incident socket and the member's other-incident socket before
removal, then proves both receive real committed records after targeted revocation.
No production backend code changed.

- `make generate`: PASS, `20260913T020143Z-p32762`.
- `make frontend-typecheck`: PASS, `20260913T020208Z-p36105`.
- Evidence stateful rows (new `incident_access_revocation` plus existing
  `verify_evidence_attach_preview_download_blocked_b2c37d6c7b`): PASS, 11/11,
  `20260913T020207Z-p35840`.
- Collaboration live-source integration: PASS within `20260913T020500Z-p74430`.
  That selection's genuine-session browser row initially FAILED before revocation:
  its shared helper's label matched both the open menu and its trigger. The helper
  in `support/collaboration/replay.ts` now selects the exact button role. The
  unchanged product scenario then PASSED, 11/11, `20260913T020659Z-p14711`.
- Coordinator scoped revocation with an actual queued draft: PASS,
  `20260913T020608Z-p8980`; invalidation retains the exact queue and never replays.
- Evidence accessibility and ordinary visual row: PASS, 13/13,
  `20260913T020728Z-p63688`.
- Directory/account-menu ordinary visual row: PASS, 11/11,
  `20260913T020728Z-p63692`. Reviewed matching directory and Evidence images and
  both reconciliation artifacts. Nine selected captures match existing goldens;
  no promotion, changed PNG or visual redesign. Unselected captures in narrow
  reconciliation are not deletion candidates.
- Auth/incident accessibility row initially FAILED, `20260913T020728Z-p63706`:
  live revocation returned to the directory before the test reloaded, so reload
  reset the intended access-loss focus. `workbook.a11y.spec.ts` now revisits the
  captured stale incident URL explicitly; PASS, 11/11, `20260913T020935Z-p76817`.

Negative control: an isolated detached checkout of `bc7664af` received only the
new browser file, Evidence routing and Make-generated topology. The new scenario
FAILED at the directory assertion, `evidence-integration.spec.ts:586`, after the
same two hidden-404 checks passed. Its primary-page diagnostic correctly shows
`Sign in to Cartulary` and `The current session ended`. Archived unchanged evidence:
`.cartulary/ir-baseline-control/20260913T020707Z-p29225/`. The temporary checkout
was removed after archival. Its initial generator attempts failed because the
ignored local `deploy/mvp/runtime-roots/tmp/.keep` was absent; creating that empty
runtime placeholder allowed ordinary Make generation. Those setup diagnostics
are also archived; they are not product failures or final-source passing evidence.

This supersedes the earlier Timeline handoff's deferred Evidence-access failure
without changing its historical results. No prerequisite remains. Next: IR-05
finalization, broader affected verification, completed ledger and final scope review.

## IR-05 final validation and handoff

IR-05 exit: DONE. All implementation, owner-selected and broadened checks below
passed. Final human scope review clarified that REQ-01-277 requires the client to
discard local terminal resume state; it does not introduce a backend token
revocation registry. No source/contract prerequisite, compatibility migration or
unresolved product failure remains.

Final verification (run IDs under `.cartulary/test-results/`):

| Command / selection | Result | Run |
| --- | --- | --- |
| `make agent-finalize` | PASS; generated outputs unchanged | `20260913T021108Z-p13138` |
| `make generate-drift` | PASS, 4/4 units | `20260913T021207Z-p16781` |
| `make frontend-unit` | PASS, 597/597 units | `20260913T021242Z-p20979` |
| `make frontend-typecheck` | PASS | `20260913T021242Z-p20973` |
| `make frontend-import-boundary-check` | PASS | `20260913T021242Z-p20997` |
| `make lint-biome` | PASS | `20260913T021242Z-p21034` |
| `make json-shape-check` | PASS | `20260913T021242Z-p20646` |
| `make generated-artifact-policy-check` | PASS | `20260913T021242Z-p20634` |
| Protocol owner family row from IR-01 | PASS | `20260913T021242Z-p20770` |
| Evidence: both stateful cases and accessibility row from IR-04 | PASS, 13/13 units | `20260913T021242Z-p20820` |
| Collaboration: revocation-source integration and genuine-session browser row | PASS, 12/12 units | `20260913T021242Z-p20840` |
| Auth/incident accessibility row from IR-04 | PASS, 11/11 units | `20260913T021242Z-p20860` |

The service selections use `make service-backed-test-slice OWNER=<owner>` with
exact semantic rows recorded in IR-04; protocol uses `make test-slice`. Unit
counts include harness prerequisites and aggregation, not just test cases.
Frontend coverage includes existing Evidence accepted-receipt/suspension and
account-retirement tests, as well as Network Flow owner recovery and the complete
application/workbook/collaboration regression inventory.

Final gap dispositions:

| Gap | Disposition and binary completion evidence |
| --- | --- |
| G1 | DONE: reviewed Core 01/03 wording and Core 04 AC-484 distinguish incident/session/uncertain outcomes and existing retention. |
| G2 | DONE: four canonical reasons/additive fields decode; invalid reasons reject; generated outputs and backend producers agree. Protocol, typechecking and drift pass. |
| G3 | DONE: canonical/invalid/close-only, duplicate, message-plus-close, foreign, replaced-socket, disposal and account-lifetime tests pass; obsolete recovery cannot navigate or replay. |
| G4 | DONE: full frontend owner tests and real-service incident/session cases pass. Queued drafts and accepted receipts follow existing suspension/retirement; directory navigation and account refresh remain independent. |
| G5 | DONE: local failures preserve the grid; incident loss clears it and preserves account/other-incident access. New scenario fails old frontend with the correct primary-member diagnostic. |
| G6 | DONE: authored routing, generated derivatives, broad frontend verification, backend/live browser isolation, accessibility, ordinary visuals and this handoff provide current evidence. |

`RESULTS_DIR` remained unset. Finalizer retained-run maintenance, performance
baseline updates and retained-run checks were skipped: no exact-source successful
full warm `check` evidence qualifies. No full `check`, CI, release, deployment or
unaffected backend sweep was requested or needed for this frontend/projection
change. Narrow ordinary visual validation passed; no golden update or mandatory
post-promotion reruns apply because no golden changed.

Final scope: only adopted owner clarification, WebSocket reason projection and
Make derivatives, scoped frontend revocation, focused test maintenance, authored
verification routing and this handoff. No digest, dependency, migration,
production backend behavior or analyst-data edits; no commit, push or deployment.
Execution branch and HEAD remain `main` at `bc7664afdf49b14870c327b20732209169a23586`.
Other existing worktrees were preserved. The negative-control scratch checkout
was removed after evidence archival.

After marking IR-05 DONE, `make lint-markdown` PASSED in
`20260913T021749Z-p87480`; `git diff --check`, the untracked handoff's whitespace
check, and the final changed-path/HEAD review also PASSED. The finished handoff
and final patch are subject to one final Markdown/whitespace recheck after this
result entry. Next action: user review of the uncommitted patch. No dependent
work is left open.

## Compatibility and rollback

Conforming `/ws/v1/` messages, routes and close codes stay unchanged. The reason
schema narrows to the adopted registry. Internal events migrate together without
aliases. Retention stays inside existing account/incident runtime lifetimes.
Rollback reverts frontend, projection, routing and documentation together. Never
restore revoked memberships or issue compensating mutations. Preserve committed
Evidence, attachments, receipts, custody history and revisions.
