---
title: Cartulary MinIO-to-SeaweedFS S3 Migration Implementation Plan
document_class: nlspec-style closed implementation plan
status: implementation-plan
created_at: 2026-05-30
scope: Replace MinIO server usage with SeaweedFS S3-compatible object storage in Cartulary development, test, disconnected, and release-support surfaces while retaining minio-go as a generic Apache-2.0 S3 client.
---

## 0. Status, authority, and revision boundary

This artifact is an implementation plan written in NLSpec voice. It is not an adopted Cartulary subsystem NLSpec unless the repository authority process later adopts it.

Core 00 through Core 04 remain the implementation-conformance authority for current Cartulary product behavior. Core 05 remains claim-publication authority only. This plan follows the existing authority model and does not supersede owner documents. Adopted subsystem NLSpecs may define bounded implementation-conformance requirements only for their named subsystem and must state scope, non-goals, owner interactions, and deployment-configuration effects.

This plan owns the migration work required to replace default MinIO server usage with SeaweedFS S3 in local development, service-backed tests, disconnected deployment support, release support, documentation, and migration tooling. It must not add public route families, alter public request or response envelopes, alter evidence lifecycle semantics, alter session or authorization semantics, or widen the deployment-configuration schema without a patch to the primary owner document.

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative inside this plan. **MUST** and **MUST NOT** define requirements. **SHOULD** and **SHOULD NOT** define strong defaults whose exceptions must remain compatible with all MUST-level requirements. **MAY** defines optional behavior only where omitted behavior, allowed profiles, and conformance effect are explicit.

Implementation is blocked when this plan requires an owner-document patch and that patch has not landed. A repository change that merely updates support docs, scripts, fixtures, examples, or comments is not sufficient when the behavior is owned by Core 00 through Core 04.

### 0.1 Closure classes and claimability

Every plan row that can affect implementation, release, or conformance claimability MUST use exactly one closure class.

| Closure class | Meaning | Claimable by default | Required evidence to become claimable |
| --- | --- | ---: | --- |
| `plan_local_closed` | The behavior is fully defined by this plan without changing Core-owned product behavior. | `true` | Retained implementation evidence named by the relevant acceptance criterion. |
| `blocked_until_owner_patch` | The behavior changes or extends a Core-owned product, security, deployment, route, error, health, or configuration contract. | `false` | The named owner patch is merged and cited by path, anchor, and requirement or acceptance ID. |
| `repo_or_external_fact_required` | The behavior depends on live repository files, image digests, SBOM/license evidence, exact command surfaces, lockfiles, or external source verification not present in this artifact. | `false` | The named fact is resolved into a retained artifact or repository-controlled file and referenced by the acceptance-evidence matrix. |

A section, phase, or acceptance criterion MUST NOT be marked complete if it depends on any unresolved `blocked_until_owner_patch` or `repo_or_external_fact_required` row.

Any row with `claimable=false` MUST include a `blocking_id`, `blocking_owner` or `required_evidence`, and `resolution_rule`. Missing blocker metadata is itself a plan defect.

### 0.2 Claimability record shape

When a tool or reviewer records claimability for this plan, the record MUST use this shape.

| Field | Type | Required rule |
| --- | --- | --- |
| `claimability_schema_id` | string | Exact `cartulary.seaweedfs_migration_claimability.v1`. |
| `subject_id` | string | Section, requirement, phase, or acceptance criterion ID. |
| `closure_class` | enum | `plan_local_closed`, `blocked_until_owner_patch`, or `repo_or_external_fact_required`. |
| `claimable` | boolean | `false` unless all blockers for the subject are resolved. |
| `blocking_id` | string or null | Required when `claimable=false`; null only when claimable. |
| `blocking_owner` | string or null | Required for owner-patch blockers. |
| `required_evidence` | string or null | Required for repo or external fact blockers. |
| `resolution_rule` | string | Exact condition that changes the row to claimable. |
| `resolved_by` | string or null | Repository path, artifact ID, or patch ID; null while unresolved. |

## 1. Executive decision

Default Cartulary development, CI, service-backed test, disconnected, and release-support object storage MUST use **SeaweedFS S3** instead of a MinIO server.

This is a deployment, harness, documentation, compatibility, backup/restore, and migration-substitution change. It is not a product-model change. Cartulary remains a modular monolith with one application deployable, one Postgres service as the authoritative structured data store, and one S3-compatible object storage service as the authoritative binary evidence store.

The migration MUST preserve all of the following invariants:

- one application deployable;
- one Postgres service;
- one S3-compatible object-storage service;
- unchanged public `/api/v1/*` and `/ws/v1/*` route families;
- unchanged blob-slot create and attach semantics;
- unchanged evidence record lifecycle semantics;
- unchanged same-origin preview and download handle semantics;
- unchanged server-side authorization as the primary incident-data access boundary.

The MinIO server and `github.com/minio/minio-go/v7` are distinct migration subjects. The MinIO server MUST NOT be shipped, started, named, or documented as the default Cartulary object-store service. The `minio-go` package MAY remain only as a generic S3-compatible client library behind the internal object-store adapter. If absent, no `minio-go` dependency is required. If `minio-go` is omitted, another S3-compatible client may satisfy this plan only by preserving the adapter contract. If `minio-go` is retained, SDK-only dependency and license claimability remain blocked until `SWFS-AC-002` and `SWFS-AC-022` pass.

## 2. Terminology

| Term | Required meaning |
| --- | --- |
| `MinIO server` | The MinIO object-storage server process, image, binary, container, service, fixture, or default runtime target. It is not a default Cartulary runtime dependency after this migration. |
| `minio-go` | The Go package `github.com/minio/minio-go/v7`, retained only as a generic S3-compatible client behind `internal/platform/objectstore` or equivalent. |
| `SeaweedFS S3` | The SeaweedFS S3-compatible endpoint used as Cartulary's default local, disconnected, CI, and service-backed test object-store target. |
| `object-store adapter` | The internal platform boundary that owns S3-compatible object operations. Evidence, workbook, reporting, backup, migration, and collaboration modules MUST NOT import SeaweedFS-specific code. |
| `direct upload target` | A short-lived upload target returned by `POST /api/v1/object-blobs` for exactly one pending blob slot. |
| `same-origin evidence handle` | An opaque application URL under `GET /api/v1/evidence-handles/{handle_token}` used for preview and download redemption. |
| `raw object key` | Any storage key, bucket-relative object name, filer path, volume identifier, backend path, or equivalent object-store identifier used to locate object bytes. |
| `storage ref` | The Core-owned authoritative object-location reference stored for an object blob. This plan does not redefine its current owner contract. |
| `legacy external S3 endpoint` | An operator-supplied S3-compatible endpoint that Cartulary does not ship, start, manage, test as the default target, or document as the default object-store service. |
| `runtime profile` | One of the migration runtime profiles defined in §6. It is not a Core conformance profile unless expressly named as such. |
| `operator-private artifact` | A retained artifact that may contain deployment-local object-store identifiers required for migration, backup, or restore. It MUST remain outside incident portability bundles and public user-facing responses. |
| `shareable summary` | A redacted human or machine summary that may be attached to release notes, tickets, or implementation reports. It MUST NOT contain secrets, credentials, raw object keys, backend URLs, or restore-only object locations. |
| `redaction object` | A structured placeholder that preserves correlation without exposing raw secret, bucket, key, or endpoint material. Its shape is defined in §2.1. |
| `claimable` | A row is claimable only when its closure class permits completion and all named blockers have been resolved. |

### 2.1 Redaction object shapes

Artifacts that need to refer to sensitive or deployment-local values without exposing them MUST use the following redaction objects.

| Redaction class | Shape | Required hashing input |
| --- | --- | --- |
| `secret_ref` | `{ "redacted": true, "redaction_class": "secret", "sha256": <lowercase-hex-64> }` | UTF-8 bytes of the exact secret value. |
| `bucket_ref` | `{ "redacted": true, "redaction_class": "bucket", "sha256": <lowercase-hex-64> }` | UTF-8 bytes of the exact bucket name. |
| `object_key_ref` | `{ "redacted": true, "redaction_class": "object_key", "sha256": <lowercase-hex-64> }` | UTF-8 bytes of the exact bucket-relative key. |
| `storage_ref` | `{ "redacted": true, "redaction_class": "storage_ref", "sha256": <lowercase-hex-64> }` | UTF-8 bytes of the exact storage reference string. |
| `endpoint_ref` | `{ "redacted": true, "redaction_class": "endpoint", "scheme": <scheme-or-null>, "host_sha256": <lowercase-hex-64>, "port_present": <boolean> }` | UTF-8 bytes of the normalized endpoint host only for `host_sha256`; scheme may be retained because it is not credential-bearing. |

`sha256` values MUST be lowercase 64-character hexadecimal strings. Redaction objects MUST NOT include raw bucket names, raw keys, credentials, endpoint hosts, query strings, or userinfo.

## 3. Non-goals and preserved invariants

### 3.1 Non-goals

This migration MUST NOT introduce any behavior in the following table.

| Non-goal | Boundary |
| --- | --- |
| New user-visible storage abstraction | Users and browser clients MUST NOT choose, identify, or reason about SeaweedFS, MinIO, AWS S3, or another backend while using ordinary workbook and evidence routes. |
| SeaweedFS-specific public routes | No public `/api/v1/*` or `/ws/v1/*` route family may be added solely for SeaweedFS. |
| Product-model change | Evidence records, object blobs, workbook rows, projections, revisions, and collaboration messages keep their current owner semantics. |
| Object-key authority | Raw object keys, bucket names, filer paths, volume IDs, object-store URLs, and storage backend identifiers MUST NOT become incident-portability source data or public row identity. |
| Bucket policy as primary authorization | Application authorization remains the primary incident-data authorization boundary. Object-store permissions are defense-in-depth and service isolation. |
| Multipart upload | Multipart upload is out of scope for this migration. Any later multipart support requires an owner document that defines thresholds, chunk ordering, abort semantics, retry behavior, finalization, cleanup, and public error behavior. |
| SeaweedFS administrative dependency | Runtime product code MUST NOT call SeaweedFS master, filer, volume, or admin APIs for ordinary application behavior. |
| New deployment-config namespace | This plan MUST NOT add a SeaweedFS deployment-configuration namespace. Any new application config keys require an owner-document patch. |
| Live migration workflow | This plan does not add live maintenance-mode product behavior. The default migration mode is application-stopped migration. |
| Presigned GET evidence access | Public preview and download MUST remain application-mediated through same-origin evidence handles. |
| Object-store listing as a user primitive | No product route or workbook surface may expose S3 bucket or prefix listing as an analyst feature. |

### 3.2 Preserved public evidence boundary

`POST /api/v1/object-blobs` remains the only public blob-slot creation route. A successful response continues to include `object_blob_id`, `upload_target`, `target_expires_at`, `pending_expires_at`, and `accepted_contract`; upload target expiry remains 60 minutes and pending slot expiry remains 24 hours under the Core-owned contract.

Preview and download MUST continue to be issued only through the existing preview-handle and download-handle routes and redeemed only through `GET /api/v1/evidence-handles/{handle_token}`. Those handles remain opaque, same-origin, short-lived, authorization-checked, and application-mediated. Public preview and download responses MUST NOT expose long-lived object-store credentials, bucket names, raw object keys, raw SeaweedFS URLs, or storage-backend-specific identifiers.

The only permitted backend-upload exposure is the `upload_target` returned by `POST /api/v1/object-blobs` when the active upload mode is `direct_presigned_put`. That exception is limited to uploading bytes for one pending blob slot. It MUST NOT be used for preview, download, listing, delete, evidence discovery, or cross-incident transfer.

## 4. Closed decision registry

| Decision ID | Decision | Required effect | Closure class |
| --- | --- | --- | --- |
| `SWFS-RD-001` | SeaweedFS S3 is the default local, disconnected, CI, service-backed test, and release-support object-store target. | Default Compose, harness, fixture, documentation, and release-support paths MUST use SeaweedFS S3 or generic S3-compatible wording. | `plan_local_closed` |
| `SWFS-RD-002` | MinIO server is no longer a supported default target. | MinIO server MAY appear only as a legacy external S3 endpoint configured and operated by the deployment owner or as a migration source label. If omitted, no MinIO server path exists. | `plan_local_closed` |
| `SWFS-RD-003` | `github.com/minio/minio-go/v7` may remain as a generic S3 client. | `minio-go` MUST remain behind `internal/platform/objectstore` or equivalent and MUST NOT justify MinIO server fixtures, docs, services, admin APIs, or default runtime behavior. | `plan_local_closed` |
| `SWFS-RD-004` | Public Cartulary API behavior is unchanged except for owner-approved additive error-code rows. | No SeaweedFS-specific public route, public storage abstraction, or browser-visible backend identity may be added. | `blocked_until_owner_patch` for new public errors; otherwise `plan_local_closed` |
| `SWFS-RD-005` | No new deployment-configuration namespace is introduced by this plan. | Any new application configuration key is blocked until the primary owner defines keys, defaults, null behavior, validation, redaction, and startup failure semantics. | `blocked_until_owner_patch` |
| `SWFS-RD-006` | The default SeaweedFS upload mode is `direct_presigned_put`. | Startup readiness MUST require presigned PUT compatibility for SeaweedFS profiles unless an owner document later defines an application-mediated upload profile. | `plan_local_closed` |
| `SWFS-RD-007` | Default migration preserves source bucket name and object keys. | Any bucket or key-shape change is a separate high-risk migration requiring explicit mapping, database migration, validation, and rollback rules. | `plan_local_closed` |
| `SWFS-RD-008` | Multipart upload is out of scope. | Compatibility, harness, backup, and migration tests MUST NOT treat multipart as required or skipped current-profile behavior. | `plan_local_closed` |
| `SWFS-RD-009` | Presigned GET is not part of public evidence access. | Public preview/download tests MUST use same-origin evidence handles, not presigned GET. | `plan_local_closed` |
| `SWFS-RD-010` | Runtime object-store dependency outages need explicit public error ownership. | The owner patch in §5.1 MUST land before route-local runtime outage behavior is claimed complete. | `blocked_until_owner_patch` |
| `SWFS-RD-011` | Backup and restore artifacts must be sufficient without prose. | Restore input MUST use an operator-private manifest that includes all object locations and computed SHA-256 values. | `plan_local_closed` |
| `SWFS-RD-012` | Migration equivalence is byte equivalence. | ETag MAY be retained as diagnostics but MUST NOT be sufficient to prove source-target object equality. If omitted, SHA-256 remains the only equivalence proof. | `plan_local_closed` |

## 5. Owner-document patch registry

This registry replaces prose-only owner-patch references. A row is not claimable until the target document is patched at the named anchor or exact replacement point.

| Owner patch ID | Target document | Target anchor | Patch class | Required change | Plan sections blocked | Claimability effect | Evidence required |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `SWFS-OWNER-ERR-001` | `01_architecture_storage_and_view_contracts.md` | Public success/error envelope and error-code registry | `normative_core` | Add public error codes `object_store_unavailable` and `object_store_access_rejected`, transport statuses, retry hints, reason-code vocabularies, and route-family applicability. | §10.1, §10.2, §10.4, `SWFS-AC-009`, `SWFS-AC-013` | `blocks_completion` | Owner diff plus generated error contract artifact. |
| `SWFS-OWNER-CONFIG-001` | `04_security_deployment_and_conformance.md` | Deployment-configuration contract | `normative_core` | Define exact backend-neutral object-store configuration keys, types, defaults, omitted behavior, explicit-`null` behavior, environment overlays, validation errors, redaction, cross-key rules, and startup failure semantics. | §6.4, §6.5, `SWFS-AC-008`, `SWFS-AC-024` | `blocks_release` | Owner diff plus config-schema drift artifact. |
| `SWFS-OWNER-HEALTH-001` | `01_architecture_storage_and_view_contracts.md` or `04_security_deployment_and_conformance.md` | Health/readiness owner section | `normative_core` | Define public readiness output for `degraded_object_store`, recovery events, and whether readiness is boolean, degraded, or structured. | §9.6, §10.3, `SWFS-AC-008`, `SWFS-AC-009` | `blocks_completion` | Owner diff plus health-route or readiness artifact evidence. |
| `SWFS-OWNER-STORAGEREF-001` | `01_architecture_storage_and_view_contracts.md` and any generated contracts | Object blob storage-ref owner | `normative_core` | Core 01 defines server-managed logical `object://{object_blob_uuid}` refs, reserves that form from user-authored direct mutations, defines `object_blob_storage_key_v1`, caps physical keys at 1024 UTF-8 bytes, and requires invalid server-generated keys to fail before backend calls. | §7.3, §8.2, §13.4, `SWFS-AC-018` | `resolved_by_owner_contract` | Owner diff plus storage-ref owner coverage artifact. |
| `SWFS-OWNER-RANGE-001` | `01_architecture_storage_and_view_contracts.md` | Evidence preview/download owner section | `normative_core` | Define whether range retrieval is required for preview only, download only, or both; define full-download fallback when range is unavailable after readiness. | §10.4, `SWFS-AC-007`, `SWFS-AC-013` | `blocks_completion` | Owner diff plus browser/backend evidence-access tests. |
| `SWFS-OWNER-BACKUP-001` | `01_architecture_storage_and_view_contracts.md` and `04_security_deployment_and_conformance.md` | Backup/restore owner sections | `normative_core` | Confirm that the operator-private object-store manifest and restore-verification schemas in §12 are adopted as backup/restore evidence for this migration. | §12, `SWFS-AC-016`, `SWFS-AC-017` | `blocks_release` | Owner diff plus backup/restore retained artifacts. |
| `SWFS-OWNER-THREAT-001` | Project threat model path | STRIDE object-storage and backup/restore sections | `security_model` | Update STRIDE coverage for SeaweedFS endpoint identity, direct upload, credential scope, admin surfaces, CORS, backup/restore, and migration validation. | §15, `SWFS-AC-021` | `blocks_release` | Threat-model diff or scanner-ready retained document. |
| `SWFS-OWNER-DOCS-001` | `cartulary-dev-guide.md`, `cartulary_repository_bootstrap_guide.md`, and any repo docs index | Object-store service defaults | `implementation_support` | Replace MinIO server default wording with SeaweedFS S3 or generic S3-compatible wording while preserving `minio-go` as SDK-only if present. | §16, §17, `SWFS-AC-001`, `SWFS-AC-002`, `SWFS-AC-023`, `SWFS-AC-025` | `blocks_docs` | Occurrence inventory plus docs diffs. |

## 6. Configuration, runtime profiles, and exposure

### 6.1 Runtime profile registry

The runtime profile registry is closed to the five profiles in this table. No other profile may claim this plan unless the table is revised.

| Field | `local_dev` | `ci_service_backed` | `disconnected_prod` | `on_prem_or_cloud_external` | `developer_debug` |
| --- | --- | --- | --- | --- | --- |
| `closure_class` | `repo_or_external_fact_required` until repo service files are inspected | `repo_or_external_fact_required` until harness files are inspected | `blocked_until_owner_patch` for config keys; repo fact required for deployment manifests | `blocked_until_owner_patch` for config keys | `repo_or_external_fact_required` |
| `production_allowed` | `false` | `false` | `true` | `true` | `false` |
| `service_name` | `seaweedfs-s3` | `seaweedfs-s3` | `seaweedfs-s3` unless operator-managed name is supplied by deployment manifest | `external_operator_managed` | `seaweedfs-s3` |
| `image_ref` | `TODO:seaweedfs-image-tag-and-digest-required` | `TODO:seaweedfs-image-tag-and-digest-required` | `TODO:seaweedfs-image-tag-and-digest-required` when Cartulary ships service manifest | `operator_owned` | Same as selected base profile |
| `entrypoint` / `command` | `TODO:repo-compose-inspection-required` | `TODO:harness-service-inspection-required` | `TODO:release-manifest-inspection-required` | `operator_owned` | Same as selected base profile |
| `s3_endpoint_internal` | Repo-defined service DNS or loopback endpoint from owner-defined object-store config | Harness endpoint from service lease artifact | Owner-defined object-store config | Owner-defined object-store config | Same as selected base profile |
| `s3_endpoint_browser_reachability` | `loopback_direct` by default | `harness_direct` | `same_origin_reverse_proxy` by default; `public_direct_operator_managed` only when owner deployment docs adopt it | `public_direct_operator_managed` or `same_origin_reverse_proxy` as configured by owner docs | `loopback_direct` only |
| `host_bindings` | S3 endpoint may bind loopback only; admin surfaces container-network-only | Harness network only | App public surface only; S3 browser reachability only through declared mode | Operator-owned | Loopback only; forbidden in release gates |
| `admin_surface_exposure` | Master, filer, volume, WebDAV, metrics, and debug surfaces not host-exposed | Not host-exposed except harness diagnostics | Forbidden unless a separate owner document defines authenticated exposure | Operator-owned but not controlled by Cartulary conformance | Loopback only and release-forbidden |
| `credential_source` | Dev-only local secret file ignored by VCS or repo-defined dev secret mechanism | Harness-generated secret for one run | Deployment-local secret procedure | Operator-managed secret procedure | Same as selected base profile |
| `bucket_bootstrap` | Application or dev tooling may create bucket | Harness may create run bucket or prefix | Application startup MUST NOT create bucket | Operator-owned | Same as selected base profile |
| `persistence_paths` | Named Docker volume or local root; exact path is repo fact | Ephemeral per run unless diagnostic retention is enabled | §6.5 under `roots.object_storage.path` | Managed service; not filesystem-root unless configured | Same as selected base profile |
| `cors_policy_id` | `cors_direct_put_v1` | `cors_direct_put_v1` | `cors_direct_put_v1` | `cors_direct_put_v1` when direct browser upload is enabled | `cors_direct_put_v1` |
| `readiness_policy_id` | `probe_required_direct_put_v1` | `probe_required_direct_put_v1` | `probe_required_direct_put_v1` | `probe_required_direct_put_v1` | `probe_required_direct_put_v1` |

The SeaweedFS image reference MUST be pinned by tag and digest in repo-control files before any release-support profile is claimable. `latest`, floating major tags, and unpinned image references are invalid for default development, CI, disconnected, and release-support profiles.

Readiness MUST be determined by the object-store capability probe in §9, not by a container-started flag alone.

### 6.2 Direct-upload exposure matrix

The default active upload mode for SeaweedFS profiles is `direct_presigned_put`.

| Profile | Default upload mode | Browser reachability | Fallback allowed | Required result when unreachable |
| --- | --- | --- | --- | --- |
| `local_dev` | `direct_presigned_put` | `loopback_direct` to the S3 endpoint or a repo-defined equivalent that resolves to the same endpoint | `false` | Startup probe fails until endpoint and CORS are configured. |
| `ci_service_backed` | `direct_presigned_put` | `harness_direct` inside harness-owned browser/service network | `false` | Product tests MUST NOT start. |
| `disconnected_prod` | `direct_presigned_put` | `same_origin_reverse_proxy` by default | `false` unless an owner patch defines application-mediated upload | Startup fails closed. |
| `on_prem_or_cloud_external` | `direct_presigned_put` | `public_direct_operator_managed` or `same_origin_reverse_proxy` as selected by owner-defined deployment configuration | `false` | Startup fails closed. |
| `developer_debug` | Inherits selected base profile | `loopback_direct` only | Inherits selected base profile | Forbidden in release gates. |

If a deployment cannot make the direct upload target reachable from the browser under this table, it MUST fail startup. It MUST NOT silently fall back to application-mediated upload unless a later owner patch defines that upload mode, route behavior, probe behavior, and public response differences.

For `local_dev`, the repo-defined equivalent endpoint is the loopback CORS-normalizing proxy started by `scripts/dev-services.sh` from `tools/s3corsproxy`. SeaweedFS S3 remains the backing service and is bound to a loopback-only upstream port; the browser-facing object-store endpoint is the proxy on the ordinary local object-store port. The proxy MUST preserve signed S3 request host semantics and MUST normalize only browser CORS response behavior to `cors_direct_put_v1`.

### 6.3 Direct-upload CORS policy

When `upload_mode='direct_presigned_put'`, the S3 endpoint or same-origin reverse proxy used for direct PUT MUST satisfy `cors_direct_put_v1`.

| CORS field | Required value |
| --- | --- |
| Allowed origins | Exactly `application.public_origin`. Wildcard origins are forbidden in every claimable profile. |
| `Origin: null` | Rejected. |
| Missing `Origin` | Request may be evaluated as non-browser traffic, but no CORS response headers are emitted. |
| Methods | `PUT`, `OPTIONS`. |
| Disallowed browser methods | `GET`, `DELETE`, `POST`, bucket listing, and non-S3 protocols. |
| Allowed request headers | Exactly `content-type` and `x-amz-checksum-sha256`, case-insensitive. If the generated upload target does not bind a checksum header, a browser MAY omit `x-amz-checksum-sha256`; the CORS allowlist remains unchanged. |
| Forbidden request headers | `authorization`, `cookie`, `x-amz-security-token`, `x-amz-meta-*`, and any header outside the allowed request-header set. |
| Exposed response headers | `etag` only. |
| Credentials | `Access-Control-Allow-Credentials=false`; browser credentials to the object-store endpoint are forbidden. |
| Max age | `600` seconds. |
| Failure before readiness | Probe fails with `object_store_access_rejected` and `reason_code=cors_rejected`. |
| Failure after readiness | Evidence upload or attach flow fails through owner-approved runtime dependency error mapping. |

The direct-upload target generation algorithm MUST NOT require browser-supplied headers outside the allowed request-header set. Any future requirement for additional signed headers is blocked until this table and the Core-owned accepted upload contract are revised together.

### 6.4 Configuration and secret registry

This migration MUST NOT add new Cartulary deployment-configuration keys. Runtime application configuration MUST continue to use owner-defined deployment configuration and owner-defined object-store configuration. Core 04 owns the deployment-configuration container, discovery, overlay, unknown-key rejection, validation errors, and fail-closed startup behavior.

Known owner-defined configuration keys that this plan may reference are listed below.

| Semantic use | Config key | Type | Required profiles | Default | Omitted behavior | Explicit `null` behavior | Redaction class | Startup effect | Closure class |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Configuration schema identity | `config_schema_id` | string | all | none | invalid | invalid | `non_secret` | `fail_closed` | `plan_local_closed` |
| Deployment profile | `deployment_profile` | enum | all | none | invalid | invalid | `non_secret` | `fail_closed` | `plan_local_closed` |
| Browser origin | `application.public_origin` | absolute `http` or `https` origin | all direct upload profiles | none | invalid | invalid | `endpoint` | `fail_closed` | `plan_local_closed` |
| Object-storage runtime root | `roots.object_storage` | runtime-root object | `disconnected_prod` | none | invalid in disconnected profile | invalid | `identifier` | `fail_closed` | `plan_local_closed` |
| Backup runtime root | `roots.backup_storage` | runtime-root object | backup/restore profiles | none | invalid when backup/restore gate runs | invalid | `identifier` | `fail_closed` | `plan_local_closed` |
| Object-store endpoint | `TODO:owner-object-store-endpoint-key` | TODO | all profiles | TODO | TODO | TODO | `endpoint` | `fail_closed` | `blocked_until_owner_patch` |
| Object-store bucket | `TODO:owner-object-store-bucket-key` | TODO | all profiles | TODO | TODO | TODO | `identifier` | `fail_closed` | `blocked_until_owner_patch` |
| Object-store access-key reference | `TODO:owner-object-store-access-key-ref` | TODO | all profiles | TODO | TODO | TODO | `secret` | `fail_closed` | `blocked_until_owner_patch` |
| Object-store secret-key reference | `TODO:owner-object-store-secret-key-ref` | TODO | all profiles | TODO | TODO | TODO | `secret` | `fail_closed` | `blocked_until_owner_patch` |
| Object-store upload mode | `TODO:owner-object-store-upload-mode-key` if configurable | enum | only when operator-selectable | `direct_presigned_put` when not configurable | Use default | explicit `null` invalid if key exists | `non_secret` | `fail_closed` | `blocked_until_owner_patch` |

Compose-only variables used solely to configure the SeaweedFS container are `service_local_variable` values. They MUST NOT be documented as Cartulary deployment-configuration keys. They MUST NOT participate in Core 04 deployment-configuration validation unless an owner patch adopts them as application configuration.

Secret values MUST be provided through deployment-local secret handling. Raw access keys, secret keys, session tokens, presigned URLs, CORS secrets, and service-local credential files MUST NOT appear in incident records, workbook rows, public route responses other than the short-lived `upload_target` exception, portability bundles, retained harness summaries, shareable migration summaries, screenshots, logs, or release notes.

### 6.5 Disconnected persistence layout

When `deployment_profile='disconnected'`, `roots.object_storage` MUST use `binding_kind='filesystem_root'` under the Core-owned runtime-root contract. SeaweedFS state for the default disconnected profile MUST be rooted under that filesystem root as follows.

| Path under `roots.object_storage.path` | Required contents | Backup criticality | Secret-bearing | Restore rule |
| --- | --- | --- | --- | --- |
| `seaweedfs/master/` | Master or equivalent topology metadata. | Required. | No. | Restore before readiness probe. |
| `seaweedfs/filer/` | Filer or equivalent namespace metadata. | Required. | No. | Restore before readiness probe. |
| `seaweedfs/volume/` | Object bytes and volume files. | Required. | Contains incident evidence bytes but not credentials by design. | Restore before readiness probe. |
| `seaweedfs/tmp/` | Runtime scratch only. | Not authoritative. | No. | MAY be omitted on backup and recreated on startup. Omitted backup behavior means no restore attempt is made for this path. |
| outside `roots.object_storage.path` | S3 access credentials and service-local secrets. | Not part of evidence backup. | Yes. | Restored only through deployment-local secret procedure. |

The application MUST NOT treat SeaweedFS metadata paths, volume files, or object keys as incident-portability source data. Whole-incident portability remains governed by the incident-portability owner documents, not by SeaweedFS filesystem layout.

## 7. Credential, bucket, key, and exposure policy

### 7.1 Credential identity registry

| Environment | Identity | Secret source | Required scope | Retention | Closure class |
| --- | --- | --- | --- | --- | --- |
| `local_dev` | `cartulary_dev` | Dev-only local secret file ignored by VCS or repo-defined dev secret mechanism | Bucket `cartulary` or repo-defined dev bucket | May persist only in local dev secret storage | `repo_or_external_fact_required` |
| `ci_service_backed` | `cartulary_test_<run_segment>` | Harness CSPRNG secret material | Run-owned bucket or run-owned prefix | Destroy at end of run unless diagnostics explicitly retain redacted evidence | `repo_or_external_fact_required` |
| `disconnected_prod` | `cartulary_app` | Deployment-local secret procedure | One production bucket | Deployment-local rotation and backup procedure | `blocked_until_owner_patch` for exact config keys |
| `migration` | `cartulary_migration_<run_segment>` | Operator or migration tool CSPRNG secret material | Source and target buckets for one migration run | Revoke after `post_cutover_verified` or `rolled_back` | `plan_local_closed` |
| `backup_restore` | `cartulary_backup_restore` | Deployment-local operator secret procedure | Backup, restore, list, and verification scope only | Deployment-local rotation and audit procedure | `blocked_until_owner_patch` if owner backup profile requires exact credential naming |

Credentials generated by Cartulary-controlled tooling MUST satisfy these bounds:

| Secret | Minimum entropy | Encoding requirement | Artifact rule |
| --- | ---: | --- | --- |
| Access key | 128 bits | Printable ASCII or URL-safe Base64 without whitespace | Raw value forbidden outside secret storage. |
| Secret key | 256 bits | Printable ASCII or URL-safe Base64 without whitespace | Raw value forbidden outside secret storage. |
| Migration run segment source | 128 bits if random; otherwise hash rule in §7.4 | Lowercase hex where represented | Redacted in shareable summaries. |

### 7.2 Logical action to S3-compatible operation mapping

| Logical action | Concrete S3-compatible operation | Allowed profiles | Forbidden uses |
| --- | --- | --- | --- |
| `read_object` | `GetObject` | all active profiles where evidence access is enabled | User-facing direct object-store URLs and presigned GET. |
| `read_object_range` | ranged `GetObject` or equivalent Range request | profiles where preview/download owner contract requires range | Treating range as optional when startup profile declares it required. |
| `head_object` | `HeadObject` or equivalent object metadata read | all active profiles | Metadata reads that expose raw keys publicly. |
| `write_object` | `PutObject` | upload target, app-mediated upload if later defined, probe, test, migration, backup restore | User-facing arbitrary object writes. |
| `delete_object` | `DeleteObject` | probe cleanup, test cleanup, authorized cleanup, migration validation cleanup | Ordinary user-facing query paths and broad production prefix deletes. |
| `list_bucket_prefix` | `ListObjectsV2` or equivalent scoped prefix listing | backup, diagnostics, restore verification, migration validation, test cleanup | User-facing workbook routes and analyst-visible object discovery. |
| `create_bucket` | bucket creation operation | `local_dev`, `ci_service_backed` only | `disconnected_prod`, `on_prem_or_cloud_external` app startup. |
| `admin_operation` | SeaweedFS master, filer, volume, admin, policy, user, lifecycle, or non-S3 protocol operation | none for runtime app code | all product runtime profiles. |

Bucket credentials MUST NOT grant `admin_operation` to runtime product code. A credential that can perform SeaweedFS admin, filer, master, volume, WebDAV, or policy-management operations is not a valid application runtime credential under this plan.

### 7.3 Bucket and key grammar

| Namespace | Required rule | Invalid behavior |
| --- | --- | --- |
| Cartulary-managed dev/test bucket names | MUST match `^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`; MUST NOT contain `..`, `.-`, `-.`, uppercase letters, underscores, path separators, leading/trailing hyphens, leading/trailing dots, or IPv4-address-like form. | Reject before backend call with `object_store_invalid_request`. |
| Production bucket names | No application default unless owner-defined. Missing production bucket value fails startup. | Startup fails closed. |
| New server-generated object keys | MUST follow Core 01 `object_blob_storage_key_v1`: `incidents/{incident_uuid}/object-blobs/{object_blob_uuid}`, lowercase RFC 4122 UUIDs, slash-separated ASCII segments, no empty/absolute/traversal/control segments, and max 1024 UTF-8 bytes. | Reject before backend call with `object_store_invalid_request`. |
| Migration source object keys | Opaque existing S3 keys. The migration MUST copy bytes without Unicode normalization, slash cleanup, dot-segment rewriting, case-folding, URL decoding, or path separator conversion. | If target backend cannot represent the key exactly, preflight fails. |
| Probe keys | Prefix `.cartulary/probes/startup/<probe_id>/` plus fixed filenames in §9.1. | Any probe outside prefix fails cleanup classification. |
| CI shared-bucket prefixes | `test-runs/<run_segment>/<scope_segment>/`. | Harness setup fails before product tests. |

Default migration MUST preserve source bucket names and object keys. A bucket/key rewrite migration is out of scope unless a separate mapping contract defines source, target, database migration, validation, and rollback.

### 7.4 Run prefix hashing

`run_segment` MUST be the first 32 lowercase hexadecimal characters of SHA-256 over the harness run ID bytes.

`scope_segment` MUST be the first 32 lowercase hexadecimal characters of SHA-256 over the package, suite, or scheduler-scope string. Empty scope string normalizes to `default` before hashing.

The byte encoding for the hash input MUST be UTF-8 without BOM. The output MUST be lowercase hexadecimal.

### 7.5 Exposure matrix

| Surface | `local_dev` | `ci_service_backed` | `disconnected_prod` | `on_prem_or_cloud_external` | `developer_debug` |
| --- | --- | --- | --- | --- | --- |
| Cartulary app HTTP/HTTPS | Exposed as configured. | Harness-owned only. | Exposed as the public app surface. | Operator-owned public app surface. | Same as base profile. |
| SeaweedFS S3 endpoint | Loopback direct or same-origin reverse proxy. | Harness-owned only. | Same-origin reverse proxy by default. | Operator-managed direct or reverse-proxy surface. | Loopback only. |
| SeaweedFS master UI/API | Container network only. | Container network only. | Forbidden. | Out of Cartulary conformance; docs must warn it is not an app dependency. | Loopback only. |
| SeaweedFS filer UI/API | Container network only. | Container network only. | Forbidden. | Out of Cartulary conformance; docs must warn it is not an app dependency. | Loopback only. |
| SeaweedFS volume UI/API | Container network only. | Container network only. | Forbidden. | Out of Cartulary conformance; docs must warn it is not an app dependency. | Loopback only. |
| WebDAV or non-S3 protocols | Disabled. | Disabled. | Forbidden. | Not used by Cartulary runtime. | May be enabled only outside release manifests. |
| Metrics or debug endpoints | Disabled unless owned by another adopted subsystem. | Harness-owned if present. | Forbidden unless owner document defines exposure and authentication. | Operator-owned. | Loopback only. |

## 8. Object-store adapter contract

### 8.1 Adapter boundary

All object-store access MUST pass through `internal/platform/objectstore` or an equivalent internal object-store adapter boundary. Product modules MUST NOT import SeaweedFS client code, SeaweedFS admin code, MinIO server code, or bucket-management APIs directly.

The adapter MAY use `github.com/minio/minio-go/v7` as its implementation client only as an SDK-only dependency. Omitted use means the adapter uses another S3-compatible client. No acceptance criterion may require `minio-go` specifically, and replacing the client MUST NOT change public API behavior, artifact schemas, retry behavior, error mapping, or migration validation.

### 8.2 Adapter input contracts

| Input | Type and bounds | Omission behavior | Explicit `null` behavior | Invalid result |
| --- | --- | --- | --- | --- |
| `bucket` | String satisfying the active namespace rule in §7.3. | Invalid when operation requires bucket. | Invalid. | `object_store_invalid_request` before backend call. |
| `server_generated_key` | Core 01 `object_blob_storage_key_v1`; max 1024 UTF-8 bytes. | Invalid when required. | Invalid. | `object_store_invalid_request`. |
| `migration_source_key` | Opaque byte-preserving S3 key represented as UTF-8 string or implementation byte sequence equivalent. | Invalid when required. | Invalid. | Preflight failure if unrepresentable. |
| `byte_size` | Integer `0..limits.object_blobs.max_declared_byte_size` where Core owner exposes the limit. | Invalid when required. | Invalid. | Owner public validation where route-facing; adapter invalid request otherwise. |
| `content_type` | Optional string without CR or LF; max 255 bytes UTF-8. | Stored as absent. | Stored as absent. | Invalid request. |
| `sha256_hex` | Optional lowercase 64-character hex. | Absent. | Absent. | Invalid request. |
| `metadata` | Object using only keys in §8.3; each value max 1024 UTF-8 bytes; total metadata max 8192 UTF-8 bytes. | Empty object. | Invalid. | Invalid request. |
| `purpose` | Closed enum in §8.4. | Invalid when operation requires purpose. | Invalid. | Invalid request. |
| `continuation_token` | Adapter-opaque string returned by the same operation. | Start first page. | Invalid. | Invalid request. |
| `expires_at` | RFC3339 UTC timestamp no later than 60 minutes after blob-slot creation for public upload targets. | Invalid for target creation. | Invalid. | Owner public validation or invalid request. |

### 8.3 Adapter metadata vocabulary

The adapter metadata vocabulary is closed for migration work.

| Metadata key | Allowed operation | Required value | Public exposure |
| --- | --- | --- | --- |
| `cartulary-object-blob-id` | `PutObject` for product evidence and migration copy | Stable `object_blob_id` when known | Forbidden in public object-store response; may appear in operator-private manifest. |
| `cartulary-upload-contract-sha256` | `PutObject` for product evidence | Lowercase SHA-256 of canonical accepted upload contract when available | Forbidden publicly. |
| `cartulary-migration-run-id` | migration copy and validation objects | Current migration `run_id` | Redacted or omitted from shareable summary. |
| `cartulary-probe-id` | probe objects | Current `probe_id` | May appear redacted in probe artifact. |

Unknown metadata keys MUST be rejected before backend calls for plan-owned operations. A future owner may add metadata keys only by extending this table or delegating a closed vocabulary to an owner document.

### 8.4 Purpose enum

| `purpose` | Allowed operations | Production allowed | Meaning |
| --- | --- | ---: | --- |
| `product_upload` | `CreateUploadTarget`, `HeadObject`, finalization metadata checks | yes | Ordinary blob-slot upload path. |
| `product_read` | `HeadObject`, `GetObject`, `GetObjectRange` | yes | Evidence preview/download through the application. |
| `probe_startup` | `PutObject`, `HeadObject`, `GetObject`, `GetObjectRange`, `DeleteObject` | yes | Startup capability probe. |
| `test_cleanup` | `ListPrefix`, `DeleteObject` | no | Harness cleanup under run-owned scope. |
| `backup_manifest` | `ListPrefix`, `HeadObject`, `GetObject` | yes | Backup manifest generation and verification. |
| `restore_verification` | `ListPrefix`, `HeadObject`, `GetObject` | yes | Restore verification only. |
| `migration_copy` | `ListPrefix`, `HeadObject`, `GetObject`, `PutObject` | yes, operator-only | Application-stopped migration copy. |
| `migration_validation` | `ListPrefix`, `HeadObject`, `GetObject`, `DeleteObject` for migration-test objects only | yes, operator-only | Migration validation. |
| `diagnostic` | `HeadObject`, bounded `ListPrefix` where owner-approved | yes, operator-only | Operator diagnostics. |

Unknown `purpose` is invalid. Product routes MUST NOT allow clients to supply `purpose`.

### 8.5 Operation contract

| Operation | Inputs | Output | Default timeout | Retry policy | Idempotency | Errors | Production permission |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `CreateUploadTarget` | `bucket`, server-generated `key`, `byte_size`, optional `content_type`, optional `sha256_hex`, `expires_at`, `upload_mode`. | `upload_target` bound to one method, one key, one size contract, one allowed header set, and one expiry. | 10s. | §8.6 with `max_total_attempts=2`; no retry after target is returned. | MUST NOT create more than one authoritative pending blob slot. | `object_store_unavailable`, `object_store_access_rejected`, `object_store_invalid_request`. | Allowed. |
| `PutObject` | `bucket`, key, replayable byte stream, declared `size`, optional metadata, `purpose`. | `etag` if available, `size_bytes`, stored metadata echo. | 30s to first byte accepted; total deadline caller-supplied. | Retry only when no bytes were accepted and stream is replayable; `max_total_attempts=2`. | Same `(bucket,key)` overwrite allowed only for probe/test/migration overwrite rules; product evidence keys are write-once. | Same plus `object_store_integrity_mismatch`. | Allowed only by purpose. |
| `HeadObject` | `bucket`, key, `purpose`. | `exists`, `size_bytes`, `etag` if available, closed metadata values when available. | 10s. | §8.6 with `max_total_attempts=2`. | Read operation. | `object_not_found`, dependency errors. | Allowed. |
| `GetObject` | `bucket`, key, `purpose`. | Stream plus `size_bytes` and metadata when available. | 30s to first byte; 30s inactivity timeout after first byte. | No adapter-level retry after bytes are emitted. | Read operation. | `object_not_found`, dependency errors. | Allowed. |
| `GetObjectRange` | `bucket`, key, required inclusive `start`, optional inclusive `end`, `purpose`. | Stream of bytes `start..end` inclusive, or `start..EOF` when `end` omitted. | 30s to first byte; 30s inactivity timeout. | No adapter-level retry after bytes are emitted. | Read operation. | `object_range_not_satisfiable`, `object_not_found`, dependency errors. | Allowed when owner range contract requires it. |
| `DeleteObject` | `bucket`, key, `purpose`. | `deleted=true` or `not_found=true`. | 30s. | §8.6 with `max_total_attempts=2`. | Idempotent when object is already absent. | Dependency errors. | Allowed only for cleanup, validation, and owner-approved cleanup flows. |
| `ListPrefix` | `bucket`, `prefix`, `purpose`, optional continuation token. | Object summaries ordered by key bytes ascending and adapter-opaque continuation token. | 30s per page. | §8.6 per page with `max_total_attempts=2`. | Read operation. | Dependency errors. | Allowed only by purpose; never user-facing query primitive. |
| `EnsureBucketForDevTest` | `bucket`, `profile`. | `created`, `already_exists`, or failure. | 30s. | §8.6 with `max_total_attempts=2`. | Idempotent for existing bucket. | Dependency errors. | Forbidden in `disconnected_prod` and `on_prem_or_cloud_external` application startup. |

### 8.6 Adapter retry algorithm

All adapter retry behavior MUST use this algorithm.

```text
run_with_retry(operation, max_total_attempts, per_attempt_timeout, total_deadline):
  attempt_index = 1
  while true:
    if now() >= total_deadline:
      return failure(reason_code="deadline_exceeded")
    attempt_deadline = min(now() + per_attempt_timeout, total_deadline)
    result = run operation with attempt_deadline
    if result is success:
      return result
    if result emitted response bytes, accepted request bytes, or returned an upload target:
      return result.failure
    if result.failure.retryable is false:
      return result.failure
    if attempt_index >= max_total_attempts:
      return failure(reason_code="retry_exhausted", cause=result.failure)
    wait deterministic_backoff(attempt_index)
    attempt_index = attempt_index + 1
```

Default deterministic backoff:

| Attempt before retry | Backoff |
| ---: | ---: |
| 1 | 100 ms |

No jitter is allowed in conformance tests. Production implementations MAY add bounded jitter only when tests can disable it and observable retry limits, attempt counts, deadlines, and final failure shape remain identical.

`max_total_attempts` includes the first attempt.

### 8.7 Range and stream rules

`GetObjectRange.start` and `GetObjectRange.end` use zero-based inclusive byte positions. `start` is required. Omitted `end` means stream from `start` through EOF. `start > end` is an adapter caller error and MUST NOT issue a backend request. A range beginning at EOF or beyond EOF MUST map to `object_range_not_satisfiable`.

All successful stream outputs MUST require the caller to close the stream. The adapter MUST make stream closure observable in tests by exposing a close hook, leak detector, or equivalent test-only fixture. A leaked stream in adapter tests is a harness or product test failure according to the owning test row.

`etag` is diagnostic only unless a future owner explicitly classifies a backend ETag as reliable for one bounded operation. ETag MUST NOT be used as migration copy equivalence.

### 8.8 Internal adapter error registry

| Internal error | Meaning | Retryable default | Public mapping family |
| --- | --- | ---: | --- |
| `object_store_unavailable` | Endpoint, bucket, or operation is temporarily unavailable. | `true` | `object_store_unavailable` after owner patch. |
| `object_store_access_rejected` | Credentials, CORS, permission, or capability blocks required behavior. | `false` | `object_store_access_rejected` after owner patch. |
| `object_not_found` | The requested object key is absent or not visible to the configured credential. | `false` | Evidence state-specific errors when tied to evidence; otherwise adapter-only. |
| `object_range_not_satisfiable` | Requested range is outside object byte domain. | `false` | Evidence access failure or adapter test failure, depending caller. |
| `object_store_integrity_mismatch` | Observed size or checksum conflicts with accepted contract. | `false` | `evidence_attach_rejected/accepted_contract_mismatch` or validation artifact mismatch. |
| `object_store_invalid_request` | Adapter caller supplied invalid bucket, key, range, metadata, purpose, unsupported profile operation, or evidence route construction found malformed persisted physical-key metadata before backend access. | `false` | Owner-owned public validation for malformed or identity-mismatched persisted `object_blobs.storage_key`; otherwise product bug in tests. |
| `object_store_cleanup_failed` | Cleanup after probe/test/migration left a reserved object behind. | `false` | Harness artifact failure, readiness failure, or warning as defined in §9.5. |
| `object_store_deadline_exceeded` | The operation exceeded the applicable total deadline. | `true` when caller may retry externally | Dependency error family. |
| `object_store_retry_exhausted` | Retryable operation failed after maximum attempts. | `false` | Dependency error family. |

## 9. Capability probe and runtime health

### 9.1 Probe inputs and defaults

| Field | Default or rule |
| --- | --- |
| `schema_id` | `cartulary.object_store_capability_probe.v1`. |
| `probe_id` | 26-character lowercase ULID or equivalent CSPRNG-backed stable ID generated once per startup attempt. |
| Probe prefix | `.cartulary/probes/startup/<probe_id>/`. |
| Primary key | `<probe_prefix>probe.bin`. |
| Secondary direct-upload key | `<probe_prefix>direct-put.bin`. |
| Primary payload | UTF-8 bytes `cartulary-object-store-probe-v1\n` followed by 16 CSPRNG bytes. |
| Secondary direct-upload payload | UTF-8 bytes `cartulary-object-store-direct-put-probe-v1\n` followed by 16 CSPRNG bytes. |
| Total deadline | 60s. |
| Endpoint attempt timeout | 5s per attempt. |
| Metadata attempt timeout | 10s per attempt. |
| Object mutation timeout | 30s per attempt. |
| Retry count | At most 2 total attempts per required step, including the first attempt. |
| Cleanup | Attempt delete for every probe object created by the current run before returning success or failure. |

### 9.2 Probe stage enum

`failed_stage` and `steps[].stage` MUST use this closed vocabulary.

| Stage | Required operation |
| --- | --- |
| `endpoint_reachability` | Resolve and connect to the configured S3 endpoint without using admin APIs. |
| `bucket_validation` | Confirm bucket existence or absence using S3-compatible behavior. |
| `bucket_creation` | Create bucket only in `local_dev` or `ci_service_backed` when missing. |
| `put_primary` | `PutObject` primary payload. |
| `head_primary` | `HeadObject` primary object. |
| `get_primary` | `GetObject` primary object and verify bytes. |
| `range_primary` | `GetObjectRange` for primary object and verify bytes. |
| `create_direct_upload_target` | Generate direct presigned PUT target for secondary key. |
| `cors_preflight` | Validate CORS policy for direct-upload browser origin when direct upload is active. |
| `direct_put` | Use generated target to upload secondary payload. |
| `head_direct` | `HeadObject` secondary object. |
| `get_direct` | `GetObject` secondary object and verify bytes. |
| `delete_primary` | Delete primary probe object. |
| `delete_direct` | Delete secondary probe object. |
| `verify_primary_deleted` | Confirm primary object absence. |
| `verify_direct_deleted` | Confirm secondary object absence. |
| `cleanup_after_failure` | Attempt deletion of any objects created before a failed stage. |

### 9.3 Probe reason-code enum

| Reason code | Meaning | Retryable default |
| --- | --- | ---: |
| `endpoint_unreachable` | S3 endpoint cannot be reached within the endpoint timeout. | `true` |
| `endpoint_tls_rejected` | TLS or endpoint identity validation fails. | `false` |
| `bucket_missing` | Required bucket does not exist in a profile where app startup cannot create it. | `false` |
| `bucket_create_forbidden` | Bucket creation was attempted in a profile where it is forbidden. | `false` |
| `bucket_create_failed` | Dev/test bucket creation failed. | `true` |
| `credential_denied` | Configured credential cannot perform required action. | `false` |
| `cors_rejected` | Direct-upload browser preflight fails policy. | `false` |
| `capability_missing` | Required S3 operation is unsupported or behaves incompatibly. | `false` |
| `size_mismatch` | Stored object size differs from expected. | `false` |
| `hash_mismatch` | Retrieved bytes differ from expected payload bytes. | `false` |
| `range_unsupported` | Required range read is unavailable. | `false` |
| `cleanup_failed` | Probe object cleanup failed. | `false` |
| `deadline_exceeded` | Probe deadline expired. | `true` |
| `retry_exhausted` | Retryable operation exhausted attempts. | `false` |
| `invalid_configuration` | Effective deployment configuration cannot be used for object store startup. | `false` |

### 9.4 Probe algorithm

```text
probe_object_store(profile, upload_mode):
  create result object with schema_id and probe_id
  derive prefix and probe keys from §9.1
  set total_deadline = now + 60 seconds

  run endpoint_reachability
  run bucket_validation
  if bucket is missing:
    if profile in {local_dev, ci_service_backed}:
      run bucket_creation
    else:
      fail bucket_validation with reason_code="bucket_missing"

  run put_primary
  run head_primary and require expected size
  run get_primary and require exact payload bytes
  run range_primary with range [4, 15] and require exact byte slice

  if upload_mode == "direct_presigned_put":
    run create_direct_upload_target
    run cors_preflight from application.public_origin
    run direct_put using only headers allowed by cors_direct_put_v1
    run head_direct and require expected size
    run get_direct and require exact direct payload bytes

  run delete_primary when primary object was created
  run delete_direct when direct object was created
  run verify_primary_deleted when primary delete returned success
  run verify_direct_deleted when direct delete returned success

  classify cleanup according to §9.5
  return pass only when all required stages pass and cleanup outcome is clean
```

The probe MUST use S3-compatible operations only. It MUST NOT call SeaweedFS master, filer, volume, WebDAV, lifecycle, user-management, or admin APIs.

### 9.5 Probe result schema

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_capability_probe.v1`. |
| `probe_id` | string | Same generated value used in probe keys. |
| `profile_id` | enum | Active runtime profile. |
| `upload_mode` | enum | `direct_presigned_put` in this migration unless later owner profile defines another value. |
| `started_at` | RFC3339 UTC timestamp | Required. |
| `completed_at` | RFC3339 UTC timestamp | Required after terminal result. |
| `result` | enum | `pass`, `fail`, or `pass_with_cleanup_warning`. |
| `failed_stage` | enum or null | Null on pass; otherwise §9.2 value. |
| `reason_code` | enum or null | Null on pass; otherwise §9.3 value. |
| `retryable` | boolean | Required. |
| `endpoint` | redaction object | `endpoint_ref` shape from §2.1. |
| `bucket_ref` | redaction object | `bucket_ref` shape from §2.1. |
| `steps[]` | array | Ordered by execution order. |
| `steps[].stage` | enum | §9.2 value. |
| `steps[].status` | enum | `pass`, `fail`, `skipped`, or `cleanup_warning`. |
| `steps[].attempt_count` | integer | `0..2`; skipped uses `0`. |
| `steps[].started_at` | timestamp or null | Null only when skipped. |
| `steps[].completed_at` | timestamp or null | Null only when skipped. |
| `steps[].reason_code` | enum or null | Null unless failed or warning. |
| `cleanup_result` | enum | §9.6 value. |
| `retained_probe_keys[]` | array | Redacted `object_key_ref` values only; empty when cleanup is clean. |

### 9.6 Probe cleanup outcomes

| Cleanup outcome | Meaning | `local_dev` | `ci_service_backed` | `disconnected_prod` | Release gate |
| --- | --- | --- | --- | --- | --- |
| `clean` | Every created probe object was deleted and verified absent. | pass | pass | pass | pass |
| `retained_under_reserved_probe_prefix` | Cleanup failed, and every retained object is under `.cartulary/probes/startup/<probe_id>/`. | `pass_with_cleanup_warning` | fail unless diagnostic retention was explicitly requested by harness | fail readiness | fail |
| `retained_outside_reserved_probe_prefix` | Any retained object is outside the reserved prefix. | fail | fail | fail | fail |
| `cleanup_not_attempted` | Cleanup was not attempted for an object that may have been created. | fail | fail | fail | fail |

A `pass_with_cleanup_warning` result MUST NOT satisfy release gates. It MAY allow local developer startup only in `local_dev`, and the result artifact MUST record retained object redaction refs.

### 9.7 Runtime health state machine

The internal health state machine is closed by this plan. Public readiness transport shape remains blocked by `SWFS-OWNER-HEALTH-001` until the owner document adopts the response shape.

| State | Entry event | Liveness | Readiness | Evidence operations | Non-evidence operations | Exit events |
| --- | --- | --- | --- | --- | --- | --- |
| `starting_object_store_probe` | Process start after configuration validation | healthy | not ready | unavailable | unavailable | `probe_passed`, `probe_failed` |
| `ready` | Probe passed with `cleanup_result=clean` | healthy | ready | allowed | allowed | `object_store_operation_failed`, `shutdown_started` |
| `degraded_object_store` | Dependency failure after ready | healthy | blocked by `SWFS-OWNER-HEALTH-001` | fail through §10.2 | continue when no object-store dependency is required | `reprobe_started`, `shutdown_started` |
| `recovering_object_store` | Runtime reprobe started | healthy | blocked by `SWFS-OWNER-HEALTH-001` | fail or retry only by owner route contract | continue when no object-store dependency is required | `probe_passed`, `probe_failed`, `shutdown_started` |
| `stopped` | Shutdown started or terminal startup failure | not applicable | not ready | unavailable | unavailable | none |

The application MUST NOT keep evidence upload, attach, preview, or download operations silently successful when the object store is known degraded. The application MAY continue ordinary non-evidence workbook operations when those operations do not require object storage.

## 10. Evidence route behavior and public error mapping

### 10.1 Claimability of public error mapping

Rows using `object_store_unavailable` or `object_store_access_rejected` have `closure_class=blocked_until_owner_patch` until `SWFS-OWNER-ERR-001` lands. Existing evidence-state errors remain claimable only where they already exist in Core-owned registries. No test or implementation report may claim runtime object-store dependency mapping complete while `SWFS-OWNER-ERR-001` is unresolved.

### 10.2 Runtime dependency outage mapping

| Condition | Route family | Required public error after owner patch | Required reason code | Retry hint |
| --- | --- | --- | --- | --- |
| Endpoint unreachable | Blob-slot create, attach finalization, preview issuance, download issuance, handle redemption | `object_store_unavailable` | `endpoint_unreachable` | retryable |
| Bucket missing after ready | Same | `object_store_unavailable` | `bucket_missing` | retryable only after operator repair |
| Credentials denied | Same | `object_store_access_rejected` | `credential_denied` | not retryable without operator repair |
| Capability missing | Same | `object_store_access_rejected` | `capability_missing` | not retryable |
| CORS rejects direct upload | Blob-slot create or browser upload validation path | `object_store_access_rejected` | `cors_rejected` | not retryable without configuration change |
| Range unavailable for required preview | Preview issuance or handle redemption | `object_store_access_rejected` | `capability_missing` | not retryable unless owner range contract defines fallback |
| Runtime timeout after max attempts | Same | `object_store_unavailable` | `retry_exhausted` | retryable |
| Malformed persisted physical key | Attach finalization, preview issuance, download issuance, handle redemption | `object_store_invalid_request` | `object_blob_storage_key_malformed` | not retryable |
| Identity-mismatched persisted physical key | Attach finalization, preview issuance, download issuance, handle redemption | `object_store_invalid_request` | `object_blob_storage_key_identity_mismatch` | not retryable |

These errors MUST NOT replace existing state-specific evidence errors. Missing, pending, failed, quarantined, inconsistent, unsupported-preview, and preview-size conditions MUST continue to use existing Core-owned evidence errors when the object-store dependency is reachable and the authoritative evidence/blob state determines the result.

### 10.3 Evidence state matrix

| State | Upload target use | Attach finalization | Preview-handle issuance | Download-handle issuance | Handle redeem after issuance |
| --- | --- | --- | --- | --- | --- |
| Pending blob | Upload may proceed before target expiry. | Reject with owner-owned pending-blob attach error. | Reject with owner-owned pending-blob access error. | Reject with owner-owned pending-blob access error. | Reject with owner-owned pending-blob access error. |
| Failed blob | Same slot cannot be refreshed. | Reject with owner-owned failed-blob attach error. | Reject with owner-owned failed-blob access error. | Reject with owner-owned failed-blob access error. | Reject with owner-owned failed-blob access error. |
| Missing backing object | No effect on new create. | Reject with owner-owned not-visible attach error unless still pending and observable. | Reject with owner-owned missing-blob access error. | Reject with owner-owned missing-blob access error. | Reject with owner-owned missing-blob access error. |
| Quarantined blob | No attach. | Reject with owner-owned quarantined-blob attach error. | Reject with owner-owned quarantined evidence access error. | Reject with owner-owned quarantined evidence access error. | Reject with owner-owned quarantined evidence access error. |
| Quarantined evidence | No attach. | Reject with owner-owned quarantined evidence attach error. | Reject with owner-owned quarantined evidence access error. | Reject with owner-owned quarantined evidence access error. | Reject with owner-owned quarantined evidence access error. |
| Oversized preview | Not applicable. | Not applicable. | Reject with owner-owned preview-too-large error. | May issue when download otherwise allowed. | Existing preview handle MUST NOT have been issued for oversized preview. |
| Unsupported preview | Not applicable. | Not applicable. | Reject with owner-owned unsupported-preview error. | May issue when download otherwise allowed. | Existing preview handle MUST NOT have been issued for unsupported preview. |
| Expired upload target | Same slot cannot refresh. | Reject if bytes were not successfully observed before expiry. | Reject while blob remains pending. | Reject while blob remains pending. | Reject while blob remains pending. |
| Expired handle | Not applicable. | Not applicable. | Fresh issuance may succeed. | Fresh issuance may succeed. | Owner-owned expired-handle response. |
| Consumed download handle | Not applicable. | Not applicable. | Not applicable. | Fresh issuance may succeed. | Owner-owned consumed-handle response. |
| Stale row version on attach | Not applicable. | Owner-owned row-version conflict. | Not applicable. | Not applicable. | Not applicable. |

This table is a mapping guard. It does not rename existing Core-owned error codes.

### 10.4 Range-failure behavior

| Condition | Required behavior | Closure class |
| --- | --- | --- |
| Range unavailable before startup | Startup fails when preview or download implementation declares range as required. If only preview declares range as required, preview readiness fails and the profile is not ready for preview. | `blocked_until_owner_patch` until `SWFS-OWNER-RANGE-001` lands. |
| Range unavailable after readiness | Preview operations requiring range fail with `object_store_access_rejected/capability_missing` after `SWFS-OWNER-ERR-001`; full download uses `GetObject` and may continue only when `SWFS-OWNER-RANGE-001` explicitly allows full-stream fallback. | `blocked_until_owner_patch`. |
| Range not satisfiable for a specific object | Map to owner-owned evidence inconsistency or range error unless preview-size or unsupported-preview owner rules are more specific. | `blocked_until_owner_patch`. |

## 11. Harness and compatibility profile

### 11.1 Canonical artifact vocabulary

Harness, compatibility, and migration artifacts MUST use backend-neutral fields. Backend identity belongs in `object_store_backend`.

| Field | Required use |
| --- | --- |
| `object_store_backend` | Exact value `seaweedfs_s3` for default SeaweedFS runs. |
| `s3_bucket` | Canonical bucket field, raw only in operator-private artifacts. |
| `s3_bucket_ref` | Redacted bucket object for shareable summaries. |
| `s3_prefix` | Canonical prefix field when prefix isolation is used. |
| `object_store_endpoint` | Endpoint value with credentials redacted, operator-private only. |
| `object_store_endpoint_ref` | Redacted endpoint object for shareable summaries. |
| `object_store_capability_probe` | Structured probe result from §9.5. |
| `object_store_failure_reason` | Stable harness reason code. |

The following artifact fields are forbidden in new or revised default artifacts: `minio_bucket`, `minio_endpoint`, `minio_container`, `minio_ready`, `minio_access_key`, and `minio_secret_key`.

### 11.2 Compatibility cases

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
| `SWFS-COMP-009` | Error classification | Missing object, denied credential, missing bucket, unreachable endpoint, integrity mismatch, and CORS rejection fixtures. | Each maps to the backend-neutral error registry or blocked owner error row. |
| `SWFS-COMP-010` | No public listing primitive | Product route inventory and UI route registry. | No public API or workbook route exposes object prefix listing. |
| `SWFS-COMP-011` | CORS preflight | Direct-upload browser origin equal to `application.public_origin`, `Origin: null`, and actual direct PUT response headers. | Browser preflight allows exactly the §6.3 direct-upload method/header set, wildcard and `Origin: null` are rejected, credentials are not enabled, and the direct PUT response exposes only `etag`. |
| `SWFS-COMP-012` | Same-origin preview/download | Existing evidence preview and download flow. | Public evidence access uses same-origin handles only. |
| `SWFS-COMP-013` | Probe cleanup classification | Forced cleanup failure under and outside probe prefix. | Outcomes classify exactly according to §9.6. |
| `SWFS-COMP-014` | Canonical artifact serialization | Probe, backup, restore, and validation artifacts. | Duplicate JSON keys invalid; canonical digests stable. |

No compatibility row may be skipped because multipart upload or presigned GET is absent. Those capabilities are outside the current migration profile.

## 12. Backup and restore contract

Core 01 requires backup and restore behavior to cover Postgres and object-store contents from the same retained `backup_set` and `consistency_point_at`; restore must restore Postgres, restore object-store contents, and rebuild projections.

### 12.1 Operator-private backup manifest

A successful backup against SeaweedFS MUST include one operator-private object-store backup manifest.

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_backup_manifest.v1`. |
| `backup_set_id` | string | Must match enclosing backup set. |
| `consistency_point_at` | RFC3339 timestamp | Must match enclosing backup set. |
| `object_store_backend` | string | `seaweedfs_s3`. |
| `bucket` | string | Required raw bucket name in operator-private artifact. |
| `object_count` | integer | Count of manifest objects. |
| `total_size_bytes` | integer | Sum of object byte sizes. |
| `objects[]` | array | Sorted by `storage_ref` bytewise ascending. |
| `objects[].object_blob_id` | string | Required when object corresponds to an authoritative blob. |
| `objects[].storage_ref` | string | Required for every restoreable object. |
| `objects[].storage_ref_sha256` | lowercase hex SHA-256 | Required. |
| `objects[].size_bytes` | integer | Required and MUST be `>=0`. |
| `objects[].sha256` | lowercase hex SHA-256 | Required and computed from backed-up object bytes. MUST NOT be null. |
| `objects[].backup_member_sha256` | lowercase hex SHA-256 | Digest of the backup artifact member bytes. |
| `manifest_sha256` | lowercase hex SHA-256 | Digest of canonical manifest bytes excluding this field. |

The manifest MUST contain enough information to restore object bytes without any external mapping. A shareable summary MUST NOT be accepted as restore input.

### 12.2 Shareable backup summary

A shareable summary MAY be emitted only as a redacted derivative of the operator-private manifest. If omitted, no shareable summary exists and restore input remains the operator-private manifest.

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_backup_summary.v1`. |
| `backup_set_id` | string | Must match private manifest. |
| `consistency_point_at` | RFC3339 timestamp | Must match private manifest. |
| `object_store_backend` | string | `seaweedfs_s3`. |
| `bucket_ref` | redaction object | Required; raw bucket forbidden. |
| `object_count` | integer | Same as private manifest. |
| `total_size_bytes` | integer | Same as private manifest. |
| `manifest_sha256` | lowercase hex SHA-256 | Digest of private canonical manifest. |
| `objects_summary[]` | array or omitted | If present, each item MUST omit raw `storage_ref` and include `storage_ref_sha256`, `size_bytes`, and `sha256`. |

The summary MUST NOT contain `bucket`, `storage_ref`, raw object keys, endpoint URLs, credentials, or raw storage backend paths.

### 12.3 Canonical manifest serialization

Backup manifests, backup summaries, and restore verification artifacts MUST use canonical JSON:

- UTF-8 without BOM;
- lexicographically sorted object keys;
- no insignificant whitespace;
- LF line ending;
- arrays in the specified order;
- duplicate JSON object keys invalid;
- digest fields computed with the digest field omitted.

### 12.4 Latest retained backup selection

Restore verification MUST select exactly one retained backup set with this deterministic algorithm.

```text
select_latest_successful_retained_backup(backups):
  candidates = backups where status == "successful" and retention_state == "retained"
  require candidates is non-empty
  sort by consistency_point_at DESC, completed_at DESC, backup_set_id bytewise DESC
  return first candidate
```

A backup MUST NOT be classified as the latest successful retained backup if the object-store manifest is missing, unreadable, inconsistent with the backup set, detached from the Postgres consistency point, or contains any object with `sha256=null`.

### 12.5 Restore verification

Restore verification MUST execute this sequence:

1. Select exactly one retained `backup_set` by §12.4.
2. Restore Postgres from that `backup_set`.
3. Restore SeaweedFS object data from the same `backup_set`.
4. Rebuild projections.
5. Verify that every authoritative blob requiring durable bytes has bytes present in the restored object store.
6. Verify manifest size and SHA-256 proofs for every manifest object.
7. If restored data contains incidents, select the lowest `incident_id` by bytewise ascending order.
8. If an incident was selected, run one built-in workbook query against `cartulary.view.timeline.v1` with owner-defined default sort and limit `1`, unless Core defines another restore-verification query.
9. Mark verification failed if any required blob byte, manifest proof, lifecycle invariant, projection rebuild, or workbook query fails.

Zero-incident backups MAY pass restore verification only when all blob and manifest checks pass and the artifact records `incident_open_check.status='skipped_no_incidents'`. If omitted, the verifier MUST fail with a closed reason code.

### 12.6 Restore verification artifact

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.restore_verification.v1`. |
| `backup_set_id` | string | Selected backup. |
| `selected_incident_id` | string or null | Null only when no incidents exist. |
| `incident_open_check.status` | enum | `pass`, `fail`, or `skipped_no_incidents`. |
| `query_view_schema_id` | string or null | `cartulary.view.timeline.v1` when query executed. |
| `blob_check_counts.total` | integer | Required. |
| `blob_check_counts.passed` | integer | Required. |
| `blob_check_counts.failed` | integer | Required. |
| `manifest_check_result` | enum | `pass` or `fail`. |
| `projection_rebuild_result` | enum | `pass` or `fail`. |
| `result` | enum | `pass` or `fail`. |
| `failure_reasons[]` | array | Empty only when `result='pass'`; closed reason codes. |
| `artifact_sha256` | lowercase hex SHA-256 | Digest of canonical artifact bytes excluding this field. |

## 13. MinIO-to-SeaweedFS migration contract

### 13.1 Migration posture

The default migration mode is application-stopped migration. The running Cartulary application MUST be stopped or otherwise incapable of accepting writes before coherent backup capture, object copy, validation, or cutover begins.

No live maintenance-mode product workflow is introduced by this plan. A later live migration feature requires a separate owner document defining write freeze, public route outcomes, operator state, background jobs, and rollback behavior.

### 13.2 Write-quiescence proof

| Proof kind | Valid? | Required evidence |
| --- | ---: | --- |
| `process_stopped` | yes | App process absent or stopped by supervisor; HTTP listener closed; WebSocket listener closed. |
| `listener_disabled` | yes only if repo has operator control for it | Listener-disable artifact and failed write probe. |
| `owner_defined_write_gate` | blocked until owner patch | Named owner patch and write-rejection route outcomes. |
| `operator_assertion_only` | no | Not sufficient. |

The migration tool MUST record exactly one valid write-quiescence proof before `backup_captured`. Operator assertion may appear as supplementary context, but it MUST NOT satisfy the proof requirement.

### 13.3 Migration run artifact

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_migration_run.v1`. |
| `run_id` | string | Stable migration run identifier. |
| `created_at` | RFC3339 UTC | Required. |
| `updated_at` | RFC3339 UTC | Required after every state transition. |
| `current_state` | enum | One state from §13.4. |
| `state_timestamps` | object | One timestamp for every entered state. |
| `events[]` | array | Ordered append-only event ledger using §13.4 events. |
| `operator_identity` | string or redacted object | Required. |
| `source_endpoint_ref` | redaction object | Required. |
| `target_endpoint_ref` | redaction object | Required. |
| `source_bucket_ref` | redaction object | Required in shareable form; raw source bucket allowed only in operator-private extension. |
| `target_bucket_ref` | redaction object | Required in shareable form; raw target bucket allowed only in operator-private extension. |
| `backup_refs[]` | array | Required when backup exists. |
| `probe_ref` | string or object | Required before `target_prepared`. |
| `copy_ledger_ref` | string or null | Required by `copying`. |
| `validation_ref` | string or null | Required by `cutover_ready`. |
| `rollback_ref` | string or null | Required when rolled back. |
| `terminal_result` | enum or null | `post_cutover_verified`, `rolled_back`, or `failed` when terminal. |

### 13.4 Migration lifecycle machine

| Event | Allowed source state | Guard | Action | Destination |
| --- | --- | --- | --- | --- |
| `plan_created` | none | endpoints supplied | create migration-run artifact | `planned` |
| `preflight_passed` | `planned` | source/target credentials validated; target feature preflight complete | record preflight proof | `preflighted` |
| `write_quiescence_verified` | `preflighted` | proof accepted by §13.2 | record proof | `application_stopped` |
| `backup_captured` | `application_stopped` | coherent backup artifact exists | attach backup refs | `backup_captured` |
| `target_prepared` | `backup_captured` | target probe passes cleanly | record probe artifact | `target_prepared` |
| `copy_started` | `target_prepared` | no blocking preflight diagnostics | open copy ledger | `copying` |
| `copy_completed` | `copying` | every object copied or already equivalent | close copy ledger | `copied` |
| `validation_started` | `copied` | copy ledger closed with no blocking errors | open validation artifact | `validating` |
| `validation_passed` | `validating` | validation artifact `result='pass'` | attach validation artifact | `cutover_ready` |
| `cutover_committed` | `cutover_ready` | endpoint config updated and app remains stopped | record cutover config diff | `cutover_committed` |
| `post_cutover_verified` | `cutover_committed` | evidence and backup/restore gates pass; migration credentials revoked | record post-cutover proof | `post_cutover_verified` |
| `rollback_requested` | `backup_captured`, `target_prepared`, `copying`, `copied`, `validating`, `cutover_ready`, `cutover_committed` | rollback still open by §13.7 | execute rollback action | `rolled_back` |
| `blocking_failure` | any nonterminal state | failure recorded | close or roll back by boundary rule | `failed` or `rolled_back` |

Terminal states are `post_cutover_verified`, `rolled_back`, and `failed`. A terminal migration run MUST NOT transition to another state. Resume from a nonterminal state is allowed only by replaying the event ledger and verifying all referenced artifacts still exist and match their recorded digests.

### 13.5 Copy semantics

Copy idempotency key MUST be:

```text
(source_bucket, source_key, target_bucket, target_key, source_size_bytes, source_sha256)
```

Default copy MUST preserve source bucket name and object keys. The migration utility MUST NOT update database `storage_ref` values during default migration.

Byte equality is authoritative. SHA-256 MUST be computed from source bytes before target equivalence is accepted. ETag MUST NOT be used as source hash or target equivalence proof.

| Source/target condition | Required result |
| --- | --- |
| Source object exists and target absent | Copy bytes and closed metadata. |
| Target exists with same size and SHA-256 | Count as `already_copied`; do not rewrite. |
| Target exists with different size or SHA-256 | Blocking failure. |
| Target exists with same ETag but unknown SHA-256 | Not equivalent; compute SHA-256 or fail. |
| Source object is zero bytes | Valid; SHA-256 is the SHA-256 of the empty byte string. |
| Source bucket uses versioning or contains delete markers in copy scope | Preflight fails unless a later migration profile defines version semantics. |
| Source object has object tags, object lock, retention policy, lifecycle policy, or bucket policy | Feature is out of scope; copy ignores non-authoritative tags and fails preflight when retention/lock prevents byte-copy validation. |
| Unknown user metadata exists | Ignored unless owner docs classify it as authoritative. |
| Closed Cartulary metadata exists | Preserve keys in §8.3 where present. |

### 13.6 Copy ledger item schema

| Field | Required rule |
| --- | --- |
| `source_bucket_ref` | Redacted bucket object in shareable form; raw value allowed only in operator-private ledger. |
| `source_key_ref` | Redacted object-key object. |
| `target_bucket_ref` | Redacted bucket object. |
| `target_key_ref` | Redacted object-key object. |
| `source_size_bytes` | Required. |
| `source_sha256` | Required. |
| `target_size_bytes` | Required when target exists. |
| `target_sha256` | Required when target exists. |
| `status` | `copied`, `already_copied`, `missing_source`, `target_mismatch`, `unsupported_source_feature`, or `error`. |
| `reason_code` | Closed copy reason code. |

### 13.7 Rollback semantics

| Migration boundary | Rollback behavior |
| --- | --- |
| Before `cutover_committed` | Leave Cartulary pointed at the source object store and retain the pre-migration backup. |
| After `cutover_committed` and before `post_cutover_verified` | Stop the app; restore the pre-migration coherent backup if any writes occurred after cutover; point the app back to the source object store; verify the source path before reopening. |
| After `post_cutover_verified` | Migration rollback is closed. Recovery uses ordinary retained-backup restore procedures. |

Migration credentials MUST be revoked or rendered unusable before `post_cutover_verified` becomes terminal.

## 14. Migration validation artifact

### 14.1 Top-level schema

The migration validation tool MUST emit one operator-private artifact and MAY emit a redacted shareable summary derived from it. If omitted, no shareable summary exists and cutover eligibility is unchanged. Automation MUST be able to decide cutover eligibility from the operator-private artifact without reading prose.

| Field | Type | Required rule |
| --- | --- | --- |
| `schema_id` | string | Exact `cartulary.object_store_migration_validation.v1`. |
| `schema_version` | string | Exact `1.0.0`. |
| `validation_tool_version` | string | Exact repo-controlled version. |
| `run_id` | string | Stable migration run identifier. |
| `started_at` | RFC3339 UTC | Required. |
| `completed_at` | RFC3339 UTC or null | Null only while running. |
| `source_backend` | string | `minio_s3` for default MinIO-source migration; otherwise backend-neutral source label. |
| `target_backend` | string | `seaweedfs_s3`. |
| `source_snapshot_id` | string | Required. |
| `target_snapshot_id` | string | Required. |
| `source_bucket` | string or redaction object | Raw required in operator-private artifact; redacted in shareable summary. |
| `target_bucket` | string or redaction object | Raw required in operator-private artifact; redacted in shareable summary. |
| `incident_count` | integer | Count of incidents checked. |
| `object_blob_count` | integer | Count of authoritative object blobs checked. |
| `objects_checked[]` | array | Sorted by `object_blob_id` bytewise ascending. |
| `preview_sample_checks[]` | array | Deterministic bounded sample from §14.4. |
| `blocking_diagnostics[]` | array | Empty required for `result='pass'`. |
| `nonblocking_warnings[]` | array | May be non-empty for `result='pass'`. |
| `result` | enum | Exactly `pass` or `fail`. |
| `artifact_sha256` | lowercase hex SHA-256 | Digest of canonical artifact bytes excluding this field. |

### 14.2 `objects_checked[]` item schema

| Field | Required rule |
| --- | --- |
| `object_blob_id` | Required. |
| `incident_id` | Required. |
| `storage_ref_sha256` | Required. |
| `source_size_bytes` | Required when source object exists. |
| `target_size_bytes` | Required when target object exists. |
| `source_sha256` | Required when source object exists. |
| `target_sha256` | Required when target object exists. |
| `status` | `pass`, `missing_source`, `missing_target`, `size_mismatch`, `hash_mismatch`, `unsupported_source_feature`, or `error`. |
| `reason_code` | Closed reason code from §14.6. |

### 14.3 Diagnostic item schema

`blocking_diagnostics[]` and `nonblocking_warnings[]` items MUST use this shape.

| Field | Required rule |
| --- | --- |
| `diagnostic_id` | Stable diagnostic identifier unique within the artifact. |
| `severity` | `blocking` or `warning`. |
| `reason_code` | Closed reason code from §14.6. |
| `object_blob_id` | Required when object-specific; otherwise null. |
| `incident_id` | Required when incident-specific; otherwise null. |
| `message` | Human-readable message with no raw bucket, key, endpoint, credential, or secret. |
| `refs[]` | Redacted refs only. |

### 14.4 Preview sample algorithm

```text
sample_preview_checks(objects):
  candidates = authoritative objects where:
    object_blob_id is non-null
    evidence_state allows preview or download
    object bytes are required for the route
    validation caller has current authorization
  sort candidates by sha256(utf8(object_blob_id)) ascending, then object_blob_id bytewise ascending
  return first min(32, len(candidates))
```

| Candidate route class | Required validation |
| --- | --- |
| Previewable | Issue preview handle, redeem handle, verify expected success envelope and non-empty or valid zero-byte response as applicable. |
| Downloadable but not previewable | Issue download handle, redeem handle, verify exact byte count and SHA-256. |
| Fewer than 32 candidates | Use all candidates; result may pass with fewer than 32. |
| Zero candidates | `preview_sample_checks=[]`; result may pass only if no other blocking arrays are non-empty. |

### 14.5 Result computation

```text
result = "pass" only if:
  every objects_checked item has status == "pass"
  blocking_diagnostics is empty
  every preview_sample_checks item has status == "pass"
else result = "fail"
```

`nonblocking_warnings[]` MAY be non-empty for `result='pass'` only when every warning reason code is classified as nonblocking in §14.6. If omitted, it defaults to an empty array.

### 14.6 Validation reason-code registry

| Reason code | Blocking? | Meaning |
| --- | ---: | --- |
| `missing_source_object` | yes | Authoritative source object expected by database is absent from source. |
| `missing_target_object` | yes | Target object expected after copy is absent. |
| `size_mismatch` | yes | Source and target sizes differ. |
| `hash_mismatch` | yes | Source and target SHA-256 differ. |
| `unsupported_source_feature` | yes | Source bucket uses unsupported versioning, delete markers, object lock, or retention behavior. |
| `preview_handle_failed` | yes | Expected preview handle could not be issued or redeemed. |
| `download_handle_failed` | yes | Expected download handle could not be issued or redeemed. |
| `authorization_failed` | yes | Validation caller lacked required authorization for ordinary route check. |
| `artifact_schema_invalid` | yes | Validation artifact shape is invalid. |
| `redaction_applied` | no | Shareable summary redacted operator-private values. |
| `fewer_than_32_preview_candidates` | no | Deterministic sample had fewer than 32 eligible candidates. |
| `zero_preview_candidates` | no | No eligible candidates existed. |

### 14.7 Canonical validation serialization

Migration validation artifacts MUST use canonical JSON:

- UTF-8 without BOM;
- lexicographically sorted object keys;
- no insignificant whitespace;
- LF line ending;
- arrays in specified order;
- duplicate JSON object keys invalid;
- `artifact_sha256` computed with `artifact_sha256` omitted.

### 14.8 Release-shareable migration preservation evidence

The strict release gate MUST emit a redacted shareable migration-preservation summary with `schema_id='cartulary.seaweedfs_migration_preservation_evidence.v2'`. This summary is downstream of the operator-private migration run, copy ledger, and validation artifacts. It MUST NOT copy raw `source_bucket`, `target_bucket`, raw object keys, endpoints, or storage refs into release-shareable evidence.

The v2 summary MUST include `source_backend`, `target_backend`, `source_bucket_ref`, `target_bucket_ref`, `bucket_preserved`, `object_blob_count`, `copy_ledger_object_count`, artifact references, findings, and `result`. `bucket_preserved=true` requires equal source and target bucket redaction hashes. Object-key preservation remains proven through copy-ledger source/target key redaction hashes.

## 15. Security and threat-model update

The threat model MUST be updated before release because this migration materially changes the object-storage access pattern, deployment profile, backup/restore behavior, and migration workflow. Core 04 already requires threat-model updates before releases that change an object-storage access pattern or backup/restore mechanism.

| STRIDE class | Required SeaweedFS-specific coverage | Required evidence |
| --- | --- | --- |
| Spoofing | S3 endpoint identity, reverse-proxy trust, credential source, direct upload target scope. | Threat-model row plus probe/endpoint identity evidence. |
| Tampering | Object overwrite, object delete, object metadata drift, migration copy mismatch, backup manifest mismatch. | Threat-model row plus adapter and migration validation evidence. |
| Repudiation | Application-owned evidence attach audit remains authoritative; object-store logs are diagnostic only. | Threat-model row plus evidence attach audit assertion. |
| Information disclosure | S3 credentials, raw object keys, upload targets, preview/download handles, SeaweedFS admin/filer/master/volume UIs. | Threat-model row plus artifact redaction tests. |
| Denial of service | Oversized evidence, prefix listing abuse, storage exhaustion, range-read abuse, probe cleanup failure. | Threat-model row plus compatibility/error tests. |
| Elevation of privilege | Bucket wildcard credentials, anonymous access, exposed admin APIs, wildcard CORS, mistaken MinIO server default. | Threat-model row plus release manifest exposure scan. |

Security acceptance for this migration requires all of the following:

- no production release manifest exposes SeaweedFS admin, master, filer, volume, or WebDAV UI ports;
- no production CORS rule uses wildcard origin for direct upload;
- no retained harness, migration, backup, restore, or validation artifact contains raw secret values;
- no release-shareable SeaweedFS evidence artifact contains raw bucket names, object keys, raw storage refs, or raw object-store endpoints;
- no public preview/download response contains bucket names, object keys, raw storage refs, raw storage URLs, or long-lived object-store credentials;
- no runtime product code calls SeaweedFS admin APIs;
- no MinIO server artifact is present in release manifests except as a migration source fixture explicitly excluded from default runtime.

## 16. Documentation and terminology occurrence inventory

### 16.1 Terminology replacement rules

| Old wording | Required replacement |
| --- | --- |
| `MinIO container` as default runtime | `SeaweedFS S3 container` or `S3-compatible object store`, according to owner context. |
| `MinIO bucket` in backend-neutral artifact | `S3 bucket`. |
| `MinIO endpoint` in backend-neutral artifact | `object-store endpoint` or `S3-compatible endpoint`. |
| `minio-go` as runtime service proof | `generic S3-compatible client dependency`. |
| `MinIO` as legacy migration source | May remain only when explicitly labeled `migration_source` or `legacy_external_endpoint`. |

### 16.2 Occurrence inventory contract

The occurrence inventory MUST emit one artifact.

| Field | Required rule |
| --- | --- |
| `schema_id` | Exact `cartulary.seaweedfs_migration_occurrence_inventory.v1`. |
| `scanned_at` | RFC3339 UTC. |
| `repo_commit` | Required when run in repository. |
| `scan_scope.included_paths[]` | Exact included path globs. |
| `scan_scope.excluded_paths[]` | Exact excluded path globs. |
| `tokens[]` | Exact tokens searched. |
| `occurrences[]` | Sorted by path, line, column, token. |
| `occurrences[].classification` | `sdk_only`, `legacy_external_endpoint`, `migration_source`, `historical_changelog`, `invalid`, or `unclassified`. |
| `occurrences[].owner` | Required for every non-invalid retained occurrence. |
| `occurrences[].rationale` | Required. |
| `result` | `pass` only when no `invalid` and no `unclassified` occurrence exists. |

Scan scope defaults:

- Include tracked regular text files.
- Exclude `.git/`, dependency caches, generated build outputs, binary files, and retained test artifacts.
- Include checked-in generated files only when they are committed and not excluded by a generated-artifact policy.
- Search tokens exactly and case-sensitively as listed: `MinIO`, `minio`, `MINIO`, `minio_bucket`, `minio_endpoint`.

| Classification | Allowed content |
| --- | --- |
| `sdk_only` | `minio-go` dependency, license, SBOM, or adapter implementation note. |
| `legacy_external_endpoint` | Operator-supplied external S3-compatible endpoint caveat. |
| `migration_source` | Source backend in MinIO-to-SeaweedFS migration tooling, validation artifact, or fixture. |
| `historical_changelog` | Release-note or changelog reference to prior default. |
| `invalid` | Any default service, fixture, runtime, readiness, bucket, or operator instruction reference. |
| `unclassified` | Any occurrence not assigned by the reviewer or inventory tool. |

`invalid` and `unclassified` occurrences block release.

## 17. Source snapshot and fact blockers

This plan was revised from the uploaded migration plan and the uploaded Cartulary documentation set. It does not inspect the live repository, Compose files, Make targets, Go module files, lockfiles, current threat model, or current runtime code. Exact repository patches MUST inspect the target repository before claiming completion.

This plan does not independently revalidate external web facts about MinIO server, SeaweedFS, licenses, image tags, or upstream support posture. Release implementation MUST rely on live repository SBOM, license, and image-pin evidence rather than on this prose.

| Fact class | Required evidence | Blocking effect |
| --- | --- | --- |
| Live repository commit | `git rev-parse HEAD` or equivalent retained artifact. | Blocks repository completion claims. |
| Compose/service manifests inspected | Path list and line ranges for every default service definition. | Blocks runtime-profile completion. |
| Make targets inspected | Path list and command registry evidence. | Blocks acceptance-evidence matrix completion. |
| Go module and lockfiles inspected | Path list and dependency evidence. | Blocks `minio-go` and SBOM claims. |
| Frontend/browser route evidence inspected | Browser/E2E evidence path list. | Blocks direct-upload and handle-access completion. |
| SeaweedFS image tag/digest | Repo-control file plus registry digest evidence. | Blocks release. |
| License/SBOM evidence | Retained SBOM/license report. | Blocks release. |
| Threat model patch | Diff or retained scanner-ready document. | Blocks security acceptance. |
| Owner error registry patch | Owner diff plus generated error contract. | Blocks runtime outage mapping. |
| Owner config registry patch | Owner diff plus generated config-schema evidence. | Blocks production startup and config claims. |

## 18. Implementation sequence

| Phase | Gate to enter | Required work | Gate to exit |
| --- | --- | --- | --- |
| A. Contract cleanup | This revised plan accepted. | Patch owner docs, support docs, error registry, config registry, health readiness, and storage-ref blockers as required by §5. | Owner patches merged or blockers explicitly retained as non-claimable; no downstream row falsely marked complete. |
| B. Local service replacement | Phase A owner blockers needed for local profile resolved or explicitly not needed. | Replace MinIO server with SeaweedFS S3 in local development and default service definitions. | `seaweedfs-s3` starts with pinned image and §9 probe passes cleanly. |
| C. Harness replacement | Phase B complete. | Replace MinIO service vocabulary, artifacts, and service-backed fixture ownership. | Harness artifacts use canonical vocabulary and compatibility suite passes. |
| D. Adapter hardening | Phase C complete. | Implement adapter input contracts, operations, retry algorithm, capability probe, backend-neutral errors, and direct upload target probe. | Adapter tests and blocked/claimable runtime mappings are recorded correctly. |
| E. Backup and restore | Phase D complete. | Add SeaweedFS private manifest, shareable summary, restore verification, and backup/restore test coverage. | Fresh Postgres plus fresh SeaweedFS restore verification passes and emits §12 artifact. |
| F. Migration tooling | Phase E complete. | Implement application-stopped migration utility, state machine, copy semantics, validation artifact, and rollback docs. | Fixture migration emits §14 artifact with `result='pass'` and blocks mismatches. |
| G. Security and release gate | Phase F complete. | Update threat model, SBOM/license notes, release docs, occurrence inventory, and security scan. | Full acceptance-evidence matrix passes with no unresolved release-blocking row. |

No phase may be marked done by prose-only completion. Each exit gate requires retained evidence, command output, repository diff evidence, or explicit unresolved blocker metadata tied to §19.

## 19. Acceptance-evidence matrix

Every `SWFS-AC-*` row MUST be evaluated using this matrix. `TODO:repo-command-required` means the row is intentionally non-claimable until live repository inspection supplies the exact command or evidence source.

Current remediation evidence note: the 2026-06-04 summary-artifact and persisted-key remediation pass refreshed current SeaweedFS child-gate evidence at `.cartulary/test-results/20260604T232308Z-p3197473` and `.cartulary/release-artifacts/seaweedfs/20260604T233023Z-p3235166`. That run root contains passing `seaweedfs-compatibility/target-summary.json`, `seaweedfs-compatibility/tool-run-summary.json`, `seaweedfs-release-gate/target-summary.json`, and `seaweedfs-release-gate/tool-run-summary.json`. The same run root has `release-check/target-summary.json` with `status='fail'` because Vite exited before frontend readiness after hitting the host `ENOSPC` file-watcher limit. The current SeaweedFS child gate can support `SWFS-AC-015` and `SWFS-AC-018`; `SWFS-AC-024` remains unclaimable until a full aggregate `make release-check` passes.

| ID | Criterion | Closure class | Owner dependencies | Command or evidence source | Artifact schema | Pass predicate | Failure class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `SWFS-AC-001` | No default development, CI, service-backed test, or release Compose/manifest starts, pulls, or names a MinIO server container or image. | `repo_or_external_fact_required` | none | `TODO:repo-compose-inventory-command` | occurrence inventory | zero invalid MinIO server service rows | `blocks_release` |
| `SWFS-AC-002` | `github.com/minio/minio-go/v7`, if present, appears only as an S3 client dependency behind the object-store adapter. | `repo_or_external_fact_required` | none | `TODO:dependency-boundary-command` | SBOM/license report | no runtime service, fixture, readiness label, or operator instruction treats SDK as server support | `blocks_release` |
| `SWFS-AC-003` | Core 01 object-storage wording names SeaweedFS S3 as the default local/disconnected S3-compatible target while preserving generic S3 compatibility. | `blocked_until_owner_patch` | `SWFS-OWNER-DOCS-001` or core owner patch | owner diff | not applicable | target wording patched at owner anchor | `blocks_docs` |
| `SWFS-AC-004` | Core 04 disconnected deployment wording names one SeaweedFS S3 container or equivalent S3-compatible object store. | `blocked_until_owner_patch` | `SWFS-OWNER-DOCS-001` or core owner patch | owner diff | not applicable | target wording patched at owner anchor | `blocks_docs` |
| `SWFS-AC-005` | The default local service is named `seaweedfs-s3`, uses a pinned SeaweedFS image tag plus digest, and exposes only the S3 endpoint in ordinary local development. | `repo_or_external_fact_required` | none | `TODO:compose-service-validation-command` | capability probe | service name, digest, and exposure table match §6 and §7 | `blocks_phase` |
| `SWFS-AC-006` | Production documentation forbids default exposure of SeaweedFS admin, master, filer, volume, WebDAV, and debug surfaces. | `repo_or_external_fact_required` | `SWFS-OWNER-DOCS-001` | occurrence inventory plus release manifest scan | occurrence inventory | no invalid exposure instructions | `blocks_release` |
| `SWFS-AC-007` | The capability probe completes required PutObject, HeadObject, full GetObject, range GetObject, DeleteObject, CORS preflight, and presigned PUT stages within timeout and retry bounds. | `plan_local_closed` | `SWFS-OWNER-RANGE-001` if range owner declares required semantics | `TODO:capability-probe-command` | `cartulary.object_store_capability_probe.v1` | `result='pass'` and every required stage `status='pass'` | `blocks_phase` |
| `SWFS-AC-008` | In production profile, missing bucket, denied credentials, endpoint unreachable, CORS failure, or missing required capability fails startup before ready state. | `blocked_until_owner_patch` | `SWFS-OWNER-CONFIG-001`, `SWFS-OWNER-HEALTH-001` | `TODO:startup-failure-e2e-command` | probe artifact plus startup diagnostics | no listener becomes ready; diagnostic reason matches §9.3 | `blocks_release` |
| `SWFS-AC-009` | After a post-ready object-store outage, ordinary non-evidence workbook row editing remains available while evidence operations fail through mapped public dependency errors. | `blocked_until_owner_patch` | `SWFS-OWNER-ERR-001`, `SWFS-OWNER-HEALTH-001` | `TODO:runtime-outage-e2e-command` | public error contract evidence | non-evidence route succeeds; evidence route errors match §10.2 | `blocks_phase` |
| `SWFS-AC-010` | `POST /api/v1/object-blobs` still returns the Core-owned blob-slot response shape and timers. | `plan_local_closed` | none | `TODO:blob-slot-contract-command` | public route evidence | response includes required fields; timers unchanged | `blocks_phase` |
| `SWFS-AC-011` | Browser E2E creates a pending blob slot, uploads bytes to SeaweedFS, attaches blob to evidence, receives projection row, and emits collaboration update. | `repo_or_external_fact_required` | none | `TODO:browser-evidence-e2e-command` | browser/evidence artifact | full two-step flow succeeds without raw preview/download object URLs | `blocks_phase` |
| `SWFS-AC-012` | Preview and download issuance return only same-origin opaque evidence handles and never return bucket names, object keys, raw storage refs, raw SeaweedFS URLs, or long-lived object-store credentials. | `plan_local_closed` | none | `TODO:evidence-handle-e2e-command` | public route evidence | response fields are same-origin handle only; forbidden values absent | `blocks_phase` |
| `SWFS-AC-013` | Evidence negative cases for missing, pending, failed, quarantined, oversized, unsupported, expired, consumed, stale, and expired-upload-target states produce exact owner-mapped errors. | `blocked_until_owner_patch` | `SWFS-OWNER-ERR-001`, `SWFS-OWNER-RANGE-001` | `TODO:evidence-negative-matrix-command` | public error matrix | every case matches owner registry and §10 | `blocks_phase` |
| `SWFS-AC-014` | Harness artifacts use backend-neutral object-store vocabulary and contain no MinIO server readiness fields. | `repo_or_external_fact_required` | none | `TODO:harness-artifact-scan-command` | harness artifact summaries | forbidden fields absent; required fields present | `blocks_phase` |
| `SWFS-AC-015` | The SeaweedFS compatibility suite passes every `SWFS-COMP-*` case and contains no multipart or presigned-GET skip row. | `repo_or_external_fact_required` | none | `make seaweedfs-compatibility` | Current run-root `seaweedfs-compatibility/object-store-compatibility-report.json`, sibling passing `target-summary.json`, sibling passing `tool-run-summary.json`, and `cartulary.seaweedfs_compatibility_evidence.v1` | every case pass; no forbidden skip; report path is bound to the current `seaweedfs-compatibility` target run and not a stable release-artifact copy | `blocks_phase` |
| `SWFS-AC-016` | Each successful backup set against SeaweedFS includes a private manifest tied to the same backup set and consistency point as Postgres. | `blocked_until_owner_patch` | `SWFS-OWNER-BACKUP-001` | `TODO:backup-command` | `cartulary.object_store_backup_manifest.v1` | manifest valid; every object SHA-256 non-null | `blocks_release` |
| `SWFS-AC-017` | Restoring the latest successful retained backup into fresh Postgres and fresh SeaweedFS rebuilds projections and preserves blob lifecycle consistency. | `blocked_until_owner_patch` | `SWFS-OWNER-BACKUP-001` | `TODO:restore-verification-command` | `cartulary.restore_verification.v1` | `result='pass'` | `blocks_release` |
| `SWFS-AC-018` | Default MinIO-to-SeaweedFS migration preserves bucket name and object keys and does not mutate database `storage_ref` values. | `repo_or_external_fact_required` | none | `make backend-process` plus `make seaweedfs-release-gate` owner coverage | migration run, validation artifacts, redacted `cartulary.seaweedfs_migration_preservation_evidence.v2`, and storage-ref owner coverage artifact | database refs unchanged; copy ledger target key refs match source key refs; preservation summary has `bucket_preserved=true` | `blocks_phase` |
| `SWFS-AC-019` | Migration validation emits `cartulary.object_store_migration_validation.v1` with `result='pass'` only when blocking arrays are empty and every preview sample passes. | `plan_local_closed` | none | `TODO:migration-validation-command` | `cartulary.object_store_migration_validation.v1` | result computation matches §14.5 | `blocks_phase` |
| `SWFS-AC-020` | Any target-side object existing with a different size or SHA-256 than source blocks migration cutover. | `plan_local_closed` | none | `TODO:migration-target-mismatch-command` | copy ledger and validation artifact | mismatch produces blocking failure; no cutover | `blocks_phase` |
| `SWFS-AC-021` | Threat model update includes every STRIDE row listed in §15 and names SeaweedFS direct upload, credentials, admin surfaces, backup/restore, and migration validation. | `blocked_until_owner_patch` | `SWFS-OWNER-THREAT-001` | threat-model diff or scanner document | not applicable | every row covered with control and verification hook | `blocks_release` |
| `SWFS-AC-022` | Release SBOM and license gates identify no MinIO server artifact; if `minio-go` remains, release notes identify it as client dependency only. | `repo_or_external_fact_required` | none | `TODO:sbom-license-command` | SBOM/license report | no MinIO server; SDK-only classification for `minio-go` | `blocks_release` |
| `SWFS-AC-023` | Default docs no longer describe MinIO server as default local, disconnected, CI, service-backed test, or release-support object-store target. | `repo_or_external_fact_required` | `SWFS-OWNER-DOCS-001` | occurrence inventory | `cartulary.seaweedfs_migration_occurrence_inventory.v1` | zero invalid occurrences | `blocks_docs` |
| `SWFS-AC-024` | Full release gate runs required compatibility, object-store reachability, evidence, backup/restore, security, license/SBOM, and full repository check gates. | `repo_or_external_fact_required` | none | `make seaweedfs-release-gate` and `make release-check` | release gate summary, release-check target summary, current compatibility target run-root evidence, and SeaweedFS child `target-summary.json` artifacts | all child predicates pass, current `seaweedfs-compatibility` report provenance is verified, every sequence-produced SeaweedFS summary target retains `target-summary.json`, `release-check/target-summary.json` has `status='pass'`, and no unresolved release blockers remain | `blocks_release` |
| `SWFS-AC-025` | Post-migration occurrence inventory classifies every remaining MinIO token with zero invalid and zero unclassified rows. | `plan_local_closed` | none | `TODO:occurrence-inventory-command` | `cartulary.seaweedfs_migration_occurrence_inventory.v1` | `result='pass'` | `blocks_release` |

## 20. Revision acceptance criteria

The document revision is complete only when all criteria below pass.

| ID | Acceptance criterion |
| --- | --- |
| `SWFS-RP-AC-001` | Every section that depends on Core-owned behavior names a blocker or cites the owner patch that closed it. |
| `SWFS-RP-AC-002` | No `MAY` remains unless omitted behavior, allowed profiles, and conformance effect are explicit. |
| `SWFS-RP-AC-003` | Every runtime profile has endpoint, exposure, persistence, credential, bucket, CORS, and readiness behavior or an explicit blocker. |
| `SWFS-RP-AC-004` | Direct upload has one closed browser-reachability result per profile and no silent fallback to app-mediated upload. |
| `SWFS-RP-AC-005` | CORS behavior covers allowed origins, null origins, missing origins, methods, request headers, exposed headers, credentials, max age, and failure mapping. |
| `SWFS-RP-AC-006` | Adapter inputs, outputs, retry behavior, continuation tokens, list ordering, metadata, content type, checksum, and purpose vocabulary are fully defined. |
| `SWFS-RP-AC-007` | Probe stages, probe result schema, cleanup outcomes, and readiness effects are fully defined or owner-blocked. |
| `SWFS-RP-AC-008` | Backup manifest and shareable summary are separate schemas; restore input is operator-private and sufficient without external prose. |
| `SWFS-RP-AC-009` | Restore selection, incident selection, query selection, zero-incident behavior, and verification artifact schema are deterministic. |
| `SWFS-RP-AC-010` | Migration lifecycle has closed states, events, guards, transition actions, persistence schema, terminal-state rules, and rollback boundaries. |
| `SWFS-RP-AC-011` | Copy equivalence uses SHA-256 byte equality, not ETag reliability. Unsupported S3 features have fail-closed behavior. |
| `SWFS-RP-AC-012` | Migration validation artifact defines every child array schema, reason-code registry, canonical serialization, result computation, and preview-sample route check. |
| `SWFS-RP-AC-013` | Documentation occurrence inventory has a closed scan scope, artifact schema, classification vocabulary, owner field, and release-blocking predicate. |
| `SWFS-RP-AC-014` | Every `SWFS-AC-*` row maps to exact evidence, artifact schema, command or TODO blocker, and binary pass predicate. |
| `SWFS-RP-AC-015` | `## 17. Source snapshot and fact blockers` contains no passive caveats; each missing repo or external fact is a named blocker with required evidence. |
