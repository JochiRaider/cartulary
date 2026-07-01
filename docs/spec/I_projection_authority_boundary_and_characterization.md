# Appendix I: Projection Authority, Boundary, and Characterization Evidence

Appendix I is non-normative. It records current evidence for the projection authority, ownership, descriptor, restore-rebuild, query-characterization, and import-boundary decisions owned by Core 00, Core 01, and Core 04.

## I.1 Authority Posture

`docs/graph_projection_nlspec.md` currently declares `status: adopted/current` and is an adopted projections-specific subsystem NLSpec for graph-oriented projection behavior only. It does not govern workbook-grid projection tables, `view_row_v1`, workbook query routes, saved views, import owner facades, provider descriptors for workbook projections, or restore projection rebuild behavior. Research reports R01 through R09 remain informative unless a finding is explicitly promoted into Core, ADR, SPEC, or adopted NLSpec material.

Core 00 REQ-00-062 is the authority rule for projections. Adoption or substantive revision of a projections-specific NLSpec requires projection-related Core sections, implementation trackers, provider descriptors, rebuild behavior, query behavior, and boundary guard tests to be re-audited before accepting new projection changes.

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
| Timeline, hosts, identities, indicators | Owner-specific projection/query paths outside generic query surfaces. | Route dispatch, row shape, projection refresh, rebuild behavior. | Timeline/entity/indicator integration tests. | Provider descriptors may list ownership; generic query split must not claim these surfaces without characterization. |

## I.4 Restore Rebuild Characterization

Recovery owns restore orchestration. Projection modules own projection rebuild mechanics. The current implementation path may delegate to `projections.RebuildRestoreProjections` while the recovery-owned `RestoreProjectionRebuilder` adapter contract is introduced.

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
| `owner_module` | Source-owner module or subsystem. | Required. |
| `view_schema_ids` | View schemas served by provider. | Required when provider participates in query behavior. |
| `projection_table_ids` | Projection tables or derived stores owned by provider. | Required when provider owns persisted projection state. |
| `source_authorities` | Authoritative source records used to build projection rows. | Required. |
| `capabilities` | Explicit capability map using exactly `query`, `refresh_row`, `incident_rebuild`, and `restore_rebuild`. | Missing capabilities are invalid in code-backed descriptors and manifests. |
| `restore_rebuild` | Restore rebuild participation. | Must be `required`, `nonparticipating`, or `unsupported`. |
| `status` | Provider status. | Must be `active`, `deprecated`, or `experimental`. |
| `facade_packages` | Approved facade package boundary. | Package-level production import allowlist; test imports remain separate. |

`make json-shape-check` validates the manifest shape. `internal/modules/projections/provider_manifest_test.go` compares the manifest to the code-backed registry and `SupportsQuerySurface`.

## I.6 Import Graph Characterization

S-04 production import guardrails are package-import based and distinguish production imports from test-only imports. The current owner-approved production allowlist for direct `internal/modules/projections` imports is:

| Importer |
| -------- |
| `internal/app/operator.go` |
| `internal/modules/artifacts/import_projection.go` |
| `internal/modules/artifacts/linkednotes/facade.go` |
| `internal/modules/assessments/store.go` |
| `internal/modules/entities/store.go` |
| `internal/modules/evidence/import_projection.go` |
| `internal/modules/evidence/store.go` |
| `internal/modules/incidentbundles/source.go` |
| `internal/modules/parties/store.go` |
| `internal/modules/revisions/delete_restore_store.go` |
| `internal/modules/revisions/rollback_store.go` |
| `internal/modules/tasksdecisions/import_projection.go` |
| `internal/modules/tasksdecisions/supersede_facade.go` |
| `internal/modules/timeline/ports.go` |
| `internal/modules/workbook/mutation_store.go` |
| `internal/modules/workbook/store.go` |

Test-only imports are intentionally not production permissions. Production imports of projection internals, projection provider internals, rebuild internals, and projection test fixtures remain forbidden outside the approved facades/adapters/contracts.

## I.7 Boundary Guard Test Guide

The S-04 guard test lives at `internal/modules/projections/boundary_guard_test.go`. It parses Go imports with `go/parser`, ignores `_test.go` files, and fails on unapproved production imports of `github.com/JochiRaider/cartulary/internal/modules/projections`.

To add a new production facade, first update the owner-approved list in this appendix and the validation-only manifest if provider ownership is affected. Then update the guard allowlist with the package-level facade path. Do not add permissions for incidental file structure, test fixtures, provider internals, or temporary migration helpers.
