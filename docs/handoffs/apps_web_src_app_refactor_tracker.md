# apps/web/src/app Refactor Tracker And Handoff Planner

## 1. Session Header

| Field | Value |
|---|---|
| Target directory | `apps/web/src/app` |
| Branch / commit | `main` / `96476122001a6b3b3e1133b5d7a4e73f4b8fbf66` |
| Dirty tree | Clean before creating this planning artifact |
| Date/time | `2026-06-27 22:27:40 EDT` |
| Agent/session | Codex implementation session |
| Mode | Planning and handoff only |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Source limits | Live repo is source of truth; no validation passed; only this planning artifact was created |

## 2. Scope And Non-Scope

| In scope | Non-scope / excluded | Preserved public behavior |
|---|---|---|
| App shell, root/provider wrapping, route-level browser entry surfaces, authentication gateway, landing/admin surfaces, incident admin entry, reference-pack admin entry, debug harness entrypoints, app-shell tests | Workbook internals, grid vendor integration, generated outputs, backend contracts, package manifests, source docs, behavior changes | Root route selection, auth/session readiness, deployment admin routing, incident landing/admin affordances, reference-pack admin controls, debug harness route, test selectors, app-to-workbook handoff |

## 3. Source And Authority Map

| Source | Class | Use |
|---|---|---|
| `docs/spec/00` through Core 04 | `behavior_owner` | Route/auth/admin/workbook startup/security boundaries |
| `docs/spec/05` | `behavior_owner` only for claim evidence | Do not use for ordinary app behavior |
| `docs/domain.md` | `behavior_owner` | Vocabulary and concept boundaries |
| `docs/testing-harness-nlspec.md` | `behavior_owner` | Harness command/artifact rules |
| Framework handoff doc | `prior_analysis` | Planning structure only |
| `apps/web/src/README.md` | `implementation_support` | App-shell ownership context; verified against live tree |
| Live `apps/web/src/app/**` | `current_repo_state` | File inventory/import/coupling truth |
| Nearby tests and ledgers | `test_evidence` | Existing characterization only if current files/tests exist |
| Makefiles, package manifests, import-boundary configs | `current_repo_state` | Available commands and boundary rules |
| Contradictions | `BLOCKED` | None found; coupling risks are implementation risks, not owner contradictions |

## 4. Current-State Inventory

| Path | Current responsibility | Target responsibility | Public/private/test/support | Imports in | Imports out | Coupling risk | Notes |
|---|---|---|---|---|---|---|---|
| `App.tsx` | App route/session composition | App-shell composition/router | public | `main`, tests | services, workbook lazy/type, app surfaces | high | Root, debug, admin, workbook handoff |
| `AppRoot.tsx` | Root/provider wrapper | Root/provider wrapper | public | `main` | UI provider, `App` | low | Theme/provider boundary |
| `AuthGateway.tsx` | Login/MFA/enterprise auth | Auth gateway | public | `Phase1Surface` | `phase1Client`, UI contracts | high | Auth readiness behavior |
| `DebugHarnessShell.tsx` | Phase harness aggregate | Debug entrypoint | support | `App` lazy | Phase harnesses | medium | Explicit debug route only |
| `IncidentAdminPanel.tsx` | Incident admin controls | Incident admin app-shell panel | public | `WorkbookShell`, tests | services, UI/view contracts, workbook startup model | high | App/workbook coupling |
| `LandingAdminSurface.tsx` | Landing, account, deployment admin, audit, import | Landing/admin surfaces | public | `App`, tests | services, `phase1Client`, UI contracts | high | Large mixed surface |
| `Phase1Harness.tsx` | Auth/account/admin debug harness | Debug support | support | `DebugHarnessShell` | services, `phase1Client` | medium | Raw selectors allowed locally |
| `Phase1Surface.tsx` | Auth/account/admin panels | Auth/account/admin surfaces | public | `App`, tests | `AuthGateway`, `phase1Client` | high | Phase-named runtime surface |
| `Phase2Harness.tsx` | Incident/ext debug harness | Debug support | support | `DebugHarnessShell` | services | medium | Duplicate fetch helper |
| `ReferencePackAdminPanel.tsx` | Reference pack admin | Reference-pack admin surface | public | `App`, tests | services, UI contracts, `SessionData` | high | Extension/job controls |
| `phase1Client.ts` | Auth/account/admin route client | App-shell client facade | public support | app files, tests, test support | protocol facade, browser API | high | Phase-named runtime helper |
| `App.landing.test.tsx` | Landing/admin route characterization | App-shell tests | test | vitest | app/test helpers | medium | Strong route/admin coverage |
| `App.phase1.support.test.tsx` | Selector/support coverage | App-shell tests | test | vitest | app/test helpers | medium | Selector contracts |
| `App.phase1.test.tsx` | Auth/admin/account coverage | App-shell tests | test | vitest | app/test helpers | medium | Broad auth/admin coverage |
| `App.test.tsx` | Workbook invalidation behavior | adjacent test | test | vitest | `WorkbookShell` | high | Workbook internals; non-scope unless handoff touched |
| `IncidentAdminPanel.test.tsx` | Incident admin coverage | app-shell tests | test | vitest | panel/test helpers | medium | Role/prefs/membership coverage |
| `ReferencePackAdminPanel.test.tsx` | Ref-pack coverage | app-shell tests | test | vitest | panel/test helpers | medium | Job/search/admin controls |
| `fontBundle.test.ts` | Font bundle boundary | readiness test | test | vitest | fs/path/url | low | Design readiness evidence |
| `fontRoles.test.tsx` | Font role behavior | adjacent readiness test | test | vitest | `App`, `WorkbookShell` | medium | App+workbook |
| `otelBoundary.test.ts` | Browser OTel boundary | readiness test | test | vitest | globals | low | Browser runtime boundary |
| `phase1Client.routeBoundary.test.ts` | Client route contract | app-shell tests | test | vitest | `phase1Client`, browser API | medium | Route/body/header coverage |

## 5. App-Shell Responsibility Map

| Seam | Files |
|---|---|
| Top-level composition and routing | `App.tsx`, `AppRoot.tsx` |
| Root/provider wrapping | `AppRoot.tsx` |
| Authentication gateway and account readiness | `AuthGateway.tsx`, `Phase1Surface.tsx`, `phase1Client.ts` |
| Landing/admin surfaces | `LandingAdminSurface.tsx`, `Phase1Surface.tsx` |
| Incident administration | `IncidentAdminPanel.tsx`; consumed by `workbook/WorkbookShell.tsx` |
| Reference-pack administration | `ReferencePackAdminPanel.tsx` |
| Debug/phase harness entrypoints | `DebugHarnessShell.tsx`, `Phase1Harness.tsx`, `Phase2Harness.tsx` |
| App-level route/client helpers | `phase1Client.ts`, route helpers inside `App.tsx` |
| App-shell tests and support tests | All tests under `apps/web/src/app`; workbook-focused tests are adjacent evidence, not scope expansion |

## 6. Contract And Behavior Preservation Matrix

| Surface | Files | Observable behavior | Existing tests | Missing characterization | Risk | Required next action |
|---|---|---|---|---|---|---|
| Route selection | `App.tsx` | `/`, `incident_id`, debug, deployment admin, sole-incident auto-open | `App.landing`, `App.phase1` | TODO: popstate/back-forward edge coverage | high | Freeze before splitting |
| Auth/session readiness | `AuthGateway`, `Phase1Surface`, `phase1Client` | Login, revoked session, MFA, enterprise auth, profile/preferences | `App.phase1`, route-boundary tests | TODO: visual e2e mapping | high | Characterize route envelopes |
| Deployment admin entry | `App`, `LandingAdminSurface`, `Phase1Surface` | Admin-only entry; non-admin route returns `/`; extension-gated panels | `App.landing`, support tests | TODO: admin audit panel direct coverage | high | Keep menu/denial semantics |
| Incident admin panel | `IncidentAdminPanel` | Role gates, TLP tokens, membership, prefs, close/reopen | `IncidentAdminPanel.test` | TODO: workbook embedded drawer coverage mapping | high | Isolate app/workbook coupling last |
| Reference-pack admin | `ReferencePackAdminPanel` | Admin controls, import/action/refresh/job cancel/search | panel tests, phase 11 ledger hits | TODO: extension claim integration coverage | high | Preserve job polling semantics |
| Debug harnesses | `DebugHarnessShell`, phase harnesses | Only explicit `?debug=harness`; support selectors | `App.landing` debug mock | TODO: direct harness smoke tests | medium | Keep support-only |
| Font/Otel boundaries | font/otel tests | Font manifest/roles; no browser telemetry exporters | font/otel tests | none known | low | Run when touching root/runtime |
| Test selectors | app code, selector policy | Shared builders plus app-local allowlist | selector policy tests | TODO: promote only with contract owner | medium | Do not churn selectors casually |
| App-to-workbook handoff | `App`, `WorkbookShell`, `IncidentAdminPanel` | Workbook open, account menu, density, access lost, incident controls | app mocks, workbook tests | TODO: facade-level characterization | high | Plan dedicated slice |

## 7. Boundary And Coupling Scan

| Finding | Class | Evidence | Next action |
|---|---|---|---|
| Runtime phase names in `Phase1Surface.tsx`, `phase1Client.ts` | `should_fix` | Framework prefers module-shaped production code | Add compatibility facade before renames |
| Debug phase harness names | `intentional` | README allows phase-named tests/harness evidence | Keep support-only |
| `IncidentAdminPanel` imports workbook startup model | `should_fix` | App code reaches into workbook model | Move behind neutral facade after tests |
| `WorkbookShell` imports `IncidentAdminPanel` | `should_fix` | Bidirectional app/workbook dependency | Design handoff/slot slice |
| App imports workbook shell lazily | `intentional` | App-to-workbook handoff is in scope | Preserve |
| Raw `data-testid` literals in panels/harnesses | `defer` | Selector policy allowlist covers app-local IDs | Promote only in selector slice |
| Runtime importing test helpers | `intentional` none found | Import-boundary scan evidence | Keep boundary check |
| Direct generated imports | `intentional` none found | Uses package facades | Do not edit generated roots |
| Direct grid vendor import | `intentional` none found | No `react-data-grid` under target | Keep excluded |
| Path aliases/app-local barrels | `defer` | No TS path aliases; `Phase1Surface` re-export exists | Avoid new barrels |

## 8. Workflow Dependency Map

| Workflow | Previous | Subsequent | Outputs | Binary exit criteria |
|---|---|---|---|---|
| WF-00 session/source bootstrap | none | WF-01 | Header, sources, command inventory | Branch, commit, dirty state, owners recorded |
| WF-01 live target scan | WF-00 | WF-02 | File list/import scan | Every live target file seen or marked unseen |
| WF-02 app responsibility inventory | WF-01 | WF-03 | Inventory and seam map | Each file has current/target role |
| WF-03 public contract freeze | WF-02 | WF-04/WF-05 | Preservation matrix | Observable behaviors and tests listed |
| WF-04 boundary/coupling scan | WF-02 | WF-05/WF-06 | Risk table | Each coupling classified |
| WF-05 characterization test plan | WF-03/WF-04 | WF-06 | Missing-test plan | Required characterization named |
| WF-06 refactor slice candidate design | WF-05 | WF-07 | Slice backlog | Slices are behavior-preserving or blocked |
| WF-07 validation plan | WF-06 | WF-08 | Command ladder | Commands discovered before assertion |
| WF-08 handoff and next-slice bootstrap | WF-07 | next agent | Session record and prompt | Restart path is explicit |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence/artifact | Exit condition |
|---|---|---|---|---|---|---|
| T-001 | Scope/source bootstrap | Scope | DONE | none | Header/source map | Target and limits explicit |
| T-002 | Live file inventory | Evidence | DONE | T-001 | `rg --files`, import scan | 21 files inventoried |
| T-003 | Responsibility map | App shell | DONE | T-002 | Seam table | No workbook internals absorbed |
| T-004 | Contract freeze | Tests | DONE | T-002 | Matrix | Behaviors/tests/missing char listed |
| T-005 | Boundary scan | Boundary | DONE | T-002 | Risk table | Risks classified |
| T-006 | Slice backlog | Planning | DONE | T-004/T-005 | Backlog | Only preserving slices |
| T-007 | Validation ladder | Validation | DONE | T-006 | Make/package scan | Commands discovered |
| T-008 | Create output artifact | Handoff | DONE | T-001-T-007 | `temp/apps_web_src_app_refactor_tracker.md` | File exists with this content |
| T-009 | Run chosen validation after future edits | Validation | TODO | implementation slice | Make run artifacts | Results reported exactly |

## 10. Refactor Slice Backlog

| Slice ID | Dependency | Proposed change | Files likely involved | Behavior expected unchanged | Characterization needed | Validation | Risk | Handoff note |
|---|---|---|---|---|---|---|---|---|
| S-001 | WF-05 | Freeze app route/auth/admin characterization | app tests only | All current UI/API behavior | Route/auth/admin gaps | `make frontend-unit` | low | Do before moves |
| S-002 | S-001 | Add module-shaped client facade, keep `phase1Client` compatibility | `phase1Client.ts`, imports | Route shapes/bodies/headers | route-boundary tests | frontend unit/typecheck/import boundary | medium | No endpoint changes |
| S-003 | S-001 | Split `Phase1Surface` into auth/account/admin modules with stable exports | `Phase1Surface`, `AuthGateway` | Auth/account/admin UI | selector support tests | frontend unit | high | Keep public export first |
| S-004 | S-001 | Extract route-state helpers from `App.tsx` | `App.tsx`, new app helper | URL/history behavior | route edge tests | frontend unit/typecheck | high | No route semantics change |
| S-005 | S-001 | Split landing/admin/account panels from `LandingAdminSurface` | landing surface files | Landing/admin/account behavior | admin audit TODO | frontend unit | high | Keep selectors stable |
| S-006 | S-001 | Introduce incident admin/workbook handoff facade | `IncidentAdminPanel`, `WorkbookShell` | Drawer/admin controls | facade tests | frontend unit + workbook tests | high | Do after smaller slices |
| S-007 | S-006 | Remove app import of workbook startup internals | incident panel/shared facade | Sheet ref display/prefs | prefs formatting tests | frontend unit/typecheck | medium | Neutral helper only |
| S-008 | S-001 | Isolate debug harness support code | debug/phase harnesses | `?debug=harness` behavior | direct harness smoke TODO | frontend unit | medium | Support-only rename |
| S-009 | S-001 | Ref-pack panel helper extraction | `ReferencePackAdminPanel` | Job/search/admin controls | extension integration TODO | frontend unit | medium | Preserve polling/cancel |
| S-010 | S-001 | Promote selected app-local selectors to contract package | panel + UI contracts | Selector strings | selector policy | import boundary/unit | medium | Only with owner approval |

## 11. Validation Plan

| Order | Command | Evidence type | Use |
|---|---|---|---|
| 1 | `make task-guide ROLE=feature-dev PHASE=phase1` / `phase2` / `phase11` | implementation support | Choose narrow phase by touched surface |
| 2 | `make explain-target TARGET=frontend-unit DETAIL=rows` | implementation support | Confirm unit coverage rows before rerun |
| 3 | `make frontend-unit` | implementation support/test evidence | Cheapest broad app-shell unit gate |
| 4 | `make frontend-typecheck` | implementation support | Type/API refactor gate |
| 5 | `make frontend-import-boundary-check` | implementation support | Runtime/test/generated/vendor boundary gate |
| 6 | `make lint-biome` | implementation support | Frontend lint/format conformance |
| 7 | `make generated-artifact-policy-check` and `make json-shape-check` | harness/config | If generated policy/contracts are touched |
| 8 | `make browser-e2e-webserver-backed` or targeted browser target from task guide | design/readiness | Route/UI workflow confidence after shell moves |
| 9 | `make browser-e2e-a11y` / `make browser-e2e-visual` | design readiness | Only for visual/a11y-impacting UI changes |
| 10 | `make test-fast`, then `make check` | broad implementation support | End-of-slice or release-risk expansion |
| 11 | `make agent-finalize` | handoff hygiene | Before broad final verification; pass `RESULTS_DIR` only if retained run exists |

No command above is claimed passing in this session.

## 12. Handoff Sections By Workstream

| Workstream | Date | Files/Surface | Current state | Evidence/command | Next action |
|---|---|---|---|---|---|
| Scope and evidence | 2026-06-27 | target/framework/README/domain | Scoped to app shell | `rg`, `wc`, `make help-all` | Begin WF-05 characterization |
| App-shell composition | 2026-06-27 | `App`, `AppRoot` | Large route/session composition | import scan/tests | Freeze route behavior first |
| Auth and route readiness | 2026-06-27 | `AuthGateway`, `Phase1Surface`, `phase1Client` | Runtime phase naming remains | tests/import scan | Add facade before rename |
| Landing/admin surfaces | 2026-06-27 | `LandingAdminSurface`, ref-pack/admin panels | Mixed but covered | app tests | Split after characterization |
| Debug harnesses | 2026-06-27 | `DebugHarnessShell`, phase harnesses | Support-only debug route | file scan | Keep out of product architecture |
| Tests and harness | 2026-06-27 | app tests | Existing unit coverage, gaps noted | test inventory | Run Make targets only after edits |
| Boundary risks | 2026-06-27 | app/workbook coupling | Bidirectional dependency exists | import scan | Plan dedicated high-risk slice |
| Generated artifacts | 2026-06-27 | generated roots | No target generated files | policy scan | Do not hand-edit generated roots |
| Open blockers | 2026-06-27 | none | No owner contradictions | source map | TODO: validation not run |

## 13. Session Handoff Record Template

| Field | Value |
|---|---|
| Date/time | `2026-06-27 22:27:40 EDT` |
| Branch/commit | `main` / `96476122001a6b3b3e1133b5d7a4e73f4b8fbf66` |
| Target seam | `apps/web/src/app` app shell |
| Current workflow | WF-08 handoff planning |
| Completed workflows | WF-00 through WF-08 by inspection and artifact creation only |
| Changed planning files | `temp/apps_web_src_app_refactor_tracker.md` |
| Commands run | `git status`, `git rev-parse`, `date`, `ls`, `rg`; prior planning evidence included `wc`, `sed`, `make help`, `make help-all` |
| Passing validation | None claimed |
| Failing validation | None |
| Decisions made | App-shell scope excludes workbook internals; phase runtime names are refactor candidates; generated roots excluded |
| Open questions | TODO: exact browser e2e rows for app shell; TODO: admin audit direct characterization |
| Blockers | None |
| Next recommended workflow | WF-05 characterization test plan |
| Safe restart command | `make task-guide ROLE=feature-dev PHASE=phase1` |

## 14. Binary Acceptance Criteria

- Target directory and non-scope are explicit.
- Every live file under `apps/web/src/app` is inventoried.
- App-shell behavior to preserve is listed.
- Existing tests and missing characterization are identified.
- Workflow dependencies are explicit.
- Slice candidates are behavior-preserving or explicitly TODO.
- Validation commands were discovered before being named.
- Generated files are excluded from hand edits.
- No phase identity is introduced as runtime architecture.
- Handoff is sufficient for another agent to resume without rediscovery.

## Next Recommended Local-Agent Prompt

Execute WF-05 for `apps/web/src/app`: inspect current app-shell route/auth/admin test coverage, add no production changes, identify the narrowest missing characterization tests for the first behavior-preserving refactor slice, and report the Make validation targets to run before any implementation.
