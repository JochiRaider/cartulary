---
title: SeaweedFS S3 Migration Progress Handoff
document_class: implementation-support handoff
source_plan: docs/seaweedfs_s3_migration_implementation_plan.md
status: draft
snapshot_commit: 92fa79edbedc95c9c6ad51e666b107b09dc0755c
updated_at: 2026-06-03
---

# SeaweedFS S3 Migration Progress Handoff

## 0. Authority, objective, and scope

This handoff is an implementation-support progress artifact. It is subordinate to `docs/seaweedfs_s3_migration_implementation_plan.md` and MUST NOT be treated as a replacement for that implementation plan, Core 00 through Core 04, Core 05, or `docs/testing-harness-nlspec.md`.

The controlling migration objective remains: replace default Cartulary MinIO server usage with SeaweedFS S3-compatible object storage in development, test, disconnected, and release-support surfaces while preserving the object-store adapter boundary and public product behavior. The MinIO server and `github.com/minio/minio-go/v7` remain distinct migration subjects. The MinIO server MUST NOT remain a default runtime, harness, or release-support service after the migration. `minio-go` MAY remain only as a generic S3-compatible SDK dependency behind `internal/platform/objectstore` or an equivalent internal adapter.

This handoff uses the plan's phase structure, terminology, closure classes, acceptance criteria, and claimability model. Unknown repository state, owner-document status, external facts, command results, image digests, artifact identities, or validation outcomes are marked with `TODO:` and the evidence needed to close them.

## 1. Current repository snapshot

Planning inspection was performed against commit `92fa79edbedc95c9c6ad51e666b107b09dc0755c`. Phase A contract cleanup was implemented in the working tree against that same commit. At the start of Phase A implementation, `git status --short` showed this handoff file as untracked and no tracked file modifications.

The following are inspection facts only. They are not `SWFS-AC-*` acceptance evidence unless a later retained artifact or command result ties them to the acceptance-evidence matrix.

| Area | Inspected fact | Evidence |
| --- | --- | --- |
| Default local Compose service | `docker-compose.dev.yml` still defines service `minio`, image `minio/minio:RELEASE.2025-09-07T16-13-09Z`, console address `:9001`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, ports `9000` and `9001`, and volume `minio-data`. | `docker-compose.dev.yml:18` through `docker-compose.dev.yml:30`. |
| Test S3 fixture | `internal/testutil/s3test/s3test.go` still imports `github.com/minio/minio-go/v7`, uses `github.com/minio/minio-go/v7/pkg/credentials`, and defines `minioImage = "minio/minio:RELEASE.2025-09-07T16-13-09Z"`. | `internal/testutil/s3test/s3test.go:18` through `internal/testutil/s3test/s3test.go:29`. |
| Local dev service scripts | `scripts/dev-services.sh` still owns `wait_minio`, `init_minio`, starts `postgres minio`, initializes `MINIO_BUCKET`, and uses MinIO-specific reset wording. | `scripts/dev-services.sh:82` through `scripts/dev-services.sh:182`. |
| Harness public target registry | `docs/testing-harness-nlspec.md` still lists `minio-init` with command ID `cartulary.harness.command.minio_init.v1`. | `docs/testing-harness-nlspec.md:267`. |
| Generated harness manifests | `tools/task_surface_manifest.json` and `tools/execution_topology_manifest.json` still include `minio-init`, MinIO descriptions, and `cartulary.harness.command.minio_init.v1`. | Planning search output; TODO: retained occurrence inventory artifact with path, line, column, token, classification, owner, and rationale. |
| Testcontainers fixtures | `tools/testservices`, `internal/testutil/testcontainersx` tests, and harness recovery docs still contain MinIO service, startup, bucket, and image references. | Planning search output; TODO: retained occurrence inventory artifact with complete scan scope. |
| `minio-go` dependency | `go.mod` contains `github.com/minio/minio-go/v7 v7.0.100`; `go.sum` contains matching module checksums. | `go.mod:12`, `go.sum:91`, `go.sum:92`. |
| SeaweedFS references | During planning search, SeaweedFS references were found only in `docs/seaweedfs_s3_migration_implementation_plan.md`. | TODO: retained occurrence inventory artifact that proves complete scan scope and exact results. |
| SeaweedFS image tag and digest | No repo-controlled SeaweedFS image tag plus digest was verified. | TODO: repo-controlled service manifest plus registry digest evidence. |
| Owner-document patches | Phase A owner diffs were added for public object-store dependency errors, backend-neutral managed object-store service binding, readiness state, range fallback semantics, backup/restore artifact adoption, threat-model coverage, and default SeaweedFS S3 wording. `SWFS-OWNER-STORAGEREF-001` remains an explicit owner-decision TODO at the Core 01 storage-ref anchor. | `docs/spec/01_architecture_storage_and_view_contracts.md`; `docs/spec/04_security_deployment_and_conformance.md`; `contracts/errors/index.json`; generated contract embeddings from `make generate`. |
| Acceptance command results | No `SWFS-AC-*` acceptance command was run or retained for this handoff. Phase A validation commands ran only owner/doc/contract hygiene checks. | See Section 5.2. |

## 2. Phase-by-phase progress status

No phase may be marked complete by prose-only completion. Each exit gate requires retained evidence, command output, repository diff evidence, or explicit unresolved blocker metadata tied to the implementation plan's acceptance-evidence matrix.

| Phase | Status | Current handoff state | Evidence needed to close |
| --- | --- | --- | --- |
| A. Contract cleanup | `complete_for_phase_a_with_retained_blockers` | Phase A owner cleanup is implemented in the working tree. Seven owner rows have local owner diffs or authored registry changes; `SWFS-OWNER-STORAGEREF-001` remains blocked with an explicit Core 01 TODO blocker rather than invented behavior. Downstream acceptance rows remain unclaimed. | Commit or review the Phase A diffs, then carry remaining implementation evidence into later phases. |
| B. Local service replacement | `blocked` | Incomplete. The inspected repo still defines a default `minio` service and `minio/minio` image in `docker-compose.dev.yml`. SeaweedFS image tag/digest evidence remains unresolved. | Repo-controlled `seaweedfs-s3` service with pinned SeaweedFS image tag plus digest, ordinary local exposure limited to the S3 endpoint, and clean Section 9 capability probe. |
| C. Harness replacement | `blocked` | Incomplete. Harness docs/manifests, testservice code, and fixture terminology still include MinIO surfaces. | Backend-neutral object-store harness vocabulary, generated manifest updates from owner inputs, service-backed fixture replacement, and retained artifact scan proving forbidden MinIO readiness fields are absent. |
| D. Adapter hardening | `TODO:evidence-required` | Not verified. Existing adapter still uses `minio-go`; this is allowed only as SDK-only if boundary and artifact evidence prove it is not MinIO server support. No adapter-hardening tests or probe artifacts were retained. | Adapter tests, capability probe artifact, direct-upload probe evidence, retry/error mapping evidence, and owner-blocked public mapping status. |
| E. Backup and restore | `TODO:evidence-required` | Not verified. Phase A added owner text for backup/restore artifact shapes, but no SeaweedFS backup manifest, shareable summary, restore verification artifact, or tooling evidence was produced. | `cartulary.object_store_backup_manifest.v1`, `cartulary.object_store_backup_summary.v1` if emitted, and `cartulary.restore_verification.v1` with `result='pass'`. |
| F. Migration tooling | `TODO:evidence-required` | Not verified. No application-stopped migration utility, lifecycle artifact, copy ledger, validation artifact, or rollback evidence was verified. | `cartulary.object_store_migration_run.v1`, copy ledger evidence, `cartulary.object_store_migration_validation.v1`, storage-ref owner citation, mismatch-blocking fixture, and rollback documentation evidence. |
| G. Security and release gate | `blocked` | Blocked. Phase A added threat-model owner text, but occurrence inventory, scanner/release evidence, SBOM/license, release-gate, and remaining owner dependency evidence are not available. | Occurrence inventory with zero invalid and zero unclassified rows, SBOM/license report, release manifest exposure scan, `make release-check` or plan-owned release gate evidence after prior blockers resolve. |

## 3. Completed work

Phase A contract cleanup completed only owner-document, authored registry, generated contract, support-doc, and handoff work. No Phase B or later implementation was performed.

Completed Phase A decisions and edits:

- `SWFS-OWNER-ERR-001`: Core 01 now defines `object_store_unavailable` and `object_store_access_rejected`, transport status, retry hints, reason-code vocabularies, and route-family applicability. `contracts/errors/index.json` mirrors those codes, and `make generate` updated generated contract embeddings.
- `SWFS-OWNER-CONFIG-001`: Core 04 now keeps object storage backend-neutral under `roots.object_storage`, defines managed-service `service_ref` environment binding, defaults, omitted/null behavior, redaction, validation reasons, and fail-closed readiness behavior without adding a SeaweedFS config namespace.
- `SWFS-OWNER-HEALTH-001`: Core 04 now distinguishes `/healthz` process liveness from structured `/readyz` object-store readiness states including `degraded_object_store` and recovery behavior.
- `SWFS-OWNER-STORAGEREF-001`: Core 01 now contains an explicit TODO blocker at the `evidence.storage_ref` anchor; no storage-ref grammar or key behavior was invented.
- `SWFS-OWNER-RANGE-001`: Core 01 now requires byte-range support for preview-handle redemption and permits full-stream download fallback for ordinary download-handle redemption.
- `SWFS-OWNER-BACKUP-001`: Core 01 and Core 04 now adopt the SeaweedFS S3 operator-private backup manifest and restore-verification artifact shapes as sufficient owner-level evidence shapes when tied to the same backup set and consistency point with non-null object SHA-256 proofs and redaction boundaries.
- `SWFS-OWNER-THREAT-001`: Core 04 §4.4 STRIDE coverage now names SeaweedFS S3 endpoint identity, direct upload target scope, object tampering, migration mismatch, credentials, admin surfaces, CORS, storage exhaustion, backup/restore, and migration validation risks.
- `SWFS-OWNER-DOCS-001`: Authored README and guide default wording now uses SeaweedFS S3-compatible or backend-neutral S3-compatible object-store terminology. Remaining SDK-only and current scheduler/generated-surface references are retained as blockers for later phases.

Planning inspection also identified current blocking MinIO server surfaces in local Compose, local service scripts, service-backed harness/testservice fixtures, generated harness manifests, harness docs, testcontainer fixtures, and dependency files. Those inspection results are evidence of remaining work, not completion evidence for any `SWFS-AC-*` row.

## 4. Remaining work and blockers

### 4.1 Owner-document patches still required

| Owner patch ID | Target document or owner area | Required evidence to close | Current status |
| --- | --- | --- | --- |
| `SWFS-OWNER-ERR-001` | Core 01 public success/error envelope and error-code registry | Owner diff plus generated error contract artifact for `object_store_unavailable` and `object_store_access_rejected`. | `resolved_in_phase_a_worktree`; evidence: Core 01 error rows and reason registries, `contracts/errors/index.json`, `internal/gen/contracts/contracts_gen.go`, `packages/protocol-ts/src/generated/contracts.ts`, `make generate` run root `.cartulary/test-results/20260603T184912Z-p563803`. |
| `SWFS-OWNER-CONFIG-001` | Core 04 deployment-configuration contract | Owner diff plus config-schema drift artifact for backend-neutral object-store configuration keys and fail-closed validation semantics. | `resolved_in_phase_a_worktree_for_owner_text`; evidence: Core 04 managed `roots.object_storage` service-ref binding and validation reasons. Runtime implementation and config-schema conformance evidence remain TODO for `SWFS-AC-008` and `SWFS-AC-024`. |
| `SWFS-OWNER-HEALTH-001` | Core 01 or Core 04 health/readiness owner section | Owner diff plus health-route or readiness artifact evidence for `degraded_object_store` and recovery output. | `resolved_in_phase_a_worktree_for_owner_text`; evidence: Core 04 `/healthz` and `/readyz` structured status wording. Health-route implementation evidence remains TODO. |
| `SWFS-OWNER-STORAGEREF-001` | Core 01 and generated contracts for object blob storage-ref ownership | Owner diff or verified existing owner citation for storage-ref grammar, canonicalization, key generation, maximum length, and invalid-state behavior. | `blocked_with_explicit_todo`; Core 01 storage-ref anchor now records the exact owner decision required. No grammar, canonicalization, or key-generation behavior was invented. |
| `SWFS-OWNER-RANGE-001` | Core 01 evidence preview/download owner section | Owner diff plus browser/backend evidence-access tests defining required range retrieval and full-download fallback behavior. | `resolved_in_phase_a_worktree_for_owner_text`; evidence: Core 01 §16 preview byte-range requirement and full-download fallback wording. Evidence-access tests remain TODO. |
| `SWFS-OWNER-BACKUP-001` | Core 01 and Core 04 backup/restore owner sections | Owner diff plus backup/restore retained artifacts adopting the Section 12 operator-private manifest and restore-verification schemas. | `resolved_in_phase_a_worktree_for_owner_text`; evidence: Core 01 backup and restore verification artifact adoption plus Core 04 recovery-artifact confidentiality wording. Backup/restore tool artifacts remain TODO. |
| `SWFS-OWNER-THREAT-001` | Project threat model path | Threat-model diff or scanner-ready retained document covering the Section 15 STRIDE rows. | `resolved_in_phase_a_worktree_for_owner_text`; project threat model path is Core 04 §4.4, and STRIDE rows now include SeaweedFS S3-specific coverage. Scanner/release evidence remains TODO. |
| `SWFS-OWNER-DOCS-001` | `cartulary-dev-guide.md`, `cartulary_repository_bootstrap_guide.md`, and docs index | Occurrence inventory plus docs diffs replacing MinIO server default wording with SeaweedFS S3 or generic S3-compatible wording while preserving SDK-only `minio-go` where present. | `partially_resolved_with_retained_blockers`; authored README and guide default wording patched. Complete occurrence inventory, generated/harness docs cleanup, scheduler-resource vocabulary, and service replacement evidence remain TODO. |

### 4.2 Repo or external facts still required

| Fact class | Required evidence | Current status |
| --- | --- | --- |
| Live repository commit | `git rev-parse HEAD` or equivalent retained artifact. | Planning value recorded as `92fa79edbedc95c9c6ad51e666b107b09dc0755c`; TODO: retained artifact for completion claims. |
| Compose/service manifests inspected | Path list and line ranges for every default service definition. | Partial inspection found `docker-compose.dev.yml` MinIO service; TODO: complete manifest inventory. |
| Make targets inspected | Path list and command registry evidence. | `make help-all` was consulted during planning; TODO: retained command-surface artifact if used as acceptance evidence. |
| Go module and lockfiles inspected | Path list and dependency evidence. | `go.mod` and `go.sum` show `minio-go`; TODO: SBOM/license evidence and dependency-boundary scan. |
| Frontend/browser route evidence inspected | Browser/E2E evidence path list. | TODO: browser route inventory and evidence-flow tests. |
| SeaweedFS image tag/digest | Repo-control file plus registry digest evidence. | TODO: pinned image tag plus digest. |
| License/SBOM evidence | Retained SBOM/license report. | TODO: release SBOM/license artifact. |
| Threat model patch | Diff or retained scanner-ready document. | Core 04 §4.4 owner text patched; TODO: scanner/release evidence. |
| Owner error registry patch | Owner diff plus generated error contract. | Core 01 and generated error contracts patched; TODO: review/commit evidence if used for acceptance. |
| Owner config registry patch | Owner diff plus generated config-schema evidence. | Core 04 owner text patched; TODO: runtime config-schema and startup evidence. |

### 4.3 Unresolved evidence requirements

- TODO: `cartulary.seaweedfs_migration_occurrence_inventory.v1` over the full plan-defined scan scope.
- TODO: `cartulary.object_store_capability_probe.v1` with `result='pass'` and clean cleanup classification.
- TODO: backend-neutral harness artifact summaries with forbidden MinIO fields absent.
- TODO: SeaweedFS compatibility report covering every `SWFS-COMP-*` case with no multipart or presigned-GET skip row.
- TODO: public route evidence for blob-slot shape, timers, same-origin evidence handles, and negative evidence-state matrix.
- TODO: backup/restore manifests and restore verification artifacts.
- TODO: migration run, copy ledger, validation, mismatch-blocking, and rollback artifacts.
- TODO: SBOM/license, release gate, and later-phase `make agent-finalize` retained evidence after implementation work.

## 5. Validation status

### 5.1 Planning inspection commands run

These commands were used to prepare this handoff. They are not acceptance evidence for any `SWFS-AC-*` row.

| Command | Result summary |
| --- | --- |
| `sed -n '1,260p' docs/seaweedfs_s3_migration_implementation_plan.md` and later ranges through the end of the file | Read controlling implementation plan, phases, owner patch registry, acceptance matrix, and claimability model. |
| `sed -n '1,220p' docs/testing-harness-nlspec.md` and targeted line inspection | Confirmed Make-owned harness authority and current `minio-init` public target row. |
| `sed -n '1,220p' docs/domain.md` | Confirmed domain vocabulary boundary and implementation-support treatment. |
| `sed -n '1,220p' docs/spec/00_document_set_status_and_precedence.md` | Confirmed Core 00 authority and owner-document precedence. |
| `git rev-parse HEAD` | Returned `92fa79edbedc95c9c6ad51e666b107b09dc0755c`. |
| `git status --short` | Returned no output before this handoff file was added. |
| `rg -n "SeaweedFS\|seaweedfs\|MinIO server\|minio/minio\|minio-go\|minio_" ...` | Found current MinIO server, harness, fixture, docs, manifest, and SDK references; SeaweedFS was only found in the implementation plan during the planning search. |
| `make help-all | rg -i 'seaweed\|object\|s3\|minio\|backup\|restore\|release-check\|agent-finalize\|harness\|browser-e2e\|test-fast\|check\|generate\|drift\|sbom\|license\|gosec\|vuln'` | Showed current public command surface still includes `db-up`, `services-up`, and `minio-init` MinIO wording. |

### 5.2 Phase A verification commands

The commands below verify only Phase A owner/doc/contract hygiene. They are not acceptance evidence for any `SWFS-AC-*` row.

| Command | Status | Run root |
| --- | --- | --- |
| `make generate` | `pass`; regenerated contract embeddings from authored `contracts/errors/index.json`. | `.cartulary/test-results/20260603T184912Z-p563803` |
| `make generated-artifact-policy-check` | `pass`. | `.cartulary/test-results/20260603T184946Z-p565502` |
| `make json-shape-check` | `pass`. | `.cartulary/test-results/20260603T184951Z-p565808` |
| `make generate-drift` | `pass`; generated outputs match owner inputs after regeneration. | `.cartulary/test-results/20260603T184956Z-p566103` |
| `make agent-finalize` | `pass`; generated outputs unchanged and no retained full-check `RESULTS_DIR` was supplied. | `.cartulary/test-results/20260603T185324Z-p571651` |

Skipped checks:

- `make harness-contract` was skipped because Phase A did not edit `docs/testing-harness-nlspec.md`, harness command topology, generated harness manifests, or harness public command surfaces.
- Broad `make check` was skipped because Phase A touched owner docs, authored contract registry input, generated contract embeddings, support docs, and this handoff only; no implementation behavior or service/harness replacement was changed.

### 5.3 Acceptance command backlog

All plan-owned acceptance commands below remain unrun and all rows remain unclaimed unless a later retained artifact supplies the required evidence.

| Command or evidence source | Status |
| --- | --- |
| `TODO:repo-compose-inventory-command` | `TODO: unrun; required for SWFS-AC-001`. |
| `TODO:dependency-boundary-command` | `TODO: unrun; required for SWFS-AC-002`. |
| owner diff for `SWFS-AC-003` and `SWFS-AC-004` | `TODO: working-tree owner wording exists, but acceptance completion is not claimed until reviewed/retained as evidence`. |
| `TODO:compose-service-validation-command` | `TODO: unrun; required for SWFS-AC-005`. |
| occurrence inventory plus release manifest scan | `TODO: unrun; required for SWFS-AC-006`. |
| `TODO:capability-probe-command` | `TODO: unrun; required for SWFS-AC-007`. |
| `TODO:startup-failure-e2e-command` | `TODO: unrun; required for SWFS-AC-008`. |
| `TODO:runtime-outage-e2e-command` | `TODO: unrun; required for SWFS-AC-009`. |
| `TODO:blob-slot-contract-command` | `TODO: unrun; required for SWFS-AC-010`. |
| `TODO:browser-evidence-e2e-command` | `TODO: unrun; required for SWFS-AC-011`. |
| `TODO:evidence-handle-e2e-command` | `TODO: unrun; required for SWFS-AC-012`. |
| `TODO:evidence-negative-matrix-command` | `TODO: unrun; required for SWFS-AC-013`. |
| `TODO:harness-artifact-scan-command` | `TODO: unrun; required for SWFS-AC-014`. |
| `TODO:seaweedfs-compat-command` | `TODO: unrun; required for SWFS-AC-015`. |
| `TODO:backup-command` | `TODO: unrun; required for SWFS-AC-016`. |
| `TODO:restore-verification-command` | `TODO: unrun; required for SWFS-AC-017`. |
| `TODO:migration-fixture-command` | TODO: unrun; blocked by `SWFS-OWNER-STORAGEREF-001`; required for `SWFS-AC-018`. |
| `TODO:migration-validation-command` | `TODO: unrun; required for SWFS-AC-019`. |
| `TODO:migration-target-mismatch-command` | `TODO: unrun; required for SWFS-AC-020`. |
| threat-model diff or scanner document | `TODO: owner text patched in Core 04 §4.4; scanner/release evidence unrun; required for SWFS-AC-021`. |
| `TODO:sbom-license-command` | `TODO: unrun; required for SWFS-AC-022`. |
| occurrence inventory | `TODO: unrun; required for SWFS-AC-023 and SWFS-AC-025`. |
| `TODO:release-gate-command` | `TODO: unrun; required for SWFS-AC-024`. |

## 6. Acceptance criteria status

Every `SWFS-AC-*` row remains non-claimable in this handoff unless the current status explicitly names retained evidence. No row currently names retained pass evidence.

| ID | Criterion | Closure class | Owner dependencies | Command or evidence source | Artifact schema | Pass predicate | Failure class | Current status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `SWFS-AC-001` | No default development, CI, service-backed test, or release Compose/manifest starts, pulls, or names a MinIO server container or image. | `repo_or_external_fact_required` | none | `TODO:repo-compose-inventory-command` | occurrence inventory | zero invalid MinIO server service rows | `blocks_release` | `not_claimable`; planning inspection found current MinIO service/image in `docker-compose.dev.yml`; TODO: complete inventory artifact. |
| `SWFS-AC-002` | `github.com/minio/minio-go/v7`, if present, appears only as an S3 client dependency behind the object-store adapter. | `repo_or_external_fact_required` | none | `TODO:dependency-boundary-command` | SBOM/license report | no runtime service, fixture, readiness label, or operator instruction treats SDK as server support | `blocks_release` | `not_claimable`; `minio-go` is present in `go.mod`; TODO: dependency-boundary and SBOM/license report. |
| `SWFS-AC-003` | Core 01 object-storage wording names SeaweedFS S3 as the default local/disconnected S3-compatible target while preserving generic S3 compatibility. | `blocked_until_owner_patch` | `SWFS-OWNER-DOCS-001` or core owner patch | owner diff | not applicable | target wording patched at owner anchor | `blocks_docs` | `blocked`; working-tree Core 01 wording exists, but no acceptance completion is claimed until retained owner-diff evidence is reviewed. |
| `SWFS-AC-004` | Core 04 disconnected deployment wording names one SeaweedFS S3 container or equivalent S3-compatible object store. | `blocked_until_owner_patch` | `SWFS-OWNER-DOCS-001` or core owner patch | owner diff | not applicable | target wording patched at owner anchor | `blocks_docs` | `blocked`; working-tree Core 04 wording exists, but no acceptance completion is claimed until retained owner-diff evidence is reviewed. |
| `SWFS-AC-005` | The default local service is named `seaweedfs-s3`, uses a pinned SeaweedFS image tag plus digest, and exposes only the S3 endpoint in ordinary local development. | `repo_or_external_fact_required` | none | `TODO:compose-service-validation-command` | capability probe | service name, digest, and exposure table match Sections 6 and 7 | `blocks_phase` | `not_claimable`; current Compose service is `minio`; TODO: pinned SeaweedFS service and probe evidence. |
| `SWFS-AC-006` | Production documentation forbids default exposure of SeaweedFS admin, master, filer, volume, WebDAV, and debug surfaces. | `repo_or_external_fact_required` | `SWFS-OWNER-DOCS-001` | occurrence inventory plus release manifest scan | occurrence inventory | no invalid exposure instructions | `blocks_release` | `blocked`; Phase A touched default wording only; TODO: occurrence inventory and release manifest scan. |
| `SWFS-AC-007` | The capability probe completes required PutObject, HeadObject, full GetObject, range GetObject, DeleteObject, CORS preflight, and presigned PUT stages within timeout and retry bounds. | `plan_local_closed` | `SWFS-OWNER-RANGE-001` if range owner declares required semantics | `TODO:capability-probe-command` | `cartulary.object_store_capability_probe.v1` | `result='pass'` and every required stage `status='pass'` | `blocks_phase` | `not_claimable`; TODO: capability probe command and artifact; range owner dependency may apply. |
| `SWFS-AC-008` | In production profile, missing bucket, denied credentials, endpoint unreachable, CORS failure, or missing required capability fails startup before ready state. | `blocked_until_owner_patch` | `SWFS-OWNER-CONFIG-001`, `SWFS-OWNER-HEALTH-001` | `TODO:startup-failure-e2e-command` | probe artifact plus startup diagnostics | no listener becomes ready; diagnostic reason matches Section 9.3 | `blocks_release` | `blocked`; owner text patched in working tree; TODO: startup failure E2E artifact. |
| `SWFS-AC-009` | After a post-ready object-store outage, ordinary non-evidence workbook row editing remains available while evidence operations fail through mapped public dependency errors. | `blocked_until_owner_patch` | `SWFS-OWNER-ERR-001`, `SWFS-OWNER-HEALTH-001` | `TODO:runtime-outage-e2e-command` | public error contract evidence | non-evidence route succeeds; evidence route errors match Section 10.2 | `blocks_phase` | `blocked`; owner text and generated error contracts patched in working tree; TODO: runtime outage E2E artifact. |
| `SWFS-AC-010` | `POST /api/v1/object-blobs` still returns the Core-owned blob-slot response shape and timers. | `plan_local_closed` | none | `TODO:blob-slot-contract-command` | public route evidence | response includes required fields; timers unchanged | `blocks_phase` | `not_claimable`; TODO: blob-slot contract command and public route evidence. |
| `SWFS-AC-011` | Browser E2E creates a pending blob slot, uploads bytes to SeaweedFS, attaches blob to evidence, receives projection row, and emits collaboration update. | `repo_or_external_fact_required` | none | `TODO:browser-evidence-e2e-command` | browser/evidence artifact | full two-step flow succeeds without raw preview/download object URLs | `blocks_phase` | `not_claimable`; TODO: browser evidence E2E artifact. |
| `SWFS-AC-012` | Preview and download issuance return only same-origin opaque evidence handles and never return bucket names, object keys, raw storage refs, raw SeaweedFS URLs, or long-lived object-store credentials. | `plan_local_closed` | none | `TODO:evidence-handle-e2e-command` | public route evidence | response fields are same-origin handle only; forbidden values absent | `blocks_phase` | `not_claimable`; TODO: evidence-handle E2E command and public route evidence. |
| `SWFS-AC-013` | Evidence negative cases for missing, pending, failed, quarantined, oversized, unsupported, expired, consumed, stale, and expired-upload-target states produce exact owner-mapped errors. | `blocked_until_owner_patch` | `SWFS-OWNER-ERR-001`, `SWFS-OWNER-RANGE-001` | `TODO:evidence-negative-matrix-command` | public error matrix | every case matches owner registry and Section 10 | `blocks_phase` | `blocked`; owner text and generated error contracts patched in working tree; TODO: negative matrix artifact. |
| `SWFS-AC-014` | Harness artifacts use backend-neutral object-store vocabulary and contain no MinIO server readiness fields. | `repo_or_external_fact_required` | none | `TODO:harness-artifact-scan-command` | harness artifact summaries | forbidden fields absent; required fields present | `blocks_phase` | `not_claimable`; current harness surfaces still include MinIO vocabulary; TODO: artifact scan. |
| `SWFS-AC-015` | The SeaweedFS compatibility suite passes every `SWFS-COMP-*` case and contains no multipart or presigned-GET skip row. | `repo_or_external_fact_required` | none | `TODO:seaweedfs-compat-command` | compatibility report | every case pass; no forbidden skip | `blocks_phase` | `not_claimable`; TODO: SeaweedFS compatibility suite and report. |
| `SWFS-AC-016` | Each successful backup set against SeaweedFS includes a private manifest tied to the same backup set and consistency point as Postgres. | `blocked_until_owner_patch` | `SWFS-OWNER-BACKUP-001` | `TODO:backup-command` | `cartulary.object_store_backup_manifest.v1` | manifest valid; every object SHA-256 non-null | `blocks_release` | `blocked`; owner text patched in working tree; TODO: backup manifest artifact. |
| `SWFS-AC-017` | Restoring the latest successful retained backup into fresh Postgres and fresh SeaweedFS rebuilds projections and preserves blob lifecycle consistency. | `blocked_until_owner_patch` | `SWFS-OWNER-BACKUP-001` | `TODO:restore-verification-command` | `cartulary.restore_verification.v1` | `result='pass'` | `blocks_release` | `blocked`; owner text patched in working tree; TODO: restore verification artifact. |
| `SWFS-AC-018` | Default MinIO-to-SeaweedFS migration preserves bucket name and object keys and does not mutate database `storage_ref` values. | `blocked_until_owner_patch` | `SWFS-OWNER-STORAGEREF-001` | `TODO:migration-fixture-command` | migration run and validation artifacts | database refs unchanged; copy ledger target keys match source | `blocks_phase` | `blocked`; TODO: storage-ref owner citation/patch and migration fixture evidence. |
| `SWFS-AC-019` | Migration validation emits `cartulary.object_store_migration_validation.v1` with `result='pass'` only when blocking arrays are empty and every preview sample passes. | `plan_local_closed` | none | `TODO:migration-validation-command` | `cartulary.object_store_migration_validation.v1` | result computation matches Section 14.5 | `blocks_phase` | `not_claimable`; TODO: migration validation command and artifact. |
| `SWFS-AC-020` | Any target-side object existing with a different size or SHA-256 than source blocks migration cutover. | `plan_local_closed` | none | `TODO:migration-target-mismatch-command` | copy ledger and validation artifact | mismatch produces blocking failure; no cutover | `blocks_phase` | `not_claimable`; TODO: mismatch fixture and artifacts. |
| `SWFS-AC-021` | Threat model update includes every STRIDE row listed in Section 15 and names SeaweedFS direct upload, credentials, admin surfaces, backup/restore, and migration validation. | `blocked_until_owner_patch` | `SWFS-OWNER-THREAT-001` | threat-model diff or scanner document | not applicable | every row covered with control and verification hook | `blocks_release` | `blocked`; Core 04 §4.4 owner text patched in working tree; TODO: scanner/release evidence. |
| `SWFS-AC-022` | Release SBOM and license gates identify no MinIO server artifact; if `minio-go` remains, release notes identify it as client dependency only. | `repo_or_external_fact_required` | none | `TODO:sbom-license-command` | SBOM/license report | no MinIO server; SDK-only classification for `minio-go` | `blocks_release` | `not_claimable`; TODO: SBOM/license report and release notes check. |
| `SWFS-AC-023` | Default docs no longer describe MinIO server as default local, disconnected, CI, service-backed test, or release-support object-store target. | `repo_or_external_fact_required` | `SWFS-OWNER-DOCS-001` | occurrence inventory | `cartulary.seaweedfs_migration_occurrence_inventory.v1` | zero invalid occurrences | `blocks_docs` | `blocked`; authored default wording patched, but current generated/harness/service surfaces and complete occurrence inventory remain TODO. |
| `SWFS-AC-024` | Full release gate runs required compatibility, object-store reachability, evidence, backup/restore, security, license/SBOM, and full repository check gates. | `repo_or_external_fact_required` | owner blockers resolved | `TODO:release-gate-command` | release gate summary | all child predicates pass and no unresolved release blockers remain | `blocks_release` | `blocked`; TODO: all prior blockers plus release-gate summary. |
| `SWFS-AC-025` | Post-migration occurrence inventory classifies every remaining MinIO token with zero invalid and zero unclassified rows. | `plan_local_closed` | none | `TODO:occurrence-inventory-command` | `cartulary.seaweedfs_migration_occurrence_inventory.v1` | `result='pass'` | `blocks_release` | `not_claimable`; TODO: occurrence inventory artifact. |

## 7. Risks, assumptions, and next actions

### 7.1 Risks and assumptions

- Current default local and service-backed surfaces still appear MinIO-centered. Treat Phase B and Phase C as blocked until the service replacement and harness replacement are implemented from owner inputs and regenerated surfaces where applicable.
- `minio-go` presence is not itself invalid, but any claim that it is SDK-only is currently unverified.
- Runtime implementation evidence remains missing for dependency errors, readiness shape, production object-store configuration, range behavior, backup/restore, and threat-model release checks. Storage-ref grammar and key-generation behavior remain owner-blocked by `SWFS-OWNER-STORAGEREF-001`.
- The implementation plan requires byte-equivalence proof by SHA-256. Do not use ETag-only evidence for migration equivalence.
- The handoff's planning inspection used local command output that is summarized here. Release or acceptance claims require retained artifacts, not this prose summary.

### 7.2 Recommended next actions

1. Review and commit the Phase A owner, registry, generated-contract, support-doc, and handoff diffs if accepted.
2. Resolve `SWFS-OWNER-STORAGEREF-001` with an owner decision for storage-ref grammar, canonicalization, key generation, maximum length, and invalid-state behavior.
3. Build `cartulary.seaweedfs_migration_occurrence_inventory.v1` and use it to classify all MinIO tokens before editing broad service and docs surfaces.
4. Replace the default local service with `seaweedfs-s3`, pin the SeaweedFS image tag plus digest, remove ordinary local admin-surface exposure, and add a clean Section 9 capability probe.
5. Replace harness/testservice MinIO vocabulary and service ownership with backend-neutral object-store vocabulary and SeaweedFS S3 fixtures, updating generated manifests from owner inputs rather than hand-editing generated outputs.
6. Harden the object-store adapter and probe path while preserving public blob-slot, attach, preview, download, authorization, and evidence lifecycle semantics.
7. Implement backup/restore and migration tooling only after the owner dependencies and adapter/probe evidence are in place.
8. Finish with SBOM/license, occurrence inventory, release manifest exposure scan, `make agent-finalize`, and the full release gate after all release blockers are closed.
