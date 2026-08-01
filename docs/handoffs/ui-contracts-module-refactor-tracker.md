# ui-contracts Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | `packages/ui-contracts` |
| Normalized target label | `ui-contracts` |
| Output path | `docs/handoffs/ui-contracts-module-refactor-tracker.md` |
| Mode | Authorized sequential implementation of S-00 through S-08 |
| Allowed change in this session | Authored specifications, contracts, tests, catalog inputs, UI Contracts/app sources, harness tooling, guides, generated output through `make generate`, retained evidence, and this tracker, limited to the approved remediation plan |
| Non-goals | No database, persisted-data, HTTP, WebSocket, route, authorization, mutation, revision, projection, or deployment contract change; no public selector subpath; no hand edit of protected generated output or lockfiles |
| Discovery baseline | Commit `24f3010158934a61bed02804fc97093ef625ec56`; the planning pass observed unrelated evidence/backend/generated worktree edits and did not modify them |
| Tracker-creation baseline | Commit `e2b57a73bf7388a35404fa9b74f73c317d3e2cd8` (`main`); the worktree was clean before this tracker was created |
| Current implementation baseline | `main` commit `9f0a22c6f5d8016061a074dc0d0b969b09609cc8`; immediately before S-00 the only tracked worktree change was this tracker, already staged by the user |
| Current implementation status | S-00 through S-07 are complete. S-08 implementation, API/caller accounting, impact mapping, focused owners, static/build gates, and four broad checks pass; retained-run `agent-finalize` is blocked only by reproducible scheduler timing-health thresholds. |
| Baseline mismatch | The repository advanced after tracker creation. The tracked target inventory remains six files, but `src/index.ts` gained two coordination selectors and grew from 2,349 to 2,353 lines; live importing files changed from 155 to 151. This revision uses the current repository state. Older commits, counts, and validation run roots remain historical evidence only. |
| Authorization | The user expressly authorized the complete S-00 through S-08 plan. Each slice remains gated by its predecessor's validation and tracker checkpoint. |

The target path exists. Its presence is not treated as proof that it is a valid
permanent module boundary. The live implementation is assessed below as a
legitimate shared selector facade that also contains responsibilities whose
long-term ownership requires further evidence.

This tracker is the controlling implementation-support artifact subordinate to
the adopted owners. It MUST NOT create, amend, or infer product behavior.
Implementation MUST preserve valuable observable contracts named in Section 4,
while obsolete private TypeScript structure MUST be removed by coordinated
workspace migration when the declaration-accurate baseline authorizes it.

The key words `MUST`, `MUST NOT`, `SHOULD`, and `MAY` are normative in this
tracker:

| Term | Meaning in this tracker |
| --- | --- |
| `MUST` / `MUST NOT` | An implementation conformance requirement or prohibition. Failure blocks completion. |
| `SHOULD` / `SHOULD NOT` | The expected implementation posture. A deviation requires recorded owner evidence, rationale, risk, and compensating verification. |
| `MAY` | An optional action that remains subject to scope, authority, and all `MUST` requirements. |

Decision and risk states have the following closed meanings:

| State | Meaning |
| --- | --- |
| `RESOLVED` | Planning authority and evidence determine one default decision. Implementation, when required, still needs authorization and passing evidence. |
| `IMPLEMENTATION_GATED` | The decision boundary is known, but source movement MUST wait for the named baseline, artifact, or compatibility gate. |
| `DEFERRED` | No current implementation need or sufficient evidence justifies action. The current behavior and surface MUST remain unchanged. |
| `BLOCKED: <reason>` | Work within the named blocked scope MUST stop until the stated contradiction or missing authority is repaired. Unrelated slices MAY continue. |

The authority order used for this tracker is:

1. Adopted subsystem NLSpecs, within their named subsystems.
2. Core 00 through Core 04, for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior plans, handoffs, and this framework as evidence only.

If owner documents disagree, this tracker records `BLOCKED: owner
contradiction` and does not select a side. Repository/framework mismatches are
planning findings, not inferred behavior.

### Owner and planning documents inspected

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/testing-harness-nlspec.md`
- `docs/network-flow-activity-nlspec.md`
- `docs/design.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`
- `docs/reference-pack-subsystem-nlspec.md`, as draft evidence only
- `docs/research/nlspec-spec.md`, for NLSpec writing form only
- `temp/analysis-notes.md`, as non-authoritative revision input

The frontend development guide describes `packages/ui-contracts` narrowly as
the owner of runtime-safe selector/test-ID builders. The live package also
contains generated design-token exports, density parsing, application-shell
types and labels, and feature-specific vocabulary. That difference is a
repository/framework mismatch and is carried forward as an architectural
finding rather than resolved here.

### Repository files inspected

All tracked files under the target were inspected directly:

- `packages/ui-contracts/package.json`
- `packages/ui-contracts/tsconfig.json`
- `packages/ui-contracts/src/index.ts`
- `packages/ui-contracts/src/index.test.ts`
- `packages/ui-contracts/src/design-tokens.test.ts`
- `packages/ui-contracts/src/generated/design-tokens.ts`

Related contract, generation, policy, and accounting inputs inspected were:

- `contracts/design/tokens.v1.json`
- `contracts/verification/owners/package.ui.json`
- `tools/frontend_import_boundaries.json`
- `tools/generated_artifact_policy.json`
- `tools/harness/generated-artifacts/design-tokens/design-token-cli.mjs`
- `tools/harness_helper_ownership.json`
- `tools/test_catalog_owner.json`
- `tools/test_families/package.ui.json`

Inbound imports were enumerated across `apps` and `packages`. Representative
callers inspected directly included:

- `apps/web/src/app/AppRoot.tsx`
- `apps/web/src/app/IncidentAdminPanel.tsx`
- `apps/web/src/app/landingAdminTypes.ts`
- `apps/web/src/shared/workbookShellContracts.ts`
- `apps/web/src/testing/selectorContractPolicy.test.ts`
- `apps/web/src/workbook/components/WorkbookShellSlots.tsx`
- `apps/web/src/workbook/hooks/useIncidentControlsDrawer.ts`
- `apps/web/src/workbook/models/workbookSurfaceRegistry.ts`
- `packages/grid-adapter/src/SemanticDataGrid.tsx`
- `packages/grid-adapter/src/index.test.tsx`
- `packages/grid-adapter/src/test-support.tsx`
- `packages/test-utils/src/focus.ts`
- `packages/test-utils/src/grid-editing.ts`
- `packages/test-utils/src/grouping.ts`
- `packages/test-utils/src/index.test.ts`
- `packages/test-utils/src/marker.ts`
- `packages/test-utils/src/scrolling.ts`
- `packages/protocol-ts/package.json`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `apps/web/src/shared/workbookSheetRef.ts`

Search results were used to enumerate callers and contract families. A search
hit alone was not treated as proof of a file's behavior.

## 2. Current-State Repository Inventory

At current planning commit `9f0a22c6f5d8016061a074dc0d0b969b09609cc8`,
`packages/ui-contracts` has six tracked files. The two additional files found
under the directory are ignored tool/dependency artifacts and are explicitly
out of scope.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `packages/ui-contracts/package.json` | Declares the private workspace package and its single public root entry point. | Package `@cartulary/ui-contracts`; only `"."` is exported as `./src/index.ts`. | Workspace package resolution and manifests, including `apps/web`, `packages/grid-adapter`, and `packages/test-utils`. | `@cartulary/protocol-ts`; development-only Vitest dependency. | Indirectly exercised by all package-owned and consumer tests. | None directly. | `package.ui` | Medium | Changing the root export or adding subpaths is a public package-surface decision and is deferred. |
| `packages/ui-contracts/tsconfig.json` | Strict, source-only TypeScript configuration extending the repository base configuration. | No runtime surface; includes `src`, uses bundler resolution, and emits no files. | Repository frontend type-check tooling. | `../../tsconfig.base.json`. | `make frontend-typecheck`. | None. | `package.ui` | Low | Configuration change is not part of the planned structural split. |
| `packages/ui-contracts/src/index.ts` | Root facade for runtime-safe selectors/test IDs, validation and encoding helpers, selector vocabularies, shell labels/types, semantic grid contracts, generated-token access, density parsing, and feature-specific selector families. | 2,353 lines; 229 exported functions; 274 authored exported declarations plus five generated re-exports. | Current TypeScript import inventory finds 151 files: 90 under `apps/web/src`, 51 under `apps/web/e2e`, and 10 under `packages`. | `listViewSchemaRegistryEntries` from `@cartulary/protocol-ts`; `./generated/design-tokens`. | `src/index.test.ts`, `src/design-tokens.test.ts`, consumer unit tests, grid-adapter tests, test-utils tests, and browser tests. | Reads the generated view-schema registry through `protocol-ts`; re-exports the protected generated design-token projection. | Shared selector facade: `package.ui`; embedded responsibilities have candidates in Section 3. | High | The current source includes the coordination selectors `party-partial-completion` and `party-retry-created-link`. It is a legitimate facade but also a mixed-responsibility, high-fan-in monolith. No backend, transport, persistence, authorization, mutation, revision, or projection implementation was found. |
| `packages/ui-contracts/src/index.test.ts` | Package-level selector-contract characterization. | No production exports; 23 Vitest cases. | Package test catalog row and local test runner. | `./index`; Vitest. | The file is the test. | Indirectly validates registry-backed view identifiers through the public facade. | `package.ui` test ownership | Medium | Existing cases cover many selector families but do not directly cover every exported public family or invalid-input branch. |
| `packages/ui-contracts/src/design-tokens.test.ts` | Characterizes the generated-token facade, theme CSS, and density values. | No production exports; three Vitest cases. | Package test catalog row and local test runner. | `./index`; Vitest. | The file is the test. | Tests exports generated from `contracts/design/tokens.v1.json`. | `package.ui` test ownership with design-token projection ownership | Medium | It must consume the authored facade, not hand-edit or bypass the generated output. |
| `packages/ui-contracts/src/generated/design-tokens.ts` | Generated TypeScript projection containing token variables, default theme identity, types, and CSS text. | Generated `cartularyDesignTokenVars`, `cartularyDefaultThemeId`, generated token-name/theme types, and `cartularyDesignThemeCssText`. | Re-exported through `src/index.ts`; used by runtime consumers through the package facade. | None at runtime. | `src/design-tokens.test.ts`; drift and artifact-policy checks. | Generated from `contracts/design/tokens.v1.json` by the design-token harness helper. | Design-token owner projection plus generated-artifact harness | High | Begins with `Code generated ... DO NOT EDIT` and is protected by `tools/generated_artifact_policy.json`. Never hand-edit. |
| `packages/ui-contracts/node_modules/**` | Installed dependency/tool links, including `.bin/vitest`. | Not a package-owned source surface. | Local package tooling only. | Workspace dependency installation. | Not applicable. | Tool-managed installation artifact. | Workspace package manager | Low | Ignored and explicitly out of scope. Do not edit or account as authored inventory. |
| `packages/ui-contracts/tsconfig.tsbuildinfo` | TypeScript incremental-build cache. | None. | TypeScript tooling only. | TypeScript compiler. | Not applicable. | Tool-managed cache. | Repository-local tooling | Low | Ignored and explicitly out of scope. Do not edit or treat as a contract. |

## 3. Module Boundary Diagnosis

The current target is:

- a legitimate thin public application-support facade for stable selectors and
  test IDs;
- a mixed-responsibility package and internally accidental catch-all;
- transport-adjacent only in the sense that browser automation and runtime DOM
  surfaces consume it, not because it implements HTTP or WebSocket transport;
- view/projection-adjacent only because it validates view-schema identifiers
  against a generated registry and supplies projection-facing UI anchors;
- frontend shell/controller-adjacent because it exports shell slots, labels, and
  state vocabulary;
- grid-adapter-adjacent because it supplies semantic grid classes, attributes,
  selectors, and density-derived row heights, without importing the grid vendor;
- a misplaced possible home for some vocabulary currently owned semantically by
  `apps/web`, `packages/grid-adapter`, view contracts, or adopted subsystem
  owners.

It is not a persistence adapter, mutation coordinator, backend module,
authorization implementation, revision implementation, projection refresh
implementation, or direct grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Stable selector construction, opaque identifier encoding, CSS attribute escaping, and fail-closed validation | `src/index.ts` | `package.ui` / shared UI-contract facade | keep | Public builders are imported by runtime code, E2E tests, grid adapter, and test utilities; frontend guide assigns this seam to `packages/ui-contracts`. | This is the package's legitimate deep-module responsibility. Exact output bytes are frozen. |
| Auth, account, admin, landing, incident, workbook, saved-view, timeline, collaboration, entity, evidence, coordination, import, assessment, reference-pack, and debug selector families | `src/index.ts` | `package.ui` facade, with semantic inputs governed by the matching subsystem or application owner | split | Large owner-specific type/function families coexist in one 2,353-line file and are consumed across runtime and tests. | Split means private source grouping with unchanged root re-exports, not a decision to relocate public imports. |
| Workbook shell slots, shell labels, admin panel tokens, load state, and controller-facing state types | `src/index.ts` | `apps/web` for composition, controller/navigation state, transient load state, and private copy; `package.ui` for selector bytes, semantic DOM contracts, stable test IDs, and intentionally shared accessibility names | split | S-04 moved the three app-state types, retained valuable shared shell contracts, and removed obsolete private exports without aliases. | RB-002 is `IMPLEMENTED`; the S-04 report and gates prove the resulting 236-name boundary. |
| `WorkbookSurface = string` and view-schema identifier validation | `src/index.ts` | Generated `SheetRef` from `@cartulary/protocol-ts/core-http` for workbook identity; registry-validated strings for view-schema-specific selectors; `package.ui` for encoded selector tokens only | move | The package imports `listViewSchemaRegistryEntries`; generated Core HTTP types define a discriminated `SheetRef` union for view schema, saved view, and extension workspace. | RB-003 is resolved. S-05 removes `WorkbookSurface`, adopts `SheetRef` for semantic identity, and retains exact selector validation and bytes. |
| Grid density parsing, row-height lookup, semantic scrollport class, role/data attributes, row/cell/group/filter/sort selectors | `src/index.ts` | Shared selector facade for automation; `packages/grid-adapter` for semantic DOM implementation and density application | split | `SemanticDataGrid.tsx`, grid-adapter tests, and test-utils consume these exports; no `react-data-grid` import exists in `ui-contracts`. | Preserve semantic selector outputs. Vendor integration remains correctly isolated in the adapter. |
| Generated design-token exports and theme CSS facade | `src/generated/design-tokens.ts` and root re-exports | Design-token owner projection plus generated-artifact harness; `package.ui` remains the authored access facade under current policy | keep | Token registry, generator, import-boundary policy, artifact policy, and design-token tests all point to this seam. | Generator/input redesign is blocked by RB-001. Generated output is never hand-edited. |
| Network Flow-specific selector types and builders | `src/index.ts` | Adopted Network Flow owner for semantics; `package.ui` for the stable cross-runtime/test automation facade | split | Network Flow runtime, unit, and E2E callers use the shared builders; the adopted Network Flow NLSpec governs the subsystem behavior. | Group privately and add direct characterization before movement. Do not infer subsystem behavior from function names. |
| Test/debug harness selectors used by production-rendered DOM | `src/index.ts` | `package.ui` automation facade; `packages/test-utils` owns choreography, not selector bytes | keep | Runtime components attach IDs and test-utils/E2E use them; the development guide explicitly makes shared test IDs a runtime-safe contract. | Production use is intentional, not evidence that test choreography belongs in production. |
| HTTP, WebSocket, storage, authorization, mutation, revision, and projection behavior | Not implemented in this package | Matching backend/platform/application owners | defer | Direct inspection found no fetch, route handler, WebSocket, SQL/storage, authorization-decision, source mutation, revision-write, or projection-refresh implementation. | The package supplies anchors used by tests of those behaviors; phase/test maps are evidence accounting only. |

The following allocation is normative for later ownership decisions:

| Concern | Required owner posture | Prohibited inference |
| --- | --- | --- |
| Shell composition, controller/navigation state, transient load state, private application copy | `apps/web` MUST own runtime state and composition after the coordinated S-04 migration. | An exported declaration's current location MUST NOT be treated as permanent semantic ownership or require a forwarding alias. |
| Selector bytes, semantic DOM contracts, stable test IDs, intentionally shared accessibility names | `package.ui` MUST retain the cross-runtime/test contract unless an adopted owner explicitly relocates it. | Production use of test-ID builders MUST NOT be classified as test-only leakage. |
| Workbook, view-schema, saved-view, and extension identity | The applicable generated protocol or registry owner MUST supply semantic identity. | Labels, routes, visible tab order, selector strings, and `WorkbookSurface` MUST NOT define identity. |
| Grid implementation and row-height application | `packages/grid-adapter` MUST own the semantic DOM implementation and vendor lifecycle; `package.ui` MAY expose stable semantic selectors and derived values. | Shared grid selectors MUST NOT introduce a vendor import into `ui-contracts`. |
| Authorization and availability | Server and adopted subsystem owners remain authoritative. | UI types, selector presence, registry membership, or `SheetRef.kind` MUST NOT authorize access. |

## 4. Public Contract and Behavior Freeze Map

The default for every structural slice is exact compatibility for valuable
observable behavior: selector/test-ID bytes, validation failures, CSS escaping,
closed-set order used by consumers, semantic DOM attributes, accessibility
names, token values, CSS text, density results, and the active package root
import. Private TypeScript declarations and forwarding aliases are not
observable contracts merely because the private `0.0.0` package once exported
them. Their default disposition is coordinated removal when S-02 proves that no
continuing consumer or owner value exists.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Package root import and public TypeScript surface | `package.ui` | `package.json` exposes only `@cartulary/ui-contracts`; 151 current importing files. | Type-check, package unit tests, consumer unit/E2E tests. | Produce a current-commit TypeScript-program API and export-ownership baseline before internal extraction. | High | Preserve the single root import path. S-03 preserves all 279 baseline names; S-04 and S-05 then remove only the baseline-approved private structure through coordinated migration, with target counts of 236 and 235. Public selector subpaths and compatibility aliases are prohibited. |
| Exact test-ID and selector strings | `package.ui`, with subsystem-owned semantic inputs | Public builders in `src/index.ts`; runtime components and tests share outputs. | Broad coverage in `index.test.ts`, app tests, grid tests, test-utils, and E2E. | Direct tests for currently uncovered families and output bytes. | High | Percent encoding, separators, prefixes, CSS escaping, and DOM query shapes are observable contracts. |
| Fail-closed selector input validation | `package.ui` | `require*` helpers reject empty, malformed, or unknown values and validate view-schema IDs. | Several invalid-input cases in `index.test.ts`. | Add missing invalid-input branches for under-tested public families. | High | Opaque IDs remain syntax-opaque and losslessly encoded; do not infer identifier structure. |
| Exported vocabulary arrays and ordering | `package.ui`, with matching semantic owner | Exported `as const` arrays and union types in `src/index.ts`. | Partial direct unit coverage and consumer compilation. | Characterize array membership and order where runtime consumers or tests rely on iteration. | Medium | Order is frozen until evidence establishes that it is unobservable. |
| Workbook shell/controller types and ARIA labels | `apps/web` candidate; currently exported by `package.ui` | `workbookShellContracts.ts`, `WorkbookShellSlots.tsx`, incident-admin/controller callers. | App unit tests and browser tests. | Directly freeze exact shell slot labels and compile-time public type compatibility before any move. | High | Label text is user-facing accessibility behavior, not merely an internal constant. |
| Grid semantic DOM and UI-selector contract | `packages/grid-adapter` implementation plus `package.ui` selector facade | `SemanticDataGrid.tsx`, grid-adapter tests, selector policy test. | Grid-adapter unit tests, package selector tests, test-utils, E2E. | Add direct coverage for filter helpers, history-open helpers, and any untested semantic attributes. | High | Preserve scrollport class, roles, data attributes, row/cell/group selectors, and absence of vendor imports outside the adapter. |
| Density-derived grid row heights | Design-token projection and `package.ui` facade; adapter consumes result | `workbookGridRowHeightPx` parses generated fixed-pixel density tokens. | Token tests assert token values; callers compile. | Add direct tests for all densities and the invalid-token failure path. | Medium | Exact numeric results and thrown error semantics are frozen. |
| Generated design-token exports, token order, default theme, and theme CSS | Design-token projection/harness; exposed by `package.ui` | `tokens.v1.json`, generated file, generator, artifact and import-boundary policies. | `design-tokens.test.ts`; drift, policy, and JSON-shape checks. | Existing tests cover the current facade; a generator redesign additionally requires the machine-contract evidence below and remains blocked by RB-001. | High | Never hand-edit generated output. CSS text and token iteration order are observable. |
| Saved-view and view-schema selector behavior | Saved-view/view-contract owners plus `package.ui` facade | Registry-backed validation and saved-view selector builders. | Package tests and saved-view app/E2E tests. | Freeze registry acceptance/rejection and exact saved-view selectors before changing registry access. | High | No saved-view persistence or schema parsing is implemented here. |
| Entity row/query/mutation UI anchors | Entity owner semantics plus `package.ui` facade | Entity selector families and entity-focused callers. | App/E2E coverage; partial package coverage. | Direct tests for reusable-entity selectors, merge preconditions, and invalid entity tokens. | Medium | The package does not perform entity query or mutation behavior. |
| Evidence, timeline, revision/change-set, history, conflict, and projection-refresh UI anchors | Matching subsystem owners plus `package.ui` facade | Feature selector families and app/E2E callers. | Cross-layer app and browser tests; partial package tests. | Direct tests for row-history open helpers, draft-row creation, paste-conflict helpers, and projection-refresh anchors before movement. | Medium | The package supplies stable DOM anchors only; source mutation, revisions, conflicts, and projection refresh remain outside it. |
| Network Flow selector surface | Adopted Network Flow subsystem owner plus `package.ui` facade | Network Flow types/builders and runtime/unit/E2E callers. | Consumer tests; incomplete direct package characterization. | Directly characterize every public Network Flow builder, output byte, validation branch, and relevant exported array order. | High | Retain the shared automation facade even if source is privately grouped. |
| Authorization-related selector surface | Security/application owners plus `package.ui` facade | Auth/account/admin selectors and runtime callers. | Auth/admin app and E2E tests; package selector tests. | Freeze exact selectors when moved; no authorization-decision characterization belongs in this package. | Medium | No authorization check or outcome is computed here. |
| HTTP routes, request/response envelopes, and WebSocket paths/events | Backend transport/application owners; not `ui-contracts` | No route handler, fetch client, WebSocket implementation, or event logic found in the target. | Service/browser suites outside this package use UI anchors. | None for a selector-only mechanical split; run affected browser targets if selector bytes or shell behavior change. | Low, indirect | Route and event behavior remains frozen globally but is not implemented by this package. |
| Storage semantics, source mutation, revision writes, projection refresh, and authorization decisions | Matching domain/platform/application owners; not `ui-contracts` | Direct target inspection found none of these implementations. | Backend and browser evidence outside package ownership. | None in this package. | Low, indirect | Test rows and UI anchors do not transfer runtime ownership to this package. |
| Harness/test accounting | Testing Harness owner and `package.ui` catalog inputs | `package.ui` owner/family manifests route two unit rows and four static/boundary rows. | `make test-slice OWNER=package.ui` passed 6/6 at the discovery commit. | Update authored catalog inputs only if a later slice adds a separately routed test file or changes ownership. | Medium | Harness maps are verification accounting, not runtime architecture. |

### Root API and export-ownership contract

Before any source movement, S-02 MUST retain a current-commit API and
export-ownership baseline. The report MUST use the installed TypeScript compiler
API and MUST NOT add a reporting dependency or a permanent compatibility
framework.
Consumer analysis MUST use the TypeScript program so type-only imports,
re-exports, aliases, and indirect root consumers are counted. Text search alone
is not completion evidence.

Each of the 279 current root names MUST have exactly one row with these fields:

| Field | Required value |
| --- | --- |
| Export identity | Export name, declaration kind, generated status, and exact public signature or type shape. |
| Consumption | Runtime consumers, test consumers, and type-only consumers at the baseline commit. |
| Effects | Exact observable outputs, failures, ordering, labels, DOM semantics, token/CSS values, or `none`. |
| Ownership | Semantic owner and implementation owner, recorded separately. |
| Disposition | One of `keep`, `move_coordinated`, `internalize`, or `remove_coordinated`; absence of a consumer is evidence only and MUST be corroborated by complete program, catalog, generated, and dynamic-consumer accounting. |
| Browser evidence | Exact catalog row IDs required by the observable-change matrix, or `none` with a reason. |
| Compatibility | Compatibility class, reason, deprecation state, named replacement when one exists, and rollback export. |

Compatibility classes are closed:

| Class | Meaning and default |
| --- | --- |
| `observable` | Retained output, failure, order, DOM, accessibility, token, density, or active-import behavior; MUST remain byte- and behavior-compatible. |
| `workspace_active` | TypeScript surface with a current in-repository consumer or clear continuing owner value; MUST remain until a coordinated owner migration completes. |
| `generated` | Protected projection surfaced by the root facade; MUST be regenerated, never hand-edited. |
| `internal` | Non-exported implementation detail; MAY move privately while public behavior remains exact. |
| `obsolete_private` | Private-workspace surface with no current consumer or continuing owner value; MUST be internalized or removed in the coordinated cleanup and MUST NOT receive a deprecated alias. |

Repository-unreferenced exports are removal candidates, not automatic deletion
authority. Coordinated removal requires TypeScript-program consumer evidence,
generated and catalog searches, namespace/dynamic-consumer rejection, exact
browser-impact accounting, a recorded owner disposition, and a passing
compatibility baseline. The package is private and version `0.0.0`, so a name
that passes this gate MUST be removed without a deprecated export, forwarding
module, or indefinite alias.

### Identity contract

Semantic identity and selector representation MUST remain separate:

| Use | Required identity |
| --- | --- |
| Navigation, startup, and presence | Generated `SheetRef` from `@cartulary/protocol-ts/core-http`. |
| View-schema APIs | Opaque, registry-validated `ViewSchemaId`; the client MUST NOT duplicate the registry. |
| Saved views | A distinct saved-view identity; a view-schema ID or label MUST NOT substitute for it. |
| Extension workspaces | The `SheetRef` extension-workspace variant. |
| Selector string segment | An opaque encoded selector token; encoding MUST be lossless and deterministic. |
| Labels, routes, visible tab order | Never identity. |

`WorkbookSurface = string` is obsolete selector vocabulary. New route, startup,
presence, authorization, saved-view, extension, or other semantic identity APIs
MUST NOT use it. S-05 MUST remove it; view-schema-specific selector builders
MUST accept strings and validate them against the canonical registry, while
semantic callers use generated `SheetRef`. Implementations
MUST NOT create custom sheet/schema variants, infer authorization from identity
or registry membership, duplicate the protocol registry, or introduce a
`protocol-ts` to `ui-contracts` dependency cycle.

### Design-token machine contract

The sole executable authored input is
`contracts/design/tokens.v1.json`. Executable tests, generators, conformance,
and release tooling MUST NOT read, stat, hash, or otherwise access
`docs/design.md` or any other `docs/` path. `docs/design.md` owns the human YAML
registry and explicitly defines the JSON file as its already-resolved machine
projection. The Testing Harness owner row names the same machine input.

| Property | Required contract |
| --- | --- |
| Schema and owner | The input MUST declare a stable schema/version, owner provenance, verification identity, and exactly one default theme. Unknown schema versions, missing owner fields, duplicate keys after decoding, or an absent default theme MUST fail. |
| Namespaces | Only the owner-projected `border`, `colors`, `component`, `density`, `elevation`, `layout`, `motion`, `rounded`, `spacing`, and `typography` token namespaces are admitted by the current contract. New namespaces require an adopted design-owner revision and matching machine projection. |
| Ordering | Input objects MUST be normalized into deterministic Unicode code-point lexical order by emitted token name. Exported arrays, type unions, and CSS declarations MUST use that same stable order. Source-object insertion order MUST NOT affect output. |
| References | A reference MUST resolve to one admitted token in the same registry. Missing references, cross-registry references, and cycles MUST fail with deterministic diagnostics. A generator that does not support references MUST fail on reference syntax rather than emit it literally. |
| Failure atomicity | Schema, namespace, value, reference, cycle, and output-collision validation MUST complete before protected output replacement. A failed run MUST leave the previous output bytes unchanged and MUST exit non-zero. |
| Provenance | Generated output or its retained generation evidence MUST record the SHA-256 of the exact machine-input bytes and a stable generator identity/version. Paths or Markdown hashes MUST NOT serve as provenance. |
| Output | `packages/ui-contracts/src/generated/design-tokens.ts` is protected output. Repeated generation from identical input and generator identity MUST be byte-identical and MUST preserve the current token names, TypeScript types, default-theme export, CSS text, ordering, and density values unless an owner-authorized behavior change says otherwise. |

A generator redesign is conformant only when focused tests prove execution with
no `docs/` tree, a file-access guard rejects documentation access, two clean
generations are byte-identical, negative schema/reference/cycle cases fail
before replacement, and the unchanged fixture produces unchanged token, CSS,
type, order, and density output.

### Protocol-import comparison contract

RB-004 defaults to keeping the current protocol root-registry import. A narrow
seam MAY be adopted only after an authorized comparison uses the same commit,
lockfile, toolchain, build mode, browser, cache policy, and fixture for:

| Variant | Definition |
| --- | --- |
| A | Current root-registry import and current initialization path. |
| B | Exactly one candidate narrow import with equivalent registry behavior. |

The retained report MUST include the commit and lockfile hashes, commands,
toolchain, Vite manifest/chunk graph, compressed JavaScript by initial and lazy
chunk, critical requests, module side effects/cycles, and at least 20 comparable
cold starts per variant with raw and p50/p95 results. Noise and exclusion rules
MUST be declared before comparison.

Variant B MAY replace A only if it removes a side effect or dependency cycle,
eliminates an unrelated eager module or critical request, reduces compressed
JavaScript by at least `max(10 KiB, 2%)`, or improves p95 startup by at least
5 ms beyond measured noise. If no threshold is crossed, evidence is missing,
or equivalence fails, the implementation MUST keep Variant A.

### Observable-change to browser-evidence mapping

An authorized runtime change MUST be classified before implementation. The
current test catalog supplies the exact row IDs; the table below selects the
required Make-owned stage and prohibits discretionary under-testing.

| Observable change class | Required browser evidence |
| --- | --- |
| Private partition with identical exports, bytes, initialization order, and component files | No browser row; package tests, frontend unit, type-check, boundary check, and `make build-web` are required. |
| Selector/test-ID byte, escaping, prefix, validation, or DOM-query change | Every exact catalog row that consumes the changed export; at minimum the applicable `make browser-e2e` row. Zero matched rows fails closed. |
| ARIA name, role, focus, keyboard, or accessibility-state change | Exact consuming functional rows plus the applicable `make browser-e2e-a11y` row. |
| Client-only shell composition, navigation, or transient-state change | Exact consuming rows through `make browser-e2e-webserver-backed`; add a11y evidence when accessible output changes. |
| Server-backed authorization, mutation, collaboration, save-state, conflict, or projection behavior | Exact consuming rows through `make browser-e2e-stateful`; add functional/a11y rows required by changed output. |
| CSS token, theme CSS, density, geometry, or visual-layout change | Exact consuming functional rows plus `make browser-e2e-visual`; add a11y when contrast, focus, or semantic output changes. |
| Import seam, code splitting, critical request, or startup performance change | `make browser-e2e-measurement` plus an exact functional smoke row and the RB-004 report. |
| Generated-token regeneration with byte-identical output | No browser row; generated drift, artifact-policy, JSON-shape, and package-owner evidence are required. |

The retained impact record MUST name its schema/version, commit, changed
exports, change classes, exact catalog row IDs, stages, projects, commands, run
roots, omitted stages, and omission reasons. An observable runtime change that
maps to zero catalog rows MUST fail closed until ownership/accounting is fixed.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Direct package characterization is incomplete for public Network Flow builders, row-height parsing, filter helpers, history-open helpers, reusable-entity selectors, and several invalid-input paths. | Export-to-test comparison against `src/index.test.ts` and `src/design-tokens.test.ts`; consumer tests provide indirect coverage but do not isolate all package contracts. | A mechanical split could silently change selector bytes or failure behavior while broad tests remain insensitive. | `must_fix` | `package.ui` test ownership, with semantic inputs from matching owners | Add behavior-focused direct characterization before moving each affected family. Update owner-catalog accounting only through authored inputs if routing changes. |
| `src/index.ts` combines selector core, application-shell vocabulary, grid semantics, domain-feature families, Network Flow vocabulary, and design-token facade logic. | 2,353 lines, 229 functions, 274 authored declarations plus five generated re-exports, and 151 importing files. | Large blast radius and difficult review; unrelated owners are co-located. | `should_fix` | `package.ui` root facade with private owner-grouped modules | Partition privately by selector core, app/shell, workbook/grid, domain feature, Network Flow, and token facade while preserving root exports, initialization, bytes, failures, and order. |
| Shell/controller types and labels live in the selector package despite the guide assigning application shell and state to `apps/web`. | Runtime callers include shell-contract, admin-panel, and controller files. | Moving them can break public types and user-facing ARIA labels. | `fixed` | `apps/web` for state; `package.ui` for shared automation labels where required | S-04 gives application state one app owner and retains shared observable labels; compiler, unit, boundary, Fallow, and build evidence passes. |
| `WorkbookSurface` is an unconstrained `string` while canonical workbook identity is the generated `SheetRef` union. | Root type declaration; generated `core-http` `SheetRef`; current app-local handwritten sheet-reference type. | Continued semantic use can accept invalid identity; abrupt tightening can break callers or create a dependency cycle. | `must_fix` | Protocol Core HTTP for `SheetRef`; view-contract owner for registry-validated identifiers; `package.ui` for selector encoding only | RB-003 is `RESOLVED`: S-05 performs one coordinated migration, removes the obsolete alias and old helper path, and preserves selector bytes. |
| Root module initialization imports the `protocol-ts` registry facade to validate view-schema IDs. | `listViewSchemaRegistryEntries` import; protocol root aggregates generated registry families. | Possible broad module loading or bundle impact has not been measured. | `defer` | View-contract/protocol owner with a narrow consumer seam | RB-004 is `RESOLVED: keep by default`. Change only when the comparison contract crosses a materiality threshold. |
| Apparently unused exported arrays, types, and selectors remain part of the root surface. | Repository-wide search finds no consumers for some exports, but text search misses type-only/indirect use and cannot grant removal authority. | Keeping obsolete exports increases coupling and makes future removal harder. | `must_fix` | `package.ui` | RB-005 is `RESOLVED: coordinated cleanup`. S-02 must prove the exact consumer/disposition inventory; S-04 internalizes 36 declarations, deletes four builders, and retains three intentional generated facade exports. |
| Production components consume shared test-ID builders. | Runtime imports from `apps/web` and grid adapter; frontend guide requires runtime-safe shared selectors. | Misclassifying this as test leakage would fragment selector strings. | `intentional/no_action` | `package.ui` | Retain the shared runtime/test facade; keep choreography in `packages/test-utils`. |
| Semantic grid selectors and density helpers are outside the grid adapter, but direct vendor integration is not. | `ui-contracts` has no `react-data-grid` import; `SemanticDataGrid.tsx` consumes the shared semantic contracts. | Moving all grid selectors into the adapter could break shared automation imports. | `intentional/no_action` | `package.ui` facade plus `packages/grid-adapter` implementation | Preserve dependency direction and enforce the existing vendor import boundary. |
| Generated design tokens are accessed through the authored root facade. | Import-boundary policy, artifact policy, generator, and tests all enforce this path. | Bypassing the facade or editing output risks generated drift. | `intentional/no_action` | Design-token projection/harness and `package.ui` facade | Keep current access path; regenerate only through Make-owned tooling when separately authorized. |
| The design-token owner and Testing Harness helper row previously disagreed about executable input. | `docs/design.md`; `docs/testing-harness-nlspec.md`; `contracts/design/tokens.v1.json`; repository procedure. | Leaving the contradiction unresolved would invite documentation parsing and circular authority. | `must_fix` | Design-token owner and Testing Harness owner | S-00 defines the human-owner/machine-projection boundary and names only the JSON projection for executable use. S-07 must prove the no-docs boundary. |
| No domain logic under transport/persistence, direct SQL/storage coupling, hidden mutation/revision/projection side effects, or wrong-layer authorization checks were found in the target. | Direct inspection of every tracked target file and dependency/import search. | Inventing such workstreams would expand scope and obscure the actual frontend seam. | `intentional/no_action` | Existing backend/platform/module owners | Keep backend facade, route, storage, and authorization refactors out of this target plan. |
| Caller counts changed from 153 at discovery commit `24f301...`, to 155 at tracker commit `e2b57a...`, to 151 at current commit `9f0a22c...`; the current target also gained two coordination selectors. | Current TypeScript import inventory and commit-range inspection. | A stale caller or API inventory could omit consumers or exports during a future move. | `should_fix` | Future implementing session | Re-run TypeScript-program caller and root-export analysis at the implementing commit. Retain older counts as historical evidence only. |
| Browser verification was previously selected case by case without a closed observable-change mapping. | Existing tracker browser row was `TODO`; catalog rows and Make-owned browser stages are available. | An observable change could receive no evidence, or a private no-op split could trigger disproportionate evidence. | `must_fix` | Testing Harness accounting plus affected semantic owner | RB-006 is `RESOLVED`: apply the Section 4 matrix and exact row IDs; a runtime change mapping to zero rows fails closed. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Support-owner and contract admission | root | none | WF-01, WF-07 | Repair the token authority contradiction and adopt the coordinated-removal and browser-impact contracts before source work. | Adopted design and Testing Harness owners, frontend guide, and this tracker; never generated output by hand. | Owner review; harness, JSON shape, generated-policy, and documentation checks. | RB-001 is resolved; compatibility and impact policies match the approved remediation. |
| WF-01 | Direct characterization | chain | WF-00 | WF-02, WF-03 | Characterize every public family selected for movement, including positive bytes, failures, and ordering. | Package tests and authored catalog inputs only when test routing changes. | `make test-slice OWNER=package.ui`; `make frontend-unit`. | Every moved family has direct package evidence tied to the current source state. |
| WF-02 | Current-commit root API and owner baseline | chain | WF-01 | WF-03, WF-04, WF-05 | Generate the complete root API/export-ownership matrix with TypeScript-program consumer analysis. | Read-only target and consumer graph plus retained refactor report. | `make frontend-typecheck`; report self-validation; source/name audit. | Every current export has one disposition, owners, consumers, effects, browser rows, compatibility class, and rollback seam. |
| WF-03 | Private facade partition | chain | WF-01, WF-02 | WF-04, WF-05, WF-06 | Extract owner-grouped private modules while preserving the single root facade exactly. | Authored `packages/ui-contracts/src/**`, excluding `src/generated/**`. | Package owner slice, frontend unit/type/boundary, build, and generated drift. | All public exports, selector/failure/order/label/token/density bytes, and module evaluation are unchanged. |
| WF-04 | Application-state ownership migration | chain | WF-02, WF-03 | WF-05 | Move matrix-classified controller state to `apps/web`; internalize or remove approved obsolete declarations without aliases. | `packages/ui-contracts`, `apps/web`, focused tests, and ownership guide. | Static gates plus exact Section 4 browser rows when observable output is affected. | Root API has 236 names, no package-to-app edge exists, and app state has one owner. |
| WF-05 | Canonical identity migration | chain | WF-04 | WF-06 | Replace handwritten workbook identity with generated `SheetRef`, remove `WorkbookSurface`, and keep view-schema selector validation. | App, protocol/view-contract consumers, tests, and `ui-contracts` selector signatures. | Type/boundary/unit/build and exact affected browser rows. | Root API has 235 names; no old path, custom variant, registry duplicate, auth inference, cycle, or selector drift remains. |
| WF-06 | Protocol import decision | chain | WF-05 | WF-07 | Inspect the post-refactor graph and retain the current import unless the static trigger requires the one controlled narrow-seam comparison. | Retained graph/report and only a qualifying source seam. | `make build-web`; measurement and functional smoke only if the candidate is exercised. | A retained same-source-state report records keep/change; missing or sub-threshold evidence records keep. |
| WF-07 | Token generator hardening | chain | WF-06 | WF-08 | Implement strict, safe, deterministic, provenance-bearing, atomic generation from the flat machine projection. | Authored registry schema, generator helper/CLI/tests, policies; generated output only via `make generate`. | `make harness-contract`; `make generate`; package owner slice; drift, policy, JSON-shape, and script checks. | Documentation-free positive, negative, determinism, provenance, and atomicity evidence passes with runtime-identical token exports. |
| WF-08 | Accumulated validation and handoff | chain | WF-07 | none | Run required package, frontend, generated, browser-impact, and repository gates; update restartable evidence. | Complete implementation diff and this tracker. | Section 8 ladder, `make agent-finalize`, and `make check`. | All applicable `UI-RF-CLOSE-*` rows pass at one source state; rollback points, commands, run roots, scope, and omissions are current. |

## 7. Proposed Refactor Slice Plan

The user authorized every implementation slice below as one sequential effort.
A slice is complete only after its focused checks pass, this tracker records the
files, decisions, commands, run roots, failures, rollback point, and next action,
and Markdown plus patch hygiene pass. A failed slice is recorded as blocked and
the next slice MUST NOT begin.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | Current tracker | Repair the design/Testing Harness token authority boundary, align the frontend guide, and adopt coordinated-removal and impact policies. No product behavior changes. | `docs/design.md`, Testing Harness NLSpec, frontend guide, and this tracker. | Owner text could create a documentation dependency or change evidence routing. | Existing harness/schema/policy evidence; no generator change in this slice. | `make lint-markdown`; `make harness-contract`; `make json-shape-check`; `make generated-artifact-policy-check` | Revert the four authored documentation changes as one unit; the existing JSON generator path remains intact. | RB-001 is resolved; executable input is the resolved JSON projection; compatibility policy matches the approved cleanup. |
| S-01 | S-00 | Split root-facade characterization by contract family and add Network Flow, density, filter/history/draft/paste/reusable-entity/merge, order, invalid-input, and coordination cases. | Package tests plus authored `package.ui` family inputs and generated topology. | Tests could bless inferred rather than live behavior. | Assert exact current bytes, errors, labels, arrays, tokens, CSS, and density values through the public facade. | `make generate`; `make test-slice OWNER=package.ui`; `make frontend-unit`; `make generate-drift` | Revert only test/catalog authored inputs and regenerate; production remains unchanged. | Every family selected for movement or cleanup has direct isolated evidence. |
| S-02 | S-01 | Emit and validate the current root API/ownership report using the installed TypeScript compiler API. | Read-only package/consumer graph and retained run-root artifact. | Missing type-only, alias, namespace, dynamic, catalog, or generated consumers could falsely authorize removal. | Self-validate 279 names, declaration shapes, consumers, effects, owners, and dispositions. | Report validation; `make frontend-typecheck` | Discard the report without changing package source. | Every baseline export has one complete disposition; the 43 zero-consumer names match the approved cleanup classes. |
| S-03 | S-02 | Partition `src/index.ts` into acyclic private selector, token, schema, shell/grid, interaction, entity/evidence, saved-view, and Network Flow modules behind explicit root exports. | Authored `packages/ui-contracts/src/**`; never `src/generated/**`. | Export/signature omission, selector/failure/order drift, side-effect reordering, cycle, token initialization drift. | Preserve S-01 and consumer tests unchanged; compare the S-02 API report exactly. | Package owner, frontend unit/type/boundary, build, Fallow, and generated-drift gates. | Revert one private module extraction at a time; the root facade is the rollback seam. | Root imports, 279 names, signatures, observable outputs, failures, order, and initialization are identical. |
| S-04 | S-03 | Move app-owned state to `apps/web`, internalize 36 approved declarations, delete four unused helpers, and retain the three intentional generated-token facade exports. No aliases. | `packages/ui-contracts`, `apps/web`, focused tests, and ownership guide. | Reverse dependency, duplicate authority, missed consumer, shell/ARIA/focus drift. | App-state unit/compile evidence plus facade and exact impact-selected tests. | Package/frontend static, unit, build, and generated gates. | Restore the removed declarations to their extracted owner modules without undoing S-03. | Root API has 236 names; state has one owner; all removed names have zero consumers; outputs remain exact. |
| S-05 | S-04 | Adopt generated `SheetRef`, rename the app helper seam to `sheetRef.ts`, remove the old path and `WorkbookSurface`, and preserve registry validation in view-schema selectors. | App callers/tests, protocol types, E2E helpers, and `ui-contracts` signatures. | Identity conflation, rejected extensions, duplicated registry, auth inference, cycle, selector-byte drift. | Unit owners plus exact startup, saved-view, presence, and claimed/unclaimed Network Analysis browser rows. | Focused owners, frontend gates, build, generated drift, and exact browser evidence. | Restore each semantic call site and helper file as one migration unit; do not add a forwarding module. | Root API has 235 names; identity distinctions match owners and no obsolete identity path remains. |
| S-06 | S-05 | Inspect the final bundle graph; keep the current protocol root import unless an eager unrelated module, cycle, or critical-request trigger exists. If triggered, compare only `@cartulary/protocol-ts/view-schema-registry`. | Retained same-source-state graph/report; source only after a qualifying comparison. | Non-comparable runs, graph/init drift, speculative subpath complexity. | Registry equivalence; at least 20 comparable cold starts and one smoke row only if a candidate is tested. | `make build-web`; measurement and smoke evidence only when triggered. | Measurement changes no source; the current root import is the rollback default. | Complete keep/change report applies exact thresholds; missing or sub-threshold evidence records keep. |
| S-07 | S-06 | Harden the flat token generator with strict parsing/validation, safe values, code-point sorting, SHA-256 provenance, v2 identity, atomic replacement, and no-docs tests. | Registry schema, generator helper/CLI/tests; protected output only via `make generate`. | Documentation dependency, unsafe CSS, nondeterminism, partial replacement, provenance drift, runtime token change. | Positive, negative, no-docs, determinism, provenance, collision, and atomicity fixtures. | `make harness-contract`; `make generate`; package owner; drift/policy/JSON/script gates. | Revert authored helper/schema/tests and regenerate the prior output; never edit output directly. | All machine-contract assertions pass and runtime token/CSS exports remain identical. |
| S-08 | S-07 | Re-run API/caller inventories, classify the complete diff, execute all applicable gates, finalize retained evidence, and complete this tracker. | Entire implementation diff, evidence roots, and tracker. | Stale evidence, skipped rows, generated drift, scope expansion. | All package/consumer/generated/browser evidence selected by the impact matrix. | Section 8 ladder; `make agent-finalize`; `make check` | Revert the smallest failing slice; do not merge unrelated cleanup into rollback. | Every applicable `UI-RF-CLOSE-*` criterion passes at one source state with current handoff evidence. |

No slice may hand-edit `packages/ui-contracts/src/generated/design-tokens.ts`.
During S-03, test files and generated output stay unchanged. If S-01 adds a new
test family or changes catalog ownership, that accounting work updates authored
inputs and then regenerates the applicable topology. Worktree state MUST be
inspected before every implementation slice; no slice may overwrite unrelated
changes. The tracker MUST be updated and validated after every slice before the
next begins.

## 8. Validation Plan

### Slice checkpoint ledger

The ledger is the restart boundary. A successor remains `NOT STARTED` until its
predecessor is `COMPLETE` and the predecessor's final Markdown and patch-hygiene
checks pass.

| Slice | Status | Source state and substantive files | Decisions and impact | Validation and retained evidence | Failure, rollback, and next action |
| --- | --- | --- | --- | --- | --- |
| S-00 | COMPLETE | Baseline `9f0a22c6f5d8016061a074dc0d0b969b09609cc8` plus the implementation worktree. Updated `docs/design.md`, `docs/testing-harness-nlspec.md`, the frontend implementation guide, this tracker, and the stale catalog-count assertions in `tools/harness/tests/test-harness-contracts.mjs`. | The YAML registry is human authority; `contracts/design/tokens.v1.json` is the resolved executable projection. Valuable observable behavior is stable; obsolete private TypeScript structure uses coordinated removal without shims. No runtime, selector, token, wire, route, persistence, authorization, or generated-output change. | `make lint-markdown` PASS at `.cartulary/test-results/20260801T023229Z-p2157548` and final checkpoint root `.cartulary/test-results/20260801T023612Z-p2167776`; `make json-shape-check` PASS at `.cartulary/test-results/20260801T023241Z-p2159201`; `make generated-artifact-policy-check` PASS at `.cartulary/test-results/20260801T023241Z-p2159199`; `make generate-drift` PASS at `.cartulary/test-results/20260801T023332Z-p2161442`; `make harness-contract` PASS at `.cartulary/test-results/20260801T023432Z-p2166050`; final `git diff --check` PASS. | Initial harness run failed at `.cartulary/test-results/20260801T023241Z-p2159438` because its hard-coded catalog totals were stale (`1006/1871` versus live `1008/1891`); updating the existing exact-count guard made the rerun pass. Rollback is the five authored S-00 files as one unit; the JSON generator path remains intact. Next: S-01 test partition. |
| S-01 | COMPLETE | Replaced `packages/ui-contracts/src/index.test.ts` with public-facade suites `selector-core.test.ts`, `application-selectors.test.ts`, `workbook-shell-grid-selectors.test.ts`, `workbook-interaction-selectors.test.ts`, and `network-flow-selectors.test.ts`; updated `tools/test_families/package.ui.json`, exact harness catalog totals, and generated `tools/execution_topology_render_index.json` through `make generate`. | Production is unchanged. All 23 prior cases remain partitioned; new direct cases cover exact shell/group order and labels, every density and malformed-token parsing, filter builders, history-open, draft/paste, reusable entity, merge preconditions, both coordination selectors, and all Network Flow selector/identity families. `package.ui` now has nine rows; global catalog is 60 owners, 215 active families, 1,011 rows, and 1,894 selectors. | `make generate` PASS at `.cartulary/test-results/20260801T024445Z-p2182976`; `make test-slice OWNER=package.ui` PASS 9/9 at `.cartulary/test-results/20260801T024543Z-p2186360`; `make frontend-unit` PASS at `.cartulary/test-results/20260801T024632Z-p2187861`; `make generate-drift` PASS at `.cartulary/test-results/20260801T024632Z-p2187745`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260801T024710Z-p2193941`; `make harness-contract` PASS at `.cartulary/test-results/20260801T024710Z-p2193968`; `make lint-biome` PASS at `.cartulary/test-results/20260801T024816Z-p2199829`; final tracker lint and `git diff --check` follow this update. | Generation failed at `.cartulary/test-results/20260801T024333Z-p2175868`, `.cartulary/test-results/20260801T024359Z-p2178258`, and `.cartulary/test-results/20260801T024423Z-p2180632` until row IDs/titles were ASCII-sorted and overlapping ownership was removed. The first package slice failed at `.cartulary/test-results/20260801T024456Z-p2185222`; it correctly exposed live density values `24/32/40` and Network Flow's opaque percent-encoded field segment. Initial Biome formatting failed at `.cartulary/test-results/20260801T024755Z-p2195918`; `make format` at `.cartulary/test-results/20260801T024807Z-p2196573` changed only the new tests. Rollback is the five test files, manifest/count input, and regenerated topology as one unit. Next: S-02 declaration-accurate baseline. |
| S-02 | COMPLETE | Used TypeScript `6.0.2` compiler analysis over the 279-name root facade and 154 current importing files. Retained `.cartulary/test-results/20260801T025200Z-ui-contracts-s02/cartulary.ui_contracts_api_baseline.v1.json` with SHA-256 `1b5f6a37fb73f2671288675c054ee3831b42c7bbde6b9c276bd1e53b2bc7524e`. Removed the package-local namespace-introspection alias test and its catalog row because namespace/dynamic consumers are prohibited; regenerated the topology index and refreshed exact catalog totals. | Every export has declaration kind/shape/file, generated status, runtime/type/test consumers, observable effect, semantic and implementation owners, disposition, compatibility class, browser impact, and rollback export. There are 28 names with no import at all and exactly 43 with no continuing consumer outside package characterization: 36 `internalize`, four helper removals plus `WorkbookSurface` as five `remove_coordinated`, three app-state `move_coordinated`, and 235 `keep`. The 43-name set contains three retained generated facade exports. Namespace, dynamic, default, unknown, and non-workspace consumer counts are zero. Current `package.ui` routing has eight rows; global catalog totals are 60 owners, 215 active families, 1,010 rows, and 1,893 selectors. | Report self-validation and `sha256sum -c` PASS; `make generate` PASS at `.cartulary/test-results/20260801T025051Z-p2202690`; `make test-slice OWNER=package.ui` PASS 8/8 at `.cartulary/test-results/20260801T025423Z-p2205880`; `make generate-drift` PASS at `.cartulary/test-results/20260801T025423Z-p2205802`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260801T025423Z-p2205969`; catalog/generated symbol search, final tracker lint, and `git diff --check` pass. | The first analyzer attempt counted package characterization imports as continuing API demand and reported 28 rather than the approved 43-name cleanup set; the retained schema now records both counts explicitly. The only namespace consumer was the obsolete alias-introspection test, removed with its exact catalog row rather than exempted. Rollback is the retained report plus that test/catalog/topology accounting unit; production remains unchanged. Next: S-03 private production partition using the retained 279-name report as the equality gate. |
| S-03 | COMPLETE | Replaced the 2,353-line authored monolith with an explicit 309-line root facade and ten private modules: `selectorCore.ts`, `designTokens.ts`, `viewSchemaSelectors.ts`, `applicationSelectors.ts`, `workbookShellSelectors.ts`, `gridSelectors.ts`, `workbookInteractionSelectors.ts`, `entityEvidenceSelectors.ts`, `savedViewSelectors.ts`, and `networkFlowSelectors.ts`. Updated `tools/frontend_import_boundaries.json` so only the private design-token facade may import protected token output. Tests, catalog inputs, and generated token output were not edited in this mechanical slice. | Root package path and all 279 names, declaration kinds/shapes, generated status, consumers, selector bytes, failures, closed order, labels, DOM semantics, tokens/CSS, densities, and initialization results remain compatible. Static graph validation finds ten modules, no cycle, no root `export *`, and no top-level expression statement. Private dependency direction is toward selector core/view-schema/token seams; no vendor import or package subpath was added. No observable change maps to browser evidence. | Post-partition report `.cartulary/test-results/20260801T030400Z-ui-contracts-s03/cartulary.ui_contracts_api_baseline.v1.json`, SHA-256 `79f70af28664d6619e7ab8983f598dcd1620da28e365f55aa816f266b4332e02`; API/graph/generated-byte comparison PASS. `make test-slice OWNER=package.ui` PASS 8/8 at `.cartulary/test-results/20260801T030238Z-p2218634`; `make frontend-unit` PASS at `.cartulary/test-results/20260801T030238Z-p2218714`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260801T030313Z-p2221967`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260801T030344Z-p2228069`; `make frontend-fallow-static` PASS at `.cartulary/test-results/20260801T030313Z-p2222048`; `make lint-biome` PASS at `.cartulary/test-results/20260801T030313Z-p2222083`; `make generate-drift` PASS at `.cartulary/test-results/20260801T030313Z-p2221766`; `make build-web` PASS at `.cartulary/test-results/20260801T030351Z-p2228688`; `make json-shape-check` PASS at `.cartulary/test-results/20260801T030426Z-p2232499`; `make harness-contract` PASS at `.cartulary/test-results/20260801T030426Z-p2232639`; final Markdown/diff hygiene follows. | Initial partition generation was rejected twice before file mutation because dependency analysis found core-to-application and core-to-interaction cycles; moving the feature-specific validators to their owners produced the acyclic graph. The first boundary run failed at `.cartulary/test-results/20260801T030313Z-p2222010` because policy still named `index.ts`; the authored policy now names `designTokens.ts` while consumers remain root-only. The first API comparison reported only an order difference in the compiler's expanded `CartularyDesignTokenVarName` union; normalized member comparison and unchanged generated bytes prove no API drift. Rollback is one module at a time through the explicit facade, with the pre-partition file recoverable as the S-03 unit. Next: S-04 coordinated cleanup. |
| S-04 | COMPLETE | Added app-owned `LandingAdminPanelId`, `IncidentControlsSection`, and `IncidentControlsLoadState` declarations in `apps/web/src/app/landingAdminTypes.ts`; migrated source and E2E callers. Removed the corresponding package types, internalized the 36 approved declarations, and deleted `systemViewSelectorTestId`, `gridSavedRowSelector`, `gridGroupRowTestIdPrefix`, and `entityMentionResolutionStatusTestId` without aliases. Package builders now accept strings and validate them against private closed sets. | Root API is exactly 236 names. The generated-token facade exports remain; `EntityType` remains because the S-02 compiler report records an active test-support consumer. Selector/test-ID bytes, errors, labels, order, ARIA/focus/shell behavior, routes, tokens, and density values are unchanged. Change class is coordinated type ownership/root cleanup with no observable browser output, so no browser row is selected. There is one app owner, no package-to-app edge, package subpath, forwarding module, namespace/dynamic/default/unknown consumer, or removed-name import. | Post-cleanup report `.cartulary/test-results/20260801T031300Z-ui-contracts-s04/cartulary.ui_contracts_api_baseline.v1.json`, SHA-256 `59ab75c6337ab3b6b40cf7e575866e375f221c7083a23df9da17e3273e3392fe`. `make test-slice OWNER=package.ui` PASS 8/8 at `.cartulary/test-results/20260801T031254Z-p2239985`; `make frontend-unit` PASS at `.cartulary/test-results/20260801T031254Z-p2240227`; `make frontend-typecheck` PASS at `.cartulary/test-results/20260801T031359Z-p2251598`; `make frontend-import-boundary-check` PASS at `.cartulary/test-results/20260801T031254Z-p2240169`; `make frontend-fallow-static` PASS at `.cartulary/test-results/20260801T031254Z-p2240235`; `make generate-drift` PASS at `.cartulary/test-results/20260801T031254Z-p2239886`; `make build-web` PASS at `.cartulary/test-results/20260801T031359Z-p2251802`; `make lint-biome` PASS at `.cartulary/test-results/20260801T031359Z-p2251641`; `make lint-markdown` PASS at `.cartulary/test-results/20260801T031503Z-p2256989`; final `git diff --check` PASS. | The first typecheck failed at `.cartulary/test-results/20260801T031054Z-p2237809`: private string validation needed explicit narrowing, package characterization still imported `NetworkAnalysisSelector`, and app test support retained an active `EntityType` import. The first API check correctly found 235 rather than 236 because `EntityType` had been removed despite that active consumer; it was restored. Rollback is the three app-owned types plus the approved 40-name obsolete surface as one S-04 unit; S-03 modules remain. Next: S-05 canonical identity migration. |
| S-05 | COMPLETE | Replaced the handwritten app union with generated `SheetRef` projected by the sole authorized app helper facade `apps/web/src/shared/sheetRef.ts`; renamed helpers to `isSheetRef`, `sheetRefKey`, and `sheetRefsEqual`; deleted `workbookSheetRef.ts` without forwarding. Runtime, test, E2E, startup, saved-view, collaboration, preference, and extension callers migrated together. Removed root `WorkbookSurface`; view-schema-specific package/app/test helpers use strings and existing registry validation. Updated the frontend guide, import-boundary allowlist, and source-ownership manifest. | Root API is exactly 235 names. Navigation/startup/presence/saved-view/extension identity resolves to generated `SheetRef`; the app facade contains validation/equality functions but no authored type alias or variant. `WorkbookStartupRejectedReference` models invalid cleared-pointer diagnostics, not valid identity. No old path/name, custom valid identity, selector-vocabulary cast, duplicate registry, authorization inference, package cycle, or package-to-app edge remains. Exact selector bytes, route/wire/storage/auth behavior, startup fallback, saved-view separation, presence, and claimed/unclaimed extension semantics are unchanged. Browser impact is the five exact S-02 rows named below. | API report `.cartulary/test-results/20260801T033900Z-ui-contracts-s05/cartulary.ui_contracts_api_baseline.v1.json`, SHA-256 `736239dae00cca4881bf0853a9dc6e0f17234c7195a3afde152ba78098dea509`; 235 exports, 145 facade-importing files, and zero forbidden consumers. Focused slices PASS: `package.ui` 8/8 at `.cartulary/test-results/20260801T032334Z-p2266746`; `web.workbook` 118/118 at `.cartulary/test-results/20260801T032335Z-p2266763`; `module.workbook` 12/12 and 88 tests at `.cartulary/test-results/20260801T032334Z-p2266756`; `module.savedviews` 1/1 and 17 tests at `.cartulary/test-results/20260801T032335Z-p2266799`; `module.collaboration` 5/5 and 32 tests at `.cartulary/test-results/20260801T032335Z-p2266794`; `module.networkflow` 1/1 and 58 tests at `.cartulary/test-results/20260801T032718Z-p2376881`. Exact service-backed rows PASS: workbook startup `module.workbook.browser.browser_startup_honors_explicit_sheet_home_defau_ee3b02ee01` at `.cartulary/test-results/20260801T032942Z-p2408118`; saved-view persistence/default startup `module.savedviews.browser_stateful.verify_saved_view_persistence_default_startup_su_3fd1edb0fd` at `.cartulary/test-results/20260801T033020Z-p2426427`; multi-client presence `module.collaboration.browser.two_analysts_on_the_same_incident_see_each_other_d099712cac` at `.cartulary/test-results/20260801T033057Z-p2444533`; claimed Network Analysis `module.networkflow.browser.network_flow_selector_covers_when_the_extension_4ce49e0d6f` at `.cartulary/test-results/20260801T033256Z-p2483297`; unclaimed Network Flow `module.networkflow.browser.network_flow_selector_covers_when_the_extension_c90fca78ee` at `.cartulary/test-results/20260801T033331Z-p2501385`. Final `make frontend-unit` PASS at `.cartulary/test-results/20260801T033737Z-p2542673`; typecheck at `.cartulary/test-results/20260801T033816Z-p2545435`; boundary at `.cartulary/test-results/20260801T033816Z-p2545520`; Fallow at `.cartulary/test-results/20260801T033816Z-p2545501`; build at `.cartulary/test-results/20260801T033816Z-p2545926`; generated drift at `.cartulary/test-results/20260801T033816Z-p2545055`; JSON shape at `.cartulary/test-results/20260801T033816Z-p2545100`; Biome at `.cartulary/test-results/20260801T033816Z-p2545585`; final Markdown/diff hygiene follows. | The first generated-type compile exposed a base/saved-only E2E preference assertion; it now compares extension identity too. Boundary failure `.cartulary/test-results/20260801T033440Z-p2523420` required the single app facade rather than direct runtime protocol imports. Frontend-unit failures `.cartulary/test-results/20260801T033440Z-p2523357` and `.cartulary/test-results/20260801T033624Z-p2540244` exposed the stale source-ownership filename; the final rerun passes. Concurrent Network Flow unit root `.cartulary/test-results/20260801T032335Z-p2266778` and combined two-profile browser root `.cartulary/test-results/20260801T033135Z-p2462668` failed only harness scheduler/result accounting; isolated owner/profile reruns pass. Rollback is the generated-type/app-facade migration plus the one root export as one S-05 unit. Next: S-06 protocol-import evidence gate. |
| S-06 | COMPLETE | Inspected the final production bundle, then compared the current root import with exactly the authorized temporary `@cartulary/protocol-ts/view-schema-registry` candidate. The candidate source, export, boundary allowance, and import were removed after measurement; the retained source state is variant A. | Static inspection triggered comparison because the initial chunk contains 14 protocol sources, 12 unrelated to the registry. Candidate behavior was equivalent but retained the same 12 unrelated eager sources and ten critical requests, introduced no cycle improvement, increased initial gzip by 238 bytes, and improved p95 by only 0.09 ms wall/0.30 ms navigation. No threshold passed, so RB-004 resolves to `keep`; no protocol subpath or compatibility seam persists. | Decision report `.cartulary/test-results/20260801T034700Z-ui-contracts-s06/cartulary.ui_contracts_protocol_import_decision.v1.json`, SHA-256 `d8d1a90a98e139684e7176dc28aa5a51031065581161d02d99afa99d04149328`; bundle report SHA-256 `1f597e4a6a0959a0973c99ac1cdebd8b72f8768c65d270d086f76df194646764`; cold-start report SHA-256 `0dd501e5a356ff5ed8bb5c9c8964a8292608752cae47025dd8f5c60c8ab43556`. Each variant has five discarded warmups and 20 retained alternating fresh-browser samples with no exclusions. Temporary candidate package/type/boundary/build gates pass at roots recorded in the report. Retained variant passes typecheck `.cartulary/test-results/20260801T035108Z-p2587065`, boundary `.cartulary/test-results/20260801T035119Z-p2587649`, build `.cartulary/test-results/20260801T035124Z-p2588268`, `package.ui` 8/8 `.cartulary/test-results/20260801T035203Z-p2596182`, `package.protocol_ts` 5/5 `.cartulary/test-results/20260801T035216Z-p2597028`, measurement `.cartulary/test-results/20260801T035220Z-p2597423`, and startup smoke 1/1 `.cartulary/test-results/20260801T035258Z-p2615951`. | Temporary-candidate typecheck `.cartulary/test-results/20260801T034550Z-p2563318` exposed an insufficient module-scope narrowing, and boundary root `.cartulary/test-results/20260801T034716Z-p2572711` required explicit authorization while the candidate existed. The first A artifact included the temporary root re-export, so it was rejected as non-canonical; the retained final comparison rebuilt A after complete candidate removal. Rollback is already applied: variant A is the current source. Next: S-07 token-generator hardening. |
| S-07 | COMPLETE | Added the shared strict JSON parser at `tools/harness/contract/strict-json.mjs` and preserved the test-catalog facade behavior; hardened the design-token schema, loader, renderer, CLI, facade, and harness tests. Regenerated topology and protected `packages/ui-contracts/src/generated/design-tokens.ts` only through `make generate`. The source registry bytes are unchanged. | The flat projection now enforces exact schema/owner/verification/default-theme values, ten closed namespaces, safe resolved scalar syntax, no reference syntax or cycles, fatal UTF-8 and duplicate-key parsing, Unicode code-point ordering, exact-input SHA-256, generator `cartulary.design_token_generation.v2`, documentation-path rejection before reads, and exclusive sibling-temporary atomic replacement. The only generated diff is the v2/provenance header; the 18,102-byte runtime body remains SHA-256 `9778ab7dd448737da2c7a7369d1bffab6e1d1005a22fc972e3a4986d7204922a`, so no browser row is selected. | Report `.cartulary/test-results/20260801T040500Z-ui-contracts-s07/cartulary.design_token_generation_remediation.v1.json`, SHA-256 `b9e1ce51ae3ee62d52b3540474736c777de67ddb2a267ce21a1a6776ac656430`. `make harness-contract` PASS 102/102 at `.cartulary/test-results/20260801T040703Z-p2656884`; repeated `make generate` PASS at `.cartulary/test-results/20260801T040731Z-p2658158`; drift `.cartulary/test-results/20260801T040738Z-p2660345`; script lint `.cartulary/test-results/20260801T040750Z-p2664377`; `package.ui` 8/8 `.cartulary/test-results/20260801T040811Z-p2665052`; protected policy `.cartulary/test-results/20260801T040824Z-p2665878`; JSON shape `.cartulary/test-results/20260801T040826Z-p2666222`; typecheck `.cartulary/test-results/20260801T040829Z-p2666793`; boundary `.cartulary/test-results/20260801T040841Z-p2667425`. Tests cover duplicate/schema/owner/verification/theme/namespace/reference/cycle/unsafe/control/path/collision failures, no-docs CLI operation, determinism, provenance, and prior-output preservation. | JSON shape root `.cartulary/test-results/20260801T040328Z-p2642980` rejected literal restricted-root fixture paths; the final test exercises the pure pre-read guard without touching that tree. Root `.cartulary/test-results/20260801T040417Z-p2643796` then correctly required topology regeneration. Rollback is the authored strict-parser/schema/generator/CLI/facade/test unit followed by `make generate`; never edit protected output directly. Next: S-08 accumulated validation and handoff. |
| S-08 | BLOCKED: retained-run timing health | Final compiler report records 235 `keep` exports and 145 importing files (83 app source, 50 E2E, 12 package), with zero namespace/dynamic/default/unknown/non-workspace consumers. Retained final impact report accounts for all 44 coordinated removals, authorized scope, change classes, exact browser rows, commands, omissions, rollback units, staging preservation, and failures. No product source changed during S-08. | Every functional and repository criterion passes at one source state. Six final owner slices pass, including the five exact identity rows. Final harness/generated/frontend/static/build gates pass. Four independent `make check` runs each pass 197/197 work units and 853 tests with zero failures/missing evidence. Pre-check `agent-finalize` passes with retained maintenance skipped because `RESULTS_DIR` is unset. The latest exact retained-run finalization remains timing-only failed, so UI-RF-CLOSE-013 and overall completion cannot be marked pass. | API report `.cartulary/test-results/20260801T041100Z-ui-contracts-s08/cartulary.ui_contracts_api_baseline.v1.json`, SHA-256 `1f7680aae26ba5c03dbe57dd38d854b46ef6747ab7632293d40cea21ab062589`. Final impact/validation report in the same root, SHA-256 `d263c04eb3d808fa695e2114f0bbcf69df81020e4c7726db0a6f0b68c188f042`. Focused roots: `package.ui` `.cartulary/test-results/20260801T041224Z-p2670655`; `web.workbook` `.cartulary/test-results/20260801T041236Z-p2671501`; `module.workbook` `.cartulary/test-results/20260801T041358Z-p2676708`; saved views `.cartulary/test-results/20260801T041712Z-p2716012`; collaboration `.cartulary/test-results/20260801T041848Z-p2744073`; Network Flow `.cartulary/test-results/20260801T042114Z-p2773573`. Static/build roots span `.cartulary/test-results/20260801T042347Z-p2804405` through `.cartulary/test-results/20260801T042531Z-p2815620` and are enumerated in the report. Latest broad check PASS is `.cartulary/test-results/20260801T043838Z-p3181827`; pre-check finalization PASS is `.cartulary/test-results/20260801T042610Z-p2819236`. Final Markdown checkpoint is assigned `.cartulary/test-results/20260801T044500Z-ui-contracts-s08-final`; worktree and cached diff hygiene pass. | Exact latest-root finalization `.cartulary/test-results/20260801T044230Z-p3289247` fails only `duration_baseline_drift`: server/operator/harness build readiness is 16,599/15,905/15,759 ms against 15,000 ms and service-backed duration is 161,900 ms against 155,000 ms. Three prior checks also pass functionally but reproduce timing pressure. Make-owned artifact prewarm passed and did not clear it; thresholds and unrelated harness policy were not changed. Source rollback is not indicated. Next: on a host meeting current thresholds, run a latest `make check`, then `make agent-finalize RESULTS_DIR=<that exact root>` and change only S-08/UI-T-015/UI-RF-CLOSE-013 to pass. |

Canonical commands were discovered through `make help`,
`make task-guide ROLE=module-author OWNER=package.ui`,
`make explain-test-owner OWNER=package.ui`, and targeted `make explain-target`
queries. The current `package.ui` owner has eight work units and zero
service-backed rows.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=package.ui`, then `make frontend-unit` | Package-owner unit/static rows followed by broader frontend unit consumers | yes | Run the owner slice before movement to establish a current-commit baseline; complete direct characterization before extraction. |
| integration | N/A | `package.ui` currently has zero service-backed test rows | no | Re-run owner explanation before implementation. Add integration only if an in-scope change introduces a real service boundary. |
| e2e/browser | Exact rows selected through Section 4; stage is one or more of `make browser-e2e`, `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`, `make browser-e2e-measurement`, `make browser-e2e-a11y`, or `make browser-e2e-visual` | Observable runtime changes only | no for a proven private no-op partition; yes otherwise | Selection MUST be deterministic by change class and exact catalog row. A runtime change with zero rows fails closed. A retained impact record MUST explain every selected and omitted stage. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated token/protocol drift, protected-root policy, and authored JSON shape | yes | Required before and after any authorized contract/generator-adjacent slice; generated files are never hand-edited. |
| API/boundary/static | Current-commit API/export report; `make frontend-typecheck`; `make frontend-import-boundary-check` | Root export compatibility, type-only consumer accounting, package/vendor/generated-token import rules | yes | Report mechanism is intentionally unspecified. A new dependency needs authorization. Every source extraction requires all three evidence classes. |
| web build | `make build-web` | Vite build, module graph, root facade evaluation, and bundle inputs | yes for source movement | Required for S-03 and any import-seam study. Build success alone does not prove selector or startup equivalence. |
| full check | `make agent-finalize`, followed by `make check` | Final retained-run maintenance and repository-wide gate | no | Required at S-08. Supply a successful `RESULTS_DIR` when retaining a full warm run; otherwise report the unset skip honestly. |
| tracker checkpoint | `make lint-markdown`; `git diff --check` | Documentation and patch hygiene after every slice | yes | The tracker is updated and validated before the next workstream begins. |

### Recorded discovery baseline

These commands actually ran and passed during the discovery session at commit
`24f3010158934a61bed02804fc97093ef625ec56`. They are retained historical
evidence, not a claim that the same commands ran at tracker-creation commit
`e2b57a...` or current planning commit `9f0a22c...`.

| Command | Result | Evidence |
| --- | --- | --- |
| `make test-slice OWNER=package.ui` | PASS, 6/6 work units | `.cartulary/test-results/20260731T015525Z-p3878389` |
| `make frontend-typecheck` | PASS, exit 0 | Command produced no failure output. |
| `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260731T015533Z-p3878916` |
| `make generate-drift` | PASS | `.cartulary/test-results/20260731T015537Z-p3879377` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260731T015603Z-p3883984` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260731T015625Z-p3885091` |

`make check`, `make frontend-unit`, and browser targets did not run during that
discovery baseline. They must not be reported as passing.

### Current implementation baseline and slice evidence

| Command | Result | Evidence |
| --- | --- | --- |
| `make lint-markdown` | PASS at live commit `e2b57a73bf7388a35404fa9b74f73c317d3e2cd8` after tracker creation | `.cartulary/test-results/20260731T021608Z-p3895031` |
| `make lint-markdown` | PASS after the current tracker revision at `9f0a22c6f5d8016061a074dc0d0b969b09609cc8` | `.cartulary/test-results/20260801T015948Z-p2143397` |
| `git diff --check HEAD -- docs/handoffs/ui-contracts-module-refactor-tracker.md` | PASS, exit 0 with no output | Tracker-only whitespace audit. |
| `git diff --name-only --cached` before S-00 | Existing staged user change | Only `docs/handoffs/ui-contracts-module-refactor-tracker.md`; implementation preserves and amends it without deliberate staging changes. |
| `make frontend-fallow-static` before S-00 | PASS | `.cartulary/test-results/20260801T021135Z-p2150864`; ignored retained evidence only. |

Slice-specific evidence is appended here and in the handoff log after each
workstream. Historical planning roots do not prove the implementing source
state.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| UI-T-001 | Normalize `packages/ui-contracts` to the safe label `ui-contracts` and constrain the remediation scope | WF-00 | DONE | none | Section 1 scope posture | Target, output, allowed write, and non-goals are explicit. |
| UI-T-002 | Inventory every target file and live inbound dependency | WF-01 | DONE | UI-T-001 | Section 2; current file and import analysis | Six tracked files, two ignored artifacts, 151 importing files, 279 root names, and the two newer coordination selectors are recorded at `9f0a22c...`. |
| UI-T-003 | Map selector, shell, grid, view-schema, Network Flow, token, and harness owners | WF-02 | DONE | UI-T-002 | Sections 3 and 4 | Each discovered responsibility has an owner candidate and keep/split/defer posture. |
| UI-T-004 | Freeze observable package and UI contract behavior | WF-02 | DONE | UI-T-003 | Section 4 | Root API, bytes, failures, order, labels, DOM attributes, tokens, CSS, and density results are named. |
| UI-T-005 | Identify characterization gaps for high-risk public families | WF-03 | DONE | UI-T-003 | Section 4 test postures and S-01 | Missing direct evidence is enumerated without modifying tests. |
| UI-T-006 | Repair the token support-owner contradiction and adopt machine compatibility/browser-impact inputs | WF-00 | DONE | RB-001 | S-00 checkpoint ledger and retained roots | Design and Testing Harness owners name the resolved JSON projection, forbid executable docs access, and support gates pass. |
| UI-T-007 | Add complete direct package characterization | WF-01 | DONE | UI-T-006 | S-01 checkpoint ledger; package slice root `.cartulary/test-results/20260801T024543Z-p2186360` | Every family selected for movement has exact positive, negative, and ordering evidence. |
| UI-T-008 | Create the current-commit root API/export-ownership baseline | WF-02 | DONE | UI-T-007 | S-02 report and digest in the checkpoint ledger | All 279 exports have TypeScript-program consumers, effects, owners, disposition, compatibility, and browser rows; forbidden consumer classes are zero. |
| UI-T-009 | Partition the monolithic authored source behind the unchanged root facade | WF-03 | DONE | UI-T-008 | S-03 report, graph check, and retained gate roots | Root surface, initialization effects, and every observable output remain compatible across ten acyclic private modules. |
| UI-T-010 | Migrate approved application controller state and clean the root surface | WF-04 | DONE | UI-T-009, RB-002 | S-04 report, compiler/static gates, and checkpoint ledger | Root API has 236 names; app state has one owner; no reverse dependency, forwarding alias, or accessibility drift exists. |
| UI-T-011 | Migrate semantic identity to generated `SheetRef`/registry-validated view-schema IDs | WF-05 | DONE | UI-T-010, RB-003 | S-05 API report, six focused owners, and five exact browser rows | Root API has 235 names; no semantic `WorkbookSurface`, custom valid variant, registry duplicate, auth inference, old path, or cycle remains. |
| UI-T-012 | Inspect the protocol registry seam and retain the current import by default | WF-06 | DONE | UI-T-011, RB-004 | S-06 decision report and bundle/cold-start digests | Same-toolchain A/B comparison found equivalence but no side-effect, eager-source, request, gzip, or p95 threshold; the candidate was removed and the root import remains. |
| UI-T-013 | Harden the generated-token seam | WF-07 | DONE | UI-T-012, RB-001 | S-07 report and retained Make roots | Strict duplicate/schema/scalar/reference validation, code-point determinism, exact-byte provenance, no-docs execution, and failure-atomic replacement pass; runtime exports remain byte-identical. |
| UI-T-014 | Apply deterministic browser-impact accounting to every observable implementation change | WF-08 | DONE | Applicable future slice, RB-006 | S-08 final impact report | All 44 removed exports, change classes, exact identity rows, output-neutral omissions, commands, roots, and rollback units are retained; final owner slices include every selected row. |
| UI-T-015 | Run final implementation verification and update handoff | WF-08 | BLOCKED | Applicable future slices | S-08 API/impact reports, Make roots, and retained-finalization failure | All product, focused, static, build, and four broad checks pass; exact retained-run finalization reproducibly misses timing-health thresholds. |
| UI-T-016 | Maintain the restartable tracker through every implementation slice | WF-08 | DONE | UI-T-001 through UI-T-015 | This file and per-slice evidence rows | S-00 through S-07 are complete; S-08 records the single exact blocker, evidence, rollback posture, and one-command continuation without rediscovery. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Planning complete at live commit `e2b57a...`; repository was clean before tracker creation; prior discovery was at `24f301...`. | Inspected framework, `domain.md`, Core 00-04, adopted Testing Harness and Network Flow NLSpecs, design and frontend guides; touched only this tracker. | `sed`; `rg`; `find`; `git rev-parse`; `git status`; `git log`; `git diff`; `wc`; `date`. | Target exists, normalized label is safe, target unchanged across commits, live caller inventory refreshed. | RB-001 affects generator redesign only. | Obtain later authorization for S-01 characterization. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | Planning rebased to clean `main` commit `9f0a22c...`; prior commits and run roots retained as historical evidence. | Inspected this tracker, analysis notes, NLSpec writing guide, current target/consumer/protocol/token evidence; touched only this tracker. | `git status`; `git rev-parse`; `git ls-files`; `rg`; `sed`; `wc`; TypeScript AST/import analysis. | Authority, normative terms, closed risk states, current counts, and later-authorization boundary are explicit. | RB-001 blocks generator redesign only. | Complete tracker-only validation, then seek separate authorization for S-00/S-01/S-02 as applicable. |
| 2026-07-31 22:35 EDT | Codex / S-00 | Specification and tracker admission complete at implementation baseline `9f0a22c...`; the user's pre-existing staged tracker change is preserved and amended. | Updated the design owner, Testing Harness owner row, frontend boundary guide, tracker policy/ledger, and stale harness catalog-count guard. | `make lint-markdown`; `make harness-contract`; `make json-shape-check`; `make generated-artifact-policy-check`; `make generate-drift`; `git diff --check`. | RB-001 is resolved without runtime or generated drift; retained roots and the recovered initial harness failure are in Section 8. | None. | Begin S-01 test partition; production source remains unchanged. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Backend refactor is out of scope; no backend responsibility was found in the target. | Inspected all tracked target files and searched imports/symbols; touched only this tracker. | Target file inspection and repository searches. | No route handler, WebSocket transport, SQL/storage, authorization decision, mutation coordinator, revision writer, or projection refresh implementation found. | None for planning. | Keep backend workflows excluded unless new live evidence appears. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | Backend boundary remains out of scope and unchanged at `9f0a22c...`. | Rechecked target dependencies and observable-contract posture; touched only this tracker. | Current source/import inspection. | The tracker now states that UI anchors do not transfer route, socket, storage, auth, mutation, revision, or projection ownership. | None for tracker acceptance. | Keep those behaviors with adopted/backend owners during every future slice. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Legitimate shared selector facade; mixed responsibilities require private partition and later owner decisions. | Inspected `src/index.ts`, representative app-shell, grid-adapter, test-utils, runtime, unit, and E2E callers; touched only this tracker. | `rg -l '@cartulary/ui-contracts' apps packages`; source inspection; export and line counts. | 155 live caller files; no grid-vendor import in target; shell/grid/view/Network Flow seams mapped. | RB-002 through RB-005 defer public ownership/API changes. | Execute S-01 before any source movement. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | Current facade has 151 importing files and 279 root names; shell and identity decisions are closed without source movement. | Reinspected target API, app shell/identity callers, generated `SheetRef`, and representative package consumers; touched only this tracker. | Current TypeScript AST/import analysis; source inspection. | RB-002 is implementation-gated; RB-003 defines canonical identity; no package-to-app back edge or selector-byte change is authorized. | Later implementation authorization and export baseline. | Characterize, baseline, then partition privately before any state or identity migration. |
| 2026-07-31 22:55 EDT | Codex / S-02 | Declaration-accurate baseline is complete at 279 exports and 154 importing files after test partition. | Inspected all workspace TypeScript import forms and root declarations through TypeScript `6.0.2`; retained the report and digest under the S-02 run root. | Compiler-program analyzer; `sha256sum -c`; `make generate`; `make test-slice OWNER=package.ui`; `make frontend-typecheck`; `make generate-drift`. | 43 names have no continuing consumer; dispositions are 36 internalize, five remove, three move, and 235 keep. Forbidden namespace/dynamic/default/unknown/non-workspace consumers are zero. | None. | Use the 279-row report as S-03's pre/post equality gate; do not combine source movement with cleanup. |
| 2026-07-31 23:14 EDT | Codex / S-04 | App controller/transient vocabulary has one owner in `apps/web`; the package root contains 236 exports. | Migrated seven app/E2E source files plus the app type owner; removed the approved package root names and four obsolete helpers; adjusted root-facade characterization only for removed APIs. | TypeScript API report; removed-name/import/catalog scan; package/frontend unit, type, boundary, Fallow, build, generated-drift, Biome, Markdown, and diff gates. | All retained runtime outputs are exact; compiler/static evidence passes; no observable browser change exists, so browser omission is required by the impact matrix. | None. | Begin S-05 as one identity migration; do not add an old-path forwarding file. |
| 2026-07-31 23:39 EDT | Codex / S-05 | Canonical identity migration is complete; the package root contains 235 exports and generated `SheetRef` is the only valid workbook identity type. | Migrated app/test/E2E identity callers, renamed and authorized the single app helper facade, deleted the old path, removed `WorkbookSurface`, updated ownership policy and the frontend guide, and retained view-schema registry validation. | API/source audits; six focused unit owners; five exact service-backed rows; frontend unit/type/boundary/Fallow/build; generated drift; JSON shape; Biome; Markdown/diff hygiene. | Identity variants remain distinct and all selected startup, saved-view, presence, claimed-extension, and unclaimed-extension evidence passes. | Two concurrent/combined Network Flow runs had harness-only accounting failures; isolated reruns pass. | Begin S-06 with a read-only final-bundle graph inspection; retain the root registry import absent a static trigger. |
| 2026-07-31 23:54 EDT | Codex / S-06 | Protocol root import retained; no narrow subpath remains. | Inspected Vite output/source maps; temporarily tested the sole authorized registry subpath; retained A/B bundles, graph/request/gzip data, 20 cold starts per variant, and the final keep report; restored variant A before final gates. | `make build-web`; package UI/protocol slices; frontend type/boundary; `make browser-e2e-measurement`; exact workbook-startup smoke; controlled Playwright cold-start runner. | Candidate behavior is equivalent but retains all unrelated eager sources and requests, adds 238 gzip bytes, and is below the p95 threshold. RB-004 and UI-RF-CLOSE-009 pass with `keep`. | None; temporary-candidate compile/boundary failures and the rejected preliminary A artifact are recorded in the S-06 ledger. | Begin S-07 from the restored root-import source; do not recreate the measured subpath. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Generated design tokens remain protected and accessed through the root facade. | Inspected generated token output, token registry, generator, generated-artifact policy, import-boundary policy, and verification owner inputs; touched only this tracker. | Discovery baseline: `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`. | All three passed at commit `24f301...` with run roots recorded in Section 8; not rerun at `e2b57a...`. | RB-001: `BLOCKED: owner contradiction`. | Preserve live JSON-to-generated path; repair adopted owner text before redesign. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | Live JSON-to-generated path remains authoritative in execution; generated output remains protected and unchanged. | Reinspected token registry, helper, policy, generated facade, and contradictory owner row; touched only this tracker. | Read-only source/search inspection; no generation command. | Machine contract now closes schema, ordering, references, atomicity, provenance, determinism, and no-docs evidence. Historical drift roots remain tied to `24f301...`. | RB-001: `BLOCKED: owner contradiction` for redesign only. | Repair owner text under separate authorization before S-07; preserve current seam meanwhile. |
| 2026-08-01 00:09 EDT | Codex / S-07 | Design-token generation v2 is strict, safe, deterministic, provenance-bearing, documentation-free, and atomic. | Added the shared strict JSON contract parser; hardened token schema/helper/CLI/facade and harness tests; regenerated topology and protected output through Make. | Harness contract; repeated generation; package owner; drift/policy/JSON/script/type/boundary gates; pre/post runtime-body digest comparison. | Input is unchanged; generated runtime exports/CSS/order/densities are byte-identical; only v2 and exact input SHA-256 header lines changed. Negative and failure-atomic cases pass. | None; two expected recovery failures are recorded in the S-07 ledger. | Begin S-08 at this source state; rerun the API/caller inventory and all accumulated gates before final tracker completion. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Existing package tests cover many but not all high-risk public families; owner has six non-service-backed rows. | Inspected both target tests, `package.ui` family/verification inputs, catalog owner, helper ownership, and representative consumer tests; touched only this tracker. | `make help`; `make task-guide ROLE=module-author OWNER=package.ui`; `make explain-test-owner OWNER=package.ui`; targeted `make explain-target`; discovery `make test-slice OWNER=package.ui`. | Discovery owner slice passed 6/6 at `24f301...`; run root `.cartulary/test-results/20260731T015525Z-p3878389`. | Missing direct characterization blocks safe extraction until S-01. | Add S-01 tests in a later authorized task, then rerun current-commit owner and frontend unit targets. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | Test counts remain 23 selector cases and three token cases; deterministic browser-impact accounting replaces discretionary target choice. | Rechecked target tests, package owner/family inputs, Make targets, and consumer distribution; touched only this tracker. | Read-only catalog/Make/source inspection; no product test command. | S-01 has closed direct-characterization obligations; RB-006 requires exact rows and fails closed for a zero-row runtime change. | Missing characterization blocks source movement, not tracker acceptance. | Under authorization, run S-01, then create the S-02 TypeScript-program API baseline. |
| 2026-07-31 22:49 EDT | Codex / S-01 | Selector characterization is partitioned into five root-facade suites; `package.ui` has nine independently accounted rows. | Replaced the monolithic test, updated the authored family manifest and exact catalog guard, and regenerated the topology render index. | `make generate`; `make test-slice OWNER=package.ui`; `make frontend-unit`; `make frontend-typecheck`; `make generate-drift`; `make harness-contract`; `make lint-biome`. | All final gates pass; exact roots and recovered failures are in the S-01 ledger. Production and generated token output are unchanged. | None. | Begin S-02 and refuse source movement until all 279 exports have one declaration-accurate disposition. |
| 2026-08-01 00:43 EDT | Codex / S-08 | Implementation is functionally complete; retained-run timing health is the sole closure blocker. | Regenerated the final TypeScript API/caller inventory; retained the complete impact/validation report; inspected every exact identity row and final diff/staging scope; changed only evidence and this tracker in S-08. | Six focused owner slices; harness/generated/frontend/static/build gates; pre-check finalization; four independent `make check` runs; exact latest-root retained finalization; Make-owned build cache prewarm; Markdown and diff hygiene. | API is 235 `keep` names with 145 importers and zero forbidden consumers. All focused/static/build checks pass. Every broad check passes 197/197 and 853 tests with zero failed/missing evidence. | Latest retained finalization fails fixed timing health only at `.cartulary/test-results/20260801T044230Z-p3289247`; source rollback is not indicated. | On a qualifying host, run a latest `make check`, then `make agent-finalize RESULTS_DIR=<exact root>`; update S-08, UI-T-015, and UI-RF-CLOSE-013 only. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Target exports auth/account/admin UI anchors but implements no authorization policy or decision. | Inspected target selector families, representative auth/admin callers, Core 04 relevant sections, and selector policy test; touched only this tracker. | Source and caller searches. | Security/authorization outcomes remain owned outside this package; exact UI anchors remain frozen. | None for selector-only planning. | Do not move authorization logic into this package; run affected browser owner tests only if anchors change under later authorization. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | Authorization remains server/owner authoritative; identity and registry membership are explicitly non-authorizing. | Rechecked identity types, selector validation, and Core/security posture; touched only this tracker. | Read-only source inspection. | Custom identity variants, registry duplication, and client-side authorization inference are prohibited; no Core amendment is currently required. | None for tracker acceptance. | Preserve authorization boundaries and select stateful evidence only for an authorized server-backed observable change. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:16 EDT | Codex / tracker implementation | Tracker created; planning workflows are complete; implementation is not authorized. | Touched only this tracker; inspected live status and all evidence named above. | `make lint-markdown`. | PASS; run root `.cartulary/test-results/20260731T021608Z-p3895031`. Safe next slice is characterization, not source movement. | RB-001 blocked; RB-002 through RB-005 deferred. | Run `make task-guide ROLE=module-author OWNER=package.ui`, refresh caller inventory, then implement S-01 only under later authorization. |
| 2026-07-31 21:56 EDT | Codex / NLSpec tracker revision | NLSpec-style closure contracts are complete for planning; future implementation remains unauthorized. | Touched only this tracker; `temp/analysis-notes.md` and all implementation/owner/generated inputs remain unchanged. | `make lint-markdown`; tracker `git diff --check`; staged and unstaged name audits. | Final Markdown lint passed at `.cartulary/test-results/20260801T015948Z-p2143397`; whitespace audit passed; staged audit was empty; unstaged audit listed only this tracker. | RB-001 blocks token redesign; all source work requires later authorization. | Under separate authorization, begin S-00 only for support-owner repair or S-01/S-02 for characterization and API baselining; do not move source first. |

## 11. Open Questions and Blockers

There are no remaining open-ended planning questions. The following closure
decisions are normative. `RESOLVED` records a decision; the named slice and
evidence still determine whether implementation is complete.

| ID | Closure decision | Implementation prerequisite | Binary evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | `docs/design.md` owns the human YAML registry; `contracts/design/tokens.v1.json` is its already-resolved executable projection. Executable tooling MUST NOT access `docs/`. | Complete S-00 owner alignment before S-07. | Owner text and machine projection name the JSON input and forbid docs access; S-07 no-docs, negative, determinism, provenance, and atomicity evidence passes. | `IMPLEMENTED` by S-00 owner repair and S-07 generator hardening. |
| RB-002 | `apps/web` owns controller/navigation state, transient load state, and private copy. `package.ui` retains selector bytes, semantic DOM contracts, stable test IDs, and intentionally shared accessibility names. | Complete S-01 characterization and S-02 export matrix. | Every moved declaration is matrix-classified; no forwarding alias or package-to-app edge exists; state has one owner; exact static/unit/build/browser/a11y evidence passes. | `IMPLEMENTED` by S-04; no browser/a11y row was selected because observable output is unchanged. |
| RB-003 | Generated `SheetRef` is canonical workbook identity. Registry-validated view-schema IDs, saved-view identity, and the extension variant retain distinct semantics. `WorkbookSurface` is obsolete. | S-02 caller matrix, S-03 partition, and S-04 cleanup. | Required identity mapping holds; no old path, custom variant, registry duplicate, authorization inference, selector drift, or package cycle exists. | `IMPLEMENTED` by S-05 with exact unit, static, build, and browser evidence. |
| RB-004 | Keep the current protocol root-registry import unless the post-refactor graph shows an eager unrelated module, cycle, or critical-request trigger and the controlled candidate crosses a Section 4 threshold. | S-06 graph inspection; controlled comparison only if triggered. | Same-source-state report, graph/request data, and, only if tested, at least 20 comparable runs, equivalence, and one qualifying threshold. | `IMPLEMENTED: keep` by S-06; the sole candidate crossed no threshold and was removed. |
| RB-005 | Valuable observable contracts and active imports are stable. Obsolete private TypeScript structure defaults to coordinated removal after complete accounting. | S-02 TypeScript-program matrix and the approved S-04/S-05 dispositions. | All 279 baseline names are accounted; removal candidates have zero program/catalog/generated/dynamic consumers and exact impact evidence; no compatibility aliases remain. | `RESOLVED: coordinated cleanup`. |
| RB-006 | Browser evidence is selected deterministically from observable change classes and exact catalog rows. A runtime change mapping to zero rows fails closed. | Actual authorized diff plus current catalog/target explanation. | Retained impact record lists changed exports, classes, exact row IDs, stages, projects, omissions/reasons, commit, commands, and run roots; every required row passes. | `RESOLVED`. |

No current decision requires a Core amendment. An adopted-owner amendment
becomes mandatory only if characterization contradicts an owner or the effort
seeks new product behavior. All migrations remain incomplete until their slice
and final closure criteria pass.

## 12. Binary Completion Criteria

### Tracker acceptance

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `packages/ui-contracts` is inventoried or explicitly out of scope. | PASS | Section 2 accounts for all six tracked files plus ignored `node_modules/**` and `tsconfig.tsbuildinfo`. |
| Every discovered public contract risk has an owner and test posture. | PASS | Sections 3 and 4 map selectors, validation, shell labels/types, grid DOM, view schemas, Network Flow, tokens, indirect backend surfaces, and harness accounting. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 provides predecessor/successor relationships, validation, and handoff checkpoints for WF-00 through WF-08. |
| Every proposed implementation slice states dependency, exact change, preserved contracts, tests, command, rollback, completion, and authorization. | PASS | Section 7 defines S-00 through S-08; behavior changes require authorization beyond the structural refactor. |
| Validation commands and deterministic browser-selection rules are defined. | PASS | Sections 4 and 8 map every change class to Make-owned stages and require exact catalog rows. |
| Token authority contradiction has one durable owner/projection resolution. | PASS | S-00 aligns `docs/design.md`, the Testing Harness owner row, the frontend guide, and RB-001; focused gates pass at the retained roots in Section 8. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Sections 1, 3, and 5 record the narrow framework description versus the mixed live package, commit/API/caller changes, and the live machine-input path. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records scope, owners, target posture, commands, evidence roots, blockers, and the safe next slice. |
| Closed planning decisions distinguish resolved, gated, deferred, and blocked work. | PASS | Section 11 closes RB-001 through RB-006 and states binary prerequisites/evidence. |
| Protected generated files are changed only through owner inputs and `make generate`. | PASS | S-07 changed only the protected token header through repeated `make generate`; generated policy and drift gates pass at the roots in the S-07 ledger. |
| Tracker Markdown and patch hygiene pass. | PASS | Completed S-08 tracker is assigned final lint root `.cartulary/test-results/20260801T044500Z-ui-contracts-s08-final`; worktree and cached `git diff --check` pass without altering the user's staged-file scope. |
| Sequential implementation is expressly authorized and hard-gated by tracker checkpoints. | PASS | Section 1 and Section 7 record the authorized S-00 through S-08 workflow. |

### Future implementation Definition of Done

Every row is binary. `NOT PASSED` means implementation is incomplete; it does
not invalidate a complete planning tracker.

| ID | Binary criterion | Traceability | Current state |
| --- | --- | --- | --- |
| UI-RF-CLOSE-001 | One implementing source state records the current target inventory, exact caller distribution, and complete root API baseline. | Section 2; S-02; UI-T-002/UI-T-008 | PASS: final S-08 report records 235 exports and 145 importing files split 83 app source, 50 E2E, and 12 package files. |
| UI-RF-CLOSE-002 | Every public family moved by the refactor has direct positive, invalid-input, and order characterization at the implementing commit. | Section 4; `must_fix` finding; S-01 | PASS: final `package.ui` 8/8 and all five broader feature owners pass at the roots in the S-08 ledger. |
| UI-RF-CLOSE-003 | The root import path and every baseline export name, declaration kind, signature, type, generated status, and initialization effect are unchanged unless separately authorized. | Root API contract; S-02/S-03 | PASS: final report contains 235 authorized `keep` dispositions, no added exports, and exactly the 44 approved S-04/S-05 removals. |
| UI-RF-CLOSE-004 | Exact selector/test-ID bytes, percent/CSS escaping, validation failures, arrays/order, shared labels, semantic grid attributes, token exports, theme CSS, and density-derived row heights match the baseline. | Freeze map; S-01/S-03 | PASS: final package/feature owners, frontend suite, four broad checks, and identical token runtime-body digest confirm accumulated compatibility. |
| UI-RF-CLOSE-005 | Private partitioning introduces no cycle, evaluation-order drift, eager side effect, or generated-token initialization drift. | WF-03; S-03 | PASS: graph, API, generated-byte, package, frontend, Fallow, and build evidence are retained in the S-03 ledger. |
| UI-RF-CLOSE-006 | Every root export has one complete export-ownership row and an authorized disposition; no declaration is moved on filename or text-search inference. | Export-ownership contract; RB-002/RB-005; S-02 | PASS: final compiler report contains one `keep` row for each of 235 exports and zero duplicate or forbidden consumer classes. |
| UI-RF-CLOSE-007 | S-04 leaves one `apps/web` state owner, no package-to-app back edge or forwarding aliases, a 236-name root API, and no copy, selector, ARIA, keyboard, focus, route, or shell-composition drift. | RB-002; WF-04; S-04 | PASS: S-04 compiler report, source audit, package/frontend tests, boundary/Fallow checks, and build evidence are retained in the checkpoint ledger. |
| UI-RF-CLOSE-008 | Semantic callers use generated `SheetRef`, registry-validated view-schema strings, saved-view identity, or extension identity as mapped; `WorkbookSurface` and the old app helper path are absent. | Identity contract; RB-003; S-05 | PASS: S-05 API/source audits, six focused owners, and five exact browser rows are retained. |
| UI-RF-CLOSE-009 | A protocol-import keep/change decision uses a conformant retained comparison; a narrow seam is adopted only after equivalence and one materiality threshold pass. | RB-004 comparison contract; S-06 | PASS: the retained report contains static graph/request/gzip evidence, exact source/toolchain controls, equivalence, and 20 cold starts per variant; no threshold passed, so variant A is restored and the candidate is absent. |
| UI-RF-CLOSE-010 | Testing Harness owner text is repaired and token generation is documentation-free, deterministic, safe, atomic on failure, provenance-bearing, and runtime-compatible for unchanged input. | Design-token machine contract; RB-001; S-00/S-07 | PASS: S-00 owner text and S-07 strict loader/renderer/CLI tests, repeated generation, provenance/runtime digests, protected policy, and no-docs/atomic failure evidence are retained. |
| UI-RF-CLOSE-011 | Every removed export passes the complete program/catalog/generated/dynamic-consumer gate; obsolete private surfaces and compatibility aliases are absent. | Compatibility classes; RB-005; S-02/S-04/S-05 | PASS: S-08 compiler delta accounts for all 36 internalized, three moved, and five removed names; final forbidden-consumer counts are zero. |
| UI-RF-CLOSE-012 | Every observable runtime change has a retained deterministic impact record and all exact selected browser rows pass; a zero-row mapping fails closed. | Browser mapping; RB-006; S-04 through S-08 | PASS: S-08 report maps every change class and omission; the five exact identity rows are present in passing workbook/saved-view/collaboration/Network Flow owner roots. |
| UI-RF-CLOSE-013 | All applicable Make-owned package, frontend unit/type/boundary, build, generated, browser-impact, `agent-finalize`, and broader final gates pass at the implementing commit. | Section 8; WF-08; S-08 | BLOCKED: all product and four broad checks pass, and pre-check finalization passes; exact latest-root finalization fails only the fixed scheduler timing-health gate at `.cartulary/test-results/20260801T044230Z-p3289247`. |
| UI-RF-CLOSE-014 | The handoff records the authorized-file scope, current commit, API/caller reports, impact/measurement artifacts, rollback points, commands, run roots, failures, and justified omissions. | Sections 9-10; WF-08; S-08 | PASS: S-08 ledger, session handoff, and retained final impact report record all required scope, evidence, failures, rollback units, staging state, and exact continuation. |

The planning tracker is complete only when every tracker-acceptance row passes.
Future implementation is complete only when all fourteen `UI-RF-CLOSE-*`
rows pass at the same implementing commit. A separately authorized product
change MAY replace an exact-compatibility criterion only when its adopted owner,
new expected behavior, migration, and evidence are recorded explicitly.

## Appendix A. Non-Authoritative Supporting Evidence

This appendix records corroboration only. Research, framework guidance, and
external documentation MUST NOT supersede adopted owners, typed projections,
or current repository evidence.

| Evidence | Decision corroborated | Authority limit |
| --- | --- | --- |
| Analysis R04 and its identity/representation research | `SheetRef` is semantic identity; selector tokens and labels are representations. | Does not define Cartulary variants, authorization, or compatibility. |
| Analysis R08 plus official Vite and Node module guidance | Root-versus-narrow imports require graph, chunk, side-effect, request, and measured startup evidence. | Does not prove that the current root import is costly or authorize a new subpath. |
| Analysis R09 plus API-reporting and TypeScript-program guidance | Public root exports need declaration-accurate accounting, including type-only and re-export consumers. | Does not choose a reporting dependency or authorize export removal. |
| Official Playwright project/selection guidance | Browser evidence can be selected deterministically by catalog row, stage, and project. | The adopted Testing Harness and live catalog, not Playwright documentation, own Cartulary routing. |
| React state and composition guidance | Application controller/transient state belongs near the app owner; shared immutable contracts may remain package-owned. | Does not independently move any current export or define shell behavior. |
| OWASP authorization guidance | Client identity and registry membership cannot be authorization authority. | Core 04 and the server remain authoritative for Cartulary outcomes. |
| Grid accessibility and adapter guidance | Stable semantic DOM contracts may be shared while vendor lifecycle remains in the grid adapter. | Does not amend Cartulary ARIA, selector, density, or vendor-boundary requirements. |
