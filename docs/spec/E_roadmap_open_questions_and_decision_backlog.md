# Appendix E: Roadmap, Open Questions, and Decision Backlog

This appendix is **non-normative**.

It preserves roadmap and open-question material from the exploratory source artifact and adds an editorial decision backlog used to keep the provisional normative core internally consistent.

## E.1 Editorial decision backlog

The source artifact mixed current requirements, roadmap positioning, and open questions in a few places. The provisional normative core resolves those tensions with bounded current-profile rules and extension profiles. This backlog records the remaining editorial pressure points.

| Decision ID | Topic | Current normative treatment | Why it remains on the backlog |
| --- | --- | --- | --- |
| `DQ-001` | Dedicated **Notes** tab versus artifact-backed note capture only | The normative core keeps **Notes** as a built-in sheet in the current profile. | The source artifact lists Notes as a built-in sheet and day-one surface, but also asks whether analysts want a dedicated Notes tab long term. |
| `DQ-002` | **Full XLSX import** status | The normative core places full XLSX import in the **Import Extension Profile**. Clipboard paste remains base-profile functionality. | The source artifact defines XLSX import behavior and acceptance expectations, but also places the import assistant in Phase 2. |
| `DQ-003` | **Immutable snapshots and report exports** status | The normative core places snapshot and report generation in the **Snapshot and Reporting Extension Profile**. | The source artifact treats snapshot-based reporting as architecturally important while also placing productized export delivery in later phases. |
| `DQ-004` | **Reference-pack refresh and distribution** status | The normative core places refresh and integrity-checked distribution in the **Reference Pack Extension Profile** while keeping degradation semantics normative. | The source artifact makes packs optional in the base design but mixes current operational expectations with later disconnected-distribution work. |
| `DQ-005` | **Indicator storage promotion** | The normative core fixes the indicator projection/API contract but does not require a dedicated indicator source table. | The source artifact intentionally leaves storage promotion contingent on real usage. |
| `DQ-006` | **Sensitive evidence controls** | The normative core keeps incident-level roles only and treats sensitive-evidence controls as future work. | The source artifact repeatedly frames restricted evidence as a possible later refinement rather than a day-one requirement. |
| `DQ-007` | **Assessment vocabulary and confidence model** | The normative core preserves the current assessment-history concept and stateful record shape. | The exact vocabulary fit remains an explicit source question tied to analyst practice. |
| `DQ-008` | **Generated presentation depth** | The normative core preserves the snapshot boundary and self-contained output rule without specifying richer narrative-control features. | The source artifact explicitly asks how far generated narrative outputs should go before they blur evidence versus presentation. |

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
- immutable incident snapshots and self-contained report exports with stable identifiers
- integrity-checked ATT&CK/D3FEND/VERIS/reference-pack distribution for disconnected deployments
- first-class indicator objects if teams outgrow artifact-backed handling
- sensitive-evidence access controls if truly needed

### Phase 3

- SAML support
- incident export/import bundles with integrity manifest
- duplicate detection and resolution suggestions
- cross-record correlation helpers and graph-oriented exploration
- richer presentation outputs built from immutable snapshots, with a clear boundary between source evidence and generated narrative material
- optional integrations for ingesting external exports

What can safely wait: advanced merge suggestions, enterprise SSO, richer importers, restricted-evidence policies, and cross-incident analytics. What cannot wait: the grid, paste, evidence attach, progressive structuring, history, collaboration, and the contracts that keep workbook views, overlays, and exports aligned.


## E.3 Source open-questions extract

## 16. Open questions

Resolved in this revision:

- Typed host/account alias auto-resolution is no longer open for MVP. See `### Auto-resolution policy for typed host/account strings`. The system MAY auto-resolve only in interactive mention-capture flows when `auto_resolution_confidence = 100`, with same-surface disclosure, direct undo, and reversible history.
- History granularity is no longer open for MVP. Storage MUST capture history at `change_set` plus mutation-target granularity. The reviewer UI MUST remain row-centric, expose single-entry rollback plus whole-row restore and whole-change-set rollback, and does not require arbitrary field-picker rollback from historical snapshots in MVP.
- Same-field conflict resolution affordance is no longer open for MVP. See Core 03 §3.3 and Core 04 AC-037 through AC-042. Same-field conflicts now resolve through a same-surface compare drawer or equivalent same-surface panel anchored to the conflicted cell, with the saved value retained in the grid, the unsaved local draft kept client-local until explicit resolution, contract-declared `conflict_resolution_class`, and grouped paste conflict handling.
- The performance envelope for large-grid and evidence-heavy incidents is no longer open for the base profile. Core 01 now requires bounded viewport virtualization, projection-backed hot-path retrieval, deterministic cursor or viewport/block access, metadata-only grid and inspector reads, and binary evidence off the hot path. Core 04 now defines two reference performance fixtures plus AC-043 through AC-047.

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
1. Is clipboard paste enough to validate adoption before the prioritized XLSX import assistant ships?
1. How often do teams need restricted evidence visibility inside a single incident workspace, and is that worth the UX cost early?
1. At what point do artifact-backed indicators outgrow their storage implementation and warrant promotion to a dedicated source table, given that the indicator projection/API contract is fixed from MVP?
1. Which assessment-state vocabulary and confidence model best match analyst practice for incident-scoped host/identity compromise assessments?
1. When do tags and notes stop being enough, and which analyst-work concepts most need explicit modeling: tasks, hypotheses, decisions, or ownership?
1. What redaction, approval, and template controls should sit on top of immutable incident snapshots once report generation is productized?
1. How far should generated presentation outputs such as Mermaid, Slidev, or Asciinema-style reenactments go before they blur the distinction between evidence and narrative?
1. Which optional reference packs should ship in the smallest disconnected deployment, and what update/attestation flow is acceptable operationally?
