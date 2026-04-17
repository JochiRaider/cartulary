# Appendix C: Schema Reference and DDL Source Extract

This appendix is **non-normative**.

No table name, column name, helper column, generated column, functional index, trigger, or DDL topology in this appendix is authoritative unless the same requirement is restated in Core 01 through Core 04.

It preserves the schema-oriented source extract, including ER diagram, concrete DDL sketch, and indexing notes.

## 7. Postgres schema proposal

### ER diagram

```mermaid
erDiagram
    USERS ||--o{ INCIDENT_MEMBERSHIPS : participates
    INCIDENTS ||--o{ INCIDENT_MEMBERSHIPS : has
    USERS ||--o{ AUTH_IDENTITIES : maps
    AUTH_PROVIDERS ||--o{ AUTH_IDENTITIES : issues

    INCIDENTS ||--o{ RECORDS : contains
    USERS ||--o{ RECORDS : creates

    RECORDS ||--|| TIMELINE_EVENTS : "is a"
    RECORDS ||--|| HOSTS : "is a"
    RECORDS ||--|| IDENTITIES : "is a"
    RECORDS ||--|| PARTIES : "is a"
    RECORDS ||--|| ARTIFACTS : "is a"
    RECORDS ||--|| INDICATORS : "is a"
    RECORDS ||--|| EVIDENCE_RECORDS : "is a"
    RECORDS ||--|| COMPROMISE_ASSESSMENTS : "is a"

    RECORDS ||--o{ ENTITY_ALIASES : aliases
    RECORDS ||--o{ ENTITY_MENTIONS : source
    RECORDS ||--o{ INDICATOR_OBSERVATIONS : source
    INDICATORS ||--o{ INDICATOR_OBSERVATIONS : resolves
    INDICATORS ||--o{ INDICATOR_STATE_INTERVALS : tracks
    RECORDS ||--o{ RECORD_LINKS : src
    RECORDS ||--o{ RECORD_LINKS : dst
    RECORDS ||--o{ RECORD_TAGS : tagged
    TAGS ||--o{ RECORD_TAGS : applied

    CHANGE_SETS ||--o{ CHANGE_SET_MUTATIONS : contains
    RECORDS ||--o{ CHANGE_SET_MUTATIONS : anchors
    CHANGE_SETS ||--o{ RECORD_REVISIONS : groups
    RECORDS ||--o{ RECORD_REVISIONS : revised

    OBJECT_BLOBS ||--o{ EVIDENCE_RECORDS : backs
    EVIDENCE_RECORDS ||--o{ EVIDENCE_CUSTODY_EVENTS : tracks

    VIEW_SCHEMAS ||--o{ SAVED_VIEWS : shapes
    REFERENCE_PACKS ||--o{ TYPE_REGISTRY_ENTRIES : provides
    INCIDENTS ||--o{ SAVED_VIEWS : has
```

### Key tables

One conformant non-normative realization uses a **`records` envelope table**. Every user-visible object gets one row there. That costs an extra join but buys strong generic linking, tagging, revisions, and consistent UI routing. Canonical indicators fit this same envelope pattern, while source-bound `indicator_observations` remain separate structured observation rows. In this realization, built-in view behavior lives in `view_schemas` and reference-pack tables rather than in visible headers or tab names.

### Illustrative realization notes for view-query sort and group contract

One conformant non-normative realization of the current sort and grouping contract needs explicit storage for:

- ordered default-sort tuples rather than one `default_sort_key` plus one `default_sort_direction`,
- an explicit per-view sortable-field whitelist,
- optional field-level `header_sort_field_key` metadata inside field-registry storage or equivalent child rows,
- canonical saved-view persistence where `query_json.sort` stores only normalized user sort overrides, `[]` is the only stored representation of `no user sort override`, and inactive grouping omits `query_json.group_by` rather than storing JSON `null`.

One additional non-normative realization note is now explicit: generated comparison columns, expression indexes, token tables, or `tsvector`-style realizations are conformant only when they reproduce the Core 01 comparison and tokenization contract exactly. Database-default locale, collation, or tokenizer behavior is not authoritative when it differs from the normative contract.

### Illustrative realization notes for snapshot-stable cursor pagination

One conformant non-normative realization of `snapshot_stable` cursor continuation needs explicit runtime state for:

- a snapshot descriptor or equivalent opaque snapshot anchor,
- a deterministic continuation position within that snapshot,
- `last_used_at` or equivalent inactivity-expiry state for the cursor chain.

A conformant realization can use:

- an exported or equivalent MVCC snapshot descriptor,
- a server-side materialized ordered `record_id` list plus continuation position,
- an equivalent snapshot descriptor plus deterministic seek anchor that reproduces the same externally observable continuation behavior.

Snapshot descriptors and continuation state are deployment-local runtime state. They are not incident-portability content, not part of the authoritative incident source model, and not a reason to promote projection tables into source-of-truth tables. Projection tables remain disposable caches and do not become the authoritative snapshot contract.

### Illustrative realization notes for canonical view-row wire family

One conformant non-normative realization of the current full-row and sparse-patch contract needs explicit serializer and projection discipline for the canonical row family:

- no new authoritative source tables are required solely to satisfy `view_row_v1` or `view_row_patch_v1`,
- the serializer can reconstruct `view_row_v1` from the active `view_schema` field registry, the projection row, and authoritative joined source tables,
- full-row serialization in one conformant realization emits explicit JSON `null` for schema-declared non-technical fields whose authoritative value is null rather than omitting those fields from `cells`,
- sparse collaboration patches can be materialized as field-subset row objects keyed by `field_key` rather than as flat cell maps unrelated to the canonical row envelope,
- `group_values` can be derived alongside the row and need not live as authoritative columns on base record tables so long as full-row and sparse-patch derivation remains deterministic for the active schema.

### Illustrative realization notes for collaboration-array canonicalization

One conformant non-normative realization of the current collaboration-array contract needs explicit emitter discipline for the canonical public arrays:

- no new authoritative source tables are required solely to satisfy collaboration-array canonicalization,
- `presence_snapshot.payload.presences[]` can be materialized from a map keyed by exact `connection_id`, with expired presence rows pruned before serialization and the remaining entries sorted before serialization under the Core 01 exact-identifier ordering contract,
- `record_changed.payload.changed_field_keys[]` can be materialized from a set of exact public `field_key` identifiers and sorted before serialization under the Core 01 exact-identifier ordering contract,
- `record_changed.payload.affected_views[]` can be materialized from a map keyed by base `view_schema_id` and sorted before serialization under the Core 01 exact-identifier ordering contract,
- this realization does not rely on database-default locale behavior, query row order, map iteration order, or object insertion order to satisfy the public collaboration wire contract.

### Illustrative realization notes for mention/stub provenance

One conformant realization can include the following explicit contract fields beyond the high-level tables named above:

- `entity_mentions` rows in one conformant realization store `source_field_key`, `origin_kind`, and `origin_locator`, plus resolution metadata such as `resolved_at`, `resolved_by_user_id`, and `resolution_method`.
- Host and identity rows in one conformant realization store `entity_origin` and structured provenance, including an optional seed mention reference when the entity was created from a mention.
- `view_schemas.writeback_contract` and import mappings in one conformant realization store `entity_binding_mode` per entity-bearing field. Same-field-conflict-capable write-back fields in the same realization also store `conflict_resolution_class` per `field_key`.
- Repeated mentions remain separate rows in this realization; repeated entity-origin inputs can upsert the same entity when exact-match rules select a unique active target.

### Illustrative realization notes for import-session and mapping reconstruction

This appendix illustrates a realization with one explicit import-session and mapping-reconstruction section so the durable read resources in Core 01 §17.2 do not depend on inferred field names or lossy persistence:

- `import_sessions` or equivalent structured rows in one conformant realization persist the durable session fields already required by Core 03, including `import_session_id`, incident anchor, creator attribution, source-file identity, parser identity, assistant profile, durable `session_status`, `selected_unit_ids[]`, `blocking_diagnostics[]`, and `nonblocking_warning_codes[]`.
- `import_units` or equivalent structured rows in one conformant realization persist the durable unit fields already required by Core 03, including locator identity, source rectangle, row-boundary refs, inferred dimensions, `warning_codes[]`, durable `unit_status`, and optional `mapping_fingerprint`.
- Each approved unit in one conformant realization persists one approved-mapping parent object or an equivalent structured row set sufficient to reconstruct the exact Core 01 §17.2 `approved_mapping` object.
- That approved-mapping realization persists one child row per discovered source column carrying at minimum `source_column_ordinal`, `source_header_text`, `field_key`, `entity_binding_mode`, `transform_id`, `transform_options`, and `empty_value_policy`.
- Read-side reconstruction of `approved_mapping.source_columns[]` in this realization remains exhaustive, ordered by `source_column_ordinal`, and lossless for deterministic `mapping_fingerprint` recomputation.

### Illustrative realization notes for incident-scoped parties

One conformant realization can include an explicit party model now that the core closes requester, collector, source, and related coordination identity on a standalone incident-scoped record:

- One conformant realization includes `party` in `records.record_type`.
- `parties` in one conformant realization persist required `display_name` and `party_kind`, plus optional `organization_name`, `role_title`, `primary_email`, `timezone_name`, `external_ref`, and `notes`.
- Task-request, evidence, and optional coordination-artifact refs such as requester, collector, source, audience, or attendee refs in this realization use same-incident `party_id` values while preserving raw or source text separately.
- Exact-match reuse for direct party creation or explicit create-from-text is incident-scoped and limited to unique exact matches on normalized `primary_email` or `external_ref`; display name, organization, role title, and phone-like text are suggestion inputs only.
- The current profile does not standardize party merge or phone-based dedupe.

### Illustrative realization notes for coordination collection fields

This appendix illustrates one explicit mapping table for the coordination-surface collection families closed by the current profile:

| Field set | Public family | `item_kind` | `item_ref` form | Authoritative target or child identity | Storage note |
| --- | --- | --- | --- | --- | --- |
| `comm_log.decision_ids[]`, `comm_log.action_task_ids[]`, `handoff.open_task_ids[]`, `handoff.open_decision_ids[]`, `status_review.blocked_task_ids[]`, `status_review.pending_evidence_ids[]`, `status_review.open_decision_ids[]`, `lesson.follow_up_task_ids[]`, `lesson.evidence_refs[]` | `record_ref` | `record_ref` | `record_ref:<linked_record_id>` | same-incident active target `record_id` constrained by the owning `field_key` | One conformant storage realization keeps the authoritative association in `record_links` with field-derived relation token `references_record`. |
| `comm_log.audience_party_ids[]`, `comm_log.attendee_party_ids[]` | `party_ref` | `party_ref` | `party_ref:<party_id>` | same-incident active `party_id` | The public target identifier remains `party_id` even when storage uses a record envelope internally. |
| `handoff.open_risk_refs[]` | `risk_ref` | `risk_ref` | `risk_ref:<risk_ref_id>` | dedicated child-row `risk_ref_id` scoped to one `handoff` record | `risk_ref_text` remains source-preserving text on the child row; the current profile does not imply a first-class `risk` record type. |

Additional contract notes:

- The base profile does not introduce public `decision_ref`, `task_ref`, or `evidence_ref` item kinds for these coordination fields.
- `add_risk_ref.risk_ref_text` uses `single_line_title_v1` and duplicate adds coalesce by normalized `risk_ref_text` within one `handoff` record.
- Same-field conflict payloads for these coordination collections reuse the same family shapes on `collection_value_v1`; the wire contract does not fall back to raw string arrays.

#### Illustrative child-row sketch for handoff risk refs

One conformant realization of the `risk_ref` family is a dedicated child-row table or equivalent child-row structure with these properties:

- each child row belongs to exactly one parent `handoff` record;
- minimum child-row state is the parent `handoff` `record_id`, stable `risk_ref_id`, source-preserving `risk_ref_text`, and normalized `risk_ref_text` used for duplicate coalescing;
- `risk_ref_id` is the child identity that backs public `item_ref="risk_ref:<risk_ref_id>"`;
- the implementation can enforce at most one active child row per `(handoff_record_id, normalized_risk_ref_text)`;
- the child row is not a record-envelope row, does not consume a `record_id`, and does not imply a standalone public route family or a reusable future `risk` object;
- add/remove history remains in the existing mutation/history substrate, while whole-row restore remains parent-row behavior on the `handoff` artifact.

### Illustrative realization notes for Timeline supersede replacement relation

This appendix illustrates one explicit non-normative realization note for Timeline supersession with direct replacement:

- `record_links.link_type='supersedes'` is valid for both `decision -> decision` and `timeline_event -> timeline_event` endpoint pairs.
- For Timeline rows, the authoritative replacement relation is one active `supersedes` link from the replacement Timeline row to the superseded Timeline row.
- A conformant realization can enforce at most one active incoming Timeline `supersedes` link for any one superseded Timeline row, while still allowing one replacement Timeline row to supersede multiple older Timeline rows.
- Because the generic `record_links` table does not encode endpoint record types, that Timeline-specific cardinality constraint can require a trigger, helper columns, or an equivalent realization.
- If the implementation materializes a convenience projection, `timeline_grid_projection` can add nullable `replacement_record_id uuid`, but that field remains read-only and derived from the authoritative `record_links` relation.
- The current profile does not require a default visible grid column, filter key, grouping key, or default index for `timeline.replacement_record_id`.

### Illustrative realization notes for canonical indicators

One conformant realization can include the following explicit fields and separations for indicators:

- One conformant realization includes `indicator` in `records.record_type`.
- Canonical indicators in one conformant realization store `indicator_type`, `value_kind`, canonical display value, `normalized_value` when applicable, deterministic incident-scoped `dedupe_key`, optional `defanged_value`, optional hash fields, and optional `stix_pattern`.
- In the current profile, the stored canonical-identity inputs are `indicator_type`, `value_kind`, canonical display value, and `normalized_value` when applicable, plus the pair `hash_algorithm` and `hash_value` when both are populated and incorporated into the canonical dedupe key; `defanged_value` and `stix_pattern` are stored fields, not canonical identity inputs.
- Source-bound `indicator_observations` in one conformant realization store `source_record_id`, `source_field_key`, `origin_kind`, `origin_locator`, observed text, optional parsed indicator type and normalized candidate, deterministic span or selection locator when the source is inline text, resolution metadata, and attribution.
- Indicator lifecycle windows in one conformant realization are stored separately from observations in append-only `indicator_state_intervals` or equivalent structured rows keyed to the canonical indicator.
- `indicator_grid_projection` in one conformant realization is keyed by canonical indicator `record_id`, not by source artifact or observation identity.

### Illustrative realization notes for compromise assessments

One conformant realization can include the following explicit fields and separations for compromise assessments:

- In one conformant realization, `assessment_state` uses the closed vocabulary `unknown`, `suspected`, `confirmed`, `disproven`, and `cleared`.
- Operational-response terms such as `contained`, `isolated`, `disabled`, `reset`, or `monitored` do not appear in `assessment_state` in this realization.
- Compromise assessments in one conformant realization store nullable `confidence_score` in the range `0..100` and expose deterministic derived `confidence_band` values of `unset`, `low`, `medium`, or `high`.
- Compromise assessment history in this realization remains append-only and incident-scoped to a host or identity subject rather than overwriting a mutable compromise flag on the subject row.
- Assessor attribution in this realization is preserved either on the assessment row itself or through the owning assessment record envelope.
- `assessment_grid_projection` in one conformant realization is keyed by assessment `record_id` or equivalent stable assessment-row identity, not by a mutable subject-state overwrite.

### Illustrative realization notes for rollback granularity

One conformant history realization includes a mutation log in addition to row-snapshot revisions.

- `change_sets` remain the attribution unit for actor, source, reason, and transaction grouping.
- A `change_set_mutations`-style table or equivalent in one conformant realization records reversible entries at mutation-target granularity and orders them deterministically within the parent `change_set`.
- Illustrative mutation targets in this realization include row-field edits, `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, `indicator_state_intervals`, compromise assessments, evidence associations, and merge/repoint fan-out.
- Stable mutation target identities in this realization use a canonical target-kind-specific serialization. Composite targets serialize deterministically, for example `record_tag:<record_id>:<tag_id>`.
- `record_revisions` in one conformant realization can retain `before_json` / `after_json` row snapshots for audit and whole-row restore, but they are not the sole rollback substrate.

One conformant realization of the base-profile destructive-operation concurrency contract is an internal lock service or transaction-scoped database locking keyed to the protected first-class `record_id` set computed by Core 01 §3.3.5.0. Advisory locks, row-level locks, or an equivalent internal coordinator are all acceptable if they preserve the owner-defined public behavior. In all such patterns, lock acquisition proceeds in canonical ascending `record_id` order, fails fast rather than queueing, remains deployment-local runtime state rather than portable incident data, and releases on commit, rollback, or request termination.

### Illustrative realization notes for same-field conflict resolution

One conformant realization of same-field conflict handling includes one explicit contract hook beyond base row-versioning:

- `view_schemas.writeback_contract` in one conformant realization stores `conflict_resolution_class` per write-back-capable `field_key`.
- The closed vocabulary in the owner sections is `atomic_replace`, `text_compare_merge`, and `collection_review`. In this realization, unknown or omitted values fall back to `atomic_replace`.
- A conformant same-field conflict payload realization carries `conflict_token`, `record_id`, `field_key`, `conflict_resolution_class`, `base_row_version`, `current_row_version`, `client_value`, `server_value`, `server_updated_by`, and `server_updated_at`. For `text_compare_merge`, the base profile uses `base_value` rather than `base_revision_ref` alone, and `client_value`, `server_value`, `base_value`, and optional `suggested_merged_value` are raw text or `null`. For `collection_review`, `client_value`, `server_value`, and `base_value` use `collection_value_v1`.
- `text_compare_merge` denotes a plain-text, line-based merge-capable conflict class. Same-field detection stays keyed by `field_key`, not by textual subrange. For merge computation only, `null` becomes the empty string and `CRLF` or `CR` normalize to `LF`. A clean deterministic suggestion can be surfaced as `suggested_merged_value` only when normalized client and server change hunks do not overlap and do not both insert at the same base boundary.
- Explicit `merged_value` resolution for `text_compare_merge` accepts only the final text scalar or `null`; it is not a diff script, token list, AST, or field-specific merge object.
- Client implementations in one conformant realization keep same-field conflicts in a conflict queue keyed by the canonical composite `record_id:field_key` rather than mixing them into the transient retry queue.
- The base profile does not require an authoritative persisted conflict-draft table. Unresolved conflict drafts remain client-local unsaved state until the analyst explicitly resolves them.

### Illustrative realization notes for direct-scalar timestamp contracts

One conformant realization includes one explicit contract hook for writable temporal scalars:

- The field registry for one conformant realization stores `direct_scalar_contract_id` and `clearable` explicitly for writable direct temporal scalar fields.
- The base profile currently closes exactly one such contract, `timestamp_instant_v1`.
- `timestamp_instant_v1` admits only RFC 3339 timestamp strings with an explicit timezone designator, compares canonical equality in UTC `Z` form, and uses explicit JSON `null` as the only authoritative clear representation when the bound field declares `clearable=true`.
- Original timestamp text, original offset, and precision caveats remain source-preserving text or metadata rather than part of the canonical scalar column.

### Illustrative realization notes for direct-reference scalar contracts

One conformant realization includes one explicit contract hook for writable direct-reference scalars:

- The field registry for one conformant realization stores `direct_reference_contract_id` and `clearable` explicitly for writable direct-reference scalar fields.
- The base profile currently closes exactly two such contracts, `same_incident_party_ref_v1` and `same_incident_decision_ref_v1`.
- `same_incident_party_ref_v1` admits only exact `party_id` strings as non-null input and uses explicit JSON `null` as the only authoritative clear representation when the bound field declares `clearable=true`.
- `same_incident_decision_ref_v1` admits only exact `record_id` strings that resolve to same-incident active `decision` records and uses explicit JSON `null` as the only authoritative clear representation when the bound field declares `clearable=true`.
- If `task.decision_record_id` is realized as a denormalized convenience scalar, set and clear operations remain atomically consistent with the authoritative `record_links` representation rather than creating dual authority.

### Illustrative realization notes for reference-pack lifecycle

The schema sketch models reference-pack lifecycle through two linked table sets:

- `reference_packs` for version-scoped verification and availability state,
- `reference_pack_activation_state` and `reference_pack_attestations` for the active-version pointer and import or activation events.

The version-scoped public durable conditions are `staged`, `verified_available`, `disabled`, `failed`, and `missing`. `active` is not a stored version-state token; it is derived when `reference_packs.status='available'`, `verification_result='passed'`, and `reference_pack_activation_state.active_version` for the same `pack_key` equals that `pack_version`. The public durable condition `verified_available` is derived when `reference_packs.status='available'`, `verification_result='passed'`, and the version is not currently active for its `pack_key`. The owner sections treat `verified_available` as the activation-eligible condition, and this derivation therefore never leaves a disabled, failed, or missing version as the active pointer for its `pack_key`.

Core 01 §11.3, Core 01 §11.3.1, Core 01 §11.4, and Core 04 §4.1 own the public lifecycle semantics, durable conditions, and verification rules; this appendix illustrates one storage derivation only. Core 01 §17.4 owns the public `reference_pack_version resource`; this appendix describes one storage realization only and does not own the public JSON shape. Public `pack_version_state` is derived from storage `status`, `verification_result`, and the activation pointer. Public `active` is the derived activation-pointer boolean rather than a stored version token. Public `payload_sha256` in this realization can be reconstructed from one or more stored payload SHA-256 digests or an equivalent canonical aggregate digest.

### Illustrative realization notes for snapshot artifact lifecycle

This appendix illustrates three distinct storage surfaces so snapshot-boundary state, release-boundary state, and approval state do not collapse into one implied artifact row:

- snapshot descriptor rows in one conformant realization persist only the snapshot-boundary fields from revised REQ-02-140: `snapshot_id`, `incident_id`, `created_by_user_id`, `created_at`, `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, and `export_model_sha256`; they do not carry template selectors, redaction-profile selectors, release-state metadata, or rendered-output hashes.
- release rows in one conformant realization persist the release-owned fields from revised REQ-02-145: `release_id`, `incident_id`, `snapshot_id`, template selectors, redaction-profile selectors, `output_kind`, `release_scope`, `output_sha256`, `release_state`, creator attribution, lifecycle timestamps, and optional `invalidation_reason`.
- approval rows in one conformant realization bind to `release_id` rather than to mutable incident rows.

A public release resource can expose snapshot-boundary fields such as `snapshot_at`, `source_change_set_high_watermark`, `derivation_version`, and `export_model_sha256` by deterministic join to the bound snapshot descriptor rather than by forcing redundant release-row storage. In this realization, a superseding render for the same logical output slot or a byte change creates a new `pending_approval` candidate and does not inherit prior approval state.

### Illustrative realization notes for blob-upload and evidence lifecycle

Evidence-access `media_class` and `preview_kind` are contract vocabularies that can be derived from authoritative object metadata and server-side inspection. This appendix does not imply that same-named physical columns are required.

The schema sketch keeps blob upload and evidence lifecycle separate:

- `object_blobs.upload_state` holds conditions equivalent to `pending`, `available`, `failed`, or `quarantined`,
- `evidence_records.lifecycle_state` holds states equivalent to `requested`, `pending_receipt`, `received`, `available`, `quarantined`, or `released`,
- the bridge between them is the optional `object_blob_id` plus custody events.

The core now also requires `object_blobs` to persist the incident anchor, the route-scoped blob-create idempotency key, the accepted upload contract, the observed object metadata used for finalization checks, and the structured timeout and cleanup fields `target_expires_at`, `pending_expires_at`, `finalize_attempt_count`, `terminal_reason`, `failed_at`, `cleanup_due_at`, and `cleaned_up_at`.

Timeout, retry exhaustion, and terminal contract mismatch remain instances of `upload_state='failed'`; the sketch does not need a separate expired state. The base-profile blob slot behaves as a single-upload lease with a short-lived upload target and a longer pending-slot timeout.

A blob slot left in `pending` without successful finalization is not treated as attached evidence in this realization. A declared-size or expected-hash mismatch fails finalization and leaves no attached evidence. An evidence row does not surface as available, previewable, or released while its linked blob is `pending`, `failed`, or missing. If structured state becomes inconsistent, the application fails closed for preview and download until repaired.

The deployment-level resource ceilings closed by the current core are configuration keys, not new authoritative schema columns. Core 04 §12.3.1 owns the numeric registry for blob-create ceilings, structured-import ceilings, archive extraction and compression limits, reference-pack and incident-bundle extracted-byte overrides, and preview ceilings. The schema remains responsible for accepted upload contract, observed object metadata, timeout and cleanup state, terminal reasons, and evidence or blob lifecycle state only.

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email citext NOT NULL UNIQUE,
    display_name text NOT NULL CHECK (char_length(display_name) <= 256),
    password_hash text NOT NULL,
    mfa_required boolean NOT NULL DEFAULT true,
    totp_secret_enc bytea,
    webauthn_credentials jsonb NOT NULL DEFAULT '[]'::jsonb,
    is_active boolean NOT NULL DEFAULT true,
    is_deployment_admin boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid REFERENCES users(id),
    last_login_at timestamptz,
    user_version bigint NOT NULL DEFAULT 1
);

-- Informative: public `POST /api/v1/users` treats omitted `mfa_required`
-- as `true`, omitted `is_deployment_admin` as `false`, and always
-- initializes `is_active` to `true`. The public create contract does not
-- accept client-supplied `is_active`, and the first deployment admin is
-- realized through the deployment-local bootstrap-admin manifest contract
-- rather than through an unauthenticated public bootstrap route.
-- The illustrative `char_length(...)` checks above capture the bounded-size
-- part of the writable-string contract. Unicode NFC normalization,
-- leading/trailing Unicode-whitespace trimming, control-character rejection,
-- clear-to-null behavior, and exact Unicode-scalar counting remain
-- service-layer rules owned by Core 01 §18.
-- In the base profile, `users.email` is the authoritative local login
-- identifier. `POST /api/v1/auth/login` retains the wire member name
-- `username` for v1 compatibility only and does not imply a second
-- persisted local username namespace.
-- Non-normative example: `email citext UNIQUE` is one conformant
-- realization of the local-user email comparison and uniqueness contract.
-- An equivalent generated comparison column or functional unique index
-- over the same `email_address_v1` comparison substrate is also
-- conformant; `citext` itself is not required.

### Informative note on deployment-local credential lifecycle realization

One conformant realization is to keep credential lifecycle state in deployment-local
administrative tables or columns such as `users.password_changed_at`,
`users.totp_enrolled_at`, wrapped or encrypted-at-rest active TOTP secret
material, one pending-enrollment row keyed by `enrollment_id`, and one
non-reversible bootstrap-token lookup substrate such as `bootstrap_token_hash`.
That illustrative name is one conformant realization of the required
non-reversible lookup substrate rather than a required column name or required
co-located storage shape. Those rows remain deployment-local auth state. They
do not participate in the record-envelope model, workbook mutation routes, or
incident-portability bundles.

One conformant realization of the one-time bootstrap marker is a dedicated
deployment-local table written in the same transaction as the first created
admin user:

```sql
CREATE TABLE deployment_bootstrap_state (
    slot text PRIMARY KEY CHECK (slot = 'first_deployment_admin'),
    bootstrap_schema_id text NOT NULL CHECK (bootstrap_schema_id = 'cartulary.bootstrap_admin.v1'),
    bootstrap_artifact_id uuid NOT NULL UNIQUE,
    artifact_sha256 bytea NOT NULL,
    created_user_id uuid NOT NULL REFERENCES users(id),
    consumed_at timestamptz NOT NULL DEFAULT now()
);
```

In that realization, `artifact_sha256` is computed from the exact raw manifest
bytes consumed, and the same commit would also append one deployment-local
administrative audit event for bootstrap consumption.

CREATE TABLE auth_providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_key text NOT NULL UNIQUE,
    provider_type text NOT NULL CHECK (provider_type IN ('local','oidc','saml')),
    display_name text NOT NULL,
    config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_enabled boolean NOT NULL DEFAULT true
);

CREATE TABLE auth_identities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id uuid NOT NULL REFERENCES auth_providers(id),
    provider_subject text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    last_auth_at timestamptz,
    retired_at timestamptz,
    retired_by_user_id uuid REFERENCES users(id),
    retire_reason text,
    replaced_by_auth_binding_id uuid REFERENCES auth_identities(id),
    CHECK ((retired_at IS NULL) = (retired_by_user_id IS NULL))
);

CREATE UNIQUE INDEX auth_identities_active_provider_subject_uniq
    ON auth_identities (provider_id, provider_subject)
    WHERE retired_at IS NULL;

CREATE UNIQUE INDEX auth_identities_active_user_provider_uniq
    ON auth_identities (user_id, provider_id)
    WHERE retired_at IS NULL;

-- Non-normative example: the `auth_identities` table plus the partial unique
-- indexes below is one conformant realization of the enterprise-auth binding
-- persistence and active-binding uniqueness contract. Equivalent helper
-- columns, partial indexes, triggers, or storage-engine-specific constraints
-- are also conformant when they preserve the same public invariants.
-- Informative: `auth_providers.config_json` is the natural home for protocol
-- type and subject-mapping configuration sufficient to declare one stable
-- authoritative SAML subject source and browser-interactive provider behavior.
-- `auth_identities.provider_subject` is the authoritative external bind key.
-- Public `auth_binding_id` can be the safe stable exposure of
-- `auth_identities.id`. Rotation is realized as retire-plus-create in one
-- transaction, with the old row retained for lineage through
-- `replaced_by_auth_binding_id`. `last_auth_at` is updated only on a
-- successful provider-auth callback resolved through an active binding.
-- For `provider_type='local'`, this sketch relies on `users.email` and
-- `users.created_at` to materialize exactly one local safe-user binding
-- summary with `provider_key='local'`, derived `username` equal to
-- `users.email`, and derived `created_at` equal to `users.created_at`.
-- That local summary is not backed by an `auth_identities` row and does not
-- carry local `auth_binding_id`, `provider_subject`, or `last_auth_at`
-- state. Enterprise binding summaries arise only from active
-- provider-backed `auth_identities` rows. Omitting the derived local summary
-- from the safe user resource is non-conformant in the current profile. Any
-- persisted in-flight enterprise-auth transaction state remains
-- deployment-local ephemeral auth state and is excluded from incident
-- portability.

CREATE TABLE incidents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_key text NOT NULL,
    incident_key_canonical text NOT NULL,
    title text NOT NULL,
    description text,
    status text NOT NULL DEFAULT 'active',
    severity text,
    tlp text,
    current_phase text,
    primary_external_case_ref text,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid REFERENCES users(id),
    incident_version bigint NOT NULL DEFAULT 1,
    closed_at timestamptz,
    CHECK (incident_key_canonical <> ''),
    CHECK (octet_length(convert_to(incident_key_canonical, 'UTF8')) <= 128),
    CHECK (char_length(title) <= 512),
    CHECK (description IS NULL OR char_length(description) <= 16384),
    CHECK (severity IS NULL OR char_length(severity) <= 128),
    CHECK (tlp IS NULL OR char_length(tlp) <= 128),
    CHECK (current_phase IS NULL OR char_length(current_phase) <= 128),
    CHECK (primary_external_case_ref IS NULL OR char_length(primary_external_case_ref) <= 128),
    UNIQUE (incident_key_canonical)
);

-- Public `POST /api/v1/incidents` uses one committed create timestamp for
-- both `created_at` and `updated_at`, initializes `incident_version` to 1,
-- sets `status='active'`, leaves `closed_at NULL`, and binds both
-- `created_by_user_id` and `updated_by_user_id` to the creating local user.
--
-- Non-normative example: `incident_key_canonical` stores the trimmed,
-- Unicode NFC-normalized uniqueness form used by the public create
-- contract. A functional unique index over the same canonicalization rule
-- is equivalent. Full Unicode whitespace trimming, NFC normalization, and
-- control-character rejection remain service-layer validation rules rather
-- than pure SQL checks in this illustrative sketch.

CREATE TABLE incident_memberships (
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('viewer','editor','reviewer','admin')),
    joined_at timestamptz NOT NULL DEFAULT now(),
    added_by_user_id uuid NOT NULL REFERENCES users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid NOT NULL REFERENCES users(id),
    membership_version bigint NOT NULL DEFAULT 1,
    PRIMARY KEY (incident_id, user_id)
);

-- Informative: internal users and incident memberships are deployment-local
-- authorization state and are excluded from whole-incident portability
-- bundles.

CREATE TABLE reference_packs (
    pack_key text NOT NULL,
    version text NOT NULL,
    -- Illustrative local deployments MAY constrain pack_kind to a smaller subset.
    -- The public read contract keeps `pack_kind` as an open metadata string.
    pack_kind text NOT NULL,
    source_identifier text,
    manifest_sha256 text NOT NULL,
    -- Storage may retain multiple payload digests even though the public read
    -- contract exposes one canonical `payload_sha256` value.
    payload_sha256_list text[] NOT NULL DEFAULT '{}'::text[],
    pack_contract_version text,
    verification_method text,
    signer_key_id text,
    signature_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    trusted_source_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL DEFAULT 'staged' CHECK (
        status IN ('staged','available','disabled','failed','missing')
    ),
    imported_at timestamptz NOT NULL DEFAULT now(),
    imported_by_user_id uuid REFERENCES users(id),
    verification_result text NOT NULL DEFAULT 'pending' CHECK (
        verification_result IN ('pending','passed','failed')
    ),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (pack_key, version)
);

CREATE TABLE reference_pack_activation_state (
    pack_key text PRIMARY KEY,
    active_version text NOT NULL,
    previous_active_version text,
    activated_at timestamptz NOT NULL DEFAULT now(),
    activated_by_user_id uuid REFERENCES users(id),
    operator_note text,
    change_ticket text,
    CHECK (
        previous_active_version IS NULL
        OR previous_active_version <> active_version
    ),
    FOREIGN KEY (pack_key, active_version)
        REFERENCES reference_packs(pack_key, version),
    FOREIGN KEY (pack_key, previous_active_version)
        REFERENCES reference_packs(pack_key, version)
);

CREATE TABLE reference_pack_attestations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    -- Illustrative local deployments MAY constrain pack_kind to a smaller subset.
    -- The public read contract keeps `pack_kind` as an open metadata string.
    pack_kind text NOT NULL,
    event_kind text NOT NULL CHECK (
        event_kind IN ('import','activate')
    ),
    manifest_sha256 text NOT NULL,
    -- Storage may retain multiple payload digests even though the public read
    -- contract exposes one canonical `payload_sha256` value.
    payload_sha256_list text[] NOT NULL DEFAULT '{}'::text[],
    source_identifier text,
    verification_method text,
    signer_key_id text,
    trusted_source_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    previous_active_version text,
    verification_result text NOT NULL CHECK (
        verification_result IN ('passed','failed')
    ),
    actor_user_id uuid REFERENCES users(id),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    operator_note text,
    change_ticket text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    FOREIGN KEY (pack_key, pack_version)
        REFERENCES reference_packs(pack_key, version),
    FOREIGN KEY (pack_key, previous_active_version)
        REFERENCES reference_packs(pack_key, version)
);

CREATE TABLE type_registry_entries (
    registry_key text NOT NULL,
    type_key text NOT NULL,
    display_label text NOT NULL,
    category text,
    icon_key text,
    sort_order integer NOT NULL DEFAULT 0,
    pack_key text,
    pack_version text,
    is_local_override boolean NOT NULL DEFAULT false,
    config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (registry_key, type_key),
    FOREIGN KEY (pack_key, pack_version)
        REFERENCES reference_packs(pack_key, version)
);

CREATE TABLE view_schemas (
    id text PRIMARY KEY,
    sheet_type text NOT NULL,
    source_record_types text[] NOT NULL,
    base_projection text NOT NULL,
    field_registry_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    computed_columns jsonb NOT NULL DEFAULT '[]'::jsonb,
    required_reference_pack_keys text[] NOT NULL DEFAULT '{}'::text[],
    default_sort_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    sort_fields_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    grouping_fields_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    filter_contract jsonb NOT NULL DEFAULT '{}'::jsonb,
    writeback_contract jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    record_type text NOT NULL CHECK (
        record_type IN ('timeline_event','host','identity','party','indicator','artifact','evidence','assessment')
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid NOT NULL REFERENCES users(id),
    row_version bigint NOT NULL DEFAULT 1,
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users(id)
);

CREATE INDEX idx_records_incident_type
    ON records (incident_id, record_type)
    WHERE deleted_at IS NULL;
```

```sql
CREATE TABLE timeline_events (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_seq bigint GENERATED ALWAYS AS IDENTITY,
    occurred_at timestamptz,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    summary text,
    details text,
    source_text text,
    capture_state text NOT NULL DEFAULT 'rough' CHECK (
        capture_state IN ('rough','enriched','reviewed','superseded')
    ),
    confidence smallint CHECK (confidence BETWEEN 0 AND 100),
    raw_capture jsonb NOT NULL DEFAULT '{}'::jsonb,
    custom_attrs jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_timeline_events_incident_occurred
    ON timeline_events (incident_id, occurred_at NULLS LAST, event_seq);
CREATE INDEX idx_timeline_events_incident_recorded
    ON timeline_events (incident_id, recorded_at DESC);

CREATE TABLE hosts (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    display_name text NOT NULL CHECK (char_length(display_name) <= 256),
    hostname citext,
    fqdn citext,
    asset_id text,
    host_type_key text NOT NULL DEFAULT 'host',
    aad_device_id uuid,
    os_platform text,
    host_state text NOT NULL DEFAULT 'stub' CHECK (
        host_state IN ('stub','canonical','merged','retired')
    ),
    merged_into_record_id uuid REFERENCES records(id),
    custom_attrs jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX uq_hosts_incident_hostname
    ON hosts (incident_id, lower(hostname::text))
    WHERE hostname IS NOT NULL AND merged_into_record_id IS NULL;

CREATE TABLE identities (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    display_name text NOT NULL CHECK (char_length(display_name) <= 256),
    upn citext,
    sam_account_name citext,
    email citext,
    aad_object_id uuid,
    sid text,
    identity_type text NOT NULL DEFAULT 'user' CHECK (
        identity_type IN ('user','service','shared','external')
    ),
    identity_state text NOT NULL DEFAULT 'stub' CHECK (
        identity_state IN ('stub','canonical','merged','disabled')
    ),
    merged_into_record_id uuid REFERENCES records(id),
    custom_attrs jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX uq_identities_incident_upn
    ON identities (incident_id, lower(upn::text))
    WHERE upn IS NOT NULL AND merged_into_record_id IS NULL;

CREATE TABLE parties (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    display_name text NOT NULL CHECK (char_length(display_name) <= 256),
    party_kind text NOT NULL CHECK (
        party_kind IN ('person','team','organization','distribution_list','other')
    ),
    organization_name text CHECK (organization_name IS NULL OR char_length(organization_name) <= 256),
    role_title text CHECK (role_title IS NULL OR char_length(role_title) <= 256),
    primary_email citext CHECK (primary_email IS NULL OR char_length(primary_email::text) <= 320),
    timezone_name text CHECK (timezone_name IS NULL OR char_length(timezone_name) <= 128),
    external_ref text CHECK (external_ref IS NULL OR char_length(external_ref) <= 1024),
    notes text CHECK (notes IS NULL OR char_length(notes) <= 16384),
    custom_attrs jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_parties_incident_primary_email
    ON parties (incident_id, lower(primary_email::text))
    WHERE primary_email IS NOT NULL;

CREATE INDEX idx_parties_incident_external_ref
    ON parties (incident_id, external_ref)
    WHERE external_ref IS NOT NULL;

-- Exact-match reuse for active same-incident parties, same-incident ref validation,
-- and delete rejection while referenced remain authoritative application-layer
-- rules because they depend on record-envelope delete state and inbound refs.

CREATE TABLE indicators (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    indicator_type text NOT NULL,
    value_kind text NOT NULL CHECK (
        value_kind IN ('atomic','pattern','reference')
    ),
    display_value text NOT NULL,
    normalized_value text,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    custom_attrs jsonb NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (incident_id, indicator_type, dedupe_key)
);

CREATE INDEX idx_indicators_lookup
    ON indicators (incident_id, indicator_type, dedupe_key);

CREATE INDEX idx_indicators_normalized
    ON indicators (incident_id, indicator_type, normalized_value)
    WHERE normalized_value IS NOT NULL;

CREATE TABLE entity_aliases (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    alias_type text NOT NULL,
    alias_text citext NOT NULL CHECK (char_length(alias_text::text) <= 256),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (record_id, alias_text)
);

CREATE INDEX idx_entity_aliases_incident_alias_trgm
    ON entity_aliases USING gin (lower(alias_text::text) gin_trgm_ops);
```

One conformant realization can keep ordinary `suggestion_only` aliases and host or identity `exact_match_reuse` carry-forward values in one child store by adding a server-owned internal classification field or equivalent companion rows that are not surfaced through the ordinary alias editor. Another conformant realization can keep `entity_aliases` for `suggestion_only` values only and place active reusable identifiers in a separate child table keyed by same-incident `record_id`, identifier class, normalized value, and active-state boundary. In either realization, the normative core still requires incident-scoped uniqueness checks for active host and identity `exact_match_reuse` values and fail-closed third-record collision detection during merge.

```sql
CREATE TABLE artifacts (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    artifact_type text NOT NULL CHECK (
        artifact_type IN ('note','query','excerpt','export','other')
    ),
    title text CHECK (title IS NULL OR char_length(title) <= 512),
    content_text text CHECK (content_text IS NULL OR char_length(content_text) <= 16384),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    content_tsv tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title,'')), 'A') ||
        setweight(to_tsvector('simple', coalesce(content_text,'')), 'B')
    ) STORED
);

CREATE INDEX idx_artifacts_search
    ON artifacts USING gin (content_tsv);
```

```sql
CREATE TABLE object_blobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    create_client_txn_id text NOT NULL,
    storage_backend text NOT NULL CHECK (storage_backend IN ('s3','filesystem')),
    bucket_name text NOT NULL,
    object_key text NOT NULL UNIQUE,
    filename_hint text,
    content_type_hint text,
    declared_size_bytes bigint NOT NULL CHECK (declared_size_bytes >= 0),
    expected_sha256 text,
    observed_content_type text,
    observed_size_bytes bigint CHECK (observed_size_bytes IS NULL OR observed_size_bytes >= 0),
    observed_sha256 text,
    upload_state text NOT NULL DEFAULT 'pending' CHECK (
        upload_state IN ('pending','available','failed','quarantined')
    ),
    target_expires_at timestamptz,
    pending_expires_at timestamptz,
    finalize_attempt_count integer NOT NULL DEFAULT 0 CHECK (finalize_attempt_count >= 0),
    terminal_reason text,
    failed_at timestamptz,
    cleanup_due_at timestamptz,
    cleaned_up_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    UNIQUE (incident_id, created_by_user_id, create_client_txn_id)
);

CREATE TABLE evidence_records (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    object_blob_id uuid REFERENCES object_blobs(id),
    title text CHECK (title IS NULL OR char_length(title) <= 512),
    evidence_type_key text NOT NULL DEFAULT 'other',
    description text,
    lifecycle_state text NOT NULL DEFAULT 'requested' CHECK (
        lifecycle_state IN (
            'requested','pending_receipt','received','available','quarantined','released'
        )
    ),
    owner_user_id uuid REFERENCES users(id),
    collector_party_id uuid REFERENCES records(id),
    source_party_id uuid REFERENCES records(id),
    requested_at timestamptz NOT NULL DEFAULT now(),
    received_at timestamptz,
    available_at timestamptz,
    released_at timestamptz,
    storage_ref text CHECK (storage_ref IS NULL OR char_length(storage_ref) <= 1024),
    collector_party_text text CHECK (collector_party_text IS NULL OR char_length(collector_party_text) <= 256),
    source_party_text text CHECK (source_party_text IS NULL OR char_length(source_party_text) <= 256),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

-- Same-incident validation for `collector_party_id` and `source_party_id`, plus
-- delete rejection while those refs remain active, are application-layer rules in
-- this sketch because they depend on record-envelope incident ownership and
-- active-state checks beyond a simple foreign key.

CREATE INDEX idx_evidence_incident_type
    ON evidence_records (incident_id, evidence_type_key);
-- The full task-request and decision table sketches are omitted from this
-- extract. When materialized, their title, body, locator, party, and reason
-- fields should apply the same illustrative size caps aligned to
-- `single_line_title_v1`, `multiline_body_v1`, `locator_text_v1`,
-- `party_text_v1`, and `reason_note_v1`. Task requests should pair
-- `requester_party_text` with optional `requester_party_id` rather than
-- treating requester identity as free text or `user_id`.

CREATE TABLE evidence_custody_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    evidence_record_id uuid NOT NULL REFERENCES evidence_records(record_id) ON DELETE CASCADE,
    custody_event_type text NOT NULL CHECK (
        custody_event_type IN (
            'requested','received','made_available','transferred','quarantined','released'
        )
    ),
    actor_user_id uuid REFERENCES users(id),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    location_text text,
    note text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_evidence_custody_events_record_time
    ON evidence_custody_events (evidence_record_id, occurred_at DESC);

CREATE TABLE compromise_assessments (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    subject_record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    subject_type text NOT NULL CHECK (subject_type IN ('host','identity')),
    assessment_state text NOT NULL CHECK (
        assessment_state IN ('unknown','suspected','confirmed','disproven','cleared')
    ),
    confidence_score smallint CHECK (confidence_score BETWEEN 0 AND 100),
    rationale text,
    supporting_record_id uuid REFERENCES records(id),
    assessed_at timestamptz NOT NULL DEFAULT now(),
    supersedes_record_id uuid REFERENCES records(id),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_compromise_assessments_subject
    ON compromise_assessments (incident_id, subject_record_id, assessed_at DESC);

CREATE TABLE indicator_state_intervals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    indicator_record_id uuid NOT NULL REFERENCES indicators(record_id) ON DELETE CASCADE,
    state_key text NOT NULL,
    valid_from timestamptz,
    valid_to timestamptz,
    confidence smallint CHECK (confidence BETWEEN 0 AND 100),
    rationale text,
    supporting_record_id uuid REFERENCES records(id),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    supersedes_interval_id uuid REFERENCES indicator_state_intervals(id),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_indicator_state_intervals_indicator
    ON indicator_state_intervals (incident_id, indicator_record_id, created_at DESC);

CREATE TABLE indicator_observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    source_record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    source_field_key text NOT NULL,
    origin_kind text NOT NULL CHECK (
        origin_kind IN ('manual_entry','clipboard_paste','csv_import','xlsx_import','api_import','extraction','system')
    ),
    origin_locator text NOT NULL,
    observed_text text NOT NULL,
    parsed_indicator_type text,
    normalized_candidate text,
    span_locator text,
    resolution_status text NOT NULL DEFAULT 'unresolved' CHECK (
        resolution_status IN ('unresolved','resolved','dismissed')
    ),
    resolved_indicator_record_id uuid REFERENCES indicators(record_id),
    extraction_method text,
    extraction_confidence smallint CHECK (extraction_confidence BETWEEN 0 AND 100),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    updated_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    row_version bigint NOT NULL DEFAULT 1,
    deleted_at timestamptz
);

CREATE INDEX idx_indicator_observations_source
    ON indicator_observations (source_record_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_indicator_observations_resolved
    ON indicator_observations (incident_id, resolution_status, resolved_indicator_record_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_indicator_observations_normalized
    ON indicator_observations (incident_id, parsed_indicator_type, normalized_candidate)
    WHERE deleted_at IS NULL AND normalized_candidate IS NOT NULL;

CREATE TABLE entity_mentions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    source_record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    entity_type text NOT NULL CHECK (entity_type IN ('host','identity','artifact')),
    raw_text text NOT NULL,
    normalized_text text GENERATED ALWAYS AS (
        lower(regexp_replace(raw_text, '\s+', ' ', 'g'))
    ) STORED,
    ordinal int NOT NULL DEFAULT 0,
    resolution_status text NOT NULL DEFAULT 'unresolved' CHECK (
        resolution_status IN ('unresolved','resolved','dismissed')
    ),
    resolved_record_id uuid REFERENCES records(id),
    confidence smallint CHECK (confidence BETWEEN 0 AND 100),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    updated_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    row_version bigint NOT NULL DEFAULT 1,
    deleted_at timestamptz
);

CREATE INDEX idx_mentions_source
    ON entity_mentions (source_record_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_mentions_resolved
    ON entity_mentions (incident_id, entity_type, resolution_status, resolved_record_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_mentions_raw_trgm
    ON entity_mentions USING gin (normalized_text gin_trgm_ops)
    WHERE deleted_at IS NULL;
```

```sql
-- Persisted record links are directed. Reverse traversal is derived from query or projection state.
CREATE TABLE record_links (
    record_link_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    src_record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    dst_record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    link_type text NOT NULL CHECK (
        link_type IN (
            'observed_on_host',
            'observed_as_identity',
            'references_indicator',
            'attached_evidence',
            'references_artifact',
            'derived_from',
            'merged_into',
            'supported_by',
            'references_record',
            'supersedes'
        )
    ),
    provenance text NOT NULL DEFAULT 'manual' CHECK (
        provenance IN ('manual','auto_match','import','rollback','system')
    ),
    confidence smallint CHECK (confidence BETWEEN 0 AND 100),
    note text,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users(id),
    CHECK (src_record_id <> dst_record_id)
);

CREATE UNIQUE INDEX uq_active_links
    ON record_links (incident_id, src_record_id, dst_record_id, link_type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_links_src
    ON record_links (incident_id, src_record_id, link_type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_links_dst
    ON record_links (incident_id, dst_record_id, link_type)
    WHERE deleted_at IS NULL;

-- Enforce the same-incident endpoint invariant with composite foreign keys or an equivalent constraint or trigger.

CREATE TABLE tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    name citext NOT NULL CHECK (char_length(name::text) <= 64),
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (incident_id, name)
);

CREATE TABLE record_tags (
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    tag_id uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (record_id, tag_id)
);

CREATE TABLE change_sets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    actor_user_id uuid NOT NULL REFERENCES users(id),
    reason text CHECK (reason IS NULL OR char_length(reason) <= 4096),
    source text NOT NULL DEFAULT 'ui',
    client_txn_id uuid,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE change_set_mutations (
    id bigserial PRIMARY KEY,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    change_set_id uuid NOT NULL REFERENCES change_sets(id) ON DELETE CASCADE,
    sequence_no integer NOT NULL,
    target_kind text NOT NULL CHECK (
        target_kind IN (
            'record_field',
            'record_link',
            'record_tag',
            'entity_mention',
            'evidence_association',
            'merge_repoint'
        )
    ),
    target_id text NOT NULL,
    parent_record_id uuid REFERENCES records(id) ON DELETE CASCADE,
    field_key text,
    operation text NOT NULL CHECK (
        operation IN (
            'insert','update','soft_delete','restore','resolve','dismiss',
            'attach','detach','repoint','rollback'
        )
    ),
    before_version_id text,
    after_version_id text,
    before_json jsonb,
    after_json jsonb,
    patch_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (change_set_id, sequence_no)
);

CREATE INDEX idx_change_set_mutations_parent_record
    ON change_set_mutations (parent_record_id, change_set_id, sequence_no);

CREATE INDEX idx_change_set_mutations_change_set
    ON change_set_mutations (change_set_id, sequence_no);

CREATE TABLE record_revisions (
    id bigserial PRIMARY KEY,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    record_id uuid NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    record_type text NOT NULL,
    change_set_id uuid NOT NULL REFERENCES change_sets(id) ON DELETE CASCADE,
    revision_no bigint NOT NULL,
    operation text NOT NULL CHECK (
        operation IN ('insert','update','soft_delete','restore','rollback')
    ),
    before_json jsonb,
    after_json jsonb,
    patch_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    changed_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (record_id, revision_no)
);

CREATE INDEX idx_record_revisions_record
    ON record_revisions (record_id, revision_no DESC);

CREATE INDEX idx_record_revisions_incident_time
    ON record_revisions (incident_id, changed_at DESC);

CREATE TABLE saved_views (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    owner_user_id uuid REFERENCES users(id),
    view_scope text NOT NULL CHECK (view_scope IN ('private','shared','system')),
    view_schema_id text NOT NULL REFERENCES view_schemas(id),
    display_name text NOT NULL CHECK (char_length(display_name) <= 256),
    query_json jsonb NOT NULL,
    layout_json jsonb NOT NULL,
    saved_view_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        (view_scope = 'system' AND owner_user_id IS NULL)
        OR (view_scope IN ('private','shared') AND owner_user_id IS NOT NULL)
    )
);

CREATE TABLE user_workbook_preferences (
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    home_sheet_ref jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (incident_id, user_id)
);

CREATE TABLE incident_workbook_preferences (
    incident_id uuid PRIMARY KEY REFERENCES incidents(id) ON DELETE CASCADE,
    default_sheet_ref jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid REFERENCES users(id)
);

-- Informative: public `POST /api/v1/incidents/{incident_id}/saved-views`
-- requires non-null `display_name` and `query_json`, defaults omitted
-- `scope` to `private`, normalizes omitted or `{}` `layout_json` to the
-- canonical schema-derived `cartulary.layout.v1` object, and keeps
-- startup/default surface selection in the separate workbook-preference
-- objects above rather than on saved-view rows.
--
-- Informative: public `PUT /api/v1/incidents/{incident_id}/workbook-
-- preferences/me` accepts only `{ "home_sheet_ref": <sheet_ref|null> }`,
-- and the `default` route accepts only
-- `{ "default_sheet_ref": <sheet_ref|null> }`. Both routes create the
-- preference object if absent, replace only the named pointer if present,
-- and leave `updated_at` unchanged on no-op updates.
--
-- Informative: for required base coordination surfaces
-- `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`,
-- `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`, any
-- saved-view-shaped helper row used for defaults, layout, or query state is
-- implementation detail or a distinct saved-view object. It is not the
-- authoritative public identity of the required base surface.
```

### Illustrative public `view_schema_resource_v1` example

```json
{
  "view_schema_id": "cartulary.view.evidence.v1",
  "surface_kind": "built_in_sheet",
  "title": "Evidence",
  "source_record_types": ["evidence"],
  "technical_fields": ["record_id", "row_version"],
  "required_reference_pack_keys": [],
  "default_sort": [
    { "field_key": "evidence.requested_at", "direction": "desc" },
    { "field_key": "record_id", "direction": "asc" }
  ],
  "sort_fields": [
    "evidence.blob_hash",
    "evidence.collector_party_text",
    "evidence.edited_at",
    "evidence.lifecycle_state",
    "evidence.linked_record_count",
    "evidence.received_at",
    "evidence.requested_at",
    "evidence.source_party_text",
    "evidence.storage_ref",
    "evidence.title",
    "evidence.upload_state"
  ],
  "filter_fields": [
    "evidence.blob_hash",
    "evidence.collector_party_text",
    "evidence.lifecycle_state",
    "evidence.received_at",
    "evidence.requested_at",
    "evidence.source_party_text",
    "evidence.storage_ref",
    "evidence.upload_state"
  ],
  "synthetic_filter_predicates": [],
  "grouping_fields": [],
  "fields": [
    {
      "field_key": "evidence.title",
      "label": "Title",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "prefix"],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "direct_value",
      "conflict_resolution_class": "text_compare_merge",
      "entity_binding_mode": null,
      "string_contract_id": "single_line_title_v1",
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    },
    {
      "field_key": "evidence.lifecycle_state",
      "label": "Lifecycle State",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq"],
      "groupable": false,
      "read_kind": "enum",
      "write_kind": "direct_value",
      "conflict_resolution_class": "atomic_replace",
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": ["requested", "pending_receipt", "received", "available", "quarantined", "released"]
    },
    {
      "field_key": "evidence.requested_at",
      "label": "Requested At",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "range"],
      "groupable": false,
      "read_kind": "timestamp",
      "write_kind": "direct_value",
      "conflict_resolution_class": "atomic_replace",
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": "timestamp_instant_v1",
      "direct_reference_contract_id": null,
      "clearable": true,
      "enum_values": null
    },
    {
      "field_key": "evidence.received_at",
      "label": "Received At",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "range"],
      "groupable": false,
      "read_kind": "timestamp",
      "write_kind": "direct_value",
      "conflict_resolution_class": "atomic_replace",
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": "timestamp_instant_v1",
      "direct_reference_contract_id": null,
      "clearable": true,
      "enum_values": null
    },
    {
      "field_key": "evidence.storage_ref",
      "label": "Storage Ref",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "prefix"],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "direct_value",
      "conflict_resolution_class": "atomic_replace",
      "entity_binding_mode": null,
      "string_contract_id": "locator_text_v1",
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    },
    {
      "field_key": "evidence.blob_hash",
      "label": "Blob Hash",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "prefix"],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "read_only",
      "conflict_resolution_class": null,
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    },
    {
      "field_key": "evidence.collector_party_text",
      "label": "Collector",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "prefix"],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "direct_value",
      "conflict_resolution_class": "text_compare_merge",
      "entity_binding_mode": null,
      "string_contract_id": "party_text_v1",
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    },
    {
      "field_key": "evidence.collector_party_id",
      "label": "Collector Party",
      "default_hidden": true,
      "sortable": false,
      "header_sort_field_key": null,
      "filter_ops": [],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "direct_value",
      "conflict_resolution_class": "atomic_replace",
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": "same_incident_party_ref_v1",
      "clearable": true,
      "enum_values": null
    },
    {
      "field_key": "evidence.source_party_text",
      "label": "Source",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq", "prefix"],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "direct_value",
      "conflict_resolution_class": "text_compare_merge",
      "entity_binding_mode": null,
      "string_contract_id": "party_text_v1",
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    },
    {
      "field_key": "evidence.source_party_id",
      "label": "Source Party",
      "default_hidden": true,
      "sortable": false,
      "header_sort_field_key": null,
      "filter_ops": [],
      "groupable": false,
      "read_kind": "text",
      "write_kind": "direct_value",
      "conflict_resolution_class": "atomic_replace",
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": "same_incident_party_ref_v1",
      "clearable": true,
      "enum_values": null
    },
    {
      "field_key": "evidence.upload_state",
      "label": "Upload State",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": ["eq"],
      "groupable": false,
      "read_kind": "enum",
      "write_kind": "read_only",
      "conflict_resolution_class": null,
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": ["pending", "available", "failed", "quarantined"]
    },
    {
      "field_key": "evidence.linked_record_count",
      "label": "Linked Record Count",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": [],
      "groupable": false,
      "read_kind": "number",
      "write_kind": "read_only",
      "conflict_resolution_class": null,
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    },
    {
      "field_key": "evidence.edited_at",
      "label": "Edited At",
      "default_hidden": false,
      "sortable": true,
      "header_sort_field_key": null,
      "filter_ops": [],
      "groupable": false,
      "read_kind": "timestamp",
      "write_kind": "read_only",
      "conflict_resolution_class": null,
      "entity_binding_mode": null,
      "string_contract_id": null,
      "direct_scalar_contract_id": null,
      "direct_reference_contract_id": null,
      "clearable": false,
      "enum_values": null
    }
  ]
}
```

### Illustrative public `view_field_entry_v1` example

```json
{
  "field_key": "evidence.collector_party_id",
  "label": "Collector Party",
  "default_hidden": true,
  "sortable": false,
  "header_sort_field_key": null,
  "filter_ops": [],
  "groupable": false,
  "read_kind": "text",
  "write_kind": "direct_value",
  "conflict_resolution_class": "atomic_replace",
  "entity_binding_mode": null,
  "string_contract_id": null,
  "direct_scalar_contract_id": null,
  "direct_reference_contract_id": "same_incident_party_ref_v1",
  "clearable": true,
  "enum_values": null
}
```

### Illustrative canonical default `layout_json` example

```json
{
  "layout_schema_id": "cartulary.layout.v1",
  "column_order": [
    "evidence.title",
    "evidence.lifecycle_state",
    "evidence.requested_at",
    "evidence.received_at",
    "evidence.storage_ref",
    "evidence.blob_hash",
    "evidence.collector_party_text",
    "evidence.collector_party_id",
    "evidence.source_party_text",
    "evidence.source_party_id",
    "evidence.upload_state",
    "evidence.linked_record_count",
    "evidence.edited_at"
  ],
  "hidden_field_keys": [
    "evidence.collector_party_id",
    "evidence.source_party_id"
  ],
  "column_widths": []
}
```

### Illustrative non-default `layout_json` example

```json
{
  "layout_schema_id": "cartulary.layout.v1",
  "column_order": [
    "evidence.title",
    "evidence.requested_at",
    "evidence.received_at",
    "evidence.lifecycle_state",
    "evidence.storage_ref",
    "evidence.collector_party_text",
    "evidence.source_party_text",
    "evidence.blob_hash",
    "evidence.collector_party_id",
    "evidence.source_party_id",
    "evidence.upload_state",
    "evidence.linked_record_count",
    "evidence.edited_at"
  ],
  "hidden_field_keys": [
    "evidence.blob_hash",
    "evidence.collector_party_id",
    "evidence.source_party_id"
  ],
  "column_widths": [
    { "field_key": "evidence.collector_party_text", "width_px": 260 },
    { "field_key": "evidence.source_party_text", "width_px": 260 },
    { "field_key": "evidence.title", "width_px": 420 }
  ]
}
```

### Non-normative migration note for saved-view normalization

Legacy omitted `filters` normalize to `filters=[]`. Legacy omitted or `{}` `layout_json` normalizes to the canonical schema-derived `cartulary.layout.v1` object for the owning `view_schema_id`. Ambiguous legacy saved views are better left unrepaired at runtime. If an incident-portability implementation imports a saved view that cannot be normalized against the resolved `view_schema_id`, a defensible current-profile choice is to skip that saved-view object, import the core incident state, and emit a deterministic diagnostic rather than fail the entire bundle import. The underlying rationale for the closed shared-layout grammar is to preserve portable workbook semantics while keeping per-session UI state client-local.

```sql
CREATE TABLE timeline_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_seq bigint NOT NULL,
    occurred_at timestamptz,
    recorded_at timestamptz NOT NULL,
    sort_ts timestamptz NOT NULL,
    summary text,
    details_excerpt text,
    capture_state text NOT NULL,
    host_labels text[] NOT NULL DEFAULT '{}'::text[],
    unresolved_host_tokens text[] NOT NULL DEFAULT '{}'::text[],
    identity_labels text[] NOT NULL DEFAULT '{}'::text[],
    unresolved_identity_tokens text[] NOT NULL DEFAULT '{}'::text[],
    artifact_labels text[] NOT NULL DEFAULT '{}'::text[],
    evidence_count integer NOT NULL DEFAULT 0,
    tag_names text[] NOT NULL DEFAULT '{}'::text[],
    author_display text,
    last_editor_display text,
    row_version bigint NOT NULL,
    search_tsv tsvector
);

CREATE INDEX idx_timeline_grid_sort
    ON timeline_grid_projection (incident_id, sort_ts DESC, event_seq DESC);

CREATE INDEX idx_timeline_grid_state
    ON timeline_grid_projection (incident_id, capture_state);

CREATE INDEX idx_timeline_grid_search
    ON timeline_grid_projection USING gin (search_tsv);

CREATE INDEX idx_timeline_grid_tags_gin
    ON timeline_grid_projection USING gin (tag_names);

CREATE TABLE indicator_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES records(id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    indicator_type text NOT NULL,
    value_kind text NOT NULL,
    indicator_value text NOT NULL,
    normalized_value text,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    first_observed_at timestamptz,
    last_observed_at timestamptz,
    current_state_key text,
    current_valid_from timestamptz,
    current_valid_to timestamptz,
    observation_count integer NOT NULL DEFAULT 0,
    linked_event_count integer NOT NULL DEFAULT 0,
    linked_evidence_count integer NOT NULL DEFAULT 0,
    linked_host_count integer NOT NULL DEFAULT 0,
    linked_identity_count integer NOT NULL DEFAULT 0,
    row_version bigint NOT NULL,
    search_tsv tsvector
);

CREATE INDEX idx_indicator_grid_lookup
    ON indicator_grid_projection (incident_id, indicator_type, dedupe_key);

CREATE INDEX idx_indicator_grid_normalized
    ON indicator_grid_projection (incident_id, indicator_type, normalized_value)
    WHERE normalized_value IS NOT NULL;

CREATE INDEX idx_indicator_grid_search
    ON indicator_grid_projection USING gin (search_tsv);
```

### Indexing strategy

- **Filtering / sorting**: b-tree on `(incident_id, occurred_at)`, `(incident_id, recorded_at)`, `(incident_id, capture_state)`, `(incident_id, record_type)`.
- **Full-text search**: GIN on `tsvector` using the **`simple`** text search config, not English, because IR tokens like hostnames, hashes, UPNs, and account names do not stem well.
- **Lookup by linked entities**: composite indexes on `record_links` source and destination columns.
- **Canonical indicator lookup**: composite indexes on `(incident_id, indicator_type, dedupe_key)` and `(incident_id, indicator_type, normalized_value)` for exact and normalized indicator matching.
- **Observation traversal**: indexes on `indicator_observations` by source record, resolution state, and normalized candidate.
- **Fuzzy matching**: `pg_trgm` on alias text and unresolved mention text.
- **Array containment** on projection tables: GIN for tags and denormalized label arrays when useful.

### Soft delete vs hard delete

- User-facing records are **soft-deleted** via `records.deleted_at`.
- Links are soft-deletable.
- Revisions are append-only and never deleted in normal operation.
- Blobs can be **hard-deleted** only when their owning evidence record is purged by an explicit admin workflow or retention policy.

### Rollback, lineage, revision history

Every mutation creates a `change_set`. In one conformant realization, storage is authoritative at `change_set` plus mutation-target granularity, not at row-snapshot granularity alone. `record_revisions` remain useful row-centric snapshot history for audit and fast restore, but they are not the sole rollback substrate.

#### Storage history granularity

For each committed action, one conformant history realization creates one immutable `change_set` and one or more reversible mutation entries. Each mutation entry is queryable together with the parent change set's actor, timestamp, source, and reason, and it records:

- target kind and stable target id,
- operation kind,
- deterministic ordering within the `change_set`,
- pre-change and post-change version identifiers,
- either `before_value` / `after_value` or an equivalent reversible patch.

Illustrative mutation targets include, at minimum:

- scalar/document fields on first-class records,
- `record_links`,
- `record_tags`,
- `entity_mentions` lifecycle changes, including resolve, dismiss, and restore,
- `indicator_observations` lifecycle changes, including resolve, dismiss, and restore,
- `indicator_state_intervals` append or supersession changes,
- evidence association changes, including evidence-record linkage,
- merge/repoint fan-out caused by entity merge or restore.

For `entity_mentions`, ordinary dismiss and restore semantics are lifecycle transitions on the existing mention row, not delete-and-recreate operations. In this realization, a dismiss transition preserves `raw_text`, derived `normalized_text`, stable mention identity, and provenance, clears active resolution metadata such as `resolved_record_id`, `resolved_by_user_id`, `resolved_at`, and `resolution_method`, and removes or tombstones any corresponding active resolved `record_link` in the same `change_set`. An ordinary restore transition returns that same mention row to `unresolved` with resolution metadata null; recovering a prior resolved target is handled by rollback of the dismissal change, not by ordinary restore.

The history model in this realization preserves enough detail to reconstruct both the full row snapshot at any revision and the exact field/link/mention/tag/evidence delta introduced by a `change_set`. Row snapshots such as `before_json` / `after_json` can be retained for audit and fast restore, but they are not the only rollback substrate. Projection tables are not authoritative history; they remain derived state only.

This yields a deliberate split: attribution unit = `change_set`; rollback unit = mutation entry or whole `change_set`; primary reviewer lens = row.

For entity merges specifically, active mention resolutions and active links in this realization never continue to point at a merged-away record. The merge change set repoints those live references to the survivor, preserves the losing record for audit/history, and emits enough revision detail to reconstruct both the pre-merge graph and the post-merge graph.

### Postgres-native features worth using

- **JSONB** for low-frequency enrichment and view config.
- **GIN + `pg_trgm`** for alias and mention matching.
- **`citext`** for case-insensitive identifiers.
- **`tsvector`** for search over timeline and artifact content.
- **TOASTed JSONB snapshots** in `record_revisions` for fast row restore and audit diffs.
