# Appendix E: Roadmap, Open Questions, and Decision Backlog

This appendix is **non-normative**.

It preserves roadmap and open-question material from the exploratory source artifact and adds an editorial decision backlog used to keep the provisional normative core internally consistent.

## E.1 Editorial decision backlog

The source artifact mixed current requirements, roadmap positioning, and open questions in a few places. The provisional normative core resolves those tensions with bounded current-profile rules and extension profiles. This backlog records the remaining editorial pressure points and any recently closed decisions that materially affect roadmap interpretation.

| Decision ID | Topic | Current normative treatment | Current status or follow-on scope |
| --- | --- | --- | --- |
| `DQ-002` | **File-based structured import** closure | The normative core now closes this by defining a Phase 2 Workbook Import Assistant inside the **Import Extension Profile**. Clipboard paste remains the base hot-path surface. File-based workbook import remains isolated behind the dedicated imports module and stable tabular-ingest contract. Whole-workbook import is now defined as an orchestrated batch of explicit `import_unit` objects inside one `import_session`, with deterministic `mapping_fingerprint` values, a closed `warning_code[]` vocabulary, explicit used-range, table, named-range, and region discovery, no auto-resolution during ingest, and no execution of formulas, VBA, automation, or external links. | Resolved in the current normative core. Future work MAY add richer parser coverage or operator conveniences without widening the base workbook surface or leaking XLSX semantics outside `imports`. |
| `DQ-003` | **Immutable snapshots and report exports** status | The normative core places snapshot and report generation in the **Snapshot and Reporting Extension Profile** and fixes the immutable release tuple, release scopes, template contracts, redaction profiles, and artifact release records inside that profile. | The source artifact treats snapshot-based reporting as architecturally important while also placing productized export delivery in later phases. |
| `DQ-004` | **Reference-pack refresh and distribution** status | The normative core places refresh and integrity-checked distribution in the **Reference Pack Extension Profile** and now fixes the smallest disconnected bundle, offline import flow, explicit activation, retained prior active version, and structured attestation metadata. | Future work may still add richer mirroring, signing, or distribution tooling without changing the current disconnected-pack contract. |
| `DQ-005` | **Incident portability contract** | The normative core now places whole-incident export/import in the **Incident Portability Extension Profile** and fixes the logical bundle layout, manifest, deterministic JSON/NDJSON serialization, portable actor descriptors, import phases, and fail-closed bundle verification rules. | Future work may still add richer migration tooling, alternative transport wrappers, or operator conveniences without changing the current portability contract. |
| `DQ-006` | **Restricted evidence visibility versus export-scoped withholding** | The normative core keeps live workspace visibility incident-scoped, routes recipient-specific withholding through snapshot, render, and release controls, and leaves narrow live sensitive-evidence controls to future work. | The source artifact mixed restricted-evidence questions with redaction and reporting controls; the current core separates export disclosure from live workspace authorization while preserving a future path if repeated practice proves it necessary. |
| `DQ-007` | **Saved-view scope and startup/default selection** | The normative core now closes saved-view scope as exactly `private`, `shared`, and `system`, adds owner/admin mutability rules and ordinary delete semantics, separates saved-view discoverability from underlying incident-data authorization, and defines separate per-user and incident-wide `sheet_ref` preference objects for workbook startup selection. | Resolved in the current normative core. Future work MAY add richer saved-view curation, seeding, or suggestion tooling without changing the base scope/default contract. |
| `DQ-008` | **Generated presentation depth** | The normative core now fixes `content_class`, `release_scope`, `support_refs[]`, versioned template and redaction controls, and the current generated-presentation boundary, including `internal_review`-only reenactments marked `generated_presentation=true`. | Future work may still add richer authoring or visualization families without relaxing the current evidence-versus-presentation boundary. |
| `DQ-009` | **`task_request` and `decision` legal transition matrices** | The normative core now defines explicit lifecycle machines for `task_request.status` and `decision.status`, including legal transitions, reciprocal guard behavior, direct-write rejection of `decision.status='superseded'`, and stable illegal-transition failure semantics. | Resolved in the current normative core. Future work MAY add richer workflow presentation or operator conveniences without changing these base lifecycle semantics or introducing a generalized workflow engine. |
| `DQ-010` | **Public interface surface contract** | The normative core now closes the browser-facing wire contract as a versioned HTTP+JSON API rooted at `/api/v1/` plus a bounded WebSocket stream rooted at `/ws/v1/`, with session-based auth, stable route families, field-key-based mutation, explicit record-scoped soft-delete and restore routes, history-exposed tombstone version tokens for restore concurrency, generic success and error envelopes, cursor pagination, a closed view-query `filters[]` wire contract keyed by stable `field_key` values with operator-specific operands, stable synthetic filter keys for filter-only predicates such as full-text search, blob-slot and job routes, and a concrete v1 collaboration wire contract covering the common message envelope, typed `hello` and `resume` handshakes, typed presence payloads, replayable-versus-ephemeral delivery classes, bounded resume and reset behavior, application-level heartbeats, per-incident ordering, and lifecycle `record_changed` semantics that surface delete as `remove` and restore as `invalidate`. | Resolved in the current normative core. Future work MAY add route families or optional members within the same major version, but breaking changes to required route fields, required filter operators or operand shapes, required message members, required message types, or message semantics require a new major version root. |
| `DQ-011` | **Pending blob timeout and cleanup thresholds** | The normative core now fixes separate `target_expires_at` and `pending_expires_at` timers, treats a pending blob slot as a single-upload lease, bounds failed finalization attempts to 3 per slot, requires timeout and retry exhaustion to reuse `upload_state='failed'` with structured `terminal_reason`, and sets a minimum cleanup cadence plus metadata-retention window for failed unattached slots. | Resolved in the current normative core. Future work MAY add richer resumable-upload or same-slot refresh behavior only as an additive extension that preserves the current fail-closed lease and timeout contract. |
| `DQ-012` | **Session lifecycle boundaries** | The normative core now closes session lifecycle boundaries by fixing server-side opaque sessions with `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, and derived `session_expires_at`; a 30-minute idle timeout; a 12-hour absolute lifetime; a 5-session human-user concurrency cap with least-recently-used non-current eviction; current-session logout versus revoke-all triggers; preserved local unsaved work across auth expiry; and a replay-only `resume_token` bound to `{session, incident, client_instance}` that cannot outlive the underlying session. | Resolved in the current normative core. Future work MAY add device-trust or step-up-auth features only as additive extensions that preserve the base opaque-session contract, fail-closed revocation, and provider-agnostic API surface. |
| `DQ-013` | **Deployment-local user-account and incident-membership administration** | The normative core now separates deployment-scoped account administration from incident-scoped membership writes by adding a narrow `deployment_admin` capability, a base `/api/v1/users` route family for local internal users, and explicit incident membership write routes with optimistic concurrency, last-admin guards, and stable `user_id` binding. | Resolved in the current normative core. Future work MAY add invites, provider-driven provisioning, SCIM, password-reset routes, or MFA-reset routes only as additive extensions that preserve the current deployment-local `user_id` and incident-membership contracts. |

## E.2 Source roadmap extract

## 15. Recommended MVP and phased roadmap

### MVP

The MVP must include the spreadsheet-like UX on day one, or it is not an MVP for this problem.

Must-have on day one:

- browser workbook with Timeline, Hosts, Identities, Evidence, Notes tabs
- contract-backed system views for indicators and compromise assessments, with canonical indicators modeled as first-class records, source-bound indicator observations, and incident-scoped lifecycle intervals while preserving raw source text in Timeline, Notes, and Evidence
- inline grid editing with keyboard navigation
- multi-row clipboard paste
- blank-row quick entry
- unresolved mention capture for hosts and identities
- canonical host/identity records and manual resolution flow
- incident-scoped compromise assessment history for hosts and identities
- screenshot/file attachment to rows and standalone evidence requests
- typed links and tags
- data-driven type registries for host/evidence presentation, with local overrides
- row-centric history, attribution, and rollback with single-entry and whole-change-set restore paths
- multi-user live updates and presence
- local auth with TOTP MFA
- background jobs for imports, report snapshots, evidence processing, and optional reference-pack refresh
- container deployment with Postgres + MinIO

### Phase 2

- OIDC integration
- richer host/identity merge workflows
- richer saved-view curation, seeding, and suggestion tooling beyond the base scope and startup-default contract
- first-class `task_request` and `decision` records, plus workbook-native system or saved views for owned work, decisions, communications logs, handoffs, status reviews, and lessons learned without adding new built-in sheets
- stronger evidence previews and metadata extraction
- incident import assistant for XLSX files, prioritizing Timeline, Systems/Hosts, Accounts/Identities, Indicators, Evidence Tracker, and VERIS-like sheets when present; preserving unknown columns in `raw_capture` or `custom_attrs`
- immutable incident snapshots and self-contained report exports with stable identifiers, versioned recipient-specific redaction profiles, and supported templates
- integrity-checked ATT&CK/D3FEND/VERIS/reference-pack distribution for disconnected deployments
- narrow live sensitive-evidence access controls only if repeated real-world incidents show export-scoped withholding is insufficient

### Phase 3

- SAML support
- richer portability tooling, compatibility helpers, and convenience packaging around the Incident Portability Extension Profile
- duplicate detection and resolution suggestions
- cross-record correlation helpers and graph-oriented exploration
- promote `hypothesis` to a first-class `record_type` only if repeated use shows artifact-backed tracking is insufficient for competing explanations, support or contradiction sets, state transitions, or reviewer-visible history
- richer presentation outputs built from immutable snapshots, with a clear boundary between source evidence and generated narrative material
- optional integrations for ingesting external exports

What can safely wait: advanced merge suggestions, enterprise SSO, richer importers, live restricted-evidence policies, and cross-incident analytics. What cannot wait: the grid, paste, evidence attach, progressive structuring, history, collaboration, and the contracts that keep workbook views, overlays, and exports aligned.


## E.3 Source open-questions extract

## 16. Open questions

Resolved in this revision:

- Typed host/account alias auto-resolution is no longer open for MVP. See `### Auto-resolution policy for typed host/account strings`. The system MAY auto-resolve only in interactive mention-capture flows when `auto_resolution_confidence = 100`, with same-surface disclosure, direct undo, and reversible history.
- History granularity is no longer open for MVP. Storage MUST capture history at `change_set` plus mutation-target granularity. The reviewer UI MUST remain row-centric, expose single-entry rollback plus whole-row restore and whole-change-set rollback, and does not require arbitrary field-picker rollback from historical snapshots in MVP.
- Same-field conflict resolution affordance is no longer open for MVP. See Core 03 §3.3 and Core 04 AC-037 through AC-042. Same-field conflicts now resolve through a same-surface compare drawer or equivalent same-surface panel anchored to the conflicted cell, with the saved value retained in the grid, the unsaved local draft kept client-local until explicit resolution, contract-declared `conflict_resolution_class`, and grouped paste conflict handling.
- The public client/server interface surface is no longer open for the current profile. Core 01 now fixes a versioned HTTP+JSON API rooted at `/api/v1/` plus a bounded WebSocket stream rooted at `/ws/v1/`, including a concrete v1 collaboration wire contract with a common envelope, typed `hello` and `resume` handshakes, typed presence payloads, replayable-versus-ephemeral delivery classes, bounded resume and reset behavior, application-level heartbeats, and per-incident ordering; Core 03 aligns same-field conflicts, presence, evidence upload, evidence access, and Timeline read or write flows to that contract; Core 04 fixes the session-based authentication surface, CSRF expectations, and conformance checks for query, mutation, conflict, blob-slot, job, and live-update behavior.
- Session lifecycle boundaries for authenticated HTTP and WebSocket access are no longer implicit. Core 04 now fixes a 30-minute qualifying-activity idle timeout, a 12-hour absolute lifetime, a 5-session human-user concurrency cap with least-recently-used non-current eviction, immediate current-session logout versus revoke-all triggers, and incident-membership loss that closes only the affected incident subscription. Core 01 now exposes `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at` on `GET /api/v1/auth/session`, keeps `resume_token` as a replay-only token bound to `{session, incident, client_instance}`, and forbids heartbeats, passive push, reconnect or replay, and `GET /api/v1/auth/session` from renewing idle expiry; Core 03 preserves client-local pending patches and same-field drafts across auth expiry until re-authentication; Core 04 adds conformance checks for timeout, revocation, and concurrent-session eviction.
- Pending blob timeout and cleanup thresholds are no longer implicit. Core 01 now fixes `target_expires_at` and `pending_expires_at` on the blob-slot route; Core 02 adds structured expiry, failure, and cleanup metadata; Core 03 fixes a 60-minute upload-target expiry, 24-hour pending-slot expiry, a single-upload lease, a maximum of 3 failed finalization attempts per slot, a 15-minute cleanup sweep cadence, and a 7-day failed-metadata retention floor; Core 04 adds conformance checks for timeout, retry exhaustion, and orphaned-byte cleanup.
- The performance envelope for large-grid and evidence-heavy incidents is no longer open for the base profile. Core 01 now requires bounded viewport virtualization, projection-backed hot-path retrieval, deterministic cursor or viewport/block access, metadata-only grid and inspector reads, and binary evidence off the hot path. Core 04 now defines two reference performance fixtures plus AC-043 through AC-047.
- The timeline-sheet grouping-key whitelist is no longer open for the current profile. Core 03 now fixes the current five-key whitelist as sufficient for GA, forbids any sixth base-profile grouping key, and explicitly excludes raw `timeline.event_type` as a base-profile grouping key. Core 04 now makes that whitelist boundary explicit in conformance. See `### Timeline grouping-key whitelist` for the gate a future proposal must satisfy before widening the whitelist.
- The base-profile built-in and system view registry is no longer implicit. Core 01 now fixes exactly nine pack-independent `view_schema_id` entries for Timeline, Hosts, Identities, Evidence, Notes, Indicators, Assessments, Task Requests, and Decisions, each with an ordered field set, deterministic default sort tuple, filter whitelist, and explicit write-back contract; Core 03 aligns Timeline, Indicators, and Assessments interaction semantics; and Core 04 adds conformance checks for registry inspection, no-pack loading, and fail-closed writes to derived or read-only fields.
- The incident-specific custom-metadata promotion question is no longer open for the current profile. Core 02 now defines the promotion rule: a field leaves `custom_attrs` only when it recurs across independent SoD exemplars and is operationally load-bearing for sort, filter, join, lifecycle, dedupe, queueing, or reporting. The current profile now fixes the minimum promoted field sets for incident metadata, task requests, hosts, identities, canonical indicators, and evidence records; keeps framework overlays, enrichment metadata, presentation flags, and low-frequency governance text flexible; requires multi-valued relationships to use typed links or child tables rather than JSONB; and keeps authoritative custody narrative in append-only custody events rather than a single top-level text field.
- Snapshot and report release controls are no longer open for the current profile. Core 01 now fixes the immutable release tuple, `content_class`, `release_scope`, versioned template contracts, versioned redaction profiles, fail-closed redaction, `support_refs[]` on externally releasable curated narrative, and redaction manifests. Core 04 now defines the bounded artifact-release gate and approval invalidation rules.
- Generated presentation depth is no longer open for the current profile. Core 01 now allows arrangement and deterministic templating over snapshot facts, forbids invented facts or synthesized unobserved operator activity, allows `mermaid` and `slidev` external release only within the chosen release scope, and keeps `reenactment` outputs internal-review-only with `generated_presentation=true`.
- Whole-incident portability is no longer open for the current profile. Core 00 now adds the Incident Portability Extension Profile; Core 01 fixes the logical bundle layout, manifest, deterministic JSON/NDJSON bundle format, authoritative-source-only export boundary, portable actor descriptors, optional embedded snapshot/reference-pack sections, and fail-closed import phases; Core 04 adds the corresponding trust-boundary and conformance checks.
- Clipboard paste versus XLSX adoption is no longer open as a single yes-or-no question. The current core keeps clipboard paste on the base workbook interaction path and closes richer file-based onboarding as a Phase 2 Workbook Import Assistant inside the Import Extension Profile behind a dedicated imports module. Whole-workbook import is now defined as an orchestrated batch of explicit `import_unit` objects within one `import_session`; bounded CSV plus discovered XLSX used ranges, tables, eligible named ranges, and operator-selected regions are the current bridge for brownfield workbook migration, while workbook semantics remain downgraded metadata or rejection cases rather than first-class logic.
- Restricted evidence visibility inside a single incident workspace is no longer open for the current profile. Live workbook views remain incident-scoped for authenticated incident participants. Recipient-specific withholding is handled only at snapshot, render, and release time through versioned redaction profiles and optional `disclosure_partition_refs[]`. Future live sensitive-evidence controls remain out of scope unless repeated real-world incidents show that export-scoped withholding is insufficient.
- Shared/private saved views and workbook startup defaults are no longer open for the current profile. Core 01 now adds saved-view delete plus explicit saved-view and workbook-preference API contracts; Core 02 defines the saved-view schema contract plus separate per-user and incident-wide `sheet_ref` preference objects; Core 03 defines the `private`, `shared`, and `system` scope model, immutable `view_schema_id`, owner/admin mutability rules, and deterministic startup fallback from explicit `sheet_ref` to user home to incident default to Timeline; Core 04 adds authorization and conformance checks that saved-view scope affects only discoverability of the saved-view object, not underlying incident-data visibility.
- The dedicated Notes tab is no longer open for the current profile. Core 01 now fixes Notes as a built-in sheet keyed by stable `view_schema_id` and backed by `artifact_grid_projection` filtered to `artifact_type='note'`; Core 02 keeps note storage artifact-backed and links contextual notes through generic `record_links`; Core 04 adds acceptance criteria for direct and contextual note creation and for excluding raw note working material from `external_release`. Current analyst-work coordination objects now use distinct `record_type`s or `artifact_type`s and MUST NOT overload `artifact_type='note'`. A future NLSpec MAY revisit whether Notes remains first-class only after those richer objects ship and adoption evidence shows the dedicated sheet is redundant; it is not a current-profile open question.
- Indicator storage promotion is no longer open for the current profile. Core 01 now fixes the Indicators system view as a projection over canonical indicator records; Core 02 requires first-class canonical indicators, source-bound `indicator_observation` rows, and append-only indicator lifecycle intervals; Core 03 keeps indicator capture embedded in raw source fields and places indicator linking in the same-surface enrichment flow; Core 04 adds acceptance criteria for distinct observations, canonical dedupe, lifecycle history, and stable one-row-per-indicator system-view behavior.
- The minimum disconnected reference-pack bundle is no longer open for the current profile. Core 01 now fixes the smallest supported disconnected bundle as exactly three preinstalled type-registry packs, `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`; keeps framework, enrichment, template, and separately distributed `view_contract` packs outside that minimum bundle; and requires offline staged import plus explicit activation. Core 02 now requires structured activation-state and attestation metadata. Core 04 adds fail-closed verification, retained prior active version, no-live-fetch disconnected operation, and conformance criteria for the pack lifecycle.
- Assessment vocabulary and confidence model are no longer open for the current profile. Core 01 now fixes the Compromise Assessments system view as a projection over append-only assessment records; Core 02 fixes closed compromise-assessment states `unknown`, `suspected`, `confirmed`, `disproven`, and `cleared`, separates assessment judgment from response posture, and defines nullable `confidence_score` plus derived `confidence_band`; Core 03 makes interactive assessment entry band-first and keeps filtering separate by `assessment_state` and `confidence_band`; Core 04 adds acceptance criteria for distinct `disproven` versus `cleared` behavior and for keeping response actions separate from assessment state.
- The analyst-work tracking boundary is no longer open for the current profile. Core 00 now places `task_request`, `decision`, and structured coordination artifacts inside the current system boundary; Core 01 adds task and decision system views plus projection support while keeping communications, handoff, status-review, and lesson surfaces off the built-in-sheet set; Core 02 defines the promotion rule, first-class `task_request` and `decision` records, ownership as a field on coordination objects, artifact-backed `comm_log`, `handoff`, `status_review`, and `lesson`, and current artifact-backed `hypothesis`; Core 03 keeps routine timeline capture free of mandatory owner or approval fields while adding workbook-native coordination surfaces; Core 04 adds authorization, release-boundary, and conformance checks for these objects.

### Analyst-work tracking boundary

Tags and notes remain sufficient only for unassigned, non-lifecycle working material. The first analyst-work concepts promoted in the current profile are `task_request` and `decision`.

Ownership remains a required field on coordination objects rather than a standalone record type or a mandatory field on every timeline row. `comm_log`, `handoff`, `status_review`, and `lesson` remain artifact-backed coordination surfaces. `hypothesis` remains artifact-backed until repeated use shows a need for first-class competing-hypothesis state, explicit support or contradiction sets, or reviewer-visible hypothesis history.

Current ordering: `task_request`, `decision`, ownership on coordination objects, `comm_log`, `handoff`, `status_review`, `lesson`, and only later `hypothesis`.

Explicit lifecycle machines for `task_request` and `decision` statuses are now part of the current core. The core now defines legal transitions, reciprocal guard normalization, explicit supersession semantics, and illegal-transition plus idempotent-retry conformance hooks. Future work MAY add richer workflow presentation or operator conveniences without changing these base status semantics or introducing a generalized workflow engine.

### Performance envelope for large incidents and evidence-heavy incidents

This revision closes the performance question with two reference fixtures rather than a single vague "large incident" threshold. That keeps grid pressure and evidence-path pressure separable.

**Fixture A: large-grid incident**

- 20,000 timeline rows
- 1,000 host rows
- 1,000 identity rows
- 25 concurrently connected analyst sessions on one incident, with presence enabled and representative live row-update traffic
- representative tags, mentions, and links, but not evidence-heavy per row

**Fixture B: evidence-heavy incident**

- 5,000 timeline rows
- 10,000 evidence records
- tens of GB of binary evidence stored in object storage
- at least one timeline row linked to 100 evidence records
- evidence blobs MAY be stubbed for throughput tests, but evidence metadata, counts, attachment state, and preview handles MUST be real

**Immediate hot path**

- selection move, focus move, and typing acknowledgment: at or below 100 ms p95 on LAN
- blank-row creation in the timeline sheet with one non-empty user-entered value: at or below 150 ms p95 on LAN
- row identity and selection remain anchored to `record_id` during live updates, filtering, sorting, and grouping

**Short-delay interactive**

- sort, filter, and grouping changes: first useful viewport at or below 250 ms p95
- full stable viewport and result ordering: at or below 1.0 s p95
- inspector metadata shell for a row linked to 100 evidence records: at or below 300 ms p95

**Staged/background**

- imports, evidence processing, projection rebuilds, snapshot generation, report generation, and reference-pack refresh remain background jobs
- progress and cancellation appear within 1 second of job start
- grid editing and row creation remain responsive while those jobs run

### Timeline grouping-key whitelist

This revision closes the timeline grouping-key question by keeping the base-profile whitelist at exactly five keys: `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions`. The current whitelist is sufficient for GA.

Raw `timeline.event_type` remains out of scope for base-profile grouping because it would couple grouping behavior to evolving source vocabularies and imported taxonomies. If a later revision adds one more grouping key, the next candidate should be a derived bucket such as `timeline.event_family`, not raw `timeline.event_type`.

A future proposal to widen the base-profile whitelist is eligible only if all of the following gates are met:

- the candidate key is a scalar contract-backed value already available from `timeline_grid_projection` or an equivalent projection, with no scan of free text, no join against blob or evidence metadata, and no dependence on visible labels
- the candidate key uses a stable closed vocabulary or deterministic bucketization, not raw imported labels
- analyst testing shows that sort, filter, saved views, or dedicated derived views are materially worse for the target task
- the added key preserves the sort, filter, and grouping interaction envelope defined in Core 04

Remaining open questions:

None in the current profile.
