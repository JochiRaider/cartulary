# Phase 10 Implementation Plan

## Summary

This file is the execution roadmap, progress marker, and handoff aid for Cartulary Phase 10: operational backup, restore, and restore verification.

`docs/guides/cartulary_implementation_testing_guide.md` section 12, Phase 10, is the controlling implementation-planning row set. Normative implementation-conformance behavior remains owned by Core 00 through Core 04. Core 05 applies only to claim-bearing timed or fixture-sensitive publication boundaries.

This planning artifact does not implement Phase 10 behavior. Operational backup and restore are deployment-local recovery behavior, not Incident Portability Extension Profile behavior, not snapshot/reporting behavior, and not benchmark or release-publication behavior.

Authority model:
- Core 00 through Core 04 own implementation-conformance behavior.
- Core 05 owns claim-bearing timed or fixture-sensitive publication only.
- `PHASE10_IMPLEMENTATION_PLAN.md` is a planning artifact only.
- Phase 5 through Phase 9 plans are structural examples only.
- Generated ledgers, generated schedules, support-only tests, visual goldens, and retained artifacts are not behavior authorities.
- `docs/domain.md` is vocabulary and concept support only.

Current repo status after Sprint 0 setup:
- Phase 5 through Phase 9 implementation plans are present.
- Phase 10 is registered as `planned` in `tools/phase_registry.json`.
- `tools/phase10_test_map.json` exists with every authoritative Phase 10 row marked `claim_status=blocked`.
- `docs/testing/phase10_coverage_ledger.md` must be generated from the manifest and must not be hand-edited.
- `configs/dev/config.toml` declares `[roots.backup_storage]`.
- Config validation already requires `roots.backup_storage` and validates filesystem-root overlap; Phase 10 still needs direct AC-403 evidence.
- No operational `backup_set`, `backup_attestation`, restore, or restore-verification implementation exists.
- No public backup, restore, or restore-verification route family is currently registered. This remains support context only until direct Phase 10 evidence captures it.
- TODO(Phase10-retained-root): replace with exact retained result root, run ID, run root, target, row ID, and artifact paths only after direct Phase 10 evidence exists.

## Phase Objective

By Phase 10 exit, a deployment-local operator must be able to inspect the latest successful retained backup set, prove retention and verification metadata, restore the latest successful retained backup into a fresh environment, rebuild projections before readiness, verify restored incident/workbook behavior, and fail closed when required backup artifacts or integrity proof are absent.

User-observable exit state:
- The latest successful retained `backup_set` exposes `backup_set_id`, `consistency_point_at`, `created_at`, `retained_until`, `verification_state`, `last_verified_restore_at`, `postgres_restore_anchor`, and `object_store_restore_anchor`.
- `verification_state` is exactly one of `unverified`, `verified`, or `failed`; `last_verified_restore_at` is null only while unverified.
- Restore selects exactly one retained `backup_set` and uses the same declared point for Postgres and object storage.
- A restored deployment becomes ready only after Postgres restore, object-store restore, and projection rebuild complete.
- Missing required artifacts or missing checksum/integrity proof prevents readiness.
- Restore verification runs in an isolated environment and records verified or failed truthfully.
- Public HTTP and WebSocket inventories expose no backup, restore, or restore-verification route family.
- Any built-in control surface is deployment-local, operator-facing, requires `deployment_admin`, and is not incident-scoped.
- `roots.backup_storage` is present, uses only allowed binding kinds, and is not satisfied by export or temporary-work roots.

## Implementation Scope

In scope:
- Phase 10 ownership manifest, generated coverage ledger, schedule-drift behavior, and explicit blocker sentinels until implementation rows exist.
- Persistent retained-backup metadata and restore anchors.
- `verification_state` vocabulary, state transitions, verification timestamps, and retention floors.
- Backup storage binding through `roots.backup_storage`.
- Restore-readiness orchestration over Postgres, object storage, and projection rebuild.
- Fail-closed integrity proof and missing-artifact behavior.
- Isolated restore-verification cadence and state transitions.
- Deployment-local operator inspection/control boundaries.
- Service-backed and browser/workbook recovery evidence.

Out of scope:
- Treating this plan, prior plans, generated ledgers, schedules, support-only tests, visual goldens, or retained artifacts as behavior authority.
- Incident Portability Extension Profile semantics, whole-incident bundles, import/export portability, snapshot/reporting routes, release publication, or benchmark publication.
- Arbitrary cross-store point-in-time restore to operator-supplied timestamps beyond the owner-required latest successful retained `backup_set`.
- Treating projections as authoritative backup inputs.
- Raw object-store browser access.
- Destructive restore or cleanup outside declared isolated test or restore-verification environments with explicit proof predicates.

Owner anchors:
- Core 01 section 12.1 Backup.
- Core 01 section 12.2 Restore.
- Core 04 section 2 deployment-admin authorization boundary.
- Core 04 section 6 runtime roots.
- Core 04 section 9.14 AC-398 through AC-403.
- Core 04 section 12.3.3 backup storage binding.
- Testing Harness NLSpec sections 1 through 4, 6 through 9, 11, 13, and 15 for public Make invocation, target selection, artifact identity, generated-artifact handling, service-backed fixtures, failure classification, cleanup predicates, output modes, and retained result roots.

## Interfaces And Data Model

Planned internal/deployment-local additions:
- Add persistent operational recovery state, likely `backup_sets` or `backup_attestations`, with `backup_set_id`, `consistency_point_at`, `postgres_restore_anchor`, `object_store_restore_anchor`, `created_at`, `retained_until`, `verification_state`, and `last_verified_restore_at`.
- Enforce exact `verification_state` vocabulary: `unverified`, `verified`, `failed`.
- Add a backup storage adapter bound only through `roots.backup_storage`.
- Add restore and restore-verification orchestration that can run only in declared isolated restore/test environments and records destructive proof predicates before mutation.
- Add an all-required-projection rebuild coordinator for restore readiness.
- Add deployment-local operator inspection/control only when needed; do not add public `/api/v1/backups*`, `/api/v1/restores*`, or `/api/v1/restore-verifications*` families.

Generated boundary:
- Do not hand-edit `internal/gen/**`.
- Do not hand-edit `packages/protocol-ts/src/generated/**`.
- Do not hand-edit generated coverage ledgers or generated schedules.
- Do not hand-edit `go.sum` or `pnpm-lock.yaml`.
- If schema changes land, add numbered migrations and keep codegen drift separate from migration drift.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers |
| --- | --- | --- | --- |
| [x] | 0. Ownership manifest and harness setup | `make phase-map-check`, `make explain-phase PHASE=phase10` | Phase 10 is planned, not executable, while blocker sentinels remain. |
| [ ] | 1. Backup storage root and deployment configuration | `make backend-unit`, `make backend-process` | AC-403 needs direct Phase 10 evidence. |
| [ ] | 2. Backup metadata and retention floors | `make backend-store`, `make migration-drift` | No backup schema exists. |
| [ ] | 3. Restore readiness and coherent stores | `make backend-process`, `make service-backed-slice PHASE=phase10` | Restore orchestration absent. |
| [ ] | 4. Fail-closed integrity handling | `make backend-process` | Integrity proof contract absent. |
| [ ] | 5. Isolated restore verification | `make backend-process` | Verification cadence/state machine absent. |
| [ ] | 6. Route absence and deployment-admin boundary | `make backend-unit`, `make backend-process` | Operator surface must remain deployment-local. |
| [ ] | 7. Service-backed and workbook recovery evidence | `make backend-integration`, `make browser-e2e-webserver-backed` | Requires real backup/restore substrate. |
| [ ] | 8. Final public wrappers, drift gates, finalizer, check, handoff | public Phase 10 wrappers, `make agent-finalize`, `make check` | TODO(Phase10-final-roots): record final retained roots. |

## Evidence Layer Matrix

Every authoritative Phase 10 row must have exactly one authoritative owner in `tools/phase10_test_map.json`. Support-only tests may be listed as support evidence, but support-only evidence must not satisfy authoritative rows.

| Row | Evidence layer | Claim intent |
| --- | --- | --- |
| `U-10-01` | `backend_store` | Metadata, vocabulary, anchors, and retention floors require persistent state. |
| `U-10-02` | `backend_process` | Restore readiness needs real Postgres/object-store orchestration and readiness gating. |
| `U-10-03` | `backend_process` | Missing artifacts/proofs and verification transitions must fail before target readiness. |
| `U-10-04` | `backend_unit` | Public route inventory absence and deployment-local guard registration can be tested without services. |
| `U-10-05` | `backend_unit` | Deployment configuration binding and root separation are config-contract checks. |
| `I-10-01` | `backend_integration` | Real backing-storage metadata and verification transitions. |
| `I-10-02` | `backend_process` | Fresh environment restore, projection rebuild, incident open, workbook query, and consistency proof. |
| `I-10-03` | `backend_process` | Broken retained backup must fail before readiness in an isolated environment. |
| `E-10-01` | `backend_process` | Deployment-local operator inspection without public API exposure. |
| `E-10-02` | `browser_functional` | Restored workbook surface and built-in workbook query through the real web stack. |
| `E-10-03` | `backend_process` | Public HTTP/WebSocket inventory from a running deployment. |
| `E-10-04` | `backend_process` | Effective deployment configuration rejects invalid backup-root binding. |

## Sprint 0. Ownership Manifest And Harness Setup

Objective: establish Phase 10 traceability without making false product claims.

Relevant IDs: all Phase 10 rows.

Files and areas:
- `tools/phase_registry.json`
- `tools/phase10_test_map.json`
- `docs/testing/phase10_coverage_ledger.md`
- `internal/modules/recovery/phase10_recovery_sentinel_test.go`
- `cmd/server/main_phase10_recovery_sentinel_test.go`
- `apps/web/e2e/phase10.restore.spec.ts`

Test-first sequence:
1. Add every Phase 10 guide row to `tools/phase10_test_map.json`.
2. Mark every authoritative row `claim_status=blocked` until direct evidence replaces the sentinel.
3. Keep Phase 10 `planned` in `tools/phase_registry.json`.
4. Generate the Phase 10 ledger through `make phase-ledgers`; do not hand-edit the ledger.
5. Do not run destructive backup, restore, cleanup, or environment-reset operations for Sprint 0.

Implementation tasks:
- Add the planned registry entry with manifest and ledger paths.
- Add blocked sentinel tests that skip at runtime and therefore do not break broad existing gates.
- Keep support-only carryover out of the Phase 10 row inventory.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase10`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `make target-plan-json`
- `git diff --check`

Deliverables:
- `tools/phase10_test_map.json`
- Updated `tools/phase_registry.json`
- Generated `docs/testing/phase10_coverage_ledger.md`
- Non-claiming blocker sentinel tests.

Risks:
- Phase 10 must not become active until executable row owners are coherent.
- Skipped sentinels are traceability placeholders only; they are not product evidence.

Exit criteria:
- Planned Phase 10 passes phase-map validation.
- The generated ledger renders from the manifest.
- `make phase-slice PHASE=phase10` is not described as passing while Phase 10 is planned and blocked.

## Sprint 1. Backup Storage Root And Deployment Configuration

Objective: prove `roots.backup_storage` is required, has only allowed bindings, and is distinct from export and temporary-work roots.

Rows: `U-10-05`, `E-10-04`.

Files and areas:
- `internal/platform/config`
- `internal/app`
- `configs/dev/config.toml`
- config golden diagnostics
- startup preflight tests

Test-first sequence:
1. Add Phase 10 config tests for missing `roots.backup_storage`.
2. Add allowed binding tests for disconnected, on-prem, and cloud profiles.
3. Add rejection tests for using `roots.export_outputs` or `roots.temporary_work` as backup storage.
4. Add effective-startup evidence for invalid backup-root configuration.

Implementation tasks:
- Tighten config validation only if current overlap checks do not satisfy AC-403 directly.
- Add clear diagnostics for attempted export/temp backup-root substitution.

Validation commands:
- `make backend-unit`
- `make backend-process`
- `make phase-map-check`
- `git diff --check`

Deliverables:
- Direct AC-403 row evidence.
- Updated manifest rows replacing blocker sentinels for `U-10-05` and `E-10-04`.

Risks:
- Existing Phase 0 runtime-root evidence is support only and cannot satisfy Phase 10 rows.

Exit criteria:
- `roots.backup_storage` AC-403 behavior has direct Phase 10 evidence.

## Sprint 2. Retained Backup Metadata And Retention Floors

Objective: create and inspect durable backup metadata without exposing public backup routes.

Rows: `U-10-01`, `I-10-01`, `E-10-01`.

Files and areas:
- `db/migrations`
- `db/queries`
- `internal/modules/recovery`
- `internal/platform/postgres`
- `internal/platform/objectstore`
- runtime assembly and deployment-admin patterns

Test-first sequence:
1. Assert `backup_set`/`backup_attestation` fields and restore anchors.
2. Assert exact `verification_state` vocabulary.
3. Assert null rule for `last_verified_restore_at`.
4. Assert latest-success no older than 24 hours.
5. Assert each successful backup has `retained_until >= created_at + 30 days`.

Implementation tasks:
- Add backup attestation schema and store.
- Add latest successful retained backup lookup.
- Add retention-floor enforcement.
- Add deployment-local operator inspection path that is not a public backup route family.

Validation commands:
- `make backend-store`
- `make backend-integration`
- `make migration-drift`
- `make generate-drift`
- `git diff --check`

Deliverables:
- Metadata and retention evidence for AC-398.
- Updated manifest rows replacing relevant blockers.

Risks:
- Do not treat a metadata row as successful unless both Postgres and object-store anchors plus attestation are durable for the same point.

Exit criteria:
- Latest successful retained backup metadata is inspectable without public backup routes.

## Sprint 3. Restore Readiness And Coherent Stores

Objective: restore exactly one retained backup set into a fresh environment and expose readiness only after projection rebuild.

Rows: `U-10-02`, `I-10-02`.

Files and areas:
- `internal/platform/postgres`
- `internal/platform/objectstore`
- `internal/modules/projections`
- `tools/testservices`
- workbook query helpers

Test-first sequence:
1. Assert restore selects exactly one retained `backup_set`.
2. Assert Postgres and object-store restore use the same declared point.
3. Assert order: Postgres, object store, projections.
4. Assert readiness is withheld until projections rebuild.
5. Assert row/change/blob hash consistency after restore.

Implementation tasks:
- Add restore runner for isolated target environments.
- Add projection rebuild coordinator.
- Add consistency checks over authoritative rows, change sets, and blob hashes.

Validation commands:
- `make backend-process`
- `make service-backed-slice PHASE=phase10`
- `git diff --check`

Deliverables:
- Direct restore-readiness evidence.

Risks:
- Do not claim arbitrary cross-store point-in-time restore.
- Do not treat projection tables as authoritative backup inputs.

Exit criteria:
- Latest successful retained backup restores coherent authoritative state into a fresh environment.

## Sprint 4. Fail-Closed Integrity Proof Handling

Objective: fail restore before readiness when required artifacts or integrity proof are absent.

Rows: `U-10-03`, `I-10-03`.

Files and areas:
- backup artifact layout
- checksum or manifest proof handling
- failure taxonomy and readiness gating

Test-first sequence:
1. Missing Postgres artifact fails before readiness.
2. Missing object-store artifact fails before readiness.
3. Missing checksum or integrity proof fails before readiness.
4. Failed verification records `failed` with non-null `last_verified_restore_at` when verification ran.

Implementation tasks:
- Define proof verification before restore exposure.
- Preserve failed verification state truthfully.

Validation commands:
- `make backend-process`
- focused broken-artifact rows
- `git diff --check`

Deliverables:
- AC-400 direct evidence.

Risks:
- A failed restore must never be represented as ready or verified.

Exit criteria:
- No broken backup can produce a ready target environment.

## Sprint 5. Isolated Restore Verification Cadence And Transitions

Objective: verify retained backups in an isolated environment and transition state truthfully.

Rows: `U-10-03`, `I-10-01`, `I-10-02`.

Files and areas:
- scheduler/job shells
- testservices isolation
- clock controls
- backup mechanism and root-binding change hooks

Test-first sequence:
1. Successful verification sets `verified` and non-null timestamp.
2. Failed verification sets `failed` and non-null timestamp.
3. Verification runs at least every 7 days and after backup, database, object-store, or backup-storage binding changes.
4. Failed verification is never reported as ready.

Implementation tasks:
- Add restore verification orchestration.
- Add cadence trigger logic.
- Keep this out of Core 05 benchmark publication.

Validation commands:
- `make backend-process`
- `make service-backed-slice PHASE=phase10`
- `git diff --check`

Deliverables:
- AC-401 direct evidence.

Risks:
- Verification artifacts may carry incident data and must remain under protected runtime roots.

Exit criteria:
- Verification state reflects actual isolated restore result.

## Sprint 6. Public Route Absence And Deployment-Admin Boundary

Objective: prove public route absence and deployment-local operator authorization.

Rows: `U-10-04`, `E-10-03`.

Files and areas:
- `internal/app/runtime.go`
- `internal/platform/httpapi`
- module route registrars
- WebSocket routes
- extension family registry

Test-first sequence:
1. Assert public `/api/v1/backups*`, `/api/v1/restores*`, and `/api/v1/restore-verifications*` are absent.
2. Assert matching `/ws/v1/*` route families are absent.
3. Assert any built-in operator control requires `deployment_admin`.
4. Assert controls are not incident-scoped.

Implementation tasks:
- Add deployment-local operator guard tests.
- Avoid incident-scoped backup/restore routes.

Validation commands:
- `make backend-unit`
- `make backend-process`
- `make browser-e2e-webserver-backed` only if browser-visible operator checks exist
- `git diff --check`

Deliverables:
- AC-402 direct evidence.

Risks:
- Do not invent public backup, restore, or restore-verification route families.

Exit criteria:
- Public HTTP and WebSocket inventories expose no recovery families.

## Sprint 7. Service-Backed And Browser Workbook Recovery Evidence

Objective: prove restored workbook usability from a fresh deployment.

Rows: `E-10-02`, support for `I-10-02`.

Files and areas:
- workbook E2E harness
- built-in query helpers
- incident seed data
- reset and isolation predicates

Test-first sequence:
1. Restore latest successful retained backup into a fresh deployment.
2. Open at least one incident when incident data exists.
3. Execute at least one built-in workbook query.
4. Prove row/change/blob consistency.
5. Do not expose raw object-store access as browser evidence.

Implementation tasks:
- Add browser/process restore fixture with explicit isolated environment proof.
- Keep workbook proof through ordinary product surfaces.

Validation commands:
- `make backend-integration`
- `make browser-e2e-webserver-backed`
- `make service-backed-slice PHASE=phase10`
- `git diff --check`

Deliverables:
- AC-399 and AC-401 workbook recovery evidence.

Risks:
- Browser evidence must not depend on diagnostic object-store shortcuts.

Exit criteria:
- Restored workbook surface is usable after restore.

## Sprint 8. Final Public Wrappers, Drift Gates, Finalizer, Handoff

Objective: complete Phase 10 public verification and handoff.

Rows: all Phase 10 rows.

Files and areas:
- generated ledgers and schedules
- finalizer outputs
- retained result roots

Test-first sequence:
1. Replace all blockers with direct evidence or keep blockers that prevent completion claims.
2. Refresh generated artifacts through canonical commands.
3. Run final public wrappers and broader gates.
4. Record retained roots and blockers.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase10`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `make target-plan-json`
- focused owner targets
- `make phase-slice PHASE=phase10`
- `make service-backed-slice PHASE=phase10`
- `make generate-drift`
- `make migration-drift`
- `make agent-finalize`
- `make check`
- `git diff --check`

Deliverables:
- TODO(Phase10-final-evidence): replace with exact retained roots or explicit blockers.
- Updated manifest, ledger, schedules, and final handoff notes.

Risks:
- Outside-Phase-10 blockers must be recorded with exact target, artifact root, failing row/test, failure class, and out-of-scope rationale.

Exit criteria:
- All binary exit criteria below are met.

## Blocker Recording Rules

For every failed Phase 10 validation command, record:
- Exact command.
- Exact failing target or scheduler work unit.
- Result root, run ID, run root, and artifact path.
- Row ID or test title when applicable.
- Failure class and failure reason when exposed: `product`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, or `unknown`.
- Ownership: Phase 10-owned, harness-owned, infra-owned, or outside Phase 10.
- Minimum follow-up required to make the blocker actionable.
- Whether destructive cleanup or restore was skipped, blocked by proof predicates, or ran only inside an isolated declared environment.

## Binary Exit Criteria

Phase 10 is complete only when all are true:
- Phase 10 is in `tools/phase_registry.json` and selectable only after row owners exist.
- `tools/phase10_test_map.json` covers every authoritative Phase 10 guide row exactly once.
- `docs/testing/phase10_coverage_ledger.md` is generated from the manifest and is not hand-edited.
- Generated schedule artifacts are refreshed and drift-checked through canonical commands.
- All TODO markers are replaced with direct evidence, or remaining blocker sentinels explicitly prevent Phase 10 completion claims.
- Operational backup and restore remain deployment-local recovery behavior, not public incident routes and not incident-portability behavior.
- `roots.backup_storage` is validated by deployment configuration and is not satisfied by export or temporary-work roots.
- Restore of the latest successful retained `backup_set` restores Postgres and object-store contents from the same retained point, rebuilds projections, opens at least one incident when incident data exists, and executes at least one built-in workbook query.
- Missing backup artifacts or missing integrity proof fails before readiness.
- Restore verification runs in an isolated environment and records `verified` or `failed` truthfully.
- Public `/api/v1/*` and `/ws/v1/*` route inventories expose no backup, restore, or restore-verification family.
- Any built-in backup, restore, or restore-verification control surface is deployment-local, operator-facing, requires `deployment_admin`, and is not incident-scoped.
- `make phase-slice PHASE=phase10` passes after Phase 10 becomes active.
- `make service-backed-slice PHASE=phase10` passes after Phase 10 becomes active, unless the manifest intentionally declares no service-backed work and the wrapper reports an explicit no-op.
- `make agent-finalize` outcome is recorded.
- `make check` passes, or every blocker outside Phase 10 is recorded with exact target, artifact root, failing row or test, failure class where available, and out-of-scope rationale.
- `git diff --check` passes.

## Global References

Controlling row set:
- `docs/guides/cartulary_implementation_testing_guide.md`, section 12, Phase 10.
- `U-10-01` through `U-10-05`.
- `I-10-01` through `I-10-03`.
- `E-10-01` through `E-10-04`.

Owner anchors:
- `docs/spec/00_document_set_status_and_precedence.md`.
- `docs/spec/01_architecture_storage_and_view_contracts.md`, section 12.1 and section 12.2.
- `docs/spec/04_security_deployment_and_conformance.md`, section 2, section 6, section 9.14, and section 12.3.3.
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`, claim publication only.
- `docs/testing-harness-nlspec.md`, harness mechanics only.
- `docs/domain.md`, vocabulary support only.

Generated-boundary rules:
- Do not hand-edit `docs/testing/phase10_coverage_ledger.md`.
- Do not hand-edit `tools/scheduler_manifest.json` or `tools/execution_topology_render_index.json`; refresh through `make phase-schedules`.
- Keep codegen drift separate from migration drift.
