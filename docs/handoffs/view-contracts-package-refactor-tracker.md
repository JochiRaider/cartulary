# view-contracts Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

This tracker is the controlling execution and handoff artifact for the authorized S-00 through S-08 remediation. It does not supersede an adopted owner; each completed slice records its authored inputs, generated projections, implementation, evidence, compatibility effect, and rollback posture here before the next slice begins.

| Item | Value |
| --- | --- |
| Target path | `packages/view-contracts` |
| Target label | `view-contracts` |
| Output path | `docs/handoffs/view-contracts-package-refactor-tracker.md` |
| Status | `DONE` 2026-08-04; S-00 through S-08 implemented, validated, and handed off |
| Allowed change in this session | Adopted owners, authored contracts, generator inputs, generated projections through Make, package and web implementation, tests, authored harness inputs, and this tracker |
| Repository baseline | Clean `main` worktree at `ea622cf7f09451d87fd2be955d9cea6959d4e0f6` before the tracker was created |
| Prior tracker history | None; the output path did not exist at session start |
| Implementation authority | Granted by the task owner on 2026-08-04 for the complete S-00 through S-08 remediation plan; generated files remain generator-owned and MUST NOT be hand-edited |
| Compatibility decision | The task owner confirmed that no supported previously shipped client rejects the current indicator-specific inspector values |

The source hierarchy used for this plan is:

1. Adopted subsystem NLSpecs for their named subsystem. No adopted subsystem NLSpec was found that directly owns this TypeScript package.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication. It was inspected and is not applicable to this refactor.
4. Domain vocabulary and implementation-support guides.
5. Current repository code, contracts, tests, and harness inputs.
6. The planning framework and prior handoffs as evidence only.

Owner and planning documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`
- `docs/research/nlspec-spec.md`, used as writing doctrine rather than behavioral authority
- `temp/analysis-notes.md`, used as non-authoritative decision input

Repository evidence inspected includes:

- Every tracked file under `packages/view-contracts`, plus its ignored dependency and compiler artifacts.
- The `@cartulary/protocol-ts/view-schemas` entry point, its generated source types, `contracts/view-schemas/index.json`, the indicator view-schema source, its JSON schema, and the contract generator input chain.
- The `platform.viewschema` and `module.workbook` authored OpenAPI sources and `contracts/ws/index.schema.json`.
- Representative web consumers for contract rows, surface registration, inspector state, saved views, import mapping, query patches, timeline collaboration, and generic/entity/assessment query flows.
- Backend route ownership in `internal/modules/viewschemas/routes.go`, `internal/modules/workbook/routes.go`, and collaboration protocol code.
- `tools/test_families/package.view_contracts.json`, its verification owner, related workbook/view-schema/protocol owner mappings, the test catalog, execution topology inputs, frontend import-boundary policy, and relevant Make target explanations.

The framework correctly anticipates a frontend contract-adapter role, but it is not proof of the live boundary. The execution baseline has 81 direct TypeScript importer files and four current harness rows. A historical phase-row comment in `src/index.ts` is not current runtime or harness authority.

Non-goals are implementing the selected owner repair locally in the package, moving domain workflows, editing generated files by hand, weakening closed-enum validation, or treating harness phase maps as runtime architecture. Valid owner-defined wire behavior is preserved. Intentional remediation includes strict malformed-source rejection, positive row-version enforcement, a distinct sparse-row type, and removal of public symbols that have no repository importer or continuing architectural value.

### 1.1 Normative language and local requirement IDs

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** in this tracker state conditions for the later authorized refactor. They do not amend Core 00 through Core 04. A contradiction between this tracker and an adopted owner MUST be resolved in the owner before implementation proceeds.

Identifiers beginning with `VCRT-` are tracker-local traceability aids. They MUST NOT be emitted into runtime contracts, generated artifacts, harness identity, or public APIs.

**VCRT-SRC-001**  
Implementation MUST derive behavior from adopted owners and their typed projections. It MUST NOT parse Markdown, this tracker, or research notes as runtime or conformance input.

**VCRT-SRC-002**  
The semantic decision for RB-001 is resolved: the indicator-specific values MUST be preserved. Delivery is complete; Core 01 and downstream evidence satisfy section 12.2.

**VCRT-COMPAT-001**  
Because no supported prior client rejects the current values, `cartulary.view.timeline.v2`, `cartulary.view.indicators.v1`, and `cartulary.inspector_config.v1` MUST retain their identities. Unknown inspector values MUST continue to fail closed. Implementations MUST NOT add fallback aliases, accept arbitrary strings, or introduce a protocol-version transition for this owner-corpus repair.

## 2. Execution-Baseline Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `packages/view-contracts/.gitkeep` | Historical directory marker | None | None | None | None | None | Package maintenance | Low | The package is no longer empty. Retain during planning; removal is not required for the refactor. |
| `packages/view-contracts/package.json` | Declares private ESM workspace package `@cartulary/view-contracts`, a single `.` source export, Vitest script, and dependency metadata | `.` maps to `src/index.ts` | Root TypeScript project and `apps/web` workspace dependency | Runtime: `@cartulary/protocol-ts`; development: Vitest and shared TypeScript tooling | Package test invocation and workspace build/typecheck | Does not generate artifacts | Frontend contract-adapter package | Medium | The public subpath surface must remain unchanged during a structural refactor. |
| `packages/view-contracts/src/index.ts` | Defines public view/row/inspector/surface types; parses and validates view-schema JSON; normalizes full rows and sparse patches; joins workbook-surface registry metadata; exposes schema IDs, status arrays, lookup helpers, sort resolution, field capabilities, and visible fields | `SortEntry`; view-field, inspector, surface, and row types; `normalizeViewRowV1`; `normalizeViewRowPatchV1`; `parseViewContractJSON`; surface build/list/get/require functions; 17 schema ID constants; surface-status arrays; view list/get/require functions; sort/capability/visibility helpers | 81 direct TypeScript importer files at execution bootstrap | `@cartulary/protocol-ts/view-schemas` generated source types, registry, and artifacts | `src/index.test.ts`; direct web and browser-E2E consumers; indirect workbook tests | Reads generated view-schema projections; owns no generated output | `package.view_contracts` as browser adapter, with contract meaning owned by Core 01 and interaction meaning owned by Core 03 | High | At 1,931 lines, this is a stable external facade with mixed internal responsibilities. It imports no backend, SQL, platform, grid-vendor, or test-util package. |
| `packages/view-contracts/src/index.test.ts` | Characterizes parsing, registry order, field metadata, row variants, inspector configuration, and invalid closed vocabulary values | No runtime exports; 22 Vitest cases | Package test family and a workbook collaborator test-family row keyed to an existing test title | Vitest and local public facade | Self-contained package unit tests | Exercises generated view-schema projections through the public facade | `package.view_contracts` test evidence | Medium | Coverage is meaningful but does not close all structural-refactor risks listed in section 4. Keep path and titles stable unless authored harness accounting is updated. |
| `packages/view-contracts/tsconfig.json` | Strict no-emit TypeScript project configuration | TypeScript project boundary | Root `tsconfig.json` project reference and Make-owned typecheck/build targets | Shared root TypeScript configuration | `make frontend-typecheck` | Produces no tracked output | Frontend tooling | Low | Preserve strictness and no-emit behavior. |
| `packages/view-contracts/node_modules/.bin/vitest` | Installed dependency shim | Out of scope | Tool invocation only | Package manager installation | Not source evidence | Tool-managed install artifact | Package manager | Low | Ignored and explicitly outside refactor scope. Do not edit or inventory as authored source. |
| `packages/view-contracts/tsconfig.tsbuildinfo` | Incremental compiler cache | Out of scope | TypeScript tooling only | TypeScript compiler | Not source evidence | Ignored build artifact | TypeScript tooling | Low | Do not edit, depend on, or treat as retained evidence. |

The generated contract chain is downstream of adopted owners. The completed refactor consumes those projections through `@cartulary/protocol-ts/view-schemas`; it did not hand-edit `packages/protocol-ts/src/generated/**`, `internal/gen/**`, or authored generated-artifact outputs.

## 3. Module Boundary Diagnosis

The live evidence supports retaining `@cartulary/view-contracts` as a thin browser-facing contract facade. That finding comes from its protocol-only runtime dependency, stable single-entry-point surface, and extensive browser consumers, not from the directory name. The current implementation is nevertheless a mixed-responsibility package internally: row-wire adaptation, view-contract parsing, inspector metadata validation, surface registry assembly, and field-capability helpers are concentrated in one source file.

The target is a view/projection-adjacent and transport-adjacent adapter because it interprets generated wire and view-schema projections. It does not materialize projections, execute transport, persist records, coordinate mutations or revisions, own frontend controller state, implement a grid vendor, or enforce authorization.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Public browser contract facade | `packages/view-contracts/src/index.ts` | `@cartulary/view-contracts` | Keep and simplify | One package export and 81 direct TypeScript importer files | Preserve the external entry point; remove only the unused raw parser and unused raw row type after the final importer audit. |
| Generated view-schema parsing and lookup | `src/index.ts` | View-contracts adapter over Core 01 projections | Split | Imports only the generated view-schema registry, artifacts, and source types | Separate implementation by semantic concern without changing module-load validation. |
| Full `view_row_v1` normalization | `src/index.ts` | Core 01 row-wire browser adapter | Split | Used by contract-row and timeline models | Preserve record identity, version, cell completeness, technical cells, and frozen output. |
| Sparse `view_row_patch_v1` normalization | `src/index.ts` | Core 01 row-wire browser adapter | Split | Used by generic, entity, assessment, and timeline live-patch consumers | Preserve sparse and additive-cell semantics; distinguish the concept internally without changing the public return type in the initial slice. |
| Inspector configuration metadata and closed vocabularies | `src/index.ts` | Core 01 emitted metadata; Core 03 interaction algorithms | Defer, then split | Core 02 REQ-02-258 delegates emitted registry content to Core 01; the semantic resolution is selected but Core 01 is not repaired | Preserve the current values verbatim until S-00 passes; S-06 MUST NOT begin earlier. |
| Workbook surface registry join, order, status, and constants | `src/index.ts` | View-contracts adapter, consumed by Workbook Interaction | Split | Joins generated artifact and registry projections for 17 surfaces | Preserve canonical order and required/optional omission behavior. |
| Sort, capability, and visible-field helpers | `src/index.ts` | View-contracts adapter | Split | Used by workbook grid/view models without importing a grid vendor | Keep vendor-neutral. |
| Frontend shell, inspector UI, and controller state | `apps/web/src/workbook/**` | Workbook Interaction frontend | Keep outside target | Representative consumers own registration, panels, and state | The target supplies metadata and types only. |
| Saved-view query/layout behavior | `apps/web/src/workbook/models/workbookSavedViews.ts` and saved-view owner modules | Saved Views and Workbook Interaction | Keep outside target | Target contributes `ViewContract` identity and fields only | Do not move persistence or startup behavior into the package. |
| Collaboration refresh and patch application | `apps/web/src/workbook/collaboration/**`, query hooks, and backend collaboration module | Collaboration plus owning workbook/query modules | Keep outside target | `record_changed` consumers call target normalizers; target does not own the stream | Preserve event admission semantics indirectly. |
| Import/tabular ingest workflow | `apps/web/src/workbook/features/ImportAssistantFeature.tsx` and imports owners | Imports plus source-record owners | Keep outside target | Target provides field metadata for mapping | No import execution belongs in view-contracts. |
| Projection refresh and storage | Backend projection, workbook, and source-owner modules | Projections and Search plus source owners | Keep outside target | No SQL, object store, or backend import exists in the package | The package validates returned browser rows only. |
| Authorization enforcement | Core 04 and backend route/source owners | Security and route owners | Keep outside target | Role and disabled-state metadata are presentation hints | Never elevate package metadata into an authorization decision. |
| Grid vendor integration | Web grid adapter | Grid adapter | Keep outside target | No vendor import exists in the package | The package may expose vendor-neutral capability metadata only. |

**VCRT-BND-001**  
`@cartulary/view-contracts` MUST remain a validating browser adapter over generated contract inputs. It MUST NOT become the semantic owner of view schemas, Inspector routes, authorization, persistence, projection refresh, mutation coordination, collaboration publication, saved views, imports, frontend shell state, or grid-vendor behavior.

**VCRT-BND-002**  
The package MUST retain the public `.` entry point and every used or strategically coherent exported name. It MUST add `NormalizedViewRowPatchV1`, make sparse and full normalized rows non-assignable at compile time, and remove `parseViewContractJSON` plus `ViewRowV1` after a repository-wide importer audit. It MUST NOT add a public subpath. Valid runtime result and freeze behavior remains stable; malformed source inputs and `row_version=0` intentionally fail closed.

**VCRT-BND-003**  
Internal source decomposition MAY vary from the filenames proposed in section 7 when that variation is not observable. Semantic responsibility, dependency direction, and slice acceptance criteria MUST remain unchanged.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `@cartulary/view-contracts` `.` export, exported names, object shapes, error strings, freezing, and module-load parsing | Package facade, semantics downstream of Core 01/Core 03 | `package.json` and `src/index.ts` | Package unit suite plus compile-time consumer coverage | Snapshot or explicit assertions for export availability, frozen results, and representative error classes/messages | High | No public subpath or symbol pruning in a behavior-preserving refactor. |
| View-schema discovery: `GET /api/v1/view-schemas` and `GET /api/v1/view-schemas/{view_schema_id}` | `platform.viewschema` and Core 01 | Authored OpenAPI and `internal/modules/viewschemas/routes.go` | Platform view-schema integration and contract rows | Preserve admission of every generated discovery artifact and registry/artifact identity agreement | High | The target is an indirect browser adapter, not a route handler. |
| Workbook query rows: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | `module.workbook` and source/projection owners | Authored OpenAPI, workbook routes, and web query adapters | Backend workbook tests and web query tests | Missing/invalid identity and version; complete cell set; unknown/additive fields; group values; technical cells; schema mismatch | High | Preserve the query envelope and full `view_row_v1` interpretation. |
| Row creation: `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | `module.workbook` plus source owner selected by view schema | Authored OpenAPI and workbook mutation adapters | Workbook integration and package grid-editable/create tests | Characterize returned row admission for all supported surface classes and replay responses | High | Target does not choose mutation owner or implement create semantics. |
| Row-refresh mutation envelopes from `PATCH /api/v1/records/{record_id}`, linked-note creation, same-field conflict resolution, and other workbook mutation routes | Revision coordinator, source owners, and `module.workbook` facade | `ViewMutationEnvelope` OpenAPI references and web mutation ports | Backend integration, revision, conflict, and web mutation tests | Preserve complete refresh-row parsing, record/version identity, change-set-adjacent fields, and rejection boundaries | High | No revision, conflict, or idempotency coordination moves into the target. |
| Sparse live row patches | Core 01 row wire; Collaboration transports owner-produced change data | `normalizeViewRowPatchV1` consumers in generic, entity, assessment, and timeline query paths | Package patch cases and web collaboration/query tests | Unknown field, technical cell, empty patch, invalid cell shape, view mismatch, and version edge cases | High | S-03 now returns branded `NormalizedViewRowPatchV1`; full and sparse results are compile-time distinct without runtime wire-member changes. |
| `GET /ws/v1/incidents/{incident_id}` and replayable `record_changed` event | Collaboration and source/revision owners | `contracts/ws/index.schema.json`, collaboration protocol, and web coordinator | Backend collaboration tests and web collaboration tests | Preserve affected-view selection, changed-field handling, sparse patch admission, refresh fallback, and replay behavior | High | The target owns neither authorization nor event emission. |
| Inspector configuration, panels, route bindings, seed bindings, conditions, and result behavior | Core 01 metadata and Core 03 interaction algorithms | Core 01 REQ-01-615 and section 7.4; Core 02 REQ-02-258; Core 03 REQ-03-247 and REQ-03-306; indicator schema | Package inspector tests and workbook inspector collaborator row | Exact mapping, generated-registry closure, strict parsing, and four-feature consumer evidence | Repaired | S-00 amended Core 01 and aligned every downstream projection/consumer; S-06 consumes the resulting registry without an independent map. |
| Workbook surface registry, 17 schema ID constants, status arrays, canonical order, required/optional surface handling | Core 01 surface registry projected through protocol-ts | `contracts/view-schemas/index.json` and surface registry code | Registry order/status tests | Required artifact omission; optional artifact omission; unknown artifact; duplicate or mismatched ID/path/status/kind; frozen output | High | There are 14 required surfaces and three standardized optional surfaces in the live registry. |
| Field keys, visibility, query metadata, grouping, filter operators, sort mapping, edit/create capability, enums, mutation metadata, and reference-pack metadata | Core 01 view contracts | Authored view schemas and package parser/helpers | Package field, filter, sort, group, capability, enum, mutation, and reference-pack tests | Cross-schema field-key identity and complete generated-contract closure | Medium | Preserve vendor-neutral behavior and exact field keys. |
| Saved-view identity, query, and layout interpretation | Saved Views owner and Workbook Interaction | `workbookSavedViews.ts` consumes `ViewContract` | Web saved-view and shell tests | Representative saved-view round trip against unchanged view and field IDs | Medium | Target affects identity compatibility but owns no saved-view storage or startup policy. |
| Import target field mapping | Imports owner and source owners | `ImportAssistantFeature.tsx` consumes view contracts | Web import-assistant tests and backend import tests | Representative writable/hidden/technical field mapping | Medium | No import preview, validation, or commit behavior belongs in this package. |
| Projection refresh and source-owner row semantics | Projections and Search plus each source owner | Backend query/mutation implementations; package has no backend dependency | Source-owner and workbook integration tests | Representative query and mutation refresh for affected surface families | Medium | Treat as consumer compatibility, not target architecture. |
| Role, route, and disabled-state hints | Core 01 metadata for presentation; Core 04 for enforcement | Inspector/config metadata and Core 04 REQ-04-150 | Package metadata tests and backend authorization tests | Ensure structural extraction preserves hints; rely on backend tests for actual authorization outcomes | High | Client metadata must never become authoritative access control. |
| Generated protocol/view-schema surface | Contract generator and `package.protocol_ts` | Generated source types and view-schema entry point | Protocol and generate-drift families | Every registry entry parses; generated source and artifact identities agree | High | Generated files are regenerated from owners and authored inputs, never hand-edited. |
| Test-util row/query/mutation/save-state/conflict/inspector behavior | Respective test-support and owning modules, not this package | No test-util import in production target; representative tests consume the facade | Package, web, workbook, collaboration, and owner integration tests | Preserve facade outcomes used by test support; do not move test fixtures into production | Medium | The absence of test-util production coupling is intentional. |
| Harness/test accounting | Verification owners and test-family catalog | Four active `package.view_contracts` rows and workbook collaborator mapping | `make explain-test-owner OWNER=package.view_contracts` | New or renamed test titles must be authored in owner inputs and regenerated through Make | Medium | Accounting is verification evidence, not runtime ownership. |

### 4.1 Required Indicator Inspector interface

**VCRT-INSP-001**  
The S-00 owner repair defines the following four feature objects exactly. Array order in `disabled_when[]` MUST match the order shown. `minimum_incident_role=null` means an authenticated caller with ordinary incident row-read access; it MUST NOT mean anonymous access.

| `feature_group_key` | Surface | `panel_id` | `route_binding.kind` | `route_binding.owner` | `minimum_incident_role` | `mutates` | `requires_confirmation` | `disabled_when[]` | Success | Failure |
| --- | --- | --- | --- | --- | --- | ---: | ---: | --- | --- | --- |
| `indicator.observations.manage` | Timeline | `relationships` | `indicator_observations` | `indicator_observations_route` | `editor` | `true` | `false` | `no_row_selected`, `incident_closed`, `authorization_lost`, `row_version_changed`, `record_deleted` | `preserve_selected_row` | `show_same_shell_error_invalidate_pending_action` |
| `indicator.observations.pivot` | Indicators | `relationships` | `indicator_observations` | `indicator_observations_route` | `null` | `false` | `false` | `no_row_selected`, `authorization_lost`, `record_deleted` | `preserve_selected_row` | `show_same_shell_error_preserve_selection` |
| `indicator.lifecycle.read` | Indicators | `history` | `indicator_lifecycle` | `indicator_lifecycle_route` | `null` | `false` | `false` | `no_row_selected`, `authorization_lost`, `record_deleted` | `preserve_selected_row` | `show_same_shell_error_preserve_selection` |
| `indicator.lifecycle.manage` | Indicators | `history` | `indicator_lifecycle` | `indicator_lifecycle_route` | `editor` | `true` | `false` | `no_row_selected`, `incident_closed`, `authorization_lost`, `row_version_changed`, `record_deleted` | `preserve_selected_row` | `show_same_shell_error_invalidate_pending_action` |

**VCRT-INSP-002**  
Each row in Table 4.1 MUST use `seed_bindings=[]`. Its `route_binding.action_key` MUST equal its `feature_group_key`. `route_binding.target_view_schema_id` MUST be omitted, not emitted as `null` or an empty string.

**VCRT-INSP-003**  
Core 01 REQ-01-615 MUST add `indicator_observations` and `indicator_lifecycle` to its closed kind vocabulary and `indicator_observations_route` and `indicator_lifecycle_route` to its closed owner vocabulary. The kind tokens identify semantic child-resource families. They MUST NOT identify HTTP methods, components, handlers, storage relations, or arbitrary route strings.

**VCRT-INSP-004**  
Core 01 section 7.4.1A MUST include exact rows for the four keys in Table 4.1 and MUST include `indicator.lifecycle.manage` in the `cartulary.view.indicators.v1` required feature list. The Timeline list MUST retain `indicator.observations.manage`.

**VCRT-INSP-005**  
The execution-baseline Timeline and Indicators schemas emitted the four keys and owner tokens but placed the four feature groups in `workflow`. S-00 moved observation features to `relationships` and lifecycle features to `history` together with the Core 01 amendment and aligned acceptance evidence.

### 4.2 Deterministic feature resolution and dispatch

**VCRT-DISPATCH-001**  
For one emitted feature group, resolution MUST execute in this order:

1. Match the complete `feature_group_key` against an exact Core 01 row.
2. If an exact row exists, use every routing, role, mutation, confirmation, disabled-state, seed, success, and failure member from that row.
3. Consult a wildcard family only when no exact row exists.
4. Validate that the resolved owner token has a registered narrow client adapter.
5. Omit the feature when no handler exists; do not render an inert control.
6. Dispatch through the registered semantic owner adapter and revalidate current selection state after completion.

**VCRT-DISPATCH-002**  
`indicator.observations.manage`, `indicator.observations.pivot`, `indicator.lifecycle.read`, and `indicator.lifecycle.manage` MUST be excluded from generic `*.manage` or other `record_patch` expansion. Observation features MUST use the Indicator observation child-route family. Lifecycle features MUST use `GET` or `POST /api/v1/indicators/{indicator_id}/state-intervals` as selected by the feature. Neither family may use `PATCH /api/v1/records/{record_id}`.

**VCRT-DISPATCH-003**  
The client MUST supply stable selected-record identity and current row version. Timeline observation management MUST additionally preserve the selected source field and exact byte-span context. Success MUST preserve selection by stable record ID rather than row position. Server routes MUST rederive membership, role, target visibility, lifecycle state, idempotency, and concurrency; disabled states are presentation hints only.

**VCRT-DISPATCH-004**  
No route-binding token, action key, grid coordinate, visible label, React component identity, backend handler name, SQL identity, or grid-vendor type may be converted into an arbitrary route string. No Indicator Inspector operation may introduce external enrichment or incident-data egress.

### 4.3 Compatibility and forward evolution

**VCRT-COMPAT-002**  
This repair is an owner-corpus correction, not a new wire feature: current authored view schemas, authored OpenAPI input, generated protocol projections, package validators, and package tests already contain the indicator-specific values, and the task owner confirmed that no supported prior client rejects them. No identity bump is permitted for this repair.

**VCRT-COMPAT-003**  
The two specialized kinds MUST remain limited to the existing Indicator child-resource families. They MUST NOT become a callback bus, plugin interface, authorization capability, external-enrichment mechanism, route-string escape hatch, or runtime registration precedent. A future generic child-resource binding requires a separately versioned owner contract.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| RB-001 owner-corpus contradiction has a selected resolution but is not delivered | Core 01 omits values required by Core 03 and already present downstream; the task owner selected preservation | Local package repair would still violate owner order | `must_fix` | Core 01 metadata owner, with Core 03/Core 04 cross-references | Amend Core 01 exactly as section 4 requires, align projections, and pass section 12.2 before S-06. |
| One 1,931-line public entry point combines five semantic concerns | Direct inspection of `src/index.ts` | Structural edits have broad blast radius | `should_fix` | View-contracts facade with internal concern modules | Sequence characterization before extraction and keep the public facade stable. |
| Per-view inspector feature-key facts are hardcoded in package code as well as owner projections | Package registry and generated view-schema metadata | Drift can be internally consistent yet owner-inconsistent | `should_fix` | Core 01 projection chain | After S-00, consume one owner-derived typed source and remove independently authoritative package registry facts without changing the public API. |
| Authored Indicator and Timeline feature objects place specialized observation and lifecycle features in `workflow` | Exact authored schema objects compared with the selected mapping in section 4.1 | UI grouping and same-surface behavior would remain owner-inconsistent | `should_fix` | Core 01 and authored view-schema inputs | Move observation features to `relationships` and lifecycle features to `history` only in the owner-aligned S-00 change. |
| Characterization does not cover several invalid-input and required/optional registry branches | Direct comparison of exported branches with 22 tests | A structural split could change rejection behavior unnoticed | `should_fix` | `package.view_contracts` tests | Add focused characterization before moving implementation. |
| Full and sparse normalized rows share `NormalizedViewRowV1` despite distinct invariants | Both normalizers return the same public type | Future callers may accidentally assume full-cell completeness | `should_fix` | Core 01 row adapter | Separate internal concepts first; defer any public type change. |
| Historical `unit.stage-0.row-02` comment remains in production code | Comment in `src/index.ts`; current owner has four rows | Historical phase vocabulary can be mistaken for current authority | `should_fix` | Package maintenance | Remove or rewrite only in a later structural slice; no behavior change. |
| Some public helpers/types have no discovered external production importer | Import aggregation compared with exports | Removal could break unsearched or future consumers | `defer` | Public package API owner | Do not prune without separate compatibility authority and complete usage evidence. |
| Tightening acceptance of malformed values beyond current tests is unspecified | Exported parser accepts JSON text and validates selected closed shapes | Validation changes are observable behavior | `defer` | Core 01 plus package adapter | Characterize current behavior; require later authorization for tightening. |
| Runtime dependency is limited to generated protocol contracts | `package.json` and imports | Low coupling and correct projection direction | `intentional/no_action` | View-contracts facade | Preserve the dependency direction. |
| No backend, platform, SQL, storage, test-util, or grid-vendor import exists | Direct target import scan | Moving such behavior here would erode boundaries | `intentional/no_action` | Existing backend, frontend, and adapter owners | Keep those responsibilities outside the package. |
| Saved views, imports, collaboration, projection refresh, and UI state remain in semantic owners | Representative caller inspection | Target could become a catch-all if consumer workflows move inward | `intentional/no_action` | Existing semantic owners | Retain only contract adaptation in this package. |
| Generated artifacts are consumed but not owned | Protocol entry point and generated-artifact policy | Hand edits would drift or be overwritten | `intentional/no_action` | Contract generator | Change owners/authored inputs, regenerate, and run drift checks when authorized. |
| Existing public identities remain compatible with all supported clients | Task-owner compatibility confirmation and current downstream acceptance | An unnecessary version fork would create incompatible duplicate identities | `intentional/no_action` | Core 01 compatibility owner | Retain both view-schema IDs and `cartulary.inspector_config.v1`; keep unknown-value rejection strict. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Establish authority order, safe scope, baseline, and handoff record | This tracker; owner and framework documents as evidence | `make lint-markdown` for the tracker | Execution authority and baseline are recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for every target file, public symbol family, caller class, dependency, and artifact | `packages/view-contracts/**` and representative consumers | Read-only scans; `make explain-test-owner OWNER=package.view_contracts` | Inventory is rebaselined at 81 direct importer files. |
| WF-02 | Contract-owner mapping and owner-corpus repair | parallel | WF-01 | WF-05 | Map every affected observable contract, then deliver the selected RB-001 repair through the owner chain | Core 01-04, Appendix E/F, authored OpenAPI/view schemas, generated projections | Section 8 owner/projection ladder | `DONE`; section 12.2 passes and RB-001 is closed. |
| WF-03 | Characterization test gap analysis | parallel | WF-01 | WF-05 | Identify tests required before code moves | `src/index.test.ts` and consumer tests | `make test-slice OWNER=package.view_contracts` | Existing coverage and exact missing branches are documented. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Distinguish legitimate facade behavior from misplaced workflows | Target imports plus representative web/backend owners | `make frontend-import-boundary-check` | No domain workflow, grid vendor, persistence, transport, or authorization implementation is assigned to the target. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Preserve one public facade while dividing implementation by semantic concern | Internal type, row, view, inspector, and surface modules | Typecheck and focused tests | `DONE`; explicit façade and acyclic semantic modules pass. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07 | Order the smallest reversible behavior-preserving moves | Target source and tests | Per-slice commands in section 7 | Each slice has dependencies, risks, rollback, and completion criteria. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Preserve or intentionally update authored test ownership without treating it as architecture | Authored test-family and verification-owner inputs only if tests move or are renamed | Test catalog/generation targets discovered through Make | Existing titles stay stable where practical; generated topology remains untouched by hand. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Run focused-to-broad evidence and leave a resumable handoff | Tracker plus authorized later changes | Section 8 validation ladder | Results, failures, skipped checks, and remaining blocker are recorded. |

Planning, owner repair, package decomposition, compatibility cleanup, authored accounting, validation, and handoff are complete. WF-00 through WF-08 are closed with slice evidence in section 13.

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | Authorization granted; `DONE` 2026-08-04 | Deliver the Core 01 vocabulary, exact rows, panel placement, precedence, typed Inspector registry, typed Indicator operations, narrow dispatch, and owner/browser evidence | Core 01, Core 03/Core 04 cross-references, Appendix E/F, authored view-schema/protocol inputs, generator/projection outputs, package and workbook adapters | Partial owner repair; panel or dispatch drift; hand-edited generated output; generic record patch fallback | Section 8 owner/projection, security/service, package, UI, and browser criteria | Focused S-00 ladder and standalone webserver-backed browser gate | Revert the owner/projection/application/UI/test slice as one compatibility unit and regenerate prior projections | Section 12.2 passes; RB-001 and VC-006 are `DONE`; S-06 is owner-unblocked. |
| S-01 | S-00; `DONE` 2026-08-04 | Added named characterization for all valid/rejection branches selected by the adopted schemas and approved cleanup decisions. Did not characterize malformed required documents or `row_version=0` as supported. | `packages/view-contracts/src/index.test.ts`; `tools/test_families/package.view_contracts.json`; generated topology render index | Tests could canonize accidental permissiveness | Invalid JSON/root/identity/version, cells, technical cells, additive members, sparse patches, registry/order, required/optional surfaces, freezing, deterministic errors | `make test-slice OWNER=package.view_contracts`; no web source changed, so no web owner slice was affected | Revert the S-01 test, authored manifest row, and regenerated topology projection together | All 11 new branches pass in their authored owner row; package owner now has 5 authored rows and 6 scheduled units including install. |
| S-02 | S-01; `DONE` 2026-08-04 | Extracted shared public types and invariant utilities behind the unchanged `.` facade. | `src/index.ts`, `src/types.ts`, `src/invariants.ts`; protocol view-schema literal type aliases | Export identity, error text, validation order, and cyclic imports | Characterization, package slice, repository typecheck, and formatting/lint preserve behavior | `make test-slice OWNER=package.view_contracts`; `make frontend-typecheck`; `make lint-biome` | Revert the two internal modules, restore definitions in `index.ts`, and remove only the added protocol type aliases | Public exports/runtime results are unchanged; dependency direction is protocol types → package types/invariants → façade with no internal cycle. |
| S-03 | S-02; `DONE` 2026-08-04 | Extracted row normalization, added branded `NormalizedViewRowPatchV1`, kept full and sparse runtime objects unchanged, and now requires positive safe `row_version >= 1`. | `src/index.ts`, `src/types.ts`, `src/rows.ts`, package row tests and authored family title | Sparse/full semantic drift; invalid concurrency admission; query/live patch regressions | Compile-time bidirectional non-assignability and full/sparse runtime matrix | Package and `web.workbook` slices, repository typecheck, Biome | Revert `rows.ts`, restore row functions/types in their S-02 locations, remove the one added title, and regenerate topology | Branded separation, positive versions, freezing, additive-member behavior, query/collaboration consumers, and 120/120 web owner units pass. |
| S-04 | S-01, S-02; `DONE` 2026-08-04 | Extracted view parsing/helpers and validates raw source documents with a generated runtime decoder before semantic invariants. | One-line façade, `src/view-contracts.ts`, invariant source diagnostics, protocol generator/decoder/entrypoint, generated validator | Diagnostic ordering, module-load timing, valid artifact order | Invalid JSON/root/required/type/unknown-member/closed-value tests and all 17 valid artifacts | Package/protocol slices, typecheck, drift, JSON shape, Biome | Revert parser module move, decoder entrypoint/generator work, source-diagnostic tests, and regenerated outputs without affecting `rows.ts` | Malformed source documents fail closed with deterministic source/path/reason errors; all 17 artifacts and valid helper behavior are unchanged. |
| S-05 | S-02, S-04; `DONE` 2026-08-04 | Extracted workbook-surface assembly, status arrays, schema constants, and required/optional joining while retaining `buildWorkbookSurfaceContracts` and injected sources. | `src/index.ts`, `src/view-contracts.ts`, `src/workbook-surfaces.ts`, package surface tests/accounting | Required/optional omission, ordering, artifact/registry mismatch | Required/optional branches, status partitions, all constants, canonical order, injected construction, kind/reference-pack drift | Package and 120-unit `web.workbook` slices, typecheck, Biome | Revert only the surface module, restore its block to `view-contracts.ts`, remove one test title, and regenerate topology | Registry order, identities, failures, omissions, constants, freeze boundaries, and production injection pass. |
| S-06 | S-00, S-02, S-04; `DONE` 2026-08-04 | Extracted Inspector parsing after owner/projection repair and consumes only owner-derived typed registry facts. The public façade and selected semantics remain unchanged. | `src/view-contracts.ts`, `src/inspector.ts`, exact package Inspector tests/accounting | Closed-vocabulary, route, lifecycle, authorization-hint, result, initialization, and freezing drift | Exact four-row mapping; mismatch/generic-patch rejection; strict unknown rejection; all 17 artifacts; representative Inspector consumers | Package, 120-unit web workbook, and 66-unit service-backed workbook slices; typecheck; Biome | Restore Inspector parsing to `view-contracts.ts`, remove `inspector.ts` and one title, then regenerate topology; retain S-00 owner repair | Section 4 interfaces pass, no independent registry remains, metadata stays non-authoritative, and package plus consumer evidence passes. |
| S-07 | S-03, S-04, S-05, S-06; `DONE` 2026-08-04 | Reduced `src/index.ts` to exact value/type re-exports, kept `.` and the coherent Inspector graph, removed public `parseViewContractJSON` and `ViewRowV1`, and added no alias or subpath. | `src/index.ts`, `src/types.ts`, internal parser test import, exact façade test | Accidental real-consumer removal or initialization-order change | Exact runtime export set, importer/subpath scans, internal module load, package consumers | Package slice, typecheck, import boundary, Biome | Restore the S-06 star exports and row type; no generated rollback is needed | One package entrypoint remains; parser is internal-only; removed type is absent; every workspace consumer compiles. |
| S-08 | S-07; `DONE` 2026-08-04 | Audited authored test accounting/topology, repaired two broad-gate omissions, reran the exact focused-to-broad ladder, and completed the implementation handoff. | Owner manifests, frontend source ownership, generator cleanup, generated topology through Make, tracker | Unaccounted evidence, generated drift, or incomplete handoff | Every added/renamed test is owner-routable; source ownership and Go staticcheck pass | Section 8.4 exact sequence; final `check` 715/715 | Revert S-08 ownership/helper fixes only if the corresponding S-00 files are also reverted; regenerate rather than edit outputs | All focused and broad checks pass; no unaccounted test/generated drift remains; overall tracker is `DONE`. |

### 7.1 S-00 ordered owner-corpus repair

S-00 MUST execute in this order and MUST stop at the first failed gate:

1. Amend Core 01 with the two kinds, two owner tokens, four exact feature rows, `indicator.lifecycle.manage`, exact-over-wildcard precedence, and the generic-record-patch exclusion.
2. Align Core 03 cross-references and Core 04 acceptance criteria without transferring metadata, interaction, source-state, or authorization ownership.
3. Update Appendix E historical status and Appendix F traceability; neither appendix may become an owner.
4. Align the authored Timeline and Indicators view schemas plus the platform view-schema OpenAPI owner input with section 4.
5. Regenerate every downstream projection through Make-owned generation. Generated roots MUST NOT be hand-edited.
6. Change `@cartulary/view-contracts` to validate owner-projected data without an independently authoritative per-view feature registry.
7. Bind workbook owner tokens to narrow Indicator adapters; React components and grid callbacks MUST NOT construct routes.
8. Run the complete section 8 ladder and record exact evidence before changing RB-001 delivery to `DONE`.

### 7.2 Slice independence rule

**VCRT-SLICE-001**  
S-01 through S-05 MAY proceed before S-00 only when they preserve every inspector value, feature object, error, freeze boundary, and initialization effect byte-for-byte or structurally equivalently. They MUST NOT reinterpret, remove, normalize, relocate between panels, or add aliases for inspector values.

**VCRT-SLICE-002**  
S-06 MUST NOT begin until S-00 and section 12.2 are complete. S-07 MUST NOT finalize the public facade while S-06 is blocked.

Every code-bearing slice MUST be independently reviewable and reversible. No behavior change is authorized unless the slice explicitly says `requires later authorization`.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | This tracker and repository Markdown policy | yes | Required after every slice and after the final implementation handoff. |
| unit | `make test-slice OWNER=package.view_contracts` | Four owned package rows | yes | Run before characterization and after each source extraction. |
| integration | `make service-backed-test-slice OWNER=module.workbook` | Workbook routes, row refresh, conflicts, inspector action integration | no | Required before completing a later slice that changes consumer-visible row or inspector behavior. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser/workbook behavior against a real server | no | Required when a later slice changes consumer-visible row, registry, inspector, or collaboration flow. |
| generated drift | `make generate-drift` | Contract and generated artifact projections | yes for owner/contract slices | A passing drift check proves projection consistency, not completeness against prose owners. |
| generated policy | `make generated-artifact-policy-check` | Generated-root ownership and edit policy | yes for owner/contract slices | Required after regeneration; generated roots MUST NOT be hand-edited. |
| JSON shape | `make json-shape-check` | Authored contract JSON shape | yes for owner/contract slices | Required after authored view-schema or OpenAPI input changes. |
| import-boundary/static | `make frontend-typecheck` | TypeScript package and consumers | yes | Baseline passed. |
| import-boundary/static | `make frontend-import-boundary-check` | Browser/runtime facade and package boundary policy | yes | Baseline passed. |
| import-boundary/static | `make lint-biome` | Authored frontend source style and static checks | no | Run for later TypeScript edits; not needed for the tracker-only write. |
| consumer unit | `make test-slice OWNER=platform.viewschema` | View-schema discovery owner | no | Required when discovery or generated schema adaptation changes. |
| consumer unit | `make test-slice OWNER=package.protocol_ts` | Generated TypeScript protocol entry points | no | Required when generator inputs or entry points change. |
| owner unit/service | `make test-slice OWNER=module.indicators` and `make service-backed-test-slice OWNER=module.indicators` | Indicator observations, lifecycle, storage, routes, and UI ownership | no | Required for S-00; live owner has 23 rows, 11 service-backed. |
| consumer unit | `make test-slice OWNER=web.workbook` | Workbook frontend consumers | no | Required after row, view, surface, or inspector extraction. |
| finalization | `make agent-finalize` | Retained run evidence and end-of-run maintenance | no | Passed in S-08 before broad verification; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| full fast check | `make test-fast` | Broad fast repository evidence | no | Run after all later structural slices. |
| full check | `make check` | Full repository check | no | Passed 715/715 in the final S-08 sequence. |

Commands were discovered through `make help`, `make task-guide ROLE=module-author OWNER=package.view_contracts`, `make explain-test-owner OWNER=package.view_contracts`, and `make explain-target` for the listed targets. Baseline evidence already run during planning:

| Command | Result | Retained run root |
| --- | --- | --- |
| `make test-slice OWNER=package.view_contracts` | Passed, 5/5 execution rows | `.cartulary/test-results/20260804T122125Z-p358192` |
| `make generate-drift` | Passed, 4/4 execution rows | `.cartulary/test-results/20260804T122133Z-p358556` |
| `make frontend-typecheck` | Passed, 2/2 execution rows | `.cartulary/test-results/20260804T122144Z-p361175` |
| `make frontend-import-boundary-check` | Passed, 2/2 execution rows | `.cartulary/test-results/20260804T122145Z-p361540` |
| `make lint-markdown` | Passed | `.cartulary/test-results/20260804T123720Z-p368339` |

No unrun command is claimed as successful.

### 8.1 Owner and projection acceptance

The following criteria are binary and MUST pass before S-00 is complete:

1. `indicator_observations` and `indicator_lifecycle` occur exactly once in every applicable closed kind projection.
2. `indicator_observations_route` and `indicator_lifecycle_route` occur exactly once in every applicable closed owner projection.
3. Unknown, misspelled, case-folded, padded, or extension-prefixed kind and owner values fail validation.
4. Every required surface emits exactly its Core 01 feature set; each implemented optional surface does the same.
5. Indicators emits `indicator.lifecycle.manage` exactly once.
6. Timeline resolves `indicator.observations.manage` to `indicator_observations_route`, never `record_patch_route`.
7. All four exact rows in section 4.1 override every wildcard family.
8. All 17 current view-schema artifacts parse and agree with the registry.
9. Authored JSON, generated source types, generated artifacts, OpenAPI discovery enums, and package runtime validation agree.
10. Generated-artifact policy and drift checks prove that regeneration came from authored inputs rather than hand edits.

### 8.2 Security and service acceptance

| Scenario | Required result |
| --- | --- |
| Authenticated incident viewer reads observation or lifecycle collection | Allowed when the selected record is visible and active. |
| Viewer attempts create, resolve, dismiss, restore, or lifecycle append | Denied without disclosing a hidden target. |
| Incident editor performs a declared mutation | Allowed only through the dedicated Indicator child route after current state validation. |
| `deployment_admin` without incident membership | No incident read or mutation access. |
| Cookie-authenticated mutation without valid CSRF | Rejected before mutation. |
| Hidden, foreign-incident, deleted, or wrong-type target | No existence disclosure; no child-resource side channel. |
| Exact idempotent replay | Returns the original successful result after current authorization. |
| Divergent replay or stale row version | Fails with no partial success. |
| Failed operation | Creates no history, projection refresh, idempotency-success record, or Collaboration publication. |
| Inspector dispatch | Calls the real observation or state-interval route; never generic record patch. |
| Missing client handler | Omits the feature instead of rendering an inert control. |
| Base-profile operation | Performs no third-party enrichment or external incident-data egress. |

### 8.3 Package and browser acceptance

1. The package accepts the four exact feature objects in section 4.1 and rejects unknown kind or owner tokens.
2. Parser results and normalized contract data remain recursively immutable to the same observable extent as the current package.
3. Saved views over the same `view_schema_id` inherit the same Inspector config and persist no Inspector UI state.
4. A pending lifecycle or observation form invalidates on selected-record, row-version, authorization, incident-lifecycle, deletion, or active-surface change.
5. Successful same-surface operations preserve the selected Indicator or Timeline row, scroll context, and focus target by stable identity.
6. Browser tests prove that no generic record-patch adapter is called for the four exact features.
7. Route dispatch accepts no grid-vendor type, cell coordinate, row index, visible label, or component identity.
8. Unknown or unsupported features remain fail-closed and follow `omit_feature` behavior.

### 8.4 S-08 final validation order

S-08 MUST run the following commands in order and MUST stop on the first failure:

```text
make lint-markdown
make lint-biome
make generate-drift
make generated-artifact-policy-check
make json-shape-check
make test-slice OWNER=platform.viewschema
make test-slice OWNER=package.protocol_ts
make test-slice OWNER=package.view_contracts
make test-slice OWNER=module.indicators
make service-backed-test-slice OWNER=module.indicators
make test-slice OWNER=web.workbook
make service-backed-test-slice OWNER=module.workbook
make frontend-typecheck
make frontend-import-boundary-check
make browser-e2e-webserver-backed
make agent-finalize
make test-fast
make check
git diff --check
git status --short
```

### 8.5 Current documentation revision evidence

Only documentation validation applies to this tracker revision.

| Command | Result | Evidence |
| --- | --- | --- |
| `make lint-markdown` | Passed | `.cartulary/test-results/20260804T132354Z-p385630` |
| `git diff --check` | Passed | No output |
| `git status --short` | Passed scope check | Only `?? docs/handoffs/view-contracts-package-refactor-tracker.md` |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| VC-001 | Establish source hierarchy, safe scope, baseline, and tracker | WF-00 | DONE | None | Section 1 | Tracker is the only touched file and implementation is explicitly unauthorized. |
| VC-002 | Inventory every target file, public surface, caller class, and dependency | WF-01 | DONE | VC-001 | Section 2 | All tracked and ignored target files are accounted for. |
| VC-003 | Map observable contracts to owners and select RB-001 semantics | WF-02 | DONE | VC-002 | Sections 4 and 11; task-owner compatibility confirmation | Owners, exact values, interfaces, and selected no-version-change posture are explicit. |
| VC-004 | Identify characterization gaps | WF-03 | DONE | VC-002 | Sections 4 and 7, S-01 | Exact pre-extraction test gaps are listed without inventing behavior. |
| VC-005 | Diagnose package boundary and coupling | WF-04 | DONE | VC-002 | Sections 3 and 5 | Facade responsibilities are separated from consumer-owned workflows. |
| VC-006 | Deliver the selected Core 01 owner-corpus repair | WF-02 | DONE | VC-003; implementation authority granted | RB-001, S-00, section 12.2, and the S-00 checkpoint | Core 01, authored inputs, generated projections, application dispatch, and conformance evidence agree. |
| VC-007 | Add behavior-preserving characterization tests | WF-03 | DONE | VC-004 | S-01 and checkpoint 13.2 | Required branches pass before implementation moves. |
| VC-008 | Extract types, invariants, and row normalization | WF-05/WF-06 | DONE | VC-007 | S-02 and S-03; checkpoints 13.3 and 13.4 | Public facade is stable; row runtime shape is unchanged; full and sparse results are type-distinct. |
| VC-009 | Extract view parsing and surface registry | WF-05/WF-06 | DONE | VC-007, VC-008 | S-04 and S-05; checkpoints 13.5 and 13.6 | All 17 schema artifacts, parser invariants, registry joins, constants, and consumer evidence pass. |
| VC-010 | Extract inspector metadata handling | WF-05/WF-06 | DONE | VC-006, VC-008, VC-009 | S-06 and checkpoint 13.7 | Owner conflict is closed; one owner-derived registry drives strict metadata parsing and exact specialized signatures. |
| VC-011 | Finalize stable facade and authored harness accounting | WF-07 | DONE | VC-008, VC-009, VC-010 | S-07 and S-08; checkpoints 13.8 and 13.9 | Explicit public façade and final catalog/topology/source-ownership accounting pass. |
| VC-012 | Run final implementation validation and hand off | WF-08 | DONE | VC-011 | Section 8 and checkpoint 13.9 | Required focused and broad evidence is recorded with no unreported failure. |
| VC-013 | Confirm supported-client compatibility posture | WF-02 | DONE | VC-003 | Task-owner confirmation recorded in section 1 | Existing schema and view IDs remain valid; no fallback alias or version transition is planned. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Framework doctrine and Core 00-05 authority mapped; planning-only boundary active | Inspected framework, domain vocabulary, Core 00-05, target package, contracts, consumers, and harness inputs; touched only this tracker | `git status --short`; representative `rg`, `sed`, and `jq` reads | Clean baseline confirmed; target exists; prior tracker absent | RB-001 | Obtain later authorization and resolve RB-001 before inspector semantic work. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Semantic decision resolved; adopted-owner delivery remains blocked | Inspected `nlspec-spec.md`, `analysis-notes.md`, live tracker, Core owner sections, authored contracts, package validators, generator validation, history, and compatibility evidence; touched only this tracker | `git status --short`; `sha256sum`; `rg`; `sed`; `jq`; `git log`; `git branch` | Tracker-local normative rules, exact mapping, compatibility posture, and acceptance gates added | RB-001 delivery gate | Seek separate authority for S-00; do not treat the tracker as an owner amendment. |
| 2026-08-04T09:53:44-04:00 | Codex remediation session | Execution authority granted; S-00 starting | Rebaselined tracker and live package importers; tracker changed from planning-only to active remediation | `git status --short`; `git rev-parse`; importer `rg`; `make task-guide ROLE=module-author OWNER=package.view_contracts` | Baseline remains `ea622cf7f09451d87fd2be955d9cea6959d4e0f6`; tracker is untracked; 81 direct importer files; focused owner command reconfirmed | None | Execute S-00 as one owner/projection/application compatibility unit. |
| 2026-08-04T11:00:27-04:00 | Codex remediation session | S-00 complete; RB-001 delivered | Core owners, typed registry and protocol inputs, generators, generated projections, package validation, workbook transport/dispatch/UI, focused tests, visual goldens, and owner-slice browser harness | S-00 checkpoint section 13.1 records exact commands, roots, failures, and visual refresh evidence | Owner/projection/application/browser agreement established; 24 Indicator owner rows, including two frontend rows and 11 service-backed rows | None | Run tracker close checks, then begin S-01 characterization. |
| 2026-08-04T11:10:45-04:00 | Codex remediation session | S-01 complete; characterization baseline frozen | Added 11 named package tests, registered one authored owner row, and regenerated topology through Make | `make test-slice OWNER=package.view_contracts`; `make generate`; tracker close checks in checkpoint 13.2 | Parser/row/patch/surface/export/freeze branches selected for preservation or intentional later cleanup now have explicit evidence; malformed required source tolerance and zero row versions were not blessed | None | Begin S-02 shared types and invariant extraction. |
| 2026-08-04T11:16:11-04:00 | Codex remediation session | S-02 complete; shared foundations extracted | Moved the public type graph to `types.ts`, moved owner-neutral invariant primitives to `invariants.ts`, retained explicit type re-exports from `.`, and exposed missing registry-derived protocol type aliases | Package slice, repository typecheck, Biome, and tracker close checks in checkpoint 13.3 | Runtime initialization, deterministic errors, public names, and all characterized outputs are unchanged; internal dependencies are one-way | None | Begin S-03 branded full/sparse row adapters. |
| 2026-08-04T11:21:13-04:00 | Codex remediation session | S-03 complete; row concepts separated | Extracted `rows.ts`, added compile-time-only full/patch brands, changed the sparse normalizer return type, rejected `row_version < 1`, expanded the runtime/type matrix, and regenerated topology | Package and complete `web.workbook` owner slices, repository typecheck, Biome, generation, and tracker close checks in checkpoint 13.4 | No runtime wire-member change; ReturnType consumers migrated automatically; invalid zero versions intentionally fail | None | Begin S-04 strict view-source decoding and parser/helper extraction. |
| 2026-08-04T11:31:17-04:00 | Codex remediation session | S-04 complete; malformed source tolerance removed | Generated an AJV standalone validator from the authored view-source schema, exposed a payload-safe typed decoder, applied it before semantic parsing, moved parser/lookups/helpers behind the façade, and added deterministic path/reason coverage | Package/protocol slices, typecheck, generator drift, JSON shape, Biome, and tracker close checks in checkpoint 13.5 | Valid artifacts and helper results remain stable; malformed documents now intentionally fail closed; decoder diagnostics identify required/unknown members exactly | None | Begin S-05 workbook surface assembly extraction. |
| 2026-08-04T11:37:34-04:00 | Codex remediation session | S-05 complete; surface assembly isolated | Moved registry joining, injected construction, lookup index, schema constants, and status partitions to `workbook-surfaces.ts`; expanded mismatch/constant evidence and regenerated topology | Package and complete `web.workbook` owner slices, typecheck, Biome, generation, and tracker close checks in checkpoint 13.6 | Required/optional behavior, order, identities, frozen outputs, and production injection remain stable | None | Begin S-06 Inspector module extraction. |
| 2026-08-04T11:46:55-04:00 | Codex remediation session | S-06 complete; Inspector metadata isolated | Moved Inspector parsing/cross-field validation to `inspector.ts`, retained only generic parser algorithms elsewhere, added exact four-signature and record-patch mismatch evidence, and regenerated topology | Package, complete web workbook, service-backed workbook, typecheck, Biome, generation, and tracker close checks in checkpoint 13.7 | No metadata, public type, initialization, authorization, transport, workflow, or UI behavior changed | None | Begin S-07 façade and compatibility cleanup. |
| 2026-08-04T11:51:57-04:00 | Codex remediation session | S-07 complete; intended package façade finalized | Replaced star exports with explicit value/type exports, made the raw parser internal-only, removed the unused raw row type, and pinned the exact runtime surface | Package slice, workspace typecheck, import-boundary gate, Biome, importer/subpath scans, and tracker close checks in checkpoint 13.8 | One `.` entrypoint remains; no importer migration, alias, subpath, initialization change, or valid runtime change was needed | None | Begin S-08 accounting audit and exact final validation sequence. |
| 2026-08-04T12:22:37-04:00 | Codex remediation session | S-08 and overall remediation complete | Audited six affected owners, repaired frontend source ownership and unused generator-helper omissions exposed by the first full check, regenerated through Make, reran the exact final sequence, and inventoried the final worktree | Checkpoint 13.9 records every final run root; `test-fast` 349/349; `check` 715/715; browser 62/62 | All planned owner, projection, package, UI, test, generated, and handoff outcomes are complete; no retained-run maintenance was performed because `RESULTS_DIR` was unset | None | Handoff complete; changes are ready for review and commit. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Target owns no backend route, projection, mutation, revision, persistence, or storage implementation | Inspected module workbook/view-schema OpenAPI, `internal/modules/workbook/routes.go`, `internal/modules/viewschemas/routes.go`, collaboration route/protocol evidence; touched only this tracker | `rg` route/operation scans; `jq` OpenAPI path and operation extraction | Route ownership remains in `platform.viewschema`, `module.workbook`, Collaboration, Revisions, and source owners | None outside RB-001 metadata ownership | Keep backend behavior in current owners; validate it only as consumer compatibility. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Observation and lifecycle route families remain adopted and outside the target package | Inspected Core 01 REQ-01-652/654, module ownership, OpenAPI operations, and Core 04 AC-532; touched only this tracker | Exact owner and route searches; `make task-guide ROLE=module-author OWNER=module.indicators`; `make explain-test-owner OWNER=module.indicators` | S-00 now requires dedicated child-route dispatch, no generic patch, and 23 owner rows including 11 service-backed rows | RB-001 delivery gate | Amend owners first, then align narrow adapters and service evidence in one authorized sequence. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Legitimate public facade with mixed internals; 82 direct TypeScript importers | Inspected all target source plus representative row, surface, inspector, saved-view, import, query, timeline, and collaboration consumers; touched only this tracker | `rg` import aggregation and exact-source `sed` reads | Keep package facade; split internals; keep controller/workflow/vendor state outside target | RB-001 blocks inspector semantics only | Add characterization, then execute S-02 through S-05 as isolated later slices. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Public facade remains frozen; exact Indicator dispatch and panel ownership are now specified | Inspected package types/registries/tests and authored Timeline/Indicators feature objects; touched only this tracker | `rg` exact token/key reads; `jq` exact feature-object extraction | Live authored objects use `workflow`; target mapping requires observation in `relationships` and lifecycle in `history` | RB-001 delivery gate | S-01 through S-05 may remain neutral; S-06 waits for owner-aligned panel and dispatch repair. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Package consumes 17 generated view-schema projections and owns no generated root | Inspected protocol view-schema entry point/types, registry, indicator source, OpenAPI, WebSocket schema, generator inputs, and generated-artifact policy; touched only this tracker | `jq` registry/schema reads; `make generate-drift` | Drift passed at `.cartulary/test-results/20260804T122133Z-p358556`; this does not resolve owner inconsistency | RB-001 | Reconcile owners, change authored inputs if authorized, then regenerate; never hand-edit generated files. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Downstream authored/generated/package inputs already contain the disputed tokens; Core 01 remains behind them | Inspected Core 01 vocabularies/registry, Core 03 REQ-03-306, OpenAPI enums, view-schema inputs, generator registry, and package validator; touched only this tracker | `rg`; `sed`; `jq`; repository history inspection | No identity bump is required by supported-client compatibility; exact owner repair and regeneration remain mandatory | RB-001 delivery gate | Apply S-00 in order and fail if any owner, authored input, generated projection, or runtime validator disagrees. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Four active package owner rows and one workbook collaborator dependency mapped; characterization gaps identified | Inspected target tests, package/workbook/view-schema/protocol test families, verification owners, catalog, topology, and Make target guides; touched only this tracker | `make task-guide ROLE=module-author OWNER=package.view_contracts`; `make explain-test-owner OWNER=package.view_contracts`; `make explain-target`; focused baseline targets | Package slice, typecheck, import boundary, generated drift, and Markdown lint passed | RB-001 affects inspector expected values | In a later authorized task, add S-01 tests before source extraction. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Owner/projection, security/service, package/browser, and binary closure matrices are explicit | Inspected current test owners and the analysis-note acceptance inventory; touched only this tracker | Owner discovery commands; `make lint-markdown`; `git diff --check`; `git status --short` | Future ladder is ordered and stops on first failure; documentation checks passed with only this tracker changed | RB-001 delivery gate | Add implementation test evidence only in a separately authorized S-00/S-01 task. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Target parses presentation metadata but performs no authoritative access check | Inspected Core 04 authorization requirements, inspector metadata, route security declarations, and relevant consumers; touched only this tracker | `rg`/`sed` owner and consumer reads; `jq` OpenAPI security reads | Server route/source owners remain authoritative; role and disabled hints are presentation only | RB-001 includes route-binding vocabulary, not authorization authority | Preserve hints exactly and require backend authorization evidence for any later action-flow change. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Security boundary is defined as an explicit acceptance matrix | Inspected Core 04 REQ-04-150/AC-532 and current feature roles/disabled conditions; touched only this tracker | `rg`; `sed`; `jq` | Viewer/editor, membership, CSRF, non-disclosure, replay, concurrency, atomicity, omission, and non-egress outcomes are binary | RB-001 delivery gate | Server tests must remain authoritative; client disabled conditions must never grant or deny access. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T08:26:07-04:00 | Codex planning session | Structural plan is decision-complete; implementation remains unstarted | Inspected all evidence listed in section 1; touched only this tracker | Read-only repository discovery and five passing validation targets, including Markdown lint | Non-inspector structural work is sequenced; no production refactor performed | RB-001 | First later session: authorize and perform S-00 or explicitly restrict work to S-01 through S-05 without inspector semantic changes. |
| 2026-08-04T09:18:36-04:00 | Codex NLSpec revision session | Product choice and no-version posture are complete; owner delivery and package refactor remain unstarted | Inspected all revision sources and live contradiction evidence; touched only this tracker | Read-only discovery; documentation validation and scope checks | RB-001 is no longer a design question; documentation checks pass; delivery remains a binary owner gate | RB-001 delivery gate | Authorize S-00 or restrict the next task to neutral S-01 through S-05 work. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | **Owner-corpus delivery gate. Decision status: RESOLVED. Delivery status: DONE.** Preserve `indicator_observations`, `indicator_lifecycle`, `indicator_observations_route`, `indicator_lifecycle_route`, and `indicator.lifecycle.manage`; use the exact mapping in section 4. | Partial repair could route a child-resource mutation through generic patching, publish inconsistent discovery, or make client presentation metadata disagree with authorization and service behavior. | Exact Core 01 amendment; aligned Core 03/Core 04 references; Appendix E/F maintenance; authored input alignment; generated projection refresh; narrow client adapters; every section 12.2 criterion passing | **DONE in S-00; rechecked through S-08.** Final focused, browser, fast, and full gates retain the resolved registry and expose no blocker. |

## 12. Binary Completion Criteria

### 12.1 Revised tracker completion

- [x] Every tracked target file and ignored target artifact is inventoried with an explicit posture.
- [x] The source hierarchy distinguishes adopted owners, typed projections, current implementation evidence, writing doctrine, and research input.
- [x] The RB-001 semantic decision, delivery status, compatibility posture, and ownership boundaries are unambiguous.
- [x] The four target feature objects define exact panels, roles, mutation flags, confirmation flags, disabled-state defaults, seed defaults, route bindings, and result behavior.
- [x] Exact-over-wildcard resolution and real-route dispatch are deterministic.
- [x] Public package compatibility, fail-closed unknown handling, and no-version-change behavior are explicit.
- [x] Every discovered public contract risk has an owner, test posture, and acceptance boundary.
- [x] Every workflow and slice has dependencies, validation, rollback, and a binary exit condition.
- [x] Owner/projection, security/service, and package/browser acceptance criteria are testable.
- [x] Generated files and generated harness outputs remain downstream projections and MUST NOT be hand-edited.
- [x] The framework/live mismatch, Core/downstream mismatch, and authored-panel mismatch are recorded.
- [x] Prior handoff history is preserved and the NLSpec revision session is appended.
- [x] Production implementation authority was granted on 2026-08-04 for S-00 through S-08.

### 12.2 S-00 owner-repair completion

RB-001 delivery remained blocked until every item below was reverified against the same resulting change; all S-00 items are now closed:

- [x] Core 01 lists `indicator_observations` and `indicator_lifecycle` exactly once in the closed kind vocabulary.
- [x] Core 01 lists `indicator_observations_route` and `indicator_lifecycle_route` exactly once in the closed owner vocabulary.
- [x] Core 01 contains the four exact feature rows from section 4.1.
- [x] The Indicators per-surface registry requires `indicator.lifecycle.manage` exactly once.
- [x] Exact feature rows precede and override wildcard families, and the four specialized keys cannot resolve to `record_patch`.
- [x] Core 03 references the same bindings without becoming a second wire-registry owner.
- [x] Core 02 source-state and Core 04 authorization/non-egress boundaries remain unchanged and correctly referenced.
- [x] Appendix E records the contradiction as closed historical material and Appendix F maps the repaired requirements to AC-454, AC-455, and AC-532.
- [x] Authored Timeline and Indicators schemas exactly match section 4.1, including panels, disabled-state order, empty seeds, action keys, and omitted target schema IDs.
- [x] Authored OpenAPI discovery enums and existing Indicator child routes match the adopted Core without aliases or generic-patch substitutes.
- [x] Generated OpenAPI, Go, protocol-ts, validator, registry, and other declared projections are regenerated and drift-clean.
- [x] Negative tests prove unknown tokens fail closed and all four features avoid generic record patching.
- [x] Security, authorization, CSRF, non-disclosure, idempotency, concurrency, atomicity, selection continuity, omission, and non-egress criteria pass.
- [x] `@cartulary/view-contracts` remains an adapter and no longer maintains an independently authoritative per-view feature registry.
- [x] The no-version-change posture remains valid and no fallback alias, permissive validation, or identity bump is introduced.
- [x] Separate implementation authority has been granted.
- [x] The focused S-00 section 8 ladder and standalone webserver-backed browser gate pass, VC-006 is `DONE`, and S-06 is owner-unblocked; final broad gates remain owned by S-08.

### 12.3 Package-refactor completion

- [x] S-01 characterizes all identified parser, row, patch, registry, freezing, and error boundaries before production source moves.
- [x] S-02 preserves public exports and completes independently with focused package and repository type evidence.
- [x] S-03 introduces the approved branded patch type and positive row versions without runtime wire drift.
- [x] S-04 closes malformed source documents and extracts parsing/helpers without valid-artifact drift.
- [x] S-05 preserves registry assembly, ordering, required failures, optional omission, constants, and production injection.
- [x] S-06 begins only after section 12.2 and consumes owner-derived Inspector projections.
- [x] S-07 preserves the `.` facade with no new public subpath, adds only the approved branded patch type, and removes only the approved dead parser/row seams after the importer audit.
- [x] Test titles and paths remain stable or authored harness inputs are updated and generated topology is regenerated.
- [x] Every required focused, service-backed, frontend, browser, finalization, fast, and full validation command passes.
- [x] Final handoff records exact changes, run roots, failures, skipped checks, rollback posture, and any remaining blocker.

Sections 12.1 through 12.3 pass, checkpoint 13.9 marks S-08 `DONE`, and the overall effort is complete.

## 13. Execution Checkpoints

### 13.1 S-00 — Owner, projection, and Indicator behavior repair

| Item | Result |
| --- | --- |
| Status | `DONE` on 2026-08-04; completed as one owner/projection/application compatibility unit |
| Dependencies | Implementation authority granted; baseline and compatibility posture recorded |
| Authored specification and contract changes | Core 01 exact vocabularies/rows/precedence/exclusion; Core 03/Core 04 references and AC-532; Appendices E/F; `view-inspector` registry family and JSON Schema; Timeline/Indicators feature objects; eight protocol operation catalog entries; protocol entrypoint/import-boundary ownership |
| Generated projections | Regenerated only through `make generate`: Go contract registry, view-schema/extension artifacts, protocol HTTP request/response/validator/binding types, Inspector registry TypeScript, and execution-topology render index |
| Implementation changes | `view-contracts` consumes the owner-derived registry; workbook Indicator operations use `WorkbookOperationExecutor`; exact complete-tuple dispatch mounts observations in `relationships` and lifecycle in `history`; paging, mutation metadata, retry/loading/empty states, stable-ID refresh, and unsupported-tuple omission are explicit |
| Test/accounting changes | Added exact transport/dispatch/paging tests and a dedicated accessible UI/paging/metadata row; `module.indicators` now has 24 rows, two frontend rows, and 11 service-backed rows; topology regenerated through Make |
| Compatibility/migration | Preserved `cartulary.view.timeline.v2`, `cartulary.view.indicators.v1`, `cartulary.inspector_config.v1`, HTTP paths, server authorization, and package `.`. No database/storage migration, identity bump, alias, permissive fallback, or generic patch compatibility was added. Panel placement and unsupported inert-control omission are intentional UI changes. |
| Rollback | Revert the complete S-00 authored owner/contract/generator/package/web/test/harness/golden change set together, restore the prior authored registry inputs, and rerun `make generate`; do not selectively retain generated outputs. |
| Next slice | S-01 characterization baseline; do not begin until the checkpoint Markdown and diff checks pass. |

#### S-00 passing evidence

| Command | Result | Run root or evidence |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260804T145847Z-p779607` |
| `make lint-markdown` | PASS after the completed checkpoint write | `.cartulary/test-results/20260804T150535Z-p800915` |
| `make lint-biome` | PASS | `.cartulary/test-results/20260804T150425Z-p796340` |
| `make generate-drift` | PASS | `.cartulary/test-results/20260804T150436Z-p796751` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260804T150446Z-p799320` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260804T150451Z-p799757` |
| `make test-slice OWNER=platform.viewschema` | PASS, 3/3 units | `.cartulary/test-results/20260804T144139Z-p621256` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 units | `.cartulary/test-results/20260804T144916Z-p663365` |
| `make test-slice OWNER=package.view_contracts` | PASS, 5/5 units | `.cartulary/test-results/20260804T144927Z-p668379` |
| `make test-slice OWNER=module.indicators` | PASS, 27/27 units | `.cartulary/test-results/20260804T145949Z-p789179` |
| `make service-backed-test-slice OWNER=module.indicators` | PASS, 13/13 units | `.cartulary/test-results/20260804T144956Z-p672848` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T145011Z-p674298` |
| `make service-backed-test-slice OWNER=module.workbook` | PASS, 66/66 units | `.cartulary/test-results/20260804T145049Z-p687286` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T150006Z-p792964` |
| `make frontend-import-boundary-check` | PASS, 2/2 units | `.cartulary/test-results/20260804T150501Z-p800288` |
| `make browser-e2e-webserver-backed` | PASS, 62/62 units | `.cartulary/test-results/20260804T145253Z-p723486` |
| `make explain-test-owner OWNER=module.indicators` | PASS | 24 rows; frontend 2; service-backed 11 |
| `make target-plan` | PASS | 614 units; digest `sha256:0fb8b2a04bf627a25e4b61b2c5dc7d20b28a1f7b794ebc9aa33788e7efc49be5` |
| `git diff --check` | PASS | No output after the S-00 tracker checkpoint |

#### S-00 failures encountered and resolved

| Run root | Failure and disposition |
| --- | --- |
| `.cartulary/test-results/20260804T140454Z-p404667` | Initial generation exposed an accidentally duplicated panel; authored schema corrected, then regeneration passed. |
| `.cartulary/test-results/20260804T141738Z-p412825` | Formatter exposed an implicit-any branch and forbidden test assertions; source/test typing corrected. |
| `.cartulary/test-results/20260804T141911Z-p425638`, `.cartulary/test-results/20260804T141935Z-p426262` | JSON-shape validation exposed missing active-family accounting and then stale generated topology; authored checker updated and generation rerun. |
| `.cartulary/test-results/20260804T142022Z-p435138` | Package tests exposed deterministic registry-order diagnostic changes; assertions aligned with the owner-derived ordered registry. |
| `.cartulary/test-results/20260804T142305Z-p455919`, `.cartulary/test-results/20260804T142611Z-p467140`, `.cartulary/test-results/20260804T142755Z-p475106` | Owner-scoped service validation exposed missing selection provenance and illegal direct Playwright row scheduling; the harness now forwards selection identity and projects Playwright rows through filtered browser groups. |
| `.cartulary/test-results/20260804T143206Z-p514340` | Functional browser rows passed; the intentional Indicator panel move produced reviewed visual deltas. Canonical refresh and later comparison validation passed. |
| `.cartulary/test-results/20260804T144323Z-p645189`, `.cartulary/test-results/20260804T144442Z-p649530`, `.cartulary/test-results/20260804T144756Z-p656203` | Type/import gates exposed an over-wide parsed-type field, an untyped spy, protocol imports outside adapters, a missing generated-entrypoint allowlist, and a panel narrowing gap; each was structurally corrected. |
| `.cartulary/test-results/20260804T145859Z-p781853` | The new UI row used unavailable DOM matcher extensions; assertions were rewritten against standard DOM properties and the owner slice passed. |

All failures were related to S-00 implementation or its required owner-scoped validation path and are resolved. No S-00 validation was skipped. Broad `agent-finalize`, `test-fast`, and `check` remain deliberately sequenced in S-08, not claimed here.

#### S-00 visual refresh record

- Accepted trigger: adopted Inspector behavior intentionally moved Indicator observation management from `workflow` to `relationships` and omitted unsupported inert relationship controls.
- Affected owner rows: `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` and `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`.
- Affected fixtures: `visual.fixture.base_inspector`, `visual.fixture.destructive_actions`, and `visual.fixture.mention_chip_state_matrix`.
- Changed goldens: `entity-mention-chip-states-linux.png`, `workbook-inspector-destructive-confirmation-linux.png`, `workbook-inspector-history-linux.png`, `workbook-inspector-public-error-linux.png`, `workbook-inspector-relationships-linux.png`, and `workbook-inspector-rollback-preview-linux.png` under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
- Capture contract: viewport, browser zoom, masks, scroll normalization, screenshot scope, fixture identities, typography, grid layout, and focus contract did not change.
- Refresh: `make browser-e2e-visual-update` passed at `.cartulary/test-results/20260804T143614Z-p554985`.
- Review: all six refreshed images were inspected; deltas were limited to the supported Indicator control and the resulting downstream Inspector vertical layout.
- Ordinary comparison validation: workbook owner service-backed comparison passed at `.cartulary/test-results/20260804T145049Z-p687286`; standalone webserver-backed browser validation passed at `.cartulary/test-results/20260804T145253Z-p723486`.

### 13.2 S-01 — Characterization baseline

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-00 complete; owner-derived Indicator semantics already active |
| Substantive authored files | `packages/view-contracts/src/index.test.ts`; `tools/test_families/package.view_contracts.json`; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Test accounting | Added `package.view_contracts.frontend_unit.characterization_baseline_31fd4bbd8a` with 11 exact titles under the existing shared-contract verification owner; the package manifest now has 5 active rows |
| Coverage added | Supported facade availability; 17-artifact initialization; freeze boundaries; invalid JSON and malformed roots; missing/malformed row identity and version; full-cell completeness; technical-cell rejection; additive row/cell members; sparse patches; view mismatch; required/optional surface handling; registry mismatch and order |
| Intentional exclusions | No assertion accepts `row_version=0`; no assertion treats missing required view-source members as supported. Strict raw source-schema validation remains an intentional S-04 change. |
| Compatibility impact | Test and harness metadata only. Existing valid runtime behavior is unchanged; the baseline protects deliberate later structural moves without preserving approved malformed-input or zero-version cleanup targets. |
| Rollback | Revert the new test block and its authored family row, rerun `make generate`, and remove only the resulting topology digest change. S-00 remains independently intact. |
| Next slice | S-02 shared types and invariant infrastructure |

#### S-01 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units after topology regeneration | `.cartulary/test-results/20260804T151059Z-p809954` |
| `make generate` | PASS | `.cartulary/test-results/20260804T151021Z-p807514` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T151145Z-p810612` |
| `git diff --check` | PASS | No output after the S-01 checkpoint |

The first owner-slice run failed at `.cartulary/test-results/20260804T150937Z-p805317` because the export-availability assertion accidentally depended on JavaScript module-key enumeration order. The assertion was corrected to verify each supported function independently; this was a characterization-test defect, not a product behavior failure. No web implementation or web test changed in S-01, so no web owner slice was affected or skipped.

### 13.3 S-02 — Shared types and invariant infrastructure

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-01 characterization baseline complete |
| Substantive authored files | `packages/view-contracts/src/types.ts`; `packages/view-contracts/src/invariants.ts`; `packages/view-contracts/src/index.ts`; `packages/protocol-ts/src/entrypoints/view-schemas.ts`; the S-01 facade test import was made statically lintable; this tracker |
| Generated files | None; no generated root or harness output changed in S-02 |
| Structure | `types.ts` owns the coherent public metadata/row/surface type graph; `invariants.ts` owns source-aware contract/row failures and shared key/object/enum primitives; `index.ts` explicitly type-re-exports the public graph and consumes both modules |
| Dependency direction | Protocol registry literal types → `types.ts`; no-dependency invariant primitives → semantic code in `index.ts`; the two extracted modules do not import the façade or each other, so the internal graph is acyclic |
| Behavior | Existing exception envelopes, validation order, frozen runtime objects, module initialization, and public names remain unchanged; the historical stage-row comment was replaced by owner-neutral invariant wording |
| Compatibility impact | No runtime or public import-path change. Two additive type-only aliases (`ViewInspectorIncidentRole` and `ViewInspectorSpecializedActionKey`) expose facts already present in the protocol registry. |
| Test accounting | No title or path was added/renamed; the S-01 authored package row still executes the affected behavior, so topology regeneration was not required |
| Rollback | Restore public types and invariant helpers to `index.ts`, remove the two internal files and the two additive protocol aliases; no generated rollback is needed |
| Next slice | S-03 full and sparse row adapters |

#### S-02 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T151539Z-p815783` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T151547Z-p818938` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T151548Z-p819054` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T151548Z-p819091` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T151702Z-p821018` |
| `git diff --check` | PASS | No output after the S-02 checkpoint |

The first Biome run failed at `.cartulary/test-results/20260804T151507Z-p814794` because the newly added files needed formatting and the S-01 facade test used dynamic namespace access. `make format` normalized the authored sources and the test now statically imports every retained function. The concurrent typecheck already passed at `.cartulary/test-results/20260804T151507Z-p814764`; the complete post-fix gate set above also passes.

### 13.4 S-03 — Full and sparse row adapters

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-02 shared types and invariant infrastructure complete |
| Substantive authored files | `packages/view-contracts/src/rows.ts`; `packages/view-contracts/src/types.ts`; `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/index.test.ts`; `tools/test_families/package.view_contracts.json`; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Type contract | `NormalizedViewRowV1` and new `NormalizedViewRowPatchV1` carry distinct required `unique symbol` phantom brands. The brands emit no runtime members and are non-assignable in either direction. |
| Runtime contract | Full rows still require every declared data cell; patches remain sparse; technical/unknown cells fail; additive row/cell members remain ignored; nested values, group semantics, source identity, and freeze boundaries are unchanged |
| Concurrency cleanup | Both adapters now require a positive safe integer and reject zero, negative, fractional, string, missing, infinite, and unsafe versions with one deterministic invariant |
| Consumer migration | All sparse consumers already use `ReturnType<typeof normalizeViewRowPatchV1>`, so they received the stronger type without local casts or aliases; repository typecheck and the full web owner slice confirm compatibility |
| Test accounting | Added the exact title `keeps normalized full rows and sparse patches non-assignable` to the existing characterization row; package manifest remains 5 rows and topology was regenerated through Make |
| Compatibility impact | Intentional compile-time distinction and intentional rejection of `row_version=0`; no valid wire-object shape or schema identity change |
| Rollback | Restore the S-02 row types/functions, remove `rows.ts` and the added test title, then rerun `make generate`; no other slice must be reverted |
| Next slice | S-04 view parser and helper extraction with strict generated decoding |

#### S-03 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T151951Z-p824989` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T152001Z-p828088` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T152019Z-p829496` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T152001Z-p828160` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T152019Z-p829595` |
| `make generate` | PASS | `.cartulary/test-results/20260804T152057Z-p842883` |
| `make explain-test-owner OWNER=package.view_contracts` | PASS | 5 rows; 4 frontend-unit and 1 regression row; Vitest owner route confirmed |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T152206Z-p845654` |
| `git diff --check` | PASS | No output after the S-03 checkpoint |

No S-03 validation failed. No database, route, generated schema, or web implementation file changed. The complete `web.workbook` slice supplies the requested query and collaboration regression evidence rather than relying only on narrower local tests.

### 13.5 S-04 — View parser and helper extraction

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-01 characterization and S-02 shared foundations complete; S-03 remains independently intact |
| Substantive authored files | `tools/protocol-ts/generate-protocol-types.mjs`; `contracts/protocol-ts/frontend-entrypoints.v2.json`; `packages/protocol-ts/src/internal/decoder.ts`; `packages/protocol-ts/src/entrypoints/view-schemas.ts`; protocol tests; `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/view-contracts.ts`; `packages/view-contracts/src/invariants.ts`; package tests and family accounting; this tracker |
| Generated files | Added `packages/protocol-ts/src/generated/view-schema-source-validator.ts`; refreshed generated protocol outputs and `tools/execution_topology_render_index.json` through `make generate`; no generated file was hand-edited |
| Decoder source | The existing authored `tools/schemas/cartulary.view_schema_source.v1.schema.json` is compiled by the protocol generator into a standalone AJV validator; the view-schema entrypoint exposes `viewSchemaSourceDocumentDecoder` typed as `ViewSchemaSourceDocument` |
| Validation order | JSON syntax → generated source-schema decoder → semantic and cross-document invariants. Parser defaults can no longer turn missing required members into accepted values. |
| Diagnostics | Syntax and shape failures use `View contract source validation failed: <source> path=<JSON pointer> reason=<category>`; required and additional-property decoder paths include the exact member while payload values remain absent |
| Extraction | The package `index.ts` is now a one-line façade over `view-contracts.ts`; source parsing, generated artifact initialization, lookup, sort, field-reference, capability, visibility, and still-to-be-extracted surface/Inspector semantics remain behind it without import cycles |
| Behavior evidence | All 17 generated artifacts initialize in registry order; semantic duplicate/reference/Inspector checks remain deterministic; valid field capabilities, ordering, visibility, mutation metadata, and freeze results are unchanged |
| Compatibility impact | Intentionally rejects invalid JSON, malformed roots, missing required members, wrong types, unknown source-schema members, and invalid closed values. Valid source artifacts and browser-facing runtime objects are unchanged. |
| Test accounting | Added one exact package title for source-schema shape branches to the existing characterization row and regenerated topology; protocol decoder evidence extends an already-accounted title/path |
| Rollback | Move the implementation back to `index.ts`, remove the source decoder export and generator output/input changes, restore prior decoder path behavior/tests, remove the added package title, and rerun `make generate`; row adapters remain untouched |
| Next slice | S-05 workbook surface assembly |

#### S-04 passing evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make generate` | PASS after final authored test accounting | `.cartulary/test-results/20260804T153039Z-p876424` |
| `make format` | PASS | `.cartulary/test-results/20260804T153006Z-p863541` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T153016Z-p866842` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 units | `.cartulary/test-results/20260804T153016Z-p866851` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T153016Z-p867093` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T153016Z-p867248` |
| `make generate-drift` | PASS, 4/4 units | `.cartulary/test-results/20260804T153051Z-p878657` |
| `make json-shape-check` | PASS, 3/3 units | `.cartulary/test-results/20260804T153051Z-p878659` |
| `make explain-test-owner OWNER=package.view_contracts` | PASS | 5 rows; package owner route confirmed |
| `make explain-test-owner OWNER=package.protocol_ts` | PASS | 6 rows; protocol owner route confirmed |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T153219Z-p882541` |
| `git diff --check` | PASS | No output after the S-04 checkpoint |

#### S-04 failures encountered and resolved

| Run root | Failure and disposition |
| --- | --- |
| `.cartulary/test-results/20260804T152436Z-p849488` | The first generation run rejected an unsorted authored generated-module allowlist. The new validator path was placed in lexical order; generation passed at `.cartulary/test-results/20260804T152500Z-p851069`. |
| `.cartulary/test-results/20260804T152648Z-p853855`, `.cartulary/test-results/20260804T152708Z-p854895`, `.cartulary/test-results/20260804T152820Z-p856142` | Strict decoding exposed intentionally incomplete test fixtures and assertions that expected later semantic errors. Fixtures were made schema-valid and shape assertions were updated to the new deterministic decoder boundary; the package slice then passed at `.cartulary/test-results/20260804T152918Z-p857338` and again in the final gate set. |
| `.cartulary/test-results/20260804T152928Z-p858250` | Exact-member decoder paths intentionally changed an existing Network Flow unknown-member expectation from the parent path to `/raw_source_value`. The payload-safe assertion was aligned and the protocol slice passed in the final gate set. |
| `.cartulary/test-results/20260804T153016Z-p866768`, `.cartulary/test-results/20260804T153016Z-p866828` | Drift and JSON-shape gates correctly detected the newly added package title after the preceding generation. A final `make generate` refreshed topology; both gates then passed. |

All failures were related to the S-04 implementation or its authored test-accounting order and are resolved. No database, backend route, view identity, Inspector identity, alias, fallback, or package subpath was introduced.

### 13.6 S-05 — Workbook surface assembly

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-02 shared foundations and S-04 parser module complete |
| Substantive authored files | `packages/view-contracts/src/workbook-surfaces.ts`; `packages/view-contracts/src/view-contracts.ts`; `packages/view-contracts/src/index.ts`; package tests and family accounting; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Extracted responsibility | Registry-order joining, injected source construction, required/optional handling, registry/contract agreement, frozen surface records, surface lookup index, 17 schema constants, and three status partitions |
| Dependency direction | `workbook-surfaces.ts` depends on the protocol registry, shared types, and `listViewContracts`; parser/view helpers do not depend on surface assembly, and the façade exports both without a cycle |
| Retained seam | `buildWorkbookSurfaceContracts(sourceContracts?)` remains public and injectable. Its default source is the initialized view-contract list, preserving the web surface registry’s production construction path. |
| Validation matrix | Missing required surface throws; missing optional surface omits; surface-kind and ordered reference-pack mismatches throw; registry order and all constants/partitions match; output arrays and entries remain frozen |
| Compatibility impact | None. Public names, package `.` entrypoint, identities, ordering, omission behavior, errors, and production injection are unchanged. |
| Test accounting | Added one exact constant/partition title to the existing characterization row and regenerated topology; no test path changed |
| Rollback | Restore the surface block to `view-contracts.ts`, remove `workbook-surfaces.ts` and its façade export, remove the added title, and rerun `make generate`; parser and row modules remain intact |
| Next slice | S-06 Inspector module extraction |

#### S-05 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T153610Z-p886743` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T153621Z-p889922` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T153622Z-p890050` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T153622Z-p890072` |
| `make generate` | PASS | `.cartulary/test-results/20260804T153640Z-p891657` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T153653Z-p893895` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T153855Z-p907395` |
| `git diff --check` | PASS | No output after the S-05 checkpoint |

No S-05 validation failed or was skipped. No web implementation file needed migration because its existing package-level import and injected builder use were preserved exactly.

### 13.7 S-06 — Inspector module extraction

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-00 owner/projection repair, S-02 shared foundations, and S-04 strict parser complete |
| Substantive authored files | `packages/view-contracts/src/inspector.ts`; `packages/view-contracts/src/view-contracts.ts`; package Inspector tests; `tools/test_families/package.view_contracts.json`; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Extracted responsibility | Inspector source parsing, closed-value validation, cross-field route and seed invariants, per-view feature closure, exact specialized signatures, and frozen metadata construction |
| Registry authority | `inspector.ts` derives every closed vocabulary, ordered per-view feature list, and specialized signature from `viewInspectorRegistry`; repository search finds no second independent TypeScript registry or feature-key map |
| Exact behavior | All four Indicator features match their complete owner-derived tuple before wildcard families, use the required panels/owners/actions/conditions/results, and reject any `record_patch` substitution or other tuple mismatch |
| Scope boundary | Authorization, transport execution, workflow state, frontend controls, and server enforcement remain outside the package. Inspector roles and disabled conditions remain presentation metadata only. |
| Compatibility impact | None. Valid metadata, public types, initialized artifacts, exception envelopes, and package entrypoint are unchanged. Unknown tokens and unsupported tuples continue to fail closed. |
| Test accounting | Added one exact four-signature title to the existing characterization row and regenerated topology; the package manifest remains at 5 active rows |
| Rollback | Restore Inspector parsing to `view-contracts.ts`, remove `inspector.ts` and the added title, then rerun `make generate`; the S-00 owner repair remains independently required |
| Next slice | S-07 façade and compatibility cleanup |

#### S-06 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T154322Z-p911864` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T154335Z-p915065` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T154335Z-p915170` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T154335Z-p915197` |
| `make generate` | PASS | `.cartulary/test-results/20260804T154351Z-p916773` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T154409Z-p919037` |
| `make service-backed-test-slice OWNER=module.workbook` | PASS, 66/66 units | `.cartulary/test-results/20260804T154448Z-p932038` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T154851Z-p969688` |
| `git diff --check` | PASS | No output after the S-06 checkpoint |

No S-06 validation failed or was skipped. The extraction changed module cohesion only; it did not move security, workflow, transport, or UI responsibilities into the package.

### 13.8 S-07 — Façade and compatibility cleanup

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-03 through S-06 complete |
| Substantive authored files | `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/types.ts`; `packages/view-contracts/src/view-contracts.ts`; `packages/view-contracts/src/index.test.ts`; this tracker |
| Generated files | None; no generated root or harness output changed in S-07 |
| Final façade | `index.ts` contains only explicit runtime and type re-exports. The package manifest still exposes only `.`, and repository search finds no package subpath import. |
| Approved removals | `parseViewContractJSON` is absent from the public runtime export set but remains directly available to package tests and module initialization in the internal parser module. `ViewRowV1` was removed entirely. |
| Retained API | Both normalized row types/adapters, all used view/surface functions and constants, `buildWorkbookSurfaceContracts`, and the coherent Inspector metadata type graph remain public. |
| Importer audit | No production or test consumer imported either removed name through `@cartulary/view-contracts`; no workspace caller required migration or an alias. All 83 direct package-import occurrences remain on the `.` entrypoint. |
| Initialization evidence | The package slice still loads and validates all 17 artifacts. Its exact runtime-key assertion confirms the raw parser is not accidentally re-exported and every intended value remains available. |
| Test accounting | The existing façade test was strengthened without changing its title or path; no authored family or topology update was required |
| Compatibility impact | Intentional removal of two unused private-workspace API seams only. There is no valid wire behavior, identity, package path, or runtime initialization change. |
| Rollback | Restore the S-06 star exports and `ViewRowV1`; restore the test import to the façade. No generator or consumer rollback is needed. |
| Next slice | S-08 validation and handoff completion |

#### S-07 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS after final test correction | `.cartulary/test-results/20260804T155105Z-p979049` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T155108Z-p982034` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T155116Z-p983182` |
| `make frontend-import-boundary-check` | PASS, 2/2 units | `.cartulary/test-results/20260804T155116Z-p983201` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T155116Z-p983214` |
| Repository-wide removed-symbol and subpath scans | PASS | Parser references are internal implementation/test references only; `ViewRowV1` and package subpath imports have zero matches |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T155248Z-p984940` |
| `git diff --check` | PASS | No output after the S-07 checkpoint |

The first package slice failed at `.cartulary/test-results/20260804T155029Z-p976568` because the new expected export-name array was not itself sorted before comparison; the runtime export set was already correct. Sorting both sides fixed the assertion. One initial composite audit shell command also had a quoting error and was replaced by separate read-only scans. No product, consumer, or harness-accounting failure remains, and no S-07 check was skipped.

### 13.9 S-08 — Validation and handoff completion

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04; overall remediation `DONE` |
| Dependencies | S-07 and every preceding slice complete |
| Authored owner accounting | `platform.viewschema`: 3 rows; `package.protocol_ts`: 6 rows; `package.view_contracts`: 5 rows; `module.indicators`: 24 rows, 11 service-backed; `web.workbook`: 119 rows; `module.workbook`: 89 rows, 56 service-backed |
| Test-title/path audit | All new titles resolve through their authored manifests. Only `module.indicators.json` and `package.view_contracts.json` required title/row changes in this effort; no S-08 title or path changed. |
| Topology audit | `make target-plan` reported 616 planned `check` units, 572 cacheable units, 214 fixtures, digest `sha256:a299ea816654335788b36bbd2288d0a9382b93b67931a2e20bfa96c08ed3a27a`; final generation/drift and full execution passed. |
| S-08 authored fixes | Added the two new Indicator source paths to `tools/frontend_source_ownership.json`; removed the now-unused `requireEnumStringFromSet` helper from `tools/contractgen/validation.go`. Both omissions were exposed by the first broad `check`, not by a change in selected product semantics. |
| Generated confirmation | Final `make generate` passed at `.cartulary/test-results/20260804T160836Z-p1276381`. All Go/protocol/view-schema/Inspector validators, bindings, artifacts, topology outputs, and visual goldens were produced through Make-owned generation/update paths; none was hand-edited. |
| Compatibility summary | Existing HTTP paths, `cartulary.view.timeline.v2`, `cartulary.view.indicators.v1`, `cartulary.inspector_config.v1`, and the package `.` entrypoint are preserved. No database migration, alias, fallback, identity bump, or backend authorization transfer was introduced. Intentional incompatibilities are limited to malformed source rejection, `row_version < 1` rejection, compile-time full/sparse separation, corrected Inspector panel placement/dispatch, and removal of two unused private-workspace exports. |
| Retained-run maintenance | Skipped by design because `RESULTS_DIR` was unset for `make agent-finalize`; the target itself passed and no successful retained full warm-check root was claimed. |
| Remaining risks | No known functional, specification, generated-drift, ownership-accounting, security-boundary, or validation blocker remains. The uncommitted implementation still requires ordinary reviewer acceptance of the owner edits and refreshed visual baselines. |
| Rollback | Revert S-07 through S-01 in reverse checkpoint order for isolated package changes. Revert S-00 owner/projection/application/UI/tests/goldens as one compatibility unit, then rerun `make generate`; do not retain Indicator runtime changes against the prior owner corpus. The two S-08 cleanup edits should follow the files that made them necessary. |
| Handoff | Review the owner amendments and generated Inspector registry first, then the typed Indicator executor/handler/UI flow, package module split, strict decoder/row types, test accounting, and S-00 visual evidence. Every required focused and broad gate passes on this worktree. |
| Tracker closure | `make lint-markdown` passed at `.cartulary/test-results/20260804T162619Z-p1560317`; `git diff --check` passed with no output after the final handoff was written |

#### S-08 final validation sequence

| Ordered command | Result | Run root or detail |
| --- | --- | --- |
| 1. `make lint-markdown` | PASS | `.cartulary/test-results/20260804T160853Z-p1278732` |
| 2. `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T160901Z-p1280253` |
| 3. `make generate-drift` | PASS, 4/4 units | `.cartulary/test-results/20260804T160911Z-p1280677` |
| 4. `make generated-artifact-policy-check` | PASS, 3/3 units | `.cartulary/test-results/20260804T160920Z-p1283252` |
| 5. `make json-shape-check` | PASS, 3/3 units | `.cartulary/test-results/20260804T160924Z-p1283680` |
| 6. `make test-slice OWNER=platform.viewschema` | PASS, 3/3 units | `.cartulary/test-results/20260804T160932Z-p1284133` |
| 7. `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 scheduled units | `.cartulary/test-results/20260804T160938Z-p1284703` |
| 8. `make test-slice OWNER=package.view_contracts` | PASS, 6/6 scheduled units | `.cartulary/test-results/20260804T160947Z-p1288932` |
| 9. `make test-slice OWNER=module.indicators` | PASS, 27/27 scheduled units | `.cartulary/test-results/20260804T160953Z-p1289838` |
| 10. `make service-backed-test-slice OWNER=module.indicators` | PASS, 13/13 scheduled units | `.cartulary/test-results/20260804T161010Z-p1293859` |
| 11. `make test-slice OWNER=web.workbook` | PASS, 120/120 scheduled units | `.cartulary/test-results/20260804T161025Z-p1295311` |
| 12. `make service-backed-test-slice OWNER=module.workbook` | PASS, 66/66 scheduled units | `.cartulary/test-results/20260804T161108Z-p1308312` |
| 13. `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T161305Z-p1344026` |
| 14. `make frontend-import-boundary-check` | PASS, 2/2 units | `.cartulary/test-results/20260804T161318Z-p1344537` |
| 15. `make browser-e2e-webserver-backed` | PASS, 62/62 units | `.cartulary/test-results/20260804T161324Z-p1344980` |
| 16. `make agent-finalize` | PASS, 1/1 unit | `.cartulary/test-results/20260804T161655Z-p1397125`; `RESULTS_DIR` unset |
| 17. `make test-fast` | PASS, 349/349 units | `.cartulary/test-results/20260804T161711Z-p1399669` |
| 18. `make check` | PASS, 715/715 units | `.cartulary/test-results/20260804T161836Z-p1435478` |
| 19. `git diff --check` | PASS | No output |
| 20. `git status --short` | PASS, expected dirty handoff | Only the intended authored, generated, visual-evidence, test-accounting, and tracker paths were present |

#### S-08 failure and recovery record

The first exact sequence stopped at `make check` root `.cartulary/test-results/20260804T160326Z-p1148447` with 713/715 units passing. Both failures were related to this remediation:

- `web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` found the new `IndicatorInspectorWorkflow.test.tsx` and `indicatorInspectorHandlers.ts` absent from the authored frontend source-ownership manifest.
- Go staticcheck found `requireEnumStringFromSet` unused after Inspector closed-vocabulary validation moved to the owner-derived registry.

The ownership manifest and obsolete helper were corrected. `make test-slice OWNER=web.architecture` then passed 12/12 at `.cartulary/test-results/20260804T160824Z-p1275215`; `make lint-go-staticcheck` passed; `make generate` refreshed downstream artifacts; and the complete 20-command sequence above was restarted from command 1. No final validation was skipped and no unrelated failure remains.
