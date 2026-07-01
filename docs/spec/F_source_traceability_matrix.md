# Appendix F: Source Traceability Matrix

This appendix is **non-normative**.

It records where the content of the exploratory source artifact was carried into the derived document set. Appendix G preserves the full original artifact verbatim.

## F.1 Level-2 source section coverage

| Source section or topic | Current owner section(s) | Non-normative preservation target(s) | Supporting guidance | Preservation note |
| --- | --- | --- | --- | --- |
| `1. Executive summary` | Core 00 | Appendix A, Appendix G | — | rationale only |
| `2. Problem framing` | — | Appendix A, Appendix G | — | rationale only |
| `3. Design principles and assumptions` | Core 00, Core 01, Core 03 | Appendix A, Appendix G | — | rationale only |
| `4. Recommended architecture` | Core 01, Core 04 | Appendix B, Appendix G | — | illustration only |
| `5. Collaboration and consistency model` | Core 03 | Appendix G | Appendix H | history only |
| `6. Domain model and schema strategy` | Core 02 | Appendix G | — | history only |
| `7. Postgres schema proposal` | Core 02 | Appendix C, Appendix G | — | illustration only |
| `8. Record lifecycle and IR workflow model` | Core 02, Core 03 | Appendix D, Appendix G | — | illustration only |
| `9. UI concepts focused on preserving the spreadsheet feel` | Core 01, Core 03 | Appendix D, Appendix G | — | illustration only |
| `10. UX acceptance criteria` | Core 04, Core 05 | Appendix G | — | history only |
| `11. Security, auth, and access control` | Core 01, Core 04 | Appendix G | — | history only |
| `12. Deployment model` | Core 01, Core 04 | Appendix G | — | history only |
| `13. Trade-offs, risks, and rejected alternatives` | — | Appendix A, Appendix G | — | rationale only |
| `14. Comparison table` | — | Appendix A, Appendix G | — | rationale only |
| `15. Recommended MVP and phased roadmap` | — | Appendix E, Appendix G | — | future-only backlog |
| `16. Open questions` | Core 00, Core 01, Core 02, Core 03, Core 04 | Appendix E, Appendix G | Appendix H | history only |
| `Design sanity check` | — | Appendix A, Appendix G | — | rationale only |

## F.2 Critical subsection coverage

| Source section or topic | Current owner section(s) | Non-normative preservation target(s) | Supporting guidance | Preservation note |
| --- | --- | --- | --- | --- |
| `6. Mention, stub, and entity-origin contract` | Core 02 §6 | Appendix G | — | history only |
| `6. Required provenance` | Core 02 §7 | Appendix G | — | history only |
| `6. Deduplication and auto-upsert rules` | Core 02 §7.3 and §8 | Appendix G | — | history only |
| `6. Merge behavior` | Core 01 §3.3.3 and §3.3.5.4, Core 02 §7.3, §8, §9, and §15.4, Core 03 §5 and §16.2, Core 04 §9.1 and §9.10 | Appendix G | — | history only |
| `7. Additional schema requirements for rollback granularity` | Core 00 §8.1, Core 02 §14-15 | Appendix C, Appendix G | — | history only |
| History retention and rollback horizon | Core 01 §3.3.4.2 and §12.3, Core 02 §15.3.1, Core 04 §9.0.1, §9.0.6, §9.1, and §9.11 | Appendix E | — | history only |
| Destructive-operation concurrency and public contention failure for restore, rollback, and merge | Core 00 §5.1, Core 01 §3.3.5.0, Core 03 §5, Core 04 §9.1 | Appendix C, Appendix D, Appendix E | — | history only |
| Timeline supersede replacement relation | Core 00 §5.1, Core 01 §3.3.5 and §7.4.1, Core 02 §12.1-§12.3, Core 03 §6, Core 04 §9.1 and §9.11 | Appendix C, Appendix D, Appendix E | — | history only |
| `5. Same-field conflict resolution UX` | Core 03 §3.3, Core 04 §9.6 | Appendix E | — | illustration only |
| Performance benchmark environment and reproducibility | Core 00 §5.1, Core 05 | Appendix B, Appendix D, Appendix E | — | history only |
| Cross-cutting client/server interface surface for workbook, evidence, and jobs | Core 01 §3.3.2, §3.3.4, §3.3.5, and §3.3.10.1, Core 03 §3.3.4, §4.3, §4.3.1, §4.4, §8, and §15, Core 04 §1, §2, §4.5, and §9.10 | Appendix E | — | owner-local contract tables now densify the boundary in Core 01; appendices remain navigation only |
| `16. Open question on local pending queue behavior` | Core 03 §4.2 and §4.4, Core 04 §9.1 | Appendix B, Appendix D, Appendix E | — | history only |
| `8. Bulk paste/import from existing spreadsheet or clipboard` | Core 01 §17.2, Core 03 §11 | Appendix D, Appendix G | — | history only |
| `8. Auto-resolution policy for typed host/account strings` | Core 03 §12 | Appendix D, Appendix G | — | history only |
| `9. Sorting / filtering / grouping` | Core 00 §5.1, Core 01 §3.3.4 and §7.4, Core 02 §16, Core 03 §14 | Appendix D, Appendix G | — | history only |
| `9. How denormalized timeline views are composed` | Core 01 §7.4 and Core 03 §15 | Appendix D, Appendix G | — | illustration only |
| `9. How denormalized entity/evidence views are composed` | Core 01 §7.4, §18, and §19, Core 02 §8.2, §18, and §19, Core 03 §20, Core 04 §9.1 | Appendix D, Appendix E, Appendix G | — | illustration only |
| `11. Reference-pack and export trust boundaries` | Core 04 §4 | Appendix G | — | history only |
| `11. Local users, incident roles, and admin surfaces` | Core 01 §3.3.2, §3.3.3, §3.3.5.1, and §18, Core 02 §3 and §14.1, Core 04 §1-§3 and §9.10, §9.12, and §12 | Appendix C, Appendix E, Appendix G | — | history only |
| Current-account profile and density preference boundary | Core 00 §5.1, Core 01 §3.3.2.3, Core 02 §14.1, Core 03 §2.4, Core 04 §1.1.1, §2, §3, and §9.10 | Appendix C, Appendix D, Appendix E | `docs/domain.md`; `docs/design.md`; UI/UX guide | owner-level self-service account and density closure only |
| Inspector feature-group registry and row-context workflow routing | Core 00 §5.1; Core 01 §7.4; Core 02 §2; Core 03 §2.3A; Core 04 §2 and §9 | Appendix D, Appendix E, Appendix H | `docs/domain.md`; `docs/design.md`; UI/UX guide; frontend guide; visual golden guide; dev guide | Core 01 owns emitted config and feature registry. Core 03 owns interaction algorithms. Core 02 owns no-source-state boundary. Core 04 owns authorization and egress. Appendices and guides are usage, design, implementation-support, or traceability only. |
| Deployment-admin application-level capability and public-route authorization closure | Core 04 §2 and §9.10; Core 01 §3.3.3, §17, and §20 | Appendix E | `docs/domain.md`; UI/UX guide; developer guide | owner-level authorization matrix and reference-pack route-family traceability only |
| Administrative audit read projections and deployment-local audit retention | Core 00 §5.1, Core 01 §3.3.5.1A and §12, Core 02 §14.1, Core 04 §2, §3, and §9.10 | Appendix C, Appendix D, Appendix E | Appendix H; UI/UX guide | owner-level route, resource, authorization, redaction, retention, backup, and portability closure only |
| Shared authenticated post-login landing and visible-incident directory selection | Core 00 §5.1, Core 01 §3.3.2.1A and §3.3.5.3.1, Core 03 §2.4, Core 04 §9.10 | Appendix D, Appendix E | UI/UX guide | owner navigation and composition only |
| Deployment administration browser context and prohibited aggregate administration concepts | Core 00 §5.1, Core 01 §3.3.2.1B, Core 03 §2.4, Core 04 §2 and §9.10 | Appendix D, Appendix E | `docs/design.md`; UI/UX guide | owner-level browser context, panel eligibility, and no all-incident catalog or generic settings closure only |
| Incident-bundle import initial-admin bootstrap membership | Core 00 §5.1, Core 01 §12.3.6 and §17.5, Core 03 §2.4, Core 04 §9.11 | Appendix D, Appendix E | Appendix H; UI/UX guide | target-local membership bootstrap and import-completion navigation only |
| Reference Pack list-query search and exact filters | Core 00 §5.1, Core 01 §17.4, Core 04 §9.4 | Appendix D, Appendix E | `docs/design.md`; UI/UX guide | owner-level complete-collection search/filter and request sequencing only |
| Deployment topology, application-unit boundary, and authoritative service separation | Core 00 §5.1, Core 01 §1, Core 04 §5.2-§8 | Appendix B | — | illustration only |
| `4. Backup, restore, portability, failure modes` - operational backup, backup creation admission/publication, coherent restore, and recovery operator interface | Core 00 §5.1, Core 01 §12.1-§12.2, Core 04 §2, §4.4, §4.5, §6, §9.0.1, §9.14, and §12 | Appendix B, Appendix E | Appendix H; development guide | illustration and operator-practice guidance only |
| `4. Backup, restore, portability, failure modes` - incident portability | Core 00 §4.2 and §6, Core 01 §12.3, Core 04 §4.2 and §9.11 | Appendix E, Appendix G | — | illustration only |
| Deployment configuration surface | Core 00 §5.1, Core 01 §14, Core 04 §6, §9.12, and §12 | Appendix B, Appendix E | — | history only |
| Resource-limit registry and numeric safety boundaries | Core 00 §5.1, Core 01 §3.3.8, §16, §17.2, §17.4, and §12.3.6, Core 04 §9.13 and §12.3.1 | Appendix B, Appendix E | — | history only |
| `16. Open question on minimum disconnected reference-pack bundle and update flow` | Core 01 §5 and §11, Core 02 §4.1, §14.1, and §17, Core 04 §4.1, §5.1, and §9.4 | Appendix C, Appendix E | — | history only |
| Reference-pack durable version conditions and verification/activation lifecycle owner boundary | Core 00 §5.1, Core 01 §11.3, §11.3.1, §11.4, and §11.4.1, Core 02 §14.1 and §17, Core 04 §9.4 | Appendix C | — | history only |
| `16. Open questions on snapshot release controls and generated presentation depth` | Core 01 §10, Core 02 §10.5 and §14.5, Core 04 §2.1, §4.2, and §9.3 | Appendix E | — | history only |
| `16. Open question on restricted evidence visibility` | Core 01 §10.5, Core 02 §10.5, §13, and §14.5, Core 04 §2, §4.2, and §9.3 | Appendix E | — | history only |
| `16. Open question on clipboard paste versus XLSX adoption` | Core 01 §2, Core 03 §11, Core 04 §4.5 and §9.2 | Appendix E | — | history only |
| `16. Open question on dedicated Notes tab and note-capture shape` | Core 01 §7.1, §7.3, and §7.4, Core 02 §10.1 and §12, Core 04 §9.1 and §9.3 | Appendix E | — | history only |
| Artifact-backed note/coordination/finding tagged-variant registry | Core 02 §10.4.4A; Core 01 §7.4 and §19; Core 03 §2.1-§2.2; Core 04 §9.1 | Appendix E | — | history only |
| Cross-layer workbook-surface identity mapping table | Core 01 §7.4; Core 02 §10.4.4A-§10.4.6; Core 03 §2 | — | — | owner-level normalization only |
| Workbook public-surface, source-record, projection/query, saved-view, revision/conflict, and collaboration-publisher owner vocabulary | Core 01 §7.4; Core 02 §10.4.4; Core 03 §2.2 and §3.3 | — | `docs/domain.md`; workbook refactor handoff | owner-boundary clarification only |
| Projection authority, provider descriptors, query characterization, restore rebuild, and import guardrails | Core 00 §5.1; Core 01 §3.3.4, §8, and §12.2; Core 04 §9.1A | Appendix I | `docs/domain.md`; projections refactor tracker | owner-level projection authority and boundary closure; appendix evidence only |
| `16. Open question on shared/private saved views and startup defaults` | Core 01 §3.3.3 and §3.3.5.2, Core 02 §11.1-§11.2 and §14.1, Core 03 §2.3-§2.4 and §14.3-§14.7, Core 04 §2, §9.1, and §9.10 | Appendix E | — | history only |
| `Saved-view object grammar and public discovery-resource closure` | Core 01 §3.3.5.2, §6, and §7.4, Core 02 §4.3 and §11.1, Core 03 §2.3 and §14, Core 04 §9.1 and §9.10 | Appendix C, Appendix E | — | history only |
| Incident lifecycle close/reopen and closed-incident source-state boundary | Core 00 §5.1, Core 01 §3.3.5.3.2 and §3.3.10, Core 02 §18, Core 03 §4.3.2 and §4.4, Core 04 §9.10 | Appendix C, Appendix D, Appendix E | `docs/domain.md`; UI/UX guide | route, migration, operation-boundary, race, and UI traceability only |
| `16. Open question on incident-scoped party identity, dedupe, and coordination references` | Core 01 §7.2, §7.4, §18, and §19, Core 02 §2, §4.5, §8.2-§8.3, §10.4, §13, §14.1, §18, and §19, Core 03 §2.2, §16.2, §16.4, and §20, Core 04 §2, §9.0, and §9.1 | Appendix C, Appendix D, Appendix E | — | history only |
| `16. Open question on analyst-work tracking beyond tags and notes` | Core 00 §4.3 and §6, Core 01 §2, §7.2, §7.4, §8.1, §8.5, and §10.3, Core 02 §2, §3, §10.1, §10.4, §12-§14, Core 03 §2.2, §16.4, and §19, Core 04 §2, §4.2, §9.1, and §9.3 | Appendix E | Appendix H | history only |
| `16. Open question on indicator storage promotion` | Core 01 §7.2, §7.4, §8.1, and §8.5, Core 02 §2, §7.4, §10.2, §12, and §14, Core 03 §2.2, §7, §9.1, §15, and §16.2, Core 04 §9.1 | Appendix B, Appendix C, Appendix D, Appendix E | — | history only |
| `16. Open question on assessment vocabulary and confidence model` | Core 01 §7.2, §7.4, and §8.1-§8.5, Core 02 §10.3 and §14.1-§14.3, Core 03 §2.2 and §16.3, Core 04 §9.1 | Appendix C, Appendix E | — | history only |
| `16. Open question on timeline grouping-key whitelist` | Core 03 §14.3, Core 04 §9.1 | Appendix E | — | history only |
| `16. Open question on incident-specific custom metadata promotion` | Core 01 §7.2, §8.5, and §19, Core 02 §4.1-§4.5, §10.4, §13, §14.1, §18, and §19, Core 03 §2.2, §16.1-§16.4, and §20, Core 04 §9.1 and §9.8 | Appendix E | — | history only |
| `16. Open question on writable-string boundary completeness` | Core 01 §3.3.5, §7.4, and §18, Core 04 §9.1 and §9.10 | Appendix C, Appendix D, Appendix E | — | history only |
| `16. Open question on direct-scalar timestamp boundary completeness` | Core 00 §5.1, Core 01 §3.3.5, §7.4, §18A, and §19, Core 03 §13.1, Core 04 §9.0 and §9.1 | Appendix C, Appendix D, Appendix E | — | history only |
| `16. Open question on nullable direct-reference scalar boundary completeness` | Core 00 §5.1, Core 01 §3.3.5, §7.4, §18B, and §19, Core 02 §10.4.1 and §19, Core 03 §16.2, §16.4, and §20, Core 04 §9.0 and §9.1 | Appendix C, Appendix D, Appendix E | — | history only |
| Define-once create-time defaults and initial-value ownership | Core 00 §2 and §5.1, Core 01 §7.4 and §19, Core 02 §10.4.1.1, §10.4.4, and §10.4.6 | Appendix E | — | history only |
| Manual-link confidence default and client-supplied relationship confidence | Core 01 §7.4, Core 02 §12.3, Core 03 §13.1, Core 04 §9.0.1 and §9.1 | Appendix D, Appendix E | — | history only |
| Extension route-family public-contract parity | Core 00 §5.1, Core 01 §17 and §20, Core 04 §9.0, §9.2, §9.3, §9.4, §9.5, and §9.11 | Appendix E | — | owner-local route, resource, and error tables now carry the compact extension-family contract; appendices remain history and navigation only |
| Enterprise-auth identity-binding lifecycle and deployment-admin binding surface | Core 00 §5.1, Core 01 §20, Core 02 §14.1, Core 04 §1.2, §2, §3, and §9.5 | Appendix C, Appendix E | — | Core 01 §20 now carries owner-local protocol and binding-management tables; appendices remain history and realization only |
| Enterprise-auth provider deployment configuration and reconciliation | Core 00 §5.1, Core 01 §20, Core 02 §14.1, Core 04 §1.2, §9.5, §12.3.4, and §12.6 | Appendix B, Appendix C, Appendix D, Appendix E | developer guide; implementation testing guide; frontend testing guide | provider definitions are startup-only deployment-local configuration with Core-owned manifest validation, reconciliation, safe discovery, and no runtime provider mutation |

## F.3 Completeness note

Appendix F is navigation only.

Current-profile behavior is controlled by Core 00 through Core 05. Appendix E preserves historical source-question context, future-only backlog, and editorial-hardening support. Appendix H preserves non-normative operating-model guidance for operator practice such as tracker hygiene, handoff quality, status-review cadence, debrief discipline, workload redistribution, and challenge/escalation practice. Appendix I preserves projection authority, boundary, and characterization evidence only.

No appendix text is authoritative for current-profile runtime behavior unless the same behavior is restated in the normative core.

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
| REQ-00-013 | Core 00 §7 Supported operating envelope | base, reference_pack | AC-043..AC-046, AC-231, AC-234 |
| REQ-00-014 | Core 00 §8.1 Core identifiers | base | AC-116, AC-118, AC-123..AC-125, AC-127..AC-129, AC-231 |
| REQ-00-015 | Core 00 §9 Global invariants | base, snapshot_reporting, reference_pack | AC-231, AC-233..AC-234 |
| REQ-00-016 | Core 00 §9.1 Lifecycle state-machine notation | base | AC-107..AC-111, AC-231, AC-313 |
| REQ-00-017 | Core 00 §9.1 Lifecycle state-machine notation | base | AC-107..AC-111, AC-231 |
| REQ-00-019 | Core 00 §9.1 Lifecycle state-machine notation | base | AC-231, AC-313 |
| REQ-00-020 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-021 | Core 00 §5.1 Contract-owner matrix | enterprise_authentication | AC-235 |
| REQ-00-022 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-023 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-024 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-025 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-026 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-027 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-028 | Core 00 §5.1 Contract-owner matrix | claim_publication | PC-006 |
| REQ-00-029 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-030 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-031 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-032 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-033 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-034 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-035 | Core 00 §5.1 Contract-owner matrix | base | AC-099, AC-170, AC-212, AC-214, AC-231, AC-418..AC-426 |
| REQ-00-036 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-037 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-038 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-039 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-040 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-041 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-042 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-043 | Core 00 §5.1 Contract-owner matrix | reference_pack | AC-234 |
| REQ-00-044 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-045 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-046 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-047 | Core 00 §5.1 Contract-owner matrix | import, snapshot_reporting, incident_portability, reference_pack | AC-232, AC-233, AC-234, AC-236 |
| REQ-00-048 | Core 00 §5.1 Contract-owner matrix | enterprise_authentication | AC-235 |
| REQ-00-049 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-050 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-051 | Core 00 §1 Status | base | AC-231 |
| REQ-00-052 | Core 00 §5.1 Contract-owner matrix | base | AC-231 |
| REQ-00-053 | Core 00 §5.1 Contract-owner matrix | base | AC-414 |
| REQ-00-054 | Core 00 §5.1 Contract-owner matrix | base | AC-429..AC-432 |
| REQ-00-055 | Core 00 §5.1 Contract-owner matrix | enterprise_authentication | AC-433..AC-436 |
| REQ-00-056 | Core 00 §5.1 Contract-owner matrix | base | AC-437..AC-440 |
| REQ-00-057 | Core 00 §5.1 Contract-owner matrix | base | AC-414, AC-427, AC-441 |
| REQ-00-058 | Core 00 §5.1 Contract-owner matrix | incident_portability | AC-442 |
| REQ-00-059 | Core 00 §5.1 Contract-owner matrix | reference_pack | AC-443 |
| REQ-00-062 | Core 00 §1 Status | base | AC-469 |
| REQ-00-063 | Core 00 §5.1 Contract-owner matrix | base | AC-469..AC-473 |
| REQ-01-001 | Core 01 §1 Architecture pattern | base | AC-231, AC-404 |
| REQ-01-002 | Core 01 §1 Architecture pattern | base | AC-231, AC-404, AC-405 |
| REQ-01-003 | Core 01 §1 Architecture pattern | base | AC-231, AC-404 |
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
| REQ-01-021 | Core 01 §3.3.1 Versioning and compatibility | base | AC-124..AC-125, AC-127, AC-131, AC-135, AC-219..AC-220, AC-231 |
| REQ-01-022 | Core 01 §3.3.1 Versioning and compatibility | base | AC-124..AC-125, AC-127, AC-131, AC-135, AC-231 |
| REQ-01-023 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-024 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-025 | Core 01 §3.3.2 Session and authentication routes | base | AC-123, AC-130, AC-156..AC-162, AC-231, AC-244..AC-250, AC-311, AC-414 |
| REQ-01-026 | Core 01 §3.3.2.1 Session resource and expiry contract | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-027 | Core 01 §3.3.2.1 Session resource and expiry contract | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-028 | Core 01 §3.3.2.1 Session resource and expiry contract | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-029 | Core 01 §3.3.2.1 Session resource and expiry contract | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-030 | Core 01 §3.3.2.1 Session resource and expiry contract | base | AC-123, AC-130, AC-156..AC-162, AC-231 |
| REQ-01-031 | Core 01 §3.3.2.1 Session resource and expiry contract | base, enterprise_authentication | AC-123, AC-130, AC-156..AC-162, AC-231, AC-235, AC-250, AC-290..AC-291 |
| REQ-01-580 | Core 01 §3.3.2.1A Authenticated root landing contract | base | AC-289..AC-291, AC-414 |
| REQ-01-032 | Core 01 §3.3.3 Route families | base | AC-175..AC-180, AC-186..AC-187, AC-231, AC-251..AC-255, AC-418, AC-427, AC-429..AC-432, AC-437..AC-439 |
| REQ-01-033 | Core 01 §3.3.3 Route families | base, import, snapshot_reporting, incident_portability, reference_pack, enterprise_authentication | AC-175..AC-180, AC-186..AC-187, AC-231..AC-236, AC-427 |
| REQ-01-034 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-366..AC-368 |
| REQ-01-035 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-238..AC-240, AC-243, AC-372..AC-375, AC-387 |
| REQ-01-036 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-238, AC-241, AC-243, AC-366..AC-367, AC-372..AC-374, AC-387 |
| REQ-01-037 | Core 01 §3.3.4 View-shaped read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-243, AC-387 |
| REQ-01-038 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-360, AC-387 |
| REQ-01-039 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-040 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-041 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-042 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-043 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-044 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-045 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-046 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-360..AC-361, AC-387 |
| REQ-01-047 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-387 |
| REQ-01-048 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-049 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-050 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-051 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-052 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-053 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-054 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231, AC-383..AC-384 |
| REQ-01-055 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-231 |
| REQ-01-056 | Core 01 §3.3.4.2 Record-history read contract | base | AC-124, AC-127, AC-184..AC-185, AC-215, AC-231 |
| REQ-01-057 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-058 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-299 |
| REQ-01-059 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-060 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-061 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-062 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-063 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-299 |
| REQ-01-064 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-065 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-066 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-067 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-068 | Core 01 §3.3.5 Mutation contract | base, snapshot_reporting | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-233 |
| REQ-01-069 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-299 |
| REQ-01-070 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231, AC-299 |
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
| REQ-01-086 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-196, AC-200..AC-218, AC-221..AC-225, AC-231, AC-329, AC-331 |
| REQ-01-087 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-199..AC-200, AC-201..AC-218, AC-221..AC-225, AC-231, AC-329..AC-330 |
| REQ-01-088 | Core 01 §3.3.5 Mutation contract | base | AC-125..AC-126, AC-181..AC-183, AC-188..AC-190, AC-200..AC-218, AC-221..AC-225, AC-231 |
| REQ-01-089 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-090 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-091 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-092 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-093 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-094 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-095 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-096 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231, AC-384 |
| REQ-01-097 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-098 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base, snapshot_reporting | AC-215..AC-218, AC-231, AC-233 |
| REQ-01-099 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-100 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-101 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-102 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-103 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-104 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-182, AC-187, AC-215..AC-218, AC-231, AC-353 |
| REQ-01-105 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-106 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-107 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-108 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-109 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-110 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-111 | Core 01 §3.3.5.0 Destructive-operation concurrency and rollback contract | base | AC-215..AC-218, AC-231 |
| REQ-01-112 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-113 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-114 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-115 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-116 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231, AC-348, AC-351 |
| REQ-01-117 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-127, AC-175..AC-180, AC-231, AC-417 |
| REQ-01-118 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-127, AC-175..AC-180, AC-231 |
| REQ-01-119 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231, AC-312 |
| REQ-01-120 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231, AC-244..AC-245, AC-312 |
| REQ-01-121 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231, AC-343..AC-347 |
| REQ-01-122 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231, AC-312 |
| REQ-01-123 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231 |
| REQ-01-124 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-175..AC-180, AC-231, AC-312 |
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
| REQ-01-142 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231, AC-360 |
| REQ-01-143 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231 |
| REQ-01-144 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-127, AC-146..AC-153, AC-231 |
| REQ-01-145 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231, AC-360 |
| REQ-01-146 | Core 01 §3.3.5.2 View-schema, saved-view, and workbook-preference contracts | base | AC-146..AC-153, AC-231, AC-360 |
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
| REQ-01-157 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-099, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-158 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-099, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-159 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-160 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-161 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-162 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-163 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-164 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-165 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-166 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-167 | Core 01 §3.3.5.3 Incident resource and creation contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-168 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-127, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231, AC-414, AC-416, AC-424 |
| REQ-01-169 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-170 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-171 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-172 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-173 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-099, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-174 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-099, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-175 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-099, AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-176 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-177 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-178 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-179 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-180 | Core 01 §3.3.5.3.1 Incident list, retrieval, and metadata patch contract | base | AC-170..AC-174, AC-211..AC-214, AC-219..AC-220, AC-231 |
| REQ-01-585 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base | AC-418, AC-422, AC-425 |
| REQ-01-586 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base | AC-418, AC-421 |
| REQ-01-587 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base | AC-421 |
| REQ-01-588 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base | AC-419..AC-421, AC-425 |
| REQ-01-589 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base | AC-419..AC-421 |
| REQ-01-590 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base, import, snapshot_reporting, incident_portability | AC-422, AC-424 |
| REQ-01-591 | Core 01 §3.3.5.3.2 Incident lifecycle close/reopen contract | base, import | AC-421..AC-423 |
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
| REQ-01-199 | Core 01 §3.3.5.5 Entity-mention action contract | base | AC-019..AC-021, AC-188..AC-190, AC-201, AC-221..AC-225, AC-231 |
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
| REQ-01-234 | Core 01 §3.3.6.1 Canonical public error-code registry | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231, AC-239..AC-240, AC-245..AC-247, AC-249..AC-255, AC-260..AC-261, AC-293, AC-321, AC-323..AC-326, AC-328, AC-415..AC-418, AC-421..AC-422, AC-427, AC-438 |
| REQ-01-235 | Core 01 §3.3.6.1 Canonical public error-code registry | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-236 | Core 01 §3.3.6.1 Canonical public error-code registry | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-237 | Core 01 §3.3.6.1 Canonical public error-code registry | base | AC-126, AC-187, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-238 | Core 01 §3.3.6.2 Canonical public reason-code registries | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231, AC-239..AC-240, AC-252, AC-255, AC-260, AC-293, AC-321..AC-328, AC-375, AC-415..AC-418, AC-421, AC-438 |
| REQ-01-239 | Core 01 §3.3.6.2 Canonical public reason-code registries | base | AC-126, AC-203..AC-208, AC-211, AC-213..AC-214, AC-218..AC-219, AC-231 |
| REQ-01-240 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238..AC-240, AC-415..AC-417, AC-438 |
| REQ-01-241 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-239, AC-416..AC-417, AC-438 |
| REQ-01-242 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-116, AC-127, AC-151, AC-171, AC-175, AC-178, AC-215, AC-231, AC-238..AC-239, AC-241..AC-242, AC-416..AC-417, AC-438 |
| REQ-01-243 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-321 |
| REQ-01-244 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-245 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-321 |
| REQ-01-246 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-01-247 | Core 01 §3.3.8 Evidence and blob routes | base | AC-015..AC-016, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-251..AC-255 |
| REQ-01-248 | Core 01 §3.3.9 Background-job routes | base | AC-046, AC-129, AC-231, AC-257 |
| REQ-01-249 | Core 01 §3.3.9 Background-job routes | base | AC-046, AC-129, AC-231, AC-257..AC-261 |
| REQ-01-250 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-251 | Core 01 §3.3.10 WebSocket collaboration stream | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-252 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-253 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-254 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-255 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-256 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-257 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-258 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-259 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-260 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-261 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-262 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-263 | Core 01 §3.3.10.1 v1 collaboration wire contract | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-264 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-265 | Core 01 §3.3.10.1 v1 collaboration wire contract | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-266 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-267 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-368 |
| REQ-01-268 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-257..AC-259 |
| REQ-01-269 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-270 | Core 01 §3.3.10.1 v1 collaboration wire contract | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-271 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-272 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-273 | Core 01 §3.3.10.1 v1 collaboration wire contract | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-274 | Core 01 §3.3.10.1 v1 collaboration wire contract | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231, AC-233 |
| REQ-01-275 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-276 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-277 | Core 01 §3.3.10.1 v1 collaboration wire contract | base | AC-129, AC-131..AC-136, AC-156..AC-163, AC-231 |
| REQ-01-592 | Core 01 §3.3.10 WebSocket collaboration stream | base | AC-419, AC-426 |
| REQ-01-278 | Core 01 §4.1 Postgres | base, snapshot_reporting, reference_pack | AC-231, AC-233..AC-234, AC-405 |
| REQ-01-279 | Core 01 §4.2 Object storage | base, snapshot_reporting | AC-231, AC-233, AC-405 |
| REQ-01-280 | Core 01 §4.3 Storage exclusions | base | AC-231, AC-405 |
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
| REQ-01-296 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-116, AC-121..AC-122, AC-231, AC-277 |
| REQ-01-297 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-298 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-299 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-300 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-301 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-01-302 | Core 01 §7.2 Contract-backed system views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231, AC-281..AC-284 |
| REQ-01-303 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-304 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-305 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-306 | Core 01 §7.3 Notes sheet contract | base | AC-068..AC-070, AC-112, AC-185, AC-231 |
| REQ-01-307 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-116..AC-122, AC-124..AC-125, AC-231, AC-281..AC-284 |
| REQ-01-308 | Core 01 §7.4 Authoritative base-profile view schema registry | base, reference_pack | AC-116..AC-122, AC-124..AC-125, AC-231, AC-234, AC-285..AC-287 |
| REQ-01-579 | Core 01 §7.4 Authoritative base-profile view schema registry | base, reference_pack | AC-411 |
| REQ-01-309 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-116..AC-122, AC-124..AC-125, AC-231, AC-281..AC-287, AC-410 |
| REQ-01-310 | Core 01 §7.4 Authoritative base-profile view schema registry | base, reference_pack | AC-116..AC-122, AC-124..AC-125, AC-184..AC-185, AC-231, AC-234, AC-281..AC-287, AC-300..AC-303, AC-366..AC-368 |
| REQ-01-311 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-116..AC-122, AC-124..AC-125, AC-196, AC-231, AC-281..AC-284, AC-331..AC-332 |
| REQ-01-312 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-118, AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231, AC-300..AC-301, AC-303, AC-331..AC-332, AC-444..AC-449, AC-452 |
| REQ-01-313 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231, AC-444, AC-448, AC-452 |
| REQ-01-314 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231, AC-444, AC-448, AC-452 |
| REQ-01-315 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-118, AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231, AC-388..AC-391, AC-448 |
| REQ-01-316 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-118, AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231, AC-388..AC-391, AC-448 |
| REQ-01-317 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-318 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-319 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-320 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-321 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-322 | Core 01 §7.4.1 `cartulary.view.timeline.v2` | base | AC-119, AC-124..AC-125, AC-184, AC-191..AC-198, AC-231 |
| REQ-01-323 | Core 01 §7.4.2 `cartulary.view.hosts.v1` | base | AC-097, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-324 | Core 01 §7.4.2 `cartulary.view.hosts.v1` | base | AC-097, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-325 | Core 01 §7.4.2 `cartulary.view.hosts.v1` | base | AC-097, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-326 | Core 01 §7.4.3 `cartulary.view.identities.v1` | base | AC-098, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-327 | Core 01 §7.4.3 `cartulary.view.identities.v1` | base | AC-098, AC-118, AC-124..AC-125, AC-231 |
| REQ-01-328 | Core 01 §7.4.4 `cartulary.view.evidence.v1` | base | AC-100, AC-118, AC-124..AC-125, AC-128, AC-231, AC-278..AC-280, AC-300..AC-301, AC-303, AC-315..AC-318 |
| REQ-01-329 | Core 01 §7.4.5 `cartulary.view.notes.v1` | base | AC-068..AC-070, AC-112, AC-118, AC-124..AC-125, AC-185, AC-231 |
| REQ-01-330 | Core 01 §7.4.5 `cartulary.view.notes.v1` | base | AC-068..AC-070, AC-112, AC-118, AC-124..AC-125, AC-185, AC-231 |
| REQ-01-331 | Core 01 §7.4.6 `cartulary.view.indicators.v1` | base | AC-017, AC-072..AC-079, AC-118, AC-122, AC-124, AC-231 |
| REQ-01-332 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231, AC-300, AC-302..AC-303 |
| REQ-01-333 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-334 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-335 | Core 01 §7.4.7 `cartulary.view.assessments.v1` | base | AC-018, AC-080..AC-084, AC-118, AC-121, AC-124, AC-231 |
| REQ-01-336 | Core 01 §7.4.8 `cartulary.view.task_requests.v1` | base | AC-085, AC-118, AC-124, AC-137..AC-140, AC-145, AC-231, AC-278..AC-280, AC-300..AC-301, AC-303..AC-304, AC-315..AC-319 |
| REQ-01-337 | Core 01 §7.4.8 `cartulary.view.task_requests.v1` | base | AC-085, AC-118, AC-124, AC-137..AC-140, AC-145, AC-231 |
| REQ-01-338 | Core 01 §7.4.8 `cartulary.view.task_requests.v1` | base | AC-085, AC-118, AC-124, AC-137..AC-140, AC-145, AC-231, AC-304 |
| REQ-01-339 | Core 01 §7.4.9 `cartulary.view.decisions.v1` | base | AC-086, AC-118, AC-124, AC-141..AC-145, AC-231, AC-300, AC-302..AC-303 |
| REQ-01-340 | Core 01 §7.4.9 `cartulary.view.decisions.v1` | base | AC-086, AC-118, AC-124, AC-141..AC-145, AC-231 |
| REQ-01-341 | Core 01 §7.4.9 `cartulary.view.decisions.v1` | base | AC-086, AC-118, AC-124, AC-141..AC-145, AC-231 |
| REQ-01-342 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-343 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231, AC-277 |
| REQ-01-344 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-345 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-346 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-347 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-348 | Core 01 §8.1 Projection tables | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-349 | Core 01 §8.2 Projection-row identity | base | AC-124, AC-125, AC-231 |
| REQ-01-350 | Core 01 §8.2 Projection-row identity | base | AC-013, AC-125, AC-231 |
| REQ-01-351 | Core 01 §8.3 Projection maintenance | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-352 | Core 01 §8.3 Projection maintenance | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-353 | Core 01 §8.3 Projection maintenance | base | AC-032, AC-046, AC-210, AC-231 |
| REQ-01-354 | Core 01 §8.4 Projection corruption | base | AC-231 |
| REQ-01-355 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-356 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-357 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231 |
| REQ-01-358 | Core 01 §8.5 Hot-path retrieval and evidence boundary | base | AC-015..AC-017, AC-045, AC-053..AC-054, AC-100, AC-128, AC-210, AC-231, AC-281..AC-287 |
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
| REQ-01-377 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233, AC-333 |
| REQ-01-378 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233, AC-333 |
| REQ-01-379 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233, AC-333 |
| REQ-01-380 | Core 01 §10.3 Export-model classification and release scopes | snapshot_reporting | AC-057, AC-059..AC-062, AC-071, AC-091, AC-113..AC-115, AC-233, AC-333 |
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
| REQ-01-432 | Core 01 §12.3.2 Authoritative-state boundary | incident_portability | AC-164, AC-167, AC-236, AC-440 |
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
| REQ-01-448 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236, AC-327..AC-328, AC-332 |
| REQ-01-449 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236, AC-327..AC-328, AC-332 |
| REQ-01-450 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-165..AC-167, AC-169, AC-236 |
| REQ-01-451 | Core 01 §12.4 Failure handling | base | AC-166, AC-231 |
| REQ-01-452 | Core 01 §13 Long-running operations and background jobs | base, snapshot_reporting, reference_pack | AC-030, AC-033, AC-046, AC-129, AC-169, AC-231, AC-233..AC-234 |
| REQ-01-453 | Core 01 §13 Long-running operations and background jobs | base | AC-030, AC-033, AC-046, AC-129, AC-169, AC-231, AC-258, AC-260 |
| REQ-01-454 | Core 01 §13 Long-running operations and background jobs | base | AC-030, AC-033, AC-046, AC-129, AC-169, AC-231 |
| REQ-01-455 | Core 01 §14 Runtime roots and packaging | base, reference_pack | AC-051, AC-055, AC-169, AC-231, AC-234, AC-294..AC-295, AC-297 |
| REQ-01-456 | Core 01 §14 Runtime roots and packaging | base | AC-051, AC-055, AC-169, AC-231, AC-296 |
| REQ-01-457 | Core 01 §15 Architecture invariants | base, import, snapshot_reporting, incident_portability | AC-231..AC-233, AC-236 |
| REQ-01-458 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-252..AC-254 |
| REQ-01-459 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-251, AC-255 |
| REQ-01-460 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-252..AC-253, AC-256 |
| REQ-01-461 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-252, AC-322 |
| REQ-01-462 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-253..AC-254 |
| REQ-01-463 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-254..AC-255 |
| REQ-01-464 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-256 |
| REQ-01-465 | Core 01 §16 Evidence-access handle contract | base | AC-231, AC-251..AC-255, AC-322 |
| REQ-01-466 | Core 01 §17.1 Common parity rules | import, snapshot_reporting, incident_portability, reference_pack | AC-262..AC-276 |
| REQ-01-467 | Core 01 §17.1 Common parity rules | import, snapshot_reporting, incident_portability, reference_pack | AC-262, AC-264, AC-266..AC-268, AC-270..AC-271, AC-273..AC-275, AC-309, AC-369 |
| REQ-01-468 | Core 01 §17.1 Common parity rules | import, snapshot_reporting, incident_portability, reference_pack | AC-263, AC-266, AC-270, AC-274 |
| REQ-01-469 | Core 01 §17.1 Common parity rules | import, snapshot_reporting, incident_portability, reference_pack | AC-262..AC-264, AC-266..AC-268, AC-270..AC-271, AC-273..AC-275, AC-305, AC-308, AC-369 |
| REQ-01-470 | Core 01 §17.1 Common parity rules | import, snapshot_reporting, incident_portability, reference_pack | AC-262, AC-264, AC-266..AC-268, AC-270..AC-271, AC-273, AC-275, AC-305, AC-308, AC-369 |
| REQ-01-471 | Core 01 §17.1 Common parity rules | import, snapshot_reporting, incident_portability, reference_pack | AC-265, AC-269, AC-272, AC-276, AC-307, AC-310, AC-427 |
| REQ-01-542 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-370, AC-427 |
| REQ-01-543 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-370, AC-427 |
| REQ-01-544 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-370..AC-371, AC-427 |
| REQ-01-545 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-370, AC-427 |
| REQ-01-546 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-370..AC-371, AC-427 |
| REQ-01-547 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-371, AC-427 |
| REQ-01-548 | Core 01 §3.3.3.1 Runtime extension discovery and reserved-unclaimed extension semantics | base | AC-371, AC-427 |
| REQ-01-549 | Core 01 §17.1.1 Shared upload-envelope contract for upload-style extension routes | import, incident_portability, reference_pack | AC-262, AC-270, AC-275 |
| REQ-01-550 | Core 01 §17.1.1 Shared upload-envelope contract for upload-style extension routes | import, incident_portability, reference_pack | AC-262, AC-270, AC-275 |
| REQ-01-551 | Core 01 §17.1.1 Shared upload-envelope contract for upload-style extension routes | import, incident_portability, reference_pack | AC-262, AC-270, AC-275 |
| REQ-01-552 | Core 01 §17.1.1 Shared upload-envelope contract for upload-style extension routes | import, incident_portability, reference_pack | AC-262, AC-270, AC-275 |
| REQ-01-553 | Core 01 §17.1.1 Shared upload-envelope contract for upload-style extension routes | import, incident_portability, reference_pack | AC-262, AC-265, AC-270, AC-272, AC-275, AC-276 |
| REQ-01-554 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-372..AC-375 |
| REQ-01-555 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-372..AC-374 |
| REQ-01-556 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-372..AC-373 |
| REQ-01-557 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-373..AC-374 |
| REQ-01-558 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-374..AC-375 |
| REQ-01-559 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-375 |
| REQ-01-560 | Core 01 §3.3.7 Pagination and cursor contract | base | AC-231, AC-372..AC-375 |
| REQ-01-581 | Core 01 §3.3.7.1 GET collection list-query contract | base | AC-415..AC-417 |
| REQ-01-582 | Core 01 §3.3.7.1 GET collection list-query contract | base | AC-415..AC-417 |
| REQ-01-583 | Core 01 §3.3.7.1 GET collection list-query contract | base | AC-415..AC-417, AC-438 |
| REQ-01-584 | Core 01 §3.3.7.1 GET collection list-query contract | base | AC-416..AC-417, AC-438 |
| REQ-01-561 | Core 01 §3.3.4.2 Record-history read contract | base | AC-231, AC-383 |
| REQ-01-562 | Core 01 §3.3.4.2 Record-history read contract | base | AC-231, AC-383 |
| REQ-01-563 | Core 01 §3.3.4.2 Record-history read contract | base | AC-231, AC-384 |
| REQ-01-564 | Core 01 §12.3 Incident portability | incident_portability | AC-236, AC-386 |
| REQ-01-565 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-231, AC-387 |
| REQ-01-566 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-231, AC-387 |
| REQ-01-567 | Core 01 §3.3.4.1 Filter predicate wire contract | base | AC-231, AC-387 |
| REQ-01-472 | Core 01 §17.2 Import Extension Profile public contract | import | AC-262..AC-264 |
| REQ-01-473 | Core 01 §17.2 Import Extension Profile public contract | import | AC-262, AC-323 |
| REQ-01-474 | Core 01 §17.2 Import Extension Profile public contract | import | AC-263..AC-264, AC-324..AC-325 |
| REQ-01-475 | Core 01 §17.2 Import Extension Profile public contract | import | AC-265, AC-323..AC-325 |
| REQ-01-618 | Core 01 §17.2 Import Extension Profile public contract | import | AC-463..AC-465 |
| REQ-01-619 | Core 01 §17.2 Import Extension Profile public contract | import | AC-465..AC-467 |
| REQ-01-620 | Core 01 §17.2 Import Extension Profile public contract | import | AC-464, AC-466 |
| REQ-01-621 | Core 01 §8.3 Projection maintenance | base | AC-470, AC-472 |
| REQ-01-622 | Core 01 §8.3 Projection maintenance | base | AC-470 |
| REQ-01-623 | Core 01 §3.3.4 Sheet row-query route contract | base | AC-471 |
| REQ-01-624 | Core 01 §12.2 Restore | base | AC-472 |
| REQ-01-625 | Core 01 §12.2 Restore | base | AC-472 |
| REQ-01-626 | Core 01 §8.6 Projection module boundary and import policy | base | AC-473 |
| REQ-01-476 | Core 01 §17.3 Snapshot and Reporting Extension Profile public contract | snapshot_reporting | AC-266..AC-268 |
| REQ-01-477 | Core 01 §17.3 Snapshot and Reporting Extension Profile public contract | snapshot_reporting | AC-266..AC-267 |
| REQ-01-478 | Core 01 §17.3 Snapshot and Reporting Extension Profile public contract | snapshot_reporting | AC-268, AC-305..AC-306 |
| REQ-01-479 | Core 01 §17.3 Snapshot and Reporting Extension Profile public contract | snapshot_reporting | AC-269, AC-307 |
| REQ-01-480 | Core 01 §17.4 Reference Pack Extension Profile public contract | reference_pack | AC-270..AC-271, AC-427 |
| REQ-01-481 | Core 01 §17.4 Reference Pack Extension Profile public contract | reference_pack | AC-270..AC-271, AC-308..AC-309, AC-326, AC-369, AC-427 |
| REQ-01-482 | Core 01 §17.4 Reference Pack Extension Profile public contract | reference_pack | AC-272, AC-310, AC-326 |
| REQ-01-483 | Core 01 §17.5 Incident Portability Extension Profile public contract | incident_portability | AC-273..AC-275 |
| REQ-01-484 | Core 01 §17.5 Incident Portability Extension Profile public contract | incident_portability | AC-273..AC-274 |
| REQ-01-485 | Core 01 §17.5 Incident Portability Extension Profile public contract | incident_portability | AC-275 |
| REQ-01-486 | Core 01 §17.5 Incident Portability Extension Profile public contract | incident_portability | AC-276, AC-327..AC-328, AC-332 |
| REQ-01-487 | Core 01 §18 Writable-string contract registry | base | AC-015, AC-068, AC-085..AC-086, AC-112, AC-118, AC-152, AC-175..AC-176, AC-181..AC-182, AC-186, AC-194, AC-196, AC-200, AC-202, AC-216, AC-221, AC-225, AC-231, AC-300..AC-303, AC-315..AC-319 |
| REQ-01-488 | Core 01 §18 Writable-string contract registry | base | AC-015, AC-068, AC-085..AC-086, AC-112, AC-118, AC-152, AC-175..AC-176, AC-181..AC-182, AC-184..AC-186, AC-194, AC-196, AC-200, AC-202, AC-216, AC-221, AC-225, AC-231, AC-300..AC-303, AC-315..AC-319 |
| REQ-01-489 | Core 01 §18 Writable-string contract registry | base | AC-118, AC-152, AC-175..AC-176, AC-231 |
| REQ-01-490 | Core 01 §18 Writable-string contract registry | base | AC-015, AC-068, AC-085..AC-086, AC-112, AC-118, AC-231 |
| REQ-01-491 | Core 01 §18 Writable-string contract registry | base | AC-068, AC-085..AC-086, AC-112, AC-118, AC-216, AC-231 |
| REQ-01-491.1 | Core 01 §18 Writable-string contract registry | base | AC-099, AC-170, AC-212, AC-214, AC-219..AC-220, AC-231 |
| REQ-01-492 | Core 01 §18 Writable-string contract registry | base | AC-015, AC-085, AC-100, AC-118, AC-231 |
| REQ-01-493 | Core 01 §18 Writable-string contract registry | base | AC-015, AC-085, AC-100, AC-118, AC-231 |
| REQ-01-494 | Core 01 §18 Writable-string contract registry | base | AC-118, AC-200, AC-231 |
| REQ-01-495 | Core 01 §18 Writable-string contract registry | base | AC-118, AC-202, AC-231 |
| REQ-01-496 | Core 01 §18 Writable-string contract registry | base | AC-085, AC-118, AC-181..AC-182, AC-186, AC-194, AC-196, AC-216, AC-221, AC-225, AC-231, AC-305, AC-308, AC-418 |
| REQ-01-497 | Core 01 §18 Writable-string contract registry | base | AC-175, AC-176, AC-178, AC-231, AC-247, AC-277, AC-279, AC-311, AC-312 |
| REQ-01-498 | Core 01 §18 Writable-string contract registry | base | AC-231, AC-277 |
| REQ-01-499 | Core 01 §19 Parties system-view addendum | base | AC-116..AC-118, AC-231, AC-277 |
| REQ-01-500 | Core 01 §19 Parties system-view addendum | base | AC-117..AC-118, AC-231, AC-277 |
| REQ-01-501 | Core 01 §19 Parties system-view addendum | base | AC-118, AC-231, AC-277 |
| REQ-01-502 | Core 01 §19 Parties system-view addendum | base | AC-118, AC-231, AC-278..AC-279, AC-318 |
| REQ-01-503 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-281, AC-300..AC-303 |
| REQ-01-504 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-282, AC-300..AC-303 |
| REQ-01-505 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-283, AC-300..AC-303 |
| REQ-01-506 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-284, AC-300, AC-302..AC-303 |
| REQ-01-507 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-285 |
| REQ-01-508 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-286 |
| REQ-01-509 | Core 01 §19 Additional coordination and optional artifact-backed surface addenda | base | AC-116..AC-118, AC-231, AC-287 |
| REQ-01-510 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-235, AC-288, AC-290..AC-291, AC-436 |
| REQ-01-511 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-235, AC-288..AC-289, AC-435..AC-436 |
| REQ-01-512 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-235, AC-290..AC-291 |
| REQ-01-513 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-235, AC-292..AC-293 |
| REQ-01-514 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-235, AC-293 |
| REQ-01-515 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-235, AC-289..AC-291 |
| REQ-01-537 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-348, AC-352 |
| REQ-01-538 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-348..AC-349 |
| REQ-01-539 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-350, AC-352 |
| REQ-01-540 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-351..AC-352 |
| REQ-01-541 | Core 01 §20 Enterprise Authentication Extension Profile public contract | enterprise_authentication | AC-349..AC-352 |
| REQ-01-516 | Core 01 §18B Writable direct-reference-scalar contract registry | base | AC-315, AC-317, AC-319 |
| REQ-01-517 | Core 01 §18B Writable direct-reference-scalar contract registry | base | AC-315..AC-317, AC-319 |
| REQ-01-518 | Core 01 §18B Writable direct-reference-scalar contract registry | base | AC-317, AC-319 |
| REQ-01-519 | Core 01 §18B Writable direct-reference-scalar contract registry | base | AC-315..AC-318 |
| REQ-01-520 | Core 01 §18B Writable direct-reference-scalar contract registry | base | AC-315..AC-317, AC-319 |
| REQ-01-521 | Core 01 §18 Writable-string contract registry | base | AC-175..AC-176, AC-231, AC-244..AC-245 |
| REQ-01-568 | Core 01 §18 Writable-string contract registry | base | AC-118, AC-231, AC-388..AC-391 |
| REQ-01-569 | Core 01 §7.4 Authoritative base-profile view schema registry | base | AC-396 |
| REQ-01-570 | Core 01 §3.3.3 Route families | base | AC-402 |
| REQ-01-571 | Core 01 §12.1 Backup | base | AC-398..AC-399, AC-440 |
| REQ-01-572 | Core 01 §12.1 Backup | base | AC-398, AC-401 |
| REQ-01-596 | Core 01 §12.1 Backup | base | AC-398, AC-428 |
| REQ-01-573 | Core 01 §12.1 Backup | base | AC-398 |
| REQ-01-574 | Core 01 §12.1 Backup | base | AC-398..AC-399, AC-401 |
| REQ-01-575 | Core 01 §12.2 Restore | base | AC-399..AC-400 |
| REQ-01-576 | Core 01 §12.2 Restore | base | AC-400 |
| REQ-01-577 | Core 01 §12.2 Restore | base | AC-399 |
| REQ-01-578 | Core 01 §12.2 Restore | base | AC-401 |
| REQ-01-593 | Core 01 §12.2.1 Operator recovery CLI contract | base | AC-428 |
| REQ-01-594 | Core 01 §12.2.1 Operator recovery CLI contract | base | AC-428 |
| REQ-01-595 | Core 01 §12.2.1 Operator recovery CLI contract | base | AC-428 |
| REQ-01-522 | Core 01 §3.3.2.2 Credential lifecycle and TOTP bootstrap routes | base | AC-334..AC-339 |
| REQ-01-523 | Core 01 §3.3.2.2 Credential lifecycle and TOTP bootstrap routes | base | AC-335, AC-339 |
| REQ-01-524 | Core 01 §3.3.2.2 Credential lifecycle and TOTP bootstrap routes | base | AC-338..AC-339 |
| REQ-01-525 | Core 01 §3.3.2.2 Credential lifecycle and TOTP bootstrap routes | base | AC-336, AC-339 |
| REQ-01-526 | Core 01 §3.3.2.2 Credential lifecycle and TOTP bootstrap routes | base | AC-337, AC-339 |
| REQ-01-597 | Core 01 §3.3.2.3 Current-account profile and preference routes | base | AC-429..AC-432 |
| REQ-01-598 | Core 01 §3.3.2.3 Current-account profile and preference routes | base | AC-429, AC-432 |
| REQ-01-599 | Core 01 §3.3.2.3 Current-account profile and preference routes | base | AC-429..AC-430, AC-432 |
| REQ-01-600 | Core 01 §3.3.2.3 Current-account profile and preference routes | base | AC-431..AC-432 |
| REQ-01-601 | Core 01 §3.3.2.3 Current-account profile and preference routes | base | AC-431..AC-432 |
| REQ-01-602 | Core 01 §3.3.2.3 Current-account profile and preference routes | base | AC-432 |
| REQ-01-527 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-340..AC-341 |
| REQ-01-528 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-341 |
| REQ-01-529 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-342 |
| REQ-01-530 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-343..AC-346 |
| REQ-01-531 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-343..AC-344 |
| REQ-01-532 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-343..AC-344 |
| REQ-01-533 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-343..AC-346 |
| REQ-01-534 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-343..AC-344 |
| REQ-01-535 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-344..AC-346 |
| REQ-01-536 | Core 01 §3.3.5.1 Deployment-local user-account and incident-membership administration contracts | base | AC-343, AC-347 |
| REQ-01-603 | Core 01 §3.3.5.1A Administrative audit read projections | base | AC-437..AC-439 |
| REQ-01-604 | Core 01 §3.3.5.1A Administrative audit read projections | base | AC-437, AC-440 |
| REQ-01-605 | Core 01 §3.3.5.1A Administrative audit read projections | base | AC-437 |
| REQ-01-606 | Core 01 §3.3.5.1A Administrative audit read projections | base, enterprise_authentication | AC-437, AC-440 |
| REQ-01-607 | Core 01 §3.3.5.1A Administrative audit read projections | base | AC-438 |
| REQ-01-608 | Core 01 §3.3.2.1B Deployment administration browser context | base | AC-414, AC-427, AC-441 |
| REQ-01-609 | Core 01 §12.3.6 Export and import execution semantics | incident_portability | AC-442 |
| REQ-01-610 | Core 01 §17.4 Reference Pack Extension Profile public contract | reference_pack | AC-443 |
| REQ-01-611 | Core 01 §3.3.5.0A Timeline time-conversion profile | base | AC-444, AC-449, AC-451 |
| REQ-01-612 | Core 01 §3.3.5.0A Timeline time-conversion profile | base | AC-444, AC-449, AC-451 |
| REQ-01-613 | Core 01 §3.3.5.0A Timeline time-conversion profile | base | AC-451 |
| REQ-01-614 | Core 01 §18 timeline_visible_text_v1 | base | AC-444..AC-449, AC-452 |
| REQ-01-615 | Core 01 §7.4 `inspector_config_v1` | base | AC-453..AC-455, AC-460..AC-462 |
| REQ-01-616 | Core 01 §7.4 per-surface inspector matrix | base | AC-453..AC-455, AC-461..AC-462 |
| REQ-01-617 | Core 01 §7.4.1A Inspector feature-group registry | base | AC-454..AC-455, AC-458, AC-460..AC-462 |
| REQ-02-001 | Core 02 §1 Domain-model goals | base, reference_pack | AC-231, AC-234 |
| REQ-02-002 | Core 02 §1 Domain-model goals | base | AC-231 |
| REQ-02-003 | Core 02 §2 Core record types | base, reference_pack | AC-231, AC-234, AC-277 |
| REQ-02-004 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-005 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-006 | Core 02 §3 Record envelope contract | base | AC-231, AC-277 |
| REQ-02-007 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-008 | Core 02 §3 Record envelope contract | base | AC-231 |
| REQ-02-009 | Core 02 §4.1 Normalized data | base, snapshot_reporting, reference_pack | AC-097..AC-101, AC-231, AC-233..AC-234, AC-277 |
| REQ-02-010 | Core 02 §4.3 JSONB discipline | base | AC-097..AC-101, AC-231 |
| REQ-02-011 | Core 02 §4.3 JSONB discipline | base | AC-097..AC-101, AC-231 |
| REQ-02-012 | Core 02 §4.4 Promotion rule for recurrent incident-specific fields | base | AC-097..AC-101, AC-231 |
| REQ-02-013 | Core 02 §4.4 Promotion rule for recurrent incident-specific fields | base | AC-097..AC-101, AC-231 |
| REQ-02-014 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-015 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-016 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-017 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231, AC-278, AC-318 |
| REQ-02-018 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-019 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-020 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-021 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231, AC-278, AC-280, AC-318 |
| REQ-02-022 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231, AC-277..AC-280, AC-318 |
| REQ-02-023 | Core 02 §4.5 Current-profile promoted field sets | base | AC-097..AC-101, AC-211..AC-214, AC-231 |
| REQ-02-024 | Core 02 §5 Partial and uncertain data | base | AC-231, AC-406 |
| REQ-02-025 | Core 02 §5 Partial and uncertain data | base | AC-231, AC-406 |
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
| REQ-02-259 | Core 02 §7.2 File-based import provenance | import | AC-463, AC-465, AC-467 |
| REQ-02-054 | Core 02 §7.3 Entity provenance | base | AC-186, AC-209, AC-231 |
| REQ-02-055 | Core 02 §7.3 Entity provenance | base | AC-186, AC-209, AC-231 |
| REQ-02-056 | Core 02 §7.4 Indicator observation provenance | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-057 | Core 02 §7.4 Indicator observation provenance | base | AC-017, AC-072..AC-079, AC-122, AC-231 |
| REQ-02-058 | Core 02 §8.1 Mention deduplication | base | AC-021, AC-028, AC-188, AC-231 |
| REQ-02-059 | Core 02 §8.1 Mention deduplication | base | AC-021, AC-028, AC-188, AC-231 |
| REQ-02-060 | Core 02 §8.2 Entity exact-match precedence | base | AC-022, AC-028, AC-186..AC-187, AC-231, AC-279 |
| REQ-02-061 | Core 02 §8.2 Entity exact-match precedence | base | AC-022, AC-028, AC-186..AC-187, AC-231, AC-279 |
| REQ-02-062 | Core 02 §8.3 Suggestion boundary | base | AC-028..AC-029, AC-231, AC-279 |
| REQ-02-063 | Core 02 §8.3 Suggestion boundary | base | AC-028..AC-029, AC-231, AC-279..AC-280 |
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
| REQ-02-102 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-103 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-104 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-105 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-106 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-107 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-108 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-109 | Core 02 §10.4.1.1 Task-request lifecycle machine | base | AC-085, AC-137..AC-140, AC-145, AC-231 |
| REQ-02-110 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-111 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-112 | Core 02 §10.4.2 `decision` record type | base, snapshot_reporting | AC-086, AC-141..AC-145, AC-231, AC-233 |
| REQ-02-113 | Core 02 §10.4.2 `decision` record type | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-114 | Core 02 §10.4.2.1 Decision lifecycle machine | base | AC-086, AC-141..AC-145, AC-231, AC-314 |
| REQ-02-115 | Core 02 §10.4.2.1 Decision lifecycle machine | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-116 | Core 02 §10.4.2.1 Decision lifecycle machine | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-117 | Core 02 §10.4.2.1 Decision lifecycle machine | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-118 | Core 02 §10.4.2.1 Decision lifecycle machine | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-119 | Core 02 §10.4.2.1 Decision lifecycle machine | base | AC-086, AC-141..AC-145, AC-231 |
| REQ-02-120 | Core 02 §10.4.3 Ownership and hot-path boundary | base | AC-085..AC-086, AC-090, AC-231, AC-278 |
| REQ-02-121 | Core 02 §10.4.3 Ownership and hot-path boundary | base | AC-085..AC-086, AC-090, AC-231 |
| REQ-02-122 | Core 02 §10.4.3 Ownership and hot-path boundary | base | AC-085..AC-086, AC-090, AC-231 |
| REQ-02-123 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-281..AC-284 |
| REQ-02-124 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-278, AC-281 |
| REQ-02-125 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-278, AC-281 |
| REQ-02-126 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-282 |
| REQ-02-127 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-282 |
| REQ-02-128 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-283 |
| REQ-02-129 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-283 |
| REQ-02-130 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-284 |
| REQ-02-131 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-284 |
| REQ-02-132 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-281..AC-284 |
| REQ-02-133 | Core 02 §10.4.4 Structured coordination artifact types | base | AC-087..AC-089, AC-231, AC-281..AC-284 |
| REQ-02-134 | Core 02 §10.4.5 Hypothesis boundary | base | AC-089, AC-231, AC-410 |
| REQ-02-135 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231, AC-285..AC-287 |
| REQ-02-136 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231, AC-285 |
| REQ-02-137 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231, AC-286 |
| REQ-02-138 | Core 02 §10.4.6 Optional structured findings, investigative-query, and forensic-keyword surfaces | base | AC-101, AC-231, AC-287 |
| REQ-02-139 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233, AC-333 |
| REQ-02-140 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-141 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-142 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233 |
| REQ-02-143 | Core 02 §10.5 Snapshot and reporting extension objects | snapshot_reporting | AC-030..AC-032, AC-056..AC-062, AC-071, AC-091, AC-104..AC-106, AC-113..AC-115, AC-233, AC-333 |
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
| REQ-02-168 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231, AC-329 |
| REQ-02-169 | Core 02 §12.1 `record_link` object contract | base | AC-196, AC-205..AC-210, AC-231, AC-332 |
| REQ-02-170 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-171 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-172 | Core 02 §12.1 `record_link` object contract | base | AC-205..AC-210, AC-231 |
| REQ-02-173 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-174 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-175 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-196, AC-205..AC-210, AC-231, AC-332 |
| REQ-02-176 | Core 02 §12.2 Relationship vocabulary and canonical direction | base | AC-205..AC-210, AC-231 |
| REQ-02-177 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-178 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-179 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-180 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-181 | Core 02 §12.3 Link metadata semantics | base | AC-196, AC-205..AC-210, AC-231, AC-331..AC-332 |
| REQ-02-182 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-183 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-184 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-185 | Core 02 §12.3 Link metadata semantics | base | AC-205..AC-210, AC-231 |
| REQ-02-186 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-187 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-188 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-189 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-190 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-191 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-192 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-193 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-194 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-195 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-196 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-02-197 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-278, AC-280 |
| REQ-02-198 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-280 |
| REQ-02-199 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-278 |
| REQ-02-200 | Core 02 §13 Evidence and object metadata | base, snapshot_reporting | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231, AC-233 |
| REQ-02-201 | Core 02 §13 Evidence and object metadata | base | AC-015..AC-016, AC-053, AC-100, AC-102..AC-103, AC-128, AC-154..AC-155, AC-231 |
| REQ-02-202 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base, import, snapshot_reporting, reference_pack | AC-017..AC-018, AC-072..AC-075, AC-118, AC-128, AC-154..AC-155, AC-188..AC-190, AC-200..AC-204, AC-231..AC-234, AC-277..AC-278, AC-280, AC-440 |
| REQ-02-243 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-170, AC-174, AC-231 |
| REQ-02-244 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-175..AC-176, AC-178, AC-231 |
| REQ-02-245 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-175, AC-177, AC-335..AC-342, AC-231 |
| REQ-02-255 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-429, AC-431, AC-432 |
| REQ-02-246 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-231, AC-343..AC-344, AC-346 |
| REQ-02-247 | Core 02 §14.1 Persistence realization status and deployment-local invariants | reference_pack | AC-093..AC-095, AC-234 |
| REQ-02-256 | Core 02 §14.1 Persistence realization status and deployment-local invariants | enterprise_authentication | AC-435..AC-436 |
| REQ-02-257 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-437, AC-439, AC-440 |
| REQ-02-248 | Core 02 §12.3 Link metadata semantics | base | AC-394..AC-397 |
| REQ-02-249 | Core 02 §14.1 Persistence realization status and deployment-local invariants | incident_portability | AC-236, AC-409, AC-432, AC-440 |
| REQ-02-250 | Core 02 §10.4.4A Tagged-variant registry for artifact-backed notes, coordination artifacts, and structured findings | base | AC-410 |
| REQ-02-251 | Core 02 §10.4.4A Tagged-variant registry for artifact-backed notes, coordination artifacts, and structured findings | base | AC-410 |
| REQ-02-252 | Core 02 §10.4.4A Tagged-variant registry for artifact-backed notes, coordination artifacts, and structured findings | base | AC-410 |
| REQ-02-253 | Core 02 §10.4.4A Tagged-variant registry for artifact-backed notes, coordination artifacts, and structured findings | base | AC-410 |
| REQ-02-203 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base | AC-170, AC-231 |
| REQ-02-204 | Core 02 §14.1 Persistence realization status and deployment-local invariants | base, incident_portability | AC-231, AC-236, AC-409, AC-432, AC-440 |
| REQ-02-205 | Core 02 §14.2 Rollback granularity substrate | base | AC-215..AC-218, AC-231 |
| REQ-02-206 | Core 02 §14.2 Rollback granularity substrate | base, snapshot_reporting | AC-215..AC-218, AC-231, AC-233 |
| REQ-02-207 | Core 02 §14.2 Rollback granularity substrate | base, snapshot_reporting | AC-215..AC-218, AC-231, AC-233 |
| REQ-02-208 | Core 02 §14.3 Mutation targets | base | AC-125, AC-200..AC-208, AC-231 |
| REQ-02-209 | Core 02 §14.3 Mutation targets | base | AC-125, AC-200..AC-208, AC-231 |
| REQ-02-210 | Core 02 §14.4 Soft delete | base | AC-181..AC-183, AC-231 |
| REQ-02-211 | Core 02 §14.5 Snapshot and reporting extension fields | snapshot_reporting | AC-057, AC-060..AC-062, AC-113..AC-115, AC-233, AC-333 |
| REQ-02-212 | Core 02 §15.1 Attribution unit | base | AC-215..AC-217, AC-231, AC-412 |
| REQ-02-213 | Core 02 §15.1 Attribution unit | base | AC-215..AC-217, AC-231, AC-412 |
| REQ-02-214 | Core 02 §15.2 Mutation-entry unit | base | AC-215..AC-217, AC-231, AC-412 |
| REQ-02-215 | Core 02 §15.2 Mutation-entry unit | base | AC-215..AC-217, AC-231, AC-412 |
| REQ-02-216 | Core 02 §15.2 Mutation-entry unit | base | AC-215..AC-217, AC-231, AC-384, AC-412 |
| REQ-02-217 | Core 02 §15.3 Reconstruction requirement | base, snapshot_reporting | AC-215, AC-217, AC-231, AC-233, AC-412 |
| REQ-02-218 | Core 02 §15.3 Reconstruction requirement | base | AC-215, AC-217, AC-231, AC-412 |
| REQ-02-219 | Core 02 §15.4 Entity-merge history | base | AC-023, AC-186, AC-209, AC-231, AC-412 |
| REQ-02-220 | Core 02 §15.4 Entity-merge history | base | AC-023, AC-186, AC-209, AC-231, AC-412 |
| REQ-02-221 | Core 02 §17 Domain invariants | base, reference_pack | AC-231, AC-234 |
| REQ-02-222 | Core 02 §18 Canonical closed-vocabulary registry | base, incident_portability | AC-076..AC-084, AC-099, AC-121..AC-122, AC-137..AC-145, AC-170, AC-212, AC-214, AC-219..AC-220, AC-231, AC-236, AC-252..AC-253, AC-277, AC-284, AC-285, AC-287, AC-416, AC-425 |
| REQ-02-223 | Core 02 §18 Canonical closed-vocabulary registry | base | AC-076..AC-084, AC-099, AC-121..AC-122, AC-137..AC-145, AC-170, AC-212, AC-214, AC-219..AC-220, AC-231, AC-252..AC-253, AC-277, AC-416, AC-425 |
| REQ-02-254 | Core 02 §18 Canonical closed-vocabulary registry | base | AC-425 |
| REQ-02-224 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-277 |
| REQ-02-225 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-277 |
| REQ-02-226 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-278, AC-280 |
| REQ-02-227 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-278 |
| REQ-02-228 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-278 |
| REQ-02-229 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-279 |
| REQ-02-230 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-279 |
| REQ-02-231 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-280, AC-318 |
| REQ-02-232 | Core 02 §19 Incident-scoped party model | base | AC-231, AC-279..AC-280 |
| REQ-02-233 | Core 02 §10.4.1 `task_request` record type | base | AC-319 |
| REQ-02-234 | Core 02 §14.1 Persistence realization status and deployment-local invariants | enterprise_authentication | AC-348..AC-351 |
| REQ-02-235 | Core 02 §14.1 Persistence realization status and deployment-local invariants | enterprise_authentication | AC-349..AC-351 |
| REQ-02-236 | Core 02 §14.1 Persistence realization status and deployment-local invariants | enterprise_authentication | AC-348, AC-351..AC-352 |
| REQ-02-237 | Core 02 §14.1 Persistence realization status and deployment-local invariants | enterprise_authentication | AC-175, AC-348, AC-351 |
| REQ-02-238 | Core 02 §15.3.1 Retained history and rollback horizon | base | AC-231, AC-383, AC-386 |
| REQ-02-239 | Core 02 §15.3.1 Retained history and rollback horizon | base | AC-231, AC-383, AC-385 |
| REQ-02-240 | Core 02 §15.3.1 Retained history and rollback horizon | base | AC-231, AC-383 |
| REQ-02-241 | Core 02 §15.3.1 Retained history and rollback horizon | base | AC-231, AC-384, AC-386 |
| REQ-02-242 | Core 02 §15.3.1 Retained history and rollback horizon | base | AC-231 |
| REQ-02-258 | Core 02 §2 inspector source-state boundary | base | AC-453 |
| REQ-03-001 | Core 03 §1 Interaction model | base | AC-001..AC-002, AC-005, AC-043, AC-231 |
| REQ-03-002 | Core 03 §1 Interaction model | base | AC-001..AC-002, AC-005, AC-043, AC-231 |
| REQ-03-003 | Core 03 §1 Interaction model | base | AC-001..AC-002, AC-005, AC-043, AC-231 |
| REQ-03-004 | Core 03 §2.1 Built-in tabs | base | AC-112, AC-116, AC-231, AC-410 |
| REQ-03-005 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231, AC-277 |
| REQ-03-006 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-007 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-008 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-009 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231 |
| REQ-03-010 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231, AC-281..AC-284 |
| REQ-03-011 | Core 03 §2.2 System views | base | AC-078, AC-085..AC-090, AC-121..AC-122, AC-231, AC-281..AC-284, AC-410, AC-411 |
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
| REQ-03-030 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231, AC-290..AC-291, AC-414 |
| REQ-03-031 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231, AC-290..AC-291, AC-414 |
| REQ-03-032 | Core 03 §2.4 Startup and default surface selection | base | AC-150, AC-153, AC-231 |
| REQ-03-289 | Core 03 §2.4 Startup and default surface selection | base | AC-431, AC-432 |
| REQ-03-290 | Core 03 §2.4 Startup and default surface selection | base, incident_portability | AC-441, AC-442 |
| REQ-03-291 | Core 03 §2.3 Saved views and inspector behavior | base | AC-453, AC-456..AC-457, AC-462 |
| REQ-03-292 | Core 03 §2.3A Inspector workflow interaction semantics | base | AC-453, AC-456..AC-458, AC-462 |
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
| REQ-03-072 | Core 03 §3.3.4 Same-field conflict transport contract | base | AC-126, AC-203..AC-204, AC-226..AC-231, AC-381 |
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
| REQ-03-089 | Core 03 §4.2 Save-state presentation | base | AC-043, AC-231, AC-376 |
| REQ-03-090 | Core 03 §4.3 Presence | base | AC-008, AC-132, AC-231 |
| REQ-03-091 | Core 03 §4.3 Presence | base | AC-008, AC-132, AC-231 |
| REQ-03-092 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-093 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-094 | Core 03 §4.3.1 Collaboration message application | base, snapshot_reporting | AC-129, AC-131..AC-136, AC-231, AC-233 |
| REQ-03-095 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231, AC-376 |
| REQ-03-096 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-097 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231, AC-368 |
| REQ-03-098 | Core 03 §4.3.1 Collaboration message application | base | AC-129, AC-131..AC-136, AC-231 |
| REQ-03-287 | Core 03 §4.3.2 Closed incident workbook mode | base, snapshot_reporting, incident_portability | AC-424 |
| REQ-03-288 | Core 03 §4.3.2 Closed incident workbook mode | base | AC-424, AC-426 |
| REQ-03-099 | Core 03 §4.4 Local pending queue | base | AC-156..AC-163, AC-231, AC-376..AC-382 |
| REQ-03-100 | Core 03 §4.4 Local pending queue | base | AC-156..AC-163, AC-231, AC-376..AC-382 |
| REQ-03-101 | Core 03 §5 Locking policy | base | AC-182, AC-187, AC-218, AC-231, AC-353 |
| REQ-03-102 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-103 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-104 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-105 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231 |
| REQ-03-106 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231, AC-329, AC-331 |
| REQ-03-107 | Core 03 §6 Record lifecycle | base | AC-107..AC-111, AC-137..AC-145, AC-191..AC-199, AC-231, AC-331 |
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
| REQ-03-121 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-03-122 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-03-123 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-03-124 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-03-125 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-03-126 | Core 03 §8.3 Blob and evidence lifecycle bridge | base | AC-015..AC-016, AC-103, AC-107..AC-111, AC-128, AC-154..AC-155, AC-231, AC-313 |
| REQ-03-127 | Core 03 §8.4 Evidence access | base | AC-053..AC-054, AC-103, AC-128, AC-231, AC-252, AC-255 |
| REQ-03-128 | Core 03 §8.4 Evidence access | base | AC-053..AC-054, AC-103, AC-128, AC-231, AC-252..AC-254 |
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
| REQ-03-293 | Core 03 §11.2.4 Discovery and batch semantics | import | AC-463..AC-464, AC-466..AC-467 |
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
| REQ-03-276 | Core 03 §12.1 Allowed scope | base | AC-205, AC-231, AC-388, AC-392, AC-393 |
| REQ-03-277 | Core 03 §12.2 Required confidence conditions | base | AC-205, AC-231, AC-388..AC-391 |
| REQ-03-278 | Core 03 §12.2 Required confidence conditions | base | AC-231, AC-389..AC-390 |
| REQ-03-279 | Core 03 §12.2 Required confidence conditions | base | AC-231, AC-391 |
| REQ-03-280 | Core 03 §13.1 Grid editing | base | AC-231, AC-394..AC-396 |
| REQ-03-281 | Core 03 §13.1 Grid editing | base | AC-231, AC-354 |
| REQ-03-205 | Core 03 §12.2 Required confidence conditions | base | AC-205, AC-231, AC-389..AC-391 |
| REQ-03-206 | Core 03 §12.3 Required write effects | base | AC-205, AC-231 |
| REQ-03-207 | Core 03 §12.3 Required write effects | base | AC-205, AC-231 |
| REQ-03-208 | Core 03 §12.4 Allowed and forbidden workflows | base, import | AC-205, AC-231..AC-232, AC-392..AC-393 |
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
| REQ-03-223 | Core 03 §14.1 Sort and filter behavior | base | AC-013..AC-014, AC-044, AC-047, AC-124, AC-184..AC-185, AC-231, AC-387 |
| REQ-03-224 | Core 03 §14.1 Sort and filter behavior | base | AC-013..AC-014, AC-044, AC-047, AC-124, AC-184..AC-185, AC-231, AC-360, AC-387 |
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
| REQ-03-247 | Core 03 §16.2 Inspector | base | AC-006, AC-020, AC-023, AC-072..AC-075, AC-186..AC-187, AC-209..AC-210, AC-231, AC-278..AC-279, AC-366 |
| REQ-03-248 | Core 03 §16.2 Inspector | base | AC-006, AC-020, AC-023, AC-072..AC-075, AC-186..AC-187, AC-209..AC-210, AC-231 |
| REQ-03-249 | Core 03 §16.2 Inspector | base | AC-006, AC-020, AC-023, AC-072..AC-075, AC-186..AC-187, AC-209..AC-210, AC-231 |
| REQ-03-250 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-251 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-252 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-253 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-254 | Core 03 §16.3 Compromise-assessment surfaces | base | AC-018, AC-080..AC-084, AC-121, AC-231 |
| REQ-03-255 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-256 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231, AC-278..AC-279 |
| REQ-03-257 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-258 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-259 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231, AC-279, AC-281..AC-287 |
| REQ-03-260 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-085..AC-090, AC-137..AC-145, AC-231 |
| REQ-03-261 | Core 03 §17 Authorship and attribution in the UI | base | AC-007, AC-231 |
| REQ-03-262 | Core 03 §17 Authorship and attribution in the UI | base | AC-007, AC-231 |
| REQ-03-263 | Core 03 §18.1 Excel-like behaviors | base | AC-005, AC-043, AC-231 |
| REQ-03-264 | Core 03 §18.2 Intentional differences | base | AC-090, AC-231 |
| REQ-03-265 | Core 03 §19 Interaction invariants | base | AC-231, AC-281..AC-287 |
| REQ-03-266 | Core 03 §20 Parties system-view and linking flows | base | AC-231, AC-277 |
| REQ-03-267 | Core 03 §20 Parties system-view and linking flows | base | AC-231, AC-279 |
| REQ-03-268 | Core 03 §20 Parties system-view and linking flows | base | AC-231, AC-278..AC-279 |
| REQ-03-269 | Core 03 §20 Parties system-view and linking flows | base | AC-231, AC-278 |
| REQ-03-270 | Core 03 §20 Parties system-view and linking flows | base | AC-231, AC-278..AC-279 |
| REQ-03-271 | Core 03 §20 Parties system-view and linking flows | base | AC-231, AC-278..AC-279 |
| REQ-03-272 | Core 03 §16.2 Inspector | base | AC-315, AC-318 |
| REQ-03-273 | Core 03 §16.4 Analyst-work coordination surfaces | base | AC-315, AC-319 |
| REQ-03-274 | Core 03 §20 Parties system-view and linking flows | base | AC-315, AC-318 |
| REQ-03-275 | Core 03 §4.3.1 Collaboration message application | base | AC-231 |
| REQ-03-282 | Core 03 §4.2 Save-state presentation | base | AC-043, AC-125, AC-126, AC-181, AC-183, AC-200, AC-231 |
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
| REQ-04-016 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231, AC-430..AC-431 |
| REQ-04-095 | Core 04 §1.1.1 Session lifecycle boundaries | enterprise_authentication | AC-350..AC-352 |
| REQ-04-017 | Core 04 §1.1.1 Session lifecycle boundaries | base | AC-123, AC-131, AC-136, AC-156..AC-163, AC-231 |
| REQ-04-018 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-036, AC-235, AC-288..AC-291, AC-293 |
| REQ-04-019 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-036, AC-235, AC-292..AC-293 |
| REQ-04-020 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-036, AC-235, AC-290..AC-291, AC-293, AC-433..AC-434, AC-436 |
| REQ-04-093 | Core 04 §1.2 Enterprise Authentication Extension Profile | enterprise_authentication | AC-348, AC-352, AC-435..AC-436 |
| REQ-04-021 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-022 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231, AC-280 |
| REQ-04-023 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231, AC-254..AC-255, AC-257, AC-260..AC-261, AC-427 |
| REQ-04-024 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231, AC-280 |
| REQ-04-025 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-026 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231 |
| REQ-04-027 | Core 04 §2 Authorization model | base, snapshot_reporting | AC-054, AC-149, AC-178..AC-180, AC-231, AC-233 |
| REQ-04-028 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231, AC-343..AC-346, AC-414, AC-427, AC-432, AC-439, AC-441 |
| REQ-04-029 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231, AC-261, AC-414, AC-427, AC-439, AC-441 |
| REQ-04-030 | Core 04 §2 Authorization model | base | AC-054, AC-149, AC-178..AC-180, AC-231, AC-427, AC-439 |
| REQ-04-114 | Core 04 §2 Authorization model | base | AC-429..AC-432 |
| REQ-04-123 | Core 04 §2 Authorization model | base | AC-438..AC-439 |
| REQ-04-126 | Core 04 §2 Authorization model | base | AC-441 |
| REQ-04-031 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-032 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-033 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-034 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233, AC-306 |
| REQ-04-035 | Core 04 §2.1 Snapshot and Reporting Extension Profile release gate | snapshot_reporting | AC-059..AC-060, AC-104..AC-106, AC-233 |
| REQ-04-036 | Core 04 §3 Attribution and audit requirements | base | AC-231, AC-407 |
| REQ-04-037 | Core 04 §3 Attribution and audit requirements | base | AC-231, AC-408 |
| REQ-04-038 | Core 04 §3 Attribution and audit requirements | base, incident_portability | AC-175, AC-231, AC-236, AC-343..AC-344, AC-346, AC-409, AC-430, AC-432, AC-440 |
| REQ-04-039 | Core 04 §3 Attribution and audit requirements | base | AC-231 |
| REQ-04-040 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-041 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-042 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-043 | Core 04 §4.1 Reference packs | reference_pack | AC-033, AC-035, AC-052, AC-092..AC-096, AC-234 |
| REQ-04-044 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-045 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236, AC-333 |
| REQ-04-046 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-047 | Core 04 §4.2 Export outputs | snapshot_reporting, incident_portability | AC-031, AC-057, AC-059..AC-062, AC-091, AC-113..AC-115, AC-164..AC-169, AC-233, AC-236 |
| REQ-04-048 | Core 04 §4.3 Evidence uploads | base | AC-053, AC-128, AC-231 |
| REQ-04-049 | Core 04 §4.4 STRIDE threat model | base | AC-048, AC-231 |
| REQ-04-050 | Core 04 §4.4 STRIDE threat model | base | AC-048, AC-231 |
| REQ-04-051 | Core 04 §4.4 STRIDE threat model | base | AC-048, AC-231 |
| REQ-04-052 | Core 04 §4.5 Focused MITRE CWE constraints | base | AC-049..AC-055, AC-130..AC-131, AC-231 |
| REQ-04-053 | Core 04 §4.5 Focused MITRE CWE constraints | base, import, reference_pack | AC-049..AC-055, AC-130..AC-131, AC-231..AC-232, AC-234, AC-252..AC-255 |
| REQ-04-054 | Core 04 §5.1 Flyaway or disconnected deployment | base | AC-055, AC-092, AC-096, AC-169, AC-231 |
| REQ-04-055 | Core 04 §5.1 Flyaway or disconnected deployment | base, reference_pack | AC-055, AC-092, AC-096, AC-169, AC-231, AC-234 |
| REQ-04-056 | Core 04 §5.2 On-prem deployment | base | AC-231 |
| REQ-04-057 | Core 04 §5.3 Cloud deployment | base | AC-231 |
| REQ-04-058 | Core 04 §6 Runtime roots and storage paths | base, reference_pack | AC-051, AC-055, AC-169, AC-231, AC-234, AC-294..AC-295, AC-297 |
| REQ-04-059 | Core 04 §6 Runtime roots and storage paths | base | AC-051, AC-055, AC-169, AC-231, AC-296 |
| REQ-04-060 | Core 04 §7 Container boundary | base | AC-231 |
| REQ-04-061 | Core 04 §8.1 Required services | base | AC-231 |
| REQ-04-065 | Core 04 §11 Operational posture | base | AC-164..AC-169, AC-231 |
| REQ-04-066 | Core 04 §12.1 Scope and owner | base | AC-294, AC-298, AC-320 |
| REQ-04-067 | Core 04 §12.2 Canonical artifact and discovery | base | AC-294, AC-297 |
| REQ-04-068 | Core 04 §12.2 Canonical artifact and discovery | base | AC-294 |
| REQ-04-069 | Core 04 §12.2 Canonical artifact and discovery | base | AC-294..AC-295, AC-297 |
| REQ-04-070 | Core 04 §12.2 Canonical artifact and discovery | base | AC-294 |
| REQ-04-071 | Core 04 §12.3 Key registry and binding model | base, reference_pack | AC-294..AC-295, AC-297 |
| REQ-04-072 | Core 04 §12.3 Key registry and binding model | base | AC-295 |
| REQ-04-073 | Core 04 §12.3 Key registry and binding model | base, reference_pack | AC-295, AC-297 |
| REQ-04-074 | Core 04 §12.4 Filesystem-root path contract | base, reference_pack | AC-296..AC-297 |
| REQ-04-075 | Core 04 §12.4 Filesystem-root path contract | base, reference_pack | AC-296 |
| REQ-04-076 | Core 04 §12.5 Canonical disconnected-layout defaults | base, reference_pack | AC-297 |
| REQ-04-077 | Core 04 §12.6 Validation error contract and startup behavior | base | AC-294..AC-296, AC-298, AC-320, AC-433..AC-435 |
| REQ-04-078 | Core 04 §12.6 Validation error contract and startup behavior | base | AC-298, AC-433..AC-435 |
| REQ-04-079 | Core 04 §12.3.1 Resource-limit registry | base | AC-320..AC-328 |
| REQ-04-080 | Core 04 §12.3.1 Resource-limit registry | base | AC-320, AC-322..AC-328 |
| REQ-04-081 | Core 04 §12.3.1 Resource-limit registry | base | AC-320 |
| REQ-04-083 | Core 04 §1.1 Base authentication | base | AC-334..AC-342 |
| REQ-04-084 | Core 04 §1.1 Base authentication | base | AC-334..AC-339, AC-347 |
| REQ-04-085 | Core 04 §2 Authorization model | base | AC-340..AC-342, AC-427 |
| REQ-04-094 | Core 04 §2 Authorization model | enterprise_authentication | AC-352, AC-427 |
| REQ-04-086 | Core 04 §3 Attribution and audit requirements | base | AC-336..AC-338, AC-340..AC-342, AC-440 |
| REQ-04-096 | Core 04 §3 Attribution and audit requirements | enterprise_authentication | AC-352, AC-440 |
| REQ-04-124 | Core 04 §3 Attribution and audit requirements | base, enterprise_authentication | AC-437, AC-440 |
| REQ-04-125 | Core 04 §3 Attribution and audit requirements | base | AC-437, AC-439, AC-440 |
| REQ-04-087 | Core 04 §12.3.2 First-admin bootstrap binding | base | AC-343..AC-346 |
| REQ-04-088 | Core 04 §12.3.2 First-admin bootstrap binding | base | AC-343..AC-344 |
| REQ-04-089 | Core 04 §12.3.2 First-admin bootstrap binding | base | AC-343..AC-345 |
| REQ-04-090 | Core 04 §12.3.2 First-admin bootstrap binding | base | AC-343..AC-346 |
| REQ-04-091 | Core 04 §12.3.2 First-admin bootstrap binding | base | AC-345 |
| REQ-04-092 | Core 04 §12.3.2 First-admin bootstrap binding | base | AC-344, AC-346 |



| REQ-04-105 | Core 04 §2 Authorization model | base | AC-370..AC-371, AC-427 |
| REQ-04-106 | Core 04 §2 Authorization model | base | AC-402, AC-427, AC-428 |
| REQ-04-113 | Core 04 §2 Authorization model | base | AC-428 |
| REQ-04-107 | Core 04 §12.3.3 Backup storage binding | base | AC-403 |
| REQ-04-108 | Core 04 §12.3.3 Backup storage binding | base | AC-403 |
| REQ-04-115 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-433, AC-436 |
| REQ-04-116 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-433..AC-434, AC-436 |
| REQ-04-117 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-434 |
| REQ-04-118 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-434..AC-435 |
| REQ-04-119 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-434..AC-435 |
| REQ-04-120 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-434..AC-435 |
| REQ-04-121 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-288, AC-348, AC-435..AC-436 |
| REQ-04-122 | Core 04 §12.3.4 Enterprise-auth provider manifest binding | enterprise_authentication | AC-352, AC-427, AC-435..AC-436 |
| REQ-04-109 | Core 04 §4 Trust boundaries | base | AC-413 |
| REQ-04-110 | Core 04 §12.3 Key registry and binding model | base | AC-131, AC-294, AC-298 |
| REQ-05-001 | Core 05 §1 Scope and separation | claim_publication | PC-006 |
| REQ-05-002 | Core 05 §1 Scope and separation | claim_publication | PC-006 |
| REQ-05-003 | Core 05 §1 Scope and separation | claim_publication | PC-001, PC-002, PC-006 |
| REQ-05-004 | Core 05 §2 Benchmark fixtures and observable-timing rules | claim_publication | PC-003 |
| REQ-05-005 | Core 05 §2 Benchmark fixtures and observable-timing rules | claim_publication | PC-003 |
| REQ-05-006 | Core 05 §3 Benchmark profile registry and benchmark-manifest contract | claim_publication | PC-001, PC-002 |
| REQ-05-007 | Core 05 §3 Benchmark profile registry and benchmark-manifest contract | claim_publication | PC-002 |
| REQ-05-008 | Core 05 §3 Benchmark profile registry and benchmark-manifest contract | claim_publication | PC-002 |
| REQ-05-009 | Core 05 §3 Benchmark profile registry and benchmark-manifest contract | claim_publication | PC-001, PC-005 |
| REQ-05-010 | Core 05 §3 Benchmark profile registry and benchmark-manifest contract | claim_publication | PC-004 |
| REQ-05-011 | Core 05 §4 Measurement-predicate registry | claim_publication | PC-003 |
| REQ-05-012 | Core 05 §4 Measurement-predicate registry | claim_publication | PC-002, PC-003 |
| REQ-05-013 | Core 05 §4 Measurement-predicate registry | claim_publication | PC-002 |

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
| AC-013 | REQ-01-350, REQ-03-033..REQ-03-035, REQ-03-086, REQ-03-223..REQ-03-224 |
| AC-014 | REQ-03-223..REQ-03-224 |
| AC-015 | REQ-01-243..REQ-01-247, REQ-01-291..REQ-01-295, REQ-01-355..REQ-01-366, REQ-01-487..REQ-01-488, REQ-01-490, REQ-01-492..REQ-01-493, REQ-02-186..REQ-02-201, REQ-03-120..REQ-03-126, REQ-03-242..REQ-03-246 |
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
| AC-043 | REQ-00-013, REQ-01-015..REQ-01-017, REQ-03-001..REQ-03-003, REQ-03-087..REQ-03-089, REQ-03-217..REQ-03-219, REQ-03-263, REQ-03-282 |
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
| AC-068 | REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-01-487..REQ-01-488, REQ-01-490..REQ-01-491, REQ-02-067..REQ-02-071 |
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
| AC-085 | REQ-01-296..REQ-01-302, REQ-01-336..REQ-01-338, REQ-01-487..REQ-01-488, REQ-01-490..REQ-01-493, REQ-01-496, REQ-02-094..REQ-02-109, REQ-02-120..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-086 | REQ-01-296..REQ-01-302, REQ-01-339..REQ-01-341, REQ-01-487..REQ-01-488, REQ-01-490..REQ-01-491, REQ-02-094, REQ-02-110..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-087 | REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-088 | REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-123..REQ-02-133, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-089 | REQ-01-296..REQ-01-302, REQ-02-067..REQ-02-071, REQ-02-094, REQ-02-123..REQ-02-134, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260 |
| AC-090 | REQ-01-296..REQ-01-302, REQ-02-094, REQ-02-120..REQ-02-122, REQ-03-005..REQ-03-011, REQ-03-255..REQ-03-260, REQ-03-264 |
| AC-091 | REQ-01-377..REQ-01-384, REQ-02-139..REQ-02-146, REQ-04-044..REQ-04-047 |
| AC-092 | REQ-01-282..REQ-01-284, REQ-01-400..REQ-01-406, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055 |
| AC-093 | REQ-01-407..REQ-01-413, REQ-02-247, REQ-04-040..REQ-04-043 |
| AC-094 | REQ-01-407..REQ-01-418, REQ-02-247, REQ-04-040..REQ-04-043 |
| AC-095 | REQ-01-409..REQ-01-421, REQ-02-247, REQ-04-040..REQ-04-043 |
| AC-096 | REQ-01-282..REQ-01-284, REQ-01-407..REQ-01-413, REQ-04-040..REQ-04-043, REQ-04-054..REQ-04-055 |
| AC-097 | REQ-01-323..REQ-01-325, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246 |
| AC-098 | REQ-01-326..REQ-01-327, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246 |
| AC-099 | REQ-01-157..REQ-01-158, REQ-01-173..REQ-01-175, REQ-01-491..REQ-01-491.1, REQ-02-009..REQ-02-023, REQ-03-242..REQ-03-246 |
| AC-100 | REQ-01-328, REQ-01-355..REQ-01-366, REQ-01-492..REQ-01-493, REQ-02-009..REQ-02-023, REQ-02-186..REQ-02-201, REQ-03-242..REQ-03-246 |
| AC-101 | REQ-02-009..REQ-02-023, REQ-02-135..REQ-02-138 |
| AC-102 | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-120 |
| AC-103 | REQ-01-243..REQ-01-247, REQ-02-186..REQ-02-201, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128 |
| AC-104 | REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035 |
| AC-105 | REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035 |
| AC-106 | REQ-01-374..REQ-01-376, REQ-02-139..REQ-02-146, REQ-04-031..REQ-04-035 |
| AC-107 | REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126 |
| AC-108 | REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126 |
| AC-109 | REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126 |
| AC-110 | REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126 |
| AC-111 | REQ-00-016..REQ-00-017, REQ-02-189..REQ-02-196, REQ-03-102..REQ-03-110, REQ-03-121..REQ-03-126 |
| AC-112 | REQ-01-303..REQ-01-306, REQ-01-329..REQ-01-330, REQ-01-487..REQ-01-488, REQ-01-490..REQ-01-491, REQ-02-067..REQ-02-071, REQ-03-004, REQ-03-242..REQ-03-246 |
| AC-113 | REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-114 | REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-115 | REQ-01-377..REQ-01-380, REQ-01-385..REQ-01-393, REQ-02-139..REQ-02-146, REQ-02-211, REQ-04-044..REQ-04-047 |
| AC-116 | REQ-00-014, REQ-01-240..REQ-01-242, REQ-01-285..REQ-01-296, REQ-01-307..REQ-01-310, REQ-01-499, REQ-01-503..REQ-01-509, REQ-03-004, REQ-03-242..REQ-03-246 |
| AC-117 | REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-310, REQ-01-499..REQ-01-500, REQ-01-503..REQ-01-509, REQ-03-242..REQ-03-246 |
| AC-118 | REQ-00-014, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-310, REQ-01-323..REQ-01-341, REQ-01-487..REQ-01-509, REQ-01-568, REQ-02-028..REQ-02-029, REQ-02-202..REQ-02-204, REQ-03-052..REQ-03-053, REQ-03-242..REQ-03-246 |
| AC-119 | REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-322, REQ-03-236..REQ-03-241 |
| AC-120 | REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-311, REQ-03-236..REQ-03-241 |
| AC-121 | REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-332..REQ-01-335, REQ-02-083..REQ-02-093, REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011, REQ-03-250..REQ-03-254 |
| AC-122 | REQ-01-296..REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-331, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082, REQ-02-222..REQ-02-223, REQ-03-005..REQ-03-011 |
| AC-123 | REQ-00-014, REQ-01-023..REQ-01-031, REQ-04-001..REQ-04-017 |
| AC-124 | REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-341, REQ-01-349, REQ-03-223..REQ-03-224, REQ-03-236..REQ-03-241 |
| AC-125 | REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-057..REQ-01-088, REQ-01-285..REQ-01-290, REQ-01-307..REQ-01-330, REQ-01-349..REQ-01-350, REQ-02-208..REQ-02-209, REQ-03-086, REQ-03-111..REQ-03-115, REQ-03-236..REQ-03-241, REQ-03-282 |
| AC-126 | REQ-01-019, REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-03-036..REQ-03-040, REQ-03-063..REQ-03-076, REQ-03-282 |
| AC-127 | REQ-00-014, REQ-01-019..REQ-01-022, REQ-01-034..REQ-01-056, REQ-01-117..REQ-01-118, REQ-01-129, REQ-01-144, REQ-01-168, REQ-01-240..REQ-01-242, REQ-01-288 |
| AC-128 | REQ-00-014, REQ-01-019, REQ-01-243..REQ-01-247, REQ-01-328, REQ-01-355..REQ-01-366, REQ-02-186..REQ-02-204, REQ-03-116..REQ-03-119, REQ-03-121..REQ-03-128, REQ-04-048 |
| AC-129 | REQ-00-014, REQ-01-018..REQ-01-019, REQ-01-248..REQ-01-277, REQ-01-452..REQ-01-454, REQ-03-092..REQ-03-098 |
| AC-130 | REQ-01-023..REQ-01-031, REQ-01-154, REQ-04-001..REQ-04-004, REQ-04-052..REQ-04-053 |
| AC-131 | REQ-01-019..REQ-01-022, REQ-01-250..REQ-01-277, REQ-03-092..REQ-03-098, REQ-04-005..REQ-04-017, REQ-04-052..REQ-04-053, REQ-04-110 |
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
| AC-152 | REQ-01-138..REQ-01-151, REQ-01-487..REQ-01-489, REQ-02-147..REQ-02-157, REQ-03-012..REQ-03-016, REQ-03-022..REQ-03-026 |
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
| AC-170 | REQ-01-152..REQ-01-180, REQ-01-491..REQ-01-491.1, REQ-02-203, REQ-02-222..REQ-02-223, REQ-02-243 |
| AC-171 | REQ-01-152..REQ-01-180, REQ-01-240..REQ-01-242 |
| AC-172 | REQ-01-152..REQ-01-180 |
| AC-173 | REQ-01-152..REQ-01-180 |
| AC-174 | REQ-01-152..REQ-01-180, REQ-02-243 |
| AC-175 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-240..REQ-01-242, REQ-01-487..REQ-01-489, REQ-01-497, REQ-01-521, REQ-02-237, REQ-02-244..REQ-02-245, REQ-04-038 |
| AC-176 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-487..REQ-01-489, REQ-01-497, REQ-01-521, REQ-02-244 |
| AC-177 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-02-245 |
| AC-178 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-01-240..REQ-01-242, REQ-01-497, REQ-02-244, REQ-04-021..REQ-04-030 |
| AC-179 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030 |
| AC-180 | REQ-01-032..REQ-01-033, REQ-01-112..REQ-01-137, REQ-04-021..REQ-04-030 |
| AC-181 | REQ-01-057..REQ-01-088, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-210, REQ-03-282 |
| AC-182 | REQ-01-057..REQ-01-088, REQ-01-104, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-210, REQ-03-101 |
| AC-183 | REQ-01-057..REQ-01-088, REQ-02-210, REQ-03-282 |
| AC-184 | REQ-01-034..REQ-01-056, REQ-01-310, REQ-01-312..REQ-01-322, REQ-01-488, REQ-03-223..REQ-03-224 |
| AC-185 | REQ-01-034..REQ-01-056, REQ-01-303..REQ-01-306, REQ-01-310, REQ-01-329..REQ-01-330, REQ-01-488, REQ-02-067..REQ-02-071, REQ-03-223..REQ-03-224 |
| AC-186 | REQ-01-032..REQ-01-033, REQ-01-181..REQ-01-195, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-054..REQ-02-055, REQ-02-060..REQ-02-061, REQ-02-064..REQ-02-066, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249 |
| AC-187 | REQ-01-032..REQ-01-033, REQ-01-104, REQ-01-181..REQ-01-195, REQ-01-237, REQ-02-060..REQ-02-061, REQ-02-064..REQ-02-066, REQ-03-101, REQ-03-247..REQ-03-249 |
| AC-188 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-02-058..REQ-02-059, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241 |
| AC-189 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241 |
| AC-190 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-02-202..REQ-02-204, REQ-03-129..REQ-03-134, REQ-03-209..REQ-03-216, REQ-03-236..REQ-03-241 |
| AC-191 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241 |
| AC-192 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110, REQ-03-236..REQ-03-241 |
| AC-193 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-115, REQ-03-236..REQ-03-241 |
| AC-194 | REQ-01-312..REQ-01-322, REQ-01-487..REQ-01-488, REQ-01-496, REQ-03-102..REQ-03-110 |
| AC-195 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-196 | REQ-01-086, REQ-01-311..REQ-01-312, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-169, REQ-02-175, REQ-02-181, REQ-03-102..REQ-03-106 |
| AC-197 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-198 | REQ-01-312..REQ-01-322, REQ-03-102..REQ-03-110 |
| AC-199 | REQ-01-087, REQ-03-102..REQ-03-110 |
| AC-200 | REQ-01-057..REQ-01-088, REQ-01-487..REQ-01-488, REQ-01-494, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209, REQ-03-282 |
| AC-201 | REQ-01-057..REQ-01-088, REQ-02-030..REQ-02-036, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209 |
| AC-202 | REQ-01-057..REQ-01-088, REQ-01-487..REQ-01-488, REQ-01-495, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209 |
| AC-203 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209, REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076 |
| AC-204 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-202..REQ-02-204, REQ-02-208..REQ-02-209, REQ-03-048..REQ-03-053, REQ-03-063..REQ-03-076 |
| AC-205 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 |
| AC-206 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209 |
| AC-207 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209 |
| AC-208 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209 |
| AC-209 | REQ-01-057..REQ-01-088, REQ-01-181..REQ-01-195, REQ-02-054..REQ-02-055, REQ-02-064..REQ-02-066, REQ-02-163..REQ-02-185, REQ-02-219..REQ-02-220, REQ-03-247..REQ-03-249 |
| AC-210 | REQ-01-057..REQ-01-088, REQ-01-342..REQ-01-348, REQ-01-351..REQ-01-353, REQ-01-355..REQ-01-366, REQ-02-163..REQ-02-185, REQ-03-247..REQ-03-249 |
| AC-211 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023 |
| AC-212 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-491..REQ-01-491.1, REQ-02-014..REQ-02-023, REQ-02-222..REQ-02-223 |
| AC-213 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-02-014..REQ-02-023 |
| AC-214 | REQ-01-057..REQ-01-088, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-01-491..REQ-01-491.1, REQ-02-014..REQ-02-023, REQ-02-222..REQ-02-223 |
| AC-215 | REQ-01-056..REQ-01-111, REQ-01-240..REQ-01-242, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-218, REQ-03-139..REQ-03-142 |
| AC-216 | REQ-01-057..REQ-01-111, REQ-01-487..REQ-01-488, REQ-01-491, REQ-01-496, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-216, REQ-03-141..REQ-03-144 |
| AC-217 | REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-218, REQ-03-141..REQ-03-144 |
| AC-218 | REQ-01-057..REQ-01-111, REQ-01-228..REQ-01-239, REQ-02-205..REQ-02-207, REQ-03-101, REQ-03-141..REQ-03-144 |
| AC-219 | REQ-01-021, REQ-01-152..REQ-01-180, REQ-01-228..REQ-01-239, REQ-01-491..REQ-01-491.1, REQ-02-222..REQ-02-223 |
| AC-220 | REQ-01-021, REQ-01-152..REQ-01-180, REQ-01-491..REQ-01-491.1, REQ-02-222..REQ-02-223 |
| AC-221 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134 |
| AC-222 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134 |
| AC-223 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-026..REQ-02-036, REQ-02-039..REQ-02-044, REQ-03-129..REQ-03-134 |
| AC-224 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041, REQ-03-129..REQ-03-134 |
| AC-225 | REQ-01-057..REQ-01-088, REQ-01-196..REQ-01-227, REQ-01-487..REQ-01-488, REQ-01-496, REQ-02-030..REQ-02-036, REQ-02-039..REQ-02-041, REQ-03-129..REQ-03-134 |
| AC-226 | REQ-03-048..REQ-03-076 |
| AC-227 | REQ-03-048..REQ-03-076 |
| AC-228 | REQ-03-048..REQ-03-076 |
| AC-229 | REQ-03-048..REQ-03-076 |
| AC-230 | REQ-03-048..REQ-03-076 |
| AC-231 | REQ-00-001..REQ-00-003, REQ-00-005..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-020, REQ-00-022..REQ-00-027, REQ-00-029..REQ-00-042, REQ-00-044..REQ-00-046, REQ-00-049..REQ-00-051, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-465, REQ-01-487..REQ-01-509, REQ-01-516..REQ-01-521, REQ-01-554..REQ-01-563, REQ-01-568, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-233, REQ-02-238..REQ-02-246, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-282, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-061, REQ-04-065..REQ-04-092 |
| AC-232 | REQ-00-004, REQ-00-047, REQ-01-006..REQ-01-014, REQ-01-033, REQ-01-457, REQ-02-032, REQ-02-045..REQ-02-053, REQ-02-202, REQ-03-146..REQ-03-147, REQ-03-153..REQ-03-204, REQ-03-208, REQ-04-053 |
| AC-233 | REQ-00-004, REQ-00-015, REQ-00-047, REQ-01-004, REQ-01-018, REQ-01-033, REQ-01-068, REQ-01-078, REQ-01-098, REQ-01-251, REQ-01-263, REQ-01-265, REQ-01-270, REQ-01-273..REQ-01-274, REQ-01-278..REQ-01-279, REQ-01-368..REQ-01-398, REQ-01-452, REQ-01-457, REQ-02-009, REQ-02-112, REQ-02-139..REQ-02-146, REQ-02-200, REQ-02-202, REQ-02-206..REQ-02-207, REQ-02-211, REQ-02-217, REQ-03-021, REQ-03-078, REQ-03-094, REQ-04-027, REQ-04-031..REQ-04-035, REQ-04-044..REQ-04-047 |
| AC-234 | REQ-00-004, REQ-00-013, REQ-00-015, REQ-00-043, REQ-00-047, REQ-01-018, REQ-01-033, REQ-01-278, REQ-01-282..REQ-01-284, REQ-01-286, REQ-01-308, REQ-01-310, REQ-01-399..REQ-01-422, REQ-01-452, REQ-01-455, REQ-02-001, REQ-02-003, REQ-02-009, REQ-02-202, REQ-02-221, REQ-02-247, REQ-04-040..REQ-04-043, REQ-04-053, REQ-04-055, REQ-04-058 |
| AC-235 | REQ-00-004, REQ-00-021, REQ-00-048, REQ-01-031, REQ-01-033, REQ-01-510..REQ-01-515, REQ-04-018..REQ-04-020 |
| AC-236 | REQ-00-004, REQ-00-047, REQ-01-033, REQ-01-425..REQ-01-450, REQ-01-457, REQ-01-564, REQ-02-204, REQ-02-222, REQ-02-249, REQ-04-038, REQ-04-044..REQ-04-047 |
| AC-238 | REQ-01-035..REQ-01-036, REQ-01-240, REQ-01-242 |
| AC-239 | REQ-01-035, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242 |
| AC-240 | REQ-01-035, REQ-01-234, REQ-01-238, REQ-01-240 |
| AC-241 | REQ-01-036, REQ-01-242 |
| AC-242 | REQ-01-242 |
| AC-243 | REQ-01-035..REQ-01-037 |
| AC-244 | REQ-01-025, REQ-01-120, REQ-01-521 |
| AC-245 | REQ-01-025, REQ-01-120, REQ-01-234, REQ-01-521 |
| AC-246 | REQ-01-025, REQ-01-234 |
| AC-247 | REQ-01-025, REQ-01-234, REQ-01-497 |
| AC-248 | REQ-01-025 |
| AC-249 | REQ-01-025, REQ-01-234 |
| AC-250 | REQ-01-025, REQ-01-031, REQ-01-234 |
| AC-251 | REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-459, REQ-01-465 |
| AC-252 | REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-247, REQ-01-458, REQ-01-460..REQ-01-461, REQ-01-465, REQ-02-222..REQ-02-223, REQ-03-127..REQ-03-128, REQ-04-053 |
| AC-253 | REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-458, REQ-01-460, REQ-01-462, REQ-01-465, REQ-02-222..REQ-02-223, REQ-03-128, REQ-04-053 |
| AC-254 | REQ-01-032, REQ-01-234, REQ-01-247, REQ-01-458, REQ-01-462..REQ-01-463, REQ-01-465, REQ-03-128, REQ-04-023, REQ-04-053 |
| AC-255 | REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-247, REQ-01-459, REQ-01-463, REQ-01-465, REQ-03-127, REQ-04-023, REQ-04-053 |
| AC-256 | REQ-01-460, REQ-01-464 |
| AC-257 | REQ-01-248..REQ-01-249, REQ-01-268, REQ-04-023 |
| AC-258 | REQ-01-249, REQ-01-268, REQ-01-453 |
| AC-259 | REQ-01-249, REQ-01-268 |
| AC-260 | REQ-01-234, REQ-01-238, REQ-01-249, REQ-01-453, REQ-04-023 |
| AC-261 | REQ-01-234, REQ-01-249, REQ-04-023, REQ-04-029 |
| AC-262 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-472..REQ-01-473, REQ-01-549..REQ-01-553 |
| AC-263 | REQ-01-466, REQ-01-468..REQ-01-469, REQ-01-472, REQ-01-474 |
| AC-264 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-472, REQ-01-474 |
| AC-265 | REQ-01-466, REQ-01-471, REQ-01-475, REQ-01-553 |
| AC-266 | REQ-01-466..REQ-01-470, REQ-01-476..REQ-01-477 |
| AC-267 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-476..REQ-01-477 |
| AC-268 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-476, REQ-01-478 |
| AC-269 | REQ-01-466, REQ-01-471, REQ-01-479 |
| AC-270 | REQ-01-466..REQ-01-470, REQ-01-480..REQ-01-481, REQ-01-549..REQ-01-553 |
| AC-271 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-480..REQ-01-481 |
| AC-272 | REQ-01-466, REQ-01-471, REQ-01-482, REQ-01-553 |
| AC-273 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-483..REQ-01-484 |
| AC-274 | REQ-01-466..REQ-01-469, REQ-01-483..REQ-01-484 |
| AC-275 | REQ-01-466..REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-483, REQ-01-485, REQ-01-549..REQ-01-553 |
| AC-276 | REQ-01-466, REQ-01-471, REQ-01-486, REQ-01-553 |
| AC-277 | REQ-01-296, REQ-01-343, REQ-01-497..REQ-01-501, REQ-02-003, REQ-02-006, REQ-02-009, REQ-02-022, REQ-02-202, REQ-02-222..REQ-02-225, REQ-03-005, REQ-03-266 |
| AC-278 | REQ-01-328, REQ-01-336, REQ-01-502, REQ-02-017, REQ-02-021..REQ-02-022, REQ-02-120, REQ-02-124..REQ-02-125, REQ-02-197, REQ-02-199, REQ-02-202, REQ-02-226..REQ-02-228, REQ-03-247, REQ-03-256, REQ-03-268..REQ-03-271 |
| AC-279 | REQ-01-328, REQ-01-336, REQ-01-497, REQ-01-502, REQ-02-022, REQ-02-060..REQ-02-063, REQ-02-229..REQ-02-230, REQ-02-232, REQ-03-247, REQ-03-256, REQ-03-259, REQ-03-267..REQ-03-268, REQ-03-270..REQ-03-271 |
| AC-280 | REQ-01-328, REQ-01-336, REQ-02-021..REQ-02-022, REQ-02-063, REQ-02-197..REQ-02-198, REQ-02-202, REQ-02-226, REQ-02-231..REQ-02-232, REQ-04-022, REQ-04-024 |
| AC-281 | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-503, REQ-02-123..REQ-02-125, REQ-02-132..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 |
| AC-282 | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-504, REQ-02-123, REQ-02-126..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 |
| AC-283 | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-505, REQ-02-123, REQ-02-128..REQ-02-133, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 |
| AC-284 | REQ-01-302, REQ-01-307..REQ-01-311, REQ-01-358, REQ-01-506, REQ-02-123, REQ-02-130..REQ-02-133, REQ-02-222, REQ-03-010..REQ-03-011, REQ-03-259, REQ-03-265 |
| AC-285 | REQ-01-308..REQ-01-310, REQ-01-358, REQ-01-507, REQ-02-135..REQ-02-136, REQ-02-222, REQ-03-259, REQ-03-265 |
| AC-286 | REQ-01-308..REQ-01-310, REQ-01-358, REQ-01-508, REQ-02-135, REQ-02-137, REQ-03-259, REQ-03-265 |
| AC-287 | REQ-01-308..REQ-01-310, REQ-01-358, REQ-01-509, REQ-02-135, REQ-02-138, REQ-02-222, REQ-03-259, REQ-03-265 |
| AC-288 | REQ-01-510..REQ-01-511, REQ-04-018, REQ-04-121 |
| AC-289 | REQ-01-510..REQ-01-511, REQ-01-514..REQ-01-515, REQ-01-580, REQ-04-018 |
| AC-290 | REQ-01-031, REQ-01-510, REQ-01-512, REQ-01-515, REQ-01-580, REQ-03-030..REQ-03-031, REQ-04-018, REQ-04-020 |
| AC-291 | REQ-01-031, REQ-01-510, REQ-01-512, REQ-01-515, REQ-01-580, REQ-03-030..REQ-03-031, REQ-04-018, REQ-04-020 |
| AC-292 | REQ-01-513, REQ-04-019 |
| AC-293 | REQ-01-234, REQ-01-238, REQ-01-513..REQ-01-514, REQ-04-018..REQ-04-020 |
| AC-348 | REQ-01-033, REQ-01-116, REQ-01-537..REQ-01-538, REQ-02-234, REQ-02-236..REQ-02-237, REQ-04-093, REQ-04-121 |
| AC-349 | REQ-01-234, REQ-01-238, REQ-01-538, REQ-01-541, REQ-02-234..REQ-02-235 |
| AC-350 | REQ-01-513..REQ-01-514, REQ-01-539..REQ-01-541, REQ-02-234..REQ-02-235, REQ-04-020, REQ-04-095 |
| AC-351 | REQ-01-116, REQ-01-234, REQ-01-238, REQ-01-540..REQ-01-541, REQ-02-234..REQ-02-237, REQ-04-095 |
| AC-352 | REQ-01-537..REQ-01-541, REQ-02-236, REQ-04-093..REQ-04-096, REQ-04-122 |
| AC-294 | REQ-01-455, REQ-04-058, REQ-04-066..REQ-04-071, REQ-04-077, REQ-04-110 |
| AC-295 | REQ-01-455, REQ-04-058, REQ-04-069, REQ-04-071..REQ-04-073, REQ-04-077 |
| AC-296 | REQ-01-456, REQ-04-059, REQ-04-074..REQ-04-075, REQ-04-077 |
| AC-297 | REQ-01-455, REQ-04-058, REQ-04-067, REQ-04-069, REQ-04-071..REQ-04-076 |
| AC-298 | REQ-04-066, REQ-04-077..REQ-04-078, REQ-04-110, REQ-04-115..REQ-04-122 |
| AC-299 | REQ-01-058, REQ-01-063, REQ-01-069..REQ-01-070 |
| AC-300 | REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-487..REQ-01-488, REQ-01-503..REQ-01-506 |
| AC-301 | REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-336, REQ-01-488, REQ-01-503..REQ-01-505 |
| AC-302 | REQ-01-310, REQ-01-332, REQ-01-339, REQ-01-488, REQ-01-503..REQ-01-506 |
| AC-303 | REQ-01-310, REQ-01-312, REQ-01-328, REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-487..REQ-01-488, REQ-01-503..REQ-01-506 |
| AC-354 | REQ-03-281 |
| AC-304 | REQ-01-336, REQ-01-338 |
| AC-305 | REQ-01-469..REQ-01-470, REQ-01-478, REQ-01-496 |
| AC-306 | REQ-01-478, REQ-04-034 |
| AC-307 | REQ-01-471, REQ-01-479 |
| AC-308 | REQ-01-469..REQ-01-470, REQ-01-481, REQ-01-496 |
| AC-309 | REQ-01-467, REQ-01-481 |
| AC-369 | REQ-01-467, REQ-01-469..REQ-01-470, REQ-01-481 |
| AC-310 | REQ-01-471, REQ-01-482 |
| AC-311 | REQ-01-025, REQ-01-497 |
| AC-312 | REQ-01-119..REQ-01-120, REQ-01-122, REQ-01-124, REQ-01-497 |
| AC-313 | REQ-00-016, REQ-00-019, REQ-02-189..REQ-02-196, REQ-03-121..REQ-03-126 |
| AC-314 | REQ-02-114 |
| AC-315 | REQ-01-059..REQ-01-060, REQ-01-328, REQ-01-336, REQ-01-516..REQ-01-517, REQ-01-519..REQ-01-520, REQ-03-272..REQ-03-274 |
| AC-316 | REQ-01-061, REQ-01-328, REQ-01-336, REQ-01-517, REQ-01-519..REQ-01-520 |
| AC-317 | REQ-01-328, REQ-01-336, REQ-01-517..REQ-01-520 |
| AC-318 | REQ-01-328, REQ-01-336, REQ-01-502, REQ-02-017, REQ-02-021..REQ-02-022, REQ-02-231, REQ-03-272, REQ-03-274 |
| AC-319 | REQ-01-336, REQ-01-517..REQ-01-520, REQ-02-233, REQ-03-273 |
| AC-320 | REQ-04-066, REQ-04-077, REQ-04-079..REQ-04-081 |
| AC-321 | REQ-01-234, REQ-01-238, REQ-01-243..REQ-01-245, REQ-04-079..REQ-04-080 |
| AC-322 | REQ-01-238, REQ-01-461, REQ-01-465, REQ-04-079..REQ-04-080 |
| AC-323 | REQ-01-234, REQ-01-238, REQ-01-473..REQ-01-475, REQ-04-079..REQ-04-081 |
| AC-324 | REQ-01-234, REQ-01-238, REQ-01-474..REQ-01-475, REQ-04-079..REQ-04-080 |
| AC-325 | REQ-01-234, REQ-01-238, REQ-01-474..REQ-01-475, REQ-04-079..REQ-04-080 |
| AC-463 | REQ-01-618, REQ-02-259, REQ-03-293 |
| AC-464 | REQ-01-618, REQ-01-620, REQ-03-293 |
| AC-465 | REQ-01-618..REQ-01-619, REQ-02-259 |
| AC-466 | REQ-01-619..REQ-01-620, REQ-03-293 |
| AC-467 | REQ-01-619, REQ-02-259, REQ-03-293 |
| AC-468 | REQ-01-618, REQ-03-293 |
| AC-469 | REQ-00-062..REQ-00-063 |
| AC-470 | REQ-00-063, REQ-01-621..REQ-01-622 |
| AC-471 | REQ-00-026, REQ-01-623 |
| AC-472 | REQ-00-023, REQ-00-063, REQ-01-621, REQ-01-624..REQ-01-625 |
| AC-473 | REQ-00-063, REQ-01-626 |
| AC-326 | REQ-01-234, REQ-01-238, REQ-01-481..REQ-01-482, REQ-04-079..REQ-04-080 |
| AC-327 | REQ-01-238, REQ-01-448..REQ-01-449, REQ-01-486, REQ-04-079..REQ-04-080 |
| AC-328 | REQ-01-234, REQ-01-238, REQ-01-448..REQ-01-449, REQ-01-486, REQ-04-079..REQ-04-080 |
| AC-329 | REQ-01-086..REQ-01-087, REQ-02-168, REQ-03-106 |
| AC-330 | REQ-01-087 |
| AC-331 | REQ-01-086, REQ-01-311..REQ-01-312, REQ-02-181, REQ-03-106..REQ-03-107 |
| AC-332 | REQ-01-311..REQ-01-312, REQ-01-448..REQ-01-449, REQ-01-486, REQ-02-169, REQ-02-175, REQ-02-181 |
| AC-333 | REQ-01-377..REQ-01-380, REQ-02-139, REQ-02-143, REQ-02-211, REQ-04-045 |
| AC-334 | REQ-01-024..REQ-01-025, REQ-01-234, REQ-01-522, REQ-04-083, REQ-04-084 |
| AC-335 | REQ-01-024, REQ-01-522..REQ-01-523, REQ-02-222, REQ-02-245, REQ-04-083, REQ-04-084 |
| AC-336 | REQ-01-024, REQ-01-234, REQ-01-522, REQ-01-525, REQ-02-245, REQ-04-083, REQ-04-084, REQ-04-086 |
| AC-337 | REQ-01-024, REQ-01-234, REQ-01-238, REQ-01-522, REQ-01-526, REQ-02-245, REQ-04-083, REQ-04-084, REQ-04-086 |
| AC-338 | REQ-01-024, REQ-01-234, REQ-01-522, REQ-01-524, REQ-02-245, REQ-04-016, REQ-04-083, REQ-04-084, REQ-04-086 |
| AC-339 | REQ-01-234, REQ-01-238, REQ-01-522..REQ-01-526, REQ-02-245, REQ-04-083, REQ-04-084 |
| AC-340 | REQ-01-032, REQ-01-527, REQ-02-245, REQ-04-016, REQ-04-083, REQ-04-085, REQ-04-086 |
| AC-341 | REQ-01-025, REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-528, REQ-02-245, REQ-04-001, REQ-04-016, REQ-04-083, REQ-04-084, REQ-04-085, REQ-04-086 |
| AC-342 | REQ-01-032, REQ-01-529, REQ-02-245, REQ-04-016, REQ-04-083, REQ-04-085, REQ-04-086 |
| AC-343 | REQ-01-121, REQ-01-530..REQ-01-536, REQ-02-007..REQ-02-008, REQ-02-202, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-087..REQ-04-090 |
| AC-344 | REQ-01-121, REQ-01-530..REQ-01-535, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-087..REQ-04-092 |
| AC-345 | REQ-01-121, REQ-01-530, REQ-01-533, REQ-01-535, REQ-04-028, REQ-04-089..REQ-04-091 |
| AC-346 | REQ-01-121, REQ-01-530, REQ-01-533..REQ-01-535, REQ-02-007..REQ-02-008, REQ-02-202, REQ-02-246, REQ-04-028, REQ-04-038, REQ-04-090, REQ-04-092 |
| AC-347 | REQ-01-121, REQ-01-536, REQ-04-084 |
| AC-353 | REQ-01-104, REQ-03-101 |
| AC-359 | REQ-01-035, REQ-01-286, REQ-01-310 |
| AC-360 | REQ-01-035, REQ-01-038, REQ-01-046, REQ-01-142, REQ-01-145..REQ-01-146, REQ-02-010, REQ-02-155, REQ-03-224, REQ-03-227 |
| AC-361 | REQ-01-035..REQ-01-036, REQ-01-046 |
| AC-362 | REQ-01-286, REQ-01-310, REQ-01-312, REQ-01-323, REQ-01-326, REQ-01-328, REQ-01-329, REQ-01-331..REQ-01-332, REQ-01-336, REQ-01-339, REQ-01-499, REQ-01-503..REQ-01-509 |
| AC-363 | REQ-01-310, REQ-03-223 |
| AC-364 | REQ-03-225..REQ-03-235 |
| AC-365 | REQ-01-310 |
| AC-366 | REQ-01-034, REQ-01-036, REQ-01-310, REQ-01-328, REQ-01-336, REQ-03-247 |
| AC-367 | REQ-01-034, REQ-01-036, REQ-01-070, REQ-01-310, REQ-01-328, REQ-01-336 |
| AC-368 | REQ-01-034, REQ-01-267, REQ-01-310, REQ-03-097 |
| AC-370 | REQ-00-022, REQ-01-032..REQ-01-033, REQ-01-542..REQ-01-546, REQ-04-105 |
| AC-371 | REQ-00-022, REQ-01-033, REQ-01-234, REQ-01-544, REQ-01-546..REQ-01-548, REQ-04-105 |
| AC-372 | REQ-01-035, REQ-01-554..REQ-01-556 |
| AC-373 | REQ-01-035..REQ-01-036, REQ-01-555..REQ-01-557 |
| AC-374 | REQ-01-035..REQ-01-036, REQ-01-554..REQ-01-558, REQ-01-560 |
| AC-375 | REQ-01-035, REQ-01-238, REQ-01-554, REQ-01-558..REQ-01-560 |
| AC-376 | REQ-03-089, REQ-03-095, REQ-03-099..REQ-03-100 |
| AC-377 | REQ-03-099..REQ-03-100 |
| AC-378 | REQ-03-099..REQ-03-100 |
| AC-379 | REQ-03-099..REQ-03-100 |
| AC-380 | REQ-03-099..REQ-03-100 |
| AC-381 | REQ-03-072, REQ-03-099..REQ-03-100 |
| AC-382 | REQ-03-099..REQ-03-100 |
| AC-383 | REQ-01-054, REQ-01-561..REQ-01-562, REQ-02-238..REQ-02-240 |
| AC-384 | REQ-01-054, REQ-01-096, REQ-01-563, REQ-02-216, REQ-02-241 |
| AC-385 | REQ-02-239 |
| AC-386 | REQ-01-564, REQ-02-238, REQ-02-241 |
| AC-387 | REQ-01-035..REQ-01-047, REQ-01-565..REQ-01-567, REQ-03-223..REQ-03-224 |
| AC-388 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 |
| AC-389 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 |
| AC-390 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 |
| AC-391 | REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-02-208..REQ-02-209, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 |
| AC-392 | REQ-02-163..REQ-02-185, REQ-03-208, REQ-03-276 |
| AC-393 | REQ-01-010..REQ-01-014, REQ-02-032, REQ-03-208, REQ-03-276 |



| AC-394 | REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-163..REQ-02-180, REQ-02-248, REQ-03-280 |
| AC-395 | REQ-01-311, REQ-01-333..REQ-01-340, REQ-01-503..REQ-01-506, REQ-02-163..REQ-02-180, REQ-02-248, REQ-03-280 |
| AC-396 | REQ-01-569, REQ-03-280 |
| AC-397 | REQ-02-248 |
| AC-398 | REQ-01-572..REQ-01-573 |
| AC-399 | REQ-01-571, REQ-01-575, REQ-01-423..REQ-01-424, REQ-01-577 |
| AC-400 | REQ-01-575..REQ-01-576 |
| AC-401 | REQ-01-572, REQ-01-578 |
| AC-402 | REQ-01-570, REQ-04-106 |
| AC-428 | REQ-01-593..REQ-01-595, REQ-04-106, REQ-04-113 |
| AC-429 | REQ-00-054, REQ-01-032, REQ-01-597..REQ-01-599, REQ-02-255, REQ-04-114 |
| AC-430 | REQ-01-599, REQ-04-016, REQ-04-038, REQ-04-114 |
| AC-431 | REQ-01-600..REQ-01-601, REQ-02-255, REQ-03-289, REQ-04-016, REQ-04-114 |
| AC-432 | REQ-00-054, REQ-01-597..REQ-01-602, REQ-02-204, REQ-02-249, REQ-02-255, REQ-03-289, REQ-04-028, REQ-04-038, REQ-04-114 |
| AC-433 | REQ-00-055, REQ-04-020, REQ-04-077..REQ-04-078, REQ-04-115..REQ-04-116 |
| AC-434 | REQ-00-055, REQ-04-077..REQ-04-078, REQ-04-116..REQ-04-120 |
| AC-435 | REQ-00-055, REQ-01-511, REQ-01-538, REQ-02-234..REQ-02-237, REQ-02-256, REQ-04-077..REQ-04-078, REQ-04-121..REQ-04-122 |
| AC-436 | REQ-00-055, REQ-01-510..REQ-01-511, REQ-01-537..REQ-01-541, REQ-02-236, REQ-02-256, REQ-04-020, REQ-04-093, REQ-04-115..REQ-04-122 |
| AC-437 | REQ-00-056, REQ-01-032, REQ-01-603..REQ-01-606, REQ-02-257, REQ-04-124..REQ-04-125 |
| AC-438 | REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-583..REQ-01-584, REQ-01-607, REQ-04-123 |
| AC-439 | REQ-00-056, REQ-01-032, REQ-01-603, REQ-02-257, REQ-04-028..REQ-04-030, REQ-04-123, REQ-04-125 |
| AC-440 | REQ-00-056, REQ-01-432, REQ-01-571, REQ-01-604..REQ-01-606, REQ-02-202, REQ-02-204, REQ-02-249, REQ-02-257, REQ-04-038, REQ-04-086, REQ-04-096, REQ-04-124..REQ-04-125 |
| AC-441 | REQ-00-057, REQ-01-608, REQ-03-290, REQ-04-028..REQ-04-029, REQ-04-126 |
| AC-442 | REQ-00-058, REQ-01-448..REQ-01-450, REQ-01-485..REQ-01-486, REQ-01-609, REQ-03-290 |
| AC-443 | REQ-00-059, REQ-01-481, REQ-01-583..REQ-01-584, REQ-01-610 |
| AC-444 | REQ-01-307, REQ-01-312, REQ-01-614, REQ-03-236..REQ-03-241 |
| AC-445 | REQ-01-057..REQ-01-070, REQ-01-312, REQ-01-614 |
| AC-446 | REQ-01-312, REQ-03-236..REQ-03-241 |
| AC-447 | REQ-01-312, REQ-01-614, REQ-03-236..REQ-03-241 |
| AC-448 | REQ-01-312, REQ-01-614, REQ-03-276 |
| AC-449 | REQ-01-312, REQ-01-611..REQ-01-614, REQ-03-236..REQ-03-241 |
| AC-450 | REQ-01-057..REQ-01-080, REQ-04-021..REQ-04-030 |
| AC-451 | REQ-01-611..REQ-01-613, REQ-04-021..REQ-04-030 |
| AC-452 | REQ-00-014, REQ-01-312, REQ-01-614, REQ-03-236..REQ-03-241 |
| AC-453 | REQ-00-061, REQ-01-615..REQ-01-617, REQ-02-258, REQ-03-291..REQ-03-292, REQ-04-127 |
| AC-454 | REQ-01-615..REQ-01-617 |
| AC-455 | REQ-01-615..REQ-01-617 |
| AC-456 | REQ-03-291..REQ-03-292 |
| AC-457 | REQ-03-291..REQ-03-292 |
| AC-458 | REQ-01-617, REQ-03-292 |
| AC-459 | REQ-04-021..REQ-04-030, REQ-04-127 |
| AC-460 | REQ-01-615..REQ-01-617, REQ-04-127 |
| AC-461 | REQ-01-615..REQ-01-617 |
| AC-462 | REQ-01-615..REQ-01-617, REQ-03-291..REQ-03-292 |
| AC-403 | REQ-04-053, REQ-04-058, REQ-04-071..REQ-04-073, REQ-04-076, REQ-04-107..REQ-04-108 |
| AC-404 | REQ-01-001..REQ-01-003 |
| AC-405 | REQ-01-002, REQ-01-278..REQ-01-280 |
| AC-406 | REQ-02-024..REQ-02-025 |
| AC-407 | REQ-04-036 |
| AC-408 | REQ-04-037 |
| AC-409 | REQ-02-204, REQ-02-249, REQ-04-038 |
| AC-410 | REQ-01-309, REQ-02-134, REQ-02-250..REQ-02-253, REQ-03-004, REQ-03-011 |
| AC-411 | REQ-01-579, REQ-03-011 |
| AC-412 | REQ-01-057..REQ-01-111, REQ-02-205..REQ-02-207, REQ-02-212..REQ-02-220, REQ-03-141..REQ-03-144 |
| AC-413 | REQ-04-109 |
| AC-414 | REQ-00-053, REQ-00-057, REQ-01-025, REQ-01-168, REQ-01-580, REQ-01-608, REQ-03-030..REQ-03-031, REQ-04-028..REQ-04-029 |
| AC-415 | REQ-01-234, REQ-01-238, REQ-01-240, REQ-01-581..REQ-01-583 |
| AC-416 | REQ-01-168, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-581..REQ-01-584, REQ-02-222..REQ-02-223 |
| AC-417 | REQ-01-117, REQ-01-234, REQ-01-238, REQ-01-240..REQ-01-242, REQ-01-581..REQ-01-584, REQ-04-038 |
| AC-418 | REQ-00-035, REQ-01-032, REQ-01-234, REQ-01-238, REQ-01-496, REQ-01-585..REQ-01-586 |
| AC-419 | REQ-01-587..REQ-01-589, REQ-01-592 |
| AC-420 | REQ-01-587..REQ-01-589 |
| AC-421 | REQ-01-234, REQ-01-238, REQ-01-587..REQ-01-589, REQ-01-591 |
| AC-422 | REQ-01-234, REQ-01-585, REQ-01-590..REQ-01-591 |
| AC-423 | REQ-01-591 |
| AC-424 | REQ-01-168, REQ-01-585, REQ-01-590, REQ-03-287..REQ-03-288 |
| AC-425 | REQ-01-585, REQ-01-588, REQ-02-222..REQ-02-223, REQ-02-254 |
| AC-426 | REQ-01-592, REQ-03-288 |
| AC-427 | REQ-00-057, REQ-01-032..REQ-01-033, REQ-01-234, REQ-01-471, REQ-01-480..REQ-01-481, REQ-01-542..REQ-01-548, REQ-01-608, REQ-04-023, REQ-04-028..REQ-04-030, REQ-04-085, REQ-04-094, REQ-04-105..REQ-04-106, REQ-04-114, REQ-04-122, REQ-04-126 |
| PC-001 | REQ-05-003, REQ-05-006, REQ-05-009 |
| PC-002 | REQ-05-003, REQ-05-006..REQ-05-008, REQ-05-012..REQ-05-013 |
| PC-003 | REQ-05-004..REQ-05-005, REQ-05-011..REQ-05-012 |
| PC-004 | REQ-05-010 |
| PC-005 | REQ-05-009 |
| PC-006 | REQ-00-028, REQ-05-001..REQ-05-003 |

## F.6 Profile and companion claim Definition-of-Done navigation

| Profile or companion claim | Prerequisite claim | Required REQs | Required ACs |
| --- | --- | --- | --- |
| base | — | REQ-00-001..REQ-00-003, REQ-00-005..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-020, REQ-00-022..REQ-00-027, REQ-00-029..REQ-00-042, REQ-00-044..REQ-00-046, REQ-00-049..REQ-00-057, REQ-00-062..REQ-00-063, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-465, REQ-01-487..REQ-01-509, REQ-01-516..REQ-01-536, REQ-01-542..REQ-01-548, REQ-01-554..REQ-01-563, REQ-01-565..REQ-01-578, REQ-01-580..REQ-01-608, REQ-01-611..REQ-01-617, REQ-01-621..REQ-01-626, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-233, REQ-02-238..REQ-02-246, REQ-02-248, REQ-02-250..REQ-02-255, REQ-02-257..REQ-02-258, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-292, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-061, REQ-04-065..REQ-04-081, REQ-04-083..REQ-04-092, REQ-04-105..REQ-04-108, REQ-04-113..REQ-04-114, REQ-04-123..REQ-04-126 | AC-001..AC-026, AC-037..AC-055, AC-068..AC-070, AC-072..AC-090, AC-097..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-231, AC-238..AC-261, AC-277..AC-287, AC-294..AC-304, AC-311..AC-322, AC-329..AC-331, AC-334..AC-347, AC-353..AC-354, AC-359..AC-368, AC-370..AC-385, AC-387..AC-392, AC-394..AC-408, AC-410..AC-432, AC-437..AC-441, AC-444..AC-462, AC-469..AC-473 |
| import | base | REQ-00-001..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-020, REQ-00-022..REQ-00-027, REQ-00-029..REQ-00-042, REQ-00-044..REQ-00-047, REQ-00-049..REQ-00-054, REQ-00-057, REQ-00-062..REQ-00-063, REQ-01-001..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-475, REQ-01-487..REQ-01-509, REQ-01-516..REQ-01-536, REQ-01-542..REQ-01-563, REQ-01-565..REQ-01-578, REQ-01-580..REQ-01-584, REQ-01-608, REQ-01-618..REQ-01-626, REQ-02-001..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-233, REQ-02-238..REQ-02-246, REQ-02-248, REQ-02-250..REQ-02-253, REQ-02-259, REQ-03-001..REQ-03-281, REQ-03-290, REQ-03-293, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-061, REQ-04-065..REQ-04-081, REQ-04-083..REQ-04-092, REQ-04-105..REQ-04-108, REQ-04-126 | AC-001..AC-029, AC-037..AC-055, AC-063..AC-070, AC-072..AC-090, AC-097..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-232, AC-238..AC-265, AC-277..AC-287, AC-294..AC-304, AC-311..AC-325, AC-329..AC-331, AC-334..AC-347, AC-353..AC-354, AC-359..AC-368, AC-370..AC-385, AC-387..AC-408, AC-410..AC-417, AC-441, AC-463..AC-473 |
| snapshot_reporting | base | REQ-00-001..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-020, REQ-00-022..REQ-00-027, REQ-00-029..REQ-00-042, REQ-00-044..REQ-00-047, REQ-00-049..REQ-00-054, REQ-00-057, REQ-00-062..REQ-00-063, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-398, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-471, REQ-01-476..REQ-01-479, REQ-01-487..REQ-01-509, REQ-01-516..REQ-01-536, REQ-01-542..REQ-01-548, REQ-01-554..REQ-01-563, REQ-01-565..REQ-01-578, REQ-01-580..REQ-01-584, REQ-01-608, REQ-01-621..REQ-01-626, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-233, REQ-02-238..REQ-02-246, REQ-02-248, REQ-02-250..REQ-02-253, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-282, REQ-03-290, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-039, REQ-04-044..REQ-04-061, REQ-04-065..REQ-04-081, REQ-04-083..REQ-04-092, REQ-04-105..REQ-04-108, REQ-04-126 | AC-001..AC-026, AC-030..AC-032, AC-037..AC-062, AC-068..AC-091, AC-097..AC-163, AC-170..AC-231, AC-233, AC-238..AC-261, AC-266..AC-269, AC-277..AC-287, AC-294..AC-307, AC-311..AC-322, AC-329..AC-331, AC-333..AC-347, AC-353..AC-354, AC-359..AC-368, AC-370..AC-385, AC-387..AC-392, AC-394..AC-408, AC-410..AC-417, AC-441, AC-469..AC-473 |
| reference_pack | base | REQ-00-001..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-020, REQ-00-022..REQ-00-027, REQ-00-029..REQ-00-047, REQ-00-049..REQ-00-054, REQ-00-057, REQ-00-059, REQ-00-062..REQ-00-063, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-399..REQ-01-424, REQ-01-451..REQ-01-471, REQ-01-480..REQ-01-482, REQ-01-487..REQ-01-509, REQ-01-516..REQ-01-536, REQ-01-542..REQ-01-563, REQ-01-565..REQ-01-578, REQ-01-580..REQ-01-584, REQ-01-608, REQ-01-610, REQ-01-621..REQ-01-626, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-233, REQ-02-238..REQ-02-248, REQ-02-250..REQ-02-253, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-282, REQ-03-290, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-043, REQ-04-048..REQ-04-061, REQ-04-065..REQ-04-081, REQ-04-083..REQ-04-092, REQ-04-105..REQ-04-108, REQ-04-126 | AC-001..AC-026, AC-033..AC-035, AC-037..AC-055, AC-068..AC-070, AC-072..AC-090, AC-092..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-231, AC-234, AC-238..AC-261, AC-270..AC-272, AC-277..AC-287, AC-294..AC-304, AC-308..AC-322, AC-326, AC-329..AC-331, AC-334..AC-347, AC-353..AC-354, AC-359..AC-385, AC-387..AC-392, AC-394..AC-408, AC-410..AC-417, AC-427, AC-441, AC-443, AC-469..AC-473 |
| incident_portability | base | REQ-00-001..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-020, REQ-00-022..REQ-00-027, REQ-00-029..REQ-00-042, REQ-00-044..REQ-00-047, REQ-00-049..REQ-00-054, REQ-00-057, REQ-00-058, REQ-00-062..REQ-00-063, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-471, REQ-01-483..REQ-01-509, REQ-01-516..REQ-01-536, REQ-01-542..REQ-01-578, REQ-01-580..REQ-01-584, REQ-01-608, REQ-01-609, REQ-01-621..REQ-01-626, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-233, REQ-02-238..REQ-02-246, REQ-02-248..REQ-02-249, REQ-02-250..REQ-02-253, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-282, REQ-03-290, REQ-04-001..REQ-04-017, REQ-04-021..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-044..REQ-04-061, REQ-04-065..REQ-04-081, REQ-04-083..REQ-04-092, REQ-04-105..REQ-04-108, REQ-04-126 | AC-001..AC-026, AC-037..AC-055, AC-068..AC-070, AC-072..AC-090, AC-097..AC-103, AC-107..AC-112, AC-116..AC-231, AC-236, AC-238..AC-261, AC-273..AC-287, AC-294..AC-304, AC-311..AC-322, AC-327..AC-332, AC-334..AC-347, AC-353..AC-354, AC-359..AC-368, AC-370..AC-392, AC-394..AC-417, AC-440..AC-442, AC-469..AC-473 |
| enterprise_authentication | base | REQ-00-001..REQ-00-008, REQ-00-013..REQ-00-017, REQ-00-019..REQ-00-027, REQ-00-029..REQ-00-042, REQ-00-044..REQ-00-046, REQ-00-048..REQ-00-055, REQ-00-057, REQ-00-062..REQ-00-063, REQ-01-001..REQ-01-009, REQ-01-015..REQ-01-368, REQ-01-423..REQ-01-424, REQ-01-451..REQ-01-465, REQ-01-487..REQ-01-548, REQ-01-554..REQ-01-563, REQ-01-565..REQ-01-578, REQ-01-580..REQ-01-584, REQ-01-608, REQ-01-621..REQ-01-626, REQ-02-001..REQ-02-044, REQ-02-054..REQ-02-138, REQ-02-147..REQ-02-210, REQ-02-212..REQ-02-246, REQ-02-248, REQ-02-250..REQ-02-253, REQ-02-256, REQ-03-001..REQ-03-152, REQ-03-205..REQ-03-282, REQ-03-290, REQ-04-001..REQ-04-030, REQ-04-036..REQ-04-039, REQ-04-048..REQ-04-061, REQ-04-065..REQ-04-081, REQ-04-083..REQ-04-096, REQ-04-105..REQ-04-108, REQ-04-115..REQ-04-122, REQ-04-126 | AC-001..AC-026, AC-036..AC-055, AC-068..AC-070, AC-072..AC-090, AC-097..AC-103, AC-107..AC-112, AC-116..AC-163, AC-170..AC-231, AC-235, AC-238..AC-261, AC-277..AC-304, AC-311..AC-322, AC-329..AC-331, AC-334..AC-354, AC-359..AC-368, AC-370..AC-385, AC-387..AC-392, AC-394..AC-408, AC-410..AC-417, AC-433..AC-436, AC-441, AC-469..AC-473 |
| claim_publication | relevant implementation claim for each published timed or fixture-sensitive criterion | REQ-00-028, REQ-05-001..REQ-05-013 | PC-001..PC-006 |
