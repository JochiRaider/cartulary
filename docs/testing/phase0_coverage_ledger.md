# Phase 0 Coverage Ledger

This ledger is generated from `tools/phase0_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: infrastructure, deployment configuration, runtime roots, schema bootstrap, bootstrap-admin preflight, object-store reachability, and fail-closed startup only.
- Normative owners: Core 01 `§1`; Core 01 `§3.3.5.1`; Core 04 `§5–§8`; Core 04 `§12`; Core 04 `§9.0.1`.
- Authority: `tools/phase0_test_map.json` is the enforced Phase 0 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Browser E2E note: no Phase 0 browser-visible surface exists under `apps/web`, so authoritative `E-*` evidence lives on the real `cmd/server` process boundary.

## Authoritative Execution

- `backend-unit` selects authoritative `U-0-*` rows through `cartulary-runner.mjs go-target backend-unit`, with target-plan selection constrained by the Phase 0 manifest and `backend_unit` execution dependency.
- `backend-integration` selects authoritative `I-0-*` rows through `cartulary-runner.mjs go-target backend-integration`, with target-plan selection constrained by the Phase 0 manifest and `backend_integration` execution dependency.
- `backend-process` selects authoritative `E-0-*` rows only through the manifest-owned `backend-process` execution family.

## Support-Only Execution

- `internal/platform/postgres/postgres_phase0_support_test.go` runs through `backend-unit` with `TestSupportPhase0_` and is forbidden from claiming `U-0-*` identifiers.
- `internal/platform/objectstore/objectstore_phase0_support_test.go` runs through `backend-integration-support` with `TestSupportPhase0_` and is forbidden from claiming `I-0-*` identifiers.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-0-01` | `implemented` | `internal/platform/config/config_phase0_test.go::TestPhase0_ConfigDiscovery_U_0_01` | `backend_unit` | Configuration discovery enforces canonical selector precedence, rejects unknown keys, and fails closed on unsupported schema or profile combinations before startup composition. | Runtime-root compatibility and startup filesystem canonicalization belong to later Phase 0 rows. |
| `U-0-02` | `implemented` | `internal/platform/config/config_phase0_test.go::TestPhase0_RuntimeRoots_U_0_02` | `backend_unit` | Every supported deployment profile requires the full runtime-root set, rejects unknown or malformed bindings, accepts `cloud` managed-service bindings for the service-backed roots, and keeps the always-filesystem roots closed. | Lexical path rejection, overlap detection, and effective escape-write prevention are owned by `U-0-03`. |
| `U-0-03` | `implemented` | `internal/platform/config/config_phase0_test.go::TestPhase0_FilesystemRootPaths_U_0_03` | `backend_unit` | Filesystem-root validation rejects non-absolute, shell-expanded, `.`, `..`, NUL-bearing, overlapping, non-writable, and attempted escape-write paths against configured roots. | Real runtime fail-closed startup consequences remain integration and process evidence. |
| `U-0-04` | `implemented` | `internal/platform/config/config_phase0_test.go::TestPhase0_DisconnectedDefaults_U_0_04` | `backend_unit` | Disconnected defaults apply only where allowed and never synthesize missing or malformed roots away for other profiles. | Profile-specific binding-kind compatibility is covered by `U-0-02`. |
| `U-0-05` | `implemented` | `internal/app/runtime_phase0_test.go::TestPhase0_FailClosedStartup_U_0_05` | `backend_unit` | Invalid deployment config and bootstrap-preflight failure stop dependency wiring before HTTP, WebSocket, or job startup begins. | Whole-payload diagnostics and durable bootstrap side effects are integration or process evidence. |
| `U-0-06` | `implemented` | `internal/platform/config/config_phase0_test.go::TestPhase0_EnterpriseAuthenticationConfig_U_0_06` | `backend_unit` | Enterprise-authentication claim configuration defaults to unclaimed, rejects provider manifests for unclaimed deployments, requires an absolute provider manifest path when claimed, and preserves nested environment overlays. | Provider manifest file parsing, secret resolution, verifier wiring, and provider reconciliation are enterprise-authentication runtime evidence. |
| `U-0-07` | `implemented` | `internal/platform/bootstrap/bootstrap_phase0_test.go::TestPhase0_BootstrapManifestValidation_U_0_07` | `backend_unit` | Bootstrap manifest validation accepts only the v1 schema, defaults omitted `mfa_required` to true, and rejects explicit false, unknown members, and forbidden deployment-admin, incident, or provider fields. | Manifest filesystem errors and durable bootstrap persistence stay with higher Phase 0 layers. |
| `U-0-08` | `implemented` | `internal/platform/bootstrap/bootstrap_phase0_test.go::TestPhase0_BootstrapPreflight_U_0_08` | `backend_unit` | Bootstrap preflight queries bootstrap state first, skips manifest access when an active admin exists, and emits canonical diagnostics for missing, unreadable, non-regular, malformed, and recovery-failure cases. | Real Postgres-backed bootstrap creation and rollback belong to `I-0-04` through `I-0-06`. |
| `U-0-09` | `implemented` | `internal/platform/config/config_phase0_test.go::TestPhase0_ResourceLimits_U_0_09` | `backend_unit` | Resource-limit validation resolves defaults deterministically, enforces closed numeric domains, rejects unknown limit keys, and preserves the fixed public ceilings. | Runtime use of those limits on later import, archive, and preview routes belongs to later phases. |
| `U-0-10` | `implemented` | `internal/app/operator_migration_evidence_test.go::TestPhase0_MigrationEvidenceCommand_U_0_10` | `backend_unit` | Migration-history evidence parsing audits embedded source against the migration history manifest, emits redacted evidence-only JSON, and reports missing migration metadata without authorizing rewrite, squash, reset, or rebaseline work. | Real PostgreSQL goose ledger collection and DB-only applied-version findings are covered by `I-0-07`; any deployment-local operator wrapper is implementation support and does not close recovery conformance. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-0-01` | `implemented` | `internal/platform/postgres/postgres_phase0_test.go::TestPhase0_SchemaBootstrap_I_0_01` | `backend_integration` | Real PostgreSQL bootstrap creates the required extensions and base tables and reruns without schema drift or duplicate objects. | Migration-text style guards remain support-only. |
| `I-0-02` | `implemented` | `internal/platform/objectstore/objectstore_phase0_test.go::TestPhase0_ObjectStoreInitialization_I_0_02` | `backend_integration` | Disconnected `filesystem_root` object-store setup uses only `roots.object_storage.path`, ignores generic S3 service coordinates, and completes filesystem-backed write, read, stat, list, delete, upload-target, and path-escape checks. | Managed-service object-store binding coverage remains supplemental. |
| `I-0-03` | `implemented` | `internal/app/runtime_phase0_integration_test.go::TestPhase0_InvalidConfigNeverReachesReady_I_0_03` | `backend_integration` | Config failures on real Postgres and object-store services prevent ready state and stop handler, WebSocket, and job startup even when backing services are healthy. | Real process exit behavior and connection refusal remain `E-0-02`. |
| `I-0-04` | `implemented` | `internal/app/runtime_phase0_integration_test.go::TestPhase0_FirstAdminBootstrap_I_0_04` | `backend_integration` | Successful bootstrap commits exactly one active deployment admin, one bootstrap marker, one audit row, zero incident memberships, and a local-auth-only safe user shape before readiness; rollback leaves no partial state. | Real OS process-boundary readiness and diagnostics remain `E-0-03`. |
| `I-0-05` | `implemented` | `internal/app/runtime_phase0_integration_test.go::TestPhase0_BootstrapFailures_I_0_05` | `backend_integration` | Required bootstrap failures for missing, unreadable regular-file, non-regular, malformed, schema-invalid, and email-conflicting manifests block readiness and leave no partial bootstrap state. | Process-boundary refusal and exit diagnostics remain `E-0-04`. |
| `I-0-06` | `implemented` | `internal/app/runtime_phase0_integration_test.go::TestPhase0_BootstrapSkipAndRecovery_I_0_06` | `backend_integration` | Existing active admins skip stale or invalid manifests, while bootstrap-completion with zero active admins fails closed without implicit recovery. | Real process-boundary refusal remains `E-0-05`. |
| `I-0-07` | `implemented` | `internal/app/operator_migration_evidence_integration_test.go::TestPhase0_MigrationEvidenceCommand_I_0_07` | `backend_integration` | Against real PostgreSQL, migration-history evidence reports the migrated goose ledger current version and source audit through manifest head, classifies protected applied history, and surfaces synthetic DB-only ledger versions as evidence findings without mutating schema. | Deployment-local wrapper invocation is implementation support; rewrite, squash, reset, rebaseline authorization, and recovery operator conformance remain separate owner-policy decisions. |

## Process E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `E-0-01` | `implemented` | `cmd/server/main_phase0_e2e_test.go::TestPhase0_ReadyState_E_0_01` | `backend_process` | A fresh process reaches health and ready only after configuration validation, Postgres bootstrap, object-store reachability, and successful bootstrap preflight. | Browser-visible UI evidence is out of scope for Phase 0. |
| `E-0-02` | `implemented` | `cmd/server/main_phase0_e2e_test.go::TestPhase0_InvalidConfigDiagnostics_E_0_02` | `backend_process` | Invalid config exits non-zero, emits structured startup diagnostics, and exposes no HTTP readiness or WebSocket surface. | Bootstrap-manifest-specific failure variants remain `E-0-04`. |
| `E-0-03` | `implemented` | `cmd/server/main_phase0_e2e_test.go::TestPhase0_FirstAdminBootstrap_E_0_03` | `backend_process` | A real process requiring bootstrap becomes ready only after durable bootstrap success and leaves one deployment admin, one bootstrap marker, one audit row, and zero incident memberships. | Bootstrap failure matrix coverage belongs to `E-0-04` and `E-0-05`. |
| `E-0-04` | `implemented` | `cmd/server/main_phase0_e2e_test.go::TestPhase0_BootstrapFailures_E_0_04` | `backend_process` | Required bootstrap failures on the real process boundary block startup before HTTP, readiness, or WebSocket surfaces exist and emit canonical invalid-deployment diagnostics. | Unreadable regular-file manifest parity is a follow-on process-boundary improvement, not part of the current authoritative row. |
| `E-0-05` | `implemented` | `cmd/server/main_phase0_e2e_test.go::TestPhase0_BootstrapSkipAndRecovery_E_0_05` | `backend_process` | Stale or invalid manifests are ignored when an active deployment admin exists, while lost-admin recovery fails closed with `bootstrap_recovery_not_supported` and no exposed surface. | It does not prove every bootstrap failure reason at the process boundary; those remain `E-0-04`. |

## Shared Harness Coverage

| Harness | Phase 0 evidence |
| --- | --- |
| Startup diagnostics and real process boundary | `internal/testutil/diagnosticstest`, `internal/testutil/configtest`, and `internal/testutil/processtest` keep unit, integration, and process startup diagnostics on shared whole-payload goldens. |
| Fail-closed HTTP readiness and health gating | `internal/testutil/processtest.WaitForReady`, `RequireStatus`, and `RequireConnectionRefused` prove `/healthz` and `/readyz` behavior across success and failure flows. |
| Fail-closed WebSocket boundary | `internal/testutil/processtest.RequireWebsocketConnectionRefused`, built on `internal/testutil/wstest`, proves Phase 0 startup failures expose no WebSocket surface. |
| Mutation attribution, secret-safe payloads, and bootstrap auth-binding shape | `internal/testutil/auditassert`, `internal/testutil/securityassert`, and `internal/modules/auth/testsupport/bootstraptest.RequireBootstrapUserLocalAuthOnly` cover startup audit attribution, secret-safe payloads, and bootstrap-created-user auth-binding shape. |
| Real Postgres bootstrap harness | `internal/platform/postgres/postgres_phase0_test.go::TestPhase0_SchemaBootstrap_I_0_01` and the runtime integration suite prove migration and bootstrap state against real PostgreSQL. |
| Real object-store binding harness | `internal/platform/objectstore/objectstore_phase0_test.go::TestPhase0_ObjectStoreInitialization_I_0_02` proves disconnected `filesystem_root` object storage through the configured root path, while the real process suite proves runtime startup consumes validated root bindings. |

## Support-Only Evidence

- `internal/platform/postgres/postgres_phase0_support_test.go` keeps the migration-text regression guard; authoritative schema-bootstrap evidence stays `I-0-01`.
- `internal/platform/objectstore/objectstore_phase0_support_test.go` keeps managed-service object-store binding coverage; authoritative object-store reachability stays `I-0-02`.
- `tools/testservices/integration_test.go` remains harness-development noise and is intentionally outside Phase 0 traceability.
