# Recovery Specification and Implementation Remediation Tracker

## 1. Control

- **Controlling artifact:** this file.
- **Execution base:** `d45f3fbf`.
- **Active branch at execution start:** `main`.
- **Program status:** `COMPLETE`.
- **Active workstream:** none; remediation complete.
- **Execution rule:** workstreams run strictly in the order in section 5.
- **Failure rule:** if a workstream exit criterion cannot be met, mark that
  workstream `BLOCKED`, record the evidence, and stop.
- **Domain decision:** `docs/domain.md` remains unchanged. Backup and Restore
  remain generic contexts, and this remediation adds neither domain vocabulary
  nor a browser route or capability.

The planning-only tracker at base commit `d45f3fbf` is superseded. Its
RS-00--RS-08 behavior-preservation proposal remains available in repository
history but is not authorization for implementation. This tracker authorizes
the forward-looking RS-00--RS-12 and RS-99 sequence below.

### 1.1 Required slice protocol

For every workstream:

1. Confirm the preceding tracker-update commit exists.
2. Implement only the active workstream.
3. Run the narrow validation named for that workstream.
4. Commit the substantive specification, contract, implementation, test,
   harness, or documentation changes as one rollback unit or an explicitly
   recorded series.
5. Update sections 5, 7, and 8 with status, substantive changes, compatibility
   impact, commands and results, run roots, rollback state, and exact
   substantive commit.
6. Run `make lint-markdown`.
7. Commit the tracker update before beginning the next workstream.

A tracker commit cannot contain its own hash. The exact substantive commit is
therefore recorded in the completed workstream row; the following commit in
history is the required tracker ledger commit.

### 1.2 Non-negotiable compatibility boundary

- Recovery gains no HTTP or WebSocket route, browser capability, scheduler, or
  sixth operator command.
- The five operator commands, result and progress schema v1, flags, defaults,
  sort order, output streams, and exit codes remain stable.
- New backup writers emit only vNext formats. Strict historical readers remain
  until no retained backup metadata references them.
- Historical readers do not guess, normalize, alias, or coerce tokens.
- MinIO-source migration support is removed immediately. A deployment needing
  that transition must migrate externally before upgrade or remain on an older
  release.
- The MinIO Go SDK remains because it is the current S3-compatible client.
- No database migration is planned. Discovering a need for one blocks the
  active slice until the tracker records an independently justified schema
  change.
- Runtime, tests, generation, conformance, and release evidence must not depend
  on this Markdown file or any other file under `docs/`.

## 2. Adopted Design Decisions

| Gap | Remediation and owning areas | Long-term rationale and benefit | Compatibility or migration impact | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- | --- | --- |
| Planning-only tracker freezes behavior | Make this tracker the sequential execution ledger. **Documentation.** | Prevents known defects from becoming compatibility constraints and makes execution resumable. | No product impact; old plan remains in Git history. | Work can run out of order or silently retain nonconformance. | One active sequence, no unresolved design choice, Markdown lint green. |
| Cross-owner table predicate | Adopt owner-authored `recovery_state_contribution.v1` inputs and one frozen `recovery_state_catalog.v1`. **Specification, contracts, implementation, tests.** | Makes ownership and rebuild policy explicit, rejects future unclassified state, and supports new storage kinds without Recovery learning owner schemas. | New backups capture the exact 82-table set in section 3. Historical artifacts retain their original interpretation. | New state can be silently omitted or derived/transient state treated as authority. | All 109 authored tables classify exactly once; unknown, duplicate, and missing contributions fail closed. |
| Incoherent, memory-heavy capture | Use one read-only repeatable-read PostgreSQL snapshot for rows and owner object inventories; stream immutable referenced objects and table members. **Specification, contracts, implementation, security, tests.** | Makes `consistency_point_at` real, excludes orphans, and bounds memory for large deployments. | New vNext manifests, units, summaries, and envelope; retained v1/v2 readers remain. | Published backups can mix points in time, omit bytes, capture stale data, or exhaust memory. | Concurrent-write, mutation, missing-object, extra-object, large-object, and zero-byte tests pass. |
| Go-only persisted shapes | Add strict schemas, canonical fixtures, registries, bounds, and historical codec rules under `contracts/recovery`. **Specification, contracts.** | Makes durable bytes versioned and independently reviewable. | Current schema IDs become read-only historical codecs at cutover. | Refactors can silently alter bytes, omission semantics, or redaction. | Unknown/duplicate member, bound, canonical byte, digest, and drift tests pass. |
| Fake stopped-target proof | Require bound marker v2 and a shared-server/exclusive-restore serving lease. **Specification, platform, operator tooling, tests, documentation.** | Proves the target is correct and not serving without a startup race. | Ordinary and verification restore require marker v2; the old marker is rejected. | A live or misaddressed deployment can be overwritten. | Active server, purpose, binding, stale marker, and concurrent startup fail before mutation. |
| CLI DTO and error-text coupling | Add transport-neutral requests, results, progress, and typed `FailureKind`; leave parsing and wire encoding in Operator. **Implementation, tests.** | Gives Recovery a cohesive API and exhaustive failure mapping. | Wire v1 stays stable; invalid journal/audit code-reason combinations disappear. | Wording changes can alter protocol behavior or produce non-adopted pairs. | Every closed pair has a typed test; no error-string routing or CLI DTO dependency remains. |
| Open journal/audit maps | Adopt typed admission/completion records and atomically write terminal encrypted journal plus safe audit derivative. **Specification, contracts, implementation, security, tests.** | Prevents secret leakage and partial terminal evidence. | New private payload v2; envelope v1 and historical rows remain readable. | Attempts can lack closure evidence or journal and audit can disagree. | Failure-injection atomicity and exact-field/redaction tests cover all operations. |
| SQLC/platform leakage | Add private repositories, owner object capabilities, streaming storage, and target-admission ports. **Implementation, tests, boundary policy.** | Keeps semantics cohesive and concrete adapters independently replaceable. | Authored SQL stays unchanged unless separately justified. | Recovery remains a universal storage/configuration owner. | Semantic APIs expose no SQLC/platform store/settings types; resource-lifetime tests pass. |
| Hard-coded, double-selected workbook probe | Adopt one owner registration: Timeline owns query, Workbook validates/executes, Recovery selects once. **Specification, contracts, implementation, tests.** | Preserves the valuable probe while supporting future registered probes. | Selector, query, zero-incident skip, and zero-row success remain exact. | Artifact identity can diverge from the incident/query actually executed. | Exact fixture and conflict tests pass; Recovery has no Timeline implementation import. |
| One batch-wide due-verification context | Snapshot due order, then use per-backup deadline, lock, admission, journal pair, attempt ID, reset, and aggregation. **Specification, implementation, tests.** | Bounds failures and satisfies the per-verification contract. | Final CLI envelope stays v1; no-due keeps safe scheduler evidence. | One slow item consumes the batch and later attempts lack evidence. | Independent deadline, lock, evidence, continuation, unsafe-stop, reset, and order tests pass. |
| Retired migration in Recovery | Remove code, tests, six schemas, target, release dependency, occurrence policy, and proposed owner. **Specification, implementation, tests, harness, documentation.** | Avoids permanent security and ownership cost for a command that does not exist. | Immediate breaking removal for MinIO-source deployments; current S3 client remains. | A retired transition becomes a permanent subsystem and release gate. | No migration state/schema/target remains; compatibility and release checks stay green. |
| Incomplete test ownership | Add exact rows for each durable behavior and remove migration routing. **Tests, harness.** | Gives every postcondition one accountable owner. | Generated routing/topology may change only through generators. | Regressions can be unowned or multiply owned. | Every symbol selects exactly once and collaborator/topology checks pass. |

## 3. Recovery-State Catalog Target

Every one of the 109 authored public base tables must have exactly one source
owner and one classification. The vNext capture set contains exactly 82 tables.

| Owner | Count | Required vNext tables |
| --- | --- | --- |
| `artifacts` | 5 | `artifact_findings`, `artifact_forensic_keywords`, `artifact_investigative_queries`, `artifacts`, `handoff_risk_refs` |
| `assessments` | 1 | `assessments` |
| `audit` | 1 | `administrative_audit_projections` |
| `auth` | 5 | `account_preferences`, `bootstrap_tokens`, `enterprise_auth_bindings`, `enterprise_auth_providers`, `users` |
| `deployment_admin` | 2 | `deployment_admin_audit_events`, `deployment_bootstrap_state` |
| `entities` | 5 | `entity_aliases`, `entity_mentions`, `entity_preserved_identifiers`, `hosts`, `identities` |
| `evidence` | 3 | `evidence`, `evidence_custody_events`, `object_blobs` |
| `extensions` | 6 | `extension_job_cancellation_observations`, `extension_job_commit_proofs`, `extension_migration_ledger`, `extension_staged_object_references`, `extension_staged_objects`, `extension_state_metadata` |
| `imports` | 6 | `import_apply_journal`, `import_apply_unit_plans`, `import_sessions`, `import_source_streams`, `import_unit_apply_outcomes`, `import_units` |
| `incidentbundles` | 5 | `incident_bundle_exports`, `incident_bundle_imported_actors`, `incident_bundle_imported_attributions`, `incident_bundle_job_payloads`, `incident_bundle_manifest_files` |
| `incidents` | 4 | `incident_memberships`, `incident_workbook_preferences`, `incidents`, `user_workbook_preferences` |
| `indicators` | 3 | `indicator_observations`, `indicator_state_intervals`, `indicators` |
| `links` | 2 | `record_links`, `record_tags` |
| `networkflow` | 4 | `network_flow_indicator_bindings`, `network_flow_rejected_row_diagnostics`, `network_flow_rows`, `network_flow_tables` |
| `parties` | 1 | `parties` |
| `platform_jobs` | 1 | `jobs` |
| `recovery` | 1 | `operator_recovery_journal` |
| `reference_data` | 4 | `reference_pack_activation_state`, `reference_pack_attestations`, `reference_pack_job_payloads`, `reference_packs` |
| `reportcomposition` | 4 | `report_composition_preview_attempts`, `report_composition_release_bindings`, `report_composition_versions`, `report_compositions` |
| `reporting` | 8 | `reporting_composition_preview_output_files`, `reporting_composition_preview_outputs`, `reporting_job_payloads`, `reporting_release_approvals`, `reporting_releases`, `reporting_render_bundle_files`, `reporting_render_bundles`, `reporting_snapshots` |
| `revisions` | 5 | `change_set_mutations`, `change_sets`, `record_history_entry_refs`, `record_revisions`, `records` |
| `savedviews` | 1 | `saved_views` |
| `tasksdecisions` | 2 | `decisions`, `task_requests` |
| `timeline` | 3 | `timeline_events`, `timeline_source_provenance`, `timeline_time_conversion_profiles` |
| **Total** | **82** | No additional table. |

The current 92-table set loses ten transient or derived tables:

- all five `graph_projection_*` tables, rebuilt from authoritative state;
- all four `collaboration_*` durable-stream tables, invalidated across the
  restore generation; and
- `enterprise_auth_transactions`, invalidated as stale protocol correlation
  state.

The seven exact exclusions remain `backup_sets`, `evidence_access_handles`,
`pending_totp_enrollments`, `restore_verification_runs`, `route_idempotency`,
`schema_migration_lineage`, and `user_sessions`. Synthetic
`goose_db_version` remains excluded. The ten `*_grid_projection` exclusions
remain:

```text
artifact_grid_projection
assessment_grid_projection
decision_grid_projection
evidence_grid_projection
host_grid_projection
identity_grid_projection
indicator_grid_projection
party_grid_projection
task_request_grid_projection
timeline_grid_projection
```

Object contributions cover Evidence blobs, import source streams, Extension
staged objects, Incident Bundle files, Reference Pack members, and Reporting
render/preview members. Unclaimed object-store members are not silently copied:
they are explicitly transient or cause coverage failure before publication.

## 4. Versioned Contract Target

The active owner work must adopt and project these identities:

| Identity | Purpose |
| --- | --- |
| `cartulary.recovery_state_contribution.v1` | Source-owner table, object, restore, rebuild, and invalidation contribution. |
| `cartulary.recovery_state_catalog.v1` | Frozen complete catalog and canonical digest. |
| `cartulary.restore_workbook_probe_registration.v1` | Owner query registration executed by Workbook. |
| `cartulary.restore_target_marker.v2` | Purpose, target generation, and non-secret database/object binding digests. |
| `cartulary.backup_artifact_envelope.v2` | Chunked authenticated streaming envelope. |
| `cartulary.backup_integrity_manifest.v3` | Backup-wide vNext identity and proofs. |
| `cartulary.postgres_snapshot_artifact.v2` | Structured snapshot index. |
| `cartulary.postgres_snapshot_unit.v1` | One streamed canonical table unit. |
| `cartulary.object_store_backup_manifest.v2` | Private owner-selected object restore manifest. |
| `cartulary.object_store_backup_summary.v2` | Safe redacted manifest derivative. |
| `cartulary.operator_recovery_journal_payload.v2` | Typed private admission/completion evidence. |
| `cartulary.operator_recovery_audit_summary.v2` | Typed safe administrative-audit derivative. |
| `cartulary.restore_verification.v2` | Verification identity bound to catalog, codecs, and executed probe. |

Envelope v2 uses fixed 4 MiB plaintext chunks, HKDF-SHA256 to derive a
per-artifact AES-256 key from a random 32-byte salt, and AES-GCM nonces composed
from a random 8-byte prefix plus a big-endian `uint32` chunk index. AAD binds
schema ID, logical reference, content type, chunk index, plaintext length, and
final-chunk flag. Zero-byte input emits one authenticated final chunk.

Current backup manifest v2, PostgreSQL snapshot v1, object snapshot v2, backup
envelope v1, object manifest/summary v1, restore verification v1, and journal
payload/envelope formats remain strict historical readers where applicable.
New writers stop emitting them at RS-09. All six object-store migration
artifact identities are removed in RS-02.

## 5. Sequential Workstreams

| ID | Scope and dependency | Narrow validation | Exit criterion | Status | Substantive commit |
| --- | --- | --- | --- | --- | --- |
| RS-00 | Convert tracker and capture executable baseline. Depends on `d45f3fbf`. | Recovery unit/service slices; browser restore; operational smoke; migration preservation before deletion; boundary check; Markdown lint. | Tracker is executable and every later slice has scope, compatibility, rollback, validation, and exit fields. | DONE | `a1ebb471` |
| RS-01 | Adopt Core 00/01/04, Extension, Graph Projection, Testing Harness, schemas, fixtures, and generated projections. Depends on RS-00. | Markdown, JSON/schema/contract tests, `make generate-drift`, artifact policy. | Owners, strict contracts, and historical rules are authoritative outside this tracker. | DONE | `fe74a282` |
| RS-02 | Remove legacy migration implementation, tests, schemas, policy, target, and release dependency. Depends on RS-01. | Target/occurrence scans; harness; SeaweedFS compatibility; release projection. | No MinIO-source migration behavior or evidence remains. | DONE | `70cfa8d3` |
| RS-03 | Add typed Recovery facade and Operator wire adapter. Depends on RS-02. | Recovery owner slice and `make build-operator`. | Semantic operations have typed failures/progress and wire v1 parity. | DONE | `8a5a9dad` |
| RS-04 | Add private repositories and typed, atomic terminal evidence. Depends on RS-03. | Recovery service slice, journal/audit security tests, targeted gosec. | No SQLC/raw Evidence leakage; terminal journal/audit is atomic and secret-safe. | DONE | `73e42313` |
| RS-05 | Add bound target marker v2 and shared/exclusive serving lease. Depends on RS-04. | Recovery/server/operator slices and operational recovery smoke. | Every restore holds real target-safety proof before mutation. | DONE | `b0edf689` |
| RS-06 | Aggregate complete recovery-state catalog in shadow mode. Depends on RS-05. | Catalog fixtures, owner slices, boundary and generation drift. | All 109 tables and object families classify exactly once; 82 tables are required. | DONE | `a2c6db0e` |
| RS-07 | Add streaming backup storage and envelope v2 beside v1. Depends on RS-06. | Codec/security/fuzz-style cases, bounded-memory test, targeted gosec. | Corruption/reordering/truncation/duplication/trailing data fail closed; memory is bounded. | DONE | `38661274` |
| RS-08 | Add non-default repeatable-read vNext capture and restore codecs. Depends on RS-07. | Recovery service slice and vNext end-to-end fixtures. | vNext works in parallel without changing the active writer. | DONE | `b7a27f3a` |
| RS-09 | Atomically cut writer/reader selection to vNext and retain strict historical readers. Depends on RS-08. | New/historical restore, concurrent capture, owner rebuild/invalidation tests. | New backups are coherent vNext; retained historical backups restore. | DONE | `5f2f8ee3` |
| RS-10 | Add Timeline registration, Workbook registry/executor, and single Recovery selection. Depends on RS-09. | Recovery, Timeline, Workbook owner slices and browser restore. | Exact query behavior passes without Recovery importing Timeline implementation. | DONE | `a10c723a` |
| RS-11 | Give each due backup an independent attempt lifecycle. Depends on RS-10. | Recovery due-batch unit/service tests. | All safe selected backups are attempted with complete per-item evidence. | DONE | `391c0a8e` |
| RS-12 | Privatize leftovers and update boundary/verification ownership and generated topology. Depends on RS-11. | Owner explanations/slices, boundary, harness, generation and artifact drift. | Exact-once ownership, valid collaborators, and no obsolete shim remain. | DONE | `a4e2e194` |
| RS-99 | Run full validation and complete handoff. Depends on RS-12. | Commands in section 6. | All required checks pass or have owner-approved exception; no TODO/BLOCKED remains. | DONE | `7d9950a5` |

Every workstream is its own rollback unit. A later slice may use multiple
substantive commits only when section 7 records their order and the tracker
ledger lists the complete range.

## 6. Validation and Acceptance

### 6.1 Final validation

RS-99 runs `make agent-finalize` first, using the successful warm-check
`RESULTS_DIR` when one exists. It then runs:

```text
make test-slice OWNER=module.recovery
make service-backed-test-slice OWNER=module.recovery
make backend-module-boundary-check
make generate-drift
make generated-artifact-policy-check
make harness-contract
make go-gosec-targeted
make build-operator
make browser-e2e-webserver-backed
make standup-operational-recovery-smoke
make seaweedfs-compatibility
make check
make release-check
```

Run `make migration-drift` only if a separately justified migration was
introduced. Otherwise record that it was skipped because no schema change was
needed.

### 6.2 Required scenarios

- exact 109-table classification and 82-table capture set;
- rejection of an unclassified table or object family before publication;
- coherent concurrent database/object capture and deterministic restore;
- multi-chunk, zero-byte, wrong-key, corruption, reorder, duplicate,
  truncation, trailing-data, and digest failures;
- historical backup restoration during the retained window;
- active server, missing/wrong/stale marker, wrong binding, startup race, and
  lease-loss rejection;
- every closed CLI error mapping without text inspection;
- atomic, secret-safe journal and administrative-audit completion;
- exact Timeline registration, zero-incident skip, zero-row success, and
  browser restore;
- due ordering, independent timeout/lock/evidence, safe continuation,
  indeterminate stop, and reset;
- absence of migration code, schemas, tests, task surface, and release
  dependency; and
- exact-once verification ownership, clean generated artifacts, and final
  repository/release success.

## 7. Workstream Evidence Ledger

Complete one row before its tracker-update commit. Run roots are repository
relative. Compatibility entries must name intentional breaks as well as
preserved surfaces.

| ID | Date | Status | Substantive changes | Compatibility impact | Validation and run roots | Rollback state | Exact substantive commit |
| --- | --- | --- | --- | --- | --- | --- | --- |
| RS-00 | 2026-07-29 | DONE | Replaced planning-only plan with active remediation ledger; recorded design decisions, 82-table target, contract identities, sequential gates, and baseline inventory. | Documentation only; old plan remains at `d45f3fbf`. | All focused baselines in section 8 passed; `make lint-markdown` passed at `.cartulary/test-results/20260729T152635Z-p1430749`. | Revert `a1ebb471`; no product or generated state changed. | `a1ebb471` |
| RS-01 | 2026-07-29 | DONE | Adopted coherent capture, source-owner catalog, vNext codecs, typed failures/evidence, bound marker/serving lease, workbook registration, per-attempt due verification, and immediate migration retirement in Core and subsystem owners. Added a private Recovery contract family with 14 strict schemas, 13 fixtures, generated Go projection, duplicate-member rejection, and migration-to-catalog exact coverage validation. | Normative current behavior changes to the vNext target; implementation remains intentionally transitional until later slices. No public/browser contract or database schema changed. Historical recovery schema IDs are read-only; migration schema prefix is retired. | `make generate` passed at `.cartulary/test-results/20260729T155013Z-p1452942`; `make json-shape-check` passed at `.cartulary/test-results/20260729T155025Z-p1455182`; `make generate-drift` passed at `.cartulary/test-results/20260729T155031Z-p1455754`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260729T155050Z-p1459597`; `make test-fast` passed 933 tests at `.cartulary/test-results/20260729T155057Z-p1460071`; `make lint-markdown` passed at `.cartulary/test-results/20260729T155334Z-p1513610`; `make lint-biome` passed. | Revert `fe74a282`; generated Recovery and Extension projections revert with their authored inputs. No persistence migration exists. | `fe74a282` |
| RS-02 | 2026-07-29 | DONE | Deleted the legacy Recovery object-store transition implementation and its unit/integration tests; removed its six artifact identities, command target, release-DAG edge, duration baseline, topology claim, and transition-specific occurrence classifications. Replaced release evidence with current SeaweedFS compatibility, Recovery backup-integrity contract coverage, redaction, and storage-reference ownership checks; regenerated task and topology projections. | Immediate intentional break for deployments still requiring a MinIO-source transition: migrate externally before upgrade or remain on an earlier release. The current MinIO Go SDK remains as the S3-compatible client. The five Recovery commands, browser surface, database schema, current SeaweedFS backend, and retained historical backup readers are unchanged. | Exact repository token/target scans passed. `make generate` passed at `.cartulary/test-results/20260729T160301Z-p1520533`; `make generate-drift` passed at `.cartulary/test-results/20260729T160330Z-p1523039`; `make harness-contract` initially found the intentional public-target digest change at `.cartulary/test-results/20260729T160344Z-p1526944`, then passed after updating that digest at `.cartulary/test-results/20260729T160426Z-p1528367`; `make lint-biome` passed at `.cartulary/test-results/20260729T160507Z-p1529733`; `make lint-scripts` passed at `.cartulary/test-results/20260729T160515Z-p1530341`; `make test-fast` passed 933 tests at `.cartulary/test-results/20260729T160525Z-p1530916`; `make seaweedfs-compatibility` first stopped on a stale July 26 repository-owned proxy at `.cartulary/test-results/20260729T160758Z-p1587792`, then passed after terminating that stale process at `.cartulary/test-results/20260729T160839Z-p1588799`; `make seaweedfs-release-gate` passed at `.cartulary/test-results/20260729T160852Z-p1589656`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260729T161011Z-p1592358`; `make json-shape-check` passed at `.cartulary/test-results/20260729T161018Z-p1592737`. | Revert `70cfa8d3` as one rollback unit. No persistence migration or external state mutation was introduced; the stale local proxy was terminated before the successful compatibility rerun. | `70cfa8d3` |
| RS-03 | 2026-07-29 | DONE | Introduced a transport-neutral Recovery application facade with five operation-specific request methods, typed results, progress sink, and a closed 39-kind failure catalog. Moved parsing, flags, JSON/JSONL encoding, diagnostic wording, wire error pairs, sorting, and exit selection under the Operator facade. Replaced message-fragment routing with typed sentinels and restore-stage failures, exhaustively tested every Core failure pair, and added boundary rules prohibiting Operator transport imports and error-text routing in the Recovery application package. | The exact five commands, flags, defaults, result/progress schema v1, stdout/stderr split, sorting, and exit codes remain stable, as confirmed by canonical process evidence. The prior non-Core `journal_append_failed` and `audit_append_failed` combinations can no longer be emitted. No route, browser, persistence schema, backup format, or historical decoder changed. | `make build-operator` passed at `.cartulary/test-results/20260729T162652Z-p1618986`; `make test-slice OWNER=module.recovery` passed 11 tests at `.cartulary/test-results/20260729T162734Z-p1628773`; `make test-fast` passed 933 tests at `.cartulary/test-results/20260729T162903Z-p1657799`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260729T163221Z-p1715587`; `make backend-unit` passed 421 tests at `.cartulary/test-results/20260729T163308Z-p1716396`; `make service-backed-test-slice OWNER=module.recovery` passed 9 tests at `.cartulary/test-results/20260729T163335Z-p1719537`; `make lint-markdown` passed at `.cartulary/test-results/20260729T163617Z-p1747850`. Exact source scans found no former Recovery CLI/operations package imports and no error-text routing in the new application package. | Revert `8a5a9dad` as one rollback unit. Package moves and typed sentinels are source-level only; no persistent data or external state requires rollback. | `8a5a9dad` |
| RS-04 | 2026-07-29 | DONE | Replaced Recovery's open journal maps with exact typed v2 admission and completion records. Composition now encrypts journal payloads and commits the terminal journal plus its safe administrative-audit derivative in one PostgreSQL transaction. Added private backup and verification repository boundaries, moved Evidence object discovery and row counting behind an Evidence-owned recovery provider, confined SQLC to the Recovery store adapter, and added reverse-order resource-lifetime coverage. Added exact-field, sorting, redaction, atomic rollback, closed-result, terminal-write override, and resource-closure tests with exact owner rows. | New attempts write private journal payload v2 and safe audit summary v2 while retaining the journal envelope v1 and all historical rows as forensic evidence. A terminal evidence transaction failure maps to the existing operation-specific `journal_write_failed` pair. Operator wire v1, the five commands, current backup formats, database schema, and historical decoders are unchanged. No database migration was introduced. | Focused atomic evidence passed at `.cartulary/test-results/20260729T165526Z-p1859005`; typed failure coverage passed at `.cartulary/test-results/20260729T165652Z-p1866209`; resource lifetimes passed at `.cartulary/test-results/20260729T165417Z-p1853035`; Operator mapping passed at `.cartulary/test-results/20260729T165224Z-p1846327`; `make test-slice OWNER=module.recovery` passed 14 tests at `.cartulary/test-results/20260729T165703Z-p1866971`; the focused Recovery service slice passed 10 tests at `.cartulary/test-results/20260729T165810Z-p1895963`; full `make test-service-backed` passed at `.cartulary/test-results/20260729T170258Z-p1948822`; `make test-fast` passed all 934 rows at `.cartulary/test-results/20260729T170910Z-p2040770`; `make harness-contract` passed at `.cartulary/test-results/20260729T170805Z-p2034584`; `make go-test-duration-baseline-coverage` passed at `.cartulary/test-results/20260729T170849Z-p2036093`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260729T170910Z-p2040667`; `make go-gosec-targeted` passed at `.cartulary/test-results/20260729T170910Z-p2040818`; `make build-operator` passed at `.cartulary/test-results/20260729T170910Z-p2040957`; `make generate-drift` passed at `.cartulary/test-results/20260729T170849Z-p2035962`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260729T170849Z-p2035955`; and `make lint-markdown` passed at `.cartulary/test-results/20260729T171243Z-p2125735`. Exact production-source scans found no raw Evidence table tokens outside the Evidence owner and no SQLC import outside the Recovery store adapter. | Revert `73e42313` as one rollback unit. The change introduces no database migration or external data rewrite; reverting returns new writes to the former v1 open-map payload behavior while preserving already-written encrypted rows. | `73e42313` |
| RS-05 | 2026-07-29 | DONE | Replaced the fake `Stopped` target flag and verification-only marker v1 with a strict marker v2 plus separate target-generation proof for ordinary and verification restores. The marker binds purpose, canonical generation UUID, non-secret database/object binding digests, and a maximum 24-hour validity interval. Every PostgreSQL-backed server runtime now acquires and monitors a shared serving advisory lease before publication or listeners; restore acquires the exclusive counterpart, cancels mutation on lease uncertainty/loss, and holds the lease through the atomic terminal journal/audit transaction. Added strict nested duplicate-member JSON validation, deployment marker-generation tooling, startup/active-server/lease-loss coverage, exact verification rows, and refreshed generated scheduling evidence. | Intentional operational break: ordinary restore now requires `restore-target-marker.json` and `restore-target-generation`, and the former `restore-verification-target.json` v1 marker is rejected. Deployment tooling generates the v2 material after target preparation. The five commands and wire v1 remain byte-compatible, replicated servers remain supported through shared leases, credentials do not affect binding digests, and no route, browser capability, backup codec, historical decoder, or database schema changed. | `make test-slice OWNER=module.recovery` passed 16 tests at `.cartulary/test-results/20260729T180958Z-p2710279`; `make service-backed-test-slice OWNER=module.recovery` passed 11 tests at `.cartulary/test-results/20260729T181111Z-p2739379`; app-server unit and service slices passed 31 and 24 tests at `.cartulary/test-results/20260729T174731Z-p2352613` and `.cartulary/test-results/20260729T174902Z-p2381237`; canonical Operator process evidence passed at `.cartulary/test-results/20260729T174328Z-p2241179`; `make test-fast` passed 946 tests at `.cartulary/test-results/20260729T180539Z-p2641785`; `make standup-operational-recovery-smoke` passed at `.cartulary/test-results/20260729T181407Z-p2812705`; `make harness-contract` passed at `.cartulary/test-results/20260729T180827Z-p2701642`; duration coverage passed at `.cartulary/test-results/20260729T180826Z-p2701452`; final boundary, targeted gosec, operator/server build, generation-drift, and artifact-policy checks passed under `.cartulary/test-results/20260729T181348Z-*`. The first broad fast run exposed only a stale renamed selector and a generated schedule based on pre-refresh durations; both owner inputs were corrected, regenerated, and the full rerun passed. No migration was run because no schema change was introduced. | Revert `b0edf689` as one rollback unit. The change creates no database migration or external persistent rewrite; rollback requires restoring the v1 marker preparation procedure and removes the serving lease safety guarantee. | `b0edf689` |
| RS-06 | 2026-07-29 | DONE | Added 28 source-owner contributions covering all 109 authored tables and six authoritative object families, then assembled them into an immutable, canonically digested catalog behind the neutral `platform/recoverystate` boundary. Catalog construction validates closed fact combinations, exact fixture parity, duplicate/missing/unknown ownership, contribution digests, the 82-table vNext set, and object inventory/validation/restore algorithm identities. Operator assembly now freezes the catalog before any Recovery command, validates the live `public` base-table set against the catalog, and shadow-compares the active legacy 92-table snapshot without selecting the vNext writer. Added exact unit and real-database drift-rejection rows, source-owner import approvals, projection-manifest evidence, generated topology/timing updates, and a pressure-safe lease test cleanup discovered by the broad suite. | Backup and restore admission now intentionally fail closed when the deployed schema differs from the 109 authored tables plus `goose_db_version`, or when the current writer does not produce its exact historical 92-table set. The active writer, backup formats, object capture behavior, five commands, wire v1, routes, browser surface, historical readers, and database schema remain unchanged until later slices. A future table or object family must receive an adopted owner contribution before Recovery can admit it. No alias, guessing, migration, or persistent rewrite was introduced. | Catalog unit and real-database rows passed at `.cartulary/test-results/20260729T183132Z-p2901878` and `.cartulary/test-results/20260729T183141Z-p2903735`; Recovery owner slices passed 18 and 12 tests at `.cartulary/test-results/20260729T183554Z-p2936369` and `.cartulary/test-results/20260729T183710Z-p2967180`; exact Projections boundary and manifest rows passed at `.cartulary/test-results/20260729T184126Z-p3053676` and `.cartulary/test-results/20260729T185854Z-p3176546`; the pressure-safe serving-lease row passed at `.cartulary/test-results/20260729T185452Z-p3117351`. The broad gate first exposed an unapproved Projections import at `.cartulary/test-results/20260729T183822Z-p2995321`, then a two-session pool test-cleanup defect at `.cartulary/test-results/20260729T184138Z-p3056015`, and finally authored projection-manifest drift at `.cartulary/test-results/20260729T185507Z-p3118835`; each owner input or test-lifetime defect was corrected and its exact row passed before the final rerun. Final `make test-fast` passed 948 tests with no missing or unmapped evidence at `.cartulary/test-results/20260729T190453Z-p3277268`; `make generate`, duration coverage, generation drift, and artifact policy passed at `.cartulary/test-results/20260729T190207Z-p3231515`, `.cartulary/test-results/20260729T190215Z-p3233754`, `.cartulary/test-results/20260729T190217Z-p3233925`, and `.cartulary/test-results/20260729T190228Z-p3237747`. Final backend unit, harness, Operator build, targeted gosec, boundary, and JSON-shape checks passed under `.cartulary/test-results/20260729T190240Z-*`; `make lint-markdown` passed at `.cartulary/test-results/20260729T190849Z-p3331669`. No migration check ran because the slice introduced no schema change. | Revert `a2c6db0e` as one rollback unit, including its generated timing/topology projections. No persistent format, database state, or external resource changed; rollback removes fail-closed catalog admission and returns Recovery to the legacy predicate alone. | `a2c6db0e` |
| RS-07 | 2026-07-29 | DONE | Added a separate streaming backup-storage capability without widening the historical byte-slice interface. Envelope v2 writes fixed 4 MiB chunks directly into an atomic rooted-storage publication, derives one AES-256 key per artifact with HKDF-SHA256 and a random 32-byte salt, forms nonces from a random 8-byte prefix plus big-endian `uint32` index, and binds every required field in AAD. Reads authenticate and digest-check the complete envelope into a discard preflight before a second authenticated pass writes plaintext to a caller-owned staging sink. Plaintext and envelope sizes/digests, strict canonical member order, exact chunk sequence/finality, bounds, base64, algorithm identities, and trailing bytes all fail closed. Added an exact adversarial owner row and corrected the envelope schema's maximum chunk count from `2^32-1` to `2^32` and its maximum encoded ciphertext length from a four-byte-loose bound to the exact 5,592,428 characters. | This is additive and deliberately inactive: the current v1 writer/readers and `BackupStorage` methods are unchanged, and production has no caller of `WriteArtifactStream` or `ReadArtifactStream` until RS-08. Zero-byte v2 artifacts are newly representable as one authenticated final chunk. The stricter schema corrections affect no persisted v2 artifact because no v2 writer existed. The five commands, wire v1, routes, browser surface, database schema, historical readers, and managed-service posture are unchanged. | The exact adversarial row passed at `.cartulary/test-results/20260729T193758Z-p3631885`; the full Recovery unit and service-backed slices passed 19 and 12 tests at `.cartulary/test-results/20260729T192514Z-p3362332` and `.cartulary/test-results/20260729T192657Z-p3415810`. Final `make test-fast` passed 949 tests with no missing or unmapped evidence at `.cartulary/test-results/20260729T193306Z-p3513233`; the duration baseline refresh, generation, coverage, and harness contract passed at `.cartulary/test-results/20260729T193148Z-p3504939`, `.cartulary/test-results/20260729T193154Z-p3505153`, `.cartulary/test-results/20260729T193202Z-p3507397`, and `.cartulary/test-results/20260729T193218Z-p3511804`. Final generation drift and artifact policy passed at `.cartulary/test-results/20260729T193810Z-p3632322` and `.cartulary/test-results/20260729T193810Z-p3632324`; backend unit passed 431 tests at `.cartulary/test-results/20260729T193546Z-p3567148`; final boundary, JSON-shape, Operator build, and targeted gosec passed under `.cartulary/test-results/20260729T193638Z-*`; `make lint-markdown` passed at `.cartulary/test-results/20260729T193947Z-p3636782`. The boundary gate first rejected a redundant test-only generated-package import at `.cartulary/test-results/20260729T193546Z-p3567186`; contract shape/drift coverage remained in its proper owner, the import was removed, and the exact row plus boundary rerun passed. Source scans found no `io.ReadAll` or active production call of the v2 methods. No migration check ran because no schema change was introduced. | Revert `38661274` as one rollback unit, including the generated contract and timing/topology projections. Because the capability was not selected, rollback has no persisted v2 bytes, database state, or external resource to translate. | `38661274` |
| RS-08 | 2026-07-29 | DONE | Added a parallel vNext capture/restore engine over typed snapshot, inventory, streaming-storage, atomic-mutation, and catalog-algorithm boundaries. The PostgreSQL adapter owns one read-only repeatable-read transaction and streams all 82 catalog tables as canonical NDJSON header/row/trailer units. Six source owners now provide exact, catalog-checked object inventories; capture rejects missing, duplicate, changed, or multiply owned members before publishing the final manifest. Canonically self-digested PostgreSQL index v2, object manifest/summary v2, integrity manifest v3, and a canonically digested codec registry bind every unit to the frozen catalog. Restore authenticates all top-level and owner-object artifacts before entering one target-owned atomic staging boundary, then streams rows/objects and runs every catalog validator, rebuilder, and invalidator. Added exact end-to-end and production-registration evidence and generated topology accounting. | This slice is deliberately additive: no application composition calls `NewVNextCaptureService` or `NewVNextRestoreService`, so the current manifest v2/PostgreSQL v1/object snapshot v2 writer and historical readers remain active and unchanged until RS-09. The exact five commands, wire v1, routes, browser surface, database schema, and managed-service posture are unchanged. The Import owner currently exposes its database-inline source bytes through the same immutable inventory contract; no persistence migration or byte relocation was introduced. | The exact vNext row passed at `.cartulary/test-results/20260729T195619Z-p3721979`, and the vNext plus production six-provider catalog rows passed together at `.cartulary/test-results/20260729T200316Z-p3853845`. `make test-fast` passed 950 tests with zero missing evidence at `.cartulary/test-results/20260729T195841Z-p3735811`. Generation passed at `.cartulary/test-results/20260729T195532Z-p3716040`; final generation drift, artifact policy, boundary, and targeted gosec passed at `.cartulary/test-results/20260729T200219Z-p3822800`, `.cartulary/test-results/20260729T200219Z-p3822758`, `.cartulary/test-results/20260729T200220Z-p3823157`, and `.cartulary/test-results/20260729T200220Z-p3823256`; harness contract passed after updating its intentional row/selector totals at `.cartulary/test-results/20260729T200316Z-p3853941`; Markdown lint passed at `.cartulary/test-results/20260729T200504Z-p3856926`. The first exact fixture run exposed and corrected a self-digest preimage bug at `.cartulary/test-results/20260729T195544Z-p3718323`. The first fast run stopped only because a stale 14 GB `/tmp` Go build cache exhausted tmpfs at `.cartulary/test-results/20260729T195631Z-p3723878`; the exact cache was relocated, the suite passed, and the quarantined cache was then deleted. The first harness run correctly found its old 940/1,740 totals at `.cartulary/test-results/20260729T200220Z-p3823147`; the authored expectation now matches 941/1,741. Source scans find no production selector outside the vNext implementation. No migration check ran because no schema change was introduced. | Revert `b7a27f3a` as one rollback unit, including the generated topology index and authored owner row. Because vNext remains unselected, rollback has no published metadata row, retained backup, database state, or external migration to translate. | `b7a27f3a` |
| RS-09 | 2026-07-29 | DONE | Replaced the application backup-create path with the vNext capture assembly and atomically publishes only after every unit, object, index, summary, and integrity envelope succeeds. Existing backup metadata columns carry an explicit `backup-stream-v2://` transport selector plus plaintext proof; manifest v3 then selects the exact codec registry and catalog. Candidate durability authenticates all top-level vNext proofs before selection. Versioned restore dispatches exact-selector records to streaming vNext restore and all other records to the unchanged strict historical decoder. The vNext target clears only restore/rebuild/invalidate catalog tables, streams rows and owner objects, runs catalog algorithms, invokes existing owner projection rebuilders, and derives closed consistency evidence. Verification basis now binds catalog and codec digests. Canonical process evidence creates and restores vNext while retained verification fixtures continue to restore historical artifacts. | Intentional format cutover: every new backup is manifest v3/PostgreSQL index v2/object manifest and summary v2/envelope v2. No new historical v1/v2 backup is written. Retained metadata without the exact vNext selector continues through its strict historical codecs; there is no alias, normalization, or fallback guessing. Metadata reuse required no database migration. The five commands, wire v1, flags, streams, routes, browser surface, target marker, and serving lease remain unchanged. | The canonical process row passed after the cutover at `.cartulary/test-results/20260729T201405Z-p3895562`; full Recovery unit and service-backed slices passed 20 and 12 tests at `.cartulary/test-results/20260729T201512Z-p3907386` and `.cartulary/test-results/20260729T201512Z-p3907404`. Operator build and boundary checks passed at `.cartulary/test-results/20260729T201512Z-p3907775` and `.cartulary/test-results/20260729T201512Z-p3907528`. `make test-fast` passed 950 tests with zero missing evidence at `.cartulary/test-results/20260729T201706Z-p3981385`. Final generation drift, artifact policy, harness, targeted gosec, and JSON shape checks passed at `.cartulary/test-results/20260729T201958Z-p4045491`, `.cartulary/test-results/20260729T201958Z-p4045590`, `.cartulary/test-results/20260729T201958Z-p4045866`, `.cartulary/test-results/20260729T201958Z-p4046047`, and `.cartulary/test-results/20260729T201958Z-p4045544`. Production scans find no current application call to the monolithic PostgreSQL/object snapshot capture or historical capture service. No migration check ran because existing metadata columns carry the exact versioned selector and proofs without schema change. | Revert `5f2f8ee3` as one rollback unit. Rollback is safe only before relying on newly emitted vNext metadata; after publication, retain the RS-08 vNext reader or preserve this commit so those backups remain readable. Existing historical backups require no translation. | `5f2f8ee3` |
| RS-10 | 2026-07-29 | DONE | Added a neutral workbook-probe registration and execution contract, an exact Timeline-owned Base registration, and a Workbook-owned registry that validates view/query capabilities, rejects duplicate IDs or profile-default conflicts at composition, realizes the registered query, permits zero rows, and returns registration/view/row-count identity. Recovery now selects the incident once, passes only that identity to the injected executor, and binds the returned identity into `restore_verification.v2`. The current verification codec now includes a typed target/backup binding basis, catalog and codec digests, manifest identity, restored-object count, selected incident, and executed-or-skipped probe result; the strict v1 codec remains historical. Added exact Recovery, Timeline, and Workbook owner rows, canonical fixture coverage, an Operator artifact round-trip assertion, browser parity, and generated topology accounting. The broad gate also exposed and corrected the serving-lease acquisition edge where pool-session timeout was not normalized to the closed active-process result, plus cleanup ordering that could strand that failed test. | New verification attempts intentionally emit `cartulary.restore_verification.v2`; v1 remains a strict read-only historical codec. Recovery's internal verification API now accepts the typed basis rather than an opaque hash, while the Operator command/result/progress wire remains unchanged. The exact selector, two-field ascending sort, zero-incident skip, and zero-row success are preserved. Recovery imports neither Timeline/view-schema implementation nor Workbook implementation; application composition is the only concrete Workbook dependency. No command, route, browser capability, scheduler, database migration, or backup-reader compatibility change was introduced. | Final Recovery unit plus canonical process evidence passed at `.cartulary/test-results/20260729T210021Z-p151651`; Workbook and Timeline exact rows passed at `.cartulary/test-results/20260729T204503Z-p51217` and `.cartulary/test-results/20260729T204506Z-p51557`; browser restore passed at `.cartulary/test-results/20260729T204705Z-p63665`. The repaired serving-lease row passed at `.cartulary/test-results/20260729T210257Z-p200740`, and the complete `make test-fast` rerun passed 956 tests with zero missing evidence at `.cartulary/test-results/20260729T210310Z-p202183`. Final generation drift, boundary, targeted gosec, and Operator build passed at `.cartulary/test-results/20260729T210548Z-p266793`, `.cartulary/test-results/20260729T210548Z-p267037`, `.cartulary/test-results/20260729T210548Z-p267171`, and `.cartulary/test-results/20260729T210548Z-p267311`; artifact policy, JSON shape, and harness contract passed at `.cartulary/test-results/20260729T210107Z-p166943`, `.cartulary/test-results/20260729T210125Z-p171507`, and `.cartulary/test-results/20260729T210125Z-p171710`. Early gates correctly found the technical `record_id` default-sort exception, forbidden source-owner Workbook import placement, intentional generated digest/count drift, and the serving-lease timeout edge; each was structurally repaired and its exact/final gate passed. Source scans find one incident selector and no forbidden implementation import. No migration check ran because no schema change was introduced. | Revert `a10c723a` as one rollback unit, including the three owner rows and generated topology. No database or external-state migration exists. If v2 verification artifacts have become retained operational evidence, preserve the v2 decoder when rolling back the writer; historical v1 material requires no translation. | `a10c723a` |
| RS-11 | 2026-07-29 | DONE | Replaced the batch-wide due-verification context, advisory lock, target admission, and evidence pair with a snapshotted due plan and one complete lifecycle per selected backup. Every attempt now has a caller-generated UUID shared by its verification run, v2 artifact, encrypted journal admission/completion, and safe audit derivative; an independent deadline; a transaction-scoped exclusion lock; a fresh bound target admission; and an explicit reset decision. Determinate failures continue only after reset, release, and fresh re-admission; cancellation, deadline, lease uncertainty, mutation-stage uncertainty, reset failure, and terminal-evidence failure stop the batch. Aggregation preserves the first due-order failure and every safe attempt artifact while keeping the first selected `backup_set_id`. The no-due path retains scheduler journal/audit evidence without taking the mutating lock. A failure-artifact gap discovered by the process scenario added the closed v2 workbook skip reason `verification_failed_before_probe`, prevents incident selection from partial restores, and preserves the original typed failure. Added exact Recovery and Operator owner rows, due-order/deadline/unsafe-stop tests, process reset/re-admission/evidence tests, no-due lock proof, and refreshed generated topology and timing inputs. | The five commands, flags/defaults, result/progress schema v1, streams, exit mappings, artifact-ref sorting, routes, browser surface, and scheduler boundary remain unchanged. `--timeout-seconds` for `restore_verify_due` now applies independently per selected backup instead of once to the outer invocation. Current verification v2 adds one strict failure-only skipped-probe reason; passing artifacts and historical v1 readers are unchanged. No database migration, historical backup decoder change, or external data rewrite was introduced. | Focused due-batch and v2 artifact rows passed at `.cartulary/test-results/20260729T212844Z-p488696`; the Operator timeout-adapter row passed at `.cartulary/test-results/20260729T212429Z-p447777`; determinate continuation/reset/re-admission evidence passed at `.cartulary/test-results/20260729T212652Z-p462450`; canonical five-command process evidence, including per-attempt pairs and no-due lock independence, passed at `.cartulary/test-results/20260729T212852Z-p489188`. Full Recovery unit and service-backed slices passed 23 and 13 tests at `.cartulary/test-results/20260729T213111Z-p529348` and `.cartulary/test-results/20260729T212949Z-p500774`; full `make test-service-backed` passed at `.cartulary/test-results/20260729T213424Z-p564271`; `make test-fast` passed 961 tests with zero missing evidence at `.cartulary/test-results/20260729T214126Z-p700191`. Generation, duration coverage, harness contract, generation drift, artifact policy, JSON shape, boundary, Operator build, targeted gosec, and Markdown lint passed at `.cartulary/test-results/20260729T213918Z-p657880`, `.cartulary/test-results/20260729T213929Z-p660161`, `.cartulary/test-results/20260729T213936Z-p660464`, `.cartulary/test-results/20260729T214013Z-p661737`, `.cartulary/test-results/20260729T214028Z-p665551`, `.cartulary/test-results/20260729T214034Z-p665929`, `.cartulary/test-results/20260729T214042Z-p666533`, `.cartulary/test-results/20260729T214051Z-p666994`, `.cartulary/test-results/20260729T214107Z-p676811`, and `.cartulary/test-results/20260729T214546Z-p757670`. Early gates exposed the transaction query-interface mismatch, obsolete single-journal progress expectation, failure-artifact probe-state gap, intentional catalog totals, and missing process duration evidence; each authored input or implementation defect was corrected before the final passes. No migration check ran because no schema change was introduced. | Revert `391c0a8e` as one rollback unit, including the v2 failure-only reason, three owner rows, timing data, and generated topology. No persistent schema or external resource requires rollback; retain the updated v2 decoder if failure artifacts using the new reason have been retained. | `391c0a8e` |
| RS-12 | 2026-07-29 | DONE | Removed the dead exported due-verification interval and the unused historical v1 writer, privatized same-package object-manifest encoders and historical validators/schema constants, and retained the one strict v1 decoder as the only historical verification codec surface. Removed Recovery's current-artifact dependency on Timeline's exact view ID: verification v2 now binds any syntactically valid owner-registered view identity while the Timeline fixture remains exact at its owner boundary. Added direct strict-decoder evidence, updated exact selector accounting to 947 rows and 1,753 selectors, regenerated the Recovery contract and execution-topology projections, and added production boundary rules forbidding Recovery dependencies on `docs/` or `tools/` paths and resurrection of the superseded Timeline constant. | Intentional internal source cleanup removes unused exported writer/validator names; all affected packages are under `internal/` and had no live caller. New v2 verification artifacts can represent future registered workbook views without a Recovery change; the current Timeline registration and canonical fixture remain unchanged. Historical v1 artifacts remain strictly decodable, with no alias, normalization, or coercion. The five commands, wire v1, routes, browser surface, scheduler, database schema, current backup writer, and retained metadata require no migration. | The focused artifact/probe row passed at `.cartulary/test-results/20260729T215126Z-p774932`; the first boundary run at `.cartulary/test-results/20260729T215135Z-p776541` correctly caught the superseded token inside a historical internal constant name, and the corrected boundary passed at `.cartulary/test-results/20260729T215151Z-p776997`. Generation and JSON shape passed at `.cartulary/test-results/20260729T215202Z-p777329` and `.cartulary/test-results/20260729T215225Z-p779915`; harness contract first reported the intentional selector total change at `.cartulary/test-results/20260729T215232Z-p780535`, then passed with the exact 1,753 total at `.cartulary/test-results/20260729T215319Z-p782237`. Duration coverage passed at `.cartulary/test-results/20260729T215358Z-p783567`; final generation passed at `.cartulary/test-results/20260729T215404Z-p783762`. Full Recovery unit and service-backed slices passed 23 and 13 rows at `.cartulary/test-results/20260729T215413Z-p786012` and `.cartulary/test-results/20260729T215543Z-p816915`. Generation drift and artifact policy passed at `.cartulary/test-results/20260729T215659Z-p844935` and `.cartulary/test-results/20260729T215713Z-p848773`; `make test-fast` passed 962 tests with zero missing evidence at `.cartulary/test-results/20260729T215721Z-p849245`; final formatting and Markdown lint passed at `.cartulary/test-results/20260729T220009Z-p909777` and `.cartulary/test-results/20260729T220129Z-p913350`. Exact production scans found no superseded export/token and no `docs/` or `tools/` path dependency. No migration check ran because no database schema change was introduced. | Revert `a4e2e194` as one rollback unit, including the strict-decoder selector, generated contract/topology updates, and boundary rules. No persistence or external-state rollback is required; keep the v2 generic registered-view validation if a future non-Timeline probe has begun emitting retained artifacts. | `a4e2e194` |
| RS-99 | 2026-07-29 | DONE | Executed the complete final validation sequence from clean RS-12 ledger commit `391f2dc4`, then ran every directly declared Recovery collaborator owner slice, the reverse Projections collaborator slice, and Timeline's exact restore-probe registration row. Added the final validation table, compatibility posture, exact base-to-handoff changed-file inventory, and explicit seven-reader historical-decoder inventory. | Documentation and validation only. No product, wire, schema, generated artifact, persistent data, external resource, or deployment behavior changed in this slice. Historical reader removal remains gated by retained metadata; MinIO-source migration remains intentionally absent; `docs/domain.md` remains unchanged. | Every required final command passed at the run roots in section 9. The full warm `make check` passed 189/189 work units and 818 tests with zero missing evidence at `.cartulary/test-results/20260729T221327Z-p1066897`; `make release-check` passed 11/11 work units and 818 tests with zero missing evidence at `.cartulary/test-results/20260729T221625Z-p1168691`. All collaborator rows passed at the section 9.1 roots. `make lint-markdown` passed the substantive handoff at `.cartulary/test-results/20260729T223135Z-p1405942`. `make migration-drift` was correctly skipped because no database schema change or migration exists. No exception was required. | Revert `7d9950a5` to remove only the RS-99 handoff report and changed-file inventory. Product rollback remains the per-workstream units in this ledger; no final-slice external state requires rollback. | `7d9950a5` |

## 8. RS-00 Baseline

Repository state at start:

- `HEAD=d45f3fbf`;
- branch `main`, two commits ahead of `origin/main`;
- clean worktree;
- 29 files under `internal/modules/recovery`;
- 11 `module.recovery` owner rows, of which 9 are service-backed;
- legacy migration implementation and `seaweedfs-migration-preservation`
  release target present;
- no `contracts/recovery` directory.

| Command | Result | Run root or evidence |
| --- | --- | --- |
| `make task-guide ROLE=module-author OWNER=module.recovery` | PASS | Focused unit and service-backed slice commands confirmed. |
| `make explain-test-owner OWNER=module.recovery` | PASS | 11 rows; 9 service-backed. |
| `make test-slice OWNER=module.recovery` | PASS, 11 tests | `.cartulary/test-results/20260729T151530Z-p1326305` |
| `make service-backed-test-slice OWNER=module.recovery` | PASS, 9 tests | `.cartulary/test-results/20260729T151637Z-p1355383` |
| `make seaweedfs-migration-preservation` | PASS | `.cartulary/test-results/20260729T151742Z-p1382406` |
| `make backend-module-boundary-check` | PASS | `.cartulary/test-results/20260729T151834Z-p1383694` |
| `make browser-e2e-webserver-backed` | PASS, 2/2 work units | `.cartulary/test-results/20260729T151850Z-p1384114` |
| `make standup-operational-recovery-smoke` | PASS | `.cartulary/test-results/20260729T152337Z-p1412464` |

### 8.1 Baseline persisted identities

The pre-remediation writer/reader inventory is:

- `cartulary.backup_integrity_manifest.v2`;
- `cartulary.postgres_snapshot_artifact.v1`;
- `cartulary.object_store_snapshot_artifact.v2`;
- `cartulary.backup_artifact_envelope.v1`;
- `cartulary.operator_recovery_journal_envelope.v1`;
- `cartulary.object_store_backup_manifest.v1`;
- `cartulary.object_store_backup_summary.v1`;
- `cartulary.restore_verification.v1`;
- `cartulary.extension_backup_registry.v1`; and
- six `cartulary.object_store_migration_*.v1` identities, removed at RS-02.

### 8.2 RS-02 removal checkpoint

- Recovery now contains 26 source/test files; the transition implementation
  and its two test files were deleted.
- The authored and generated task surfaces expose no dedicated transition
  command, and the release DAG no longer allocates its service-validation
  branch.
- The occurrence policy is backend-current and retains only SDK, external
  endpoint, historical fixture, invalid, and unclassified categories.
- SeaweedFS release evidence now derives its storage claim from current
  compatibility, the Recovery backup-integrity contract registry, redaction,
  and source-owner storage-reference controls.
- Exact scans outside `docs/archive/**` and this tracker found no removed
  schema ID, state token, implementation symbol, environment input, target, or
  release dependency.
- `make lint-markdown` validated the completed ledger at
  `.cartulary/test-results/20260729T161204Z-p1594592`.

### 8.3 RS-05 target-safety checkpoint

- `RestoreTarget.Stopped` and `ErrRestoreTargetNotStopped` no longer exist;
  target safety is an application/platform capability rather than caller
  assertion.
- Server runtimes acquire the shared PostgreSQL serving lease before
  publication or listener startup. Ordinary and verification restore acquire
  the exclusive counterpart before marker validation and target inspection.
- Marker v2 and its separate generation proof are strict, bounded,
  purpose-specific, target-bound, short-lived, and credential-independent.
- Active serving processes, server startup during restore, missing or v1
  markers, wrong purpose/generation/bindings, stale markers, and lease loss
  fail before or cancel target mutation.
- Three exact owner rows and four selectors increased the catalog to 937 rows
  and 1,737 selectors; duration evidence and generated scheduling topology were
  refreshed through their authored generators.
- No database migration, new command, route, browser capability, backup format,
  or historical-decoder change was introduced.

### 8.4 RS-06 recovery-state checkpoint

- Twenty-eight source owners now contribute exactly 109 tables and six object
  families through typed, closed facts; the frozen catalog contains exactly 82
  `authoritative_required` table units.
- Catalog construction compares the assembled owner facts with the generated
  strict contract projection, computes real contribution and catalog digests,
  and rejects duplicate, missing, unknown, or internally inconsistent facts.
- Recovery assembly validates the live schema against all 109 authored tables
  plus `goose_db_version` before command semantics proceed. A real database test
  proves an unclassified future table fails this gate.
- The active legacy writer remains unchanged and emits its historical 92-table
  snapshot. Shadow validation proves that set is the vNext 82 plus the five
  Graph Projection, four Collaboration, and one enterprise-auth transitional
  exclusions.
- Owner declarations for Evidence blobs, import source streams, Extension
  staged objects, Incident Bundle files, Reference Pack members, and Reporting
  render/preview members are frozen now; their active inventories and bytes
  remain work for RS-08.
- Two exact owner rows increased the authored catalog to 939 rows and 1,739
  selectors. Duration evidence and generated scheduling topology were refreshed
  from a successful complete fast run.
- No database migration, writer/reader cutover, new command, route, browser
  capability, object-copy behavior, or historical-decoder change was
  introduced.

### 8.5 RS-07 streaming-storage checkpoint

- Envelope v2 encryption and decryption operate through a distinct streaming
  capability. The legacy `WriteArtifact` and `ReadArtifact` byte-slice methods
  remain the only methods used by the active writer and historical readers.
- Atomic filesystem publication hashes and counts bytes while writing and opens
  immutable regular files for bounded reads; neither layer materializes a
  complete artifact.
- The authenticated preflight rejects wrong keys, ciphertext corruption,
  reordered or duplicate chunks/indices, missing final chunks, truncation,
  trailing data, duplicate members, AAD changes, and plaintext or envelope
  digest mismatches before the destination receives bytes.
- Zero-byte input produces one index-zero, zero-length, authenticated final
  chunk. A 64 MiB-plus fixture round-trips with source reads and staging writes
  bounded to the fixed 4 MiB chunk size.
- The machine schema now permits all `uint32` nonce indices as exactly
  4,294,967,296 possible chunks and uses the exact maximum base64 length for a
  full chunk plus the 16-byte GCM tag.
- One exact owner row increased the authored catalog to 940 rows and 1,740
  selectors. Duration evidence and generated scheduling topology were refreshed
  from a successful complete fast run.
- No writer selection, database migration, new command, route, browser
  capability, object-capture behavior, or historical-decoder change was
  introduced.

### 8.6 RS-08 parallel-vNext checkpoint

- Recovery can now capture exactly 82 canonically ordered PostgreSQL units and
  all six owner inventories through one repository-owned repeatable-read,
  read-only snapshot.
- Every unit is canonical NDJSON with a table-bound header, canonical row
  records, and a trailer binding exact row count and row-record digest.
- Owner object bytes are streamed through envelope v2 and must match the
  snapshot-visible size and digest. Missing, changed, duplicate, unclassified,
  or multiply owned objects stop capture before the final integrity manifest.
- PostgreSQL index v2, object manifest/summary v2, integrity manifest v3, and
  the codec registry have deterministic domain-separated digest preimages.
- Restore authenticates every referenced artifact before target mutation, then
  streams rows and objects through a target-owned atomic staging boundary and
  executes every exact catalog restore/validation algorithm.
- The production assembly contains exactly six owner inventory registrations;
  missing or mismatched owner, family, or algorithm identities fail startup.
- One exact owner row increased the authored catalog to 941 rows and 1,741
  selectors. Generated topology and harness totals were updated from authored
  inputs.
- The active application still selects only the historical writer/readers.
  RS-09 owns the atomic selection cutover and retained historical dispatch.
- No database migration, new command, route, browser capability, scheduler, or
  public transport change was introduced.

### 8.7 RS-09 writer/reader cutover checkpoint

- Backup create now has one active writer: vNext capture over the frozen
  catalog and streaming envelope v2.
- Publication occurs only after every table, object, index, summary, and
  integrity artifact has been written and authenticated.
- Existing metadata columns carry an exact `backup-stream-v2://` transport
  selector. Manifest v3 remains the authority for schema, codec-registry, and
  catalog selection after decryption.
- Records without that exact selector retain their prior strict v1/v2
  interpretation. Malformed selectors fail closed and never fall back to a
  different decoder.
- Canonical process evidence now performs a real vNext create and restore.
  The same row's verification fixtures continue to exercise retained
  historical artifacts.
- Graph and Workbook projections rebuild from restored authoritative state;
  Collaboration and authentication protocol state is invalidated through the
  catalog-driven target reset.
- Restore verification basis now includes the exact recovery-state catalog and
  codec-registry digests.
- No database migration, command, route, browser capability, scheduler, or
  transport-wire change was introduced.

### 8.8 RS-10 owner-registered workbook checkpoint

- Timeline owns the exact Base restore registration and Workbook owns registry
  validation and query realization. The shared contract is implementation
  neutral.
- Recovery selects the first incident exactly once and receives only
  registration ID, view schema ID, and row count from the executor.
- Zero incidents produce an explicit skipped result. Zero query rows remain a
  successful executed probe.
- Current verification artifacts are strict v2 and bind the typed target and
  backup-storage basis, catalog, codecs, manifest, selected incident, and
  executed registration. Verification v1 remains a strict historical codec.
- Duplicate registrations and conflicting or missing profile defaults fail
  during application composition.
- Recovery source imports no Timeline/view-schema or Workbook implementation.
  The concrete registry is confined to application composition.
- Three exact owner rows increased the authored catalog to 944 rows and 1,747
  selectors. Generated topology and harness totals were updated from authored
  inputs.
- The five commands, wire v1, database schema, routes, browser capability, and
  scheduler surface are unchanged.

### 8.9 RS-11 per-backup due-verification checkpoint

- The due list is snapshotted once in required
  `consistency_point_at ASC, backup_set_id ASC` order. The first selected
  backup remains the final v1 result identity.
- Every selected backup receives a distinct attempt UUID, deadline context,
  transaction-scoped exclusion lock, marker-bound exclusive serving lease,
  journal admission/completion pair, safe audit summary, attestation decision,
  and reset decision.
- A determinate failure can continue only after target reset and lease release;
  the next attempt then performs a fresh complete admission. The first failure
  remains the final error even when later safe attempts succeed.
- Timeout, cancellation, serving-lease loss, database/object mutation-stage
  uncertainty, reset failure, and terminal-evidence failure stop later
  attempts. Timed-out or indeterminate targets are not reset or reused.
- Failed attempts publish a safe v2 verification artifact whenever artifact
  storage remains available. Failure before the workbook probe uses only
  `verification_failed_before_probe` and never selects an incident from a
  partial restore.
- A no-due invocation keeps its null-identity no-op journal and audit proof
  while acquiring no mutating-operation lock or target admission.
- Three exact owner rows increased the authored catalog to 947 rows and 1,752
  selectors. Generated topology and duration evidence were refreshed from
  authored inputs and a successful full service-backed run.
- No database migration, command, route, browser capability, scheduler, wire
  schema, or historical backup-decoder change was introduced.

### 8.10 RS-12 boundary and verification checkpoint

- Recovery's current verification artifact validates owner-returned workbook
  registration and view identities as typed identifiers; Timeline's exact view
  ID remains in its registration contract and canonical fixture rather than a
  Recovery constant.
- The unused historical v1 verification writer is gone. Its strict decoder,
  decoded shape, canonical-byte check, digest verification, closed
  vocabularies, and exact Timeline interpretation remain available for retained
  evidence.
- Dead or same-package-only Recovery exports were removed or privatized. No
  live production, test, tool, or command caller was changed to a compatibility
  wrapper.
- Backend boundaries now reject production Recovery source references to
  `docs/` and `tools/` paths and reject the removed current Timeline constant.
- The authored ownership catalog remains 947 rows and now has 1,753 exact
  selectors. Harness contract, duration coverage, collaborator routing, both
  Recovery slices, and generated topology all validate.
- No database migration, command, route, browser capability, scheduler, wire
  schema, historical decoder, or external data changed.

## 9. RS-99 Validation and Handoff Evidence

The final validation ran from clean commit `391f2dc4`. No owner exception was
needed, and no required gate failed. The exact changed-file inventory and
retained historical-reader posture are recorded in the
[changed-file inventory](recovery-module-refactor-changed-files.md).

| Command | Result | Run root |
| --- | --- | --- |
| `make agent-finalize` | PASS; generated output unchanged; retained-run maintenance skipped because `RESULTS_DIR` was unset before a successful warm check existed | `.cartulary/test-results/20260729T220230Z-p916727` |
| `make test-slice OWNER=module.recovery` | PASS, 23 rows | `.cartulary/test-results/20260729T220301Z-p923158` |
| `make service-backed-test-slice OWNER=module.recovery` | PASS, 13 rows | `.cartulary/test-results/20260729T220432Z-p951594` |
| `make backend-module-boundary-check` | PASS | `.cartulary/test-results/20260729T220549Z-p979433` |
| `make generate-drift` | PASS | `.cartulary/test-results/20260729T220555Z-p979739` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260729T220613Z-p983577` |
| `make harness-contract` | PASS | `.cartulary/test-results/20260729T220620Z-p984025` |
| `make go-gosec-targeted` | PASS | `.cartulary/test-results/20260729T220657Z-p985388` |
| `make build-operator` | PASS | `.cartulary/test-results/20260729T220709Z-p1008457` |
| `make browser-e2e-webserver-backed` | PASS, 2/2 sessions | `.cartulary/test-results/20260729T220721Z-p1018004` |
| `make standup-operational-recovery-smoke` | PASS | `.cartulary/test-results/20260729T221225Z-p1047147` |
| `make seaweedfs-compatibility` | PASS | `.cartulary/test-results/20260729T221313Z-p1065986` |
| `make check` | PASS, 189/189 work units and 818 tests with zero missing evidence | `.cartulary/test-results/20260729T221327Z-p1066897` |
| `make release-check` | PASS, 11/11 work units and 818 tests with zero missing evidence | `.cartulary/test-results/20260729T221625Z-p1168691` |
| `make migration-drift` | SKIPPED; no database schema change or migration was introduced | Not applicable |

### 9.1 Collaborator owner evidence

Every directly declared Recovery collaborator owner passed its complete
focused slice. Projections also passed as the reverse collaborator that
declares Recovery, and Timeline's exact registration row passed explicitly.

| Owner or row | Result | Run root |
| --- | --- | --- |
| `app.operator` | PASS, 5 rows | `.cartulary/test-results/20260729T222214Z-p1323287` |
| `app.server` | PASS, 31 rows | `.cartulary/test-results/20260729T222229Z-p1324904` |
| `harness.browser` | PASS, 14 rows | `.cartulary/test-results/20260729T222352Z-p1353753` |
| `module.workbook` | PASS, 87 rows | `.cartulary/test-results/20260729T222411Z-p1355438` |
| `platform.audit` | PASS, 2 rows | `.cartulary/test-results/20260729T222724Z-p1392894` |
| `platform.objectstore` | PASS, 2 rows | `.cartulary/test-results/20260729T222740Z-p1394406` |
| `platform.postgres` | PASS, 5 rows | `.cartulary/test-results/20260729T222759Z-p1395996` |
| `platform.rootedfs` | PASS, 2 rows | `.cartulary/test-results/20260729T222829Z-p1399659` |
| `web.application` | PASS, 58 rows | `.cartulary/test-results/20260729T222836Z-p1400070` |
| `module.projections` | PASS, 8 rows | `.cartulary/test-results/20260729T222918Z-p1401711` |
| Timeline restore-probe registration row | PASS, 1 row | `.cartulary/test-results/20260729T222953Z-p1404371` |

### 9.2 Final compatibility and operational posture

- `docs/domain.md` remains intentionally unchanged: no domain concept, route,
  browser capability, scheduler, or sixth command was introduced.
- Operator result/progress schema v1, flags, defaults, ordering, streams, and
  exit codes remain stable.
- New backup writes use only the vNext catalog, manifests, units, object
  inventory, and streaming envelope. Historical readers remain strict and
  read-only as inventoried in the handoff.
- MinIO-source migration support remains intentionally absent. Affected
  deployments must migrate externally before upgrade or stay on an earlier
  release.
- No database migration, persistent data rewrite, or owner-approved validation
  exception was needed.
- The validated implementation and substantive handoff range is
  `d45f3fbf..7d9950a5`. This final tracker-only ledger commit is the immediate
  successor and, by the section 1.1 protocol, records no self-referential hash.
- The worktree was clean at RS-99 entry, after validation, and after the
  substantive handoff commit. It must be clean again after this final
  tracker-only ledger commit.

## 10. Final Handoff Checklist

- [x] Every RS-00--RS-12 row is `DONE` with an exact substantive commit and
  following tracker ledger commit.
- [x] RS-99 validation is complete.
- [x] No TODO or BLOCKED status item remains.
- [x] All changed files and retained historical decoders are listed.
- [x] Every command result and run root is recorded.
- [x] Skipped `migration-drift` is justified, or its passing run is recorded.
- [x] `domain.md` no-change decision remains valid.
- [x] The worktree state and final commit range are recorded.
- [x] Program status is `COMPLETE`.
