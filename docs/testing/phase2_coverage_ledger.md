# Phase 2 Coverage Ledger

This ledger is the human-readable companion to `tools/phase2_test_map.json`.

- Scope: incident create or list or get or patch, bootstrap membership and workbook preferences, incident membership administration, incident-scoped authorization, extension discovery and reserved-family dispatch, plus the ordinary landing and incident-shell Phase 2 UI.
- Normative owners: Core 01 `§3.3.1`, `§3.3.3`, `§3.3.5.1`, `§3.3.5.2`, `§3.3.5.3`; Core 02 `§4.5`; Core 04 `§2`, `§3`.
- Authority: `tools/phase2_test_map.json` is the enforced Phase 2 traceability source. This ledger summarizes the same surface in prose and does not control the mechanical row inventory.
- Authoritative execution:
  - `backend-unit` selects authoritative Go-backed `U-2-01..U-2-10` rows through the Phase 2 manifest.
  - `frontend-unit` selects authoritative Vitest-backed `U-2-11..U-2-13` rows through `run-vitest-manifest-phase phase2 authoritative frontend_unit`.
  - `backend-integration` selects authoritative `I-2-*` rows through the Phase 2 manifest.
  - `browser-e2e-webserver-backed` keeps authoritative real-browser `E-2-01..E-2-03`.
- Support-only evidence:
  - `internal/modules/incidents/phase2_http_conformance_test.go`, `phase2_pagination_integration_test.go`, and `phase2_extra_integration_test.go` remain support coverage only and must not claim authoritative Phase 2 IDs.
  - `apps/web/e2e/phase2.support.spec.ts` and browser-authenticated request probes remain supplemental browser evidence only.
  - `cmd/server/main_phase2_smoke_test.go` remains process-smoke evidence only.

## Unit

| Row | Evidence | Task target | Real services | Major assertions | Remaining gap |
| --- | --- | --- | --- | --- | --- |
| `U-2-01` | `internal/modules/incidents/phase2_request_test.go::TestPhase2_U_2_01_IncidentCreateAcceptsDeclaredMembersAndNormalizesIncidentKey` | `backend-unit via manifest` | None | Incident create decode rejects undeclared members, rejects `initial_memberships`, and normalizes `incident_key` with trim plus NFC before uniqueness. | Durable create persistence remains integration evidence. |
| `U-2-02` | `internal/modules/incidents/phase2_unit_test.go::TestPhase2_U_2_02_IncidentCreateBootstrapPlanIncludesCreatorAdminAndWorkbookPreferences` | `backend-unit via manifest` | None | Create planning bootstraps the creator as incident `admin` and schedules both workbook-preference objects. | Live HTTP shape remains support or integration evidence. |
| `U-2-03` | `internal/modules/incidents/phase2_unit_test.go::TestPhase2_U_2_03_IncidentCreateLocationIsRootedAtIncidentMember` | `backend-unit via manifest` | None | Successful create exposes the stable `Location` rooted at `/api/v1/incidents/{incident_id}`. | Runtime envelope proof remains support evidence. |
| `U-2-04` | `internal/modules/incidents/phase2_unit_test.go::TestPhase2_U_2_04_IncidentCreateIdempotencyScopesByActorAndNormalizedRequest` | `backend-unit via manifest` | None | Incident create idempotency scopes by actor plus normalized request and rejects divergent replay. | Durable replay storage remains integration evidence. |
| `U-2-05` | `internal/modules/incidents/phase2_request_test.go::TestPhase2_U_2_05_IncidentPatchAllowsPromotedFieldsAndKeepsNoOpVersionStable` | `backend-unit via manifest` | None | Incident patch decode allows only promoted fields, requires `base_incident_version`, and keeps structural no-ops version-stable. | Persisted patch consequences remain integration evidence. |
| `U-2-06` | `internal/modules/incidents/phase2_request_test.go::TestPhase2_U_2_06_MembershipCreateDecodeRejectsInvalidSelectorsAndInvitationFields` | `backend-unit via manifest` | None | Membership create requires exactly one selector, keeps a closed role vocabulary, and rejects invitation-only fields. | Runtime user lookup consequences remain integration or support evidence. |
| `U-2-07` | `internal/modules/incidents/phase2_request_test.go::TestPhase2_U_2_07_MembershipPatchAndDeleteDecodeEnforceBaseVersionAndLastAdminGuard` | `backend-unit via manifest` | None | Membership patch and delete require `base_membership_version` and keep the last-admin guard stable. | Durable mutation rows remain integration evidence. |
| `U-2-08` | `internal/modules/incidents/phase2_unit_test.go::TestPhase2_U_2_08_DeploymentAdminWithoutMembershipDoesNotGainIncidentAccess` | `backend-unit via manifest` | None | Deployment-admin alone does not imply incident-scoped read, write, membership, preference, or socket access. | Exhaustive route-matrix coverage is support or integration evidence. |
| `U-2-09` | `internal/modules/incidents/phase2_request_test.go::TestPhase2_U_2_09_ExtensionDiscoveryReturnsExactSingletonProfileShape` | `backend-unit via manifest` | None | Extension discovery keeps the singleton route shape, exact profile fields, canonical ordering, and no pagination members. | Runtime auth and leak checks remain integration evidence. |
| `U-2-10` | `internal/modules/incidents/phase2_request_test.go::TestPhase2_U_2_10_ReservedExtensionDispatchHonorsBaseRoutesClaimedFamiliesAndOutsideFallback` | `backend-unit via manifest` | None | Reserved-family dispatch preserves base-route precedence, claimed-family routing, reserved-but-unclaimed canonical 404s, and ordinary outside-family 404s. | Real authenticated dispatch remains integration evidence. |
| `U-2-11` | `apps/web/src/App.landing.test.tsx::Phase 2 U-2-11 ordinary landing shell creates an incident, refreshes session-visible membership, routes to the workbook by incident_id, and falls back when a stale incident selection is no longer visible` | `frontend-unit via manifest` | None | The ordinary landing shell creates an incident, refreshes session-visible membership on workbook entry, stores the selected `incident_id` in the route, and returns cleanly to landing when a stale selection is no longer visible. | Real browser confirmation stays `E-2-01` and `E-2-02`. |
| `U-2-12` | `apps/web/src/IncidentAdminPanel.test.tsx::Phase 2 U-2-12 ordinary incident shell gates promoted-field controls by incident role, hides membership-admin controls from non-admin members, and returns to landing when incident access is lost` | `frontend-unit via manifest` | None | The ordinary incident shell exposes promoted-field controls only for reviewer or admin roles, hides membership-administration controls from non-admin members, and triggers the landing fallback when incident access disappears. | Real browser visibility remains `E-2-02` and `E-2-03`. |
| `U-2-13` | `apps/web/src/IncidentAdminPanel.test.tsx::Phase 2 U-2-13 ordinary incident shell issues membership create, patch, and delete requests with versioned payloads and refreshes session role after each mutation` | `frontend-unit via manifest` | None | The ordinary incident shell emits the expected membership mutation payloads, includes versioned patch or delete bodies, and refreshes the session-visible incident role after each mutation. | Real browser mutation flow remains `E-2-03`. |

## Integration

| Row | Evidence | Task target | Real services | Major assertions | Remaining gap |
| --- | --- | --- | --- | --- | --- |
| `I-2-01` | `internal/modules/incidents/phase2_integration_test.go::TestPhase2_I_2_01_IncidentCreatePersistsBootstrapStateAndRollsBackAtomically` | `backend-integration via manifest` | PostgreSQL + real runtime | Incident create persists the incident row, bootstrap membership, workbook preferences, and owner-aware mutation artifacts atomically, while hook-forced rollback leaves no surviving Phase 2 mutation artifacts. | None for the declared create contract. |
| `I-2-02` | `internal/modules/incidents/phase2_integration_test.go::TestPhase2_I_2_02_IncidentCreateReplayAndDuplicateKeyConflictUseNormalizedState` | `backend-integration via manifest` | PostgreSQL + real runtime | Real create replay returns the original committed resource, rejects divergent replay, and keeps replay counters stable. | None for the declared replay contract. |
| `I-2-03` | `internal/modules/incidents/phase2_integration_test.go::TestPhase2_I_2_03_MembershipChangesReDeriveAuthorizationImmediately` | `backend-integration via manifest` | PostgreSQL + real runtime + real websocket boundary | One same-session role progression `no membership -> viewer -> reviewer -> admin -> removed` re-derives access immediately for every inventory-owned incident-scoped control route, including the Timeline websocket boundary. | None for the declared authorization re-derivation contract. |
| `I-2-04` | `internal/modules/incidents/phase2_integration_test.go::TestPhase2_I_2_04_IncidentPatchPersistsOnlyPromotedFieldsAndAdvancesOnMaterialChange` | `backend-integration via manifest` | PostgreSQL + real runtime | Real incident patch persists only promoted fields, rejects stale versions, and advances `incident_version` only on material change. | None for the declared patch contract. |
| `I-2-05` | `internal/modules/incidents/phase2_integration_test.go::TestPhase2_I_2_05_ExtensionDiscoveryReturnsExactZeroMembershipShapeWithoutLeaks` | `backend-integration via manifest` | PostgreSQL + real runtime | Extension discovery succeeds for an authenticated zero-membership session and leaks no provider secrets or live claim payloads. | None for the declared discovery contract. |
| `I-2-06` | `internal/modules/incidents/phase2_integration_test.go::TestPhase2_I_2_06_UnclaimedReservedFamiliesReturnCanonical404AndOutsidePathsDoNot` | `backend-integration via manifest` | PostgreSQL + real runtime | Reserved but unclaimed family roots and descendants return canonical `extension_profile_not_claimed`, while outside-family paths preserve ordinary 404 handling. | None for the declared dispatch contract. |

## Browser E2E

| Row | Evidence | Task target | Real services | Major assertions | Remaining gap |
| --- | --- | --- | --- | --- | --- |
| `E-2-01` | `apps/web/e2e/phase2.spec.ts::E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface` | `browser-e2e-webserver-backed via manifest` | Real browser + Go server + PostgreSQL + Vite | Real browser incident create bootstraps the creator as incident admin and lands on the ordinary workbook shell. | None for the declared browser create-flow contract. |
| `E-2-02` | `apps/web/e2e/phase2.spec.ts::E-2-02 shows incident discovery, direct retrieval, and promoted-field-only patching on the ordinary incident shell` | `browser-e2e-webserver-backed via manifest` | Real browser + Go server + PostgreSQL + Vite | The ordinary shell exposes incident discovery, direct retrieval, and promoted-field-only patching on the real browser surface. | None for the declared browser patch-flow contract. |
| `E-2-03` | `apps/web/e2e/phase2.spec.ts::E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell` | `browser-e2e-webserver-backed via manifest` | Real browser + Go server + PostgreSQL + Vite | Incident admins manage memberships on the ordinary shell and non-admin incident members do not receive those controls. | None for the declared browser membership contract. |

## Shared Harness Coverage

| Harness | Phase 2 evidence |
| --- | --- |
| Public-route inventory ownership | `internal/testutil/phase2test.PublicRouteInventory()` now owns the Phase 2 HTTP success-envelope inventory, including success status, CSRF expectation, envelope coverage, and mutation-owner metadata for Phase 2 owned routes. |
| Role-aware control boundary inventory | `internal/testutil/phase2test.ControlBoundaryInventory()` now owns the incident-scoped control boundary matrix across incident reads or writes, membership administration, workbook preferences, Timeline or entity queries, Timeline row create, record patch, lifecycle actions, and the websocket boundary. |
| Owner-aware mutation helpers | `internal/testutil/phase2test.LookupOwnerMutations()`, `RequireOwnerMutationEvent()`, `CountMutationArtifacts()`, and `RequireNoMutationArtifacts()` keep Phase 2 tests tied to incident resource or membership mutation ownership instead of the current storage table name. |
| Reusable incident route fixture helpers | `internal/modules/incidents/phase2_inventory_helpers_test.go` centralizes the seeded incident, membership, and record fixture setup plus route execution helpers shared by the Phase 2 support and integration matrices. |
| Frontend ordinary-shell component coverage | `apps/web/src/App.landing.test.tsx` and `apps/web/src/IncidentAdminPanel.test.tsx` keep the authoritative frontend-unit evidence on the ordinary landing or incident shell rather than the debug-only `Phase2Harness`. |

## Support-Only Evidence

- `internal/modules/incidents/phase2_http_conformance_test.go::TestSupportPhase2_PublicRouteInventoryEnvelopes` loops every `PublicRouteInventory()` entry and proves the success-envelope contract stays attached to the shared inventory rather than scattered route lists.
- `internal/modules/incidents/phase2_http_conformance_test.go::TestSupportPhase2_ControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied` loops every `ControlBoundaryInventory()` entry and proves deployment-admin-without-membership denial across the entire incident-scoped surface.
- `internal/modules/incidents/phase2_extra_integration_test.go` keeps replay, before-or-after mutation payload, NFC key normalization, and other regression coverage that strengthens confidence but does not replace the authoritative `I-2-*` rows.
- `apps/web/e2e/phase2.support.spec.ts` and browser-authenticated request probes remain supplemental browser evidence for request-shape regressions and reserved-route behavior. They do not replace the authoritative `E-2-*` rows.
