# Reference Pack regulated production-readiness plan

This is the sole controlling tracker for the next Reference Pack and application cryptography iteration. It is implementation support, not a normative owner, adoption record or conformance claim. Tests, generators, runtime metadata and release evidence must not read, stat, hash or otherwise depend on this document.

## 1. Scope and source posture

**Current authorization is document-only.** The user approved the plan below for recording in this tracker on 2026-10-05. Production implementation requires a later authorized task. All new implementation workstreams S16–S25 are TODO. The completed structural and package iterations remain DONE; their plans, failures and original validation identities are retained as historical records below.

| Scope fact | Decision |
| --- | --- |
| Current baseline | Existing working tree based on `e7a1497ff9e8219f7ccdad7a09af9e356cb146ec`, inspected 2026-10-05, including completed remediation changes and other pre-existing edits. This is **not a clean checkout**. Preserve all pre-existing changes. |
| Allowed change now | Update only `docs/handoffs/reference-pack-remediation.md`, append the planning-session record, run `make lint-markdown` and `git diff --check`. No implementation, generation, product verification, deployment or retained-state mutation. |
| Target and growth constraint | Preserve the 197-file Reference Data/assembly inventory; extend planning across the application, authentication, sealing, recovery, integration and build boundaries in §2. Keep domain formats and authorization with their owners so later phases do not require a shared domain-token or provider framework. |
| S07 disposition | Expand the formerly deferred validated-cryptography capability into S16–S25. S07 is a TODO umbrella, not another implementation slice. Its old DEFERRED records remain historical. |
| Claim boundary | Security cryptographic operations performed by `server`, `migrate` and `operator`, including enabled optional integrations. Database servers, object-storage servers, external TLS terminators, browsers and host certification remain separately owned deployment dependencies. Application-side transport and certificate verification remain in scope. |
| Production policy | One supported cryptographic implementation and release policy. No standard-versus-regulated algorithm branches, crypto provider plug-in framework, silent fallback or security-operation bypass. |
| Initial module candidate | Go Cryptographic Module `v1.0.0-c2097c7c` in the repository's pinned Go `1.27.1` toolchain; CMVP certificate `5247`. Archive SHA-256: `daf3614e0406f67ae6323c902db3f953a1effb199142362a039e7526dfb9368b`. Verify and pin exact archive/build identity; never use floating `latest`, `certified` or `inprocess` selectors. Certificate and security disposition must be rechecked before qualification. |
| Reference environment | Linux/amd64, Debian 12, using the existing digest-pinned application image. Record host OS/kernel, container runtime, CPU/features, image/index and resolved platform image identities; establish their relationship to the module's permitted operating environments before qualification. A Debian-derived image name alone is insufficient. |
| Execution mode | Build against the pinned module and require enabled FIPS mode at startup. Use `fips140=only` for targeted diagnostic assessment, not production deployment or sole compliance evidence; it can produce false positives and negatives. |
| Compatibility | Fresh regulated deployments only. Reject incompatible application state and backup formats before mutation. Historical deployments and backups stay with their matching releases. No password conversion, legacy cryptography fallback, dual writers or automatic reset. |
| Reference Pack signatures | Preserve Ed25519, canonical bytes, trust rotation and `tuf_1_0_35_offline_bundle_v1`. Amend the owner's FIPS exclusion to distinguish execution qualification from signature semantics. The selected module includes EdDSA; a second signature algorithm or method would add unnecessary format/trust complexity. |
| Permitted non-security use | Explicitly reviewed protocol-only hashes such as the WebSocket handshake hash may remain. They must never provide authentication, integrity protection, key derivation or authorization. Indicator MD5/SHA-1 strings remain domain data. An exception is not an enforcement bypass. |
| Non-goals | No high availability, new pack profiles, external corpus distribution, unrelated module reshuffling, UI redesign or automatic customer rollout. Do not reopen completed runtime/test separation or recreate retired reporting dossiers. |
| Authority | Adopted subsystem NLSpecs and normative Core sections govern their scopes; typed contracts are downstream projections. Verification routing belongs to its machine owners, not this tracker. `docs/domain.md` owns vocabulary/navigation and `docs/design.md` supplies design direction. Owner adoption, package qualification and customer deployment approval are separate decisions. |

The candidate and constraints above follow the [NIST certificate](https://csrc.nist.gov/projects/cryptographic-module-validation-program/certificate/5247), [module security policy](https://csrc.nist.gov/CSRC/media/projects/cryptographic-module-validation-program/documents/security-policies/140sp5247.pdf), and [Go FIPS guidance](https://go.dev/doc/security/fips140). They define a qualification target, not a claim that the current package is qualified. Password parameters below follow the [OWASP password-storage guidance](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html). The [WebSocket security discussion](https://www.rfc-editor.org/rfc/rfc6455#section-10.8) informs the narrow protocol exception.

The existing image pin is `gcr.io/distroless/base-debian12:nonroot@sha256:7f0c72cd138b442ae0deeb69c08b1acf5525439ba251a49ad93c320a061567e5` in `tools/toolchain_pins.json`. It is a reproducibility input; S24 must establish actual execution-environment suitability.

The previous complete release remains **historical passing evidence**, not regulated qualification: `.cartulary/test-results/20261005T070744Z-p19802` passed 1264/1264 units, zero failures/skips/cancellations, all 17 readiness projections, and 665 cleanup steps without failure or block. Its manifest records the dirty candidate based on `e7a1497ff9e8219f7ccdad7a09af9e356cb146ec`, source digest `sha256:8ae8e9ab302a68ca099c6fb203158c202f25f477ac130330c334b0d30cf05fce` and graph digest `sha256:b2e3a5199571067bcca2addf20247eddbc46980e950a25b064e3a4f6640f40c7`. Do not relabel that run, the earlier failed release `20261005T061847Z-p6495`, or the older passing `20261005T030730Z-p45700` as evidence for a future candidate.

The four prior behavioral corrections remain regression obligations: strict composition identities, framework lookup-key deduplication, typed early upload rejection and pre-parse URL admission. Completed module boundaries, canonical identity checks, export-root recovery binding and deterministic presence capture remain intact.

## 2. Current-state repository inventory

The previous clean baseline contained **186 files**. The current direct target contains **197 files**, all Go, inventoried below and checked against the filesystem with no missing or stale entries. S02–S06 added thirteen direct-target files and removed `api.go`, `routes.go`, and `worker_hooks.go` (net +10); S13 added the fixture/runtime boundary regression (net +1). The HTTP adapter and request files are additions, not deletions. Adjacent additions include the instance barrier in `internal/testutil/appsupport/reference_pack_barrier.go`; it is intentionally outside this count.

| Path / inventory group | Current responsibility | Public surface | Inbound callers | Outbound dependencies | Tests touching it | Contracts / generated artifacts | Target owner and disposition | Risk / notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Reference Data root: API and routes | HTTP decoding, resources, authorization, pagination and application interface binding | `RegisterRoutes`, `AdministrativeApplication`, request decoders, administrative resources | Server route assembly; HTTP tests | Auth, HTTP, Jobs, pagination; coordinator | `api_test.go`, `openapi_contract_test.go`, lifecycle and HTTP integration | Core 01 OpenAPI owner input and generated clients | Reference Data HTTP adapter; private adapter; complete construction in assembly | High: preserve authorization, idempotency and exact errors. |
| Root: coordinator and admission | Immutable input preparation, frozen admission, execution, publication and invalidation | `Coordinator`, `NewCoordinator`, `ImportAdmission`, `PendingImport`, operation types | HTTP, Jobs, operator, integrity and portability paths | postgres.DB, owner storage, Jobs and shared finalizer | Admission, Jobs, concurrency, mixed refresh and terminal-effect tests | Configuration/Jobs v2; attempt/envelope schemas | Reference Data application; dedicated admission component and private persistence seams | High: no second executor or altered finality. |
| Root: verification and historical validation | Ordered verification, retained identity, restore-time integrity | Private verifier; required-state and restore entry points | Coordinator preparation; historical restore | Private format logic, storage, trust history | Verification engine/precedence, historical and recovery evidence | Check registry, limits, trust and diagnostics | Reference Data; owner-local verifier surface is private | High: preserve registry order and cleanup ownership. |
| Root: consumers, provenance, assignments and retention | Exact-set reads, canonical evaluation, pins and registry usage | Five-operation `Consumer`; `Retention`; `RegistryAssignments` | Indicators, observations, Network Flow, Reporting, revisions/import participation | Canonical repository, immutable content, pagination and transaction-bound guards | Consumer/profile/provenance/retention and downstream tests | Five request/result pairs; sets, provenance and portability | Reference Data facade; root-owned DTOs with explicit conversions | High: no table/file access leaks or changed identity. |
| Root: incident and portable operations | Destination verification, exact local reuse, export catalogs and parent publication | `IncidentReferences`, prepared imports, reference codecs | Incident Bundles and Reporting through assembly | Storage, trust, shared transactions, retained history | Portable admission/execution, incident export/retention | Incident-bundle v5; `reference_pack_refs.v1` | Reference Data participation; keep semantics | High: destination trust and activation preservation. |
| Root: Base, storage, persistence, collection and lifecycle | Release-bound registries, roots, revisions, audit, readiness and cleanup | Base reconciliation, storage ports, collector, preflight and recovery helpers | Server/operator startup, migration/recovery owners, Jobs | PostgreSQL, confined storage and owner finalizers | Base, canonical persistence, storage, collection and retired-table tests | Built-in release binding; authored SQL/migrations 46–63 | Reference Data state; keep root-confined adapter outside owner logic | High: preserve required readiness and reference retention. |
| Test scheduling | Instance-scoped test DB dependency | No production setter | Per-server test composition | Existing postgres.DB | Four race scenarios and two-instance isolation | Authored semantic row routing | Global hook removed in S06 | No cross-instance pause. |
| `internal/packformat/` | Pure archive, TUF, schema, profile, normalization, identity, index and diagnostic logic | Owner-private package; exported symbols are not cross-owner API | Reference Data and private tests | Canonical JSON, generated typed contracts, standard cryptography | Complete listed package tests and independent vectors | Authored Reference Pack schemas/catalogs/fixtures | Keep algorithms private; preserve the runtime/test projection separation completed in S13 | High for trust; no general profile rewrite. |
| `internal/packstate/` | Pure eligibility, cohort, transition, fallback and revision proposals | Private state functions | Application coordinator | Typed state inputs | `state_test.go` plus live integration | State contracts and transition acceptance | Keep private | High: proposals must not acquire persistence authority. |
| `recoverycontribution/` | Owner-built recovery catalog/inventory participation | Contribution and inventory provider | Recovery composition | Recovery contracts and Reference Data owner state | Recovery/service-backed integration | Versioned recovery contributions | Keep owner contribution | High: include pinned, failed and disabled retained content. |
| `testsupport/` | Reusable signed/reference fixtures and indicator helpers | Test support constructors | Other owner test packages | Reference Data contracts and fixture infrastructure | Caller suites | Synthetic fixture data | Keep test support; no runtime dependency | Medium: never distribute test trust as operator trust. |
| `internal/app/referenceassembly/` | Storage capabilities, bootstrap admission, telemetry, registry and operator adapters | Root/recovery storage, operator admission, observer and registry adapters | Server/operator facades | Rooted FS, Reference Data ports, participating owners and Extensions | Storage/telemetry tests and composed integration | Operational configuration and telemetry projections | Keep composition; own runtime wiring, not domain decisions | High: preserve borrowed-resource ownership and root confinement. |

### Exact target file register

#### `internal/app/referenceassembly/`

16 files; 4 test files.

```text
administration.go
administration_test.go
artifact_storage.go
object_collection.go
operator_admission.go
recovery_storage.go
registry_transactions.go
registry_usage.go
staging_bounds_test.go
storage.go
storage_test.go
telemetry.go
telemetry_test.go
trust_bootstrap.go
verification_storage.go
workspace_recovery.go
```

#### `internal/modules/reference_data/`

106 files; 39 test files.

```text
administration.go
administration_repository.go
administrative_actions.go
api_test.go
attestation.go
audit.go
base_reconciliation.go
base_reconciliation_integration_test.go
base_release.go
byte_retention.go
canonical_index.go
canonical_persistence_integration_test.go
canonical_publication.go
canonical_repository.go
concurrent_publication_integration_test.go
configuration.go
consumer.go
consumer_contract_test.go
consumer_integrity_integration_test.go
consumer_test.go
coordinator.go
coordinator_admission.go
coordinator_admission_integration_test.go
coordinator_artifacts.go
coordinator_execution.go
coordinator_guards.go
coordinator_identity.go
coordinator_integrity.go
coordinator_jobs_integration_test.go
coordinator_observation.go
coordinator_preparation.go
coordinator_publication.go
cutover.go
dependency_admission.go
historical_attempts.go
historical_identity.go
historical_validation.go
http_adapter.go
http_requests.go
idempotency_receipts.go
incident_export.go
incident_export_integration_test.go
incident_portability.go
incident_portability_test.go
incident_retention.go
incident_retention_integration_test.go
incident_retention_integrity.go
integrity.go
job_finalization.go
job_ports.go
job_reconciliation.go
job_terminal_effects.go
job_terminal_effects_integration_test.go
lifecycle_fixture_integration_test.go
limits.go
mixed_refresh_integration_test.go
object_collection.go
object_collection_integration_test.go
object_collection_storage_integration_test.go
object_collector.go
observation.go
openapi_contract_test.go
operation_types.go
operator_contract.go
operator_import_integration_test.go
portable_admission.go
portable_admission_integration_test.go
portable_admission_test.go
portable_execution.go
portable_execution_test.go
portable_history.go
portable_input.go
portable_input_test.go
portable_publication.go
portable_verification.go
prepared_test.go
profile_consumers_integration_test.go
provenance.go
provenance_test.go
public_types.go
public_types_test.go
readiness.go
reference_pack_integration_test.go
reference_pack_storage_reference_migration_test.go
reference_pack_test_helpers_test.go
registry_assignment.go
requests.go
retention.go
retired_format_fixture_test.go
retired_tables_integration_test.go
staging_rejection.go
staging_rejection_test.go
storage_port.go
storage_port_test.go
successful_envelope.go
successful_envelope_test.go
trust_repository.go
upload_limit_integration_test.go
validation_summaries.go
verification_diagnostics.go
verification_engine.go
verification_engine_test.go
verification_precedence_integration_test.go
verification_program.go
verification_repository.go
worker_isolation_integration_test.go
```

#### `internal/modules/reference_data/internal/packformat/`

68 files; 31 test files.

```text
admission_fixture_test.go
archive.go
archive_limit_test.go
archive_program.go
archive_ratio_test.go
archive_test.go
attempt_result.go
attempt_result_test.go
attestation.go
attestation_test.go
bootstrap.go
builtin_binding.go
builtin_binding_test.go
consumer_contract.go
content.go
content_admission_test.go
content_diagnostics.go
content_diagnostics_test.go
content_limit_test.go
content_test.go
dependencies.go
dependencies_test.go
dependency_fixture_test.go
diagnostic_collector.go
diagnostic_safety.go
diagnostics.go
diagnostics_test.go
duplicate_diagnostics.go
envelope.go
fixture_contract_test.go
fixture_projection_test.go
fixture_scratch_test.go
framework_lookup_test.go
identity.go
indicator.go
indicator_test.go
inventory_diagnostics.go
inventory_diagnostics_test.go
license.go
limit_diagnostics_test.go
manifest.go
manifest_diagnostics.go
manifest_limit_test.go
manifest_program.go
manifest_test.go
notice.go
portability.go
portability_retention.go
portability_retention_test.go
portable_content.go
portable_content_test.go
portable_context.go
portable_context_test.go
prepared.go
prepared_test.go
profile.go
provenance.go
semantic_findings.go
shape.go
shape_diagnostics.go
trust_fixture_test.go
tuf.go
tuf_diagnostics.go
tuf_limit_test.go
tuf_program.go
tuf_root_shape_test.go
tuf_test.go
validation_summary.go
```

#### `internal/modules/reference_data/internal/packstate/`

2 files; 1 test files.

```text
state.go
state_test.go
```

#### `internal/modules/reference_data/recoverycontribution/`

2 files; 0 test files.

```text
contribution.go
inventory.go
```

#### `internal/modules/reference_data/testsupport/`

3 files; 0 test files.

```text
indicators.go
owner_fixture.go
portable_fixture.go
```

### Adjacent inspected boundaries and explicit exclusions

The direct register above is retained in full. The following inspected boundaries extend this iteration; they do not add files to the 197-file Reference Data count or authorize unrelated refactors. Paths identify owner entry points and supporting callers/tests, not a new exhaustive repository-wide inventory.

| Boundary / inspected entry points | Current responsibility and callers | Dependencies / public contract / tests | Owner and planned disposition | Risk |
| --- | --- | --- | --- | --- |
| `internal/app/server/server.go`, `internal/app/migrate/migrate.go`, `internal/app/operator/operator.go`; `cmd/server`, `cmd/migrate`, `cmd/operator` | Compose application dependencies and perform startup/cutover admission | Configuration, PostgreSQL, object store, bootstrap and domain owners; application/process tests and package smokes | Exact application facades enforce S17 policy admission before service acquisition/mutation; `cmd/*` stays composition-only. Preserve borrowed-resource ownership and reverse-order idempotent cleanup. | High: an invalid binary can otherwise begin mutating state. |
| `internal/platform/authn/authn.go`, `internal/platform/bootstrap/bootstrap.go`, `internal/modules/auth/` | Password encoding/checking, TOTP, secret sealing, bootstrap account creation and authentication lifecycle | Existing Argon2id, SHA-1 TOTP and caller-nonce GCM; Auth/bootstrap unit, integration and browser routes | Platform owns primitive mechanics; Auth owns credential/session/MFA semantics. S18–S20 remove duplicate encoders and incompatible primitive use. | High: lockout, resource exhaustion or divergent bootstrap credentials. |
| `internal/platform/pagination/pagination.go`, `internal/modules/revisions/conflicts/conflict_token.go`, `internal/modules/networkflow/security.go` | Cursor and conflict-token security for owning routes | Existing AES-GCM formats including `cft3`/`nfc2`, entropy options, claims/AAD and expiry; owning token and HTTP tests | S20 adopts the narrow S18 seal/open primitive; payload, key lifecycle, authorization, replay binding and error mapping remain local. | High: shared machinery must not weaken owner-specific binding. |
| `internal/modules/evidence/upload_token.go`, `internal/platform/secretpurpose/registry.go` | Upload-token authentication and purpose/reuse checks | HMAC-SHA-256 and SHA-256; callers of Auth key derivation | Preserve supported algorithms and owner semantics; review S18 derivation changes and S23 service mapping. No blanket hash replacement. | Medium: an indirect key-derivation caller can be missed. |
| `internal/modules/recovery/encryption.go`, `streaming_encryption.go`, recovery assembly and operator flows | Journal sealing, small-artifact encryption and streaming backup/restore | Existing v1/v2 encryption representations, manually constructed stream nonces and key truncation; `contracts/recovery/`, export-root, due-verification and restore tests | Recovery owns one current bounded-memory envelope and admission in S21; platform supplies primitives. Preserve storage confinement and no-publication-before-validation. | High: unreadable backups or unsafe target admission. |
| `internal/modules/reference_data/internal/packformat/tuf.go`, `tuf_program.go`, `internal/app/referenceassembly/trust_bootstrap.go` | Actual Ed25519/TUF verification, bootstrap trust and lifecycle composition | Standard-library Ed25519; strict key/signature sizes; runtime contracts in `contracts/reference-packs/` and fixtures in `contracts/reference-pack-fixtures/` | S22 qualifies shipped verification paths without changing canonical formats, signer rules or immutable provenance. | High: a primitive test does not prove the real verifier. |
| `internal/platform/postgres/postgres.go`, `internal/platform/objectstore/objectstore.go` | Application-side database and S3 adapters | pgx SCRAM/TLS; MinIO static V4 signing, hashes/checksums and transport; platform owner tests and service-backed tests | S23 reviews reachable services/parameters and transport at adapters; retain `postgres.DB` and existing library protocols. Server products are external dependencies. | High: transitive security services and insecure transport configuration. |
| `internal/platform/enterpriseauth/enterpriseauth.go`, `configuration.go`, `manifest.go`; TLS, telemetry and WebSocket callers | OIDC/SAML verification, certificate configuration and network integration | Vendor verifier behavior, claims and key handling; enterprise auth tests; telemetry/TLS and collaboration routing | S23 retains library verification, constrains approved security use and records narrow protocol-only exceptions. | High: an enabled optional path can invalidate the overall claim. |
| `contracts/openapi-source/owners/module.auth/openapi.json`, `apps/web/src/app/authenticationModel.ts` and auth fixtures | Public MFA algorithm enum and browser enrollment/login flows | Current SHA-1 projection; generated clients, frontend type/unit tests and Auth browser rows | S19 changes the authored enum and normal projections; no workbook, saved-view, grid or visual redesign. | Medium: stale generated/client assumptions prevent enrollment. |
| `tools/toolchain_pins.json`, `Makefile`, `tools/harness/backend/build-go-artifact.sh`, `deploy/mvp/Containerfile` | Pinned toolchain/image and shared binary build/cache entry points | Go 1.27.1; existing cache identities do not establish the planned module policy; toolchain/build regressions | S17 makes the exact module and actual binary identity authoritative throughout supported build paths. S24 qualifies the package. | High: cache contamination, missing metadata or environment override. |
| `deploy/mvp/docker-compose*.yml`, package configuration/guide, `tools/release-evidence/` and existing smoke orchestration | Shipped package, operational recovery and cleanup | Current package permits HTTP and PostgreSQL `sslmode=disable`; three package smoke targets already participate in release | S23 defines secure production admission; S24 extends existing isolated qualification and canonical evidence. No second release system. | High: source success can conceal package or transport failure. |
| `tools/test_families/module.auth.json`, `module.recovery.json`, `module.reference_data.json`, `harness.release.json` and affected platform families | Semantic verification routing and release membership | `contracts/verification/`, owner catalog, authored topology and generated schedules | S17/S24 author diagnostic/module qualification routing; all slices update owning routes as needed. Never hand-edit generated roots. | Medium: unrouted checks are absent from release acceptance. |
| Core 01/Core 04, Reference Pack, Recovery/configuration and Testing Harness owner sections | Requirements and projection authority | `docs/spec/`, `docs/reference-pack-subsystem-nlspec.md`, `docs/testing-harness-nlspec.md`; typed contract inputs | S16 coordinates the candidate amendments. Ordinary human review establishes projection fidelity; no runtime or harness dependency on prose. | High: tests cannot resolve contradictory requirements or adopt a draft. |
| Applied SQL migrations, completed Reference Data boundaries, external servers/browsers/host certification | Retained history and separately owned deployment responsibilities | `db/migrations`, source-owner storage ports and deployment approvals | Preserve applied migrations; add a new migration only if the adopted cutover design needs it. No rewrite of existing schema history or expansion of the application claim into external products. | Intentional scope limit; environment evidence is still required. |

Dependency source inspection included pgx `v5.9.2` SCRAM, MinIO `v7.0.100` signing/checksums, SAML `v0.5.1` and coder/websocket `v1.8.14`. pgx uses the standard PBKDF2 implementation, and MinIO's V4 path uses standard SHA-256/HMAC; their mere presence is not a finding of noncompliance. S23 must map the reachable operation, parameters and purpose to supported module services. An import scan cannot decide that question.

## 3. Module boundary diagnosis

Reference Data remains a cohesive supporting subsystem. The completed private format/state packages, root-owned DTOs, operation-focused persistence and application assembly should remain. This iteration centralizes cryptographic mechanics and startup policy, while preserving subsystem ownership of formats, authorization and lifecycle.

| Responsibility | Current location / evidence | Correct owner | Disposition | Design constraint / long-term benefit |
| --- | --- | --- | --- | --- |
| Required cryptographic policy | Distributed owner clauses, build flags and runtime callers | Core security requirements plus a narrow platform policy boundary | Establish | One immutable production policy and explicit admission; no provider registry, alternative policy branch or configuration-driven downgrade. |
| Process startup admission | Server, migrate and operator facades | Each matching application facade using the platform boundary | Extend | Check actual module/build identity and enabled mode before acquiring services or mutating state; assembly remains testable through explicit dependencies. |
| KDF and GCM nonce mechanics | Authn, pagination, conflict tokens, Network Flow and Recovery | Small platform primitive boundary | Consolidate | Domain-separated derivation and module-generated nonces; keep payload schemas and key ownership out of the shared API. |
| Password and MFA semantics | Authn primitives, duplicated bootstrap encoding, Auth module | Auth owns account/session/MFA lifecycle; shared supported password primitive serves bootstrap and normal creation | Simplify | One encoder/parser policy; bounded work; deliberate version evolution without legacy fallback. |
| Token claims and replay protection | Existing pagination/Revisions/Network Flow/Auth owners | Existing owners | Keep | Shared sealing must not become a shared domain-token framework or replace owner-specific authorization/AAD. |
| Recovery artifact and journal formats | Parallel small/streaming encryption paths | Recovery | Unify | One current streaming codec for all sizes; bounded memory and authenticated finality; historical readers stay with historical releases. |
| Pack signatures and canonicalization | Reference Data private verifier | Reference Data | Keep and qualify | Execution qualification changes independently of stable signed bytes, trust rotation and method identity. |
| External protocol verification | PostgreSQL, object-store, enterprise-auth and transport adapters | Existing platform adapters and vetted libraries | Constrain and qualify | Enforce supported algorithms/parameters and transport at the adapter; do not implement replacement SCRAM, XML-signature, OIDC or S3 protocols. |
| Qualification and evidence | Existing harness, package smokes and release artifacts | Testing Harness/build/package owners | Extend | Reuse semantic routing, real binaries, canonical results and accountable cleanup; no new reporting dossiers. |
| HA, new pack algorithms/profiles and unrelated owner moves | Outside identified gaps | Future separately authorized work | Defer | Avoid speculative abstractions and additional compatibility surfaces. |

## 4. Public contract and behavior freeze map

This is a planned cutover, not a blanket behavior-preserving refactor. Changes marked below **require later implementation authorization**. Existing behavior is retained where it carries a durable domain or security guarantee, not merely because it exists.

| Contract | Owner / evidence | Retain or planned change | Existing / required validation | Risk |
| --- | --- | --- | --- | --- |
| Pack signed bytes, sets and provenance | Reference Pack owner, private verifier and canonical fixtures | Retain Ed25519, method identity, canonical digests, complete signer sets, immutable anchors and sixteen existing profiles | Independent vectors; actual bootstrap/TUF/rotation/replay/historical/destination paths under pinned module, S22 | High |
| Five consumer operations and source participation | Reference Data DTOs, retention and assignment ports | Retain operation set, exact-set selection, authorization, pins, transaction/lock order and source-owner revision effects | Consumer/provenance/retention tests and affected source-owner slices | High |
| Administrative HTTP/operator/Jobs behavior | Core 01/Core 04, adapters and assembly | Retain routes, actor attribution, replay, idempotency, bounded admission, finality, atomic publication and safe errors | Existing HTTP/Jobs/concurrency/operator tests plus S24 packaged parity and startup rejection | High |
| Password and MFA records | Core 04, Auth/bootstrap and OpenAPI owner | Change to bounded versioned PBKDF2-HMAC-SHA-256 records and SHA-256 TOTP; fresh credentials/enrollment required | S19 bootstrap/login/enrollment/replacement/reset/revocation tests, concurrent-load bound, enum compatibility and browser flows | High; breaking, requires later authorization |
| Secret, cursor and conflict-token representations | Auth, pagination, Revisions and Network Flow | Change sealed versions and derived keys; reject old formats; retain claim/AAD/expiry/error semantics | S18/S20 tamper, wrong-purpose/key, user/session/route replay, expiry and rotation | High; breaking, requires later authorization |
| Backup, journal and recovery admission | Recovery owner and `contracts/recovery/` | One current encryption codec and explicit profile binding; preserve bounded streaming, target safety and export-root binding | S21 real backup/due verification/restore, adversarial chunk/framing tests, interruption and no-publication tests | High; breaking, requires later authorization |
| Runtime/configuration admission and transport | Core 04, application facades and platform adapters | Ordinary/disabled/wrong-module binaries and insecure production transport reject; fresh-state cutover before mutation | S17/S23/S24 fail-closed tests and disposable package scenarios using actual binaries | High; breaking, requires later authorization |
| Authentication browser and generated contracts | Auth OpenAPI owner, web authentication model and fixtures | Update algorithm projection through normal compatibility process; preserve routes, session UX and safe error handling | Auth frontend type/unit/browser evidence; generated drift and OpenAPI compatibility disposition | Medium; public enum change, requires later authorization |
| WebSocket protocol and domain indicators | Transport owner and indicator owners | Preserve reviewed handshake hash and opaque indicator digest strings; no security use of an exception | S23 reachable-call review, positive/negative security evidence and targeted diagnostic findings | Medium |
| UI/workbook/view schemas/selectors | Existing frontend and source owners | No planned redesign, entity-row/view-schema changes or selector/golden refresh | Preserve existing suites; expand only if an actual auth/integration change reaches these surfaces | Low |
| Runtime/test asset split, harness and release | Contract-generation and Testing Harness owners | Preserve separated fixture family, semantic ownership and canonical artifacts; add module/cache/diagnostic/package acceptance inputs | Drift/boundary/harness checks, real-binary negatives and complete final release | High; release policy change requires later authorization |
| Deployment trust, clock and external data | Reference Pack and deployment owners | Preserve explicit trusted-clock assertion, deployment-supplied trust and licensing/source-identity constraints | Existing admission/rejection tests and ordinary deployment review | High; test trust is not deployment approval |

## 5. Coupling and boundary findings

RP-F01–F19 and their historical outcomes remain below. The former RP-F09/S07 deferral is now expanded, not erased. No completed structural or package repair is reopened without new evidence.

| Finding | Evidence | Risk if unresolved | Classification | Owner | Required action |
| --- | --- | --- | --- | --- | --- |
| RP-F20: current owner exclusions do not support the chosen regulated profile | RP-REQ-261 requires a new method for FIPS; RP-REQ-264 calls it unsupported; Core security projections still describe current primitives | Implementation can contradict owners or make an unsupported claim | must_fix | Core security / Reference Pack / Recovery | S16 separates execution qualification from unchanged signature semantics and coordinates all security-service requirements. |
| RP-F21: build settings are not authoritative runtime admission | Go toolchain/image pins exist, but shared build/cache and all three facades lack the proposed exact module policy gate | Ordinary, stale or overridden binaries execute under a regulated label | must_fix | Build / platform policy / application facades | S17 pins identity, includes it in caches and artifacts, and rejects invalid startup before acquisition/mutation. |
| RP-F22: duplicated nonce/KDF mechanics and permissive key handling | Authn hash-based derivation and fallback secrets; Recovery truncation; GCM nonce construction across multiple owners | Approved algorithm names conceal unsupported use; changes drift across callers | must_fix | Platform primitive boundary and calling owners | S18 explicit key contracts, domain-separated HKDF and module-generated-nonce sealing; S20/S21 migrate callers. |
| RP-F23: password/MFA primitives and bootstrap encoding conflict with the selected profile | Argon2id in Authn/bootstrap; SHA-1 TOTP and public enum/client assumptions | Base authentication prevents qualification; encoders or clients diverge | must_fix | Auth / bootstrap / Core 04 / OpenAPI | S19 one password owner path and bounded PBKDF2; SHA-256 enrollment and complete lifecycle/browser validation. |
| RP-F24: token migration can erase owner-specific binding | Pagination, `cft3`, `nfc2` and Auth secrets have separate claim/AAD/expiry rules | New shared machinery permits replay, wrong-purpose acceptance or unsafe error disclosure | must_fix | Existing token owners | S20 adopts only primitive mechanics and retains owner-local semantics with adversarial tests. |
| RP-F25: recovery has parallel encryption formats and caller nonces | `encryption.go` and `streaming_encryption.go`; journal and backup envelope schemas | Policy-incompatible backup creation or unusable/unsafe restores | must_fix | Recovery | S21 one current streaming format for all sizes, authenticated chunk identity/finality and profile-bound admission. |
| RP-F26: module support alone does not qualify the actual pack verifier | Real bootstrap/TUF/historical/portable code paths differ from an isolated Ed25519 test | A different runtime path or malformed input behavior escapes evidence | must_fix | Reference Data / assembly | S22 execute independent vectors and complete lifecycle paths under the pinned module. |
| RP-F27: dependency and transport behavior is not established by direct imports | pgx SCRAM, MinIO signing/checksums, OIDC/SAML, TLS/telemetry and WebSocket; insecure current package transport options | An optional integration or unprotected transport invalidates the application claim | must_fix | Platform adapter / configuration owners | S23 service/purpose/parameter mapping, certificate/transport enforcement and narrow reviewed non-security exceptions. |
| RP-F28: existing package acceptance has no regulated binary/environment contract | Existing three smokes and passing release establish the prior package only | Source tests pass while cache, image, startup inputs or environment undermine the package | must_fix | Package / build / Testing Harness | S24 real-binary identity and negative admission, fresh-state flows, reference-environment record and canonical release routing. |
| RP-F29: readiness states can be conflated | Previous release passed; Reference Pack candidate and deployment acceptance remain separately governed | Historical results or draft changes are represented as current compliance/rollout approval | must_fix | Specification / release / deployment owners | S25 complete one final-candidate release and separately record implementation, qualification, adoption and deployment decisions. |
| RP-F30: stable formats and existing owner boundaries remain useful | Completed DTO/persistence/fixture separation; canonical pack identity; protocol-only hash/data uses | Unnecessary redesign introduces migration and coupling without a security benefit | intentional/no_action | Existing subsystem owners | Preserve these boundaries; no algorithm negotiation, blanket hash bans or unrelated module moves. |

## 6. Refactor workstreams and phase sequencing

Each implementation slice is its own workstream. The authorized plan is intentionally sequential: **S16 → S17 → S18 → S19 → S20 → S21 → S22 → S23 → S24 → S25**. Later discoveries become explicit prerequisite slices; they do not silently expand an active slice. S25 must complete last.

| Phase | Workstream / class / owner | Required previous | Required next | Goal and affected boundary | Validation / handoff checkpoint |
| --- | --- | --- | --- | --- | --- |
| 1: Specification and execution boundary | S16 / root / Core security and affected specification owners | Later authorized implementation task | S17 | Consistent profile/cutover requirements and projection design | Owner review and complete service/exception routing; no unresolved owner contradiction. |
| 1: Specification and execution boundary | S17 / chain / build, platform policy and application facades | S16 | S18 | Exact build/module identity and startup admission | Actual binaries accept the candidate and reject wrong/missing/disabled/stale identities before side effects. |
| 2: Application cryptography | S18 / chain / platform primitives | S17 | S19 | Explicit keys, HKDF and random-nonce sealing | Known-answer, domain-separation, misuse and key-use-bound tests. |
| 2: Application cryptography | S19 / chain / Auth and bootstrap | S18 | S20 | One password path and SHA-256 TOTP | Full credential/MFA/session/browser lifecycle and resource bounds. |
| 2: Application cryptography | S20 / chain / Auth, pagination, Revisions and Network Flow | S19 | S21 | Owner-specific tokens on shared primitives | Binding, replay, tamper, expiry and rotation evidence for every caller. |
| 2: Application cryptography | S21 / chain / Recovery | S20 | S22 | One current backup/journal encryption policy | Real restore/due verification and adversarial stream/admission checks. |
| 3: Verification and integrations | S22 / chain / Reference Data and reference assembly | S21 | S23 | Qualify the actual pack verifier | Shipped paths, lifecycle and unchanged canonical identities. |
| 3: Verification and integrations | S23 / chain / platform adapter and configuration owners | S22 | S24 | Qualify enabled transitive services and transport | Supported service mapping plus positive/negative runtime evidence; no unresolved security use. |
| 4: Package and handoff | S24 / chain / package, build and Testing Harness | S23 | S25 | Regulated package acceptance and environment evidence | Disposable scenarios, canonical routing/results and accountable cleanup. |
| 4: Package and handoff | S25 / chain / release and handoff owners | S24 and every discovered prerequisite | None | One complete final-candidate validation and maintainable handoff | All implementation slices DONE, final release/cleanup pass, distinct approval dispositions recorded. Completes last. |

| Phase | Principal risks / sequencing reason | Binary phase exit |
| --- | --- | --- |
| 1 | Mistaking build selection for operational qualification; contradicting adopted owners; admission after resource mutation | The profile is a consistent candidate and the policy is executable through typed/build/runtime inputs with no Markdown dependency. Any actual owner contradiction is BLOCKED and resolved before dependent implementation. |
| 2 | Authentication lockout, unbounded password work, token-binding regressions, nonce misuse and unusable backups | Every directly owned security primitive follows the selected policy; incompatible state rejects before mutation and affected owner lifecycle/negative tests pass. |
| 3 | Hidden transitive behavior, unsupported parameters/purposes or an overly broad exception | Every enabled application security service has an approved-use mapping and execution evidence; unresolved use blocks qualification. Pack format identity remains stable. |
| 4 | Incorrect host/image assumptions, incomplete cleanup, stale evidence or premature deployment claims | Package scenarios and one complete final release pass with cleanup; implementation/qualification/adoption/deployment statuses are explicit, with promotion blocked where required acceptance is missing. |

**Execution rule:** after later authorization, mark the next workstream IN_PROGRESS before its first implementation change. At its exit, record the changed owners/areas, compatibility consequence, commands and canonical run identities, failures, residual risks and next dependency; mark DONE only when the binary exit is satisfied. **Update this tracker after each completed workstream and before beginning the next one.** Failed prerequisites remain BLOCKED, not DONE. Allowed statuses are TODO, IN_PROGRESS, BLOCKED, DONE, DEFERRED and DROPPED. Compact canonical evidence replaces duplicate narrative dossiers.

## 7. Proposed refactor slices and exits

**Every slice below is TODO and requires later implementation authorization.** No new validation result is claimed. Evidence cited here diagnoses the baseline; planned tests and commands are acceptance requirements, not completed checks. §8 gives the discovered Make routes. Each slice must retain its owner-local regression coverage and report any additional command/semantic-row discovery in its checkpoint.

### S16 — Define the regulated profile and amend owners

**Owner:** Core security, with Reference Pack, Recovery, configuration, OpenAPI and Testing Harness owners. **Areas:** specifications, documentation and typed contract design. **Gaps:** RP-F20/F29. **Depends on:** later authorized implementation task. **Next:** S17.

- **Remediation:** define one application cryptographic profile covering passwords, TOTP, key derivation, encryption, security signatures, enabled integrations and production admission. Amend Core 04 and affected Core 01/Recovery/configuration clauses together. Amend RP-REQ-261 and RP-REQ-264's FIPS exclusion without changing Ed25519 or `tuf_1_0_35_offline_bundle_v1`. Identify the corresponding authored Recovery schemas, configuration contracts, Auth OpenAPI enum, build inputs and harness projections before coding.
- **Boundary:** distinguish domain semantics, module-service usage and deployment qualification. Enumerate each security service and owner, and each permitted non-security protocol use with its reason and validation route. Keep the review in normal owner changes and existing machine projections; do not create an approval registry or replacement evidence dossier.
- **Rationale / long-term benefit:** a future module update can change qualified build/environment identity without changing pack signatures, trust history or unrelated domain contracts. Requirements become coherent enough for implementation and human projection review.
- **Compatibility / migration:** explicitly declare the fresh-state cutover, breaking credential/MFA/sealed-state formats, startup rejection and changed transport admission. Define how incompatible state is identified before mutation. Applied migrations remain immutable; historical deployments/backups use matching releases.
- **Unresolved risk / risk if left open:** the present Reference Pack exclusion is incompatible with the selected direction until amended. Tests cannot waive it. Missing or contradictory service requirements can produce implementation that overclaims qualification.
- **Validation / evidence:** review Core security, Reference Pack, Recovery/configuration, OpenAPI and harness owners as one candidate; use `make lint-markdown` and `git diff --check`. No current owner adoption or new crypto execution evidence is asserted.
- **Binary exit:** every affected service, protocol exception, cutover rule and projection has one consistent owner and validation route; no unresolved owner contradiction. An actual contradiction is recorded as **BLOCKED: owner contradiction** and blocks dependent implementation. Formal adoption remains a separately recorded status.
- **Rollback:** withdraw conflicting candidate amendments together and keep production promotion blocked; do not implement against contradictory requirements.

### S17 — Make build identity and startup admission authoritative

**Owner:** build/toolchain and narrow platform cryptographic-policy boundary; all three application facades own invocation. **Areas:** implementation, build configuration, tests and documentation. **Gap:** RP-F21. **Depends on:** S16. **Next:** S18.

- **Remediation:** pin `v1.0.0-c2097c7c` and its verified archive digest in authored build inputs, alongside the pinned toolchain. Include module selection, digest, compiler/build settings and relevant mode inputs in shared build cache identities, harness fingerprints and release artifacts. Cover supported server/migrate/operator builds and packaged binaries; do not rely on a label or caller-supplied version string.
- **Admission:** each exact application facade checks actual module version, available build information and enabled FIPS mode before acquiring application services or mutating state. Runtime mode/environment overrides cannot downgrade the policy. Reject wrong module, absent/contradictory metadata and disabled mode with bounded safe diagnostics. Keep feature logic out of `cmd/*`.
- **Rationale / long-term benefit:** one authoritative policy closes gaps between source, cached binary, image and process. The platform boundary stays small and upgrades remain reviewable without a provider framework.
- **Compatibility / migration:** ordinary builds and `fips140=off` cannot run production application commands. Development/test composition supplies explicit supported inputs and owned dependencies; it does not add a production bypass. Preserve borrowed PostgreSQL/object-store ownership and idempotent cleanup.
- **Unresolved risk / risk if left open:** stale cache entries, local/global toolchain mismatch, spoofed metadata or production environment overrides can admit an unqualified binary. Module metadata must be verified using the pinned toolchain's supported interfaces.
- **Validation / evidence:** build/admission tests run actual binaries for correct/wrong/missing/disabled cases and assert no service acquisition or state mutation on rejection; exercise cross-policy cache reuse attempts. Run toolchain, boundary, harness and affected facade checks through discovered Make routes. Author the targeted strict-mode assessment route; do not invent or invoke an unregistered target.
- **Binary exit:** all three binaries prove the required actual identity and enabled mode; all negative cases and cache-contamination tests reject before side effects; canonical build/release artifacts carry matching identities.
- **Rollback:** revert an incomplete build/admission change as a unit and leave regulated production unavailable. Never restore an ordinary-build fallback.

### S18 — Establish approved key derivation and sealing primitives

**Owner:** narrow platform primitive boundary, with existing key/purpose owners as callers. **Areas:** implementation, contracts and tests. **Gap:** RP-F22. **Depends on:** S17. **Next:** S19.

- **Remediation:** replace custom hash-based derivation with standard-library HKDF-SHA-256 and explicit, unambiguous domain separation. Require explicit master keys of the profile's exact supported size; remove production fallback secrets, permissive excess lengths and silent truncation. Keep key identifiers, rotation, authorization and secret-purpose reuse checks with their owners.
- **Primitive boundary:** provide a small seal/open API using the module's AES-GCM random-nonce facility. Production callers never provide GCM nonces or test entropy. Define versioned framing and associated-data inputs without owning domain payload schemas. Independently derive per-object/message keys where necessary so safety does not depend on a deployment-global nonce counter. Specify and enforce per-key use limits, including the module random-nonce limit, over the actual key lifetime.
- **Rationale / long-term benefit:** consolidate difficult derivation, framing and nonce mechanics once; keep future token/artifact growth local to each domain owner. Avoid a generic crypto service, provider API or shared domain-token framework.
- **Compatibility / migration:** derived keys and sealed representations change. Old state is rejected by explicit version/profile admission, not probed with a legacy decoder. Existing HMAC/SHA-256 callers such as upload tokens retain their semantics while adopting correctly separated keys.
- **Unresolved risk / risk if left open:** primitive names alone do not establish approved usage; ambiguous context encoding, reused keys, malformed framing or lifetime-limit assumptions can undermine isolation.
- **Validation / evidence:** known-answer HKDF vectors; distinct owner/purpose/object contexts; malformed/oversized inputs, tampering, wrong keys/AAD and key-use boundaries. Verify production code cannot inject nonce/entropy. Run affected platform/owner slices and boundary checks; migrate callers only in their owning subsequent slices.
- **Binary exit:** the primitive API and exact key contract pass all positive/negative vectors and usage-bound checks; no production caller of the new primitive supplies nonces or test entropy; caller migration obligations are explicitly assigned.
- **Rollback:** keep the new primitives unexposed until their contract is complete; do not introduce automatic fallback in consumers.

### S19 — Replace password and MFA cryptography

**Owner:** Auth lifecycle and shared authentication primitives; bootstrap, Core 04 and OpenAPI owners participate. **Areas:** specifications, implementation, OpenAPI/contracts, tests and documentation. **Gap:** RP-F23. **Depends on:** S18. **Next:** S20.

- **Remediation:** use standard-library PBKDF2-HMAC-SHA-256 with an initial fixed cost of **600,000 iterations**, **16-byte random salt** and **32-byte result**. Use a versioned, bounded record format and one creation/verification policy for bootstrap and ordinary credentials. Remove duplicate Argon2 encoders and legacy verification paths.
- **MFA:** use SHA-256 TOTP with **32-byte secrets**, retaining six digits, 30-second periods and the existing bounded acceptance window. Update enrollment, verification and the authored OpenAPI algorithm enum through the normal compatibility process, then regenerate clients/fixtures. Keep authentication, session revocation, replacement/reset authorization and browser behavior owned by Auth.
- **Rationale / long-term benefit:** one credential policy removes divergent encoders and unsupported authentication primitives. Explicit versions and bounded parameters permit deliberate future cost strengthening without arbitrary work factors. The selected initial password cost follows the linked OWASP FIPS guidance.
- **Compatibility / migration:** existing password records and TOTP enrollment are incompatible; no conversion or automatic reset. Fresh deployment provisions credentials and enrollment under the new profile. Document authenticator compatibility and the intentional public enum break; do not add a SHA-1 fallback.
- **Unresolved risk / risk if left open:** authentication alone can prevent an application-wide claim. A permissive parser or unmeasured cost can expose CPU exhaustion; unsupported clients can lock users out.
- **Validation / evidence:** bootstrap, ordinary password creation, successful/failed login, MFA enrollment, replacement/reset and session revocation; independent PBKDF2/TOTP vectors and skew/replay/invalid-code cases. Reject malformed versions, lengths and work factors before expensive derivation. Measure representative concurrent authentication and record an explicit bounded resource/latency acceptance decision. Run Auth/bootstrap owner slices, relevant service-backed/browser rows, OpenAPI compatibility and generated/client checks.
- **Binary exit:** both credential creation paths use one policy; lifecycle/browser cases pass; malformed records reject cheaply; representative concurrency stays within the agreed bound; old formats fail explicitly and the enum/client projections agree.
- **Rollback:** fail fresh deployment qualification and correct the candidate; do not mix old authentication code with newly created regulated state.

### S20 — Move secrets and short-lived tokens onto approved sealing

**Owner:** Auth secrets, platform pagination, Revisions conflict tokens and Network Flow cursors. **Areas:** implementation, contracts and tests. **Gap:** RP-F24. **Depends on:** S19. **Next:** S21.

- **Remediation:** migrate all identified authentication-secret, pagination, conflict-token and Network Flow sealing callers to S18. Version changed sealed/token formats and explicitly reject older formats. Keep claims, associated-data encoding, owner/purpose labels, user/session/route binding, expiry, rotation and public error mapping inside the existing owners.
- **Boundary:** share cryptographic mechanics only. Retain authorization before/after decoding where required, bounded parsing and existing replay protections; do not let one owner's token become consumable by another. Review indirect derivation consumers such as Evidence upload-token signing without gratuitously replacing supported HMAC semantics.
- **Rationale / long-term benefit:** nonce and key-derivation fixes are maintained once while future token fields remain cohesive with the domain that interprets them.
- **Compatibility / migration:** old cursors/conflict tokens expire through explicit incompatibility rejection; clients follow existing restart/reload/error behavior. Persisted authentication secrets require fresh state. No dual reader, dual writer or compatibility alias.
- **Unresolved risk / risk if left open:** a less-visible token path can retain incompatible crypto; a shared implementation can accidentally erase AAD, actor/session binding, expiry or safe failure behavior.
- **Validation / evidence:** each owner tests positive round trips plus cross-user, cross-session and cross-route replay, wrong-purpose keys, tamper, expiry boundaries, rotation, malformed/oversized records and old-version rejection. Exercise public error mappings and owner lifecycle flows, not only primitive helpers. Use the affected owner/platform slices and service-backed rows selected through `make task-guide`.
- **Binary exit:** every identified caller uses S18 and every owner-specific binding/rejection scenario passes; no legacy sealing caller or unassigned key-derivation consumer remains.
- **Rollback:** keep the candidate unqualified until all affected callers agree; never deploy mixed format writers or restore legacy decoding as a tactical fix.

### S21 — Replace recovery encryption with one current format

**Owner:** Recovery, with operator/application assembly and Core recovery requirements. **Areas:** specifications, implementation, contracts, tests and operator documentation. **Gap:** RP-F25. **Depends on:** S20. **Next:** S22.

- **Remediation:** introduce one current streaming backup envelope using independently derived per-artifact HKDF keys and module-generated nonces per chunk. Authenticate artifact identity, chunk position, length and finality; bound chunk sizes/counts and per-key use. Route small artifacts through the same codec and remove parallel legacy encryption paths from the new release.
- **Admission and journals:** update journal sealing through S18 and bind backup/restore/verification admission to the cryptographic profile. Preserve confined storage, export-root binding, freshness/overlap checks, bounded-memory streaming, interruption safety and validation before publication/target admission. Reject incompatible profile/envelope state before any mutation.
- **Rationale / long-term benefit:** one current codec prevents divergent integrity rules as artifact types and sizes grow, while actual recovery remains bounded and independently verifiable.
- **Compatibility / migration:** new encrypted-envelope versions are intentional breaks. Historical backup readers remain with historical releases; document matching release/backup retention and fresh-deployment boundaries. Keep applied SQL migrations immutable; coordinate any necessary new authored migration and typed schema changes with their owners.
- **Unresolved risk / risk if left open:** a backup can be created successfully yet violate module usage limits, lose finality, leak partial output or fail when needed. Old recovery proofs must not authorize a new profile.
- **Validation / evidence:** actual backup, due verification and restore through operator/package paths; small/large/empty boundary artifacts; truncation, reordering, duplication, substitution, length/finality corruption, wrong keys, interrupted operations and incompatible versions. Assert no publication, target admission or retained-state mutation on failure. Run Recovery owner/service-backed slices, shape/drift checks and operational recovery smoke.
- **Binary exit:** one current codec and journal policy serve all supported artifacts; actual restore and due verification pass; all adversarial/incompatibility cases reject before publication/admission, with bounded memory and accounted cleanup.
- **Rollback:** preserve historical deployments/backups with their matching releases. Do not add a legacy reader or run a converter to rescue an incomplete new format.

### S22 — Qualify the real Reference Pack verifier

**Owner:** Reference Data and reference assembly. **Areas:** implementation verification, tests and contract review. **Gap:** RP-F26. **Depends on:** S21. **Next:** S23.

- **Remediation:** exercise bootstrap admission and the complete TUF verifier under the pinned module, including root/key rotation, signature thresholds and malformed inputs, replay, historical validation and destination portability. Use shipped execution paths and independently authored vectors, not merely a standalone Ed25519 call.
- **Preserved contract:** Ed25519, canonical bytes/digests, method identity, immutable provenance, trust history and destination trust remain unchanged. The owner amendment changes execution qualification, not signed-pack interpretation. Runtime projections remain separated from fixtures.
- **Rationale / long-term benefit:** the actual verifier receives credible qualification while valid packs and historical identities avoid unnecessary migration. Future module upgrades reuse these stable lifecycle and vector obligations.
- **Compatibility / migration:** valid canonical pack inputs retain identity; malformed keys/signatures and unsupported algorithms still reject. This does not authorize converting incompatible application databases or importing trust from a historical deployment automatically.
- **Unresolved risk / risk if left open:** passing primitive tests can miss another runtime verification path, signature-validation edge cases, or altered canonicalization after a build change.
- **Validation / evidence:** independent positive/negative vectors plus bootstrap/full TUF/rotation/replay/historical/destination scenarios through Reference Data owner/service-backed rows and packaged Reference Pack smoke. Compare valid canonical digests to preserved vectors; run targeted strict assessment through authored routing and explain any protocol-only findings.
- **Binary exit:** all actual verifier paths and lifecycle scenarios pass with the pinned actual module identity; valid canonical digests are unchanged, malformed inputs reject and no algorithm fallback exists.
- **Rollback:** block qualification and fix the verifier/module disposition; do not substitute algorithms or reissue pack identities to conceal an execution failure.

### S23 — Close transitive cryptographic and transport gaps

**Owner:** existing platform PostgreSQL, object-store, enterprise-auth, TLS/telemetry, WebSocket and configuration adapters; security owner reviews service usage. **Areas:** platform adapters, configuration, specifications, tests and documentation. **Gap:** RP-F27. **Depends on:** S22. **Next:** S24.

- **Remediation:** map reachable PostgreSQL/SCRAM, S3 signing/checksum, OIDC/SAML, TLS, telemetry and WebSocket operations to the module services actually used, their parameters and permitted purpose. Review enabled optional features as part of the application boundary. Enforce approved algorithms/parameters at existing adapters while retaining vetted library protocol/signature verification.
- **Transport:** validate certificate chains, names and trust handling; reject insecure production transport settings. Specify application-side behavior when TLS termination is external, without claiming qualification of the terminator or external servers. Current HTTP and PostgreSQL `sslmode=disable` package settings need explicit disposition under the adopted policy.
- **Exceptions:** retain only explicitly reviewed non-security uses such as the WebSocket handshake hash. Treat indicator MD5/SHA-1 strings as data. Do not blanket-ban imports, wrap security calls to evade enforcement, or silently exempt a dependency. PBKDF2's password-storage service restrictions require purpose review for SCRAM; using a standard-library implementation is insufficient by itself.
- **Rationale / long-term benefit:** qualification follows actual behavior instead of library names or import counts. Adapter-owned policy keeps integration constraints cohesive and avoids hand-written replacements for complex protocols.
- **Compatibility / migration:** incompatible provider algorithms, credentials, certificates and transport settings fail explicitly. Operators configure supported providers/transports for fresh deployments; no legacy negotiation fallback is introduced.
- **Unresolved risk / risk if left open:** third-party defaults, optional checksums or provider negotiation can execute unqualified services. An unresolved supported-purpose mapping blocks qualification; do not assume the selected candidate can satisfy every dependency before this review.
- **Validation / evidence:** positive/negative integration execution for every enabled security path, including unsupported algorithms/parameters, malformed cryptographic inputs and rejected certificate/transport cases. Review dependency source with runtime evidence; targeted strict-mode assessment supplements both. Use platform/affected owner and service-backed slices, configuration checks and package scenarios.
- **Binary exit:** every enabled security service has a supported module-service/purpose mapping and passing positive/negative execution evidence; every protocol exception has a bounded non-security rationale; no unresolved usage or insecure production setting remains.
- **Rollback:** stop qualification for an unresolved required integration and resolve its supported design with the owner. Do not silently narrow the claimed boundary, disable a required feature or add a second crypto policy.

### S24 — Integrate regulated package acceptance

**Owner:** deployment package, build/toolchain and Testing Harness/release owners. **Areas:** deployment package, harness, tests, release artifacts and documentation. **Gap:** RP-F28. **Depends on:** S23. **Next:** S25.

- **Remediation:** extend existing package qualifications to verify actual identities of server, migrate and operator; startup rejection; fresh-state admission; authentication/MFA; token flows; Reference Pack lifecycle; and actual recovery. Include wrong-module/disabled-mode/missing-metadata and incompatible retained-state/backup scenarios in disposable environments with rejection before mutation.
- **Environment and artifacts:** record host OS/kernel, runtime/version, CPU/features, application image/index and resolved platform image, binary/module/archive/build identities and their relationship to permitted operating environments. Recheck certificate/security disposition before qualification. Use existing canonical package/release artifacts and semantic routing, not a new report or approval dossier.
- **Harness:** author task, owner and topology inputs through existing mechanisms, regenerate outputs through Make, and require the regulated package results in release acceptance. Retain all owned-resource/workspace cleanup attempts and fail on incomplete cleanup. Keep strict-mode diagnostics distinct from production mode and security-service qualification.
- **Rationale / long-term benefit:** qualify the exact package operators receive and reuse the established release system as features grow. Source-only tests cannot detect packaging, cache, configuration or runtime overrides.
- **Compatibility / migration:** operators receive explicit fresh-deployment guidance and preflight rejection for incompatible installations. No automatic rollout, reset, conversion or retained deployment operation.
- **Unresolved risk / risk if left open:** the reference environment may not yet be available or accepted; a distro-derived image alone does not establish the operating environment. Stale artifacts and incomplete cleanup can make a passing package result misleading.
- **Validation / evidence:** `make standup-package-smoke`, `make standup-reference-pack-smoke` and `make standup-operational-recovery-smoke` with the new semantic cases; identity/cache negatives, harness/shape/drift/toolchain checks and canonical artifact review. Establish the target environment relationship explicitly; if unresolved, record the blocker rather than qualifying a substitute host.
- **Binary exit:** disposable package scenarios and all cleanup pass on the recorded reference target with verified identities; release acceptance requires those canonical results; the module/environment relationship has an explicit supported disposition.
- **Rollback:** keep release acceptance/promotion blocked until the package and routing agree. Never mark helper or partial results as completed package qualification.

### S25 — Final validation and handoff completion

**Owner:** release/validation and handoff owners, with affected specification and operational owners. **Areas:** validation, documentation and this tracker. **Gap:** RP-F29. **Depends on:** S24 and all implementation/prerequisite slices. **Next:** none; **must finish last**.

- **Remediation:** finish focused owner and changed-input checks, run `make agent-finalize` before broader final verification, then `make test-fast` and one complete `make release-check` on the final candidate. Inspect canonical results, readiness projections, actual binary/module/environment identities and cleanup. A changed candidate or failing release needs its own corrected final validation; never combine partial runs into a pass.
- **Handoff:** review fresh-deployment and incompatible-state handling, matching-release backup retention, operational failure diagnostics, key provisioning/rotation, integration limits and recovery procedures. Document the module-update process: choose an exact candidate, recheck certificate/security status and environment support, amend pins/cache identities, rerun service/verifier/package qualification and record ordinary review. Future updates are deliberate, never floating-selector refreshes.
- **Rationale / long-term benefit:** a reproducible release boundary and actionable maintenance instructions survive beyond the current authors without multiplying signoff artifacts.
- **Compatibility / migration:** no customer rollout or retained-state mutation is automatic. Clearly distinguish implementation completion, package qualification, specification adoption and customer deployment approval. Missing adoption or deployment/environment acceptance continues to block production promotion even when engineering checks pass.
- **Unresolved risk / risk if left open:** partial, historical or stale evidence can be mistaken for readiness; unresolved security disposition or environment assumptions can be lost during handoff.
- **Validation / evidence:** focused owner/drift/OpenAPI/boundary/toolchain/harness checks as needed; finalization, fast checks and one complete final release. Use `make explain-run RESULTS_DIR=<run-root>` for canonical inspection. Record exact source/graph/run identities, failure history, cleanup, skipped checks and their reasons. If retained-run maintenance is intended, use the permitted successful full warm-check `RESULTS_DIR`; otherwise report it skipped because unset.
- **Binary exit:** S16–S24 and every discovered implementation prerequisite are DONE; one complete final-candidate release and cleanup pass; canonical evidence, compatibility guidance, remaining limitations and all four approval/qualification statuses are recorded. Then mark S25 DONE last and close the S07 umbrella. No unresolved implementation or package-qualification failure is hidden as a deployment-only issue.
- **Rollback:** leave S25 incomplete or BLOCKED with the exact failed boundary and retained artifacts. Do not relabel the prior release, erase failures or declare promotion from an incomplete candidate.

## 8. Validation plan

Use public Make targets from the repository root and choose the narrowest owner rows first. `make task-guide ROLE=module-author OWNER=<owner-id>`, `make help` and `make help-all` remain the live navigation surface; this table records the relevant discovered routes, not a second command registry. Test routing does not define requirements or establish specification completeness.

| Layer / workstreams | Discovered route | Required timing and evidence |
| --- | --- | --- |
| Immediate document update | `make lint-markdown`; `git diff --check` | Required now. Also compare the exact 197-file register, preserved historical text/run identities, new TODO statuses and pre-existing working-tree file hashes. No product validation. |
| Owner unit/behavior checks, S18–S23 | `make task-guide ROLE=module-author OWNER=module.auth`, `OWNER=module.recovery`, `OWNER=module.reference_data`; then `make test-slice OWNER=<owner-id> [ROWS=<discovered-row-ids>]` | Guides for these three owners were inspected successfully during planning. Discover affected Revisions, Network Flow, bootstrap, configuration and other platform row selection when their slices begin; do not invent semantic row IDs. |
| Integration/service-backed checks, S17–S24 | `make service-backed-test-slice OWNER=<owner-id> [ROWS=<discovered-row-ids>]` after the owning task guide | Exercise real persistence, provider, auth/session, verification and recovery boundaries where reached. Keep isolated services and accountable cleanup. |
| Auth browser/client projection, S19/S20 | Auth task guide and its routed browser rows; `make frontend-typecheck` and `make frontend-unit` when their inputs change | Real enrollment/login/reset/session flows with the new enum and formats. No unsolicited UI redesign or golden refresh. Discover exact browser row selection from authored routing at execution time. |
| Authored contracts and generated outputs | `make generate`, `make generate-drift`, `make json-shape-check`, `make generated-artifact-policy-check` | Run only when relevant owner inputs change during implementation. Preserve generated-root ownership and runtime/test separation; never use Markdown as a generator/test input. |
| Public API compatibility, S19 | `make openapi-compatibility-check` | Resolve the intentional algorithm-enum change through the normal compatibility process; a known break is not silently ignored or hidden behind legacy support. |
| Owner/module boundaries and pins | `make backend-module-boundary-check`; `make toolchain-drift` | Required when the platform primitive/policy boundary, imports or build pins change. Use normal frontend boundary routing if clients change. |
| Harness and release routing, S17/S24 | `make harness-contract`; `make explain-target TARGET=release-check DETAIL=summary` | Verify authored routes, cache/fingerprint inputs, release membership and generated topology. Explain-target is discovery, not execution evidence. |
| Targeted strict-mode assessment, S17/S22/S23/S24 | **TODO: author/discover the exact public Make/harness route in S17, then require its results in S24.** | No such new route is claimed today. Pin the module, select relevant security paths and retain actionable diagnostic findings. `fips140=only` is diagnostic, not deployment mode or sole compliance criterion; expected protocol-only findings are reviewed, not suppressed through bypasses. |
| Actual package acceptance, S21/S22/S24 | `make standup-package-smoke`; `make standup-reference-pack-smoke`; `make standup-operational-recovery-smoke` | Extend the existing routed smokes. Verify all three real binaries, fresh-state/negative admission, auth/tokens, pack lifecycle and actual recovery on the recorded reference target, including cleanup. |
| Final candidate, S25 last | Focused checks, then `make agent-finalize`, `make test-fast` and one complete `make release-check` | Finalization precedes broader end-of-run verification. Record source/module/image/environment identities and all failures. No merging partial runs or substituting historical release evidence. |
| Canonical handoff review | `make explain-run RESULTS_DIR=<run-root>`; inspect canonical summary/manifest, package artifacts and cleanup receipts | Require complete passing results and correct identities. Distinguish retained-run maintenance from read-only result inspection. `RESULTS_DIR` for finalization must meet the repository's retained-run requirements or maintenance is explicitly skipped. |

Current `lint-markdown` configuration selects authored documentation globs but does not directly include this handoff path. Run the required public target and report its configured scope honestly; supplement it with read-only tracker structure, inventory, history and diff checks. Do not edit lint configuration or bypass the Make surface solely to broaden this document-only task.

## 9. Top-level work tracker

Historical DONE rows below retain their original outcomes; their validation does not qualify S16–S25. All new implementation rows start TODO.

| ID | Work item / workstream | Status | Depends on | Evidence / disposition | Binary exit |
| --- | --- | --- | --- | --- | --- |
| RP-P02 | Prior package-plan document update | DONE | Prior user-selected plan | Markdown `20261005T042618Z-p49665` and original session retained below | Prior document-only step completed. |
| RP-P03 | Regulated-readiness document update | DONE | Current document-only request | Markdown `20261005T122414Z-p70033` passed for its configured scope; diff, tracker structure, 197-file inventory, history and 6,684-path scope checks passed | Only this tracker changed; S16–S25 remain TODO and require later implementation authorization. |
| RP-S01 | Reconcile owner/tracker authority | DONE | Prior iteration | Original execution retained below | Prior slice exit passed. |
| RP-S02 | Remove dead interfaces/misleading fixtures | DONE | S01 | Original execution retained below | Prior slice exit passed. |
| RP-S03 | Stable owner DTOs/private verification | DONE | S02 | Original execution retained below | Prior slice exit passed. |
| RP-S04 | Complete application construction | DONE | S03 | Original execution retained below | Prior slice exit passed. |
| RP-S05 | Separate transport and semantics | DONE | S04 | Original execution retained below | Prior slice exit passed. |
| RP-S06 | Persistence seams and instance isolation | DONE | S05 | Original execution retained below | Prior slice exit passed. |
| RP-S07 | Validated cryptography umbrella, expanded into S16–S25 | TODO | S16–S25 | Former DEFERRED disposition retained in history; no separate implementation slice | S25 finishes last and the expanded capability's implementation/qualification exits pass. |
| RP-S08 | Candidate owner/projection review | DONE | S06 | Historical review, not formal adoption or recurring reporting | Prior slice exit passed. |
| RP-S09 | Disposable operational qualification | DONE | S14, S09a | Prior packaged lifecycle/recovery/cleanup results retained | Prior package exit passed. |
| RP-S09a | Exported-bundle recovery storage repair | DONE | Discovered during S09 | Confined export-root capture/restore and admission binding retained | Prior repair exit passed. |
| RP-S10 | Prior structural validation/handoff | DONE | S08 | Release `20261005T030730Z-p45700`, 1261/1261 on its original dirty candidate | Completed last in that iteration. |
| RP-S11 | Simplify specification/reporting obligations | DONE | Prior implementation request | Historical plan and implementation checkpoint retained | Owner cleanup retained substantive controls. |
| RP-S12 | Remove accounting-only machine contracts | DONE | S11 | Historical plan and checkpoint retained | Behavioral fixture completeness retained. |
| RP-S13 | Separate runtime projections/test assets | DONE | S12 | Runtime/fixture families and boundary regression retained | Runtime excludes fixture assets. |
| RP-S14 | Complete packaged administration support | DONE | S13 | Optional configuration and pinned image retained | Prior package configuration exit passed. |
| RP-S15a | Deterministic presence capture framing | DONE | Discovered during S15 | Unchanged-golden repair and failed-release history retained | Prior visual prerequisite passed. |
| RP-S15 | Integrated package release/final handoff | DONE | S09, S15a | `20261005T070744Z-p19802`, 1264/1264; 17 readiness projections and cleanup pass | Completed last in the prior package iteration. |
| RP-S16 | Define regulated profile/amend owners | TODO | Later authorized implementation task | RP-F20/F29; §7; no new execution evidence | Consistent candidate, service/exception owners and routes; no owner contradiction. |
| RP-S17 | Build identity/startup admission | TODO | S16 | RP-F21; §7; no new execution evidence | Three actual binaries enforce identity/mode before side effects; cache negatives pass. |
| RP-S18 | Approved derivation/sealing primitives | TODO | S17 | RP-F22; §7; no new execution evidence | Exact keys, separated derivation and random-nonce primitive pass misuse/use-bound tests. |
| RP-S19 | Password/MFA cryptography | TODO | S18 | RP-F23; §7; no new execution evidence | One credential policy, SHA-256 TOTP, lifecycle/client and resource-bound checks pass. |
| RP-S20 | Secrets and short-lived tokens | TODO | S19 | RP-F24; §7; no new execution evidence | All caller migrations and owner-specific binding/rejection tests pass. |
| RP-S21 | One current recovery encryption format | TODO | S20 | RP-F25; §7; no new execution evidence | Real restore/due verification and negative framing/admission tests pass. |
| RP-S22 | Actual Reference Pack verifier qualification | TODO | S21 | RP-F26; §7; no new execution evidence | Shipped verifier/lifecycle pass with unchanged canonical identities and no fallback. |
| RP-S23 | Transitive crypto/transport qualification | TODO | S22 | RP-F27; §7; no new execution evidence | Every enabled security path has a supported service mapping and positive/negative evidence. |
| RP-S24 | Regulated package acceptance | TODO | S23 | RP-F28; §7; no new execution evidence | Real binaries, reference environment, package scenarios, routing and cleanup pass. |
| RP-S25 | Final validation/handoff completion | TODO | S24 and all prerequisites | RP-F29; §7; no new execution evidence | All slices DONE, complete final release/cleanup, explicit approval states; completes last. |

## 10. Session handoff log

This is the appended regulated-readiness planning/document-update session. Prior planning sessions, implementation checkpoints and failures remain in the historical sections without reassigned run identities.

### Scope and authority

| Date / session | State / files inspected or touched | Commands / decision | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 / regulated planning | Repository AGENTS, refactor-tracker skill/format, local planning framework, domain/design navigation, owner context and controlling tracker | User selected the full regulated plan; current action is explicitly document-only | Preserve the dirty `e7a1497` working tree and all completed history. Only this tracker is editable now; later implementation starts S16. |

### Backend boundary

| Date / session | State / files | Commands / evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 / regulated planning | 197-file Reference Data/assembly register; app facades; Authn/bootstrap; cursor/conflict/Network Flow; Recovery; actual pack verifier; platform adapters | Targeted `rg`/file reads and dependency source inspection; working-tree content snapshot | Keep completed Reference Data ownership. Add only the narrow policy/primitive boundaries and owner-specific migrations described in S17–S23. |
| 2026-10-05 / document update | Existing tracker and repository scope | Exact register/file comparison and before/after content snapshots across 6,684 tracked/nonignored paths passed | All 197 direct files remain accounted for. Only this tracker changed; pre-existing source, contract, test, configuration and other edits were preserved. |

### Frontend boundary

| Date / session | State / files | Evidence / decision | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 / regulated planning | `apps/web/src/app/authenticationModel.ts` and Auth/OpenAPI fixture assumptions | Current SHA-1 TOTP enum reaches the browser | S19 changes the authored contract and projections with browser lifecycle checks. No rendered UI audit, selector redesign or visual-golden update is part of this document task. |

### Contracts and code generation

| Date / session | State / files | Commands / evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 / regulated planning | Core security, RP-REQ-261/264, Recovery envelope/journal contracts, Auth OpenAPI owner, build pins and authored harness inputs | Targeted owner/projection reads | S16 coordinates requirements before implementation; S19/S21 version intentional incompatible formats. No contracts or generated artifacts changed now; no generator reads this tracker. |

### Tests and harness

| Date / session | State / files | Commands / evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 / planning discovery | Auth, Recovery and Reference Data verification routes | `make task-guide ROLE=module-author OWNER=module.auth` / `module.recovery` / `module.reference_data` succeeded; `make explain-target TARGET=release-check DETAIL=summary` succeeded | Read-only command discovery, not product execution or regulated qualification. Discover additional exact owner/browser/strict-mode rows during their slices. |
| 2026-10-05 / historical evidence review | Original final release and earlier failures | Retained summaries/manifests and tracker checkpoints inspected | Prior `20261005T070744Z-p19802` remains historical passing evidence with original source/graph identity; failed `20261005T061847Z-p6495` remains recorded. |
| 2026-10-05 / document completion | Required current checks | `make lint-markdown` PASS at `20261005T122241Z-p67499` and after closeout at `20261005T122414Z-p70033`; summaries: `.cartulary/test-results/<run-id>/adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` PASS | Lint passed its configured scope. Supplemental read-only checks passed: twelve active sections, all per-slice fields, S16–S25 TODO, table/heading/fence structure, unchanged 197-file register, preserved historical plan/checkpoints/tail and only-tracker content change. RP-P03 DONE. |

Read-only discovery encountered guessed paths that did not exist, including `internal/platform/postgres/settings.go` and a `docs/core-04*` glob; file enumeration resolved the actual owners/entry points. These lookup failures changed nothing and are not product-test failures. The current Markdown target does not directly select this handoff path; its configured-scope result and supplemental tracker checks are recorded separately.

Product tests, generation/drift, package/recovery runs, release verification and `make agent-finalize` are deliberately skipped for this document-only step. Retained-run maintenance is skipped because `RESULTS_DIR` is unset. No prior successful run is refreshed or relabeled.

### Security and authorization

| Date / session | State / files | Evidence / decision | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 / regulated planning | Go 1.27.1 frozen module candidate, certificate/security policy, local library paths, existing image pin and package transport | One policy; all application binaries/integrations; fresh state; preserve Ed25519; reviewed protocol-only exceptions | Candidate pin/digest and Linux/amd64 Debian 12 reference target recorded. Actual module usage, certificate disposition and environment suitability must be qualified later; no adoption, customer rollout or external-server certification is asserted. |

### Open risks and next session

| Date / session | Current state | Remaining risk | Next action |
| --- | --- | --- | --- |
| 2026-10-05 / document handoff | S16–S25 all TODO; previous package iteration complete | Current code still uses the old cryptography; owner exclusion, transitive service mapping and actual reference-environment suitability remain unresolved implementation prerequisites | After later authorization, start S16, then follow the dependency chain and update this tracker at every workstream exit. S25 finishes last. Keep normal review/canonical evidence; do not recreate retired dossiers. |

## 11. Open prerequisites and production boundaries

There is no unresolved user preference preventing this plan. Planning choices do not establish adoption, compliance or authority to mutate a retained deployment. TODO here means required future evidence/work; BLOCKED is reserved for an actual contradiction or failed prerequisite encountered during execution.

| ID | Prerequisite / question | Why it matters | Required owner / evidence | Status |
| --- | --- | --- | --- | --- |
| RP-PRE-01 | Prior coordinated reporting/accounting amendments | Completed S11 requirement cleanup remains valid history | Original owner changes and S11 checkpoint | DONE |
| RP-PRE-02 | Prior package release routing and cleanup | Three smokes already participate in the historical package release | Original S09/S15 results, including `20261005T070744Z-p19802` | DONE |
| RP-PRE-03 | Formal owner adoption and actual deployment suitability | Engineering success does not adopt a draft or approve a customer environment | Existing owner-status process and applicable deployment/security/licensing acceptance; ordinary records, no new dossier | TODO |
| RP-B02 | Regulated security policy and Reference Pack owner amendment | Existing exclusion cannot be ignored; signature identity should remain stable | S16 coordinated owner candidate and service/exception definitions | TODO |
| RP-B03 | Exact module/service/build/runtime/environment disposition | Certificate association or a Debian-derived image alone is insufficient | S17/S22/S23/S24 actual identities, supported service/purpose mappings, current security disposition and permitted environment relationship | TODO |
| RP-B04 | Pinned-module and diagnostic Make/harness routing | Floating selection, stale caches and unrouted assessment undermine qualification | S17 authored build/cache/routing policy, S24 canonical release acceptance | TODO |
| RP-PRE-04 | Fresh-state admission and incompatible-format detection | Rejection must precede mutation for all three binaries and recovery | S16 contract; S17/S19–S21 implementation; S24 real-binary negative evidence | TODO |
| RP-PRE-05 | PBKDF2 cost and authenticator compatibility | Fixed cost must remain operationally bounded; SHA-256 TOTP clients must work | S19 representative concurrency evidence, explicit acceptance bound and supported-client guidance | TODO |
| RP-PRE-06 | Enabled integration service usage and production transport | SCRAM, S3, OIDC/SAML, TLS/telemetry and optional paths may have unsupported uses/defaults | S23 adapter-owned positive/negative evidence and supported module-service/purpose mapping; unresolved use blocks qualification | TODO |
| RP-PRE-07 | Reference environment availability and acceptance | Host/runtime/CPU/image relationship must be established before package qualification | S24 actual Linux/amd64 Debian 12 reference-target evidence; separate acceptance for a customer target | TODO |

RP-B02–RP-B04 are expanded from historical DEFERRED constraints into active TODO prerequisites; their historical entries are not rewritten. Prior RP-B01 owner contradiction and RP-B05 pre-production release blocker remain resolved in their original iteration. Do not reopen them merely because another capability is planned.

| Readiness dimension | Current status | Completion authority / evidence |
| --- | --- | --- |
| Previous structural/package implementation | DONE | Historical S01–S15/prerequisite exits and original releases; no regulated claim. |
| New regulated implementation | TODO | S16–S25 completion and exact final-candidate evidence. |
| Regulated package qualification | TODO | S22–S24 actual-service/verifier/package evidence and S25 complete release on the established reference target. |
| Specification adoption for production | TODO; not claimed | Coordinated owner-status process; the Reference Pack 0.2.0 candidate remains draft at this planning point. Tracker/test results cannot adopt it. |
| Customer deployment approval | TODO; outside this implementation authorization | Acceptance of the actual target, external dependencies, trust/clock inputs and applicable security/licensing requirements by their owners. No automatic rollout. |

Production promotion remains blocked wherever required owner adoption or environment/deployment acceptance is absent, even if implementation and reference-package qualification later pass. A failed implementation or unsupported application security service cannot be reclassified as an external deployment responsibility.

## 12. Binary completion criteria

**The current document update is complete only when:**

- Only this controlling tracker changed during this step; all pre-existing working-tree changes remain intact and the baseline is correctly described as dirty `e7a1497` with completed remediation.
- The exact 197-file Reference Data inventory is preserved and verified; the inspected adjacent application/authentication/encryption/integration/build boundaries are recorded with owners, callers, contracts and test posture.
- Active sections contain S16–S25 as separate TODO workstreams, the complete dependency chain, phase risks/exits, and per-slice remediation, areas, rationale/benefit, compatibility, unresolved risks, validation and binary exit.
- S07 is explicitly expanded; prior completed workstreams, failures and original validation identities remain historical without relabeling.
- All behavior changes are marked as requiring later implementation authorization; this task changes no product behavior, generated input/output, conformance or release evidence.
- The appended session records `make lint-markdown`, `git diff --check`, supplemental tracker checks, actual results and skipped checks. RP-P03 is marked DONE only after they pass.

**The later regulated implementation is complete only when:**

- S16–S24 and any discovered prerequisites are DONE and S25 completes last, with this tracker updated after every completed workstream before the next begins.
- The coherent owner candidate, typed projections, build/runtime policy and approved primitive/service usage agree; no owner contradiction or unsupported enabled security service remains hidden.
- Fresh-state admission rejects incompatible state and backups before mutation; no legacy crypto reader/writer, password converter, automatic reset or fallback remains in the new production policy.
- Pack signatures, canonical identities, trust rotation, immutable provenance and other retained domain/security guarantees pass the actual verifier and lifecycle tests.
- Actual package identities and reference environment are established, disposable scenarios and cleanup pass, and one complete final-candidate release supplies current canonical evidence. Historical/partial results are not substituted.
- Operational failure handling, key management, fresh-deployment/recovery guidance and the exact-module update policy are actionable; implementation completion, package qualification, specification adoption and customer deployment approval each have an explicit disposition.
- Missing formal adoption or deployment/environment acceptance still blocks promotion. Completing the engineering iteration never authorizes retained-state mutation or customer rollout.

## Historical completed package planning — S11–S15

The following is the preserved previous package-iteration plan and session record. Its authorizations, clean-baseline assumption, current/next labels, S07 deferral and S15 final-slice instructions belong to that historical iteration only. The active §1–§12 above control S16–S25 and correctly identify the current dirty working tree. The exact file register remains in active §2 instead of being duplicated here. Original completed outcomes, failures and validation identities are not reassigned to the new plan. The same historical-context rule applies to all older records below.

### Archived package plan — 1. Scope and source posture

The user authorized implementation of the proposed plan on 2026-10-05. Execute the active workstreams in order and update this tracker after each exit before starting the next workstream. The new plan supersedes the previous iteration's current-state sequencing and production deferrals only where explicitly stated below; completed execution records and their original evidence identities remain historical.

| Scope fact | Decision |
| --- | --- |
| Current baseline | Clean `e7a1497ff9e8219f7ccdad7a09af9e356cb146ec`, inspected 2026-10-05; prior structural remediation is committed. |
| Targets | `internal/modules/reference_data/`, `internal/app/referenceassembly/`, their machine contracts, and the package/harness boundaries explicitly listed in §2. |
| Current authorized change | Implement S11 → S12 → S13 → S14 → S09 → S15, including owner amendments, contracts, runtime/test boundaries, package qualification and final handoff. Preserve the pre-existing tracker edits. |
| Current completion | COMPLETE: S11–S14, S09, discovered prerequisites S09a/S15a, and S15 are DONE. S15 completed last with one complete passing final-candidate release and verified cleanup. |
| Previous completion | S01–S06, S08 and S10 remain DONE; S10 completed the prior pre-production iteration. Its results do not qualify future changes. |
| Reporting decision | Remove compulsory human evidence dossiers, companion-document hashes, repeated signoff records and accounting-only requirement/criterion catalogs. Retain normal repository review, owner authority and the existing coordinated adoption process. |
| Production boundary | Qualify the current non-FIPS Ed25519 package through repeatable disposable-environment tests. Actual customer deployment, infrastructure certification and destructive operations against retained environments are excluded. |
| S09 scope change | The previously deferred on-prem rehearsal is DONE as automated disposable package qualification after S14. Retained deployment operation remains outside this iteration. |
| S07 disposition | DEFERRED: validated cryptography remains a separate capability requiring a target posture, owner amendment and actual verifier/build qualification. No algorithm substitution or negotiation. |
| Compatibility and data | Preserve public routes, authorization, actor attribution, replay, signed bytes, canonical identities, immutable provenance, atomic publication and retention. Preserve applied migrations and fail-without-mutation preflight; add no aliases, converter, dual writer or historical trust reconstruction. |
| Evidence retained | Behavioral fixtures, independent vectors, semantic test routing, canonical run artifacts, security checks and useful machine-produced SBOM/license outputs. Test totals and mappings do not prove behavioral completeness. |
| Operational constraints | Trust roots, explicit clock assertion, redistribution restrictions and suitability for the actual deployment remain substantive decisions. Test trust is not deployment trust. |
| Authority | Adopted subsystem owners govern their scopes, then normative Core owners; typed inputs project those owners. Core 05 applies only to its separately claimed publication boundary. Domain owns vocabulary; code and historical plans supply evidence. The Reference Pack 0.2.0 candidate remains draft today. |

The existing four behavioral corrections remain regression obligations: strict composition identities, framework lookup-key deduplication, typed early upload rejection and pre-parse URL admission. Do not redo the completed structural work or the browser acquisition repair merely because older historical entries describe their earlier failures.

The accessible prior release at `.cartulary/test-results/20261005T030730Z-p45700/run-summary.json` passed 1261/1261 units with zero failures, skips or cancellations. Its manifest records the original dirty candidate based on `c9f5b366fe0d295057a30aaf6e488af9c587fdcb`, source digest `sha256:b5d0f1c25905104352a03faae5c571510902d31ac89b4b16d1632dabe671d6ae`; do not relabel it as a new run on the current commit. Exact prior finalization, test-fast and cleanup evidence is preserved below.

#### Adjacent boundaries and explicit exclusions

| Boundary | Iteration treatment | Reason / owner |
| --- | --- | --- |
| `internal/app/server/`, `internal/app/operator/`, `internal/testutil/appsupport/` | Preserve assembly; repair only reproduced package defects and add isolated test composition | Application facades compose dependencies; `cmd/*` remains composition-only. |
| `internal/platform/postgres/`, Jobs, rooted FS and Extensions finalization | Reuse existing ports, storage confinement and protocols | No generic repository, transaction lifecycle, executor or infrastructure rewrite. |
| `contracts/reference-packs/` | Remove accounting-only catalog; retain runtime schemas, algorithms, limits and original Base assets; move test-only inputs in S13 | Sixteen profiles and nine algorithm families remain supported; no optional external corpus is newly distributed. |
| Planned `contracts/reference-pack-fixtures/` and its generated test artifact package | Separate test contract family through existing `contracts/index.json` generation | Fixture growth must not expand runtime artifacts or diagnostic vocabulary; no production import of test assets. |
| Reference Pack, Core, Extensions, Reporting/Composition, recovery and Testing Harness owners | Coordinated reporting/adoption-process amendments, preserving substantive behavior | Normal review establishes projection fidelity; this tracker cannot waive an adopted requirement. |
| Four auxiliary Reference Pack evidence handoffs | Preserve as historical records during S11; remove perpetual refresh requirements | No new owner/obligation/limit/security dossier is required per slice. Current operational instructions belong in maintained guides. |
| Indicators, observations, Network Flow, Reporting, Incident Bundles and recovery | Preserve owner interfaces and run affected regression slices | No replacement canonicalizers or unrelated feature work. |
| `apps/web` and shared protocol/UI packages | Preserve administration behavior, selectors and generated client contracts | No UI redesign, workbook surface, saved-view change or unsolicited golden refresh. |
| `db/migrations`, `db/queries` | Preserve applied history and existing storage semantics | No schema migration planned; a reproduced defect needs its own owning invariant and regression. |
| Generated roots, contract registry, boundary policy and harness routing | Amend authored inputs, then regenerate through Make | `tools/generated_artifact_policy.json` owns generated membership; never hand-edit outputs or lockfiles. |
| `deploy/mvp`, toolchain/build ownership and `tools/release-evidence` | Optional administration wiring, immutable base-image pin, isolated qualification and release inclusion | Extend existing package/recovery infrastructure; no parallel evidence system or broad deployment framework. |
| Authentication redesign, unrelated subsystems, external corpora, actual deployment targets and validated cryptography | Excluded or S07 DEFERRED | No new compliance claim, licensing approval or deployment authorization is inferred. |

### Archived package plan — 3. Module boundary diagnosis

Reference Data remains one cohesive supporting subsystem. Private algorithms remain in `internal/packformat` and `internal/packstate`; the root owns supported DTOs, semantic operations and private persistence adapters. No directory rename or generic repository/transaction framework was introduced.

| Responsibility | Current boundary | Rationale and evidence |
| --- | --- | --- |
| HTTP and safe public errors | Private `httpAdapter` in `http_adapter.go` and decoding in `http_requests.go` | Owner-local same-package adapter avoids exporting conversion scaffolding. Semantic operations no longer import HTTP/auth types. |
| Admission / verification / integrity | Separate complete private components; only `NewCoordinator` constructs administration | Operator admission and independent integrity do not inherit irrelevant Jobs dependencies. Missing dependencies fail construction. |
| Worker registration | `internal/app/referenceassembly/administration.go` before route binding | One lifecycle worker, usable without HTTP; duplicate registration rejects. |
| Receipts | Private `idempotency_receipts.go` bridge | Preserves Core storage/status representation within the publication transaction and exact commit-proof checks. |
| Persistence | Existing `postgres.DB`, private administrative and frozen-verification repositories | Explicit operation-focused SQL ownership; no platform transaction lifecycle abstraction. |
| Consumer/provenance/portable data | Root-owned `public_types.go` and explicit private conversions | Cross-owner consumers no longer depend on private format representations. Canonical identity/null/order tests pass. |
| Storage and external-owner dependencies | Application assembly builds confined capabilities and injects owner ports | Borrowed resources stay borrowed; runtime cleanup semantics unchanged. |
| Test scheduling | Instance-scoped DB barrier in `internal/testutil/appsupport` | Four former global-hook uses and explicit two-server isolation retain deterministic race evidence. |
| Negative fixtures and cutover guards | Retired fixture isolated by name; live canonical fixtures always sign | Preserves retired-format rejection without implying a second supported import path. |

The completed module boundaries remain appropriate. This iteration changes the following supporting boundaries:

| Responsibility | Current location | Correct owner | Disposition | Evidence and design constraint |
| --- | --- | --- | --- | --- |
| Human reports and document revision binding | Reference Pack Table 4-C and auxiliary handoffs | Ordinary repository review and owner documents | Simplify | Repeated document hashes and dossiers add no executable guarantee; do not replace them with another report registry. |
| Requirement-count accounting | Traceability JSON, fixture family and private format test | No runtime owner; behavioral verification stays with its semantic owner | Remove | The validator checks 264/54 identifiers and mappings, not behavior. S11 must amend the owning clauses before S12 removes it. |
| Runtime and fixture artifact aggregation | One `contractreferencepacks.Artifacts` collection | Reference Data runtime projections versus test support | Split | Production `shape.go`, Base loading and diagnostic vocabulary scan the same collection that carries fixtures. No fixture-backed production fallback. |
| Package administration and recovery choreography | MVP Compose package and existing smoke scripts | Deployment composition and Testing Harness mechanics | Extend | Existing server/operator/recovery semantics remain with their owners; introduce neither another executor nor a second evidence format. |
| Production compliance beyond current Ed25519 profile | Target-specific authority and build posture | Security/build owners | Defer | S07 is not a prerequisite for qualifying the explicitly non-FIPS package. |

### Archived package plan — 4. Public contract and behavior freeze map

| Contract | Current owner | Evidence | Existing tests | Retained evidence / new qualification | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| HTTP routes, envelopes and authorization | Core 01/Core 04; Reference Data adapter | OpenAPI owner input, `http_requests.go`, `http_adapter.go` | API/OpenAPI/lifecycle/operator tests | Retain exact public mapping; add packaged HTTP/operator parity and dependency rejection | High | Keep `/api/v1`, authorization order, idempotency, removal and action errors. |
| Five consumer operations | Reference Pack candidate and machine projections | `ResolveCurrentPackSet`, `GetPackEntry`, `LookupPackEntries`, `EvaluateIndicatorValue`, `GetPackProvenance` | Consumer contract/profile/provenance tests | DTO round trips, exhaustive error unions, cursor omission/inheritance and exact expiry | High | No new operation or algorithm negotiation. |
| Pins, usage and source mutation | Reference Data plus participating source owners | `Retention`, `RegistryAssignments`, transaction-bound consumers | Retention, concurrent publication and downstream owner suites | Two-instance isolation and unchanged shared lock order | High | Preserve exact tokens and source-owner revision effects. |
| Jobs, admission, deadline and finality | Core Jobs/Extensions; Reference Data semantic effects | Frozen inputs, shared finalizers, terminal effects | Jobs, mixed refresh, commit-fault and restart tests | Route-free worker construction, lost acknowledgement, stale state and canceled admission | High | Configuration/Jobs v2; one execution budget, successful envelopes required. |
| Formats, sets and historical provenance | Reference Pack candidate | Private format/state and authored canonical fixtures | Engine, TUF, immutable persistence and historical validation | Fixed vectors unchanged after runtime/test asset separation | High | Retain canonical `.v1`, profile major 2, complete signer sets and immutable anchors. |
| Reporting, portability and recovery | Respective owners with Reference Data ports | Exact-set bindings and incident-bundle v5 | Export/import, snapshot, retention and restore suites | Changed activation between admission/render; exact reuse; historical restore and required loss | High | No current-set substitution, reactivation on import or historical trust reconstruction. |
| UI and shared projections | Core/admin controller and generated clients | Existing administration model and OpenAPI | Frontend unit/type/import-boundary and browser owner rows | Existing disabled, mixed refresh, removal and uncertain-outcome behavior | Medium | No new WebSocket, workbook surface, saved-view or grid-selector behavior. |
| Configuration and readiness | Core 04/Extensions | Existing configuration/limits/Base readiness | Base startup and configuration tests | Claimed/unclaimed startup | High | Clock false; timeout 1,800 in 60–86,400; lookup 50 in 1–200; cursor 900 seconds; null rules unchanged. |
| Verification accounting | Testing Harness owner | `tools/test_families/module.reference_data.json` and owner catalog | Current owner routing; historical counts are not acceptance evidence | Retain semantic row ownership; remove accounting-only catalogs and test totals in S12; integrate package qualification in S15 | Medium | Test counts and mappings do not establish adoption or completeness. |

### Archived package plan — 5. Coupling and boundary findings

The original RP-F01–F11 evidence and dispositions remain in the historical iteration record below. F01–F08 and F10 require no repeated remediation; F09 remains deferred with S07, and F11's fail-closed cutover guards remain intentional. The prior tracker’s incorrectly named deleted files are corrected in §2.

| Finding | Evidence | Risk if unresolved | Classification | Owner | Planned disposition |
| --- | --- | --- | --- | --- | --- |
| RP-F12: compulsory human reporting and duplicate promotion bookkeeping | RP-REQ-020, Table 4-C, Tables 1-A/28-A, §32 and the four auxiliary handoffs | Records drift; editorial changes trigger repeated reporting; paperwork is mistaken for acceptance | must_fix | Reference Pack and companion specification owners | S11 removes hashes/dossiers and duplicate signoff records, retaining ordinary review, coordinated adoption and substantive constraints. |
| RP-F13: accounting-only catalogs and hard-coded fixture totals | `traceability.v1.json`, `rpfx_traceability`, `spec_traceability`, RP-REQ-259/AC-049 and `fixture_contract_test.go` | Future profile/test growth requires unrelated numbering maintenance; zero-unmapped can obscure missing behavior | must_fix | Reference Pack / Testing Harness | S11 amends owners; S12 removes accounting and replaces 163-total assertion with executable fixture completeness. |
| RP-F14: test assets share the production projection collection | `contracts/index.json`, contract generator, `packformat/shape.go`, `diagnostic_safety.go`, `base_release.go` | Fixture growth expands production artifacts and can alter diagnostic vocabulary | must_fix | Reference Data / contract generation | S13 separates test assets using the existing family mechanism and enforces import boundaries. |
| RP-F15: packaged administration lacks an explicit supported configuration path | MVP configuration has storage roots but no administration example; operator requires server worker, claim, roots and clock inputs | Correct source-level behavior remains difficult to configure and qualify in the shipped package | must_fix | Deployment/application assembly | S14 adds optional package wiring with operator-supplied trust; S09 exercises it. |
| RP-F16: package/recovery smokes are outside the release gate | `make explain-target` reports both targets as `helper_only`; current recovery smoke exercises backup/restore but not Reference Pack administration | A release can pass without packaged Reference Pack lifecycle qualification | must_fix | Testing Harness / deployment | S09 adds the proposed Reference Pack smoke; S15 integrates all relevant package checks into release. |
| RP-F17: package smoke cleanup suppresses failures | Package/recovery script cleanup uses `compose down ... || true` and image-removal suppression | Leaked owned resources can accompany a successful qualification result | must_fix | Package test composition / Testing Harness | S09 records cleanup results, attempts all owned-resource cleanup and fails qualification on remaining resources. |
| RP-F18: mutable application base-image reference | `deploy/mvp/Containerfile` uses `gcr.io/distroless/base-debian12:nonroot` without a digest | The same source can build with different base bytes | must_fix | Build/toolchain owner | S14 pins an immutable digest through existing pin ownership and drift checks. |
| RP-F19: meaningful controls could be mistaken for reporting | Clock/trust admission, licensing rules, immutable attestations, canonical run evidence and current unsupported-FIPS policy | Removing controls would weaken authorization, reproducibility or truthful deployment claims | intentional/no_action | Reference Data, security, licensing and harness owners | Retain operational controls and machine evidence; remove only report/accounting obligations. |

### Archived package plan — 6. Refactor workstreams and phase sequencing

| Phase | Workstream / class | Depends on | Next | Goal and likely boundaries | Principal risk | Validation / phase exit |
| --- | --- | --- | --- | --- | --- | --- |
| 1: Simplify obligations | S11 / root | User implementation request | S12 | Owner specification and documentation cleanup | Removing an actual control or implying adoption | Owner text agrees on simplified process; Markdown checks pass. |
| 2: Simplify executable inputs | S12 / chain | S11 | S13 | Accounting removal in contracts, fixtures and routing | Deleting real behavioral evidence with accounting | Fixtures remain closed, covered and routed; focused checks pass. |
| 2: Simplify executable inputs | S13 / chain | S12 | S14 | Runtime/test asset separation, generated families and boundaries | Missing runtime schema or changed canonical/diagnostic behavior | Runtime dependency exclusion, vectors, Base and privacy checks pass. |
| 3: Qualify the supported package | S14 / chain | S13 | S09 | Optional deployment wiring and immutable image input | Implicit trust, permissions errors or accidental profile activation | Base/claimed package setup and negative configuration checks pass. |
| 3: Qualify the supported package | S09 / chain | S14 | S15 | Disposable lifecycle/recovery qualification and accountable cleanup | Unsafe target selection, incomplete recovery or leaked resources | Complete isolated scenario and cleanup pass using packaged binaries. |
| 3: Repair discovered recovery boundary | S09a / prerequisite | Reproduced during S09 | Resume S09 | Confined export-root capture/restore and target admission | Source/target overlap or unbound recovery proof | Exported bytes survive recovery; fresh bindings, isolation and cleanup pass. |
| 4: Repair validation prerequisite | S15a / prerequisite | Reproduced during S15 | Resume S15 | Deterministic presence-capture framing | Stable screenshot at the wrong scroll position | Exact visual row and complete visual suite pass against unchanged goldens. |
| 4: Integrate and hand off | S15 / chain | S09, S15a | None | Release graph, complete validation and final tracker closeout | Helper checks omitted or partial runs aggregated | One complete final-candidate release includes package qualification; tracker completed last. |
| Future capability | S07 / deferred | Target crypto posture and owner amendment | Separate future plan | Validated cryptographic build | Unsupported compliance claim | Not part of this iteration's exit. |

Before each active workstream, change its row to IN_PROGRESS. After its checks finish, record the outcome, compatibility consequence, canonical run location, unresolved prerequisite and next dependency; mark DONE only when its exit passes. **Update this tracker before beginning the next workstream.** Keep the checkpoint concise: link canonical artifacts instead of reproducing logs or requirement-by-requirement narratives. A failed prerequisite is BLOCKED, not DONE. Allowed statuses: TODO, IN_PROGRESS, BLOCKED, DONE, DEFERRED, DROPPED.

### Archived package plan — 7. Proposed refactor slices and exits

All slices below are authorized for implementation by the user request of 2026-10-05. Changes to accounting, package configuration or release acceptance are intentional changes, not behavior silently frozen from the current implementation.

#### S11 — Simplify specification and approval obligations

**Areas / gaps:** specification and documentation; RP-F12/F13/F19. **Dependency:** user implementation request; then S12.

- **Remediation:** remove Table 4-C's human-maintained hashes and RP-REQ-020's compulsory document-byte/revision rebinding. Use owner references and ordinary repository history. Amend the accounting portion of RP-GATE-012, RP-REQ-257/259, RP-AC-049, Table 31-B and the matching Reference Pack companion under Testing Harness TH-HARNESS-REQ-657. Retire obsolete accounting identifiers without renumbering unrelated requirements or reassigning their meanings. Consolidate duplicate promotion bookkeeping in Tables 1-A/28-A and §32 into coordinated owner review/adoption through the existing process; do not create another approval manifest.
- **Documentation:** mark the owner, obligation, limit and security evidence handoffs as historical and remove perpetual refresh requirements. Preserve their old results. Keep actionable trust/configuration/recovery instructions in maintained operational guides. No separate recreation, interchangeability, security-disposition dossier or per-requirement report is required for each change.
- **Rationale / long-term benefit:** reviewers assess actual contracts and behavior without synchronizing multiple accounts of the same work; future phases do not multiply reporting obligations.
- **Compatibility / migration:** process/specification amendment only; no runtime or stored-data conversion. Retain normal review and adoption authority, trust/clock configuration, actual deployment suitability and applicable redistribution decisions. A passing test or tracker edit cannot promote draft/proposed documents.
- **Risk if unresolved:** duplicated records drift and repeated signoffs obscure missing behavior. **Rollback:** revert the coordinated editorial change together if it leaves conflicting owners; do not restore hashes as executable dependencies.
- **Validation / exit:** no active mandatory dossier, document-hash or exhaustive requirement-mapping obligation remains in the amended scope; companion owners agree; meaningful controls remain explicit; Markdown/diff checks pass. Record an actual owner contradiction as BLOCKED and resolve it through its owner before S12.

#### S12 — Remove accounting-only machine contracts

**Areas / gaps:** authored contracts, tests, generation and harness routing; RP-F13. **Dependency:** S11; then S13.

- **Remediation:** remove `contracts/reference-packs/traceability.v1.json`, `fixture-manifests/traceability.v1.json`, the `spec_traceability` schema member and `validateTraceabilityProjection`. Rename the retained fixture test to describe executable fixture contracts; update its selector atomically, retaining its semantic row identity where meaning is unchanged. No compatibility reader remains.
- **Completeness:** replace the 163-manifest assertion with checks that every discovered fixture has an expectation runner or actual owner execution binding, required behavioral families have coverage, and every supported content profile retains valid, malformed and semantic-error cases. Keep strict shapes, required nulls, unique IDs, safe paths, ordered effects, independent vectors and actual scenario assertions. A mapping alone cannot replace execution.
- **Rationale / long-term benefit:** new behavior grows the relevant tests without maintaining unrelated requirement counts or ranges.
- **Compatibility / migration:** internal accounting removal only; no public or stored-data migration. Amend authored inputs and regenerate through Make. Do not generalize the removal to other owners' traceability systems.
- **Risk if unresolved:** zero-unmapped and exact totals continue to imply completeness without checking behavior. **Rollback:** revert the slice's authored/test/generated changes together; preserve behavioral failures for diagnosis rather than weakening assertions.
- **Validation / exit:** no live Reference Pack accounting-only consumer remains; missing, malformed, duplicate and unrouted fixtures still fail; profile/family coverage remains; focused owner, harness, generation, shape and artifact-policy checks pass.

#### S13 — Separate runtime projections from test assets

**Areas / gaps:** authored contracts, generated packaging, implementation boundaries and tests; RP-F14. **Dependency:** S12; then S14.

- **Remediation:** move fixtures, fixture manifests and fixture-only schemas into `contracts/reference-pack-fixtures/`, registered as a separate family using the existing contract-family mechanism. Generate a separate fixture artifact package; allow it only in tests and test-support packages. Keep runtime schemas, algorithms, limits, source/license inputs needed by the packaged Base release and the Base assets in `contracts/reference-packs/`.
- **Boundary:** update input paths and test helpers atomically. Fixture-only schema helpers remain test-side; runtime lookup must never fall back to the fixture package. Derive diagnostic vocabulary from production schemas and explicit diagnostic inputs so a new test cannot expand allowed runtime detail. Update contract-family and backend boundary inputs, then generate through Make. The existing generated-root policy already covers internal/gen; do not add a redundant root.
- **Rationale / long-term benefit:** runtime contents, initialization and diagnostics become independent of corpus growth; the existing generator remains the single mechanism.
- **Compatibility / migration:** internal repository paths/imports change; canonical pack bytes, identifiers, stored provenance and supported public diagnostic behavior remain stable. Do not rename contract IDs merely because an asset moves. No aliases, public API change or data conversion.
- **Risk if unresolved:** fixture growth enlarges shipped artifacts and incidental test schemas affect production policy. **Rollback:** revert asset moves, registry inputs and all callers together; do not add runtime test-asset fallback.
- **Validation / exit:** production dependency closure excludes the fixture package/assets; all fixture runners still execute; Base loading, fixed-byte/digest vectors, consumer/provenance/portability and diagnostic-privacy checks pass; generation and boundary checks pass.

#### S14 — Complete packaged administration support

**Areas / gaps:** deployment packaging, configuration examples, operational documentation and package tests; RP-F15/F18. **Dependency:** S13; then S09.

- **Remediation:** provide an explicit optional administration configuration and Compose integration using existing server/operator binaries, read-only operator-supplied trust mounting and confined incoming-bundle access. Preserve Base-only startup as the default, with administration unclaimed and clock trust false until explicitly configured. Server and operator share configured storage and the existing durable Job worker. Do not add a listener, executor or command family.
- **Build inputs:** pin the application image base by immutable digest through existing toolchain/build ownership and extend its drift check. Keep synthetic trust, fixture bundles and test credentials in disposable test composition, outside shipped defaults.
- **Rationale / long-term benefit:** operators can reproduce the supported setup using the actual package and stable build inputs, without undocumented assembly or special production paths.
- **Compatibility / migration:** optional deployment configuration; no automatic claim, trust installation or activation. No database migration. Retain existing public routes and operator result schema.
- **Risk if unresolved:** shipped permissions, mounts and configuration may fail despite source tests; mutable image inputs undermine reproducibility. **Rollback:** revert optional package wiring and pin changes with their tests; do not mutate retained deployments.
- **Validation / exit:** packaged Base and explicitly claimed configurations start; missing or invalid required dependencies reject safely; operator admission uses the server worker with the same identity/replay semantics; image and sanitized configuration identities are available in ordinary run artifacts. Run applicable package/configuration, toolchain drift and boundary checks.

#### S09 — Automate disposable operational qualification

**Areas / gaps:** deployment tests, recovery integration and harness composition; RP-F15/F16/F17. **Dependency:** S14; then S15. This replaces the old deferred S09 scope with isolated package qualification only.

- **Remediation:** introduce the proposed public Make target `standup-reference-pack-smoke` using existing package/recovery smoke infrastructure, authored target inputs and semantic owner routing. It does not exist at this baseline. Share only narrow setup/cleanup helpers; do not build a new runner framework or a parallel evidence format.
- **Scenarios:** exercise actual packaged binaries for Base-only and claimed startup; HTTP/operator import, initiating actor, replay and explicit activation; missing trust, untrusted clock, malformed input and incompatible-state rejection; pre-admission rejection leaves durable state unchanged, while admitted verification failures retain required Job/audit evidence without partial content, index, trust or activation publication; restart, root rotation and rollback rejection; snapshot/report exact-set binding, destination-trusted portability and historical restore; required-content loss/readiness, pin-protected removal and eventual collection. Retain focused race, cancellation, mixed-refresh and commit-proof tests at their current owner boundaries instead of duplicating every case in the package smoke.
- **Isolation / cleanup:** allocate unique test-owned volumes, database, object-store resources and trust roots; never accept an arbitrary retained deployment as the default target. Exercise destructive initialization only inside that allocation. Attempt all cleanup, retain safe diagnostics and record failures; qualification fails if owned resources remain or required cleanup cannot be verified. Do not retain secrets in artifacts.
- **Rationale / long-term benefit:** the supported lifecycle becomes repeatable executable evidence rather than a manually written rehearsal report.
- **Compatibility / migration:** test infrastructure only; repair any reproduced product failure through its owning contract with a regression case. Preserve source environments and their identities; no reset or conversion of retained data.
- **Risk if unresolved:** package, permissions, restart and recovery defects escape owner tests, and cleanup failures can be hidden. **Rollback:** remove only test-owned resources and revert the scoped harness change; keep safe diagnostics for failures.
- **Validation / exit:** all listed scenarios pass on disposable resources using the package's real entry points, source environments remain untouched, cleanup passes, and canonical artifacts identify the actual image/configuration. Existing `standup-package-smoke` and `standup-operational-recovery-smoke` continue to pass with accountable cleanup.

#### S09a — Repair exported-bundle recovery storage (discovered dependency)

**Areas:** Core recovery specification, implementation, contracts, tests, package documentation. The package portability scenario exposes a real pre-existing mismatch: Incident Bundles publishes into the configured export root, while Recovery inventories those files through the object-store adapter. Backup therefore fails once an export exists. Reproduced by the packaged operator at `20261005T053223Z-p16885` and the focused lifecycle row at `20261005T054026Z-p91140`; the latter identifies `incident_bundles.files` as the missing-object family.

**Remediation / rationale:** inject the confined export-root capability into recovery capture and restore; retain source-owner object-family inventory and exact bytes. Extend target binding, freshness and overlap checks to that root. Do not reroute durable files through unrelated storage or omit exported artifacts from backup. This keeps storage responsibility explicit and makes populated-deployment recovery reliable.

**Impact / risk:** ephemeral restore-target marker and verification bindings must include the newly admitted root; refresh operator examples and generated contracts together. No authored SQL migration or automatic retained-deployment mutation. Leaving this unresolved makes successful export disable backup and prevents faithful recovery. Risks are accidental source-root overlap, incomplete cleanup and missing historical files.

**Validation / exit:** real exported bytes survive capture/restore; absent, overlapping or nonempty targets reject before mutation; wrapper-produced bindings match the target; package qualification and cleanup pass. Record the checkpoint before returning to S09. S15 remains last.

#### S15 — Integrate release acceptance and finish the handoff

**Areas / gaps:** harness routing, validation and documentation; RP-F16. **Dependency:** S09; final workstream.

- **Remediation:** add package, operational recovery and Reference Pack smoke qualification to the authored release dependency graph and canonical release-readiness projections. The first two are helper-only today. Register real service/resource requirements and serialize incompatible fixtures through existing scheduler mechanisms. Do not insert a static pass declaration or aggregate unrelated partial runs.
- **Rationale / long-term benefit:** one release qualifies the supported package, and maintainers read canonical results instead of assembling a new readiness dossier.
- **Compatibility / migration:** release acceptance strengthens; product contracts remain unchanged. Ordinary owner review/adoption remains explicit and cannot be inferred from test success.
- **Risk if unresolved:** helper-only qualification is easily omitted and historical results can be mistaken for the new release. **Rollback:** revert a defective routing change together with its tests; keep the slice incomplete until the complete gate actually includes qualification.
- **Validation / exit:** run `make agent-finalize` before `make test-fast` and one complete `make release-check` on the final implementation candidate. Verify qualification inclusion, accessible artifacts and cleanup. Resolve failures at the responsible owner and rerun affected checks plus the complete gate. Finish Markdown/diff/scope checks and this tracker last. All active slices must be DONE; S07 remains DEFERRED. Describe the result as qualification of the supported production package, without claiming approval for an untested deployment.

### Archived package plan — 8. Validation plan

Use public Make targets from the repository root. The original plan-only update used documentation validation. The authorized implementation follows the complete validation sequence below. Discover exact semantic rows at each workstream rather than copying a static target or row inventory into the tracker.

| Layer | Public command / scope | When | Acceptance |
| --- | --- | --- | --- |
| Discovery | `make task-guide ROLE=module-author OWNER=module.reference_data`; `make explain-test-owner OWNER=module.reference_data`; equivalent affected owners | Before each implementation slice | Narrow semantic routing and affected consumers identified. |
| Focused behavior | `make test-slice OWNER=module.reference_data`; `make service-backed-test-slice OWNER=module.reference_data`, with discovered ROWS | S12–S14 and reproduced product repairs | Independent vectors, hostile input, trust/rotation/rollback, no-egress, privacy, limits, finality and retained four-defect regressions pass. |
| Participating owners | Equivalent owner slices for application/server/operator, Reporting/Composition, Incident Bundles, recovery and affected consumers | As affected | Exact-set binding, actor/replay parity, destination trust, retention and historical restore preserved. |
| Static/public boundaries | `make backend-module-boundary-check`; `make openapi-compatibility-check`; applicable frontend type/import/unit targets | S13–S15 as affected | No runtime fixture dependency, transport drift or changed administration behavior. |
| Authored/generated inputs | `make generate`, then `make generate-drift`, `make json-shape-check`, `make generated-artifact-policy-check`; `make harness-contract` | S12/S13 and harness changes | Generated outputs follow authored inputs; no Markdown dependency or accounting-only remnant; malformed fixtures remain rejected. |
| Build/package | `make toolchain-drift`; `make standup-package-smoke`; `make standup-operational-recovery-smoke` | S14/S09 | Immutable build input, real entry points, configuration and cleanup checks pass. |
| Reference Pack qualification | `make standup-reference-pack-smoke`, authored in S09 | S09; then included in S15 release | Full isolated scenario above passes; no retained environment touched. |
| Gate investigation | `make explain-target TARGET=release-check DETAIL=summary`; `make target-plan-json TARGET=release-check`; `make explain-run RESULTS_DIR=<run-dir>` | Before/after final gate | Package checks actually selected; no missing artifacts, partial pass aggregation or unsupported readiness claim. |
| Final candidate | `make agent-finalize`, then `make test-fast` and `make release-check` | S15 | One complete final-candidate release passes with cleanup. Record RESULTS_DIR use or why retained-run maintenance was skipped. Omit duplicate `make check` only after confirming release includes it. |
| Documentation/scope | `make lint-markdown`; `git diff --check`; `git status --short` | Current document update and final handoff | Review all implementation changes and complete the tracker last. |

No migration is planned; use migration checks only if a separately justified storage change becomes necessary. Do not widen production limits, weaken assertions, exclude failing rows or rerun already completed historical work to manufacture acceptance. Record failed or skipped verification with its reason and canonical artifact location. The proposed new target must be authored before it is invoked.

### Archived package plan — 9. Top-level work tracker

| ID | Work item / workstream | Status | Depends on | Evidence / current disposition | Exit condition |
| --- | --- | --- | --- | --- | --- |
| RP-P02 | Current document update | DONE | User-selected plan | Only this tracker changed; inventory/history verified; Markdown passed at `20261005T042618Z-p49665`; diff/scope checks passed | Initial document-only plan complete; implementation outcomes are recorded below. |
| RP-S01 | Reconcile owner and tracker authority | DONE | Prior iteration | Historical execution below | Prior slice exit passed. |
| RP-S02 | Remove dead interfaces and misleading fixtures | DONE | S01 | Historical execution below | Prior slice exit passed. |
| RP-S03 | Stable owner DTOs and private verification | DONE | S02 | Historical execution below | Prior slice exit passed. |
| RP-S04 | Complete application construction | DONE | S03 | Historical execution below | Prior slice exit passed. |
| RP-S05 | Separate transport and semantics | DONE | S04 | Historical execution below | Prior slice exit passed. |
| RP-S06 | Persistence seams and instance isolation | DONE | S05 | Historical execution below | Prior slice exit passed. |
| RP-S07 | Validated cryptographic build | DEFERRED | Future target posture and owner amendment | Current profile remains non-FIPS | Separate future actual verifier/build and target qualification. |
| RP-S08 | Candidate owner and projection review | DONE | S06 | Historical review below; not recurring reporting | Prior slice exit passed; adoption was not claimed. |
| RP-S09 | Automated disposable operational qualification | DONE | S14 | Rescoped from the prior on-prem deferral; §7 | Packaged lifecycle/recovery scenarios and cleanup pass. |
| RP-S10 | Prior pre-production validation/handoff | DONE | S08 | Release 1261/1261 at `20261005T030730Z-p45700` | Completed last in the previous iteration. |
| RP-S11 | Simplify specification and approval obligations | DONE | User implementation request | RP-F12/F13/F19; §7 | Owner cleanup agrees and retains actual controls. |
| RP-S12 | Remove accounting-only machine contracts | DONE | S11 | RP-F13; §7 | Accounting removed; behavioral fixture completeness passes. |
| RP-S13 | Separate runtime projections from test assets | DONE | S12 | RP-F14; §7 | Runtime closure excludes fixtures; identities/privacy pass. |
| RP-S14 | Complete packaged administration support | DONE | S13 | RP-F15/F18; §7 | Optional package configuration and immutable build pass. |
| RP-S09a | Repair exported-bundle recovery storage | DONE | Reproduced during S09 | Backup inventory reads exported bundles from the wrong storage capability | Confined export-root capture and restore, target binding/isolation, and regression pass; then resume S09. |
| RP-S15a | Stabilize declared presence capture framing | DONE | Reproduced during S15 | Test capture inherits horizontal scroll after density/reload setup | Reset and establish the declared cell/trailing-context anchor after presentation readiness; unchanged golden passes; then resume S15. |
| RP-S15 | Integrated release and final handoff | DONE | S09, S15a | Final release `20261005T070744Z-p19802` passed 1264/1264; all 17 readiness projections and cleanup pass | Complete final release and cleanup passed; tracker completed last. |

### Archived package plan — 10. Session handoff log

These compact entries preserve the original planning/document-update session. Implementation checkpoints follow and link canonical run artifacts; they do not require new dossiers or duplicate requirement inventories.

#### Scope and authority

| Date | State / files | Commands or decision | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 | Inspected AGENTS, planning framework, Domain, Reference Pack/Core/Extensions/Reporting owner context and prior handoffs; only this tracker touched | User selected removal of reports/duplicate gates and the accounting-only catalog; current non-FIPS package with disposable qualification | Ordinary review, owner authority and operational constraints retained. Implementation is a later task, beginning with S11. |

#### Backend boundary

| Date | State / files | Commands or evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 | Reference Data and assembly inventory; `shape.go`, `diagnostic_safety.go`, `base_release.go`, operator admission and readiness | `rg --files`, imports/callers, exact file reads and read-only register comparison | All 196 direct files accounted for. Preserve completed owner boundaries; split only runtime/test assets in S13. |

#### Frontend boundary

| Date | State / files | Evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 | Existing public/admin behavior and prior browser evidence inspected through tracker/owner routing; no rendered UI audit | No frontend feature change in this plan | Retain existing selectors and behavior; run affected tests during later qualification, without unsolicited golden changes. |

#### Contracts and code generation

| Date | State / files | Commands or evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 | Reference Pack §29/31, Testing Harness companion, traceability JSON, fixture schema/test, contract registry/generator and generated collection | Targeted reads/searches show accounting-only validation, fixed 163 total and mixed runtime/test collection | S11 amends owners before S12; S13 uses existing family generation. No authored machine input or generated output changed now. |

#### Tests and harness

| Date | State / files | Commands or evidence | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 | Public routing and package/recovery smoke scripts inspected | `make task-guide ROLE=module-author OWNER=module.reference_data`, `make help-all`, and `make explain-target TARGET=standup-package-smoke DETAIL=summary` / `TARGET=standup-operational-recovery-smoke` succeeded | Both package targets are helper-only. S09 adds Reference Pack qualification; S15 integrates release. |
| 2026-10-05 | Prior release summary and manifest inspected | Read-only artifact inspection: `20261005T030730Z-p45700`, PASS 1261/1261 on its recorded dirty source | Historical baseline only; no fresh product validation in this document task. |
| 2026-10-05 | Current tracker documentation validation | `make lint-markdown` PASS: `.cartulary/test-results/20261005T042618Z-p49665/adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` PASS; inventory/history/status assertions PASS; `git status --short` confirms only this tracker changed | RP-P02 DONE. No documentation verification failed. All implementation slices remain unstarted. |

Initial discovery searches guessed nonexistent `tools/standup*`, `tools/task_surface.json`, `tools/execution_topology.json`, `docs/extensions-subsystem-nlspec.md` and `tools/harness/execution/topology*.json` paths. `rg --files` resolved the actual release scripts, authored task-surface/topology manifests and singular `docs/extension-subsystem-nlspec.md`. These read-only lookup errors changed nothing; no verification failure is hidden by them.

Product tests, generation, package/deployment runs, migration checks, release checks and `make agent-finalize` are skipped for this document-only update. Retained-run maintenance is skipped because RESULTS_DIR is unset; the prior successful release is not rerun or relabeled.

#### Security and authorization

| Date | State / files | Evidence / decision | Result / next action |
| --- | --- | --- | --- |
| 2026-10-05 | Trust/clock contract, owner security/licensing boundaries, MVP image/configuration and smoke cleanup | Non-FIPS target selected; image is tag-only; smoke cleanup suppresses failures | S14 pins build input; S09 confines destructive tests and verifies cleanup. No compliance, adoption, licensing or deployment approval manufactured. |

#### Open risks and next session

| Date | Current state | Remaining risk | Next action |
| --- | --- | --- | --- |
| 2026-10-05 | Document update only; all new implementation slices TODO | Current owners still require accounting until S11; no new target exists or production package has been requalified | Later implementation starts S11, updates this tracker at each checkpoint and ends with S15. Preserve historical records; do not regenerate the four auxiliary reports. |

### Archived package plan — 11. Open prerequisites and deferred production constraints

There is no unresolved user preference blocking this plan. The choices above do not themselves change owner requirements or grant authority to operate an actual deployment.

| ID | Prerequisite / disposition | Why it matters | Needed authority or evidence | Status |
| --- | --- | --- | --- | --- |
| RP-PRE-01 | Coordinated owner amendments before accounting removal | Reference Pack and Testing Harness now require executable coverage instead of accounting | S11 amendments reviewed together; no owner contradiction found | DONE |
| RP-PRE-02 | Disposable package target and release routing | Three package projections and accountable cleanup passed in the complete final release | `20261005T070744Z-p19802`; S14/S09/S15 | DONE |
| RP-PRE-03 | Production owner adoption and actual deployment suitability | Passing package tests does not adopt draft owners or approve a specific environment | Existing document-status process and applicable security/licensing decisions in ordinary review/operational records; no new dossier | TODO |
| RP-B02 | Validated crypto policy and method | Current Ed25519 profile makes no FIPS claim | Future target-specific posture and coordinated owner amendment | DEFERRED |
| RP-B03 | Validated module/build/runtime/environment disposition | General verification tests do not establish compliance | Actual verifier/build and target authority qualification if S07 resumes | DEFERRED |
| RP-B04 | Fixed-module Make routing | No speculative build mode belongs in this iteration | Future authored routing and verifier/process evidence after posture selection | DEFERRED |

Prior RP-B01 owner contradiction and RP-B05 pre-production release blocker were resolved in the completed iteration; their historical evidence is preserved below. They are not reopened by the new plan. Actual retained-environment rollout remains outside this iteration; never use disposable qualification as implicit authorization for it.

### Archived package plan — 12. Binary completion criteria

The original plan-only update RP-P02 is complete; its document-only evidence remains in §10. Implementation now accounts for 197 direct target files plus the explicitly identified adjacent package, Recovery and harness changes.

The implementation iteration is complete only when S11–S14, the rescoped S09, discovered prerequisites S09a/S15a, and S15 are DONE, with S15 completed last, and all of the following hold:

- Compulsory human reports, document hashes, duplicate signoff records and accounting-only requirement catalogs are removed from the amended scope; normal review, owner authority and substantive constraints remain.
- Runtime dependencies exclude test assets; fixture growth does not expand production diagnostics; independent canonical and behavioral evidence remains intact.
- Optional administration is configurable in the shipped package using explicit trust and clock inputs, immutable build inputs and the shared server worker.
- Real packaged lifecycle, portability, recovery, negative admission and cleanup scenarios pass in disposable environments, without touching retained deployments.
- One complete final-candidate release includes package qualification with accessible canonical artifacts, safe diagnostics and accounted cleanup. No partial-run aggregation or static readiness declaration substitutes for it.
- The final tracker contains a compact outcome, compatibility decisions, canonical run location, actual adoption/deployment disposition and deferred S07 prerequisites. S07 remains DEFERRED. Qualification of the supported package is not approval of an untested deployment.

## Historical package implementation checkpoints — 2026-10-05

### S11 started

User authorized the complete plan. The existing tracker edit is preserved. S11 is IN_PROGRESS; subsequent slices remain TODO. Planning clarifications are incorporated: admitted verification failures retain their required durable failure evidence, runtime fixture imports have no fallback, and the existing internal/gen policy covers the new generated package. Actual adoption and retained-environment deployment remain separate from disposable package qualification.

### S11 completed

Removed human document-hash binding and the duplicate owner-amendment ledger; Table 1-A now supplies the single coordinated owner boundary. Retired RP-REQ-259 and RP-AC-049 without reusing IDs; removed the accounting family/range mapping and amended the Harness companion to require executable fixture coverage. Marked all four auxiliary handoffs historical without changing their prior results. Trust, clock, redistribution, security suitability and actual adoption remain explicit. No runtime/data migration or product check changed.

Validation: `make lint-markdown` PASS at `.cartulary/test-results/20261005T043506Z-p58291/adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` PASS; reviewed active Reference Pack/Harness clauses and companion references for retired obligations. Product checks intentionally skipped for this documentation-only slice. S11 DONE; S12 is next.

### S12 completed

Deleted the Reference Pack requirement/criterion catalog and its manifest; removed the accounting family and validator. The renamed executable fixture test keeps its semantic row ID. Coverage now follows the fixture-family enum and supported runtime profiles, checks valid/schema-invalid/semantic-invalid cases, and rejects removal of any required family or case class. Existing strict shape/path/null/effect checks and actual expectation runners remain. No public or persisted-data migration.

Validation PASS: format `20261005T043656Z-p61011` (2/2); generate `20261005T043704Z-p65946`; focused fixture row `20261005T043723Z-p69183` (1/1); generate-drift `20261005T043723Z-p69098` (4/4); harness-contract `20261005T043723Z-p69325` (2/2); json-shape-check `20261005T043930Z-p9424` (3/3); generated-artifact-policy-check `20261005T043937Z-p9986` (3/3). Canonical artifacts are under `.cartulary/test-results/<run-id>/`. S12 DONE; S13 is next. Full product/release validation remains S15 work.

### S13 completed

Runtime contracts and Base assets now remain in `contracts/reference-packs/`; fixture schemas, manifests and vectors live in the separately generated `reference-pack-fixtures` family. Only tests and explicitly scoped test support import that family. The deployable-shape check traverses all three binary dependency closures, and the diagnostic regression rejects fixture-only vocabulary. Existing byte identities and contract IDs are unchanged. The direct target register gains `internal/modules/reference_data/internal/packformat/fixture_projection_test.go` (197 direct files total); generated output remains covered by the existing policy.

Validation PASS: generate-drift `20261005T044305Z-p28233` (4/4), JSON shape `20261005T044326Z-p71527` (3/3), generated policy `20261005T044334Z-p94644` (3/3), boundary `20261005T044418Z-p22106` (3/3), OpenAPI `20261005T044425Z-p26783` (4/4), deployable-shape, Reporting `20261005T044315Z-p42837` (3/3), and Incident Bundles `20261005T044417Z-p21781` (3/3). Reference Data `20261005T044305Z-p28296` passed 22/23 units; the fixture-manifest row failed because moved paths changed lexical ordering. Authored refs were sorted and regenerated (`20261005T044842Z-p44242`); the corrective focused run passed (`20261005T044853Z-p47219`, 1/1). The initial boundary failure (`20261005T044339Z-p10979`) was an intended Indicators test-support import and is fixed by its narrow allowance. Canonical artifacts are under `.cartulary/test-results/<run-id>/`. S13 DONE; S14 starts next. Full final-candidate validation remains S15.

### S14 completed

Added optional administration TOML and Compose overlay, read-only trust/incoming mounts, shared server/operator storage, and operational setup instructions. Base-only defaults remain unchanged. Recovery wrappers accept the same overlay; backup creation now reports failure if application restart fails. The application base image is pinned to the registry-resolved multi-platform digest in the existing toolchain owner, with mutable-image drift regression coverage. The package smoke now runs missing/invalid trust, untrusted clock, claimed startup, and real operator import/replay using fresh test-only signatures, retaining private image/configuration identities.

PASS: `standup-package-smoke` `20261005T050517Z-p46183` (9/9, retained scan included); `toolchain-drift` `20261005T050308Z-p18033` (2/2); toolchain regression in `20261005T045842Z-p26721/harness-smoke-toolchain-pins/tool-run-summary.json`; lint-scripts `20261005T045808Z-p24743` (2/2), lint-shell `20261005T045811Z-p25252` (4/4), Markdown `20261005T050311Z-p21545`. Canonical artifacts remain under `.cartulary/test-results/<run-id>/`.

Corrected qualification failures: missing-bind directory transition (`20261005T045425Z-p15409`), invalid test repository identifier (`20261005T045556Z-p71327`), Docker Desktop stale config inode (`20261005T050009Z-p36974`), and non-private artifact permissions (`20261005T050248Z-p89872`, execution 9/9 but public result artifact failure). The final run supersedes these, not their historical records. An attempted private harness-check name was rejected by Make; its public extended suite was used instead. That suite's toolchain regression passed but its unrelated task-surface-report UI locator assertion failed; investigate during S15. Package-owned Docker resources from the completed smoke are absent; retained development services remain untouched. S14 DONE; S09 starts next.

### S09a completed — exported-bundle recovery

Recovery now receives the confined export-root capability used by Incident Bundles, captures the correct bytes, restores through owner/family dispatch, and includes that destination in freshness, overlap, marker and verification bindings. Marker v4, verification v4 and journal v5 are coordinated with Core 01/Core 04 and generated contracts. Operators must issue fresh markers and verification; historical v4 journals remain readable but cannot authorize replay without the export-root proof. No SQL migration or retained-environment mutation. The additional Reference Data repair compares persisted JSONB transition evidence canonically while preserving exact signed root bytes.

PASS: export storage atomicity/confinement `20261005T054851Z-p88963`; full Recovery owner `20261005T055546Z-p80480` (24/24); Reference Data root rotation `20261005T052713Z-p27573`; generate `20261005T055347Z-p72908`; boundary `20261005T055414Z-p89683`; JSON shape `20261005T055420Z-p3931`; shell `20261005T055428Z-p30621`. Packaged qualification `20261005T060026Z-p34287` passed 11/11, including backup, due verification, actual restore, exact exported bytes and all owned-resource/workspace cleanup receipts. Canonical artifacts are under `.cartulary/test-results/<run-id>/`.

Earlier failures exposed the missing export source, stale marker producer and JSONB byte comparison. Corrected test/build attempts included a helper-name collision (`054700`), missing caller/marker projections (`054818`), a changing-input build invalidation (`055358`), and an assertion that read a deliberately reset verification target (`055412`). The final scenario now admits a fresh generation for actual `restore latest` before asserting recovered state. S09a DONE; resume S09 and validate the two existing smoke entry points before S15.

### S09 completed — disposable operational qualification

Added the public Reference Pack package target and semantic owner row. Its source/destination stacks use shipped binaries and isolated databases, buckets and volumes; scenarios cover HTTP/operator actor and replay behavior, admission versus terminal failures, explicit activation, rotation/restart/rollback/signature rejection, exact historical reporting, destination-trusted portability, matching-package recovery, physical collection and required-content loss. Existing owner suites continue to own races/cancellation and mixed refresh. Shared cleanup attempts every owned resource, records failures and residual groups, repairs only the private test workspace's ownership, and fails qualification on incomplete cleanup. Fault tests cover rejected ownership, failed removal/listing, interruption and workspace failure.

PASS: full Reference Pack smoke `20261005T060026Z-p34287` (11/11 units and eleven scenario groups); existing package `20261005T060306Z-p10194` (9/9); operational recovery `20261005T060428Z-p65685` (9/9); cleanup/release contract `20261005T060317Z-p24153`; script lint `20261005T060323Z-p35884`; shell lint `20261005T060326Z-p39231`. All package cleanup receipts pass, including source, destination and private workspace where allocated. Results are under `.cartulary/test-results/<run-id>/`. S09 DONE; S15 starts now with release graph integration, the known task-surface regression correction, final-candidate validation and handoff. No retained deployment was operated.

### S15 release integration checkpoint

The three packaged checks now have release-tier semantic rows, declared service/capacity needs and a shared volume-capacity constraint. `release-check` and its readiness marker require all three; canonical readiness validation requires their target projections. The regression proves exactly one execution per target, real service claims, fresh evidence and closure before readiness. The unrelated UI launcher assertion now distinguishes creating a locator from commands that require one. Performance roster tests follow authored target membership instead of a fixed count. The moved browser-fixture path comment and direct inventory are current.

PASS: generate `20261005T060718Z-p18901`; release contract `20261005T060847Z-p53121`; harness contract `20261005T060852Z-p53546` (2/2); extended harness `20261005T060954Z-p67031`; generate-drift `20261005T061050Z-p2101` (4/4); Markdown `20261005T061020Z-p83818`; diff check. Initial new routing assertion and stale 47-target assertion failed at `20261005T060749Z-p22783` / `p22784`; both are corrected and superseded by the passing checks above. Public release inspection selects 1264 units, including the three package rows. These focused results do not replace the required full release. S15 remains IN_PROGRESS; run finalization, fast tests and one complete final-candidate release next. Retained-run maintenance is skipped because RESULTS_DIR is unset.

Finalization passed at `20261005T061133Z-p6893` (1/1); fast validation passed at `20261005T061214Z-p11280` (739/739). The complete final-candidate release is now running; S15 remains IN_PROGRESS until it and the final cleanup/artifact review pass.

### S15a — discovered validation prerequisite

The complete release `20261005T061847Z-p6495` failed 1262/1264: one presence screenshot differed and its visual summary failed. All three packaged qualification rows passed, including cleanup. Private artifact-mode inspection through the repository UI review skill confirmed identical presence content but a horizontal grid shift of approximately 75 pixels. Both runs have the same viewport, density, fonts and renderer. The scenario uses nearest-target scrolling after density/reload exercises without resetting its origin; capture readiness verifies the body, not the intended grid framing.

**Areas / remediation:** tests and handoff; establish the declared cell/trailing-context anchor from a known grid origin after presentation readiness, and verify the resulting framing at capture. **Rationale / benefit:** remove dependence on prior scroll history while preserving product behavior and the existing visual acceptance boundary. **Impact:** no product, schema or golden migration. **Risk if unresolved:** release qualification can fail nondeterministically or capture the wrong stable frame. **Validation / exit:** the exact failing semantic row passes against the unchanged golden, private review cleanup completes, then finalization/fast validation and a complete release are rerun. S15 remains BLOCKED pending S15a; no partial release results qualify the candidate.

### S15a completed — deterministic presence framing

The exact failing Collaboration visual row passed against the unchanged golden at `20261005T065951Z-p3426` (11/11 units). The scenario now resets shell/scrollport origin through the existing grid normalization helper, establishes the declared cell and neighboring context after presentation readiness, and verifies identical framing before/after capture. No production UI, tolerance, fixture registry or golden changed. Format passed at `20261005T065944Z-p98137`. The artifact-only UI review terminal receipt at `20261005T065511Z-p94405/ui-review/terminal.json` records `cleanup=complete`; task-owned scratch requests were removed. The two imported PNGs provided image evidence only; canonical renderer diagnostics supplied the matching font/density/viewport context.

S15a DONE; S15 resumes with finalization, fast validation and a fresh complete release. Finalization passed at `20261005T070252Z-p41022`; fast validation passed at `20261005T070313Z-p45275` (739/739). The replacement complete release `20261005T070744Z-p19802` is running. The earlier release's three package projections and all cleanup receipts passed, but its visual failure prevents using it as overall qualification.

### S15 final validation and handoff

The final candidate passed `make agent-finalize` at `20261005T070252Z-p41022` (1/1), `make test-fast` at `20261005T070313Z-p45275` (739/739), and one complete `make release-check` at `20261005T070744Z-p19802` (1264/1264; zero failed, skipped or cancelled units). `make explain-run RESULTS_DIR=.cartulary/test-results/20261005T070744Z-p19802` confirms the canonical identity and 665 cleanup steps with zero failures or blocks. All three package target projections, workbook visual validation and the release-readiness projection pass in that same run. The eleven Reference Pack scenarios and actual restore result are retained under its package artifacts. Separate partial runs are not combined into this result.

Final cleanup review confirmed passing source, destination and workspace receipts and absence of every package-owned container, volume and network. Only the pre-existing `cartulary-postgres-1` and `cartulary-seaweedfs-s3-1` services remain, both still up nine days. No visual goldens changed. The earlier failed release remains recorded and is superseded by this complete passing candidate.

Handoff: runtime Reference Pack contracts remain in `contracts/reference-packs/`, test assets in `contracts/reference-pack-fixtures/`; `deploy/mvp/README.md` owns package setup and recovery instructions. Optional administration still needs deployment-supplied trust and an explicit verified-clock assertion. Recovery target markers must be regenerated as v4 and verification rerun for the export-root binding; current completions use journal v5, with historical v4 readable but not replay-authorizing. Exported bundles now use the confined export root throughout capture/restore. No SQL migration, public-route change or retained deployment operation is required by this iteration.

A separate `make check` was not repeated: release inherits the `ci`/`check` policy closures and all lower semantic tiers through the authored graph. Retained-run maintenance was skipped because RESULTS_DIR was unset during finalization. S07 remains DEFERRED; formal owner adoption, FIPS qualification and approval of an actual deployment are not claimed. Final Markdown passed at `20261005T074809Z-p92909`; diff and scope checks passed, all 17 readiness projections were verified, and lockfiles/visual goldens are unchanged. S15 is DONE, completed last. No required work remains in this authorized iteration.

## Historical completed structural remediation — S01–S10

The following scope, findings, slice definitions, statuses and execution evidence describe the previous completed pre-production iteration. Their references to current work, required report refreshes, production deferrals, uncommitted state or final completion apply to that iteration only. The present §1–§12 control the new plan; in particular S09 is now planned for disposable qualification, and S15 is the new final slice. Original run identities and outcomes are preserved rather than rewritten as current evidence.

### Previous scope and source posture

The current execution supersedes the earlier ledger-only and production-readiness plans. Historical A–F and four-defect implementation records remain below. Existing code already fixes strict composition identities, framework lookup-key deduplication, typed early upload rejection and pre-parse URL admission; this iteration preserves those corrections.

| Scope fact | Decision |
| --- | --- |
| Baseline | Clean `c9f5b366fe0d295057a30aaf6e488af9c587fdcb`, inspected 2026-10-04. |
| Targets | `internal/modules/reference_data/`, `internal/app/referenceassembly/` and affected owner/caller/test boundaries. |
| Authorized work | S01–S06, S08 and S10: specifications, implementation, tests, authored projections when necessary, generated outputs through Make, and handoff evidence. |
| Completion | **PRE-PRODUCTION COMPLETE**: S01–S06, S08 and S10 DONE; S10 completed last with one fresh complete passing release and verified cleanup. |
| Deferred | S07 validated cryptographic build; S09 deployment rehearsal; formal adoption and target security/licensing approval. No deployment, reset or activation. |
| Compatibility | Preserve deliberate trust, identity, public transport, authorization, atomicity, retention and reproducibility contracts. Remove dead/private coupling without deprecated aliases. |
| Data policy | Fresh initialization; preserve applied migrations and fail-without-mutation preflight. No converter, dual writer, historical trust inference or identity rewrite. |
| Cryptography | Existing strict Ed25519 offline verification; no algorithm negotiation, FIPS claim or speculative build mode. |
| Evidence | No second implementation/recreation report required. Independent fixed vectors and real acceptance evidence remain required. |
| Authority | Adopted owners govern their scopes; typed inputs project them. The Reference Pack 0.2.0 candidate remains draft. Domain owns vocabulary only. |

The later [release reliability handoff](release-reliability-remediation.md) reports a complete passing run for `36ed8c239`. The old browser failure is not an outstanding implementation diagnosis. Its referenced run root was absent during planning (`make explain-run` returned ENOENT); neither that report nor older results qualify new changes.

### Previous 5. Coupling and boundary findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| RP-F01: owner provenance contradiction | Core 00 REQ-00-065 and Extensions §1.1 now consistently prohibit executable document provenance | Stale blocker could cause a second incompatible repair | intentional/no_action | Core 00 and Extensions | Candidate repair verified in S01; coordinated production adoption remains unclaimed. |
| RP-F02: Reference Data vocabulary | Required Base, optional imports and overlay/framework wording aligned | Ambiguous feature gating and ownership | resolved | Domain and planning framework | S01 DONE; vocabulary and framework aligned, Reporting keeps templates. |
| RP-F03: dead refresh/wrapper/error surface | Definition-only `NormalizeRefreshRequest`, `ValidateRefreshPackKeys`, `ValidateTrustBinding`; orphaned `referencePackVerificationFailed` tested only directly | Unnecessary maintenance and misleading alternative paths | resolved | Reference Data | S02 DONE; removed without aliases; real rejection tests retained. |
| RP-F04: vestigial and incomplete HTTP service state | `jobManager`, `jobSuccessFinalizer`, `storage`, `limits` stored without Service reads; absent Jobs permits nil coordinator | Invalid construction and transport-owned application lifecycle | resolved | Reference Data/application assembly | S02/S04/S05 DONE; complete construction including registry usage; HTTP takes application. |
| RP-F05: global test hook | `worker_hooks.go`, four integration call sites and execution hook | Cross-instance interference, hidden production testing surface | resolved | Test composition/application assembly | S06 DONE; four instance-scoped barriers and explicit two-server isolation. |
| RP-F06: excessive public verification/format surface | `VerificationAttempt`, `VerifiedContent`, `VerifyCanonicalContainer`; public aliases and embedding | Private algorithm evolution leaks to callers | resolved | Reference Data | S03 DONE; private verification and explicit owner DTO conversions. |
| RP-F07: mixed transport, persistence and application concerns | Coordinator imports HTTP/auth and concrete pool; several constructors create partial integrity coordinators | More coupling as consumers and operations grow | resolved | Reference Data/application assembly | S04–S06 DONE; complete components, private HTTP/receipt/persistence seams and postgres.DB. |
| RP-F08: ambiguous fixture naming | Two different `referencePackBundle` helpers; canonical helper ignores `Signed`; old helper proves rejection | Cleanup can remove valid evidence or imply fake trust support | resolved | Reference Data tests | S02 DONE; retired negative fixture isolated; canonical producer always signs. |
| RP-F09: validated cryptographic deployment | RP-REQ-261/264 explicitly exclude FIPS under the current profile | Unsupported production compliance claim | defer | Reference Pack/Core 04/security authority | S07 deferred by user; retain strict Ed25519 and the current policy until a target posture is selected. |
| RP-F10: framework Base scope | Framework catalog now separates required Base from optional administration | Misleading responsibility split | resolved | Planning framework | S01 DONE; framework reflects required Base and optional administration. |
| RP-F11: retained safety boundaries look legacy by name | Cutover preflight, retired-table checks, negative formats | Removing fail-closed evidence weakens the cutover | intentional/no_action | Migration owner and Reference Data | Keep; do not reintroduce retired runtime support. |


### Previous 6. Workstream sequencing

Each implementation slice is a separate workstream. Execute S01 → S02 → S03 → S04 → S05 → S06 → S08 → S10. Do not require deferred S07/S09 before pre-production closure.

Before work starts, mark its row IN_PROGRESS. After its checks finish, record changes, compatibility impact, command/results/artifact locations and remaining findings, then mark DONE only if its exit passes. **Update this tracker after each completed workstream and before beginning the next.** A failing slice remains incomplete. Status vocabulary: TODO, IN_PROGRESS, BLOCKED, DONE, DEFERRED, DROPPED.

### Previous 7. Implementation slices and exits

| Slice/workstream | Depends on | Remediation | Exit criteria |
| --- | --- | --- | --- |
| RP-S01: Reconcile owner and tracker authority | None | Verify candidate Core/Extensions repairs, review Domain, correct framework, refresh inventory and findings. | Editorial consistency and Markdown lint; candidate labels retained. |
| RP-S02: Remove dead interfaces and misleading fixtures | S01 | Delete unused refresh/trust/error helpers and write-only fields; isolate retired rejection fixtures and remove ignored signing option. | No deleted callers; live negative and focused owner tests pass. |
| RP-S03: Stable owner DTOs and private verification | S02 | Replace public private-format aliases/embedding with owner DTOs and explicit conversions; privatize verifier internals. | Canonical bytes, result unions, portability and provenance unchanged. |
| RP-S04: Complete application construction | S03 | Dedicated admission/integrity components; full coordinator and worker registration in assembly; HTTP receives an application port. | Required-dependency, worker registration, Base/admin and operator parity evidence. |
| RP-S05: Separate transport and semantics | S04 | Owner-local HTTP adapter and typed semantic outcomes; narrow transactional Core receipt bridge. | Exact HTTP/auth/idempotency/publication contract and boundary checks pass. |
| RP-S06: Persistence seams and instance isolation | S05 | Use postgres.DB; consolidate private owner persistence; replace global hook with instance-scoped test dependencies. | Two-instance isolation, race, finality, retention and restart checks pass. |
| RP-S07: Validated cryptographic build | Target security decision | No build/policy change; later owner amendment and exact module/environment evidence required. | Future target-security disposition, verifier/process tests and fail-closed posture. |
| RP-S08: Candidate owner and projection review | S06 | Review all 264 requirements and 54 criteria with sub-obligations; refresh owner/limit/obligation/security evidence. | No unexplained current-behavior contradiction; all sixteen gates have honest dispositions. |
| RP-S09: On-prem deployment rehearsal | Formal adoption and deployment approvals | No deployment/reset/activation; retain future prerequisites and matching-version recovery policy. | Future disposable target and recorded initialization/recovery scenario evidence. |
| RP-S10: Validation and handoff completion | S08 | Final review, finalization, fresh complete gate and controlling tracker completion. | All active slices pass; accessible current evidence; production gates explicitly deferred. |

The main risks are DTO identity drift (S03), incomplete/duplicate construction (S04), altered error/receipt atomicity (S05), and lock/finality or test isolation regressions (S06). Revert each structural slice with its callers if its exit cannot be met; never restore retired acceptance, weaken checks or rewrite retained data. S08 does not promote any draft/proposed document. Deferred compliance and deployment risks remain explicit.

### Previous 8. Validation plan

Use public Make targets from the repository root. Discover semantic rows with `make task-guide ROLE=module-author OWNER=module.reference_data` and `make explain-test-owner OWNER=module.reference_data`; use the equivalent discovery for affected owners. Run narrow checks first.

| Layer | Commands / criteria |
| --- | --- |
| Focused owner | `make test-slice OWNER=module.reference_data`; `make service-backed-test-slice OWNER=module.reference_data`, narrowed with discovered ROWS when appropriate. |
| Participating owners | Affected application, Indicators, Network Flow, Reporting, Report Composition, Incident Bundles, Extensions, Recovery and platform slices through their owners. |
| Boundaries and public shape | `make backend-module-boundary-check`; `make openapi-compatibility-check`; affected frontend type/import/unit checks. |
| Authored input changes | `make generate`, `make generate-drift`, `make json-shape-check`, `make generated-artifact-policy-check`; migration checks if SQL changes. Never hand-edit generated roots. |
| Final gate | `make agent-finalize`, then `make test-fast` and `make release-check`. Avoid duplicate `make check` only when the current release graph includes it. Record RESULTS_DIR or why retained-run maintenance was skipped. |
| Documentation | `make lint-markdown`; `git diff --check`; scope/status review. |

Retain hostile-container/signature/rotation/rollback, no-egress, diagnostic privacy, exact-set/pin, destination trust, historical restore, limits, cancellation, concurrency and commit-proof evidence. Preserve the four already fixed behavioral defects as regression anchors. Do not widen limits, exclude rows or combine partial runs to claim a release pass.

### Previous 9. Top-level work tracker

| Slice | Workstream | Status | Depends on | Evidence / current disposition |
| --- | --- | --- | --- | --- |
| RP-S01 | Reconcile owner and tracker authority | DONE | None | Current baseline, inventory and candidate authority reconciled; documentation checks passed. |
| RP-S02 | Remove dead interfaces and misleading fixtures | DONE | S01 | Dead helper callers absent; focused owner tests passed. |
| RP-S03 | Stable owner DTOs and private verification | DONE | S02 | Canonical DTO and live persistence/profile tests passed; no public format aliases. |
| RP-S04 | Complete application construction | DONE | S03 | Dedicated admission, integrity and portable verification components; assembly-owned worker registration. |
| RP-S05 | Separate transport and semantics | DONE | S04 | Typed semantic outcomes and private HTTP/receipt adapters. |
| RP-S06 | Persistence seams and instance isolation | DONE | S05 | Narrow database ports, private persistence seams and instance-scoped blocking dependencies. |
| RP-S07 | Validated cryptographic build | DEFERRED | Target security decision | User-selected production deferral. |
| RP-S08 | Candidate owner and projection review | DONE | S06 | All 264 requirements/54 criteria reviewed; candidate contradictions repaired; indexes and sixteen gate dispositions refreshed. |
| RP-S09 | On-prem deployment rehearsal | DEFERRED | Formal adoption and deployment approvals | User-selected production deferral. |
| RP-S10 | Validation and handoff completion | DONE | S08 | Finalize, test-fast 739/739 and release-check 1261/1261 passed; cleanup, evidence and documentation verified. Completed last. |

### Previous 10. Session handoff log

| Slice | State / edits | Validation and artifacts | Compatibility / risks | Next action |
| --- | --- | --- | --- | --- |
| S01 | DONE: refreshed 186-file inventory, one row per slice, scoped exits and historical posture; corrected Domain overlay summary and framework Base wording. | `make task-guide ROLE=module-author OWNER=module.reference_data` passed; `make lint-markdown` passed at `20261005T020210Z-p39801` (`adhoc/lint-markdown/tool-run-summary.json`); `git diff --check` passed. | No runtime change; draft/proposed authority preserved. Historical release artifacts unavailable locally. | S02 dead-interface cleanup. |
| S02 | DONE: removed dead refresh/trust/error helpers, refresh-only fields, write-only Service state and retired production contract constant; isolated `retired_format_fixture_test.go`; removed ignored canonical Signed option. | `make format` passed `20261005T020319Z-p42391`; focused `make test-slice OWNER=module.reference_data ROWS=…` (canonical consumers/manifest, requests and verifier/limits) passed 2/2 at `20261005T020344Z-p47637`; no callers of deleted names; `git diff --check` passed. | Internal-only removal; canonical signatures and retired signed/unsigned rejection preserved. | S03 owner DTO boundary. |
| S03 | DONE: added owner DTOs and explicit set/license/provenance/portable/diagnostic conversions; privatized verifier attempt/result/function; adapted test producers and retained-content boundaries. Added canonical round-trip coverage including nil/empty collections. | Focused consumers/requests/verifier passed 1/1 at `20261005T020908Z-p63787`; live all-profile and canonical-persistence rows passed 4/4 at `20261005T020944Z-p64668`; backend boundary passed 3/3 at `20261005T020944Z-p64751`; whitespace passed. Earlier compile-only failures at `020603Z-p54800`, `020713Z-p55576`, `020741Z-p56076`, `020835Z-p61683`, `020854Z-p62880` exposed incomplete conversions and were corrected. | Internal Go representation changes only; canonical identities and stored/wire shapes preserved. | S04 complete construction, including portability verifier. |
| S04 | DONE: admission, integrity and verification constructors return complete dedicated components; only NewCoordinator constructs a coordinator. Application assembly registers the shared worker before route binding. Added missing-dependency, route-free construction and duplicate-worker tests and authored routing. | Construction passed 1/1 `20261005T021555Z-p3335`; live canonical persistence, disconnected Base startup and shared upload/operator parity passed 4/4 `20261005T021555Z-p3346`; boundary passed 3/3 `20261005T021555Z-p3479`; generate passed `20261005T021520Z-p99638`; format passed. Intermediate compile failure `20261005T021330Z-p92249` corrected stale embedded field/test construction. | Internal construction API changed atomically; no stored or wire changes. Optional administration remains independent of required Base and borrowed-resource ownership is unchanged. | S05 transport and semantics. |
| S05 | DONE: private owner-local HTTP adapter now owns routing, auth, request decoding and resource projection; semantic operations return typed rejections/results. A private Core receipt bridge alone maps semantic outcomes to retained HTTP statuses, inside the same publication transaction. Full construction additionally requires RegistryUsage. | Request/consumer checks passed 1/1 `20261005T022322Z-p40010`; live lifecycle/auth/upload/replay/persistence passed 4/4 `20261005T022322Z-p40016`; final constructor refinement passed 1/1 `20261005T022516Z-p68343` and persistence 3/3 `20261005T022516Z-p68349`. Boundary 3/3 `20261005T022118Z-p59439`; OpenAPI 4/4 `20261005T022118Z-p59330`; format/whitespace passed. Compile-only attempts `20261005T022050Z-p33078`, `20261005T022118Z-p59359`, `20261005T022205Z-p75747` and `20261005T022205Z-p75755` failed on moved/missing imports and were corrected. | Go ActionResult now carries AdministrativeVersion. Stored receipt payload/status, canonical proof, routes and public projections are unchanged. Adapter remains private within the owner package, avoiding a new exported conversion API. | S06 database and timing dependencies. |
| S06 | DONE: Reference Data uses postgres.DB; private administrative/frozen-verification persistence separated from models/preparation; index writes retain bounded transactions and historical caller-owned transactions. Deleted global hook; four tests use per-server DB barriers; explicit two-server isolation passes. Reproduced owner-routed browser startup exposed stale runtime-binary default paths; corrected authored topology to build/bin and added cross-layout regression coverage. | Full owner `make test-slice OWNER=module.reference_data` passed 23/23 `20261005T024203Z-p93991`; affected isolation/lifecycle rows 3/3 `20261005T023959Z-p22363`; browser rows 11/11 `20261005T024027Z-p43071`; Extensions finality 3/3 `20261005T023148Z-p9478`; harness-contract 2/2 `20261005T024026Z-p42893`; boundary 3/3 `20261005T023122Z-p7510`; lint-scripts 2/2 `20261005T024143Z-p93506`; final generate `20261005T023937Z-p14178`, drift 4/4 `20261005T024237Z-p31652`, shapes 3/3 `20261005T024237Z-p31656`, artifact policy 3/3 `20261005T024237Z-p31654`. | No SQL/schema, lock order or isolation change. Explicit index-batch transaction preserves prior implicit batch atomicity. Pool retained only in operator assembly because Jobs construction requires it. Scope includes one reproduced harness path repair; no repeat of historical acquisition repair. | S08 candidate-owner review. |

| S08 | DONE: all requirement/criterion bodies, candidate owners and projections reviewed; stale error/signer/lifecycle text corrected and evidence indexes refreshed. | 264 requirements, 54 criteria, sixteen gate dispositions; Markdown passed `20261005T025943Z-p43994`; detailed checkpoint below. | No executable identity/schema or stored/wire change; no production adoption. | S10 final acceptance. |
| S10 | DONE: complete gate, cleanup, evidence indexes and final tracker closeout passed. | Finalize 1/1, test-fast 739/739, release-check 1261/1261; final Markdown/whitespace passed; exact identities below. | No further production edits or migrations; production gates remain deferred. | Authorized pre-production effort complete. |

S06 failed attempts retained: `20261005T022928Z-p97716` caught stale concrete-pool test calls/missing import; corrected. Full-owner attempts `20261005T023148Z-p9477` and `20261005T023148Z-p9479` exposed the stale browser binary path and a new two-server fixture lease deadlock. The two stuck test processes were stopped with SIGQUIT for diagnostics; the fixture now allocates a distinct second bucket. Both runs report successful cleanup (27 and 17 steps). `make explain-run` initially failed while no completed summary existed, then succeeded for both. `make harness-smoke-execution-topology` has no public rule; it was replaced with public `make harness-contract` and the actual browser owner run. None of these partial runs is acceptance evidence.


#### S08 completion checkpoint

S08 is DONE before S10 begins. Reviewed every requirement body RP-REQ-001–264, all 54 acceptance criteria and sub-obligations; the [obligation index](reference-pack-obligation-evidence.md#s08-requirement-and-sub-obligation-review) records each inclusive range and its evidence boundary. The [owner inventory](reference-pack-owner-inventory.md) has a separate disposition and responsible authority for every RP-GATE-001–016. Refreshed the limit and focused security indexes and inventoried all 196 current direct-target files.

Additional candidate-only findings resolved: RP-REQ-073 still demanded the retired scalar signer; Table 26-C named a retired set-limit error family; Core 01's generic registry duplicated pre-major-2 errors; Core 04 REQ-04-041, AC-035/093–095/272/309/310/369/443 and LF-07 retained old signer, activation, refresh or filter behavior. These now agree with the current route owner and typed/OpenAPI projections. Registry definitions are referenced at their exact owner instead of copied again. Clarified RP-REQ-135/Table 18-A producer obligations for framework/enrichment subject/type continuity: deployments enforce immutable version-qualified identity and intra-pack constraints, not an inferred global subject registry. This prevents deployment history from deciding source semantics while retaining producer accountability and pinned provenance.

Compatibility: no public/schema/algorithm identifier, signed bytes, stored identity or runtime behavior changed in S08. No data migration or generated artifact change was required. Core 01/Core 04/Domain human SHA references were refreshed; candidate draft/proposed labels remain. Production approvals are pending. Existing real signature, profile/lookup, lifecycle/error, pin/provenance and OpenAPI tests cover the retained contracts; S10 reruns the integrated graph.

Validation: `make lint-markdown` passed at `20261005T025943Z-p43994/adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` passed. `make explain-target TARGET=release-check DETAIL=summary` succeeded and resolves 1261 units/2023 semantic rows, graph `sha256:0f2f8449a6f08477faab9e44a2423e2ed16fc2b51b980b9995c4eff4f0930e2d`. Public target ownership explicitly includes check within release-check; a duplicate standalone check is unnecessary. There is no qualifying retained full warm check RESULTS_DIR for this candidate, so finalization will report retained-run maintenance skipped. Next dependency: S10 final validation and handoff; S07/S09 remain DEFERRED.

#### S10 final validation and handoff

The final machine-source candidate passed one complete release. This run, rather than historical reports or a collection of retries, is the pre-production acceptance basis. All production changes were complete before finalization; later edits are documentation only.

| Validation | Result | Exact accessible artifact |
| --- | --- | --- |
| `make agent-finalize` | PASS, 1/1; generated outputs unchanged | `.cartulary/test-results/20261005T030036Z-p46135/run-summary.json`; `unit-artifacts/finalize-summary.json` |
| `make test-fast` | PASS, 739/739; zero failed/skipped/canceled | `.cartulary/test-results/20261005T030116Z-p50484/run-summary.json` |
| `make release-check` | PASS, 1261/1261; zero failed/skipped/canceled; 2077767 ms | `.cartulary/test-results/20261005T030730Z-p45700/run-summary.json`; `target-summaries/release-check.json` |
| Final focused owner/consumer evidence inside release | All passed: Reference Data 38 rows, Reporting 16, Composition 2, Incident Bundles 26, Recovery 27, server 37 | Same release root `rows/` and `unit-results/` |
| Security/static/contract/browser evidence | Go lint, vulnerability checks, both gosec targets, boundaries, OpenAPI, frontend, browser/a11y/visual and generation/schema/migration checks passed | Same release root `unit-results/`, `unit-logs/` and target summaries |
| Cleanup | All 708 release actions and all 731 test-fast actions completed; zero cleanup failures | Each run's `cleanup-results.json`; all release-summary artifact references resolve |
| Check inclusion | All 1404 check semantic rows and every non-runner policy unit included in release; Go units may regroup | `make target-plan-json TARGET=check` and `TARGET=release-check`; retained `.cartulary/reference-pack-remediation/check-plan.json` and `release-plan.json` |
| Documentation | Evidence indexes and final tracker Markdown passed | `.cartulary/test-results/20261005T034450Z-p32657/adhoc/lint-markdown/tool-run-summary.json`; final `.cartulary/test-results/20261005T034656Z-p35172/adhoc/lint-markdown/tool-run-summary.json` |
| Scope/whitespace | `git diff --check` passed; all 87 changed paths belong to this effort | 70 modified, 3 deleted, 14 new; no staged changes, no commit/push |

Exact final release identities from `run-manifest.json`:

- Source commit: `c9f5b366fe0d295057a30aaf6e488af9c587fdcb`; source state: `dirty` (the reviewed working-tree remediation).
- Machine-source digest: `sha256:b5d0f1c25905104352a03faae5c571510902d31ac89b4b16d1632dabe671d6ae`, identical for finalization, test-fast and release.
- Toolchain digest: `sha256:461a0521f7b3649ff41783a87f34a739b043899881ee43995da8642c47879992`.
- System digest: `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`.
- Executed release graph digest: `sha256:f533f1f21b08b8fcb669e767152363ce1fbde7bba8c1b2abe9e5087cd0d66824`. The S08 target-plan digest above is a planning identity, not this executed-run identity.
- Cache mode: normal; 724 hits, 120 misses, 417 bypasses. The complete current run retains its own result/accounting; no partial historical run supplies missing acceptance.

No final-gate command failed. Earlier slice failures, the initial two-server fixture deadlock, nonexistent public target and absent historical release root remain recorded with their original identities. Retained-run maintenance in `agent-finalize` was **skipped because RESULTS_DIR was unset**: canonical retained-evidence drift, scheduler event-order and warm timing maintenance were not selected. A duplicate standalone `make check` was deliberately omitted because the current release graph includes every selected check row and policy unit. No deployment, reset, activation rehearsal, FIPS build, owner-adoption action or target security approval was attempted; these are excluded production gates, not missing pre-production tests.

The focused [security review](reference-pack-security-review.md) covers the changed construction, DTO, HTTP/receipt, database/index and scheduling boundaries plus the four retained behavioral regressions. The [owner inventory](reference-pack-owner-inventory.md), [obligation review](reference-pack-obligation-evidence.md) and [limit index](reference-pack-limit-evidence.md) separate implemented behavior, actual execution, human projection review and authority decisions. The direct 196-file inventory above is current. Applied SQL, OpenAPI release history and dependency lockfiles were unchanged.

Compatibility handoff: internal callers now use root DTOs, complete administration construction, a prebuilt HTTP application port and `postgres.DB`; the deleted helpers/test hook have no aliases. Receipt bytes/statuses, route/error/authorization behavior, canonical identities, trust, atomic publication, immutable provenance and stored schemas are preserved. No converter or data migration is required for these structural changes. The only adjacent implementation repair aligns authored runtime-binary defaults with Make's `build/bin` outputs and has harness regression coverage.

S07 and S09 remain DEFERRED. Their named authorities and exact resumption prerequisites are in [resuming deferred production slices](reference-pack-owner-inventory.md#resuming-the-deferred-production-slices). The candidate remains draft 0.2.0/profile 2, Reporting/Composition amendments remain proposed, and no production conformance or compliance claim is made.

Final tracker closeout: `make lint-markdown` passed at `20261005T034656Z-p35172`; `git diff --check` passed, and every current-session run root referenced above resolves. S10 is DONE as the final completed slice. All authorized pre-production exit criteria pass. No active implementation work remains; only the explicitly deferred production gates await their authorities and prerequisites.

### Previous 11. Production gates and remaining risks

The current sixteen-gate disposition and named responsible authorities are in the [S08 owner inventory](reference-pack-owner-inventory.md#s08-gate-dispositions-and-responsible-authorities). This supersedes gate descriptions in historical execution records. None of those records independently grants adoption.

| ID | Disposition | Evidence required to resume |
| --- | --- | --- |
| RP-B01 | Candidate owner contradiction repaired; no implementation blocker | S08 completed coordinated projection review; production adoption remains an authority decision. |
| RP-B02 | DEFERRED with S07; current FIPS exclusion remains intentional | Target posture decision and corresponding owner amendment before build-policy changes. |
| RP-B03 | DEFERRED production security disposition | Exact target module/build/runtime/environment and security authority decision. |
| RP-B04 | DEFERRED fixed-module Make routing | Authored harness routing and actual-verifier evidence if S07 is resumed. |
| RP-B05 | RESOLVED for pre-production | One complete final-candidate release passed with accessible artifacts and cleanup; historical reported success is not substituted. |

### Previous 12. Binary completion criteria

Pre-production completion requires S01–S06, S08 and S10 DONE, with S10 completed last. It requires no dead/misleading facade paths, complete construction, narrow owner DTOs, transport separation, instance-isolated scheduling tests, preserved intended semantics, reviewed owner/projection/evidence boundaries and one fresh complete passing release gate with cleanup accounting. The final handoff records exact source identities, artifacts, skipped/failed checks and how to resume production gates. No executable evidence depends on Markdown. S07/S09 remain DEFERRED; formal adoption, production security/licensing approvals and deployment readiness remain unclaimed.

## Historical planning-session log

The following entries describe the earlier planning-only session and do not control current scope or status.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | Approved planning-only write; baseline refreshed | Tracker touched; AGENTS, framework, NLSpec reference, Reference Pack/Core/Extensions/Domain and prior handoffs inspected | `git status --short`; `git log -1 --oneline`; baseline diff; targeted reads | Clean initial tree at `54f421955`; production target unchanged from `18c093002` | RP-B01; draft policy mismatch RP-B02 | Implement S01 only in a later authorized implementation task. |

### Backend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | 181 direct target files accounted; concrete cleanup and boundary findings recorded | Target package/assembly, selected source-owner ports and server/operator callers; no source edits | `rg --files`, caller/import/declaration searches and selected source reads | Confirmed dead helpers, global hook, partial constructor and private-type exposure | No missing target path | Preserve existing algorithms; implement S02–S06 with focused characterization. |

### Frontend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | Existing administration behavior frozen; no frontend change planned independently | Prior UI evidence and owner routing; no new rendered-UI audit | `make explain-test-owner OWNER=module.reference_data`; public target discovery | Three browser and five frontend rows present | None for document update | Reinspect exact affected callers if adapter/DTO changes reach generated public types. |

### Contracts and code generation

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | Formats and operational versions retained; owner promotion planned | Owner/version inventory, generated policy, recognition facts, generator/fixture baseline diff | Read-only diff and searches | Current generator source-escaping change preserved; no generated output edited | RP-B01/RP-B02 | Refresh exact final editorial revisions after owner amendments; no runtime Markdown dependency. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | Read-only routing refreshed; product tests not rerun | Owner catalog, Reference Data family and public task inventory | `make task-guide ROLE=module-author OWNER=module.reference_data`; `make explain-test-owner OWNER=module.reference_data`; `make help-all` | Passed; 37 rows / 15 service-backed | Dedicated restricted-module route must be authored in S07 | Use current owner routing per slice; historical release remains historical. |
| 2026-10-04 | Ledger update | Document validation passed | This ledger only | `make lint-markdown`; `git diff --check`; `git status --short` | Passed; lint run `.cartulary/test-results/20261004T133610Z-p95493`, summary `adhoc/lint-markdown/tool-run-summary.json`; only this ledger modified | No blocker to the document update | Production slices remain unstarted; prerequisite authority work is next. |

Product tests, generation, migration/release checks, deployment rehearsal and `make agent-finalize` were not run for this documentation-only edit. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. The successful prior release remains historical evidence, not a result of this session.

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | FIPS scope limited to Reference Pack verification; no approval manufactured | RP-REQ-261/264, prior security handoff, toolchain pin/module metadata and packaging; linked Go/NIST sources reviewed during planning | Read-only source/configuration inspection | Selected fixed module; process-wide effect and target-environment limits explicit | RP-B02/RP-B03 | Amend owner policy, execute real verifier/build evidence and record actual disposition. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-04 | Ledger update | This session changes planning support only | This ledger | Read-only exploration and documentation validation | No code, owner status, data or deployment mutation | Adoption and deployment evidence remain future work | Start with S01; do not repeat the completed A–F implementation or request excluded recreation reports. |


## Historical completed remediation — A–F

The remainder preserves the prior remediation's implementation and evidence record. Its “current” and “final” statements describe that completed iteration, not the planned work above. The source digest and uncommitted-tree descriptions identify the original runs and are intentionally not relabeled as results for `54f421955` or a future refactor.

**The authorized pre-production remediation is implemented and validated.** HTTP administration and lifecycle Jobs call the canonical application coordinator. The legacy checksum-labelled verifier, mutable store implementation and empty built-in generator have been removed. The live verifier uses bounded confined storage, strict JSON and offline Ed25519 verification. Production HTTP and browser acceptance fixtures use canonical signed containers. The final coordinated release run passed all 1,259 execution units, with no failures, skips or cancellations.

Migrations 00046–00063 implement the immutable stores, exact Reporting binding, identity preflight, audit/diagnostic history, portable catalogs and preparation, retired-store removal, and immutable imported Reporting files. They have been exercised only in disposable PostgreSQL databases. No real deployment has been migrated or reset; no incident identities or historical provenance have been converted. Reference Pack profile major 2, lifecycle Jobs v2, configuration v2, incident-bundle format 5 and recovery projections are implemented.

The production consumer boundary now serves all sixteen profiles and nine algorithm families. Indicators, observations, Network Flow, import and revision mutations use that boundary and transactional usage guards. Native snapshots pin exact sets through rendering. Incident portability validates embedded bytes under destination trust in its parent finalizer, reuses exact local content without changing activation, and retains/re-exports immutable source Reporting artifacts. Recovery validates historical signatures, successful-time freshness and retained source bindings. Physical collection and backup retention now serialize across processes and preserve successful containers, pins and removal history.

The final visual and release gates are complete. Production conformance separately requires owner adoption and the target deployment's cryptographic compliance disposition; this handoff does not claim either. The project owner explicitly removed independent recreation and interchangeability reports from this pre-production cutover on 2026-10-04; those reports are not outstanding implementation work. The focused implementer security review is recorded separately. The [obligation evidence index](reference-pack-obligation-evidence.md), [limit evidence index](reference-pack-limit-evidence.md), [owner/version inventory](reference-pack-owner-inventory.md) and [security disposition](reference-pack-security-review.md) record the implemented boundaries and the distinction between executed evidence and approval. There are 163 fixture manifests across all 27 required families. Reporting identity and admission binding, live concurrency and commit faults, complete reachable boundary recipes and all five closed consumer machine contracts have focused and coordinated release evidence. A fixture count does not establish exhaustive scenario coverage. The subsystem NLSpec remains draft; its formal production conformance gates are unchanged. Dated checkpoints below are historical; this current boundary supersedes their descriptions of missing implementation or validation.

## Fixed decisions

- Replace the legacy pack format without a compatibility reader or dual writes.
- Correct never-adopted schema identifiers in place; increment adopted contracts when observable behavior changes.
- Evaluate freshness once at verification-attempt start, excluding Jobs queue time.
- Reverify and refresh require a successful retained envelope; never-successful candidates require explicit reimport.
- Verification preserves administrative disablement; explicit activation can clear it.
- Complete exact-set consumer, reporting, portability, recovery, and removal integration before claiming completion.
- Never silently discard incident data, invent historical trust, or rewrite indicator identities during migration.

## Current obligation disposition

All rows have passed the final coordinated implementation gates. “Implemented” below identifies live production code, not a production deployment approval or conformance claim. The project owner has scoped this work to pre-production. Exact test routes, retained results and evidence limits are in the linked indexes.

| Plan gap | Implemented boundary / evidence | Pre-production disposition and scope |
| --- | --- | --- |
| A1 | Companion amendments, profile/configuration/Job/public/portability/recovery version inventory and pending OpenAPI change set. | Exact companion content revisions and typed projections are recorded; production adoption is unclaimed. |
| A2 | Strict original JSON, closed nested schemas, explicit null/default policy, normalized alias rejection; nested mutation evidence for manifests, profiles, envelopes, provenance and consumer results. | Nested admission, nullable fields and union branches reviewed against the typed projections; focused and final coordinated checks passed. |
| A3 | All 49 live ordered checks, applicability, bounded exact deduplication, safe diagnostics and durable summaries. | Projection/precedence/privacy review complete; fast/check pass and complete release passed. |
| B1 | Bounded confined ZIP/ustar/GZIP, original names/types, collisions, exact ratios and crash cleanup. Complete structural and configured-limit fixtures are indexed. | Focused implementer security review complete; no additional independent audit is required by the plan. |
| B2 | Live offline Ed25519, root union/both thresholds, 128 root-update signatures, exact historical replay, rollback and complete signer sets. | Focused trust review complete. RP-GATE-010 remains a target-production compliance decision. |
| B3 | Execution-start freshness, clock assertion, one monotonic cohort budget, cancellation/recovery and Extensions finality; queue/expiry/commit-fault evidence. | Fast/check and complete release regression passed; production owner adoption unclaimed. |
| C1 | Candidates, attempts, immutable successful envelopes, independent health/disablement, eligibility and interrupted work. | Transition/projection review and focused plus broad regression passed. |
| C2 | Atomic mixed cohort outcomes, preserved failed renewal, invalidation/fallback, trust/attestation/audit publication and lost-acknowledgement classification. | Shared-finalizer regression passed; focused review includes parent transaction changes. |
| C3 | Durable cohort/dependency/revision captures and ordered guards; live shared-root, activation, usage and pin/removal race evidence. | Focused race evidence and complete release regression passed. |
| C4 | Fixed-point dependency pruning and mandatory Base fallback; startup/content-loss and effective-registry evidence. | Current Host/Evidence owners expose no custom-token assignment fields and Indicator admits the nine Base-preserved types. The normative scope argument is explicit; future assignment owners must adopt token-preserving degradation before expanding that boundary. |
| C5 | Immutable tuples, envelopes, roots, sets, first-success provenance and operation-idempotent attestations; collision/replay/collection evidence. | Database/recovery, fast/check and complete release regression passed. |
| D1 | Versioned event IDs with rightmost parsing and 266/192/184-byte profile bounds; longest identities pass actual Get/lookup. | All sixteen profiles pass the actual consumer boundary and complete release regression. |
| D2 | One evaluator and transactional usage port across Indicators, observations, Network Flow, import and revision mutations; nine families and explicit identity preflight. | Caller and functional browser regression passed. |
| D3 | Five consumer operations with authored closed request/result schemas, shared validation, staged precedence and confidential 900-second cursors. | Closed result unions, downstream callers and complete release regression passed. |
| E1 | Three original nonempty release-bound Base packs, sixteen implemented profiles, all deployment-mode reconciliation and imported-selection revalidation. | Distribution inventory reviewed: project-owned Base and synthetic optional fixtures; no external corpus is redistributed. |
| E2 | Admission-pinned snapshots and frozen render identities, incident-bundle v5, destination trust, exact inactive reuse and retained source artifacts. | Reporting/portability regression passed; exact companion content revisions recorded; production adoption unclaimed. |
| E3 | Historical-time recovery and retained-reference inventory; cross-process backup/collection serialization and exact reimport protection. | Recovery/process regression passed; exact companion content revisions recorded; production adoption unclaimed. |
| E4 | HTTP/Job/operator coordinator, typed metadata/errors, administration details and common audit/telemetry publication. | Client/type/import-boundary/unit/functional/a11y checks, reviewed visual refresh, two ordinary visual comparisons and final release passed. |
| F1 | Obligation and per-limit evidence indexes, 163 manifests/27 families, independent vectors, explicit unreachability arguments and disposable reset rehearsal. | Independent vectors, limit arguments, pre-production scope disposition and complete final validation are recorded. |

## Final coordinated verification

All run roots below are under `.cartulary/test-results/`. Each retains `run-summary.json`; the full release also retains its source manifest, unit results, target summaries and visual reconciliation. Failed and canceled diagnostic runs remain recorded in the historical checkpoints.

| Command | Final result | Retained run |
| --- | --- | --- |
| `make format` | Passed 2/2 after the last browser fixture preparation change. | `20261004T063334Z-p30649` |
| `make generate` | Passed after the last authored source change. | `20261004T063341Z-p35615` |
| `make agent-finalize` | Passed 1/1 before the complete release rerun. | `20261004T063610Z-p76079` |
| `make test-fast` | Passed 737/737 on the final production implementation; subsequent changes affected browser fixtures, goldens and documentation. | `20261004T044530Z-p34338` |
| `make check` | Passed 990/990 on the final production implementation. | `20261004T045139Z-p26860` |
| `make browser-e2e-visual` | Passed 12/12; first ordinary comparison after reviewed golden promotion. | `20261004T055013Z-p41727` |
| `make release-check` | Passed 1259/1259 in 2148.180 seconds; zero failed, skipped or canceled units. Includes the second ordinary visual comparison and final browser fixture changes. | `20261004T063630Z-p80139` |
| `make lint-markdown` | Passed after the completed handoff edits and again after the result-only ledger/owner-inventory additions. | `20261004T071627Z-p84809`; follow-up `20261004T071911Z-p87560` |
| `git diff --check` | Passed after the handoff edits. | Local whitespace check; no harness run root. |

The release source manifest identifies uncommitted source digest `sha256:a0078d33960af5195218b0d56d7e856c5fe739e7034126596b2e24330efc2785`, based on commit `78a4effc23aac4b46ca414e11b91e17892f790c7`. This is the tested working tree, not a new commit. Final handoff edits are documentation only. Both ordinary visual runs use identical renderer identity and golden-manifest SHA-256 `8fbbe83aab213faba870618041d0609d31d5d3a1d4563f37fa6025912eca50d9`; both reconcile all 255 active captures and 29 registered fixtures with zero orphan, missing, ambiguous or unresolved mappings.

Retained-run maintenance was skipped because `RESULTS_DIR` is unset. No mandatory execution unit was skipped in the successful release run. No real deployment migration/reset, commit, push or production conformance publication was performed.

## Phase exits and handoff

| Phase | Completed pre-production exit |
| --- | --- |
| 0 — Contracts | Defaults, omission/null semantics, checks, transitions and owner boundaries reconciled; exact companion content revisions and typed projections recorded. Formal owner adoption is not inferred. |
| 1 — Persistence | Appended migrations 46–63, immutable constraints and fail-without-mutation preflight verified. Explicit reset rehearsed twice against a disposable database and cleaned up. |
| 2 — Verification and operations | Live canonical verifier, retained trust, frozen admission and atomic finalization validated with adversarial, timing, concurrency and commit-fault evidence. |
| 3 — Base and consumers | All deployment branches establish Base; five consumer operations, sixteen profiles and nine algorithm families execute through owner interfaces. Current custom-token reachability limits are documented. |
| 4 — Integration | Exact Reporting binding, format-5 portability, historical recovery, retention/collection, operator administration and UI pass focused and complete release validation. |
| 5 — Evidence | Retired production paths removed; projections regenerated; obligation/limit evidence, licensing inventory, focused implementer review and operational handoff recorded. Recreation/interchangeability reports are outside the owner's pre-production scope. |

Use the [owner/version inventory](reference-pack-owner-inventory.md) for changed contracts and retired behavior, and the [development cutover procedure](../guides/reference-pack-development-cutover.md) for preflight, explicit reset, backup and recovery. Historical backups require their matching historical application; reset does not convert them. No unresolved implementation finding is carried into this pre-production handoff. Production adoption and deployment compliance remain separately unclaimed.

## Files and ownership

- `docs/reference-pack-subsystem-nlspec.md`: draft boundary closure, ordered checks, state/freshness/portability decisions, acceptance amendments. This document has not been adopted.
- `contracts/reference-packs/`: authored schemas, profiles, checks, limits, pinned SPDX identifiers, and independent vectors. The generated Go projection is produced through the existing contract-family generator.
- `internal/modules/reference_data/internal/packformat/`: pure archive, trust, manifest, license, identity, profile, normalization, indexing and diagnostic logic.
- `internal/modules/reference_data/internal/packstate/`: pure eligibility, frozen cohort, transition, fallback and revision-comparison logic. These are proposals, not durable publication.
- `internal/platform/canonicaljson/`: strict original-byte JSON admission before canonicalization; duplicate keys and malformed Unicode are rejected without echoing input.
- `internal/modules/reference_data/storage_port.go` and `internal/app/referenceassembly/verification_storage.go`: streaming, root-free verification storage; fresh private workspaces with confined cleanup.
- `internal/platform/rootedfs/`: random-access regular readers and identity-guarded empty-directory removal, with no recursive deletion.
- Contract-family, generation-scratch and test-routing authored inputs are updated; generated outputs are refreshed using Make.

## Historical verification record — foundations (2026-10-02)

Read-only owner discovery completed for `module.reference_data` and `platform.rootedfs`. Focused verification passed for strict JSON, archives, trust, manifests/licenses/set identity, all declared profiles, indicators, diagnostics, pure transitions, workspace isolation/cleanup, and rooted filesystem containment.

Selected successful run roots under `.cartulary/test-results/`:

| Run | Evidence |
| --- | --- |
| `20261002T194459Z-p55946` | Strict canonical JSON admission and existing RFC 8785 vectors. |
| `20261002T202140Z-p42489` | Storage workspace row passed; the combined archive row failed due to the ratio fixture described below. This is not a successful whole run. |
| `20261002T202227Z-p48543` | Rooted filesystem containment and empty-directory rejection. |
| `20261002T202325Z-p54064` | Archive, trust, manifest, indicator and profile rows after ratio-fixture correction. |
| `20261002T202550Z-p59877` | Diagnostic ordering, deterministic retention and operational-abort semantics. |
| `20261002T203151Z-p71770` | Pure verification transitions, refresh freezing, dependency fallback and revision guards. |
| `20261002T203444Z-p77971` | Trust combined-fault precedence and profile validation. |
| `20261002T203658Z-p79788` | Generated contract and topology outputs. |
| `20261002T203911Z-p84435` | Markdown lint. |

A newly authored invalid-text fixture initially used an invalid UTF-8 byte instead of the intended Unicode C1 control; it was corrected to U+0085. The compression-ratio regression fixture initially used a hard-coded ratio that did not straddle its intended two-member boundary; the test now chooses the effective ratio from a fixed 150,000-byte threshold and the complete fixture size. Production limits were not enlarged. Both failures were related to newly authored tests and passed after correction.

`agent-finalize` initially failed because test-routing edits made the topology inputs stale. After regeneration, `json-shape-check` exposed the separate contract-family allowlist; the authored validator now includes `reference-packs`. That validator edit itself requires topology regeneration before finalization. Regeneration and finalization subsequently passed; see the final checkpoint below.

Retained-run maintenance is skipped because `RESULTS_DIR` is unset. No full service-backed, frontend, migration, `check`, `release-check`, or security-review disposition has been established for the coordinated cutover. Those mandatory release obligations are **unfulfilled**, not waived or represented by the focused foundation tests.

## Completed pre-production acceptance sequence

1. Reviewed and transactionally promoted thirteen explained Reference Pack goldens. Preserved the Evidence golden after repairing its capture anchor.
2. Passed two fresh ordinary visual comparisons, including the complete release run after `agent-finalize`; exact results are recorded above.
3. Recorded obligation, version, migration/reset, licensing and focused security-review dispositions. Production conformance and deployment-specific compliance approval remain distinct; no recreation or interchangeability report is required for this pre-production work.

## Historical foundation checkpoint verification

- `make generate`: passed, `.cartulary/test-results/20261002T204057Z-p21931`.
- `make agent-finalize`: passed, `.cartulary/test-results/20261002T204108Z-p27395`; retained-run maintenance skipped because `RESULTS_DIR` was unset.
- `make test-slice OWNER=module.reference_data`: passed, 23/23 execution units, `.cartulary/test-results/20261002T204009Z-p87002`.
- `make test-slice OWNER=platform.rootedfs`: passed, 2/2 execution units, `.cartulary/test-results/20261002T204120Z-p28688`.
- `make test-catalog-check`: passed.
- `git diff --check`: passed before the final documentation update.

These checks validate the current foundation changes and regression boundary. They do not establish that the unimplemented cutover, integrations, or adoption gates are complete.

The broad `make test-fast` regression passed all 736 execution units at `.cartulary/test-results/20261002T204329Z-p32720`. Subsequent archive review added a draft fixed maximum of 10,000 empty directory markers, independent of regular-file counts. This closes collision-index memory growth through logically empty directories without changing logical identity or compression-ratio accounting. Complete TAR fixtures exercise equality and one-over. The subsequent focused archive/profile check covers that follow-up.

The directory-bound follow-up passed generation at `.cartulary/test-results/20261002T205101Z-p27971`, the focused archive/profile slice at `.cartulary/test-results/20261002T205116Z-p35624`, and `agent-finalize` at `.cartulary/test-results/20261002T205121Z-p36073`. `git diff --check` also passed. The 736-unit fast run preceded this follow-up; the focused slice covers its changed production code. The post-ledger Markdown lint passed at `.cartulary/test-results/20261002T204330Z-p32923`; a final lint includes this evidence update.

## Reachability arguments requiring separate guard evidence

- **Pack-set member ceiling:** the current closed catalog contains 16 pack keys and a set permits at most one version per key. A valid current set therefore cannot contain 64 members. The 64-member ceiling must remain a future-compatible structural guard; an accepted current-profile 64-member fixture would contradict key uniqueness or the closed catalog. Use complete fixtures through the reachable 16-key catalog and a separate cardinality-guard test at 64/65. The independent complete 16-key vector and isolated 64/65 guard pass at `20261003T015302Z-p88217`; all 16 profiles pass live consumer integration at `20261003T015842Z-p15435`.
- **Framework relationship ceiling:** every canonical relationship contains a 68-byte `relationship_id` (`rpr_` plus 64 hexadecimal digits), even before JSON framing or any other required field. Five million relationships therefore require more than 340,000,000 bytes, exceeding the 268,435,456-byte payload-file ceiling. A complete valid fixture cannot reach five million relationships. Verify the isolated count guard at 5,000,000/5,000,001 and use complete payload fixtures for reachable file-size boundaries. The isolated count guard passes at `20261003T015302Z-p88217`. A complete uncompressed archive now reaches the 268,435,456-byte file ceiling using a streamed additional notice alongside the valid Base manifest, original license and registry payload; manifest inventory, profile content, exact notice digest and streaming notice admission are checked. A complete one-byte-over archive fails before destination writes. This passes in the Reference Data slice at `20261003T210850Z-p53778`; no production limits were enlarged.

Final `make lint-markdown` passed at `.cartulary/test-results/20261002T205247Z-p40364`. The final whitespace check also passed. No migrations, resets, commits, pushes, or release claims were performed.


## Persistence, Base and consumer implementation checkpoint

This checkpoint supersedes the earlier foundation-only status and its claim that no migration was authored. It does not close any complete A–F workstream.

- Appended migration `00046_reference_pack_immutable_state.sql` introduces candidates, attempts, immutable versions/envelopes/root history/sets/provenance, retained objects, indexes, revisions, operations, pins and release bindings. The application preflight reports incompatible categories before migration; the migration repeats the check under locks. Neither path converts or deletes retained data.
- Added bounded stream staging/publication and a canonical verification engine over confined per-attempt workspaces. The engine is **not yet the live HTTP/Jobs verifier**, and its phase integration and full diagnostic enumeration remain incomplete.
- Base reconciliation now runs before readiness in every deployment mode. Three nonempty compiled registries carry manifests, notices and release digests. Claimed startup admits configured offline bootstrap roots and persists exact root bytes; retained roots cannot be replaced through a configuration edit. Asset provenance/licensing disposition and complete release-binding review remain open.
- Added exact-set consumer operations, immutable first-success provenance, transactional current-set pin acquisition, and authenticated confidential lookup cursors. Request admission is being hardened before downstream owners are connected. Indicators, observations, Network Flow and Reporting still require migration to these interfaces.
- Configuration now owns `reference_packs.clock_trusted` (default false), the required claimed-profile bootstrap path, and the 1,800-second execution budget bounded by 60–86,400. The Extensions configuration contract advances to v4 with an explicit configuration namespace; the Reference Pack configuration contract advances to major 2. Profile major and persisted lifecycle job cutover remain pending.
- Recovery state catalog advances to v2 and the changed graph restore implementation binding to v5; the graph restore algorithm remains v4. The Reference Data contribution inventories retained object references rather than a health allowlist. Historical signature/freshness restore, index reconstruction and collection serialization are still incomplete.

Current validation evidence:

| Command | Result and run root |
| --- | --- |
| `make generate` | Passed, `20261002T222018Z-p25624`; subsequent source/test edits require final regeneration. |
| `make test-slice OWNER=platform.config` | Passed 2/2 units, `20261002T222138Z-p30416`. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3 units, `20261002T222139Z-p32168`: fresh Base, pins, immutable provenance, restart reuse, read-only preflight, migration rejection without mutation. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers` | Earlier consumer checkpoint passed, `20261002T215733Z-p83854`; new request-decoding edits require a rerun. |

The broader Reference Data slice failed at `20261002T221420Z-p56807` (20/23 units) while the configuration adapter still assumed that a profile ID was its namespace. That assumption has been corrected. Legacy integration fixtures and routes still require the coordinated production cutover; no broad passing result is claimed. A concurrent service run at `20261002T221421Z-p57037` failed on a shared harness warm-stamp race and supplies no product evidence; the later sequential persistence run above passed.

The earlier 736-unit fast run predates these changes. Current `agent-finalize`, broad fast/check/release, browser, focused security review and companion-owner adoption remain outstanding. `RESULTS_DIR` is unset, so retained-run maintenance remains skipped. No reset, commit, push or conformance claim has been made.

### Publication preparation follow-up

The draft migration now also retains private index generations and immutable operation-member/repository captures. Index rows are prepared in bounded batches before publication; the public candidate points only to a complete generation bound to exact manifest/payload digests. Per-file index hashes are compared with the signed payload. The version sequence uniqueness scope is `(repository_id, pack_key, pack_release_sequence)`, with built-in sequences separate from operator repositories. A previous per-key-only projection was corrected before any real deployment migration.

Frozen admission and publication guard functions now use repository/key/current-set/usage lock order. Pending work is checked at the selected key boundary before missing-envelope eligibility. Cohort members reference immutable envelopes instead of copying all content into a large Job payload. These functions are not yet integrated with production Jobs admission or finalization; acceptance coverage and coordinator wiring remain mandatory.

Base configuration identity now includes the clock assertion and verification/archive limits. A restart with unchanged configuration should preserve its revision; a policy change must invalidate affected frozen input. Recovery catalog counts advance with the new authored tables; the catalog remains v2 within this unpublished cutover.

- Consumer JSON admission and the independent complete signed-container test passed at `20261002T225513Z-p43403`. The test verifies exact expected digests/signers, start-time expiry equality, untrusted-clock rejection and workspace ownership. Earlier failures identified mistakes in the newly authored fixture (root threshold, hint schema ID, target inventory and required source timestamp); the fixture was corrected without weakening admission.
- The shared Extensions finalizer now consults an immutable success proof after a commit-acknowledgement error. A matching final commit identity, timestamp, result and digest proves success; absence or inability to read the proof does not prove rollback. Its focused service-backed row passed at `20261002T224314Z-p8758`. This is not complete cancellation/deadline or failed-outcome reconciliation evidence.
- The index-generation persistence checkpoint passed at `20261002T223528Z-p69529`; subsequent admission/configuration/sequence changes require the next persistence run.
- Generation passed at `20261002T230013Z-p55007`; later schema and source changes still require regeneration.

The authoritative release block is unchanged: live legacy import verification, the full outcome coordinator, consumer injection, portability, restore, collection, administration projections, owner adoption and focused security disposition are incomplete.

### Coordinator execution checkpoint

The application coordinator now admits immutable operation inputs, starts attempts under the Jobs execution lock, streams verification into private index generations, and retains individual member results in `reference_pack_attempt_members`. It processes a frozen cohort one member at a time. The publication transaction merges compatible root proposals, publishes successful envelopes and failed verdicts, restores mandatory Base registries where needed, retains first-success set provenance, and finalizes the attempt and operation through Extensions. Cancellation and operational aborts use a separate mutation that preserves established health. No transport has yet been switched to this coordinator.

The migration now retains 25 new Reference Data tables. The complete recovery catalog classifies 140 authored tables, including 109 required authoritative tables. Frozen operation headers, cohort rows, terminal attempts and member results have database mutation guards. These are still unpublished changes to migration 00046, exercised only in disposable databases.

Queued imports now finish an immutable container object before admission. The operation's object reference commits with the Job and is included in recovery inventory; temporary upload staging is removed. Successful envelopes reuse that exact retained container. Uncertain admission keeps potentially referenced objects. This closes the earlier temporary-staging reference gap without inventing trust or historical provenance.

The coordinator uses the Auth-owned transaction lock for an idempotency identity before creating a Job or taking owner mutation locks. Attempt startup validates the current Jobs lease in the same transaction that records interruption of an older attempt. Frozen trust loading is bounded independently of accumulated repository history. Reverify checks every retained logical member before reading the successful container; it cannot silently reconstruct lost payload bytes.

Evidence completed before the latest input-retention and attempt-guard follow-up:

- `make generate`: passed at `20261002T234801Z-p52000`.
- `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence`: passed at `20261002T234130Z-p20427`. This covers signed coordinator publication, disabled-state preservation, failed-renewal preservation, active invalidation/fallback, immutable historical sets, leased import, terminal replay, queued expiry and cancellation. It does not establish mixed-cohort, all deadline faults, or live HTTP coverage.
- The preceding leased test failed at `20261002T233928Z-p91752` because its fixed signing-vector clock produced an already-expired PostgreSQL Job lease. Jobs now uses the real lease clock in that fixture while Reference Data uses the independent verification clock. No production lease validation was weakened.
- Shared finalizer cancellation, lost success/failure acknowledgements, proven absence and indeterminate outcome coverage passed at `20261002T231204Z-p14992` under the Extensions finalization row.
- Generation at `20261002T233057Z-p51149` failed because the updated recovery catalog's canonical artifact digest had not been propagated into the generation registry. The authored registry was corrected, and generation passed at `20261002T233321Z-p58653`.

Current mandatory gaps remain live coordinator routing, complete ordered diagnostics, all public/version projections, activation/disable/removal, consumer injection, exact reporting binding, incident portability, historical restore, collection, UI/operator integration, focused security review, companion adoption and release validation. No full-plan completion or conformance claim is made.

### Live administration and retired verifier checkpoint

- Removed the production legacy verifier, mutable store and empty built-in generator. HTTP imports, reverify and refresh now use canonical coordinator admission and Jobs execution; activation, disablement and removal use its administration operations.
- Administrative resources expose health separately from disablement, explicit nullable success metadata, complete signer IDs and current verification timestamps. Core 01 section 17.4 and the pending OpenAPI release describe the changed contract. Historical OpenAPI baselines remain unchanged.
- Administrative JSON admission rejects duplicate keys, malformed Unicode and trailing input before ordinary decoding. Requests have a 65,536-byte bound; normalized reasons have a 4,096-scalar bound. Removal requires a reason. Public diagnostics no longer echo arbitrary unknown member names or internal storage errors.
- Added removal through the existing administration command/replay controller and updated generated operation bindings. Runtime and recovery grants are explicit for the new tables; the initial full owner slice exposed missing runtime grants, which have been corrected in the unpublished migration.
- Focused canonical persistence passed 3/3 at `.cartulary/test-results/20261003T005834Z-p46872`; frontend type checks passed 2/2 at `.cartulary/test-results/20261003T005835Z-p47128`. Generation passed at `20261003T005803Z-p43734` before subsequent HTTP fixture edits. The complete Reference Data slice at `20261003T005539Z-p6721` failed 19/23 with missing runtime table grants and a retired disabled-state expectation; a full rerun remains required.
- Canonical signed HTTP fixtures and lifecycle expectations are being replaced. This work does not close browser, cross-owner consumer, audit, telemetry, security-review or adoption obligations. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.

### Exact binding, portability shape and historical recovery checkpoint

This checkpoint records implementation through 2026-10-03 02:30 UTC. The full A–F remediation remains incomplete and release remains blocked.

- Removed the whole-byte storage API. The confined streaming port is the only live Reference Pack storage path. The Extensions finalizer now classifies synchronous administration commit acknowledgement loss through a durable operation proof, including cancellation after commit.
- Reporting export/derivation/job contracts and render-bundle manifest advance to v2. Snapshot admission acquires a transactional exact-set pin and retains first-success provenance; release/render jobs retain the binding. Missing required content is the closed `required_reference_pack_unavailable` conflict. Committed admission replay precedes fresh content availability checks. Migration 00047 rejects the retired render manifest representation; migration 00046 preflight additionally inventories retained portability Jobs and historical incident-bundle exports without deleting them.
- Incident-bundle format 5 is current-only. The authored source catalog and every source owner accept format 5. The required `reference_pack_refs.v1` object has strict nested schemas, canonical sets, unique/comprehensive version rows and independent fixtures. This closes representation admission, **not** full portability: actual artifact reference discovery, embeddings, destination verification/reuse, pins and redistribution scenarios are still open.
- All 16 profiles now pass actual HTTP import/activation followed by Get/lookup through the consumer interface. All nine indicator families pass independent evaluation/idempotence vectors through that boundary. Indicators, observations and Network Flow still require injection of this interface and retirement of their private canonicalizers.
- Historical validation verifies root transitions against both predecessor/successor thresholds, retained envelope identity, signed bytes at the recorded successful time, logical inventories, Base release bindings, immutable set provenance and required indexes. Restore rebuilds indexes without changing health, disablement or provenance. Startup invokes required-state validation before readiness. Additional corruption and concurrency evidence remains open.
- Recovery now has a separate confined Reference Pack object source/restore port. It restores exact opaque references with streaming digest/size verification, rejects occupied destinations and confines disposable-target cleanup. Restore marker v3 and restore-verification artifact v3 include the Reference Pack root; the unpublished journal v4 completion retains all target binding digests for replay. The full operator acceptance run is in progress; the older process sentinel still needs its legacy fixture updated.

Selected evidence (all run roots are under `.cartulary/test-results/`):

| Command / boundary | Result |
| --- | --- |
| Extensions service slice, finalization cancellation/reconciliation row | Passed 3/3, `20261003T012931Z-p63504`. |
| Reference Data consumer/profile HTTP service rows | Passed 3/3, `20261003T015842Z-p15435`. |
| Reference Data request/manifest unit rows, including portable references and isolated limits | Passed 2/2, `20261003T015302Z-p88217`. |
| Reporting snapshot binding service row | Passed 3/3, `20261003T020439Z-p39461`; includes missing-content rejection and committed replay. |
| Canonical persistence service row with historical expiry/index rebuild and required notice loss | Passed 3/3, `20261003T021251Z-p86460`; later set-provenance validation changes need the next rerun. |
| Reference Pack confined storage row with backup/restore identity and cleanup | Passed 1/1, `20261003T022013Z-p22173`. |
| Recovery marker and verification artifact unit rows | Passed 2/2, `20261003T022610Z-p96854`; later terminal binding evidence changes need rerun. |
| Generation | Passed `20261003T022418Z-p93316`; subsequent journal/schema edits require regeneration. |

Failures retained for attribution:

- `20261003T015819Z-p14118`: generation rejected one unclassified pending OpenAPI description change. The Reporting-owned change was classified; generation passed at `20261003T020400Z-p36259`.
- `20261003T015301Z-p87991`: the new Windows lookup fixture expected one result for omitted version. The normative result includes both versions; only the assertion changed before the passing profile run.
- `20261003T014638Z-p58586`: a new preflight query named the wrong incident-bundle export table. Corrected before the passing migration/Reporting runs. Current migration schema hash is `b11048c0cac06f34b352ad3085ffb02ee5b6c2534c64aea82b12bf9d96eef282`.
- `20261003T020620Z-p59872`: incident-bundle service failures exposed a remaining Timeline format-4 switch. It now admits format 5. `20261003T022015Z-p22521` then passed 5/6 units; the remaining restart fixture reused the database with a fresh Reference Pack root. The fixture now retains that root; rerun is required.
- `20261003T022237Z-p48686`: the packaged-server recovery sentinel failed before readiness because its process fixture omitted the now-required trust bootstrap. That fixture is corrected; its old backup representation still requires migration to the current recovery path.

No real deployment migration, reset, commit or push occurred. Current broad `agent-finalize`, `test-fast`, `check`, `release-check`, browser evidence, focused security disposition, asset licensing disposition and companion-owner adoption are outstanding. The old 736-unit fast run is not evidence for the live cutover. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.


### Recovery round-trip and admission checkpoint

Implementation through 2026-10-03 03:00 UTC remains incomplete overall. The live operator and packaged-server recovery paths now have passing Reference Pack evidence:

- The unpublished migration 00046 no longer seeds an empty current-set singleton. Base establishes it in the same transaction as the first usable set. This leaves fresh migration and disposable-reset targets empty without a restore-admission exception. The current migration source hash is `2513be78d24b2419129665b9416f24bc1f831f52aa547ea08a3827d4e87ad6d1`; only disposable databases were used.
- PostgreSQL restore supplies writable columns explicitly and checks generated values against the backup after PostgreSQL recomputes them. The focused round trip exposed the attempted write to `reference_pack_lookup_keys.lookup_sha256`; no database integrity constraint was weakened. Restore now also rejects missing or additional row columns, and capture admits original JSON without duplicate-key or Unicode repair.
- The process sentinel uses current catalog-driven backup and restore, including Base objects in their separate root, historical validation, derived index rebuilding, graph/workbook reconstruction and actual server readiness. Operator terminal replay retains the root bindings already captured in journal v4.
- Root admission checks filesystem overlap across source packs, generic objects, backups and destination storage. Current journal decoding requires explicit nullable members, validates nested typed shape and rejects lossy JSON. These latest admission changes have focused unit coverage; the preceding operator/process pass predates them and requires a follow-up service run.

| Command | Result and retained run |
| --- | --- |
| `make generate` | Passed `20261003T023832Z-p68594`; later source changes require regeneration. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3, `20261003T024838Z-p8943`, including Base backup/restore, pinned provenance and readiness. Predates the final exact-column shape check. |
| `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9,module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e` | Passed 9/9, `20261003T024953Z-p28590`. |
| `make service-backed-test-slice OWNER=module.incidentbundles` | Passed 6/6, `20261003T025136Z-p67889`. This validates existing format-5 scenarios, not the unfinished pack embedding/import behavior. |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.closed_typed_failure_catalog_2d89c71ab4,module.recovery.unit.exact_recovery_state_catalog_and_synthetic_goose_774b28fd25,module.recovery.unit.graph_restore_v4_contract_projection,module.recovery.unit.owner_registered_workbook_probe_and_v2_artifact_a881cb6053,module.recovery.unit.restore_target_marker_v2_validates_purpose_gener_492d40e257,module.recovery.unit.runtime_resource_lifetimes_9c37e10ad6,module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | Passed 3/3, `20261003T025153Z-p82166`. |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.restore_target_marker_v2_validates_purpose_gener_492d40e257,module.recovery.unit.graph_restore_v4_contract_projection,module.recovery.unit.closed_typed_failure_catalog_2d89c71ab4` | Passed 2/2, `20261003T025709Z-p94402`, including current journal nested admission and cross-root overlap. |

Earlier failures were change-related and fixed: `20261003T023131Z-p27602` rejected the seeded empty singleton; `20261003T023922Z-p76495` and `20261003T024614Z-p84624` exposed generated-column restore; `20261003T024236Z-p21209` used an insufficient retention duration in the new process fixture; `20261003T024458Z-p65593` caught an unused test import; `20261003T025610Z-p93520` caught an incomplete storage-binding test fixture. The later passes above supersede those specific failures. Required broad checks, security review and owner adoption are still outstanding. `RESULTS_DIR` remains unset, so retained-run maintenance remains skipped.


### Shared identity and registry transaction checkpoint

Implementation through 2026-10-03 04:20 UTC remains incomplete overall.

- Indicators now obtains canonical evaluation from Reference Data for explicit creation, observations, Network Flow participants, import, rollback and incident portability. Private type aliases and identity canonicalizers are retired. Raw display and supplied normalized text reach owner evaluation without generic line normalization. All nine algorithm families and a 6,000-character URL are covered through live HTTP.
- Source transactions acquire mandatory registry keys, current-set state and usage guards before incident or record locks. Application composition supplies the guarded transaction opener to Revisions, Imports and the Extensions cross-owner backend. Source owners increment only affected registry usage; unchanged identities and unrelated edits do not create a new usage revision. Dedicated concurrency coverage is still required beyond the current guard/rollback tests.
- Migration 00048 introduces non-null canonical values, type-specific dedupe constraints and bounded digest indexes with exact logical comparisons. The generated active-identity digest is recomputed during restore. No historical key is reinterpreted. Preflight now also rejects retained Indicator identities in intermediate migration-47 databases. The current authored migration source hash is `05bcbeb5a0b6af91ec0a8657c52d4b1a242ec85a6da21536a1f7f5f5600e5ee9`.
- Indicator portable rows and browser creation constraints advance to v2; the Network Flow find-or-create participant is v2. The source-state catalog names exact row schema IDs rather than synthesizing `.v1`. Browser value-kind choices derive from the Reference Pack algorithm catalog. Public raw-value requirements and bounds are recorded in the pending OpenAPI release.
- Administrative metadata projection is being completed with release sequence, source profile/version/time, license, redistribution and repository identity. Signers use an explicit empty array before success. Core and subsystem prose are reconciled on that omission rule. These latest projection changes still need generation and acceptance evidence.

| Command | Result and retained run |
| --- | --- |
| `make test-slice OWNER=module.indicators` | Passed 20/20, `20261003T041702Z-p50745`, after source-schema v2 correction. |
| `make service-backed-test-slice OWNER=module.indicators` | Passed 8/8, `20261003T035910Z-p6775`; predates the final portable-row v2 and browser projections. |
| `make generate` | Passed `20261003T041627Z-p47354`; latest administrative projection changes require regeneration. |
| `make frontend-typecheck` | Passed 2/2, `20261003T040451Z-p35860`; latest administrative fields require rerun. |
| `make frontend-unit` | Failed 687/691, `20261003T040642Z-p63075`. Four change-related failures: retired pack fixtures, derived Indicator defaults/client validation, and source-ownership inventory. Corrections are authored. |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.indicator_canonical_transport` | Passed 2/2, `20261003T041702Z-p50749`. |
| `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.generated_http_operation_bindings` | Passed 2/2, `20261003T041702Z-p50768`; predates the additional metadata fields. |
| `make service-backed-test-slice OWNER=module.networkflow` | Failed 15/36 units, `20261003T040642Z-p63009`. Backend behavior rows passed except a database cleanup timeout. Browser startup lacked the required trust bootstrap; its harness configuration is corrected and awaits rerun. |
| Canonical persistence service row | `20261003T041702Z-p50742` failed before product execution because concurrent owner slices raced on the shared service-image warm stamp. This supplies no product evidence; service-backed runs must be sequential. |

No real data reset or migration occurred. Full diagnostic execution, historical artifact portability, collection/backup serialization, custom-token degradation, operator import, complete public/audit/telemetry evidence, focused security review, licensing disposition and owner adoption remain open. Current finalization and broad release gates have not run. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.

### Live verification, browser and consumer integrity checkpoint

Implementation through 2026-10-03 05:40 UTC remains incomplete overall.

- Browser fixtures now produce canonical Ed25519/TUF containers, with a confined harness bootstrap and an explicit trusted-clock assertion. The real administration lifecycle passed. The duplicate administration unit fixture now uses the shared typed fixture. The closed administrative schema is checked for exact parity with the authored OpenAPI resource, including all eight source/release/license/repository fields added in the cutover.
- Mixed refresh acceptance now covers one successful renewal and one missing active member: one transaction renews the healthy envelope while preserving disablement, marks the missing member, applies Base fallback and retains a failed Job outcome. Replaying the operation creates neither another attempt nor another member attestation; historical set provenance remains unchanged.
- Verification separates manifest shape, logical identity, payload identity, content shape, content semantics and compatibility. Release-sequence and logical-version checks run before content validation; historical restore supplies exact retained identity instead of current release-sequence policy. Every content-schema pass finishes before semantic checks or index construction. Combined-fault HTTP fixtures exercise these winners. This does **not** close the complete 49-check runner, enumerable finding collection or diagnostic safety obligation.
- Consumer cursors retain subsecond timestamp precision for their exact 900-second lifetime. Missing or incomplete indexes fail explicitly. Consumer-detected missing or altered authoritative bytes now invoke coordinator invalidation with revision comparison and atomic fallback. Transactional consumers roll back the complete source transaction before independent invalidation, avoiding a publication-lock deadlock and preventing source writes from committing after content loss. Repeated detection preserves the first event. Required Base loss with no healthy fallback, complete normative attestations/audit and broader races still need closure.
- Recovery rejects the retired whole-snapshot representation before reading its artifacts. The live restore path requires the current state catalog and admitted operation identity. Catalog-driven readiness, metadata-store, tampered-binding and missing-artifact fixtures replace retired restore fixtures. The old whole-snapshot capture helpers still require retirement; this checkpoint does not claim they are gone.

All retained roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| Reference Data live administration browser row (`make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_real_lifecycle`) | Passed 11/11, `20261003T043822Z-p30392`. |
| Recovery browser row (`module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0`) through `make service-backed-test-slice` | Passed 11/11, `20261003T042734Z-p54690`. |
| Recovery service rows for readiness, selected-backup validation, missing artifact and durable metadata creation | Passed 4/4, `20261003T045458Z-p27107`. Exact row selectors are retained in that run's plan. |
| `make generate` | Passed `20261003T053605Z-p94987`, after preceding passes `20261003T050206Z-p47760` and `20261003T051054Z-p78107`. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f,module.reference_data.frontend.the_deployment_admin_reference_pack_panel_starts_1cf94e4f79,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Passed 6/6, `20261003T051614Z-p29302`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T052050Z-p54258`; predates consumer invalidation. |
| `make test-slice OWNER=module.reference_data` | Failed 22/23, `20261003T053638Z-p3186`. New consumer invalidation acceptance passed. One retired symlink expectation was corrected; rerun required. |
| `make frontend-typecheck` | Passed 2/2, `20261003T053638Z-p3262`. |
| `make format` | Passed 2/2, `20261003T053625Z-p98194`; subsequent edits need formatting. |

Earlier change-related failures and repairs remain attributable: browser `20261003T043547Z-p95929` omitted required source time; recovery `20261003T045049Z-p3328` used an obsolete evidence binding and misplaced service selector; Reference Data `20261003T051149Z-p86498` exposed stale administrative fixtures and a combined-fault fixture that unintentionally reused TUF metadata version 1; frontend typecheck `20261003T052051Z-p54519` found an unused import after fixture replacement. Those specific defects were corrected before the corresponding passes above. Generation `20261003T044825Z-p91423` rejected selector ordering; corrected generation passed at `20261003T044955Z-p94699`. The current generation result above supersedes the ordering failure.

No deployment reset, migration, commit or push occurred. Release/security/adoption gates remain open. `RESULTS_DIR` is unset; retained-run maintenance remains skipped.

### Canonical evidence, audit and release binding checkpoint

This checkpoint covers work through 2026-10-03 06:31 UTC. Full-plan completion remains blocked.

- Replaced incomplete lifecycle event objects with the exact closed `cartulary.reference_pack_attestation.v1` shape and content-derived identity. Every required nullable member is emitted. Failed verification does not borrow previous successful signer or freshness evidence. Historical validation checks canonical bytes, identity, relational event binding and operation attribution. An independently hashed fixture exercises identity and nested schema rejection.
- Lifecycle publication now appends canonical activation/rollback, verification, disablement/removal, consumer invalidation, safety fallback, dependency invalidation, exact reimport and trust-update evidence. Rollback derives from prior successful activation or lower sequence in the same repository/key scope. First-success set provenance remains unchanged.
- Migration 00049 and the authored audit registry allow the exact Reference Pack actions. Admission, terminal attempts and semantic publication append raw/projection audit pairs in their owner transactions. Audit carries operation identity and opaque version references, with no payload, source identifier, user reason, path or raw error. Replay and repeated consumer detection do not duplicate events. Bootstrap-root import and the complete rejected-admission audit matrix remain open.
- Core 01's retired table-based lifecycle, disabled reverify rule, blanket failure behavior and open pack-key extension clauses were replaced with the coordinated contract. Adoption gates explicitly include Extensions, incident portability and recovery; exact approved revisions remain unrecorded, so the subsystem remains draft.
- Packaged Base assets now carry the exact eight-member release binding; startup and restore validate its schema, sequence, profile and digest identity. Packaging wrappers remain separate. Reconciliation checks conflicting retained bindings rather than ignoring them. The browser fixture reads the new wrapper. This is a correction of unpublished packaging, not a compatibility reader.
- A narrow operation observer is now wired into live import, verification, refresh/reverify, lifecycle, Base reconciliation and all consumer paths. The telemetry owner and typed registry declare ten exact operation names and one duration histogram. Only operation/result tokens cross the port. These latest telemetry changes require verification before this obligation can be closed.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make test-slice OWNER=module.reference_data` | Passed 23/23 at `20261003T055032Z-p58622` and `20261003T061012Z-p38528`; the latter includes live audit, fallback and rollback assertions. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest` | Passed 1/1 at `20261003T055456Z-p4980`, including the independent attestation fixture. |
| `make service-backed-test-slice OWNER=module.extensions ROWS=module.extensions.integration.job_finalization_cancellation_and_reconciliation_2c712826c2` | Passed 3/3 at `20261003T055547Z-p5606`, including the Base-only mutation finalizer and lost acknowledgement. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13 at `20261003T061253Z-p75768`, before the exact Base-binding follow-up. |
| `make test-slice OWNER=platform.audit` | Passed 4/4 at `20261003T061434Z-p10296`. |
| `make test-slice OWNER=module.reference_data` after Base binding changes | Failed 21/23 at `20261003T062021Z-p39302`; all Go checks passed, but the browser fixture still read three retired wrapper fields. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_real_lifecycle` after fixture repair | Passed 11/11 at `20261003T062254Z-p76628`. |
| `make generate` | Passed at `20261003T060113Z-p27294`, `20261003T061852Z-p30996` and `20261003T063104Z-p13843`. Earlier `20261003T060008Z-p26097` failed because migration 49 lacked an owner assignment; that authored catalog defect was corrected. |
| `make format` | Passed 2/2 at `20261003T061950Z-p34400`; latest telemetry changes still require formatting and verification. |

No real deployment has been reset or migrated. Security authority, licensing disposition, full portability, garbage collection, operator import, complete check precedence/findings, companion adoption and broad release evidence remain mandatory open work. `RESULTS_DIR` is unset, so retained-run maintenance remains skipped.

### Archive, operational failures and bootstrap audit checkpoint

This checkpoint covers work through 2026-10-03 06:58 UTC. Full-plan completion remains blocked.

- ZIP admission now parses one central record at a time instead of constructing an unbounded metadata inventory through `archive/zip.Reader`. It skips comments, bounds names before allocation, checks local/central metadata and intervals, and independently verifies decoded size, CRC and exact compressed-stream termination. A valid 32-member fixture with 1.9 MB of comments proves bounded metadata reads. Malformed CRC, size and compressed-trailer fixtures are rejected.
- The draft and typed limits now bound the terminating TAR zero run to 2–20 blocks, including the mandatory pair. This also applies after GZIP decompression, closing work expansion through non-payload padding. Complete TAR and GZIP fixtures exercise 2, 20 and 21 blocks. Regular-file byte accounting and logical identity remain unchanged.
- Bootstrap root installation retains a terminal attributed operation and administrative audit in the same transaction as the exact root. Reconciliation replay creates neither another operation nor another audit event. Operator/user admission rejection coverage remains open.
- Retained-content checks now distinguish definite absence/prohibited filesystem replacement from permission or device/read failures. Operational failures return unavailable to the caller without condemning established content or publishing fallback. Fault injection covers permission and mid-read failures through live consumers, and failures during archive parsing through the production verification boundary. A decoder cannot erase the underlying I/O failure into a content verdict.
- Proven import-admission replay remains successful if cleanup of its unused newly uploaded copy fails; that copy remains cleanup work. Restart/orphan collection is still incomplete.
- Migration generation and schema ownership validation now import one authored owner allocation. This removes the duplicated maps that disagreed about migrations 47–49.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.operation_telemetry` | Passed 1/1, `20261003T063621Z-p23329`. |
| `make test-slice OWNER=platform.telemetry` | Passed 2/2, `20261003T063644Z-p24623`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T063718Z-p26402`, including exact Base binding restore and mixed-refresh telemetry. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_archive_admission` | Passed 1/1, `20261003T064311Z-p72665`, including bounded metadata and TAR padding. |
| `make test-slice OWNER=module.reference_data` | Passed 23/23, `20261003T064533Z-p79107`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T064720Z-p16130`, including bootstrap-root audit replay. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Passed 4/4, `20261003T065547Z-p67426`, including operational read fault classification and the live lifecycle. |
| `make generate` | Passed `20261003T065453Z-p59435`. |
| `make json-shape-check` | Passed 3/3, `20261003T065656Z-p87861`. Earlier failure `20261003T064853Z-p50473` exposed the duplicated migration-owner map; `20261003T065253Z-p57979` then required topology regeneration. Both are superseded by this pass. |
| `make migration-drift` | Passed 5/5, `20261003T065708Z-p88374`. |
| `make format` | Passed 2/2, `20261003T065527Z-p62705`. |

No deployment reset, commit, push or conformance claim occurred. Portability still emits empty references in its live exporter; actual snapshot section transport, permitted pack embedding and destination publication must be implemented together. Garbage collection, local operator import, complete diagnostics, human adoption/licensing/security disposition and full release evidence remain mandatory. `RESULTS_DIR` is unset; retained-run maintenance remains skipped.

### Closed envelopes, historical retention and invalidation evidence

This checkpoint covers work through 2026-10-03 08:05 UTC. The full plan remains incomplete.

- Added the closed successful-envelope projection, including required nullable success fields, exact nested trust snapshots/proposals, canonical Base64 metadata, root-history keys, sorted complete signer sets, transitions and target binding. All production envelope reads and writes use the same strict admission boundary. Historical restore checks envelope identity against its relational row and requires exact canonical root-transition evidence. Independent signed fixtures exercise nested omission, null, type, unknown-member and binding failures. Signed TUF bytes themselves are unchanged.
- Separated immutable set retention from active-set publication. Retaining a healthy disabled member does not clear disablement or change the active-set revision. The owner retention interface can pin a named historical set and export its exact closed reference catalog; duplicate uses coalesce to one set/version row. Pin acquisition locks the same ordered pack-key guards as removal. The live Incident Bundles exporter still requires integration and must not be described as complete.
- Migration 00050 adds a publication fact to attempt members for first invalidation of established logical content. A narrow column grant and trigger permit only false-to-true publication before attempt completion; prepared verification facts, repeated publication and terminal evidence remain immutable. Refresh/reverify retain distinct payload-invalidation audit and attestation evidence with the complete resulting set. Failed import renewal and operational aborts do not assert invalidation. The current independent migration source hash is `61885889af6e2cbeb9e9d68b8f47807a720441920a8bac28772afadc753a80a5`.
- Updated stale bundle-version, Extensions Job/worker, owner API disposition and migration-evidence fixtures. The administrative audit filter now includes the twenty Reference Pack actions and operation target in both the authored OpenAPI owner and existing UI model. Its breaking parameter fingerprint is recorded in the pending release; historical baselines remain untouched.
- Corrected remaining Core 02 prose that described the retired Reporting v1 model/derivation and a legacy reader. The companion now names the current v2 exact-set binding and explicit cutover rejection. No owner adoption or security/licensing authorization is inferred from these edits.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make test-fast` before compatibility-fixture corrections | Failed 725/736, `20261003T065957Z-p96820`. |
| `make test-fast` after the first fixture corrections | Failed 734/736, `20261003T071755Z-p5303`; the two remaining failures were an obsolete incident-bundle rejection vector and the administrative audit filter model. Both are corrected. A current broad rerun remains required. |
| Incidents source-codec row through `make test-slice` | Passed 1/1, `20261003T072338Z-p87149`. |
| `make test-slice OWNER=web.application ROWS=web.application.regression.administrative_audit_exact_query` | Passed 2/2, `20261003T072449Z-p88292`. |
| Reference Data verifier row through `make test-slice` | Passed 1/1, `20261003T073639Z-p4452`, including closed-envelope mutations; subsequent metadata/linkage follow-ups are included in the persistence passes below. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3, `20261003T073852Z-p11245` (envelopes and historical validation), then 3/3, `20261003T074844Z-p38917` (historical retention, catalog export and pin/removal guard). |
| `make generate` | Passed `20261003T080230Z-p6754`, including migration 50 and its owner assignment. |
| `make service-backed-test-slice OWNER=module.reference_data` | Failed 12/13, `20261003T075849Z-p70250`: the initial migration-50 projection did not permit its new publication field through the existing immutable-preparation boundary. The narrow grant/guard correction is authored and regenerated. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Passed 3/3, `20261003T080426Z-p15191`, after that correction; includes mixed refresh, first-invalidation replay and terminal-evidence mutation rejection. |
| `make format` | Passed 2/2, `20261003T080413Z-p10334`. |

Earlier change-related failures are retained: generation at `20261003T071353Z-p93561` required the pending OpenAPI fingerprint update; `20261003T073310Z-p92187` required test-selector ordering; operator evidence at `20261003T075719Z-p68660` required the new migration-catalog golden. The current golden is independently calculated from the exact emitted JSON, the appended SQL bytes and regenerated manifest; a fresh test pass remains required. A read-only `explain-target` output piped into an early-closing pager produced EPIPE; it did not run product checks or change product state.

No real migration/reset, data conversion, commit or push occurred. Full live portability, collection/backup serialization, local operator import, complete diagnostic execution/enumeration, remaining concurrency/crash evidence and exact companion adoption remain open. Focused security authority and licensing disposition have not been obtained. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.

### Closed extension members and broad fast verification

The canonical manifest and every profile object now require `extensions={}` and reject arbitrary namespaced members. This corrects a remaining draft contradiction with the approved closed-schema decision. Only the explicitly permitted TUF `signed` extension exception remains; a mutation test proves that these bytes participate in signature verification. Existing canonical fixtures and their signatures/digests remain unchanged.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make agent-finalize` | Passed 1/1, `20261003T080834Z-p35451`; retained-run maintenance skipped because `RESULTS_DIR` is unset. |
| `make test-fast` | Passed 736/736, `20261003T080900Z-p39874`, including the corrected migration-50 golden. This run precedes the extension-schema correction. |
| `make generate` | Passed `20261003T081822Z-p24580`. |
| `make format` | Passed 2/2, `20261003T081839Z-p27520`. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_trust` | Passed 1/1 execution unit, `20261003T081853Z-p32210`, covering all three selected rows. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T081918Z-p32718`, superseding the earlier migration-50 publication failure. |

Full-plan completion remains unclaimed. The mandatory production and adoption gaps listed above remain open.

### Retained provenance and broad integration repairs

This checkpoint covers work through 2026-10-03 09:08 UTC. The full remediation remains incomplete.

- Added the closed twenty-member provenance projection and strict canonical admission for retained set anchors. Publication, consumer reads, transactional pin capture and historical restore now share that boundary. Validation binds every anchor to the exact ordered member, profile authority, verification method and successful-time expiry; operator signer sets are nonempty, unique and sorted, while built-in expiry and signers retain their distinct rules. Nested omission, null, type, unknown-member and identity mutations run through independently signed fixtures. Portable source evidence still cannot confer destination trust.
- Fixed transaction-backed Indicator import to finish and close its row stream before consulting the Reference Data evaluator. This removes a same-connection query conflict during incident import. Rollback compares canonical identity without substituting the creation default for an explicitly null retained defanged presentation. Obsolete underscore-domain fixtures were rebuilt rather than accepted through a compatibility canonicalizer.
- Required Base loss now maps to the closed consumer/reporting unavailable error even if its consequence transaction cannot find a healthy fallback. Durable health publication and readiness consequences when Base itself is lost remain open; this error correction does not close that obligation.
- Disposable application restarts now preserve the Reference Pack storage root along with their database. Tests still create independent credential roots. Bootstrap audit assertions select their own action because Base/trust reconciliation legitimately emits additional audit records.
- Completed migration 46's explicit foreign-key names, supporting indexes, composite-type ACLs and append-only table privilege projections. The finalizer's one-way invalidation field has a separate narrow column grant. Migrations 46–50 remain the appended, undeployed cutover sequence; migrations 1–45 were not rewritten. Reversible-catalog tests now cover the pre-cutover sequence through 45, and independent tests prove attempted downgrade of 46, 47 and 48 fails without changing schema or version state. No real deployment was migrated or reset.
- The independently calculated migration source digest is `cc3746b50b09108486aca979383a1b7b41fb9680c89504878e3475bad4964fac`. Migration projections and exact CLI evidence fixtures were regenerated/checked against the authored SQL bytes. Removed unused legacy helpers identified by static analysis.
- Updated the exact consumer import allowlists after reviewing each boundary, preserving owner table and storage confinement. Added Reference Pack duration inventory to the telemetry checker and aligned the two new public error classifications with the telemetry owner and projection.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make agent-finalize` | Passed 1/1, `20261003T082117Z-p68261`; `RESULTS_DIR` unset. |
| `make check` before the integration repairs | Failed 973/988, `20261003T082129Z-p70212`. Focused replacement evidence follows; a fresh broad run is still required. |
| Database Migrations recurrence and preflight rows through `make test-slice` | Passed 3/3, `20261003T084703Z-p36584`. |
| Incident Bundles affected round-trip/atomicity rows through `make test-slice` | Passed 3/3, `20261003T084811Z-p56439`. |
| Reporting pinned-content row through `make test-slice` | Passed 3/3, `20261003T085013Z-p96716`. |
| Imports crash-recovery row through `make test-slice` | Passed 3/3, `20261003T085014Z-p96938`. |
| Revisions affected rollback/restart rows through `make test-slice` | Passed 3/3, `20261003T085251Z-p40397`. |
| Application/serverprocess six affected startup and restart rows through `make test-slice` | Passed 9/9, `20261003T085250Z-p40169`. |
| `make test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.purpose_identity_roles_and_acl` | Passed 3/3, `20261003T090200Z-p42991`. |
| Reference Data verifier/provenance row through `make test-slice` | Passed 1/1, `20261003T090159Z-p42763`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T090249Z-p81582`. |
| Evidence attach/failure-mapping row through `make test-slice` | Passed 3/3, `20261003T090320Z-p13868`. |
| `make test-slice OWNER=app.operator ROWS=app.operator.unit.migration_evidence_cli_transport_and_redaction_906768574c` | Passed 1/1, `20261003T090320Z-p13874`. |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_network_flow_claimed` | Passed 3/3, `20261003T090355Z-p31481`; the earlier readiness timeout did not reproduce. No timeout was enlarged. |
| `make test-slice OWNER=module.records ROWS=module.records.architecture.boundary` | Passed 3/3, `20261003T090750Z-p92377`. |
| `make otel-conformance` | Passed 6/6, `20261003T090750Z-p92271`. |
| `make generate` | Passed, `20261003T090103Z-p34509`. |
| `make format` | Passed 2/2, `20261003T090146Z-p37812`. |

The broad testutil execution unit has not yet been independently re-proved. Earlier failed generation, lint, operator-golden, privilege, boundary and telemetry runs remain retained; the focused results above supersede their corrected assertions, not the complete broad gate. Full release validation, owner adoption, security review and licensing disposition remain mandatory.

Operator import has an additional explicit Core/Jobs dependency: RP-REQ-182 requires an admitted background Job, while the existing Jobs resource, persistence and route-scoped identity require a human `submitted_by_user_id`. Implementing this requires an adopted, typed local-operator attribution variant across Jobs, Extensions finalization, recovery and public Job projection. Fabricating a user UUID or introducing a synchronous second verifier would violate the approved contracts. This dependency remains open along with live snapshot/pack portability, complete diagnostic execution and enumeration, collection/backup serialization, and the remaining concurrency/crash scenarios.

### Verification precedence and relevant usage guards

This checkpoint covers work through 2026-10-03 09:44 UTC. It does not close the full plan.

- Manifest attribution no longer selects an error before trust selection. Missing or oversized hints use `bundle_shape`; unknown repositories precede TUF metadata admission. The bounded member reader now preserves each member class's declared byte-limit error instead of mapping manifest and hint limits to archive structure.
- Live TUF admission validates present descriptor types, numeric domains, digest encoding and closed nested members before signature checks. Missing role files use the metadata-schema verdict. Missing integrity fields are checked across all roles and targets before link, expiry or inventory mismatches. Root filename errors no longer preempt metadata schema or canonical-byte errors. Cartulary target repository binding is checked with metadata links; exact manifest binding remains at its later payload check. Independent signed-vector mutations and isolated at-limit byte guards cover these changes. Complete issue enumeration, bounded deduplication and public validation summaries remain open.
- Registry-usage captures now represent actual read dependencies. Normal activation retains and checks usage revisions; verification and safety fallback preserve assignments and omit that dependency. Integration evidence proves a selected-key assignment change does not abort refresh, while a captured activation usage change still rejects as stale. Other key, trust, configuration and set guards remain in force.
- Added focused routing for the HTTP test harness's restart scenario. Its repeated runtimes now retain pack storage at the parent test's lifetime while reusing their database.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make agent-finalize` | Passed 1/1, `20261003T091702Z-p97020`; retained-run maintenance skipped because `RESULTS_DIR` is unset. |
| `make check` before this checkpoint's repairs | Failed 987/988, `20261003T091724Z-p1428`. The only failure was `raw_go:backend-integration-testutil`; its aggregate log tail omitted the original failing assertion. The repeated HTTP harness storage lifetime defect was identified by source review and repaired, with focused evidence below. A new broad pass remains required. |
| `make generate` | Passed, `20261003T093633Z-p41868`. |
| `make format` | Passed 2/2, `20261003T093644Z-p44998` and `20261003T094046Z-p71482`. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2` | Passed 2/2, `20261003T093716Z-p49820`. |
| `make service-backed-test-slice OWNER=app.server ROWS=app.server.integration.http_harness_restart_storage` | Passed 3/3, `20261003T093716Z-p49828`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T094151Z-p76852`, including relevant-usage capture and live verification changes. |
| `make lint-go` | Passed in the same command session as the Markdown check; this target emitted no retained summary location. |
| `make lint-markdown` | Passed, `20261003T094212Z-p8022`, summary `adhoc/lint-markdown/tool-run-summary.json`. |

Read-only routing discovery also confirmed that `backend-integration-testutil` is an execution unit rather than a public Make target; asking `explain-target` for that name failed with `usage_error` and ran no tests. The public owner slice above now supplies narrow evidence for its application-restart boundary.

### Retained verification order and broad gate findings

This checkpoint begins with the broad verification performed through 2026-10-03 10:11 UTC. Full remediation and release adoption remain incomplete.

- `make agent-finalize` passed 1/1 at `20261003T094549Z-p30254`; retained-run maintenance was skipped because `RESULTS_DIR` is unset.
- `make test-fast` passed 736/736 at `20261003T094610Z-p34537`.
- `make check` failed with 988/989 units passed at `20261003T095132Z-p17119`. The prior raw HTTP harness restart failure is resolved. The remaining failure was `harness.browser.boundary_support.ui_review_lifecycle`: its cancellation test expected a browser action to start within two seconds of spawning Make. It observed `ready` instead of `busy`. The test now synchronizes on the borrowed server receiving the actual navigation before measuring the unchanged stop deadline.
- `make release-check` at `20261003T100105Z-p73199` was interrupted after detecting the change-related `module.artifacts.linked_notes.note_association_resource` failure. The retained Links round-trip fixture still named bundle format 4; it now names format 5. The runner emitted `status=cancelled`, with 954/1258 units passed before interruption. This is not successful release evidence; unexecuted and canceled units remain required.

The next source revision separates retained presence, length and hash checks at their declared verification ranks. Reverify and refresh cannot select a retained hash failure before an earlier trust error. Consumer integrity checks use the same presence/length/hash order. Retained length failures now use `target_length_mismatch/member_length` and participate in atomic first-invalidation evidence. Combined-fault tests cover missing retained bytes, altered bytes, trust expiry, and the absence of indexing after rejection.

Root-update metadata now admits every required nested object and field type before canonical-byte and rotation checks. The signed-field extension exception does not permit arbitrary nested key, role or binding members. Tests mutate omissions, nulls, types and unknown members while also supplying later canonical and filename faults. Cryptographic algorithm policy, derived key identities, role-key references and threshold sufficiency remain separate trust/rotation checks.

The local operator dependency is now an explicit `RP-GATE-016` promotion condition. No Job attribution workaround, synthetic user identity, local synchronous verifier or conformance claim has been introduced.

`make format` passed 2/2 at `20261003T101218Z-p69734`; `make generate` passed at `20261003T101225Z-p74525`. The focused canonical-consumer and trust rows passed 2/2 at `20261003T101248Z-p77595`. The Reference Data service slice passed 13/13 at `20261003T101248Z-p77617`; the Artifacts Note association row passed 3/3 at `20261003T101248Z-p77634`. `make harness-ui-review-lifecycle` exited successfully but emitted no retained result location. A later broad gate must still cover its aggregate scheduling behavior.

### Closed preparation and destination catalog retention

This checkpoint records continuing implementation on 2026-10-03. All mandatory completion gates remain in force.

- Durable prepared results now use a closed canonical schema, strict manifest/envelope admission, exact complete ordered inventory and internally consistent object identities. Operator publication accepts only persisted objects owned by the current operation, or the exact captured envelope's container. It cannot manufacture an object from a descriptor. The completed index generation must belong to the same operation. Tests cover nested malformed preparations and attempted publication of absent or unrelated stored objects.
- The preparation ceiling is 71,303,168 bytes. A complete valid at-ceiling fixture is unreachable under the nested limits: the 67,108,864-byte envelope plus a 1,398,104-character padded manifest and at most 68 descriptors is below 70,608,000 bytes even allowing six-byte JSON escaping for every bounded path/reference byte and conservative framing. The isolated guard is tested at 71,303,168/71,303,169; production limits were not raised. Schema references resolve only to packaged relative leaf projections, reject external references and sibling keywords, and detect cycles.
- Migration 00051 adds immutable source reference catalogs and destination-envelope version pins. A catalog may retain absent optional versions without fabricating trusted versions. Exact local collisions fail; available members receive individual pins, and complete sets receive immutable set retention and set pins. The live Incident Bundles module now requires this owner interface for import and export. Publication shares the incident final transaction, uses ordered key guards, preserves activation/disablement, and replays committed attribution before new revision checks.
- Recovery contributions include both new tables, advancing this unpublished cutover's catalog to 142 authored tables and 111 required tables. Required-state validation includes version pins and validates catalog, result, operation and pin linkage. Core 01 and the draft Reference Pack owner describe the retained evidence and its closed typed input/result projections. This is not owner adoption.
- Native Reporting portability, new embedded-pack verification, required-artifact resolution, collection, operator attribution, complete diagnostic enumeration, security review and licensing disposition remain incomplete. The new catalog boundary must not be presented as completion of E2 or E3.

Evidence before the new portability integration:

| Command / boundary | Result |
| --- | --- |
| `make format` / `make generate` | Passed at `20261003T102147Z-p41598` / `20261003T102153Z-p46439`. |
| Canonical consumer and trust rows | Passed 2/2 at `20261003T102226Z-p49476`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13 at `20261003T102226Z-p49503`, including prepared-object ownership checks. |
| `make json-shape-check` | Passed 3/3 at `20261003T102226Z-p49431`; `make lint-go` also passed without emitting a result location. |
| `make format` / `make generate` | Passed at `20261003T102832Z-p6253` / `20261003T102839Z-p11087`. |
| Canonical consumer and trust rows, including isolated prepared bound and schema-reference confinement | Passed 2/2 at `20261003T103558Z-p15998`. An earlier invocation used abbreviated row names and failed before execution with `usage_error`; the corrected command used complete owner row IDs. |

During portability implementation, generation failed at `20261003T104914Z-p28214` and `20261003T105030Z-p30773` because dependent authored recovery binding/result fixtures still carried the prior catalog or binding digest. Those inputs were updated together; generation then passed at `20261003T105118Z-p32726`, `20261003T105356Z-p41581` and `20261003T105457Z-p49826`. Format passed 2/2 at `20261003T105450Z-p45020`. Focused portability and persistence acceptance is in progress; no pass is claimed until its result is retained.

### Portability retention acceptance and diagnostic execution

The full remediation remains incomplete. These checkpoints do not close the portability, diagnostic, security-review, or adoption gates.

The imported-catalog boundary passed its owner and HTTP tests. The exact unavailable tuples survive import and re-export without activation changes. Migration 51 now rejects older imported incidents whose source reference catalogs were never retained; category counts are reported before mutation. Portable catalogs have explicit inclusive bounds of 64 MiB, 1,024 sets and 16,384 versions. A complete independently hashed 16,801,852-byte catalog reaches both collection ceilings. A valid empty catalog with admitted surrounding JSON whitespace reaches the raw-byte ceiling; the next byte is rejected before decoding.

| Command / boundary | Result |
| --- | --- |
| `make test-slice OWNER=module.incidentbundles` | Passed 8/8, `20261003T105530Z-p52800`. |
| Reference Data canonical-persistence service row | Initial compile failure at `20261003T105651Z-p94577` used an unnecessarily broad database interface. After narrowing it to the read it needs, passed 3/3 at `20261003T105829Z-p19178`. |
| Incident Bundles HTTP round-trip service row | Initial run `20261003T105531Z-p53024` failed before product execution because concurrent warm-image setup shared one temporary stamp name. The readiness helper now uses an owned `mktemp` file and atomic replacement. The row passed 3/3 at `20261003T105932Z-p38917`. |
| `make format` / `make generate` | Passed at `20261003T110800Z-p61341` / `20261003T110807Z-p66220`. |
| Canonical manifest unit row, including complete portable bounds | Passed 1/1, `20261003T110845Z-p69292`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T110851Z-p70572`, including catalog/pin linkage and migration preflight. |
| `make test-slice OWNER=module.recovery` | Passed 24/24, `20261003T110846Z-p69521`. |
| `make lint-shell` | Passed 4/4, `20261003T111021Z-p52379`. |
| Operator migration-evidence transport row | Failed `20261003T111015Z-p51863` on the old source-catalog digest. Exact SQL/manifest hashes, ledger, binding and findings were independently checked before updating the golden; passed 1/1 at `20261003T112653Z-p67226`. This pass predates migration 52. |

The diagnostic collector no longer retains an unbounded issue map or receives a complete finding slice. It streams fixed 4,096-digest chunks into confined runs, merges with two readers, deletes obsolete runs immediately and retains only the first 1,000 issue objects. Safe path/member names and detail tokens are checked against typed projections. Unknown input names use the fixed sentinel. Tests include 40,002 unique findings with repeats across spill runs, reverse traversal, hostile fields, caller mutation, cancellation and scratch faults. Initial collector/trust selection passed 1/1 at `20261003T112653Z-p67220` after format/generation at `20261003T112618Z-p59524` / `20261003T112626Z-p64135`.

Live content-schema verification now uses the same schema evaluator as admission and enumerates every safely bounded row. Migration 52 retains canonical failed summaries and their content-derived references. Committed Jobs and attestations bind those references; the new authenticated read operation exposes the closed object. Private preparation and operationally aborted attempts remain unreadable. Historical validation checks issue identity, check mapping, ordering, counts and attestation linkage. Missing old findings cannot be reconstructed: both read-only and transactional cutover guards reject incompatible failed-attempt history.

The 16 MiB retained-summary ceiling is conservative and unreachable for a valid 1,000-issue object: admitted paths need at most twice their 4,096-byte bound after escaping, admitted IDs at most twice 512 bytes, and fixed registry tokens, bounded related identities and framing keep each complete issue below 12 KiB. This bound does not authorize larger issue counts. Remaining verification phases still produce aggregate root findings, and the full registry program is not yet the live orchestration authority. A3 remains open.

Format passed at `20261003T114317Z-p76707`. Generation at `20261003T114325Z-p81541` required four additive summary-operation/schema entries in the pending OpenAPI release; generation then passed at `20261003T114459Z-p83047`. The first owner slice at `20261003T114601Z-p91202` passed 20/23 execution units but failed compilation of the root package because the new HTTP test redeclared an existing response variable. That test is corrected; a fresh pass is pending. No release gate is claimed from these focused runs. No real migration, reset, commit, push or owner adoption occurred. `RESULTS_DIR` remains unset, so retained-run maintenance remains skipped.


### Diagnostic administration and local operator execution

The full remediation and mandatory release/adoption gates remain incomplete. The diagnostic administration UI now fetches only authenticated content-derived summary references, checks the summary against its terminal Job, and keeps diagnostic read retries separate from mutation replay. Concealment, dismissal, stale responses and authority loss fence protected results. Focus stays on the existing action control. This work used the [Cartulary UI refactor skill](/mnt/c/Users/jochi/.codex/skills/cartulary-ui-ux-refactor/SKILL.md), specifically its existing-controller, local-feedback, retry and focus rules; no redesign or dependency was introduced. Rendered browser evidence for this new panel is still required.

The collector now hashes scratch runs and verifies their lengths and record counts during bounded merge. The canonical-persistence test also checks that local import renewal preserves disablement. The initial assertion expected two audit records; the actual required admission, import and verification-completion sequence produces three. That assertion is corrected and awaits the current focused rerun.

| Command / boundary | Result |
| --- | --- |
| Reference Data full service slice | Passed 13/13, `20261003T115232Z-p80016`, including migration 52 and committed diagnostics. |
| Operator migration-evidence transport | Passed 1/1, `20261003T115232Z-p80017`; predates migration 53. |
| Diagnostic and UI focused rows | Initial run `20261003T120830Z-p37685` passed 4/5. A malformed edit to an existing controller assertion caused the fifth failure; corrected controller row passed 2/2 at `20261003T121025Z-p46416`. Scratch fault, transport and panel tests passed in the initial run. |
| Frontend typecheck / import boundary | Typecheck passed 2/2 at `20261003T121025Z-p46495` after fixing the same accidental argument edit; import boundary passed 2/2 at `20261003T120852Z-p40247`. Latest nullable Job submitter typecheck passed 2/2 at `20261003T122659Z-p44830`. |
| Jobs and Extensions focused owners | Passed 6/6 and 24/24 at `20261003T122541Z-p68336` / `20261003T122541Z-p68344`; these predate the latest durable operator attribution column. |
| Reference Data canonical persistence | `20261003T122541Z-p68352` passed 2/3; the only failure was the audit-count assertion above. Local admission, worker execution, null human submitter, proof and preserved disablement were reached successfully. |
| Format / generation | Passed at `20261003T124403Z-p55074` / `20261003T124409Z-p59905`. Earlier generation failures required the pending nullable-submitter OpenAPI entry and the identity-schema catalog metadata; historical baselines were not edited. |

Migration 53 adds nullable human submission and a durable local operator operation UUID. It refuses retired import attribution before schema mutation. The full `.v2` identity schema is closed and admits either a human actor or a local operation; only import v2 adopts it, while unchanged human-only operations retain their declared v1 identity. The same Extensions finalizer validates this identity and publishes Job outcomes. Local operations omit human Auth receipt reconciliation and retain the common proof/Job identity. Expiry may clear request material without erasing the local operation UUID.

The operator facade admits bounded incoming bytes through the same coordinator, then observes the deployment worker. It holds the existing serving lease, creates no listener and does not duplicate verification execution. Interruption stops observation, leaving the durable Job. Command, confinement, worker, cutover and expiry evidence is still being completed. RP-GATE-016 stays open until that evidence and owner adoption are recorded. No real migration, reset, commit or push occurred; `RESULTS_DIR` is unset.


The operator worker scenario now passed 3/3 at `20261003T124854Z-p15639`; its earlier run `20261003T124602Z-p87705` failed on a test-only obsolete `active` column assertion after both actual command outcomes succeeded. The assertion now reads current-set membership and `verified_available` health. Operator owner tests passed 12/12 at `20261003T124854Z-p15649`. The Jobs service run `20261003T124854Z-p15646` passed 4/5; its new expiry fixture needed an explicit UUID cast for a parameter shared with JSON text. The fixture is corrected and its focused rerun is pending. Canonical persistence and cutover preflight passed 3/3 at `20261003T124452Z-p63370`. Migration 53's source hash and migration-evidence output digest were independently reconstructed from the SQL/manifest bytes before updating goldens.

Native Reporting reference discovery is now connected to incident export through a narrow owner interface. It validates stored snapshot model identity and digest, then Reference Data validates the exact binding and unions it with the retained imported catalog. Duplicate facts coalesce, conflicting tuples fail, and unknown imported optional tuples are retained. This closes the empty-native-catalog production path once current integration evidence passes. It does not implement embedded new-pack verification, required embedded artifacts or optional snapshot payload transport; E2 remains incomplete. Focused native export/import, missing-binding, catalog-union and expired operator attribution evidence is in progress.


### Required Base loss, native reference export and schema diagnostics

The coordinated remediation remains incomplete. Definitive required-registry loss now commits its health and attestation even when no healthy Base fallback exists. The same transaction clears the current selection and advances its revision, preserves immutable historical sets, and makes the live readiness probe fail. System integrity detection remains admissible with an explicitly null frozen set; current-set consumers and new administrative operations fail through their declared errors. A healthy imported replacement remains current after inactive Base loss. Disabling the final healthy replacement is rejected without mutation. Core 01 and the Reference Pack draft now describe this boundary explicitly.

Native Reporting snapshot discovery and the retained imported catalog are now exported as one exact reference graph. The previously dormant Reporting boundary tests have explicit owner routing; their stale allowance list was reconciled with existing narrow Incidents admission and Graph source interfaces. Direct Reference Data table access is prohibited. The union test now checks both native and unavailable imported references instead of incorrectly expecting the unmerged source catalog. The real browser scenario covers rejected-pack diagnostic reading, transport failure, read retry and unchanged mutation count. The fixture declaration now exposes its admitted invalid-manifest option to TypeScript.

| Validation | Result location under `.cartulary/test-results/` |
| --- | --- |
| Real administration browser, including diagnostic retry | Passed 11/11, `20261003T130245Z-p80787`. |
| Jobs local-operator compaction | Passed 3/3, `20261003T125625Z-p3685`. |
| Native snapshot incident-bundle round trip | Passed 3/3, `20261003T125625Z-p3689`. |
| Initial Reporting ownership / persistence / frontend checks | Failed at `20261003T130245Z-p80761`, `20261003T130245Z-p80770`, `20261003T130245Z-p80936` on the stale boundary allowance, incorrect union assertion and missing fixture declaration described above. |
| Reporting boundary rerun | Failed at `20261003T131345Z-p55704` on remaining existing Graph source imports; complete corrected row passed 1/1 at `20261003T131830Z-p29135`. |
| Incident Bundles public API boundary | Passed 1/1, `20261003T131851Z-p43144`. |
| Canonical persistence and lifecycle, including absent current selection | Passed 4/4, `20261003T131345Z-p55720`. |
| Complete Reference Data unit/frontend owner slice | Passed 23/23, `20261003T131345Z-p55719`. |
| Frontend typecheck | Passed 2/2, `20261003T131515Z-p19696`. |
| Lifecycle plus healthy replacement / rejected unsafe disablement | Passed 3/3, `20261003T131830Z-p29156`. |
| Markdown / JSON shape | Passed at `20261003T131830Z-p29298` / `20261003T132017Z-p55135`. |
| Manifest enumeration, diagnostics and canonical consumers | Passed 2/2, `20261003T132743Z-p65498`; an earlier invocation named a nonexistent verifier row and failed routing before execution. The actual verifier tests belong to `canonical_consumers`. |

Manifest and bundle-hint schema diagnostics now share ordinary admission rules, enumerate known structural fields safely, and collect independently applicable cross-field findings before canonical-byte checks. TUF schema projections and complete metadata shape enumeration are being connected and need their next focused validation. All-check registry orchestration, remaining diagnostic phases, embedded new-pack portability, complete retention/collection races, licensing disposition, focused security review and owner adoption remain open. The existing Reporting canonical-object helper was also found to use ordinary JSON marshaling and unprefixed hashes despite its owner's distinct canonical/hash rules; do not treat native reference round-trip acceptance as Reporting canonicalization conformance. This needs an explicit companion-contract remediation before adoption.

No real migration, reset, commit, push or owner adoption occurred. `RESULTS_DIR` remains unset; retained-run maintenance remains skipped.


### TUF registry execution checkpoint

The full Reference Data owner slice passed 23/23 at `20261003T133606Z-p83983`, its service-backed slice passed 13/13 at `20261003T133607Z-p84213`, and JSON shape passed 3/3 at `20261003T133743Z-p58001`. These runs cover the typed TUF metadata schema additions, complete shape diagnostics and required-Base loss behavior.

The next implementation replaces the TUF/target verification sequence with one complete registry-keyed program used by both live and historical readers. It collects complete winning-check findings for canonical metadata, independent roles, supplied signatures and thresholds, rollback scopes, integrity descriptors, links, expiry and target inventory. Rejected rotation chains do not guess a predecessor for later thresholds. The live error adapter now rejects an ambiguous code mapping instead of choosing the first matching registry row. A malformed Go field tag in retained metadata was corrected without changing its intended JSON member.

Initial `make format` calls failed before execution with a harness `artifact_error`: the newly routed test name was not ASCII-sorted. `make explain-target TARGET=format DETAIL=summary` exposed the exact catalog error. Correcting authored routing restored format and generation, passing at `20261003T135114Z-p67520` and `20261003T135122Z-p72155`. Acceptance of this new execution program is pending. Remaining archive precedence, non-TUF diagnostic phases, portability, retention and adoption gaps are not closed by this checkpoint.


### Archive preflight and exact extraction

The registry-keyed TUF program passed its focused canonical trust/consumer/diagnostic selection, 2/2 at `20261003T135232Z-p75535`, and the full Reference Data service slice, 13/13 at `20261003T135259Z-p76823`. An earlier focused invocation named the nonexistent `canonical_diagnostics` row; it failed routing and ran no tests. The correct row is `canonical_pack_diagnostics`.

Archive admission now uses the registry range from format identification through member type. It performs a bounded framing/decompression scan without named writes, enumerates paths and collision participants in stable original-name order, then extracts only the exact admitted headers and bytes. Changes between scans abort operationally. Format precedes container size; later framing errors precede earlier path faults when both are safely observable. Resource guards and opaque prohibited bodies stop further interpretation as explicitly described by the owner. ZIP comments remain skipped with bounded reads in each pass; production byte/member/ratio limits were not raised. The derived header ceiling uses overflow-safe comparison.

`make format` and `make generate` passed at `20261003T140326Z-p27585` / `20261003T140333Z-p32219`. The full Reference Data owner slice passed 23/23 at `20261003T140344Z-p35319`, including real administration browser coverage, traversal permutations, complete collision findings, no writes from rejected names and changed-source rejection. A subsequent TAR boundary repair validates original name/prefix padding and header checksum before the standard reader or type dispatch; its acceptance is pending. This checkpoint does not close all A3 phases, crash cleanup, portability, retention, licensing, security review, adoption or release gates.


The final TAR name/prefix/checksum tests and canonical consumers passed 2/2 at `20261003T140945Z-p88740`; the service slice passed 13/13 at `20261003T140945Z-p88751`. JSON shape passed 3/3 at `20261003T140945Z-p88701`; Markdown passed at `20261003T140953Z-p94861`. Latest archive generation passed at `20261003T140900Z-p85590` after format at `20261003T140854Z-p80956`.

`make lint-go` first failed Go vet at `20261003T141104Z-p34224`: Network Flow's retained-Job test fixture still assigned a string to the now-nullable submitter. The fixture now supplies the human submitter pointer. The next lint run passed formatting and vet, then failed staticcheck at `20261003T141250Z-p58835` on the schema evaluator's obsolete unused `inBounds` helper; that helper is removed. The subsequent lint run passed formatting, vet and staticcheck at `20261003T141410Z-p61809`, `20261003T141414Z-p66436` and `20261003T141425Z-p76628`. The affected Network Flow unit row passed 1/1 at `20261003T141514Z-p78391`.

Base asset source inspection found only project-owned data: host and evidence contain the required `unknown` fallback entry, and indicator contains the nine declared types. Their source-profile digest is `243087a41dfefd3ad706b756f44f4cdc96e66554edb2aff568e6823b090645e2`. All three license declarations are `Apache-2.0`, redistribution `allowed`, with no LicenseRef bindings. Each packaged notice is exactly the repository Apache license plus the final LF required for a notice member; the repository file itself lacks that LF. This records the actual included material and transformation. It does not authorize redistributing any optional upstream dataset or close companion-owner adoption, independent review or release gates.


Retained logical-member checks now enumerate all missing, length-mismatched or hash-mismatched objects in their winning check, using ordinals over the immutable logical inventory. Storage failures still abort without a content verdict. The canonical-persistence service selection passed 3/3 at `20261003T141830Z-p85384` after format at `20261003T141823Z-p80679`. A dependency audit then found that live verification checks manifest policy but does not resolve retained dependency tuples or their reachable graph before indexing; activation checks do not satisfy this verification obligation. The next implementation closes this gap and its admission/publication revision guard.


### Dependency verification and admission closure

Live verification now resolves every exact retained dependency, checks its canonical member bytes, traverses reachable usable versions, rejects cycles by strongly connected component and checks candidate self/dependency conflicts before registry compatibility or indexing. Healthy disabled versions remain eligible. Missing, removed, failed or digest-mismatched tuples produce bounded identifier diagnostics; operational reads abort. Historical verification ignores later administrative health/removal changes while requiring exact retained identities and bytes. Current dependency state is protected by separate immutable admission read guards; it does not falsely assert pending mutation work. Admission rechecks the reachable key closure after ordered locking; preparation and publication reject relevant changed revisions.

Migration 00054 appends those read guards and fails before mutation on nonterminal verification inputs or preexisting dependent content lacking this evidence. The application preflight reports the same two categories. Fresh-state initialization and both rejection categories passed in disposable databases. Recovery includes the guard table (143 authored, 112 required); its catalog digest is `8ce98d4607faab2fbc75e99e132a0a2fd3c245af452effdf1215db679aa87682`. No real deployment was migrated. Core 02's mutable local-override sentence is replaced with the immutable complete-pack rule; no adoption approval is claimed.

Format passed at `20261003T143900Z-p26618`, generation at `20261003T143907Z-p31507`, focused pure manifest/dependency/consumer tests 2/2 at `20261003T143939Z-p34756`, and canonical-persistence plus live verification/activation service tests 4/4 at `20261003T143939Z-p34770`. Earlier generation failures at `20261003T143508Z-p16924` and `20261003T143700Z-p24296` caught stale recovery fixture bindings and exact catalog-count checks; both were updated from their authored inputs. The migration source digest is independently calculated as `3012414fce2cc3d206a6a8ada629dc9c78c3364b8ddaafc9031b5a18d84b8302`; the exact operator evidence stdout digest is `3023fe8ca0f90e49ec3d89c098a5f93a0852c3c09229d374c142a427225bf974`. Fresh migration-owner/operator/recovery evidence is pending. Full A–F completion, collection, portability, security review and owner adoption remain open.


### Inventory, semantic diagnostics and cleanup classification

The coordinated remediation remains **incomplete**. This checkpoint records production behavior and focused evidence; it does not close portability, physical collection, the complete verification program, security review, adoption or release gates.

The live payload inventory now reports every missing, extra, length-mismatched and hash-mismatched member. Retained-member and container findings combine only for the same winning check; no truncated summary is expanded into an invented full count. Target binding reports all independent mismatches. Strict JSON checks whole-document syntax before duplicate detection, preventing an early duplicate from hiding later malformed syntax. Manifest duplicate diagnostics enumerate declared paths and collapse undeclared subtrees to a fixed safe segment.

Semantic validators now share one field-predicate implementation between ordinary admission, index construction and diagnostics. The diagnostic pass visits every schema-admitted row, including independent profile fields, aliases, source references, external references, relationship identities/endpoints, replacement chains, license bindings, notices and registry policies. Unknown indicator types do not fabricate expected algorithms. A live combined-fault import retains all four independently specified semantic findings ahead of a later compatibility failure and creates no index. Source text and unknown field names are absent from diagnostics.

Review found a cleanup classification defect: joining cleanup errors with a content rejection left the rejection discoverable through `errors.As`. Verification now returns the operational cleanup failure without a publishable content verdict. Historical validation also closes its verification workspace before successful index publication. Fault tests cover extraction-workspace cleanup and a 5,000-finding diagnostic spill workspace; neither returns a content rejection.

| Command / boundary | Result and retained run root |
| --- | --- |
| Full Reference Data service slice after dependency cutover | Passed 13/13, `20261003T144333Z-p67876`. |
| Migration drift | Passed 5/5, `20261003T144508Z-p9916`. |
| Focused inventory/consumer unit rows | Passed 2/2, `20261003T145001Z-p23435`. |
| Canonical persistence and live verification service rows | Passed 4/4, `20261003T145010Z-p26424`. |
| Recovery catalog/parallel restore unit rows | Passed 3/3, `20261003T145001Z-p23427`. |
| Migration public surface and production catalog unit rows | Passed 1/1, `20261003T145001Z-p23445`. |
| Operator migration evidence transport | Passed 1/1, `20261003T145012Z-p27319`. |
| Strict JSON combined-fault unit row | Passed 1/1, `20261003T145817Z-p56253`. |
| Semantic/profile/manifest/consumer unit rows | Passed 2/2, `20261003T150839Z-p66062`. |
| Strict JSON plus semantic/duplicate/consumer rows | Passed 3/3, `20261003T151141Z-p73462`. |
| Live verification/activation service row | Passed 3/3, `20261003T151142Z-p73685`. |
| Cleanup faults plus profile/manifest/consumer unit rows | Passed 2/2, `20261003T151440Z-p99852`. |
| Latest format | Passed 2/2, `20261003T151424Z-p94898`. |

The recovery/migration failures at `20261003T144333Z-p67854` and `20261003T144333Z-p67860` were stale expected catalog counts/digests after migration 54; the focused passes above supersede those failures. Latest inventory generation passed at `20261003T144922Z-p20240`. No source projection input changed during the subsequent semantic/cleanup implementation, but fresh final generation/drift and broader verification remain required. No real deployment migration, destructive reset, commit, push or adoption approval occurred. `RESULTS_DIR` is unset; retained-run maintenance remains skipped.


### Private workspace restart recovery

Server readiness and local operator import now reconcile abandoned private verification workspaces under descriptor-confined cross-process directory leases. A parent lease serializes creation and collection; each live workspace retains an inode lease. Process death releases ownership without PID or age heuristics. Enumeration uses bounded batches, accepts only the private workspace filename vocabulary, rejects links and special files, and never descends into authoritative storage. Descriptor duplication uses atomic close-on-exec. This closes private scratch restart cleanup only; retained-object collection and retention-reference serialization remain open.

Format passed at `20261003T153038Z-p81165`; generation passed at `20261003T153113Z-p86270`. RootedFS passed 2/2 at `20261003T153150Z-p89428`, including independent capabilities, canceled waits, symlink rejection, streaming enumeration and actual killed-process lease release. Reference Data passed 23/23 at `20261003T153150Z-p89439` and its full service slice passed 13/13 at `20261003T153150Z-p89469`. Operator passed 12/12 at `20261003T153150Z-p89468`. Workspace recovery tests preserve live work, staged inputs and published artifacts and prove repeated sweeping is idempotent. Full A–F status remains incomplete.


### Complete live verification program

The live verifier now binds all 49 callbacks before reading source bytes and executes them through the typed ordered registry. Archive, trust, manifest and dependency owners supply complete deferred ranges; missing, nil, unknown and overlapping implementations fail operationally. No content-verdict selection remains in transport orchestration. Retained inventory summaries are combined only within their proven 67-object bound; truncated summaries are never expanded into invented counts. A source failure, cleanup failure or misplaced content verdict cannot escape as a publishable rejection.

Manifest admission now gives raw/escaped Unicode and the top-level-object requirement their correct encoding/JSON ranks. Missing and oversized manifests have explicit applicability rules; oversized content is not decoded. The current closed profiles exclude active roles before `disallowed_content`, whose data-only allowlist is still asserted. The typed registry names each applicability class, and the owner records why this particular failure rank is unreachable after earlier admission. Inert examples are never classified by guessed executable syntax.

Generation passed at `20261003T155019Z-p11588` after format at `20261003T155000Z-p6584`. The six selected verification/consumer owner rows coalesced to two passing Go units at `20261003T155038Z-p14551`. Go format/vet/staticcheck passed at `20261003T155038Z-p14828`, `20261003T155042Z-p19326` and `20261003T155058Z-p35614`. JSON shape passed 3/3 at `20261003T155121Z-p38239`; the full Reference Data service slice passed 13/13 at `20261003T155121Z-p38292`. Subsequent retained-trust-guard and misplaced-verdict tests passed with consumers and strict JSON, 2/2 at `20261003T155426Z-p79698`, after format at `20261003T155335Z-p74752`. The independent signed fixture proves exact execution of every rank; removing each callback rejects before storage access. A mistakenly requested `platform.canonicaljson` owner does not exist and ran no tests; its actual routing is the Reference Data strict-JSON row used in the passing selection. Further portability, collection, security review, owner adoption and final release evidence remain open.


### Exact portability requirements and migration 55

Incident-bundle v5 now admits `reference_pack_content.v1`: a closed original-byte-validated inventory of exact operator-container identities and required set/member pairs. Container paths derive exclusively from their admitted manifest digests. Unknown, duplicate, unreferenced or aliased paths fail before owner preparation. Optional unavailable tuples remain historical references; required unavailable tuples fail before owner state is published. Required pairs survive references-only re-export. Proven exact replay precedes fresh byte availability and revision checks, and replay cannot erase or change required-member classification.

Migration 00055 adds the retained-input constraint only after rejecting prior portability operations without this evidence. The read-only preflight reports `unclassified_portable_content`; the transactional migration repeats rejection and preserves both rows and schema head. The cutover guide documents that category. No deployed database was migrated or reset.

| Command / boundary | Result and retained run root |
| --- | --- |
| Generation after migration 55 | Passed, `20261003T161955Z-p16704`. |
| Reference Data manifest and consumer units | Passed 2/2, `20261003T162456Z-p21157`. |
| Incident Bundles unit slice | Passed 8/8, `20261003T162456Z-p21165`. |
| Reference Data full service slice | Passed 13/13, `20261003T162810Z-p82737`; includes required-member failure, exact replay without storage, re-export preservation and disposable migration rejection. |
| Incident Bundles service slice | Passed 6/6 execution units, `20261003T162810Z-p82750`. |

The initial manifest test at `20261003T161025Z-p98535` used an operator fixture where its new assertion expected a built-in. The fixture was corrected; the focused pass above supersedes it. Operator runs `20261003T162456Z-p21173` (11/12) and `20261003T162810Z-p82726` (0/1) failed the stale independent migration-evidence digest after adding migration 55. The expected schema hash is now `a6cc19b74ad44b89e959f12ccde266c63ae49755df8208a38a0be630bba9433c`; the reviewed output digest is `05cceebd83114d8d4fd865408f4bb91cfdf1eb98b07398bc1cc2073eb7ce5d49`. Its focused rerun passed 1/1 at `20261003T163511Z-p49719`. Migration drift passed 5/5 at `20261003T163511Z-p49673`.

Subsequent changes preserve staged incident-bundle input until authoritative terminal Job observation, retain it across interrupted attempts or uncertain outcomes, and add independent complete embedded-content boundary fixtures. The input-retention changes passed with the Incident Bundles unit slice (8/8 at `20261003T163511Z-p49706`); the independent complete content boundaries passed with Reference Data manifest/consumer rows (2/2 at `20261003T163511Z-p49697`). Source-side exact-container export was added afterward and is under focused verification. New embedded-container verification/publication, snapshot payload transport, physical collection, security review, owner adoption and complete release evidence remain open. `RESULTS_DIR` is unset, so retained-run maintenance remains skipped.


### Source-side embedded container export

Incident export now selects exact referenced operator containers through a narrow Reference Data stream. It checks immutable manifest/envelope/object linkage, honors the manifest's `allowed` redistribution classification, skips built-ins and absent optional container bytes, and hashes the stream before and during copying. Restricted/prohibited classifications remain references-only; the implementation does not infer legal permission from the SPDX expression. Generated content inventory names only actually copied containers and preserves retained required pairs. Incident Bundles bounds allocation before reading a container and checks complete archive bytes/member counts, including generated manifest and checksums, before ZIP construction.

Reference Data's full service slice passed 13/13 at `20261003T164346Z-p85825`; it exercises the independently signed container through the exporter, exact bytes/digests, refs-only behavior, short sinks, sink failure, missing/altered bytes and mutation during copying. Incident Bundles passed 8/8 at `20261003T164346Z-p85827`, including at-limit/over-limit allocation and complete-archive metadata accounting. Staticcheck passed at `20261003T164346Z-p86362`. Format passed 2/2 at `20261003T164324Z-p80954`. Subsequent review preserves staged outputs immediately after proven incident commit, before inspecting returned participant values, and preserves cancellation as an operational export outcome. Those final small changes require rerun.

This is source-side embedding only. Destination verification/publication of new embedded containers, source snapshot artifact/provenance transport, physical retention collection, complete fault/concurrency evidence, security review, companion adoption and final release checks remain mandatory and incomplete. The next embedded-import slice must freeze one durable cohort, use the existing finalizer for incident and pack publication, and retain inputs across Jobs recovery. Submitting independent Reference Pack Jobs from Incident Portability would violate the required atomic publication boundary and is not an acceptable substitute.


Final source-export follow-up validation passed: format `20261003T164826Z-p49631`, generation `20261003T164840Z-p54392`, Incident Bundles services 6/6 `20261003T164947Z-p57928`, JSON shape 3/3 `20261003T164947Z-p57860`, and Markdown lint `20261003T164947Z-p58143`. Go-format validation also passed. No final release/conformance claim follows from these focused checks.

### Destination cohort work in progress

Migration 00056 and the closed portable verification-context projection are being added for immutable destination preparation. The tables retain exact source selections, original reusable envelopes and per-container operation-member bindings; they do not publish content or alter legacy catalogs. Recovery contribution inventory includes these facts. This foundation is not yet connected to destination verification or the incident finalizer, and its migration/projection validation is pending. The A–F status remains incomplete.


### Destination foundation and terminal recovery checkpoint

The A–F remediation remains **incomplete**. Migration 00056 and portable-context generation passed at `20261003T170519Z-p7490`; canonical schema tests passed 1/1 at `20261003T171000Z-p11614`, canonical persistence/live coordinator services passed 3/3 at `20261003T171001Z-p11842`, and recovery-catalog validation passed 1/1 at `20261003T171002Z-p12087`. The operator digest failure at `20261003T171003Z-p12349` was the expected stale fingerprint after migration 56. Its reviewed replacement is `1bdbf96d2b0182a1751f487b857fce85404f5adbe82bb39c24b9927dda94680c`; the focused rerun passed 1/1 at `20261003T171909Z-p50899`. The new preparation tables still need explicit runtime/recovery privileges; that correction and live destination wiring remain open.

Jobs now has an injected transaction-scoped terminal-effects port. Production composition supplies Reference Data's participant, so ordinary finalization is a no-op after successful owner publication, while exhausted retries and inactive-profile cancellation close pending owner work atomically. The participant preserves established health, disablement, trust and activation, never publishes private preparation, and does not invent a verification attempt for a queued abort. Source and Jobs owner amendments describe the same boundary. Focused tests are running; startup repair of previously orphaned terminal work and full adoption evidence remain open. Generation passed at `20261003T171827Z-p47343` after format `20261003T171801Z-p42256`. `RESULTS_DIR` remains unset.


### Atomic terminal effects and startup repair evidence

Jobs terminal transitions now invoke the required owner participant before returning from the same transaction. Reference Data's participant closes exhausted or inactive-profile work, keeps never-started verification without an attempt timestamp, preserves established health and disablement, and rejects success without owner publication. Startup repair uses a narrow Jobs lifecycle projection that remains available after public expiry, preserves recoverable nonterminal work, and uses the existing mutation finalizer and exact owner receipt. It performs no verification or content/trust publication.

The new `attempt_result.v1` projection closes result shapes and count/outcome relationships. Migration 00057 appends the missing immutable-preparation runtime and recovery grants without rewriting migration 56. The schema fingerprint is `609576f1a8b7679e76b475a27193e4fa4ddbddd380b97008646d1fcec519af8d`. No real deployment migration or reset occurred.

| Validation | Result / retained root |
| --- | --- |
| Initial terminal-effect Reference Data services | Passed 3/3, `20261003T171906Z-p50432`. |
| Complete Jobs service slice | Passed 5/5, `20261003T171908Z-p50654`. |
| Generation after migration 57 and closed results | Passed, `20261003T172520Z-p8578`. |
| Closed Reference Data schema unit row | Passed 1/1, `20261003T172601Z-p11717`. |
| Reference Data persistence, live verification, terminal faults and startup repair | Passed 3/3, `20261003T172603Z-p11947`. |
| Production migration catalog/privilege/rollback recurrence | Passed 3/3, `20261003T172604Z-p12190`. |

The operator transport golden at `20261003T172605Z-p12482` failed only its stale migration-56 digest. The reviewed migration-57 output digest is `1a7f5a6eb7f7778be20631d75484bba689648edbff67d5595c69335a9f1ca7a9`; rerun is pending. The full A–F status and destination-cohort wiring remain incomplete. No security approval, companion adoption, conformance claim, or final release validation is implied. `RESULTS_DIR` is unset; retained-run maintenance is skipped.


### Retained lifecycle validation and completed checkpoint

Restore/readiness now validates closed frozen lifecycle inputs, canonical attempt results, stored outcome agreement, admitted cohort counts, complete published content results, and matching successful operation/attempt evidence. Terminal operations cannot hide unfinished attempts. Corruption tests change outcomes, canonical bytes, counts and frozen inputs inside rolled-back transactions and require rejection. Common Job success without terminal owner publication is also rejected in the live integration path.

Latest focused results:

| Command | Result / retained root |
| --- | --- |
| `make test-slice OWNER=module.reference_data` | Passed 23/23, `20261003T173505Z-p76194`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T173506Z-p76422`. |
| `make service-backed-test-slice OWNER=module.incidentbundles` | Passed 6/6, `20261003T173507Z-p76650`. |
| `make migration-drift` | Passed 5/5, `20261003T173508Z-p76889`. |
| Operator migration transport row | Passed 1/1, `20261003T173509Z-p78377`, superseding the stale golden failure above. |

Latest generation passed at `20261003T173053Z-p62444`; latest format at `20261003T173420Z-p71221`. Subsequent documentation moved attempt-result ownership into persistence §24.2 and records restore obligations. New destination verification is still unwired. Portability artifact transport, physical collection, full obligation evidence, security review, companion adoption and final release gates remain open. No approval or conformance claim has been fabricated. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset.


### Next destination integration boundary

The single live coordinator preparation method is now extracted as `prepareVerificationMember`; ordinary lifecycle execution still calls it. The extraction itself is awaiting focused rerun. Destination integration must use that same ordered verification/index/artifact path and publish within the parent Incident Portability transaction.

Before wiring the cohort, resolve three concrete boundary details in its contract and persistence projection: an embedded source-reference tuple alone cannot justify a candidate before safe manifest attribution; several newly embedded versions of one key must not acquire traversal-dependent release-sequence rollback outcomes; and redistribution denial must remain a transport eligibility decision rather than a fabricated cryptographic verdict. The current preparation-to-operation-member foreign key assumes every freshly verified container is attributable. Either explicitly close unattributable transport admission or correct that private model before accepting such input; do not create a false candidate to satisfy the foreign key. Required/optional classification remains exact and failures cannot silently publish partial trust.

Shared deadline/cancellation observation must continue through the remaining source preparation and common finalizer. Per-container independent Jobs, separate commits, fresh current-set resolution, retry rebasing and nil required checks remain prohibited. The parent Job terminal-effects participant is now available to close an aborted portable cohort without promoting its private preparation. These notes record remaining implementation work, not completed behavior or adopted evidence.

### Destination cohort production integration checkpoint

The A–F remediation remains **incomplete**. The destination path now connects admitted embedded containers to the production verifier and the parent Incident Portability transaction. This supersedes the unwired-foundation notes above, but does not close complete portability, recovery, security review or owner adoption.

- Transport admission streams exact descriptor-bound inputs, separates redistribution/unattributable denial from content verdicts, rejects contradictory source identity and prioritizes operational cleanup faults. Missing optional inventory members are not opened. The independent `portable-input.v1.json` fixture includes allowed and restricted transport variants.
- Durable preparation captures original local selections, retained input objects, canonical source catalog, dependency schedule, repository/key revisions and configuration. The parent Jobs lease authorizes admission and replay; retries do not reopen source containers or rebase onto later local versions. Pending work rejects before clock eligibility. Rejected admission removes only its uncommitted private objects; uncertain preparation commit retains potentially referenced bytes.
- The shared live verifier/index/artifact path verifies every new cohort member. Only fully verified predecessors may resolve fresh dependencies. Restarted attempts preserve interrupted history and use a new freshness instant. Publication merges compatible roots first, orders successful versions by numeric release sequence and retains destination envelopes, attestations, exact sets and pins within the incident transaction. Active selection and administrative disablement remain unchanged.
- The parent execution scope carries one budget and cancellation observer through source preparation and finalization. Its abort classification uses the adopted Extensions policy and finalizer, preserves recoverable lease/process interruptions and records typed operational outcomes. Restore now validates preparation identity, selection and object-reference consistency and terminal-history linkage.

The new focused tests cover transport contradictions and I/O faults, stable dependency ordering including cycles, expired execution budget and cancellation/deadline equality, durable replay, conflicting work, stale publication, interrupted-attempt freshness, real signed destination verification, transaction rollback, exact replay and unchanged activation. The service scenario now calls the public Reference Data `PrepareImport` boundary for the final successful import. It is not yet a complete HTTP embedded-pack acceptance matrix.

| Command / evidence | Result |
|---|---|
| `make generate` | Passed at `20261003T181105Z-p72637`, including the authored portable-input fixture. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_pack_manifest` | Passed 2/2 at `20261003T184243Z-p8261`. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3 at `20261003T184244Z-p8491`. |
| `make service-backed-test-slice OWNER=module.incidentbundles` | Passed 6/6 at `20261003T184245Z-p8732`. |

During this slice, the initially mistyped test-row selector executed no tests. Subsequent failures were a test constant spelling, a missing reloaded input logical path, a fixture bootstrap-field spelling and a fixture hash that incorrectly included the transport newline. The corrected rows above pass. No production limit was increased and no historical data was rewritten. Migration head remains 57.

Remaining destination evidence includes HTTP embedded import under destination trust, missing/invalid required versus optional bytes, multiple fresh versions and dependent cohorts, conflicting roots, source-publication/commit faults, timeout/cancellation through the common finalizer and corrupt preparation recovery. Complete snapshot/artifact transport, authoritative-byte collection, obligation-level release evidence, independent security-review disposition and companion-owner adoption remain open. Prior broad check/release evidence is stale; no new release pass or adoption claim is made. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset.


### Destination rejection and recovery evidence

Required embedded content rejection now commits attributable failed-attempt and candidate findings with the failed parent Job, without publishing successful sibling content, trust proposals, a catalog or the incident. Operational aborts remain separate. Rejection publication retains the original monotonic budget and frozen revision guards; proven cancellation, expiry or stale state reclassifies uncommitted work without re-verifying or rebasing. Restore admits a catalog-free content rejection only when it exactly matches a completed rejected attempt.

| Command / evidence | Result |
|---|---|
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers` | Passed 1/1, `20261003T191629Z-p2666`. |
| `make service-backed-test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | Passed 3/3, `20261003T191632Z-p3173`. Includes live HTTP untrusted clock, missing required input, required/optional invalid signature, valid destination verification, unchanged activation, exact reuse and re-export. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3, `20261003T191849Z-p56439`. Includes catalog-free rejection recovery and deliberate corrupted-history rejection. |
| `make lint-go-staticcheck` | Passed, `20261003T191633Z-p3680`. |

The preceding Reference Data service run `20261003T191631Z-p2890` failed because its intentional corruption fixture was stopped by the immutable-operation trigger. The corrected test disables that trigger only inside its disposable rolled-back corruption transaction, leaving the production guard unchanged. An earlier format run `20261003T191251Z-p87057` caught a missing source newline; the subsequent formats passed.

HTTP test setup failures preceding this checkpoint were an isolated-server fixture sharing its parent's object-store prefix (the run `20261003T185345Z-p72234` was canceled) and a synthetic bootstrap below the adopted root threshold (`20261003T185924Z-p6611`). The fixture now uses a distinct prefix and three root keys with threshold two; the passing HTTP rows above supersede both. The renewal test producer is not counted as independent cryptographic-vector evidence.

### Legacy schema retirement in progress

Appended migration 00058 removes only the four empty retired mutable stores after locked row-count preflight. It does not rewrite migration 26 or convert trust/history. Read-only cutover reporting now includes legacy activation and audit rows even on an otherwise current schema. The current recovery projection removes these four stores, with 141 authored tables and 110 required authoritative tables. No real deployment migration or reset has occurred. Generation and focused retirement/preflight validation are pending.

The full remediation remains **incomplete**. Historical snapshot/artifact transport, authoritative-byte collection, the remaining cohort/finalization/race evidence, focused security review and companion-owner adoption remain mandatory. Earlier broad check/release runs remain stale. Retained-run maintenance is skipped because `RESULTS_DIR` is unset.


### Legacy retirement validation checkpoint

Migration 00058 and recovery generation now pass initialization and current-schema drift checks. The safe preflight scenarios cover each retired table, preserve category counts, leave all four tables present on rejection and keep migration head 57. Source code uses the canonical stores only; the retired stores occur solely in historical migrations and explicit preflight/retirement evidence.

| Command | Result / retained root |
|---|---|
| `make generate` | Passed, `20261003T192643Z-p85550`. |
| `make migration-drift` | Passed 5/5, `20261003T192715Z-p88889`. |
| `make service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v2_recurrence` | Passed 3/3, `20261003T192715Z-p88933`. |
| Recovery frozen-catalog unit row | Passed 1/1, `20261003T192715Z-p88930`. |
| Reference Data canonical-persistence service row | Passed 3/3, `20261003T193103Z-p63578`. |
| Operator migration-evidence transport row | Passed 1/1, `20261003T193103Z-p63583`. |
| `make json-shape-check` | Passed 3/3, `20261003T193103Z-p63501`. |
| `make lint-markdown` | Passed, `20261003T193103Z-p63847`, before this checkpoint addition. |

Generation initially failed at `20261003T192513Z-p83456` because dependent Graph restore fixtures still named the old recovery catalog binding. The authored fixtures were updated together. The first retirement test run `20261003T192715Z-p88924` passed 12/13 units; the failed new fixture attempted a restricted replication setting. It now constructs valid FK parents with normal migration-role privileges. The stale operator digest at `20261003T192715Z-p88946` was updated to the reviewed schema-58 output and passed above. The broader migration unit slice `20261003T193103Z-p63636` passed 8/9 and exposed a separate stale schema fingerprint in pgtest; that expectation is corrected. The complete migration owner rerun passed 9/9 at `20261003T193356Z-p10162`.

Current schema SHA-256 is `df58d0d882dcfd7cbf12b1b4d97851e3908f2d6836922c1ab704080e91deb2cb`. Recovery semantic catalog digest is `e0b2195e6dff2bb482a058c60d8088faf9400b4627fbb558a42939ebe6c005d3`; canonical artifact digest is `1b00cd1c3e2347e961d96794d323023ba6c7e8e247dc7b03db43e6107809029b`. The Graph restore binding digest is `38798b2c3711d452e28cfc3a3731036df419d92e1a0e2c30d1f39f6a53f47347`, and operator transport output digest is `d2b667e4af077b7470e5d9238bb179a3dba8cd1834ae9ba3d83cb89103984358`.

The recovery catalog schema now uses the existing format bound of 1–4096 tables; the current generation still fixes exact counts at 141 authored and 110 required tables. The prior 113–114 schema range was stale. No production table limit was raised. Legacy schema retirement does not close physical byte collection, historical artifact transport, independent security review, owner adoption or final release validation. Overall status remains incomplete.


### Historical artifact transport implementation in progress

Reporting now owns a closed optional artifact catalog, strict byte/identity/binding admission, explicit sensitive-map omissions, and transactional immutable source-evidence retention. Incident Bundles delegates through a narrow artifact port and does not read Reporting tables. Migration 00059 adds the retained files; the recovery catalog becomes 142 authored and 111 required tables. Startup validates restored source artifacts through the Reporting and Reference Data boundaries. These changes are awaiting focused evidence; the A–F remediation remains incomplete. No security approval or owner adoption is implied.

The first generation attempt at `20261003T195518Z-p41792` rejected migration 59's missing explicit foreign-key actions, constraint name and routine search path. Those declarations were corrected before rerunning generation. No deployment database was changed.


### Historical artifact transport first evidence

The optional snapshot path now exports native immutable models and completed render members, preserves source provenance, retains imported bytes in the incident transaction and re-exports them unchanged. It excludes sensitive reveal maps with exact omission records. The first live round trip also verifies no destination native snapshots, releases or approvals are fabricated, no active set changes, and bad hashes, undeclared files, altered models and supplied sensitive files publish no partial incident state. A subsequent recovery-corruption scenario is added and awaiting rerun.

| Command | Result / retained root |
|---|---|
| `make generate` | Passed, `20261003T200040Z-p51721`. |
| `make test-slice OWNER=module.reporting` | Passed 7/7, `20261003T200244Z-p60465`, including nested catalog closure and isolated cardinality/path guards. |
| Migration production-DDL recurrence service row | Passed 3/3, `20261003T200244Z-p60493`. |
| Incident Bundles artifact and embedded-pack round-trip service row | Passed 3/3, `20261003T200534Z-p37979`. |
| Recovery frozen-catalog unit row | Passed 1/1, `20261003T200534Z-p38001`. |
| `make lint-go-staticcheck` | Passed, `20261003T200245Z-p61314`, before the final recovery-test edits. |

The first HTTP fixture at `20261003T200244Z-p60489` incorrectly tried to patch the immutable incident title; it now changes the mutable phase and proves the snapshot still retains the earlier phase. JSON-shape checks at `20261003T200244Z-p60402` and `20261003T200534Z-p37887` found the new Reporting family missing from the fixed family registry and then its generated topology fingerprint awaiting regeneration. The Incident Bundles unit slice at `20261003T200534Z-p37986` passed 7/8; its export-surface guard exposed the earlier cancellation-finalization type missing from the allowlist. The operator golden at `20261003T200534Z-p38018` needs the reviewed migration-59 fingerprint. These projections are corrected; focused reruns are pending.

Migration head is 59. Schema SHA-256 is `17571ef18d83b6a9d163b7a54f7a69bc75fb4f03912e079e67d6f644a652b176`. Recovery catalog digest is `ff769efe796c3ac9194cbe6e531e82084d39d3d48b181d63b43f07840f40264b`; canonical catalog artifact digest is `7a8df5eebeaf7e1cdbcd826bd9d2e8183fa2fc5e035b81b2d63dbfa9667d2318`. The Graph restore binding is `01a6b736bc33c211fb246bdf64fabc19b115cb431c0cc3f2c6902ee6f889a314`; operator transport digest is `b92d01e607131301f04916a933c271503bd824e5e3c6ef84fb8fdaee692a5727`.

The full remediation remains incomplete. Physical authoritative-byte collection, the remaining multi-member/cohort/finalization/race evidence, Reporting canonicalization discrepancies, obligation-level completeness, focused security review, companion adoption and fresh final release gates remain open. Retained-run maintenance is skipped because `RESULTS_DIR` is unset. No deployment migration or reset has occurred.


## Historical artifact transport — completed focused reruns

All result roots below are under `.cartulary/test-results/`. These are focused implementation evidence, not adoption or release approval.

- Incident Bundle artifact and pack HTTP round trips: `20261003T201149Z-p97400`, 3/3 passed.
- Incident Bundle public surface: `20261003T201149Z-p97409`, 1/1 passed.
- Operator migration transport golden: `20261003T201149Z-p97432`, 1/1 passed.
- JSON shape: `20261003T201149Z-p97364`, 3/3 passed.
- Migration drift: `20261003T201149Z-p97358`, 5/5 passed.
- Reporting snapshot/render integration: `20261003T201149Z-p97501`, 3/3 passed.

Migration 59 retains source Reporting artifacts separately from native snapshots, releases and approvals. Startup and restore validate their closed catalog, exact source pack bindings, checksums and member inventory. Sensitive reveal maps remain explicitly omitted. Current schema inventory is 142 authored tables / 111 required authoritative tables. No deployment database was migrated.


## Authoritative byte collection and backup retention

Production collection now runs after terminal-Job reconciliation before readiness and retries durable removal tombstones once per minute. The storage port returns a shared publication namespace lease; the registering transaction holds a shared database retention guard. Collection acquires the exclusive filesystem namespace before its exclusive database guard, retires unreferenced objects before physical deletion, and retains bytes on an uncertain retirement commit. Backup capture takes its guard before opening the repeatable-read snapshot and keeps it through object streaming.

Removed extracted copies are reclaimed only while the version is still removed, inactive, unpinned and free of pending work. Successful envelopes retain their exact signed containers; full restore authenticates those historical bytes at the successful instant and does not recreate removed indexes. Terminal preparations may release byte ownership only with another retained owner or a recorded unavailable generation. Exact reimport publishes distinct physical identities and keeps removal history. Unknown publication leftovers are collectible only after all live publisher leases and database registrations have resolved.

Focused evidence currently includes passing scenarios for live publication exclusion, lost retirement acknowledgement retaining bytes, backup/snapshot lock order, capture-wide exclusion, unregistered orphan collection, removed extraction collection, full historical verification, queued exact reimport, subsequent collection and Base integrity. The encompassing persistence row still awaits a clean rerun after tightening terminal portable-history ownership checks; do not treat these scenarios as a complete row pass yet.

- `make generate`: passed, `20261003T203529Z-p59508`.
- `make test-slice OWNER=platform.rootedfs`: 2/2 passed, `20261003T203551Z-p62757`.
- `make test-slice OWNER=module.reference_data`: 22/23 passed, `20261003T203551Z-p62756`; failed portable replay compared a process-local publication lease. Corrected to compare durable identity.
- `make test-slice OWNER=module.recovery`: first 21/24, then 23/24 at `20261003T204140Z-p93922`. Stale catalog golden/count expectations and missing Base fixtures were corrected. The remaining operator process rows passed 8/8 in focused rerun `20261003T204711Z-p90865`.
- Incident Bundle artifact/pack HTTP round trips: 3/3 passed, `20261003T204140Z-p93946`.
- `make lint-markdown`: passed, `20261003T204140Z-p94146`.
- `make lint-go-staticcheck`: passed, `20261003T204711Z-p91394`.

No migration, reset, deployment, commit or push occurred. Generation and focused checks do not close owner adoption, licensing, security review or the final release gates. `RESULTS_DIR` remains unset, so retained-run maintenance is skipped.


The complete persistence rerun passed 3/3 at `20261003T204937Z-p57894` using `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence`. This supersedes the collection checkpoint's pending-row status. The preceding failures exposed and corrected process-local replay comparison, terminal ownership validation, and a nullable SQL projection for selections with no input object. Historical corruption rejection remains covered.


## Administrative eligibility and reachable limit evidence

Administrative resources now expose required, non-null `pending_work` and `reproducibility_pinned` booleans from the same database snapshot. Pending work is selected-key-wide. The administration model blocks known ineligible actions and displays health independently of disablement, retention blockers, and current verification separately from historical snapshot provenance. The coordinator still rechecks every mutation transactionally. Core 01 and the pending OpenAPI 2.0.0 change set carry the new fields; the historical baseline is unchanged.

The Reference Pack draft now uses Core's single activation/removal error registry, including common operation rejections and pending-before-envelope precedence. It distinguishes definitive missing bytes from operational read failure and closes timeout equality. Current record owners expose no custom Host/Evidence registry assignment, and Indicators requires exactly the nine Base tokens; §18 documents why that custom-token fallback scenario is currently unreachable without inventing a record field.

All result roots are under `.cartulary/test-results/`:

| Command | Result / retained root |
|---|---|
| `make generate` | Passed, `20261003T210312Z-p96369`. |
| `make test-slice OWNER=module.reference_data` | Passed 23/23, `20261003T210850Z-p53778`, including administration, browser and the complete streamed file boundary. |
| `make frontend-typecheck` | Passed 2/2, `20261003T210850Z-p53855`. |
| `make frontend-import-boundary-check` | Passed 2/2, `20261003T210850Z-p53863`. |
| `make openapi-compatibility-check` | Passed 4/4, `20261003T210413Z-p4932`. |
| `make json-shape-check` | Passed 3/3, `20261003T210413Z-p4915`. |
| `make lint-markdown` | Passed, `20261003T210413Z-p5120`, before this ledger update. |

Superseded failures: generation `20261003T210101Z-p90986` and compatibility `20261003T210151Z-p92427` detected the newly required release fingerprints; generation `20261003T210230Z-p93177` detected unsorted authored test selectors. Reference Data `20261003T210413Z-p4957` passed 20/23 while the public required-member assertion and browser fixtures still had the retired resource shape; frontend type checking `20261003T210413Z-p5071` identified that same fixture omission. The later passes above cover those corrections.

Reporting canonical serialization is under repair and is not covered by those earlier passes. Full remediation and adoption remain incomplete. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset.


## Reporting canonical serialization repair

Reporting now implements its own UTF-8 byte-ordered, shortest-escape, integer-only canonical JSON. It rejects duplicate keys, malformed Unicode, fractions/exponents/negative zero, unsafe integers and opaque binary producer values before hashing. This remains distinct from Reference Data's RFC 8785 canonicalizer. Independent literal vectors cover the differing Unicode order, scalar boundaries, idempotence and the schema-plus-LF hash domain. A complete model also survives JSON persistence without changing canonical bytes. The obsolete golden for Go struct declaration order was replaced with those independent byte/hash vectors and a full model persistence check.

Export models, redacted models, profile views, redaction/token/reveal manifests and render metadata now use schema-domain object digests. Physical file and Extensions output checksums remain raw SHA-256; the render participant validates and translates those separate identities explicitly. Portable source-object admission now applies Reporting's number and Unicode rules to original bytes before reconstructing JSONB values. Its catalog continues using its separately declared RFC 8785 transport serialization. The private release Job codec carries normalized request bytes as base64 and does not use the canonical-object encoder.

Incident artifact/pack HTTP round trips passed 3/3 at `20261003T212032Z-p49114`. Reporting's first reruns exposed a removed encoder import (`20261003T211537Z-p98449` and companion Incident run `20261003T211537Z-p98462`), the private binary Job codec and obsolete golden (`20261003T211728Z-p22142`), and an integration assertion that conflated object digests and file checksums (`20261003T212032Z-p49106`, 6/7). Those issues are corrected; a complete Reporting rerun is pending. Staticcheck at `20261003T212032Z-p49639` found an unused fixture manifest assignment in the new file-boundary test; that assignment is corrected.

This closes neither the separately identified generated-ID/timestamp discrepancy nor full Reporting owner conformance. Reference Pack release/adoption gates and fresh broad final evidence remain incomplete. `RESULTS_DIR` is unset.

The complete Reporting slice subsequently passed 7/7 at `20261003T212347Z-p99337`, covering the corrected source object/hash boundaries and native snapshot/release flows. Staticcheck passed in the corresponding final focused rerun. These supersede the Reporting and fixture failures above; generated identifiers/timestamps and the remaining full-plan obligations are still open.


## Broad verification and fallback attribution checkpoint

The full remediation remains incomplete. `make test-fast` completed 735/737 units in `20261003T212557Z-p24748`; stale fixtures omitted the server collection lifecycle field and required administration blocker fields. Those fixtures are corrected, with focused passes in `20261003T213314Z-p18563` (server, 1/1) and `20261003T213322Z-p19236` (protocol, 2/2).

`make agent-finalize` passed in `20261003T213420Z-p20029`. `make check` then completed 988/990 units in `20261003T213440Z-p24271`. Its failures identified missing exact importer registrations for the new owner-port wiring and a direct execute grant on Reporting's private artifact trigger. The authored import registry now lists the reviewed files; migration 00060 revokes the unnecessary grant without editing migration 00059. `release-check` did not execute because `check` failed. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. These corrections require fresh evidence.

Migration 00061 and the coordinator now persist the displaced version when safety fallback selects Base. The required nullable `fallback_from_version` administrative field is separate from health and immutable provenance. Unchanged claimed-profile reconciliation preserves it; different selection and unclaimed reconciliation clear current attribution. Preflight rejects intermediate fallback history whose current cause cannot be reconstructed safely. HTTP, restart/reconciliation, explicit replacement, schema, UI and preflight acceptance additions are pending verification.

The fixture-manifest contract required by RP-REQ-256 remains absent. Obligation-level evidence and fixture registration must be completed explicitly; the broad passing units do not close that conformance requirement. Reporting identity/timestamp reconciliation, remaining concurrency/finalization cases, security authority and companion adoption remain open.


Fallback acceptance passed in the complete Reference Data slice, 23/23 at `20261003T215904Z-p98469`, and artifact portability passed 3/3 at `20261003T215904Z-p98532`. Frontend types passed 2/2 at `20261003T215904Z-p98692`, importer boundaries passed 3/3 at `20261003T215904Z-p98721`, and migration drift passed 5/5 at `20261003T215904Z-p98441`. Generation passed at `20261003T215817Z-p95017` after recording the new field in the pending OpenAPI release.

PostgreSQL parity initially exposed missing append-only classifications for the two portable preparation tables and imported Reporting artifact files (`20261003T215904Z-p98492`); the authored generator now projects their existing intended immutable grants. The operator golden initially retained the old migration inventory (`20261003T215904Z-p98491`). After both fixes and generation (`20261003T220123Z-p91277`), PostgreSQL parity passes 3/3 at `20261003T220205Z-p94753`, operator evidence passes 1/1 at `20261003T220205Z-p94768`, and the recovery operator slice passes 8/8 at `20261003T220205Z-p94782`. Migration source digest is `5c4e74637434b98fbd36b56b9f84208dd016f107d4fe5b8e7fb6dc0f199d7d97`.

The administration resource now additionally projects exact verified dependency tuples, using explicit null before successful verification and an empty array for a verified empty list. Expandable UI details show source/profile, dates, logical digests, prior activation, missing reason, and dependencies. Keyboard browser acceptance and safe text rendering tests are added; this subsequent metadata slice requires its own verification.


### Administrative metadata, fixture registration and rendered review checkpoint

- The administrative projection now includes exact dependencies (`null` before success, an explicit array afterward) and retained Base fallback attribution. The UI separates fallback from health and expands source/profile dates, digests, prior activation, missing reason and dependencies. Core 01 and the pending OpenAPI release carry the same required fields; generated clients were refreshed.
- Reference Data passed all 23 units at `20261003T221141Z-p69963`, including the new metadata and keyboard browser assertions. The combined invocation then rejected an inapplicable `OWNER` input for the frontend target; the separate frontend type and boundary invocations passed at `20261003T221330Z-p14222` and `20261003T221330Z-p14230`.
- The private seeded UI review used the cartulary-ui-review skill. Both inspected captures showed wrapping digests, readable details and visible keyboard focus; both axe scans completed with zero violations or incomplete findings. The terminal receipt at `20261003T221321Z-p9692/ui-review/terminal.json` reports two captures, zero console errors/failed requests and complete cleanup. Private screenshots and app observations were disposed by the harness. This is supporting review evidence, not a conformance claim. The first launch rejected machine-output mode before session creation; the corrected foreground invocation succeeded.
- Added the closed harness fixture-manifest schema and the first three expectation-bound manifests (independent manifest identity, complete set, signed container). Added the manually projected 264-requirement/54-criterion range registry. The test validates explicit null/omission/type/unknown fields, sorted/disjoint effects, root-confined regular inputs, symlink rejection and expectation identity. It rejects prose paths before filesystem access. The projection consistency report is `unmapped=0 unknown_requirement=0 unknown_acceptance_criterion=0`; this does not mean all obligations have implementation evidence.
- The fixture and traceability test passed at `20261003T222525Z-p57780`. JSON-shape initially rejected the schema attachment's validation token at `20261003T222526Z-p58002`; it was corrected to the registry's required `json-shape-check` token. A rerun is required. Registration of all remaining scenario fixtures and all 27 required families remains incomplete.
- Corrected the stale RP-REQ-262 clock assumption to reference RP-REQ-075's adopted design decision, preserving Base, historical-use and empty-refresh exceptions. No time-service inference or future subsystem is needed for the operator assertion.

The security-authority compliance-posture decision has been requested from the user. No approval, licensing determination for external datasets, companion-owner adoption, or full-plan completion is inferred from this work. No deployment database was migrated or reset. `RESULTS_DIR` remains unset.


## Cohort publication and boundary checkpoint

The full `make test-fast` passed 737/737 units at `.cartulary/test-results/20261003T223007Z-p67231`. The following `make check` failed 988/990 at `.cartulary/test-results/20261003T223007Z-p67243`; `release-check` did not run. Both failures were related to this work: a test-fixture setup triggered the generated-root write heuristic, and migration 61's partial fallback FK index did not satisfy the adopted full-index policy. Temporary fixture creation now lives in a separate test helper. Appended migration 62 replaces the partial index with a full leading-column index; applied migration bytes remain unchanged.

- `make backend-module-boundary-check`: passed 3/3, `.cartulary/test-results/20261003T225332Z-p52973`.
- Reference Data `make test-slice`: passed 23/23, `.cartulary/test-results/20261003T225332Z-p52700`, including request-shape precedence and content-count reachability.
- Database Migrations service-backed `production_ddl_v2_recurrence`: passed 3/3, `.cartulary/test-results/20261003T225332Z-p52743`.
- Incident Bundles service-backed round-trip row: passed 3/3, `.cartulary/test-results/20261003T230150Z-p95227`. Two fresh historical versions share one verification instant and publish in release-sequence order even when lexical order differs. Compatible root proposals advance trust once, publish both envelopes and pins, retain four attributable events and leave activation unchanged. Individually valid conflicting root bytes roll back all semantic effects, retain terminal operational-abort evidence and replay the same Job.

The new HTTP scenario initially failed at `.cartulary/test-results/20261003T225332Z-p52719`, `.cartulary/test-results/20261003T225735Z-p43317` and `.cartulary/test-results/20261003T225955Z-p69441`. Portability incorrectly copied the container digest onto root-update attestations; only the verification event now carries it, matching ordinary lifecycle publication and the closed attestation contract. Temporary bounded diagnostic instrumentation used to isolate that failure was removed.

Generation failed at `.cartulary/test-results/20261003T225235Z-p48573` because migration 62 had not yet received an authored source-owner allocation. After that allocation, `make generate` passed at `.cartulary/test-results/20261003T225311Z-p49413`. The operator digest row failed at `.cartulary/test-results/20261003T225332Z-p52747` because its expected catalog digest predated migration 62; its expected bytes have been updated and await rerun. The schema hash is now `94f4cf02311af1bed7c0ff4f8602148a6f3e27f0c337a11ced9a8bb40f76b53a`.

Forty-eight separately registered content fixtures now provide valid, malformed-structure and semantic-error inputs for all sixteen profiles. Their expected issue preimages/digests are calculated independently with Python JSON/SHA-256, never by the production verifier; the fixture runner must compare the exact winning error and ordered issues. These additions await focused validation. They do not complete every required fixture family or owner adoption.


The 48 content-profile fixture manifests passed the canonical fixture/traceability row at `.cartulary/test-results/20261003T230642Z-p25227`. The updated operator digest row passed at `20261003T230642Z-p25234`; JSON-shape passed 3/3 at `20261003T230642Z-p25187`, migration drift 5/5 at `20261003T230642Z-p25199`, module boundaries 3/3 at `20261003T230642Z-p25367`, and Markdown lint passed at `20261003T230642Z-p25390`. Generation passed at `20261003T230610Z-p21894`. These are focused checks, not a complete release pass.

## Final-commit classification and live concurrency checkpoint

The late parent-publication fault scenario passed 3/3 at `.cartulary/test-results/20261003T231602Z-p90553`. A test-only nontransactional sequence proves that the incident, two successful envelopes, four attestations, two pins and root advancement were visible inside the parent transaction before the injected catalog failure. The entire semantic mutation rolls back; terminal attempt evidence remains attributable without condemning content. Earlier runs `20261003T231037Z-p39815` and `20261003T231324Z-p65444` exposed a missing grant on that disposable fixture sequence; only the fixture grant was corrected.

The shared Incident Bundles transaction previously classified every lost commit acknowledgement as unknown without consulting the already-written Job proof. It now registers the existing Extensions owner finalizer's exact-success classifier during terminal publication. The physical classifier is shared by success, failure, cancellation and synchronous owner mutations; proof reads use a bounded context independent of caller cancellation. Missing or unreadable proof still does not prove rollback. Proven server rollback errors preserve their original classification, and unresolved mutation still enters the fatal integrity path. Server composition preserves the admitted database port when applying instrumentation, allowing the narrow transaction decorator to exercise the real parent transaction without a production fault toggle.

- Incident Bundles HTTP cohort, lost acknowledgement and explicit commit rollback: passed 3/3 at `20261003T232611Z-p18441` and again after classifier consolidation at `20261003T232958Z-p84905`.
- Extensions Job finalization: passed 3/3 at `20261003T232627Z-p29710`; the expanded success/failure/timeout/cancellation matrix passed 3/3 at `20261003T233233Z-p30452`. Every committed outcome survives context cancellation after commit. Run `20261003T232957Z-p84677` failed compilation because the new test incorrectly named a separate timeout status; Jobs correctly exposes timeout as failed, and the fixture was corrected.
- Shared cross-owner PostgreSQL protocol: passed 3/3 at `20261003T232627Z-p29712`.
- Reference Data live lifecycle row: passed 3/3 at `20261003T233907Z-p64651`. A paused import retains root-1 inputs while a different key commits root 2; it verifies its frozen bytes, rejects stale publication, preserves activation, replays its original failed Job and accepts only a new explicit import under root 2. Concurrent activation requests both capture the prior set; exactly one commits, the other returns stale state, and a new explicit request preserves both selections. Historical pin acquisition blocks a live HTTP removal, which rejects as pinned after the pin commits.
- Module boundaries: passed 3/3 at `20261003T233907Z-p64747`. Generation passed at `20261003T233603Z-p50968`.

Routing discovery initially used the nonexistent owner `platform.extensions`; it was corrected to the actual `module.extensions` and `module.crossownertransaction` owners before verification. Full remediation, Reporting identity/timestamp closure, complete fixture/evidence registration, security-authority disposition and companion adoption remain incomplete. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.


## Full-check checkpoint and admitted Reporting identity work

Before the subsequent Reporting identity correction, `make agent-finalize` passed 1/1 at `20261003T234052Z-p85331`, `make test-fast` passed 737/737 at `20261003T234131Z-p89465`, and `make check` passed 990/990 at `20261003T234131Z-p89477`. `make release-check` was deliberately interrupted to resume required implementation; it completed 958/1259 units and reports `cancelled` at `20261003T234131Z-p89481`. This is not release evidence or a product failure. The earlier completed suites do not cover the subsequent Reporting edits. `RESULTS_DIR` remains unset.

Reporting now separates immutable snapshot materialization from admitted release/preview models. Snapshot admission allocates its real ID before constructing nested source references. Release admission freezes the exact release ID and a microsecond timestamp; canonical export, deck and token identifiers use their declared schema-domain tuples. Previews use distinct hash domains with the companion's actual attempt ID. A failed render retains its admitted release identity, and execution cannot fill a missing admission time from its current clock. Generated render artifacts retain the same identity/time. These changes require focused acceptance before being considered complete.

Appended migration 63 and Reporting's read-only preflight reject incompatible retained snapshot, job, imported-artifact and preview history without reconstructing it. This has not been run against a real deployment. Reporting 1.4.0 and Report Composition 1.3.0 are explicitly proposed amendments over their adopted bases; no owner approval is inferred. The implementing-agent security review is recorded in `docs/handoffs/reference-pack-security-review.md`; independent authority/compliance disposition remains open. Full remediation and adoption remain incomplete.


The first Reporting identity slice failed 6/7 at `20261004T001610Z-p80401`; its focused diagnostic run `20261004T001910Z-p2657` identified a three-segment test path that created no record. The corrected fixture reached the intended nested-source assertion. The next slice failed 6/7 at `20261004T002706Z-p30945` because the new migration fixture omitted required Job metadata. After correcting that setup, Reporting passed 7/7 at `20261004T003053Z-p98066`, including independent ID vectors, exact timestamp admission, immutable participant input, live snapshot/release metadata and safe migration rejection. Report Composition passed 3/3 at `20261004T002706Z-p30954`; Incident Bundle round trips passed 3/3 at `20261004T002706Z-p30976`.

The new preflight's dynamic relation spelling was rejected by the SQL ownership checker (`20261004T002706Z-p31147`). It now uses explicit owner-table queries. Module boundaries passed 3/3 at `20261004T003053Z-p98240`, JSON shape passed 3/3 at `20261004T003053Z-p98001`, and migration DDL recurrence passed 3/3 at `20261004T003053Z-p98107`. The migration source hash is `9052cea15514ace7ede6cbc4d214c585c27c3ced742595b7b9bfd5269e6176f3`. The operator fingerprint run `20261004T003053Z-p98087` failed against its prior 62-migration inventory; the expected 63-migration digest is now `9925ebfb54dc9883d3fca57c01b9c0512c4a2cadd32d4ab18cf0f3e5d810890f` and requires rerun.

A subsequent schema-version correction advances changed redaction, token, reveal-map and deck artifacts independently to v2, as required for the release/preview identity union. This correction requires fresh focused validation; it does not invalidate the preceding findings about behavior. Full release evidence, complete obligation registration, reset rehearsal, security-authority disposition and companion adoption remain open.


## Disposable reset rehearsal

The public reset procedure was rehearsed against unique Compose project `cartulary-reference-pack-reset-k6w7meiq`, bound only to `127.0.0.1:54215`. Its PostgreSQL data used a 1 GiB tmpfs, with no named database volume. The selected Compose input lived at `/tmp/cartulary-reference-pack-reset-k6w7meiq/compose.yml`; no normal development service or retained incident database was selected.

- `make db-reset CARTULARY_CLEANUP_DRY_RUN=1` with that explicit Compose/port/runtime environment passed at `20261004T003507Z-p50193`. The retained log names the isolated Compose path.
- `make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` passed first initialization at `20261004T003609Z-p50844` and a second deliberate reset of the initialized database at `20261004T003700Z-p51853`.
- Read-only inspection after the second reset confirmed migration head 63, zero snapshots and zero Reference Pack versions before application startup. This is schema/reset evidence; Base readiness remains covered by the separate startup acceptance tests.
- Cleanup with the same isolated environment and `make services-down` passed at `20261004T003723Z-p52539`; read-only inspection confirmed no project container remained. Historical backup conversion was neither attempted nor implied. The guide still requires preserving a matching historical application and backup before resetting affected disposable data.


## Reporting and canonical-admission verification

The final independent artifact-version correction passed Reporting 7/7 (`20261004T003936Z-p58677`), Incident Bundles 3/3 (`20261004T003936Z-p58688`), and the operator inventory row 1/1 (`20261004T003936Z-p58690`). The changed render artifacts use their own v2 identifiers; source snapshots have a separate model identity, release/preview identities have distinct domains, and all retained render timestamps derive from admission. Module boundaries passed 3/3 (`20261004T003937Z-p58878`), JSON shape 3/3 (`20261004T003936Z-p58631`), migration drift 5/5 (`20261004T003936Z-p58647`), and Markdown lint passed (`20261004T003937Z-p58912`). These close the specific model-identity/timestamp mismatch, not full Reporting owner adoption.

Nested profile admission now mutates every present nested object field and array item in every canonical profile row through the full content validator. Omission, disallowed null, wrong type and unknown member must reject before indexing. Explicit nullable-field expectations are independent of the compiled schemas. The expanded profile row passed 1/1 at `20261004T004731Z-p16509`.

Forty-nine additional canonical admission manifests retain independently produced archive bytes and original JSON errors. Equivalent ZIP, ustar and GZIP-ustar containers carry the same signed operator pack and pass the full verifier with exact content, signer and expiry identity. Archive attacks retain exact public errors and all ordered issue tuples; invalid input writes no extracted member. JSON vectors preserve malformed Unicode, duplicate members and noncanonical input that ordinary decoding could erase. Independent Python standard-library archive/JSON/SHA-256 construction supplies the expected bytes and digests; the Go implementation does not generate its own oracle. Both fixture and live-engine rows passed 2/2 at `20261004T005431Z-p34922`, with 100 registered fixture manifests. An earlier invocation used an invalid row selector and ran no tests; the corrected command was `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_consumers`.

Nineteen additional trust manifests passed for exact root/signer proposals, retained replay, rollback, all-role signature rejection, root thresholds, expiry equality and target substitution at `20261004T010225Z-p46051`. Eleven license/source-profile manifests passed at `20261004T010517Z-p56019`. These validate runtime provenance admission, not rights to redistribute external datasets. The overall remediation remains incomplete.


## Lifecycle, dependency and startup closure

The canonical fixture runner now has a separate narrow row, `module.reference_data.unit.canonical_fixture_manifests`, so a failed canonical expectation remains visible without being displaced by unrelated nested-schema output. Its 155-manifest checkpoint passed at `20261004T014402Z-p69789`; JSON shape passed 3/3 at `20261004T014402Z-p69751`. The inventory includes isolated 64/65 set guards, the traceability projection, registry-key/algorithm/alias failures, and missing, cyclic, conflicting and transitive dependency inputs. Dependency graphs are isolated authenticated-resolution inputs; they are not represented as independently signed containers. Expected diagnostic order follows the declared tuple, including the issue-ID tie break.

Fourteen canonical lifecycle steps now run through the real coordinator, PostgreSQL, Jobs and the shared finalizer: import, same-transaction replay, activation, active reverify, disablement, envelope renewal, refresh, cancellation, empty refresh, removal, exact reimport, reactivation, and claimed/unclaimed Base reconciliation. They assert independently calculated set identities, administrative state, temporary cleanup and declared side effects. The integration row passed 3/3 at `20261004T013408Z-p28442`. Initial fixture failures were an overstrict no-attestation expectation on unchanged startup and an omitted registry-usage test port; neither was a production authorization bypass.

A subsequent production audit found that startup could preserve an active imported version without rechecking its retained bytes. Base reconciliation now captures one coherent selection and relevant revisions, checks retained bytes and current content/runtime compatibility outside publication locks, and compares captured revisions before mutation. Definitive loss atomically publishes health, fallback/dependency consequences, attestations and the resulting set. Storage execution failures publish none of those effects. Freshness is not reapplied; healthy expired metadata remains usable at startup. Six database-backed scenarios cover healthy expiry, missing bytes, altered bytes, operational read failure, stale relevant state and unrelated revisions. The expanded persistence row passed 3/3 at `20261004T014402Z-p69801`.

Canonical owner scenarios now register snapshot/report binding, embedded destination verification and historical required-content loss. With these additions all 27 fixture families have a registered execution boundary, across 158 manifests. Registration is not proof of exhaustive coverage within a family. Snapshot admission now has a real publication-barrier test that changes activation before worker completion, checks rendering and rerendering against the admitted pin, and rejects loss of the pinned Base manifest while an imported replacement remains active. Current validation is in progress. Destination verification passed 3/3 at `20261004T015240Z-p3875`; the canonical manifest inventory passed 1/1 at `20261004T015240Z-p3849`.

Related new-test failures at `20261004T015240Z-p3863` and `20261004T015240Z-p3903` respectively selected the redaction manifest instead of the render-bundle manifest and incorrectly expected an active pointer to survive definitive required Base loss. Both expectations were corrected. The boundary checker at `20261004T015240Z-p4131` required the exact new Reporting test importer to be registered for Reference Data's test-support port; the production import boundary was not widened. Fresh validation follows below.

## Final boundary-fixture audit checkpoint

The [obligation index](reference-pack-obligation-evidence.md) now links the [per-limit evidence inventory](reference-pack-limit-evidence.md). Neither document is executable authority or an adoption record. The canonical inventory contains 163 manifests across all 27 declared families. The latest five cover complete longest Windows-event, CVE and SID identities and alias/source-ref excesses. Production Get/lookup exercises the longest identities, including a colon-bearing event provider, omitted version and explicit numeric version.

Additional compact complete recipes cover a 1048576-byte indexed NDJSON line, 64 profile-specific elements, 66 declared files, 64 source artifacts/dependencies/conflicts, exact original path/segment bounds, an exact compression ratio of 100 with a one-byte-over numerator, 64 authorized ordinary signers, two-MiB TUF files, and eight-MiB aggregate metadata and deployment bootstrap. The independent 128-signature disjoint-root vector remains the cryptographic cross-language oracle. Configured container/extracted/member limits also execute against a complete independently signed container at equality and one-over. Lookup defaults/endpoints and exact 1000/1001 diagnostic retention now have direct assertions. Unreachable canonical object ceilings use explicit guards and constraint arguments; production limits were not increased.

| Command | Result / retained run |
| --- | --- |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary` | PASS 3/3, `20261004T020443Z-p30137`; all sixteen profiles and longest identity round trips. |
| `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` | PASS 3/3, `20261004T020006Z-p78396`; admission barrier, activation change, rerender and required pinned-content loss. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | PASS 3/3, `20261004T020006Z-p78407`; startup/lifecycle/history and exact owner-fixture assertions. |
| `make backend-module-boundary-check` | PASS 3/3, `20261004T020006Z-p78578`; exact Reporting test-support exception is registered. |
| `make lint-markdown` | PASS, `20261004T020006Z-p78615`; predates this limit-index update. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_consumers` | PASS 2/2, `20261004T021226Z-p62233`; 163 fixtures and complete configured archive boundaries. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.canonical_pack_diagnostics` | PASS 1/1 grouped execution unit, `20261004T022751Z-p7856`; new complete boundary recipes and exact retention assertions. |
| `make generate` | PASS, `20261004T022619Z-p4548`; latest authored source and fixture projection generation. |
| `make format` | PASS 2/2, `20261004T022605Z-p99221`; subsequent numeric recipe correction requires no formatting change. |

Related failures are retained: `20261004T020443Z-p30128` had an incorrect expected diagnostic omitting the declared 64/65 count tokens; the independently specified expectation and its canonical issue hash were corrected. `20261004T021922Z-p82982` rejected the new test producer's one-key bootstrap because the adopted root policy requires three keys and threshold two; the producer now uses a valid 64-key root. `20261004T022352Z-p97534` failed compilation because two local recipe variables used the same declaration; the second now assigns the existing variable. A separate invocation using the nonexistent selector `module.reference_data.unit.canonical_diagnostics` was rejected before execution and corrected to the registered `canonical_pack_diagnostics` row. None of these failures is represented as passing evidence.

Owner adoption, final independent security review and the full final-tree gates remain unrecorded at this checkpoint. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset. No deployment reset or data migration was performed outside disposable test databases.


## Consumer machine-contract closure checkpoint

The five consumer operations now have authored closed request/result schemas and an operation catalog defining scalar check order. Native callers and strict original-JSON adapters use the same compiled request projection. Required nullable result arms remain explicit, and success/error are mutually exclusive. UTF-8 byte bounds and Unicode-scalar bounds are distinct owner keywords. Actual Get/lookup results are additionally checked against their selected profile in integration evidence; ordinary production reads retain typed constructors over verified immutable indexes rather than reparsing every result.

The focused live all-profile slice passed 3/3 at `20261004T025627Z-p4834`, including longest identities and all five result projections. Consumer units passed 1/1 at `20261004T025627Z-p4828`. JSON shape passed 3/3 at `20261004T025425Z-p95335`; generation passed at `20261004T025557Z-p1786`. The first unit invocation, `20261004T025425Z-p95367`, failed compilation on an unused import introduced by this refactor; that import was removed before the passing rerun. Owner boundary checks passed 3/3 at `20261004T030245Z-p30047`.

The prior broad runs were deliberately canceled after this missing projection was identified: `test-fast` at `20261004T023224Z-p14613` completed 141/737 units, `check` at `20261004T023224Z-p14660` completed 231/990, and `release-check` at `20261004T023224Z-p14683` completed 173/1259. No failure was reported before cancellation, but none of these runs is complete passing evidence. Their prerequisite `agent-finalize` passed at `20261004T023155Z-p10216` and is superseded by the next finalization.

Consumer result mutation evidence now also checks explicit nulls and array elements using independently stated nullable fields and deterministic traversal. Formatting passed at `20261004T030342Z-p30741` and generation at `20261004T030358Z-p35572`; the final consumer unit rerun passed 1/1 at `20261004T030413Z-p38440`. Fresh broad gates follow. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset. No human adoption or security approval is inferred.


## Full-suite failure diagnosis and correction checkpoint

`make agent-finalize` passed 1/1 at `20261004T030610Z-p40218`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. Markdown lint passed at `20261004T030610Z-p40332`, and whitespace validation passed before the broad runs.

- `make test-fast` passed 737/737 at `20261004T030640Z-p46384`.
- `make check` finished 989/990 at `20261004T030640Z-p46428`. The sole failed execution unit was Recovery's `selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a` row: all restore assertions passed, but force-dropping a disposable PostgreSQL database exceeded the fixture cleanup deadline. No production restore failure was observed. The exact row is being repeated without concurrent broad suites; neither the cleanup deadline nor the acceptance requirement is widened.
- `make release-check` was deliberately stopped after confirmed failures so corrections could be applied. `20261004T030640Z-p46445` retains 1139 passed, 4 failed and 116 canceled units of 1259; it is not complete release evidence. The completed failed groups were timeline range scrolling, timeline keyboard support, incident import and canonical Indicator creation.
- The isolated timeline scrolling rerun also failed at `20261004T031959Z-p98398` (9/11 execution units, including its failed group summary). The range ended 52 pixels short, compared with 30 in the full run, while the five-second wall-clock assertion expired. The production scroller intentionally caps work per animation frame. Functional acceptance now uses Playwright's controlled animation clock, preserving exact range, no-query and no-document-scroll assertions. It does not increase product speed or turn the test into a performance claim.
- The existing keyboard fixture expected obsolete retention text. It now checks the named Resume/Discard controls, activates Resume through the keyboard, verifies the exact retained draft and checks Escape/focus recovery. No production inspector behavior changed.
- The browser incident-import producer still emitted bundle 4 and an unversioned empty reference array. It now selects catalog version 5 and emits the canonical empty `reference_pack_refs.v1` object with deterministic file ordering. Server integration separately continues to reject versions 1–4 without semantic effects.
- Canonical Indicator browser and accessibility fixtures still supplied `NEW[.]EXAMPLE`/`ALPHA[.]EXAMPLE` as valid domain input. RP-REQ-123 admits ASCII domain labels; defanging is a separate output operation. Positive fixtures now supply `NEW.EXAMPLE`/`ALPHA.EXAMPLE`, and an independently stated negative domain vector retains bracketed-dot rejection. No refanging alias was introduced.

A read-only audit also identified the original JSON decoder's previously unnamed depth limit. RP-REQ-029 and the fixed typed limit now define 10000 open object/array containers, top-level container depth 1, with scalar leaves adding no depth. A complete authenticated TUF signed-extension fixture reaches equality; 10001 is rejected during original JSON admission. Byte bounds remain unchanged. Format passed at `20261004T034826Z-p77805`, generation at `20261004T034844Z-p82674`, and trust/indicator units passed 1/1 at `20261004T034914Z-p85619`. Focused browser, Recovery and final sequential full-suite evidence follow.

The candidate definition of done additionally requires recreatability and interchangeable independent implementations. The checked-in cross-language vectors establish their named byte/digest/signature boundaries; they are not evidence that a second complete subsystem implementation was recreated. That broader adoption proof remains unrecorded, alongside exact companion-owner adoption and the security-authority compliance decision. This checkpoint does not promote the draft or claim completion.


The targeted corrections passed: timeline functional scrolling and keyboard retained-draft recovery, 13/13 at `20261004T034933Z-p86329`; canonical Indicator creation and accessibility, 13/13 at `20261004T035132Z-p47611`; Recovery's exact previously failed row, 3/3 at `20261004T035031Z-p26546`; JSON shape, 3/3 at `20261004T035031Z-p26493`; frontend types, 2/2 at `20261004T035031Z-p26672`. The Recovery cleanup failure did not recur without concurrent broad suites. No cleanup deadline was changed. Whitespace validation passes. The incident-import browser row subsequently passed 11/11 at `20261004T035241Z-p87079`. Sequential final full-tree gates follow the diagnostic-limit closure below.


### Diagnostic limit contract closure

The final projection review found draft Table 26-C used names that differed from the typed limit catalog, and several streaming guards omitted numeric measurements. The draft now uses the catalog's single vocabulary; the bootstrap, source-artifact, dependency and conflict limits are also present in that catalog. Typed schema annotations attach the registered limit ID to array/object cardinality findings. Shared bounded numeric constructors preserve safe counts across archive, manifest, TUF and NDJSON guards, including the distinct 128-signature root-update limit. Count strings admit unsigned 64-bit measurements without widening signed JSON numbers or path indexes. The depth guard explicitly uses the enclosing document's original-JSON failure. Failed initial candidate retention is consistent with the outcome matrix. New focused evidence and the sequential final gates remain pending until recorded below.

The numeric diagnostic changes passed `make format` at `20261004T040533Z-p27451`, `make generate` at `20261004T040547Z-p32315`, the combined diagnostic/manifest/trust row (1/1 execution units) at `20261004T040606Z-p35294`, and `make json-shape-check` (3/3) at `20261004T040606Z-p35256`. The regression checks assert actual emitted IDs and measurements, including 64/65 ordinary signatures, 128/129 root-update signatures, manifest arrays, complete container byte admission, path/segment lengths, NDJSON line stop, and unsigned-count overflow without widening path indexes. The complete owner slice and final gates remain pending.

The full Reference Data owner run at `20261004T040652Z-p36726` passed 22/23 units, including production integration and administration browser rows. The sole failure was the independent SID alias/source-reference overflow fixture still expecting `limit_id=null`; the focused manifest runner reproduced it at `20261004T040927Z-p78618`. Both authored expectations now name their registered limits, with issue identities computed from the specification's seven-field hash domain using Python JSON/SHA-256, not copied from the verifier result. Markdown lint passed at `20261004T040652Z-p36834` before this evidence addition. A subsequent owner run validates the corrected fixture expectations.

The corrected fixture row passed 1/1 at `20261004T041051Z-p82212`; regeneration passed at `20261004T041033Z-p79220`. Fresh `make agent-finalize` passed 1/1 at `20261004T041051Z-p82202`. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. The source tree is now frozen for sequential `test-fast`, `check` and `release-check`.

For clarity, earlier checkpoints used “independent security review” too broadly. The requested focused implementation review and the security-authority compliance decision are distinct. The former is recorded in the security review; RP-GATE-010 explicitly requires the latter. This handoff does not add a separate mandatory external audit. The owner's explicit independent-implementation/recreatability proof and companion adoption gates remain as written.

### Sequential final gate evidence

`make test-fast` passed **737/737** execution units at `20261004T041155Z-p88329` in 325.388 seconds on the tree including the limit-diagnostic closure and corrected independent fixture expectations. `make check` and `make release-check` follow sequentially. No conformance or adoption claim follows from this fast-suite result.

`make check` passed **990/990** units at `20261004T041748Z-p71111` in 633.086 seconds, including the previously failing Recovery row. A subsequent read-only configuration review found the development example still claimed Reference Packs while omitting the newly required bootstrap. The example now explicitly uses `reference_pack.claimed=false`, preserving required Base startup; the guide explains explicit post-adoption opt-in with a real operator-selected bootstrap. It also removes obsolete “overlay” wording. No new capability toggle was added. This example/documentation follow-up is covered by the next focused configuration and final release runs; the two broad results above precede it.

The development-default follow-up passed platform configuration 2/2 at `20261004T042914Z-p58703` and fresh finalization 1/1 at `20261004T042914Z-p58689`. The release attempt `20261004T042948Z-p63192` exposed a browser harness coupling: it injected fixture trust settings while inheriting the development example's claim flag, producing `extension_config_without_claim` during startup. That run was deliberately canceled after diagnosis: **209 passed, 56 failed, 994 canceled** in 280.676 seconds. Most failures were fixture admission/reset cascades; the Jobs group also reported three disposable-database cleanup timeouts during the failed startup churn. The harness now explicitly claims Reference Packs alongside its existing fixture-only trust settings. The development default stays unclaimed. No production claim rule or cleanup deadline was weakened. Focused browser/Base startup evidence and a fresh release run follow.

The three Jobs cleanup cases passed 3/3 at `20261004T043802Z-p43324` without deadline changes. Generation passed at `20261004T043720Z-p40216`, shell lint 4/4 at `20261004T043802Z-p43486`, and finalization 1/1 at `20261004T043802Z-p43310`. Startup diagnostics also exposed a remaining namespace/identity bug: inactive Reference Pack settings reported `profile_id=reference_packs` rather than the registered `reference_pack`. The neutral configuration adapter now takes identity from the admitted claim registration instead of splitting the settings path. A composition-level regression asserts the exact owner identity for both rejected trust settings. These subsequent source changes require refreshed validation.

The repaired browser harness and Base startup slice passed **12/12** at `20261004T044205Z-p76698`, covering administration recovery and the disconnected/optional failure modes. The first new configuration regression at `20261004T044205Z-p76693` failed only because the fixture supplied one trust setting while asserting two findings; the actual diagnostic already reported the correct registered identity. The fixture now explicitly supplies both the clock assertion and a must-not-be-read bootstrap path. Finalization at `20261004T044205Z-p76688` passed before that test-input correction.


### Pre-production scope clarification and final visual correction

The project owner confirmed that this is pre-production and that no recreation or interchangeability reports exist or are required. The draft owner §32 now reflects that scope: independent vectors, behavioral checks, safe cutover and focused security review remain required; a second full implementation does not. No production compliance approval or formal owner adoption is inferred from that clarification. Core 00's remaining profile-major-1 table row was corrected to match the intended major-2 companion contract, with its pre-production status explicit.

Fresh finalization passed 1/1 at `20261004T044452Z-p29892`; configuration passed 2/2 at `20261004T044452Z-p29898`. `make test-fast` passed 737/737 at `20261004T044530Z-p34338` and `make check` passed 990/990 at `20261004T045139Z-p26860`. The completed release run `20261004T050233Z-p18405` passed 1257/1259 units in 2013.918 seconds. Its two failed units were the visual group and its required summary: the group had two failing screenshot scenarios and 44 passing scenarios. No backend, functional browser, stateful, accessibility, recovery or migration unit failed in that run.

Artifact review used the UI-review skill in artifact mode at `20261004T053640Z-p45892`. All 14 imported capture bundles were inspected, including expected/actual/diff Evidence framing and all 13 Reference Pack actual states; image digests were verified before viewing. Reconciliation accounts for all 255 active captures/goldens, with no orphan, missing, ambiguous or unresolved registered fixture. The Reference Pack differences follow the planned health, current-verification, immutable-provenance, details and removal UI, replacing the retired checksum-labelled verifier. The unrelated Evidence difference is a retained horizontal offset at a sticky gutter. Its anchor now establishes the declared left edge before capture rather than inheriting prior interaction scroll. No production grid behavior, tolerance, renderer, mask or viewport changed. UI-review cleanup completed successfully; artifact-only review supplied no DOM or accessibility observations.

Format and generation passed at `20261004T053942Z-p50028` and `20261004T053949Z-p54982`. Focused framing verification and the maintained transactional visual refresh follow; final release evidence remains incomplete until those results are recorded.


### Reviewed visual refresh

Accepted trigger: the existing Reference Pack golden images described retired checksum verification and omitted the implemented health/current verification, historical-provenance explanation, version details and removal controls. The already-passing functional administration scenarios establish those changes. `make browser-e2e-visual-update` passed 12/12 at `20261004T054154Z-p95826`, promoting only the thirteen Reference Pack goldens below and the maintained manifest. Each promoted PNG is byte-identical to the corresponding actual image inspected through artifact-mode UI review. The renderer, fonts, viewport, zoom, density, masks and screenshot scopes are unchanged.

Affected semantic owner row: `web.design.visual.reference_pack_administration`, scenario `scenario_034f30c7dcac`. This is an active nonregistry visual scenario, so it has no separate stable registry fixture ID; its thirteen exact capture intents are the filename stems below. Ordinary-run reconciliation accounts for them through their catalog/scenario/capture IDs. It has zero orphan, ambiguous, missing or unresolved fixture mappings. No invented fixture ID or filename-based ownership was used.

Changed files under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

- `reference-pack-catalog-linux.png`
- `reference-pack-selection-linux.png`
- `reference-pack-submitted-linux.png`
- `reference-pack-recovery-short-linux.png`
- `reference-pack-recovery-zoom-linux.png`
- `reference-pack-recovery-spacing-linux.png`
- `reference-pack-queued-linux.png`
- `reference-pack-running-linux.png`
- `reference-pack-observation-recovery-linux.png`
- `reference-pack-cancel-requested-linux.png`
- `reference-pack-canceled-linux.png`
- `reference-pack-succeeded-linux.png`
- `reference-pack-failed-linux.png`

The separate Evidence anchor repair belongs to `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`, scenario `scenario_2bff3f0ed45f`, capture intent `evidence-timeline-evidence-count`. It passed 11/11 at `20261004T054018Z-p58587` with its existing golden; no Evidence image was refreshed. The repaired helper establishes the declared left edge for the sticky-gutter anchor before capture and is also exercised by the other evidence-action scenario in the full visual group. Fresh finalization passed 1/1 at `20261004T054953Z-p37655`. Two ordinary visual comparisons remain pending, with the second scheduled inside the final release run. `RESULTS_DIR` is unset, so retained-run maintenance was skipped.

The first fresh ordinary visual run passed **12/12** at `20261004T055013Z-p41727` in 436.454 seconds. Reconciliation again passed all 255 captures, with no missing, ambiguous, orphan or unresolved mapping. The second ordinary visual run is part of the final full `make release-check`, started after this pass with the same promoted manifest.


### Functional clock-driving follow-up

The next full release attempt, `20261004T055737Z-p82919`, was deliberately canceled after the timeline auto-scroll scenario exhausted its unchanged 60-second overall test budget. The trace shows 13.002 real seconds spent advancing 2 simulated seconds and 27.055 real seconds advancing another 8 simulated seconds; the test expired while advancing the horizontal interval, before its final boundary assertion could inspect the page. Eighteen other tests in that browser group passed. This is retained as a failed/canceled run, not successful release evidence.

The functional test now advances one 16 ms frame at a time, stops immediately when each declared scroll boundary is reached, and retains a bounded maximum frame count. It no longer executes several seconds of unrelated idle frames after reaching an edge. Production auto-scroll, production deadlines, the overall test timeout, exact selected membership, focus, query-count and document-scroll assertions are unchanged. This repairs the incomplete earlier fixed-interval clock test approach; focused and full reruns are required. The accepted visual manifest is unchanged.

The canceled run totals were **437 passed, 1 failed, 821 canceled**, duration 405.639 seconds. Format/generation passed at `20261004T060502Z-p95740` and `20261004T060509Z-p1046`. The repaired functional scrolling selector passed **11/11** at `20261004T060525Z-p3967`; no assertion or timeout was removed or weakened. Finalization and the complete release gate follow sequentially.

The final editorial version review also corrected Core 00's capability prohibition to name Extensions contract major 2, matching the existing Extensions owner. Required capability arrays remain empty. The exact Core 00 and Reference Pack content revision bindings were refreshed manually; no runtime or generated artifact consumes those Markdown bytes.


### Undo-continuity fixture preparation

The following full release run, `20261004T060707Z-p45314`, passed the repaired timeline range group but found an independent fixture precondition error in the existing auto-resolution Undo scenario. The test never established its initial grid offset and always wheeled upward by 120 pixels; depending on the preceding relationship-authoring layout, that could validly clamp to zero before the test checked that it remained positive. Fourteen other scenarios in that group passed. The run was canceled after diagnosis to repair the fixture and restart validation; it is not successful release evidence.

The scenario now establishes a middle scroll position before Undo captures its owner state, verifies that the position is scrollable, then dispatches a real native wheel movement for half the observed offset while the request is held. It still requires a positive newer position, a decrease from the initial position, BODY focus, one request and exact preservation after late acceptance. Product scrolling/focus logic, the timeout and comparison tolerances are unchanged. This is a test-preparation repair; the accepted visual manifest remains unchanged.

Canceled-run totals: **1154 passed, 1 failed, 104 canceled** in 1336.966 seconds.

The combined range-scrolling and auto-resolution owner selection passed **12/12** at `20261004T063044Z-p89867`. A final setup refinement waits for two animation frames after the initial programmatic scroll, before Undo admission, so setup cannot be mistaken for post-admission user intent. Format/generation passed at `20261004T063334Z-p30649` and `20261004T063341Z-p35615`; the focused follow-up and final release run remain pending.

The final Undo fixture follow-up passed **11/11** at `20261004T063352Z-p38466` in 125.264 seconds. The production tree and accepted golden manifest are frozen for fresh finalization and the complete release rerun.

### Completed pre-production handoff — 2026-10-04

Fresh `make agent-finalize` passed at `20261004T063610Z-p76079`. The final complete `make release-check` passed **1259/1259** at `20261004T063630Z-p80139`, including both repaired Timeline browser groups and the second ordinary visual comparison. Its source and golden identities, complete counts and phase exits are recorded in the current sections above. This closes the earlier pending implementation and validation statements for the authorized pre-production scope. The prior failed/canceled runs remain diagnostic history.

Final handoff Markdown lint passed at `20261004T071627Z-p84809`; its follow-up after the ledger and owner-inventory result additions passed at `20261004T071911Z-p87560`. `git diff --check` passed. This final record adds only the successful follow-up result. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. No commit, push, real deployment reset or production conformance claim was made.

## Four-defect remediation follow-up — 2026-10-04

This is the execution record for the user's subsequently authorized **Reference Pack remediation and readiness plan**, based on `54f42195572c04b2bc9bb9003f2338eef0ba37db` on `main`. The only pre-existing uncommitted edit was this ledger; its contents above are preserved. This scoped execution supersedes the earlier ledger-only edit restriction for these four defects and their specification/evidence gates. It does not execute the broader production-readiness refactors, choose a validated cryptographic deployment, perform a production rollout, or adopt any specification.

The initial read-only inspection covered `docs/design.md`, `docs/domain.md`, the affected owners, contracts, callers, fixtures, routing and storage adapters. Domain remains vocabulary guidance and Design remains presentation guidance. Historical `make explain-run` inspections confirmed 737/737 at `20261004T041155Z-p88329`, 990/990 at `20261004T041748Z-p71111`, and 1259/1259 at `20261004T063630Z-p80139`; all identify a dirty tree based on `78a4effc2`. Equivalence to `54f4219` was not established. None closes this follow-up's regression or release gate.

### Corrections and ownership

| Gap | Correction and source owners | Validation boundary | Compatibility and unresolved risk |
| --- | --- | --- | --- |
| W1 — Composition identities | Extracted the unchanged Reporting serializer to `internal/modules/reporting/canonicaljson` (`Marshal`, `Canonicalize`). Composition imports it and owns `CompositionDigest` and `PreviewSourceDigest`; Reporting calls the preview digest facade. Raw tokens undergo strict admission before permissive decoding can erase duplicates, malformed Unicode or numeric spelling. REQ-RPT-047–049 and REQ-RC-019a/040/041 now explicitly distinguish literal UTF-8 escaping, self-member exclusion and Reporting's separate domain-prefixed object hashes. | Independent literal bytes and Python SHA-256 vectors, materialized defaults, malformed JSON/numbers, inline tampering, release-tuple resolution, actual draft/version Jobs, internal-only output and replayed identity/time/snapshot/set binding. Existing Reporting canonical vectors still pass. | Affected previously nonconforming composition/preview identities change. Valid Reporting canonical bytes remain identical. Fresh disposable state is required; no alternate hash acceptance or immutable history rewrite. |
| W2 — Framework indexing | `packformat/content.go` coalesces exact external-ID lookup keys per object in first-occurrence order. RP-REQ-109 distinguishes provenance tuple uniqueness from index-key uniqueness. Original canonical `external_refs` and database uniqueness constraints are unchanged. | ATT&CK/D3FEND/VERIS fixtures retain distinct source and URL references sharing an ID. Real signed imports preserve canonical content, create one key per object/value and page through distinct objects once. Exact duplicate references still fail unit admission. | Previously rejected valid packs import successfully. No schema, signed-byte, public lookup or digest change. |
| W3 — Early upload limit | Reference Data owns `StagingLimitError` measurements; RootStorage streams at most the bound plus one probe byte. The owner maps a sole typed rejection through the existing `container_bytes` diagnostic machinery. Core 01 and the authored OpenAPI specify 409 `reference_pack_verification_failed`, `container_bytes_exceeded`, check ID and a single inline existing validation summary. Operator admission returns the same code/reason with null Job and unestablished identities. RootedFS now preserves operational sealing/cleanup failures over callback rejection; mixed probe/read errors remain operational. RP-REQ-246/247 distinguish the early guard from complete-container verification order. | Real filesystem below/equality/one-over and first-excess consumption; MaxInt64-safe diagnostic strings; I/O/cancellation/joined-error cases; known-length and streaming multipart HTTP responses, independent issue ID and unchanged durable table counts; temporary cleanup; operator transport classification. | Erroneous HTTP 500 becomes the specified 409. No Job, candidate, attestation, persisted summary resource, trust advancement or published bytes are created by early excess. Genuine operational failures remain operational. |
| W4 — URL admission | `NormalizeURL` rejects forbidden raw URI bytes after the single permitted outer trim and before component parsing. RP-REQ-124 now states this ordering explicitly. | IPv6 brackets, host, path, query and fragment reject space, C0, DEL and non-ASCII/forbidden punctuation. Outer trim and encoded whitespace remain permitted, valid canonical vectors remain stable, public consumer outputs are null on `invalid_http_url`, and framework URLs retain HTTPS-only canonical-input rules. | Corrects the existing algorithm identifier; no new normalizer version. Formerly repaired malformed input is rejected. Stored identities are not rewritten. |
| W5 — Authority | Core 00 REQ-00-065 and residual Extensions vocabulary, startup/staleness rules, manifest limit and adoption criterion now follow Extensions §1.1: executable dependencies contain typed operational artifacts, never document manifests/hashes. Domain distinguishes required Base registries from optional imported packs. | Human comparison of owner clauses and projections; current editorial companion references in Reference Pack Table 4-C. Tests and generation do not consume Markdown. | The textual contradiction is repaired in the candidate amendment set. Coordinated human adoption remains pending; executable compatibility or document-provenance readers were not added. |

The shared serializer is deliberately separate from Reference Pack's RFC 8785 implementation. Only strict JSON admission is reused across those serializers; UTF-8 member order and finite-integer encoding remain Reporting-owned. The new serializer is a pure subpackage, with no application-package import cycle or generic configurable hashing layer.

Authored implementation/test changes are confined to `internal/modules/reportcomposition`, `internal/modules/reporting`, `internal/modules/reference_data`, `internal/app/referenceassembly`, the existing Reference Pack operator adapter/test, and RootedFS's operational-error handling/test. Tests extend existing semantic rows except `module.reportcomposition.unit.canonical_digests`; the cross-module preview integration retains Report Composition ownership with Reporting as collaborator. The exact operator test facade allowance is in `tools/backend_module_boundaries.json`; the Reporting source import guard allows only the new pure package from `canonical_json.go`.

Authored projection inputs are `contracts/openapi-source/owners/module.reference_data/openapi.json`, `contracts/openapi-releases/2.0.0.change-set.json`, and `tools/test_families/module.reportcomposition.json`. The new synchronous conflict shape is recorded in the existing pre-production 2.0.0 change set; the historical baseline is untouched. Make generation produced the assembled OpenAPI, embedded OpenAPI artifacts/operation catalog and execution-topology render index. No generated file, migration, lock file or signed built-in pack was hand-edited. The owner inventory, obligation index and limit inventory now direct readers to this scoped evidence instead of attributing earlier cutover totals to this candidate.

### Regression evidence and diagnostic history

Run IDs below are under `.cartulary/test-results/`. Counts are execution units, not claims of exhaustive requirement coverage. All commands ran from the repository root through public Make targets.

| Command | Result and attribution |
| --- | --- |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_indicators,module.reference_data.unit.canonical_pack_profiles` | Before the corresponding production fixes: FAIL 0/1, `20261004T140852Z-p12567`; both newly extended semantic rows reproduce the URL and derived-key defects. |
| `make service-backed-test-slice OWNER=module.reportcomposition` | Before the composition fix: FAIL 2/3, `20261004T140903Z-p13241`; actual Reporting worker rejects preview source binding with ordinary text. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_upload_envelope_failures_create_n_710dd1d70f` | Before the staging fix: FAIL 2/3, `20261004T141223Z-p54360`; fixed-length and streaming cases return 500 instead of 409. Earlier `20261004T141046Z-p34017` had a new-test table-name mistake and is not defect evidence. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_indicators,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_consumers,module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda` | PASS 3/3, `20261004T141838Z-p81891`. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary,module.reference_data.integration.reference_pack_upload_envelope_failures_create_n_710dd1d70f` | PASS 3/3, `20261004T142056Z-p52870`; actual signed imports, pagination and both HTTP transports. |
| `make test-slice OWNER=module.reportcomposition` | PASS 4/4, `20261004T141930Z-p89436`. Expanded canonical/default/inline and release-tuple assertions subsequently passed in 3/4-unit run `20261004T142809Z-p96059`; its preview build failed due to the new test using the wrong database helper signature. That test-only error was corrected. |
| `make service-backed-test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a` | PASS 3/3, `20261004T143032Z-p28309`; includes the expanded replay, timestamp, snapshot and pack-set assertions. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a` | PASS 1/1, `20261004T142809Z-p96062`; includes the mixed extra-byte/read-error probe. |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | PASS 1/1, `20261004T142536Z-p87960`; unchanged canonical byte vectors execute through extracted serializer. |
| `make test-slice OWNER=platform.rootedfs` | PASS 2/2, `20261004T142536Z-p87940`. Earlier `20261004T142009Z-p46243` failed a new test's same-package qualifier, corrected before the passing run. |
| `make test-slice OWNER=app.operator ROWS=app.operator.unit.operator_registry_eight_paths_and_collaboration_01217009b8` | PASS 1/1, `20261004T142536Z-p87949`. Earlier full operator slice `20261004T141932Z-p89811` and format `20261004T141916Z-p84284` failed a missing brace in the new test; corrected. |
| `make backend-module-boundary-check` | PASS 3/3, `20261004T142914Z-p20929`; earlier `20261004T142754Z-p95499` identified the missing exact operator test allowance. |
| `make format` | Latest PASS 2/2, `20261004T143007Z-p22778`. |
| `make generate` | PASS, `20261004T143031Z-p27744`. Earlier `20261004T142013Z-p50307`, `20261004T142102Z-p60277` and `20261004T142914Z-p20792` correctly rejected absent/stale authored OpenAPI change-set fingerprints. The four current changes are now explicitly recorded; no guard was weakened. |

These failed runs remain retained diagnostic history. New fixture/compiler mistakes are distinguished from the four successful baseline reproductions. The three baseline test invocations cover all four defects; production fixes were applied after the relevant failures were captured.

### Promotion and adoption gate ledger

The authority roles below identify the responsible decision maker; no individual approval or external authority action has been fabricated. “Candidate ready for review” describes an amendment or implementation, never adoption. Reference Pack remains **draft 0.2.0**, Reporting remains **proposed 1.4.0 over adopted 1.3.0**, and Report Composition remains **proposed 1.3.0 over adopted 1.2.0**. The current Extensions 0.12.0 header does not retroactively adopt this follow-up's edits.

| Gate | Responsible authority | Evidence / disposition and next action |
| --- | --- | --- |
| RP-GATE-001 | Core 00 and Extensions owners, project adoption authority | REQ-00-065 and Extensions §1.1 are now textually consistent; residual document-manifest requirements removed. Former `BLOCKED: owner contradiction` becomes **candidate repair awaiting coordinated adoption**. Review typed recognition/dependency projections and adopt together. |
| RP-GATE-002 | Core 01 owner | Required Base and optional imported administration remain separate; Domain corrections align terminology. **Companion adoption pending**. |
| RP-GATE-003 | Core 01 public-contract owner | Inline early-upload rejection now matches authored/generated OpenAPI, existing reason registry and route evidence. Removal, lifecycle and exact-set amendments remain the named companion set. **Adoption pending**. |
| RP-GATE-004 | Core 02 owner | Immutable pack/envelope/set schema and retained persistence evidence; this follow-up changes no SQL or uniqueness constraint. **Companion adoption pending**; review existing §24 mapping. |
| RP-GATE-005 | Core 04 security/deployment owner | Existing trust bootstrap, clock and timeout projections; operational/content separation strengthened. **Companion adoption pending**; deployment posture is not selected here. |
| RP-GATE-006 | Reporting and Report Composition owners | Shared serialization, companion-owned digests, real draft/version preview and release-tuple evidence above. **Coordinated 1.4.0/1.3.0 adoption pending**. |
| RP-GATE-007 | OpenTelemetry owner | Existing closed operation/attribute registry and telemetry semantic row remain unchanged; full candidate verification must still execute them. **Companion adoption pending**. |
| RP-GATE-008 | Testing Harness owner | Existing typed fixtures/routing plus one new uniquely owned canonical-digest row; Make-generated topology and fresh results. **Companion adoption pending**; routing counts alone are not evidence of completeness. |
| RP-GATE-009 | Domain vocabulary maintainer | Candidate rows now distinguish required Base registries from optional imports and Reporting templates. **Editorial review pending**; Domain creates no behavior authority. |
| RP-GATE-010 | Target-deployment security authority | Ed25519-only TUF profile remains unchanged. **Blocked for production promotion: compliance approval not supplied**. Obtain the target posture decision; a different cryptographic profile requires a separate scoped revision. |
| RP-GATE-011 | Licensing/distribution authority | Existing inventory contains original Base assets and synthetic test fixtures; this effort distributes no external dataset. **No new external-corpus approval triggered by this diff**. Authority must confirm the distribution inventory at promotion; future external redistribution remains blocked without approved classification/notices. Runtime license parsing is not legal approval. |
| RP-GATE-012 | Corpus maintainer and subsystem owner | Existing human obligation/limit indexes plus the concrete corrections and fresh regression results here. **Requirement-level review pending**: confirm every RP requirement's behavioral mapping and no normative placeholder. A test count or document binding cannot waive that review. |
| RP-GATE-013 | Extensions owner and project adoption authority | Existing profile 2/configuration 2/Jobs 2 projections unchanged; Core/Extensions authority repair follows §1.1. **Companion adoption pending** for the coordinated Reference Pack set. No document-provenance executable input is restored. |
| RP-GATE-014 | Core 01 portability owner | Existing format-5/destination verification contracts and release routes unchanged. **Companion adoption pending**; review fresh complete-gate results alongside existing mapping. |
| RP-GATE-015 | Core 01/Core 04 recovery owners | Existing current recovery contributions/historical-time verification remain unchanged. **Companion adoption pending**. No history converter, trust reconstruction or identity backfill is introduced. |
| RP-GATE-016 | Core 01 Jobs and Extensions owners | Local operator early limit error preserves owner code/reason without fabricated Job identity; admitted imports still use the same coordinator/Jobs finalizer. **Companion adoption pending**. |

No second implementation, recreation/interchangeability report, real deployment reset or production rollout is required by the authorized pre-production scope. No legacy reader, converter, dual write, alternate digest acceptance, schema-version bypass or historical identity rewrite was added. Disposable development state with defective immutable identities must be recreated through the existing development cutover procedure; retained historical backups stay paired with their historical application.

### Phase exits and final verification

Phases 0–3 have concrete baseline, owner amendments and passing boundary/identity/indexing evidence above. Phase 4 requires fresh projection checks, finalization, `test-fast` and the complete `release-check` against a recorded unchanged candidate. Phase 5 requires this actionable gate ledger and handoff; production promotion remains blocked independently of successful implementation verification. Fresh final results are appended below rather than replacing historical entries.

Final projection checks passed: `make generate-drift` 4/4 at `20261004T143354Z-p52434`; `make json-shape-check` 3/3 at `20261004T143354Z-p52457`; `make generated-artifact-policy-check` 3/3 at `20261004T143354Z-p52450`; `make openapi-compatibility-check` 4/4 at `20261004T143354Z-p52507`; `make lint-markdown` at `20261004T143354Z-p52847`. `git diff --check` passed. The subsequent changes added test assertions only: exact duplicate signed references publish no partial index, the OpenAPI immediate-response fields remain closed and reuse the existing summary, and MaxInt64 diagnostics retain exact decimal strings.

Those final assertions passed: `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary` 3/3 at `20261004T143604Z-p65643`; `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda` 1/1 at `20261004T143620Z-p80758`. The signed negative cases cover all three framework profiles. Final `make format` passed 2/2 at `20261004T143553Z-p60553`, and `make agent-finalize` passed 1/1 at `20261004T143605Z-p65857`. `RESULTS_DIR` was unset; retained-run maintenance was skipped. `make explain-target TARGET=release-check DETAIL=summary` resolved the complete candidate graph before broader execution.

The final human authority scan also removed the retired companion-manifest declarations from Core 01–04 and aligned Extensions diagnostic declarations/registry identifiers with the existing §1.1 typed projections (`v3` declarations, `v2` condition registry). The obsolete document-manifest resolver and static closure/source-linter completeness claims were replaced by operational contract resolution and human review. These are documentation repairs to the adopted §1.1 boundary; no runtime projection changed or approval was inferred. Table 4-C and the owner inventory received renewed editorial references after this review.

### Follow-up file inventory

This inventory includes the preserved pre-existing ledger edit plus the scoped append above; other listed changes belong to this execution. Generated paths are the OpenAPI assembly, its two Go projections, and the topology render index, all produced through Make. No lock, SQL migration, deployment or visual-golden file changed.

```text
contracts/openapi-releases/2.0.0.change-set.json
contracts/openapi-source/owners/module.reference_data/openapi.json
contracts/openapi/cartulary.openapi.yaml
docs/domain.md
docs/extension-subsystem-nlspec.md
docs/handoffs/reference-pack-limit-evidence.md
docs/handoffs/reference-pack-obligation-evidence.md
docs/handoffs/reference-pack-owner-inventory.md
docs/handoffs/reference-pack-remediation.md
docs/reference-pack-subsystem-nlspec.md
docs/report-composition-nlspec.md
docs/reporting-subsystem-nlspec.md
docs/spec/00_document_set_status_and_precedence.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/02_domain_model_schema_and_history.md
docs/spec/03_workbook_interaction_collaboration_and_workflows.md
docs/spec/04_security_deployment_and_conformance.md
internal/app/operator/operator_reference_pack.go
internal/app/operator/operator_reference_pack_test.go
internal/app/referenceassembly/artifact_storage.go
internal/app/referenceassembly/staging_bounds_test.go
internal/app/referenceassembly/storage_test.go
internal/gen/contractopenapi/artifacts_gen.go
internal/gen/openapioperations/catalog_gen.go
internal/modules/reference_data/api.go
internal/modules/reference_data/api_test.go
internal/modules/reference_data/consumer_test.go
internal/modules/reference_data/coordinator.go
internal/modules/reference_data/internal/packformat/content.go
internal/modules/reference_data/internal/packformat/content_test.go
internal/modules/reference_data/internal/packformat/framework_lookup_test.go
internal/modules/reference_data/internal/packformat/indicator.go
internal/modules/reference_data/internal/packformat/indicator_test.go
internal/modules/reference_data/openapi_contract_test.go
internal/modules/reference_data/profile_consumers_integration_test.go
internal/modules/reference_data/reference_pack_integration_test.go
internal/modules/reference_data/reference_pack_test_helpers_test.go
internal/modules/reference_data/routes.go
internal/modules/reference_data/staging_rejection.go
internal/modules/reference_data/staging_rejection_test.go
internal/modules/reference_data/storage_port.go
internal/modules/reference_data/upload_limit_integration_test.go
internal/modules/reportcomposition/api.go
internal/modules/reportcomposition/canonical.go
internal/modules/reportcomposition/canonical_digest_test.go
internal/modules/reportcomposition/release_tuple_facade_test.go
internal/modules/reportcomposition/routes.go
internal/modules/reporting/boundary_guard_test.go
internal/modules/reporting/canonical_json.go
internal/modules/reporting/canonicaljson/canonical.go
internal/modules/reporting/composition_preview_integration_test.go
internal/modules/reporting/store.go
internal/platform/rootedfs/root_linux.go
internal/platform/rootedfs/root_linux_test.go
tools/backend_module_boundaries.json
tools/execution_topology_render_index.json
tools/test_families/module.reportcomposition.json
```

### Full candidate validation and Timeline timeout investigation

`make test-fast` passed **738/738** at `20261004T143722Z-p91994` in 334.233 seconds, with zero failed, skipped or canceled units. `make explain-run RESULTS_DIR=.cartulary/test-results/20261004T143722Z-p91994` independently reports the same result and 656 cleanup steps with no failed or blocked cleanup. Its source is dirty `54f42195572c04b2bc9bb9003f2338eef0ba37db`, source fingerprint `sha256:5c9373a4378d601d416e9a2371c7275a784640d4ff1560eef8467312432c73fb`, toolchain fingerprint `sha256:461a0521f7b3649ff41783a87f34a739b043899881ee43995da8642c47879992`, graph `sha256:80d3622fb85f6a7098ac247e0c3841ec4313d996c2ba4f3f4947bdff33106e10`. Normal harness cache policy was retained.

The first complete `make release-check` finished **FAIL, 1258/1260**, at `20261004T144412Z-p77087` in 2054.281 seconds. It has the exact same source and toolchain fingerprints; graph `sha256:388776b635dc545b0dfd6d1144d47be8b0033843263262d14465fe269bbc6319`. The two failed units are the existing `module.timeline.browser.range_scrolling` scenario's group and its dependent browser target summary. The other eighteen scenarios in that group passed, and all remaining release units passed. This run is retained as failed evidence, never as a passing release gate.

The Playwright report records a 60000 ms test-budget exhaustion near the final drag in `apps/web/e2e/timeline-range-selection.spec.ts`, with a 64968 ms terminal duration including teardown. Trace timings include 172 animation-frame clock calls consuming about 30.6 seconds and 407 setup HTTP calls consuming about 9.2 seconds. The scenario and Timeline production code were unchanged in this remediation. This appears to be an existing functional-test budget sensitivity; a narrow unchanged-scenario rerun and a fresh complete gate are required before closure. No timeout, assertion, golden or source file was changed to hide the failure. The progress monitor initially counted only `completed` events and missed the harness's separate `failed` event; that reporting mistake was corrected after inspecting the retained logs. The final result and failure attribution here supersede interim no-failure progress statements.

The diagnostic report is `browser-e2e-webserver-backed/browser-groups/functional-support-default-timeline-range-selection/playwright-report.json` beneath that run root, with its trace under the group's `playwright-output/`. The investigation uses `make task-guide ROLE=module-author OWNER=module.timeline` and the exact `make test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_scrolling` selector. Final documentation lint after the coordinated owner cleanup passed at `20261004T144413Z-p77395`; this later evidence append will receive a final lint pass.

The unchanged narrow Timeline scenario passed **11/11** at `20261004T151912Z-p37134` in 68.985 seconds including its complete fixture/build lifecycle. The first failed full gate's `make explain-run` confirms all 1260 units executed, zero skipped/canceled units, and 689 cleanup steps with no failed or blocked cleanup. The next `make release-check` uses the same frozen candidate; it is a whole-gate reproducibility rerun, not a substituted subset or a combined success claim assembled from separate runs.

The second complete attempt, `20261004T152104Z-p74695`, exposed a separate existing harness timing failure: `target:harness-contract`'s child-command diagnostic case received `timeout_failure` instead of its expected `artifact_error`. The helper allows 10000 ms for a Node/Make child startup; no harness source changed in this remediation. That attempt was canceled through the harness's SIGINT/cleanup path after diagnosis (189.332 seconds), rather than continuing a known failed gate. It is canceled/failed history, not success evidence.

The next full run uses the supported `CARTULARY_HARNESS_CAPACITY_OVERRIDE` to reduce host contention: `.cartulary/validation-config/reference-pack-capacity.json` contains `{"schema_id":"cartulary.harness_capacity_override.v1","cpu_tokens":8,"postgres_lanes":4,"object_store_lanes":2}`. The automatic snapshot had 19 CPU tokens, 8 PostgreSQL lanes and 3 object-store lanes. This is a run-local, ignored validation input; it does not change production configuration, source, assertions, timeout values, selected semantic rows, or service availability. The harness validates the override against detected bounds and retains its resolved capability snapshot. An isolated `make harness-contract` check precedes the bounded-concurrency whole-gate rerun.

The isolated `make harness-contract` passed **2/2** at `20261004T152459Z-p24229` in 27.528 seconds. `make explain-run` for the canceled second attempt reports 901 passed, 1 failed, 1 skipped and 357 canceled units, plus 79 successful cleanup steps and no failed/blocked cleanup. Those nonexecuted units are attributable to the deliberate cancellation and cannot satisfy validation. The next exact command is `make release-check CARTULARY_HARNESS_CAPACITY_OVERRIDE=.cartulary/validation-config/reference-pack-capacity.json`.

### Complete bounded-capacity run and remaining release blocker

That command completed **FAIL, 1259/1260**, at `.cartulary/test-results/20261004T152543Z-p28254` in 2567.943 seconds. All 1260 units executed; none were skipped or canceled. `make explain-run RESULTS_DIR=.cartulary/test-results/20261004T152543Z-p28254` independently confirms the result. Source and toolchain fingerprints remain exactly those of the passing fast run above. The execution graph is `sha256:730c7e016529d47d9134ce51a18ddebc0ed0a359c438ff968da617d0ce299bf2`, and the resolved capability snapshot is `sha256:837d0601513fe514ade92145d54094c8c0e37ad6c9e1c0032f889ad27a340b5b`. The top-level cleanup receipt records 588 steps with no failed or blocked step. This does not establish successful cleanup inside every nested test.

All four remediation workstreams' tests passed in that complete run, as did the previously failing Timeline group and `target:harness-contract`. The sole failed unit is `row:harness.browser.integration.ui_review_seeded_default`. Its retained stderr is `unit-logs/row-harness.browser.integration.ui_review_seeded_default/stderr.log` under the run root: the unchanged `realStackCleanupCases` helper failed in its `finally` recovery path with `browser recovery requires one retained managed suite proof`. The error occurs before the controlled cleanup-failure case and masks the original browser-allocation failure; the retained evidence does not establish that original failure's cause.

The nested fixture root is `.cartulary/test-results/fixture-cleanup-1791130012864-product`. Its service lifecycle records readiness followed by successful managed-service cleanup. The private runtime `/tmp/cartulary-harness-scratch/suite-runtime/suite-fixture-cleanup-1791130012864-product-ShMirv` retains a pending `browser_stack` recovery record whose lease target was never published, with no remaining managed-suite recovery record. It was preserved for owner investigation, not manually deleted or represented as complete private cleanup. No credentials or private lease contents are copied into this handoff.

Static inspection places the inconsistent recovery state in unchanged harness code: browser fixture acquisition registers a pending proof, its catch path only retires it when a lease exists, and preparation with a borrowed runtime can close the managed suite before the test's final recovery. The recovery helper requires exactly one suite proof whenever any browser record is present. This is independent of the changed Go staging adapter and composition/reference contracts. The original allocation failure remains unclassified; do not relabel it as a proven capacity timeout. The next diagnostic command is `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_default`, selected after `make task-guide ROLE=module-author OWNER=harness.browser`, with unchanged source and assertions.

The Testing Harness/browser owner must resolve or explain the failed allocation and the pending-proof recovery lifecycle, retain a focused regression if code changes, verify exact private-resource cleanup, and rerun the complete gate on the resulting source fingerprint. The earlier default-capacity timing failures also remain relevant reproducibility evidence. No Timeline, browser-harness, timeout, assertion, golden or routing exclusion was changed to obtain a green result. A successful isolated retry cannot substitute for a successful complete release run.

The unchanged isolated UI-review row passed **3/3** at `.cartulary/test-results/20261004T160912Z-p17945` in 253.730 seconds. It exercises the public seeded workflow and both real-stack cleanup cases. This narrows the reproducibility problem but does not repair the failed run's pending proof or make its complete gate successful. No further complete rerun was substituted for diagnosing the unresolved harness lifecycle.

### Execution handoff disposition

- **Implementation remediation complete for the four reported defects:** shared strict composition serialization and owner digests, set-valued framework lookup keys with intact provenance, typed bounded upload rejection across HTTP/operator adapters, and pre-parse URL character rejection. Focused tests, full fast tests, projection/boundary checks, and those workstreams' units in the complete release graph pass. The fresh-initialization policy applies; no compatibility reader or immutable identity rewrite was added.
- **Integrated acceptance remains blocked:** Phase 4 lacks a successful complete `release-check`. Preserve all three broad attempts and the focused diagnostic results above. Testing Harness/browser owns the failed-allocation/recovery follow-up; the project release authority must accept fresh complete-gate evidence before closing this exit. Counts assembled across runs are insufficient.
- **Specifications are not yet adopted:** coordinated Core/Extensions repairs and Reference Pack 0.2.0, Reporting 1.4.0, Report Composition 1.3.0 companion amendments are reviewable, with the responsible authorities and all sixteen gate dispositions above. Draft/proposed labels remain. Requirement-level human review, security posture approval and distribution-inventory confirmation remain explicit promotion gates.
- **Production conformance is not approved:** no deployment, production cutover, second implementation, license approval or cryptographic-posture decision was performed or inferred. Phase 5's actionable documentation handoff is prepared; overall plan acceptance and promotion remain blocked by the stated validation and authority gates.

The existing ledger material was preserved. Final source changes remain uncommitted on `main` based on `54f42195572c04b2bc9bb9003f2338eef0ba37db`; all three complete/canceled attempts and the passing fast run identify source fingerprint `sha256:5c9373a4378d601d416e9a2371c7275a784640d4ff1560eef8467312432c73fb`. Later handoff edits are documentation-only and are excluded from executable evidence by repository policy. `make agent-finalize` ran before broader verification; retained-run maintenance was skipped because no qualifying successful full warm `check` root was supplied as `RESULTS_DIR`. A separate duplicate `make check` was omitted because `release-check` already covers its graph. No requested product check was skipped to produce success.

Final documentation verification: `make lint-markdown` passed at `.cartulary/test-results/20261004T161416Z-p43948` (summary `adhoc/lint-markdown/tool-run-summary.json`), and `git diff --check` passed. `make explain-run` for the isolated UI-review retry confirms the same source/toolchain fingerprints, zero failed/skipped/canceled units, and five successful top-level cleanup steps. These final documentation edits do not change product validation evidence.
