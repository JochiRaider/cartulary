# grid-adapter Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `packages/grid-adapter`
- **Normalized target label:** `grid-adapter`
- **Output path:** `docs/handoffs/grid-adapter-module-refactor-tracker.md`
- **Repository baseline:** branch `main`, commit `a10e14f184ee626b00a28716426a7668f6091262`, two commits ahead of `origin/main`; no working-tree changes were observed before this tracker was created.
- **Status:** remediation direction decided; planning and documentation only in this session.
- **Allowed change in this session:** this tracker file only.
- **Non-goals:** no production code, test, contract, generated artifact, dependency or package configuration, migration, phase map, ledger, schedule, or harness edit; no behavior change; no implementation patch.
- **Later authorization:** every implementation, test, contract, code-generation, package, and harness change described below requires a separate authorized task.
- **Owner decision recorded on 2026-07-14:** the production runtime will become a real `react-data-grid` component integration. Ordinary workbook grids will use `DataGrid`; grouped workbook grids will use `TreeDataGrid`. The custom CSS-grid renderer is not a retained target architecture.

Source hierarchy used for this tracker:

1. No adopted subsystem NLSpec was found for grid-vendor integration or the `grid-adapter` package. Adopted NLSpecs remain authoritative only for their named subsystems.
2. Core 00 through Core 04 own current implementation-conformance behavior.
3. Core 05 applies only to claim-bearing timed or fixture-sensitive publication; this tracker creates no Core 05 claim.
4. `docs/domain.md`, design and implementation guides, and the testing-harness NLSpec provide terminology, package-boundary, design-direction, and execution support.
5. Live code, tests, manifests, and generated ledgers describe current repository state.
6. The modular-refactor framework, research R09, existing handoffs, phase labels, and ledgers are supporting evidence only.

Owner and support documents inspected:

- `AGENTS.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially architecture, view-row, view-schema, and projection sections
- `docs/spec/02_domain_model_schema_and_history.md`, especially record identity, view-contract, schema, and history boundaries
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, especially workbook, addressing, paste, keyboard, grouping, and interaction invariants
- `docs/spec/04_security_deployment_and_conformance.md`, especially incident authorization
- `docs/design.md`, especially density, grid layout, keyboard, visual, and accessibility direction
- `docs/testing-harness-nlspec.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`
- `docs/research/R09-react-data-grid-research-report.md`

Repository files inspected include every tracked file under the target, ignored target-local install/cache entries, and the following relevant callers and evidence owners:

- `apps/web/package.json`, `apps/web/vite.config.ts`, root `package.json`, root TypeScript configs, and `pnpm-workspace.yaml`
- Workbook contract-row, focus, Timeline anchor, Timeline grid, entity, generic, assessment, density, loader, and renderer modules under `apps/web/src/workbook/**`
- Grid-related workbook unit tests and browser selector/evidence usage under `apps/web/src/workbook/**` and `apps/web/e2e/**`
- `packages/ui-contracts/src/index.ts` and grid-related selector tests; grid-related helpers under `packages/test-utils/src/**`
- `tools/frontend_import_boundaries.json`, `tools/fallow/cartulary-boundaries.rulepack.json`, `tools/generated_artifact_policy.json`, and `tools/test_accounting_classification.json`
- `tools/phase3_test_map.json`, `tools/frontend_phase_maps/fe_p3_test_map.json`, and their generated Phase 3 ledgers
- Public Make target discovery and target guidance for frontend unit, type-check, import-boundary, browser, generated-policy, JSON-shape, Markdown, and full-check targets

No Core owner contradiction was found. Core 00 through Core 04 remain vendor-neutral behavior owners. The material mismatch is between the live custom renderer and implementation-support descriptions of an RDG-backed adapter; the owner has now resolved that mismatch in favor of a real RDG component integration. This tracker specifies the remediation but does not implement it.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `packages/grid-adapter/package.json` | Declares the private workspace package, runtime/dev dependencies, and export map. | Package roots `@cartulary/grid-adapter` and `@cartulary/grid-adapter/test-support`. | pnpm workspace; `apps/web/package.json`; root TypeScript project references. | React 19, ReactDOM 19, `@cartulary/ui-contracts`, `react-data-grid@7.0.0-beta.59`; test libraries. | Frontend install, type-check, import-boundary, unit, and build surfaces consume the manifest indirectly. | Lockfile records the dependency but is tool-managed and not touched. | `grid-adapter` package facade/configuration. | Medium | Exact RDG version is pinned here; changing it is outside this planning task. |
| `packages/grid-adapter/tsconfig.json` | Defines the strict no-emit TypeScript project using bundler resolution and DOM/JSX libraries. | Build/type-check project surface only. | Root `tsconfig.json` project reference and frontend type-check. | `../../tsconfig.base.json`. | `make frontend-typecheck`. | None. | Frontend toolchain support. | Low | Authored configuration; not changed by this task. |
| `packages/grid-adapter/src/core.ts` | Defines semantic grid row/column models; identity validation; row reconciliation; presentation grouping; stable cell anchors; keyboard navigation; paste targeting; renderer/editor resolution and cleanup. | Root re-exports include grid types, `assertGridRows`, `buildGridPresentationRows`, `cleanupGridAdapters`, `isGridColumnEditable`, `navigateGridCellAnchor`, `reconcileRecordRows`, `resolveGridCellAnchor`, `resolveGridPasteTargets`, and `resolveGridRenderer`. `./test-support` re-exports a subset. | Production workbook row loaders, focus controllers, Timeline/entity paste and navigation controllers, surface renderers through types, and package tests. | React types only. | `src/core.test.ts`, `src/index.test.tsx`, app grid-anchor tests, and workbook tests through the root facade. | No generated imports. Stable `record_id`/`field_key` meaning is owner-derived but represented through authored types. | Cartulary-native semantic grid adapter core, except app-owned row reconciliation. | High | The replacement design separates committed rows from drafts, makes vendor translation private, and moves `reconcileRecordRows` to workbook query/collaboration state. |
| `packages/grid-adapter/src/index.tsx` | Production package entry; re-exports core; imports RDG CSS; renders the grid using custom React elements and CSS Grid; owns ARIA/data attributes, grouping collapse state, sort headers, row gutter, density, sizing, scroll padding, and styling. | All root exports plus `gridAdapterVendor`, `GridViewport`, and `GridTable`. | Timeline, entity, generic, and assessment workbook surfaces; authoritative workbook tests; browser E2E and visual selectors indirectly. | React, `@cartulary/ui-contracts`, `react-data-grid/lib/styles.css`, local core. | `src/index.test.tsx`; workbook grid tests; browser functional, visual, and accessibility evidence through live rendering. | Consumes the authored UI-contract selector facade; no protocol/view generated artifact. | Grid rendering and vendor-integration facade. | Critical | Replacement is decided: retire this custom renderer in favor of real `DataGrid`/`TreeDataGrid` orchestration, then remove `GridTable` and `gridAdapterVendor`. |
| `packages/grid-adapter/src/test-support.tsx` | Table-based semantic renderer used by explicitly allowed unit-test mocks while preserving core row/group and selector semantics. | `@cartulary/grid-adapter/test-support`: `GridViewport`, `GridTable`, `gridAdapterVendor="semantic-test-grid"`, and a subset of core types/helpers. | Workbook unit tests that explicitly mock the production adapter. Runtime import guards exclude production code. | React types, `@cartulary/ui-contracts`, local core. | Workbook autosave, payload, action-sequencing, save-state, query, Phase 5/6/7/8/9, and inspector tests. | No generated artifacts. Its import boundary is authored in `tools/frontend_import_boundaries.json`. | Adapter-owned lightweight test double. | High | Retain only for fast non-vendor tests, migrate it to the new semantic facade, share semantic compilation, and prohibit it from closing RDG interaction, accessibility, or visual evidence. |
| `packages/grid-adapter/src/core.test.ts` | Direct characterization of grouping, identity, semantic anchors, navigation, paste, editability, renderer precedence, and cleanup. | No production surface. Test titles include FE-P3 evidence scenarios and support coverage. | `make frontend-unit` through the app Vitest project. | Vitest and local core. | This is the test source. | Scenario/test accounting is mapped by authored frontend maps and test-accounting classification; generated ledgers are downstream only. | Grid-adapter unit evidence. | Medium | Some tests are authoritative FE-P3 rows; other cases are classified support evidence. |
| `packages/grid-adapter/src/index.test.tsx` | Direct regression coverage for production DOM, selectors, sort/group behavior, density, sizing, row identity, rerenders, and editable descendants. | No production surface. | `make frontend-unit` through the app Vitest project. | Testing Library, Vitest, React, UI-contracts, package root. | This is the test source. | Classified as support evidence in test accounting; no generated artifact is edited. | Grid-adapter renderer evidence. | High | Test description says RDG-backed, but assertions do not prove that an RDG component renders. |
| `packages/grid-adapter/src/css.d.ts` | Declares the RDG stylesheet module for TypeScript. | Ambient module declaration. | TypeScript compiler while checking `src/index.tsx`. | `react-data-grid/lib/styles.css` module name. | `make frontend-typecheck` indirectly. | None. | Grid vendor build seam. | Low | Required only while the stylesheet import remains. |
| `packages/grid-adapter/node_modules/**` | Workspace dependency symlinks and `.bin` shims, including `vite`, `vitest`, and `yaml`. | None; package-manager install state. | Local tools only. | pnpm store. | None as authored source. | Tool-managed dependency artifacts. | Out of scope. | Low | Ignored by `.gitignore`; all entries under this root are explicitly out of scope. |
| `packages/grid-adapter/tsconfig.tsbuildinfo` | Incremental TypeScript build cache. | None. | TypeScript compiler only. | TypeScript toolchain. | None. | Tool-managed cache artifact. | Out of scope. | Low | Ignored by `*.tsbuildinfo`; explicitly out of scope. |

The target contains no route handler, SQL, storage adapter, protocol decoder, view-contract parser, WebSocket client, authorization policy, revision writer, or generated source. Live outbound imports confirm only React, the UI-contract facade, local files, and the RDG stylesheet.

## 3. Module Boundary Diagnosis

Current classification:

- legitimate frontend grid integration facade;
- Cartulary-native presentation and semantic-coordinate orchestration layer;
- grid-vendor integration boundary in package ownership, but only a stylesheet-level vendor integration in live runtime code;
- frontend rendering surface with mixed semantic-core, presentation, style, and test-double responsibilities;
- not a transport-adjacent, persistence-adjacent, backend mutation-coordinator, saved-view owner, view-contract owner, projection owner, collaboration owner, or authorization owner.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Public Cartulary grid types and facade | `package.json`, `src/index.tsx`, `src/core.ts` | `packages/grid-adapter` | replace | Package exports and all production grid imports use the package facade. | Replace `GridTable` with `WorkbookDataGrid`; retain `GridViewport` as the useful shell boundary; migrate all in-repo callers atomically and keep no deprecated private-package alias. |
| Stable saved-row identity validation | `src/core.ts` | `packages/grid-adapter` | keep | Core 01/03 stable-identity rules; dev guide §6.6/§6.9; package tests. | Current API also exposes a separate `GridRow.key`; its relationship to `recordId` needs stronger characterization. |
| Record/cell anchor translation and navigation | `src/core.ts` | `packages/grid-adapter` | keep | App focus controllers call `navigateGridCellAnchor`; FE-P3 anchor tests. | App owns focus orchestration; adapter owns semantic coordinate translation. |
| Presentation grouping and rejection of group rows as targets | `src/core.ts`, `src/index.tsx` | `packages/grid-adapter`, driven by app-supplied contract data | keep | Core 03 grouping boundary; app passes stable `groupBy` and labels; core tests. | Adapter must not decide allowed grouping keys from labels. |
| Paste target translation | `src/core.ts` | `packages/grid-adapter` plus app sync/mutation owners | keep | `resolveGridPasteTargets` is used by Timeline and entity controllers. | Package returns semantic targets only and performs no HTTP mutation. |
| Row reference reconciliation after refresh | `src/core.ts`; called by workbook loaders | `apps/web` workbook query/collaboration state | move | Dev guide assigns query and collaboration state to the app; the helper is consumed only by app loaders and has no vendor-coordinate responsibility. | Move it to an app-owned workbook row-reconciliation utility, migrate both loader families, preserve structural sharing, and remove the grid-adapter export without a compatibility shim. |
| Renderer/editor registry and cleanup abstractions | `src/core.ts` | `packages/grid-adapter` | replace | FE-P3 tests cover the public helpers; no production consumer was found. | Replace row-only signatures with semantic render/edit contexts wired to RDG. Remove unused public registry helpers after callers migrate; retain only cleanup behavior that has a live adapter owner. |
| Production grid DOM, density, sizing, scroll, group collapse, and styles | `src/index.tsx` | `packages/grid-adapter`, with design tokens from `ui-contracts`/design owners | replace | Direct workbook consumers, semantic selectors, design direction, and visual evidence. | Preserve stable semantic selectors and accessibility state, not the custom element hierarchy, inline CSS-grid geometry, or current goldens. Let RDG own layout-critical mechanics. |
| Direct RDG integration | `src/index.tsx`, `package.json`, `src/css.d.ts` | `packages/grid-adapter` | replace | The exact RDG dependency and stylesheet already exist; the owner selected real component integration. | Use `DataGrid` for ordinary grids and `TreeDataGrid` for grouped grids, keep all direct imports contained, and remove the custom production renderer before the integration workstream exits. |
| Timeline-specific keep-all-rows-mounted workaround | `src/index.tsx` | `packages/grid-adapter` virtualization seam plus app anchor controllers | remove | `buildVirtualizedRows` ignores viewport inputs and mounts all rows; the development guide selects fixed row heights and enabled virtualization. | Delete the workaround and spacer machinery, use fixed numeric density heights and RDG virtualization, and restore focus/edit state through stable Cartulary anchors. |
| Semantic unit-test renderer | `src/test-support.tsx` | `packages/grid-adapter` test surface | split | Export map, explicit test imports, production test-helper boundary guard. | Retain the test-only facade; consolidate shared semantics after parity tests exist. |
| View-schema parsing and field capability decisions | `apps/web` and `packages/view-contracts` callers | `packages/view-contracts` plus app policy/controllers | keep outside target | `workbookContractRows.ts` supplies contract-derived columns; target has no view-contract dependency. | Grid adapter must not infer mutability, grouping, sorting, or write-back from labels. |
| HTTP mutation, saved views, projection refresh orchestration, revisions, conflicts, collaboration, authorization | `apps/web` and server owner modules | Existing app/server owners | keep outside target | Negative target searches and live callers show callback/type-only integration. | Preserve this exclusion during refactor. |

Architectural decision: `packages/grid-adapter` remains the permanent seam for Cartulary-native grid semantics and direct vendor containment. Its current custom renderer and broad public helper surface are not retained architecture. The replacement separates pure semantic logic, private RDG compilation, React orchestration, adapter-owned styles, and a narrow semantic facade. Core-owned behavior and stable selectors are compatibility inputs; custom DOM, inline grid-template values, synthetic RDG evidence, and all-rows-mounted behavior are not.

### 3.1 Approved Target State

- `WorkbookDataGrid` replaces `GridTable`; `GridViewport` remains only as the shell/layout boundary.
- Ordinary surfaces render `DataGrid`; grouped surfaces render `TreeDataGrid` with exactly one adapter-private group column.
- The public row model separates required committed `GridRecordRow` identity from an optional recordless `GridDraftRow`; committed RDG keys are `recordId`, while the draft uses a namespaced `draftKey` and remains outside record grouping, selection, and mutation targets.
- The public adapter exports `GridRecordRow`, `GridDraftRow`, `GridCellTarget`, `GridColumn`, `GridEditorAdapter`, `GridGroupingDescriptor`, semantic interaction intents, and a restricted `GridHandle`; it exports no RDG type, coordinate, or imperative handle.
- Row gutter and actions are adapter-private RDG columns. Custom RDG row/cell renderers preserve stable `ui-contracts` attributes and test IDs without treating RDG-generated class names or element nesting as contracts.
- Query sort arrays map to controlled `sortColumns`; saved-view widths map to controlled `columnWidths`; committed selection maps to record-ID sets; and Cartulary anchors map to RDG selection/scroll coordinates. Active cell, group expansion, edit, copy, paste, fill, and focus exit are controlled or translated through explicit Cartulary owners before reaching `apps/web`.
- The exact existing `react-data-grid@7.0.0-beta.59` pin remains unchanged during remediation. Dependency upgrade risk is handled only after the integration is complete.

## 4. Contract Compatibility and Characterization Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Root package exports and exported types/helpers | `packages/grid-adapter` | `package.json`, `src/index.tsx`, production imports | Type-check, package tests, app unit tests | Compile-time migration of every live consumer to the approved semantic facade | High | The package is private. Remove obsolete exports after atomic caller migration rather than carrying a deprecated alias layer. |
| `./test-support` export and test-only import boundary | Grid adapter plus frontend boundary manifest | Export map, `tools/frontend_import_boundaries.json`, fixture-accounting guard | Import-boundary shell tests and workbook mocked tests | Production/test facade shape and selector parity | High | Runtime code must continue to reject test-support imports. |
| Vendor identity and stylesheet singleton | Grid adapter implementation-support boundary | Actual component import/render, CSS import, dependency manifest, FE-S-P3-01 | Import-boundary target and new live-component tests | Verify a rendered RDG root and exercised vendor callbacks; verify exactly one stylesheet import | Critical | Remove `gridAdapterVendor`; a string constant is not evidence. |
| Saved-row identity | Core 01/03 behavior; adapter translation | `assertGridRows`, `GridRow`, app row builders | Core/index tests and U-3-GRID rows | Saved `key`/`recordId` agreement, duplicate/missing IDs, drafts with null ID, reorder stability | High | Row position must never become mutation identity. |
| Stable cell anchors and keyboard navigation | Core 03; adapter translation; app focus controllers | `resolveGridCellAnchor`, `navigateGridCellAnchor`, app focus modules | FE-U-P3-02, app anchor tests, Phase 9 browser evidence | Boundary movement, group/draft rejection, focus restoration after reorder/refresh | High | Preserve current behavior of clearing rather than inventing an anchor on presentation rows. |
| Paste target resolution | Core 03; adapter translation plus app mutation owners | `resolveGridPasteTargets`, Timeline/entity controllers | Core tests, Phase 9 sentinel and browser paste evidence | Sorted/filtered/grouped targets, overflow creates, create-disabled, recordless/group rejection | Critical | Package must continue to return semantic targets and never submit a mutation. |
| Presentation grouping and collapse | Core 03; app query state plus adapter presentation | `buildGridPresentationRows`, `GridTable` | Core/index tests; Phase 8 unit/browser; Phase 3 visual | Null/empty labels, committed versus draft rows, one header per bucket, collapse/expand, target exclusion | High | Query/contract owner decides allowed keys and row order. |
| Sort-header translation | Core 01/03; app query state plus adapter UI | `sortableFieldKey`, `onToggleSort`, sort props | Index tests and workbook query tests | Stable field-key dispatch, disabled headers, current sort label, no local authoritative reorder | High | Visible labels are non-authoritative. |
| Renderer/editor capability and cleanup | Grid adapter support boundary; contract capability remains external | Core registry helpers and public types | FE-U-P3-03/04 in `core.test.ts` | Production wiring or explicit non-wiring, lifecycle cleanup, renderer precedence, read-only cells | High | Current `GridTable` calls `column.renderCell` directly. |
| DOM, ARIA, and selector surface | Grid adapter plus `ui-contracts` selector owner | Semantic roles, data attributes, scrollport class, group toggle, row gutter | Live RDG tests, UI-contract/test-utils tests, app tests, browser suites | Stable semantic selectors and accessible state across RDG row/cell renderers | Critical | Custom element nesting and pixel geometry are intentionally not frozen; selectors expressing Cartulary identity remain stable. |
| Density, sizing, and scroll geometry | Design direction plus grid adapter implementation | Density types/tokens, `GridViewport`, RDG controlled widths and row heights | Live component, browser, and visual evidence | All density modes, controlled widths, shell fill, fixed-height virtualization, deep-scroll continuity | High | Numeric RDG row heights must derive through an authored `ui-contracts` projection of generated density tokens; design evidence is not Core conformance. |
| Query/projection refresh structural sharing | App query/collaboration state; helper currently in grid adapter | `reconcileRecordRows` and two loader families | FE-I-P3-01 and loader tests | Sparse/full refresh, unchanged references, changed row replacement, drafts/local pending state | High | Ownership is resolved to the app; preserve behavior while moving the helper and remove the adapter export without a compatibility shim. |
| View-schema and saved-view behavior | Core 01/03; `packages/view-contracts`; `apps/web` | Contract-derived columns/query state passed into adapter | U-3-GRID, Phase 8 query/saved-view tests | Confirm adapter remains driven by stable keys and does not parse labels or persist view state | Critical | No view-schema or saved-view implementation belongs in the target. |
| HTTP route shapes and envelopes | Core 01 route owners and `apps/web` transport/controllers | No route or fetch code in target; app callbacks perform requests | Workbook unit, backend integration, browser functional tests | Confirm refactor changes no callback payload or route invocation | Critical | Applicable indirectly; there is no target-owned HTTP contract. |
| WebSocket paths/events and projection refresh | Core 01/03 collaboration owners and app collaboration state | No WebSocket code in target; row identity must survive refresh | Phase 6 unit/browser and Phase 9 continuity evidence | Refresh/invalidate/remove continuity by record ID after renderer changes | Critical | Applicable indirectly; no target-owned WebSocket contract. |
| Authorization and closed-incident behavior | Core 04 and server/app policy owners | No authorization code in target; controls are supplied by callers | Workbook role/closed-incident and browser tests | Verify renderer never treats visibility/editability as authorization | Critical | Server re-derivation remains authoritative. |
| Revision/change-set/conflict behavior | Core 01/02/03 and app sync/conflict owners | No revision writer in target; paste/anchors feed app owners | Phase 3/4/6/7/9 tests | Confirm stable targets and no renderer-owned mutation or conflict decision | Critical | Package must not create hidden side effects. |
| Generated protocol/view contracts | Contract/codegen owners | No generated import or output in target | Generated policy, JSON-shape, type-check | Confirm refactor adds no generated hand edit or protected-root import | Medium | Any later contract change starts at its owner and generator. |
| Harness and evidence accounting | Testing-harness NLSpec; authored maps/manifests | FE-P3 map, Phase 3 map, test accounting, generated ledgers | Make target accounting and drift checks | Preserve scenario titles/paths or update authored owners then regenerate | High | Phase identity is evidence accounting, not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Production code claims an RDG vendor but imports only the RDG stylesheet and renders a custom CSS grid. | `src/index.tsx` has no `DataGrid`/`TreeDataGrid` import; `gridAdapterVendor` is a constant; package tests assert the constant. | Misleading architecture/evidence claims; a renderer change affects all grid behavior. | `must_fix` | Architecture owner plus `packages/grid-adapter` | Owner direction is resolved: implement real `DataGrid`/`TreeDataGrid`, remove the custom renderer and vendor marker, and prove component behavior through live evidence. |
| Semantic core, production renderer, public barrel, styles, and vendor marker are combined in one production entry/core pair. | `src/core.ts` and 1002-line `src/index.tsx`; broad root re-export. | High review, upgrade, and rollback cost. | `must_fix` | `packages/grid-adapter` | Split pure semantics, private RDG compilation, orchestration, styles, and the public facade; freeze semantic contracts rather than obsolete DOM. |
| Renderer/editor registries and cleanup are public and tested but unused by production consumers and `GridTable`. | Global symbol search finds use only in core and core tests. | Tests imply capabilities the live renderer does not exercise. | `must_fix` | `packages/grid-adapter` with view-contract/app inputs | Replace them with live semantic render/edit contexts wired to RDG; remove unused registry exports after caller migration. |
| Production and semantic test renderers differ in element roles/controls, collapse, and keyboard behavior. | `src/index.tsx` uses div/CSS grid and collapsible group buttons; `test-support.tsx` uses a table and no collapse state. | Mocked unit tests can pass while production behavior regresses. | `must_fix` | Grid adapter test surface | Share semantic compilation, limit the double to fast non-vendor tests, and require the live component for RDG interaction, accessibility, and visual evidence. |
| Public saved-row shape permits `key` to differ from `recordId`. | `GridRow` exposes both; app builders currently set saved keys from record IDs; validation checks only record IDs. | React identity could drift from mutation identity. | `must_fix` | Grid adapter semantic core | Replace it with explicit committed-record and recordless-draft row types; remove the duplicate committed key. |
| `reconcileRecordRows` sits in the adapter but is called from query/collaboration loaders. | App loaders import it; guide assigns query state to the app. | The vendor seam owns app refresh mechanics and encourages future state leakage. | `must_fix` | `apps/web` workbook query/collaboration state | Move it to an app-owned utility, migrate both loader families, preserve behavior, and remove the adapter export without an alias. |
| Generic renderer contains a Timeline-specific all-rows-mounted workaround; viewport state inputs are currently ignored by row materialization. | `buildVirtualizedRows` comment and implementation in `src/index.tsx`. | Large grids scale poorly and evidence does not exercise real virtualization. | `must_fix` | Grid adapter virtualization seam plus app anchor controllers | Delete the workaround, use fixed-height RDG virtualization, and preserve focus/edit continuity through stable anchors. |
| Direct RDG package/style imports are confined to the adapter. | Global search and `tools/frontend_import_boundaries.json`. | Low while the guard remains. | `intentional/no_action` | Grid adapter and frontend boundary manifest | Preserve `make frontend-import-boundary-check`. |
| Selector/test-id construction remains in `ui-contracts`; adapter consumes the facade. | `gridScrollportClassName()` import and selector helpers/tests. | Low; avoids duplicated lossy selector construction. | `intentional/no_action` | `packages/ui-contracts` | Preserve facade dependency; do not move selector ownership into the adapter. |
| Contract-derived columns and mutation callbacks are built in the app, not inferred by the adapter. | `workbookContractRows.ts`, Timeline/entity/generic/assessment callers. | Low while stable keys remain authoritative. | `intentional/no_action` | `packages/view-contracts` and `apps/web` | Preserve dependency direction and label-independent behavior. |
| No platform, SQL, storage, HTTP, WebSocket, authorization, revision, or generated-code coupling exists in the target. | Target import and token searches are empty for these families. | Low, but future leakage would be severe. | `intentional/no_action` | Existing app/server/contract owners | Keep these concerns outside `packages/grid-adapter`; enforce via review/static checks. |
| Test-only adapter imports are explicitly allowed only from test sources, and authoritative live-adapter rows are guarded against semantic-test substitution. | Frontend import-boundary manifest and phase fixture-accounting script. | Medium if tests or import paths move. | `intentional/no_action` | Harness/frontend boundary owners | Retain the guard and update authored inputs before any generated accounting output. |

No duplicate SQL/storage row logic, hidden revision/projection side effect, misplaced authorization check, or hand-edited generated file was found inside the target. The principal duplication risk is production/test-renderer semantic drift, not duplicated source-state behavior.

### 5.1 Decision-Complete Remediation Register

#### RG-001 — Real RDG Runtime Integration

- **Remediation:** Replace the custom `div`/CSS-grid renderer with `DataGrid` for ordinary workbook surfaces and `TreeDataGrid` when grouping is active. Remove `gridAdapterVendor` and require live component behavior as vendor evidence.
- **Areas:** Implementation, tests, and implementation-support documentation.
- **Rationale:** The current runtime claims RDG ownership without using its component, so package, guide, and evidence statements are materially misleading.
- **Expected long-term benefit:** Layout, virtualization, active-cell, header, resizing, editing, and treegrid mechanics have one maintained vendor implementation behind a Cartulary-owned boundary.
- **Compatibility or migration impact:** RDG DOM structure, geometry, focus choreography, and visual goldens change intentionally. Stable Cartulary selectors, route contracts, storage, and server authorization remain unchanged. No permanent custom-renderer fallback remains.
- **Risk if unresolved:** Future work continues to build on false capability claims, duplicated grid mechanics, unbounded row mounting, and evidence that cannot detect vendor integration regressions.
- **Validation criteria:** A live RDG root renders on every workbook surface; vendor callbacks are exercised; the RDG stylesheet is imported exactly once; functional, stateful, visual, and accessibility grid evidence passes.

#### RG-002 — Internal Package Decomposition

- **Remediation:** Split pure Cartulary semantics, private RDG row/column/event compilation, React orchestration, adapter-owned styles, and a narrow root facade. Export no vendor type or raw imperative handle.
- **Areas:** Implementation and tests.
- **Rationale:** The present core/entry pair couples unrelated change reasons and makes vendor upgrades, semantic changes, and styling changes share one high-risk file.
- **Expected long-term benefit:** Each layer can evolve and be tested independently; RDG can be upgraded or replaced without leaking its coordinate model into app code.
- **Compatibility or migration impact:** Replace `GridTable` with `WorkbookDataGrid`, retain `GridViewport` only as the useful shell, migrate all private in-repo callers atomically, and remove obsolete facade symbols without deprecation aliases.
- **Risk if unresolved:** Review, rollback, and future phase work remain expensive, and vendor APIs can escape through convenience exports.
- **Validation criteria:** Import-boundary checks prove only the adapter imports RDG; pure mapping tests run without a browser; the root facade exposes only approved semantic types and components.

#### RG-003 — Committed and Draft Row Identity

- **Remediation:** Replace `GridRow` with `GridRecordRow` and `GridDraftRow`. Committed rows require non-empty unique `recordId` and explicit `rowVersion`; the RDG key is `recordId`. One optional draft uses a namespaced `draftKey` and RDG bottom-row rendering outside committed grouping, selection, and mutation targeting.
- **Areas:** Implementation, tests, and specification-support documentation.
- **Rationale:** A second committed `key` permits React identity to diverge from mutation identity, while the null-record draft exception is implicit.
- **Expected long-term benefit:** Record position can never become authority, draft handling is explicit, and future collaboration/refresh work has a single stable identity model.
- **Compatibility or migration impact:** Row builders and all grid consumers migrate to the union; the redundant `key` is removed. The development guide clarifies that `rowKeyGetter=recordId` applies to committed rows and that the draft is a recordless affordance.
- **Risk if unresolved:** Reorder, grouping, or refresh can silently retarget selection or mutation state to the wrong record.
- **Validation criteria:** Missing/blank/duplicate committed IDs fail before mutation-capable render; reorder and refresh preserve anchors; the draft never joins a group, selected-record set, or record mutation intent.

#### RG-004 — Live Renderer and Editor Adapters

- **Remediation:** Replace row-only renderer/editor signatures with semantic contexts carrying row data, stable cell target, selection state, and adapter-owned update/close functions. Emit RDG `renderEditCell` only when contract writeability and an explicit editor adapter are both present. Remove unused registry APIs.
- **Areas:** Implementation and tests.
- **Rationale:** Current public helpers are tested but unwired, while Timeline renders editors permanently and bypasses RDG edit lifecycle.
- **Expected long-term benefit:** Display, editing, commit, cancel, cleanup, and field capability have one explicit path that future field families can extend safely.
- **Compatibility or migration impact:** Saved Timeline inputs become display renderers plus editor adapters; draft-create controls remain in the recordless bottom row. App mutation owners continue to submit writes.
- **Risk if unresolved:** Tests keep claiming capabilities the runtime lacks, editability can drift from contracts, and editor cleanup/focus behavior remains renderer-specific.
- **Validation criteria:** Direct typing, Enter/Tab commit, Escape cancel, invalid local timestamp retention, cleanup, renderer precedence, and read-only blocking pass against the live RDG component.

#### RG-005 — Controlled State and Semantic Event Translation

- **Remediation:** Add semantic callbacks for active cell, record selection, sort arrays, column widths, edit intent, copy/paste, fill, and focus exit. Translate every RDG coordinate to `recordId + rowVersion + fieldKey` before calling the app and fail closed for group, draft, missing, read-only, incompatible, or stale targets. Expose only semantic focus, scroll-to-anchor, and scroll-element methods through `GridHandle`.
- **Areas:** Implementation, tests, and implementation-support documentation.
- **Rationale:** RDG-controlled state currently has no explicit Cartulary owner, and app focus/paste helpers reconstruct presentation coordinates outside the integration seam.
- **Expected long-term benefit:** Sorting remains server-owned, mutation intent remains auditable, and vendor-local selection/focus state cannot become authoritative workbook state.
- **Compatibility or migration impact:** App controllers consume semantic callbacks instead of DOM/presentation reconstruction; saved-view layout continues to own durable widths; no raw RDG handle reaches `apps/web`.
- **Risk if unresolved:** Vendor row indexes can leak into writes, hidden state can diverge from saved/query state, and keyboard or fill behavior can retarget after refresh.
- **Validation criteria:** The exhaustive Core/design keyboard matrix, stable multi-cell TSV paste, readable-cell copy, stable-target fill, controlled resize, sort-query dispatch, and focus restoration pass without authoritative local sorting or row mutation.

#### RG-006 — Contract-Backed Treegrid Grouping

- **Remediation:** Use `TreeDataGrid` with one adapter-private group column and a typed descriptor containing authoritative `fieldKey`, canonical scalar extraction, label formatting, and test-ID construction. Encode type-aware lossless group IDs, preserve server row/group order, and keep null as the explicit unassigned bucket. Expansion is client-local, initially expanded, keyed by surface plus grouping field, and reconciled as buckets change.
- **Areas:** Implementation, tests, and documentation.
- **Rationale:** Current grouping uses display strings and custom rows rather than the specified treegrid boundary; several legal grouping fields are not visible columns.
- **Expected long-term benefit:** Group identity is label-independent, accessibility semantics come from a purpose-built treegrid, and later grouping-field growth does not require new DOM machinery.
- **Compatibility or migration impact:** Group-row DOM and goldens change. Drafts remain outside grouped records; grouped vendor drag-fill is disabled while explicit stable-target bulk commands remain available.
- **Risk if unresolved:** Group labels can become accidental identity, group rows can acquire mutation affordances, and accessibility behavior remains custom and incomplete.
- **Validation criteria:** One group level, deterministic order, null handling, keyboard expand/collapse, ARIA treegrid metadata, no draft-created buckets, and no edit/copy/paste/fill mutation on group rows all pass.

#### RG-007 — Real Fixed-Height Virtualization

- **Remediation:** Delete `buildVirtualizedRows`, spacer calculations, viewport state that does not affect materialization, and the Timeline-specific all-rows-mounted workaround. Keep RDG virtualization enabled and derive fixed numeric row heights from generated design tokens through an authored `ui-contracts` projection. Variable-height mode remains unsupported until separately justified and tested.
- **Areas:** Implementation, tests, and design-support documentation.
- **Rationale:** Current code advertises virtualization inputs while mounting every row, contradicting the development guide and impairing scale.
- **Expected long-term benefit:** Large workbook surfaces have bounded DOM cost and use RDG's maintained scroll, frozen-column, measuring, and active-cell machinery.
- **Compatibility or migration impact:** Ordinary cell content must fit fixed rows through truncation, compact chips, or adjacent detail surfaces; custom wrapping behavior is not preserved solely for legacy reasons.
- **Risk if unresolved:** Future phases inherit poor large-result performance and fragile focus workarounds that become harder to remove.
- **Validation criteria:** Bounded rendered-row counts, deep scrolling, active edit/focus continuity, frozen gutter behavior, density modes, and record-keyed interaction after virtualization pass.

#### RG-008 — Live Evidence and Deliberate Test Double

- **Remediation:** Retain `./test-support` only as a lightweight fast double that shares semantic compilation and cannot close RDG interaction, accessibility, or visual rows. Replace string-marker assertions and injected specimen DOM with live-component unit/browser evidence and deterministic production-adapter captures.
- **Areas:** Tests, harness accounting, visual evidence, and documentation.
- **Rationale:** Production and test renderers currently diverge, and synthetic specimens can claim resize, fill, or RDG states that the runtime does not expose.
- **Expected long-term benefit:** Fast app tests remain economical while authoritative evidence exercises the same component users receive.
- **Compatibility or migration impact:** Mapped scenario paths/titles and visual goldens may change; authored maps, accounting, and fixture registries change first, followed by Make-owned ledger/schedule regeneration.
- **Risk if unresolved:** Green tests and goldens can coexist with a broken or absent vendor integration.
- **Validation criteria:** No RDG claim relies on a constant or fabricated vendor class; live tests cover resize, fill handle, editor, grouping, empty state, themes, keyboard, and accessibility; generated accounting passes drift checks.

#### RG-009 — App-Owned Row Reconciliation

- **Remediation:** Move `reconcileRecordRows` to an app-owned workbook query/collaboration utility, migrate both loader families, and remove it from grid-adapter exports and tests.
- **Areas:** Implementation, tests, and tracker ownership.
- **Rationale:** Structural sharing across query refresh and collaboration is app-state behavior, not vendor-coordinate translation.
- **Expected long-term benefit:** The adapter remains cohesive and query/draft/freshness rules evolve beside their actual consumers.
- **Compatibility or migration impact:** This is an in-repo private import move with no compatibility re-export; structural sharing, local draft preservation, freshness, and record-ID lookup remain behaviorally unchanged.
- **Risk if unresolved:** The adapter becomes a catch-all frontend state package and future refresh semantics split across owners.
- **Validation criteria:** Sparse/full refresh and local-draft loader tests pass; unchanged references are retained; changed rows are replaced; no query/collaboration helper remains in the adapter.

#### RG-010 — Specification-Support and Evidence Alignment

- **Remediation:** Mark RB-001 `RESOLVED: real RDG integration selected` and RB-002 `RESOLVED: app-owned reconciliation`. Amend only implementation-support guides needed to define the semantic facade, draft exception, controlled-state mapping, live evidence, and fixed-height virtualization. Do not add vendor requirements to Core 00 through Core 04 or create a vendor-specific NLSpec.
- **Areas:** Specification-support documentation, tracker, and tests.
- **Rationale:** Product behavior must remain vendor-neutral while implementation-support statements must accurately describe the chosen runtime.
- **Expected long-term benefit:** Future implementers have a coherent owner hierarchy without coupling Base Profile conformance to one beta package.
- **Compatibility or migration impact:** Guide and evidence wording changes; product requirements, routes, schemas, storage, authorization, and conformance profiles do not.
- **Risk if unresolved:** Guides continue to overstate runtime capabilities or vendor evidence is incorrectly represented as product conformance.
- **Validation criteria:** Tracker, guides, phase maps, test accounting, and live runtime agree; design/vendor evidence remains correctly classified; Core 05 is not invoked absent a timed or fixture-publication claim.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-R00 | Tracker and authority cleanup | root | none | WF-R01 | Record the selected real-RDG direction, resolved owners, durable target architecture, compatibility posture, and later-authorization boundary. | Tracker only in this session. | Structure, source cross-check, and tracker diff inspection. | Sections 1 and 3 through 12 contain no unresolved RB-001/RB-002 dependency and claim no production change. |
| WF-R01 | Behavior and evidence baseline | chain | WF-R00 | WF-R02 | Freeze Core-owned behavior, stable semantic selectors, and deliberate app contracts without blessing custom DOM, inline grid geometry, or all-rows-mounted behavior. | Package/app tests, browser evidence maps, owner docs. | `make frontend-unit`; live-component red-first assertions within the implementation slice. | Acceptance matrix covers identity, grouping, editing, keyboard, paste/fill, sizing, virtualization, visuals, and accessibility. |
| WF-R02 | Semantic API and ownership migration | chain | WF-R01 | WF-R03 | Introduce committed/draft row types and semantic intents, migrate consumers, and move row reconciliation to app state. | Grid adapter facade/core; workbook row builders, loaders, focus/paste controllers. | `make frontend-typecheck`; `make frontend-unit`; import-boundary check. | No duplicate committed key, no query helper in the adapter, and no vendor type outside the adapter. |
| WF-R03 | RDG rendering foundation | chain | WF-R02 | WF-R04 | Implement private RDG compilation, `DataGrid`/`TreeDataGrid`, controlled sort/width/selection/group state, bottom draft row, stable selectors, adapter styles, and fixed-height virtualization. | Grid adapter runtime/styles; `ui-contracts` density projection; workbook component call sites. | Type-check, unit, functional browser, visual inspection. | Every workbook surface renders through RDG; any temporary private renderer switch and the custom production renderer are removed. |
| WF-R04 | Interaction completion | chain | WF-R03 | WF-R05 | Wire editor adapters, active-cell restoration, exhaustive keyboard behavior, copy/paste, fill, resize, focus exit, and semantic imperative operations. | Grid adapter event compiler/orchestrator; app mutation, keyboard, focus, and paste owners. | Unit, webserver-backed, stateful, and accessibility browser targets. | Every mutation-capable event has a stable Cartulary target; presentation and read-only paths fail closed. |
| WF-R05 | Evidence and documentation alignment | chain | WF-R04 | WF-R06 | Replace synthetic RDG evidence, update live scenarios/goldens, revise support guides, update authored accounting inputs, and regenerate downstream ledgers/schedules. | Package/app/browser tests; guides; authored maps/registries; generated accounting through Make only. | Visual/a11y targets; JSON/generated policy; ledger/schedule drift. | Evidence exercises the production component, classifications remain correct, and generated artifacts match owners. |
| WF-R06 | Final validation and handoff | chain | WF-R05 | none | Run the narrow-to-broad gate, record exact results and failures, update tracker statuses, and publish the implementation handoff. | No new production scope; tracker and retained test artifacts. | `make agent-finalize`; `make check` after all narrow gates. | Custom renderer and obsolete facade are absent; required checks pass or exact related failures and artifacts are recorded. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GA-S01 | WF-R01 | Add or revise characterization around Core behavior, stable selectors, deliberate app contracts, and a live-component smoke assertion. Do not freeze custom DOM hierarchy, CSS-grid templates, fake vendor markers, or unlimited row mounting. | Target tests; selected workbook/browser tests. | Accidental preservation of obsolete implementation details or red tests left outside the implementation slice. | Identity, grouping, edit, keyboard, paste/fill, density, virtualization, selector, visual, and a11y matrix. | `make frontend-unit` | Keep red-first vendor assertions and their implementation in one reviewable slice; revert assertions that claim behavior without an owner. | The matrix distinguishes protected behavior from intentionally replaceable implementation details. |
| GA-S02 | GA-S01 | Introduce `GridRecordRow`, `GridDraftRow`, `GridCellTarget`, semantic interaction intents, `GridHandle`, and the `WorkbookDataGrid` facade; migrate callers and remove redundant/obsolete exports. | Grid adapter; workbook row/column builders and component consumers. | Private API break, draft placement, row-version propagation, selector continuity. | Compile-time consumer coverage; missing/duplicate ID; reorder/refresh/draft isolation. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check` | Migrate the package and all callers atomically; do not retain a deprecated facade. | No caller uses `GridTable`, `gridAdapterVendor`, duplicate committed `key`, raw RDG types, or obsolete registry helpers. |
| GA-S03 | GA-S01 | Move `reconcileRecordRows` to an app-owned workbook row-reconciliation utility and migrate both loader families without changing structural sharing, local drafts, freshness, or ordering. | App workbook utilities/loaders; grid-adapter core/tests. | Refresh ordering, local pending data, row object identity. | FE-I-P3-01 plus sparse/full refresh and local-draft loader cases. | `make frontend-unit`; `make browser-e2e-stateful` when continuity is affected | Revert the import/ownership move as one unit; add no compatibility re-export. | Loader tests pass and no query/collaboration helper remains in the adapter. |
| GA-S04 | GA-S02, GA-S03 | Split private semantics/compiler/orchestration/styles and implement actual `DataGrid` plus fixed-height virtualization, controlled widths/sort/selection, row/cell renderers, gutter/actions columns, bottom draft row, and stable selectors. | Grid adapter; `ui-contracts` density projection; workbook surfaces. | RDG beta API, CSS/layout, selectors, fixed-height content, scrolling, draft semantics. | Live component, density, width, empty, row identity, selector, deep-scroll, and bounded-DOM cases. | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make browser-e2e-webserver-backed` | A private temporary switch is permitted only during development and must be removed before merge; the rollback point is the entire renderer slice. | Every ungrouped surface renders RDG with virtualization enabled and no custom production renderer. |
| GA-S05 | GA-S04 | Add conditional `TreeDataGrid`, the private group column, canonical typed group IDs, controlled local expansion, server-order preservation, null bucket handling, and draft/group mutation exclusion. | Grid adapter grouping compiler/orchestrator; app grouping descriptors; browser tests. | Hidden grouping fields, order stability, ARIA/tree keyboard, group target leakage. | Core 03 grouping matrix; treegrid ARIA; expand/collapse; copy/paste/edit/fill rejection. | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make browser-e2e-a11y` | Revert grouped integration independently while leaving ungrouped RDG foundation reviewable; no custom grouped renderer remains in the completed slice. | Grouped surfaces use one-level `TreeDataGrid` and all presentation-row invariants pass. |
| GA-S06 | GA-S04, GA-S05 | Wire semantic editor contexts, active-cell restoration, keyboard state machine, row selection, copy/paste, fill, sort/resize callbacks, focus exit, and restricted imperative operations; migrate app focus/paste controllers. | Grid adapter event compiler; Timeline and generic/entity controllers/renderers. | Async mutation intent, editor commit/cancel, browser default suppression, grouped restrictions, stale anchors. | Direct typing; Enter/Tab/Escape; range/copy/paste; fill; resize; refresh/reorder focus; closed/read-only state. | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make browser-e2e-a11y` | Roll back by capability slice only when its semantic API is also removed; never fall back to raw vendor coordinates or hidden local authoritative rows. | All mutation-capable paths emit stable targets and every unsupported/presentation path fails closed. |
| GA-S07 | GA-S06 | Limit `./test-support` to fast non-vendor use, replace synthetic RDG evidence with production-component scenarios, update guides/authored accounting/fixture owners, regenerate ledgers/schedules, review visual diffs, then finalize. | Test-support, app/browser tests, guides, maps/registries, Make-generated accounting. | Evidence loss, misleading classification, golden churn, manual generated edits. | Full live RDG functional/stateful/visual/a11y matrix and existing affected workbook cases. | Section 8 ladder; `make agent-finalize`; `make check` | Revert authored accounting/golden changes and regenerate; never repair generated outputs manually. | Evidence proves the production component, generated artifacts agree, exact results are logged, and no unresolved architecture blocker remains. |

All production, test, guide, authored-accounting, and generated-artifact changes in GA-S01 through GA-S07 require a later authorized implementation task. The owner decisions are complete; later work is not blocked on architecture selection.

## 8. Validation Plan

No dedicated public `grid-adapter` target was found. `make frontend-unit` is the current Make-owned package-inclusive unit surface. `browser-e2e-support` exists as an internal helper and is not used here as the primary public recommendation.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make frontend-unit` | Target core/renderer tests plus mapped workbook unit and integration rows | yes | Establish a baseline before movement and rerun after every runtime/test slice. |
| integration | `make browser-e2e-webserver-backed` | Public functional workbook behavior against the webserver-backed stack | no | Required after renderer, paste, query, mutation-binding, or live-workbook interaction changes. |
| e2e/browser | `make browser-e2e-stateful`; `make browser-e2e-a11y-preflight`; `make browser-e2e-a11y`; `make browser-e2e-visual` | Focus/selection/pending continuity; accessibility; intentional visual geometry | no | Run every affected public grid-fragility surface. After reviewing intentional diffs, refresh goldens only through `make browser-e2e-visual-update`, then rerun the non-update visual target. |
| generated/accounting drift | `make generated-artifact-policy-check`; `make json-shape-check`; conditionally `make phase-ledgers`, `make phase-ledger-drift`, `make phase-schedules`, and `make phase-schedule-drift` | Protected generated roots, authored JSON owners, and downstream accounting | no | When scenario paths/titles, maps, classifications, or fixture registries change, edit authored owners first and regenerate through Make; never hand-edit generated ledgers or schedules. |
| import-boundary/static | `make frontend-typecheck`; `make lint-biome`; `make frontend-import-boundary-check` | Type/public facade, authored frontend lint, RDG and test-helper containment | yes | Import-boundary verification also owns the singleton RDG stylesheet check. |
| full check | `make agent-finalize`; `make check` | Harness maintenance followed by the broad developer gate | no | Required before final implementation handoff after narrow checks; use `RESULTS_DIR` only with an eligible retained successful full warm check root. |

Implementation validation order is `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome`, `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`, `make browser-e2e-a11y-preflight`, `make browser-e2e-a11y`, and `make browser-e2e-visual`; conditional accounting regeneration/drift checks follow authored owner changes; `make agent-finalize` and `make check` are the final gates.

For this tracker-only task, validation is limited to direct structure/diff inspection and `git diff --check -- docs/handoffs/grid-adapter-module-refactor-tracker.md`. `make lint-markdown` currently excludes `docs/handoffs/**`, so it is not evidence for this file. No product, frontend, browser, generated-drift, or full validation suite is claimed as executed for this planning-only session.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| GT-001 | Normalize `packages/grid-adapter` to safe label `grid-adapter` and freeze planning-only scope | WF-R00 | DONE | none | Section 1 | Target, output, authority, and only permitted write are explicit. |
| GT-002 | Inventory every tracked target file and ignored install/cache entry | WF-R00 | DONE | GT-001 | Section 2; tracked/ignored file inspection | Every target file is inventoried or explicitly out of scope. |
| GT-003 | Map Core, app, package, contract, design, and harness owners | WF-R00 | DONE | GT-002 | Sections 3 and 4 | Every discovered contract risk has an owner and evidence posture. |
| GT-004 | Diagnose gaps and record decision-complete remediations | WF-R00 | DONE | GT-003 | Section 5 | RG-001 through RG-010 include remediation, areas, rationale, benefit, migration impact, unresolved risk, and validation criteria. |
| GT-005 | Define sequenced remediation workstreams and slices | WF-R00 | DONE | GT-004 | Sections 6 and 7 | Dependencies, validation, rollback, handoff checkpoints, and exit criteria are explicit. |
| GT-006 | Discover the Make-owned validation ladder | WF-R00 | DONE | GT-005 | Section 8; Make help/task guidance | Canonical commands are named without inventing a package target. |
| GT-007 | Establish the behavior and evidence baseline | WF-R01 | TODO | GT-003 | GA-S01; Section 4 | Core-owned behavior and deliberate app contracts are protected without freezing the custom renderer. |
| GT-008 | Decide real RDG integration versus revised custom-renderer guidance | WF-R00 | DECIDED | GT-004 | RB-001; Section 3.1 | Real `DataGrid`/`TreeDataGrid` integration is selected and no permanent custom-renderer fallback is planned. |
| GT-009 | Decide ownership of `reconcileRecordRows` | WF-R00 | DECIDED | GT-004 | RB-002; RG-009 | Reconciliation is assigned to app-owned workbook query/collaboration state with no compatibility re-export. |
| GT-010 | Execute the semantic API, ownership, renderer, interaction, and evidence slices | WF-R02 through WF-R05 | TODO | GT-007, GT-008, GT-009 | GA-S02 through GA-S07; later authorized task | Every authorized slice satisfies its validation and exit criteria; no custom production grid fallback remains. |
| GT-011 | Publish the decision-complete planning handoff | WF-R00 | DONE | GT-001 through GT-009 | Sections 10 through 12 | Another agent can begin an authorized implementation slice without architecture rediscovery. |
| GT-012 | Complete broad validation and implementation handoff | WF-R06 | TODO | GT-010 | Section 8; implementation-session tracker updates | Required checks pass and exact results, residual risks, and the rollback point are recorded. |

`DONE` and `DECIDED` above apply to this planning artifact and its owner decisions. They do not mean that characterization, production refactoring, test migration, guide revision, generated-accounting work, or browser validation occurred in this session.

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Owner direction is recorded on `main` at `a10e14f...`; the tracker remains the only write and all production work requires later authorization. | Inspected: framework, `AGENTS.md`, domain, Core 00-04, guides, design, R09, target code/tests, callers, maps, and Make guidance. Touched: `docs/handoffs/grid-adapter-module-refactor-tracker.md` only. | `sed`, `rg`, `git status`, `git rev-parse`, `date`, tracker patching and inspection | No Core-owner contradiction; real RDG integration and app-owned reconciliation are decision-complete. | None for planning. | Begin GA-S01 only in a separately authorized implementation task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Target is not a backend, transport, persistence, projection-store, revision, or authorization module. | Inspected: target imports; relevant Core architecture/projection/history sections. Touched: tracker only. | Target token/import searches for route, WebSocket, storage, protocol, view-schema, revision, and authorization families | No backend or platform coupling found in the target. | None for planning. | Keep all backend/domain/store behavior out of later grid-adapter slices. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Legitimate grid seam with mixed semantics, custom rendering, styles, vendor claims, and test-double responsibilities; the replacement boundary is approved. | Inspected: every target source/config/test file; Timeline/entity/generic/assessment renderers; focus/paste/load controllers; UI/test helpers; local RDG API. Touched: tracker only. | `rg --files`, `find`, `git ls-files`, `wc -l`, direct reads, import/symbol searches | Target is `WorkbookDataGrid` over real `DataGrid`/`TreeDataGrid`, with private vendor translation, semantic public events, and no custom production fallback. | None for planning. | Authorize GA-S01, then execute GA-S02 through GA-S06 in dependency order. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Core 00-04 remain vendor-neutral; the private package facade is approved for an atomic breaking in-repo migration; semantic selectors remain compatibility inputs, but custom DOM geometry does not. | Inspected: package export map, root/project TypeScript configs, `workbookContractRows.ts`, generated-artifact policy, frontend boundary manifest. Touched: tracker only. | Export/import searches; `git check-ignore`; generated-policy/config inspection | No generated source changes are required by the architecture decision; later evidence-map changes must start at authored owners and regenerate through Make. | None for planning. | Implement the semantic facade in GA-S02 and regenerate only if GA-S07 changes authored accounting owners. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Existing package/unit/browser evidence is mapped; authoritative RDG closure is assigned to live-component unit/browser/a11y/visual tests, while `./test-support` is limited to fast semantic coverage. | Inspected: target tests, workbook tests/imports, Phase 3 maps/ledgers, FE-P3 map/ledger, test accounting, harness boundary scripts. Touched: tracker only. | `make help-all`; `make task-guide ROLE=feature-dev PHASE=phase3`; `make explain-target` for frontend, browser, drift, Markdown, and check targets; evidence inspection | Command discovery only; no product validation suite was executed or claimed passing in this documentation-only task. | None for planning; GA-S01 remains unimplemented. | Establish the baseline and red-first live-component smoke assertion within the first authorized implementation slice. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Target contains no authorization logic; UI visibility/editability must never become authorization. | Inspected: Core 04 authorization section, target sources, caller responsibilities. Touched: tracker only. | Direct owner read; target authorization-token search | Server/app owners remain authoritative; no misplaced check found in target. | None for planning. | Preserve callback-driven, authorization-neutral adapter behavior. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-14T18:41:16-04:00 | Codex `/root`, tracker-remediation-planning session | Architecture planning is decision-complete; production work remains deferred to later authorized slices. | Inspected: tracker against the requested remediations, workstreams, public facade, validation ladder, and repository evidence. Touched: tracker only. | Structural and terminology inspection followed by tracker diff checks; no product suite | RB-001 and RB-002 are resolved. Remaining risks are implementation risks around the pinned beta API, interaction parity, treegrid accessibility, virtualization, and reviewed visual changes—not owner blockers. | None for planning. | Authorize GA-S01 and retain the pre-renderer commit as the initial rollback point; record exact results after each later slice. |

## 11. Decision and Blocker Register

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Should the runtime become a real `react-data-grid` component integration, or should implementation guidance/package claims be revised around the current custom renderer? | The two directions have different dependencies, DOM, keyboard, editing, focus, grouping, virtualization, paste/fill, visual, and accessibility risk. | The architecture/product owner has selected real component integration; GA-S01 supplies a behavior baseline but is not a prerequisite for the direction decision. | `RESOLVED: real RDG integration selected` — ordinary grids use `DataGrid`, grouped grids use `TreeDataGrid`, and the custom renderer has no permanent fallback role. |
| RB-002 | Should `reconcileRecordRows` remain an adapter identity/performance helper or move behind app-owned workbook query/collaboration state? | Ownership affects structural sharing, local-draft preservation, refresh ordering, and the adapter's conceptual boundary. | Live callers and the development-guide ownership model place refresh/query reconciliation in app-owned workbook state. | `RESOLVED: app-owned reconciliation` — move the helper and its tests to the workbook app, migrate both loader families, and add no compatibility re-export. |

There is no `BLOCKED: owner contradiction` entry because no conflicting Core owner requirements were found. There is no remaining architecture or ownership blocker in this tracker. Later execution still requires explicit implementation authorization and must satisfy the slice dependencies and evidence gates in Sections 6 through 8.

## 12. Binary Completion Criteria

This planning tracker is complete only when every criterion below passes:

- [x] Every tracked file in `packages/grid-adapter` is inventoried.
- [x] Ignored `node_modules/**` install artifacts and `tsconfig.tsbuildinfo` are explicitly out of scope.
- [x] Every discovered public contract risk has an owner and test posture.
- [x] RG-001 through RG-010 each state the remediation, affected areas, rationale, long-term benefit, migration impact, unresolved risk, and validation criteria.
- [x] Every workflow has dependencies, validation, and a handoff checkpoint.
- [x] Every proposed slice distinguishes protected product behavior from obsolete custom-renderer details and is marked as requiring later authorization.
- [x] `DataGrid` for ordinary grids and `TreeDataGrid` for grouped grids are the selected runtime integration, with no permanent custom-renderer fallback.
- [x] The semantic facade, committed/draft identity split, vendor containment, controlled-state translation, and app-owned reconciliation target are explicit.
- [x] Canonical validation commands are discovered; the absence of a dedicated public package target is recorded.
- [x] No generated file hand edit is proposed.
- [x] Phase maps and ledgers are treated as evidence accounting, not runtime architecture.
- [x] Core 00 through Core 04 remain vendor-neutral; no vendor-specific NLSpec or Base Profile claim is proposed.
- [x] No contradiction was found; any future contradiction must be written exactly as `BLOCKED: owner contradiction` without choosing a side.
- [x] RB-001 and RB-002 are marked resolved with their owner decisions and implementation consequences.
- [x] The handoff tables identify current state, evidence, commands, blockers, and next action.
- [x] The tracker states that implementation requires a later authorized task and that no production refactor occurred.

These checks complete the remediation plan, not the refactor. GT-007, GT-010, and GT-012 remain the authoritative indication that baseline characterization, implementation, validation, and final implementation handoff have not been completed. GT-008 and GT-009 record decisions only; no production, test, guide, generated-artifact, dependency, or harness change occurred in this tracker-edit session.
