# Auth Module Refactoring Tracker and Handoff

Created: 2026-07-01

## 1. Scope and Source Posture

Target directory: `internal/modules/auth`.

Allowed changes for this artifact:

- Document current repository state for `internal/modules/auth`.
- Identify public contracts that must be preserved by later refactors.
- Record boundary findings, candidate owners, workstreams, and safe slices.
- Mark unknowns as `TODO:` and owner-document contradictions as
  `BLOCKED: owner contradiction`.

Non-goals:

- Do not rename or split the module in this tracker.
- Do not treat `auth` as a permanent valid module boundary solely because the
  directory exists.
- Do not move workbook, projection, revision, import, entity, evidence, link,
  saved-view, or view-contract behavior as part of this document.
- Do not hand-edit generated roots or phase ledgers.

Authority posture:

1. Adopted subsystem NLSpecs govern only their named subsystem. No auth-specific
   adopted subsystem NLSpec was found during this pass.
2. Core 00 through Core 04 govern implementation-conformance behavior.
3. Core 05 is relevant only for claim-bearing timed, fixture-sensitive, or
   publication evidence.
4. `docs/domain.md` and implementation-support guides provide terminology,
   package-boundary, harness, and execution support.
5. Current repository code and tests describe current implementation state.
6. Prior plans and framework files are planning evidence, not authority.

Framework posture: `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
was used as the planning template and doctrine. It is not proof of current
repository state.

Architectural finding: `internal/modules/auth` is not a workbook package and
does not currently own workbook row/query/mutation behavior, saved views, view
schemas, projection refresh, revisions, collaboration semantics, imports or
tabular ingest, entities or indicators, evidence, links, or view contracts. It
currently acts as a mixture of auth HTTP/application facade, deployment
account/admin behavior, enterprise-auth extension facade, session revocation
signaling, and a small number of retained compatibility aliases.

## 2. Current-State Repository Inventory

| path | current responsibility | exported/public symbols or package surface | inbound callers | outbound dependencies | tests touching it | generated artifacts or contracts touched | suspected target owner module | risk level | notes |
| ---- | ---------------------- | ------------------------------------------ | --------------- | --------------------- | ----------------- | ---------------------------------------- | ----------------------------- | ---------- | ----- |
| `internal/modules/auth/account_handlers.go` | Account profile and account preference HTTP handlers. | Package-private methods on `Service`. | Registered through `RegisterRoutes` in `routes.go`; frontend/account clients through HTTP. | `authn`, `httpapi`, account decoders/resources in `users.go`, shared auth/session helpers. | Phase 1 handler, integration, and support-integration tests. | OpenAPI account route schemas and error registry if wire behavior changes. | auth/account facade. | medium | Account preferences remain deployment-local account state, not workbook saved-view or incident preference state. |
| `internal/modules/auth/admin_audit_handlers.go` | Administrative audit query parsing, filtering, resource shaping, pagination, and handler. | Package-private helper functions and method on `Service`. | Registered through `RegisterRoutes` in `routes.go`; frontend `DeploymentAuditPanel` through HTTP. | `authn`, `httpapi`, `listquery`, `pagination`, JSON/time/UUID utilities, shared session helpers. | Phase 1 integration/support-integration tests. | OpenAPI administrative audit path and list-query envelopes if changed. | deployment admin/account audit facade. | medium | Split from route assembly; deployment-admin authorization remains unchanged. |
| `internal/modules/auth/api.go` | Auth request decoding, CSRF helper, bootstrap route helper, credential-state and TOTP setup resource helpers. | `APIError`, `AuthSource`, `LoginRequest`, `SecondFactorAssertion`, `PasswordChangeRequest`, `TOTPBeginRequest`, `TOTPCompleteRequest`, decode helpers, `ValidateCSRF`, `AllowsBootstrapTokenRoute`, credential/TOTP builders, error builders. | Same-package handlers; package tests; no non-test external production caller found for most helpers. | `internal/platform/httpapi`, `internal/platform/httpauth`, `internal/platform/authn`; Go JSON/HTTP/time/url utilities. | `phase1_request_test.go`, `phase1_support_test.go`, `phase1_handlers_test.go`, `phase1_integration_test.go`, `phase1_support_integration_test.go`. | OpenAPI auth request/response schemas and error code registry are affected if wire shape changes. | auth/account facade; pure helper area can remain in auth. | medium | Retains active compatibility aliases to platform API/auth concepts; removed unused session auth wrappers in WS-06. |
| `internal/modules/auth/auth_session_helpers.go` | Shared auth/session helper methods for test touch, credential/session authentication, bootstrap tokens, cookies, session resources, idle sliding, revocation publishing, idempotency hashes, stored responses, and API-error forwarding. | Package-private methods/helpers on `Service`; `buildSessionResource` and `buildSafeUserResource` stay package-private. | Same-package handlers and tests. | `authn`, `httpapi`, retained `AuthSource` alias, JSON/SHA/context/HTTP/time/UUID utilities. | Phase 1 handler, support, integration, store-backed, and support-integration tests. | Session response schemas, cookies, CSRF, and error envelopes if changed. | auth/session facade over platform primitives. | high | Membership summaries are informational bootstrap state only; authorization uses route-specific checks. |
| `internal/modules/auth/credential_handlers.go` | Credential state, password change, TOTP begin/complete, and active TOTP validation handlers. | Package-private methods on `Service`. | Registered through `RegisterRoutes` in `routes.go`; frontend auth/account surfaces through HTTP. | `authn`, `httpapi`, UUID/time/errors utilities, request decoders/resources in `api.go`, shared session helpers. | Phase 1 handler, request, support, integration, and support-integration tests. | OpenAPI credential/password/TOTP routes and error registry if changed. | auth credential lifecycle facade. | high | Preserves bootstrap-token route limits, idempotency, cookie clearing, and session revocation behavior. |
| `internal/modules/auth/enterprise_protocol.go` | Package-local protocol aliases for enterprise auth verifiers. | Package-private aliases `oidcVerifier`, `samlVerifier`. | `routes.go` service construction and `enterprise_routes.go`. | `internal/platform/enterpriseauth`. | `phase11_enterprise_auth_integration_test.go`. | Enterprise-auth extension profile and OpenAPI enterprise route families if protocol boundary changes. | enterprise-auth extension facade with platform verifier boundary. | low | Thin indirection only; no workbook ownership. |
| `internal/modules/auth/enterprise_routes.go` | Enterprise provider discovery/begin, OIDC callback, SAML ACS/complete, and deployment-admin enterprise auth binding lifecycle. | `EnterpriseAuthBeginRequest`, binding request structs, decode helpers; route handlers are methods on `Service`. | Registered through `RegisterRoutes` in `routes.go`; exercised by frontend app shell clients and phase 11 tests through HTTP. | `internal/platform/authn`, `internal/platform/httpapi`, verifier aliases from `enterprise_protocol.go`; cookie, URL, SHA256, UUID utilities. | `phase11_enterprise_auth_integration_test.go`; frontend `phase11.enterprise-auth` and app shell tests through HTTP contract. | `contracts/extensions/index.json`, OpenAPI enterprise auth paths, generated `protocol-ts` contract bundle, auth error registry. | enterprise-auth extension facade. | high | Route shape, redirect behavior, transaction cookies, provider profile gating, idempotency, and session revocation are observable. |
| `internal/modules/auth/routes.go` | Auth service assembly, narrowed private store seams, session revocation publisher contract, extension verifier setup, and public/test route registration. | `Service`, `SessionPrincipal`, `CredentialAuthContext`, `EnterpriseOIDCVerifierOverrideKey`, `EnterpriseSAMLVerifierOverrideKey`, `RegisterRoutes`, `RegisterTestRoutes`. | `internal/app/runtime.go`; `cmd/server/main.go` test-route registration; integration harnesses; frontend and workbook shells through HTTP. | `internal/platform/authn`, `enterpriseauth`, `httpapi`, `httpauth`, `pagination`; UUID and stdlib context/HTTP/time. | All phase 1 auth tests and phase 11 enterprise test touch the service path. | OpenAPI auth/account/user/admin audit paths, error registry, generated protocol contracts, phase 1 and phase 11 harness maps. | auth/account/admin facade plus enterprise extension entrypoint. | high | Route registration order and envelopes remain stable; handler logic is now split into cohesive files. |
| `internal/modules/auth/session_auth.go` | Removed compatibility wrapper over platform session authentication. | Removed `SessionAuthStore`, `SessionRevoker`, `SessionAuthOptions`, and `AuthenticateSessionRequest`. | Focused caller inventory found no repo caller outside the removed file. | N/A. | Phase 1 slice and `make test-fast` passed after removal. | None. | platform `httpauth` owns shared session authentication primitives. | low | Deleted in WS-06 to remove accidental API surface. |
| `internal/modules/auth/session_handlers.go` | Login, session inspection, and logout HTTP handlers. | Package-private methods on `Service`. | Registered through `RegisterRoutes` in `routes.go`; frontend/app shells through HTTP. | `authn`, `httpapi`, request decoders, shared cookies/session helpers. | Phase 1 handler, integration, store-backed, and support-integration tests. | OpenAPI login/session/logout paths and cookie behavior if changed. | auth/session facade. | high | Preserves public cookies, CSRF, MFA challenges, session concurrency, and revocation side effects. |
| `internal/modules/auth/users.go` | Deployment user, account profile, account preference, and admin credential-action request decoders and resource builders. | User/account/admin request structs; decode helpers; `BuildSafeUserResource`, `BuildSafeUserResourceWithEnterpriseBindings`, `BuildAccountProfileResource`, `BuildAccountPreferencesResource`, `RequireDeploymentAdmin`. | Local handlers; `internal/testutil/phase0test/phase0test.go` calls `BuildSafeUserResource`; package tests. | `internal/platform/authn`, `internal/platform/httpauth`; JSON/HTTP/string utilities. | `phase1_support_test.go`, `phase1_handlers_test.go`, `phase1_integration_test.go`, `phase1_support_integration_test.go`, phase 0 testutil indirectly. | OpenAPI account/user/admin schemas, generated protocol contracts, UI admin/account clients. | auth/account/admin facade. | high | Account preferences are deployment-local account preferences per domain, not workbook saved views. |
| `internal/modules/auth/user_admin_handlers.go` | Deployment user collection/member handlers, admin password/TOTP/session actions, list parsing, and pagination. | Package-private helper functions and methods on `Service`. | Registered through `RegisterRoutes` in `routes.go`; frontend deployment-admin clients through HTTP. | `authn`, `httpapi`, `listquery`, `pagination`, UUID/string/errors utilities, shared safe-user/session helpers. | Phase 1 handler, integration, store-backed, and support-integration tests. | OpenAPI user/admin credential routes and error registry if changed. | deployment user/admin facade. | high | Preserves deployment-admin guard, last-admin constraints, idempotency, and target session revocation. |
| `internal/modules/auth/phase1_handlers_test.go` | Handler-level characterization for phase 1 auth, account, user admin, credential actions, CSRF, and route contracts using stubs. | Test-only helpers, stubs, and `Test*` functions. | Go test runner. | Auth package internals; platform `authn`; HTTP test utilities. | It is a test file. | Phase 1 evidence accounting. | tests and harness evidence. | medium | Broad stub mirrors `authStore`; future seam narrowing will likely touch this file. |
| `internal/modules/auth/phase1_integration_test.go` | Integration characterization for auth/session lifecycle, bootstrap, WS rejection, user admin, audit, idempotency, and test route setup. | External-package `auth_test`; `Test*` functions and test harness helpers. | Go test runner; route registrars. | `internal/modules/auth`, runtime/testutil support, platform stores, websocket/test helpers. | It is a test file. | Phase 1 ledger and route inventory evidence. | tests and harness evidence. | high | Protects observable behavior including session revocation and bootstrap token boundaries. |
| `internal/modules/auth/phase1_request_test.go` | Unit tests for login request shape and bootstrap route helper behavior. | `Test*` functions. | Go test runner. | Auth package helpers. | It is a test file. | Phase 1 request contract evidence. | tests and harness evidence. | low | Useful guard before decoder/helper splits. |
| `internal/modules/auth/phase1_store_test.go` | Store-backed auth behavior tests for session concurrency, idle sliding, revocation consequences, user patching, and last-admin guard. | `Test*` functions. | Go test runner. | Platform auth store and database test fixtures. | It is a test file. | Phase 1 storage behavior evidence. | tests and harness evidence. | high | Tests storage semantics but production auth code depends on platform store facade rather than raw SQL. |
| `internal/modules/auth/phase1_support_integration_test.go` | Inventory-driven phase 1 integration sweeps for envelopes, bootstrap boundaries, CSRF, replay payloads, audit, session revocation, authorization, and request contracts. | `Test*` functions. | Go test runner; `phase1test.PublicRouteInventory`. | Auth route registration, runtime harness, phase 1 route inventory. | It is a test file. | Phase 1 harness accounting and route inventory evidence. | tests and harness evidence. | high | This is the main safety net for preserving route/accounting behavior during refactors. |
| `internal/modules/auth/phase1_support_test.go` | Pure helper tests for session query helpers, CSRF, credential state, TOTP setup, safe user shape, and deployment admin guard. | `Test*` functions. | Go test runner. | Auth helper functions and platform authn records. | It is a test file. | Phase 1 helper behavior evidence. | tests and harness evidence. | low | Should remain green for helper file splits. |
| `internal/modules/auth/phase11_enterprise_auth_integration_test.go` | Enterprise auth integration characterization for provider discovery/begin, OIDC, SAML, and binding lifecycle. | `Test*` functions and deterministic verifier fixtures. | Go test runner. | Auth route registration, platform enterprise auth/test verifiers, runtime harness. | It is a test file. | Phase 11 enterprise-auth extension evidence. | tests and harness evidence. | high | Deterministic verifier evidence does not prove real IdP interoperability. |

## 3. Workbook Boundary Diagnosis

`internal/modules/auth` is not a legitimate workbook facade. The directory does
not contain workbook row/query/mutation orchestration, workbook preferences,
view-schema logic, saved-view logic, projection refresh behavior, revision or
change-set coordination, import/tabular ingest behavior, entity/indicator
ownership, evidence ownership, or link ownership.

The package is a mixture of auth service facade, deployment account/admin
surface, enterprise-auth extension facade, and session/authorization adapter.
Workbook-facing impact is indirect: clients such as
`WorkbookShell` call `/api/v1/auth/session`, authorization derives from the
session, bootstrap tokens are rejected from `/ws/v1/*`, and auth can notify the
WebSocket hub when sessions are revoked.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| -------------------- | ---------------- | ----------------------- | --------------------------- | -------- | ----- |
| Login, logout, session issuance, idle/max expiry, session inspection | `session_handlers.go`, `auth_session_helpers.go`, `api.go` | auth/session facade with platform `httpauth`/`authn` primitives | keep | Core 01/Core 04 route and session requirements; phase 1 tests | Preserve cookie, bearer, envelope, and idle-slide behavior. |
| Password change, credential state, TOTP bootstrap/replacement | `credential_handlers.go`, `api.go` | auth credential lifecycle | keep | Core 01/Core 04 credential and TOTP requirements; phase 1 tests | Bootstrap token remains limited to TOTP begin/complete. |
| Account profile and account preferences | `account_handlers.go`, `users.go` | auth/account facade | keep | Domain says account preference is deployment-local density override | Do not move to workbook saved views or incident preferences. |
| Deployment user admin and credential admin actions | `user_admin_handlers.go`, `users.go` | deployment admin/account facade | keep | Core 01/Core 04 user/admin routes and authorization matrix | Deployment admin is not incident admin or content bypass. |
| Administrative audit event read | `admin_audit_handlers.go` | deployment admin/account audit facade | keep | Core 01 route family; frontend `DeploymentAuditPanel` | Preserve list query, pagination, and deployment-admin guard. |
| Enterprise provider discovery and login | `enterprise_routes.go`, `enterprise_protocol.go`, `routes.go` | enterprise-auth extension facade | keep, maybe split from base auth routing | Core 01/Core 04 enterprise auth profile; phase 11 tests | Profile gating and redirect/cookie semantics are observable. |
| Enterprise auth binding create/rotate/retire | `enterprise_routes.go`, `users.go`, `routes.go` | enterprise-auth extension facade plus deployment user admin | keep, split from base auth routing | Core 01 binding routes; phase 11 tests | Retire/rotate revoke target sessions. |
| WebSocket session revocation notification | `auth_session_helpers.go` via `sessionRevocationPublisher` | auth emits revocation; collaboration/platform WS owns socket semantics | keep | Phase 1 integration tests for session revocation and WS rejection | Auth publishes revocation reason codes; socket semantics remain outside auth. |
| Incident membership summaries in session resource | `auth_session_helpers.go` via `buildSessionResource` and a named membership-summary interface | session resource owner; incident membership data remains incident/authn storage concern | keep | Session response consumers and tests | Informational bootstrap state only; authorization does not derive from this member. |
| Workbook row/query/mutation behavior | Not present under auth | workbook module | no_action | Search found no auth production ownership | Auth only influences workbook access through session/authorization. |
| Saved-view or view-schema behavior | Not present under auth | savedviews/viewschemas modules | no_action | Search found no auth production ownership | Account density preference is not saved-view state. |
| Projection refresh, revision/change-set, import ingest, entity/evidence/link behavior | Not present under auth | projections/revisions/imports/entities/evidence/links owners | no_action | Search found no auth production ownership | Do not invent movement out of auth. |

## 4. Public Contract and Behavior Freeze Map

| Contract area | Observable contract | Existing tests or evidence | Required characterization before implementation |
| ------------- | ------------------- | -------------------------- | ---------------------------------------------- |
| Auth HTTP routes | `/api/v1/auth/login`, `/logout`, `/session`, `/credential-state`, `/password/change`, `/mfa/totp/begin`, `/mfa/totp/complete`. | Phase 1 handler, request, support, integration, and support-integration tests; OpenAPI paths. | Add targeted tests only if a slice changes request decoding, cookie handling, status codes, response envelopes, or idempotency. |
| Account routes | `/api/v1/account/profile` and `/api/v1/account/preferences`. | Phase 1 handler/integration tests; frontend app shell client and tests; OpenAPI/account schemas. | Characterize profile/preference conflict and replay behavior before splitting account handlers. |
| Deployment user routes | `/api/v1/users`, `/api/v1/users/{user_id}`, password reset, TOTP reset, revoke-all. | Phase 1 handler, store, integration, and support-integration tests; frontend admin clients. | Preserve deployment-admin guard, safe user shape, pagination/list query, last-admin guard, and session revocation effects. |
| Administrative audit route | `/api/v1/administrative-audit-events`. | Phase 1 integration/support-integration tests; frontend `DeploymentAuditPanel`; UI contract token. | Add focused characterization if list filters, resource shape, or authorization layer is touched. |
| Enterprise auth provider routes | `/api/v1/auth/providers`, provider begin, OIDC callback, SAML ACS, SAML ACS complete. | Phase 11 integration tests; frontend phase 11 tests; extension profile contract. | Preserve profile gating, transaction storage, redirect status, browser binding cookie, return_to validation, and error envelopes. |
| Enterprise auth binding routes | `/api/v1/users/{user_id}/auth-bindings`, rotate, delete. | Phase 11 integration tests; frontend app shell support and route-boundary tests; OpenAPI paths. | Preserve deployment-admin guard, idempotency, safe user response, audit, and target session revocation. |
| WebSocket paths/events | Bootstrap tokens rejected from `/ws/v1/*`; session revocation closes accepted session-bound sockets. | Phase 1 integration and support-integration tests; collaboration owns socket semantics. | Any auth revocation or bootstrap-token change must rerun WS/session tests and collaboration-adjacent checks. |
| Workbook interaction | Workbook clients depend on `/api/v1/auth/session` for current user/session context. | `WorkbookShell` and workbook visual/a11y/unit tests mock or call the session route. | No workbook mutation characterization needed unless auth session shape changes. |
| Saved-view/view-schema behavior | No direct auth ownership found. | Saved views/view schemas are registered separately in `internal/app/runtime.go`. | Required only if account preferences are incorrectly conflated with workbook preferences. |
| Projection refresh behavior | No direct auth ownership found. | No auth production imports of projection owner code found. | TODO: characterize only if future slice introduces projection coupling. |
| Authorization checks | Session required, CSRF for cookie state changes, deployment-admin checks, bootstrap-token route limits, account-current-session checks. | Phase 1 support/integration sweeps; Core 04 authorization matrix. | Required for any movement of `RequireDeploymentAdmin`, CSRF, bootstrap credential context, or session lookup logic. |
| Revision/change-set behavior | No incident revision/change-set ownership found in auth. | Auth admin actions use deployment audit/idempotency, not incident change sets. | No characterization required unless future implementation touches revision modules. |
| Generated protocol/view contracts | OpenAPI, errors, extensions, `protocol-ts` generated bundle, `ui-contracts` admin token. | Contract files and frontend tests reference auth/account/admin/enterprise paths. | Run generated drift checks if owner inputs, route shape, error codes, or schemas change. |
| Harness/test accounting | Phase 1 and phase 11 ledgers, phase maps, route inventory, task guides. | `docs/testing/*`, `internal/testutil/phase1test/inventory/routes.go`, `make task-guide`. | Treat as evidence accounting only; update owner inputs before generated ledgers if future changes require it. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| ------- | -------- | ---- | -------------- | -------------- | ------------------------ |
| No immediate owner contradiction found. | Core 01/Core 04, domain vocabulary, current route ownership, and tests align for auth/account/admin and enterprise auth. | low | intentional/no_action | auth/account/admin and enterprise-auth extension | Keep monitoring during implementation; mark any later document conflict as `BLOCKED: owner contradiction`. |
| `routes.go` was an oversized service and handler catch-all. | Handler code is now split into session, credential, account, user-admin, admin-audit, and shared helper files. | low | fixed | auth/account/admin facade with internal subareas | Preserve the split and keep future route families cohesive. |
| `authStore` was broad and crossed many responsibilities. | `authStore` now composes private responsibility interfaces for login, session, membership summaries, credential lifecycle, account, user admin, administrative audit, and enterprise auth. | low | fixed | auth facade over platform `authn` store | Keep new methods on the narrowest responsibility interface. |
| Enterprise auth shares base service state intentionally. | `enterprise_routes.go` uses shared session issuance, cookies, profile gating, and verifier dependencies through the same service facade. | medium | defer | enterprise-auth extension facade sharing session issuance primitive | Production claim enablement remains separate provider-manifest and real-verifier work. |
| Compatibility aliases remain in auth package where still useful. | `api.go` and `users.go` retain active aliases/helpers; `session_auth.go` was deleted after caller inventory. | low | fixed/defer | platform packages for primitives; auth package for stable local helper surface | Do not remove retained helpers until active package/testutil callers are migrated or deliberately preserved. |
| Session resource includes incident membership summaries. | `buildSessionResource` calls membership summary storage and clients consume `/auth/session`. | medium | intentional/no_action | auth session resource with incident membership data behind named interface | Preserve required wire member; treat it as informational bootstrap state, not authorization state. |
| WS revocation side effects are centralized behind auth helper actions. | Logout, admin revoke, credential changes, and enterprise binding actions now publish through `sessionRevocationPublisher`. | low | fixed | auth emits revocation; collaboration/platform WS handles socket semantics | Keep reason-code fanout centralized when adding new revocation causes. |
| Account preferences may be confused with workbook preferences. | `users.go` handles density mode; domain says account preference is deployment-local and not saved-view/workbook preference. | medium | intentional/no_action | auth/account facade | Preserve distinction in tracker and future implementation notes. |
| Platform imports are present in module logic. | Auth imports `authn`, `httpapi`, `httpauth`, `enterpriseauth`, `listquery`, and `pagination`. | medium | intentional/no_action | platform primitives consumed by module facade | No action unless domain logic is moved under transport/persistence packages. |
| Production auth does not directly own workbook/projection/revision/import/entity/evidence/link logic. | Target searches found no production imports or handlers for those owner areas under auth. | low | intentional/no_action | respective modules | Do not invent movement or classify auth as workbook orchestration. |
| Test-only SQL/storage assumptions exist. | Store/integration tests inspect sessions, audit, idempotency, and enterprise auth behavior through fixtures. | medium | defer | tests/harness evidence | Keep as characterization; avoid leaking test assumptions into production refactor design. |
| Deterministic enterprise verifier tests do not prove real IdP interoperability. | Phase 11 integration uses deterministic verifier overrides/testutil. | medium | defer | enterprise-auth extension and platform verifier adapters | Mark as evidence limitation, not current implementation defect. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| ----------- | ---- | -------------------------- | --------------------------- | ----------------------------- | ---- | --------------------- | ---------- | ------------------ |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01, WF-02 | Establish authority posture, inspect framework first, create tracker. | `docs/handoffs/auth-module-refactor-tracker-2.md` | `make lint-markdown` | Tracker exists with scope, authority, and no-production-refactor statement. |
| WF-01 | Auth package inventory | chain | WF-00 | WF-02, WF-04 | Inventory every file in `internal/modules/auth`. | All target files. | `make lint-markdown` for tracker; no code validation yet. | Inventory table names each file and risk. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Map routes, generated contracts, frontend callers, harness evidence, and owner docs. | Auth routes, OpenAPI/errors/extensions, frontend clients, phase inventory. | `make generated-artifact-policy-check`; `make json-shape-check` if contracts are touched later. | Freeze map identifies owners and tests for each public contract. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-06 | Determine where existing tests are enough and where future slices need tests first. | Phase 1/11 tests, frontend route-boundary tests, phase ledgers. | `make phase-slice PHASE=phase1`; `make phase-slice PHASE=phase11` for enterprise. | Test posture documented before any production split. |
| WF-04 | Boundary/coupling scan | chain | WF-01, WF-02 | WF-05 | Classify coupling findings by action level. | `routes.go`, `enterprise_routes.go`, `api.go`, `users.go`, `session_auth.go`. | Static search plus phase validation after changes. | Findings table has classification, owner, and action. |
| WF-05 | Facade or ownership redesign plan | chain | WF-04 | WF-06 | Define behavior-preserving internal cohesion splits without changing public contracts. | Auth service files and tests. | Phase 1 and phase 11 slices as applicable. | Proposed split plan preserves route registration and exported symbols. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07 | Sequence smallest safe implementation slices. | Tracker plus future production/test files. | Slice-specific Make targets. | Each slice has dependency, rollback, validation, and exit criteria. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Plan any needed test, harness, phase-map, or generated-accounting updates. | Phase inventory, ledgers, Make targets, contract files if changed. | Harness and generated drift targets. | Accounting updates are separated from runtime architecture. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Finalize evidence and handoff for implementation sessions. | Tracker and implementation diffs if later authorized. | `make agent-finalize`, `make test-fast`, `make check` as risk requires. | Handoff log updated with commands, results, blockers, and next action. |

## 7. Proposed Refactor Slice Plan

| Slice | Dependency | Exact intended change | Files or packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| ----- | ---------- | --------------------- | --------------------------------- | -------------- | ------------------------ | ------------------ | ------------- | -------------------- |
| S-00 | None | Create this tracker only. | `docs/handoffs/auth-module-refactor-tracker-2.md`. | None beyond documentation accuracy. | No tests added. Preserve no production behavior. | `make lint-markdown`. | Delete the tracker file. | Tracker contains all requested sections and repo-derived findings. |
| S-01 | S-00 | Characterize gaps before production edits. | Phase 1/11 auth tests, frontend route-boundary tests, harness inventory. | Missing characterization could hide route/envelope/session drift. | Add tests only for touched behavior not already covered. | `make phase-slice PHASE=phase1`; `make phase-slice PHASE=phase11` for enterprise. | Remove new characterization tests if no implementation follows. | Each planned production slice has an existing or new test posture. |
| S-02 | S-01 | Split `routes.go` by cohesive handler area without changing routes, service construction, exported symbols, or response shapes. | `internal/modules/auth/routes.go` plus new same-package files. | Route registration order, cookie/CSRF behavior, session sliding, idempotency, admin authorization, WS revocation. | Preserve phase 1 handler/integration/support tests. | `make phase-slice PHASE=phase1`; broaden to `make service-backed-slice PHASE=phase1` if store paths are touched. | Revert file split to previous `routes.go`. | Same public routes and tests pass. |
| S-03 | S-01 | Split pure decoder/resource helpers while preserving exported function/type names. | `api.go`, `users.go`, possible new helper files. | Request validation, defaulting, error codes, safe user/account resource shape. | Preserve support/request/handler tests; add targeted decoder tests for any moved branch not covered. | `make phase-slice PHASE=phase1`. | Restore helper definitions to original files. | Exported API and wire behavior remain unchanged. |
| S-04 | S-02, S-03 | Narrow broad store/service seams by responsibility while keeping platform store implementation unchanged. | `routes.go`, `enterprise_routes.go`, tests and stubs. | Store-backed behavior, idempotency, audit, session concurrency, enterprise binding revocation. | Preserve `phase1_store_test.go`, support-integration sweeps, and phase 11 integration tests. | `make service-backed-slice PHASE=phase1`; `make service-backed-slice PHASE=phase11` if enterprise touched. | Recombine sub-interfaces into the prior `authStore` shape. | No storage or observable route behavior changes. |
| S-05 | S-04 | Migrate or remove compatibility aliases only after caller inventory proves safe. | `session_auth.go`, `api.go`, `users.go`, `internal/testutil/phase0test`, any external package imports found by `rg`. | Go source compatibility for tests and internal consumers. | Add or preserve compile coverage for migrated callers. | `make test-fast`; `make phase-slice PHASE=phase1`. | Restore aliases/wrappers. | No remaining unauthorized caller depends on removed auth symbols. |
| S-06 | S-04 | Review enterprise-auth adapter boundary without changing production verifier behavior. | `enterprise_routes.go`, `enterprise_protocol.go`, platform `enterpriseauth`, phase 11 tests. | Provider discovery, redirects, transaction cookie, profile gating, binding idempotency, target session revocation. | Preserve phase 11 tests; add characterization before any real verifier adapter change. | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11`. | Restore previous verifier alias/override wiring. | Enterprise route behavior and deterministic tests remain stable. |
| S-07 | S-02 through S-06 as needed | Update owner inputs and regenerate only if public routes, schemas, error codes, extension profiles, or harness accounting intentionally change. | Owner specs/contracts, generated protocol outputs, phase maps. | Generated contract drift, client incompatibility, harness accounting drift. | Preserve protocol and frontend route-boundary tests. | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make frontend-import-boundary-check`. | Revert owner-input and generated changes together. | No generated drift remains unless intentionally committed from owner inputs. |
| S-08 | S-07 | Final validation and handoff after an implementation sequence. | Changed implementation/test/docs files. | Missed broad regression. | Preserve all previous characterization. | `make agent-finalize`; `make test-fast`; `make check` when risk warrants. | Revert final slice or offending implementation slice. | Handoff log has commands, results, artifacts, and remaining blockers. |

## 7A. Execution Remediation Control Plan

This section is the controlling execution plan for the remediation session that
started on 2026-07-01T08:30:31-04:00. Each workstream MUST update this tracker
after completion and before the next workstream starts. Updates MUST record
status, touched files, validation commands, results, blockers, and next action.

Execution posture:

- Public route paths, response envelopes, cookies, CSRF behavior, idempotency,
  authorization, extension-profile gating, and generated contracts stay stable
  unless an owner-spec change explicitly precedes implementation.
- `enterprise_authentication` remains unclaimed by default until production
  provider-manifest startup validation and real OIDC/SAML verification exist.
- Deterministic enterprise-auth verifiers remain harness evidence only.
- Generated roots and generated ledgers remain downstream artifacts and MUST NOT
  be hand-edited.

### 7A.1 Gap Remediation Matrix

| Gap ID | Gap | Remediation | Areas | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation |
| ------ | --- | ----------- | ----- | ----------------- | --------------------------------- | ------------------ | ---------- |
| GAP-AUTH-001 | Tracker lacks execution-grade remediation state. | Add this matrix, workstream status, and handoff checkpoints. | documentation | Prevents slice drift across sessions. | Docs-only. | Later agents skip sequencing or repeat discovery. | `make lint-markdown`. |
| GAP-AUTH-002 | Enterprise-auth claim posture is easy to misread. | Clarify authored phase inputs that direct evidence exists while the profile remains unclaimed by default; regenerate or drift-check downstream outputs. | specification, tests, documentation | Safer conformance accounting. | Update owner inputs only; generated artifacts remain downstream. | Premature or underclaimed enterprise-auth status. | `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift`. |
| GAP-AUTH-003 | `routes.go` is a catch-all. | Split same-package handlers into auth/session, credential lifecycle, account, deployment users, admin audit, and shared auth/session helpers. | implementation, tests | Smaller review units and cleaner future phase expansion. | No wire change. | Fragile edits and accidental cross-route regressions. | `make phase-slice PHASE=phase1`. |
| GAP-AUTH-004 | `authStore` is too broad. | Replace it with private responsibility interfaces while keeping `authn.Store` as the backing implementation. | implementation, tests | Easier unit stubs and safer extension growth. | Test stubs change; storage behavior stays stable. | Every handler can depend on every storage concern. | `make service-backed-slice PHASE=phase1`; phase 11 if enterprise touched. |
| GAP-AUTH-005 | Enterprise auth is coupled to the base auth service. | Keep shared session issuance, but isolate enterprise profile gating, protocol handlers, binding handlers, and verifier dependencies as a cohesive extension facade. | implementation, tests | Cleaner claim enablement and verifier replacement later. | No route, cookie, or redirect behavior change. | Extension behavior remains hard to reason about. | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11`. |
| GAP-AUTH-006 | Compatibility aliases and wrappers create accidental API surface. | Inventory callers, migrate repo-local uses to platform APIs or local helpers, delete unused wrappers, and unexport helpers where safe. | implementation, tests | Reduces maintenance burden and false ownership. | `internal/testutil/phase0test` may need migration if exported helper cleanup touches it. | Future code imports auth as a platform primitive. | `make test-fast`; `make phase-slice PHASE=phase1`. |
| GAP-AUTH-007 | Session resource includes membership summaries. | Preserve the required wire member, isolate membership-summary reads behind a named private interface, and document it as informational bootstrap state. | implementation, documentation, tests | Clear boundary between bootstrap display and authorization. | No wire change. | Developers may treat session memberships as authorization source. | Existing session tests plus `make phase-slice PHASE=phase1`. |
| GAP-AUTH-008 | WebSocket revocation side effects are hidden. | Narrow auth's hub dependency to a revocation publisher and centralize reason-code fanout. | implementation, tests, documentation | Easier audit of session-loss causes. | No socket event change. | Missed revocation on credential or admin changes. | Route revocation tests and WS support integration. |
| GAP-AUTH-009 | Account preferences can be confused with workbook preferences. | Keep account preferences under account handlers and preserve terminology guardrails. | documentation, implementation, tests | Prevents wrong ownership moves. | No behavior change. | Portability or saved-view coupling leaks in later phases. | Account preference tests and import-boundary scan. |
| GAP-AUTH-010 | Test-only SQL/storage assumptions can leak into design. | Keep SQL fixture checks as characterization only and prefer route-contract assertions for behavior. | tests | Cleaner tests after seam narrowing. | Test refactor only. | Implementation starts matching fixtures instead of contracts. | Phase 1 and phase 11 slices stay green. |
| GAP-AUTH-011 | Enterprise deterministic verifiers do not prove production interoperability. | Keep profile unclaimed and track provider-manifest config plus real verifier adapters as future claim-enablement work. | specification, implementation, tests, documentation | Secure future enterprise rollout. | No current route claim flip. | Unsafe conformance claim or unusable production enterprise auth. | Config/startup tests, phase 11 slices, final claim gate when implemented. |
| GAP-AUTH-012 | Generated contract drift risk. | Update owner inputs first, regenerate only through Make targets, and never hand-edit generated roots. | specification, contracts, tests | Keeps clients and contracts consistent. | Expected none for structural refactor. | Silent protocol/client divergence. | `make generate-drift`; `make generated-artifact-policy-check`. |
| GAP-AUTH-013 | `.gitkeep` is obsolete. | Remove after code split confirms the directory remains populated. | implementation | Small repository cleanup. | None. | Minor repository noise. | `git status`; ordinary tests. |

### 7A.2 Execution Workstreams

| Workstream | Depends on | Required work | Tracker checkpoint | Exit criteria | Status |
| ---------- | ---------- | ------------- | ------------------ | ------------- | ------ |
| WS-00 | None | Add this execution plan, remediation matrix, and update-before-next-workstream rule. | Record baseline in Section 10. | Tracker is current and lintable. | DONE |
| WS-01 | WS-00 | Clarify enterprise-auth direct-evidence versus unclaimed-default wording in authored phase inputs; regenerate or drift-check downstream outputs. | Record owner inputs changed and drift results. | Phase guide/ledger wording no longer implies missing direct evidence or a production claim. | DONE |
| WS-02 | WS-01 | Map every touched behavior to existing tests; add narrow tests only for uncovered behavior before production edits. | Update `AUTH-TRACK-007`. | Each implementation slice has a named test posture. | DONE |
| WS-03 | WS-02 | Split `routes.go` by cohesive handler area while keeping route registration order and response behavior stable. | Update `AUTH-TRACK-008`. | Phase 1 behavior is unchanged. | DONE |
| WS-04 | WS-03 | Split decoders, resource builders, errors, cookies, session resources, and idempotency helpers into cohesive files; keep needed symbols until cleanup. | Update `AUTH-TRACK-009`. | Decoder/resource tests remain stable. | DONE |
| WS-05 | WS-04 | Introduce narrowed private store interfaces and a revocation publisher; update tests/stubs. | Update `AUTH-TRACK-010`. | Store-backed auth and enterprise revocation behavior remains stable. | DONE |
| WS-06 | WS-05 | Remove unused platform aliases/wrappers, migrate repo callers, delete obsolete `.gitkeep` if appropriate. | Resolve `AUTH-TRACK-011`. | No remaining unauthorized caller depends on removed auth symbols. | DONE |
| WS-07 | WS-06 | Keep deterministic enterprise tests as evidence only; document production claim-enablement follow-up unless implementing provider manifest and real verifiers. | Update `AUTH-TRACK-012` and `AUTH-TRACK-013` if owner/generator changes occur. | Enterprise routes stay gated when unclaimed and pass when test profile override claims them. | DONE |
| WS-08 | WS-07 | Run final validation and update handoff. | Mark `AUTH-TRACK-014` done. | Handoff is complete and no required validation is unexplained. | DONE |

### 7A.3 Characterization Mapping For Implementation Slices

No new tests are required before the planned behavior-preserving structural
splits. Existing tests cover the observable behavior that each slice may touch.
If a later edit changes public route behavior rather than only moving code, add
a focused characterization test before that behavior change.

| Planned slice | Existing characterization evidence | Coverage decision |
| ------------- | ---------------------------------- | ----------------- |
| Auth/session handler split | `U-1-01` through `U-1-05`, `I-1-01`, `E-1-01`, `E-1-03`, `E-1-04`; support tests for session helpers. | Existing coverage is sufficient for code movement. |
| Credential lifecycle split | `U-1-10` through `U-1-12`, `I-1-04`, `E-1-02`, `E-1-06`, `E-1-07`; support tests for password/TOTP helpers. | Existing coverage is sufficient for decoder/helper movement. |
| Account profile and preferences split | Phase 1 handler/support-integration account route sweeps plus Core 01 account preference tests through app-shell callers. | Existing route and request-contract sweeps are sufficient unless wire shape changes. |
| Deployment user/admin split | `U-1-07`, `U-1-08`, `U-1-13`, `I-1-03`, `I-1-05`, `E-1-05`, `E-1-08`, `E-1-12`. | Existing coverage is sufficient for code movement and seam narrowing. |
| Administrative audit split | Phase 1 support-integration surface, audit attribution, authorization, and request-contract sweeps; `I-1-03`, `I-1-05`. | Existing coverage is sufficient unless filter/resource shape changes. |
| Store seam narrowing | `phase1_handlers_test.go` stub coverage, `phase1_store_test.go`, Phase 1 support-integration route sweeps. | Existing coverage is sufficient; stubs will be updated to match narrower interfaces. |
| Revocation publisher narrowing | `U-1-05`, `U-1-06`, `I-1-02`, support-integration session revocation, phase 11 binding lifecycle. | Existing coverage is sufficient; no event-shape change planned. |
| Enterprise-auth facade split | `I-11-ENTERPRISE-AUTH-01`, `I-11-ENTERPRISE-AUTH-02`, `I-11-ENTERPRISE-AUTH-03`, frontend/browser enterprise rows. | Existing deterministic evidence is sufficient for structure; production IdP interoperability remains future claim-enablement work. |

### 7A.4 Enterprise Authentication Claim Gate

Enterprise Authentication remains unclaimed by default after this remediation.
The current route and binding implementation has deterministic harness evidence,
but production claim enablement requires all of the following future work before
the profile can be claimed:

- Add config/runtime support for
  `enterprise_authentication.provider_manifest_path`.
- Validate provider manifests during production startup, including disabled,
  missing, malformed, duplicate, and provider-key mismatch cases.
- Add production OIDC and SAML verifier adapters backed by real provider
  metadata, key material, issuer/audience checks, response freshness, relay
  state, and subject binding validation.
- Reconcile manifest-backed provider definitions without adding browser
  provider-definition mutation APIs.
- Publish an owner-spec claim change before flipping generated or runtime
  conformance accounting.

Deterministic OIDC/SAML verifiers under the test harness remain implementation
evidence only. They are not interoperability or security evidence for a
production extension-profile claim.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| ---------------- | ------- | ----- | ------------------------------- | ----- |
| unit | `make test-fast` | Broad fast backend/frontend test surface. | no | Use as a broader fast gate after code changes; not required for docs-only tracker creation. |
| integration | `make service-backed-slice PHASE=phase1` | Service-backed auth/session/account/admin behavior. | yes, for behavior-affecting auth changes | Pair with `make phase-slice PHASE=phase1` when route/helper behavior changes. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser-visible app/session route behavior. | no | Required if session route shape, login flow, account/admin UI behavior, or enterprise begin UI behavior changes. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift` | Generated roots, JSON shape, and generated artifact drift. | no | Required if owner inputs/contracts/harness manifests change; not expected for docs-only tracker. |
| import-boundary/static | `make lint`; `make frontend-import-boundary-check`; `make frontend-typecheck` | Static and frontend import/type boundaries. | no | Required for package split, exported symbol cleanup, or frontend contract changes. |
| full check | `make agent-finalize`; `make check` | End-of-run hygiene and full repository check. | no | Run when implementation risk is broad. For retained successful runs, pass `RESULTS_DIR` to `make agent-finalize`; otherwise record skip reason. |

Docs-only tracker validation:

| Validation layer | Command | Scope | Required before implementation? | Notes |
| ---------------- | ------- | ----- | ------------------------------- | ----- |
| documentation | `make lint-markdown` | Auth tracker Markdown formatting. | yes | Narrow validation for creating this file. |

Enterprise-specific validation:

| Validation layer | Command | Scope | Required before implementation? | Notes |
| ---------------- | ------- | ----- | ------------------------------- | ----- |
| enterprise auth | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11` | Enterprise provider and auth-binding behavior. | yes, before enterprise-auth production edits | Phase 11 task guide notes enterprise-auth evidence rows, but deterministic verifier coverage is not real IdP interoperability proof. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| AUTH-TRACK-001 | Create planning-only tracker file. | WF-00 | DONE | None | This file. | File exists at `docs/handoffs/auth-module-refactor-tracker-2.md`. |
| AUTH-TRACK-002 | Confirm framework/domain/source authority posture. | WF-00 | DONE | AUTH-TRACK-001 | Framework, Core 00/Core 01/Core 04, domain, harness NLSpec inspection. | Source hierarchy is recorded and no auth-specific NLSpec is assumed. |
| AUTH-TRACK-003 | Inventory every target file. | WF-01 | DONE | AUTH-TRACK-002 | Section 2 inventory table. | Each file in `internal/modules/auth` is listed or explicitly out of scope. |
| AUTH-TRACK-004 | Diagnose workbook boundary. | WF-01 | DONE | AUTH-TRACK-003 | Section 3 diagnosis. | Auth is recorded as not owning workbook orchestration. |
| AUTH-TRACK-005 | Map public contracts and generated surfaces. | WF-02 | DONE | AUTH-TRACK-003 | Section 4 freeze map. | Every discovered auth/account/admin/enterprise/WS-adjacent risk has owner and test posture. |
| AUTH-TRACK-006 | Classify coupling findings. | WF-04 | DONE | AUTH-TRACK-005 | Section 5 findings. | Findings have classification, risk, proposed owner, and planning action. |
| AUTH-TRACK-007 | Add characterization tests for any future touched gaps. | WF-03 | DONE | AUTH-TRACK-005 | Section 7A.3 characterization mapping. | Every touched behavior has an existing or new characterization test. |
| AUTH-TRACK-008 | Split route handlers by cohesive area. | WF-05 | DONE | AUTH-TRACK-007 | `routes.go` reduced to service setup and registration; added `session_handlers.go`, `credential_handlers.go`, `account_handlers.go`, `admin_audit_handlers.go`, and `user_admin_handlers.go`; `make phase-slice PHASE=phase1` pass at `.cartulary/test-results/20260701T123554Z-p1063697`. | Public route behavior and phase 1 tests remain stable. |
| AUTH-TRACK-009 | Split pure helper files while preserving exported symbols. | WF-05 | DONE | AUTH-TRACK-007 | Added `auth_session_helpers.go` for auth/session cookies, session resource, idempotency hashes, stored responses, sliding, and API-error forwarding; `make phase-slice PHASE=phase1` pass at `.cartulary/test-results/20260701T123554Z-p1063697`. | Decoder/resource helper tests remain stable. |
| AUTH-TRACK-010 | Narrow broad store/service seams. | WF-05 | DONE | AUTH-TRACK-008, AUTH-TRACK-009 | `authStore` is now composed from private responsibility interfaces; session membership summaries are isolated as informational bootstrap state; WebSocket dependency is narrowed to `sessionRevocationPublisher`; direct revocation calls fan out through shared helpers. Validation passed: `make service-backed-slice PHASE=phase1` at `.cartulary/test-results/20260701T123839Z-p1085638`; `make service-backed-slice PHASE=phase11` at `.cartulary/test-results/20260701T123925Z-p1104168`. | Store-backed phase 1 and phase 11 behavior remains stable. |
| AUTH-TRACK-011 | Decide compatibility alias cleanup. | WF-06 | DONE | AUTH-TRACK-010 | Removed unused `session_auth.go` platform wrapper and obsolete `internal/modules/auth/.gitkeep` after focused caller inventory found only the wrapper file used `AuthenticateSessionRequest`, `SessionAuthOptions`, `SessionAuthStore`, and `SessionRevoker`. Retained `APIError`, `AuthSource`, and safe user resource helpers because package tests and `internal/testutil/phase0test` still have active repo-local callers. Validation passed: `make phase-slice PHASE=phase1` at `.cartulary/test-results/20260701T124024Z-p1116788`; `make test-fast` at `.cartulary/test-results/20260701T124108Z-p1135775`. | No remaining unauthorized caller depends on removed auth symbols. |
| AUTH-TRACK-012 | Review enterprise auth adapter boundary. | WF-06 | DONE | AUTH-TRACK-010 | Section 7A.4 records the production claim gate; deterministic enterprise verifier evidence stays harness-only. Validation passed: `make phase-slice PHASE=phase11` at `.cartulary/test-results/20260701T124355Z-p1186484`. | Deterministic tests remain stable and evidence limits are documented. |
| AUTH-TRACK-013 | Update generated contracts only from owner inputs if needed. | WF-07 | DONE | Any route/schema/error/profile change | No public route/schema/error/profile changes were made by the structural refactor, so generated contracts were not regenerated. WS-01 owner-input phase accounting changes were made through `tools/phase11_test_map.json` and `tools/phase_registry.json`, then regenerated/drift-checked with `make phase-ledgers`, `make phase-schedules`, `make json-shape-check`, `make phase-ledger-drift`, and `make phase-schedule-drift`. | No generated contract drift remains after intentional owner-input changes. |
| AUTH-TRACK-014 | Run final validation and handoff. | WF-08 | DONE | Implementation slices | Final validation passed: `make agent-finalize` at `.cartulary/test-results/20260701T124458Z-p1199419`; `make lint-markdown`; `make json-shape-check` at `.cartulary/test-results/20260701T124524Z-p1201248`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260701T124527Z-p1201573`; `make generate-drift` at `.cartulary/test-results/20260701T124529Z-p1201759`; `make phase-ledger-drift` at `.cartulary/test-results/20260701T124534Z-p1202744`; `make phase-schedule-drift` at `.cartulary/test-results/20260701T124539Z-p1203100`; `make test-fast` at `.cartulary/test-results/20260701T124542Z-p1203279`; `make check` at `.cartulary/test-results/20260701T124752Z-p1249751`. `make agent-finalize` retained-run maintenance was skipped because `RESULTS_DIR` was unset. | Commands, results, skipped checks, and blockers are recorded. |
| AUTH-TRACK-015 | Add execution-grade remediation matrix and workstream checkpoints. | WS-00 | DONE | AUTH-TRACK-006 | Section 7A. | Tracker names gaps, sequencing, validation, and update-before-next-workstream rule. |

## 10. Session Handoff Log

### Scope and Authority Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-07-01T08:19:08-04:00 | Codex implementation session | Planning-only tracker created from live repo inspection and prior planning pass. | Inspected framework, prior auth tracker, domain, Core 00/Core 01/Core 04 excerpts, harness NLSpec; touched this tracker. | `rg`, `sed`, `find`, `wc`, `git status --short`, `make help`, `make task-guide ROLE=feature-dev PHASE=phase1`, `make task-guide ROLE=feature-dev PHASE=phase11`. | Source hierarchy and no-production-refactor posture recorded. | None for tracker creation. | Use this tracker before authorizing production refactor. |
| 2026-07-01T08:30:31-04:00 | Codex implementation session | WS-00 remediation baseline added. | Touched this tracker. | `sed`, `date -Iseconds`. | Section 7A now records gap matrix, workstream sequence, and tracker-update rule. | None. | Start WS-01 authored phase-input posture cleanup. |

### Backend Module Boundary Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-07-01T08:19:08-04:00 | Codex implementation session | Auth is recorded as auth/account/admin/enterprise facade, not workbook orchestration. | All production files under `internal/modules/auth`; `internal/app/runtime.go`; caller searches. | `find internal/modules/auth`; `rg "workbook|saved_view|view_schema|projection|revision|imports|entities|evidence|links"`; import/caller `rg`. | No direct production ownership of workbook/projection/revision/import/entity/evidence/link behavior found. | None. | If implementing, start with characterization and cohesion splits, not ownership moves. |

### Contract and Codegen Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-07-01T08:19:08-04:00 | Codex implementation session | Auth/account/admin/enterprise routes and generated surfaces are mapped as freeze contracts. | `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, `contracts/extensions/index.json`, generated protocol references, frontend route callers. | Contract and frontend `rg`; targeted `sed` on contract files. | Public route families and generated contract risks recorded. | None for docs-only tracker. | Run drift checks only if future owner inputs or contracts change. |
| 2026-07-01T08:32:00-04:00 | Codex implementation session | WS-01 enterprise-auth posture cleanup completed. | Touched `tools/phase11_test_map.json`, `tools/phase_registry.json`; generated `docs/testing/phase11_coverage_ledger.md` and `tools/execution_topology_render_index.json`; touched this tracker. | `make phase-ledgers`; `make phase-schedules`; `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift`; `git status --short`. | Authored phase inputs now say Enterprise Authentication has direct deterministic implementation evidence while production profile remains unclaimed by default; all drift checks pass. | None. | Start WS-02 characterization mapping before production code splits. |

### Tests and Harness Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-07-01T08:19:08-04:00 | Codex implementation session | Phase 1 and phase 11 tests are identified as the primary auth characterization net. | All auth test files, phase 1 route inventory, phase ledgers/maps, Make task guide output. | `rg "^func Test"` in auth tests; `sed` route inventory excerpts; `make task-guide` for phase 1 and phase 11. | Validation commands and accounting posture recorded. | TODO: future slices must decide whether touched behavior needs new characterization tests. | Before production edits, map each touched contract to an existing or new test. |
| 2026-07-01T08:33:00-04:00 | Codex implementation session | WS-02 characterization mapping completed. | Touched this tracker; inspected auth tests, phase ledgers, and route inventory references. | `rg "^func Test"`; targeted `sed` of phase 1 and phase 11 ledgers; `rg` for route inventory, CSRF, revocation, and enterprise route evidence. | Existing Phase 1 and Phase 11 tests cover the planned behavior-preserving code movement; no new tests added before structural split. | None. | Start WS-03 route/service cohesion split. |
| 2026-07-01T08:35:54-04:00 | Codex implementation session | WS-03 and WS-04 route/helper split completed. | Touched `internal/modules/auth/routes.go`, `session_handlers.go`, `credential_handlers.go`, `account_handlers.go`, `admin_audit_handlers.go`, `user_admin_handlers.go`, and `auth_session_helpers.go`. | `make format`; `make phase-slice PHASE=phase1` after fixing one split-only unused import. | Phase 1 slice passed with 9/9 work units and 71 tests at `.cartulary/test-results/20260701T123554Z-p1063697`; the earlier failed import-only run was `.cartulary/test-results/20260701T123527Z-p1057339`. | None. | Start WS-05 seam narrowing and revocation publisher cleanup. |
| 2026-07-01T08:39:25-04:00 | Codex implementation session | WS-05 seam narrowing and revocation publisher cleanup completed. | Touched `internal/modules/auth/routes.go`, `auth_session_helpers.go`, `session_handlers.go`, `credential_handlers.go`, `user_admin_handlers.go`, `enterprise_routes.go`, and `phase1_handlers_test.go`. | `make format`; `make service-backed-slice PHASE=phase1`; `make service-backed-slice PHASE=phase11`. | Phase 1 service-backed slice passed with 7/7 work units and 49 tests at `.cartulary/test-results/20260701T123839Z-p1085638`; Phase 11 service-backed slice passed with 2/2 work units and 32 tests at `.cartulary/test-results/20260701T123925Z-p1104168`. | None. | Start WS-06 compatibility wrapper and placeholder cleanup. |
| 2026-07-01T08:43:27-04:00 | Codex implementation session | WS-06 compatibility cleanup completed. | Deleted `internal/modules/auth/session_auth.go` and `internal/modules/auth/.gitkeep`; inspected active callers for wrapper aliases and retained helpers. | Focused `rg` for `AuthenticateSessionRequest`, `SessionAuthOptions`, `SessionAuthStore`, and `SessionRevoker`; `find internal/modules/auth -name .gitkeep`; `make phase-slice PHASE=phase1`; `make test-fast`. | Phase 1 slice passed with 9/9 work units and 71 tests at `.cartulary/test-results/20260701T124024Z-p1116788`; `make test-fast` passed with 2/2 work units and 970 tests at `.cartulary/test-results/20260701T124108Z-p1135775`. | None. | Start WS-07 enterprise boundary posture review and final claim-enablement notes. |
| 2026-07-01T08:43:55-04:00 | Codex implementation session | WS-07 enterprise boundary posture review completed. | Touched this tracker; reviewed existing Phase 11 posture and claim-gate requirements. | `make phase-slice PHASE=phase11`. | Phase 11 slice passed with 4/4 work units and 50 tests at `.cartulary/test-results/20260701T124355Z-p1186484`; no production provider-manifest or real verifier work was implemented, so Enterprise Authentication remains unclaimed by default. | None. | Start WS-08 final validation and handoff completion. |
| 2026-07-01T08:50:19-04:00 | Codex implementation session | WS-08 final validation and handoff completed. | Touched this tracker; final diff includes auth file split, narrowed auth store/revocation seams, wrapper and placeholder deletion, and phase-accounting owner/generated outputs. | `make agent-finalize`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check`; `make generate-drift`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make test-fast`; `make check`. | All final validation passed. `make check` passed with 276/276 work units and 973 tests at `.cartulary/test-results/20260701T124752Z-p1249751`. `make agent-finalize` passed at `.cartulary/test-results/20260701T124458Z-p1199419`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. | None. | Remediation effort complete; future work is limited to separate Enterprise Authentication production claim enablement. |

### Security and Authorization Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-07-01T08:19:08-04:00 | Codex implementation session | Session required, CSRF, bootstrap-token limits, deployment-admin checks, current-account guards, and enterprise profile gating are contract risks. | `api.go`, `routes.go`, `users.go`, `enterprise_routes.go`, Core 04 authorization/session excerpts. | Targeted `rg` for auth, CSRF, bootstrap, deployment_admin, session, enterprise provider terms. | Authorization outcomes are frozen as observable behavior. | None for tracker. | Any implementation touching auth checks must run phase 1 and relevant phase 11 validation. |

### Open Risks and Next Session Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-07-01T08:50:19-04:00 | Codex implementation session | Original risks for oversized route/service files, broad store seam, unused session wrapper, and hidden WS revocation fanout are remediated; enterprise production evidence limits remain. | This tracker and implementation diffs under `internal/modules/auth`. | Final validation set in AUTH-TRACK-014. | Remediation complete with stable public auth contracts. | None for this remediation. | Future work is Enterprise Authentication production claim enablement only: provider manifest config/startup validation and real OIDC/SAML verifier adapters before any claim flip. |

## 11. Open Questions and Blockers

No blocking open questions were found for creating this planning tracker. Future
implementation may uncover blockers when a concrete slice changes code or
contracts.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| -- | ------------------- | -------------- | ---------------------------- | -------------- |

## 12. Binary Completion Criteria

This tracker is complete when all of the following are true:

- Every file in `internal/modules/auth` is inventoried or explicitly out of
  scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every implementation slice is behavior-preserving unless explicitly marked as
  requiring later authorization.
- Validation commands are discovered or marked as `TODO:` with reason.
- Handoff sections are current enough for another agent to continue without
  rediscovery.

Completion status for this tracker:

- Files inventoried: complete for the current target directory listing.
- Public contract risks mapped: complete for discovered auth/account/admin,
  enterprise-auth, generated-contract, WS-adjacent, and workbook-indirect
  contracts.
- Proposed workflows and slices: complete as planning guidance.
- Validation commands: discovered from Make-owned public targets and task guide
  output.
- Production refactor: completed for the authorized behavior-preserving auth
  module remediation; public HTTP contracts remained stable.
- Tracker state: complete and current through WS-08.
- Output path: `docs/handoffs/auth-module-refactor-tracker-2.md`.
- Remaining blockers: none for this remediation. Enterprise Authentication
  production claim enablement remains future work requiring provider-manifest
  startup validation and real OIDC/SAML verifier adapters before any claim flip.
