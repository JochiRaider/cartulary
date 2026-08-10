# db-migrations Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `db/migrations`
- **Target label:** `db-migrations` (derived from the path and normalized to safe lowercase kebab case)
- **Output path:** `docs/handoffs/db-migrations-refactor-tracker.md`
- **Status:** Authority adoption and implementation handoff only. No product implementation is performed by this checkpoint.
- **Allowed changes in this session:** Core 00, Core 01, Core 04, and this tracker only.
- **Non-goals:** No production code, tests, contracts, generated artifacts, package configuration, SQL migrations, queries, manifests, harness files, route behavior, storage behavior, or authorization behavior may change.
- **Implementation authorization:** The adopted owner text authorizes the target behavior defined below. Production code, characterization tests, machine contracts, generated projections, and harness changes remain `TODO` work for a later implementation task.

`MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, and `MAY` are normative within this implementation handoff only where they restate the adopted owners. This tracker MUST NOT create behavior independently of those owners. If this tracker and an adopted owner differ, the owner controls and the mismatch blocks implementation until this tracker is corrected.

The primary architectural finding is established by Core 01 §2.1A: `db/migrations` is schema-evolution and database-contract infrastructure, not a domain module. The directory contains one centralized, ordered, append-only authored SQL catalog and a small persistence-adjacent Go embedding adapter. The physical catalog location does not transfer schema meaning from the source owners named in `tools/schema_object_ownership_manifest.json` to a putative `db-migrations` domain module.

The planning source hierarchy is:

1. Adopted subsystem NLSpecs within their named subsystem.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication; it is not applicable to this structural plan.
4. `docs/domain.md` and implementation-support guides for vocabulary, boundaries, and harness mechanics.
5. Live repository code and tests for current implementation state.
6. This tracker, the planning framework, and prior handoffs as evidence only.

Owner and planning documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` in full before repository claims were made;
- `docs/spec/00_document_set_status_and_precedence.md`, especially status, precedence, and §5.1;
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially §§1, 2, and 2.1A;
- `docs/spec/02_domain_model_schema_and_history.md`, especially §§14.1 and 15;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, especially saved-view and collaboration sections;
- `docs/spec/04_security_deployment_and_conformance.md`, especially AC-537;
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` for scope separation only;
- adopted/current `docs/extension-subsystem-nlspec.md`, `docs/graph_projection_nlspec.md`, `docs/network-flow-activity-nlspec.md`, `docs/reporting-subsystem-nlspec.md`, `docs/report-composition-nlspec.md`, and `docs/opentelemetry-instrumentation-nlspec.md` for their named scopes;
- draft `docs/reference-pack-subsystem-nlspec.md` as evidence only, not current authority;
- `docs/testing-harness-nlspec.md`, `docs/domain.md`, and `docs/research/nlspec-spec.md` for execution, vocabulary, and specification discipline.
- `temp/analysis-notes.md` as decision-support evidence only; its recommendations were checked against the live repository before adoption.

Repository files inspected include every target file inventoried in §2, plus:

- `internal/modules/database_migrations/**` and `internal/modules/database_migrations/migrationevidence/**`;
- source callers in `internal/app/migrate`, `internal/app/operator`, `internal/app/server`, `internal/testutil/pgschema`, `internal/testutil/pgtest`, `tools/testservices`, and `tools/recoverybrowserrestore`;
- dedicated migration tests under the applicable source-owner packages and `internal/platform/jobs`;
- `sqlc.yaml`, `tools/migration_history_manifest.json`, `tools/schema_object_ownership_manifest.json`, `tools/generated_artifact_policy.json`, `contracts/recovery/fixtures/recovery-state-catalog.v1.json`, `contracts/verification/owners/module.database_migrations.json`, `tools/test_catalog_owner.json`, `tools/test_families/module.database_migrations.json`, and `tools/backend_module_boundaries.json`;
- public Make help, owner routing, and target explanations described in §§8 and 10.

No owner-owner contradiction was found in the inspected scope. This authority checkpoint resolves the former owner/projection ambiguity and the evidence-version decision. The live implementation remains non-conforming at the raw-source and evidence-v2 boundaries until the `TODO` slices below pass.

### Authority and closure map

| Decision | Adopted owner | Adopted target | Current repository evidence | Authority status | Implementation status |
| --- | --- | --- | --- | --- | --- |
| RB-001 recovery cardinality | Core 01 REQ-01-647 and REQ-01-575 | Exactly 111 catalog entries; exactly 83 `authoritative_required`; one Revisions-owned authoritative `record_revision_conflict_facts` entry | RC and `recoverystate.AuthoredTableCount` already report 111; RC reports 83 authoritative entries and the required unique Revisions entry | CLOSED | Projection already matches; no SQL or schema change is required |
| RB-002 production/test source split | Core 01 REQ-01-657; Testing Harness TH-HARNESS-REQ-810 and TH-HARNESS-AC-094 | One opaque cached production source and one sealed harness-only targeted capability over the same canonical embedded catalog | `Source()` exists, but raw `Files`/`EmbeddedPath`, caller-selected construction, `pgschema`, and `pgtest` still escape the target boundary | CLOSED | TODO: S-02 through S-05 |
| RB-003 evidence path disclosure | Core 01 REQ-01-657; Core 04 AC-537 | Current evidence is v2; `manifest.path` is absent without replacement; v1 is not current input or output | Current producer emits v1 and serializes `manifest.path` | CLOSED | TODO: S-07; current implementation mismatch is explicit |

The current-state and target-state distinction is normative: a closed authority decision does not mean the implementation is complete. Work items MUST use `DONE` only after their executable completion criteria pass.

## 2. Current-State Repository Inventory

Table abbreviations: `MH` is `tools/migration_history_manifest.json`; `SO` is `tools/schema_object_ownership_manifest.json`; `RC` is the recovery-state catalog and its generated projection; `SQLC` is `internal/gen/sql/**`. “Catalog-wide” test evidence means the source validation, schema-bootstrap integration, and/or `make migration-drift` path consumes the ordered catalog, not that every migration has a dedicated boundary-upgrade fixture.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `db/migrations/.gitkeep` | Empty directory sentinel retained after the directory gained content. | None. | None found. | None. | None. | Ignored sentinel in generated-artifact policy. | Repository maintenance. | low | Inventory only; `intentional/no_action` for this plan. |
| `db/migrations/source.go` | Embeds SQL and caches construction of the canonical migration source. | `Files`, `EmbeddedPath`, `LineageID`, `LineageBoundary`, `Source()`. | Server startup, migrate CLI, operator migration evidence, database-migration tests, `pgschema`, `pgtest`, recovery browser tool, testservices. | `embed`, `sync`, `internal/modules/database_migrations`. | Database-migration contract/readiness/remediation tests; operator evidence; pgtest. | Migration source identity, evidence, recovery metadata, build inputs. | `module.database_migrations` through a thin persistence-adjacent adapter. | high | `Source()` is the legitimate facade candidate; raw FS and metadata exports are boundary leaks. |
| `db/migrations/00001_database_infrastructure.sql` | Installs `pgcrypto`/`citext` and creates the production lineage table. | `schema_migration_lineage`; extensions. | Embedded source, SQLC, migration drift, pgtest. | PostgreSQL extensions. | Bootstrap guard; catalog-wide. | MH, SO, RC, SQLC schema input. | `database_migrations` plus `deployment_admin`. | high | Mixed historical migration; lineage and byte identity are frozen. |
| `db/migrations/00002_auth_accounts_and_enterprise.sql` | Creates local accounts, sessions, bootstrap/TOTP, idempotency, preferences, and enterprise-auth state. | Nine auth/account relations and their constraints/indexes. | Common catalog consumers. | Self-references among users, sessions, tokens, and auth bindings. | Catalog-wide; auth integration at head. | MH, SO, RC, SQLC. | `auth`. | high | Physical schema realizes Core auth contracts; not migration-owner semantics. |
| `db/migrations/00003_deployment_admin_and_recovery.sql` | Creates backup, bootstrap, administrative-audit, restore-verification, and operator-journal relations. | Five deployment/recovery relations. | Common catalog consumers. | `users`, `backup_sets`. | Catalog-wide; recovery/audit head-state tests. | MH, SO, RC, SQLC. | `deployment_admin` plus `recovery`. | high | Mixed historical migration; operator journal remains Recovery-owned. |
| `db/migrations/00004_incidents_and_preferences.sql` | Creates incidents, memberships, and incident/user workbook preference anchors. | Four incident relations. | Common catalog consumers. | `users`, `incidents`, deployment audit. | Catalog-wide; incident integration at head. | MH, SO, RC, SQLC. | `incidents`. | high | Authorization and startup-preference persistence are indirectly affected. |
| `db/migrations/00005_platform_jobs.sql` | Creates the shared background-job table. | `jobs` plus indexes/checks. | Common catalog consumers. | `incidents`, `users`. | Catalog-wide; platform jobs tests at head. | MH, SO, RC, SQLC. | `platform_jobs`. | high | Public job routes remain `jobapi`-owned. |
| `db/migrations/00006_record_revision_substrate.sql` | Creates current record envelopes and retained revision/history substrate. | `records`, `change_sets`, `change_set_mutations`, `record_revisions`, `record_history_entry_refs`. | Common catalog consumers. | `incidents`, `users`, record/revision self-relations. | Catalog-wide; Revisions integration at head. | MH, SO, RC, SQLC. | `records` plus `revisions`. | high | Mixed historical migration; current envelope and retained history are distinct owners. |
| `db/migrations/00007_timeline_source.sql` | Creates Timeline source rows and time-conversion profiles. | `timeline_events`, `timeline_time_conversion_profiles`. | Common catalog consumers. | `records`, `incidents`, `users`. | Catalog-wide; Timeline source/workbook tests at head. | MH, SO, RC, SQLC. | `timeline`. | high | Source state, not projection state. |
| `db/migrations/00008_entity_source.sql` | Creates aliases, mentions, preserved identifiers, hosts, and identities. | Five entity relations. | Common catalog consumers. | `records`, `incidents`, `users`, entity self-relations. | Catalog-wide; entity integration at head. | MH, SO, RC, SQLC. | `entities`. | high | Alias upgrade behavior is extended by migration 31. |
| `db/migrations/00009_indicator_source.sql` | Creates indicators, observations, and state intervals. | Three indicator relations. | Common catalog consumers. | `records`, `incidents`, `users`, indicators. | Catalog-wide; indicator integration at head. | MH, SO, RC, SQLC. | `indicators`. | high | Revision and resolution effects remain collaborators, not owners. |
| `db/migrations/00010_assessment_source.sql` | Creates compromise-assessment source state. | `assessments`. | Common catalog consumers. | `records`, `incidents`, `users`. | Catalog-wide; assessment integration at head. | MH, SO, RC, SQLC. | `assessments`. | high | Projection read models remain Projections-owned. |
| `db/migrations/00011_links_and_tags.sql` | Creates typed record links and tags. | `record_links`, `record_tags`. | Common catalog consumers. | `records`, `incidents`, `users`. | Catalog-wide; links/tags tests at head. | MH, SO, RC, SQLC. | `links`. | high | Generic relationship storage, not migration orchestration. |
| `db/migrations/00012_parties.sql` | Creates incident-scoped party source state. | `parties`. | Common catalog consumers. | `records`, `incidents`. | Catalog-wide; party integration at head. | MH, SO, RC, SQLC. | `parties`. | high | Party is distinct from user/identity. |
| `db/migrations/00013_tasks_and_decisions.sql` | Creates task-request and decision source state. | `task_requests`, `decisions`. | Common catalog consumers. | `records`, `incidents`, `users`, parties. | Catalog-wide; tasks/decisions tests at head. | MH, SO, RC, SQLC. | `tasksdecisions`. | high | Workflow behavior remains source-owner/Core-owned. |
| `db/migrations/00014_artifacts_and_optional_surfaces.sql` | Creates artifacts, findings, investigative queries, forensic keywords, handoff risks, and confidence helper. | Five relations and `cartulary_confidence_band`. | Common catalog consumers. | `records`, `incidents`, `users`, artifacts. | Catalog-wide; artifact contract/integration at head. | MH, SO, RC, SQLC. | `artifacts`. | high | Optional surface schema does not make this a workbook owner. |
| `db/migrations/00015_evidence_and_object_blobs.sql` | Creates evidence, blob, access-handle, and custody state. | Four evidence/blob relations. | Common catalog consumers. | `records`, `incidents`, `users`, evidence/blob self-relations. | Catalog-wide; evidence integration at head. | MH, SO, RC, SQLC. | `evidence`. | high | Indirectly affects object-store and handle contracts. |
| `db/migrations/00016_workbook_projections.sql` | Creates ten derived workbook grid projections. | Timeline, entity, indicator, assessment, task, decision, artifact, evidence, and party projection relations. | Common catalog consumers; SQLC query generation. | Source-owner tables, `incidents`, `users`. | Catalog-wide; projection/query/rebuild tests at head. | MH, SO, RC, SQLC. | `projections`. | high | Derived state; not authoritative source or history. |
| `db/migrations/00017_saved_views.sql` | Creates saved-view persistence. | `saved_views`. | Common catalog consumers. | `incidents`, `users`. | Catalog-wide; saved-view tests at head. | MH, SO, RC, SQLC. | `savedviews`. | high | Core 01/Core 03 own public and interaction contracts. |
| `db/migrations/00018_imports.sql` | Creates import sessions, units, and apply journal. | Three Import relations. | Common catalog consumers. | `incidents`, `users`, `jobs`, `records`, `change_sets`. | Catalog-wide; Imports integration at head. | MH, SO, RC, SQLC. | `imports`. | high | File ingest, provenance, and apply semantics remain Imports-owned. |
| `db/migrations/00019_reference_data.sql` | Creates reference-pack metadata, activation, attestations, and job payloads. | Four reference-data relations. | Common catalog consumers. | `jobs`, `users`, `reference_packs`. | Catalog-wide; Reference Data integration at head. | MH, SO, RC, SQLC. | `reference_data`. | high | Draft Reference Pack NLSpec is evidence only; current Core owners govern. |
| `db/migrations/00020_graph_projection.sql` | Creates graph-projection view, run, vertex, edge, and idempotency state. | Five graph-projection relations. | Common catalog consumers. | Graph projection self-relations. | Catalog-wide; Graph Projection storage tests at head. | MH, SO, RC, SQLC. | `graphprojection`. | high | Derived graph state; adopted Graph Projection NLSpec applies. |
| `db/migrations/00021_reporting.sql` | Creates snapshot, release, approval, render-bundle, and reporting-job state. | Six Reporting relations. | Common catalog consumers. | `incidents`, `users`, `jobs`, reporting self-relations. | Catalog-wide; Reporting tests at head. | MH, SO, RC, SQLC. | `reporting`. | high | Extension-profile schema only. |
| `db/migrations/00022_report_composition.sql` | Creates report-composition resources, versions, preview attempts, and release bindings. | Four Report Composition relations. | Common catalog consumers. | `incidents`, `users`, `jobs`, reporting releases, composition self-relations. | Catalog-wide; Report Composition tests at head. | MH, SO, RC, SQLC. | `reportcomposition`. | high | Authoring boundary remains distinct from Reporting render effects. |
| `db/migrations/00023_incident_bundles.sql` | Creates Incident Portability export/import metadata and file/job state. | Five incident-bundle relations. | Common catalog consumers. | `incidents`, `users`, `jobs`, bundle self-relations. | Catalog-wide; Incident Bundle tests at head. | MH, SO, RC, SQLC. | `incidentbundles`. | high | Last version inside the manifest's protected-through boundary. |
| `db/migrations/00024_enterprise_auth_provider_manifest.sql` | Adds manifest/configuration identity to enterprise-auth persistence. | Alters provider and transaction relations. | Common catalog consumers. | Existing auth tables. | Catalog-wide; enterprise-auth tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `auth`. | high | Preserve provider and transaction semantics. |
| `db/migrations/00025_links_projection_read_contracts.sql` | Adds active-link and active-tag read views. | `active_record_links_v1`, `active_record_tags_v1`. | Common catalog consumers; Links stores. | `record_links`, `record_tags`. | Catalog-wide; `links_tags_test.go`. | MH, SO, SQLC. | `links`. | high | Read contract belongs to Links, not generic Projections. |
| `db/migrations/00026_indicator_child_rollback_tombstones.sql` | Adds tombstone/version fields and active-only indexes for indicator child rows. | Alters observations and intervals. | Common catalog consumers. | Indicator relations and `users`. | Catalog-wide; indicator rollback/child route tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `indicators`, with Revisions collaboration. | high | Mixed historical migration; rollback semantics do not transfer source ownership. |
| `db/migrations/00027_platform_job_handlers.sql` | Adds handler lease/recovery fields to jobs. | Alters `jobs`. | Common catalog consumers. | Existing `jobs`. | Catalog-wide; platform jobs tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `platform_jobs`. | high | Shared job lifecycle substrate. |
| `db/migrations/00028_import_source_streams_and_targets.sql` | Adds import targets and authoritative source-stream storage. | Alters import units; creates `import_source_streams`. | Common catalog consumers. | `import_sessions`, `import_units`. | Catalog-wide; Imports integration at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `imports`. | high | Object inventory participates in recovery. |
| `db/migrations/00029_network_flow_storage.sql` | Creates Network Flow table, immutable row, and rejected-diagnostic storage. | Three relations, immutable-update function, and triggers. | Common catalog consumers. | Import, incident, user, and Network Flow relations. | Catalog-wide; Network Flow storage/import tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `networkflow`. | high | Adopted Network Flow Activity NLSpec applies when claimed. |
| `db/migrations/00030_network_flow_indicator_bindings.sql` | Creates explicit Network Flow-to-Indicator binding provenance. | `network_flow_indicator_bindings`. | Common catalog consumers. | `incidents`, `indicators`, `users`. | Catalog-wide; Network Flow/Indicator tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `networkflow`, with Indicators as target owner. | high | Binding persistence does not transfer Indicator identity ownership. |
| `db/migrations/00031_entity_alias_contract.sql` | Converts normalized aliases to `citext`, validates legacy rows, and tombstones case-equivalent duplicates. | Alters `entity_aliases`. | Common catalog consumers; targeted pgtest upgrades. | Existing entity aliases and PostgreSQL `citext`. | Dedicated `entity_alias_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `entities`. | high | Dedicated fresh, dedupe, and rejection characterization exists. |
| `db/migrations/00032_graph_projection_conformance.sql` | Resets/adapts legacy derived graph state and adds conformance fields/constraints. | Alters graph views, runs, and idempotency. | Common catalog consumers; targeted pgtest upgrades. | Graph state and Reporting references. | Dedicated `graph_projection_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `graphprojection`. | high | Blocks unsafe reset when Reporting references exist. |
| `db/migrations/00033_network_flow_keyset_indexes.sql` | Replaces Network Flow ordering indexes with keyset-pagination indexes. | Concurrent index operations. | Common catalog consumers. | Network Flow row/diagnostic relations. | Catalog-wide; query tests at head; no dedicated upgrade fixture found. | MH, SO, SQLC schema input. | `networkflow`. | high | Uses `NO TRANSACTION`; marker/order behavior is contract-sensitive. |
| `db/migrations/00034_extension_coordination.sql` | Creates generic extension metadata, migration ledger, staged-object/proof state, and adds extension job metadata. | Six extension relations and altered `jobs`. | Common catalog consumers; targeted pgtest upgrades. | Extension state, staged objects, `jobs`; known Network Flow lineage seed. | Dedicated `extension_job_cutover_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `extensions` plus `platform_jobs`. | high | Mixed historical migration; forward-only Down marker and preflight are observable. |
| `db/migrations/00035_import_unit_discovery_order.sql` | Adds deterministic import-unit discovery order. | Alters/backfills `import_units`. | Common catalog consumers. | Existing import units. | Catalog-wide; Imports integration at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `imports`. | high | Preserve backfill ordering. |
| `db/migrations/00036_reporting_composition_preview_outputs.sql` | Adds internal-draft composition preview output/file metadata. | Two Reporting relations. | Common catalog consumers. | `jobs`, composition preview attempts. | Catalog-wide; Reporting preview tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `reporting`. | high | Reporting owns outputs; Report Composition supplies authoring inputs. |
| `db/migrations/00037_incident_bundle_storage_references.sql` | Replaces physical bundle paths with logical storage references. | Alters bundle exports and job payloads. | Common catalog consumers; targeted pgtest upgrades. | Existing incident-bundle tables. | Dedicated `incident_bundle_storage_reference_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `incidentbundles`. | high | Clean cutover rejects unsafe populated legacy storage paths. |
| `db/migrations/00038_reference_pack_storage_references.sql` | Replaces physical reference-pack paths with logical storage references. | Alters pack and job-payload tables. | Common catalog consumers; targeted pgtest upgrades. | Existing reference-data tables. | Dedicated `reference_pack_storage_reference_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `reference_data`. | high | Current authority is Core, not the draft subsystem NLSpec. |
| `db/migrations/00039_administrative_audit_projections.sql` | Creates immutable safe administrative-audit projection and source trigger protection. | Audit projection relation, rejection function, triggers. | Common catalog consumers; targeted pgtest upgrades. | `deployment_admin_audit_events`. | Administrative-audit integration, including migration 39; catalog-wide. | MH, SO, RC, SQLC. | `audit`, with `deployment_admin` source collaboration. | high | Mixed historical migration; raw journal and projection have distinct roles. |
| `db/migrations/00040_remove_legacy_administrative_audit_projections.sql` | Removes legacy projected rows/shape while preserving current immutable audit state. | Alters audit projection and trigger. | Common catalog consumers; targeted pgtest upgrades. | Existing administrative-audit projection. | Administrative-audit cleanup integration; catalog-wide. | MH, SO, RC, SQLC. | `audit`. | high | Paired with migration 39 in upgrade characterization. |
| `db/migrations/00041_collaboration_durable_stream.sql` | Creates durable collaboration sequencing/replay/resume state and initially installs source-table event producers. | Four relations, producer functions, and triggers. | Common catalog consumers. | `incidents`, `user_sessions`, records/revisions, jobs, Network Flow. | Catalog-wide; Collaboration stream and producer-consumer tests at head. | MH, SO, RC, SQLC. | `collaboration`, with source-owner producers. | high | Mixed historical migration; migration 44 later removes these triggers/functions at current head. |
| `db/migrations/00042_timeline_projection_collections_expand.sql` | Expand phase for Timeline projection collection columns. | Alters `timeline_grid_projection`. | Common catalog consumers. | Timeline projection. | Catalog-wide; projection/workbook collection tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `projections`. | high | Expand/contract pair is migration sequencing, not runtime phase architecture. |
| `db/migrations/00043_timeline_projection_collections_contract.sql` | Backfills and constrains Timeline projection collections. | Alters `timeline_grid_projection`. | Common catalog consumers. | Timeline projection. | Catalog-wide; projection/workbook collection tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `projections`. | high | Empty-array defaults and JSON-array checks affect view rows. |
| `db/migrations/00044_collaboration_owner_producers.sql` | Removes database triggers/functions that produced collaboration events from peer source tables. | Drops producer triggers/functions; Down recreates legacy producers. | Common catalog consumers. | Revisions, jobs, Network Flow, collaboration intents. | Catalog-wide; application-owned Collaboration producer tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `collaboration`, with peer source owners. | high | Mixed historical migration; current head intentionally has no listed producer triggers. |
| `db/migrations/00045_timeline_source_provenance_expand.sql` | Adds normalized Timeline source-provenance relation. | `timeline_source_provenance`. | Common catalog consumers. | `timeline_events`. | Catalog-wide; Imports, portability, and Timeline tests at head. | MH, SO, RC, SQLC. | `timeline`. | high | Source provenance, not Import ownership. |
| `db/migrations/00046_timeline_source_provenance_contract.sql` | Contracts legacy Timeline provenance into normalized required state. | Alters/backfills `timeline_events`. | Common catalog consumers. | Timeline events and provenance. | Catalog-wide; Timeline/import/portability tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `timeline`. | high | Expand/contract sequence must remain ordered. |
| `db/migrations/00047_collaboration_stream_operations_expand.sql` | Adds quarantine, operation, and retry fields for Collaboration stream operations. | Alters cursor, intent, and replay relations. | Common catalog consumers. | Collaboration durable stream. | Catalog-wide; stream/requeue tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `collaboration`. | high | Authorization remains outside SQL. |
| `db/migrations/00048_collaboration_stream_operations_contract.sql` | Backfills/constrains Collaboration operation state and replaces dispatch indexes. | Alters `collaboration_event_intents`; indexes. | Common catalog consumers. | Collaboration durable stream. | Catalog-wide; stream/requeue tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `collaboration`. | high | Expand/contract sequence must preserve deterministic operations. |
| `db/migrations/00049_import_unit_apply_outcomes.sql` | Adds durable Import apply plans/outcomes and unit finalization links. | Alters units; creates two relations. | Common catalog consumers. | Imports, jobs, users, change sets. | Catalog-wide; `imports_integration_test.go` at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `imports`. | high | Durable outcome, idempotency, and finalization semantics are observable. |
| `db/migrations/00050_import_owner_error_outcomes.sql` | Tightens Import owner-error/action outcome vocabulary and backfills rows. | Alters `import_unit_apply_outcomes`. | Common catalog consumers. | Import outcomes. | Catalog-wide; Imports integration at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `imports`. | high | Preflight avoids silently admitting retired `use_null`. |
| `db/migrations/00051_import_operator_regions.sql` | Adds validated operator-region selections to import units. | Alters `import_units`. | Common catalog consumers. | Import units. | Catalog-wide; Imports region tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `imports`. | high | UI assistant state is persisted only through Imports-owned semantics. |
| `db/migrations/00052_saved_views_storage_hardening.sql` | Validates/backfills saved views and adds owner/version/time constraints. | Alters `saved_views`. | Common catalog consumers; targeted pgtest upgrades. | Saved views. | Dedicated `saved_views_storage_hardening_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `savedviews`. | high | Preflight diagnostics intentionally report counts, not row identities. |
| `db/migrations/00053_evidence_blob_association_uniqueness.sql` | Enforces one Evidence association per object blob after safe preflight. | Replaces Evidence blob index with unique index. | Common catalog consumers; targeted pgtest upgrades. | Evidence and object blobs. | Dedicated `evidence_blob_uniqueness_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `evidence`. | high | Dedicated Down/Up and redacted-diagnostic coverage exists. |
| `db/migrations/00054_revisions_incident_bundle_sequence_repair.sql` | Adds Revisions functions for Incident Bundle sequence repair. | Two Revisions functions. | Common catalog consumers. | Revision/history and Incident Bundle import. | Catalog-wide; portability/revisions tests at head; no dedicated upgrade fixture found. | MH, SO, SQLC. | `revisions`, with `incidentbundles` collaboration. | high | Functions are Revisions-owned coordination, not bundle-owned history. |
| `db/migrations/00055_indicator_observation_origin_constraint.sql` | Adds origin consistency constraint for indicator observations. | Alters `indicator_observations`. | Common catalog consumers. | Indicator observations. | Catalog-wide; indicator tests at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `indicators`. | high | Source-observation provenance is observable. |
| `db/migrations/00056_indicator_active_identities_expand.sql` | Adds active identity registry, synchronization triggers, rebuild, and validation helpers. | Relation, functions, triggers; alters indicators/records. | Common catalog consumers. | `indicators`, `records`. | Catalog-wide; active-identity integration at head; no dedicated upgrade fixture found. | MH, SO, RC, SQLC. | `indicators`, with Records envelope collaboration. | high | Expand phase establishes derived identity consistency. |
| `db/migrations/00057_indicator_envelope_contract.sql` | Contracts Indicator envelope/support-reference invariants and synchronization behavior. | Functions, constraints, triggers; alters four Indicator relations. | Common catalog consumers; targeted pgtest upgrades. | Indicators, active identities, observations, intervals, users. | Dedicated `envelope_contract_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `indicators`. | high | Dedicated Down/Up characterization exists. |
| `db/migrations/00058_platform_jobs_definition_contract.sql` | Converts jobs to v2 job-kind/progress-unit definitions with guarded backfill. | Alters `jobs`. | Common catalog consumers; targeted pgtest upgrades. | Jobs and collaboration replay state. | Dedicated `jobs_definition_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `platform_jobs`. | high | Preflight and exact mapping are observable. |
| `db/migrations/00059_platform_jobs_execution_fencing.sql` | Replaces legacy handler-attempt fields with execution-fencing state. | Alters `jobs`. | Common catalog consumers; targeted pgtest upgrades. | Jobs. | Dedicated `jobs_execution_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `platform_jobs`. | high | Unsafe legacy identity fails before mutation. |
| `db/migrations/00060_platform_jobs_expiry_tombstones.sql` | Adds durable expiry tombstone and guarded downgrade behavior. | Alters `jobs`; replaces retention index. | Common catalog consumers; targeted pgtest upgrades. | Jobs. | Dedicated `jobs_expiry_migration_test.go`; platform cutover tests; catalog-wide. | MH, SO, RC, SQLC. | `platform_jobs`. | high | Down is blocked after compaction facts exist. |
| `db/migrations/00061_revisions_history_associations.sql` | Adds canonical history associations and retained conflict-fact state. | Validation function, altered mutations, GIN index, `record_revision_conflict_facts`. | Common catalog consumers; targeted pgtest upgrades. | `change_set_mutations`, `record_revisions`. | Dedicated `history_associations_migration_test.go`; catalog-wide. | MH, SO, RC, SQLC. | `revisions`. | high | Adds the 111th recovery-catalog entry; the live 111/83 projection now agrees with adopted Core 01. |

Every target file is inventoried. No file is excluded from discovery; `.gitkeep` is explicitly outside any future behavior refactor because it has no behavior.

## 3. Module Boundary Diagnosis

The target is a **persistence-adjacent adapter plus a mixed-semantic authored catalog**. It is not a legitimate permanent domain module, application facade, mutation coordinator, frontend controller, transport adapter, or grid integration layer. The catalog's physical centralization is intentional because the database has one ordered migration line; semantic ownership remains distributed.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Ordered append-only production SQL history | `db/migrations/*.sql` | Central migration catalog governed by source owners and `database_migrations` lifecycle policy | keep | Core 01 §2.1A; MH; migration drift | Do not move, rename, rewrite, squash, or infer domain ownership from filenames. |
| Schema meaning for source relations | Central SQL files | Owners in SO: Auth, Incidents, Records, Revisions, Timeline, Entities, Indicators, and others | keep | SO object patterns and Core owner matrix | The current physical catalog is not an accidental catch-all runtime package. |
| Cached canonical embedded source | `db/migrations/source.go` | `module.database_migrations` behind a necessary embedding adapter | split | Core 01 REQ-01-657; current `Source()` call graph | Retain a thin adapter if Go embedding locality requires it; remove unrelated public capability. |
| Raw embedded filesystem access | `Files`/`EmbeddedPath` in `source.go` | Opaque migration source; Testing Harness gets a separate narrow capability | move | Operator evidence, `pgschema`, `pgtest`, bootstrap guard call sites | Production and harness needs are currently conflated. |
| Generic source construction | `database_migrations.NewSource` | Owner-private `sourcecatalog` construction seam with one adapter caller | move | Core 01 prohibits caller-selected production sources; Go embedding remains package-relative | RB-002 is decision-complete; S-02 MUST implement the exact boundary below. |
| Lineage identity constants | `source.go` exports | Database Migrations owner | move | Core 01 §2.1A; remediation/readiness tests | Preserve exact values and report behavior. |
| SQLC schema input | `sqlc.yaml` points at `db/migrations` | Generated SQL adapter pipeline | keep | `sqlc.yaml`; generated artifact policy | Generated files are refreshed only through Make generators. |
| Cross-owner triggers/functions in historical migrations | Migrations 39, 41, 44, 54, 56, 57 | Named semantic owners and application composition | keep | Exact SQL and current head tests | Historical DDL is evidence; future ownership changes use new forward migrations. |
| Targeted apply/rollback | `internal/testutil/pgtest` using raw catalog FS | Testing Harness | move | Core 01 REQ-01-657; AC-537; pgtest implementation | Must remain unavailable to production callers. |
| HTTP, WebSocket transport, frontend shell, view-contract adapter, grid-vendor integration | Not present in target | Existing transport/web/package owners | defer | Direct target inspection and import search | Only indirect schema contract risks apply. |

Architectural classification:

- legitimate thin service facade: only `Source()` as an adapter candidate;
- persistence-adjacent adapter: yes;
- mixed-responsibility package: yes at the physical catalog level, intentionally, with owner mapping required;
- misplaced home for logic owned by other modules: raw source/evidence/test capabilities require redesign, but historical SQL does not move;
- accidental catch-all, view orchestration, mutation coordinator, transport adapter, frontend shell, or grid-vendor layer: no evidence found.

### Required production source interface

The supported production entry point MUST be exactly:

```go
func migrations.Source() (*database_migrations.Source, error)
```

`Source()` MUST take no arguments, bind lineage `cartulary.prod_ddl_rebaseline.v1` and boundary `prod_ddl_rebaseline_v1`, read only the compile-time embedded `db/migrations/*.sql` catalog, validate the complete catalog before returning, and cache both success and failure. Repeated successful calls MUST return the same immutable source pointer. It MUST NOT panic, access a database, discover the operating-system filesystem, consult environment variables or the current working directory, accept a caller-selected lineage, or expose filesystem, path, root, raw SQL, mutable internal-entry, or lineage accessors. `database_migrations.Source` MUST have no exported fields.

The owner-private construction and consumption mapping is:

| Surface | Exact role | Allowed callers | Required behavior | Prohibited behavior |
| --- | --- | --- | --- | --- |
| `database_migrations.BuildCanonicalEmbedded(fs.FS, string) (*Source, error)` | Unsupported construction seam used by the physical embed adapter | One non-test caller: `db/migrations/source.go`; owner tests MAY construct malformed fixtures | Fix lineage inside the owner, copy/freeze bytes, and apply all source validation before return | Caller-selected lineage, runtime filesystem fallback, or use as a general production constructor |
| `internal/modules/database_migrations/sourcecatalog` | Owner-private representation and provider bridge | Database Migrations implementation and `internal/testutil/pgtest` only | Keep fields private and provide invocation-local provider input | Import from any application, platform, source-owner, command, or other test package |
| `database_migrations.InspectSource(*Source) (SourceInspection, error)` | Safe source metadata for evidence | Database Migrations evidence code | Return defensive metadata copies containing version, filename, SHA-256, marker facts, and ordering facts | Return SQL bytes, `fs.FS`, roots, paths, or mutable source storage |
| `database_migrations.SchemaHash(*Source, string) (string, error)` | Canonical digest operation for `pgschema` | `internal/testutil/pgschema` | Preserve the existing runner-identity, sorted filename, separator, and byte hashing algorithm exactly; reject an empty runner identity | Expose source bytes or permit a caller-selected source other than `migrations.Source()` |
| `migrationevidence.Build(..., *database_migrations.Source)` | Evidence source input | Operator evidence assembly | Audit the opaque canonical source and preserve current finding/order semantics | Accept raw `fs.FS` or serialize a source locator |

Static enforcement MUST reject every additional non-test call to `BuildCanonicalEmbedded`, every import of `sourcecatalog` outside the two allowed owner/harness paths, every raw `embed.FS` export from `db/migrations`, every production import of the harness capability, and every reintroduced generic source constructor or physical-adapter lineage export.

### Required Testing Harness interface

The only version-targeted interface MUST be:

```go
type MigrationDatabase struct {
	// all fields unexported
}

func (h *Harness) MigrationDatabaseT(testing.TB) *MigrationDatabase
func (h *Harness) MigrationDatabaseThroughT(testing.TB, int64) *MigrationDatabase
func (m *MigrationDatabase) SQL() *sql.DB
func (m *MigrationDatabase) ApplyThrough(context.Context, int64) error
func (m *MigrationDatabase) RollbackThrough(context.Context, int64) error
```

The Harness MUST issue this sealed concrete type only for a disposable `postgres_migration` scratch lease. It MUST bind an unexported harness identity and the harness-owned database handle, obtain `migrations.Source()` internally, accept no source, filesystem, root, lineage, prefix, DSN, or arbitrary database, and assign its own collision-safe scratch identity. `SQL()` returns a borrowed assertion handle for the same scratch database. `ApplyThrough` MUST reject targets less than or equal to zero before source or database access. `RollbackThrough` MUST reject negative targets before access and MUST accept zero. An unissued or zero capability MUST reject operations as `migration database capability is not harness-issued`.

Each targeted operation MUST create one invocation-local PostgreSQL Goose Provider from the capability database, the canonical private source representation, the owner locker, disabled global Go registration, and discarded logging. Cleanup MUST close and destroy only the owned scratch database and MUST NOT close or delete borrowed state. The capability MUST NOT establish production rollback, restore, downgrade, or recovery behavior.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Migration filenames, versions, contiguous order, Up/Down markers, bytes, and hashes | Database Migrations lifecycle plus each schema owner | MH, `NewSource` validation, AC-537 | Source contract tests; migration drift | Freeze exact 1–61 inventory and hashes before any source-boundary edit | critical | Existing SQL is immutable for this refactor. |
| Production lineage ID and remediation boundary | Database Migrations | Core 01 §2.1A; `source.go`; migration 1 | Remediation/readiness matrices; migrate/server diagnostics | Preserve exact `cartulary.prod_ddl_rebaseline.v1` and `prod_ddl_rebaseline_v1` framing | critical | No historical-line bridge. |
| Apply-to-head locking, cancellation, postcondition, and borrowed DB lifetime | Database Migrations | `migrations.go`, lock/state code | Owner unit and integration rows | Existing evidence is sufficient if SQL and apply implementation are untouched | high | Target refactor must not create targeted production operations. |
| Server readiness ordering and failure mapping | `app.server` plus Database Migrations | Core 01 §2.1A; runtime assembly | Server and readiness tests | Preserve source acquisition and typed error mapping | high | Readiness must precede schema-dependent subsystem setup. |
| Migrate CLI grammar and output | `app.migrate` plus Database Migrations | `migrate.go` | `app.migrate` unit row | Preserve exact `migrate up`, exit codes, reason/report output | high | No generic Goose grammar. |
| Operator migration-evidence JSON | Database Migrations evidence plus `app.operator` transport | Adopted Core 01 REQ-01-657; AC-537; current `migrationevidence.Result`; operator tests | Operator unit/integration and owner rows | Freeze v1 for S-01 through S-06; S-07 MUST cut over atomically to the v2 mapping below | critical | Authority is adopted; current v1 plus `manifest.path` is an explicit implementation mismatch. |
| Test template schema hash | Testing Harness | `internal/testutil/pgschema` | pgtest/testservices coverage | Prove identical hash before and after catalog-capability cutover | high | Hash identity drives reusable templates. |
| Harness-only targeted apply/rollback | Testing Harness | `internal/testutil/pgtest` | Migration capability validation and dedicated owner upgrade tests | Preserve capability issuance, positive/nonnegative targets, and Down/Up fixtures | critical | Must not escape to production API. |
| HTTP route and envelope behavior backed by migrated schema | Named route/source owners | Core 01/02/04; SO downstream surfaces | Owner route/integration suites | No new cross-owner characterization unless schema bytes change, which is forbidden | medium | No route handler exists in target. |
| WebSocket sequencing, replay, resume, and event semantics | Collaboration | Migrations 41, 44, 47, 48; Core 03 | Collaboration stream/requeue tests | Preserve current-head absence of retired producer triggers and application producer behavior | high | SQL history affects durable stream state indirectly. |
| Entity row/query/mutation and alias behavior | Entities | Migrations 8 and 31 | Dedicated migration 31 and entity suites | Existing dedicated upgrade coverage is sufficient if bytes remain unchanged | high | Includes preflight diagnostic behavior. |
| Indicator observation, identity, support, and rollback behavior | Indicators with Revisions | Migrations 9, 26, 55–57 | Indicator suites; dedicated migration 57 | Characterize versions 26, 55, and 56 before any later schema-semantic change | high | No such change is authorized here. |
| Saved-view and view-schema behavior | Saved Views, Core 01/Core 03, Projections | Migrations 16, 17, 42, 43, 52 | Saved-view migration 52; projection/workbook suites | Preserve empty-array/default/owner/version constraints and view row shape | high | Physical schema is non-normative but observable behavior is fixed. |
| Projection refresh/query behavior | Projections and Graph Projection | Migrations 16, 20, 32, 42, 43 | Projection suites; graph migration 32 | Existing head-state coverage plus dedicated graph upgrade coverage | high | Derived tables remain disposable only under owner rules. |
| Import session/unit/apply behavior | Imports | Migrations 18, 28, 35, 49–51 | Imports integration | Add upgrade-specific characterization only if a future slice changes these migrations | high | Current refactor must not touch them. |
| Evidence/blob storage and access behavior | Evidence | Migrations 15 and 53 | Evidence suites; dedicated migration 53 | Existing dedicated uniqueness upgrade coverage is sufficient | high | Object-store bytes remain outside these SQL relations. |
| Revision/history/conflict behavior | Revisions and Records | Migrations 6, 54, 61 | Revisions suites; dedicated migration 61 | Preserve history association, conflict-fact, and portability repair behavior | critical | Current record envelopes remain Records-owned. |
| Authorization persistence outcomes | Auth, Incidents, Deployment Admin | Migrations 2–4, 24, 39–40 | Auth, incident, audit, server suites | No target-specific addition unless semantic SQL changes | high | SQL must not become authorization logic owner. |
| Generated SQL models and queries | SQLC pipeline | `sqlc.yaml`, generated policy | Generate drift/check | Prove no generated diff for structural source-only slices | high | Never hand-edit generated roots. |
| Recovery catalog coverage | Recovery plus every source owner | Adopted REQ-01-647/575; RC and recovery contributions | Recovery catalog drift tests | Prove the existing 111/83 projection and unique Revisions contribution remain aligned | critical | No SQL, schema, migration 62, or recovery data change is required. |
| Harness/test accounting | Testing Harness | Verification owner and test-family manifests | Catalog checks and owner explanations | Update routing only if tests move/add; never infer architecture from rows | medium | Test rows are evidence accounting only. |
| Browser, UI selectors, grid adapter | Web/package owners | No target implementation found | Existing owner browser tests | Not applicable for source-boundary-only refactor | low | Required only if a later schema-semantic change affects a public surface. |

### Recovery classification contract

| Category | Required count | Owner-defined treatment | Acceptance rule |
| --- | --- | --- | --- |
| `authoritative_required` | 83 | Back up and restore through the exact owner codec | The admitted authoritative set contains 83 unique entries. |
| `excluded_rebuildable` graph projections | 5 | Rebuild through the Graph Projection owner algorithm | All five `graph_projection_*` entries are excluded from backup and rebuilt. |
| `excluded_security_state` | 5 | Invalidate across restore generation | Four `collaboration_*` streams plus `enterprise_auth_transactions` are invalidated. |
| Explicit exclusions | 7 | Apply each declared ignore/invalidate/rebuild action | Every exclusion has one owner and one action. |
| Synthetic migration metadata | 1 | Treat `goose_db_version` through the declared schema-metadata rule | It appears once and is not authoritative application state. |
| Workbook grid projections | 10 | Rebuild through the Workbook projection algorithm | Every `*_grid_projection` entry is excluded from authoritative backup. |
| Total | 111 | Aggregate exactly once | `83 + 5 + 5 + 7 + 1 + 10` MUST equal 111. |

| `record_revision_conflict_facts` field | Required value |
| --- | --- |
| Table identity | `record_revision_conflict_facts` |
| Source owner | `module.revisions` |
| State class | `authoritative` |
| Backup inclusion | `authoritative_required` |
| Restore action | `restore` exact retained rows |
| Codec | `cartulary.postgres_snapshot_unit.v1` |
| Rebuild algorithm | None |
| Invalidation algorithm | None |
| Incident Portability | Excluded under the existing non-portable conflict-fact rule |

Historical backups MUST use the catalog and decoder embedded in the retained backup. They MUST NOT be reinterpreted with the current 111/83 catalog. Migrations 1–61 MUST retain identical filenames, versions, ordering, bytes, hashes, and lineage. Migration 62 MUST NOT be created for this authority correction.

### Migration-history evidence v2 contract

Current implementation v1 is the frozen input to S-07. The v2 producer and schema MUST use the following exhaustive mapping. A row marked “unchanged” retains its current v1 type, presence, ordering, and semantics.

| JSON field | v1 | Required v2 type/presence | Normative v2 rule |
| --- | --- | --- | --- |
| `schema_id` | `cartulary.migration_history_evidence.v1` | required string | Exact value `cartulary.migration_history_evidence.v2`. |
| `collected_at` | present | required UTC timestamp | Preserve the caller-selected or owner-clock collection instant in current canonical JSON timestamp form. |
| `evidence_only` | `true` | required boolean | MUST be `true`. |
| `rewrite_authorized` | `false` | required boolean | MUST be `false`; evidence never authorizes migration rewrite. |
| `database_binding` | object | required closed object | Preserve binding normalization and privacy. |
| `database_binding.binding_kind` | string | required string | Trim surrounding whitespace; preserve the logical binding kind. |
| `database_binding.service_ref` | optional string | optional string | Omit when empty; preserve the normalized non-secret service reference otherwise. |
| `manifest` | object | required closed object | Preserve logical manifest identity and digest facts; reject unknown members. |
| `manifest.schema_id` | string | required string | Preserve the audited manifest schema identity. |
| `manifest.path` | required path string | prohibited | Remove without replacement; any occurrence MUST fail v2 validation. |
| `manifest.sha256` | lowercase SHA-256 | required string | Preserve the digest of the exact manifest bytes. |
| `manifest.migration_root` | logical root string | required string | Preserve as an opaque manifest contract identity; it MUST NOT be resolved, joined, or advertised as a filesystem locator. |
| `manifest.immutable_through_version` | integer | required integer | Preserve the protected migration boundary. |
| `manifest.expected_min_version` | integer | required integer | Preserve the manifest-derived minimum version. |
| `manifest.expected_max_version` | integer | required integer | Preserve the manifest-derived maximum version. |
| `manifest.expected_version_count` | integer | required integer | Preserve the exact manifest entry count. |
| `source_audit` | ordered array | required array | Preserve deterministic order by version ascending, then filename ascending. |
| `source_audit[].version` | integer | required positive integer | Preserve the parsed migration version. |
| `source_audit[].filename` | string | required string | Preserve the migration filename only, never a containing path. |
| `source_audit[].sha256` | lowercase SHA-256 | required string | Preserve the digest of the exact embedded migration bytes. |
| `source_audit[].has_goose_up` | boolean | required boolean | Preserve Up-marker detection. |
| `source_audit[].has_goose_down` | boolean | required boolean | Preserve Down-marker detection. |
| `source_audit[].phase_shaped_name` | boolean | required boolean | Preserve phase-shaped filename classification. |
| `source_audit[].immutability_class` | string | required enum | Exact values remain `protected` or `current`. |
| `source_audit[].manifest_filename` | optional string | optional string | Present only when the version exists in the manifest. |
| `source_audit[].manifest_sha256` | optional string | optional string | Present only when the version exists in the manifest. |
| `source_audit[].manifest_hash_matches` | boolean | required boolean | Preserve exact digest equality result. |
| `goose_ledger` | object | required closed object | Preserve ledger collection and missing-metadata behavior. |
| `goose_ledger.metadata_present` | boolean | required boolean | Preserve metadata-table presence. |
| `goose_ledger.row_count` | integer | required non-negative integer | Preserve the exact ledger row count; zero when metadata is absent. |
| `goose_ledger.current_effective_applied_version` | integer | required non-negative integer | Preserve the highest effective applied version; zero when none exists. |
| `goose_ledger.latest_effective_states` | array or `null` | required array or `null` | Preserve `null` when metadata is absent and version-ascending rows otherwise. |
| `goose_ledger.latest_effective_states[].version` | integer | required integer | Preserve the ledger version. |
| `goose_ledger.latest_effective_states[].is_applied` | boolean | required boolean | Preserve the latest effective application state. |
| `goose_ledger.latest_effective_states[].tstamp` | timestamp | required timestamp | Preserve the ledger timestamp. |
| `findings` | ordered array | required array | Preserve sorting by effective version, reason code, then filename. |
| `findings[].severity` | string | required enum | Exact values remain `blocking`, `warning`, or `info`. |
| `findings[].reason_code` | string | required string | Preserve the current closed reason-code behavior. |
| `findings[].version` | optional integer | optional positive integer | Omit when the finding has no positive version. |
| `findings[].filename` | optional string | optional string | Omit when no migration filename applies; never emit a path. |
| `findings[].detail` | string | required string | Preserve the safe logical detail and MUST NOT disclose a filesystem locator. |

The v2 finding set and triggers MUST remain:

| Reason code | Severity | Exact trigger class |
| --- | --- | --- |
| `manifest_schema_unsupported` | `blocking` | Manifest `schema_id` is not the supported migration-history manifest identity. |
| `manifest_duplicate_version` | `blocking` | More than one manifest entry declares the same version. |
| `manifest_version_gap` | `blocking` | Sorted manifest versions are not contiguous from 1. |
| `source_filename_invalid` | `blocking` | An embedded SQL filename is outside the exact five-digit lower-snake form. |
| `source_duplicate_version` | `blocking` | More than one embedded file declares the same parsed version. |
| `manifest_filename_mismatch` | `blocking` | Source and manifest filenames differ for one version. |
| `manifest_hash_mismatch` | `blocking` | Source and manifest SHA-256 values differ for one version. |
| `source_marker_missing` | `blocking` | A source file lacks either its Up or Down marker. |
| `future_phase_shaped_filename` | `warning` | A version beyond the immutable boundary has a phase-shaped filename. |
| `manifest_version_not_in_source` | `blocking` | A manifest version has no embedded source file. |
| `source_version_not_in_manifest` | `blocking` | An embedded source version has no manifest entry. |
| `source_version_gap` | `blocking` | Sorted source versions are not contiguous from 1. |
| `db_version_not_in_manifest` | `blocking` | The effective Goose ledger references a nonzero version absent from the manifest. |
| `db_applied_version_gap` | `blocking` | An effective applied version is missing or down at or below the current applied maximum. |
| `migration_metadata_missing` | `blocking` | `goose_db_version` is absent, so applied-version evidence is unavailable. |
| `protected_boundary_applied` | `info` | The current applied maximum reaches or exceeds the positive immutable boundary. |

The Operator command contract remains:

| Input or outcome | Exact v2 behavior |
| --- | --- |
| Command grammar | `operator migration-evidence capture [-source-config <path>] [-manifest <path>] [-as-of <RFC3339>]` |
| Omitted `-source-config` | Preserve the current deployment-config default resolution. |
| Omitted `-manifest` | Use `tools/migration_history_manifest.json`. |
| Empty `-manifest` | Emit the existing safe invocation diagnostic on stderr and exit `2`. |
| Omitted `-as-of` | Use the injected owner clock and serialize the UTC instant. |
| Invalid `-as-of` or flag grammar | Emit the existing safe invocation diagnostic on stderr and exit `2`. |
| Help | Preserve current help output and exit `0`. |
| Successful capture | Emit exactly one v2 JSON object followed by one LF on stdout, emit no success stderr, close the acquired Postgres pool, and exit `0`. |
| Runtime capture failure | Emit no partial evidence object, log only the existing safe failure, close acquired resources, and exit `1`. |

The future machine projection MUST be `contracts/database-migrations/migration-history-evidence.v2.schema.json`, MUST close every object against unknown members, MUST register through `tools/harness_schema_attachments.json`, and MUST include valid v2 and rejected-v1 fixtures. Current producers and validators MUST use v2 only. They MUST NOT add a v1 flag, legacy mode, dual output, wrapper containing both versions, runtime v1 translation, or replacement locator.

| Prohibited v2 member or value | Required result |
| --- | --- |
| `manifest.path` | Schema rejection. |
| `manifest.source_path`, `manifest.repository_path`, `manifest.embedded_path`, `manifest.file`, or `manifest.uri` | Schema rejection as unknown members. |
| `manifest.path_sha256` or another path-derived identity | Schema rejection as an unknown member. |
| Absolute manifest path, repository-relative manifest-file path, current working directory, or runtime embedded root in any field, diagnostic, log, telemetry, stderr, or stdout | Conformance failure. |
| Relocated repository/executable with identical logical inputs | Evidence MUST be identical apart from an intentionally supplied `collected_at`. |

S-07 MAY change only the top-level schema ID and removal of `manifest.path`. The logical command and grammar, authorization, database open/close behavior, stdout object plus trailing LF, exit-code mapping, safe stderr behavior, all other fields, nullability, array order, findings, and ledger semantics MUST remain unchanged.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `Files` and `EmbeddedPath` expose raw catalog filesystem access to production operator code and test infrastructure. | `source.go`; operator evidence; `pgschema`; `pgtest`; bootstrap guard | Production and harness capabilities are not separated; arbitrary filesystem-style use can recur. | `must_fix` | Database Migrations plus Testing Harness | Implement S-02 through S-04 using the exact §3 interfaces; RB-002 is closed. |
| `database_migrations.NewSource` accepts caller-selected FS/root/lineage and is exported to satisfy the embedding adapter. | `migrations.go`; only production caller is `source.go`; tests also construct synthetic sources | Conflicts with the narrow production surface required by REQ-01-657. | `must_fix` | Database Migrations | Replace it with the fixed-lineage single-caller construction seam; owner tests MAY retain private malformed-catalog constructors. |
| `LineageID` and `LineageBoundary` are exported from the physical adapter and imported by owner tests. | `source.go`; remediation/readiness tests | Physical placement can masquerade as semantic ownership. | `must_fix` | Database Migrations | Localize identity while preserving exact remediation wire behavior. |
| No guard prevents future raw migration catalog access. | Backend boundary rules cover owner imports but not `db/migrations` exported FS use. | Boundary can regress after refactor. | `should_fix` | Database Migrations/Testing Harness static policy | Add a source/API/import guard after the new capability is established. |
| Central SQL history contains many semantic owners. | All 61 migrations; SO maps 29 owner labels. | A directory-shaped refactor could incorrectly centralize schema meaning. | `intentional/no_action` | Named source owners plus Database Migrations lifecycle | Keep the ordered catalog and record owner per object/file; do not create a domain facade over all schema. |
| Existing migrations contain cross-owner FKs, functions, triggers, and expand/contract steps. | Migrations 1, 3, 6, 26, 34, 39, 41, 44, 54, 56, 57 | Rewriting history could change applied databases and observable failures. | `intentional/no_action` | Historical source owners | Use new forward migrations for future owner-authorized changes. |
| SQLC consumes the entire migration directory. | `sqlc.yaml` | Structural SQL edits can cause broad generated drift. | `intentional/no_action` | SQLC pipeline | Do not alter schema inputs; use Make generation only if a later authorized schema change requires it. |
| Operator evidence serializes `manifest.path`. | `ManifestSummary.Path`; operator payload; adopted REQ-01-657 and AC-537 | Current implementation emits rejected v1 evidence and a prohibited locator. | `must_fix` | Database Migrations evidence and Operator transport | Preserve v1 only through S-06, then perform the owner-authorized atomic S-07 v2 cutover. |
| Recovery projection contains 111 entries and 83 authoritative entries, including the unique Revisions conflict-fact entry. | Adopted REQ-01-647/575; migration 61; RC; `recoverystate.AuthoredTableCount` | Treating the projection as authority or changing SQL would recreate drift. | `intentional/no_action` | Core/Recovery/Revisions owners | Keep the projection downstream, prove its counts, and make no SQL/schema/catalog edit in this authority checkpoint. |
| Dedicated upgrade tests exist only for selected later migrations. | Test-family manifests and migration test search | Future changes to uncovered upgrade boundaries could lack characterization. | `defer` | Each source owner | No gap blocks a source-only refactor because SQL bytes stay frozen; add tests before any later semantic migration change. |
| Reference Pack subsystem document remains draft. | Front matter of `docs/reference-pack-subsystem-nlspec.md` | Treating it as adopted would invert authority. | `intentional/no_action` | Core 00–04 and Reference Data | Use Core owners for current behavior; retain the draft as evidence only. |
| No platform import, frontend shell, direct grid-vendor, UI-selector, or test-only production assumption was found inside the target SQL/adapter beyond raw catalog access. | Direct target inspection and import searches | Low. | `intentional/no_action` | Existing platform/web/package owners | Keep these concerns out of the target. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Authority adoption and tracker bootstrap | root | none | WF-01 | Adopt the 111/83, opaque-source, and evidence-v2 owner decisions before implementation. | Core 00, Core 01, Core 04, tracker | `make lint-markdown`; `git diff --check`; read-only count checks | Owner decisions are non-contradictory and RB-001 through RB-003 are closed at the authority layer. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Retain the exact 63-file inventory, callers, dependencies, owner maps, tests, and generated consumers. | `db/migrations/**`, caller and manifest files | Read-only inventory checks; future `make migration-drift` | Every target file has one current-state row and SQL diff is empty. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map lifecycle mechanics to Database Migrations, schema meaning to source owners, and Recovery contribution semantics to adopted Core. | Owner specs, SO, RC, verification manifests | Owner review and RC count query | No physical-path ownership inference or count ambiguity remains. |
| WF-03 | Characterization gap analysis | chain | WF-01, WF-02 | WF-05, WF-06 | Freeze v1 baseline behavior, source identity, readiness, schema hash, targeted harness behavior, and migration bytes before code movement. | Owner tests and test-family manifests | Focused unit/service-backed slices; migration drift | Every observable changed by S-02 through S-07 has a named baseline assertion. |
| WF-04 | Boundary and coupling enforcement design | parallel | WF-01 | WF-05, WF-07 | Map raw FS, generic constructor, lineage, codegen, evidence, and harness leaks to exact forbidden-import/call rules. | Source adapter, Database Migrations, operator, pgtest/pgschema, boundary manifest | `make backend-module-boundary-check` | Every prohibited surface has a negative fixture and an allowed caller set. |
| WF-05 | Opaque facade and ownership design | chain | WF-02, WF-03, WF-04 | WF-06 | Implement the exact §3 production, inspection, digest, evidence, and harness interfaces. | Adapter, Database Migrations, migration evidence, harness | Unit slice and static boundary | Only `migrations.Source()` is the supported production catalog entry and targeted operations remain harness-only. |
| WF-06 | Behavior-preserving caller cutover | chain | WF-05 | WF-07, WF-08 | Cut over production, evidence, `pgschema`, and `pgtest` callers before deleting obsolete exports. | Files named by S-02 through S-05 | Validation attached to every slice | Zero callers use raw FS, generic construction, or adapter-owned lineage exports. |
| WF-07 | Evidence-v2 and harness accounting | parallel | WF-03, WF-04, WF-05 | WF-08 | Add the closed v2 schema/fixtures and preserve harness source, hash, lease, event, and routing behavior. | Contracts, operator evidence, `internal/testutil`, authored harness inputs | JSON shape, owner slices, boundary, generated drift | Current validators accept only v2; harness capability and evidence routing are complete. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow-to-broad checks, prove SQL/generated integrity, and close executable work items. | Whole later implementation diff | Commands in §8 | All binary criteria pass with retained roots and no unexplained diff. |

## 7. Proposed Refactor Slice Plan

These are later implementation slices. This checkpoint adopts their owner contract but performs no product, test, contract, generated, or harness implementation. Each later task MUST land one slice atomically and MUST NOT use compatibility aliases to span slices.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | none | Freeze exact filenames, versions, order, bytes, hashes, lineage, source construction result, source-audit output, readiness outcomes, operator v1 bytes, schema hash, and targeted scratch behavior. | Database Migrations, operator, pgtest, and pgschema tests | An incomplete baseline could conceal structural or wire drift. | Preserve all current owner rows; add exact source/hash/v1 characterization only where absent. | `make test-slice OWNER=module.database_migrations` and `make service-backed-test-slice OWNER=module.database_migrations` | Revert characterization-only changes if they assert behavior not present at the baseline. | Every S-02 through S-07 observable has a passing named baseline and SQL diff is empty. |
| S-02 | S-01 | Add owner-private `sourcecatalog`, pointer-valued opaque `Source()`, fixed-lineage construction, defensive inspection, and digest operations while retaining SQL in place. | `db/migrations/source.go`, `internal/modules/database_migrations`, new owner-private bridge | Cache identity/failure, source validation, lineage, digest, and provider behavior. | Construction rejection matrix, mutation-after-build, pointer identity, digest parity, public-surface tests. | `make test-slice OWNER=module.database_migrations` | Revert S-02 as a unit; do not expose raw FS as a fallback. | Exact §3 production interfaces compile and pass; existing callers remain behaviorally unchanged. |
| S-03 | S-02 | Change migration evidence and Operator assembly to receive the opaque source; preserve v1 output byte-for-byte for this slice. | Migration evidence, `internal/app/operator`, source adapter | Audit ordering, hashes, findings, command grammar, resource cleanup, framing. | Evidence unit/integration and exact v1 fixture parity. | `make test-slice OWNER=module.database_migrations` and `make test-slice OWNER=app.operator` | Revert the caller cutover without changing the v1 contract. | Operator no longer imports raw catalog FS and emits the frozen v1 baseline. |
| S-04 | S-02 | Make `pgschema` use the owner digest and make `pgtest` use the exact sealed no-prefix capability over the private provider bridge. | `internal/testutil/pgschema`, `internal/testutil/pgtest`, source-owner migration tests | Template hash, scratch naming/ownership, target validation, rollback fixtures, preparation events. | Hash parity, unissued/invalid target tests, cleanup/borrowed-handle tests, every dedicated upgrade fixture. | `make service-backed-test-slice OWNER=module.database_migrations` | Revert the harness cutover as a unit; never add production targeted operations. | Hash is unchanged, all targeted tests use `MigrationDatabase`, and no caller supplies source/DB/DSN/prefix. |
| S-05 | S-03, S-04 | Remove `Files`, `EmbeddedPath`, adapter lineage constants, generic `NewSource`, raw scratch issuance, and obsolete signatures; add exact import/call guards. | Source adapter, Database Migrations, pgtest, boundary owner input/tests | Compile surface, remediation identity, and accidental test-to-production leaks. | Public-surface and positive/negative boundary fixtures. | `make backend-module-boundary-check` | Remove only after zero callers remain; rollback the whole removal/guard slice on failure. | Only approved §3 surfaces remain and every prohibited fixture is rejected. |
| S-06 | S-05 | Prove structural work caused no SQL, lineage, manifest, Recovery, SQLC, or generated drift. | No intended authored SQL or generated changes | Accidental catalog or projection mutation. | Existing migration/recovery/generated drift evidence. | `make migration-drift` and `make generated-artifact-policy-check` | Revert structural code; MUST NOT update SQL, hashes, or generated output to accommodate drift. | Migrations 1–61 and all protected identities are unchanged; Recovery remains 111/83. |
| S-07 | S-03, S-05, S-06 | Perform the owner-authorized atomic v2 cutover: add the closed v2 schema and fixtures, emit v2 only, remove `ManifestSummary.Path`, and reject current v1 evidence. | Core-projected contract root, migration evidence, Operator tests, schema attachment and verification inputs | Intentional breaking schema-shape change, path leakage, stale v1 acceptance, unrelated output drift. | Exhaustive §4 field matrix, v1 negative fixture, unknown-path rejection, relocation invariance, exact Operator framing/exit/resource tests. | `make json-shape-check`; unit and service-backed test slices for `module.database_migrations` and `app.operator` | Revert S-07 as one contract/producer/validator/test unit; no dual-version fallback is permitted. | Current output and validators use only v2, no locator is observable, and every unrelated v1 behavior is unchanged. |
| S-08 | S-06, S-07 | Run focused, static, generated, broader, and full checks; record roots, failures, retries, and final integrity. | Whole authorized implementation diff | Hidden cross-owner regression or incomplete evidence accounting. | Every affected owner and harness row. | `make test-fast`, then `make check` after all narrow gates | Return to the last passing slice; database rollback and SQL-history rewrite are forbidden. | Every §12 executable criterion passes with named evidence and no unexplained diff. |

No slice may edit migrations 1–61 or add migration 62 for this refactor. No slice may copy or generate migration SQL, create Go byte literals, add a runtime filesystem fallback, or hand-edit generated output. A future unrelated schema change requires its own owner-authorized forward migration task.

## 8. Validation Plan

Commands were discovered through public Make help and target/owner explanation. This authority checkpoint MUST run only Markdown, diff, inventory, and read-only owner/projection consistency checks. Product, service-backed, migration, generated, boundary, and JSON-schema commands below are mandatory for their later slices and MUST NOT be reported as run by this checkpoint.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.database_migrations` | Seven fast/full owner unit rows plus selected collaborators | yes | Establish source, error, evidence, public-surface, and harness-capability baseline. |
| integration | `make service-backed-test-slice OWNER=module.database_migrations` | Five PostgreSQL-backed owner rows | yes | Covers locking, readiness, bootstrap, remediation, and evidence semantics. |
| operator unit | `make test-slice OWNER=app.operator` | Operator command grammar, framing, resource, and evidence projection rows | yes | Freeze v1 before S-03 and prove v2 after S-07. |
| operator integration | `make service-backed-test-slice OWNER=app.operator` | PostgreSQL-backed Operator evidence semantics | yes | Required for source cutover and v2 completion. |
| e2e/browser | N/A for a source-boundary-only refactor | Browser surfaces | no | Required only if a later schema-semantic change affects an owner public surface. |
| generated drift | `make migration-drift`, `make generate-drift`, and `make generated-artifact-policy-check` | Migration inputs/scratch application, generated projections, and generated-file policy | yes | Do not update hashes or generated outputs to make an unintended diff pass. |
| evidence schema | `make json-shape-check` | Closed migration-evidence v2 schema, fixtures, and attachment | no | Required in S-07 after the machine contract exists; v1 MUST be a negative current fixture. |
| import-boundary/static | `make backend-module-boundary-check` | Backend ownership/import rules | yes | Extend this guard only in the later authorized boundary slice. |
| full check | `make check` | Developer verification gate | no | Run after narrow and service-backed checks pass. |
| broader fast loop | `make test-fast` | Repository fast verification | no | Run before the full gate for a later implementation. |
| authority documents | `make lint-markdown` and `git diff --check` | Core 00, Core 01, Core 04, and tracker | yes | Required for this checkpoint together with the read-only 111/83 and 63-file assertions. |

`make agent-finalize` is not planned for this documentation-only authority checkpoint because no product snapshot is being validated and no harness-maintenance file may change. The later final implementation slice MUST run it under the repository handoff rules.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define `db/migrations` as infrastructure, not a permanent domain module | WF-00 | DONE | none | Core 01 §2.1A; §1 of this tracker | Target, exclusions, and authority are explicit. |
| T-002 | Inventory all 63 files and direct consumers | WF-01 | DONE | T-001 | §2 inventory | Every target file has one populated row. |
| T-003 | Map schema families to semantic owners | WF-02 | DONE | T-002 | SO and §3 | Every SQL file has an owner candidate or mixed-owner note. |
| T-004 | Freeze migration/runtime/operator/harness contracts | WF-03 | DONE | T-002, T-003 | §4 | Contract risks and characterization posture are explicit. |
| T-005 | Classify coupling and intentional centralization | WF-04 | DONE | T-002 | §5 | Findings use the required four classifications. |
| T-006 | Adopt Recovery 111/83 and conflict-fact ownership | WF-00, WF-02 | DONE | T-003 | Core 00/Core 01 amendments; RB-001 | Owner states 111/83 exactly, RC matches, and no SQL/schema change is proposed. |
| T-007 | Specify opaque production and harness catalog design | WF-05 | DONE | T-004, T-005 | REQ-01-657, TH-HARNESS-REQ-810/AC-094, §3 | Exact supported, restricted, and prohibited interfaces are decision-complete. |
| T-008 | Implement opaque production source cutover | WF-05 | TODO | T-007 | S-02/S-03 | Production callers no longer consume raw FS or caller-selected sources. |
| T-009 | Implement harness-only catalog capability | WF-07 | TODO | T-007 | S-04 | Targeted apply/rollback remains test-only with hash parity. |
| T-010 | Remove obsolete exports and install guard | WF-04 | TODO | T-008, T-009 | S-05 | Raw filesystem/lineage exports have no callers and cannot recur. |
| T-011 | Implement evidence v2 and remove path disclosure | WF-07 | TODO | T-008, T-010 | Adopted REQ-01-657/AC-537; S-07 | Closed v2 schema/output is current-only and every locator/v1 compatibility path is absent. |
| T-012 | Validate no migration/generated drift and full conformance | WF-08 | TODO | T-010, T-011 | S-06/S-08 command artifacts | Required focused, schema, boundary, drift, and broad checks pass. |
| T-013 | Frontend/grid refactor | WF-01 | DROPPED | T-002 | No target frontend/grid code found | Remains out of scope unless later evidence changes. |
| T-014 | Complete authority-first NLSpec handoff | WF-00 | DONE | T-001–T-007 | Adopted owner docs and §§1–12 | Another agent can implement S-01 through S-08 without an unresolved design decision. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | Branch `main`; commit `48c36c03666821937d58318586eb956a4ef19a9f`; initial worktree clean; planning only | Inspected framework, Core 00–05 applicability, domain/NLSpec guidance, adopted subsystem owners; touched only this tracker | `git branch --show-current`; `git rev-parse HEAD`; `git status --short`; `sed`/`rg` owner reads; `make lint-markdown` | Target is infrastructure, not a domain module; no owner contradiction found; tracker Markdown lint passed | RB-001, RB-002, RB-003 | Begin WF-05 only after blocker ownership is assigned. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | Branch and commit unchanged; Core owners now adopt 111/83, the opaque source split, and evidence v2 | Inspected analysis notes, NLSpec guidance, Core owners, RC, source/evidence/harness implementation, boundary and test manifests; touched Core 00, Core 01, Core 04, and tracker only | `git status --short`; `rg`/`sed`/`find`/`jq` reads; `make help-all`; `make lint-markdown`; `git diff --check` | Authority decisions are closed without claiming product implementation; documentation validation passed | none | Begin S-01 retained characterization in a later product task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | `Source()` is a thin facade candidate; raw FS, generic constructor, and lineage exports cross the desired boundary | Inspected `source.go`, Database Migrations implementation/tests, server/migrate/operator callers, pgtest/pgschema, tools; touched tracker only | `rg` import/symbol searches; `sed` exact source reads; `jq` manifest projections | Production and harness catalog capabilities are conflated; SQL remains centrally located | RB-002 | Produce an owner-approved opaque-source interface design. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | §3 defines the exact opaque source, restricted construction/bridge, safe inspection/digest, and sealed harness interfaces | Same implementation/boundary paths revalidated; touched authority documents and tracker only | Exact symbol/import searches and boundary-manifest reads | RB-002 is closed; raw FS, generic construction, and caller cutovers remain TODO | none | Implement S-02 through S-05 without moving SQL. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | Not applicable to direct target implementation | Inspected target and repository references; touched tracker only | `rg` target/caller searches | No HTTP handler, browser controller, UI-selector, or grid-vendor code exists in target | none | Keep frontend work out of scope unless a later schema-semantic change creates owner-specific risk. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | Still not applicable to the direct target or adopted owner changes | Revalidated target responsibilities; touched authority documents and tracker only | `rg` target/caller searches | No frontend, browser, selector, or grid contract is added | none | Keep frontend work out of S-01 through S-08 unless live evidence changes. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | 61 contiguous migrations; MH protected through 23; SQLC consumes entire directory; recovery projection asserts 111 tables | Inspected all target SQL, MH, SO, RC, `sqlc.yaml`, generated policy; touched tracker only | `find`, `wc`, `awk`, `rg`, `jq` inventory/object/dependency reads | All files mapped; generated files must not be hand-edited; Core 00 count differs from projection | RB-001 | Reconcile owner count before schema-affecting work; otherwise preserve exact bytes. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | Core and RC now agree on 111 total/83 authoritative; v2 machine contract location and closed shape are prescribed but not created | Revalidated Core, RC, Revisions contribution, evidence structs/tests, schema attachment conventions; touched authority documents and tracker only | `jq` 111/83/unique-entry assertions; `rg` schema and policy reads | RB-001 and RB-003 authority closed; no SQL, RC, generated, or contract file changed | none | S-07 adds the v2 projection only after S-01 through S-06. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | Twelve Database Migrations owner rows: seven unit and five service-backed; selected owner migrations have dedicated upgrades | Inspected test catalog/families, dedicated migration tests, Make guidance; touched tracker only | `make help`; `make help-all`; `make task-guide ROLE=module-author OWNER=module.database_migrations`; `make explain-test-owner OWNER=module.database_migrations`; target explanations for migration drift, generated policy, test-fast, test, and check | Canonical focused commands discovered; no implementation validation claimed | RB-002 | Design harness-only catalog access, then run focused unit/service-backed slices. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | Exact future harness signatures, target bounds, ownership, defaults, cleanup, and validation commands are specified | Revalidated TH-HARNESS-REQ-810/AC-094, pgtest/pgschema, owner rows, Make surface; touched authority documents and tracker only | `sed`/`rg` source and owner reads; `make help-all` | Test plan is decision-complete; no product or service-backed command ran in this checkpoint | none | S-01 retains baseline; S-04 performs the harness cutover. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | Target owns no authorization decision; schema supports owner authorization state; evidence output currently includes `manifest.path` | Inspected Core 04 AC-537, migration evidence, operator tests; touched tracker only | `rg`/`sed` exact contract and implementation reads | Filesystem-path output differs from owner conformance text; structural refactor must not silently alter it | RB-003 | Obtain later authorization for a separate owner correction. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | Core 01 and AC-537 authorize v2-only evidence with no locator or compatibility path | Revalidated evidence DTOs, Operator grammar/framing/resource tests, Core security text; touched authority documents and tracker only | Exact field/reason/command searches; Markdown/diff validation | RB-003 is closed; the live v1 producer remains a named TODO mismatch | none | Preserve v1 through S-06; perform the atomic S-07 cutover. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-10T15:30:45-04:00 | Codex `/root`, planning session | Inventory, owner map, freeze map, workflows, slices, validation, and blocker IDs are current | Touched only `docs/handoffs/db-migrations-refactor-tracker.md` | Planning discovery commands above; `make lint-markdown` | Tracker lint passed; no production refactor performed; no generated or harness file edited | RB-001, RB-002, RB-003 | Next recommended workflow: WF-05 facade or ownership redesign plan plus blocker resolution. |
| 2026-08-10T16:26:26-04:00 | Codex `/root`, authority-revision session | All authority decisions, mappings, interfaces, defaults, workstreams, slices, and binary criteria are current | Touched only the three Core owner documents and this tracker | Read-only consistency commands; `make lint-markdown`; `git diff --check` | No open blocker; no production, SQL, contract, generated, test, or harness change occurred | none | Next implementation checkpoint is S-01 retained characterization, followed by S-02. |

## 11. Open Questions and Blockers

No open planning question or owner contradiction remains. The table preserves the stable blocker IDs and records their authority-layer closure; downstream `TODO` implementation is tracked in §9 and MUST NOT reopen a closed decision without new owner evidence.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | What is the current Recovery catalog cardinality and classification for `record_revision_conflict_facts`? | Backup admission and historical interpretation require one authoritative answer. | Adopted REQ-01-647/575 plus RC parity: 111 total, 83 authoritative, one Revisions-owned exact-restore conflict-fact entry. | CLOSED: authority adopted; projection already matches; no SQL/schema change. |
| RB-002 | What exact interface supplies one opaque production source and one harness-only targeted catalog? | Raw FS or caller-selected source access would violate the production/test boundary. | REQ-01-657, TH-HARNESS-REQ-810/AC-094, and the exact §3 interface/import map. | CLOSED: behavior and implementation mapping are decision-complete; S-02 through S-05 remain TODO. |
| RB-003 | What replaces v1 `manifest.path`, and how is the breaking change versioned? | Current output violates the filesystem-disclosure boundary and consumers require one current schema. | Adopted REQ-01-657 and AC-537 plus the exhaustive §4 v2 mapping. | CLOSED: v2 removes the member without replacement or compatibility; S-07 remains TODO. |

There is no `BLOCKED: owner contradiction` finding in the inspected owner documents. The remaining v1 producer and raw-source surfaces are implementation/owner mismatches with exact completion criteria, not unresolved authority.

## 12. Binary Completion Criteria

- [x] Every file in `db/migrations` has exactly one inventory row; the inventory contains 63 rows.
- [x] Every discovered public contract risk has an owner, current-state statement, target-state statement, and test posture.
- [x] Recovery cardinality, category counts, conflict-fact ownership, restore treatment, and historical interpretation are exact and testable.
- [x] Production source, construction, inspection, digest, evidence, harness, and static-boundary interfaces are unambiguous, including invalid inputs and defaults.
- [x] The evidence v2 mapping accounts for every current field, removal, presence rule, ordering rule, forbidden locator, compatibility rule, and preserved Operator behavior.
- [x] Every workflow names predecessors, successors, validation, and an exit checkpoint.
- [x] Every slice names exact change, dependencies, risks, tests, validation, rollback, and binary completion. S-07 is the sole authorized observable schema-shape change.
- [x] Later unit, integration, schema, drift, static, broader, and full commands are discovered; browser work is inapplicable unless later schema semantics change.
- [x] RB-001 through RB-003 are closed at the authority layer without claiming their downstream implementation complete.
- [x] No `BLOCKED: owner contradiction` exists; implementation/owner mismatches remain visible as `TODO` work.
- [x] The framework's generic module vocabulary is narrowed by Core 01 §2.1A: the target is infrastructure, not a domain module.
- [x] Prior handoff history is preserved and the authority checkpoint records exact files, commands, results, and next action.
- [x] Only Core 00, Core 01, Core 04, and this tracker change; SQL, code, tests, contracts, generated outputs, manifests, and harness files remain unchanged.
- [x] Production implementation is explicitly reserved for S-01 through S-08 in a later task.
