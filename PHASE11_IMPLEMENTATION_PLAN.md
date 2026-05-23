# Phase 11 Implementation Plan

## Summary

This file is the execution roadmap, progress marker, and handoff aid for Cartulary Phase 11: extension profile testing hooks.

Phase 11 is not Base Profile conformance. Phases 0 through 10 are the base-profile implementation sequence. Phase 11 keeps future extension-profile work aligned with the current extension route families, profile claim boundaries, acceptance criteria, and harness mechanics.

Core 00 through Core 04 own implementation-conformance behavior. Core 05 owns claim-bearing timed or fixture-sensitive publication only and does not define ordinary runtime implementation behavior. This file is non-normative: it records execution status, remediation decisions, and handoff notes. Runtime behavior is owned by the normative core, authored contracts and migrations, implementation code, generated artifacts, and direct tests.

Current remediation state:
- Import is the only selected and claimed Phase 11 extension profile.
- Phase 11 now has an active shared manifest at `tools/phase11_test_map.json` with Import and common-job rows only.
- Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication remain reserved and unclaimed.
- The remaining profile sprints are open future implementation work. They are not closed, failed, or descoped by the Import remediation.
- Helper-only Sprint 1 evidence is not used by itself to claim route parity; Import claim evidence is carried by the completed route-family implementation and direct Phase 11 rows.

Implemented remediation record:
- Owner/spec alignment was completed for the shared upload envelope and Import request reasons: valid JSON metadata that is not an object maps to route-family `request_not_object`, while `invalid_value` is owner-backed for scalar, type, and format validation.
- Common substrate is in place for upload-envelope parsing, durable common jobs, route-scoped cancel idempotency, terminal job summaries, and request-time authorization re-derivation.
- The selected Import route family is implemented for session create/read, unit list/read, preview, mapping approval, select, skip, and apply.
- Import discovery covers bounded CSV and XLSX used-range sources, source byte hashing, deterministic mapping fingerprints, provenance through apply, duplicate apply blocking, and durable Import state that is separate from job status.
- OpenAPI, generated Go/TypeScript contracts, SQL-derived generated code, phase manifest, generated phase ledger, generated schedules, and duration-maintenance inputs were refreshed through canonical targets.
- Phase 2, process, and browser reserved-extension expectations were adjusted to use an unclaimed profile root instead of Import now that Import is claimed.
- Retained-run finalizer maintenance was run after the successful `make check` root and the warm `check-service-backed` retained-run budget was raised to `120000ms` to match observed successful timing.

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
- Phase 11 uses one shared manifest, `tools/phase11_test_map.json`, with profile-selected rows only. Current active rows cover Import and common jobs.
- Existing public wrappers, `make phase-slice PHASE=phase11` and `make service-backed-slice PHASE=phase11`, are sufficient for the current active Import claim.

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
| Snapshot and Reporting | Passing Base claim plus `profile:snapshot_reporting` requirements | `/api/v1/snapshots`, `/api/v1/releases` | Immutable snapshot descriptor, release record, approval records, output hashes, redaction metadata | Snapshot create and release create use common job resource; approve/publish/invalidate are synchronous | Export output, self-contained assets, recipient redaction without live workspace withholding | `AC-030..AC-032`, `AC-056..AC-062`, `AC-071`, `AC-091`, `AC-104..AC-106`, `AC-113..AC-115`, `AC-233`, `AC-266..AC-269`, `AC-305..AC-307`, `AC-333` | `not_started` |
| Reference Pack | Passing Base claim plus `profile:reference_pack` requirements | `/api/v1/reference-packs` | `reference_pack_version`, activation pointer, verification metadata, prior active version | Import, reverify, and refresh are jobs; activate and disable may be sync or jobs | Hostile local pack bundles, disconnected operation, integrity and content screening | `AC-033..AC-035`, `AC-092..AC-096`, `AC-234`, `AC-270..AC-272`, `AC-308..AC-310`, `AC-326`, `AC-369` | `not_started` |
| Incident Portability | Passing Base claim plus `profile:incident_portability` requirements | `/api/v1/incident-bundles` | Durable export descriptor, deterministic bundle, staged import state, imported incident | Export and import use common job resource; terminal summaries emit one bundle or incident ref | Whole-incident bundle integrity, temporary-work staging, encrypted roots, no deployment-local auth state | `AC-164..AC-169`, `AC-236`, `AC-273..AC-276`, `AC-327..AC-328`, `AC-332`, `AC-386`, `AC-409` | `not_started` |
| Enterprise Authentication | Passing Base claim plus `profile:enterprise_authentication` requirements | `/api/v1/auth/providers`, `/api/v1/auth/oidc`, `/api/v1/auth/saml`, `/api/v1/users/{user_id}/auth-bindings` | Enterprise provider config, server-side auth transaction, active auth binding, binding audit lineage | Protocol routes are intentionally non-idempotent and not job-backed; binding routes are synchronous idempotent mutations | Provider metadata, raw assertions, tokens, provider subjects, deployment-admin binding control | `AC-036`, `AC-235`, `AC-288..AC-293`, `AC-348..AC-352` | `not_started` |

## Evidence Layer Matrix

Phase 11 now has authoritative active rows in `tools/phase11_test_map.json`. Current adopted rows are limited to Import and common jobs: `U-11-JOBS-01`, `U-11-JOBS-02`, `I-11-IMPORT-01`, `I-11-IMPORT-02`, and `I-11-IMPORT-03`. Other profile evidence categories remain planning-only until future profile rows are explicitly adopted.

| Profile | Backend unit evidence | Backend store evidence | Backend integration/process evidence | Browser/E2E evidence | Generated/drift evidence | Claim blockers |
| --- | --- | --- | --- | --- | --- | --- |
| Import | Upload-envelope parser, request normalization, mapping fingerprint, error registries, parser isolation guards | Import sessions/units, provenance, source hash, mapping persistence, duplicate-apply and overlap state | Real Postgres import lifecycle, CSV/XLSX bounded discovery, apply jobs, idempotent replay | No browser Import workflow is claimed in the current API-only evidence set | OpenAPI, generated Go/TS contracts, SQL-derived models, phase ledger, phase schedule, and duration baselines refreshed through generators | No open blocker for the current selected Import API route family; future UI evidence is separate and unclaimed |
| Snapshot and Reporting | Snapshot/release request validation, release-state guards, approval tuple, redaction policy selection, error registries | Immutable snapshots, release records, approval records, output hashes, invalidation persistence | Render jobs, self-contained output production, snapshot boundary high-watermark behavior | User-visible progress, release approve/publish/invalidate, recipient-specific output inspection | Contract/OpenAPI, rendered-output test fixtures, ledgers/schedules through generators | Base claim freshness; release template/redaction fixture policy TODO; no route implementation |
| Reference Pack | Upload envelope, activation-policy validation, state conflict registry, derived `active` rules | Pack metadata, verification result, active pointer, prior active version, durable conditions | Bundle staging, verification, activation, disable, reverify, refresh, disconnected no-network fixtures | Operator pack import/activate/disable/reverify/refresh flows if product UI exists | Contract/OpenAPI, reference-pack fixture manifests, ledgers/schedules through generators | Base claim freshness; disconnected bundle fixture shape TODO; no route implementation |
| Incident Portability | Export/import request validation, bundle selector canonicalization, error registries, forbidden-mode rejection | Export descriptor, manifest checksum records, staged import bookkeeping, imported actor descriptors | Whole-incident export/import with Postgres/object store, projection rebuild, missing-file/checksum failures | Export/import progress and imported incident open if product UI exists | Contract/OpenAPI, bundle fixture generation, ledgers/schedules through generators | Base claim freshness; destructive import fixture isolation TODO; no route implementation |
| Enterprise Authentication | Provider discovery, begin validation, callback/ACS rejection, binding request validation, error registries | Provider config, auth transaction, binding lifecycle, active uniqueness, audit lineage | OIDC/SAML protocol simulation, session issuance, session revocation on rotate/retire | Provider sign-in redirect/callback/ACS and post-login landing behavior if harness can simulate IdP | Contract/OpenAPI and generated protocol drift if route schemas change; ledgers/schedules through generators | Base claim freshness; IdP fixture strategy TODO; no route implementation |

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers | Follow-up notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Phase 11 ownership model, profile-selection policy, and harness setup | `make phase-map-check`, `make explain-phase PHASE=phase11`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-test-name-check` | None. | Sprint 0 is historical. Phase 11 is now active because Import/common-job rows were adopted. |
| [x] | 1. Common extension substrate and upload-envelope foundation | `go test ./internal/platform/jobs ./internal/modules/jobapi ./internal/modules/imports ./internal/modules/imports/tabularingest ./internal/platform/httpapi ./internal/app`; `make generate`; `make phase-slice PHASE=phase11` | None for common substrate. | Helper evidence remains substrate only; Import route-family evidence in Sprint 2 carries the claim. |
| [x] | 2. Import Extension Profile | `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, `make check` | None for the selected API route family. | Import is the only selected and claimed profile; no browser Import workflow is claimed. |
| [ ] | 3. Snapshot and Reporting Extension Profile | profile-selected targets | Pending implementation. | Covers snapshots, releases, approval tuple, state machine, rendering, invalidation, redaction boundary. |
| [ ] | 4. Reference Pack Extension Profile | profile-selected targets | Pending implementation. | Covers pack import, verification, activation, disable, reverify, refresh, disconnected bundle constraints. |
| [ ] | 5. Incident Portability Extension Profile | profile-selected targets | Pending implementation. | Covers export/import bundle layout, checksums, authoritative source state, blob/history preservation, no deployment-local admin import. |
| [ ] | 6. Enterprise Authentication Extension Profile | profile-selected targets | Pending implementation. | Covers providers, begin/callback/ACS, same session family, binding management, no auto-provisioning. |
| [ ] | 7. Profile claim gates, generated artifacts, finalizers, and handoff | `make check`, `make agent-finalize`, drift gates, profile wrappers as applicable | Open until remaining profile sprints are implemented or explicitly rescoped. | Import finalizer evidence is recorded below; full Phase 11 closeout remains open. |

## Sprint 0. Phase 11 Ownership Model, Profile-Selection Policy, And Harness Setup

Objective: Establish Phase 11 traceability and profile-selection policy without making product claims.

Relevant IDs:
- Extension profile selectors: `profile:import`, `profile:snapshot_reporting`, `profile:reference_pack`, `profile:incident_portability`, `profile:enterprise_authentication`.
- Aggregate claim gates: `AC-232`, `AC-233`, `AC-234`, `AC-235`, `AC-236`.
- Active authoritative Phase 11 rows now exist for Import and common jobs only: `U-11-JOBS-01`, `U-11-JOBS-02`, `I-11-IMPORT-01`, `I-11-IMPORT-02`, and `I-11-IMPORT-03`.

Files and areas:
- `tools/phase_registry.json`
- `tools/phase11_test_map.json` is the active shared Phase 11 manifest and contains only Import/common-job rows.
- `docs/testing/phase11_coverage_ledger.md` is generated from the active shared manifest.
- `scripts/check-phase-maps.sh`, `scripts/render-phase-ledgers.mjs`, `scripts/render-execution-topology-artifacts.mjs`, and `scripts/check-phase-test-names.mjs`.
- `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, and generated schedule artifacts define the public wrapper and service-backed selection behavior.

Test-first sequence:
1. Sprint 0 first registered Phase 11 as planned and non-claiming.
2. Remediation adopted only the direct Import/common-job rows required for the selected profile.
3. Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication remain `not_started`.
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
- Historical Sprint 0 profile statuses were all `not_started`. After remediation, Import is selected and implemented; Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication remain `not_started`.
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
- Updated plan notes record Import as selected/claimed and all other profiles as unclaimed.

Risks and assumptions:
- The previous `make explain-phase PHASE=phase11` unknown-phase failure was expected while Phase 11 was unregistered. After Sprint 0 registration, this command is expected to pass and report `status: planned`.
- Current `make explain-phase PHASE=phase11` reports active Import/common-job coverage, not planned coverage.
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
- Current selected profile set in the default registry after remediation: Import only. Import is claimed only after its Sprint 2 route-family evidence; all other Phase 11 profiles remain unclaimed.
- Phase 11 manifest policy after remediation: one shared active `tools/phase11_test_map.json` with Import/common-job rows only. Generated ledgers and schedules are produced from that manifest.
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
- Explicit non-claim status for not-yet-implemented profiles: Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication remain unclaimed in the default extension registry.

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
- Default extension discovery claims Import and keeps Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication unclaimed.
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
- Default extension discovery reports Import as claimed and leaves all other extension profiles unclaimed.
- The active Phase 11 manifest contains only Import and common-job rows.
- Direct tests cover non-object metadata early failure, CSV mapping/select/apply, XLSX used-range discovery, and incident/deployment common-job authorization re-derivation.

Sprint 2 validation record:
- `go test ./internal/platform/httpapi ./internal/platform/jobs ./internal/modules/jobapi ./internal/modules/imports ./internal/modules/imports/tabularingest ./internal/app`: passed.
- `make phase-slice PHASE=phase11`: passed; retained root `.cartulary/test-results/20260523T153803Z-p3120091`.
- `make service-backed-slice PHASE=phase11`: passed; retained root `.cartulary/test-results/20260523T154329Z-p3131974`.
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, and `make phase-test-name-check`: passed.

## Sprint 3. Snapshot And Reporting Extension Profile

Objective: Implement and prove the Snapshot and Reporting Extension Profile only when selected.

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
- Likely areas: `internal/modules/reporting`, export-model derivation, job helpers, object storage/output roots, generated public contracts, and browser report/release surfaces if exposed.
- TODO: Confirm concrete source owners and template/redaction fixture locations before implementation.

Test-first sequence:
1. Prove `POST /api/v1/snapshots` request validation, omitted high-watermark resolution at admission, replay using original boundary, job summary, and exact snapshot read shape.
2. Prove `POST /api/v1/releases` exact version selectors, `release_scope` defaulting to `internal_draft`, closed vocabulary, no `latest/current`, render job summary, and exact release resource shape.
3. Prove approve, publish, and invalidate action routes, approval tuple binding, replay before fresh state checks, legal state transitions, and exact `release_state` vocabulary.
4. Prove self-contained disconnected output and deterministic ordering/hashes.
5. Prove recipient-specific redaction profiles operate at snapshot/render/release time and do not change live workbook query results, field visibility, row visibility, or evidence visibility.

Implementation tasks:
- Add snapshot and release route family exactly under `/api/v1/snapshots` and `/api/v1/releases`.
- Persist immutable snapshot descriptors separate from release records.
- Persist release records, approval state, output hashes, invalidation fields, and release-state transitions.
- Render self-contained outputs without runtime remote assets.
- Implement redaction manifest and recipient-specific profile handling.
- Maintain snapshot/release family error and reason-code registries.

Validation commands:
- Profile-selected backend unit/store/integration/process tests.
- Browser or E2E tests for progress, release actions, and output inspection if UI is exposed.
- `make phase-slice PHASE=phase11`
- `make service-backed-slice PHASE=phase11`
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `git diff --check`.

Deliverables:
- Snapshot and release route family.
- Immutable snapshot and release durable state.
- Self-contained output and redaction evidence.
- Direct non-aggregate evidence for every Snapshot and Reporting delta family.

Risks and assumptions:
- Live-recipient-specific workbook withholding is out of scope and non-conformant.
- Release approvals bind to the exact release tuple and rendered bytes.
- Core 05 applies only if timed or fixture-sensitive claim publication is made.

Exit criteria:
- Passing Base claim evidence is identified.
- All Snapshot and Reporting delta ACs have direct evidence.
- The profile can be claimed independently without implying other extension profiles.

## Sprint 4. Reference Pack Extension Profile

Objective: Implement and prove the Reference Pack Extension Profile only when selected.

Relevant IDs:
- `AC-033..AC-035`
- `AC-092..AC-096`
- `AC-234`
- `AC-270..AC-272`
- `AC-308..AC-310`
- `AC-326`
- `AC-369`

Files and areas:
- Likely areas: `internal/modules/reference_data`, reference-pack storage root configuration, archive verification helpers, generated public contracts, and optional UI controls.
- TODO: Confirm concrete source owners, fixture bundle format, and disconnected test harness strategy before implementation.

Test-first sequence:
1. Prove list/read route shapes, paging, sorting by `pack_key asc` then exact `pack_version asc`, singleton pagination rejection, and exact `reference_pack_version` resource.
2. Prove `POST /api/v1/reference-packs/import` shared upload envelope, `activation_policy='staged_only'`, no auto-activation, file hash idempotency, and terminal `reference_pack_imported` summary.
3. Prove activation, disable, reverify, and refresh validation, route-scoped idempotency, legal conditions, derived `active`, and exact resource refs.
4. Prove durable metadata: `pack_key`, `pack_kind`, `pack_version`, `manifest_sha256`, canonical `payload_sha256`, verification method/result, source identifier, signer key, and prior active version.
5. Prove disconnected operation, smallest disconnected bundle expectations, integrity failure handling, prior-active retention, and archive/path/content-screening failures.

Implementation tasks:
- Add the route family exactly under `/api/v1/reference-packs`.
- Persist or derive durable condition vocabulary: `staged`, `verified_available`, `disabled`, `failed`, `missing`.
- Keep `active` derived from activation pointer, not as an additional durable condition.
- Stage and verify local bundles, enforce archive/resource limits, and reject active content/path traversal/integrity failures.
- Maintain reference-pack family error and reason-code registries.

Validation commands:
- Profile-selected backend unit/store/process tests.
- Disconnected/no-network integration tests.
- Browser or E2E tests for operator controls if UI is exposed.
- `make phase-slice PHASE=phase11`
- `make service-backed-slice PHASE=phase11`
- `make generate-drift`, `make migration-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `git diff --check`.

Deliverables:
- Reference-pack route family and durable state.
- Verification and activation lifecycle evidence.
- Direct non-aggregate evidence for every Reference Pack delta family.

Risks and assumptions:
- Pack-dependent workbook overlays must not become Base Profile surfaces.
- Live internet fetch must not be required for disconnected verification or activation.
- Prior active version must remain retained on failed candidate activation.

Exit criteria:
- Passing Base claim evidence is identified.
- All Reference Pack delta ACs have direct evidence.
- The profile can be claimed without requiring Import, Snapshot and Reporting, Incident Portability, or Enterprise Authentication.

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
- Import remediation finalization evidence is recorded here because Import is closed.
- Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication are still open implementation sprints and are not closed out by this status record.

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
- Profile-selected final evidence for Import and common jobs.
- Refreshed generated artifacts through canonical commands.
- Updated plan/handoff status for selected, claimed, and still-open profiles.
- Explicit blocker list: no open blocker for the selected Import API route family.
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
- `make explain-phase PHASE=phase11` reports an active manifest with Import/common-job rows only.
- `make phase-slice PHASE=phase11` and `make service-backed-slice PHASE=phase11` execute the selected rows.
- `make check` passed at retained root `.cartulary/test-results/20260523T161735Z-p3268642`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260523T161735Z-p3268642` passed at retained root `.cartulary/test-results/20260523T162808Z-p3327944`.
- Do not fabricate successful retained run roots.
