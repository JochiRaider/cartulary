# Appendix E: Roadmap, Open Questions, and Decision Backlog

This appendix is **non-normative**.

It preserves roadmap and open-question material from the exploratory source artifact and adds an editorial decision backlog used to keep the provisional normative core internally consistent.

## E.1 Editorial decision backlog

The source artifact mixed current requirements, roadmap positioning, and open questions in a few places. The provisional normative core resolves those tensions with bounded current-profile rules and extension profiles. This backlog records the remaining editorial pressure points.

| Decision ID | Topic | Current normative treatment | Why it remains on the backlog |
| --- | --- | --- | --- |
| `DQ-002` | **File-based structured import** boundary | The normative core keeps clipboard paste in the base workbook surface and places file-based structured import in the **Import Extension Profile** as a dedicated imports-module boundary with a bounded CSV and selected-sheet or selected-region XLSX contract. | The source artifact defined richer XLSX-assistant behavior, but the current core intentionally defers whole-workbook fidelity and other spreadsheet-specific compatibility work. |
| `DQ-003` | **Immutable snapshots and report exports** status | The normative core places snapshot and report generation in the **Snapshot and Reporting Extension Profile** and fixes the immutable release tuple, release scopes, template contracts, redaction profiles, and artifact release records inside that profile. | The source artifact treats snapshot-based reporting as architecturally important while also placing productized export delivery in later phases. |
| `DQ-004` | **Reference-pack refresh and distribution** status | The normative core places refresh and integrity-checked distribution in the **Reference Pack Extension Profile** and now fixes the smallest disconnected bundle, offline import flow, explicit activation, retained prior active version, and structured attestation metadata. | Future work may still add richer mirroring, signing, or distribution tooling without changing the current disconnected-pack contract. |
| `DQ-005` | **Incident portability contract** | The normative core now places whole-incident export/import in the **Incident Portability Extension Profile** and fixes the logical bundle layout, manifest, deterministic JSON/NDJSON serialization, portable actor descriptors, import phases, and fail-closed bundle verification rules. | Future work may still add richer migration tooling, alternative transport wrappers, or operator conveniences without changing the current portability contract. |
| `DQ-006` | **Restricted evidence visibility versus export-scoped withholding** | The normative core keeps live workspace visibility incident-scoped, routes recipient-specific withholding through snapshot, render, and release controls, and leaves narrow live sensitive-evidence controls to future work. | The source artifact mixed restricted-evidence questions with redaction and reporting controls; the current core separates export disclosure from live workspace authorization while preserving a future path if repeated practice proves it necessary. |
| `DQ-008` | **Generated presentation depth** | The normative core now fixes `content_class`, `release_scope`, `support_refs[]`, versioned template and redaction controls, and the current generated-presentation boundary, including `internal_review`-only reenactments marked `generated_presentation=true`. | Future work may still add richer authoring or visualization families without relaxing the current evidence-versus-presentation boundary. |

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
- shared/private saved views with better defaults
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
- The performance envelope for large-grid and evidence-heavy incidents is no longer open for the base profile. Core 01 now requires bounded viewport virtualization, projection-backed hot-path retrieval, deterministic cursor or viewport/block access, metadata-only grid and inspector reads, and binary evidence off the hot path. Core 04 now defines two reference performance fixtures plus AC-043 through AC-047.
- The timeline-sheet grouping-key whitelist is no longer open for the current profile. Core 03 now fixes the current five-key whitelist as sufficient for GA, forbids any sixth base-profile grouping key, and explicitly excludes raw `timeline.event_type` as a base-profile grouping key. Core 04 now makes that whitelist boundary explicit in conformance. See `### Timeline grouping-key whitelist` for the gate a future proposal must satisfy before widening the whitelist.
- The incident-specific custom-metadata promotion question is no longer open for the current profile. Core 02 now defines the promotion rule: a field leaves `custom_attrs` only when it recurs across independent SoD exemplars and is operationally load-bearing for sort, filter, join, lifecycle, dedupe, queueing, or reporting. The current profile now fixes the minimum promoted field sets for incident metadata, task requests, hosts, identities, canonical indicators, and evidence records; keeps framework overlays, enrichment metadata, presentation flags, and low-frequency governance text flexible; requires multi-valued relationships to use typed links or child tables rather than JSONB; and keeps authoritative custody narrative in append-only custody events rather than a single top-level text field.
- Snapshot and report release controls are no longer open for the current profile. Core 01 now fixes the immutable release tuple, `content_class`, `release_scope`, versioned template contracts, versioned redaction profiles, fail-closed redaction, `support_refs[]` on externally releasable curated narrative, and redaction manifests. Core 04 now defines the bounded artifact-release gate and approval invalidation rules.
- Generated presentation depth is no longer open for the current profile. Core 01 now allows arrangement and deterministic templating over snapshot facts, forbids invented facts or synthesized unobserved operator activity, allows `mermaid` and `slidev` external release only within the chosen release scope, and keeps `reenactment` outputs internal-review-only with `generated_presentation=true`.
- Whole-incident portability is no longer open for the current profile. Core 00 now adds the Incident Portability Extension Profile; Core 01 fixes the logical bundle layout, manifest, deterministic JSON/NDJSON bundle format, authoritative-source-only export boundary, portable actor descriptors, optional embedded snapshot/reference-pack sections, and fail-closed import phases; Core 04 adds the corresponding trust-boundary and conformance checks.
- Clipboard paste versus XLSX adoption is no longer open as a single yes-or-no question. The current core keeps clipboard paste on the base workbook interaction path and treats file-based structured import as a separate Import Extension Profile behind a dedicated imports module. Clipboard paste is therefore sufficient to validate the grid hot path, but not brownfield workbook migration readiness; bounded CSV and selected-sheet or selected-region XLSX onboarding are the current file-based bridge.
- Restricted evidence visibility inside a single incident workspace is no longer open for the current profile. Live workbook views remain incident-scoped for authenticated incident participants. Recipient-specific withholding is handled only at snapshot, render, and release time through versioned redaction profiles and optional `disclosure_partition_refs[]`. Future live sensitive-evidence controls remain out of scope unless repeated real-world incidents show that export-scoped withholding is insufficient.
- The dedicated Notes tab is no longer open for the current profile. Core 01 now fixes Notes as a built-in sheet keyed by stable `view_schema_id` and backed by `artifact_grid_projection` filtered to `artifact_type='note'`; Core 02 keeps note storage artifact-backed and links contextual notes through generic `record_links`; Core 04 adds acceptance criteria for direct and contextual note creation and for excluding raw note working material from `external_release`. Current analyst-work coordination objects now use distinct `record_type`s or `artifact_type`s and MUST NOT overload `artifact_type='note'`. A future NLSpec MAY revisit whether Notes remains first-class only after those richer objects ship and adoption evidence shows the dedicated sheet is redundant; it is not a current-profile open question.
- Indicator storage promotion is no longer open for the current profile. Core 01 now fixes the Indicators system view as a projection over canonical indicator records; Core 02 requires first-class canonical indicators, source-bound `indicator_observation` rows, and append-only indicator lifecycle intervals; Core 03 keeps indicator capture embedded in raw source fields and places indicator linking in the same-surface enrichment flow; Core 04 adds acceptance criteria for distinct observations, canonical dedupe, lifecycle history, and stable one-row-per-indicator system-view behavior.
- The minimum disconnected reference-pack bundle is no longer open for the current profile. Core 01 now fixes the smallest supported disconnected bundle as exactly three preinstalled type-registry packs, `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`; keeps framework, enrichment, template, and separately distributed `view_contract` packs outside that minimum bundle; and requires offline staged import plus explicit activation. Core 02 now requires structured activation-state and attestation metadata. Core 04 adds fail-closed verification, retained prior active version, no-live-fetch disconnected operation, and conformance criteria for the pack lifecycle.
- Assessment vocabulary and confidence model are no longer open for the current profile. Core 01 now fixes the Compromise Assessments system view as a projection over append-only assessment records; Core 02 fixes closed compromise-assessment states `unknown`, `suspected`, `confirmed`, `disproven`, and `cleared`, separates assessment judgment from response posture, and defines nullable `confidence_score` plus derived `confidence_band`; Core 03 makes interactive assessment entry band-first and keeps filtering separate by `assessment_state` and `confidence_band`; Core 04 adds acceptance criteria for distinct `disproven` versus `cleared` behavior and for keeping response actions separate from assessment state.
- The analyst-work tracking boundary is no longer open for the current profile. Core 00 now places `task_request`, `decision`, and structured coordination artifacts inside the current system boundary; Core 01 adds task and decision system views plus projection support while keeping communications, handoff, status-review, and lesson surfaces off the built-in-sheet set; Core 02 defines the promotion rule, first-class `task_request` and `decision` records, ownership as a field on coordination objects, artifact-backed `comm_log`, `handoff`, `status_review`, and `lesson`, and current artifact-backed `hypothesis`; Core 03 keeps routine timeline capture free of mandatory owner or approval fields while adding workbook-native coordination surfaces; Core 04 adds authorization, release-boundary, and conformance checks for these objects.

### Analyst-work tracking boundary

Tags and notes remain sufficient only for unassigned, non-lifecycle working material. The first analyst-work concepts promoted in the current profile are `task_request` and `decision`.

Ownership remains a required field on coordination objects rather than a standalone record type or a mandatory field on every timeline row. `comm_log`, `handoff`, `status_review`, and `lesson` remain artifact-backed coordination surfaces. `hypothesis` remains artifact-backed until repeated use shows a need for first-class competing-hypothesis state, explicit support or contradiction sets, or reviewer-visible hypothesis history.

Current ordering: `task_request`, `decision`, ownership on coordination objects, `comm_log`, `handoff`, `status_review`, `lesson`, and only later `hypothesis`.

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
