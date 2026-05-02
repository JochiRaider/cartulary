---
title: Cartulary Domain Vocabulary and Boundaries
class: domain-language reference
---

# Cartulary Domain Vocabulary and Boundaries

## 1. Status and authority

This document is the first-class repository reference for Cartulary domain vocabulary, domain boundaries, and terminology interpretation. It exists so developers, reviewers, specification authors, and coding agents use the same project-specific meanings for terms such as `party`, `artifact`, `view schema`, `saved view`, `system view`, `entity mention`, `object blob`, and `workbook surface`.

This document does not replace the current implementation-conformance corpus. Core 00 through Core 04 remain the implementation-conformance authority for the current profile, Core 05 remains a normative companion for claim-bearing publication only, and non-normative appendices and guides remain subordinate according to the existing authority model.

The normative words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** in this document govern only repository vocabulary use, domain interpretation, documentation discipline, review discipline, and coding-agent context. **SHOULD** and **SHOULD NOT** indicate strong repository defaults whose exceptions must remain compatible with the owner sections and with this document's vocabulary boundaries. They MUST NOT be read as adding, widening, narrowing, or replacing runtime behavior unless the same behavior is owned by the applicable Core 00 through Core 04 section.

If this document and a primary owner section differ, the owner section governs. The difference is documentation drift and MUST be repaired by updating `docs/domain.md`, the owner section, or both through the repository's ordinary specification-change process. If two owner sections appear to conflict, that conflict is a corpus defect and MUST NOT be resolved by this document alone.

Creation of this document identifies no authority-model conflict requiring a behavior or conformance-scope change to Core 00. The vocabulary resolutions in §5 are least-disruptive interpretations of terms already present in the current document set.

## 2. Purpose

`docs/domain.md` MUST provide a compact, rigorous domain reference that makes Cartulary-specific language operationally usable during implementation and review.

The document MUST serve these functions:

1. Define canonical domain terms and their forbidden substitutions.
2. Map domain terms to their primary owner sections without copying full owner contracts.
3. Distinguish domain concepts from implementation modules, physical tables, component names, route helpers, and external-system concerns.
4. Preserve the current authority model by treating owner sections as the behavioral source of truth.
5. Give coding agents enough stable context to avoid semantic aliasing, storage-model leakage, visible-label inference, and forms-first or module-first implementation mistakes.
6. Provide binary acceptance criteria for evaluating whether the document is complete enough to use during development.

`docs/domain.md` MUST NOT be used as an API reference, schema reference, implementation guide, operating handbook, UI design guide, route inventory, or test plan. It MAY point to those artifacts when they are the correct owner for a term or behavior.

## 3. Domain thesis

Cartulary is a workbook-native incident workspace. Analysts act on visible rows, cells, chips, counts, filters, groups, previews, saved views, and system views, while authoritative source state remains typed, relational, versioned, attributed, and auditable underneath. The workbook surface preserves spreadsheet speed and direct manipulation; the source model rejects spreadsheet identity, overwrite, evidence, history, and relationship semantics where those would undermine collaboration or recoverability.

For domain interpretation, the following statement is controlling:

> Cartulary preserves the spreadsheet mental model at the view layer, not at the storage layer.

A proposed term, feature name, route helper, UI label, or implementation module that contradicts this thesis MUST be treated as ambiguous until mapped to the correct owner section.

## 4. Document relationship map

The following map preserves the existing security, implementation-support, and design-direction boundaries rather than creating a new document hierarchy.

| Artifact               | Owns                                                                                                                                                                          | `docs/domain.md` role                                                                                              |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Core 00                | Document status, precedence, profile model, contract-owner matrix, conformance separation.                                                                                    | MUST follow its authority order and MUST NOT create new runtime conformance requirements.                          |
| Core 01                | Architecture, route families, public interfaces, view schemas, projections, jobs, portability, extension routes, snapshot and reference-pack route families.                  | MAY summarize domain-facing meanings of stable identifiers, workbook surfaces, projections, and route families.    |
| Core 02                | Record model, mention and entity semantics, party model, task requests, decisions, artifacts, indicators, assessments, relationships, history substrate, closed vocabularies. | MUST use Core 02 as the primary owner for entity and record vocabulary.                                            |
| Core 03                | Workbook interaction, built-in tabs, system views, saved views, startup surface selection, collaboration, same-field conflicts, workflows, grouping, coordination behavior.   | MUST use Core 03 as the primary owner for interaction-domain vocabulary.                                           |
| Core 04                | Authentication, authorization, deployment, trust boundaries, runtime roots, conformance criteria.                                                                             | MUST use Core 04 as the primary owner for users, sessions, deployment administration, and security-boundary terms. |
| Core 05                | Claim-bearing publication and benchmark reproducibility.                                                                                                                      | MUST NOT treat Core 05 as Base Profile runtime behavior.                                                           |
| Appendices A through H | Rationale, illustrations, non-normative operating guidance, source preservation, backlog, traceability.                                                                       | MAY use as evidence or orientation only when not in conflict with the core.                                        |
| UI/UX design guide     | Derived design-direction specification.                                                                                                                                       | MAY use for user-facing mental-model language only when behavior remains owned by the core.                        |
| Development guide      | Repo-local implementation baseline and developer workflow.                                                                                                                    | MAY use for package/module mappings, but implementation modules are not domain definitions.                        |
| README                 | Public orientation and onboarding.                                                                                                                                            | SHOULD point readers to `docs/domain.md` for vocabulary.                                                           |
| Code comments          | Local implementation rationale.                                                                                                                                               | MUST use canonical terms from `docs/domain.md` and MUST NOT redefine domain concepts locally.                      |
| `AGENTS.md`            | Coding-agent and contributor procedure when present.                                                                                                                          | SHOULD instruct agents to consult `docs/domain.md` before domain-facing changes.                                   |

## 5. Resolved terminology decisions

The following decisions resolve overloaded or easily-confused language for current-profile repository use. These decisions do not change owner-section behavior.

| Issue                                       | Canonical interpretation                                                                                                                                                      | Forbidden interpretation                                                                                          | Primary owner                       |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| `artifact` versus Notes                     | `artifact` is the structured text object family. Notes is one workbook surface backed by `artifact_type='note'`.                                                              | Treating `artifact` as a synonym for the Notes tab or as binary evidence.                                         | Core 02 §2, §10.4.4A; Core 01 §7.4  |
| `party` versus user                         | `party` is an incident-scoped coordination identity. A user is a deployment-local login and attribution identity.                                                             | Treating `party_id`, `user_id`, email text, auth subject, or incident membership as interchangeable.              | Core 02 §19; Core 04 §2             |
| `identity` versus party                     | `identity` is a host/account/persona investigation entity. `party` is a coordination stakeholder identity.                                                                    | Using identities to model requesters, audiences, attendees, or collectors when the domain needs party references. | Core 02 §2, §19                     |
| `system view` versus `system` saved view    | A system view is a contract-backed workbook surface identified by `view_schema_id`. A `scope='system'` saved view is an implementation-owned saved-view configuration object. | Treating a `system` saved view as the required system view itself.                                                | Core 01 §7.4; Core 03 §2.3          |
| `view_schema_id` versus visible label       | `view_schema_id` is the canonical workbook-surface identity. Visible titles and labels are display hints.                                                                     | Deriving behavior from tab labels, column labels, visible row order, projection names, or storage names.          | Core 01 §3.3.1, §7.4                |
| Projection versus source state              | A projection is a derived read model for workbook use. Source state is the authoritative typed record, link, evidence, history, or admin state.                               | Mutating projections as source of truth or treating projection corruption as source-state corruption.             | Core 01 §8; Core 02 §1, §15         |
| Entity mention versus stub entity           | An entity mention is a source-bound observation. A stub entity is a real host or identity record with its own `record_id`.                                                    | Treating unresolved mentions as weak entities or auto-creating stubs from every mention.                          | Core 02 §6                          |
| Indicator observation versus indicator      | An indicator observation is source-bound. A canonical indicator is a first-class incident-scoped record.                                                                      | Treating every observed indicator-like string as an automatically created canonical indicator.                    | Core 02 §6, §10                     |
| Evidence record versus object blob          | Evidence record is the user-facing evidence envelope. Object blob is binary-object metadata and upload state.                                                                 | Treating object blobs as workbook evidence rows or storing raw evidence inside timeline cells.                    | Core 02 §13; Core 03 §8             |
| Coordination surface versus workflow engine | Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson are workbook-native coordination surfaces.                                                   | Creating a generalized approval/workflow engine for ordinary row edits.                                           | Core 02 §10.4; Core 03 §16.4        |
| Import unit versus worksheet/table          | `import_unit` is the contract object for one candidate ingestable unit. Worksheet ranges and tables are locator kinds.                                                        | Using “selected sheet import” or “Excel table import” as alternate public contract nouns.                         | Core 01 §2.1; development guide §11 |
| Reference pack versus incident data         | Reference packs are versioned optional vocabularies, frameworks, or enrichment datasets outside incident source records.                                                      | Treating reference-pack activation as incident record mutation or blocking core capture on pack availability.     | Core 01 §11; Core 02 §17            |
| Snapshot/report versus live workbook        | Snapshot/report artifacts are immutable export or publication inputs under extension rules. Live workbook state is the operational incident workspace.                        | Applying recipient-specific export redaction by hiding live workbook rows from incident members.                  | Core 01 §10; Core 04 §2.1, §4.2     |

## 6. Domain versus implementation detail

A term belongs in `docs/domain.md` when misunderstanding it would cause a developer, reviewer, or coding agent to build the wrong behavior, address the wrong contract, mutate the wrong source of truth, or write a misleading specification.

A term usually belongs outside `docs/domain.md` when it is only one of the following:

- package, module, or directory name;
- SQL table, generated column, trigger, index, or migration filename;
- React component, hook, CSS class, or grid vendor coordinate;
- Go type, helper function, adapter method, or internal service interface;
- object-store key realization;
- test harness implementation detail;
- deployment-specific secret, path, or operator command.

An implementation detail MAY appear in `docs/domain.md` only as an implementation-facing mapping that prevents ambiguity. Such a mapping MUST NOT become the definition of the domain term.

| Domain term      | May mention implementation mapping                        | Boundary                                                                                                    |
| ---------------- | --------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| Workbook surface | `view_schema_id`, `sheet_ref`, query route, grid adapter. | The domain identity is not the React component, route handler, or visible tab label.                        |
| Party            | `record_type='party'`, `party_id`, Parties view.          | The domain identity is not a deployment-local user, email string, auth subject, or contact-directory entry. |
| Notes            | `cartulary.view.notes.v1`, `artifact_type='note'`.        | Notes is one artifact-backed surface, not the whole artifact family.                                        |
| Projection       | Projection table or equivalent denormalized read model.   | Projection state is disposable and derived, not authoritative source state.                                 |
| Object blob      | Object storage metadata and upload slot.                  | The object-store key is implementation realization, not public evidence identity.                           |
| Import unit      | `import_unit`, locator kind, mapping fingerprint.         | XLSX worksheet, used range, table, and named range are locator examples, not cross-module semantics.        |

## 7. Core distinctions

Each distinction below is review-critical for vocabulary and documentation discipline. A change that blurs one of these distinctions MUST cite the owner section that allows the change.

| Distinction                                      | Required interpretation                                                                                                                                | Common failure                                                                                 |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------- |
| Workbook surface versus source state             | A surface is what analysts query and edit. Source state is the record, link, mention, evidence, or admin model underneath.                             | Updating projection rows directly or storing relationship meaning in visible columns only.     |
| Built-in tab versus system view                  | Built-in tabs are five primary required sheets. System views are additional required workbook-native surfaces reached through the same workbook model. | Making coordination or indicator surfaces separate application modules.                        |
| System view versus saved view                    | A system view is a required `view_schema_id`. A saved view is an incident-bound configuration over one `view_schema_id`.                               | Replacing required surfaces with saved-view presets.                                           |
| Visible label versus stable identifier           | Public behavior binds to identifiers such as `view_schema_id`, `field_key`, `record_id`, `row_version`, `client_txn_id`, and `party_id`.               | Inferring write behavior from labels, row order, or SQL names.                                 |
| User versus party versus identity                | User authenticates and receives attribution. Party coordinates stakeholder references. Identity is an incident entity such as an account or persona.   | Using `user_id` as requester, using party as login, or using identity as stakeholder audience. |
| Mention versus entity                            | A mention is an observed string in a source record and field. An entity is a host or identity record.                                                  | Auto-creating entities from every host-like or account-like string.                            |
| Indicator observation versus canonical indicator | Observation is source-bound occurrence. Canonical indicator is first-class linkable record.                                                            | Deduping observations into one record without preserving occurrences.                          |
| Evidence record versus object blob               | Evidence record models request, receipt, custody, availability, and workbook linkage. Object blob models binary upload and storage metadata.           | Treating raw blob existence as sufficient evidence availability.                               |
| Artifact versus binary artifact                  | In Cartulary, `artifact` means structured analyst text material. Binary evidence uses evidence records and object blobs.                               | Calling uploaded files “artifacts” without qualifying them as evidence blobs.                  |
| Task/decision versus timeline field              | Task Requests and Decisions are coordination records. Timeline rows capture observations and chronology.                                               | Adding mandatory task or approval fields to timeline row creation.                             |
| Reference pack versus overlay surface            | Reference packs provide optional vocabularies and enrichment. Pack-dependent framework overlays are not base workbook surfaces.                        | Exposing ATT&CK, D3FEND, or VERIS as base `view_schema` resources.                             |
| Snapshot/report versus live workspace            | Snapshot/report is an export or release construct. Live workspace is incident collaboration.                                                           | Applying external-release redaction to live workbook visibility.                               |

## 8. Canonical identifier vocabulary

| Identifier          | Domain target                                                                                       | Stable use                                                                                | Not this                                                                   | Owner            |
| ------------------- | --------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- | ---------------- |
| `incident_id`       | Incident workspace.                                                                                 | Scope for incident records, views, jobs, and authorization checks.                        | Deployment, tenant, or object-store bucket identity.                       | Core 01, Core 02 |
| `incident_key`      | Deployment-unique incident key.                                                                     | Human-meaningful incident identity normalized for uniqueness.                             | Public row identity or authorization token.                                | Core 02, Core 01 |
| `record_id`         | User-visible first-class incident record envelope.                                                  | Row identity, record mutation, links, tags, history, rollback.                            | User identity, party reference, risk-ref child identity, or blob identity. | Core 02          |
| `row_version`       | Current optimistic-concurrency version for a record row.                                            | Server-emitted version anchor for current row state.                                      | Global revision number or history-entry selector.                          | Core 03          |
| `base_row_version`  | Client's version anchor for an attempted record write.                                              | Conflict detection for record-scoped writes.                                              | Current version after successful commit.                                   | Core 03, Core 01 |
| `field_key`         | Stable view-field identity.                                                                         | Write target, conflict key, sort/filter/group capability key.                             | Column label, SQL column, visible header, or translated label.             | Core 01          |
| `view_schema_id`    | Stable public workbook-surface identity.                                                            | Required built-in sheet, required system view, or standardized optional surface identity. | Visible tab label, saved view ID, or projection name.                      | Core 01          |
| `sheet_ref`         | Stable reference to a workbook surface or saved view.                                               | Startup/default surface pointers and presence state.                                      | Visible shell location.                                                    | Core 01, Core 03 |
| `saved_view_id`     | Saved-view configuration object.                                                                    | Incident-bound saved view identity.                                                       | Required system view identity.                                             | Core 03          |
| `party_id`          | Same-incident party record.                                                                         | Stable coordination-party reference.                                                      | Email text, user ID, auth-provider subject, or incident membership.        | Core 02          |
| `entity_mention_id` | Source-bound textual entity mention.                                                                | Explicit resolve or dismiss target.                                                       | Host or identity `record_id`.                                              | Core 02, Core 01 |
| `object_blob_id`    | Binary upload slot or object metadata.                                                              | Blob create, upload, and attach flow.                                                     | Evidence record identity.                                                  | Core 01, Core 02 |
| `client_txn_id`     | Client-supplied idempotency key for one multi-change user action where the owner route requires it. | Replay-safe mutation boundaries.                                                          | Record ID, operation type, or global transaction ID.                       | Core 01          |
| `history_entry_ref` | Public row-history selector for eligible rollback targets.                                          | Record-history and rollback selection.                                                    | Storage primary key or mutation-entry ID.                                  | Core 02, Core 01 |
| `risk_ref_id`       | Child identity scoped to one `handoff` artifact.                                                    | Public item reference for `handoff.open_risk_refs[]`.                                     | First-class risk record or generic `record_id`.                            | Core 02, Core 01 |
| `user_id`           | Deployment-local user and attribution identity.                                                     | Authentication, session, ownership, actor attribution.                                    | Party, identity, or incident membership itself.                            | Core 04, Core 01 |

## 9. Workbook-surface registry

The following table is a domain-facing copy of the current-profile standardized workbook-surface identity map. Core 01 §7.4 owns the authoritative registry, exhaustive field membership, field defaults, write targets, sort/filter/group capability registries, and public discovery behavior.

| Surface                | `view_schema_id`                          | Surface kind     | Source record types | Canonical discriminator or filter                                                                              | Surface status                         |
| ---------------------- | ----------------------------------------- | ---------------- | ------------------- | -------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| Timeline               | `cartulary.view.timeline.v1`              | `built_in_sheet` | `timeline_event`    | `record_type='timeline_event'`                                                                                 | required built-in sheet                |
| Hosts                  | `cartulary.view.hosts.v1`                 | `built_in_sheet` | `host`              | `record_type='host'`                                                                                           | required built-in sheet                |
| Identities             | `cartulary.view.identities.v1`            | `built_in_sheet` | `identity`          | `record_type='identity'`                                                                                       | required built-in sheet                |
| Evidence               | `cartulary.view.evidence.v1`              | `built_in_sheet` | `evidence`          | `record_type='evidence'`                                                                                       | required built-in sheet                |
| Notes                  | `cartulary.view.notes.v1`                 | `built_in_sheet` | `artifact`          | `artifact_type='note'`                                                                                         | required built-in sheet                |
| Indicators             | `cartulary.view.indicators.v1`            | `system_view`    | `indicator`         | `record_type='indicator'`                                                                                      | required system view                   |
| Compromise Assessments | `cartulary.view.assessments.v1`           | `system_view`    | `assessment`        | `record_type='assessment'`                                                                                     | required system view                   |
| Task Requests          | `cartulary.view.task_requests.v1`         | `system_view`    | `task_request`      | `record_type='task_request'`                                                                                   | required system view                   |
| Decisions              | `cartulary.view.decisions.v1`             | `system_view`    | `decision`          | `record_type='decision'`                                                                                       | required system view                   |
| Parties                | `cartulary.view.parties.v1`               | `system_view`    | `party`             | `record_type='party'`                                                                                          | required system view                   |
| Communications Log     | `cartulary.view.comm_log.v1`              | `system_view`    | `artifact`          | `artifact_type='comm_log'`                                                                                     | required system view                   |
| Handoff                | `cartulary.view.handoff.v1`               | `system_view`    | `artifact`          | `artifact_type='handoff'`                                                                                      | required system view                   |
| Status Review          | `cartulary.view.status_review.v1`         | `system_view`    | `artifact`          | `artifact_type='status_review'`                                                                                | required system view                   |
| Lesson                 | `cartulary.view.lesson.v1`                | `system_view`    | `artifact`          | `artifact_type='lesson'`                                                                                       | required system view                   |
| Findings               | `cartulary.view.findings.v1`              | `system_view`    | `artifact`          | `artifact_type='finding'`; subtype dimension `finding.kind`                                                    | standardized optional workbook surface |
| Investigative Queries  | `cartulary.view.investigative_queries.v1` | `system_view`    | `artifact`          | implementation-declared structured investigative-query subtype; no current-profile fixed durable discriminator | standardized optional workbook surface |
| Forensic Keywords      | `cartulary.view.forensic_keywords.v1`     | `system_view`    | `artifact`          | implementation-declared structured forensic-keyword subtype; no current-profile fixed durable discriminator    | standardized optional workbook surface |

### 9.1 Surface interpretation rules

- Required built-in tabs MUST be treated as primary workbook surfaces, not as storage tables.
- Required system views MUST remain workbook-native surfaces, not separate application modules.
- Standardized optional workbook surfaces MAY be exposed only when implemented under their owner contracts.
- A saved view over any `view_schema_id` is additive and non-canonical.
- Pack-dependent framework overlays such as ATT&CK, D3FEND, and VERIS MUST NOT be described as base-profile workbook surfaces.
- Display labels MAY change without changing `view_schema_id` when field semantics do not change.

## 10. Bounded contexts

Bounded contexts describe domain responsibility. They MUST NOT be treated as mandatory package names, deployable names, database schemas, or UI navigation labels.

| Bounded context                   | Owns domain language for                                                                                                            | Must not own                                                                                                | Primary owner                      |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Incident Workspace                | Incident, incident key, incident metadata, membership, startup workbook surface, incident-scoped preferences.                       | Password/MFA semantics, deployment-wide users, object storage.                                              | Core 01, Core 02, Core 03, Core 04 |
| Workbook Interaction              | Built-in tabs, system views, saved views, grid action, paste, sort, filter, group, inspector, presence, save/conflict state.        | Authoritative storage schema or physical projection topology.                                               | Core 03, Core 01                   |
| Capture and Timeline              | Timeline events, rough capture, capture state, timeline grouping, timeline source text, unresolved tokens.                          | Canonical host/identity lifecycle except through mention and resolution workflows.                          | Core 02, Core 03                   |
| Entities and Observations         | Hosts, identities, aliases, entity mentions, indicators, indicator observations, assessments, exact-match reuse, explicit merge.    | Parties, deployment users, generalized contact management.                                                  | Core 02, Core 03                   |
| Evidence                          | Evidence records, object blobs, upload slot, attach flow, preview/download handle, custody and availability lifecycle.              | Long-form report release state or raw telemetry storage.                                                    | Core 01, Core 02, Core 03, Core 04 |
| Coordination                      | Parties, task requests, decisions, communications logs, handoffs, status reviews, lessons, owner fields, follow-through references. | Generalized workflow engine or mandatory timeline approval path.                                            | Core 02, Core 03, Core 04          |
| Links and Tags                    | Typed record relationships, record tags, relationship cells, collection-review fields.                                              | Raw binary evidence storage or user account binding.                                                        | Core 02, Core 01                   |
| Revisions and Audit               | Change sets, mutation entries, record revisions, row history, rollback, merge history.                                              | Live presence transport or deployment-local credential secrets.                                             | Core 02, Core 01, Core 04          |
| Projections and Search            | Workbook projections, view-query contract, filter/sort/group behavior, cursor semantics.                                            | Authoritative source decisions or persisted incident facts except where owner sections define source state. | Core 01, Core 03                   |
| Reference Data                    | Reference packs, type registries, activation, attestation, optional overlays.                                                       | Incident record lifecycle or live capture eligibility.                                                      | Core 01, Core 02, Core 04          |
| Reporting and Snapshots           | Immutable snapshots, export models, release records, redaction profiles, rendered outputs.                                          | Live workbook write path or incident-member visibility.                                                     | Core 01, Core 04                   |
| Authentication and Administration | Local users, sessions, MFA, deployment admin, credential lifecycle, enterprise-auth bindings, incident roles.                       | Incident-scoped party identity or workbook row mutation.                                                    | Core 01, Core 04                   |
| Imports and Tabular Ingest        | Import sessions, import units, source adapters, locator kinds, mapping fingerprints, warning codes, provenance.                     | Runtime workbook semantics outside the stable tabular-ingest contract.                                      | Core 01, Core 03, Core 04          |
| Backup and Restore                | Backup sets, restore verification, runtime roots, operator-facing recovery.                                                         | Workbook-surface route families or incident-scoped workflow.                                                | Core 01, Core 04                   |

## 11. Core domain glossary

The `Owner` column points to the primary section that controls behavior. This table defines vocabulary, not exhaustive field behavior.

| Term                         | Definition                                                                                                                                                   | Not this                                                                                             | Canonical identifiers or tokens                                   | Owner                               |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------- |
| Incident                     | Workspace boundary for incident-scoped records, collaboration, authorization, and workbook state.                                                            | Tenant, deployment, case-file path, or report artifact.                                              | `incident_id`, `incident_key`                                     | Core 01 §3.3.5.3; Core 02 §2        |
| Record envelope              | Common identity, version, attribution, and delete-state wrapper for user-visible first-class incident records.                                               | Administrative user, session, membership, bootstrap marker, blob metadata alone.                     | `record_id`, `record_type`, `row_version`                         | Core 02 §3                          |
| Timeline event               | Primary rough-capture unit for incident chronology, observations, and later structuring.                                                                     | Task, decision, evidence blob, or final narrative paragraph.                                         | `record_type='timeline_event'`, `cartulary.view.timeline.v1`      | Core 02 §2; Core 03 §6-§7           |
| Host                         | Canonical or stub device or host record used for scoping and entity relationships.                                                                           | Raw host-like text in another row.                                                                   | `record_type='host'`, `cartulary.view.hosts.v1`                   | Core 02 §2, §6, §8                  |
| Identity                     | Canonical or stub account or persona record used for scoping and entity relationships.                                                                       | Deployment-local user or stakeholder party.                                                          | `record_type='identity'`, `cartulary.view.identities.v1`          | Core 02 §2, §6, §8                  |
| Party                        | Incident-scoped coordination identity for a person, team, organization, distribution list, requester, collector, source, audience, attendee, or stakeholder. | Deployment-local user, auth identity, email text, identity record, incident membership, CRM contact. | `record_type='party'`, `party_id`, `party_kind`                   | Core 02 §19                         |
| User                         | Deployment-local login and attribution identity.                                                                                                             | Party, host/identity entity, incident membership, or stakeholder audience.                           | `user_id`, session resource, `is_deployment_admin`                | Core 01 §3.3.2; Core 04 §1-§2       |
| Incident membership          | Authorization relationship between a user and an incident.                                                                                                   | Party reference or user account itself.                                                              | role tokens `viewer`, `editor`, `reviewer`, `admin`               | Core 04 §2                          |
| Deployment admin             | Deployment-scoped capability for account and deployment-local administration.                                                                                | Incident admin role or incident data access.                                                         | `deployment_admin`, `is_deployment_admin`                         | Core 04 §2                          |
| Artifact                     | Structured source-preserving analyst material, including notes and selected coordination or finding variants.                                                | Binary evidence, object blob, or the Notes surface alone.                                            | `record_type='artifact'`, `artifact_type`                         | Core 02 §2, §10.4.4A                |
| Note                         | Artifact-backed text material exposed through the required Notes built-in sheet.                                                                             | Every artifact, binary evidence, or external findings document.                                      | `artifact_type='note'`, `cartulary.view.notes.v1`                 | Core 01 §7.3-§7.4; Core 02 §10.4.4A |
| Task request                 | First-class owned unit of work or request with lifecycle state, owner, priority, and queue semantics.                                                        | Tag, note, generic ticket integration, or mandatory timeline field.                                  | `record_type='task_request'`, `task_request.status`               | Core 02 §10.4.1                     |
| Decision                     | First-class rationale-bearing incident coordination choice with status and support references.                                                               | Generalized approval for row edits or release approval record.                                       | `record_type='decision'`, `decision.status`                       | Core 02 §10.4.2                     |
| Communications Log           | Artifact-backed durable record of stakeholder-impacting communication or meeting context.                                                                    | Chat transcript store or generalized notification system.                                            | `artifact_type='comm_log'`, `comm_id`                             | Core 02 §10.4.4                     |
| Handoff                      | Artifact-backed continuity record for phase, shift, or ownership handoff.                                                                                    | Ordinary timeline edit or generic task comment.                                                      | `artifact_type='handoff'`, `handoff_id`, `handoff.ack_state`      | Core 02 §10.4.4                     |
| Status Review                | Artifact-backed coordination checkpoint for blockers, pending evidence, decisions, risks, and next reporting.                                                | Dashboard state or mandatory row-edit ritual.                                                        | `artifact_type='status_review'`, `status_review_id`               | Core 02 §10.4.4                     |
| Lesson                       | Artifact-backed retrospective or improvement record with follow-up linkage.                                                                                  | Closed task or final report section alone.                                                           | `artifact_type='lesson'`, `lesson_id`, `lesson.closure_state`     | Core 02 §10.4.4                     |
| Finding                      | Optional artifact-backed structured finding or hypothesis surface when implemented.                                                                          | First-class hypothesis record in the current profile.                                                | `artifact_type='finding'`, `finding.kind`                         | Core 02 §10.4.4A-§10.4.6            |
| Entity mention               | Source-bound raw textual reference to a host or identity captured before canonical resolution.                                                               | Stub entity, alias, host row, identity row.                                                          | `entity_mention_id`, `resolution_status`                          | Core 02 §6                          |
| Stub entity                  | Host or identity record created when a real entity is needed but canonical details remain incomplete.                                                        | Unresolved mention.                                                                                  | `record_type='host'` or `record_type='identity'`, `entity_origin` | Core 02 §6, §8                      |
| Entity alias                 | Alternate name or identifier associated with a canonical or stub host or identity.                                                                           | Source-bound mention occurrence.                                                                     | alias rows, exact-match reuse inputs                              | Core 02 §7-§8                       |
| Indicator                    | Canonical incident-scoped linkable value or pattern.                                                                                                         | Raw IOC-like string occurrence.                                                                      | `record_type='indicator'`, `indicator.value_kind`                 | Core 02 §10                         |
| Indicator observation        | Source-bound occurrence of an indicator value or pattern inside another record field.                                                                        | Canonical indicator or lifecycle interval.                                                           | observation source record and field, `origin_kind`                | Core 02 §6, §10                     |
| Indicator lifecycle interval | Append-only state window attached to a canonical indicator.                                                                                                  | Observation time or overwrite of indicator state.                                                    | lifecycle interval row                                            | Core 02 §10                         |
| Compromise assessment        | Incident-scoped assessment record about a host or identity.                                                                                                  | Field on host or identity, or overwritten current verdict.                                           | `record_type='assessment'`, `assessment_state`                    | Core 02 §10.3                       |
| Evidence record              | User-facing evidence envelope that can model requested, pending, received, available, quarantined, or released evidence.                                     | Raw blob, file path, screenshot pasted into notes.                                                   | `record_type='evidence'`, `evidence_records.lifecycle_state`      | Core 02 §13                         |
| Object blob                  | Storage metadata and upload state for binary content.                                                                                                        | Evidence row, workbook attachment count, or object-store key contract.                               | `object_blob_id`, `object_blobs.upload_state`                     | Core 02 §13; Core 03 §8             |
| Record link                  | Typed relationship between two records.                                                                                                                      | JSON list, text mention, or visible chip alone.                                                      | `record_links.link_type`, `record_links.provenance`               | Core 02 §12                         |
| Tag                          | Lightweight incident-scoped label.                                                                                                                           | Owner, task, lifecycle state, or relationship.                                                       | `tag_id`, `record_tags`                                           | Core 02 §12                         |
| View schema                  | Contract for a built-in sheet, required system view, or standardized optional workbook surface.                                                              | Physical table schema or saved view.                                                                 | `view_schema_id`                                                  | Core 01 §7.4                        |
| Workbook surface             | A visible or addressable sheet-like working surface backed by a `view_schema_id` or distinct saved view over one schema.                                     | Module, route family, projection table, or visible label.                                            | `sheet_ref`                                                       | Core 01 §7.4; Core 03 §2            |
| Built-in tab                 | Required primary sheet-like workbook surface in the base profile.                                                                                            | Every required surface.                                                                              | Timeline, Hosts, Identities, Evidence, Notes                      | Core 03 §2.1                        |
| System view                  | Required workbook-native surface beyond the five built-in tabs.                                                                                              | `scope='system'` saved view.                                                                         | required system `view_schema_id` values                           | Core 03 §2.2                        |
| Saved view                   | Incident-bound workbook configuration over exactly one immutable `view_schema_id`.                                                                           | Required system view or projection row.                                                              | `saved_view_id`, `scope`                                          | Core 03 §2.3                        |
| Projection                   | Denormalized read model used for workbook query, sorting, filtering, grouping, and row refresh.                                                              | Source of truth or history substrate.                                                                | projection row with `record_id` and `row_version`                 | Core 01 §8; Core 02 §15             |
| Change set                   | Immutable attribution unit for one committed action.                                                                                                         | UI action only or untracked transaction.                                                             | `change_set_id`, actor, timestamp, source                         | Core 02 §15                         |
| Mutation entry               | Reversible target-level delta within a change set.                                                                                                           | Human-facing history summary only.                                                                   | target kind, target identifier, operation kind                    | Core 02 §15                         |
| Record revision              | Row-centric historical state used for review and rollback.                                                                                                   | Projection snapshot or audit log only.                                                               | revision number, history entry ref where eligible                 | Core 02 §15; Core 01 §3.3.4.2       |
| Rollback                     | Scoped historical corrective action that appends history.                                                                                                    | Silent overwrite or delete-and-recreate.                                                             | rollback route and `history_entry_ref` where eligible             | Core 01 §3.3.5; Core 02 §15         |
| Reference pack               | Versioned optional vocabulary, framework, type registry, template, or enrichment dataset.                                                                    | Incident record, live workbook surface, or required capture dependency.                              | `pack_key`, `pack_version`, activation metadata                   | Core 01 §11; Core 02 §17            |
| Import session               | One uploaded source file plus one operator-driven import workflow.                                                                                           | Whole-workbook runtime behavior.                                                                     | `import_session_id`                                               | Core 01 §2.1, §17                   |
| Import unit                  | One candidate ingestable unit discovered from an import source.                                                                                              | Worksheet identity or table identity.                                                                | `import_unit`, `locator_kind`, `mapping_fingerprint`              | Core 01 §2.1                        |
| Snapshot                     | Immutable incident export-model anchor when the Snapshot and Reporting Extension Profile is implemented.                                                     | Live workbook state or saved view.                                                                   | `snapshot_id`, `snapshot_at`                                      | Core 01 §10                         |
| Release                      | Artifact-scoped rendered-output approval and publication record.                                                                                             | General row approval workflow.                                                                       | `release_id`, `release_state`                                     | Core 04 §2.1; Core 01 §10           |

## 12. Entity and relationship model

### 12.1 Record-envelope membership

A user-visible first-class incident record MUST consume one `record_id` through the record-envelope model. The following current-profile record types are first-class incident records:

| Record type      | Domain object                                                                  |
| ---------------- | ------------------------------------------------------------------------------ |
| `timeline_event` | Timeline event                                                                 |
| `host`           | Host                                                                           |
| `identity`       | Identity                                                                       |
| `party`          | Party                                                                          |
| `indicator`      | Canonical indicator                                                            |
| `artifact`       | Notes, coordination artifacts, findings, and other structured analyst material |
| `task_request`   | Task request                                                                   |
| `decision`       | Decision                                                                       |
| `evidence`       | Evidence record                                                                |
| `assessment`     | Compromise assessment                                                          |

Administrative users, auth-provider identities, sessions, incident memberships, local-account credential state, bootstrap-completion state, object blobs, import-unit child state, and handoff risk-ref child rows MUST NOT be called first-class incident records unless a later owner spec changes their model.

### 12.2 Relationship families

| Relationship family                         | Authoritative representation                                              | Domain rule                                                                   |
| ------------------------------------------- | ------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Host or identity observed in another record | `entity_mention` plus optional resolution and typed link                  | The raw mention remains preserved after resolution.                           |
| Indicator observed in another record        | `indicator_observation` plus optional canonical indicator link            | Repeated observations remain distinct.                                        |
| Evidence attached to a record               | Evidence record plus typed record link and optional object blob           | Binary upload success alone is not sufficient to imply evidence availability. |
| Record-to-record association                | `record_links` with exact `link_type` token                               | Relationship semantics are typed, not inferred from text columns.             |
| Tags                                        | `record_tags` association to incident-scoped tag                          | Tags are labels, not lifecycle or ownership state.                            |
| Coordination collections                    | `record_ref`, `party_ref`, or `risk_ref` item families depending on field | Public item kind is field-family-specific and not a free-form string array.   |
| Entity merge                                | Explicit merge action preserving pre- and post-merge graph                | Merges are explicit and auditable.                                            |
| Timeline supersession                       | Typed `supersedes` relation when replacement row is selected              | Supersession is reviewer action, not ordinary edit.                           |

### 12.3 Current-profile invariants

The following invariants MUST be used during review and agent prompting:

1. Mentions and entities are different object types.
2. Repeated mentions remain distinct observations.
3. Source-bound indicator observations and canonical indicators are different object types.
4. Repeated indicator observations remain distinct observations.
5. Exact-match reuse follows deterministic owner-defined precedence.
6. Merges are explicit and auditable.
7. Notes are artifacts.
8. Notes and tags remain insufficient for owned lifecycle work.
9. Owned work uses `task_request`, `decision`, or coordination-artifact objects.
10. Current-profile ownership is a field on coordination objects, not a standalone domain object.
11. Relationship semantics are typed.
12. Projection state is derived.
13. History visibility is retained for extant incident records within the current deployment.
14. Hypotheses remain artifact-backed as findings with `finding.kind='hypothesis'`.
15. Reference-pack manifests and activation state are incident-external structured state.

## 13. Workflow vocabulary

### 13.1 Rough timeline capture

Rough capture is analyst-entered partial or uncertain incident information that is valid before canonical structure is complete. A timeline row MAY be created with nullable time, one non-empty field, unresolved host or account text, indicator-bearing text, unstructured details, or only an attached screenshot. Creation MUST NOT block on canonical host, identity, or indicator records.

Domain defaults:

| Event                                       | Domain interpretation                                                           |
| ------------------------------------------- | ------------------------------------------------------------------------------- |
| New Timeline row                            | Persists with `capture_state='rough'`.                                          |
| First later capture-state-material mutation | Moves `rough` to `enriched`.                                                    |
| Reviewer mark-reviewed action               | Moves `rough` or `enriched` to `reviewed`.                                      |
| Material mutation to reviewed row           | Moves `reviewed` to `enriched`.                                                 |
| Reviewer supersede action                   | Moves eligible row to terminal `superseded`.                                    |
| Link creation or mention resolution         | Creates a derived `linked` milestone but not a stored `capture_state='linked'`. |
| Rollback                                    | History or reviewer outcome, not a `capture_state` token.                       |

### 13.2 Mention resolution

Mention resolution is the explicit process of linking a source-bound textual observation to an existing or newly created host or identity. The raw mention MUST remain preserved. Creating a stub from one mention resolves only that selected mention by default. Bulk resolution of sibling mentions requires a separate explicit action.

Automatic background matching MAY suggest candidates. It MUST NOT create stubs or merge entities unless the owner-defined binding mode and explicit action rules permit it.

### 13.3 Auto-resolution

Auto-resolution is a narrow current-profile eligibility path for interactive mention capture on Timeline relationship cells. It is not general fuzzy matching. If any required condition fails, the system follows the ordinary unresolved path and MAY present non-mutating suggestions.

Auto-resolution vocabulary MUST NOT be used for:

- party references;
- import-time entity creation outside the owner-defined binding mode;
- indicator canonicalization;
- fuzzy or partial alias matching;
- batch merge or dedupe;
- visible-label based matching.

### 13.4 Indicator capture

Indicator capture from Timeline, Notes, Evidence, or other supported source fields preserves the raw source field. It does not require dedicated IOC columns on non-indicator sheets. A canonical indicator is useful when the value or pattern needs linkability, lifecycle intervals, dedupe, filtering, or history beyond the source-bound occurrence.

### 13.5 Evidence request, receipt, and attachment

Evidence work has two linked but distinct state families:

| State family             | Domain object   | Exact tokens                                                                       |
| ------------------------ | --------------- | ---------------------------------------------------------------------------------- |
| Blob upload state        | Object blob     | `pending`, `available`, `failed`, `quarantined`                                    |
| Evidence lifecycle state | Evidence record | `requested`, `pending_receipt`, `received`, `available`, `quarantined`, `released` |

Binary evidence attachment uses a two-step upload and finalization flow. A pending or requested evidence record may exist without a blob. Upload hints such as filename and content type are advisory metadata, not authorization, storage-key, preview-allowlist, or release-posture authority.

### 13.6 Task request workflow

Task Requests model owned work. A task request is the correct domain object when work needs an accountable owner, lifecycle state, due or blocker tracking, queue visibility, handoff durability, or later reconstruction.

| State         | Meaning                                                             |
| ------------- | ------------------------------------------------------------------- |
| `open`        | Task exists and is not currently being worked.                      |
| `in_progress` | Task is actively being worked.                                      |
| `blocked`     | Task cannot proceed until a prerequisite or dependency is resolved. |
| `done`        | Task is complete.                                                   |
| `canceled`    | Task is no longer being pursued.                                    |

A task request MUST NOT be used as a hidden substitute for ordinary timeline capture. A tag or note MUST NOT be used as a substitute for a task when the work requires lifecycle state, ownership, or queueing.

### 13.7 Decision workflow

Decisions model rationale-bearing coordination choices. A decision record is the correct domain object when the incident team needs an inspectable choice, owner, rationale, status, support references, affected records, or supersession history.

| State        | Meaning                                                                     |
| ------------ | --------------------------------------------------------------------------- |
| `proposed`   | Candidate decision under consideration.                                     |
| `approved`   | Accepted as current intended course of action, not necessarily carried out. |
| `rejected`   | Considered and declined.                                                    |
| `executed`   | Carried out.                                                                |
| `superseded` | Overtaken before execution by a later accepted decision.                    |

`approved` in a decision is an incident-coordination state. It MUST NOT be described as a generalized reviewer or administrator approval gate for ordinary row edits.

### 13.8 Coordination artifacts

| Artifact        | Domain workflow role                                                                                                       | Required surface                  |
| --------------- | -------------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| `comm_log`      | Durable communication memory for meetings, notifications, approvals, briefings, and handoffs.                              | `cartulary.view.comm_log.v1`      |
| `handoff`       | Continuity record for outgoing/incoming ownership, current state, open work, open decisions, risks, and next checks.       | `cartulary.view.handoff.v1`       |
| `status_review` | Coordination checkpoint for state summary, blocked tasks, pending evidence, open decisions, risks, and next report timing. | `cartulary.view.status_review.v1` |
| `lesson`        | Retrospective or process-learning record with owner, follow-up tasks, evidence refs, and closure state.                    | `cartulary.view.lesson.v1`        |

Coordination artifacts MUST remain workbook-native and artifact-backed in the current profile. They MUST NOT become required timeline fields or a standalone workflow module.

### 13.9 Saved-view workflow

A saved view captures query and layout configuration over exactly one immutable `view_schema_id`. Scope controls saved-view discoverability and mutability only. It does not change underlying incident-row authorization, evidence visibility, export redaction, or field visibility.

| Scope     | Meaning                                                                                                                                       |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `private` | Visible only to owner and incident admins. Any incident member may create their own private saved view.                                       |
| `shared`  | Visible to all incident members. Owner and incident admins may update or delete it in place.                                                  |
| `system`  | Implementation-owned or admin-seeded saved-view configuration, visible to all incident members, immutable through ordinary saved-view routes. |

Workbook startup selection follows explicit launch reference, user home reference, incident default reference, then `cartulary.view.timeline.v1`.

### 13.10 Same-field conflict workflow

Cartulary uses field-level optimistic concurrency on top of row versioning. Different-field concurrent edits may auto-rebase. Same-field concurrent edits produce a conflict keyed by `field_key` and must remain unresolved until an analyst explicitly chooses a resolution.

A same-field conflict MUST NOT be described as a text-range conflict, row lock, form edit lock, or silent overwrite.

### 13.11 Import workflow

Clipboard paste is part of the base workbook hot path. File-based structured import is extension-profile work isolated behind import sessions, import units, approved mappings, mapping fingerprints, diagnostics, and provenance.

Import vocabulary MUST preserve this boundary:

| Term                  | Correct use                                                       | Incorrect use                                        |
| --------------------- | ----------------------------------------------------------------- | ---------------------------------------------------- |
| `import_session`      | One uploaded source file and one operator-driven workflow.        | A workbook surface or incident.                      |
| `import_unit`         | One candidate ingestable unit discovered from a source.           | A worksheet/table/range as public contract identity. |
| `locator_kind`        | Source locator category such as CSV file or eligible XLSX region. | Runtime workbook semantics outside `imports`.        |
| `mapping_fingerprint` | Deterministic identity for an approved source-to-field mapping.   | Row identity or import-session identity.             |

### 13.12 Snapshot, report, and release workflow

Snapshots, rendered outputs, and release records belong to the Snapshot and Reporting Extension Profile. They are not the live workbook. Recipient-specific withholding belongs at snapshot, render, and release time, not by hiding live workbook content from authenticated incident participants under the base incident-role model.

## 14. Defaults and boundary conditions with domain significance

This table lists defaults and omitted cases that commonly affect terminology. Owner sections remain authoritative for exact request shapes and failure semantics.

| Domain area                                     | Default or boundary condition                                                                                                                                                                                                                                             | Owner                     |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- |
| Incident create                                 | `status` is server-managed as `active`; optional incident fields omitted or explicit `null` expose `null` where the owner contract permits.                                                                                                                               | Core 01 §3.3.5.3          |
| Timeline create                                 | New row persists with `capture_state='rough'`; row creation may use one non-empty value or only an attached screenshot.                                                                                                                                                   | Core 03 §6-§7             |
| Task request inline create                      | Omitted `task.status`, `task.owner_user_id`, and `task.priority` default to `open`, current actor, and `normal`; those defaults do not satisfy the minimum create signal.                                                                                                 | Core 01 §7.4.8            |
| Decision inline create                          | Omitted `decision.status`, `decision.owner_user_id`, and `decision.decided_at` default to `proposed`, current actor, and commit timestamp; those defaults do not satisfy the minimum create signal.                                                                       | Core 01 §7.4.9            |
| Communications Log create                       | Omitted `comm_id` is generated; omitted `timestamp_utc` defaults to commit timestamp; collection fields default to empty; nullable coordination fields default to `null`; defaults do not satisfy the minimum create signal.                                              | Core 01 §19               |
| Handoff create                                  | Omitted `handoff_id` is generated; omitted timestamp defaults to commit timestamp; outgoing owner defaults to current actor; open collections default to empty; `next_checks` and `acknowledged_at` default to `null`; defaults do not satisfy the minimum create signal. | Core 01 §19               |
| Status Review create                            | Omitted ID is generated; timestamp defaults to commit timestamp; review owner defaults to current actor; coordination collections default to empty; risk summary and next report default to `null`; defaults do not satisfy the minimum create signal.                    | Core 01 §19               |
| Lesson create                                   | Omitted ID is generated; timestamp defaults to commit timestamp; owner defaults to current actor; follow-up/evidence collections default to empty; `closure_state` defaults to `open`; defaults do not satisfy the minimum create signal.                                 | Core 01 §19               |
| Finding create when optional surface is exposed | `finding.kind` defaults to `finding`; `finding.state` defaults to `open`; owner defaults to current actor; confidence and close timestamp default to `null`.                                                                                                              | Core 01 §19               |
| Saved-view create                               | Omitted `scope` defaults to `private`; ordinary create rejects `scope='system'`.                                                                                                                                                                                          | Core 03 §2.3              |
| Saved-view query state                          | Omitted or empty `sort` means no user sort override; omitted or empty `filters` means no filters; inactive grouping omits `group_by`, never JSON `null`.                                                                                                                  | Core 02 §4.3; Core 03 §14 |
| Workbook startup                                | Fallback order is explicit launch `sheet_ref`, user home `sheet_ref`, incident default `sheet_ref`, then `cartulary.view.timeline.v1`. Invalid pointers are cleared and fallback continues.                                                                               | Core 03 §2.4              |
| Party text/ref pair                             | Text and `party_id` preserve independent meanings; clearing one does not clear the other unless the action explicitly clears both.                                                                                                                                        | Core 02 §19               |
| Object-blob upload                              | `filename_hint` and `content_type_hint` are advisory only; accepted upload contract is server-approved.                                                                                                                                                                   | Core 01 §16; Core 03 §8   |
| Pagination cursor                               | Current profile uses snapshot-stable continuation; later live changes do not alter later pages in the same cursor chain.                                                                                                                                                  | Core 01 §3.3.7            |
| Terminal jobs                                   | Terminal job resources are retained for at least 7 days, but expiring a job resource does not delete durable outputs.                                                                                                                                                     | Core 01 §3.3.9            |

## 15. Closed vocabulary quick registry

This section is a quick domain reference. Core 02 §18 owns exact token sets and must be updated first for behavior-affecting vocabulary changes.

| Token family                                                          | Exact tokens                                                                                                                                                                                     |
| --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `entity_mentions.resolution_status`                                   | `unresolved`, `resolved`, `dismissed`                                                                                                                                                            |
| `entity_mentions.origin_kind` and `indicator_observation.origin_kind` | `manual_entry`, `clipboard_paste`, `csv_import`, `xlsx_import`, `api_import`, `extraction`, `system`                                                                                             |
| `host.entity_origin` and `identity.entity_origin`                     | `entity_sheet`, `entity_import`, `created_from_mention`, `system_upsert`                                                                                                                         |
| Host and identity preserved identifier classification                 | `exact_match_reuse`, `suggestion_only`, `provenance_only`                                                                                                                                        |
| `party.party_kind`                                                    | `person`, `team`, `organization`, `distribution_list`, `other`                                                                                                                                   |
| `indicator.value_kind`                                                | `atomic`, `pattern`, `reference`                                                                                                                                                                 |
| `assessment_state`                                                    | `unknown`, `suspected`, `confirmed`, `disproven`, `cleared`                                                                                                                                      |
| `task_request.task_kind`                                              | `question`, `request`, `collection`, `containment`, `follow_up`                                                                                                                                  |
| `task_request.status`                                                 | `open`, `in_progress`, `blocked`, `done`, `canceled`                                                                                                                                             |
| `task_request.priority`                                               | `low`, `normal`, `high`, `urgent`                                                                                                                                                                |
| `decision.decision_type`                                              | `scope`, `containment`, `communication`, `evidence`, `reporting`                                                                                                                                 |
| `decision.status`                                                     | `proposed`, `approved`, `rejected`, `superseded`, `executed`                                                                                                                                     |
| `record_links.link_type`                                              | `observed_on_host`, `observed_as_identity`, `references_indicator`, `attached_evidence`, `references_artifact`, `derived_from`, `merged_into`, `supported_by`, `references_record`, `supersedes` |
| `record_links.provenance`                                             | `manual`, `auto_match`, `import`, `rollback`, `system`                                                                                                                                           |
| `artifact.comm_type` for `artifact_type='comm_log'`                   | `meeting`, `notification`, `approval`, `briefing`, `handoff`                                                                                                                                     |
| `handoff.ack_state`                                                   | `pending`, `acknowledged`                                                                                                                                                                        |
| `lesson.closure_state`                                                | `open`, `closed`                                                                                                                                                                                 |
| `finding.kind`                                                        | `finding`, `hypothesis`                                                                                                                                                                          |
| `finding.state`                                                       | `open`, `closed`                                                                                                                                                                                 |
| `finding.confidence_band`                                             | `unset`, `low`, `medium`, `high`                                                                                                                                                                 |
| `forensic_keyword.match_mode`                                         | `literal`, `regex`                                                                                                                                                                               |
| `release_state`                                                       | `pending_approval`, `approved`, `invalidated`, `published`                                                                                                                                       |
| `object_blobs.upload_state`                                           | `pending`, `available`, `failed`, `quarantined`                                                                                                                                                  |
| `object_blobs.terminal_reason`                                        | `pending_timeout`, `finalize_retry_exhausted`, `declared_size_mismatch`, `expected_sha256_mismatch`                                                                                              |
| `evidence_records.lifecycle_state`                                    | `requested`, `pending_receipt`, `received`, `available`, `quarantined`, `released`                                                                                                               |
| `evidence-access media_class`                                         | `image`, `pdf`, `text`, `audio`, `video`, `archive`, `office_document`, `binary`, `active_content`                                                                                               |
| `evidence-access preview_kind`                                        | `image_inline`, `pdf_inline`, `text_inline`                                                                                                                                                      |
| Base incident roles                                                   | `viewer`, `editor`, `reviewer`, `admin`                                                                                                                                                          |
| Saved-view scope                                                      | `private`, `shared`, `system`                                                                                                                                                                    |
| Local credential recovery model                                       | `admin_assisted`                                                                                                                                                                                 |
| TOTP state                                                            | `not_enrolled`, `pending`, `active`                                                                                                                                                              |

Display labels MAY map these tokens for users, but authoritative structured state and machine-readable payloads MUST use the exact tokens owned by the applicable core section.

## 16. External-system and extension boundaries

| External or adjacent concern          | Domain boundary                                                                                                         | Required language                                                               |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| SIEM, EDR, telemetry stores           | External source or pivot target. Cartulary may reference queries, indicators, evidence, or findings derived from them.  | Do not call raw telemetry Cartulary source state.                               |
| Object storage                        | Authoritative binary evidence backing service. Cartulary owns object metadata and evidence access semantics.            | Do not expose raw object-store URLs as evidence identity.                       |
| Enterprise IdP                        | Optional enterprise authentication provider. Successful provider auth maps to internal user and server-managed session. | Do not call provider subject a party, incident identity, or authorization role. |
| CMDB or asset inventory               | External enterprise master data. Cartulary host records remain incident-scoped investigation records.                   | Do not treat external asset identity as `record_id`.                            |
| Ticketing system                      | External task or request system. Cartulary task requests may store `external_ticket_ref`.                               | Do not replace task lifecycle with ticket state unless an owner spec says so.   |
| Reference-data sources                | Optional packs or enrichment inputs.                                                                                    | Do not block base capture on live reference data.                               |
| Report templates and rendered outputs | Snapshot/reporting extension artifacts.                                                                                 | Do not treat report release state as live workbook approval.                    |
| Backup and restore systems            | Deployment-local operator-facing recovery.                                                                              | Do not expose backup/restore as workbook route families in the current profile. |

## 17. Coding-agent rules

A coding agent working on Cartulary domain-facing code, tests, specs, comments, or generated contracts MUST apply these rules before making changes.

### 17.1 Required orientation

1. Identify the bounded context in §10.
2. Identify the domain terms in §11.
3. Identify the primary owner sections before asserting behavior.
4. Use stable identifiers from §8 instead of visible labels.
5. Treat implementation modules as mappings, not definitions.
6. When a behavior is unclear, use `TODO: owner unresolved` or request an owner-section decision rather than inventing behavior.

### 17.2 Prohibited shortcuts

A coding agent MUST NOT:

1. Treat `party`, `user`, `identity`, `incident membership`, and email text as interchangeable.
2. Treat `artifact` as a synonym for Notes or binary evidence.
3. Treat a required system view as a saved view.
4. Infer behavior from visible tab labels, visible column labels, row order, SQL names, projection table names, or React component names.
5. Mutate projection state as authoritative source state.
6. Auto-create host or identity records from `mention_origin` fields.
7. Auto-create or auto-link party records from ordinary party text.
8. Treat an object blob as an evidence row or evidence availability proof.
9. Store relationships, canonical identifiers, timestamps used for sorting, or evidence retrieval metadata only in JSON or raw text.
10. Add mandatory owner, approver, challenge, checklist, task, or decision fields to ordinary timeline capture.
11. Implement coordination surfaces as separate application modules that leave the workbook interaction model.
12. Expose ATT&CK, D3FEND, VERIS, or other pack overlays as base workbook `view_schema` resources.
13. Treat release approval as ordinary row-edit approval.
14. Use findings or hypotheses as first-class `record_type`s in the current profile.
15. Treat import worksheet/range/table identity as runtime workbook identity.

### 17.3 Review checklist for generated or agent-authored text

A reviewer MUST reject generated text that:

- defines a Cartulary term without an owner reference;
- introduces a synonym for an exact token in §15;
- presents implementation package names as domain definitions;
- omits the `Not this` boundary for a high-risk term;
- claims behavior from an appendix, guide, or research report without checking whether the core owns that behavior;
- describes a future or extension capability as Base Profile behavior;
- uses display names where stable identifiers are required.

## 18. Maintenance rules

`docs/domain.md` MUST remain small enough to be used as prompt context and rigorous enough to prevent semantic drift.

### 18.1 Required update triggers

A change set MUST update `docs/domain.md` or explicitly state `domain vocabulary unchanged` when it changes any of the following:

- first-class record type membership;
- standardized `view_schema_id` membership;
- stable identifier family;
- closed vocabulary token set;
- entity-binding mode behavior;
- party-reference semantics;
- artifact-backed variant membership;
- evidence or blob lifecycle vocabulary;
- saved-view scope semantics;
- workbook startup surface semantics;
- import object vocabulary;
- snapshot/report/release vocabulary;
- extension-profile claim vocabulary;
- a term listed in §5, §8, §9, §11, or §15.

### 18.2 Drift-control rules

1. `docs/domain.md` MUST link or point to owner sections instead of copying full route contracts or field registries.
2. Any copied exact token list MUST identify the owner section and be updated in the same change set as the owner change.
3. New glossary entries MUST include a definition, at least one forbidden interpretation when ambiguity is plausible, canonical identifiers when applicable, and an owner reference.
4. Future-only terms MUST be labeled as future-only and MUST NOT be mixed into Base Profile current terms.
5. Deprecated or migrated terms MUST name the canonical replacement and the owner section that permits migration handling.
6. A reviewer MUST treat near-duplicate terms as drift risk unless the distinction is explicit.

### 18.3 Suggested repository checks

The repository MAY add lightweight checks that verify:

- every `cartulary.view.*.v1` string in `docs/domain.md` exists in the generated view-schema registry, is listed by Core 01 as a standardized optional surface, or is labeled future-only;
- every exact token table in §15 matches the owner-derived generated token registry;
- every glossary owner reference names a known core, companion, guide, appendix, or accepted later NLSpec;
- no section defines behavior using a visible label when a stable identifier exists.

## 19. Acceptance criteria

The document is useful and complete enough for practical repository use only when every criterion in this section passes.

| ID            | Criterion                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DOMAIN-AC-001 | The document states that it is a domain-language reference and does not replace Core 00 through Core 04 implementation conformance.                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| DOMAIN-AC-002 | The document states that Core 05 governs claim-bearing publication only and does not broaden Base Profile runtime behavior.                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| DOMAIN-AC-003 | The document contains a relationship map that distinguishes Core 00, Core 01, Core 02, Core 03, Core 04, Core 05, appendices, guides, README, code comments, and `AGENTS.md`.                                                                                                                                                                                                                                                                                                                                                                                                                             |
| DOMAIN-AC-004 | The document includes explicit resolved terminology decisions for `artifact`, Notes, `party`, user, identity, incident membership, system view, saved view, projection, mention, indicator observation, evidence record, object blob, and coordination surfaces.                                                                                                                                                                                                                                                                                                                                          |
| DOMAIN-AC-005 | The document lists all fourteen required base-profile `view_schema_id` values and labels the three standardized optional workbook surfaces as optional.                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| DOMAIN-AC-006 | The document states that saved views over required view schemas are additive and non-canonical.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| DOMAIN-AC-007 | The document states that visible labels, row order, column labels, projection names, and storage names are not public mutation identifiers.                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| DOMAIN-AC-008 | The document distinguishes domain concepts from implementation details and external-system concerns.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| DOMAIN-AC-009 | The document includes a bounded-context table that describes domain responsibility without making package layout authoritative.                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| DOMAIN-AC-010 | The glossary includes, at minimum, Incident, Record envelope, Timeline event, Host, Identity, Party, User, Incident membership, Deployment admin, Artifact, Note, Task request, Decision, Communications Log, Handoff, Status Review, Lesson, Finding, Entity mention, Stub entity, Indicator, Indicator observation, Compromise assessment, Evidence record, Object blob, Record link, Tag, View schema, Workbook surface, Built-in tab, System view, Saved view, Projection, Change set, Mutation entry, Record revision, Rollback, Reference pack, Import session, Import unit, Snapshot, and Release. |
| DOMAIN-AC-011 | The document includes exact current-profile token sets or points to the generated owner-derived token registry for exact token sets.                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| DOMAIN-AC-012 | The document includes workflow vocabulary for rough timeline capture, mention resolution, indicator capture, evidence attachment, task requests, decisions, coordination artifacts, saved views, conflicts, imports, and snapshots/releases.                                                                                                                                                                                                                                                                                                                                                              |
| DOMAIN-AC-013 | The document includes defaults or omitted-case rules for the domain areas most likely to cause agent divergence.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| DOMAIN-AC-014 | The document includes a coding-agent rule section with explicit prohibited shortcuts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| DOMAIN-AC-015 | The document includes maintenance triggers for vocabulary updates.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| DOMAIN-AC-016 | A reviewer can answer “which owner section should I inspect next?” for every glossary term.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| DOMAIN-AC-017 | A coding agent can determine whether a term is a domain concept, implementation detail, external concern, Base Profile term, optional standardized surface, extension-profile term, or future-only term.                                                                                                                                                                                                                                                                                                                                                                                                  |
| DOMAIN-AC-018 | The document does not duplicate full route contracts, exhaustive field registries, or acceptance criteria already owned by Core 01 through Core 04.                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| DOMAIN-AC-019 | The document does not introduce a new authority model, new runtime behavior, new route family, new record type, new closed-vocabulary token, or new workbook surface.                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| DOMAIN-AC-020 | The document can be committed as `docs/domain.md` without requiring behavior or conformance-scope changes to Core 00 through Core 05.                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |

## 20. Open issues and future additions

This revision intentionally leaves the following as owner-driven future work rather than resolving them in `docs/domain.md`:

| Topic                                          | Current status                                                                                  | Required next step                                                                                                          |
| ---------------------------------------------- | ----------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| New first-class hypothesis record type         | Not current-profile behavior. Hypotheses remain artifact-backed findings.                       | Define a later NLSpec or profile if promotion becomes necessary.                                                            |
| Party merge and phone-based dedupe             | Not standardized in the current profile.                                                        | Define explicit party merge, phone normalization, and dedupe semantics before implementing.                                 |
| Additional framework overlay workbook surfaces | Not standardized as Base Profile or current Reference Pack Extension Profile workbook surfaces. | Define exact `view_schema_id`, fields, write behavior, pack dependencies, and compatibility rules in a later owner section. |
| Generalized approval workflows                 | Out of scope for ordinary row edits.                                                            | Define a future bounded workflow profile only if specific domain evidence justifies it.                                     |
| Live sensitive-evidence visibility model       | Out of scope for the base live workspace model.                                                 | Define a later profile if export-scoped withholding is insufficient.                                                        |
| Cross-incident analytics                       | Reserved for future specification work.                                                         | Define cross-incident data model, privacy boundary, query surface, and conformance criteria.                                |
