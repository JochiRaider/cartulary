# Appendix A: Problem Framing, Rationale, Trade-offs, and Sanity Check

This appendix is **non-normative**.

It preserves explanatory context from the exploratory source artifact so future editors can see why the normative core was shaped the way it was.

## 1. Executive summary

I recommend a **modular monolith**: a browser-based workbook UI backed by **Postgres as the source of truth** and **S3-compatible object storage** for binary evidence, packaged for three-container deployment in the smallest useful form. The core design thesis is that the spreadsheet metaphor should survive **at the view layer**, not at the storage layer: analysts work in sheet-like grids, while the system stores normalized records, typed relationships, revisions, versioned reference packs, and evidence metadata underneath. Built-in sheets should remain intentionally few, with additional tabs expressed as saved or system views over the same projections rather than new storage silos. The key mechanism is **capture first, structure later**: a timeline row can be created with partial facts and unresolved host/account strings, and those rough tokens are preserved as first-class records until another analyst resolves them into canonical hosts and identities. That preserves the low-friction spreadsheet feel without collapsing into an uncontrolled spreadsheet. The system justifies replacing Excel by adding capabilities spreadsheets handle badly: **explicit links across events/entities/evidence, reliable search/correlation, attributed edits, revision history, rollback, binary evidence attachment, and self-contained report/export snapshots**. Reports and presentation artifacts should be generated from immutable incident snapshots using the same canonical derivation layer as operator-facing views, so UI and export do not drift. If the product cannot get within one interaction of Excel for creating and editing rows, it will fail regardless of how good the schema is.


## 2. Problem framing

The SOD exists because it is operationally honest. Excel does four things extremely well in an M365 incident response workflow: it is familiar, it is instantly shareable, it tolerates incomplete information, and it lets analysts type directly into cells with almost no ceremony. That is not a cosmetic advantage; during active IR, it is the difference between capturing a fact now and intending to capture it later.

Replacing it is hard because most purpose-built systems optimize for structure before speed. They assume records should be created through forms, validation should happen up front, and relationships should be explicit at entry time. That is exactly backwards for live investigations, where the first useful version of a fact is often messy: “probably jdoe on WS-023 via VPN; screenshot attached; time maybe 09:14.” A good replacement must preserve that immediacy while adding structure later.

The primary jobs-to-be-done are different by role:

- **Front-line analysts** need to capture observations fast, paste blocks of data, attach screenshots/files, and keep working when facts are incomplete or uncertain.
- **Incident commanders / leads** need shared visibility, common views, rapid filtering, a trustworthy timeline, and awareness of what has or has not been normalized.
- **Reviewers / report writers** need lineage, edit attribution, stable references, evidence retrieval, and the ability to inspect or revert mistaken changes.

Low-friction capture is the make-or-break requirement because IR work is lossy under stress. If an analyst has to leave the grid, open a form, satisfy validation, or decide where a fact “belongs” before typing it, the system is already worse than the workbook it is trying to replace.


## 3. Design principles and assumptions

### Assumptions

- A single incident workspace usually has **2–8 active users**, occasionally up to **25**.
- A serious incident may accumulate **1k–20k timeline rows**, **hundreds to low thousands** of host/identity records, and **tens of GB** of evidence.
- Deployments are **single-tenant** initially.
- “Disconnected” means the deployment can run in an isolated environment; it does **not** imply fully offline browser sync or multi-master replication.
- The default client is a **browser UI**, not a desktop app.
- The system must support **rough capture first** and **progressive normalization** without losing original user-entered text.
- Binary evidence may be large; it should not be stored inline in Postgres.
- Optional reference packs such as framework mappings, type registries, and enrichment datasets may be present or absent by deployment; the core workbook must remain usable without them.
- Report and presentation exports may need to run in disconnected environments and therefore cannot require remote runtime assets.

### Design principles

1. **Grid first, forms second.**\
   The primary interaction model is a workbook-like grid with inline editing, keyboard navigation, and paste. Forms belong in a side inspector for enrichment, not as the default path.

1. **Capture first, structure later.**\
   Rough rows, unresolved host/account strings, uncertain timestamps, and ad hoc notes are valid first-class records. The system must preserve raw input and support later normalization.

1. **Collaborative editing with attribution by default.**\
   Every mutation must be tied to an authenticated user and written with revision history. Collaboration cannot assume single-user ownership of a row.

1. **Preserve context during live investigations.**\
   Common actions—create row, attach screenshot, link host, inspect history—must happen without forcing the analyst into full-page navigation or modal-heavy flows.

1. **Operational simplicity beats architectural purity.**\
   Put the complexity in the data model and UI behavior, not in distributed infrastructure. Prefer a modular monolith over microservices.

1. **Views may be denormalized; source data may not be sloppy.**\
   Spreadsheet-style sheets should be projections over disciplined source tables, not the source of truth themselves.

1. **Few first-class sheets, many saved/system views.**
   The workbook should expose a small set of core tabs and treat indicator, assessment, and framework overlays as contract-backed views until usage justifies promotion.

1. **Reference packs and enrichment are optional overlays.**
   Framework mappings, type registries, and other enrichment datasets must be versioned separately from incident data and cannot sit on the hot capture path.

1. **Behavior follows explicit contracts, not labels.**
   View semantics, computed columns, and write-back rules must be keyed by stable identifiers such as `view_schema_id`, not by visible tab names, column headers, or UI control names.

1. **One canonical derivation layer feeds every derived surface.**
   Visualizations, report sections, framework rollups, and exports should reuse the same extraction/query logic so filters and counts do not drift between UI and output.


## 13. Trade-offs, risks, and rejected alternatives

### Rejected approach 1: strict normalized forms-first application

This is the easiest backend to design and the easiest schema to keep “clean.” It is also the fastest way to lose analyst adoption. Forcing users to decide whether a rough fact is a timeline item, host, identity, or note **before typing it** is exactly the wrong interaction model for live IR.

### Rejected approach 2: single-user local notebook/tool

Single-user notebooks can be fast and flexible. They also fail the collaboration, attribution, shared visibility, and evidence governance requirements. That makes them a poor replacement for an M365-hosted workbook that multiple analysts already use simultaneously.

### Recommended approach: spreadsheet-like grid on top of relational source tables

This is the hardest option to implement well, but it is the only one that attacks the actual problem. It preserves near-spreadsheet entry friction while adding the relational and audit capabilities that justify migration.

### Hardest implementation risks

1. **Write-back correctness from denormalized cells to normalized source tables.**\
   This is where bugs will hide. Relationship cells cannot be implemented as naive string fields.

1. **Concurrency UX.**\
   Field-level optimistic merge is manageable; same-cell conflicts need careful UX to avoid annoyance or silent data loss.

1. **Projection freshness and rebuild discipline.**\
   Projection tables must be treated as disposable caches and consistently rebuilt on change.

1. **Entity resolution UX.**\
   If resolving rough host/account strings feels slow or unclear, analysts will stay in unresolved text forever.

1. **Grid performance at realistic incident sizes.**\
   Virtualization, batched mutations, and projection indexing are not optional.

1. **Allowing view behavior to depend on headers, labels, or duplicated derivation logic.**
   If UI views, framework overlays, and exports each implement their own extraction rules, the product will drift into brittle inconsistencies.

### One more rejected alternative: full CRDT/Google-Sheets-style cell engine

It is possible to build a true collaborative cell engine. It is also the wrong place to spend complexity budget for this product. IR needs auditable records, typed links, and rollback more than it needs character-by-character co-editing of the same cell.


## 14. Comparison table

| Dimension                 | Excel-based SOD              | Proposed system                                                    | Typical single-user IR notebook/tool |
| ------------------------- | ---------------------------- | ------------------------------------------------------------------ | ------------------------------------ |
| Entry friction            | Excellent                    | **Near-Excel by design; must be close enough to be credible**      | Good for one user                    |
| Collaboration             | Good in M365                 | **Strong: shared views, presence, live updates, attributed edits** | Weak                                 |
| Structure / relationships | Weak, implicit               | **Strong: typed links, canonical entities, unresolved mentions**   | Medium                               |
| Auditability              | Weak                         | **Strong: per-user revisions, change sets, rollback**              | Medium                               |
| Evidence handling         | Weak; attachments awkward    | **Strong: object-backed evidence linked to rows/entities**         | Medium                               |
| Search / correlation      | Weak beyond filters          | **Strong: links, aliases, FTS, relation-driven navigation**        | Medium                               |
| Deployment simplicity     | Excellent if already in M365 | Good: 3-container footprint                                        | Good locally, poor for teams         |
| Training burden           | Very low                     | Moderate but workbook-shaped                                       | Moderate                             |
| Rollback / versioning     | Weak and workbook-wide       | **Strong: row-centric history with mutation-entry rollback**        | Variable                             |
| Data discipline           | Low                          | **Balanced: fast capture with progressive structure**              | Medium                               |

The proposed system does **not** need to beat Excel on raw familiarity. It **does** need to clearly beat Excel on **relationships, evidence handling, auditability, rollback, and collaborative visibility**, while staying close enough on entry speed that analysts do not resent it.


## Design sanity check

- **Does the design preserve spreadsheet-like low-friction entry?** Yes — the grid is the primary UI, row creation is inline, and unresolved mentions allow rough entry without forms.
- **Does it support multi-user concurrent editing?** Yes — WebSocket presence plus field-level optimistic concurrency, row versioning, and stable `record_id` binding.
- **Does it support progressive structuring of partial data?** Yes — rough rows and unresolved mentions can later resolve into canonical hosts and identities.
- **Does it support explicit links between events, identities, hosts, artifacts, evidence, and compromise assessments?** Yes — typed `record_links` are first-class, evidence is modeled as a lifecycle-aware envelope, and assessments are incident-scoped records rather than static flags.
- **Does it keep incident data portable while making framework and enrichment overlays optional?** Yes — incident data and reference packs are separated, and missing packs degrade overlays rather than blocking capture.
- **Does it prevent UI/export drift?** Yes — views, overlays, and exports are contract-backed and share a canonical derivation layer or immutable snapshot boundary.
- **Does it provide attribution, version history, and rollback?** Yes — every mutation writes `change_sets`, `change_set_mutations`, and row-centric `record_revisions`, and rollback is an attributed new change.
- **Can it run as a standalone portable deployment?** Yes — the recommended footprint is app + Postgres + MinIO containers with explicit pack and export roots.
- **Does the proposed UI feel closer to Excel than to a ticketing system?** Yes — the main surface is a workbook grid with direct typing, paste, sheet views, and overlays that stay within the workbook mental model.
