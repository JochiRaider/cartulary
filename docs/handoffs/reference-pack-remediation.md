# Reference Pack cryptographic remediation — Windows 11 / WSL2

This is the sole controlling tracker and completion record for the Reference Pack and application cryptography iteration. It is implementation support, not a normative owner, adoption record or conformance claim. Tests, generators, runtime metadata and release evidence must not read, stat, hash or otherwise depend on this document.

## 1. Scope and source posture

**Current authorization is implementation of S16–S25, refocused on WSL2 under Windows 11.** On 2026-10-05 the user authorized execution, then selected WSL2 and deferred pure/native Linux indefinitely. This supersedes the earlier native-Linux target requirement and access request. RP-P04, S16, S16a and S16b are DONE for their documented exits; S16c is DONE; S17–S20 and prerequisites S19a/S19b/S20a/S20b are DONE; S21 and its prerequisites are DONE; S22 and S22a are DONE; S23 and S23v–S23y are DONE; S23a is DONE and RP-F21 is closed; S24 and S24a are DONE; S25 is DONE after final-candidate release `20261006T091921Z-p85023` passed and its handoff was recorded. Completed work, failures and original validation identities remain historical records below. Earlier document-only restrictions do not constrain the existing implementation authorization.

| Scope fact | Decision |
| --- | --- |
| Current baseline | Execution began at clean `df720535b6b6622cab8b3d399e868dc968e0f6c4` on 2026-10-05. The final candidate retains the original uncommitted work plus all authorized implementation changes; no commit, reset or automatic deployment was performed. Earlier baselines and run identities remain historical. Preserve any concurrent user changes. |
| Execution disposition | S16c → S17 → S18 → S19 → S20 → S21 → S22 → S23 → S23a → S24 → S25 completed with recorded prerequisite checkpoints and one complete final-candidate release. Future candidate changes require renewed validation. No automatic deployment, destructive reset, retained-state conversion or customer rollout. |
| Target and growth constraint | Preserve the 197-file Reference Data/assembly inventory; extend planning across the application, authentication, sealing, recovery, integration and build boundaries in §2. Keep domain formats and authorization with their owners so later phases do not require a shared domain-token or provider framework. |
| S07 disposition | S16–S25 implement the single pinned-module policy and qualify the WSL2 package. S07 is DONE for the Windows 11 / WSL2 engineering scope after S25 completed and release `20261006T091921Z-p85023` passed. Formal CMVP operating-environment qualification is a separate, unestablished claim, not implied by umbrella completion. Its old DEFERRED records remain historical. |
| Claim boundary | Security cryptographic operations performed by `server`, `migrate` and `operator`, including enabled optional integrations. Database servers, object-storage servers, external TLS terminators, browsers and host certification remain separately owned deployment dependencies. Application-side transport and certificate verification remain in scope. |
| Production policy | One supported cryptographic implementation and release policy. No standard-versus-regulated algorithm branches, crypto provider plug-in framework, silent fallback or security-operation bypass. |
| Initial module candidate | Go Cryptographic Module `v1.0.0-c2097c7c` in the repository's pinned Go `1.27.1` toolchain; CMVP certificate `5247`. Archive SHA-256: `daf3614e0406f67ae6323c902db3f953a1effb199142362a039e7526dfb9368b`. Verify and pin exact archive/build identity; never use floating `latest`, `certified` or `inprocess` selectors. Certificate and security disposition must be rechecked before qualification. |
| Reference environment | Windows 11 x64 → WSL2 → Ubuntu x86_64, with Docker Desktop's WSL2 backend and the existing digest-pinned Linux/amd64 application image. The inspected reference guest is Ubuntu 26.04 LTS. Record the Windows build, WSL package/kernel, guest, Docker backend/daemon, CPU/resource limits and resolved image separately. Run the complete release here, not a reduced compatibility subset. |
| Native Linux disposition | Pure/native Linux deployment, platform qualification, release rehearsal and operational documentation are DEFERRED indefinitely, with no date or dependency on S16–S25. Resume only on an explicit future scope decision. Linux binaries, container userlands and filesystem primitives remain necessary inside WSL2; this is not a Windows-native port. |
| Qualification terminology | In active S16–S25, package/verifier qualification means engineering acceptance on the recorded WSL2 target. Formal CMVP applicability to the exact Windows/WSL/guest/container/CPU stack is UNESTABLISHED and not claimed. It gates any regulated deployment claim, not completion of this WSL2 engineering iteration. No alternate crypto mode or weakened service policy follows from that distinction. |
| Execution mode | Build against the pinned module and require enabled FIPS mode at startup. Use `fips140=only` for targeted diagnostic assessment, not production deployment or sole compliance evidence; it can produce false positives and negatives. |
| Compatibility | Fresh WSL2 deployments under the single cryptographic policy only. Reject incompatible application state and backup formats before mutation. Historical deployments and backups stay with their matching releases. No password conversion, legacy cryptography fallback, dual writers or automatic reset. |
| Early feasibility | S16 must establish supported service/purpose mappings, enforceable driver admission and an executable WSL2 validation route before S17. Missing required WSL2/dependency capabilities become explicit prerequisites. Formal environment claims remain separate; a missing native-Linux host is no longer a blocker. |
| PostgreSQL authentication | Runtime, migration and recovery use separately provisioned client certificates with verified TLS and existing effective-role admission. Prohibited authentication requests must reject before prohibited cryptography; clearing a driver password alone is insufficient. No password/passfile fallback. |
| Enforcement staging | S17 delivered independently tested components; completed S23a activates mandatory facade invocation and default pinned application/harness builds. RP-F21 is closed by its actual binary/state evidence. No runtime bypass or second production policy. Intermediate candidates and their state are disposable, not release-ready. |
| WSL2 operation | Operator-started only: no Windows startup automation; guest backup timers run while available, with explicit backup-age and due-verification checks after restart. Docker Desktop availability replaces any assumption of guest docker.service. Downtime never waives freshness. |
| Startup stages | Admit actual binary identity and configuration before service acquisition; then use only necessary resources for read-only retained-state inspection before mutations, bootstrap, leases, journals, jobs or publication. Application cryptographic-format identity is distinct from module build identity. |
| Transport | Protect network hops including proxy-to-application traffic. External TLS termination does not justify an unprotected application network listener. |
| Reference Pack signatures | Preserve Ed25519, canonical bytes, trust rotation and `tuf_1_0_35_offline_bundle_v1`. Amend the owner's FIPS exclusion to distinguish execution qualification from signature semantics. The selected module includes EdDSA; a second signature algorithm or method would add unnecessary format/trust complexity. |
| Permitted non-security use | Explicitly reviewed protocol-only hashes such as the WebSocket handshake hash may remain. They must never provide authentication, integrity protection, key derivation or authorization. Indicator MD5/SHA-1 strings remain domain data. An exception is not an enforcement bypass. |
| Non-goals | No high availability, new pack profiles, external corpus distribution, unrelated module reshuffling, UI redesign or automatic customer rollout. Do not reopen completed runtime/test separation or recreate retired reporting dossiers. |
| Authority | Adopted subsystem NLSpecs and normative Core sections govern their scopes; typed contracts are downstream projections. Verification routing belongs to its machine owners, not this tracker. `docs/domain.md` owns vocabulary/navigation and `docs/design.md` supplies design direction. Owner adoption, package qualification and customer deployment approval are separate decisions. |

The candidate and constraints above follow the [NIST certificate](https://csrc.nist.gov/projects/cryptographic-module-validation-program/certificate/5247), [module security policy](https://csrc.nist.gov/CSRC/media/projects/cryptographic-module-validation-program/documents/security-policies/140sp5247.pdf), and [Go FIPS guidance](https://go.dev/doc/security/fips140). They define a qualification target, not a claim that the current package is qualified. Password parameters below follow the [OWASP password-storage guidance](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html). The [WebSocket security discussion](https://www.rfc-editor.org/rfc/rfc6455#section-10.8) informs the narrow protocol exception.

The existing image pin is `gcr.io/distroless/base-debian12:nonroot@sha256:7f0c72cd138b442ae0deeb69c08b1acf5525439ba251a49ad93c320a061567e5` in `tools/toolchain_pins.json`. Keep it as the Linux container userland on WSL2; it does not require a Debian host. S24 must record the resolved platform image and the actual WSL2 execution stack. [Microsoft's WSL architecture description](https://learn.microsoft.com/en-us/windows/wsl/compare-versions) distinguishes the Linux guest kernel from its Windows host. Neither a Windows entry nor a Linux distribution entry in the module policy alone establishes formal applicability to this complete stack.

Repository sources, caches, private keys, confined application roots, recovery staging and service volumes belong on Linux filesystems inside the WSL2/Docker environment. Windows-mounted paths are not the reference storage profile. Operator export/import across that boundary needs explicit copy, permission and integrity checks; it must not weaken confinement or introduce another storage implementation. S23/S24 cover Windows-to-application HTTPS/WSS, certificate trust in each actual client, Docker/WSL networking, and failure after service interruption. System-wide WSL shutdown, host reboot, trust-store changes and firewall changes are operator procedures, never unannounced test cleanup.

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

This is an authorized fresh-deployment cutover, not a blanket behavior-preserving refactor. The user authorized the behavior changes below on 2026-10-05, subject to sequential workstream exits. Existing behavior is retained where it carries a durable domain or security guarantee, not merely because it exists.

| Contract | Owner / evidence | Retain or planned change | Existing / required validation | Risk |
| --- | --- | --- | --- | --- |
| Pack signed bytes, sets and provenance | Reference Pack owner, private verifier and canonical fixtures | Retain Ed25519, method identity, canonical digests, complete signer sets, immutable anchors and sixteen existing profiles | Independent vectors; actual bootstrap/TUF/rotation/replay/historical/destination paths under pinned module, S22 | High |
| Five consumer operations and source participation | Reference Data DTOs, retention and assignment ports | Retain operation set, exact-set selection, authorization, pins, transaction/lock order and source-owner revision effects | Consumer/provenance/retention tests and affected source-owner slices | High |
| Administrative HTTP/operator/Jobs behavior | Core 01/Core 04, adapters and assembly | Retain routes, actor attribution, replay, idempotency, bounded admission, finality, atomic publication and safe errors | Existing HTTP/Jobs/concurrency/operator tests plus S24 packaged parity and startup rejection | High |
| Password and MFA records | Core 04, Auth/bootstrap and OpenAPI owner | Change to bounded versioned PBKDF2-HMAC-SHA-256 records and SHA-256 TOTP; fresh credentials/enrollment required | S19 bootstrap/login/enrollment/replacement/reset/revocation tests, concurrent-load bound, enum compatibility and browser flows | High; authorized fresh-deployment break |
| Secret, cursor and conflict-token representations | Auth, pagination, Revisions and Network Flow | Change sealed versions and derived keys; reject old formats; retain claim/AAD/expiry/error semantics | S18/S20 tamper, wrong-purpose/key, user/session/route replay, expiry and rotation | High; authorized fresh-deployment break |
| Backup, journal and recovery admission | Recovery owner and `contracts/recovery/` | One current encryption codec and explicit profile binding; preserve bounded streaming, target safety and export-root binding | S21 real backup/due verification/restore, adversarial chunk/framing tests, interruption and no-publication tests | High; authorized fresh-deployment break |
| Runtime/configuration admission and transport | Core 04, application facades and platform adapters | Ordinary/disabled/wrong-module binaries and insecure production transport reject; fresh-state cutover before mutation | S17/S23/S24 fail-closed tests and disposable package scenarios using actual binaries | High; authorized fresh-deployment break |
| Authentication browser and generated contracts | Auth OpenAPI owner, web authentication model and fixtures | Update algorithm projection through normal compatibility process; preserve routes, session UX and safe error handling | Auth frontend type/unit/browser evidence; generated drift and OpenAPI compatibility disposition | Medium; authorized public enum change |
| WebSocket protocol and domain indicators | Transport owner and indicator owners | Preserve reviewed handshake hash and opaque indicator digest strings; no security use of an exception | S23 reachable-call review, positive/negative security evidence and targeted diagnostic findings | Medium |
| UI/workbook/view schemas/selectors | Existing frontend and source owners | No planned redesign, entity-row/view-schema changes or selector/golden refresh | Preserve existing suites; expand only if an actual auth/integration change reaches these surfaces | Low |
| Runtime/test asset split, harness and release | Contract-generation and Testing Harness owners | Preserve separated fixture family, semantic ownership and canonical artifacts; add module/cache/diagnostic/package acceptance inputs | Drift/boundary/harness checks, real-binary negatives and complete final release | High; authorized release policy change |
| Deployment trust, clock and external data | Reference Pack and deployment owners | Preserve explicit trusted-clock assertion, deployment-supplied trust and licensing/source-identity constraints | Existing admission/rejection tests and ordinary deployment review | High; test trust is not deployment approval |

## 5. Coupling and boundary findings

RP-F01–F19 and their historical outcomes remain below. The former RP-F09/S07 deferral is now expanded, not erased. No completed structural or package repair is reopened without new evidence.

| Finding | Evidence | Risk if unresolved | Classification | Owner | Required action |
| --- | --- | --- | --- | --- | --- |
| RP-F20: current owner exclusions do not support the chosen regulated profile | RP-REQ-261 requires a new method for FIPS; RP-REQ-264 calls it unsupported; Core security projections still describe current primitives | Implementation can contradict owners or make an unsupported claim | must_fix | Core security / Reference Pack / Recovery | S16 separates execution qualification from unchanged signature semantics and coordinates all security-service requirements. |
| RP-F21: build settings are not authoritative runtime admission | Go toolchain/image pins exist, but shared build/cache and all three facades lack the proposed exact module policy gate | Ordinary, stale or overridden binaries execute under a regulated label | must_fix | Build / platform policy / application facades | CLOSED by S23a; exact binary and retained-state admission pass before mutation. |
| RP-F22: duplicated nonce/KDF mechanics and permissive key handling | Authn hash-based derivation and fallback secrets; Recovery truncation; GCM nonce construction across multiple owners | Approved algorithm names conceal unsupported use; changes drift across callers | must_fix | Platform primitive boundary and calling owners | S18 explicit key contracts, domain-separated HKDF and module-generated-nonce sealing; S20/S21 migrate callers. |
| RP-F23: password/MFA primitives and bootstrap encoding conflict with the selected profile | Argon2id in Authn/bootstrap; SHA-1 TOTP and public enum/client assumptions | Base authentication prevents qualification; encoders or clients diverge | must_fix | Auth / bootstrap / Core 04 / OpenAPI | S19 one password owner path and bounded PBKDF2; SHA-256 enrollment and complete lifecycle/browser validation. |
| RP-F24: token migration can erase owner-specific binding | Pagination, `cft3`, `nfc2` and Auth secrets have separate claim/AAD/expiry rules | New shared machinery permits replay, wrong-purpose acceptance or unsafe error disclosure | must_fix | Existing token owners | S20 adopts only primitive mechanics and retains owner-local semantics with adversarial tests. |
| RP-F25: recovery has parallel encryption formats and caller nonces | `encryption.go` and `streaming_encryption.go`; journal and backup envelope schemas | Policy-incompatible backup creation or unusable/unsafe restores | must_fix | Recovery | S21 one current streaming format for all sizes, authenticated chunk identity/finality and profile-bound admission. |
| RP-F26: module support alone does not qualify the actual pack verifier | Real bootstrap/TUF/historical/portable code paths differ from an isolated Ed25519 test | A different runtime path or malformed input behavior escapes evidence | must_fix | Reference Data / assembly | S22 execute independent vectors and complete lifecycle paths under the pinned module. |
| RP-F27: dependency and transport behavior is not established by direct imports | pgx SCRAM, MinIO signing/checksums, OIDC/SAML, TLS/telemetry and WebSocket; insecure current package transport options | An optional integration or unprotected transport invalidates the application claim | must_fix | Platform adapter / configuration owners | S23 service/purpose/parameter mapping, certificate/transport enforcement and narrow reviewed non-security exceptions. |
| RP-F28: existing package acceptance lacks the selected module/WSL2 environment contract | Existing three smokes and passing release establish the prior package only | Source tests pass while cache, image, startup inputs or Windows/guest/daemon boundaries undermine the package | must_fix | Package / build / Testing Harness | S24 real-binary identity and negative admission, fresh-state flows, WSL2 environment/boundary evidence and canonical release routing. |
| RP-F29: readiness states can be conflated | Previous release passed; Reference Pack candidate and deployment acceptance remain separately governed | Historical results or draft changes are represented as current compliance/rollout approval | must_fix | Specification / release / deployment owners | S25 complete one final-candidate release and separately record implementation, qualification, adoption and deployment decisions. |
| RP-F30: stable formats and existing owner boundaries remain useful | Completed DTO/persistence/fixture separation; canonical pack identity; protocol-only hash/data uses | Unnecessary redesign introduces migration and coupling without a security benefit | intentional/no_action | Existing subsystem owners | Preserve these boundaries; no algorithm negotiation, blanket hash bans or unrelated module moves. |

## 6. Refactor workstreams and phase sequencing

Each implementation slice is its own workstream. The authorized plan is intentionally sequential: **S16c → S17 → S18 → S19 → S20 → S21 → S22 → S23 → S23a → S24 → S25**. Later discoveries become explicit prerequisite slices; they do not silently expand an active slice. S25 must complete last.

| Phase | Workstream / class / owner | Required previous | Required next | Goal and affected boundary | Validation / handoff checkpoint |
| --- | --- | --- | --- | --- | --- |
| 1: Specification and execution boundary | S16 / root / Core security and affected specification owners | User implementation authorization, received 2026-10-05 | S17 | Consistent profile/cutover requirements, supported integration/environment feasibility and projection design | Owner review and complete service/exception routing; no unresolved owner contradiction or unsupported driver admission. |
| 1: Specification and execution boundary | S16c / prerequisite / Core security | S16 completed | S17 | Reconcile acceptance criteria and staged execution | Owner requirements and criteria agree; document checks pass. |
| 1: Specification and execution boundary | S17 / chain / build, platform policy and application facades | S16c | S18 | Exact build/module identity and admission components | Isolated pinned component tests pass; production enforcement is S23a. |
| 2: Application cryptography | S18 / chain / platform primitives | S17 | S19 | Explicit keys, HKDF and random-nonce sealing | Known-answer, domain-separation, misuse and key-use-bound tests. |
| 2: Application cryptography | S19 / chain / Auth and bootstrap | S18 | S20 | One password path and SHA-256 TOTP | Full credential/MFA/session/browser lifecycle and resource bounds. |
| 2: Application cryptography | S20 / chain / Auth, pagination, Revisions and Network Flow | S19 | S21 | Owner-specific tokens on shared primitives | Binding, replay, tamper, expiry and rotation evidence for every caller. |
| 2: Application cryptography | S21 / chain / Recovery | S20 | S22 | One current backup/journal encryption policy | Real restore/due verification and adversarial stream/admission checks. |
| 3: Verification and integrations | S22 / chain / Reference Data and reference assembly | S21 | S23 | Qualify the actual pack verifier | Shipped paths, lifecycle and unchanged canonical identities. |
| 3: Verification and integrations | S23 / chain / platform adapter and configuration owners | S22 | S23a | Qualify enabled transitive services and transport | Supported service mapping plus positive/negative runtime evidence; no unresolved security use. |
| 3: Verification and integrations | S23a / chain / application facades and build | S23 and S17 | S24 | Activate mandatory binary/state admission and pinned defaults | Original S17 actual-binary acceptance passes; close RP-F21. |
| 4: Package and handoff | S24 / chain / package, build and Testing Harness | S23a | S25 | WSL2 package acceptance and complete environment evidence | Three package scenarios, Windows/WSL boundaries, canonical routing/results and accountable cleanup. |
| 4: Package and handoff | S25 / chain / release and handoff owners | S24 and every active implementation prerequisite | None | One complete WSL2 final-candidate validation and maintainable handoff | Active implementation slices DONE, final release/cleanup pass; native Linux remains indefinitely deferred and formal environment claims remain separate. Completes last. |

| Phase | Principal risks / sequencing reason | Binary phase exit |
| --- | --- | --- |
| 1 | Mistaking build selection for operational qualification; contradicting adopted owners; admission after resource mutation | The profile is a consistent candidate and the policy is executable through typed/build/runtime inputs with no Markdown dependency. Any actual owner contradiction is BLOCKED and resolved before dependent implementation. |
| 2 | Authentication lockout, unbounded password work, token-binding regressions, nonce misuse and unusable backups | Every directly owned security primitive follows the selected policy; incompatible state rejects before mutation and affected owner lifecycle/negative tests pass. |
| 3 | Hidden transitive behavior, unsupported parameters/purposes or an overly broad exception | Every enabled application security service has an approved-use mapping and execution evidence; unresolved use blocks qualification. Pack format identity remains stable. |
| 4 | Incorrect host/image assumptions, incomplete cleanup, stale evidence or premature deployment claims | Package scenarios and one complete final release pass with cleanup; implementation/qualification/adoption/deployment statuses are explicit, with promotion blocked where required acceptance is missing. |

**Execution rule:** mark the next workstream IN_PROGRESS before its first implementation change. At its exit, record the changed owners/areas, compatibility consequence, commands and canonical run identities, failures, residual risks and next dependency; mark DONE only when the binary exit is satisfied. **Update this tracker after each completed workstream and before beginning the next one.** Failed prerequisites remain BLOCKED, not DONE. Allowed statuses are TODO, IN_PROGRESS, BLOCKED, DONE, DEFERRED and DROPPED. Compact canonical evidence replaces duplicate narrative dossiers.

## 7. Implementation workstreams and exits

**S16, S16a and S16b remain DONE; S16c/S17/S18/S19 and S19a/S19b are DONE; S20 is DONE; S21 and its prerequisites are DONE; S22 and S22a are DONE; S23 and all recorded prerequisites, S23a, S24 and S24a are DONE; S25 is DONE with final-candidate acceptance recorded below.** The native-target access blocker is superseded by the user's target decision, not a passing native qualification. S16 closes source/design feasibility only; runtime, generated projections and package evidence remain with their assigned slices. §8 gives the discovered Make routes. Each slice records its own evidence; planned checks are not completed checks.

### S16a — Establish supported PostgreSQL authentication admission

**Status:** DONE. **Owner:** platform.postgres and dependency inputs. **Areas:** implementation dependency, tests and harness routing. **Discovered prerequisite for:** S16; S17 cannot begin until S16 and this prerequisite pass.

- **Gap / remediation:** the inspected pgx v5.9.2 dispatches password mechanisms without a pre-authentication policy. Upgrade to released v5.11.0, whose `pgconn.Config.RequireAuth` admits `none` for certificate-authenticated TLS and rejects password, MD5, SCRAM, GSS and OAuth requests before mechanism dispatch. Use the supported driver policy; do not implement a production wire-protocol substitute.
- **Rationale / benefit:** enforce authentication admission at the dependency boundary that owns dispatch, retaining normal driver maintenance and all existing purpose/effective-role checks. Certificate presentation and verified TLS remain separate required S23 controls; `none` alone does not authenticate a server or client.
- **Compatibility:** no application authentication change in this prerequisite. The dependency upgrade must preserve existing connections and DB port behavior. S23 deliberately removes the old production password connection policy.
- **Risk if unresolved:** a malicious or misconfigured server can trigger prohibited cryptography before application admission, invalidating the certificate-only design.
- **Validation:** exercise challenged authentication messages through the released driver's connection path with a synthetic in-process server and assert rejection without a password/mechanism response; exercise direct AuthenticationOk as the positive protocol case. Run PostgreSQL owner tests, affected interface/build checks and service-backed role/identity tests. Tool-managed dependency sums must be produced through Make.
- **Exit / next:** released dependency capability and regression routes pass, with upgrade compatibility and failures recorded here; then resume S16 feasibility/specification work. This prerequisite does not qualify production TLS, the module or the operating environment.

### S16b — Establish Windows 11 / WSL2 environment feasibility

**Status:** DONE. **Owner:** deployment and Testing Harness.
**Areas:** environment inspection, specifications, validation design and handoff.
**Discovered prerequisite for:** S16, before S17; separate from S24 package execution.

- **Remediation:** use the accessible Windows 11 / WSL2 / Ubuntu / Docker Desktop
  stack. Record host, guest and daemon identities separately; establish Linux
  filesystem placement, verified network hops, resource limits and the full Make
  validation route. Amend the harness platform matrix so native Linux is no
  longer required for this release. No Windows-native application binaries.
- **Rationale / benefit:** one actual target and one release path remove an
  unused platform dependency without creating a second cryptographic policy.
- **Compatibility:** native-Linux support/qualification work is indefinitely
  deferred. Existing Linux container images, Go targets and guest confinement
  remain. Inspection does not modify a deployment or grant a formal CMVP claim.
- **Risk if unresolved:** Windows/guest trust differences, mounted-filesystem
  semantics, daemon placement and shared resource contention can invalidate tests
  or operational assumptions even when individual primitives pass.
- **Validation / exit:** observed target identities and `make doctor` pass;
  existing pinned-module feasibility evidence is accounted for without relabeling;
  the WSL2 performance, network, storage and interruption obligations have owners
  and public validation routes; formal module-environment applicability has an
  explicit unestablished disposition. Full package evidence remains S24 work.
- **Completion / next:** observed target and doctor evidence, coordinated owner
  requirements and validation routes pass the feasibility exit; §10 records the
  exact evidence and limitations. Finish S16's owner checkpoint before S17.
  No native-Linux machine or access method is required.

### S16 — Define the cryptographic policy and amend owners

**Status:** DONE for coordinated source/design feasibility; not runtime completion or formal adoption.

**Owner:** Core security, with Reference Pack, Recovery, configuration, OpenAPI and Testing Harness owners. **Areas:** specifications, documentation and typed contract design. **Gaps:** RP-F20/F29. **Depends on:** user implementation authorization, received 2026-10-05. **Next:** S17 only after feasibility and owner exits pass.

- **Remediation:** define one application cryptographic profile covering passwords, TOTP, key derivation, encryption, security signatures, enabled integrations and production admission. Amend Core 04 and affected Core 01/Recovery/configuration clauses together. Amend RP-REQ-261 and RP-REQ-264's FIPS exclusion without changing Ed25519 or `tuf_1_0_35_offline_bundle_v1`. Identify the corresponding authored Recovery schemas, configuration contracts, Auth OpenAPI enum, build inputs and harness projections before coding.
- **Boundary:** distinguish domain semantics, module-service usage and deployment qualification. Enumerate each security service and owner, and each permitted non-security protocol use with its reason and validation route. Keep the review in normal owner changes and existing machine projections; do not create an approval registry or replacement evidence dossier.
- **Early feasibility:** verify the actual module/toolchain and WSL2 execution route; map PostgreSQL authentication rejection, S3 signing/checksums, OIDC/SAML, TLS, telemetry and WebSocket operations before S17. Client-certificate authentication must reject prohibited requests through supported dependency capabilities. Unsupported required capability is an explicit prerequisite, never a replacement protocol. Freeze password-work capacity, overload behavior and WSL2 resource/latency acceptance before measurement; keep the existing numeric bounds. Record formal environment applicability separately without demanding a native-Linux substitute.
- **Rationale / long-term benefit:** a future module update can change qualified build/environment identity without changing pack signatures, trust history or unrelated domain contracts. Requirements become coherent enough for implementation and human projection review.
- **Compatibility / migration:** explicitly declare the fresh-state cutover, breaking credential/MFA/sealed-state formats, startup rejection and changed transport admission. Define how incompatible state is identified before mutation. Applied migrations remain immutable; historical deployments/backups use matching releases.
- **Unresolved risk / risk if left open:** the present Reference Pack exclusion is incompatible with the selected direction until amended. Tests cannot waive it. Missing or contradictory service requirements can produce implementation that overclaims qualification.
- **Validation / evidence:** review Core security, Reference Pack, Recovery/configuration, OpenAPI and harness owners as one candidate; use `make lint-markdown` and `git diff --check`. No current owner adoption or new crypto execution evidence is asserted.
- **Binary exit:** every affected service, protocol exception, cutover rule and projection has one consistent owner and validation route; integration and environment feasibility are supported, including driver rejection before prohibited cryptography. An actual contradiction is recorded as **BLOCKED: owner contradiction**; any unresolved feasibility prerequisite also blocks dependent implementation. Formal adoption remains a separately recorded status.
- **Rollback:** withdraw conflicting candidate amendments together and keep production promotion blocked; do not implement against contradictory requirements.

### S16c — Reconcile remaining source acceptance criteria

**Status:** DONE. **Owner:** Core security/configuration, Revisions and Testing Harness. **Areas:** specifications and tracker. **Depends on:** completed S16. **Next:** S17.

- **Remediation:** reconcile AC-526 with configuration v3, Revisions key-ring v2 and conflict tokens v4; review Recovery version and WSL2 acceptance requirements. Record operator-started operation and S17/S23a staging without changing the final production policy.
- **Rationale / benefit:** consistent source targets prevent contradictory implementation and verification. Completed source feasibility and historical evidence remain intact.
- **Compatibility / risk:** the existing fresh-deployment break is unchanged. Unresolved criteria would admit incorrect formats or misrepresent interim builds as qualified.
- **Exit:** affected requirements/criteria agree; owning guide, Markdown lint, diff and document review pass. Formal CMVP applicability, adoption and customer deployment approval remain separate.

### S17 — Implement build identity and admission components

**Status:** DONE. **Owner:** build/toolchain, platform cryptographic execution boundary, migration-owned state admission and application facades. **Areas:** implementation, authored inputs, harness and tests. **Gap:** RP-F21; infrastructure completed here, production closure recorded in S23a. **Depends on:** S16c. **Next:** S18.

- **Remediation:** pin Go `1.27.1`, selector `v1.0.0-c2097c7c` and archive SHA-256 `daf3614e0406f67ae6323c902db3f953a1effb199142362a039e7526dfb9368b` in authored inputs; extend caches, receipts and fingerprints. Check `crypto/fips140.Version()` equals `v1.0.0`, resolved build metadata equals the full selector, actual toolchain and enabled mode independently.
- **Admission:** implement narrow binary admission and separate read-only application state inspection. No bucket/directory/journal/lease/bootstrap creation during inspection. Migration-owned singleton `cartulary.application_crypto_format.v1` is committed only after successful admitted fresh initialization, with one authoritative exclusion boundary across freshness recheck, migration and identity commit. Partial initialization without identity rejects; a genuinely empty target may retry. Ordinary server/operator commands only inspect, never stamp. Coordinate bucket initialization after database admission.
- **Staging:** exercise components and pinned builds in isolated harness outputs. Default pinned application/test builds and mandatory facade invocation activate in S23a after owners and integrations are ready. No runtime bypass flag or alternate production policy. Intermediate state cannot supply final format evidence.
- **Assessment:** author `cryptographic-policy-assessment` in the existing public task/harness inputs; strict mode remains diagnostic.
- **Rationale / benefit:** executable identity and stored-format compatibility remain separate, independently testable boundaries. Future module qualification does not force domain-format changes.
- **Compatibility / risk:** fresh deployment only; no stamping retained state, legacy fallback or automatic reset. Premature enforcement would break unmigrated owners; premature release would admit unsupported operations.
- **Exit:** correct/contradictory identities, cache contamination, read-only rejection, concurrency, interruption and cleanup pass component/candidate tests. This completes infrastructure, not production admission. S17's original production acceptance is retained intact in S23a.

### S17a — Consolidate generated Node readiness recipes

**Status:** DONE. **Owner:** harness.command_surface. **Areas:** generator, generated Make outputs and harness tests. **Discovered prerequisite for:** S17, before S18.

- **Gap / remediation:** the required assessment target increases generated bindings to 186757 bytes, above the existing 184320-byte budget. Factor the repeated Node readiness shell prelude into one generated runtime macro; keep ordering, environment stripping, quiet prerequisite execution and skip behavior identical.
- **Rationale / benefit:** remove repeated mechanics instead of raising a growth limit. Later public targets gain room without duplicating executable policy.
- **Compatibility / risk:** no public input/output change. Incorrect Make expansion could alter prerequisite ordering or command visibility and block all public targets.
- **Validation / exit:** normal Make generation passes unchanged density limits; command-surface/harness generation checks and representative public invocations pass. Then resume S17. Historical generated roots are changed only through Make.

### S17b — Enclose migration discovery in initialization admission

**Status:** DONE. **Owner:** module.database_migrations and PostgreSQL/application composition. **Areas:** migration lifecycle, admission and tests. **Discovered prerequisite for:** S17.

- **Gap / remediation:** pgx admission is supported, but Goose v3.27.0 `Provider.Up` calls `HasPending` before its session-lock hook; discovery may create the ledger. Put serialized freshness admission around the complete provider operation, including discovery, successful postconditions and format identity commit. Use supported library/capability boundaries; preserve borrowed handles, bounded cancellation and cleanup. Do not infer freshness from an empty ledger or add a historical-state exception.
- **Rationale / benefit:** one exclusion lifetime makes all mutation ordering explicit and survives library internals. A callback that runs after discovery cannot enforce read-only rejection.
- **Compatibility / risk:** only staged components change; production activation remains S23a. Without this prerequisite, an incompatible launch can mutate the database and a fresh launch can reject its own prematurely created ledger.
- **Exit:** fresh/current, retained-state, concurrent initialization, cancellation, partial interruption and cleanup tests execute through the complete provider path and prove rejection precedes ledger creation. Then resume S17; do not begin S18 while blocked.

### S18 — Establish approved key derivation and sealing primitives

**Status:** DONE. **Owner:** narrow platform primitive boundary, with existing key/purpose owners as callers. **Areas:** implementation, contracts and tests. **Gap:** RP-F22. **Depends on:** S17. **Next:** S19.

- **Remediation:** replace custom hash-based derivation with standard-library HKDF-SHA-256 and explicit, unambiguous domain separation. Require explicit master keys of the profile's exact supported size; remove production fallback secrets, permissive excess lengths and silent truncation. Keep key identifiers, rotation, authorization and secret-purpose reuse checks with their owners.
- **Primitive boundary:** provide a small seal/open API using the module's AES-GCM random-nonce facility. Production callers never provide GCM nonces or test entropy. Define versioned framing and associated-data inputs without owning domain payload schemas. Independently derive per-object/message keys where necessary so safety does not depend on a deployment-global nonce counter. Specify and enforce per-key use limits, including the module random-nonce limit, over the actual key lifetime.
- **Rationale / long-term benefit:** consolidate difficult derivation, framing and nonce mechanics once; keep future token/artifact growth local to each domain owner. Avoid a generic crypto service, provider API or shared domain-token framework.
- **Compatibility / migration:** derived keys and sealed representations change. Old state is rejected by explicit version/profile admission, not probed with a legacy decoder. Existing HMAC/SHA-256 callers such as upload tokens retain their semantics while adopting correctly separated keys.
- **Unresolved risk / risk if left open:** primitive names alone do not establish approved usage; ambiguous context encoding, reused keys, malformed framing or lifetime-limit assumptions can undermine isolation.
- **Validation / evidence:** known-answer HKDF vectors; distinct owner/purpose/object contexts; malformed/oversized inputs, tampering, wrong keys/AAD and key-use boundaries. Verify production code cannot inject nonce/entropy. Run affected platform/owner slices and boundary checks; migrate callers only in their owning subsequent slices.
- **Binary exit:** the primitive API and exact key contract pass all positive/negative vectors and usage-bound checks; no production caller of the new primitive supplies nonces or test entropy; caller migration obligations are explicitly assigned.
- **Rollback:** keep the new primitives unexposed until their contract is complete; do not introduce automatic fallback in consumers.

### S19 — Replace password and MFA cryptography

**Status:** DONE. **Owner:** Auth lifecycle and shared authentication primitives; bootstrap, Core 04 and OpenAPI owners participate. **Areas:** specifications, implementation, OpenAPI/contracts, tests and documentation. **Gap:** RP-F23. **Depends on:** S18. **Next:** S20.

- **Remediation:** use standard-library PBKDF2-HMAC-SHA-256 with an initial fixed cost of **600,000 iterations**, **16-byte random salt** and **32-byte result**. Use a versioned, bounded record format and one creation/verification policy for bootstrap and ordinary credentials. Remove duplicate Argon2 encoders and legacy verification paths.
- **MFA:** use SHA-256 TOTP with **32-byte secrets**, retaining six digits, 30-second periods and the existing bounded acceptance window. Update enrollment, verification and the authored OpenAPI algorithm enum through the normal compatibility process, then regenerate clients/fixtures. Keep authentication, session revocation, replacement/reset authorization and browser behavior owned by Auth.
- **Rationale / long-term benefit:** one credential policy removes divergent encoders and unsupported authentication primitives. Explicit versions and bounded parameters permit deliberate future cost strengthening without arbitrary work factors. The selected initial password cost follows the linked OWASP FIPS guidance.
- **Compatibility / migration:** existing password records and TOTP enrollment are incompatible; no conversion or automatic reset. Fresh deployment provisions credentials and enrollment under the new profile. Document authenticator compatibility and the intentional public enum break; do not add a SHA-1 fallback.
- **Unresolved risk / risk if left open:** authentication alone can prevent an application-wide claim. A permissive parser or unmeasured cost can expose CPU exhaustion; unsupported clients can lock users out.
- **Validation / evidence:** bootstrap, ordinary password creation, successful/failed login, MFA enrollment, replacement/reset and session revocation; independent PBKDF2/TOTP vectors and skew/replay/invalid-code cases. Reject malformed versions, lengths and work factors before expensive derivation. Measure representative concurrent authentication and record an explicit bounded resource/latency acceptance decision. Run Auth/bootstrap owner slices, relevant service-backed/browser rows, OpenAPI compatibility and generated/client checks.
- **Binary exit:** both credential creation paths use one policy; lifecycle/browser cases pass; malformed records reject cheaply; representative concurrency stays within the agreed bound; old formats fail explicitly and the enum/client projections agree.
- **Rollback:** fail fresh deployment qualification and correct the candidate; do not mix old authentication code with newly created regulated state.

### S19a — Maintain the fixed-container credential assessment

**Status:** DONE. **Owner:** module.auth / harness.command_surface. **Areas:** harness routing, isolated application fixture, performance evidence. **Discovered prerequisite for:** S19 completion, before S20.

- **Gap/remediation:** existing owner tests do not measure the real application under the frozen two-logical-CPU, 2-GiB/no-additional-swap profile. Add a Make-owned assessment using a pinned server/migrate build, disposable database and container, actual credential HTTP routes, and canonical identity/measurement/cleanup artifacts. Reuse the production facades and password service; add no production timing bypass or tuning flag.
- **Benefit/rationale:** makes the cost/capacity acceptance reproducible now and reusable for the final WSL2 package, without substituting a primitive benchmark for the application path.
- **Compatibility:** isolated intermediate state is destroyed; it cannot establish final-format/package compatibility. S23/S23a will advance the fixture's transport/configuration/admission alongside production; S24 repeats qualification on the final package.
- **Risk:** host load or Docker resource differences can invalidate latency evidence; capture actual Windows/WSL/Docker/CPU/container/module identities and effective limits. Do not relax frozen thresholds. Failure blocks S19 and S20.
- **Exit:** 20 synchronized ten-request bursts satisfy 2000-ms p95, 64-request overload rejects excess work within 250 ms, cancellation/recovery and existing budget invariants pass, and owned resource cleanup is confirmed. Record failures and original evidence identities.

### S19b — Preserve migration-owned identity during browser reset

**Status:** DONE. **Owner:** harness.browser (platform/harnessruntime implementation). **Areas:** disposable-fixture reset and regression tests. **Discovered prerequisite for:** S19 browser completion.

- **Gap/remediation:** service-backed Auth run `20261005T161754Z-p28248` reaches browser startup but its reset selects `application_crypto_format` as mutable application data and receives SQLSTATE `42501`. Preserve migration-owned identity metadata alongside the existing migration ledger/lineage metadata. Do not grant the recovery role write access or stamp a missing identity during reset.
- **Rationale/benefit:** fixture resets model ordinary operation on an admitted schema and retain the same security boundary as production; a reset cannot erase or upgrade format identity.
- **Compatibility:** disposable harness state only. Existing deployment admission, fresh initialization and applied migrations remain unchanged.
- **Risk/exit:** an incorrect exclusion can leave domain data behind or erase the startup marker. Test full domain reset with identity unchanged, absent identity remaining absent, narrow role grants intact, and real browser reset/startup passing. Owner discovery initially rejected the guessed `platform.harnessruntime` id; `harness.browser` is the catalog owner and its task guide was consulted. Record the failed run and cleanup before resuming S19; S20 remains blocked on S19 completion.

### S20 — Move secrets and short-lived tokens onto approved sealing

**Status:** DONE. **Owner:** Auth secrets, platform pagination, Revisions conflict tokens and Network Flow cursors. **Areas:** implementation, contracts and tests. **Gap:** RP-F24. **Depends on:** S19. **Next:** S21.

- **Remediation:** migrate all identified authentication-secret, pagination, conflict-token and Network Flow sealing callers to S18. Version changed sealed/token formats and explicitly reject older formats. Keep claims, associated-data encoding, owner/purpose labels, user/session/route binding, expiry, rotation and public error mapping inside the existing owners.
- **Boundary:** share cryptographic mechanics only. Retain authorization before/after decoding where required, bounded parsing and existing replay protections; do not let one owner's token become consumable by another. Review indirect derivation consumers such as Evidence upload-token signing without gratuitously replacing supported HMAC semantics.
- **Rationale / long-term benefit:** nonce and key-derivation fixes are maintained once while future token fields remain cohesive with the domain that interprets them.
- **Compatibility / migration:** old cursors/conflict tokens expire through explicit incompatibility rejection; clients follow existing restart/reload/error behavior. Persisted authentication secrets require fresh state. No dual reader, dual writer or compatibility alias.
- **Unresolved risk / risk if left open:** a less-visible token path can retain incompatible crypto; a shared implementation can accidentally erase AAD, actor/session binding, expiry or safe failure behavior.
- **Validation / evidence:** each owner tests positive round trips plus cross-user, cross-session and cross-route replay, wrong-purpose keys, tamper, expiry boundaries, rotation, malformed/oversized records and old-version rejection. Exercise public error mappings and owner lifecycle flows, not only primitive helpers. Use the affected owner/platform slices and service-backed rows selected through `make task-guide`.
- **Binary exit:** every identified caller uses S18 and every owner-specific binding/rejection scenario passes; no legacy sealing caller or unassigned key-derivation consumer remains.
- **Rollback:** keep the candidate unqualified until all affected callers agree; never deploy mixed format writers or restore legacy decoding as a tactical fix.

### S20a — Retire caller-controlled cursor entropy

**Status:** DONE. **Owner:** module.networkflow / harness.browser. **Areas:** Testing Harness owner, harness schema, fixture controls, implementation seam and tests. **Discovered prerequisite for:** S20's cursor migration.

- **Gap/remediation:** Testing Harness REQ-466–469 still promises a `network_flow.cursor_nonce` control and the module exposes `CursorNonceEntropy`. Remove that cryptographic control and its injection seam; retain table-ID determinism. Advance the active control response schema to v3 and retain v2 solely as historical evidence schema. Ordinary cursor nonce generation remains internal until S20 replaces the old codec with S18; no new nonce/entropy hook is introduced.
- **Rationale/benefit:** prevents a harness contract from forcing unsafe production capability into the shared primitive. Future cryptographic changes no longer carry deterministic-nonce compatibility.
- **Compatibility/risk:** old nonce control requests reject. Tests must assert opacity, binding, tampering and uniqueness rather than fixed encrypted bytes. Preserve the existing verification row identities and record the retired subcase explicitly.
- **Exit:** unauthorized/invalid/retired control requests reject; table-ID ordering, exhaustion and reset remain covered; cursor public behavior passes with no entropy option; normal schema generation/drift and owning checks pass. Resume S20 only after recording this checkpoint.

### S20b — Conceal Evidence storage locators in upload capabilities

**Status:** DONE. **Owner:** module.evidence. **Areas:** Core 01, implementation, owner token tests and route validation. **Discovered prerequisite for:** S20 closure.

- **Gap/remediation:** the indirect-consumer audit found that a signed base64 JSON upload capability carries `storage_key` verbatim, despite REQ-01-244 forbidding storage-key disclosure. Replace that member with a domain-separated HMAC binding, preserve the remaining claims and live lease/actor/session/method checks, and advance the internal token version to v3. Bound and canonically decode the token before expensive work.
- **Rationale/benefit:** keeps object-store addressing private while retaining the supported HMAC capability design and owner-local authorization. Shared cryptographic mechanics do not acquire Evidence payload rules.
- **Compatibility:** public upload-target shape remains unchanged. Existing v1/v2 capabilities reject; clients create a new slot/target using existing recovery behavior. No legacy reader or translation.
- **Risk/exit:** unresolved locator disclosure contradicts the owner contract. Verify no raw path in decoded public token bytes, exact HMAC binding, wrong-purpose/key rejection, malformed/oversized/old-format rejection, and real upload lifecycle authorization before resuming S20.

### S21 — Replace recovery encryption with one current format

**Status:** DONE. **Owner:** Recovery, with operator/application assembly and Core recovery requirements. **Areas:** specifications, implementation, contracts, tests and operator documentation. **Gap:** RP-F25. **Depends on:** S20. **Next:** S22.

- **Remediation:** introduce one current streaming backup envelope using independently derived per-artifact HKDF keys and module-generated nonces per chunk. Authenticate artifact identity, chunk position, length and finality; bound chunk sizes/counts and per-key use. Route small artifacts through the same codec and remove parallel legacy encryption paths from the new release.
- **Admission and journals:** update journal sealing through S18 and bind backup/restore/verification admission to the cryptographic profile. Preserve confined storage, export-root binding, freshness/overlap checks, bounded-memory streaming, interruption safety and validation before publication/target admission. Reject incompatible profile/envelope state before any mutation.
- **Rationale / long-term benefit:** one current codec prevents divergent integrity rules as artifact types and sizes grow, while actual recovery remains bounded and independently verifiable.
- **Compatibility / migration:** new encrypted-envelope versions are intentional breaks. Historical backup readers remain with historical releases; document matching release/backup retention and fresh-deployment boundaries. Keep applied SQL migrations immutable; coordinate any necessary new authored migration and typed schema changes with their owners.
- **Unresolved risk / risk if left open:** a backup can be created successfully yet violate module usage limits, lose finality, leak partial output or fail when needed. Old recovery proofs must not authorize a new profile.
- **Validation / evidence:** actual backup, due verification and restore through operator/package paths; small/large/empty boundary artifacts; truncation, reordering, duplication, substitution, length/finality corruption, wrong keys, interrupted operations and incompatible versions. Assert no publication, target admission or retained-state mutation on failure. Run Recovery owner/service-backed slices, shape/drift checks and operational recovery smoke.
- **Binary exit:** one current codec and journal policy serve all supported artifacts; actual restore and due verification pass; all adversarial/incompatibility cases reject before publication/admission, with bounded memory and accounted cleanup.
- **Rollback:** preserve historical deployments/backups with their matching releases. Do not add a legacy reader or run a converter to rescue an incomplete new format.

### S21a — Confined authenticated read staging

**Status:** DONE. **Owner:** platform.rootedfs and Recovery storage/application assembly. **Areas:** platform capability, Recovery storage port, implementation and adversarial tests. **Discovered prerequisite for:** S21 codec/publication completion.

- **Gap/remediation:** current `ReadArtifactStream` authenticates one open then reopens the path for output; `OpenRegular` returns a live file descriptor, not an immutable snapshot. Add a bounded, private, unlinked staging capability under the admitted filesystem root. Recovery decrypts and authenticates complete input into that capability, then copies those exact staged bytes to the caller; no second source open. No global temporary directory or extra shared lifecycle facade.
- **Rationale/benefit:** eliminates the read/verify/read race and keeps storage mechanics with rootedfs, artifact authentication with Recovery. The capability exposes no pathname and cannot publish failed plaintext.
- **Compatibility:** internal storage port requires staging support; unsupported backends reject, without a memory-unbounded or legacy fallback. On errors/cancellation the private staging descriptor closes, leaving no named artifact.
- **Risk/exit:** replacement or in-place mutation between reads must never expose unauthenticated output. Test root confinement/replacement, bounds, callback failure, cancellation, cleanup and storage-source mutation; verify only one source open and no destination use on authentication failure. Complete this prerequisite and checkpoint before resuming S21.

### S21b — Retire remaining historical Recovery readers

**Status:** DONE. **Owner:** module.recovery. **Areas:** implementation, fixture helpers, owner contracts and tests. **Discovered prerequisite for:** S21 completion.

- **Gap/remediation:** inspection found exported historical backup-capture and verification-v1 readers retained for test fixtures. Remove historical readers/writers from production; retain valuable negative cases through current entry points and move necessary test-only data construction to test support. Preserve original verification row identities while explicitly updating retired subcases.
- **Rationale/benefit:** one supported Recovery representation avoids a hidden compatibility burden or a route that could revive historical state.
- **Compatibility/risk:** historical backups still require matching releases. Current callers reject old schemas and cannot upgrade old evidence by changing a label/digest. Leaving dead legacy readers would contradict the fresh-release contract.
- **Exit:** production callers use only current envelopes/manifests/proofs; historical inputs reject before successful publication or target admission; owner, generation/boundary and real recovery lifecycle checks pass. Complete and checkpoint before S21 closes.

### S21c — Preserve migration metadata during verification reset

**Status:** DONE. **Owner:** module.recovery. **Areas:** target admission/reset implementation and lifecycle tests. **Discovered prerequisite for:** S21 lifecycle qualification.

- **Gap/remediation:** due verification still chooses target tables through a legacy hard-coded snapshot exclusion list. It attempts to truncate the migration-owned cryptographic-format singleton, which Recovery may only read. Use the frozen Recovery state catalog to validate table coverage and distinguish schema metadata from mutable target state for emptiness and disposable reset. Preserve the format record exactly, including absence; never create it during restore/reset.
- **Rationale/benefit:** one owner catalog governs new table classes, preventing future metadata additions from silently becoming restore payloads or reset targets.
- **Compatibility/risk:** disposable reset remains scoped to exclusively admitted verification targets. Ordinary restore cannot reset retained state. Failure to fix this blocks repeated verification and could erase or fabricate state identity if grants were broadened.
- **Exit:** real due-verification batches pass with current schema; populated/absent format markers survive reset, unknown table coverage rejects, effective-role constraints remain unchanged and cleanup passes. Checkpoint before resuming S21/S21b.

### S22 — Qualify the real Reference Pack verifier

**Status:** DONE. **Owner:** Reference Data and reference assembly. **Areas:** implementation verification, tests and contract review. **Gap:** RP-F26. **Depends on:** S21. **Next:** S23.

- **Remediation:** exercise bootstrap admission and the complete TUF verifier under the pinned module, including root/key rotation, signature thresholds and malformed inputs, replay, historical validation and destination portability. Use shipped execution paths and independently authored vectors, not merely a standalone Ed25519 call.
- **Preserved contract:** Ed25519, canonical bytes/digests, method identity, immutable provenance, trust history and destination trust remain unchanged. The owner amendment changes execution qualification, not signed-pack interpretation. Runtime projections remain separated from fixtures.
- **Rationale / long-term benefit:** the actual verifier receives credible qualification while valid packs and historical identities avoid unnecessary migration. Future module upgrades reuse these stable lifecycle and vector obligations.
- **Compatibility / migration:** valid canonical pack inputs retain identity; malformed keys/signatures and unsupported algorithms still reject. This does not authorize converting incompatible application databases or importing trust from a historical deployment automatically.
- **Unresolved risk / risk if left open:** passing primitive tests can miss another runtime verification path, signature-validation edge cases, or altered canonicalization after a build change.
- **Validation / evidence:** independent positive/negative vectors plus bootstrap/full TUF/rotation/replay/historical/destination scenarios through Reference Data owner/service-backed rows and packaged Reference Pack smoke. Compare valid canonical digests to preserved vectors; run targeted strict assessment through authored routing and explain any protocol-only findings.
- **Binary exit:** all actual verifier paths and lifecycle scenarios pass with the pinned actual module identity; valid canonical digests are unchanged, malformed inputs reject and no algorithm fallback exists.
- **Rollback:** block qualification and fix the verifier/module disposition; do not substitute algorithms or reissue pack identities to conceal an execution failure.

### S22a — Align Reference Data continuation with shared keyset pagination

**Status:** DONE. **Owner:** module.reference_data and platform.pagination. **Areas:** implementation and owner regression tests. **Discovered prerequisite for:** S22 completion.

- **Gap/remediation:** S20's current cursor decoder admits only offset/keyset modes, while Reference Data still writes a private `reference_pack_entry` mode. The resulting continuation rejects and breaks multi-page lookups. Use the existing shared keyset mode and make generic cursor write/read admission symmetric so an encoder cannot issue a cursor its decoder refuses; keep route, consumer, pack-set, pack, normalized lookup, limit, position and 900-second lifetime bindings in Reference Data.
- **Rationale/benefit:** removes an unnecessary platform-mode extension and keeps the generic mechanics separate from owner authorization and query semantics. Do not broaden the decoder with an owner-specific exception.
- **Compatibility/risk:** old private-mode cursors reject through the existing reload flow; fresh current cursors work. No pack identity, signed bytes, trust history or public consumer operation changes. Leaving the gap blocks every lookup that needs a continuation.
- **Exit:** unit continuation/binding/expiry and real all-profile pagination pass under the pinned module; shared cursor tests and boundary checks pass. Record failures/cleanup and checkpoint before resuming S22.

### S23 — Close transitive cryptographic and transport gaps

**Status:** DONE. **Owner:** existing platform PostgreSQL, object-store, enterprise-auth, TLS/telemetry, WebSocket and configuration adapters; security owner reviews service usage. **Areas:** platform adapters, configuration, specifications, tests and documentation. **Gap:** RP-F27. **Depends on:** S22. **Next:** S23v, then S23a.

- **Remediation:** map reachable PostgreSQL/SCRAM, S3 signing/checksum, OIDC/SAML, TLS, telemetry and WebSocket operations to the module services actually used, their parameters and permitted purpose. Review enabled optional features as part of the application boundary. Enforce approved algorithms/parameters at existing adapters while retaining vetted library protocol/signature verification.
- **Transport:** validate certificate chains, names and trust handling; reject insecure production transport settings. Specify application-side behavior when TLS termination is external, without claiming qualification of the terminator or external servers. Current HTTP and PostgreSQL `sslmode=disable` package settings need explicit disposition under the adopted policy.
- **WSL2 boundary:** keep application/private-key/service storage on Linux filesystems. Document Docker Desktop WSL2 integration and Windows-to-application HTTPS/WSS access with stable certificate names; do not assume guest and Windows trust stores are shared. Check published ports, proxy-to-app encryption and expired/untrusted/replaced certificates. Certificate provisioning, trust import and firewall changes are explicit operator procedures, not automatic global test changes.
- **Exceptions:** retain only explicitly reviewed non-security uses such as the WebSocket handshake hash. Treat indicator MD5/SHA-1 strings as data. Do not blanket-ban imports, wrap security calls to evade enforcement, or silently exempt a dependency. PBKDF2's password-storage service restrictions require purpose review for SCRAM; using a standard-library implementation is insufficient by itself.
- **Rationale / long-term benefit:** qualification follows actual behavior instead of library names or import counts. Adapter-owned policy keeps integration constraints cohesive and avoids hand-written replacements for complex protocols.
- **Compatibility / migration:** incompatible provider algorithms, credentials, certificates and transport settings fail explicitly. Operators configure supported providers/transports for fresh deployments; no legacy negotiation fallback is introduced.
- **Unresolved risk / risk if left open:** third-party defaults, optional checksums or provider negotiation can execute unqualified services. An unresolved supported-purpose mapping blocks qualification; do not assume the selected candidate can satisfy every dependency before this review.
- **Validation / evidence:** positive/negative integration execution for every enabled security path, including unsupported algorithms/parameters, malformed cryptographic inputs and rejected certificate/transport cases. Review dependency source with runtime evidence; targeted strict-mode assessment supplements both. Use platform/affected owner and service-backed slices, configuration checks and package scenarios.
- **Binary exit:** every enabled security service has a supported module-service/purpose mapping and passing positive/negative execution evidence; every protocol exception has a bounded non-security rationale; no unresolved usage or insecure production setting remains.
- **Rollback:** stop qualification for an unresolved required integration and resolve its supported design with the owner. Do not silently narrow the claimed boundary, disable a required feature or add a second crypto policy.

### S23b — Shared verified TLS construction

**Status:** DONE. **Owner:** platform.cryptography, with securefile and adapter callers. **Areas:** narrow transport mechanics and tests. **Discovered prerequisite for:** S23 adapter/fixture migration.

- **Gap/remediation:** current adapters independently construct insecure or TLS-1.2 transports. Supply one TLS-1.3 client/server configuration boundary with P-256/P-384 exchange, module-selected AES-GCM, verified chain/name checks, admitted certificate/key parameters, and no-follow bounded PEM reads. Callers retain endpoints, trust bindings, authentication and lifecycle; no provider framework or verification bypass is added.
- **Rationale/benefit:** qualify common transport mechanics once and prevent adapter defaults from silently diverging. Disable session tickets/resumption to avoid carrying another key lifecycle across certificate replacement; reconnection performs full current certificate validation.
- **Compatibility/risk:** unsupported protocol/certificate/key/path inputs reject; operator replacement takes effect on validated restart. Until adapters adopt this component, its completion does not establish secure application transport.
- **Exit:** pinned-module client/server and mutual-certificate handshakes pass; wrong names/roots, expired/future/mismatched/disallowed certificates, obsolete protocols, unsafe/oversized files and certificate replacement reject or succeed as specified. Authored verification routing, generation/drift and boundary checks pass; checkpoint before S23 integration resumes.

### S23c — Certificate-only PostgreSQL and disposable service fixtures

**Status:** DONE. **Owner:** platform.postgres and existing PostgreSQL/service harness owners. **Areas:** adapter, fixture/provisioning inputs, owner specifications and tests. **Discovered prerequisite for:** S23 integration completion and all application-running service evidence.

- **Gap/remediation:** driver parsing still consults inherited/default credentials and every shared PostgreSQL fixture currently supplies passwords/plaintext. Admit a closed explicit URI before driver parsing, reject inherited `PG*` inputs without interpreting values, load TLS files through S23b, suppress driver file/default fallback during controlled construction, and require `RequireAuth=none`. Preserve purpose-role checks and opaque diagnostics. Convert owned PostgreSQL fixtures/provisioning to isolated CA trust, distinct client identities and password-disabled logins; no replacement authentication protocol or legacy fixture mode.
- **Rationale/benefit:** production and service-backed owners exercise the same admission path. Removing password fixtures prevents later enforcement from hiding insecure dependencies behind injected handles.
- **Compatibility/risk:** fresh certificate-authenticated databases only; keyword/password/fallback DSNs reject. Fixture trust/private files remain owned disposable guest state with accountable cleanup; interrupted setup must clean its certificates as well as containers.
- **Exit:** runtime/migration/recovery through pool and SQL paths prove TLS 1.3, certificate identity and effective roles; wrong/expired/replaced certificates and prohibited server authentication fail before application work; parser/file/environment negative tests and representative migration/owner lifecycles pass. Generation, harness routing, docs and cleanup pass; checkpoint before other S23 integrations.

### S23d — Remove PUBLIC access to cryptographic-format metadata row type

**Status:** DONE. **Owner:** module.database_migrations, with PostgreSQL ACL validation. **Areas:** new migration, authored catalog inputs, generation and tests. **Discovered prerequisite for:** S23c completion.

- **Gap/remediation:** migration 64 restricts the singleton table but omitted the explicit PUBLIC row-type privilege revocation required for newly created tables. Add a forward-only privilege migration, preserving applied migration 64 and all historical digests. The metadata table remains migration-owned and read-only for runtime/Recovery; no new runtime grant is needed.
- **Rationale/benefit:** restore the established object-privilege invariant through the canonical migration lifecycle, rather than weakening the ACL check or adding fixture-only cleanup SQL.
- **Compatibility/risk:** fresh initialization includes the new migration; historical deployment compatibility remains rejected. Leaving PUBLIC type access violates the adopted PostgreSQL boundary and blocks owner acceptance.
- **Exit:** complete PostgreSQL ACL regression and migration/catalog checks pass, generated identities are current, cleanup succeeds and tracker is checkpointed before S23c resumes.

### S23e — Enterprise provider algorithm and transport admission

**Status:** DONE. **Owner:** module.auth's existing enterpriseauth adapter. **Areas:** adapter, provider admission, secure provider fixtures and verification routing. **Discovered prerequisite for:** S23 integration completion.

- **Gap/remediation:** production provider verification inherits broad library algorithm/key defaults and default HTTP clients. Use a verified TLS-only HTTP client for exchange/discovery/JWKS; constrain discovered endpoints and admit JWKS signing keys before library signature dispatch. Admit SAML RSA-SHA256/SHA-256 and RSA keys before vetted XML validation, rejecting encrypted assertions before decryption. Retain issuer, audience, nonce, correlation, time, subject and replay ownership.
- **Rationale/benefit:** a narrow admission layer around maintained protocol libraries avoids a replacement JWT/XML verifier and prevents future dependency defaults from widening policy. Route existing production-verifier tests explicitly, alongside new adversarial checks.
- **Compatibility/risk:** providers outside the fixed policy reject; no legacy algorithm or plaintext fallback. Private provider TLS trust remains operator-provisioned system trust; isolated tests supply local CA trust without global changes.
- **Exit:** real verified-HTTPS OIDC exchange/discovery/JWKS and signed SAML assertions pass under the pinned module; allowed algorithms, key rotation and certificate replacement pass; excluded keys/algorithms, insecure endpoints/redirects, encrypted assertions and identity/correlation failures reject before excluded operations. Owning route/lifecycle regressions, generation/drift, boundaries and cleanup pass before resuming S23.

### S23f — Admit RSA service parameters before verification

**Status:** DONE. **Owner:** platform.cryptography, with enterpriseauth and transport adapters. **Areas:** shared service admission, vetted-library adapters, trust construction, owner source and tests. **Discovered prerequisite for:** S23e/S23 integration completion.

- **Gap/remediation:** RSA size alone is incomplete admission; the selected archive additionally requires an odd modulus with even bit length and an odd exponent within 65537..2147483647. JOSE's default PSS verifier auto-detects salt length. TLS's default chain verifier runs before the application's current admission callback. Consolidate RSA parameter admission and expose fixed SHA-256 signature verification through supported JOSE hooks. For TLS, admit peer and trust-anchor parameters before calling standard x509 chain/name/EKU verification; bind the expected destination explicitly and preserve complete verification without an operator bypass.
- **Rationale/benefit:** accepted algorithms must imply accepted service parameters and ordering, rather than merely rejecting a connection after excluded cryptographic work. Retain standard library TLS, X.509 and JOSE parsing/signature semantics; do not implement a replacement wire protocol or general provider framework.
- **Compatibility/risk:** unsupported RSA parameters and non-32-byte PS256 salts reject. The selected Ubuntu/distroless guest profile supplies a bounded immutable system CA bundle; only policy-admitted anchors may enter chain construction. Trust/certificate replacement requires restart. No native-Linux scope is introduced.
- **Exit:** independent positive/adversarial RSA verification, PS256 salt-length rejection, and real TLS chains with excluded intermediate/anchor parameters prove rejection at the required boundary. Wrong chain/name/EKU/time and trust replacement remain enforced. Renew PostgreSQL and enterprise provider execution under the pinned module; routing, documentation and cleanup pass before S23e resumes.

### S23g — Qualify S3 operations and secure disposable fixtures

**Status:** DONE. **Owner:** platform.objectstore and existing object-store/service harness owners. **Areas:** SDK adapter, dependency inputs, fixture transport, contract tests and owner documentation. **Discovered prerequisite for:** S23 completion.

- **Gap/remediation:** MinIO bucket creation retains an unconditional MD5 path for non-default or corrected regions, and current fixtures use plaintext. Replace the operation binding with the maintained AWS SDK v2 S3 API/transfer manager under one explicit SigV4, SHA-256 and verified-TLS configuration; remove the former client after all production, fixture and probe callers migrate. Preserve adapter-owned errors, retries, bounded streams, purpose checks and bucket initialization ordering. Use SeaweedFS's native TLS-only listener and isolated process-local CA bundles, with accountable certificate/container cleanup.
- **Rationale/benefit:** using supported SDK operations removes reachable excluded mechanics without a dependency fork, fabricated checksum, response-rewriting guard or hand-written S3 protocol. One qualified adapter avoids two client policies and preserves future backend portability.
- **Compatibility/risk:** plaintext endpoints and unsigned checksum-less direct uploads reject. Presigned PUT requires the existing typed SHA-256 input and returns signed checksum headers; the public Evidence same-origin upload lease retains its own semantics. Bucket initialization and existing object keys remain supported; final qualification uses fresh state.
- **Exit:** known-length, unknown-length, multipart and presigned operations pass against the TLS-enabled reference backend. Wrong certificates, checksum/body substitution, expired presigns, adversarial region responses, cancellation and failed multipart cleanup reject without excluded operations. All owning contract/fixture/probe rows, read-only inspection, generation, boundaries and cleanup pass; backend limitations become separate blocking prerequisites.

### S23h — Supply the pinned runtime to build receipt generation

**Status:** DONE. **Owner:** harness build mechanics. **Areas:** authored Make build inputs and narrow build validation. **Discovered prerequisite for:** S23g and later public application builds.

- **Gap/remediation:** public binary recipes invoke S17's receipt-producing wrapper without exporting its required `NODE_BIN`; a successful compiler step can therefore be followed by a failing receipt step. Declare the pinned Node artifact prerequisite and pass its exact path to every affected cached build invocation. Preserve receipt production inside cache publication and receipt refresh on cache hits.
- **Rationale/benefit:** explicit tool dependencies work independently of inherited shell state and prevent a binary without its receipt from being cached as complete.
- **Compatibility/risk:** no runtime or product behavior changes. Leaving this unresolved breaks clean public builds and can make nested compiler summaries appear stronger than the enclosing command result.
- **Exit:** public ordinary and pinned migration builds, cached reuse and the existing build-wrapper harness row pass with receipts; toolchain/harness boundary checks and tracker update precede resuming S23g.

### S23i — Route existing S3 probe tests through the owner catalog

**Status:** DONE. **Owner:** harness.test_catalog / platform.objectstore. **Areas:** exact test-root admission, owner routing and generated topology. **Discovered prerequisite for:** S23g validation.

- **Gap/remediation:** the existing probe's CORS tests were outside the catalog's exact approved Go roots. Admit only `tools/objectstoreprobe` alongside the existing tool root and register its existing tests under the object-store owner. Preserve selector containment and complete test inventory checks; do not admit arbitrary tool packages.
- **Rationale/benefit:** the migrated probe receives reproducible verification through public Make routing instead of an untracked direct test command.
- **Compatibility/risk:** no product behavior changes. Unrouted probe tests could conceal SDK or TLS regressions. Narrow admission avoids weakening the harness's repository boundary.
- **Exit:** generation, exact probe row, catalog/boundary checks and cleanup pass; checkpoint before resuming S23g.

### S23j — Establish effective S3 checksum enforcement

**Status:** DONE. **Owner:** platform.objectstore / reference backend qualification. **Areas:** supported SDK/backend capabilities, negative integration evidence, configuration/dependency inputs if needed. **Discovered prerequisite for:** S23g completion.

- **Gap/remediation:** the first live SHA-256 run returned successful uploads but no retained checksum metadata, and accepted a changed body at a checksum-bound presigned URL. Determine the actual supported header/trailer representation and backend enforcement. Use a supported configuration or qualifying backend release; no weakened assertions, alternate hash, fabricated checksum, or response rewriting.
- **Rationale/benefit:** successful transfer alone cannot establish the integrity contract. Pin and qualify an interoperable supported service boundary before exposing its capabilities.
- **Compatibility/risk:** the reference backend must reject body substitution and preserve the required checksum semantics for empty, known, unknown-length and multipart uploads. An unsupported current backend blocks later work; no silent fallback or new TLS-only proxy.
- **Exit:** real positive and negative checksum/presign operations pass, including non-publication/cleanup; any changed pin passes normal generation/toolchain routes. Record the original failure and checkpoint before resuming S23g.

### S23k — Diagnose and recover early strict-mode fixture failure

**Status:** DONE. **Owner:** existing service harness. **Areas:** diagnostic routing and owned-resource recovery. **Discovered prerequisite for:** S23g completion.

- **Gap/remediation:** strict-mode service helper exited before start evidence and retained a pending runtime lease. Inspect fixture-only SHA-1 namespace hashing and testcontainers' Docker credential-cache MD5 separately from application security. Preserve a useful bounded failure classification and recover only proven owned resources using existing lifecycle machinery. Do not suppress prohibited application operations or equate strict-mode harness compatibility with production qualification.
- **Benefit/compatibility/risk:** actionable diagnostics and accounted cleanup prevent hidden leaks and misleading product claims; any disposable namespace change has no application-state migration. Existing failed evidence remains immutable.
- **Exit:** the cause and its scope are established, targeted diagnostics execute affected application operations under the pinned policy, and the retained run's owned resources are accounted for. Checkpoint before S23g closes.

### S23l — Route retained CORS proxy verification

**Status:** DONE. **Owner:** harness.test_catalog / platform.objectstore. **Areas:** exact test-root admission and owner routing. **Discovered prerequisite for:** S23g transport validation.

- **Gap:** the existing CORS proxy has behavior and process-ownership tests outside the public owner catalog. TLS changes must retain both semantics and safe process cleanup.
- **Remediation:** admit only `tools/s3corsproxy` alongside the existing exact helper roots and register its retained tests under Object Store. Generate normal topology; no broad tools-directory test admission.
- **Compatibility:** verification routing only. Existing plaintext tests establish historical proxy behavior, never positive cryptographic-policy evidence.
- **Exit:** exact proxy row, catalog contracts, JSON shape and generation drift pass; checkpoint before S23g migrates the proxy transport.

### S23m — Qualify telemetry exporter transport

**Status:** DONE. **Owner:** platform.telemetry, with platform.cryptography. **Areas:** exporter construction, endpoint admission, owner tests and dependency identity.

- **Gap/remediation:** OTLP currently permits plaintext and TLS 1.2 and inherits HTTP transport defaults. Admit only HTTPS, construct immutable shared-policy TLS before exporter activation, suppress proxy/redirect and insecure alternatives, and close owned HTTP resources after exporters. Qualify HTTP/gRPC traces, metrics and logs through real isolated collectors.
- **Identity:** remove the stale hard-coded exporter User-Agent version; derive each exporter identity from actual executable dependency metadata through S23n and retain existing safe header/secret handling.
- **Compatibility:** plaintext collectors and certificates outside Core 04 fail explicitly. Runtime system trust remains the only supported trust source; no custom telemetry trust/client-certificate configuration is introduced.
- **Exit:** all signal/transport paths pass positive and certificate/algorithm rejection checks under the pinned module, existing telemetry lifecycle/admission tests pass, generated routing and boundary checks pass, cleanup completes, and this tracker is checkpointed before the next S23 slice.

### S23n — Bind exporter identity to executable dependency metadata

**Status:** DONE. **Owner:** platform.telemetry. **Areas:** dependency identity and exporter option composition. **Discovered prerequisite for:** S23m.

- **Gap:** the selected metric HTTP module is v1.45.0 but its `Version()` reports 1.43.0; library version helpers are not reliable executable dependency evidence. Separate `WithDialOption` calls also replace, rather than append, application gRPC options.
- **Remediation:** derive per-signal/per-transport identity from actual executable dependency metadata, rejecting missing/replaced/ambiguous identities. Compose the supported gRPC options once so proxy rejection and application User-Agent coexist. No new version fallback or runtime override.
- **Compatibility:** correct diagnostic/protocol identification only; retained telemetry payload, privacy and retry rules stay owner-local.
- **Exit:** actual exporter requests carry their selected module identities; missing/replaced metadata rejects; transport qualification and cleanup pass. Checkpoint before completing S23m.

### S23o — Configuration v3 and application TLS admission

**Status:** DONE. **Owner:** platform.config / app.server / platform.httpruntime. **Areas:** typed configuration, listener assembly, configuration/process fixtures and owner validation.

- **Gap/remediation:** Core 04 requires v3, HTTPS and explicit listener certificate/key bindings, while the current projection admits v2/plaintext. Advance the closed schema, require normalized certificate/key paths and HTTPS origin, and admit the immutable server identity before runtime resource acquisition. Every ordinary/inherited network listener must serve qualified TLS; no plaintext fallback or external-termination exception.
- **Ownership:** configuration owns syntax; the server facade admits its transport identity; HTTP runtime owns listener/publication/shutdown. Shared cryptography owns certificate mechanics. Migrate/operator retain purpose-specific resources rather than acquiring a listener identity they do not use.
- **Fixtures:** isolate server identities and client trust through maintained test composition, preserve process/listener ownership and reverse cleanup, and keep pure handler tests distinct from network qualification.
- **Compatibility:** v1/v2 configuration and missing/invalid TLS bindings reject. Fresh configuration and provisioned certificates are required; no compatibility mode or retained-state rewrite.
- **Exit:** configuration/HTTP/server owner checks and actual HTTPS/WSS process startup, rejection, cancellation and cleanup pass; public generated/boundary checks pass. Browser/deployment composition prerequisites must be recorded separately if they block this boundary.

### S23q — Align remaining application unit fixtures with admitted credentials

**Status:** DONE. **Owner:** app.server / app.migrate / app.operator test composition. **Areas:** unit fixtures and verification. **Discovered prerequisite for:** S23o completion.

- **Gap:** broad backend unit routing reaches retained fake PostgreSQL DSNs and missing master-key bindings that now fail before the intended bootstrap/cleanup/operator transport assertions. Narrow earlier adapter checks did not execute these cases.
- **Remediation:** provide explicitly admitted syntactic certificate-DSN and exact-key fixture inputs at the unit composition boundary, preserve mocked connectivity and existing failure/lifecycle assertions, and retain real TLS qualification separately. Inspect any other affected assertions rather than weakening production admission or changing historical evidence.
- **Compatibility:** test composition only; no runtime fallback, altered authority or retained-state migration.
- **Exit:** exact server startup, migration facade and operator CLI/evidence rows pass, broader affected unit routing passes, and cleanup is accounted for before S23o resumes.

### S23r — Replace the audit test's removed Recovery API dependency

**Status:** DONE. **Owner:** platform.audit / Recovery state composition. **Areas:** owner integration regression.

- **Gap:** the audit package test still calls a snapshot-table helper moved to historical test-only Recovery support in S21; the production dependency no longer exists, preventing audit compilation.
- **Remediation:** assert authoritative backup membership through the current typed state catalog and owner contribution, preserving the audit retention claim without restoring a legacy codec API.
- **Compatibility:** tests only; no legacy reader, new production dependency or historical evidence rewrite.
- **Exit:** audit unit/service rows and boundary checks pass; current Recovery membership is asserted through its maintained owner port.

### S23p — Provision isolated HTTPS browser and development fixtures

**Status:** DONE. **Owner:** harness browser/service composition and development provisioning. **Areas:** authored fixture lifecycle, frontend proxy settings, local TLS material, browser trust and owner validation. **Discovered prerequisite for:** S23 completion, after S23o.

- **Gap:** browser startup still projects plaintext public origins and lacks application TLS bindings; the first broader Auth run correctly rejects it before readiness. The existing Vite/application/object-store paths must share explicitly provisioned fixture trust without disabling certificate verification.
- **Remediation:** provision confined, short-lived per-session certificates and explicit trusted HTTPS/WSS transports at existing boundaries; update origins, health probes and renewal/cleanup together. Keep host/global Windows trust unchanged. Development inputs require operator-provisioned identities rather than fixture or production fallback secrets.
- **Compatibility:** fresh configuration v3 and HTTPS browser URLs; obsolete plaintext fixture assumptions are removed. No production bypass or second policy.
- **Exit:** Auth browser, type/unit and owned web-e2e lifecycle checks execute through admitted HTTPS/WSS, negative trust cases reject, certificate/material and owned-service cleanup pass. Seeded-review lifecycle and development/package provisioning are separately owned by discovered prerequisites S23s/S23t and remain required before S23 closes.

### S23s — Carry isolated trust through seeded review sessions

**Status:** DONE. **Owner:** harness browser review/preparation lifecycle. **Areas:** fixture attachment, client trust, owned browser resources and cleanup. **Discovered prerequisite for:** S23 completion, after S23p.

- **Gap:** seeded review preparation performs Node requests in an already-running process, and its browser adapter launches local Chromium without the fixture's isolated trust database. Merely changing an environment variable after process startup does not establish client trust. The seeded TOTP helper still uses SHA-1, and its cleanup test asserts an obsolete single-cause representation instead of the current aggregate.
- **Remediation:** carry the admitted fixture CA capability through the existing request boundary; use the qualified owned renderer for seeded sessions and preserve exact resource proof, interruption/recovery and idempotent cleanup. Do not modify host trust or enable certificate-error bypasses. Keep externally selected developer review origins separate from canonical package evidence. Update the seeded TOTP consumer to the current exact 32-byte SHA-256 policy with independent vectors, and assert aggregate cleanup failures without dropping the primary failure.
- **Compatibility:** fresh HTTPS fixtures and their current attachment identity are required; no old attachment or plaintext production fallback.
- **Exit:** seeded review preparation, browser operation, interruption/recovery and cleanup pass with verified HTTPS and exact ownership; applicable review lifecycle and contract checks pass before S23 advances.

### S23t — Provision development and WSL2 package services with certificates

**Status:** DONE. **Owner:** deployment/dev-service composition and Recovery scheduling. **Areas:** authored Compose inputs, certificate/role provisioning, operator procedures and smoke support. **Discovered prerequisite for:** S23 completion, after S23s.

- **Gap:** the updated adapters reject retained development/package password DSNs and plaintext dependency defaults. Existing Compose examples and bootstrap scripts still project those obsolete inputs; guest scheduling also requires the selected Docker Desktop operation model.
- **Remediation:** provision distinct certificate-authenticated PostgreSQL purposes with password-disabled application roles, native SeaweedFS TLS, application TLS and explicit immutable trust bindings. Keep local database bootstrap isolated. Remove password and plaintext production alternatives; provision persisted owner keys explicitly. Reconcile operator-started scheduling and overdue backup/due-verification checks with Docker Desktop availability.
- **Compatibility:** fresh disposable engineering state and fresh final package state; old deployments and backups stay with matching releases. Do not reset the user's existing development containers or install host/Windows trust.
- **Exit:** authored service/configuration inputs admit all current application purposes, certificates can be replaced through documented owned restart, negative authentication/trust cases fail, scheduling and cleanup checks pass. Package acceptance remains S24; mandatory facade activation remains S23a.

### S23u — Prevent credential entry during session bootstrap

**Status:** DONE. **Owner:** Auth browser shell. **Areas:** frontend lifecycle and regression tests. **Discovered prerequisite for:** S23p, before its completion.

- **Gap:** the HTTPS/container Auth run exposed an existing timing race: the loading gateway accepts a username before the initial session response remounts the anonymous gateway, clearing it. The final click then fails local validation without sending a login request.
- **Remediation:** make credential fields unavailable while the initial session observation is pending; keep existing controller retirement and secret clearing. Add a delayed-session regression proving controls become usable only when the anonymous surface is ready.
- **Benefit / compatibility:** users cannot lose input across this mandatory initialization transition; no credential, API or stored-format change. Leaving the race creates intermittent failed sign-in and unreliable browser evidence.
- **Exit:** delayed-session unit assertions and the affected real HTTPS browser scenario pass; typechecking passes. Return to S23p after its tracker checkpoint.

### S23v — Expose genuinely fresh package storage to initialization

**Status:** DONE. **Owner:** package composition and build inputs. **Areas:** image context, shared storage bindings and package tests. **Discovered prerequisite for:** S23a, after S23t.

- **Gap:** the image currently copies `.keep` sentinel files into every authoritative filesystem volume, while S17 correctly treats every entry as retained state. The migration service also lacks the source object-store binding and shared filesystem mounts, so checking its private image paths would not inspect the deployment’s actual storage.
- **Remediation:** exclude source-control sentinels from the shipped runtime roots without weakening read-only freshness rules. Give migration exactly the source storage visibility required for read-only inspection, with the same bindings as the application and read-only filesystem mounts. Keep distinct PostgreSQL purpose credentials and nonroot ownership.
- **Benefit / compatibility / risk:** genuinely empty, correctly owned roots can initialize; retained shared content cannot be mistaken for empty private container paths. Fresh-only packaging changes; no deletion or compatibility stamping of historical state. Leaving this open either blocks every fresh launch or falsely admits retained storage.
- **Exit:** image roots contain no sentinel/content files, nonroot startup works, migration mounts resolve to the same authoritative volumes as the app with read-only access, and package smoke/cleanup pass. S23a then proves retained-state rejection through actual binaries.

### S23w — Initialize fresh test databases through state admission

**Status:** DONE. **Owner:** PostgreSQL test composition and test-service templates. **Areas:** reusable fixtures and tests. **Discovered prerequisite for:** S23a.

- **Gap:** fresh reusable database fixtures and suite templates currently call the low-level migration engine, leaving the format singleton empty. Mandatory facade admission would correctly reject them; adding an identity after fixture data creation would falsely stamp retained state.
- **Remediation:** use the migration owner's admitted initialization with distinct handles when creating genuinely fresh full-schema fixtures and templates. Keep historical migration-scratch capabilities on the low-level engine so they continue testing migration boundaries. Retain source-aware fixture build fingerprints; admitted template data carries the explicit format identity.
- **Benefit / compatibility / risk:** application tests exercise the same fresh-format gate as production without bypasses or post-hoc stamps. Disposable old template state rejects or is replaced only through existing owned fixture lifecycle.
- **Exit:** fixture ownership/lifecycle tests and service-backed application composition pass with the singleton established by admitted initialization; historical migration tests retain their coverage and owned resources are cleaned up. Checkpoint before resuming S23a.

### S23x — Distinguish inspection from writable-root admission

**Status:** DONE. **Owner:** configuration platform and application composition. **Areas:** configuration API, migration invocation, tests and owner clarification. **Discovered prerequisite for:** S23a.

- **Gap:** the startup filesystem validator requires write access. Migration now correctly receives authoritative volumes read-only, so applying the server's writable-root validation rejects every fresh packaged initialization before inspection.
- **Remediation:** expose an explicit read-only inspection validator for structurally admitted configuration. Share canonical root/type/overlap checks while requiring only readable searchable paths for inspection. Retain writable-root admission for serving and other owners that write; add no runtime policy toggle and create no probe files.
- **Benefit / compatibility / risk:** permission checks match each operation's capability. Migration cannot acquire unnecessary storage write access merely to satisfy an unrelated startup check. Without the fix, least-privilege packaging cannot initialize.
- **Exit:** read-only roots admit inspection, unsafe/overlapping/unreadable roots reject, no inspection creates state, existing writable startup negatives remain valid, and the fresh package smoke/cleanup passes. Checkpoint before resuming S23a.

### S23y — Retire staged credential assessment provisioning

**Status:** DONE. **Owner:** Auth capacity qualification and package harness. **Areas:** maintained assessment route, shared package fixtures and TLS client. **Discovered prerequisite for:** S23a.

- **Gap:** the S19 capacity assessment still provisions its intermediate password/plaintext database and HTTP listener. The now-mandatory facades correctly reject those inputs; leaving it would retain a second, unusable staging composition.
- **Remediation:** drive the existing assessment through the current shipped package and its certificate/fresh-state provisioning and cleanup. Use verified client TLS and record actual binary/image/host identities. Preserve two CPUs, 2 GiB, two active workflows, eight pending, 2,000-ms p95 and 250-ms excess-work rejection without tuning after measurement.
- **Benefit / compatibility / risk:** performance evidence exercises supported production composition; shared package ownership avoids divergent setup, credential and cleanup rules. Historical measurements retain their intermediate disposition and are never relabeled.
- **Exit:** the public assessment passes against the current package with exact effective limits, current admissions and TLS, records canonical measurements, and cleans all owned resources. Checkpoint before resuming S23a. S24 still renews final-candidate qualification.

### S23a — Activate mandatory application admission

**Owner:** build/toolchain and narrow platform cryptographic-policy boundary; all three application facades own invocation. **Areas:** implementation, build configuration, tests and documentation. **Gap:** RP-F21. **Status:** DONE. **Depends on:** S23, S23v, S23w, S23x, S23y and S17. **Next:** S24.

- **Activation:** wire the S17 components into server, migrate and operator; make default application builds and application-running harness paths use the pinned module; remove temporary staging machinery and stale ordinary application artifacts. RP-F21 closes only after this workstream passes.
- **Remediation:** pin `v1.0.0-c2097c7c` and its verified archive digest in authored build inputs, alongside the pinned toolchain. Include module selection, digest, compiler/build settings and relevant mode inputs in shared build cache identities, harness fingerprints and release artifacts. Cover supported server/migrate/operator builds and packaged binaries; do not rely on a label or caller-supplied version string.
- **Admission:** each exact application facade checks actual module version, resolved GOFIPS140 build metadata and enabled FIPS mode before acquiring application services. Then open only resources needed for read-only retained-state admission, rejecting incompatible state before migrations, bootstrap, leases, recovery journals, bucket creation, jobs or publication. Establish a separate application cryptographic-format identity only during admitted fresh initialization with race/interruption guards; never stamp an existing deployment as compatible. Runtime mode/environment overrides cannot downgrade the policy. Reject wrong module, absent/contradictory metadata and disabled mode with bounded safe diagnostics. Keep feature logic out of `cmd/*`.
- **Rationale / long-term benefit:** one authoritative policy closes gaps between source, cached binary, image and process. The platform boundary stays small and upgrades remain reviewable without a provider framework.
- **Compatibility / migration:** ordinary builds and `fips140=off` cannot run production application commands. Development/test composition supplies explicit supported inputs and owned dependencies; it does not add a production bypass. Preserve borrowed PostgreSQL/object-store ownership and idempotent cleanup.
- **Unresolved risk / risk if left open:** stale cache entries, local/global toolchain mismatch, spoofed metadata or production environment overrides can admit an unqualified binary. Module metadata must be verified using the pinned toolchain's supported interfaces.
- **Validation / evidence:** build/admission tests run actual binaries for correct/wrong/missing/disabled cases and assert no service acquisition or state mutation on rejection; exercise cross-policy cache reuse attempts. Run toolchain, boundary, harness and affected facade checks through discovered Make routes. Author the targeted strict-mode assessment route; do not invent or invoke an unregistered target.
- **Binary exit:** all three binaries prove the required actual identity and enabled mode; all negative cases and cache-contamination tests reject before side effects; canonical build/release artifacts carry matching identities.
- **Rollback:** revert an incomplete build/admission change as a unit and leave regulated production unavailable. Never restore an ordinary-build fallback.

### S24 — Qualify the Windows 11 / WSL2 package

**Owner:** deployment package, build/toolchain and Testing Harness/release owners. **Areas:** deployment package, harness, tests, release artifacts and documentation. **Gap:** RP-F28. **Status:** DONE. **Depends on:** S23a. **Next:** S25.

- **Remediation:** extend existing package qualifications to verify actual identities of server, migrate and operator; startup rejection; fresh-state admission; authentication/MFA; token flows; Reference Pack lifecycle; and actual recovery. Include wrong-module/disabled-mode/missing-metadata and incompatible retained-state/backup scenarios in disposable environments with rejection before mutation.
- **Environment and artifacts:** record Windows edition/build, WSL package/kernel, Ubuntu guest, CPU/features, effective VM and measured-container limits, Docker Desktop/backend/Engine/Compose versions, daemon kernel/context, storage filesystem, network mode, image index/resolved linux/amd64 image and binary/module/archive/build identities. Record the formal module-environment relationship as established only with supporting authority; otherwise explicitly unestablished. An enabled module or Windows/Linux policy entry is not sufficient. Use canonical package/release artifacts, not a new dossier.
- **Harness:** project the WSL2 required platform and indefinitely deferred native-Linux profile through existing authored task/owner/topology inputs and Make generation where applicable. Run the complete release matrix on WSL2, including all three smokes; do not substitute the former reduced WSL compatibility subset. Retain owned-resource cleanup and fail on incomplete cleanup. Keep strict diagnostics and formal CMVP claims separate from engineering acceptance.
- **WSL2 scenarios:** qualify Windows-to-application HTTPS/WSS and separate trust stores, guest/container service TLS, confined guest filesystem paths, certificate replacement and resource-constrained authentication. Exercise owned-service interruption/restart during jobs/recovery and clock/expiry rejection after a discontinuity; no partial backup or successful target admission may escape. Document host sleep, Windows restart and WSL shutdown as outages requiring restart/preflight. Never shut down the user's WSL VM or reboot Windows as routine smoke cleanup.
- **Rationale / long-term benefit:** qualify the exact package operators receive and reuse the established release system as features grow. Source-only tests cannot detect packaging, cache, configuration or runtime overrides.
- **Compatibility / migration:** operators receive explicit fresh-deployment guidance and preflight rejection for incompatible installations. No automatic rollout, reset, conversion or retained deployment operation.
- **Unresolved risk / risk if left open:** guest/daemon identity can differ; Windows updates, shared CPU contention, network changes and trust-store assumptions can make local success irreproducible. Native-Linux evidence cannot stand in for WSL2 acceptance; a Debian image cannot stand in for the observed kernel.
- **Validation / evidence:** `make standup-package-smoke`, `make standup-reference-pack-smoke` and `make standup-operational-recovery-smoke` on the recorded WSL2 target, with the new cases, identity/cache negatives, applicable harness/shape/drift/toolchain checks and canonical review. Record every failed WSL2 prerequisite. Actual Windows client-path evidence must be distinguishable from guest-browser tests; extend existing routes before execution rather than inventing a second test system.
- **Binary exit:** all package scenarios, WSL2 boundary checks, fixed performance criteria and cleanup pass with verified identities and required release routing. Formal environment applicability has an honest separate disposition; unestablished formal applicability prevents a regulated claim, not a WSL2 engineering pass. Native Linux remains deferred and outside this exit.
- **Rollback:** keep release acceptance/promotion blocked until the package and routing agree. Never mark helper or partial results as completed package qualification.

### S24a — SeaweedFS identity and readiness across service replacement

**Status:** DONE. **Owner:** deployment package and object-storage fixture owners. **Areas:** authored service configuration, package validation and operator documentation. **Discovered prerequisite for:** S24.

- **Gap:** the renewal scenario reaches renewed PostgreSQL credentials but S3 initialization returns `retry_exhausted` after service recreation (`20261005T235956Z-p98004`). The package leaves SeaweedFS's node identity at its detected container address. The tagged 4.48 server source uses that value as the master identity and resumes persisted Raft state; concurrently recreated service addresses are not stable.
- **Remediation:** explicitly bind the single packaged node to its stable Compose DNS identity, with an explicit listener binding. Preserve the existing persistent volume and TLS-only S3 policy. Qualify replacement against already-created storage; do not reset Raft state, erase data or add a retry/fallback that conceals the failure. Confirm the source-based diagnosis through actual renewed service execution.
- **Benefit / compatibility:** service identity survives certificate replacement and container recreation. This is part of the fresh deployment candidate, not a conversion of existing user development volumes. Isolated one-lifetime fixtures use disposable state and are never presented as retained-state restart evidence.
- **Risk if unresolved:** renewed or recreated service containers may be unable to read retained storage, making certificate maintenance and restart unreliable.
- **Exit:** the actual package's purpose rejection, renewed migration/object-store credentials, certificate replacement and owned restart pass with the same database/volume identity and successful cleanup. Checkpoint before resuming S24. Source: https://github.com/seaweedfs/seaweedfs/blob/4.48/weed/command/server.go and matching master.go.

### S25 — Final validation and handoff completion

**Owner:** release/validation and handoff owners, with affected specification and operational owners. **Areas:** validation, documentation and this tracker. **Gap:** RP-F29. **Status:** DONE. **Depends on:** S24 and all active implementation/prerequisite slices; excludes indefinitely deferred native Linux and the separate formal deployment-claim decision. **Next:** none; **must finish last**.

- **Remediation:** finish focused owner and changed-input checks, run `make agent-finalize` before broader final verification, then `make test-fast` and one complete `make release-check` on the final candidate. Inspect canonical results, readiness projections, actual binary/module/environment identities and cleanup. A changed candidate or failing release needs its own corrected final validation; never combine partial runs into a pass.
- **Handoff:** review fresh-deployment and incompatible-state handling, matching-release backup retention, operational failure diagnostics, key provisioning/rotation, integration limits and recovery procedures. Document the module-update process: choose an exact candidate, recheck certificate/security status and environment support, amend pins/cache identities, rerun service/verifier/package qualification and record ordinary review. Future updates are deliberate, never floating-selector refreshes.
- **WSL2 operations:** hand off Windows/WSL/Docker versions and update revalidation, Ubuntu setup, guest filesystem/volume placement, certificate provisioning and renewal for Windows and Linux clients, HTTPS/WSS access, restart after sleep/shutdown, clock checks, interrupted-work recovery and cleanup. Retain backups with their matching application releases. Include the indefinitely deferred native-Linux scope and formal CMVP claim limitation; neither requires a parallel package or a fallback policy.
- **Rationale / long-term benefit:** a reproducible release boundary and actionable maintenance instructions survive beyond the current authors without multiplying signoff artifacts.
- **Compatibility / migration:** no customer rollout or retained-state mutation is automatic. Distinguish implementation completion, WSL2 package acceptance, formal CMVP environment applicability, specification adoption and customer deployment approval. Missing deployment/adoption acceptance blocks the corresponding promotion; missing formal applicability blocks regulated claims. Neither indefinite native-Linux deferral nor an unestablished formal claim prevents completion of an otherwise passing WSL2 engineering iteration.
- **Unresolved risk / risk if left open:** partial, historical or stale evidence can be mistaken for readiness; unresolved security disposition or environment assumptions can be lost during handoff.
- **Validation / evidence:** focused owner/drift/OpenAPI/boundary/toolchain/harness checks as needed; finalization, fast checks and one complete final release. Use `make explain-run RESULTS_DIR=<run-root>` for canonical inspection. Record exact source/graph/run identities, failure history, cleanup, skipped checks and their reasons. If retained-run maintenance is intended, use the permitted successful full warm-check `RESULTS_DIR`; otherwise report it skipped because unset.
- **Binary exit:** S16–S24, S16c, S23a and every active implementation prerequisite are DONE; one complete final-candidate WSL2 release and cleanup pass; canonical evidence, WSL2 operating guidance and all five readiness dispositions are recorded. Native Linux stays DEFERRED indefinitely; formal environment applicability stays unclaimed unless independently established. Mark S25 DONE last and close S07 for the selected scope without claiming formal certification. No implementation, approved-service or WSL2 package failure may be moved into the deferred scope.
- **Rollback:** leave S25 incomplete or BLOCKED with the exact failed boundary and retained artifacts. Do not relabel the prior release, erase failures or declare promotion from an incomplete candidate.

## 8. Validation plan

### S16 service feasibility and downstream projection review

This preserves S16 source/design feasibility evidence, distinct from completed
S23 execution and S24 package qualification. The original MinIO design row is
historical and was superseded by S23g: the current adapter uses AWS SDK v2,
SigV4 and explicit SHA-256 checksums, with qualified known-length, streaming,
multipart and presigned operations against SeaweedFS 4.48. Its actual adapter
and package receipts, not the earlier source inference, establish engineering
acceptance. The selected module archive is unchanged. Its frozen
PBKDF2 implementation marks its internal HMAC as KDF use; local password rules
need not be changed merely because standalone HMAC has a minimum key length.

| Service / source boundary inspected | Supported design / projection | Required execution route |
| --- | --- | --- |
| Go 1.27.1 frozen module and standard cryptographic APIs | Exact archive digest verified; driver regression compiles/runs with the selected GOFIPS140 and enabled mode. `crypto/fips140.Version`, `Enabled` and actual build settings supply separate admission inputs. | S17 actual three-binary metadata/mode/cache rejection; S18 primitive vectors. Development compilation is not environment qualification. |
| PostgreSQL / pgx v5.11.0 | `RequireAuth=none` guards mechanisms before dispatch; full TLS plus separate certificate-to-login and effective-role checks are still required. The driver's connection-string allowed-key capability can reject unsupported options before default expansion; inherited PG defaults must separately reject. | S16a DONE; S23 certificate/chain/name/purpose and prohibited-mechanism execution; preserve existing platform.postgres rows. |
| Passwords and MFA / Authn, bootstrap, public Auth owner | Standard PBKDF2 parameters and versioned record, 2 active/8 pending/2000-ms queue; SHA256 TOTP32 bytes; no legacy fallback. Core 01 owns overload, Core 04 owns capacity/reference acceptance. | S19 Auth/Bootstrap owner, frontend and browser routes; OpenAPI owner source, generated clients, compatibility change set and fixtures. |
| Owner secrets/tokens and Recovery | Standard HKDF plus `NewGCMWithRandomNonce`; per-message salt, per-artifact salt and 1048576-chunk bound; context/nonce mechanics shared, payloads/keys/rotation local. | S18–S21 owner tests, schema/fixture generation; exact affected IDs described in Core 01/Core 04; real restore and no-publication negatives. |
| S3 / MinIO v7.0.100 actual adapter calls | Static SigV4 uses standard HMAC/SHA256. `TrailingHeaders` plus explicit `PutObjectOptions.Checksum=ChecksumSHA256` disables Content-MD5 for known/unknown-length and multipart writes. Current adapters use single-object removal and do not invoke MD5-based multi-delete, bucket-CORS/versioning/encryption configuration. Qualify the actual server's required checksum support, never silently fall back. | S23 platform.objectstore/SeaweedFS compatibility including upload sizes, presigning, probes, initialization and deletion over TLS; S24 packaged object operations. |
| OIDC / go-oidc and go-jose | SupportedSigningAlgs and vetted verifier/key-set boundaries allow the fixed RS256/PS256/ES256 policy; validate key type/size before verification and preserve issuer/audience/nonce/time checks. | S23 enterprise-auth positive/negative provider tests; reject unsupported algorithm/key material before dispatch. |
| SAML / crewjam and goxmldsig | ServiceProvider.SignatureVerifier provides a supported admission wrapper around the library ValidationContext; inspect signature/digest parameters and provisioned keys before library verification. Library retains namespace/canonicalization and signature checks. No protocol replacement. | S23 real signed response, algorithm/key rejection, issuer/audience/request/time/replay negatives. |
| TLS / application, PostgreSQL, object store, OTLP | Standard crypto/tls with fixed TLS 1.3, supported curves/suites and verified certificates; application PEM bindings require configuration v3. OTLP retains system trust-store ownership and loses insecure HTTP. | S23 transport/adapters and negative certificates; S24 proxy-to-app and packaged dependency hops, provisioning/rotation. |
| Reference Pack and WebSockets | Ed25519 and canonical TUF semantics unchanged. WebSocket SHA-1 is only the RFC 6455 handshake protocol exception; it has no authentication or payload-integrity role. | S22 actual verifier lifecycle and stable digests; S23 handshake/origin/session/subscription/TLS and diagnostic disposition. |
| Reference operating environment | Windows 11 Pro build 26200.9457; WSL 2.7.3.0, guest kernel 6.6.114.1-microsoft-standard-WSL2; Ubuntu 26.04 LTS x86_64; Intel Core Ultra 9 285HX; Docker Desktop Engine 29.5.2 / Compose 5.1.3, daemon on the same kernel, 24 visible CPUs and 33353146368 memory bytes. Repository backing filesystem is ext4. S24 receipts now record Desktop 4.75.0.227598, actual daemon identity, ext4/NAT and effective resource limits; the original S16 observation remains distinct. | S16b uses observed host/guest/daemon identity and passing doctor; S24 records final identities and runs the package; S25 runs the complete release on WSL2. Module policy review does not establish formal CMVP applicability to this stack. Native Linux is not a dependency. |

Relevant authored projections for follow-on slices are `tools/toolchain_pins.json`,
Make build/cache inputs, harness owner/topology inputs, `internal/platform/config/`
and `internal/testutil/fixtures/config/`,
`contracts/openapi-source/owners/module.auth/openapi.json`, Recovery schemas and
fixtures, owner-local token/envelope contracts, and deployment inputs. Core 04's
current configuration projection is the authored Go schema/catalog and fixtures;
there is no `contracts/configuration/` family. No generator may discover these
inputs by reading this table or other Markdown.

Use public Make targets from the repository root and choose the narrowest owner rows first. `make task-guide ROLE=module-author OWNER=<owner-id>`, `make help` and `make help-all` remain the live navigation surface; this table records the relevant discovered routes, not a second command registry. Test routing does not define requirements or establish specification completeness.

| Layer / workstreams | Discovered route | Required timing and evidence |
| --- | --- | --- |
| Specification/tracker edits | `make lint-markdown`; `git diff --check` | At workstream checkpoints; preserve the exact 197-file register and historical text/run identities. Product validation follows the changed implementation owners. |
| Owner unit/behavior checks, S18–S23 | `make task-guide ROLE=module-author OWNER=module.auth`, `OWNER=module.recovery`, `OWNER=module.reference_data`; then `make test-slice OWNER=<owner-id> [ROWS=<discovered-row-ids>]` | Guides for these three owners were inspected successfully during planning. Discover affected Revisions, Network Flow, bootstrap, configuration and other platform row selection when their slices begin; do not invent semantic row IDs. |
| Integration/service-backed checks, S17–S24 | `make service-backed-test-slice OWNER=<owner-id> [ROWS=<discovered-row-ids>]` after the owning task guide | Exercise real persistence, provider, auth/session, verification and recovery boundaries where reached. Keep isolated services and accountable cleanup. |
| Auth browser/client projection, S19/S20 | Auth task guide and its routed browser rows; `make frontend-typecheck` and `make frontend-unit` when their inputs change | Real enrollment/login/reset/session flows with the new enum and formats. No unsolicited UI redesign or golden refresh. Discover exact browser row selection from authored routing at execution time. |
| Authored contracts and generated outputs | `make generate`, `make generate-drift`, `make json-shape-check`, `make generated-artifact-policy-check` | Run only when relevant owner inputs change during implementation. Preserve generated-root ownership and runtime/test separation; never use Markdown as a generator/test input. |
| Public API compatibility, S19 | `make openapi-compatibility-check` | Resolve the intentional algorithm-enum change through the normal compatibility process; a known break is not silently ignored or hidden behind legacy support. |
| Owner/module boundaries and pins | `make backend-module-boundary-check`; `make toolchain-drift` | Required when the platform primitive/policy boundary, imports or build pins change. Use normal frontend boundary routing if clients change. |
| Harness and release routing, S17/S24 | `make harness-contract`; `make explain-target TARGET=release-check DETAIL=summary` | Verify authored routes, cache/fingerprint inputs, release membership and generated topology. Explain-target is discovery, not execution evidence. |
| Targeted strict-mode assessment, S17/S22/S23/S24 | `make cryptographic-policy-assessment` (authored and validated in S17); require final package identities in S24. | S17 isolated build/assessment results are component evidence. Pin the module, select relevant security paths and retain actionable diagnostic findings. `fips140=only` is diagnostic, not deployment mode or sole compliance criterion; expected protocol-only findings are reviewed, not suppressed through bypasses. |
| Actual package acceptance, S21/S22/S24 | `make standup-package-smoke`; `make standup-reference-pack-smoke`; `make standup-operational-recovery-smoke` | Extend the existing routed smokes. Verify all three real binaries, fresh-state/negative admission, auth/tokens, pack lifecycle and actual recovery on the recorded reference target, including cleanup. |
| Final candidate, S25 last | Focused checks, then `make agent-finalize`, `make test-fast` and one complete `make release-check` | Finalization precedes broader end-of-run verification. Record source/module/image/environment identities and all failures. No merging partial runs or substituting historical release evidence. |
| Canonical handoff review | `make explain-run RESULTS_DIR=<run-root>`; inspect canonical summary/manifest, package artifacts and cleanup receipts | Require complete passing results and correct identities. Distinguish retained-run maintenance from read-only result inspection. `RESULTS_DIR` for finalization must meet the repository's retained-run requirements or maintenance is explicitly skipped. |

Current `lint-markdown` configuration selects authored documentation globs but does not directly include this handoff path. Run the required public target and report its configured scope honestly; supplement it with read-only tracker structure, inventory, history and diff checks. The changed owner documents are included by `docs/*.md` and `docs/spec/**/*.md`; this controlling handoff is not directly linted. Do not change lint configuration merely to broaden coverage.

## 9. Top-level work tracker

Historical DONE rows below retain their original outcomes; their validation does not qualify S16–S25. Current execution statuses are explicit below.

| ID | Work item / workstream | Status | Depends on | Evidence / disposition | Binary exit |
| --- | --- | --- | --- | --- | --- |
| RP-P02 | Prior package-plan document update | DONE | Prior user-selected plan | Markdown `20261005T042618Z-p49665` and original session retained below | Prior document-only step completed. |
| RP-P03 | Regulated-readiness document update | DONE | Current document-only request | Markdown `20261005T122414Z-p70033` passed for its configured scope; diff, tracker structure, 197-file inventory, history and 6,684-path scope checks passed | Only this tracker changed; S16–S25 remain TODO and require later implementation authorization. |
| RP-P04 | Windows 11 / WSL2 plan and owner-source refocus | DONE | User target change, 2026-10-05 | Markdown `20261005T134653Z-p81528`, diff and supplemental history/inventory review passed; §10 checkpoint | Consistent active target, indefinite native-Linux deferral, WSL2 acceptance and separate formal-claim disposition; earlier implementation preserved. |
| RP-S01 | Reconcile owner/tracker authority | DONE | Prior iteration | Original execution retained below | Prior slice exit passed. |
| RP-S02 | Remove dead interfaces/misleading fixtures | DONE | S01 | Original execution retained below | Prior slice exit passed. |
| RP-S03 | Stable owner DTOs/private verification | DONE | S02 | Original execution retained below | Prior slice exit passed. |
| RP-S04 | Complete application construction | DONE | S03 | Original execution retained below | Prior slice exit passed. |
| RP-S05 | Separate transport and semantics | DONE | S04 | Original execution retained below | Prior slice exit passed. |
| RP-S06 | Persistence seams and instance isolation | DONE | S05 | Original execution retained below | Prior slice exit passed. |
| RP-S07 | Pinned-module cryptography and WSL2 acceptance umbrella | DONE | S16–S25 | S25 completed last; final release `20261006T091921Z-p85023` passes 1,279/1,279 and all 673 cleanup steps | WSL2 engineering scope closed; native Linux remains indefinitely deferred, formal CMVP applicability unestablished, adoption and customer approval separate. |
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
| RP-S16 | Define cryptographic policy/amend owners | DONE | Authorized implementation; S16a/S16b DONE | RP-F20/F29; coordinated source/service and downstream projection review; Markdown and diff checks pass; §10 checkpoint | Consistent candidate, service/exception owners and routes; WSL2 feasibility established. Runtime/adoption claims remain separate; next S17. |
| RP-S16a | Supported PostgreSQL authentication admission | DONE | Discovered during S16 | pgx v5.11.0 upgrade, real challenge regression, operator fixture repair; owner/build/fast/vulnerability/finalization checks passed; §10 retains failures and identities | Supported driver capability proven; production certificate policy remains S23 work. |
| RP-S16b | Windows 11 / WSL2 environment feasibility | DONE | User retarget, 2026-10-05 | Host/guest/daemon inspected; doctor `20261005T134056Z-p77862` and Markdown `20261005T134653Z-p81528` passed; coordinated source review complete | WSL2 identities, storage/resource/transport assumptions and full validation route established; formal applicability unestablished. S24 remains unexecuted. |
| RP-ENV-NATIVE | Pure/native Linux deployment and qualification | DEFERRED | Explicit future user scope decision only | Indefinite deferral; no deadline or active release dependency; no new native-host evidence required | Reactivate only under a separately scoped future plan; not part of S25/S07 engineering completion. |
| RP-S16c | Remaining source consistency | DONE | S16 | Markdown `20261005T143313Z-p459`, diff/document review passed; checkpoint below | Consistent requirements/criteria and document checks. |
| RP-S17 | Build identity/admission components | DONE | S16c | RP-F21 remains open; §7 | Isolated component/candidate admission, concurrency, interruption and cache tests pass. |
| RP-S17a | Generated Node readiness consolidation | DONE | Discovered during S17 | Generation `20261005T144847Z-p19036` exceeded existing binding-size limit | Generator/command-surface checks pass without budget increase; then resume S17. |
| RP-S17b | Migration discovery admission boundary | DONE | Discovered during S17 | Failed focused DB run `20261005T145239Z-p59016`; provider source confirms early ledger mutation | Full provider discovery/application enclosed by admission; lifecycle negatives pass. |
| RP-S18 | Approved derivation/sealing primitives | DONE | S17 | RP-F22; §7; primitive completion evidence below | Exact keys, separated derivation and random-nonce primitive pass misuse/use-bound tests. |
| RP-S19 | Password/MFA cryptography | DONE | S18 | RP-F23; S19/S19a/S19b completion evidence below | One credential policy, SHA-256 TOTP, lifecycle/client and resource-bound checks pass. |
| RP-S19a | Fixed-container credential capacity | DONE | Discovered during S19 | `20261005T163057Z-p36924`; exact profile, latency and cleanup recorded | Frozen profile and overload thresholds pass. |
| RP-S19b | Preserve format identity during reset | DONE | Discovered during S19 | `20261005T162310Z-p95444` / `20261005T162403Z-p28456` | Present/absent marker and narrow grants preserved. |
| RP-S20a | Retire cursor entropy injection | DONE | Discovered during S20 | `20261005T164118Z-p71795`; original row identity retained | Only non-cryptographic table-ID determinism remains. |
| RP-S20b | Conceal Evidence storage locators | DONE | Discovered during S20 | REQ-01-244 versus upload-token payload | Opaque HMAC binding and current-version lifecycle pass. |
| RP-S20 | Secrets and short-lived tokens | DONE | S19 | RP-F24; §7; owner migration/validation checkpoints below | All caller migrations and owner-specific binding/rejection tests pass. |
| RP-S21a | Confined authenticated read staging | DONE | Discovered during S21 | Live reader/two-open inspection; no qualification yet | One source read, bounded private staging and failure cleanup pass. |
| RP-S21b | Retire historical Recovery readers | DONE | Discovered during S21 | Historical fixture-only reader/writer inventory | Production current-only formats and negative legacy tests pass. |
| RP-S21c | Preserve reset schema metadata | DONE | Discovered during S21 | Failed due-verification runs below; readonly singleton is in legacy truncate set | Catalog-driven target checks/reset and due lifecycle pass. |
| RP-S21 | One current recovery encryption format | DONE | S20 | RP-F25; §7; no new execution evidence | Real restore/due verification and negative framing/admission tests pass. |
| RP-S22a | Reference Data shared keyset continuation | DONE | Discovered during S22 | Current cursor mode mismatch in failed consumer rows | Bound/expiring continuations pass through real consumers. |
| RP-S22 | Actual Reference Pack verifier qualification | DONE | S21 and S22a | RP-F26; pinned owner and strict runs below | Shipped verifier/lifecycle pass with unchanged canonical identities and no fallback. |
| RP-S23b | Shared verified TLS construction | DONE | Discovered during S23 | Divergent transport defaults; new execution pending | Shared secure transport and adversarial handshake/path checks pass. |
| RP-S23d | Cryptographic metadata row-type privilege | DONE | Discovered during S23c | PostgreSQL ACL failure in `20261005T185159Z-p88332` | New immutable migration restores zero PUBLIC application type grants. |
| RP-S23f | RSA parameters and TLS verification ordering | DONE | Discovered during S23e | Automatic PSS salt and post-verification TLS admission | Vetted dispatch receives only admitted parameters; full certificate verification remains mandatory. |
| RP-S23e | Enterprise provider algorithm and transport admission | DONE | Discovered during S23 | Broad provider defaults and unrouted production verifier checks | Approved library dispatch, secure fixtures and owning lifecycle evidence pass. |
| RP-S23h | Explicit build-receipt runtime binding | DONE | Discovered during S23g | Public Make wrapper omits NODE_BIN | Ordinary/pinned builds and cache reuse produce verified receipts. |
| RP-S23i | Exact S3 probe test routing | DONE | Discovered during S23g | Existing probe package outside exact allowed roots | Generated routing and probe/catalog checks pass. |
| RP-S23j | Effective S3 checksum enforcement | DONE | Discovered during S23g | Successful write accepts a substituted body | Qualified backend/request representation and negative checks. |
| RP-S23k | Strict fixture diagnostics and recovery | DONE | Discovered during S23g | Helper exits before evidence and cleanup fails | Classified diagnostic and proven owned cleanup. |
| RP-S23l | Retained CORS proxy verification | DONE | Discovered during S23g | Existing helper tests outside public catalog | Exact routing and ownership regressions pass. |
| RP-S23n | Exporter executable dependency identity | DONE | Discovered during S23m | Stale upstream Version API and replaced gRPC options | Actual module identity and composed options pass. |
| RP-S23o | Configuration v3 and application TLS | DONE | S23m | v2/plaintext projection contradicts Core 04 | Config, listener and real process admission pass. |
| RP-S23q | Admitted application unit fixtures | DONE | Discovered during S23o | Old fake DSNs, missing explicit key and stale current evidence fixture | Server/migrate/operator assertions reach intended boundaries. |
| RP-S23r | Audit Recovery contribution regression | DONE | Discovered during S23o | Audit test calls removed legacy snapshot helper | Current owner contribution and audit persistence assertions pass. |
| RP-S23u | Auth bootstrap credential admission | DONE | Discovered during S23p | Input accepted before anonymous gateway remount | Delayed-session and HTTPS browser regressions pass. |
| RP-S23p | Isolated HTTPS browser/development fixtures | DONE | S23o | Browser public origins and transports remain plaintext | Verified HTTPS/WSS, browser ownership and cleanup pass. |
| RP-S23s | Seeded review trust and ownership | DONE | S23p | Existing review clients do not acquire fixture trust | Verified seeded review and exact cleanup pass. |
| RP-S23t | Certificate provisioning and WSL2 scheduling | DONE | S23s | Compose/bootstrap defaults retain password/plaintext paths | Current purposes and operator-started overdue checks pass. |
| RP-S23m | Telemetry exporter transport | DONE | Discovered during S23 | Plaintext/TLS 1.2 and stale exporter identity | All signal/transport paths and lifecycle checks pass. |
| RP-S23g | Qualified S3 operations and secure fixtures | DONE | Discovered during S23 | MinIO bucket-creation MD5 retry and plaintext fixtures | One supported SDK policy and live SHA-256/TLS operations with cleanup. |
| RP-S23c | Certificate-only PostgreSQL and fixtures | DONE | S23b; discovered during S23 | Existing password/plaintext fixture coupling | Real purpose identities, rejected auth, secure fixtures and cleanup pass. |
| RP-S23 | Transitive crypto/transport qualification | DONE | S22 and S23b–S23u | RP-F27; owner and package checkpoints below | Every enabled security path has a supported service mapping and positive/negative evidence. |
| RP-S23v | Genuine fresh package roots and inspection bindings | DONE | S23t | Image sentinels and absent migration storage visibility | Empty owned roots and identical read-only source bindings pass package checks. |
| RP-S23w | Admit fresh database test fixtures and templates | DONE | S23v; discovered in S23a | Pinned migration 9/9, PostgreSQL 6/6 and service 4/4; §10 checkpoint | Fresh singleton established through production initializer; historical scratch preserved; cleanup complete. |
| RP-S23x | Read-only configuration inspection admission | DONE | S23v; discovered in S23a | Package `20261005T231708Z-p63310` rejected read-only roots; §10 | Read-only configuration, migration and fresh package pass; cleanup complete. |
| RP-S23y | Current-package credential capacity assessment | DONE | S23x; discovered in S23a | Retire intermediate password/HTTP fixture; frozen limits retained | Current package assessment 9/9, effective limits and cleanup pass; §10. |
| RP-S23a | Activate mandatory application admission | DONE | S23, S23v and S17 | Original S17 production exit retained in §7; closes RP-F21 only here | Three actual binaries/default builds enforce identity/mode/state before effects. |
| RP-S24a | SeaweedFS service identity/readiness | DONE | Discovered in S24 | Restart returns retry_exhausted | Same persistent state survives certificate/container replacement and cleanup. |
| RP-S24 | Windows 11 / WSL2 package acceptance | DONE | S23a | RP-F28; frozen `20261006T001353Z` package/assessment results and §10 checkpoint | Real binaries, WSL2 boundaries, full package scenarios, routing and cleanup pass. |
| RP-S25 | Final validation/handoff completion | IN_PROGRESS | S24 and all active prerequisites | RP-F29; finalization `20261006T001750Z-p46630`, final fast/release validation pending | Active slices DONE, complete WSL2 final release/cleanup, five readiness dispositions; completes last. Native Linux remains deferred. |

## 10. Session handoff log

Prior planning sessions, implementation checkpoints and failures remain below without reassigned run identities. Rows labeled planning/document update describe those earlier sessions; the execution rows govern current work.

### Implementation execution — 2026-10-05

| Workstream | Changes / compatibility | Commands and canonical results | Failures / residual risk / next dependency |
| --- | --- | --- | --- |
| S16 start | Confirmed clean `df720535b6b6622cab8b3d399e868dc968e0f6c4`; updated active scope, staged admission, certificate policy and prerequisite routing before implementation | Git baseline/status, owner/source inspection and `make help-all` succeeded; local module archive matches recorded SHA-256 | No implementation qualification claimed. Current WSL host is not the reference target; availability/access requested. Finish S16a before S16 exit and S17. |
| S16a start | Selected released pgx v5.11.0 after source review of RequireAuth dispatch guards; dependency/test work authorized as explicit prerequisite | [Released policy](https://github.com/jackc/pgx/blob/v5.11.0/pgconn/require_auth.go) and [dispatch](https://github.com/jackc/pgx/blob/v5.11.0/pgconn/pgconn.go) inspected; PostgreSQL task-guide discovery started | Existing v5.9.2 cannot enforce policy before dispatch. No driver/runtime changes or tests completed yet. |
| S16a implementation | Updated `go.mod`; Make produced `go.sum`; added real driver challenge regression and authored semantic row; `make generate` refreshed the topology render index. Updated the operator Rows test double for pgx's added `TypeMap` method. Production connection policy remains unchanged until S23. | Bootstrap PASS `20261005T130154Z-p85232`; `GOFLAGS=-mod=mod make build-migrate` PASS `20261005T130639Z-p96501`; PostgreSQL owner slice PASS `20261005T130704Z-p541` (4/4 units, cleanup 7/7); generate PASS `20261005T130823Z-p24014`; agent-finalize PASS `20261005T130843Z-p26951`; server/operator builds PASS `20261005T130910Z-p31566` / `20261005T130910Z-p31572` | Finalization retained-run maintenance skipped: RESULTS_DIR unset. Wider upgrade validation still pending after test-double repair; S16a remains IN_PROGRESS. |
| S16a failure history | Corrected new-row ASCII ordering and tool-managed missing archive sum; generated stale topology through Make; corrected added Rows interface in operator test fixture | Initial test-slice `20261005T130342Z-p87812` and agent-finalize `20261005T130448Z-p94039` stopped before canonical summaries (row ordering). `test-catalog-check` FAIL `20261005T130516Z-p94798`; test-slice FAIL `20261005T130610Z-p95512` (missing sum; cleanup 3/3); agent-finalize FAIL `20261005T130706Z-p808` and json-shape FAIL `20261005T130747Z-p23024` (stale generated topology). `test-fast` FAIL `20261005T130909Z-p31265` (736/739 units; three operator compile failures); go-vulncheck FAIL `20261005T131222Z-p8493` (same test-double compile failure) | These failures are related to this dependency/test change. Canonical summaries are `<run-root>/run-summary.json`, except leaf catalog/bootstrap/build/generate summaries under their target `tool-run-summary.json`; initial graph failures have no summary. Investigation invocations without RESULTS_DIR and with undeclared debug/OWNER inputs also rejected as usage/configuration errors, with no product execution. No failed or partial run is qualification evidence. |
| S16a completion | Upgraded supported dependency and regression route; test-only Rows adapter now implements TypeMap. No production authentication policy or stored format changed. | Focused operator row PASS `20261005T131657Z-p73743`; agent-finalize PASS `20261005T131604Z-p68853`; `make test-fast` PASS `20261005T131726Z-p75497` (739/739, no skips/cancellations, cleanup 646/646); `make go-vulncheck` PASS `20261005T131658Z-p74091` (4/4, cleanup 6/6). `make explain-run` inspected these canonical summaries and identities. | S16a DONE. Retained-run maintenance skipped because RESULTS_DIR was unset. Source digest for the corrected implementation is `sha256:b89537d41f4e63ee4f0c676e94587831157d910cc66ca2c91f349df8d637dca0`, based on dirty `df720535b6b6622cab8b3d399e868dc968e0f6c4`. Resume S16; S16b still blocks its exit. |
| S16 module feasibility | Frozen archive source inspected, including PBKDF2's KDF-specific HMAC treatment; exact archive digest preserved. Real driver challenge test also builds/runs with `GOFIPS140=v1.0.0-c2097c7c GODEBUG=fips140=on` and graph cache disabled. | Pinned-module PostgreSQL row PASS `20261005T132052Z-p25251` (1/1, cleanup 3/3); exact command: `GOFIPS140=v1.0.0-c2097c7c GODEBUG=fips140=on make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_authentication_policy_rejects_passwo_d477ceca4c CARTULARY_HARNESS_CACHE_MODE=off` | Development feasibility evidence only. This command does not supply S17's actual-binary identity assertions, S23's mTLS execution, or S24's operating-environment qualification. |
| S16 coordinated source candidate | Amended Core 00 status, Core 01 public credential/Recovery requirements, Core 04 security/configuration, Reference Pack method qualification, telemetry HTTPS and Testing Harness requirements. Preserved domain vocabulary/navigation and completed Reference Data boundaries. | `make lint-markdown` PASS `20261005T132422Z-p66181` for configured coverage; `git diff --check` PASS before closeout. Machine contract/client/runtime updates remain in S17–S24, not falsely claimed complete by source edits. | Fresh-deployment breaks are explicit. Password capacity fixed before measurement: 2 active, 8 pending, 2000-ms queue; public 503/Retry-After 1; fixed reference acceptance in REQ-04-166. S16b environment admission remains unresolved; package/release validation has not begun. |
| S16 blocked handoff | S16a DONE; S16 and S16b BLOCKED; S17–S25 TODO; S07 remains open. Source amendments are a reviewable candidate, not implementation completion, package qualification, formal Reference Pack adoption or deployment approval. | Final configured-scope Markdown PASS `20261005T133058Z-p72630`, summary `.cartulary/test-results/20261005T133058Z-p72630/adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` PASS. Document review confirmed unchanged historical tracker tail, unique added requirement declarations and unchanged 197-file direct target inventory. | `lint-markdown` covers the changed `docs/*.md` and `docs/spec/**/*.md` owners but not this handoff path. S24 smokes, final release-check and S25 completion were not run because S16b blocks progression. Retained-run maintenance remains skipped because RESULTS_DIR was unset. No automatic deployment, reset, conversion or historical-evidence relabeling occurred. |
| RP-P04 / S16b retarget start | User selected WSL2 under Windows 11 and indefinite native-Linux deferral. Marked plan/source refocus and S16b IN_PROGRESS before source amendments. Preserved the dirty implementation from S16a. | Inspected Core 00/04, Reference Pack, Testing Harness platform matrix, domain authority, tracker/framework and driver history; reviewed official Microsoft, Go and NIST sources linked in §1. | Prior native-Linux access blocker and request are superseded, not resolved by a native qualification run. One cryptographic policy, exact module/archive and pack semantics are unchanged. |
| S16b completion | Recorded Windows 11 Pro 26200.9457, WSL 2.7.3.0, Ubuntu 26.04, WSL kernel/CPU and Docker Desktop backend/Engine/Compose; repository is ext4. Reconciled Core 00/04, Reference Pack and Harness target/claim requirements. Defined full WSL2 release, storage/trust/network/interruption checks and fixed measured-container resource limits. | Read-only PowerShell OS query, `wsl.exe --version`, distro listing, `uname`, `lscpu`, `findmnt` and bounded Docker metadata; `make doctor` PASS `.cartulary/test-results/20261005T134056Z-p77862/doctor/tool-run-summary.json`; `make lint-markdown` PASS `.cartulary/test-results/20261005T134653Z-p81528/adhoc/lint-markdown/tool-run-summary.json`; diff and supplemental document review PASS. | S16b DONE for engineering feasibility. Bare `wsl` lookup inside PowerShell failed; resolved with `wsl.exe` from WSL. `.wslconfig` is absent; defaults are not assumed to be fixed allocations. A guessed doctor-source glob missed and was resolved by file enumeration. No product failure or host settings change occurred. S24 must record final limits, full Desktop version, network/image identities and new package results. Formal CMVP applicability remains unestablished. Next: S16 completion review. |
| S16 / RP-P04 completion | Coordinated Core 00/01/04, Reference Pack, telemetry and harness candidate has one service policy, explicit fresh-state admission/formats, bounded credentials and consistent WSL2 target/claim boundaries. S16a proves supported driver admission; S16b establishes the accessible execution route. This refocus edited only the tracker, Core 00/04, Reference Pack and Harness prose; prior implementation remains intact. | Configured-scope Markdown PASS `20261005T134653Z-p81528`; `git diff --check` PASS. Supplemental human-document review verified the unchanged historical tail, exact 197-file inventory, twelve active sections and S16–S25 sequencing. Owner/projection paths and per-service routes remain in §8. | S16 and RP-P04 DONE. Lint covers `docs/*.md` and `docs/spec/**/*.md`, not this handoff. Product tests, generation, agent-finalize, package smokes and final release were not rerun for this documentation-only retarget; retained-run maintenance skipped because RESULTS_DIR is unset. No revised runtime/package acceptance is claimed. Next: mark S17 IN_PROGRESS before its first change; complete S25 last and then S07 for WSL2 scope. |

**Resume:** S17 is IN_PROGRESS before implementing the pinned build identity
and staged startup admission. S16/S16a/S16b completion is recorded above. The
native-target access request is historical and superseded by the WSL2 decision.
Follow S17–S25 in order using the complete WSL2 release matrix; historical runs
cannot close S24/S25. Native Linux has no dependency in this chain. Formal
environment claims remain separate; no security-service or cleanup gate is waived.

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

### Finishing implementation checkpoint — 2026-10-05

S16c marked IN_PROGRESS before source edits. Latest authorization adopts the finishing sequence and staged enforcement above. `make task-guide ROLE=module-author OWNER=module.revisions` passed. Existing dirty work and historical records are preserved; no implementation validation or package acceptance is claimed by this checkpoint.

S16c DONE: corrected AC-526 to configuration v3, key-ring v2 and cft4; aligned the existing Recovery journal payload identifier with v5 (envelope v2 remains S21); clarified partial-versus-empty initialization retry, workflow admission, operator-started WSL2 operation and Windows-client evidence. Reviewed Recovery envelope v3, integrity manifest v4, marker/proof v5 requirements and affected acceptance criteria. No applied migrations or runtime behavior changed. Revisions task guide passed; `make lint-markdown` passed at `.cartulary/test-results/20261005T143313Z-p459/adhoc/lint-markdown/tool-run-summary.json` (prior pre-final-edit run `20261005T143237Z-p97782` also passed). `git diff --check` and human document review passed; historical completed-plan tail remains byte-identical. Markdown's configured scope includes owner documents, excludes this tracker. No resources created; no cleanup required. Next dependency: S17.

S17 marked IN_PROGRESS before implementation. Database-migrations and server task guides passed. A guessed `harness.build-readiness` guide rejected with usage error because that owner does not exist; discover the maintained owner before continuing. No product check failed. Production enforcement remains S23a; no intermediate candidate is release-ready.

S17 checkpoint: implemented candidate execution identity checks, a migration-owned singleton/serialized admission component, authored module pins/cache identity inputs, diagnostic target routing and current Recovery metadata participation. These changes are unvalidated and not production-active. Generation failures are retained: `20261005T143858Z-p4887` (SQL qualification/named constraints), `20261005T144119Z-p6275` (redundant Go language/toolchain directive; restored language floor), `20261005T144228Z-p7542` and `20261005T144526Z-p9979` (dependent Recovery bindings), `20261005T144615Z-p12087` (catalog count), `20261005T144651Z-p13918` (catalog ordering), `20261005T144754Z-p15826` (assessment routing metadata), and `20261005T144847Z-p19036` (generated Make density). Canonical summaries are `<run-root>/generate/tool-run-summary.json`; detailed child stderr identifies each failure. Source/projection repairs precede reruns. S17a is now IN_PROGRESS and S17 BLOCKED on it; do not advance to S18. No application state or package was published.

S17a DONE: factored Node readiness into one generated runtime macro; generated binding size is 165337 bytes against the unchanged 184320-byte limit, total generated Make bytes 170034. `make generate` PASS `20261005T145032Z-p22695` (`generate/tool-run-summary.json`); `make harness-command-surface-contract` PASS (exit 0); `make test-slice OWNER=harness.command_surface` PASS `20261005T145100Z-p27848` (2/2 units, canonical `run-summary.json`). `make lint-markdown` PASS `20261005T145100Z-p27961` for configured owner-document coverage, excluding this tracker; `git diff --check` PASS. No product behavior or public inputs changed. Generation and harness fixtures cleaned their owned temporary resources. S17 resumes IN_PROGRESS; S18 remains unstarted.

S17 validation checkpoint: `make format-go` and platform.cryptography task guide passed. Execution-identity matrix PASS `20261005T145239Z-p58983` (1/1). `cryptographic-policy-assessment` at `20261005T145239Z-p59438` passed 3/3 units but failed final retained-artifact admission (`artifact_error`): new receipt directory/file permissions were not owner-only. This run is NOT a target pass. Database admission FAIL `20261005T145239Z-p59016` (2/3 units, new row failed). Inspection confirmed Goose `Up` invokes mutating `HasPending` before SessionLock; S17b is IN_PROGRESS and blocks S17. Do not treat failed component evidence as package or production qualification. Runtime facade enforcement remains unactivated.

S17b DONE: a dedicated admission connection holds initialization lock `4097083627` across read-only freshness recheck, admitted ledger creation, full provider invocation, postconditions and identity commit. Goose's supported store creates the normal zero-version ledger on that admission connection before work begins; losing the connection leaves incompatible retained state and cannot publish identity. Work retains the existing migration lock `4097083626`; separate borrowed pools avoid waiter/work starvation and target/effective-role bindings are compared before mutation. Core 04 records the ordering/loss obligations. Applied SQL remains immutable. Go language floor remains 1.27.0; actual compiler remains pinned to go1.27.1.

S17b validation: `make generate` PASS `20261005T150821Z-p64522`; focused database admission PASS `20261005T150903Z-p67684` (3/3 units, 6/6 cleanup), including fresh/current concurrency, incompatible state, empty retry, partial failure and real connection termination. Earlier focused PASS `20261005T150116Z-p93235` predates the added connection-loss guard and is not its evidence. Database owner slice PASS `20261005T150322Z-p38353` (9/9); its prior FAIL `20261005T150135Z-p8736` (7/9) exposed two current-schema digest expectations and one table-count expectation, updated for appended migration 64. Boundary PASS `20261005T150904Z-p68004` (3/3, 5/5 cleanup). Markdown PASS `20261005T151007Z-p88958` (`adhoc/lint-markdown/tool-run-summary.json`, excludes tracker); diff check PASS. Canonical graph summaries and cleanup live in each run root. No package publication or deployment occurred. S17 resumes IN_PROGRESS; S18 remains TODO.

S17 component checkpoint: added bounded, no-create filesystem/object-store inspection and mandatory caller-owned fresh-dependency inspection inside migration exclusion, rechecked before identity commit. Real read-only database admission and retained external-state rejection pass. All Go artifact caches include module/mode/toolchain/pin inputs and binary receipts; receipts are validated again after cache hits. The public assessment builds all three applications in owned disposable output, verifies exact archive/metadata, rejects ordinary artifacts and tests enabled/disabled/strict diagnostic modes. Default builds and facade invocation remain staged until S23a.

S17 evidence: generation PASS `20261005T152518Z-p4987` (prior `20261005T152012Z-p5490`); assessment PASS `20261005T152551Z-p8362` (3/3, with build-helper receipt regression; prior `20261005T152116Z-p9910` also passed). Rooted filesystem PASS `20261005T152116Z-p9444` (2/2); database admission PASS `20261005T152116Z-p9455` (3/3, 6/6 cleanup); managed object storage PASS `20261005T152116Z-p9454` (3/3, 6/6 cleanup); cryptography owner PASS `20261005T152201Z-p59228` (1/1); harness contract PASS `20261005T152201Z-p59393` (2/2), including cross-mode/selector/toolchain cache rejection. Generate drift PASS `20261005T152201Z-p59123` (4/4), JSON shape PASS `20261005T152201Z-p59136` (3/3), generated policy PASS `20261005T152201Z-p59130` (3/3), toolchain PASS `20261005T152201Z-p59145` (2/2), boundary PASS `20261005T152201Z-p59377` (3/3). Canonical graph evidence is `run-summary.json` and `cleanup-results.json` under each run root; assessment detail is `cryptographic-policy-assessment/assessment.json`. Scripts lint PASS `20261005T152413Z-p99917`, shell lint PASS `20261005T152413Z-p99923`, Markdown PASS `20261005T152413Z-p99919` (excludes this tracker). Diff and unchanged historical-tail review pass. Discovery attempts using nonexistent owner `harness.scheduler` and internal smoke identifiers as Make targets rejected with usage/no-rule errors; the maintained public assessment and harness contract supply the relevant execution instead. No package publication, current-state certification, final validation or deployment approval is claimed.

S17 DONE: all component exits pass. Database owner slice PASS `20261005T152552Z-p8479` (9/9); final JSON shape/policy PASS `20261005T152629Z-p31221` / `20261005T152629Z-p31219`. Recovery catalog row initially FAIL `20261005T152627Z-p30970` because the authored new contribution digest encoded absent object families as `[]`, while the established canonical owner encoding uses `null`. Corrected the authored contribution/catalog/dependent binding digests and current expectation; generated normally. Final generation PASS `20261005T152851Z-p42099`; Recovery catalog PASS `20261005T152920Z-p45125` (1/1); Markdown PASS `20261005T152921Z-p45444` (configured owner coverage, excludes tracker); diff check PASS. No migration history was rewritten; the new format table is migration-owned excluded schema metadata. Owned resources cleaned, borrowed handles remain usable. Historical results remain intact. RP-F21 remains OPEN for S23a; production startup and package qualification are not completed by this infrastructure exit. Next dependency: S18.

S18 marked IN_PROGRESS before its first implementation change. Platform cryptography guide consulted. Implement exact-key admission, explicit-length domain separation, bounded per-value sealing and opaque bounded artifact capabilities; owner migrations follow in S19–S21.

S18 checkpoint: implemented exact 32-byte admitted keys, separated HKDF constructions with length/count framing, envelope v1 with fresh 32-byte salt and module nonces, bounded inputs and separate opaque artifact writer/reader capabilities. Added authored primitive contract/vector family and normal generation; independent vectors use Python standard-library HKDF and cryptography/OpenSSL AES-GCM, with test-only fixed nonce bytes. Pinned-module race-enabled owner slice PASS `20261005T153734Z-p67130` (both execution and primitive rows); ordinary development slice PASS `20261005T153826Z-p71947`. JSON shape FAIL `20261005T153734Z-p67094` required registering the new active family in the maintained shape validator; follow-up `20261005T153826Z-p71911` rejected stale generated topology after that validator edit. Regenerate before retrying checks. No owner migration, compatibility fallback or production enforcement has been activated.

S18 DONE: independent purpose/message/artifact vectors, exact key admission, ambiguous-context separation, full-byte tampering, wrong key/AAD/context, inclusive bounds, malformed framing, non-resumable artifact reader and concurrent artifact-use limit all pass. Final pinned-module race-enabled owner slice PASS `20261005T154056Z-p81959` (both semantic rows in one unit; actual enabled `v1.0.0` asserted when pinned); generation PASS `20261005T153950Z-p73631`; drift PASS `20261005T154056Z-p81915`; JSON shape PASS `20261005T154056Z-p81923`; generated policy PASS `20261005T154056Z-p81919`; boundary PASS `20261005T154056Z-p82043`; Markdown PASS `20261005T154056Z-p82058` (excludes tracker); diff PASS. Canonical summaries and cleanup are in those run roots; tests leave no persisted application state. Core 04 now records exact primitive framing/component bound, with authored machine contracts and independent vector provenance. Existing owners still use their old mechanics until S19–S21; no legacy reader was added to the new platform API. Next dependency: S19.

S19 marked IN_PROGRESS before implementation. Replace password encoding/verification and MFA together with route-wide admission, public overload contracts, client projections, replay handling and fixed-profile load evidence. S18 primitives remain domain-neutral.

S19 checkpoint: one context-aware PBKDF2 service replaces both Argon2 encoders; a shared two-active/eight-pending workflow budget preserves sequential derivations and completed mutation replay. SHA-256 TOTP uses canonical 32-byte secrets; authored OpenAPI/error contracts and clients advance together. Pinned focused password/capacity slice PASS `20261005T155708Z-p22796` (2/2); bootstrap PASS `20261005T155708Z-p22797` (1/1). Generation FAIL `20261005T155124Z-p98789` required eight candidate release-ledger entries; repaired generation PASS `20261005T155453Z-p6384` and `20261005T155654Z-p19748`. Later schema tightening invalidated one candidate fingerprint in FAIL `20261005T155808Z-p29316`; updating that fingerprint and three secret constraints preserves the released baseline.

Auth owner run FAIL `20261005T155809Z-p30600` (20/38 units passed, 7 failed, 11 skipped): one exact response-status expectation needed the new 503; concurrent generation invalidated the run's frontend source snapshot and prevented browser startup. Those derived missing browser summaries are not product passes. Generation and source-snapshot execution will now be sequential. Canonical run summary and cleanup remain in that original run root; no historical run is relabeled. Fixed-container performance evidence is still required before S19 completion.

S19a checkpoint: added `make credential-capacity-assessment` through the authored task surface/work graph, a pinned real server/migrate build, isolated PostgreSQL/application containers, a public HTTP bootstrap/MFA/load client and private canonical environment/binary/observation/cleanup artifacts. First FAIL `20261005T161033Z-p82653` exposed a fixture key encoded as base64 rather than Revisions' required base64url; second FAIL `20261005T161126Z-p92252` exposed the client's incorrect second-factor assertion shape. Both cleaned all owned resources. Corrected assessment PASS `20261005T161221Z-p93789` (3/3): 200 successful valid logins, p95 323.151 ms; overload 10 accepted/54 rejected, maximum rejection 67.679 ms; cancellation and recovery passed. Effective cgroups: CPUs 0–1, quota 200000/100000, memory 2147483648, swap 0. This is intermediate credential evidence, not final package qualification.

Environment receipt refinement FAIL `20261005T161420Z-p7923` because Docker Desktop omits its default WSL engine setting; it does not establish a non-WSL backend. Record actual Desktop executable version, running `docker-desktop` WSL2 distribution and daemon Microsoft kernel instead of requiring that optional setting to be present. CLI plugin version is not the Desktop application version. The temporary workspace was removed before any container allocation in this failed run. Added interrupt cleanup and authored service side effects; final S19a validation remains pending. Generation PASS `20261005T161522Z-p9054`.

S19a DONE: final assessment PASS `20261005T161656Z-p26687` (3/3; canonical `credential-capacity-assessment/{binary-identities,environment,observations,cleanup}.json` and root `run-summary.json`). p95 321.246 ms for all 200 successful password+TOTP logins; 10 accepted/54 rejected under synchronized overload, slowest 503 58.940 ms; cancellation/recovery passed. Desktop application 4.75.0.227598, running WSL2 backend and actual cgroup limits recorded; the optional settings value is retained as absent, never inferred as enabled. All three containers, private network and workspace were removed. Command-surface contract PASS `20261005T161543Z-p12322`; script/shell lint PASS `20261005T161543Z-p12103` / `20261005T161543Z-p12109`. No limits/costs were relaxed. Final-package/transport enforcement and renewed package measurements remain S23/S23a/S24. Resume S19 lifecycle/client validation; S20 remains TODO.

S19 validation checkpoint: pinned Auth unit slice PASS `20261005T161753Z-p27975` (3/3); frontend typecheck PASS `20261005T161755Z-p28573` (2/2), full frontend unit PASS `20261005T161755Z-p28575` (691/691). Service-backed Auth FAIL `20261005T161754Z-p28248` (23/26, with reset failure and two dependent skips; 36/36 cleanup) discovered S19b. All executed credential/browser scenarios passed, but the overall run is not acceptance. Drift/JSON/policy PASS `20261005T161914Z-p91117` / `20261005T161914Z-p91131` / `20261005T161914Z-p91124`; boundary/toolchain/Markdown PASS `20261005T161914Z-p91270` / `20261005T161914Z-p91122` / `20261005T161914Z-p91294`. Markdown excludes the controlling handoff and guides.

S19b DONE: reset explicitly excludes migration-owned cryptographic identity and compares its full before/after value without initializing missing identity. The existing integration row now proves both absent and populated identity preservation, complete domain reset and unchanged read-only Recovery grants. Pinned reset integration PASS `20261005T162310Z-p95444` (3/3); real stateful Auth-to-clock browser reset PASS `20261005T162403Z-p28456` (13/13). Canonical root summaries, reset diagnostic/attempt artifacts and cleanup retain those run identities. `git diff --check` PASS; the historical S11–S15 tracker tail remains byte-identical to HEAD. No production grants or historical migration were changed. Resume S19; renew its complete service-backed slice before closure.

S19 final-owner validation: removed bootstrap's remaining duplicate password validator; its parser now delegates to the shared password policy. Pinned Auth vector/capacity/password-replay slice PASS `20261005T162657Z-p76103` (2/2), complete pinned service-backed Auth slice PASS `20261005T162657Z-p76102` (26/26), including all selected browser groups and reset boundaries. Bootstrap unit FAIL `20261005T162657Z-p76096` (0/1) was the authored diagnostic golden's old short-password wording; update only that message to the shared bounded-policy diagnostic, preserving the path/reason and validation identity. Generation PASS `20261005T162616Z-p68289`. Renew bootstrap and fixed-profile evidence before closing S19.

S19 DONE: final `GOFIPS140=v1.0.0-c2097c7c GODEBUG=fips140=on make test-slice OWNER=platform.bootstrap` PASS `20261005T162958Z-p36340` (1/1), and renewed `make credential-capacity-assessment` PASS `20261005T163057Z-p36924` (3/3). All 200 password+SHA-256-TOTP logins succeeded with p95 319.621 ms; overload accepted ten/rejected 54, maximum rejection 36.409 ms; cancellation and capacity recovery passed. Exact binary/module/host/cgroup identities and removed-container/network/workspace evidence are in `credential-capacity-assessment/` under that run root. Pinned complete Auth service-backed, Auth unit, bootstrap and full frontend evidence above satisfy lifecycle, replay, malformed-record, overload, browser and client exits. No compatibility verifier remains for Argon2 or SHA-1 TOTP; the historical row identifiers retain their original spelling. Authenticator requirements and retry behavior are documented in the developer guide. Intermediate state is disposable; no final package or deployment was published. S19a/S19b are complete and next dependency is S20.

S20 marked IN_PROGRESS before implementation. Account for Auth active/pending/enterprise secrets, pagination v2, Revisions cft4, Network Flow nfc3 and indirect HMAC consumers. Add only new migrations for changed persistence; key admission/derivation is shared, but payloads, record binding, authorization and rotation stay with the owners. Read owner guides before the first implementation change.

S20a checkpoint: Testing Harness source/control schema now permits deterministic table IDs only; active response is v3, with the unchanged v2 schema retained for historical artifacts. Removed nonce consumer/registry API and module dependency. The original randomness verification row remains; its obsolete `TestNetworkFlowRandomnessRouteConsumesHexBytes` subcase is replaced by `TestNetworkFlowRandomnessRouteRejectsRetiredNonceControl`. Generation PASS `20261005T163904Z-p50893`; shape/drift/policy PASS `20261005T164005Z-p59110` / `20261005T164005Z-p59102` / `20261005T164005Z-p59106`; boundary/Markdown PASS `20261005T164006Z-p59229` / `20261005T164006Z-p59246`. Focused pinned slice FAIL `20261005T164004Z-p58780` (1/2): retired nonce-byte assertions left an unused base64 import. Remove it and rerun the same owning rows. Module-generated sealing itself remains S20; this prerequisite removes the incompatible test control.

S20a DONE: renewed pinned module rows PASS `20261005T164118Z-p71795` (2/2). The unchanged randomness row identity now includes explicit retired-control rejection, and the cursor continuation/rotation row passes without an injectable entropy source. Shape, drift, policy, boundary and Markdown results above passed; the actual Markdown coverage excludes the tracker. Retained schema v2 has no active producer; v3 is the only emitted control response. No fallback or production bypass was added. Resume S20 owner sealing and persistence migrations.

S20 pagination checkpoint: current `pc2` envelopes now use S18 with exact-size keys, strict framing and bounded decode; actor/route/query/continuation interpretation stays local. Added the previously missing `platform.pagination` owner route without changing historical validation identities. Generation PASS `20261005T164642Z-p73945`; owner guide succeeded and pinned `make test-slice OWNER=platform.pagination` PASS `20261005T165158Z-p83091` (1/1). Revisions and Network Flow codec migrations follow; S20 is not complete.

S20 generation failure `20261005T165348Z-p83990`: replacing the obsolete caller-entropy subcase with exact-key rejection changed selector sort order. Original row IDs are retained; restored ASCII order in the authored family. Regeneration PASS `20261005T165503Z-p92024`. No product execution or cleanup was involved.

S20 token checkpoint: pinned Revisions row PASS `20261005T165609Z-p5151` and Network Flow continuation/rotation row PASS `20261005T165609Z-p5154`. Current wire formats are cft4/nfc3; key-ring schemas advance to v2 with HKDF/GCM mechanics and unchanged HMAC safe-digest semantics. Auth migration 65 replaces split ciphertext/nonce columns and explicitly rejects retained old secrets; no applied migration changes. Generation FAIL `20261005T170210Z-p10134` identified the missing authored migration-owner allocation; add Auth ownership before regenerating. Auth component/application validation is pending.

S20 Auth checkpoint: generation PASS `20261005T170416Z-p16249`. Pinned Auth slice FAIL `20261005T170443Z-p24021` (1/2 work units): the new key/envelope vectors passed, while two adjacent seed helpers had a renamed declaration left mismatched at scan/return. JSON shape FAIL `20261005T170443Z-p23984` (2/3): the new explicit Auth fixture root needed classification in the authored support inventory. Repair both before renewed validation. Test credentials are now explicit fixture inputs, and browser credentials must survive owned-backend restart; ordinary development requires a persisted operator-provisioned key. No production fallback remains.

S20 renewed Auth component/handler slice PASS `20261005T170638Z-p32043` (2/2), and JSON shape PASS `20261005T170638Z-p32013` (3/3). The chained boundary command used an unavailable target spelling and did not run; `make help-all` confirms the public target is `backend-module-boundary-check`. Full lifecycle, browser and changed-schema validation remain pending.

S20 application checkpoint: pinned complete Auth service-backed slice FAIL `20261005T170800Z-p38730` (25/26): one store fixture still contained a 20-byte TOTP secret, now rejected by exact owner payload admission. Every browser group, enterprise route and remaining lifecycle row passed. Migration owner FAIL `20261005T170800Z-p38737` (8/9): `pgtest` retained the earlier schema-hash expectation. Preserve those failed identities and cleanup records; fix the fixture/hash and renew affected validation. Boundary/scripts/shell PASS `20261005T170800Z-p38856` / `20261005T170800Z-p38874` / `20261005T170800Z-p38893`. Add actual-route pending/active/enterprise substitution tests and require a current PKCE envelope for every OIDC callback before invoking the verifier.

S20b DONE: current Evidence upload tokens v3 replace readable storage keys with a separately derived HMAC binding; the owner recomputes it against live blob state before storage access. Payload claims, authorization and lease lifecycle remain Evidence-owned; HMAC-SHA-256 remains the supported authenticator. Tokens reject old versions, noncanonical framing, unknown members and values exceeding 4096 bytes. Generation PASS `20261005T171819Z-p85787`; pinned token row PASS `20261005T171843Z-p93906` (1/1) and actual upload/authorization row PASS `20261005T171843Z-p93919` (3/3); cleanup completed 3/3 and 6/6. JSON/boundary/Markdown PASS `20261005T171843Z-p93863` / `20261005T171843Z-p94044` / `20261005T171843Z-p94068`; Markdown excludes the handoff, guides and module READMEs. Diff check passed and historical tracker tail is unchanged. Public target shapes remain unchanged; old targets must be reissued. Resume S20 final owner validation before S21.

S20 remaining-validation checkpoint: Auth substitution/reset rows PASS `20261005T171326Z-p31569` (4/4); targeted migration-harness validation PASS `20261005T171326Z-p31568` (1/1). Earlier Evidence service row PASS `20261005T171326Z-p31590` predates S20b and is not its acceptance. Drift/policy PASS `20261005T171326Z-p31505` / `20261005T171326Z-p31514` precede S20b generation and will be renewed as needed. Failed Auth/migration runs above completed cleanup 38/38 and 13/13. S20 still IN_PROGRESS.

S20 renewed validation: complete Auth service-backed slice PASS `20261005T172143Z-p23110` (26/26); Revisions history pagination PASS `20261005T172143Z-p23125` (3/3); bound-secret vectors PASS `20261005T172206Z-p17264`; frontend typecheck PASS `20261005T172206Z-p17502`. Drift/policy/harness PASS `20261005T172206Z-p17033` / `20261005T172206Z-p17044` / `20261005T172206Z-p17631`. Workbook FAIL `20261005T172143Z-p23126` (2/3), because its authenticated-invalid-version fixture still constructed legacy AES-GCM. Network Flow FAIL `20261005T172143Z-p23119` (11/12), because its functional fixture still expected the retired nonce injection route. Replace those fixtures with current-format adversarial tests and renew owning rows; no production fallback is added. Frontend unit validation remains running.

S20 DONE: renewed Workbook routes PASS `20261005T172846Z-p13475` (3/3), Network Flow pagination PASS `20261005T172846Z-p13497` (3/3), Revisions key-ring/token row PASS `20261005T172846Z-p13481`, Network Flow key-ring/rotation row PASS `20261005T172941Z-p58204`; explicit v1 key-ring rejection is covered. Frontend unit PASS `20261005T172206Z-p17520` (691/691). Format PASS `20261005T172802Z-p99743`; Markdown PASS `20261005T172846Z-p13709` for configured owner coverage (excludes handoff/guides/module READMEs); diff check PASS. All latest owning, frontend and failed-route cleanup results are `completed`: Auth 38/38, Revisions 6/6, renewed Workbook/NF 6/6 each, NF key-ring 3/3, frontend 693/693; failed Workbook/NF runs cleaned 6/6 and 16/16. Canonical identities remain in each named run's summary and cleanup artifacts.

S20 final compatibility/caller disposition: Auth active/pending TOTP and enterprise PKCE use record/purpose-bound S18 envelopes with migration 65; enrollment activation reseals. Pagination v2, cft4, nfc3 and key-ring v2 reject old formats through owner error/reload flows. Evidence v3 retains HMAC authorization with an opaque storage binding. Fingerprint, CSRF/request and remaining purpose-key consumers use separated HKDF without a fallback master key; dev/harness fixtures supply explicit keys. Production AES/GCM duplication remains only in Recovery, the next S21 dependency. No intermediate package/state is release acceptance; RP-F21 remains open until S23a.

S21 start: S20 completion recorded before advancing. Marked S21 and discovered S21a IN_PROGRESS before implementation. Inspect Recovery and rootedfs guides, then establish private bounded staging before codec replacement.

S21a DONE: `Root.Stage` uses root-relative `O_TMPFILE|O_EXCL` private storage, enforces bounds/cancellation, retires callback write access, and returns only a read capability after success. No pathname is published; close/process exit reclaims the inode. Recovery authenticates once into staging, then copies those exact bytes. The filesystem must support unnamed staging (WSL2 guest ext4 does); unsupported storage rejects, without fallback. Linux [open semantics](https://man7.org/linux/man-pages/man2/open.2.html) were reviewed. Pinned rootedfs containment/staging PASS `20261005T173337Z-p70226`, streaming mutation/negative suite PASS `20261005T173337Z-p70231`; cleanup 3/3 each. Format/generation PASS `20261005T173254Z-p61785` / `20261005T173254Z-p61705`; JSON/boundary PASS `20261005T173337Z-p70186` / `20261005T173337Z-p70363`; drift/policy PASS `20261005T173400Z-p75529` / `20261005T173400Z-p75531`; Markdown PASS `20261005T173400Z-p75607` (configured coverage excludes this handoff); diff check PASS. No failed execution in this prerequisite. Resume S21 codec, proof and journal migration; package filesystem acceptance remains S24.

S21 implementation checkpoint: replacing parallel small/stream encryption with one v3 codec, exact Recovery key admission, shared artifact capabilities, bounded JSON fields and journal envelopes v2 via new migration 66. Current manifests/markers/proofs advance with explicit application-format binding. Format PASS `20261005T174016Z-p83799`; generation FAIL `20261005T174016Z-p83719` because retired schema paths were still in the generator catalog but absent from the current registry. Remove retired authored schema assets from this release (their identities remain listed as retired), then regenerate. An editing helper also attempted an unchanged write to a read-only vendored normalization file and stopped; no vendor content changed. Narrowed that edit before continuing. S21b separately tracks the discovered production historical-reader cleanup. No new codec/journal qualification is claimed yet.

S21 focused codec/marker PASS `20261005T174217Z-p98880` (2/2), JSON PASS `20261005T174217Z-p98849`; repaired generation PASS `20261005T174139Z-p90707`. Added an independent Python HKDF/AESGCM vector, exact-key/journal-context/use-bound tests and new schema generation PASS `20261005T174416Z-p1600`. Broader pinned Recovery test slice FAIL `20261005T174518Z-p10324` (20/24) and service-backed slice FAIL `20261005T174518Z-p10336` (16/19). Findings: one Graph binding digest expectation predates S17; browser restore helper still calls the pre-S19 password signature; due-verification reset attempts to truncate read-only application format metadata. JSON PASS `20261005T174518Z-p10281`. Preserve failed runs/cleanup identities; S21c separately fixes the reset boundary before renewed lifecycle validation.










S21c implementation: target emptiness and disposable reset now validate the complete frozen catalog and exclude schema metadata. Added real-role reset checks for present/absent format identity, unchanged migration ledgers, restricted grants and unknown-table rejection before mutation. Format/generation PASS `20261005T175356Z-p18420` / `20261005T175356Z-p18321`. Prior failed broad Recovery runs completed cleanup: `20261005T174518Z-p10324` 34/34; `20261005T174518Z-p10336` 29/29. Focused reset, due-process and browser checks are pending; no completion claimed.

S21c focused rerun FAIL `20261005T175421Z-p26767` (8/9): due-process lifecycle passed, but the new unknown-table regression showed information_schema silently excludes tables inaccessible to Recovery. Both target-reset and assembly coverage now use pg_catalog to detect those tables without broadening grants. Browser/Graph rerun FAIL `20261005T175421Z-p26759` (10/12): Graph passed; the browser helper omitted the explicit S20 Auth master key. Fixed the isolated fixture input; no fallback was added.

S21c DONE: frozen-catalog target checks/reset preserve present or absent format identity and both migration ledgers, without new grants or compatibility stamping. Public catalog inspection catches inaccessible unclassified tables before mutation; the assembly admission query follows the same rule. Pinned reset/coverage PASS `20261005T175649Z-p21684` (4/4); restored browser PASS `20261005T175650Z-p21959` (11/11). Due-process rows already passed in `20261005T175421Z-p26767`; that complete run remains FAIL because its first unknown-table test failed. Drift/policy PASS `20261005T175710Z-p63797` / `20261005T175710Z-p63810`; boundary/Markdown PASS `20261005T175508Z-p9018` / `20261005T175508Z-p9026`; format PASS `20261005T175637Z-p16426`. Failed focused runs completed cleanup 13/13 and 15/15. Next dependency: S21b, then final S21 qualification.

S21c final cleanup: reset/coverage 8/8, browser 14/14 completed. S21b marked IN_PROGRESS after the S21c checkpoint; Recovery guide consulted. Remove obsolete production capture/read APIs and unused supporting helper chains; retain legacy data only as explicit test fixtures and rejection cases, preserving row identities.

S21b checkpoint: moved historical capture/object-manifest data helpers into non-shipped `_test.go` fixtures, deleted the unused cross-package helper chain, and changed the original verification-v1 decoder test to current-reader rejection. Retired the unused support subcases for old CaptureParams, CaptureInput and SnapshotDigest; retained their containing row and useful evidence-permissions/environment checks. Pinned Recovery owner PASS `20261005T180003Z-p97359` (24/24), boundary/JSON/policy PASS `20261005T180003Z-p97486` / `20261005T180003Z-p97307` / `20261005T180003Z-p97300`; Markdown PASS `20261005T180003Z-p97529`. A follow-up audit found journal v2–v4 dispatch and an unused transitional catalog-shadow API; removing both before S21b completion. Historical journal test names remain as rejection identities.

S21b DONE: no historical backup/snapshot/object-summary/proof readers or writers remain in the production package; historical journal v2–v4 dispatch and transitional catalog-shadow API are removed. The only current journal payload is v5, with unchanged lifecycle semantics and strict shape checks. Current markers/proofs reject v4 and a mismatched application format. Retired identities remain in the registry and historical record; required old data construction is test-only. Final pinned Recovery unit/support slice PASS `20261005T180314Z-p67795` (5/5); complete service-backed Recovery PASS `20261005T180314Z-p67765` (19/19), including real backup, restore, due verification and restored browser. Generation/format PASS `20261005T180229Z-p54222` / `20261005T180256Z-p62282`; drift/JSON/boundary PASS `20261005T180314Z-p67710` / `20261005T180314Z-p67721` / `20261005T180314Z-p67901`; diff PASS. No execution failed in S21b. Resume S21 final codec bounds/interruption checks and operator documentation, then checkpoint before S22. Package transport/admission activation and final fresh-state smokes remain S23/S23a/S24.

S21b final cleanup: owner 34/34, service-backed 29/29, unit/support 7/7 completed. S21 resumed after its prerequisite checkpoint: add zero-byte small-artifact interoperability, fresh per-write salt evidence, interrupted/cancelled publication and bounded hostile JSON parsing; update current-format operator documentation. S21 remains IN_PROGRESS.

S21 final-owner run FAIL `20261005T180757Z-p35523` (23/24): all lifecycles passed; the newly added small-artifact test incorrectly expected immutable storage to overwrite a path. Preserve exclusive publication and fix that test to write a separate envelope with identical logical context, checking fresh salt and overwrite rejection. Migration owner PASS `20261005T180820Z-p78372` (9/9); real encrypted/atomic journal rows PASS `20261005T180820Z-p78367` (3/3). Markdown/policy PASS `20261005T180757Z-p35685` / `20261005T180757Z-p35477`; configured Markdown scope excludes the controlling handoff and package README.

S21 DONE: one current v3 streaming codec serves small/large/empty artifacts; exact keys, fresh artifact derivation, module nonces, 4-MiB chunks and the 1,048,576-use bound are enforced. Private staging authenticates the exact bytes exposed; incomplete/corrupt/reordered/duplicated/substituted inputs never publish successful output or admit a restore. Manifests v4 and markers/proofs v5 bind application format; migration 66 installs record-bound journal envelopes v2 with payload v5. Final pinned codec/bounds/interruption/independent-vector row PASS `20261005T181051Z-p37243` (1/1), correcting the test-only overwrite assumption from the retained failed owner run. The unchanged real lifecycle rows passed in that owner run and in complete service-backed `20261005T180314Z-p67765`; migration and final atomic journal checks passed as recorded above. No partial run is relabeled as a complete final-candidate pass. Format PASS `20261005T181034Z-p31857`; diff PASS; current operator format/key/backup instructions updated. Cleanup counts are recorded below. RP-F25 implementation is resolved; exact secure-package acceptance remains S24 after S23a. Next dependency: S22.

S21 final cleanup: failed final-owner run 33/33, migrations 12/12, atomic journal rows 6/6, final codec row 3/3 completed. S22 marked IN_PROGRESS after the S21 checkpoint; Reference Data guide consulted. Qualify unchanged canonical/TUF semantics through independent vectors, real bootstrap and owner lifecycle paths using the exact pinned execution identity.

S22 qualification checkpoint: strict real TUF diagnostic PASS `20261005T181325Z-p44133` (1/1); cryptographic build/identity/cache assessment PASS `20261005T181325Z-p44485` (3/3), with exact module/archive and all three binary receipts in its canonical assessment artifact. Full pinned Reference Data owner FAIL `20261005T181325Z-p44125` (21/23): canonical consumer continuation and all-profile continuation failed; verifier, lifecycle and browser rows passed. Narrow consumer rerun FAIL `20261005T181750Z-p91674` (0/1). Inspection found Reference Data writing a private cursor mode rejected by the S20 decoder. S22a is IN_PROGRESS before its implementation; consult both owner guides, use shared keyset mechanics and preserve owner-local bindings.

S22a narrow all-profile reproduction FAIL `20261005T181750Z-p91679` (2/3); framework continuation errors confirm the shared-mode mismatch. Implementation uses the supported keyset mode, tests mode/route/consumer substitution, and rejects invalid generic cursor payloads before encryption as well as after decryption. No owner-specific mode is added to the platform.

S22a shared pagination PASS `20261005T182105Z-p20926` (1/1). Reference Data reruns FAIL `20261005T182105Z-p20939` / `20261005T182105Z-p20951` because the new test reused a local variable name with a different type; renamed that test variable. Boundary/Markdown PASS `20261005T182105Z-p21127` / `20261005T182105Z-p21158`. The failure is in the new test, not a new runtime policy exception.

### Open risks and next session

| Date / session | Current state | Remaining risk | Next action |
| --- | --- | --- | --- |
| 2026-10-05 / document handoff | S16–S25 all TODO; previous package iteration complete | Current code still uses the old cryptography; owner exclusion, transitive service mapping and actual reference-environment suitability remain unresolved implementation prerequisites | After later authorization, start S16, then follow the dependency chain and update this tracker at every workstream exit. S25 finishes last. Keep normal review/canonical evidence; do not recreate retired dossiers. |

S22a DONE: pinned canonical-consumer row PASS `20261005T182258Z-p55799` (1/1, cleanup 3/3), actual all-profile consumer integration PASS `20261005T182258Z-p55809` (3/3, cleanup 6/6), and shared pagination PASS `20261005T182105Z-p20926` (1/1, cleanup 3/3). Current Reference Data cursors use shared keyset mode with unchanged query/actor/route/expiry binding; unsupported modes reject symmetrically at write/read boundaries. Format PASS `20261005T182226Z-p50577`; prior boundary/Markdown passes retained; diff check PASS. Pack bytes, trust history and consumer operations are unchanged. Resume S22 algorithm/key rejection and complete pinned verifier/lifecycle validation; secure-package acceptance remains S24.

S22 DONE: `GOFIPS140=v1.0.0-c2097c7c GODEBUG=fips140=on make test-slice OWNER=module.reference_data` PASS `20261005T182757Z-p86041` (23/23, cleanup 27/27); targeted `GODEBUG=fips140=only` canonical-pack-trust row PASS `20261005T182757Z-p86047` (1/1, cleanup 3/3). Actual verifier vectors assert toolchain `go1.27.1`, selector `v1.0.0-c2097c7c`, module `v1.0.0` and enabled mode. Added algorithm/scheme/key-length rejection before signature dispatch and through bootstrap/historical entry points. Full owner includes real bootstrap, complete TUF, thresholds, rotation, replay, historical checks, portability, immutable provenance, consumers and browser import. The four independent fixture files remain byte-identical to HEAD: manifest SHA-256 `246cf7ab27a3dd604205ada7ba838622399408a4bee367ef2e9c14d95861d6b7`, TUF `312e843d86b440fb988fa6f6895b62557a6045af4dfe27ea8bf3fee62a6b7732`, signed container `513faea6ec3615faedfd645497c6f89ef3858766a139e7d4fa5f48837e5b335c`, portable input `e1282ce9c618e475fb45f6439a4ae891c9b40d1bcfbdbc24c0a74fd4c71a44ce`. Preserved expected canonical manifest/payload digests `c7b6ef26aba1be1de4af440b8689af75a80616a8780eb06a1226182c29200ecd` / `e38ff4e85cc452ea298b402bd1003c3ba199db81e9f2fd9a8f3931369965e584`. Format PASS `20261005T182728Z-p80284`; Markdown PASS `20261005T182757Z-p86202` (configured coverage excludes this handoff); diff/historical-tail comparison PASS. `make explain-run` confirmed pass, exact dirty candidate identity and complete cleanup. Earlier failures remain recorded; no algorithm fallback or pack reissue. RP-F26 source/verifier execution is resolved; package qualification remains S24.

S23 marked IN_PROGRESS after S22 checkpoint. Source review confirms pgx accepts an explicit authentication allow-policy but still loads inherited/default credentials during parsing; closed adapter admission and controlled construction must prevent those reads. Shared TLS construction is needed before migrating PostgreSQL, object storage, enterprise HTTP, telemetry and listeners; keep trust/configuration/lifecycle with those owners. Consult owning guides before implementation. S23a mandatory activation, final package acceptance and formal CMVP applicability remain separate and incomplete.

S23b first generation FAIL `20261005T183530Z-p37855`: the new row named `platform.securefile`, but secure-file tests are routed by existing owner `platform.rootedfs`. Correct the authored collaborator to that established owner; no new ownership boundary is needed.

S23b DONE: the shared platform TLS boundary admits TLS 1.3/P-256/P-384, module-selected AES-GCM, current certificate key/signature parameters and verified chain/name checks. PEM bindings use immutable no-follow bounded reads; no bypass, fallback identity, plaintext retry or session resumption is exposed. Short-lived fixture CA/identity generation is test infrastructure only and changes no host trust. Pinned complete cryptography owner PASS `20261005T183651Z-p49740` (1/1 grouped unit); final expanded live TLS row PASS `20261005T183900Z-p63033` and strict diagnostic PASS `20261005T183900Z-p63038` (each 1/1). Live tests cover RSA-2048/P-256/P-384, rejected P-521/Ed25519, expired/future/name/root failures, TLS 1.2 rejection, mutual identity and certificate/trust replacement. File/key/PEM adversarial tests pass. All three runs cleaned 3/3 steps. Generation PASS `20261005T183552Z-p41069`; final format PASS `20261005T183831Z-p57264`; boundary/JSON/drift/artifact-policy PASS `20261005T183651Z-p49950` / `-p49666` / `-p49650` / `-p49693`; Markdown PASS `20261005T183901Z-p63195` (excludes handoff); diff PASS. The initial wrong collaborator failure remains recorded. Resume S23; this component does not imply that existing adapters or application listeners are already secure.

S23c first generation FAIL `20261005T184810Z-p70234`: the authored fixture-row collaborator list was not ASCII-sorted. Sorted the owner inputs; no runtime failure or policy fallback. Closed URI parsing, controlled driver construction, certificate-only fixture provisioning and private certificate cleanup are implemented but not yet qualified.

S23c first pinned PostgreSQL owner FAIL `20261005T185159Z-p88332` (4/5, cleanup 8/8). Parser/default-file/environment negatives, driver pre-dispatch rejection and owned certificate fixture lifecycle passed. The integration unit connected successfully but its added TLS observation queried restricted statistics as an application role, yielding no rows; inspect these facts through the existing fixture administrator instead of broadening grants. Separately, the full ACL check found the new `application_crypto_format` composite row type retains PostgreSQL's default PUBLIC usage because migration 64 omitted the established explicit revocation. S23d is IN_PROGRESS before its new migration; S23c pauses for that prerequisite. Boundary/JSON PASS `20261005T185159Z-p88464` / `-p88290`; generation/format PASS `20261005T184901Z-p73750` / `20261005T185114Z-p82792`.

S23d validation checkpoint: generation PASS `20261005T185437Z-p13045`; drift/Markdown PASS `20261005T185538Z-p21996` / `20261005T185539Z-p22291`. PostgreSQL service slice FAIL `20261005T185539Z-p22074` (2/3, cleanup 6/6): complete ACL assertions now pass, but new TLS assertions expected `CN=...` instead of PostgreSQL's observed `/CN=...`; correct that test expectation. Migration owner FAIL `20261005T185539Z-p22069` (7/9): a second frozen schema-hash expectation needs the new migration digest, and operator unit fixtures still provide now-rejected implicit-default placeholder DSNs before their injected database double. Update those current fixture inputs; preserve historical negative tests. The new complete migration digest is `5111d8c8b6d5d70fdba0e4c664bb980daadccfa51b105976096cf02cf2c20061`; migration 64 remains unchanged.

S23d DONE: migration 67 revokes PUBLIC usage on the singleton's implicit row type without rewriting migration 64 or widening runtime/Recovery grants. Authored owner allocation and generated migration history are current; complete schema digest is `5111d8c8b6d5d70fdba0e4c664bb980daadccfa51b105976096cf02cf2c20061`. Pinned complete PostgreSQL service slice PASS `20261005T185842Z-p76645` (3/3, cleanup 6/6), including all purpose identities, exact privileges, actual TLS 1.3/client DNs and wrong-client/root rejection. Pinned migration owner PASS `20261005T185842Z-p76644` (9/9, cleanup 12/12), including catalog, real fresh initialization and operator evidence. Observation uses the fixture administrator's existing statistics visibility; no monitoring grant was added. Current unit placeholders/hash expectations were updated; historical failure identities remain intact. Artifact policy PASS `20261005T185842Z-p76594`; final format PASS `20261005T185806Z-p71315`; diff PASS. Prior failed migration run cleanup 12/12 completed. Resume S23c certificate replacement and affected caller validation; S23/S23a/S24 remain incomplete.

S23c renewed PostgreSQL owner PASS `20261005T190251Z-p31242` (5/5, cleanup 8/8), including real certificate replacement; harness lifecycle row PASS `20261005T190251Z-p31243` (1/1, cleanup 3/3). Auth service-backed PASS `20261005T190251Z-p31260` (26/26, cleanup 38/38). Recovery service-backed FAIL `20261005T190251Z-p31269` (18/19, cleanup 29/29): one metadata fixture reconstructed a purpose DSN from pgx's deliberately sanitized construction string, losing its certificate bindings. Use the retained fixture database's authoritative purpose DSN instead; the other Recovery rows, including real restore/browser execution, passed. Format PASS `20261005T190211Z-p25639`. These remain component/caller results, not final package acceptance.

S23c DONE: closed explicit certificate DSNs, TLS 1.3 with verified chain/name, distinct certificate identities, password-disabled roles and pgx authentication pre-dispatch rejection pass. Driver construction suppresses inherited `PG*`, home passfiles and default certificate bindings; default-file FIFO tests prove no fallback reads. Actual SQL/pool purpose connections and trusted certificate renewal pass; wrong-client/root/replacement negatives reject. Final Recovery metadata slice PASS `20261005T190908Z-p71321` (3/3, cleanup 6/6), repairing the fixture-only DSN reconstruction; the earlier full Recovery failure remains distinct. Drift/Markdown PASS `20261005T190908Z-p71250` / `-p71474`; Markdown coverage excludes this handoff; diff PASS. Previously recorded complete PostgreSQL, migration, Auth and harness lifecycle results supply the other required exits. S23d remains a separate completed privilege prerequisite. Production package provisioning, other adapters, listener transport and S23a activation are still pending; next dependency is remaining S23 integrations.

S23e first pinned provider slice FAIL `20261005T191613Z-p9387` (0/1 grouped unit): real HTTPS OIDC RS256/PS256/ES256, key rotation, unsupported inputs and signed RSA-SHA256 SAML succeeded; five new malformed-certificate fixtures retained a parsed RSA certificate signature-algorithm field while asking the ECDSA fixture CA to sign. Clear that test-template field before issuance; do not weaken production admission. Generation/format PASS `20261005T191547Z-p709` / `20261005T191603Z-p4048`; boundary/JSON/drift PASS `20261005T191613Z-p9564` / `-p9332` / `-p9334`. The source now explicitly rejects certificate-bearing JWKs before the pinned JOSE decoder's unconditional SHA-1 thumbprint computation; provider migration requires public RSA/P-256 JWK representations.

S23e checkpoint: final pinned provider/manifest slice PASS `20261005T192501Z-p57662` and strict diagnostic provider row PASS `20261005T192501Z-p57669` (each 1/1, cleanup 3/3). Earlier expanded passes `20261005T191829Z-p25381` / `-p25382` also cleaned 3/3; the initial failed fixture run cleaned 3/3. Auth enterprise route/lifecycle slice PASS `20261005T192008Z-p27466` (3/3, cleanup 6/6). Artifact policy PASS `20261005T192008Z-p27399`; final format/drift/Markdown PASS `20261005T192341Z-p52045` / `20261005T192501Z-p57595` / `20261005T192502Z-p57894`; diff PASS. These checks cover admitted algorithms, key/trust replacement, claims, replay/binding and excluded inputs, but source inspection uncovered an additional dispatch prerequisite: go-jose PSS verification accepts automatic salt length, and Go TLS performs chain cryptography before the current policy callback. The selected archive requires an even RSA modulus bit length and approved exponent; PS256 additionally requires hash-length salt. S23f is IN_PROGRESS before changes, and S23e remains IN_PROGRESS pending that prerequisite and renewed verifier acceptance. No source-only success is promoted to completed integration qualification.

S23f initial implementation checks PASS: pinned complete cryptography owner `20261005T193651Z-p78989` (1/1 grouped unit), provider/manifest slice `20261005T193651Z-p78985` (1/1) and complete PostgreSQL owner `20261005T193651Z-p78995` (5/5). Boundary PASS `20261005T193651Z-p79181`; generation/format PASS `20261005T193541Z-p70148` / `20261005T193615Z-p73562`. TLS now admits explicit destination, peer and immutable trust-anchor parameters before full standard-library chain/name/EKU verification; the internal Go flag suppresses only the duplicate earlier verifier, with no operator bypass. PS256 uses the supported JOSE opaque-verifier hook with exact SHA-256 salt length. Add independent Wycheproof public vectors and final strict diagnostics before closing the prerequisite. Source fixture identity: `rsa_pss_2048_sha256_mgf1_32_test.json` SHA-256 `7f6efafc160f4816b96cbf1c12188a31051d7e3f001e27505d9edb5f2a0e325c`, cases 1/62/108; no runtime or test downloads.

S23f DONE: fixed RSA parameter admission and SHA-256 PKCS#1/PSS services pass independent public vectors and adversarial checks. Supported JOSE dispatch enforces the exact 32-byte PS256 salt. TLS admits peer and immutable root parameters before mandatory standard X.509 chain/name/EKU/time verification; destination identity is explicit, and trust-pool metadata cannot widen captured trust. The system bundle is bounded and frozen per process, with unsupported anchors excluded before path construction; explicit private roots reject unsupported entries. No operator verification bypass or global trust modification exists. Final pinned cryptography owner PASS `20261005T194015Z-p20316` (1/1, cleanup 3/3); strict RSA/TLS PASS `-p20319` (1/1, cleanup 3/3); strict enterprise provider/manifest PASS `-p20338` (1/1, cleanup 3/3); full PostgreSQL service slice PASS `-p20355` (3/3, cleanup 6/6); targeted security scan PASS `-p20756` (4/4, cleanup 6/6). Generation/format PASS `20261005T193831Z-p11218` / `20261005T193940Z-p14779`; JSON/drift/Markdown PASS `20261005T194319Z-p88293` / `-p88285` / `-p88508`. Markdown excludes this handoff; diff PASS. No new failure occurred. Certificate/trust replacement requires restart; S23e resumes for its final owning-route check. S23 and production/package enforcement remain incomplete.

S23e DONE after S23f: the final pinned Auth enterprise route/lifecycle slice PASS `20261005T194742Z-p96377` (3/3, cleanup 6/6), and generated-artifact policy PASS `20261005T194742Z-p96332`. The renewed strict provider/manifest result in S23f covers real HTTPS RS256/PS256/ES256, fixed PSS salt, trust/key replacement, signed SAML and adversarial provider admission. Existing issuer/audience/correlation/replay checks remain owner-local. Public JWK-only signing representations, HTTPS endpoint equality, bounded responses and current RSA signing certificates are deliberate fresh-provider compatibility requirements; unsupported representations reject before excluded operations. Prior fixture failure remains recorded. Diff PASS; prior final Markdown check excludes the handoff. Next is S3 integration. Inspection found MinIO bucket creation can compute MD5 on region correction even with SHA-256 upload settings; the current v7.3.0 API retains that path. Record and resolve S23g before progressing to other S23 adapters.

S23g first build FAIL `20261005T195236Z-p19991`: extraction of the S3 implementation also removed the filesystem adapter's `sort` import. Restore that import; the SDK compiled and dependency sums were produced through `GOFLAGS=-mod=mod make build-migrate`. This is an intermediate compile check, not acceptance; no services were acquired.

S23g paused for S23h: second `make build-migrate` compiled successfully, but its receipt wrapper failed because `NODE_BIN` was unset. The nested compiler summary `20261005T195446Z-p24220` says pass; the enclosing Make command exited 2 and is recorded as FAIL. Do not use that compiler-only summary as complete build evidence. Format PASS `20261005T195513Z-p27392`.

S23h DONE: all five public Go artifact recipes now declare the pinned Node runtime as a prerequisite/cache input and pass its path explicitly to the receipt-producing build wrapper. Ordinary migration build PASS `20261005T195616Z-p33396`, pinned build PASS `20261005T195637Z-p39797`, and pinned cache reuse PASS `20261005T195719Z-p56005` with matching receipt/binary SHA-256 `6a37e8b63264de142f30a8363e5bd439990a038f3588a566d56bdc7df6438328`. Full cryptographic-policy assessment PASS `20261005T195643Z-p43645` (3/3, cleanup 5/5), including the existing wrapper regression and all three actual binary receipts. Toolchain drift PASS `20261005T195617Z-p34376`; `make harness-command-surface-contract` exited 0. Markdown PASS `20261005T195727Z-p58453` (handoff excluded); diff PASS. Discovery mistakes `OWNER=harness` and direct `harness-smoke-cache-artifact` returned usage/no-target errors; the correct owner is `harness.command_surface`, and its build-wrapper check is executed by the registered cryptographic assessment. The earlier enclosing receipt failure remains a failure despite its successful nested compiler summary. Resume S23g; this is build infrastructure, not S23a activation or final acceptance.

S23g unit compile FAIL `20261005T195757Z-p60490`: the SDK error-type conversion accidentally replaced the fixture readiness helper block, causing missing helpers and invalid return types. Restore those existing lifecycle helpers exactly and change only their typed error inspection. No service acquisition occurred. This editing failure remains distinct from pending live S3 qualification.

S23g first renewed unit PASS `20261005T195957Z-p68038` (1/1, cleanup 3/3). Live managed-service slice FAIL `20261005T195957Z-p68041` (2/3): object-store readiness timed out during listing on the running native-TLS container. Canonical diagnostic: `_shared/test-services/a2e637efea484fd39a601141/service-scope.json`, 32 attempts (25 operation timeouts, 7 transport failures). Its readiness diagnostic records cleanup failed; inspect the enclosing cleanup and owned-resource ledger before claiming clean completion. Investigate the actual listener/SDK path; no plaintext or verification fallback is authorized.

S23g readiness diagnosis: the pinned SeaweedFS entrypoint drops from root to `seaweed`; copied 0600 identities were still root-owned, so its certificate watcher could not supply a handshake identity. Fixture construction now assigns only those two owned identity files to the image account before running the unchanged entrypoint, retaining private permissions. The failed run's outer cleanup completed 5/5 and its ledger contains no retained object-store resource; the inner readiness-cleanup failure remains recorded. Native TLS and certificate verification remain mandatory.

S23g next failures: native-TLS parent readiness now succeeds, but managed-service child `20261005T200552Z-p99347` failed because the fixture broker discarded its isolated `SSL_CERT_FILE` binding; admit that one owned environment field through service and browser projections. Fixture row `20261005T200552Z-p99345` failed compilation on one remaining MinIO `BucketExists` test call; use the new SDK helper. Both runs cleaned 6/6. Certificate/configuration rejections now fail fixture admission immediately instead of spending the transient-readiness deadline. Boundary PASS `20261005T200552Z-p99479`.

S23g native TLS and child trust projection now pass: managed-service owner slice `20261005T201140Z-p64570` (3/3) and fixture admission/lifecycle row `20261005T201140Z-p64569` (3/3). Expanded checksum/multipart/presign adversarial execution is still pending; these results do not close S23g.

S23g paused for S23i: `make generate` FAIL `20261005T201447Z-p9671` because the new probe row names a Go package outside the schema/catalog's exact allowed roots. The task-guide lookup also rejected that invalid candidate. Temporarily withdraw only this newly added row to restore catalog discovery, consult the owning guide, then amend the exact root policy and restore the row. No unrelated routing or previous evidence is removed. The original generation failure remains recorded.

S23i discovery also rejected unsorted new row identifiers and the streaming row's mismatched family prefix; both are corrected in authored routing. Generation FAIL `20261005T202053Z-p15783` then identified unsorted probe test selectors. Sort those authored selectors before regeneration; no generated file is hand-edited.

S23i routing generation PASS `20261005T202122Z-p19244`; format PASS `20261005T202137Z-p22439` after the earlier pre-generation graph admission failed (`20261005T202105Z-p19004`, no canonical result). Catalog owner PASS `20261005T202143Z-p26517` (1/1), JSON shapes PASS `20261005T202158Z-p30795` (3/3), Markdown PASS `20261005T202158Z-p30976` (handoff excluded). Newly routed probe FAIL `20261005T202143Z-p26516`: compilation found a stale short declaration after SDK conversion in bucket cleanup. Correct that declaration and rerun the exact row. The preceding managed S3/fixture passes each completed cleanup 6/6.

S23i DONE: exact `tools/objectstoreprobe` admission is now consistent across selector schema, runner registry and catalog containment. The existing three CORS checks run through the object-store owner. Repeat probe `20261005T202417Z-p65088` still failed on the same stale declaration because the first edit did not match; actual corrected probe PASS `20261005T202440Z-p65598` (1/1). Generation drift PASS `20261005T202442Z-p65847` (4/4); catalog owner, JSON and Markdown results above remain applicable. Cleanup: catalog 3/3, probe failures 3/3 each, corrected probe 3/3, drift 6/6. `git diff --check` passes. No runtime compatibility changed; only the exact tool package is admitted. Resume S23g with qualified streaming/presign execution and remaining probe transport migration before S23a.

S23g strict policy unit PASS `20261005T202546Z-p71060` (1/1). Strict service diagnostic FAIL `20261005T202546Z-p71066` before service startup; outer cleanup 4/5 with services_close failed and a retained pending runtime lease. The service ledger records no started database/object-store resource; this does not itself discharge the lease. Source inspection found fixture-only SHA-1 namespace hashing and testcontainers Docker-config MD5; S23k must establish the executed cause and recovery. Pinned enabled-mode live run FAIL `20261005T202703Z-p93655` (2/3): all five checksum-metadata assertions fail and presigned changed-body input is accepted; failed multipart abort and cancelled-upload checks pass. S23g pauses for S23j. No fallback or final acceptance is claimed.

S23j candidate: upstream 4.17 source does not implement additional SHA-256 checksum handling in PUT/HEAD. Released SeaweedFS 4.48 (2026-09-28) has explicit SHA-256 header/query/trailer validation and retained checksum metadata. Registry index bytes independently hash to `sha256:4e61d15fd35994cb1e43e1e553dff106794841fd9a99ade2fc8c8bfce4d7872d`; its linux/amd64 manifest is `sha256:aba492e2a4e4c90bff795745e8e660affa1f09e7650f5981bd7bccd1a06cd931`. Advance authored fixture/dev/package/probe pins together and qualify this candidate with unchanged integrity assertions. The Go module archive selection/digest remain unchanged; this is not package acceptance. Sources: https://github.com/seaweedfs/seaweedfs/releases/tag/4.48 and the matching tagged S3 PUT implementation.

S23j DONE: the user explicitly confirmed upgrading SeaweedFS to latest stable during this work. The independently verified latest stable 4.48 is pinned by index digest in dev/package Compose, test fixtures, local lifecycle defaults and probe identity defaults. First 4.48 live PASS `20261005T203031Z-p15919` (3/3); expanded unchanged-integrity and explicit bad-checksum/part rejection PASS `20261005T203214Z-p43117` (3/3). Empty, known/unknown-length, multipart and presigned transfers retain SHA-256 metadata; changed-body/method/expired presigns reject, original bytes survive a rejected overwrite, incorrect PUT/part digests reject before publication, interrupted multipart abort and cancellation leave no object. Both passing runs clean 6/6; the 4.17 failed run also cleaned 6/6. Toolchain drift PASS `20261005T203048Z-p28117` (2/2); format PASS `20261005T203205Z-p37945`; Markdown PASS `20261005T203215Z-p43442` (handoff excluded); diff PASS. No persisted operator state was upgraded or reset, and no package was published. Historical 4.17 failure/identities remain retained. Go module/archive pins are unchanged. Next resolve S23k, then resume S23g's remaining adapter/probe checks; full package acceptance remains S24.

S23k begins after the harness.browser guide. Read-only Docker inspection corrects the earlier incomplete cleanup inference: exited container `40ae358fba9c` belongs to strict run `20261005T202546Z-p71066`, and exited container `80896a7169d9` belongs to original TLS failure `20261005T195957Z-p68041`. Both carry exact managed/run/suite/service labels. Empty retained ledgers did not prove their absence; both failures remain failures. Unrelated long-running development containers are outside this cleanup. The strict run also retained its private runtime directory and certificate material. Establish the fixture-only hash failure and discharge these exact resources; do not erase historical evidence.

S23k correction: disposable suite/process namespace suffixes now use SHA-256. With that sole fixture hash change, the complete live S3 row passes under `GODEBUG=fips140=only` and the pinned module at `20261005T203642Z-p72424` (3/3, cleanup 6/6); no container with that run label remains. This identifies the fixture namespace SHA-1 as the reached diagnostic obstruction. The separate testcontainers Docker-config MD5 is non-product tooling source inventory, not a production exception or a claim that every possible harness operation passes strict mode. The broker now captures bounded stderr only to classify known excluded-hash panics safely, without retaining/printing raw panic arguments; its regression verifies private text cannot escape.

Exact failed-run environment cleanup was performed after matching managed/run/suite/service labels and stopped state: `docker rm --volumes` removed `40ae358fba9c` and `80896a7169d9`. The strict run's exact private runtime was removed after matching its owner UID/run/lease marker, proving its owner process inactive and confirming zero matching containers. These were owned-environment repairs, not validation reruns; historical failed summaries and cleanup records remain unchanged. Unrelated development containers were preserved. Format PASS `20261005T203616Z-p67084`; harness diagnostic regression remains pending before S23k closes.

S23k harness catalog FAIL `20261005T203642Z-p72423` (96/97 internal cases): generated topology input fingerprints are stale after the fixture-broker implementation edit. Run normal generation before renewing the affected harness check; the strict live S3 pass is independent of this projection failure. Markdown PASS `20261005T203820Z-p7507` (handoff excluded).

S23k DONE: normal generation PASS `20261005T203909Z-p9730`; renewed harness catalog PASS `20261005T204254Z-p13805` (1/1 routed row, 97/97 internal cases), including safe early-panic classification and retained-secret scanning. Cleanup 3/3; strict live S3 evidence and exact failed-resource repairs remain as recorded above. No production compatibility or policy exception was introduced. The original failures remain failures. Resume S23g after this checkpoint; S23a and package acceptance remain pending.

S23g resumed after its object-store task guide. The maintained SDK treats size as a hint and automatically enlarges part buffers; the adapter now owns exact-length admission and a fixed 8-MiB/10,000-part transport bound with two concurrent part requests. Known oversize rejects before body consumption; unknown/mismatched streams must abort before publication. Core 04 records this intentional S3 capacity limit separately from Recovery filesystem framing. New live checks cover short/extra bodies, multipart/boundary mismatches and huge size hints. Validation pending.

S23g pauses for S23l: the existing CORS proxy test package is not admitted by the public catalog. Preserve its behavior/process-proof tests and establish exact routing before changing TLS. The ongoing upload-size row is independent validation of the already authored S23g change.

S23l DONE: exact helper root added to registry/schema/catalog; existing eight proxy tests routed through Object Store. Generation PASS `20261005T204701Z-p55908`; proxy row PASS `20261005T204751Z-p59621` (1/1, cleanup 3/3); harness catalog PASS `20261005T204751Z-p59612` (1/1, cleanup 3/3); JSON PASS `20261005T204751Z-p59555` (3/3, cleanup 5/5); drift PASS `20261005T204751Z-p59551` (4/4, cleanup 6/6); Markdown PASS `20261005T204911Z-p89476` (handoff excluded); diff PASS. No unrelated helper root or runtime compatibility admitted. S23g upload bounds also PASS under pinned strict diagnostic mode at `20261005T204620Z-p33649` (3/3, cleanup 6/6). Resume S23g TLS migration after this checkpoint.

S23g TLS candidate: the existing CORS proxy now uses qualified TLS for listener, upstream and ownership-health checks, a separate fixture private key, immutable certificate/root fingerprints and v2 lifecycle projections. No TLS-only proxy was added; retained CORS and pidfd proof semantics remain. Strict policy/probe/proxy unit PASS `20261005T205213Z-p12326` (3/3). First HTTPS compatibility FAIL `20261005T205214Z-p12726` (2/3, cleanup 6/6): SDK HEAD maps a denied credential to `Forbidden`, absent from the previous MinIO error mapping. All reached storage/CORS/presign checks pass. Correct owner/probe classification, and use the body-bearing GET error when distinguishing a missing bucket from a missing object; HTTP HEAD cannot carry that S3 error distinction. Historical failure retained. Generation `20261005T205146Z-p3825` and format `20261005T205205Z-p7201` passed before this correction; renew after v2 projection changes.

S23g DONE: one supported AWS S3 connection/transfer policy replaces MinIO; HTTPS and explicit SHA-256 checksums cover initialization, known/unknown streams, multipart, typed presigns, fixtures and the public probe. Exact size admission prevents SDK hints from increasing buffers; failed streams abort without publication. The existing CORS proxy uses separate TLS identity, verified upstream/health, v2 lifecycle records and unchanged process ownership. Core 04, harness owner and development guidance describe bounds, binding/renewal and incompatibility. SeaweedFS 4.48 remains digest-pinned; no retained development deployment was changed.

Final S23g evidence: full pinned strict Object Store slice PASS `20261005T205447Z-p43303` (8/8, cleanup 13/13); pinned service slice PASS `20261005T205447Z-p43318` (5/5, cleanup 10/10); public HTTPS `seaweedfs-compatibility` PASS `20261005T205447Z-p43664` (3/3, all 14 compatibility cases pass, probe cleanup clean/no retained keys, outer cleanup 6/6, zero matching Docker containers). Boundary PASS `20261005T205447Z-p43488` (3/3); generation `20261005T205412Z-p34742` and format `20261005T205432Z-p37915` pass. JSON `20261005T205607Z-p9024` (3/3), drift `20261005T205607Z-p9014` (4/4), targeted security `20261005T205607Z-p9413` (4/4), artifact policy `20261005T205652Z-p49715` (3/3), scripts `20261005T205652Z-p50096` (2/2), fixture/lifecycle rows `20261005T205652Z-p49797` (4/4) and harness catalog `20261005T205652Z-p49789` (1/1) all pass with every recorded cleanup completed. Markdown PASS `20261005T205607Z-p9319` (configured coverage excludes this handoff and development guide); diff PASS. Historical failures and repairs remain unchanged. S23 remains IN_PROGRESS: telemetry, application/configuration transport, complete browser/deployment fixtures and WSL2 scheduling still need implementation before S23a; package and final acceptance are not claimed.

S23m candidate: endpoint/configuration admission now requires HTTPS; exporters receive shared-policy immutable TLS with no ambient proxy or redirect, and owned HTTP transport closes after provider shutdown. Removed the stale `v1.41.0` exporter identity; maintained library APIs supply each signal's version. Owner text explicitly preserves the supported gRPC library's mandatory `grpc-go/VERSION` suffix rather than replacing protocol internals. Initial default telemetry slice PASS `20261005T210154Z-p91874` (2/2); this default selection does not execute all release-tier owner rows. The new real HTTP/gRPC signal matrix and complete explicit owner selection remain pending.

S23m validation correction: the default owner run above contained all 10 pre-existing rows grouped into two execution units; the prior inference that release-tier rows were omitted was incorrect. Explicit 11-row candidate FAIL `20261005T210408Z-p2708` (0/2 grouped units, cleanup 4/4) because the new log fixture used the pre-upgrade `log.StringValue` API; the selected OTel version uses `attribute.StringValue`. Correct the fixture and rerun. This is a new-test compile failure, not collector qualification evidence. Added explicit expired-certificate and redirect rejection scenarios before renewed validation.

S23m real collector FAIL `20261005T210537Z-p8768` (1/2 grouped units): positive peers were rejected because the fixture replaced `SSL_CERT_FILE` inside a process that had already captured immutable system trust. This confirms the required restart boundary; do not weaken it for tests. Move each collector/trust scenario into a fresh test subprocess with isolated CA material. Boundary `20261005T210538Z-p9329`, JSON `20261005T210538Z-p9162`, Markdown `20261005T210538Z-p9354` passed; renewed collector evidence remains required.

S23m pauses for S23n after collector FAIL `20261005T210709Z-p19457` (1/2 grouped units): valid TLS now works, but the metric HTTP library reports stale version 1.43.0 inside its selected v1.45.0 module, and the later gRPC no-proxy option overwrites the User-Agent option. All certificate rejection scenarios pass. Establish executable dependency identity and compose supported gRPC options in one call. Security `20261005T210718Z-p21194` and drift `20261005T210719Z-p21335` pass but do not replace failed collector acceptance.

S23n candidate now reads actual per-protocol exporter module versions from `debug.ReadBuildInfo`, rejects absent/replaced/duplicate metadata, and composes User-Agent/no-proxy gRPC options once. Actual pinned strict collector row PASS `20261005T210912Z-p68356` (1/1): all HTTP/gRPC traces/metrics/logs reach their qualified collectors with actual v1.45.0/v0.21.0 identities; wrong root/name, expiry, TLS 1.2 and redirects reject. Add independent malformed-metadata checks and renew generation/owner validation before this prerequisite closes.

S23n DONE: runtime identity comes from each actual selected module in executable build metadata, with missing/replaced/duplicate identities rejected. One supported gRPC option composition retains both application identification and proxy rejection. Independent metadata cases and real strict HTTP/gRPC signals PASS in the full 11-row telemetry owner selection at `20261005T211045Z-p78688` (2/2 execution groups, cleanup 4/4). Generation `20261005T211007Z-p70056`, format `20261005T211029Z-p73314`, drift `20261005T211045Z-p78607` (4/4, cleanup 6/6), JSON `20261005T211045Z-p78626` (3/3, cleanup 5/5) and targeted security `20261005T211045Z-p79015` (4/4, cleanup 6/6) pass. Earlier failed collector runs cleaned 4/4 each; the first corrected actual transport row cleaned 3/3. Corrected owner text makes executable dependency metadata authoritative, preserving historical unsuccessful Version-helper evidence above. Resume S23m after this checkpoint.

S23m DONE after its resumed owner guide: all 11 telemetry owner rows pass on the completed S23n candidate, including real TLS 1.3 HTTP/gRPC traces, metrics and logs, correct per-module User-Agent, wrong-root/name/expired-certificate/TLS 1.2 rejection, plaintext admission failure and no redirect forwarding. Positive evidence uses isolated process trust and no verification bypass. SDK privacy, retry, queue, secret-redaction and activation/containment tests remain covered. Source policy now requires HTTPS, verified immutable system trust, bounded owned transports, no proxy/redirect fallback and reverse cleanup. JSON, generation/drift, module-boundary and security evidence above pass; diff PASS. Remaining S23 work is configuration/application HTTPS/WSS, browser/deployment trust provisioning and WSL2 operational scheduling. Mandatory application cryptographic admission remains S23a; package/final acceptance remain S24/S25.

S23m final Markdown PASS `20261005T211208Z-p23132` (handoff excluded). S23o begins after the telemetry completion checkpoint; all existing dirty work and historical tracker content remain preserved.

## 11. Open prerequisites and production boundaries

There is no unresolved user preference preventing this plan. Planning choices do not establish adoption, compliance or authority to mutate a retained deployment. TODO here means required future evidence/work; BLOCKED is reserved for an actual contradiction or failed prerequisite encountered during execution.

| ID | Prerequisite / question | Why it matters | Required owner / evidence | Status |
| --- | --- | --- | --- | --- |
| RP-PRE-01 | Prior coordinated reporting/accounting amendments | Completed S11 requirement cleanup remains valid history | Original owner changes and S11 checkpoint | DONE |
| RP-PRE-02 | Prior package release routing and cleanup | Three smokes already participate in the historical package release | Original S09/S15 results, including `20261005T070744Z-p19802` | DONE |
| RP-PRE-03 | Formal owner adoption and actual deployment suitability | Engineering success does not adopt a draft or approve a customer environment | Existing owner-status process and applicable deployment/security/licensing acceptance; ordinary records, no new dossier | TODO |
| RP-B02 | Cryptographic security policy and Reference Pack source amendment | Existing exclusion cannot be ignored; signature identity remains stable | S16 coordinated source candidate and service/exception definitions; formal adoption separate | DONE for source/design; not formal adoption |
| RP-B03 | Exact module/service/build/runtime and WSL2 engineering disposition | Certificate association or a Debian-derived image alone is insufficient | S17/S22/S23/S24 actual identities, supported service/purpose mappings and current security disposition; WSL2 package evidence and explicit separate formal-claim status | DONE for WSL2 engineering; final release remains S25 |
| RP-B04 | Pinned-module and diagnostic Make/harness routing | Floating selection, stale caches and unrouted assessment undermine qualification | S17 authored build/cache/routing policy, S24 canonical release acceptance | DONE for WSL2 engineering; final release remains S25 |
| RP-PRE-04 | Fresh-state admission and incompatible-format detection | Rejection must precede mutation for all three binaries and recovery | S16 contract; S17/S19–S21 implementation; S24 real-binary negative evidence | DONE for WSL2 engineering; final release remains S25 |
| RP-PRE-05 | PBKDF2 cost and authenticator compatibility | Fixed cost must remain operationally bounded; SHA-256 TOTP clients must work | S19 complete; fixed-profile assessment `20261005T163057Z-p36924`, lifecycle/client evidence and authenticator guidance above | DONE |
| RP-PRE-06 | Enabled integration service usage and production transport | SCRAM, S3, OIDC/SAML, TLS/telemetry and optional paths may have unsupported uses/defaults | S23 adapter-owned positive/negative evidence and supported module-service/purpose mapping; unresolved use blocks qualification | DONE for WSL2 engineering; final release remains S25 |
| RP-PRE-07 | WSL2 environment feasibility | Establish actual host/guest/daemon and executable validation route during S16, then qualify the package in S24 | S16b inspected target, doctor, owner/source review and complete route; package evidence remains S24, customer acceptance separate | DONE for feasibility; native-target blocker superseded |
| RP-CLAIM-WSL | Formal CMVP applicability to the exact WSL2 stack | Engineering acceptance cannot assert a formal regulated deployment claim | Applicable security authority plus exact module-policy/environment evidence; neither Windows-native listings nor generic distro listings suffice alone | TODO; unestablished, outside engineering completion dependencies |
| RP-ENV-NATIVE | Pure/native Linux deployment and qualification | Avoid restoring an unwanted platform requirement through release or handoff checks | Explicit future user scope decision; preserve useful Linux guest/container mechanics | DEFERRED indefinitely; no dependency or deadline |

RP-B02–RP-B04 were expanded from historical DEFERRED constraints; RP-B02 now has a completed source/design disposition while RP-B03/RP-B04 now have completed engineering dispositions with final release validation retained in S25. Historical entries are not rewritten. Prior RP-B01 owner contradiction and RP-B05 pre-production release blocker remain resolved in their original iteration. Do not reopen them merely because another capability is planned.

| Readiness dimension | Current status | Completion authority / evidence |
| --- | --- | --- |
| Previous structural/package implementation | DONE | Historical S01–S15/prerequisite exits and original releases; no regulated claim. |
| New single-policy cryptographic implementation | IN_PROGRESS; S16–S24 and every implementation prerequisite complete; final validation is S25 | S25 finalization, fast checks and one complete final-candidate release. |
| Windows 11 / WSL2 package engineering acceptance | S24 PASS; complete final release pending in S25 | Frozen three-smoke and capacity results `20261006T001353Z`; exact platform, binary/module and cleanup receipts. S25 renews acceptance through the complete release. |
| Formal CMVP operating-environment applicability | TODO; unestablished, not claimed | Applicable security authority and exact stack/module evidence; separate from WSL2 engineering completion. |
| Pure/native Linux deployment and qualification | DEFERRED indefinitely | Only an explicit future scope decision reopens this work; it is not an S16–S25 dependency. |
| Specification adoption for production | TODO; not claimed | Coordinated owner-status process; the Reference Pack 0.2.0 candidate remains draft at this planning point. Tracker/test results cannot adopt it. |
| Customer deployment approval | TODO; outside this implementation authorization | Acceptance of the actual target, external dependencies, trust/clock inputs and applicable security/licensing requirements by their owners. No automatic rollout. |

Promotion remains blocked wherever required owner adoption or actual deployment acceptance is absent, even after WSL2 engineering acceptance. Formal regulated claims additionally require established CMVP environment applicability. Those separate dispositions and indefinite native-Linux deferral do not keep a completed WSL2 engineering iteration open. A failed implementation, unsupported application security service or failed WSL2 check cannot be reclassified as an external responsibility.

## 12. Binary completion criteria

**The completed planning-only update used these criteria (historical):**

- Only this controlling tracker changed during this step; all pre-existing working-tree changes remain intact and the baseline is correctly described as dirty `e7a1497` with completed remediation.
- The exact 197-file Reference Data inventory is preserved and verified; the inspected adjacent application/authentication/encryption/integration/build boundaries are recorded with owners, callers, contracts and test posture.
- Active sections contain S16–S25 as separate TODO workstreams, the complete dependency chain, phase risks/exits, and per-slice remediation, areas, rationale/benefit, compatibility, unresolved risks, validation and binary exit.
- S07 is explicitly expanded; prior completed workstreams, failures and original validation identities remain historical without relabeling.
- All behavior changes are marked as requiring later implementation authorization; this task changes no product behavior, generated input/output, conformance or release evidence.
- The appended session records `make lint-markdown`, `git diff --check`, supplemental tracker checks, actual results and skipped checks. RP-P03 is marked DONE only after they pass.

**The currently authorized implementation is complete only when:**

- S16–S24, S16c, S23a and every active implementation prerequisite are DONE and S25 completes last, with this tracker updated after each workstream before the next. RP-ENV-NATIVE remains DEFERRED indefinitely; RP-CLAIM-WSL remains a separate claim disposition and neither is an engineering completion dependency.
- The coherent owner candidate, typed projections, build/runtime policy and approved primitive/service usage agree; no owner contradiction or unsupported enabled security service remains hidden.
- Fresh-state admission rejects incompatible state and backups before mutation; no legacy crypto reader/writer, password converter, automatic reset or fallback remains in the new production policy.
- Pack signatures, canonical identities, trust rotation, immutable provenance and other retained domain/security guarantees pass the actual verifier and lifecycle tests.
- Actual Windows/WSL2/guest/daemon/package identities are recorded, full disposable scenarios and cleanup pass, and one complete final-candidate WSL2 release supplies canonical evidence. Historical/partial runs or native-Linux runs are not substitutes.
- WSL2 failure/restart handling, key/certificate management across trust stores, guest storage, fresh provisioning/recovery and exact-module/platform update procedures are actionable. Implementation completion, WSL2 package acceptance, formal CMVP applicability, specification adoption and customer deployment approval each have an explicit disposition.
- Missing adoption or actual deployment acceptance blocks the corresponding promotion; unestablished formal applicability blocks regulated claims. Engineering completion does not authorize retained-state mutation or rollout. Close S07 only after S25, explicitly for the WSL2 scope.

S23o foundation candidate: configuration v3 now requires HTTPS and explicit normalized TLS certificate/key paths. Server admission captures and validates its identity before runtime acquisition; every ordinary/inherited listener requires TLS. Config slice PASS `20261005T212004Z-p37221` (2/2 execution groups); HTTP runtime PASS `20261005T212004Z-p37223` (1/1), including inherited-listener drain and no-TLS pre-publication rejection. Generation `20261005T211914Z-p28332` and format `20261005T211952Z-p31852` pass. An attempted server slice used nonexistent row `app.server.unit.server_runner` and failed routing before execution; no acceptance or cleanup evidence is inferred. Use the authored exact runner/lifecycle rows for renewed checks. Real process clients still require explicit isolated trust injection; S23o remains IN_PROGRESS.

S23o actual pinned process candidate PASS `20261005T212251Z-p49138` (7/7 execution units, two selected owner rows) covers HTTPS readiness and authentication/session behavior. Backend boundary PASS `20261005T212251Z-p49258` (3/3). Broader Auth FAIL `20261005T212251Z-p49145` (24/38 passed, 12 failed, 2 skipped): retained browser startup stderr explicitly reports `application.public_origin` / `invalid_origin` because existing browser fixtures use HTTP. This is related integration work, now tracked separately as S23p; it is not a product-policy exception or a browser pass. Earlier exact server runner/lifecycle slice PASS `20261005T212032Z-p40563` (1/1). Generation after new process-admission/boundary rows PASS `20261005T212422Z-p40655`. Renew process acceptance on the expanded candidate and inspect cleanup before S23o completion.

S23o expanded pinned `backend-process` PASS `20261005T212535Z-p49847` (16/16; 20 semantic rows and retained process-helper tests, cleanup 27/27). Actual binaries cover TLS 1.3, HTTPS/WSS, missing certificate/key rejection before root mutation, untrusted/TLS 1.2/plaintext rejection, process restart and cleanup. Full server service slice FAIL `20261005T212535Z-p49829` (16/18, cleanup 27/27) solely in the known plaintext browser support acquisition and its summary; S23p owns that prerequisite. Previous Auth browser failure cleanup completed 39/39. Exact `cartulary.test-services.run-id` queries show zero remaining containers for all three runs; an earlier query used the wrong label key and is not cleanup evidence.

S23o broader unit candidate `20261005T212840Z-p87505` exposes a new fixture import cycle (`cryptography` tests → PKI fixture → new transport helper → `cryptography`) and stale server/operator fake credential inputs. Move the qualified HTTP fixture to the separate `tlstest/transport` subpackage, preserving PKI issuer independence; no production API changes. S23o pauses for S23q's explicit credential-fixture repair before closing. Strict HTTP runtime PASS `20261005T212603Z-p32844` (1/1, cleanup 3/3); config PASS `20261005T212603Z-p32818` (2/2, cleanup 3/3); drift PASS `20261005T212603Z-p32696` (4/4, cleanup 6/6); targeted security PASS `20261005T212603Z-p33195` (4/4, cleanup 6/6); boundary `20261005T212841Z-p87624` and generated-policy `20261005T212840Z-p87448` pass. Markdown `20261005T212809Z-p85292` passes with configured handoff exclusion; diff passes. Renew affected tests after corrections; no broad or final acceptance is claimed.

S23q discovery correction: completed broad unit run `20261005T212840Z-p87505` failed 5/164 execution units (159 passed), cleanup 164/164. In addition to server/operator fixtures and the new TLS-helper cycle, migration facade tests retain the same obsolete DSN shape; include them in S23q. Audit compilation also references Recovery's removed legacy helper, separately recorded as S23r before further changes. The earlier Auth browser failure cleanup count is **41/41**, correcting the hand-counted 39/39 above; all outcomes completed. Do not infer individual row failures solely from a failed Go execution group; targeted reruns will resolve each assertion.

S23q focused server startup PASS `20261005T213414Z-p37095`, migration facade PASS `20261005T213414Z-p37109`, and repaired shared cryptography strict slice PASS `20261005T213414Z-p37121`. Operator requeue assertions pass, but its combined group FAIL `20261005T213414Z-p37098` because the migration-evidence test bans the substring `secret`, now present in the legitimate migration 65 filename, and freezes the pre-64–67 manifest's complete output digest. Replace the stale password sentinel with the actual admitted DSN, hostname and encoded/decoded certificate locators; preserve redaction checks. Review the four additive migrations and current manifest, then advance the current exact-byte expectation from historical `9925ebfb54dc9883d3fca57c01b9c0512c4a2cadd32d4ab18cf0f3e5d810890f` to `9cdc19f52ff1be7ceb04f0c46d39d0baaa7825c2b3e37f014c369ed85d2e1e47`. The production evidence serializer and applied migration history are unchanged; the original identity and failed run remain recorded here.

S23q DONE: all affected server, migration and operator facade assertions now execute past explicitly valid test credentials to their intended failure/ownership boundaries. Migrate uses a non-connecting `database/sql` connector rather than a plaintext network DSN for its fake handle. Server startup PASS `20261005T213414Z-p37095`; migration facade PASS `20261005T213414Z-p37109`; operator CLI/evidence PASS `20261005T213526Z-p40629`; each run has 1/1 execution unit and cleanup 3/3. Exact-byte current migration evidence and locator redaction pass after the reviewed fixture correction; no production behavior or applied SQL changed. The independent TLS-helper cycle repair also passes the pinned strict cryptography slice `20261005T213414Z-p37121`. Format PASS `20261005T213525Z-p40418`. Broad backend-unit acceptance remains failed at its original identity; S23r is the remaining independent compile prerequisite before S23o's renewed broad validation. Begin S23r after this checkpoint and its owning guide.

S23r DONE: audit regression now asserts its current authoritative, required PostgreSQL Recovery contribution through `administrativeaudit.RecoveryStateContribution`, including owner, restore mode and codec identity. The removed legacy snapshot helper remains test-only in Recovery. Pinned audit unit/service slice PASS `20261005T213643Z-p52277` (4/4, cleanup 7/7); module boundary PASS `20261005T213643Z-p52405` (3/3, cleanup 5/5); format PASS `20261005T213619Z-p46708`; Markdown PASS `20261005T213643Z-p52436` (configured handoff exclusion). No runtime compatibility or applied history change. S23o resumes with broad backend-unit validation; browser/development prerequisite S23p remains pending.

S23o DONE: renewed pinned `make backend-unit` PASS `20261005T214029Z-p76548` (164/164 execution units, cleanup 167/167). This closes the separately recorded fixture/cycle/audit failures without rewriting their original outcomes. Configuration v3, pre-resource immutable TLS identity admission, mandatory ordinary/inherited HTTPS listeners and per-process verified clients are implemented. Actual HTTPS/WSS process evidence remains `20261005T212535Z-p49847`; the subsequent helper package move is covered by renewed unit compilation and strict primitive checks. JSON, generation/drift, security, generated-policy and module boundaries passed as recorded above; diff and historical-tail equality pass. Public endpoint/configuration compatibility intentionally breaks; fresh certificate bindings are required. Known full Auth/server browser runs remain failed until S23p. Next: S23p, marked IN_PROGRESS after this checkpoint; `make task-guide ROLE=module-author OWNER=harness.browser` consulted.

S23p first HTTPS browser candidate FAIL `20261005T214827Z-p29869` (9/11 execution units; cleanup 13/13 completed). Both real backend/frontend HTTPS readiness probes passed; terminal startup evidence still used the HTTP-only diagnostics v2 schema, preventing publication. Advance current diagnostics to v3 together with HTTPS stack v8, generation v2 and explicit fixture trust identity; old schema consumers reject. No browser acceptance is claimed. Pinned NSS tool preparation PASS `20261005T214549Z-p19577`; archive SHA-256 `e5b42390b02c21851bc04e9557ec48539a913f9b09637220c4b1db0d1386b3fb` is from Ubuntu’s official `libnss3-tools` 2:3.98-1ubuntu0.2 amd64 download metadata. It is extracted only in owned containers; no host/global trust or package installation. Generation PASS `20261005T214750Z-p21224`, format `20261005T214814Z-p24758`, JSON `20261005T214828Z-p30137` precede this diagnostics correction and require renewal.

S23p renewed browser candidates `20261005T215101Z-p73332` and `20261005T215230Z-p12357` FAIL (9/11 each) after successful HTTPS stack publication. The latter’s bounded NSS diagnostic identifies archive-read permission denial: inherited private-process umask reduced Docker-copied public tool/CA files to 0600, owned by root in the container. Explicitly set those public files to 0644 inside their private 0700 staging directory before copying; private keys never enter the container. Renderer contract slice PASS `20261005T215231Z-p12648` (1/1). Renew actual browser execution; do not infer qualification from mocked lifecycle checks.

S23p browser sign-in PASS `20261005T215336Z-p51684` (11/11) after public-copy permissions were corrected. Independent browser trust/TLS row PASS `20261005T215810Z-p4541` (11/11): Chromium accepts the isolated authority and rejects wrong authority/hostname; explicit TLS 1.2 probes reject at both frontend and backend listeners. PKI/service-fixture/renderer contract rows PASS `20261005T215811Z-p4919` (5/5). Generation attempt `20261005T215625Z-p91612` failed because a newly introduced tool test package was outside the closed Go selector roots; move reusable fixture behavior/tests into `internal/testutil/browserpki`, leaving the tool as a thin composition root. Format and typecheck attempted during that invalid routing candidate rejected before execution. Renewed generation PASS `20261005T215726Z-p95784`, format `20261005T215759Z-p99022`. Typecheck `20261005T215811Z-p4987` failed on an explicitly undefined Vite HTTPS option; omit absent optional fields. Renewed typecheck PASS `20261005T215900Z-p65174` (2/2); shell lint PASS `20261005T215939Z-p79536` (4/4), JSON `20261005T215939Z-p79337` (3/3), boundary `20261005T215939Z-p79530` (3/3). Auth/frontend/lifecycle checks are ongoing; exact trust-bundle attachment binding was strengthened during the broader intermediate run and needs renewed frozen-candidate validation. S23s/S23t record remaining consumers/provisioning separately before implementation; S23a remains blocked.

S23p checkpoint: frontend-unit PASS `20261005T215902Z-p65503` (691/691, cleanup 693/693). Broad pinned Auth FAIL `20261005T215937Z-p78732` (24/26 execution units; one failing browser row; cleanup 38/38). The trace proves the final username was entered while session bootstrap was loading, cleared by the anonymous gateway transition, then rejected locally; no final login request occurred. Record S23u before fixing this owner issue. S23p TLS/login successes `20261005T215810Z-p4541` and `20261005T215336Z-p51684` each cleaned 14/14, as did the two failed NSS runs. Lifecycle FAIL `20261005T220033Z-p29639`: current web-e2e lifecycle passes, then seeded-review test fails on obsolete cleanup `cause` assertion; no canonical cleanup-results file was emitted by this direct runner. S23s now explicitly includes that assertion and the missed SHA-1 seeded TOTP helper. Toolchain drift PASS `20261005T220854Z-p69944`, script lint PASS `20261005T220857Z-p74804`; generation drift FAIL `20261005T220855Z-p72309`, pending artifact inspection. An attempted narrow rerun passed `GODEBUG` as a Make input and was rejected before execution; repeat with environment assignments. Auth and harness guides consulted before S23u. No final acceptance claimed.

S23u DONE: the ordinary gateway disables username/password input while the initial session observation is loading, preserving its existing secret retirement behavior. The delayed-session unit row passes `20261005T221026Z-p27405` (2/2); the affected real HTTPS browser row passes `20261005T221006Z-p96906` (11/11) after reproducing its failure at `20261005T220838Z-p46162`. Typecheck PASS `20261005T221005Z-p96347` (2/2); format `20261005T220956Z-p90432` and generation `20261005T221004Z-p95695` pass. The earlier drift failure was solely the topology render index after the last authored harness edit; generation refreshed it through Make. No API or state compatibility change. Resume S23p with renewed full Auth and fixture evidence; S23s remains the explicitly separate seeded-review lifecycle prerequisite.

S23u reopened before S23p completion: broader Auth `20261005T221113Z-p39745` fails 24/26 on the deferred-session accessibility row, which deliberately requires keyboard-reachable credential controls while loading. The earlier sign-in fix passed its focused exit but disabled those controls. Use read-only loading fields instead: they remain keyboard reachable and named while preventing credential entry until the anonymous controller is ready. Preserve the existing accessibility assertions and rerun both scenarios; no test weakening or timeout increase. TLS trust row PASS `20261005T221113Z-p39755` (11/11), renewed generation drift `20261005T221113Z-p39649` (4/4), generated-policy `20261005T221113Z-p39679` (3/3), Markdown `20261005T221113Z-p40006` and diff pass. Source review also reconciled the older nonvisual-host-browser sentence with the new isolated-renderer requirement; Markdown coverage still excludes the handoff.

S23u DONE after renewal: loading fields are read-only, not disabled, preserving named keyboard access while preventing input before anonymous initialization. Full pinned Auth PASS `20261005T221443Z-p46412` (26/26 execution units, cleanup 38/38), including the original sign-in failure and unchanged deferred-session accessibility assertions. Delayed-session unit PASS `20261005T221443Z-p46416` (2/2, cleanup 4/4); typecheck PASS `20261005T221443Z-p46564` (2/2). Format `20261005T221425Z-p41031`, Markdown `20261005T221443Z-p46635` pass. Earlier DONE/reopen and failure identities remain intact. Return to S23p after this checkpoint.

S23p DONE: fresh private PKI, independent frontend/backend keys, verified HTTPS readiness/proxy/application hops, stack v8/generation v2/diagnostics v3, exact attachment CA digest and pinned isolated NSS trust are implemented. Browser trust negatives pass at `20261005T221113Z-p39755` (11/11); full Auth is the renewed frozen implementation candidate above (26/26, cleanup 38/38). Frontend-unit `20261005T215902Z-p65503` remains the broad 691/691 baseline; the subsequent gateway change has renewed owning unit/browser/type evidence. Generation drift `20261005T221113Z-p39649`, generated-policy `20261005T221113Z-p39679`, toolchain/script/shell/JSON/boundary and Markdown results are recorded above. The owned web-e2e lifecycle section passes in `20261005T220033Z-p29639`; its later seeded-review assertion still fails, explicitly owned by S23s rather than relabeled as a passing composite run. Development launch now requires persisted keys and explicit certificate bindings; provisioning remains S23t. No global trust changes, certificate bypasses, intermediate package publication or retained-state conversion. Next S23s; S23a remains blocked until the integration umbrella and its prerequisites finish.

S23s implementation checkpoint: seeded preparation and review now acquire independent pinned renderer containers with the admitted fixture trust bundle before requests. Seeding uses the existing Playwright request boundary in that trusted process; actor authentication stays on the public Auth API. Seed TOTP now requires canonical 32-byte secrets and SHA-256, checked against RFC 6238 vectors. A typed renderer proof records daemon/image/name/random-token ownership before Docker creation, supports exact interrupted recovery and rejects mismatches. Seeded diagnostic attachment uses a separately published host-loopback port for the same owned browser, preserving the existing upstream CLI and page invariants. Dev review remains externally selected/live-unattested; it does not qualify application transport. The cleanup test now checks aggregate failures and preserves primary errors. Generation PASS `20261005T222024Z-p6804`, format PASS `20261005T222051Z-p10109`; actual seeded/lifecycle/contract execution is in progress. S23s remains IN_PROGRESS and S23t has not begun.

S23s DONE: verified seeded enrollment/seeding, editor/viewer sign-in, current stack attachment, diagnostic connection to the same owned page, private capture/report lifecycle and cleanup pass for both profiles. Initial default PASS `20261005T222054Z-p11386`; Network Flow claimed PASS `20261005T222201Z-p51705`. Renewed pinned default PASS `20261005T222530Z-p94918` adds actual controller SIGKILL, exact resource recovery, retained `session_lost` failure/complete-cleanup receipt, repeated-stop identity and private-root removal; its two real fixture cleanup/failure-recovery cases also pass. No live/stopped renderer containers remain after the checks. Ownership tests reject changed daemon, token, name, image, ambiguous IDs and failed removal, and retain pre-acquisition proof. Review contract PASS `20261005T222113Z-p30774`, execution PASS `20261005T222115Z-p32226`, lifecycle PASS `20261005T222112Z-p29280`, renewed public workflow/diagnostic PASS `20261005T222531Z-p95463` (7 tests). Composite `make run-harness-smoke-lifecycle` PASS `20261005T222053Z-p10564`, including the previously failing seeded aggregate assertion and current dev/web-e2e lifecycle. Script lint `20261005T222159Z-p50377`, JSON `20261005T222200Z-p50770`, generation drift `20261005T222532Z-p95748`, Markdown `20261005T222348Z-p37695`, and diff pass. Markdown excludes the handoff; historical tail equality passes. Compatibility: current fixture CA/stack and SHA-256 TOTP only; host trust and canonical golden identity unchanged. Next S23t. No package/final release readiness is claimed.

S23t IN_PROGRESS after the S23s checkpoint and release/PostgreSQL/object-store/Recovery guides. Provisioning will use explicit purpose-specific certificates, fixed certificate-only PostgreSQL host rules, isolated local bootstrap, native SeaweedFS TLS and HTTPS application inputs. Preserve nonroot application ownership and mount only each purpose’s credential files. Existing development containers/volumes are untouched; use fresh disposable Compose projects for engineering validation. Scheduling will check Docker Desktop availability and reuse existing `backup inspect latest`/`restore-verify due` owner commands on operator restart, with no Windows startup automation or guest `docker.service` dependency.

S23t checkpoint: shared certificate-only PostgreSQL provisioning, per-purpose read-only mounts, native SeaweedFS TLS, HTTPS package/dev inputs and operator-started timer/restart scripts are implemented; acceptance is still pending. Generation PASS `20261005T223726Z-p30083`, format PASS `20261005T223820Z-p33499`. Shell lint FAIL `20261005T223820Z-p33492` on obsolete unused host/port defaults after explicit-DSN migration; remove them and rerun. First disposable pinned package attempt FAIL `20261005T223830Z-p39505` (10/11 units), PostgreSQL initialization exited before app start. Original cleanup receipts remain canonical; add bounded owned-service failure logs before rerunning, rather than inferring the cause from Compose’s dependency error. S23t remains IN_PROGRESS; no package qualification is claimed.

S23t renewed package FAIL `20261005T224011Z-p15879` (10/11), with clean workspace/source/destination receipts. Bounded logs identify the TLS wrapper’s 077 umask making PostgreSQL’s intermediate parent root-only before upstream drops privileges. Restore upstream 022 only after the explicitly protected TLS copy; retain 0700/0600 service key modes. Shell lint PASS `20261005T223939Z-p14617` (4/4); Markdown PASS `20261005T224030Z-p57944`, configured root/spec coverage excludes this handoff and package/development guides. No user containers or volumes changed.

S23t package FAIL `20261005T224147Z-p94076` (10/11), all three owned cleanup receipts pass. PostgreSQL now initializes and real certificate-authenticated migration succeeds. SeaweedFS rejected the newly added unqualified `-telemetry=false` spelling; verify the pinned 4.48 source and use its supported `-master.telemetry=false`, disabling unused Lance/Iceberg listeners explicitly. Use local `gosu postgres` for health observation to avoid root/peer mismatch logs. Lifecycle PASS `20261005T224031Z-p58253` (dev/web-e2e suites); generation PASS `20261005T224247Z-p68264`. Dev-service tests now require an explicit certificate DSN and verify missing identity rejects before Docker acquisition and ordinary migration never reprovisions retained state.

S23t package PASS `20261005T224347Z-p71960` (11/11, all owned receipts clean): real certificate migration, native HTTPS object-store initialization, app readiness/assets, WebSocket Origin rejection and packaged Reference Pack import/replay. Recovery attempt FAIL `20261005T224459Z-p65106` (10/11): a fixture string replacement accidentally altered `RESTORE_VERIFY_POSTGRES_DB` and duplicated the TLS-directory binding. Compose created empty root-owned directories despite `create_host_path=false`; source resource cleanup passed but workspace cleanup failed. Correct the authored fixture, extend bounded workspace reclamation to those owned directories, and preserve the original failed receipt. Manually remove only the exact observed empty directories via the pinned, project-labeled helper; record the separate retry result without relabeling historical acceptance. Harness contracts PASS `20261005T224349Z-p72479`; release contracts including restart/freshness/cleanup scenarios PASS `20261005T224441Z-p56318`. Script lint `20261005T224441Z-p55595`, shell `20261005T224441Z-p55684`, drift `20261005T224538Z-p27871`, JSON `20261005T224538Z-p27892`, artifact policy `20261005T224538Z-p27894` and boundary `20261005T224557Z-p45913` pass. Attempted `module-boundary-check` was rejected as nonexistent before execution; the discovered public `backend-module-boundary-check` passed. Diff and historical-tail equality pass.

S23t recovery FAIL `20261005T224724Z-p46993` (10/11), cleanup passed completely. Real backup creation/latest inspection/due restore verification all succeeded; the trailing public-route probe still used Node HTTP for HTTPS. Replace that fixture transport with explicit scoped CA and TLS 1.3. Actual logs also revealed `compose start app` restarts completed dependencies; restart only the exact previously running container ID, rejecting ambiguous identity. Original failed workspace at `20261005T224459Z-p65106` now has a separate `workspace-cleanup-retry.json` confirming removal of only its observed empty directories; original failed receipt is unchanged. Reference Pack fixture TOTP now validates canonical 32-byte secrets and uses SHA-256’s final-byte truncation, with an independent RFC 6238 vector. S23t stays IN_PROGRESS pending renewed frozen-candidate checks.

S23t DONE: current Compose/dev inputs use distinct certificate purposes, verified PostgreSQL TLS 1.3, NULL application passwords, fixed HBA/mappings, private service key copies and native HTTPS SeaweedFS 4.48. Ordinary startup does not reapply provisioning. Application mounts stay nonroot and purpose-scoped. Timers are explicitly started, independent of guest docker.service; restart checks use owner backup-age and due-verification commands. Backup cleanup restores only the exact previously running app container. Package PASS `20261005T224347Z-p71960` (11/11); renewed operational Recovery PASS `20261005T225004Z-p96480` (11/11, cleanup 13/13); full Reference Pack PASS `20261005T224854Z-p22355`, then renewed frozen candidate PASS `20261005T225149Z-p77536` (11/11, cleanup 13/13 and clean workspace/source/destination receipts). The latter covers all 11 named scenarios in `standup-reference-pack-smoke/artifacts/scenarios.json`, including SHA-256 MFA, actor/replay/rotation, historical rendering, destination portability, actual matching-package restore, collection and content-loss rejection. Renewed operator/cleanup contracts PASS `20261005T225006Z-p96996`; shell `20261005T225103Z-p73872`, scripts `20261005T225103Z-p73860`, Markdown `20261005T225212Z-p25563`, diff and historical-tail equality pass. Applicable generation/drift/JSON/boundary/policy evidence is recorded above. All historical failures and original identities remain. Fresh-only compatibility; no user development state or host trust changed. S23v is the newly recorded admission-packaging prerequisite. S24 will repeat qualification on fresh final state with Windows boundary/identity/performance evidence; no final or formal CMVP acceptance is claimed here.

S23 DONE after S23t: PostgreSQL dispatch guard/certificate purpose, admitted OIDC/SAML key and signature parameters, S3 SigV4/SHA-256 operations, verified application/dependency/telemetry HTTPS, real WebSockets and isolated browser trust all have their owning positive/negative execution evidence in S23b–S23u. Plaintext/password production configuration is removed; non-security protocol hashes retain their narrow owner justification. Enabled integration engineering closes RP-F27; actual facade enforcement remains open RP-F21 in S23a. Before activation, finish separately discovered S23v so migration inspects genuinely fresh shared package storage. Windows-client/package qualification and formal applicability remain separate S24/disposition work. No earlier failure is relabeled.

S23v IN_PROGRESS after the S23t/S23 completion checkpoints and release owner guide. Exclude only source sentinels from the package context, retain empty nonroot-owned directories and expose the actual source/target object-store and shared read-only filesystem bindings to migration. Do not weaken InspectEmpty or stamp intermediate state.

S23v DONE: `deploy/mvp/Containerfile.dockerignore` excludes only source `.keep` sentinels; shipped runtime roots remain empty and owned by nonroot. Migration receives the same four authoritative app volumes read-only and explicit verified source S3 credentials; verification-target migration receives the target S3 binding and read-only target-root bind. Package checks assert no image-root contents and exact shared volume identity/read-only flags using mount-only inspection. Pinned package PASS `20261005T225615Z-p61204` (11/11, cleanup 13/13 and clean owned receipts); operational Recovery PASS `20261005T225616Z-p61448` (11/11, cleanup 13/13 and clean owned receipts), proving nonroot write/publication survives the empty-root change. Generation `20261005T225601Z-p58236`, drift `20261005T225640Z-p50553` (4/4), shell `20261005T225640Z-p51189` (4/4), release contracts `20261005T225641Z-p52722` and diff pass. No validation failure in this prerequisite. No state conversion or freshness exception added. Next S23a, with RP-F21 still open until actual mandatory facade/build admission passes; S24/S25 retain final package/release acceptance.

S23a IN_PROGRESS after S23v's completion checkpoint. Cryptography, server, migrate, operator and database-migration guides consulted. Activate executable admission before resources in every facade; wire migration-owned exclusion/identity with authoritative read-only storage inspection, and require current identity before ordinary leases/bootstrap/jobs/publication. Make default supported builds and application-running tests pinned; diagnostic ordinary/wrong identities remain isolated negative fixtures. Do not add a runtime policy bypass or stamp preexisting state.

S23w marked IN_PROGRESS after S23a inspection: shared fresh fixture and template initialization still bypasses the format initializer. No production facade activation has been implemented yet. Preserve migration-scratch mechanics; finish and validate this prerequisite before continuing S23a.

S23w validation history: initial generation failed `20261005T230710Z-p27226` because the authored row list was not ASCII sorted; corrected owner input and regenerated PASS `20261005T230744Z-p30344`. Three initial Make-argument test invocations rejected unregistered `GOFIPS140` before execution (no canonical run); use the established environment input. Pinned retries `20261005T230815Z-p39387` / `-p39391` failed the test-service build on a now-unused migration import; `-p39393` also exposed a previously unrouted mock lifecycle test's missing service declaration. Removed the unused import and declared the mocked capability; no production admission weakening. Corrected runs are pending.

S23w DONE: new database-only fixture initialization uses the production migration admission boundary and independent exclusion handle. Full fresh fixtures and test-service templates use it; historical migration-scratch capabilities retain low-level migration replay. Pinned `make test-slice OWNER=module.database_migrations` PASS `20261005T230911Z-p76928` (9/9), PostgreSQL owner PASS `20261005T230913Z-p77495` (6/6), and complete PostgreSQL service-backed owner PASS `20261005T230912Z-p77202` (4/4, including admitted fresh/template fixtures). Drift PASS `20261005T230929Z-p7958` (4/4); Markdown PASS `20261005T231020Z-p45151` (configured root/spec coverage excludes this handoff); diff PASS. Cleanup is complete in all passing runs; no deployment state was stamped or reset. Earlier failed generation/build/mock-declaration runs remain unchanged. Next dependency: resume S23a mandatory facade/default-build activation; S24/S25 remain incomplete.

S23a activation is implemented and validation is IN_PROGRESS: all three facades admit actual execution identity, server inspects current state before leases/storage/telemetry/bootstrap, migration holds fresh initialization exclusion through read-only authoritative storage checks and format commit, and every operator PostgreSQL path plus bucket initialization requires current state. Make defaults and build receipts now require the exact module; ordinary staging receipts are removed. Common secret/configuration material is resolved before server acquisition; normative leased extension claim resolution remains in its original stage. Initial migrate unit run `20261005T231316Z-p53277` failed an outdated private callback signature; repaired. Initial operator run `-p53279` failed the bucket-init process fixture's intentionally unused database URL; the now-required admitted database and PostgreSQL routing are supplied. Server owner PASS `-p53278` (26/26). Initial cryptographic assessment `20261005T231431Z-p36867` failed the build-helper fixture's missing GO variable; repaired through the existing launcher fallback. Assessment PASS `20261005T231628Z-p48502`, renewed PASS `20261005T231807Z-p23225` (3/3) after binding the FIFO probe to the actual `CARTULARY_CONFIG_FILE` input. All three actual facades reject disabled mode, ordinary/wrong-module builds and missing/contradictory metadata before reading the FIFO configuration. Fixture binaries are isolated and removed; state/package acceptance remains separately pending.

S23x marked IN_PROGRESS before implementation. S23a package `20261005T231708Z-p63310` failed at migration configuration admission (10/11 units); read-only volume mounts correctly cause the shared writable-root access check to reject. Package-owned cleanup receipts remain the authoritative outcome; no retained deployment changed. S23a owner reruns PASS: migrate `20261005T231705Z-p62380` (4/4), operator `20261005T231706Z-p62658` (11/11), server `20261005T231707Z-p62956` (26/26), operator service `20261005T231901Z-p37989` (8/8), server service `-p38001` (18/18), toolchain `-p37930` (2/2), shell `-p38217` (4/4). S23a remains incomplete pending this prerequisite and package rejection evidence.

S23x DONE: explicit structurally admitted read-only configuration inspection shares canonicalization/type/overlap checks while requiring only read/search access. Migration uses this API; writable serving validation is retained. Core 04 clarifies the capability distinction. Config owner PASS `20261005T232156Z-p19571` (2/2), migrate owner PASS `-p19580` (4/4), fresh package PASS `-p19545` (11/11; cleanup 13/13 plus package/source/destination workspace receipts passed). Markdown PASS `-p19793`, configured docs root/spec coverage excludes this tracker and package README; diff PASS. No S23x run failed. Resume S23a final actual-binary retained-state negatives and integration/build checks before S24.

S23a package/build validation PASS: package `20261005T232525Z-p23719` (11/11), Reference Pack `20261005T232749Z-p12616` (11/11), Recovery `-p12275` (11/11). Each source package rejects disabled mode and incompatible/missing identities through actual image binaries; database counts and filesystem digests remain unchanged on rejection, and tampered disposable state is destroyed rather than repaired. Exact image-extracted server/migrate/operator receipts bind the pinned compiler/module/archive and binary hashes. Boundary `20261005T232733Z-p4033` (3/3), harness `-p4093` (2/2), drift `-p3650` (4/4), shape `-p3682` (3/3), generated policy `-p3696` (3/3), shell `-p4149` (4/4) pass. S23y marked IN_PROGRESS before implementation: inspection found the S19 assessment's old password/HTTP provisioning path, which must be replaced before staging machinery can be considered retired. S23a remains IN_PROGRESS and RP-F21 stays open until this prerequisite passes.

S23y DONE: the public credential assessment now runs the shipped certificate-authenticated/TLS package with exact execution/state admission, frozen two-CPU/two-GiB cgroup limits and verified TLS client. Existing package ownership handles cleanup; the old password/HTTP staging composition is removed. Initial current-package PASS `20261005T233411Z-p87940` (3/3) depended on warm artifacts; inspection found the graph dependency incomplete, corrected explicitly to deployable-shape before acceptance. Final PASS `20261005T233951Z-p97558` (9/9; cleanup 11/11 plus source/destination/workspace receipts passed): p95 352.055021 ms, maximum excess-work rejection 43.72293 ms, exact 10 accepted / 54 rejected. Cgroup observations, current image and binary receipts and host metadata are canonical under `credential-capacity-assessment/artifacts`. Generation PASS `20261005T233933Z-p93347`, harness PASS `20261005T233951Z-p97167` (2/2, cleanup 4/4), drift `20261005T233950Z-p96754` (4/4), scripts `20261005T233951Z-p97176` (2/2), shell `-p97234` (4/4), Markdown `-p97251`, diff and historical-tail review pass. No failing execution in S23y; the initial warm routing limitation remains recorded. Configured Markdown coverage excludes this handoff. Next resume S23a completion checkpoint.

S23a DONE after S23y: all three production facades and default supported application/test builds now require the pinned compiler/module and enabled mode. Actual binary/configuration admission precedes read-only state admission, and state admission precedes leases, storage creation, bootstrap, jobs and publication. Migration alone establishes the format through the authoritative fresh initialization exclusion; partial/interrupted retained state rejects. Actual wrong/missing/contradictory/disabled binaries and incompatible/missing retained identities reject with bounded diagnostics; rejection preserves observed state, borrowed ownership and reverse idempotent cleanup. Original S17 production acceptance is discharged by the component, owner, actual binary and package identities recorded in the preceding checkpoints, including all three fresh package smokes and their cleanup. RP-F21 CLOSED. No runtime bypass, compatibility stamping, legacy format reader or ordinary production build remains. All failures retain original identities. Intermediate state is disposed; S24 must qualify fresh package state and the Windows boundary, then S25 validates one complete final candidate. Next S24.

S24 IN_PROGRESS after the S23y and S23a completion checkpoints. Release and Recovery task guides consulted. Extend existing Make-owned package routes with a Windows executable using isolated trust, certificate replacement and owned-service interruption, actual WSL2/daemon/filesystem/network identities and required full-release routing. Never reboot the host, shut down WSL, change global trust/firewall or reuse intermediate retained state. Formal CMVP applicability remains unestablished.

S24 initial package validation: all three Windows HTTPS clients pass isolated chain/name/expiry/TLS-version checks. Recovery PASS `20261005T234524Z-p85197` (11/11). Base package FAIL `-p85174` (10/11): the new renewal probe incorrectly assumed SeaweedFS has a Compose healthcheck; use PostgreSQL health plus real object-store initialization as the service readiness boundary. Reference Pack FAIL `-p85813` (10/11): new confinement test expected exit 1, but the documented operator runtime-failure status is 3; fix the assertion, retaining the valid-archive symlink rejection and no-publication checks. Both failures are qualification-helper errors and require fresh reruns. Generation PASS `20261005T234504Z-p76723`, release-routing contract PASS `20261005T234717Z-p26464` (1/1), scripts `20261005T234524Z-p85507` and shell `-p85516` pass. These initial runs overlap additions to the qualification scenarios and are not frozen final-candidate acceptance.

S24 renewal checkpoint: Windows authenticated WSS, origin rejection and valid-archive symlink confinement execute successfully in the renewed Reference Pack run. Recovery PASS `20261005T234856Z-p33682` (11/11); current fixed-profile assessment PASS `-p35033` (9/9); harness PASS `-p34422` (2/2). Base package FAIL `-p33671` (10/11) because the new certificate-purpose assertion inspected stderr for a raw PostgreSQL error; the server correctly emits its bounded admission diagnostic through stdout. Correct the observation channel and assert the bounded owner code. Earlier failed packages and Recovery all have complete canonical cleanup and passed private-workspace/service receipts. No production authentication fallback or diagnostic disclosure is introduced.

S24 certificate rerun FAIL `20261005T235211Z-p49351` (10/11): corrected certificate-purpose rejection passed, then a subsequent Compose command failed with an insufficiently specific helper diagnostic. Add bounded stage/reason observations before rerunning; no failure is promoted. Full Reference Pack PASS `20261005T234856Z-p34930` (11/11, all 13 scenarios). Its Windows HTTPS/WSS receipt asserts the actual Windows process and isolated trust. Renewed capacity p95 373.630968 ms / maximum rejection 49.067246 ms. The Recovery package is additionally extended to interrupt only its owned database at an observed real snapshot read, assert no successful publication, then exercise backup/restore after restart. Final qualification remains pending.

S24 interruption PASS `20261005T235607Z-p39201` (11/11): the actual backup reaches a blocked snapshot read, its owned database is stopped, operator exits 4, no backup metadata or successful journal is published, and the same database/application restart before real backup and due verification succeed. Base package FAIL `-p39193` (10/11): renewed migration credentials pass, but renewed object-store initialization fails; record text-form bounded reason codes as well as JSON codes to isolate this remaining integration boundary. Shell/scripts PASS `-p39457` / `-p39444`. S24 remains incomplete.

S24a marked IN_PROGRESS before production configuration changes. The final bounded diagnostic identifies S3 retry exhaustion after recreation, while renewed migration and deliberate certificate-purpose rejection pass. Upstream tagged source ties the detected address to persisted Raft identity. Introduce an explicit stable service identity and verify this inference using the same real replacement scenario; do not weaken TLS, reset storage or reinterpret the previous failures. S24 pauses until this prerequisite exits.

S24a stable-identity candidate FAIL `20261006T000414Z-p79369` (10/11), still at S3 retry exhaustion after credential renewal. The stable-address change alone does not resolve the observed failure, so the earlier source-based diagnosis is not confirmed. Retain bounded service state and filtered startup diagnostics in the next isolated run to distinguish resumed-service readiness from storage/transport failure. Generation `20261006T000400Z-p76415`, toolchain `20261006T000416Z-p79744`, shell `-p79954`, Markdown `-p79943` pass; package cleanup remains part of failure inspection. S24a stays IN_PROGRESS and blocks S24.

S24a diagnostic FAIL `20261006T000731Z-p63526` (10/11): SeaweedFS remains running, recovers the same persisted topology and filer store, but has no Raft leader throughout the observed first eight seconds. S3 requests begin before resumed master readiness. The initial address-only diagnosis was incomplete. Use SeaweedFS's local `/cluster/healthz` readiness check and require healthy dependency admission in authored Compose, preserving the stable node identity and resumed state. This probe is container-local control-plane observation; application S3 still requires verified TLS and SigV4. The generic S3 `/readyz` handler only returns HTTP 200 and is not an adequate backing-store readiness check. Tagged source and exact commit handlers reviewed. No data reset, TLS bypass or indefinite operation retry is introduced.

S24a DONE: authored package/development Compose use a stable SeaweedFS DNS identity and native local master readiness; dependent services require healthy state. Existing storage and resume policy are preserved. Actual package PASS `20261006T001055Z-p44366` (11/11, cleanup 13/13 plus all source/destination/workspace receipts passed). All purpose certificates are replaced, wrong database purpose rejects, renewed migration and Recovery/S3 operations pass, the original bucket remains present (creation is forbidden), the application survives an owned abrupt interruption, and the format identity remains unchanged. Canonical `certificate-replacement.json`, progress and binary/Windows receipts bind this evidence. Generation PASS `20261006T001043Z-p41402`, shell PASS `20261006T001056Z-p44670`; configured Markdown and diff checks pass, historical tracker tail is unchanged. Previous failures retain their diagnostics and complete cleanup; their address-only inference is superseded by the observed master-readiness finding. No user development volumes, Raft reset, legacy conversion, host trust or firewall changes occurred. Resume S24 final fresh-state package/assessment and routing validation, then S25.

S24 DONE after S24a: frozen package validation PASS `20261006T001353Z-p29429` (11/11), full Reference Pack PASS `-p30662` (11/11), operational Recovery PASS `-p29456` (11/11), fixed capacity PASS `-p30745` (9/9). Three smokes each clean 13/13; capacity cleans 11/11; every owned workspace/source/destination receipt passes. Windows HTTPS/WSS, isolated CA/name/expiry/TLS-version rejection, certificate replacement, wrong database purpose, unchanged current format, original-bucket retention, actual snapshot-read interruption with no successful publication, restart/backup/restore and guest filesystem confinement are covered. Exact Windows/WSL/Ubuntu/Docker/daemon/ext4/NAT/CPU and effective limit receipts plus resolved images and all three binary/module/archive receipts are under each target's canonical artifacts. Capacity p95 is 360.608743 ms and maximum excess rejection 75.385803 ms under unchanged two-CPU/two-GiB limits. Full release membership is 1,274 units / 2,048 rows, with both assessments and all three smokes required by readiness; no reduced WSL subset. Harness PASS `-p30199` (2/2), drift `-p29366` (4/4), shapes `-p29452` (3/3), generated policy `-p29457` (3/3), boundary `-p30138` (3/3), toolchain `-p29592` (2/2). Markdown PASS `20261006T001324Z-p26892` for configured root/spec coverage (handoff/package README excluded), diff and direct documentation review pass. Every earlier failure retains its identity; no intermediate state is preserved as final compatibility evidence. Formal CMVP applicability remains unestablished; specification adoption and customer deployment approval are separate. Next S25 finalization and one complete final-candidate release.

S25 IN_PROGRESS after the completed S24 checkpoint. Release owner guide and full target discovery consulted. Finalize operational handoff and generated/maintenance work, then run agent-finalize, test-fast and one complete release-check in order. Inspect canonical projections, source/graph identities, all package/Windows/capacity receipts and cleanup through explain-run. RESULTS_DIR is unset, so retained-run maintenance will be reported skipped. Formal CMVP applicability, specification adoption and customer deployment approval remain separate recorded dispositions.

S25 finalization PASS `20261006T001750Z-p46630` (1/1): generated artifacts unchanged, no rollback required, structural/schema/catalog checks pass. `RESULTS_DIR` is unset; retained-run, canonical-evidence, performance and scheduler maintenance are explicitly skipped with `results-dir-not-provided` in `unit-artifacts/finalize-summary.json`. Markdown PASS `20261006T001824Z-p51332` for configured root/spec coverage; handoff/package README remain excluded and directly reviewed. `make test-fast` is executing as `20261006T001823Z-p51074`; no full final release has yet been accepted. Current handoff preserves the superseded S16 MinIO feasibility inference while pointing to completed S23g AWS SDK v2 execution evidence.

S25 fast validation PASS `20261006T001823Z-p51074`: 744/744 units, zero failures/skips/cancellations, 746/746 cleanup steps completed; `make explain-run` inspected the canonical run. Source digest `sha256:bf89c18c765ef6856edc1b9740d4fb38474fac6c1608485c3ed8e878205c7058`, graph `sha256:3c5c29d04022b17d333f1f9cd1ddb211f9002cfda0f85a4769df2b9b326e6dd4`, dirty baseline `df720535b6b6622cab8b3d399e868dc968e0f6c4`. Direct operational-document review removed the obsolete PostgreSQL-16 cutover/reset instructions and corrected Recovery certificate diagnostics; these Markdown-only edits are outside runtime evidence inputs. Diff and historical-tail preservation checks pass. Start one complete final `make release-check`; S25 remains IN_PROGRESS until all projections and cleanup are inspected.

### S25 discovered prerequisites — final release findings

The complete candidate run `20261006T002507Z-p51747` finished with
failed gates. It cannot supply final acceptance. Keep its source frozen until
cleanup completes; do not combine its successful parts with a later run. S25
remains IN_PROGRESS and resumes final validation after these separate slices.

| Workstream | Status / owner | Remediation, compatibility and exit |
| --- | --- | --- |
| S25a — performance roster integrity | DONE; harness.command_surface / evidence accounting | Sort the authored required-target roster after the new capacity target addition; preserve exact membership and measurement bindings. Generate normally; duration-baseline coverage, owner/harness and drift checks pass. No product compatibility effect. |
| S25b — migrated test call sites | DONE; Auth, Network Flow, collaboration | Update the old socket-helper invocation to explicit trusted TLS, avoid copying a key-ring mutex, and expect current cft4 in the actual conflict browser flow. Compile/vet/vulnerability loading and narrow owner/browser rows pass; no legacy fallback or changed public behavior. |
| S25c — overload telemetry mapping | DONE; platform.telemetry | Project the public authentication-capacity error into its appropriate bounded telemetry class and regenerate affected outputs. Conformance and owner checks pass with complete unique mappings; retain empty public details and no account-existence signal. |
| S25d — performance-fixture trust propagation | DONE; harness.browser / performance fixtures | Carry the actual isolated S3 CA path across the performance clone adapter into its exported environment; retain verified TLS and lifecycle ownership. All four Timeline measurement rows and cleanup must pass with fresh fixtures. Missing trust must fail explicitly. |
| S25e — saved-view selection completion | DONE; Saved Views / browser support | Diagnose why the saved-view flow observes the preceding selection after requesting the system view; repair the responsible owner/helper without loosening authorization assertions. Reproduce and pass the complete selected lifecycle row plus focused regression evidence. |
| S25f — historical migration fixtures | DONE; Entities / Parties / Auth test support | Seed only schema-era inputs when testing earlier immutable migrations; never use a later Auth column or edit applied SQL. Entity and Party historical migration rows must pass their original upgrade and atomic-rejection assertions. |
| S25g — reset identity fixture semantics | DONE; Recovery / harness.browser | Explicitly prepare missing/current marker scenarios now that ordinary fixtures correctly establish current-format identity. Preserve reset's prohibition on stamping or deleting migration identity; both owner rows and cleanup pass. |
| S25h — retained-note recovery navigation | DONE; Notes / browser support | Diagnose the existing retained operation and panel navigation in the response-loss scenario; preserve exact replay bytes and read-only refresh recovery. Repair the responsible lifecycle/helper and pass the complete browser row plus focused support checks. |
| S25i — Reference Pack accessibility recovery | DONE; Reference Data / browser accessibility | Diagnose the absent Retry observation control after keyboard activation; preserve keyboard reachability and observation/cancellation recovery. The complete affected accessibility row and focused owner checks must pass without weakening assertions. |
| S25j — saved-graph refresh observation | DONE; Network Flow / browser support | Make the real saved-graph refresh scenario retain an observable in-flight interval through an explicit request barrier, then release it and prove successful current results. Preserve the last-result visibility assertion and original completion bound; no polling delay or threshold relaxation. |
| S25k — SeaweedFS release Recovery projection | DONE; object-store release / Recovery | Replace obsolete required envelope v2/manifest v3 references with current v3/v4 in authored release checks and dependent fixtures; preserve storage-ref ownership and all 14 real compatibility operations. SeaweedFS release gate, contracts and cleanup must pass against the current registry. |
| S25l — HTTPS object-store fixture origins | DONE; platform.objectstore / S3 test support | Align authored browser-origin fixtures and assertions with verified HTTPS application clients; remove stale mixed HTTP entries rather than adding plaintext support. S3 helper rows, CORS positive/negative execution and testutil integration must pass. |
| S25m — Recovery reset verification ownership | DONE; Recovery / Extensions / test catalog | Route the new reset-identity test through a distinct appropriate row rather than silently expanding the existing exact restore-rejection selector. Preserve original row identities and coverage; generate routing and pass catalog, Recovery and Extensions checks. |
| S25n — visual mention fixture readiness | DONE; Entities / browser visual | Diagnose the missing unresolved-mention control before visual capture. Preserve real mention states and visibility; the complete original visual row must pass without weakening assertions. |
| S25o — saved-view visual differences | DONE; Saved Views / browser visual | Inspect the exact expected/actual/diff artifacts through the maintained UI-review workflow, repair incorrect state or capture readiness, and change goldens only if owner-approved behavior warrants it. Preserve renderer identities and original visual coverage. |
| S25p — observation replay lifecycle | DONE; Indicator observations / browser | Diagnose the missing uncertain-outcome state during the real commit/response-loss transition. Preserve exact replay and no-duplicate-history assertions; the full affected lifecycle row and owner regression checks must pass. |
| S25q — contextual reference selection readiness | DONE; contextual authoring / browser | Diagnose the absent selected-reference removal controls after selecting three candidates; preserve selection identity, keyboard focus and pagination. Pass the complete original browser row and relevant owner regression checks. |
| S25r — Timeline stale-row replay readiness | DONE; Timeline / browser | Diagnose the missing two prepared stale rows before the replay assertion. Establish authoritative read/render completion without relaxing atomic paste, replay or visibility requirements; pass the complete original lifecycle row. |
| S25s — secure restore browser fixture | DONE; Recovery / browser restore | Replace the remaining HTTP origin in the real restore-browser source/target composition with isolated purpose-correct TLS and verified clients; retain fresh-state/format admission and resource cleanup. The actual restore/workbook-query browser row must pass without disabling verification. |

These findings include `go-test-duration-baseline-coverage` (unsorted roster),
`go-vulncheck`/`lint-go` (socket signature and mutex-copy vet), `otel-conformance`
(missing public error mapping), four Timeline setup failures and their dependent
missing summaries, the saved-view browser row, the collaboration cft3 assertion,
three Entity historical migration rows and two reset rows. Final counts and
cleanup will be recorded after the run; no readiness claim follows from a
preliminary count. Resolve S25d first because its failed fixture retirement blocks safe service closure; then execute S25a–S25c and S25e–S25s sequentially, recording each completion before
the next. Any additional distinct failure receives its own prerequisite.

S25 first complete candidate FAIL `20261006T002507Z-p51747`: 1,239/1,274 units pass, 33 fail, 2 dependent skips, zero cancellations, 2,546.252 seconds. `make explain-run` confirms source `sha256:bf89c18c765ef6856edc1b9740d4fb38474fac6c1608485c3ed8e878205c7058`, graph `sha256:03c25ba76ab14dd760a5a1c38d8542c16f9ad233aadd509f10b134a0a0d1b012`, dirty baseline `df720535b6b6622cab8b3d399e868dc968e0f6c4`. All three package projections, cryptographic-policy assessment and credential-capacity assessment pass. They do not override the full failure. Cleanup: 718 completed, 4 failed fixture-acquisition retirements (lease-000009 through lease-000012), 1 blocked service closure. Failed Timeline clone setup and retirement are S25d's immediate priority; inspect exact retained ownership and repair safe idempotent cleanup before proceeding. Historical run results will never be overwritten by later cleanup. Browser owner guide consulted; S25d is IN_PROGRESS, all other newly discovered prerequisites remain TODO. No product source changed during this complete run.

### S25t — Idempotent performance-fixture retirement prerequisite

DONE; harness.browser. Inspection of all four original lease receipts proves bucket cleanup failed after session, database, process and credential cleanup completed. A repeated retirement currently attempts to revoke sessions in a removed database and rewrite an immutable lease. Preserve exact resource ownership and original immutable evidence; treat an already removed database as no remaining sessions, allow identical successful terminal observations without rewriting bytes, and reject any conflicting retained observation. Test repeated cleanup and partial-failure retries; no failed run may become passing evidence. S25d pauses for this separate prerequisite. Browser owner guide consulted.

The abandoned run's two exact owned service containers and private runtime were removed after checking run/suite/lease labels and all closed producer proofs. Separate recovery receipt: `.cartulary/maintenance-recovery/20261006T002507Z-p51747/manual-owned-cleanup.json`. This is manual environment recovery, not canonical qualification. Original failed run and lease artifacts remain unchanged. Only the user's existing PostgreSQL 18.6 and SeaweedFS 4.17 development containers remain; neither was modified. S25d trust propagation and its four-clone regression are authored; `make format` PASS `20261006T011116Z-p38262`. Validation resumes after S25t.

S25t DONE: repeated session retirement accepts only PostgreSQL's missing-database condition; other errors remain failures. Terminal lease observations compare full validated identities/results while preserving original bytes and timestamp. Conflicting or failed-to-successful observations reject. Independent cleanup still runs after failures, and incomplete process cleanup now returns an error even when other resources clean. Existing owner-routed tests cover repeat success, conflicting identity, partial failure/retry and incomplete process closure. Four real isolated clones each retire twice without residue. Unit PASS `20261006T011814Z-p77126` (2/2); real lifecycle PASS `20261006T011656Z-p50315` (3/3, all 6 cleanup steps complete); format PASS `20261006T011759Z-p71989`. Configured Markdown and diff checks pass; handoff coverage remains excluded from lint and directly reviewed. No product compatibility change. S25d resumes IN_PROGRESS with the Timeline guide consulted; run all four measurement rows on fresh disposable state.

S25d measurement candidate FAIL `20261006T011840Z-p79583`: 18/20 units pass, no skips/cancellations; all 27 cleanup steps complete and all four immutable clone leases report complete cleanup. Verified TLS startup and resource retirement are repaired. Typing p95 40.5 ms, selection 32.4 ms and focus 39.2 ms pass their unchanged 100-ms bounds. Blank-row creation p95 175.7 ms exceeds the unchanged 150-ms bound (p50 94.8 ms); its aggregate consequently fails. Preserve these results. S25d remains IN_PROGRESS pending the separate performance prerequisite below.

### S25u — Timeline blank-row performance prerequisite

DONE; module.timeline / workbook client. The now-executing current candidate exposes blank-row paint latency under the frozen populated fixture and 25-session traffic. Inspect stage observations and fix demonstrated unnecessary work in the owning create/render path. Preserve the real commit, exact visible-paint boundary, 100 measured samples, one warm-up, traffic, fixture and 150-ms p95 requirement. Do not tune thresholds or reclassify the failed run. Exit: owning regression tests plus the original measurement row and all-four measurement closure pass with complete cleanup. Timeline guide consulted; S25d pauses until this prerequisite passes.

S25u diagnostic step: collect a bounded Chrome CPU profile through the existing Make-owned blank-row measurement route. This temporary test instrumentation is diagnostic only and will be removed before acceptance. No threshold, sampling or traffic policy changes. The slowest samples show accepted-action-to-request and client processing costs; do not infer a cryptographic bottleneck from total latency alone.

S25u diagnostic FAIL `20261006T012612Z-p29027` (12/14, 18/18 cleanup complete): instrumented blank-row p95 196.6 ms. CPU attachment in its canonical Playwright report shows `visitNode → captureSnapshot` consumes 1,996.59 ms self time, plus 246.22 ms child traversal, during 100 measured operations. This is Playwright diagnostic DOM traversal, not application cryptography. Remove the temporary profiler; the shared Playwright configuration now omits diagnostic DOM snapshots/screenshots only for the existing exclusive measurement profile, retaining failure action/transport traces and all ordinary functional/visual diagnostics. Harness owner text states this observer-isolation requirement. Preserve all workload, sample, traffic, paint and threshold identities; renew full four-row qualification. No application UI behavior is changed.

S25u observer-only candidate FAIL `20261006T013159Z-p80049` (18/20; cleanup inspection pending): blank-row p95 222.1 ms remains above 150 ms. Snapshot overhead exists but its removal does not resolve the threshold; revert that configuration and owner-text experiment, preserving the original measurement diagnostics. Earlier typecheck `20261006T013118Z-p74447`, harness `-p74504`, configured Markdown `-p74496` passed; they do not qualify performance.

S25u structural correction: current source-mapped CPU profile attributes substantial self time to semantic cell marking and layout reads, and shows every rerender removes/re-adds unchanged semantic classes and rewrites ARIA values. Grid Adapter now updates only changed owned classes/attributes, preserving vendor/range classes, precedence, markers, focus, authorization and cell registration. No domain, storage, route or migration changes; rollback is local to this DOM projection. Benefit: any existing/future grid surface avoids unnecessary style invalidation through the sole semantic presentation owner. Risk if unresolved is repeated work across all mounted cells during real editing. Core 03 REQ-03-001..003/286 and Core 04 AC-043 govern behavior; `package.grid_adapter` owns verification. Its guide, source guide and current React 19.2.5/Playwright 1.63.0 inputs were reviewed on dirty `main` at the original baseline. The UI/UX refactor skill's narrow offline React query returned three results; ADAPT R012's DOM-write concern through existing ownership, REJECT an unrelated Activity remount change, and defer dependency memoization without evidence. New regression asserts unchanged semantic projection causes no DOM attribute mutations while existing transitions still update access and conflict state. Acceptance remains pending owner checks and the original four timing predicates. No generated roots or goldens are edited.

S25u semantic-projection candidate FAIL `20261006T014214Z-p47394` (18/20): typing/selection/focus remain passing, blank-row p95 177.7 ms still exceeds 150 ms. The new no-redundant-attribute-write regression passes (`20261006T014144Z-p45761`, 2/2), as do typecheck (`-p45870`) and import boundaries (`-p45892`). Retain the demonstrated projection simplification but do not claim it fixes the entire performance gap. Source mapping additionally identifies `WorkbookSurfaceLayout` as the largest application self-time frame (717.55 ms), with a synchronous `visualViewport.width` read during every rerender. Replace repeated reads with a layout-owned width subscription preserving the same raw viewport/fallback metric and resize semantics. Validate unchanged rerenders, resize/clamp/ARIA, then renew measured evidence. No mutation dispatch or authorization ordering change is justified by this evidence.

S25u viewport fixture initial FAIL `20261006T014945Z-p4916`: the new isolated surface fixture omitted its required query-browsing provider; add the real provider. Newly routed existing zoom coverage also retained an obsolete 900-pixel expectation at 200% root zoom; the effective 450-CSS-pixel block size matches current geometry and design §12.5's zoom-accounting rule. Correct the fixture without changing runtime zoom calculations. Generation PASS `20261006T014923Z-p1635`; typecheck `20261006T014945Z-p5063`, drift `-p4839` and shapes `-p4861` pass. These results do not close performance acceptance.

S25u viewport candidate FAIL `20261006T015129Z-p17448`: blank-row p95 182.3 ms remains above 150 ms; do not close the prerequisite. The width subscription's corrected owner regression PASS `20261006T015109Z-p16488` (2/2). The previous grid candidate's 27 cleanup steps are complete. A further source-confirmed unnecessary yield exists before dispatch: the Timeline owner wraps even local captured drafts and verified cached rows in an async function. Permit its private observation port to return an immediate value; await only real authoritative reads. Preserve all epoch, FIFO, current-version, lifetime and immutable-request checks. This is a performance hypothesis requiring owner lifecycle regression and real measurement, not a new admission shortcut.

S25u immediate-observation candidate focused PASS `20261006T015925Z-p74725` (14/14; 18 cleanup steps complete): blank-row p95 119.9 ms, p50 76.7 ms, with the unchanged 150-ms limit, 100 samples, one warm-up and 25-session workload. Workbook queue/autosave/retained-draft checks PASS `20261006T015825Z-p71395` (8/8), Timeline checks `-p71396` (4/4), collaboration `-p71399` (2/2), and typecheck `20261006T015826Z-p71550` (2/2). The earlier observer-only and viewport failed candidates each have all 27 cleanup steps complete. Full four-predicate measurement closure and responsive/accessibility checks remain required before S25u completion.

S25u DONE: full measurement PASS `20261006T020322Z-p15416` (20/20; all 27 cleanup steps complete). p95: blank-row 150 ms, typing 32.6 ms, selection 35.9 ms, focus 31.6 ms. Blank-row meets its unchanged bound exactly; its narrow margin remains a final-candidate risk, not a reason to relax acceptance. Responsive frame PASS `20261006T020901Z-p63961` (11/11); real grid accessibility PASS `20261006T020902Z-p64233` (11/11); both have complete cleanup. The structural changes remove redundant semantic DOM writes, subscribe to viewport changes rather than reading layout on each render, and avoid a microtask yield for already available owner observations. Real authoritative reads remain asynchronous and retain cancellation, version, epoch and queue checks. No payload, token, storage, authorization or compatibility changes. UI/UX review: A001–A004, A008–A009 and applicable A011–A013/A019 obligations are supported by the owner mappings, focused lifetime/queue tests, measured scenarios and responsive/accessibility evidence; theme, density, unrelated creation/editor workflows and visual redesign are unchanged and not claimed as newly qualified. Temporary profiling/configuration experiments are removed. Markdown PASS `20261006T020904Z-p64554`; configured coverage excludes this handoff, which was directly reviewed. Diff check passes; historical completed tail is byte-identical to HEAD. Next: close S25d using its complete trust/lifecycle and four-predicate evidence.

S25d DONE: performance fixtures now carry the actual isolated object-store CA through their prepared-fixture adapter and environment projection. Four real clone lifecycle tests pass and each retires twice (`20261006T011656Z-p50315`, 3/3; 6 cleanup steps complete). The repaired four Timeline measurements pass on fresh fixtures (`20261006T020322Z-p15416`, 20/20; 27 cleanup steps complete), preserving verified TLS, exact ownership and all original timing identities. S25t closes repeated retirement and S25u closes observed application latency. Existing failure artifacts remain immutable; the separate manual cleanup receipt does not change their failed outcomes. No product compatibility effect; no trust bypass or development-container change. Next: S25a, then S25b–S25c and S25e–S25s in the recorded sequence.

S25a DONE: sorted `observability_policy.required_targets` in its authored owner, retaining identical membership including credential-capacity assessment. Normal generation PASS `20261006T021102Z-p45124`; duration coverage `20261006T021120Z-p48551` (3/3), command-surface owner `-p48465` (2/2), harness contract `-p48763` (2/2), drift `-p48347` (4/4), shape `-p48391` (3/3), generated policy `-p48381` (3/3), configured Markdown `20261006T021141Z-p58262` and diff checks PASS. Cleanup complete. No product, compatibility or release-membership change. This closes the original unsorted-roster failure, not the entire release. Next: S25b; Auth, Network Flow and collaboration guides consulted (initial `module.network_flow` guide lookup was rejected before execution; correct owner is `module.networkflow`).

### S25v — Qualified-transport static-analysis prerequisite

DONE; platform.cryptography / platform.enterpriseauth. S25b's vet now passes, exposing three static-analysis findings in `20261006T021435Z-p14396`: deprecated CertPool.Subjects in the root-admission test and two redundant `any` declarations. Replace the deprecated representation assertion with exact expected-pool equality and simplify the inferred declarations; preserve all trust/algorithm behavior. Exit: narrow transport/policy owner tests and complete lint-go pass. S25b pauses. Cryptography and enterprise-auth guides must be consulted before edits.

### S25w — gRPC security patch prerequisite

DONE; platform.telemetry / dependency inputs. `go-vulncheck` FAIL `20261006T021423Z-p98013` (3/4) reports reachable GO-2026-6443 in gRPC 1.83.1 through the isolated telemetry collector; upstream fixed version is 1.83.2. Verify primary advisory/release details, update through the maintained Make dependency path, then validate telemetry TLS, dependencies/toolchain and vulnerability loading. Retain the pinned Go cryptographic module. Do not suppress the finding. The module-only x/crypto finding also requires disposition from the scanner's reachability evidence. No deployment is authorized by dependency qualification.

### S25x — Key-ring result-publication prerequisite

DONE; harness evidence. The new key-ring row's process and canonical summary pass in `20261006T021423Z-p97337` (1/1; 3 cleanup steps complete), but the public target returns artifact_error and retained row output is absent. Preserve this contradictory incomplete run and do not count it as acceptance. Diagnose publication/finalization and repair the responsible boundary; exact row evidence, successful public target and cleanup must agree. Do not weaken artifact or retained-secret checks.

### S25y — Collaboration presence readiness prerequisite

DONE; collaboration browser support. Browser FAIL `20261006T021423Z-p97381` (9/11) times out awaiting editing presence before the current-token assertion; the remote-cell action is still pending when the waiter expires. Diagnose the actual interaction/authoritative state boundary, preserving observed cross-client presence, live-row refresh and conflict resolution. The complete original row must pass, including cft4, without longer timeouts or weakened assertions.

S25b partial evidence: actual isolated trusted socket PASS `20261006T021423Z-p97341` (1/1); generate `20261006T021406Z-p93572`, drift `20261006T021423Z-p97257` (4/4), shapes `-p97272` (3/3), generated policy `-p97295` (3/3) pass. Two explicit owner rows now execute the existing socket and key-ring regressions rather than merely compiling them. Original row identities are preserved. Follow S25v → S25w → S25x → S25y before closing S25b and proceeding to S25c. Full finalization/release must be renewed after these prerequisites.

S25v DONE: the root-filter regression compares complete certificate-pool identity with an independently constructed expected pool, retaining explicit-binding rejection. OIDC/SAML inferred interface variables remove redundant type declarations without changing verification. Cryptography/Auth guides were consulted; `platform.enterpriseauth` is a source path, not a verification owner (that guide lookup rejected before execution). TLS owner PASS `20261006T021740Z-p71173` (1/1), enterprise policy `-p71174` (1/1), both with 3 completed cleanup steps. Complete lint-go PASS: formatting `-p71768`, vet `20261006T021746Z-p78630`, staticcheck `20261006T021748Z-p79261`. Configured Markdown `20261006T021740Z-p71387` and diff checks pass. No behavior/compatibility change; no suppressions. Original failed staticcheck remains preserved. Next S25w; telemetry guide already consulted.

S25w dependency preparation: upstream [GO-2026-6443](https://pkg.go.dev/vuln/GO-2026-6443) and [gRPC 1.83.2](https://github.com/grpc/grpc-go/releases/tag/v1.83.2) confirm the HTTP/2 missing-authority rejection patch. The scanner reaches transport through a test collector; this is not a claim that the application exposes the advisory's xDS configuration. Authored go.mod now selects 1.83.2. Make bootstrap PASS `20261006T021940Z-p80391` acquired its module metadata but omitted the archive checksum; normal telemetry `20261006T022001Z-p84224`, vulnerability `-p84602` and server build `-p84721` consequently fail missing-go.sum admission. Preserve these preparation failures. Use the existing Make server build with explicit `GOFLAGS=-mod=mod` dependency maintenance to let Go record verified checksums; do not hand-edit go.sum. Toolchain drift `-p84164` passes. [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) concerns unmaintained x/crypto OpenPGP, which has no package/symbol reachability in the scan; retaining other x/crypto consumers does not justify using OpenPGP. Renew the unmodified scan after the patch.

S25w DONE: gRPC 1.83.2 selected in go.mod; Go-owned maintenance build PASS `20261006T022058Z-p2047` (4/4) recorded archive checksum without manual go.sum edits. Normal server build PASS `20261006T022144Z-p17354` (4/4), verified TLS telemetry `-p16935` (1/1), unmodified vulnerability scan `-p17249` (4/4), drift `-p16866` (4/4); cleanup complete. The scanner has no reachable vulnerability after the patch; the OpenPGP module-only advisory remains an explicit unused-package disposition, not a suppression. Toolchain/module policy and stored-format identity are unchanged. No migration impact; final package/binary hashes must be regenerated and qualified. Earlier missing-checksum failures remain recorded. Configured Markdown and diff checks pass. Next S25x, then S25y, then resume S25b.

S25x diagnosis: the retained-file guard explicitly forbids basenames matching `key[-_.]?ring`. The newly authored, unqualified row `module.networkflow.unit.current_key_ring_lifecycle` unintentionally used that reserved spelling in its result filename. Rename only this new row to `module.networkflow.unit.current_cursor_digest_lifecycle`; preserve its tests, ownership, original failed evidence and every pre-existing row identity. No guard or runtime change is needed. Network Flow and command-surface guides were consulted; generate from the authored catalog, execute the exact row and inspect complete publication/cleanup.

S25x DONE: reserved-name collision removed solely from the new catalog row. Generation PASS `20261006T022301Z-p39682`; exact row PASS `20261006T022324Z-p43004` (1/1), with matching published row, passing retained-secret scan and all 3 cleanup steps complete. Drift `-p42903` (4/4), shapes `-p42925` (3/3), generated policy `-p42932` (3/3), configured Markdown `-p43300` and diff checks pass. Prior incomplete publication remains rejected evidence. No changes to runtime, secret retention policy or existing validation identities. Next S25y; collaboration guide consulted before implementation.

S25y readiness candidate FAIL `20261006T022456Z-p57687` (9/11; all 14 cleanup steps complete): the helper selected only ARIA grid while this grouped scenario correctly renders a treegrid. Correct the helper to accept the two existing grid roles within the exact surface shell. Typecheck PASS `20261006T022458Z-p57994` (2/2). The preceding trace shows visible columns reset between pre-query scrolling and click; wait for both grids' existing aria-busy=false state, scroll the remote target last, then await click/focus and the bounded socket observation together. Preserve all timeouts, predicates and real interaction. No product change or role weakening.

S25y DONE: the owning browser helper waits for the exact surface grid/treegrid's current read to finish, aligns the remote target last, and awaits the real click/focus together with the bounded socket observation. This removes the pre-query virtual-column race and unhandled waiter rejection without changing product timing, presence semantics, assertions or timeouts. Complete original multi-client lifecycle PASS `20261006T022635Z-p5504` (11/11; all 14 cleanup steps complete), including live rows, presence, stale requery, reset and cft4 conflict resolution. Helper regression `20261006T022638Z-p6011` (2/2), typecheck `20261006T022637Z-p5810` (2/2), configured Markdown `20261006T022529Z-p93582` and diff checks pass. No product or compatibility change. Historical tracker tail is unchanged. Next: resume and close S25b against its repaired compilation, routing, security and browser evidence.

S25b DONE: socket helper now uses an isolated qualified TLS identity and an explicitly trusted client; the dotted-key cursor fixture constructs its own key-ring value without copying a mutex; the actual browser flow requires cft4. Focused Auth PASS `20261006T021423Z-p97341` and current cursor/digest lifecycle PASS `20261006T022324Z-p43004`, both 1/1 with complete cleanup; new owner rows retain this execution coverage. S25v supplies complete lint-go, S25w supplies normal build and vulnerability qualification, S25x supplies complete result publication, and S25y supplies the complete original collaboration browser pass. Generation/drift/shapes/policy evidence and original failures remain recorded above. No compatibility fallback, changed public token behavior or test suppression. Next S25c; telemetry guide consulted and overload classification remains to be projected.

S25c DONE: OTEL-REQ-142 and its typed registry map `authentication_capacity_exhausted` uniquely to `capacity_exhausted`, distinguishing service admission from authentication failure and telemetry queue overflow. Logs reuse the same closed runtime class list as signal redaction; focused tests preserve the bounded public code/class without account details. Public status, empty details and Retry-After behavior are unchanged. Generation PASS `20261006T022845Z-p51014`; owner log/privacy checks `20261006T022906Z-p54512` (1/1), conformance `20261006T022905Z-p54202` (6/6), drift `20261006T022906Z-p54382` (4/4), shapes `-p54428` (3/3), generated policy `-p54440` (3/3), configured Markdown `-p54863` and diff checks pass. Cleanup complete; Markdown coverage includes the amended owner specification but excludes this handoff, which was directly reviewed. Telemetry consumers may now see the new bounded class; no persisted format migration. Next S25e; Saved Views guide consulted.

S25e DONE: selection helper waits for the requested stable saved-view ID, sheet-reference kind and surface identity before returning; action-menu opening waits for actual visibility. A controlled delayed-selection regression proves the helper cannot report completion while the old selection is still published. No application behavior, permission rule or timeout changed. Support PASS `20261006T023102Z-p72154` (2/2), complete original create/update/select/default/system-view lifecycle `-p72165` (11/11), typecheck `-p72309` (2/2), configured Markdown `-p72372` and diff checks pass. Browser cleanup 14/14 and support cleanup complete. This preserves disabled system-view actions rather than weakening their assertions. Next S25f; Entities and Parties guides consulted. Source inspection corrects the earlier prerequisite label: active-key-claims fixtures belong to Parties, not Indicators.

S25f DONE: Entities and Parties historical migration fixtures create inactive attribution identities using only schema-era fields, with deliberately unusable credential text. They no longer invoke current Auth seeding or read future envelope columns. Current-schema runtime/concurrency fixtures retain ordinary application composition. No applied migration, production credential path or compatibility reader changed. Original Entities upgrade/preflight/mention rows PASS `20261006T023317Z-p22520` (3/3); Parties claims/mutation-hash/concurrency row PASS `-p22528` (3/3), with complete cleanup. Configured Markdown `-p22671` and diff checks pass. These remain historical schema regression tests, not retained-state admission evidence. Next S25g; Recovery and harness-browser guides consulted.

S25g DONE: both reset fixtures first verify their normal admitted identity, then explicitly remove it through a migration-purpose handle to prepare the missing-marker case. Reset continues to preserve both absent and present markers, migration ledgers and lineage; Recovery still has no identity write privileges. No production reset, admission or cleanup behavior changed. Recovery owner PASS `20261006T023522Z-p72689` (3/3), harness reset `-p72697` (3/3), each with all 6 cleanup steps complete; configured Markdown `20261006T023523Z-p72841` and diff checks pass. S25m retains the separate pending routing cleanup for the new reset test. Next S25h; Notes verification belongs to module.workbook, whose guide is consulted (an initial module.notes lookup rejected before execution).

S25h DONE: the response-loss browser scenario now waits for a real successful commit receipt and completed response abort before detaching the inspector. Disabled authoring fields no longer stand in for dispatch completion. The original exact-request replay, single Note/link, cross-sheet retention and read-only accepted-refresh recovery assertions remain unchanged. Complete browser PASS `20261006T023706Z-p22314` (11/11; 14 cleanup steps complete); Note atomic recovery owner `20261006T023722Z-p41926` (2/2), typecheck `20261006T023708Z-p22620` (2/2), configured Markdown `20261006T023714Z-p29786` and diff checks pass. Test-only correction, no compatibility impact. Next S25i; web.design guide consulted.

S25i DONE: the keyboard scenario holds the second failed admission behind its existing fixture barrier, verifies the replay-disabled state and one admitted request, releases the failure, and waits for the same focused control to become available before the next activation. It then retains the original observation/cancellation recovery, viewport, zoom, text-spacing and naming assertions. Complete accessibility PASS `20261006T023903Z-p70143` (11/11; 14 cleanup steps complete); Reference Pack panel owner `20261006T023912Z-p81550` (2/2), typecheck `-p81770` (2/2), configured Markdown `-p81864` and diff checks pass. No product, pack identity or compatibility change; no timeout relaxation. Next S25j; Network Flow guide consulted.

S25j first barrier candidate FAIL `20261006T024051Z-p17875` (9/11): removing the handler while its released request was still pending caused a double-handled route error; retirement focus also failed in that interrupted run. Keep the released pass-through handler scoped to the test page so pending requests complete normally. Typecheck PASS `20261006T024052Z-p18186`. Preserve this failure and retest the complete lifecycle before changing product focus behavior.

S25j DONE: the real refresh test holds only job-status observation while checking retained results, then releases the request and proves completion within the unchanged 15-second bound. The pass-through handler remains scoped to the test page, avoiding premature route teardown. Complete original lifecycle PASS `20261006T024415Z-p58381` (11/11; 14 cleanup steps complete), including retirement focus; owner lifecycle `20261006T024414Z-p58097` (2/2), typecheck `20261006T024052Z-p18186` (2/2), configured Markdown `20261006T024426Z-p71410` and diff checks pass. First failed run also completed cleanup. No product or compatibility change. Next S25k; consult release and Recovery guides before correcting the current contract projection.

S25k DONE: the authored SeaweedFS gate now requires Recovery envelope v3 and integrity manifest v4, retaining object-store manifest/summary v2 and storage-ref ownership. Its contract test asserts the complete current list. Release gate PASS `20261006T024542Z-p777` (9/9; 12 cleanup steps complete), canonical evidence `.cartulary/release-artifacts/seaweedfs/20261006T024621Z-p25265`, with all 14 real SeaweedFS 4.48 compatibility operations passing. Release contract PASS `20261006T024543Z-p919` (1/1), configured Markdown `20261006T024544Z-p1326` and diff checks pass. No runtime or stored-format change; this repairs the downstream evidence projection. Next S25l; consult harness-browser and object-store guides.

S25l routing candidate generation FAIL `20261006T024729Z-p34413`: the proposed real roundtrip row duplicated an already-owned reset selector. Preserve the original admission row and remove only that duplicate from the new row; origin-default and real roundtrip/namespace-confinement tests receive distinct routing. No product execution or resource mutation occurred in this generator failure.

S25l DONE: default browser origins consistently use HTTPS for localhost and 127.0.0.1 across the existing port range. The origin regression rejects any plaintext default, wildcard or unspecified address. New distinct owner rows execute origin admission and real roundtrip/namespace confinement; the original reset/admission selector remains intact. Generation PASS `20261006T024804Z-p37499`; fixture checks `20261006T024821Z-p40926` (5/5), CORS positive/negative contract `-p40930` (1/1), actual SeaweedFS compatibility `20261006T024838Z-p57390` (3/3), all cleanup complete. Drift `20261006T024821Z-p40807` (4/4), shapes `-p40844` (3/3), generated policy `-p40857` (3/3), configured Markdown `-p41286` and diff checks pass. No production compatibility extension; secure fixtures follow HTTPS clients. Next S25m; Recovery and Extensions guides precede routing repair.

S25m DONE: the original restore-rejection row again contains exactly its three contracted tests; the reset-identity regression has a separate Recovery integration row. Authored catalog generation PASS `20261006T024944Z-p96761`; both Recovery rows PASS `20261006T025002Z-p451` (3/3), Extensions exact boundary-routing contract `-p436` (3/3), all cleanup complete. Drift `-p300` (4/4), shapes `-p326` (3/3), generated policy `-p352` (3/3), configured Markdown `20261006T025003Z-p773` and diff checks pass. No runtime/compatibility change and no original validation identity reassigned. Next S25n; consult Entities guide and maintained UI-review evidence workflow.

S25n first readiness candidate FAIL `20261006T025159Z-p55561` (9/11): waiting for aria-busy=false did not prevent the missing grid mention. Recorded DOM still shows the default column order. Preserve all visual assertions and diagnose whether configuration failed or was later reset before adopting a fix. Mention model PASS `20261006T025158Z-p54709` (2/2), typecheck `20261006T025149Z-p49700` (2/2), configured Markdown `20261006T025216Z-p74474`. No golden or production change.

S25n diagnostic rerun FAIL `20261006T025432Z-p2336` (9/11): the new immediate header assertion passes, proving column configuration succeeds; the original later chip check still fails. Narrow the observation to column headers and scroll offsets across inspector actions before choosing a repair. Temporary bounded diagnostics carry no credentials or domain content and will be removed.

S25n bounded diagnosis `20261006T025621Z-p41419` FAIL (9/11) shows column configuration is retained: the inspector action scrolls the selected Synopsis cell to offsets 1449/1922, and pre-capture normalization is undone by focus retention. The earlier inference of a default-layout reset was incorrect. Remove the ineffective initial-read helper changes and all temporary diagnostics. Open the final mention detail and clear focus before normalizing the capture viewport, preserving the intended existing blurred capture state and all painted-control assertions. No product change or golden update.

S25n DONE: final mention detail is opened and focus cleared before positioning the grid for its existing blurred screenshot. This removes focus-driven viewport restoration; column order and production behavior were already correct. All temporary diagnostics and ineffective helper changes are removed. Complete original visual row PASS `20261006T025857Z-p85420` (11/11), with existing goldens and painted unresolved/auto/dismissed controls unchanged; 14 cleanup steps complete in this and all three diagnostic failures. Mention model PASS `20261006T025158Z-p54709` (2/2), renewed typecheck `20261006T025858Z-p85773` (2/2), configured Markdown `20261006T025923Z-p17772` and diff checks pass. Historical tracker tail remains unchanged. No compatibility effect. Next S25o; consult Saved Views guide and import exact canonical visual differences using the UI-review artifact workflow.

S25o artifact review: exact canonical capture `visual.capture.4f5e321cc67de64bdcfa` from failed full run was imported through artifact-mode UI review `20261006T030018Z-p27678`. Digest/length-verified expected, actual and diff images were inspected at their original size. The difference is Recovery (1) versus (0), shifting adjacent top-bar text; grid/query layout is unchanged. No DOM channel accompanies this imported capture. Source review shows fixture Mark Reviewed reconciliation was still pending when query edits began. Wait for its explicit completed result before configuring the saved-view fixture. Preserve all goldens, capture identities and thresholds. Review terminal receipt confirms closed/cleanup complete; temporary request removed and private artifact links expired.

S25o DONE: the fixture observes the real Timeline review completion through Recovery before changing its query, then closes Recovery and verifies zero outstanding attention. The complete saved-view visual row PASS `20261006T030308Z-p34186` (11/11; 14 cleanup steps complete), including all previously differing desktop, narrow, compact, zoom and text-spacing captures against unchanged goldens. Timeline recovery owner PASS `20261006T030446Z-p75851` (2/2), typecheck `20261006T030308Z-p34605` (2/2), configured Markdown `-p34654` and diff checks pass. Artifact review and its clean terminal receipt are recorded above; no product, renderer, threshold or compatibility change. Next S25p; Workbook guide precedes observation lifecycle repair.

S25p DONE: the browser replay loop waits for one newly completed records/history refresh before admitting its next transition. The original failure was a safe preflight rejection of the stale child captured during the preceding refresh. Complete source-capture/resolve/dismiss/restore replay lifecycle PASS `20261006T030534Z-p82154` (11/11; 14 cleanup steps complete), retaining original bytes/transaction identity, single history effects and backend snapshots. Owner/reconciliation PASS `20261006T030541Z-p89639` (3/3), typecheck `-p89835` (2/2), configured Markdown `-p89917` and diff checks pass. No product, expiry, authorization, timing or compatibility change. Next S25q; Workbook guides precede initial candidate readiness correction.

S25q DONE: the keyboard-removal fixture waits for its known 100-candidate first page before reading and selecting the first three stable IDs. It no longer silently selects an empty list during discovery. Complete original five-scenario Party/candidate browser row PASS `20261006T030702Z-p25155` (11/11; 14 cleanup steps complete), preserving six-item continuation, off-page selection, removal identity, scroll and keyboard focus. Candidate presentation owner PASS `20261006T030707Z-p29957` (2/2), typecheck `20261006T030708Z-p30182` (2/2), configured Markdown `-p30284` and diff checks pass. No production or compatibility change. Next S25r; Workbook guide precedes stale-row readiness correction.

S25r DONE: the stale-client fixture waits for its two committed rows before collecting their rendered order and checking exact identities. The mutation substrate being mounted no longer substitutes for query completion. Both complete original public-mutation/remount-safe recovery scenarios PASS `20261006T030838Z-p67249` (11/11; 14 cleanup steps complete), retaining atomic paste, stale versions, exact replay and conflict resolution. Batch retention/refresh owner PASS `20261006T030858Z-p95945` (3/3), typecheck `20261006T030856Z-p91123` (2/2), configured Markdown `-p91222` and diff checks pass. No product or compatibility change. Next S25s, the remaining known prerequisite before renewed finalization and full release acceptance; consult Recovery and harness-browser guides.

S25s first focused attempt FAIL `20261006T031638Z-p22841` (9/11; all 14 harness cleanup steps complete): the helper expected inherited suite PostgreSQL administration, which browser attachment deliberately excludes. Read the existing private runtime Recovery credential file instead, validate its certificate URI, and derive explicit purpose identities through the maintained fixture adapter; no password or localhost fallback. Frontend typecheck PASS `20261006T031639Z-p23155` (2/2), Go lint PASS `20261006T031641Z-p23511`. TLS, purpose admission and cleanup changes remain under validation; no S25 acceptance follows from this attempt.

S25s DONE: the real Recovery browser source/target now use isolated verified HTTPS and exact public origins, purpose-specific PostgreSQL certificates and the existing private fixture credential channel. Obsolete password/localhost/HTTP and shell-env fallback readers are removed. Borrowed databases, object stores, backup storage and listeners close before owned databases are dropped; cleanup errors affect exit status. Browser shutdown sends EOF to the helper through its Go supervisor, waits for successful cleanup and fails on nonzero exit or forced group retirement; ready payloads containing disposable credentials are not printed in error diagnostics. Current key-ring v2, Auth master-key fixture and context-aware password hashing remain aligned with the finished owners. Complete actual backup/restore/workbook-query browser row PASS `20261006T031919Z-p90593` (11/11; all 14 cleanup steps complete), after the preserved failed attempt above. Typecheck `20261006T031639Z-p23155` and Go lint `20261006T031641Z-p23511` passed; format `20261006T031911Z-p85493`, configured Markdown `20261006T031920Z-p90919`, diff and byte-identical historical-tail checks pass. The final release will renew Go lint after the private binding/backup-close adjustment. No product compatibility or recovery semantic change. Next: S25 finalization, fast suite and one renewed complete release; every known prerequisite is DONE.

S25 renewed final candidate begins after all prerequisite checkpoints and release guide consultation. Operational handoff is in `deploy/mvp/README.md`, including fresh provisioning, purpose certificates/renewal, matching-release backups, rejected-state recovery, integration restrictions, operator-started overdue checks and exact platform/module requalification. Native Linux remains indefinitely deferred. Run `make agent-finalize` with RESULTS_DIR unset (retained-run maintenance skipped), then `make test-fast`, then one complete `make release-check`. Inspect canonical identities/readiness/cleanup with `make explain-run`. Implementation completion and WSL2 acceptance stay pending final release; formal CMVP applicability is UNESTABLISHED, specification adoption and customer deployment approval remain separate PENDING dispositions. No intermediate or failed run is final acceptance.

### S25z — Recovery browser fixture composition boundary

DONE; Recovery/browser test support and module boundaries. The renewed complete candidate `20261006T032808Z-p35487` found `module.records.architecture.boundary`: `tools/recoverybrowserrestore/main.go` imports `internal/testutil/pgtest`, which the existing production/tooling boundary correctly excludes outside declared test infrastructure. Move the service-starting browser fixture composition into shared test support and leave its tool as a narrow CLI adapter. Preserve certificate-only connections, full admission, backup/restore behavior and cleanup; do not broaden import exceptions or duplicate connection policy. Exit: original boundary row, actual Recovery browser row, applicable Go/security checks and renewed complete-candidate validation pass. The Records and browser/Recovery task guides are available. The full current run continues unchanged to discover all failures; no implementation edit occurs until its evidence is complete.

### S25aa — Timeline exact-action recovery preflight

DONE; Timeline/browser and mutation owner, after S25z. The same complete candidate's original `functional-support-default-timeline-workbook` group fails the replacement supersession case in “Timeline exact action recovery preserves committed transitions change sets and replacement links.” The UI reports preflight rejection (“No action was sent”) instead of the intended committed-response-loss state. Diagnose query/row readiness and current-epoch preflight using the exact scenario; preserve preflight safety, base versions, replacement identity, response-loss injection, byte-identical retry and single history transition. No timeout increase, assertion weakening or unconditional retry. Exit: the complete original lifecycle group and relevant owner regression checks pass, followed by renewed final-candidate release validation. No implementation change made while the current complete run is still collecting evidence.

### S25ab — XLSX partial-import fixture readiness

DONE; Imports/workbook browser, after S25aa. The renewed complete candidate's `functional-support-network-flow-claimed-import-assistant` group timed out while selecting the second unit's Target view in “Workbook Import Assistant reviews XLSX partial outcomes and navigates only to committed targets.” Diagnose mapping/selection settlement and preview readiness; preserve the real two-unit import, one committed/one failed outcome and navigation only to committed targets. Do not extend the timeout or bypass owner state. Exit: original complete Import Assistant group and relevant owner lifecycle regression checks pass, followed by renewed final-candidate release validation. Current run remains unchanged while collecting evidence.

### S25ac — Workbook continuation query lifecycle

DONE; workbook query/browser, after S25ab. The renewed full release's original `stateful-default-workbook-query-browsing` group reports four initial Timeline query reads against its existing maximum of three, and a later Identities route handler reports a disposed fetch response. Diagnose request ownership, page readiness and route teardown through the full multi-surface lifecycle. Preserve the original initial-read bound, cursor/page identity, continuation/return behavior and authorization. Do not relax counts or conceal transport errors. Exit: complete original stateful group and owning query lifecycle regressions pass, with cleanup, followed by renewed final-candidate release validation.

### S25ad — Saved-view reload render readiness

DONE; saved views/workbook stateful browser, after S25ac. The renewed complete candidate's original `stateful-default-workbook` group fails “verifies saved-view persistence, default/startup surface persistence, and query replay through /api/v1/ after reload”: expected beta identity is read as an empty visible-row list immediately after the correct saved-view URL is observed. Diagnose actual query/render completion and wait for the authoritative rendered state, retaining exact row identity, grouping, sort/filter replay and startup/default persistence. Exit: complete original stateful workbook group and relevant saved-view owner checks pass, followed by renewed final-candidate release validation; no broadened timeout or weakened identity assertion.

S25 second complete candidate FAIL `20261006T032808Z-p35487`: 1,271/1,278 units pass, seven fail, zero skips/cancellations, 2,676.991 seconds; all 724 cleanup steps completed. `make explain-run` inspected the canonical run. Source `sha256:3876aae8b7db6793b16593464ed28fb464234c963c818f831b3698a7f70ab888`, graph `sha256:08aeea0ccf7bbeda20b01afb189a5cd191d4f09e77a1cdafdb77c9a4eec833f0`, dirty baseline `df720535b6b6622cab8b3d399e868dc968e0f6c4`; toolchain `sha256:e28906e907c215e18aa7df021f87305dfae38bd6bdbda606f10dcce4969c5d99`, system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. All three package rows, SeaweedFS release gate, crypto/capacity assessments, Go lint/vulnerability gates and all Timeline/Network Flow measurements pass. Five underlying failures are S25z–S25ad; two browser aggregate summaries fail accordingly. Prior renewed finalization PASS `20261006T032043Z-p32317` (1/1; no generated changes; retained maintenance skipped because RESULTS_DIR unset), test-fast PASS `20261006T032116Z-p36688` (746/746; all 748 cleanup steps complete), and its `make explain-run` are supporting evidence, not a final release pass. No source changed during the complete release. Resolve S25z → S25aa → S25ab → S25ac → S25ad, checkpointing each, then renew finalization/fast/full release. Original run results remain unchanged.

S25z IN_PROGRESS after complete-run/cleanup inspection and Records/browser/Recovery guides. Move the browser restoration lifecycle into shared test infrastructure, leaving command argument/signal handling at the tool composition root. Preserve the existing production-no-pgtest-import rule and all runtime behavior. No exception or alternate credential policy is introduced.

S25z DONE: `internal/testutil/recoverybrowsertest` owns the disposable lifecycle and receives context/input/output; `tools/recoverybrowserrestore` is a narrow argument/signal adapter. No boundary exceptions, duplicated certificate policy or product behavior change. Original Records boundary row PASS `20261006T041431Z-p90833` (3/3; five cleanup steps), actual Recovery browser PASS `20261006T041431Z-p90840` (11/11; all 14 cleanup steps complete), targeted security PASS `20261006T041431Z-p91171` (4/4; six cleanup steps), Go lint exit 0. Format and configured Markdown `20261006T041447Z-p31564` pass, as does diff review. No failed qualification attempt in this prerequisite. Next S25aa; mark IN_PROGRESS and consult Timeline/Workbook guides before its request-settlement correction. Full candidate acceptance remains pending renewed S25 validation.

S25aa DONE: the exact-action fixture keeps the reviewed editor open until its first injected abort/malformed response has actually completed. Opening Recovery while the replacement lookup was still in preflight correctly invalidated that editor; the fixture now exercises the intended submitted-operation uncertainty. Original complete four-case lifecycle PASS `20261006T041656Z-p71757` (11/11; all 14 cleanup steps complete), preserving base versions, reason/replacement identity, exact retry, single change set and history. Capture owner PASS `-p71815` (2/2), typecheck `-p71966` (2/2), configured Markdown `-p72024`, format and diff pass. No production behavior, timeout or compatibility change; no failed focused attempt. Next S25ab IN_PROGRESS; consult Imports/Workbook guides and wait for the first unit's completed mapping/selection before requesting the next preview.

S25ab DONE: the XLSX fixture waits for “Mapping approved and selected” on its first unit before loading the second preview. The selection stage otherwise cancels that prematurely admitted read, correctly exposing preview recovery. All four rows in the complete original claimed Import Assistant group PASS `20261006T041919Z-p19918` (11/11; all 14 cleanup steps complete), retaining real partial outcomes, committed-only navigation, keyboard recovery and density checks. Import lifecycle PASS `-p19909` (2/2), typecheck `-p20079` (2/2), configured Markdown `-p20142`, format and diff pass. No product, timeout or compatibility change; no failed focused attempt. Next S25ac IN_PROGRESS; consult Workbook guide and diagnose the four identical initial query requests and route teardown before changing the observer or query owner.

S25ac diagnostic step: retained trace confirms four identical initial `{limit:100}` Timeline requests. Add temporary bounded callback-dependency/load-reason observations and run the original Timeline continuation row through Make; no query body, credentials or response content is logged. Remove diagnostics before acceptance. The original three-read startup bound and twelve-read traversal bound remain unchanged.

S25ac diagnostic runs PASS `20261006T042243Z-p62425`, `20261006T042508Z-p1979` and complete original group `20261006T043017Z-p42071` (30 browser scenarios). All isolated surfaces issue one startup read; the full failure includes a socket reconnection and four identical requests. Source review confirms committed writes can still await durable stream sequencing, so HTTP fixture completion alone does not establish quiescence. Make fixture publication observable through the real socket before a fresh navigation, retaining the original request limits and live-reconciliation scenarios. Drain only owned observer work before page disposal. Remove all temporary diagnostics; no production query or authorization behavior changes. Renew full-group and owner validation before completion.

S25ac DONE: continuation fixtures await each real durable `record_changed` publication with their existing eight-worker bound, detach setup, and then measure a fresh page. Successful HTTP fixture writes alone did not guarantee publication quiescence; production query, authorization and collaboration behavior remain unchanged. Observer handlers are removed and their pending fetch/fulfill work finishes before page disposal, without hiding transport failures. All 30 original stateful browsing scenarios PASS `20261006T043506Z-p86870` (11/11 harness units; all 14 cleanup steps complete), including unchanged startup/traversal limits, cursor bytes, cross-surface continuation and live recovery. Query owner PASS `-p86848` (3/3), typecheck `-p86992` (2/2), configured Markdown `-p87053`, format and diff pass. Diagnostics removed; no compatibility change. The diagnostic full group and focused runs above are retained separately. Next S25ad IN_PROGRESS; consult Saved Views/Workbook guides before fixing render readiness. Full release acceptance remains pending.

S25ad DONE: the saved-view fixture waits for its exact beta row after default startup and reload, retaining the complete saved identity, filter/sort/group query, home/default persistence and direct API checks. URL/shell admission no longer substitutes for completed rendering. Original complete stateful Workbook group PASS `20261006T044016Z-p35977` (11/11; all 14 cleanup steps complete), saved-view/startup owners PASS `-p35993` (3/3), typecheck `-p36139` (2/2), configured Markdown `-p36209`, format and diff pass. No product, timeout or compatibility change; no failed focused attempt. Direct review confirms the historical tracker tail remains byte-identical to HEAD. All known prerequisites are DONE.

S25 third complete candidate: consult release guide, run `make agent-finalize` with RESULTS_DIR unset (retained-run maintenance skipped), then `make test-fast`, then one complete `make release-check`. Preserve the candidate during execution and inspect canonical run/readiness/identity/cleanup projections with `make explain-run`. Prior failed full runs and their cleanup remain unchanged; this attempt must stand on its own. Implementation completion and WSL2 acceptance remain pending, CMVP applicability UNESTABLISHED, adoption/customer approval independently PENDING.

S25 third candidate finalization PASS `20261006T044218Z-p77898` (1/1; generated artifacts unchanged; retained-run maintenance skipped because RESULTS_DIR unset). Fast suite PASS `20261006T044257Z-p82220` (746/746; 287.656 seconds), canonical `make explain-run` inspected before full release. Start one complete `make release-check` against the unchanged candidate; S25 remains IN_PROGRESS until complete final acceptance and cleanup inspection.

### S25ae — Viewer keyboard surface-entry readiness

DONE; Incident Administration/Workbook browser prerequisite discovered by third complete release `20261006T044838Z-p57124`. Original `functional-support-default-incident-administration` fails “Verify viewer built-in selection focuses committed cells then the empty grid.” The expected Evidence cell is absent at line 1265; retained DOM still shows Timeline and an expanded incident-details trigger. Diagnose initial selector focus and startup readiness from the retained trace, preserving actual keyboard navigation, viewer restrictions, committed-cell focus and empty-grid focus. Do not replace the keyboard flow with a direct click, weaken focus assertions or increase timeouts. Exit: original complete browser group and relevant owner focus/startup tests pass, then renewed finalization/fast/full release. The current run continues without source changes to gather all failures and cleanup.

### S25af — Promoted-metadata conflict event ordering

DONE; Incident Metadata browser prerequisite after S25ae. Third complete release `20261006T044838Z-p57124` fails “metadata edits each live field sparsely clears normalizes and reviews a real version conflict” in `functional-support-default-incident-metadata-editing`: Save promoted fields remains disabled at the conflict submission (line 115). Inspect no-op normalization completion versus the following edited draft and competing write. Preserve sparse requests, unchanged-version no-op acknowledgment, retained draft, actual optimistic conflict, explicit version adoption and lifecycle checks. Exit: complete original group and relevant metadata owner regressions pass, followed by renewed full candidate validation. No source change while the complete run is active.

### S25ag — Mention-chip visual state completion

DONE; Workbook/mentions visual prerequisite after S25af. Third complete release `20261006T044838Z-p57124` reports 38,926 differing pixels (5%) for `entity-mention-chip-states.png` in the original unresolved/resolved/auto-resolved/dismissed/manual-resolution visual scenario. The visual aggregate fails accordingly. Inspect the exact retained capture through UI-review routing and compare semantic state, focus, scroll and mutation completion against the unchanged golden. Preserve the original fixture and thresholds; no blanket snapshot refresh. Exit: exact original visual row and relevant owner regressions pass, then one renewed full candidate release. The current source remains unchanged while the release completes.

### S25ah — Recovery browser helper shutdown lifecycle

DONE; Recovery/browser test infrastructure after S25ag. Third complete release `20261006T044838Z-p57124` passes the restored workbook/query assertions but the original `functional-support-default-restore` scenario fails during helper shutdown: the parent’s five-second EOF cleanup allowance expires. Inspect browser connection detachment, cancellation, reverse resource ownership and the helper’s bounded server/database retirement before changing the supervisor budget. Cleanup must remain explicit, bounded and reported; forced process retirement remains failure. Exit: actual original backup/restore/browser scenario, lifecycle/boundary checks and cleanup pass, followed by renewed complete final validation. Record any owned residue from the completed run separately without rewriting its original receipts.

S25 third complete candidate FAIL `20261006T044838Z-p57124`: 1,272/1,278 units pass, six fail, zero skips/cancellations, 2,719.433 seconds. `make explain-run` inspected source `sha256:f06d36f901c144cd579ee1a6ecb9f2faae4fcf4700a019390314e0079bff16ba`, graph `sha256:08aeea0ccf7bbeda20b01afb189a5cd191d4f09e77a1cdafdb77c9a4eec833f0`, dirty baseline `df720535b6b6622cab8b3d399e868dc968e0f6c4`, unchanged toolchain `sha256:e28906e907c215e18aa7df021f87305dfae38bd6bdbda606f10dcce4969c5d99` and system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. All 676 harness cleanup steps completed; no outstanding harness-owned cleanup action. The Recovery helper’s forced shutdown is still a failed scenario, not rewritten as successful helper cleanup. All three package rows, SeaweedFS gate, credential-capacity assessment, stateful/a11y browser stages and fixed measurements pass. Four underlying failures are S25ae–S25ah, plus functional/visual aggregates. Prior finalization `20261006T044218Z-p77898` and fast `20261006T044257Z-p82220` remain separate; fast cleanup is 559/559 complete. No product source changed during this complete run. Resolve each new prerequisite with its own checkpoint before renewing the entire final sequence. S25ae IN_PROGRESS after completed-run inspection.

S25ae DONE: the viewer fixture waits for initial grid admission; desktop keyboard entry first requires visible tabs, confirms the preceding control’s focus and verifies Tab reached a surface tab before arrows/activation. No direct-click replacement, timeout increase or product focus change. All thirteen default scenarios from the original multi-owner group pass through their four exact owner slices: Workbook `20261006T053645Z-p1743`, Design `-p1752`, Incidents `-p1751`, Saved Views `-p1762` (each 11/11; each all 14 cleanup steps complete). Selector/grid-entry regressions PASS `20261006T053646Z-p2608` (3/3), typecheck `20261006T053647Z-p2922` (2/2), configured Markdown `20261006T053722Z-p17529`, format/diff pass. No failed focused attempt. The full candidate will renew the combined group in its canonical routing. Next S25af IN_PROGRESS with Incidents/metadata guides: retained DOM proves the competing update already opened conflict review before the attempted Save. Use an explicit outbound-write barrier to establish a real in-flight optimistic conflict; keep proactive conflict behavior unchanged.

S25af DONE: the fixture waits for normalized no-op confirmation, admits the local write behind an owned transport barrier, commits the competing API edit, then releases the local request and requires HTTP 409. This fixes event ordering while preserving proactive review, retained raw input and separate explicit version adoption/submission. All five original metadata browser scenarios PASS `20261006T053952Z-p66285` (11/11; all 14 cleanup steps complete), metadata owner PASS `-p66278` (2/2), typecheck `-p66430` (2/2), configured Markdown `-p66492`, format/diff pass. No product or compatibility change and no failed focused attempt. Next S25ag IN_PROGRESS; inspect exact capture `visual.capture.bda571d19ea0163393fc` from the third full run through artifact-only UI review before changing fixture completion.

S25ag artifact review: imported exact third-run capture `visual.capture.bda571d19ea0163393fc` through `make ui-review UI_MODE=artifacts` / `ui-capture`, session `uireview-0d7f7e686cfbe597f0c4316a9e7ca5c3`, run `20261006T054159Z-p8982`. Verified returned actual/expected byte lengths and digests and inspected both original images. Mention states agree; actual Inspector heading is clipped by upward scroll while expected starts at its top. Replace the fixture’s nearest-scroll-container adjustment with a declared Inspector anchor and explicit body origin after presentation readiness, plus a fully visible heading assertion. Goldens/tolerances and production rendering remain unchanged. Exact stop returned a successful terminal receipt with cleanup completed; private images expired and caller scratch was removed. Renew the original row before completion.

S25ag first framed candidate FAIL `20261006T054509Z-p15885` (9/11; original visual comparison still differs by 14,606 pixels, 2%; heading/framing assertions pass). Owner regressions `-p15876` (3/3), typecheck `-p16030` (2/2) and configured Markdown `20261006T054510Z-p16094` pass. Retain this result and inspect its exact remaining difference before revising framing further; no threshold or golden change.

S25ag second artifact review: session `uireview-938e27242dba26dc25317edf148b5c1f`, run `20261006T054658Z-p57828`, imported the exact first-candidate capture and verified component digests. Actual/diff inspection shows the proposed body-zero reset exposed Details, an introduced fixture mistake. Restore the established Relationships item/eight-pixel inset using the existing declared-scrollport geometry helper, which also clears outer scroll offsets; verify active Relationships, full heading and exact inset. Terminal receipt is successful with cleanup complete; private artifacts expired and caller scratch removed. First candidate cleanup is 14/14 complete. No golden, tolerance or product change.

S25ag DONE: the original mention-chip visual row PASS `20261006T054841Z-p63969` (11/11; all 14 cleanup steps complete) against the unchanged golden. The fixture declares the actual Inspector body, preserves its original eight-pixel mention inset, clears outer scroll offsets through the existing geometry helper and verifies active Relationships/full heading/exact inset before and after capture. The old heuristic scrolling helper is removed. Owner regression PASS `20261006T054509Z-p15876` (3/3); final typecheck `20261006T054841Z-p64070`, configured Markdown `-p64117`, format/diff pass. The first candidate failure and both successful artifact-review cleanup receipts remain recorded above. No production, tolerance or compatibility change. Next S25ah IN_PROGRESS; consult Recovery/browser owner guides and resolve helper cancellation/connection shutdown before final renewal.

S25ah DONE: Recovery fixture servers now cancel request contexts before draining and await active HTTP/hijacked WebSocket handlers before releasing borrowed databases; close is idempotent. The browser detaches before EOF, and its bounded parent allowance covers the child’s two five-second drains and two ten-second database retirements with closure margin. Forced retirement still fails. A verified-TLS regression leaves both HTTPS and WebSocket requests active and requires cancellation/completion; owner routing was authored and generated through Make. No production format or compatibility change. `make generate` PASS `20261006T055819Z-p13140`; lifecycle row PASS `20261006T055838Z-p16274` (1/1); original Recovery browser PASS `20261006T055839Z-p16552` (11/11, all 14 cleanup steps complete); typecheck `20261006T055840Z-p16880` (2/2), generation drift `20261006T055842Z-p17082` (4/4), correct Records boundary `20261006T055853Z-p32118` (3/3), targeted security `20261006T055854Z-p32927` (4/4), Go lint, catalog check, format and diff pass. Initial boundary invocation used nonexistent row `module.records.boundary.module_boundaries` and failed graph admission before execution; it is a command-selection error, corrected to `module.records.architecture.boundary`, not a product failure. Configured Markdown PASS `20261006T055854Z-p32674`; coverage still excludes this handoff, which was directly reviewed. No failed implementation test. Next resume S25 with renewed finalization, fast checks and a complete release on one frozen candidate. The third full run and all earlier failures retain their original identities.

S25 fourth final-candidate sequence: `make agent-finalize` PASS `20261006T060018Z-p97082` (1/1). `RESULTS_DIR` was unset, so retained-run maintenance was skipped. All S25 prerequisites through S25ah are DONE. Operating guidance was directly reviewed for provisioning, certificate/key renewal, historical backup/release pairing, rejected initialization, verified integrations, operator-started restart/overdue checks and exact-module/platform requalification. `make test-fast` started; the complete release follows only after it passes. Product source will remain frozen during final qualification.

S25 fourth fast candidate PASS `20261006T060049Z-p1748`: 746/746, zero failures/skips/cancellations, 371.911 seconds, all 748 cleanup steps complete. Source `sha256:5cd3b9762774b2d3317e34539ea6ef37c0e960e0777c5ac8e22dd71054dda612`, graph `sha256:98b7df56111459087a9350e29f072f36140bcf8b71f155d6200f20ee433834e2`, dirty baseline `df720535b6b6622cab8b3d399e868dc968e0f6c4`, toolchain `sha256:e28906e907c215e18aa7df021f87305dfae38bd6bdbda606f10dcce4969c5d99`, system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. `make release-check` started for the same frozen candidate; no partial result establishes final acceptance.

### S25ai — Timeline Create fixture initial-query readiness

DONE; Timeline browser prerequisite discovered by fourth full release `20261006T060721Z-p93508`. The original `functional-support-default-timeline-workbook` group fails “Timeline Add row and Evidence stay independent of Create and closed or viewer admission” at its first draft-focus assertion. Its retained trace shows `aria-busy=true` and no draft data row at Add-row pointer dispatch; the fixture currently waits only for mutation-substrate readiness and Saved. Inspect owner readiness and focus lifecycle, retain native activation/no-create/Evidence/viewer/closed assertions and add an explicit completed initial-query prerequisite where appropriate. Do not weaken focus expectations or extend their timeout. Source remains frozen while the complete run finishes; no correction has been applied yet. Exit: actual original group, relevant focus/draft owner regressions, typechecking and cleanup pass, followed by a new complete final candidate. Record all further failures from this fourth run separately.

### S25aj — Narrow-layout recovery completion reachability

DONE; workbook accessibility support after S25ai. Fourth full release `20261006T060721Z-p93508` fails the ordinary-reference/recovery keyboard scenario after accepted recovery, when `expectDecisionControlReachable` checks account navigation at width 390: right edge 392.578125 exceeds the unchanged viewport-plus-one bound of 391. Both recovery-control viewport captures and the retained recovery assertions passed. Inspect the asynchronous completed-state layout and the helper’s immediate geometry sample; determine transient settling versus persistent product overflow before repair. Keep bounds, focus, keyboard and document-overflow assertions intact. Source remains frozen until the complete run finishes. Exit: original accessibility row/group, relevant owner regression/type checks and cleanup, then renewed full-candidate validation; no new tolerance or golden.

S25 fourth complete candidate FAIL `20261006T060721Z-p93508`: 1,275/1,279 pass, four fail, zero skips/cancellations, 2,696.953 seconds; all 723 cleanup steps complete. `make explain-run` inspected source `sha256:5cd3b9762774b2d3317e34539ea6ef37c0e960e0777c5ac8e22dd71054dda612`, graph `sha256:7429555b27f6c38420953a3e970e577cbee7be97f66e6a2a37b5116de1ba3d0a`, unchanged toolchain `sha256:e28906e907c215e18aa7df021f87305dfae38bd6bdbda606f10dcce4969c5d99` and system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. All three package smokes, visual suite, credential capacity, vulnerability checks, earlier metadata/viewer/Recovery repairs and full stateful browser stage pass. Two underlying failures are S25ai/S25aj, plus their functional/accessibility aggregates. No source changed while this run executed. S25ai remains IN_PROGRESS, its Timeline guide is consulted, and its completion checkpoint must precede S25aj. No outstanding cleanup action. Prior failed runs retain their original identities.

S25ai DONE: `openTimelineCreateFixture` now waits for completed initial-query state (`aria-busy=false`) and the single draft row after mutation readiness/Saved. The original native Add-row focus assertion, no-create behavior, Evidence selection and viewer/closed admission remain unchanged. All 15 original Timeline browser scenarios PASS `20261006T065403Z-p36922` (11/11; all 14 cleanup steps complete). Focus/row-model owners PASS `-p36929` (3/3); frontend typecheck `20261006T065404Z-p37090` (2/2); configured Markdown `-p37152`, format and diff pass. No product or compatibility change and no failed focused attempt. Next S25aj IN_PROGRESS; consult workbook owner guidance and test the existing narrow-layout bounds through completed recovery before final renewal.

S25aj DONE: the unchanged isolated scenario PASS `20261006T065642Z-p79389` (11/11) confirmed the full-run failure was not consistently reproducible. The shared reachability helper now observes focused element geometry and document overflow in one browser evaluation and retries the existing bounds for the normal five-second assertion budget after native focus/scroll. It performs no corrective focus, scroll, scale or layout action while retrying; persistent overflow/focus loss still fails. Original one-pixel containment tolerance and keyboard assertions remain intact. The trace supports a transition-order sensitivity, not a proven persistent product-layout defect; no production CSS changed. Full `make browser-e2e-a11y` PASS `20261006T065923Z-p24850` (20/20, all 24 cleanup steps complete), including all 49 main workbook scenarios and the other stage groups. Viewport owner PASS `-p24311` (2/2), account-menu owner `-p24310` (2/2), frontend typecheck `-p24513` (2/2), configured Markdown `-p24603`, format/diff pass. No failed corrected run. No compatibility change. All S25 prerequisites through S25aj are DONE; renew finalization → fast → one complete release. Earlier full failures remain failed evidence.

S25 fifth candidate: `make agent-finalize` PASS `20261006T070455Z-p73448` (1/1); `RESULTS_DIR` remains unset and retained-run maintenance was skipped. S25ai/S25aj are complete; the tracker’s historical tail is byte-identical to the baseline and `git diff --check` passes. The unchanged S25aj reproduction also completed all 14 cleanup steps. Fresh fast validation started before the next complete release; no executable source edits are planned during those runs.

S25 fifth fast candidate PASS `20261006T070550Z-p77825`: 746/746, zero failures/skips/cancellations, 277.678 seconds, all 559 cleanup steps complete. Source `sha256:952d2bb30297cc2b2cd1c5ae02443ab57eec3d06dea350e300c93ea7fa00d4b2`, graph `sha256:98b7df56111459087a9350e29f072f36140bcf8b71f155d6200f20ee433834e2`, unchanged baseline/toolchain/system identities. The fifth complete `make release-check` started on this frozen candidate. No prior passing subset is substituted for its exit.

### S25ak — Metadata conflict authorization observation

DONE; incident metadata/browser and authorization observation after completed S25aj. Fifth full release `20261006T071042Z-p52311` fails the original metadata live-field/version-conflict scenario: local PATCH correctly returns 409, but review retains version 11/Unset and “Refreshing promoted fields”. Trace records the subsequent `/api/v1/auth/session` observation as pending (no response) through the five-second value assertion; no subsequent metadata GET occurs. The competing PATCH succeeds. Inspect the actual authorization/transport boundary and pending-request lifecycle; do not extend the value timeout or weaken the real-conflict assertion. This failure differs from S25af’s preflight ordering issue, which remains fixed. Keep source frozen until the complete run finishes. Exit: root cause resolved, owner/adversarial lifecycle tests and original browser group/cleanup pass, followed by a renewed complete final candidate. Further failures and cleanup retain this run’s identity.

S25ak trace refinement: at monotonic 6879.880 ms, the fixture removes its last request-interception pattern immediately after the 409; at 6880.373 ms, the owner’s automatic session revalidation starts and remains pending through failure. Earlier session reads completed normally. The overlap points to premature transport-fixture retirement during the conflict’s follow-up observation, rather than a proven metadata-controller bug. Retain the barrier through the complete conflict review/readiness assertion before disposing it, and validate that lifecycle without changing production authentication, retry bounds or the five-second review assertion. This is a source-based diagnosis to verify after the current complete run ends.

### S25al — Constrain narrow account-menu layout

DONE; application menu / workbook layout after S25ak. Fifth full release `20261006T071042Z-p52311` reproduces exactly the 392.578125-pixel right edge at width 390 throughout S25aj’s five-second observation. This disproves the transient-settling hypothesis: S25aj’s observation improvement passed its slice but did not resolve the full candidate’s persistent layout defect. Inspect the top-bar’s horizontal scroll extent, the account slot’s eight-rem maximum, fieldset containment and button sizing at the real browser boundary. Repair structural sizing/overflow ownership; preserve the one-pixel bound, focus and keyboard access. No timeout, pixel tolerance or golden adjustment. Exit: retained full-run layout reproduced and corrected; actual narrow browser evidence, owner tests and visual/a11y regressions pass, then renewed complete release.

### S25am — Saved-view keyboard pagination readiness

DONE; Saved Views/workbook accessibility after S25al. The same fifth full release fails the saved-view discovery keyboard matrix at its First-button focus assertion. The button is initially disabled and later enabled; inspect the admitted page transition and native Shift-Tab target before changing behavior. Keep real keyboard, scope, zoom/density/text-spacing and bounded-layout requirements. Exit: original discovery accessibility row/group plus owner focus/pagination regressions and cleanup pass, followed by final renewal. Source remains frozen until the fifth run finishes.

### S25an — Assessment late-acceptance focus observation

DONE; Assessments/workbook browser after S25am. Fifth full release `20261006T071042Z-p52311` fails “Late Assessment acceptance preserves navigation focus and retained receipt” at the Hosts-tab focus assertion after releasing the delayed response; the Hosts tab is selected but inactive. Inspect surface-entry focus completion versus the explicit focus whose retention is being tested. Preserve delayed acceptance, navigation, receipt and keyboard assertions, and distinguish owner focus behavior from fixture timing. Exit: original Assessment recovery browser group, relevant focus/owner regressions and cleanup pass before renewed final candidate. No source changes while the fifth full release remains active.

S25 fifth complete candidate FAIL `20261006T071042Z-p52311`: 1,274/1,279, five failed units, zero skips/cancellations, 2,657.139 seconds. All 675 cleanup steps completed. `make explain-run` inspected source `sha256:952d2bb30297cc2b2cd1c5ae02443ab57eec3d06dea350e300c93ea7fa00d4b2`, graph `sha256:7429555b27f6c38420953a3e970e577cbee7be97f66e6a2a37b5116de1ba3d0a`, unchanged baseline/toolchain/system identities. All three package smokes, SeaweedFS gates, pinned policy, credential capacity, measurements, stateful/visual suites and Timeline readiness repair pass. Four underlying scenarios map to S25ak–S25an, with three failed browser groups plus two aggregates. S25aj’s persistent-layout diagnosis is corrected explicitly in S25al. Source remained frozen; no outstanding cleanup. Next complete S25ak, checkpoint, then S25al → S25am → S25an before renewing final validation.

S25ak DONE: the real 409 fixture retains its released transport interceptor until automatic authorization/current-value observation completes, exact raw draft is verified and Use this version is enabled. Disposal still runs in finally, including failure. Production authorization/read deadlines and review assertions are unchanged. Original five metadata browser scenarios PASS `20261006T075610Z-p93408` (11/11; all 14 cleanup steps complete), metadata owner `-p93398` (2/2), typecheck `-p93552` (2/2), configured Markdown `20261006T075611Z-p93606`, format/diff pass. No failed corrected run or compatibility change. Next S25al IN_PROGRESS: inspect actual narrow account-menu containment, then repair and validate structural layout before S25am.

S25al seeded diagnosis: UI-review session `uireview-84a9dc031b44e4f2ea9a872ca5d99a7f`, run `20261006T075745Z-p35969`, reviewed 390×480 top-bar/account containment through the maintained harness. Exact bundles 7/13 had verified bytes/digests and their originals were inspected. Both the initial synthetic admin label and a disposable profile renamed to the browser fixture’s longer label produce a contained 128-pixel button/right edge 378.359375. This does not reproduce the original recovery transition; do not infer a fieldset sizing fix from source alone. One fill request used the wrong schema field and was rejected before mutation; corrected after schema/status inspection. Exact stop returned status ok and terminal cleanup completed; foreground exited, private images expired and caller request scratch removed. Next obtain bounded geometry diagnostics from the original failing browser path before choosing a structural correction.

S25al diagnosis continuation: the original row now explicitly selects default density and restores the prior preference in finally; this alone passes `20261006T081512Z-p23369` (11/11), so density is not yet established as the cause. The unchanged prior diagnostic row also passed `20261006T080937Z-p77805`. Ancestor geometry diagnostics preserve the original one-pixel bound and never repair focus/scroll on retry. An attempted `make browser-e2e-a11y ROWS=...` was rejected before execution because that target does not declare ROWS; corrected to the owner `test-slice`. The full accessibility target is now running to inspect sequence-dependent geometry; S25al remains IN_PROGRESS.

S25al structural decision: full accessibility diagnostic `20261006T081700Z-p62416` failed only the already tracked S25am scenario (48/49 workbook scenarios; 18/20 graph units; all 24 cleanup steps complete). Passing geometry run `20261006T082501Z-p14214` confirms a 128-pixel fieldset/button with correct containment; its top-bar scroll is 143 versus the retained failure’s 128, exactly 15 pixels of displacement. A native-scrollbar-style probe `20261006T082749Z-p58338` also passes and does not establish scrollbar coupling as the cause. All three narrow diagnostic runs have 14 completed cleanup steps. Remove temporary metrics/style probes. Keep the account actions at the visible inline edge only in the below-minimum scrollable bar, so neighboring layout/scroll movement cannot remove global account navigation; other controls retain existing scrolling. Strengthen the original recovery row to require account visibility before focus and through completion, retaining native keyboard and one-pixel bounds. No changed pixel tolerances or golden replacement.

S25al design refinement: the strengthened original row first fails `20261006T083022Z-p2789` (account intersection ratio zero before focus). Sticky actions pass the functional row but fail the existing visual row in `20261006T083154Z-p47239` (11/13 graph units); exact artifact review session `uireview-4527b47157e701e8d79bb01e24c05869`, run `20261006T083258Z-p92566`, verified and inspected the 390-pixel original and found overlap with System Views. Reject that intermediate design; stop completed with terminal cleanup complete and caller scratch removed. Final design separates the below-minimum horizontally scrolling workbook controls from a stable account region, avoiding overlay and retaining existing desktop layout through display:contents. The original stronger functional row passes `20261006T083432Z-p98369` (11/11), workbook viewport/surface owners `20261006T083455Z-p29672` (3/3), typecheck `-p29857` (2/2); earlier account menu owner `20261006T083154Z-p47236` also passes. A complete ordinary visual run now determines exact affected baselines. The earlier no-golden-change intent is superseded only for reviewed images stale relative to this validated structural correction, following the golden maintenance guide; no renderer, viewport, normalization or tolerance changes. S25al stays IN_PROGRESS through visual review, any required Make-owned maintenance and renewed verification.

S25al visual review: ordinary full visual `20261006T083606Z-p39496` fails only eight top-bar comparisons in seven scenarios; 39 workbook scenarios and the claimed Network Flow scenario pass, all functional assertions complete. Reconciliation accounts for all 255 active captures/goldens with zero missing, ambiguous or orphan mappings; its only error is the failed target attempt. All 16 cleanup steps complete. Exact canonical captures were imported into artifact session `uireview-3afb8b20a7cf567d85bb133b899e4f7a` / `20261006T084455Z-p82515`; all eight actuals and diffs were verified and inspected, with a representative expected original. Diffs are confined to the intentional visible account region; main content, focus, typography and drawer geometry remain unchanged. Session stop/terminal cleanup complete; private scratch removed. The initial added System Views observation in `20261006T084547Z-p89940` fails when focused without reveal; use the same existing native focus-and-scroll reachability helper as all controls, then require full intersection and no overlap with account. No pixel tolerance change. Golden refresh remains pending that check.

Accepted refresh candidates (no viewport, zoom, renderer, masks, normalization or scope changes; Coordination declares stable fixture `visual.fixture.contextual_coordination_creation`; the other seven images are active nonregistry captures without declared fixture IDs):

| Owner row | Capture identity | Golden filename |
| --- | --- | --- |
| `web.design.visual.membership_management_visual` | `visual.capture.00f7c068d36017c70d3c` | `membership-management-removal-narrow-linux.png` |
| `module.workbook.visual.ordinary_create_authoring_recovery` | `visual.capture.19f35bdbc6ff546cd0ea` | `ordinary-recovery-390-linux.png` |
| `module.workbook.visual.ordinary_create_authoring_recovery` | `visual.capture.1aef889a4b404d1f3aa2` | `ordinary-closed-retained-narrow-linux.png` |
| `module.workbook.visual.coordination_create_authoring_recovery` | `visual.capture.1b5164dd291daeff516f` | `coordination-recovery-narrow-linux.png` |
| `module.workbook.visual.preferences` | `visual.capture.81c8c654c29ea74cd73e` | `workbook-preferences-uncertain-narrow-linux.png` |
| `web.design.visual.membership_audit_browsing` | `visual.capture.914ce670aa5017b301a7` | `membership-audit-inspected-narrow-linux.png` |
| `web.design.visual.metadata_editing` | `visual.capture.b870f1c3884a0a5b59d1` | `metadata-review-narrow-linux.png` |
| `web.design.visual.lifecycle` | `visual.capture.e6a743acbe50272f803f` | `lifecycle-review-narrow-linux.png` |

S25al final focus/reveal regression PASS `20261006T084751Z-p34685` (11/11): account visible before focus and throughout accepted recovery, System Views fully revealed without overlap, native account menu open/escape/focus retained. Proceed with Make-owned refresh of the eight reviewed stale images; exact pre-refresh file hashes are retained in caller scratch to verify the mutation set. Refresh completion alone will not close S25al; require actual changed-image review, two ordinary visual passes and applicable projection checks.

S25al refresh PASS `20261006T084919Z-p73509` (12/12, 47 visual scenarios, all 16 cleanup steps complete). Transactional promotion changes exactly the eight listed PNGs and `tools/frontend_visual_golden_manifest.json`. Every promoted PNG is SHA-256-identical to its exact previously inspected actual, so the eight-image review covers the promoted bytes without inference from filenames. No unrelated golden changes. The before/after functional attempts `20261006T083022Z-p2789`, `20261006T083154Z-p47239`, `20261006T083432Z-p98369`, `20261006T084547Z-p89940`, `20261006T084751Z-p34685` have respectively 14/17/14/14/14 completed cleanup steps, no unresolved cleanup. Two independent ordinary visual runs against the promoted manifest plus projection checks are in progress; no advance to S25am yet.

S25al projection checks PASS: generate-drift `20261006T085747Z-p17678` (4/4), JSON shape `-p17738` (3/3), generated-artifact policy `-p17724` (3/3), frontend typecheck `-p18204` (2/2), test-catalog-check exit 0, configured Markdown `-p18314`, format/diff. Markdown coverage remains `docs/*.md` and `docs/spec/**/*.md`, excluding this controlling handoff; direct review and diff supplement it.

S25al DONE: below-minimum workbook controls own their horizontal scroller; global account navigation has a separate stable region. Supported-width layout is preserved. The original recovery row now sets/restores its density, keeps account visible before focus and through completion, reveals System Views without overlap, and preserves original keyboard/focus and one-pixel geometry assertions. Temporary diagnostics and the rejected sticky overlay are removed. Eight reviewed goldens were refreshed through Make with unchanged capture/renderer policy; all promoted bytes match the reviewed actuals. Two fresh ordinary visual runs PASS `20261006T085747Z-p18381` and `-p18376` (12/12, all 47 scenarios each, 16 completed cleanup steps each); both reconcile all 255 active images against the same final golden manifest. Owner/type/drift/JSON/policy/catalog/Markdown/format/diff evidence is recorded above. Compatibility: only below-minimum visual placement changes; no API/state/credential change. Historical failures remain; final complete release still required. Next dependency S25am.

S25am DONE: the discovery accessibility matrix waits for the First control to be enabled before native End/Tab traversal and before reopened option activation. Visible cached options alone do not establish paging readiness. Real keyboard activation, scope, dismissal, density/zoom/text-spacing and bounds assertions remain; no production change or timeout increase. Original row PASS `20261006T090813Z-p15989` (11/11, all 14 cleanup steps complete), selector focus owner `20261006T090840Z-p47060` (2/2), pagination owner `-p47067` (2/2), typecheck `20261006T090814Z-p16097` (2/2), configured Markdown `-p16147`, format/diff. No corrected-run failure, compatibility effect or outstanding cleanup. Next dependency S25an, then renew the complete final candidate.

S25an DONE: the late-acceptance fixture observes completed Hosts grid loading and actual entry focus before deliberately focusing the Hosts tab, asserts that focus, then releases the original held receipt. It preserves late acceptance, retained receipt, source/support history and focus-retention assertions; no production focus change or timeout increase. All three original recovery browser scenarios PASS `20261006T091024Z-p64030` (11/11; all 14 cleanup steps complete), semantic focus owner `-p64031` (2/2), typecheck `-p64185` (2/2), configured Markdown `-p64239`, format/diff. No corrected-run failure, compatibility effect or residual cleanup. All active S16c–S24 and S25 prerequisites are now DONE. S25 remains IN_PROGRESS: renew agent-finalize → test-fast → one complete release-check, inspect canonical identities/readiness/cleanup and finish handoff. Earlier failed final candidates remain failures. Retained-run maintenance will be reported skipped when RESULTS_DIR is unset.

S25 sixth final-candidate sequence: agent-finalize PASS `20261006T091301Z-p6084` (1/1; generated artifacts unchanged; RESULTS_DIR unset, retained-run selection/performance maintenance/run checks explicitly skipped). Then test-fast PASS `20261006T091355Z-p10625` (746/746, zero failed/skipped/cancelled; 281.801 seconds; all 560 cleanup steps complete). Candidate source `sha256:36bdf0e90ffa1ebd71b32aa13b70d7b2f27bdc42395186fb13b5a4dbab984b63`, fast graph `sha256:98b7df56111459087a9350e29f072f36140bcf8b71f155d6200f20ee433834e2`, baseline `df720535b6b6622cab8b3d399e868dc968e0f6c4` with preserved dirty tree; toolchain `sha256:e28906e907c215e18aa7df021f87305dfae38bd6bdbda606f10dcce4969c5d99`, system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. Begin one complete release-check on this fixed candidate. No intermediate or historical result is final acceptance.

### S25 final candidate acceptance and operating handoff — 2026-10-06

**S25 DONE**, after S16c–S24 and every recorded prerequisite completed. The sixth complete candidate is the acceptance run; the first five failed candidates remain failed historical evidence. No partial-run results were combined into this disposition.

- Finalization: `make agent-finalize`, run `20261006T091301Z-p6084`, PASS 1/1; generated artifacts unchanged. `RESULTS_DIR` was unset, so retained-run selection, retained performance maintenance and retained-run checks were skipped explicitly.
- Fast verification: `make test-fast`, run `20261006T091355Z-p10625`, PASS 746/746, zero failed/skipped/cancelled, 281.801 seconds; all 560 cleanup steps completed.
- Complete release: `make release-check`, run **`20261006T091921Z-p85023`**, PASS **1,279/1,279**, zero failed/skipped/cancelled, 2,751.136 seconds. All **673 cleanup steps completed**, none failed or blocked. `make explain-run RESULTS_DIR=.cartulary/test-results/20261006T091921Z-p85023` confirms the canonical result and identities.
- All 19 required canonical readiness projections pass, including the three package smokes, pinned cryptographic policy, credential capacity, SeaweedFS gate, browser support/accessibility/visual evidence, builds, harness, security audit and release inventory. The closure-marker log is not the readiness result; the terminal `target-summaries/*.json` projections are.
- Source identity: `sha256:36bdf0e90ffa1ebd71b32aa13b70d7b2f27bdc42395186fb13b5a4dbab984b63`; baseline commit `df720535b6b6622cab8b3d399e868dc968e0f6c4`, preserved dirty working tree. Release graph `sha256:7429555b27f6c38420953a3e970e577cbee7be97f66e6a2a37b5116de1ba3d0a`; toolchain `sha256:e28906e907c215e18aa7df021f87305dfae38bd6bdbda606f10dcce4969c5d99`; system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. The candidate stayed fixed from finalization through release; only human tracker evidence was appended.
- Package receipts verify Go `1.27.1`, resolved selector `v1.0.0-c2097c7c`, service version `v1.0.0`, archive SHA-256 `daf3614e0406f67ae6323c902db3f953a1effb199142362a039e7526dfb9368b`. Packaged binary SHA-256: server `ef8d0f4c948506029668d2ddfac74059741fecf6526d1de1d4d90cba14d9e570`; migrate `6f8ebbdc22142d518d7b788679371bd01f2305eb59803b454cfd135f4891d188`; operator `1a926da7334e5d4859b216ca63ef0a7cb97347bba4407b275448dfa16971cd72`.
- SeaweedFS is pinned to stable `4.48`, index `sha256:4e61d15fd35994cb1e43e1e553dff106794841fd9a99ade2fc8c8bfce4d7872d`, resolved amd64 manifest `sha256:aba492e2a4e4c90bff795745e8e660affa1f09e7650f5981bd7bccd1a06cd931`. Final package environment receipts record this image identity. The general package application image is `sha256:118b010a7ddbc26d976eee8484912ed087eddb9bccc5794169126a481e90d1b6`; separately built Reference Pack and Recovery package image identities remain in their own receipts.
- Final reference environment: Windows 11 Pro `26200.9457`; WSL `2.7.3.0`, kernel `6.6.114.1-microsoft-standard-WSL2`; Ubuntu `26.04`; Intel Core Ultra 9 285HX, 24 visible CPUs; Docker Desktop `4.75.0.227598`, Engine/client `29.5.2`, Compose `5.1.3`, daemon `76196c3d-6136-43bc-831a-499136887d0e`, 33,353,146,368 bytes memory; ext4 `/dev/sdd`, NAT. The Desktop settings flag is unavailable/null, while the running `docker-desktop` WSL2 distribution and daemon/kernel identity are observed; no formal operating-environment claim is inferred from an image label.
- Credential capacity: actual cgroup limits two CPUs (`0-1`, quota `200000 100000`), 2,147,483,648 bytes memory, swap zero. P95 **346.286735 ms** against 2,000 ms; overload **10 accepted / 54 rejected**, maximum excess-work rejection **86.175468 ms** against 250 ms; cancellation and subsequent recovery pass.
- Package admission proves three facade receipts and three each disabled-mode, missing-identity and incompatible-state rejections, with rejected state unchanged and no disposable-state repair. Actual Windows HTTPS/WSS probes use isolated trust, reject TLS 1.2, wrong names, untrusted authorities, expiry and bad WSS origin. Certificate replacement covers all purpose leaves, wrong database purpose, owned-service recreation and application interruption. Recovery interruption creates zero successful backups/journals; matching-package backup, due verification and restore pass. Package, destination and workspace cleanup receipts all pass with zero remaining resource groups. Retained-secret scan passes across 9,229 files.

Canonical evidence is under `.cartulary/test-results/20261006T091921Z-p85023/`: `run-manifest.json`, `run-summary.json`, `cleanup-results.json`, `target-summaries/`, and each package's `artifacts/` receipts (`environment.json`, facade `*.crypto.json`, admission, Windows-client, rotation/interruption and cleanup results). Credential observations and effective limits are in `credential-capacity-assessment/artifacts/`. Main workbook accessibility, metadata conflicts, Assessment recovery, stateful query browsing and final visual reconciliation all pass in this same run; the eight reviewed golden updates also have the two separate ordinary passes recorded in S25al.

Operating handoff is [the MVP package guide](../../deploy/mvp/README.md), reviewed for fresh provisioning, certificate/key rotation, matching-release backup retention, restore, rejected-startup handling, integration restrictions, operator-started WSL2 restart/overdue-work checks and exact-module/platform upgrade qualification. The application cryptographic-format identity is `cartulary.application_crypto_format.v1`; existing deployments and historical backups remain with their matching releases. No conversion, automatic reset, Windows startup automation, customer rollout or global trust/firewall modification was performed. Existing user-owned development services remain borrowed and untouched. Changes remain uncommitted for review.

| Disposition | Final state | Basis / remaining authority |
| --- | --- | --- |
| Implementation completion | COMPLETE | S16c–S25 and every active implementation prerequisite passed their exits; complete current-candidate release passed. |
| WSL2 package acceptance | ACCEPTED for the recorded engineering reference environment | All three current package smokes, actual Windows-client boundaries, fixed performance checks and cleanup pass. |
| Formal CMVP applicability | UNESTABLISHED; no claim | Requires a separate formal applicability determination for the exact module/service/platform relationship. |
| Specification adoption | PENDING separate owner decision | Source amendments and machine projections are reconciled; engineering evidence does not itself adopt a specification. |
| Customer deployment approval | PENDING; no rollout authorized by this completion | Customer environment acceptance and operational approval remain separate. |

Native Linux remains indefinitely DEFERRED. No implementation or qualification failure has been reassigned to that deferred scope. There are no outstanding engineering prerequisites or owned-resource cleanup failures. Post-acceptance changes require appropriately renewed validation. S25 is completed last; S07 may now close for the Windows 11 / WSL2 engineering scope while the three external dispositions above remain separate.

S07 closure checkpoint: after S25 completion and the five dispositions above, **RP-S07 DONE for Windows 11 / WSL2 engineering**. RP-F29 is closed for accurate engineering readiness and handoff; the explicit external decisions remain pending/unestablished. Final `make lint-markdown` PASS `20261006T100919Z-p19767` for its configured `docs/*.md` and `docs/spec/**/*.md` coverage (which excludes this handoff); direct tracker/history review and `git diff --check` pass. The historical section remains byte-identical to the baseline. No engineering workstream or owned cleanup remains open.

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
