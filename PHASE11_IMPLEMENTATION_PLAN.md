# Phase 11 Implementation Plan

## Summary

This file is the execution roadmap, progress marker, and handoff aid for Cartulary Phase 11: extension profile testing hooks.

Phase 11 is not Base Profile conformance. Phases 0 through 10 are the base-profile implementation sequence. Phase 11 keeps future extension-profile work aligned with the current extension route families, profile claim boundaries, acceptance criteria, and harness mechanics.

Core 00 through Core 04 own implementation-conformance behavior. Core 05 owns claim-bearing timed or fixture-sensitive publication only and does not define ordinary runtime implementation behavior. This file is non-normative: it records execution status, remediation decisions, and handoff notes. Runtime behavior is owned by the normative core, authored contracts and migrations, implementation code, generated artifacts, and direct tests.

Current remediation state:
- Import and Snapshot and Reporting are the selected and claimed Phase 11 extension profiles.
- Reference Pack is selected and claimed for the full `profile:reference_pack` contract after the remediation work added real background jobs, minimum disconnected seed packs, refresh replay stability, operator progress/cancel UI, and direct evidence rows.
- Phase 11 now has an active shared manifest at `tools/phase11_test_map.json` with Import, Snapshot/Reporting, Reference Pack, and common-job rows.
- Incident Portability and Enterprise Authentication remain reserved and unclaimed.
- The remaining unimplemented profile sprints are open future implementation work. They are not closed, failed, or descoped by the Import, Snapshot/Reporting, or Reference Pack remediation.
- Helper-only Sprint 1 evidence is not used by itself to claim route parity; Import claim evidence is carried by the completed route-family implementation and direct Phase 11 rows.

Implemented remediation record:
- Owner/spec alignment was completed for the shared upload envelope and Import request reasons: valid JSON metadata that is not an object maps to route-family `request_not_object`, while `invalid_value` is owner-backed for scalar, type, and format validation.
- Common substrate is in place for upload-envelope parsing, durable common jobs, route-scoped cancel idempotency, terminal job summaries, and request-time authorization re-derivation.
- The selected Import route family is implemented for session create/read, unit list/read, preview, mapping approval, select, skip, and apply.
- Import discovery covers bounded CSV and XLSX used-range sources, source byte hashing, deterministic mapping fingerprints, provenance through apply, duplicate apply blocking, and durable Import state that is separate from job status.
- OpenAPI, generated Go/TypeScript contracts, SQL-derived generated code, phase manifest, generated phase ledger, generated schedules, and duration-maintenance inputs were refreshed through canonical targets.
- Phase 2, process, and browser reserved-extension expectations were adjusted to use an unclaimed profile root instead of Import now that Import is claimed.
- Retained-run finalizer maintenance was run after the successful `make check` root and the warm `check-service-backed` retained-run budget was raised to `120000ms` to match observed successful timing.
- Snapshot and Reporting now targets the `/api/v1/snapshots` and `/api/v1/releases` route families on the v3 remediation model: `cartulary.snapshot_export_model.v3`, `cartulary.source_boundary.v1:<sha256>` source-boundary tokens, explicit export-family collection, exact `cartulary.report.default@1` template contracts, closed `output_kind` and `release_scope` OpenAPI resource schemas, explicit `recipient_partition_refs`, durable `render_failed` release rows, route-scoped hidden-resource errors, async snapshot/release create jobs, generated OpenAPI/contract coverage, SQLC-backed reporting queries, approval/publish/invalidate actions, exact-shape route evidence, and Phase 11 direct evidence rows.
- Reference Pack now targets the `/api/v1/reference-packs` route family with local bundle import through the shared upload envelope, `activation_policy='staged_only'`, no auto-activation, file-hash idempotency, explicit activation/disable, genuinely backgrounded import/reverify/refresh jobs, activation-pointer-derived `active`, closed durable/public condition vocabulary, prior-active retention, local-only archive verification, the required fresh-disconnected three-pack floor, closed Reference Pack error and reason registries, generated OpenAPI/contract coverage, deployment-admin UI progress/cancel controls, and Phase 11 direct evidence rows.
- Historical retained finalizer roots that predate the explain-run tool-summary diagnostic cleanup remain historical evidence only. Current remediation handoff must cite fresh retained roots that are explainable through `make explain-run`.

Authority model:
- Future adopted Cartulary NLSpecs, if present and explicitly adopted by the repository authority process, govern first.
- Core 00 through Core 04 govern current implementation-conformance behavior.
- Core 05 governs only claim-bearing timed or fixture-sensitive publication.
- `docs/guides/cartulary_implementation_testing_guide.md` guides implementation planning, especially Phase 11.
- `PHASE11_IMPLEMENTATION_PLAN.md` is a planning artifact only.
- Testing Harness NLSpec governs only harness mechanics: public Make invocation, target selection, scheduling, fixture lifecycle, artifact identity, generated-artifact handling, service ownership, failure classification, cleanup predicates, output modes, and retained result roots.
- Generated ledgers, generated schedules, visual goldens, retained artifacts, prior summaries, support-only tests, and prior implementation plans are evidence or diagnostics only. They are not behavior authorities.
- `docs/domain.md` is vocabulary and concept support only. It does not replace owner specs.

## Phase Objective

By Phase 11 exit, Cartulary must have a profile-by-profile extension implementation plan and, when a profile is selected for implementation, the repository must have direct evidence that the profile's route family, durable resources, lifecycle semantics, idempotency, authorization, trust boundary, error registry, generated artifacts, and harness selection behavior satisfy the extension delta on top of a passing Base Profile.

Phase 11 work may be implemented profile by profile. A partial Phase 11 implementation may support one extension profile without claiming all five. Each extension profile is independently claimable only when its profile-specific Definition of Done passes.

## Claim and User-Observable Exit State

Base prerequisite state:
- A claimed extension profile requires a passing Base claim first.
- The implementation must continue to expose reserved but unclaimed extension families as unclaimed, not partially active.
- Aggregate ACs `AC-232..AC-236` must not be the sole proof for substantive runtime behavior. Every substantive requirement family needs direct non-aggregate evidence.
- Import remediation completed against a passing `make check` root, `.cartulary/test-results/20260523T161735Z-p3268642`. Future extension claims must refresh or identify current retained Base evidence through the canonical public wrappers and final gate policy.

Import claim exit state:
- Observable outcome: users can upload bounded CSV or XLSX sources through `POST /api/v1/import-sessions`, inspect import sessions and units, preview read-only source rows, approve exhaustive ordered mappings, select or skip units, and apply selected units without blocking ordinary workbook editing.
- Proof required: direct evidence for shared upload-envelope handling, source byte hashing, `assistant_profile='phase2_workbook_import_v1'`, import-session and import-unit resources, deterministic `mapping_fingerprint`, provenance, hostile workbook constraints, no forbidden auto-resolution during ingest, import family error and reason-code registries, route-scoped idempotency, long-running job summaries, generated artifact boundaries, and harness selection.

Snapshot and Reporting claim exit state:
- Observable outcome: users can create immutable snapshots, create releases from exact snapshot/template/redaction/output tuples, approve, publish, invalidate, and render self-contained recipient-specific outputs without changing live workbook authorization or withholding.
- Proof required: direct evidence for snapshot boundary selection and high-watermark handling, `release_scope` and `release_state` vocabularies, approval tuple binding, invalidation behavior, deterministic self-contained output, recipient-specific redaction at snapshot/render/release time, snapshot/release error and reason-code registries, idempotency, job summaries, generated artifacts, and harness selection.

Reference Pack claim exit state:
- Observable outcome: operators can list and read reference-pack versions, import local bundles through the shared upload envelope, verify staged candidates, activate explicitly, disable, reverify, refresh, and retain the prior active version on failure.
- Proof required: direct evidence for `activation_policy='staged_only'`, `pack_key`, `pack_kind`, `pack_version`, integrity metadata, verification method/result, durable condition vocabulary, derived `active`, disconnected bundle behavior, integrity/archive/path/content-screening failures, error and reason-code registries, idempotency, jobs, generated artifacts, and harness selection.
- Current status: selected and claimed for the full profile after remediation direct evidence and final validation. Operator progress/cancel controls are exposed only to deployment-admin sessions on the existing authenticated landing/admin surface.

Incident Portability claim exit state:
- Observable outcome: deployment administrators can export a full-fidelity whole-incident bundle, read the durable export descriptor, and import a checksum-verified bundle into an empty incident namespace without importing deployment-local login-capable administrative state.
- Proof required: direct evidence for deterministic bundle layout, authoritative-source-only export, fixed `history_mode='full'` and `blob_mode='full'`, optional sections and required capabilities, manifest and checksum verification, staged import under temporary-work root, no projections or deployment-local runtime state, rejection of clone/merge/identifier-remap/remote-fetch modes, error and reason-code registries, idempotency, jobs, generated artifacts, and harness selection.

Enterprise Authentication claim exit state:
- Observable outcome: users can discover enabled OIDC/SAML providers, start provider sign-in, complete OIDC authorization-code with PKCE `S256` and `nonce` or SAML SP-initiated ACS, receive the same opaque server-managed session family as base auth, and deployment admins can create, rotate, and retire enterprise auth bindings for existing local users.
- Proof required: direct evidence for provider discovery, begin, callback, ACS, no implicit or hybrid OIDC, no IdP-initiated SAML, no public durable auth-transaction resource, no auto-created local users or incident memberships, deployment-admin-only binding management, binding audit and session revocation on rotate/retire, forbidden secret/assertion retention, error and reason-code registries, route-specific idempotency where applicable, generated artifacts, and harness selection.

## Implementation Scope

In scope:
- Phase 11 ownership and traceability planning.
- Extension route-family parity under Core 01 §17 and §20.
- Core 01 §17.1 common parity rules and shared upload-envelope behavior for import, reference-pack import, and incident-bundle import.
- Profile-specific manifests, ledgers, schedules, and public-wrapper planning only when repository conventions require them.
- Extension profile claim manifests and direct evidence boundaries.
- Common job-resource behavior for long-running extension work.
- Route-scoped idempotency and divergent replay for mutating extension control-plane routes, except where Enterprise Authentication protocol routes are explicitly non-idempotent.
- Extension error-code and reason-code registries.
- Authorization re-derivation and deployment-admin-only boundaries where applicable.
- Trust boundaries for hostile files, export outputs, reference packs, portability bundles, and enterprise-auth provider data.
- Generated-artifact and drift planning for future owner-input changes.
- Profile-by-profile handoff notes that keep implemented profiles distinct from unimplemented or unclaimed profiles.

Out of scope:
- Treating Phase 11 as a Base Profile phase.
- Treating this plan, prior plans, generated ledgers, generated schedules, visual goldens, support-only tests, retained artifacts, or previous handoff notes as behavior authority.
- Implementing all extension profiles merely because Phase 11 exists.
- Adding extension routes without claiming and testing the corresponding extension profile.
- Reinterpreting clipboard paste as file-based import.
- Reinterpreting operational backup/restore as incident portability.
- Claim-bearing benchmarks or public timed/fixture-sensitive publication unless Core 05 publication requirements are separately satisfied.
- WebAuthn/passkeys, JIT provisioning, SCIM, generalized workflow engines, live-recipient-specific workbook withholding, cross-incident analytics, or any other future area not currently defined by owner sections.
- Treating `contracts/extensions/index.json`, `internal/gen/**`, or generated TypeScript protocol files as behavior owners.

Optional/profile-selected scope:
- A selected extension profile may get its own implementation sprint, manifest rows, generated ledger, schedule coverage, and final claim gate.
- A selected profile may be implemented and claimed while other Phase 11 profiles remain `not_started`.
- Phase 11 uses one shared manifest, `tools/phase11_test_map.json`, with profile-selected rows only. Current active rows cover Import, Snapshot/Reporting, Reference Pack, and common jobs.
- Existing public wrappers, `make phase-slice PHASE=phase11` and `make service-backed-slice PHASE=phase11`, are sufficient for the current active Import, Snapshot/Reporting, and Reference Pack claims.

## Owner Anchors

- Future adopted Cartulary NLSpecs, if present and explicitly adopted by the repository authority process.
- Core 00 §4.2 extension profiles and §5.1 owner matrix.
- Core 01 §17 extension route-family public contracts.
- Core 01 §17.1 common parity rules and shared upload-envelope contract.
- Core 01 §17.2 Import Extension Profile public contract.
- Core 01 §17.3 Snapshot and Reporting Extension Profile public contract.
- Core 01 §17.4 Reference Pack Extension Profile public contract.
- Core 01 §17.5 Incident Portability Extension Profile public contract.
- Core 01 §20 Enterprise Authentication Extension Profile public contract.
- Core 02 extension-owned provenance, snapshot/release, reference-pack, incident-bundle, and enterprise-auth state where applicable.
- Core 03 import assistant and workbook-startup reuse where applicable.
- Core 04 extension claim manifests and security/trust-boundary sections.
- Core 05 only for claim-bearing publication.
- `docs/guides/cartulary_implementation_testing_guide.md` Phase 11 as implementation-planning guidance.
- Testing Harness NLSpec only for command, scheduler, output, artifact, service, and failure-class mechanics.

## Extension Profile Matrix

| Profile | Claim prerequisite | Primary route family | Durable resource or state | Long-running job behavior | Trust boundary | AC delta | Planned claim status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Import | Passing Base claim plus `profile:import` requirements | `/api/v1/import-sessions` | `import_session`, `import_unit`, approved mapping, provenance, `mapping_fingerprint`, `source_content_sha256` | Discovery and apply use common job resource; terminal summaries emit exactly one `import_session` ref | Hostile CSV/XLSX source bytes, inert workbook behavior, bounded parser limits, imports module boundary | `AC-027..AC-029`, `AC-063..AC-067`, `AC-232`, `AC-262..AC-265`, `AC-323..AC-325`, `AC-393` | `selected_implemented` |
| Snapshot and Reporting | Passing Base claim plus `profile:snapshot_reporting` requirements | `/api/v1/snapshots`, `/api/v1/releases` | Immutable snapshot descriptor, release record, approval records, output hashes, redaction metadata | Snapshot create and release create use common job resource; approve/publish/invalidate are synchronous | Export output, self-contained assets, recipient redaction without live workspace withholding | `AC-030..AC-032`, `AC-056..AC-062`, `AC-071`, `AC-091`, `AC-104..AC-106`, `AC-113..AC-115`, `AC-233`, `AC-266..AC-269`, `AC-305..AC-307`, `AC-333` | `selected_implemented` |
| Reference Pack | Passing Base claim plus `profile:reference_pack` requirements | `/api/v1/reference-packs` | `reference_pack_version`, activation pointer, verification metadata, prior active version | Import, reverify, and refresh are jobs; activate and disable are synchronous | Hostile local pack bundles, disconnected operation, integrity and content screening | `AC-033..AC-035`, `AC-092..AC-096`, `AC-234`, `AC-270..AC-272`, `AC-308..AC-310`, `AC-326`, `AC-369` | `selected_implemented` |
| Incident Portability | Passing Base claim plus `profile:incident_portability` requirements | `/api/v1/incident-bundles` | Durable export descriptor, deterministic bundle, staged import state, imported incident | Export and import use common job resource; terminal summaries emit one bundle or incident ref | Whole-incident bundle integrity, temporary-work staging, encrypted roots, no deployment-local auth state | `AC-164..AC-169`, `AC-236`, `AC-273..AC-276`, `AC-327..AC-328`, `AC-332`, `AC-386`, `AC-409` | `not_started` |
| Enterprise Authentication | Passing Base claim plus `profile:enterprise_authentication` requirements | `/api/v1/auth/providers`, `/api/v1/auth/oidc`, `/api/v1/auth/saml`, `/api/v1/users/{user_id}/auth-bindings` | Enterprise provider config, server-side auth transaction, active auth binding, binding audit lineage | Protocol routes are intentionally non-idempotent and not job-backed; binding routes are synchronous idempotent mutations | Provider metadata, raw assertions, tokens, provider subjects, deployment-admin binding control | `AC-036`, `AC-235`, `AC-288..AC-293`, `AC-348..AC-352` | `not_started` |

## Evidence Layer Matrix

Phase 11 now has authoritative active rows in `tools/phase11_test_map.json`. Current adopted rows cover common jobs, Import, Snapshot/Reporting, and Reference Pack: `U-11-JOBS-01`, `U-11-JOBS-02`, `I-11-IMPORT-01`, `I-11-IMPORT-02`, `I-11-IMPORT-03`, `U-11-REPORTING-01..06`, `I-11-REPORTING-01..04`, `U-11-REFERENCE-PACK-01..06`, and `I-11-REFERENCE-PACK-01..07`. Other profile evidence categories remain planning-only until future profile rows are explicitly adopted.

| Profile | Backend unit evidence | Backend store evidence | Backend integration/process evidence | Browser/E2E evidence | Generated/drift evidence | Claim blockers |
| --- | --- | --- | --- | --- | --- | --- |
| Import | Upload-envelope parser, request normalization, mapping fingerprint, error registries, parser isolation guards | Import sessions/units, provenance, source hash, mapping persistence, duplicate-apply and overlap state | Real Postgres import lifecycle, CSV/XLSX bounded discovery, apply jobs, idempotent replay | No browser Import workflow is claimed in the current API-only evidence set | OpenAPI, generated Go/TS contracts, SQL-derived models, phase ledger, phase schedule, and duration baselines refreshed through generators | No open blocker for the current selected Import API route family; future UI evidence is separate and unclaimed |
| Snapshot and Reporting | Snapshot/release request validation, release-state guards, approval tuple, redaction policy selection, disclosure partition checks, error registries, generated OpenAPI closed enum/resource-shape assertions | Immutable snapshots, release records, approval records, output hashes, invalidation persistence, SQLC-backed reporting queries | Real Postgres snapshot/release lifecycle, self-contained output production, snapshot boundary high-watermark replay, approval/publish idempotency, singleton exact-shape and route-scoped visibility/auth assertions | No browser reporting workflow is claimed in the current API-only evidence set | OpenAPI, generated Go/TS contracts, SQL-derived models and queries, phase ledger, phase schedule, and duration baselines refreshed through generators | No open blocker for the current selected API route family; future UI evidence and richer template packs are separate work |
| Reference Pack | Upload envelope, activation-policy validation, state conflict registry, closed error/reason registries, verifier failure mapping, derived `active` rules | Pack metadata, verification result, active pointer, prior active version, durable conditions, job payloads and summaries | Bundle staging, verification, activation, disable, reverify, refresh, file-hash replay, divergent replay, disconnected seed fixtures, integrity/archive/path/content-screening failures | Deployment-admin landing/admin panel lists packs, imports bundles, runs actions, polls progress, and cancels non-terminal jobs | Contract/OpenAPI, SQL-derived models, TypeScript contracts, phase ledger, schedules, and duration baselines refreshed through generators/finalizers | No open blocker for the selected full profile |
| Incident Portability | Export/import request validation, bundle selector canonicalization, error registries, forbidden-mode rejection | Export descriptor, manifest checksum records, staged import bookkeeping, imported actor descriptors | Whole-incident export/import with Postgres/object store, projection rebuild, missing-file/checksum failures | Export/import progress and imported incident open if product UI exists | Contract/OpenAPI, bundle fixture generation, ledgers/schedules through generators | Base claim freshness; destructive import fixture isolation TODO; no route implementation |
| Enterprise Authentication | Provider discovery, begin validation, callback/ACS rejection, binding request validation, error registries | Provider config, auth transaction, binding lifecycle, active uniqueness, audit lineage | OIDC/SAML protocol simulation, session issuance, session revocation on rotate/retire | Provider sign-in redirect/callback/ACS and post-login landing behavior if harness can simulate IdP | Contract/OpenAPI and generated protocol drift if route schemas change; ledgers/schedules through generators | Base claim freshness; IdP fixture strategy TODO; no route implementation |

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers | Follow-up notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Phase 11 ownership model, profile-selection policy, and harness setup | `make phase-map-check`, `make explain-phase PHASE=phase11`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-test-name-check` | None. | Sprint 0 is historical. Phase 11 is active because selected-profile rows were adopted. |
| [x] | 1. Common extension substrate and upload-envelope foundation | `go test ./internal/platform/jobs ./internal/modules/jobapi ./internal/modules/imports ./internal/modules/imports/tabularingest ./internal/platform/httpapi ./internal/app`; `make generate`; `make phase-slice PHASE=phase11` | None for common substrate. | Helper evidence remains substrate only; Import route-family evidence in Sprint 2 carries the claim. |
| [x] | 2. Import Extension Profile | `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, `make check` | None for the selected API route family. | Import is selected and claimed; no browser Import workflow is claimed. |
| [x] | 3. Snapshot and Reporting Extension Profile | `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, latest post-remediation `make test-fast`; historical Sprint 3 `make check` root recorded below | None for the selected API route family. | Covers snapshots, releases, approval tuple, state machine, rendering, invalidation, redaction boundary, generated contracts, exact shapes, and route-scoped visibility/auth. |
| [x] | 4. Reference Pack Extension Profile | `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, `make check` | None for the selected full profile. | Covers pack import, verification, activation, disable, reverify, refresh, disconnected bundle constraints, operator progress/cancel UI, generated artifacts, and direct evidence rows. |
| [ ] | 5. Incident Portability Extension Profile | profile-selected targets | Pending implementation. | Covers export/import bundle layout, checksums, authoritative source state, blob/history preservation, no deployment-local admin import. |
| [ ] | 6. Enterprise Authentication Extension Profile | profile-selected targets | Pending implementation. | Covers providers, begin/callback/ACS, same session family, binding management, no auto-provisioning. |
| [ ] | 7. Profile claim gates, generated artifacts, finalizers, and handoff | `make check`, `make agent-finalize`, drift gates, profile wrappers as applicable | Open until remaining profile sprints are implemented or explicitly rescoped. | Import finalizer evidence is recorded below; full Phase 11 closeout remains open. |

## Sprint 0. Phase 11 Ownership Model, Profile-Selection Policy, And Harness Setup

Objective: Establish Phase 11 traceability and profile-selection policy without making product claims.

Relevant IDs:
- Extension profile selectors: `profile:import`, `profile:snapshot_reporting`, `profile:reference_pack`, `profile:incident_portability`, `profile:enterprise_authentication`.
- Aggregate claim gates: `AC-232`, `AC-233`, `AC-234`, `AC-235`, `AC-236`.
- Active authoritative Phase 11 rows now exist for Import, Snapshot/Reporting, Reference Pack, and common jobs: `U-11-JOBS-01`, `U-11-JOBS-02`, `I-11-IMPORT-01`, `I-11-IMPORT-02`, `I-11-IMPORT-03`, `U-11-REPORTING-01..06`, `I-11-REPORTING-01..04`, `U-11-REFERENCE-PACK-01..06`, and `I-11-REFERENCE-PACK-01..07`.

Files and areas:
- `tools/phase_registry.json`
- `tools/phase11_test_map.json` is the active shared Phase 11 manifest and contains Import, Snapshot/Reporting, Reference Pack, and common-job rows.
- `docs/testing/phase11_coverage_ledger.md` is generated from the active shared manifest.
- `scripts/check-phase-maps.sh`, `scripts/render-phase-ledgers.mjs`, `scripts/render-execution-topology-artifacts.mjs`, and `scripts/check-phase-test-names.mjs`.
- `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, and generated schedule artifacts define the public wrapper and service-backed selection behavior.

Test-first sequence:
1. Sprint 0 first registered Phase 11 as planned and non-claiming.
2. Remediation adopted only the direct Import, Snapshot/Reporting, Reference Pack, and common-job rows required for the selected profiles.
3. Incident Portability and Enterprise Authentication remain `not_started`.
4. Generated ledgers and schedules were refreshed only through canonical commands.
5. Unselected profiles remain unclaimed and continue to return reserved-family `extension_profile_not_claimed`.

Implementation tasks:
- Register Phase 11 as a planned, non-executable phase before adopting manifest rows. Complete.
- Encode profile-selection policy in this handoff and active shared manifest. Complete.
- Keep aggregate ACs separate from direct behavior rows.
- Mark every profile as independently claimable only after its own Definition of Done passes.

Sprint 0 policy:
- Sprint 0 is historical. Phase 11 was originally registered with `status: planned`; after Import remediation it is active with a shared manifest.
- The current registry schema supports one manifest path per phase, so Phase 11 uses one shared `tools/phase11_test_map.json` with profile-selected rows.
- Profile-specific manifests are not supported by the current registry schema because each phase entry has one `manifest_path` and the path must end in `phaseN_test_map.json`.
- The current phase manifest row schema does not support `not_started`; row `claim_status` values are limited to `implemented`, `blocked`, and `not_applicable`. Unselected profile status remains recorded here until profile rows are adopted.
- Planning rows for future profiles remain in this file only. Blocker rows may be added to the shared manifest only when the blocker must participate in harness selection or generated ledgers.
- Skipped rows are not claim evidence. Future non-executable profile work should remain absent from the active manifest unless an intentional blocker row is needed.
- Claimable profile evidence must be direct, profile-selected, and non-aggregate. Aggregate ACs `AC-232..AC-236` remain profile claim gates and must not substitute for direct runtime behavior evidence.
- Historical Sprint 0 profile statuses were all `not_started`. After remediation, Import, Snapshot/Reporting, and Reference Pack are selected and implemented; Incident Portability and Enterprise Authentication remain `not_started`.
- Valid current public wrappers: `make explain-phase PHASE=phase11`, `make phase-slice PHASE=phase11`, and `make service-backed-slice PHASE=phase11`.
- Historical planned-phase invalid-command behavior was normalized for clarity, but it no longer applies to current `phase11` because the phase is active.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase11`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `git diff --check`

Deliverables:
- Phase 11 manifest policy: one shared manifest with selected-profile rows only.
- Generated ledger and schedule status: Phase 11 ledger and schedule artifacts are generated from `tools/phase11_test_map.json`.
- Updated plan notes record Import and Snapshot/Reporting as selected/claimed and all other profiles as unclaimed.

Risks and assumptions:
- The previous `make explain-phase PHASE=phase11` unknown-phase failure was expected while Phase 11 was unregistered. After Sprint 0 registration, this command is expected to pass and report `status: planned`.
- Current `make explain-phase PHASE=phase11` reports active Import, Snapshot/Reporting, Reference Pack, and common-job coverage, not planned coverage.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.
- A broad Phase 11 manifest must not imply all five profiles are implemented.

Exit criteria:
- Harness policy is explicit.
- Profile claim statuses are not ambiguous.
- Every profile-selected row has a direct evidence plan or a blocker.
- No generated artifact is edited by hand.

## Sprint 1. Common Extension Route Parity And Upload Envelope

Objective: Build shared extension route-family substrate only for profiles selected for implementation.

Sprint 1 execution status:
- Sprint 1 final scope is helper substrate only. It is not route-parity evidence by itself.
- Current selected profile set in the default registry after remediation: Import, Snapshot/Reporting, and Reference Pack. Import is claimed by Sprint 2 route-family evidence; Snapshot/Reporting is claimed by Sprint 3 route-family evidence; Reference Pack is claimed by Sprint 4 route-family evidence; Incident Portability and Enterprise Authentication remain unclaimed.
- Phase 11 manifest policy after remediation: one shared active `tools/phase11_test_map.json` with Import, Snapshot/Reporting, Reference Pack, and common-job rows. Generated ledgers and schedules are produced from that manifest.
- Implemented helper-only substrate: `internal/platform/httpapi.ParseUploadEnvelope` with `UploadEnvelopePolicy`, `UploadEnvelope`, `UploadEnvelopeError`, closed shared reason-code helpers, route-local file media-type allowlists, metadata duplicate-key rejection, metadata BOM/non-UTF-8 rejection, nested multipart rejection, and SHA-256 calculation for accepted file bytes.
- Common job substrate: `/api/v1/jobs/{job_id}` and `/api/v1/jobs/{job_id}/cancel` are listed in the OpenAPI route inventory and backed by durable job storage, route-scoped cancel idempotency, request-time job authorization, terminal summaries, retention timestamps, and incident-scoped `job_progress` publication.
- Import substrate status: Sprint 1 introduced the helper and scaffold surface; Sprint 2 completes the selected Import route family.
- Selected-profile evidence: mapping, select, skip, apply, bounded XLSX used-range discovery, apply provenance through workbook/timeline writes, OpenAPI schemas, and Phase 11 manifest rows are adopted for Import.

Relevant IDs:
- Common parity: `AC-262..AC-276` where profile-selected.
- Import upload envelope: `AC-262`, `AC-265`.
- Reference-pack upload envelope: `AC-270`, `AC-272`.
- Incident-bundle upload envelope: `AC-275`, `AC-276`.
- Common job/resource behavior: profile-selected `AC-262`, `AC-264`, `AC-266`, `AC-267`, `AC-268`, `AC-270`, `AC-271`, `AC-273`, `AC-274`, `AC-275`, `AC-309`, `AC-369`.

Files and areas:
- `internal/platform/httpapi` for extension discovery and reserved-family dispatch.
- `internal/platform/httpapi/upload_envelope.go` for shared upload-envelope parsing and route-family error translation.
- `internal/platform/jobs` and `internal/modules/jobapi` for durable common job behavior and job HTTP routes.
- `internal/modules/imports` for the selected Import route family.
- `contracts/openapi/cartulary.openapi.yaml` for common job and Import route contracts.
- Future not-yet-implemented profile modules remain unclaimed and are not made public by Sprint 1 substrate work.

Test-first sequence:
1. Add shared upload-envelope tests for exact multipart requirements before adding profile-specific route behavior.
2. Add common job-result summary tests for long-running extension actions selected for the profile.
3. Add route-scoped idempotency and divergent replay tests for each mutating control-plane route that owns `client_txn_id`.
4. Add family error and reason-code registry tests before handler implementation.
5. Add authorization re-derivation tests for incident-scoped and deployment-admin routes.

Implementation tasks:
- Implement shared upload-envelope parsing for exactly `metadata` and `file`, UTF-8 BOM-free JSON metadata, duplicate-key rejection, route-local file media-type allowlists, and early-fail no-durable-state behavior.
- Reuse the common job resource for long-running actions and keep durable family states separate from job-status tokens.
- Keep reserved but unclaimed route families unchanged until the selected profile is claimed and tested.
- Keep family-specific error codes and reason codes closed.

Validation commands:
- Targeted backend unit tests for upload envelopes and route parity.
- Targeted backend store/process tests for idempotency and jobs.
- `make phase-slice PHASE=phase11` once active with an adopted manifest.
- `make service-backed-slice PHASE=phase11` once active and profile rows require services.
- `make generate-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, and `git diff --check` when owner inputs change.

Deliverables:
- Shared upload-envelope helper with direct tests: complete for shared helper scope.
- Closed shared upload-envelope reason-code registry evidence: present for shared upload reasons and Import request/source reason families.
- Common job resource evidence: targeted jobs tests cover durable creation, cancellation, idempotent cancel replay, terminal summaries, and retention fields.
- Import route substrate evidence: targeted integration tests cover route-level upload failure with zero import/job/idempotency rows, exact multipart replay, divergent replay rejection, durable session/unit reads, preview reads, pagination metadata, and discovery-job terminal summary.
- Full selected-profile route parity evidence: complete for Import through the Sprint 2 remediation rows.
- Explicit non-claim status for not-yet-implemented profiles: Incident Portability and Enterprise Authentication remain unclaimed in the default extension registry.

Risks and assumptions:
- Enterprise Authentication protocol routes are intentionally non-idempotent and not upload-envelope routes.
- Implementing shared helpers without any selected profile must not expose public extension behavior.

Exit criteria:
- Common substrate passes targeted direct and service-backed tests.
- Unselected profile roots remain reserved and unclaimed.
- Generated contract artifacts are refreshed through `make generate` after the error registry update.
- Import Sprint 2 evidence carries the selected-profile route-family claim.

Historical Sprint 1 validation record before remediation:
- `go test ./internal/platform/jobs ./internal/modules/jobapi ./internal/modules/imports ./internal/modules/imports/tabularingest ./internal/platform/httpapi ./internal/app`: passed.
- `make generate`: passed and refreshed contract/sql-derived generated artifacts.
- `make agent-finalize`: passed with `generated=unchanged`, retained-run maintenance skipped because `RESULTS_DIR` was unset.
- `make explain-phase PHASE=phase11`: passed; Phase 11 remains planned, `coverage: none`, and has no execution map.
- `make phase-slice PHASE=phase11`: failed as expected with `phase phase11 is planned and is not executable`.
- `make service-backed-slice PHASE=phase11`: failed as expected with `phase phase11 is planned and is not executable`.
- `make generate-drift`: passed after generation.
- `make migration-drift`: passed after adding Phase 11 job and import-session migrations.
- `make phase-ledger-drift`: passed; no Phase 11 ledger adopted.
- `make phase-schedule-drift`: passed; no Phase 11 schedule adopted.
- `make phase-test-name-check`: passed; support tests do not use `TestPhase11...` names before a Phase 11 manifest exists.
- `make test-fast`: passed.
- `make check`: passed, run root `.cartulary/test-results/20260523T021024Z-p1860583`.
- `git diff --check`: passed.

## Sprint 2. Import Extension Profile

Objective: Implement and prove the Import Extension Profile only when selected.

Sprint 2 execution status:
- Complete for the selected Import API route family.
- At Sprint 2 close, default extension discovery claimed Import and kept Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication unclaimed. Sprint 3 later claimed Snapshot/Reporting, and Sprint 4 later claimed Reference Pack.
- Implemented routes: `POST /api/v1/import-sessions`, `GET /api/v1/import-sessions/{import_session_id}`, `GET /api/v1/import-sessions/{import_session_id}/units`, `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}`, `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview`, `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping`, `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select`, `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip`, and `POST /api/v1/import-sessions/{import_session_id}/apply`.
- CSV discovery, XLSX bounded used-range discovery, mapping approval, select, skip, apply, duplicate apply blocking, terminal job summaries, and durable session/unit state are covered by active Phase 11 rows.
- XLSX table, named-range, and manual-region locators are not claimed by the current evidence set.

Relevant IDs:
- `AC-027..AC-029`
- `AC-063..AC-067`
- `AC-232`
- `AC-262..AC-265`
- `AC-323..AC-325`
- `AC-393`

Files and areas:
- `internal/modules/imports/api.go`, `internal/modules/imports/routes.go`, `internal/modules/imports/store.go`, `internal/modules/imports/xlsx.go`, and `internal/modules/imports/imports_integration_test.go`.
- `internal/modules/imports/tabularingest/tabularingest.go` and `internal/modules/imports/tabularingest/tabularingest_test.go`.
- `db/migrations/00023_phase11_jobs.sql` and `db/migrations/00024_phase11_import_sessions.sql`.
- `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, generated Go protocol code, generated TypeScript protocol code, and SQL-derived generated code.
- `tools/phase11_test_map.json`, `docs/testing/phase11_coverage_ledger.md`, generated schedules, and duration-maintenance baselines.

Test-first sequence:
1. Prove `POST /api/v1/import-sessions` shared upload envelope, exact file media types, default and explicit `assistant_profile='phase2_workbook_import_v1'`, `source_content_sha256`, no durable state on envelope failure, and idempotent replay.
2. Prove import-session and import-unit reads, pagination rules, exact empty arrays, absent/present `mapping_fingerprint`, preview shape, closed `cell_kind`, source order, and first-50-row cap.
3. Prove `PUT .../mapping` exhaustive ordered `source_columns[]`, deterministic `mapping_fingerprint`, invalid mapping failures, and approved mapping readback.
4. Prove select, skip, and apply behavior, including persisted `selected_unit_ids[]`, no-op select/skip, skipped mapping preservation, omitted `selected_unit_ids[]`, apply job summaries, and durable state not using job tokens.
5. Prove CSV and bounded XLSX discovery, inert formulas/macros/external links, unsupported workbook handling, no forbidden auto-resolution during ingest, duplicate apply blocking, overlap blocking, deterministic apply order, and provenance.

Implementation tasks:
- Add the route family exactly under `/api/v1/import-sessions`.
- Keep parser, discovery, workbook downgrade warnings, XLSX/OpenXML dependencies, and workbook heuristics isolated in `imports`.
- Persist import sessions, units, mapping plans, source hashes, parser profile/version, provenance, warning codes, blockers, and terminal states.
- Apply mappings through the stable tabular-ingest contract and shared mapping engine.
- Preserve unknown columns with source-unit and ordinal identity.
- Enforce import and archive resource limits.
- Maintain closed import error and reason-code registries.

Validation commands:
- Profile-selected backend unit/store/integration tests.
- Browser or E2E tests for upload/progress/preview/mapping/select/skip/apply if UI is exposed.
- `make phase-slice PHASE=phase11`
- `make service-backed-slice PHASE=phase11`
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `git diff --check`.

Deliverables:
- Import profile route family and durable resources.
- Direct non-aggregate evidence for every Import delta family.
- Claim manifest rows or profile-selected equivalent.
- Updated claim status only after proof passes.

Risks and assumptions:
- Clipboard paste remains Base Profile behavior and must not be reinterpreted as file-based import.
- Imported host or account aliases must not auto-resolve where owner sections forbid it.
- XLSX fixtures must not execute workbook behavior.

Exit criteria:
- Passing Base claim evidence is identified.
- All Import delta ACs have direct evidence.
- `profile:import` is the only extension profile moved beyond `not_started` if other profiles remain unimplemented.

Sprint 2 remediation status:
- Import is selected and implemented.
- At Sprint 2 close, default extension discovery reported Import as claimed and left all other extension profiles unclaimed. Sprint 3 later moved Snapshot/Reporting into the selected/claimed set.
- At Sprint 2 close, the active Phase 11 manifest contained Import and common-job rows. The current manifest also includes Snapshot/Reporting rows.
- Direct tests cover non-object metadata early failure, CSV mapping/select/apply, XLSX used-range discovery, and incident/deployment common-job authorization re-derivation.

Sprint 2 validation record:
- `go test ./internal/platform/httpapi ./internal/platform/jobs ./internal/modules/jobapi ./internal/modules/imports ./internal/modules/imports/tabularingest ./internal/app`: passed.
- `make phase-slice PHASE=phase11`: passed; retained root `.cartulary/test-results/20260523T153803Z-p3120091`.
- `make service-backed-slice PHASE=phase11`: passed; retained root `.cartulary/test-results/20260523T154329Z-p3131974`.
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, and `make phase-test-name-check`: passed.

## Sprint 3. Snapshot And Reporting Extension Profile

Objective: Implement and prove the Snapshot and Reporting Extension Profile only when selected.

Status: complete for the selected API route family. Browser reporting workflows and richer future template packs remain unclaimed future work.

Relevant IDs:
- `AC-030..AC-032`
- `AC-056..AC-062`
- `AC-071`
- `AC-091`
- `AC-104..AC-106`
- `AC-113..AC-115`
- `AC-233`
- `AC-266..AC-269`
- `AC-305..AC-307`
- `AC-333`

Files and areas:
- Runtime and persistence: `internal/modules/reporting`, `internal/app/runtime.go`, `internal/platform/httpapi/extensions.go`, `db/migrations/00025_phase11_snapshot_reporting.sql`, `db/migrations/00026_phase11_reporting_remediation.sql`, and `db/queries/reporting_phase11.sql`.
- Owner/spec and contracts: Core 01, Core 02, Core 04, `contracts/openapi/cartulary.openapi.yaml`, and `contracts/errors/index.json`.
- Generated artifacts: `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `docs/testing/phase11_coverage_ledger.md`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, and duration baseline JSON files were refreshed through generators/finalizers only.

Test-first sequence:
1. Prove `POST /api/v1/snapshots` request validation, omitted high-watermark resolution at admission, replay using original boundary, queued/running/terminal job behavior, and exact snapshot read shape.
2. Prove `cartulary.snapshot_export_model.v3` collects current base-profile workbook surfaces, active relationship/evidence metadata, deterministic ordering, support refs, disclosure partition refs, and no raw blob bytes, blob hashes, storage refs, job records, auth/admin state, mutable history, or idempotency records.
3. Prove `POST /api/v1/releases` exact version selectors, `release_scope` defaulting to `internal_draft`, explicit canonical `recipient_partition_refs`, closed `output_kind` and `release_scope` vocabulary, no `latest/current`, render job summary, and exact release resource shape.
4. Prove durable `render_failed` release rows expose `render_failed_reason_code`, nullable output fields, terminal failed jobs, and state-conflict behavior for approve, publish, and invalidate.
5. Prove approve, publish, and invalidate action routes, approval tuple binding including recipient partitions, replay before fresh state checks, legal state transitions, singleton pagination rejection, action response parity with `GET /api/v1/releases/{release_id}`, admin-only publish/invalidate authorization, and exact `release_state` vocabulary.
6. Prove self-contained disconnected output and deterministic ordering/hashes.
7. Prove recipient-specific redaction profiles operate at snapshot/render/release time and do not change live workbook query results, field visibility, row visibility, or evidence visibility.
8. Prove generated OpenAPI references reusable closed enum schemas from both release create and durable release resource schemas, and keeps snapshot/release resource schemas closed with exact required members.
9. Prove hidden snapshot member and hidden release member/action routes return reporting route-family not-found codes instead of incident-family not-found codes.

Implementation tasks:
- Add snapshot and release route family exactly under `/api/v1/snapshots` and `/api/v1/releases`.
- Persist immutable snapshot descriptors separate from release records.
- Replace the provisional export summary with `cartulary.snapshot_export_model.v3` over current base-profile workbook surfaces and active reporting metadata.
- Add `recipient_partition_refs` to the release tuple, release resource, idempotency hash, output slot, redaction manifest, and approval binding.
- Persist release records, approval state, output hashes, invalidation fields, nullable render output fields for failed renders, `render_failed_reason_code`, and release-state transitions.
- Admit snapshot and release creation as reporting jobs with queued/running/terminal state, cancellation checkpoints before durable commit, and exactly one success ref.
- Render self-contained outputs without runtime remote assets.
- Implement incident-aware immutable redaction profile resolution backed by a closed built-in catalog.
- Maintain the closed snapshot/release family error and reason-code registries without carrying old catch-all reasons.
- Centralize current-profile reporting vocabularies for validators and template contracts while returning copy-safe slices to callers.
- Keep generated OpenAPI contract precision aligned with runtime and database closed vocabularies by using reusable `ReleaseOutputKind` and `ReleaseScope` schemas.
- Map non-visible snapshot and release resources to `snapshot_not_found` or `release_not_found`, while preserving visible insufficient-role publish/invalidate failures as `authorization_denied`.
- Add authoritative Phase 11 evidence rows for OpenAPI contract precision and endpoint-specific exact-shape/visibility/auth behavior.

Validation commands:
- `go test ./internal/modules/reporting`
- `go test ./internal/app ./internal/platform/httpapi ./internal/modules/incidents`
- `make generate`, `make generate-drift`, and `make migration-drift`
- `make phase-ledgers`, `make phase-schedules`, `make phase-ledger-drift`, `make phase-schedule-drift`, and `make phase-test-name-check`
- `make phase-slice PHASE=phase11`: latest post-contract-remediation pass at `.cartulary/test-results/20260524T160036Z-p2155281`
- `make service-backed-slice PHASE=phase11`: latest post-contract-remediation pass at `.cartulary/test-results/20260524T160051Z-p2157836`
- `make agent-finalize`: latest post-contract-remediation pass at `.cartulary/test-results/20260524T155911Z-p2149879`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- `make test-fast`: latest post-contract-remediation pass at `.cartulary/test-results/20260524T160600Z-p2167905`
- `make check`: historical Sprint 3 broad gate passed at `.cartulary/test-results/20260524T142209Z-p1934659`; it was not rerun after the later generated-contract precision and exact-shape/auth evidence hardening.
- `git diff --check`

Deliverables:
- Snapshot and release route family.
- Immutable snapshot and release durable state.
- Export model v3 over current workbook surfaces.
- Explicit recipient-partition release tuple and manifest binding.
- Durable render-failed release lifecycle.
- Reporting job payload/executor boundary for snapshot and release create.
- Self-contained output and redaction evidence.
- Direct non-aggregate evidence for every Snapshot and Reporting delta family.
- Reusable OpenAPI enum schemas for release output kind and release scope, referenced by both create and durable resource schemas.
- Endpoint-specific exact-shape, singleton pagination, route-scoped hidden-resource error, and visible insufficient-role action evidence.

Risks and assumptions:
- Live-recipient-specific workbook withholding is out of scope and non-conformant.
- Release approvals bind to the exact release tuple and rendered bytes.
- Core 05 applies only if timed or fixture-sensitive claim publication is made.
- Browser/network instrumentation remains out of scope for this API-rendered output boundary unless future renderer asset behavior changes.

Exit criteria:
- Passing Base claim evidence is identified.
- All Snapshot and Reporting delta ACs have direct evidence.
- The profile can be claimed independently without implying other extension profiles.
- `profile:snapshot_reporting.claimed` is true only after the direct Phase 11 slices, drift gates, finalizer maintenance, and selected broad gate passed. After later contract/evidence hardening, the direct slices, drift gates, finalizer without retained-run input, and `make test-fast` passed; `make check` remains the recommended pre-merge gate.

## Sprint 4. Reference Pack Extension Profile

Objective: Implement and prove the Reference Pack Extension Profile only when selected.

Status: complete for the selected full profile after remediation. Reference Pack import, reverify, and refresh are durable background jobs; disconnected deployments receive the required minimum built-in type-registry packs; omitted refresh replay uses the admission-time selector; and the authenticated landing/admin surface exposes deployment-admin progress/cancel controls.

Relevant IDs:
- `AC-033..AC-035`
- `AC-092..AC-096`
- `AC-234`
- `AC-270..AC-272`
- `AC-308..AC-310`
- `AC-326`
- `AC-369`

Files and areas:
- Runtime and persistence: `internal/modules/reference_data`, `internal/app/runtime.go`, `internal/platform/httpapi/extensions.go`, `internal/platform/jobs/runner.go`, `db/migrations/00027_phase11_reference_packs.sql`, and `db/migrations/00028_phase11_reference_pack_async_payloads.sql`.
- Owner-derived contracts and registries: `contracts/openapi/cartulary.openapi.yaml` and `contracts/errors/index.json`.
- Generated artifacts refreshed through canonical commands: `internal/gen/contracts/contracts_gen.go`, `internal/gen/sql/models.go`, `packages/protocol-ts/src/generated/contracts.ts`, `docs/testing/phase11_coverage_ledger.md`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, and `tools/go_test_duration_baselines.json`.
- Support expectation updates: Phase 2 reserved-extension tests use Incident Portability as the unclaimed root and accept Reference Pack as claimed; Reference Pack UI evidence is now carried by frontend-unit panel rows.

Test-first sequence:
1. Proved list/read route shapes, paging, sorting by `pack_key asc` then exact `pack_version asc`, singleton pagination rejection, not-found behavior, and exact `reference_pack_version` resources.
2. Proved `POST /api/v1/reference-packs/import` shared upload envelope, `activation_policy='staged_only'`, no auto-activation, queued-first file-hash idempotency, divergent replay rejection, durable job payloads, pre-verification cancellation, and terminal `reference_pack_imported` summaries.
3. Proved activation, disable, reverify, and refresh validation, route-scoped idempotency, legal transitions, authorization re-derivation, derived `active`, and exact resource refs.
4. Proved durable metadata: `pack_key`, `pack_kind`, `pack_version`, `manifest_sha256`, canonical `payload_sha256`, verification method/result, source identifier, signer key where applicable, `previous_active_version`, and prior-active retention.
5. Proved disconnected local-bundle behavior and no-network strategy with in-memory fixture bundles, integrity failure handling, prior-active retention, archive/path/content-screening failures, malformed bundles, missing payloads, signature failures, and refresh no-op behavior.

Implementation tasks:
- Added the route family exactly under `/api/v1/reference-packs`: list, read, import, activate, disable, reverify, and refresh.
- Persisted durable Reference Pack version metadata, activation pointers, attestation rows, and worker-owned job payloads including staged bundle paths and admission-time refresh selectors. Public `active` is derived from `reference_pack_activation_state`.
- Kept durable/public condition vocabulary limited to `staged`, `verified_available`, `disabled`, `failed`, and `missing`.
- Implemented local-only archive verification for zip, tar, tar.gz/gzip, and sniffed octet-stream archives with required `manifest.json`, safe relative paths, passive-content allowlist, canonical sorted payload digest summary, integrity checks, and no network fetch path.
- Maintained closed Reference Pack error and reason-code registries and generated OpenAPI/Go/TypeScript contracts from authored inputs.

Validation commands:
- Targeted direct tests:
  - `go test ./internal/modules/reference_data`
  - `go test ./internal/app ./internal/platform/httpapi`
  - `go test ./internal/modules/incidents -run 'TestPhase2_U_2_09|TestPhase2_U_2_10' -count=1`
  - `go test ./internal/modules/incidents -run 'TestPhase2_I_2_05|TestPhase2_I_2_06' -count=1`
  - `go test ./cmd/server -run TestPhase2_ExtensionDiscoveryAndReservedRoutes_E_2_SMOKE_01_ProcessSmoke -count=1`
  - `make lint-go`
  - `make lint-biome`
  - `make browser-e2e-support`
  - `make frontend-unit`
- Generation and harness maintenance:
  - `make generate`
  - `make phase-ledgers`
  - `make phase-schedules`
  - `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260524T195847Z-p2632792`
  - `make go-test-duration-baseline-coverage`
  - `make agent-finalize`
- Required profile and drift gates:
  - `make phase-slice PHASE=phase11`: passed, retained root `.cartulary/test-results/20260524T195317Z-p2620061`
  - `make service-backed-slice PHASE=phase11`: passed, retained root `.cartulary/test-results/20260524T195847Z-p2632792`
  - `make generate-drift`: passed, retained root `.cartulary/test-results/20260524T200408Z-p2642630`
  - `make migration-drift`: passed, retained root `.cartulary/test-results/20260524T200408Z-p2642691`
  - `make phase-ledger-drift`: passed, retained root `.cartulary/test-results/20260524T200408Z-p2642644`
  - `make phase-schedule-drift`: passed, latest retained root `.cartulary/test-results/20260524T200557Z-p2650662`
  - `make phase-test-name-check`: passed
  - `git diff --check`: passed
  - `make check`: passed, retained root `.cartulary/test-results/20260524T204455Z-p2852189`
  - Final `make agent-finalize`: passed with `generated=unchanged files=0 duration=skipped run_checks=skipped`, retained root `.cartulary/test-results/20260524T205045Z-p2902181`

Deliverables:
- Reference-pack route family and durable state.
- Verification, import, activation, disable, reverify, and refresh lifecycle evidence.
- Direct non-aggregate evidence for every Reference Pack delta family:
  - `U-11-REFERENCE-PACK-01` request validation, upload-envelope translation, action/selector normalization, closed registries, idempotency guards.
  - `U-11-REFERENCE-PACK-02` local verifier success and archive/integrity/path/content/signature/malformed/missing-payload failure mapping.
  - `U-11-REFERENCE-PACK-03` direct extracted-byte, compression-ratio, and archive-member limit failures.
  - `U-11-REFERENCE-PACK-04` and `U-11-REFERENCE-PACK-06` deployment-admin UI gating, job progress polling, and cancellation controls.
  - `U-11-REFERENCE-PACK-05` OpenAPI and error registry precision.
  - `I-11-REFERENCE-PACK-01` import/list/read, pagination, sorting, replay, job payloads, and terminal summaries.
  - `I-11-REFERENCE-PACK-02` activation, disable, reverify, refresh, exact refs, authorization re-derivation, and legal lifecycle transitions.
  - `I-11-REFERENCE-PACK-03` disconnected failure coverage, prior-active retention, local storage-root behavior, and no-network bundle processing.
  - `I-11-REFERENCE-PACK-04..07` queued-first admission, pre-commit cancellation, exact minimum disconnected seeding, omitted-selector refresh replay, upload-envelope no-durable-state behavior, and deployment-admin route gating.
- Profile claim status: `profile:reference_pack` is claimed in the default extension registry. Import and Snapshot/Reporting claims were preserved. Incident Portability and Enterprise Authentication remain unclaimed.

Risks and assumptions:
- Pack-dependent workbook overlays must not become Base Profile surfaces.
- Live internet fetch must not be required for disconnected verification or activation.
- Prior active version must remain retained on failed candidate activation.
- The repo-local test bundle manifest format is an implementation fixture and is not a behavior owner.
- Reference Pack operator UI is intentionally limited to deployment-admin sessions and the existing authenticated landing/admin surface; it does not create a separate product area.

Exit criteria:
- Passing Base prerequisite evidence is supported by the final `make check` root `.cartulary/test-results/20260524T204455Z-p2852189` plus direct Phase 11 wrapper evidence.
- All Reference Pack delta ACs have direct non-aggregate evidence in `tools/phase11_test_map.json` and the generated ledger.
- The profile is claimed without requiring Import, Snapshot and Reporting, Incident Portability, or Enterprise Authentication to be claimed.
- Skipped checks: browser/E2E operator workflow evidence remains a future hardening target; frontend-unit evidence now covers admin UI gating, progress polling, and cancellation controls.
- Unresolved blockers: none for the selected Reference Pack full profile.
- Remaining non-claim state: Incident Portability and Enterprise Authentication remain reserved and unclaimed.

## Sprint 5. Incident Portability Extension Profile

Objective: Implement and prove the Incident Portability Extension Profile only when selected.

Relevant IDs:
- `AC-164..AC-169`
- `AC-236`
- `AC-273..AC-276`
- `AC-327..AC-328`
- `AC-332`
- `AC-386`
- `AC-409`

Files and areas:
- Likely areas: portability/export/import services, object storage helpers, projection rebuild paths, temporary-work root handling, generated public contracts, and browser/admin surfaces if exposed.
- TODO: Confirm concrete module owner. Do not assume `internal/modules/recovery` owns portability; Phase 10 recovery is operational backup/restore, not incident portability.

Test-first sequence:
1. Prove `POST /api/v1/incident-bundles/export` request validation, selector canonicalization, fixed `history_mode='full'`, fixed `blob_mode='full'`, forbidden user modes, job summary, and idempotency.
2. Prove durable export descriptor read, pagination rejection, canonical arrays, manifest hash, and descriptor existence only after successful export.
3. Prove `POST /api/v1/incident-bundles/import` shared upload envelope, file hash idempotency, job summary, imported incident ref, no durable import resource, and rejection of clone/merge/identifier-remap/remote-fetch modes.
4. Prove deterministic bundle layout, authoritative-source-only export, no projections/search indexes/runtime state, full history/blob preservation, checksum verification, staged import under temporary-work root, and no partial visible incident on failure.
5. Prove no deployment-local login-capable admin/user/auth state is exported or imported, while historical actors remain inert descriptors or map only through explicit deployment-local administrative action.

Implementation tasks:
- Add the route family exactly under `/api/v1/incident-bundles`.
- Build deterministic export layout from authoritative incident source state and referenced blob bytes.
- Stage import under temporary-work root and verify manifest, checksums, signatures where applicable, and blob hashes before visibility.
- Rebuild projections after import before normal open.
- Reject unsupported required capabilities and invalid optional section handling as owned.
- Maintain incident-bundle family error and reason-code registries.

Validation commands:
- Profile-selected backend unit/store/process tests.
- Service-backed export/import with Postgres and object store.
- Browser or E2E test that imported incident opens normally if UI is exposed.
- `make phase-slice PHASE=phase11`
- `make service-backed-slice PHASE=phase11`
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `git diff --check`.

Deliverables:
- Incident-bundle route family and bundle format.
- Full-fidelity export/import evidence.
- Direct non-aggregate evidence for every Incident Portability delta family.

Risks and assumptions:
- Operational backup/restore is Phase 10 Base Profile behavior and must not be reinterpreted as incident portability.
- Import must not create deployment-local login-capable users, memberships, auth bindings, credentials, sessions, or deployment admins.
- Destructive fixture isolation and cleanup predicates must be explicit before service-backed import tests run.

Exit criteria:
- Passing Base claim evidence is identified.
- All Incident Portability delta ACs have direct evidence.
- The profile can be claimed independently.

## Sprint 6. Enterprise Authentication Extension Profile

Objective: Implement and prove the Enterprise Authentication Extension Profile only when selected.

Relevant IDs:
- `AC-036`
- `AC-235`
- `AC-288..AC-293`
- `AC-348..AC-352`

Files and areas:
- Likely areas: `internal/modules/auth`, deployment-local auth persistence, session issuance/revocation, provider configuration, generated public contracts, and browser login flows.
- TODO: Confirm IdP fixture strategy and provider metadata storage boundary before implementation.

Test-first sequence:
1. Prove `GET /api/v1/auth/providers` lists only enabled interactive providers, sorted by `display_name asc` then `provider_key asc`, exact item shape, no secrets/metadata/claim maps, and pagination rejection.
2. Prove `POST /api/v1/auth/providers/{provider_key}/begin` accepts only optional `return_to`, normalizes omitted/null to `/`, rejects `client_txn_id`, validates same-origin relative return paths, and creates no public durable transaction.
3. Prove OIDC callback using authorization-code flow with PKCE `S256` and `nonce`, same session resource family, `303 See Other`, replay/expiry/state/nonce/code/provider failures, and startup fallback reuse.
4. Prove SAML SP-initiated ACS, same session resource family, `303 See Other`, replay/expiry/RelayState/issuer/audience/signature/assertion failures, and no IdP-initiated SAML.
5. Prove provider subject mapping to existing local users only, no JIT local user, no auth identity auto-create, no incident membership auto-create, no group-claim-to-role mapping.
6. Prove deployment-admin binding create, rotate, and retire routes, exact idempotency, safe user resource summaries, active uniqueness, audit lineage, session revocation on rotate/retire, and forbidden secret/assertion retention.

Implementation tasks:
- Add Enterprise Authentication protocol routes exactly under `/api/v1/auth/providers`, `/api/v1/auth/oidc`, and `/api/v1/auth/saml`.
- Add binding routes exactly under `/api/v1/users/{user_id}/auth-bindings`.
- Persist server-side single-use auth transactions and active provider bindings as deployment-local auth state.
- Reuse the same opaque server-managed session family as base auth.
- Keep provider secrets, raw assertions, provider tokens, and raw metadata server-side and out of public resources.
- Maintain enterprise-auth and binding-management error and reason-code registries.

Validation commands:
- Profile-selected backend unit/store/integration/process tests.
- Browser or E2E sign-in flows with deterministic fake OIDC/SAML providers if supported by harness.
- `make phase-slice PHASE=phase11`
- `make service-backed-slice PHASE=phase11`
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `git diff --check`.

Deliverables:
- Enterprise-auth provider protocol and binding-management route families.
- Provider-backed session convergence evidence.
- Direct non-aggregate evidence for every Enterprise Authentication delta family.

Risks and assumptions:
- The profile defines no WebAuthn/passkeys, SCIM, JIT provisioning, self-service linking, or enterprise-only user-creation path.
- Protocol route failures must create no session.
- Rotate and retire must revoke all active sessions for the target user.

Exit criteria:
- Passing Base claim evidence is identified.
- All Enterprise Authentication delta ACs have direct evidence.
- The profile can be claimed independently.

## Sprint 7. Profile Claim Gates, Generated Artifacts, Finalizers, And Handoff

Objective: Maintain the Phase 11 finalization lane, close profile claims as each profile is implemented, refresh downstream artifacts through generators, and leave explicit handoff state for profiles that remain open future work.

Sprint 7 status:
- Open.
- Import, Snapshot/Reporting, and Reference Pack remediation finalization evidence is recorded here because those selected API route families are closed.
- Incident Portability and Enterprise Authentication are still open implementation sprints and are not closed out by this status record.

Relevant IDs:
- All selected profile AC deltas.
- Aggregate claim gate for each selected profile: `AC-232`, `AC-233`, `AC-234`, `AC-235`, or `AC-236`.
- Core 05 `PC-*` only if claim-bearing timed or fixture-sensitive publication is made.

Files and areas:
- `tools/phase_registry.json`
- `tools/phase11_test_map.json`.
- Generated ledgers and schedules derived from the active Phase 11 manifest.
- Generated contract outputs under `internal/gen/**` and `packages/protocol-ts/src/generated/**`.
- Duration-maintenance inputs refreshed from retained successful runs.
- `PHASE11_IMPLEMENTATION_PLAN.md` for final status and retained roots.

Test-first sequence:
1. Confirm Base claim prerequisite evidence.
2. Confirm every selected profile has direct non-aggregate evidence for its substantive behavior families.
3. Confirm each profile without completed implementation remains unclaimed and reserved-family behavior is unchanged.
4. Run drift gates for generated contracts, migrations, ledgers, schedules, and phase names.
5. Run `make agent-finalize` first for end-of-run maintenance, with `RESULTS_DIR=<successful retained run root>` only when one exists and is valid.
6. Run `make check` or narrower profile-final gate as adopted by repository convention.

Implementation tasks:
- Update profile claim status only after direct evidence passes.
- Record retained roots and final gates without fabricating run IDs.
- Keep generated artifacts downstream and refreshed through canonical targets.
- Record blockers as corpus defects if owner sections conflict.
- Document any skipped retained-run maintenance when `RESULTS_DIR` is unset.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase11`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `make generate-drift`
- `make migration-drift`
- `make agent-finalize`
- `make check`
- `git diff --check`

Deliverables:
- Profile-selected final evidence for Import, Snapshot/Reporting, Reference Pack, and common jobs.
- Refreshed generated artifacts through canonical commands.
- Updated plan/handoff status for selected, claimed, and still-open profiles.
- Explicit blocker list: no open blocker for the selected Import, Snapshot/Reporting, or Reference Pack API route families.
- Retained final gate roots recorded below.
- Full Phase 11 final handoff remains open until the remaining profile sprints are implemented or explicitly rescoped.

Import remediation validation record:
- `make explain-phase PHASE=phase11`: passed and reported active Import/common-job rows.
- `make phase-slice PHASE=phase11`: passed; retained root `.cartulary/test-results/20260523T153803Z-p3120091`.
- `make service-backed-slice PHASE=phase11`: passed; retained root `.cartulary/test-results/20260523T154329Z-p3131974`.
- `make generate-drift`, `make migration-drift`, `make phase-map-check`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-test-name-check`, `make json-shape-check`, `make lint-scripts`, and `make harness-contract-tests`: passed.
- `make check`: passed; retained root `.cartulary/test-results/20260523T161735Z-p3268642`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260523T161735Z-p3268642`: passed; retained root `.cartulary/test-results/20260523T162808Z-p3327944`; generated outputs unchanged after finalizer refresh.
- An earlier retained-run finalizer attempt exposed a stale warm `check-service-backed` budget. The budget was raised to `120000ms`, then finalizer maintenance passed against the successful `make check` root.

Snapshot and Reporting remediation implementation record:
- Owner specs now bind `cartulary.snapshot_export_model.v3`, `cartulary.source_boundary.v1:<sha256>` source-boundary tokens, explicit workbook export-family collection, closed `output_kind` and `release_scope` durable resource vocabularies, `recipient_partition_refs`, durable `render_failed` release resources, release-create job behavior, approval tuple provenance, deterministic redaction rule precedence, reserved `hash` handling, post-redaction validation, disclosure-partition eligibility, route-scoped hidden-resource errors, admin-only publish/invalidate behavior, and the opaque-binary boundary.
- OpenAPI now declares `/api/v1/snapshots`, `/api/v1/snapshots/{snapshot_id}`, `/api/v1/releases`, `/api/v1/releases/{release_id}`, release `approve`, `publish`, and `invalidate` action routes, reusable `ReleaseOutputKind` and `ReleaseScope` enum schemas, release recipient partitions, render failure reason codes, and nullable output fields for `render_failed`.
- Implementation added `internal/modules/reporting` with immutable async snapshot creation, export model v3 collection over workbook surfaces, release rendering through the versioned default template contract, deterministic redaction manifests, persisted rendered artifacts, route-scoped idempotency, recipient-partition-aware external releases, durable render-failed rows, distinct external reviewer/admin approvals, publish/invalidate actions, route-local snapshot/release visibility errors, centralized current-profile vocabulary helpers, SQLC-backed reporting-table queries, and runtime claim registration for `profile:snapshot_reporting`.
- Persistence added `reporting_snapshots`, `reporting_releases`, `reporting_release_approvals`, reporting job payloads, recipient partition columns, render-failed nullable output constraints, and authored SQLC query inputs in `db/queries/reporting_phase11.sql`.
- SQLC ownership cleanup moved superseded-release invalidation back to the generated `InvalidateSupersededReportingReleases` query, including canonical `recipient_partition_refs[]`, and removed the duplicate inline invalidation SQL from the reporting store.
- Migration `00026_phase11_reporting_remediation.sql` refuses to run against non-empty reporting tables instead of deleting incompatible pre-claim rows, adds recipient partition and reporting job payload storage, allows output fields to be nullable only for durable `render_failed`, and avoids legacy renderer support.
- Direct evidence added Phase 11 reporting unit rows `U-11-REPORTING-01..06` and integration rows `I-11-REPORTING-01..04`. The added post-remediation rows assert OpenAPI enum/resource precision, exact snapshot/release response key sets, singleton pagination rejection, route-scoped hidden-resource errors, visible insufficient-role action errors, and action response parity with `GET /api/v1/releases/{release_id}`.
- Harness diagnostic cleanup added `explain-run` support for retained public tool-run roots that contain `<target>/tool-run-summary.json`, including bounded `agent-finalize/finalize-summary.json` summaries and child log access without fabricating `run-summary.json`.
- Focused validation: `go test ./internal/modules/reporting`, `make check-harness-smoke`, `make generate`, `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, and `git diff --check` passed during this cleanup.
- Retained roots from this cleanup: `make generate` `.cartulary/test-results/20260524T140301Z-p1887464`; `make phase-slice PHASE=phase11` `.cartulary/test-results/20260524T141115Z-p1914505`; `make service-backed-slice PHASE=phase11` `.cartulary/test-results/20260524T141644Z-p1925389`; `make check` `.cartulary/test-results/20260524T142209Z-p1934659`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T142209Z-p1934659` `.cartulary/test-results/20260524T142930Z-p1990229`.
- Finalizer notes: retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T142209Z-p1934659` passed with `generated=updated files=6 duration=refreshed run_checks=pass`; updated files were the browser, Go, harness-smoke, service-backed duration baselines plus `tools/execution_topology_render_index.json` and `tools/scheduler_manifest.json`.
- Broad gate note: an initial `make check` attempt at `.cartulary/test-results/20260524T000651Z-p491510` exposed two unused reporting integration helpers after async job polling replaced direct table helpers. They were removed, `go test ./internal/modules/reporting` and `make lint-go-staticcheck` passed, and the broad gate was rerun successfully.
- Historical Sprint 3 broad gate before generated-contract precision remediation: `make check` passed at `.cartulary/test-results/20260524T142209Z-p1934659` with 144/144 work units and 695 tests. Retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T142209Z-p1934659` passed at `.cartulary/test-results/20260524T142930Z-p1990229`, and both retained roots are explainable through `make explain-run`.

Snapshot and Reporting generated-contract precision remediation record:
- The confirmed gap was OpenAPI resource precision only: runtime code and database constraints already enforced closed `release_scope` and `output_kind` vocabularies, but `ReleaseResource` exposed those fields as plain strings.
- Remediation added reusable OpenAPI enum schemas for `ReleaseOutputKind` (`html`, `markdown`, `slidev`, `mermaid`, `reenactment`) and `ReleaseScope` (`internal_draft`, `internal_review`, `external_release`) and referenced them from both `ReleaseCreateRequest` and `ReleaseResource`.
- Route behavior was aligned with clarified owner text: hidden snapshot reads and release-create attempts against hidden snapshots return `snapshot_not_found`; hidden release reads and release actions return `release_not_found`; visible publish/invalidate callers without incident `admin` role keep `authorization_denied`; approve/publish/invalidate reject singleton pagination query members.
- Reporting template contracts and validators now share the same runtime vocabulary source with copy-safe slices rather than duplicated literal lists.
- Generated artifacts refreshed by canonical commands: OpenAPI-derived Go/TypeScript contracts, generated Phase 11 ledger, generated schedules, and Go duration baselines.
- Latest post-remediation validation: `make generate`, `make phase-ledgers`, `make phase-schedules`, isolated `go test ./internal/modules/reporting -run TestPhase11_U_11_REPORTING_06_OpenAPIReleaseEnumsAndExactResources`, `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260524T155233Z-p2135664`, `make go-test-duration-baseline-coverage`, `make agent-finalize`, `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, `make test-fast`, and `git diff --check` passed.
- Latest retained roots: `make agent-finalize` `.cartulary/test-results/20260524T155911Z-p2149879`; `make generate-drift` `.cartulary/test-results/20260524T160011Z-p2152395`; `make migration-drift` `.cartulary/test-results/20260524T160016Z-p2153189`; `make phase-slice PHASE=phase11` `.cartulary/test-results/20260524T160036Z-p2155281`; `make service-backed-slice PHASE=phase11` `.cartulary/test-results/20260524T160051Z-p2157836`; `make test-fast` `.cartulary/test-results/20260524T160600Z-p2167905`.
- Latest finalizer note: `make agent-finalize` passed with `generated=unchanged files=0 duration=skipped run_checks=skipped results_dir=-`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. An earlier finalizer attempt failed because the new integration shard lacked a Go duration baseline; a successful Phase 11 service-backed retained root was used to refresh the missing observed baseline without pruning partial evidence, and baseline coverage then passed.
- `make check` was not rerun after this generated-contract precision remediation. The historical Sprint 3 `make check` root above remains evidence for the earlier full Sprint 3 implementation state; `make check` remains the recommended pre-merge broad gate for the latest remediation.

Reference Pack remediation implementation record:
- Reconnaissance confirmed `profile:reference_pack` owner sections and AC mappings, reusable shared upload-envelope parsing and durable common jobs, the `internal/modules/reference_data` module boundary, authored contract/migration locations, the need for operator progress/cancel controls, and a local archive fixture strategy with no network dependency.
- Implementation added `internal/modules/reference_data` for request normalization, route handling, durable store access, local archive verification, queued-first worker execution, disconnected seed packs, and direct tests. Runtime assembly registers the module and keeps `profile:reference_pack` claimed after remediation.
- Persistence added `reference_packs`, `reference_pack_activation_state`, `reference_pack_attestations`, and `reference_pack_job_payloads` in `db/migrations/00027_phase11_reference_packs.sql`; `db/migrations/00028_phase11_reference_pack_async_payloads.sql` adds staged bundle payload storage for worker admission.
- OpenAPI declares list/read/import/activate/disable/reverify/refresh routes, exact `reference_pack_version` resources, closed action request/resource schemas, and job summary refs. `contracts/errors/index.json` declares closed Reference Pack errors and reason-code registries.
- Direct evidence added Phase 11 Reference Pack unit rows `U-11-REFERENCE-PACK-01..06`, integration rows `I-11-REFERENCE-PACK-01..07`, and browser row `E-11-01`.
- Generated artifacts refreshed by canonical commands: OpenAPI-derived Go/TypeScript contracts, SQL-derived Go models, generated Phase 11 ledger, generated schedules, and Go duration baselines.
- Focused validation passed: `go test ./internal/modules/reference_data ./internal/platform/httpapi ./internal/app ./internal/modules/jobapi -count=1`, `make frontend-typecheck`, `make frontend-unit`, `make lint-biome`, `make lint-go`, and `pnpm -C apps/web exec playwright test e2e/phase11.reference-pack.spec.ts --workers=1`.
- Required final validation passed before the final staging-path hardening change: `make generate`, `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-test-name-check`, `git diff --check`, `make check`, and final `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T230021Z-p3178170`. After the staging-path hardening change and browser row addition, targeted Reference Pack validation and fresh Phase 11 public wrappers passed.
- Latest retained roots: `make phase-slice PHASE=phase11` `.cartulary/test-results/20260524T232929Z-p3346924`; `make service-backed-slice PHASE=phase11` `.cartulary/test-results/20260524T232929Z-p3346928`; `make browser-e2e-webserver-backed` `.cartulary/test-results/20260524T233702Z-p3372186`; `make browser-e2e-duration-baselines RESULTS_DIR=.cartulary/test-results/20260524T233702Z-p3372186` `.cartulary/test-results/20260524T233821Z-p3380034`; `make browser-e2e-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260524T233702Z-p3372186` `.cartulary/test-results/20260524T233825Z-p3380259`; `make generate` `.cartulary/test-results/20260524T224015Z-p3104816`; `make generate-drift` `.cartulary/test-results/20260524T230921Z-p3236064`; `make migration-drift` `.cartulary/test-results/20260524T230923Z-p3236729`; `make phase-ledger-drift` `.cartulary/test-results/20260524T233915Z-p3381853`; `make phase-schedule-drift` `.cartulary/test-results/20260524T233915Z-p3381859`; `make frontend-unit` `.cartulary/test-results/20260524T232326Z-p3325069`; prior successful `make check` `.cartulary/test-results/20260524T230021Z-p3178170`; retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T230021Z-p3178170` `.cartulary/test-results/20260524T230747Z-p3232178`; latest no-`RESULTS_DIR` `make agent-finalize` `.cartulary/test-results/20260524T233919Z-p3382354`.
- Finalizer notes: an initial `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T224115Z-p3114969` failed retained-run preflight because the retained root was a Phase 11 service-backed slice, not a warm `check` run. A no-`RESULTS_DIR` finalizer then passed with retained-run checks skipped. Missing duration baselines for the new Reference Pack service-backed rows were refreshed from `.cartulary/test-results/20260524T224115Z-p3114969`, schedules were regenerated, and retained-run `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T230021Z-p3178170` passed with `generated=updated files=3 duration=refreshed run_checks=pass`. After the browser row addition and browser duration baseline refresh, `make agent-finalize` passed again with `generated=unchanged files=0 duration=skipped run_checks=skipped results_dir=-`.
- Broad gate notes: the first `make check` attempt exposed Biome formatting/import organization and staticcheck findings in the new Reference Pack code. After cleanup, `make check` passed with 151/151 work units and 710 tests at `.cartulary/test-results/20260524T230021Z-p3178170`. Later reruns after the staging-path hardening change failed outside the Reference Pack surface: `.cartulary/test-results/20260524T231031Z-p3240870` failed the existing Phase 3 browser support assertion expecting `Zulu anchored` but receiving `Alpha summary`; `.cartulary/test-results/20260524T231556Z-p3282510` failed the existing Phase 9 frontend sentinel assertion expecting three calls but receiving one; and final `.cartulary/test-results/20260524T234000Z-p3384502` failed the existing Phase 3 support row-version assertion at `apps/web/e2e/phase3.support.spec.ts:166`, expecting row version `2` but receiving `1`. Standalone `make frontend-unit`, targeted Reference Pack Go tests, full `make browser-e2e-webserver-backed`, drift checks, and Phase 11 wrappers passed on the final tree.
- Skipped checks: no Reference Pack-specific checks were skipped. The final full `make check` attempt was run and failed outside the Reference Pack surface as noted above.
- Unresolved blockers: none for the selected Reference Pack full profile. Remaining non-claim state: Incident Portability and Enterprise Authentication remain reserved and unclaimed.

Risks and assumptions:
- Broad gates may be heavy and should run only when product implementation occurs or repository convention requires them.
- Successful retained roots must not be fabricated.
- Prior plans and retained artifacts are handoff inputs, not behavior authorities.

Exit criteria:
- Each claimed profile has a passing Base prerequisite plus complete profile-specific direct evidence.
- Open future profile sprints remain explicitly open until implemented or explicitly rescoped.
- Profiles without completed implementation remain unclaimed.
- No generated or tool-managed file was hand-edited.
- Final handoff identifies commands run, retained roots, unsupported commands, and unresolved blockers.

## Generated Boundaries

- Do not hand-edit `internal/gen/**`.
- Do not hand-edit `packages/protocol-ts/src/generated/**`.
- Do not hand-edit generated coverage ledgers or generated schedules.
- Do not hand-edit generated Make includes.
- Do not hand-edit `go.sum` or `pnpm-lock.yaml`.
- If contract or SQL generation is required in future implementation work, edit owner inputs and use canonical generator targets.
- Generated ledgers and schedules are downstream artifacts and must not be treated as behavior authorities.
- `contracts/*` is downstream of the normative core and upstream of generated code. Hand-edit contracts only as owner-driven contract updates, not as independent behavior authority.
- Keep codegen drift and migration drift separate.

## Validation Command Policy

This file began as a planning artifact and now also records implemented remediation status. Owner specs, contracts, implementation code, tests, and generated artifacts remain the source of executable truth.

Planning-task validation commands:
- `git status --short`
- `make phase-map-check`
- `make explain-phase PHASE=phase11`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `git diff --check`

Expected remediation-task result:
- `make explain-phase PHASE=phase11` reports an active manifest with Import, Snapshot/Reporting, Reference Pack, and common-job rows.
- `make phase-slice PHASE=phase11` and `make service-backed-slice PHASE=phase11` execute the selected rows.
- `make check` passed at retained root `.cartulary/test-results/20260523T161735Z-p3268642`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260523T161735Z-p3268642` passed at retained root `.cartulary/test-results/20260523T162808Z-p3327944`.
- Snapshot/Reporting cleanup `make check` passed at retained root `.cartulary/test-results/20260524T142209Z-p1934659`.
- Snapshot/Reporting cleanup `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260524T142209Z-p1934659` passed at retained root `.cartulary/test-results/20260524T142930Z-p1990229`.
- Reference Pack closeout `make check` passed at retained root `.cartulary/test-results/20260524T204455Z-p2852189`.
- Reference Pack closeout final `make agent-finalize` passed at retained root `.cartulary/test-results/20260524T205045Z-p2902181`.
- Do not fabricate successful retained run roots.
