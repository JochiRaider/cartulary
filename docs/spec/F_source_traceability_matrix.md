
# Appendix F: Source Traceability Matrix

This appendix is **non-normative**.

It records where the content of the exploratory source artifact was carried into the derived document set. Appendix G preserves the full original artifact verbatim.

## F.1 Level-2 source section coverage

| Source section | Derived target(s) | Coverage note |
| --- | --- | --- |
| `1. Executive summary` | Core 00, Appendix A, Appendix G | Vision and design thesis preserved as rationale; closed requirements extracted into core invariants. |
| `2. Problem framing` | Appendix A, Appendix G | Problem statement, user roles, and adoption rationale preserved as informative context. |
| `3. Design principles and assumptions` | Core 00, Core 01, Core 03, Appendix A, Appendix G | Operating envelope and load-bearing principles promoted; rationale retained in Appendix A. |
| `4. Recommended architecture` | Core 01, Appendix B, Appendix G | Topology, storage, projection, snapshot, and background-job contracts promoted; diagrams and explanatory notes retained. |
| `5. Collaboration and consistency model` | Core 03, Appendix G | Optimistic concurrency, autosave, presence, and same-surface same-field conflict semantics promoted. |
| `6. Domain model and schema strategy` | Core 02, Appendix G | Record types, mention/entity contract, indicator records and observations, provenance, dedupe, merge, notes, assessments, and JSONB discipline promoted. |
| `7. Postgres schema proposal` | Core 02, Appendix C, Appendix G | Schema invariants and history rules promoted; ER diagram and DDL retained as informative reference. |
| `8. Record lifecycle and IR workflow model` | Core 03, Appendix D, Appendix G | Lifecycle, timeline create flow, evidence attach flow, resolution flow, rollback flow, paste/import behavior, and auto-resolution policy promoted. |
| `9. UI concepts focused on preserving the spreadsheet feel` | Core 03, Appendix D, Appendix G | Workbook interaction contract promoted; mockups and explanatory notes retained. |
| `10. UX acceptance criteria` | Core 04, Appendix G | Rewritten as conformance criteria with base and extension profiles, including explicit same-field conflict-resolution checks. |
| `11. Security, auth, and access control` | Core 04, Appendix G | Authentication, authorization, attribution, and trust boundaries promoted. |
| `12. Deployment model` | Core 01, Core 04, Appendix G | Topology, runtime roots, and deployment variants promoted. |
| `13. Trade-offs, risks, and rejected alternatives` | Appendix A, Appendix G | Preserved as informative rationale and risk context. |
| `14. Comparison table` | Appendix A, Appendix G | Preserved as informative comparison context. |
| `15. Recommended MVP and phased roadmap` | Appendix E, Appendix G | Preserved as roadmap context; mixed-status features normalized into extension profiles in core. |
| `16. Open questions` | Core 01, Core 02, Core 04, Appendix E, Appendix G | Preserved as non-normative open issues and editorial decision backlog, with questions closed by later normative decisions promoted into the normative core and marked resolved in Appendix E. |
| `Design sanity check` | Appendix A, Appendix G | Preserved as informative validation summary. |


## F.2 Critical subsection coverage

| Source subsection or topic | Derived target(s) | Coverage note |
| --- | --- | --- |
| `6. Mention, stub, and entity-origin contract` | Core 02 §6, Appendix G | Promoted as an authoritative contract. |
| `6. Required provenance` | Core 02 §7, Appendix G | Promoted as an authoritative provenance contract. |
| `6. Deduplication and auto-upsert rules` | Core 02 §8, Appendix G | Promoted as deterministic matching rules. |
| `6. Merge behavior` | Core 02 §9, Appendix G | Promoted as authoritative merge semantics. |
| `7. Additional schema requirements for rollback granularity` | Core 02 §14-15, Appendix C, Appendix G | Promoted as authoritative rollback substrate. |
| `5. Same-field conflict resolution UX` | Core 03 §3.3, Core 04 §9.6, Appendix E | Promoted as an authoritative same-surface conflict-resolution contract and removed from the remaining MVP open questions. |
| `8. Bulk paste/import from existing spreadsheet or clipboard` | Core 03 §11, Appendix D, Appendix G | Split into base clipboard workflow, bounded file-based import profile, and dedicated imports-module boundary. |
| `8. Auto-resolution policy for typed host/account strings` | Core 03 §12, Appendix D, Appendix G | Promoted as authoritative bounded auto-resolution behavior. |
| `9. Sorting / filtering / grouping` | Core 03 §14, Appendix D, Appendix G | Promoted as authoritative grouping and row-addressing behavior. |
| `9. How denormalized timeline views are composed` | Core 03 §15, Appendix D, Appendix G | Promoted as read/write contract. |
| `11. Reference-pack and export trust boundaries` | Core 04 §4, Appendix G | Promoted as trust-boundary requirements. |
| `16. Open questions on snapshot release controls and generated presentation depth` | Core 01 §10, Core 02 §10.5 and §14.5, Core 04 §2.1, §4.2, and §9.3, Appendix E | Promoted as bounded snapshot release controls, explicit content-class and release-scope contracts, and an evidence-versus-generated-presentation boundary. |
| `16. Open question on restricted evidence visibility` | Core 01 §10.5, Core 02 §10.5, §13, and §14.5, Core 04 §2, §4.2, and §9.3, Appendix E | Closed as export-scoped recipient withholding through snapshot redaction and release controls while keeping live workspace visibility incident-scoped; narrow live sensitive-evidence controls remain future-only. |
| `16. Open question on clipboard paste versus XLSX adoption` | Core 01 §2, Core 03 §11, Core 04 §4.5 and §9.2, Appendix E | Closed as a split between base clipboard UX and bounded file-based import through a dedicated imports module. |
| `16. Open question on dedicated Notes tab and note-capture shape` | Core 01 §7.1 and §7.3, Core 02 §10.1 and §12, Core 04 §9.1 and §9.3, Appendix E | Closed as Notes remaining a built-in workbook sheet in the base profile, backed by the shared artifact model and generic `record_links`, with raw note working material excluded from `external_release` unless explicitly curated. |
| `16. Open question on indicator storage promotion` | Core 01 §7.2, §8.1, and §8.5, Core 02 §2, §7.4, §10.2, §12, and §14, Core 03 §2.2, §7, §9.1, §15, and §16.2, Core 04 §9.1, Appendix B, Appendix C, Appendix D, and Appendix E | Closed as first-class canonical indicators plus source-bound observations and append-only lifecycle intervals, while keeping Indicators as a contract-backed system view over canonical records and preserving raw source text. |
| `16. Open question on assessment vocabulary and confidence model` | Core 01 §7.2 and §8.1-§8.5, Core 02 §10.3 and §14.1-§14.3, Core 03 §2.2 and §16.3, Core 04 §9.1, Appendix C, and Appendix E | Closed as a closed compromise-assessment vocabulary `unknown | suspected | confirmed | disproven | cleared`, append-only incident-scoped assessment history, nullable `confidence_score`, derived `confidence_band`, band-first interactive entry, and strict separation between evidentiary assessment state and operational response actions. |

## F.3 Completeness note

Every source section lands in at least one derived target and in the full-source archive. The normative core is intentionally more compact than the exploratory artifact; rationale, alternatives, mockups, diagrams, roadmap material, and unresolved questions were moved to non-normative appendices so the authoritative core can later be promoted into NLSpecs with minimal churn.
