# ui-contracts Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | `packages/ui-contracts` |
| Normalized target label | `ui-contracts` |
| Output path | `docs/handoffs/ui-contracts-module-refactor-tracker.md` |
| Mode | Planning and documentation only |
| Allowed change in this session | This tracker file only |
| Non-goals | No production, test, contract, generated, package-configuration, migration, harness, or dependency-lock changes; no selector, type, token, route, storage, authorization, mutation, revision, projection, or browser behavior changes |
| Discovery baseline | Commit `24f3010158934a61bed02804fc97093ef625ec56`; the planning pass observed unrelated evidence/backend/generated worktree edits and did not modify them |
| Live execution baseline | Commit `e2b57a73bf7388a35404fa9b74f73c317d3e2cd8` (`main`); the worktree was clean before this tracker was created |
| Baseline mismatch | The repository advanced between discovery and tracker creation. `packages/ui-contracts` did not change, but the live `@cartulary/ui-contracts` caller search increased from 153 to 155 files. Relevant Core 01 through Core 04 Evidence requirements also changed. This tracker uses live repository state and retains the earlier validation evidence with its original commit context. |
| Later authorization | Any implementation, characterization-test, catalog, code-generation, or package-boundary change requires a later expressly authorized task. |

The target path exists. Its presence is not treated as proof that it is a valid
permanent module boundary. The live implementation is assessed below as a
legitimate shared selector facade that also contains responsibilities whose
long-term ownership requires further evidence.

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

Search results were used to enumerate callers and contract families. A search
hit alone was not treated as proof of a file's behavior.

## 2. Current-State Repository Inventory

At live commit `e2b57a73bf7388a35404fa9b74f73c317d3e2cd8`,
`packages/ui-contracts` has six tracked files. The two additional files found
under the directory are ignored tool/dependency artifacts and are explicitly
out of scope.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `packages/ui-contracts/package.json` | Declares the private workspace package and its single public root entry point. | Package `@cartulary/ui-contracts`; only `"."` is exported as `./src/index.ts`. | Workspace package resolution and manifests, including `apps/web`, `packages/grid-adapter`, and `packages/test-utils`. | `@cartulary/protocol-ts`; development-only Vitest dependency. | Indirectly exercised by all package-owned and consumer tests. | None directly. | `package.ui` | Medium | Changing the root export or adding subpaths is a public package-surface decision and is deferred. |
| `packages/ui-contracts/tsconfig.json` | Strict, source-only TypeScript configuration extending the repository base configuration. | No runtime surface; includes `src`, uses bundler resolution, and emits no files. | Repository frontend type-check tooling. | `../../tsconfig.base.json`. | `make frontend-typecheck`. | None. | `package.ui` | Low | Configuration change is not part of the planned structural split. |
| `packages/ui-contracts/src/index.ts` | Root facade for runtime-safe selectors/test IDs, validation and encoding helpers, selector vocabularies, shell labels/types, semantic grid contracts, generated-token access, density parsing, and feature-specific selector families. | 2,349 lines; 229 exported functions and 274 exported declarations, plus named generated-token re-exports. | Live search finds 155 importing app/package files: 90 under `apps/web/src`, 51 under `apps/web/e2e`, 13 under `packages`, and one other app file. | `listViewSchemaRegistryEntries` from `@cartulary/protocol-ts`; `./generated/design-tokens`. | `src/index.test.ts`, `src/design-tokens.test.ts`, consumer unit tests, grid-adapter tests, test-utils tests, and browser tests. | Reads the generated view-schema registry through `protocol-ts`; re-exports the protected generated design-token projection. | Shared selector facade: `package.ui`; embedded responsibilities have candidates in Section 3. | High | This is a legitimate facade but also a mixed-responsibility, high-fan-in monolith. No backend, transport, persistence, authorization, mutation, revision, or projection implementation was found. |
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
| Auth, account, admin, landing, incident, workbook, saved-view, timeline, collaboration, entity, evidence, coordination, import, assessment, reference-pack, and debug selector families | `src/index.ts` | `package.ui` facade, with semantic inputs governed by the matching subsystem or application owner | split | Large owner-specific type/function families coexist in one 2,349-line file and are consumed across runtime and tests. | Split means private source grouping with unchanged root re-exports, not a decision to relocate public imports. |
| Workbook shell slots, shell labels, admin panel tokens, load state, and controller-facing state types | `src/index.ts` | `apps/web` for application state; `package.ui` only where an exact automation contract must remain shared | defer | `apps/web` runtime files consume these types and labels; the frontend guide assigns shell/controllers/state to `apps/web`. | Relocation could change public types or user-facing ARIA labels. A later owner decision is required. |
| `WorkbookSurface = string` and view-schema identifier validation | `src/index.ts` | `packages/view-contracts` and/or `@cartulary/protocol-ts`, with `package.ui` consuming a narrow contract | defer | The package imports `listViewSchemaRegistryEntries` and constructs validation state from generated registry entries. | Do not replace the broad type or dependency direction without compatibility and bundle evidence. |
| Grid density parsing, row-height lookup, semantic scrollport class, role/data attributes, row/cell/group/filter/sort selectors | `src/index.ts` | Shared selector facade for automation; `packages/grid-adapter` for semantic DOM implementation and density application | split | `SemanticDataGrid.tsx`, grid-adapter tests, and test-utils consume these exports; no `react-data-grid` import exists in `ui-contracts`. | Preserve semantic selector outputs. Vendor integration remains correctly isolated in the adapter. |
| Generated design-token exports and theme CSS facade | `src/generated/design-tokens.ts` and root re-exports | Design-token owner projection plus generated-artifact harness; `package.ui` remains the authored access facade under current policy | keep | Token registry, generator, import-boundary policy, artifact policy, and design-token tests all point to this seam. | Generator/input redesign is blocked by RB-001. Generated output is never hand-edited. |
| Network Flow-specific selector types and builders | `src/index.ts` | Adopted Network Flow owner for semantics; `package.ui` for the stable cross-runtime/test automation facade | split | Network Flow runtime, unit, and E2E callers use the shared builders; the adopted Network Flow NLSpec governs the subsystem behavior. | Group privately and add direct characterization before movement. Do not infer subsystem behavior from function names. |
| Test/debug harness selectors used by production-rendered DOM | `src/index.ts` | `package.ui` automation facade; `packages/test-utils` owns choreography, not selector bytes | keep | Runtime components attach IDs and test-utils/E2E use them; the development guide explicitly makes shared test IDs a runtime-safe contract. | Production use is intentional, not evidence that test choreography belongs in production. |
| HTTP, WebSocket, storage, authorization, mutation, revision, and projection behavior | Not implemented in this package | Matching backend/platform/application owners | defer | Direct inspection found no fetch, route handler, WebSocket, SQL/storage, authorization-decision, source mutation, revision-write, or projection-refresh implementation. | The package supplies anchors used by tests of those behaviors; phase/test maps are evidence accounting only. |

## 4. Public Contract and Behavior Freeze Map

The default for every later structural slice is exact observable compatibility.
No public export, selector byte, validation outcome, label, token, or density
value may drift as part of a mechanical move.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Package root import and public TypeScript surface | `package.ui` | `package.json` exposes only `@cartulary/ui-contracts`; 155 live import hits. | Type-check, package unit tests, consumer unit/E2E tests. | Snapshot or compile-time accounting of all root export names and relevant public types before internal extraction. | High | Preserve the single root import path, symbol names, types, and module evaluation behavior. Subpaths are deferred. |
| Exact test-ID and selector strings | `package.ui`, with subsystem-owned semantic inputs | Public builders in `src/index.ts`; runtime components and tests share outputs. | Broad coverage in `index.test.ts`, app tests, grid tests, test-utils, and E2E. | Direct tests for currently uncovered families and output bytes. | High | Percent encoding, separators, prefixes, CSS escaping, and DOM query shapes are observable contracts. |
| Fail-closed selector input validation | `package.ui` | `require*` helpers reject empty, malformed, or unknown values and validate view-schema IDs. | Several invalid-input cases in `index.test.ts`. | Add missing invalid-input branches for under-tested public families. | High | Opaque IDs remain syntax-opaque and losslessly encoded; do not infer identifier structure. |
| Exported vocabulary arrays and ordering | `package.ui`, with matching semantic owner | Exported `as const` arrays and union types in `src/index.ts`. | Partial direct unit coverage and consumer compilation. | Characterize array membership and order where runtime consumers or tests rely on iteration. | Medium | Order is frozen until evidence establishes that it is unobservable. |
| Workbook shell/controller types and ARIA labels | `apps/web` candidate; currently exported by `package.ui` | `workbookShellContracts.ts`, `WorkbookShellSlots.tsx`, incident-admin/controller callers. | App unit tests and browser tests. | Directly freeze exact shell slot labels and compile-time public type compatibility before any move. | High | Label text is user-facing accessibility behavior, not merely an internal constant. |
| Grid semantic DOM and UI-selector contract | `packages/grid-adapter` implementation plus `package.ui` selector facade | `SemanticDataGrid.tsx`, grid-adapter tests, selector policy test. | Grid-adapter unit tests, package selector tests, test-utils, E2E. | Add direct coverage for filter helpers, history-open helpers, and any untested semantic attributes. | High | Preserve scrollport class, roles, data attributes, row/cell/group selectors, and absence of vendor imports outside the adapter. |
| Density-derived grid row heights | Design-token projection and `package.ui` facade; adapter consumes result | `workbookGridRowHeightPx` parses generated fixed-pixel density tokens. | Token tests assert token values; callers compile. | Add direct tests for all densities and the invalid-token failure path. | Medium | Exact numeric results and thrown error semantics are frozen. |
| Generated design-token exports, token order, default theme, and theme CSS | Design-token projection/harness; exposed by `package.ui` | `tokens.v1.json`, generated file, generator, artifact and import-boundary policies. | `design-tokens.test.ts`; drift, policy, and JSON-shape checks. | Existing tests are sufficient for a no-change internal split; generator/input redesign remains blocked. | High | Never hand-edit generated output. CSS text and token iteration order are observable. |
| Saved-view and view-schema selector behavior | Saved-view/view-contract owners plus `package.ui` facade | Registry-backed validation and saved-view selector builders. | Package tests and saved-view app/E2E tests. | Freeze registry acceptance/rejection and exact saved-view selectors before changing registry access. | High | No saved-view persistence or schema parsing is implemented here. |
| Entity row/query/mutation UI anchors | Entity owner semantics plus `package.ui` facade | Entity selector families and entity-focused callers. | App/E2E coverage; partial package coverage. | Direct tests for reusable-entity selectors, merge preconditions, and invalid entity tokens. | Medium | The package does not perform entity query or mutation behavior. |
| Evidence, timeline, revision/change-set, history, conflict, and projection-refresh UI anchors | Matching subsystem owners plus `package.ui` facade | Feature selector families and app/E2E callers. | Cross-layer app and browser tests; partial package tests. | Direct tests for row-history open helpers, draft-row creation, paste-conflict helpers, and projection-refresh anchors before movement. | Medium | The package supplies stable DOM anchors only; source mutation, revisions, conflicts, and projection refresh remain outside it. |
| Network Flow selector surface | Adopted Network Flow subsystem owner plus `package.ui` facade | Network Flow types/builders and runtime/unit/E2E callers. | Consumer tests; incomplete direct package characterization. | Directly characterize every public Network Flow builder, output byte, validation branch, and relevant exported array order. | High | Retain the shared automation facade even if source is privately grouped. |
| Authorization-related selector surface | Security/application owners plus `package.ui` facade | Auth/account/admin selectors and runtime callers. | Auth/admin app and E2E tests; package selector tests. | Freeze exact selectors when moved; no authorization-decision characterization belongs in this package. | Medium | No authorization check or outcome is computed here. |
| HTTP routes, request/response envelopes, and WebSocket paths/events | Backend transport/application owners; not `ui-contracts` | No route handler, fetch client, WebSocket implementation, or event logic found in the target. | Service/browser suites outside this package use UI anchors. | None for a selector-only mechanical split; run affected browser targets if selector bytes or shell behavior change. | Low, indirect | Route and event behavior remains frozen globally but is not implemented by this package. |
| Storage semantics, source mutation, revision writes, projection refresh, and authorization decisions | Matching domain/platform/application owners; not `ui-contracts` | Direct target inspection found none of these implementations. | Backend and browser evidence outside package ownership. | None in this package. | Low, indirect | Test rows and UI anchors do not transfer runtime ownership to this package. |
| Harness/test accounting | Testing Harness owner and `package.ui` catalog inputs | `package.ui` owner/family manifests route two unit rows and four static/boundary rows. | `make test-slice OWNER=package.ui` passed 6/6 at the discovery commit. | Update authored catalog inputs only if a later slice adds a separately routed test file or changes ownership. | Medium | Harness maps are verification accounting, not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Direct package characterization is incomplete for public Network Flow builders, row-height parsing, filter helpers, history-open helpers, reusable-entity selectors, and several invalid-input paths. | Export-to-test comparison against `src/index.test.ts` and `src/design-tokens.test.ts`; consumer tests provide indirect coverage but do not isolate all package contracts. | A mechanical split could silently change selector bytes or failure behavior while broad tests remain insensitive. | `must_fix` | `package.ui` test ownership, with semantic inputs from matching owners | Add behavior-focused direct characterization before moving each affected family. Update owner-catalog accounting only through authored inputs if routing changes. |
| `src/index.ts` combines selector core, application-shell vocabulary, grid semantics, domain-feature families, Network Flow vocabulary, and design-token facade logic. | 2,349 lines, 229 functions, 274 declarations, and 155 inbound import hits. | Large blast radius and difficult review; unrelated owners are co-located. | `should_fix` | `package.ui` root facade with private owner-grouped modules | Plan an internal partition that retains the exact single root surface and output bytes. |
| Shell/controller types and labels live in the selector package despite the guide assigning application shell and state to `apps/web`. | Runtime callers include shell-contract, admin-panel, and controller files. | Moving them can break public types and user-facing ARIA labels. | `defer` | `apps/web` for state; `package.ui` for shared automation labels where required | Reassess in an authorized owner decision after characterization. Do not move as part of the first mechanical split. |
| `WorkbookSurface` is an unconstrained `string` while view-schema identity is generated elsewhere. | Root type declaration and protocol registry dependency. | Tightening or relocating the type could break many callers or create a dependency cycle. | `defer` | `packages/view-contracts` / `@cartulary/protocol-ts` | Map compile-time and runtime callers before proposing a replacement. |
| Root module initialization imports the `protocol-ts` registry facade to validate view-schema IDs. | `listViewSchemaRegistryEntries` import; protocol root aggregates generated registry families. | Possible broad module loading or bundle impact has not been measured. | `defer` | View-contract/protocol owner with a narrow consumer seam | Measure build/bundle behavior and dependency direction before changing the import. Do not infer impact from source size alone. |
| Apparently unused exported arrays, types, and selectors remain part of the root surface. | Repository-wide caller search finds no external uses for some exports, but public package consumers outside the repository are not proven absent. | Removing them could be a breaking API change. | `defer` | `package.ui` | Preserve by default; removal requires explicit later authorization and compatibility evidence. |
| Production components consume shared test-ID builders. | Runtime imports from `apps/web` and grid adapter; frontend guide requires runtime-safe shared selectors. | Misclassifying this as test leakage would fragment selector strings. | `intentional/no_action` | `package.ui` | Retain the shared runtime/test facade; keep choreography in `packages/test-utils`. |
| Semantic grid selectors and density helpers are outside the grid adapter, but direct vendor integration is not. | `ui-contracts` has no `react-data-grid` import; `SemanticDataGrid.tsx` consumes the shared semantic contracts. | Moving all grid selectors into the adapter could break shared automation imports. | `intentional/no_action` | `package.ui` facade plus `packages/grid-adapter` implementation | Preserve dependency direction and enforce the existing vendor import boundary. |
| Generated design tokens are accessed through the authored root facade. | Import-boundary policy, artifact policy, generator, and tests all enforce this path. | Bypassing the facade or editing output risks generated drift. | `intentional/no_action` | Design-token projection/harness and `package.ui` facade | Keep current access path; regenerate only through Make-owned tooling when separately authorized. |
| `BLOCKED: owner contradiction` — the adopted Testing Harness NLSpec says tests and generators do not consume documentation, while its design-token helper row names `docs/design.md` token parsing; the live generator reads `contracts/design/tokens.v1.json`. | `docs/testing-harness-nlspec.md`; `tools/harness/generated-artifacts/design-tokens/design-token-cli.mjs`; `contracts/design/tokens.v1.json`; repository procedure. | Choosing either prose statement as the redesign rule would contradict the other owner text. | `defer` | Testing Harness owner and design-token owner | Block generator/input-ownership redesign until the adopted owner text is repaired. Preserve the live machine-input path meanwhile. |
| No domain logic under transport/persistence, direct SQL/storage coupling, hidden mutation/revision/projection side effects, or wrong-layer authorization checks were found in the target. | Direct inspection of every tracked target file and dependency/import search. | Inventing such workstreams would expand scope and obscure the actual frontend seam. | `intentional/no_action` | Existing backend/platform/module owners | Keep backend facade, route, storage, and authorization refactors out of this target plan. |
| Repository callers increased from 153 at discovery commit `24f301...` to 155 at live commit `e2b57a...`, while target files remained unchanged. | Live `rg` caller count and commit-range diff. | A stale caller inventory could omit consumers during a future move. | `should_fix` | Future implementing session | Re-run caller enumeration immediately before each implementation slice and record any new consumers. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish target, authority, commit, scope, dirty-tree posture, and allowed writes. | This tracker; framework; authority documents. | `git status --short`; `git rev-parse HEAD`; framework inspection. | Scope, source order, planning-only constraint, and commit mismatch are recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every tracked and ignored target file plus live callers and dependencies. | All files under `packages/ui-contracts`; representative callers. | File/caller searches; line/export/test counts. | Every target file is inventoried or explicitly out of scope. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map selector, shell, grid, view-schema, Network Flow, token, and harness responsibilities without assuming relocation. | `src/index.ts`; owner docs; consumer sources; generated inputs/policies. | Owner-document comparison and import-direction review. | Every responsibility has an owner candidate and keep/move/split/defer posture. |
| WF-03 | Characterization-test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06 | Identify direct evidence needed before moving high-risk public families. | `src/index.test.ts`; `src/design-tokens.test.ts`; consumer tests; test family manifests. | `make test-slice OWNER=package.ui`; later `make frontend-unit`. | Each public contract family has an existing or required test posture. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Check platform/domain, grid-vendor, generated, test-util, protocol, and shell ownership seams. | Target sources; grid adapter; test-utils; app shell; boundary and artifact policies. | `make frontend-import-boundary-check`; `make generated-artifact-policy-check`; `make frontend-typecheck`. | Findings are classified and unsupported workstreams are excluded. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Define a private internal partition while preserving the root facade; isolate deferred owner decisions. | `src/index.ts`; prospective authored source modules; no generated file edits. | Compile-time root compatibility, unit tests, import-boundary check. | Proposed seam keeps exact root exports and names every deferred public decision. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Sequence characterization, mechanical extraction, deferred ownership study, and final verification. | Files listed per slice in Section 7. | Cheapest Make-owned target per checkpoint. | Each slice has dependency, rollback, validation, and binary completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Keep package owner rows, generated policies, and any new test routing aligned without making them runtime architecture. | Tests and authored `package.ui` verification/test-family inputs only if a later test slice requires changes. | `make explain-test-owner OWNER=package.ui`; `make test-slice OWNER=package.ui`; drift/policy checks. | No hand-edited generated output; owner accounting matches any authorized test changes. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow-to-broad validation, review scope, and leave a restartable handoff. | Implementation diff in a later task; this tracker. | Section 8 command matrix. | Passing commands and run roots are recorded; failures and deferrals are explicit. |

## 7. Proposed Refactor Slice Plan

Every implementation slice below requires a later authorized task. Slices S-01,
S-02, S-04's measurement step, and S-06 are intended to preserve behavior.
Any public type removal, package subpath change, selector-byte change, label
change, dependency-direction change, or behavior correction is marked as
requiring separate later authorization.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | Completed planning tracker | Add direct package characterization for Network Flow builders, all density row heights and invalid token parsing, filter helpers, row-history open helpers, draft-row and paste-conflict helpers, reusable-entity/merge selectors, exported ordering where observed, and missing invalid-input paths. Update authored owner-catalog accounting only if a new routed test file requires it. | `packages/ui-contracts/src/index.test.ts`; possibly a new authored package test; possibly authored `tools/test_families/package.ui.json` and `contracts/verification/owners/package.ui.json` in a separately authorized accounting edit. | Tests could accidentally bless an inferred behavior rather than current output. | Add exact byte/failure characterization; preserve all existing package and consumer tests. | `make test-slice OWNER=package.ui`, then `make frontend-unit` | Revert only the new characterization/accounting edits; no production source or generated output is changed in this slice. | Every family selected for S-02 has direct behavior evidence and remains in correct owner accounting. |
| S-02 | S-01 | Extract private authored modules for selector core/encoding, app and shell, workbook/grid, owner-grouped feature selectors, Network Flow selectors, and the authored design-token facade; retain one root entry point with byte-compatible exports and module behavior. | `packages/ui-contracts/src/index.ts`; new authored files under `packages/ui-contracts/src/`; never `src/generated/**`. | Export omission, type drift, validation-order drift, selector-byte drift, cyclic initialization, or changed generated-token evaluation. | Keep S-01 tests and all existing tests unchanged during the mechanical split. Add no behavior changes. | `make test-slice OWNER=package.ui`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make generate-drift` | Revert one owner-grouped extraction at a time while the root facade remains the stable rollback point. | Root import path, public names/types, exact outputs, thrown failures, array order, labels, token exports, and caller behavior are unchanged. |
| S-03 | S-02 | Reassess ownership of workbook shell/controller types and labels. Move only state that is proven application-owned while retaining any shared automation contract. **Requires later authorization** for public type removal, replacement, or label change. | `packages/ui-contracts`; `apps/web/src/shared/workbookShellContracts.ts`; admin/controller/shell callers. | Public API breakage, duplicate vocabulary, circular imports, ARIA-label drift. | Add/retain compile-time caller tests, shell component tests, selector tests, and affected browser coverage. | `make frontend-typecheck`; `make frontend-unit`; affected Make-owned browser target if runtime labels or shell behavior could change | Keep compatibility re-exports until all callers and browser behavior pass; roll back relocation without touching S-02's private grouping. | An owner decision is documented and either the type remains intentionally shared or an authorized compatibility-preserving migration is complete. |
| S-04 | S-02 | Measure protocol-registry loading and bundle impact, map dependency direction, and decide whether a narrower view-schema identity seam is warranted. Any dependency-direction or public type change **requires later authorization**. | `packages/ui-contracts`; `packages/protocol-ts`; `packages/view-contracts`; relevant web build entry points. | Bundle growth, module initialization changes, registry drift, dependency cycles, invalid view-ID acceptance changes. | Preserve registry acceptance/rejection tests and build/type-check evidence; add a focused measurement only through repository-owned tooling. | `make frontend-typecheck`; `make build-web`; `make generate-drift` | The measurement step has no source rollback. If an authorized change follows, retain the current registry path as the rollback seam until equivalence passes. | Measured evidence supports an explicit keep/change decision; no source-size inference is treated as bundle proof. |
| S-05 | RB-001 resolved; S-02 | Reassess the generated-token facade and generator input ownership. **BLOCKED: owner contradiction** and **requires later authorization**. | Adopted Testing Harness owner text; design-token contract input; harness helper; `packages/ui-contracts/src/generated/design-tokens.ts` only as generator output. | Owner violation, documentation-dependent generation, hand-edited generated output, token/CSS/order drift. | Preserve token facade tests; add owner-aligned generator tests only after the contradiction is repaired. | `TODO: command discovery required after owner repair`; expected existing gates include `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check`. | Do not begin until owner repair. If later generated output drifts unexpectedly, revert authored inputs and regenerate; never hand-edit output. | Adopted owner text is internally consistent, the machine input is explicit, and Make-owned generation reproduces protected output. |
| S-06 | S-01 through applicable authorized slices | Run final narrow-to-broad verification, inspect scope, maintain retained evidence, and update this handoff. | All files changed by the later authorized task; this tracker. | False confidence from stale run roots or skipped consumer/browser checks. | Preserve package, consumer, boundary, generated, and affected browser evidence. | `make agent-finalize`, followed by `make check`; use affected browser target when required. | Revert the smallest failing slice; do not combine cleanup with an unresolved failure. | All required current-commit commands pass, generated output is clean, only authorized files changed, and the handoff names run roots and any skipped checks. |

No slice may hand-edit `packages/ui-contracts/src/generated/design-tokens.ts`.
During S-02, test files and generated output stay unchanged. If S-01 adds a new
test family or changes catalog ownership, that accounting work is a separately
authorized authored-input edit followed by the applicable generator/drift
checks.

## 8. Validation Plan

Canonical commands were discovered through `make help`,
`make task-guide ROLE=module-author OWNER=package.ui`,
`make explain-test-owner OWNER=package.ui`, and targeted `make explain-target`
queries. The current owner has six work units and zero service-backed rows.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=package.ui`, then `make frontend-unit` | Package-owner unit/static rows followed by broader frontend unit consumers | yes | Run the owner slice before movement to establish a current-commit baseline; broaden after characterization or extraction. |
| integration | N/A | `package.ui` currently has zero service-backed test rows | no | Re-run owner explanation before implementation. Add integration only if a later authorized change introduces a real service boundary. |
| e2e/browser | `TODO: select the affected Make-owned browser target` | Selector-byte, runtime shell, focus, accessibility, or browser workflow impact only | no | Required if selector bytes or runtime shell behavior might change. Choose from the repository's public browser targets after identifying the affected owner/family; do not invent a direct Playwright command. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated token/protocol drift, protected-root policy, and authored JSON shape | yes | Required before and after any authorized contract/generator-adjacent slice; generated files are never hand-edited. |
| import-boundary/static | `make frontend-typecheck`; `make frontend-import-boundary-check` | Public TypeScript compatibility and package/vendor/generated-token import rules | yes | Required for characterization and every source extraction. |
| full check | `make agent-finalize`, followed by `make check` | Final retained-run maintenance and repository-wide gate | no | Required at the final implementation checkpoint, not for this tracker-only edit. Supply a successful `RESULTS_DIR` when retaining a full warm run; otherwise report the unset skip honestly. |
| tracker-only | `make lint-markdown` | Markdown documentation | yes | The only product-tree change in this session is this tracker. |

### Recorded discovery baseline

These commands actually ran and passed during the discovery session at commit
`24f3010158934a61bed02804fc97093ef625ec56`. They are retained historical
evidence, not a claim that the same commands ran at live commit `e2b57a...`.

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

### Current tracker-only validation

| Command | Result | Evidence |
| --- | --- | --- |
| `make lint-markdown` | PASS at live commit `e2b57a73bf7388a35404fa9b74f73c317d3e2cd8` after tracker creation | `.cartulary/test-results/20260731T021608Z-p3895031` |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| UI-T-001 | Normalize `packages/ui-contracts` to the safe label `ui-contracts` and constrain the session to tracker-only planning | WF-00 | DONE | none | Section 1 scope posture | Target, output, allowed write, and non-goals are explicit. |
| UI-T-002 | Inventory every target file and live inbound dependency | WF-01 | DONE | UI-T-001 | Section 2; live file and import searches | Six tracked files and two ignored artifacts are accounted for; 155 current import hits are recorded. |
| UI-T-003 | Map selector, shell, grid, view-schema, Network Flow, token, and harness owners | WF-02 | DONE | UI-T-002 | Sections 3 and 4 | Each discovered responsibility has an owner candidate and keep/split/defer posture. |
| UI-T-004 | Freeze observable package and UI contract behavior | WF-02 | DONE | UI-T-003 | Section 4 | Root API, bytes, failures, order, labels, DOM attributes, tokens, CSS, and density results are named. |
| UI-T-005 | Identify characterization gaps for high-risk public families | WF-03 | DONE | UI-T-003 | Section 4 test postures and S-01 | Missing direct evidence is enumerated without modifying tests. |
| UI-T-006 | Add the missing direct package characterization | WF-03 | TODO | UI-T-005, later authorization | Future S-01 test evidence | Each family selected for extraction has direct current-behavior coverage. |
| UI-T-007 | Partition the monolithic authored source behind the unchanged root facade | WF-05 | TODO | UI-T-006, later authorization | Future S-02 source diff | Root surface and all observable outputs remain exactly compatible. |
| UI-T-008 | Decide application-shell/controller type ownership | WF-05 | DEFERRED | UI-T-007, RB-002 | Future owner decision | Shared automation labels and app-only state have evidence-backed owners and compatibility posture. |
| UI-T-009 | Measure and decide the protocol registry/view-schema seam | WF-05 | DEFERRED | UI-T-007, RB-004 | Future build/bundle evidence | Keep/change decision is based on measured loading and dependency evidence. |
| UI-T-010 | Repair the design-token generator owner contradiction | WF-07 | BLOCKED | RB-001 | Adopted owner-document repair | Testing Harness text consistently names the allowed machine input and no docs-dependent generation is introduced. |
| UI-T-011 | Reassess the generated-token seam | WF-07 | BLOCKED | UI-T-010, later authorization | Future S-05 decision | Owner contradiction is resolved and Make-owned generation proves any authorized seam. |
| UI-T-012 | Run final implementation verification and update handoff | WF-08 | TODO | Applicable future slices | Current-commit Make run roots and final diff | Required commands pass and skipped checks have reasons. |
| UI-T-013 | Create the target-specific tracker and current session handoff | WF-08 | DONE | UI-T-001 through UI-T-005 | This file | Another agent can resume without rediscovering scope, inventory, risks, commands, or blockers. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Planning complete at live commit `e2b57a...`; repository was clean before tracker creation; prior discovery was at `24f301...`. | Inspected framework, `domain.md`, Core 00-04, adopted Testing Harness and Network Flow NLSpecs, design and frontend guides; touched only this tracker. | `sed`; `rg`; `find`; `git rev-parse`; `git status`; `git log`; `git diff`; `wc`; `date`. | Target exists, normalized label is safe, target unchanged across commits, live caller inventory refreshed. | RB-001 affects generator redesign only. | Obtain later authorization for S-01 characterization. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Backend refactor is out of scope; no backend responsibility was found in the target. | Inspected all tracked target files and searched imports/symbols; touched only this tracker. | Target file inspection and repository searches. | No route handler, WebSocket transport, SQL/storage, authorization decision, mutation coordinator, revision writer, or projection refresh implementation found. | None for planning. | Keep backend workflows excluded unless new live evidence appears. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Legitimate shared selector facade; mixed responsibilities require private partition and later owner decisions. | Inspected `src/index.ts`, representative app-shell, grid-adapter, test-utils, runtime, unit, and E2E callers; touched only this tracker. | `rg -l '@cartulary/ui-contracts' apps packages`; source inspection; export and line counts. | 155 live caller files; no grid-vendor import in target; shell/grid/view/Network Flow seams mapped. | RB-002 through RB-005 defer public ownership/API changes. | Execute S-01 before any source movement. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Generated design tokens remain protected and accessed through the root facade. | Inspected generated token output, token registry, generator, generated-artifact policy, import-boundary policy, and verification owner inputs; touched only this tracker. | Discovery baseline: `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`. | All three passed at commit `24f301...` with run roots recorded in Section 8; not rerun at `e2b57a...`. | RB-001: `BLOCKED: owner contradiction`. | Preserve live JSON-to-generated path; repair adopted owner text before redesign. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Existing package tests cover many but not all high-risk public families; owner has six non-service-backed rows. | Inspected both target tests, `package.ui` family/verification inputs, catalog owner, helper ownership, and representative consumer tests; touched only this tracker. | `make help`; `make task-guide ROLE=module-author OWNER=package.ui`; `make explain-test-owner OWNER=package.ui`; targeted `make explain-target`; discovery `make test-slice OWNER=package.ui`. | Discovery owner slice passed 6/6 at `24f301...`; run root `.cartulary/test-results/20260731T015525Z-p3878389`. | Missing direct characterization blocks safe extraction until S-01. | Add S-01 tests in a later authorized task, then rerun current-commit owner and frontend unit targets. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:10 EDT | Codex / tracker implementation | Target exports auth/account/admin UI anchors but implements no authorization policy or decision. | Inspected target selector families, representative auth/admin callers, Core 04 relevant sections, and selector policy test; touched only this tracker. | Source and caller searches. | Security/authorization outcomes remain owned outside this package; exact UI anchors remain frozen. | None for selector-only planning. | Do not move authorization logic into this package; run affected browser owner tests only if anchors change under later authorization. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30 22:16 EDT | Codex / tracker implementation | Tracker created; planning workflows are complete; implementation is not authorized. | Touched only this tracker; inspected live status and all evidence named above. | `make lint-markdown`. | PASS; run root `.cartulary/test-results/20260731T021608Z-p3895031`. Safe next slice is characterization, not source movement. | RB-001 blocked; RB-002 through RB-005 deferred. | Run `make task-guide ROLE=module-author OWNER=package.ui`, refresh caller inventory, then implement S-01 only under later authorization. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | `BLOCKED: owner contradiction` — does the adopted Testing Harness owner intend design-token generation to parse `docs/design.md`, or to consume only the machine projection `contracts/design/tokens.v1.json` as the live generator and repository procedure require? | A generator/input redesign cannot choose one conflicting owner statement without violating the other. | Repair the adopted Testing Harness design-token helper row so it agrees with its no-document-dependency rule and identifies the authoritative machine input. | BLOCKED for S-05 and UI-T-010/UI-T-011; does not block S-01 or S-02. |
| RB-002 | Which shell/controller declarations are application-only state, and which are stable runtime/test automation contracts that should remain exported? | Moving the wrong declaration can break public TypeScript callers or change user-facing ARIA labels. | Complete S-01 characterization, enumerate compile-time consumers, and obtain an `apps/web` versus `package.ui` owner decision. | DEFERRED to S-03. |
| RB-003 | Should `WorkbookSurface = string` remain deliberately open, or be replaced by a generated/view-contract-owned identity? | Tightening it may reject extensions or create package dependency cycles; leaving it broad may conceal invalid values. | Adopted view-schema/extension owner evidence plus a complete caller and compatibility analysis. | DEFERRED; no change proposed. |
| RB-004 | Does importing the `protocol-ts` root registry from `ui-contracts` materially affect browser module loading or bundle size, and is a narrower registry seam warranted? | Source-file size alone does not prove runtime cost, while changing the seam can alter initialization and dependency direction. | Repository-owned build/bundle measurement and import-graph evidence at the implementing commit. | DEFERRED to S-04. |
| RB-005 | Are repository-unreferenced exports safe to remove, or are they intentional public compatibility surface? | The package is private but has a declared root API; repository search cannot prove the absence of all consumers or future generated use. | Explicit package compatibility policy and later authorization, backed by caller and history evidence. | DEFERRED; preserve all exports. |
| RB-006 | Which browser target is required for a concrete slice? | Running no browser evidence after a selector-byte or shell behavior change could miss observable regressions; running every browser family for a private no-op split may be disproportionate. | Inspect the actual authorized diff and use Make owner/family explanation to select the affected public browser target. | TODO only when a later slice can affect selector bytes or runtime shell behavior. |

These blockers and deferred questions do not prevent this tracker from being
complete. They prevent unsafe later implementation decisions from being made
without owner repair, characterization, measurement, or authorization.

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `packages/ui-contracts` is inventoried or explicitly out of scope. | PASS | Section 2 accounts for all six tracked files plus ignored `node_modules/**` and `tsconfig.tsbuildinfo`. |
| Every discovered public contract risk has an owner and test posture. | PASS | Sections 3 and 4 map selectors, validation, shell labels/types, grid DOM, view schemas, Network Flow, tokens, indirect backend surfaces, and harness accounting. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 provides predecessor/successor relationships, validation, and handoff checkpoints for WF-00 through WF-08. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | PASS | Section 7 freezes behavior and marks public API, label, dependency, generator, and behavior changes as later-authorized decisions. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 names Make-owned commands, marks browser selection conditional, and records zero service-backed owner rows. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | RB-001 and Sections 5 and 7 block design-token generator/input redesign. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Sections 1, 3, and 5 record the narrow framework description versus the mixed live package, the commit/caller-count change, and the live machine-input path. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records scope, owners, target posture, commands, evidence roots, blockers, and the safe next slice. |
| No generated file hand-edit or production behavior change is planned in this session. | PASS | Only this tracker is touched; generated roots and all implementation files remain unchanged. |
| Implementation requires a later authorized task. | PASS | Section 1 and every future slice state the authorization boundary. |

The planning tracker is complete. Later implementation remains intentionally
blocked or deferred where the available authority and repository evidence do not
support a safe decision.
