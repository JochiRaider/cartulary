# Appendix I: Projection Authority, Boundary, and Characterization Evidence

Appendix I is non-normative. It records current evidence for the projection authority, ownership, descriptor, restore-rebuild, query-characterization, and import-boundary decisions owned by Core 00, Core 01, and Core 04.

## I.1 Authority Posture

`docs/graph_projection_nlspec.md` currently declares `status: adopted/current` and is an adopted projections-specific subsystem NLSpec for graph-oriented projection behavior only. It does not govern workbook-grid projection tables, `view_row_v1`, workbook query routes, saved views, import owner facades, provider descriptors for workbook projections, or restore projection rebuild behavior. Research reports R01 through R09 remain informative unless a finding is explicitly promoted into Core, ADR, SPEC, or adopted NLSpec material.

Core 00 REQ-00-062 is the authority rule for projections. Adoption or substantive revision of a projections-specific NLSpec requires projection-related Core sections, implementation trackers, provider descriptors, rebuild behavior, query behavior, and boundary guard tests to be re-audited before accepting new projection changes.

Accepted workbook projection boundary: timeline owns authoritative timeline source semantics and the timeline workbook projection row DTO exposed by `internal/modules/timeline/workbookprojection`; projections owns the physical `timeline_grid_projection` storage lifecycle, delete/upsert behavior, incident rebuild orchestration, and restore rebuild orchestration. Core 01 and Core 03 own observable workbook query and startup behavior. Graph Projection NLSpec rules remain out of scope for workbook-grid projection tables, workbook query routes, saved views, restore rebuilds, and `view_row_v1`.

## I.2 R01 Through R09 Evidence Crosswalk

| Report | Evidence posture | Affected decision | Adopted into Core text? | Notes |
| ------ | ---------------- | ----------------- | ----------------------- | ----- |
| R01 | TODO: evidence report path required before citation. | Q-001 | No | Informative unless promoted by adopted-document process. |
| R02 | TODO: evidence report path required before citation. | Q-002 | No | Do not use as authority for public query behavior. |
| R03 | TODO: evidence report path required before citation. | Q-002 | No | Query characterization remains required before provider split. |
| R04 | TODO: evidence report path required before citation. | Q-003 | No | Recovery adapter ownership is now Core-owned. |
| R05 | TODO: evidence report path required before citation. | Q-003 | No | Restore characterization evidence only. |
| R06 | TODO: evidence report path required before citation. | Q-004 | No | Descriptor manifest remains validation-only. |
| R07 | TODO: evidence report path required before citation. | Q-004 | No | Code-backed registry remains authoritative. |
| R08 | TODO: evidence report path required before citation. | Q-005 | No | Import-graph evidence only. |
| R09 | TODO: evidence report path required before citation. | Q-005 | No | Boundary guard allowlist requires owner approval before expansion. |

## I.3 Query Characterization Matrix

Public query behavior is owned by Core 01 §3.3.4 and §3.3.4.1. `internal/modules/projections/query.go` is the current route-facing implementation facade for generic projection-backed query behavior, but it is not a permanent normative source of truth.

| Surface family | Current query path | Behavior to characterize before movement | Current evidence | Required parity posture |
| -------------- | ------------------ | ---------------------------------------- | ---------------- | ----------------------- |
| Assessment rows | `internal/modules/projections/query.go` via workbook query dispatch. | Field-key validation, null cells, collection cells, sort/filter/group, pagination, row envelope. | `internal/modules/projections/query_test.go`; provider registry tests. | Store-backed tests must fail on `view_row_v1` or query-contract drift. |
| Artifact-backed rows | `internal/modules/projections/query.go` for notes, communications log, handoff, status review, lesson, findings, investigative queries, and forensic keywords. | Artifact subtype filtering, collection cell shape, default sort, unsupported field behavior, saved-view validation. | `internal/modules/projections/query_test.go`; workbook phase9 coordination tests. | Split only behind a stable provider interface with parity tests. |
| Evidence rows | `internal/modules/projections/query.go` through generic evidence surface. | Attachment-state cells, null cells, filter/sort semantics, row version, paging. | `internal/modules/projections/query_test.go`; evidence integration tests. | Preserve route-owned validation and `internal_error` behavior for unexpected provider failures. |
| Parties rows | `internal/modules/projections/query.go` through generic party surface. | Party text cells, scope/authorization envelope, grouping, row refresh shape. | `internal/modules/projections/query_test.go`; workbook parties integration tests. | Characterize before moving owner-specific logic. |
| Task and decision rows | `internal/modules/projections/query.go` through generic task/decision surfaces. | Queue fields, supersession cells, collection fields, filters, sort order, row snapshots. | `internal/modules/projections/query_test.go`; task/decision store tests. | Preserve revision/change-set row snapshots. |
| Timeline, hosts, identities, indicators | Owner-specific projection/query paths outside generic query surfaces; timeline source extraction is behind `internal/modules/timeline/workbookprojection`. | Route dispatch, row shape, projection refresh, rebuild behavior, and source/storage ownership. | Timeline/entity/indicator integration tests; projection provider manifest parity tests. | Provider descriptors may list ownership; generic query split must not claim these surfaces without characterization. |

## I.4 Restore Rebuild Characterization

Recovery owns restore orchestration. Projection modules own projection rebuild mechanics. The current implementation path delegates through a narrow projections restore rebuilder adapter rather than using the full projection store as the recovery-facing interface.

| Restore condition | Characterized default | Evidence to preserve |
| ----------------- | --------------------- | -------------------- |
| Request shape | `restore_projection_rebuild_request_v1` carries `restore_operation_id`, non-empty `restored_source_state_ref`, `rebuild_scope='all_active_providers'`, provider registry reference or snapshot, and caller context. | Recovery tests must prove the adapter receives a non-empty restore operation identifier and source-state reference. |
| Result shape | `restore_projection_rebuild_result_v1` carries `status`, `readiness_outcome`, ordered `provider_results[]`, warnings, and errors. | Restore tests must assert readiness from structured result state rather than from a nil error alone. |
| No active projection providers | Projection readiness may be `not_applicable`. | Recovery tests must show readiness is explicit rather than implied by skipped work. |
| Active provider lacks rebuild support | Fail closed unless Core marks provider `nonparticipating`. | Provider descriptor validation must catch unsupported active providers. |
| Partial rebuild failure | Restore readiness remains `incomplete` or `degraded`. | Store-backed restore tests should assert readiness outcome and warnings/errors. |
| Retry | Rebuild is idempotent for the same restored source state and scope. | Rebuild tests should compare provider results and row counts when available. |
| Existing projection data | Derived state is replaced or reconciled deterministically, not silently merged with stale rows. | Restore/rebuild tests should run against preexisting projection rows. |
| Missing restored source-state reference | Rebuild fails before touching projection state. | Recovery adapter tests must assert fail-before-touch behavior. |

## I.5 Provider Descriptor Manifest Design

The current runtime authority is the code-backed registry in `internal/modules/projections/provider_registry.go`. `contracts/projection-providers/index.json` is a canonical validation artifact for drift detection and review only.

| Field | Purpose | Validation posture |
| ----- | ------- | ------------------ |
| `provider_id` | Stable unique provider key. | Must be unique among active providers. |
| `schema_version` | Descriptor shape version. | Unknown versions fail validation. |
| `source_owner_module` | Source-owner module or subsystem that owns projection source semantics and query descriptor intent. | Required. |
| `projection_storage_owner_module` | Module or subsystem that owns physical derived projection storage lifecycle. | Required. |
| `view_schema_ids` | View schemas served by provider. | Required when provider participates in query behavior. |
| `projection_table_ids` | Projection tables or derived stores owned by provider. | Required when provider owns persisted projection state. |
| `source_authorities` | Authoritative source records used to build projection rows. | Required. |
| `capabilities` | Explicit capability map using exactly `query`, `refresh_row`, `incident_rebuild`, and `restore_rebuild`. | Missing capabilities are invalid in code-backed descriptors and manifests. |
| `restore_rebuild` | Restore rebuild participation. | Must be `required`, `nonparticipating`, or `unsupported`. |
| `status` | Provider status. | Must be `active`, `deprecated`, or `experimental`. |
| `facade_packages` | Source-owner facade package boundary declared by each provider. | Package-level owner evidence; test imports remain separate. Timeline declares `internal/modules/timeline/workbookprojection` so source extraction stays timeline-owned while projection storage writes stay in `projections`. |
| `import_policy` | Validation-manifest import policy for projection root, adapter, and contract packages. | Root production importer list is empty; adapter and contract packages are approved by exact package path. |

`make json-shape-check` validates the manifest shape. `internal/modules/projections/provider_manifest_test.go` compares the manifest to the code-backed registry and `SupportsQuerySurface`.

## I.6 Import Graph Characterization

S-04 production import guardrails are package-import based and distinguish production imports from test-only imports. The current validation-manifest import policy is:

| Policy member | Approved values |
| ------------- | --------------- |
| `approved_root_importers` | Empty. Production code outside `internal/modules/projections/**` must not import root `internal/modules/projections`. |
| `approved_adapter_packages` | `internal/modules/projections/adapters` |
| `approved_contract_packages` | `internal/modules/projections/providercontract` |

Test-only imports are intentionally not production permissions. Production imports of projection internals, projection provider internals, rebuild internals, and projection test fixtures remain forbidden outside approved adapters/contracts. Exact production imports of the stable projection provider contract package are allowed so source-owner providers can publish descriptors without importing the projection runtime.

## I.7 Boundary Guard Test Guide

The S-04 guard test lives at `internal/modules/projections/boundary_guard_test.go`. It parses Go imports with `go/parser`, ignores `_test.go` files, fails on every production root import of `github.com/JochiRaider/cartulary/internal/modules/projections`, allows the exact adapter and contract packages listed in the validation manifest, and fails on other projections subpackages.

To add a new production adapter or contract package, first update the owner-approved list in this appendix and the validation-only manifest if provider ownership is affected. Then update the guard policy with the exact package path. Do not add permissions for incidental file structure, test fixtures, provider internals, root imports, or temporary migration helpers.
