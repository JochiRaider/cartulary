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

Planning inspection was performed against commit `92fa79edbedc95c9c6ad51e666b107b09dc0755c`. Phase A contract cleanup and Phase B local service replacement were implemented in the working tree against that same commit. Phase C harness replacement was implemented in the working tree at live repository commit `92e679dcc8031e476e5586f0b11adfcaa93647d1`. Phase D adapter hardening was implemented in the current working tree without changing generated contracts, lockfiles, database schema, or public route families.

The following are inspection facts only. They are not `SWFS-AC-*` acceptance evidence unless a later retained artifact or command result ties them to the acceptance-evidence matrix.

| Area | Inspected fact | Evidence |
| --- | --- | --- |
| Default local Compose service | Phase B replaces the default local object-store service with `seaweedfs-s3`, pinned to `docker.io/chrislusf/seaweedfs:4.17@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`. The only host-published object-store port is S3 port `8333`. | `docker-compose.dev.yml`; `make services-up` run root `.cartulary/test-results/20260603T195218Z-p648759`; direct post-run `docker compose ps` inspection showed Postgres plus `seaweedfs-s3` only and host-published object-store port `0.0.0.0:8333`. |
| Test S3 fixture | Phase C replaces the owned testcontainer server with SeaweedFS S3 pinned to `docker.io/chrislusf/seaweedfs:4.17@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`, exposing S3 port `8333/tcp` and using the Phase B local credentials. `github.com/minio/minio-go/v7` remains only as the generic S3 SDK in the fixture client/probe code. | `internal/testutil/s3test/s3test.go`; `make backend-integration` run root `.cartulary/test-results/20260603T204514Z-p691818`; `backend-integration/backend-integration-testutil/phase-summary.json`; `_shared/backend-integration-testutil-shard-02/runner.jsonl`. |
| Local dev service scripts | Phase B replaces local wait/init/reset behavior with provider-neutral object-store helpers backed by `tools/objectstoreprobe`. `make services-up`, `make db-up`, `make object-store-init`, and `make object-store-reset` now use `seaweedfs-s3` and `OBJECT_STORE_BUCKET`. | `scripts/dev-services.sh`; `scripts/dev-stack.sh`; `make services-up` run root `.cartulary/test-results/20260603T195218Z-p648759`; `make object-store-init` run root `.cartulary/test-results/20260603T195740Z-p652397`. |
| Harness public target registry | Phase B replaces local `minio-init` with `object-store-init` and generated `object-store-wait` from owner inputs. | `docs/testing-harness-nlspec.md:267`; `tools/execution_topology_manifest.json`; `make phase-schedules` run root `.cartulary/test-results/20260603T194856Z-p641032`. |
| Generated harness manifests | Generated task-surface and scheduler outputs now include backend-neutral `object_store` service-backed resource claims. Generated scheduler and phase-ledger outputs were regenerated only through Make-owned generators. | `tools/scheduler_manifest.json`; `tools/execution_topology_render_index.json`; `docs/testing/phase*_coverage_ledger.md`; `make phase-schedules` run root `.cartulary/test-results/20260603T204304Z-p686001`; `make phase-ledgers` run root `.cartulary/test-results/20260603T204317Z-p686275`; drift roots `.cartulary/test-results/20260603T204424Z-p687908` and `.cartulary/test-results/20260603T204428Z-p688151`. |
| Testcontainers fixtures | Phase C replaces service-backed suite/service artifacts, leases, scheduler lanes, cleanup stages, and fixture activity with backend-neutral `object_store` vocabulary. New `service-scope.json` and `service-lease.json` artifacts do not emit `minio` compatibility fields. | `tools/testservices`; `internal/testutil/suiteservices`; retained artifacts `.cartulary/test-results/20260603T204514Z-p691818/_shared/test-services/de5b40060835285b2dfb6796/service-scope.json` and `service-lease.json`; run-root scan `rg -n "MinIO\|MINIO\|minio\|minio_endpoint\|minio_container\|minio_ready\|minio_bucket\|minio_access_key\|minio_secret_key" .cartulary/test-results/20260603T204514Z-p691818` returned no matches. |
| `minio-go` dependency | `go.mod` contains `github.com/minio/minio-go/v7 v7.0.100`; `go.sum` contains matching module checksums. | `go.mod:12`, `go.sum:91`, `go.sum:92`. |
| SeaweedFS references | Phase B adds repo-controlled local SeaweedFS S3 references in Compose, local service scripts, dev env defaults, task-surface owner inputs, generated task-surface outputs, and the local probe. | Phase B diff; `make help-all` output showed SeaweedFS S3 local command wording. |
| SeaweedFS image tag and digest | SeaweedFS image digest verified from current registry metadata and `docker buildx imagetools inspect`: index digest `sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`. | `docker-compose.dev.yml`; `docker image inspect` showed repo digest `chrislusf/seaweedfs@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`. |
| Owner-document patches | Phase A owner diffs were added for public object-store dependency errors, backend-neutral managed object-store service binding, readiness state, range fallback semantics, backup/restore artifact adoption, threat-model coverage, and default SeaweedFS S3 wording. `SWFS-OWNER-STORAGEREF-001` remains an explicit owner-decision TODO at the Core 01 storage-ref anchor. | `docs/spec/01_architecture_storage_and_view_contracts.md`; `docs/spec/04_security_deployment_and_conformance.md`; `contracts/errors/index.json`; generated contract embeddings from `make generate`. |
| Acceptance command results | Phase D retained exact adapter, route, capability-probe, and partial SeaweedFS compatibility artifacts. `SWFS-AC-007` is claimable from the Phase D probe. `SWFS-COMP-001` through `SWFS-COMP-006` and `SWFS-COMP-011` are pass in the compatibility report. `SWFS-AC-015` remains unclaimed because not every `SWFS-COMP-*` case passed. Release-wide rows, SDK-only dependency rows, docs-wide rows, backup/restore rows, migration rows, and release-gate rows remain unclaimed. | See Sections 5.5, 6, and 8. |

## 2. Phase-by-phase progress status

No phase may be marked complete by prose-only completion. Each exit gate requires retained evidence, command output, repository diff evidence, or explicit unresolved blocker metadata tied to the implementation plan's acceptance-evidence matrix.

| Phase | Status | Current handoff state | Evidence needed to close |
| --- | --- | --- | --- |
| A. Contract cleanup | `complete_for_phase_a_with_retained_blockers` | Phase A owner cleanup is implemented in the working tree. Seven owner rows have local owner diffs or authored registry changes; `SWFS-OWNER-STORAGEREF-001` remains blocked with an explicit Core 01 TODO blocker rather than invented behavior. Downstream acceptance rows remain unclaimed. | Commit or review the Phase A diffs, then carry remaining implementation evidence into later phases. |
| B. Local service replacement | `complete_for_phase_b_with_retained_local_evidence` | Implemented. The default local object-store service is `seaweedfs-s3`, uses the pinned SeaweedFS tag plus digest, publishes only S3 port `8333` to the host, and passes the local S3 capability probe. | Keep the retained `services-up` evidence root and do not widen the claim to Phase C harness fixtures or release support. |
| C. Harness replacement | `complete_for_phase_c_with_retained_harness_evidence` | Implemented. Service-backed/testcontainer fixture ownership now uses SeaweedFS S3 or backend-neutral object-store vocabulary; generated scheduler/ledger outputs were regenerated from owner inputs; browser service-backed startup uses `seaweedfs-s3`; retained backend-integration artifacts prove `object_store` service scope/lease vocabulary and absence of forbidden MinIO fields. | Keep the retained Phase C evidence roots and do not widen this to Phase D adapter hardening, Phase E/F tooling, or release-wide occurrence closure. |
| D. Adapter hardening | `implemented_for_phase_d_with_retained_adapter_and_route_evidence` | Implemented. `internal/platform/objectstore` now has typed internal operation requests, backend-neutral adapter errors, deterministic adapter retry bounds, startup bucket/capability validation without managed-service bucket creation, stream-close observation, and retained Phase D support tests. Evidence routes continue to preserve public product envelopes while mapping adapter dependency errors to owner-owned public dependency errors. `tools/objectstoreprobe` now emits the Phase D capability probe plus a partial SeaweedFS compatibility report. | Keep retained Phase D evidence roots. Do not claim full compatibility, release-wide dependency-boundary, storage-ref migration, backup/restore, or occurrence-inventory rows. |
| E. Backup and restore | `TODO:evidence-required` | Not verified. Phase A added owner text for backup/restore artifact shapes, but no SeaweedFS backup manifest, shareable summary, restore verification artifact, or tooling evidence was produced. | `cartulary.object_store_backup_manifest.v1`, `cartulary.object_store_backup_summary.v1` if emitted, and `cartulary.restore_verification.v1` with `result='pass'`. |
| F. Migration tooling | `TODO:evidence-required` | Not verified. No application-stopped migration utility, lifecycle artifact, copy ledger, validation artifact, or rollback evidence was verified. | `cartulary.object_store_migration_run.v1`, copy ledger evidence, `cartulary.object_store_migration_validation.v1`, storage-ref owner citation, mismatch-blocking fixture, and rollback documentation evidence. |
| G. Security and release gate | `blocked` | Blocked. Phase A added threat-model owner text, but occurrence inventory, scanner/release evidence, SBOM/license, release-gate, and remaining owner dependency evidence are not available. | Occurrence inventory with zero invalid and zero unclassified rows, SBOM/license report, release manifest exposure scan, `make release-check` or plan-owned release gate evidence after prior blockers resolve. |

## 3. Completed work

Completed work is grouped by migration phase. Later release-wide acceptance rows remain governed by the implementation plan and require their own retained evidence.

Completed Phase A decisions and edits:

- `SWFS-OWNER-ERR-001`: Core 01 now defines `object_store_unavailable` and `object_store_access_rejected`, transport status, retry hints, reason-code vocabularies, and route-family applicability. `contracts/errors/index.json` mirrors those codes, and `make generate` updated generated contract embeddings.
- `SWFS-OWNER-CONFIG-001`: Core 04 now keeps object storage backend-neutral under `roots.object_storage`, defines managed-service `service_ref` environment binding, defaults, omitted/null behavior, redaction, validation reasons, and fail-closed readiness behavior without adding a SeaweedFS config namespace.
- `SWFS-OWNER-HEALTH-001`: Core 04 now distinguishes `/healthz` process liveness from structured `/readyz` object-store readiness states including `degraded_object_store` and recovery behavior.
- `SWFS-OWNER-STORAGEREF-001`: Core 01 now contains an explicit TODO blocker at the `evidence.storage_ref` anchor; no storage-ref grammar or key behavior was invented.
- `SWFS-OWNER-RANGE-001`: Core 01 now requires byte-range support for preview-handle redemption and permits full-stream download fallback for ordinary download-handle redemption.
- `SWFS-OWNER-BACKUP-001`: Core 01 and Core 04 now adopt the SeaweedFS S3 operator-private backup manifest and restore-verification artifact shapes as sufficient owner-level evidence shapes when tied to the same backup set and consistency point with non-null object SHA-256 proofs and redaction boundaries.
- `SWFS-OWNER-THREAT-001`: Core 04 §4.4 STRIDE coverage now names SeaweedFS S3 endpoint identity, direct upload target scope, object tampering, migration mismatch, credentials, admin surfaces, CORS, storage exhaustion, backup/restore, and migration validation risks.
- `SWFS-OWNER-DOCS-001`: Authored README and guide default wording now uses SeaweedFS S3-compatible or backend-neutral S3-compatible object-store terminology. Remaining SDK-only and current scheduler/generated-surface references are retained as blockers for later phases.

Completed Phase B decisions and edits:

- Local Compose now defines `seaweedfs-s3` as the default object-store service and pins `docker.io/chrislusf/seaweedfs:4.17@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`.
- Ordinary local Compose publishes only host S3 port `8333` for object storage. SeaweedFS admin, master, filer, volume, WebDAV, console, Iceberg, and debug ports are not host-published by the local Compose file.
- `scripts/dev-services.sh`, `scripts/dev-stack.sh`, `.env.example`, `AGENTS.md`, `Makefile`, `docs/testing-harness-nlspec.md`, and task-surface owner inputs now use `OBJECT_STORE_BUCKET`, `object-store-init`, `object-store-wait`, and SeaweedFS S3 local defaults.
- `tools/objectstoreprobe` provides the Phase B local S3 capability probe using the existing S3-compatible SDK dependency. It validates bucket create-or-exists, PutObject, HeadObject, full GetObject, range read, presigned PUT, CORS preflight, and cleanup.
- `make phase-schedules` regenerated task-surface outputs from owner inputs; generated files were not hand-edited.

Phase C replaced the service-backed harness/testservice fixture, browser E2E service-backed startup, scheduler resource vocabulary, and generated scheduler/ledger surfaces. Remaining MinIO references are classified in Section 7 and are not completion evidence for any release-wide `SWFS-AC-*` row.

Completed Phase C decisions and edits:

- `internal/testutil/s3test` now starts SeaweedFS S3 with the Phase B pinned image, S3 port `8333/tcp`, Phase B credentials, SeaweedFS command flags, and object-store/SeaweedFS S3 user-visible wording. The stable `CARTULARY_S3TEST_*` attach API remains unchanged.
- `internal/testutil/suiteservices` and `tools/testservices` now use `ServiceObjectStore = "object_store"` in service scope summaries, lease resources, startup attempts, timing labels, fixture events, cleanup stages, container labels, and failure summaries. New artifacts do not emit `minio` compatibility fields.
- `tools/scheduler_resource_registry.json`, `tools/execution_topology_manifest.json`, scheduler planners, validators, shell tests, and browser task-surface checks now use `object_store` for service-backed object-store resources. `make phase-schedules` regenerated downstream task/scheduler/browser outputs.
- `scripts/start-web-e2e.sh` standalone service startup now starts `seaweedfs-s3` with endpoint `localhost:8333` and Phase B credentials; browser/service-backed messages use object-store wording.
- Harness NLSpec, implementation testing guide, active phase-map support notes, and harness recovery support docs now use object-store wording for Phase C service-backed fixture surfaces. Generated phase ledgers were regenerated through `make phase-ledgers`.
- `make backend-integration` retained a passing service-backed proof at `.cartulary/test-results/20260603T204514Z-p691818`; `RAW-backend-integration-testutil` included `./internal/testutil/s3test`, and `TestHarnessStartsSeaweedFSS3AndRoundTripsObjects` passed.

Completed Phase D decisions and edits:

- `internal/platform/objectstore` now exposes typed internal adapter requests/results for upload-target creation, put, head, get/range get, list prefix, delete, and dev/test bucket ensure. Bucket selection remains immutable validated store configuration, not caller/browser input.
- Adapter errors now use backend-neutral codes and reason codes. MinIO SDK errors are mapped at the adapter boundary, and evidence routes translate dependency failures to `object_store_unavailable` or `object_store_access_rejected` without changing owner-owned `object_not_found` or evidence lifecycle errors.
- Adapter retry behavior is deterministic at the adapter boundary: maximum total adapter attempts are two, test backoff can be set to zero, and the SDK client is configured with `MaxRetries: 1` so SDK retries do not multiply adapter attempts. Put operations do not retry after accepted request bytes.
- Managed-service startup validates the configured bucket and required capabilities. Bucket creation remains in dev/test helpers and local object-store initialization, not production startup.
- Successful adapter read streams are close-observable through a test hook, and evidence preview/download code closes returned streams.
- Evidence blob-slot, attach, preview, and download route envelopes remain public-contract stable. Phase D added route tests for object-store dependency error mapping without changing generated contracts.
- `tools/objectstoreprobe` now emits `cartulary.object_store_capability_probe.v1` and `cartulary.seaweedfs_s3_compatibility_report.v1` with backend-neutral artifact fields and no `minio_*` fields. The compatibility report is intentionally partial for Phase D and keeps `SWFS-AC-015` unclaimed.

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
| `SWFS-OWNER-DOCS-001` | `cartulary-dev-guide.md`, `cartulary_repository_bootstrap_guide.md`, and docs index | Occurrence inventory plus docs diffs replacing MinIO server default wording with SeaweedFS S3 or generic S3-compatible wording while preserving SDK-only `minio-go` where present. | `partially_resolved_with_retained_blockers`; authored README/guide default wording, current harness docs/support wording, generated scheduler/ledger wording, and service-backed harness vocabulary are patched. Complete occurrence inventory and release/docs cleanup remain TODO. |

### 4.2 Repo or external facts still required

| Fact class | Required evidence | Current status |
| --- | --- | --- |
| Live repository commit | `git rev-parse HEAD` or equivalent retained artifact. | Phase C implementation command returned `92e679dcc8031e476e5586f0b11adfcaa93647d1`. |
| Compose/service manifests inspected | Path list and line ranges for every default service definition. | Phase B local Compose evidence is `docker-compose.dev.yml` plus `make services-up` run root `.cartulary/test-results/20260603T195218Z-p648759`; full repo occurrence inventory remains TODO for release-wide claims. |
| Make targets inspected | Path list and command registry evidence. | `make help-all` and `make explain-target TARGET=services-up DETAIL=summary` were run after Phase B; `services-up` reports `Postgres,SeaweedFS S3`. |
| Go module and lockfiles inspected | Path list and dependency evidence. | `go.mod` and `go.sum` show `minio-go` as the S3 SDK dependency; no lockfile edits were made. SBOM/license evidence remains TODO for release rows. |
| Frontend/browser route evidence inspected | Browser/E2E evidence path list. | TODO: browser route inventory and evidence-flow tests. |
| SeaweedFS image tag/digest | Repo-control file plus registry digest evidence. | Resolved for Phase B local Compose: `docker.io/chrislusf/seaweedfs:4.17@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`. |
| License/SBOM evidence | Retained SBOM/license report. | TODO: release SBOM/license artifact. |
| Threat model patch | Diff or retained scanner-ready document. | Core 04 §4.4 owner text patched; TODO: scanner/release evidence. |
| Owner error registry patch | Owner diff plus generated error contract. | Core 01 and generated error contracts patched; TODO: review/commit evidence if used for acceptance. |
| Owner config registry patch | Owner diff plus generated config-schema evidence. | Core 04 owner text patched; TODO: runtime config-schema and startup evidence. |

### 4.3 Unresolved evidence requirements

- TODO: `cartulary.seaweedfs_migration_occurrence_inventory.v1` over the full plan-defined scan scope.
- Phase D retained `cartulary.object_store_capability_probe.v1` with `result='pass'`, every required stage `status='pass'`, and `cleanup_result='clean'` at `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-capability-probe.json`. This supports `SWFS-AC-007`.
- Phase C retained backend-neutral harness artifact summaries with forbidden MinIO fields absent: `.cartulary/test-results/20260603T204514Z-p691818/_shared/test-services/de5b40060835285b2dfb6796/service-scope.json`, `.cartulary/test-results/20260603T204514Z-p691818/_shared/test-services/de5b40060835285b2dfb6796/service-lease.json`, and run-root forbidden-token scan with no matches.
- Phase D retained a partial `cartulary.seaweedfs_s3_compatibility_report.v1` at `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-compatibility-report.json`. `SWFS-COMP-001` through `SWFS-COMP-006` and `SWFS-COMP-011` have `status='pass'`; `SWFS-COMP-007`, `SWFS-COMP-008`, `SWFS-COMP-009`, `SWFS-COMP-010`, `SWFS-COMP-012`, `SWFS-COMP-013`, and `SWFS-COMP-014` remain `status='not_run'`. `SWFS-AC-015` remains unclaimed.
- Phase D retained backend route evidence for object-store dependency error mapping at `.cartulary/test-results/20260603T224415Z-p821138/_shared/backend-integration-evidence-shard-02/runner.jsonl` under `TestPhase5_AttachRouteContract_I_5_05/object-store_dependency_errors_use_owner_public_mapping`. This covers only the dependency-error mapping portion; full blob-slot timers, same-origin handle forbidden-value scans, and the complete negative evidence-state matrix remain unclaimed.
- TODO: backup/restore manifests and restore verification artifacts.
- TODO: migration run, copy ledger, validation, mismatch-blocking, and rollback artifacts.
- TODO: SBOM/license, release gate, and later-phase `make agent-finalize` retained evidence after implementation work.

## 5. Validation status

### 5.1 Planning inspection commands run

These commands were used to prepare this handoff. They are not acceptance evidence for any `SWFS-AC-*` row.

| Command | Result summary |
| --- | --- |
| `sed -n '1,260p' docs/seaweedfs_s3_migration_implementation_plan.md` and later ranges through the end of the file | Read controlling implementation plan, phases, owner patch registry, acceptance matrix, and claimability model. |
| `sed -n '1,220p' docs/testing-harness-nlspec.md` and targeted line inspection | Initial planning confirmed Make-owned harness authority and the then-current `minio-init` public target row; Phase B replaced that local target wording before Phase C. |
| `sed -n '1,220p' docs/domain.md` | Confirmed domain vocabulary boundary and implementation-support treatment. |
| `sed -n '1,220p' docs/spec/00_document_set_status_and_precedence.md` | Confirmed Core 00 authority and owner-document precedence. |
| `git rev-parse HEAD` | Returned `92fa79edbedc95c9c6ad51e666b107b09dc0755c`. |
| `git status --short` | Returned no output before this handoff file was added. |
| `rg -n "SeaweedFS\|seaweedfs\|MinIO server\|minio/minio\|minio-go\|minio_" ...` | Initial planning found MinIO server, harness, fixture, docs, manifest, and SDK references. Phase B and Phase C replaced the local and service-backed harness surfaces; remaining references are classified in Section 7. |
| `make help-all | rg -i 'seaweed\|object\|s3\|minio\|backup\|restore\|release-check\|agent-finalize\|harness\|browser-e2e\|test-fast\|check\|generate\|drift\|sbom\|license\|gosec\|vuln'` | Initial planning showed the public command surface still included `minio-init`; Phase B replaced that with object-store public target wording. |

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

- `make harness-contract` was skipped during Phase A because Phase A did not edit `docs/testing-harness-nlspec.md`, harness command topology, generated harness manifests, or harness public command surfaces.
- Broad `make check` was skipped during Phase A because Phase A touched owner docs, authored contract registry input, generated contract embeddings, support docs, and this handoff only; no implementation behavior or service/harness replacement was changed.

### 5.3 Phase B verification commands

The commands below verify the Phase B local-service replacement only. They do not claim Phase C service-backed harness replacement, Phase D adapter hardening, release SBOM/license closure, or release-wide occurrence inventory closure.

| Command | Status | Run root or evidence |
| --- | --- | --- |
| `make phase-schedules` | `pass`; regenerated task-surface outputs from `tools/execution_topology_manifest.json`. | `.cartulary/test-results/20260603T194856Z-p641032` |
| `make generated-artifact-policy-check` | `pass`. | `.cartulary/test-results/20260603T194954Z-p642411` |
| `make json-shape-check` | `pass`. | `.cartulary/test-results/20260603T194959Z-p642656` |
| `make help-all` | `pass`; output names SeaweedFS S3 for local `db-up`, `services-up`, and `services-down`, and lists `object-store-init`. | terminal output summarized here |
| `make explain-target TARGET=services-up DETAIL=summary` | `pass`; output reports `services: Postgres,SeaweedFS S3`. | terminal output summarized here |
| `make phase-schedule-drift` | `pass`. | `.cartulary/test-results/20260603T195015Z-p643182` |
| `make lint-scripts` | `pass`. | `.cartulary/test-results/20260603T195025Z-p643537` |
| `make lint-shell` | `pass` after final shell edits. | `.cartulary/test-results/20260603T195749Z-p652769` |
| `make lint-go-format` | `pass`; no wrapper output. | terminal status |
| `make lint-go-vet` | `pass`; no wrapper output. | terminal status |
| `make services-up` | `pass`; started Postgres plus `seaweedfs-s3`, removed the prior MinIO orphan, and emitted the local capability probe artifact. | `.cartulary/test-results/20260603T195218Z-p648759`; probe `.cartulary/test-results/20260603T195218Z-p648759/services-up/object-store-capability-probe.json` |
| `make object-store-init` | `pass`; public provider-neutral helper target. | `.cartulary/test-results/20260603T195740Z-p652397` |
| `make agent-finalize` | `pass`; generated outputs unchanged; finalizer reported `results_dir=-`. | `.cartulary/test-results/20260603T200800Z-p658561` |
| `make services-down` | `pass`; stopped local Compose services after evidence capture. | `.cartulary/test-results/20260603T195829Z-p654702` |

Skipped Phase B checks:

- Broad `make check` was skipped because Phase B touched local-service Compose/scripts, local task-surface owner inputs, generated task surfaces, local docs/support wording, and a narrow support probe. It did not implement Phase C service-backed fixtures, Phase D adapter hardening, or broad runtime behavior.
- Phase C service-backed harness/browser fixture replacement checks were skipped because they are out of Phase B scope and remain blockers.

### 5.4 Phase C verification commands

The commands below verify the Phase C harness replacement only. They do not claim Phase D adapter hardening, Phase E/F tooling, release SBOM/license closure, or release-wide occurrence inventory closure.

| Command | Status | Run root or evidence |
| --- | --- | --- |
| `make phase-schedules` | `pass`; regenerated task/scheduler/browser outputs from owner inputs after `object_store` resource changes. | `.cartulary/test-results/20260603T204304Z-p686001` |
| `make phase-ledgers` | `pass`; regenerated phase ledgers after phase-map support wording changed. | `.cartulary/test-results/20260603T204317Z-p686275` |
| `make generated-artifact-policy-check` | `pass`. | `.cartulary/test-results/20260603T204358Z-p687336` |
| `make json-shape-check` | `pass`. | `.cartulary/test-results/20260603T204402Z-p687576` |
| `make phase-schedule-drift` | `pass`. | `.cartulary/test-results/20260603T204424Z-p687908` |
| `make phase-ledger-drift` | `pass`. | `.cartulary/test-results/20260603T204428Z-p688151` |
| `make lint-scripts` | `pass`. | `.cartulary/test-results/20260603T204435Z-p688595` |
| `make lint-shell` | `pass`. | `.cartulary/test-results/20260603T204441Z-p689096` |
| `make lint-go-format` | `pass`; no wrapper output. | terminal status |
| `make lint-go-vet` | `pass`; no wrapper output. | terminal status |
| `make backend-integration` | `pass`; 155 tests. Retained `RAW-backend-integration-testutil` includes `./internal/testutil/s3test`; `TestHarnessStartsSeaweedFSS3AndRoundTripsObjects` passed; service artifacts use `object_store` and the pinned SeaweedFS image. | `.cartulary/test-results/20260603T204514Z-p691818`; `backend-integration/backend-integration-testutil/phase-summary.json`; `_shared/backend-integration-testutil-shard-02/runner.jsonl`; `_shared/test-services/de5b40060835285b2dfb6796/service-scope.json`; `_shared/test-services/de5b40060835285b2dfb6796/service-lease.json` |
| `rg -n "MinIO\|MINIO\|minio\|minio_endpoint\|minio_container\|minio_ready\|minio_bucket\|minio_access_key\|minio_secret_key" .cartulary/test-results/20260603T204514Z-p691818` | `pass`; no matches. | terminal status |
| `make agent-finalize` | `pass`; generated outputs unchanged; finalizer reported `results_dir=-`. | `.cartulary/test-results/20260603T205417Z-p704687` |

Skipped Phase C checks:

- Broad `make check` was skipped because Phase C touched harness/testservice fixtures, scheduler/browser startup surfaces, owner inputs, generated schedules/ledgers, and harness docs/support wording. The plan selected `make backend-integration` as the narrow service-backed fixture proof.
- Phase D adapter-hardening, Phase E backup/restore, Phase F migration, and Phase G release/SBOM/release-gate checks were skipped as out of Phase C scope.

### 5.5 Phase D verification commands

The commands below verify Phase D adapter hardening, object-store capability probing, partial compatibility reporting, and backend route dependency-error mapping. They do not claim full `SWFS-AC-015`, browser E2E, backup/restore, migration, SBOM/license, release-gate, or release-wide occurrence-inventory rows.

| Command | Status | Run root or evidence |
| --- | --- | --- |
| `make format` | `pass`; applied authored Go formatting before final checks. | `.cartulary/test-results/20260603T223759Z-p804433` |
| `make lint-go-format` | `pass`. | `.cartulary/test-results/20260603T224355Z-p817452` |
| `make lint-go-vet` | `pass`. | `.cartulary/test-results/20260603T224355Z-p817480` |
| `make generated-artifact-policy-check` | `pass`. | `.cartulary/test-results/20260603T224355Z-p817469` |
| `make json-shape-check` | `pass`. | `.cartulary/test-results/20260603T224355Z-p817454` |
| `make backend-integration-support` | `pass`; retained Phase D adapter subtests under `TestSupportPhase0_ManagedServiceObjectStoreBinding/phase_d_adapter_contract_hardening` for input contracts, retry behavior, managed-service bucket startup validation, and read-stream closure observability. | `.cartulary/test-results/20260603T223803Z-p806186`; runner `.cartulary/test-results/20260603T223803Z-p806186/_shared/backend-integration-platform-shard-03/runner.jsonl` |
| `make services-up` | `pass`; emitted Phase D capability probe and partial compatibility report. | `.cartulary/test-results/20260603T224410Z-p820796`; probe `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-capability-probe.json`; report `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-compatibility-report.json` |
| `make backend-integration` | `pass`; 155 tests. Retained evidence route dependency-error mapping subtests under `TestPhase5_AttachRouteContract_I_5_05/object-store_dependency_errors_use_owner_public_mapping`. | `.cartulary/test-results/20260603T224415Z-p821138`; runner `.cartulary/test-results/20260603T224415Z-p821138/_shared/backend-integration-evidence-shard-02/runner.jsonl` |

Skipped Phase D checks:

- `make lint-scripts` and `make lint-shell` were skipped because Phase D did not change shell scripts or the shell invocation path for `tools/objectstoreprobe`.
- `make phase-schedules`, `make phase-ledgers`, and their drift checks were skipped because Phase D did not change scheduler/topology owner inputs or phase-map owner inputs.
- Browser/service-backed evidence targets were skipped because Phase D did not change browser flow implementation; route-envelope and dependency-error behavior were covered by backend integration tests.
- Broad `make check` was skipped because the edits were scoped to the adapter, evidence route mapping, probe/report support, and narrow tests, and the required narrower targets passed.

### 5.6 Acceptance command backlog

All plan-owned acceptance commands below remain unrun and all rows remain unclaimed unless a retained artifact supplies the required evidence. Phase B makes `SWFS-AC-005` claimable with the retained local evidence named in Section 5.3. Phase C makes `SWFS-AC-014` claimable for harness artifacts with the retained evidence named in Section 5.4. Phase D makes `SWFS-AC-007` claimable with the retained capability probe named in Section 5.5, and makes only the explicitly passing `SWFS-COMP-*` rows in Section 6 claimable. No release-wide row is claimed.

| Command or evidence source | Status |
| --- | --- |
| `TODO:repo-compose-inventory-command` | `TODO: unrun; required for SWFS-AC-001`. |
| `TODO:dependency-boundary-command` | `TODO: unrun; required for SWFS-AC-002`. |
| owner diff for `SWFS-AC-003` and `SWFS-AC-004` | `TODO: working-tree owner wording exists, but acceptance completion is not claimed until reviewed/retained as evidence`. |
| `TODO:compose-service-validation-command` | Phase B local evidence retained for `SWFS-AC-005`; full occurrence inventory still TODO for release-wide closure. |
| occurrence inventory plus release manifest scan | `TODO: unrun; required for SWFS-AC-006`. |
| `make services-up` | Phase D capability probe retained for `SWFS-AC-007`; compatibility report remains partial, so `SWFS-AC-015` is unclaimed. |
| `TODO:startup-failure-e2e-command` | `TODO: unrun; required for SWFS-AC-008`. |
| `TODO:runtime-outage-e2e-command` | `TODO: unrun; required for SWFS-AC-009`. |
| `make backend-integration` | Phase D retained route dependency-error mapping evidence. Full blob-slot timer evidence for `SWFS-AC-010` remains unclaimed. |
| `TODO:browser-evidence-e2e-command` | `TODO: unrun; required for SWFS-AC-011`. |
| `TODO:evidence-handle-e2e-command` | TODO: unrun; required for full `SWFS-AC-012`; Phase D backend route evidence covers only dependency-error mapping, not forbidden-value same-origin handle scans. |
| `TODO:evidence-negative-matrix-command` | TODO: unrun; required for full `SWFS-AC-013`; Phase D backend route evidence covers only object-store dependency-error reason codes. |
| `TODO:harness-artifact-scan-command` | Phase C retained exact backend-integration harness artifact scan evidence for `SWFS-AC-014`; broader release-wide occurrence inventory remains TODO. |
| `make services-up` compatibility report | Partial; `SWFS-COMP-001` through `SWFS-COMP-006` and `SWFS-COMP-011` pass, but `SWFS-AC-015` remains unclaimed because not every `SWFS-COMP-*` row passed. |
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

Every `SWFS-AC-*` row remains non-claimable in this handoff unless the current status explicitly names retained evidence. Phase D names retained pass evidence for `SWFS-AC-007`; Phase C names retained pass evidence for `SWFS-AC-014`; Phase B local evidence remains limited to `SWFS-AC-005`.

| ID | Criterion | Closure class | Owner dependencies | Command or evidence source | Artifact schema | Pass predicate | Failure class | Current status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `SWFS-AC-001` | No default development, CI, service-backed test, or release Compose/manifest starts, pulls, or names a MinIO server container or image. | `repo_or_external_fact_required` | none | `TODO:repo-compose-inventory-command` | occurrence inventory | zero invalid MinIO server service rows | `blocks_release` | `not_claimable`; Phase B and Phase C replaced local and service-backed harness defaults, but a complete repo/release occurrence inventory is still TODO. |
| `SWFS-AC-002` | `github.com/minio/minio-go/v7`, if present, appears only as an S3 client dependency behind the object-store adapter. | `repo_or_external_fact_required` | none | `TODO:dependency-boundary-command` | SBOM/license report | no runtime service, fixture, readiness label, or operator instruction treats SDK as server support | `blocks_release` | `not_claimable`; `minio-go` is present in `go.mod`; TODO: dependency-boundary and SBOM/license report. |
| `SWFS-AC-003` | Core 01 object-storage wording names SeaweedFS S3 as the default local/disconnected S3-compatible target while preserving generic S3 compatibility. | `blocked_until_owner_patch` | `SWFS-OWNER-DOCS-001` or core owner patch | owner diff | not applicable | target wording patched at owner anchor | `blocks_docs` | `blocked`; working-tree Core 01 wording exists, but no acceptance completion is claimed until retained owner-diff evidence is reviewed. |
| `SWFS-AC-004` | Core 04 disconnected deployment wording names one SeaweedFS S3 container or equivalent S3-compatible object store. | `blocked_until_owner_patch` | `SWFS-OWNER-DOCS-001` or core owner patch | owner diff | not applicable | target wording patched at owner anchor | `blocks_docs` | `blocked`; working-tree Core 04 wording exists, but no acceptance completion is claimed until retained owner-diff evidence is reviewed. |
| `SWFS-AC-005` | The default local service is named `seaweedfs-s3`, uses a pinned SeaweedFS image tag plus digest, and exposes only the S3 endpoint in ordinary local development. | `repo_or_external_fact_required` | none | `make services-up`; `docker-compose.dev.yml`; post-run Compose inspection | capability probe plus repo-controlled Compose file | service name, digest, and exposure table match Sections 6 and 7 | `blocks_phase` | `claimable_for_phase_b_local`; evidence: `docker-compose.dev.yml`, probe artifact `.cartulary/test-results/20260603T195218Z-p648759/services-up/object-store-capability-probe.json`, and post-run inspection showing only host-published object-store port `8333`. Does not claim Phase C or release-wide occurrence closure. |
| `SWFS-AC-006` | Production documentation forbids default exposure of SeaweedFS admin, master, filer, volume, WebDAV, and debug surfaces. | `repo_or_external_fact_required` | `SWFS-OWNER-DOCS-001` | occurrence inventory plus release manifest scan | occurrence inventory | no invalid exposure instructions | `blocks_release` | `blocked`; Phase A touched default wording only; TODO: occurrence inventory and release manifest scan. |
| `SWFS-AC-007` | The capability probe completes required PutObject, HeadObject, full GetObject, range GetObject, DeleteObject, CORS preflight, and presigned PUT stages within timeout and retry bounds. | `plan_local_closed` | `SWFS-OWNER-RANGE-001` if range owner declares required semantics | `make services-up` | `cartulary.object_store_capability_probe.v1` | `result='pass'` and every required stage `status='pass'` | `blocks_phase` | `claimable_for_phase_d`; retained artifact `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-capability-probe.json` has `result='pass'`, `cleanup_result='clean'`, and pass stages for `endpoint_reachability`, `bucket_validation`, `put_primary`, `head_primary`, `get_primary`, `range_primary`, `create_direct_upload_target`, `cors_preflight`, `direct_put`, `head_direct`, `get_direct`, `delete_primary`, `verify_primary_deleted`, `delete_direct`, and `verify_direct_deleted`. |
| `SWFS-AC-008` | In production profile, missing bucket, denied credentials, endpoint unreachable, CORS failure, or missing required capability fails startup before ready state. | `blocked_until_owner_patch` | `SWFS-OWNER-CONFIG-001`, `SWFS-OWNER-HEALTH-001` | `TODO:startup-failure-e2e-command` | probe artifact plus startup diagnostics | no listener becomes ready; diagnostic reason matches Section 9.3 | `blocks_release` | `blocked`; owner text patched in working tree; TODO: startup failure E2E artifact. |
| `SWFS-AC-009` | After a post-ready object-store outage, ordinary non-evidence workbook row editing remains available while evidence operations fail through mapped public dependency errors. | `blocked_until_owner_patch` | `SWFS-OWNER-ERR-001`, `SWFS-OWNER-HEALTH-001` | `TODO:runtime-outage-e2e-command` | public error contract evidence | non-evidence route succeeds; evidence route errors match Section 10.2 | `blocks_phase` | `blocked`; Phase D retained evidence-route dependency-error mapping at `.cartulary/test-results/20260603T224415Z-p821138/_shared/backend-integration-evidence-shard-02/runner.jsonl`, but no runtime-outage E2E artifact proves ordinary non-evidence workbook editing remains available. |
| `SWFS-AC-010` | `POST /api/v1/object-blobs` still returns the Core-owned blob-slot response shape and timers. | `plan_local_closed` | none | `TODO:blob-slot-contract-command` | public route evidence | response includes required fields; timers unchanged | `blocks_phase` | `not_claimable`; Phase D retained dependency-error mapping for blob-slot create at `.cartulary/test-results/20260603T224415Z-p821138/_shared/backend-integration-evidence-shard-02/runner.jsonl`, but no exact retained artifact proves the complete success envelope and timer predicate for this row. |
| `SWFS-AC-011` | Browser E2E creates a pending blob slot, uploads bytes to SeaweedFS, attaches blob to evidence, receives projection row, and emits collaboration update. | `repo_or_external_fact_required` | none | `TODO:browser-evidence-e2e-command` | browser/evidence artifact | full two-step flow succeeds without raw preview/download object URLs | `blocks_phase` | `not_claimable`; TODO: browser evidence E2E artifact. |
| `SWFS-AC-012` | Preview and download issuance return only same-origin opaque evidence handles and never return bucket names, object keys, raw storage refs, raw SeaweedFS URLs, or long-lived object-store credentials. | `plan_local_closed` | none | `TODO:evidence-handle-e2e-command` | public route evidence | response fields are same-origin handle only; forbidden values absent | `blocks_phase` | `not_claimable`; Phase D retained preview dependency-error mapping at `.cartulary/test-results/20260603T224415Z-p821138/_shared/backend-integration-evidence-shard-02/runner.jsonl`, but no exact forbidden-value same-origin handle artifact was retained. |
| `SWFS-AC-013` | Evidence negative cases for missing, pending, failed, quarantined, oversized, unsupported, expired, consumed, stale, and expired-upload-target states produce exact owner-mapped errors. | `blocked_until_owner_patch` | `SWFS-OWNER-ERR-001`, `SWFS-OWNER-RANGE-001` | `TODO:evidence-negative-matrix-command` | public error matrix | every case matches owner registry and Section 10 | `blocks_phase` | `blocked`; Phase D retained covered dependency-error reason-code evidence at `.cartulary/test-results/20260603T224415Z-p821138/_shared/backend-integration-evidence-shard-02/runner.jsonl`, but the complete negative-state matrix remains unclaimed. |
| `SWFS-AC-014` | Harness artifacts use backend-neutral object-store vocabulary and contain no MinIO server readiness fields. | `repo_or_external_fact_required` | none | `make backend-integration` plus retained run-root forbidden-token scan | harness artifact summaries | forbidden fields absent; required fields present | `blocks_phase` | `claimable_for_phase_c_harness`; evidence: `.cartulary/test-results/20260603T204514Z-p691818/_shared/test-services/de5b40060835285b2dfb6796/service-scope.json` has top-level `object_store.started=true` and no `minio` top-level field; `.cartulary/test-results/20260603T204514Z-p691818/_shared/test-services/de5b40060835285b2dfb6796/service-lease.json` records service `object_store` and image `docker.io/chrislusf/seaweedfs:4.17@sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9`; forbidden-token scan over `.cartulary/test-results/20260603T204514Z-p691818` found no `MinIO`, `minio`, `minio_endpoint`, `minio_container`, `minio_ready`, `minio_bucket`, `minio_access_key`, or `minio_secret_key` matches. |
| `SWFS-AC-015` | The SeaweedFS compatibility suite passes every `SWFS-COMP-*` case and contains no multipart or presigned-GET skip row. | `repo_or_external_fact_required` | none | `make services-up` compatibility report | `cartulary.seaweedfs_s3_compatibility_report.v1` | every case pass; no forbidden skip | `blocks_phase` | `not_claimable`; retained report `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-compatibility-report.json` has `result='fail'`: `SWFS-COMP-001` through `SWFS-COMP-006` and `SWFS-COMP-011` pass, but `SWFS-COMP-007`, `SWFS-COMP-008`, `SWFS-COMP-009`, `SWFS-COMP-010`, `SWFS-COMP-012`, `SWFS-COMP-013`, and `SWFS-COMP-014` are `not_run`. |
| `SWFS-AC-016` | Each successful backup set against SeaweedFS includes a private manifest tied to the same backup set and consistency point as Postgres. | `blocked_until_owner_patch` | `SWFS-OWNER-BACKUP-001` | `TODO:backup-command` | `cartulary.object_store_backup_manifest.v1` | manifest valid; every object SHA-256 non-null | `blocks_release` | `blocked`; owner text patched in working tree; TODO: backup manifest artifact. |
| `SWFS-AC-017` | Restoring the latest successful retained backup into fresh Postgres and fresh SeaweedFS rebuilds projections and preserves blob lifecycle consistency. | `blocked_until_owner_patch` | `SWFS-OWNER-BACKUP-001` | `TODO:restore-verification-command` | `cartulary.restore_verification.v1` | `result='pass'` | `blocks_release` | `blocked`; owner text patched in working tree; TODO: restore verification artifact. |
| `SWFS-AC-018` | Default MinIO-to-SeaweedFS migration preserves bucket name and object keys and does not mutate database `storage_ref` values. | `blocked_until_owner_patch` | `SWFS-OWNER-STORAGEREF-001` | `TODO:migration-fixture-command` | migration run and validation artifacts | database refs unchanged; copy ledger target keys match source | `blocks_phase` | `blocked`; TODO: storage-ref owner citation/patch and migration fixture evidence. |
| `SWFS-AC-019` | Migration validation emits `cartulary.object_store_migration_validation.v1` with `result='pass'` only when blocking arrays are empty and every preview sample passes. | `plan_local_closed` | none | `TODO:migration-validation-command` | `cartulary.object_store_migration_validation.v1` | result computation matches Section 14.5 | `blocks_phase` | `not_claimable`; TODO: migration validation command and artifact. |
| `SWFS-AC-020` | Any target-side object existing with a different size or SHA-256 than source blocks migration cutover. | `plan_local_closed` | none | `TODO:migration-target-mismatch-command` | copy ledger and validation artifact | mismatch produces blocking failure; no cutover | `blocks_phase` | `not_claimable`; TODO: mismatch fixture and artifacts. |
| `SWFS-AC-021` | Threat model update includes every STRIDE row listed in Section 15 and names SeaweedFS direct upload, credentials, admin surfaces, backup/restore, and migration validation. | `blocked_until_owner_patch` | `SWFS-OWNER-THREAT-001` | threat-model diff or scanner document | not applicable | every row covered with control and verification hook | `blocks_release` | `blocked`; Core 04 §4.4 owner text patched in working tree; TODO: scanner/release evidence. |
| `SWFS-AC-022` | Release SBOM and license gates identify no MinIO server artifact; if `minio-go` remains, release notes identify it as client dependency only. | `repo_or_external_fact_required` | none | `TODO:sbom-license-command` | SBOM/license report | no MinIO server; SDK-only classification for `minio-go` | `blocks_release` | `not_claimable`; TODO: SBOM/license report and release notes check. |
| `SWFS-AC-023` | Default docs no longer describe MinIO server as default local, disconnected, CI, service-backed test, or release-support object-store target. | `repo_or_external_fact_required` | `SWFS-OWNER-DOCS-001` | occurrence inventory | `cartulary.seaweedfs_migration_occurrence_inventory.v1` | zero invalid occurrences | `blocks_docs` | `blocked`; authored default wording and Phase C harness docs/support wording are patched, but complete occurrence inventory and release/docs cleanup remain TODO. |
| `SWFS-AC-024` | Full release gate runs required compatibility, object-store reachability, evidence, backup/restore, security, license/SBOM, and full repository check gates. | `repo_or_external_fact_required` | owner blockers resolved | `TODO:release-gate-command` | release gate summary | all child predicates pass and no unresolved release blockers remain | `blocks_release` | `blocked`; TODO: all prior blockers plus release-gate summary. |
| `SWFS-AC-025` | Post-migration occurrence inventory classifies every remaining MinIO token with zero invalid and zero unclassified rows. | `plan_local_closed` | none | `TODO:occurrence-inventory-command` | `cartulary.seaweedfs_migration_occurrence_inventory.v1` | `result='pass'` | `blocks_release` | `not_claimable`; TODO: occurrence inventory artifact. |

### 6.1 SeaweedFS compatibility case status

Retained report: `.cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-compatibility-report.json` with `schema_id='cartulary.seaweedfs_s3_compatibility_report.v1'`, `object_store_backend='seaweedfs_s3'`, `result='fail'`, and `forbidden_skip_rows=[]`.

| ID | Retained status | Current claim status |
| --- | --- | --- |
| `SWFS-COMP-001` | `status='pass'`; source `bucket_validation`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-002` | `status='pass'`; source `put_primary`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-003` | `status='pass'`; source `head_primary`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-004` | `status='pass'`; source `get_primary`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-005` | `status='pass'`; source `range_primary`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-006` | `status='pass'`; source `verify_primary_deleted`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-007` | `status='not_run'`; `reason_code='phase_d_evidence_not_retained'`. | `unclaimed` |
| `SWFS-COMP-008` | `status='not_run'`; `reason_code='after_expiry_path_not_retained'`. | `unclaimed` |
| `SWFS-COMP-009` | `status='not_run'`; `reason_code='phase_d_evidence_not_retained'`. | `unclaimed` |
| `SWFS-COMP-010` | `status='not_run'`; `reason_code='phase_d_evidence_not_retained'`. | `unclaimed` |
| `SWFS-COMP-011` | `status='pass'`; source `cors_preflight`. | `claimable_from_phase_d_report` |
| `SWFS-COMP-012` | `status='not_run'`; `reason_code='phase_d_evidence_not_retained'`. | `unclaimed` |
| `SWFS-COMP-013` | `status='not_run'`; `reason_code='phase_d_evidence_not_retained'`. | `unclaimed` |
| `SWFS-COMP-014` | `status='not_run'`; `reason_code='phase_d_evidence_not_retained'`. | `unclaimed` |

## 7. Remaining MinIO classification

Phase D reclassified remaining `MinIO`/`minio` references after adapter hardening and probe/report changes.

| Class | Status | Remaining references or evidence |
| --- | --- | --- |
| Invalid Phase C harness/test fixture surface | `resolved_for_phase_c` | Replaced in `internal/testutil/s3test`, `internal/testutil/suiteservices`, `internal/testutil/testcontainersx`, `tools/testservices`, `scripts/start-web-e2e.sh`, service-backed scheduler/task-surface validators and tests, scheduler resource owner inputs, browser task-surface checks, `docs/testing-harness-nlspec.md`, active phase-map support notes, and generated scheduler/ledger outputs. |
| Allowed `minio-go` SDK-only reference | `allowed_for_phase_d` | `go.mod`, `go.sum`, `internal/platform/objectstore`, `internal/testutil/s3test`, `tools/objectstoreprobe`, and dependency-list docs retain `github.com/minio/minio-go/v7` or `minio.*` SDK symbols as generic S3-compatible client usage. Phase D keeps the SDK behind the object-store adapter/probe/fixture boundary. This does not claim `SWFS-AC-002` or `SWFS-AC-022`; dependency-boundary and SBOM/license evidence remain later-phase work. |
| Already-resolved Phase B local default or negative guard | `allowed/resolved` | Local default service surfaces now use `seaweedfs-s3`. Negative checks in `scripts/check-backend-task-surface.sh` intentionally reject retired `minio-init` and `minio-reset` compatibility aliases. |
| Generated/downstream output requiring owner-input changes | `resolved_for_phase_c` | `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, browser batch/scheduler outputs, and `docs/testing/phase*_coverage_ledger.md` were regenerated through `make phase-schedules` and `make phase-ledgers`. A targeted search of generated scheduler/ledger outputs found no `MinIO`/`minio` references after regeneration. |
| Legacy external S3 or migration-source reference | `allowed/historical_or_plan_context` | The implementation plan and this handoff intentionally discuss MinIO as the migration source or historical server state. Historical recovery handoffs/audits, archive/source-extract documents, and explanatory prior-state docs retain historical MinIO mentions and are not current Phase C harness instructions. |
| Later-phase release/SBOM/docs blocker | `blocked_outside_phase_d` | `scripts/generate-sbom-license-evidence.mjs` still scans for `minio/minio` images for release evidence; `tools/phase10browserrestore` still has later-phase backup/restore credential defaults; release/docs/source-archive cleanup, SBOM/license evidence, and full occurrence inventory remain Phase E/F/G or broader migration blockers. |
| Redaction-pattern compatibility fixtures | `allowed` | `tools/harness_redaction_manifest.json`, `scripts/test-harness-contracts.mjs`, `internal/testutil/harnessredact` keep `minio` credential-name examples to verify old secret names are still redacted. |
| Backend-neutral config/service-ref test label | `later_phase_or_owner_cleanup` | `internal/platform/config/config_phase0_test.go` still uses the literal `minio-primary` as a managed object-storage service-ref fixture. It is not a Phase D adapter, route, probe, or service startup surface; classifying or replacing service-ref label examples belongs with broader config/owner or occurrence-inventory cleanup. |
| Adapter/user-facing wording | `resolved_for_phase_d` | Invalid adapter wording such as `create minio client` was replaced with backend-neutral S3 wording. Remaining adapter references to `minio` are SDK symbol/import references only. |

## 8. Risks, assumptions, and next actions

### 8.1 Risks and assumptions

- Current default local surfaces and service-backed harness/browser fixture surfaces now use SeaweedFS S3 or backend-neutral object-store vocabulary. Release-wide occurrence closure remains unclaimed.
- `minio-go` presence is not itself invalid, but any release-wide claim that it is SDK-only is currently unverified by SBOM/license and dependency-boundary evidence.
- Runtime implementation evidence remains partial: Phase D retained adapter and backend route dependency-error evidence, but full runtime-outage E2E, browser evidence flows, readiness route shape, backup/restore, migration, and threat-model release checks remain missing. Storage-ref grammar and key-generation behavior remain owner-blocked by `SWFS-OWNER-STORAGEREF-001`.
- The implementation plan requires byte-equivalence proof by SHA-256. Do not use ETag-only evidence for migration equivalence.
- The handoff's planning inspection and Phase B command output are summarized here. Release-wide claims require retained artifacts and occurrence inventory, not this prose summary alone.

### 8.2 Recommended next actions

1. Review and commit the Phase A owner, registry, generated-contract, support-doc, and handoff diffs if accepted.
2. Resolve `SWFS-OWNER-STORAGEREF-001` with an owner decision for storage-ref grammar, canonicalization, key generation, maximum length, and invalid-state behavior.
3. Build `cartulary.seaweedfs_migration_occurrence_inventory.v1` and use it to classify all remaining MinIO tokens before release-wide docs or SBOM/license claims.
4. Complete the remaining `SWFS-COMP-*` compatibility cases and retain a full passing compatibility report before claiming `SWFS-AC-015`.
5. Add browser/runtime outage evidence for blob-slot success timers, same-origin handle forbidden-value scans, and the complete evidence negative-state matrix before claiming `SWFS-AC-010`, `SWFS-AC-012`, or `SWFS-AC-013`.
6. Implement backup/restore and migration tooling only after the owner dependencies and adapter/probe evidence are in place.
7. Finish with SBOM/license, occurrence inventory, release manifest exposure scan, `make agent-finalize`, and the full release gate after all release blockers are closed.
