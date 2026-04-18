# Phase 1 Coverage Ledger

This ledger is the authoritative repo-local map for Phase 1 test evidence.

- Scope: authentication, sessions, and deployment-local user administration only.
- Normative owners: Core 01 `§3.3.2`, `§3.3.2.2`, `§3.3.5.1`, `§3.3.10`; Core 04 `§1`, `§3`.
- Support-only evidence: `cmd/server/main_phase1_process_test.go` process smoke tests are intentionally excluded from the authoritative `E-*` rows below.

## Unit

| Row | Owner sections | Evidence |
| --- | --- | --- |
| `U-1-01` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/modules/auth/phase1_request_test.go::TestPhase1_LoginRequestShape_U_1_01` |
| `U-1-02` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/platform/authn/phase1_primitives_test.go::TestPhase1_LoginNormalizationAndPasswordExactness_U_1_02` |
| `U-1-03` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/platform/authn/phase1_primitives_test.go::TestPhase1_SessionTiming_U_1_03` |
| `U-1-04` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/modules/auth/phase1_request_test.go::TestPhase1_SessionInspectionContracts_U_1_04` |
| `U-1-05` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/platform/authn/phase1_primitives_test.go::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05` |
| `U-1-06` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/platform/authn/phase1_primitives_test.go::TestPhase1_RevocationScopes_U_1_06` |
| `U-1-07` | Core 01 `§3.3.5.1`; Core 04 `§3` | `internal/modules/auth/phase1_users_test.go::TestPhase1_UserCreateDefaultsAndSafeShape_U_1_07` |
| `U-1-08` | Core 01 `§3.3.5.1`; Core 04 `§3` | `internal/modules/auth/phase1_users_test.go::TestPhase1_UserPatchAndLastAdminGuard_U_1_08` |
| `U-1-09` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/modules/auth/phase1_request_test.go::TestPhase1_CSRFFailClosed_U_1_09` |
| `U-1-10` | Core 01 `§3.3.2.2`; Core 04 `§1` | `internal/modules/auth/phase1_request_test.go::TestPhase1_CredentialStateInspection_U_1_10` |
| `U-1-11` | Core 01 `§3.3.2.2`; Core 04 `§1` | `internal/modules/auth/phase1_request_test.go::TestPhase1_PasswordChangeRequest_U_1_11` |
| `U-1-12` | Core 01 `§3.3.2.2`; Core 04 `§1` | `internal/modules/auth/phase1_request_test.go::TestPhase1_TOTPBootstrapAndSetupRules_U_1_12` |
| `U-1-13` | Core 01 `§3.3.5.1`; Core 04 `§3` | `internal/modules/auth/phase1_users_test.go::TestPhase1_AdminCredentialActionsRequireDeploymentAdmin_U_1_13` |

## Integration

| Row | Owner sections | Evidence |
| --- | --- | --- |
| `I-1-01` | Core 01 `§3.3.2`; Core 04 `§1` | `internal/modules/auth/phase1_integration_test.go::TestPhase1_LoginSessionLifecycle_I_1_01` |
| `I-1-02` | Core 01 `§3.3.10`; Core 04 `§1` | `internal/modules/auth/phase1_integration_test.go::TestPhase1_SessionRevocationClosesAttachedSocket_I_1_02` |
| `I-1-03` | Core 01 `§3.3.5.1`; Core 04 `§3` | `internal/modules/auth/phase1_integration_test.go::TestPhase1_UserAdminLifecycle_I_1_03`, `internal/modules/auth/phase1_integration_test.go::TestPhase1_UserAdminAudit_I_1_03` |
| `I-1-04` | Core 01 `§3.3.2.2`; Core 04 `§1` | `internal/modules/auth/phase1_integration_test.go::TestPhase1_CredentialStateAndBootstrapFlows_I_1_04`, `internal/modules/auth/phase1_integration_test.go::TestPhase1_CredentialStateTransitions_I_1_04`, `internal/modules/auth/phase1_integration_test.go::TestPhase1_BootstrapEnrollmentConsumption_I_1_04` |
| `I-1-05` | Core 01 `§3.3.5.1`; Core 01 `§3.3.10`; Core 04 `§3` | `internal/modules/auth/phase1_integration_test.go::TestPhase1_AdminCredentialActions_I_1_05`, `internal/modules/auth/phase1_integration_test.go::TestPhase1_AdminCredentialAuditAndScope_I_1_05` |
| `I-1-06` | Core 01 `§3.3.2.2`; Core 01 `§3.3.10`; Core 04 `§1` | `internal/modules/auth/phase1_integration_test.go::TestPhase1_BootstrapTokenRouteBoundaries_I_1_06` |

## Browser E2E

| Row | Owner sections | Evidence |
| --- | --- | --- |
| `E-1-01` | Core 01 `§3.3.2`; Core 04 `§1` | `apps/web/e2e/phase1.spec.ts::E-1-01 logs in as a local user and inspects the singleton session resource` |
| `E-1-02` | Core 01 `§3.3.2`; Core 01 `§3.3.2.2`; Core 04 `§1` | `apps/web/e2e/phase1.spec.ts::E-1-02 requires MFA when the account has an active factor, rejects wrong codes, and accepts a valid TOTP code` |
| `E-1-03` | Core 01 `§3.3.2`; Core 04 `§1` | `apps/web/e2e/phase1.spec.ts::E-1-03 rejects invalid credentials without issuing a session cookie` |
| `E-1-04` | Core 01 `§3.3.2`; Core 04 `§1` | `apps/web/e2e/phase1.spec.ts::E-1-04 forces the idle expiry boundary and requires a fresh login afterwards` |
| `E-1-05` | Core 01 `§3.3.5.1`; Core 04 `§3` | `apps/web/e2e/phase1.spec.ts::E-1-05 lets deployment admins create and patch users, rejects stale versions, and preserves the last-admin guard` |
| `E-1-06` | Core 01 `§3.3.2.2`; Core 04 `§1` | `apps/web/e2e/phase1.spec.ts::E-1-06 follows the bootstrap-token enrollment sequence and proves first-time completion alone issues no session` |
| `E-1-07` | Core 01 `§3.3.2.2`; Core 04 `§1` | `apps/web/e2e/phase1.spec.ts::E-1-07 requires the current password and current TOTP code, revokes the session immediately, and requires re-login with the new password` |
| `E-1-08` | Core 01 `§3.3.5.1`; Core 04 `§3` | `apps/web/e2e/phase1.spec.ts::E-1-08 keeps credential actions deployment-admin only and denies the same routes to a non-deployment-admin incident admin` |

## Shared Harness Coverage

| Harness | Phase 1 evidence |
| --- | --- |
| Envelope consistency | `internal/testutil/httptestx/httptestx.go` plus all `httptestx.RequireSuccessEnvelope` and `httptestx.RequireErrorEnvelope` assertions in `internal/modules/auth/phase1_integration_test.go`, `cmd/server/main_phase1_process_test.go`, and `apps/web/src/Phase1Harness.tsx`. |
| Authorization re-derivation | `internal/modules/auth/phase1_integration_test.go::TestPhase1_AdminCredentialAuditAndScope_I_1_05` proves incident-admin scope does not widen into deployment-admin credential authority. |
| Mutation attribution and history emission | `internal/modules/auth/phase1_integration_test.go::TestPhase1_UserAdminAudit_I_1_03` and `internal/modules/auth/phase1_integration_test.go::TestPhase1_AdminCredentialAuditAndScope_I_1_05` use `httptestx.RequireMutationAttribution` against `deployment_admin_audit_events`. |
| Idempotent replay and divergent replay | `internal/modules/auth/phase1_integration_test.go::TestPhase1_CredentialStateTransitions_I_1_04`, `internal/modules/auth/phase1_integration_test.go::TestPhase1_BootstrapEnrollmentConsumption_I_1_04`, and `internal/modules/auth/phase1_integration_test.go::TestPhase1_UserAdminLifecycle_I_1_03`. |
| Closed-vocabulary rejection | `internal/modules/auth/phase1_request_test.go::TestPhase1_CredentialStateInspection_U_1_10` and `internal/modules/auth/phase1_request_test.go::TestPhase1_TOTPBootstrapAndSetupRules_U_1_12`. |
| Writable-string normalization | `internal/platform/authn/phase1_primitives_test.go::TestPhase1_LoginNormalizationAndPasswordExactness_U_1_02` and `internal/modules/auth/phase1_users_test.go::TestPhase1_UserPatchAndLastAdminGuard_U_1_08`. |
| WebSocket lifecycle | `internal/modules/auth/phase1_integration_test.go::TestPhase1_SessionRevocationClosesAttachedSocket_I_1_02` and `internal/modules/auth/phase1_integration_test.go::TestPhase1_AdminCredentialActions_I_1_05`. |
| Topology, preservation, and audit-source invariants | `internal/modules/auth/phase1_integration_test.go::TestPhase1_UserAdminAudit_I_1_03`, `internal/modules/auth/phase1_integration_test.go::TestPhase1_AdminCredentialAuditAndScope_I_1_05`, and `internal/testutil/httptestx/crosscutting.go::RequireSecretSafePayload`. |
