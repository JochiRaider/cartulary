# Recovery Specification and Implementation Remediation Tracker

## 1. Control

- **Controlling artifact:** this file.
- **Execution base:** `d45f3fbf`.
- **Active branch at execution start:** `main`.
- **Program status:** `IN PROGRESS`.
- **Active workstream:** `RS-06`.
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
| RS-06 | Aggregate complete recovery-state catalog in shadow mode. Depends on RS-05. | Catalog fixtures, owner slices, boundary and generation drift. | All 109 tables and object families classify exactly once; 82 tables are required. | TODO | Pending |
| RS-07 | Add streaming backup storage and envelope v2 beside v1. Depends on RS-06. | Codec/security/fuzz-style cases, bounded-memory test, targeted gosec. | Corruption/reordering/truncation/duplication/trailing data fail closed; memory is bounded. | TODO | Pending |
| RS-08 | Add non-default repeatable-read vNext capture and restore codecs. Depends on RS-07. | Recovery service slice and vNext end-to-end fixtures. | vNext works in parallel without changing the active writer. | TODO | Pending |
| RS-09 | Atomically cut writer/reader selection to vNext and retain strict historical readers. Depends on RS-08. | New/historical restore, concurrent capture, owner rebuild/invalidation tests. | New backups are coherent vNext; retained historical backups restore. | TODO | Pending |
| RS-10 | Add Timeline registration, Workbook registry/executor, and single Recovery selection. Depends on RS-09. | Recovery, Timeline, Workbook owner slices and browser restore. | Exact query behavior passes without Recovery importing Timeline implementation. | TODO | Pending |
| RS-11 | Give each due backup an independent attempt lifecycle. Depends on RS-10. | Recovery due-batch unit/service tests. | All safe selected backups are attempted with complete per-item evidence. | TODO | Pending |
| RS-12 | Privatize leftovers and update boundary/verification ownership and generated topology. Depends on RS-11. | Owner explanations/slices, boundary, harness, generation and artifact drift. | Exact-once ownership, valid collaborators, and no obsolete shim remain. | TODO | Pending |
| RS-99 | Run full validation and complete handoff. Depends on RS-12. | Commands in section 6. | All required checks pass or have owner-approved exception; no TODO/BLOCKED remains. | TODO | Pending |

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

## 9. Final Handoff Checklist

- [ ] Every RS-00--RS-12 row is `DONE` with an exact substantive commit and
  following tracker ledger commit.
- [ ] RS-99 validation is complete.
- [ ] No TODO or BLOCKED item remains.
- [ ] All changed files and retained historical decoders are listed.
- [ ] Every command result and run root is recorded.
- [ ] Skipped `migration-drift` is justified, or its passing run is recorded.
- [ ] `domain.md` no-change decision remains valid.
- [ ] The worktree state and final commit range are recorded.
- [ ] Program status is `COMPLETE`.
