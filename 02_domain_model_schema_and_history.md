# Cartulary Normative Core 02: Domain Model, Schema, and History

## 1. Domain-model goals

**REQ-02-001**
Cartulary MUST separate:

- raw capture from canonical entities,
- incident data from optional reference packs and derived overlays,
- authoritative source state from denormalized workbook projections.
Profiles: base, reference_pack
Verified by: AC-231, AC-234

**REQ-02-002**
The domain model MUST preserve low-friction rough capture while enabling progressive normalization, typed relationships, revision history, rollback, and export derivation.
Profiles: base
Verified by: AC-231

## 2. Core record types

**REQ-02-003**
The base profile MUST support the following core objects:

- **Incident**: workspace boundary.
- **Record envelope**: common identity, version, attribution, and delete state.
- **Timeline event**: primary capture unit.
- **Host**: canonical or stub device or host record.
- **Identity**: canonical or stub account or persona record.
- **Artifact**: structured text object including notes, queries, excerpts, exports, hypotheses, communications logs, handoffs, status reviews, lessons, and other source-preserving analyst material.
- **Task request**: owned unit of work or request with lifecycle state.
- **Decision**: owned incident-scoped rationale-bearing decision record.
- **Indicator**: canonical incident-scoped linkable value or pattern.
- **Evidence record**: user-facing evidence envelope that may later reference a blob.
- **Object blob**: storage metadata for binary content.
- **Entity mention**: rough textual reference captured before canonical resolution.
- **Indicator observation**: source-bound occurrence of an indicator value or pattern inside another record field.
- **Indicator lifecycle interval**: append-only incident-scoped state window attached to a canonical indicator.
- **Compromise assessment**: incident-scoped assessment record about a host or identity.
- **Record link**: typed relationship between two records.
- **Tag**: lightweight incident-scoped label.
- **Reference pack**: versioned optional vocabulary, framework, or enrichment dataset.
- **View schema**: contract for a built-in sheet or system view.
- **Saved view**: workbook configuration over a projection or view schema.
- **User workbook preferences**: incident-scoped per-user startup surface preference.
- **Incident workbook preferences**: incident-wide fallback startup surface preference.
Profiles: base, reference_pack
Verified by: AC-231, AC-234

## 3. Record envelope contract

**REQ-02-004**
Every user-visible first-class record MUST own one record-envelope row.
Profiles: base
Verified by: AC-231

**REQ-02-005**
The record envelope MUST include, at minimum:

- stable `record_id`,
- owning `incident_id`,
- `record_type`,
- creation attribution,
- update attribution,
- current `row_version`,
- soft-delete state.
Profiles: base
Verified by: AC-231

**REQ-02-006**
The base record types MUST include, at minimum:

- `timeline_event`,
- `host`,
- `identity`,
- `indicator`,
- `artifact`,
- `task_request`,
- `decision`,
- `evidence`,
- `assessment`.
Profiles: base
Verified by: AC-231

**REQ-02-007**
Internal users, auth-provider identities, sessions, and incident memberships are authoritative normalized administrative state but are not user-visible first-class incident records. They MUST NOT consume `record_id` values, MUST NOT be modeled as generic record-envelope rows, and MUST NOT be the target of workbook row-mutation routes.
Profiles: base
Verified by: AC-231

**REQ-02-008**
User-account and incident-membership mutations MUST remain auditable, but they MUST use a dedicated deployment-local administrative audit substrate rather than incident `change_set` or `record_revisions` rows.
Profiles: base
Verified by: AC-231

## 4. Normalized versus flexible data

### 4.1 Normalized data

**REQ-02-009**
The following data classes MUST remain normalized structured state:

- incidents, users, and memberships,
- primary records,
- typed links,
- tags,
- revisions and change sets,
- blob metadata,
- canonical host, identity, and indicator fields,
- indicator observations and indicator lifecycle intervals,
- compromise assessments,
- reference-pack manifests, activation state, attestation metadata, and type registries,
- view schemas, saved views, and workbook preference objects,
- promoted operational fields on incident, host, identity, task-request, indicator, and evidence surfaces as defined in §4.4 and §4.5,
- snapshot descriptors, canonical export-model metadata, versioned template-contract metadata, versioned redaction-profile metadata, redaction manifests, and artifact release records when the Snapshot and Reporting Extension Profile is implemented.
Profiles: base, snapshot_reporting, reference_pack
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-231, AC-233, AC-234

### 4.2 Flexible data

The following MAY remain flexible or extensible:

- `raw_capture` on timeline events,
- `custom_attrs` on major record types,
- free-text details and notes,
- low-frequency incident-specific metadata that is not promoted by §4.4 and §4.5.

### 4.3 JSONB discipline

JSONB MAY be used for:

- raw import remnants,
- optional enrichment metadata,
- low-frequency custom fields,
- normalized saved-view query state,
- saved-view layout configuration.

**REQ-02-010**
JSONB MUST NOT be the authoritative storage for:

- timestamps used for sorting,
- canonical names or primary identifiers,
- record relationships,
- evidence metadata required for retrieval,
- saved-view identity, scope, ownership, or startup/default surface selection,
- any field that requires reliable filtering or indexing at operational scale.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-231

Saved-view `query_json` MAY remain in JSONB only when it is normalized against the owning `view_schema_id` and uses stable `field_key` values, ordered sort/filter entries, optional `group_by`, and normalized scalar values. Saved-view `layout_json` MAY describe presentation only. Neither `query_json` nor `layout_json` MAY be the authoritative source for saved-view identity, scope, ownership, authorization, or startup/default surface selection.

**REQ-02-011**
Multi-valued relationships MUST NOT remain only in `custom_attrs` or other JSONB fields. They MUST use typed link rows or dedicated child tables. Scalar convenience projections MAY exist, but they MUST NOT be the authoritative relationship store.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-231

### 4.4 Promotion rule for recurrent incident-specific fields

**REQ-02-012**
A field MUST be promoted from `custom_attrs` or other JSONB storage to first-class structured state when both of the following are true:

- the field recurs across independent SoD-style incident workbooks or equivalent real-world incident trackers,
- analysts need the field for deterministic sort, filter, join, lifecycle tracking, dedupe, queueing, or reporting.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-231

A field SHOULD remain flexible when it is low-frequency, framework-specific, integration-specific, presentation-only, or better modeled as a typed link or child record.

**REQ-02-013**
A promoted field MUST be available as named structured state on the authoritative record model or on a deterministic projection derived from authoritative structured inputs. It MUST NOT require JSON parsing of `custom_attrs`, scans of raw note text, or scans of raw evidence text on the hot path.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-231

### 4.5 Current-profile promoted field sets

**REQ-02-014**
The current profile closes the incident-specific custom-metadata question with the following minimum promoted field sets. An implementation MAY add more structured fields. It MUST NOT demote these fields into JSONB-only storage.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-015**
The incident record MUST expose, as structured state:

- `tlp`,
- `current_phase`,
- `primary_external_case_ref`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-016**
An incident MUST also persist a first-class `incident_key`. The deployment MUST enforce deployment-wide uniqueness on the canonical `incident_key` form produced by trimming leading and trailing Unicode whitespace and applying Unicode NFC normalization. The implementation MAY store that canonical form explicitly or derive it through an equivalent deterministic uniqueness mechanism.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

Core 01 §3.3.5.3 owns the public incident resource and create contract. That contract also fixes `title`, `description`, `status`, `severity`, attribution fields, timestamps, `incident_version`, and `closed_at` as first-class incident state. This section governs only the additional incident-specific operational fields that were open in the source artifact.

**REQ-02-017**
The `task_request` model MUST expose, as structured state:

- `priority`,
- `task_kind`,
- optional `workstream`,
- optional `due_at`,
- optional `requester_party_text`,
- optional `blocked_reason`,
- optional `completed_at`,
- optional `external_ticket_ref`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-018**
The current `host` record type is the structured home for recurrent asset-scoping fields. A host record MUST expose, as structured state:

- `location`,
- `os_platform`,
- `business_owner`,
- `criticality`,
- `containment_status`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-019**
An `identity` record MUST expose, as structured state:

- `privilege_level`,
- `mfa_state`,
- `reset_status`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-020**
The canonical indicator model and indicator system-view contract MUST expose, as structured state:

- observation-derived `first_observed_at`,
- observation-derived `last_observed_at`,
- `defanged_value` when applicable,
- optional `hash_algorithm`,
- optional `hash_value`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-021**
An evidence record or its deterministic joined projection from authoritative object metadata MUST expose, as structured state:

- `requested_at`,
- `received_at`,
- `storage_ref`,
- `blob_hash`,
- `collector_party_text`,
- `source_party_text`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-022**
The current profile does not standardize a standalone incident-contact record. Until such a model exists, requester, collector, and source-party promotion MUST use first-class structured text fields or an equivalent stable party reference. They MUST NOT live only in `custom_attrs`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

**REQ-02-023**
If the implementation exposes structured findings, communications logs, investigative queries, or forensic keywords as dedicated surfaces or typed artifact subtypes, their defining fields MUST also be structured state rather than JSON-only payload:

- findings: `state`, `confidence_score`, `owner_user_id`, optional `closed_at`,
- communications logs: `comm_type`, `audience`,
- investigative queries: `query_id`, `platform`, `purpose`, `query_text`, `created_by_user_id`,
- forensic keywords: `keyword_id`, `pattern`, `reason`.
Profiles: base
Verified by: AC-097, AC-098, AC-099, AC-100, AC-101, AC-211, AC-212, AC-213, AC-214, AC-231

The following SHOULD remain flexible or separately modeled unless a later profile standardizes them:

- framework-specific overlays such as ATT&CK, D3FEND, VERIS, kill-chain, or procedure mapping,
- threat-intel enrichment such as `asn`, `geo`, or `reputation`,
- campaign, attribution, or reporting-specific metadata such as `intrusion_set`, `threat_actor_name`, `target_country`, or `target_sector`,
- presentation-only flags such as `visual`, `pin`, or similar UI-state markers,
- long governance or free-text metadata such as bridge details, legal-hold narrative, confidentiality narrative, working-hours text, or notification-scope text.

## 5. Partial and uncertain data

**REQ-02-024**
The implementation MUST permit rough and uncertain input.
Profiles: base
Verified by: AC-231

At minimum:

- `occurred_at` MAY be null,
- summary MAY be null if another field or attachment exists,
- host and account text MAY remain unresolved,
- confidence MAY be unset,
- details MAY remain unstructured text,
- `source_text` MAY remain unstructured text when a timeline event captures verbatim or excerpted source material.

**REQ-02-025**
The system MUST preserve original rough capture even after later normalization or resolution.
Profiles: base
Verified by: AC-231

## 6. Mention, stub, and entity-origin contract

### 6.1 Separation invariant

An unresolved mention is **not** a low-quality entity.

**REQ-02-026**
An `entity_mention` is an observation tied to a source record and field. A stub entity is a host or identity record with its own `record_id` and lifecycle. The implementation MUST keep these object types separate.
Profiles: base
Verified by: AC-019, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

**REQ-02-027**
The same source-observation-versus-canonical-record separation MUST apply to indicators. An `indicator_observation` is a source-bound occurrence tied to a record and field. A canonical `indicator` is a first-class record with its own identity and lifecycle.
Profiles: base
Verified by: AC-019, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

### 6.2 Binding modes

**REQ-02-028**
Every entity-bearing field in a view contract, import mapping, or API contract MUST declare `entity_binding_mode` with one of the following values:

- **`mention_origin`**: the field captures an observed reference within another record. The write path MUST create `entity_mentions`. It MUST NOT implicitly create host or identity records.
- **`entity_origin`**: the field creates or updates a host or identity record directly. The write path MUST create or upsert a host or identity record. It MUST NOT synthesize `entity_mentions` unless the action is explicitly anchored to an existing mention.
Profiles: base
Verified by: AC-118, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

**REQ-02-029**
Binding behavior MUST depend on the contract, not on the visible column header.
Profiles: base
Verified by: AC-118, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

### 6.3 Required binding behavior by context

| Context | Binding mode | Required behavior | Requirement ID | Profiles | Verified by |
| --- | --- | --- | --- | --- | --- |
| Timeline, Notes, and other non-entity record cells that reference hosts or identities | `mention_origin` | MUST persist a raw `entity_mention`; MUST NOT auto-create a stub entity; MAY return candidate matches | REQ-02-030 | base | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231 |
| Clipboard paste into non-entity sheets | `mention_origin` | MUST create one mention per observed cell value and source row; MUST NOT coalesce repeated mentions or auto-create stubs | REQ-02-031 | base | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231 |
| XLSX, CSV, or API import mapped into non-entity records | `mention_origin` | MUST preserve mentions as mentions; MUST attach import provenance; MUST NOT auto-create stubs during ingest | REQ-02-032 | base, import | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231, AC-232 |
| Inspector action `Resolve` to existing entity | `mention_origin` | MUST resolve the selected mention to an existing entity; MUST keep the raw mention unchanged | REQ-02-033 | base | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231 |
| Inspector action `Create host` or `Create identity` from a mention | `mention_origin` plus explicit create | MUST create a stub if no unique exact-match entity exists; MUST keep the mention; MUST resolve the selected mention to the created stub | REQ-02-034 | base | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231 |
| Direct row creation or paste on Hosts or Identities sheets | `entity_origin` | MUST create or upsert a host or identity record; if identifiers are incomplete or unverified, the record MUST start in `stub` state | REQ-02-035 | base | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231 |
| Entity-primary imports such as Systems/Hosts or Accounts/Identities mappings | `entity_origin` | MUST upsert an existing entity on a unique exact match; otherwise MUST create a stub; MUST preserve source values as provenance and aliases; MUST NOT auto-merge pre-existing entities | REQ-02-036 | base | AC-019, AC-020, AC-022, AC-028, AC-029, AC-188, AC-189, AC-190, AC-201, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231 |

### 6.4 Suggestion boundary

**REQ-02-037**
Automatic background matching MAY suggest candidates or pre-fill a resolution UI. It MUST NOT create stubs or merge entities without either:

- an explicit `entity_origin` contract, or
- an explicit analyst action.
Profiles: base
Verified by: AC-020, AC-028, AC-029, AC-231

**REQ-02-038**
Creating a stub from one mention MUST resolve only that selected mention by default. Bulk resolution of sibling mentions MUST be a separate explicit action.
Profiles: base
Verified by: AC-020, AC-028, AC-029, AC-231

### 6.5 Entity-mention lifecycle

**REQ-02-039**
`entity_mentions.resolution_status` MUST use the closed vocabulary `unresolved`, `resolved`, and `dismissed`.
Profiles: base
Verified by: AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-02-040**
- **`unresolved`**: the mention remains a source-bound observation awaiting analyst action. It participates in unresolved-mention flags and resolution queues.
- **`resolved`**: the mention remains a source-bound observation bound to an existing entity through resolution metadata such as `resolved_record_id`. The raw mention and normalized mention text remain preserved alongside the canonical link.
- **`dismissed`**: the mention remains preserved for provenance and history, but it is excluded from active resolution work. A dismissed mention MUST retain `raw_text`, `normalized_text`, stable row identity, source locator, and other provenance, MUST clear active resolution metadata such as `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`, and MUST NOT contribute to active relationship-cell values or unresolved-mention derived flags.
Profiles: base
Verified by: AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-02-041**
Ordinary user-facing restore of a dismissed mention MUST return it to `unresolved`. Recovering the exact pre-dismiss state, including any prior resolved target, MUST be handled by rollback rather than ordinary restore.
Profiles: base
Verified by: AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

## 7. Provenance requirements

### 7.1 Mention provenance

**REQ-02-042**
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
Profiles: base
Verified by: AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

**REQ-02-043**
`origin_kind` MUST use the exact closed vocabulary defined in §18.
Profiles: base
Verified by: AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

**REQ-02-044**
`origin_locator` MUST identify the source position deterministically, such as sheet/row/column for import or view/field for interactive entry.
Profiles: base
Verified by: AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-231

### 7.2 File-based import provenance

**REQ-02-045**
When a record, mention, indicator observation, or alias is created or updated through file-based import, the implementation MUST persist provenance sufficient to identify the import source deterministically.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-046**
At minimum, file-based import provenance MUST include:

- `source_file_kind`,
- `source_content_sha256`,
- `parser_version`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-047**
If the source is sheet-based or region-based, the provenance MUST also identify the selected sheet and rectangular region deterministically.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

### 7.2.1 Import-session, unit, and mapping identity

**REQ-02-048**
When file-based import is implemented, the implementation MUST persist enough metadata to reconstruct which uploaded bytes, discovered source unit, and operator-approved mapping produced a given record, mention, indicator observation, or alias.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-049**
At minimum, file-based import-created or file-based import-updated data MUST be attributable to:

- `import_session_id`,
- `import_unit_id`,
- `mapping_fingerprint`,
- `parser_profile_id`,
- `source_file_kind`,
- `source_content_sha256`,
- `parser_version`,
- `origin_locator`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-050**
For imported rows, mentions, aliases, and indicator observations, `origin_locator` MUST be derivable from the canonical import-unit locator plus deterministic row and column coordinates within that unit.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-051**
An `import_unit` MUST have a semantic identity equal to the tuple `source_content_sha256 + canonical_locator + parser_version`. If the implementation stores one derived key for that identity, it MUST compute that key as the SHA-256 of the canonical JSON serialization of `{source_content_sha256, locator, parser_version}`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-052**
`mapping_fingerprint` MUST identify the operator-approved header-to-field mapping plan deterministically. The fingerprint input MUST include, at minimum, mapping contract version, target view schema identifier, header-row reference, data-start-row reference, unknown-column policy, and one ordered entry per source column containing source ordinal, raw imported header text, target field key, `entity_binding_mode`, and any declared transform or empty-value policy. The implementation MUST serialize that mapping input canonically with lexicographically sorted object keys and source columns ordered by `source_column_ordinal`, then hash the result as lower-hex SHA-256.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

**REQ-02-053**
Preserved unknown columns from file-based import MUST NOT be keyed only by visible header text. The preserved structure MUST retain, at minimum, `import_unit_id`, `source_column_ordinal`, `source_header_text`, and raw value for each retained cell or row value. Duplicate visible headers MUST remain distinguishable by ordinal and source unit.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-067, AC-232

### 7.3 Entity provenance

**REQ-02-054**
For every host or identity record created outside direct entity sheets, the implementation MUST persist:

- `entity_origin`,
- an optional seed mention reference when created from a mention,
- seed identifier values as aliases or equivalent structured provenance,
- full `change_set` and revision lineage.
Profiles: base
Verified by: AC-186, AC-209, AC-231

**REQ-02-055**
`entity_origin` MUST use the exact closed vocabulary defined in §18.
Profiles: base
Verified by: AC-186, AC-209, AC-231

### 7.4 Indicator observation provenance

**REQ-02-056**
For every `indicator_observation`, the implementation MUST persist:

- `source_record_id`,
- `source_field_key`,
- `origin_kind`,
- `origin_locator`,
- `observed_text`,
- optional parsed `indicator_type`,
- optional `normalized_candidate`,
- when the source is inline text, a deterministic span or selection locator,
- creator or ingest actor,
- creation timestamp,
- when resolved: `resolved_indicator_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`,
- when machine-assisted: extraction method or parser version and confidence.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-057**
`origin_kind` MUST use the exact closed vocabulary defined in §18.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

## 8. Deduplication and auto-upsert rules

### 8.1 Mention deduplication

**REQ-02-058**
Repeated mentions with identical `normalized_text` MUST remain separate mention rows.
Profiles: base
Verified by: AC-021, AC-028, AC-188, AC-231

**REQ-02-059**
The implementation MAY group such mentions for review. It MUST NOT merge or coalesce them, because each mention is a distinct observation with distinct provenance.
Profiles: base
Verified by: AC-021, AC-028, AC-188, AC-231

### 8.2 Entity exact-match precedence

Entity deduplication is incident-scoped and applies only to active entities.

**REQ-02-060**
Exact-match selection MUST follow this precedence order:

- hosts: `aad_device_id`, then `fqdn`, then `hostname`,
- identities: `aad_object_id`, then `sid`, then `upn`, then `email`, then `sam_account_name`.
Profiles: base
Verified by: AC-022, AC-028, AC-186, AC-187, AC-231

**REQ-02-061**
A unique exact match on one of these keys MUST reuse the existing active entity.
Profiles: base
Verified by: AC-022, AC-028, AC-186, AC-187, AC-231

### 8.3 Suggestion boundary

**REQ-02-062**
Exact alias matches and fuzzy or trigram matches MAY be surfaced as suggestions. They MUST NOT:

- auto-resolve mentions,
- auto-create stubs,
- auto-merge entities.
Profiles: base
Verified by: AC-028, AC-029, AC-231

**REQ-02-063**
If a create or update would assign one of the exact-match keys to more than one active entity, the operation MUST fail as a merge or conflict case rather than silently rekeying or merging entities.
Profiles: base
Verified by: AC-028, AC-029, AC-231

## 9. Merge behavior

**REQ-02-064**
Entity merges MUST be explicit reviewer or admin actions.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-02-065**
The implementation MUST NOT auto-merge entities based only on alias similarity, repeated mention text, or fuzzy matching.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

**REQ-02-066**
When two entities are merged:

- the survivor `record_id` MUST remain unchanged,
- the losing entity MUST remain as a historical row with state `merged` and `merged_into_record_id` set,
- active `entity_mentions.resolved_record_id`, active `record_links`, active assessments, and active tags pointing at the losing entity MUST be repointed to the survivor in the same `change_set`, or otherwise tombstoned and recreated deterministically,
- duplicate links or tags created by repointing MUST be deduplicated without losing revision history,
- non-conflicting alias values and seed identifiers from the losing entity SHOULD be copied to the survivor,
- raw mention text MUST NOT be rewritten or deleted.
Profiles: base
Verified by: AC-023, AC-186, AC-187, AC-209, AC-231

## 10. Notes, artifacts, indicators, and assessments

### 10.1 Notes

**REQ-02-067**
Ad hoc notes MUST be modeled as artifacts with `artifact_type='note'`.
Profiles: base
Verified by: AC-068, AC-070, AC-089, AC-112, AC-185, AC-231

**REQ-02-068**
The base profile MUST NOT introduce a Notes-specific source table or Notes-only persistence model. Standalone note creation from the built-in Notes sheet and contextual `add linked note` actions MUST create the same underlying artifact record shape and use the same revision, tag, and link semantics.
Profiles: base
Verified by: AC-068, AC-070, AC-089, AC-112, AC-185, AC-231

**REQ-02-069**
A note artifact MUST be able to persist, at minimum, fields equivalent to `title` and `body`, plus generic tags through `record_tags` or an equivalent shared tag mechanism.
Profiles: base
Verified by: AC-068, AC-070, AC-089, AC-112, AC-185, AC-231

**REQ-02-070**
Lightweight free-text fields MAY remain on timeline events, hosts, identities, evidence, or other records. Text that requires standalone history, tags, search, or reuse across records MUST be modeled as an artifact with `artifact_type='note'`.
Profiles: base
Verified by: AC-068, AC-070, AC-089, AC-112, AC-185, AC-231

**REQ-02-071**
Analyst-work objects other than ad hoc notes MUST use either a distinct `artifact_type` or a distinct first-class `record_type`. In the current profile, `task_request` and `decision` are first-class `record_type`s, while `comm_log`, `handoff`, `status_review`, `lesson`, and current `hypothesis` tracking remain artifact-backed. None of these objects MAY overload `artifact_type='note'`.
Profiles: base
Verified by: AC-068, AC-070, AC-089, AC-112, AC-185, AC-231

### 10.2 Indicator contract

**REQ-02-072**
The base profile MUST model canonical indicators as first-class `indicator` records.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-073**
A canonical indicator record MUST be incident-scoped and MUST own, at minimum:

- `indicator_type` as a registry key,
- `value_kind` from the exact closed vocabulary defined in §18,
- a canonical display value,
- `normalized_value` when applicable,
- a deterministic incident-scoped dedupe key,
- `defanged_value` when applicable,
- optional hash algorithm and hash value,
- optional `stix_pattern`,
- current `row_version` and revision lineage.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-074**
Canonical indicator identity MUST be derived from the incident plus the deterministic type-specific dedupe key or an equivalent stable canonical identity. In the current profile, the stored field inputs to that canonical identity are `indicator_type`, `value_kind`, the canonical display value, and `normalized_value` when applicable, plus the pair `hash_algorithm` and `hash_value` when both are populated and incorporated into the canonical dedupe key for the row. `defanged_value` MAY be stored for presentation or export, and `stix_pattern` MAY be stored for interoperability or export. Neither `defanged_value` nor `stix_pattern` MAY be the source of truth for canonical identity. Any later additional stored identity-basis field MUST be introduced explicitly by a versioned schema change.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-075**
A canonical indicator record is not the raw source occurrence. Source-bound occurrences inside timeline, artifact, note, evidence, or other record fields MUST be stored separately as `indicator_observation` rows or an equivalent structured observation model. `indicator_observation` rows MUST remain distinct from generic `record_links`.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-076**
Every `indicator_observation` MUST bind to a source record and field and MUST preserve the raw observed value or text span. It MUST be able to store, at minimum, the observed text, parsed indicator type guess, normalized parse output when available, deterministic source locator, resolution status, optional resolved indicator reference, and attribution timestamps. Repeated identical observed values across different source rows MUST remain distinct observations with distinct provenance.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-077**
A source value such as a hostname or MAC address MAY participate in both the host or identity model and the indicator model. The implementation MUST NOT force one object type to subsume the other. A single source span MAY therefore yield both an `entity_mention` and an `indicator_observation`, with separate later resolution to a host or identity record and a canonical indicator record.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-078**
The base profile MUST preserve raw source text when capturing or resolving indicator observations. It MUST NOT require dedicated IOC columns on Timeline or other non-indicator sheets.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-079**
Indicator lifecycle state MUST be modeled as append-only incident-scoped history attached to the canonical indicator, not as an overwrite of the indicator row and not as a rewrite of the source observation. Each lifecycle interval MUST carry, at minimum:

- `indicator_record_id`,
- incident-scoped state or disposition,
- `valid_from`,
- `valid_to`,
- confidence,
- rationale,
- optional supporting record links,
- assessor,
- timestamp.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-080**
`first_observed_at` and `last_observed_at` are derived from indicator observations. They MUST remain distinct from lifecycle `valid_from` and `valid_to`.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-081**
The system MUST expose a stable indicator system-view and API contract over canonical indicators with fields equivalent to:

- `indicator_type`,
- `value_kind`,
- canonical indicator value,
- `normalized_value`,
- deterministic dedupe key,
- `defanged_value`,
- optional hash algorithm and hash value,
- `stix_pattern`,
- observation-derived first and last observed timestamps,
- lifecycle summary fields,
- supporting link counts.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

**REQ-02-082**
This contract MUST remain stable across import, export, reporting, and future storage evolution.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-122, AC-231

### 10.3 Compromise assessments

**REQ-02-083**
Compromise state MUST be modeled as incident-scoped assessment history attached to a host or identity rather than as a static property on the entity.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-084**
`assessment_state` MUST use this closed vocabulary:

- `unknown`: no meaningful assessment yet; evidence is insufficient or not yet reviewed.
- `suspected`: there is credible evidence of compromise, but not enough to confirm.
- `confirmed`: compromise is supported by sufficient evidence.
- `disproven`: the subject was investigated and compromise is not supported for the scoped incident context.
- `cleared`: the subject was previously suspected or confirmed compromised, and is now assessed as no longer actively compromised as of this assessment.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-085**
`assessment_state` MUST represent evidentiary judgment about compromise for the subject at the time of assessment.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-086**
Operational-response terms such as `contained`, `isolated`, `disabled`, `reset`, or `monitored` MUST NOT be used as assessment states. Response posture belongs in linked timeline events, evidence, tasks, decisions, or a future separate response-state model.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-087**
`disproven` and `cleared` MUST remain distinct because the former records a negative investigation result and the latter records post-remediation or post-validation state after actual or plausible compromise.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-088**
Each assessment MUST carry:

- `assessment_state`,
- `assessed_at`,
- assessor,
- nullable `confidence_score` in the range `0..100`,
- rationale,
- optional supporting record links.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-089**
`confidence_score=NULL` means unspecified. `confidence_score` MUST express confidence in the correctness of the assessment entry itself. It MUST NOT be reused as severity, impact, priority, or urgency.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-090**
Assessment history MUST be append-only. A new judgment MUST append a new assessment record for the incident-scoped subject. It MUST NOT overwrite or silently mutate prior assessments.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-091**
The system MUST expose a deterministic derived `confidence_band` with the following mapping:

- `unset` when `confidence_score IS NULL`,
- `low` when `confidence_score BETWEEN 0 AND 39`,
- `medium` when `confidence_score BETWEEN 40 AND 69`,
- `high` when `confidence_score BETWEEN 70 AND 100`.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-092**
The system MUST expose a stable compromise-assessment system-view and API contract over assessment history with fields equivalent to:

- subject record and subject type,
- `assessment_state`,
- `confidence_score`,
- `confidence_band`,
- rationale,
- assessor,
- `assessed_at`,
- optional supporting link counts.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-02-093**
This contract MUST remain stable across filtering, reporting, and future storage evolution.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

### 10.4 Analyst-work tracking

Tags and notes remain sufficient only for unassigned, non-transactional working material such as rough observations, ad hoc reminders, local analytic prose, and other low-stakes annotations that do not need accountable ownership, explicit lifecycle state, handoff durability, queue-oriented views, or later reconstruction as coordination artifacts.

**REQ-02-094**
Analyst work MUST be promoted out of tags or notes and into a structured analyst-work object when any of the following becomes true:

- it needs one accountable owner,
- it needs explicit lifecycle state,
- it needs due-date or blocker tracking,
- it must survive handoff or status review,
- it must be queryable as a queue, board, or owner workload view,
- it must explain or justify an operational action later,
- it must link multiple records as one durable unit of work.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

#### 10.4.1 `task_request` record type

**REQ-02-095**
The base profile MUST model owned work and requests as first-class `task_request` records.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-096**
A `task_request` record MUST carry, at minimum:

- stable `record_id`,
- owning `incident_id`,
- `task_kind` from the exact closed vocabulary defined in §18,
- `title`,
- `status` from the exact closed vocabulary defined in §18,
- `owner_user_id`,
- `priority` from the exact closed vocabulary defined in §18,
- `created_at`,
- `updated_at`.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-097**
A `task_request` record MUST also be able to persist, at minimum, the following optional fields when present:

- `workstream`,
- `due_at`,
- `requester_party_text`,
- `blocked_reason`,
- `completed_at`,
- `external_ticket_ref`,
- `linked_record_ids[]`,
- `decision_record_id`,
- `closure_summary`.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-098**
If `status='blocked'`, `blocked_reason` MUST be non-empty.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-099**
If `status='done'`, `completed_at` MUST be present and MUST NOT be earlier than `created_at`.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-100**
An active `task_request` with `status` not in `done` or `canceled` MUST NOT be ownerless.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-101**
If `linked_record_ids[]` or `decision_record_id` are persisted as denormalized convenience fields, authoritative cross-record association MUST still be representable through generic `record_links`.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

##### 10.4.1.1 Task-request lifecycle machine

The current profile defines `task_request.status` as a normative lifecycle machine rather than only a closed vocabulary.

The authoritative machine condition is determined by persisted `status` together with `blocked_reason`, `completed_at`, `owner_user_id`, and `created_at`.

The machine states mean:

- `open`: the task exists and is not currently being worked,
- `in_progress`: the task is actively being worked,
- `blocked`: the task cannot proceed until a prerequisite or dependency is resolved,
- `done`: the task is complete,
- `canceled`: the task is no longer being pursued.

Allowed lifecycle triggers are record creation that sets an initial `status` and later record patches that change `status` or one of the guard fields above.

Initial creation MAY set `status` to any value in the closed vocabulary. Interactive blank-row creation SHOULD default to `open` unless the operator explicitly chooses another allowed initial state.

The legal persisted status transitions are:

- `open` -> `in_progress`, `blocked`, `done`, or `canceled`,
- `in_progress` -> `open`, `blocked`, `done`, or `canceled`,
- `blocked` -> `open`, `in_progress`, `done`, or `canceled`,
- `done` -> `open`, `in_progress`, or `blocked`,
- `canceled` -> `open`, `in_progress`, or `blocked`.

**REQ-02-102**
A write that leaves `status` unchanged MUST be treated as a legal idempotent in-state write if the resulting record still satisfies all guards in this subsection.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-103**
The resulting authoritative row MUST satisfy all of the following guards:

- if `status='blocked'`, `blocked_reason` MUST be non-empty,
- if `status!='blocked'`, `blocked_reason` MUST be absent or empty,
- if `status='done'`, `completed_at` MUST be present and MUST NOT be earlier than `created_at`,
- if `status!='done'`, `completed_at` MUST be absent,
- if `status` is `open`, `in_progress`, or `blocked`, `owner_user_id` MUST be present.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-104**
When a successful write changes `status` away from `blocked`, the implementation MUST clear persisted `blocked_reason` before commit.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-105**
When a successful write changes `status` away from `done`, the implementation MUST clear persisted `completed_at` before commit.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-106**
When a successful write sets `status='done'` and the client did not supply `completed_at`, the implementation MUST set `completed_at` to the commit timestamp before commit.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-107**
Any other attempted write set that would leave the resulting record outside these invariants MUST be rejected as an illegal transition.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-108**
A requested transition from `done` to `canceled` or from `canceled` to `done` MUST be rejected as an illegal transition.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

**REQ-02-109**
A successful lifecycle transition MUST be committed through the ordinary record-mutation path, increment `row_version`, append a `change_set` and mutation entries, update derived projections, and emit the ordinary `record_changed` signal for that record. A rejected lifecycle transition MUST return the stable lifecycle-validation error defined by Core 01 §3.3.6 and MUST leave authoritative persisted state unchanged.
Profiles: base
Verified by: AC-085, AC-137, AC-138, AC-139, AC-140, AC-145, AC-231

#### 10.4.2 `decision` record type

**REQ-02-110**
The base profile MUST model rationale-bearing incident coordination choices as first-class `decision` records.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-02-111**
A `decision` record MUST carry, at minimum:

- stable `record_id`,
- owning `incident_id`,
- `decision_type` from the exact closed vocabulary defined in §18,
- `summary`,
- `status` from the exact closed vocabulary defined in §18,
- `owner_user_id`,
- `decided_at`,
- `rationale`.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-02-112**
A `decision` record MUST also be able to persist, at minimum, the following optional fields when present:

- `support_refs[]`,
- `affected_record_ids[]`,
- `supersedes_record_id`,
- `review_class`.
Profiles: base, snapshot_reporting
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231, AC-233

**REQ-02-113**
If `affected_record_ids[]` or `supersedes_record_id` are persisted as denormalized convenience fields, authoritative cross-record association MUST still be representable through generic `record_links`.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

##### 10.4.2.1 Decision lifecycle machine

**REQ-02-114**
The current profile defines `decision.status` as a normative lifecycle machine rather than only a closed vocabulary. The `approved` state is an incident-coordination state. It MUST NOT be interpreted as a generalized reviewer or administrator approval gate for ordinary row edits.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

The authoritative machine condition is determined by persisted `status` together with `supersedes_record_id` or an equivalent authoritative decision-to-decision link, `owner_user_id`, and `decided_at`.

The machine states mean:

- `proposed`: candidate decision under consideration,
- `approved`: accepted as the current intended course of action but not necessarily yet carried out,
- `rejected`: considered and declined,
- `executed`: carried out,
- `superseded`: overtaken before execution by a later accepted decision.

Allowed lifecycle triggers are record creation that sets an initial `status`, later record patches that change `status`, and the explicit supersession flow that links one decision to another.

**REQ-02-115**
Initial creation MAY set `status` to `proposed`, `approved`, `rejected`, or `executed`. Initial creation MUST NOT set `status='superseded'`.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

The legal direct status transitions are:

- `proposed` -> `approved`, `rejected`, or `executed`,
- `approved` -> `executed`.

**REQ-02-116**
A direct write that leaves `status` unchanged MUST be treated as a legal idempotent in-state write if the resulting record still satisfies this subsection, except that a direct write whose requested or resulting `status` is `superseded` MUST still be rejected because `superseded` is reachable only through the explicit supersession flow.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

`rejected`, `superseded`, and `executed` are terminal persisted statuses for direct status writes.

**REQ-02-117**
A direct write that would change `status` from `approved` to `rejected`, from `rejected` to `proposed`, from `executed` to any other value, or to `superseded` MUST be rejected as an illegal transition. A later contrary choice MUST be modeled as a new decision rather than as an in-place reversal of a `rejected`, `superseded`, or `executed` record.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-02-118**
The explicit supersession flow MUST satisfy all of the following:

- the superseding decision and the target decision MUST be different records in the same incident,
- the superseding decision MUST already have `status` `approved` or `executed`,
- the target decision MAY have `status` `proposed`, `approved`, or `executed`,
- the committed action MUST persist the supersession relation on the superseding decision through `supersedes_record_id` or an equivalent authoritative decision link,
- if the target decision has `status` `proposed` or `approved`, the same committed action MUST set the target decision's persisted `status` to `superseded`,
- if the target decision has `status` `executed`, the target decision MUST remain `executed`; a projection MAY still surface `decision.is_superseded=true` for that record, but the persisted `status` MUST NOT change away from `executed`.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-02-119**
A successful lifecycle transition or explicit supersession action MUST be committed through the ordinary record-mutation path, increment `row_version` on every changed record, append a `change_set` and mutation entries, update derived projections, and emit the ordinary `record_changed` signal for each changed record. A rejected lifecycle transition or supersession action MUST return the stable lifecycle-validation error defined by Core 01 §3.3.6 and MUST leave authoritative persisted state unchanged.
Profiles: base
Verified by: AC-086, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

#### 10.4.3 Ownership and hot-path boundary

`ownership` is not a standalone current-profile record type.

**REQ-02-120**
Ownership MUST instead be modeled as a required `owner_user_id` or equivalent stable assignee field on coordination objects such as `task_request`, `decision`, `status_review`, `lesson`, handoff artifacts that require accountable follow-through, and evidence-request records when the evidence model is used in request mode.
Profiles: base
Verified by: AC-085, AC-086, AC-090, AC-231

**REQ-02-121**
Routine timeline, host, identity, note, and evidence editing MUST NOT require owner, approver, challenge, or checklist fields merely to preserve the primary capture path.
Profiles: base
Verified by: AC-085, AC-086, AC-090, AC-231

**REQ-02-122**
The base profile MUST NOT introduce a generalized approval workflow for ordinary row edits in order to support analyst-work coordination objects.
Profiles: base
Verified by: AC-085, AC-086, AC-090, AC-231

#### 10.4.4 Structured coordination artifact types

**REQ-02-123**
The current profile MUST keep the following coordination surfaces artifact-backed rather than promoting them to first-class `record_type`s:

- communications log artifacts with `artifact_type='comm_log'`,
- handoff artifacts with `artifact_type='handoff'`,
- status-review artifacts with `artifact_type='status_review'`,
- lesson or follow-up artifacts with `artifact_type='lesson'`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-124**
A `comm_log` artifact MUST carry, at minimum:

- `comm_id`,
- `comm_type` from the exact closed vocabulary defined in §18,
- `timestamp_utc`,
- `audience`,
- `channel_or_meeting`,
- `summary`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-125**
A `comm_log` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `decision_ids[]`,
- `action_task_ids[]`,
- `next_report_at`,
- `privilege_tag`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-126**
A `handoff` artifact MUST carry, at minimum:

- `handoff_id`,
- `timestamp_utc`,
- `outgoing_owner_user_id`,
- `incoming_owner_user_id`,
- `current_state_summary`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-127**
A `handoff` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `open_task_ids[]`,
- `open_decision_ids[]`,
- `open_risk_refs[]`,
- `next_checks`,
- `acknowledged_at`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-128**
A `status_review` artifact MUST carry, at minimum:

- `status_review_id`,
- `timestamp_utc`,
- `review_owner_user_id`,
- `current_state_summary`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-129**
A `status_review` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `blocked_task_ids[]`,
- `pending_evidence_ids[]`,
- `open_decision_ids[]`,
- `active_risks_summary`,
- `next_report_at`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-130**
A `lesson` artifact MUST carry, at minimum:

- `lesson_id`,
- `timestamp_utc`,
- `summary`,
- `owner_user_id`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-131**
A `lesson` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `follow_up_task_ids[]`,
- `closure_state`,
- `evidence_refs[]`.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-132**
These coordination artifact types MUST reuse the shared artifact model, revision history, tags, `record_links`, projections, and `view_schema_id` contracts. They MUST NOT introduce Notes-specific or artifact-type-specific standalone storage silos.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

**REQ-02-133**
Any `*_ids[]` fields on coordination artifacts are denormalized convenience fields only. Authoritative many-to-many association MUST remain representable through generic `record_links` or dedicated child tables when per-link metadata is required.
Profiles: base
Verified by: AC-087, AC-088, AC-089, AC-231

#### 10.4.5 Hypothesis boundary

**REQ-02-134**
Current-profile hypotheses MUST remain artifact-backed, using either `artifact_type='hypothesis'` or a structured findings subtype. Hypotheses MUST NOT be promoted to a first-class `record_type` until repeated usage demonstrates a need for explicit competing-hypothesis tracking, support and contradiction sets, state transitions, or reviewer-visible hypothesis history that artifact-backed tracking cannot satisfy.
Profiles: base
Verified by: AC-089, AC-231

#### 10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces

The base profile does not require dedicated built-in sheets for findings, investigative queries, or forensic keywords.

**REQ-02-135**
If the implementation exposes any of these surfaces as typed artifact subtypes, contract-backed system views, or future first-class record types, their defining fields MUST be persisted as structured state rather than only in `custom_attrs`.
Profiles: base
Verified by: AC-101, AC-231

**REQ-02-136**
A structured finding or hypothesis subtype MUST be able to persist, at minimum:

- `statement`,
- `state`,
- `confidence_score`,
- `owner_user_id`,
- optional `closed_at`,
- optional supporting or contradictory record references.
Profiles: base
Verified by: AC-101, AC-231

**REQ-02-137**
A structured investigative-query subtype MUST be able to persist, at minimum:

- `query_id`,
- `platform`,
- `purpose`,
- `query_text`,
- `created_by_user_id`.
Profiles: base
Verified by: AC-101, AC-231

**REQ-02-138**
A structured forensic-keyword subtype MUST be able to persist, at minimum:

- `keyword_id`,
- `pattern`,
- `reason`,
- optional regex or literal flags,
- optional case-sensitivity flags.
Profiles: base
Verified by: AC-101, AC-231

### 10.5 Snapshot and reporting extension objects

**REQ-02-139**
If the Snapshot and Reporting Extension Profile is implemented, the domain model MUST also support structured metadata for:

- immutable snapshot descriptors,
- canonical export-model fields and blocks with stable export-model paths,
- versioned redaction profiles,
- versioned template contracts,
- redaction manifests,
- artifact release records.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

**REQ-02-140**
At minimum, an immutable snapshot descriptor MUST persist the release tuple defined by Core 01 §10.2.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

**REQ-02-141**
For rendered-output lifecycle, the authoritative persisted state MUST be artifact-scoped release records plus any bound approval records.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

For this lifecycle, the logical output slot is the bound release tuple excluding `output_sha256`.

**REQ-02-142**
Each release record MUST also persist, at minimum:

- `release_state` from the exact closed vocabulary defined in §18,
- `approved_at`,
- `invalidated_at`,
- `published_at`,
- optional `invalidation_reason`.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

A rendered output enters `pending_approval` when bytes and `output_sha256` exist for one immutable release tuple but the required approvals are not yet satisfied. It enters `approved` only when the required approvals are satisfied for that exact release record. It enters `published` only through an explicit publish action after approval. It enters `invalidated` when a different artifact for the same logical output slot supersedes it, when `output_sha256` changes for that slot, or when the implementation can no longer attest that the required approval set applies to that exact artifact.

**REQ-02-143**
Each exportable field or block in the canonical export model MUST persist:

- a stable export-model path,
- exactly one `content_class`,
- deterministic order metadata sufficient to reproduce export ordering,
- when `content_class='curated_narrative'`, zero or more `support_refs[]`.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

**REQ-02-144**
When recipient-specific reporting is implemented, each exportable field or block MUST also be able to carry zero or more stable incident-local `disclosure_partition_refs[]`. Each `disclosure_partition_ref` names an export-only withholding boundary such as an affected party or other approved recipient partition. `disclosure_partition_refs[]` MUST be stored as snapshot or export metadata only. They MUST NOT change live workspace authorization.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

**REQ-02-145**
A release record MUST bind, at minimum:

- `snapshot_id`,
- `template_id`,
- `template_version`,
- `redaction_profile_id`,
- `redaction_profile_version`,
- `output_kind`,
- `release_scope`,
- `output_sha256`,
- `release_state`,
- `approved_at`,
- `invalidated_at`,
- `published_at`.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

**REQ-02-146**
If approval state is stored, it MUST bind to the release record rather than to mutable incident rows.
Profiles: snapshot_reporting
Verified by: AC-030, AC-031, AC-032, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-071, AC-091, AC-104, AC-105, AC-106, AC-113, AC-114, AC-115, AC-233

## 11. Type registries and view contracts

**REQ-02-147**
Type and icon choices MUST be driven by stable registry keys rather than hard-coded display text.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-02-148**
Host types, evidence types, indicator types, and similar display taxonomies MUST resolve through versioned registries keyed by stable IDs, with display labels, categories, icon keys, optional local overrides, and indicator-type-specific normalization or validation behavior where applicable.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-02-149**
Indicator-type registries MUST define normalization, validation, defanging, and dedupe behavior for each indicator type. Optional STIX export or mapping data MAY be present in the registry, but it MUST NOT control canonical indicator identity.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-02-150**
Entity-bearing columns and import mappings MUST carry stable `field_key` values and `entity_binding_mode` so that mention-versus-entity behavior never depends on a visible header such as `Host`, `User`, or `Account`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

**REQ-02-151**
View contracts and import mappings that support structured indicator capture from source text MUST declare the eligible `field_key` values or equivalent stable field identifiers. Visible labels MUST NOT determine indicator-capture behavior.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-150, AC-151, AC-152, AC-153, AC-231

### 11.1 Saved-view contract

**REQ-02-152**
A saved view MUST be an incident-scoped configuration object over exactly one `view_schema`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-02-153**
A saved view MUST persist, at minimum:

- stable `saved_view_id`,
- owning `incident_id`,
- immutable `view_schema_id`,
- `scope` with one of `private`, `shared`, or `system`,
- `display_name`,
- normalized `query_json`,
- `layout_json`,
- `owner_user_id`,
- `created_at`,
- `updated_at`,
- monotonically increasing `saved_view_version`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-02-154**
`owner_user_id` MUST be present for `private` and `shared` saved views. It MAY be null only for `system` saved views.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-02-155**
`query_json` MUST preserve saved-view sort, filter, and grouping state using stable `field_key` values, stable grouping identifiers, and normalized scalar values. It MUST NOT store visible labels as the authoritative identifier for any saved sort, filter, or grouping element. `query_json.filters[]` MUST use the exact filter predicate wire shape defined by Core 01 §3.3.4.1, preserve the operator-specific `arg` object, and persist canonical `filters[]` ordering by `field_key asc`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-02-156**
A saved view created by duplicating another saved view MUST persist a full normalized copy of the source `view_schema_id`, `query_json`, and `layout_json`. The new saved view MUST NOT carry a runtime dependency on the source `saved_view_id`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-02-157**
The schema MUST enforce immutability of `saved_view_id`, `incident_id`, and `view_schema_id` after creation.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

### 11.2 Workbook startup preference objects

**REQ-02-158**
The schema MUST support two distinct workbook-startup preference objects:

- `user_workbook_preferences`, keyed by `(incident_id, user_id)`,
- `incident_workbook_preferences`, keyed by `incident_id`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-02-159**
`user_workbook_preferences` MUST persist, at minimum:

- `incident_id`,
- `user_id`,
- nullable `home_sheet_ref`,
- `created_at`,
- `updated_at`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-02-160**
`incident_workbook_preferences` MUST persist, at minimum:

- `incident_id`,
- nullable `default_sheet_ref`,
- `created_at`,
- `updated_at`,
- `updated_by_user_id`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-02-161**
Both `home_sheet_ref` and `default_sheet_ref` MUST use the `sheet_ref` union defined by Core 01 §3.3.10.1 and MUST be nullable so the implementation can clear an invalid or unwanted pointer without deleting the preference object.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-02-162**
The schema MUST NOT overload saved-view rows with a generic `is_default` flag or any equivalent bit that conflates per-user startup selection, incident-wide fallback selection, or system-seeded queue surfaces.
Profiles: base
Verified by: AC-150, AC-153, AC-231

## 12. Typed relationships

**REQ-02-163**
The implementation MUST provide a generic typed relationship store equivalent to `record_links`.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

### 12.1 `record_link` object contract

A `record_link` is a stable, incident-scoped, directed, soft-deletable non-record mutation target between two first-class record-envelope rows in the same incident.

**REQ-02-164**
A `record_link` MUST include, at minimum:

- stable `record_link_id`,
- owning `incident_id`,
- `src_record_id`,
- `dst_record_id`,
- `link_type` from the exact closed vocabulary defined in §18,
- `provenance` from the exact closed vocabulary defined in §18,
- nullable `confidence` in the range `0..100`,
- optional `note`,
- `created_by_user_id`,
- `created_at`,
- nullable `deleted_at`,
- nullable `deleted_by_user_id`.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

Stable `record_link_id` values are the mutation-target identity for history, rollback, and deterministic link-level reconstruction.

**REQ-02-165**
Both endpoints MUST reference first-class record-envelope rows. A `record_link` MUST NOT target `entity_mentions`, `indicator_observations`, or any other non-record mutation target.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-166**
`src_record_id` and `dst_record_id` MUST belong to the same `incident_id` as the `record_link` row. A `record_link` MUST NOT connect a record to itself.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-167**
A `record_link` is not a record-envelope row. It MUST NOT participate in the generic record-scoped delete or restore routes defined for first-class records, and it MUST NOT introduce a standalone `row_version` optimistic-concurrency contract in the current profile.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-168**
There MUST be at most one non-deleted `record_link` for the same `(incident_id, src_record_id, dst_record_id, link_type)` tuple.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-169**
For projection, count, and ordinary export purposes, a link is active only when the `record_link` row is not soft-deleted and both endpoint records are not soft-deleted. Soft-deleted links and links whose endpoint records are soft-deleted MUST remain reconstructible through history and rollback.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-170**
Merge-time repoint semantics from §9 apply to `record_links`. Incoming and outgoing active links to a losing entity MUST be repointed or deterministically recreated on the survivor in the same `change_set`, and duplicate active tuples created by repointing MUST be deduplicated without losing history.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-171**
All persisted `record_links` MUST be directed. The base profile MUST NOT require mirrored reverse rows, per-row symmetry flags, or inverse-type columns. Reverse traversal, reverse counts, and equivalent "linked from" UI affordances MUST be derived from the current directed links.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-172**
`record_links` MUST remain distinct from `entity_mentions` and `indicator_observations`. Mention-origin references MUST use `entity_mentions`. Source-bound indicator occurrences MUST use `indicator_observation` rows. A canonical indicator association MAY also be represented by a `record_link` when the applicable `link_type` is `references_indicator`.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

### 12.2 Relationship vocabulary and canonical direction

**REQ-02-173**
The relationship vocabulary MUST be controlled by application code or equivalent authoritative configuration. It MUST NOT be free-text user input in the UI.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-174**
The relationship vocabulary MUST support, at minimum, the exact base-profile `link_type` tokens defined in §18:

- `observed_on_host`,
- `observed_as_identity`,
- `references_indicator`,
- `attached_evidence`,
- `references_artifact`,
- `derived_from`,
- `merged_into`,
- `supported_by`,
- `references_record`,
- `supersedes`.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-175**
For persisted `record_links`, `src_record_id` and `dst_record_id` MUST use these canonical directions:

- `observed_on_host`: source record -> host,
- `observed_as_identity`: source record -> identity,
- `references_indicator`: source record -> canonical indicator,
- `attached_evidence`: anchor record -> evidence record,
- `references_artifact`: anchor record -> artifact record,
- `derived_from`: derived record -> source record,
- `merged_into`: losing record -> survivor,
- `supported_by`: assessment or decision -> supporting record,
- `references_record`: task request or other coordination record -> referenced record,
- `supersedes`: superseding decision -> superseded decision.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-176**
Client-visible reverse traversal or reverse counts MUST be derived from these canonical directions. They MUST NOT imply a second persisted link.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

### 12.3 Link metadata semantics

The exact `provenance` tokens are defined in §18. The following semantics apply:

- `manual`: explicit analyst action created the link,
- `auto_match`: the link was created by the interactive host or identity auto-resolution flow defined by Core 03 §12.3,
- `import`: the link was created by a file or API import flow,
- `rollback`: the link was recreated or restored by rollback,
- `system`: the link was created by another system-managed action.

`provenance` is required.

**REQ-02-177**
`confidence` is a nullable `0..100` score about the correctness of a machine-produced link assertion. It MUST NOT be reused as evidence strength, assessment confidence, severity, impact, or priority.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-178**
A link created by explicit analyst action in the base profile MUST use `provenance='manual'` unless the applicable write path is the Core 03 §12.3 auto-resolution flow.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-179**
`provenance='auto_match'` is valid only for interactive host and identity auto-resolution. Such links MUST set `confidence=100`.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

Explicit analyst-created links SHOULD leave `confidence` null unless a later profile defines a stable machine-produced scoring contract for that write path.

**REQ-02-180**
`note` MAY hold free text and operator context, but it MUST NOT carry conformance-critical semantics.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-181**
A timeline event MUST be able to relate to:

- hosts and identities via typed record links,
- canonical indicators via typed record links,
- unresolved host and account strings via entity mentions,
- source-bound indicator occurrences via `indicator_observation` rows,
- notes and other artifacts via typed record links,
- evidence via typed record links to evidence records.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-182**
Hosts, identities, and evidence records MUST also be able to relate to note artifacts and canonical indicators through the same generic typed relationship store.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-183**
When a note is created from timeline, host, identity, or evidence context, the association MUST be persisted through generic `record_links` rather than a Notes-specific foreign key on the source record.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

A single note artifact MAY relate to zero, one, or many source records through `record_links`.

**REQ-02-184**
Task requests, decisions, and coordination artifacts MUST also be able to relate to timeline events, hosts, identities, canonical indicators, evidence records, notes, and each other through the same generic typed relationship store.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

**REQ-02-185**
The implementation MUST NOT rely on task-specific, decision-specific, or handoff-specific foreign-key columns on timeline, host, identity, or evidence rows as the only supported linkage mechanism.
Profiles: base
Verified by: AC-205, AC-206, AC-207, AC-208, AC-209, AC-210, AC-231

## 13. Evidence and object metadata

**REQ-02-186**
The implementation MUST store authoritative blob-slot create-contract fields and observed object metadata in structured storage, including:

- blob identifier,
- owning `incident_id`,
- route-scoped create idempotency key,
- storage backend,
- bucket and key,
- accepted `filename_hint`,
- accepted `content_type_hint`,
- declared `byte_size`,
- optional expected SHA-256,
- observed content type,
- observed size,
- observed SHA-256,
- upload state.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-187**
The binary content MUST live in object storage. Structured storage MUST remain the authoritative pointer and metadata source. Caller-supplied `filename_hint` and `content_type_hint` are advisory accepted-contract metadata only. They MUST NOT determine storage keys, authorization, portability layout, preview allowlisting, or release posture. After upload, observed size and verified hash are authoritative object metadata, and observed media metadata drives preview policy.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-188**
The evidence model MUST support both:

- requested or pending evidence with no blob yet attached,
- received or available evidence with custody events and optional blob linkage.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

Cartulary defines two linked but separate evidence-related lifecycle machines:

- a blob-upload machine authoritative on `object_blobs.upload_state`,
- an evidence-custody and availability machine authoritative on `evidence_records.lifecycle_state` plus append-only custody events.

The blob-upload machine uses the exact closed vocabulary defined in §18.

The evidence-custody and availability machine uses the exact closed vocabulary defined in §18.

**REQ-02-189**
These machines MUST remain separate. `object_blobs.upload_state` MUST NOT be treated as the user-facing evidence lifecycle, and an evidence record MUST be able to exist with no blob at all.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-190**
The bridge rules are:

- an evidence record in `requested` or `pending_receipt` MAY have no `object_blob_id`,
- if an evidence record has a linked `object_blob_id`, it MUST NOT enter `available` unless that blob is in `upload_state='available'`,
- if a linked blob is `quarantined`, the evidence record MUST be `quarantined` or remain non-available,
- a blob in `pending` or `failed` MUST NOT make an evidence record appear attached, available, previewable, or released,
- `released` evidence MUST reference a blob in `upload_state='available'` when a blob is present.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-191**
Abandoned pending uploads MUST fail closed. A blob slot left in `pending` without successful finalization MUST NOT create or imply an attached evidence record, MUST NOT increment visible evidence counts, and MUST remain eligible only for retry, timeout handling, or administrative cleanup.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-192**
Each pending blob slot MUST persist `target_expires_at` and `pending_expires_at`. In the base profile, `target_expires_at` MUST be 60 minutes after upload-target issuance and `pending_expires_at` MUST be 24 hours after blob-slot creation. These timers MUST remain separate, and timeout handling MUST reuse `upload_state='failed'` rather than introduce a distinct expired lifecycle state.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-193**
The base profile MUST treat each pending blob slot as a single-upload lease bound to one accepted create contract. If the upload target expires before successful upload, retry requires a fresh blob-slot creation call rather than same-slot refresh or lease renewal. The base profile MUST NOT require resumable upload semantics.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-194**
Explicit finalization MUST compare observed bytes against the accepted create contract. If observed size differs from declared `byte_size`, finalization MUST fail immediately, the slot MUST transition to `upload_state='failed'`, `terminal_reason` MUST be `declared_size_mismatch`, and `failed_at` MUST be recorded. If expected SHA-256 is present and differs from observed SHA-256, finalization MUST fail immediately, the slot MUST transition to `upload_state='failed'`, `terminal_reason` MUST be `expected_sha256_mismatch`, and `failed_at` MUST be recorded. These mismatch failures are terminal on first detection and MUST NOT consume or extend the ordinary finalization retry budget. A mismatch between observed content type and advisory `content_type_hint` alone MUST update observed metadata and preview policy and MUST NOT by itself fail the slot. Filename mismatch alone MUST NOT be a finalization failure in the base profile. For other explicit failed finalization attempts on a pending blob slot, the implementation MUST track `finalize_attempt_count`. Only unsuccessful explicit finalization attempts count toward this total; an idempotent replay after already-committed success MUST NOT consume retry budget. A pending blob slot MUST allow 3 non-terminal failed explicit finalization attempts. On the 4th such failed attempt, it MUST transition to `upload_state='failed'`, persist `terminal_reason='finalize_retry_exhausted'`, and record `failed_at`. Later retry then requires a fresh blob slot.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-195**
A pending blob slot that is not successfully finalized by `pending_expires_at` MUST transition from `pending` to `failed`, persist `terminal_reason='pending_timeout'`, and record `failed_at`. For a failed blob slot that remains unattached to evidence, `cleanup_due_at` MUST be no later than 1 hour after terminal failure, orphaned blob bytes MUST be deleted by that deadline, and failed unattached slot metadata MUST remain queryable for at least 7 days before automatic hard deletion is allowed. `cleaned_up_at` MUST record completion of object-byte cleanup when that cleanup occurs.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-196**
If structured state becomes inconsistent, such as an evidence record claiming `available` or `released` while the linked blob is `pending`, `failed`, missing, or otherwise unusable, the implementation MUST fail closed: preview and download MUST be blocked and the record MUST surface as inconsistent until repaired by an explicit corrective action or re-finalization path.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-197**
An evidence record or its deterministic joined projection from authoritative object metadata MUST be able to expose, at minimum:

- `requested_at`,
- `received_at`,
- `storage_ref`,
- `blob_hash`,
- `collector_party_text`,
- `source_party_text`.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-198**
`requested_at` and `received_at` MUST remain first-class structured state. `storage_ref` MAY be a human-usable repository locator derived from the authoritative object-blob pointer, but it MUST be filterable and exportable without parsing `custom_attrs`.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-199**
When an evidence record represents requested or pending evidence rather than already received material, the model MUST be able to store `owner_user_id` or an equivalent stable assignee field so accountable follow-up does not require a separate timeline field. A linked `task_request` MAY also exist.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-02-200**
An implementation MAY persist a non-authoritative `sensitivity_class` on evidence records and MAY project that label onto exportable evidence-derived blocks. In the base profile, `sensitivity_class` is metadata for preview labeling, default export-redaction behavior, and audit or reporting only. It MUST NOT change incident authorization or hide live workspace content.
Profiles: base, snapshot_reporting
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231, AC-233

**REQ-02-201**
When custody narrative or handoff commentary is preserved as authoritative workflow state, it MUST be modeled as append-only custody events or equivalent child rows. A convenience `custody_note` or similar summary field MAY exist, but it MUST NOT be the sole authoritative custody history.
Profiles: base
Verified by: AC-015, AC-016, AC-053, AC-100, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

## 14. Schema invariants

### 14.1 Mention, indicator, provenance, and coordination fields

**REQ-02-202**
The schema MUST support:

- `source_field_key`,
- `origin_kind`,
- `origin_locator`,
- file-based import provenance fields for `import_session_id`, `import_unit_id`, `mapping_fingerprint`, `parser_profile_id`, source file kind, source content hash, parser version, and selected sheet or region locator,
- unknown-column preservation fields sufficient to persist `import_unit_id`, source column ordinal, source header text, and raw value without conflating duplicate visible headers,
- resolution metadata on mentions,
- source-bound indicator-observation fields sufficient to persist observed text, optional parsed indicator type, optional normalized candidate, deterministic source locator or span, and resolution metadata,
- `record_link` fields sufficient to persist stable `record_link_id`, directed `src_record_id` and `dst_record_id`, closed-vocabulary `link_type`, closed-vocabulary `provenance`, nullable `confidence`, optional `note`, creation attribution, creation timestamp, and soft-delete attribution,
- append-only indicator lifecycle interval fields separate from observation timestamps,
- append-only compromise-assessment fields sufficient to persist closed-vocabulary `assessment_state`, `assessed_at`, assessor attribution, nullable `confidence_score`, rationale, optional supporting record references, and deterministic derivation of `confidence_band`,
- reference-pack manifest fields sufficient to persist `pack_key`, `pack_kind`, `pack_version`, `source_identifier`, `manifest_sha256`, one or more payload SHA-256 digests in deterministic member order or an equivalent canonical aggregate digest, declared `pack_contract_version`, signature or trusted-source metadata, `verification_method`, and non-active availability state,
- reference-pack activation and attestation fields sufficient to persist one active-version pointer per `pack_key`, imported and activated actor attribution with timestamps, `previous_active_version`, `verification_result`, and optional operator note or change ticket,
- incident fields sufficient to persist `incident_key`, `title`, optional `description`, `status`, optional `severity`, optional `tlp`, optional `current_phase`, optional `primary_external_case_ref`, `created_at`, `updated_at`, nullable `closed_at`, creation attribution, update attribution, and monotonically increasing `incident_version`, plus a canonical `incident_key` uniqueness form or equivalent deterministic uniqueness substrate derived by trimming leading and trailing Unicode whitespace and applying Unicode NFC normalization,
- internal-user fields sufficient to persist stable `user_id`, a canonical case-insensitive email uniqueness form or equivalent deterministic lookup substrate, `display_name`, `password_hash`, `mfa_required`, `is_active`, `is_deployment_admin`, `created_at`, `updated_at`, nullable `updated_by_user_id`, nullable `last_login_at`, and monotonically increasing `user_version`,
- auth-binding fields sufficient to persist provider-backed or local binding data keyed to stable internal `user_id`, including `provider_key`, `provider_type`, provider subject or equivalent identity key, optional `username`, and creation and last-auth timestamps, without making incident memberships depend on provider subject,
- incident-membership fields sufficient to persist `(incident_id, user_id)`, `role`, `joined_at`, `added_by_user_id`, `updated_at`, `updated_by_user_id`, and monotonically increasing `membership_version`,
- saved-view fields sufficient to persist immutable `view_schema_id`, `scope`, `display_name`, normalized `query_json`, `layout_json`, nullable `owner_user_id`, and monotonically increasing `saved_view_version`,
- workbook-preference fields sufficient to persist per-user `home_sheet_ref` and incident-wide `default_sheet_ref` separately without overloading a saved-view row flag,
- `entity_origin` and seed provenance on host and identity records,
- host fields sufficient to persist `location`, `os_platform`, `business_owner`, `criticality`, and `containment_status`,
- identity fields sufficient to persist `privilege_level`, `mfa_state`, and `reset_status`,
- `entity_binding_mode` in view write-back contracts and import mappings,
- canonical indicator fields sufficient to persist `defanged_value`, optional `hash_algorithm`, optional `hash_value`, and deterministic storage or derivation of `first_observed_at` and `last_observed_at`,
- `task_request` fields sufficient to persist `task_kind`, `status`, `owner_user_id`, `priority`, optional `workstream`, `due_at`, `blocked_reason`, `requester_party_text`, `completed_at`, `external_ticket_ref`, and optional linked-record or linked-decision references,
- `decision` fields sufficient to persist `decision_type`, `status`, `owner_user_id`, `decided_at`, rationale, support references, affected-record references, and optional supersession linkage,
- evidence fields sufficient to persist separate blob-upload state, evidence lifecycle state, the bridge between `object_blob_id` and evidence availability, owning `incident_id`, blob-slot create idempotency key, declared upload-contract fields for `byte_size`, advisory `filename_hint`, advisory `content_type_hint`, and optional expected SHA-256, observed object metadata for content type, size, and verified SHA-256, `target_expires_at`, `pending_expires_at`, `finalize_attempt_count`, `terminal_reason`, `failed_at`, `cleanup_due_at`, `cleaned_up_at`, `requested_at`, `received_at`, `storage_ref`, `blob_hash`, `collector_party_text`, `source_party_text`, and append-only custody events,
- snapshot and release fields sufficient to persist artifact-scoped `release_state`, approval binding, `approved_at`, `invalidated_at`, `published_at`, optional `invalidation_reason`, and deterministic identification of the logical output slot,
- coordination-artifact fields sufficient to persist `comm_type` and other `artifact_type`-specific metadata for `comm_log`, `handoff`, `status_review`, `lesson`, and optional current-profile `hypothesis` tracking,
- optional structured artifact subtype fields sufficient to persist findings, investigative queries, and forensic keywords when those surfaces are implemented.
Profiles: base, import, snapshot_reporting, reference_pack
Verified by: AC-017, AC-018, AC-072, AC-073, AC-074, AC-075, AC-118, AC-128, AC-154, AC-155, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-231, AC-232, AC-233, AC-234

**REQ-02-203**
For incidents created through the public base-profile route, the schema MUST support `status='active'`, `incident_version=1`, `closed_at=NULL`, and one committed create timestamp written to both `created_at` and `updated_at`; that public create route MUST bind both creation and initial update attribution to the creating local `user_id`.
Profiles: base
Verified by: AC-017, AC-018, AC-072, AC-073, AC-074, AC-075, AC-118, AC-128, AC-154, AC-155, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-231

**REQ-02-204**
Internal-user and incident-membership state MUST remain deployment-local authorization state. Whole-incident portability import MAY map historical actor descriptors to existing local users, but import MUST NOT synthesize login-capable users, deployment-admin flags, or active memberships without explicit deployment-local administrative action.
Profiles: base, incident_portability
Verified by: AC-017, AC-018, AC-072, AC-073, AC-074, AC-075, AC-118, AC-128, AC-154, AC-155, AC-188, AC-189, AC-190, AC-200, AC-201, AC-202, AC-203, AC-204, AC-231, AC-236

### 14.2 Rollback granularity substrate

**REQ-02-205**
A conformant history schema MUST include both:

- immutable `change_set` attribution records, and
- mutation-entry records at target granularity.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-02-206**
`record_revisions` MAY retain row snapshots such as `before_json` and `after_json` for audit and whole-row restore. They MUST NOT be the sole rollback substrate.
Profiles: base, snapshot_reporting
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231, AC-233

**REQ-02-207**
Whole-row restore driven by a `record_revisions` snapshot MUST restore only authoritative row-backed fields for the selected `record_id` and revision number. It MUST NOT, by itself, recreate or delete non-row mutation targets such as `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, or evidence associations.
Profiles: base, snapshot_reporting
Verified by: AC-215, AC-216, AC-217, AC-218, AC-231, AC-233

### 14.3 Mutation targets

**REQ-02-208**
The mutation-entry model MUST support, at minimum:

- row-field edits,
- record links,
- record tags,
- entity mentions,
- indicator observations,
- indicator lifecycle intervals,
- compromise assessments,
- evidence associations,
- merge and repoint fan-out.
Profiles: base
Verified by: AC-125, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-231

**REQ-02-209**
Stable mutation targets MUST use deterministic serialization, for example `record_link:<record_link_id>`. Composite targets MUST serialize canonically, for example `record_tag:<record_id>:<tag_id>`.
Profiles: base
Verified by: AC-125, AC-200, AC-201, AC-202, AC-203, AC-204, AC-205, AC-206, AC-207, AC-208, AC-231

### 14.4 Soft delete

**REQ-02-210**
User-visible records MUST be soft-deletable. Links MUST be soft-deletable. Revisions MUST be append-only in normal operation.
Profiles: base
Verified by: AC-181, AC-182, AC-183, AC-231

Blobs MAY be hard-deleted only through an explicit administrative purge or retention workflow.

### 14.5 Snapshot and reporting extension fields

**REQ-02-211**
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
Profiles: snapshot_reporting
Verified by: AC-057, AC-060, AC-061, AC-062, AC-113, AC-114, AC-115, AC-233

## 15. History and rollback

### 15.1 Attribution unit

**REQ-02-212**
For each committed action, the implementation MUST create one immutable `change_set`.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-231

**REQ-02-213**
A `change_set` MUST record, at minimum:

- incident,
- actor,
- timestamp,
- source,
- optional reason,
- optional client transaction identifier.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-231

### 15.2 Mutation-entry unit

**REQ-02-214**
Each committed action MUST also create one or more reversible mutation entries.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-231

**REQ-02-215**
Each mutation entry MUST record:

- target kind,
- stable target identifier,
- operation kind,
- deterministic order within the parent `change_set`,
- pre-change and post-change version identifiers,
- reversible before/after values or an equivalent reversible patch.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-231

**REQ-02-216**
The history substrate MUST also support a stable opaque public `history_entry_ref` for any row-centric logical history item that maps to exactly one reversible mutation target. The public rollback interface MUST round-trip that reference without exposing storage-primary-key mutation-entry identifiers.
Profiles: base
Verified by: AC-215, AC-216, AC-217, AC-231

### 15.3 Reconstruction requirement

**REQ-02-217**
The history model MUST preserve enough information to reconstruct:

- the full row snapshot at any revision,
- the exact field, link, mention, tag, evidence, or merge delta introduced by a `change_set`.
Profiles: base, snapshot_reporting
Verified by: AC-215, AC-217, AC-231, AC-233

**REQ-02-218**
Projection tables MUST NOT be authoritative history.
Profiles: base
Verified by: AC-215, AC-217, AC-231

### 15.4 Entity-merge history

**REQ-02-219**
When entities are merged, the merge `change_set` MUST preserve enough detail to reconstruct both:

- the pre-merge graph,
- the post-merge graph.
Profiles: base
Verified by: AC-023, AC-186, AC-209, AC-217, AC-231

**REQ-02-220**
Live mention resolutions and live links MUST NOT continue to point at a merged-away record after the merge `change_set` commits.
Profiles: base
Verified by: AC-023, AC-186, AC-209, AC-217, AC-231

## 16. Search and indexing expectations

The source artifact included a concrete indexing strategy. The exact SQL realization MAY vary, but a conformant implementation SHOULD provide equivalent lookup behavior for:

- filtering and sorting by incident and key timeline dimensions,
- exact and normalized lookup over canonical indicator values and deterministic indicator dedupe keys,
- full-text search using a configuration appropriate for IR tokens rather than English stemming,
- link traversal by source and destination,
- traversal from canonical indicators to source-bound observations and linked records,
- filtering and sorting over `task_request` owner, status, priority, due, blocker, and last-updated state,
- filtering and sorting over `decision` owner, status, type, and `decided_at`,
- fuzzy matching over alias and unresolved mention text,
- array or equivalent containment lookups for denormalized tag and label sets.

## 17. Domain invariants

**REQ-02-221**
A conformant implementation MUST preserve all of the following invariants:

1. mentions and entities are different object types,
2. repeated mentions remain distinct observations,
3. exact-match entity reuse follows a deterministic precedence,
4. merges are explicit and auditable,
5. notes are artifacts,
6. notes and tags remain only for unassigned, non-lifecycle working material,
7. current-profile owned work uses structured `task_request`, `decision`, or coordination-artifact objects rather than overloading notes,
8. current-profile ownership is a field on coordination objects rather than a standalone object,
9. canonical indicators and source-bound indicator observations are different object types,
10. repeated indicator observations remain distinct observations,
11. indicator lifecycle intervals append rather than overwrite and remain distinct from observation times,
12. assessments append rather than overwrite,
13. relationship semantics are typed,
14. history is reversible at mutation-entry granularity,
15. projection state is derived rather than authoritative,
16. current-profile hypotheses remain artifact-backed until later promotion criteria are met,
17. reference-pack manifests, activation state, and attestation metadata remain incident-external structured state, and at most one version per `pack_key` is active at a time.
Profiles: base, reference_pack
Verified by: AC-231, AC-234

## 18. Canonical closed-vocabulary registry

**REQ-02-222**
This section owns the exact token sets for closed vocabularies that are persisted in structured state or surfaced through contract-backed views, portability bundles, or public API payloads. Earlier sections that described a vocabulary as "equivalent to" these values are resolved by this registry. A conformant implementation MUST persist and emit the exact tokens listed here.
Profiles: base, incident_portability
Verified by: AC-076, AC-077, AC-078, AC-079, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-122, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231, AC-236, AC-252, AC-253

Where an earlier section also defines lifecycle rules, semantic meanings, or guard behavior for a token family, that earlier section remains authoritative for those semantics. This registry owns the exact token spellings and membership of the token set.

| Structured field or token family | Exact tokens |
| --- | --- |
| `entity_mentions.resolution_status` | `unresolved`, `resolved`, `dismissed` |
| `entity_mentions.origin_kind` and `indicator_observation.origin_kind` | `manual_entry`, `clipboard_paste`, `csv_import`, `xlsx_import`, `api_import`, `extraction`, `system` |
| `host.entity_origin` and `identity.entity_origin` | `entity_sheet`, `entity_import`, `created_from_mention`, `system_upsert` |
| `indicator.value_kind` | `atomic`, `pattern`, `reference` |
| `assessment_state` | `unknown`, `suspected`, `confirmed`, `disproven`, `cleared` |
| `task_request.task_kind` | `question`, `request`, `collection`, `containment`, `follow_up` |
| `task_request.status` | `open`, `in_progress`, `blocked`, `done`, `canceled` |
| `task_request.priority` | `low`, `normal`, `high`, `urgent` |
| `decision.decision_type` | `scope`, `containment`, `communication`, `evidence`, `reporting` |
| `decision.status` | `proposed`, `approved`, `rejected`, `superseded`, `executed` |
| `record_links.link_type` | `observed_on_host`, `observed_as_identity`, `references_indicator`, `attached_evidence`, `references_artifact`, `derived_from`, `merged_into`, `supported_by`, `references_record`, `supersedes` |
| `record_links.provenance` | `manual`, `auto_match`, `import`, `rollback`, `system` |
| `artifact.comm_type` for `artifact_type='comm_log'` | `meeting`, `notification`, `approval`, `briefing`, `handoff` |
| `release_state` | `pending_approval`, `approved`, `invalidated`, `published` |
| `object_blobs.upload_state` | `pending`, `available`, `failed`, `quarantined` |
| `object_blobs.terminal_reason` | `pending_timeout`, `finalize_retry_exhausted`, `declared_size_mismatch`, `expected_sha256_mismatch` |
| `evidence_records.lifecycle_state` | `requested`, `pending_receipt`, `received`, `available`, `quarantined`, `released` |
| `evidence-access media_class` | `image`, `pdf`, `text`, `audio`, `video`, `archive`, `office_document`, `binary`, `active_content` |
| `evidence-access preview_kind` | `image_inline`, `pdf_inline`, `text_inline` |

A token family listed here MAY be emitted by deriving from authoritative metadata, server-side inspection, or lifecycle state rather than by persisting a same-named storage column. The presence of a token family in this registry MUST NOT be interpreted as a requirement to add a same-named column to the physical schema.

**REQ-02-223**
A current-profile implementation MUST NOT persist alternate spellings, display labels, or semantically equivalent synonyms for any token listed here in authoritative structured state. Client and export presentation layers MAY map these tokens to display labels, but the canonical token value MUST remain stable in stored state and machine-readable payloads.
Profiles: base
Verified by: AC-076, AC-077, AC-078, AC-079, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-122, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231, AC-252, AC-253
