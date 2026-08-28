# links Module Refactoring Tracker and Handoff

> **Current execution authority:** Section 14 is the sole forward plan. Sections
> 1 through 12 are the completed Iteration 1 historical record, and Section 13 is
> the completed Iteration 2 historical record. Their inventories, compatibility
> decisions, future-work language, and execution instructions are descriptive
> evidence only and are superseded wherever Section 14 differs.

## 1. Scope and Source Posture

| Item | Posture |
| --- | --- |
| Target path | `internal/modules/links` |
| Target label | `links`, derived from the final path segment and normalized to lowercase kebab case |
| Output path | `docs/handoffs/links-module-refactor-tracker.md` |
| Status | Iterations 1 and 2 complete; Iteration 3 planned; implementation not started |
| Allowed change in this session | This tracker only. No implementation, owner specification, migration, generated artifact, harness input, or other documentation change is authorized by this document-update step. |
| Non-goals | No historical migration rewrite, hand-edited generated artifact, unrelated module redesign, frontend redesign, public route-shape change, authorization relocation, transaction-lifecycle relocation, or AC043 redistribution. |
| Current default posture | Production-hardening cutover: remove remaining legacy readers, open DTOs, duplicate validation, dead exports, and forwarding seams atomically. Preserve current canonical data and version-3 bundles; no migration or reset is expected. |
| Document class | NLSpec-style refactor execution tracker; not an adopted subsystem NLSpec or product-behavior owner |

The target exists. Iteration 1 began with the inventory recorded in Section 2 and
completed with the evidence in Sections 6 through 12. Section 13 records the
completed Iteration 2 cutover and evidence. Section 14 records the live
post-Iteration-2 inventory and the only current forward plan. The existence of a
historical seam or compatibility test is not, by itself, a reason to retain it.

Iterations 1 and 2 were explicitly authorized and completed. The Iteration 3
document plan was authorized on 2026-08-27. This tracker is the controlling
execution artifact. Owner specifications remain authoritative for behavior, and
each structural or corrective slice remains independently reviewable and
reversible.

### Execution checkpoint protocol

After every workstream, including a blocked or conditionally dropped branch, the
executor MUST update the active iteration work tracker and append exact handoff
evidence: changed files, commands, run roots, rollback posture, blockers, and
the next executable slice. The executor MUST then run `make lint-markdown`.
The next workstream MUST NOT begin until that checkpoint passes. A failed
checkpoint leaves the current workstream `IN_PROGRESS` or `BLOCKED`; it is not
completion evidence.

The current execution baseline is recorded in Section 14.1. Baseline statements
inside Sections 1 through 13 remain historical evidence for their completed
iterations and MUST NOT be used as current source-state claims.

### Normative language and evidence classes

The key words **MUST**, **MUST NOT**, and **MAY** are normative only for the
future refactor execution requirements identified as `LRT-REQ-*`, their
incorporated tables, and their acceptance criteria. They do not amend Core or an
adopted subsystem NLSpec. A proposed behavior remains blocked until its named
owner adopts it. After adoption, the owner text governs and this tracker is
supporting implementation material.

`MUST` and `MUST NOT` define binary refactor conformance. `MAY` defines an
optional implementation choice whose omission behavior is stated in the same
requirement. `Default` means the required outcome when an optional input is
omitted. Identifiers, field keys, schema IDs, link types, provenance tokens,
fixture IDs, contribution IDs, and error codes compare as exact code-point
sequences; the refactor MUST NOT trim, case-fold, label-resolve, or normalize
them unless an adopted owner requires an exact named transformation.

This tracker uses three evidence classes:

| Evidence class | Meaning | Authority consequence |
| --- | --- | --- |
| Current-state evidence | A directly inspected repository fact at the recorded commit | Descriptive only; it cannot resolve an owner contradiction. |
| Refactor execution requirement | A `LRT-REQ-*` rule for sequencing, interfaces, preservation, validation, or handoff | Normative for this authorized remediation. |
| Owner amendment | Exact behavior selected for Core or an adopted subsystem owner | Normative after adoption; O-01 and O-02 adoption evidence is recorded below. |

Decision closure and execution readiness are separate. The tracker statuses
remain exactly `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and
`DROPPED`. Phrases such as `RESOLVED — OWNER EDIT REQUIRED` describe decision
posture and MUST NOT be used as top-level execution statuses.

### Source hierarchy

1. Adopted subsystem NLSpecs, within their named subsystem. No adopted
   Links-specific subsystem NLSpec was found.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior plans, handoffs, and this framework as evidence only.

When owners conflict, this tracker uses `BLOCKED: owner contradiction` and does
not select a behavior. Phase maps and test rows are evidence-accounting inputs,
not runtime architecture. The planning framework supplies the template and
doctrine; every repository claim below was checked against the live tree.

### Owner documents inspected

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`
- `docs/reporting-subsystem-nlspec.md` (`adopted/current`, Reporting scope only)
- `docs/domain.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`

Core 01 and Core 02 supply the principal Links contracts. Core 03 governs
revision, projection, and Collaboration publication effects. Core 04 governs the
authorization outcomes enforced before Links capabilities are called. Core 05 is
relevant only to the AC043-sensitive performance-fixture contribution. The
Reporting Subsystem NLSpec governs canonical reporting export-model and
support-reference behavior within its named scope.

### Repository evidence inspected

- Every entry under `internal/modules/links`, individually reflected in Section 2.
- `db/migrations/00014_links.sql`.
- `tools/backend_module_boundaries.json`,
  `tools/schema_object_ownership_manifest.json`,
  `tools/test_support_inventory.json`,
  `tools/performance_fixture_snapshot_owner.json`,
  `tools/performance_fixture_builder_policy.json`, and
  `tools/test_catalog_owner.json`.
- `tools/test_families/module.links.json` and relevant cross-owner family files for
  Artifacts, Assessments, Incident Bundles, Projections, Reporting, Timeline, and
  Tasks and Decisions.
- `contracts/verification/owners/module.links.json`,
  `contracts/performance/ac043.v1.json`, relevant view-schema contracts, and the
  Incident Bundle, projection-provider, recovery, and revision registries.
- Direct Links callers under `internal/app/{assessmentassembly,importassembly,incidentportabilityassembly,recoveryassembly,revisionassembly,timelineassembly,workbookassembly}`.
- Direct Links callers under the Artifacts, Entities, Reporting, Revisions,
  Tasks and Decisions, and Timeline modules, plus performance-fixture assembly.
- Generated protocol and view-contract surfaces were searched as consumers only;
  no generated file is an edit source for this plan.

The inspected tree was clean at commit
`22d33ee5f197a00e52789625326226f82d7a04a3` on branch `main`, which was four
commits ahead of `origin/main` when this session began.

### Refactor requirement index and traceability

| Requirement | Required outcome | Behavioral owner or adoption gate | Workflow/slice | Acceptance IDs | Canonical Make evidence |
| --- | --- | --- | --- | --- | --- |
| `LRT-REQ-001` | Preserve authority, behavior-freeze, generated-file, caller-transaction, and upstream-authorization boundaries. | Existing Core 00–04 and repository procedure | All workflows | `LRT-AC-FREEZE-001..004` | Focused owner slices, boundary check, drift checks |
| `LRT-REQ-002` | Adopt and implement field-aware active-link identity without client storage overrides. | Core 01 and Core 02 repair; Core 04 acceptance additions | O-01, S-01, P-01, S-03, S-05, S-06A | `LRT-AC-FIELD-001..010` | Links store/integration rows plus affected owner slices |
| `LRT-REQ-003` | Remove link-local narrative from the current profile and reject it at exact-shape inputs. | Core 02 correction; Core 04/Incident Bundle acceptance | O-01, S-01, P-01, B-02 | `LRT-AC-NOTE-001..006` | Links, Incident Bundle, Reporting, drift, and migration checks as applicable |
| `LRT-REQ-004` | Delete `links/readshape` with no shim after the zero-consumer gate. | Repository implementation evidence; no Core amendment | S-02 | `LRT-AC-READSHAPE-001..003` | Links slice, boundary, generated drift, JSON shape, `test-fast` |
| `LRT-REQ-005` | Replace Links-to-Reporting DTO coupling with Links facts, an application adapter, and a Reporting-owned port. | Core 01 and/or adopted Reporting NLSpec amendment | O-02, S-01, P-01, S-06C | `LRT-AC-REPORT-001..010` | Reporting/Links focused evidence, boundary, canonical parity |
| `LRT-REQ-006` | Preserve the current AC043 fixture contribution and claim bindings exactly. | Existing Core 04, Core 05, and harness owners; no amendment | P-01, S-08, S-09 | `LRT-AC-AC043-001..004` | Fixture assembly tests and generated drift when relevant |
| `LRT-REQ-007` | Complete only independently reversible slices with retained handoff evidence. | Repository procedure and this tracker | S-01 through S-09 | `LRT-AC-DONE-001..006` | `make agent-finalize`, focused slices, risk-based broader checks |

### Owner-document and supporting-material allocation

| Topic | Required authoritative location | Supporting-only location |
| --- | --- | --- |
| Nullable `field_key`, field-aware active identity, merge/rollback/portability preservation | Core 02 | Migration SQL, Store/codec queries, test fixtures, this tracker |
| Field-routed derivation, field-scoped removal, and client override prohibition | Core 01 | Store and route adapter implementation notes |
| Negative conformance for field identity and absent link-local narrative | Core 04 and applicable owner acceptance sections | Exact test files, selectors, and run artifacts |
| Removal of `record_link.note` from the current profile | Core 02 | Preflight report and future-feature rationale |
| Source-owner-to-Reporting dependency direction | Core 01 and/or the adopted Reporting Subsystem NLSpec | Concrete Go types, package paths, and migration order |
| Reporting field/fact/support-ref schemas, ordering, and missing-reference behavior | Adopted Reporting Subsystem NLSpec | Application-adapter implementation notes and parity fixtures |
| `links/readshape` deletion | No Core amendment | This tracker, retained search evidence, and implementation handoff |
| Current AC043 topology | Existing Core 04, Core 05, and harness owners | This tracker and path-only implementation notes |
| Future AC043 redistribution | New Core 04/Core 05/harness owner revision | Migration procedure and benchmark rerun handoff |

## 2. Iteration 1 Baseline Repository Inventory (Historical)

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/links/.gitkeep` | Empty historical directory placeholder | None | None | None | None | None | Out of scope | Low | Explicitly out of scope because it has no content or runtime effect; remove only as incidental cleanup after a later authorized code move. |
| `internal/modules/links/assessment_facts.go` | Reads active Links-owned support references for Assessment projection facts | `AssessmentSupportFacts`, `AssessmentFactReader.LoadSupportFactsTx` | Assessment assembly projection source | `pgx`, active `record_links` data | Assessment contract and integration tests | Assessment view-schema support-reference fields | Links, with an Assessment assembly adapter | Medium | Keep the source query Links-owned; hide it behind a narrower contribution/provider seam. |
| `internal/modules/links/collection_actions.go` | Validates and applies record-reference, party-reference, and tag collection actions | Collection validation/command/result input types, `CollectionValidationError`, Store validation/apply methods | Artifacts and Tasks and Decisions mutation facades; Links tests | Store persistence, `pgx`, UUIDs | `links_tags_test.go`; cross-owner mutation tests | Workbook collection-action and view-field contracts | Links | High | Legitimate Links mutation semantics. Preserve typed validation errors, add/remove ordering, normalization, and idempotence. |
| `internal/modules/links/collection_mutations.go` | Wraps collection changes with canonical before/after mutation values for Revisions | `CollectionMutationResult`, `RecordLinkMutation`, `RecordTagMutation`, mutation-aware Store methods | Artifacts and Tasks and Decisions mutation paths | Links store/value codec, `pgx` | Links integration tests; mutation composition and revision tests | Revision target semantics and `record_changed` field invalidation inputs | Links | High | Legitimate mutation coordinator inside the source owner; preserve transaction and mutation ordering. |
| `internal/modules/links/commands.go` | Public typed command DTOs for link upsert, field-reference sync, and supersedes insertion | `LinkType`, `LinkProvenance`, `UpsertLinkCommand`, `SyncFieldReferenceCommand`, `InsertSupersedesCommand` | Store callers across owner modules and application assemblies | UUID/time types | Links tests and cross-owner integration tests | Record-link storage and route-derived mutation contracts | Links facade | High | Keep a small stable command surface while private mechanics move. |
| `internal/modules/links/field_refs.go` | Persists field-keyed record/party references and record tags; validates active targets | Store field-reference methods, `TagStore`, `NewTagStore`, tag mutation methods | Workbook mutation capabilities, Tasks and Decisions, Artifacts | PostgreSQL transaction and authoritative Records checks | `links_tags_test.go`; saved-view, timeline, entity, and task/decision tests | View-schema collection fields, active-record-link/tag views | Links | High | Source-owned persistence. Its field-aware identity is now adopted; implementation changes remain gated by S-01/P-01 evidence. |
| `internal/modules/links/incident_bundle_portability.go` | Exports/imports Links source rows and validates portable link/tag data | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Incident Bundle source-port adapter | Incident portability contracts, JSON, `pgx` | Incident Bundle owner tests through the catalog | Incident Bundle source catalog and file paths | Links internal portability provider | High | Retain Links source ownership; internalize mechanics behind the source-port contribution. |
| `internal/modules/links/incident_bundle_source_port.go` | Publishes the Links owner port to Incident Bundle orchestration | `NewIncidentBundleSourcePort` | `internal/app/incidentportabilityassembly/catalog.go`, import owner registry | Incident Bundle `sourceport` | Incident Bundle catalog/import tests | `contracts/incident-bundles/source_catalog.json` | Links thin contribution | Medium | Keep as a narrow public contribution; implementation details belong behind it. |
| `internal/modules/links/item_refs.go` | Encodes and strictly parses canonical collection item references | `RecordRefItemRef`, `ParseRecordRefItemRef`, `PartyRefItemRef`, `ParsePartyRefItemRef`, `RecordTagItemRef`, `ParseRecordTagItemRef` | Links collection logic and cross-owner mutation callers | UUID parsing | `item_refs_test.go`; mutation tests | Workbook collection-action `item_ref` wire values | Links facade | High | Preserve exact prefixes, separators, UUID canonicalization, and parse failures. |
| `internal/modules/links/item_refs_test.go` | Characterizes canonical and strict Links item-reference encoding | Test-only surface | Go test runner | Links item-ref API | Self | Collection-action item-reference contract | Links tests | Low | Preserve and extend if public facade movement changes call paths. |
| `internal/modules/links/links_tags_test.go` | Covers SQL view shape, owner validation, typed links/tags, HTTP mutation/query, projection, revision, and WebSocket effects | Test-only constants/helpers and four major test cases | Go test runner and service-backed harness | App support, database, HTTP/WebSocket test helpers | Self | Module Links verification rows and observable route/event contracts | Links tests | High | Existing broad characterization is valuable but spans unit, store, and integration concerns. Keep behavior assertions stable during movement. |
| `internal/modules/links/merge_effects.go` | Repoints and deduplicates links/tags during Entity merge and emits revision/invalidation mutations | Merge command/result/mutation types; `Store.RepointMergedLinksTx`, `Store.RepointMergedTagsTx` | Entities merge ports/adapters | Links SQL/value codec, record-type invalidation mapping, `pgx` | Entity merge/support tests | Entity merge route effects, revisions, projections | Links internal merge-effects provider | High | The data effects are source-owned; expose only a narrow capability to Entities. Preserve counts, ordering, dedupe, and invalidations. |
| `internal/modules/links/readshape/sql.go` | Builds SQL expressions for active Links views and canonical item refs | Five exported SQL helper functions | No live production, test, generator, or configuration caller found; only historical documentation/archive references remain | String construction only | No direct test found | Active view and item-ref shape if adopted | No owner after the evidence gate | Medium | `LRT-REQ-004` requires deletion without a compatibility shim after the final zero-consumer gate. |
| `internal/modules/links/recovery_state.go` | Declares authoritative Links tables to Recovery | `RecoveryStateContribution` | Recovery assembly state catalog | Recovery state contract | Recovery catalog/fixture tests | Recovery state catalog fixtures/registries | Links thin contribution | Medium | Keep the contribution public and table list unchanged. |
| `internal/modules/links/reportingprovider/provider.go` | Collects Links support references and reporting fields/facts | `CollectSupportRefsTx`, `CollectFieldsTx`, `CollectFactsTx` | Reporting export materializer | Reporting export-provider DTOs, Links SQL, `pgx` | Reporting evidence-provider integration tests | Reporting export field/fact contracts | Links facts plus application adapter plus Reporting-owned port | High | `LRT-REQ-005` selects the joint design. Migration remains blocked until owner-boundary adoption and method-by-method parity evidence. |
| `internal/modules/links/revision_provider_contribution.go` | Registers the Links rollback provider for `record_link` and `record_tag` targets | `RevisionProviderContribution` | Revision assembly | Revisions catalog and Links revision provider | `revision_provider_contribution_test.go`; Revisions integration tests | Revision target-semantics registry | Links thin contribution | High | Keep this narrow public entry point while provider mechanics become private. |
| `internal/modules/links/revision_provider_contribution_test.go` | Proves the contribution owns current Link and Tag history targets | Test-only surface | Go test runner | Revisions catalog, Links contribution | Self | Module Links unit verification row | Links tests | Low | Preserve as the facade ownership check. |
| `internal/modules/links/revisionprovider/provider.go` | Describes and inverses non-row link/tag revisions, including restore/tombstone and affected-field calculation | `Provider`, `NewProvider`, validation/load/restore/tombstone methods, rollback-provider methods | Links contribution; no external constructor caller found | Revisions rollback contracts, Links value codec, SQL/`pgx` | Contribution and Revisions integration tests | Revision history/rollback envelopes and projection invalidations | Links internal revision provider | High | Public subpackage is broader than current callers require. Internalize without changing target descriptors, legacy decode, or field invalidations. |
| `internal/modules/links/store.go` | Core active record-link validation and persistence, tombstoning, and supersedes checks | `Store`, errors, link/provenance constants, `RecordLink`, `SupersedesLink`, constructor and Store methods | Application assemblies, Artifacts, Tasks and Decisions, Entities, Timeline/Assessment adapters | PostgreSQL transaction, Records table/type checks, UUID/time | Links tests and many cross-owner integration tests | `record_links`, `active_record_links_v1`, link-type contracts | Links facade and private source adapter | Critical | Central legitimate owner capability. Preserve SQL semantics, errors, idempotence, transaction ownership, and canonical link direction. |
| `internal/modules/links/testsupport/links.go` | Owner-local database seeds, lookups, and assertions for links/tags | Fixture/expectation types and test helper functions | Links and cross-owner tests | PostgreSQL test DB and testing assertions | Test-only consumers | `tools/test_support_inventory.json` | Links test support | Low | Intentional owner-local test support; runtime excluded and no service starts. |
| `internal/modules/links/testsupport/performancefixture/production.go` | Production-backed validator for expected link, tag, and Entity-mention fixture counts | `ProductionApplication`, constructor, `ValidateFixtureAssociations` | Performance-fixture assembly | PostgreSQL DB; authoritative Links rows plus Entities `entity_mentions` | Performance-fixture assembly tests | AC043 fixture counts and claim evidence | Current Links test support | High | `LRT-REQ-006` resolves ownership posture as no change. Cross-owner counting is harness accounting, not runtime ownership. |
| `internal/modules/links/testsupport/performancefixture/provider.go` | Contributes `links.timeline_associations.v1` fixture validation to the builder | `Expectations`, `Application`, `Provider`, `New`, descriptor/apply methods | Performance-fixture assembly | Fixture descriptor/build state | Performance-fixture assembly tests | `tools/performance_fixture_snapshot_owner.json`, AC043 fixture | Current Links test support | Critical | Contribution identity, dependency, counts, receipts, order, digest inputs, snapshot key, and verification bindings MUST remain unchanged. |
| `internal/modules/links/timeline_facts.go` | Reads Links facts and returns Timeline workbook-projection DTOs | `TimelineFacts`, `TimelineFactReader.LoadTx` | Timeline assembly/projection source | `timeline/workbookprojection`, Links SQL, `pgx` | Timeline projection/store/resolution tests; Links integration test | Timeline projection and view-schema fields | Links facts with assembly conversion | High | `must_fix`: root Links code imports a consumer module DTO. Define Links-owned facts and convert at the composition edge. |
| `internal/modules/links/timeline_history_facts.go` | Determines Links-owned Timeline collection fields changed at a revision time | `Store.LoadTimelineCollectionFieldsChangedTx` | Timeline history/revision composition | Links mutation tables, `pgx` | Timeline and Revisions integration tests | History `fields_changed` and projection invalidation behavior | Links internal history-facts provider | High | Keep source query semantics Links-owned; hide consumer-specific orchestration behind a narrow port. |
| `internal/modules/links/value_codecs.go` | Adapts Store methods to canonical link/tag mutation-value loading | Four exported Store load methods | Collection mutation and revision paths | Links `valuecodec`, `pgx` | Links integration and codec tests | Revision before/after value maps | Links facade over internal codec | High | Preserve map shapes and error behavior while internalizing codec mechanics. |
| `internal/modules/links/valuecodec/valuecodec.go` | Builds, decodes, and validates canonical and legacy non-row mutation values and restore plans | Mutation-value, identity, restore-plan, and input types; load/build/decode/parse helpers | Links Store, merge effects, revision provider | SQL/`pgx`, JSON, UUID | `valuecodec_test.go`; revision tests | Revision target values and rollback compatibility | Links internal mutation codec | High | Move behind Links-private APIs in stages. Preserve `field_key`, compact legacy link shapes, and the legacy `tag_id` alias. |
| `internal/modules/links/valuecodec/valuecodec_test.go` | Characterizes canonical mutation maps and legacy decode/restore compatibility | Test-only surface | Go test runner | Links value codec | Self | Revision compatibility behavior | Links tests | Medium | Must move with the codec and retain all legacy fixtures. |

No file in the target is silently excluded. The only explicit non-code exclusion is
the empty `.gitkeep` row.

## 3. Module Boundary Diagnosis

The current target is a **legitimate Links-and-Tags source-owner module with a
mixed-responsibility package surface**. It is not an accidental catch-all, a
frontend shell, or a grid-vendor integration layer. It combines a valid
persistence/mutation core with projection facts, merge effects, portability,
rollback, reporting, recovery, and test contributions that should be hidden
behind smaller owner facades.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Typed relationship and tag validation/persistence | Links root Store and collection files | Links | Keep, then split facade from private source adapter | Core 02; migration `00014`; schema ownership manifest; callers/tests | Central design decision hidden by the module. Do not change identity or SQL behavior. |
| Canonical collection item refs and mutation values | `item_refs.go`, `value_codecs.go`, `valuecodec/` | Links | Keep item-ref facade; move codec implementation private | Workbook and Revision contracts; codec tests | Legacy decode is observable rollback compatibility. |
| Timeline projection facts | `timeline_facts.go` | Links facts plus Timeline/application adapter | Split | Direct import of `timeline/workbookprojection`; timeline assembly callers | Remove peer DTO dependency; preserve facts, ordering, and nil/empty behavior. |
| Assessment support facts | `assessment_facts.go` | Links facts plus Assessment/application adapter | Split | Assessment assembly and projection tests | Source query stays Links-owned; consumer DTO adaptation belongs at composition. |
| Timeline history field facts | `timeline_history_facts.go` | Links private history provider | Move internally | Timeline/Revisions integration | Preserve exact changed-field calculation. |
| Entity merge repoint/dedupe | `merge_effects.go` | Links internal merge-effects provider, exposed through an Entities-facing port | Split | Entities merge adapters/tests | Links owns its rows; Entities owns merge orchestration. |
| Revision inverse application | `revisionprovider/` and root contribution | Links private provider plus thin Revisions contribution | Split | Revision catalog and integration tests | Keep contribution; internalize provider and mapping mechanics. |
| Incident Bundle source portability | Portability files | Links private provider plus thin Incident Bundle contribution | Split | Source catalog and assembly | Preserve file names, ordering, validation, attribution, and atomic import. |
| Recovery table registration | `recovery_state.go` | Links thin contribution | Keep | Recovery catalog fixtures | No reason to move table meaning into Recovery. |
| Reporting support references/facts | `reportingprovider/` | Links source facts, application-composed adapter, Reporting-owned port/DTOs | Split after characterization | Reporting materializer, adopted Reporting NLSpec, and integration tests | The joint design is adopted by LRT-REQ-005; implementation is gated by characterization and parity. |
| Owner-local test support | `testsupport/links.go` | Links test support | Keep | Test-support inventory | Intentional and excluded from runtime. |
| AC043 fixture association validation | `testsupport/performancefixture/` | Current Links fixture contribution | Keep | Snapshot owner manifest and Core 05 AC043 contract | RB-005 is resolved as no change. Harness ownership is evidence accounting, not runtime architecture. |
| Active-view SQL helper strings | `readshape/sql.go` | None unless the final gate discovers a live consumer | Move/delete | Repository-wide tracked-source search found no runtime caller | Delete without a shim after LRT-AC-READSHAPE-001 passes. |
| Authorization | Route/source facades outside Links | Workbook, Entities, Revisions, Collaboration, and Security owners | Keep outside Links | Core 04 and inspected route/application flows | Links receives authorized transaction-scoped calls; do not duplicate transport authorization here. |
| Frontend collection rendering/controller state | Generated contracts and generic web collection handlers | Web shell/view contracts/grid adapter | Defer/no Links move | No direct Links frontend or grid-vendor import found | Only revisit if an owner-approved public contract changes. |

## 4. Public Contract and Behavior Freeze Map

The adopted owner repair changes no public API, schema, route, WebSocket event,
or generated interface.

### LRT-REQ-002: field-aware active-link identity

Core 01 and Core 02 now adopt the field-aware owner repair. Implementation
movement remains gated by S-01 characterization and P-01 preflight.

A field-routed active link MUST have identity
`(incident_id, src_record_id, dst_record_id, link_type, field_key)`. The stored
`field_key` MUST equal the exact canonical field key validated from the active
view contract. Two null `field_key` values compare equal for identity. Two
otherwise-identical bindings with different non-null field keys are distinct.

| Creation or lifecycle path | Required `field_key` | Active identity and default | Client authority | Removal or collision scope |
| --- | --- | --- | --- | --- |
| View-schema collection mutation | Exact validated active field key; non-null | Full five-member field-aware tuple | Client supplies only ordinary `changes[].field_key`; no storage override | Only the active field binding identified by its valid `item_ref` |
| Explicit action route with no owning collection field | `null` | Four non-null identity members plus the single null-field identity value | Route owner derives type/direction; client supplies no storage member | Route-specific target identity |
| Timeline or Decision supersession | `null` in the current routes | Null-field tuple plus the independent unique active superseded-destination rule | Client supplies business target IDs only | One active `supersedes` destination; one source MAY supersede multiple destinations |
| Revision rollback or restore | Exact original value, including `null` | Restored row re-enters its original identity | No override | Full restored identity |
| Incident Bundle import/export | Exact source value, including `null` | Round trip MUST preserve identity | Bundle shape does not translate labels to field keys | Full imported identity |
| Entity merge repoint | Exact original value, including `null` | Collision detection uses `IS NOT DISTINCT FROM` field semantics | No override | Deduplicate only a full field-aware collision |

Merge-time survivor selection MUST be deterministic because `record_link_id` is
history and rollback identity. S-01 MUST characterize the current ordered
survivor rule before S-05 moves it. The refactor MUST NOT select a survivor from
unspecified database return order.

Counts MUST use one of the exact meanings below. A producer MUST name the count
kind and MUST NOT substitute another kind silently.

| Count kind | Exact definition |
| --- | --- |
| Binding count | Number of active `record_link` rows; different field keys count separately. |
| Field collection count | Number of active bindings whose `field_key` equals the requested exact field key. |
| Distinct-target count | Number of distinct target record IDs after applying a separately named direction and link-type filter. |

The live migration already has one unique active partial index over the full
tuple for non-null keys and one unique active partial index over the four
non-null identity members for null keys. Merge collision lookup uses
`field_key IS NOT DISTINCT FROM`; revision values and Incident Bundle rows carry
`field_key`. Therefore the required default preflight result is `no migration`.
P-01 MUST fail closed if live DDL or data differs. Any correction MUST be a new,
separately authorized forward migration; historical migration files MUST NOT be
rewritten, and no preflight MAY discard or choose a survivor automatically.

### LRT-REQ-003: no link-local narrative

Core 02 now adopts the no-narrative correction. The current profile MUST NOT
define or accept
`record_link.note`, `record_link.description`, `record_link.comment`, or an
equivalent link-local free-text extension. Narrative requiring authorship,
history, search, reporting, or reuse belongs to an owning source record, a
Decision, or an Artifact related through a typed link.

| Surface | Required current-profile shape | Omission/default | Invalid-input behavior |
| --- | --- | --- | --- |
| Database and Store DTO | No narrative column or member | Absent; no reserved null field | A discovered legacy column or data blocks the no-migration path. |
| Public collection/action input | Closed owner action shape, no narrative member | Omission is the only valid state | Unknown narrative member fails `400 invalid_mutation_payload`; it is not ignored. |
| Mutation/revision/rollback value | No narrative member | Absent in canonical and legacy output | An unrecognized narrative member MUST NOT become restored state. |
| Incident Bundle link row | Exact owner-approved row shape without narrative | Absent | Unknown narrative data MUST fail exact-shape validation rather than be ignored. |
| Snapshot and Reporting value | No link-local narrative field | Absent | An implementation-local column or extension map MUST NOT be surfaced. |
| Generated protocol/view contracts | No link-local narrative member | Absent | Drift adding such a member fails generated/shape validation. |

Current fixed-row Incident Bundle import proves required-column presence but does
not prove rejection of unknown members before `jsonb_populate_record`. B-02 MUST
add the owner-approved negative characterization and any required exact-shape
correction; S-06A MUST NOT claim LRT-REQ-003 complete merely because no column is
populated.

A future profile MAY introduce link-local narrative only after one adopted owner
revision defines, together, its scalar type and maximum length; omission,
explicit-null, and empty-value behavior; create/edit authorization; conflict and
idempotency behavior; revision and rollback form; portability version; snapshot,
Reporting, disclosure, and redaction classification; search/index behavior; and
migration/compatibility policy. Until then no ignored compatibility field,
reserved field, or local JSON extension is permitted.

### LRT-REQ-004: `readshape` zero-consumer deletion

`internal/modules/links/readshape` MUST be deleted without a compatibility shim
after LRT-AC-READSHAPE-001 passes against the implementation baseline. The final
gate MUST search tracked Go and non-Go sources for the package path, directory,
five exported symbols, string literals, generators, templates, Make inputs,
JSON/YAML/TOML configuration, build-tag profiles, testdata, embed directives,
architecture allowlists, and generated-owner inputs.

Current external matches are historical projection handoff text and an archived
binary document; neither is a runtime consumer. A newly discovered consumer does
not authorize a generic shim. Its semantic owner MUST characterize the required
shape, move it behind an owner-typed interface, migrate the consumer, and then
delete `readshape`.

Direct `go list` output MAY support investigation but is not canonical repository
validation. The deletion evidence MUST use the Make-owned checks in Sections 7
and 8.

### LRT-REQ-005: Links-to-Reporting interface

The required dependency direction is:

```text
Links-owned immutable source facts
        -> application reporting adapter
        -> Reporting-owned provider port and DTOs
        -> Reporting materializer and canonical export model
```

Only application composition MAY import both the Links fact interface and the
Reporting provider contract. Links MUST NOT import Reporting DTOs or select
report sections, labels, redaction classes, support-reference IDs, or canonical
serialization. Reporting MUST NOT query `record_links` or `record_tags`
directly. The adapter MUST NOT issue source SQL, start or commit a transaction,
or duplicate materializer orchestration. No neutral `common/reporting` package
may become a third DTO authority.

The Links reader accepts `(context, caller_transaction, incident_id)` and returns
one closed immutable fact set or one typed owner error. It MUST NOT return a
partial fact set, expose SQL text/database details, perform authorization,
perform redaction, or select report content. Empty link or tag families default
to empty ordered collections, not omitted unknown state.

| `RecordLinkFact` member | Type | Nullable | Rule |
| --- | --- | --- | --- |
| `record_link_id` | UUID | No | Stable mutation and Reporting path identity. |
| `src_record_id` | UUID | No | Canonical source endpoint. |
| `dst_record_id` | UUID | No | Canonical destination endpoint. |
| `link_type` | closed Core 02 token | No | Exact persisted token. |
| `field_key` | exact canonical field key | Yes | `null` means explicit route binding without an owning field. |
| `provenance` | closed Core 02 token | No | Exact persisted token. |
| `confidence` | integer `0..100` | Yes | `null` is preserved; current manual links require null. |
| `owner_user_id` | UUID | No | Current owner attribution. |
| `created_by_user_id` | UUID | No | Creation attribution. |
| `decided_at` | timestamp with time zone | No | Exact source instant. |
| `created_at` | timestamp with time zone | No | Exact source instant. |

| `RecordTagFact` member | Type | Nullable | Rule |
| --- | --- | --- | --- |
| `record_tag_id` | UUID | No | Stable tag-assignment and Reporting path identity. |
| `record_id` | UUID | No | Tagged record. |
| `tag_name` | string | No | Exact stored display form. |
| `normalized_tag_name` | string | No | Exact stored normalized identity. |
| `created_by_user_id` | UUID | No | Creation attribution. |
| `created_at` | timestamp with time zone | No | Exact source instant. |
| `updated_at` | timestamp with time zone | No | Exact source instant. |

`incident_id` is an interface input and MUST NOT be duplicated in the Reporting
value object. The fact variants represent active output; tombstone members do
not belong to their conceptual type. To preserve the current Reporting schema,
the adapter MUST emit `deleted_at: null` and `deleted_by_user_id: null` in each
active value object where current canonical output contains those members.

Before S-06C, P-01 MUST compare the current Reporting row set with Core 02 active
semantics, which require the link row and both endpoint records to be active. If
the sets differ for any supported case, S-06C is blocked: an independently
authorized normative-correction slice must land before the final active-fact
reader. Structural adapter migration MUST NOT hide an endpoint-eligibility
behavior change.

The application adapter MUST apply this exact mapping:

| Source fact | Reporting field path | Source family | Content class | Value and support behavior |
| --- | --- | --- | --- | --- |
| Link | `/relationships/{record_link_id}` | `record_link` | `derived_analytic` | Exact current link value members except `incident_id`; explicit active tombstone nulls; support refs from the Reporting-owned support map keyed by path identity. |
| Tag | `/tags/{record_tag_id}` | `record_tag` | `derived_analytic` | Exact current tag value members except `incident_id`; explicit active tombstone nulls. |

The provider key MUST equal `links`. Final fields MUST sort by path ascending.
Support-edge derivation MUST include exactly active links whose `link_type` is
`supported_by`, `references_record`, or `attached_evidence`; process source ID
ascending then destination ID ascending; use a non-empty Reporting logical target
when one exists and otherwise `/record_envelopes/{dst_record_id}`; and preserve
the current repeated-target behavior until the Reporting owner adopts a different
canonical duplicate rule. A nil or missing logical-target entry uses the same
fallback. Zero support edges return an empty map.

`CollectSupportRefsTx`, `CollectFieldsTx`, and `CollectFactsTx` MUST migrate one at
a time. The old implementation remains until that method has canonical-model or
byte parity for included link types, tags, row eligibility, ordering, duplicates,
nulls, missing targets, support IDs, and errors.

### LRT-REQ-006: AC043 semantic freeze

RB-005 is resolved as `DONE — no change`. The fact that the contribution checks
an Entities-owned mention count is harness accounting and MUST NOT be interpreted
as runtime ownership.

| Frozen member | Exact value |
| --- | --- |
| Fixture profile | `ac043_large_grid_snapshot_v1` |
| Fixture version / fixture ID | `cartulary.perf.large_grid.v1` |
| Seed | `20260405` |
| Contribution ID and version | `links.timeline_associations.v1` |
| Owner | `module.links` |
| Dependency | `timeline.large_grid.v1` |
| Source contract | `cartulary.performance.ac043.v1` |
| Receipt count `links` | `1000` |
| Receipt count `mentions` | `1000` |
| Receipt count `tags` | `1000` |

Contribution order, dependency graph, receipt identity and shape, semantic digest
inputs, generated snapshot key, verification bindings, and claim bindings MUST
remain unchanged. Test-support files MAY move within Links only when all semantic
artifacts remain identical. A future redistribution to Entities is a new fixture
contract and requires a new fixture identity/version, new contributions and
order, regenerated descriptors/key vectors, amended Core 04/Core 05/harness
bindings, rebuilt claim artifacts, and classification of prior results as
historical evidence.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `record_links` identity, types, provenance, confidence, active/tombstone state | Links under Core 02 | Core 02; `store.go`; migration and schema manifest | Links store/integration tests | Preserve exact error, idempotence, direction, and active-row cases before Store movement | Critical | LRT-REQ-002 field-aware identity is adopted; implementation waits on characterization and preflight. |
| Field-keyed record/party references | Links with Workbook field routing | Core 01 collection rules; `field_refs.go`; migration indexes | Links, Artifacts, Timeline, Tasks and Decisions tests | Multi-field same-target, removal, concurrency, merge, rollback, and portability cases | Critical | Existing SQL is the expected no-migration projection of LRT-REQ-002. |
| Tags and `record_tags` mutation identity | Links | Core 02; field/tag Store; active view | Links store/integration and codec tests | Preserve normalization, catalog reuse, add/remove, tombstone, and item-ref identity | High | Canonical item ref is `record_tag:<record-id>:<record-tag-id>`. |
| Collection action request/result behavior | Workbook transport; Links source mutation | Core 01 and Links collection commands | Workbook-facing Links integration; cross-owner mutation tests | Preserve validation errors, atomicity, save state, and conflict outcomes | High | Links owns source semantics, not the HTTP envelope. |
| Query, row creation, patch, linked-notes, and supersede routes | Workbook and Tasks and Decisions | Route assembly/capabilities and Core 01 | Links integration, Tasks and Decisions and saved-view tests | Existing route shapes/envelopes must remain black-box stable | High | No Links-owned route handler was found. |
| Entity merge and entity-mention resolution effects | Entities orchestration; Links rows | Entity merge/mention ports and Links merge effects | Entities unit/support/resolution tests | Characterize repoint, dedupe, counts, mutation order, and invalidation fields | High | Keep the transaction caller-owned. |
| Canonical `item_ref` strings | Links source contract consumed by Workbook | `item_refs.go`; Core 01 collection actions | `item_refs_test.go` and mutation tests | Existing strict parser tests are sufficient; add cross-facade tests only if API moves | High | Exact prefixes and UUID shapes are frozen. |
| Revision before/after mutation maps and legacy decode | Links values; Revisions coordination | value codecs and revision target registry | Codec tests and Revisions integration tests | Preserve legacy `tag_id`, compact link shape, defaults, and canonical emitted maps | Critical | Never rewrite old stored revision values during structural refactor. |
| History and rollback route behavior | Revisions | Core 03; Links revision contribution/provider | Revisions integration and Links integration tests | Preserve descriptors, affected record IDs/fields, restore/tombstone, and responses | Critical | Provider implementation may move; target ownership may not drift. |
| Projection refresh and field invalidation | Projection owners, sourced by Links facts/mutations | Core 01/Core 03; Timeline/Assessment adapters | Links, Timeline, Assessment, and Projections tests | Preserve field ordering, affected views, refresh timing, nil/empty distinctions | High | Fact DTO movement must be adapter-only. |
| Saved-view and view-schema behavior | Workbook/Saved Views/View Contracts | Authored view schemas and generated adapters | Saved Views and module integration tests | Preserve field keys, types, capabilities, and stored-query compatibility | High | Do not infer storage ownership from visible column labels. |
| WebSocket `/ws/v1/incidents/{incident_id}` and replayable `record_changed` | Collaboration | Core 03; collaboration session and Workbook mutation publication | Links integration and Collaboration/browser tests | Preserve path, authorization recheck, incident scope, affected fields/views, ordering | Critical | Links contributes mutation facts; Collaboration owns publication. |
| Authorization outcomes | Security plus owning route/application facade | Core 04 and route/application composition | Route and integration authorization tests | Retain incident membership/role checks and deployment-admin non-bypass | Critical | Do not move transport authorization into Links. |
| Incident Bundle link/tag export/import | Incident Bundles orchestration; Links source port | Source catalog, portability files, Core 01 portability | Incident Bundle family | Preserve file paths, deterministic export, validation, attribution, and atomic import | High | Generated/source catalogs must be updated only through owners if the contribution changes. |
| Reporting support refs and fields | Reporting orchestration; Links source facts | Reporting provider/materializer and adopted Reporting NLSpec | Reporting evidence-provider integration | Characterize all three current methods and canonical output before seam change | Critical | Reporting §7.1.2 and Core 01 REQ-01-664 adopt the boundary; parity remains the implementation gate. |
| Recovery authoritative tables | Recovery catalog; Links table meaning | `RecoveryStateContribution`; recovery fixtures | Recovery catalog/fixture tests | Existing catalog coverage is sufficient unless table list changes | Medium | Keep `record_links` and `record_tags` registration stable. |
| Generated protocol/view contracts and UI selectors | Their authored contract owners | Generated TypeScript searches and view-schema contracts | Frontend unit/browser families | Run drift checks if authored inputs change; no hand edits | High | No direct grid-vendor coupling was found in Links. |
| Module Links harness accounting | Verification/harness owners | `module.links` owner and family JSON | Four active rows: browser, integration, store, unit | Add/update authored rows only when coverage ownership or paths change | Medium | Test rows prove execution/evidence accounting, not runtime boundaries. |
| AC043 fixture contribution | Core 05 and harness fixture owner | AC043 contract and snapshot-owner manifest | Performance-fixture assembly tests | Prove semantic artifacts unchanged after any path-only movement | Critical | LRT-REQ-006 requires no topology or expectation change. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Links root imports Timeline workbook-projection DTOs. | `timeline_facts.go` imports `internal/modules/timeline/workbookprojection`. | High | `must_fix` | Links facts plus application adapter | Introduce Links-owned immutable fact types, convert in timeline assembly, and characterize projection output first. |
| Links exposes broad provider and codec subpackages although current construction is owner-local. | `revisionprovider/`, `valuecodec/`, root Store methods and direct importer scan. | High | `should_fix` | Links private implementation plus thin contributions | Internalize one provider at a time; retain facade contracts until all callers migrate. |
| `readshape` exports SQL builders with no discovered caller. | Repository-wide symbol/import search. | Medium | `should_fix` | No owner if confirmed dead | Repeat the no-consumer search, run focused tests, then delete in an isolated slice. |
| Reporting provider imports Reporting-owned DTOs. | `reportingprovider/provider.go`, Reporting materializer, and adopted Reporting owner. | High | `must_fix` | Links facts, application adapter, Reporting-owned port | Apply adopted LRT-REQ-005 only after characterization and method parity. |
| Revision provider knows Timeline/Evidence/Decision invalidation fields. | `revisionprovider/provider.go` affected-field logic. | High | `should_fix` | Links source-owned rollback semantics behind a narrow mapping contribution | Characterize all target descriptors and invalidations; internalize mapping without changing output. |
| Performance fixture validates Entity mentions as part of a Links contribution. | Performance-fixture provider/production code and snapshot-owner manifest. | Critical | `intentional/no_action` | Existing harness contribution | LRT-REQ-006 freezes the current topology; no redistribution work exists. |
| Direct SQL and `pgx.Tx` occur in Links source logic. | Store, facts, merge, codec, and portability files. | Medium | `intentional/no_action` | Links source adapter | Preserve caller-owned transactions; do not introduce a generic transaction facade without an owner decision. |
| Authorization is performed upstream rather than in Links. | Core 04 plus route/application composition; no auth transport import in Links. | Critical | `intentional/no_action` | Route/source facades and Security owners | Preserve authorization outcomes and avoid duplicate checks in the source store. |
| No direct grid-vendor import exists outside the grid adapter for this surface. | Repository search for direct grid-vendor imports and Links consumers. | Low | `intentional/no_action` | Grid adapter | Keep frontend/grid work out of this refactor unless new evidence appears. |
| No duplicate alternate Links persistence implementation was found. | SQL ownership manifest, migration, and caller scan. | Low | `intentional/no_action` | Links | Keep the authoritative store centralized; duplication in DTO adaptation is assessed separately. |
| Generated artifacts expose affected collection/view contracts. | Generated protocol/view-contract searches and generated-artifact policy. | High | `intentional/no_action` | Authored contract owners and generators | Never hand-edit generated outputs; change owner inputs and run Make generation/drift targets if authorized. |
| Legacy codec assumptions are deliberately retained for stored revisions. | `valuecodec_test.go` legacy alias/compact-shape cases. | Critical | `intentional/no_action` | Links and Revisions | Treat legacy decode as public rollback compatibility, not removable test-only behavior. |
| Link uniqueness clauses disagreed on whether `field_key` participates. | Core 02 REQ-02-168 versus Core 01 field-routed collections and migration indexes. | Critical | `fixed_in_owner` | Core 01/Core 02 owners | O-01 adopted field-aware identity; P-01 still must prove the no-migration baseline. |
| Core 02 described optional `record_link.note`, absent from live schema/code. | Core 02 record-link contract versus migration, Store DTOs, codecs, and generated surfaces. | High | `fixed_in_owner` | Core 02 plus affected acceptance owners | O-01 removed narrative from the current profile; preflight and exact-shape validation remain. |
| Fixed-row Incident Bundle import does not prove rejection of unknown link members. | `ImportFixedRows` validates required columns, then uses owner SQL with `jsonb_populate_record`. | High | `must_fix` | Links source shape plus Incident Bundle validation | Characterize unknown `note` input and implement only the owner-approved exact-shape correction in B-02. |
| Current Reporting link queries filter link tombstones but do not join endpoint activity. | `CollectSupportRefsTx` and `CollectFactsTx` query `record_links` with `deleted_at IS NULL`; Core 02 REQ-02-169 requires active endpoints for ordinary export. | High | `must_fix` | Core 02 source semantics plus Reporting | P-01 compares row sets. If they differ, separate normative correction B-04 from structural S-06C. |

No test-only helper was found imported by Links production code. No generated file
is proposed as an edit source.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Fix scope, authority, commit, constraints, and output. Discovery is complete. | This tracker and owner sources | `make lint-markdown` for tracker revisions | Scope, baseline, staging posture, and only-write rule are recorded. |
| WF-01 | Target inventory and contract scan | chain | WF-00 | WF-02 | Account for every target entry, caller, dependency, behavior, owner, test, and generated surface. Discovery is complete. | All 27 target entries, callers, contracts, manifests | Task guide, owner explanation, repository searches | Sections 2–5 have no unaccounted target or discovered contract. |
| WF-02 | Closure-decision and owner-repair specification | chain | WF-01 | WF-03, WF-04 | Specify RB-001/RB-002 owner repairs and RB-005 no-change closure without treating tracker text as adopted behavior. | Core 01, Core 02, Core 04, Core 05, owner navigation | Owner review plus affected acceptance mapping | O-01 is adoptable, exact, and implementation remains blocked until adoption. |
| WF-03 | Reporting boundary owner specification | parallel | WF-01 | WF-04, WF-06 | Adopt LRT-REQ-005 across Core 01 and/or the Reporting NLSpec without changing public route/export schemas. | Reporting NLSpec, Core 01, Reporting materializer, Links provider | Reporting owner review and parity-fixture design | O-02 defines one owner per fact, DTO, adapter, ordering rule, and error outcome. |
| WF-04 | Characterization and preflight | chain | WF-02, WF-03 for Reporting cases | WF-05, WF-06 | Freeze field identity, no-note, Reporting, readshape, AC043, authorization, route, event, and generated behavior before movement. | Existing owner tests plus new characterization and preflight evidence | Focused owner slices and Make-owned drift/static targets | S-01/P-01 pass or record one exact fail-closed blocker per mismatch. |
| WF-05 | Independent structural internalization | chain | WF-04 | WF-06, WF-07 | Delete readshape, internalize codec/revision/fact/merge/portability/recovery mechanics, and preserve narrow public contributions. | Slice-specific Links packages and adapters | Narrowest affected owner targets | Each slice is independently reversible and leaves public behavior unchanged. |
| WF-06 | Links/Reporting adapter migration | chain | WF-03, WF-04, S-02 | WF-07 | Migrate support refs, fields, and facts one at a time through the approved source-fact/adapter/port boundary. | Links facts, application reporting assembly, Reporting providers | Canonical-model/byte parity plus focused owner tests | Old method is removed only after its replacement passes parity. |
| WF-07 | Caller, boundary, and harness accounting completion | chain | WF-05, WF-06 when applicable | WF-08 | Remove superseded exports, enforce import direction, retain test ownership, and preserve AC043 semantic identity. | Direct callers, boundary owner input, authored test-family inputs when needed | Boundary, focused slices, drift/policy/shape checks | Only approved facades remain and generated/harness artifacts are consistent. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Run final narrow-to-broad verification and publish exact continuation evidence. | All authorized slice changes and this tracker | `make agent-finalize`, focused slices, `make test-fast`, risk-based broader checks | Results/run roots, skipped checks, blockers, rollback points, and next action are recorded. |

## 7. Proposed Refactor Slice Plan

All unqualified slices below are behavior-preserving. The remediation request
authorizes owner repairs O-01/O-02 and the evidence-gated B-01, B-02, and B-04
correction branches; conditional branches MUST NOT be folded into a structural
slice or executed before their recorded preflight condition is met. The required
order is O-01, O-02, S-01, P-01, S-02, the
remaining internalization slices, S-06C, caller/accounting cleanup, and final
validation. LRT-REQ-006 applies unchanged throughout.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| O-01 | None | **Completed 2026-08-21:** adopt one owner repair for field-aware identity and no link-local narrative. Amend Core 02 shape/identity/merge/no-note rules, Core 01 field-routing/removal rules, and Core 04 acceptance. | Core 01, Core 02, Core 04; Domain/traceability only as downstream navigation | An incomplete edit could replace one contradiction with another. | Owner-document consistency and requirement-to-acceptance review. | `make lint-markdown`; owner-selected contract checks if authored projections change | Revert the owner revision as one unit before implementation begins. | RB-001/RB-002 owner text is adopted, internally consistent, and matches LRT-REQ-002/003. |
| O-02 | O-01 | **Completed 2026-08-21:** adopt the source-owner/Reporting dependency direction and exact provider responsibilities without changing public route/export schemas. | Core 01 and adopted Reporting NLSpec | DTO ownership, immutable snapshot boundary, ordering, support IDs, and errors | Owner review plus parity-fixture specification. | `make lint-markdown`; generated drift only if authored projections change | Revert the owner amendment before any adapter migration. | LRT-REQ-005 has one normative owner for every interface behavior. |
| S-01 | O-01; O-02 for Reporting cases | **Completed 2026-08-21:** add/freeze behavior-level characterization for field identity, no-note shapes, Store/views, item refs, codecs, mutation, merge, rollback, portability, Reporting output, AC043, routes/events, and browser behavior. | Links and affected owner tests; authored test-family input only when accounting changes | Tests could accidentally encode pre-repair Core contradictions or change harness ownership. | Acceptance matrices in Section 8; preserve all current focused rows. | Focused module Links rows and affected-owner task-guide selections | Revert only new tests/accounting and regenerate; production remains unchanged. | Passing pre-move evidence is retained; current unknown-member and deleted-endpoint behavior are exact P-01 branch inputs. |
| P-01 | S-01 | **Completed 2026-08-21:** run fail-closed preflights for field-aware DDL/data, absence of link-local narrative, revision/bundle field preservation, Reporting endpoint eligibility/parity, readshape consumers, and AC043 current identity. | Migration/schema inspection, migrated service database, codecs, bundle fixtures, Reporting fixtures, tracked-source search, fixture owner inputs | Silent data loss, hidden narrative, unsupported consumer, output drift, claim drift | Preflight fixtures cover every negative branch; no corrective production mutation. | Make-owned owner checks; commands in Section 8 | Preflight is read-only; retain its report and make no corrective edit in this slice. | B-01 condition is false; B-02 and B-04 conditions are true; readshape has zero consumers and AC043 is unchanged. |
| S-02 | P-01 | **Completed 2026-08-21:** delete `internal/modules/links/readshape` with no compatibility shim after the zero-consumer gate. | Deleted `readshape/sql.go`; no caller migration was required | Hidden consumer or SQL-shape drift | Active-view and item-ref contract tests remain. | Links slice; boundary; generation/policy/shape drift; `make test-fast` | Restore the deleted package unchanged if any gate fails. | LRT-AC-READSHAPE-001/002 pass; the no-consumer search remains empty. |
| S-03 | P-01 | **Completed 2026-08-21:** move codec and revision-provider mechanics under `links/internal`, retain the root revision contribution and map-returning Store facade, and remove root methods that exposed codec types. | `internal/valuecodec/`, `internal/revisionprovider/`, `value_codecs.go`, revision contribution, boundary owner input | Stored revision compatibility, field identity, restore/tombstone, descriptors, invalidations | Canonical/legacy codec, contribution, Links integration, and Revisions integration cases pass. | Links unit/integration rows; boundary/drift/shape; `make test-fast` | Revert package/import movement and boundary input as one slice; stored data is never migrated. | Private mechanics have no external importer and canonical/legacy maps and rollback outputs are equivalent. |
| S-04 | P-01 | **Completed 2026-08-21:** replace Timeline DTO coupling with Links-owned facts, Timeline-owned collection inputs, and an application adapter; audit Assessment and history readers and retain them unchanged because they already expose source-owned values. | Links/Timeline fact files, application Timeline fact assembly, projection and Timeline composition callers | Projection rows, ordering, timestamps, nil/empty collections, refresh/invalidation | Links, Timeline, Assessment, Projection, and Saved Views evidence passes. | Links integration row; affected-owner slices; boundary; `make test-fast` | Restore prior DTOs/adapters; no database change is involved. | Links has no Timeline package import, only application composition imports both fact boundaries, and observable projections are equivalent. |
| S-05 | P-01, S-03 | **Completed 2026-08-21:** internalize Links merge SQL/value mechanics, define the Entities-owned link-effects port and command/result values, and supply the Links-backed application adapter explicitly during merge composition. | Private Links merge effects, root thin facade, Entities port/store, application adapter and composition | Full-tuple dedupe, deterministic survivor, counts, mutation order, revision values, invalidations | LRT-AC-FIELD-006/007 and focused Entity merge/protected-set/route evidence pass. | Relevant owner slices; boundary/drift; `make test-fast` | Revert capability adapter and internal move together. | Entities receives equivalent values without importing Links, only application composition imports both boundaries, and Links hides SQL/value mechanics. |
| B-02 | O-01, S-01, P-01 evidence | **Completed 2026-08-21:** extend fixed-row import with an optional owner-populated allowed-column set, validate every row before attribution or SQL, populate exact Links link/tag columns, and map violations to `links_tags.row_shape_exact`. | Incident portability fixed-row validation, Links bundle source shape/descriptor, affected tests and authored test selector | Import compatibility and error mapping | LRT-AC-NOTE-002..006 pass for link rows; valid export/import remains compatible. | Incident Bundle/Links focused tests, generation/drift/policy/shape checks | Revert the correction independently; no data or migration change is permitted. | `note`, `description`, `comment`, and arbitrary undeclared link members reject before attribution or insertion; canonical outputs remain note-free. |
| S-06A | P-01, B-02 when required | **Completed 2026-08-21:** move Incident Bundle export/import/provider mechanics under Links-private implementation, retain only the root `NewIncidentBundleSourcePort`, and route direct owner tests through Links-local test support. | Private Links Incident Bundle package, thin root contribution, test support, boundary owner input | Paths/order, exact shape, field key, validation, attribution, atomicity | LRT-AC-FIELD-008 and NOTE bundle cases plus Incident Bundle catalog/round-trip pass. | Owner task-guide-selected focused slice; boundary/drift/shape; `make test-fast` | Restore prior provider/catalog binding. | Catalog identity and bundle bytes/validation remain stable; production mechanics are private. |
| S-06B | P-01 | **Completed 2026-08-21 as validation-only:** confirm the existing Recovery contribution is already a thin owner declaration and make no structural change. | Audited Links Recovery contribution, application catalog, and frozen Recovery fixtures; no file changed in this slice | Missing/reordered tables and fixture drift | Recovery catalog/fixture tests pass. | Recovery owner task-guide-selected exact catalog rows | No rollback is needed because the implementation was unchanged. | Exactly `record_links` and `record_tags` remain authoritative under `module.links`. |
| B-04 | O-01, S-01, P-01 mismatch | **Completed 2026-08-21:** source legacy Reporting link/tag output from Core-active views, remove view-only type columns to preserve admitted value shapes, and prove old snapshots immutable while new snapshots omit links with deleted endpoints. | Links legacy Reporting provider and Reporting integration acceptance | Observable export/snapshot content | Deleted-source/destination support-ref/field/fact cases, active tags, immutable snapshot, and canonical output pass. | Links/Reporting focused suites, boundary, and `make test-fast` | Revert independently from adapter restructuring. | Core 02-active eligibility is implemented; inactive endpoint rows are omitted only from newly materialized output. |
| S-06C | O-02, P-01, S-02, B-04 when required | **Completed 2026-08-21:** add the typed Links active-fact reader, Reporting-owned support-reference port, and application adapter; prove support/field/fact parity before deleting the legacy provider and wire production composition through both Reporting ports. | Links active facts, application Reporting assembly, Reporting port/materializer/routes, server composition, boundary acceptance; deleted legacy provider | Fact shape, support target fallback, duplicates, nulls, order, errors, immutable boundary | LRT-AC-REPORT-001..010 and Reporting integration/export-model parity pass. | Reporting and Links focused slices; boundary/generation/drift/shape; `make test-fast` | Revert active reader/ports/adapter/composition as one set and restore the legacy provider only with its parity baseline. | Links imports no Reporting DTO, Reporting reads no Links tables/views, only application composition imports both boundaries, and legacy methods/package are removed after parity. |
| S-07 | S-02 through S-06 as applicable | **Completed 2026-08-21:** migrate the remaining Timeline caller from its consumer-specific Links reader to record-scoped Links facts, delete the superseded reader, and audit all remaining Links imports/exports without adding compatibility shims. | Links active/record facts, application Timeline adapter, projection ownership test, direct importers, and boundary owner input | Compile-time breakage can expose unrecorded behavioral coupling | Links and Timeline owner rows preserve projections, ordering, nil/empty posture, and transaction ownership. | Boundary; exact Links rows; Timeline unit/service rows; `make test-fast` | Restore the deleted reader and adapter source selection as one slice; no data or public contract changed. | Searches show no superseded provider, readshape, codec, revision, or Timeline-reader import; only approved mutation, fact, contribution, and testsupport seams remain. |
| S-08 | Any slice that moves tests or changes owner inputs | **Completed 2026-08-21:** retain the four existing Links rows, keep S-01 coverage additions within their current store row, regenerate managed topology through Make, and prove the AC043 fixture/contribution artifacts byte-identical to the baseline. | Authored Links family, generated render index, AC043 owner/contract/key/binding artifacts, and fixture assembly tests | Missing/duplicated execution or AC043 semantic drift | Four Links rows and LRT-AC-AC043-001..004 pass without redistributing the fixture contribution. | `make generate`; drift/policy/shape; fixture key/assembler rows; exact Links rows | Revert the authored selector additions and regenerate only if coverage must be removed; AC043 inputs require no rollback because they did not change. | Accounting is exact and all AC043 semantic artifacts are unchanged. |
| S-09 | S-07 and S-08 when applicable | **Completed 2026-08-21:** perform the terminal dependency/diff audit, rerun high-risk Links/Timeline/Reporting rows, finalize harness structure, run fast/browser/full gates, and close this handoff. | All files changed by authorized slices and this tracker | Late wire, integration, generated, or boundary regression | All characterized evidence remains passing; no stale compatibility surface or unapproved contract change remains. | Focused owner rows; `make agent-finalize`; `make test-fast`; `make browser-e2e-webserver-backed`; `make check` | Revert only the independently identified failing slice; no terminal rollback was needed. | All required checks pass, the one skipped retained-run action is classified, and the handoff contains exact evidence with no remaining blocker. |
| B-01 | O-01, P-01 mismatch | **DROPPED 2026-08-21:** P-01 found no schema/data-projection mismatch, so no forward migration or collision remediation is justified. | No migration or schema file changed | An unnecessary migration would create avoidable lineage and rollback burden. | Field identity and full-tuple merge characterizations remain the evidence. | No migration command is required because the branch condition is false. | Restore branch eligibility only if a separately inspected deployment disproves the recorded baseline. | Existing DDL matches LRT-REQ-002 and migrated databases enforce both null postures plus supersedes uniqueness. |

## 8. Validation Plan

The public commands below were discovered from the live Make task surface.
Executed commands and retained run roots are recorded per checkpoint; unrecorded
commands are not claimed as passing.

### Binary acceptance matrix

Every acceptance row is binary. A later handoff MUST record `pass`, `fail`, or
`not run`; qualitative terms such as "looks correct" do not satisfy a row.

#### Field-aware identity and no-note acceptance

| Acceptance ID | Input or condition | Required outcome |
| --- | --- | --- |
| `LRT-AC-FIELD-001` | Add the same target twice through the same field and link type. | Exactly one active binding remains; retry is idempotent. |
| `LRT-AC-FIELD-002` | Add the same target through two different non-null fields that map to the same link type. | Two active bindings remain, one per exact field key. |
| `LRT-AC-FIELD-003` | Remove field-A `item_ref` while the same target is active in field B. | Field A is tombstoned and field B remains active. |
| `LRT-AC-FIELD-004` | Submit an `item_ref` not active in the patched record and field. | Request fails with `400 invalid_mutation_payload`; no binding changes. |
| `LRT-AC-FIELD-005` | Submit storage `field_key`, raw `link_type`, direction, or table metadata inside an action. | Request fails with `400 invalid_mutation_payload`; the member is not ignored. |
| `LRT-AC-FIELD-006` | Race two same-field adds. | At most one active full-tuple binding commits. |
| `LRT-AC-FIELD-007` | Merge creates same-endpoint collisions in same and different fields. | Same full tuple deduplicates deterministically; different field keys remain distinct. |
| `LRT-AC-FIELD-008` | Export/import a null-field link and a non-null field binding. | Exact `field_key`, including null, round trips without translation. |
| `LRT-AC-FIELD-009` | Tombstone then rollback/restore both binding forms. | Original `record_link_id` and exact field identity are restored. |
| `LRT-AC-FIELD-010` | Insert competing Timeline `supersedes` destinations. | Independent one-active-link-per-superseded-destination invariant remains enforced. |
| `LRT-AC-NOTE-001` | Inspect live DDL, Store DTOs, mutation/revision values, generated contracts, and reporting shapes. | No authoritative link-local narrative member or data exists. |
| `LRT-AC-NOTE-002` | Supply `note`, `description`, or `comment` in a closed relationship action. | Exact owner error is returned; no side effect occurs. |
| `LRT-AC-NOTE-003` | Import a link row containing an unknown narrative member. | Incident Bundle validation fails closed before row insertion or attribution commit. |
| `LRT-AC-NOTE-004` | Encode and restore canonical/legacy revision values. | No link-local narrative is emitted, accepted as restored state, or synthesized. |
| `LRT-AC-NOTE-005` | Build snapshot and Reporting output from links. | No hidden narrative member appears in facts, support metadata, or canonical bytes. |
| `LRT-AC-NOTE-006` | Run generated drift/shape validation. | No generated public contract gains a link-local narrative member. |

#### Reporting, deletion, fixture, and global-freeze acceptance

| Acceptance ID | Input or condition | Required outcome |
| --- | --- | --- |
| `LRT-AC-REPORT-001` | Construct the final dependency graph. | Links imports no Reporting DTO; Reporting imports no Links storage/provider; only application composition imports both contracts. |
| `LRT-AC-REPORT-002` | Reader receives a caller transaction and incident ID. | It starts/commits no transaction, performs no authorization/redaction/selection, and returns complete facts or one typed safe error. |
| `LRT-AC-REPORT-003` | Collect support refs for all included/excluded link types. | Only the three declared types contribute, with exact fallback, order, duplicate, null, and missing-target behavior. |
| `LRT-AC-REPORT-004` | Materialize link facts. | Paths, provider key, source family, content class, value members, explicit nulls, and ordering match canonical pre-change output. |
| `LRT-AC-REPORT-005` | Materialize tag facts. | Paths, source family, value members, explicit nulls, and ordering match canonical pre-change output. |
| `LRT-AC-REPORT-006` | No eligible link/tag rows exist. | Reader returns empty ordered fact families; Reporting output remains valid and deterministic. |
| `LRT-AC-REPORT-007` | A reader query or adapter mapping fails. | No partial provider output is published; caller receives the owner-approved safe error mapping. |
| `LRT-AC-REPORT-008` | Soft-delete a link endpoint while retaining the link row. | P-01 proves whether current and Core-active sets differ; structural migration stops if correction is needed. |
| `LRT-AC-REPORT-009` | Migrate each of the three current methods. | Each replacement passes canonical-model or byte parity before its old method is removed. |
| `LRT-AC-REPORT-010` | Materialize from an admitted immutable snapshot boundary. | Adapter performs no later live source read and canonical export bytes remain deterministic. |
| `LRT-AC-READSHAPE-001` | Run the final tracked-source/generator/build-input search. | No runtime, test, generator, embedded, configuration, or supported-profile consumer exists. |
| `LRT-AC-READSHAPE-002` | Delete the package and run focused/static/drift checks. | All required commands pass and no generated artifact changes unexpectedly. |
| `LRT-AC-READSHAPE-003` | A consumer is discovered. | Deletion stops; the consumer gets an owner-typed migration and no generic shim is added. |
| `LRT-AC-AC043-001` | Compare authored fixture identity before and after the refactor. | Profile, fixture version, seed, contribution, dependency, source contract, and order are exact matches. |
| `LRT-AC-AC043-002` | Validate contribution receipt. | `links=1000`, `mentions=1000`, and `tags=1000` with unchanged receipt identity/shape. |
| `LRT-AC-AC043-003` | Compare generated semantic artifacts. | Digest inputs, snapshot key, verification bindings, and claim bindings are unchanged. |
| `LRT-AC-AC043-004` | Move only a test-support path. | Fixture assembly passes and generated semantic artifacts remain byte-identical. |
| `LRT-AC-FREEZE-001` | Exercise affected HTTP routes and envelopes. | Route, request/response, save-state, conflict, and authorization behavior is unchanged except an independently authorized correction. |
| `LRT-AC-FREEZE-002` | Exercise `/ws/v1/incidents/{incident_id}` after mutations. | Authorization recheck, ordering, event kind, affected fields/views, and replay semantics are unchanged. |
| `LRT-AC-FREEZE-003` | Compare projections, saved views, generated contracts, and frontend selectors. | Observable fields/types/order/capabilities remain unchanged; no direct grid-vendor dependency is introduced. |
| `LRT-AC-FREEZE-004` | Inspect transaction and authorization ownership. | Callers retain transaction lifecycle and route/source facades retain authorization; Links duplicates neither. |

#### Completion acceptance

| Acceptance ID | Required outcome |
| --- | --- |
| `LRT-AC-DONE-001` | Every implemented requirement maps to an adopted owner and passing acceptance evidence. |
| `LRT-AC-DONE-002` | Every slice is independently reviewable and reversible; no behavior correction is hidden in structural movement. |
| `LRT-AC-DONE-003` | Direct import search contains only approved Links facades, contributions, and testsupport. |
| `LRT-AC-DONE-004` | Generated drift, boundary, focused owner, and risk-based broader checks pass or have an exact related failure record. |
| `LRT-AC-DONE-005` | No generated file or historical migration was hand-edited. |
| `LRT-AC-DONE-006` | Final handoff names changed files, commands, run roots, skipped checks, remaining blockers, and the next executable slice. |

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | This tracker and repository Markdown policy | yes | Run for this documentation-only change. |
| unit | `make test-slice OWNER=module.links ROWS=module.links.unit.links_revision_contribution_owns_current_target_08116fb575` | Links revision contribution ownership | yes | Establish before revision-provider movement and repeat afterward. |
| integration | `make service-backed-test-slice OWNER=module.links ROWS=module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1,module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906` | Store, mutation, projection, history, and event behavior | yes | Requires service-backed harness readiness. |
| e2e/browser | `make service-backed-test-slice OWNER=module.links ROWS=module.links.browser.the_notes_tab_supports_browser_visible_creation_ae6bdcc2da` | Browser-visible linked-note workflow | no | Required before and after a route/generated/frontend-facing slice; otherwise run based on risk. |
| broader browser | `make browser-e2e-webserver-backed` | Webserver-backed browser families | no | Use when a public Workbook or browser seam is affected. |
| generated drift | `make generate-drift` | Generated outputs versus authored inputs | no | Required after any authorized authored contract/harness input change. |
| generated policy | `make generated-artifact-policy-check` | Generated-root edit policy | no | Pair with drift checks; generated files are never hand-edited. |
| JSON shape | `make json-shape-check` | Authored/generated JSON contract shapes | no | Required if owner manifests or contracts change. |
| migration drift | `make migration-drift` | Migration history and schema projections | no | Run only for separately authorized B-01/B-02 schema work. |
| import-boundary/static | `make backend-module-boundary-check` | Backend import and SQL ownership boundaries | yes | Establish before facade work and repeat after each caller-migration slice. |
| fast broad suite | `make test-fast` | Repository fast verification | no | Run after focused slices pass. |
| full check | `make check` | Broad repository checks | no | End-of-series or high-risk validation; do not substitute it for focused evidence. |
| finalization | `make agent-finalize` | Repository finalization procedure | no | Run before broader end-of-run verification; if `RESULTS_DIR` is unset, report retained-run maintenance as skipped. |

Discovery completed successfully with:

- `make task-guide ROLE=module-author OWNER=module.links`
- `make explain-test-owner OWNER=module.links`

Those commands identified four active Links rows: one browser, one integration,
one store, and one unit row; three are service-backed. No unit, integration,
browser, generation, migration, boundary, or full validation suite was executed
during planning discovery.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Activate target, authority, scope, and checkpoint protocol | WF-00 | DONE | None | Section 1; 2026-08-21 14:29 EDT execution checkpoint | Target, exclusions, baseline, and between-workstream tracker updates are explicit. |
| T-002 | Inventory all 26 Go files and `.gitkeep` | WF-01 | DONE | T-001 | Section 2 | Every target entry has a disposition and evidence posture. |
| T-003 | Map current contract owners, adopted Reporting authority, callers, and closure decisions | WF-01, WF-02 | DONE | T-002 | Sections 3–5 and requirement index | Every discovered risk has an owner, decision posture, execution status, and acceptance mapping. |
| T-004 | Adopt field-aware identity and no-note owner repair | WF-02 | DONE | RB-001, RB-002 | Core 01 REQ-01-311; Core 02 REQ-02-164/168/170/180; Core 04 AC-207/208/209/332; checkpoint run `.cartulary/test-results/20260821T183337Z-p767953` | Core 01/Core 02/Core 04 repairs are adopted and self-consistent. |
| T-005 | Adopt Links/Reporting dependency and interface boundary | WF-03 | DONE | RB-004, T-004 | Reporting §7.1.2; Core 01 REQ-01-664; Core 04 AC-553; checkpoint run `.cartulary/test-results/20260821T183852Z-p771527` | Core 01 and Reporting owner text adopt LRT-REQ-005. |
| T-006 | Freeze pre-move characterization | WF-04 | DONE | T-004; T-005 for Reporting | New Links field/merge/no-note/bundle tests; expanded Reporting provider integration; updated `module.links` selectors and generated topology; retained runs listed in the 2026-08-21 14:49 EDT checkpoint | Every risky move has passing evidence or one exact blocker. |
| T-007 | Run read-only schema/data/output/consumer/fixture preflights | WF-04 | DONE | T-006 | Repository DDL and migrated test database; S-01 characterizations; exhaustive source searches; boundary pass `.cartulary/test-results/20260821T185139Z-p903301`; P-01 checkpoint log | No migration/no narrative data/no readshape consumer/AC043-current are proved; B-02 and B-04 are opened. |
| T-008 | Delete `readshape` without a shim | WF-05 | DONE | T-007 | Deleted `internal/modules/links/readshape/sql.go`; zero-match search; Links unit `.cartulary/test-results/20260821T185409Z-p924097`; Links service `.cartulary/test-results/20260821T185424Z-p927777`; boundary `.cartulary/test-results/20260821T185410Z-p924230`; fast `.cartulary/test-results/20260821T185504Z-p943753` | LRT-AC-READSHAPE-001/002 pass; LRT-AC-READSHAPE-003 was not triggered. |
| T-009 | Internalize codec and revision-provider mechanics | WF-05 | DONE | T-007 | Private package paths; root map facade; Links unit `.cartulary/test-results/20260821T185757Z-p957473`; Links service `.cartulary/test-results/20260821T185812Z-p962295`; Revisions `.cartulary/test-results/20260821T185902Z-p993967`; fast `.cartulary/test-results/20260821T185948Z-p1009396` | Field-aware canonical/legacy maps and rollback outputs remain equivalent. |
| T-010 | Decouple Timeline/Assessment/history facts | WF-05 | DONE | T-007 | Links/Timeline fact types and application adapter; boundary `.cartulary/test-results/20260821T190601Z-p1025443`; Timeline unit `.cartulary/test-results/20260821T190610Z-p1025852`; Links service `.cartulary/test-results/20260821T190627Z-p1027994`; Timeline service `.cartulary/test-results/20260821T190710Z-p1043838`; Assessments `.cartulary/test-results/20260821T190801Z-p1059089`; fast `.cartulary/test-results/20260821T190859Z-p1100918` | Links owns facts, Timeline owns projection inputs, application adapters preserve ordering and value semantics, and already-conforming Assessment/history readers remain unchanged. |
| T-011 | Internalize merge effects | WF-05 | DONE | T-009 | Private Links merge mechanics; Entities link-effects port; application adapter; boundary `.cartulary/test-results/20260821T191551Z-p1115769`; Links service `.cartulary/test-results/20260821T191717Z-p1133838`; Entities merge `.cartulary/test-results/20260821T191624Z-p1117608`; drift `.cartulary/test-results/20260821T191817Z-p1149226`; fast `.cartulary/test-results/20260821T191828Z-p1152255` | Full-tuple deterministic merge semantics, counts, mutation values/order, and invalidations pass with no Entities-to-Links dependency. |
| T-012 | Enforce exact no-note Incident Bundle input shape if preflight proves a gap | WF-04 | DONE | T-004, T-006, T-007 | Exact fixed-row validation and Links invariant; Links service `.cartulary/test-results/20260821T192238Z-p1169297`; Incident Bundle catalog `.cartulary/test-results/20260821T192349Z-p1186585`; round-trip `.cartulary/test-results/20260821T192359Z-p1187725`; drift/policy/shape roots `.cartulary/test-results/20260821T192441Z-p1202883`, `.cartulary/test-results/20260821T192441Z-p1202893`, `.cartulary/test-results/20260821T192441Z-p1202905` | Unknown members fail before attribution/SQL with no migration/data loss; valid bundle round-trip remains accepted. |
| T-013 | Internalize portability and recovery providers | WF-05 | DONE | T-007, T-012 when required | S-06A evidence: Links `.cartulary/test-results/20260821T192822Z-p1221946`, Incident Bundle catalog `.cartulary/test-results/20260821T192906Z-p1237736`, round-trip `.cartulary/test-results/20260821T192916Z-p1238834`, fast `.cartulary/test-results/20260821T193010Z-p1257761`. S-06B Recovery catalog `.cartulary/test-results/20260821T193209Z-p1267849`. | Portability mechanics are private, contribution identity/bytes/errors are stable, and Recovery still owns exactly the two Links tables. |
| T-014 | Correct Reporting endpoint eligibility if preflight proves drift | WF-04 | DONE | T-004, T-006, T-007 | Active-view provider queries and Reporting snapshot acceptance `.cartulary/test-results/20260821T193654Z-p1312629`; Links `.cartulary/test-results/20260821T193744Z-p1328454`; fast `.cartulary/test-results/20260821T193823Z-p1343648`; boundary `.cartulary/test-results/20260821T193907Z-p1349208` | Deleted source/destination relationships and source-deleted tags are excluded; prior snapshots remain immutable and new snapshots omit inactive rows. |
| T-015 | Migrate Links/Reporting support refs, fields, and facts | WF-06 | DONE | T-005, T-007, T-008, T-014 when required | Pre-removal support/field/fact parity `.cartulary/test-results/20260821T194452Z-p1357523`; post-removal Reporting `.cartulary/test-results/20260821T194646Z-p1378511`; Links `.cartulary/test-results/20260821T194741Z-p1394934`; boundary `.cartulary/test-results/20260821T194639Z-p1378122`; drift/policy/shape roots `.cartulary/test-results/20260821T194839Z-p1413233`, `.cartulary/test-results/20260821T194839Z-p1413247`, `.cartulary/test-results/20260821T194839Z-p1413263`; fast `.cartulary/test-results/20260821T194851Z-p1417028` | All three legacy methods passed exact typed parity before removal; peer DTO/table imports and `links/reportingprovider` are gone. |
| T-016 | Migrate callers and retire obsolete exports | WF-07 | DONE | T-008 through T-015 as applicable | Deleted `timeline_facts.go`; record-scoped source facts and application Timeline adapter; boundary `.cartulary/test-results/20260821T195733Z-p1477029`; Timeline unit `.cartulary/test-results/20260821T195230Z-p1431301`; Timeline service `.cartulary/test-results/20260821T195240Z-p1432312`; Links unit `.cartulary/test-results/20260821T195432Z-p1448441`; Links service `.cartulary/test-results/20260821T195443Z-p1449525`; fast `.cartulary/test-results/20260821T195740Z-p1477486` | Direct import scan contains only approved mutation, source-fact, contribution, and testsupport seams; all obsolete packages and temporary shims are absent. |
| T-017 | Preserve current AC043 topology decision | WF-02 | DONE | None | LRT-REQ-006, RB-005 | No redistribution work is introduced. |
| T-018 | Maintain test/harness accounting and AC043 semantic identity | WF-07 | DONE | T-006, T-016 | Four-row assertion; byte comparison to baseline; fixture key/assembler `.cartulary/test-results/20260821T200054Z-p1495126`; generate `.cartulary/test-results/20260821T200023Z-p1488100`; drift/policy/shape `.cartulary/test-results/20260821T200037Z-p1491121`, `.cartulary/test-results/20260821T200037Z-p1491134`, `.cartulary/test-results/20260821T200037Z-p1491148`; Links unit/service `.cartulary/test-results/20260821T200131Z-p1495894`, `.cartulary/test-results/20260821T200138Z-p1496310` | Authored ownership remains one browser, integration, store, and unit row; frozen AC043 profile, version, seed, contribution order/dependency, receipts, key vector, verification bindings, claim bindings, and test support are byte-identical to baseline. |
| T-019 | Conditional forward migration for field identity | WF-04 | DROPPED | T-004, T-007 mismatch | P-01 no-migration proof; migration 00014 field/non-field/supersedes indexes; passing field/merge characterization | Branch condition is false; no historical or forward migration is changed. |
| T-020 | Complete final validation and handoff | WF-08 | DONE | T-016, T-018 | S-09: Links unit/service `.cartulary/test-results/20260821T200523Z-p1544758`, `.cartulary/test-results/20260821T200530Z-p1545077`; Reporting `.cartulary/test-results/20260821T200404Z-p1514326`; Timeline `.cartulary/test-results/20260821T200442Z-p1529484`; finalize `.cartulary/test-results/20260821T200610Z-p1560271`; fast `.cartulary/test-results/20260821T200625Z-p1563174`; browser `.cartulary/test-results/20260821T200639Z-p1563957`; check `.cartulary/test-results/20260821T201046Z-p1616597`; terminal lint `.cartulary/test-results/20260821T201706Z-p1739054` | Exact evidence, changed files, rollback points, the retained-run skip, and zero remaining blockers are current; no next implementation slice remains. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Scope and authority grounded; documentation-only tracker created | Inspected framework, Domain, Core 00–05, live target, manifests, and callers; touched only this tracker | `git status --short --branch`; `git rev-parse HEAD`; repository searches; framework/source reads | Target exists; label is safely normalized as `links`; clean baseline recorded | RB-001 and RB-002 do not block tracker creation | Validate Markdown and confirm sole-file diff. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Requirements, evidence classes, owner gates, defaults, mappings, acceptance IDs, and completion states added | Inspected this tracker, analysis notes, NLSpec guide, adopted Reporting NLSpec, DDL, Links/Reporting/Incident Bundle sources, owner inputs; touched only this tracker | Repository reads/searches; `make help-all` command discovery; `make lint-markdown` | Markdown passed; run root `.cartulary/test-results/20260821T181213Z-p757028` | RB-001/RB-002/RB-004 remain execution-blocked on adopted owner/evidence gates | Begin later authorized work with O-01, not implementation. |
| 2026-08-21 14:29 EDT | Codex remediation execution | W-00 activated the user-approved remediation while preserving the staged tracker and recorded commit baseline | Touched only this tracker | `git status --short --branch`; `git rev-parse HEAD`; `make lint-markdown` | Active scope and mandatory between-workstream checkpoint protocol recorded | None for O-01 | Adopt O-01 owner repairs, then checkpoint before O-02. |
| 2026-08-21 14:34 EDT | Codex remediation execution | O-01 adopted field-aware active identity and prohibited link-local narrative | Core 01, Core 02, Core 04, and this tracker; Domain navigation audited with no owner-reference change | Targeted owner-clause searches; `make lint-markdown` | Owner clauses and acceptance are self-consistent; checkpoint lint passed at `.cartulary/test-results/20260821T183337Z-p767953` | Runtime/data assertions remain for S-01/P-01; no owner blocker remains | Execute O-02 Reporting boundary owner repair. |
| 2026-08-21 14:38 EDT | Codex remediation execution | O-02 adopted the Links-fact/application-adapter/Reporting-port dependency direction and exact field/support behavior | Reporting NLSpec, Core 01, Core 04, and this tracker | Owner/interface searches; `make lint-markdown` | Owner clauses and AC-553 are self-consistent; checkpoint lint passed at `.cartulary/test-results/20260821T183852Z-p771527` | Characterization and parity remain for S-01/P-01 | Execute S-01 characterization before implementation movement. |
| 2026-08-21 14:49 EDT | Codex remediation execution | S-01 froze field-aware identity, scoped removal, deterministic full-tuple merge, no-note canonical values, exact current bundle admission, Reporting value/order/duplicate/null behavior, and deleted-endpoint current behavior | Links and Reporting tests; `tools/test_families/module.links.json`; generated execution-topology render index | `make format`; `make generate`; drift/policy/shape checks; exact Links unit/store/integration rows; focused Reporting, Timeline, Assessments, Entities, Revisions, Incident Bundles, and Recovery rows | All selected rows passed. Primary runs: Links unit `.cartulary/test-results/20260821T184506Z-p786859`, Links service `.cartulary/test-results/20260821T184515Z-p787282`, Reporting `.cartulary/test-results/20260821T184558Z-p802571`; affected-owner roots are recorded in Tests and harness below | P-01 must classify the characterized unknown-member admission as B-02 and deleted-endpoint inclusion as B-04; production remains unchanged | Execute read-only P-01 and open/drop only the conditional branches proved by evidence. |
| 2026-08-21 14:52 EDT | Codex remediation execution | P-01 proved the repository field-aware DDL/index baseline and migrated-database behavior, no production narrative shape, exact revision/bundle field carriage, zero readshape consumers, and unchanged AC043 identity; it confirmed both conditional behavior gaps | Migration 00014, Links Store/merge/codec/portability, generated and authored shapes, Reporting characterization, readshape symbols/imports/configuration, AC043 owner/key/contract artifacts | Exhaustive `rg`; focused migrated-database rows; `make backend-module-boundary-check`; `make generate-drift` | B-01 condition false. B-02 condition true because unknown link members reach fixed-row insert. B-04 condition true because raw Reporting queries retain links after endpoint deletion. Boundary initially failed only for a test-only owner-port import at `.cartulary/test-results/20260821T185027Z-p895518`; the test was relocated behind the Links facade and the gate passed at `.cartulary/test-results/20260821T185139Z-p903301`. Updated store characterization passed at `.cartulary/test-results/20260821T185139Z-p903174`; drift passed at `.cartulary/test-results/20260821T185139Z-p903118`. | No external deployment database was in scope; repository DDL plus freshly migrated service databases are the available baseline. B-02 and B-04 are required. | Execute B-01 as DROPPED with no-migration proof, then checkpoint before S-02. |
| 2026-08-21 14:54 EDT | Codex remediation execution | B-01 was conditionally DROPPED with no migration | Migration 00014, P-01 evidence, tracker | No product command; P-01 checkpoint lint passed at `.cartulary/test-results/20260821T185305Z-p921788` | Existing field/non-field active uniqueness and independent supersedes uniqueness match the adopted owner; no collision mapping is needed | A separately scoped deployed database with divergent DDL/data would require a new reviewed preflight, not automatic remediation | Execute S-02 readshape deletion. |
| 2026-08-21 14:56 EDT | Codex remediation execution | S-02 deleted the dead `links/readshape` package without a forwarding shim | Deleted `internal/modules/links/readshape/sql.go`; no caller or policy input changed | Final symbol/path/config search; exact Links rows; boundary, drift, generated policy, JSON shape, and `make test-fast` | Zero consumer matches. Boundary `.cartulary/test-results/20260821T185410Z-p924230`, drift `.cartulary/test-results/20260821T185409Z-p924040`, policy `.cartulary/test-results/20260821T185424Z-p927698`, shape `.cartulary/test-results/20260821T185424Z-p927725`, Links service `.cartulary/test-results/20260821T185424Z-p927777`, fast `.cartulary/test-results/20260821T185504Z-p943753` all passed | None | Execute S-03 codec and revision-mechanics internalization. |
| 2026-08-21 15:00 EDT | Codex remediation execution | S-03 moved `valuecodec` and `revisionprovider` under Links-private packages, preserved the root revision contribution and map facade, and removed the typed codec methods from the root Store surface | Moved codec/provider files and tests; updated Links imports and boundary owner input | `make format`; `make generate`; exact Links/Revisions rows; boundary, drift, policy, shape, and `make test-fast` | Format `.cartulary/test-results/20260821T185734Z-p950707`, generate `.cartulary/test-results/20260821T185741Z-p954341`, boundary `.cartulary/test-results/20260821T185757Z-p957656`, drift `.cartulary/test-results/20260821T185757Z-p957400`, policy `.cartulary/test-results/20260821T185757Z-p957418`, shape `.cartulary/test-results/20260821T185812Z-p962244`, Links/Revisions/fast roots in T-009 all passed | Concurrent first Revisions attempt failed in harness image warm-stamp coordination at `.cartulary/test-results/20260821T185812Z-p962306`; sequential rerun passed and the failure is unrelated to the slice | Execute S-04 Links-owned Timeline fact decoupling. |
| 2026-08-21 15:09 EDT | Codex remediation execution | S-04 replaced Links-to-Timeline DTO ownership with Links-owned source facts, Timeline-owned collection inputs, and a deep-copying application adapter | Added `internal/app/timelinefactassembly`; updated Links facts, Timeline collection/projection constructors, application projection/Timeline assembly, and projection test composition; Assessment/history readers audited and retained | `make format`; task guides; exact Links rows; Timeline collection/projection unit rows; Timeline projection service row; Assessment slice; boundary; `make test-fast` | Format `.cartulary/test-results/20260821T190530Z-p1020278`, boundary `.cartulary/test-results/20260821T190601Z-p1025443`, Timeline unit `.cartulary/test-results/20260821T190610Z-p1025852`, Links unit `.cartulary/test-results/20260821T190618Z-p1026928`, Links service `.cartulary/test-results/20260821T190627Z-p1027994`, Timeline service `.cartulary/test-results/20260821T190710Z-p1043838`, Assessment `.cartulary/test-results/20260821T190801Z-p1059089`, and fast `.cartulary/test-results/20260821T190859Z-p1100918` passed | No blocker; rollback is the adapter/type/constructor change set only, with no database or public-contract change | Execute S-05 Entities-owned merge-effects port and Links-backed application adapter. |
| 2026-08-21 15:19 EDT | Codex remediation execution | S-05 moved merge SQL/value mechanics under Links-private implementation, replaced the Entities-owned Links adapter/DTO dependency with an Entities-owned link-effects port, and injected a Links-backed application adapter | Added private merge mechanics and `internal/app/entitymergeassembly`; retained a thin Links facade; updated Entities ports/store/composition and focused test composition | `make format`; task guide; boundary; exact Links rows; focused Entities protected-set/store/route rows; `make generate-drift`; `make test-fast` | Format `.cartulary/test-results/20260821T191544Z-p1111980`, boundary `.cartulary/test-results/20260821T191551Z-p1115769`, Links unit `.cartulary/test-results/20260821T191559Z-p1116157`, Entities `.cartulary/test-results/20260821T191624Z-p1117608`, Links service `.cartulary/test-results/20260821T191717Z-p1133838`, drift `.cartulary/test-results/20260821T191817Z-p1149226`, and fast `.cartulary/test-results/20260821T191828Z-p1152255` passed | No blocker; rollback is the private-mechanics/root-facade/Entities-port/application-adapter change set with no schema or wire change | Execute B-02 exact Incident Bundle row-shape enforcement. |
| 2026-08-21 15:25 EDT | Codex remediation execution | B-02 closed the characterized bundle-smuggling gap with owner-populated exact row shapes validated before all attribution and SQL work | Incident portability fixed-row contract; Links link/tag allowed columns and descriptor invariant; negative Links test; authored Links selector and regenerated topology | `make format`; `make generate`; exact Links service rows; Incident Bundle source-catalog and round-trip rows; drift/policy/shape; boundary | Format `.cartulary/test-results/20260821T192217Z-p1162673`, generate `.cartulary/test-results/20260821T192223Z-p1166332`, Links `.cartulary/test-results/20260821T192238Z-p1169297`, catalog `.cartulary/test-results/20260821T192349Z-p1186585`, round-trip `.cartulary/test-results/20260821T192359Z-p1187725`, drift/policy/shape roots in T-012, and boundary `.cartulary/test-results/20260821T192453Z-p1206627` passed | Invalid version-2 rows with undeclared members are intentionally no longer accepted; no valid bundle, schema, data, or public wire shape changed | Execute S-06A portability internalization. |
| 2026-08-21 15:31 EDT | Codex remediation execution | S-06A moved portability/provider mechanics private and left only the root source-port contribution; direct owner tests use Links-local test support rather than a production forwarder | Moved Links portability/source-port implementation; added thin root contribution and test-support helper; updated boundary owner paths | `make format`; `make generate`; boundary; exact Links rows; Incident Bundle catalog/round-trip; drift/policy/shape; `make test-fast` | Format `.cartulary/test-results/20260821T192806Z-p1217791`, generate `.cartulary/test-results/20260821T192715Z-p1212846`, boundary `.cartulary/test-results/20260821T192730Z-p1215810`, Links unit `.cartulary/test-results/20260821T192814Z-p1221482`, Links service `.cartulary/test-results/20260821T192822Z-p1221946`, catalog/round-trip roots in T-013, drift `.cartulary/test-results/20260821T192958Z-p1253980`, policy `.cartulary/test-results/20260821T192958Z-p1254002`, shape `.cartulary/test-results/20260821T192958Z-p1253997`, and fast `.cartulary/test-results/20260821T193010Z-p1257761` passed | First Links unit attempt failed at `.cartulary/test-results/20260821T192740Z-p1216233` because the moved test helper import was omitted; adding the owner-local import fixed the slice. Rollback is the portability move/contribution/test-support/boundary change set. | Execute S-06B Recovery validation-only closure. |
| 2026-08-21 15:32 EDT | Codex remediation execution | S-06B confirmed that Recovery is already the intended thin contribution and deliberately made no implementation change | Audited `internal/modules/links/recovery_state.go`, recovery assembly catalog tests, and current/pre-ownership/graph-v2 fixtures | Recovery task guide; exact current/frozen catalog unit rows; source/fixture searches | `.cartulary/test-results/20260821T193209Z-p1267849` passed; all inspected catalogs assign exactly `record_links` and `record_tags` to `module.links` | No blocker and no rollback point because this was validation-only | Execute B-04 Reporting active-endpoint correction. |
| 2026-08-21 15:39 EDT | Codex remediation execution | B-04 aligned the legacy Reporting provider with active endpoint/record semantics and added immutable-old/new-snapshot acceptance | Links Reporting provider and Reporting evidence integration test | `make format`; Reporting task guide and focused service row; exact Links rows; boundary; `make test-fast` | Format `.cartulary/test-results/20260821T193645Z-p1308935`, Reporting `.cartulary/test-results/20260821T193654Z-p1312629`, Links unit `.cartulary/test-results/20260821T193735Z-p1327965`, Links service `.cartulary/test-results/20260821T193744Z-p1328454`, fast `.cartulary/test-results/20260821T193823Z-p1343648`, and boundary `.cartulary/test-results/20260821T193907Z-p1349208` passed | First Reporting attempt failed at `.cartulary/test-results/20260821T193427Z-p1273879` because a test-only SQL `CASE` inferred text for a timestamp; two typed updates fixed the fixture. New snapshot output intentionally omits inactive relationships; stored snapshots are unchanged. | Execute S-06C Links active-fact/Reporting port adapter migration. |
| 2026-08-21 15:49 EDT | Codex remediation execution | S-06C established the Links-fact/application-adapter/Reporting-port seam, proved all three legacy method outputs equal, migrated production composition, and deleted `links/reportingprovider` | Added Links active facts and application Reporting adapter; added Reporting support port and injection; updated server/test composition and boundary guard/policy; deleted legacy provider | `make format`; pre-removal parity Reporting row; post-removal Reporting and exact Links rows; boundary; `make generate`; drift/policy/shape; `make test-fast`; final dependency/table searches | Format `.cartulary/test-results/20260821T194631Z-p1374388`, parity `.cartulary/test-results/20260821T194452Z-p1357523`, post-removal Reporting `.cartulary/test-results/20260821T194646Z-p1378511`, Links unit `.cartulary/test-results/20260821T194730Z-p1393766`, Links service `.cartulary/test-results/20260821T194741Z-p1394934`, boundary `.cartulary/test-results/20260821T194639Z-p1378122`, generate `.cartulary/test-results/20260821T194822Z-p1410234`, drift/policy/shape and fast roots in T-015 all passed | Initial boundary run `.cartulary/test-results/20260821T194421Z-p1356728` correctly rejected the new Evidence provider import until the authorized application adapter path was added to the owner policy. Rollback is the active-reader/ports/adapter/composition change set plus legacy provider restoration. | Execute S-07 caller and facade cleanup. |
| 2026-08-21 15:58 EDT | Codex remediation execution | S-07 consolidated Timeline onto Links-owned record facts, removed the final consumer-specific Links reader, and completed the facade/private-package/import audit without retaining a shim | Added record-scoped fact reading to `active_facts.go`; updated the application Timeline adapter and projection ownership test; deleted `timeline_facts.go`; audited production imports and private-package reachability | `make format`; exhaustive symbol/path/import searches; boundary; Timeline unit/service rows; exact Links rows; `make test-fast` | Format `.cartulary/test-results/20260821T195655Z-p1472667`, boundary `.cartulary/test-results/20260821T195733Z-p1477029`, Timeline unit `.cartulary/test-results/20260821T195230Z-p1431301`, Timeline service `.cartulary/test-results/20260821T195240Z-p1432312`, Links unit `.cartulary/test-results/20260821T195432Z-p1448441`, Links service `.cartulary/test-results/20260821T195443Z-p1449525`, and fast `.cartulary/test-results/20260821T195740Z-p1477486` passed | First post-cleanup boundary run `.cartulary/test-results/20260821T195704Z-p1476405` exposed three unrelated record-envelope allowlist entries accidentally dropped while normalizing moved paths; restoring them fixed the policy-only regression. Rollback is the record-fact/Timeline-adapter/deleted-reader set. | Execute S-08 harness and AC043 accounting. |
| 2026-08-21 16:02 EDT | Codex remediation execution | S-08 confirmed exact four-row Links ownership and preserved AC043 semantic identity without fixture redistribution or authored ownership changes beyond the existing S-01 selector additions | Audited Links owner/family manifests, generated topology, AC043 contract/snapshot owner/builder/key vectors, Timeline verification and test-family bindings, browser batch bindings, and Links fixture test support | `make task-guide`; `make generate`; drift/policy/shape; fixture snapshot-key and source-owner-assembler rows; exact Links rows; baseline byte comparison and exact JSON assertions | Generate `.cartulary/test-results/20260821T200023Z-p1488100`, drift/policy/shape roots in T-018, fixture rows `.cartulary/test-results/20260821T200054Z-p1495126`, Links unit `.cartulary/test-results/20260821T200131Z-p1495894`, and Links service `.cartulary/test-results/20260821T200138Z-p1496310` passed; all named AC043 semantic files are byte-identical to `22d33ee5` | No blocker. No AC043 rollback is needed because its inputs and generated semantic artifacts did not change; the generated topology index remains downstream of the intentionally expanded Links store selector. | Execute S-09 final validation and handoff completion. |
| 2026-08-21 16:15 EDT | Codex remediation execution | S-09 completed final narrow-to-broad validation and closed the remediation with no remaining implementation blocker or compatibility shim | Final source/import/diff audit plus every file in the changed-file inventory below; `docs/domain.md`, historical migrations, public route schemas, generated public contracts, frontend code, and AC043 artifacts remain unchanged | `git diff --check`; exhaustive stale-path/private-import/peer-table searches; task guides; focused Links/Timeline/Reporting rows; `make agent-finalize`; `make test-fast`; `make browser-e2e-webserver-backed`; `make check` | All focused and broad gates passed at the roots in T-020. `make check` passed 637/637 units and the browser gate passed 58/58 units. | `agent-finalize` passed but retained-run validation was skipped because `RESULTS_DIR` was unset and no prior successful full warm check existed at invocation time; its summary records `results_dir_status=skipped`. No product check was skipped and no blocker remains. | No implementation slice remains. Preserve the independent rollback points above and hand the completed worktree to review/commit. |

### Final changed-file inventory

- Specifications and handoff: `docs/reporting-subsystem-nlspec.md`,
  `docs/spec/01_architecture_storage_and_view_contracts.md`,
  `docs/spec/02_domain_model_schema_and_history.md`,
  `docs/spec/04_security_deployment_and_conformance.md`, and this tracker.
- Application adapters and composition: `internal/app/entitymergeassembly/links.go`,
  `internal/app/reportingassembly/links.go`,
  `internal/app/timelinefactassembly/links.go`,
  `internal/app/projectionassembly/build.go`,
  `internal/app/projectionassembly/source_ownership_test.go`,
  `internal/app/server/runtime_assembly.go`, and
  `internal/app/timelineassembly/assembly.go`.
- Entities and shared portability: `internal/modules/entities/merge/link_effects.go`,
  `internal/modules/entities/merge/merge_store.go`,
  `internal/modules/entities/merge/ports.go`,
  `internal/modules/entities/merge/store.go`, both
  `merge_protected_set*_test.go` files, and
  `internal/modules/incidentportability/portability.go`.
- Links owner surface and tests: `internal/modules/links/active_facts.go`,
  `collection_mutations.go`, `incident_bundle_source_port.go`,
  `links_tags_test.go`, `merge_effects.go`,
  `revision_provider_contribution.go`, `testsupport/links.go`, and
  `value_codecs.go`.
- Links private implementations: `internal/modules/links/internal/incidentbundle/`
  `portability.go` and `source_port.go`;
  `internal/modules/links/internal/mergeeffects/merge_effects.go`;
  `internal/modules/links/internal/revisionprovider/provider.go`; and
  `internal/modules/links/internal/valuecodec/valuecodec.go` plus its test.
- Removed obsolete Links surfaces: `incident_bundle_portability.go`,
  `readshape/sql.go`, `reportingprovider/provider.go`,
  `revisionprovider/provider.go`, `timeline_facts.go`, and the old
  `valuecodec/valuecodec.go` plus its test.
- Reporting, Timeline, and Projection consumers/tests:
  `internal/modules/reporting/boundary_guard_test.go`,
  `evidence_provider_integration_test.go`,
  `evidence_provider_validation_test.go`, `export_materializer.go`,
  `exportprovider/provider.go`, and `routes.go`;
  `internal/modules/timeline/collectionfacts/reader.go` plus its test and
  `projectionprovider/source.go`; and
  `internal/modules/projections/internal/runtime/rebuild_test.go` plus
  `internal/modules/projections/testsupport/build.go`.
- Authored/generated harness surfaces: `tools/backend_module_boundaries.json`,
  `tools/test_families/module.links.json`, and Make-generated
  `tools/execution_topology_render_index.json`.

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Links is a legitimate but mixed source-owner module; Timeline DTO coupling is the first must-fix boundary | All Links Go files; application assemblies; peer module callers; boundary manifest; touched only this tracker | `rg`/`find`/`sed` inspection; direct-import and exported-symbol searches | Public/private seam and future slice order documented | Pre-move characterization remains TODO; RB-001 blocks identity changes | In a later authorized task, execute S-01 before any package move. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Closure direction is exact: field-aware identity, no link narrative, no readshape shim, source-fact adapters, unchanged AC043 | Links Store/field/merge/codec/portability/readshape sources and module callers; touched only this tracker | Exact source reads and path/symbol/import searches | Existing DDL/merge/revision/portability support expected field identity; structural and correction slices are separated | Owner adoption precedes S-01; B-01/B-02/B-04 are fail-closed conditional/correction branches | Adopt O-01/O-02, then characterize and preflight. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Frontend is a generated-contract/generic collection consumer, not a Links implementation owner | Generated protocol/view-contract consumers and grid-adapter import policy; touched only this tracker | Repository searches for Links fields, selectors, and direct grid-vendor imports | No direct Links frontend or out-of-adapter grid-vendor coupling found | None for backend structural work | Keep frontend out of scope unless an authorized public contract change supplies new evidence. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Frontend remains a frozen consumer surface; no new UI, selector, or vendor seam is introduced | Current frontend/generated findings and acceptance mapping; touched only this tracker | Tracker/source comparison and Markdown validation | LRT-AC-FREEZE-003 now states the binary no-drift outcome | None before owner repairs; any future public-contract proposal reopens frontend review | Keep frontend out of structural slices and run browser evidence only when affected. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Storage, route, event, projection, revision, bundle, reporting, recovery, generated, and harness surfaces mapped | Core owners, migration, view schemas, generated consumers, source/revision/recovery registries; touched only this tracker | Contract/manifests searches and exact source reads | No Links-owned HTTP route; generated files are consumers/projections only | RB-001 owner contradiction; RB-002 spec/repo drift | Preserve all contract shapes; seek owner resolution only in a later behavior-change task. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Proposed owner repairs and Reporting boundary are specified without being misrepresented as adopted | Core 01–05, adopted Reporting NLSpec, migration, generated/fixture owner inputs; touched only this tracker | Owner-clause searches, exact DDL/source reads, Markdown validation | Owner/support allocation, closure decisions, exact mappings, and adoption consequences are complete | O-01/O-02 not yet adopted; Incident Bundle unknown-member behavior needs owner-approved characterization | Authorize and adopt owner documents before code or generated projections change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Existing Links tests and four catalog rows mapped; future focused commands recorded | Links tests, cross-owner tests, module owner/family files, fixture policies; touched only this tracker | `make task-guide ROLE=module-author OWNER=module.links`; `make explain-test-owner OWNER=module.links` | Both discovery targets passed; no actual test suite was run | Missing-characterization assessment remains TODO; AC043 topology changes deferred | Run S-01 under later authorization, then the narrowest listed owner rows. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Binary acceptance matrices cover field/no-note, Reporting, readshape, AC043, global freeze, and final completion | Existing test/harness evidence, AC043 owner/contract, acceptance IDs; touched only this tracker | Manifest/source inspection; `make lint-markdown` only | Documentation lint passed; no unit, integration, browser, migration, generation, boundary, or product suite ran | S-01 and P-01 remain TODO; AC043 redistribution is closed as no action | Add/retain evidence under later authorization without changing fixture identity or inferring runtime architecture. |
| 2026-08-21 14:49 EDT | Codex remediation execution | S-01 added behavior-level characterization without changing production or the four-row Links owner topology | Links/Reporting tests, authored Links family manifest, generated render index | `make format` root `.cartulary/test-results/20260821T184335Z-p775007`; `make generate` `.cartulary/test-results/20260821T184432Z-p780005`; generate drift `.cartulary/test-results/20260821T184451Z-p783173`; policy `.cartulary/test-results/20260821T184451Z-p783181`; shape `.cartulary/test-results/20260821T184451Z-p783193`; focused runs | Links unit/store/integration, Reporting, Timeline unit/store, Assessments, Entities, Revisions, Incident Bundles, and Recovery passed. Affected-owner service roots: `.cartulary/test-results/20260821T184708Z-p819031`, `.cartulary/test-results/20260821T184708Z-p819036`, `.cartulary/test-results/20260821T184708Z-p819044`, `.cartulary/test-results/20260821T184750Z-p864000`, `.cartulary/test-results/20260821T184750Z-p864006`; unit roots `.cartulary/test-results/20260821T184657Z-p817904`, `.cartulary/test-results/20260821T184657Z-p817908`, `.cartulary/test-results/20260821T184657Z-p817917` | Current bundle unknown-member admission and deleted-endpoint Reporting inclusion are characterized gaps, not accepted owner behavior | Run P-01 read-only baseline and route those findings only to B-02/B-04. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Authorization remains upstream in route/source facades; Links owns authorized source operations | Core 04 and relevant application/route composition; touched only this tracker | Source/import searches and owner reads | No reason found to add transport authorization to Links | None for behavior-preserving package moves | Preserve incident membership/role outcomes and re-run owner route tests after any adapter change. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Upstream authorization and caller-owned transaction lifecycle are explicit LRT-REQ-001 invariants | Core 04, application composition, Links and Reporting interfaces; touched only this tracker | Source/interface reads and Markdown validation | LRT-AC-FREEZE-001/002/004 make unchanged security/event/transaction outcomes binary | No security design blocker; owner repairs must not relocate authorization | Re-run owning route/event authorization evidence after affected adapters move. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21 12:15 EDT | Codex planning/tracker session | Tracker is decision-complete for later characterization and structural slices | This tracker; no production/test/contract/generated/migration/harness file touched | Inspection/discovery commands only before tracker validation | Slices are independently reversible and behavior-preserving; later-authorized changes are separated | RB-001, RB-002; Reporting seam and AC043 topology deliberately deferred | Start future work with S-01, not with package movement or schema edits. |
| 2026-08-21 14:04 EDT | Codex NLSpec-style tracker revision | Design choices are closed; execution order begins with owner repair and ends with retained evidence | This tracker only touched; all source documents remain unchanged | Repository reads/searches and `make lint-markdown` | Tracker now distinguishes decision closure, execution blockers, conditional corrections, and future Definition of Done | RB-001/RB-002/RB-004 blocked; RB-003 evidence-gated; RB-005 done | Execute O-01, O-02, S-01, P-01, then only the unblocked slice sequence. |

## 11. Open Questions and Blockers

The design questions are closed below. `BLOCKED` means the selected direction is
not executable until its owner/evidence gate passes; it does not mean the design
is left to the implementer.

| ID | Normative closure decision | Required closure evidence | Decision posture | Execution status |
| --- | --- | --- | --- | --- |
| RB-001 | Active identity includes exact `field_key` for field-routed links; `null` is one identity value for explicit routes without a field; different non-null fields are distinct; superseded-destination uniqueness remains independent. | Adopted Core 01/Core 02 repair; Core 04 cases; LRT-AC-FIELD-001..010; P-01 no-migration proof. | ADOPTED — OWNER GATE PASSED | NO-MIGRATION BASELINE PROVED |
| RB-002 | The current profile has no link-local narrative. Core 02 removes `note`; exact shapes omit and reject `note`, `description`, `comment`, and equivalents. | Adopted Core 02 correction; no-data/no-shape preflight; LRT-AC-NOTE-001..006; required B-02 exact import correction. | ADOPTED — OWNER GATE PASSED | DONE |
| RB-003 | Delete `internal/modules/links/readshape` without a compatibility shim. A discovered consumer receives an owner-specific migration, not package retention. | Final tracked-source/generator/configuration gate plus LRT-AC-READSHAPE-001..003 and required Make checks. | CLOSED — PACKAGE DELETED | DONE |
| RB-004 | Links owns immutable facts; Reporting owns its port, DTOs, support identity, ordering, missing-reference behavior, and canonical output; application composition maps between them. | Adopted Reporting §7.1.2 and Core 01 REQ-01-664; current/Core-active eligibility preflight; LRT-AC-REPORT-001..010; method-by-method parity; no peer DTO/table imports. | ADOPTED — OWNER GATE PASSED | DONE |
| RB-005 | No redistribution is permitted for the current AC043 fixture. Preserve the exact LRT-REQ-006 identity, counts, order, receipts, digests, keys, and bindings. | LRT-AC-AC043-001..004 and unchanged generated/fixture semantic artifacts. | RESOLVED — NO CHANGE | DONE |

No execution blocker remains. B-01 was dropped on repository and freshly migrated
database evidence; a separately scoped deployed database with divergent DDL or
data would require a new reviewed preflight. Authorization stays upstream, AC043
stays unchanged, and historical migrations stay append-only.

## 12. Binary Completion Criteria

### Tracker-revision completion

This NLSpec-style tracker revision is complete only when all of the following are
true:

- [x] Every Go file and `.gitkeep` under `internal/modules/links` is inventoried or
  explicitly out of scope with a reason.
- [x] Every discovered public contract risk has a named current owner, evidence,
  existing-test posture, characterization posture, and risk rating.
- [x] Every selected closure has exact normative behavior, defaults, boundaries,
  owner-adoption consequences, and an execution status.
- [x] Field identity, no-note shape, Reporting facts/mappings, readshape deletion,
  and AC043 freeze have closed interface or mapping tables.
- [x] Every `LRT-REQ-*` requirement maps to binary `LRT-AC-*` criteria, workflows,
  slices, and Make-owned evidence.
- [x] Every workflow has dependencies, validation, and a handoff checkpoint.
- [x] Every proposed implementation slice is behavior-preserving unless marked
  `requires later authorization`.
- [x] Validation commands are repository-discovered Make targets; commands not
  executed are not represented as passing.
- [x] Owner contradictions were repaired in the named adopted owners before their
  implementation slices began.
- [x] Repository/owner mismatches, including optional Link narrative, unknown
  Incident Bundle members, and Reporting endpoint eligibility, have fail-closed
  branches rather than guessed outcomes.
- [x] Phase maps, test rows, and fixture contributions are treated as evidence
  accounting rather than runtime architecture.
- [x] Generated artifacts are identified as downstream surfaces and are never
  proposed for hand editing.
- [x] The handoff is current enough for review or commit without rediscovering
  target scope, owner posture, callers, risks, commands, or rollback points.
- [x] The authorized remediation is implemented and final evidence is retained.

### Iteration 1 Definition of Done

The refactor is complete because retained implementation evidence now satisfies
every item below.

- [x] O-01 and O-02 are adopted by their normative owners.
- [x] S-01 characterization and P-01 preflights pass; every mismatch is routed to
  its independent correction branch.
- [x] LRT-AC-FIELD-001..010 and LRT-AC-NOTE-001..006 pass.
- [x] `links/readshape` is deleted and LRT-AC-READSHAPE-001..003 pass.
- [x] Codec, revision, projection-fact, merge, portability, and recovery mechanics
  are private behind approved narrow facades with equivalent behavior.
- [x] LRT-AC-REPORT-001..010 pass and `links/reportingprovider` is removed.
- [x] LRT-AC-AC043-001..004 prove the current fixture contract is unchanged.
- [x] LRT-AC-FREEZE-001..004 prove routes, events, projections, authorization,
  transactions, generated contracts, frontend selectors, and grid boundaries did
  not drift.
- [x] LRT-AC-DONE-001..006 pass and the final handoff contains exact retained
  evidence.

### Non-normative implementation references

The following primary references support mechanics only. They do not own
Cartulary behavior and do not override Make-owned repository commands:

- [PostgreSQL 16 constraints](https://www.postgresql.org/docs/16/ddl-constraints.html)
  documents unique constraints, null treatment, and partial-index availability
  relevant only if conditional B-01 is authorized.
- [Go command documentation](https://pkg.go.dev/cmd/go) describes dependency and
  test-package enumeration that MAY support RB-003 investigation; direct Go
  commands are not canonical Cartulary validation.

## 13. Iteration 2: Production Readiness and Legacy Removal

This section was the controlling plan for the completed second Links iteration. It superseded
all Iteration 1 instructions to retain compact link mutation values, the legacy
`tag_id` alias, restore-time actor or provenance defaults, bare tag target IDs,
obsolete facade methods, or compatibility shims. Iteration 1 completion evidence
remains valid evidence of the state from which this iteration began; it did not
grant compatibility authority beyond the completed cutover.

### 13.1 Current Scope and Baseline

| Item | Current posture |
| --- | --- |
| Baseline commit | `21c226ce353961f47e438afaf350760bc1e82c6f` on `main`, five commits ahead of `origin/main` |
| Baseline worktree | Clean at planning inspection |
| Live target inventory | 26 Go files plus one tracked `.gitkeep` under `internal/modules/links` |
| Iteration status | Planned; implementation not started |
| Product posture | Pre-production; every database and retained bundle is disposable |
| Compatibility posture | Hard cutover with reset/regeneration only; no migration, backfill, alias, inference, dual read, dual write, forwarding shim, or historical-value rewrite |
| Authorized document-step change | This tracker only |
| Future implementation scope | Links, directly affected owner specifications and acceptance, application composition, Entity-mentions and Entity merge composition, direct Links mutation callers, tests, and boundary policy |
| Public behavior freeze | Preserve current HTTP and WebSocket shapes, authorization outcomes, caller-owned transactions, active source-row bundle shape, Recovery tables, frontend behavior, and AC043 identity |
| Default database action | Reset after the hard-cutover implementation and before service-backed completion evidence |

Planning inspection retained the following passing baselines:

| Evidence | Result | Run root |
| --- | --- | --- |
| Markdown lint | Pass | `.cartulary/test-results/20260821T211047Z-p1761480` |
| Backend module boundary | Pass | `.cartulary/test-results/20260821T211105Z-p1762443` |
| Links non-service slice | Pass, 14 of 14 units | `.cartulary/test-results/20260821T211111Z-p1762829` |

These runs establish the planning baseline only. They do not prove any Iteration 2
requirement implemented.

### 13.2 Live Inventory

| Current paths | Count | Current responsibility | Iteration 2 disposition |
| --- | ---: | --- | --- |
| `.gitkeep` | 1 | Obsolete placeholder in a non-empty package | Delete in PR-05 |
| `active_facts.go`, `assessment_facts.go`, `timeline_history_facts.go` | 3 | Incident/record active facts plus consumer-specific Assessment and Timeline reads | Consolidate behind one owner-neutral `FactReader` in PR-06 |
| `commands.go`, `store.go`, `field_refs.go`, `value_codecs.go` | 4 | Link commands, persistence DTOs, low-level reference/tag operations, and mutation-value loads | Replace broad results and dead helpers with canonical mutation-result commands in PR-02 through PR-05 |
| `collection_actions.go`, `collection_mutations.go` | 2 | Collection validation plus mutation-aware and legacy non-mutation application paths | Retain mutation-aware paths; delete non-mutation duplicates in PR-05 |
| `merge_effects.go`, `internal/mergeeffects/merge_effects.go` | 2 | Thin root merge facade and private Links-owned merge mechanics | Retain the facade; canonicalize tag targets and values in PR-04 |
| `internal/valuecodec/valuecodec.go`, `internal/valuecodec/valuecodec_test.go` | 2 | Canonical encoders plus compact, alias, and defaulting compatibility readers | Make exact and fail closed; delete every legacy branch and fixture in PR-04 |
| `revision_provider_contribution.go`, `revision_provider_contribution_test.go`, `internal/revisionprovider/provider.go` | 3 | Thin contribution and private history/rollback mechanics | Retain contribution; consume only canonical values and targets after PR-04 |
| `incident_bundle_source_port.go`, `internal/incidentbundle/portability.go`, `internal/incidentbundle/source_port.go` | 3 | Thin contribution and private Links source-row portability | Retain current source-row shape; reject old full bundles through canonical Revisions validation after PR-04 |
| `item_refs.go`, `item_refs_test.go` | 2 | Canonical collection item references | Retain exact current item-reference grammar |
| `recovery_state.go` | 1 | Thin authoritative-table contribution | Retain exactly `record_links` and `record_tags` |
| `links_tags_test.go` | 1 | 1,004-line mixed store, portability, route, projection, history, and rollback suite | Split without renaming catalog-selected tests in PR-07 |
| `testsupport/links.go` | 1 | Shared source-owner fixtures plus unused compatibility expectations | Retain live fixtures; delete unused expectation types/constants in PR-05 |
| `testsupport/performancefixture/production.go`, `testsupport/performancefixture/provider.go` | 2 | Frozen AC043 fixture validation contribution | Retain byte- and identity-equivalent behavior |

Empty untracked directories left by Iteration 1 package moves have no Git identity
and are not product work. The tracked `.gitkeep` is the only placeholder deletion
in this iteration.

### 13.3 Gap and Remediation Decisions

| Gap | Remediation and affected areas | Rationale and long-term benefit | Compatibility impact | Risk if unresolved | Completion validation |
| --- | --- | --- | --- | --- | --- |
| Current producers still emit compact link history values | Make Links the sole canonical mutation-value producer and return those values with each command. Migrate Entity-mentions, Decision supersession, Timeline supersession, and linked-note creation. Areas: specification, Links, callers, application composition, tests. | One source owner controls reconstruction vocabulary, eliminating partial maps and making future versioning explicit. | Old pre-production history and bundles are discarded; current public responses do not change. | New history remains dependent on permissive decoding and cannot guarantee exact rollback. | Every current producer persists the exact link member set; no production caller constructs a link history map. |
| Codec accepts `tag_id`, compact links, and restore defaults | Define exact closed current shapes, validate them before description or inverse application, and delete aliases/defaults. Areas: specification, codec, revision provider, tests. | Fail-closed decoding prevents silent attribution invention and converts schema growth into an explicit owner decision. | Deliberate hard break for old disposable values; no migration or shim. | Malformed history can be reported as reversible and restored with invented state. | Alias, compact, missing, malformed, and unknown-member fixtures all fail as not reversible. |
| Entity merge emits bare tag mutation targets | Emit `record_tag:<pre-mutation-record-id>:<record-tag-id>` for patch/delete and require the same grammar in the provider. Areas: Links merge, Revisions acceptance, tests. | Keeps collection identity deterministic and consistent with the current canonical item reference. | Old merge history is discarded by reset. | Merge-generated history uses a second target grammar and forces permanent fallback support. | Merge patch/delete, history lookup, Incident Bundle round trip, and rollback pass with composite targets; bare UUID fails. |
| Link commands expose persistence DTOs and post-write reloads | Return one narrow `RecordLinkCommandResult` with minimal identity and an optional canonical mutation. Remove raw row DTOs and value reloads after migration. Areas: Links API, direct callers, tests. | Commands become complete mutation capabilities, reducing race-prone reloads and accidental storage coupling. | Internal Go compile migration only. | Consumers keep duplicating serialization and newly added fields are omitted from history. | Create, patch, delete, and no-op cases return exact results; no retired DTO or loader remains. |
| Entity-mentions constructs Links internally and serializes Links values | Define an Entities-owned `LinkOperationsPort`, compose a Links adapter in application assembly, and inject one configured mention store into Timeline, mention routes, and Entity merge. Areas: Entities, application composition, Links adapter, tests, boundary policy. | Brings Entity-mentions into the same consumer-port/application-adapter model as Reporting, Timeline facts, and Entity merge effects. | Internal constructor/options change only. | Hidden construction, peer imports, and duplicate mention stores make future owner changes brittle. | Entity-mentions has no Links import; missing link operations fail composition; all mention lifecycle behavior remains equivalent. |
| Root package retains unused mutation paths and helpers | Delete non-mutation collection application, `TagStore`, boolean-only wrappers, low-level exported field/tag primitives, obsolete errors, specialized linked-note insertion, unused loaders, `.gitkeep`, and dead test expectations. Areas: implementation, tests, boundary policy. | Removes multiple ways to perform the same write and makes mutation/history accounting unavoidable. | Internal Go callers must use the canonical command paths; no forwarding shim. | A future caller can bypass revision-value production or adopt misleading APIs. | Exhaustive tracked-source/config/generator searches find no retired symbol; focused tests and boundary policy pass. |
| Fact reads have three reader types and embed consumer vocabulary | Use one Links-owned `FactReader` with incident, record, and collection-change methods. Return typed link values and owner-neutral tag-change facts; map `timeline.tags` in Timeline assembly. Areas: Links, application adapters, Assessment, Timeline, Reporting, tests. | One cohesive source-fact boundary is easier to extend without new consumer-specific readers. | Internal Go compile migration only; output remains equivalent. | Reader proliferation and consumer tokens in Links recreate peer coupling. | Ordering, non-nil empty sets, typed errors, active endpoint filtering, projection output, and history fields remain equivalent. |
| Mixed test file and unused support obscure ownership | Split tests by concern while keeping exact selected test names and remove dead support. Add guards for retired APIs and Entity-to-Links imports. Areas: tests, boundary policy, harness audit. | Smaller suites make failures and future additions easier to locate without changing verification accounting. | No row-ID or test-name change; no AC043 redistribution. | Cleanup regressions become hard to diagnose and deleted APIs can silently return. | Four Links rows and AC043 inputs remain exact; boundary and generated drift pass. |

### 13.4 Canonical History Contract

PR-01 MUST add the following owner rules to Core 02 and matching negative
acceptance to Core 04 before the strict implementation is admitted.

#### Record-link mutation value

The exact member set is:

1. `record_link_id`
2. `incident_id`
3. `src_record_id`
4. `dst_record_id`
5. `link_type`
6. `field_key`
7. `provenance`
8. `confidence`
9. `owner_user_id`
10. `created_by_user_id`
11. `decided_at`
12. `created_at`
13. `deleted_at`
14. `deleted_by_user_id`

Every member MUST be present. `field_key`, `confidence`, `deleted_at`, and
`deleted_by_user_id` MAY be JSON `null`; the remaining members MUST be non-null.
UUIDs MUST use canonical UUID strings. Timestamps MUST be UTC RFC 3339 values
using the source owner's canonical encoder. Link type, provenance, confidence,
endpoint, active-identity, and tombstone invariants MUST satisfy current Core 02.

#### Record-tag mutation value

The exact member set is:

1. `record_tag_id`
2. `incident_id`
3. `record_id`
4. `tag_name`
5. `normalized_tag_name`
6. `created_by_user_id`
7. `created_at`
8. `updated_at`
9. `deleted_at`
10. `deleted_by_user_id`

Every member MUST be present. Only `deleted_at` and `deleted_by_user_id` MAY be
JSON `null`. The target ID MUST be
`record_tag:<record_id>:<record_tag_id>`. Patch and delete entries use the
pre-mutation value's `record_id`; create entries use the after value's
`record_id`.

Both shapes are closed for the current target-semantics version. Unknown members,
`tag_id`, bare tag UUID targets, omitted nullable members, invalid values, and
partial maps MUST fail before history admission, import mutation, description, or
inverse application. A later shape requires an adopted owner revision and a new
explicit compatibility decision; permissive member acceptance is not an
extension mechanism.

### 13.5 Target Interfaces and Boundaries

The Links root retains typed commands, mutation results/errors, canonical item
references, immutable fact readers, thin Revision/Incident Bundle/Recovery
contributions, merge effects, and owner-local test support. It MUST NOT expose
source-row DTOs or alternate persistence primitives.

`RecordLinkCommandResult` has exactly these conceptual fields:

| Field | Meaning |
| --- | --- |
| `RecordLinkID` | Stable affected link identity |
| `SrcRecordID` | Canonical source endpoint |
| `DstRecordID` | Canonical destination endpoint |
| `LinkType` | Typed Links-owned relation token |
| `Mutation` | Optional canonical `RecordLinkMutation`; `nil` means the command was an idempotent no-op |

The command contract is:

- `UpsertLinkCommandTx` returns create or patch mutation data when state changes
  and `Mutation=nil` when the existing canonical state already matches.
- `InsertSupersedesCommandTx` returns a create mutation or an error.
- `TombstoneActiveLinkCommandTx` accepts exact incident, endpoints, link type,
  actor, and time and returns `(result, found, error)`; `found=false` is the sole
  absent-link result and carries no mutation.
- Field and tag collections continue returning `CollectionMutationResult` and
  become the only collection application paths.
- Linked-note creation uses the ordinary typed `references_artifact` upsert;
  there is no Links-local specialized linked-note method.
- All returned maps are fresh deep copies. A consumer cannot mutate a later
  result, provider value, or another caller's result through shared map storage.

Entity-mentions defines its own link command, result, mutation, and not-found
semantics in `LinkOperationsPort`. Application composition performs the only
Links-to-Entities conversion. `mentions.NewStore` requires
`mentions.WithLinkOperations`; it MUST fail fast if the port is absent. Timeline
assembly constructs one fully configured mention store and reuses it for Timeline
collections, mention routes, and `merge.WithMentionStore`. Entity merge no longer
constructs a mention store internally.

The consolidated Links `FactReader` exposes:

- `LoadIncidentTx(context, caller-owned transaction, incident_id) ActiveFacts`;
- `LoadRecordTx(context, caller-owned transaction, incident_id, record_id) ActiveFacts`;
- `LoadCollectionChangesTx(context, caller-owned transaction, incident_id,
  record_id, changed_at) CollectionChangeFacts`.

`CollectionChangeFacts` contains non-nil sorted unique `LinkFieldKeys` and a
`TagsChanged` boolean. It contains no Timeline or Assessment field key. Timeline
assembly maps `TagsChanged=true` to `timeline.tags`, merges it with link and
mention facts, deduplicates, and sorts the final field set. `RecordLinkFact` uses
`LinkType` and `LinkProvenance`, not raw strings. All methods return one typed
Links fact-read error without partial facts.

### 13.6 Refactor Requirements

| Requirement | Required outcome | Primary workstreams | Acceptance IDs |
| --- | --- | --- | --- |
| `LPR-REQ-001` | Current owner specifications define exact closed link/tag mutation values, target identities, and pre-production reset posture. | PR-01 | `LPR-AC-HISTORY-001..008` |
| `LPR-REQ-002` | Links alone creates link/tag mutation values and returns them atomically with commands. | PR-02, PR-03 | `LPR-AC-MUTATION-001..008` |
| `LPR-REQ-003` | Entity-mentions owns its port and has no Links implementation import or serializer. | PR-03 | `LPR-AC-ENTITY-001..008` |
| `LPR-REQ-004` | Every legacy alias, compact/defaulting decoder, target fallback, and old pre-production retained value is removed. | PR-04 | `LPR-AC-LEGACY-001..008` |
| `LPR-REQ-005` | Every listed dead facade, wrapper, loader, DTO, error, placeholder, and test-support value is absent without a shim. | PR-05 | `LPR-AC-DEAD-001..006` |
| `LPR-REQ-006` | One owner-neutral `FactReader` replaces the three current readers without output drift. | PR-06 | `LPR-AC-FACTS-001..008` |
| `LPR-REQ-007` | Test ownership remains exact while tests are split and retired surfaces receive durable guards. | PR-07 | `LPR-AC-TEST-001..006` |
| `LPR-REQ-008` | Public behavior, source-row portability, Recovery, authorization, transactions, and AC043 remain frozen through final validation. | All, PR-08 | `LPR-AC-FREEZE-001..006`, `LPR-AC-DONE-001..006` |

### 13.7 Binary Acceptance Matrix

| Acceptance ID | Pass condition |
| --- | --- |
| `LPR-AC-HISTORY-001` | Every link create, patch, and delete value has exactly the 14 named members with valid types and nullability. |
| `LPR-AC-HISTORY-002` | Every tag create, patch, and delete value has exactly the 10 named members with valid types and nullability. |
| `LPR-AC-HISTORY-003` | Link/tag values emitted by ordinary collections, mention resolution, supersession, linked-note creation, merge, and rollback use the same owner encoder. |
| `LPR-AC-HISTORY-004` | Tag merge patch/delete uses the pre-mutation record ID in its canonical composite target. |
| `LPR-AC-HISTORY-005` | Exact canonical values export, import, describe, inverse, and re-export without shape drift. |
| `LPR-AC-HISTORY-006` | Unknown members fail before history admission or Incident Bundle mutation. |
| `LPR-AC-HISTORY-007` | Nullable members remain explicitly present when null. |
| `LPR-AC-HISTORY-008` | Core 02 and Core 04 owner text is adopted before strict implementation completion. |
| `LPR-AC-MUTATION-001` | Upsert create returns one canonical create mutation. |
| `LPR-AC-MUTATION-002` | Upsert metadata change returns one canonical patch mutation. |
| `LPR-AC-MUTATION-003` | Idempotent upsert returns the current identity with no mutation. |
| `LPR-AC-MUTATION-004` | Active-link tombstone returns exact before/after values and absence returns `found=false`. |
| `LPR-AC-MUTATION-005` | Supersedes and linked-note callers consume the returned mutation without a value reload. |
| `LPR-AC-MUTATION-006` | Collection results preserve deterministic mutation order and exact no-op behavior. |
| `LPR-AC-MUTATION-007` | Mutating one returned map cannot alter another result or provider value. |
| `LPR-AC-MUTATION-008` | Every command runs entirely in the caller-owned transaction. |
| `LPR-AC-ENTITY-001` | `internal/modules/entities/mentions` has no production import of Links. |
| `LPR-AC-ENTITY-002` | Entity-mentions contains no link mutation-value serializer. |
| `LPR-AC-ENTITY-003` | Missing `LinkOperationsPort` fails composition before serving. |
| `LPR-AC-ENTITY-004` | One configured mention store is reused by Timeline collection, route, and merge composition. |
| `LPR-AC-ENTITY-005` | Resolve, idempotent resolve, retarget, dismiss, and revert produce equivalent active link state and route payloads. |
| `LPR-AC-ENTITY-006` | Mention-driven link mutations use exact canonical values and remain reversible. |
| `LPR-AC-ENTITY-007` | Mention projection, changed fields, collaboration intents, and row versions remain equivalent. |
| `LPR-AC-ENTITY-008` | Upstream authorization and incident-open checks remain in their current owners. |
| `LPR-AC-LEGACY-001` | `tag_id` mutation values fail closed. |
| `LPR-AC-LEGACY-002` | Compact link values fail closed. |
| `LPR-AC-LEGACY-003` | Missing field, confidence, tombstone, attribution, or timestamp members fail closed even when a former default exists. |
| `LPR-AC-LEGACY-004` | Invalid owner UUID and provenance fail rather than defaulting to the actor or `rollback`. |
| `LPR-AC-LEGACY-005` | Bare record-tag target UUIDs fail closed. |
| `LPR-AC-LEGACY-006` | No legacy helper, alias branch, fallback function, dual reader, or compatibility fixture remains. |
| `LPR-AC-LEGACY-007` | Existing pre-cutover databases are reset; no history row is rewritten. |
| `LPR-AC-LEGACY-008` | Pre-cutover bundles are rejected and regenerated; no version-2 permissive mode is retained. |
| `LPR-AC-DEAD-001` | `.gitkeep`, `TagStore`, `NewTagStore`, and `Store.Tags` are absent. |
| `LPR-AC-DEAD-002` | Non-mutation `Apply*CollectionTx` methods and boolean-only field/tag wrappers are absent. |
| `LPR-AC-DEAD-003` | `RecordLink`, `SupersedesLink`, old get/tombstone/value-load methods, and specialized linked-note insertion are absent. |
| `LPR-AC-DEAD-004` | Obsolete exported errors and lower-level record-returning field/tag helpers are private or deleted. |
| `LPR-AC-DEAD-005` | Unused test-support compatibility expectations are absent; live owner fixtures remain. |
| `LPR-AC-DEAD-006` | Tracked-source, configuration, generator, and boundary searches find no retired symbol or forwarding shim. |
| `LPR-AC-FACTS-001` | `FactReader.LoadIncidentTx` preserves Reporting facts and deterministic ordering. |
| `LPR-AC-FACTS-002` | `FactReader.LoadRecordTx` preserves Timeline facts, tags, and replacement selection. |
| `LPR-AC-FACTS-003` | Assessment derives only outbound `supported_by` targets from record facts. |
| `LPR-AC-FACTS-004` | Collection-change facts are incident-scoped, non-nil, sorted, and unique. |
| `LPR-AC-FACTS-005` | Links returns `TagsChanged`; only Timeline maps it to `timeline.tags`. |
| `LPR-AC-FACTS-006` | Empty facts are non-nil empty collections. |
| `LPR-AC-FACTS-007` | Any query/scan/iteration failure returns the typed Links error and no partial result. |
| `LPR-AC-FACTS-008` | Links contains no Assessment- or Timeline-named fact reader or field token. |
| `LPR-AC-TEST-001` | The four active Links verification rows retain their row IDs, test names, evidence classes, and ownership. |
| `LPR-AC-TEST-002` | Splitting files does not change catalog selectors or duplicate execution. |
| `LPR-AC-TEST-003` | Boundary policy rejects Entity-mention imports of Links. |
| `LPR-AC-TEST-004` | Boundary policy rejects every retired production symbol. |
| `LPR-AC-TEST-005` | Dead test support is removed without breaking cross-owner fixtures. |
| `LPR-AC-TEST-006` | AC043 owner, descriptor, counts, stride, order, receipt, digests, keys, and bindings are unchanged. |
| `LPR-AC-FREEZE-001` | HTTP and WebSocket paths, envelopes, statuses, error codes, replay behavior, and payload shapes remain equivalent. |
| `LPR-AC-FREEZE-002` | Caller-owned transactions and upstream authorization remain unchanged. |
| `LPR-AC-FREEZE-003` | Active Links source-row bundle paths, field order, attribution, and bytes remain unchanged for newly generated bundles. |
| `LPR-AC-FREEZE-004` | Recovery continues to register exactly `record_links` and `record_tags`. |
| `LPR-AC-FREEZE-005` | Reporting, Assessment, Timeline, projections, saved views, and frontend behavior remain equivalent. |
| `LPR-AC-FREEZE-006` | No historical migration or generated artifact is hand edited. |
| `LPR-AC-DONE-001` | Every workstream checkpoint passes Markdown lint before the next workstream begins. |
| `LPR-AC-DONE-002` | Every required focused owner row passes after its affected slice. |
| `LPR-AC-DONE-003` | Boundary and applicable generation/drift/shape checks pass. |
| `LPR-AC-DONE-004` | The pre-production reset and bundle regeneration posture is recorded with no migration artifact. |
| `LPR-AC-DONE-005` | `make agent-finalize`, `make test-fast`, the webserver-backed browser gate, and `make check` pass. |
| `LPR-AC-DONE-006` | Final handoff names changed files, commands, run roots, failures, resets, rollback points, and any remaining blocker. |

### 13.8 Workstreams and Sequence

Every row is a separate implementation workstream. After each row, including a
blocked or dropped row, update Sections 13.9 and 13.10 and run
`make lint-markdown`. The next row MUST NOT begin until that checkpoint passes.

| Workstream | Dependencies | Intended change and principal risk | Exit criteria |
| --- | --- | --- | --- |
| PR-00 Tracker update | None | Make Section 13 the current authority, record the live baseline, close design choices, and preserve Iteration 1 evidence. Risk: historical and current instructions remain ambiguous. | This section is decision-complete; only the tracker changed; Markdown lint and whitespace validation pass. |
| PR-01 Canonical history owner repair | PR-00 | Amend Core 02 and Core 04 with Section 13.4 exact shapes, target grammar, negative cases, and reset-only compatibility posture. Risk: implementation otherwise invents stored-history behavior. | Owner text is adopted, self-consistent, and maps `LPR-AC-HISTORY-001..008` plus all legacy negatives. |
| PR-02 Links-owned mutation results | PR-01 | Add canonical command results, atomic before/after production, deep-copy semantics, active tuple tombstoning, and complete create/patch/delete/no-op behavior. Keep old callers temporarily compiling only inside this slice. Risk: incomplete results or shared maps. | `LPR-AC-MUTATION-001..008` pass in Links tests; canonical result API exists before any decoder removal. |
| PR-03 Caller and Entity-mention migration | PR-02 | Migrate Timeline, Tasks/Decisions, Artifacts, and Entity-mentions; add the Entities-owned port and application adapter; compose one mention store and inject it into merge. Risk: route/history/projection drift or hidden nil dependencies. | All current producers consume Links mutations; Entity-mentions imports no Links; `LPR-AC-ENTITY-001..008` pass. |
| PR-04 Legacy hard cutover | PR-03 | Canonicalize merge tag targets, make the codec exact, delete aliases/defaults/fallbacks, update current fixtures, reset databases, and reject old bundles. Risk: removing a fallback before the last producer migrates or rewriting immutable history. | `LPR-AC-HISTORY-*` and `LPR-AC-LEGACY-*` pass; reset is recorded; no migration/backfill exists. |
| PR-05 Dead surface deletion | PR-04 | Delete every Section 13.3 dead facade, DTO, wrapper, loader, error, placeholder, and test-support value without a shim. Risk: an undiscovered caller or alternate unrevisioned write path. | `LPR-AC-DEAD-001..006` pass and direct source/config/generator searches are empty. |
| PR-06 Fact-reader consolidation | PR-03, PR-05 | Introduce the single owner-neutral reader, migrate Reporting/Timeline/Assessment adapters, move tag-field mapping to Timeline, and delete old readers. Risk: ordering, empty-set, error, or projection drift. | `LPR-AC-FACTS-001..008` and affected owner parity evidence pass. |
| PR-07 Test and boundary cleanup | PR-04 through PR-06 | Split the mixed Links test file, preserve test names/accounting, remove dead support, and add durable import/symbol guards. Risk: catalog drift or accidental AC043 change. | `LPR-AC-TEST-001..006`, boundary, drift, and exact Links rows pass. |
| PR-08 Validation and handoff | PR-07 | Run final narrow-to-broad validation, close all checklists, and retain reset/evidence/rollback details. Risk: declaring completion from partial or stale evidence. | Every `LPR-AC-FREEZE-*` and `LPR-AC-DONE-*` case passes; no TODO, blocker, legacy seam, or unrecorded skip remains. |

### 13.9 Iteration 2 Work Tracker

| Workstream | Status | Changed areas when executed | Required checkpoint evidence | Next executable slice |
| --- | --- | --- | --- | --- |
| PR-00 Tracker update | DONE | This tracker only | Markdown lint `.cartulary/test-results/20260821T211724Z-p1802679`; `git diff --check` passed; one-file diff audit passed | PR-01 after separate implementation authorization |
| PR-01 Canonical history owner repair | DONE | Core 02 exact Links mutation contracts; Core 04 `AC-554..AC-557` and claim mappings | Markdown lint `.cartulary/test-results/20260821T214452Z-p1813175`; `git diff --check` passed; owner acceptance mapped | PR-02 |
| PR-02 Links-owned mutation results | DONE | Canonical Links command results, complete internal state, tuple tombstoning, mutation-aware collections, recursive copies, focused tests; compile-only caller adaptations | Links rows `.cartulary/test-results/20260821T220018Z-p1986280` and `.cartulary/test-results/20260821T220102Z-p2026043`; format and boundary passed | PR-03 |
| PR-03 Caller and Entity-mention migration | DONE | Canonical mutation consumption in Assessment, Timeline, Artifacts, Tasks/Decisions, Entity-mentions, Entity merge, and application composition | Affected owner slices, boundary, format, and Markdown checkpoint passed | PR-04 |
| PR-04 Legacy hard cutover | DONE | Exact Links codec/provider/merge targets, owner validation through Revisions, fresh-state reset and bundle evidence | Links, Revisions, and Incident Bundle slices passed on reset state; boundary and Markdown checkpoint passed | PR-05 |
| PR-05 Dead surface deletion | DONE | Deleted alternate Links mutation facades, DTOs, loaders, exported sentinels, placeholder, and dead test expectations; canonicalized characterization coverage | Zero-consumer searches, all Links rows, boundary, fast tests, format, and Markdown checkpoint passed | PR-06 |
| PR-06 Fact-reader consolidation | DONE | One typed owner-neutral Links fact reader; Reporting, Timeline, and Assessment adapters; incident-scoped collection changes and assembly-owned tag mapping | Affected owner rows, parity, boundary, fast tests, source audit, and Markdown checkpoint passed | PR-07 |
| PR-07 Test and boundary cleanup | DONE | Concern-specific Links tests and durable Entity-import/retired-symbol boundary rules; no test-family ownership change | Exact four Links rows, boundary, generation/drift/shape, AC043 audit, fast tests, and Markdown checkpoint passed | PR-08 |
| PR-08 Validation and handoff | DONE | Deterministic Entity merge history fixture chronology and final tracker evidence | Focused Entities rows, retained-run finalization, fast and browser gates, authoritative full check, freeze audits, and final Markdown/whitespace checks passed | None |

### 13.10 Checkpoint and Handoff Protocol

Each workstream handoff entry MUST record:

- workstream ID and status;
- exact changed files and substantive changes;
- commands and selected row IDs;
- run roots and summary artifacts;
- failures, including whether they are related to the slice;
- database reset or bundle-regeneration action when applicable;
- reversible rollback point for source movement;
- non-reversible compatibility consequence of the approved hard cutover;
- blockers and the next executable slice.

The only allowed statuses are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`,
`DEFERRED`, and `DROPPED`. A skipped required service-backed check leaves the
workstream `BLOCKED`, not `DONE`. Because all environments are confirmed
disposable, discovery of pre-cutover data does not open a migration branch; the
required action is reset and regeneration.

### 13.11 Validation Order

1. After every specification or tracker checkpoint, run `make lint-markdown`.
2. Use `make task-guide ROLE=module-author OWNER=<owner-id>` before selecting
   narrow owner evidence.
3. Run `make test-slice OWNER=module.links` and
   `make service-backed-test-slice OWNER=module.links`, preserving all four active
   Links rows.
4. Run affected owner slices for Revisions, Entities, Timeline, Assessments,
   Artifacts, Tasks and Decisions, Incident Bundles, and Reporting.
5. After package, caller, or policy changes, run
   `make backend-module-boundary-check`.
6. When authored contract or harness inputs change, run `make generate`,
   `make generate-drift`, `make generated-artifact-policy-check`, and
   `make json-shape-check`. Generated files are never hand edited.
7. Run `make db-reset` at the PR-04 checkpoint, then rerun all required
   service-backed evidence against fresh state. Do not add or edit a migration.
8. Run `make agent-finalize`; pass a retained successful `RESULTS_DIR` when one
   exists, otherwise record the skipped retained-run maintenance explicitly.
9. Finish with `make test-fast`, `make browser-e2e-webserver-backed`, and
   `make check`.

### 13.12 Iteration 2 Definition of Done

- [x] PR-00 records a passing document-only checkpoint.
- [x] PR-01 adopts exact current mutation value and reset rules.
- [x] PR-02 makes Links the sole canonical link/tag mutation-value producer.
- [x] PR-03 brings Entity-mentions and all other current producers onto the
  canonical result boundary.
- [x] PR-04 removes every legacy reader, alias, default, target fallback, old
  retained value, and old bundle expectation through a reset-only cutover.
- [x] PR-05 deletes every proven dead API and placeholder without a shim.
- [x] PR-06 leaves one owner-neutral Links fact reader and no consumer token.
- [x] PR-07 preserves exact verification ownership and AC043 while adding durable
  boundary guards.
- [x] PR-08 passes all focused and broad gates and publishes a complete handoff.

### 13.13 Iteration 2 Handoff Log

| Date and workstream | Status and substantive change | Changed files | Validation and retained evidence | Rollback point | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-21 PR-00 | DONE. Made Section 13 the sole current forward authority; recorded the live post-Iteration-1 inventory, pre-production hard cutover, exact history shapes, target interfaces, workstreams, acceptance, reset posture, and completion gates. | `docs/handoffs/links-module-refactor-tracker.md` only | `make lint-markdown` passed at `.cartulary/test-results/20260821T211724Z-p1802679`; `git diff --check` passed; `git diff --name-only` returned only this tracker. Planning baselines are recorded in Section 13.1. | Revert the PR-00 tracker-only diff; no product, schema, data, generated, harness, or other documentation state changed. | Obtain separate implementation authorization, then execute PR-01 and checkpoint it before PR-02. |
| 2026-08-21 PR-01 | DONE. Adopted exact closed record-link and record-tag mutation values, canonical tag target grammar, operation and validation invariants, exact restoration, and reset-only legacy rejection. Added `AC-554..AC-557` and mapped them to Base and Incident Portability claims. No database reset or bundle regeneration occurred in this specification-only slice. | `docs/spec/02_domain_model_schema_and_history.md`; `docs/spec/04_security_deployment_and_conformance.md`; this tracker | `make lint-markdown` passed at `.cartulary/test-results/20260821T214452Z-p1813175`; `git diff --check` passed. No owner runtime row applied to a specification-only change. | Revert the two owner-document edits and this PR-01 checkpoint while retaining PR-00. No schema, data, generated artifact, harness, or product source changed. The adopted hard cutover intentionally provides no compatibility path once PR-04 resets state. | Execute PR-02 Links-owned mutation results; checkpoint it before PR-03. |
| 2026-08-21 PR-02 | DONE. Added `RecordLinkCommandResult`, canonical create/patch/delete/no-op mutations, exact active-tuple tombstoning, complete private link/tag row state, scalar-safe canonical encoding, recursive copy isolation, and mutation-aware collections that capture values from locked/returned state without post-write value reloads. `auto_match` now requires exactly `100`. Minimal caller signature adaptations preserve compilation; mutation consumption remains PR-03. No database reset or bundle regeneration occurred, and no compatibility artifact was added. | `internal/modules/links/commands.go`; `internal/modules/links/record_link_commands.go`; `internal/modules/links/store.go`; `internal/modules/links/field_refs.go`; `internal/modules/links/collection_mutations.go`; `internal/modules/links/internal/valuecodec/valuecodec.go`; `internal/modules/links/internal/valuecodec/valuecodec_test.go`; `internal/modules/links/links_tags_test.go`; `internal/app/assessmentassembly/adapters.go`; `internal/app/timelineassembly/assembly.go`; `internal/modules/entities/mentions/ports.go`; `internal/modules/tasksdecisions/mutation_capabilities.go`; this tracker | `make task-guide ROLE=module-author OWNER=module.links`; `make format` passed at `.cartulary/test-results/20260821T220007Z-p1982168`; all four active Links rows passed through `make test-slice OWNER=module.links` at `.cartulary/test-results/20260821T220018Z-p1986280` and `make service-backed-test-slice OWNER=module.links` at `.cartulary/test-results/20260821T220102Z-p2026043`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260821T220011Z-p1985865`; `git diff --check` passed. Related intermediate failures: `.cartulary/test-results/20260821T215234Z-p1825410` exposed caller assignment mismatches and `.cartulary/test-results/20260821T215827Z-p1980961` exposed a file-scoped `records` boundary; both were repaired and rerun. | Revert the twelve PR-02 source/test files and this checkpoint while retaining PR-00/PR-01. This slice changed no schema, generated artifact, database, or bundle. After PR-04, operational rollback instead requires source revert plus another disposable-state reset because no backward conversion will exist. | Execute PR-03 caller and Entity-mention migration; checkpoint it before PR-04. |
| 2026-08-21 PR-03 | DONE. Migrated Assessment support creation, Timeline auto-resolution, explicit mention lifecycle, clipboard collections and supersession, Decision supersession, and contextual linked-note creation to consume Links-owned canonical mutations in their owning change sets. Entity-mentions now owns `LinkOperationsPort` and no Links import or serializer; Timeline assembly creates one configured mention store and injects it into collection operations, routes, and required Entity merge composition. No database reset or bundle regeneration occurred, and no compatibility shim was added. | `internal/app/assessmentassembly/adapters.go`; `internal/app/timelineassembly/assembly.go`; `internal/app/workbookassembly/assessment_facade.go`; `internal/modules/artifacts/mutation_create.go`; `internal/modules/artifacts/mutation_shared.go`; `internal/modules/assessments/assessments_integration_test.go`; `internal/modules/assessments/export_surface_test.go`; `internal/modules/assessments/facade.go`; `internal/modules/assessments/facade_contract_test.go`; `internal/modules/entities/mentions/mention_lifecycle.go`; `internal/modules/entities/mentions/mention_resolution.go`; `internal/modules/entities/mentions/ports.go`; `internal/modules/entities/mentions/ports_test.go`; `internal/modules/entities/merge/merge_protected_set_composition_test.go`; `internal/modules/entities/merge/merge_protected_set_test.go`; `internal/modules/entities/merge/ports.go`; `internal/modules/entities/merge/store.go`; `internal/modules/tasksdecisions/mutation_capabilities.go`; `internal/modules/tasksdecisions/supersede_facade.go`; `internal/modules/timeline/clipboard_paste_store.go`; `internal/modules/timeline/lifecycle_store.go`; `internal/modules/timeline/mentions_collections_store.go`; `internal/modules/timeline/performance_fixture_store.go`; `internal/modules/timeline/ports.go`; `internal/modules/timeline/resolution_integration_test.go`; `internal/modules/timeline/store.go`; this tracker | Task guides ran for `module.links`, `module.entities`, `module.timeline`, `module.assessments`, `module.artifacts`, and `module.tasksdecisions`. Non-service slices passed: Links `.cartulary/test-results/20260821T221452Z-p2076443`, Entities `.cartulary/test-results/20260821T222416Z-p2401947`, Timeline `.cartulary/test-results/20260821T222704Z-p2496275`, Assessments `.cartulary/test-results/20260821T222602Z-p2455045`, Artifacts `.cartulary/test-results/20260821T221452Z-p2076536`, and Tasks/Decisions `.cartulary/test-results/20260821T222129Z-p2358160`. Service-backed slices passed: Links `.cartulary/test-results/20260821T223140Z-p2552926`, Entities `.cartulary/test-results/20260821T223633Z-p2644825`, Timeline `.cartulary/test-results/20260821T223819Z-p2697528`, Assessments `.cartulary/test-results/20260821T224300Z-p2753848`, Artifacts `.cartulary/test-results/20260821T224631Z-p2810430`, and Tasks/Decisions `.cartulary/test-results/20260821T224710Z-p2825545`. `make format` passed at `.cartulary/test-results/20260821T222404Z-p2398185`; boundary passed at `.cartulary/test-results/20260821T222223Z-p2396891`; Markdown checkpoint passed at `.cartulary/test-results/20260821T224849Z-p2864331`; zero-consumer and Entity-to-Links searches plus `git diff --check` passed. Repaired product evidence: Timeline’s first run `.cartulary/test-results/20260821T221452Z-p2076492` exposed old mutation-count assertions. Parallel Assessment and Tasks/Decisions runs `.cartulary/test-results/20260821T221452Z-p2076490` and `.cartulary/test-results/20260821T221452Z-p2076541` contended while building test-service images; serial reruns passed. Entities `.cartulary/test-results/20260821T223225Z-p2591306` and Artifacts `.cartulary/test-results/20260821T224358Z-p2794767` had object-store readiness timeouts; immediate serial reruns passed with no source change. | Revert the listed PR-03 application, module, and test edits plus this checkpoint while retaining PR-00 through PR-02. This slice changed no schema, migration, generated artifact, database, or retained bundle. Once PR-04 removes compatibility, operational rollback requires source revert and another disposable-state reset; no backward data conversion will exist. | Execute PR-04 legacy hard cutover, reset disposable databases, regenerate retained bundles, and checkpoint before PR-05. |
| 2026-08-21 PR-04 | DONE. Replaced permissive Links history decoding with exact closed 14-member link and 10-member tag values, canonical scalar and operation validation, exact restoration of creation attribution/timestamps, reversible link patches, and composite pre-mutation merge tag targets. Added the pure Links validator to Revisions generic target semantics and invoked it before local append, description, inverse planning/application, and Incident Bundle writes without teaching Revisions Links grammar. Added strict negative import matrices, canonical merge patch/delete and rollback evidence, and fresh-state bundle round trips. Removed alias, compact, defaulting, and bare-target acceptance; no migration, backfill, history rewrite, bundle translation, fallback, or dual reader/writer was added. | `internal/modules/links/internal/valuecodec/valuecodec.go`; `internal/modules/links/internal/valuecodec/valuecodec_test.go`; `internal/modules/links/internal/revisionprovider/provider.go`; `internal/modules/links/internal/mergeeffects/merge_effects.go`; `internal/modules/links/revision_provider_contribution.go`; `internal/modules/links/revision_provider_contribution_test.go`; `internal/modules/revisions/provider_contributions.go`; `internal/modules/revisions/target_history_facets.go`; `internal/modules/revisions/target_semantics_catalog.go`; `internal/modules/revisions/target_semantics_compiler.go`; `internal/modules/revisions/target_semantics_lookup.go`; `internal/modules/revisions/appender_mutation_storage.go`; `internal/modules/revisions/incident_bundle_validation.go`; `internal/modules/revisions/incident_bundle_apply.go`; `internal/modules/revisions/rollback_query_companions.go`; `internal/modules/revisions/rollback_apply_nonrow.go`; `internal/modules/revisions/catalog_admission_test.go`; `internal/modules/revisions/history_components_test.go`; `internal/modules/revisions/incident_bundle_portability_test.go`; `internal/modules/revisions/integration_test.go`; `internal/modules/revisions/rollback_test.go`; `internal/modules/incidentbundles/routes_helpers_integration_test.go`; `internal/modules/incidentbundles/routes_admission_integration_test.go`; this tracker | Task guides ran for `module.links`, `module.revisions`, and `module.incidentbundles`. `make backend-unit` passed at `.cartulary/test-results/20260821T231709Z-p2996631`. Non-service slices passed: Links 14/14 at `.cartulary/test-results/20260821T232221Z-p3112596`, Revisions 27/27 at `.cartulary/test-results/20260821T233154Z-p3312265`, and Incident Bundles 8/8 at `.cartulary/test-results/20260821T232622Z-p3191653`. After the hard reset, service-backed slices passed: Links 13/13 at `.cartulary/test-results/20260821T232804Z-p3208982`, Revisions 20/20 at `.cartulary/test-results/20260821T232857Z-p3247870`, and Incident Bundles 6/6 at `.cartulary/test-results/20260821T233012Z-p3292594`; those Incident Bundle rows generated, imported, and re-exported current canonical bundles and no tracked retained bundle artifact exists. Latest format passed at `.cartulary/test-results/20260821T233150Z-p3308606`; boundary passed at `.cartulary/test-results/20260821T233120Z-p3307982`; Markdown checkpoint passed at `.cartulary/test-results/20260821T233410Z-p3356707`; `git diff --check`, migration-diff search, and legacy-branch searches passed. `make db-reset` required its Make-variable confirmation: unconfirmed attempts failed safely at `.cartulary/test-results/20260821T232723Z-p3207177` and `.cartulary/test-results/20260821T232729Z-p3207580`; `make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` passed at `.cartulary/test-results/20260821T232736Z-p3207988`. Repaired related failures: Links `.cartulary/test-results/20260821T225856Z-p2873400` exposed PostgreSQL timestamp encoding drift; Revisions `.cartulary/test-results/20260821T231836Z-p3020608` exposed noncanonical fixture chronology and a duplicate fixture ID; Incident Bundles `.cartulary/test-results/20260821T232307Z-p3151165` exposed generic target parsing of owner-validated composite IDs and `.cartulary/test-results/20260821T232438Z-p3170520` exposed a retired non-reversible tag expectation. | Revert the PR-04 source/test/tracker edits, then reset disposable database and bundle state again. The cutover deliberately has no backward conversion: operational rollback after this point is source revert plus reset, and pre-cutover history/bundles remain rejected. | Execute PR-05 dead surface deletion; checkpoint it before PR-06. |
| 2026-08-21 PR-05 | DONE. Deleted the non-mutation collection methods, boolean-only field/tag wrappers, `TagStore`, raw root link DTOs, old get/tombstone/value loaders, specialized linked-note insertion, exported internal sentinels, `.gitkeep`, and unused test expectation wrappers without aliases or forwarding shims. Merge now calls the private canonical tombstone state operation, and field-aware and Timeline characterization uses canonical mutation commands plus live test tokens/helpers. No database reset, bundle regeneration, schema, migration, generated artifact, contract, or harness input changed in this slice. | Deleted `internal/modules/links/.gitkeep` and `internal/modules/links/value_codecs.go`; changed `internal/modules/links/store.go`; `internal/modules/links/field_refs.go`; `internal/modules/links/collection_actions.go`; `internal/modules/links/collection_mutations.go`; `internal/modules/links/record_link_commands.go`; `internal/modules/links/merge_effects.go`; `internal/modules/links/internal/revisionprovider/provider.go`; `internal/modules/links/internal/valuecodec/valuecodec.go`; `internal/modules/links/links_tags_test.go`; `internal/modules/links/testsupport/links.go`; `internal/modules/timeline/resolution_integration_test.go`; this tracker | `make task-guide ROLE=module-author OWNER=module.links`; `make format` passed at `.cartulary/test-results/20260821T234228Z-p3362122`; all four active Links rows passed through `make test-slice OWNER=module.links` at `.cartulary/test-results/20260821T234249Z-p3366247` and `make service-backed-test-slice OWNER=module.links` at `.cartulary/test-results/20260821T234340Z-p3406424`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260821T234249Z-p3366350`; `make test-fast` passed 416/416 at `.cartulary/test-results/20260821T234427Z-p3444883`; the Markdown checkpoint passed at `.cartulary/test-results/20260821T234608Z-p3453784`; exact retired-symbol searches across tracked Go, configuration, generator, and policy inputs, `.gitkeep` absence, and `git diff --check` passed. No failure occurred in this slice. | Revert the listed PR-05 source/test deletions and edits plus this checkpoint while retaining PR-00 through PR-04. No state reset is needed to reverse this source-only slice; because PR-04 remains a hard cutover, any rollback across PR-04 still requires source revert plus disposable-state reset and offers no backward conversion. | Execute PR-06 fact-reader consolidation; checkpoint it before PR-07. |
| 2026-08-22 PR-06 | DONE. Replaced `ActiveFactReader`, `RecordFactReader`, `AssessmentFactReader`, and the Links Timeline-named history method with one `FactReader` exposing incident, record, and incident-scoped collection-change reads. Link facts now use typed Links link/provenance values; all reads provide non-nil empty collections and one `FactReadError` without partial facts. Reporting and Timeline fact adapters consume the unified reader, Assessment derives outbound `supported_by` targets from record facts, and Timeline assembly alone maps `TagsChanged` to `timeline.tags`. Active filtering, deterministic ordering, replacement selection, route/projection behavior, and reporting output remain covered. No database reset, bundle regeneration, schema, migration, generated artifact, contract, or harness input changed. | Deleted `internal/modules/links/assessment_facts.go` and `internal/modules/links/timeline_history_facts.go`; changed `internal/modules/links/active_facts.go`; `internal/modules/links/links_tags_test.go`; `internal/app/reportingassembly/links.go`; `internal/app/timelinefactassembly/links.go`; `internal/app/assessmentassembly/projection_source.go`; `internal/app/timelineassembly/assembly.go`; `internal/modules/timeline/ports.go`; `internal/modules/timeline/store.go`; `internal/modules/timeline/clipboard_paste_store.go`; this tracker | Task guides ran for `module.links`, `module.reporting`, `module.timeline`, and `module.assessments`. Latest non-service slices passed: Links 14/14 at `.cartulary/test-results/20260822T000547Z-p3820795`, Reporting 5/5 at `.cartulary/test-results/20260821T235154Z-p3506735`, Timeline 51/51 at `.cartulary/test-results/20260821T235154Z-p3506730`, and Assessments 27/27 at `.cartulary/test-results/20260821T235154Z-p3506743`. Service-backed slices passed: Links 13/13 at `.cartulary/test-results/20260822T000636Z-p3859488`, Reporting 4/4 at `.cartulary/test-results/20260821T235810Z-p3697424`, Timeline 29/29 at `.cartulary/test-results/20260821T235955Z-p3753470`, and Assessments 18/18 at `.cartulary/test-results/20260821T235853Z-p3712616`. Latest format passed at `.cartulary/test-results/20260822T000537Z-p3816978`; boundary passed at `.cartulary/test-results/20260821T235041Z-p3462677`; `make test-fast` passed 416/416 at `.cartulary/test-results/20260822T000726Z-p3898093`; the Markdown checkpoint passed at `.cartulary/test-results/20260822T000839Z-p3903449`; corrected owner-scoped reader/token searches and `git diff --check` passed. One preliminary all-application symbol search exited 1 because it intentionally encountered the live Entities mention-history method; the corrected Links-reader scope passed and required no source change. | Revert the listed PR-06 source/test edits and deletions plus this checkpoint while retaining PR-00 through PR-05. No state reset is needed for this source-only slice; rollback across the retained PR-04 hard cutover still requires source revert plus disposable-state reset and has no backward conversion. | Execute PR-07 test and boundary cleanup; checkpoint it before PR-08. |
| 2026-08-22 PR-07 | DONE. Split the 1,291-line mixed Links suite into store, merge, portability, route/projection/history, and shared-helper files without renaming any catalog-selected test. Added boundary-owner rules that reject production Entity-mention imports of Links, every retired Links declaration in the owner root, and calls to every retired mutation facade across production. The authored Links family, its four row IDs/selectors/evidence classes/ownership, generated topology, and AC043 profile, version, seed, contribution order/dependency, receipts, counts, stride-bearing implementation, digests, keys, verification bindings, and claim bindings remain unchanged. No database reset, bundle regeneration, product source, schema, migration, contract, test-family manifest, or generated artifact changed. | Deleted `internal/modules/links/links_tags_test.go`; added `internal/modules/links/store_test.go`; `internal/modules/links/merge_test.go`; `internal/modules/links/portability_test.go`; `internal/modules/links/route_projection_history_test.go`; `internal/modules/links/test_helpers_test.go`; changed `tools/backend_module_boundaries.json`; this tracker | `make format` passed at `.cartulary/test-results/20260822T001426Z-p3951683`; `make generate` passed at `.cartulary/test-results/20260822T001435Z-p3955392` with no generated diff; `make generate-drift` passed at `.cartulary/test-results/20260822T001530Z-p3959489`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260822T001530Z-p3959509`; `make json-shape-check` at `.cartulary/test-results/20260822T001530Z-p3959524`; and `make backend-module-boundary-check` at `.cartulary/test-results/20260822T001531Z-p3959778`. `make explain-test-owner OWNER=module.links` reports exactly four active rows: one browser, integration, store, and unit row with three service-backed. Those exact rows passed 14/14 at `.cartulary/test-results/20260822T001549Z-p3963716` and 13/13 service-backed at `.cartulary/test-results/20260822T001650Z-p4003193`. The AC043 runtime/assembler, snapshot-key, and source-owner-assembler rows passed at `.cartulary/test-results/20260822T001549Z-p3963719`; every named AC043 and Links accounting artifact is byte-identical to baseline `21c226ce353961f47e438afaf350760bc1e82c6f`. `make test-fast` passed 416/416 at `.cartulary/test-results/20260822T001734Z-p4041790`; the Markdown checkpoint passed at `.cartulary/test-results/20260822T001914Z-p4047559`; retired-source/import searches and `git diff --check` passed. The first post-split Links run failed at `.cartulary/test-results/20260822T001315Z-p3911579` because the route test file lacked the live Timeline test-support import; adding that import repaired all three Go row builds and the exact rerun passed. | Revert the PR-07 test split, boundary-owner edit, and this checkpoint while retaining PR-00 through PR-06. No data reset is needed; generated output did not change. Rollback across PR-04 still requires source revert plus disposable-state reset and provides no backward conversion. | Execute PR-08 final validation and handoff completion. |
| 2026-08-22 PR-08 | DONE. Closed every `LPR-AC-FREEZE-*` and `LPR-AC-DONE-*` case with current-source evidence. The broad gate exposed a stale Entity merge fixture whose database-default creation time followed its fixed mutation clock; the fixture now deterministically places link/tag creation one minute before merge, preserving strict production admission rather than weakening it. Public HTTP/WebSocket and frontend surfaces, authorization and caller-owned transaction ownership, active Links bundle sources, Recovery membership, projections/reporting behavior, and AC043 identity remain frozen. No migration, generated artifact, retained history, or bundle was rewritten; PR-04's reset and fresh bundle round trips remain authoritative. There is no blocker, skip, TODO workstream, compatibility seam, or next slice. | `internal/modules/entities/unit_test.go`; this tracker | The first `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260821T201046Z-p1616597` failed safely at `.cartulary/test-results/20260822T002038Z-p4050083` because the old run's source digest did not match current source; it changed no file. The first preliminary `make check` failed 636/637 at `.cartulary/test-results/20260822T002115Z-p4050867`, correctly identifying the related Entity merge fixture chronology. An initial task-guide call with `OWNER=entities` failed as a usage error without a run root; the corrected `make task-guide ROLE=module-author OWNER=module.entities` passed. `make format` passed at `.cartulary/test-results/20260822T002833Z-p4174107`; Entity rows passed 33/33 at `.cartulary/test-results/20260822T002841Z-p4177883` and 29/29 service-backed at `.cartulary/test-results/20260822T002841Z-p4177889`. A current-source preliminary `make check` passed 637/637 at `.cartulary/test-results/20260822T003032Z-p89495`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260822T003032Z-p89495` then passed at `.cartulary/test-results/20260822T003455Z-p201534`, validating performance/evidence and reporting zero generated updates. The prescribed final sequence passed: `make test-fast` 416/416 at `.cartulary/test-results/20260822T003515Z-p204499`, `make browser-e2e-webserver-backed` 58/58 at `.cartulary/test-results/20260822T003528Z-p205454`, and `make check` 637/637 at `.cartulary/test-results/20260822T003932Z-p257609`. Baseline diff audits found no change under migrations, generated roots, `apps/web`, `docs/domain.md`, Links active bundle-source files, or Recovery; the contribution still registers exactly `record_links` and `record_tags`. The final Markdown checkpoint passed at `.cartulary/test-results/20260822T004530Z-p364782`; `git diff --check` passed. | Revert the PR-08 fixture and tracker checkpoint while retaining PR-00 through PR-07; no additional state reset is needed for PR-08 alone. Any operational rollback across PR-04 remains source revert plus another disposable-state reset, with no backward conversion for rejected pre-cutover history or bundles. | None. Iteration 2 is complete. |

## 14. Iteration 3: Structural Production Hardening

This section is the sole current forward plan. It supersedes every future-work
statement, compatibility posture, execution instruction, and stale inventory in
Sections 1 through 13. Those sections remain immutable historical evidence
except for the navigation corrections that identify their completed status.

Iteration 3 removes remaining indirect history inference, peer-owner physical
source reads, open mutation carriers, duplicate vocabulary validation, dead
codec exports, and permissive portability seams. It does not add product
features. A behavior or compatibility surface is retained only when an adopted
owner or the freeze map below requires it.

### 14.1 Current Scope and Baseline

| Item | Current posture |
| --- | --- |
| Baseline commit | `0f92592f9fb1e6fec5ac7364d6cb330b4e946db1` on `main` |
| Branch posture | Baseline commit is synchronized with `origin/main`; the sole pre-execution diff is the user-owned Section 14 tracker addition and MUST be preserved |
| Live target inventory | 27 files and 5,906 lines under `internal/modules/links` |
| Iteration status | Planned; implementation not started |
| Authorized document-step change | This tracker only |
| Future implementation scope | Links, directly affected Timeline and Entity conflict composition, Tasks/Decisions fact composition, portability and Recovery assembly, tests, verification routing, and boundary policy |
| Owner posture | Core 01 and Core 02 already own the required source, portability, history, and link/tag behavior; no owner or domain amendment is planned |
| Compatibility posture | Atomic internal compile-time cutover; no alias, forwarding method, dual path, compatibility DTO, fallback, or inferred legacy state |
| Data posture | Current canonical database state and version-3 bundles remain valid; no migration, rewrite, reset, or regeneration is expected |
| Public behavior posture | Frozen by Section 14.3 |
| Stateless facade posture | `Store` remains the stateless implementation of consumer-owned capabilities; an empty receiver alone is not dead code |

Current-source planning evidence:

| Evidence | Result | Run root or repository fact |
| --- | --- | --- |
| Git baseline | Clean `main` at the baseline commit and synchronized with `origin/main` | Direct inspection |
| Links owner inventory | Four active rows: browser, integration, store, and unit; three are service-backed | `make explain-test-owner OWNER=module.links` |
| Links non-service slice | Pass, 14/14 | `.cartulary/test-results/20260828T004751Z-p2527679` |
| Backend module boundary | Pass, 3/3 | `.cartulary/test-results/20260828T004838Z-p2568366` |
| Markdown lint | Pass | `.cartulary/test-results/20260828T005220Z-p2569962` |

These runs establish planning health only. They do not prove an Iteration 3
change. Product checks are not rerun merely to validate this tracker-only edit.

### 14.2 Live Inventory and Disposition

| Current area | Current responsibility or liability | Iteration 3 disposition |
| --- | --- | --- |
| Root `Store` and command facades | Stateless implementation of multiple narrow consumer capabilities | Retain the facade; narrow result and value surfaces |
| `active_facts.go` | Active owner facts plus timestamp-correlated collection-change inference | Retain incident/record facts; delete timestamp inference |
| `commands.go` and `store.go` | Open string-backed link/provenance values and independent command validation | Replace with one closed vocabulary and command-admission layer |
| `collection_mutations.go` | Separate mutable link/tag DTOs with caller-owned maps | Replace with one closed immutable mutation value |
| `merge_effects.go` | A third generic mutation carrier and mutable result slices | Reuse the canonical mutation value and defensive accessors |
| `internal/valuecodec` | Canonical encoding, decoding, transition policy, PostgreSQL loading, restore plans, and dead helpers in one package | Keep a pure focused codec; move SQL and delete duplicate/dead surfaces |
| `internal/revisionprovider` | Restore mechanics plus wrappers that have no caller | Own private source loading; delete unused wrappers |
| Incident Bundle and Recovery files | Paths, tables, columns, identities, and invariant IDs repeated across four locations | Derive all source-state contributions from one validated manifest |
| Timeline conflict composition | Reconstructs collection changes by timestamp after Revisions already stored exact conflict facts | Consume the canonical Revisions conflict window and delete inference |
| Entities mention history helper | Repeats timestamp inference for Timeline conflict reconstruction | Delete with the Timeline legacy path |
| Tasks/Decisions source facts | Mutation and portability policy reads Links physical source state directly | Consume application-injected owner-neutral Links facts |
| Links verification family | Does not directly route the private codec package | Add owner-owned unit coverage and regenerate topology through Make |
| Boundary policy | Retired declarations are root-scoped and two Tasks/Decisions physical-table exceptions remain | Extend removal guards to the subtree and delete both exceptions |

### 14.3 Behavior and Artifact Freeze

| Surface | Required Iteration 3 outcome |
| --- | --- |
| HTTP, WebSocket, OpenAPI, and frontend | No route, payload, error, event, selector, or interaction change |
| Authorization | Existing upstream admission and outcomes remain unchanged |
| Transactions | Every Links operation continues to use the caller-supplied transaction; no begin, commit, rollback, retry, or publication moves into Links |
| PostgreSQL | No schema, migration, index, view, constraint, or persisted token change |
| Link/tag state | Exact active identity, endpoint eligibility, direction, field key, uniqueness, deletion tuple, tag normalization, and deterministic ordering remain owner-conformant |
| Revisions | Exact `record_link` 14-member values, `record_tag` 10-member values, operation rules, target grammar, timestamps, attribution, and mutation order remain byte-equivalent |
| Incident Bundle v3 | Exact three Links paths, path order, content roles, stable identities, valid row bytes, attribution, and import/export/re-export behavior remain unchanged |
| Recovery | `module.links` continues to contribute exactly `record_links` and `record_tags` |
| Reporting and projections | Existing owner-neutral facts, missing-reference behavior, eligibility, ordering, and output remain unchanged |
| AC043 | Profile, version, seed, contribution order, dependency, receipts, counts, stride, digests, keys, verification bindings, and claim bindings remain byte-identical |
| Generated artifacts | Never hand edited; only Make-owned regeneration may update declared outputs after an authored harness change |

No compatibility claim overrides this table. A proposed change outside it is
`BLOCKED: scope expansion` and requires separate authorization.

### 14.4 Gap and Remediation Decisions

| ID | Finding | Decision |
| --- | --- | --- |
| G3-01 | Timeline and Entities infer changed collection fields by equating source-row timestamps with `change_sets.created_at` | Delete the inference path and use retained Revisions conflict facts |
| G3-02 | Tasks/Decisions mutation and portability policy reads `record_links` and `active_record_links_v1` from its source package | Inject a Tasks/Decisions-owned fact port mapped from Links |
| G3-03 | Link, tag, and merge mutations expose mutable maps and three shapes | Replace them atomically with one Links-owned immutable representation |
| G3-04 | `LinkType` and `LinkProvenance` accept arbitrary strings; command, fact, codec, and portability paths can drift | Introduce one closed vocabulary with exact parse and serialization |
| G3-05 | The codec mixes pure encoding with SQL and duplicate restore plans; four exported helpers have no caller | Move SQL to the provider and delete caller-free or duplicate surfaces |
| G3-06 | Portability declares seven invariant IDs but validates only a subset explicitly and uses an unchecked prepared-value assertion | Validate every declared invariant through typed preparation and checked application |
| G3-07 | Bundle, import, descriptor, and Recovery metadata repeat source-state facts | Replace the literals with one validated source-state manifest |
| G3-08 | `Store` has no fields | Retain it as the stable stateless capability implementation; do not replace it with package globals |

Owner-specific field meaning remains with the field owner. Links centralizes
only the generic link/tag vocabulary and invariants adopted for its source
state.

### 14.5 Target Interfaces and Boundaries

#### Closed vocabulary

`LinkType` and `LinkProvenance` become closed enum-like values with an invalid
zero value. The public Links surface provides exact `ParseLinkType`,
`ParseLinkProvenance`, and `String` operations. Parsing rejects blank, unknown,
aliased, normalized, or differently cased input.

The exact link tokens remain:

1. `observed_on_host`;
2. `observed_as_identity`;
3. `references_indicator`;
4. `attached_evidence`;
5. `references_artifact`;
6. `derived_from`;
7. `merged_into`;
8. `supported_by`;
9. `references_record`;
10. `supersedes`.

The exact provenance tokens remain `manual`, `auto_match`, `import`,
`rollback`, and `system`. History and portability admit the full vocabulary.
Live mutation commands continue to admit only `manual` and `auto_match`, with
the current exact confidence rules.

Database and bundle readers MUST scan a string and parse it. They MUST NOT cast
an unchecked string into a typed value. An unknown persisted token returns the
existing owner-safe error family and no partial facts or mutations.

#### Canonical mutation value

Links exposes one `Mutation` value with unexported state and these read methods:

- `TargetKind() string`;
- `TargetID() string`;
- `OperationKind() string`;
- `BeforeValue() map[string]any`;
- `AfterValue() map[string]any`.

Nullable before/after values return `nil`. Non-nil maps are recursive defensive
copies. Result-level mutation accessors return copied slices. Callers may map
the immutable value into their own ports, but cannot mutate or construct the
Links-owned history payload.

`CollectionMutationResult` exposes one deterministically ordered mutation
sequence. `RecordLinkCommandResult` exposes `Mutation() (Mutation, bool)`.
Repoint-link and repoint-tag results expose defensive `[]Mutation` accessors
while retaining their existing count and invalidation facts. The following
types are deleted with no aliases:

- `RecordLinkMutation`;
- `RecordTagMutation`;
- `MergeMutation`.

#### Conflict boundary

Timeline's Revisions capability returns `[]conflicts.RevisionWindowRow`
directly. Timeline calls `conflicts.BuildCanonicalPatchConflictWindow` with its
field descriptors and snapshot projector, then maps the owner-neutral result
to the existing Timeline error/response surface.

The Revisions `ConflictFacts` array is the only collection-change source for a
revision window. The following legacy surfaces are deleted:

- `RecordRevisionWindowEntry`;
- `CollectionChangeFacts`;
- `FactReader.LoadCollectionChangesTx`;
- `LinkPort.LoadCollectionFieldsChangedTx`;
- `EntityMentionPort.LoadTimelineCollectionFieldsChangedTx`;
- their application adapters and timestamp-based SQL/tests.

#### Tasks/Decisions fact boundary

Tasks/Decisions owns `LinkFactsCapability` and a minimal immutable fact DTO
containing source record ID, destination record ID, link type, and optional
field key. Application composition maps `links.FactReader.LoadRecordTx` into
that DTO.

Mutation dependencies and the Tasks/Decisions Incident Bundle source port
receive this capability explicitly. Tasks/Decisions continues to interpret
Decision status, supersession legality, field ownership, and target types. No
Tasks/Decisions mutation or portability source file reads Links physical tables
or redefines the Links token catalog.

Projection and rollback providers that consume the adopted active view are
characterized but are outside this physical source-table slice.

#### Pure codec and source-state package

`internal/valuecodec` retains only canonical link/tag models, encoding,
decoding, target parsing, and history-transition validation. PostgreSQL loading
moves into private revision-provider repository code. Restoration consumes the
decoded canonical values directly.

The following caller-free or duplicate surfaces are deleted:

- `DecodeMutationValue`;
- `UUIDFromMap`;
- `revisionprovider.Provider.ValidateRecordLinkValue`;
- `revisionprovider.Provider.ParseRecordTagIdentity`;
- the provider's `RecordTagIdentity` alias;
- `RecordLinkRestorePlan` and `RecordTagRestorePlan` plus their decoders.

A private `internal/sourcestate` package owns one validated manifest, Incident
Bundle export/import/application, exact invariant mapping, and Recovery table
derivation. The root constructors become:

- `NewIncidentBundleSourcePort() (sourceport.Port, error)`;
- `RecoveryStateContribution() (recoverystate.Contribution, error)`.

Application assembly MUST wrap these startup failures with Links context.
Prepared import state has one private concrete type; Apply uses a checked type
assertion and returns an owner-safe error instead of panicking.

When one prepared input violates multiple declared invariants, validation
selects exactly one failure in this priority order: source identity, generic
link tuple, deletion tuple, tag normalization, tag-catalog equality, active
uniqueness, then endpoint and owner same-incident scope.

### 14.6 Refactor Requirements

**LPR3-REQ-001 — Authority and freeze.** Implementation MUST remain inside the
future scope and freeze map in Sections 14.1 and 14.3. It MUST NOT amend owners,
schemas, public transports, authorization, transaction ownership, current
bundle formats, Recovery membership, or AC043.

**LPR3-REQ-002 — Canonical conflict facts.** Timeline conflict reconstruction
MUST use the Revisions canonical conflict builder and retained conflict facts.
Timestamp correlation across change sets, record links, record tags, or Entity
mentions MUST NOT remain in production or test-support APIs.

**LPR3-REQ-003 — Links source ownership.** Tasks/Decisions mutation and
portability policy MUST consume a consumer-owned fact capability. Production
code outside Links MUST NOT read `record_links` or `record_tags` through an
exception added for this work. Consumer lifecycle and field policy MUST NOT
move into Links.

**LPR3-REQ-004 — Closed vocabulary.** One exact closed catalog MUST own all ten
link types and five provenance values. Command, fact, codec, history, merge, and
portability paths MUST parse or serialize through it and fail closed on an
unknown value.

**LPR3-REQ-005 — Immutable mutations.** Every Links-produced non-row mutation
MUST use the single closed representation, exact target and operation grammar,
recursive defensive copies, and deterministic order. No caller may construct
or corrupt the canonical before/after maps.

**LPR3-REQ-006 — Pure codec and dead-code deletion.** Valuecodec MUST contain no
SQL or transaction dependency. The exact dead and duplicate surfaces in
Section 14.5 MUST be deleted without forwarding functions, aliases, or copied
replacement DTOs.

**LPR3-REQ-007 — Source-state manifest.** One validated manifest MUST derive
exactly two Recovery tables, three ordered version-3 paths, their roles,
identities, columns, and all seven Links invariant IDs. Malformed, duplicate,
unsafe, empty, or internally inconsistent authored manifest input MUST make
startup fail.

**LPR3-REQ-008 — Fail-closed portability.** Prepare MUST admit exact typed and
canonical rows before Apply. It MUST validate source identity, endpoints, the
generic link tuple, active uniqueness, deletion tuples, tag normalization via
`fieldnorm.NormalizeTagLabel`, and exact tag-catalog equality. Every failure
MUST select its declared invariant deterministically, leave no partial state,
and never depend on database error text or panic.

**LPR3-REQ-009 — Verification and deletion proof.** Every removed surface,
forbidden source read, exact vocabulary, immutable copy boundary, manifest
derivation, portability invariant, valid-bundle round trip, and frozen public
surface MUST have owner-routed evidence. Generated topology MUST be produced
only from authored inputs through Make.

### 14.7 Binary Acceptance Matrix

| Acceptance ID | Binary pass condition | Primary evidence |
| --- | --- | --- |
| LPR3-AC-FREEZE-001 | No diff exists in migrations, generated contract roots, `apps/web`, public transport contracts, `docs/domain.md`, or adopted owners except Make-generated topology explicitly caused by the authored test-family change | Baseline diff audit |
| LPR3-AC-FREEZE-002 | Authorization, caller-owned transactions, reporting/projections, Recovery's exact two tables, and AC043 remain unchanged | Owner slices, boundary, broad check, artifact comparison |
| LPR3-AC-CONFLICT-001 | Timeline consumes `conflicts.RevisionWindowRow` and the canonical conflict builder without a copied revision-window DTO | Compile-time surface test and Timeline conflict tests |
| LPR3-AC-CONFLICT-002 | All five legacy conflict symbols in Section 14.5 and their timestamp SQL have zero tracked production/test consumers | Subtree guard and exact search |
| LPR3-AC-CONFLICT-003 | Scalar and collection conflicts preserve current field keys, base row, actor, time, error shape, and empty collection defaults | Timeline unit and service-backed conflict matrix |
| LPR3-AC-FACTS-001 | Tasks/Decisions mutations and source-port validation receive an injected consumer-owned Links fact capability | Surface contract and composition tests |
| LPR3-AC-FACTS-002 | `internal/modules/tasksdecisions/internal/source/facts.go` contains no `record_links` or `active_record_links_v1` read | Boundary policy and exact search |
| LPR3-AC-FACTS-003 | Both Tasks/Decisions paths are absent from `links-source-tables.allowed_paths` and the boundary check passes | Boundary check |
| LPR3-AC-VOCAB-001 | Exactly ten link and five provenance tokens parse and serialize round trip; zero, unknown, aliased, and case-shifted input fails | Pure vocabulary unit matrix |
| LPR3-AC-VOCAB-002 | Live commands admit only `manual` and `auto_match` with current confidence semantics; history/import admit the adopted full provenance set | Command, history, and portability tests |
| LPR3-AC-VOCAB-003 | Fact, codec, merge, and portability code has no independent token switch or unchecked typed cast | Surface guard and negative persisted-token tests |
| LPR3-AC-MUTATION-001 | Collection, command, and merge results use only `Mutation` and preserve existing sequence, targets, operations, and canonical values | Unit and service-backed history tests |
| LPR3-AC-MUTATION-002 | Mutating any returned map or slice cannot alter a retained result or a later read | Recursive copy-isolation tests |
| LPR3-AC-MUTATION-003 | Links-owned `RecordLinkMutation`, `RecordTagMutation`, and `MergeMutation` have zero declarations or consumers; independently owned Entity mutation types are outside this guard | Links subtree guard and exact qualified search |
| LPR3-AC-CODEC-001 | `internal/valuecodec` imports no `pgx`, contains no SQL, and its exact canonical/transition matrix still passes | Import guard and private package unit row |
| LPR3-AC-CODEC-002 | Every dead helper, provider wrapper, alias, and restore-plan type named in Section 14.5 is absent | Subtree guard and exact search |
| LPR3-AC-MANIFEST-001 | One manifest defensively derives exactly two tables and three paths with exact order, roles, versions, identities, columns, and invariant IDs | Manifest contract tests |
| LPR3-AC-MANIFEST-002 | Empty, unsafe, duplicate, unordered, or internally inconsistent manifest fixtures fail with the Links manifest error | Negative manifest matrix |
| LPR3-AC-PORT-001 | Every one of the seven declared invariant IDs has at least one exact pre-apply or post-apply negative case | Portability invariant matrix |
| LPR3-AC-PORT-002 | Wrong prepared-value types return an error without panic and every failed import leaves no partial Links state | Checked-apply and transaction-atomicity tests |
| LPR3-AC-PORT-003 | Valid current bundles export, import, and re-export with the same three paths and byte-equivalent content | Service-backed round trip |
| LPR3-AC-DONE-001 | The authored Links verification family directly routes the private vocabulary/codec tests and generated topology is current | Test-family explanation and generation drift |
| LPR3-AC-DONE-002 | Narrow affected-owner slices, service-backed slices, finalization, fast tests, browser freeze, and broad check pass at current source | Retained run roots |
| LPR3-AC-DONE-003 | The final handoff records every changed file, command, run root, failure, rollback point, blocker, and skipped check | Section 14.11 audit |

### 14.8 Workstreams and Sequence

| Workstream | Required change | Depends on | Required checkpoint |
| --- | --- | --- | --- |
| I3-00 | Capture exact surfaces, conflict behavior, source ownership, mutation bytes/order, bundle descriptor, Recovery membership, and pre-cutover removal-search hits | Document plan authorized | Characterization passes and expected pre-cutover hits are recorded; permanent guards activate in the deleting slice |
| I3-01 | Replace Timeline conflict inference with canonical Revisions conflict facts; delete Links/Entities timestamp paths | I3-00 | Links, Timeline, Entities, and Revisions narrow evidence |
| I3-02 | Add Tasks/Decisions fact port and composition; remove peer physical reads and boundary exceptions | I3-00 | Links and Tasks/Decisions narrow plus service-backed evidence |
| I3-03 | Introduce closed vocabulary and immutable mutation value; migrate all callers atomically | I3-01, I3-02 | Links and every affected caller owner compile/test evidence |
| I3-04 | Make valuecodec pure; move SQL; delete dead helpers, wrappers, aliases, and restore plans | I3-03 | Links and Revisions codec/history/rollback evidence |
| I3-05 | Add validated source-state manifest and strict typed portability; propagate fallible constructors | I3-03, I3-04 | Links, Incident Bundles, Recovery, and affected source-owner evidence |
| I3-06 | Route private unit tests, strengthen subtree/boundary guards, remove obsolete tests, and run final gates | I3-01 through I3-05 | Generated drift, boundary, narrow, service-backed, and broad evidence |

Workstreams execute in table order. A workstream MAY be split into smaller
reviewable commits, but its checkpoint MUST stay `IN_PROGRESS` until every
listed acceptance case passes. No later workstream may recreate an interface
deleted by an earlier one.

### 14.9 Iteration 3 Work Tracker

| Workstream | Status | Expected changed areas | Exit condition | Next executable slice |
| --- | --- | --- | --- | --- |
| I3-00 | DONE | Characterization tests and boundary/removal guards | Baseline contracts pass and pre-cutover removal-search hits are retained in the handoff | I3-01 |
| I3-01 | DONE | Links facts, Timeline conflict composition, Entities mentions, Revisions adapter, tests | LPR3-AC-CONFLICT-001..003 pass | I3-02 |
| I3-02 | DONE | Tasks/Decisions capabilities/source facts/source port, application composition, boundary policy, tests | LPR3-AC-FACTS-001..003 pass | I3-03 |
| I3-03 | DONE | Links commands, facts, collections, merge, all direct callers and adapters | LPR3-AC-VOCAB-* and LPR3-AC-MUTATION-* pass | I3-04 |
| I3-04 | DONE | Valuecodec, revision provider, history tests and verification routing | LPR3-AC-CODEC-001..002 pass | I3-05 |
| I3-05 | DONE | Private source state, root contributions, portability/Recovery assembly and tests | LPR3-AC-MANIFEST-* and LPR3-AC-PORT-* pass | I3-06 |
| I3-06 | DONE | Test-family owner input, Make-generated topology, boundary/removal guards, handoff | Every `LPR3-AC-*` case passes | None — Iteration 3 complete |

Statuses remain exactly `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`,
and `DROPPED`. A discovered owner contradiction sets the affected workstream to
`BLOCKED: owner contradiction`; the executor MUST NOT resolve it in
implementation or this tracker.

### 14.10 Validation Order and Definition of Done

After each implementation workstream:

1. run `make task-guide ROLE=module-author OWNER=<owner-id>` for every changed
   owner;
2. use `make explain-test-owner` and the narrowest `make test-slice` or
   `make service-backed-test-slice` selected by the guide;
3. run `make format` when authored Go or frontend sources changed;
4. run `make backend-module-boundary-check` when a production import, source
   table, or retired-symbol boundary changed;
5. run `make lint-markdown` after appending the checkpoint;
6. run `git diff --check` and a changed-file scope audit.

When the authored Links test-family manifest changes, the same workstream MUST
run `make generate`, `make generate-drift`,
`make generated-artifact-policy-check`, and `make json-shape-check`. Generated
outputs are reviewed but never hand edited.

Before broad end-of-run verification, run `make agent-finalize` with the
`RESULTS_DIR` from a successful current-source warm `make check` run. Then run:

1. `make test-fast`;
2. `make browser-e2e-webserver-backed`;
3. `make check`.

Iteration 3 is `DONE` only when:

- [x] every `LPR3-REQ-*` requirement maps to passing binary acceptance;
- [x] every legacy symbol and dead export named in Section 14.5 is absent;
- [x] Timeline uses retained conflict facts with no timestamp inference;
- [x] Tasks/Decisions mutation and portability code has no Links physical-table
  read or boundary exception;
- [x] one closed vocabulary and one immutable mutation representation remain;
- [x] valuecodec is pure and the source-state manifest is the only repeated-fact
  owner;
- [x] all seven portability invariants fail closed and valid bundle bytes remain
  equivalent;
- [x] current canonical data requires no migration or reset;
- [x] public, owner, generated, Recovery, and AC043 freezes pass;
- [x] all focused and broad validation has retained current-source evidence;
- [x] the final handoff records changed files, failures, run roots, rollback,
  blockers, and skipped checks.

### 14.11 Iteration 3 Handoff Log

| Date and workstream | Status and substantive change | Changed files | Validation and retained evidence | Rollback point | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-27 DOC-00 | DONE. Activated Section 14 as the sole forward authority and recorded the decision-complete Iteration 3 structural hardening plan. Marked Iterations 1 and 2 complete while retaining their inventories and evidence as historical. No product, owner, migration, generated, harness, or other documentation changed. | This tracker only | `make lint-markdown` passed at `.cartulary/test-results/20260828T010455Z-p2574028`; `git diff --check` and the tracker-only changed-file audit passed. Product tests were not rerun for this Markdown-only change; the current planning baselines remain recorded in Section 14.1. | Revert the DOC-00 tracker-only diff; no runtime or data rollback is required. | Obtain separate implementation authorization, then execute I3-00 and checkpoint it before I3-01 or I3-02. |
| 2026-08-27 I3-00 | DONE. Preserved and re-baselined the user-owned Section 14 diff at `0f92592f9fb1e6fec5ac7364d6cb330b4e946db1`; clarified that mutation removal is Links-qualified, recorded pre-cutover searches instead of activating impossible failing guards, fixed the portability validation priority, and added an owner-routed executable characterization of the exact three v3 paths, roles, identities, seven invariant IDs, and two Recovery tables. Existing Links mutation/history tests retain canonical value shape, order, and defensive database-read behavior; current Timeline, Entities, and Revisions owner rows retain the conflict payload baseline. Expected pre-cutover hits remain for the five conflict surfaces, three Links mutation carriers, and five Tasks/Decisions physical Links reads; independently owned Entity and Assessment mutation types are intentionally outside the removal scope. No owner, domain, public transport, migration, database, frontend, or AC043 source changed. | Added `internal/modules/links/iteration3_characterization_test.go`; changed `tools/test_families/module.links.json`; Make-generated `tools/execution_topology_render_index.json`; this tracker | One initial `make task-guide ROLE=module-author OWNER=links` invocation failed as a usage error because owner IDs require the `module.` prefix; the corrected Links, Timeline, Entities, and Revisions guides passed. `make format` passed at `.cartulary/test-results/20260828T013301Z-p2584562`; `make generate` at `.cartulary/test-results/20260828T013307Z-p2588550`; `make generate-drift` at `.cartulary/test-results/20260828T013321Z-p2591540`; generated policy at `.cartulary/test-results/20260828T013321Z-p2591571`; JSON shape at `.cartulary/test-results/20260828T013321Z-p2591564`; and boundary at `.cartulary/test-results/20260828T013336Z-p2595511`. Links passed 14/14 at `.cartulary/test-results/20260828T013336Z-p2595374` and 13/13 service-backed at `.cartulary/test-results/20260828T013433Z-p2637164`; Revisions passed 27/27 at `.cartulary/test-results/20260828T013524Z-p2677867`; Entities passed 40/40 at `.cartulary/test-results/20260828T013524Z-p2677865`; Timeline passed 52/52 at `.cartulary/test-results/20260828T013524Z-p2677866`; Markdown passed at `.cartulary/test-results/20260828T014147Z-p2838369`. The exact pre-cutover searches, `git diff --check`, and changed-file scope audit matched the expected inventory. | Revert the characterization test, authored family row, generated render index, and this checkpoint while retaining DOC-00. This slice changed no runtime or data and needs no migration, state reset, or bundle conversion. | Execute I3-01 only; checkpoint it before I3-02. |
| 2026-08-27 I3-01 | DONE. Timeline now receives the Revisions field resolver at composition, resolves its immutable descriptor set at startup, consumes raw `conflicts.RevisionWindowRow` values, and maps `BuildCanonicalPatchConflictWindow` output to its unchanged conflict surface. The canonical builder now fails safely when a base row is absent before retained facts are read. Deleted the copied Timeline revision DTO, Links collection-change fact API and timestamp SQL, Entity-mention timestamp reader, both application adapters, and their obsolete tests. Added owner-routed retained-fact coverage for scalar plus all four Timeline collections, same-timestamp rows, empty base collection cells, actor/time metadata, and missing-base failure. Permanent declaration/call guards now reject recreation of every retired path. No public transport, persisted history, migration, bundle, Recovery, frontend, owner, domain, or AC043 source changed. | Added `internal/modules/revisions/conflicts/conflict_window_test.go`; deleted `internal/modules/entities/mentions/timeline_history_facts.go`; changed `internal/modules/links/active_facts.go`, Links store/portability/route tests; Timeline ports/store/conflict/batch/test composition and integration tests; Revisions conflict builder; Timeline application assembly and adapters; server/import test composition; workbook and performance test composition; Entity test composition/export guard; `internal/testutil/httptestx/httptestx.go`; `tools/backend_module_boundaries.json`; `tools/test_families/module.revisions.json`; Make-generated `tools/execution_topology_render_index.json`; this tracker | Two preliminary focused Timeline runs failed safely at `.cartulary/test-results/20260828T014748Z-p2847135` and `.cartulary/test-results/20260828T014827Z-p2860263` on compile-only missing imports; adding the exact imports fixed them. Focused Timeline conflict passed 3/3 at `.cartulary/test-results/20260828T014857Z-p2872957`; focused Revisions facts passed 1/1 at `.cartulary/test-results/20260828T015130Z-p2897741`. Latest format passed at `.cartulary/test-results/20260828T015111Z-p2890791`; generate at `.cartulary/test-results/20260828T015115Z-p2894781`; generation drift at `.cartulary/test-results/20260828T015214Z-p2914828`; generated policy at `.cartulary/test-results/20260828T015214Z-p2914801`; JSON shape at `.cartulary/test-results/20260828T020237Z-p3319703`; boundary at `.cartulary/test-results/20260828T020237Z-p3319820`. Full slices passed: Links 14/14 at `.cartulary/test-results/20260828T015227Z-p2918994`, Revisions 27/27 at `.cartulary/test-results/20260828T015227Z-p2919013`, Entities 40/40 at `.cartulary/test-results/20260828T015227Z-p2919001`, Timeline 52/52 at `.cartulary/test-results/20260828T015227Z-p2919006`; service-backed Links 13/13 at `.cartulary/test-results/20260828T015715Z-p3121170`, Revisions 20/20 at `.cartulary/test-results/20260828T015715Z-p3121181`, Entities 31/31 at `.cartulary/test-results/20260828T015715Z-p3121159`, and Timeline 29/29 at `.cartulary/test-results/20260828T015715Z-p3121157`; Markdown at `.cartulary/test-results/20260828T020328Z-p3320784`. Exact retired-symbol/timestamp searches, `git diff --check`, and the freeze-scope audit passed. | Revert the I3-01 source, test, boundary, authored Revisions family, generated render-index, and checkpoint edits while retaining I3-00. No data rollback, migration reversal, or bundle conversion is required. | Execute I3-02 only; checkpoint it before I3-03. |
| 2026-08-27 I3-02 | DONE. Added the Tasks/Decisions-owned scalar `LinkFact` and `LinkFactsCapability` with explicit optional-field presence, plus one shared application adapter over `links.FactReader.LoadRecordTx`. Mutation dependencies and the fallible Tasks/Decisions Incident Bundle source-port constructor now require that capability. Decision lifecycle/supersession and portability policy interpret copied facts while retaining all Records and Tasks/Decisions endpoint, type, field, and lifecycle rules. Removed all five physical Links reads from Tasks/Decisions source policy and both source-table boundary exceptions. The adapter returns non-nil empty success and nil/no-partial failure, and startup fails closed when the capability is absent. Projection and rollback consumers of the adopted active view remain unchanged. No schema, migration, transport, frontend, owner, domain, Links storage, bundle bytes, Recovery, or AC043 source changed. | Added `internal/modules/tasksdecisions/internal/linkfacts/linkfacts.go`, `internal/app/tasksdecisionassembly/link_facts.go`, and its unit test; changed Tasks/Decisions source facts, Incident Bundle provider/root/test, mutation capability/construction/patch/supersession/helpers/composition/export inventory; `internal/app/workbookassembly/tasksdecisions_capabilities.go`; `internal/app/incidentportabilityassembly/catalog.go`; `tools/backend_module_boundaries.json`; `tools/test_families/module.tasksdecisions.json`; Make-generated `tools/execution_topology_render_index.json`; this tracker | Preliminary focused runs failed safely at `.cartulary/test-results/20260828T020910Z-p3329230` and `.cartulary/test-results/20260828T020935Z-p3333892` on compile-only state-shape/call-site errors; exact fixes preserved the prior policy types. One `make format` preflight rejected an overlong authored row ID before formatting and had no run root; shortening only the ID fixed it. The first tightened boundary run failed at `.cartulary/test-results/20260828T021214Z-p3348168` until the new private consumer-owned package was added to the two exact import allowlists. Latest format passed at `.cartulary/test-results/20260828T021146Z-p3340188`; generate at `.cartulary/test-results/20260828T021150Z-p3344151`; generation drift at `.cartulary/test-results/20260828T021214Z-p3347907`; generated policy at `.cartulary/test-results/20260828T021214Z-p3347927`; JSON shape at `.cartulary/test-results/20260828T021214Z-p3347966`; boundary at `.cartulary/test-results/20260828T021306Z-p3352986`. Adapter/composition passed 2/2 at `.cartulary/test-results/20260828T021203Z-p3347060`; focused portability passed 3/3 at `.cartulary/test-results/20260828T021306Z-p3352908`. Full Links passed 14/14 at `.cartulary/test-results/20260828T021353Z-p3370472` and 13/13 service-backed at `.cartulary/test-results/20260828T021353Z-p3370460`; Tasks/Decisions passed 21/21 at `.cartulary/test-results/20260828T021353Z-p3370491` and 15/15 service-backed at `.cartulary/test-results/20260828T021353Z-p3370493`; Markdown at `.cartulary/test-results/20260828T021538Z-p3534370`. Exact source-table searches, both exception-removal checks, `git diff --check`, and the freeze-scope audit passed. | Revert the I3-02 capability, adapter, source-policy, composition, test, boundary, authored family, generated render-index, and checkpoint edits while retaining I3-00/I3-01. No data rollback, migration reversal, or bundle conversion is required. | Execute I3-03 only; checkpoint it before I3-04. |
| 2026-08-27 I3-03 | DONE. Replaced open string aliases with one exact Links-internal enum catalog for all ten link types and five provenance values; root parsers and typed constants are the only production construction route, invalid zero values serialize empty, live commands still admit only `manual` and `auto_match`, and fact/state/merge/codec readers reject unknown persisted tokens without partial facts. Replaced Links-owned collection, command, and merge carriers with one internally constructible immutable `Mutation`; its target/operation/side grammar is closed, construction inputs and every result/value accessor recursively copy maps, slices, arrays, and pointers, and link invalidation maps are deeply copied. Migrated Timeline, Tasks/Decisions, Artifacts, mentions, Assessments, and Entity merge while preserving canonical values and mutation order. Deleted the three Links-owned retired carrier types without aliases, retained the stateless `Store`, and activated guards against their return and against production typed-string casts. No persisted bytes, schema, migration, transport, frontend, owner, domain, Recovery, bundle, reporting/projection, authorization, transaction, or AC043 source changed. | Added `internal/modules/links/internal/vocabulary/vocabulary.go` and tests plus `internal/modules/links/internal/mutationvalue/mutation.go` and tests; changed Links command/store/active-fact/field-reference/collection/merge/valuecodec implementations and Links store/merge tests; Timeline, Tasks/Decisions, Artifacts, Assessments, and Entity-merge application/caller adapters; `tools/backend_module_boundaries.json`; `tools/test_families/module.links.json`; Make-generated `tools/execution_topology_render_index.json`; this tracker | Compile-cutover runs failed safely at `.cartulary/test-results/20260828T022304Z-p3543797`, `.cartulary/test-results/20260828T022612Z-p3558228`, and `.cartulary/test-results/20260828T022640Z-p3570904` while direct callers were migrated. Links test compilation failed at `.cartulary/test-results/20260828T022738Z-p3583687`; the first immutable-copy execution exposed and fixed nil-interface cloning at `.cartulary/test-results/20260828T022945Z-p3605546` and `.cartulary/test-results/20260828T023041Z-p3622441`. The first Tasks/Decisions broad run at `.cartulary/test-results/20260828T023448Z-p3705897` found one typed-token/string assertion, fixed without runtime change. Latest format passed at `.cartulary/test-results/20260828T025612Z-p28266`; generate at `.cartulary/test-results/20260828T023256Z-p3644116`; drift at `.cartulary/test-results/20260828T025543Z-p24332`; generated policy at `.cartulary/test-results/20260828T025555Z-p27326`; JSON shape at `.cartulary/test-results/20260828T025555Z-p27324`; boundary at `.cartulary/test-results/20260828T024532Z-p3963494`. Focused vocabulary/mutation isolation passed 2/2 at `.cartulary/test-results/20260828T023310Z-p3647055`; Links passed 16/16 at `.cartulary/test-results/20260828T023400Z-p3664874` and 13/13 service-backed at `.cartulary/test-results/20260828T024540Z-p3963884`; Timeline 52/52 at `.cartulary/test-results/20260828T023558Z-p3747445` and 29/29 service-backed at `.cartulary/test-results/20260828T024717Z-p4045281`; Tasks/Decisions 21/21 at `.cartulary/test-results/20260828T024415Z-p3922372` and 15/15 service-backed at `.cartulary/test-results/20260828T024626Z-p4004530`; Artifacts 7/7 at `.cartulary/test-results/20260828T024038Z-p3806622` and 3/3 service-backed at `.cartulary/test-results/20260828T025200Z-p4103686`; Entities 40/40 at `.cartulary/test-results/20260828T024121Z-p3823547` and 31/31 service-backed at `.cartulary/test-results/20260828T025243Z-p4120218`; Assessments 27/27 at `.cartulary/test-results/20260828T024317Z-p3878971` and 18/18 service-backed at `.cartulary/test-results/20260828T025438Z-p4175173`; Markdown checkpoint at `.cartulary/test-results/20260828T025705Z-p32562`. Exact Links-qualified retired-symbol and production-cast searches, `git diff --check`, and the changed-file/freeze audit passed; no checks were skipped. | Revert the I3-03 vocabulary/mutation packages, caller cutover, tests, boundary, authored Links family, generated render index, and this checkpoint while retaining I3-00 through I3-02. This is source-only rollback; no data rollback, migration reversal, bundle conversion, or state reset is required. | Execute I3-04 only; checkpoint it before I3-05. |
| 2026-08-27 I3-04 | DONE. Removed all PostgreSQL and transaction dependencies from the Links valuecodec. Private revision-provider repository functions now own link/tag row loading, parse persisted vocabulary strings before constructing canonical values, and return no partial decoded state for unknown tokens. Inverse restoration decodes canonical link/tag values once and writes their exact fields directly; the duplicate restore-plan DTOs and decoders are gone. Deleted `DecodeMutationValue`, `UUIDFromMap`, the provider's `RecordTagIdentity` alias, `ValidateRecordLinkValue`, and `ParseRecordTagIdentity` wrappers without replacements. Added permanent import/declaration guards and direct owner routing for the complete six-test private codec matrix. Canonical maps, history transitions, rollback changed-field results, persisted values, and bundle bytes remain unchanged; no migration, reset, compatibility decoder, transport, frontend, owner, domain, Recovery, or AC043 source changed. | Added `internal/modules/links/internal/revisionprovider/repository.go`; changed `internal/modules/links/internal/revisionprovider/provider.go`, `internal/modules/links/internal/valuecodec/valuecodec.go` and its tests, `tools/backend_module_boundaries.json`, `tools/test_families/module.links.json`, Make-generated `tools/execution_topology_render_index.json`, and this tracker | No failed command or test occurred in this slice. Format passed at `.cartulary/test-results/20260828T030120Z-p36071`; generate at `.cartulary/test-results/20260828T030130Z-p40130`; drift at `.cartulary/test-results/20260828T030800Z-p252814`; generated policy at `.cartulary/test-results/20260828T030813Z-p255834`; JSON shape at `.cartulary/test-results/20260828T030813Z-p255841`; final boundary at `.cartulary/test-results/20260828T030813Z-p256015`. The directly routed codec row passed 1/1 at `.cartulary/test-results/20260828T030144Z-p43064`; Links passed 17/17 at `.cartulary/test-results/20260828T030201Z-p43927` and 13/13 service-backed at `.cartulary/test-results/20260828T030359Z-p132576`; Revisions passed 27/27 at `.cartulary/test-results/20260828T030253Z-p85859` and 20/20 service-backed at `.cartulary/test-results/20260828T030446Z-p173347`; Incident Bundles passed 8/8 at `.cartulary/test-results/20260828T030552Z-p218861` and 6/6 service-backed at `.cartulary/test-results/20260828T030652Z-p236041`. Exact codec SQL/import and dead-surface searches plus `git diff --check` passed; no checks were skipped. | Revert the I3-04 repository extraction, provider/codec/test changes, boundary, authored Links family, generated render index, and this checkpoint while retaining I3-00 through I3-03. This is source-only rollback; no database rollback, migration reversal, bundle conversion, or state reset is required. | Execute I3-05 only; checkpoint it before I3-06. |
| 2026-08-27 I3-05 | DONE. Replaced repeated Links portability and Recovery metadata with one private, fallibly validated, defensively projected source-state manifest owning the exact two authoritative tables, three ordered version-3 paths, roles, identities, columns, and invariant registry. Export and import SQL remain fixed code selected only by closed path kinds. Links now strictly decodes all three files to an opaque bound prepared value, rejects blank, multivalue, missing, unknown, wrongly typed, noncanonical, duplicate, illegal-tuple, catalog, uniqueness, and incident-scope input before persistence wherever possible, and selects multi-invalid failures in the required source/link/deletion/tag/catalog/unique/endpoint order. Apply uses a checked assertion, fixed manifest-derived import specifications, pre-write identity/uniqueness/endpoint checks, attribution remapping, affected-row enforcement, and final admitted-state/attribution comparison; the old unchecked assertion and portability package are deleted. The Links source-port and Recovery constructors are fallible, and both application assemblies context-wrap Links construction failures. The manifest now drives exact Recovery membership, descriptor, export order, and import shape. Private negative/defensive tests and a service-backed owner-export/import/re-export test prove no panic, no partial writes, exact attribution, and byte-equivalent valid v3 round trips. Existing direct-SQL Incident Bundle fixtures were corrected from noncanonical normalized tags to the same values produced by owner admission; no product compatibility decoder or normalization path was added. No owner specification, domain vocabulary, migration, public transport, frontend, authorization, transaction coordination, reporting/projection, or AC043 behavior changed. | Added `internal/modules/links/internal/sourcestate/manifest.go`, `prepare.go`, `port.go`, `manifest_test.go`, `prepare_test.go`, and `internal/modules/links/source_state_portability_integration_test.go`; deleted `internal/modules/links/internal/incidentbundle/portability.go` and `source_port.go`; changed Links root Incident Bundle/Recovery contributions, portability/characterization tests and test support; Incident Portability and Recovery assemblies; two Incident Bundles fixture/assertion tests; `tools/backend_module_boundaries.json`; `tools/test_families/module.links.json`; Make-generated `tools/execution_topology_render_index.json`; this tracker | One pre-format catalog check rejected the initially misspelled Incident Bundles collaborator ID without creating a retained run. Two preliminary service-backed Links runs at `.cartulary/test-results/20260828T032643Z-p333319` and `.cartulary/test-results/20260828T032749Z-p355282` exposed that current valid v3 PostgreSQL export uses canonical microsecond `+00:00` timestamps; the strict reader was corrected to admit exactly that frozen representation without changing bytes. The first boundary run at `.cartulary/test-results/20260828T033110Z-p398738` identified the two new owner tests missing from the source-port import allowlist. Incident Bundles runs at `.cartulary/test-results/20260828T033251Z-p447209` and `.cartulary/test-results/20260828T033515Z-p469450` exposed an over-broad pre-publication membership check plus noncanonical direct-SQL fixture values and their stale assertion; Links validation was restored to endpoint/owning-record scope and the invalid fixtures were made owner-canonical. Latest format passed at `.cartulary/test-results/20260828T033631Z-p487021`; generate at `.cartulary/test-results/20260828T032634Z-p330439`; drift at `.cartulary/test-results/20260828T034151Z-p673411`; generated policy at `.cartulary/test-results/20260828T034159Z-p676352`; JSON shape at `.cartulary/test-results/20260828T034200Z-p676794`; boundary at `.cartulary/test-results/20260828T034204Z-p677257`; migration drift at `.cartulary/test-results/20260828T034206Z-p677593`. The routed source-state unit rows passed 1/1 at `.cartulary/test-results/20260828T032456Z-p325331` and `.cartulary/test-results/20260828T032915Z-p376457`. Full Links passed 18/18 at `.cartulary/test-results/20260828T033205Z-p405197` and 13/13 service-backed at `.cartulary/test-results/20260828T033851Z-p562441`; Incident Bundles passed 8/8 at `.cartulary/test-results/20260828T033635Z-p491060` and 6/6 service-backed at `.cartulary/test-results/20260828T033936Z-p603321`; Recovery passed 24/24 at `.cartulary/test-results/20260828T033731Z-p508116` and 19/19 service-backed at `.cartulary/test-results/20260828T034031Z-p620060`. Links-scoped retired-package/helper searches, manifest SQL/import checks, exact three-path/two-table parity, `git diff --check`, and the changed-file/freeze audit passed. No required check was skipped. | Revert the I3-05 source-state package, root/app constructor cutover, portability/Recovery tests, fixture corrections, old-package deletion, boundary, authored Links family, generated render index, and this checkpoint while retaining I3-00 through I3-04. This is source-only rollback; no data rollback, migration reversal, bundle conversion, or state reset is required. | Execute I3-06 only; complete final validation and handoff before marking Iteration 3 done. |
| 2026-08-28 I3-06 | DONE. Replaced the temporary Iteration-3 characterization name/file with durable root contribution parity evidence, split it into its own owner-routed row, regenerated Make-owned topology, reconfirmed every permanent removal/import/source-table guard, and audited all `LPR3-REQ-*` requirements against the binary matrix. Final lint exposed and removed one unused I3-03 map-copy helper and corrected one I3-02 internal error string; no behavior or public diagnostic changed. All seven workstreams are `DONE`; no timestamp inference, physical peer read, retired carrier, dead codec/provider surface, unchecked prepared assertion, old source-state package, compatibility seam, migration, reset requirement, or next implementation slice remains. The exact cumulative inventory follows this table. | I3-06 changed `internal/modules/links/source_state_contributions_test.go`, `internal/app/entitymergeassembly/links.go`, `internal/modules/tasksdecisions/internal/source/facts.go`, `tools/test_families/module.links.json`, Make-generated `tools/execution_topology_render_index.json`, and this tracker; it removed the temporary untracked `internal/modules/links/iteration3_characterization_test.go`. The final cumulative 88-file inventory below is authoritative. | Final topology generation passed at `.cartulary/test-results/20260828T034518Z-p686936`; directly routed durable/private Links evidence passed 2/2 at `.cartulary/test-results/20260828T034527Z-p689840`; generation drift, generated policy, JSON shape, boundary, and migration drift passed at `.cartulary/test-results/20260828T034619Z-p690900`, `.cartulary/test-results/20260828T034627Z-p693836`, `.cartulary/test-results/20260828T034628Z-p694246`, `.cartulary/test-results/20260828T034632Z-p694708`, and `.cartulary/test-results/20260828T034634Z-p695045`. Current Links passed 18/18 at `.cartulary/test-results/20260828T034641Z-p697911` and 13/13 service-backed at `.cartulary/test-results/20260828T034725Z-p739047`. The first warm check reached 667/668 and failed only staticcheck ST1005 at `.cartulary/test-results/20260828T034817Z-p780049`; the next direct `make lint-go` exposed the now-deleted unused helper at `.cartulary/test-results/20260828T035334Z-p923679`. After both exact fixes, Go lint passed; Tasks/Decisions passed 21/21 at `.cartulary/test-results/20260828T035438Z-p935630` and 15/15 service-backed at `.cartulary/test-results/20260828T035526Z-p977196`; Entities passed 40/40 at `.cartulary/test-results/20260828T035612Z-p1017897` and 31/31 service-backed at `.cartulary/test-results/20260828T035805Z-p1073270`. The successful warm check passed 668/668 at `.cartulary/test-results/20260828T040001Z-p1128430`; retained-evidence finalization passed 1/1 at `.cartulary/test-results/20260828T040441Z-p1247561`; fast tests passed 438/438 at `.cartulary/test-results/20260828T040504Z-p1250591`; webserver-backed browser evidence passed 58/58 at `.cartulary/test-results/20260828T040521Z-p1251571`; the final check passed 668/668 at `.cartulary/test-results/20260828T040931Z-p1305876`. Exact dead-surface, timestamp-SQL, peer-table, generated/freeze-scope, changed-file, and `git diff --check` audits passed. The final tracker-only Markdown/diff audit is deliberately sequenced immediately after this completed row; its actual retained root is reported in the completion response because any self-recording edit would invalidate that final pass. There are no blockers and no skipped required checks. Product checks were intentionally not repeated after this final tracker-only handoff edit. | Revert the cumulative 88-file source diff against baseline `0f92592f9fb1e6fec5ac7364d6cb330b4e946db1`, including the three deletions, 16 additions, 69 modifications, authored test-family inputs, generated render index, and tracker. Rollback is source-only: no data rollback, migration reversal, bundle conversion, state reset, or external cleanup is required. | None. Iteration 3 remediation, validation, and handoff are complete; no compatibility seam or follow-on slice remains. |

#### I3-06 Exact Cumulative Changed-File Inventory

Added (16):

- `internal/app/tasksdecisionassembly/link_facts.go`
- `internal/app/tasksdecisionassembly/link_facts_test.go`
- `internal/modules/links/internal/mutationvalue/mutation.go`
- `internal/modules/links/internal/mutationvalue/mutation_test.go`
- `internal/modules/links/internal/revisionprovider/repository.go`
- `internal/modules/links/internal/sourcestate/manifest.go`
- `internal/modules/links/internal/sourcestate/manifest_test.go`
- `internal/modules/links/internal/sourcestate/port.go`
- `internal/modules/links/internal/sourcestate/prepare.go`
- `internal/modules/links/internal/sourcestate/prepare_test.go`
- `internal/modules/links/internal/vocabulary/vocabulary.go`
- `internal/modules/links/internal/vocabulary/vocabulary_test.go`
- `internal/modules/links/source_state_contributions_test.go`
- `internal/modules/links/source_state_portability_integration_test.go`
- `internal/modules/revisions/conflicts/conflict_window_test.go`
- `internal/modules/tasksdecisions/internal/linkfacts/linkfacts.go`

Deleted (3):

- `internal/modules/entities/mentions/timeline_history_facts.go`
- `internal/modules/links/internal/incidentbundle/portability.go`
- `internal/modules/links/internal/incidentbundle/source_port.go`

Modified (69):

- `docs/handoffs/links-module-refactor-tracker.md`
- `internal/app/assessmentassembly/adapters.go`
- `internal/app/entitymergeassembly/links.go`
- `internal/app/importassembly/tasksdecisions_integration_test.go`
- `internal/app/incidentportabilityassembly/catalog.go`
- `internal/app/recoveryassembly/state_catalog.go`
- `internal/app/server/runtime_assembly.go`
- `internal/app/timelineassembly/assembly.go`
- `internal/app/timelineassembly/core_adapters.go`
- `internal/app/timelineassembly/entity_evidence_adapters.go`
- `internal/app/timelineassembly/relationship_adapters.go`
- `internal/app/workbookassembly/tasksdecisions_capabilities.go`
- `internal/modules/artifacts/mutation_collections.go`
- `internal/modules/artifacts/mutation_create.go`
- `internal/modules/artifacts/mutation_patch.go`
- `internal/modules/entities/boundary_guard_test.go`
- `internal/modules/entities/unit_support_test.go`
- `internal/modules/incidentbundles/routes_admission_integration_test.go`
- `internal/modules/incidentbundles/routes_helpers_integration_test.go`
- `internal/modules/links/active_facts.go`
- `internal/modules/links/collection_mutations.go`
- `internal/modules/links/commands.go`
- `internal/modules/links/field_refs.go`
- `internal/modules/links/incident_bundle_source_port.go`
- `internal/modules/links/internal/mergeeffects/merge_effects.go`
- `internal/modules/links/internal/revisionprovider/provider.go`
- `internal/modules/links/internal/valuecodec/valuecodec.go`
- `internal/modules/links/internal/valuecodec/valuecodec_test.go`
- `internal/modules/links/merge_effects.go`
- `internal/modules/links/merge_test.go`
- `internal/modules/links/portability_test.go`
- `internal/modules/links/record_link_commands.go`
- `internal/modules/links/recovery_state.go`
- `internal/modules/links/route_projection_history_test.go`
- `internal/modules/links/store.go`
- `internal/modules/links/store_test.go`
- `internal/modules/links/testsupport/links.go`
- `internal/modules/revisions/conflicts/conflict_window.go`
- `internal/modules/tasksdecisions/exported_surface_test.go`
- `internal/modules/tasksdecisions/import_create.go`
- `internal/modules/tasksdecisions/incident_bundle_contribution.go`
- `internal/modules/tasksdecisions/incident_bundle_source_port_test.go`
- `internal/modules/tasksdecisions/internal/providers/incidentbundle/source_port.go`
- `internal/modules/tasksdecisions/internal/source/facts.go`
- `internal/modules/tasksdecisions/mutation_capabilities.go`
- `internal/modules/tasksdecisions/mutation_collections.go`
- `internal/modules/tasksdecisions/mutation_composition_test.go`
- `internal/modules/tasksdecisions/mutation_construction.go`
- `internal/modules/tasksdecisions/mutation_create.go`
- `internal/modules/tasksdecisions/mutation_patch.go`
- `internal/modules/tasksdecisions/mutation_supersede.go`
- `internal/modules/tasksdecisions/mutations.go`
- `internal/modules/timeline/batch_mutation_store.go`
- `internal/modules/timeline/conflict_values.go`
- `internal/modules/timeline/ports.go`
- `internal/modules/timeline/store.go`
- `internal/modules/timeline/store_patch.go`
- `internal/modules/timeline/store_test.go`
- `internal/modules/timeline/test_composition_test.go`
- `internal/modules/timeline/timeline_event_integration_test.go`
- `internal/modules/workbook/notes_indicators_test.go`
- `internal/testutil/appsupport/performancefixture/owners.go`
- `internal/testutil/appsupport/workbook.go`
- `internal/testutil/httptestx/httptestx.go`
- `tools/backend_module_boundaries.json`
- `tools/execution_topology_render_index.json`
- `tools/test_families/module.links.json`
- `tools/test_families/module.revisions.json`
- `tools/test_families/module.tasksdecisions.json`
