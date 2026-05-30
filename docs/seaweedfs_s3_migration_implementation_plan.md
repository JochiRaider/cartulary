---
title: Cartulary MinIO-to-SeaweedFS S3 Migration Implementation Plan
document_class: nlspec-style implementation plan
status: implementation-plan
created_at: 2026-05-30
revision_scope: Replace MinIO server usage with SeaweedFS S3-compatible object storage in Cartulary development, test, disconnected, and release-support surfaces while retaining minio-go as a generic Apache-2.0 S3 client.
---

# Cartulary MinIO-to-SeaweedFS S3 Migration Implementation Plan

## 0. Status, authority, and revision boundary

This artifact is an implementation plan written in NLSpec voice. It is not an adopted Cartulary subsystem NLSpec unless the repository authority process later adopts it.

Core 00 through Core 04 remain the implementation-conformance authority for current Cartulary product behavior. Core 05 remains claim-publication authority only. Future adopted subsystem NLSpecs may define bounded implementation-conformance requirements for their named subsystem only, and each adopted subsystem NLSpec must state scope, non-goals, owner interactions, and deployment-configuration effects. This plan follows that authority model and does not supersede it.[^1]

This plan owns the migration work required to replace default MinIO server usage with SeaweedFS S3 in local development, service-backed tests, disconnected deployment support, release support, documentation, and migration tooling. It must not add public route families, alter public request or response envelopes, alter evidence lifecycle semantics, alter session or authorization semantics, or widen the deployment-configuration schema without a patch to the primary owner document.

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative inside this plan. **SHOULD** and **SHOULD NOT** define strong migration defaults whose exceptions must remain compatible with all MUST-level requirements. **MAY** defines optional behavior only where omitted behavior is explicit.

Implementation is blocked when this plan requires an owner-document patch and that patch has not been made. A repository change that merely updates support docs, scripts, fixtures, or comments is not sufficient when the behavior is owned by Core 00 through Core 04.

## 1. Executive decision

Default Cartulary development, CI, service-backed test, disconnected, and release-support object storage MUST use **SeaweedFS S3** instead of a MinIO server.

This is a deployment, harness, documentation, compatibility, backup/restore, and migration-substitution change. It is not a product-model change. Cartulary remains a modular monolith with one application deployable, one Postgres service as the authoritative structured data store, and one S3-compatible object storage service as the authoritative binary evidence store.[^2]

The migration MUST preserve all of the following invariants:

- one application deployable;
- one Postgres service;
- one S3-compatible object-storage service;
- unchanged public `/api/v1/*` and `/ws/v1/*` route families;
- unchanged blob-slot create and attach semantics;
- unchanged evidence record lifecycle semantics;
- unchanged same-origin preview and download handle semantics;
- unchanged server-side authorization as the primary incident-data access boundary.

The MinIO server and `github.com/minio/minio-go/v7` are distinct migration subjects. The MinIO server MUST NOT be shipped, started, named, or documented as the default Cartulary object-store service. The `minio-go` package MUST remain allowed as a generic Apache-2.0 S3-compatible client library behind the internal object-store adapter.[^3]

## 2. Terminology

| Term | Required meaning |
| --- | --- |
| `MinIO server` | The MinIO object-storage server process, image, binary, container, service, fixture, or default runtime target. It is not a default Cartulary runtime dependency after this migration. |
| `minio-go` | The Apache-2.0 Go package `github.com/minio/minio-go/v7`, retained only as a generic S3-compatible client behind `internal/platform/objectstore`. |
| `SeaweedFS S3` | The SeaweedFS S3-compatible endpoint used as Cartulary's default local, disconnected, CI, and service-backed test object-store target. |
| `object-store adapter` | The internal platform boundary that owns S3-compatible object operations. Evidence, workbook, reporting, backup, and collaboration modules MUST NOT import SeaweedFS-specific code. |
| `direct upload target` | A short-lived upload target returned by `POST /api/v1/object-blobs` for exactly one pending blob slot. |
| `same-origin evidence handle` | An opaque application URL under `GET /api/v1/evidence-handles/{handle_token}` used for preview and download redemption. |
| `raw object key` | Any storage key, bucket-relative object name, filer path, volume identifier, backend path, or equivalent object-store identifier used to locate object bytes. |
| `legacy external S3 endpoint` | An operator-supplied S3-compatible endpoint that Cartulary does not ship, start, manage, test as the default target, or document as the default object-store service. |
| `profile` | One of the migration runtime profiles defined in §6. It is not a Core profile unless expressly named as such. |
| `operator-private artifact` | A retained artifact that may contain deployment-local object-store identifiers required for migration, backup, or restore. It MUST remain outside incident portability bundles and public user-facing responses. |
| `shareable summary` | A redacted human summary that may be attached to release notes, tickets, or implementation reports. It MUST NOT contain secrets, credentials, raw object keys, or backend URLs. |

## 3. Non-goals and preserved invariants

### 3.1 Non-goals

This migration MUST NOT introduce any behavior in the following table.

| Non-goal | Boundary |
| --- | --- |
| New user-visible storage abstraction | Users and browser clients MUST NOT choose, identify, or reason about SeaweedFS, MinIO, AWS S3, or another backend while using ordinary workbook and evidence routes. |
| SeaweedFS-specific public routes | No public `/api/v1/*` or `/ws/v1/*` route family may be added solely for SeaweedFS. |
| Product-model change | Evidence records, object blobs, workbook rows, projections, revisions, and collaboration messages keep their current owner semantics. |
| Object-key authority | Raw object keys, bucket names, filer paths, volume IDs, and object-store URLs MUST NOT become incident-portability source data or public row identity. |
| Bucket policy as primary authorization | Application authorization remains the primary incident-data authorization boundary. Object-store permissions are defense-in-depth and service isolation. |
| Multipart upload | Multipart upload is out of scope for this migration. Any later multipart support requires an owner document that defines thresholds, chunk ordering, abort semantics, retry behavior, and finalization. |
| SeaweedFS administrative dependency | Runtime product code MUST NOT call SeaweedFS master, filer, volume, or admin APIs for ordinary application behavior. |
| New deployment-config namespace | This plan MUST NOT add a SeaweedFS deployment-configuration namespace. Any new application config keys require an owner-document patch. |
| Live migration workflow | This plan does not add live maintenance-mode product behavior. The default migration mode is application-stopped migration. |

### 3.2 Preserved public evidence boundary

`POST /api/v1/object-blobs` remains the only public blob-slot creation route. A successful response continues to include `object_blob_id`, `upload_target`, `target_expires_at`, `pending_expires_at`, and `accepted_contract`; upload target expiry remains 60 minutes and pending slot expiry remains 24 hours under the Core-owned contract.[^4]

Preview and download MUST continue to be issued only through the existing preview-handle and download-handle routes and redeemed only through `GET /api/v1/evidence-handles/{handle_token}`. Those handles remain opaque, same-origin, short-lived, authorization-checked, and application-mediated. Public preview and download responses MUST NOT expose long-lived object-store credentials, bucket names, raw object keys, raw SeaweedFS URLs, or storage-backend-specific identifiers.[^5]

The only permitted backend-upload exposure is the `upload_target` returned by `POST /api/v1/object-blobs` when the active upload mode is `direct_presigned_put`. That exception is limited to uploading bytes for one pending blob slot. It MUST NOT be used for preview, download, listing, delete, or evidence discovery.

## 4. Closed decisions

| Decision ID | Decision | Required effect |
| --- | --- | --- |
| `SWFS-RD-001` | SeaweedFS S3 is the default local, disconnected, CI, service-backed test, and release-support object-store target. | Default Compose, harness, fixture, documentation, and release-support paths MUST use SeaweedFS S3 or generic S3-compatible wording. |
| `SWFS-RD-002` | MinIO server is no longer a supported default target. | MinIO server MAY appear only as a legacy external S3 endpoint configured and operated by the deployment owner. |
| `SWFS-RD-003` | Keep `github.com/minio/minio-go/v7` as a generic Apache-2.0 S3 client. | `minio-go` MUST remain behind `internal/platform/objectstore` and MUST NOT justify MinIO server fixtures, docs, services, admin APIs, or default runtime behavior. |
| `SWFS-RD-004` | Public Cartulary API behavior is unchanged except for any owner-approved additive error-code patch in §5. | No SeaweedFS-specific public route, public storage abstraction, or browser-visible backend identity may be added. |
| `SWFS-RD-005` | No new deployment-configuration namespace is introduced by this plan. | Any new application configuration key is blocked until the primary owner defines keys, defaults, null behavior, validation, redaction, and startup failure semantics. |
| `SWFS-RD-006` | The default SeaweedFS upload mode is `direct_presigned_put`. | Startup readiness MUST require presigned PUT compatibility for SeaweedFS profiles unless an owner document later defines an application-mediated upload profile. |
| `SWFS-RD-007` | Default migration preserves source bucket name and object keys. | Any bucket or key-shape change is a separate high-risk migration requiring explicit mapping, database migration, validation, and rollback rules. |
| `SWFS-RD-008` | Multipart upload is out of scope. | Compatibility, harness, backup, and migration tests MUST NOT treat multipart as required or skipped current-profile behavior. |
| `SWFS-RD-009` | Presigned GET is not part of public evidence access. | Public preview/download tests MUST use same-origin evidence handles, not presigned GET. |
| `SWFS-RD-010` | Runtime object-store dependency outages need explicit public error ownership. | The owner patch in §5.3 MUST land before any route-local runtime outage behavior is claimed complete. |

## 5. Owner-document patch plan

### 5.1 Required normative-core patches

| Target | Anchor | Required replacement |
| --- | --- | --- |
| `01_architecture_storage_and_view_contracts.md` | `§4.2 Object storage` | Replace any MinIO-as-default local/disconnected example with SeaweedFS S3 while preserving generic S3 compatibility and the same evidence, backup, restore, and access-handle contracts. |
| `04_security_deployment_and_conformance.md` | `§5.1 Flyaway or disconnected deployment` | Replace “one MinIO container or equivalent S3-compatible object store” with “one SeaweedFS S3 container or equivalent S3-compatible object store.” |
| `01_architecture_storage_and_view_contracts.md` | `§3.3.6.1 Canonical public error-code registry` | Add the two runtime object-store dependency error codes in §5.3, or identify an existing owner-approved equivalent with identical status, retryability, and reason-code semantics. |
| `04_security_deployment_and_conformance.md` | `§4.4 STRIDE threat model` | Add SeaweedFS S3 endpoint identity, credentials, object-store dependency errors, direct upload target scope, exposed admin surfaces, backup/restore, and migration validation to the required threat-model coverage. |

Core 04 owns the operator-facing deployment-configuration artifact, environment overlay grammar, unknown-key rejection, validation envelope, and fail-closed startup behavior. Implementation-support examples and `.env.example` MUST NOT widen that schema by implication.[^6]

### 5.2 Required implementation-support patches

| Target | Required update |
| --- | --- |
| `cartulary-dev-guide.md` | Distinguish `minio-go` as a generic S3 client from MinIO server. Replace default local and deployment prose that still names MinIO server. |
| `cartulary_implementation_testing_guide.md` | Replace service-backed fixture references to Postgres plus MinIO with Postgres plus SeaweedFS S3. Replace MinIO bucket terminology with canonical object-store terminology. |
| `testing-harness-nlspec.md` | Replace MinIO service and artifact vocabulary with generic object-store vocabulary and SeaweedFS-specific runtime profile rows where the default service is named. |
| `cartulary_repository_bootstrap_guide.md` | Replace default local service guidance and Docker Compose expectations with SeaweedFS S3. Keep `s3test` only as a generic package or helper name if the repository treats it as backend-neutral. |
| `.env.example` and Compose examples | Remove MinIO-server-specific defaults. Use existing owner-defined object-store configuration keys, or mark application-key changes as blocked until the owner patch exists. |
| README and release notes | State that SeaweedFS S3 is the default local/disconnected/test target and that `minio-go`, if present, is SDK-only. |

### 5.3 Required public error-code owner patch

The current Core registry has state-specific evidence errors and blob-create validation errors, but this migration needs route-local dependency-outage behavior after startup when object storage becomes unavailable. The repository MUST either add the following Core 01 rows or cite an owner-approved equivalent before implementation completion.

| New `error.code` | Required `error.status` | Required `error.retryable` | Canonical meaning |
| --- | ---: | --- | --- |
| `object_store_unavailable` | `503` | `true` | The route cannot complete because the configured object-store endpoint, bucket, or operation is temporarily unavailable after startup readiness previously succeeded. |
| `object_store_access_rejected` | `503` | `false` | The route cannot complete because object-store credentials, bucket permissions, CORS, or required capability no longer satisfy the configured runtime contract after startup readiness previously succeeded. |

Required `error.details.reason_code` values for both codes are:

| `reason_code` | Meaning | Allowed code |
| --- | --- | --- |
| `endpoint_unreachable` | Endpoint cannot be reached or returns transport failure before a storage operation can complete. | `object_store_unavailable` |
| `operation_timeout` | The object-store adapter deadline expires. | `object_store_unavailable` |
| `bucket_unavailable` | The configured bucket cannot be reached after prior readiness. | `object_store_unavailable` |
| `storage_internal_error` | Backend returns a transient server-side storage error. | `object_store_unavailable` |
| `permission_denied` | Credentials do not allow the required action. | `object_store_access_rejected` |
| `capability_missing` | The backend lacks a required operation for the active profile. | `object_store_access_rejected` |
| `cors_rejected` | Direct upload cannot proceed because CORS rejects the required browser upload path. | `object_store_access_rejected` |

These error codes MUST NOT replace existing state-specific errors. Missing, pending, failed, quarantined, inconsistent, unsupported preview, and preview-size conditions MUST continue to use existing Core-owned evidence errors when the object-store dependency is reachable and the authoritative evidence/blob state determines the result.

## 6. SeaweedFS runtime profiles

### 6.1 Runtime profile matrix

| Profile | Service name | Required topology | Host exposure | Persistence | Bucket creation | Readiness result |
| --- | --- | --- | --- | --- | --- | --- |
| `local_dev` | `seaweedfs-s3` | One SeaweedFS service containing master, filer, volume, and S3 gateway roles for one local deployment. | S3 endpoint MAY bind to localhost. Admin, master, filer, volume, and WebDAV surfaces MUST remain container-network-only unless `developer_debug` profile is explicitly selected. | Named Docker volume or `roots.object_storage`-equivalent local path. | Allowed for bucket `cartulary`. | Ready only after §10 capability probe succeeds. |
| `ci_service_backed` | `seaweedfs-s3` | One isolated SeaweedFS service per run, service suite, or scheduler lease. | Harness-owned network only. No public host exposure except the harness-controlled endpoint. | Ephemeral per run unless artifact retention explicitly preserves service state for diagnostics. | Allowed only for the run bucket or run prefix. | Product tests MUST NOT start before §10 capability probe succeeds. |
| `disconnected_prod` | Deployment-selected; default example `seaweedfs-s3` | One SeaweedFS S3 service paired with Cartulary app and Postgres. | Public exposure defaults to Cartulary app only. S3 endpoint is externally reachable only when direct upload requires it and §8 CORS rules pass. | Backed by `roots.object_storage` filesystem-root binding. | Application startup MUST NOT create buckets. Bucket must be pre-created or provisioned by operator tooling. | Startup MUST fail closed if endpoint, bucket, credentials, CORS, or required capability is missing. |
| `on_prem_or_cloud_external` | Not owned by this plan | Any operator-managed S3-compatible endpoint, including operator-managed SeaweedFS or legacy MinIO server. | Operator-owned. | Managed-service binding outside filesystem-root semantics. | Operator-owned. | Generic S3 capability probe MUST pass before ready state. |
| `developer_debug` | `seaweedfs-s3` | Same as `local_dev`. | May expose admin surfaces only on loopback or harness-owned network. | Same as selected base profile. | Same as selected base profile. | MUST NOT be used in production, CI release gates, or release manifests. |

### 6.2 Runtime service closure

`local_dev`, `ci_service_backed`, and `disconnected_prod` MUST use a single SeaweedFS service that provides an S3-compatible endpoint backed by colocated metadata and object storage. A multi-service SeaweedFS topology is out of scope for this migration unless a later owner document defines service count, port exposure, persistence layout, backup scope, and readiness semantics for that topology.

The SeaweedFS image reference MUST be pinned by tag and digest in repo-control files. `latest`, floating major tags, and unpinned image references are invalid for default development, CI, disconnected, and release-support profiles.

Readiness MUST be determined by the object-store capability probe in §10, not by a container-started flag alone.

### 6.3 Disconnected persistence layout

When `deployment_profile='disconnected'`, Core 04 requires `roots.object_storage` to use `binding_kind='filesystem_root'`.[^7] SeaweedFS state for the default disconnected profile MUST be rooted under that filesystem root as follows.

| Path under `roots.object_storage.path` | Required contents | Backup criticality | Secret-bearing | Restore rule |
| --- | --- | --- | --- | --- |
| `seaweedfs/master/` | Master or equivalent topology metadata. | Required. | No. | Restore before readiness probe. |
| `seaweedfs/filer/` | Filer or equivalent namespace metadata. | Required. | No. | Restore before readiness probe. |
| `seaweedfs/volume/` | Object bytes and volume files. | Required. | Contains incident evidence bytes but not credentials by design. | Restore before readiness probe. |
| `seaweedfs/tmp/` | Runtime scratch only. | Not authoritative. | No. | MAY be omitted on backup and recreated on startup. |
| outside `roots.object_storage.path` | S3 access credentials and service-local secrets. | Not part of evidence backup. | Yes. | Restored only through deployment-local secret procedure. |

The application MUST NOT treat SeaweedFS metadata paths, volume files, or object keys as incident-portability source data. Whole-incident portability remains governed by the incident-portability owner documents, not by SeaweedFS filesystem layout.

## 7. Configuration and secret boundary

This migration MUST NOT add new Cartulary deployment-configuration keys. Runtime application configuration MUST continue to use the adopted object-store configuration contract. If the repository currently uses MinIO-branded application configuration names, renaming them to backend-neutral names requires an owner patch that defines exact keys, types, defaults, omitted behavior, explicit-`null` behavior, environment overlays, validation errors, secret handling, and startup failure semantics.

Compose-only variables used solely to configure the SeaweedFS container are service-local variables. They MUST NOT be documented as Cartulary deployment-configuration keys. They MUST NOT participate in Core 04 deployment-configuration validation unless an owner patch adopts them as application configuration.

Secret values MUST be provided through deployment-local secret handling. Raw access keys, secret keys, session tokens, presigned URLs, CORS secrets, and service-local credential files MUST NOT appear in incident records, workbook rows, public route responses other than the short-lived `upload_target` exception, portability bundles, retained harness summaries, shareable migration summaries, screenshots, logs, or release notes.

The public browser origin used for direct-upload CORS MUST be `application.public_origin`, the stable deployment-configuration key owned by Core 04.[^8]

## 8. Credential, bucket, prefix, and exposure policy

### 8.1 Credential matrix

| Environment | Identity | Required actions | Forbidden actions | Scope | Retention |
| --- | --- | --- | --- | --- | --- |
| `local_dev` | `cartulary_dev` | Read, Write, Head, Delete, List for bucket `cartulary`; dev-only bucket bootstrap. | Anonymous access; production credentials; wildcard service credentials. | Bucket `cartulary`. | Secret may live in dev-only local secret file ignored by VCS. |
| `ci_service_backed` | `cartulary_test_<run_segment>` | Read, Write, Head, Delete, List for run-owned bucket or run-owned prefix; test-only bucket bootstrap. | Cross-run access; persisted credentials after run cleanup. | One run bucket or one run prefix. | Destroy at end of run unless diagnostics retained. |
| `disconnected_prod` | `cartulary_app` | Read, Write, Head, Delete only where required by evidence cleanup, and List only for backup, diagnostics, and restore verification. | `Admin`, anonymous access, wildcard access, app startup bucket creation. | One production bucket. | Deployment-local secret procedure. |
| `migration` | `cartulary_migration_<run_segment>` | Read/List/Head/Get from source, Write/List/Head/Get/Delete for migration-owned target test objects, and Write for copied target objects. | Reuse as application credential after migration. | Source and target buckets for one migration run. | Revoke after `post_cutover_verified` or `rolled_back`. |

### 8.2 Bucket and prefix rules

| Environment | Bucket default | Bucket creation | Prefix rule | Cleanup rule |
| --- | --- | --- | --- | --- |
| `local_dev` | `cartulary` | Allowed. | No required prefix. | Developer reset MAY clear only the local development bucket after an explicit reset command. |
| `ci_service_backed` | `cartulary-test` unless per-run buckets are used. | Allowed only by harness-owned setup. | Prefix MUST be `test-runs/<run_segment>/<scope_segment>/` when shared bucket is used. | Harness MAY delete only the run-owned bucket or run-owned prefix. |
| `disconnected_prod` | No implicit application default unless already owner-defined. | Forbidden during application startup. | Application-generated server-side storage refs only. | Application MUST NOT bulk-delete production prefixes outside authorized cleanup flows. |
| `migration` | Preserve source bucket name by default. | Target bucket must exist before copy unless operator tooling owns creation. | Preserve source object keys by default. | Validation cleanup MAY delete only probe and migration-test objects. |

`run_segment` MUST be the first 32 lowercase hexadecimal characters of SHA-256 over the harness run ID bytes. `scope_segment` MUST be the first 32 lowercase hexadecimal characters of SHA-256 over the package, suite, or scheduler-scope string. Empty scope string normalizes to `default` before hashing.

Bucket names used by Cartulary-managed dev and CI profiles MUST match this grammar:

```text
^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$
```

They MUST NOT contain `..`, `.-`, `-.`, uppercase letters, underscores, path separators, leading or trailing hyphens, leading or trailing dots, or an IPv4-address-like form.

### 8.3 Exposure matrix

| Surface | `local_dev` | `ci_service_backed` | `disconnected_prod` | `developer_debug` |
| --- | --- | --- | --- | --- |
| Cartulary app HTTP/HTTPS | Exposed as configured. | Harness-owned only. | Exposed as the public app surface. | Same as base profile. |
| SeaweedFS S3 endpoint | MAY bind localhost. | Harness-owned only. | Not publicly exposed unless direct upload requires browser reachability. | MAY bind loopback only. |
| SeaweedFS master UI/API | Container network only. | Container network only. | Forbidden. | MAY bind loopback only. |
| SeaweedFS filer UI/API | Container network only. | Container network only. | Forbidden. | MAY bind loopback only. |
| SeaweedFS volume UI/API | Container network only. | Container network only. | Forbidden. | MAY bind loopback only. |
| WebDAV or non-S3 protocols | Disabled. | Disabled. | Forbidden. | MAY be enabled only outside release manifests. |
| Metrics or debug endpoints | Disabled unless already owned by another adopted subsystem. | Harness-owned if present. | Forbidden unless owner document defines exposure and authentication. | MAY bind loopback only. |

### 8.4 Direct-upload CORS policy

When `upload_mode='direct_presigned_put'`, CORS for the SeaweedFS S3 endpoint MUST satisfy all rows below.

| Rule | Required value |
| --- | --- |
| Allowed origin | Exactly `application.public_origin`. Wildcard origins are forbidden outside throwaway local development. |
| Allowed methods | `PUT` and `OPTIONS` only. |
| Disallowed browser methods | `GET`, `DELETE`, `POST`, `LIST`, and bucket enumeration. |
| Allowed headers | Only headers required by the generated upload target, including content type and checksum headers when present. |
| Exposed headers | `ETag` only when finalization or client diagnostics require it; otherwise no exposed object-store headers. |
| Credentials | Browser credentials to the object-store endpoint are forbidden. Authorization comes from the short-lived upload target. |
| Expiry | The upload target expiry remains the Core-owned 60 minutes. CORS MUST NOT extend or refresh it. |

## 9. Object-store adapter contract

### 9.1 Adapter boundary

All object-store access MUST pass through `internal/platform/objectstore` or an equivalent internal object-store adapter boundary. Product modules MUST NOT import SeaweedFS client code, SeaweedFS admin code, MinIO server code, or bucket-management APIs directly.

The adapter MAY use `github.com/minio/minio-go/v7` as its implementation client. That dependency is SDK-only and MUST remain replaceable by another S3-compatible client without changing public API behavior.

### 9.2 Operation contract

| Operation | Inputs | Output | Default timeout | Retry policy | Idempotency | Errors | Production permission |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `CreateUploadTarget` | `bucket`, server-generated `key`, `byte_size`, optional `content_type`, optional `sha256_hex`, `expires_at`, `upload_mode`. | `upload_target` bound to one method, one key, one size contract, and one expiry. | 10s. | Retry at most 2 attempts on transport timeout before target is returned. | MUST NOT create more than one authoritative pending blob slot. | `object_store_unavailable`, `object_store_access_rejected`, `object_store_invalid_request`. | Allowed. |
| `PutObject` | `bucket`, server-generated `key`, byte stream, declared `size`, optional metadata. | `etag` if available, `size_bytes`, stored metadata echo. | 30s to first byte accepted; stream read governed by caller context. | Retry only when no bytes were accepted by backend and client can replay stream. Max 2 attempts. | Same `(bucket,key)` overwrite is allowed only for probe/test/migration overwrite rules; product evidence keys are write-once. | Same plus `object_store_integrity_mismatch`. | Allowed for app-mediated upload if owner profile exists; otherwise probe/test/migration only. |
| `HeadObject` | `bucket`, `key`. | `exists`, `size_bytes`, `etag` if available, metadata required for finalization. | 10s. | Max 2 attempts on transient transport failure. | Read operation. | `object_not_found`, dependency errors. | Allowed. |
| `GetObject` | `bucket`, `key`. | Stream plus `size_bytes` and metadata when available. | 30s to first byte; 30s inactivity timeout after first byte. | No adapter-level retry after bytes are emitted. | Read operation. | `object_not_found`, dependency errors. | Allowed. |
| `GetObjectRange` | `bucket`, `key`, required inclusive `start`, optional inclusive `end`. | Stream of bytes `start..end` inclusive, or `start..EOF` when `end` omitted. | 30s to first byte; 30s inactivity timeout. | No adapter-level retry after bytes are emitted. | Read operation. | `object_range_not_satisfiable`, `object_not_found`, dependency errors. | Allowed. |
| `DeleteObject` | `bucket`, `key`, `purpose`. | `deleted=true` or `not_found=true`. | 30s. | Max 2 attempts on transient transport failure. | Idempotent when object is already absent. | Dependency errors. | Allowed only for probe cleanup, test cleanup, authorized cleanup, and migration validation cleanup. |
| `ListPrefix` | `bucket`, `prefix`, `purpose`, optional continuation token. | Ordered object summaries and continuation token. | 30s per page. | Max 2 attempts per page before returning failure. | Read operation. | Dependency errors. | Allowed only for test cleanup, backup verification, migration validation, and diagnostics. Never a user-facing query primitive. |
| `EnsureBucketForDevTest` | `bucket`, `profile`. | `created`, `already_exists`, or failure. | 30s. | Max 2 attempts. | Idempotent for existing bucket. | Dependency errors. | Forbidden in `disconnected_prod` and `on_prem_or_cloud_external` application startup. |

### 9.3 Range and stream rules

`GetObjectRange.start` and `GetObjectRange.end` use zero-based inclusive byte positions. `start` is required. Omitted `end` means stream from `start` through EOF. `start > end` is an adapter caller error and MUST NOT issue a backend request. A range beginning at EOF or beyond EOF MUST map to `object_range_not_satisfiable`.

All successful stream outputs MUST require the caller to close the stream. The adapter MUST make stream closure observable in tests by exposing a close hook or leak detector in test builds. A leaked stream in adapter tests is a harness or product test failure according to the owning test row.

### 9.4 Internal adapter error registry

| Internal error | Meaning | Public mapping family |
| --- | --- | --- |
| `object_store_unavailable` | Endpoint, bucket, or operation is temporarily unavailable. | `object_store_unavailable`. |
| `object_store_access_rejected` | Credentials, CORS, permission, or capability blocks required behavior. | `object_store_access_rejected`. |
| `object_not_found` | The requested object key is absent or not visible to the configured credential. | Evidence state-specific errors when tied to evidence; otherwise adapter-only. |
| `object_range_not_satisfiable` | Requested range is outside object byte domain. | Evidence access failure or adapter test failure, depending caller. |
| `object_store_integrity_mismatch` | Observed size or checksum conflicts with accepted contract. | `evidence_attach_rejected/accepted_contract_mismatch` or validation artifact mismatch. |
| `object_store_invalid_request` | Adapter caller supplied invalid bucket, key, range, metadata, purpose, or unsupported profile operation. | Product bug in tests; public route must fail earlier with owner-owned validation where possible. |
| `object_store_cleanup_failed` | Cleanup after probe/test/migration left a reserved object behind. | Harness artifact failure or startup warning as defined in §10. |

## 10. Capability probe contract

### 10.1 Probe inputs and defaults

| Field | Default or rule |
| --- | --- |
| `probe_id` | 26-character lowercase ULID or equivalent CSPRNG-backed stable ID generated once per startup attempt. |
| Probe prefix | `.cartulary/probes/startup/<probe_id>/`. |
| Primary payload | UTF-8 bytes `cartulary-object-store-probe-v1\n` followed by 16 CSPRNG bytes. |
| Secondary direct-upload payload | UTF-8 bytes `cartulary-object-store-direct-put-probe-v1\n` followed by 16 CSPRNG bytes. |
| Total deadline | 60s. |
| Endpoint attempt timeout | 5s per attempt. |
| Metadata attempt timeout | 10s per attempt. |
| Object mutation timeout | 30s per attempt. |
| Retry count | At most 2 attempts per required step, no unbounded retry. |
| Cleanup | Delete every probe object created by the current run before returning success or failure. |

### 10.2 Probe algorithm

```text
probe_object_store(profile, upload_mode):
  probe_id = generate_probe_id()
  prefix = ".cartulary/probes/startup/" + probe_id + "/"
  key_primary = prefix + "probe.bin"
  key_secondary = prefix + "direct-put.bin"
  payload = utf8("cartulary-object-store-probe-v1\n") + random_16_bytes()
  direct_payload = utf8("cartulary-object-store-direct-put-probe-v1\n") + random_16_bytes()
  deadline = now() + 60 seconds

  require endpoint reachable within 5 seconds

  if bucket is missing:
      if profile in {local_dev, ci_service_backed}:
          EnsureBucketForDevTest(bucket, profile)
      else:
          fail object_store_access_rejected/capability_missing

  PutObject(bucket, key_primary, payload, size=len(payload), purpose="startup_probe")
  HeadObject(bucket, key_primary) and require size_bytes == len(payload)
  GetObject(bucket, key_primary) and require bytes == payload
  GetObjectRange(bucket, key_primary, start=0, end=8) and require bytes == payload[0:9]

  if upload_mode == "direct_presigned_put":
      target = CreateUploadTarget(bucket, key_secondary, len(direct_payload), expires_at=now()+5 minutes)
      perform HTTP PUT through target with exact direct_payload
      HeadObject(bucket, key_secondary) and require size_bytes == len(direct_payload)
      GetObject(bucket, key_secondary) and require bytes == direct_payload

  DeleteObject(bucket, key_primary, purpose="startup_probe_cleanup")
  DeleteObject(bucket, key_secondary, purpose="startup_probe_cleanup") if created
  require HeadObject(bucket, key_primary) maps to object_not_found
  require HeadObject(bucket, key_secondary) maps to object_not_found if created
  return success

on any failure:
  attempt DeleteObject for every probe key created in this run
  return one normalized failure with stage, reason_code, retryable, and redacted endpoint identity
```

### 10.3 Probe failure matrix

| Failure | Startup result | Runtime result after prior readiness |
| --- | --- | --- |
| Endpoint unreachable | Fail readiness. HTTP, WebSocket, and job listeners MUST NOT enter ready state when object storage is required. | Evidence upload, attach finalization, preview, and download fail with `object_store_unavailable/endpoint_unreachable`. Ordinary non-evidence row editing continues. |
| Bucket missing | Dev and CI MAY create if profile allows. Production startup fails. | Evidence paths fail with `object_store_unavailable/bucket_unavailable` unless an operator repairs the bucket and readiness is re-established. |
| Permission denied | Startup fails. | Evidence paths fail with `object_store_access_rejected/permission_denied`. |
| Range unavailable | Startup fails when preview/download implementation requires range reads. | Preview/download requiring range fails; full-download MAY continue only if the owner contract allows full-stream download without range. |
| Presigned PUT unavailable | Startup fails when `upload_mode='direct_presigned_put'`. | New blob-slot creation fails with `object_store_access_rejected/capability_missing`; existing pending slots do not attach unless uploaded bytes are observed. |
| CORS rejected | Startup fails for browser-reachable direct-upload production profiles after CORS probe or preflight validation. | New direct upload targets fail with `object_store_access_rejected/cors_rejected`. |
| Cleanup delete fails after all functional steps pass | Startup MAY continue only when the retained object is under the reserved probe prefix and diagnostics record cleanup failure. | Harness cleanup failure is an artifact/harness failure, not product success. |

## 11. Evidence route behavior and public error mapping

### 11.1 Runtime dependency outage mapping

| Adapter condition | Blob create / upload target | Attach finalization | Preview/download issuance | Handle redeem |
| --- | --- | --- | --- | --- |
| `object_store_unavailable` | `503 object_store_unavailable` with precise reason. | `503 object_store_unavailable` if uploaded bytes cannot be observed. | `503 object_store_unavailable` if bytes or metadata cannot be reached. | `503 object_store_unavailable` if bytes cannot be reached. |
| `object_store_access_rejected` | `503 object_store_access_rejected` with precise reason. | `503 object_store_access_rejected`; evidence MUST NOT attach. | `503 object_store_access_rejected`; handle MUST NOT issue. | `503 object_store_access_rejected`; bytes MUST NOT stream. |
| `object_not_found` | Not applicable before upload. | `409 evidence_attach_rejected/blob_not_visible` when authoritative blob is absent or not visible. | `409 evidence_access_unavailable/blob_missing` when metadata points at unavailable bytes. | `409 evidence_access_unavailable/blob_missing`. |
| `object_store_integrity_mismatch` | Not applicable. | `409 evidence_attach_rejected/accepted_contract_mismatch`. | `409 evidence_access_unavailable/evidence_inconsistent` if discovered after attach. | `409 evidence_access_unavailable/evidence_inconsistent`. |
| `object_range_not_satisfiable` | Not applicable. | Not applicable. | `409 evidence_access_unavailable/evidence_inconsistent` unless preview-size or unsupported-preview owner rules are more specific. | Same as issuance. |

This table is contingent on the owner patch in §5.3. If the owner patch is not present, any route-local runtime dependency outage row in this table is blocked and MUST NOT be claimed complete.

### 11.2 Evidence state matrix

| State | Upload target use | Attach finalization | Preview-handle issuance | Download-handle issuance | Handle redeem after issuance |
| --- | --- | --- | --- | --- | --- |
| Pending blob | Upload may proceed before target expiry. | Reject with `evidence_attach_rejected/blob_pending`. | Reject with `evidence_access_unavailable/blob_pending`. | Reject with `evidence_access_unavailable/blob_pending`. | Reject with `evidence_access_unavailable/blob_pending`. |
| Failed blob | Same slot cannot be refreshed. | Reject with `evidence_attach_rejected/blob_failed`. | Reject with `evidence_access_unavailable/blob_failed`. | Reject with `evidence_access_unavailable/blob_failed`. | Reject with `evidence_access_unavailable/blob_failed`. |
| Missing backing object | No effect on new create. | Reject with `evidence_attach_rejected/blob_not_visible` unless still pending and observable. | Reject with `evidence_access_unavailable/blob_missing`. | Reject with `evidence_access_unavailable/blob_missing`. | Reject with `evidence_access_unavailable/blob_missing`. |
| Quarantined blob | No attach. | Reject with `evidence_attach_rejected/blob_quarantined`. | Reject with `evidence_access_unavailable/evidence_quarantined`. | Reject with `evidence_access_unavailable/evidence_quarantined`. | Reject with `evidence_access_unavailable/evidence_quarantined`. |
| Quarantined evidence | No attach. | Reject with `evidence_attach_rejected/evidence_quarantined`. | Reject with `evidence_access_unavailable/evidence_quarantined`. | Reject with `evidence_access_unavailable/evidence_quarantined`. | Reject with `evidence_access_unavailable/evidence_quarantined`. |
| Oversized preview | Not applicable. | Not applicable. | Reject with `evidence_access_unavailable/preview_payload_too_large`. | May issue when download otherwise allowed. | Existing preview handle MUST NOT have been issued for oversized preview. |
| Unsupported preview | Not applicable. | Not applicable. | Reject with `evidence_access_unavailable/unsupported_preview`. | May issue when download otherwise allowed. | Existing preview handle MUST NOT have been issued for unsupported preview. |
| Expired upload target | Same slot cannot refresh. | Reject if bytes were not successfully observed before expiry. | Reject while blob remains pending. | Reject while blob remains pending. | Reject while blob remains pending. |
| Expired handle | Not applicable. | Not applicable. | Fresh issuance may succeed. | Fresh issuance may succeed. | `410 handle_expired`. |
| Consumed download handle | Not applicable. | Not applicable. | Not applicable. | Fresh issuance may succeed. | `410 handle_consumed`. |
| Stale row version on attach | Not applicable. | `409 row_version_conflict`. | Not applicable. | Not applicable. | Not applicable. |

### 11.3 Startup outage versus runtime outage

| Condition | Readiness | Non-evidence workbook operations | Evidence operations | Diagnostics |
| --- | --- | --- | --- | --- |
| Invalid object-store configuration before startup | Not ready; process exits non-zero or refuses listeners. | Unavailable because application is not ready. | Unavailable because application is not ready. | Startup diagnostic with redacted endpoint and reason. |
| Object-store service unreachable before startup | Not ready. | Unavailable because application is not ready. | Unavailable because application is not ready. | Startup probe failure artifact. |
| Object-store outage after ready | Ready or degraded according to health contract; app remains able to authorize ordinary requests. | Must continue when no object-store dependency is required. | Route-local failures per §11.1. | Runtime dependency diagnostic with reason code. |
| Object-store restored after outage | Readiness may return to healthy after probe or operation succeeds. | Continues. | New evidence operations may succeed; failed or expired blob slots remain governed by lifecycle rules and do not auto-refresh. | Recovery event in diagnostics. |

## 12. Harness and compatibility profile

### 12.1 Canonical artifact vocabulary

Harness, compatibility, and migration artifacts MUST use backend-neutral fields. Backend identity belongs in `object_store_backend`.

| Field | Required use |
| --- | --- |
| `object_store_backend` | Exact value `seaweedfs_s3` for default SeaweedFS runs. |
| `s3_bucket` | Canonical bucket field. |
| `s3_prefix` | Canonical prefix field when prefix isolation is used. |
| `object_store_endpoint` | Endpoint value with credentials redacted. |
| `object_store_capability_probe` | Structured probe result. |
| `object_store_failure_reason` | Stable harness reason code. |

The following artifact fields are forbidden in new or revised default artifacts: `minio_bucket`, `minio_endpoint`, `minio_container`, `minio_ready`, `minio_access_key`, and `minio_secret_key`.

### 12.2 Compatibility cases

| Case ID | Capability | Required test input | Required result |
| --- | --- | --- | --- |
| `SWFS-COMP-001` | Bucket validation | Configured dev/test bucket. | Bucket exists or is created only in dev/test profile. |
| `SWFS-COMP-002` | Put object | Payloads: 0 bytes, 37 bytes, and 1 MiB + 13 bytes. | Exact byte length stored for each payload. |
| `SWFS-COMP-003` | Head object | Objects from `SWFS-COMP-002`. | Existence and size match input. |
| `SWFS-COMP-004` | Full get object | Objects from `SWFS-COMP-002`. | Returned bytes equal input bytes. |
| `SWFS-COMP-005` | Range get | Payload at least 4096 bytes; range `[4, 15]`. | Returned bytes equal input byte offsets 4 through 15 inclusive. |
| `SWFS-COMP-006` | Delete object | One created object. | Subsequent head maps to `object_not_found`. |
| `SWFS-COMP-007` | Prefix isolation | At least two prefixes in one shared bucket. | Listing one test prefix returns only objects under that prefix. |
| `SWFS-COMP-008` | Presigned PUT | One upload target with 5-minute test expiry. | PUT before expiry succeeds; PUT after expiry fails; neither path attaches evidence without finalization. |
| `SWFS-COMP-009` | Error classification | Missing object, denied credential, missing bucket, unreachable endpoint, and integrity mismatch fixtures. | Each maps to the backend-neutral error registry. |
| `SWFS-COMP-010` | No public listing primitive | Product route inventory and UI route registry. | No public API or workbook route exposes object prefix listing. |
| `SWFS-COMP-011` | CORS preflight | Direct-upload browser origin equal to `application.public_origin`. | Browser preflight allows only the §8.4 direct-upload method/header set. |
| `SWFS-COMP-012` | Same-origin preview/download | Existing evidence preview and download flow. | Public evidence access uses same-origin handles only. |

No compatibility row may be skipped because multipart upload or presigned GET is absent. Those capabilities are outside the current migration profile.

## 13. Backup and restore contract

Core 01 requires successful backup and restore behavior to cover Postgres and object-store contents from the same retained `backup_set` and `consistency_point_at`; restore must restore Postgres, restore object-store contents, and rebuild projections.[^9]

### 13.1 SeaweedFS object manifest

A successful backup against SeaweedFS MUST include one object-store backup manifest with this schema.

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_backup_manifest.v1`. |
| `backup_set_id` | string | Must match enclosing backup set. |
| `consistency_point_at` | RFC3339 timestamp | Must match enclosing backup set. |
| `object_store_backend` | string | `seaweedfs_s3`. |
| `bucket` | string | Required in operator-private artifact; may be redacted in shareable summary. |
| `object_count` | integer | Count of manifest objects. |
| `total_size_bytes` | integer | Sum of object byte sizes. |
| `objects[]` | array | Sorted by `storage_ref` bytewise ascending when present, otherwise by `storage_ref_sha256`. |
| `objects[].object_blob_id` | string | Required when object corresponds to an authoritative blob. |
| `objects[].storage_ref` | string or omitted | Required in restoreable operator-private manifest unless equivalent restore mapping exists. |
| `objects[].storage_ref_sha256` | lowercase hex SHA-256 | Required. |
| `objects[].size_bytes` | integer | Required. |
| `objects[].sha256` | lowercase hex SHA-256 or null | Null only when authoritative hash is absent and backup mechanism cannot compute one. |
| `objects[].backup_member_sha256` | lowercase hex SHA-256 | Digest of the backup artifact member bytes. |
| `manifest_sha256` | lowercase hex SHA-256 | Digest of canonical manifest bytes excluding this field. |

### 13.2 Canonical manifest serialization

The manifest MUST be serialized as UTF-8 JSON with lexicographically sorted object keys, no insignificant whitespace, LF line ending, and arrays in their specified order. `manifest_sha256` is computed over the canonical bytes with `manifest_sha256` omitted. A manifest with duplicate JSON object keys is invalid.

### 13.3 Restore verification

Restore verification MUST execute this sequence:

1. Select exactly one retained `backup_set`.
2. Restore Postgres from that `backup_set`.
3. Restore SeaweedFS object data from that same `backup_set`.
4. Rebuild projections.
5. Verify that every authoritative blob requiring durable bytes has bytes present in the restored object store.
6. Verify manifest size and hash proofs where hashes exist.
7. Open at least one restored incident when incident data exists.
8. Run one built-in workbook query when incident data exists.
9. Mark verification failed if any required blob byte, manifest proof, lifecycle invariant, or workbook query fails.

A backup MUST NOT be classified as the latest successful retained backup if the object-store manifest is missing, unreadable, inconsistent with the backup set, or detached from the Postgres consistency point.

## 14. MinIO-to-SeaweedFS migration contract

### 14.1 Migration posture

The default migration mode is application-stopped migration. The running Cartulary application MUST be stopped or otherwise incapable of accepting writes before coherent backup capture, object copy, validation, or cutover begins.

No live maintenance-mode product workflow is introduced by this plan. A later live migration feature requires a separate owner document defining write freeze, public route outcomes, operator state, background jobs, and rollback behavior.

### 14.2 Migration state machine

| State | Entry condition | Allowed next states | Blocking failure result |
| --- | --- | --- | --- |
| `planned` | Operator has selected source and target endpoints. | `preflighted`, `failed`. | `failed`. |
| `preflighted` | Source and target credentials validated; source backup target available. | `application_stopped`, `failed`. | `failed`. |
| `application_stopped` | Cartulary app cannot accept writes. | `backup_captured`, `failed`. | `failed`. |
| `backup_captured` | Coherent pre-migration Postgres plus source object-store backup exists. | `target_prepared`, `rolled_back`, `failed`. | `rolled_back` when source remains authoritative. |
| `target_prepared` | SeaweedFS target started; target bucket exists; capability probe passes. | `copying`, `failed`. | `failed`. |
| `copying` | Byte-for-byte object copy is running. | `copied`, `failed`. | `failed`; source remains authoritative. |
| `copied` | Copy tool reports completion. | `validating`, `failed`. | `failed`; source remains authoritative. |
| `validating` | Validation artifact is being produced. | `cutover_ready`, `failed`. | `failed`; source remains authoritative. |
| `cutover_ready` | Validation artifact has `result='pass'`. | `cutover_committed`, `rolled_back`. | `rolled_back`. |
| `cutover_committed` | Application object-store endpoint points to SeaweedFS. | `post_cutover_verified`, `rolled_back`, `failed`. | `rolled_back` if verification fails before release. |
| `post_cutover_verified` | Evidence and backup/restore gates pass against SeaweedFS. | terminal. | N/A. |
| `rolled_back` | Endpoint restored to source and pre-migration backup retained. | terminal. | N/A. |
| `failed` | Blocking precondition or validation gate failed. | terminal unless operator restarts from `planned`. | N/A. |

### 14.3 Copy semantics

Copy idempotency key MUST be:

```text
(source_bucket, source_key, target_bucket, target_key, source_size_bytes, source_hash_or_etag)
```

Default copy MUST preserve source bucket name and object keys. The migration utility MUST NOT update database `storage_ref` values during default migration.

If a target object exists with the same size and hash, copy MAY skip it and count it as already copied. If a target object exists with different size or hash, copy MUST fail that object and block cutover. If no reliable source hash exists, the copy utility MUST compute SHA-256 from source bytes before treating a target object as equivalent.

### 14.4 Rollback semantics

| Migration boundary | Rollback behavior |
| --- | --- |
| Before `cutover_committed` | Leave Cartulary pointed at the source object store and retain the pre-migration backup. |
| After `cutover_committed` and before `post_cutover_verified` | Stop the app; restore the pre-migration coherent backup if any writes occurred after cutover; point the app back to the source object store; verify the source path before reopening. |
| After `post_cutover_verified` | Migration rollback is closed. Recovery uses ordinary retained-backup restore procedures. |

## 15. Migration validation artifact

### 15.1 Schema

The migration validation tool MUST emit one operator-private artifact conforming to this schema and MAY emit a redacted shareable summary derived from it.

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_migration_validation.v1`. |
| `run_id` | string | Stable migration run identifier. |
| `started_at` | RFC3339 timestamp | Required. |
| `completed_at` | RFC3339 timestamp or null | Null only while running. |
| `source_backend` | string | `minio_s3` for default MinIO-source migration; otherwise backend-neutral source label. |
| `target_backend` | string | `seaweedfs_s3`. |
| `source_bucket` | string or redacted object | Required in operator-private artifact. |
| `target_bucket` | string or redacted object | Required in operator-private artifact. |
| `incident_count` | integer | Count of incidents checked. |
| `object_blob_count` | integer | Count of authoritative object blobs checked. |
| `objects_checked[]` | array | Sorted by `object_blob_id asc`; may be omitted from shareable summary if full artifact is retained. |
| `missing_source_objects[]` | array | Empty required for trusted migration. |
| `missing_target_objects[]` | array | Empty required for cutover. |
| `size_mismatches[]` | array | Empty required for cutover. |
| `hash_mismatches[]` | array | Empty required for cutover when expected hash exists. |
| `preview_sample_checks[]` | array | Deterministic bounded sample from §15.3. |
| `blocking_diagnostics[]` | array | Empty required for `result='pass'`. |
| `nonblocking_warnings[]` | array | May be non-empty for `result='pass'`. |
| `result` | enum | Exactly `pass` or `fail`. |

### 15.2 Mismatch item shape

| Field | Required rule |
| --- | --- |
| `object_blob_id` | Required. |
| `incident_id` | Required. |
| `storage_ref_sha256` | Required. |
| `expected_size_bytes` | Required when size mismatch is reported. |
| `actual_size_bytes` | Required when target exists. |
| `expected_sha256` | Required when authoritative hash exists. |
| `actual_sha256` | Required when hash was computed. |
| `reason_code` | Closed reason token from the validation tool registry. |

### 15.3 Preview sample algorithm

```text
sample_preview_checks(objects):
  candidates = objects where evidence is previewable or downloadable through ordinary app routes
  sort candidates by sha256(object_blob_id) ascending
  return first min(32, len(candidates))
```

### 15.4 Result computation

```text
result = "pass" only if:
  missing_source_objects is empty
  missing_target_objects is empty
  size_mismatches is empty
  hash_mismatches is empty
  blocking_diagnostics is empty
  every preview_sample_checks item has status == "pass"
else result = "fail"
```

Automation MUST be able to decide cutover eligibility from the artifact without reading prose.

## 16. Security and threat-model update

The threat model MUST be updated before release because this migration materially changes the object-storage access pattern, deployment profile, backup/restore behavior, and migration workflow. Core 04 already requires threat-model updates before releases that change an object-storage access pattern or backup/restore mechanism.[^10]

| STRIDE class | Required SeaweedFS-specific coverage |
| --- | --- |
| Spoofing | S3 endpoint identity, reverse-proxy trust, credential source, direct upload target scope. |
| Tampering | Object overwrite, object delete, object metadata drift, migration copy mismatch, backup manifest mismatch. |
| Repudiation | Application-owned evidence attach audit remains authoritative; object-store logs are diagnostic only. |
| Information disclosure | S3 credentials, raw object keys, upload targets, preview/download handles, SeaweedFS admin/filer/master/volume UIs. |
| Denial of service | Oversized evidence, prefix listing abuse, storage exhaustion, range-read abuse, probe cleanup failure. |
| Elevation of privilege | Bucket wildcard credentials, anonymous access, exposed admin APIs, wildcard CORS, mistaken MinIO server default. |

Security acceptance for this migration is:

- no production release manifest exposes SeaweedFS admin, master, filer, volume, or WebDAV UI ports;
- no production CORS rule uses wildcard origin for direct upload;
- no retained harness, migration, backup, or restore artifact contains raw secret values;
- no public preview/download response contains bucket names, object keys, raw storage refs, backend URLs, or long-lived credentials;
- direct upload targets are bound to one pending blob slot and one target expiry.

## 17. Documentation and terminology patch plan

| Old term | Replacement | Rule |
| --- | --- | --- |
| `MinIO container` | `SeaweedFS S3 container` | Use only for default local/disconnected service. |
| `MinIO bucket` | `S3 bucket` | Use in generic docs and harness artifacts. |
| `minio_bucket` | `s3_bucket` | Required artifact field rename. |
| `minio_endpoint` | `object_store_endpoint` | Use one canonical field per artifact schema. |
| `MinIO readiness` | `object-store readiness` | Backend identified by `object_store_backend`. |
| `MinIO server dependency` | Forbidden unless describing legacy external endpoint. | MUST NOT describe `minio-go`. |
| `minio-go` | `Apache-2.0 S3-compatible Go client` | Allowed only in dependency inventory, SBOM, license notes, or adapter implementation notes. |

Required documentation patches:

- README: default services, setup, and migration note.
- Development guide: dependency inventory and local services.
- Bootstrap guide: Compose, `db-up`, and object-store fixture guidance.
- Implementation/testing guide: service-backed fixtures and Phase 0, Phase 5, and Phase 10 gates.
- Testing harness NLSpec: service lifecycle and artifact vocabulary.
- `.env.example`: remove MinIO-server-specific defaults unless clearly legacy external endpoint examples.
- Threat model: add SeaweedFS coverage from §16.
- Release notes: state MinIO server removed as default; `minio-go` retained as generic S3 client if present.

A repository search for `MinIO`, `minio`, `MINIO`, `minio_bucket`, and `minio_endpoint` after migration MUST classify every remaining occurrence as one of:

| Classification | Allowed content |
| --- | --- |
| `sdk_only` | `minio-go` dependency, license, SBOM, or adapter implementation note. |
| `legacy_external_endpoint` | Operator-supplied external S3-compatible endpoint caveat. |
| `migration_source` | Source backend in MinIO-to-SeaweedFS migration tooling or validation artifact. |
| `historical_changelog` | Release-note or changelog reference to prior default. |
| `invalid` | Any default service, fixture, runtime, readiness, bucket, or operator instruction reference. |

`invalid` occurrences block release.

## 18. Implementation sequence

| Phase | Gate to enter | Required work | Gate to exit |
| --- | --- | --- | --- |
| A. Contract cleanup | This revised plan accepted. | Patch owner docs, support docs, and error registry as required by §5. | Owner patches merged; SDK decision closed; no blocked owner dependency remains. |
| B. Local service replacement | Phase A complete. | Replace MinIO server with SeaweedFS S3 in local development and default service definitions. | `seaweedfs-s3` starts with pinned image and capability probe passes. |
| C. Harness replacement | Phase B complete. | Replace MinIO service vocabulary, artifacts, and service-backed fixture ownership. | Harness artifacts use canonical vocabulary and compatibility suite passes. |
| D. Adapter hardening | Phase C complete. | Implement adapter operations, capability probe, backend-neutral errors, and direct upload target probe. | Adapter tests and runtime outage mappings pass. |
| E. Backup and restore | Phase D complete. | Add SeaweedFS object manifest, restore verification, and backup/restore test coverage. | Fresh Postgres plus fresh SeaweedFS restore verification passes. |
| F. Migration tooling | Phase E complete. | Implement application-stopped migration utility, state machine, copy semantics, validation artifact, and rollback docs. | Fixture migration emits `result='pass'` and blocks mismatches. |
| G. Security and release gate | Phase F complete. | Update threat model, SBOM/license notes, release docs, and security scan. | Full release gate and acceptance criteria pass. |

No phase may be marked done by prose-only completion. Each exit gate requires retained evidence, command output, or repository diff evidence tied to the acceptance criteria in §19.

## 19. Acceptance criteria

| ID | Criterion |
| --- | --- |
| `SWFS-AC-001` | No default development, CI, service-backed test, or release Compose/manifest starts, pulls, or names a MinIO server container or image. |
| `SWFS-AC-002` | `github.com/minio/minio-go/v7`, if present, appears only as an Apache-2.0 generic S3 client dependency behind `internal/platform/objectstore`; no runtime service, fixture, readiness label, or operator instruction treats it as MinIO server support. |
| `SWFS-AC-003` | Core 01 object-storage wording names SeaweedFS S3 as the default local/disconnected S3-compatible target while preserving generic S3 compatibility. |
| `SWFS-AC-004` | Core 04 disconnected deployment wording names one SeaweedFS S3 container or equivalent S3-compatible object store. |
| `SWFS-AC-005` | The default local service is named `seaweedfs-s3`, uses a pinned SeaweedFS image tag plus digest, and exposes only the S3 endpoint in ordinary local development. |
| `SWFS-AC-006` | Production documentation forbids default exposure of SeaweedFS admin, master, filer, volume, WebDAV, and debug surfaces. |
| `SWFS-AC-007` | The capability probe completes PutObject, HeadObject, full GetObject, range GetObject, DeleteObject, and presigned PUT for the active direct-upload mode within the defined timeout and retry bounds. |
| `SWFS-AC-008` | In production profile, missing bucket, denied credentials, endpoint unreachable, CORS failure, or missing required capability fails startup before ready state. |
| `SWFS-AC-009` | After a post-ready object-store outage, ordinary non-evidence workbook row editing remains available while evidence upload, attach finalization, preview, and download fail through the mapped public dependency error behavior. |
| `SWFS-AC-010` | `POST /api/v1/object-blobs` still returns the Core-owned blob-slot response shape, including `upload_target`, `target_expires_at`, `pending_expires_at`, and `accepted_contract`; target expiry remains 60 minutes and pending slot expiry remains 24 hours. |
| `SWFS-AC-011` | A browser E2E flow creates a pending blob slot, uploads bytes to SeaweedFS through the active upload target, attaches the blob to an evidence record, receives a workbook projection row, and emits the expected collaboration update. |
| `SWFS-AC-012` | Preview and download issuance return only same-origin opaque evidence handles and never return bucket names, object keys, raw storage refs, raw SeaweedFS URLs, or long-lived object-store credentials. |
| `SWFS-AC-013` | Evidence negative cases for missing blob, pending blob, failed blob, quarantined blob, oversized preview, unsupported preview, expired handle, consumed handle, stale row version, and expired upload target produce the exact mapped public error code and reason code. |
| `SWFS-AC-014` | Harness artifacts use `object_store_backend`, `s3_bucket`, `s3_prefix`, `object_store_endpoint`, and `object_store_capability_probe`; they contain no `minio_bucket`, `minio_endpoint`, `minio_container`, or MinIO server readiness labels. |
| `SWFS-AC-015` | The SeaweedFS compatibility suite passes every required `SWFS-COMP-*` case and contains no conditional multipart or presigned-GET skip row. |
| `SWFS-AC-016` | Each successful backup set against SeaweedFS includes a `cartulary.object_store_backup_manifest.v1` manifest tied to the same `backup_set_id` and `consistency_point_at` as the Postgres restore artifact. |
| `SWFS-AC-017` | Restoring the latest successful retained backup into fresh Postgres and fresh SeaweedFS rebuilds projections, opens at least one incident when data exists, and preserves evidence/blob lifecycle consistency. |
| `SWFS-AC-018` | Default MinIO-to-SeaweedFS migration preserves bucket name and object keys and does not mutate database `storage_ref` values. |
| `SWFS-AC-019` | Migration validation emits `cartulary.object_store_migration_validation.v1` with `result='pass'` only when all blocking arrays are empty and every deterministic preview sample passes. |
| `SWFS-AC-020` | Any target-side object existing with a different size or hash than source blocks migration cutover. |
| `SWFS-AC-021` | The threat model update includes every STRIDE row listed in §16 and names SeaweedFS direct upload, credentials, admin surfaces, backup/restore, and migration validation. |
| `SWFS-AC-022` | Release SBOM and license gates identify no MinIO server artifact; if `minio-go` remains, release notes identify it as an Apache-2.0 S3-compatible client dependency only. |
| `SWFS-AC-023` | Default docs no longer describe MinIO server as the default local, disconnected, CI, service-backed test, or release-support object-store target. |
| `SWFS-AC-024` | Full release gate runs the SeaweedFS compatibility suite, Phase 0 object-store reachability, Phase 5 evidence paths, Phase 10 backup/restore paths, security scan, license/SBOM gate, and full `make check`. |
| `SWFS-AC-025` | The post-migration repository occurrence inventory classifies every remaining `MinIO`, `minio`, `MINIO`, `minio_bucket`, and `minio_endpoint` occurrence under §17, with zero `invalid` classifications. |

## 20. Source limits

This plan was revised from the uploaded migration plan and the uploaded Cartulary documentation set. It does not inspect the live repository, Compose files, Make targets, Go module files, lockfiles, current threat model, or current runtime code. Exact repository patches MUST inspect the target repository before claiming completion.

This plan does not independently revalidate external web facts about MinIO server, SeaweedFS, licenses, image tags, or upstream support posture. Release implementation MUST rely on live repository SBOM, license, and image-pin evidence rather than on this prose.

## Sources

[^1]: `00_document_set_status_and_precedence.md`, §1 Status, §2 Precedence, §3 Normative language, and adopted subsystem NLSpec boundary, lines 5-20, 22-47, and 96-113; `nlspec-spec.md`, “The Nature of the Artifact” and “Why NLSpecs Work When They Work,” lines 11-72.

[^2]: `01_architecture_storage_and_view_contracts.md`, §1 Architecture pattern and §2 Required modules and boundaries, lines 3-49.

[^3]: `cartulary-dev-guide.md`, §2.1 Backend runtime dependencies, lines 137-148.

[^4]: `01_architecture_storage_and_view_contracts.md`, §3.3.8 Evidence and blob routes, lines 3120-3167.

[^5]: `01_architecture_storage_and_view_contracts.md`, §16 Evidence-access handle contract, lines 5523-5577.

[^6]: `04_security_deployment_and_conformance.md`, §12.1 Scope and owner and §12.2 Canonical artifact and discovery, lines 1641-1653 and 1662-1677.

[^7]: `04_security_deployment_and_conformance.md`, §12.3 runtime-root binding rules, lines 1729-1750.

[^8]: `04_security_deployment_and_conformance.md`, `application.public_origin` key ownership, lines 1681-1686.

[^9]: `01_architecture_storage_and_view_contracts.md`, §12.1 Backup and §12.2 Restore, lines 5124-5174.

[^10]: `04_security_deployment_and_conformance.md`, §4.4 STRIDE threat model update triggers and matrix, lines 379-417.
