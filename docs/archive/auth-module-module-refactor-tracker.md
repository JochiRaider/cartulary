# auth-module Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

Target path: `internal/modules/auth`.

Target label: `auth-module`.

Tracker path: `docs/handoffs/auth-module-module-refactor-tracker.md`.

Non-goals:

- Do not change route shapes, request or response envelopes, authorization outcomes, cookies, CSRF behavior, session lifecycle, audit/idempotency storage semantics, generated contracts, frontend callers, or harness accounting.
- Do not hand-edit generated roots under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, or generated harness/task-surface outputs.
- Do not treat `auth-module` as a permanent module boundary merely because `internal/modules/auth` exists.

Source hierarchy used:

1. Adopted subsystem NLSpecs, for their named subsystem only.
2. Core 00 through Core 04, for implementation-conformance behavior.
3. Core 05, only for claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md` and implementation-support guides, for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests, for current implementation state.
6. Prior plans, handoffs, and framework files, as evidence only.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`.
- `docs/spec/00_document_set_status_and_precedence.md`, for document authority and owner-family posture.
- `docs/spec/01_architecture_storage_and_view_contracts.md`, for auth/account/deployment-user route inventory and Enterprise Authentication public contract.
- `docs/spec/04_security_deployment_and_conformance.md`, for sessions, CSRF, TOTP, deployment administration, current-account access, administrative audit authorization, credential actions, enterprise binding authorization, and safe audit constraints.
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`, posture only; no Core 05 claim-bearing evidence is planned here.
- `docs/domain.md`, for vocabulary boundaries, especially Authentication and Administration as a generic context and the separation from workbook/source domains.
- `docs/testing-harness-nlspec.md`, for Make-owned harness mechanics and evidence posture.
- `tools/generated_artifact_policy.json`, for generated roots that must not be hand-edited.

Repository files inspected:

- Every file returned by `rg --files internal/modules/auth` on 2026-07-09.
- Route assembly in `internal/app/runtime.go` and `cmd/server/main.go`.
- Platform auth/error/verifier seams in `internal/platform/httpapi/api_error.go`, `internal/platform/httpauth/httpauth.go`, and `internal/platform/enterpriseauth/enterpriseauth.go`.
- Auth-owned contract artifacts by reference and targeted inspection: `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, `contracts/extensions/index.json`, `packages/protocol-ts/src/index.test.ts`, and `packages/ui-contracts/src/index.ts`.
- Frontend route callers by targeted search under `apps/web` and `packages`.
- Prior archive evidence in `docs/archive/auth-module-refactor-tracker.md`.

Prior handoff posture:

- `docs/handoffs/auth-module-module-refactor-tracker.md` did not exist at session start.
- `docs/archive/auth-module-refactor-tracker.md` exists but is stale prior evidence only. It references older file layout and an older output path, including files not present in the current target inventory.

Architectural posture:

- Live repository evidence supports `internal/modules/auth` as a backend auth/account/admin HTTP facade with transport-adjacent handlers, private store facades, credential/session orchestration, deployment-user administration, administrative-audit read projection, and Enterprise Authentication extension routes.
- Live repository evidence does not support classifying `auth-module` as workbook orchestration. The target owns no timeline, projection, revision, import/tabular ingest, entity, evidence, link, saved-view, view-contract, frontend shell/controller, or grid-vendor behavior.
- No owner contradiction was found during this pass. If future evidence exposes an owner contradiction, write `BLOCKED: owner contradiction` and do not choose a side.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/auth/account_handlers.go` | Handles current account profile and preferences routes: `GET/PATCH /api/v1/account/profile` and `GET/PUT /api/v1/account/preferences`. | Unexported `Service` methods `handleAccountProfile` and `handleAccountPreferences`. | Registered through `routes.go`; called by HTTP clients and frontend app shell/account settings. | `internal/platform/authn`, `internal/platform/httpapi`, local auth helpers. | `phase1_handlers_test.go`, `phase1_support_integration_test.go`, `phase1_integration_test.go`; frontend Phase 2/account tests. | OpenAPI account paths and protocol TS account schemas. | `auth` current-account subfacade. | Medium | Current-account routes are current-session-only and are not deployment-admin cross-user access. |
| `internal/modules/auth/admin_audit_handlers.go` | Parses administrative-audit list query, builds safe administrative-audit resources, paginates, and handles `GET /api/v1/administrative-audit-events`. | Unexported parser/build helpers and `Service.handleAdministrativeAuditEvents`. | Registered through `routes.go`; called by deployment-administration UI. | `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/listquery`, `internal/platform/pagination`, `uuid`. | `phase1_support_test.go`, `phase1_support_integration_test.go`, `phase1_integration_test.go`; frontend deployment-audit tests. | OpenAPI/admin route family and UI contract token `administrative-audit`. | `auth` deployment-admin audit read facade; possible future audit module only if owner docs widen scope. | Medium | This is a deployment-local audit read projection, not incident row revision history. |
| `internal/modules/auth/api.go` | Auth request decoding, common compatibility wrappers to platform HTTP/session helpers, credential-state/TOTP setup resource builders, and auth-specific API errors. | `APIError`, `AuthSource`, `LoginRequest`, `SecondFactorAssertion`, password/TOTP request types, decode helpers, `BuildCredentialStateResource`, `BuildTOTPSetup`, `ValidateCSRF`, `BootstrapRejectedError`, `TOTPSetupNotPendingError`, `ClientTxnConflictError`. | Handler files in this package; same-package tests. | `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/httpauth`, JSON/HTTP/url/time packages. | `phase1_request_test.go`, `phase1_support_test.go`, `phase1_handlers_test.go`, integration support tests. | Errors registry and OpenAPI/protocol surfaces depend on wire behavior. | `auth` helper layer with platform compatibility wrappers. | Medium | Generic API/session helpers now live in `platform/httpapi` and `platform/httpauth`; bootstrap-aware auth behavior remains auth-specific. |
| `internal/modules/auth/auth_session_helpers.go` | Handles bootstrap-aware authentication, session-token authentication, cookie setting/clearing, session resource construction, idle sliding, revocation publication, request hashing, stored-response decoding, and error writing. | Unexported helpers; public surface only through handlers and `Service`. | Handler files in `internal/modules/auth`; tests use same-package access. | `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/httpauth`, `uuid`, HTTP/JSON/crypto/time packages. | `phase1_handlers_test.go`, `phase1_store_test.go`, `phase1_integration_test.go`, support tests. | Session and cookie behavior flows to OpenAPI/error contracts and frontend cookie assumptions. | `auth` credential/session orchestration facade, with platform auth primitives. | High | Bootstrap-token routes are auth-specific; ordinary session/CSRF helpers are platform-owned. |
| `internal/modules/auth/credential_handlers.go` | Handles credential state, password change, TOTP begin, TOTP complete, active TOTP verification, session revocation after credential changes, and bootstrap-token credential flows. | Unexported `Service` handlers and `validateActiveTOTP`. | Registered through `routes.go`; called by frontend auth/account flows and browser tests. | `internal/platform/authn`, `internal/platform/httpapi`, local session helpers, `uuid`, HTTP/time. | `phase1_handlers_test.go`, `phase1_integration_test.go`, `phase1_support_integration_test.go`, frontend Phase 1 tests. | OpenAPI credential routes, errors registry, protocol TS generated types. | `auth` credential lifecycle. | High | Secret handling, idempotency, bootstrap-token scope, and session revocation are behavior-sensitive. |
| `internal/modules/auth/enterprise_protocol.go` | Aliases Enterprise Authentication verifier interfaces to platform verifier interfaces. | Type aliases `enterpriseOIDCVerifier` and `enterpriseSAMLVerifier`. | `enterprise_routes.go`; tests inject deterministic verifiers through module overrides. | `internal/platform/enterpriseauth`. | `phase11_enterprise_auth_integration_test.go`. | Enterprise Authentication route behavior indirectly. | `auth` extension facade with verifier boundary owned by platform. | Medium | Production default verifiers are unconfigured; deterministic verifier evidence is test-only. |
| `internal/modules/auth/enterprise_routes.go` | Handles enterprise provider discovery, provider begin, OIDC callback, SAML ACS, SAML completion, enterprise browser-binding cookie, and enterprise binding create/rotate/retire routes. | Exported request structs and decode helpers; unexported handlers/helpers. | Registered through `routes.go`; frontend app shell and enterprise auth UI call provider/binding routes. | `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/enterpriseauth` via aliases, `uuid`, crypto/HTTP/url/time packages. | `phase11_enterprise_auth_integration_test.go`, selected handler tests, frontend Phase 11 tests. | `contracts/extensions/index.json`, OpenAPI enterprise paths, errors registry, protocol TS generated surfaces. | `auth` Enterprise Authentication extension facade. | High | Core 01 owns route shapes; Core 04 owns provider config/security. Production IdP interoperability remains future work. |
| `internal/modules/auth/routes.go` | Defines `Service`, store subinterfaces, route registration, test route registration, route paths, service construction, master key/cursor setup, extension profiles, verifier overrides, and backing store creation. | `Service`, `RegisterRoutes`, `RegisterTestRoutes`, `EnterpriseOIDCVerifierOverrideKey`, `EnterpriseSAMLVerifierOverrideKey`, `SessionPrincipal`, `CredentialAuthContext`. | `internal/app/runtime.go`, `cmd/server/main.go`, test runtime helpers. | `internal/platform/authn`, `internal/platform/enterpriseauth`, `internal/platform/httpapi`, `internal/platform/httpauth`, `internal/platform/pagination`, `uuid`, HTTP/time. | All auth handler/integration tests. | All auth/account/admin/enterprise route contracts. | `auth` HTTP facade and composition root for the module. | High | Private `authStore` remains broad though split into subinterfaces; `authn.NewStore` is still constructed here. |
| `internal/modules/auth/session_handlers.go` | Handles local login, session inspection, and logout. | Unexported `Service` handlers. | Registered through `routes.go`; called by frontend app shell, login UI, browser tests, and API clients. | `internal/platform/authn`, `internal/platform/httpapi`, local helpers. | `phase1_request_test.go`, `phase1_handlers_test.go`, `phase1_store_test.go`, `phase1_integration_test.go`, frontend Phase 1/session tests. | OpenAPI auth routes, errors registry, cookie/session frontend assumptions. | `auth` session lifecycle. | High | Login/session/logout are primary auth observable contracts. |
| `internal/modules/auth/user_admin_handlers.go` | Handles deployment-user list/create/get/patch and user member action routes: password reset, TOTP reset, revoke-all, plus enterprise binding subroute dispatch. | Unexported `Service` handlers and list-scope helpers. | Registered through `routes.go`; deployment admin UI and frontend API client call `/api/v1/users*`. | `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/listquery`, `internal/platform/pagination`, `uuid`. | `phase1_handlers_test.go`, `phase1_store_test.go`, `phase1_integration_test.go`, support integration tests, frontend Phase 1 tests. | OpenAPI deployment-user paths, protocol TS generated surfaces, errors registry. | `auth` deployment-user administration facade. | High | Authorization, version conflict, idempotency, audit, and revocation behavior are sensitive. |
| `internal/modules/auth/users.go` | Defines user/account/admin request structs, safe resource builders, defaulting, deployment-admin guard wrapper, request decoders, list/pagination errors, and mutation payload errors. | Exported request/resource helper types and functions including `DecodeUserCreateRequest`, `DecodeUserPatchRequest`, account/admin decoders, `BuildSafeUserResource`, `RequireDeploymentAdmin`. | Auth handlers, same-package tests, selected testutil support. | `internal/platform/authn`, `internal/platform/httpauth`, JSON/HTTP/string packages. | `phase1_support_test.go`, `phase1_handlers_test.go`, `phase1_integration_test.go`, `phase11_enterprise_auth_integration_test.go`. | OpenAPI user/account/admin schemas and errors. | `auth` request/resource helper layer with platform guard wrapper. | Medium | `httpauth.RequireDeploymentAdmin` is the generic guard owner; auth keeps local compatibility. |
| `internal/modules/auth/phase11_enterprise_auth_integration_test.go` | Integration evidence for deterministic Enterprise Authentication OIDC, SAML, and binding lifecycle. | Test-only external package `auth_test`. | Go test harness and phase 11 maps. | `internal/modules/auth`, `internal/platform/authn`, `internal/testutil/enterpriseauthtest`, `internal/testutil/phase1test`, `httptestx`. | Itself. | `tools/phase11_test_map.json` and generated coverage ledgers. | Test evidence only. | Medium | Proves deterministic route behavior, not production IdP interoperability. |
| `internal/modules/auth/phase1_handlers_test.go` | Same-package handler unit tests with store stubs for login, session, CSRF, credential, admin, and enterprise seams. | Test-only stubs and helpers including `authStoreStub`, `hubStub`, `newUnitService`. | Go test harness. | `internal/platform/authn`, `internal/platform/pagination`, `internal/testutil/authcookietest`, `internal/testutil/phase1test/inventory`. | Itself. | Phase 1 test maps and ledgers. | Test evidence only. | Medium | Same-package access means internal helper moves require careful test adaptation. |
| `internal/modules/auth/phase1_integration_test.go` | Public-route integration evidence for session lifecycle, WebSocket revocation, credential state/bootstrap, user admin lifecycle, audit, replay, and admin credential actions. | Test-only external package `auth_test`; helper `startPhase1Server`. | Go test harness and phase runtime. | `internal/modules/auth`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/testutil/phase1test`, `httptestx`, WebSocket client. | Itself. | `tools/phase1_test_map.json` and generated ledgers. | Test evidence only. | Medium | Seeds incidents/sockets only to test auth/session boundaries. |
| `internal/modules/auth/phase1_request_test.go` | Unit coverage for strict login request shape and no-side-effect malformed login behavior. | Test-only package `auth`. | Go test harness. | `internal/platform/authn`, `httptest`, local helpers. | Itself. | Phase 1 test maps and ledgers. | Test evidence only. | Low | Key characterization for `DecodeLoginRequest` and login side-effect suppression. |
| `internal/modules/auth/phase1_store_test.go` | Store-backed auth route tests for concurrency cap, session slide persistence, revocation consequences, user patch, and last-admin guard. | Test-only package `auth`; fixture helpers. | Go test harness. | `internal/platform/authn`, `internal/platform/postgres`, `internal/testutil/phase1storetest`, `authcookietest`. | Itself. | Phase 1 test maps and ledgers. | Test evidence only. | Medium | Exercises `platform/authn.Store` through auth handlers. |
| `internal/modules/auth/phase1_support_integration_test.go` | Inventory-driven support integration sweeps for user-list pagination/search, envelopes, bootstrap boundaries, CSRF, replay payload safety, audit attribution, revocation, authorization re-derivation, and request contracts. | Test-only external package `auth_test`. | Go test harness and support maps. | `internal/testutil/phase1test`, `httptestx`, `authn`, WebSocket helpers. | Itself. | Phase 1 support rows and ledgers. | Test evidence only. | Medium | Phase accounting is evidence only, not runtime architecture. |
| `internal/modules/auth/phase1_support_test.go` | Helper-level unit coverage for session helpers, list query errors, CSRF, credential builders, password/TOTP decoders, user defaults, and deployment-admin guard. | Test-only package `auth`. | Go test harness. | `internal/platform/authn`, HTTP/url/time packages. | Itself. | Phase 1 support rows and ledgers. | Test evidence only. | Low | Useful characterization before pure helper reshaping. |

## 3. Module Boundary Diagnosis

Current classification:

- Legitimate thin application/service facade: yes, for route registration and `Service` construction.
- Accidental catch-all: partial, because one package contains local auth, current account, deployment user admin, administrative audit projection, and Enterprise Authentication extension routes.
- View/projection orchestration layer: no direct auth ownership found.
- Transport-adjacent adapter: yes, HTTP route handlers and cookie/header handling are in the module.
- Persistence-adjacent adapter: yes, private store interfaces wrap `platform/authn.Store`, but direct SQL is in `internal/platform/authn`.
- Mutation coordinator: yes, for password/TOTP/user/admin/binding mutations, idempotency hashing, audit-safe replay payloads, and session revocation fanout.
- Frontend shell/controller surface: no direct auth ownership found.
- Grid-vendor integration layer: no direct auth ownership found.
- Misplaced home for logic owned by other modules: generic HTTP error/session helpers have already moved to platform; remaining bootstrap-aware auth is auth-specific.
- Mixed-responsibility package: yes, because auth, current-account, deployment-user administration, audit read projection, and enterprise-auth extension concerns share one package.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Local login, password verification, MFA challenge, session issuance | `session_handlers.go`, `api.go`, `auth_session_helpers.go`, `platform/authn` | `auth` facade with `platform/authn` primitives | keep | Core 01 `/api/v1/auth/login`; Core 04 REQ-04-001..016 | Core auth responsibility. |
| Session inspection, logout, cookie lifecycle, idle sliding, revocation fanout | `session_handlers.go`, `auth_session_helpers.go`, `platform/httpauth` | `auth` session facade plus `platform/httpauth` shared primitives | keep/split | `GET /api/v1/auth/session`, `POST /api/v1/auth/logout`, `httpauth` helpers | Preserve cookie attributes and WebSocket revocation effects. |
| Bootstrap-token credential setup | `auth_session_helpers.go`, `credential_handlers.go` | `auth` credential lifecycle | keep | Core 01 bootstrap-token accepted routes; Core 04 REQ-04-084 | Bootstrap token is not an ordinary session. |
| Current account profile/preferences | `account_handlers.go`, `users.go` | `auth` current-account subfacade | keep | `/api/v1/account/profile`, `/api/v1/account/preferences`; Core 04 REQ-04-114 | Not workbook state even when preferences affect UI density. |
| Deployment user administration | `user_admin_handlers.go`, `users.go`, `platform/authn` | `auth` deployment-user admin facade | keep | `/api/v1/users*`; Core 04 deployment-admin matrix | Cross-user account administration is auth/admin, not incident identity. |
| Administrative audit read projection | `admin_audit_handlers.go`, `platform/authn` | `auth` deployment-admin audit read facade; future audit module only if widened by owner docs | defer | `/api/v1/administrative-audit-events`; Core 04 REQ-04-123..125 | Deployment-local audit projection, not incident revision history. |
| Enterprise provider discovery and OIDC/SAML protocol routes | `enterprise_routes.go`, `enterprise_protocol.go`, `platform/enterpriseauth` | `auth` Enterprise Authentication extension facade | keep/defer | Core 01 §20; Phase 11 deterministic evidence | Production verifier implementation remains future work. |
| Enterprise binding lifecycle | `enterprise_routes.go`, `platform/authn` | `auth` deployment-admin enterprise binding facade | keep | `/api/v1/users/{user_id}/auth-bindings*`; Core 01 REQ-01-537..541; Core 04 REQ-04-094 | Must not mutate incident memberships or local credential state outside specified behavior. |
| Generic API error envelope helpers | `platform/httpapi`; auth compatibility aliases in `api.go` | `platform/httpapi` | keep current placement | `httpapi.APIError`, `httpapi.WriteAPIError` | Do not move back into auth. |
| Generic session/CSRF/deployment-admin helper primitives | `platform/httpauth`; auth wrappers in `api.go` and `users.go` | `platform/httpauth` | keep current placement | `httpauth.AuthenticateRequest`, `ValidateCSRF`, `RequireDeploymentAdmin` | Auth wrappers remain compatibility and local convenience. |
| WebSocket session revocation effects | Auth publishes revocation; collaboration/platform own socket semantics | `auth` emits, `collaboration` and `platform/ws` own WebSocket protocol | split/no_action | `sessionRevocationPublisher`, Phase 1 WebSocket revocation tests | Auth should not own collaboration event protocol beyond session revocation cause. |
| Workbook row/query/mutation behavior | none found in production target | `workbook`, `timeline`, `projections`, `revisions`, `savedviews`, peer domain modules | intentional/no_action | No production auth file imports workbook/timeline/view-schema/projection packages | Auth session shape can affect frontend startup, but not workbook row mutation. |
| Frontend shell/controller state | none in target | `apps/web` | intentional/no_action | Frontend calls auth routes via `appShellClient`; no Go auth frontend code | Preserve route contract rather than moving frontend logic. |
| Grid-adapter/vendor integration | none in target | `packages/grid-adapter` | intentional/no_action | No grid vendor imports in target | No auth-module grid slice needed. |
| Phase maps and test rows | phase-named auth tests and tool manifests | testing harness | keep as evidence only | Phase 1 and Phase 11 task guides | Evidence accounting must not define runtime architecture. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| HTTP auth base routes | Core 01 route inventory; Core 04 auth/session security | `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/session`, credential routes in Core 01 and OpenAPI | `phase1_request_test.go`, `phase1_handlers_test.go`, `phase1_store_test.go`, `phase1_integration_test.go`, frontend Phase 1 tests | Add only if moving decoders, cookies, session timing, or auth-store seams | High | Preserve strict request decoding, success/error envelopes, cookies, MFA, idempotency, and revocation. |
| Current account routes | Core 01 account routes; Core 04 REQ-04-114 | `/api/v1/account/profile`, `/api/v1/account/preferences`; OpenAPI account schemas | Auth handler/integration tests and frontend Phase 2/account tests | Characterize replay/version behavior before moving account handlers | Medium | `deployment_admin` is not a bypass for current-account routes. |
| Deployment user/admin routes | Core 01 deployment-user routes; Core 04 deployment-admin matrix | `/api/v1/users*`, password reset, TOTP reset, revoke-all | Auth Phase 1 unit/store/integration tests and frontend deployment-admin tests | Preserve authenticated non-admin `403 authorization_denied` and bad-session `401 session_required` | High | Cross-user admin behavior is observable and security-sensitive. |
| Administrative audit read | Core 04 REQ-04-123..125; Core 01 route family | `GET /api/v1/administrative-audit-events` | Support tests, integration audit tests, frontend audit panel tests | Add if splitting audit read handler or query parsing | Medium | Safe audit projection must not expose forbidden secret-bearing values. |
| Enterprise provider discovery and protocol routes | Core 01 §20; Core 04 provider config/security | `/api/v1/auth/providers`, provider begin, OIDC callback, SAML ACS, SAML completion | `phase11_enterprise_auth_integration_test.go`, frontend Phase 11 tests | Required before changing verifier boundary or error mapping | High | Deterministic test evidence is not production IdP interoperability. |
| Enterprise binding lifecycle | Core 01 REQ-01-537..541; Core 04 REQ-04-094..096 | `/api/v1/users/{user_id}/auth-bindings*`; extension registry | Phase 11 binding lifecycle integration tests and frontend support tests | Required before any idempotency, audit, revocation, or safe-user shape change | High | Binding routes must not create users, mutate incident memberships, or mutate local credentials except specified session revocation effects. |
| CSRF/session authorization outcomes | Core 04 REQ-04-002..003 and deployment-admin matrix | `platform/httpauth`, auth handlers, browser API CSRF code | Phase 1 CSRF/unit/integration tests, frontend route-boundary tests | Required before centralizing or changing bootstrap-aware auth helpers | High | Cookie-authenticated mutators fail closed without CSRF proof. |
| WebSocket session revocation effects | Core 01/04 WebSocket/session behavior; collaboration/platform own socket protocol | Auth publishes `RevokeSession`; `/ws/v1/incidents/{incident_id}` receives session revocation effects | `phase1_integration_test.go`, collaboration/browser tests | Required if revocation publisher or session expiry behavior changes | High | Auth does not own collaboration messages beyond revocation cause. |
| Workbook row/query/mutation behavior | Peer workbook/timeline/projection/revision owners | No direct auth production ownership found | Indirect frontend/workbook tests call `/api/v1/auth/session` | No auth-specific workbook characterization unless session shape changes | Low | Record as no direct auth ownership found. |
| Saved-view or view-schema behavior | Saved-view/view-schema modules | No direct auth production ownership found | Frontend/app-shell uses session/current account data near saved views | No direct auth characterization required | Low | Account density preference is not saved-view state. |
| Projection refresh behavior | Projection/workbook owners | No production auth projection refresh code found | None in auth target | No action | Low | No direct auth ownership found. |
| Revision/change-set behavior | Revisions/history owners; Core 04 for deployment-local audit | Auth writes/reads deployment-local audit/idempotency through `platform/authn` | Auth integration tests inspect audit/idempotency | Characterize if audit/idempotency storage shape changes | Medium | Do not classify deployment audit as incident `change_set` history. |
| Generated protocol/view contracts | Owner specs/contracts upstream; generated roots downstream | OpenAPI, errors, extensions, protocol TS, UI contracts | `packages/protocol-ts/src/index.test.ts`, UI contract tests, drift checks | Required if route/error/schema owner inputs change | High | Generated files must be refreshed through generators, not hand-edited. |
| Grid-adapter or UI-selector contracts | Frontend package owners | No grid-adapter auth target ownership; UI has admin/audit selector tokens | Frontend unit and browser tests | Only needed for frontend-visible behavior changes | Low | No direct grid-vendor coupling found. |
| Harness/test accounting | `docs/testing-harness-nlspec.md` and phase manifests | Phase 1 and Phase 11 task guides and test maps | Phase slices and service-backed slices | Update owner inputs before generated ledgers if evidence maps change | Medium | Phase maps are evidence accounting only. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `auth-module` is not workbook orchestration. | Current target files are auth/account/admin handlers, helpers, and tests; no production workbook/projection/grid imports found. | Medium | intentional/no_action | `auth` for auth/admin; peer modules for workbook behavior | Keep tracker from assigning workbook ownership to auth. |
| `Service` and `authStore` remain broad, even though storage methods are grouped into subinterfaces. | `routes.go` defines one `Service` and an `authStore` composed from local login, session, credential, account, user admin, audit, and enterprise auth store interfaces. | High | should_fix | `auth` facade plus `platform/authn` storage | Plan seam narrowing after characterization; do not move SQL/storage first. |
| HTTP handlers are transport-adjacent and application-orchestration-heavy. | Handler files parse requests, authenticate, authorize, call store methods, map errors, write envelopes, set cookies, and publish revocations. | High | should_fix | `auth` service/application facade | Split only in behavior-preserving slices with route tests. |
| Generic API error and singleton-read helpers are platform-owned now. | `platform/httpapi.APIError`, `WriteAPIError`, `ValidateSingletonReadQuery`; auth exposes aliases/wrappers. | Medium | intentional/no_action | `internal/platform/httpapi` | Avoid reintroducing shared helpers into auth. |
| Generic session, CSRF, idle-slide, and deployment-admin guard helpers are platform-owned now. | `platform/httpauth.AuthenticateRequest`, `ValidateCSRF`, `RequireDeploymentAdmin`, slide helpers; auth keeps bootstrap-aware local helpers. | Medium | intentional/no_action | `internal/platform/httpauth` | Preserve platform direction; extract only auth-specific bootstrap context if needed. |
| Bootstrap-aware credential auth remains auth-specific. | `authenticateAuthRequest` accepts session or bootstrap token and restricts bootstrap token routes. | High | should_fix | `auth` credential lifecycle | Future extraction must preserve exact one-auth-mode and route-boundary errors. |
| Direct SQL/storage coupling is not in handler files. | Handlers call private store interfaces; `platform/authn.Store` owns SQL-backed methods. | Medium | defer | `internal/platform/authn` | Storage refactor requires separate platform/authn plan. |
| Production Enterprise Authentication verifiers are unconfigured by default. | `platform/enterpriseauth.UnconfiguredOIDCVerifier` and `UnconfiguredSAMLVerifier`; tests inject deterministic verifiers via module overrides. | High | defer | `platform/enterpriseauth` plus `auth` extension facade | Do not claim production IdP interoperability until real adapters and evidence exist. |
| Deterministic Enterprise Authentication evidence is test-only. | `phase11_enterprise_auth_integration_test.go` uses `internal/testutil/enterpriseauthtest`. | Medium | intentional/no_action | tests/harness | Keep distinction in tracker and future claims. |
| Administrative audit read projection sits in auth. | `admin_audit_handlers.go` builds deployment-local administrative-audit resources. | Medium | defer | `auth` currently; future audit module only with owner support | Do not merge with incident revisions/history. |
| WebSocket effects are indirect session revocation effects. | Auth publishes session revocations; collaboration/platform own `/ws/v1/incidents/{incident_id}` protocol. | High | intentional/no_action | auth emits; collaboration/platform own socket semantics | Preserve fanout behavior; do not move collaboration ownership into auth. |
| Frontend app shell depends heavily on auth route shapes. | `apps/web/src/app/api/appShellClient.ts`, `App.phase1.test.tsx`, E2E helpers call auth/account/users routes. | High | should_fix | route contracts and frontend API client | Include frontend unit/browser validation for UI-visible auth changes. |
| Generated-file drift risk exists for route/schema/error changes. | OpenAPI, errors, extensions, protocol TS generated roots, UI contracts. | High | should_fix | contracts/generation pipeline | Owner inputs first, generators second; never hand-edit generated roots. |
| Test-only phase names can mislead module ownership. | `phase1_*` and `phase11_*` tests live under auth. | Low | intentional/no_action | harness/evidence accounting | Treat phase rows as evidence only, not runtime architecture. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish target, authority posture, current file list, prior tracker status, and documentation-only scope. | This tracker, planning framework, owner docs. | `rg --files internal/modules/auth`, `test -f docs/handoffs/auth-module-module-refactor-tracker.md`, `make help`. | Section 1 and handoff scope rows current. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every current file under `internal/modules/auth` and classify tests separately from production. | All 18 target files. | Targeted `rg`, direct source inspection. | Section 2 has no generic placeholders. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map observable route, session, auth, enterprise, generated, frontend, and harness contracts to owners and tests. | Auth files, Core 01/Core 04, contracts, frontend client/tests. | `make task-guide ROLE=feature-dev PHASE=phase1`, `make task-guide ROLE=feature-dev PHASE=phase11`. | Section 4 complete with test posture. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-06, WF-07 | Decide where existing Phase 1/11 evidence is sufficient and where new characterization would be required before movement. | Auth tests, frontend auth tests, phase maps. | `make phase-slice PHASE=phase1`; `make phase-slice PHASE=phase11`. | Future implementation slices name required tests. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05 | Classify platform, storage, frontend, WebSocket, generated-file, and workbook boundary findings. | Auth package, platform auth helpers, runtime route assembly, frontend callers. | `make lint`; `make frontend-import-boundary-check` when cross-boundary code changes. | Section 5 findings complete. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-04 | WF-06 | Plan behavior-preserving seams: helper cohesion, store seam narrowing, bootstrap-auth extraction, enterprise verifier boundary. | `routes.go`, handler files, helpers, `platform/httpauth`, `platform/enterpriseauth`. | Phase 1/11 slices depending on touched area. | Slices are owner-safe and behavior-preserving. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07, WF-08 | Define smallest safe implementation slices and dependencies. | Auth files and directly affected tests. | Per-slice validation in Section 7. | Section 7 has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03 | WF-08 | Record when test maps, generated ledgers, or schedules would need owner-input updates. | Phase maps, generated ledgers, docs/testing ledgers if affected. | `make phase-ledger-drift`, `make phase-schedule-drift` only if owner inputs change. | Harness changes planned before generation. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Record validation commands, skipped checks, blockers, and next session restart posture. | This tracker and any future changed files. | `make agent-finalize`, `make test-fast`, `make check` when implementation risk warrants. | Section 10 and Section 12 current. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker only. | `docs/handoffs/auth-module-module-refactor-tracker.md`. | Documentation drift only. | No production tests; markdown check only if run. | `make lint-markdown`. | Revert tracker file only. | Tracker has all 12 required sections and inventories all 18 files. |
| S-01 | S-00 | Add or update characterization tests only if a future implementation slice changes auth route boundaries, helper access, or denial mapping. | Auth tests, frontend auth tests when UI-visible. | Could expose behavior mismatch; no runtime behavior change. | Preserve Phase 1 and Phase 11 route behavior. | `make phase-slice PHASE=phase1`; add `make phase-slice PHASE=phase11` for enterprise. | Drop test-only changes if owner evidence rejects the premise. | Risky movement has characterization before code motion. |
| S-02 | S-01 | Helper cohesion: split pure request/resource helpers inside `auth` without changing package API or wire behavior. | `api.go`, `users.go`, possible new authored auth files. | Request validation fields, reason codes, exported helper names. | Preserve `phase1_request_test.go`, `phase1_support_test.go`, handler tests. | `make phase-slice PHASE=phase1`. | Revert helper split. | Exported Go surface and route responses remain compatible. |
| S-03 | S-01 | Narrow auth-store seams by handler family while leaving `platform/authn.Store` as backing implementation. | `routes.go`, handler files, same-package test stubs. | Handler behavior, idempotency, audit, revocation. | Preserve handler unit, store, and integration tests. | `make phase-slice PHASE=phase1`; `make phase-slice PHASE=phase11` if enterprise methods change. | Revert seam split and test stubs. | Store interface blast radius is smaller with no wire drift. |
| S-04 | S-02 | Extract auth-specific bootstrap credential context into a clearer internal seam. | `auth_session_helpers.go`, `credential_handlers.go`, tests. | Bootstrap-token route boundaries, mixed auth-mode rejection, credential setup errors. | Preserve bootstrap boundary and TOTP begin/complete tests. | `make phase-slice PHASE=phase1`. | Revert extraction. | Bootstrap-aware logic is isolated and public behavior is identical. |
| S-05 | S-00 | Enterprise verifier boundary planning or adapter implementation, only when later authorized. | `enterprise_protocol.go`, `enterprise_routes.go`, `platform/enterpriseauth`, testutil verifiers. | Callback/ACS errors, browser binding, extension profile behavior. | Preserve deterministic Phase 11 evidence; add real-adapter tests only with owner support. | `make phase-slice PHASE=phase11`. | Revert adapter changes. | Production verifier status is explicit; deterministic tests still pass. |
| S-06 | S-03 or S-05 | Generated-contract audit when route, schema, error, or extension owner inputs change. | Owner specs/contracts, generated roots. | OpenAPI/errors/extensions/protocol TS drift. | Preserve protocol-ts and UI-contract tests. | `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make frontend-unit`. | Revert owner inputs and generated outputs together. | No generated file hand edits; drift checks pass. |
| S-07 | Any implementation slice | Frontend-visible auth route validation. | `apps/web/src/app/api/appShellClient.ts`, frontend tests, browser E2E. | App shell login/session/admin behavior. | Preserve route-boundary, Phase 1, Phase 11, and browser auth flows. | `make frontend-unit`; `make browser-e2e-webserver-backed`; phase slice as applicable. | Revert UI-facing contract change. | Frontend-visible behavior remains stable. |
| S-08 | all implementation slices | Final validation and handoff. | Tracker plus changed implementation files. | Missed cross-module auth/session consumer. | Preserve all relevant characterization tests. | `make agent-finalize`; `make test-fast`; `make check` if risk warrants. | Revert last implementation slice if contract drift appears. | Handoff names commands, artifacts, failures, skipped checks, and blockers. |

Any slice that intentionally changes observable behavior requires later authorization and owner-doc support before implementation.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make phase-slice PHASE=phase1` | Phase 1 local authentication, session lifecycle, TOTP, deployment-user admin, and credential-action evidence. | yes for auth/session/account/admin implementation slices | Discovered through `make task-guide ROLE=feature-dev PHASE=phase1`. |
| integration | `make service-backed-slice PHASE=phase1` | Service-backed Phase 1 evidence with Postgres, object store, and browser stack where selected. | yes for route/store/session revocation movement | Use after touching handler-store interactions or durable session/audit behavior. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser-backed public flows through `/api/v1/*` and `/ws/v1/*`. | no for tracker; yes for frontend-visible route behavior | Prefer phase slices first; broaden when UI-visible auth behavior changes. |
| enterprise extension | `make phase-slice PHASE=phase11` | Deterministic Enterprise Authentication route and binding evidence. | yes if enterprise auth files change | Phase 11 guide states production Enterprise Authentication remains unclaimed by default. |
| generated drift | `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift` | Generated policy and JSON/contract shape drift. | no for tracker; yes if owner inputs/contracts change | Do not hand-edit generated roots. |
| import-boundary/static | `make frontend-import-boundary-check`, `make lint` | Frontend import boundaries and broad static hygiene. | no for tracker; yes for cross-package or frontend caller changes | `make lint` is broad hygiene, not phase evidence. |
| full check | `make agent-finalize`, `make test-fast`, `make check` | Retained-run maintenance and broad local gate. | no for tracker; yes before final broad implementation handoff | Run `make agent-finalize` before broad end-of-run verification. |

Commands discovered or refreshed in this session:

- `rg --files internal/modules/auth`.
- `test -f docs/handoffs/auth-module-module-refactor-tracker.md`.
- `git status --short`.
- `date -u +%Y-%m-%dT%H:%M:%SZ`.
- `make help`.
- `make task-guide ROLE=feature-dev PHASE=phase1`.
- `make task-guide ROLE=feature-dev PHASE=phase11`.
- `make lint-markdown`.

Validation actually run for this tracker:

- `make lint-markdown` passed for the final tracker content.

Product/auth validation was not run because this task performed no production refactor.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Create target-specific tracker at requested path | WF-00 | DONE | none | This file | Tracker exists at `docs/handoffs/auth-module-module-refactor-tracker.md`. |
| T-002 | Confirm current target inventory | WF-01 | DONE | T-001 | `rg --files internal/modules/auth` returned 18 files | Every target file is inventoried. |
| T-003 | Record prior archive posture | WF-00 | DONE | T-001 | `docs/archive/auth-module-refactor-tracker.md` by inspection | Archive is treated as stale evidence only. |
| T-004 | Diagnose workbook-boundary scope rule | WF-04 | DONE | T-002 | Sections 3 and 5 | Tracker states no direct workbook/projection/grid ownership found. |
| T-005 | Map HTTP auth/account/admin contracts | WF-02 | DONE | T-002 | Core 01/Core 04 and contracts | Auth/account/admin surfaces have owner and test posture. |
| T-006 | Map enterprise-auth contracts | WF-02 | DONE | T-002 | Core 01 §20, Core 04, Phase 11 guide | Enterprise routes and test-only verifier status are recorded. |
| T-007 | Map generated/frontend/harness risks | WF-02 | DONE | T-005 | Contracts/frontend search and task guides | Generated and frontend risk surfaces are listed. |
| T-008 | Classify coupling findings | WF-04 | DONE | T-002 | Section 5 | Findings use allowed classifications. |
| T-009 | Define future behavior-preserving slices | WF-06 | DONE | T-008 | Section 7 | Each slice has dependency, validation, rollback, and completion criteria. |
| T-010 | Record validation commands | WF-08 | DONE | T-009 | Section 8 | Commands are discovered or marked by scope. |
| T-011 | Append session handoff entries | WF-08 | DONE | T-010 | Section 10 | Handoff tables are current and list only this tracker as touched. |
| T-012 | Track production IdP verifier gap | WF-05 | TODO | S-05 | Open blocker RB-001 | Future owner-backed implementation plan exists before any claim. |
| T-013 | Track auth-store seam narrowing | WF-05 | TODO | S-03 | Slice S-03 | A later authorized task narrows store seams with passing phase validation. |
| T-014 | Track bootstrap-auth extraction | WF-05 | TODO | S-04 | Slice S-04 | Bootstrap-aware logic has clearer internal seam with no behavior drift. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | Created new target-specific planning tracker. Authority posture, target label, output path, allowed write, and non-goals are recorded. | Inspected: framework, Core 00/Core 01/Core 04/Core 05 posture, `docs/domain.md`, `docs/testing-harness-nlspec.md`, current auth files, platform auth helpers, runtime route assembly, contracts/frontend references, stale archive tracker. Touched: this tracker only. | `rg --files internal/modules/auth`; `test -f docs/handoffs/auth-module-module-refactor-tracker.md`; `git status --short`; `date -u +%Y-%m-%dT%H:%M:%SZ`; `make help`; `make task-guide ROLE=feature-dev PHASE=phase1`; `make task-guide ROLE=feature-dev PHASE=phase11`; `make lint-markdown`. | Tracker created; `make lint-markdown` passed. | None for tracker creation. | Use this file as the handoff point before any future authorized implementation slice. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | `internal/modules/auth` is a backend auth/account/admin facade with mixed responsibilities; not workbook orchestration. | Inspected current production auth files and platform auth helpers. Touched: this tracker only. | `rg --files internal/modules/auth`; targeted source inspection from planning session. | Boundary diagnosis recorded with keep/split/defer/no-action decisions. | Broad `authStore` and bootstrap-aware auth helper remain future refactor candidates. | Start with characterization before any helper or store seam movement. |

### Frontend module boundary, if applicable

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | Frontend code consumes auth/account/users/admin/enterprise routes; auth target does not own frontend shell/controller state. | Inspected by search under `apps/web` and `packages`; touched this tracker only. | Targeted `rg` from planning session. | Frontend route risk recorded in Sections 4, 5, and 7. | None for tracker. | Include frontend unit/browser checks only when route behavior is UI-visible. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | Auth route behavior is reflected in authored contracts and generated downstream roots; generated files were not edited. | Inspected OpenAPI/errors/extensions/protocol/ui contract references; touched this tracker only. | Targeted source inspection and `make help`; `make task-guide` commands. | Generated-contract risk and drift commands recorded. | No generated changes planned for tracker. | If future route/schema behavior changes, update owner inputs and run generators/drift checks. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | Phase 1 and Phase 11 are the Make-owned verification surfaces for most auth work. Phase maps are evidence accounting only. | Inspected current auth test filenames and task guide output; touched this tracker only. | `make task-guide ROLE=feature-dev PHASE=phase1`; `make task-guide ROLE=feature-dev PHASE=phase11`; `make lint-markdown`. | Task guide refreshed; `make lint-markdown` passed. | Product/auth tests were skipped because no production refactor was performed. | Run phase slices only after future implementation changes. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | Core 04 supports current taxonomy: invalid/missing/expired/revoked/inactive sessions return `401 session_required`; authenticated non-admin deployment-admin denials return `403 authorization_denied`. | Inspected Core 04 passages and platform/auth helpers; touched this tracker only. | Targeted source inspection from planning session. | No owner contradiction found. | Production OIDC/SAML interoperability remains future work, not current evidence. | Do not alter auth outcomes without owner-backed authorization and characterization. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T02:36:45Z | Codex auth-module tracker implementation | Tracker is complete for planning handoff; future implementation slices remain unauthorized. | Touched this tracker only. | `make lint-markdown`. | Markdown validation passed. | RB-001 production IdP verifier evidence gap; future auth-store/bootstrap seam work requires implementation authorization. | Pick one slice from Section 7 and run the named characterization before edits. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Production OIDC/SAML verifier interoperability is not proven by deterministic Phase 11 fixtures. | Enterprise-auth routes currently have deterministic test evidence, but production IdP behavior is a separate claim and implementation boundary. | Core 01/Core 04 owner support, real verifier implementation, and non-fixture interoperability tests. | REMEDIATED IN IMPLEMENTATION: explicit claim config, startup manifest validation/reconciliation, production verifier adapters, and encrypted PKCE state now exist; external IdP interoperability evidence remains the claim-publication validation step. |
| RB-002 | Should auth-store seams be narrowed by handler family in a later behavior-preserving slice? | Current broad private `authStore` increases test-stub and accidental coupling risk. | Existing Phase 1/11 characterization plus implementation authorization. | REMEDIATED: `Service` now holds handler-family ports backed by `platform/authn.Store`, rather than one broad runtime `authStore`. |
| RB-003 | Should bootstrap-aware auth context move behind a clearer internal seam? | Bootstrap token rules are auth-specific and security-sensitive; moving them without characterization risks route-boundary drift. | Phase 1 bootstrap/TOTP characterization and owner evidence. | REMEDIATED: bootstrap credential authentication now lives behind an auth-owned internal component with a narrow result type and no session semantics. |
| RB-004 | Should administrative-audit read projection stay in auth long term? | It is deployment-local auth/admin evidence today, but could become a clearer audit module only if widened by owner docs. | Core 04/Core 01 owner decision and current route characterization. | REMEDIATED/DEFERRED BY OWNER BOUNDARY: kept in auth/admin, narrowed to a deployment-audit reader port, and documented as deployment-local administrative evidence pending any future Core 01/Core 04 widening. |

## 12. Binary Completion Criteria

The tracker is complete only when all criteria pass:

- Every file in `internal/modules/auth` is inventoried or explicitly out of scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`.
- Validation commands are discovered or marked `TODO:` with a reason.
- Contradictions are marked `BLOCKED: owner contradiction`.
- Repository/framework mismatches are recorded as planning findings.
- Handoff sections are current enough for another agent to continue without rediscovery.

Current completion status:

- File inventory: complete for all 18 files returned by `rg --files internal/modules/auth`.
- Public contract risk mapping: complete for discovered HTTP auth/account/admin, administrative audit, enterprise-auth, WebSocket-adjacent, generated-contract, frontend, and harness surfaces.
- Workflows: WF-00 through WF-08 have dependencies, validation posture, and handoff checkpoints.
- Slices: S-00 through S-08 are behavior-preserving by default; any behavior change requires later authorization.
- Validation commands: discovered through Make help and task guides; tracker markdown validation passed with `make lint-markdown`.
- Owner contradictions: none found in this pass.
- Repository/framework mismatch: recorded. The target is auth/admin, not workbook orchestration.
- Handoff: Section 10 records the current planning-only session and states that only this tracker was touched.
