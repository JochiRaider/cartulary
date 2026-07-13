---
title: Cartulary Domain Vocabulary and Boundaries
class: domain-language reference
status: nlspec-grade-domain-artifact
---

## 1. Status and authority

This document is the first-class repository reference for Cartulary domain vocabulary, domain boundaries, and terminology interpretation. It exists so developers, reviewers, specification authors, and coding agents use the same project-specific meanings for terms such as `party`, `artifact`, `view schema`, `saved view`, `system view`, `entity mention`, `object blob`, and `workbook surface`.

This document does not replace the current implementation-conformance corpus. Core 00 through Core 04 remain the implementation-conformance authority for the current profile. Core 05 remains a normative companion for claim-bearing publication only. Non-normative appendices, research reports, implementation guides, design guides, and operating guidance remain subordinate according to the existing authority model.

`docs/domain.md` owns repository vocabulary, concept classification, owner-section navigation, domain-boundary interpretation, cross-context terminology hygiene, and coding-agent terminology discipline. It MUST NOT create implementation-conformance behavior, route families, record types, field registries, closed-vocabulary token membership, security behavior, deployment behavior, benchmark-publication behavior, or workbook surfaces.

If this document and a primary owner section differ, the owner section governs. The difference is documentation drift and MUST be repaired by updating `docs/domain.md`, the owner section, or both through the repository's ordinary specification-change process. If two owner sections appear to conflict, that conflict is a corpus defect and MUST NOT be resolved by this document alone.

Creation or revision of this document identifies no authority-model conflict requiring a behavior or conformance-scope change to Core 00. The vocabulary resolutions in this document are least-disruptive interpretations of terms already present in the current document set.

### 1.1 Normative language and statement classes

The normative words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** in this document govern only repository vocabulary use, domain interpretation, documentation discipline, review discipline, and coding-agent context.

- **MUST** and **MUST NOT** define domain-document requirements inside this document's authority boundary.
- **SHOULD** and **SHOULD NOT** define strong domain-document defaults whose exceptions must remain compatible with the owner sections and this document's vocabulary boundaries.
- **MAY** defines optional domain-document behavior only when omission behavior is specified in the same paragraph, table row, or immediately following paragraph.
- `default` defines the required value, interpretation, or classification when the relevant caller, document, or owner section omits a more specific value.

Normative text in this document MUST NOT be read as adding, widening, narrowing, or replacing runtime behavior unless the same behavior is owned by the applicable Core 00 through Core 04 section. A sentence that states behavior owned by another document is a domain-facing restatement only; the owner section remains the behavioral authority.

| Statement class | Meaning in `docs/domain.md` | Required handling |
| --- | --- | --- |
| Domain-owned requirement | Vocabulary, classification, owner navigation, boundary interpretation, or agent/reviewer discipline owned by this document. | Binding within this document's authority boundary. |
| Owner restatement | Compact summary of behavior owned by Core 00 through Core 05 or a later adopted NLSpec. | MUST name the owner; MUST NOT become independent authority. |
| Rationale | Explanation of why a term or boundary exists. | MUST NOT override a requirement. |
| Future-only note | Current-profile omission or reserved concept. | MUST include current handling in §20. |
| External boundary note | Anti-corruption or external-system translation rule. | MUST map external language to a Cartulary canonical target in §16. |

A domain-owned `MAY` without explicit omission behavior is invalid in this document. Lowercase `may` is non-normative unless it appears inside a quoted source title, source text, or example label.

### 1.2 Applicability and current-status value sets

Every glossary row, bounded-context row, surface row, and future-only row that classifies a domain concept MUST use the following closed values.

| Applicability value | Meaning |
| --- | --- |
| `base` | Current Base Profile concept or vocabulary term that is not itself a surface-status row. |
| `base-required-surface` | Required current-profile workbook surface. |
| `standardized-optional-surface` | Current-profile standardized optional workbook surface when implemented under its owner contract. |
| `extension-profile` | Concept that applies only when its named extension profile is claimed. |
| `future-only` | Reserved concept that is not current-profile behavior. |
| `external` | Concept owned by an external system or external model and admitted only through an anti-corruption boundary. |
| `implementation-support` | Repo-local implementation, testing, guide, or generated-artifact term that is not domain behavior. |
| `non-normative-guidance` | Operating, rationale, or design-support term that does not define implementation conformance. |

| Current-profile status value | Meaning |
| --- | --- |
| `current-required` | Required by the current Base Profile or required for current domain vocabulary interpretation. |
| `current-optional-when-implemented` | Valid in the current profile only when the standardized optional surface or optional owner-defined feature is implemented. |
| `current-extension-when-claimed` | Valid only when the matching extension profile is claimed. |
| `current-external-boundary` | Valid only as external input, reference, or translation source. |
| `implementation-support-only` | Valid only as repository implementation-support language. |
| `non-normative-only` | Valid only as rationale, guidance, or appendix-style explanation. |
| `not-current` | Not valid as current-profile behavior unless a later owner spec defines it. |

A row that cannot be classified by these tables MUST use `TODO: owner decision required` and MUST NOT be treated as current-profile behavior.

## 2. Purpose

`docs/domain.md` MUST provide a compact, rigorous domain reference that makes Cartulary-specific language operationally usable during implementation and review.

The document MUST serve these functions:

1. Define canonical domain terms and their forbidden substitutions.
2. Map domain terms to their language home and behavioral owner sections without copying full owner contracts.
3. Distinguish domain concepts from implementation modules, physical tables, component names, route helpers, guides, tests, and external-system concerns.
4. Preserve the current authority model by treating owner sections as the behavioral source of truth.
5. Give coding agents enough stable context to avoid semantic aliasing, storage-model leakage, visible-label inference, external-model leakage, and forms-first or module-first implementation mistakes.
6. Define Strategic Domain-Driven Design boundaries: ubiquitous language, bounded contexts, context relationships, upstream/downstream direction, anti-corruption boundaries, and subdomain classification.
7. Provide binary acceptance criteria for evaluating whether this document is complete enough to use during development.

`docs/domain.md` MUST NOT be used as an API reference, schema reference, implementation guide, operating handbook, UI design guide, route inventory, migration inventory, generated-contract registry, benchmark specification, or test plan. It MAY point to those artifacts when they are the correct owner for a term or behavior. Omission behavior: if this document does not name a route, field, token, or runtime behavior, the omission MUST be interpreted as delegation to the owner section, not as permission to invent local behavior.

## 3. Domain thesis

Cartulary is a workbook-native incident workspace. Analysts act on visible rows, cells, chips, counts, filters, groups, previews, saved views, and system views, while authoritative source state remains typed, relational, versioned, attributed, and auditable underneath. The workbook surface preserves spreadsheet speed and direct manipulation. The source model rejects spreadsheet identity, overwrite, evidence, history, and relationship semantics where those semantics would undermine collaboration or recoverability.

For domain interpretation, the following statement is controlling:

> Cartulary preserves the spreadsheet mental model at the view layer, not at the storage layer.

A proposed term, feature name, route helper, UI label, implementation module, guide section, or generated artifact that contradicts this thesis MUST be treated as ambiguous until mapped to the correct owner section.

Inspector vocabulary follows the workbook thesis: analysts act on visible rows, cells, chips, previews, saved views, and inspector affordances, while behavior is keyed by stable owner identifiers. If a proposed inspector term implies hidden source state, workflow engine identity, saved-view persistence, or route-helper identity, it is ambiguous until remapped to Core 01/Core 03/Core 04.

## 4. Document relationship map

The following map preserves the existing security, implementation-support, design-direction, and publication boundaries rather than creating a new document hierarchy.

| Artifact | Owns | `docs/domain.md` role |
| --- | --- | --- |
| Core 00 | Document status, precedence, profile model, contract-owner matrix, conformance separation. | MUST follow its authority order and MUST NOT create new runtime conformance requirements. |
| Core 01 | Architecture, route families, public interfaces, view schemas, projections, jobs, portability, extension routes, snapshot and reference-pack route families. | Summarizes domain-facing meanings of stable identifiers, workbook surfaces, projections, and route families only when needed for vocabulary navigation. |
| Core 02 | Record model, mention and entity semantics, party model, task requests, decisions, artifacts, indicators, assessments, relationships, history substrate, closed vocabularies. | MUST use Core 02 as the primary behavioral owner for entity, record, and token vocabulary when that behavior is not owned elsewhere by Core 00 §5.1. |
| Core 03 | Workbook interaction, built-in tabs, system views, saved views, startup surface selection, collaboration, same-field conflicts, workflows, grouping, coordination interaction behavior. | MUST use Core 03 as the primary behavioral owner for interaction-domain vocabulary. |
| Core 04 | Authentication, authorization, deployment, trust boundaries, runtime roots, conformance criteria. | MUST use Core 04 as the primary behavioral owner for users, sessions, deployment administration, incident roles, security boundaries, and release gates. |
| Core 05 | Claim-bearing publication and benchmark reproducibility. | MUST NOT treat Core 05 as Base Profile runtime behavior. |
| Adopted subsystem NLSpecs | Bounded subsystem behavior after repository adoption. | References subsystem terms only inside the adopted subsystem boundary and MUST NOT let subsystem terms redefine root domain vocabulary. |
| Report Composition NLSpec | Report-composition authoring resource lifecycle, schema, operations, semantic anchors, authored presentation text, composition diagram declarations, closed diagram layout data, and builder UI conformance boundaries when adopted. | Provides extension-profile vocabulary for composition authoring only and MUST NOT redefine reports, templates, workbook artifacts, snapshots, or release outputs. |
| Network Flow Activity NLSpec | Adopted analytical-table, normalized flow-row, graph-adapter, indicator-binding, and Network Analysis workspace owner under the Core 00 `network_flow_activity` boundary. | Supplies current extension vocabulary when claimed; MUST NOT redefine Core records, view schemas, saved views, import sessions/units, or Base Profile surfaces. |
| Appendices A through I | Rationale, illustrations, non-normative operating guidance, source preservation, backlog, traceability, and projection authority/boundary evidence. | Uses appendix material as evidence or orientation only when not in conflict with the core. |
| Research reports | Evidence, comparative analysis, design rationale, or source archaeology. | Inform rationale only and MUST NOT define current-profile behavior. |
| UI/UX design guide and `design.md` | Derived design-direction specifications. | Use for user-facing mental-model language only when behavior remains owned by the core. |
| Development, bootstrap, frontend, visual, and testing guides | Repo-local implementation baseline, implementation-support procedure, harness mechanics, visual maintenance, and planning. | Use for package/module mappings, but implementation modules and test rows are not domain definitions. |
| README | Public orientation and onboarding. | Default repository guidance points readers to `docs/domain.md` for vocabulary. |
| Code comments | Local implementation rationale. | MUST use canonical terms from `docs/domain.md` and MUST NOT redefine domain concepts locally. |
| `AGENTS.md` | Coding-agent and contributor procedure when present. | Default coding-agent procedure instructs agents to consult `docs/domain.md` before domain-facing changes. |

### 4.1 Cross-spec dependency and delegation matrix

| Adjacent spec or document | Concept retained by `docs/domain.md` | Behavior delegated | Boundary rule |
| --- | --- | --- | --- |
| Core 00 | Authority vocabulary, profile vocabulary, conformance-boundary terms, owner-section navigation. | Authority order, profile claims, contract-owner matrix, current/future profile closure. | `docs/domain.md` imports Core 00 authority and MUST NOT reopen it. |
| Core 01 | Domain-facing stable identifiers, workbook-surface names, projection vocabulary, route-family nouns, import and extension-profile object names. | Route shapes, public envelopes, view-schema resources, field registries, sort/filter/group algorithms, background jobs, import/session route behavior, portability, and extension routes. | `docs/domain.md` can name `view_schema_id` values and route-family nouns for vocabulary navigation but MUST NOT restate route contracts or field registries. |
| Core 02 | Record-family vocabulary, entity/mention/party/artifact/indicator/evidence/history terms, closed-token-family names. | Record-type membership, schema invariants, exact token membership, create/update defaults, provenance, merge, dedupe, history, rollback substrate. | Exact token membership MUST be owner-derived or pointer-only in this document. |
| Core 03 | Workbook-surface vocabulary, saved-view vocabulary, workflow terms, interaction-state names. | Interaction algorithms, saved-view lifecycle, startup selection, conflict resolution, workbook write-back, grouping, paste, collaboration. | Workflow vocabulary in this document MUST remain orientation and MUST NOT define transition matrices unless it cites the owner. |
| Core 04 | User, session, role, deployment-admin, trust-boundary, release, and authorization vocabulary. | Authentication, authorization, credential lifecycle, incident role behavior, deployment, evidence access security, release gate behavior, acceptance criteria. | `party`, `identity`, `user`, and incident-role language MUST NOT imply authorization behavior beyond Core 04. |
| Core 05 | Publication and benchmark claim vocabulary. | Benchmark profiles, manifests, measurement predicates, claim-bearing publication conditions. | Document-readiness criteria in §19 are not Core 05 publication evidence. |
| Report Composition NLSpec | Report composition resource, composition draft, composition version, composition operation, semantic composition anchor, authored presentation text, composition diagram declaration, composition diagram layout, composition preview, and report builder vocabulary. | Authoring routes, resource lifecycle, canonical composition schema, operation wire vocabulary, builder-facing validation, preview admission, and builder UI conformance boundaries. | Composition terms are incident-scoped authoring inputs and MUST NOT be promoted to report, template, workbook artifact, snapshot, or release-output terms. |
| Appendices and research reports | Rationale terms and historical context. | No current-profile behavior. | A term from an appendix or research report becomes domain vocabulary only when this document or a core owner admits it. |
| Guides | Implementation-support terms and planning labels. | Package boundaries, harness mechanics, visual fixture maintenance, frontend phase evidence. | Guide terms MUST NOT become domain concepts by repetition in code or tests. |

### 4.2 Strategic DDD source posture

`docs/domain.md` uses Strategic Domain-Driven Design terminology only inside this document's authority boundary.

| Source family | Status in this document | Permitted use |
| --- | --- | --- |
| Evans DDD Reference and Evans original DDD text | Definition baseline | Define the meanings of ubiquitous language, bounded context, context map, and original strategic relationship vocabulary. |
| Fowler bounded-context explanation | Practitioner clarification | Clarify bounded-context interpretation where it does not conflict with Evans or Cartulary owner sections. |
| DDD Crew context-mapping and bounded-context artifacts | Practitioner pattern and documentation reference | Inform relationship-type naming, team-contact awareness, and bounded-context documentation structure. |
| Context Mapper and related strategic-DDD tooling literature | Tooling and validation reference | Inform machine-readable context-map and bounded-context validation shapes. |
| Academic DDD research | Evidence only | Identify boundary-maintenance risks, tooling value, and empirical limits. Academic sources MUST NOT override Cartulary owner sections or this document's closed vocabulary. |
| Cartulary owner sections | Behavioral authority | Govern product behavior, conformance, route shape, record type membership, security behavior, and exact token membership. |

When Strategic DDD terminology and Cartulary owner sections appear to conflict, the owner section governs product behavior and `docs/domain.md` MUST either repair the terminology or mark the row `TODO: owner decision required`.

A future revision MAY add source rows to this table. Omission behavior: omitted Strategic DDD sources are not authority for this document unless a later accepted revision adds them here.

## 5. Resolved terminology decisions

The following decisions resolve overloaded or easily-confused language for current-profile repository use. These decisions do not change owner-section behavior.

| Issue | Canonical interpretation | Forbidden interpretation | Primary owner |
| --- | --- | --- | --- |
| `artifact` versus Notes | `artifact` is the structured text object family. Notes is one workbook surface backed by `artifact_type='note'`. | Treating `artifact` as a synonym for the Notes tab or as binary evidence. | Core 02 §2, §10.4.4A; Core 01 §7.4 |
| `party` versus user | `party` is an incident-scoped coordination identity. A user is a deployment-local login and attribution identity. | Treating `party_id`, `user_id`, email text, auth subject, or incident membership as interchangeable. | Core 02 §19; Core 04 §2 |
| `identity` versus party | `identity` is a host/account/persona investigation entity. `party` is a coordination stakeholder identity. | Using identities to model requesters, audiences, attendees, or collectors when the domain needs party references. | Core 02 §2, §19 |
| `system view` versus `required system view` | `system view` is the non-built-in workbook-surface kind `surface_kind='system_view'`. `required system view` is a `surface_status` subset. | Treating every `surface_kind='system_view'` row as Base Profile required. | Core 01 §7.4; Core 03 §2.2 |
| `system view` versus `system` saved view | A system view is a contract-backed workbook surface identified by `view_schema_id`. A `scope='system'` saved view is an implementation-owned saved-view configuration object. | Treating a `system` saved view as the required system view itself. | Core 01 §7.4; Core 03 §2.3 |
| `view_schema_id` versus visible label | `view_schema_id` is the canonical workbook-surface identity. Visible titles and labels are display hints. | Deriving behavior from tab labels, column labels, visible row order, projection names, or storage names. | Core 01 §3.3.1, §7.4 |
| Projection versus source state | A workbook projection is a derived read model for workbook query, row refresh, sorting, filtering, and grouping. Source state is the authoritative typed record, link, evidence, history, or admin state. | Mutating projections as source of truth or treating projection corruption as source-state corruption. | Core 01 §8; Core 02 §1, §15 |
| Administrative audit versus incident revision history | Administrative audit is deployment-local administrative evidence for auth, deployment administration, incident-membership administration, import bootstrap, and recovery-operation summaries. Incident revision history is source-state row/version history inside incident data. | Treating `/api/v1/administrative-audit-events` as incident row revision history, exporting it in incident portability bundles, or creating a separate audit bounded context before Core 01/Core 04 widen audit beyond deployment-local admin/auth evidence. | Core 00 §5.1; Core 01 §3.3.5.1A; Core 04 §9.10 |
| Workbook projection versus graph projection | Workbook projections are Core-owned workbook read models. Graph projection is an adopted graph-oriented derivation subsystem governed by `docs/graph_projection_nlspec.md`. | Applying graph-projection lifecycle or query rules to workbook-grid projection tables, workbook query routes, saved views, restore rebuilds, or `view_row_v1`. | Core 01 §8; `docs/graph_projection_nlspec.md` |
| Entity mention versus stub entity | An entity mention is a source-bound observation. A stub entity is a real host or identity record with its own `record_id`. | Treating unresolved mentions as weak entities or auto-creating stubs from every mention. | Core 02 §6 |
| Indicator observation versus indicator | An indicator observation is source-bound. A canonical indicator is a first-class incident-scoped record. | Treating every observed indicator-like string as an automatically created canonical indicator. | Core 02 §6, §10 |
| Evidence record versus object blob | Evidence record is the user-facing evidence envelope. Object blob is binary-object metadata and upload state. | Treating object blobs as workbook evidence rows or storing raw evidence inside timeline cells. | Core 02 §13; Core 03 §8 |
| Coordination surface versus workflow engine | Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson are workbook-native coordination surfaces. | Creating a generalized approval/workflow engine for ordinary row edits. | Core 02 §10.4; Core 03 §16.4 |
| Import unit versus worksheet/table | `import_unit` is the contract object for one candidate ingestable unit. Worksheet ranges and tables are locator kinds. | Using “selected sheet import” or “Excel table import” as alternate public contract nouns. | Core 01 §2.1; development guide §11 |
| Reference pack versus incident data | Reference packs are versioned optional vocabularies, frameworks, or enrichment datasets outside incident source records. | Treating reference-pack activation as incident record mutation or blocking core capture on pack availability. | Core 01 §11; Core 02 §17 |
| Incident TLP token versus display label | `tlp` machine state stores and exchanges only the canonical Core 02 token or `null`. UI labels, colors, and localized text are presentation. | Persisting display labels, aliases such as `WHITE`, or case variants as incident TLP state. | Core 01 §3.3.5.3; Core 02 §18 |
| Incident severity or phase suggestion versus stored metadata | `severity` and `current_phase` are organization-specific bounded incident metadata text; reference packs may suggest values or labels only. | Rejecting otherwise valid bounded text, reinterpreting stored values, requiring a reference pack, or blocking incident creation or capture because a value is outside a suggested list. | Core 01 §3.3.5.3; Core 02 §4.5 |
| Incident lifecycle state versus archive/delete/purge | `incident.status` is the current-profile lifecycle state vocabulary for the active or closed incident workspace. Closure is a read-only source-state boundary, not an incident-removal model. | Treating `closed` as archived, deleted, purged, hidden from current members, or equivalent to any future retention/tombstone behavior. | Core 01 §3.3.5.3.2; Core 02 §18; Core 03 §4.3.2 |
| Snapshot/report versus live workbook | Snapshot/report artifacts are immutable export or publication inputs under extension rules. Live workbook state is the operational incident workspace. | Applying recipient-specific export redaction by hiding live workbook rows from incident members. | Core 01 §10; Core 04 §2.1, §4.2 |
| Report composition versus report/template/workbook artifact | A report composition is an incident-scoped authoring input that can be bound to a report render by digest. It is not itself a report, template, workbook artifact, snapshot, or release output. | Treating composition authoring as generated-source editing, workbook mutation, template mutation, snapshot mutation, or release-byte editing. | `docs/report-composition-nlspec.md`; `docs/reporting-subsystem-nlspec.md` |

## 6. Domain versus implementation detail

A term belongs in `docs/domain.md` when misunderstanding it would cause a developer, reviewer, or coding agent to build the wrong behavior, address the wrong contract, mutate the wrong source of truth, violate a bounded context, or write a misleading specification.

A term usually belongs outside `docs/domain.md` when it is only one of the following:

- package, module, or directory name;
- SQL table, generated column, trigger, index, or migration filename;
- React component, hook, CSS class, or grid vendor coordinate;
- Go type, helper function, adapter method, or internal service interface;
- object-store key realization;
- test harness implementation detail;
- deployment-specific secret, path, or operator command.

An implementation detail MAY appear in `docs/domain.md` only as an implementation-facing mapping that prevents ambiguity. Omission behavior: if a mapping is omitted, the implementation detail MUST NOT be treated as a domain term merely because it appears in code, tests, generated artifacts, or documentation.

| Domain term | May mention implementation mapping | Boundary |
| --- | --- | --- |
| Workbook surface | `view_schema_id`, `sheet_ref`, query route, grid adapter. | The domain identity is not the React component, route handler, projection table, or visible tab label. |
| Party | `record_type='party'`, `party_id`, Parties view. | The domain identity is not a deployment-local user, email string, auth subject, or contact-directory entry. |
| Notes | `cartulary.view.notes.v1`, `artifact_type='note'`. | Notes is one artifact-backed surface, not the whole artifact family. |
| Projection | Workbook projection table or equivalent denormalized read model. | Projection state is disposable and derived, not authoritative source state; graph projection is a separate adopted subsystem and is not workbook authority. |
| Object blob | Object storage metadata and upload slot. | The object-store key is implementation realization, not public evidence identity. |
| Import unit | `import_unit`, locator kind, mapping fingerprint. | XLSX worksheet, used range, table, and named range are locator examples, not cross-module semantics. |

### 6.1 Term classification decision tree

A reviewer or coding agent MUST classify a candidate term by the first matching row in this table. Later rows MUST NOT override an earlier row.

| Order | Candidate condition | Required classification | Required location | Omission behavior |
| ---: | --- | --- | --- | --- |
| 1 | Exact stable public identifier named in §8 or an owner section. | Domain concept. | `docs/domain.md` §8 or owner section. | If absent from §8, add only when it is domain-facing and owner-backed. |
| 2 | Current-profile record, surface, workflow, party, entity, evidence, history, or token-family term in Core 00 through Core 04. | Domain concept. | `docs/domain.md` §11 or relevant section. | If owner is unclear, write `TODO: owner decision required`. |
| 3 | Public wire/resource member or route-family noun whose misunderstanding changes behavior. | Domain-boundary term. | Owner spec; optional summary in `docs/domain.md`. | Do not copy route or field contract. |
| 4 | `view_schema_id`, `field_key`, saved-view, projection, or workbook-surface field identity. | Domain-boundary term. | Core 01/Core 03 owner, summarized in §8 or §9. | Do not infer from labels. |
| 5 | Domain-facing UI label whose text may change but maps to a canonical term. | Display label. | UI/design docs may define label; `docs/domain.md` may map only if ambiguity exists. | Do not treat the label as stable identity. |
| 6 | Generated contract name derived from owner specs. | Generated artifact. | Generated registry or contracts directory. | Do not define in `docs/domain.md` unless it names a domain-facing stable identifier. |
| 7 | Package, module, component, helper, table, migration, test, or harness term. | Implementation-support term. | Development guide, harness spec, code, or test docs. | Omit from `docs/domain.md` unless needed as an ambiguity-prevention mapping. |
| 8 | External-system term, external ID, provider claim, telemetry concept, CMDB object, ticket status, or SIEM/EDR object. | External term. | §16 anti-corruption map or integration owner. | Preserve only in the allowed raw/source form. |
| 9 | Roadmap, future profile, rejected alternative, or non-normative research term. | Future-only or rationale. | §20, appendix, roadmap, or research report. | Must not appear as current-profile behavior. |

### 6.2 Alias, deprecated-term, and forbidden-term registry

Rows in this table are exhaustive for the high-risk aliases currently known to this document. A new known alias that affects domain interpretation MUST add one row here or add `TODO: owner decision required`.

| Observed term | Allowed? | Canonical term | Allowed context | Forbidden context | Owner |
| --- | --- | --- | --- | --- | --- |
| case | Limited | Incident | Informal human prose when the intended object is an incident. | Stable identifiers, route families, record types, or specs that need `incident`. | Core 01, Core 02 |
| worksheet | Limited | Import unit locator or visible workbook surface | Import-source discussion where worksheet is an XLSX locator kind. | Runtime workbook identity or `view_schema_id` identity. | Core 01 §2.1 |
| sheet | Limited | Built-in tab, workbook surface, or import locator depending on context | User-facing explanation after canonical term is established. | Public contract identity when `view_schema_id` or `sheet_ref` is required. | Core 01, Core 03 |
| tab | Limited | Built-in tab or visible shell label | UI copy for five built-in tabs. | Any required system view or saved view identity. | Core 03 §2 |
| IOC | Limited | Indicator observation or indicator | Practitioner prose where ambiguity is immediately resolved. | Record model, token registry, or automated creation rules. | Core 02 §6, §10 |
| artifact | Yes, canonical only | Artifact | Structured analyst text object family. | Binary evidence, object blob, Notes surface alone. | Core 02 §2, §10.4.4A |
| binary artifact | Discouraged | Evidence record plus object blob | Explanatory migration text that names the canonical replacement. | Current domain specs and generated contracts. | Core 02 §13 |
| approval | Limited | Release gate or decision status depending on context | Snapshot/reporting extension release gate. | Ordinary row editing, timeline capture, task lifecycle, or generalized workflow engine. | Core 04 §2.1 |
| owner | Limited | `owner_user_id`, party reference, or task owner depending on field | Field labels after owner field is named. | Standalone domain object or substitute for authorization role. | Core 02, Core 04 |
| actor | Limited | User or system process | Audit and attribution prose. | Party, incident member, or identity. | Core 04 §3 |
| contact | Limited | Party | UI copy only when it maps to a `party` record or raw party text. | Deployment user, identity, or external directory object. | Core 02 §19 |
| customer | Limited | Party or recipient partition | External-facing explanation with explicit mapping. | Authorization role or incident workspace identity. | Core 02 §19; Core 04 §2.1 |
| analyst | Limited | User or party, explicitly disambiguated | Human-readable role prose. | Stable `user_id`, `party_id`, or incident role field. | Core 04 §2; Core 02 §19 |
| ticket | Limited | `external_ticket_ref` or task request | External ticket references on task-request rows. | Task lifecycle source of truth. | Core 02 §10.4 |
| report | Limited | Snapshot/rendered output/release artifact | Snapshot and Reporting Extension Profile discussion. | Live workbook state or ordinary saved view. | Core 01 §10; Core 04 §2.1 |
| report composition | Yes, extension-specific | Report composition | Snapshot and Reporting Extension Profile authoring-input discussion. | Report, report template, workbook artifact, snapshot, release output, live workbook record, or generated source file. | Report Composition NLSpec |
| release | Yes, extension-specific | Release | Artifact-scoped rendered-output approval/publication record. | General row approval or live visibility gate. | Core 04 §2.1 |
| Excel table | Limited | `import_unit` locator kind | Import-source locator explanation. | Runtime workbook surface or source-of-truth table. | Core 01 §2.1 |
| flow table | Yes, extension-specific | Network Flow table | Network Flow Activity analytical-resource discussion. | SQL table, import unit, worksheet, `view_schema`, saved view, or raw telemetry stream. | Core 00 profile posture; Network Flow Activity NLSpec |
| Network Analysis | Limited visible label | Network Analysis workspace | User-facing extension-workspace label after canonical identity is stated. | Stable identifier, Base built-in tab, system view, saved view, or module identity. | Core 01/Core 03; Network Flow Activity NLSpec |
| findings document | Limited | Companion findings document | Non-normative operating guidance. | Current-profile source record unless summarized as structured finding or artifact. | Appendix H; Core owners when structured |

## 7. Core distinctions

Each distinction below is review-critical for vocabulary and documentation discipline. A change that blurs one of these distinctions MUST cite the owner section that allows the change.

| Distinction | Required interpretation | Common failure |
| --- | --- | --- |
| Workbook surface versus source state | A surface is what analysts query and edit. Source state is the record, link, mention, evidence, history, or admin model underneath. | Updating projection rows directly or storing relationship meaning in visible columns only. |
| Surface kind versus surface status | `surface_kind` identifies the surface family. `surface_status` identifies required or optional current-profile status. | Inferring requiredness from `surface_kind='system_view'`. |
| Built-in tab versus system view | Built-in tabs are five primary required sheets. System views are non-built-in contract-backed workbook surfaces. | Making coordination or indicator surfaces separate application modules. |
| Required system view versus standardized optional workbook surface | Both can have `surface_kind='system_view'`; only rows with `surface_status='required system view'` are required Base Profile surfaces. | Treating Findings, Investigative Queries, or Forensic Keywords as required surfaces. |
| System view versus saved view | A system view is a contract-backed `view_schema_id`. A saved view is an incident-bound configuration over one `view_schema_id`. | Replacing required surfaces with saved-view presets. |
| Visible label versus stable identifier | Public behavior binds to identifiers such as `view_schema_id`, `field_key`, `record_id`, `row_version`, `client_txn_id`, and `party_id`. | Inferring write behavior from labels, row order, or SQL names. |
| User versus party versus identity | User authenticates and receives attribution. Party coordinates stakeholder references. Identity is an incident entity such as an account or persona. | Using `user_id` as requester, using party as login, or using identity as stakeholder audience. |
| Mention versus entity | A mention is an observed string in a source record and field. An entity is a host or identity record. | Auto-creating entities from every host-like or account-like string. |
| Indicator observation versus canonical indicator | Observation is source-bound occurrence. Canonical indicator is first-class linkable record. | Deduping observations into one record without preserving occurrences. |
| Evidence record versus object blob | Evidence record models request, receipt, custody, availability, and workbook linkage. Object blob models binary upload and storage metadata. | Treating raw blob existence as sufficient evidence availability. |
| Artifact versus binary evidence | In Cartulary, `artifact` means structured analyst text material. Binary evidence uses evidence records and object blobs. | Calling uploaded files “artifacts” without qualifying them as evidence blobs. |
| Task/decision versus timeline field | Task Requests and Decisions are coordination records. Timeline rows capture observations and chronology. | Adding mandatory task or approval fields to timeline row creation. |
| Reference pack versus overlay surface | Reference packs provide optional vocabularies and enrichment. Pack-dependent framework overlays are not base workbook surfaces. | Exposing ATT&CK, D3FEND, or VERIS as base `view_schema` resources. |
| Snapshot/report versus live workspace | Snapshot/report is an export or release construct. Live workspace is incident collaboration. | Applying external-release redaction to live workbook visibility. |
| Analytical extension resource versus Core source record | A Network Flow table and its immutable normalized rows are extension-owned analytical resources only after the profile is claimable. They are not Core record envelopes, workbook projections, or raw external telemetry identity. | Assigning `record_id`, `view_schema_id`, row-history, or Core record mutation semantics to a flow table or flow row. |
| Extension workspace versus Base workbook surface | An extension workspace is contributed by a claimed profile under owner-defined discovery and `sheet_ref` rules. It is not a Base built-in tab, system view, saved view, or member of the exhaustive Base `view_schema_id` registry. | Adding Network Analysis to §9.2 or inferring stable identity from its visible label. |

## 8. Canonical identifier vocabulary

| Identifier | Domain target | Stable use | Not this | Owner | Applicability | Current-profile status |
| --- | --- | --- | --- | --- | --- | --- |
| `incident_id` | Incident workspace. | Scope for incident records, views, jobs, and authorization checks. | Deployment, tenant, or object-store bucket identity. | Core 01, Core 02 | `base` | `current-required` |
| `incident_key` | Deployment-unique incident key. | Human-meaningful incident identity normalized for uniqueness. | Public row identity or authorization token. | Core 02, Core 01 | `base` | `current-required` |
| `record_id` | User-visible first-class incident record envelope. | Row identity, record mutation, links, tags, history, rollback. | User identity, party reference, risk-ref child identity, or blob identity. | Core 02 | `base` | `current-required` |
| `row_version` | Current optimistic-concurrency version for a record row. | Server-emitted version anchor for current row state. | Global revision number or history-entry selector. | Core 03 | `base` | `current-required` |
| `base_row_version` | Client version anchor for an attempted record write. | Conflict detection for record-scoped writes. | Current version after successful commit. | Core 03, Core 01 | `base` | `current-required` |
| `field_key` | Stable view-field identity. | Write target, conflict key, sort/filter/group capability key. | Column label, SQL column, visible header, or translated label. | Core 01 | `base` | `current-required` |
| `view_schema_id` | Stable public workbook-surface identity. | Built-in sheet, system view, or standardized optional surface identity. | Visible tab label, saved view ID, or projection name. | Core 01 | `base` | `current-required` |
| `feature_group_key` | Inspector feature group. | Stable feature-group identity within one `view_schema_id`. | Visible label, route helper, React component, CSS selector, storage column, or grid-vendor identifier. | Core 01 §7.4 | `base` | `current-required` |
| `route_binding.owner` | Inspector route owner token. | Closed token identifying the Core-owned route family or current-row data source used by a feature group. | Implementation function name, handler name, SQL table, component name, or route-helper name. | Core 01 §7.4 | `base` | `current-required` |
| `sheet_ref` | Stable reference to a workbook surface or saved view. | Startup/default surface pointers and presence state. | Visible shell location. | Core 01, Core 03 | `base` | `current-required` |
| `saved_view_id` | Saved-view configuration object. | Incident-bound saved view identity. | Required system view identity. | Core 03 | `base` | `current-required` |
| `party_id` | Same-incident party record. | Stable coordination-party reference. | Email text, user ID, auth-provider subject, or incident membership. | Core 02 | `base` | `current-required` |
| `entity_mention_id` | Source-bound textual entity mention. | Explicit resolve or dismiss target. | Host or identity `record_id`. | Core 02, Core 01 | `base` | `current-required` |
| `object_blob_id` | Binary upload slot or object metadata. | Blob create, upload, and attach flow. | Evidence record identity. | Core 01, Core 02 | `base` | `current-required` |
| `client_txn_id` | Client-supplied idempotency key where the owner route requires it. | Replay-safe mutation boundaries. | Record ID, operation type, or global transaction ID. | Core 01 | `base` | `current-required` |
| `history_entry_ref` | Public row-history selector for eligible rollback targets. | Record-history and rollback selection. | Storage primary key or mutation-entry ID. | Core 02, Core 01 | `base` | `current-required` |
| `risk_ref_id` | Child identity scoped to one `handoff` artifact. | Public item reference for `handoff.open_risk_refs[]`. | First-class risk record or generic `record_id`. | Core 02, Core 01 | `base` | `current-required` |
| `user_id` | Deployment-local user and attribution identity. | Authentication, session, attribution, administrative account state. | Party, identity, or incident stakeholder identity. | Core 04, Core 01 | `base` | `current-required` |
| `snapshot_id` | Immutable export-model anchor. | Snapshot and Reporting Extension Profile. | Live workbook surface or saved view. | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| `release_id` | Rendered-output release record. | Snapshot/report release gate. | General row approval workflow. | Core 04, Core 01 | `extension-profile` | `current-extension-when-claimed` |
| `composition_id` | Incident-scoped report-composition resource. | Report Composition NLSpec authoring routes and Reporting release tuple references. | Report identity, template identity, workbook artifact identity, snapshot identity, or release output identity. | Report Composition NLSpec | `extension-profile` | `current-extension-when-claimed` |
| `composition_version` | Immutable report-composition version under one `composition_id`. | Digest-bound composition selection for Reporting consumption. | Template version, release version, snapshot version, workbook row version, or generated report revision. | Report Composition NLSpec | `extension-profile` | `current-extension-when-claimed` |
| `import_session_id` | Uploaded source file plus operator-driven import workflow. | Import Extension Profile source workflow. | Runtime workbook identity. | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| `import_unit` | Candidate ingestable unit from an import source. | Import source locator and mapping boundary. | Worksheet identity or runtime workbook surface. | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| `network_flow_table_id` | Network Flow analytical table resource. | Stable incident-scoped identity for one successfully published normalized flow table. | `record_id`, `view_schema_id`, `saved_view_id`, import unit, worksheet, filename, or visible tab label. | Network Flow Activity NLSpec; Core 01 for public reference unions | `extension-profile` | `current` |

### 8.1 Stable identity and display-label rule

Stable identifiers in §8 MUST be compared and discussed as exact contract identifiers according to their owner sections. Display labels MAY change without changing domain identity when the owner section allows the label change and field or surface semantics do not change. Omission behavior: when a document or code comment uses only a visible label where a stable identifier exists, reviewers MUST require the stable identifier to be added or MUST classify the text as non-authoritative UI copy.

The Base `sheet_ref` vocabulary remains owner-defined by Core 01/Core 03. A
later claimed extension may add an owner-adopted extension-workspace variant,
but this document does not define that wire shape and §9.2 MUST remain the
exhaustive Base/standardized-optional `view_schema_id` registry.

## 9. Workbook-surface registry

This registry is a domain-facing mirror of the current-profile standardized workbook-surface identity set. Core 01 owns the authoritative cross-layer workbook-surface mapping. This section MUST NOT define or restate exhaustive field registries, write targets, route shapes, projection-table names, or per-field defaults.

### 9.1 Surface axes

| Axis | Allowed values in this document | Meaning | Omission behavior |
| --- | --- | --- | --- |
| `surface_kind` | `built_in_sheet`, `system_view` | Identifies whether the surface is one of the five primary built-in sheets or a non-built-in contract-backed workbook surface. | Missing value in a registry row is a defect. |
| `surface_status` | `required built-in sheet`, `required system view`, `standardized optional workbook surface` | Identifies current-profile requiredness or optional standardization. | Missing value in a registry row is a defect. |
| `applicability` | Values from §1.2 | Identifies profile and current-status class. | Missing value in a registry row is a defect. |

`surface_kind='system_view'` MUST NOT be interpreted as requiredness. Requiredness is determined only by `surface_status`.

### 9.2 Surface identity mapping

Declared scope: every current-profile standardized workbook surface named by the Core 01 cross-layer workbook-surface mapping. Completion rule: this table is exhaustive when it contains the fourteen required Base Profile surfaces plus the three standardized optional artifact-backed surfaces and no additional current-profile surface.

| Surface | `view_schema_id` | `surface_kind` | `source_record_types` | `canonical_source_discriminator_or_filter` | `surface_status` | `required_reference_pack_keys` | `applicability` |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Timeline | `cartulary.view.timeline.v2` | `built_in_sheet` | `["timeline_event"]` | `record_type='timeline_event'` | required built-in sheet | `[]` | `base-required-surface` |
| Hosts | `cartulary.view.hosts.v1` | `built_in_sheet` | `["host"]` | `record_type='host'` | required built-in sheet | `[]` | `base-required-surface` |
| Identities | `cartulary.view.identities.v1` | `built_in_sheet` | `["identity"]` | `record_type='identity'` | required built-in sheet | `[]` | `base-required-surface` |
| Evidence | `cartulary.view.evidence.v1` | `built_in_sheet` | `["evidence"]` | `record_type='evidence'` | required built-in sheet | `[]` | `base-required-surface` |
| Notes | `cartulary.view.notes.v1` | `built_in_sheet` | `["artifact"]` | `artifact_type='note'` | required built-in sheet | `[]` | `base-required-surface` |
| Indicators | `cartulary.view.indicators.v1` | `system_view` | `["indicator"]` | `record_type='indicator'` | required system view | `[]` | `base-required-surface` |
| Compromise Assessments | `cartulary.view.assessments.v1` | `system_view` | `["assessment"]` | `record_type='assessment'` | required system view | `[]` | `base-required-surface` |
| Task Requests | `cartulary.view.task_requests.v1` | `system_view` | `["task_request"]` | `record_type='task_request'` | required system view | `[]` | `base-required-surface` |
| Decisions | `cartulary.view.decisions.v1` | `system_view` | `["decision"]` | `record_type='decision'` | required system view | `[]` | `base-required-surface` |
| Parties | `cartulary.view.parties.v1` | `system_view` | `["party"]` | `record_type='party'` | required system view | `[]` | `base-required-surface` |
| Communications Log | `cartulary.view.comm_log.v1` | `system_view` | `["artifact"]` | `artifact_type='comm_log'` | required system view | `[]` | `base-required-surface` |
| Handoff | `cartulary.view.handoff.v1` | `system_view` | `["artifact"]` | `artifact_type='handoff'` | required system view | `[]` | `base-required-surface` |
| Status Review | `cartulary.view.status_review.v1` | `system_view` | `["artifact"]` | `artifact_type='status_review'` | required system view | `[]` | `base-required-surface` |
| Lesson | `cartulary.view.lesson.v1` | `system_view` | `["artifact"]` | `artifact_type='lesson'` | required system view | `[]` | `base-required-surface` |
| Findings | `cartulary.view.findings.v1` | `system_view` | `["artifact"]` | `artifact_type='finding'`; subtype dimension `finding.kind` | standardized optional workbook surface | `[]` | `standardized-optional-surface` |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | `system_view` | `["artifact"]` | `artifact_type='investigative_query'`; separately governed optional structured subtype | standardized optional workbook surface | `[]` | `standardized-optional-surface` |
| Forensic Keywords | `cartulary.view.forensic_keywords.v1` | `system_view` | `["artifact"]` | `artifact_type='forensic_keyword'`; separately governed optional structured subtype | standardized optional workbook surface | `[]` | `standardized-optional-surface` |

### 9.3 Surface interpretation rules

- Required built-in tabs MUST be treated as primary workbook surfaces, not as storage tables.
- Required system views MUST remain workbook-native surfaces, not separate application modules.
- Standardized optional workbook surfaces MAY be exposed only when implemented under their owner contracts. Omission behavior: absence of a standardized optional surface does not make the Base Profile incomplete unless the implementation claims the corresponding optional owner-defined behavior.
- A saved view over any `view_schema_id` is additive and non-canonical.
- Pack-dependent framework overlays such as ATT&CK, D3FEND, and VERIS MUST NOT be described as base-profile workbook surfaces.
- Display labels MAY change without changing `view_schema_id` when field semantics do not change. Omission behavior: when a display-label change is not recorded here, the stable `view_schema_id` remains the domain identity.

## 10. Bounded contexts

Bounded contexts describe domain responsibility and language boundaries. They MUST NOT be treated as mandatory package names, deployable names, database schemas, route namespaces, or UI navigation labels.

Declared scope: all repository-wide domain contexts required to classify current-profile vocabulary in this document. Completion rule: every glossary row in §11 maps to one bounded context or to an external/future-only boundary row.

A bounded context in this document MUST have exactly one `context_kind`.

| `context_kind` | Meaning | Use when |
| --- | --- | --- |
| `domain_context` | Primary Cartulary problem-domain language boundary. | The context defines incident-response, investigation, evidence, entity, or coordination concepts. |
| `interaction_context` | Product-interaction language boundary. | The context defines workbook, surface, grid, saved-view, or inspector interaction vocabulary. |
| `supporting_context` | Supporting capability with domain-facing terms. | The context supports domain work but is not itself the primary problem-domain language. |
| `generic_context` | Conventional security, administration, backup, deployment, or platform language with Cartulary-specific constraints. | The context mostly uses industry-standard terms and this document only prevents leakage into domain language. |
| `external_boundary` | External model admitted only through anti-corruption mapping. | The context is not a Cartulary source model and must be translated through §16. |

| Bounded context | Context kind | Owns domain language for | Must not own | Language home | Behavior owner disposition | Tie-break rule | Applicability |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Incident Workspace | `domain_context` | Incident, incident key, incident metadata, membership, startup workbook surface, incident-scoped preferences. | Password/MFA semantics, deployment-wide users, object storage. | §11 glossary rows for Incident and Incident membership. | Use Core 00 §5.1 owner matrix by contract family. | When behavior crosses incident, membership, and workbook startup, inspect Core 01/Core 03/Core 04 owner rows before local prose. | `base` |
| Workbook Interaction | `interaction_context` | Built-in tabs, system views, saved views, grid action, paste, sort, filter, group, inspector, presence, save/conflict state. | Authoritative storage schema or physical projection topology. | §9 and §11 workbook-surface rows. | Core 03 for interaction behavior; Core 01 for view-schema/public interface behavior. | Stable identifiers and public contracts resolve through Core 01; interaction semantics resolve through Core 03. | `base` |
| Capture and Timeline | `domain_context` | Timeline events, rough capture, capture state, timeline grouping, timeline source text, unresolved tokens. | Canonical host/identity lifecycle except through mention and resolution workflows. | §13.1 through §13.3. | Core 02 for record semantics; Core 03 for interaction behavior. | Source-model disputes resolve through Core 02; interaction disputes resolve through Core 03. | `base` |
| Entities and Observations | `domain_context` | Hosts, identities, aliases, entity mentions, indicators, indicator observations, assessments, exact-match reuse, explicit merge. | Parties, deployment users, generalized contact management. | §12.2 and §13.2 through §13.4. | Core 02 primary; Core 03 for workbook interaction. | Model semantics resolve through Core 02. | `base` |
| Evidence | `domain_context` | Evidence records, object blobs, upload slot, attach flow, preview/download handle, custody and availability lifecycle. | Long-form report release state or raw telemetry storage. | §11 Evidence record and Object blob rows; §13.5. | Core 02 for evidence/blob model; Core 01 for object/blob routes and handles; Core 04 for access security. | Security/access disputes resolve through Core 04; source-model disputes resolve through Core 02. | `base` |
| Coordination | `domain_context` | Parties, task requests, decisions, communications logs, handoffs, status reviews, lessons, owner fields, follow-through references. | Generalized workflow engine or mandatory timeline approval path. | §13.6 through §13.8. | Core 02 for coordination records; Core 03 for workbook surfaces; Core 04 for authorization/release boundaries. | Routine coordination resolves through Core 02/Core 03; release gates resolve through Core 04. | `base` |
| Links and Tags | `domain_context` | Typed record relationships, record tags, relationship cells, collection-review fields. | Raw binary evidence storage or user account binding. | §12.2. | Core 02 for relationship semantics; Core 01 for public field/write contracts. | Relationship semantics resolve through Core 02. | `base` |
| Revisions and Audit | `supporting_context` | Change sets, mutation entries, record revisions, row history, rollback, merge history. | Live presence transport or deployment-local credential secrets. | §11 history rows; §13.10. | Core 02 for history substrate; Core 01 for public routes; Core 04 for administrative audit. | Incident record history resolves through Core 02; deployment-local admin audit resolves through Core 04. | `base` |
| Projections and Search | `supporting_context` | Workbook projections, view-query contract, filter/sort/group behavior, cursor semantics. | Authoritative source decisions or persisted incident facts except where owner sections define source state. | §11 Projection row; §14 defaults. | Core 01 primary; Core 03 for interaction behavior. | Public query and cursor semantics resolve through Core 01. | `base` |
| Reference Data | `supporting_context` | Reference packs, type registries, activation, attestation, optional overlays. | Incident record lifecycle or live capture eligibility. | §11 Reference pack row; §16 anti-corruption rows. | Core 01/Core 02/Core 04 by pack family. | Trust and activation disputes resolve through Core 04; data-model disputes through Core 02. | `extension-profile` |
| Reporting and Snapshots | `supporting_context` | Immutable snapshots, export models, release records, redaction profiles, report compositions, rendered outputs. | Live workbook write path, workbook artifacts, report templates, snapshots as mutable state, or incident-member visibility. | §11 Snapshot, Release, Recipient partition, and Report composition rows; §13.13. | Core 01 for snapshots/reporting; Core 04 for release gates and redaction; Report Composition NLSpec for composition authoring. | Release/authorization disputes resolve through Core 04; composition authoring disputes resolve through the Report Composition NLSpec. | `extension-profile` |
| Authentication and Administration | `generic_context` | Local users, current account profile, account preferences, sessions, MFA, deployment admin, credential lifecycle, enterprise-auth bindings, incident roles. | Incident-scoped party identity, incident-scoped workbook preferences, or workbook row mutation. | §11 User, Current account profile, Account preference, Incident membership, Deployment admin rows. | Core 04 primary; Core 01 for public auth/admin routes. | Security and authorization disputes resolve through Core 04. | `base` |
| Imports and Tabular Ingest | `supporting_context` | Import sessions, import units, source adapters, locator kinds, mapping fingerprints, warning codes, provenance. | Runtime workbook semantics outside the stable tabular-ingest contract. | §11 Import rows; §13.12. | Core 01 primary; Core 03 for paste and workbook interaction. | Import object semantics resolve through Core 01. | `extension-profile` |
| Network Flow Analysis | `supporting_context` | Network Flow analytical tables, immutable normalized flow rows, graph adaptation, contributor queries, and explicit indicator bindings. | Core record envelopes, `view_schema` resources, saved views, raw SIEM/EDR identity, live collection, or whole-incident removal. | §11 Network Flow rows; §13.12; §16. | Network Flow Activity NLSpec; Core 01–04 for imported interfaces, interaction, identity, and security. | Core owners govern shared interfaces; the adopted extension NLSpec governs only extension-local analytical semantics. | `extension-profile` |
| Backup and Restore | `generic_context` | Backup sets, restore verification, runtime roots, operator-facing recovery. | Workbook-surface route families or incident-scoped workflow. | §16 external boundary rows. | Core 01/Core 04 by backup or restore contract family. | Deployment/security disputes resolve through Core 04. | `base` |

### 10.1 Strategic subdomain classification

Declared scope: every bounded context in §10. Completion rule: every bounded context has exactly one subdomain class and one modeling-latitude rule.

A bounded context MAY be classified as `core` only when changing its language or model would materially change Cartulary's product identity, incident-response workflow fit, or correctness posture. A context that is necessary but replaceable, conventional, or mainly infrastructural MUST be classified as `supporting` or `generic`.

Every `core` row MUST satisfy this pressure test:

| Test question | Required answer for `core` |
| --- | --- |
| Product identity | The context names a concept that makes Cartulary recognizably Cartulary rather than a generic case, ticket, spreadsheet, or CRUD product. |
| Domain correctness | Misclassifying or weakening the context would cause incident-response, evidence, coordination, timeline, or entity-model drift. |
| Replacement risk | The context cannot be replaced by a conventional library, admin subsystem, import adapter, or storage mechanism without preserving Cartulary-specific vocabulary. |

| Bounded context | Subdomain class | Reason | Modeling latitude | Forbidden leakage |
| --- | --- | --- | --- | --- |
| Incident Workspace | `core` | Defines the incident-scoped workspace boundary that every domain concept depends on. | Preserve Cartulary-specific language. | Tenant, case-file, ticket, or project vocabulary cannot replace incident vocabulary. |
| Workbook Interaction | `core` | The workbook-first interaction model is a primary product differentiator. | Preserve workbook-surface and stable-identifier vocabulary. | Grid-vendor, React, SQL, or visible-label terms cannot define workbook behavior. |
| Capture and Timeline | `core` | Rough capture and progressive structuring are central to the domain. | Preserve capture-first language. | Forms-first or pre-normalization language cannot redefine capture. |
| Entities and Observations | `core` | Mention/entity/indicator separation prevents model drift. | Preserve observation versus canonical-entity distinctions. | IOC, contact, CMDB asset, or auth subject cannot collapse distinct objects. |
| Evidence | `core` | Evidence envelope versus blob separation is critical to IR correctness. | Preserve evidence lifecycle and object-blob distinctions. | File path, object key, screenshot, or raw blob cannot become evidence identity. |
| Coordination | `core` | Task, decision, party, and handoff semantics model incident teamwork. | Preserve bounded coordination objects. | General workflow-engine or approval-system language cannot replace coordination surfaces. |
| Links and Tags | `core` | Typed relationships and lightweight tags preserve recoverable graph semantics. | Preserve typed relationship vocabulary. | JSON arrays, visible chips, or note text cannot become relationship authority. |
| Revisions and Audit | `supporting` | Enables recoverability and accountability across the core domain. | Use domain-specific terms only at incident-history boundaries. | Generic audit-log terms cannot replace record revision or rollback semantics. |
| Projections and Search | `supporting` | Enables workbook views without becoming source state. | Use implementation latitude below the projection contract. | Projection table names cannot become source truth. |
| Reference Data | `supporting` | Optional vocabularies and enrichments support incident work. | Keep pack vocabulary external to incident source state. | Framework overlays cannot become Base Profile surfaces. |
| Reporting and Snapshots | `supporting` | Derives output artifacts and composition authoring inputs from incident state without changing live collaboration. | Keep release vocabulary artifact-scoped and composition vocabulary authoring-input scoped. | Export redaction cannot become live workspace authorization, and composition authoring cannot become source-record mutation. |
| Authentication and Administration | `generic` | Uses conventional account/session/admin concepts with Cartulary-specific boundaries. | Prefer conventional security terminology except where party/user/role separation matters. | User, party, identity, incident role, and deployment admin cannot collapse. |
| Imports and Tabular Ingest | `supporting` | Bridges spreadsheets and structured workbook state. | Keep XLSX/CSV semantics inside import boundaries. | Worksheet/table semantics cannot leak into runtime workbook identity. |
| Network Flow Analysis | `supporting` | Provides bounded incident-scoped analysis of normalized external flow data after the extension is adopted and claimed. | Keep analytical tables and rows outside the Core record/view model. | Raw telemetry, flow rows, graph nodes, and workspace tabs cannot become Base Profile source records or surfaces. |
| Backup and Restore | `generic` | Deployment-local recovery concern with domain safety boundaries. | Use conventional backup/restore terminology behind Core 04 boundaries. | Backup jobs cannot become incident-scoped workbook workflow. |

Review note: `Links and Tags` remains `core` in this revision because typed recoverable relationship semantics are treated as product-identity language. A later revision that demotes it MUST update §10.1, §11.1, and §19 in the same change set.

### 10.2 Strategic context map

Declared scope: current cross-context relationships that affect repository vocabulary, owner navigation, or anti-corruption boundaries. Completion rule: every bounded context in §10 appears in at least one relationship row or has an explicit `separate_ways` disposition.

| Upstream context | Downstream context | Relationship type | Published language | Translation owner | Allowed dependency | Forbidden dependency | Change obligation |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Incident Workspace | Workbook Interaction | `published_language` | `incident_id`, `sheet_ref`, incident preferences. | Core 01/Core 03 | Workbook startup and surface selection may depend on incident workspace identity. | Workbook UI MUST NOT redefine incident or membership semantics. | Any change to incident identity, membership, preference, or `sheet_ref` vocabulary MUST update the downstream workbook startup/surface-selection language or mark a breaking corpus change. |
| Workbook Interaction | Capture and Timeline | `customer_supplier` | `view_schema_id`, `field_key`, row mutation vocabulary. | Core 03/Core 01 | Timeline capture uses workbook interaction contracts. | Timeline source model MUST NOT be derived from visible columns alone. | Any change to `view_schema_id`, `field_key`, row mutation vocabulary, or paste semantics MUST update Timeline vocabulary before capture-language text changes. |
| Capture and Timeline | Entities and Observations | `published_language` | `entity_mention`, `mention_origin`, host/identity refs. | Core 02 | Timeline observations can resolve to canonical host/identity records through owner-defined actions. | Mentions MUST NOT auto-create entities outside owner-defined actions. | Any change to mention or resolution vocabulary MUST update entity/observation terminology and anti-auto-create guardrails in the same change set. |
| Entities and Observations | Workbook Interaction | `published_language` | host, identity, indicator, assessment projections. | Core 02/Core 03 | Workbook views may expose entity and observation state. | Grid labels MUST NOT redefine entity identity or dedupe. | Any change to host, identity, indicator, assessment, or dedupe language MUST update workbook-facing labels and glossary mappings without changing stable identifiers by label. |
| Evidence | Workbook Interaction | `published_language` | evidence record, object blob, preview/download affordance. | Core 01/Core 02/Core 04 | Workbook surfaces may show evidence state and handles. | Object-store keys MUST NOT enter workbook identity. | Any change to evidence/blob/access-handle vocabulary MUST update workbook evidence affordance terminology and §16 object-storage anti-corruption rows. |
| Coordination | Workbook Interaction | `published_language` | party, task_request, decision, comm_log, handoff, status_review, lesson. | Core 02/Core 03 | Coordination surfaces remain workbook-native. | Coordination MUST NOT become a separate workflow engine outside workbook grammar. | Any change to party, task, decision, or coordination-artifact vocabulary MUST update workbook-surface terminology and §11.1 mapping in the same change set. |
| Links and Tags | Workbook Interaction | `shared_kernel` | record link, tag, collection item families. | Core 02/Core 01 | Workbook fields and inspector actions may emit typed relationship mutations. | Relationship semantics MUST NOT be inferred from visible chips or text. | Any change to shared relationship/tag vocabulary MUST update both relationship semantics and workbook/inspector terminology together. |
| Revisions and Audit | Workbook Interaction | `published_language` | change_set, mutation entry, record revision, rollback. | Core 02/Core 03 | Workbook history UI may expose row-history and rollback vocabulary. | Live presence and save state MUST NOT become history authority. | Any change to history, rollback, change-set, or revision vocabulary MUST update workbook history terminology without converting live presence/save state into audit state. |
| Projections and Search | Workbook Interaction | `published_language` | projection row, sort/filter/group/cursor vocabulary. | Core 01/Core 03 | Workbook query consumes derived projections. | Projection state MUST NOT become source state. | Any change to projection/query/cursor vocabulary MUST update workbook query language and must not promote projection state to source state. |
| Reference Data | Entities and Observations | `anti_corruption_layer` | reference pack, type registry, enrichment metadata. | Core 01/Core 02/Core 04 | Reference data may enrich incident records through owner-defined fields. | Pack terminology MUST NOT block capture or become incident source identity. | Any change in external pack vocabulary MUST update the translation row before incident-domain terms may use it. |
| Reporting and Snapshots | Incident Workspace | `published_language` | snapshot, release, recipient partition, report composition. | Core 01/Core 04/Report Composition NLSpec/Reporting Subsystem NLSpec | Snapshot/reporting may derive immutable outputs and composition authoring inputs from incident state. | Release redaction MUST NOT hide live workspace content, and composition authoring MUST NOT mutate source records. | Any change to snapshot, release, redaction, recipient, or composition vocabulary MUST preserve the live-workspace/export boundary or identify a breaking corpus change. |
| Authentication and Administration | Incident Workspace | `published_language` | user, session, incident role, deployment admin. | Core 04/Core 01 | Incident routes use current membership and role checks. | Auth provider identity MUST NOT become party or incident identity. | Any change to user/session/role/deployment-admin vocabulary MUST update party/user/identity distinctions and Core 04 owner references. |
| Imports and Tabular Ingest | Workbook Interaction | `anti_corruption_layer` | import_session, import_unit, locator kind, mapping fingerprint. | Core 01/Core 03 | Import may compile source columns into stable field-key mappings. | Worksheet/table semantics MUST NOT become runtime workbook semantics. | Any change to import-source locator language MUST update spreadsheet anti-corruption rows and must not redefine runtime workbook surface identity. |
| Imports and Tabular Ingest | Network Flow Analysis | `anti_corruption_layer` | import session/unit identity, approved mapping, opaque source capability, extension target result. | Core 01; Network Flow Activity NLSpec | Imports owns upload and orchestration; the extension owns normalized flow validation and analytical-table publication. | Import units, filenames, worksheet locators, and parser DTOs MUST NOT become Network Flow table identity. | Any extension-target change MUST update both import vocabulary and Network Flow target/result vocabulary before implementation. |
| Network Flow Analysis | Workbook Interaction | `published_language` | extension workspace identity, active table label, invalidation, current authorization. | Core 01/Core 03; Network Flow Activity NLSpec | A claimed extension may contribute Network Analysis without expanding the Base `view_schema_id` registry. | Visible labels, React components, inner tabs, and routes MUST NOT define extension resource identity. | Any workspace-identity change MUST preserve the extension/Base surface boundary and owner-defined `sheet_ref` vocabulary. |
| Network Flow Analysis | Entities and Observations | `customer_supplier` | canonical IP indicator identity and explicit binding. | Core 02; Network Flow Activity NLSpec | The extension may request owner-defined canonical indicator find/create and bind an endpoint explicitly. | A flow row MUST NOT become an indicator observation, host, identity, or canonical indicator automatically. | Any binding change MUST update canonical IP identity and transaction-boundary language together. |
| Backup and Restore | Incident Workspace | `separate_ways` | deployment-local recovery vocabulary. | Core 01/Core 04 | Recovery may preserve or restore authoritative incident state under owner rules. | Backup/restore MUST NOT appear as workbook route families or incident workflow. | Any proposed dependency from workbook or coordination vocabulary to recovery vocabulary MUST be rejected unless an owner section creates an explicit operational effect. |

### 10.3 Context relationship type vocabulary

Declared scope: relationship-type tokens that may appear in §10.2 or be recognized during domain review. Completion rule: every relationship type is classified exactly once, and only rows with `status='current-used'` may appear in §10.2.

| Relationship type | Status | Meaning in this document | Omission or use behavior |
| --- | --- | --- | --- |
| `published_language` | `current-used` | Upstream context exposes stable concepts that downstream context may use without importing upstream implementation internals. | Valid in §10.2 only with named published language and translation owner. |
| `anti_corruption_layer` | `current-used` | Downstream context may consume external or adjacent concepts only through a named translation boundary. | Valid in §10.2 only when §16 or an owner-defined integration boundary maps the external language. |
| `customer_supplier` | `current-used` | Downstream context depends on upstream concepts and upstream changes MUST preserve downstream vocabulary compatibility or identify a breaking corpus change. | Valid in §10.2 only with a change obligation. |
| `shared_kernel` | `current-used` | Contexts share a narrow owner-defined model. Changes require owner-section coordination. | Valid in §10.2 only when the shared model is explicitly named. |
| `separate_ways` | `current-used` | Contexts MUST remain conceptually separate except for explicit owner-defined operational effects. | Valid in §10.2 as an explicit no-integration disposition. |
| `external_adapter` | `current-used` | External system vocabulary enters only through §16 translation rows. | Valid only for external-system rows or future integration-owner text. |
| `partnership` | `recognized-unused` | Two contexts mutually evolve a model or language. | MUST NOT appear in §10.2 unless this table is revised to `current-used` and the row names mutual change obligations. |
| `conformist` | `recognized-unused` | Downstream adopts an upstream model without translation. | MUST NOT be implied by consuming an external API or reference pack. Use `anti_corruption_layer` unless an owner section explicitly adopts conformist behavior. |
| `open_host_service` | `recognized-unused` | Upstream exposes a stable service protocol for downstream consumers. | MUST NOT appear in §10.2 in this revision. Use Core-owned public route or interface terminology instead. |
| `big_ball_of_mud` | `recognized-diagnostic-only` | Boundary failure label for uncontrolled model mixing. | MUST NOT be a target relationship. MAY appear only in review findings or rationale that identify a defect. |

### 10.4 Context stewardship and review triggers

Context stewardship is documentation and review responsibility only. It MUST NOT create runtime authorization, team ownership, deployable ownership, package ownership, incident membership, or module ownership.

| Bounded context | Vocabulary steward | Change-review trigger |
| --- | --- | --- |
| Incident Workspace | Owner sections named by the row in §10. | Any change to incident, membership, preference, startup, or workspace-boundary language. |
| Workbook Interaction | Core 01/Core 03 owner sections. | Any change to workbook surface, `view_schema_id`, `sheet_ref`, saved-view, grid, inspector, or interaction-state language. |
| Capture and Timeline | Core 02/Core 03 owner sections. | Any change to rough capture, timeline event, source-text, mention-origin, or timeline grouping language. |
| Entities and Observations | Core 02 owner sections. | Any change to host, identity, indicator, observation, assessment, mention, dedupe, auto-resolution, or merge language. |
| Evidence | Core 01/Core 02/Core 04 owner sections. | Any change to evidence record, object blob, upload, custody, preview, download, availability, or access-language boundaries. |
| Coordination | Core 02/Core 03/Core 04 owner sections. | Any change to party, task request, decision, communications log, handoff, status review, lesson, owner, or follow-through vocabulary. |
| Links and Tags | Core 01/Core 02 owner sections. | Any change to typed relationship, tag, record-link, or collection-review vocabulary. |
| Revisions and Audit | Core 01/Core 02/Core 04 owner sections. | Any change to change set, mutation entry, record revision, rollback, merge history, or administrative audit language. |
| Projections and Search | Core 01/Core 03 owner sections. | Any change to projection, query, sort, filter, grouping, search, cursor, or row-refresh language. |
| Reference Data | Core 01/Core 02/Core 04 owner sections. | Any change to reference-pack, type-registry, activation, attestation, optional overlay, or enrichment vocabulary. |
| Reporting and Snapshots | Core 01/Core 04 owner sections, Report Composition NLSpec, and Reporting Subsystem NLSpec. | Any change to snapshot, export model, release, redaction profile, rendered output, recipient-partition vocabulary, or report-composition vocabulary. |
| Authentication and Administration | Core 01/Core 04 owner sections. | Any change to user, session, credential, account profile, account preference, deployment admin, or enterprise-auth language. |
| Imports and Tabular Ingest | Core 01/Core 03 owner sections. | Any change to import session, import unit, locator kind, mapping fingerprint, warning code, or tabular-ingest vocabulary. |
| Network Flow Analysis | Core 00 profile posture; Core 01–04 shared interfaces; Network Flow Activity NLSpec. | Any change to Network Flow table/row identity, extension workspace, import target/result, graph boundary, indicator binding, authorization, or retention vocabulary. |
| Backup and Restore | Core 01/Core 04 owner sections. | Any change to backup set, restore, verification, recovery, runtime root, or operator-facing recovery language. |

## 11. Core domain glossary

Declared scope: repository-wide terms whose misunderstanding would cause behavior, contract, or review drift. Completion rule: every row has a definition, forbidden interpretation, owner, applicability, and current-profile status.

| Term | Definition | Not this | Canonical identifiers or tokens | Language owner | Behavior owner | Applicability | Current-profile status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Incident | Incident-scoped workspace for live investigation, coordination, source records, and workbook surfaces. | Tenant, deployment, project, ticket, or case file. | `incident_id`, `incident_key` | `docs/domain.md` §11 | Core 01/Core 02/Core 04 by contract family | `base` | `current-required` |
| Record envelope | Common identity, attribution, version, and delete-state envelope for user-visible first-class incident records. | SQL inheritance mechanism or generic JSON object. | `record_id`, `record_type`, `row_version` | §11 | Core 02 §3 | `base` | `current-required` |
| Timeline event | Primary rough-capture and chronology record. | Task, finding, note, or raw telemetry event store. | `record_type='timeline_event'`, Timeline view | §11 | Core 02/Core 03 | `base` | `current-required` |
| Host | Canonical or stub device or host record in one incident. | CMDB asset, object-store host, or deployment host. | `record_type='host'` | §11 | Core 02 | `base` | `current-required` |
| Identity | Canonical or stub account or persona record in one incident. | Party, user account, auth-provider identity, requester, or attendee. | `record_type='identity'` | §11 | Core 02 | `base` | `current-required` |
| Party | Incident-scoped person, team, organization, distribution list, or other stakeholder record used for stable coordination identity. | Deployment user, auth identity, incident role, or canonical investigation identity. | `record_type='party'`, `party_id` | §11 | Core 02 §19 | `base` | `current-required` |
| User | Deployment-local login and attribution identity. | Party, identity, incident member, or contact. | `user_id` | §11 | Core 04/Core 01 | `base` | `current-required` |
| Current account profile | Current authenticated user's self-service profile projection containing read-only email and editable display name. | Deployment-user administration, self-service login-identifier change, party identity, or incident-scoped preference. | `/api/v1/account/profile`, `display_name`, `user_version` | §11 | Core 01 §3.3.2.3; Core 04 §2 | `base` | `current-required` |
| Account preference | Deployment-local current-user preference resource whose current profile content is only nullable density override. | Saved view, incident workbook preference, global home/default incident setting, locale, time zone, notifications, theme selector, custom row height, or portability content. | `/api/v1/account/preferences`, `density_mode`, `preferences_version` | §11 | Core 01 §3.3.2.3; Core 02 §14.1; Core 03 §2.4 | `base` | `current-required` |
| Incident membership | User-to-incident authorization relationship carrying an incident role. | Party participation, task owner, or identity relationship. | `incident_id`, `user_id`, role token family | §11 | Core 04 | `base` | `current-required` |
| Deployment admin | Current-profile application-level deployment administration capability whose exact authorization matrix is owned by Core 04 REQ-04-028. | Incident admin, incident member, party, user type, or incident-content access by itself. | `deployment_admin` capability | §11 | Core 04 | `base` | `current-required` |
| Artifact | Structured analyst text object family, including notes and structured coordination/finding material where owner sections define the subtype. | Binary file, object blob, or Notes surface alone. | `record_type='artifact'`, `artifact_type` | §11 | Core 02 | `base` | `current-required` |
| Note | Artifact-backed built-in workbook surface row with `artifact_type='note'`. | Entire artifact family or arbitrary scratchpad outside records. | `cartulary.view.notes.v1`, `artifact_type='note'` | §11 | Core 01/Core 02/Core 03 | `base-required-surface` | `current-required` |
| Task request | Owned unit of work or request with lifecycle state and queue visibility. | Timeline checkbox, tag, generalized ticket state, or row approval. | `record_type='task_request'` | §11 | Core 02/Core 03 | `base-required-surface` | `current-required` |
| Decision | Owned incident-scoped rationale-bearing decision record. | Release approval, informal note, or workflow engine step. | `record_type='decision'` | §11 | Core 02/Core 03 | `base-required-surface` | `current-required` |
| Communications Log | Artifact-backed coordination record for durable communication memory. | Chat archive, raw email store, or release approval. | `artifact_type='comm_log'`, `cartulary.view.comm_log.v1` | §11 | Core 02/Core 03 | `base-required-surface` | `current-required` |
| Handoff | Artifact-backed coordination record for shift or responsibility transfer. | Ordinary row edit, chat message, or task assignment alone. | `artifact_type='handoff'`, `cartulary.view.handoff.v1` | §11 | Core 02/Core 03 | `base-required-surface` | `current-required` |
| Status Review | Artifact-backed coordination record for incident status checkpoints. | Dashboard, live workbook filter, or mandatory per-edit ritual. | `artifact_type='status_review'`, `cartulary.view.status_review.v1` | §11 | Core 02/Core 03 | `base-required-surface` | `current-required` |
| Lesson | Artifact-backed coordination record for debrief and improvement follow-through. | Training guide, issue tracker, or automatic postmortem. | `artifact_type='lesson'`, `cartulary.view.lesson.v1` | §11 | Core 02/Core 03 | `base-required-surface` | `current-required` |
| Finding | Structured artifact subtype for finding or hypothesis rows when the optional surface is implemented. | First-class `record_type='finding'` or final report. | `artifact_type='finding'`, `finding.kind` | §11 | Core 02/Core 01 | `standardized-optional-surface` | `current-optional-when-implemented` |
| Entity mention | Source-bound textual reference to a host or identity candidate. | Weak entity, alias, or auto-created canonical record. | `entity_mention_id`, `resolution_status` token family | §11 | Core 02 | `base` | `current-required` |
| Stub entity | Real host or identity record created with incomplete information. | Unresolved mention or suggestion. | `record_id`, `entity_origin` token family | §11 | Core 02 | `base` | `current-required` |
| Indicator | Canonical incident-scoped linkable value or pattern. | Raw IOC-like string occurrence. | `record_type='indicator'`, `indicator.value_kind` token family | §11 | Core 02 | `base-required-surface` | `current-required` |
| Indicator observation | Source-bound occurrence of an indicator value or pattern inside another record field. | Canonical indicator or lifecycle interval. | observation source record and field, `origin_kind` token family | §11 | Core 02 | `base` | `current-required` |
| Indicator lifecycle interval | Append-only state window attached to a canonical indicator. | Observation time or overwrite of indicator state. | lifecycle interval row | §11 | Core 02 | `base` | `current-required` |
| Compromise assessment | Incident-scoped assessment record about a host or identity. | Field on host or identity, or overwritten current verdict. | `record_type='assessment'`, `assessment_state` token family | §11 | Core 02 | `base-required-surface` | `current-required` |
| Evidence record | User-facing evidence envelope that models requested, pending, received, available, quarantined, or released evidence. | Raw blob, file path, screenshot pasted into notes. | `record_type='evidence'`, evidence lifecycle token family | §11 | Core 02 | `base-required-surface` | `current-required` |
| Object blob | Storage metadata and upload state for binary content. | Evidence row, workbook attachment count, or object-store key contract. | `object_blob_id`, upload-state token family | §11 | Core 01/Core 02 | `base` | `current-required` |
| Record link | Typed relationship between two records. | JSON list, text mention, or visible chip alone. | `record_links.link_type` token family | §11 | Core 02 | `base` | `current-required` |
| Tag | Lightweight incident-scoped label. | Owner, task, lifecycle state, or relationship. | `tag_id`, `record_tags` | §11 | Core 02 | `base` | `current-required` |
| View schema | Contract for a built-in sheet, system view, or standardized optional workbook surface. | Physical table schema or saved view. | `view_schema_id` | §11 | Core 01 | `base` | `current-required` |
| Inspector | Explicit row-context secondary surface derived from active `view_schema_id` inspector config. | Workflow engine, saved-view state, full-page editor, dashboard, ticketing module, or hidden source state. | `inspector_config_v1`, `feature_group_key`, `route_binding.owner` | §11 | Core 01/Core 03/Core 04 | `base` | `current-required` |
| Feature group key | Stable key for one declared inspector feature group within one `view_schema_id`. | Visible label, route helper, React component, CSS selector, storage column, or grid-vendor identifier. | `feature_group_key` | §11 | Core 01 §7.4 | `base` | `current-required` |
| Route binding owner | Closed token identifying the Core-owned route family or current-row data source used by a feature group. | Implementation function name, handler name, SQL table, component name, or route-helper name. | `route_binding.owner` | §11 | Core 01 §7.4 | `base` | `current-required` |
| Workbook surface | Visible or addressable sheet-like working surface backed by a `view_schema_id` or distinct saved view over one schema. | Module, route family, projection table, or visible label. | `sheet_ref` | §11 | Core 01/Core 03 | `base` | `current-required` |
| Built-in tab | Required primary sheet-like workbook surface in the base profile. | Every required surface or every visible shell tab. | Timeline, Hosts, Identities, Evidence, Notes | §11 | Core 03 | `base-required-surface` | `current-required` |
| System view | Non-built-in contract-backed workbook surface with `surface_kind='system_view'`. | `scope='system'` saved view or requiredness by itself. | `surface_kind='system_view'`, `view_schema_id` | §11 | Core 01/Core 03 | `base` | `current-required` |
| Required system view | System view with `surface_status='required system view'`. | All rows whose `surface_kind` is `system_view`. | required system `view_schema_id` values in §9.2 | §11 | Core 01/Core 03 | `base-required-surface` | `current-required` |
| Saved view | Incident-bound workbook configuration over exactly one immutable `view_schema_id`. | Required system view or projection row. | `saved_view_id`, `scope` token family | §11 | Core 03 | `base` | `current-required` |
| Projection | Denormalized workbook read model used for workbook query, sorting, filtering, grouping, and row refresh. | Source of truth, history substrate, or graph-projection result. | projection row with `record_id` and `row_version` | §11 | Core 01/Core 02 | `base` | `current-required` |
| Change set | Immutable attribution unit for one committed action. | UI action only or untracked transaction. | `change_set_id`, actor, timestamp, source | §11 | Core 02 | `base` | `current-required` |
| Mutation entry | Reversible target-level delta within a change set. | Human-facing history summary only. | target kind, target identifier, operation kind | §11 | Core 02 | `base` | `current-required` |
| Record revision | Row-centric historical state used for review and rollback. | Projection snapshot or audit log only. | revision number, history entry reference where eligible | §11 | Core 02/Core 01 | `base` | `current-required` |
| Rollback | Scoped historical corrective action that appends history. | Silent overwrite or delete-and-recreate. | rollback route and `history_entry_ref` where eligible | §11 | Core 01/Core 02 | `base` | `current-required` |
| Reference pack | Versioned optional vocabulary, framework, type registry, template, or enrichment dataset. | Incident record, live workbook surface, or required capture dependency. | `pack_key`, `pack_version`, activation metadata | §11 | Core 01/Core 02/Core 04 | `extension-profile` | `current-extension-when-claimed` |
| Import session | One uploaded source file plus one operator-driven import workflow. | Whole-workbook runtime behavior. | `import_session_id` | §11 | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| Import unit | One candidate ingestable unit discovered from an import source. | Worksheet identity or table identity. | `import_unit`, `locator_kind`, `mapping_fingerprint` | §11 | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| Import apply dispatcher | Import-owned internal dispatcher that applies approved import units only through registry-selected owner create facades. | Workbook store, parser, grid adapter, or owner source module. | `import_apply_dispatcher_v1` | §11 | Core 01/Core 03 | `extension-profile` | `current-extension-when-claimed` |
| Import target registry | Registry that declares which `view_schema_id` values are importable and which owner create facade handles each target. | View-schema registry replacement or visible worksheet list. | `target_view_schema_id`, `import_apply_status` | §11 | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| Network Flow table | Incident-scoped analytical resource containing one atomically published set of normalized immutable flow rows while the Network Flow Activity profile is claimed. | Core record, view schema, saved view, import unit, worksheet, filename, raw telemetry stream, or physical SQL table. | `network_flow_table_id` | §11 | Network Flow Activity NLSpec; Core 01 for import/public reference interfaces | `extension-profile` | `current` |
| Network Flow row | Immutable normalized analytical row owned by one Network Flow table. | `view_row_v1`, Core record revision, indicator observation, graph node, raw SIEM/EDR event, or mutable spreadsheet row. | table-scoped Network Flow row identity | §11 | Network Flow Activity NLSpec | `extension-profile` | `current` |
| Network Analysis workspace | Claimed extension workspace for working with active Network Flow tables and derived graph/query results. | Base built-in tab, system view, saved view, `view_schema_id`, full application module, or identity inferred from the `Network Analysis` label. | owner-adopted extension workspace identity and `sheet_ref` variant | §11 | Core 01/Core 03; Network Flow Activity NLSpec | `extension-profile` | `current` |
| Network Flow indicator binding | Explicit incident-scoped relationship between one normalized IP endpoint and one canonical family-matching indicator, committed through the owner transaction boundary. | Indicator observation, automatic host/identity creation, record link inferred from graph adjacency, or cross-incident identity. | Network Flow binding identity plus canonical indicator `record_id` | §11 | Core 02; Network Flow Activity NLSpec | `extension-profile` | `current` |
| Owner create facade | Source-owner import creation boundary that consumes normalized field-keyed row plans and returns authoritative record and row refresh results. | Public API route, parser DTO, workbook row store, or grid row adapter. | `import_owner_create_request_v1`, `import_owner_create_response_v1` | §11 | Core 01/Core 02/Core 03 | `extension-profile` | `current-extension-when-claimed` |
| Snapshot | Immutable incident export-model anchor when the Snapshot and Reporting Extension Profile is implemented. | Live workbook state or saved view. | `snapshot_id`, `snapshot_at` | §11 | Core 01 | `extension-profile` | `current-extension-when-claimed` |
| Release | Artifact-scoped rendered-output approval and publication record. | General row approval workflow. | `release_id`, release-state token family | §11 | Core 04/Core 01 | `extension-profile` | `current-extension-when-claimed` |
| Recipient partition | Export-only disclosure boundary selected through release-time `recipient_partition_refs[]`. | Live workbook authorization or row visibility. | `disclosure_partition_refs[]`, `recipient_partition_refs[]` | §11 | Core 01/Core 02/Core 04 | `extension-profile` | `current-extension-when-claimed` |
| Report composition | Incident-scoped authoring input that carries presentation operations, authored presentation text, composition diagram declarations, and optional closed diagram layout data for one report template version. | Report, report template, workbook artifact, snapshot, release output, live workbook record, or generated source file. | `composition_id`, `cartulary.report_composition.v1` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for consumption | `extension-profile` | `current-extension-when-claimed` |
| Report composition resource | Mutable incident-scoped authoring resource with one active draft, fixed template binding, and zero or more immutable digest-bound versions. | Report record, report template, workbook artifact, snapshot, release output, generated source file, or cross-template authoring bucket. | `composition_id`, `/report-compositions` | §11 | Report Composition NLSpec | `extension-profile` | `current-extension-when-claimed` |
| Composition draft | Mutable server-versioned body of a report composition resource before it is frozen as immutable canonical composition bytes. | Release candidate, approval evidence, immutable version, workbook row version, or browser-local editor state. | `draft_version`, `base_draft_version` | §11 | Report Composition NLSpec | `extension-profile` | `current-extension-when-claimed` |
| Composition version | Immutable digest-bound version of one report composition resource. | Report revision, template version, workbook row version, snapshot version, release output, or mutable draft. | `composition_version`, `composition_sha256` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for release tuple consumption | `extension-profile` | `current-extension-when-claimed` |
| Composition operation | Closed authoring instruction that requests a presentation effect through semantic anchors and typed payloads. | Workbook mutation, template edit, snapshot edit, report-byte edit, raw Markdown patch, or raw Mermaid patch. | `composition_op.v1`, `op_kind` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for render effects | `extension-profile` | `current-extension-when-claimed` |
| Semantic composition anchor | Stable authoring target for a section, record, block, or diagram declaration. | Row number, visible label, generated slide ID, CSS selector, DOM node, report output coordinate, or workbook artifact reference. | `section_anchor`, `record_anchor`, `block_anchor`, `diagram_anchor` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for render-time resolution | `extension-profile` | `current-extension-when-claimed` |
| Authored presentation text | Composition-authored text admitted only as presentation text under explicit disclosure partition and redaction rules. | Source fact, finding, causal narrative, evidence assertion, workbook artifact, report template text, snapshot content, or release output. | `authored_text.v1`, `text_role` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for redaction and admission | `extension-profile` | `current-extension-when-claimed` |
| Composition diagram declaration | Composition-owned diagram declaration that uses Reporting selection rules, presentation labels, and optional closed layout data without raw Mermaid or graph mutation. | Raw diagram source, arbitrary graph node, arbitrary graph edge, template pack mutation, workbook artifact, snapshot, release output, generated Mermaid ID, DOM ID, SVG ID, or UI-library coordinate state. | `composition_diagram_decl.v1`, `composition_diagram_layout.v1`, `decl_id` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for diagram materialization and rendering effects | `extension-profile` | `current-extension-when-claimed` |
| Composition diagram layout | Closed presentation layout data for exact node placement and manual routes over retained selected diagram items. | Graph projection output, graph mutation, rendered SVG edit, Mermaid source edit, DOM coordinate state, React Flow state, workbook artifact, snapshot, or release output. | `composition_diagram_layout.v1`, `layout_mode` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for deterministic rendering effects | `extension-profile` | `current-extension-when-claimed` |
| Composition preview | Authoritative internal-draft render attempt admitted through the composition preview route and Reporting preview boundary. | Release evidence, approval evidence, release tuple binding, browser-local preview, generated source file, or immutable release bytes. | `cartulary.report_composition_preview_source.v1`, `preview_attempt_id`, `render_attempt_id` | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec for render attempt ownership | `extension-profile` | `current-extension-when-claimed` |
| Builder boundary | The report builder is a client authoring surface whose durable contract is server route requests, generated schemas, server validation, and authoritative preview admission. | React component contract, UI library state, local lint authority, render engine, release approver, template editor, or conformance owner. | generated schemas, validation route, preview route | §11 | Report Composition NLSpec; Reporting Subsystem NLSpec; Core 01/Core 04 for route envelope and authorization | `extension-profile` | `current-extension-when-claimed` |
| Report builder | Authoring UI surface that creates, validates, versions, and previews report composition data. | Workbook editor, template editor, report renderer, post-redaction editor, release approval surface, or owner of generated report bytes. | `/api/v1/incidents/{incident_id}/report-compositions` | §11 | Report Composition NLSpec; Core 01/Core 04 for envelopes and authorization | `extension-profile` | `current-extension-when-claimed` |

### 11.1 Term-to-bounded-context map

Declared scope: every term row in §11. Completion rule: each §11 term appears exactly once in this table and maps to one bounded context from §10 or to an explicit external/future-only boundary. This table does not change behavior ownership; behavior remains owned by the `Behavior owner` column in §11.

| Term | Bounded context | Context role |
| --- | --- | --- |
| Incident | Incident Workspace | Primary language. |
| Record envelope | Incident Workspace | Shared incident-record foundation. |
| Timeline event | Capture and Timeline | Primary language. |
| Host | Entities and Observations | Primary language. |
| Identity | Entities and Observations | Primary language. |
| Party | Coordination | Primary language. |
| User | Authentication and Administration | Primary language. |
| Current account profile | Authentication and Administration | Primary language. |
| Account preference | Authentication and Administration | Primary language. |
| Incident membership | Authentication and Administration | Primary authorization-language boundary. |
| Deployment admin | Authentication and Administration | Primary administration-language boundary. |
| Artifact | Coordination | Primary artifact-backed coordination and analyst-material language. |
| Note | Workbook Interaction | Workbook-surface language over artifact-backed source state. |
| Task request | Coordination | Primary language. |
| Decision | Coordination | Primary language. |
| Communications Log | Coordination | Primary language. |
| Handoff | Coordination | Primary language. |
| Status Review | Coordination | Primary language. |
| Lesson | Coordination | Primary language. |
| Finding | Coordination | Standardized optional artifact-backed language. |
| Entity mention | Entities and Observations | Primary language. |
| Stub entity | Entities and Observations | Primary language. |
| Indicator | Entities and Observations | Primary language. |
| Indicator observation | Entities and Observations | Primary language. |
| Indicator lifecycle interval | Entities and Observations | Primary language. |
| Compromise assessment | Entities and Observations | Primary language. |
| Evidence record | Evidence | Primary language. |
| Object blob | Evidence | Primary language. |
| Record link | Links and Tags | Primary language. |
| Tag | Links and Tags | Primary language. |
| View schema | Workbook Interaction | Stable workbook contract language. |
| Inspector | Workbook Interaction | Row-context interaction language. |
| Feature group key | Workbook Interaction | Stable inspector contract language. |
| Route binding owner | Workbook Interaction | Stable inspector route-binding vocabulary. |
| Workbook surface | Workbook Interaction | Primary interaction language. |
| Built-in tab | Workbook Interaction | Primary interaction language. |
| System view | Workbook Interaction | Primary interaction language. |
| Required system view | Workbook Interaction | Primary interaction language. |
| Saved view | Workbook Interaction | Primary interaction language. |
| Projection | Projections and Search | Primary supporting language. |
| Change set | Revisions and Audit | Primary language. |
| Mutation entry | Revisions and Audit | Primary language. |
| Record revision | Revisions and Audit | Primary language. |
| Rollback | Revisions and Audit | Primary language. |
| Reference pack | Reference Data | Extension-profile language. |
| Import session | Imports and Tabular Ingest | Extension-profile language. |
| Import unit | Imports and Tabular Ingest | Extension-profile language. |
| Import apply dispatcher | Imports and Tabular Ingest | Internal apply-boundary language. |
| Import target registry | Imports and Tabular Ingest | Registry and target-ownership language. |
| Network Flow table | Network Flow Analysis | Extension-profile analytical-resource language; not current until adoption. |
| Network Flow row | Network Flow Analysis | Extension-profile normalized-row language; not current until adoption. |
| Network Analysis workspace | Network Flow Analysis | Extension-workspace language distinct from the Base surface registry; not current until adoption. |
| Network Flow indicator binding | Network Flow Analysis | Explicit binding language delegated to Core 02 identity and transaction owners; not current until adoption. |
| Owner create facade | Imports and Tabular Ingest | Owner-dispatch boundary language. |
| Snapshot | Reporting and Snapshots | Extension-profile language. |
| Release | Reporting and Snapshots | Extension-profile language. |
| Recipient partition | Reporting and Snapshots | Extension-profile language. |
| Report composition | Reporting and Snapshots | Extension-profile authoring-input language. |
| Composition version | Reporting and Snapshots | Extension-profile authoring-input version language. |
| Composition operation | Reporting and Snapshots | Extension-profile authoring operation language. |
| Semantic composition anchor | Reporting and Snapshots | Extension-profile authoring target language. |
| Authored presentation text | Reporting and Snapshots | Extension-profile presentation-text language. |
| Composition diagram declaration | Reporting and Snapshots | Extension-profile diagram authoring language. |
| Composition diagram layout | Reporting and Snapshots | Extension-profile diagram layout authoring language. |
| Report builder | Reporting and Snapshots | Extension-profile builder UI language. |

## 12. Entity and relationship model

### 12.1 Record-envelope membership

A user-visible first-class incident record MUST consume one `record_id` through the record-envelope model. The following current-profile record types are first-class incident records.

| Record type | Domain object | Language owner | Behavior owner |
| --- | --- | --- | --- |
| `timeline_event` | Timeline event | §11 | Core 02 |
| `host` | Host | §11 | Core 02 |
| `identity` | Identity | §11 | Core 02 |
| `party` | Party | §11 | Core 02 |
| `indicator` | Canonical indicator | §11 | Core 02 |
| `artifact` | Notes, coordination artifacts, findings, and other structured analyst material | §11 | Core 02 |
| `task_request` | Task request | §11 | Core 02 |
| `decision` | Decision | §11 | Core 02 |
| `evidence` | Evidence record | §11 | Core 02 |
| `assessment` | Compromise assessment | §11 | Core 02 |

Administrative users, auth-provider identities, sessions, incident memberships, local-account credential state, bootstrap-completion state, object blobs, import-unit child state, and handoff risk-ref child rows MUST NOT be called first-class incident records unless a later owner spec changes their model.

### 12.2 Relationship families

| Relationship family | Authoritative representation | Domain rule | Behavior owner |
| --- | --- | --- | --- |
| Host or identity observed in another record | `entity_mention` plus optional resolution and typed link. | The raw mention remains preserved after resolution. | Core 02 |
| Indicator observed in another record | `indicator_observation` plus optional canonical indicator link. | Repeated observations remain distinct. | Core 02 |
| Evidence attached to a record | Evidence record plus typed record link and optional object blob. | Binary upload success alone is not sufficient to imply evidence availability. | Core 01/Core 02/Core 04 |
| Record-to-record association | `record_links` with exact `link_type` token family. | Relationship semantics are typed, not inferred from text columns. | Core 02 |
| Tags | `record_tags` association to incident-scoped tag. | Tags are labels, not lifecycle or ownership state. | Core 02 |
| Coordination collections | `record_ref`, `party_ref`, or `risk_ref` item families depending on field. | Public item kind is field-family-specific and not a free-form string array. | Core 01/Core 02 |
| Entity merge | Explicit merge action preserving pre- and post-merge graph. | Merges are explicit and auditable. | Core 02/Core 01 |
| Timeline supersession | Typed `supersedes` relation when replacement row is selected. | Supersession is reviewer action, not ordinary edit. | Core 02/Core 03 |

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

Workflow vocabulary in this section is domain-language orientation. Exact lifecycle transitions, token membership, route request shapes, errors, idempotency, authorization, and persistence behavior remain owned by the cited owner sections.

### 13.1 Rough timeline capture

Rough capture is analyst-entered partial or uncertain incident information that is valid before canonical structure is complete. A timeline row MAY be created with nullable time, one non-empty owner-accepted field, unresolved host or account text, indicator-bearing text, unstructured details, or only an attached screenshot when the owner route contract permits that create signal. Omission behavior: canonical host, identity, indicator, or evidence normalization MUST NOT be required before the rough capture term may be used.

| Vocabulary point | Domain interpretation | Owner |
| --- | --- | --- |
| Rough capture | First durable capture state for incomplete but useful timeline input. | Core 02/Core 03 |
| Enrichment | Later addition of structured or resolved information. | Core 02/Core 03 |
| Review | Reviewer action or state owned by the timeline workflow. | Core 03 |
| Supersession | Reviewer-selected replacement relation, not ordinary edit. | Core 02/Core 03 |
| Rollback | Historical corrective action, not a timeline lifecycle token. | Core 01/Core 02 |

### 13.2 Mention resolution

Mention resolution is the explicit process of linking a source-bound textual observation to an existing or newly created host or identity. The raw mention MUST remain preserved. Creating a stub from one mention resolves only that selected mention by default. Bulk resolution of sibling mentions requires a separate explicit action defined by the owner section.

Automatic background matching MAY suggest candidates. Omission behavior: when no owner-defined explicit action or binding mode permits mutation, suggestions remain non-mutating and MUST NOT create stubs or merge entities.

### 13.3 Auto-resolution

Auto-resolution is a narrow current-profile eligibility path for interactive mention capture on owner-defined Timeline relationship cells. It is not general fuzzy matching. If any required owner-defined condition fails, the ordinary unresolved path applies and the UI MAY present non-mutating suggestions. Omission behavior: if suggestions are not presented, no entity or relationship mutation is implied.

Auto-resolution vocabulary MUST NOT be used for party references, import-time entity creation outside the owner-defined binding mode, indicator canonicalization, fuzzy matching, partial alias matching, batch merge, batch dedupe, or visible-label based matching.

### 13.4 Indicator capture

Indicator capture from Timeline, Notes, Evidence, or other supported source fields preserves the raw source field. It does not require dedicated IOC columns on non-indicator sheets. A canonical indicator is useful when the value or pattern needs linkability, lifecycle intervals, dedupe, filtering, or history beyond the source-bound occurrence.

### 13.5 Evidence request, receipt, and attachment

Evidence work has two linked but distinct state families: object-blob upload state and evidence-record lifecycle state. Exact token membership is owner-derived and referenced in §15 rather than copied here.

Binary evidence attachment uses a two-step upload and finalization flow. A pending or requested evidence record may exist without a blob when the owner route contract permits it. Upload hints such as filename and content type are advisory metadata, not authorization, storage-key, preview-allowlist, or release-posture authority.

An upload target is an opaque capability for completing one pending blob upload. In the base profile it is an app-owned same-origin route, not a domain identifier, storage key, bucket reference, or object-store URL.

### 13.6 Task request workflow

Task Requests model owned work. A task request is the correct domain object when work needs an accountable owner, lifecycle state, due or blocker tracking, queue visibility, handoff durability, or later reconstruction.

A task request MUST NOT be used as a hidden substitute for ordinary timeline capture. A tag or note MUST NOT be used as a substitute for owned work when lifecycle, ownership, blockers, due dates, or follow-through are required.

### 13.7 Decision workflow

Decisions model incident-scoped rationale-bearing choices. A decision is the correct domain object when the team needs durable rationale, owner accountability, status, support references, supersession, or later reconstruction.

A decision MUST NOT be used as a general approval gate for ordinary row edits. Release approval is a separate Snapshot and Reporting Extension Profile concern owned by Core 04.

### 13.8 Coordination artifacts

Communications Log, Handoff, Status Review, and Lesson are workbook-native coordination artifact surfaces. They remain part of the workbook model and MUST NOT become separate application modules in the current profile.

| Coordination artifact | Domain use | Not this | Owner |
| --- | --- | --- | --- |
| Communications Log | Durable communication memory. | Raw chat archive or approval engine. | Core 02/Core 03 |
| Handoff | Shift or responsibility continuity. | Routine edit note. | Core 02/Core 03 |
| Status Review | Checkpoint, blockers, risks, and next report timing. | Dashboard or mandatory per-edit ritual. | Core 02/Core 03 |
| Lesson | Debrief and improvement follow-through. | Generic knowledge-base article. | Core 02/Core 03 |

### 13.9 Saved views

A saved view is an incident-bound workbook configuration over exactly one immutable `view_schema_id`. Saved-view scope controls saved-view discoverability and mutability only. It MUST NOT widen or narrow access to underlying incident rows, fields, search results, export redaction behavior, or evidence visibility.

A `scope='system'` saved view is an implementation-owned saved-view configuration object. It is not a contract-backed system view and MUST NOT replace a required workbook surface.

### 13.10 Conflicts, history, and rollback

Same-field conflicts, record history, and rollback are review and recovery concepts. They MUST preserve source-state authority and MUST NOT be described as edits to projection rows. A conflict resolver, history view, or rollback action MAY appear in the UI only as an affordance over the owner-defined history and mutation contracts. Omission behavior: if the UI affordance is not present in a context, projection state still remains non-authoritative and the owner-defined history contract is unchanged.

### 13.11 Grouping, filtering, sorting, and search

Grouping, filtering, sorting, and search operate over workbook query and projection contracts. Domain language MUST refer to stable `field_key` and `view_schema_id` identities when behavior or compatibility matters. Visible headers, labels, chip text, and localized labels MUST NOT define query semantics.

### 13.12 Imports and tabular ingest

Import terminology is scoped to the Import Extension Profile or to base-profile clipboard paste where the owner sections reuse the tabular-ingest contract. `import_session` and `import_unit` are canonical contract nouns. Worksheet, table, used range, named range, and region are locator kinds or explanatory source terms; they are not runtime workbook identities. `import_apply_dispatcher_v1` and owner create facades name internal implementation boundaries only: imports owns source compatibility and apply dispatch, while source owners own durable record semantics.

The recognized Network Flow Activity boundary does not change those nouns. If
Core 01 later admits an analytical extension target, imports still owns the
source session, unit, mapping approval, opaque source capability, orchestration,
and terminal target-result publication; Network Flow Analysis owns only its
normalized analytical resource. A filename, worksheet, import unit, source
column, or parser DTO MUST NOT be called a Network Flow table or workspace.

### 13.13 Snapshots, reports, releases, recipient partitions, and report compositions

Snapshot/report/release and report-composition terminology applies only when the Snapshot and Reporting Extension Profile is implemented. Snapshots and reports derive from incident state. They do not define live workbook state. Recipient partitions are export-only disclosure boundaries and MUST NOT be described as live workbook authorization, row visibility, field visibility, or saved-view scope. Report compositions are incident-scoped authoring inputs for report rendering and MUST NOT be described as reports, templates, workbook artifacts, snapshots, release outputs, live workbook records, or generated source files.

## 14. Defaults and boundary conditions with domain significance

This section summarizes only defaults and omitted-case rules that materially affect domain language or agent divergence. Exact runtime defaults, validation, and error behavior remain delegated to the owner sections.

| Area | Domain-significant default or omission rule | Behavior owner |
| --- | --- | --- |
| Incident create | Incident creation defaults and required create fields are owner-route behavior; domain text MUST NOT invent incident identity, membership, or startup defaults. | Core 01/Core 04 |
| Timeline create | Rough capture may precede canonical host, identity, indicator, or evidence normalization when the owner route accepts the minimum create signal. | Core 02/Core 03 |
| Task request create | Task-request defaults belong to the task owner sections; tags or notes do not substitute for omitted task lifecycle state. | Core 02/Core 03 |
| Decision create | Decision defaults belong to owner sections; a decision is not a release approval unless the release owner defines it. | Core 02/Core 04 |
| Coordination creates | Coordination artifacts are workbook-native artifact-backed records; omission of a dedicated module does not omit the domain concept. | Core 02/Core 03 |
| Finding create | Structured findings are standardized optional artifact-backed surface behavior; absence of the optional surface is not a Base Profile gap. | Core 01/Core 02 |
| Saved-view create | Saved-view scope and omitted query/layout behavior are owner-owned; a system saved view never replaces a system view. | Core 03 |
| Saved-view query state | Empty sort/filter/group state has canonical owner-defined serialization; domain text MUST NOT invent alternate null/omitted semantics. | Core 01/Core 03 |
| Workbook startup | Startup surface selection uses owner-defined `sheet_ref` fallback semantics; domain text MUST distinguish saved views from base surfaces. | Core 03/Core 01 |
| Party text/reference pairs | Raw party text and optional `party_id` references have distinct meanings; omission of a `party_id` does not erase preserved text. | Core 02 |
| Object-blob upload | Blob slot creation and finalization are distinct; pending or failed blob state is not evidence availability proof. | Core 01/Core 02/Core 04 |
| Pagination cursor | Cursor behavior belongs to public interface owner; domain text may only say cursors are opaque continuation state. | Core 01 |
| Terminal jobs | Job retention behavior belongs to Core 01; expiring job inspection state MUST NOT be described as deleting durable incident outputs. | Core 01 |
| Authorization | Incident roles, saved-view scope, deployment admin, and release approvals are distinct; omission of one does not imply another. | Core 04/Core 03 |
| Reference packs | Missing optional reference packs may degrade overlays or enrichment; they MUST NOT block base capture unless an owner section explicitly says so. | Core 01/Core 04 |
| Report composition authoring | Composition defaults, draft/version lifecycle, semantic anchors, authored text, closed diagram layout data, and builder validation belong to the Report Composition NLSpec; Reporting owns only render consumption and effects. | Report Composition NLSpec; Reporting Subsystem NLSpec |

### 14.1 Defaults and omitted-case coverage disposition

Declared scope: every bounded context in §10. Completion rule: each bounded context has exactly one disposition row.

| Domain area | Disposition | Omission behavior in `docs/domain.md` | Owner |
| --- | --- | --- | --- |
| Incident Workspace | `owner-only` | Missing route-level defaults remain delegated. Domain text may summarize only incident/workspace meaning. | Core 01/Core 02/Core 04 |
| Workbook Interaction | `summarized here` | Distinctions among system view, saved view, surface status, and startup surface are summarized here; interaction behavior remains owner-owned. | Core 03/Core 01 |
| Capture and Timeline | `summarized here` | Rough-capture semantics are summarized; exact transition behavior remains owner-owned. | Core 02/Core 03 |
| Entities and Observations | `summarized here` | Mention/entity/observation separation is summarized; dedupe and merge behavior remain owner-owned. | Core 02 |
| Evidence | `summarized here` | Evidence/blob distinction is summarized; access, preview, and upload contracts remain owner-owned. | Core 01/Core 02/Core 04 |
| Coordination | `summarized here` | Coordination object vocabulary is summarized; exact lifecycle and fields remain owner-owned. | Core 02/Core 03 |
| Links and Tags | `owner-only` | Domain text states relationship families; exact link-type tokens and field mappings remain owner-owned. | Core 02/Core 01 |
| Revisions and Audit | `owner-only` | Domain text names history concepts; rollback and audit behavior remain owner-owned. | Core 02/Core 01/Core 04 |
| Projections and Search | `owner-only` | Domain text states derived-state boundary; sort/filter/group/cursor behavior remains owner-owned. | Core 01/Core 03 |
| Reference Data | `extension-only` | Missing reference packs are not Base Profile defects. | Core 01/Core 04 |
| Reporting and Snapshots | `extension-only` | Snapshot/report/release and report-composition terms are current only when the extension profile is claimed. | Core 01/Core 04; Report Composition NLSpec; Reporting Subsystem NLSpec |
| Authentication and Administration | `owner-only` | Domain text distinguishes user, party, incident role, and deployment admin; security behavior remains owner-owned. | Core 04/Core 01 |
| Imports and Tabular Ingest | `extension-only` | File-based import concepts are current only when the Import Extension Profile is claimed; base clipboard semantics remain owner-owned. | Core 01/Core 03 |
| Network Flow Analysis | `extension-only` | Network Flow terms are current extension vocabulary when the adopted profile is claimed; omission adds no route, workspace, resource, or Base surface. | Core 00; Core 01–04; Network Flow Activity NLSpec |
| Backup and Restore | `owner-only` | Domain text keeps backup/restore out of workbook semantics; recovery behavior remains owner-owned. | Core 01/Core 04 |

## 15. Closed vocabulary quick registry

This section is a quick domain reference for closed-vocabulary families. It is pointer-only by default. Exact token membership MUST be read from the owner sections or an owner-derived generated registry. This document MUST NOT become a second owner for token membership.

Declared scope: token families named in current `domain.md` because the family is likely to cause domain drift. Completion rule: every exact token family copied into this document MUST appear in §18.2 as a generated mirror with validation, otherwise it MUST remain pointer-only here.

| Token family | Exact-token source | Domain use | Mirror policy |
| --- | --- | --- | --- |
| `incident.status` | Core 02 closed-vocabulary registry; lifecycle semantics in Core 01 | Incident lifecycle visibility and source-state boundary. | Pointer-only. |
| `entity_mentions.resolution_status` | Core 02 closed-vocabulary registry | Mention resolution state. | Pointer-only. |
| `entity_mentions.origin_kind` and `indicator_observation.origin_kind` | Core 02 closed-vocabulary registry | Source of mention or indicator observation. | Pointer-only. |
| `host.entity_origin` and `identity.entity_origin` | Core 02 closed-vocabulary registry | Host/identity creation provenance. | Pointer-only. |
| Host and identity preserved identifier classification | Core 02 closed-vocabulary registry | Exact-match reuse, suggestions, and provenance treatment. | Pointer-only. |
| `party.party_kind` | Core 02 closed-vocabulary registry | Party classification. | Pointer-only. |
| `indicator.value_kind` | Core 02 closed-vocabulary registry | Indicator value family. | Pointer-only. |
| `assessment_state` | Core 02 closed-vocabulary registry | Compromise-assessment state. | Pointer-only. |
| `task_request.task_kind` | Core 02 closed-vocabulary registry | Task-request type. | Pointer-only. |
| `task_request.status` | Core 02/Core 03 owner sections | Task-request lifecycle. | Pointer-only. |
| `task_request.priority` | Core 02 closed-vocabulary registry | Queue priority. | Pointer-only. |
| `decision.decision_type` | Core 02 closed-vocabulary registry | Decision category. | Pointer-only. |
| `decision.status` | Core 02 closed-vocabulary registry | Decision lifecycle. | Pointer-only. |
| `record_links.link_type` | Core 02/Core 01 owner sections | Typed relationship semantics. | Pointer-only. |
| `record_links.provenance` | Core 02 closed-vocabulary registry | Relationship provenance. | Pointer-only. |
| `artifact.comm_type` for `artifact_type='comm_log'` | Core 02 closed-vocabulary registry | Communication-log category. | Pointer-only. |
| `handoff.ack_state` | Core 02 closed-vocabulary registry | Handoff acknowledgement state. | Pointer-only. |
| `lesson.closure_state` | Core 02 closed-vocabulary registry | Lesson closure state. | Pointer-only. |
| `finding.kind`, `finding.state`, and `finding.confidence_band` | Core 02/Core 01 owner sections | Optional structured finding vocabulary. | Pointer-only. |
| `forensic_keyword.match_mode` | Core 02/Core 01 owner sections | Optional forensic-keyword matching mode. | Pointer-only. |
| `release_state` | Core 04 release owner | Release artifact lifecycle. | Pointer-only. |
| `object_blobs.upload_state` and `object_blobs.terminal_reason` | Core 01/Core 02 owner sections | Object-blob lifecycle. | Pointer-only. |
| `evidence_records.lifecycle_state` | Core 02 owner section | Evidence-record lifecycle. | Pointer-only. |
| Evidence-access media and preview classes | Core 01/Core 04 owner sections | Evidence preview/download classification. | Pointer-only. |
| Base incident roles | Core 04 authorization owner | Incident role model. | Pointer-only. |
| Saved-view scope | Core 03 saved-view owner | Saved-view discoverability and mutability. | Pointer-only. |
| Account density mode | Core 01 current-account preference owner | Deployment-local density override and no-override state. | Pointer-only. |
| Local credential recovery model and TOTP state | Core 04/Core 01 auth owners | Credential lifecycle vocabulary. | Pointer-only. |

Display labels MAY map these tokens for users. Omission behavior: when a display label is omitted from this document, the exact machine-readable token still remains owner-owned and MUST be used in structured state and public payloads according to the owner section.

## 16. External-system and extension boundaries

External systems are upstream or adjacent models. They MUST NOT own Cartulary domain language. External vocabulary enters Cartulary only through the anti-corruption mappings in §16.1 or an owner-defined integration contract.

| External or adjacent concern | Domain boundary | Required language |
| --- | --- | --- |
| SIEM, EDR, telemetry stores | External source or pivot target. Cartulary can reference queries, indicators, evidence, or findings derived from them; after Network Flow Activity adoption, an approved import may also normalize bounded flow data into an extension-owned analytical table. | Do not call raw telemetry Core source state, a Core record, a Network Flow table, or a Network Flow row before owner-defined normalization and atomic publication. |
| Object storage | Authoritative binary evidence backing service. Cartulary owns object metadata and evidence access semantics. | Do not expose raw object-store URLs as evidence identity. |
| Enterprise IdP | Optional enterprise authentication provider. Successful provider auth maps to internal user and server-managed session. | Do not call provider subject a party, incident identity, or authorization role. |
| CMDB or asset inventory | External enterprise master data. Cartulary host records remain incident-scoped investigation records. | Do not treat external asset identity as `record_id`. |
| Ticketing system | External task or request system. Cartulary task requests can store `external_ticket_ref`. | Do not replace task lifecycle with ticket state unless an owner spec says so. |
| Reference-data sources | Optional packs or enrichment inputs. | Do not block base capture on live reference data. |
| Report templates, report compositions, and rendered outputs | Snapshot/reporting extension artifacts and authoring inputs. | Do not treat report composition authoring or report release state as live workbook approval. |
| Backup and restore systems | Deployment-local operator-facing recovery. | Do not expose backup/restore as workbook route families in the current profile. |
| Spreadsheet files | Import or clipboard source material. | Do not treat workbook file objects, worksheets, tables, or parser DTOs as runtime workbook surfaces or source-owner mutation contracts. |
| Threat-intelligence APIs | Optional enrichment or pivot sources. | Do not promote provider reputation, verdict, or object IDs into canonical incident state without owner-defined fields. |
| Coding agent or AI assistant | Implementation-support actor that may propose text, code, mappings, summaries, or terminology. It does not own Cartulary domain language. | Do not treat agent-generated names, inferred entities, summaries, or suggested mappings as Cartulary vocabulary unless accepted through owner-backed review. |

### 16.1 External anti-corruption map

Declared scope: every external or adjacent concern in §16. Completion rule: each §16 row has at least one canonical Cartulary target and one forbidden promotion.

| External system or model | External term or identifier | Allowed preservation form | Canonical Cartulary target | Translation owner | Forbidden promotion |
| --- | --- | --- | --- | --- | --- |
| SIEM/EDR/telemetry stores | Event IDs, query IDs, detector names, raw event fields, exported flow records. | Source text, evidence reference, investigative query, finding support reference, indicator observation origin; owner-approved import source material for Network Flow. | Timeline event, indicator observation, evidence record, investigative query, or finding when owner fields permit; normalized Network Flow row only through the adopted import/extension boundary. | Core 02/Core 01; Network Flow Activity NLSpec | Raw telemetry event MUST NOT become Core source state, `record_id`, `view_row_v1`, Network Flow table identity, or a published Network Flow row without normalization and atomic commit. |
| Object storage | Bucket, key, presigned URL, storage class. | Object-blob metadata and owner-defined access handles. | `object_blob_id`, evidence access handle, evidence record link. | Core 01/Core 04 | Raw object-store URL/key MUST NOT become evidence identity or workbook cell authority. |
| Enterprise IdP | Provider subject, group claim, assertion, tenant ID. | Deployment-local auth binding or server-side configuration. | Internal `user_id` and server-managed session. | Core 04/Core 01 | Provider subject MUST NOT become `party_id`, incident role, identity, or incident membership. |
| CMDB or asset inventory | Asset ID, device ID, owner, business service. | Source/provenance text, optional external reference, enrichment metadata. | Host record fields or references only when owner-defined. | Core 02/Core 01 | External asset ID MUST NOT become `record_id` or host canonical identity by itself. |
| Ticketing system | Ticket key, ticket status, assignee, priority. | `external_ticket_ref` or note/source text. | Task request, decision, or communications-log row when owner-defined. | Core 02/Core 03 | Ticket status MUST NOT replace task-request lifecycle or authorization. |
| Reference-data source | Framework object ID, version, vocabulary term. | Reference-pack payload or activation metadata. | Reference pack, type registry, enrichment metadata. | Core 01/Core 04 | Pack object MUST NOT become incident record or required capture dependency. |
| Report template/composition/rendered output | Template ID, composition ID, composition version, rendered file path, release bundle. | Snapshot/reporting extension artifact metadata and composition authoring metadata. | Snapshot, release, recipient partition, report composition, rendered output. | Core 01/Core 04/Report Composition NLSpec/Reporting Subsystem NLSpec | Release state MUST NOT become live row approval or saved-view scope, and report composition MUST NOT become a workbook artifact, template, snapshot, or release output. |
| Backup/restore system | Backup job ID, snapshot path, restore target. | Deployment-local operator-facing state. | Backup set, restore verification, runtime root binding. | Core 01/Core 04 | Backup/restore object MUST NOT become incident-scoped workbook route or coordination artifact. |
| Spreadsheet file | Worksheet name, table name, named range, used range. | `import_unit.locator_kind`, source locator, provenance. | Import session, import unit, mapping fingerprint. | Core 01 | Worksheet/table identity MUST NOT become `view_schema_id`, `sheet_ref`, or runtime surface identity. |
| Threat-intelligence API | Provider verdict, reputation, lookup ID, enrichment field. | Optional enrichment metadata or reference-pack content. | Indicator, indicator observation, finding support, or reference pack only through owner-defined mapping. | Core 02/Core 01/Core 04 | Provider verdict MUST NOT overwrite canonical assessment or indicator lifecycle without owner-defined field semantics. |
| Coding agent or AI assistant | Suggested term, inferred entity, generated summary, generated mapping, code-derived vocabulary cluster. | Draft text, review note, implementation-support artifact, or TODO marker. | Canonical domain term only after §6.1 classification, §11 glossary mapping, §11.1 bounded-context mapping, and owner-backed review. | `docs/domain.md` plus applicable Core owner section | Agent output MUST NOT become domain vocabulary, record type, surface identity, field identity, or owner behavior by repetition in generated code, tests, comments, or prompts. |

## 17. Coding-agent rules

A coding agent working on Cartulary domain-facing code, tests, specs, comments, prompts, generated contracts, or user-facing documentation MUST apply these rules before making changes.

### 17.1 Required orientation

1. Identify the bounded context in §10.
2. Identify the domain terms in §11.
3. Identify each term's bounded context in §11.1.
4. Identify the owner sections before asserting behavior.
5. Use stable identifiers from §8 instead of visible labels.
6. Apply the term-classification decision tree in §6.1.
7. Apply the alias registry in §6.2.
8. Treat implementation modules as mappings, not definitions.
9. Treat external vocabulary through §16 anti-corruption mappings.
10. Treat Strategic DDD source posture through §4.2 when using DDD terms such as bounded context, ubiquitous language, context map, published language, or anti-corruption layer.
11. When a behavior is unclear, use `TODO: owner decision required` or `TODO: owner lookup required` rather than inventing behavior.

### 17.2 Prohibited shortcuts

A coding agent MUST NOT:

1. Treat `party`, `user`, `identity`, `incident membership`, and email text as interchangeable.
2. Treat `artifact` as a synonym for Notes or binary evidence.
3. Treat a required system view as a saved view.
4. Treat every `surface_kind='system_view'` row as a required surface.
5. Infer behavior from visible tab labels, visible column labels, row order, SQL names, projection table names, React component names, or generated filenames.
6. Mutate projection state as authoritative source state.
7. Auto-create host or identity records from `mention_origin` fields.
8. Auto-create or auto-link party records from ordinary party text.
9. Treat an object blob as an evidence row or evidence availability proof.
10. Store relationships, canonical identifiers, timestamps used for sorting, or evidence retrieval metadata only in JSON or raw text.
11. Add mandatory owner, approver, challenge, checklist, task, or decision fields to ordinary timeline capture.
12. Implement coordination surfaces as separate application modules that leave the workbook interaction model.
13. Expose ATT&CK, D3FEND, VERIS, or other pack overlays as base workbook `view_schema` resources.
14. Treat release approval as ordinary row-edit approval.
15. Use findings or hypotheses as first-class `record_type`s in the current profile.
16. Treat import worksheet/range/table identity as runtime workbook identity.
17. Treat external provider identifiers as Cartulary stable identifiers without an anti-corruption mapping.
18. Treat coding-agent, AI-assistant, static-analysis, or generated-code vocabulary as Cartulary domain language without §6.1 classification, §11 glossary coverage, §11.1 bounded-context mapping, and owner-backed review.
19. Copy exact token membership into `docs/domain.md` unless the copy is generated or validated as required by §18.2.

### 17.3 Review checklist for generated or agent-authored text

A reviewer MUST reject generated text that:

- defines a Cartulary term without an owner reference;
- introduces a synonym for a canonical term without a §6.2 registry row;
- presents implementation package names as domain definitions;
- omits the `Not this` boundary for a high-risk term;
- claims behavior from an appendix, guide, research report, or generated artifact without checking whether the core owns that behavior;
- describes a future or extension capability as Base Profile behavior;
- uses display names where stable identifiers are required;
- describes external-system concepts without a §16 anti-corruption mapping;
- promotes coding-agent or AI-assistant output to domain vocabulary without §6.1 classification, §11 glossary coverage, §11.1 bounded-context mapping, and owner-backed review;
- treats document-readiness acceptance criteria as product conformance or publication evidence.

## 18. Maintenance rules

`docs/domain.md` MUST remain small enough to be used as prompt context and rigorous enough to prevent semantic drift.

### 18.1 Required update triggers

A change set MUST update `docs/domain.md` or explicitly state `domain vocabulary unchanged` when it changes any of the following:

- first-class record type membership;
- standardized `view_schema_id` membership;
- stable identifier family;
- closed vocabulary token family or exact token membership;
- entity-binding mode behavior;
- party-reference semantics;
- artifact-backed variant membership;
- evidence or blob lifecycle vocabulary;
- saved-view scope semantics;
- workbook startup surface semantics;
- import object vocabulary;
- snapshot/report/release and report-composition vocabulary;
- extension-profile claim vocabulary;
- external-system translation boundary;
- context-map relationship;
- a term listed in §5, §6.2, §8, §9, §10, §11, §15, §16, or §20.

### 18.2 Copied owner facts policy

Declared scope: every owner-owned exact fact copied or summarized by this document. Completion rule: each copied fact family has exactly one copy class and validation requirement.

| Copied table or fact family | Owner source | Copy class in `docs/domain.md` | Allowed detail level | Validation requirement | Failure handling |
| --- | --- | --- | --- | --- | --- |
| Document authority and profile posture | Core 00 | Manual summary | Authority boundary summary only. | Reviewer verifies against Core 00 during domain updates. | Owner governs; repair documentation drift. |
| Workbook-surface identity mapping in §9.2 | Core 01 Table 7.4-A | Manual identity mirror | `surface`, `view_schema_id`, `surface_kind`, `source_record_types`, discriminator/filter, `surface_status`, reference-pack keys. | Must be checked against Core 01 or generated view-schema registry before change acceptance. | Emit no domain update until mismatch is resolved. |
| Glossary owner references in §11 | Core 00 through Core 04 | Manual summary | Term meaning, Not-this boundary, identifiers, owner. | Reviewer verifies owner existence and no route/field-contract duplication. | Replace unsupported owner with `TODO: owner lookup required`. |
| Defaults summary in §14 | Core 01 through Core 04 | Manual summary | Domain-significant omitted-case interpretation only. | Reviewer verifies no route-shape or field-registry duplication. | Remove runtime-default text or delegate to owner. |
| Token quick registry in §15 | Core 02/Core 03/Core 04 owner sections | Owner pointer only | Token family name, owner, domain use. | Exact tokens MUST NOT appear unless generated mirror validation exists. | Remove manual exact token list. |
| External anti-corruption map in §16.1 | Core owners plus external-system boundary rows | Manual summary | External term, allowed preservation, canonical target, forbidden promotion. | Reviewer verifies canonical target exists or is marked future/external. | Add `TODO: owner decision required` or remove row. |
| Future-only table in §20 | Core 00 unsupported future areas and roadmap owner decisions | Manual summary | Current handling and allowed locations only. | Reviewer verifies no future-only row is described as current behavior. | Reject current-profile wording. |
| Strategic DDD source posture in §4.2 | External DDD source posture plus this document | Manual summary | Source-family role and authority posture only. | Reviewer verifies the source family supports the stated DDD concept and that no external source overrides Cartulary owner sections. | Remove source claim or reclassify as rationale. |
| Context-kind and subdomain classification in §10 and §10.1 | This document | Domain-owned classification | Context kind, subdomain class, modeling latitude, and pressure-test language. | Reviewer verifies every §10 row has one `context_kind` and every §10.1 row satisfies the closed subdomain table. | Add `TODO: owner decision required` only when owner evidence is insufficient; otherwise reject the update. |
| Term-to-bounded-context map in §11.1 | This document plus §11 glossary | Domain-owned validation mirror | One row per §11 term; context name; context role. | Validation verifies exact one-to-one coverage between §11 term names and §11.1 term names. | Emit no domain update until missing, duplicate, or orphan mappings are resolved. |
| Context relationship type status in §10.3 | This document plus Strategic DDD source posture | Domain-owned vocabulary | Relationship type, status, meaning, omission or use behavior. | Validation verifies every §10.2 relationship type is declared `current-used` in §10.3. | Reject unclassified or recognized-unused relationship use. |
| Acceptance criteria in §19 | This document | Domain-owned | Document-readiness pass/fail criteria. | Criteria MUST trace to sections and not duplicate Core 04 product ACs. | Rewrite subjective or duplicated criteria. |

### 18.3 Drift-control rules

1. `docs/domain.md` MUST point to owner sections instead of copying full route contracts or field registries.
2. Any copied exact token list MUST identify the owner section and MUST be generated from or validated against the owner-derived token registry in the same change set. In this revision, §15 is pointer-only.
3. New glossary entries MUST include definition, forbidden interpretation when ambiguity is plausible, canonical identifiers when applicable, language owner, behavior owner, applicability, and current-profile status.
4. Future-only terms MUST be labeled as future-only and MUST NOT be mixed into Base Profile current terms.
5. Deprecated or migrated terms MUST name the canonical replacement and the owner section that permits migration handling.
6. A reviewer MUST treat near-duplicate terms as drift risk unless the distinction is explicit.
7. A table that declares itself exhaustive MUST state its declared scope and completion rule.
8. Every §11 glossary term MUST appear exactly once in §11.1.
9. Every §11.1 bounded-context value MUST match one §10 bounded-context name unless the row explicitly uses an external or future-only boundary.
10. Every §10 bounded-context row MUST have exactly one `context_kind`.
11. A §10.2 relationship type MUST appear in §10.3 with `status='current-used'`.
12. A relationship type with `status='recognized-unused'` or `status='recognized-diagnostic-only'` MUST NOT appear in §10.2.
13. Every §10.2 row MUST state a change obligation.
14. Every `core` row in §10.1 MUST satisfy the core-domain pressure test.
15. Coding-agent or AI-assistant vocabulary MUST enter only through §16 and §16.1 until accepted as canonical vocabulary by §6.1, §11, §11.1, and the applicable owner section.

### 18.4 Repository checks

A repository validation suite for `docs/domain.md` MUST fail the domain-document check when any of the following conditions is true:

- a `cartulary.view.*.v1` string in §9 is absent from the owner-derived view-schema registry, is not listed by Core 01 as a standardized optional surface, and is not labeled future-only;
- an exact token membership list appears outside an owner-derived generated mirror;
- a glossary row lacks behavior owner, applicability, or current-profile status;
- a bounded context lacks a subdomain classification or context-map relationship;
- an external-system row lacks an anti-corruption mapping;
- a future-only row lacks current handling;
- an acceptance criterion lacks trace, pass condition, or failure condition.
- a §10 bounded-context row lacks `context_kind` or uses a value outside the closed `context_kind` table;
- a §10.1 row with `Subdomain class='core'` lacks pressure-test support in its reason or modeling-latitude text;
- a §10.2 relationship row lacks `Change obligation`;
- a §10.2 relationship type is absent from §10.3 or is not `status='current-used'`;
- a §10.3 relationship type with `status='recognized-unused'` or `status='recognized-diagnostic-only'` appears in §10.2;
- a §11 glossary term is absent from §11.1;
- a §11.1 term does not exactly match one §11 glossary term;
- a §11.1 bounded context does not exactly match one §10 bounded context and is not explicitly marked external or future-only;
- coding-agent or AI-assistant vocabulary appears as a domain source without a §16.1 anti-corruption row and owner-backed glossary mapping;
- a domain-language manifest or lint artifact, when present, disagrees with §10, §10.2, §10.3, §11, §11.1, or §16.1.

Omission behavior: if the repository does not yet contain such a validation suite, reviewers MUST perform the same checks manually before accepting a behavior-affecting domain update.

### 18.5 Domain-language manifest validation shape

A repository validation suite MAY materialize a domain-language manifest for automated checks. Omission behavior: if no manifest exists, §18.4 manual review remains required and sufficient for this document.

When present, the manifest MUST be derived from `docs/domain.md` and MUST NOT become an independent source of domain truth.

| Manifest family | Required content |
| --- | --- |
| `bounded_contexts[]` | §10 bounded-context names, `context_kind`, applicability, §10.1 subdomain class, and stewardship trigger reference. |
| `context_relationships[]` | §10.2 upstream, downstream, relationship type, published language, translation owner, and change obligation. |
| `relationship_types[]` | §10.3 relationship type, status, meaning, and omission or use behavior. |
| `glossary_terms[]` | §11 term, canonical identifiers or tokens, behavior owner, applicability, current-profile status, and §11.1 bounded context. |
| `external_mappings[]` | §16.1 external system or model, external term or identifier, allowed preservation form, canonical Cartulary target, translation owner, and forbidden promotion. |

A generated manifest mismatch MUST be treated as evidence of documentation drift, not as authority to reinterpret `docs/domain.md`.

## 19. Acceptance criteria

The document is useful and complete enough for practical repository use only when every criterion in this section passes. These are document-readiness criteria only. They are not Core 04 implementation-conformance criteria and not Core 05 publication criteria.

| ID | Criterion | Traces to | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| DOMAIN-AC-AUTH-001 | The document states its authority boundary. | §1, §1.1 | It says `docs/domain.md` governs vocabulary, classification, owner navigation, and review discipline only. | Any text states or implies this document owns runtime behavior not owned by Core 00 through Core 04. |
| DOMAIN-AC-AUTH-002 | Core 05 is publication-only. | §1, §4, §4.1 | Core 05 is described as claim-bearing publication authority only. | Core 05 is described as Base Profile runtime behavior. |
| DOMAIN-AC-DELEGATION-001 | Adjacent-spec boundaries are explicit. | §4.1 | Every adjacent artifact family in §4 appears with retained concept, delegated behavior, and boundary rule. | An adjacent spec is named without a delegation boundary. |
| DOMAIN-AC-DDD-SOURCE-001 | Strategic DDD source posture is explicit. | §4.2 | The document identifies which DDD source families define DDD terms, which inform tooling/practice, which are evidence only, and that Cartulary owner sections govern behavior. | DDD terms are used without a source posture or an external DDD source is allowed to override owner sections. |
| DOMAIN-AC-TERMS-001 | Resolved terminology decisions cover high-risk distinctions. | §5, §7 | The table includes artifact/Notes, party/user, identity/party, system view/saved view, projection/source state, mention/entity, indicator observation/indicator, evidence/blob, coordination/workflow, import unit, reference pack, snapshot/report, and report composition versus report/template/workbook artifact. | Any named high-risk distinction is missing or contradicted. |
| DOMAIN-AC-CLASSIFY-001 | Candidate term classification is deterministic. | §6.1 | Each row in the decision tree has one terminal classification and omission behavior. | A candidate term matches no row or produces conflicting required locations. |
| DOMAIN-AC-ALIAS-001 | Known aliases have canonical handling. | §6.2, §17.2 | Every prohibited shortcut in §17.2 appears in §6.2 or is covered by a more specific canonical row. | A prohibited shortcut lacks canonical term, allowed context, or forbidden context. |
| DOMAIN-AC-ID-001 | Stable identifiers are separated from labels and implementation names. | §8, §8.1 | Every identifier row has target, stable use, Not-this boundary, owner, applicability, and status. | Any identifier row lacks those fields or treats a label as stable identity. |
| DOMAIN-AC-SURFACE-001 | Surface kind and surface status are distinct. | §9.1, §9.2 | The registry has both `surface_kind` and `surface_status`, and `system_view` requiredness is determined only by `surface_status`. | A row or rule implies every `surface_kind='system_view'` surface is required. |
| DOMAIN-AC-SURFACE-002 | Current-profile surface registry is exhaustive in scope. | §9.2 | The table contains exactly the fourteen required surfaces plus three standardized optional surfaces. | A required surface is missing, an optional surface is unlabeled, or an extra current-profile surface appears. |
| DOMAIN-AC-CONTEXT-001 | Bounded contexts have owner and tie-break closure. | §10 | Every bounded context has language home, behavior-owner disposition, tie-break rule, and applicability. | Any bounded-context row leaves owner resolution to reader judgment. |
| DOMAIN-AC-SUBDOMAIN-001 | Strategic DDD subdomain classes are explicit. | §10.1 | Every bounded context has exactly one subdomain class and modeling-latitude rule. | Any context lacks a subdomain class or allows concept leakage. |
| DOMAIN-AC-CONTEXTMAP-001 | Context relationships are explicit. | §10.2, §10.3 | Every bounded context appears in at least one context-map row or a `separate_ways` disposition. | Any bounded context has no relationship disposition. |
| DOMAIN-AC-CONTEXT-KIND-001 | Bounded-context kind is closed and explicit. | §10 | Every bounded-context row has exactly one `context_kind` from the closed table. | A bounded context lacks `context_kind`, uses an undeclared value, or mixes domain, interaction, supporting, generic, and external meanings without classification. |
| DOMAIN-AC-CORE-PRESSURE-001 | Core subdomain classification is justified. | §10.1 | Every `core` row satisfies the product-identity, domain-correctness, and replacement-risk pressure test. | A context is classified as `core` only because it is necessary for implementation or because a package/module exists. |
| DOMAIN-AC-CONTEXTMAP-CHANGE-001 | Context-map rows define change obligations. | §10.2 | Every context-map row has one explicit change obligation tied to its relationship type. | A relationship row permits upstream vocabulary change without naming the downstream update or breaking-change handling. |
| DOMAIN-AC-RELTYPE-001 | Context relationship type status is closed. | §10.3 | Every relationship type is classified exactly once, and every §10.2 relationship type has `status='current-used'`. | A recognized-unused or diagnostic-only relationship appears in §10.2, or a §10.2 relationship type is undeclared. |
| DOMAIN-AC-CONTEXT-STEWARDSHIP-001 | Context stewardship is non-runtime and review-only. | §10.4 | Every bounded context has a vocabulary steward and change-review trigger, and §10.4 says stewardship does not create runtime authorization, team ownership, deployable ownership, package ownership, or module ownership. | Stewardship text implies runtime access control, team ownership, deployable ownership, or module ownership. |
| DOMAIN-AC-TERM-CONTEXT-001 | Glossary terms map to bounded contexts. | §11, §11.1 | Every §11 glossary term appears exactly once in §11.1 and maps to one §10 bounded context or explicit external/future-only boundary. | A glossary term is unmapped, duplicated, mapped to a non-existent context, or mapped only through `Language owner=§11`. |
| DOMAIN-AC-GLOSSARY-001 | Glossary rows are closed. | §11 | Every glossary row has definition, Not-this boundary, canonical identifiers or tokens, language owner, behavior owner, applicability, and current-profile status. | Any glossary row lacks one required column or contains unresolved multi-owner text without `TODO: owner decision required`. |
| DOMAIN-AC-RELATIONSHIP-001 | Entity and relationship invariants are explicit. | §12 | First-class record types and relationship families are named without promoting administrative, child, or external objects to records. | Administrative users, object blobs, import units, or risk refs are called first-class incident records. |
| DOMAIN-AC-WORKFLOW-001 | Workflow vocabulary is closed to domain orientation. | §13 | The section covers rough capture, mention resolution, auto-resolution, indicator capture, evidence, tasks, decisions, coordination artifacts, saved views, conflicts/history/rollback, grouping/filtering/sorting/search, imports, snapshots/releases, and report compositions. | A workflow term defines route shape, field registry, security behavior, or token membership as domain-owned behavior. |
| DOMAIN-AC-DEFAULT-001 | Defaults and omitted cases are scoped. | §14, §14.1 | Every bounded context has a default/omitted-case disposition. | A bounded context is absent from §14.1. |
| DOMAIN-AC-TOKEN-001 | Token membership uses define-once discipline. | §15, §18.2 | §15 is pointer-only or exact-token mirrors are generated/validated. | A manually maintained exact token list appears without generated validation. |
| DOMAIN-AC-EXTERNAL-001 | External anti-corruption boundaries are explicit. | §16, §16.1 | Every external concern has canonical target, allowed preservation form, translation owner, and forbidden promotion. | External vocabulary enters the model without a mapped target or forbidden-promotion rule. |
| DOMAIN-AC-AI-BOUNDARY-001 | Coding-agent and AI-assistant vocabulary is anti-corrupted. | §16, §16.1, §17 | Coding-agent or AI-assistant output is treated as implementation-support input unless owner-backed review accepts it through §6.1, §11, and §11.1. | Agent-generated terms, inferred entities, summaries, or mappings become domain vocabulary by repetition in generated code, tests, comments, prompts, or documentation. |
| DOMAIN-AC-AGENT-001 | Coding-agent rules prohibit known drift patterns. | §17 | Required orientation, prohibited shortcuts, and review checklist are present. | A prohibited shortcut is allowed without an owner-backed exception. |
| DOMAIN-AC-MAINT-001 | Copied owner facts are controlled. | §18.2 | Every copied owner fact family has copy class, allowed detail, validation, and failure handling. | An exact copied fact exists without a copy-policy row. |
| DOMAIN-AC-MANIFEST-001 | Domain-language validation artifacts are subordinate. | §18.5 | Any domain-language manifest is derived from `docs/domain.md`, includes the declared strategic DDD sections, and is not treated as independent authority. If absent, §18.4 manual review remains required. | A generated manifest redefines domain vocabulary, omits the required strategic DDD sections, or is used to bypass manual review when no validator exists. |
| DOMAIN-AC-FUTURE-001 | Future-only concepts have current handling. | §20 | Every future-only topic has current status, current handling, allowed locations now, and required next step. | A future-only topic is implemented or documented as current behavior without rejection. |
| DOMAIN-AC-ECONOMY-001 | Spec economy is preserved. | Whole document, §18 | The document does not duplicate full route contracts, field registries, or product acceptance criteria owned elsewhere. | The document contains an owner-owned route table, exhaustive field registry, or Core 04 product AC copy as domain authority. |
| DOMAIN-AC-TWO-AGENT-001 | Two independent agents would make the same boundary decisions. | §5 through §20 | For the aliases `case`, `worksheet`, `IOC`, `artifact`, `approval`, `owner`, `contact`, `ticket`, `report`, `release`, `report composition`, `system view`, `flow table`, and `Network Analysis`, the document maps each to one canonical term, owner disposition, applicability class, and current handling. | Any listed alias yields incompatible current-profile interpretations. |

## 20. Open issues and future additions

This revision intentionally leaves the following as owner-driven future work rather than resolving them in `docs/domain.md`.

Declared scope: known future-only topics that appear in the current document set or prior domain review. Completion rule: every row has current handling and allowed locations now.

| Topic | Current status | Current handling | Allowed locations now | Required next step |
| --- | --- | --- | --- | --- |
| Network Flow Activity claim and runtime behavior | Adopted/current bounded extension identity, claimable only after deployment configuration validates. | Permit unclaimed discovery identity; expose route, workspace, storage, and conformance behavior only when claimed. | Adopted NLSpec, Core owner amendments, validated deployment configuration, implementation, fixtures, and executable evidence. | Keep owner interfaces, contracts, runtime, fixtures, and evidence synchronized for each later revision. |
| Whole-incident removal and Network Flow purge cascade | Future-only; the current incident lifecycle has only active/closed plus close/reopen. | Reject a Network Flow-specific purge route, private incident-removal state, or claim that incident closure deletes extension data. | Future profile design and explicitly labeled follow-up obligations only. | Define a generic Core incident-removal lifecycle and cascade-participant interface before admitting Network Flow. |
| New first-class hypothesis record type | Not current-profile behavior. Hypotheses remain artifact-backed findings. | Reject current-profile implementation as a first-class `record_type`; allow only artifact-backed finding language. | §11 Finding row, roadmap text, non-normative rationale. | Define a later NLSpec or profile if promotion becomes necessary. |
| Party merge and phone-based dedupe | Not standardized in the current profile. | Reject implementation claims of standardized party merge or phone-based dedupe. | Roadmap text or non-normative examples labeled future-only. | Define explicit party merge, phone normalization, and dedupe semantics before implementing. |
| Additional framework overlay workbook surfaces | Not standardized as Base Profile or current Reference Pack Extension Profile workbook surfaces. | Reject current-profile `view_schema_id` discovery for ATT&CK, D3FEND, VERIS, or similar overlays. | Reference-pack rationale, roadmap text, owner-design notes labeled future-only. | Define exact `view_schema_id`, fields, write behavior, pack dependencies, and compatibility rules in a later owner section. |
| Generalized approval workflows | Out of scope for ordinary row edits. | Reject ordinary row-approval workflow engine semantics. | Release-gate extension language, roadmap text, non-normative rationale. | Define a future bounded workflow profile only if specific domain evidence justifies it. |
| Live sensitive-evidence visibility model | Out of scope for the base live workspace model. | Reject live row/field hiding as recipient-specific release withholding in the current profile. | Roadmap or security-design notes labeled future-only. | Define a later profile if export-scoped withholding is insufficient. |
| Cross-incident analytics | Reserved for future specification work. | Reject current-profile cross-incident query or analytics claims. | Roadmap, research, or non-normative product strategy text. | Define cross-incident data model, privacy boundary, query surface, and conformance criteria. |
| Local-account WebAuthn/passkeys | Not current-profile behavior. | Reject current-profile auth route, registration, assertion, credential enumeration, or recovery semantics. | Security roadmap text labeled future-only. | Define a later auth NLSpec or Core 04 revision. |
| Durable local draft persistence | Not current-profile behavior unless an owner spec adds it. | Reject claims that pending browser edits survive reload, cross-tab transfer, or offline multi-master operation. | Implementation-support rationale labeled future-only. | Define local draft storage, replay, conflict, security, and cleanup semantics. |
