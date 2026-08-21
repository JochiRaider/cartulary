# links Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Posture |
| --- | --- |
| Target path | `internal/modules/links` |
| Target label | `links`, derived from the final path segment and normalized to lowercase kebab case |
| Output path | `docs/handoffs/links-module-refactor-tracker.md` |
| Status | Remediation complete; final evidence retained |
| Allowed change in this session | Adopted owner specifications, authored contracts, implementation, tests, application composition, boundary policy, and this tracker, one workstream at a time. Generated outputs may change only through their owning Make generators. |
| Non-goals | No historical migration rewrite, hand-edited generated artifact, unrelated module redesign, frontend redesign, public route-shape change, authorization relocation, transaction-lifecycle relocation, or AC043 redistribution. |
| Default posture | Preserve valuable observable behavior. Apply owner-adopted field-identity, no-narrative, exact-input, and active-Reporting corrections only in their named slices; keep structural movement separate. |
| Document class | NLSpec-style refactor execution tracker; not an adopted subsystem NLSpec or product-behavior owner |

The target exists. Its live inventory is 26 Go files plus an empty `.gitkeep`; all
27 filesystem entries are accounted for in Section 2. The package is a real Links
and Tags source-owner boundary, but its current public surface also exposes several
consumer-specific providers and supporting mechanisms. The existence of the
directory is not, by itself, the reason to retain every current seam.

The remediation plan was explicitly authorized on 2026-08-21. This tracker is
the controlling execution artifact. Owner specifications remain authoritative
for behavior, and each structural or corrective slice remains independently
reviewable and reversible.

### Execution checkpoint protocol

After every workstream, including a blocked or conditionally dropped branch, the
executor MUST update the affected Section 9 status, append exact Section 10
evidence, record changed files, commands, run roots, rollback posture, blockers,
and the next executable slice, then run `make lint-markdown`. The next workstream
MUST NOT begin until that checkpoint passes. A failed checkpoint leaves the
current workstream `IN_PROGRESS` or `BLOCKED`; it is not completion evidence.

The execution baseline is commit
`22d33ee5f197a00e52789625326226f82d7a04a3` on `main`, four commits ahead of
`origin/main`. At activation, this tracker was the sole staged addition; it is
user-owned planning material and MUST be preserved while execution evidence is
appended.

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

## 2. Current-State Repository Inventory

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

### Future refactor Definition of Done

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
