# Appendix F: Source Traceability Matrix

This appendix is **non-normative**.

It records where the content of the exploratory source artifact was carried into the derived document set. Appendix G preserves the full original artifact verbatim.

## F.1 Level-2 source section coverage

| Source section | Derived target(s) | Coverage note |
| --- | --- | --- |
| `1. Executive summary` | Core 00, Appendix A, Appendix G | Vision and design thesis preserved as rationale; closed requirements extracted into core invariants. |
| `2. Problem framing` | Appendix A, Appendix G | Problem statement, user roles, and adoption rationale preserved as informative context. |
| `3. Design principles and assumptions` | Core 00, Core 01, Core 03, Appendix A, Appendix G | Operating envelope and load-bearing principles promoted; rationale retained in Appendix A. |
| `4. Recommended architecture` | Core 01, Appendix B, Appendix G | Topology, storage, projection, portability, snapshot, and background-job contracts promoted; diagrams and explanatory notes retained. |
| `5. Collaboration and consistency model` | Core 03, Appendix G | Optimistic concurrency, autosave, presence, and same-surface same-field conflict semantics promoted. |
| `6. Domain model and schema strategy` | Core 02, Appendix G | Record types, mention/entity contract, indicator records and observations, provenance, dedupe, merge, notes, analyst-work objects, assessments, and JSONB discipline promoted. |
| `7. Postgres schema proposal` | Core 02, Appendix C, Appendix G | Schema invariants and history rules promoted; ER diagram and DDL retained as informative reference. |
| `8. Record lifecycle and IR workflow model` | Core 03, Appendix D, Appendix G | Lifecycle, including explicit Timeline `capture_state` transition triggers and system-managed workflow semantics, timeline create flow, evidence attach flow, pending-blob timeout and cleanup defaults, resolution flow, rollback flow, paste/import behavior, and auto-resolution policy promoted. |
| `9. UI concepts focused on preserving the spreadsheet feel` | Core 01, Core 03, Appendix D, Appendix G | Workbook interaction contract and authoritative base-profile view-schema registry promoted; mockups and explanatory notes retained. |
| `10. UX acceptance criteria` | Core 04, Appendix G | Rewritten as conformance criteria with base and extension profiles, including explicit same-field conflict-resolution checks. |
| `11. Security, auth, and access control` | Core 04, Appendix G | Authentication, authorization, attribution, and trust boundaries promoted. |
| `12. Deployment model` | Core 01, Core 04, Appendix G | Topology, runtime roots, and deployment variants promoted. |
| `13. Trade-offs, risks, and rejected alternatives` | Appendix A, Appendix G | Preserved as informative rationale and risk context. |
| `14. Comparison table` | Appendix A, Appendix G | Preserved as informative comparison context. |
| `15. Recommended MVP and phased roadmap` | Appendix E, Appendix G | Preserved as roadmap context; mixed-status features normalized into extension profiles in core. |
| `16. Open questions` | Core 00, Core 01, Core 02, Core 03, Core 04, Appendix E, Appendix G | Preserved as non-normative open issues and editorial decision backlog, with questions closed by later normative decisions promoted into the normative core and marked resolved in Appendix E. |
| `Design sanity check` | Appendix A, Appendix G | Preserved as informative validation summary. |


## F.2 Critical subsection coverage

| Source subsection or topic | Derived target(s) | Coverage note |
| --- | --- | --- |
| `6. Mention, stub, and entity-origin contract` | Core 02 §6, Appendix G | Promoted as an authoritative contract. |
| `6. Required provenance` | Core 02 §7, Appendix G | Promoted as an authoritative provenance contract. |
| `6. Deduplication and auto-upsert rules` | Core 02 §8, Appendix G | Promoted as deterministic matching rules. |
| `6. Merge behavior` | Core 01 §3.3.3 and §3.3.5.4, Core 02 §9 and §15.4, Core 03 §5 and §16.2, Core 04 §9.10, Appendix G | Promoted as authoritative merge semantics plus an explicit record-scoped merge route, inspector-only initiation in the base UI, collaboration and rollback hooks, and conformance checks. |
| `7. Additional schema requirements for rollback granularity` | Core 00 §8.1, Core 02 §14-15, Appendix C, Appendix G | Promoted as authoritative rollback substrate, including row-backed whole-row restore boundaries and stable opaque `history_entry_ref` support for single-target reversal without exposing storage-primary-key mutation identifiers. |
| `5. Same-field conflict resolution UX` | Core 03 §3.3, Core 04 §9.6, Appendix E | Promoted as an authoritative same-surface conflict-resolution contract, including plain-text line-based `text_compare_merge` semantics, optional `suggested_merged_value` only for deterministic clean text merges, typed `collection_value_v1` conflict transport for `collection_review` fields, and removal from the remaining MVP open questions. |
| Cross-cutting client/server interface surface for workbook, evidence, and jobs | Core 01 §3.3 and §3.3.10.1, Core 03 §3.3.4, §4.3, §4.3.1, §4.4, §8, and §15, Core 04 §1, §2, §4.5, and §9.10, Appendix E | Closed as one versioned HTTP+JSON API plus a bounded incident-scoped WebSocket stream using stable identifiers, session-based auth, explicit session-lifecycle boundaries and expiry fields, generic envelopes, a shared cursor-pagination transport contract across workbook query plus users, incidents, memberships, saved views, view schemas, and record history, explicit blob-slot and job routes, an authoritative incident-create request and response contract with explicit create-only versus patchable incident fields, server-managed initial values, normalized idempotency comparison, and create-route error codes, record-scoped soft-delete and restore routes with optimistic concurrency and role-gated restore, a record-history route that exposes legal rollback selectors and current tombstone version tokens, a typed rollback request contract for `history_entry`, `change_set`, and `row_restore`, field-key-based mutation that distinguishes direct `value` from versioned `action_payload`, typed `collection_actions_v1` and `collection_value_v1` contracts for `collection_review` fields with stable `item_ref` targets, concrete blob-slot expiry metadata and lease semantics, and a concrete v1 collaboration protocol with a common envelope, typed handshake and presence messages, a replay-only `resume_token` bound to `{session, incident, client_instance}`, replayable-versus-ephemeral delivery classes, bounded resume and reset behavior, application-level heartbeats that do not renew idle session expiry, origin validation for cookie-authenticated browser sockets, preserved client-local unsaved work across auth expiry, per-incident ordering, and lifecycle `record_changed` semantics that surface delete as `remove` and restore as `invalidate`. |
| `8. Bulk paste/import from existing spreadsheet or clipboard` | Core 03 §11, Appendix D, Appendix G | Split into base clipboard workflow, bounded file-based import profile, and dedicated imports-module boundary. |
| `8. Auto-resolution policy for typed host/account strings` | Core 03 §12, Appendix D, Appendix G | Promoted as authoritative bounded auto-resolution behavior. |
| `9. Sorting / filtering / grouping` | Core 01 §3.3.4 and §7.4, Core 03 §14, Appendix D, Appendix G | Promoted as authoritative sort, filter, and grouping behavior, including stable `field_key` addressing, canonical saved-view filter persistence, and a closed view-query `filters[]` wire shape with synthetic predicate keys for filter-only predicates such as full-text search. |
| `9. How denormalized timeline views are composed` | Core 01 §7.4 and Core 03 §15, Appendix D, Appendix G | Promoted as the authoritative Timeline view-schema instance plus Timeline read/write contract. |
| `9. How denormalized entity/evidence views are composed` | Core 01 §7.4, Core 02 §8.2, Core 04 §9.1, Appendix D, Appendix E, and Appendix G | Promoted as the authoritative Hosts, Identities, Evidence, Notes, Indicators, Assessments, Task Requests, and Decisions view-schema instances in the base-profile registry. Hosts, Identities, Evidence, Notes, and Indicators now have exhaustive per-field registries with stable `field_key`, deterministic default sort, explicit write targets or actions, and `conflict_resolution_class`; Hosts and Identities additionally fix exact-match reuse precedence for create-or-upsert, Evidence fixes the explicit blob-action boundary, and Indicators fix a create-only existing-row contract plus an explicit identity-defining immutable field set. |
| `11. Reference-pack and export trust boundaries` | Core 04 §4, Appendix G | Promoted as trust-boundary requirements. |
| `11. Local users, incident roles, and admin surfaces` | Core 01 §3.3.2, §3.3.3, and §3.3.5.1, Core 02 §3 and §14.1, Core 04 §1-§3 and §9.10, Appendix C, Appendix E, and Appendix G | Closed as a split between deployment-local user-account administration and incident-scoped membership writes, with a narrow `deployment_admin` capability, stable internal `user_id` binding, versioned membership updates, last-admin guards, deployment-local administrative audit history, and continued exclusion of login-capable users and memberships from portability bundles. |
| `4. Backup, restore, portability, failure modes` | Core 00 §4.2 and §6, Core 01 §12.3 and §13, Core 04 §4.2 and §9.11, Appendix E, and Appendix G | Closed as a separate Incident Portability Extension Profile with a deterministic logical bundle layout, authoritative-source-only export, checksum-verified import, portable actor descriptors, optional embedded snapshot/reference-pack sections, and no-partial-visibility failure semantics. |
| `16. Open question on minimum disconnected reference-pack bundle and update flow` | Core 01 §5 and §11, Core 02 §4.1, §14.1, and §17, Core 04 §4.1, §5.1, and §9.4, Appendix C, and Appendix E | Closed as a fixed smallest disconnected bundle of three type-registry packs with framework, enrichment, template, and separately distributed `view_contract` packs kept separate, plus offline staged import, explicit activation, retained prior active version, and structured attestation metadata. |
| `16. Open questions on snapshot release controls and generated presentation depth` | Core 01 §10, Core 02 §10.5 and §14.5, Core 04 §2.1, §4.2, and §9.3, Appendix E | Promoted as bounded snapshot release controls, explicit content-class and release-scope contracts, and an evidence-versus-generated-presentation boundary. |
| `16. Open question on restricted evidence visibility` | Core 01 §10.5, Core 02 §10.5, §13, and §14.5, Core 04 §2, §4.2, and §9.3, Appendix E | Closed as export-scoped recipient withholding through snapshot redaction and release controls while keeping live workspace visibility incident-scoped; narrow live sensitive-evidence controls remain future-only. |
| `16. Open question on clipboard paste versus XLSX adoption` | Core 01 §2, Core 03 §11, Core 04 §4.5 and §9.2, Appendix E | Closed as a split between base clipboard UX and bounded file-based import through a dedicated imports module. |
| `16. Open question on dedicated Notes tab and note-capture shape` | Core 01 §7.1, §7.3, and §7.4, Core 02 §10.1 and §12, Core 04 §9.1 and §9.3, Appendix E | Closed as Notes remaining a built-in workbook sheet in the base profile, backed by the shared artifact model and generic `record_links`, with raw note working material excluded from `external_release` unless explicitly curated. |
| `16. Open question on shared/private saved views and startup defaults` | Core 01 §3.3.3 and §3.3.5.2, Core 02 §11.1-§11.2 and §14.1, Core 03 §2.3-§2.4 and §14.3-§14.7, Core 04 §2, §9.1, and §9.10, Appendix E | Closed as a three-scope saved-view model (`private`, `shared`, `system`) plus separate per-user and incident-wide `sheet_ref` startup pointers, ordinary delete semantics, owner/admin mutability, normalized `query_json` keyed by `field_key`, and explicit separation between saved-view discoverability and underlying incident-data authorization. |
| `16. Open question on analyst-work tracking beyond tags and notes` | Core 00 §4.3 and §6, Core 01 §2, §7.2, §7.4, §8.1, §8.5, and §10.3, Core 02 §2, §3, §10.1, §10.4, §12-§14, Core 03 §2.2, §16.4, and §19, Core 04 §2, §4.2, §9.1, and §9.3, Appendix E | Closed as a bounded analyst-work coordination model: first-class `task_request` and `decision` records, ownership as a field on coordination objects, artifact-backed communications, handoff, status-review, and lesson surfaces, no note overloading, no added built-in sheets, and current artifact-backed `hypothesis`. |
| `16. Open question on indicator storage promotion` | Core 01 §7.2, §7.4, §8.1, and §8.5, Core 02 §2, §7.4, §10.2, §12, and §14, Core 03 §2.2, §7, §9.1, §15, and §16.2, Core 04 §9.1, Appendix B, Appendix C, Appendix D, and Appendix E | Closed as first-class canonical indicators plus source-bound observations and append-only lifecycle intervals, while keeping Indicators as a contract-backed system view over canonical records and preserving raw source text. |
| `16. Open question on assessment vocabulary and confidence model` | Core 01 §7.2, §7.4, and §8.1-§8.5, Core 02 §10.3 and §14.1-§14.3, Core 03 §2.2 and §16.3, Core 04 §9.1, Appendix C, and Appendix E | Closed as a closed compromise-assessment vocabulary `unknown | suspected | confirmed | disproven | cleared`, append-only incident-scoped assessment history, nullable `confidence_score`, derived `confidence_band`, band-first interactive entry, and strict separation between evidentiary assessment state and operational response actions. |
| `16. Open question on timeline grouping-key whitelist` | Core 03 §14.3, Core 04 §9.1, and Appendix E | Closed as the existing five-key timeline grouping whitelist being sufficient for GA, with no sixth base-profile grouping key, explicit exclusion of raw `timeline.event_type`, and future widening gated to scalar projection-backed contract values that preserve the interaction envelope. |
| `16. Open question on incident-specific custom metadata promotion` | Core 01 §7.2 and §8.5, Core 02 §4.1-§4.5, §10.4, §13, and §14.1, Core 03 §2.2 and §16.1-§16.4, Core 04 §9.1 and §9.8, and Appendix E | Closed as a deterministic promotion rule: recurrent operational fields on incidents, task requests, hosts, identities, canonical indicators, and evidence records become first-class structured state, while framework overlays, enrichment metadata, presentation flags, and low-frequency governance text remain flexible; multi-valued relationships stay in typed links or child rows; and authoritative custody narrative moves to append-only custody events. |

## F.3 Completeness note

Every source section lands in at least one derived target and in the full-source archive. The normative core is intentionally more compact than the exploratory artifact; rationale, alternatives, mockups, diagrams, roadmap material, and unresolved questions were moved to non-normative appendices so the authoritative core can later be promoted into NLSpecs with minimal churn.

## F.4 Requirement-to-verification navigation

| Requirement ID | Owner section | Profiles | Acceptance criteria |
| --- | --- | --- | --- |
| REQ-00-001 | Core 00 §3 Normative language | base | AC-231 |
| REQ-00-002 | Core 00 §3 Normative language | base | AC-231 |
| REQ-00-003 | Core 00 §4.1 Base profile | base | AC-231 |
| REQ-00-004 | Core 00 §4.2 Extension profiles | import, snapshot_reporting, incident_portability, reference_pack, enterprise_authentication | AC-232..AC-236 |
| REQ-00-005 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-006 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-007 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-008 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-009 | Core 00 §5.2 Requirement and acceptance-criterion traceability | base | AC-231, AC-237 |
| REQ-00-010 | Core 00 §5.2 Requirement and acceptance-criterion traceability | base | AC-231, AC-237 |
| REQ-00-011 | Core 00 §5.2 Requirement and acceptance-criterion traceability | base | AC-231, AC-237 |
| REQ-00-012 | Core 00 §5.2 Requirement and acceptance-criterion traceability | base | AC-231, AC-237 |
| REQ-00-013 | Core 00 §7 Supported operating envelope | base, reference_pack | AC-043..AC-046, AC-231, AC-234 |
| REQ-00-014 | Core 00 §8.1 Core identifiers | base | AC-116, AC-118, AC-123..AC-125, AC-127..AC-129, AC-231 |
| REQ-00-015 | Core 00 §9 Global invariants | base, snapshot_reporting, reference_pack | AC-231, AC-233..AC-234 |
| REQ-00-016 | Core 00 §9.1 Lifecycle state-machine notation | base | AC-107..AC-111, AC-231 |
| REQ-00-017 | Core 00 §9.1 Lifecycle state-machine notation | base | AC-107..AC-111, AC-231 |
| REQ-01-001 | Core 01 §1 Architecture pattern | base | AC-231 |
| REQ-01-002 | Core 01 §1 Architecture pattern | base | AC-231 |
| REQ-01-003 | Core 01 §1 Architecture pattern | base | AC-231 |
| REQ-01-004 | Core 01 §2 Required modules and boundaries | base, snapshot_reporting | AC-027..AC-029, AC-046, AC-063..AC-067, AC-231, AC-233 |
| REQ-01-005 | Core 01 §2 Required modules and boundaries | base | AC-027..AC-029, AC-046, AC-063..AC-067, AC-231 |
| REQ-01-006 | Core 01 §2 Required modules and boundaries | base, import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-231..AC-232 |
| REQ-01-007 | Core 01 §2 Required modules and boundaries | base, import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-231..AC-232 |
| REQ-01-008 | Core 01 §2 Required modules and boundaries | base, import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-231..AC-232 |
| REQ-01-009 | Core 01 §2 Required modules and boundaries | base, import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-231..AC-232 |
| REQ-01-010 | Core 01 §2.1 Phase 2 Workbook Import Assistant | import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-232 |
| REQ-01-011 | Core 01 §2.1 Phase 2 Workbook Import Assistant | import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-232 |
| REQ-01-012 | Core 01 §2.1 Phase 2 Workbook Import Assistant | import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-232 |
| REQ-01-013 | Core 01 §2.1 Phase 2 Workbook Import Assistant | import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-232 |
| REQ-01-014 | Core 01 §2.1 Phase 2 Workbook Import Assistant | import | AC-027..AC-029, AC-046, AC-063..AC-067, AC-232 |
| REQ-01-015 | Core 01 §3.1 Browser client | base | AC-001, AC-003..AC-005, AC-043..AC-045, AC-047, AC-231 |
| REQ-01-016 | Core 01 §3.1 Browser client | base | AC-001, AC-003..AC-005, AC-043..AC-045, AC-047, AC-231 |
| REQ-01-017 | Core 01 §3.1 Browser client | base | AC-001, AC-003..AC-005, AC-043..AC-045, AC-047, AC-231 |
| REQ-01-018 | Core 01 §3.2 Application server | base, snapshot_reporting, reference_pack | AC-046, AC-129, AC-231, AC-233..AC-234 |
| REQ-01-019 | Core 01 §3.3 Public HTTP and WebSocket interface contract | base | AC-124..AC-129, AC-131, AC-135, AC-231 |
| REQ-01-020 | Core 01 §3.3.1 Versioning and compatibility | base | AC-124..AC-125, AC-127, AC-131, AC-135, AC-231 |
| REQ-01-021 | Core 01 §3.3.1 Versioning and compatibility | base | AC-124..AC-125, AC-127, AC-131, AC-135, AC-231 |
| REQ-01-022 | Core 01 §3.3.1 Versioning and compatibility | base | AC-124..AC-125, AC-127, AC-131, AC-135, AC-231 |
| REQ-01-023 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-024 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-025 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-026 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-027 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-028 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-029 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-030 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-031 | Core 01 §3.3.2 Session and authentication routes | base, enterprise_authentication | AC-123, AC-130, AC-156..AC-162, AC-231, AC-235 |
| REQ-01-032 | Core 01 §3.3.3 Route families | base | AC-175..AC-180, AC-186..AC-187, AC-231 |
| REQ-01-033 | Core 01 §3.3.3 Route families | base, snapshot_reporting, reference_pack | AC-175..AC-180, AC-186..AC-187, AC-231, AC-233..AC-234 |
| REQ-01-034 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-035 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-238..AC-240, AC-243 |
| REQ-01-036 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-238, AC-241, AC-243 |
| REQ-01-037 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-243 |
| REQ-01-038 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-039 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-040 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-041 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-042 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-043 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-044 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-045 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-046 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-047 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-048 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-049 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-050 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-051 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-052 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-053 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-054 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-055 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-056 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-215, AC-231 |
| REQ-01-057 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-058 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-059 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-060 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-061 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-062 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-063 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-064 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-065 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-066 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-067 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-068 | Core 01 §3.3.5 Mutation contract | base, snapshot_reporting | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-233 |
| REQ-01-069 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-070 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-071 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-072 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-073 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-074 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-075 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-076 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-077 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-078 | Core 01 §3.3.5 Mutation contract | base, snapshot_reporting | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-233 |
| REQ-01-079 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-080 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-081 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-082 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-083 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-084 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-085 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-086 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-087 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-088 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-089 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-090 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-091 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-092 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-093 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-094 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-095 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-096 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-097 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-098 | Core 01 §3.3.5.0 Rollback contract | base, snapshot_reporting | AC-215..AC-218, AC-231, AC-233 |
| REQ-01-099 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-100 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-101 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-102 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-103 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-104 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-105 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-106 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-107 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-108 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-109 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-110 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-111 | Core 01 §3.3.5.0 Rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-112 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-113 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-114 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-115 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-116 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-117 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-127, AC-175..AC-180, AC-231 |
| REQ-01-118 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-127, AC-175..AC-180, AC-231 |
| REQ-01-119 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-120 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-121 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-122 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-123 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-124 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-125 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-126 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-127 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-128 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-129 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-127, AC-175..AC-180, AC-231 |
| REQ-01-130 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-131 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-132 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-133 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-134 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-135 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-136 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-137 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-138 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-139 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-140 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-141 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-142 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-143 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-144 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-127, AC-146..AC-153, AC-231 |
| REQ-01-145 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-146 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-147 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-148 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-149 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-150 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-151 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-152 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-153 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-154 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-130, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-155 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-156 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-157 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-158 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-159 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-160 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-161 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-162 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-163 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-164 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-165 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-166 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-167 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-168 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-127, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-169 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-170 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-171 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-172 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-173 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-174 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-175 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-176 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-177 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-178 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-179 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-180 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-181 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-182 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-183 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-184 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-185 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-186 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-187 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-188 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-189 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-190 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-191 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-192 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-193 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-194 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-195 | Core 01 §3.3.5.4 Entity-merge contract | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-01-196 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-197 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-198 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-199 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-200 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-201 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-202 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-203 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-204 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-205 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-206 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-207 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-208 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-209 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-210 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-211 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-212 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-213 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-214 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-215 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-216 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-217 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-218 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-219 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-220 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-221 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-222 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-223 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-224 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-225 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-226 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-227 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-01-228 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-229 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-230 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-231 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-232 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-233 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-234 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231, AC-239..AC-240 |
| REQ-01-235 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-236 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-237 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-187, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-238 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231, AC-239..AC-240 |
| REQ-01-239 | Core 01 §3.3.6 Success and error envelopes | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-240 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238..AC-240 |
| REQ-01-241 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-239 |
| REQ-01-242 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238..AC-239, AC-241..AC-242 |
| REQ-01-243 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-244 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-245 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-246 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-247 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-248 | Core 01 §3.3.9 Background-job routes | base | AC-046, AC-129, AC-231 |
| REQ-01-249 | Core 01 §3.3.9 Background-job routes | base | AC-046, AC-129, AC-231 |
| REQ-01-250 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-251 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-252 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-253 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-254 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-255 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-256 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-257 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-258 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-259 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-260 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-261 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-262 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-263 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-264 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-265 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-266 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-267 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-268 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-269 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-270 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-271 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-272 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-273 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-274 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-275 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-276 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-277 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-278 | Core 01 §4.1 Postgres | base, snapshot_reporting, reference_pack | AC-231, AC-233..AC-234 |
| REQ-01-279 | Core 01 §4.2 Object storage | base, snapshot_reporting | AC-231, AC-233 |
| REQ-01-280 | Core 01 §4.3 Storage exclusions | base | AC-231 |
| REQ-01-281 | Core 01 §4.3 Storage exclusions | base | AC-231 |
| REQ-01-282 | Core 01 §5 Incident data versus reference packs | base, reference_pack | AC-034, AC-092, AC-096, AC-231, AC-234 |
| REQ-01-283 | Core 01 §5 Incident data versus reference packs | base, reference_pack | AC-034, AC-092, AC-096, AC-231, AC-234 |
| REQ-01-284 | Core 01 §5 Incident data versus reference packs | base, reference_pack | AC-034, AC-092, AC-096, AC-231, AC-234 |
| REQ-01-285 | Core 01 §6 View contracts | base | AC-116..AC-120, AC-124..AC-125, AC-231 |
| REQ-01-286 | Core 01 §6 View contracts | base, reference_pack | AC-116..AC-120, AC-124..AC-125, AC-231, AC-234 |
| REQ-01-287 | Core 01 §6 View contracts | base | AC-116..AC-120, AC-124..AC-125, AC-231 |
| REQ-01-288 | Core 01 §6 View contracts | base | AC-116..AC-120, AC-124..AC-125, AC-127, AC-231 |
| REQ-01-289 | Core 01 §6 View contracts | base | AC-116..AC-120, AC-124..AC-125, AC-231 |
| REQ-01-290 | Core 01 §6 View contracts | base | AC-116..AC-120, AC-124..AC-125, AC-231 |
| REQ-01-291 | Core 01 §7.1 Built-in sheets | base | AC-015, AC-116, AC-231 |
| REQ-01-292 | Core 01 §7.1 Built-in sheets | base | AC-015, AC-116, AC-231 |
| REQ-01-293 | Core 01 §7.1 Built-in sheets | base | AC-015, AC-116, AC-231 |
| REQ-01-294 | Core 01 §7.1 Built-in sheets | base | AC-015, AC-116, AC-231 |
| REQ-01-295 | Core 01 §7.1 Built-in sheets | base | AC-015, AC-116, AC-231 |
| REQ-01-296 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-297 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-298 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-299 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-300 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-301 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-302 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-303 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-304 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-305 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-306 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-307 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-116..AC-122, AC-124..AC-125, AC-231 |
| REQ-01-308 | Core 01 §7.4 Authoritative base-profile view schema registry | base, reference_pack | AC-116..AC-122, AC-124..AC-125, AC-231, AC-234 |
| REQ-01-309 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-116..AC-122, AC-124..AC-125, AC-231 |
| REQ-01-310 | Core 01 §7.4 Authoritative base-profile view schema registry | base, reference_pack | AC-116..AC-122, AC-124..AC-125, AC-231, AC-234 |
| REQ-01-311 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-116..AC-122, AC-124..AC-125, AC-231 |
| REQ-01-312 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-313 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-314 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-315 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-316 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-317 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-318 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-319 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-320 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-321 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-322 | Core 01 §7.4.1 `cartulary.view.timeline.v1` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-323 | Core 01 §7.4.2 `cartulary.view.hosts.v1` | base | AC-097, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-324 | Core 01 §7.4.2 `cartulary.view.hosts.v1` | base | AC-097, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-325 | Core 01 §7.4.2 `cartulary.view.hosts.v1` | base | AC-097, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-326 | Core 01 §7.4.3 `cartulary.view.identities.v1` | base | AC-098, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-327 | Core 01 §7.4.3 `cartulary.view.identities.v1` | base | AC-098, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-328 | Core 01 §7.4.4 `cartulary.view.evidence.v1` | base | AC-100, AC-118, AC-124..AC-125, AC-128, AC-231 |
| REQ-01-329 | Core 01 §7.4.5 `cartulary.view.notes.v1` | base | AC-068..AC-070, AC-112, AC-118, AC-124..AC-125, AC-185, AC-231 |
| REQ-01-330 | Core 01 §7.4.5 `cartulary.view.notes.v1` | base | AC-068..AC-070, AC-112, AC-118, AC-124..AC-125, AC-185, AC-231 |
| REQ-01-331 | Core 01 §7.4.6 `cartulary.view.indicators.v1` | base | AC-017, AC-072..AC-079, AC-118, AC-122, AC-124, AC-231 |
| REQ-01-332 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-333 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-334 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-335 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-336 | Core 01 §7.4.8 `cartulary.view.task_requests.v1` | base | AC-085, AC-118, AC-124, AC-137..AC-140, AC-145, AC-231 |
| REQ-01-337 | Core 01 §7.4.8 `cartulary.view.task_requests.v1` | base | AC-085, AC-118, AC-124, AC-137..AC-140, AC-145, AC-231 |
| REQ-01-338 | Core 01 §7.4.8 `cartulary.view.task_requests.v1` | base | AC-085, AC-118, AC-124, AC-137..AC-140, AC-145, AC-231 |
| REQ-01-339 | Core 01 §7.4.9 `cartulary.view.decisions.v1` | base | AC-086, AC-118, AC-124, AC-141..AC-145, AC-231 |
| REQ-01-340 | Core 01 §7.4.9 `cartulary.view.decisions.v1` | base | AC-086, AC-118, AC-124, AC-141..AC-145, AC-231 |
| REQ-01-341 | Core 01 §7.4.9 `cartulary.view.decisions.v1` | base | AC-086, AC-118, AC-124, AC-141..AC-145, AC-231 |
| REQ-01-342 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-343 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-344 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-345 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-346 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-347 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-348 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-349 | Core 01 §8.2 Projection-row identity | base | AC-231 |
| REQ-01-350 | Core 01 §8.2 Projection-row identity | base | AC-231 |
| REQ-01-351 | Core 01 §8.3 Projection maintenance | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-352 | Core 01 §8.3 Projection maintenance | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-353 | Core 01 §8.3 Projection maintenance | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-354 | Core 01 §8.4 Projection corruption | base | AC-231 |
| REQ-01-355 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-356 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-357 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-358 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-359 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-360 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-361 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-362 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-363 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-364 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-365 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-366 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-367 | Core 01 §9 Canonical derivation layer | base | AC-032, AC-231 |
| REQ-01-368 | Core 01 §9 Canonical derivation layer | base, snapshot_reporting | AC-032, AC-231, AC-233 |
| REQ-01-369 | Core 01 §10.1 Extension boundary | snapshot_reporting | AC-030, AC-046, AC-233 |
| REQ-01-370 | Core 01 §10.2 Snapshot semantics | snapshot_reporting | AC-030..AC-032, AC-056..AC-058, AC-233 |
| REQ-01-371 | Core 01 §10.2 Snapshot semantics | snapshot_reporting | AC-030..AC-032, AC-056..AC-058, AC-233 |
| REQ-01-372 | Core 01 §10.2 Snapshot semantics | snapshot_reporting | AC-030..AC-032, AC-056..AC-058, AC-233 |
| REQ-01-373 | Core 01 §10.2 Snapshot semantics | snapshot_reporting | AC-030..AC-032, AC-056..AC-058, AC-233 |
| REQ-01-374 | Core 01 §10.2.1 Rendered artifact lifecycle | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-01-375 | Core 01 §10.2.1 Rendered artifact lifecycle | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-01-376 | Core 01 §10.2.1 Rendered artifact lifecycle | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-01-377 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233 |
| REQ-01-378 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233 |
| REQ-01-379 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233 |
| REQ-01-380 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233 |
| REQ-01-381 | Core 01 §10.4 Template packs and rendering contract | snapshot_reporting | AC-058, AC-091, AC-233 |
| REQ-01-382 | Core 01 §10.4 Template packs and rendering contract | snapshot_reporting | AC-058, AC-091, AC-233 |
| REQ-01-383 | Core 01 §10.4 Template packs and rendering contract | snapshot_reporting | AC-058, AC-091, AC-233 |
| REQ-01-384 | Core 01 §10.4 Template packs and rendering contract | snapshot_reporting | AC-058, AC-091, AC-233 |
| REQ-01-385 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-386 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-387 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-388 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-389 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-390 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-391 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-392 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-393 | Core 01 §10.5 Redaction profiles and manifests | snapshot_reporting | AC-057, AC-060, AC-113..AC-115, AC-233 |
| REQ-01-394 | Core 01 §10.6 Output forms and generated-presentation boundary | snapshot_reporting | AC-031, AC-061..AC-062, AC-233 |
| REQ-01-395 | Core 01 §10.6 Output forms and generated-presentation boundary | snapshot_reporting | AC-031, AC-061..AC-062, AC-233 |
| REQ-01-396 | Core 01 §10.6 Output forms and generated-presentation boundary | snapshot_reporting | AC-031, AC-061..AC-062, AC-233 |
| REQ-01-397 | Core 01 §10.6 Output forms and generated-presentation boundary | snapshot_reporting | AC-031, AC-061..AC-062, AC-233 |
| REQ-01-398 | Core 01 §10.7 Self-contained outputs | snapshot_reporting | AC-031, AC-233 |
| REQ-01-399 | Core 01 §11.1 Extension boundary | reference_pack | AC-033..AC-035, AC-234 |
| REQ-01-400 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-401 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-402 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-403 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-404 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-405 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-406 | Core 01 §11.2 Minimum disconnected bundle | reference_pack | AC-092, AC-234 |
| REQ-01-407 | Core 01 §11.3 Offline import, update, and activation flow | reference_pack | AC-033, AC-093..AC-094, AC-096, AC-234 |
| REQ-01-408 | Core 01 §11.3 Offline import, update, and activation flow | reference_pack | AC-033, AC-093..AC-094, AC-096, AC-234 |
| REQ-01-409 | Core 01 §11.3.1 Linked reference-pack lifecycle machines | reference_pack | AC-033, AC-035, AC-093..AC-096, AC-234 |
| REQ-01-410 | Core 01 §11.3.1 Linked reference-pack lifecycle machines | reference_pack | AC-033, AC-035, AC-093..AC-096, AC-234 |
| REQ-01-411 | Core 01 §11.3.1 Linked reference-pack lifecycle machines | reference_pack | AC-033, AC-035, AC-093..AC-096, AC-234 |
| REQ-01-412 | Core 01 §11.3.1 Linked reference-pack lifecycle machines | reference_pack | AC-033, AC-035, AC-093..AC-096, AC-234 |
| REQ-01-413 | Core 01 §11.3.1 Linked reference-pack lifecycle machines | reference_pack | AC-033, AC-035, AC-093..AC-096, AC-234 |
| REQ-01-414 | Core 01 §11.4 Verification and attestation | reference_pack | AC-035, AC-094..AC-095, AC-234 |
| REQ-01-415 | Core 01 §11.4 Verification and attestation | reference_pack | AC-035, AC-094..AC-095, AC-234 |
| REQ-01-416 | Core 01 §11.4 Verification and attestation | reference_pack | AC-035, AC-094..AC-095, AC-234 |
| REQ-01-417 | Core 01 §11.4 Verification and attestation | reference_pack | AC-035, AC-094..AC-095, AC-234 |
| REQ-01-418 | Core 01 §11.4 Verification and attestation | reference_pack | AC-035, AC-094..AC-095, AC-234 |
| REQ-01-419 | Core 01 §11.4.1 Activation safety and observability | reference_pack | AC-034..AC-035, AC-095, AC-234 |
| REQ-01-420 | Core 01 §11.4.1 Activation safety and observability | reference_pack | AC-034..AC-035, AC-095, AC-234 |
| REQ-01-421 | Core 01 §11.4.1 Activation safety and observability | reference_pack | AC-034..AC-035, AC-095, AC-234 |
| REQ-01-422 | Core 01 §11.5 Degradation behavior | reference_pack | AC-034, AC-234 |
| REQ-01-423 | Core 01 §12.2 Restore | base | AC-231 |
| REQ-01-424 | Core 01 §12.2 Restore | base | AC-231 |
| REQ-01-425 | Core 01 §12.3 Incident portability | incident_portability | AC-164..AC-169, AC-236 |
| REQ-01-426 | Core 01 §12.3 Incident portability | incident_portability | AC-164..AC-169, AC-236 |
| REQ-01-427 | Core 01 §12.3.1 Logical bundle contract | incident_portability | AC-164, AC-166, AC-169, AC-236 |
| REQ-01-428 | Core 01 §12.3.1 Logical bundle contract | incident_portability | AC-164, AC-166, AC-169, AC-236 |
| REQ-01-429 | Core 01 §12.3.1 Logical bundle contract | incident_portability | AC-164, AC-166, AC-169, AC-236 |
| REQ-01-430 | Core 01 §12.3.1 Logical bundle contract | incident_portability | AC-164, AC-166, AC-169, AC-236 |
| REQ-01-431 | Core 01 §12.3.2 Authoritative-state boundary | incident_portability | AC-164, AC-167, AC-236 |
| REQ-01-432 | Core 01 §12.3.2 Authoritative-state boundary | incident_portability | AC-164, AC-167, AC-236 |
| REQ-01-433 | Core 01 §12.3.3 Manifest and integrity contract | incident_portability | AC-164, AC-166, AC-236 |
| REQ-01-434 | Core 01 §12.3.3 Manifest and integrity contract | incident_portability | AC-164, AC-166, AC-236 |
| REQ-01-435 | Core 01 §12.3.3 Manifest and integrity contract | incident_portability | AC-164, AC-166, AC-236 |
| REQ-01-436 | Core 01 §12.3.3 Manifest and integrity contract | incident_portability | AC-164, AC-166, AC-236 |
| REQ-01-437 | Core 01 §12.3.3 Manifest and integrity contract | incident_portability | AC-164, AC-166, AC-236 |
| REQ-01-438 | Core 01 §12.3.3 Manifest and integrity contract | incident_portability | AC-164, AC-166, AC-236 |
| REQ-01-439 | Core 01 §12.3.4 Structured formats and deterministic serialization | incident_portability | AC-164..AC-165, AC-236 |
| REQ-01-440 | Core 01 §12.3.4 Structured formats and deterministic serialization | incident_portability | AC-164..AC-165, AC-236 |
| REQ-01-441 | Core 01 §12.3.4 Structured formats and deterministic serialization | incident_portability | AC-164..AC-165, AC-236 |
| REQ-01-442 | Core 01 §12.3.4 Structured formats and deterministic serialization | incident_portability | AC-164..AC-165, AC-236 |
| REQ-01-443 | Core 01 §12.3.5 Portable actors and optional embedded sections | incident_portability | AC-167..AC-168, AC-236 |
| REQ-01-444 | Core 01 §12.3.5 Portable actors and optional embedded sections | incident_portability | AC-167..AC-168, AC-236 |
| REQ-01-445 | Core 01 §12.3.5 Portable actors and optional embedded sections | incident_portability | AC-167..AC-168, AC-236 |
| REQ-01-446 | Core 01 §12.3.5 Portable actors and optional embedded sections | incident_portability | AC-167..AC-168, AC-236 |
| REQ-01-447 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236 |
| REQ-01-448 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236 |
| REQ-01-449 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236 |
| REQ-01-450 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236 |
| REQ-01-451 | Core 01 §12.4 Failure handling | base | AC-166, AC-231 |
| REQ-01-452 | Core 01 §13 Long-running operations and background jobs | base, snapshot_reporting, reference_pack | AC-030, AC-033, AC-046, AC-129, AC-169, AC-231, AC-233..AC-234 |
| REQ-01-453 | Core 01 §13 Long-running operations and background jobs | base | AC-030, AC-033, AC-046, AC-129, AC-169, AC-231 |
| REQ-01-454 | Core 01 §13 Long-running operations and background jobs | base | AC-030, AC-033, AC-046, AC-129, AC-169, AC-231 |
| REQ-01-455 | Core 01 §14 Runtime roots and packaging | base, reference_pack | AC-051, AC-055, AC-169, AC-231, AC-234 |
| REQ-01-456 | Core 01 §14 Runtime roots and packaging | base | AC-051, AC-055, AC-169, AC-231 |
| REQ-01-457 | Core 01 §15 Architecture invariants | base, import, snapshot_reporting, incident_portability | AC-231..AC-233, AC-236 |
| REQ-02-001 | Core 02 §1 Domain-model goals | base, reference_pack | AC-231, AC-234 |
| REQ-02-002 | Core 02 §1 Domain-model goals | base | AC-231 |
| REQ-02-003 | Core 02 §2 Core record types | base, reference_pack | AC-231, AC-234 |
| REQ-02-004 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-005 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-006 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-007 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-008 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-009 | Core 02 §4.1 Normalized data | base, snapshot_reporting, reference_pack | AC-097..AC-101, AC-231, AC-233..AC-234 |
| REQ-02-010 | Core 02 §4.3 JSONB discipline | base | AC-097..AC-101, AC-231 |
| REQ-02-011 | Core 02 §4.3 JSONB discipline | base | AC-097..AC-101, AC-231 |
| REQ-02-012 | Core 02 §4.4 Promotion rule for recurrent incident-specific fields | base | AC-097..AC-101, AC-231 |
| REQ-02-013 | Core 02 §4.4 Promotion rule for recurrent incident-specific fields | base | AC-097..AC-101, AC-231 |
| REQ-02-014 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-015 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-016 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-017 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-018 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-019 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-020 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-021 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-022 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-023 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-024 | Core 02 §5 Partial and uncertain data | base | AC-231 |
| REQ-02-025 | Core 02 §5 Partial and uncertain data | base | AC-231 |
| REQ-02-026 | Core 02 §6.1 Separation invariant | base | AC-019, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-027 | Core 02 §6.1 Separation invariant | base | AC-019, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-028 | Core 02 §6.2 Binding modes | base | AC-118, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-029 | Core 02 §6.2 Binding modes | base | AC-118, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-030 | Core 02 §6.3 Required binding behavior by context | base | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
| REQ-02-031 | Core 02 §6.3 Required binding behavior by context | base | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
| REQ-02-032 | Core 02 §6.3 Required binding behavior by context | base, import | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231..AC-232 |
| REQ-02-033 | Core 02 §6.3 Required binding behavior by context | base | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
| REQ-02-034 | Core 02 §6.3 Required binding behavior by context | base | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
| REQ-02-035 | Core 02 §6.3 Required binding behavior by context | base | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
| REQ-02-036 | Core 02 §6.3 Required binding behavior by context | base | AC-019..AC-020, AC-022, AC-028..AC-029, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
| REQ-02-037 | Core 02 §6.4 Suggestion boundary | base | AC-020, AC-028..AC-029, AC-231 |
| REQ-02-038 | Core 02 §6.4 Suggestion boundary | base | AC-020, AC-028..AC-029, AC-231 |
| REQ-02-039 | Core 02 §6.5 Entity-mention lifecycle | base | AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-02-040 | Core 02 §6.5 Entity-mention lifecycle | base | AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-02-041 | Core 02 §6.5 Entity-mention lifecycle | base | AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-02-042 | Core 02 §7.1 Mention provenance | base | AC-021, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-043 | Core 02 §7.1 Mention provenance | base | AC-021, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-044 | Core 02 §7.1 Mention provenance | base | AC-021, AC-188..AC-190, AC-221..AC-223, AC-231 |
| REQ-02-045 | Core 02 §7.2 File-based import provenance | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-046 | Core 02 §7.2 File-based import provenance | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-047 | Core 02 §7.2 File-based import provenance | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-048 | Core 02 §7.2.1 Import-session, unit, and mapping identity | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-049 | Core 02 §7.2.1 Import-session, unit, and mapping identity | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-050 | Core 02 §7.2.1 Import-session, unit, and mapping identity | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-051 | Core 02 §7.2.1 Import-session, unit, and mapping identity | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-052 | Core 02 §7.2.1 Import-session, unit, and mapping identity | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-053 | Core 02 §7.2.1 Import-session, unit, and mapping identity | import | AC-027..AC-029, AC-063..AC-067, AC-232 |
| REQ-02-054 | Core 02 §7.3 Entity provenance | base | AC-186, AC-209, AC-231 |
| REQ-02-055 | Core 02 §7.3 Entity provenance | base | AC-186, AC-209, AC-231 |
| REQ-02-056 | Core 02 §7.4 Indicator observation provenance | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-057 | Core 02 §7.4 Indicator observation provenance | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-058 | Core 02 §8.1 Mention deduplication | base | AC-021, AC-028, AC-188, AC-231 |
| REQ-02-059 | Core 02 §8.1 Mention deduplication | base | AC-021, AC-028, AC-188, AC-231 |
| REQ-02-060 | Core 02 §8.2 Entity exact-match precedence | base | AC-022, AC-028, AC-186..AC-187, AC-231 |
| REQ-02-061 | Core 02 §8.2 Entity exact-match precedence | base | AC-022, AC-028, AC-186..AC-187, AC-231 |
| REQ-02-062 | Core 02 §8.3 Suggestion boundary | base | AC-028..AC-029, AC-231 |
| REQ-02-063 | Core 02 §8.3 Suggestion boundary | base | AC-028..AC-029, AC-231 |
| REQ-02-064 | Core 02 §9 Merge behavior | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-02-065 | Core 02 §9 Merge behavior | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-02-066 | Core 02 §9 Merge behavior | base | AC-023, AC-186..AC-187, AC-209, AC-231 |
| REQ-02-067 | Core 02 §10.1 Notes | base | AC-068, AC-070, AC-089, AC-112, AC-185, AC-231 |
| REQ-02-068 | Core 02 §10.1 Notes | base | AC-068, AC-070, AC-089, AC-112, AC-185, AC-231 |
| REQ-02-069 | Core 02 §10.1 Notes | base | AC-068, AC-070, AC-089, AC-112, AC-185, AC-231 |
| REQ-02-070 | Core 02 §10.1 Notes | base | AC-068, AC-070, AC-089, AC-112, AC-185, AC-231 |
| REQ-02-071 | Core 02 §10.1 Notes | base | AC-068, AC-070, AC-089, AC-112, AC-185, AC-231 |
| REQ-02-072 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-073 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-074 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-075 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-076 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-077 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-078 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-079 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-080 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-081 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-082 | Core 02 §10.2 Indicator contract | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-083 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-084 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-085 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-086 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-087 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-088 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-089 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-090 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-091 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-092 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-093 | Core 02 §10.3 Compromise assessments | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-02-094 | Core 02 §10.4 Analyst-work tracking | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-02-095 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-096 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-097 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-098 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-099 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-100 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-101 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-102 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-103 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-104 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-105 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-106 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-107 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-108 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-109 | Core 02 §10.4.1 `task_request` record type | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-110 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-111 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-112 | Core 02 §10.4.2 `decision` record type | base, snapshot_reporting | AC-086, AC-141..AC-145, AC-231, AC-233 |
| REQ-02-113 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-114 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-115 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-116 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-117 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-118 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-119 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-120 | Core 02 §10.4.3 Ownership and hot-path boundary | base | AC-085..AC-086, AC-090, AC-231 |
| REQ-02-121 | Core 02 §10.4.3 Ownership and hot-path boundary | base | AC-085..AC-086, AC-090, AC-231 |
| REQ-02-122 | Core 02 §10.4.3 Ownership and hot-path boundary | base | AC-085..AC-086, AC-090, AC-231 |
| REQ-02-123 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-124 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-125 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-126 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-127 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-128 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-129 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-130 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-131 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-132 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-133 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231 |
| REQ-02-134 | Core 02 §10.4.5 Hypothesis boundary | base | AC-089, AC-231 |
| REQ-02-135 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231 |
| REQ-02-136 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231 |
| REQ-02-137 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231 |
| REQ-02-138 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231 |
| REQ-02-139 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-140 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-141 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-142 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-143 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-144 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-145 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-146 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-147 | Core 02 §11 Type registries and view contracts | base | AC-146..AC-153, AC-231 |
| REQ-02-148 | Core 02 §11 Type registries and view contracts | base | AC-146..AC-153, AC-231 |
| REQ-02-149 | Core 02 §11 Type registries and view contracts | base | AC-146..AC-153, AC-231 |
| REQ-02-150 | Core 02 §11 Type registries and view contracts | base | AC-146..AC-153, AC-231 |
| REQ-02-151 | Core 02 §11 Type registries and view contracts | base | AC-146..AC-153, AC-231 |
| REQ-02-152 | Core 02 §11.1 Saved-view contract | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-02-153 | Core 02 §11.1 Saved-view contract | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-02-154 | Core 02 §11.1 Saved-view contract | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-02-155 | Core 02 §11.1 Saved-view contract | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-02-156 | Core 02 §11.1 Saved-view contract | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-02-157 | Core 02 §11.1 Saved-view contract | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-02-158 | Core 02 §11.2 Workbook startup preference objects | base | AC-150, AC-153, AC-231 |
| REQ-02-159 | Core 02 §11.2 Workbook startup preference objects | base | AC-150, AC-153, AC-231 |
| REQ-02-160 | Core 02 §11.2 Workbook startup preference objects | base | AC-150, AC-153, AC-231 |
| REQ-02-161 | Core 02 §11.2 Workbook startup preference objects | base | AC-150, AC-153, AC-231 |
| REQ-02-162 | Core 02 §11.2 Workbook startup preference objects | base | AC-150, AC-153, AC-231 |
| REQ-02-163 | Core 02 §12 Typed relationships | base | AC-205..AC-210, AC-231 |
| REQ-02-164 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-165 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-166 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-167 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-168 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-169 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-170 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-171 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-172 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-173 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-174 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-175 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-176 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-177 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-178 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-179 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-180 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-181 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-182 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-183 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-184 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-185 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-186 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-187 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-188 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-189 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-190 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-191 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-192 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-193 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-194 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-195 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-196 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-197 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-198 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-199 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-200 | Core 02 §13 Evidence and object metadata | base, snapshot_reporting | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-233 |
| REQ-02-201 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-202 | Core 02 §14.1 Mention, indicator, provenance, and coordination fields | base, import, snapshot_reporting, reference_pack | AC-017..AC-018, AC-072..AC-075, AC-118, AC-128, AC-154..AC-155, AC-188..AC-190, AC-200..AC-204, AC-231..AC-234 |
| REQ-02-203 | Core 02 §14.1 Mention, indicator, provenance, and coordination fields | base | AC-017..AC-018, AC-072..AC-075, AC-118, AC-128, AC-154..AC-155, AC-188..AC-190, AC-200..AC-204, AC-231 |
| REQ-02-204 | Core 02 §14.1 Mention, indicator, provenance, and coordination fields | base, incident_portability | AC-017..AC-018, AC-072..AC-075, AC-118, AC-128, AC-154..AC-155, AC-188..AC-190, AC-200..AC-204, AC-231, AC-236 |
| REQ-02-205 | Core 02 §14.2 Rollback granularity substrate | base | AC-215..AC-218, AC-231 |
| REQ-02-206 | Core 02 §14.2 Rollback granularity substrate | base, snapshot_reporting | AC-215..AC-218, AC-231, AC-233 |
| REQ-02-207 | Core 02 §14.2 Rollback granularity substrate | base, snapshot_reporting | AC-215..AC-218, AC-231, AC-233 |
| REQ-02-208 | Core 02 §14.3 Mutation targets | base | AC-125, AC-200..AC-208, AC-231 |
| REQ-02-209 | Core 02 §14.3 Mutation targets | base | AC-125, AC-200..AC-208, AC-231 |
| REQ-02-210 | Core 02 §14.4 Soft delete | base | AC-181..AC-183, AC-231 |
| REQ-02-211 | Core 02 §14.5 Snapshot and reporting extension fields | snapshot_reporting | AC-057, AC-060..AC-062, AC-113..AC-115, AC-233 |
| REQ-02-212 | Core 02 §15.1 Attribution unit | base | AC-215..AC-217, AC-231 |
| REQ-02-213 | Core 02 §15.1 Attribution unit | base | AC-215..AC-217, AC-231 |
| REQ-02-214 | Core 02 §15.2 Mutation-entry unit | base | AC-215..AC-217, AC-231 |
| REQ-02-215 | Core 02 §15.2 Mutation-entry unit | base | AC-215..AC-217, AC-231 |
| REQ-02-216 | Core 02 §15.2 Mutation-entry unit | base | AC-215..AC-217, AC-231 |
| REQ-02-217 | Core 02 §15.3 Reconstruction requirement | base, snapshot_reporting | AC-215, AC-217, AC-231, AC-233 |
| REQ-02-218 | Core 02 §15.3 Reconstruction requirement | base | AC-215, AC-217, AC-231 |
| REQ-02-219 | Core 02 §15.4 Entity-merge history | base | AC-023, AC-186, AC-209, AC-217, AC-231 |
| REQ-02-220 | Core 02 §15.4 Entity-merge history | base | AC-023, AC-186, AC-209, AC-217, AC-231 |
| REQ-02-221 | Core 02 §17 Domain invariants | base, reference_pack | AC-231, AC-234 |
| REQ-02-222 | Core 02 §18 Canonical closed-vocabulary registry | base, incident_portability | AC-076..AC-084, AC-121..AC-122, AC-137..AC-145, AC-231, AC-236 |
| REQ-02-223 | Core 02 §18 Canonical closed-vocabulary registry | base | AC-076..AC-084, AC-121..AC-122, AC-137..AC-145, AC-231 |
| REQ-03-001 | Core 03 §1 Interaction model | base | AC-001..AC-002, AC-005, AC-043, AC-231 |
| REQ-03-002 | Core 03 §1 Interaction model | base | AC-001..AC-002, AC-005, AC-043, AC-231 |
| REQ-03-003 | Core 03 §1 Interaction model | base | AC-001..AC-002, AC-005, AC-043, AC-231 |
| REQ-03-004 | Core 03 §2.1 Built-in tabs | base | AC-112, AC-116, AC-231 |
| REQ-03-005 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-006 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-007 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-008 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-009 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-010 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-011 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-012 | Core 03 §2.3 Saved views | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-03-013 | Core 03 §2.3 Saved views | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-03-014 | Core 03 §2.3 Saved views | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-03-015 | Core 03 §2.3 Saved views | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-03-016 | Core 03 §2.3 Saved views | base | AC-146..AC-149, AC-151..AC-152, AC-231 |
| REQ-03-017 | Core 03 §2.3.1 Scope and discoverability | base | AC-146..AC-149, AC-151, AC-231 |
| REQ-03-018 | Core 03 §2.3.1 Scope and discoverability | base | AC-146..AC-149, AC-151, AC-231 |
| REQ-03-019 | Core 03 §2.3.1 Scope and discoverability | base | AC-146..AC-149, AC-151, AC-231 |
| REQ-03-020 | Core 03 §2.3.1 Scope and discoverability | base | AC-146..AC-149, AC-151, AC-231 |
| REQ-03-021 | Core 03 §2.3.1 Scope and discoverability | base, snapshot_reporting | AC-146..AC-149, AC-151, AC-231, AC-233 |
| REQ-03-022 | Core 03 §2.3.2 Ordinary lifecycle semantics | base | AC-152, AC-231 |
| REQ-03-023 | Core 03 §2.3.2 Ordinary lifecycle semantics | base | AC-152, AC-231 |
| REQ-03-024 | Core 03 §2.3.2 Ordinary lifecycle semantics | base | AC-152, AC-231 |
| REQ-03-025 | Core 03 §2.3.2 Ordinary lifecycle semantics | base | AC-152, AC-231 |
| REQ-03-026 | Core 03 §2.3.2 Ordinary lifecycle semantics | base | AC-152, AC-231 |
| REQ-03-027 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-028 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-029 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-030 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-031 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-032 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-033 | Core 03 §3.1 Concurrency strategy | base | AC-009, AC-013, AC-047, AC-231 |
| REQ-03-034 | Core 03 §3.1 Concurrency strategy | base | AC-009, AC-013, AC-047, AC-231 |
| REQ-03-035 | Core 03 §3.1 Concurrency strategy | base | AC-009, AC-013, AC-047, AC-231 |
| REQ-03-036 | Core 03 §3.2 Server-side conflict behavior | base | AC-009, AC-126, AC-231 |
| REQ-03-037 | Core 03 §3.2 Server-side conflict behavior | base | AC-009, AC-126, AC-231 |
| REQ-03-038 | Core 03 §3.2 Server-side conflict behavior | base | AC-009, AC-126, AC-231 |
| REQ-03-039 | Core 03 §3.2 Server-side conflict behavior | base | AC-009, AC-126, AC-231 |
| REQ-03-040 | Core 03 §3.2 Server-side conflict behavior | base | AC-009, AC-126, AC-231 |
| REQ-03-041 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-042 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-043 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-044 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-045 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-046 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-047 | Core 03 §3.3.1 Resolver surface and conflict state | base | AC-037..AC-039, AC-042, AC-231 |
| REQ-03-048 | Core 03 §3.3.2 Resolver contents and safety rules | base | AC-037..AC-042, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-049 | Core 03 §3.3.2 Resolver contents and safety rules | base | AC-037..AC-042, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-050 | Core 03 §3.3.2 Resolver contents and safety rules | base | AC-037..AC-042, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-051 | Core 03 §3.3.2 Resolver contents and safety rules | base | AC-037..AC-042, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-052 | Core 03 §3.3.3 Contract-declared resolution classes | base | AC-118, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-053 | Core 03 §3.3.3 Contract-declared resolution classes | base | AC-118, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-054 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-055 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-056 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-057 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-058 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-059 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-060 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-061 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-062 | Core 03 §3.3.3.1 Operational semantics for `text_compare_merge` | base | AC-226..AC-231 |
| REQ-03-063 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-064 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-065 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-066 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-067 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-068 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-069 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-070 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-071 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-072 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-073 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-074 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-075 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-076 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231 |
| REQ-03-077 | Core 03 §3.3.5 Local draft, history, and analytics boundary | base | AC-041, AC-163, AC-231 |
| REQ-03-078 | Core 03 §3.3.5 Local draft, history, and analytics boundary | base, snapshot_reporting | AC-041, AC-163, AC-231, AC-233 |
| REQ-03-079 | Core 03 §3.3.5 Local draft, history, and analytics boundary | base | AC-041, AC-163, AC-231 |
| REQ-03-080 | Core 03 §3.3.5 Local draft, history, and analytics boundary | base | AC-041, AC-163, AC-231 |
| REQ-03-081 | Core 03 §3.3.5 Local draft, history, and analytics boundary | base | AC-041, AC-163, AC-231 |
| REQ-03-082 | Core 03 §3.3.5 Local draft, history, and analytics boundary | base | AC-041, AC-163, AC-231 |
| REQ-03-083 | Core 03 §3.3.6 Paste-time same-field conflicts | base | AC-040, AC-231 |
| REQ-03-084 | Core 03 §3.3.6 Paste-time same-field conflicts | base | AC-040, AC-231 |
| REQ-03-085 | Core 03 §3.3.6 Paste-time same-field conflicts | base | AC-040, AC-231 |
| REQ-03-086 | Core 03 §3.4 Client addressing rules | base | AC-013, AC-047, AC-125, AC-231 |
| REQ-03-087 | Core 03 §4.1 Autosave | base | AC-043, AC-231 |
| REQ-03-088 | Core 03 §4.1 Autosave | base | AC-043, AC-231 |
| REQ-03-089 | Core 03 §4.2 Save-state presentation | base | AC-043, AC-231 |
| REQ-03-090 | Core 03 §4.3 Presence | base | AC-008, AC-132, AC-231 |
| REQ-03-091 | Core 03 §4.3 Presence | base | AC-008, AC-132, AC-231 |
| REQ-03-092 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-093 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-094 | Core 03 §4.3.1 Collaboration message application | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-231, AC-233 |
| REQ-03-095 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-096 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-097 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-098 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-099 | Core 03 §4.4 Local pending queue | base | AC-156..AC-163, AC-231 |
| REQ-03-100 | Core 03 §4.4 Local pending queue | base | AC-156..AC-163, AC-231 |
| REQ-03-101 | Core 03 §5 Locking policy | base | AC-182, AC-187, AC-218, AC-231 |
| REQ-03-102 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-103 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-104 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-105 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-106 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-107 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-108 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-109 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-110 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-111 | Core 03 §7 Timeline creation workflow | base | AC-001..AC-002, AC-125, AC-191, AC-193, AC-231 |
| REQ-03-112 | Core 03 §7 Timeline creation workflow | base | AC-001..AC-002, AC-125, AC-191, AC-193, AC-231 |
| REQ-03-113 | Core 03 §7 Timeline creation workflow | base | AC-001..AC-002, AC-125, AC-191, AC-193, AC-231 |
| REQ-03-114 | Core 03 §7 Timeline creation workflow | base | AC-001..AC-002, AC-125, AC-191, AC-193, AC-231 |
| REQ-03-115 | Core 03 §7 Timeline creation workflow | base | AC-001..AC-002, AC-125, AC-191, AC-193, AC-231 |
| REQ-03-116 | Core 03 §8.1 Two-step upload | base | AC-004, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-117 | Core 03 §8.1 Two-step upload | base | AC-004, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-118 | Core 03 §8.1 Two-step upload | base | AC-004, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-119 | Core 03 §8.1 Two-step upload | base | AC-004, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-120 | Core 03 §8.2 Pending evidence without blob | base | AC-015, AC-102, AC-154..AC-155, AC-231 |
| REQ-03-121 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-122 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-123 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-124 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-125 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-126 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-03-127 | Core 03 §8.4 Evidence access | base | AC-053..AC-054, AC-103, AC-128, AC-231 |
| REQ-03-128 | Core 03 §8.4 Evidence access | base | AC-053..AC-054, AC-103, AC-128, AC-231 |
| REQ-03-129 | Core 03 §9 Mention resolution workflow | base | AC-006, AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-03-130 | Core 03 §9 Mention resolution workflow | base | AC-006, AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-03-131 | Core 03 §9 Mention resolution workflow | base | AC-006, AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-03-132 | Core 03 §9 Mention resolution workflow | base | AC-006, AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-03-133 | Core 03 §9 Mention resolution workflow | base | AC-006, AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-03-134 | Core 03 §9 Mention resolution workflow | base | AC-006, AC-019..AC-021, AC-188..AC-190, AC-221..AC-225, AC-231 |
| REQ-03-135 | Core 03 §9.1 Source-bound indicator workflow | base | AC-017, AC-072..AC-079, AC-231 |
| REQ-03-136 | Core 03 §9.1 Source-bound indicator workflow | base | AC-017, AC-072..AC-079, AC-231 |
| REQ-03-137 | Core 03 §9.1 Source-bound indicator workflow | base | AC-017, AC-072..AC-079, AC-231 |
| REQ-03-138 | Core 03 §10.1 Reviewer lens | base | AC-007, AC-231 |
| REQ-03-139 | Core 03 §10.2 Minimum history presentation | base | AC-007, AC-215, AC-231 |
| REQ-03-140 | Core 03 §10.2 Minimum history presentation | base | AC-007, AC-215, AC-231 |
| REQ-03-141 | Core 03 §10.3 Rollback granularity | base | AC-010..AC-012, AC-215..AC-218, AC-231 |
| REQ-03-142 | Core 03 §10.3 Rollback granularity | base | AC-010..AC-012, AC-215..AC-218, AC-231 |
| REQ-03-143 | Core 03 §10.4 Rollback semantics | base | AC-216..AC-218, AC-231 |
| REQ-03-144 | Core 03 §10.4 Rollback semantics | base | AC-216..AC-218, AC-231 |
| REQ-03-145 | Core 03 §11.1 Clipboard paste | base | AC-003, AC-231 |
| REQ-03-146 | Core 03 §11.1 Clipboard paste | base, import | AC-003, AC-231..AC-232 |
| REQ-03-147 | Core 03 §11.1 Clipboard paste | base, import | AC-003, AC-231..AC-232 |
| REQ-03-148 | Core 03 §11.1 Clipboard paste | base | AC-003, AC-231 |
| REQ-03-149 | Core 03 §11.1 Clipboard paste | base | AC-003, AC-231 |
| REQ-03-150 | Core 03 §11.1 Clipboard paste | base | AC-003, AC-231 |
| REQ-03-151 | Core 03 §11.1 Clipboard paste | base | AC-003, AC-231 |
| REQ-03-152 | Core 03 §11.1 Clipboard paste | base | AC-003, AC-231 |
| REQ-03-153 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-154 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-155 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-156 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-157 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-158 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-159 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-160 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-161 | Core 03 §11.2.1 Assistant boundary | import | AC-027..AC-029, AC-063, AC-067, AC-232 |
| REQ-03-162 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-163 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-164 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-165 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-166 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-167 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-168 | Core 03 §11.2.2 `import_session` | import | AC-027, AC-063..AC-064, AC-232 |
| REQ-03-169 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-170 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-171 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-172 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-173 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-174 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-175 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-176 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-177 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-178 | Core 03 §11.2.3 `import_unit` | import | AC-027..AC-029, AC-063..AC-066, AC-232 |
| REQ-03-179 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-180 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-181 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-182 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-183 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-184 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-185 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-186 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-027, AC-064..AC-066, AC-232 |
| REQ-03-187 | Core 03 §11.2.5 `mapping_fingerprint` and duplicate-apply detection | import | AC-065, AC-232 |
| REQ-03-188 | Core 03 §11.2.5 `mapping_fingerprint` and duplicate-apply detection | import | AC-065, AC-232 |
| REQ-03-189 | Core 03 §11.2.5 `mapping_fingerprint` and duplicate-apply detection | import | AC-065, AC-232 |
| REQ-03-190 | Core 03 §11.2.5 `mapping_fingerprint` and duplicate-apply detection | import | AC-065, AC-232 |
| REQ-03-191 | Core 03 §11.2.5 `mapping_fingerprint` and duplicate-apply detection | import | AC-065, AC-232 |
| REQ-03-192 | Core 03 §11.2.5 `mapping_fingerprint` and duplicate-apply detection | import | AC-065, AC-232 |
| REQ-03-193 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-194 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-195 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-196 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-197 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-198 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-199 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-200 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-201 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-202 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-203 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-204 | Core 03 §11.2.6 Closed warning vocabulary and workbook downgrade semantics | import | AC-063, AC-067, AC-232 |
| REQ-03-205 | Core 03 §12.2 Required confidence conditions | base | AC-205, AC-231 |
| REQ-03-206 | Core 03 §12.3 Required write effects | base | AC-205, AC-231 |
| REQ-03-207 | Core 03 §12.3 Required write effects | base | AC-205, AC-231 |
| REQ-03-208 | Core 03 §12.4 Allowed and forbidden workflows | base, import | AC-205, AC-231..AC-232 |
| REQ-03-209 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-210 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-211 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-212 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-213 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-214 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-215 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-216 | Core 03 §12.5 Disclosure and undo | base | AC-006, AC-188..AC-190, AC-205, AC-231 |
| REQ-03-217 | Core 03 §13.1 Grid editing | base | AC-005, AC-043, AC-231 |
| REQ-03-218 | Core 03 §13.1 Grid editing | base | AC-005, AC-043, AC-231 |
| REQ-03-219 | Core 03 §13.1 Grid editing | base | AC-005, AC-043, AC-231 |
| REQ-03-220 | Core 03 §13.2 Required keyboard actions | base | AC-005, AC-231 |
| REQ-03-221 | Core 03 §13.3 Bulk editing | base | AC-003, AC-040, AC-231 |
| REQ-03-222 | Core 03 §13.3 Bulk editing | base | AC-003, AC-040, AC-231 |
| REQ-03-223 | Core 03 §14.1 Sort and filter behavior | base | AC-013..AC-014, AC-044, AC-047, AC-124, AC-184..AC-185, AC-231 |
| REQ-03-224 | Core 03 §14.1 Sort and filter behavior | base | AC-013..AC-014, AC-044, AC-047, AC-124, AC-184..AC-185, AC-231 |
| REQ-03-225 | Core 03 §14.2 Timeline grouping boundary | base | AC-024..AC-026, AC-231 |
| REQ-03-226 | Core 03 §14.3 Allowed grouping keys | base | AC-024..AC-026, AC-231 |
| REQ-03-227 | Core 03 §14.3 Allowed grouping keys | base | AC-024..AC-026, AC-231 |
| REQ-03-228 | Core 03 §14.3 Allowed grouping keys | base | AC-024..AC-026, AC-231 |
| REQ-03-229 | Core 03 §14.4 Grouping value rules | base | AC-024..AC-026, AC-231 |
| REQ-03-230 | Core 03 §14.4 Grouping value rules | base | AC-024..AC-026, AC-231 |
| REQ-03-231 | Core 03 §14.5 Group-header behavior | base | AC-025..AC-026, AC-231 |
| REQ-03-232 | Core 03 §14.5 Group-header behavior | base | AC-025..AC-026, AC-231 |
| REQ-03-233 | Core 03 §14.6 Edit movement across groups | base | AC-026, AC-047, AC-231 |
| REQ-03-234 | Core 03 §14.7 Collaborative state boundary | base | AC-047, AC-231 |
| REQ-03-235 | Core 03 §14.7 Collaborative state boundary | base | AC-047, AC-231 |
| REQ-03-236 | Core 03 §15 Timeline read and write contract | base | AC-119..AC-120, AC-124..AC-125, AC-188..AC-193, AC-231 |
| REQ-03-237 | Core 03 §15 Timeline read and write contract | base | AC-119..AC-120, AC-124..AC-125, AC-188..AC-193, AC-231 |
| REQ-03-238 | Core 03 §15 Timeline read and write contract | base | AC-119..AC-120, AC-124..AC-125, AC-188..AC-193, AC-231 |
| REQ-03-239 | Core 03 §15 Timeline read and write contract | base | AC-119..AC-120, AC-124..AC-125, AC-188..AC-193, AC-231 |
| REQ-03-240 | Core 03 §15 Timeline read and write contract | base | AC-119..AC-120, AC-124..AC-125, AC-188..AC-193, AC-231 |
| REQ-03-241 | Core 03 §15 Timeline read and write contract | base | AC-119..AC-120, AC-124..AC-125, AC-188..AC-193, AC-231 |
| REQ-03-242 | Core 03 §16.1 Entity and evidence sheets | base | AC-015, AC-045, AC-097..AC-100, AC-112, AC-116..AC-118, AC-231 |
| REQ-03-243 | Core 03 §16.1 Entity and evidence sheets | base | AC-015, AC-045, AC-097..AC-100, AC-112, AC-116..AC-118, AC-231 |
| REQ-03-244 | Core 03 §16.1 Entity and evidence sheets | base | AC-015, AC-045, AC-097..AC-100, AC-112, AC-116..AC-118, AC-231 |
| REQ-03-245 | Core 03 §16.1 Entity and evidence sheets | base | AC-015, AC-045, AC-097..AC-100, AC-112, AC-116..AC-118, AC-231 |
| REQ-03-246 | Core 03 §16.1 Entity and evidence sheets | base | AC-015, AC-045, AC-097..AC-100, AC-112, AC-116..AC-118, AC-231 |
| REQ-03-247 | Core 03 §16.2 Inspector | base | AC-006, AC-020, AC-023, AC-072..AC-075, AC-186..AC-187, AC-209..AC-210, AC-231 |
| REQ-03-248 | Core 03 §16.2 Inspector | base | AC-006, AC-020, AC-023, AC-072..AC-075, AC-186..AC-187, AC-209..AC-210, AC-231 |
| REQ-03-249 | Core 03 §16.2 Inspector | base | AC-006, AC-020, AC-023, AC-072..AC-075, AC-186..AC-187, AC-209..AC-210, AC-231 |
| REQ-03-250 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-251 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-252 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-253 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-254 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-255 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-256 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-257 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-258 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-259 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-260 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-261 | Core 03 §17 Authorship and attribution in the UI | base | AC-007, AC-231 |
| REQ-03-262 | Core 03 §17 Authorship and attribution in the UI | base | AC-007, AC-231 |
| REQ-03-263 | Core 03 §18.1 Excel-like behaviors | base | AC-005, AC-043, AC-231 |
| REQ-03-264 | Core 03 §18.2 Intentional differences | base | AC-090, AC-231 |
| REQ-03-265 | Core 03 §19 Interaction invariants | base | AC-231 |
| REQ-04-001 | Core 04 §1.1 Base authentication | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-04-002 | Core 04 §1.1 Base authentication | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-04-003 | Core 04 §1.1 Base authentication | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-04-004 | Core 04 §1.1 Base authentication | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-04-005 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-006 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-007 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-008 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-009 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-010 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-011 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-012 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-013 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-014 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-015 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-016 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-017 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-018 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-036, AC-235 |
| REQ-04-019 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-036, AC-235 |
| REQ-04-020 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-036, AC-235 |
| REQ-04-021 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-022 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-023 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-024 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-025 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-026 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-027 | Core 04 §2 Authorization model | base, snapshot_reporting | AC-054, AC-149, AC-178..AC-180, AC-231, AC-233 |
| REQ-04-028 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-029 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-030 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-031 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-032 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-033 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-034 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-035 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-036 | Core 04 §3 Attribution and audit requirements | base | AC-231 |
| REQ-04-037 | Core 04 §3 Attribution and audit requirements | base | AC-231 |
| REQ-04-038 | Core 04 §3 Attribution and audit requirements | base, incident_portability | AC-231, AC-236 |
| REQ-04-039 | Core 04 §3 Attribution and audit requirements | base | AC-231 |
| REQ-04-040 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-041 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-042 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-043 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-044 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-045 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-046 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-047 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-048 | Core 04 §4.3 Evidence uploads | base | AC-053, AC-128, AC-231 |
| REQ-04-049 | Core 04 §4.4 STRIDE threat model | base | AC-048, AC-231 |
| REQ-04-050 | Core 04 §4.4 STRIDE threat model | base | AC-048, AC-231 |
| REQ-04-051 | Core 04 §4.4 STRIDE threat model | base | AC-048, AC-231 |
| REQ-04-052 | Core 04 §4.5 Focused MITRE CWE constraints | base | AC-049..AC-055, AC-130..AC-131, AC-231 |
| REQ-04-053 | Core 04 §4.5 Focused MITRE CWE constraints | base, import, reference_pack | AC-049..AC-055, AC-130..AC-131, AC-231..AC-232, AC-234 |
| REQ-04-054 | Core 04 §5.1 Flyaway or disconnected deployment | base | AC-055, AC-092, AC-096, AC-169, AC-231 |
| REQ-04-055 | Core 04 §5.1 Flyaway or disconnected deployment | base, reference_pack | AC-055, AC-092, AC-096, AC-169, AC-231, AC-234 |
| REQ-04-056 | Core 04 §5.2 On-prem deployment | base | AC-231 |
| REQ-04-057 | Core 04 §5.3 Cloud deployment | base | AC-231 |
| REQ-04-058 | Core 04 §6 Runtime roots and storage paths | base, reference_pack | AC-051, AC-055, AC-169, AC-231, AC-234 |
| REQ-04-059 | Core 04 §6 Runtime roots and storage paths | base | AC-051, AC-055, AC-169, AC-231 |
| REQ-04-060 | Core 04 §7 Container boundary | base | AC-231 |
| REQ-04-061 | Core 04 §8.1 Required services | base | AC-231 |
| REQ-04-062 | Core 04 §9 Conformance criteria | base | AC-231, AC-237 |
| REQ-04-063 | Core 04 §9 Conformance criteria | base | AC-231, AC-237 |
| REQ-04-064 | Core 04 §9 Conformance criteria | base | AC-231, AC-237 |
| REQ-04-065 | Core 04 §11 Operational posture | base | AC-164..AC-169, AC-231 |

## F.5 Acceptance-criterion reverse navigation

| Acceptance criterion | Requirements |
| --- | --- |
| AC-001 | REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115 |
| AC-002 | REQ-03-001..REQ-03-003, REQ-03-111..REQ-03-115 |
| AC-003 | REQ-01-015..REQ-01-017, REQ-03-145..REQ-03-152, REQ-03-221..REQ-03-222 |
| AC-004 | REQ-01-015..REQ-01-017, REQ-03-116..REQ-03-119 |
| AC-005 | REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-217..REQ-03-220, REQ-03-263 |
| AC-006 | REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-247..REQ-03-249 |
| AC-007 | REQ-03-138..REQ-03-140, REQ-03-261..REQ-03-262 |
| AC-008 | REQ-03-090..REQ-03-091 |
| AC-009 | REQ-03-033..REQ-03-040 |
| AC-010 | REQ-03-141..REQ-03-142 |
| AC-011 | REQ-03-141..REQ-03-142 |
| AC-012 | REQ-03-141..REQ-03-142 |
| AC-013 | REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224 |
| AC-014 | REQ-03-223..REQ-03-224 |
| AC-015 | REQ-01-243..REQ-01-247, REQ-01-291..REQ-01-295, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201, REQ-03-120..REQ-03-126, REQ-03-242..REQ-03-246 |
| AC-016 | REQ-01-243..REQ-01-247, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201, REQ-03-121..REQ-03-126 |
| AC-017 | REQ-01-331, REQ-01-355..REQ-01-366, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204, REQ-03-135..REQ-03-137 |
| AC-018 | REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-202..REQ-02-204, REQ-03-250..REQ-03-254 |
| AC-019 | REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-027, REQ-02-030..REQ-02-036, REQ-03-129..REQ-03-134 |
| AC-020 | REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-038, REQ-03-129..REQ-03-134, REQ-03-247..REQ-03-249 |
| AC-021 | REQ-01-196..REQ-01-227, REQ-02-042..REQ-02-044, REQ-02-058..REQ-02-059, REQ-03-129..REQ-03-134 |
| AC-022 | REQ-02-030..REQ-02-036, REQ-02-060..REQ-02-061 |
| AC-023 | REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249 |
| AC-024 | REQ-03-225..REQ-03-230 |
| AC-025 | REQ-03-225..REQ-03-232 |
| AC-026 | REQ-03-225..REQ-03-233 |
| AC-027 | REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-186 |
| AC-028 | REQ-01-004..REQ-01-014, REQ-02-030..REQ-02-038, REQ-02-045..REQ-02-053, REQ-02-058..REQ-02-063, REQ-03-153..REQ-03-161, REQ-03-169..REQ-03-178 |
| AC-029 | REQ-01-004..REQ-01-014, REQ-02-030..REQ-02-038, REQ-02-045..REQ-02-053, REQ-02-062..REQ-02-063, REQ-03-153..REQ-03-161, REQ-03-169..REQ-03-178 |
| AC-030 | REQ-01-369..REQ-01-373, REQ-01-452..REQ-01-454, REQ-02-139..REQ-02-146 |
| AC-031 | REQ-01-370..REQ-01-373, REQ-01-394..REQ-01-398, REQ-02-139..REQ-02-146, REQ-04-044..REQ-04-047 |
| AC-032 | REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-367..REQ-01-368, REQ-01-370..REQ-01-373, REQ-02-139..REQ-02-146 |
| AC-033 | REQ-01-399, REQ-01-407..REQ-01-413, REQ-01-452..REQ-01-454, REQ-04-040..REQ-04-043 |
| AC-034 | REQ-01-282..REQ-01-284, REQ-01-399, REQ-01-419..REQ-01-422 |
| AC-035 | REQ-01-399, REQ-01-409..REQ-01-421, REQ-04-040..REQ-04-043 |
| AC-036 | REQ-04-018..REQ-04-020 |
| AC-037 | REQ-03-041..REQ-03-051 |
| AC-038 | REQ-03-041..REQ-03-051 |
| AC-039 | REQ-03-041..REQ-03-051 |
| AC-040 | REQ-03-048..REQ-03-051, REQ-03-083..REQ-03-085, REQ-03-221..REQ-03-222 |
| AC-041 | REQ-03-048..REQ-03-051, REQ-03-077..REQ-03-082 |
| AC-042 | REQ-03-041..REQ-03-051 |
| AC-043 | REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-087..REQ-03-089, REQ-03-217..REQ-03-219, REQ-03-263 |
| AC-044 | REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-223..REQ-03-224 |
| AC-045 | REQ-00-013, REQ-01-015..REQ-01-017, REQ-01-355..REQ-01-366, REQ-03-242..REQ-03-246 |
| AC-046 | REQ-00-013, REQ-01-004..REQ-01-014, REQ-01-018, REQ-01-248..REQ-01-249, REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-369, REQ-01-452..REQ-01-454 |
| AC-047 | REQ-01-015..REQ-01-017, REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224, REQ-03-233..REQ-03-235 |
| AC-048 | REQ-04-049..REQ-04-051 |
| AC-049 | REQ-04-052..REQ-04-053 |
| AC-050 | REQ-04-052..REQ-04-053 |
| AC-051 | REQ-01-455..REQ-01-456, REQ-04-052..REQ-04-053, REQ-04-058..REQ-04-059 |
| AC-052 | REQ-04-040..REQ-04-043, REQ-04-052..REQ-04-053 |
| AC-053 | REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-201, REQ-03-127..REQ-03-128, REQ-04-048, REQ-04-052..REQ-04-053 |
| AC-054 | REQ-01-355..REQ-01-366, REQ-03-127..REQ-03-128, REQ-04-021..REQ-04-030, REQ-04-052..REQ-04-053 |
| AC-055 | REQ-01-455..REQ-01-456, REQ-04-052..REQ-04-055, REQ-04-058..REQ-04-059 |
| AC-056 | REQ-01-370..REQ-01-373, REQ-02-139..REQ-02-146 |
| AC-057 | REQ-01-370..REQ-01-373, REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-058 | REQ-01-370..REQ-01-373, REQ-01-381..REQ-01-384, REQ-02-139..REQ-02-146 |
| AC-059 | REQ-01-374..REQ-01-380, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035, REQ-04-044..REQ-04-047 |
| AC-060 | REQ-01-374..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-031..REQ-04-035, REQ-04-044..REQ-04-047 |
| AC-061 | REQ-01-377..REQ-01-380, REQ-01-394..REQ-01-397, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-062 | REQ-01-377..REQ-01-380, REQ-01-394..REQ-01-397, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-063 | REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-178, REQ-03-193..REQ-03-204 |
| AC-064 | REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-162..REQ-03-186 |
| AC-065 | REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-192 |
| AC-066 | REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-169..REQ-03-186 |
| AC-067 | REQ-01-004..REQ-01-014, REQ-02-045..REQ-02-053, REQ-03-153..REQ-03-161, REQ-03-193..REQ-03-204 |
| AC-068 | REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071 |
| AC-069 | REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330 |
| AC-070 | REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071 |
| AC-071 | REQ-01-377..REQ-01-380, REQ-02-139..REQ-02-146 |
| AC-072 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204, REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249 |
| AC-073 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204, REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249 |
| AC-074 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204, REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249 |
| AC-075 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-202..REQ-02-204, REQ-03-135..REQ-03-137, REQ-03-247..REQ-03-249 |
| AC-076 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-135..REQ-03-137 |
| AC-077 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-135..REQ-03-137 |
| AC-078 | REQ-01-296..REQ-01-302, REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011, REQ-03-135..REQ-03-137 |
| AC-079 | REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-135..REQ-03-137 |
| AC-080 | REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254 |
| AC-081 | REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254 |
| AC-082 | REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254 |
| AC-083 | REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254 |
| AC-084 | REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-250..REQ-03-254 |
| AC-085 | REQ-01-296..REQ-01-302, REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-120..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-086 | REQ-01-296..REQ-01-302, REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-087 | REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-088 | REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-089 | REQ-01-296..REQ-01-302, REQ-02-067..REQ-02-071, REQ-02-094, REQ-02-123..REQ-02-134, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-090 | REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-120..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260, REQ-03-264 |
| AC-091 | REQ-01-377..REQ-01-384, REQ-02-139..REQ-02-146, REQ-04-044..REQ-04-047 |
| AC-092 | REQ-01-282..REQ-01-284, REQ-01-400..REQ-01-406, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055 |
| AC-093 | REQ-01-407..REQ-01-413, REQ-04-040..REQ-04-043 |
| AC-094 | REQ-01-407..REQ-01-418, REQ-04-040..REQ-04-043 |
| AC-095 | REQ-01-409..REQ-01-421, REQ-04-040..REQ-04-043 |
| AC-096 | REQ-01-282..REQ-01-284, REQ-01-407..REQ-01-413, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055 |
| AC-097 | REQ-01-323..REQ-01-325, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246 |
| AC-098 | REQ-01-326..REQ-01-327, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246 |
| AC-099 | REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246 |
| AC-100 | REQ-01-328, REQ-01-355..REQ-01-366, REQ-02-009..REQ-02-023, REQ-02-186..REQ-02-201, REQ-03-242..REQ-03-246 |
| AC-101 | REQ-02-009..REQ-02-023, REQ-02-135..REQ-02-138 |
| AC-102 | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-120 |
| AC-103 | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128 |
| AC-104 | REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035 |
| AC-105 | REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035 |
| AC-106 | REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035 |
| AC-107 | REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110 |
| AC-108 | REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110 |
| AC-109 | REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110 |
| AC-110 | REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110 |
| AC-111 | REQ-00-016..REQ-00-017, REQ-03-102..REQ-03-110 |
| AC-112 | REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071, REQ-03-004, REQ-03-242..REQ-03-246 |
| AC-113 | REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-114 | REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-115 | REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-116 | REQ-00-014, REQ-01-240..REQ-01-242, REQ-01-285..REQ-01-295, REQ-01-307..REQ-01-311, REQ-03-004, REQ-03-242..REQ-03-246 |
| AC-117 | REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-242..REQ-03-246 |
| AC-118 | REQ-00-014, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-01-323..REQ-01-341, REQ-02-028..REQ-02-029, REQ-02-202..REQ-02-204, REQ-03-052..REQ-03-053, REQ-03-242..REQ-03-246 |
| AC-119 | REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-322, REQ-03-236..REQ-03-241 |
| AC-120 | REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-236..REQ-03-241 |
| AC-121 | REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011, REQ-03-250..REQ-03-254 |
| AC-122 | REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011 |
| AC-123 | REQ-00-014, REQ-01-023..REQ-01-031, REQ-04-001..REQ-04-017 |
| AC-124 | REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-341, REQ-03-223..REQ-03-224, REQ-03-236..REQ-03-241 |
| AC-125 | REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-057..REQ-01-088, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-330, REQ-02-208..REQ-02-209, REQ-03-086, REQ-03-111..REQ-03-115, REQ-03-236..REQ-03-241 |
| AC-126 | REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-036..REQ-03-040, REQ-03-063..REQ-03-076 |
| AC-127 | REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-117..REQ-01-118, REQ-01-129, REQ-01-144, REQ-01-168, REQ-01-240..REQ-01-242, REQ-01-288 |
| AC-128 | REQ-00-014, REQ-01-019, REQ-01-243..REQ-01-247, REQ-01-328, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128, REQ-04-048 |
| AC-129 | REQ-00-014, REQ-01-018..REQ-01-019, REQ-01-248..REQ-01-277, REQ-01-452..REQ-01-454, REQ-03-092..REQ-03-098 |
| AC-130 | REQ-01-023..REQ-01-031, REQ-01-154, REQ-04-001..REQ-04-004, REQ-04-052..REQ-04-053 |
| AC-131 | REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017, REQ-04-052..REQ-04-053 |
| AC-132 | REQ-01-250..REQ-01-277, REQ-03-090..REQ-03-098 |
| AC-133 | REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098 |
| AC-134 | REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098 |
| AC-135 | REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098 |
| AC-136 | REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017 |
| AC-137 | REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-138 | REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-139 | REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-140 | REQ-01-336..REQ-01-338, REQ-02-094..REQ-02-109, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-141 | REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-142 | REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-143 | REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-144 | REQ-01-339..REQ-01-341, REQ-02-094, REQ-02-110..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-145 | REQ-01-336..REQ-01-341, REQ-02-094..REQ-02-119, REQ-02-222..REQ-02-223, REQ-03-102..REQ-03-110, REQ-03-255..REQ-03-260 |
| AC-146 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021 |
| AC-147 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021 |
| AC-148 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021 |
| AC-149 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021, REQ-04-021..REQ-04-030 |
| AC-150 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032 |
| AC-151 | REQ-01-138..REQ-01-151, REQ-01-240..REQ-01-242, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-021 |
| AC-152 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-016, REQ-03-022..REQ-03-026 |
| AC-153 | REQ-01-138..REQ-01-151, REQ-02-147..REQ-02-151, REQ-02-158..REQ-02-162, REQ-03-027..REQ-03-032 |
| AC-154 | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126 |
| AC-155 | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-126 |
| AC-156 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-157 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-158 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-159 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-160 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-161 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-162 | REQ-01-023..REQ-01-031, REQ-01-250..REQ-01-277, REQ-03-099..REQ-03-100, REQ-04-001..REQ-04-017 |
| AC-163 | REQ-01-250..REQ-01-277, REQ-03-077..REQ-03-082, REQ-03-099..REQ-03-100, REQ-04-005..REQ-04-017 |
| AC-164 | REQ-01-425..REQ-01-442, REQ-04-044..REQ-04-047, REQ-04-065 |
| AC-165 | REQ-01-425..REQ-01-426, REQ-01-439..REQ-01-442, REQ-01-447..REQ-01-450, REQ-04-044..REQ-04-047, REQ-04-065 |
| AC-166 | REQ-01-425..REQ-01-430, REQ-01-433..REQ-01-438, REQ-01-447..REQ-01-451, REQ-04-044..REQ-04-047, REQ-04-065 |
| AC-167 | REQ-01-425..REQ-01-426, REQ-01-431..REQ-01-432, REQ-01-443..REQ-01-450, REQ-04-044..REQ-04-047, REQ-04-065 |
| AC-168 | REQ-01-425..REQ-01-426, REQ-01-443..REQ-01-446, REQ-04-044..REQ-04-047, REQ-04-065 |
| AC-169 | REQ-01-425..REQ-01-430, REQ-01-447..REQ-01-450, REQ-01-452..REQ-01-456, REQ-04-044..REQ-04-047, REQ-04-054..REQ-04-055, REQ-04-058..REQ-04-059, REQ-04-065 |
| AC-170 | REQ-01-152..REQ-01-180 |
| AC-171 | REQ-01-152..REQ-01-180, REQ-01-240..REQ-01-242 |
| AC-172 | REQ-01-152..REQ-01-180 |
| AC-173 | REQ-01-152..REQ-01-180 |
| AC-174 | REQ-01-152..REQ-01-180 |
| AC-175 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-240..REQ-01-242 |
| AC-176 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137 |
| AC-177 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137 |
| AC-178 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-240..REQ-01-242, REQ-04-021..REQ-04-030 |
| AC-179 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030 |
| AC-180 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030 |
| AC-181 | REQ-01-057..REQ-01-088, REQ-02-210 |
| AC-182 | REQ-01-057..REQ-01-088, REQ-02-210, REQ-03-101 |
| AC-183 | REQ-01-057..REQ-01-088, REQ-02-210 |
| AC-184 | REQ-01-034..REQ-01-056, REQ-01-312..REQ-01-322, REQ-03-223..REQ-03-224 |
| AC-185 | REQ-01-034..REQ-01-056, REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-02-067..REQ-02-071, REQ-03-223..REQ-03-224 |
| AC-186 | REQ-01-032..REQ-01-033, REQ-01-181..REQ-01-195, REQ-02-054..REQ-02-055, REQ-02-060..REQ-02-061, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249 |
| AC-187 | REQ-01-032..REQ-01-033, REQ-01-181..REQ-01-195, REQ-01-237, REQ-02-060..REQ-02-061, REQ-02-064..REQ-02-066, REQ-03-101, REQ-03-247..REQ-03-249 |
| AC-188 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-02-058..REQ-02-059, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241 |
| AC-189 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241 |
| AC-190 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241 |
| AC-191 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241 |
| AC-192 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110, REQ-03-236..REQ-03-241 |
| AC-193 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241 |
| AC-194 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-195 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-196 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-197 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-198 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-199 | REQ-03-102..REQ-03-110 |
| AC-200 | REQ-01-057..REQ-01-088, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209 |
| AC-201 | REQ-01-057..REQ-01-088, REQ-02-030..REQ-02-036, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209 |
| AC-202 | REQ-01-057..REQ-01-088, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209 |
| AC-203 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209, REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076 |
| AC-204 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209, REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076 |
| AC-205 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216 |
| AC-206 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209 |
| AC-207 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209 |
| AC-208 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209 |
| AC-209 | REQ-01-057..REQ-01-088, REQ-01-181..REQ-01-195, REQ-02-054..REQ-02-055, REQ-02-064..REQ-02-066, REQ-02-163..REQ-02-185, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249 |
| AC-210 | REQ-01-057..REQ-01-088, REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-355..REQ-01-366, REQ-02-163..REQ-02-185, REQ-03-247..REQ-03-249 |
| AC-211 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023 |
| AC-212 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-02-014..REQ-02-023 |
| AC-213 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023 |
| AC-214 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023 |
| AC-215 | REQ-01-056, REQ-01-057..REQ-01-111, REQ-01-240..REQ-01-242, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-218, REQ-03-139..REQ-03-142 |
| AC-216 | REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-216, REQ-03-141..REQ-03-144 |
| AC-217 | REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-220, REQ-03-141..REQ-03-144 |
| AC-218 | REQ-01-057..REQ-01-111, REQ-01-228..REQ-01-239, REQ-02-205..REQ-02-207, REQ-03-101, REQ-03-141..REQ-03-144 |
| AC-219 | REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239 |
| AC-220 | REQ-01-152..REQ-01-180 |
| AC-221 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134 |
| AC-222 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134 |
| AC-223 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134 |
| AC-224 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041, REQ-03-129..REQ-03-134 |
| AC-225 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041, REQ-03-129..REQ-03-134 |
| AC-226 | REQ-03-048..REQ-03-076 |
| AC-227 | REQ-03-048..REQ-03-076 |
| AC-228 | REQ-03-048..REQ-03-076 |
| AC-229 | REQ-03-048..REQ-03-076 |
| AC-230 | REQ-03-048..REQ-03-076 |
| AC-231 | REQ-00-001..REQ-00-003, REQ-00-005..REQ-00-017, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-457, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-223, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-265, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-065 |
| AC-232 | REQ-00-004, REQ-01-006..REQ-01-014, REQ-01-457, REQ-02-032, REQ-02-045..REQ-02-053, REQ-02-202, REQ-03-146..REQ-03-147, REQ-03-153..REQ-03-204, REQ-03-208, REQ-04-053 |
| AC-233 | REQ-00-004, REQ-00-015, REQ-01-004, REQ-01-018, REQ-01-033, REQ-01-068, REQ-01-078, REQ-01-098, REQ-01-251, REQ-01-263, REQ-01-265, REQ-01-270, REQ-01-273..REQ-01-274, REQ-01-278..REQ-01-279, REQ-01-368..REQ-01-398, REQ-01-452, REQ-01-457, REQ-02-009, REQ-02-112, REQ-02-139..REQ-02-146, REQ-02-200, REQ-02-202, REQ-02-206..REQ-02-207, REQ-02-211, REQ-02-217, REQ-03-021, REQ-03-078, REQ-03-094, REQ-04-027, REQ-04-031..REQ-04-035, REQ-04-044..REQ-04-047 |
| AC-234 | REQ-00-004, REQ-00-013, REQ-00-015, REQ-01-018, REQ-01-033, REQ-01-278, REQ-01-282..REQ-01-284, REQ-01-286, REQ-01-308, REQ-01-310, REQ-01-399..REQ-01-422, REQ-01-452, REQ-01-455, REQ-02-001, REQ-02-003, REQ-02-009, REQ-02-202, REQ-02-221, REQ-04-040..REQ-04-043, REQ-04-053, REQ-04-055, REQ-04-058 |
| AC-235 | REQ-00-004, REQ-01-031, REQ-04-018..REQ-04-020 |
| AC-236 | REQ-00-004, REQ-01-425..REQ-01-450, REQ-01-457, REQ-02-204, REQ-02-222, REQ-04-038, REQ-04-044..REQ-04-047 |
| AC-237 | REQ-00-009..REQ-00-012, REQ-04-062..REQ-04-064 |
| AC-238 | REQ-01-035, REQ-01-036, REQ-01-242 |
| AC-239 | REQ-01-035, REQ-01-234, REQ-01-238, REQ-01-241, REQ-01-242 |
| AC-240 | REQ-01-035, REQ-01-234, REQ-01-238 |
| AC-241 | REQ-01-036, REQ-01-242 |
| AC-242 | REQ-01-242 |
| AC-243 | REQ-01-035, REQ-01-036, REQ-01-037 |

## F.6 Profile Definition-of-Done navigation

| Profile | Prerequisite claim | Required REQs | Required ACs |
| --- | --- | --- | --- |
| base | — | REQ-00-001..REQ-00-003, REQ-00-005..REQ-00-017, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-457, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-223, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-265, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-065 | AC-001..AC-026, AC-037..AC-055, AC-097..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-231, AC-237..AC-243 |
| import | base | REQ-00-001..REQ-00-017, REQ-01-001..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-457, REQ-02-001..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-223, REQ-03-001..REQ-03-265, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-065 | AC-001..AC-029, AC-037..AC-055, AC-063..AC-067, AC-097..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-232, AC-237..AC-243 |
| snapshot_reporting | base | REQ-00-001..REQ-00-017, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-398, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-457, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-223, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-265, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-039, REQ-04-044..REQ-04-065 | AC-001..AC-026, AC-030..AC-032, AC-037..AC-062, AC-091, AC-097..AC-163, AC-170..AC-231, AC-233, AC-237..AC-243 |
| reference_pack | base | REQ-00-001..REQ-00-017, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-399..REQ-01-424, REQ-01-451..REQ-01-457, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-223, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-265, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-043, REQ-04-048..REQ-04-065 | AC-001..AC-026, AC-033..AC-035, AC-037..AC-055, AC-092..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-231, AC-234, AC-237..AC-243 |
| incident_portability | base | REQ-00-001..REQ-00-017, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-457, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-223, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-265, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-044..REQ-04-065 | AC-001..AC-026, AC-037..AC-055, AC-097..AC-103, AC-107..AC-112, AC-116..AC-231, AC-236..AC-243 |
| enterprise_authentication | base | REQ-00-001..REQ-00-017, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-457, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-223, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-265, REQ-04-001..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-065 | AC-001..AC-026, AC-036..AC-055, AC-097..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-231, AC-235, AC-237..AC-243 |
