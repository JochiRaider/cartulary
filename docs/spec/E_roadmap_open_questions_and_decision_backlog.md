# Appendix E: Roadmap, Open Questions, and Decision Backlog

This appendix is **non-normative**.

It preserves roadmap and open-question material from the exploratory source artifact and adds an editorial decision backlog used to keep the provisional normative core internally consistent.

## E.1 Editorial decision backlog

The source artifact mixed current requirements, roadmap positioning, and open questions in a few places. The provisional normative core resolves those tensions with bounded current-profile rules and extension profiles. This backlog records the remaining editorial pressure points.

| Decision ID | Topic | Current normative treatment | Why it remains on the backlog |
| --- | --- | --- | --- |
| `DQ-001` | Dedicated **Notes** tab versus artifact-backed note capture only | The normative core keeps **Notes** as a built-in sheet in the current profile. | The source artifact lists Notes as a built-in sheet and day-one surface, but also asks whether analysts want a dedicated Notes tab long term. |
| `DQ-002` | **File-based structured import** boundary | The normative core keeps clipboard paste in the base workbook surface and places file-based structured import in the **Import Extension Profile** as a dedicated imports-module boundary with a bounded CSV and selected-sheet or selected-region XLSX contract. | The source artifact defined richer XLSX-assistant behavior, but the current core intentionally defers whole-workbook fidelity and other spreadsheet-specific compatibility work. |
| `DQ-003` | **Immutable snapshots and report exports** status | The normative core places snapshot and report generation in the **Snapshot and Reporting Extension Profile** and fixes the immutable release tuple, release scopes, template contracts, redaction profiles, and artifact release records inside that profile. | The source artifact treats snapshot-based reporting as architecturally important while also placing productized export delivery in later phases. |
| `DQ-004` | **Reference-pack refresh and distribution** status | The normative core places refresh and integrity-checked distribution in the **Reference Pack Extension Profile** while keeping degradation semantics normative. | The source artifact makes packs optional in the base design but mixes current operational expectations with later disconnected-distribution work. |
| `DQ-005` | **Indicator storage promotion** | The normative core fixes the indicator projection/API contract but does not require a dedicated indicator source table. | The source artifact intentionally leaves storage promotion contingent on real usage. |
| `DQ-006` | **Restricted evidence visibility versus export-scoped withholding** | The normative core keeps live workspace visibility incident-scoped, routes recipient-specific withholding through snapshot, render, and release controls, and leaves narrow live sensitive-evidence controls to future work. | The source artifact mixed restricted-evidence questions with redaction and reporting controls; the current core separates export disclosure from live workspace authorization while preserving a future path if repeated practice proves it necessary. |
| `DQ-007` | **Assessment vocabulary and confidence model** | The normative core preserves the current assessment-history concept and stateful record shape. | The exact vocabulary fit remains an explicit source question tied to analyst practice. |
| `DQ-008` | **Generated presentation depth** | The normative core now fixes `content_class`, `release_scope`, `support_refs[]`, versioned template and redaction controls, and the current generated-presentation boundary, including `internal_review`-only reenactments marked `generated_presentation=true`. | Future work may still add richer authoring or visualization families without relaxing the current evidence-versus-presentation boundary. |

## E.2 Source roadmap extract

## 15. Recommended MVP and phased roadmap

### MVP

The MVP must include the spreadsheet-like UX on day one, or it is not an MVP for this problem.

Must-have on day one:

- browser workbook with Timeline, Hosts, Identities, Evidence, Notes tabs
- contract-backed system views for indicators and compromise assessments, even if indicator storage is artifact-backed initially
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
- stronger evidence previews and metadata extraction
- incident import assistant for XLSX files, prioritizing Timeline, Systems/Hosts, Accounts/Identities, Indicators, Evidence Tracker, and VERIS-like sheets when present; preserving unknown columns in `raw_capture` or `custom_attrs`
- immutable incident snapshots and self-contained report exports with stable identifiers, versioned recipient-specific redaction profiles, and supported templates
- integrity-checked ATT&CK/D3FEND/VERIS/reference-pack distribution for disconnected deployments
- first-class indicator objects if teams outgrow artifact-backed handling
- narrow live sensitive-evidence access controls only if repeated real-world incidents show export-scoped withholding is insufficient

### Phase 3

- SAML support
- incident export/import bundles with integrity manifest
- duplicate detection and resolution suggestions
- cross-record correlation helpers and graph-oriented exploration
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
- Snapshot and report release controls are no longer open for the current profile. Core 01 now fixes the immutable release tuple, `content_class`, `release_scope`, versioned template contracts, versioned redaction profiles, fail-closed redaction, `support_refs[]` on externally releasable curated narrative, and redaction manifests. Core 04 now defines the bounded artifact-release gate and approval invalidation rules.
- Generated presentation depth is no longer open for the current profile. Core 01 now allows arrangement and deterministic templating over snapshot facts, forbids invented facts or synthesized unobserved operator activity, allows `mermaid` and `slidev` external release only within the chosen release scope, and keeps `reenactment` outputs internal-review-only with `generated_presentation=true`.
- Clipboard paste versus XLSX adoption is no longer open as a single yes-or-no question. The current core keeps clipboard paste on the base workbook interaction path and treats file-based structured import as a separate Import Extension Profile behind a dedicated imports module. Clipboard paste is therefore sufficient to validate the grid hot path, but not brownfield workbook migration readiness; bounded CSV and selected-sheet or selected-region XLSX onboarding are the current file-based bridge.
- Restricted evidence visibility inside a single incident workspace is no longer open for the current profile. Live workbook views remain incident-scoped for authenticated incident participants. Recipient-specific withholding is handled only at snapshot, render, and release time through versioned redaction profiles and optional `disclosure_partition_refs[]`. Future live sensitive-evidence controls remain out of scope unless repeated real-world incidents show that export-scoped withholding is insufficient.

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

Remaining open questions:

1. Is the timeline-sheet grouping-key whitelist (`timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, `timeline.has_unresolved_mentions`) sufficient for GA, or does analyst testing show a need for one additional scalar grouping key?
1. Do analysts want a dedicated Notes tab, or is artifact-backed note capture from the timeline sufficient?
1. How much incident-specific custom metadata is real, and which of those fields become common enough to deserve first-class columns?
1. At what point do artifact-backed indicators outgrow their storage implementation and warrant promotion to a dedicated source table, given that the indicator projection/API contract is fixed from MVP?
1. Which assessment-state vocabulary and confidence model best match analyst practice for incident-scoped host/identity compromise assessments?
1. When do tags and notes stop being enough, and which analyst-work concepts most need explicit modeling: tasks, hypotheses, decisions, or ownership?
1. Which optional reference packs should ship in the smallest disconnected deployment, and what update/attestation flow is acceptable operationally?
