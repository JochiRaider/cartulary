# Appendix I: Projection Authority, Boundary, and Characterization Evidence

Appendix I is non-normative. It records current evidence for the projection authority, ownership, descriptor, restore-rebuild, query-characterization, and import-boundary decisions owned by Core 00, Core 01, and Core 04.

Behavioral ownership lives in the applicable normative Core sections and
adopted subsystem NLSpecs. Versioned machine projections and verification
routing live under `contracts/` and `tools/test_families`; they implement and
exercise those owners without replacing them. This appendix explains the
decisions for people and is not consumed by product, conformance, generation,
or release checks.

## I.1 Authority Posture

`docs/graph_projection_nlspec.md` currently declares `status: adopted/current` and is an adopted projections-specific subsystem NLSpec for graph-oriented projection behavior only. It does not govern workbook-grid projection tables, `view_row_v1`, workbook query routes, saved views, import owner facades, provider descriptors for workbook projections, or restore projection rebuild behavior. Research reports R01 through R09 remain informative unless a finding is explicitly promoted into Core, ADR, SPEC, or adopted NLSpec material.

Core 00 REQ-00-062 is the authority rule for projections. Adoption or substantive revision of a projections-specific NLSpec requires projection-related Core sections, implementation trackers, provider descriptors, rebuild behavior, query behavior, and boundary guard tests to be re-audited before accepting new projection changes.

Accepted workbook projection boundary: each named source owner owns authoritative
source semantics, canonical source snapshots, typed projection inputs, semantic
query intent, and Reporting fact meaning. Projections owns every production SQL
operation against the ten descriptor-owned `*_grid_projection` tables, private
query compilation and execution, keyset paging, incident rebuild mechanics, and
restore rebuild mechanics. Source enumeration remains source-owner supplied and
deterministic; projection storage never becomes source authority. Core 01 and
Core 03 explain observable workbook query and startup behavior. Graph Projection
NLSpec rules remain out of scope for workbook-grid projection tables, workbook
query routes, saved views, restore rebuilds, and `view_row_v1`.

## I.2 R01 Through R09 Evidence Crosswalk

| Report | Evidence posture | Affected decision | Adopted into Core text? | Notes |
| ------ | ---------------- | ----------------- | ----------------------- | ----- |
| R01 | TODO: evidence report path required before citation. | Q-001 | No | Informative unless promoted by adopted-document process. |
| R02 | TODO: evidence report path required before citation. | Q-002 | No | Do not use as authority for public query behavior. |
| R03 | TODO: evidence report path required before citation. | Q-002 | No | Query characterization remains required before provider split. |
| R04 | TODO: evidence report path required before citation. | Q-003 | No | Recovery adapter ownership is now Core-owned. |
| R05 | TODO: evidence report path required before citation. | Q-003 | No | Restore characterization evidence only. |
| R06 | TODO: evidence report path required before citation. | Q-004 | No | Descriptor manifest remains validation-only. |
| R07 | TODO: evidence report path required before citation. | Q-004 | No | The application-composed, code-backed descriptor registry remains authoritative. |
| R08 | TODO: evidence report path required before citation. | Q-005 | No | Import-graph evidence only. |
| R09 | TODO: evidence report path required before citation. | Q-005 | No | Boundary guard allowlist requires owner approval before expansion. |

## I.3 Query Characterization Matrix

Public query behavior is owned by Core 01 §3.3.4 and §3.3.4.1. Source-owner
`workbookprojection` packages contribute immutable semantic `SurfaceIntent`
values. Private `internal/modules/projections/internal/queryengine` plans bind
those intents to physical tables, expressions, joins, scanning, and row
materialization; `internal/modules/projections/internal/runtime` coordinates the
validated plans. Neither package is a normative source of truth. The production
Workbook port supplies a validated `querypage.Window`, each provider applies a
normalized keyset predicate plus `LIMIT limit+1`, and Workbook alone encodes the
last emitted row into the existing opaque cursor token.

| Surface family | Current query path | Characterized behavior | Current evidence | Required parity posture |
| -------------- | ------------------ | ---------------------- | ---------------- | ----------------------- |
| Assessment rows | Assessment `SurfaceIntent` to private compiled plan and bound Workbook query port. | Field-key validation, null cells, collection cells, sort/filter/group, pagination, row envelope. | Private runtime query tests, compiled-plan parity, and provider-registry tests. | Store-backed tests fail on `view_row_v1` or query-contract drift. |
| Artifact-backed rows | Artifact intents to private plans for notes, communications log, handoff, status review, lesson, findings, investigative queries, and forensic keywords. | Artifact subtype filtering, collection cell shape, default sort, unsupported field behavior, saved-view validation. | Private runtime query tests and Workbook coordination-surface tests. | Every artifact discriminator remains private and exactly matches its semantic intent and view schema. |
| Evidence rows | Evidence intent to private compiled plan and bound Workbook query port. | Attachment-state cells, null cells, filter/sort semantics, row version, paging. | Private runtime query tests and Evidence integration tests. | Route-owned validation and `internal_error` behavior remain unchanged for unexpected provider failures. |
| Parties rows | Party intent to private compiled plan and bound Workbook query port. | Party text cells, scope/authorization envelope, grouping, row refresh shape. | Private runtime query tests and Workbook party integration tests. | Party source meaning remains Party-owned while table access remains Projections-owned. |
| Task and decision rows | Separate task-request and decision intents to separate private plans and one typed source-owner facade. | Queue fields, supersession cells, collection fields, filters, sort order, row snapshots. | Private runtime query tests and Tasks/Decisions store tests. | Revision and change-set row snapshots remain stable. |
| Timeline and indicators | Owner-specific semantic intents to private compiled plans and bound Workbook query ports. | Route dispatch, row shape, bounded keyset retrieval, projection refresh, rebuild behavior, and source/storage ownership. | Timeline and Indicator integration tests, compiled-plan equality, and manifest parity tests. | Each provider implements the same neutral page window and nulls-last keyset semantics without moving source meaning into Projections. |
| Hosts and identities | Typed Entities query readers call private Projections host/identity plans, then hydrate only the returned bounded identifiers through Entities. | Differential filter/sort/null-order behavior, `limit+1`, exact-ID hydration, complete rows, and continuation pages. | Private query-engine differential tests and Entities/Workbook integration tests. | Host and Identity remain query-capable descriptors without leaking through the generic Workbook adapter. |

Keyset continuation preserves the normalized sort tuple, default sort tail, final `record_id` tie-breaker, direction, and nulls-last behavior. Provider queries bind cursor values and the bounded limit as SQL parameters and do not use pagination `OFFSET`. Host and Identity alias/identifier hydration is restricted to the bounded page record identifiers rather than the entire incident.

## I.4 Restore Rebuild Characterization

Recovery owns restore orchestration. Projection modules own projection rebuild mechanics. The current implementation path delegates through a narrow projections restore rebuilder adapter rather than using the full projection store as the recovery-facing interface.

| Restore condition | Characterized default | Evidence to preserve |
| ----------------- | --------------------- | -------------------- |
| Request shape | `restore_projection_rebuild_request_v1` carries `restore_operation_id`, non-empty `restored_source_state_ref`, `rebuild_scope='all_active_providers'`, provider registry reference or snapshot, and caller context. | Recovery tests must prove the adapter receives a non-empty restore operation identifier and source-state reference. |
| Result shape | `restore_projection_rebuild_result_v1` carries `status`, `readiness_outcome`, ordered `provider_results[]`, warnings, and errors. Successful provider results identify every rebuilt view schema and carry descriptor-ordered projection-table row counts; rolled-back work reports no rebuilt resources. | Restore tests must assert readiness from structured result state rather than from a nil error alone and must reject pre-commit success claims. |
| No active projection providers | Projection readiness may be `not_applicable`. | Recovery tests must show readiness is explicit rather than implied by skipped work. |
| Active provider lacks rebuild support | Fail closed unless Core marks provider `nonparticipating`. | Provider descriptor validation must catch unsupported active providers. |
| Partial rebuild failure | Restore readiness remains `incomplete` or `degraded`. | Store-backed restore tests should assert readiness outcome and warnings/errors. |
| Retry | Rebuild is idempotent for the same restored source state and scope. | Rebuild tests should compare provider results and row counts when available. |
| Existing projection data | Derived state is replaced or reconciled deterministically, not silently merged with stale rows. | Restore/rebuild tests should run against preexisting projection rows. |
| Missing restored source-state reference | Rebuild fails before touching projection state. | Recovery adapter tests must assert fail-before-touch behavior. |

## I.5 Provider Descriptor Manifest Design

The current runtime authority is the code-backed descriptor registry assembled
by `internal/app/projectionassembly/build.go` from eight required typed owner
contributions. Runtime validation and executable provider coordination are
private to Projections. `contracts/projection-providers/index.json` is an
authored canonical validation artifact for drift detection and review only. It
is outside generated-artifact policy roots and has no generator. Descriptor
changes update the owner contribution and assembled code-backed registry first,
then the manifest in the same change; manifest shape versions change only when
the JSON shape changes.

| Field | Purpose | Validation posture |
| ----- | ------- | ------------------ |
| `provider_id` | Stable unique provider key. | Must be unique among active providers. |
| `schema_version` | Descriptor shape version. | Unknown versions fail validation. |
| `source_owner_module` | Source-owner module or subsystem that owns projection source semantics and query descriptor intent. | Required. |
| `projection_storage_owner_module` | Module or subsystem that owns physical derived projection storage lifecycle. | Required. |
| `view_schema_ids` | View schemas served by provider. | Required when provider participates in query behavior. |
| `projection_table_ids` | Projection tables or derived stores owned by provider. | Required when provider owns persisted projection state. |
| `source_record_types` | First-class record types materialized by the provider. | Required, unique, and distinct from owner-module dependency declarations. |
| `source_authority_modules` | Owner modules whose authoritative source state is read to derive, refresh, rebuild, or query rows. | Required, unique, limited to schema-ownership keys, and must include `source_owner_module`; projection-owned derived-table reads do not add `projections`. |
| `capabilities` | Explicit capability map using exactly `query`, `refresh_row`, `incident_rebuild`, and `restore_rebuild`. | Missing capabilities are invalid in code-backed descriptors and manifests. |
| `restore_rebuild` | Restore rebuild participation. | Must be `required`, `nonparticipating`, or `unsupported`. |
| `status` | Provider status. | Must be `active`, `deprecated`, or `experimental`. |
| `facade_packages` | Source-owner facade package boundary declared by each provider. | Package-level owner evidence; test imports remain separate. Timeline declares `internal/modules/timeline/workbookprojection` so source extraction stays timeline-owned while projection storage writes stay in `projections`. |
| `import_policy` | Validation-manifest import policy for projection root, adapter, and contract packages. | Root production importer list is empty; adapter and contract packages are approved by exact package path. |

`make json-shape-check` validates the manifest shape.
`internal/app/projectionassembly/catalog_manifest_test.go` compares the
manifest to the fully assembled immutable descriptor set and checks consumer
port completeness. `contracts/projection-providers/README.md` owns the short
registry-first maintenance procedure.

The provider-facing `SurfaceIntent` contract is semantic and immutable. It
contains stable view and field identities plus optional semantic source-filter
tokens; it contains no table name, join, expression, predicate, alias, scan
strategy, or executable callback. Private Projections plans own all physical
SQL and bind every runtime value as a parameter. Exact intent/plan/view-schema
equality and source-ownership tests constrain the compiled plans. A new
descriptor version is required if the serialized descriptor shape changes, but
an intent-value or validation change alone does not require a version bump.

## I.6 Import Graph Characterization

Production import guardrails are package-import based and distinguish
production imports from test-only imports. The current validation-manifest
import policy is:

| Policy member | Approved values |
| ------------- | --------------- |
| `approved_root_importers` | Empty. Production code outside `internal/modules/projections/**` must not import root `internal/modules/projections`. |
| `approved_adapter_packages` | `internal/modules/projections/adapters` |
| `approved_contract_packages` | `internal/modules/projections/providercontract` |

Test-only imports are intentionally not production permissions. Production
imports of Projections internals, source-owner projection-provider internals,
rebuild internals, and projection test fixtures remain forbidden outside exact
assembly allowances. Only `internal/app/projectionassembly` imports the adapter.
Source-owner `workbookprojection` facades may import the stable provider
contract so they can contribute descriptors and semantic intent without
importing Projections runtime code.

## I.7 Boundary Guard Test Guide

The package boundary guard lives at
`internal/modules/projections/adapters/boundary_guard_test.go`. It parses Go
imports with `go/parser`, ignores `_test.go` files and registered runtime-excluded
test-support roots, requires the Projections root directory to contain no Go
files, permits only `internal/app/projectionassembly` to import `adapters`,
permits the stable `providercontract`, and rejects every production import of a
private Projections package. It also enforces exact assembly allowances for
source-owner `projectionprovider` constructors.

To change the construction or contract boundary, first update the adopted ADR
and applicable Core owner if behavior or ownership changes. Then update the
validation-only manifest and guard policy with exact package paths. Do not add
permissions for incidental file structure, test fixtures, runtime internals,
root imports, or temporary migration helpers.
