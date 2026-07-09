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

## 4. Remaining structural gaps

The next iteration should focus on these gap families:

- Production Enterprise Authentication error-contract hardening.
- Production Enterprise Authentication negative-path evidence.
- Phase 1 and Phase 11 evidence-text clarity.
- Auth facade cohesion without route or wire drift.
- Bootstrap/session context simplification without widening bootstrap semantics.
- Demand-driven persistence-port refinement.
- Administrative-audit projection and owner-boundary watchpoints.
- Final validation evidence retention after any implementation slice.

Compact gap register:

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

## 10. Test and evidence gaps

Known gaps and drift:

| Gap | Evidence | Required action |
| --- | --- | --- |
| Enterprise production verifier negative evidence is thin. | `internal/platform/enterpriseauth/production_verifier_test.go` covers success paths. | Add closed-registry negative tests after `AUTH-NEXT-001`. |
| Phase 11 claim posture can be misread. | Phase 11 map/ledger says unclaimed by default while production verifier code exists. | Keep text explicit: code exists, claim remains gated. |
| Phase 1 E-1-08 text drift. | `tools/phase1_test_map.json` and generated ledger say `401 session_required`, while current test code asserts `403 authorization_denied`. | Update owner manifest metadata, regenerate ledger, and run drift checks. |
| Audit projection shape completeness needs watch coverage before refactor. | Current auth tests cover audit safety and authorization, but future shape changes may need tighter AC-437..AC-440 characterization. | Add targeted tests only when audit handler/store behavior changes. |
| Compatibility wrapper exports may obscure ownership. | Runtime callers mostly use `platform/httpapi` and `platform/httpauth`; auth keeps local wrappers. | Inventory before narrowing exports; keep test support source-compatible unless cleanup is authorized. |

Primary validation surfaces:

- Phase 1 covers local authentication, session lifecycle, TOTP bootstrap/replacement, deployment-local user administration, and deployment-admin credential actions.
- Phase 11 covers deterministic Enterprise Authentication route behavior, binding lifecycle, selected extension profiles, and common-job substrate.
- Production Enterprise Authentication claim evidence remains separate from deterministic fixture evidence.

## 11. Proposed workstreams and sequencing

| Workstream | Items | Goal | Validation |
| --- | --- | --- | --- |
| WS-1 contract/security correction | `AUTH-NEXT-001`, `AUTH-NEXT-002`, `AUTH-NEXT-003` | Bring production Enterprise Authentication failure behavior and evidence into contract-safe shape. | `make phase-slice PHASE=phase11`; use `make service-backed-slice PHASE=phase11` if callback, binding, or runtime behavior changes. |
| WS-2 evidence metadata cleanup | `AUTH-NEXT-004` | Align Phase 1 evidence text with current denial taxonomy. | `make phase-ledger-drift`; `make phase-schedule-drift`; run `make phase-slice PHASE=phase1` if evidence claims or tests changed. |
| WS-3 cohesion-only refactors | `AUTH-NEXT-005`, `AUTH-NEXT-006`, `AUTH-NEXT-007` | Improve internal auth seams without route, envelope, cookie, CSRF, audit, or generated-contract drift. | `make phase-slice PHASE=phase1`; add `make phase-slice PHASE=phase11` if enterprise files are touched. |
| WS-4 persistence/audit watchpoints | `AUTH-NEXT-008`, `AUTH-NEXT-009`, `AUTH-NEXT-010`, `AUTH-NEXT-011` | Keep storage, audit, and runtime assembly boundaries explicit while future work grows. | Targeted phase slices based on touched area; add startup/runtime tests for claim assembly changes. |
| WS-5 final validation/handoff | `AUTH-NEXT-012` | Refresh retained evidence and record a clean continuation point. | `make agent-finalize`, then `make test-fast`; run `make check` for auth routes, enterprise auth, persistence, frontend-visible behavior, or generated contracts. |

Validation defaults:

- Docs-only tracker: `make lint-markdown`.
- Phase 1/auth changes: `make phase-slice PHASE=phase1`.
- Phase 1 store, audit, revocation, or browser-visible route behavior: `make service-backed-slice PHASE=phase1`.
- Enterprise-auth changes: `make phase-slice PHASE=phase11`.
- Enterprise callback, binding, or runtime behavior: `make service-backed-slice PHASE=phase11`.
- Contract or map updates: `make generated-artifact-policy-check`, `make json-shape-check`, `make phase-ledger-drift`, and `make phase-schedule-drift`.
- Broad handoff: `make agent-finalize`, then `make test-fast`; run `make check` when risk warrants it.

## 12. Open questions requiring owner decision

| ID | Question | Why it matters | Default until decided |
| --- | --- | --- | --- |
| `AUTH-OQ-001` | Should Core 01 widen enterprise-auth provider-response reason-code registries for production-only provider/config failures? | Current production verifier emits reason codes outside the closed registry, but widening affects public contracts. | Normalize to existing Core 01 reason codes. |
| `AUTH-OQ-002` | What exact evidence gate changes Enterprise Authentication from unclaimed-by-default to claimed production profile? | Phase 11 has deterministic evidence and production verifier code, but AC-235 claim posture remains gated. | Keep unclaimed by default. |
| `AUTH-OQ-003` | Should auth compatibility wrappers be de-exported after cross-package cleanup? | Exported helpers can imply auth owns platform behavior. | Keep wrappers until an implementation slice proves no source-compatible need remains. |
| `AUTH-OQ-004` | Should deployment audit and incident membership audit ever move into a new audit module? | A generic audit module can clarify long-term ownership, but it can also over-centralize unrelated security rules. | Keep deployment audit in auth/admin and membership audit outside auth. |
| `AUTH-OQ-005` | Should `platform/authn.Store` be split into smaller concrete stores? | The current backing store centralizes transaction and schema behavior; splitting may help only with concrete coupling. | Keep one backing store with narrow service ports. |

## 13. Exit criteria for the next refactor iteration

The next iteration is complete when:

- `AUTH-NEXT-001` has either normalized production verifier reason codes to the existing registry or has an accepted owner decision to widen the registry.
- `AUTH-NEXT-002` has negative-path verifier evidence for production OIDC/SAML behavior.
- `AUTH-NEXT-003` keeps Phase 11 claim posture clear and does not overclaim deterministic fixture evidence.
- `AUTH-NEXT-004` resolves Phase 1 E-1-08 metadata drift through owner manifest input and regenerated ledger output.
- Any cohesion refactor preserves route shape, envelopes, cookies, CSRF behavior, authorization outcomes, session lifecycle, idempotency, audit safety, generated contracts, and frontend behavior.
- No new module, public API, migration, generated-contract change, or compatibility layer was introduced without owner-backed rationale.
- Validation commands and result roots are recorded, including whether `make agent-finalize` retained-run maintenance was used.
- Any skipped checks are explicitly justified.

## 14. Handoff notes for the next implementer

Start here:

1. Read this tracker, then the current tracker it supersedes.
2. Confirm current source with `rg --files internal/modules/auth internal/platform/authn internal/platform/enterpriseauth internal/app`.
3. Re-check enterprise-auth reason-code registry in Core 01 before touching verifier failure mapping.
4. If editing Phase 1 metadata, update `tools/phase1_test_map.json` owner input first and regenerate ledgers rather than editing generated ledgers directly.
5. Choose the narrowest Make target from `make task-guide ROLE=feature-dev PHASE=phase1` or `make task-guide ROLE=feature-dev PHASE=phase11`.
6. Run `make agent-finalize` before broad end-of-run verification if implementation files changed.

Current known validation posture:

- Phase 1 and Phase 11 task guides identify the primary verification surfaces.
- This tracker is documentation-only; product/auth tests are not required for creating it.
- For a docs-only update to this tracker, run `make lint-markdown`.

Files expected to change when creating this tracker:

- `docs/handoffs/auth-module-module-refactor-tracker-next.md`.

Files not to edit for this tracker:

- Existing auth implementation files.
- Generated contract roots.
- Generated phase ledgers or schedules.
- `go.sum`, `pnpm-lock.yaml`, or tool-managed dependency/install artifacts.
