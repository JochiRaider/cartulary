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
- `indicator`,
- `artifact`,
- `task_request`,
- `decision`,
- `evidence`,
- `assessment`.

Internal users, auth-provider identities, sessions, and incident memberships are authoritative normalized administrative state but are not user-visible first-class incident records. They MUST NOT consume `record_id` values, MUST NOT be modeled as generic record-envelope rows, and MUST NOT be the target of workbook row-mutation routes.

User-account and incident-membership mutations MUST remain auditable, but they MUST use a dedicated deployment-local administrative audit substrate rather than incident `change_set` or `record_revisions` rows.

## 4. Normalized versus flexible data

### 4.1 Normalized data

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

JSONB MUST NOT be the authoritative storage for:

- timestamps used for sorting,
- canonical names or primary identifiers,
- record relationships,
- evidence metadata required for retrieval,
- saved-view identity, scope, ownership, or startup/default surface selection,
- any field that requires reliable filtering or indexing at operational scale.

Saved-view `query_json` MAY remain in JSONB only when it is normalized against the owning `view_schema_id` and uses stable `field_key` values, ordered sort/filter entries, optional `group_by`, and normalized scalar values. Saved-view `layout_json` MAY describe presentation only. Neither `query_json` nor `layout_json` MAY be the authoritative source for saved-view identity, scope, ownership, authorization, or startup/default surface selection.

Multi-valued relationships MUST NOT remain only in `custom_attrs` or other JSONB fields. They MUST use typed link rows or dedicated child tables. Scalar convenience projections MAY exist, but they MUST NOT be the authoritative relationship store.

### 4.4 Promotion rule for recurrent incident-specific fields

A field MUST be promoted from `custom_attrs` or other JSONB storage to first-class structured state when both of the following are true:

- the field recurs across independent SoD-style incident workbooks or equivalent real-world incident trackers,
- analysts need the field for deterministic sort, filter, join, lifecycle tracking, dedupe, queueing, or reporting.

A field SHOULD remain flexible when it is low-frequency, framework-specific, integration-specific, presentation-only, or better modeled as a typed link or child record.

A promoted field MUST be available as named structured state on the authoritative record model or on a deterministic projection derived from authoritative structured inputs. It MUST NOT require JSON parsing of `custom_attrs`, scans of raw note text, or scans of raw evidence text on the hot path.

### 4.5 Current-profile promoted field sets

The current profile closes the incident-specific custom-metadata question with the following minimum promoted field sets. An implementation MAY add more structured fields. It MUST NOT demote these fields into JSONB-only storage.

The incident record MUST expose, as structured state:

- `tlp`,
- `current_phase`,
- `primary_external_case_ref`.

An incident MUST also persist a first-class `incident_key`. The deployment MUST enforce deployment-wide uniqueness on the canonical `incident_key` form produced by trimming leading and trailing Unicode whitespace and applying Unicode NFC normalization. The implementation MAY store that canonical form explicitly or derive it through an equivalent deterministic uniqueness mechanism.

Core 01 §3.3.5.3 owns the public incident resource and create contract. That contract also fixes `title`, `description`, `status`, `severity`, attribution fields, timestamps, `incident_version`, and `closed_at` as first-class incident state. This section governs only the additional incident-specific operational fields that were open in the source artifact.

The `task_request` model MUST expose, as structured state:

- `priority`,
- `task_kind`,
- optional `workstream`,
- optional `due_at`,
- optional `requester_party_text`,
- optional `blocked_reason`,
- optional `completed_at`,
- optional `external_ticket_ref`.

The current `host` record type is the structured home for recurrent asset-scoping fields. A host record MUST expose, as structured state:

- `location`,
- `os_platform`,
- `business_owner`,
- `criticality`,
- `containment_status`.

An `identity` record MUST expose, as structured state:

- `privilege_level`,
- `mfa_state`,
- `reset_status`.

The canonical indicator model and indicator system-view contract MUST expose, as structured state:

- observation-derived `first_observed_at`,
- observation-derived `last_observed_at`,
- `defanged_value` when applicable,
- optional `hash_algorithm`,
- optional `hash_value`.

An evidence record or its deterministic joined projection from authoritative object metadata MUST expose, as structured state:

- `requested_at`,
- `received_at`,
- `storage_ref`,
- `blob_hash`,
- `collector_party_text`,
- `source_party_text`.

The current profile does not standardize a standalone incident-contact record. Until such a model exists, requester, collector, and source-party promotion MUST use first-class structured text fields or an equivalent stable party reference. They MUST NOT live only in `custom_attrs`.

If the implementation exposes structured findings, communications logs, investigative queries, or forensic keywords as dedicated surfaces or typed artifact subtypes, their defining fields MUST also be structured state rather than JSON-only payload:

- findings: `state`, `confidence_score`, `owner_user_id`, optional `closed_at`,
- communications logs: `comm_type`, `audience`,
- investigative queries: `query_id`, `platform`, `purpose`, `query_text`, `created_by_user_id`,
- forensic keywords: `keyword_id`, `pattern`, `reason`.

The following SHOULD remain flexible or separately modeled unless a later profile standardizes them:

- framework-specific overlays such as ATT&CK, D3FEND, VERIS, kill-chain, or procedure mapping,
- threat-intel enrichment such as `asn`, `geo`, or `reputation`,
- campaign, attribution, or reporting-specific metadata such as `intrusion_set`, `threat_actor_name`, `target_country`, or `target_sector`,
- presentation-only flags such as `visual`, `pin`, or similar UI-state markers,
- long governance or free-text metadata such as bridge details, legal-hold narrative, confidentiality narrative, working-hours text, or notification-scope text.

## 5. Partial and uncertain data

The implementation MUST permit rough and uncertain input.

At minimum:

- `occurred_at` MAY be null,
- summary MAY be null if another field or attachment exists,
- host and account text MAY remain unresolved,
- confidence MAY be unset,
- details MAY remain unstructured text,
- `source_text` MAY remain unstructured text when a timeline event captures verbatim or excerpted source material.

The system MUST preserve original rough capture even after later normalization or resolution.

## 6. Mention, stub, and entity-origin contract

### 6.1 Separation invariant

An unresolved mention is **not** a low-quality entity.

An `entity_mention` is an observation tied to a source record and field. A stub entity is a host or identity record with its own `record_id` and lifecycle. The implementation MUST keep these object types separate.

The same source-observation-versus-canonical-record separation MUST apply to indicators. An `indicator_observation` is a source-bound occurrence tied to a record and field. A canonical `indicator` is a first-class record with its own identity and lifecycle.

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

### 6.5 Entity-mention lifecycle

`entity_mentions.resolution_status` MUST use the closed vocabulary `unresolved`, `resolved`, and `dismissed`.

- **`unresolved`**: the mention remains a source-bound observation awaiting analyst action. It participates in unresolved-mention flags and resolution queues.
- **`resolved`**: the mention remains a source-bound observation bound to an existing entity through resolution metadata such as `resolved_record_id`. The raw mention and normalized mention text remain preserved alongside the canonical link.
- **`dismissed`**: the mention remains preserved for provenance and history, but it is excluded from active resolution work. A dismissed mention MUST retain `raw_text`, `normalized_text`, stable row identity, source locator, and other provenance, MUST clear active resolution metadata such as `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`, and MUST NOT contribute to active relationship-cell values or unresolved-mention derived flags.

Ordinary user-facing restore of a dismissed mention MUST return it to `unresolved`. Recovering the exact pre-dismiss state, including any prior resolved target, MUST be handled by rollback rather than ordinary restore.

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

`origin_kind` MUST use the exact closed vocabulary defined in §18.

`origin_locator` MUST identify the source position deterministically, such as sheet/row/column for import or view/field for interactive entry.

### 7.2 File-based import provenance

When a record, mention, indicator observation, or alias is created or updated through file-based import, the implementation MUST persist provenance sufficient to identify the import source deterministically.

At minimum, file-based import provenance MUST include:

- `source_file_kind`,
- `source_content_sha256`,
- `parser_version`.

If the source is sheet-based or region-based, the provenance MUST also identify the selected sheet and rectangular region deterministically.

### 7.2.1 Import-session, unit, and mapping identity

When file-based import is implemented, the implementation MUST persist enough metadata to reconstruct which uploaded bytes, discovered source unit, and operator-approved mapping produced a given record, mention, indicator observation, or alias.

At minimum, file-based import-created or file-based import-updated data MUST be attributable to:

- `import_session_id`,
- `import_unit_id`,
- `mapping_fingerprint`,
- `parser_profile_id`,
- `source_file_kind`,
- `source_content_sha256`,
- `parser_version`,
- `origin_locator`.

For imported rows, mentions, aliases, and indicator observations, `origin_locator` MUST be derivable from the canonical import-unit locator plus deterministic row and column coordinates within that unit.

An `import_unit` MUST have a semantic identity equal to the tuple `source_content_sha256 + canonical_locator + parser_version`. If the implementation stores one derived key for that identity, it MUST compute that key as the SHA-256 of the canonical JSON serialization of `{source_content_sha256, locator, parser_version}`.

`mapping_fingerprint` MUST identify the operator-approved header-to-field mapping plan deterministically. The fingerprint input MUST include, at minimum, mapping contract version, target view schema identifier, header-row reference, data-start-row reference, unknown-column policy, and one ordered entry per source column containing source ordinal, raw imported header text, target field key, `entity_binding_mode`, and any declared transform or empty-value policy. The implementation MUST serialize that mapping input canonically with lexicographically sorted object keys and source columns ordered by `source_column_ordinal`, then hash the result as lower-hex SHA-256.

Preserved unknown columns from file-based import MUST NOT be keyed only by visible header text. The preserved structure MUST retain, at minimum, `import_unit_id`, `source_column_ordinal`, `source_header_text`, and raw value for each retained cell or row value. Duplicate visible headers MUST remain distinguishable by ordinal and source unit.

### 7.3 Entity provenance

For every host or identity record created outside direct entity sheets, the implementation MUST persist:

- `entity_origin`,
- an optional seed mention reference when created from a mention,
- seed identifier values as aliases or equivalent structured provenance,
- full `change_set` and revision lineage.

`entity_origin` MUST use the exact closed vocabulary defined in §18.

### 7.4 Indicator observation provenance

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

`origin_kind` MUST use the exact closed vocabulary defined in §18.

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

The base profile MUST NOT introduce a Notes-specific source table or Notes-only persistence model. Standalone note creation from the built-in Notes sheet and contextual `add linked note` actions MUST create the same underlying artifact record shape and use the same revision, tag, and link semantics.

A note artifact MUST be able to persist, at minimum, fields equivalent to `title` and `body`, plus generic tags through `record_tags` or an equivalent shared tag mechanism.

Lightweight free-text fields MAY remain on timeline events, hosts, identities, evidence, or other records. Text that requires standalone history, tags, search, or reuse across records MUST be modeled as an artifact with `artifact_type='note'`.

Analyst-work objects other than ad hoc notes MUST use either a distinct `artifact_type` or a distinct first-class `record_type`. In the current profile, `task_request` and `decision` are first-class `record_type`s, while `comm_log`, `handoff`, `status_review`, `lesson`, and current `hypothesis` tracking remain artifact-backed. None of these objects MAY overload `artifact_type='note'`.

### 10.2 Indicator contract

The base profile MUST model canonical indicators as first-class `indicator` records.

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

Canonical indicator identity MUST be derived from the incident plus the deterministic type-specific dedupe key or an equivalent stable canonical identity. `stix_pattern` MAY be stored for interoperability or export. It MUST NOT be the source of truth for canonical identity.

A canonical indicator record is not the raw source occurrence. Source-bound occurrences inside timeline, artifact, note, evidence, or other record fields MUST be stored separately as `indicator_observation` rows or an equivalent structured observation model. `indicator_observation` rows MUST remain distinct from generic `record_links`.

Every `indicator_observation` MUST bind to a source record and field and MUST preserve the raw observed value or text span. It MUST be able to store, at minimum, the observed text, parsed indicator type guess, normalized parse output when available, deterministic source locator, resolution status, optional resolved indicator reference, and attribution timestamps. Repeated identical observed values across different source rows MUST remain distinct observations with distinct provenance.

A source value such as a hostname or MAC address MAY participate in both the host or identity model and the indicator model. The implementation MUST NOT force one object type to subsume the other. A single source span MAY therefore yield both an `entity_mention` and an `indicator_observation`, with separate later resolution to a host or identity record and a canonical indicator record.

The base profile MUST preserve raw source text when capturing or resolving indicator observations. It MUST NOT require dedicated IOC columns on Timeline or other non-indicator sheets.

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

`first_observed_at` and `last_observed_at` are derived from indicator observations. They MUST remain distinct from lifecycle `valid_from` and `valid_to`.

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

This contract MUST remain stable across import, export, reporting, and future storage evolution.

### 10.3 Compromise assessments

Compromise state MUST be modeled as incident-scoped assessment history attached to a host or identity rather than as a static property on the entity.

`assessment_state` MUST use this closed vocabulary:

- `unknown`: no meaningful assessment yet; evidence is insufficient or not yet reviewed.
- `suspected`: there is credible evidence of compromise, but not enough to confirm.
- `confirmed`: compromise is supported by sufficient evidence.
- `disproven`: the subject was investigated and compromise is not supported for the scoped incident context.
- `cleared`: the subject was previously suspected or confirmed compromised, and is now assessed as no longer actively compromised as of this assessment.

`assessment_state` MUST represent evidentiary judgment about compromise for the subject at the time of assessment.

Operational-response terms such as `contained`, `isolated`, `disabled`, `reset`, or `monitored` MUST NOT be used as assessment states. Response posture belongs in linked timeline events, evidence, tasks, decisions, or a future separate response-state model.

`disproven` and `cleared` MUST remain distinct because the former records a negative investigation result and the latter records post-remediation or post-validation state after actual or plausible compromise.

Each assessment MUST carry:

- `assessment_state`,
- `assessed_at`,
- assessor,
- nullable `confidence_score` in the range `0..100`,
- rationale,
- optional supporting record links.

`confidence_score=NULL` means unspecified. `confidence_score` MUST express confidence in the correctness of the assessment entry itself. It MUST NOT be reused as severity, impact, priority, or urgency.

Assessment history MUST be append-only. A new judgment MUST append a new assessment record for the incident-scoped subject. It MUST NOT overwrite or silently mutate prior assessments.

The system MUST expose a deterministic derived `confidence_band` with the following mapping:

- `unset` when `confidence_score IS NULL`,
- `low` when `confidence_score BETWEEN 0 AND 39`,
- `medium` when `confidence_score BETWEEN 40 AND 69`,
- `high` when `confidence_score BETWEEN 70 AND 100`.

The system MUST expose a stable compromise-assessment system-view and API contract over assessment history with fields equivalent to:

- subject record and subject type,
- `assessment_state`,
- `confidence_score`,
- `confidence_band`,
- rationale,
- assessor,
- `assessed_at`,
- optional supporting link counts.

This contract MUST remain stable across filtering, reporting, and future storage evolution.

### 10.4 Analyst-work tracking

Tags and notes remain sufficient only for unassigned, non-transactional working material such as rough observations, ad hoc reminders, local analytic prose, and other low-stakes annotations that do not need accountable ownership, explicit lifecycle state, handoff durability, queue-oriented views, or later reconstruction as coordination artifacts.

Analyst work MUST be promoted out of tags or notes and into a structured analyst-work object when any of the following becomes true:

- it needs one accountable owner,
- it needs explicit lifecycle state,
- it needs due-date or blocker tracking,
- it must survive handoff or status review,
- it must be queryable as a queue, board, or owner workload view,
- it must explain or justify an operational action later,
- it must link multiple records as one durable unit of work.

#### 10.4.1 `task_request` record type

The base profile MUST model owned work and requests as first-class `task_request` records.

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

If `status='blocked'`, `blocked_reason` MUST be non-empty.

If `status='done'`, `completed_at` MUST be present and MUST NOT be earlier than `created_at`.

An active `task_request` with `status` not in `done` or `canceled` MUST NOT be ownerless.

If `linked_record_ids[]` or `decision_record_id` are persisted as denormalized convenience fields, authoritative cross-record association MUST still be representable through generic `record_links`.

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

A write that leaves `status` unchanged MUST be treated as a legal idempotent in-state write if the resulting record still satisfies all guards in this subsection.

The resulting authoritative row MUST satisfy all of the following guards:

- if `status='blocked'`, `blocked_reason` MUST be non-empty,
- if `status!='blocked'`, `blocked_reason` MUST be absent or empty,
- if `status='done'`, `completed_at` MUST be present and MUST NOT be earlier than `created_at`,
- if `status!='done'`, `completed_at` MUST be absent,
- if `status` is `open`, `in_progress`, or `blocked`, `owner_user_id` MUST be present.

When a successful write changes `status` away from `blocked`, the implementation MUST clear persisted `blocked_reason` before commit.

When a successful write changes `status` away from `done`, the implementation MUST clear persisted `completed_at` before commit.

When a successful write sets `status='done'` and the client did not supply `completed_at`, the implementation MUST set `completed_at` to the commit timestamp before commit.

Any other attempted write set that would leave the resulting record outside these invariants MUST be rejected as an illegal transition.

A requested transition from `done` to `canceled` or from `canceled` to `done` MUST be rejected as an illegal transition.

A successful lifecycle transition MUST be committed through the ordinary record-mutation path, increment `row_version`, append a `change_set` and mutation entries, update derived projections, and emit the ordinary `record_changed` signal for that record. A rejected lifecycle transition MUST return the stable lifecycle-validation error defined by Core 01 §3.3.6 and MUST leave authoritative persisted state unchanged.

#### 10.4.2 `decision` record type

The base profile MUST model rationale-bearing incident coordination choices as first-class `decision` records.

A `decision` record MUST carry, at minimum:

- stable `record_id`,
- owning `incident_id`,
- `decision_type` from the exact closed vocabulary defined in §18,
- `summary`,
- `status` from the exact closed vocabulary defined in §18,
- `owner_user_id`,
- `decided_at`,
- `rationale`.

A `decision` record MUST also be able to persist, at minimum, the following optional fields when present:

- `support_refs[]`,
- `affected_record_ids[]`,
- `supersedes_record_id`,
- `review_class`.

If `affected_record_ids[]` or `supersedes_record_id` are persisted as denormalized convenience fields, authoritative cross-record association MUST still be representable through generic `record_links`.

##### 10.4.2.1 Decision lifecycle machine

The current profile defines `decision.status` as a normative lifecycle machine rather than only a closed vocabulary. The `approved` state is an incident-coordination state. It MUST NOT be interpreted as a generalized reviewer or administrator approval gate for ordinary row edits.

The authoritative machine condition is determined by persisted `status` together with `supersedes_record_id` or an equivalent authoritative decision-to-decision link, `owner_user_id`, and `decided_at`.

The machine states mean:

- `proposed`: candidate decision under consideration,
- `approved`: accepted as the current intended course of action but not necessarily yet carried out,
- `rejected`: considered and declined,
- `executed`: carried out,
- `superseded`: overtaken before execution by a later accepted decision.

Allowed lifecycle triggers are record creation that sets an initial `status`, later record patches that change `status`, and the explicit supersession flow that links one decision to another.

Initial creation MAY set `status` to `proposed`, `approved`, `rejected`, or `executed`. Initial creation MUST NOT set `status='superseded'`.

The legal direct status transitions are:

- `proposed` -> `approved`, `rejected`, or `executed`,
- `approved` -> `executed`.

A direct write that leaves `status` unchanged MUST be treated as a legal idempotent in-state write if the resulting record still satisfies this subsection, except that a direct write whose requested or resulting `status` is `superseded` MUST still be rejected because `superseded` is reachable only through the explicit supersession flow.

`rejected`, `superseded`, and `executed` are terminal persisted statuses for direct status writes.

A direct write that would change `status` from `approved` to `rejected`, from `rejected` to `proposed`, from `executed` to any other value, or to `superseded` MUST be rejected as an illegal transition. A later contrary choice MUST be modeled as a new decision rather than as an in-place reversal of a `rejected`, `superseded`, or `executed` record.

The explicit supersession flow MUST satisfy all of the following:

- the superseding decision and the target decision MUST be different records in the same incident,
- the superseding decision MUST already have `status` `approved` or `executed`,
- the target decision MAY have `status` `proposed`, `approved`, or `executed`,
- the committed action MUST persist the supersession relation on the superseding decision through `supersedes_record_id` or an equivalent authoritative decision link,
- if the target decision has `status` `proposed` or `approved`, the same committed action MUST set the target decision's persisted `status` to `superseded`,
- if the target decision has `status` `executed`, the target decision MUST remain `executed`; a projection MAY still surface `decision.is_superseded=true` for that record, but the persisted `status` MUST NOT change away from `executed`.

A successful lifecycle transition or explicit supersession action MUST be committed through the ordinary record-mutation path, increment `row_version` on every changed record, append a `change_set` and mutation entries, update derived projections, and emit the ordinary `record_changed` signal for each changed record. A rejected lifecycle transition or supersession action MUST return the stable lifecycle-validation error defined by Core 01 §3.3.6 and MUST leave authoritative persisted state unchanged.

#### 10.4.3 Ownership and hot-path boundary

`ownership` is not a standalone current-profile record type.

Ownership MUST instead be modeled as a required `owner_user_id` or equivalent stable assignee field on coordination objects such as `task_request`, `decision`, `status_review`, `lesson`, handoff artifacts that require accountable follow-through, and evidence-request records when the evidence model is used in request mode.

Routine timeline, host, identity, note, and evidence editing MUST NOT require owner, approver, challenge, or checklist fields merely to preserve the primary capture path.

The base profile MUST NOT introduce a generalized approval workflow for ordinary row edits in order to support analyst-work coordination objects.

#### 10.4.4 Structured coordination artifact types

The current profile MUST keep the following coordination surfaces artifact-backed rather than promoting them to first-class `record_type`s:

- communications log artifacts with `artifact_type='comm_log'`,
- handoff artifacts with `artifact_type='handoff'`,
- status-review artifacts with `artifact_type='status_review'`,
- lesson or follow-up artifacts with `artifact_type='lesson'`.

A `comm_log` artifact MUST carry, at minimum:

- `comm_id`,
- `comm_type` from the exact closed vocabulary defined in §18,
- `timestamp_utc`,
- `audience`,
- `channel_or_meeting`,
- `summary`.

A `comm_log` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `decision_ids[]`,
- `action_task_ids[]`,
- `next_report_at`,
- `privilege_tag`.

A `handoff` artifact MUST carry, at minimum:

- `handoff_id`,
- `timestamp_utc`,
- `outgoing_owner_user_id`,
- `incoming_owner_user_id`,
- `current_state_summary`.

A `handoff` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `open_task_ids[]`,
- `open_decision_ids[]`,
- `open_risk_refs[]`,
- `next_checks`,
- `acknowledged_at`.

A `status_review` artifact MUST carry, at minimum:

- `status_review_id`,
- `timestamp_utc`,
- `review_owner_user_id`,
- `current_state_summary`.

A `status_review` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `blocked_task_ids[]`,
- `pending_evidence_ids[]`,
- `open_decision_ids[]`,
- `active_risks_summary`,
- `next_report_at`.

A `lesson` artifact MUST carry, at minimum:

- `lesson_id`,
- `timestamp_utc`,
- `summary`,
- `owner_user_id`.

A `lesson` artifact MUST also be able to persist, at minimum, the following optional fields when present:

- `follow_up_task_ids[]`,
- `closure_state`,
- `evidence_refs[]`.

These coordination artifact types MUST reuse the shared artifact model, revision history, tags, `record_links`, projections, and `view_schema_id` contracts. They MUST NOT introduce Notes-specific or artifact-type-specific standalone storage silos.

Any `*_ids[]` fields on coordination artifacts are denormalized convenience fields only. Authoritative many-to-many association MUST remain representable through generic `record_links` or dedicated child tables when per-link metadata is required.

#### 10.4.5 Hypothesis boundary

Current-profile hypotheses MUST remain artifact-backed, using either `artifact_type='hypothesis'` or a structured findings subtype. Hypotheses MUST NOT be promoted to a first-class `record_type` until repeated usage demonstrates a need for explicit competing-hypothesis tracking, support and contradiction sets, state transitions, or reviewer-visible hypothesis history that artifact-backed tracking cannot satisfy.

#### 10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces

The base profile does not require dedicated built-in sheets for findings, investigative queries, or forensic keywords.

If the implementation exposes any of these surfaces as typed artifact subtypes, contract-backed system views, or future first-class record types, their defining fields MUST be persisted as structured state rather than only in `custom_attrs`.

A structured finding or hypothesis subtype MUST be able to persist, at minimum:

- `statement`,
- `state`,
- `confidence_score`,
- `owner_user_id`,
- optional `closed_at`,
- optional supporting or contradictory record references.

A structured investigative-query subtype MUST be able to persist, at minimum:

- `query_id`,
- `platform`,
- `purpose`,
- `query_text`,
- `created_by_user_id`.

A structured forensic-keyword subtype MUST be able to persist, at minimum:

- `keyword_id`,
- `pattern`,
- `reason`,
- optional regex or literal flags,
- optional case-sensitivity flags.

### 10.5 Snapshot and reporting extension objects

If the Snapshot and Reporting Extension Profile is implemented, the domain model MUST also support structured metadata for:

- immutable snapshot descriptors,
- canonical export-model fields and blocks with stable export-model paths,
- versioned redaction profiles,
- versioned template contracts,
- redaction manifests,
- artifact release records.

At minimum, an immutable snapshot descriptor MUST persist the release tuple defined by Core 01 §10.2.

For rendered-output lifecycle, the authoritative persisted state MUST be artifact-scoped release records plus any bound approval records.

For this lifecycle, the logical output slot is the bound release tuple excluding `output_sha256`.

Each release record MUST also persist, at minimum:

- `release_state` from the exact closed vocabulary defined in §18,
- `approved_at`,
- `invalidated_at`,
- `published_at`,
- optional `invalidation_reason`.

A rendered output enters `pending_approval` when bytes and `output_sha256` exist for one immutable release tuple but the required approvals are not yet satisfied. It enters `approved` only when the required approvals are satisfied for that exact release record. It enters `published` only through an explicit publish action after approval. It enters `invalidated` when a different artifact for the same logical output slot supersedes it, when `output_sha256` changes for that slot, or when the implementation can no longer attest that the required approval set applies to that exact artifact.

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
- `output_sha256`,
- `release_state`,
- `approved_at`,
- `invalidated_at`,
- `published_at`.

If approval state is stored, it MUST bind to the release record rather than to mutable incident rows.

## 11. Type registries and view contracts

Type and icon choices MUST be driven by stable registry keys rather than hard-coded display text.

Host types, evidence types, indicator types, and similar display taxonomies MUST resolve through versioned registries keyed by stable IDs, with display labels, categories, icon keys, optional local overrides, and indicator-type-specific normalization or validation behavior where applicable.

Indicator-type registries MUST define normalization, validation, defanging, and dedupe behavior for each indicator type. Optional STIX export or mapping data MAY be present in the registry, but it MUST NOT control canonical indicator identity.

Entity-bearing columns and import mappings MUST carry stable `field_key` values and `entity_binding_mode` so that mention-versus-entity behavior never depends on a visible header such as `Host`, `User`, or `Account`.

View contracts and import mappings that support structured indicator capture from source text MUST declare the eligible `field_key` values or equivalent stable field identifiers. Visible labels MUST NOT determine indicator-capture behavior.

### 11.1 Saved-view contract

A saved view MUST be an incident-scoped configuration object over exactly one `view_schema`.

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

`owner_user_id` MUST be present for `private` and `shared` saved views. It MAY be null only for `system` saved views.

`query_json` MUST preserve saved-view sort, filter, and grouping state using stable `field_key` values, stable grouping identifiers, and normalized scalar values. It MUST NOT store visible labels as the authoritative identifier for any saved sort, filter, or grouping element. `query_json.filters[]` MUST use the exact filter predicate wire shape defined by Core 01 §3.3.4.1, preserve the operator-specific `arg` object, and persist canonical `filters[]` ordering by `field_key asc`.

A saved view created by duplicating another saved view MUST persist a full normalized copy of the source `view_schema_id`, `query_json`, and `layout_json`. The new saved view MUST NOT carry a runtime dependency on the source `saved_view_id`.

The schema MUST enforce immutability of `saved_view_id`, `incident_id`, and `view_schema_id` after creation.

### 11.2 Workbook startup preference objects

The schema MUST support two distinct workbook-startup preference objects:

- `user_workbook_preferences`, keyed by `(incident_id, user_id)`,
- `incident_workbook_preferences`, keyed by `incident_id`.

`user_workbook_preferences` MUST persist, at minimum:

- `incident_id`,
- `user_id`,
- nullable `home_sheet_ref`,
- `created_at`,
- `updated_at`.

`incident_workbook_preferences` MUST persist, at minimum:

- `incident_id`,
- nullable `default_sheet_ref`,
- `created_at`,
- `updated_at`,
- `updated_by_user_id`.

Both `home_sheet_ref` and `default_sheet_ref` MUST use the `sheet_ref` union defined by Core 01 §3.3.10.1 and MUST be nullable so the implementation can clear an invalid or unwanted pointer without deleting the preference object.

The schema MUST NOT overload saved-view rows with a generic `is_default` flag or any equivalent bit that conflates per-user startup selection, incident-wide fallback selection, or system-seeded queue surfaces.

## 12. Typed relationships

The implementation MUST provide a generic typed relationship store equivalent to `record_links`.

### 12.1 `record_link` object contract

A `record_link` is a stable, incident-scoped, directed, soft-deletable non-record mutation target between two first-class record-envelope rows in the same incident.

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

Stable `record_link_id` values are the mutation-target identity for history, rollback, and deterministic link-level reconstruction.

Both endpoints MUST reference first-class record-envelope rows. A `record_link` MUST NOT target `entity_mentions`, `indicator_observations`, or any other non-record mutation target.

`src_record_id` and `dst_record_id` MUST belong to the same `incident_id` as the `record_link` row. A `record_link` MUST NOT connect a record to itself.

A `record_link` is not a record-envelope row. It MUST NOT participate in the generic record-scoped delete or restore routes defined for first-class records, and it MUST NOT introduce a standalone `row_version` optimistic-concurrency contract in the current profile.

There MUST be at most one non-deleted `record_link` for the same `(incident_id, src_record_id, dst_record_id, link_type)` tuple.

For projection, count, and ordinary export purposes, a link is active only when the `record_link` row is not soft-deleted and both endpoint records are not soft-deleted. Soft-deleted links and links whose endpoint records are soft-deleted MUST remain reconstructible through history and rollback.

Merge-time repoint semantics from §9 apply to `record_links`. Incoming and outgoing active links to a losing entity MUST be repointed or deterministically recreated on the survivor in the same `change_set`, and duplicate active tuples created by repointing MUST be deduplicated without losing history.

All persisted `record_links` MUST be directed. The base profile MUST NOT require mirrored reverse rows, per-row symmetry flags, or inverse-type columns. Reverse traversal, reverse counts, and equivalent "linked from" UI affordances MUST be derived from the current directed links.

`record_links` MUST remain distinct from `entity_mentions` and `indicator_observations`. Mention-origin references MUST use `entity_mentions`. Source-bound indicator occurrences MUST use `indicator_observation` rows. A canonical indicator association MAY also be represented by a `record_link` when the applicable `link_type` is `references_indicator`.

### 12.2 Relationship vocabulary and canonical direction

The relationship vocabulary MUST be controlled by application code or equivalent authoritative configuration. It MUST NOT be free-text user input in the UI.

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

Client-visible reverse traversal or reverse counts MUST be derived from these canonical directions. They MUST NOT imply a second persisted link.

### 12.3 Link metadata semantics

The exact `provenance` tokens are defined in §18. The following semantics apply:

- `manual`: explicit analyst action created the link,
- `auto_match`: the link was created by the interactive host or identity auto-resolution flow defined by Core 03 §12.3,
- `import`: the link was created by a file or API import flow,
- `rollback`: the link was recreated or restored by rollback,
- `system`: the link was created by another system-managed action.

`provenance` is required.

`confidence` is a nullable `0..100` score about the correctness of a machine-produced link assertion. It MUST NOT be reused as evidence strength, assessment confidence, severity, impact, or priority.

A link created by explicit analyst action in the base profile MUST use `provenance='manual'` unless the applicable write path is the Core 03 §12.3 auto-resolution flow.

`provenance='auto_match'` is valid only for interactive host and identity auto-resolution. Such links MUST set `confidence=100`.

Explicit analyst-created links SHOULD leave `confidence` null unless a later profile defines a stable machine-produced scoring contract for that write path.

`note` MAY hold free text and operator context, but it MUST NOT carry conformance-critical semantics.

A timeline event MUST be able to relate to:

- hosts and identities via typed record links,
- canonical indicators via typed record links,
- unresolved host and account strings via entity mentions,
- source-bound indicator occurrences via `indicator_observation` rows,
- notes and other artifacts via typed record links,
- evidence via typed record links to evidence records.

Hosts, identities, and evidence records MUST also be able to relate to note artifacts and canonical indicators through the same generic typed relationship store.

When a note is created from timeline, host, identity, or evidence context, the association MUST be persisted through generic `record_links` rather than a Notes-specific foreign key on the source record.

A single note artifact MAY relate to zero, one, or many source records through `record_links`.

Task requests, decisions, and coordination artifacts MUST also be able to relate to timeline events, hosts, identities, canonical indicators, evidence records, notes, and each other through the same generic typed relationship store.

The implementation MUST NOT rely on task-specific, decision-specific, or handoff-specific foreign-key columns on timeline, host, identity, or evidence rows as the only supported linkage mechanism.

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

Cartulary defines two linked but separate evidence-related lifecycle machines:

- a blob-upload machine authoritative on `object_blobs.upload_state`,
- an evidence-custody and availability machine authoritative on `evidence_records.lifecycle_state` plus append-only custody events.

The blob-upload machine uses the exact closed vocabulary defined in §18.

The evidence-custody and availability machine uses the exact closed vocabulary defined in §18.

These machines MUST remain separate. `object_blobs.upload_state` MUST NOT be treated as the user-facing evidence lifecycle, and an evidence record MUST be able to exist with no blob at all.

The bridge rules are:

- an evidence record in `requested` or `pending_receipt` MAY have no `object_blob_id`,
- if an evidence record has a linked `object_blob_id`, it MUST NOT enter `available` unless that blob is in `upload_state='available'`,
- if a linked blob is `quarantined`, the evidence record MUST be `quarantined` or remain non-available,
- a blob in `pending` or `failed` MUST NOT make an evidence record appear attached, available, previewable, or released,
- `released` evidence MUST reference a blob in `upload_state='available'` when a blob is present.

Abandoned pending uploads MUST fail closed. A blob slot left in `pending` without successful finalization MUST NOT create or imply an attached evidence record, MUST NOT increment visible evidence counts, and MUST remain eligible only for retry, timeout handling, or administrative cleanup.

Each pending blob slot MUST persist `target_expires_at` and `pending_expires_at`. In the base profile, `target_expires_at` MUST be 60 minutes after upload-target issuance and `pending_expires_at` MUST be 24 hours after blob-slot creation. These timers MUST remain separate, and timeout handling MUST reuse `upload_state='failed'` rather than introduce a distinct expired lifecycle state.

The base profile MUST treat each pending blob slot as a single-upload lease. If the upload target expires before successful upload, retry requires a fresh blob-slot creation call. The base profile MUST NOT require same-slot upload-target refresh or resumable upload semantics.

The implementation MUST track `finalize_attempt_count` for explicit failed finalization attempts on a pending blob slot. Only unsuccessful explicit finalization attempts count toward this total; an idempotent replay after already-committed success MUST NOT consume retry budget. A pending blob slot MUST allow 3 failed finalization attempts. On the 4th failed attempt, it MUST transition to `upload_state='failed'`, persist `terminal_reason='finalize_retry_exhausted'`, and record `failed_at`. Later retry then requires a fresh blob slot.

A pending blob slot that is not successfully finalized by `pending_expires_at` MUST transition from `pending` to `failed`, persist `terminal_reason='pending_timeout'`, and record `failed_at`. For a failed blob slot that remains unattached to evidence, `cleanup_due_at` MUST be no later than 1 hour after terminal failure, orphaned blob bytes MUST be deleted by that deadline, and failed unattached slot metadata MUST remain queryable for at least 7 days before automatic hard deletion is allowed. `cleaned_up_at` MUST record completion of object-byte cleanup when that cleanup occurs.

If structured state becomes inconsistent, such as an evidence record claiming `available` or `released` while the linked blob is `pending`, `failed`, missing, or otherwise unusable, the implementation MUST fail closed: preview and download MUST be blocked and the record MUST surface as inconsistent until repaired by an explicit corrective action or re-finalization path.

An evidence record or its deterministic joined projection from authoritative object metadata MUST be able to expose, at minimum:

- `requested_at`,
- `received_at`,
- `storage_ref`,
- `blob_hash`,
- `collector_party_text`,
- `source_party_text`.

`requested_at` and `received_at` MUST remain first-class structured state. `storage_ref` MAY be a human-usable repository locator derived from the authoritative object-blob pointer, but it MUST be filterable and exportable without parsing `custom_attrs`.

When an evidence record represents requested or pending evidence rather than already received material, the model MUST be able to store `owner_user_id` or an equivalent stable assignee field so accountable follow-up does not require a separate timeline field. A linked `task_request` MAY also exist.

An implementation MAY persist a non-authoritative `sensitivity_class` on evidence records and MAY project that label onto exportable evidence-derived blocks. In the base profile, `sensitivity_class` is metadata for preview labeling, default export-redaction behavior, and audit or reporting only. It MUST NOT change incident authorization or hide live workspace content.

When custody narrative or handoff commentary is preserved as authoritative workflow state, it MUST be modeled as append-only custody events or equivalent child rows. A convenience `custody_note` or similar summary field MAY exist, but it MUST NOT be the sole authoritative custody history.

## 14. Schema invariants

### 14.1 Mention, indicator, provenance, and coordination fields

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
- evidence fields sufficient to persist separate blob-upload state, evidence lifecycle state, the bridge between `object_blob_id` and evidence availability, `target_expires_at`, `pending_expires_at`, `finalize_attempt_count`, `terminal_reason`, `failed_at`, `cleanup_due_at`, `cleaned_up_at`, `requested_at`, `received_at`, `storage_ref`, `blob_hash`, `collector_party_text`, `source_party_text`, and append-only custody events,
- snapshot and release fields sufficient to persist artifact-scoped `release_state`, approval binding, `approved_at`, `invalidated_at`, `published_at`, optional `invalidation_reason`, and deterministic identification of the logical output slot,
- coordination-artifact fields sufficient to persist `comm_type` and other `artifact_type`-specific metadata for `comm_log`, `handoff`, `status_review`, `lesson`, and optional current-profile `hypothesis` tracking,
- optional structured artifact subtype fields sufficient to persist findings, investigative queries, and forensic keywords when those surfaces are implemented.

For incidents created through the public base-profile route, the schema MUST support `status='active'`, `incident_version=1`, `closed_at=NULL`, and one committed create timestamp written to both `created_at` and `updated_at`; that public create route MUST bind both creation and initial update attribution to the creating local `user_id`.

Internal-user and incident-membership state MUST remain deployment-local authorization state. Whole-incident portability import MAY map historical actor descriptors to existing local users, but import MUST NOT synthesize login-capable users, deployment-admin flags, or active memberships without explicit deployment-local administrative action.

### 14.2 Rollback granularity substrate

A conformant history schema MUST include both:

- immutable `change_set` attribution records, and
- mutation-entry records at target granularity.

`record_revisions` MAY retain row snapshots such as `before_json` and `after_json` for audit and whole-row restore. They MUST NOT be the sole rollback substrate.

Whole-row restore driven by a `record_revisions` snapshot MUST restore only authoritative row-backed fields for the selected `record_id` and revision number. It MUST NOT, by itself, recreate or delete non-row mutation targets such as `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, or evidence associations.

### 14.3 Mutation targets

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

Stable mutation targets MUST use deterministic serialization, for example `record_link:<record_link_id>`. Composite targets MUST serialize canonically, for example `record_tag:<record_id>:<tag_id>`.

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

The history substrate MUST also support a stable opaque public `history_entry_ref` for any row-centric logical history item that maps to exactly one reversible mutation target. The public rollback interface MUST round-trip that reference without exposing storage-primary-key mutation-entry identifiers.

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
- exact and normalized lookup over canonical indicator values and deterministic indicator dedupe keys,
- full-text search using a configuration appropriate for IR tokens rather than English stemming,
- link traversal by source and destination,
- traversal from canonical indicators to source-bound observations and linked records,
- filtering and sorting over `task_request` owner, status, priority, due, blocker, and last-updated state,
- filtering and sorting over `decision` owner, status, type, and `decided_at`,
- fuzzy matching over alias and unresolved mention text,
- array or equivalent containment lookups for denormalized tag and label sets.

## 17. Domain invariants

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

## 18. Canonical closed-vocabulary registry

This section owns the exact token sets for closed vocabularies that are persisted in structured state or surfaced through contract-backed views, portability bundles, or public API payloads. Earlier sections that described a vocabulary as "equivalent to" these values are resolved by this registry. A conformant implementation MUST persist and emit the exact tokens listed here.

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
| `evidence_records.lifecycle_state` | `requested`, `pending_receipt`, `received`, `available`, `quarantined`, `released` |

A current-profile implementation MUST NOT persist alternate spellings, display labels, or semantically equivalent synonyms for any token listed here in authoritative structured state. Client and export presentation layers MAY map these tokens to display labels, but the canonical token value MUST remain stable in stored state and machine-readable payloads.
