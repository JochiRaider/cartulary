# Cartulary Progressive Implementation and Testing Guide

**Status**: Derived implementation-planning artifact  
**Authoritative sources**: Core 00–04 for implementation conformance; Core 05 for claim-bearing publication only  
**Traceability source**: Appendix F  
**Scope**: Base profile, with extension-profile testing hooks in Phases 11 and 12

---

## 1. How to read this guide

This guide is a **derived implementation-planning artifact**, not an independent source of contract truth. Core 00 through Core 04 are authoritative for implementation conformance. Core 05 is authoritative for claim-bearing publication of timed or fixture-sensitive criteria only. Appendix F is the traceability aid. When this guide and an owner section diverge, the owner section governs and this guide must be repaired.[^precedence][^claim-publication]

Use `docs/domain.md` for project-wide domain vocabulary and concept boundaries while planning or reviewing phase work. It is a navigation and terminology aid; it does not replace the owner sections or Appendix F traceability.

This guide preserves the phased implementation shape because it is useful for delivery planning, but the phase order is an **implementation aid**, not a parallel specification. A phase may group work for sequencing convenience. It does not change requirement ownership, route semantics, data-model rules, or conformance scope.[^precedence][^traceability]

Phase headers therefore name **primary owner sections**, not broad REQ blocks. The phase tables carry the exact REQ and AC identifiers for each planned test. Use the phase header to find the owner section. Use the row-level mappings to build the test and to prove traceability.[^precedence][^traceability]

This guide also resolves one dependency error from the prior version: reviewer-facing history, delete or restore, and rollback remain late-phase work, but the **minimal write-side mutation substrate** cannot wait until that phase. As soon as the first record-row mutation path exists, it must already emit attributed mutations, maintain projections, and satisfy normalized idempotency and optimistic-concurrency contracts. Phase 7 now completes the reviewer-facing history and rollback surface instead of introducing mutation history for the first time.[^traceability]

### 1.1 Test categories used in this guide

- **Unit test**: verifies one module, validation rule, serializer, reducer, mapper, or deterministic algorithm in isolation.
- **Integration test**: verifies one or more modules against real backing services or real transport boundaries such as PostgreSQL, object storage, or the WebSocket boundary.
- **E2E test**: verifies a user-observable flow through the full deployed stack against the visible contract and the cited ACs.
- **Visual regression test**: verifies deterministic browser-rendered workbook states against stable screenshot or DOM-visual fixtures. It is developer-gate evidence for UI drift. It is not a claim-bearing benchmark unless Core 05 publication requirements are separately satisfied.

Visual golden refreshes follow `docs/guides/cartulary_visual_golden_maintenance.md`. That guide is implementation support only; it does not move visual snapshot refresh authority into the Core 00-04 product contract or current harness conformance profile.

Browser E2E tests that query backend state after a UI mutation must wait for the mutation boundary before querying. Prefer a route-specific `page.waitForResponse()` started before the click for the expected mutating route, or wait for the workbook save state that is set only after the route response and required refresh complete. URL, focus, scroll, and visibility assertions are context-continuity checks; they are not persistence barriers.

For focus and scroll continuity checks, wait for the mutation's required same-surface refresh boundary before asserting continuity. A visible text change from an intermediate render is not enough when the action performs a follow-up refresh, such as create-from-mention entity refresh.

Runner summaries use the following accounting buckets. `authoritative` is phase completion evidence with owned IDs. `support` is phase-owned support evidence. `raw` is an explicitly declared raw aggregate suite, owned through `tools/execution_topology_manifest.json`. `tooling_support` is helper, harness, config, or runner-support coverage. `unowned_regression` is intentional product regression coverage that has not been promoted into a phase-owned row. `unmapped` is reserved for unexpected executed tests that need an ownership decision. Public Make-owned verification targets must finish with `unmapped=0`; an otherwise passing target with unmapped executed tests fails as `failure_class="harness"` and `failure_reason="test_accounting_unmapped"`. Non-phase classifications belong in `tools/test_accounting_classification.json`; do not add filename heuristics when a manifest rule can express the ownership. Use `make explain-run RESULTS_DIR=<run-root> TARGET=<target> DETAIL=accounting` to group retained inventory during triage.

Use this triage order for new or residual tests: map the exact title or symbol when the test is canonical conformance evidence for a base or frontend row; classify it as `support`, `tooling_support`, or `unowned_regression` when it is intentional but non-authoritative coverage; delete it only when owner review finds no continuing value or a better-owned test fully supersedes it. A title containing a base row ID or `FE-*` row ID is a traceability claim, so it must either match a live map exactly or remove the row-like prefix before residual classification.

### 1.1.1 Service-backed fixture modes

Service-backed Go tests must keep service ownership centralized in `tools/testservices`; phase helpers choose only the fixture isolation mode used inside that already-running Postgres and object-store pair.

- Use isolated per-test Postgres template clones for startup, migration, rollback, process-boundary, HTTP/runtime, and unclear isolation cases.
- Use transaction-backed Postgres fixtures for store-only tests once their seed and assertion helpers run through the shared `postgres.DB` test surface.
- Use package-reused Postgres databases only for harness self-tests or rows with explicitly declared dirty-table reset scope; broad mutable-table resets are not a default fixture mode.
- Use package-reused object-store buckets for ordinary route tests; helpers clear object contents before each test. Prefix cleanup is available for tests that can route all object keys through a unique prefix.
- Treat fixture churn diagnostics in `service-scope.json` and `target-summary.json` as the source of truth when deciding whether a test needs stronger isolation. For failed harness runs, read compact `failure_class` and `failure_reason` fields in phase, target, scheduler, and run summaries before reading detailed logs: `product` means product evidence failed, `infra` means backing service/runtime readiness failed, `harness` means orchestration failed, `timing` means duration drift failed, and `artifact` means expected reports or cleanup artifacts failed.

### 1.1.2 Toolchain output policy

The repository task surface uses `CARTULARY_OUTPUT_MODE=summary` by default. Successful public verification targets should print only bounded `[RESULT]` and `[ARTIFACTS]` lines, keep child stdout/stderr in target-owned artifacts, and keep successful stderr empty. `CARTULARY_OUTPUT_MODE=ci` enables the same bounded summaries plus CI progress lines; provider-neutral CI entrypoints should invoke `make --no-print-directory ci` and set provider-specific environment explicitly when needed. `CARTULARY_OUTPUT_MODE=verbose` streams child detail for investigation, and `CARTULARY_OUTPUT_MODE=debug` is reserved for wrapper telemetry.

`CARTULARY_OUTPUT_MODE=machine` is a strict machine contract: each top-level command that accepts machine mode emits exactly one parseable JSON object followed by LF, and it must not stream raw child output, scheduler progress prose, duplicate run/target JSON objects, or pointer-only stdout. The canonical summary requires non-null RFC3339 `started_at` and `completed_at` timestamps, `result_root`, `run_id`, `run_root`, compact artifact refs, work-unit or evidence-target accounting, `failure_class`, `failure_reason`, rerun commands, and investigation commands.

Failure summaries must name the most specific available owner: `work_unit` for scheduler or sequence work, `child_target` for aggregate evidence, `failure_class`, `failure_reason`, duration, `run_root`, relative summary/log artifact refs, rerun command, and investigation command. Children skipped because a scheduler stopped after an earlier failure are skipped evidence with preserved skip provenance, not missing artifact failures. Use `-` only when the identifier or artifact genuinely does not exist.

When adding or changing public targets, update the testing harness public registry and `tools/task_surface_manifest.json`; update `tools/execution_topology_manifest.json` only when scheduler topology or child work changes. Declare `artifact_policy=none` for commands that do not create centralized artifacts. Targets that declare `tool_run_summary`, `run_and_target_summaries`, or `scheduler_and_tool_run_summaries` must route through the centralized output path or otherwise produce the promised artifact. Policy tests must cover success-output budgets for at least one leaf target, one sequence target, one scheduler target, one representative failure path, invalid usage, and machine-mode exactly-one-JSON behavior.

### 1.2 Conformance posture

Phases 0 through 10 are the base-profile implementation sequence. A base-profile claim MUST satisfy the current Base claim manifest in Core 04 rather than any historical endpoint shorthand such as `AC-299` alone. Phases 11 and 12 are intentionally not base-profile phases. They are extension-profile hook sections that keep extension work aligned to the current extension route families and claim manifests. Core 05 remains a separate normative companion for claim-bearing publication of timed or fixture-sensitive criteria and does not broaden Base Profile or extension-profile implementation conformance.[^base-manifest][^traceability][^claim-publication]

---

## 2. Phase 0 — Infrastructure, deployment configuration, and schema bootstrap

### 2.0.1 Scope

This phase establishes the deployable shell and startup control plane:

- modular-monolith application shell,
- required services and runtime roots,
- deployment-configuration artifact discovery and validation,
- canonical disconnected-layout defaults,
- fail-closed startup,
- schema bootstrap idempotency,
- first-deployment-admin bootstrap preflight,
- bootstrap-manifest validation and skip semantics,
- fail-closed lost-admin recovery when bootstrap-completion state exists but no active deployment admin remains,
- resource-limit registry validation,
- deployment-local applied-migration evidence capture,
- object-store reachability.

No domain routes beyond health or startup diagnostics should be treated as complete at the end of this phase. Bootstrap-created administrators enter the ordinary credential lifecycle later in Phase 1 rather than through a special startup-only auth path.[^base-manifest][^core01-routes]

### 2.0.2 Primary owner sections

- Core 01 §1 Architecture pattern
- Core 01 §2.1A schema ownership and migration history policy
- Core 01 §3.3.5.1 deployment-local user administration and bootstrap-admin manifest contract
- Core 04 §5–§8 deployment topology, runtime roots, container boundary, and required services
- Core 04 §12 deployment-configuration contract
- Core 04 §9.0.1 base-claim manifest for `AC-294..AC-298`, `AC-320`, and `AC-343..AC-346`

### 2.0.3 Unit tests

| ID     | Test                                                                                                                                                                                                                                                                           | Exact REQs                                                 | Exact ACs      |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------- | -------------- |
| U-0-01 | Configuration discovery follows the canonical precedence chain and rejects unknown or mismatched `config_schema_id` / `deployment_profile` combinations before startup.                                                                                                        | REQ-04-066..REQ-04-071                                     | AC-294, AC-298 |
| U-0-02 | The runtime-root registry requires exactly the configured root keys for the active profile and rejects missing or malformed bindings.                                                                                                                                          | REQ-04-058, REQ-04-069, REQ-04-071..REQ-04-073, REQ-04-077 | AC-295         |
| U-0-03 | Filesystem-root path validation canonicalizes paths and rejects traversal, root escape, and invalid overlap cases.                                                                                                                                                             | REQ-04-059, REQ-04-074..REQ-04-075, REQ-04-077             | AC-296         |
| U-0-04 | Disconnected-layout defaults are applied only where the owner section allows them. Missing required roots are not silently defaulted away.                                                                                                                                     | REQ-04-067, REQ-04-069, REQ-04-071..REQ-04-076             | AC-297         |
| U-0-05 | Invalid deployment configuration blocks dependency wiring entirely, and bootstrap-preflight failures still block HTTP startup, WebSocket startup, and background-job startup.                                                                                                     | REQ-04-077..REQ-04-078                                     | AC-298         |
| U-0-06 | Enterprise-authentication claim configuration defaults to unclaimed, rejects provider manifests for unclaimed deployments, requires an absolute provider manifest path when claimed, and supports deployment overlays.                                                          | REQ-04-115..REQ-04-116                                     | AC-433, AC-436 |
| U-0-07 | Bootstrap manifest validation accepts only `cartulary.bootstrap_admin.v1`, defaults omitted `mfa_required` to `true`, rejects explicit `false` or unknown top-level members, and never permits incident membership, provider binding, or client-chosen deployment-admin state. | REQ-01-530..REQ-01-532                                     | AC-343, AC-344 |
| U-0-08 | Startup preflight queries active deployment-admin state and bootstrap-completion state first, skips manifest consumption when an active deployment admin already exists, and fails closed when completion state exists but no active deployment admin remains.                   | REQ-01-533..REQ-01-535, REQ-04-087..REQ-04-092             | AC-345, AC-346 |
| U-0-09 | The resource-limit registry resolves omitted defaults deterministically, enforces its closed numeric domains, rejects unknown limit keys, and never widens the fixed public ceilings for `sort[]`, `filters[]`, `changes[]`, or `collection_actions_v1.actions[]`.             | REQ-04-066, REQ-04-077, REQ-04-079..REQ-04-081             | AC-320         |
| U-0-10 | The deployment-local operator migration-evidence command validates capture inputs, audits embedded migration source against the migration history manifest, emits secret-safe evidence-only JSON, and reports missing goose metadata without authorizing rewrite work.          | Core 01 §2.1A, REQ-04-077                                  | AC-231, AC-298 |

Schema-bootstrap idempotency is authoritative integration evidence at `I-0-01`. Any migration-text regression guard stays support-only and is not part of the authoritative Phase 0 traceability map.

### 2.0.4 Integration tests

| ID     | Test                                                                                                                                                                                                                                                                                 | Exact REQs                                                 | Exact ACs      |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------- | -------------- |
| I-0-01 | Against a real PostgreSQL instance, bootstrap creates the required extensions and base schema objects and can be rerun without drift.                                                                                                                                                | REQ-01-001..REQ-01-003, REQ-04-061                         | AC-231         |
| I-0-02 | Against a real S3-compatible object store, the disconnected `filesystem_root` storage binding resolves through the generic `CARTULARY_S3_*` contract and performs a minimal round-trip object write and read.                                                                      | REQ-04-058..REQ-04-061                                     | AC-231         |
| I-0-03 | A configuration with a path-validation failure or missing required root never reaches ready state, even when PostgreSQL and object storage are otherwise healthy.                                                                                                                    | REQ-04-074..REQ-04-078                                     | AC-296, AC-298 |
| I-0-04 | A fresh deployment with a valid bootstrap manifest creates exactly one active local deployment admin, one bootstrap-completion marker, one deployment-local administrative audit event, and only the ordinary local auth-binding summary before ready state.                         | REQ-01-534, REQ-02-007..REQ-02-008, REQ-04-028, REQ-04-038 | AC-343         |
| I-0-05 | When bootstrap is required, a missing-path, unreadable regular-file, non-regular, malformed, schema-invalid, or email-conflicting bootstrap manifest fails before ready state and leaves no partial user, no partial bootstrap-completion marker, and no incident membership behind. | REQ-01-530..REQ-01-535, REQ-04-087..REQ-04-092             | AC-344         |
| I-0-06 | Existing active deployment-admin state skips bootstrap consumption even when the configured manifest path is stale or invalid; completion-state with zero active deployment admins fails closed and creates no implicit replacement admin.                                           | REQ-01-533..REQ-01-535, REQ-04-090..REQ-04-092             | AC-345, AC-346 |
| I-0-07 | Against real PostgreSQL, the deployment-local migration-evidence support command uses local execution authority, reports the migrated goose ledger current version, classifies protected applied history, and surfaces DB-only ledger versions as findings without mutation.     | Core 01 §2.1A, REQ-04-077                                  | AC-231, AC-298 |

### 2.0.5 E2E tests

| ID     | Test                                                                                                                                                                                                                                                                                  | Exact REQs                                     | Exact ACs      |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | -------------- |
| E-0-01 | A fresh minimum deployment starts, passes health checks, and exposes a ready state only after PostgreSQL, object storage, and configuration validation all succeed.                                                                                                                   | REQ-04-054..REQ-04-061, REQ-04-066..REQ-04-078 | AC-294..AC-298 |
| E-0-02 | A deployment with an invalid config artifact never exposes a healthy application surface and exits with structured startup diagnostics.                                                                                                                                               | REQ-04-077..REQ-04-078                         | AC-298         |
| E-0-03 | A fresh deployment requiring bootstrap becomes ready only after successful bootstrap preflight and exposes exactly one active deployment admin with no incident memberships and only the ordinary local auth-binding summary.                                                         | REQ-01-533..REQ-01-534, REQ-04-028, REQ-04-090 | AC-343         |
| E-0-04 | Required-manifest failures on the real process boundary, including non-regular, malformed, and email-conflicting cases, block startup before HTTP, readiness, or WebSocket surfaces exist and surface the bootstrap-specific `invalid_deployment_config` reason code.          | REQ-01-530..REQ-01-535, REQ-04-087..REQ-04-092 | AC-344         |
| E-0-05 | When an active deployment admin already exists, startup ignores a stale or invalid bootstrap manifest and proceeds; when no active deployment admin exists but bootstrap-completion state already exists, startup fails closed with `reason_code='bootstrap_recovery_not_supported'` and no `/healthz`, `/readyz`, or WebSocket surface. | REQ-01-533..REQ-01-535, REQ-04-090..REQ-04-092 | AC-345, AC-346 |

---

## 3. Phase 1 — Authentication, sessions, and deployment-local user administration

### 3.1.1 Scope

This phase establishes the authenticated shell:

- local account login,
- Argon2id password hashing,
- TOTP MFA,
- CSRF protection,
- session lifecycle and concurrency cap,
- session inspection and logout,
- credential-state inspection,
- self-service password change,
- TOTP begin and complete for first enrollment and replacement enrollment,
- bootstrap-token boundaries,
- deployment-local user creation and patch,
- deployment-admin password reset, TOTP reset, and explicit revoke-all,
- immediate socket revocation behavior for revoked sessions.

Incident-specific work does not begin until the bounded credential lifecycle and deployment-local account boundary are trustworthy.[^base-manifest][^core01-routes]

Repository execution for this phase is manifest-driven. `tools/phase1_test_map.json` is the executable authoritative inventory for Phase 1 completion, and `make test`, `make check`, and `make ci` must both select and verify every mapped authoritative row. `make backend-unit` carries the pure Phase 1 handler-unit rows, while `make backend-store` carries the service-backed Phase 1 unit rows `U-1-05`, `U-1-06`, and `U-1-08`; `make frontend-unit` carries `U-1-14..U-1-17`; `make backend-integration` carries `I-1-01..I-1-06`; and the ordinary-shell browser rows continue to execute through the Phase 1 Playwright manifest under `browser_functional` and `browser_stateful`. Support-only backend sweeps and smoke coverage remain valuable, but they are not substitutes for the manifest-owned completion rows.

Shared harness owners for this phase are explicit:

- `internal/modules/auth/testsupport/routetest.PublicRouteInventory()` owns the auth public-route inventory together with the shared `surface_envelope`, `bootstrap_boundary`, `csrf`, `replay_stored_payload`, `mutation_audit`, `session_revocation`, `authorization_rederivation`, and `request_contracts` sweeps.
- `internal/modules/auth/phase1_support_integration_test.go` owns the Phase 1 support-only inventory sweeps and must stay support-only even as the route inventory expands.
- `apps/web/src/app/App.phase1.test.tsx` owns the authoritative ordinary-shell Phase 1 frontend-unit rows. `apps/web/src/app/App.phase1.support.test.tsx` remains support-only and must not claim `U-1-*` identifiers.

### 3.1.2 Primary owner sections

- Core 01 §3.3.2 session and authentication routes
- Core 01 §3.3.2.2 credential lifecycle and TOTP bootstrap routes
- Core 01 §3.3.5.1 deployment-local user administration
- Core 01 §3.3.10 WebSocket session lifecycle consequences
- Core 04 §1 authentication model and session lifecycle boundaries
- Core 04 §3 attribution and administrative audit requirements

### 3.1.3 Unit tests

| ID     | Test                                                                                                                                                                                                                                                                                                                                                     | Exact REQs                                     | Exact ACs                      |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------ |
| U-1-01 | `POST /api/v1/auth/login` rejects malformed top-level members, malformed `second_factor`, and unsupported base-profile MFA kinds with `invalid_auth_request`.                                                                                                                                                                                            | REQ-01-025, REQ-01-234                         | AC-244, AC-247, AC-249         |
| U-1-02 | `username` is trimmed for login matching, while `password` is compared exactly as supplied after JSON decoding.                                                                                                                                                                                                                                          | REQ-01-025                                     | AC-244, AC-245, AC-246         |
| U-1-03 | Successful login creates one server-side session record with `authenticated_at`, `last_qualifying_activity_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at=min(idle, absolute)`.                                                                                                                                                  | REQ-01-026..REQ-01-028, REQ-04-005..REQ-04-011 | AC-123, AC-131, AC-156..AC-163 |
| U-1-04 | `GET /api/v1/auth/session` returns the singleton session resource, rejects pagination members, and does not extend idle expiry by itself.                                                                                                                                                                                                                | REQ-01-027..REQ-01-028                         | AC-123, AC-130                 |
| U-1-05 | A sixth concurrent session revokes the least-recently-used non-current session and records `concurrency_limit`.                                                                                                                                                                                                                                          | REQ-04-013..REQ-04-014                         | AC-131, AC-136, AC-163         |
| U-1-06 | `POST /api/v1/auth/logout` revokes only the current session. Password change, MFA reset, account disablement, or explicit revoke-all revokes every active session for that user.                                                                                                                                                                         | REQ-01-029, REQ-04-015..REQ-04-016             | AC-131, AC-136, AC-156..AC-163 |
| U-1-07 | `POST /api/v1/users` rejects unsupported `auth_kind`, normalizes writable `email` and `display_name`, applies the required defaults (`mfa_required=true`, `is_deployment_admin=false`, server-managed `is_active=true`), and returns no `initial_password` echo.                                                                                     | REQ-01-117..REQ-01-123                         | AC-175..AC-177                 |
| U-1-08 | `PATCH /api/v1/users/{user_id}` enforces `base_user_version`, normalizes writable `email` and `display_name`, and rejects demotion or deactivation of the last active deployment admin.                                                                                                                                                                  | REQ-01-124..REQ-01-126                         | AC-178..AC-180                 |
| U-1-09 | Cookie-authenticated state-changing routes fail closed on missing or invalid CSRF proof.                                                                                                                                                                                                                                                                 | REQ-04-003                                     | AC-123, AC-130                 |
| U-1-10 | `GET /api/v1/auth/credential-state` returns only the safe credential-state shape, uses the closed `totp.state` vocabulary, and rejects bootstrap-token use on ordinary auth inspection routes with the route-owned rejection code.                                                                                                                       | REQ-01-523, REQ-04-083..REQ-04-084             | AC-335, AC-339                 |
| U-1-11 | `POST /api/v1/auth/password/change` verifies `current_password` exactly as supplied after JSON decoding, requires a current TOTP assertion when an active factor exists, revokes all active sessions on success, and records only secret-safe deployment-local audit or idempotency state.                                                               | REQ-01-524, REQ-04-016, REQ-04-086             | AC-338                         |
| U-1-12 | `POST /api/v1/auth/mfa/totp/begin` and `POST /api/v1/auth/mfa/totp/complete` accept exactly one auth mode, issue seed material only on begin, preserve begin idempotency within one auth scope plus `client_txn_id`, consume bootstrap tokens on successful complete, and distinguish first enrollment from replacement-enrollment session consequences. | REQ-01-522, REQ-01-525..REQ-01-526, REQ-04-084 | AC-336, AC-337, AC-339         |
| U-1-13 | `POST /api/v1/users/{user_id}/password/reset`, `POST /api/v1/users/{user_id}/mfa/totp/reset`, and `POST /api/v1/users/{user_id}/sessions/revoke-all` are callable only by `deployment_admin`, preserve their route-owned state consequences, and do not widen incident-scoped authorization.                                                             | REQ-01-527..REQ-01-529, REQ-04-085..REQ-04-086 | AC-340..AC-342                 |
| U-1-14 | The ordinary shell starts anonymous, sends the Phase 1 login request, refreshes session and credential-state resources after success, and keeps deployment-user controls denied for non-deployment-admin sessions.                                                                                                                                       | REQ-01-023..REQ-01-031, REQ-04-001..REQ-04-017 | AC-123, AC-130                 |
| U-1-15 | The ordinary shell follows `mfa_setup_required -> bootstrap_token -> totp/begin -> totp/complete`, sends the bootstrap-token requests with the ordinary client surface, and proves TOTP completion alone does not issue a session.                                                                                                                     | REQ-01-522, REQ-01-525..REQ-01-526, REQ-01-536 | AC-334, AC-336, AC-337, AC-347 |
| U-1-16 | The ordinary account-security panel issues password-change and TOTP-enrollment requests, surfaces route-owned failures on the shell, and refreshes back to the anonymous shell after success revokes the session.                                                                                                                                       | REQ-01-524..REQ-01-526, REQ-04-016, REQ-04-083 | AC-336, AC-338, AC-339         |
| U-1-17 | The ordinary deployment-admin panel creates and loads users, sends versioned patch requests, and surfaces `user_version_conflict` plus `last_deployment_admin` on the shell without relying on debug-only harness views.                                                                                                                                | REQ-01-117..REQ-01-126                         | AC-175..AC-180                 |

### 3.1.4 Integration tests

| ID     | Test                                                                                                                                                                                                                                                        | Exact REQs                                                                         | Exact ACs                      |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | ------------------------------ |
| I-1-01 | Login, session inspection, idle sliding, and logout all persist against a real PostgreSQL session store.                                                                                                                                                    | REQ-01-023..REQ-01-031, REQ-04-005..REQ-04-017                                     | AC-123, AC-131, AC-156..AC-163 |
| I-1-02 | Session revocation pushes `session_revoked` to an attached canonical incident WebSocket owned by that session and then closes the connection.                                                                                                                | REQ-01-029, REQ-01-250..REQ-01-277, REQ-04-015..REQ-04-016                         | AC-131, AC-136, AC-156..AC-163 |
| I-1-03 | Deployment-admin user creation and patch persist the expected defaults, version changes, and administrative audit records.                                                                                                                                  | REQ-01-117..REQ-01-126, REQ-04-038                                                 | AC-175..AC-180, AC-231         |
| I-1-04 | Credential-state inspection, TOTP begin and complete, and password change persist the expected deployment-local credential-state transitions and revoke sessions according to the route-owned rules.                                                        | REQ-01-523..REQ-01-526, REQ-04-016, REQ-04-083..REQ-04-084                         | AC-335..AC-339                 |
| I-1-05 | Deployment-admin password reset, TOTP reset, and revoke-all persist safe administrative audit records, update or preserve credential state exactly as required, and revoke attached sockets immediately.                                                    | REQ-01-527..REQ-01-529, REQ-01-250..REQ-01-277, REQ-04-016, REQ-04-085..REQ-04-086 | AC-340..AC-342                 |
| I-1-06 | A bootstrap token is accepted only by `POST /api/v1/auth/mfa/totp/begin` and `POST /api/v1/auth/mfa/totp/complete`; the same token fails closed on session, credential-state, ordinary incident routes, and canonical `/ws/v1/incidents/{incident_id}` with the route-owned rejection code. | REQ-01-522..REQ-01-523, REQ-04-084                                                 | AC-334, AC-339                 |

### 3.1.5 E2E tests

| ID     | Test                                                                                                                                                                                                                         | Exact REQs                                     | Exact ACs                      |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------ |
| E-1-01 | A local user logs in, retrieves the session resource, and sees `memberships[]`, expiry fields, and `provider_type='local'`.                                                                                                  | REQ-01-023..REQ-01-031, REQ-04-001..REQ-04-017 | AC-123, AC-130                 |
| E-1-02 | An MFA-required account receives `mfa_required` when `second_factor` is omitted, succeeds with a valid TOTP code, and fails with `invalid_second_factor` for a wrong but structurally valid code.                            | REQ-01-025, REQ-01-234, REQ-04-001             | AC-246, AC-248, AC-249         |
| E-1-03 | Invalid credentials return `invalid_credentials` without creating a session cookie.                                                                                                                                          | REQ-01-025, REQ-01-234                         | AC-245                         |
| E-1-04 | Session idle expiry causes subsequent authenticated requests to fail closed until the user logs in again.                                                                                                                    | REQ-04-008..REQ-04-012                         | AC-131, AC-136                 |
| E-1-05 | A deployment admin creates a user, patches it, and sees optimistic-concurrency enforcement and the last-admin guard.                                                                                                         | REQ-01-117..REQ-01-126                         | AC-175..AC-180                 |
| E-1-06 | A bootstrap-created or TOTP-reset local user follows the ordinary `mfa_setup_required -> bootstrap_token -> totp/begin -> totp/complete -> ordinary login` sequence, and first-time TOTP completion alone issues no session. | REQ-01-522, REQ-01-525..REQ-01-526, REQ-01-536 | AC-334, AC-336, AC-337, AC-347 |
| E-1-07 | A current user changes their password, must satisfy current-password and current-TOTP requirements when applicable, loses the current session immediately, and must log in again with the new password.                      | REQ-01-524, REQ-04-016, REQ-04-083             | AC-338                         |
| E-1-08 | Only a deployment admin can reset another user's password, reset another user's TOTP factor, or revoke all of that user's sessions; a non-deployment-admin incident admin cannot perform those route actions.                | REQ-01-527..REQ-01-529, REQ-04-085             | AC-340..AC-342                 |
| E-1-09 | The ordinary authenticated landing screen creates an incident through `POST /api/v1/incidents`, lists it, opens it by `incident_id`, and renders the workbook shell with current role `admin`.                              | REQ-01-154..REQ-01-160, REQ-04-021..REQ-04-026 | AC-170..AC-180, AC-231         |
| E-1-10 | A stale selected incident whose membership is removed fails closed back to the landing screen, disappears from the visible incident list, preserves the account session, and still allows another visible incident to open.  | REQ-01-127..REQ-01-137, REQ-04-021..REQ-04-026 | AC-175..AC-180, AC-231         |
| E-1-11 | A reviewer-visible incident edit submitted after current-role demotion fails through the public `authorization_denied` envelope, renders only the public error code, and reloads with the lower current role.                 | REQ-01-127..REQ-01-137, REQ-04-021..REQ-04-026 | AC-175..AC-180, AC-231         |
| E-1-12 | Deployment-admin revoke-all invalidates a target browser session, the next public session refresh returns the ordinary shell to login, and fresh login preserves unchanged incident membership access.                      | REQ-01-029, REQ-01-529, REQ-04-016, REQ-04-085 | AC-131, AC-136, AC-340..AC-342 |

Phase 1 list-search implementation addendum: when `GET /api/v1/users` implementation work lands, promote manifest rows for user-list search/filter validation, cursor query mismatch, deployment-admin gating, identifier-shaped search such as a full hyphenated `user_id` matched through ordinary `list_search_v1` source-token prefix semantics, the `limit=1` later-page match fixture, and deployment-admin frontend stale-response handling. Those rows should map to REQ-01-117, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-581..REQ-01-584, REQ-04-038, AC-415, and AC-417.

---

## 4. Phase 2 — Incidents, memberships, and the incident-scoped control envelope

### 4.2.1 Scope

This phase establishes incident-scoped administration and the first stable incident-level API behavior:

- incident create, list, get, and patch,
- creator bootstrap membership,
- incident membership create, patch, and delete,
- role-gated incident authorization,
- common response-envelope expectations on first incident routes,
- deployment-scoped extension discovery,
- reserved-but-unclaimed extension-family dispatch,
- incident create and patch idempotency or optimistic versioning,
- automatic creation of workbook preference objects at incident create.

This phase does **not** yet complete record-row history for workbook records. It also does **not** migrate Phase 2 incident mutation attribution onto `change_sets`; Phase 2 still proves owner-level mutation artifacts on the existing audit substrate, and any storage migration is separate owner work. It does establish the expectation that every new mutating route from this point forward must already satisfy its route-owned validation, idempotency, versioning, authorization, audit, and route-family discovery contracts.[^base-manifest][^core01-routes]

Repository execution for this phase is manifest-driven. `tools/phase2_test_map.json` is the executable authoritative inventory for Phase 2 completion, and `make test`, `make check`, and `make ci` must both select and verify every mapped authoritative row. Authoritative Go unit execution is split by manifest `execution_dependency`: `make backend-unit` carries `U-2-01`, `U-2-05`, `U-2-06`, `U-2-08`, `U-2-09`, and `U-2-10`, while `make backend-store` carries `U-2-02`, `U-2-03`, `U-2-04`, `U-2-07`, and `U-2-14`. `make backend-unit` also executes manifest-declared support-only unit seams, including membership patch/delete decoder support and workbook-preference PUT decoder plus OpenAPI contract support. Authoritative browser rows execute through `browser_functional` under `make browser-e2e-webserver-backed`. `backend-integration-support` emits support coverage through manifest-owned execution families rather than a second independent runtime, `make browser-e2e-support` is a helper slice already exercised inside `make browser-e2e-webserver-backed`, and `apps/web/src/app/Phase2Harness.tsx` remains a debug probe surface only. Runtime-backed HTTP conformance checks, extra regressions, process smoke coverage, and browser-authenticated request probes remain valuable support coverage, but they are supplemental evidence selected through the manifest-owned aggregate targets.

Shared harness owners for this phase are explicit:

- Owner route inventories under incidents, workbook, Timeline, records, entities, indicators, and extensions compose the Phase 2 public-envelope and role-control sweeps through `internal/testutil/routeinventory`; no single product facade owns another module's routes.
- `internal/modules/incidents/testsupport/routetest` contains only incident and membership routes. Workbook preferences, view queries, record actions, and collaboration behavior remain with their semantic owners.
- `internal/modules/incidents/testsupport/mutationtest` owns incident resource and membership mutation traceability without exposing the underlying audit table at call sites.
- `internal/modules/incidents/testsupport/storetest` owns rollback-backed incident store behavior; auth seeding comes from auth support, while `internal/testutil/appsupport.NewIncidentStore` owns cross-module store composition.
- `apps/web/src/app/App.landing.test.tsx` and `apps/web/src/app/IncidentAdminPanel.test.tsx` own the authoritative frontend-unit Phase 2 rows. They complement, not replace, Playwright `E-2-01..E-2-03`.

### 4.2.2 Primary owner sections

- Core 01 §3.3.1 versioning and compatibility
- Core 01 §3.3.3 route families and runtime extension discovery
- Core 01 §3.3.5.1 incident membership routes
- Core 01 §3.3.5.2 saved-view and workbook-preference routes
- Core 01 §3.3.5.3 incident create, list, get, and patch
- Core 02 §4.5 current-profile incident structured metadata
- Core 02 §18 incident status and TLP closed vocabularies
- Core 04 §2 authorization model
- Core 04 §3 attribution and administrative audit requirements

### 4.2.3 Unit tests

| ID     | Test                                                                                                                                                                                                                                                                                                                          | Exact REQs                                     | Exact ACs                                      |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| U-2-01 | `POST /api/v1/incidents` accepts only the declared top-level members, rejects `initial_memberships`, and normalizes `incident_key` using trim + NFC before uniqueness checks.                                                                                                                                                 | REQ-01-154..REQ-01-167, REQ-02-015..REQ-02-016 | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220 |
| U-2-02 | `Store.CreateIncident` commits the creator bootstrap membership as incident `admin`, creates both incident-wide and per-user workbook-preference rows, and workbook-preference PUT upserts preserve structural no-op timestamps plus default actors.                                                                                                                                                                      | REQ-01-156..REQ-01-159, REQ-01-145..REQ-01-151 | AC-170..AC-174, AC-150, AC-153                 |
| U-2-03 | The committed incident-create result returns a stable location value rooted at `/api/v1/incidents/{incident_id}`. HTTP header emission stays support coverage.                                                                                                                                                                  | REQ-01-160                                     | AC-170..AC-174                                 |
| U-2-04 | Incident create idempotency is keyed by `(actor_user_id, client_txn_id)` at the durable store boundary. Replay of the same normalized request returns the original result and keeps the incident row, bootstrap membership, workbook preferences, owner mutation artifacts, and route idempotency stable. Divergent replay returns `client_txn_conflict`. | REQ-01-161..REQ-01-167                         | AC-170..AC-174, AC-219..AC-220                 |
| U-2-05 | Incident patch accepts only `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`; requires `base_incident_version`; sets, replaces, clears via `null`, clears normalized-empty text fields, rejects empty or alias TLP values, rejects immutable fields without partial state, and treats a structurally valid no-op as version-stable. | REQ-01-168..REQ-01-180, REQ-01-491.1, REQ-02-015, REQ-02-222..REQ-02-223 | AC-170..AC-174, AC-211..AC-214 |
| U-2-06 | Membership create requires exactly one of `user_id` or `email`, uses the closed role vocabulary, resolves targets by lookup only, returns canonical `user_not_found` or `user_inactive`, and never auto-creates a user or invitation.                                                                                      | REQ-01-127..REQ-01-132                         | AC-175..AC-180                                 |
| U-2-07 | Store-backed membership patch and delete reject stale `base_membership_version`, and same-role membership patch keeps `membership_version` stable. Pure decoder shape and `last_incident_admin` guard assertions stay on executed support-only unit coverage.                                                                   | REQ-01-133..REQ-01-137                         | AC-178..AC-180                                 |
| U-2-08 | The incident access decision helper returns `incident_not_found` for no membership and `authorization_denied` for insufficient incident role. Deployment-admin route breadth stays paired with the control-boundary inventory.                                                                                               | REQ-04-028..REQ-04-030                         | AC-178..AC-180, AC-261                         |
| U-2-09 | `GET /api/v1/extensions` is a singleton deployment-scoped discovery route that returns the exact current-profile `profile_id` set, exact `claimed` plus `route_families[]` item shapes, canonical ordering, and `invalid_pagination_request` on pagination members.                                                           | REQ-01-542..REQ-01-545, REQ-04-105             | AC-370                                         |
| U-2-10 | Reserved-extension dispatch uses the required precedence: base routes first, claimed extension families second, reserved-but-unclaimed families return `extension_profile_not_claimed` before family-specific authorization or policy evaluation, and ordinary unknown-route handling applies only outside reserved families. | REQ-01-546..REQ-01-548                         | AC-371                                         |
| U-2-11 | The ordinary landing shell creates an incident, refreshes session-visible membership, routes to the workbook by `incident_id`, and falls back cleanly when a stale incident selection is no longer visible.                                                                                                               | REQ-01-154..REQ-01-160, REQ-04-021..REQ-04-026 | AC-170..AC-174                                 |
| U-2-12 | The ordinary incident shell gates patchable incident metadata controls by incident role, exposes `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref` only to reviewer/admin users, hides membership-administration controls from non-admin members, and returns to landing when incident access is lost. | REQ-01-127..REQ-01-137, REQ-01-168..REQ-01-180, REQ-01-491.1, REQ-02-222..REQ-02-223, REQ-04-021..REQ-04-030 | AC-175..AC-180, AC-170..AC-174 |
| U-2-13 | The ordinary incident shell sends membership create or patch or delete requests with the expected payloads and refreshes the session-visible incident role after each mutation.                                                                                                                                           | REQ-01-127..REQ-01-137, REQ-04-021..REQ-04-026 | AC-175..AC-180                                 |
| U-2-14 | Store-backed incident patch rejects stale `base_incident_version` values with typed conflict details and does not mutate durable incident state.                                                                                                                                                                      | REQ-01-168..REQ-01-180, REQ-02-015             | AC-170..AC-174, AC-211..AC-214                 |

### 4.2.4 Integration tests

| ID     | Test                                                                                                                                                                                                                                                               | Exact REQs                                     | Exact ACs                      |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------- | ------------------------------ |
| I-2-01 | Incident create persists the incident row, creator bootstrap membership, both workbook-preference objects, and owner-level mutation artifacts on the current audit substrate atomically.                                                                         | REQ-01-154..REQ-01-160, REQ-01-145..REQ-01-151 | AC-170..AC-174, AC-150, AC-153 |
| I-2-02 | Real incident create with trimmed decomposed input normalizes the response key, replay with the composed form returns the original committed resource, keeps the incident row, bootstrap membership, workbook preferences, owner mutation artifacts, and route idempotency stable, and opposite-form duplicate `incident_key` on a distinct request fails with `incident_key_conflict` after normalization. | REQ-01-161..REQ-01-165, REQ-02-016             | AC-170..AC-174, AC-211..AC-214 |
| I-2-03 | Membership changes re-derive incident authorization immediately for every `ControlBoundaryInventory()` route, including workbook-preference reads and writes, membership-admin routes, and the Timeline websocket boundary. Deployment-admin-without-membership breadth is paired with the existing support inventory sweep.          | REQ-01-127..REQ-01-137, REQ-04-021..REQ-04-030 | AC-175..AC-180                 |
| I-2-04 | Incident patch persists only the allowed patchable metadata fields, including set, replace, `null` clear, normalized-empty text clear, and no-op cases for `description` and `severity`; rejects noncanonical or empty TLP and stale versions; and advances `incident_version` only on material change. | REQ-01-168..REQ-01-180, REQ-01-491.1, REQ-02-015, REQ-02-222..REQ-02-223 | AC-170..AC-174, AC-211..AC-214 |
| I-2-05 | `GET /api/v1/extensions` succeeds for an authenticated session with zero incident memberships, returns only extension-claim state, and never exposes provider secrets, provider claim maps, or live extension-family payload.                                      | REQ-01-542..REQ-01-546, REQ-04-105             | AC-370                         |
| I-2-06 | When a profile is unclaimed, both the reserved family root and a descendant route under that family return `404 error.code='extension_profile_not_claimed'` with canonical `profile_id` and `route_family`; paths outside reserved families do not use that error. | REQ-01-544..REQ-01-548                         | AC-371                         |
| I-2-07 | Live membership patch to the current role returns `200 OK`, keeps `membership_version` and durable role stable, and writes no membership mutation artifact.                                                                                                      | REQ-01-133..REQ-01-137                         | AC-179                         |

### 4.2.5 E2E tests

| ID     | Test                                                                                                                                                                                | Exact REQs                                     | Exact ACs                      |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------ |
| E-2-01 | An authenticated user creates an incident, is bootstrapped as incident `admin`, and lands on the ordinary workbook shell with visible workbook-preference bootstrap state.          | REQ-01-154..REQ-01-160, REQ-04-021..REQ-04-026 | AC-170..AC-174                 |
| E-2-02 | The same incident appears in incident discovery, is retrieved into the ordinary incident shell through a raw `/?incident_id=...` deep link, and patches only the allowed incident metadata fields through visible default-path controls. | REQ-01-168..REQ-01-180, REQ-01-491.1, REQ-02-015, REQ-02-222..REQ-02-223 | AC-170..AC-174, AC-211..AC-214 |
| E-2-03 | An incident admin adds, changes, and removes memberships on the ordinary shell. A non-admin incident member sees the same membership state but not the admin controls.              | REQ-01-127..REQ-01-137, REQ-04-021..REQ-04-030 | AC-175..AC-180                 |

Phase 2 browser helpers that open incident controls must wait on semantic controls-readiness state rather than mutation-result copy. This is frontend implementation-support guidance only and does not add Core incident-resource behavior beyond the route and authorization contracts cited above.

Phase 2 list-search implementation addendum: when `GET /api/v1/incidents` implementation work lands, promote manifest rows for incident-list search/filter validation, cursor query mismatch, hidden-resource pagination, the `limit=1` later-page match fixture, and incident-directory frontend stale-response handling. Those rows should map to REQ-01-168, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-581..REQ-01-584, AC-415, and AC-416.

Phase 2 still keeps browser-authenticated request probes for route-owned validation errors, extension singleton discovery semantics, and reserved-family precedence, but those belong to supplemental browser support coverage rather than the authoritative E2E completion map. The debug-only `Phase2Harness` remains a probe-only support surface rather than completion evidence, while process smoke checks are supplemental Go rows selected through the `backend-process` execution family.

---

## 5. Phase 3 — Timeline, grid hot path, and first record-row mutation substrate

### 5.3.1 Scope

This phase introduces the first high-volume workbook record path and therefore the first complete record-row mutation substrate:

- Timeline row create and patch,
- partial and uncertain capture,
- autosave and exact save-state labels,
- first record-row `change_set` / revision emission,
- transactionally maintained timeline projection rows,
- explicit `mark-reviewed` and `supersede` actions,
- current Timeline `capture_state` machine: `rough`, `enriched`, `reviewed`, `superseded`.

This is the first phase where record-row history exists at all. Phase 7 will later complete the reviewer-facing history, delete or restore, and rollback surface.

### 5.3.2 Primary owner sections

- Core 01 §3.3.5 mutation contract and Timeline review actions
- Core 01 §7.4 Timeline view contract
- Core 01 §8 projection model
- Core 02 §5 partial and uncertain data
- Core 02 §14 history substrate minima
- Core 03 §1 interaction model
- Core 03 §4 save-state and autosave behavior
- Core 03 §6 Timeline lifecycle
- Core 03 §7 Timeline create workflow
- Core 03 §15 Timeline read and write contract
- Core 04 AC-131, AC-136, and AC-162 canonical incident WebSocket behavior

### 5.3.3 Unit tests

| ID          | Test                                                                                                                                                                                                                                   | Exact REQs                                                                                     | Exact ACs                              |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------- |
| U-3-01      | A Timeline row can be created with one non-empty user-entered value. System-managed fields, `record_id`, and `row_version` are assigned after commit.                                                                                  | REQ-01-057, REQ-03-111..REQ-03-115                                                             | AC-001, AC-002, AC-125                 |
| U-3-02      | Initial Timeline create sets `capture_state='rough'` on the committed row and projection state.                                                                                                                                            | REQ-03-102, REQ-03-236..REQ-03-241                                                             | AC-107..AC-111, AC-191                 |
| U-3-03      | The first capture-state-material mutation transitions `rough -> enriched`, and a store-driven `mark-reviewed` persists the explicit `reviewed` transition on real store state.                                                            | REQ-03-103..REQ-03-104, REQ-01-083..REQ-01-085                                                 | AC-107..AC-111, AC-194..AC-197         |
| U-3-04      | A later material edit on a reviewed Timeline row demotes it to `enriched`. An explicit supersede action moves a legal row to `superseded` and blocks ordinary forward editing semantics.                                               | REQ-03-105..REQ-03-110, REQ-01-086..REQ-01-088                                                 | AC-107..AC-111, AC-198..AC-199         |
| U-3-05      | Autosave commits on Enter, Tab, blur, and paste completion. No explicit Save button is required. Save-state labels are exactly `Syncing`, `Saved`, and `Conflict`. Same-record dependent actions wait for pending autosave replay to accept the latest committed row version before dispatch. | REQ-03-035, REQ-03-087..REQ-03-089                                                            | AC-043                                 |
| U-3-06      | `PATCH /api/v1/records/{record_id}` requires `view_schema_id`, `base_row_version`, `client_txn_id`, and non-empty `changes[]`. Duplicate `field_key` entries or `changes[]: []` are rejected as malformed mutation payloads.           | REQ-01-058..REQ-01-060, REQ-01-069..REQ-01-070                                                 | AC-125, AC-299                         |
| U-3-07      | Patch-route idempotent replay returns the original committed result and creates no second mutation row and no second record revision on the write substrate. Collaboration suppression remains integration or browser evidence.        | REQ-01-058, REQ-01-069..REQ-01-070                                                             | AC-299                                 |
| U-3-08      | Timeline projection rows carry the view-owned derived fields and maintain stable `record_id` / `row_version` binding for the grid.                                                                                                     | REQ-01-312..REQ-01-322, REQ-01-349..REQ-01-350, REQ-03-236..REQ-03-241                         | AC-116, AC-119, AC-120, AC-191..AC-193 |
| U-3-09      | Every successful Timeline create or patch writes an attributed mutation entry and a new row revision for that record.                                                                                                                  | REQ-02-205..REQ-02-207, REQ-04-036..REQ-04-037                                                 | AC-215, AC-231                         |
| U-3-10      | Supersede-with-replacement rejects illegal replacement targets, replays idempotently by `(record_id, client_txn_id)`, writes one coupled `timeline_record` plus one `record_link` mutation in a single change set, and rolls both back together. | REQ-01-086..REQ-01-087, REQ-01-311..REQ-01-312, REQ-02-168, REQ-02-181, REQ-03-106..REQ-03-107 | AC-329..AC-331                         |
| U-3-11      | Stale Timeline patches derive committed writable field changes since the client base row version. Different-field edits rebase, same-field direct edits return `same_field_conflict`, collection fields use `collection_value_v1`, lifecycle-only stale edits apply against current state, and replay precedence remains idempotent. | REQ-03-033..REQ-03-040, REQ-03-063..REQ-03-068, REQ-01-062..REQ-01-067                         | AC-009, AC-126, AC-203                 |
| U-3-12      | The workbook create payload builder emits a zero-field Timeline create only for explicit blank-row creation and keeps ordinary draft autosave from submitting empty creates.                                                            | REQ-01-057, REQ-03-111..REQ-03-115                                                             | AC-191                                 |
| U-3-13      | The workbook explicit blank-row action sends a `client_txn_id`-only create once during a duplicate pending-submit burst and renders the committed rough row plus a fresh draft row.                                                     | REQ-01-057, REQ-03-111..REQ-03-115                                                             | AC-191                                 |
| U-3-14      | Closed incidents remain readable and permit exact committed idempotent replay, while divergent replay conflicts before lifecycle checks and fresh Timeline source-state writes fail without new durable side effects.                    | REQ-01-590..REQ-01-591, REQ-01-057..REQ-01-070                                                  | AC-422, AC-423                         |
| U-3-15      | Timeline query accepts `timeline.evidence_count` as a declared sortable hidden field and returns projection rows in ascending or descending evidence-count order.                                               | REQ-01-312, REQ-01-322                                                                          | AC-116, AC-117, AC-118, AC-362, AC-365 |
| U-3-16      | Timeline mutation error classification emits stable API envelopes for idempotency, closed-incident, not-found, deleted-record, row-version, same-field, illegal-transition, and no-effective-change failures.     | REQ-01-057..REQ-01-070, REQ-01-086..REQ-01-088, REQ-01-590..REQ-01-591, REQ-03-033..REQ-03-040 | AC-125, AC-126, AC-203, AC-299, AC-329..AC-331, AC-422..AC-423 |
| U-3-GRID-01 | Timeline grid column bindings are produced from active `view_schema` fields and every writable cell commit uses `field_key`, not visible column label.                                                                                 | REQ-01-022, REQ-01-034..REQ-01-036, REQ-03-033..REQ-03-035                                     | AC-124, AC-125, AC-127, AC-184, AC-231 |
| U-3-GRID-02 | Timeline grid row bindings carry `record_id` and `row_version` and do not use visible row index as mutation identity.                                                                                                                  | REQ-01-015..REQ-01-017, REQ-01-022, REQ-03-033..REQ-03-035                                     | AC-001, AC-013, AC-047, AC-231         |
| U-3-GRID-03 | A local edit in a sorted or filtered Timeline viewport emits a patch with the original bound `record_id`, `base_row_version`, and `field_key`.                                                                                         | REQ-01-022, REQ-01-035, REQ-01-057..REQ-01-060, REQ-03-033..REQ-03-035                         | AC-013, AC-125, AC-231                 |

The Phase 3 workbook currently renders through the RDG-backed `@cartulary/grid-adapter`. `U-3-GRID-01/02/03` therefore own workbook binding behavior on the live adapter path, while vendor-specific RDG semantics stay with adapter-level tests.

### 5.3.4 Integration tests

| ID     | Test                                                                                                                                          | Exact REQs                                                             | Exact ACs                      |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- | ------------------------------ |
| I-3-01 | Timeline create or patch writes the source row, mutation history rows, and projection row atomically in one transaction.                      | REQ-01-057..REQ-01-070, REQ-01-351..REQ-01-353, REQ-02-205..REQ-02-207 | AC-125, AC-210, AC-215, AC-299 |
| I-3-02 | The Timeline query route reads from projection rows and returns stable row identity and deterministic ordering without scanning source blobs. | REQ-01-034..REQ-01-037, REQ-01-355..REQ-01-366, REQ-03-236..REQ-03-241 | AC-124, AC-184, AC-191..AC-193 |
| I-3-03 | Incident-role authorization, review and supersede lifecycle transitions, replay semantics, and replacement-target legality hold on the live route surface, including the coupled supersede change-set shape. | REQ-01-083..REQ-01-088, REQ-03-104..REQ-03-110                         | AC-107..AC-111, AC-194..AC-199 |
| I-3-04 | Same-field Timeline patch conflicts are transported as `409 same_field_conflict` with `error.conflict` and no mutation or idempotency write.  | REQ-03-063..REQ-03-068                                                 | AC-126                         |
| I-3-05 | The canonical incident WebSocket route proves handshake, presence snapshot ordering, membership-gated acceptance, browser-origin rejection, and incident-access revocation while preserving access to other authorized incidents. | REQ-01-250..REQ-01-277, REQ-04-005..REQ-04-017                         | AC-131, AC-136, AC-162         |
| I-3-06 | Phase 3 create, patch, mark-reviewed, supersede, query, and WebSocket pre-upgrade malformed or invalid requests return common error envelopes with route-specific details and malformed mutation requests create no durable writes. | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-233, REQ-01-250..REQ-01-277 | AC-231, AC-329                 |
| I-3-07 | Rough and uncertain Timeline capture preserves null structured fields, raw details, source text, and unresolved mention text through later explicit resolution. | REQ-02-024..REQ-02-025                                                 | AC-406                         |

### 5.3.5 E2E tests

| ID     | Test                                                                                                                                                                     | Exact REQs                                                             | Exact ACs      |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------- | -------------- |
| E-3-01 | An analyst creates a Timeline row with one non-empty value and immediately continues editing in-grid without leaving the workbook surface.                               | REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115                         | AC-001, AC-002 |
| E-3-02 | Typing acknowledgement and visible save-state transitions stay inside the declared hot-path interaction envelope on the reference fixture.                               | REQ-01-015..REQ-01-017, REQ-03-087..REQ-03-089, REQ-03-217..REQ-03-219 | AC-043         |
| E-3-03 | A reviewer-session browser flow marks a row as reviewed, later edits demote it to enriched when material, and a legal supersede action moves it to `superseded` using the post-demotion committed row version. | REQ-01-083..REQ-01-088, REQ-03-035, REQ-03-103..REQ-03-110              | AC-107..AC-111 |
| E-3-04 | Replaying the same patch request with the same `client_txn_id` returns the original committed result and does not create duplicate history or duplicate visible updates. | REQ-01-058, REQ-01-069..REQ-01-070                                     | AC-299         |

Browser tests bound to Core 05 measurement predicates or p95 fixture-sensitive envelope checks are isolated-run evidence. They must execute through the measurement browser suite and not inside the parallel heavy verification block used for functional gate work.

### 5.3.6 Visual regression tests

| ID          | Test                                                                                                                                          | Exact REQs                                     | Exact ACs              |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------- |
| V-3-GRID-01 | Timeline default viewport captures stable row identity, visible `row_version`, and save-state strip without generated incident ID leakage.    | REQ-01-015..REQ-01-017, REQ-01-022, REQ-03-087 | AC-001, AC-043, AC-231 |
| V-3-GRID-02 | Timeline edit-state visuals cover active editable cell plus `Syncing`, `Saved`, and `Conflict` save-state presentations on the browser surface. | REQ-03-033..REQ-03-040, REQ-03-087..REQ-03-089 | AC-043, AC-126         |
| V-3-GRID-03 | Timeline grouped rows and currently exposed grid chrome render deterministically through the owned visual harness.                            | REQ-01-022, REQ-03-236..REQ-03-241             | AC-124, AC-184, AC-231 |

Phase 3 visual rows intentionally do not claim row-gutter collaboration markers, drag-fill, frozen-column, resize-handle, treegrid, or full conflict-resolution fixtures until those controls are product-exposed and owned by their later phase manifests.

---

## 6. Phase 4 — Entities, mentions, resolution, merge, and canonical-indicator foundations

### 6.4.1 Scope

This phase introduces progressive normalization, exact-match entity behavior, and the remaining Phase 4 base-profile HTTP/workbook evidence:

- `mention_origin` vs `entity_origin`,
- entity-mention lifecycle (`unresolved`, `resolved`, `dismissed`),
- explicit resolve, dismiss, and ordinary restore,
- explicit create-from-mention,
- exact-match reuse precedence,
- no auto-merge of pre-existing entities,
- explicit merge,
- source-bound indicator observations vs canonical indicators,
- Parties and coordination workbook surfaces,
- object-blob creation and evidence handle issuance.

### 6.4.2 Primary owner sections

- Core 01 §3.3.5 mention action route and merge route
- Core 02 §6 binding-mode contract
- Core 02 §7 provenance requirements
- Core 02 §8 deduplication and exact-match reuse
- Core 02 §9 merge behavior
- Core 02 §10 indicator model foundations
- Core 02 §13 evidence and object metadata
- Core 02 §19 incident-scoped party model
- Core 03 §8.4 evidence access
- Core 03 §9 resolution workflows and auto-resolution boundaries
- Core 03 §16 inspector and entity/evidence interaction
- Core 03 §20 Parties system-view and linking flows

### 6.4.3 Unit tests

Execution dependency is distinct from evidence layer in this phase. `U-4-01..U-4-07` are service-backed store-domain tests that execute through `make backend-store`; `U-4-08..U-4-09` remain pure decoder tests that execute through `make backend-unit`; `U-4-WB-*` are app-level workbook component tests that execute through `make frontend-unit`. Shared package tests remain support/tooling evidence unless an app-level row consumes them through `apps/web`.

| ID     | Test                                                                                                                                                                                                                                                 | Exact REQs                                                                                                                                                 | Exact ACs                      |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| U-4-01 | A `mention_origin` write creates `entity_mention` rows and never implicitly creates a host or identity record. An `entity_origin` write creates or upserts an entity row and does not synthesize mentions.                                           | REQ-02-028..REQ-02-036                                                                                                                                     | AC-019, AC-020, AC-022         |
| U-4-02 | Repeated identical mention text values remain separate mention rows with distinct source provenance.                                                                                                                                                 | REQ-02-031..REQ-02-032, REQ-02-058                                                                                                                         | AC-019, AC-021                 |
| U-4-03 | Explicit `Create host` or `Create identity` from a mention creates a stub only when no unique exact-match entity exists, keeps the raw mention, and resolves only the selected mention by default.                                                   | REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055                                                                                                             | AC-020, AC-021, AC-186         |
| U-4-04 | Dismissing a mention preserves the mention row, clears active resolution metadata, and excludes it from active relationship-cell values. Ordinary restore returns it to `unresolved`, not to a historical resolved target.                           | REQ-02-039..REQ-02-041                                                                                                                                     | AC-188..AC-190, AC-224, AC-225 |
| U-4-05 | Exact-match reuse follows the stable precedence rules for hosts and identities. Alias and fuzzy matches remain suggestions only and never auto-resolve or auto-merge.                                                                                | REQ-02-059..REQ-02-063                                                                                                                                     | AC-021, AC-022                 |
| U-4-06 | Entity merge is explicit only. The survivor `record_id` remains unchanged, the loser remains historical with merge lineage, and raw source mentions are not rewritten.                                                                               | REQ-02-064..REQ-02-066                                                                                                                                     | AC-023, AC-186, AC-209         |
| U-4-07 | Indicator observations remain source-bound rows. Canonical indicators use incident-scoped dedupe identity and lifecycle state that is separate from observations.                                                                                    | REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082                                                                                                 | AC-017, AC-077..AC-079         |
| U-4-08 | Timeline request decoding preserves authoritative raw mention text, applies mention-token normalization for exact-match eligibility, and accepts suppressed or forbidden rewrite tokens as valid submitted values without implicitly rewriting them. | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 | AC-205, AC-388..AC-392         |
| U-4-09 | Timeline create and patch request decoding accepts manual relationship actions only when provenance or confidence metadata is omitted and fails closed on client-supplied `confidence`, `provenance`, or routing overrides.                          | REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280                                                                                                 | AC-394, AC-396, AC-397         |
| U-4-WB-01 | The assessment workbook payload builder maps band-first UI state to confidence score create fields, deduplicates support references, and does not submit derived confidence-band fields. | REQ-03-250..REQ-03-254 | AC-018, AC-080..AC-084 |
| U-4-WB-02 | The app-level assessment workbook UI submits Phase 4 create payloads through ordinary controls and renders the returned assessment row. | REQ-03-250..REQ-03-254 | AC-018, AC-080..AC-084 |
| U-4-WB-03 | The Evidence workbook surface issues preview and download handle requests with the required empty-object body and consumes opaque returned hrefs instead of deriving storage access. | REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-459, REQ-01-465 | AC-251 |
| U-4-WB-04 | The generic workbook mutation builder emits required Phase 4 create payloads with trimmed direct values, timestamp fields, and explicit null clears. | REQ-01-329..REQ-01-340, REQ-03-255..REQ-03-260 | AC-085, AC-086, AC-137..AC-145 |
| U-4-WB-05 | The generic workbook mutation builder maps direct clears plus token, party, record, and risk collection edits to typed Phase 4 collection action payloads. | REQ-01-329..REQ-01-340, REQ-03-255..REQ-03-260 | AC-085, AC-086, AC-137..AC-145 |

### 6.4.4 Integration tests

| ID     | Test                                                                                                                                                                                                                                                                                                                                                                                      | Exact REQs                                                                                                                                                 | Exact ACs                      |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| I-4-01 | `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` updates durable mention state, source-row invalidation, and active links for `resolve_item`, `dismiss_item`, and `revert_to_unresolved`, and it enforces route-owned row-version checks, idempotent replay, divergent replay rejection, live authorization re-check, and target validation.                                    | REQ-01-196..REQ-01-227, REQ-02-039..REQ-02-044                                                                                                             | AC-188..AC-190, AC-221..AC-225 |
| I-4-02 | Direct create on Hosts or Identities reuses a unique exact match, reuses preserved identifiers after canonical fields drift, keeps suggestion-only aliases as non-authoritative hints, returns merge-required conflicts for ambiguous exact matches, and never synthesizes `entity_mention` rows.                                                                                         | REQ-02-035..REQ-02-036, REQ-02-054..REQ-02-055, REQ-02-059..REQ-02-063                                                                                     | AC-022, AC-186                 |
| I-4-03 | Entity merge repoints live resolutions and live links to the survivor, preserves loser lineage and history, carries forward exact-match reuse, enforces route-owned replay semantics, re-checks current authorization, and emits the Timeline websocket invalidation boundary for dependent rows.                                                                                         | REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066                                                                                                             | AC-023, AC-186, AC-209         |
| I-4-07 | The Indicators create and query routes persist canonical indicators through the real HTTP surface, re-derive live authorization at the route boundary, emit attributed mutation history, and expose observation or lifecycle projection consequences through current-state reads without collapsing source-bound observations into one row.                                               | REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082                                                                                                 | AC-017, AC-077..AC-079         |
| I-4-08 | Interactive auto-resolution uses only the explicit exact-match eligibility contract, creates `provenance='auto_match'` links with `confidence=100` only for eligible committed Timeline tokens, never strips hedge or punctuation forms into matches, never backfills matches later through projection rebuild or background cleanup, and rolls back cleanly when the owning patch fails. | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 | AC-205, AC-388..AC-392         |
| I-4-09 | Manual Timeline relationship mutations on `timeline.host_refs` or `timeline.identity_refs` commit with `provenance='manual'`, `confidence=null`, reject client-supplied `confidence`, and preserve authoritative `null` through current-state API reads on both the real create and patch route boundaries.                                                                               | REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280                                                                                                 | AC-394, AC-396, AC-397         |
| I-4-COLLECTION-01 | Generic record patch collection fields use stable `record_id`, `field_key`, `row_version`, and `base_row_version` anchors with `collection_actions_v1` writes and `collection_value_v1` reads or conflict payloads for add, remove, and replace behavior, while rejecting legacy array-shaped writes. | REQ-01-062..REQ-01-067, REQ-03-052..REQ-03-053, REQ-03-259, REQ-03-265 | AC-201, AC-202 |
| I-4-BLOB-01 | `POST /api/v1/object-blobs` enforces the object-blob request contract, re-derives authorization, normalizes accepted optional fields, rejects oversize declared byte counts before slot creation, and preserves idempotent replay or divergent replay semantics without rejected-request side effects. | REQ-01-243..REQ-01-247, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128, REQ-04-048 | AC-128 |
| I-4-COORD-01 | Communications Log, Handoff, Status Review, and Lesson creates enforce minimum semantic fields, default server-owned values and empty collections, keep server ids immutable, distinguish clearable and non-clearable scalar fields, and include authoritative replay/history conformance without partial state on rejection. | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-503..REQ-01-506, REQ-02-123..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 | AC-281..AC-284 |
| I-4-COORD-02 | Coordination collection fields use typed `collection_value_v1` record, party, and risk references, mutate through collection actions, coalesce duplicates, reject raw arrays/nulls and invalid targets, and return typed same-field conflict payloads. | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-503..REQ-01-506, REQ-02-123..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 | AC-281..AC-284 |
| I-4-COORD-03 | Parties and coordination system views query route-created Parties and coordination rows plus seeded support rows through contract-declared filters, sorts, grouping values, row ids, and cell values. | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-503..REQ-01-506, REQ-02-123..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 | AC-277, AC-281..AC-284 |
| I-4-HANDLE-01 | Preview and download handle issuance accepts `{}` only, rejects zero-length, `null`, non-object, unknown, and `client_txn_id` bodies on both routes without side effects, re-derives authorization, and returns distinct preview versus download contracts. | REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-459, REQ-01-465 | AC-251 |
| I-4-PARTIES-01 | The Parties system view is discoverable in the base-profile schema registry, fetchable by singleton route, incident-scoped in query behavior, and enforces create-time display-name and party-kind requirements with authoritative replay/history conformance and no partial state on rejection. | REQ-01-296, REQ-01-343, REQ-01-497..REQ-01-501, REQ-02-003, REQ-02-006, REQ-02-009, REQ-02-022, REQ-02-202, REQ-02-222..REQ-02-225, REQ-03-005, REQ-03-266 | AC-277 |

### 6.4.5 E2E tests

| ID     | Test                                                                                                                                                                                                  | Exact REQs                                                                         | Exact ACs                      |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | ------------------------------ |
| E-4-01 | An analyst types raw host and identity text on a Timeline row, resolves one token to an existing entity, and creates a stub from another token in the inspector without leaving the workbook surface. | REQ-02-030..REQ-02-034, REQ-03-129..REQ-03-134, REQ-03-247..REQ-03-249             | AC-006, AC-019, AC-020         |
| E-4-02 | An analyst dismisses and later ordinarily restores a mention. The restored mention returns to the unresolved queue and does not silently recover an old resolved target.                              | REQ-02-039..REQ-02-041, REQ-03-129..REQ-03-134                                     | AC-188..AC-190, AC-224, AC-225 |
| E-4-03 | A reviewer merges two duplicate entities from the inspector. The surviving row identity remains stable and dependent links or resolutions follow the survivor.                                        | REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066, REQ-03-247..REQ-03-249             | AC-023, AC-186, AC-209         |
| E-4-04 | An eligible exact-match Timeline token auto-resolves on commit, while hedge, punctuation, parenthetical, or approximate forms remain unresolved and require explicit analyst action.                  | REQ-01-315..REQ-01-316, REQ-01-568, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 | AC-205, AC-388..AC-391         |
| E-4-05 | An analyst drives the Assessments workbook surface through ordinary controls with band-first confidence, rationale, support references, default ordering, and state or band filtering; durable append-only storage and history remain backend evidence. | REQ-03-250..REQ-03-254                                                              | AC-018, AC-080..AC-084         |
| E-4-06 | An analyst exercises the exposed generic workbook surfaces through typed controls, including defaults, collections, and visible validation failure; replay, authorization, durable history, and later queue-heavy workflow expansion remain backend or Phase 9 evidence. | REQ-01-329..REQ-01-340, REQ-03-255..REQ-03-260                                     | AC-085, AC-086, AC-137..AC-145 |

### 6.4.6 Visual regression tests

| ID          | Test                                                                                                                                     | Exact REQs                                                             | Exact ACs              |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------- |
| V-4-GRID-01 | Timeline visual evidence captures unresolved mention chips and resolved entity chips through the live workbook grid.                    | REQ-02-030..REQ-02-034, REQ-03-129..REQ-03-134                         | AC-006, AC-019, AC-020 |
| V-4-GRID-02 | Evidence visual evidence captures required access affordances and blocked-access messaging through the ordinary Evidence workbook grid. | REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-459, REQ-01-465             | AC-251                 |
| V-4-GRID-03 | Coordination visual evidence captures a required system view, such as Task Requests or Decisions, through the generic workbook grid.     | REQ-01-329..REQ-01-340, REQ-03-255..REQ-03-260                         | AC-085, AC-086         |

---

## 7. Phase 5 — Evidence, blob lifecycle, evidence access, and object storage

### 7.5.1 Scope

This phase completes the binary-evidence path:

- evidence records without blobs,
- blob-slot creation,
- accepted upload contract echo,
- final attach or replacement on an evidence record,
- evidence lifecycle vs blob lifecycle separation,
- preview-handle and download-handle issuance,
- same-origin handle redemption with current authorization re-check,
- fail-closed preview and download behavior,
- evidence projection updates on workbook surfaces.

### 7.5.2 Primary owner sections

- Core 01 §3.3.8 blob-slot and evidence-attach routes
- Core 01 §16 evidence-access handle contract
- Core 02 §4.5 promoted evidence fields
- Core 02 §13 evidence and blob schema
- Core 03 §8 evidence attachment workflow
- Core 03 §16 evidence sheet and inspector behavior
- Core 04 §4.3 evidence upload trust boundary
- Core 04 §4.5 hostile upload and preview constraints

### 7.5.3 Unit tests

| ID     | Test                                                                                                                                                                                                                                       | Exact REQs                                                                         | Exact ACs                      |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------- | ------------------------------ |
| U-5-01 | `POST /api/v1/object-blobs` requires `incident_id`, `client_txn_id`, and declared `byte_size`, rejects unknown top-level members, and echoes `accepted_contract` on success.                                                               | REQ-01-243..REQ-01-247                                                             | AC-128, AC-154, AC-155         |
| U-5-02 | Blob-create idempotency is keyed by `(actor_user_id, incident_id, client_txn_id)`. Replay of the same normalized request returns the original slot. Divergent replay returns route-owned conflict.                                         | REQ-01-243..REQ-01-247                                                             | AC-128, AC-154, AC-155         |
| U-5-03 | `POST /api/v1/evidence-records/{record_id}/attach-blob` requires `object_blob_id`, `base_row_version`, and `client_txn_id` and fails closed when the blob is pending, failed, missing, expired, or contract-mismatched.                    | REQ-01-245..REQ-01-247                                                             | AC-102, AC-103, AC-128         |
| U-5-04 | Evidence lifecycle state remains separate from blob `upload_state`. Requested or pending-receipt evidence can exist without a blob and later advance without mutating unrelated custody history.                                           | REQ-02-186..REQ-02-201                                                             | AC-102, AC-103, AC-154, AC-155 |
| U-5-05 | Preview-handle and download-handle issuance routes accept only `{}` and are intentionally non-idempotent. Each success yields a fresh opaque same-origin handle.                                                                           | REQ-01-458..REQ-01-463                                                             | AC-251, AC-252, AC-253         |
| U-5-06 | Handle redemption re-derives current session validity, current incident membership, and current evidence/blob state. Preview blocks explicitly on unsupported or unsafe preview conditions rather than silently falling back.              | REQ-01-463..REQ-01-465, REQ-04-023, REQ-04-053                                     | AC-252..AC-255                 |
| U-5-07 | Download responses use authoritative metadata for filename disposition and apply deterministic fallback naming when authoritative names are absent.                                                                                        | REQ-01-459..REQ-01-465                                                             | AC-251, AC-253                 |
| U-5-08 | Evidence attachment updates workbook-visible evidence counts and `has_evidence`-style derived flags without forcing navigation away from the current surface.                                                                              | REQ-03-116..REQ-03-126, REQ-03-242..REQ-03-246                                     | AC-004, AC-015, AC-016         |
| U-5-09 | Blob-create size ceilings reject oversize `byte_size` before slot creation, and preview-handle issuance fails with the route-owned oversized-preview reason while leaving download behavior available when the payload is otherwise legal. | REQ-01-238, REQ-01-243..REQ-01-245, REQ-01-461, REQ-01-465, REQ-04-079..REQ-04-080 | AC-321, AC-322                 |
| U-5-10 | Evidence create, attach, preview issuance, and redemption use closed public error and reason-code registries. Attach failures use `evidence_attach_rejected` without legacy unregistered reasons.                                         | REQ-01-245..REQ-01-247, REQ-01-458..REQ-01-465                                     | AC-128, AC-252..AC-255         |
| U-5-11 | Evidence and Timeline evidence-derived field keys are closed over the derived view-schema registry and reject label, storage-name, or non-canonical key fallbacks.                                                                         | REQ-03-116..REQ-03-126, REQ-03-242..REQ-03-246                                     | AC-015, AC-016                 |

### 7.5.4 Integration tests

| ID     | Test                                                                                                                                                                                  | Exact REQs                                                             | Exact ACs                              |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- | -------------------------------------- |
| I-5-01 | A real object-store upload can be created through a blob slot, finalized onto an evidence record, and then queried through the workbook projection without duplicate attachment rows. | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-126 | AC-015, AC-016, AC-128, AC-154, AC-155 |
| I-5-02 | Expired-slot replay returns the same expired slot. A fresh upload target requires a new `client_txn_id`.                                                                              | REQ-01-243..REQ-01-247                                                 | AC-128, AC-154, AC-155                 |
| I-5-03 | A redeemed preview or download handle fails closed after logout, membership removal, or evidence/blob state invalidation.                                                             | REQ-01-463..REQ-01-465, REQ-04-023, REQ-04-053                         | AC-252..AC-255                         |
| I-5-04 | If the deployment uses an upload-scanning adjunct service, the two-step attach semantics remain intact and preview never bypasses the scanning or quarantine boundary.                | REQ-04-048, REQ-04-053                                                 | AC-053                                 |
| I-5-05 | Attach route validation, authorization, route-owned error envelopes, and divergent replay rejection occur through the real HTTP surface before object observation mutates state.       | REQ-01-245..REQ-01-247, REQ-04-023                                     | AC-102, AC-103, AC-128                 |
| I-5-06 | Attached-evidence projection storage can be corrupted and deterministically rebuilt from source without mutating source rows, links, or mutation history.                             | REQ-03-116..REQ-03-126, REQ-03-242..REQ-03-246                         | AC-015, AC-016                         |
| I-5-07 | A successful evidence attach publishes canonical `record_changed` messages for the Evidence row and affected Timeline evidence projection through the real incident WebSocket emitter. | REQ-03-116..REQ-03-126, REQ-03-242..REQ-03-246                         | AC-015, AC-016                         |

### 7.5.5 E2E tests

| ID     | Test                                                                                                                                                                      | Exact REQs                                     | Exact ACs              |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------- |
| E-5-01 | An analyst attaches a screenshot to a selected Timeline row without leaving the workbook surface. The row reflects the new evidence count after commit.                   | REQ-03-116..REQ-03-126                         | AC-004, AC-015, AC-016 |
| E-5-02 | A screenshot-only Timeline row can be persisted through the two-step evidence path.                                                                                       | REQ-03-102, REQ-03-116..REQ-03-126             | AC-002, AC-102, AC-103 |
| E-5-03 | An inline-safe type receives a preview handle and renders through the same-origin redeem path. An unsupported or unsafe type returns an explicit blocked-preview outcome. | REQ-01-458..REQ-01-465, REQ-04-053             | AC-252..AC-255         |
| E-5-04 | Requested evidence can be tracked before the blob exists and later advanced to an available state without breaking workbook pivots or counts.                             | REQ-02-186..REQ-02-201, REQ-03-242..REQ-03-246 | AC-015, AC-154, AC-155 |
| E-5-05 | A second live workbook session receives the real evidence attach WebSocket event and refreshes Timeline evidence count and flags without navigation.                     | REQ-03-116..REQ-03-126, REQ-03-242..REQ-03-246 | AC-015, AC-016         |

### 7.5.6 Visual regression tests

| ID          | Test                                                                                                                                | Exact REQs                                     | Exact ACs              |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------- |
| V-5-GRID-01 | Requested Evidence surface state and the same row after workbook-surface attach render deterministically through the visual harness. | REQ-02-186..REQ-02-201, REQ-03-242..REQ-03-246 | AC-015, AC-154, AC-155 |
| V-5-GRID-02 | Blocked preview feedback and Timeline evidence-count plus `has_evidence` badge presentation render through the visual harness.      | REQ-01-458..REQ-01-465, REQ-03-116..REQ-03-126 | AC-015, AC-016, AC-252 |

---

## 8. Phase 6 — Collaboration, presence, and same-field conflict resolution

### 8.6.1 Scope

This phase completes the live multi-user path:

- incident-scoped WebSocket route,
- `hello` and `resume`,
- replay window and reset behavior,
- presence snapshots and deltas,
- `record_changed`, `job_progress`, `session_revoked`, and heartbeat behavior,
- field-level optimistic concurrency,
- same-field conflict transport,
- `atomic_replace`, `text_compare_merge`, and `collection_review`,
- same-surface resolver behavior and client-local conflict queue handling.

### 8.6.2 Primary owner sections

- Core 01 §3.3.10 WebSocket public contract
- Core 01 §3.3.5 mutation contract
- Core 03 §3 concurrency and same-field conflict resolution
- Core 03 §4 save-state, presence, pending queue, and local draft boundaries
- Core 04 §1 session lifecycle consequences
- Core 04 §2 authorization re-derivation
- Core 04 §4.5 origin and hostile-content constraints relevant to sockets

### 8.6.3 Unit tests

| ID     | Test                                                                                                                                                                                                                                                                                   | Exact REQs                                                 | Exact ACs                                      |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- | ---------------------------------------------- |
| U-6-01 | Every grid write is anchored to the route-bound `record_id` and includes body-bound `base_row_version` plus changed fields only. Different-field concurrent edits auto-rebase. Same-field concurrent edits reject with an explicit conflict payload.                                  | REQ-03-033..REQ-03-040                                     | AC-009, AC-013, AC-126                         |
| U-6-02 | Same-field conflict transport uses `409`, `error.code='same_field_conflict'`, and an `error.conflict` object with the required base and current values.                                                                                                                                | REQ-03-063..REQ-03-068                                     | AC-126, AC-203, AC-204, AC-226..AC-230         |
| U-6-03 | `text_compare_merge` treats the field as plain text, normalizes line endings only for merge computation, and never silently auto-commits a clean merge suggestion.                                                                                                                     | REQ-03-054..REQ-03-061                                     | AC-226..AC-230                                 |
| U-6-04 | `collection_review` fields use `collection_value_v1` in read or conflict payloads and `collection_actions_v1` in explicit resolution writes.                                                                                                                                           | REQ-01-062..REQ-01-067, REQ-03-052..REQ-03-053             | AC-118, AC-203, AC-204                         |
| U-6-05 | The resolver keeps the grid visible, leaves conflict state unresolved until explicit action, and returns focus to the same cell after resolution or clear.                                                                                                                             | REQ-03-041..REQ-03-051                                     | AC-037..AC-042                                 |
| U-6-06 | `keep_saved` clears the local conflict without creating a new revision. `use_unsaved` and `merged_value` create a new attributed change set.                                                                                                                                           | REQ-03-077..REQ-03-082                                     | AC-041, AC-163                                 |
| U-6-07 | The first application message on the incident socket is exactly one of `hello` or `resume`. Resume outside the replay window yields reset behavior rather than partial replay.                                                                                                         | REQ-01-250..REQ-01-277                                     | AC-129, AC-131, AC-135                         |
| U-6-08 | Presence payloads are incident-scoped and ephemeral. Heartbeats do not extend idle expiry. `session_revoked` closes the socket with the route-owned reason code set.                                                                                                                   | REQ-01-250..REQ-01-277, REQ-03-090..REQ-03-100, REQ-04-010 | AC-008, AC-131, AC-132..AC-136, AC-156..AC-163 |
| U-6-09 | Save-state labels use only `Syncing`, `Saved`, and `Conflict`; the local pending queue is FIFO, bounded to 64 non-coalescible units, same-record coalescing stays contiguous, and non-retryable failure or same-field conflict halts replay without silently evicting queued work inside one browser runtime. | REQ-03-072, REQ-03-089, REQ-03-095, REQ-03-099..REQ-03-100 | AC-376..AC-382                                 |

### 8.6.4 Integration tests

| ID     | Test                                                                                                                                                                                         | Exact REQs                                     | Exact ACs                                              |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------------------------------ |
| I-6-01 | Two real clients connect to the same incident socket, exchange presence snapshots and deltas, and observe deterministic replay ordering within the replay window.                            | REQ-01-250..REQ-01-277, REQ-03-090..REQ-03-098 | AC-129, AC-131..AC-135                                 |
| I-6-02 | Resume with a valid replay token replays replayable messages only. Presence is re-hydrated through the documented presence flow rather than replay.                                          | REQ-01-250..REQ-01-277                         | AC-129, AC-132..AC-135                                 |
| I-6-03 | Concurrent edits to different fields succeed without operator intervention. Concurrent edits to the same field produce the resolver path and preserve both saved and local drafts correctly. | REQ-03-033..REQ-03-082                         | AC-009, AC-037..AC-042, AC-203..AC-204, AC-226..AC-230 |
| I-6-04 | Cookie-authenticated browser socket upgrades reject untrusted `Origin` values before incident subscription is granted.                                                                       | REQ-01-250..REQ-01-277, REQ-04-053             | AC-131, AC-255                                         |

### 8.6.5 E2E tests

| ID     | Test                                                                                                                                                                                                                                                                                                                    | Exact REQs                                                                         | Exact ACs                              |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | -------------------------------------- |
| E-6-01 | Two analysts on the same incident see each other's presence on the workbook surface within the expected interaction window, measured only after both browser sockets have reached the accepted `hello_ack` state.                                                                                                      | REQ-03-090..REQ-03-098                                                             | AC-008, AC-132, AC-133                 |
| E-6-02 | Concurrent edits to different fields on the same row auto-merge. Concurrent edits to the same field open the same-surface resolver and require explicit analyst resolution.                                                                                                                                             | REQ-03-033..REQ-03-082                                                             | AC-009, AC-037..AC-042, AC-226..AC-230 |
| E-6-03 | Deterministic route-owned browser revocations such as logout, revoke-all, and expiry emit `session_revoked` on an established browser socket, close that socket, and preserve unsaved local work for later explicit recovery after re-authentication. Concurrency-limit victim selection and socket delivery remain backend and integration evidence. | REQ-01-029, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-013..REQ-04-016 | AC-131, AC-136, AC-156..AC-163         |
| E-6-04 | Live updates never retarget a pending local edit away from the bound `record_id` and `field_key` during sorting, filtering, grouping, virtual scrolling, or live row patch; same-field conflict markers, row-gutter presence markers, and same-cell editing hints where available remain attached to the intended cell. | REQ-01-015..REQ-01-017, REQ-03-086, REQ-03-090..REQ-03-098, REQ-03-223..REQ-03-235 | AC-008, AC-047, AC-132, AC-133         |
| E-6-05 | Within one browser runtime, queued unsent writes survive transient disconnect or session revocation, replay in FIFO order after re-authentication, halt on the first blocking non-retryable failure, and are never silently restored after a full reload.                                                               | REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100                                     | AC-377..AC-382                         |

### 8.6.6 Visual regression tests

| ID          | Test                                                                                                                          | Exact REQs                                                                         | Exact ACs                      |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | ------------------------------ |
| V-6-GRID-01 | Row-gutter presence markers and same-cell editing hints render deterministically on the Timeline grid from real browser presence state and record/field identity. | REQ-03-090..REQ-03-098                                                             | AC-008, AC-132, AC-133         |
| V-6-GRID-02 | Same-field conflict marker, same-surface resolver, and `Conflict` save-state strip render deterministically after a real backend same-field conflict; resolver behavior remains owned by `E-6-02` and unit rows. | REQ-03-033..REQ-03-082, REQ-03-099..REQ-03-100                                     | AC-037..AC-042, AC-376         |
| V-6-GRID-03 | Save-state presentations cover `Syncing`, same-field-conflict `Conflict`, and resolver recovery to `Saved` through the visual harness; queue semantics remain owned by `E-6-05` and `U-6-09`. | REQ-03-072, REQ-03-089, REQ-03-095, REQ-03-099..REQ-03-100                         | AC-376..AC-382                 |

---

## 9. Phase 7 — Reviewer-facing history, delete or restore, and rollback

### 9.7.1 Scope

This phase completes the reviewer and destructive-operation surface:

- record-history retrieval,
- history pagination,
- soft-delete and restore,
- history-entry rollback,
- whole-change-set rollback,
- merge-aware whole-change-set rollback,
- whole-row restore,
- tombstone `row_version` restore preconditions,
- `history_entry_ref`,
- `available_rollback_actions[]`,
- record-lock enforcement for destructive paths.

### 9.7.2 Primary owner sections

- Core 01 §3.3.4.2 record-history read contract
- Core 01 §3.3.5 delete, restore, review actions, and rollback routes
- Core 02 §14–§15 history and mutation-target substrate
- Core 03 §10 reviewer history and rollback workflows
- Core 04 §2 authorization model
- Core 04 §3 attribution and audit requirements

### 9.7.3 Unit tests

| ID     | Test                                                                                                                                                                                                                                                                                          | Exact REQs                                     | Exact ACs                      |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------ |
| U-7-01 | `GET /api/v1/records/{record_id}/history` returns newest-first items with deterministic order, current tombstone `row_version` for deleted rows, and canonical `available_rollback_actions[]` ordering.                                                                                       | REQ-01-048..REQ-01-056                         | AC-184, AC-185, AC-215         |
| U-7-02 | `history_entry_ref` is present only when a logical history item maps to exactly one reversible mutation target and remains opaque and stable across repeated reads.                                                                                                                           | REQ-01-054..REQ-01-055                         | AC-215, AC-216                 |
| U-7-03 | Delete requires the current `row_version`, respects role gates, and returns route-owned failures such as `record_deleted_use_restore`, `record_not_deleted`, and `record_locked` where applicable.                                                                                            | REQ-01-071..REQ-01-076, REQ-04-021..REQ-04-024 | AC-215, AC-218                 |
| U-7-04 | Restore requires the tombstone `row_version`, respects role gates, and returns the record to active state without mutating prior history rows in place.                                                                                                                                       | REQ-01-077..REQ-01-082                         | AC-215, AC-216                 |
| U-7-05 | Rollback accepts only the documented selector kinds `history_entry`, `change_set`, and `row_restore`, and creates a new `change_set` with source `rollback` rather than mutating prior history.                                                                                               | REQ-01-089..REQ-01-111                         | AC-216, AC-217, AC-412         |
| U-7-06 | Record locks are enforced for destructive or reviewer-only operations such as merge, rollback, and restore when the owner section requires them, fail fast before stale-precondition evaluation, and later replays fall through to ordinary downstream route errors after lock release.       | REQ-01-071..REQ-01-111, REQ-03-101             | AC-182, AC-187, AC-218, AC-353 |
| U-7-07 | Retained history for extant records remains fully paginatable and preserves stable `history_entry_ref` values across incident closure, delete or restore cycles, rollback, and restart; the current profile defines no history-purge route or retention-horizon setting for extant incidents. | REQ-01-054..REQ-01-056, REQ-01-561..REQ-01-563 | AC-383..AC-385                 |

### 9.7.4 Integration tests

| ID     | Test                                                                                                                                                                                                                                               | Exact REQs                                                             | Exact ACs              |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------- |
| I-7-01 | Delete, restore, and rollback update source rows, projections, history rows, and emitted collaboration events atomically.                                                                                                                          | REQ-01-071..REQ-01-111, REQ-01-351..REQ-01-353, REQ-02-205..REQ-02-220 | AC-210, AC-215..AC-218 |
| I-7-02 | History pagination remains bound to `record_id` and preserves deterministic item ordering across pages.                                                                                                                                            | REQ-01-056                                                             | AC-215                 |
| I-7-03 | A stale restore or rollback precondition fails closed and never mutates current row state.                                                                                                                                                         | REQ-01-077..REQ-01-082, REQ-01-089..REQ-01-111                         | AC-215, AC-218         |
| I-7-04 | History for an extant record remains fully paginatable and stable across service restart, incident closure, delete or restore, and rollback, with previously issued `history_entry_ref` values preserved for older single-entry-addressable items. | REQ-01-054..REQ-01-056, REQ-01-561..REQ-01-563                         | AC-383..AC-385         |
| I-7-05 | Merge-aware `change_set` rollback restores survivor and loser entity rows, carried aliases and preserved identifiers, repointed or deduped mentions, links, tags, assessments, projections, canonical affected records, append-only rollback history, idempotent replay, and stale base-version failure through the existing rollback route. | REQ-01-089..REQ-01-111, REQ-02-212..REQ-02-220                         | AC-412                 |
| I-7-06 | Indicator observation create and resolution targets and lifecycle-interval create targets appear in every affected row history, reverse through server-owned selectors without hard deletion, update active projections and every affected record envelope, append inverse history and revisions, and emit ordinary collaboration events. | REQ-01-089..REQ-01-111, REQ-02-260, REQ-03-294                         | AC-474                 |

### 9.7.5 E2E tests

| ID     | Test                                                                                                                                                  | Exact REQs                                     | Exact ACs              |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------- |
| E-7-01 | A reviewer opens row history from the workbook surface and sees actor, timestamp, operation, diff summary, and legal rollback actions.                | REQ-01-048..REQ-01-056, REQ-03-261..REQ-03-262 | AC-007, AC-215         |
| E-7-02 | A reviewer rolls back one mistaken link, tag, mention resolution, or evidence association without reverting later unrelated edits on the same row.    | REQ-01-089..REQ-01-111, REQ-02-212..REQ-02-220 | AC-010, AC-216, AC-217 |
| E-7-03 | A reviewer soft-deletes and restores a row using tombstone concurrency. Other clients observe `remove` on delete and `invalidate` on restore.         | REQ-01-071..REQ-01-082, REQ-01-250..REQ-01-277 | AC-011, AC-215, AC-218 |
| E-7-04 | Whole-row restore creates a new attributed revision and moves the visible row back to the selected historical snapshot without erasing prior history. | REQ-01-089..REQ-01-111                         | AC-011, AC-012, AC-217 |
| E-7-05 | A reviewer rolls back a merge through a server-supplied `change_set` row-history action and observes survivor and loser host rows return to the workbook surface. | REQ-01-089..REQ-01-111, REQ-02-219..REQ-02-220 | AC-010, AC-412         |

---

## 10. Phase 8 — Links, tags, saved views, sorting, filtering, grouping, startup selection, and projection-backed query semantics

### 10.8.1 Scope

This phase completes workbook configuration and projection-backed navigation:

- typed record links and incident-scoped tags,
- saved-view object lifecycle with exact scope vocabulary `private`, `shared`, `system`,
- workbook startup pointers `home_sheet_ref` and `default_sheet_ref`,
- query-shape validation for sort, filter, and group,
- Timeline grouping-key whitelist,
- group headers as presentation-only artifacts,
- no-op saved-view patch semantics,
- duplicate-view semantics.

### 10.8.2 Primary owner sections

- Core 01 §3.3.4 view-shaped read contract
- Core 01 §3.3.5.2 saved-view and workbook-preference routes
- Core 01 §7.4 base view-schema registry
- Core 01 §8 projection model
- Core 02 §11 saved views and workbook preferences
- Core 02 §12 typed relationships and tags
- Core 03 §2 workbook surfaces, saved views, and startup selection
- Core 03 §14 sorting, filtering, and grouping
- Core 04 §2 authorization model

### 10.8.3 Unit tests

| ID          | Test                                                                                                                                                                                                                                                       | Exact REQs                                                                                                                                                                                     | Exact ACs                                      |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| U-8-01      | Typed links and tags are stored as structured relationship rows, not only in JSON payloads, and use the closed base relationship vocabulary.                                                                                                               | REQ-02-011, REQ-02-163..REQ-02-176                                                                                                                                                             | AC-205..AC-210                                 |
| U-8-02      | Ordinary saved-view create defaults omitted `scope` to `private`, rejects `scope='system'`, and persists exactly one `view_schema_id` per saved view.                                                                                                      | REQ-03-012..REQ-03-023, REQ-01-138..REQ-01-151                                                                                                                                                 | AC-146..AC-152                                 |
| U-8-03      | Saved-view scope uses exactly `private`, `shared`, and `system`. No obsolete `team` scope token is accepted.                                                                                                                                               | REQ-03-017..REQ-03-020                                                                                                                                                                         | AC-146..AC-149                                 |
| U-8-04      | Ordinary saved-view patch allows only mutable members, rejects immutable-field mutation, and leaves `saved_view_version` / `updated_at` unchanged on a structurally valid no-op.                                                                           | REQ-03-024..REQ-03-026, REQ-01-138..REQ-01-151                                                                                                                                                 | AC-152                                         |
| U-8-05      | `home_sheet_ref` and `default_sheet_ref` remain separate objects. Workbook-open fallback order is explicit launch pointer, home pointer, default pointer, then Timeline. Invalid or hidden pointers are cleared before fallback continues.                 | REQ-03-027..REQ-03-032                                                                                                                                                                         | AC-150, AC-153                                 |
| U-8-06      | `filters[]`, `sort[]`, and `group_by` use stable `field_key` values only. Invalid operators, invalid arg shapes, duplicate filter keys after normalization, and invalid grouping keys are rejected by the route-owned query contract.                      | REQ-01-035..REQ-01-047, REQ-03-223..REQ-03-233                                                                                                                                                 | AC-124, AC-184, AC-243, AC-024..AC-026         |
| U-8-07      | Timeline grouping permits only the declared whitelist keys and never serializes group headers as writable rows.                                                                                                                                            | REQ-03-225..REQ-03-235                                                                                                                                                                         | AC-024..AC-026                                 |
| U-8-08      | View-query sort and filter ceilings, canonical normalization, and `meta.query` applied-sort tail expansion follow the exact route and saved-view contracts. Omitted sort means `no user sort override`, and `Group: None` is represented only by omission. | REQ-01-035..REQ-01-046, REQ-02-010, REQ-03-224, REQ-03-227                                                                                                                                     | AC-359..AC-361                                 |
| U-8-09      | View-schema discovery exposes exact `sort_fields`, `header_sort_field_key`, grouping-field whitelists, null-last ordering, and no client-sortable `record_id`. Header sort on non-sortable collection fields synthesizes no client sort.                   | REQ-01-286, REQ-01-310, REQ-01-312, REQ-01-323, REQ-01-326, REQ-01-328, REQ-01-329, REQ-01-331, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-499, REQ-01-503..REQ-01-506, REQ-03-223..REQ-03-235 | AC-362..AC-365                                 |
| U-8-10      | Full-row and sparse-patch wire families include hidden writable fields, preserve authoritative nulls, and canonicalize `changed_field_keys[]` plus `affected_views[]` ordering without guessing partial patches.                                           | REQ-00-027, REQ-01-034, REQ-01-036, REQ-01-267, REQ-01-310, REQ-03-097, REQ-03-247                                                                                                             | AC-366..AC-368                                 |
| U-8-GRID-01 | Header sort, filter controls, and group controls send only stable contract keys and never send visible labels, vendor column indexes, projection-table names, or storage-table names.                                                                      | REQ-01-022, REQ-01-035..REQ-01-047, REQ-03-223..REQ-03-235                                                                                                                                     | AC-124, AC-127, AC-184, AC-243, AC-359..AC-365 |

### 10.8.4 Integration tests

| ID     | Test                                                                                                                                                                                                                                                                              | Exact REQs                                     | Exact ACs      |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | -------------- |
| I-8-01 | Saved-view create, update, duplicate, and delete persist normalized `query_json` / `layout_json`, scope rules, and authorization consequences against a real database.                                                                                                            | REQ-01-138..REQ-01-151, REQ-03-012..REQ-03-026 | AC-146..AC-152 |
| I-8-02 | Workbook startup selection follows the documented fallback order across valid, missing, hidden, and invalid saved-view references.                                                                                                                                                | REQ-01-145..REQ-01-151, REQ-03-027..REQ-03-032 | AC-150, AC-153 |
| I-8-03 | Link and tag mutations update projections, history, and view-query results atomically.                                                                                                                                                                                            | REQ-02-163..REQ-02-176, REQ-01-351..REQ-01-353 | AC-205..AC-210 |
| I-8-04 | Live-authorized cursor pagination re-derives current authorization on continuation, rejects tampered cursor tokens, returns fetch-time row payloads, and lets fresh queries evaluate current live state from the beginning. | REQ-01-035..REQ-01-036, REQ-01-554..REQ-01-560 | AC-372..AC-375 |

### 10.8.5 E2E tests

| ID     | Test                                                                                                                                                                         | Exact REQs                                     | Exact ACs                                      |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| E-8-01 | A user creates a private saved view, converts it to shared when allowed, duplicates a visible system view, and cannot edit the system view in place through ordinary routes. | REQ-03-017..REQ-03-026                         | AC-146..AC-152                                 |
| E-8-02 | Workbook open honors explicit `sheet_ref`, then home pointer, then incident default, then Timeline, clearing invalid pointers along the way.                                 | REQ-03-027..REQ-03-032                         | AC-150, AC-153                                 |
| E-8-03 | Sorting, filtering, and grouping on Timeline produce a stable first useful viewport and deterministic final grouping order without turning group headers into rows.          | REQ-01-034..REQ-01-047, REQ-03-223..REQ-03-235 | AC-014, AC-024..AC-026, AC-044, AC-184, AC-185 |
| E-8-04 | Exact-token `full_text` and strict `prefix` behavior do not degrade into fuzzy, phrase, wildcard, stemming, transliteration, or relevance-ranked search.                     | REQ-01-042, REQ-01-565..REQ-01-567             | AC-387                                         |
| E-8-05 | Saved-view query/layout persistence, workbook-preference startup selection, and reload query replay are proven through public `/api/v1/` routes rather than in-memory state. | REQ-01-034..REQ-01-047, REQ-03-027..REQ-03-032, REQ-03-223..REQ-03-235 | AC-124, AC-127, AC-146, AC-153, AC-184, AC-185 |

---

## 11. Phase 9 — Keyboard contract, clipboard and bulk-edit behaviors, Notes, Indicators, Parties, Assessments, and analyst-work surfaces

### 11.9.1 Scope

This phase completes the remaining workbook-native operator surfaces and high-value interaction behaviors:

- full keyboard contract,
- clipboard paste and bulk-edit behaviors,
- Notes built-in tab,
- Indicators system view behavior,
- Parties system view and text-plus-link flows,
- Compromise Assessments,
- Task Requests and Decisions,
- coordination surfaces `comm_log`, `handoff`, `status_review`, and `lesson`,
- optional standardized surfaces for findings, investigative queries, and forensic keywords when the implementation exposes them.

### 11.9.2 Primary owner sections

- Core 01 §7.4 base view-schema registry
- Core 01 §19 party and coordination-surface addendum
- Core 02 §10 notes, indicators, assessments, task requests, decisions, and artifact-backed coordination rows
- Core 02 §19 party model
- Core 03 §2 workbook surfaces
- Core 03 §11 clipboard paste and bulk create
- Core 03 §13 keyboard contract
- Core 03 §16–§20 workbook-native analyst-work and party flows
- Core 04 §2 authorization model

### 11.9.3 Unit tests

| ID          | Test                                                                                                                                                                                                                                                                                                                    | Exact REQs                                                                                                                                                 | Exact ACs                                      |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| U-9-01      | The keyboard contract supports Arrow navigation, Enter, Shift+Enter, Tab, Ctrl+V, Ctrl+K, Space, Alt+H, and Esc on the workbook surface without introducing hidden macro semantics.                                                                                                                                     | REQ-03-217..REQ-03-222                                                                                                                                     | AC-003, AC-005                                 |
| U-9-02      | Multi-cell paste uses the shared tabular-ingest contract, preserves `entity_binding_mode`, and groups the paste into one visible user action while allowing later same-field conflict resolutions to create separate change sets.                                                                                       | REQ-03-145..REQ-03-152                                                                                                                                     | AC-003, AC-040                                 |
| U-9-03      | Notes are artifact-backed `note` rows exposed through the built-in Notes tab, not through a separate storage silo.                                                                                                                                                                                                      | REQ-02-067..REQ-02-071, REQ-03-004                                                                                                                         | AC-068..AC-070, AC-112                         |
| U-9-04      | Indicator system-view rows remain canonical indicator rows with pivots to source-bound observations and lifecycle history. Observation and lifecycle state stay separate from the canonical row identity.                                                                                                               | REQ-03-005..REQ-03-006, REQ-02-072..REQ-02-082                                                                                                             | AC-078, AC-079, AC-121, AC-122                 |
| U-9-05      | Party create or explicit create-from-text reuses a same-incident party only on unique exact match of normalized `primary_email` or `external_ref`. Raw requester or source text remains preserved alongside any linked `party_id`.                                                                                      | REQ-02-022, REQ-02-060..REQ-02-063, REQ-03-266..REQ-03-271                                                                                                 | AC-277..AC-280                                 |
| U-9-06      | Compromise assessments remain append-only, use only the closed assessment-state vocabulary, and keep operational response actions out of `assessment_state`. Confidence-band derivation remains deterministic.                                                                                                          | REQ-02-083..REQ-02-093, REQ-03-250..REQ-03-254                                                                                                             | AC-018, AC-080..AC-084                         |
| U-9-07      | Task Requests and Decisions are first-class workbook surfaces with bounded lifecycle transitions, queue fields, owner semantics, and fail-closed contradiction handling for derived decision-state inconsistencies. No generalized workflow engine or mandatory timeline approval fields are introduced.                | REQ-00-016, REQ-00-019, REQ-02-094..REQ-02-122, REQ-02-189..REQ-02-196, REQ-03-121..REQ-03-126, REQ-03-255..REQ-03-260                                     | AC-085, AC-086, AC-137..AC-145, AC-313, AC-314 |
| U-9-08      | `comm_log`, `handoff`, `status_review`, and `lesson` are workbook-native surfaces with the declared minimum create signal and projection-backed filter or grouping fields.                                                                                                                                              | REQ-01-503..REQ-01-506, REQ-02-123..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265                                                             | AC-281..AC-284                                 |
| U-9-09      | If the implementation exposes standardized Findings, Investigative Queries, or Forensic Keywords surfaces, each exposed surface uses its declared `view_schema_id`, minimum create signal, writable fields, and workbook-native behavior.                                                                               | REQ-01-507..REQ-01-509, REQ-02-135..REQ-02-138                                                                                                             | AC-285..AC-287                                 |
| U-9-10      | Writable timestamp fields bound to `timestamp_instant_v1` accept only RFC 3339 strings with explicit timezones, clear only through explicit JSON `null` when the field is clearable, reject timezone-less or otherwise invalid timestamp payloads, and preserve invalid timestamp drafts as client-local unsaved state. | REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-487..REQ-01-488, REQ-01-503..REQ-01-506, REQ-03-281                         | AC-300..AC-304, AC-354                         |
| U-9-11      | Direct Party-reference fields for requester, collector, and source parties accept exact same-incident active Party identifiers only, distinguish omission from explicit clear, preserve text-plus-ref independence, and reject non-direct-write clear shapes. Decision references remain covered by their own direct-reference evidence. | REQ-01-059..REQ-01-061, REQ-01-328, REQ-01-336, REQ-01-502, REQ-01-516..REQ-01-519, REQ-02-017, REQ-02-021..REQ-02-022, REQ-02-231, REQ-03-272, REQ-03-274 | AC-315..AC-318                                 |
| U-9-12      | Collection-style manual relationship fields on assessments, tasks, decisions, and coordination surfaces preserve `confidence=null`, reject client-supplied `confidence`, and keep authoritative `null` through projection and ordinary export.                                                                          | REQ-01-311, REQ-01-333..REQ-01-340, REQ-01-503..REQ-01-506, REQ-02-248, REQ-03-280                                                                         | AC-395..AC-397                                 |
| U-9-13      | The artifact-backed variant registry keeps Notes as the required built-in sheet and the coordination surfaces as standardized workbook-native `view_schema` identities. Optional standardized surfaces remain additive and do not replace the required base surface registry.                                           | REQ-01-309, REQ-02-250..REQ-02-253, REQ-03-004, REQ-03-011                                                                                                 | AC-410                                         |
| U-9-GRID-01 | Keyboard navigation updates Cartulary focus anchors, not vendor selection state alone. Arrow, Tab, Enter, Shift+Enter, and Esc leave the adapter with a valid or intentionally cleared Cartulary anchor.                                                                                                                | REQ-01-015..REQ-01-017, REQ-03-217..REQ-03-222                                                                                                             | AC-005, AC-047                                 |

### 11.9.4 Integration tests

| ID          | Test                                                                                                                                                                                  | Exact REQs                                                                                     | Exact ACs                                      |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| I-9-01      | A multi-row clipboard paste into Timeline creates ordered mutations, preserves unknown-column remnants where required, and respects `mention_origin` vs `entity_origin`.              | REQ-03-145..REQ-03-152, REQ-02-030..REQ-02-036                                                 | AC-003, AC-201                                 |
| I-9-02      | Notes, indicator, party, assessment, task, decision, and coordination surfaces persist structured state and remain queryable through workbook-native projections.                     | REQ-01-296..REQ-01-302, REQ-01-303..REQ-01-306, REQ-01-497..REQ-01-506, REQ-02-067..REQ-02-133 | AC-068..AC-090, AC-116..AC-122, AC-277..AC-284 |
| I-9-03      | Party-link helper fields on task or evidence records update only the hidden same-incident `*_party_id` link and do not overwrite source-preserving text.                              | REQ-01-502, REQ-02-021..REQ-02-022, REQ-03-268..REQ-03-271                                     | AC-278..AC-280                                 |
| I-9-GRID-01 | Multi-cell paste into sorted or filtered rows preserves visual paste order after translation to stable anchors. Existing rows update by `record_id`; new rows use create-row anchors. | REQ-03-145..REQ-03-152, REQ-01-022, REQ-01-057..REQ-01-060                                     | AC-003, AC-125, AC-126, AC-201                 |

### 11.9.5 E2E tests

| ID          | Test                                                                                                                                                                                                                                                                                            | Exact REQs                                                             | Exact ACs                                      |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------- |
| E-9-01      | Arrow keys, Tab, Enter, Shift+Enter, Ctrl+V, Ctrl+K, Space, Alt+H, and Esc all work on the grid without forcing a page or module switch.                                                                                                                                                        | REQ-03-217..REQ-03-222                                                 | AC-005                                         |
| E-9-PASTE-02 | Pasting a representative 20×5 block into the Timeline sheet creates or updates rows through the workbook surface and preserves row identity and selection state.                                                                                                                              | REQ-03-145..REQ-03-152                                                 | AC-003                                         |
| E-9-CONFLICT-02 | Grouped same-surface paste conflicts preserve workbook selection continuity while presenting conflict navigation.                                                                                                                                                                         | REQ-03-145..REQ-03-152                                                 | AC-003                                         |
| E-9-03      | Notes are available as a built-in tab, support in-grid creation, and can link to other records without leaving the workbook interaction model.                                                                                                                                                  | REQ-03-004, REQ-02-067..REQ-02-071                                     | AC-068..AC-070, AC-112, AC-116                 |
| E-9-04      | An analyst creates or links a party from requester or source text, keeps the raw text visible, and gains incident-scoped party pivots and queues.                                                                                                                                               | REQ-02-022, REQ-03-266..REQ-03-271                                     | AC-277..AC-280                                 |
| E-9-05      | Recording a new assessment appends a new attributed row. The sequence `unknown -> suspected -> confirmed -> cleared` and the alternative path `unknown -> disproven` remain distinguishable in history and filters.                                                                             | REQ-02-083..REQ-02-093, REQ-03-250..REQ-03-254                         | AC-018, AC-080..AC-084                         |
| E-9-TASKDECISION-06 | Task Requests and Decisions appear as workbook surfaces rather than separate application modules, with native queue, due-date, blocked-work, linked-record, and decision-reference behavior.                                                                                      | REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260                         | AC-085..AC-090                                 |
| E-9-COORDINATION-06 | Communications Log, Handoff, Status Review, and Lesson appear as workbook-native coordination surfaces with functional handoff, blocked-work, next-report, and lesson workflows.                                                                                                  | REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260                         | AC-281..AC-284                                 |
| E-9-07      | If the build exposes standardized Findings, Investigative Queries, or Forensic Keywords surfaces, each exposed surface behaves as a workbook-native surface with the declared minimum create semantics.                                                                                         | REQ-01-507..REQ-01-509, REQ-02-135..REQ-02-138                         | AC-285..AC-287                                 |
| E-9-08      | The workbook surface registry exposes Notes plus the required coordination surfaces by their canonical `view_schema_id` identities; any optional standardized artifact-backed surface, when exposed, is additive and does not substitute for Notes or any other required base workbook surface. | REQ-01-307..REQ-01-310, REQ-02-250..REQ-02-253, REQ-03-004..REQ-03-011 | AC-410                                         |
| E-9-GRIDANCHORS-01 | Keyboard navigation remains anchored by stable `record_id` and `field_key` values across required workbook surfaces that use the shared grid adapter.                                                                                                                            | REQ-03-005..REQ-03-011, REQ-03-217..REQ-03-222                         | AC-005, AC-085..AC-090, AC-281..AC-284         |
| E-9-GRIDHOST-01 | Host entity-origin clipboard paste uses the shared grid paste dispatcher and backend exact-match/stub behavior without routing through file import.                                                                                                                                       | REQ-03-145..REQ-03-152                                                 | AC-003                                         |
| E-9-GRIDIDENTITY-01 | Identity entity-origin clipboard paste uses the shared grid paste dispatcher and backend exact-match/stub behavior without routing through file import.                                                                                                                               | REQ-03-145..REQ-03-152                                                 | AC-003                                         |

---

## 12. Phase 10 — Operational backup, restore, and restore verification

This phase is the final base-profile implementation phase. It is intentionally late because coherent restore verification depends on a populated deployment and on at least one successful built-in workbook query after restore.[^base-manifest][^core01-routes]

### 12.10.1 Scope

This phase completes deployment-local recovery behavior for the base profile:

- retained `backup_set` metadata and retention floors,
- coherent restore of Postgres and object storage from the same retained `backup_set`,
- projection rebuild as part of restore readiness,
- isolated restore verification cadence and state transitions,
- fail-closed restore on missing required artifacts or integrity proofs,
- absence of public backup, restore, and restore-verification route families,
- deployment-local operator-facing control surfaces only,
- `roots.backup_storage` contract coverage.

### 12.10.2 Primary owner sections

- Core 01 §12.1 backup
- Core 01 §12.2 restore and operator recovery CLI contract
- Core 04 §2 deployment-admin and local operator authorization boundaries
- Core 04 §6 runtime roots and Core 04 §12.3.3 backup storage binding
- Core 04 §9.14 additional Base Profile criteria for backup and restore contract

### 12.10.3 Unit tests

| ID      | Test                                                                                                                                                                                                                                                                         | Exact REQs                                                                         | Exact ACs      |
| ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | -------------- |
| U-10-01 | Retained `backup_set` metadata uses the exact `verification_state` vocabulary, preserves one coherent `consistency_point_at`, verifies required artifact durability with redacted diagnostics before latest-success classification, and enforces the latest-success plus 30-day retention floors.       | REQ-01-571..REQ-01-574                                                             | AC-398         |
| U-10-02 | Restore readiness requires selection of exactly one retained `backup_set`, restore of Postgres and object storage from the same declared point, and projection rebuild before the target environment is ready.                                                               | REQ-01-575, REQ-01-423..REQ-01-424, REQ-01-577                                     | AC-399         |
| U-10-03 | Missing required Postgres or object-storage artifacts, or missing required checksum or integrity proof, fails selected-backup restore before readiness; due restore verification updates `verification_state` and `last_verified_restore_at` only according to the isolated verification result. | REQ-01-576, REQ-01-578                                                             | AC-400, AC-401 |
| U-10-04 | The public route inventory exposes no `/api/v1/backups*`, `/api/v1/restores*`, `/api/v1/restore-verifications*`, or matching `/ws/v1/*` family, and no browser, workbook, incident, common-job, HTTP, WebSocket, session, CSRF, Origin, or `deployment_admin` state authorizes recovery CLI invocation. | REQ-01-570, REQ-04-106                                                             | AC-402         |
| U-10-05 | `roots.backup_storage` is a required persistent runtime root with only the allowed binding kinds, distinct from `roots.export_outputs` and `roots.temporary_work`, and bound to encrypted storage when it carries incident data.                                             | REQ-04-053, REQ-04-058, REQ-04-071..REQ-04-073, REQ-04-076, REQ-04-107..REQ-04-108 | AC-403         |

### 12.10.4 Integration tests

| ID      | Test                                                                                                                                                                                                                                                                              | Exact REQs                         | Exact ACs      |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- | -------------- |
| I-10-01 | The most recent successful retained `backup_set` exposes the required metadata, restore anchors, encrypted artifact proofs, retention timestamps, and verification-state transitions against real backing storage.                                                                | REQ-01-571..REQ-01-574             | AC-398         |
| I-10-02 | Restoring the latest successful retained `backup_set` into a fresh environment rebuilds projections, opens at least one incident, executes at least one built-in workbook query when incident data is present, and preserves authoritative row, change-set, and blob consistency. | REQ-01-571, REQ-01-575..REQ-01-578 | AC-399, AC-401 |
| I-10-03 | Selecting a retained `backup_set` that is missing a required artifact or integrity proof fails before the target environment becomes ready.                                                                                                                                       | REQ-01-576                         | AC-400         |

### 12.10.5 E2E tests

| ID      | Test                                                                                                                                                                                            | Exact REQs                                                 | Exact ACs      |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- | -------------- |
| E-10-01 | A deployment-local operator can inspect the latest durable successful retained `backup_set` with redacted durability diagnostics, restore the latest durable retained backup, and run due restore verification from a marked isolated target.           | REQ-01-572..REQ-01-573, REQ-01-578                         | AC-398, AC-401 |
| E-10-02 | Restoring the latest successful retained `backup_set` into a fresh deployment recovers the workbook surface and executes at least one built-in workbook query successfully.                     | REQ-01-575..REQ-01-578                                     | AC-399, AC-401 |
| E-10-03 | Public `/api/v1/*` and `/ws/v1/*` surfaces expose no backup, restore, restore-verification, or backup-inspection families; no browser, workbook, incident, common-job, HTTP, WebSocket, session, CSRF, Origin, or `deployment_admin` state authorizes recovery CLI invocation. | REQ-01-570, REQ-04-106                                     | AC-402         |
| E-10-04 | Effective deployment configuration requires `roots.backup_storage` with allowed binding kinds and rejects attempts to satisfy backup storage with export or temporary-work roots.               | REQ-04-058, REQ-04-071..REQ-04-073, REQ-04-107..REQ-04-108 | AC-403         |

---

## 13. Phase 11 — Extension profile testing hooks

Extension profiles are not part of the base implementation sequence. This section exists so teams do not drift away from the current extension route families or profile claim boundaries while planning later work.

Each extension claim still requires the **base profile first**. The AC groups listed below are the **extension deltas** that should be planned on top of the fully passing base guide.[^profile-dod]

### 13.11.1 Import Extension Profile

**Primary owner sections**

- Core 01 §2.1 Phase 2 Workbook Import Assistant
- Core 01 §17.2 Import Extension Profile public contract
- Core 02 §7.2 file-based import provenance
- Core 03 §11 file-based import assistant
- Core 04 §4.5 hostile workbook and import-content constraints

**Extension delta ACs**

- `AC-027..AC-029`
- `AC-063..AC-067`
- `AC-232`
- `AC-262..AC-265`
- `AC-323..AC-325`
- `AC-393`

**Key boundaries to test**

- import sessions and import units,
- deterministic `mapping_fingerprint`,
- CSV plus bounded XLSX discovery and mapping,
- provenance capture for source bytes, parser version, and locator,
- inert handling of formulas, macros, automation, and external links,
- no auto-resolution during ingest,
- stable `imports`-module boundary,
- import-target registry fail-closed behavior,
- owner create facade coverage for every supported import target,
- parser isolation from owner modules, workbook stores, projection writers, and grid-vendor state,
- apply retry safety through unit-level journal and change-set evidence.

### 13.11.2 Snapshot and Reporting Extension Profile

**Primary owner sections**

- Core 01 §17.3 Snapshot and Reporting Extension Profile public contract
- Core 02 §14 snapshot and release metadata
- Core 04 §2.1 release gate and §4.2 export trust boundary

**Extension delta ACs**

- `AC-030..AC-032`
- `AC-056..AC-062`
- `AC-071`
- `AC-091`
- `AC-104..AC-106`
- `AC-113..AC-115`
- `AC-233`
- `AC-266..AC-269`
- `AC-305..AC-307`
- `AC-333`

**Key boundaries to test**

- immutable snapshot creation,
- release-state machine,
- bound approval tuple,
- invalidation on byte or tuple change,
- self-contained outputs,
- recipient-specific redaction profiles without live-workspace withholding.

### 13.11.3 Reference Pack Extension Profile

**Primary owner sections**

- Core 01 §17.4 Reference Pack Extension Profile public contract
- Core 02 reference-pack metadata and activation state
- Core 04 §4.1 reference-pack trust boundary

**Extension delta ACs**

- `AC-033..AC-035`
- `AC-092..AC-096`
- `AC-234`
- `AC-270..AC-272`
- `AC-308..AC-310`
- `AC-326`
- `AC-369`

**Key boundaries to test**

- staged import,
- verification metadata,
- explicit activation,
- retained prior active version,
- smallest disconnected bundle,
- fail-closed activation on integrity failure.

### 13.11.4 Incident Portability Extension Profile

**Primary owner sections**

- Core 01 §17.5 Incident Portability Extension Profile public contract
- Core 01 portability bundle and import phases
- Core 04 §4.2 portability trust boundary

**Extension delta ACs**

- `AC-164..AC-169`
- `AC-236`
- `AC-273..AC-276`
- `AC-327..AC-328`
- `AC-332`
- `AC-386`
- `AC-409`

**Key boundaries to test**

- deterministic bundle layout,
- authoritative-source-only export,
- checksum-verified staged import,
- preservation of attribution without importing login-capable deployment-local admin state,
- tolerance for unsupported optional embedded sections.

### 13.11.5 Enterprise Authentication Extension Profile

**Primary owner sections**

- Core 01 §20 Enterprise Authentication Extension Profile public contract
- Core 04 §1.2 enterprise-auth model
- Core 04 §12.3.4 enterprise-auth provider manifest binding

**Extension delta ACs**

- `AC-036`
- `AC-235`
- `AC-288..AC-293`
- `AC-348..AC-352`
- `AC-433..AC-436`

**Key boundaries to test**

- provider-manifest path validation before readiness,
- closed OIDC and SAML manifest schema validation, including defaults, explicit `null`, bounds, duplicate members, secret references, and referenced certificate files,
- startup reconciliation for new, updated, disabled, omitted, and type-changed providers,
- providers list and provider begin routes,
- OIDC callback and SAML ACS completion,
- provider-backed sign-in terminating into the same opaque session contract as base auth,
- deployment-admin binding create, rotate, and retire routes,
- no auto-provisioning of local users or incident memberships,
- no runtime provider-definition or provider-secret mutation route.

### 13.12 Phase 12 — Network Flow Activity extension profile

Network Flow Activity is an extension-profile adoption phase, not a Base profile phase. The repository map for this phase is active so public phase-slice targets can plan and report the row inventory, but every row remains blocked until the Network Flow implementation, immutable fixtures, transcripts, and executable selectors are available. A blocked Phase 12 row is traceability only and never closes conformance.

**Primary owner sections**

- Network Flow Activity NLSpec §23 and Table 23-A
- Core 00 extension-profile registry and final status transition sections
- Core 01 §§3.3.3.1, 3.3.6.1, 3.3.6.2, 3.3.7, 17, and 17.2
- Core 02 §§10.2, 14.4, and 18
- Core 03 §§2 and 4.3.1
- Core 04 §§2, 3, 9.1B, and 12.3
- Graph Projection NLSpec ephemeral projection sections
- Testing Harness NLSpec §§8, 11, 12, 16, and 17

**Phase 12 acceptance rows**

| ID | Test | Exact REQs | Exact ACs |
| --- | --- | --- | --- |
| E-12-NFAC001-01 | Blocked Network Flow selector for NF-AC-001: When the extension is unclaimed, the Network Analysis tab and `/api/v1/incidents/{incident_id}/network-flow/*` routes are unavailable through claimed-extension behavior. | NF-REQ-180; Network Flow Table 23-A | NF-AC-001 |
| E-12-NFAC002-02 | Blocked Network Flow selector for NF-AC-002: When the extension is claimed and an incident has no flow tables, the Network Analysis tab renders the empty state with `Import NetFlow CSV` for an authorized importer. | NF-REQ-180; Network Flow Table 23-A | NF-AC-002 |
| I-12-NFAC003-03 | Blocked Network Flow selector for NF-AC-003: Every Network Flow route uses Core success/error envelopes and rejects unknown request members unless the route declares them. | NF-REQ-180; Network Flow Table 23-A | NF-AC-003 |
| I-12-NFAC004-04 | Blocked Network Flow selector for NF-AC-004: Duplicate JSON object members at any depth are rejected before durable state or idempotency allocation. | NF-REQ-180; Network Flow Table 23-A | NF-AC-004 |
| I-12-NFAC005-05 | Blocked Network Flow selector for NF-AC-005: Invalid UTF-8, malformed JSON, non-object bodies, explicit invalid nulls, and unknown members fail request admission deterministically. | NF-REQ-180; Network Flow Table 23-A | NF-AC-005 |
| E-12-NFAC006-06 | Blocked Network Flow selector for NF-AC-006: Importing `NF-FIX-001-cisco-sna-minimal` through the Import Extension flow creates exactly one active `network_flow_table` and one inner table tab. | NF-REQ-180; Network Flow Table 23-A | NF-AC-006 |
| I-12-NFAC007-07 | Blocked Network Flow selector for NF-AC-007: The created table tab display name is derived deterministically from the uploaded filename and collision suffix rules. | NF-REQ-180; Network Flow Table 23-A | NF-AC-007 |
| I-12-NFAC008-08 | Blocked Network Flow selector for NF-AC-008: Import apply replay with the same committed `client_txn_id` returns the same table ID and creates no second table. | NF-REQ-180; Network Flow Table 23-A | NF-AC-008 |
| I-12-NFAC009-09 | Blocked Network Flow selector for NF-AC-009: Duplicate CSV headers are displayed and mapped by source column ordinal, not by label alone. | NF-REQ-180; Network Flow Table 23-A | NF-AC-009 |
| U-12-NFAC010-10 | Blocked Network Flow selector for NF-AC-010: Cisco SNA required fields are enforced exactly for `cisco_sna_netflow_csv_v1`. | NF-REQ-180; Network Flow Table 23-A | NF-AC-010 |
| I-12-NFAC011-11 | Blocked Network Flow selector for NF-AC-011: Reserved source profiles are not claimable and fail with `network_flow_unsupported_source_profile`. | NF-REQ-180; Network Flow Table 23-A | NF-AC-011 |
| U-12-NFAC012-12 | Blocked Network Flow selector for NF-AC-012: Empty CSV, forbidden header controls, header-only CSV, terminal newline, blank line, malformed quote, quote escaping, and field-count mismatch match §9.2 outcomes; the first logical record is always the header. | NF-REQ-180; Network Flow Table 23-A | NF-AC-012 |
| I-12-NFAC013-13 | Blocked Network Flow selector for NF-AC-013: Formula-looking CSV values are inert during import and never evaluated. | NF-REQ-180; Network Flow Table 23-A | NF-AC-013 |
| I-12-NFAC014-14 | Blocked Network Flow selector for NF-AC-014: Invalid rows produce deterministic diagnostics and do not appear in table query or graph results. | NF-REQ-180; Network Flow Table 23-A | NF-AC-014 |
| I-12-NFAC015-15 | Blocked Network Flow selector for NF-AC-015: A partially valid CSV creates a table containing accepted rows and exposes rejected-row counts. | NF-REQ-180; Network Flow Table 23-A | NF-AC-015 |
| I-12-NFAC016-16 | Blocked Network Flow selector for NF-AC-016: An all-invalid CSV creates no table tab and returns `network_flow_all_rows_rejected`. | NF-REQ-180; Network Flow Table 23-A | NF-AC-016 |
| U-12-NFAC017-17 | Blocked Network Flow selector for NF-AC-017: `source_row_digest_v1`, `normalized_row_digest_v1`, `mapping_fingerprint_v1`, `network_flow_graph_query_digest_v1`, and `network_flow_source_snapshot_digest_v1` produce exact expected fixture digests or IDs. | NF-REQ-180; Network Flow Table 23-A | NF-AC-017 |
| U-12-NFAC018-18 | Blocked Network Flow selector for NF-AC-018: Timestamp profile parsing rejects DST folds, DST gaps, leap seconds, invalid local-time inference, and end-before-start rows. | NF-REQ-180; Network Flow Table 23-A | NF-AC-018 |
| U-12-NFAC019-19 | Blocked Network Flow selector for NF-AC-019: IPv6 outputs follow the canonical text contract, IPv4 leading-zero octets are rejected, and IPv6 zone identifiers are rejected. | NF-REQ-180; Network Flow Table 23-A | NF-AC-019 |
| U-12-NFAC020-20 | Blocked Network Flow selector for NF-AC-020: `uint64_decimal_string_v1` rejects signs, exponents, decimals, leading zeroes except `0`, and values above max. | NF-REQ-180; Network Flow Table 23-A | NF-AC-020 |
| U-12-NFAC021-21 | Blocked Network Flow selector for NF-AC-021: The table lifecycle state machine contains no `renamed` state. | NF-REQ-180; Network Flow Table 23-A | NF-AC-021 |
| I-12-NFAC022-22 | Blocked Network Flow selector for NF-AC-022: Renaming a table changes only display metadata, increments `table_version`, and does not change row IDs, graph IDs, provenance, mapping fingerprints, diagnostics, live query cursor validity, or indicator bindings. | NF-REQ-180; Network Flow Table 23-A | NF-AC-022 |
| E-12-NFAC023-23 | Blocked Network Flow selector for NF-AC-023: Soft-deleting a table removes it from the default tab strip, removes it from `all_active_tables`, invalidates active graph queries and cursors that include it, and leaves it terminal. | NF-REQ-180; Network Flow Table 23-A | NF-AC-023 |
| I-12-NFAC024-24 | Blocked Network Flow selector for NF-AC-024: A table query uses stable `field_key` filters and rejects visible column labels. | NF-REQ-180; Network Flow Table 23-A | NF-AC-024 |
| I-12-NFAC025-25 | Blocked Network Flow selector for NF-AC-025: Filter `in` arrays reject duplicates and empty arrays. | NF-REQ-180; Network Flow Table 23-A | NF-AC-025 |
| U-12-NFAC026-26 | Blocked Network Flow selector for NF-AC-026: CIDR filtering rejects IPv4/IPv6 family mismatch and does not treat IPv4-mapped IPv6 as IPv4. | NF-REQ-180; Network Flow Table 23-A | NF-AC-026 |
| I-12-NFAC027-27 | Blocked Network Flow selector for NF-AC-027: Sort null ordering, IP ordering, decimal counter ordering, duplicate sort-key rejection, `network_flow.endpoint_ip` sort rejection, and default sort tail match §13.4. | NF-REQ-180; Network Flow Table 23-A | NF-AC-027 |
| I-12-NFAC028-28 | Blocked Network Flow selector for NF-AC-028: Cursor continuation fails after TTL expiry, actor mismatch, route mismatch, query mismatch, table soft delete, or authorization loss. | NF-REQ-180; Network Flow Table 23-A | NF-AC-028 |
| E-12-NFAC029-29 | Blocked Network Flow selector for NF-AC-029: The filter selector always includes explicit table-selection controls in graph mode. | NF-REQ-180; Network Flow Table 23-A | NF-AC-029 |
| E-12-NFAC030-30 | Blocked Network Flow selector for NF-AC-030: Default graph mode from a table tab uses `table_scope.mode='active_table'`. | NF-REQ-180; Network Flow Table 23-A | NF-AC-030 |
| I-12-NFAC031-31 | Blocked Network Flow selector for NF-AC-031: `selected_tables` with duplicate table IDs is rejected rather than silently deduplicated. | NF-REQ-180; Network Flow Table 23-A | NF-AC-031 |
| I-12-NFAC032-32 | Blocked Network Flow selector for NF-AC-032: Selecting multiple tables composes one graph using only rows from those tables after filters and time slicing. | NF-REQ-180; Network Flow Table 23-A | NF-AC-032 |
| I-12-NFAC033-33 | Blocked Network Flow selector for NF-AC-033: The same canonical endpoint IP across two selected tables emits one default vertex with both contributing table IDs. | NF-REQ-180; Network Flow Table 23-A | NF-AC-033 |
| I-12-NFAC034-34 | Blocked Network Flow selector for NF-AC-034: Default edges aggregate by `(src_ip, dst_ip, ip_protocol, dst_port)` and preserve contributing table provenance. | NF-REQ-180; Network Flow Table 23-A | NF-AC-034 |
| I-12-NFAC035-35 | Blocked Network Flow selector for NF-AC-035: Time interval filtering uses overlap semantics and handles zero-duration flows exactly as specified. | NF-REQ-180; Network Flow Table 23-A | NF-AC-035 |
| E-12-NFAC036-36 | Blocked Network Flow selector for NF-AC-036: Selecting a graph vertex pivots to contributing rows without using visible labels, graph coordinates, row order, or layout state as identity. | NF-REQ-180; Network Flow Table 23-A | NF-AC-036 |
| E-12-NFAC037-37 | Blocked Network Flow selector for NF-AC-037: Selecting a graph edge opens the cross-table contributing rows drawer grouped by table and ordered deterministically. | NF-REQ-180; Network Flow Table 23-A | NF-AC-037 |
| I-12-NFAC038-38 | Blocked Network Flow selector for NF-AC-038: Graph over-limit cases fail with deterministic errors and do not render partial unlabeled graphs. | NF-REQ-180; Network Flow Table 23-A | NF-AC-038 |
| I-12-NFAC039-39 | Blocked Network Flow selector for NF-AC-039: `example_row_refs[]`, `example_refs_truncated`, and `example_refs_total_count` match §14.5. | NF-REQ-180; Network Flow Table 23-A | NF-AC-039 |
| I-12-NFAC040-40 | Blocked Network Flow selector for NF-AC-040: Graph Projection adapter input contains the exact adapter fields and safe metadata in §14.4. | NF-REQ-180; Network Flow Table 23-A | NF-AC-040 |
| I-12-NFAC041-41 | Blocked Network Flow selector for NF-AC-041: Linking an endpoint to an existing indicator creates or returns a source-bound binding and does not rewrite the flow row. | NF-REQ-180; Network Flow Table 23-A | NF-AC-041 |
| I-12-NFAC042-42 | Blocked Network Flow selector for NF-AC-042: Creating an indicator from an endpoint uses the Core indicator owner behavior and fails if the caller lacks the required incident role. | NF-REQ-180; Network Flow Table 23-A | NF-AC-042 |
| I-12-NFAC043-43 | Blocked Network Flow selector for NF-AC-043: Duplicate indicator-link requests return the existing binding with `duplicate=true` and create no second binding. | NF-REQ-180; Network Flow Table 23-A | NF-AC-043 |
| I-12-NFAC044-44 | Blocked Network Flow selector for NF-AC-044: Selectors referencing rejected rows, soft-deleted tables, stale graph digests, or `unmapped_raw` fail. | NF-REQ-180; Network Flow Table 23-A | NF-AC-044 |
| I-12-NFAC045-45 | Blocked Network Flow selector for NF-AC-045: A `deployment_admin` without incident membership cannot list, query, import, graph, soft-delete, rename, or link flow data in that incident. | NF-REQ-180; Network Flow Table 23-A | NF-AC-045 |
| I-12-NFAC046-46 | Blocked Network Flow selector for NF-AC-046: No network-flow import, graph, table query, rejected-row query, or link action performs third-party egress in v1. | NF-REQ-180; Network Flow Table 23-A | NF-AC-046 |
| I-12-NFAC047-47 | Blocked Network Flow selector for NF-AC-047: Logs, telemetry, audit summaries, and diagnostics obey the raw-value handling matrix in §16.2. | NF-REQ-180; Network Flow Table 23-A | NF-AC-047 |
| I-12-NFAC048-48 | Blocked Network Flow selector for NF-AC-048: Effective limits are discoverable, lowerable only to declared minimums, and enforced at the phases in §20. | NF-REQ-180; Network Flow Table 23-A | NF-AC-048 |
| U-12-NFAC049-49 | Blocked Network Flow selector for NF-AC-049: Every route-local error includes the detail schema required by §21.2. | NF-REQ-180; Network Flow Table 23-A | NF-AC-049 |
| U-12-NFAC050-50 | Blocked Network Flow selector for NF-AC-050: Large-fixture timing results are classified as engineering measurements unless Core 05 claim-publication requirements are separately satisfied. | NF-REQ-180; Network Flow Table 23-A | NF-AC-050 |
| I-12-NFAC051-51 | Blocked Network Flow selector for NF-AC-051: `unmapped_raw` values are retained as inert provenance under the default policy and are not filterable, sortable, graphable, or linkable. | NF-REQ-180; Network Flow Table 23-A | NF-AC-051 |
| U-12-NFAC052-52 | Blocked Network Flow selector for NF-AC-052: Every fixture row has concrete path, byte SHA-256, approved mapping JSON, mapping fingerprint, expected IDs, expected diagnostics, and expected graph/link output before adoption. | NF-REQ-180; Network Flow Table 23-A | NF-AC-052 |
| U-12-NFAC053-53 | Blocked Network Flow selector for NF-AC-053: The document contains no `MAY` statement whose omission behavior is absent. | NF-REQ-180; Network Flow Table 23-A | NF-AC-053 |
| U-12-NFAC054-54 | Blocked Network Flow selector for NF-AC-054: No behavior is moved into appendices, research reports, UI guides, implementation guides, or vendor documentation as normative authority. | NF-REQ-180; Network Flow Table 23-A | NF-AC-054 |
| U-12-NFAC055-55 | Blocked Network Flow selector for NF-AC-055: Internal section references, table references, error-code references, and requirement references resolve without dangling anchors before adoption. | NF-REQ-180; Network Flow Table 23-A | NF-AC-055 |
| I-12-NFAC056-56 | Blocked Network Flow selector for NF-AC-056: `network_flow_flow_edge_id_v1` produces exact fixture edge IDs, and null destination ports form a distinct aggregation and edge-ID key. | NF-REQ-180; Network Flow Table 23-A | NF-AC-056 |
| I-12-NFAC057-57 | Blocked Network Flow selector for NF-AC-057: Request members `time_range.bucket` and `aggregation.include_time_buckets` fail as unknown members in v1; no time-bucket error code is exposed. | NF-REQ-180; Network Flow Table 23-A | NF-AC-057 |
| I-12-NFAC058-58 | Blocked Network Flow selector for NF-AC-058: `observation_mode` values other than `binding_only` fail, and `created_observation_refs[]` is always `[]` in binding resources and link responses. | NF-REQ-180; Network Flow Table 23-A | NF-AC-058 |
| I-12-NFAC059-59 | Blocked Network Flow selector for NF-AC-059: Graph query digests and source snapshot IDs remain unchanged when deployment graph limits or caller `limit_overrides` change without changing query semantics. | NF-REQ-180; Network Flow Table 23-A | NF-AC-059 |
| U-12-NFAC060-60 | Blocked Network Flow selector for NF-AC-060: Timestamp profile precision rejects finer source resolution, epoch modes reject fractional values, and sys-uptime parsing uses export time, exporter uptime-at-export, and per-field event uptime exactly as §9.7 specifies. | NF-REQ-180; Network Flow Table 23-A | NF-AC-060 |
| I-12-NFAC061-61 | Blocked Network Flow selector for NF-AC-061: Duplicate table rename fails with `network_flow_invalid_display_name` and `reason_code='duplicate_display_name'`, while existing cursors over the renamed table continue. | NF-REQ-180; Network Flow Table 23-A | NF-AC-061 |
| I-12-NFAC062-62 | Blocked Network Flow selector for NF-AC-062: Only `active` and `soft_deleted` are table lifecycle states; import staging is not a table, direct soft-deleted references follow NF-REQ-058, and Table 8-C limit accounting is exact. | NF-REQ-180; Network Flow Table 23-A | NF-AC-062 |
| I-12-NFAC063-63 | Blocked Network Flow selector for NF-AC-063: `network_flow_all_rows_rejected` includes ordered `diagnostics_sample[]`, `row_count_rejected`, and `diagnostics_truncated` details without creating a table. | NF-REQ-180; Network Flow Table 23-A | NF-AC-063 |
| I-12-NFAC064-64 | Blocked Network Flow selector for NF-AC-064: Omitted query limits materialize to `min(200, effective max_query_limit)`, explicit query limits outside range fail, and graph limit overrides outside `[lowerable_min, effective]` fail. | NF-REQ-180; Network Flow Table 23-A | NF-AC-064 |
| I-12-NFAC065-65 | Blocked Network Flow selector for NF-AC-065: Validation preview counts are limited to the parser preview slice, and an apply may fail with `network_flow_all_rows_rejected` even when preview accepted rows. | NF-REQ-180; Network Flow Table 23-A | NF-AC-065 |
| I-12-NFAC066-66 | Blocked Network Flow selector for NF-AC-066: Binding `source_row_refs[]`, `source_row_refs_truncated`, and `source_row_refs_total_count` are populated per Table 15-F, and duplicate `row_refs[]` selector entries fail. | NF-REQ-180; Network Flow Table 23-A | NF-AC-066 |
| I-12-NFAC067-67 | Blocked Network Flow selector for NF-AC-067: Indicator-link dedupe uses the resolved binding identity tuple; structurally different selectors resolving to the same tuple dedupe, and different source-row-ref sets create distinct bindings. | NF-REQ-180; Network Flow Table 23-A | NF-AC-067 |
| I-12-NFAC068-68 | Blocked Network Flow selector for NF-AC-068: `confirm_exact_value` mismatch fails before mutation with `network_flow_indicator_link_ambiguous` and the required indicator-link error details. | NF-REQ-180; Network Flow Table 23-A | NF-AC-068 |
| I-12-NFAC069-69 | Blocked Network Flow selector for NF-AC-069: Graph Projection adapter input uses Network Flow endpoint IDs as source entity IDs, flow edge IDs as source relationship IDs, and no invalid `ephemeral_response_only` retention token. | NF-REQ-180; Network Flow Table 23-A | NF-AC-069 |
| I-12-NFAC070-70 | Blocked Network Flow selector for NF-AC-070: Interface fields store bounded text or null only; numeric interface identifiers remain text and sort by code point. | NF-REQ-180; Network Flow Table 23-A | NF-AC-070 |
| I-12-NFAC071-71 | Blocked Network Flow selector for NF-AC-071: Unknown aggregation modes fail request admission, no unavailable-mode error is exposed, and mapping combinability accepts only `single_source_only` in v1. | NF-REQ-180; Network Flow Table 23-A | NF-AC-071 |
| I-12-NFAC072-72 | Blocked Network Flow selector for NF-AC-072: Table rename, table delete, indicator-link create, and Core import apply replay follow the idempotency comparison and replay points in Table 5-B. | NF-REQ-180; Network Flow Table 23-A | NF-AC-072 |
| I-12-NFAC073-73 | Blocked Network Flow selector for NF-AC-073: Mapping approval accepts only the three §10.4 variants; `derived`, `constant`, fake ignored-field sentinels, and variant-extra members fail deterministically. | NF-REQ-180; Network Flow Table 23-A | NF-AC-073 |
| E-12-NFAC074-74 | Blocked Network Flow selector for NF-AC-074: Alias suggestions use `source_alias_match_key_v1`; duplicate alias matches produce blocking visible warnings until explicit user approval resolves every matched column. | NF-REQ-180; Network Flow Table 23-A | NF-AC-074 |
| U-12-NFAC075-75 | Blocked Network Flow selector for NF-AC-075: Every Network Flow success response `data` object contains the schema, required members, ordering, nullable behavior, and `meta` shape defined in §14.6 or §17. | NF-REQ-180; Network Flow Table 23-A | NF-AC-075 |
| I-12-NFAC076-76 | Blocked Network Flow selector for NF-AC-076: Every Table 20-A resource limit uses the configured default/minimum behavior, enforcement phase, error/truncation behavior, and invalid-config rejection specified in §20. | NF-REQ-180; Network Flow Table 23-A | NF-AC-076 |
| I-12-NFAC077-77 | Blocked Network Flow selector for NF-AC-077: Safe samples and source-column samples follow Table 12-F exactly, including null samples for raw text/IP-like values and numeric-only samples for bounded integer-like values. | NF-REQ-180; Network Flow Table 23-A | NF-AC-077 |
| I-12-NFAC078-78 | Blocked Network Flow selector for NF-AC-078: Route-local errors include Table 21-B details and choose the first error under Table 21-C and Table 21-D ordering. | NF-REQ-180; Network Flow Table 23-A | NF-AC-078 |
| I-12-NFAC079-79 | Blocked Network Flow selector for NF-AC-079: Indicator linking accepts only v1 IP endpoint candidates, rejects every non-linkable field in Table 15-A1, and validates existing/create targets against Core canonical IP indicator identity. | NF-REQ-180; Network Flow Table 23-A | NF-AC-079 |
| I-12-NFAC080-80 | Blocked Network Flow selector for NF-AC-080: Source filename display, default table display names, explicit display-name overrides, duplicate suffixing, hidden-file stems, trailing-dot stems, and soft-delete name reuse match §8.4. | NF-REQ-180; Network Flow Table 23-A | NF-AC-080 |
| I-12-NFAC081-81 | Blocked Network Flow selector for NF-AC-081: Final import commit is atomic under failure injection and never exposes `creating`, `failed`, a ghost table ID, partial rows, partial diagnostics, or a domain audit without its table. | NF-REQ-180; Network Flow Table 23-A | NF-AC-081 |
| I-12-NFAC082-82 | Blocked Network Flow selector for NF-AC-082: Explicit import display names fail on duplicates without suffixing, while omitted display names use deterministic suffix allocation under the same incident-scoped commit lock. | NF-REQ-180; Network Flow Table 23-A | NF-AC-082 |
| U-12-NFAC083-83 | Blocked Network Flow selector for NF-AC-083: CSV preview consumes exactly the first 50 complete logical data records, includes blank and mismatched records in preview counts, and does not report an error located only after that boundary. | NF-REQ-180; Network Flow Table 23-A | NF-AC-083 |
| I-12-NFAC084-84 | Blocked Network Flow selector for NF-AC-084: `max_rows_per_csv` counts logical data records including blank and mismatched rows, excludes the header, and reports `actual=limit+1` without scanning to a total. | NF-REQ-180; Network Flow Table 23-A | NF-AC-084 |
| U-12-NFAC085-85 | Blocked Network Flow selector for NF-AC-085: Cisco SNA v1 permits only nine required targets plus two optional interface targets; exporter, TCP flags, and application label mappings fail as unsupported and public rows contain nulls for them. | NF-REQ-180; Network Flow Table 23-A | NF-AC-085 |
| U-12-NFAC086-86 | Blocked Network Flow selector for NF-AC-086: `trim_ascii_space_v1` trims only U+0020 and the empty-value policy is applied after transformation exactly as NF-REQ-095 specifies. | NF-REQ-180; Network Flow Table 23-A | NF-AC-086 |
| U-12-NFAC087-87 | Blocked Network Flow selector for NF-AC-087: Timestamp objects reject cross-variant members, enforce exact RFC3339 and epoch grammar, bind non-UTC mappings to `tzdb-2026c`, and enforce unsigned 32-bit uptime plus distinct ordinals. | NF-REQ-180; Network Flow Table 23-A | NF-AC-087 |
| I-12-NFAC088-88 | Blocked Network Flow selector for NF-AC-088: Import facade preview has no durable side effect; apply receives no filesystem path or URL; descriptors, source digest, defaults, and mapping fingerprint are server-derived; changed source fails closed. | NF-REQ-180; Network Flow Table 23-A | NF-AC-088 |
| I-12-NFAC089-89 | Blocked Network Flow selector for NF-AC-089: Every source column has one contiguous descriptor and one disposition; every required/system target has exactly one valid mapping; observation provenance is materialized server-side once. | NF-REQ-180; Network Flow Table 23-A | NF-AC-089 |
| I-12-NFAC090-90 | Blocked Network Flow selector for NF-AC-090: Public row objects always contain every nullable optional field, exact `unmapped_raw` value objects, and the namespaced `network_flow.observation_source_ref`. | NF-REQ-180; Network Flow Table 23-A | NF-AC-090 |
| I-12-NFAC091-91 | Blocked Network Flow selector for NF-AC-091: Diagnostic discovery under parallel validation yields byte-identical ordered resources, stable reason/message keys, and exact conditional limit fields. | NF-REQ-180; Network Flow Table 23-A | NF-AC-091 |
| I-12-NFAC092-92 | Blocked Network Flow selector for NF-AC-092: Table scope variants reject cross-variant members, duplicate selected IDs, over-limit selection, and zero-table resolved scopes without revealing hidden resources. | NF-REQ-180; Network Flow Table 23-A | NF-AC-092 |
| I-12-NFAC093-93 | Blocked Network Flow selector for NF-AC-093: Filter normalization canonicalizes values and `in` order, rejects post-canonicalization duplicates, and binds the normalized array rather than caller spelling or order. | NF-REQ-180; Network Flow Table 23-A | NF-AC-093 |
| I-12-NFAC094-94 | Blocked Network Flow selector for NF-AC-094: Initial and continuation request variants are mutually exclusive; cursor tokens reject non-ASCII and more than 4096 bytes; a token is invalid at `now == expires_at`. | NF-REQ-180; Network Flow Table 23-A | NF-AC-094 |
| I-12-NFAC095-95 | Blocked Network Flow selector for NF-AC-095: Accepted-row, diagnostic, and graph-contributor cursors use their specified full independent keyset tuples and never skip or duplicate equal-prefix items. | NF-REQ-180; Network Flow Table 23-A | NF-AC-095 |
| I-12-NFAC096-96 | Blocked Network Flow selector for NF-AC-096: Counter aggregates use arbitrary precision, accept values beyond uint64 when within the digit limit, and fail vertices, edges, bytes digits, then packet digits in the specified order without partial output. | NF-REQ-180; Network Flow Table 23-A | NF-AC-096 |
| I-12-NFAC097-97 | Blocked Network Flow selector for NF-AC-097: Graph Projection metadata contains exact ordered table IDs and mapping fingerprints, fixed metadata mappings, and maps adapter outcomes to Table 14-G6 without internal leakage. | NF-REQ-180; Network Flow Table 23-A | NF-AC-097 |
| I-12-NFAC098-98 | Blocked Network Flow selector for NF-AC-098: Graph success data uses exact `semantic_query`, source-table-ref, result-limit, and annotation schemas with no vague or implementation-defined members. | NF-REQ-180; Network Flow Table 23-A | NF-AC-098 |
| I-12-NFAC099-99 | Blocked Network Flow selector for NF-AC-099: Contributor queries recompute current composition and authorization, paginate all contributing accepted rows deterministically, and fail stale or missing vertex/edge selectors without falling back to example refs. | NF-REQ-180; Network Flow Table 23-A | NF-AC-099 |
| I-12-NFAC100-00 | Blocked Network Flow selector for NF-AC-100: Indicator selectors and targets are closed variants, confirmation requires byte-exact canonical IP text, and create-indicator clients cannot supply normalized or display values. | NF-REQ-180; Network Flow Table 23-A | NF-AC-100 |
| I-12-NFAC101-01 | Blocked Network Flow selector for NF-AC-101: Core indicator create/dedupe and binding insert/reuse commit atomically; `duplicate` appears only in the link result; new and reused bindings return HTTP 201 and 200 respectively. | NF-REQ-180; Network Flow Table 23-A | NF-AC-101 |
| I-12-NFAC102-02 | Blocked Network Flow selector for NF-AC-102: Every safe digest resource or audit field carries its key ID, and key rotation preserves comparison only within equal key IDs. | NF-REQ-180; Network Flow Table 23-A | NF-AC-102 |
| I-12-NFAC103-03 | Blocked Network Flow selector for NF-AC-103: Audit occurrences match Table 16-C exactly, exact replay emits no new domain occurrence, and graph truncated-ref count equals the specified sum. | NF-REQ-180; Network Flow Table 23-A | NF-AC-103 |
| I-12-NFAC104-04 | Blocked Network Flow selector for NF-AC-104: Soft delete and incident closure retain Network Flow state but make it non-queryable as specified in Table 8-D, invalidate affected cursors, preserve Core-governed audit retention, and expose no v1 whole-incident purge claim. | NF-REQ-180; Network Flow Table 23-A | NF-AC-104 |
| I-12-NFAC105-05 | Blocked Network Flow selector for NF-AC-105: Every route returns its exact success status, exact closed data schema, Table 21-A status, exhaustive reason code, safe details, and retry action. | NF-REQ-180; Network Flow Table 23-A | NF-AC-105 |
| U-12-NFAC106-06 | Implemented Network Flow selector for NF-AC-106: Every document dependency has an adopted version and immutable locator, every blocker in §24 is closed, and every Table 22-A fixture has concrete immutable bytes before status changes from draft. | NF-REQ-180; Network Flow Table 23-A | NF-AC-106 |
| I-12-NFAC107-07 | Blocked Network Flow selector for NF-AC-107: Import cancellation before commit leaves no table, while cancellation or worker failure after commit recovers and publishes the one committed success without duplicate table creation. | NF-REQ-180; Network Flow Table 23-A | NF-AC-107 |

---

## 14. Shared cross-cutting harnesses

These harnesses apply across phases and should be implemented once, then reused. The harness list is intentionally small and tied to current owner sections. When a phase introduces multiple public routes under the same surface, the harnesses should run from a shared route inventory rather than from scattered one-off assertions; Phase 1 auth and deployment-local user-administration surfaces already require that route-inventory discipline in the repo task surface, and Phase 2 incident surfaces continue the same rule. The same task-surface rule applies to browser harness ownership: webserver-backed browser rows belong in scheduler-safe service-backed browser sources and are duration-balanced by Playwright spec file, while isolated stateful, measurement, and visual browser suites keep their reset policy in the browser batch manifest under one owned shared stack per verification stage and remain scheduler-visible through browser-stage resource claims.

Browser helpers that exercise visible UI workflows should be semantic-state helpers, not bare click sequences. A helper that opens a form, submits it, or navigates into a workbook must use distinct stable selectors for those actions and wait for the durable postcondition that owns the evidence, for example the selected `incident_id` route plus workbook-shell readiness. Reusing one selector for state-dependent controls is harness debt because it makes Playwright evidence sensitive to transient React replacement rather than user intent.

### 14.1 Envelope consistency harness

Every HTTP success and error response on a public route must use the owner-level common envelope shape. No public route returns a bare object, bare array, or bespoke error wrapper.

**Owner sections**: Core 01 §3.3.6 and route-family owners in Core 01 §3.3.2, §3.3.4, §3.3.5, §3.3.8, §3.3.9, §17, and §20.

### 14.2 Authorization re-derivation harness

Route handlers, handle issuance, handle redemption, job polling, job cancel, and incident WebSocket subscription must re-derive authorization from the caller’s current role and current incident membership at request time.

**Owner sections**: Core 04 §2, plus route-specific owners in Core 01 §3.3.5, §3.3.8, §3.3.9, and §3.3.10.

### 14.3 Mutation attribution and history-emission harness

Every successful mutating route must emit the required actor, timestamp, source, and mutation detail for its owner substrate. Incident data goes through the incident history substrate. Deployment-local account and membership administration goes through the deployment-local administrative audit substrate.

**Owner sections**: Core 02 §14–§15, Core 04 §3.

### 14.4 Idempotent replay and divergent replay harness

Every route that owns idempotency must be covered for:

- first success,
- same normalized replay,
- divergent replay with the same idempotency key,
- non-idempotent routes explicitly rejecting `client_txn_id` where the owner section says they are intentionally non-idempotent.

This harness must include `PATCH /api/v1/records/{record_id}` and therefore must cover `AC-299`. It MUST also include the bounded credential-lifecycle and deployment-admin credential-action routes that own route-scoped idempotency in the current profile.[^core01-routes]

**Owner sections**: Core 01 §3.3.5, Core 01 §3.3.8, Core 01 §3.3.9, Core 01 §17, Core 01 §20.

### 14.5 Closed-vocabulary rejection harness

All write paths must reject invalid tokens for the closed vocabularies they own, including `incident.status` and `incident.tlp` wherever a route binds those token families. Public incident APIs accept only canonical TLP tokens or JSON `null`; aliases and empty strings are migration-only concerns or invalid input as defined by the owner contract.

**Owner sections**: Core 02 §18 and any route-specific write contracts that bind those tokens.

### 14.6 Writable-string normalization harness

All writable string fields bound to an owner-level string contract must normalize, trim, reject controls, clear or preserve `null`, and enforce scalar-count limits exactly as the contract requires. Incident create and patch coverage must include `description` under `multiline_body_v1` and `severity`, `current_phase`, and `primary_external_case_ref` under `incident_metadata_text_v1`.

**Owner sections**: Core 01 owner-level writable-string contract bindings and the field-level write contracts in Core 01 §3.3.5.3, §7.4, §19, and Core 03 workbook write surfaces.

### 14.7 View-schema field-key conformance harness

Clients and tests must address writable and queryable fields only by stable `field_key`. Visible column labels, tab names, storage table names, or projection column names are never mutation keys.

**Owner sections**: Core 01 §3.3.4, §3.3.5, §7.4; Core 03 §14–§16.

### 14.8 Grid-adapter identity and capability harness

This harness verifies that the frontend grid adapter maps vendor-local coordinates to stable Cartulary anchors. It must run for any change to the grid adapter, renderer/editor registration, sync-engine integration, view-schema field adapters, query UI, paste handling, presence display, or conflict display. A grid-engine replacement is not accepted until this harness passes against the replacement. The RDG-specific rows below cover browser behaviors and package surfaces identified as high-risk in the `react-data-grid` research report.[^rdg-report]

Pure adapter mapping tests must not require a real browser. Lifecycle, focus, selection, paste, and visual-state tests must use a real browser or browser-equivalent renderer.

| ID          | Category          | Required test                                                                                                                                                                                  |
| ----------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `U-GRID-01` | Unit              | Adapter maps every rendered column from the active `view_schema` using stable `field_key`; visible labels are never write identifiers.                                                         |
| `U-GRID-02` | Unit              | Vendor row, column, selection, and plugin coordinates translate to `record_id` and `field_key`, or fail closed without enqueueing a mutation.                                                  |
| `U-GRID-03` | Unit              | Group headers, loading rows, spacer rows, and presentation-only rows never emit mutation-capable events.                                                                                       |
| `U-GRID-04` | Unit              | Disabled vendor capabilities are not registered, enabled, reachable, or callable in ordinary runtime.                                                                                          |
| `U-GRID-05` | Unit              | Imperative grid APIs capable of data, metadata, sort, filter, selection, validation, plugin, or lifecycle mutation are inaccessible outside `/packages/grid-adapter`.                          |
| `U-GRID-06` | Unit              | Renderer/editor lifecycle cleanup removes portals, subscriptions, timers, observers, and stale row references on unmount, remount, and surface switch.                                         |
| `U-GRID-07` | Unit              | Adapter rejects any body row lacking a non-empty unique `record_id` before rendering a mutation-capable grid.                                                                                  |
| `U-GRID-08` | Unit              | Adapter maps RDG `sortColumns` changes to view-query `sort[]` entries keyed by sortable `field_key`; it never sorts authoritative rows locally.                                                |
| `U-GRID-09` | Unit              | Adapter resolves editability only from explicit editor adapters plus contract writeability; RDG `editable=true` without `renderEditCell` never opens an editor.                                |
| `U-GRID-10` | Unit              | Adapter preserves unchanged row object references when applying sparse patches to row arrays.                                                                                                  |
| `I-GRID-01` | Integration       | Sorted, filtered, grouped, and cursor-replaced result sets preserve edits by `record_id` and `field_key`, not visible row position.                                                            |
| `I-GRID-02` | Integration       | Sparse patches, full row refreshes, invalidates, and query replacements follow the grid-state refresh semantics table.                                                                         |
| `I-GRID-03` | Integration       | Paste into a sorted or filtered visible range creates or updates the intended records in visual paste order after stable-anchor translation.                                                   |
| `I-GRID-04` | Integration       | Column resize changes are client-local until explicitly saved through allowed layout state; hidden vendor width state is not the only durable layout state.                                    |
| `I-GRID-05` | Integration       | Copy is allowed from readable read-only cells where the view contract allows reading, but paste, fill, and editor entry are blocked.                                                           |
| `I-GRID-06` | Integration       | Renderer precedence follows field-specific, type-family, safe text/value fallback, with no mutation side effects from renderers.                                                               |
| `E-GRID-01` | E2E               | After a live update reorders the visible result, the analyst's pending edit remains attached to the original `record_id` and cannot retarget to the row now occupying the old visual position. |
| `E-GRID-02` | E2E               | Same-cell conflict markers, presence markers, and save-state labels remain attached to the intended cell after sort, filter, grouping, virtual scroll, and live row patch.                     |
| `E-GRID-03` | E2E               | Keyboard entry, exit, active-cell restoration, header navigation, summary navigation, and RTL arrow behavior preserve or intentionally clear Cartulary anchors.                                |
| `E-GRID-04` | E2E               | Drag fill updates only writable compatible target cells and fails closed on read-only, grouped, synthetic, or incompatible cells.                                                              |
| `E-GRID-05` | E2E               | Column resizing works through mouse and keyboard paths and does not break active-cell, conflict-marker, or presence-marker anchoring.                                                          |
| `E-GRID-06` | E2E               | Virtual scrolling plus scroll-to-cell preserves active edit state and same-field conflict state by `record_id + field_key`.                                                                    |
| `E-GRID-07` | E2E               | Frozen columns plus `colSpan` plus horizontal scroll preserve keyboard navigation and do not produce mutation-capable covered cells.                                                           |
| `E-GRID-08` | E2E               | Treegrid/group rows expose expand/collapse keyboard behavior and ARIA state, while copy/paste and ordinary cell editing are restricted on group rows.                                          |
| `E-GRID-09` | E2E               | The frontend bundle imports the RDG stylesheet exactly once before rendering a workbook grid.                                                                                                  |
| `V-GRID-01` | Visual regression | Stable fixtures cover default viewport, unresolved mention, resolved chip, same-field conflict marker, row-gutter presence marker, grouped result, evidence count, and save-state strip.       |
| `V-GRID-02` | Visual regression | Visual fixtures cover frozen-column shadow, resize handle, drag-fill handle, edit-cell state, grouped/tree row, and light/dark theme classes when exposed.                                     |

**Owner sections**: Core 01 §3.1, §3.3.4, §3.3.5, §7.4; Core 03 §3–§4, §11, §13–§16; development guide §6 grid adapter contract, controlled-state mapping, and capability allowlist.

**Acceptance criteria**:

- §14.8 test IDs remain unique.
- §14.8 covers keyboard/focus, editability, copy/paste, drag fill, resizing, virtualization, frozen columns, `colSpan`, treegrid, CSS import, renderer precedence, row identity, and row-object identity.
- §14.8 includes explicit disabled-capability assertions and invariant failures for missing or duplicate `record_id`, and it preserves stable anchors by `record_id + field_key` through sorted, filtered, grouped, and invalidated refreshes.
- No §14.8 row treats visible row index as a valid mutation identity.

### 14.9 Workbook visual-regression harness

The repository must provide one browser visual-regression harness for workbook surfaces. The harness must use deterministic seed data, deterministic viewport size, deterministic browser zoom, deterministic fixture ordering, and masked dynamic regions for timestamps, cursors, avatars, generated IDs, and clock-derived labels.

The default viewport for this harness is `1440x900` CSS pixels unless the repo-local visual-test configuration declares a stricter fixed viewport. Visual regression is a developer gate for changes to the grid adapter, renderer, editor, workbook shell, theme, presence UI, conflict UI, saved-view UI, and coordination-surface grid UI. Visual regression results are implementation-quality evidence only. Public timed or fixture-sensitive claims still require Core 05 publication controls.

Workbook grid visual fixtures must declare the intended capture state before the screenshot is taken. At minimum, a grid-shell capture must make the viewport size, grid scroll top/left position or named scroll anchor, focus/editor state, inspector state, dynamic masks, and post-scroll settle wait explicit in the test. A fixture must not depend on incidental scroll state inherited from a prior click, editor open, inspector action, or viewport-continuity restoration. Visual baselines may be refreshed only after the deterministic capture state has been encoded and reviewed; a baseline update is not a substitute for removing nondeterminism from the fixture.

Horizontally virtualized workbook-grid captures that intend to show right-side fields or row actions MUST use a semantic named anchor rather than a live maximum-scroll contract. The anchor must name the required field keys or action region, poll until the intended headers/cells/actions are visible at stable geometry, and fail before screenshot capture when the geometry contract is not reached. Geometry diagnostics for these failures should include scroll metrics, visible field keys, required rects, active element or focused cell, and inspector state; those diagnostics are retained browser artifacts and inherit the same secret-bearing handling as screenshots, videos, traces, and Playwright reports.

The repository uses one owned-stack Playwright screenshot suite for this harness. It MUST NOT introduce a second visual runner just to satisfy workbook screenshot coverage.

| Fixture ID     | Required visible state                                                                                      |
| -------------- | ----------------------------------------------------------------------------------------------------------- |
| `VFIX-GRID-01` | Timeline default viewport with stable row identity and visible save-state strip.                            |
| `VFIX-GRID-02` | Timeline unresolved mention token and resolved entity chip in adjacent rows.                                |
| `VFIX-GRID-03` | Same-field conflict marker with resolver entry point visible.                                               |
| `VFIX-GRID-04` | Row-gutter presence marker and same-cell editing hint where supported.                                      |
| `VFIX-GRID-05` | Evidence count and preview affordance on a Timeline row.                                                    |
| `VFIX-GRID-06` | Grouped result with group headers rendered as presentation-only rows.                                       |
| `VFIX-GRID-07` | Task Requests or Decisions system view with queue-oriented fields visible.                                  |
| `VFIX-GRID-08` | Save-state strip showing `Syncing`, `Saved`, and `Conflict` across deterministic fixture states.            |
| `VFIX-GRID-09` | Frozen first column with horizontal scroll and frozen-column shadow visible.                                |
| `VFIX-GRID-10` | Column resize handle visible on a resizable header and focus state preserved on the active cell.            |
| `VFIX-GRID-11` | Drag-fill handle visible only on an editable active cell.                                                   |
| `VFIX-GRID-12` | Active edit cell with editor chrome, display fallback, and row identity preserved.                          |
| `VFIX-GRID-13` | Treegrid/group row with expand/collapse affordance, non-writable presentation styling, and leaf rows below. |
| `VFIX-GRID-14` | Dark theme grid state when dark theme is exposed.                                                           |
| `VFIX-GRID-15` | Empty successful query using no-rows fallback distinct from loading state.                                  |

**Owner sections**: Core 01 §3.1 and §3.3.4; Core 03 §1–§4, §11, §13–§16; Core 05 for claim-bearing publication boundaries only.

**Acceptance criteria**:

- §14.9 includes `VFIX-GRID-01` through `VFIX-GRID-15`.
- Visual fixtures distinguish loading, no rows, and ordinary empty group states.
- Dynamic regions remain masked according to the harness rules in this section.
- Guide-listed vendor capabilities that are not yet product-exposed are satisfied through explicit disabled-capability assertions, not pseudo-visual fixtures that imply support.

### 14.9A Grid browser-command harness

The browser test layer must provide shared helper commands for these interaction families:

| Helper family        | Required behavior                                                                                                                                                           |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Sort, filter, group  | Toggle header sort, apply or remove filter chips, change grouping controls, and assert that subsequent refreshes preserve Cartulary anchors by `record_id + field_key`.    |
| Column resize        | Locate a header resize handle, perform pointer resize, support keyboard resize path when applicable, and assert final width or layout change through observable grid state. |
| Drag fill            | Locate active editable cell drag handle, drag over target rows, release, and assert field-keyed mutation intents.                                                           |
| Scroll-to-cell       | Scroll a target `record_id + field_key` into view and assert that focus, conflict state, and presence state attach to that anchor.                                          |
| Tree expand/collapse | Toggle group rows through mouse and keyboard paths and assert ARIA expansion state.                                                                                         |
| Paste matrix         | Paste a deterministic TSV matrix into visible cells and assert record-keyed creates or patches in visual paste order.                                                       |
| Anchor assertion     | Provide a reusable assertion path for visible cell, conflict, presence, and save-state anchors keyed by `record_id + field_key` rather than visible row position.          |

Individual tests may add local setup, but they must not reimplement low-level pointer choreography when a shared helper exists.

**Acceptance criteria**:

- The testing guide defines shared command families for column resize, drag fill, scroll-to-cell, tree expand/collapse, and paste matrix.
- The testing guide defines shared command families for sort, filter, grouping, and anchor assertions in addition to the lower-level pointer helpers.
- The command harness requires assertions against Cartulary anchors, not only vendor DOM state.

### 14.9B Timeline zero-field-create traceability note

Timeline zero-field create remains enabled when the addressed `view_schema_id` allows it. In repo-local traceability, cite Core 01 `REQ-01-057` as the owner of row-create request shape and zero-field-create eligibility, and cite Core 04 `AC-191` and `AC-192` as the conformance checks for the Timeline-specific `client_txn_id`-only create and forbidden-field rejection behavior. `contracts/view-schemas/cartulary.view.timeline.v2.json` is a derived artifact that reflects that owner rule; it is not the behavior owner.

### 14.10 Projection determinism and rebuild harness

Projection rebuilds must be deterministic from authoritative source state. Projection corruption or rebuild must never alter authoritative source rows.

**Owner sections**: Core 01 §8.

### 14.11 WebSocket lifecycle harness

The socket boundary must enforce first-message rules, replay vs reset semantics, ephemeral presence handling, heartbeat timing, `session_revoked`, and delete or restore `record_changed` semantics.

**Owner sections**: Core 01 §3.3.10 and Core 03 §4.

### 14.12 Implementation timing-validation harness

Implementation-facing timed ACs must use the Core 04 observable completion semantics and end-user-visible completion time rather than backend-only timing. Passing this harness proves implementation conformance only. It does not authorize a public benchmark claim.[^claim-publication]

**Owner sections**: Core 04 §9 for implementation-timed criteria.

### 14.13 Claim-bearing benchmark-publication boundary

Teams MUST NOT treat passing §14.12 as sufficient for a public timed or fixture-sensitive claim. Claim-bearing publication additionally requires Core 05 and remains outside Base Profile and extension-profile implementation conformance.[^precedence][^claim-publication]

| Concern                                 | Authoritative owner | Minimum evidence                                                  |
| --------------------------------------- | ------------------- | ----------------------------------------------------------------- |
| Implementation-facing timed AC pass     | Core 04 §9          | ordinary conformance evidence under this guide                    |
| Benchmark environment identity          | Core 05 §2–§3       | exact `benchmark_profile_id` plus conformant `benchmark_manifest` |
| Public timed or fixture-sensitive claim | Core 05 §1–§4       | passing `PC-001..PC-006` with retained benchmark artifacts        |

### 14.14 Topology, preservation, and audit-source invariants harness

Conformance evidence must prove that the application remains one deployable unit, binary evidence does not persist inline in the authoritative structured store, rough capture remains recoverable after later normalization, unauthenticated state-changing mutations fail closed, startup-owned non-user mutations use an explicit system-process actor, and UI, import, and rollback provenance remain distinguishable from one another.[^core01-routes][^core03-workbook]

**Owner sections**: Core 01 §1; Core 02 §5; Core 04 §1, §3, and §5–§8.

---

## 15. Phase completion checklist

A phase is complete only when all of the following are true:

1. Every row in that phase’s executable authoritative manifest passes in the intended test layer, and the default task surface verifies that those rows actually executed rather than only matching names or symbols. Browser rows that share one real web-server bootstrap must also respect the task-surface orchestration contract rather than relying on incidental startup races.
2. The shared harnesses in §14 pass for every route, event class, and mutation path introduced or materially changed by the phase.
3. All earlier phases still pass after the phase lands.
4. Any new view surface introduced in the phase is covered for:
   - query shape,
   - row identity,
   - allowed create semantics,
   - patch semantics,
   - projection maintenance,
   - authorization,
   - history or audit consequences.
5. Any mutation-bearing phase includes explicit integration coverage against the real backing boundary it depends on. Use `N/A` only when the owner section truly defines no external boundary for that work.
6. No phase claims completion by relying on behavior deferred to a later phase when the current phase already assumes that behavior.

### 15.1 Base-profile completion rule

The base implementation sequence is complete only when Phases 0 through 10 pass, the shared harnesses in §14 pass for every applicable surface, and the current Base claim manifest in Core 04 is fully covered. Historical endpoint shorthand such as `AC-299` is not sufficient on its own.[^base-manifest][^traceability]

The current Base claim manifest is:

- `AC-001..AC-026`
- `AC-037..AC-055`
- `AC-068..AC-070`
- `AC-072..AC-090`
- `AC-097..AC-103`
- `AC-107..AC-112`
- `AC-116..AC-163`
- `AC-170..AC-231`
- `AC-238..AC-261`
- `AC-277..AC-287`
- `AC-294..AC-304`
- `AC-311..AC-322`
- `AC-329..AC-331`
- `AC-334..AC-347`
- `AC-353..AC-354`
- `AC-359..AC-368`
- `AC-370..AC-371`
- `AC-372..AC-375`
- `AC-376..AC-385`
- `AC-387..AC-392`
- `AC-394..AC-408`
- `AC-410`

---

## 16. Coverage ledger

The phase tables above are the authoritative **test-id to REQ / AC mapping**. This section adds the summary views needed to keep the guide from drifting again.

`tools/phase_registry.json` owns phase catalog metadata: phase order, status, manifest path, ledger path, scope, and owner refs. Active phases are executable; planned phases are visible to phase guidance but excluded from execution; retired phases must retain explicit artifact references and a retirement reason. Repository phase maps under `tools/phaseN_test_map.json` own the executable row inventory and ledger rendering details consumed by `make phase-ledgers` and `make phase-ledger-drift`. Each phase map is a self-identifying manifest: it must declare `schema_id="cartulary.phase_test_map.v2"` and a `phase` value matching `phase0` or `phase[1-9][0-9]*`, and that declared phase must match the registry path convention. Filenames remain the storage convention, not the discovery authority. Every manifest row in every numbered phase map must declare closed `evidence_class`, closed `layer`, and `default_check_required`; service-backed Go rows must declare explicit `fixture_policy.postgres` and `fixture_budget.postgres`. Every authoritative manifest row must declare non-empty `claim` and `out_of_scope` text so future phases receive the same ledger guarantees without validator code changes.

When a change adds or promotes an authoritative service-backed Go phase row, refresh `tools/go_test_duration_baselines.json` from uncontaminated successful service-backed evidence in the same change. Do not rely on fallback weights for newly declared rows; missing explicit baselines should remain a maintenance failure.

Execution topology is owned separately by `tools/execution_topology_manifest.json`, which supplies scheduler work-unit structure, profile-expanded check schedule inputs, service-backed schedule inputs, and browser batch manifest inputs. Public target identity, output class, artifact policy, schema policy, command identity, lifecycle state, and side-effect declarations remain owned by the testing harness public registry and `tools/task_surface_manifest.json`; execution topology supplies scheduler work-unit structure only. `make check` imports service-backed evidence as flattened check-scheduler leaves: service startup is a low-resource session unit that retains the suite-service lease, and backend/browser leaves claim ordinary check-scheduler host and service resources directly instead of passing through a nested service-backed scheduler or fixed host-to-`go_cpu` bridge. Direct `make test-service-backed`, `make test-fast-service-backed`, and `make check-service-backed` continue to use `tools/scheduler_manifest.json` as standalone developer entrypoints. Scheduler capacity defaults, override environment names, and auto-limit policies are owned by `tools/scheduler_resource_registry.json` capacity profiles rather than by Make defaults or generated schedule literals. Future phase growth should start by adding or activating the phase registry row, then adding the phase map row inventory; change the execution topology only when adding scheduler topology, scheduler resources, gate policy, check-scheduler profile, capacity profile selection, or browser stage/reset policy.

Service-backed Go unit helper starts are manifest-authorized, not phase-hardcoded. `make service-backed-unit-check` treats canonical `TestPhaseN..._U_N_...` Go tests as pure unit rows unless their manifest entry is an authoritative `go_test` unit row with `execution_dependency=backend_store`; only those rows may start `pgtest`, `s3test`, or owner-local `internal/modules/<owner>/testsupport` helpers directly. The guard is convention-driven for all numbered phases, including multi-digit phases, so adding a future owner helper package must not require checker code changes.

### 16.1 Phase-to-owner-section map

| Phase    | Primary owner sections                                                                                                    |
| -------- | ------------------------------------------------------------------------------------------------------------------------- |
| Phase 0  | Core 01 §1; Core 01 §3.3.5.1 bootstrap-admin manifest contract; Core 04 §5–§8; Core 04 §12                                |
| Phase 1  | Core 01 §3.3.2; Core 01 §3.3.2.2; Core 01 §3.3.5.1; Core 01 §3.3.7.1; Core 01 §3.3.10; Core 04 §1; Core 04 §3              |
| Phase 2  | Core 01 §3.3.1; Core 01 §3.3.3; Core 01 §3.3.5.1–§3.3.5.3; Core 01 §3.3.7.1; Core 02 §4.5; Core 02 §18; Core 04 §2–§3     |
| Phase 3  | Core 01 §3.3.5; Core 01 §7.4; Core 01 §8; Core 02 §5 and §14; Core 03 §1, §4, §6, §7, and §15                             |
| Phase 4  | Core 01 §3.3.5 mention and merge routes; Core 02 §6–§10, §13, and §19; Core 03 §8.4, §9, §16, and §20                    |
| Phase 5  | Core 01 §3.3.8; Core 01 §16; Core 02 §4.5 and §13; Core 03 §8 and §16; Core 04 §4.3 and §4.5                              |
| Phase 6  | Core 01 §3.3.10; Core 01 §3.3.5; Core 03 §3–§4; Core 04 §1–§2 and §4.5                                                    |
| Phase 7  | Core 01 §3.3.4.2 and §3.3.5; Core 02 §14–§15; Core 03 §10; Core 04 §2–§3                                                  |
| Phase 8  | Core 01 §3.3.4; Core 01 §3.3.5.2; Core 01 §7.4; Core 01 §8; Core 02 §11–§12; Core 03 §2 and §14; Core 04 §2               |
| Phase 9  | Core 01 §7.4 and §19; Core 02 §10 and §19; Core 03 §2, §11, §13, and §16–§20; Core 04 §2                                  |
| Phase 10 | Core 01 §12.1–§12.2; Core 04 §2, §6, §9.14, and §12.3.3                                                                   |
| Phase 11 | Core 01 §17 and §20; Core 02 extension-owned provenance, release, and bundle sections; Core 04 extension profile sections |
| Phase 12 | Network Flow Activity NLSpec §23; Core 00 extension-profile status; Core 01 §§3.3.3.1, 3.3.6, 3.3.7, and 17.2; Core 02 §§10.2, 14.4, and 18; Core 03 §§2 and 4.3.1; Core 04 §§2, 3, 9.1B, and 12.3 |

### 16.2 Base-profile AC coverage index

| Coverage state       | AC cluster                                                                                                                                                                               | Planned owner phase or shared harness                            | Notes                                                                                                                                                    |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Phase-owned          | `AC-294..AC-298`, `AC-320`, `AC-343..AC-346`                                                                                                                                             | Phase 0                                                          | Deployment configuration, runtime roots, bootstrap preflight, and resource-limit registry.                                                               |
| Phase-owned          | `AC-123`, `AC-130..AC-163`, `AC-175..AC-180`, `AC-244..AC-250`, `AC-311..AC-312`, `AC-334..AC-347`                                                                                       | Phase 1                                                          | Session contract, credential lifecycle, deployment-local user administration, and bounded credential reset flows.                                        |
| Phase-owned          | `AC-170..AC-174`, `AC-211..AC-214`, `AC-219..AC-220`, `AC-261`, `AC-370..AC-371`                                                                                                         | Phase 2                                                          | Incident create or patch, membership control, deployment-admin incident boundary, and extension discovery.                                               |
| Phase-owned          | `AC-001..AC-002`, `AC-107..AC-111`, `AC-125`, `AC-191..AC-199`, `AC-299`, `AC-329..AC-331`                                                                                               | Phase 3                                                          | Timeline create or patch hot path, lifecycle, patch idempotency, and supersede-with-replacement.                                                         |
| Shared               | `AC-009`, `AC-126`, `AC-203`                                                                                                                                                              | Phase 3 / Phase 6                                                | Phase 3 covers server-side field-level patch rebase, same-field conflict transport, and collection conflict values. Phase 6 remains owner for resolver, concurrent-client, and local pending-queue behavior. |
| Phase-owned          | `AC-006`, `AC-017`, `AC-019..AC-023`, `AC-077..AC-079`, `AC-128`, `AC-186`, `AC-188..AC-190`, `AC-201..AC-202`, `AC-205`, `AC-209`, `AC-221..AC-225`, `AC-251`, `AC-277`, `AC-281..AC-284`, `AC-388..AC-394` | Phase 4                                                          | Mention lifecycle, exact-match reuse, alias handling, interactive auto-resolution, merge, Parties, coordination route surfaces, object-blob create, and evidence handle issuance. Phase 4 browser rows are visible workflow evidence only where the current implementation exposes the surface. |
| Phase-owned          | `AC-004`, `AC-015..AC-016`, `AC-053`, `AC-102..AC-103`, `AC-154..AC-155`, `AC-252..AC-255`, `AC-321..AC-322`, `AC-405`                                                                   | Phase 5                                                          | Evidence lifecycle, safe preview or download redemption behavior, and binary-evidence storage boundary beyond Phase 4 route-shape evidence.              |
| Phase-owned          | `AC-008`, `AC-037..AC-042`, `AC-129`, `AC-131..AC-136`, `AC-156..AC-163`, `AC-204`, `AC-226..AC-230`, `AC-376..AC-382`                                                                    | Phase 6                                                          | WebSocket lifecycle, presence, same-field conflict resolver behavior, local pending queue, and save-state semantics.                                    |
| Phase-owned          | `AC-007`, `AC-010..AC-012`, `AC-181..AC-183`, `AC-187`, `AC-215..AC-218`, `AC-353`, `AC-383..AC-385`, `AC-412`, `AC-474`                                                                  | Phase 7                                                          | Reviewer history, delete or restore, rollback including merge-aware and Indicator-child rollback, destructive-operation lock precedence, and retained-history invariants. |
| Phase-owned          | `AC-013..AC-014`, `AC-024..AC-026`, `AC-124`, `AC-127`, `AC-146..AC-153`, `AC-184..AC-185`, `AC-200`, `AC-206..AC-208`, `AC-210`, `AC-359..AC-368`, `AC-372..AC-375`, `AC-387`           | Phase 8                                                          | Links, tags, saved views, query-shape contract, view discovery, sparse patches, and live-authorized pagination.                                           |
| Phase-owned          | `AC-003`, `AC-005`, `AC-018`, `AC-068..AC-090`, `AC-112`, `AC-116..AC-122`, `AC-137..AC-145`, `AC-278..AC-280`, `AC-285..AC-287`, `AC-300..AC-304`, `AC-313..AC-319`, `AC-354`, `AC-395..AC-397`, `AC-410` | Phase 9                                                          | Keyboard contract, built-in and remaining system surfaces, timestamp and direct-reference contracts, remaining coordination and party-linking flows, queue-heavy workflow expansion, and registry closure. Phase 4 smoke evidence does not complete these later interaction obligations. |
| Phase-owned          | `AC-398..AC-403`                                                                                                                                                                         | Phase 10                                                         | Operational backup, restore, restore verification, and backup-storage binding.                                                                           |
| Shared harness-owned | `AC-043..AC-047`, `AC-048..AC-055`, `AC-097..AC-103`, `AC-231`, `AC-238..AC-260`, `AC-404..AC-408`                                                                                       | Shared harnesses in §14                                          | Timing, security, aggregate claim gate, hostile-content and authorization boundaries, topology, rough-capture preservation, and audit-source invariants. |
| Phase-owned          | `AC-285..AC-287`                                                                                                                                                                         | Phase 9                                                          | Optional standardized workbook surfaces are exposed in the current build as additive workbook-native system views.                                       |

### 16.3 Extension-only AC index

| Extension profile         | Delta ACs beyond base                                                                                                                              | Planned hook phase |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| Import                    | `AC-027..AC-029`, `AC-063..AC-067`, `AC-232`, `AC-262..AC-265`, `AC-323..AC-325`, `AC-393`                                                         | Phase 11           |
| Snapshot and Reporting    | `AC-030..AC-032`, `AC-056..AC-062`, `AC-071`, `AC-091`, `AC-104..AC-106`, `AC-113..AC-115`, `AC-233`, `AC-266..AC-269`, `AC-305..AC-307`, `AC-333` | Phase 11           |
| Reference Pack            | `AC-033..AC-035`, `AC-092..AC-096`, `AC-234`, `AC-270..AC-272`, `AC-308..AC-310`, `AC-326`, `AC-369`                                               | Phase 11           |
| Incident Portability      | `AC-164..AC-169`, `AC-236`, `AC-273..AC-276`, `AC-327..AC-328`, `AC-332`, `AC-386`, `AC-409`                                                       | Phase 11           |
| Enterprise Authentication | `AC-036`, `AC-235`, `AC-288..AC-293`, `AC-348..AC-352`, `AC-433..AC-436`                                                                            | Phase 11           |
| Network Flow Activity | `NF-AC-001..NF-AC-107` | Phase 12 |

### 16.4 Conditional surface note

The standardized Findings, Investigative Queries, and Forensic Keywords surfaces are contract-defined as optional standardized workbook surfaces. The current build exposes all three as additive `scope='system'` workbook surfaces, so Phase 9 coverage includes `AC-285..AC-287` as implemented behavior. Future profiles that omit or replace optional standardized surfaces must update the implementation-owned coverage record and verify the conformance interpretation against Core 01 §7.4 and §19 plus Core 02 §10.4.4A and §10.4.6 before asserting the remaining Base claim.[^traceability][^core03-workbook][^core02-registry]

### 16.5 Maintenance rule for manifest drift

This guide MUST NOT be used as a completion artifact unless §1.2, §15.1, §16.2, and §16.3 have been regenerated whenever Core 04 §9.0 claim manifests or Appendix F profile Definition-of-Done navigation change. Updating only one of those sections is insufficient because it can reintroduce false Base or extension completion claims.[^base-manifest][^traceability]

---

## Sources

[^precedence]: `00_document_set_status_and_precedence.md`, especially §§1–5 on authority, precedence, contract ownership, and the separation between implementation conformance and claim publication.
[^traceability]: `F_source_traceability_matrix.md`, especially §§F.2, F.4, and F.6 for requirement-to-AC mapping and profile definition-of-done navigation.
[^base-manifest]: `04_security_deployment_and_conformance.md`, especially §9.0 claim manifests, §9.14 backup and restore criteria, and §12 deployment-configuration contract.
[^profile-dod]: `F_source_traceability_matrix.md`, §F.6 Profile Definition-of-Done navigation.
[^claim-publication]: `05_claim_publication_and_benchmark_reproducibility.md`, especially §§1–4 on the separation of claim-bearing publication from implementation conformance, `benchmark_profile_id`, `benchmark_manifest`, `measurement_predicate_id`, and `PC-001..PC-006`.
[^core01-routes]: `01_architecture_storage_and_view_contracts.md`, especially §§3.3.2, 3.3.3.1, 3.3.5.1, 7.4, 12.1–12.2, 16, 17, and 20.
[^core02-registry]: `02_domain_model_schema_and_history.md`, especially §10.4.4A on the tagged-variant registry and §19 on party and coordination semantics.
[^core03-workbook]: `03_workbook_interaction_collaboration_and_workflows.md`, especially §§2, 11, 13, and 16–20 on workbook surfaces, optional standardized surfaces, keyboard behavior, and party / coordination flows.
[^rdg-report]: `R09-react-data-grid-research-report.md`, especially §§3, 12–15, 17–21, and the evidence ledger for controlled state, browser interaction coverage, grouping/treegrid behavior, performance, CSS, build packaging, and fragile grid combinations.
