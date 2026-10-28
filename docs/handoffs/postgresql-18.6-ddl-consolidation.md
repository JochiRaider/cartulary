# PostgreSQL 18.6 DDL consolidation

Status: complete. Research cutoff: **2026-10-17**;
repository work performed 2026-10-28. This is a reset-only pre-production transition.

## Working plan

1. Completed: inspect authority/clean checkout and characterize the original 95-file chain.
2. Completed: design 29 final owner files and a distinct v3 lineage; review PostgreSQL options.
3. Completed: implement final DDL, consumers, integrity fixes and owner amendments.
4. Completed: behavioral, privilege, Recovery and semantic comparison verification.
5. Completed: finalizer, final gates, cleanup evidence and handoff.

## Authority, scope and reference endpoint

Core 00 precedence applies: adopted subsystem specifications govern their scope,
then Core 00–04. Core 01 §2.1A, REQ-01-657/661, Core 02 source/history meaning,
Core 04 AC-537/542 and Testing Harness TH-HARNESS-REQ-811 govern this work.
Reference Pack, Reporting, Extensions, Graph Projection, Jobs, Projections and
Recovery retain their adopted owner contracts. `docs/domain.md` supplies vocabulary,
not independent behavior. No executable reads this document or other Markdown.

The checkout was clean at `4c22480888701a7b387074e10cd207b3e0828bc2`. Actual input:
95 numbered SQL files (13,207 lines), plus `source.go`. The final catalog has
29 SQL files (8,074 lines), with the same adapter. The original lineage was
`cartulary.prod_ddl_rebaseline.v2`, immutable through 29, actual head 95. References
to head 40/41 and Recovery 113/83 or 113/84 were stale. The actual current Recovery
catalog has 146 tables, 113 authoritative-required, including the unique Revisions
conflict-fact contribution. These numbers are preserved, not inferred from old prose.

The unmodified chain passed the complete migration owner slice at
`.cartulary/test-results/pg186-original-20261028` (40.902s). A second independent
provisioned fixture captured the endpoint at
`.cartulary/test-results/pg186-reference-capture-20261028` (17.266s).
The task-owned original catalog evidence is
`.cartulary/pg186-consolidation/original-catalog.json`; original SQL and machine
manifests are retained beside it as ignored task evidence, never a runnable legacy
catalog in the product. Both fixtures were cleaned by the harness. Actual server
observation was `180006`, checksums `on`; an image tag was not used as proof.

Original endpoint: 146 tables, 2 views, 1 owned revision sequence, 41 routines,
44 user triggers, 552 indexes, 2,248 constraints including PostgreSQL 18 NOT NULL
catalog entries, 1,724 live columns, zero RLS policies. Row security is not claimed
from role attributes. Capture includes owners, table/column/type/routine ACLs,
default privileges, trigger enable modes, extensions, generated expressions,
constraint enforcement/validation and seed contents.

The only nonempty fresh tables were lineage and application crypto identity.
The lineage row is migration initialization; crypto identity is committed by
successful application admission. No extension-profile metadata, Base packs,
trust roots or imported provenance are SQL seeds. Those remain initialized by
existing application owners and are verified by their runtime suites.

## Design and transition

The new lineage is `cartulary.prod_ddl_rebaseline.v3`, head 29, immutable through 29.
Files retain durable owner names. Infrastructure precedes auth/incidents/Jobs,
then record/history/source owners, projections, extension consumers and
cross-owner references. The Reporting preview-to-Composition FK is declared in 24,
after both owners exist. The revision sequence is created/owned in 08.

Pristine databases may initialize. Valid contiguous current-line prefixes may
advance through the migration owner; production crypto admission additionally
requires its established identity or a genuinely pristine initialization. An
interrupted fresh crypto initialization remains reset-only, as before. Exact head
passes readiness; current-line ahead is rejected; malformed/noncontiguous histories
are rejected before lineage comparison. Previous/foreign/unmarked/contaminated
histories fail before DDL with the v3 reset-only report at the migration owner.
Production facade cryptographic-state inspection retains precedence under
REQ-04-164: a missing/incompatible crypto identity on retained state reports
`ErrIncompatibleCryptoState` before schema admission. A cryptographically admitted
old lineage receives the v3 remediation report. Both paths prohibit initialization
or mutation of retained state; neither stamps compatibility. No ledger is rewritten.
The source remains one cached, validated embedded snapshot with deterministic
numeric ordering. There is no legacy source selector, compatibility view or bridge.

Each Up is transactional. FKs follow table definitions solely for dependency order;
there are no create-then-alter column cutovers. Down explicitly removes authored
objects in reverse dependency order, without CASCADE, and is disposable harness
cleanup only. Provisioned roles/extensions survive. Per-object grants replace the
old late broad grant-and-revoke phase; future defaults are closed in 01. PUBLIC row
type privileges and the narrow Reference Pack invalidated-content column grant
are retained. The old accidental Goose-sequence grants are not needed by runtime
or Recovery: ledger access remains read-only and bookkeeping is not restore data.

## Complete old-to-new mapping

Every row below includes the final effect of its original Up. `Folded` means folded
into final DDL, including constraints/indexes/routines/triggers/defaults and ACLs.
`Superseded` means intermediate state omitted; it does not discard current owner
behavior. `Retired` means objects/conversion omitted from a fresh database.
Verification codes: **C** catalog/semantic comparison + manifest parity;
**M** migration admission and full Up/Down/Up; **B** current owner behavioral suite;
**R** Recovery restore and reconstruction; **P** three-purpose privileges.

| Original migration | Final effect | New owner/file | Disposition | Verification |
| --- | --- | --- | --- | --- |
| `00001_database_infrastructure.sql` | Extension prerequisite checks and singleton lineage initialization; crypto-format table folded from 64/67. No application crypto identity is fabricated by SQL. | `00001_database_infrastructure.sql` | Folded into final DDL; retained lineage initialization | C, M, B, P |
| `00002_auth.sql` | Authentication, sessions, preferences and enrollment final shapes, including application-encrypted secret envelopes from 65. | `00002_auth.sql` | Folded into final DDL | C, M, B, P |
| `00003_incidents.sql` | Incident lifecycle, membership and workbook preference state; original scoped keys, attribution and defaults retained. | `00003_incidents.sql` | Folded into final DDL | C, M, B, P |
| `00004_recovery.sql` | Operator Recovery journal final encrypted envelope and transfer-operation vocabulary from 66/68. | `00004_recovery.sql` | Folded into final DDL | C, M, B, P, R |
| `00005_deployment_admin.sql` | Deployment bootstrap, backup and restore-verification metadata with append-only administrative events. | `00005_deployment_admin.sql` | Folded into final DDL | C, M, B, P |
| `00006_platform_jobs.sql` | Jobs and receipts retain final operator/user actor XOR and restore enumeration index from 44/53. | `00006_platform_jobs.sql` | Folded into final DDL | C, M, B, P |
| `00007_records.sql` | Record envelope identity, deletion attribution and version bounds; remove duplicate lookup index and add typed identity key. | `00007_records.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00008_revisions.sql` | Change sets, mutations, revisions, conflicts, history refs and allocation sequence; sequence belongs with its owning revision table. | `00008_revisions.sql` | Folded into final DDL | C, M, B, P, R |
| `00009_parties.sql` | Party final source without duplicate clocks; canonical active claims, private normalizer and ALWAYS triggers folded from 39/73. | `00009_parties.sql` | Folded into final DDL | C, M, B, P |
| `00010_timeline.sql` | Timeline source, provenance and time conversion profiles; interval observations and owner time semantics unchanged. | `00010_timeline.sql` | Folded into final DDL | C, M, B, P |
| `00011_entities.sql` | Entity sources, mentions, aliases and preserved identifiers with final Unicode mapping, source triggers and canonical claims from 36/37/41. | `00011_entities.sql` | Folded into final DDL; intentional modernization | C, M, B, P, R |
| `00012_indicators.sql` | Indicator source, intervals, support-reference validation and active registry identity; final digest key and value vocabulary from 38/48. | `00012_indicators.sql` | Folded into final DDL | C, M, B, P, R |
| `00013_assessments.sql` | Assessment source and confidence-band helper; current replay shape is written by application, not a SQL backfill. | `00013_assessments.sql` | Folded into final DDL | C, M, B, P |
| `00014_links.sql` | Links, tags and active views; preserve lifecycle predicates, replace split link uniqueness with NULLS NOT DISTINCT and fold lookup indexes45. | `00014_links.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00015_tasks_and_decisions.sql` | Task requests, decisions and related source state; preserve attributed lifecycle and record-envelope references. | `00015_tasks_and_decisions.sql` | Folded into final DDL | C, M, B, P |
| `00016_artifacts.sql` | Artifact sources and children; strengthen same-incident child references. | `00016_artifacts.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00017_evidence.sql` | Evidence sources, blobs, access/custody plus upload leases, cleanup claims and verification attestations from30/31/95. | `00017_evidence.sql` | Folded into final DDL; intentional modernization | C, M, B, P, R |
| `00018_saved_views.sql` | Saved View declarations and final incident/schema discovery ordering index from43. | `00018_saved_views.sql` | Folded into final DDL | C, M, B, P |
| `00019_imports.sql` | Import sessions, provenance, diagnostics and publication state; scoped Job references. | `00019_imports.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00020_network_flow.sql` | Network Flow imports, source rows, diagnostics and bindings; saved graph declarations from32 owned here. | `00020_network_flow.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00021_projections.sql` | Workbook projections and rebuildable state; current mention identities are produced by projection owner, no42 backfill. | `00021_projections.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00022_graph_projection.sql` | Only final Graph Projection results, vertices, edges and leases from32/34; five v1 tables are retired. | `00022_graph_projection.sql` | Folded into final DDL | C, M, B, P |
| `00023_reporting.sql` | Reporting final source-capture metadata, request-bound render recipe, sensitive/portable artifacts, profiles and immutable guards; later admission revisions folded. | `00023_reporting.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00024_report_composition.sql` | Report Composition versions, previews and bindings; late FK from Reporting preview outputs is declared here after target creation. | `00024_report_composition.sql` | Folded into final DDL | C, M, B, P |
| `00025_incident_bundles.sql` | Incident Bundle import/export state and staging; retain stable operation identities and scoped Jobs. | `00025_incident_bundles.sql` | Folded into final DDL; intentional modernization | C, M, B, P |
| `00026_reference_data.sql` | Only immutable Reference Pack administration, trust, sets, pins, provenance, indexes, portable preparation and dependency guards; four mutable tables retired. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P, R |
| `00027_extensions.sql` | Extension metadata, migration/operation ledgers, outbox, staging and proofs; profile rows remain startup-owned, never guessed SQL seeds. | `00027_extensions.sql` | Folded into final DDL | C, M, B, P, R |
| `00028_administrative_audit.sql` | Administrative audit projections with final Reference Pack event vocabulary and append-only trigger. | `00028_administrative_audit.sql` | Folded into final DDL | C, M, B, P |
| `00029_collaboration.sql` | Collaboration intents, cursors, replay and resume state. Broad privilege repair replaced by explicit per-object grants at creation and defaults in 01. | `00029_collaboration.sql` | Folded into final DDL | C, M, B, P |
| `00030_evidence_upload_leases.sql` | Upload leases, issuing actor/session, expiry and indexes remain in Evidence. | `00017_evidence.sql` | Folded into final DDL | C, M, B, P |
| `00031_evidence_blob_cleanup_claims.sql` | Cleanup claim states, retry/lease indexes and blob/association guards remain; blob guard uses final95 definition. | `00017_evidence.sql` | Folded into final DDL | C, M, B, P |
| `00032_graph_projection_v2.sql` | Final v2 materializations and immutable result guards retained; saved graph declarations allocated to Network Flow. No v1 reconstruction. | `00020_network_flow.sql`, `00022_graph_projection.sql` | Folded into final DDL | C, M, B, P, R |
| `00033_graph_projection_v1_removal.sql` | Five v1 Graph tables, retained-result FK and cutover-only inventory disappear. Final v2 schema and declaration/result separation preserved. | `00022_graph_projection.sql` | Retired object / obsolete conversion | C, M, B, P |
| `00034_graph_projection_cleanup_indexes.sql` | Cleanup candidate index retained with source owner, publication time and result identity. | `00022_graph_projection.sql` | Folded into final DDL | C, M, B, P |
| `00035_assessment_create_idempotency_v1.sql` | Legacy assessment idempotency JSON conversion and rollback transformation retired; fresh application writes canonical create_result.v1. | `00002_auth.sql`, `00013_assessments.sql` | Retired object / obsolete conversion | C, M, B, P |
| `00036_entities_source_integrity.sql` | Source/envelope mirror, typed children, mention vocabulary/paired fields and integrity triggers retained; empty-source preflight omitted. | `00011_entities.sql` | Folded into final DDL | C, M, B, P |
| `00037_entities_active_identifier_claims.sql` | Active identifier claim relation, exact-set validator, rebuild/refresh/release routines and ALWAYS triggers retained; empty backfill omitted; earlier normalization replaced by 41. | `00011_entities.sql` | Folded into final DDL | C, M, B, P, R |
| `00038_indicators_lifecycle_integrity.sql` | Support refs enforce unique UUID members and incident scope; prior duplicate-accepting helper superseded. No historical preflight. | `00012_indicators.sql` | Folded into final DDL | C, M, B, P |
| `00039_parties_active_key_claims.sql` | Party canonical keys and claims, unique arbitration, grants and ALWAYS triggers retained; empty backfill omitted. | `00009_parties.sql` | Folded into final DDL | C, M, B, P, R |
| `00040_parties_mutation_request_hash_v1.sql` | Disposal of old Party mutation-hash replay rows retired; no replay rows exist in fresh baseline. Current application hash semantics unchanged. | `00002_auth.sql`, `00009_parties.sql` | Retired object / obsolete conversion | C, M, B, P |
| `00041_entities_identifier_unicode_v1.sql` | Exact Unicode scalar mapping replaces earlier database casing; temporary previous-normalizer routine and reconciliation preflight retired. | `00011_entities.sql` | Folded into final DDL | C, M, B, P |
| `00042_timeline_mention_public_identity.sql` | Timeline mention public-identity projection backfill retired. Current projection/rebuild producer retains canonical mention IDs. | `00021_projections.sql` | Retired object / obsolete conversion | C, M, B, P |
| `00043_saved_views_schema_discovery.sql` | Saved View schema discovery index retained. | `00018_saved_views.sql` | Folded into final DDL | C, M, B, P |
| `00044_jobs_restore_enumeration.sql` | Jobs nonterminal restore enumeration index retained. | `00006_platform_jobs.sql` | Folded into final DDL | C, M, B, P |
| `00045_note_association_lookup.sql` | Outgoing/incoming null-field note association pagination indexes retained. | `00014_links.sql` | Folded into final DDL | C, M, B, P |
| `00046_reference_pack_immutable_state.sql` | Immutable Reference Pack final tables/routines/ACLs retained; initial attempt-member immutability replaced by50/52 guard. Legacy-empty cutover omitted. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P, R |
| `00047_reporting_reference_pack_binding.sql` | Reporting reference binding incorporated into render bundle final columns/constraints; later exact-model contract retained. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00048_indicator_registry_identity.sql` | Registry-owned Indicator dedupe, non-null normalized value, STORED digest key, unique claims and final functions retained. No historical rekeying. | `00012_indicators.sql` | Folded into final DDL | C, M, B, P |
| `00049_reference_pack_audit.sql` | Reference Pack audit event vocabulary incorporated in Administrative Audit constraint. | `00028_administrative_audit.sql` | Folded into final DDL | C, M, B, P |
| `00050_reference_pack_invalidation_evidence.sql` | Invalidated-content evidence column and narrow UPDATE(column) grant retained; final attempt guard comes from52. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00051_reference_pack_portability_retention.sql` | Portable catalogs, pins, operation kind and immutable provenance retained; no imported-incident attribution reconstruction. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00052_reference_pack_validation_summaries.sql` | Validation summary reference STORED column/index and final guarded attempt update retained. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00053_reference_pack_operator_jobs.sql` | Jobs nullable user actor and explicit operator operation identity with XOR constraint retained. | `00006_platform_jobs.sql` | Folded into final DDL | C, M, B, P |
| `00054_reference_pack_dependency_guards.sql` | Operation dependency-key records, index and immutable guard retained. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00055_reference_pack_portability_requirements.sql` | Portable-content request evidence CHECK retained; missing old required-member evidence is not invented. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00056_reference_pack_portable_preparation.sql` | Portable preparation and selections, exact member binding, pins, indexes and immutability retained. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00057_reference_pack_preparation_privileges.sql` | Required table grants and row-type USAGE incorporated when portable preparation tables are created. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00058_reference_pack_retire_legacy_tables.sql` | Four mutable Reference Pack tables omitted entirely; locked empty-only removal choreography retired. | `00026_reference_data.sql` | Retired object / obsolete conversion | C, M, B, P |
| `00059_reporting_portable_artifacts.sql` | Imported Reporting artifact file table, FK/index and immutable-update trigger retained. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00060_reporting_artifact_trigger_privileges.sql` | Narrow trigger execution privileges retained at creation; no PUBLIC execute. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00061_reference_pack_fallback_attribution.sql` | Fallback-from-version provenance and FK retained; partial FK index superseded by62. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00062_reference_pack_fallback_full_index.sql` | Full fallback-version FK covering index retained, replacing partial61 index. | `00026_reference_data.sql` | Folded into final DDL | C, M, B, P |
| `00063_reporting_admitted_model_identity.sql` | Initial snapshot-model identity CHECK superseded by75/84/85/86/91/93. Retain final source-capture identity, not retired JSON storage. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00064_application_crypto_format.sql` | Application crypto-format relation retained in infrastructure; SQL has no identity seed. Admission commits singleton only after successful fresh initialization. | `00001_database_infrastructure.sql` | Folded into final DDL; retained application initialization boundary | C, M, B, P, R |
| `00065_auth_secret_envelopes.sql` | Encrypted authentication secret envelope columns retained; obsolete plaintext columns and preflight omitted. | `00002_auth.sql` | Folded into final DDL | C, M, B, P |
| `00066_recovery_journal_envelopes.sql` | Encrypted journal envelope final columns retained; old journal payloads never converted. | `00004_recovery.sql` | Folded into final DDL | C, M, B, P, R |
| `00067_application_crypto_format_privileges.sql` | Read-only runtime/Recovery grants on crypto format retained; neither may initialize its identity. | `00001_database_infrastructure.sql` | Folded into final DDL | C, M, B, P |
| `00068_recovery_transfer_operations.sql` | Recovery transfer-operation journal vocabulary retained. | `00004_recovery.sql` | Folded into final DDL | C, M, B, P, R |
| `00069_reporting_sensitive_artifacts.sql` | Sensitive reveal-map tables and grants plus public artifact-member exclusion CHECKs retained. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00070_reporting_profile_admission.sql` | Initial render profile admission CHECK superseded by76/80/88/91/94 final recipe shape. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00071_reporting_preview_admission.sql` | Preview admission identity retained in final Reporting payload contract; old intermediate JSON checks folded/superseded. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00072_reporting_template_metadata.sql` | Explicit template metadata admission retained through final template/recipe model; no default metadata backfill. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00073_party_envelope_clock_authority.sql` | Party created_at/updated_at columns omitted; Records envelope owns clocks. No guessed clock conversion or downgrade. | `00009_parties.sql` | Folded into final DDL | C, M, B, P |
| `00074_reporting_failure_diagnostics.sql` | Failure diagnostics admission retained in final payload definition; no historical reconstruction. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00075_reporting_frozen_source_facts.sql` | Frozen source facts move through later model revisions into final source-capture representation; obsolete snapshot JSON CHECK superseded. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00076_reporting_profile_policy_cutover.sql` | Second profile-policy admission superseded by current final structured recipe contract; no retained-policy conversion. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00077_reporting_subject_identity_cutover.sql` | Historical subject-identity JSON CHECK removed with snapshot full-model column93; final capture/materializer owns subject identity. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00078_reporting_token_artifact_cutover.sql` | Token redaction manifest v3 replaced by79/88 final v5 checks. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00079_reporting_structural_removal_cutover.sql` | Structural-removal redaction v4 superseded by88 v5 checks; no legacy conversion. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00080_reporting_retained_field_policy.sql` | Retained field-policy revision 3 superseded by88/91 final recipe/profile revision; source semantics remain application-owned. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00081_reporting_retained_graph_sources.sql` | Admitted graph source digest final column and checks retained in payload. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00082_reporting_template_manifest.sql` | Template manifest v2 check superseded by91/94 final recipe; exact template binding retained. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00083_reporting_presentation_text_admission.sql` | Presentation text v1 admission superseded by84/85/86/91/94 final recipe. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00084_reporting_source_index_admission.sql` | Source-index model v3 admission superseded by85/86/91/93/94; final source-capture and recipe semantics retained. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00085_reporting_core_record_admission.sql` | Core record model v4 admission superseded by86/91/93/94; source owner identity preserved. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00086_reporting_lifecycle_admission.sql` | Lifecycle model v5 admission superseded by91/93/94; deletion/source evidence preserved in current application capture. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00087_reporting_bundle_model_identity.sql` | Bundle model identity final columns/checks retained. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00088_reporting_container_redaction.sql` | Container redaction v5 CHECKs retained; intermediate profile revision4 superseded by91. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00089_reporting_template_pack_admission.sql` | Admitted template-manifest digest column and constraint retained. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00090_reporting_theme_toolchain_inventory.sql` | Toolchain inventory v2 CHECK superseded by91 structured-redaction inventory. | `00023_reporting.sql` | Superseded intermediate state | C, M, B, P |
| `00091_reporting_structured_redaction.sql` | Structured-redaction model/inventory retained in final recipe pipeline; snapshot v6 JSON CHECK removed93 and request checks superseded94. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00092_reporting_snapshot_job_admission.sql` | Independent snapshot admission JSON/digest columns retained, nullable request permitted only in the closed final tuple. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00093_reporting_snapshot_source_capture.sql` | Snapshot full export_model_json retired; source_capture_sha256 final column/check retained. No materialized-model backfill. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00094_reporting_render_recipe.sql` | Final request-bound render recipe and admitted metadata columns/constraint retained directly. | `00023_reporting.sql` | Folded into final DDL | C, M, B, P |
| `00095_evidence_blob_verification.sql` | Verification attestations and upload-lease association retained; finalize_attempt_count removed, terminal reasons and cleanup guard use final definitions. | `00017_evidence.sql` | Folded into final DDL | C, M, B, P, R |

## PostgreSQL feature decisions

Decisions use primary version 18 documentation available at the requested cutoff.
PostgreSQL 18 features and existing capabilities available in 18.6 are distinguished
from 18.6 maintenance fixes. “Retain” is an applicability decision, not an omission.

| Area | Current realization | PostgreSQL 18.6 option | Decision | Rationale and contract impact | Verification |
| --- | --- | --- | --- | --- | --- |
| Incident references | Several child incident IDs independent of globally unique parent IDs | Composite FK | Adopt for 36 relationships | Close cross-incident associations without changing legitimate IDs; full parent keys and child index coverage | C, B; evidence/blob and link rejection |
| Entity discrimination | Trigger-enforced child type | Typed composite FK | Adopt on 3 child relations | Preserve host/identity distinction even under deferred mutations; source trigger remains for stronger owner rules | C, entity source/claims suites |
| Null identity | Two active-link indexes split on nullable field key | `NULLS NOT DISTINCT` | Adopt one exact active key | Same product identity; existing Reference Pack null-safe keys retained | Null/non-null duplicates, deletion/recreate, concurrency |
| Lifecycle uniqueness | Active/deleted and pending predicates | Partial unique indexes | Retain exact predicates | Do not widen exclusivity to deleted history, other states or intervals | Owner suites, captured predicates |
| Constraint names | Named PK/FK/CHECK/UNIQUE; column NOT NULL | Named NOT NULL catalog support | Retain named relational constraints and deterministic NOT NULL names | No staged deployment needs a separately managed NOT NULL name; three historical rename/collision names normalize in new lineage | Full enforced/validated catalog check; [CREATE TABLE](https://www.postgresql.org/docs/18/sql-createtable.html) |
| Enforcement vs validation | Ordinary enforced constraints | NOT ENFORCED / NOT VALID | Reject for baseline | Fresh installation validates everything; no indefinitely unenforced integrity | conenforced/convalidated assertions |
| Staging and locking | Fresh transactional Up | Staged NOT NULL/FK validation | Retain direct validation | Empty tables avoid live-data validation scans and rewrites; retained upgrade choreography has no supported user | [ALTER TABLE](https://www.postgresql.org/docs/18/sql-altertable.html), M |
| Deferred references | Deferred cyclic Reference Pack and entity relationships | DEFERRABLE, MATCH rules and referential actions | Retain; type FK deferred | Preserve multi-step owner transactions; uniqueness used by conflict arbitration remains immediate. No blanket cascade changes | C, concurrent claims, import/restore |
| Nullable tuples and ranges | CHECK expressions; some SQL UNKNOWN paths | Explicit non-null paired members | Correct Reporting composition tuple | Non-null composition requires version and digest; remaining owner vocabularies/bounds preserved | Named rejection test and owner lifecycle tests |
| Unicode equality | Owner NFC/scalar mapping, Party C collation, citext where adopted | casefold, pg_unicode_fast, nondeterministic collation | Retain exact owner algorithms | Alternative equality could merge stable identities or change search; no demonstrated parity | Entity closed corpus/scalar casing; [collations](https://www.postgresql.org/docs/18/collation.html) |
| Generated digest keys | Indicator dedupe and Reference Pack lookup digest STORED | VIRTUAL default in 18 | Explicitly retain STORED | Expressions use pgcrypto; avoid read-time evaluation/privilege changes and retain indexed recovery evidence | C, restore derived-value validation |
| Generated summary reference | Attempt-member summary ID STORED | VIRTUAL/expression index | Retain STORED | Summary is a queryable column with a lookup index; no measured storage pressure justifies changing contract | SQLC and Reference Pack tests |
| Generated restrictions | Immutable current-row expressions | New virtual restrictions | Retain immutable expressions | No subquery, cross-row or volatile generated logic; extension-dependent digest is unsuitable for virtual generation | [Generated columns](https://www.postgresql.org/docs/18/ddl-generated-columns.html), C |
| Generated security/replication/statistics | Column ACLs, write-time generation, no publication | Virtual privilege timing; stored replication support | Retain current model | Recovery omits generated columns on insert and checks source-derived values; no new publication or trigger access assumptions | P, R |
| Expression indexes | Owner-specific expressions and hashes | Replace stored fields with expression keys | Retain where already used | No speculative duplicate expression index; retained column consumers and collision semantics remain explicit | C, SQLC |
| Search | Owner-maintained projection/search fields and explicit normalizers | Generated tsvector/new tokenization | Retain | A different dictionary, collation or text-search configuration changes matching; cannot derive cross-record projections from row expressions | Owner query suites; [text-search indexing](https://www.postgresql.org/docs/18/textsearch-tables.html) |
| Temporal uniqueness | Indicator intervals and event observations may overlap | WITHOUT OVERLAPS | Do not adopt | Open ends and overlaps are meaningful observations, not competing temporal ownership; no invented event-time exclusion | Overlap acceptance and lifecycle suite |
| Temporal coverage | History/references use stable identities and revisions | PERIOD FK | Do not adopt | Owners do not require interval coverage. Equal endpoints, supersession and reconstruction retain existing timestamp semantics | C, history/rollback suites; [CREATE TABLE](https://www.postgresql.org/docs/18/sql-createtable.html) |
| UUID generation | Application/import IDs and random UUID defaults | uuidv7/uuidv4 aliases | Retain v4/default spelling | No measured locality benefit outweighs identity-contract changes. UUID order never represents event or commit order | C, imports/history; [UUID functions](https://www.postgresql.org/docs/18/functions-uuid.html) |
| Revision allocation | Owned bigint sequence with guarded restore allocation | Identity column | Retain explicit sequence | Revisions uses its stable name and allocation protocol; BY DEFAULT identity offers no needed guarantee | Allocation rollback gap test, P, R |
| Sequence restoration | Recovery UPDATE/setval; runtime USAGE | Transactional restart vs setval | Retain owner protocol | nextval/setval are not ordinary transactional DDL; gaps are legal and allocation does not prove commit ordering | [sequence functions](https://www.postgresql.org/docs/18/functions-sequence.html), restore tests |
| Index necessity | 552 original indexes | Redundancy removal | Remove 3 duplicates; merge 2 link indexes | Keep order/predicate/operator semantics; new composite-reference coverage accounts for added indexes | C, FK coverage manifest |
| B-tree skip scan | Incident-leading indexes | PostgreSQL 18 skip scan | Retain engine benefit, no DDL rewrite | Optimizer opportunity does not justify removing deterministic FK/index coverage | [multicolumn indexes](https://www.postgresql.org/docs/18/indexes-multicolumn.html), C |
| INCLUDE/operator classes | Existing index shapes | Covering, BRIN/GiST/GIN alternatives | Retain | No measured workload proves a new covering index; larger writes/maintenance need owner evidence | Inventory and query tests |
| Partitioning/physical layout | Logged ordinary tables | Declarative partitioning, unlogged data, compression tuning | Defer | No volume/retention benchmark or restore/unique-key design supports these changes; authoritative history must survive recovery | C, Recovery catalog |
| Concurrent index operations | Transactional migrations | CREATE/REINDEX CONCURRENTLY | Do not adopt in baseline | Requires a nontransactional recovery protocol; fresh tables have no production writers | [CREATE INDEX](https://www.postgresql.org/docs/18/sql-createindex.html), M |
| Migration coordination | Session advisory lock and per-file transactions | PostgreSQL lock modes | Retain | Explicit dependency order plus serialized admission; no live rewrite workload | Cancellation/failure tests; [locking](https://www.postgresql.org/docs/18/explicit-locking.html) |
| Roles and routines | Administrator prerequisites; three purposes | Role changes, controlled definer execution | Retain and move grants to creation | Same owners and narrow routine ACLs, controlled search_path; no runtime DDL/Recovery authority | P, connection reuse matrices |
| RLS | No policies or row-security-enabled tables | Add RLS policies | Do not adopt incidentally | Authorization remains owner/application controlled; adding RLS is a separate authorization contract | Captured zero policies, permission suites |
| Bypass restore | Replica mode skips FK and ordinary triggers | Explicit scans and owner validators | Adopt | Re-enabling triggers/VALIDATE on valid metadata does not prove loaded data; reject orphan data before commit | New bypass-load rollback test, R |
| AIO, planner and vacuum | Standard pinned engine | PostgreSQL 18 engine improvements | Retain engine defaults | No DDL action required; host tuning needs measured deployment workload | Actual server version, owner suite; [18 release](https://www.postgresql.org/docs/18/release-18.html) |
| RETURNING OLD/NEW, MERGE, COPY options | Current SQLC queries and strict imports | New query features | Defer outside this schema slice | No need to change mutation receipts or make imports tolerate malformed rows | Existing query/receipt tests |
| OAuth/authentication | Verified purpose-specific TLS certificates | PostgreSQL 18 OAuth; MD5 deprecation | Retain certificates | Different authentication flow would change provisioning authority; no MD5 dependency introduced | Platform connection tests |
| pg_upgrade/statistics and logical replication | Reset-only baseline; owner logical Recovery | Preserved optimizer stats/publication options | Not applicable | No physical upgrade or replication transition is authorized; recovery classifications remain unchanged | Deployment procedure, R |

### PostgreSQL 18.6 maintenance applicability

18.6 is an August 13 maintenance release over 18.4; 18.5 was not released.
The [18.6 release notes](https://www.postgresql.org/docs/18/release-18-6.html)
identify conditional GIN statistics, btree_gist/ltree reindex, output-plugin
allowlisting and pgcrypto PGP cleanup. This catalog is built fresh on 18.6;
there are no older index pages to repair. Installed extensions are only plpgsql,
pgcrypto and citext. No btree_gist/ltree, PGP-encrypted SQL data, custom decoding
plugin, publication, extended-statistics object, partitioned/exclusion index,
unlogged sequence or SQL-callable internal-type routine was found. No conditional
reindex, cipher override or output-plugin setting is added.

The release also fixes role-plan invalidation, type privilege checks, NFC Hangul
normalization, text-search safety, deferred-uniqueness reindex, virtual-column JIT,
serializable conflict detection and WAL/backup correctness. Running the pinned
engine adopts those fixes. Existing purpose/ACL tests and Unicode corpus test
Cartulary's exposed behavior; we do not claim to regression-test PostgreSQL itself.
Logical-replication, psql/COPY, ecpg, physical-upgrade and unrelated planner fixes
require no schema rewrite here. Retained-cluster remediation is outside this
fresh reset-only cutover, not silently declared performed.


## Intentional differences and comparison

The comparison is semantic and keyed by object identity. Catalog OIDs are not
captured. Lineage initialization timestamps are installation observations, not
product state. Physical column numbers are retained in the raw captures and
reported below; their compaction is not silently normalized away. Goose bookkeeping
rows/sequence allocation are excluded from endpoint equality because execution
count and timing intentionally change; admission and cleanup test their behavior.
No table data, ACL, expression, lifecycle predicate or owner difference is ignored.

| ID | Intentional difference | Reason and verification |
| --- | --- | --- |
| D01 | 95 historical files become 29 final owner files; v2 becomes v3 | Authorized reset-only lineage; all prefixes, rejection matrix, exact-head readiness, failure atomicity and Up/Down/Up |
| D02 | 35 existing FKs include incident identity and one additional scoped FK complements a global job FK; 15 parent incident keys and one typed record key are added | Close cross-incident writes; FK actions/deferral unchanged; catalog coverage and rejection tests |
| D03 | Three entity child tables acquire typed record FKs and supporting indexes | Preserve entity type in relational identity, alongside existing source triggers |
| D04 | Two active link unique indexes become one NULLS NOT DISTINCT index | Preserve exact active key semantics for absent/present field keys and soft-deleted history |
| D05 | Three redundant indexes are removed | Equivalent existing leading keys/predicates cover their lookups; no constraint index is dropped |
| D06 | Reporting composition CHECK explicitly rejects NULL version/digest when composition identity is present | Fix SQL UNKNOWN acceptance; named CHECK rejection test |
| D07 | Recovery scans every actual FK and source/claim predicates before commit; existing private Entity and Indicator validators are extended | Bypass-loaded rows can violate constraints whose catalog flags remain valid; negative restore proves rollback |
| D08 | 24 live column positions compact, three automatically named NOT NULL constraints normalize | Remove dropped-column and rename history. Types, nullability, defaults, generated expressions and consumers are preserved |
| D09 | Per-object grants replace late broad grants; Goose bookkeeping sequence loses accidental runtime/Recovery grants | Preserve all application-object and column ACLs; ledger needs SELECT, not sequence allocation/restore authority |
| D10 | SQLC receives explicit nullable/non-nullable public.citext overrides | Qualified extension type retains the exact existing Go string/pgtype.Text projections; no generated SQLC model difference |

The original 552 indexes become 598: remove five, add 51. Of the additions,
16 back the new parent unique constraints, 34 cover composite/typed child keys,
and one is the replacement active-link key. This is an explicit integrity/index
coverage cost, not a claim of reduced total index storage or faster queries.
Required child coverage follows the adopted manifest; partitioning and speculative
covering indexes are not introduced. Constraints increase from 2,248 to 2,268:
16 unique keys, three typed FKs and one supplemental scoped job FK; three NOT NULL names change without changing
nullability. Tables, views, sequences, routine count, user triggers, live columns,
Recovery classifications and extension inventory are unchanged.

The retained duplicate coverage is deliberate: existing global-ID or paginated
query indexes remain where a new incident-leading composite key is not an exact
substitute. The removed redundant indexes are `records_incident_lookup_idx`,
`record_tags_active_record_lookup_idx`, and
`record_history_entry_refs_record_lookup_idx`. Active-link replacements remove
`record_links_active_unique_field_idx` and `record_links_active_unique_no_field_idx`
in favor of `record_links_active_identity_idx`.

### Strengthened incident references

| Child table / existing constraint | Parent | Final identity |
| --- | --- | --- |
| `artifact_findings.artifact_findings_record_id_fkey` | `artifacts` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `artifact_forensic_keywords.artifact_forensic_keywords_record_id_fkey` | `artifacts` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `artifact_grid_projection.artifact_grid_projection_record_id_fkey` | `artifacts` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `artifact_investigative_queries.artifact_investigative_queries_record_id_fkey` | `artifacts` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `assessment_grid_projection.assessment_grid_projection_record_id_fkey` | `assessments` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `decision_grid_projection.decision_grid_projection_record_id_fkey` | `decisions` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `evidence.evidence_object_blob_fkey` | `object_blobs` | `(incident_id, object_blob_id)` → `(incident_id, object_blob_id)` |
| `evidence_access_handles.evidence_access_handles_object_blob_id_fkey` | `object_blobs` | `(incident_id, object_blob_id)` → `(incident_id, object_blob_id)` |
| `evidence_custody_events.evidence_custody_events_evidence_record_id_fkey` | `evidence` | `(incident_id, evidence_record_id)` → `(incident_id, record_id)` |
| `evidence_grid_projection.evidence_grid_projection_record_id_fkey` | `evidence` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `evidence_object_upload_leases.evidence_object_upload_leases_blob_fkey` | `object_blobs` | `(incident_id, object_blob_id)` → `(incident_id, object_blob_id)` |
| `host_grid_projection.host_grid_projection_record_id_fkey` | `hosts` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `hosts.hosts_merged_into_record_id_fkey` | `hosts` | `(incident_id, merged_into_record_id)` → `(incident_id, record_id)` |
| `identities.identities_merged_into_record_id_fkey` | `identities` | `(incident_id, merged_into_record_id)` → `(incident_id, record_id)` |
| `identity_grid_projection.identity_grid_projection_record_id_fkey` | `identities` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `import_sessions.import_sessions_apply_job_id_fkey` | `jobs` | `(incident_id, apply_job_id)` → `(incident_id, job_id)` |
| `import_sessions.import_sessions_discovery_job_id_fkey` | `jobs` | `(incident_id, discovery_job_id)` → `(incident_id, job_id)` |
| `incident_bundle_exports.incident_bundle_exports_export_job_id_fkey` | `jobs` | `(incident_id, export_job_id)` → `(incident_id, job_id)` |
| `incident_bundle_job_payloads.incident_bundle_job_payloads_incident_job_fkey` | `jobs` | `(incident_id, job_id)` → `(incident_id, job_id)` |
| `indicator_active_identities.indicator_active_identities_indicator_fkey` | `indicators` | `(incident_id, indicator_record_id)` → `(incident_id, record_id)` |
| `indicator_grid_projection.indicator_grid_projection_record_id_fkey` | `indicators` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `network_flow_graph_views.network_flow_graph_views_latest_job_id_fkey` | `jobs` | `(incident_id, latest_job_id)` → `(incident_id, job_id)` |
| `network_flow_indicator_bindings.network_flow_indicator_bindings_target_indicator_record_id_fkey` | `indicators` | `(incident_id, target_indicator_record_id)` → `(incident_id, record_id)` |
| `network_flow_rejected_row_diagnostics.network_flow_rejected_row_diagnostics_table_id_fkey` | `network_flow_tables` | `(incident_id, network_flow_table_id)` → `(incident_id, network_flow_table_id)` |
| `network_flow_rows.network_flow_rows_table_id_fkey` | `network_flow_tables` | `(incident_id, network_flow_table_id)` → `(incident_id, network_flow_table_id)` |
| `network_flow_tables.network_flow_tables_source_import_session_id_fkey` | `import_sessions` | `(incident_id, source_import_session_id)` → `(incident_id, import_session_id)` |
| `party_grid_projection.party_grid_projection_record_id_fkey` | `parties` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `report_composition_preview_attempts.report_composition_preview_attempts_composition_id_fkey` | `report_compositions` | `(incident_id, composition_id)` → `(incident_id, composition_id)` |
| `report_composition_preview_attempts.report_composition_preview_attempts_render_attempt_fk` | `jobs` | `(incident_id, render_attempt_id)` → `(incident_id, job_id)` |
| `reporting_job_payloads.reporting_job_payloads_job_id_fkey` | `jobs` | `(incident_id, job_id)` → `(incident_id, job_id)` |
| `reporting_releases.reporting_releases_create_job_id_fkey` | `jobs` | `(incident_id, create_job_id)` → `(incident_id, job_id)` |
| `reporting_releases.reporting_releases_snapshot_id_fkey` | `reporting_snapshots` | `(incident_id, snapshot_id)` → `(incident_id, snapshot_id)` |
| `reporting_snapshots.reporting_snapshots_create_job_id_fkey` | `jobs` | `(incident_id, create_job_id)` → `(incident_id, job_id)` |
| `task_request_grid_projection.task_request_grid_projection_record_id_fkey` | `task_requests` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `timeline_grid_projection.timeline_grid_projection_record_id_fkey` | `timeline_events` | `(incident_id, record_id)` → `(incident_id, record_id)` |
| `timeline_grid_projection.timeline_grid_projection_replacement_record_id_fkey` | `timeline_events` | `(incident_id, replacement_record_id)` → `(incident_id, record_id)` |

Typed references on `entity_aliases`, `entity_preserved_identifiers` and
`entity_active_identifier_claims` use `(incident_id, record_id, entity_type)` against
`records(incident_id, record_id, record_type)`, DEFERRABLE INITIALLY DEFERRED.

The import operation is an explicit scope exception under Core 01 §17.5 and Jobs:
`reporting_imported_artifact_files.operation_id` references a deployment-scoped
job by global ID. Its output incident does not become the job's scope.
`incident_bundle_job_payloads` retains the original global job FK plus the new
incident/job FK for exports. With a nullable incident key, the composite FK alone
would exempt import rows under MATCH SIMPLE and lose the existence guarantee.
The baseline tests reject an orphan import payload, accept a deployment import
and its artifact, and reject an export bound to that deployment job. The existing
portable historical Reporting round trip remains the end-to-end witness.

### Named additions and physical-name changes

The following explicit list closes the object-level difference register. Existing
index definitions not listed here are unchanged; new indexes are ordinary valid
transactional B-trees. Their exact expressions and predicates are in the machine
manifest and captured endpoint, not inferred from this document.

Added constraints:

- `incident_bundle_job_payloads.incident_bundle_job_payloads_incident_job_fkey`
- `artifacts.artifacts_incident_identity_key`
- `assessments.assessments_incident_identity_key`
- `decisions.decisions_incident_identity_key`
- `entity_active_identifier_claims.entity_active_identifier_claims_typed_record_fkey`
- `entity_aliases.entity_aliases_typed_record_fkey`
- `entity_preserved_identifiers.entity_preserved_identifiers_typed_record_fkey`
- `evidence.evidence_incident_identity_key`
- `hosts.hosts_incident_identity_key`
- `identities.identities_incident_identity_key`
- `import_sessions.import_sessions_incident_identity_key`
- `jobs.jobs_incident_identity_key`
- `network_flow_tables.network_flow_tables_incident_identity_key`
- `object_blobs.object_blobs_incident_identity_key`
- `parties.parties_incident_identity_key`
- `records.records_incident_record_type_key`
- `report_compositions.report_compositions_incident_identity_key`
- `reporting_composition_preview_output_files.reporting_composition_preview_outpu_preview_attempt_id_not_null`
- `reporting_composition_preview_outputs.reporting_composition_preview_outp_preview_attempt_id_not_null1`
- `reporting_snapshots.reporting_snapshots_incident_identity_key`
- `reporting_snapshots.reporting_snapshots_source_capture_sha256_not_null`
- `task_requests.task_requests_incident_identity_key`
- `timeline_events.timeline_events_incident_identity_key`

Added indexes:

- `artifact_grid_projection.artifact_grid_projection_record_id_scope_idx`
- `artifacts.artifacts_incident_identity_key`
- `assessment_grid_projection.assessment_grid_projection_record_id_scope_idx`
- `assessments.assessments_incident_identity_key`
- `decision_grid_projection.decision_grid_projection_record_id_scope_idx`
- `decisions.decisions_incident_identity_key`
- `entity_active_identifier_claims.entity_active_identifier_claims_typed_record_idx`
- `entity_aliases.entity_aliases_typed_record_idx`
- `entity_preserved_identifiers.entity_preserved_identifiers_typed_record_idx`
- `evidence_access_handles.evidence_access_handles_object_blob_id_scope_idx`
- `evidence_custody_events.evidence_custody_events_evidence_record_id_scope_idx`
- `evidence_grid_projection.evidence_grid_projection_record_id_scope_idx`
- `evidence_object_upload_leases.evidence_object_upload_leases_object_blob_id_scope_idx`
- `evidence.evidence_incident_identity_key`
- `evidence.evidence_object_blob_id_scope_idx`
- `host_grid_projection.host_grid_projection_record_id_scope_idx`
- `hosts.hosts_incident_identity_key`
- `hosts.hosts_merged_into_record_id_scope_idx`
- `identities.identities_incident_identity_key`
- `identities.identities_merged_into_record_id_scope_idx`
- `identity_grid_projection.identity_grid_projection_record_id_scope_idx`
- `import_sessions.import_sessions_apply_job_id_scope_idx`
- `import_sessions.import_sessions_discovery_job_id_scope_idx`
- `import_sessions.import_sessions_incident_identity_key`
- `incident_bundle_exports.incident_bundle_exports_export_job_id_scope_idx`
- `incident_bundle_job_payloads.incident_bundle_job_payloads_job_id_scope_idx`
- `indicator_grid_projection.indicator_grid_projection_record_id_scope_idx`
- `jobs.jobs_incident_identity_key`
- `network_flow_graph_views.network_flow_graph_views_latest_job_id_scope_idx`
- `network_flow_rejected_row_diagnostics.network_flow_rejected_row_diagnostics_table_id_scope_idx`
- `network_flow_rows.network_flow_rows_network_flow_table_id_scope_idx`
- `network_flow_tables.network_flow_tables_incident_identity_key`
- `network_flow_tables.network_flow_tables_source_import_session_id_scope_idx`
- `object_blobs.object_blobs_incident_identity_key`
- `parties.parties_incident_identity_key`
- `party_grid_projection.party_grid_projection_record_id_scope_idx`
- `record_links.record_links_active_identity_idx`
- `records.records_incident_record_type_key`
- `report_composition_preview_attempts.report_composition_preview_attempts_composition_id_scope_idx`
- `report_composition_preview_attempts.report_composition_preview_attempts_render_attempt_id_scope_idx`
- `report_compositions.report_compositions_incident_identity_key`
- `reporting_job_payloads.reporting_job_payloads_job_id_scope_idx`
- `reporting_releases.reporting_releases_create_job_id_scope_idx`
- `reporting_releases.reporting_releases_snapshot_id_scope_idx`
- `reporting_snapshots.reporting_snapshots_create_job_id_scope_idx`
- `reporting_snapshots.reporting_snapshots_incident_identity_key`
- `task_request_grid_projection.task_request_grid_projection_record_id_scope_idx`
- `task_requests.task_requests_incident_identity_key`
- `timeline_events.timeline_events_incident_identity_key`
- `timeline_grid_projection.timeline_grid_projection_record_id_scope_idx`
- `timeline_grid_projection.timeline_grid_projection_replacement_record_id_scope_idx`

The source-capture NOT NULL constraint is now
`reporting_snapshots_source_capture_sha256_not_null`, replacing its historical
export-model name. The long preview-output and preview-output-file automatic names
exchange PostgreSQL's collision suffix according to final creation order; their
column ownership, NOT NULL enforcement and validation are unchanged.

The changed routines are `entities_active_identifier_claims_are_valid_v1` and
`indicator_active_identities_are_valid`. Their existing Recovery-only execution
capabilities, security modes and controlled search paths are retained. Recovery
calls the owners rather than reading their source records directly. No new
routine is exposed to runtime or PUBLIC.


## Verification and deployment handoff

All commands were public Make targets from the repository root, using task-owned
run IDs and harness-coordinated disposable fixtures. Roots below are relative to
`.cartulary/test-results/`; each completed graph run retains `run-summary.json`,
unit/row evidence and `cleanup-results.json`. Tool operations retain their named
`tool-run-summary.json`. No retained database, external deployment, user service,
shared volume or unrelated worktree was reset or edited.

| Command / scope | Run root | Observed result |
| --- | --- | --- |
| `make test-slice OWNER=module.database_migrations`, unchanged original catalog | `pg186-original-20261028` | PASS, 40.902s |
| Original endpoint capture using the recurrence row | `pg186-reference-capture-20261028` | PASS, 17.266s; actual 180006/checksums on |
| `make generate` | `pg186-generate-11` | PASS; final migration/ownership hashes, contract bindings, topology and SQLC |
| `make agent-finalize` | `pg186-finalize-06` | PASS, 22.969s; retained maintenance skipped because RESULTS_DIR unset |
| `make source-check CHECK=all` | `pg186-source-02` | PASS, 29.577s; final refresh recorded below |
| Full consolidated migration owner | `pg186-migrations-04` | PASS, 60.423s; all prefixes, admission, readiness, failure/cancellation, exact manifest and Up/Down/Up |
| Consolidated recurrence plus ledger-sequence denial | `pg186-migrations-05` | PASS, 29.935s |
| `make test-slice SUITE=backend-store` | `pg186-backend-store-01` | PASS, 35 units, 238.649s |
| `make test-slice SUITE=backend-unit` | `pg186-backend-unit-01` | 187/188 units passed; old Recovery binding digest assertion failed, corrected and rerun below |
| `make test-slice SUITE=backend-integration` | `pg186-backend-integration-01` | 133/135 units passed; import operation scope and deliberately corrupt reset fixture failed, corrected and rerun below |
| `make test-slice OWNER=module.parties SCOPE=services` | `pg186-parties-01` | PASS, 140.220s; claims/concurrency and owner behavior |
| Recovery binding projection row | `pg186-recovery-02` | PASS, 11.431s; expected digest derives from the authorized lineage change |
| Incident bundle round-trip row | `pg186-import-03` | PASS, 54.281s; historical Reporting artifacts and imported provenance preserved |
| Harness managed-service reset row | `pg186-reset-03` | PASS, 10.567s; source-valid fixture plus deliberately stale derived claims |
| Recovery reset/crypto identity and bypass-integrity row | `pg186-recovery-03` | PASS, 33.240s; later owner-validator extension rerun recorded below |
| `make harness-contract SUITE=contract` | `pg186-harness-01` | PASS, 118.112s; includes updated migration row selection |
| `make format` | `pg186-format-06` | PASS, 5.509s |

The first broad runs above remain recorded as failures; their focused reruns
closed the specific defects. The final full backend reruns recorded below pass
both requested suites on the same final source. Earlier outcomes are not
relabeled or treated as proof of unchanged source. Relevant passing behavior included
incident isolation, competing Entity/Party claims, Unicode normalization, link
conflict handling, append-only history/revisions, rollback allocation, purpose
roles and connection recycling, Reference Pack initialization, extension admission,
Reporting reconstruction and full Recovery flows. Final checks below identify the
latest source coverage explicitly.

### Investigation record

- Early generation failed on the SQL-standard RETURN routine boundary, the
  changed Recovery binding digest, then selector ordering/unregistered test
  metadata. These were corrected in authored inputs; generation passed. The
  structural policy now also rejects NOT ENFORCED and tests RETURN/grant separation.
- Initial consolidated fixture builds failed because the revision sequence was
  owned before its table existed. The preserved diagnostic in
  `.cartulary/pg186-consolidation/sql-diagnostic.txt` led to moving creation and
  ownership together into migration 08. No original reference SQL was repaired.
- Early finalization rejected stale topology and the old file-owner-only manifest
  rule. The validator now checks constraints/indexes/triggers against their owning
  table, separately from dependency-driven file allocation. Finalization passed.
- The first Entity service rerun found an unused import; the next found an incorrect
  replacement-test expectation for `Ⓐ`. The owner contract maps it to `ⓐ`; the
  assertion was corrected to preserve this behavior. Broad integration passed
  those Entity rows afterward.
- The broad unit failure was the old implementation-binding checksum in
  `vnext_graph_restore_artifacts_test.go`; the new expected digest is
  `3148f767425e0d16c264b377bac333d2729e7eb9007f0d1875a2bde85bf441c8`.
- The broad integration failures established the deployment-job exception and
  typed-claim fixture requirements described in D02/D03. A first reset-fixture
  correction attempted a Recovery INSERT, which was correctly denied; the final
  fixture uses a valid host and its migration capability without changing ACLs.
- Lint found four obsolete cutover helpers and a lexical limitation around a
  standalone parenthesis string in the catalog query. Helpers were removed and
  equivalent SQL formatting simplified. The next lint correctly rejected Recovery
  reading Records directly; Indicator support validation moved into its existing
  owner routine. The boundary policy was not weakened.
- Two cross-owner ROWS attempts and one mistyped row ID failed as usage errors
  before fixture setup; subsequent runs used exact owner selectors.

### Evidence, limits and deployment

Raw endpoint captures, original SQL hashes, per-object comparison and the task's
comparison script are retained under `.cartulary/pg186-consolidation/`. The final
comparison must report zero unexplained differences against D01–D10. Counts and
full changed object names are recorded above. Application-object ACLs, types,
defaults, collation choices, triggers, views, roles, extension versions and required
seed contents compare exactly outside the stated differences. SQLC models have no
remaining diff; only authored type overrides changed.

Elapsed run durations are recorded, but no isolated migration-only speed claim is
made. Test scope, fixture setup and parallel host admission differ between the
original and final whole-owner runs. This task adds no benchmark framework and
makes no latency/storage claim for the larger composite-key index set.

No browser visual inspection or golden update was needed: no presentation behavior
changed. Release packaging, formal Core 05 publication and deployment-specific
backup restoration were not run; this is repository implementation evidence, not a
release/conformance claim. No full warm-check root was supplied, so retained-run
maintenance remains explicitly skipped. Unrelated UI, vulnerability, release and
external deployment portfolios are not claimed by the focused backend gates.

For deployment, provision a fresh PostgreSQL 18.6 target using the existing
administrator-owned roles, certificate identities, pgcrypto/citext prerequisites
and checksums. Apply v3 from version 1 through the normal migration facade, then
start the matching binary and allow its existing Base/profile initialization.
Previous v1/v2/foreign databases cannot advance in place. If replacing an existing
disposable development target, the operator must select and inspect that exact
target, stop its writers and use the documented bounded reset command. This task
did not perform that deployment action. Any newly introduced requirement to retain
old database contents needs a separately adopted migration contract; it does not
block this authorized clean baseline implementation.

### Final source checks

| Command / scope | Run root | Result |
| --- | --- | --- |
| Full migration owner, final SQL | `pg186-migrations-06` | PASS, 59.739s |
| Recovery reset/crypto/bypass row, including orphan JSON support | `pg186-recovery-04` | PASS, 41.572s |
| `make test-slice OWNER=module.indicators SCOPE=services` | `pg186-indicators-02` | PASS, 36.183s |
| `make lint` | `pg186-lint-03` | PASS, 41.766s; includes staticcheck and module SQL boundaries |
| Semantic comparison of independent endpoints | `.cartulary/pg186-consolidation/semantic-comparison-result.json` | PASS; zero unexplained differences |

The original capture SHA-256 is
`d38456c919cf31391d53a247d60198ffd958dfdfc1091cd5aff4c0e982766bc8`.
The final capture and comparison digest are in the retained JSON result. The
comparison preserves the 24 physical-column changes in its report, checks the
exact 35 changed FK definitions/actions, supplemental nullable-scope FK, constraint
and index additions/removals, owner-validator changes and seed identity. All
required constraints are enforced/validated and all indexes are ready/valid.

`make source-check CHECK=all` passed in `pg186-source-03` (13.648s).
`make lint-markdown` passed in `pg186-markdown-01` (18.380s); a final prose-only
status reconciliation is checked by `pg186-markdown-02`. `git diff --check` passed.
All 1,368 recorded task cleanup actions inspected after those runs completed with
no cleanup failure. The later full backend reruns add 551 completed cleanup
actions, recorded separately below. No deployment action remains pending inside
this checkout.

The final schema capture SHA-256 is
`2c4fbda72eb093d60ab01aaf81870b1dc3c83c70bbf55a7e93887cb2bc22e11f`.
There is no unresolved repository implementation blocker. The explicitly skipped
release/deployment and retained-run obligations above remain outside the claimed
verification scope.

### Final full backend reruns — 2026-10-28

Both complete suites were rerun against the final 29-file v3 catalog after
`pg186-finalize-06`, with harness result caching disabled. Source remained frozen
through both invocations; no repair was needed. Exact commands from the repository
root:

```sh
CARTULARY_TEST_RUN_ID=pg186-backend-unit-02 make test-slice SUITE=backend-unit CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=pg186-backend-integration-02 make test-slice SUITE=backend-integration CARTULARY_HARNESS_CACHE_MODE=off
```

| Suite / run root under `.cartulary/test-results/` | Result | Duration | Cleanup |
| --- | --- | --- | --- |
| `backend-unit` / `pg186-backend-unit-02` | PASS; 188/188 work units; outer exit 0 | 344.573s | 197/197 actions completed |
| `backend-integration` / `pg186-backend-integration-02` | PASS; 135/135 work units; outer exit 0 | 448.500s | 354/354 actions completed |

Both manifests identify commit `4c22480888701a7b387074e10cd207b3e0828bc2`,
source state `dirty` (the uncommitted consolidation), and source digest
`sha256:96faac9eb4d76a20b22a3e89a6502fdb444506521ffd5d40ec4bd0fb2a2b9a3b`.
The summaries record zero failed, skipped or cancelled work units, zero cache
hits, and 188/135 cache bypasses respectively. Work units are harness execution
units, not individual test counts. Native command completion, each manifest,
`run-summary.json`, and `cleanup-results.json` were inspected; `make explain-run
RESULTS_DIR=<run root> DETAIL=summary` also confirmed each result. Neither run
reports a failure class or reason; cleanup has no failed or blocked action.

These runs close the requested full backend unit and integration verification
on the final catalog, including the previously failing Recovery binding,
incident-bundle import and managed-service reset coverage. They do not claim a
formal owner-union evidence audit or full warm-check acceptance. The existing
finalization remains applicable because only this handoff changed after these
reruns; retained-run maintenance remains skipped because `RESULTS_DIR` was unset
for finalization.

The prose-only result update is checked by `make lint-markdown` in
`pg186-markdown-03` and by `git diff --check`.

### Original object and dependency inventory

These names are an audit of original Up statements, not executable requirements.
Column shapes/defaults are characterized by the reference endpoint; intermediate
ALTER effects and their final disposition are identified above. Every named
constraint, index, routine, trigger and view declaration is inventoried below.
FK dependencies name referenced tables; function-body runtime dependencies remain
visible in the preserved original SQL and final qualified routines. Privilege
statements are summarized per original file; their resulting ACLs are compared
against the capture and access manifest. The original file hashes are recorded
for unambiguous identification. Superseded same-name definitions resolve to the
final endpoint, not a second compatibility definition.

#### 00001_database_infrastructure.sql

Original SHA-256: `96b3d6e07cdf6bc62bef5590edd97785fb3cbdb152be316b105c105c613c9d91`.

- Tables created/altered/dropped: `schema_migration_lineage`.
- Constraints: `schema_migration_lineage_pkey`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: insert lineage singleton.

#### 00002_auth.sql

Original SHA-256: `50e722b93a8c3f3234cf503f638b1d721a8a41e79253e53abf6f3a149facd761`.

- Tables created/altered/dropped: `account_preferences`, `bootstrap_tokens`, `enterprise_auth_bindings`, `enterprise_auth_providers`, `enterprise_auth_transactions`, `pending_totp_enrollments`, `route_idempotency`, `user_sessions`, `users`.
- Constraints: `account_preferences_density_mode_check`, `account_preferences_pkey`, `account_preferences_user_id_fkey`, `bootstrap_tokens_pkey`, `bootstrap_tokens_token_fingerprint_key`, `bootstrap_tokens_user_id_fkey`, `enterprise_auth_bindings_created_by_user_id_fkey`, `enterprise_auth_bindings_pkey`, `enterprise_auth_bindings_provider_id_fkey`, `enterprise_auth_bindings_provider_subject_check`, `enterprise_auth_bindings_provider_type_check`, `enterprise_auth_bindings_replaced_by_auth_binding_id_fkey`, `enterprise_auth_bindings_retired_by_user_id_fkey`, `enterprise_auth_bindings_user_id_fkey`, `enterprise_auth_providers_display_name_check`, `enterprise_auth_providers_pkey`, `enterprise_auth_providers_provider_key_check`, `enterprise_auth_providers_provider_key_key`, `enterprise_auth_providers_provider_type_check`, `enterprise_auth_transactions_correlation_ck`, `enterprise_auth_transactions_pkey`, `enterprise_auth_transactions_provider_id_fkey`, `enterprise_auth_transactions_provider_type_check`, `enterprise_auth_transactions_saml_staging_ck`, `pending_totp_enrollments_auth_scope_bootstrap_token_id_fkey`, `pending_totp_enrollments_auth_scope_ck`, `pending_totp_enrollments_auth_scope_kind_check`, `pending_totp_enrollments_auth_scope_session_id_fkey`, `pending_totp_enrollments_pkey`, `pending_totp_enrollments_user_id_fkey`, `route_idempotency_actor_user_id_fkey`, `route_idempotency_pkey`, `route_idempotency_target_user_id_fkey`, `user_sessions_auth_binding_id_fkey`, `user_sessions_pkey`, `user_sessions_provider_type_check`, `user_sessions_token_fingerprint_key`, `user_sessions_user_id_fkey`, `users_display_name_check`, `users_email_key`, `users_pkey`, `users_updated_by_user_id_fkey`.
- Indexes: `bootstrap_tokens_user_lookup_idx`, `enterprise_auth_bindings_active_provider_subject_idx`, `enterprise_auth_bindings_active_user_provider_idx`, `enterprise_auth_bindings_created_by_user_id_fk_idx`, `enterprise_auth_bindings_provider_id_fk_idx`, `enterprise_auth_bindings_replaced_by_auth_binding_id_fk_idx`, `enterprise_auth_bindings_retired_by_user_id_fk_idx`, `enterprise_auth_bindings_user_active_idx`, `enterprise_auth_bindings_user_id_fk_idx`, `enterprise_auth_providers_discovery_idx`, `enterprise_auth_transactions_provider_expiry_idx`, `enterprise_auth_transactions_relay_state_idx`, `enterprise_auth_transactions_saml_completion_hash_idx`, `enterprise_auth_transactions_state_idx`, `pending_totp_enrollments_auth_scope_bootstrap_token_id_fk_idx`, `pending_totp_enrollments_auth_scope_session_id_fk_idx`, `pending_totp_enrollments_one_current_per_user_idx`, `pending_totp_enrollments_scope_lookup_idx`, `pending_totp_enrollments_user_id_fk_idx`, `route_idempotency_actor_lookup_idx`, `route_idempotency_route_actor_scope_client_txn_idx`, `route_idempotency_target_user_id_fk_idx`, `user_sessions_auth_binding_id_fk_idx`, `user_sessions_user_lookup_idx`, `users_updated_by_user_id_fk_idx`.
- FK prerequisites: `bootstrap_tokens`, `enterprise_auth_bindings`, `enterprise_auth_providers`, `user_sessions`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00003_incidents.sql

Original SHA-256: `8e80e7694388bd44fd7c8c292007e9e71052d54c318626ef42463dadf14275e4`.

- Tables created/altered/dropped: `incident_memberships`, `incident_workbook_preferences`, `incidents`, `user_workbook_preferences`.
- Constraints: `incident_memberships_added_by_user_id_fkey`, `incident_memberships_incident_id_fkey`, `incident_memberships_pkey`, `incident_memberships_role_check`, `incident_memberships_updated_by_user_id_fkey`, `incident_memberships_user_id_fkey`, `incident_workbook_preferences_incident_id_fkey`, `incident_workbook_preferences_pkey`, `incident_workbook_preferences_updated_by_user_id_fkey`, `incidents_closed_at_status_ck`, `incidents_created_by_user_id_fkey`, `incidents_current_phase_contract_ck`, `incidents_description_contract_ck`, `incidents_incident_key_canonical_key`, `incidents_pkey`, `incidents_primary_external_case_ref_contract_ck`, `incidents_severity_contract_ck`, `incidents_status_ck`, `incidents_tlp_ck`, `incidents_updated_by_user_id_fkey`, `user_workbook_preferences_incident_id_fkey`, `user_workbook_preferences_pkey`, `user_workbook_preferences_user_id_fkey`.
- Indexes: `incident_memberships_added_by_user_id_fk_idx`, `incident_memberships_incident_lookup_idx`, `incident_memberships_updated_by_user_id_fk_idx`, `incident_memberships_user_lookup_idx`, `incident_workbook_preferences_updated_by_user_id_fk_idx`, `incidents_created_by_user_id_fk_idx`, `incidents_status_updated_lookup_idx`, `incidents_updated_by_user_id_fk_idx`, `incidents_updated_lookup_idx`, `user_workbook_preferences_user_id_fk_idx`.
- FK prerequisites: `incidents`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00004_recovery.sql

Original SHA-256: `ac73b18aa3823d9197b8d8fe08da3defced8652709dc2a3357a59d4d5b5eaa90`.

- Tables created/altered/dropped: `backup_sets`, `operator_recovery_journal`, `restore_verification_runs`.
- Constraints: `backup_sets_integrity_manifest_key_non_empty`, `backup_sets_integrity_manifest_sha256_check`, `backup_sets_integrity_manifest_size_check`, `backup_sets_last_verification_basis_sha256_check`, `backup_sets_object_store_anchor_retained_until_floor_check`, `backup_sets_object_store_artifact_key_non_empty`, `backup_sets_object_store_artifact_sha256_check`, `backup_sets_object_store_artifact_size_check`, `backup_sets_object_store_restore_anchor_non_empty`, `backup_sets_pkey`, `backup_sets_postgres_anchor_retained_until_floor_check`, `backup_sets_postgres_artifact_key_non_empty`, `backup_sets_postgres_artifact_sha256_check`, `backup_sets_postgres_artifact_size_check`, `backup_sets_postgres_restore_anchor_non_empty`, `backup_sets_retained_until_floor_check`, `backup_sets_verification_state_check`, `backup_sets_verification_timestamp_check`, `operator_recovery_journal_encryption_mode_check`, `operator_recovery_journal_envelope_schema_id_check`, `operator_recovery_journal_key_fingerprint_sha256_check`, `operator_recovery_journal_operation_check`, `operator_recovery_journal_payload_sha256_check`, `operator_recovery_journal_pkey`, `operator_recovery_journal_result_check`, `restore_verification_runs_backup_set_id_fkey`, `restore_verification_runs_basis_sha256_check`, `restore_verification_runs_completed_after_started_check`, `restore_verification_runs_failure_shape_check`, `restore_verification_runs_pkey`, `restore_verification_runs_state_check`.
- Indexes: `backup_sets_latest_retained_idx`, `operator_recovery_journal_backup_set_idx`, `operator_recovery_journal_operation_created_idx`, `restore_verification_runs_backup_set_started_idx`.
- FK prerequisites: `backup_sets`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00005_deployment_admin.sql

Original SHA-256: `1b07a3c9076a38e2de3704fc0cf1777f74c78bc169a64605333ef5f2729f9352`.

- Tables created/altered/dropped: `deployment_admin_audit_events`, `deployment_bootstrap_state`.
- Constraints: `deployment_admin_audit_events_actor_user_id_fkey`, `deployment_admin_audit_events_incident_id_fkey`, `deployment_admin_audit_events_pkey`, `deployment_admin_audit_events_target_user_id_fkey`, `deployment_bootstrap_state_bootstrap_artifact_id_key`, `deployment_bootstrap_state_bootstrap_schema_id_check`, `deployment_bootstrap_state_created_user_id_fkey`, `deployment_bootstrap_state_pkey`, `deployment_bootstrap_state_slot_check`.
- Indexes: `deployment_admin_audit_events_actor_user_id_fk_idx`, `deployment_admin_audit_events_incident_lookup_idx`, `deployment_admin_audit_events_target_user_id_fk_idx`, `deployment_bootstrap_state_created_user_id_fk_idx`.
- FK prerequisites: `incidents`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00006_platform_jobs.sql

Original SHA-256: `0d57f1ccb9717205289d10191cef3c439b9b102ca328a22787a723725558e6e0`.

- Tables created/altered/dropped: `jobs`.
- Constraints: `jobs_auth_policy_ck`, `jobs_definition_pair_ck`, `jobs_expiry_tombstone_ck`, `jobs_extension_ownership_ck`, `jobs_handler_attempt_eligibility_ck`, `jobs_handler_failure_count_ck`, `jobs_handler_live_attempt_ck`, `jobs_handler_name_nonempty_ck`, `jobs_handler_retry_state_ck`, `jobs_handler_terminal_attempt_ck`, `jobs_incident_id_fkey`, `jobs_job_kind_nonempty_ck`, `jobs_nonterminal_definition_ck`, `jobs_pkey`, `jobs_progress_completed_check`, `jobs_progress_total_check`, `jobs_progress_total_ck`, `jobs_progress_unit_id_shape_ck`, `jobs_scope_incident_ck`, `jobs_scope_kind_check`, `jobs_status_check`, `jobs_submitted_by_user_id_fkey`, `jobs_succeeded_progress_ck`, `jobs_terminal_cancelable_ck`, `jobs_terminal_summary_ck`.
- Indexes: `jobs_expiry_candidates_idx`, `jobs_extension_nonterminal_idx`, `jobs_handler_recovery_idx`, `jobs_incident_id_fk_idx`, `jobs_incident_lookup_idx`, `jobs_submitted_by_lookup_idx`.
- FK prerequisites: `incidents`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00007_records.sql

Original SHA-256: `4f88d17cc38eb6986f28cecd38b8b2936a83c4c6f52c4115d42bb8e627e2a5a1`.

- Tables created/altered/dropped: `change_set_mutations`, `change_sets`, `record_history_entry_refs`, `record_revisions`, `records`.
- Constraints: `change_set_mutations_change_set_id_fkey`, `change_set_mutations_pkey`, `change_sets_actor_user_id_fkey`, `change_sets_incident_id_fkey`, `change_sets_pkey`, `record_history_entry_refs_change_set_id_fkey`, `record_history_entry_refs_change_set_id_mutation_sequence__fkey`, `record_history_entry_refs_pkey`, `record_history_entry_refs_record_id_change_set_id_mutation__key`, `record_history_entry_refs_record_id_fkey`, `record_revisions_change_set_id_fkey`, `record_revisions_pkey`, `record_revisions_record_id_fkey`, `record_revisions_record_id_row_version_key`, `records_created_by_user_id_fkey`, `records_delete_state_ck`, `records_deleted_by_user_id_fkey`, `records_incident_id_fkey`, `records_incident_id_record_id_key`, `records_pkey`, `records_record_type_check`, `records_row_version_check`, `records_updated_by_user_id_fkey`.
- Indexes: `change_set_mutations_target_lookup_idx`, `change_sets_actor_user_id_fk_idx`, `change_sets_incident_lookup_idx`, `record_history_entry_refs_change_set_id_fk_idx`, `record_history_entry_refs_change_set_id_mutation_seque_824b16fa`, `record_history_entry_refs_record_lookup_idx`, `record_revisions_change_set_id_fk_idx`, `record_revisions_record_lookup_idx`, `records_active_incident_type_idx`, `records_created_by_user_id_fk_idx`, `records_deleted_by_user_id_fk_idx`, `records_incident_lookup_idx`, `records_updated_by_user_id_fk_idx`.
- Views/sequences: `record_revisions_revision_id_seq`.
- FK prerequisites: `change_set_mutations`, `change_sets`, `incidents`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00008_revisions.sql

Original SHA-256: `df85d792aadf236a4da8854a84af78470198e733b367126aa3fa66025fae1c70`.

- Tables created/altered/dropped: `change_set_mutations`, `record_revision_conflict_facts`.
- Constraints: `change_set_mutations_history_entry_record_ids_ck`, `change_set_mutations_history_entry_subset_ck`, `change_set_mutations_history_record_ids_ck`, `record_revision_conflict_facts_after_value_ck`, `record_revision_conflict_facts_before_value_ck`, `record_revision_conflict_facts_field_key_ck`, `record_revision_conflict_facts_pkey`, `record_revision_conflict_facts_presence_ck`, `record_revision_conflict_facts_revision_id_fkey`.
- Indexes: `change_set_mutations_history_record_ids_idx`.
- Routines: `change_set_mutations_history_ids_are_canonical`, `revisions_incident_bundle_sequence_begin_v1`, `revisions_incident_bundle_sequence_finish_v1`.
- Views/sequences: `record_revisions_revision_id_seq`.
- FK prerequisites: `record_revisions`.
- Privilege operations: 5; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00009_parties.sql

Original SHA-256: `0fc83fc76b791c3f1baba1fec71a66d4805349f9ad63db87145c13f321658f7c`.

- Tables created/altered/dropped: `parties`.
- Constraints: `parties_incident_id_fkey`, `parties_pkey`, `parties_record_envelope_fkey`.
- Indexes: `parties_incident_display_name_idx`, `parties_incident_id_record_id_fk_idx`.
- FK prerequisites: `incidents`, `records`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00010_timeline.sql

Original SHA-256: `e77046cdc78bc1badd8f8188be9e1ee69c68d98c12cdb03eb87cfd226bff13e5`.

- Tables created/altered/dropped: `timeline_events`, `timeline_source_provenance`, `timeline_time_conversion_profiles`.
- Constraints: `timeline_events_activity_time_pair_state_check`, `timeline_events_capture_state_check`, `timeline_events_created_by_user_id_fkey`, `timeline_events_incident_id_fkey`, `timeline_events_pkey`, `timeline_events_record_envelope_fkey`, `timeline_events_reviewed_by_user_id_fkey`, `timeline_events_superseded_by_user_id_fkey`, `timeline_events_updated_by_user_id_fkey`, `timeline_source_provenance_cell_kind_ck`, `timeline_source_provenance_header_ck`, `timeline_source_provenance_identity_hash_ck`, `timeline_source_provenance_metadata_ck`, `timeline_source_provenance_ordinals_ck`, `timeline_source_provenance_pkey`, `timeline_source_provenance_raw_value_ck`, `timeline_source_provenance_record_id_fkey`, `timeline_source_provenance_source_kind_ck`, `timeline_time_conversion_profiles_enabled_offset_ck`, `timeline_time_conversion_profiles_incident_id_fkey`, `timeline_time_conversion_profiles_local_offset_minutes_check`, `timeline_time_conversion_profiles_pkey`, `timeline_time_conversion_profiles_updated_by_user_id_fkey`.
- Indexes: `idx_timeline_source_provenance_record`, `timeline_events_created_by_user_id_fk_idx`, `timeline_events_incident_id_record_id_fk_idx`, `timeline_events_incident_lookup_idx`, `timeline_events_reviewed_by_user_id_fk_idx`, `timeline_events_superseded_by_user_id_fk_idx`, `timeline_events_updated_by_user_id_fk_idx`, `timeline_time_conversion_profiles_updated_by_user_id_fk_idx`.
- FK prerequisites: `incidents`, `records`, `timeline_events`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00011_entities.sql

Original SHA-256: `3a8b71932fd0cddb9c9640f565019f1790b0dfa2f1b07148aef588a7442ee90d`.

- Tables created/altered/dropped: `entity_aliases`, `entity_mentions`, `entity_preserved_identifiers`, `hosts`, `identities`.
- Constraints: `entity_aliases_alias_text_controls_ck`, `entity_aliases_alias_text_length_ck`, `entity_aliases_alias_text_nonempty_ck`, `entity_aliases_alias_text_storage_ck`, `entity_aliases_classification_check`, `entity_aliases_created_by_user_id_fkey`, `entity_aliases_entity_type_check`, `entity_aliases_incident_id_fkey`, `entity_aliases_pkey`, `entity_mentions_created_by_user_id_fkey`, `entity_mentions_entity_type_check`, `entity_mentions_ordinal_check`, `entity_mentions_pkey`, `entity_mentions_resolution_status_check`, `entity_mentions_resolved_by_user_id_fkey`, `entity_mentions_source_record_envelope_fkey`, `entity_preserved_identifiers_classification_check`, `entity_preserved_identifiers_created_by_user_id_fkey`, `entity_preserved_identifiers_entity_type_check`, `entity_preserved_identifiers_incident_id_fkey`, `entity_preserved_identifiers_pkey`, `hosts_created_by_user_id_fkey`, `hosts_entity_origin_core02_ck`, `hosts_host_state_check`, `hosts_incident_id_fkey`, `hosts_merge_lineage_ck`, `hosts_merged_into_record_id_fkey`, `hosts_pkey`, `hosts_record_envelope_fkey`, `hosts_seed_entity_mention_id_fkey`, `hosts_updated_by_user_id_fkey`, `identities_created_by_user_id_fkey`, `identities_entity_origin_core02_ck`, `identities_identity_state_check`, `identities_incident_id_fkey`, `identities_merge_lineage_ck`, `identities_merged_into_record_id_fkey`, `identities_pkey`, `identities_record_envelope_fkey`, `identities_seed_entity_mention_id_fkey`, `identities_updated_by_user_id_fkey`.
- Indexes: `entity_aliases_created_by_user_id_fk_idx`, `entity_aliases_incident_id_fk_idx`, `entity_aliases_lookup_idx`, `entity_aliases_record_unique_idx`, `entity_mentions_created_by_user_id_fk_idx`, `entity_mentions_resolved_by_user_id_fk_idx`, `entity_mentions_source_lookup_idx`, `entity_mentions_unresolved_lookup_idx`, `entity_preserved_identifiers_created_by_user_id_fk_idx`, `entity_preserved_identifiers_exact_lookup_idx`, `entity_preserved_identifiers_incident_id_fk_idx`, `entity_preserved_identifiers_record_unique_idx`, `hosts_created_by_user_id_fk_idx`, `hosts_incident_aad_device_id_idx`, `hosts_incident_display_name_idx`, `hosts_incident_fqdn_idx`, `hosts_incident_hostname_idx`, `hosts_incident_id_record_id_fk_idx`, `hosts_incident_merged_into_idx`, `hosts_merged_into_record_id_fk_idx`, `hosts_seed_entity_mention_id_fk_idx`, `hosts_updated_by_user_id_fk_idx`, `identities_created_by_user_id_fk_idx`, `identities_incident_aad_object_id_idx`, `identities_incident_display_name_idx`, `identities_incident_email_idx`, `identities_incident_id_record_id_fk_idx`, `identities_incident_merged_into_idx`, `identities_incident_sam_account_name_idx`, `identities_incident_sid_idx`, `identities_incident_upn_idx`, `identities_merged_into_record_id_fk_idx`, `identities_seed_entity_mention_id_fk_idx`, `identities_updated_by_user_id_fk_idx`.
- FK prerequisites: `entity_mentions`, `hosts`, `identities`, `incidents`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00012_indicators.sql

Original SHA-256: `48705b34b9c5a2a6e61614a88fbddff9edff6e540dbec5ac40691cccd65166aa`.

- Tables created/altered/dropped: `indicator_active_identities`, `indicator_observations`, `indicator_state_intervals`, `indicators`, `records`.
- Constraints: `indicator_active_identities_dedupe_ck`, `indicator_active_identities_exact_dedupe_ck`, `indicator_active_identities_exact_type_ck`, `indicator_active_identities_indicator_fkey`, `indicator_active_identities_pkey`, `indicator_active_identities_record_envelope_fkey`, `indicator_active_identities_record_key`, `indicator_active_identities_type_ck`, `indicator_observations_candidate_pair_ck`, `indicator_observations_created_by_user_id_fkey`, `indicator_observations_deleted_by_user_id_fkey`, `indicator_observations_incident_id_fkey`, `indicator_observations_origin_kind_ck`, `indicator_observations_parsed_type_ck`, `indicator_observations_pkey`, `indicator_observations_required_text_ck`, `indicator_observations_resolution_status_check`, `indicator_observations_resolution_tuple_ck`, `indicator_observations_resolved_by_user_id_fkey`, `indicator_observations_resolved_indicator_incident_fkey`, `indicator_observations_row_version_ck`, `indicator_observations_source_record_envelope_fkey`, `indicator_observations_tombstone_pair_ck`, `indicator_state_intervals_confidence_check`, `indicator_state_intervals_created_by_user_id_fkey`, `indicator_state_intervals_deleted_by_user_id_fkey`, `indicator_state_intervals_incident_id_fkey`, `indicator_state_intervals_indicator_incident_fkey`, `indicator_state_intervals_lifecycle_state_ck`, `indicator_state_intervals_pkey`, `indicator_state_intervals_required_text_ck`, `indicator_state_intervals_row_version_ck`, `indicator_state_intervals_support_refs_ck`, `indicator_state_intervals_tombstone_pair_ck`, `indicator_state_intervals_validity_ck`, `indicators_dedupe_key_ck`, `indicators_hash_pair_ck`, `indicators_hash_value_ck`, `indicators_incident_id_fkey`, `indicators_incident_record_key`, `indicators_indicator_type_ck`, `indicators_ip_value_kind_ck`, `indicators_pkey`, `indicators_record_envelope_fkey`, `indicators_required_text_ck`, `indicators_value_kind_check`.
- Indexes: `indicator_active_identities_incident_id_indicator_reco_f4132796`, `indicator_observations_candidate_lookup_idx`, `indicator_observations_created_by_user_id_fk_idx`, `indicator_observations_deleted_by_user_id_fk_idx`, `indicator_observations_incident_id_resolved_indicator__f4dab861`, `indicator_observations_incident_id_source_record_id_fk_idx`, `indicator_observations_resolved_by_user_id_fk_idx`, `indicator_observations_resolved_lookup_idx`, `indicator_observations_source_lookup_idx`, `indicator_state_intervals_created_by_user_id_fk_idx`, `indicator_state_intervals_deleted_by_user_id_fk_idx`, `indicator_state_intervals_incident_id_fk_idx`, `indicator_state_intervals_indicator_incident_fk_idx`, `indicator_state_intervals_indicator_lookup_idx`, `indicators_incident_normalized_lookup_idx`.
- Routines: `enforce_indicator_support_ref_incident`, `indicator_active_identities_are_valid`, `indicator_support_refs_are_valid`, `rebuild_indicator_active_identities`, `sync_indicator_active_identity_from_indicator`, `sync_indicator_active_identity_from_record`.
- Triggers: `indicator_state_intervals_support_ref_incident`, `indicators_sync_active_identity`, `records_sync_indicator_active_identity`.
- FK prerequisites: `incidents`, `indicators`, `records`, `users`.
- Privilege operations: 6; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00013_assessments.sql

Original SHA-256: `be5596cc4bf00ddc6472931ba010110384525cf5155f71e99ee72a7352058cb9`.

- Tables created/altered/dropped: `assessments`.
- Constraints: `assessments_assessment_state_ck`, `assessments_confidence_score_ck`, `assessments_record_envelope_fkey`, `assessments_subject_type_ck`, `compromise_assessments_assessed_by_user_id_fkey`, `compromise_assessments_deleted_by_user_id_fkey`, `compromise_assessments_incident_id_fkey`, `compromise_assessments_pkey`.
- Indexes: `assessments_active_subject_lookup_idx`, `assessments_assessor_user_id_fk_idx`, `assessments_deleted_by_user_id_fk_idx`, `assessments_incident_id_record_id_fk_idx`.
- FK prerequisites: `incidents`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00014_links.sql

Original SHA-256: `e6d619e3b3c2fa45e97eb2f7a08ed632855478a7af8d7a05df8913b4e4c7ef7f`.

- Tables created/altered/dropped: `record_links`, `record_tags`.
- Constraints: `record_links_confidence_check`, `record_links_created_by_user_id_fkey`, `record_links_deleted_by_user_id_fkey`, `record_links_distinct_endpoints_ck`, `record_links_dst_record_envelope_fkey`, `record_links_incident_id_fkey`, `record_links_link_type_check`, `record_links_owner_user_id_fkey`, `record_links_pkey`, `record_links_provenance_check`, `record_links_src_record_envelope_fkey`, `record_tags_created_by_user_id_fkey`, `record_tags_deleted_by_user_id_fkey`, `record_tags_incident_id_fkey`, `record_tags_pkey`, `record_tags_record_envelope_fkey`.
- Indexes: `record_links_active_party_ref_dst_lookup_idx`, `record_links_active_src_lookup_idx`, `record_links_active_timeline_supersedes_dst_idx`, `record_links_active_unique_field_idx`, `record_links_active_unique_no_field_idx`, `record_links_created_by_user_id_fk_idx`, `record_links_deleted_by_user_id_fk_idx`, `record_links_incident_dst_record_fk_idx`, `record_links_incident_src_record_fk_idx`, `record_links_owner_user_id_fk_idx`, `record_tags_active_record_lookup_idx`, `record_tags_active_unique_idx`, `record_tags_created_by_user_id_fk_idx`, `record_tags_deleted_by_user_id_fk_idx`, `record_tags_incident_record_fk_idx`.
- Views/sequences: `active_record_links_v1`, `active_record_tags_v1`.
- FK prerequisites: `incidents`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00015_tasks_and_decisions.sql

Original SHA-256: `7e5843ba620a1032943f2a1d2f3616aecf1e0f46f42e2a52416ed0d7d68e650f`.

- Tables created/altered/dropped: `decisions`, `task_requests`.
- Constraints: `decisions_incident_id_fkey`, `decisions_owner_user_id_fkey`, `decisions_pkey`, `decisions_record_envelope_fkey`, `task_requests_incident_id_fkey`, `task_requests_owner_user_id_fkey`, `task_requests_pkey`, `task_requests_record_envelope_fkey`.
- Indexes: `decisions_incident_decided_idx`, `decisions_incident_id_record_id_fk_idx`, `decisions_owner_user_id_fk_idx`, `task_requests_incident_id_record_id_fk_idx`, `task_requests_incident_updated_idx`, `task_requests_owner_user_id_fk_idx`, `task_requests_requester_party_ref_lookup_idx`.
- FK prerequisites: `incidents`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00016_artifacts.sql

Original SHA-256: `2ec5c6f068d3c2d1c71accf1e411e50b73a28e1ef0d4c8e5da575a6af9863641`.

- Tables created/altered/dropped: `artifact_findings`, `artifact_forensic_keywords`, `artifact_investigative_queries`, `artifacts`, `handoff_risk_refs`.
- Constraints: `artifact_findings_confidence_score_check`, `artifact_findings_incident_id_fkey`, `artifact_findings_kind_check`, `artifact_findings_owner_user_id_fkey`, `artifact_findings_pkey`, `artifact_findings_record_envelope_fkey`, `artifact_findings_record_id_fkey`, `artifact_findings_state_check`, `artifact_forensic_keywords_incident_id_fkey`, `artifact_forensic_keywords_keyword_id_unique`, `artifact_forensic_keywords_match_mode_check`, `artifact_forensic_keywords_pkey`, `artifact_forensic_keywords_record_envelope_fkey`, `artifact_forensic_keywords_record_id_fkey`, `artifact_investigative_queries_created_by_user_id_fkey`, `artifact_investigative_queries_incident_id_fkey`, `artifact_investigative_queries_pkey`, `artifact_investigative_queries_query_id_unique`, `artifact_investigative_queries_record_envelope_fkey`, `artifact_investigative_queries_record_id_fkey`, `artifacts_artifact_type_check`, `artifacts_created_by_user_id_fkey`, `artifacts_incident_id_fkey`, `artifacts_incoming_owner_user_id_fkey`, `artifacts_outgoing_owner_user_id_fkey`, `artifacts_owner_user_id_fkey`, `artifacts_pkey`, `artifacts_record_envelope_fkey`, `artifacts_review_owner_user_id_fkey`, `handoff_risk_refs_created_by_user_id_fkey`, `handoff_risk_refs_delete_state_ck`, `handoff_risk_refs_deleted_by_user_id_fkey`, `handoff_risk_refs_handoff_fkey`, `handoff_risk_refs_incident_id_fkey`, `handoff_risk_refs_pkey`.
- Indexes: `artifact_findings_incident_closed_idx`, `artifact_findings_incident_confidence_idx`, `artifact_findings_incident_id_record_id_fk_idx`, `artifact_findings_incident_state_idx`, `artifact_findings_owner_user_id_fk_idx`, `artifact_forensic_keywords_incident_created_idx`, `artifact_forensic_keywords_incident_id_record_id_fk_idx`, `artifact_forensic_keywords_incident_mode_idx`, `artifact_investigative_queries_created_by_user_id_fk_idx`, `artifact_investigative_queries_incident_created_idx`, `artifact_investigative_queries_incident_id_record_id_fk_idx`, `artifact_investigative_queries_incident_platform_idx`, `artifacts_created_by_user_id_fk_idx`, `artifacts_incident_id_record_id_fk_idx`, `artifacts_incident_type_updated_idx`, `artifacts_incoming_owner_user_id_fk_idx`, `artifacts_outgoing_owner_user_id_fk_idx`, `artifacts_owner_user_id_fk_idx`, `artifacts_review_owner_user_id_fk_idx`, `handoff_risk_refs_active_text_idx`, `handoff_risk_refs_created_by_user_id_fk_idx`, `handoff_risk_refs_deleted_by_user_id_fk_idx`, `handoff_risk_refs_incident_id_fk_idx`, `handoff_risk_refs_incident_id_handoff_record_id_fk_idx`.
- Routines: `cartulary_confidence_band`.
- FK prerequisites: `artifacts`, `incidents`, `records`, `users`.
- Privilege operations: 1; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00017_evidence.sql

Original SHA-256: `d7a89168aec3dd652b03701ef1a1a4d2edf200801c6c6be27498085effd96806`.

- Tables created/altered/dropped: `evidence`, `evidence_access_handles`, `evidence_custody_events`, `object_blobs`.
- Constraints: `evidence_access_handles_disposition_check`, `evidence_access_handles_evidence_fkey`, `evidence_access_handles_handle_kind_check`, `evidence_access_handles_incident_id_fkey`, `evidence_access_handles_issued_by_user_id_fkey`, `evidence_access_handles_object_blob_id_fkey`, `evidence_access_handles_pkey`, `evidence_access_handles_preview_kind_ck`, `evidence_access_handles_size_bytes_check`, `evidence_custody_events_actor_user_id_fkey`, `evidence_custody_events_custody_event_type_check`, `evidence_custody_events_evidence_record_id_fkey`, `evidence_custody_events_incident_id_fkey`, `evidence_custody_events_pkey`, `evidence_incident_id_fkey`, `evidence_object_blob_fkey`, `evidence_pkey`, `evidence_record_envelope_fkey`, `object_blobs_byte_size_check`, `object_blobs_created_by_user_id_fkey`, `object_blobs_expected_sha256_hex_check`, `object_blobs_finalize_attempt_count_check`, `object_blobs_incident_id_fkey`, `object_blobs_observed_sha256_hex_check`, `object_blobs_pkey`, `object_blobs_storage_key_key`, `object_blobs_terminal_reason_check`, `object_blobs_terminal_state_ck`, `object_blobs_upload_state_check`.
- Indexes: `evidence_access_handles_issued_by_user_id_fk_idx`, `evidence_access_handles_lookup_idx`, `evidence_access_handles_object_blob_id_fk_idx`, `evidence_collector_party_ref_lookup_idx`, `evidence_custody_events_actor_user_id_fk_idx`, `evidence_custody_events_incident_id_fk_idx`, `evidence_custody_events_record_time_idx`, `evidence_incident_record_unique_idx`, `evidence_incident_sort_idx`, `evidence_object_blob_id_fk_idx`, `evidence_object_blob_unique_idx`, `evidence_source_party_ref_lookup_idx`, `object_blobs_created_by_user_id_fk_idx`, `object_blobs_incident_state_idx`.
- FK prerequisites: `evidence`, `incidents`, `object_blobs`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00018_saved_views.sql

Original SHA-256: `437b0ff192184590d6c8dfab9cac285db35e8ca1d6324854196ade659b701d8e`.

- Tables created/altered/dropped: `saved_views`.
- Constraints: `saved_views_incident_id_fkey`, `saved_views_owner_scope_ck`, `saved_views_owner_user_id_fkey`, `saved_views_pkey`, `saved_views_scope_check`, `saved_views_timestamp_order_ck`, `saved_views_version_positive_ck`.
- Indexes: `saved_views_incident_order_idx`, `saved_views_owner_lookup_idx`, `saved_views_owner_user_id_fk_idx`.
- FK prerequisites: `incidents`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00019_imports.sql

Original SHA-256: `ec69d5102517aa68df9e6fc15cd1ff533ff2f64203e29edbf88d883b20326449`.

- Tables created/altered/dropped: `import_apply_journal`, `import_apply_unit_plans`, `import_sessions`, `import_source_streams`, `import_unit_apply_outcomes`, `import_units`.
- Constraints: `import_apply_journal_change_set_id_fkey`, `import_apply_journal_import_session_id_fkey`, `import_apply_journal_import_unit_id_fkey`, `import_apply_journal_import_unit_id_mapping_fingerprint_sou_key`, `import_apply_journal_pkey`, `import_apply_journal_record_id_fkey`, `import_apply_journal_row_version_check`, `import_apply_journal_source_row_ref_check`, `import_apply_unit_plans_digest_ck`, `import_apply_unit_plans_job_fk`, `import_apply_unit_plans_mapping_json_ck`, `import_apply_unit_plans_pk`, `import_apply_unit_plans_sequence_ck`, `import_apply_unit_plans_session_fk`, `import_apply_unit_plans_session_sequence_uq`, `import_apply_unit_plans_target_kind_ck`, `import_apply_unit_plans_target_shape_ck`, `import_apply_unit_plans_unit_fk`, `import_apply_unit_plans_unit_uq`, `import_sessions_apply_job_id_fkey`, `import_sessions_created_by_user_id_fkey`, `import_sessions_discovery_job_id_fkey`, `import_sessions_incident_id_fkey`, `import_sessions_pkey`, `import_sessions_session_status_check`, `import_sessions_source_byte_size_check`, `import_sessions_source_file_kind_check`, `import_source_streams_import_session_id_fkey`, `import_source_streams_import_unit_id_fkey`, `import_source_streams_import_unit_id_key`, `import_source_streams_pkey`, `import_source_streams_ref_ck`, `import_source_streams_sha_ck`, `import_source_streams_source_byte_size_ck`, `import_unit_apply_outcomes_actor_fk`, `import_unit_apply_outcomes_change_set_fk`, `import_unit_apply_outcomes_commit_id_ck`, `import_unit_apply_outcomes_commit_uq`, `import_unit_apply_outcomes_error_details_ck`, `import_unit_apply_outcomes_error_shape_ck`, `import_unit_apply_outcomes_job_fk`, `import_unit_apply_outcomes_mapping_digest_ck`, `import_unit_apply_outcomes_owner_result_ck`, `import_unit_apply_outcomes_pkey`, `import_unit_apply_outcomes_plan_fk`, `import_unit_apply_outcomes_resource_refs_ck`, `import_unit_apply_outcomes_sequence_ck`, `import_unit_apply_outcomes_session_fk`, `import_unit_apply_outcomes_session_sequence_uq`, `import_unit_apply_outcomes_source_digest_ck`, `import_unit_apply_outcomes_status_ck`, `import_unit_apply_outcomes_target_kind_ck`, `import_unit_apply_outcomes_target_shape_ck`, `import_unit_apply_outcomes_unit_fk`, `import_unit_apply_outcomes_unit_uq`, `import_units_approved_target_kind_ck`, `import_units_approved_target_shape_ck`, `import_units_base_import_unit_id_fkey`, `import_units_blocking_source_columns_ck`, `import_units_data_start_row_ref_check`, `import_units_discovery_sequence_ck`, `import_units_header_row_ref_check`, `import_units_import_session_id_fkey`, `import_units_inferred_column_count_check`, `import_units_inferred_row_count_check`, `import_units_operator_region_shape_ck`, `import_units_pkey`, `import_units_session_discovery_sequence_uq`, `import_units_source_stream_ref_ck`, `import_units_unit_status_check`.
- Indexes: `import_apply_journal_change_set_id_fk_idx`, `import_apply_journal_record_idx`, `import_apply_journal_session_unit_idx`, `import_apply_unit_plans_apply_job_id_fk_idx`, `import_sessions_apply_job_id_fk_idx`, `import_sessions_created_by_lookup_idx`, `import_sessions_discovery_job_id_fk_idx`, `import_sessions_incident_lookup_idx`, `import_source_streams_session_idx`, `import_unit_apply_outcomes_actor_user_id_fk_idx`, `import_unit_apply_outcomes_apply_job_id_fk_idx`, `import_unit_apply_outcomes_change_set_id_fk_idx`, `import_unit_apply_outcomes_finalize_idx`, `import_unit_apply_outcomes_import_session_id_import_un_71ce6c33`, `import_units_base_import_unit_id_fk_idx`, `import_units_operator_region_identity_uq`, `import_units_operator_region_sequence_uq`, `import_units_session_lookup_idx`, `import_units_source_stream_ref_idx`.
- FK prerequisites: `change_sets`, `import_apply_unit_plans`, `import_sessions`, `import_units`, `incidents`, `jobs`, `records`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00020_network_flow.sql

Original SHA-256: `9112e4f5fd69e5ab9a04c0355d1158fac48f971fd3a99f38d8fcca0e8f19b82f`.

- Tables created/altered/dropped: `network_flow_indicator_bindings`, `network_flow_rejected_row_diagnostics`, `network_flow_rows`, `network_flow_tables`.
- Constraints: `network_flow_indicator_bindings_candidate_ck`, `network_flow_indicator_bindings_created_by_user_id_fkey`, `network_flow_indicator_bindings_id_ck`, `network_flow_indicator_bindings_incident_id_fkey`, `network_flow_indicator_bindings_pkey`, `network_flow_indicator_bindings_selector_kind_ck`, `network_flow_indicator_bindings_source_count_ck`, `network_flow_indicator_bindings_source_refs_ck`, `network_flow_indicator_bindings_target_ck`, `network_flow_indicator_bindings_target_indicator_record_id_fkey`, `network_flow_rejected_row_diagnostics_error_code_ck`, `network_flow_rejected_row_diagnostics_id_ck`, `network_flow_rejected_row_diagnostics_incident_id_fkey`, `network_flow_rejected_row_diagnostics_limit_ck`, `network_flow_rejected_row_diagnostics_message_args_ck`, `network_flow_rejected_row_diagnostics_message_ck`, `network_flow_rejected_row_diagnostics_message_key_ck`, `network_flow_rejected_row_diagnostics_pkey`, `network_flow_rejected_row_diagnostics_reason_code_ck`, `network_flow_rejected_row_diagnostics_sample_ck`, `network_flow_rejected_row_diagnostics_sha_ck`, `network_flow_rejected_row_diagnostics_source_column_ck`, `network_flow_rejected_row_diagnostics_source_row_ck`, `network_flow_rejected_row_diagnostics_table_id_fkey`, `network_flow_rows_created_by_user_id_fkey`, `network_flow_rows_id_ck`, `network_flow_rows_incident_id_fkey`, `network_flow_rows_ip_ck`, `network_flow_rows_nullable_text_ck`, `network_flow_rows_observation_source_ref_object_ck`, `network_flow_rows_pkey`, `network_flow_rows_ports_ck`, `network_flow_rows_positive_row_ck`, `network_flow_rows_protocol_flags_ck`, `network_flow_rows_sha_ck`, `network_flow_rows_table_id_fkey`, `network_flow_rows_table_source_row_key`, `network_flow_rows_time_order_ck`, `network_flow_rows_uint64_text_ck`, `network_flow_rows_unmapped_raw_object_ck`, `network_flow_tables_counts_ck`, `network_flow_tables_created_by_user_id_fkey`, `network_flow_tables_display_name_ck`, `network_flow_tables_id_ck`, `network_flow_tables_incident_id_fkey`, `network_flow_tables_pkey`, `network_flow_tables_profile_ids_ck`, `network_flow_tables_safe_key_id_ck`, `network_flow_tables_sha_ck`, `network_flow_tables_source_filename_display_ck`, `network_flow_tables_source_import_session_id_fkey`, `network_flow_tables_source_import_unit_id_fkey`, `network_flow_tables_source_import_unit_id_key`, `network_flow_tables_status_ck`, `network_flow_tables_status_deleted_at_ck`, `network_flow_tables_timestamps_ck`, `network_flow_tables_version_ck`.
- Indexes: `network_flow_indicator_bindings_created_by_user_id_fk_idx`, `network_flow_indicator_bindings_identity_uidx`, `network_flow_indicator_bindings_incident_created_idx`, `network_flow_indicator_bindings_target_indicator_recor_e8bba056`, `network_flow_rejected_diagnostics_keyset_idx`, `network_flow_rejected_row_diagnostics_incident_id_fk_idx`, `network_flow_rows_created_by_user_id_fk_idx`, `network_flow_rows_incident_default_keyset_idx`, `network_flow_rows_table_default_keyset_idx`, `network_flow_tables_active_display_name_uidx`, `network_flow_tables_active_order_idx`, `network_flow_tables_created_by_user_id_fk_idx`, `network_flow_tables_retained_count_idx`, `network_flow_tables_source_import_session_id_fk_idx`.
- Routines: `network_flow_reject_immutable_update`.
- Triggers: `network_flow_rejected_row_diagnostics_immutable_update`, `network_flow_rows_immutable_update`.
- FK prerequisites: `import_sessions`, `import_units`, `incidents`, `indicators`, `network_flow_tables`, `users`.
- Privilege operations: 1; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00021_projections.sql

Original SHA-256: `2605c462b3283e0325b68143f01a58e0eb26f96974a0266ce44b760468027ed5`.

- Tables created/altered/dropped: `artifact_grid_projection`, `assessment_grid_projection`, `decision_grid_projection`, `evidence_grid_projection`, `host_grid_projection`, `identity_grid_projection`, `indicator_grid_projection`, `party_grid_projection`, `task_request_grid_projection`, `timeline_grid_projection`.
- Constraints: `artifact_grid_projection_created_by_user_id_fkey`, `artifact_grid_projection_finding_owner_user_id_fkey`, `artifact_grid_projection_incident_id_fkey`, `artifact_grid_projection_incoming_owner_user_id_fkey`, `artifact_grid_projection_investigative_query_created_by_us_fkey`, `artifact_grid_projection_outgoing_owner_user_id_fkey`, `artifact_grid_projection_owner_user_id_fkey`, `artifact_grid_projection_pkey`, `artifact_grid_projection_record_id_fkey`, `artifact_grid_projection_review_owner_user_id_fkey`, `assessment_grid_projection_assessment_state_check`, `assessment_grid_projection_assessor_fkey`, `assessment_grid_projection_confidence_band_check`, `assessment_grid_projection_confidence_score_check`, `assessment_grid_projection_incident_id_fkey`, `assessment_grid_projection_pkey`, `assessment_grid_projection_record_id_fkey`, `assessment_grid_projection_subject_type_check`, `decision_grid_projection_incident_id_fkey`, `decision_grid_projection_owner_user_id_fkey`, `decision_grid_projection_pkey`, `decision_grid_projection_record_id_fkey`, `evidence_grid_projection_incident_id_fkey`, `evidence_grid_projection_pkey`, `evidence_grid_projection_record_id_fkey`, `host_grid_projection_host_state_check`, `host_grid_projection_incident_id_fkey`, `host_grid_projection_pkey`, `host_grid_projection_record_id_fkey`, `identity_grid_projection_identity_state_check`, `identity_grid_projection_incident_id_fkey`, `identity_grid_projection_pkey`, `identity_grid_projection_record_id_fkey`, `indicator_grid_projection_incident_id_fkey`, `indicator_grid_projection_pkey`, `indicator_grid_projection_record_id_fkey`, `party_grid_projection_incident_id_fkey`, `party_grid_projection_pkey`, `party_grid_projection_record_id_fkey`, `task_request_grid_projection_incident_id_fkey`, `task_request_grid_projection_owner_user_id_fkey`, `task_request_grid_projection_pkey`, `task_request_grid_projection_record_id_fkey`, `timeline_grid_projection_activity_time_pair_state_check`, `timeline_grid_projection_attached_evidence_refs_array_ck`, `timeline_grid_projection_capture_state_check`, `timeline_grid_projection_host_refs_array_ck`, `timeline_grid_projection_identity_refs_array_ck`, `timeline_grid_projection_incident_id_fkey`, `timeline_grid_projection_pkey`, `timeline_grid_projection_record_id_fkey`, `timeline_grid_projection_replacement_record_id_fkey`, `timeline_grid_projection_tags_array_ck`.
- Indexes: `artifact_grid_projection_created_by_user_id_fk_idx`, `artifact_grid_projection_finding_owner_user_id_fk_idx`, `artifact_grid_projection_finding_state_idx`, `artifact_grid_projection_incident_type_updated_idx`, `artifact_grid_projection_incoming_owner_user_id_fk_idx`, `artifact_grid_projection_investigative_query_created_b_47384917`, `artifact_grid_projection_keyword_mode_idx`, `artifact_grid_projection_note_linked_idx`, `artifact_grid_projection_outgoing_owner_user_id_fk_idx`, `artifact_grid_projection_owner_user_id_fk_idx`, `artifact_grid_projection_query_platform_idx`, `artifact_grid_projection_review_owner_user_id_fk_idx`, `assessment_grid_projection_assessor_fk_idx`, `assessment_grid_projection_incident_confidence_band_idx`, `assessment_grid_projection_incident_sort_idx`, `assessment_grid_projection_incident_state_idx`, `decision_grid_projection_incident_decided_idx`, `decision_grid_projection_owner_user_id_fk_idx`, `decision_grid_projection_review_queue_idx`, `decision_grid_projection_superseded_idx`, `evidence_grid_projection_incident_requested_idx`, `evidence_grid_projection_lifecycle_idx`, `host_grid_projection_incident_sort_idx`, `identity_grid_projection_incident_sort_idx`, `indicator_grid_projection_incident_lifecycle_idx`, `indicator_grid_projection_incident_sort_idx`, `party_grid_projection_incident_display_name_idx`, `party_grid_projection_kind_idx`, `task_request_grid_projection_due_idx`, `task_request_grid_projection_incident_updated_idx`, `task_request_grid_projection_no_owner_idx`, `task_request_grid_projection_owner_user_id_fk_idx`, `task_request_grid_projection_queue_idx`, `timeline_grid_projection_incident_activity_sort_idx`, `timeline_grid_projection_replacement_record_id_fk_idx`.
- FK prerequisites: `artifacts`, `assessments`, `decisions`, `evidence`, `hosts`, `identities`, `incidents`, `indicators`, `parties`, `task_requests`, `timeline_events`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00022_graph_projection.sql

Original SHA-256: `6bb131c75544bcba96fcc4a47654da2fa8cca5e56b647197b265ba1099f5d7a3`.

- Tables created/altered/dropped: `graph_projection_edges`, `graph_projection_idempotency`, `graph_projection_runs`, `graph_projection_vertices`, `graph_projection_views`.
- Constraints: `graph_projection_edges_pkey`, `graph_projection_edges_projection_run_id_fkey`, `graph_projection_idempotency_pkey`, `graph_projection_runs_graph_view_id_fkey`, `graph_projection_runs_pkey`, `graph_projection_runs_projection_output_digest_ck`, `graph_projection_runs_retention_policy_ck`, `graph_projection_runs_state_ck`, `graph_projection_vertices_pkey`, `graph_projection_vertices_projection_run_id_fkey`, `graph_projection_views_graph_view_key_key`, `graph_projection_views_invalidation_ck`, `graph_projection_views_pkey`, `graph_projection_views_selected_run_fkey`, `graph_projection_views_state_ck`.
- Indexes: `graph_projection_edges_endpoint_idx`, `graph_projection_edges_run_id_lookup_idx`, `graph_projection_edges_view_sort_idx`, `graph_projection_idempotency_expiry_idx`, `graph_projection_runs_failed_retention_idx`, `graph_projection_runs_invalidated_retention_idx`, `graph_projection_runs_one_active_per_view_idx`, `graph_projection_runs_replaced_retention_idx`, `graph_projection_runs_view_state_idx`, `graph_projection_vertices_run_id_lookup_idx`, `graph_projection_vertices_view_sort_idx`, `graph_projection_views_selected_projection_run_id_fk_idx`.
- FK prerequisites: `graph_projection_runs`, `graph_projection_views`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00023_reporting.sql

Original SHA-256: `6fec443360ea1a6be24fd73c661c266adf06ac41298b38cf362281f484f5e9e9`.

- Tables created/altered/dropped: `reporting_job_payloads`, `reporting_release_approvals`, `reporting_releases`, `reporting_render_bundle_files`, `reporting_render_bundles`, `reporting_snapshots`.
- Constraints: `reporting_job_payloads_actor_user_id_fkey`, `reporting_job_payloads_incident_id_fkey`, `reporting_job_payloads_job_id_fkey`, `reporting_job_payloads_job_kind_check`, `reporting_job_payloads_pkey`, `reporting_release_approvals_actor_user_id_fkey`, `reporting_release_approvals_approval_role_check`, `reporting_release_approvals_pkey`, `reporting_release_approvals_release_id_actor_user_id_approv_key`, `reporting_release_approvals_release_id_fkey`, `reporting_releases_composition_tuple_ck`, `reporting_releases_create_job_id_fkey`, `reporting_releases_created_by_user_id_fkey`, `reporting_releases_graph_projection_refs_ck`, `reporting_releases_incident_id_fkey`, `reporting_releases_output_kind_check`, `reporting_releases_output_options_ck`, `reporting_releases_pkey`, `reporting_releases_recipient_partition_refs_ck`, `reporting_releases_release_scope_check`, `reporting_releases_release_state_check`, `reporting_releases_snapshot_id_fkey`, `reporting_render_bundle_files_bundle_path_check`, `reporting_render_bundle_files_check`, `reporting_render_bundle_files_file_sha256_check`, `reporting_render_bundle_files_media_type_check`, `reporting_render_bundle_files_pkey`, `reporting_render_bundle_files_release_id_fkey`, `reporting_render_bundle_files_role_check`, `reporting_render_bundle_files_size_bytes_check`, `reporting_render_bundle_files_storage_kind_check`, `reporting_render_bundles_bundle_manifest_json_check`, `reporting_render_bundles_bundle_manifest_sha256_check`, `reporting_render_bundles_pkey`, `reporting_render_bundles_primary_bundle_path_check`, `reporting_render_bundles_primary_media_type_check`, `reporting_render_bundles_release_id_fkey`, `reporting_snapshots_create_job_id_fkey`, `reporting_snapshots_created_by_user_id_fkey`, `reporting_snapshots_incident_id_fkey`, `reporting_snapshots_pkey`, `reporting_snapshots_source_boundary_json_check`, `reporting_snapshots_source_boundary_json_ck`.
- Indexes: `reporting_job_payloads_actor_user_id_fk_idx`, `reporting_job_payloads_incident_lookup_idx`, `reporting_release_approvals_actor_user_id_fk_idx`, `reporting_release_approvals_release_lookup_idx`, `reporting_releases_create_job_id_fk_idx`, `reporting_releases_created_by_lookup_idx`, `reporting_releases_incident_lookup_idx`, `reporting_releases_snapshot_lookup_idx`, `reporting_render_bundle_files_release_idx`, `reporting_snapshots_create_job_id_fk_idx`, `reporting_snapshots_created_by_lookup_idx`, `reporting_snapshots_incident_lookup_idx`.
- FK prerequisites: `incidents`, `jobs`, `reporting_releases`, `reporting_render_bundles`, `reporting_snapshots`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00024_report_composition.sql

Original SHA-256: `5d87614407847f6728b8203b2a00a04074085de5ccefca1f463c75e7ca7b6402`.

- Tables created/altered/dropped: `report_composition_preview_attempts`, `report_composition_release_bindings`, `report_composition_versions`, `report_compositions`, `reporting_composition_preview_output_files`, `reporting_composition_preview_outputs`.
- Constraints: `report_composition_preview_attem_recipient_partition_refs_check`, `report_composition_preview_attem_redaction_profile_sha256_check`, `report_composition_preview_attempts_composition_id_fkey`, `report_composition_preview_attempts_composition_sha256_check`, `report_composition_preview_attempts_composition_version_check`, `report_composition_preview_attempts_created_by_user_id_fkey`, `report_composition_preview_attempts_draft_version_check`, `report_composition_preview_attempts_graph_projection_refs_check`, `report_composition_preview_attempts_incident_id_fkey`, `report_composition_preview_attempts_output_kind_check`, `report_composition_preview_attempts_output_options_check`, `report_composition_preview_attempts_pkey`, `report_composition_preview_attempts_preview_source_json_check`, `report_composition_preview_attempts_preview_source_sha256_check`, `report_composition_preview_attempts_render_attempt_fk`, `report_composition_preview_attempts_source_kind_check`, `report_composition_preview_source_shape_ck`, `report_composition_release_bi_composition_id_composition_v_fkey`, `report_composition_release_bindings_composition_sha256_check`, `report_composition_release_bindings_pkey`, `report_composition_release_bindings_release_fk`, `report_composition_release_bindings_release_scope_check`, `report_composition_versions_canonical_bytes_ck`, `report_composition_versions_canonical_composition_check`, `report_composition_versions_composition_id_fkey`, `report_composition_versions_composition_sha256_check`, `report_composition_versions_composition_version_check`, `report_composition_versions_created_by_user_id_fkey`, `report_composition_versions_pkey`, `report_compositions_authored_texts_array_ck`, `report_compositions_created_by_user_id_fkey`, `report_compositions_deck_ops_array_ck`, `report_compositions_diagram_decls_array_ck`, `report_compositions_draft_version_check`, `report_compositions_incident_id_fkey`, `report_compositions_latest_composition_version_check`, `report_compositions_pkey`, `reporting_composition_preview_output_files_attempt_fk`, `reporting_composition_preview_output_files_path_ck`, `reporting_composition_preview_output_files_pkey`, `reporting_composition_preview_output_files_sha_ck`, `reporting_composition_preview_output_files_size_ck`, `reporting_composition_preview_output_files_storage_ck`, `reporting_composition_preview_outputs_attempt_fk`, `reporting_composition_preview_outputs_job_fk`, `reporting_composition_preview_outputs_kind_ck`, `reporting_composition_preview_outputs_manifest_ck`, `reporting_composition_preview_outputs_pkey`, `reporting_composition_preview_outputs_render_attempt_key`, `reporting_composition_preview_outputs_scope_ck`, `reporting_composition_preview_outputs_sha_ck`.
- Indexes: `report_composition_preview_attempts_composition_lookup_idx`, `report_composition_preview_attempts_created_by_user_id_fk_idx`, `report_composition_preview_attempts_incident_id_fk_idx`, `report_composition_preview_attempts_render_attempt_id_fk_idx`, `report_composition_release_bindings_release_lookup_idx`, `report_composition_versions_created_by_user_id_fk_idx`, `report_composition_versions_digest_idx`, `report_compositions_created_by_lookup_idx`, `report_compositions_incident_lookup_idx`.
- FK prerequisites: `incidents`, `jobs`, `report_composition_preview_attempts`, `report_composition_versions`, `report_compositions`, `reporting_composition_preview_outputs`, `reporting_releases`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00025_incident_bundles.sql

Original SHA-256: `2bde61aa7a4a0e1f791107b59fbbe8b8931356dd77c1c213d417e1f4b9482a17`.

- Tables created/altered/dropped: `incident_bundle_exports`, `incident_bundle_imported_actors`, `incident_bundle_imported_attributions`, `incident_bundle_job_payloads`, `incident_bundle_manifest_files`.
- Constraints: `incident_bundle_exports_blob_mode_check`, `incident_bundle_exports_bundle_byte_size_check`, `incident_bundle_exports_bundle_sha256_check`, `incident_bundle_exports_export_job_id_fkey`, `incident_bundle_exports_export_job_id_key`, `incident_bundle_exports_exported_by_user_id_fkey`, `incident_bundle_exports_history_mode_check`, `incident_bundle_exports_incident_id_fkey`, `incident_bundle_exports_manifest_sha256_check`, `incident_bundle_exports_pkey`, `incident_bundle_exports_reference_pack_mode_check`, `incident_bundle_exports_storage_ref_ck`, `incident_bundle_imported_actors_incident_id_fkey`, `incident_bundle_imported_actors_incident_id_source_actor_id_key`, `incident_bundle_imported_actors_local_user_id_fkey`, `incident_bundle_imported_actors_pkey`, `incident_bundle_imported_attr_incident_id_source_table_sour_key`, `incident_bundle_imported_attributions_incident_id_fkey`, `incident_bundle_imported_attributions_local_user_id_fkey`, `incident_bundle_imported_attributions_pkey`, `incident_bundle_job_payloads_actor_user_id_fkey`, `incident_bundle_job_payloads_incident_id_fkey`, `incident_bundle_job_payloads_job_id_fkey`, `incident_bundle_job_payloads_job_kind_check`, `incident_bundle_job_payloads_manifest_sha256_check`, `incident_bundle_job_payloads_pkey`, `incident_bundle_job_payloads_staging_ref_ck`, `incident_bundle_job_payloads_uploaded_sha256_check`, `incident_bundle_manifest_files_bundle_id_fkey`, `incident_bundle_manifest_files_pkey`, `incident_bundle_manifest_files_sha256_check`, `incident_bundle_manifest_files_size_bytes_check`, `incident_bundle_payload_kind_ck`.
- Indexes: `incident_bundle_exports_exported_by_user_id_fk_idx`, `incident_bundle_exports_incident_lookup_idx`, `incident_bundle_imported_actors_local_user_id_fk_idx`, `incident_bundle_imported_attributions_actor_idx`, `incident_bundle_imported_attributions_local_user_id_fk_idx`, `incident_bundle_job_payloads_actor_idx`, `incident_bundle_job_payloads_incident_id_fk_idx`.
- FK prerequisites: `incident_bundle_exports`, `incidents`, `jobs`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00026_reference_data.sql

Original SHA-256: `c3581b3c7de76475fc26d6b80b8897d6efc5edfc2d918096a4ecaf1770941048`.

- Tables created/altered/dropped: `reference_pack_activation_state`, `reference_pack_attestations`, `reference_pack_job_payloads`, `reference_packs`.
- Constraints: `reference_pack_activation_sta_pack_key_previous_active_ver_fkey`, `reference_pack_activation_state_activated_by_user_id_fkey`, `reference_pack_activation_state_check`, `reference_pack_activation_state_pack_key_active_version_fkey`, `reference_pack_activation_state_pkey`, `reference_pack_attestations_actor_user_id_fkey`, `reference_pack_attestations_event_kind_check`, `reference_pack_attestations_job_id_fkey`, `reference_pack_attestations_pack_key_pack_version_fkey`, `reference_pack_attestations_pack_key_previous_active_versi_fkey`, `reference_pack_attestations_pkey`, `reference_pack_attestations_verification_result_check`, `reference_pack_job_payloads_actor_user_id_fkey`, `reference_pack_job_payloads_bundle_staging_ref_relative_check`, `reference_pack_job_payloads_job_id_fkey`, `reference_pack_job_payloads_job_kind_check`, `reference_pack_job_payloads_pkey`, `reference_packs_activated_by_user_id_fkey`, `reference_packs_bundle_storage_ref_relative_check`, `reference_packs_imported_by_user_id_fkey`, `reference_packs_pack_key_previous_active_version_fkey`, `reference_packs_pkey`, `reference_packs_status_check`, `reference_packs_verification_result_check`.
- Indexes: `reference_pack_activation_state_activated_by_user_id_fk_idx`, `reference_pack_activation_state_pack_key_active_version_fk_idx`, `reference_pack_activation_state_pack_key_previous_acti_eb58f173`, `reference_pack_attestations_actor_user_id_fk_idx`, `reference_pack_attestations_job_id_fk_idx`, `reference_pack_attestations_pack_idx`, `reference_pack_attestations_pack_key_previous_active_v_75bb245c`, `reference_pack_job_payloads_actor_idx`, `reference_packs_activated_by_user_id_fk_idx`, `reference_packs_imported_by_user_id_fk_idx`, `reference_packs_pack_key_previous_active_version_fk_idx`, `reference_packs_status_idx`.
- FK prerequisites: `jobs`, `reference_packs`, `users`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00027_extensions.sql

Original SHA-256: `84f4f5891761edb0d366247bf34c7e1142d470951dbc0a3bf53866fba11295ae`.

- Tables created/altered/dropped: `extension_job_cancellation_observations`, `extension_job_commit_proofs`, `extension_migration_ledger`, `extension_staged_object_references`, `extension_staged_objects`, `extension_state_metadata`.
- Constraints: `extension_job_cancellation_observations_id_ck`, `extension_job_cancellation_observations_job_id_fkey`, `extension_job_cancellation_observations_job_id_key`, `extension_job_cancellation_observations_pkey`, `extension_job_cancellation_observations_precommit_ck`, `extension_job_commit_proofs_audit_ck`, `extension_job_commit_proofs_commit_id_ck`, `extension_job_commit_proofs_final_commit_id_key`, `extension_job_commit_proofs_job_id_fkey`, `extension_job_commit_proofs_operation_ck`, `extension_job_commit_proofs_pkey`, `extension_job_commit_proofs_profile_ck`, `extension_job_commit_proofs_refs_ck`, `extension_job_commit_proofs_request_digest_ck`, `extension_job_commit_proofs_result_ck`, `extension_job_commit_proofs_result_digest_ck`, `extension_migration_ledger_digest_ck`, `extension_migration_ledger_lineage_ck`, `extension_migration_ledger_metadata_fk`, `extension_migration_ledger_migration_id_ck`, `extension_migration_ledger_migration_identity_uq`, `extension_migration_ledger_pk`, `extension_migration_ledger_profile_id_ck`, `extension_migration_ledger_versions_ck`, `extension_staged_object_references_digest_ck`, `extension_staged_object_references_id_ck`, `extension_staged_object_references_kind_ck`, `extension_staged_object_references_pkey`, `extension_staged_object_references_size_ck`, `extension_staged_object_references_staging_id_fkey`, `extension_staged_objects_attempt_count_ck`, `extension_staged_objects_delete_state_ck`, `extension_staged_objects_digest_ck`, `extension_staged_objects_error_code_ck`, `extension_staged_objects_expiry_ck`, `extension_staged_objects_id_ck`, `extension_staged_objects_lifecycle_ck`, `extension_staged_objects_pkey`, `extension_staged_objects_profile_id_ck`, `extension_staged_objects_size_ck`, `extension_staged_objects_state_ck`, `extension_staged_objects_storage_identity_ck`, `extension_staged_objects_storage_identity_key`, `extension_state_metadata_last_migration_ck`, `extension_state_metadata_lineage_ck`, `extension_state_metadata_metadata_version_ck`, `extension_state_metadata_pkey`, `extension_state_metadata_profile_id_ck`, `extension_state_metadata_state_version_ck`, `extension_state_metadata_time_ck`.
- Indexes: `extension_staged_objects_cleanup_idx`.
- FK prerequisites: `extension_staged_objects`, `extension_state_metadata`, `jobs`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00028_administrative_audit.sql

Original SHA-256: `5870074158100d7cdf1549055e303f0d38bc7f81ab82b3fd01bc3a55a9668189`.

- Tables created/altered/dropped: `administrative_audit_projections`.
- Constraints: `administrative_audit_projections_action_scope_check`, `administrative_audit_projections_actor_check`, `administrative_audit_projections_changes_check`, `administrative_audit_projections_pkey`, `administrative_audit_projections_raw_event_fkey`, `administrative_audit_projections_scope_check`, `administrative_audit_projections_source_check`, `administrative_audit_projections_target_check`.
- Indexes: `administrative_audit_projections_deployment_page_idx`, `administrative_audit_projections_filter_idx`, `administrative_audit_projections_incident_page_idx`.
- Routines: `reject_administrative_audit_mutation`.
- Triggers: `administrative_audit_projections_immutable`, `deployment_admin_audit_events_immutable`.
- FK prerequisites: `deployment_admin_audit_events`.
- Privilege operations: 1; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00029_collaboration.sql

Original SHA-256: `a14e3809974a118ed700a2ce5392fa27b7eb05b7d7a22fbb35577585b01be3f1`.

- Tables created/altered/dropped: `collaboration_event_intents`, `collaboration_incident_stream_cursors`, `collaboration_replay_events`, `collaboration_resume_tokens`.
- Constraints: `collaboration_event_intents_attempt_ck`, `collaboration_event_intents_dispatch_state_ck`, `collaboration_event_intents_error_ck`, `collaboration_event_intents_family_ck`, `collaboration_event_intents_incident_id_fkey`, `collaboration_event_intents_intent_key_key`, `collaboration_event_intents_key_ck`, `collaboration_event_intents_ordinal_ck`, `collaboration_event_intents_payload_ck`, `collaboration_event_intents_payload_size_ck`, `collaboration_event_intents_pkey`, `collaboration_event_intents_row_version_ck`, `collaboration_event_intents_sequence_ck`, `collaboration_event_intents_source_identity_ck`, `collaboration_event_intents_time_ck`, `collaboration_incident_stream_cursors_failure_count_ck`, `collaboration_incident_stream_cursors_high_water_ck`, `collaboration_incident_stream_cursors_incident_id_fkey`, `collaboration_incident_stream_cursors_pkey`, `collaboration_incident_stream_cursors_quarantine_ck`, `collaboration_replay_events_family_ck`, `collaboration_replay_events_incident_id_fkey`, `collaboration_replay_events_incident_sequence_uq`, `collaboration_replay_events_intent_key_key`, `collaboration_replay_events_payload_ck`, `collaboration_replay_events_payload_size_ck`, `collaboration_replay_events_pkey`, `collaboration_replay_events_sequence_ck`, `collaboration_resume_tokens_client_ck`, `collaboration_resume_tokens_expiry_ck`, `collaboration_resume_tokens_hash_ck`, `collaboration_resume_tokens_incident_id_fkey`, `collaboration_resume_tokens_pkey`, `collaboration_resume_tokens_session_id_fkey`.
- Indexes: `collaboration_event_intents_incident_id_fk_idx`, `collaboration_resume_tokens_incident_id_fk_idx`, `collaboration_resume_tokens_session_id_fk_idx`, `idx_collaboration_event_intents_dispatch`, `idx_collaboration_incident_stream_quarantine`, `idx_collaboration_replay_events_retention`, `idx_collaboration_resume_tokens_expiry`.
- FK prerequisites: `incidents`, `user_sessions`.
- Privilege operations: 26; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00030_evidence_upload_leases.sql

Original SHA-256: `b068bff7c5a93d7d82e3d4674e3cf8d2280154933c4b5162cb46d9a9cc9514cc`.

- Tables created/altered/dropped: `evidence_object_upload_leases`.
- Constraints: `evidence_object_upload_leases_blob_fkey`, `evidence_object_upload_leases_capability_digest_ck`, `evidence_object_upload_leases_capability_hash_key`, `evidence_object_upload_leases_contract_digest_ck`, `evidence_object_upload_leases_headers_ck`, `evidence_object_upload_leases_incident_fkey`, `evidence_object_upload_leases_lease_id_key`, `evidence_object_upload_leases_method_ck`, `evidence_object_upload_leases_pkey`, `evidence_object_upload_leases_session_fkey`, `evidence_object_upload_leases_state_ck`, `evidence_object_upload_leases_time_ck`, `evidence_object_upload_leases_transition_ck`, `evidence_object_upload_leases_user_fkey`.
- Indexes: `evidence_object_upload_leases_expiry_idx`, `evidence_object_upload_leases_incident_id_fk_idx`, `evidence_object_upload_leases_issuing_session_id_fk_idx`, `evidence_object_upload_leases_issuing_user_id_fk_idx`.
- FK prerequisites: `incidents`, `object_blobs`, `user_sessions`, `users`.
- Privilege operations: 3; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00031_evidence_blob_cleanup_claims.sql

Original SHA-256: `acf86625925b46198086acd93fa232d999165bf0fe2b4e4e657cacaceedcb95c`.

- Tables created/altered/dropped: `evidence_blob_cleanup_claims`.
- Constraints: `evidence_blob_cleanup_claims_attempt_count_ck`, `evidence_blob_cleanup_claims_blob_fkey`, `evidence_blob_cleanup_claims_claim_token_key`, `evidence_blob_cleanup_claims_failure_class_ck`, `evidence_blob_cleanup_claims_pkey`, `evidence_blob_cleanup_claims_state_ck`, `evidence_blob_cleanup_claims_time_ck`, `evidence_blob_cleanup_claims_transition_ck`.
- Indexes: `evidence_blob_cleanup_claims_completed_idx`, `evidence_blob_cleanup_claims_due_idx`, `evidence_blob_cleanup_claims_lease_expiry_idx`.
- Routines: `evidence_reject_cleanup_claimed_association`, `evidence_reject_cleanup_claimed_blob_update`.
- Triggers: `evidence_reject_cleanup_claimed_association`, `evidence_reject_cleanup_claimed_blob_update`.
- FK prerequisites: `object_blobs`.
- Privilege operations: 5; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00032_graph_projection_v2.sql

Original SHA-256: `393aec3abf10fa02f1cd7b26d21dfa4a98152ee1b1f7a162e419ab18725b9314`.

- Tables created/altered/dropped: `graph_projection_result_edges`, `graph_projection_result_leases`, `graph_projection_result_vertices`, `graph_projection_results`, `network_flow_graph_views`.
- Constraints: `graph_projection_result_edges_direction_ck`, `graph_projection_result_edges_dst_vertex_fkey`, `graph_projection_result_edges_id_ck`, `graph_projection_result_edges_json_ck`, `graph_projection_result_edges_kind_ck`, `graph_projection_result_edges_ordinal_ck`, `graph_projection_result_edges_pkey`, `graph_projection_result_edges_result_id_fkey`, `graph_projection_result_edges_result_ordinal_key`, `graph_projection_result_edges_src_vertex_fkey`, `graph_projection_result_leases_identity_key`, `graph_projection_result_leases_owner_ck`, `graph_projection_result_leases_pkey`, `graph_projection_result_leases_result_id_fkey`, `graph_projection_result_leases_time_ck`, `graph_projection_result_vertices_id_ck`, `graph_projection_result_vertices_json_ck`, `graph_projection_result_vertices_kind_ck`, `graph_projection_result_vertices_ordinal_ck`, `graph_projection_result_vertices_pkey`, `graph_projection_result_vertices_result_id_fkey`, `graph_projection_result_vertices_result_ordinal_key`, `graph_projection_results_binding_key`, `graph_projection_results_counts_ck`, `graph_projection_results_id_ck`, `graph_projection_results_identity_ck`, `graph_projection_results_json_ck`, `graph_projection_results_pkey`, `network_flow_graph_views_created_by_user_id_fkey`, `network_flow_graph_views_display_name_ck`, `network_flow_graph_views_failure_ck`, `network_flow_graph_views_id_ck`, `network_flow_graph_views_incident_id_fkey`, `network_flow_graph_views_latest_job_id_fkey`, `network_flow_graph_views_pkey`, `network_flow_graph_views_selected_binding_ck`, `network_flow_graph_views_semantic_query_ck`, `network_flow_graph_views_source_snapshot_ck`, `network_flow_graph_views_state_ck`, `network_flow_graph_views_state_time_ck`, `network_flow_graph_views_times_ck`, `network_flow_graph_views_versions_ck`.
- Indexes: `graph_projection_result_edges_dst_traversal_idx`, `graph_projection_result_edges_order_idx`, `graph_projection_result_edges_src_traversal_idx`, `graph_projection_result_leases_expiry_idx`, `graph_projection_result_vertices_order_idx`, `network_flow_graph_views_active_order_idx`, `network_flow_graph_views_created_by_user_id_fk_idx`, `network_flow_graph_views_latest_job_id_fk_idx`, `network_flow_graph_views_retained_count_idx`, `network_flow_graph_views_selected_result_idx`.
- Routines: `graph_projection_reject_immutable_result_update`.
- Triggers: `graph_projection_result_edges_immutable_update`, `graph_projection_result_vertices_immutable_update`, `graph_projection_results_immutable_update`.
- FK prerequisites: `graph_projection_result_vertices`, `graph_projection_results`, `incidents`, `jobs`, `users`.
- Privilege operations: 13; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00033_graph_projection_v1_removal.sql

Original SHA-256: `ca2d578d53768dda6219afe2de5010b33b3a2540a3793af7a646426c5ceb5f43`.

- Tables created/altered/dropped: `graph_projection_edges`, `graph_projection_idempotency`, `graph_projection_runs`, `graph_projection_vertices`, `graph_projection_views`.
- Constraints: `graph_projection_views_selected_run_fkey`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00034_graph_projection_cleanup_indexes.sql

Original SHA-256: `615624aeb979fe3d376a8b013666ac5f175e9c343147baf0f714bfc006e81ca3`.

- Indexes: `graph_projection_results_cleanup_candidate_idx`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00035_assessment_create_idempotency_v1.sql

Original SHA-256: `f5800e4d9b733a279d93743506d6e12ca86f9b2d8a3637161c4b6072678dfacb`.

- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: rewrite legacy assessment replay JSON; retired conversion.

#### 00036_entities_source_integrity.sql

Original SHA-256: `b42a7676d6cbf37094697d6d21ede0ad7efc8171f6de44e279dd7a6f9d2b3ca7`.

- Tables created/altered/dropped: `entity_mentions`, `entity_preserved_identifiers`.
- Constraints: `entity_mentions_origin_kind_ck`, `entity_mentions_required_text_ck`, `entity_mentions_resolution_tuple_ck`, `entity_mentions_row_version_ck`, `entity_preserved_identifiers_identifier_class_ck`, `entity_preserved_identifiers_value_nonempty_ck`.
- Routines: `entities_assert_entity_child_v1`, `entities_assert_entity_envelope_update_v1`, `entities_assert_entity_source_v1`, `entities_reject_entity_source_delete_v1`, `entities_source_codepoints_admitted_v1`, `entities_trim_unicode_space_v1`.
- Triggers: `entities_assert_entity_alias`, `entities_assert_entity_envelope_update`, `entities_assert_entity_mention`, `entities_assert_entity_preserved_identifier`, `entities_assert_host_source`, `entities_assert_identity_source`, `entities_reject_host_source_delete`, `entities_reject_identity_source_delete`.
- Privilege operations: 8; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00037_entities_active_identifier_claims.sql

Original SHA-256: `b79529d6a876c1284a14899cce2a2b4f769fb535ea4250f3b24fab95c75976fa`.

- Tables created/altered/dropped: `entity_active_identifier_claims`, `entity_preserved_identifiers`, `hosts`, `identities`, `records`.
- Constraints: `entity_active_identifier_claims_entity_type_ck`, `entity_active_identifier_claims_identifier_class_ck`, `entity_active_identifier_claims_pkey`, `entity_active_identifier_claims_record_envelope_fkey`, `entity_active_identifier_claims_value_ck`.
- Indexes: `entity_active_identifier_claims_record_envelope_idx`, `entity_active_identifier_claims_record_lookup_idx`, `entity_preserved_identifiers_exact_lookup_idx`, `hosts_incident_aad_device_id_idx`, `hosts_incident_fqdn_idx`, `hosts_incident_hostname_idx`, `identities_incident_aad_object_id_idx`, `identities_incident_email_idx`, `identities_incident_sam_account_name_idx`, `identities_incident_sid_idx`, `identities_incident_upn_idx`.
- Routines: `entities_active_identifier_claims_are_valid_v1`, `entities_expected_active_identifier_claims_v1`, `entities_normalize_identifier_v1`, `entities_rebuild_active_identifier_claims_v1`, `entities_refresh_active_identifier_claims_v1`, `entities_release_active_identifier_claims_v1`, `entities_sync_active_identifier_claims_v1`.
- Triggers: `entity_preserved_identifiers_sync_active_claims`, `hosts_sync_active_identifier_claims`, `identities_sync_active_identifier_claims`, `records_sync_entity_active_identifier_claims`.
- FK prerequisites: `records`.
- Privilege operations: 13; final grants are incorporated in the owning definitions.
- Non-DDL: rebuild existing entity claims; fresh empty rebuild omitted, routine retained.

#### 00038_indicators_lifecycle_integrity.sql

Original SHA-256: `0eaed9efe6779975b85c00c9db28790c02f4c6194fd09237526a99f9579293ed`.

- Tables created/altered/dropped: `indicator_state_intervals`.
- Constraints: `indicator_state_intervals_support_refs_ck`.
- Routines: `indicator_support_refs_are_valid`.
- Privilege operations: 2; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00039_parties_active_key_claims.sql

Original SHA-256: `02af5462b596973a1a292efda0e99b6e23a80c0887c01d44e0f6e9551a41b467`.

- Tables created/altered/dropped: `parties`, `party_active_key_claims`, `records`.
- Constraints: `parties_external_ref_claim_value_ck`, `parties_primary_email_claim_value_ck`, `party_active_key_claims_key_kind_ck`, `party_active_key_claims_pkey`, `party_active_key_claims_record_envelope_fkey`, `party_active_key_claims_record_key_unique`, `party_active_key_claims_value_ck`.
- Indexes: `party_active_key_claims_record_lookup_idx`.
- Routines: `parties_active_key_claims_are_valid_v1`, `parties_expected_active_key_claims_v1`, `parties_normalize_active_key_v1`, `parties_rebuild_active_key_claims_v1`, `parties_refresh_active_key_claims_v1`, `parties_release_active_key_claims_v1`, `parties_sync_active_key_claims_v1`, `parties_trim_unicode_space_v1`.
- Triggers: `parties_sync_active_key_claims`, `records_sync_party_active_key_claims`.
- FK prerequisites: `records`.
- Privilege operations: 15; final grants are incorporated in the owning definitions.
- Non-DDL: rebuild existing Party claims; fresh empty rebuild omitted, routine retained.

#### 00040_parties_mutation_request_hash_v1.sql

Original SHA-256: `7609c33e820cb85080a4c48279043bc020b7b15a888325ff6ca437431e365f94`.

- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: delete obsolete Party replay receipts; fresh baseline has none.

#### 00041_entities_identifier_unicode_v1.sql

Original SHA-256: `ce5928b19aba9167420017a10e24c557c5a6eb3bce7c9b3e6894e7f14f6a6453`.

- Tables created/altered/dropped: `entity_active_identifier_claims`.
- Constraints: `entity_active_identifier_claims_value_ck`.
- Routines: `entities_assert_entity_child_v1`, `entities_identifier_normalization_previous_v1`, `entities_normalize_identifier_v1`.
- Triggers: `entities_assert_entity_alias`, `entities_assert_entity_mention`, `entities_assert_entity_preserved_identifier`.
- Privilege operations: 3; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00042_timeline_mention_public_identity.sql

Original SHA-256: `7926cda9ffeb373df9af5952d002d1a9ac3ac41f10b871e9a39b36c8ea248a7b`.

- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: backfill mention identity in derived projection JSON; current producer writes it directly.

#### 00043_saved_views_schema_discovery.sql

Original SHA-256: `bd5988f4fe625ae913d1ab29b695786700313745fd5a49817fb019ed6a614b45`.

- Indexes: `saved_views_incident_schema_order_idx`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00044_jobs_restore_enumeration.sql

Original SHA-256: `c9d1367fef287c570b3637bd14590964748a44cae83d7af07019471ac06b30c1`.

- Indexes: `jobs_restored_nonterminal_order_idx`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00045_note_association_lookup.sql

Original SHA-256: `463f5bdd5672e32a8eae16227ba00a0c2a852df66e6632d89ffd701d7cddcb74`.

- Indexes: `record_links_active_null_field_incoming_page_idx`, `record_links_active_null_field_outgoing_page_idx`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00046_reference_pack_immutable_state.sql

Original SHA-256: `77579641153d293af7f02879a01b5fdcdd1bc1730aaedf3cac37594a974e711f`.

- Tables created/altered/dropped: `reference_pack_attempt_members`, `reference_pack_attempts`, `reference_pack_candidates`, `reference_pack_current_set`, `reference_pack_envelopes`, `reference_pack_events`, `reference_pack_index_generations`, `reference_pack_indexes`, `reference_pack_key_state`, `reference_pack_lookup_keys`, `reference_pack_metadata_versions`, `reference_pack_object_refs`, `reference_pack_objects`, `reference_pack_operation_keys`, `reference_pack_operation_members`, `reference_pack_operation_repositories`, `reference_pack_operations`, `reference_pack_pins`, `reference_pack_registry_usage`, `reference_pack_release_bindings`, `reference_pack_repositories`, `reference_pack_roots`, `reference_pack_set_members`, `reference_pack_sets`, `reference_pack_versions`.
- Constraints: `reference_pack_candidate_envelope_fk`, `reference_pack_current_root_fk`, `rp_attempt_members_attempt_id_fk`, `rp_attempt_members_ck1`, `rp_attempt_members_ck2`, `rp_attempt_members_ck3`, `rp_attempt_members_ck4`, `rp_attempt_members_ck5`, `rp_attempt_members_ck6`, `rp_attempt_members_pk1`, `rp_attempts_ck1`, `rp_attempts_ck2`, `rp_attempts_ck3`, `rp_attempts_operation_id_fk`, `rp_attempts_pk1`, `rp_candidates_admitted_by_user_id_fk`, `rp_candidates_ck1`, `rp_candidates_ck2`, `rp_candidates_ck3`, `rp_candidates_ck4`, `rp_candidates_ck5`, `rp_candidates_ck6`, `rp_candidates_index_fk`, `rp_candidates_pack_key_fk`, `rp_candidates_pk1`, `rp_current_set_ck1`, `rp_current_set_ck2`, `rp_current_set_ck3`, `rp_current_set_pack_set_id_fk`, `rp_current_set_pk1`, `rp_envelopes_ck1`, `rp_envelopes_ck2`, `rp_envelopes_ck3`, `rp_envelopes_ck4`, `rp_envelopes_ck5`, `rp_envelopes_ck6`, `rp_envelopes_fk1`, `rp_envelopes_fk2`, `rp_envelopes_operation_fk`, `rp_envelopes_pk1`, `rp_envelopes_repository_id_fk`, `rp_envelopes_uq1`, `rp_envelopes_uq2`, `rp_events_ck1`, `rp_events_ck2`, `rp_events_fk1`, `rp_events_operation_id_fk`, `rp_events_pk1`, `rp_events_uq1`, `rp_index_generations_ck1`, `rp_index_generations_ck2`, `rp_index_generations_ck3`, `rp_index_generations_ck4`, `rp_index_generations_ck5`, `rp_index_generations_operation_id_fk`, `rp_index_generations_pk1`, `rp_index_generations_uq1`, `rp_indexes_ck1`, `rp_indexes_ck2`, `rp_indexes_ck3`, `rp_indexes_index_id_fk`, `rp_indexes_pk1`, `rp_key_state_ck1`, `rp_key_state_pk1`, `rp_lookup_keys_fk1`, `rp_lookup_keys_pk1`, `rp_metadata_versions_ck1`, `rp_metadata_versions_ck2`, `rp_metadata_versions_ck3`, `rp_metadata_versions_fk1`, `rp_metadata_versions_pk1`, `rp_metadata_versions_repository_id_fk`, `rp_object_refs_ck1`, `rp_object_refs_object_id_fk`, `rp_object_refs_pk1`, `rp_objects_ck1`, `rp_objects_ck2`, `rp_objects_ck3`, `rp_objects_pk1`, `rp_objects_uq1`, `rp_operation_keys_ck1`, `rp_operation_keys_operation_id_fk`, `rp_operation_keys_pack_key_fk`, `rp_operation_keys_pk1`, `rp_operation_members_ck1`, `rp_operation_members_fk1`, `rp_operation_members_fk2`, `rp_operation_members_operation_id_fk`, `rp_operation_members_pk1`, `rp_operation_members_uq1`, `rp_operation_repositories_ck1`, `rp_operation_repositories_fk1`, `rp_operation_repositories_operation_id_fk`, `rp_operation_repositories_pk1`, `rp_operations_actor_user_id_fk`, `rp_operations_ck1`, `rp_operations_ck2`, `rp_operations_ck3`, `rp_operations_ck4`, `rp_operations_job_id_fk`, `rp_operations_pk1`, `rp_operations_uq1`, `rp_pins_pack_set_id_fk`, `rp_pins_pk1`, `rp_registry_usage_ck1`, `rp_registry_usage_pack_key_fk`, `rp_registry_usage_pk1`, `rp_release_bindings_fk1`, `rp_release_bindings_pk1`, `rp_repositories_ck1`, `rp_repositories_ck2`, `rp_repositories_pk1`, `rp_roots_ck1`, `rp_roots_ck2`, `rp_roots_ck3`, `rp_roots_ck4`, `rp_roots_ck5`, `rp_roots_fk1`, `rp_roots_pk1`, `rp_roots_repository_id_fk`, `rp_set_members_fk1`, `rp_set_members_fk2`, `rp_set_members_pack_set_id_fk`, `rp_set_members_pk1`, `rp_sets_ck1`, `rp_sets_ck2`, `rp_sets_ck3`, `rp_sets_first_operation_id_fk`, `rp_sets_pk1`, `rp_sets_uq1`, `rp_versions_ck1`, `rp_versions_ck2`, `rp_versions_ck3`, `rp_versions_ck4`, `rp_versions_fk1`, `rp_versions_pk1`, `rp_versions_repository_id_fk`, `rp_versions_uq1`, `rp_versions_uq2`.
- Indexes: `reference_pack_attempts_operation_idx`, `reference_pack_lookup_query_idx`, `reference_pack_object_refs_object_idx`, `reference_pack_objects_digest_idx`, `reference_pack_operation_keys_pack_idx`, `reference_pack_pins_set_idx`, `rp_candidates_actor_idx`, `rp_candidates_envelope_idx`, `rp_candidates_index_idx`, `rp_current_set_set_idx`, `rp_envelopes_root_idx`, `rp_envelopes_version_idx`, `rp_events_version_idx`, `rp_index_generations_operation_idx`, `rp_metadata_versions_version_idx`, `rp_operation_members_envelope_idx`, `rp_operation_members_version_idx`, `rp_operation_repositories_root_idx`, `rp_operations_actor_idx`, `rp_release_bindings_version_idx`, `rp_repositories_root_idx`, `rp_roots_predecessor_idx`, `rp_set_members_envelope_idx`, `rp_set_members_version_idx`, `rp_sets_operation_idx`.
- Routines: `reference_pack_attempt_guard`, `reference_pack_immutable_row`, `reference_pack_operation_guard`.
- Triggers: `reference_pack_attempt_members_immutable`, `reference_pack_attempts_guard`, `reference_pack_envelopes_immutable`, `reference_pack_events_immutable`, `reference_pack_operation_keys_immutable`, `reference_pack_operation_members_immutable`, `reference_pack_operation_repositories_immutable`, `reference_pack_operations_guard`, `reference_pack_release_bindings_immutable`, `reference_pack_roots_immutable`, `reference_pack_set_members_immutable`, `reference_pack_sets_immutable`, `reference_pack_versions_immutable`.
- FK prerequisites: `jobs`, `reference_pack_attempts`, `reference_pack_candidates`, `reference_pack_envelopes`, `reference_pack_index_generations`, `reference_pack_indexes`, `reference_pack_key_state`, `reference_pack_objects`, `reference_pack_operations`, `reference_pack_repositories`, `reference_pack_roots`, `reference_pack_sets`, `reference_pack_versions`, `users`.
- Privilege operations: 55; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00047_reporting_reference_pack_binding.sql

Original SHA-256: `e5a4ac7d64c11bafe907cc92ddb145aa9bf4a4de0121a4988109ccdb61bfe285`.

- Tables created/altered/dropped: `reporting_render_bundles`.
- Constraints: `reporting_render_bundles_bundle_manifest_json_check`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00048_indicator_registry_identity.sql

Original SHA-256: `b52810cbc07c20a14e7b70da879c68e618c52dae8239ef634ab92faabe8f4938`.

- Tables created/altered/dropped: `indicator_active_identities`, `indicators`.
- Constraints: `indicator_active_identities_exact_dedupe_ck`, `indicator_active_identities_pkey`, `indicators_dedupe_key_ck`, `indicators_registry_value_kind_ck`.
- Indexes: `indicators_incident_normalized_lookup_idx`.
- Routines: `indicator_active_identities_are_valid`.
- Privilege operations: 2; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00049_reference_pack_audit.sql

Original SHA-256: `0ea23d0e4c49a2ba81a45e82969d039a17bddf16fd3b4aa83a788b27120d8c5b`.

- Tables created/altered/dropped: `administrative_audit_projections`.
- Constraints: `administrative_audit_projections_action_scope_check`, `administrative_audit_projections_target_check`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00050_reference_pack_invalidation_evidence.sql

Original SHA-256: `9c5bbbce4aee9abdec113ffa2bd2deaae609c7db620fa73b7fc453ab12e4a751`.

- Tables created/altered/dropped: `reference_pack_attempt_members`.
- Constraints: `rp_attempt_members_invalidation_ck`.
- Routines: `reference_pack_attempt_member_guard`.
- Triggers: `reference_pack_attempt_members_guard`, `reference_pack_attempt_members_immutable`.
- Privilege operations: 2; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00051_reference_pack_portability_retention.sql

Original SHA-256: `e73779384d7256fbebca20347ccedfdc1d35653cc8e4961572c7013e69c957e2`.

- Tables created/altered/dropped: `reference_pack_operations`, `reference_pack_portable_catalogs`, `reference_pack_version_pins`.
- Constraints: `rp_operations_ck1`, `rp_portable_catalogs_bytes_ck`, `rp_portable_catalogs_digest_ck`, `rp_portable_catalogs_incident_fk`, `rp_portable_catalogs_operation_fk`, `rp_portable_catalogs_operation_uq`, `rp_portable_catalogs_pk`, `rp_version_pins_envelope_fk`, `rp_version_pins_operation_fk`, `rp_version_pins_pk`.
- Indexes: `rp_version_pins_envelope_idx`, `rp_version_pins_operation_idx`, `rp_version_pins_version_idx`.
- Triggers: `reference_pack_portable_catalogs_immutable`.
- FK prerequisites: `incidents`, `reference_pack_envelopes`, `reference_pack_operations`.
- Privilege operations: 5; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00052_reference_pack_validation_summaries.sql

Original SHA-256: `c36052455684340ad024f0c2866dc1e585a2808a9345159af80c24f85629fa27`.

- Tables created/altered/dropped: `reference_pack_attempt_members`.
- Constraints: `rp_attempt_members_summary_ck`.
- Indexes: `reference_pack_attempt_members_summary_idx`.
- Routines: `reference_pack_attempt_member_guard`.
- Triggers: `reference_pack_attempt_members_guard`.
- Privilege operations: 1; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00053_reference_pack_operator_jobs.sql

Original SHA-256: `47a2c1594575b3bb9be749df99da1b7f121a7650fa877f22da4e2f88c8cecbf8`.

- Tables created/altered/dropped: `jobs`.
- Constraints: `jobs_submitting_actor_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00054_reference_pack_dependency_guards.sql

Original SHA-256: `dc5310cc69c74558c37cddf0d544e8ba4c478dca88ac244bc21a1812c17750a4`.

- Tables created/altered/dropped: `reference_pack_operation_dependency_keys`.
- Constraints: `rp_operation_dependency_keys_key_fk`, `rp_operation_dependency_keys_operation_fk`, `rp_operation_dependency_keys_pk`, `rp_operation_dependency_keys_revision_ck`.
- Indexes: `reference_pack_operation_dependency_keys_key_idx`.
- Triggers: `reference_pack_operation_dependency_keys_immutable`.
- FK prerequisites: `reference_pack_key_state`, `reference_pack_operations`.
- Privilege operations: 4; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00055_reference_pack_portability_requirements.sql

Original SHA-256: `b143c7ce88ddbcaeb5e87417f6614ed610302b51685e83ab1cb807e951a15aa1`.

- Tables created/altered/dropped: `reference_pack_operations`.
- Constraints: `rp_portable_content_input_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00056_reference_pack_portable_preparation.sql

Original SHA-256: `938a05413daf2df31d7e647ddec011cf5862171ac5e5d1451b3a064d00721c60`.

- Tables created/altered/dropped: `reference_pack_operation_members`, `reference_pack_portable_preparations`, `reference_pack_portable_selections`.
- Constraints: `rp_operation_members_portable_binding_uq`, `rp_portable_preparations_context_ck`, `rp_portable_preparations_operation_fk`, `rp_portable_preparations_pk`, `rp_portable_preparations_references_ck`, `rp_portable_selections_availability_ck`, `rp_portable_selections_digest_ck`, `rp_portable_selections_envelope_fk`, `rp_portable_selections_input_ck`, `rp_portable_selections_member_fk`, `rp_portable_selections_object_fk`, `rp_portable_selections_operation_fk`, `rp_portable_selections_ordinal_ck`, `rp_portable_selections_pk`, `rp_portable_selections_reason_ck`, `rp_portable_selections_size_ck`, `rp_portable_selections_verification_uq`, `rp_portable_selections_version_uq`.
- Indexes: `rp_portable_selections_envelope_idx`, `rp_portable_selections_member_idx`, `rp_portable_selections_object_idx`.
- Triggers: `reference_pack_portable_preparations_immutable`, `reference_pack_portable_selections_immutable`.
- FK prerequisites: `reference_pack_envelopes`, `reference_pack_objects`, `reference_pack_operation_members`, `reference_pack_operations`, `reference_pack_portable_preparations`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00057_reference_pack_preparation_privileges.sql

Original SHA-256: `13e48ee0ada3f3853836156b59ae86b1ece5220d1590c466c447218bce116393`.

- Privilege operations: 4; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00058_reference_pack_retire_legacy_tables.sql

Original SHA-256: `fdbb180297083bd355d9648bf19edcf03299a7a51714f4818ecca96bb42c4361`.

- Tables created/altered/dropped: `reference_pack_activation_state`, `reference_pack_attestations`, `reference_pack_job_payloads`, `reference_packs`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00059_reporting_portable_artifacts.sql

Original SHA-256: `1a2cb867a2496707f7f7a7cce49f97d2d2ac6dc3adf3d3022b5c5132a680fbca`.

- Tables created/altered/dropped: `reporting_imported_artifact_files`.
- Constraints: `reporting_imported_artifact_bytes_ck`, `reporting_imported_artifact_files_pk`, `reporting_imported_artifact_incident_fk`, `reporting_imported_artifact_operation_fk`, `reporting_imported_artifact_path_ck`, `reporting_imported_artifact_sha256_ck`, `reporting_imported_artifact_size_ck`.
- Indexes: `reporting_imported_artifact_files_operation_id_fk_idx`.
- Routines: `reporting_reject_imported_artifact_update`.
- Triggers: `reporting_imported_artifacts_immutable`.
- FK prerequisites: `incidents`, `jobs`.
- Privilege operations: 6; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00060_reporting_artifact_trigger_privileges.sql

Original SHA-256: `363d0689689e49d79f68d8acd9443ae31960c3e2901a87f25e8c1905d76f2b6e`.

- Privilege operations: 1; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00061_reference_pack_fallback_attribution.sql

Original SHA-256: `5024dd336c9a001f427307b50823fb23a464209aadd1a071c8022a5d43cb784f`.

- Tables created/altered/dropped: `reference_pack_candidates`.
- Constraints: `rp_candidate_fallback_kind_ck`, `rp_candidate_fallback_version_fk`.
- Indexes: `rp_candidate_fallback_version_idx`.
- FK prerequisites: `reference_pack_candidates`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00062_reference_pack_fallback_full_index.sql

Original SHA-256: `2ca459f7a532dca1c60621b16876d321830f8d45c4d86b07af7406e56a70623d`.

- Indexes: `rp_candidate_fallback_version_idx`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00063_reporting_admitted_model_identity.sql

Original SHA-256: `9b993b4087c7ba92a17b76655ff552df46cc2bb86d8a52d9226f98de657f0ae7`.

- Tables created/altered/dropped: `reporting_snapshots`.
- Constraints: `reporting_snapshot_identity_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00064_application_crypto_format.sql

Original SHA-256: `8055e2999d3d939a78581dc861e40e444bb82a92b094a39d2085fce976b409dc`.

- Tables created/altered/dropped: `application_crypto_format`.
- Constraints: `application_crypto_format_pkey`, `application_crypto_format_singleton_check`.
- Privilege operations: 2; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00065_auth_secret_envelopes.sql

Original SHA-256: `14a2939c41a8ea5176a1b73d269533d2a0e2eb9f6fbb1cc3306eaddfb6cf0728`.

- Tables created/altered/dropped: `enterprise_auth_transactions`, `pending_totp_enrollments`, `users`.
- Constraints: `enterprise_pkce_envelope_check`, `pending_totp_secret_envelope_check`, `users_totp_secret_envelope_check`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00066_recovery_journal_envelopes.sql

Original SHA-256: `22c82803e57866619cd9d3468d510ebb595d89f55ffcb6ee1ba429de345ead24`.

- Tables created/altered/dropped: `operator_recovery_journal`.
- Constraints: `operator_recovery_journal_envelope_schema_id_check`, `operator_recovery_journal_sealed_payload_check`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00067_application_crypto_format_privileges.sql

Original SHA-256: `fef14a51a3772a811f0c82e099ae6b4a8100acd3708621d09a00fd35d164b6b4`.

- Privilege operations: 1; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00068_recovery_transfer_operations.sql

Original SHA-256: `45a4e0e4c11912cda595430a31ab909b52e585e59f9e84158748ffb81a33fc16`.

- Tables created/altered/dropped: `operator_recovery_journal`.
- Constraints: `operator_recovery_journal_operation_check`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00069_reporting_sensitive_artifacts.sql

Original SHA-256: `db24a41b15e732170cb75e107bab954a50871f7a2b74c3de07f0e6553d0d238c`.

- Tables created/altered/dropped: `reporting_composition_preview_output_files`, `reporting_preview_reveal_maps`, `reporting_release_reveal_maps`, `reporting_render_bundle_files`.
- Constraints: `reporting_bundle_public_members_ck`, `reporting_preview_public_members_ck`, `reporting_preview_reveal_maps_bytes_ck`, `reporting_preview_reveal_maps_digest_ck`, `reporting_preview_reveal_maps_parent_fk`, `reporting_preview_reveal_maps_pk`, `reporting_release_reveal_maps_bytes_ck`, `reporting_release_reveal_maps_digest_ck`, `reporting_release_reveal_maps_parent_fk`, `reporting_release_reveal_maps_pk`.
- FK prerequisites: `reporting_composition_preview_outputs`, `reporting_render_bundles`.
- Privilege operations: 4; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00070_reporting_profile_admission.sql

Original SHA-256: `9106b75e17e820782855e25817fe2b1365e46d96b421791744bf9876ead6c565`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_release_profile_admission_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00071_reporting_preview_admission.sql

Original SHA-256: `c8f5db692b3c5cac36cfa6cc2118ecc7b86f4a31efae9b4ffc9c1c2b105c7ec4`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_job_payloads_job_kind_check`, `reporting_release_profile_admission_ck`, `reporting_render_profile_admission_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00072_reporting_template_metadata.sql

Original SHA-256: `3f96368ee11a7b033f1d91b3604f8de967adf3ced75aecb56ea03359cead95cc`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_render_template_metadata_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00073_party_envelope_clock_authority.sql

Original SHA-256: `b9a7970f13f038b72e54ee642b4b2c59d36f727366fb86c955d5d77f75700563`.

- Tables created/altered/dropped: `parties`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00074_reporting_failure_diagnostics.sql

Original SHA-256: `0a05ec253da959a7645592c1f7494b5b64fa248b5e3fa50fbf2936df9e8fbe43`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_render_failure_summary_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00075_reporting_frozen_source_facts.sql

Original SHA-256: `a977813ed0164c7a2b005ce642c9ddaefefe89d81919046e2fab75ced1f41a81`.

- Tables created/altered/dropped: `reporting_snapshots`.
- Constraints: `reporting_snapshot_identity_v1_ck`, `reporting_snapshot_identity_v2_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00076_reporting_profile_policy_cutover.sql

Original SHA-256: `20f802c99f37ef55c3f456dfa4c91a0c212ff60fb3391fcb2dfcbac36dd3de51`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_render_profile_admission_ck`, `reporting_render_profile_admission_v2_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00077_reporting_subject_identity_cutover.sql

Original SHA-256: `670e39825028a4bc3cfa780ad9d0ba6d228ed90ef5bb897ef1a2ff515bd16e5a`.

- Tables created/altered/dropped: `reporting_snapshots`.
- Constraints: `reporting_subject_identity_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00078_reporting_token_artifact_cutover.sql

Original SHA-256: `9851e95564c31950736fd76d92a0dbd26e07c1608e976c6f92851b7bca1ff9fb`.

- Tables created/altered/dropped: `reporting_composition_preview_outputs`, `reporting_releases`.
- Constraints: `reporting_preview_redaction_manifest_v3_ck`, `reporting_redaction_manifest_v3_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00079_reporting_structural_removal_cutover.sql

Original SHA-256: `cfba63b326363653e2ab16c0eebc667e0572d0dfb1d7787b84b4c96cfe65b551`.

- Tables created/altered/dropped: `reporting_composition_preview_outputs`, `reporting_releases`.
- Constraints: `reporting_preview_redaction_manifest_v3_ck`, `reporting_preview_redaction_manifest_v4_ck`, `reporting_redaction_manifest_v3_ck`, `reporting_redaction_manifest_v4_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00080_reporting_retained_field_policy.sql

Original SHA-256: `458675ae79e32aea1715bcbd3785a6c9044b2c128ddb595f7eb50fa011f44806`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_render_profile_revision_v3_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00081_reporting_retained_graph_sources.sql

Original SHA-256: `892a7384a9b22212b36514efdb1b70bd3231378481a138d0f46c3e46a0e86a65`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_graph_input_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00082_reporting_template_manifest.sql

Original SHA-256: `9e918cb4dde9b96e790f92bdeeb2622b481dbac1e79e6e62179cceb41314b081`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_template_manifest_v2_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00083_reporting_presentation_text_admission.sql

Original SHA-256: `792369d314709e8f34b376ce1bdb4024b52cd017bedf33ee4be2487bb93029a7`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_presentation_text_admission_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00084_reporting_source_index_admission.sql

Original SHA-256: `3c9e7bcf1b4ac32d90c035ad1a9b9d051f68d5ebc39255c10094dbf4a44497b1`.

- Tables created/altered/dropped: `reporting_job_payloads`, `reporting_snapshots`.
- Constraints: `reporting_presentation_text_admission_v1_ck`, `reporting_snapshot_identity_v2_ck`, `reporting_snapshot_identity_v3_ck`, `reporting_source_index_admission_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00085_reporting_core_record_admission.sql

Original SHA-256: `919a207e434e3ef151469ad8a0a05ad997ab46a68ee3f595d4b639c1c909b0b4`.

- Tables created/altered/dropped: `reporting_job_payloads`, `reporting_snapshots`.
- Constraints: `reporting_core_record_admission_v1_ck`, `reporting_snapshot_identity_v3_ck`, `reporting_snapshot_identity_v4_ck`, `reporting_source_index_admission_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00086_reporting_lifecycle_admission.sql

Original SHA-256: `a7714cc77aaf7a4a0bd37428edfda595e929b8b62008e72f5948f3d8a2ee88eb`.

- Tables created/altered/dropped: `reporting_job_payloads`, `reporting_snapshots`.
- Constraints: `reporting_core_record_admission_v1_ck`, `reporting_lifecycle_admission_v1_ck`, `reporting_snapshot_identity_v4_ck`, `reporting_snapshot_identity_v5_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00087_reporting_bundle_model_identity.sql

Original SHA-256: `49627ad46d291481f138be32bef3ae60fa916a4c726aff3bb32d11af9ed5bdf3`.

- Tables created/altered/dropped: `reporting_render_bundles`.
- Constraints: `reporting_bundle_model_identity_v3_ck`, `reporting_render_bundles_bundle_manifest_json_check`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00088_reporting_container_redaction.sql

Original SHA-256: `ac319f74d435ddf6bd91df2467df2abe7025557f9e86923320930413d0117419`.

- Tables created/altered/dropped: `reporting_composition_preview_outputs`, `reporting_job_payloads`, `reporting_releases`.
- Constraints: `reporting_preview_redaction_manifest_v4_ck`, `reporting_preview_redaction_manifest_v5_ck`, `reporting_redaction_manifest_v4_ck`, `reporting_redaction_manifest_v5_ck`, `reporting_render_profile_revision_v3_ck`, `reporting_render_profile_revision_v4_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00089_reporting_template_pack_admission.sql

Original SHA-256: `971d32c316f96b10502569c2d5738abff900367c7c5972f03f2a80314a2ce26b`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_template_manifest_admission_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00090_reporting_theme_toolchain_inventory.sql

Original SHA-256: `a88d2834131e5fc02f14a518cfad5b735d3daabfe835465dab816a708014f0fb`.

- Tables created/altered/dropped: `reporting_render_bundles`.
- Constraints: `reporting_toolchain_inventory_v2_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00091_reporting_structured_redaction.sql

Original SHA-256: `e710ac0923a4b1d6b9155494c2de0bcdc1a2ce8758eaae4ee4fa9b154c1fe50c`.

- Tables created/altered/dropped: `reporting_job_payloads`, `reporting_render_bundles`, `reporting_snapshots`.
- Constraints: `reporting_lifecycle_admission_v1_ck`, `reporting_render_profile_admission_v2_ck`, `reporting_render_profile_admission_v3_ck`, `reporting_render_profile_revision_v4_ck`, `reporting_render_profile_revision_v5_ck`, `reporting_snapshot_identity_v5_ck`, `reporting_snapshot_identity_v6_ck`, `reporting_structured_input_v1_ck`, `reporting_structured_redaction_v1_ck`, `reporting_template_manifest_v2_ck`, `reporting_template_manifest_v3_ck`, `reporting_toolchain_inventory_v2_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00092_reporting_snapshot_job_admission.sql

Original SHA-256: `690ff74fb9169e6435188d17b92cc8773ac8400f117678aaccc4ca2e71366d61`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_snapshot_admission_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00093_reporting_snapshot_source_capture.sql

Original SHA-256: `93a2f41d114be55c0d6fafb24b2ebd3f438b193410540cc069f2d312981286dc`.

- Tables created/altered/dropped: `reporting_snapshots`.
- Constraints: `reporting_snapshot_capture_v1_ck`, `reporting_snapshot_identity_v6_ck`, `reporting_subject_identity_v1_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00094_reporting_render_recipe.sql

Original SHA-256: `dca31b789a7745c55d29e23e53b569f929629439aa1441e783cbad6c69ff8f72`.

- Tables created/altered/dropped: `reporting_job_payloads`.
- Constraints: `reporting_render_profile_admission_v3_ck`, `reporting_render_profile_revision_v5_ck`, `reporting_render_recipe_v1_ck`, `reporting_render_template_metadata_ck`, `reporting_structured_input_v1_ck`, `reporting_template_manifest_v3_ck`.
- Privilege operations: 0; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.

#### 00095_evidence_blob_verification.sql

Original SHA-256: `6b667ddecfcca7b46ebbada60da97690ae4a8363bdf404e562dc1d45edeacd3e`.

- Tables created/altered/dropped: `evidence_blob_verifications`, `object_blobs`.
- Constraints: `evidence_blob_verifications_blob_fkey`, `evidence_blob_verifications_content_type_ck`, `evidence_blob_verifications_lease_fkey`, `evidence_blob_verifications_pkey`, `evidence_blob_verifications_sha256_ck`, `evidence_blob_verifications_size_ck`, `object_blobs_finalize_attempt_count_check`, `object_blobs_terminal_reason_check`.
- Indexes: `evidence_blob_verifications_lease_fk_idx`.
- Routines: `evidence_reject_cleanup_claimed_blob_update`.
- Triggers: `evidence_reject_cleanup_claimed_blob_update`.
- FK prerequisites: `evidence_object_upload_leases`, `object_blobs`.
- Privilege operations: 4; final grants are incorporated in the owning definitions.
- Non-DDL: no lasting fresh-install rows; any read-only cutover/preflight checks are replaced by the lineage gate. Function-body writes remain runtime behavior.
