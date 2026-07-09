# Auth Module Refactor Tracker - Next Iteration

Last updated: 2026-07-09.

Target path: `internal/modules/auth`.

This handoff supersedes `docs/handoffs/auth-module-module-refactor-tracker.md` for future Auth Module planning. It treats that tracker and archived auth trackers as evidence only. It does not authorize runtime behavior changes, public API changes, migrations, generated-contract edits, or compatibility layers.

## 1. Scope and authority

Authority order:

1. Adopted subsystem NLSpecs, for their named subsystem only.
2. Core 00 through Core 04 for current implementation conformance.
3. Core 05 only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
4. `docs/domain.md` for vocabulary and boundary interpretation.
5. `docs/testing-harness-nlspec.md` for harness mechanics and evidence handling.
6. Current repository code and tests for current implementation state.
7. Prior handoffs and archived trackers as evidence only.

Primary inputs for this iteration:

- `docs/handoffs/auth-module-module-refactor-tracker.md`.
- `docs/archive/auth-module-refactor-tracker.md`.
- `docs/domain.md`.
- `docs/spec/00_document_set_status_and_precedence.md` through Core 04.
- `internal/modules/auth`, `internal/platform/authn`, `internal/platform/enterpriseauth`, and `internal/app`.
- `tools/phase1_test_map.json`, `tools/phase11_test_map.json`, and their generated ledgers.

Scope constraints:

- Preserve route shapes, success and error envelopes, cookies, CSRF behavior, authorization outcomes, session lifecycle, idempotency semantics, audit safety, generated contracts, frontend callers, and harness accounting unless a later owner decision explicitly authorizes a behavior change.
- Do not hand-edit generated roots under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, or `packages/ui-contracts/src/generated/**`.
- Do not create a new auth-adjacent module, public API, migration, compatibility layer, or abstraction unless it materially improves maintainability, security, extensibility, or conceptual clarity.
- Do not preserve legacy behavior merely because it exists; retain behavior only when it protects future maintainability, public contract stability, security posture, or migration safety.

## 2. Completed remediation summary

RB-001 through RB-004 are no longer open remediation blockers for the same reasons recorded in the current tracker:

| ID | Prior concern | Current state | Residual follow-up |
| --- | --- | --- | --- |
| `RB-001` | Production OIDC/SAML verifier boundary was not implemented beyond deterministic fixtures. | Structurally remediated: `internal/platform/enterpriseauth` now contains production verifier adapters, provider-manifest loading/reconciliation, and startup claim handling through `internal/app/runtime.go`. | Production verifier failure mapping and evidence are still incomplete; track as `AUTH-NEXT-001` through `AUTH-NEXT-003`. |
| `RB-002` | `Service` used one broad private auth store seam. | Remediated: `Service` now holds handler-family ports such as `loginStore`, `sessionStore`, `credentialStore`, `accountStore`, `userAdminStore`, `deploymentAuditReader`, and `enterpriseStore`, all backed by `platform/authn.Store` at assembly. | Keep further port extraction demand-driven; track as `AUTH-NEXT-008`. |
| `RB-003` | Bootstrap-aware credential authentication was mixed into ordinary auth helper flow. | Remediated: `bootstrapCredentialAuthenticator` is now an auth-owned internal component with a narrow result type. | Keep the seam narrow and do not give bootstrap tokens session semantics; track as `AUTH-NEXT-006`. |
| `RB-004` | Administrative-audit ownership was ambiguous. | Resolved by boundary decision: deployment-scoped administrative audit read remains in auth/admin through `deploymentAuditReader`; incident membership audit remains outside auth unless Core 01/Core 04 widen ownership. | Watch projection completeness and owner boundaries; track as `AUTH-NEXT-009` and `AUTH-NEXT-010`. |

The completed remediation also established or preserved:

- `internal/platform/httpapi` owns generic public API error helpers.
- `internal/platform/httpauth` owns generic session, CSRF, idle-slide, and deployment-admin guard primitives.
- Deterministic Enterprise Authentication fixtures remain test-only.
- Auth remains a backend auth/account/deployment-admin HTTP facade, not workbook orchestration.

### 2026-07-09 implementation update

The 2026-07-09 remediation slice implemented the open-gap plan:

- Enterprise-auth production verifier failures now normalize to the closed Core 01 reason-code registries; production-only verifier/config diagnostics no longer escape as public protocol `reason_code` values.
- `contracts/errors/index.json` now includes the enterprise-auth reason registries, including `auth_binding_conflict`; generated Go and TypeScript contract artifacts were regenerated through `make generate`.
- Phase 11 metadata now names the full Enterprise Authentication claim gate, including `AC-433..AC-436`, and keeps the production profile unclaimed by default until that gate is satisfied.
- Phase 1 E-1-08 metadata and generated ledger now describe authenticated non-admin denial as `403 authorization_denied` with `required_capability=deployment_admin`.
- Generic auth compatibility wrappers for platform-owned helpers were de-exported; auth internals and tests now call `internal/platform/httpapi` and `internal/platform/httpauth` directly.
- No generic audit module was introduced, and `platform/authn.Store` remains the single concrete backing store behind narrow auth service ports.

Validation evidence from this slice:

- `make json-shape-check generated-artifact-policy-check generate-drift phase-ledger-drift phase-schedule-drift` passed.
- `make phase-slice PHASE=phase11` passed: `.cartulary/test-results/20260709T122049Z-p42164`.
- `make phase-slice PHASE=phase1` passed: `.cartulary/test-results/20260709T122142Z-p62997`.
- `make service-backed-slice PHASE=phase11` passed: `.cartulary/test-results/20260709T122330Z-p90821`.
- `make agent-finalize` passed: `.cartulary/test-results/20260709T122429Z-p7495`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- `make test-fast` passed: `.cartulary/test-results/20260709T122443Z-p9835`.
- `make check` initially exposed the new test-name and Go analyzer issues; after remediation, `make check` passed: `.cartulary/test-results/20260709T123156Z-p1570`.

## 3. Current module boundary assessment

Current classification:

- `internal/modules/auth` is an auth/account/deployment-admin HTTP facade.
- It owns local login, session inspection/logout, credential lifecycle, current-account profile/preferences, deployment-user administration, deployment-scoped administrative-audit read, Enterprise Authentication protocol routes, and Enterprise Authentication binding management.
- It does not own workbook rows, workbook query, projections, revisions, saved views, grid-adapter behavior, frontend shell/controller state, or incident membership audit.
- `Service` now uses handler-family ports backed by `platform/authn.Store`.
- Bootstrap credential auth now has `bootstrapCredentialAuthenticator`.
- `internal/platform/enterpriseauth` contains production OIDC/SAML verifiers and provider-manifest reconciliation.
- `internal/app/runtime.go` applies extension-claim state, reconciles enterprise provider manifests, and assembles routes before serving traffic.

Boundary assessment:

| Area | Current owner | Current posture |
| --- | --- | --- |
| Local auth/session/credential flows | `internal/modules/auth` plus `internal/platform/authn` and `internal/platform/httpauth` | Keep in auth facade; refine only through behavior-preserving seams. |
| Current-account profile/preferences | `internal/modules/auth` | Keep with auth/account facade; not deployment-admin cross-user access. |
| Deployment-user admin | `internal/modules/auth` | Keep with auth; security-sensitive due to credential, version, audit, and revocation effects. |
| Deployment administrative-audit read | `internal/modules/auth` | Keep in auth/admin unless owner docs widen audit ownership. |
| Incident membership audit | Incident modules/Core 01/Core 04 owner path | Keep outside auth; deployment_admin alone is insufficient. |
| Enterprise provider protocol and bindings | `internal/modules/auth` facade plus `internal/platform/enterpriseauth` verifier/platform boundary | Harden production contract mapping and evidence before broader claims. |
| Auth persistence | `internal/platform/authn` | Keep SQL and durable auth state there; avoid moving storage into handlers. |
| Frontend auth consumers | `apps/web` | Validate route-visible changes with frontend/browser checks; auth should not own UI state. |

## 4. Remediation disposition and residual watchpoints

The immediate 2026-07-09 gap families are closed. Future work should treat these areas as watchpoints, not as authorization to refactor broadly:

- Preserve closed enterprise-auth public reason-code registries unless Core 01 changes them.
- Keep Enterprise Authentication unclaimed by default until the full Core 04 AC-235 gate passes.
- Keep bootstrap-token behavior route-limited and separate from ordinary session auth.
- Keep deployment administrative audit in auth/admin and incident membership audit outside auth unless Core owners widen audit.
- Keep SQL and durable auth transaction behavior in `platform/authn`.

Current disposition:

| ID | Disposition |
| --- | --- |
| `AUTH-NEXT-001` | Closed: production OIDC/SAML/unconfigured verifier paths normalize to Core 01 closed reason codes. |
| `AUTH-NEXT-002` | Closed: production verifier negative-path tests and registry guards were added. |
| `AUTH-NEXT-003` | Closed: Phase 11 claim-gate metadata now names the full gate and remains unclaimed by default. |
| `AUTH-NEXT-004` | Closed: E-1-08 metadata and generated ledger now use authenticated `403 authorization_denied`. |
| `AUTH-NEXT-005` | Residual watchpoint: no broad handler split was required for this remediation. |
| `AUTH-NEXT-006` | Preserved: bootstrap credential auth remains internal and route-limited. |
| `AUTH-NEXT-007` | Closed: generic platform helper wrappers were de-exported after caller cleanup. |
| `AUTH-NEXT-008` | Preserved: one concrete `platform/authn.Store` remains behind narrow service ports. |
| `AUTH-NEXT-009` | Preserved: deployment audit remains in auth/admin; no audit module was introduced. |
| `AUTH-NEXT-010` | Residual watchpoint: add audit shape tests only when audit behavior changes. |
| `AUTH-NEXT-011` | Preserved: startup ordering was not changed by this slice. |
| `AUTH-NEXT-012` | Closed for this slice: validation commands and run roots are recorded above. |

Historical compact gap register retained for provenance:

| ID | Title | Current state | Recommended remediation | Owner area | Rationale | Long-term benefit | Compatibility or migration impact | Risks of leaving unresolved | Validation criteria | Suggested sequencing | Dependencies |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `AUTH-NEXT-001` | Enterprise-auth production verifier emits non-registry reason codes. | `ProductionOIDCVerifier`, `ProductionSAMLVerifier`, and unconfigured verifier paths can return `provider_response_rejected` reason codes not listed in Core 01, including config/internal verifier codes. | Map production OIDC/SAML failures onto Core 01's closed `provider_response_rejected` registry, or obtain an owner spec/contract update before emitting new codes. | Implementation, tests, specification if registry changes. | AC-293 requires enterprise-auth protocol reason codes to use the declared registries. | Keeps provider failures predictable, contract-safe, and frontend-safe. | Mapping to existing reason codes is behavior-visible but contract-aligning; widening the registry requires owner input and generated contract work. | Claimed Enterprise Authentication can fail conformance and expose undocumented client behavior. | Tests cover production verifier failure mappings and no unsupported reason code is reachable from `/api/v1/auth/*`. | First. | Core 01/Core 04 registry ownership; `AUTH-NEXT-002` depends on it. |
| `AUTH-NEXT-002` | Production verifier evidence is success-heavy. | `production_verifier_test.go` covers OIDC and SAML happy paths; Phase 11 integration uses deterministic verifiers. | Add negative-path tests for OIDC discovery, token, id-token, nonce, issuer/audience/signature-equivalent, and SAML parse/signature/expiry/config failures, proving closed registry use and no session issuance. | Tests, implementation. | Success-only tests do not protect security failure paths. | Future IdP changes become safer and claim evidence is less fixture-biased. | Test-only unless failure mapping changes. | Production verifier regressions can mint sessions or return non-contract errors. | `make phase-slice PHASE=phase11` passes and negative tests assert closed reason codes. | Immediately after `AUTH-NEXT-001`. | `AUTH-NEXT-001`. |
| `AUTH-NEXT-003` | Phase 11 claim posture remains easy to misread. | Production verifier code exists, while Phase 11 map and ledger still state Enterprise Authentication is unclaimed by default until claim gates pass. | Clarify tracker and evidence language: implementation exists, deterministic evidence exists, but production Enterprise Authentication remains unclaimed by default until AC-235 evidence gates pass. | Documentation, tests. | Avoids overclaiming extension conformance from deterministic fixture rows. | Keeps future release and conformance posture honest. | Documentation and manifest text only unless claim status changes. | Teams may treat deterministic fixtures as production interoperability proof. | Phase 11 map/ledger language remains consistent with Core 04 AC-235. | With contract hardening work. | Owner decision required only if profile claim status changes. |
| `AUTH-NEXT-004` | Phase 1 E-1-08 denial text drift. | Backend and Playwright code now assert authenticated non-admin deployment-admin denials as `403 authorization_denied`, while Phase 1 map/ledger text still says `401 session_required` for E-1-08. | Update `tools/phase1_test_map.json` owner metadata to `403 authorization_denied` with `required_capability=deployment_admin`, regenerate ledgers, and run drift checks. | Tests, documentation. | Evidence text should match implemented and owner-backed RB remediation. | Prevents future agents from restoring stale denial semantics. | Harness metadata and generated ledger update only. | Confusing evidence can cause incorrect test or implementation changes. | `make phase-ledger-drift` and `make phase-schedule-drift` pass after regeneration; E-1-08 text matches code. | Second workstream after security correction or as independent docs/test cleanup. | Current test code and Core 04 denial taxonomy. |
| `AUTH-NEXT-005` | Auth handler orchestration remains broad. | Handler methods still parse requests, authenticate, authorize, call stores, map errors, write envelopes, set cookies, and publish revocations. | Do not split routes yet; first define internal subfacades around session, credential, account, user-admin, audit-read, and enterprise-auth behavior without changing route contracts. | Implementation, tests. | Cohesive internal seams reduce future phase-growth risk without inventing public APIs. | Easier reasoning, smaller tests, clearer ownership. | Behavior-preserving if kept internal; no migration. | Route handlers remain hard to extend and easy to couple accidentally. | Phase 1/11 slices pass; no OpenAPI, error, cookie, or frontend route drift. | After contract/evidence cleanup. | Existing characterization in Phase 1 and Phase 11. |
| `AUTH-NEXT-006` | Bootstrap credential context should stay narrow. | `bootstrapCredentialAuthenticator` exists and remains tied to TOTP bootstrap routes. | Keep bootstrap auth separate from ordinary sessions; preserve one-auth-mode checks, route allowlist, consumed/expired rejection, and no session semantics in future edits. | Implementation, tests. | Bootstrap token scope is security-sensitive and should not become a general auth mode. | Makes future credential and MFA expansion safer. | Behavior-preserving; no migration. | Bootstrap tokens could accidentally authorize broader routes or collide with sessions. | Phase 1 bootstrap boundary tests pass, especially TOTP begin/complete and rejection on ordinary routes. | With cohesion refactor work. | Existing bootstrap tests and Core 04 credential requirements. |
| `AUTH-NEXT-007` | Auth compatibility wrappers are mostly local convenience. | `auth.APIError`, `ValidateCSRF`, `RequireDeploymentAdmin`, and helper wrappers remain after platform extraction; cross-package runtime callers mostly use platform helpers. | Retain wrappers only where they provide local value; consider de-exporting or narrowing auth-only helpers if cross-package callers are gone. | Implementation, tests. | Reduces confusing dual ownership between auth and platform. | Cleaner import boundaries and less accidental reuse of auth internals. | May require test helper updates; no wire migration. | Future code may import auth helpers instead of platform owners. | Import search shows no unwanted runtime dependency on auth compatibility wrappers; Phase 1 tests pass. | After `AUTH-NEXT-005`, or as a small independent cleanup. | Cross-package import inventory. |
| `AUTH-NEXT-008` | Persistence ports are narrower but still backed by one `authn.Store`. | `Service` uses family-specific ports, but `newService` wires all ports to one `platform/authn.Store`. | Keep SQL in `platform/authn`; extract new persistence ports only when concrete coupling or test burden appears. | Implementation. | Avoids abstraction churn and preserves the storage boundary. | Keeps auth storage centralized while letting handler seams stay narrow. | No migration unless future store split is owner-backed. | Premature store splitting can make migrations and transaction semantics harder. | No direct SQL appears in auth handlers; store tests remain stable. | Watchpoint, not immediate work. | Any future platform/authn storage plan. |
| `AUTH-NEXT-009` | Deployment audit read belongs in auth for now. | Auth owns `/api/v1/administrative-audit-events`; Core 01/Core 04 also define incident membership audit outside auth. | Do not create an audit module unless Core 01/Core 04 broaden ownership; keep incident membership audit separate under incident ownership. | Multiple, documentation. | Administrative audit is deployment-local auth/admin evidence, not incident row history. | Prevents a premature cross-cutting audit module. | No migration now. | Auth could absorb incident audit behavior or split too early. | Tracker and future plans preserve deployment-versus-incident audit separation. | Watchpoint with any audit route work. | Owner decision required for module split. |
| `AUTH-NEXT-010` | Administrative-audit projection completeness must be watched. | Auth deployment-audit route builds resources from `platform/authn` audit rows; Core 01/Core 04 require exact shapes, redaction, ordering, and separation from incident membership audit. | Before any audit refactor, confirm deployment route shape, redaction, ordering, cursor behavior, and incident-membership separation against AC-437 through AC-440. | Tests, implementation. | Audit mistakes are security-sensitive and hard to repair after exposure. | Safer admin observability and conformance. | Test additions likely; no migration unless storage shape changes. | Secret-bearing values, wrong-scope audit records, or stale action codes can leak. | Phase 1/support audit tests cover redaction and authorization; add targeted tests if shape changes. | Before any audit handler or store refactor. | Core 01 §3.3.5.1A and Core 04 AC-437..AC-440. |
| `AUTH-NEXT-011` | Extension-profile runtime assembly and manifest reconciliation are coupled. | `internal/app/runtime.go` applies config claims, bootstraps telemetry, reconciles enterprise provider manifest, and passes extension profiles into route assembly. | Preserve order: validate config, apply claim state, reconcile provider manifest, select production verifiers for claimed profile, then serve routes. | Implementation, tests. | Startup order is part of the security boundary for claimed Enterprise Authentication. | Reduces risk of serving claimed enterprise routes with stale or unreconciled provider state. | Behavior-preserving unless claim startup rules change. | Claimed deployments could serve routes before provider definitions are safe. | Runtime/startup tests prove claimed/unclaimed provider manifest behavior and route setup order. | With enterprise-auth hardening. | Core 04 provider-manifest startup requirements. |
| `AUTH-NEXT-012` | Retained validation evidence should be refreshed after implementation slices. | Current tracker records prior retained evidence, but each future implementation slice needs fresh handoff evidence. | Run `make agent-finalize` before broad gates, then record result roots for `make test-fast` or `make check` as applicable. | Tests, documentation. | Keeps handoffs reproducible and avoids stale evidence claims. | Easier continuation and review. | Tool-managed artifacts may update; do not hand-edit them. | Future agents may rely on stale green runs. | Final handoff names commands, result roots, failures, skipped checks, and whether retained-run maintenance was used. | Final workstream after any implementation slice. | Any implementation slice. |

## 5. Security-sensitive follow-ups

Security-sensitive work should be sequenced before cohesion cleanup when both touch the same files.

Priority follow-ups:

- Normalize Enterprise Authentication production failure reason codes before treating production verifier behavior as claim-ready.
- Expand production verifier negative tests before broadening claim posture or relying on production IdP behavior.
- Keep bootstrap-token handling as credential-setup-only, not a general authenticated principal.
- Preserve authenticated non-admin deployment-admin denials as `403 authorization_denied` with `required_capability=deployment_admin`; session failures remain `401 session_required`.
- Preserve CSRF fail-closed behavior for cookie-authenticated mutators.
- Preserve audit and idempotency forbidden-secret rules for passwords, TOTP material, bootstrap/session tokens, provider assertions, provider tokens, SAML responses, ID tokens, access tokens, recovery keys, object-store credentials, DSNs, object keys, and storage secrets.

## 6. Enterprise-auth production hardening follow-ups

Current production posture:

- Enterprise Authentication routes exist behind extension-profile claim behavior.
- Provider manifest validation and reconciliation run before route serving when the profile is claimed.
- `ProductionOIDCVerifier` and `ProductionSAMLVerifier` exist in `internal/platform/enterpriseauth`.
- Deterministic Phase 11 integration tests still prove route behavior through test-only verifier overrides.
- The Phase 11 ledger states production Enterprise Authentication remains unclaimed by default.

Hardening sequence:

1. Fix or owner-resolve unsupported production reason codes (`AUTH-NEXT-001`).
2. Add negative-path production verifier tests (`AUTH-NEXT-002`).
3. Clarify claim posture in map/ledger/tracker language (`AUTH-NEXT-003`).
4. Preserve startup ordering in `internal/app/runtime.go` (`AUTH-NEXT-011`).

Do not add runtime provider-definition mutation routes, browser provider editors, JIT user creation, SCIM, group-to-role mapping, enterprise-only users, self-service linking, or passkeys under this tracker.

## 7. Bootstrap/auth-context follow-ups

Bootstrap context rules to preserve:

- Bootstrap tokens are accepted only where Core 01/Core 04 allow credential setup.
- Bootstrap token use must not mint a session directly.
- Session cookie plus bootstrap bearer token remains an invalid mixed auth mode.
- Consumed, expired, wrong-subject, or wrong-route bootstrap tokens fail closed.
- Ordinary `httpauth.AuthenticateRequest` remains the generic session boundary for peer modules.

Recommended future direction:

- Keep `bootstrapCredentialAuthenticator` internal to auth.
- Avoid moving bootstrap logic into `platform/httpauth` unless another module has a real, owner-backed need for bootstrap semantics.
- If local credential lifecycle grows, expose a small internal credential-context helper rather than passing raw token/session branches through handler code.

## 8. Store-port and persistence-boundary follow-ups

Current state:

- Auth handler-family ports are narrower than the prior broad store seam.
- `platform/authn.Store` remains the backing implementation for auth persistence.
- SQL and durable auth state remain in `internal/platform/authn`.
- Auth handlers do not directly own authored SQL.

Future extraction policy:

- Keep SQL and durable auth state in `platform/authn` unless a separate platform/authn storage plan authorizes a move.
- Add new ports only when they reduce actual coupling, test burden, or transaction ambiguity.
- Do not split ports just to mirror files.
- Preserve route-scoped idempotency and audit transaction semantics before any store reshaping.

## 9. Administrative-audit ownership watchpoints

Current owner split:

- `GET /api/v1/administrative-audit-events` is deployment-scoped and belongs in the auth/admin facade today.
- `GET /api/v1/incidents/{incident_id}/membership-audit-events` is incident-scoped and must not be authorized by `deployment_admin` alone.
- Administrative audit is not incident revision history, not workbook row history, and not incident portability source state.

Watchpoints:

- Do not create a generic audit module without a Core 01/Core 04 owner decision.
- Do not merge deployment audit and incident membership audit under auth as an implementation convenience.
- Do not emit incident membership audit records through the deployment route.
- Do not expose deployment-local administrative audit through incident portability bundles.
- Before any audit refactor, re-check AC-437 through AC-440 and the route-level authorization precedence in Core 04 REQ-04-123.

## 10. Test and evidence posture

Current posture after the 2026-07-09 remediation slice:

| Gap | Evidence | Required action |
| --- | --- | --- |
| Enterprise production verifier failure evidence. | `internal/platform/enterpriseauth/production_verifier_test.go` now covers success and negative-path closed-registry mappings. | Keep new failure classes inside the closed registry or update Core 01/contracts first. |
| Phase 11 claim posture. | Phase 11 map/ledger names the full claim gate and keeps production Enterprise Authentication unclaimed by default. | Do not flip the claim until the full AC-235 gate passes. |
| Phase 1 E-1-08 text drift. | Owner metadata and generated ledger now state authenticated non-admin `403 authorization_denied` with `required_capability=deployment_admin`. | Preserve the authentication-versus-authorization distinction. |
| Audit projection shape completeness needs watch coverage before refactor. | Current auth tests cover audit safety and authorization, but future shape changes may need tighter AC-437..AC-440 characterization. | Add targeted tests only when audit handler/store behavior changes. |
| Compatibility wrapper exports. | Generic platform-owned auth wrappers were removed; auth internals/tests call `platform/httpapi` and `platform/httpauth` directly. | Do not reintroduce auth-owned facades for generic platform behavior. |

Primary validation surfaces:

- Phase 1 covers local authentication, session lifecycle, TOTP bootstrap/replacement, deployment-local user administration, and deployment-admin credential actions.
- Phase 11 covers deterministic Enterprise Authentication route behavior, binding lifecycle, selected extension profiles, and common-job substrate.
- Production Enterprise Authentication claim evidence remains separate from deterministic fixture evidence.

## 11. Completed workstreams and sequencing

| Workstream | Items | Goal | Validation |
| --- | --- | --- | --- |
| WS-1 contract/security correction | `AUTH-NEXT-001`, `AUTH-NEXT-002`, `AUTH-NEXT-003` | Completed: production Enterprise Authentication failure behavior and evidence are contract-safe. | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11`; `make check`. |
| WS-2 evidence metadata cleanup | `AUTH-NEXT-004` | Completed: Phase 1 evidence text matches current denial taxonomy. | `make phase-ledger-drift`; `make phase-schedule-drift`; `make phase-slice PHASE=phase1`; `make check`. |
| WS-3 boundary cleanup | `AUTH-NEXT-006`, `AUTH-NEXT-007` | Completed for generic wrappers; bootstrap semantics preserved. | `make phase-slice PHASE=phase1`; `make test-fast`; `make check`. |
| WS-4 persistence/audit watchpoints | `AUTH-NEXT-008`, `AUTH-NEXT-009`, `AUTH-NEXT-010`, `AUTH-NEXT-011` | Preserved: no premature audit module, store split, or startup-order change. | Covered by phase slices and `make check`; add targeted tests only when behavior changes. |
| WS-5 final validation/handoff | `AUTH-NEXT-012` | Completed for this slice. | `make agent-finalize`, `make test-fast`, and `make check` run roots are recorded in §2. |

Validation defaults:

- Docs-only tracker: `make lint-markdown`.
- Phase 1/auth changes: `make phase-slice PHASE=phase1`.
- Phase 1 store, audit, revocation, or browser-visible route behavior: `make service-backed-slice PHASE=phase1`.
- Enterprise-auth changes: `make phase-slice PHASE=phase11`.
- Enterprise callback, binding, or runtime behavior: `make service-backed-slice PHASE=phase11`.
- Contract or map updates: `make generated-artifact-policy-check`, `make json-shape-check`, `make phase-ledger-drift`, and `make phase-schedule-drift`.
- Broad handoff: `make agent-finalize`, then `make test-fast`; run `make check` when risk warrants it.

## 12. Resolved owner decisions for this slice

| ID | Decision | Notes |
| --- | --- | --- |
| `AUTH-OQ-001` | Do not widen Core 01 enterprise-auth provider-response registries in this slice. | Production verifier diagnostics normalize to the existing closed registries; contracts now carry those registries. |
| `AUTH-OQ-002` | Keep Enterprise Authentication unclaimed by default. | The Phase 11 gate now explicitly requires Base profile evidence, listed enterprise-auth ACs including `AC-433..AC-436`, production verifier evidence without deterministic overrides, startup manifest evidence, and synchronized contracts. |
| `AUTH-OQ-003` | De-export generic auth compatibility wrappers after caller cleanup. | Auth now calls `platform/httpapi` and `platform/httpauth` directly for platform-owned helpers. |
| `AUTH-OQ-004` | Do not create a generic audit module. | Deployment administrative audit remains in auth/admin; incident membership audit remains outside auth. |
| `AUTH-OQ-005` | Do not split `platform/authn.Store` now. | One concrete backing store remains behind narrow service ports. |

## 13. Exit criteria status

The 2026-07-09 remediation slice is complete because:

- Production verifier reason codes normalize to the existing registry.
- Negative-path verifier evidence exists for production OIDC/SAML behavior.
- Phase 11 claim posture remains clear and does not overclaim deterministic fixture evidence.
- Phase 1 E-1-08 metadata drift is resolved through owner manifest input and regenerated ledger output.
- The wrapper cleanup preserved route shape, envelopes, cookies, CSRF behavior, authorization outcomes, session lifecycle, idempotency, audit safety, generated contracts, and frontend behavior.
- No new module, public API, migration, audit module, store split, or runtime provider mutation surface was introduced.
- Validation commands and result roots are recorded, including `make agent-finalize` retained-run status.

## 14. Handoff notes for the next implementer

Start here:

1. Read this tracker, then the current tracker it supersedes.
2. Confirm current source with `rg --files internal/modules/auth internal/platform/authn internal/platform/enterpriseauth internal/app`.
3. Re-check Core 01 before changing enterprise-auth public reason-code registries; do not emit new provider-response reasons from implementation alone.
4. If editing phase metadata, update the `tools/phase*_test_map.json` owner input first and regenerate ledgers rather than editing generated ledgers directly.
5. Choose the narrowest Make target from `make task-guide ROLE=feature-dev PHASE=phase1` or `make task-guide ROLE=feature-dev PHASE=phase11`.
6. Run `make agent-finalize` before broad end-of-run verification if implementation files changed.

Current known validation posture:

- Phase 1 and Phase 11 slices passed after the remediation.
- `make test-fast` and final `make check` passed; run roots are recorded in §2.
- For a docs-only update to this tracker, run `make lint-markdown`; for auth or enterprise-auth changes, choose the narrowest phase slice first and broaden only when touched behavior warrants it.

Files changed by the 2026-07-09 remediation include:

- `docs/handoffs/auth-module-module-refactor-tracker-next.md`.
- Core 01/Core 04 spec text for enterprise-auth reason-code diagnostics.
- `contracts/errors/index.json` and generated contract artifacts.
- Phase 1/Phase 11 owner manifests and generated ledgers/schedule artifacts.
- `internal/platform/enterpriseauth` verifier implementation and tests.
- `internal/modules/auth` wrapper cleanup and enterprise-auth contract guard test.

Files intentionally not changed:

- No database migrations.
- No generic audit module.
- No split of the concrete `platform/authn.Store`.
- No runtime enterprise provider mutation surface.
- Generated phase ledgers or schedules.
- `go.sum`, `pnpm-lock.yaml`, or tool-managed dependency/install artifacts.
