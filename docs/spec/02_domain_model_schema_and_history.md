# Cartulary Normative Core 02: Domain Model, Schema, and History

## 1. Domain-model goals

Cartulary MUST separate:

- raw capture from canonical entities,
- incident data from optional reference packs and derived overlays,
- authoritative source state from denormalized workbook projections.

The domain model MUST preserve low-friction rough capture while enabling progressive normalization, typed relationships, revision history, rollback, and export derivation.

## 2. Core record types

The base profile MUST support the following core objects:

- **Incident**: workspace boundary.
- **Record envelope**: common identity, version, attribution, and delete state.
- **Timeline event**: primary capture unit.
- **Host**: canonical or stub device or host record.
- **Identity**: canonical or stub account or persona record.
- **Artifact**: structured text object including notes, queries, excerpts, exports, and lightweight indicator capture in the artifact-backed path.
- **Evidence record**: user-facing evidence envelope that may later reference a blob.
- **Object blob**: storage metadata for binary content.
- **Entity mention**: rough textual reference captured before canonical resolution.
- **Compromise assessment**: incident-scoped assessment record about a host or identity.
- **Record link**: typed relationship between two records.
- **Tag**: lightweight incident-scoped label.
- **Reference pack**: versioned optional vocabulary, framework, or enrichment dataset.
- **View schema**: contract for a built-in sheet or system view.
- **Saved view**: workbook configuration over a projection or view schema.

## 3. Record envelope contract

Every user-visible first-class record MUST own one record-envelope row.

The record envelope MUST include, at minimum:

- stable `record_id`,
- owning `incident_id`,
- `record_type`,
- creation attribution,
- update attribution,
- current `row_version`,
- soft-delete state.

The base record types MUST include, at minimum:

- `timeline_event`,
- `host`,
- `identity`,
- `artifact`,
- `evidence`,
- `assessment`.

## 4. Normalized versus flexible data

### 4.1 Normalized data

The following data classes MUST remain normalized structured state:

- incidents, users, and memberships,
- primary records,
- typed links,
- tags,
- revisions and change sets,
- blob metadata,
- canonical host and identity fields,
- compromise assessments,
- reference-pack manifests and type registries,
- view schemas and saved views,
- snapshot descriptors, canonical export-model metadata, versioned template-contract metadata, versioned redaction-profile metadata, redaction manifests, and artifact release records when the Snapshot and Reporting Extension Profile is implemented.

### 4.2 Flexible data

The following MAY remain flexible or extensible:

- `raw_capture` on timeline events,
- `custom_attrs` on major record types,
- free-text details and notes,
- low-frequency incident-specific metadata.

### 4.3 JSONB discipline

JSONB MAY be used for:

- raw import remnants,
- optional enrichment metadata,
- low-frequency custom fields,
- saved view layout configuration.

JSONB MUST NOT be the authoritative storage for:

- timestamps used for sorting,
- canonical names or primary identifiers,
- record relationships,
- evidence metadata required for retrieval,
- any field that requires reliable filtering or indexing at operational scale.

## 5. Partial and uncertain data

The implementation MUST permit rough and uncertain input.

At minimum:

- `occurred_at` MAY be null,
- summary MAY be null if another field or attachment exists,
- host and account text MAY remain unresolved,
- confidence MAY be unset,
- details MAY remain unstructured text.

The system MUST preserve original rough capture even after later normalization or resolution.

## 6. Mention, stub, and entity-origin contract

### 6.1 Separation invariant

An unresolved mention is **not** a low-quality entity.

An `entity_mention` is an observation tied to a source record and field. A stub entity is a host or identity record with its own `record_id` and lifecycle. The implementation MUST keep these object types separate.

### 6.2 Binding modes

Every entity-bearing field in a view contract, import mapping, or API contract MUST declare `entity_binding_mode` with one of the following values:

- **`mention_origin`**: the field captures an observed reference within another record. The write path MUST create `entity_mentions`. It MUST NOT implicitly create host or identity records.
- **`entity_origin`**: the field creates or updates a host or identity record directly. The write path MUST create or upsert a host or identity record. It MUST NOT synthesize `entity_mentions` unless the action is explicitly anchored to an existing mention.

Binding behavior MUST depend on the contract, not on the visible column header.

### 6.3 Required binding behavior by context

| Context | Binding mode | Required behavior |
| --- | --- | --- |
| Timeline, Notes, and other non-entity record cells that reference hosts or identities | `mention_origin` | MUST persist a raw `entity_mention`; MUST NOT auto-create a stub entity; MAY return candidate matches |
| Clipboard paste into non-entity sheets | `mention_origin` | MUST create one mention per observed cell value and source row; MUST NOT coalesce repeated mentions or auto-create stubs |
| XLSX, CSV, or API import mapped into non-entity records | `mention_origin` | MUST preserve mentions as mentions; MUST attach import provenance; MUST NOT auto-create stubs during ingest |
| Inspector action `Resolve` to existing entity | `mention_origin` | MUST resolve the selected mention to an existing entity; MUST keep the raw mention unchanged |
| Inspector action `Create host` or `Create identity` from a mention | `mention_origin` plus explicit create | MUST create a stub if no unique exact-match entity exists; MUST keep the mention; MUST resolve the selected mention to the created stub |
| Direct row creation or paste on Hosts or Identities sheets | `entity_origin` | MUST create or upsert a host or identity record; if identifiers are incomplete or unverified, the record MUST start in `stub` state |
| Entity-primary imports such as Systems/Hosts or Accounts/Identities mappings | `entity_origin` | MUST upsert an existing entity on a unique exact match; otherwise MUST create a stub; MUST preserve source values as provenance and aliases; MUST NOT auto-merge pre-existing entities |

### 6.4 Suggestion boundary

Automatic background matching MAY suggest candidates or pre-fill a resolution UI. It MUST NOT create stubs or merge entities without either:

- an explicit `entity_origin` contract, or
- an explicit analyst action.

Creating a stub from one mention MUST resolve only that selected mention by default. Bulk resolution of sibling mentions MUST be a separate explicit action.

## 7. Provenance requirements

### 7.1 Mention provenance

For every `entity_mention`, the implementation MUST persist:

- `source_record_id`,
- `entity_type`,
- `source_field_key`,
- `origin_kind`,
- `origin_locator`,
- `raw_text`,
- `normalized_text`,
- creator or ingest actor,
- creation timestamp,
- when resolved: `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`.

`origin_kind` MUST use a closed vocabulary equivalent to:

- `manual_entry`,
- `clipboard_paste`,
- `csv_import`,
- `xlsx_import`,
- `api_import`,
- `extraction`,
- `system`.

`origin_locator` MUST identify the source position deterministically, such as sheet/row/column for import or view/field for interactive entry.

### 7.2 File-based import provenance

When a record, mention, or alias is created or updated through file-based import, the implementation MUST persist provenance sufficient to identify the import source deterministically.

At minimum, file-based import provenance MUST include:

- `source_file_kind`,
- `source_content_sha256`,
- `parser_version`.

If the source is sheet-based or region-based, the provenance MUST also identify the selected sheet and rectangular region deterministically.

### 7.3 Entity provenance

For every host or identity record created outside direct entity sheets, the implementation MUST persist:

- `entity_origin`,
- an optional seed mention reference when created from a mention,
- seed identifier values as aliases or equivalent structured provenance,
- full `change_set` and revision lineage.

`entity_origin` MUST use a closed vocabulary equivalent to:

- `entity_sheet`,
- `entity_import`,
- `created_from_mention`,
- `system_upsert`.

## 8. Deduplication and auto-upsert rules

### 8.1 Mention deduplication

Repeated mentions with identical `normalized_text` MUST remain separate mention rows.

The implementation MAY group such mentions for review. It MUST NOT merge or coalesce them, because each mention is a distinct observation with distinct provenance.

### 8.2 Entity exact-match precedence

Entity deduplication is incident-scoped and applies only to active entities.

Exact-match selection MUST follow this precedence order:

- hosts: `aad_device_id`, then `fqdn`, then `hostname`,
- identities: `aad_object_id`, then `sid`, then `upn`, then `email`, then `sam_account_name`.

A unique exact match on one of these keys MUST reuse the existing active entity.

### 8.3 Suggestion boundary

Exact alias matches and fuzzy or trigram matches MAY be surfaced as suggestions. They MUST NOT:

- auto-resolve mentions,
- auto-create stubs,
- auto-merge entities.

If a create or update would assign one of the exact-match keys to more than one active entity, the operation MUST fail as a merge or conflict case rather than silently rekeying or merging entities.

## 9. Merge behavior

Entity merges MUST be explicit reviewer or admin actions.

The implementation MUST NOT auto-merge entities based only on alias similarity, repeated mention text, or fuzzy matching.

When two entities are merged:

- the survivor `record_id` MUST remain unchanged,
- the losing entity MUST remain as a historical row with state `merged` and `merged_into_record_id` set,
- active `entity_mentions.resolved_record_id`, active `record_links`, active assessments, and active tags pointing at the losing entity MUST be repointed to the survivor in the same `change_set`, or otherwise tombstoned and recreated deterministically,
- duplicate links or tags created by repointing MUST be deduplicated without losing revision history,
- non-conflicting alias values and seed identifiers from the losing entity SHOULD be copied to the survivor,
- raw mention text MUST NOT be rewritten or deleted.

## 10. Notes, artifacts, indicators, and assessments

### 10.1 Notes

Ad hoc notes MUST be modeled as artifacts with `artifact_type='note'`.

### 10.2 Indicator contract

The base profile MAY keep indicator storage artifact-backed. Regardless of storage realization, the system MUST expose a stable indicator projection or API contract with fields equivalent to:

- `indicator_type`,
- `indicator_value`,
- `normalized_value`,
- `defanged_value`,
- optional hash algorithm and hash value,
- `stix_pattern`,
- supporting link counts.

This contract MUST remain stable across import, export, reporting, and any later promotion to a dedicated indicator table.

### 10.3 Compromise assessments

Compromise state MUST be modeled as incident-scoped assessment history attached to a host or identity rather than as a static property on the entity.

Each assessment MUST carry:

- state,
- timestamp,
- assessor,
- confidence,
- rationale,
- optional supporting record links.

### 10.4 Analyst-work tracking

Tags and notes are the starting point for analyst-work tracking. Explicit objects for tasks, hypotheses, decisions, and ownership are future areas preserved by Appendix E and are not current base-profile requirements.

### 10.5 Snapshot and reporting extension objects

If the Snapshot and Reporting Extension Profile is implemented, the domain model MUST also support structured metadata for:

- immutable snapshot descriptors,
- canonical export-model fields and blocks with stable export-model paths,
- versioned redaction profiles,
- versioned template contracts,
- redaction manifests,
- artifact release records.

At minimum, an immutable snapshot descriptor MUST persist the release tuple defined by Core 01 §10.2.

Each exportable field or block in the canonical export model MUST persist:

- a stable export-model path,
- exactly one `content_class`,
- deterministic order metadata sufficient to reproduce export ordering,
- when `content_class='curated_narrative'`, zero or more `support_refs[]`.

When recipient-specific reporting is implemented, each exportable field or block MUST also be able to carry zero or more stable incident-local `disclosure_partition_refs[]`. Each `disclosure_partition_ref` names an export-only withholding boundary such as an affected party or other approved recipient partition. `disclosure_partition_refs[]` MUST be stored as snapshot or export metadata only. They MUST NOT change live workspace authorization.

A release record MUST bind, at minimum:

- `snapshot_id`,
- `template_id`,
- `template_version`,
- `redaction_profile_id`,
- `redaction_profile_version`,
- `output_kind`,
- `release_scope`,
- `output_sha256`.

If approval state is stored, it MUST bind to the release record rather than to mutable incident rows.

## 11. Type registries and view contracts

Type and icon choices MUST be driven by stable registry keys rather than hard-coded display text.

Host types, evidence types, and similar display taxonomies MUST resolve through versioned registries keyed by stable IDs, with display labels, categories, icon keys, and optional local overrides.

Entity-bearing columns and import mappings MUST carry stable `field_key` values and `entity_binding_mode` so that mention-versus-entity behavior never depends on a visible header such as `Host`, `User`, or `Account`.

## 12. Typed relationships

The implementation MUST provide a generic typed relationship store equivalent to `record_links`.

The relationship vocabulary MUST be controlled by application code or equivalent authoritative configuration. It MUST NOT be free-text user input in the UI.

The relationship vocabulary MUST support, at minimum, edge types equivalent to:

- `observed_on_host`,
- `observed_as_identity`,
- `attached_evidence`,
- `references_artifact`,
- `derived_from`,
- `merged_into`.

A timeline event MUST be able to relate to:

- hosts and identities via typed record links,
- unresolved host and account strings via entity mentions,
- notes and other artifacts via typed record links,
- evidence via typed record links to evidence records.

## 13. Evidence and object metadata

The implementation MUST store authoritative object metadata in structured storage, including:

- blob identifier,
- storage backend,
- bucket and key,
- filename,
- content type,
- size,
- hash,
- upload state.

The binary content MUST live in object storage. Structured storage MUST remain the authoritative pointer and metadata source.

The evidence model MUST support both:

- requested or pending evidence with no blob yet attached,
- received or available evidence with custody events and optional blob linkage.

An implementation MAY persist a non-authoritative `sensitivity_class` on evidence records and MAY project that label onto exportable evidence-derived blocks. In the base profile, `sensitivity_class` is metadata for preview labeling, default export-redaction behavior, and audit or reporting only. It MUST NOT change incident authorization or hide live workspace content.

## 14. Schema invariants

### 14.1 Mention and provenance fields

The schema MUST support:

- `source_field_key`,
- `origin_kind`,
- `origin_locator`,
- file-based import provenance fields for source file kind, source content hash, parser version, and selected sheet or region locator,
- resolution metadata on mentions,
- `entity_origin` and seed provenance on host and identity records,
- `entity_binding_mode` in view write-back contracts and import mappings.

### 14.2 Rollback granularity substrate

A conformant history schema MUST include both:

- immutable `change_set` attribution records, and
- mutation-entry records at target granularity.

`record_revisions` MAY retain row snapshots such as `before_json` and `after_json` for audit and whole-row restore. They MUST NOT be the sole rollback substrate.

### 14.3 Mutation targets

The mutation-entry model MUST support, at minimum:

- row-field edits,
- record links,
- record tags,
- entity mentions,
- evidence associations,
- merge and repoint fan-out.

Stable mutation targets MUST use deterministic serialization. Composite targets MUST serialize canonically, for example `record_tag:<record_id>:<tag_id>`.

### 14.4 Soft delete

User-visible records MUST be soft-deletable. Links MUST be soft-deletable. Revisions MUST be append-only in normal operation.

Blobs MAY be hard-deleted only through an explicit administrative purge or retention workflow.

### 14.5 Snapshot and reporting extension fields

If the Snapshot and Reporting Extension Profile is implemented, the schema MUST support:

- stable export-model paths,
- `content_class`,
- stable incident-local `disclosure_partition_refs[]` on exportable fields or blocks,
- `support_refs[]` on `curated_narrative` blocks,
- versioned template identifiers,
- versioned redaction-profile identifiers,
- `release_scope`,
- `export_model_sha256`,
- `output_sha256`,
- redaction-manifest entries keyed by stable export-model path and rule identifier.

## 15. History and rollback

### 15.1 Attribution unit

For each committed action, the implementation MUST create one immutable `change_set`.

A `change_set` MUST record, at minimum:

- incident,
- actor,
- timestamp,
- source,
- optional reason,
- optional client transaction identifier.

### 15.2 Mutation-entry unit

Each committed action MUST also create one or more reversible mutation entries.

Each mutation entry MUST record:

- target kind,
- stable target identifier,
- operation kind,
- deterministic order within the parent `change_set`,
- pre-change and post-change version identifiers,
- reversible before/after values or an equivalent reversible patch.

### 15.3 Reconstruction requirement

The history model MUST preserve enough information to reconstruct:

- the full row snapshot at any revision,
- the exact field, link, mention, tag, evidence, or merge delta introduced by a `change_set`.

Projection tables MUST NOT be authoritative history.

### 15.4 Entity-merge history

When entities are merged, the merge `change_set` MUST preserve enough detail to reconstruct both:

- the pre-merge graph,
- the post-merge graph.

Live mention resolutions and live links MUST NOT continue to point at a merged-away record after the merge `change_set` commits.

## 16. Search and indexing expectations

The source artifact included a concrete indexing strategy. The exact SQL realization MAY vary, but a conformant implementation SHOULD provide equivalent lookup behavior for:

- filtering and sorting by incident and key timeline dimensions,
- full-text search using a configuration appropriate for IR tokens rather than English stemming,
- link traversal by source and destination,
- fuzzy matching over alias and unresolved mention text,
- array or equivalent containment lookups for denormalized tag and label sets.

## 17. Domain invariants

A conformant implementation MUST preserve all of the following invariants:

1. mentions and entities are different object types,
2. repeated mentions remain distinct observations,
3. exact-match entity reuse follows a deterministic precedence,
4. merges are explicit and auditable,
5. notes are artifacts,
6. assessments append rather than overwrite,
7. relationship semantics are typed,
8. history is reversible at mutation-entry granularity,
9. projection state is derived rather than authoritative.
