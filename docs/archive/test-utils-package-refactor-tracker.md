# test-utils-package Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Required value |
| --- | --- |
| Target path | `packages/test-utils` |
| Target label | `test-utils-package` |
| Output path | `docs/handoffs/test-utils-package-refactor-tracker.md` |
| Repository state inspected | Branch `main`, commit `df9500fc9dbf92d46410b24566dbb78f63b04269`, ahead of `origin/main` by three commits |
| Current worktree posture | Execution began with this tracker as the only untracked path; final scope inspection finds only the intended authored changes plus the Make-generated topology index |
| Planning status | Authorized remediation execution complete through S-09; final handoff is ready |
| Allowed change in this session | This tracker plus the authored specification, contract, implementation, test, guide, and harness inputs named by the active slice; generated outputs only through Make-owned generators |
| Non-goals | No database, public HTTP/WebSocket, persistence, authorization, dependency, lockfile, unrelated generated artifact, or visual-golden change |
| Later authority | The 2026-08-04 implementation request authorizes S-01 through S-09, including Core 03/Core 04 adoption, observer-only marker behavior, exact selection behavior, and zero-consumer legacy group-marker removal |

The label `test-utils-package` is the lowercase kebab-case normalization of
`packages/test-utils`. It contains no space, path separator, shell
metacharacter, or unsafe filename character.

The target exists. It is a private, route-free browser-test choreography
package. Its existence does not make `test-utils-package` a product bounded
context and does not transfer product behavior from the applicable Core,
module, application, UI, or grid-adapter owner.

### Normative statement classes

This tracker uses NLSpec-style language to make future refactor execution
testable. It does not supersede an adopted owner.

| Statement class | Meaning in this tracker | Omission behavior |
| --- | --- | --- |
| `MUST` / `MUST NOT` | Binding condition for a future slice to satisfy this approved refactor plan. | The slice is incomplete if the condition is not evidenced. |
| `SHOULD` / `SHOULD NOT` | Required default unless the implementation handoff records owner evidence for a narrower exception. | The default applies when no exception is recorded. |
| `MAY` | Optional action whose permitted effect and boundary are stated. | Omission preserves current behavior. |
| `PROPOSED` | Candidate text or interface for a higher authority. It has no normative product force until that owner adopts it. | Current adopted behavior remains authoritative. |
| `TODO:` | Evidence or command cannot yet be resolved from current repository facts. | No implementation may guess the missing fact. |
| `BLOCKED:` | A required authority decision or prerequisite is absent. | The affected slice MUST NOT begin. |
| `requires later authorization` | The slice intentionally changes observable helper or product behavior. | The slice remains `DEFERRED`. |

The authority order is:

1. Adopted subsystem NLSpecs for their exact named scope.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology,
   package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, research, analysis notes, and the planning framework
   as evidence only.

No owner contradiction was found. Core 05 is not applicable because this
tracker publishes no timed or fixture-sensitive claim. Where
`temp/analysis-notes.md` differs from live source, the live source is recorded
and the analysis is retained only as rationale.

### Owner, doctrine, and support documents inspected

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/research/nlspec-spec.md`
- `temp/analysis-notes.md`
- `docs/testing-harness-nlspec.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, especially
  REQ-03-231 through REQ-03-235
- `docs/spec/04_security_deployment_and_conformance.md`, especially AC-364
- `docs/domain.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`
- `docs/research/R04-responsive_browser_spreadsheet_ui_research_memo.md`
- `docs/research/R09-react-data-grid-research-report.md`

### Repository files inspected

Every tracked file under `packages/test-utils` was opened directly. Ignored
install and cache paths were identified separately. Adjacent evidence was
inspected in:

- `packages/grid-adapter/package.json`, `src/SemanticDataGrid.tsx`,
  `src/index.test.tsx`, `src/styles.css`, and the installed pinned
  `react-data-grid@7.0.0-beta.59` implementation;
- `packages/ui-contracts/src/gridSelectors.ts`, its public facade, and selector
  tests;
- `apps/web/src/workbook/components/WorkbookGridControls.tsx` and representative
  E2E consumers;
- `tools/frontend_import_boundaries.json` and
  `tools/harness/static-analysis/tests/test-frontend-import-boundaries.sh`;
- `tools/test_families/package.test_utils.json`,
  `tools/test_families/harness.browser.json`, `tools/test_catalog_owner.json`,
  `contracts/verification/registry.json`, and
  `contracts/verification/owners/package.test_utils.json`;
- root TypeScript project references and Make-owned task guidance.

Implementation is authorized by the 2026-08-04 remediation request. Each slice
MUST update this tracker and pass `make lint-markdown` before the next slice
begins.

## 2. Current-State Repository Inventory

The remediated package contains nineteen authored files. The two ignored
artifact families at the end of the table are inventory exclusions, not
authored package source.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `packages/test-utils/package.json` | Declares private package `@cartulary/test-utils`, dependencies, and exact public subpaths. | `./grid`, `./accessibility`, `./visual`; no root export. | Workspace resolution, `apps/web`, typecheck, import-boundary policy. | `@cartulary/ui-contracts`; Vitest as development dependency. | All package and consumer evidence depends on resolution. | `pnpm-lock.yaml` is tool-managed downstream state. | `package.test_utils` facade. | High | Public subpaths MUST remain exact during this refactor. |
| `packages/test-utils/tsconfig.json` | Source-only TypeScript project extending the repository base. | No runtime surface; includes `src` and emits no files. | Root TypeScript project and frontend typecheck. | `../../tsconfig.base.json`. | `make frontend-typecheck`. | None. | Frontend tooling. | Low | No configuration change is planned. |
| `packages/test-utils/src/accessibility.ts` | Narrow accessibility-evidence facade over focus continuity and marker anchoring. | `assertGridFocusContinuity`, `assertMarkerAnchoredToGridTarget`. | Declared public API and facade-shape evidence. | `focus.ts`, `marker.ts`. | Exact facade-shape and implementation cases. | None. | `package.test_utils`; product accessibility remains owner-local. | Medium | The guides now explicitly reject a generic accessibility-API interpretation. |
| `packages/test-utils/src/browser.ts` | Playwright-independent structural browser types, capability guards, visibility checks, and delay. | Selected types and helpers re-exported by `./grid`. | Target internals and saved-view E2E support through `./grid`. | DOM APIs and timers only. | Package fixtures and saved-view support. | None. | `package.test_utils` browser-capability seam. | Medium | Avoiding a Playwright dependency is intentional. `selectOption` is structurally optional. |
| `packages/test-utils/src/focus.ts` | Retrying, non-mutating focus and viewport-continuity observer. | `assertGridFocusContinuity` through `./grid` and `./accessibility`. | Row-mutation support and both facades. | `browser.ts`, `grid-diagnostics.ts`, `grid-observers.ts`, DOM focus/geometry. | Five cataloged unit cases and live continuity evidence. | UI selectors consumed indirectly. | `package.test_utils` observer choreography; Core 03 owns product continuity. | High | The 50 ms interval, 3000 ms timeout, and continuity rules are preserved. |
| `packages/test-utils/src/grid-actions.ts` | Sort, filter, grouping, paste, expand/collapse, and edit-commit commands. | Action functions and `GridAnchorCommandScenario` are re-exported by `./grid`. | Broad workbook, Timeline, collaboration, accessibility, and visual evidence. | UI contracts, `browser.ts`, `matrix.ts`, `grid-setup.ts`. | Positive, missing-capability, rejection-order, and live browser evidence. | Authored UI selector contracts only. | `package.test_utils` action choreography. | High | Required selection fails closed; filter values retain select-or-fill behavior. |
| `packages/test-utils/src/grid-diagnostics.ts` | Read-only scroll snapshots and bounded virtual-scan diagnostics. | Private types and functions only. | Setup and focus observers. | UI selectors and `browser.ts`. | Virtual-targeting and continuity suites. | None. | `package.test_utils` diagnostic seam. | Medium | Diagnostics exclude raw HTML and incident payloads. |
| `packages/test-utils/src/grid-observers.ts` | Viewport, mounted-row, and owner-backed group-row postconditions. | Observer functions re-exported by `./grid`. | Package facades and live browser evidence. | UI semantic selectors and `browser.ts`. | Package and owner browser rows. | None. | `package.test_utils` observer seam. | High | Observers do not depend on actions or setup. |
| `packages/test-utils/src/grid-setup.ts` | Scroll actions, target preparation, and bounded virtualized scans. | Four scrolling/targeting functions through `./grid` and `./visual`. | Actions, visual evidence, and owner support modules. | UI selectors, `browser.ts`, `grid-diagnostics.ts`. | Virtual-targeting suite and live browser stages. | None. | `package.test_utils` setup seam. | High | Defaults remain 50 ms and 3000 ms. |
| `packages/test-utils/src/grid.ts` | Primary facade over structural capabilities and the private seams. | Exact preserved runtime functions and structural types. | Application E2E/spec/support files, package tests, and static boundary policy. | Browser, action, observer, setup, focus, and marker modules. | Exact facade-shape plus all consumer evidence. | None directly. | `package.test_utils` public facade. | High | No root export or unsupported compatibility barrel was added. |
| `packages/test-utils/src/index.test.ts` | Sole discovered Vitest entry and centralized cleanup. | No production exports; registers 40 exact cases. | The one active `package.test_utils` row and frontend-unit runner. | Four non-discovered registration suites. | Complete package characterization. | Generated topology is downstream of the authored family row. | `package.test_utils` evidence owner. | High | Every title resolves exactly once. |
| `packages/test-utils/src/marker.ts` | Retrying, read-only presence-marker identity and geometry observer. | `assertMarkerAnchoredToGridTarget` through all three subpaths. | Three collaboration and four visual assertions. | UI grid-shell ID, `browser.ts`, `grid-observers.ts`. | Purity, remount, timeout, mismatch, tolerance, and live evidence. | None. | `package.test_utils`; presence behavior remains collaboration/UI-owned. | High | It performs no scrolling or other mutation after invocation. |
| `packages/test-utils/src/matrix.ts` | Converts a matrix to tab/newline-delimited clipboard text. | Private `pasteMatrixText`. | `grid-actions.ts`, package tests. | None. | One cataloged case and indirect paste evidence. | None. | `package.test_utils`. | Low | Deterministic private helper. |
| `packages/test-utils/src/test-suites/{continuity,fixtures,marker,selector-grouping,virtual-targeting}.ts` | Four non-discovered behavior-family registrations plus shared DOM/fake builders. | Test-only registration and fixture functions. | Sole `index.test.ts` entry. | Vitest, UI contracts, package facades/private seams. | Forty exact titles across four families. | Catalog discovery follows explicit registration imports. | `package.test_utils` evidence support. | Medium | Registration modules are not independently discovered as tests. |
| `packages/test-utils/src/visual.ts` | Visual-evidence facade for target setup and marker observation. | Marker assertion plus four setup functions. | `apps/web/e2e/workbook.visual.spec.ts`. | `marker.ts`, `grid-setup.ts`. | Facade-shape, package, and visual browser evidence. | Visual goldens remain unchanged consumers. | `package.test_utils`, collaborating with visual owners. | Medium | Visual evidence does not transfer product-presentation ownership. |
| `packages/test-utils/node_modules/**` | Installed dependencies and workspace links. | Not authored source. | Local tooling. | pnpm-managed state. | Not applicable. | Tool-managed artifact. | Workspace package manager. | Low | Ignored and out of scope. Installed vendor source is implementation evidence only. |
| `packages/test-utils/tsconfig.tsbuildinfo` | TypeScript incremental cache. | None. | TypeScript tooling. | TypeScript compiler. | Not applicable. | Tool-managed cache. | Repository-local tooling. | Low | Ignored and out of scope. |

The 28 application E2E/spec/support importers are:

- `apps/web/e2e/autoresolve.spec.ts`
- `apps/web/e2e/collaboration.spec.ts`
- `apps/web/e2e/evidence-integration.spec.ts`
- `apps/web/e2e/evidence.spec.ts`
- `apps/web/e2e/grid-provenance.spec.ts`
- `apps/web/e2e/history.spec.ts`
- `apps/web/e2e/incident-administration.spec.ts`
- `apps/web/e2e/inspector-actions.spec.ts`
- `apps/web/e2e/keyboard.spec.ts`
- `apps/web/e2e/measurement/timeline-grid.spec.ts`
- `apps/web/e2e/measurement/timingSupport.ts`
- `apps/web/e2e/mentions.lifecycle.spec.ts`
- `apps/web/e2e/mentions.resolve.spec.ts`
- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/e2e/support/collaboration/replay.ts`
- `apps/web/e2e/support/entities/merge.ts`
- `apps/web/e2e/support/workbook/rowMutations.ts`
- `apps/web/e2e/support/workbook/savedViews.ts`
- `apps/web/e2e/timeline-public-route.spec.ts`
- `apps/web/e2e/timeline-query.spec.ts`
- `apps/web/e2e/timeline-workbook.spec.ts`
- `apps/web/e2e/timeline.support.spec.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.assessments.spec.ts`
- `apps/web/e2e/workbook.generic.spec.ts`
- `apps/web/e2e/workbook.spec.ts`
- `apps/web/e2e/workbook.support.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`

`tools/harness/static-analysis/tests/test-frontend-import-boundaries.sh` is a
twenty-ninth textual consumer. It is architecture-policy evidence, not an
application/browser consumer. Workspace resolution and typecheck are additional
configuration consumers, not source importers.

## 3. Module Boundary Diagnosis

The target is a legitimate private test-support facade with directional setup,
action, observer, and diagnostic seams. It is browser-runtime-adjacent and
grid-adapter-adjacent, but it is not a production transport adapter,
persistence adapter, mutation coordinator, frontend controller, or grid-vendor
integration layer. No route, WebSocket, domain payload, authentication,
authorization decision, storage, SQL, revision, projection, generated contract,
or direct `react-data-grid` import exists in the target.

The permanent boundary SHOULD remain a private support package only for
reusable, route-free browser choreography. Product semantics MUST remain with
their adopted owners; selector bytes MUST remain with `packages/ui-contracts`;
vendor normalization MUST remain with `packages/grid-adapter`.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Exact `grid`, `accessibility`, and `visual` facades | Package manifest and facade files | `package.test_utils` | keep | Declared exports and import-boundary policy. | No root export is the default. |
| Structural browser capabilities | `browser.ts`, exposed partly through `grid.ts` | `package.test_utils` | keep | Route-free, payload-agnostic, Playwright-independent types. | A fourth public subpath is not justified. |
| Grid setup and actions | `grid-setup.ts`, `grid-actions.ts` | `package.test_utils` | keep split | Reusable browser choreography. | Facade signatures are preserved; required selection fails closed. |
| Focus, viewport, marker, and semantic postconditions | `focus.ts`, `marker.ts`, `grid-observers.ts` | `package.test_utils` observer layer | keep split | Helpers read owner-backed semantic DOM and geometry. | Observers do not manufacture their postconditions. |
| Bounded diagnostic state | `grid-diagnostics.ts` | `package.test_utils` diagnostic layer | keep split | Virtual scanning reports observed offsets, growth, and visibility. | Diagnostics remain bounded and payload-neutral. |
| Selector and test-ID construction | `@cartulary/ui-contracts` imports | `package.ui` | keep | Shared authored builders already exist. | Target MUST import, not duplicate, stable selector bytes. |
| Presentation-only group semantics | `grid-observers.ts`; UI semantic selector; grid-adapter/vendor DOM | Core 03, `package.grid_adapter`, `package.ui` | aligned | Core 03/04 now own the treegrid, parent-row, expansion, accessible-label, and toggle contract. | Legacy group-only mutations have zero consumers and are removed. |
| Product route, mutation, revision, projection, saved-view, authorization, and storage behavior | Only exercised indirectly by consuming tests | Applicable backend/application owners | keep outside | No implementation exists in target. | Evidence use does not transfer ownership. |
| Accessibility helper inventory | Exact narrow facades and corrected implementation guides | Guide owner plus individual product owners | aligned | No owner-backed requirement for generic wrappers exists. | The binary admission rule defaults unproven assertions to live owners. |
| Grid-vendor integration | `packages/grid-adapter` only | `package.grid_adapter` | keep outside | Direct-import policy and source inspection. | No vendor import crosses into test-utils. |

## 4. Public Contract and Behavior Freeze Map

The completed remediation preserves the public contract except for the
explicitly authorized semantic corrections recorded below.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Exact public subpaths and no root export | `package.test_utils` | `package.json`, facades, import-boundary policy | Static policy and indirect imports | Direct API-shape import test for each subpath | High | Paths and exports are frozen. |
| Structural browser types and capability guards | `package.test_utils` | `browser.ts`, saved-view support | Unit fixtures and typecheck | Missing-capability and failure-order tests | Medium | No Playwright dependency is permitted. |
| Shared selector builders | `package.ui` | `gridSelectors.ts`, policy tests | UI-contract and package tests | Facade tests MUST compare imported builders, not copied literals | High | Exact bytes remain outside test-utils. |
| Virtual target scan and diagnostics | `package.test_utils` | `grid-setup.ts`, `grid-diagnostics.ts` | Accounted package cases plus live browser evidence | Hydration, remount, growing-range, portal, deadline, and final-snapshot cases are preserved | High | S-01 repaired accounting before S-02/S-03 movement. |
| Focus continuity | Core 03 product behavior; package observer | REQ-03-283, `focus.ts` | Five package cases and live mutation evidence | Preserve default and strict-scroll branches plus purity | High | Observer does not restore focus or scroll. |
| Marker anchoring | Collaboration/UI owners; package observer | `marker.ts`, seven live calls | Purity, remount, mismatch, tolerance, browser, and visual rows | Read-only purity, remount, mismatch, and tolerance boundaries are covered | High | S-06 is implemented; every supported caller prepares before the marker-producing action. |
| Presentation-only group rows | Core 03, grid adapter, UI contracts | REQ-03-231..235, AC-364, shared selector, and adapter binding | UI-contract, adapter, package, product, accessibility, and visual rows | Production-binding row semantics, selector mapping, and keyboard/pointer equivalence pass | High | S-05 adopted owner-backed ARIA semantics and removed zero-consumer legacy group mutations. |
| Filter and grouping actions | Application/view owners; package choreography | `grid-actions.ts`, live controls | Negative ordering, branch, rejection, package, and five-stage browser evidence | Required-selector failure, value-control branch, call order, and rejection are covered | High | S-08 is implemented with fail-closed required selection and explicit value fallback. |
| TSV paste formatting | Product write owners; package formatting helper | `matrix.ts`, paste callers | One package unit case and live paste evidence | Preserve tab/newline bytes and existing product evidence | Medium | No route or mutation payload is built here. |
| Harness ownership and exact title accounting | Testing Harness NLSpec and authored family manifests | One active row names all 40 current tests | Harness validation, focused slice, and cumulative frontend unit evidence | Every current title resolves exactly once with no overlap | High | Phase/evidence rows are accounting only; later characterization titles were added to the same row. |
| HTTP, WebSocket, saved-view, mutation, revision, projection, authorization, and generated contracts | Applicable product owners | Consumer tests and architecture policy | Product-owned evidence | Only exact affected owner rows after impact discovery | High but indirect | Target MUST remain route-free and payload-agnostic. |

### Explicit current defaults

| Interface | Omitted-input default | Required invariant |
| --- | --- | --- |
| `assertGridFocusContinuity.intervalMs` | `50` ms | Negative effective retry delay is not introduced. |
| `assertGridFocusContinuity.timeoutMs` | `3000` ms | Existing retry and timeout diagnostics remain stable. |
| `allowContainingGridCell` | `false` | The exact target retains focus unless explicitly opted into the containing cell. |
| `requireExactHorizontalScroll` | `false` | Visibility-preserving horizontal movement is allowed by default. |
| `requireExactVerticalScroll` | `false` | Visibility-preserving vertical movement is allowed by default. |
| `scrollGridTargetIntoView.intervalMs` | `50` ms | Scans remain bounded by the effective deadline. |
| `scrollGridTargetIntoView.timeoutMs` | `3000` ms | Failure emits bounded scan diagnostics. |
| Marker geometry tolerance | `2` px | Row/cell identity remains mandatory regardless of tolerance. |
| `BrowserLocator.selectOption` | Capability may be absent | A required selection action fails closed; a declared polymorphic value input follows its explicit select-or-fill mapping. |

### Adopted group-row owner interface

Core 03 REQ-03-231/232 and Core 04 AC-364 now adopt the following product
contract.

When `group_by` is active, the workbook grid MUST expose `role="treegrid"`.
Every visible derived group header MUST be a parent `role="row"` with
`aria-level="1"` and row-level `aria-expanded="true"` or
`aria-expanded="false"` matching client-local expansion state. Ordinary
record, draft, loading, summary, and other non-parent rows MUST NOT acquire
`aria-expanded` solely because grouping is active. A group header MUST preserve
its bucket label for accessible-name computation, expose exactly one visible
keyboard-operable toggle, and remain recordless, non-editable, non-pasteable,
non-exportable, absent from history, and ineligible as a mutation target.

The authored UI-contract function is
`gridGroupRowSelector(expanded?: boolean)`. It MUST be distinct from and MUST
NOT change the existing
`gridGroupRowsSelector(viewSchemaId: string, fieldKey: string)` test-ID-prefix
builder.

| `expanded` input | Exact output |
| --- | --- |
| omitted / `undefined` | `[role="row"][aria-level="1"][aria-expanded]` |
| `true` | `[role="row"][aria-level="1"][aria-expanded="true"]` |
| `false` | `[role="row"][aria-level="1"][aria-expanded="false"]` |

The selector MUST NOT use `.rdg-*`, bucket content, record IDs, field
values, row positions, or visible-label interpolation. A caller needing one
bucket MUST combine the structural contract with the existing owner-built test
ID or an accessible locator. `packages/test-utils` MUST consume the shared
builder instead of copying its bytes.

S-05 qualification: the pinned vendor supplies `treegrid`, row role, row level,
and row expansion state, while the adapter toggle independently supplies
`aria-expanded`. Initial semantic validation passed before the compatibility
bridge was removed. Fresh repository scans found no supported consumer of the
group-only `data-grid-row-kind`, `data-grid-primary-state`, or ref-added styling
class mutations. The adapter now styles group rows through the owner-approved
structural selector and no longer publishes those legacy markers.

### Observer-only marker interface for S-06

After invocation, `assertMarkerAnchoredToGridTarget`:

| Capability class | Permitted operations | Forbidden operations |
| --- | --- | --- |
| Locator resolution | Resolve marker, target, row, cell, and grid-shell locators. | Remount or otherwise prepare a target. |
| Waiting | Retry, wait for visibility, and tolerate ordinary remounts without mutation. | Scroll to rescue an unmounted or offscreen target. |
| Observation | Read attributes, visibility, bounding rectangles, row identity, and field identity. | Focus, click, press, fill, select, or dispatch input/pointer/keyboard/clipboard/scroll events. |
| Diagnostics | Emit bounded identity and geometry diagnostics. | Include raw HTML or incident payloads. |

Every caller MUST prepare the target before the action that produces the marker,
trigger the owner behavior, wait for the owner-defined consequence, and only
then invoke the assertion. Moving the scroll to immediately before the
assertion after the trigger does not satisfy this contract.

### Shared-helper admission rule

A new public test-utils capability MAY be admitted only when every row is true.
If any row is false or unproven, the assertion remains local to the owner test
or owner-specific support module.

| Admission condition | Binary evidence |
| --- | --- |
| A named owner-backed postcondition exists. | Owner requirement and verification row are identified. |
| Reuse or protocol complexity justifies sharing. | Two independent owner rows use equivalent choreography, or one unsafe complex protocol is demonstrated. |
| The helper is route-free and payload-agnostic. | Import and architecture-policy scan passes. |
| No controller, app registry, authentication state, control token, or vendor import enters the package. | Import-boundary evidence passes. |
| Required browser capabilities and omission behavior are explicit. | Interface table and negative tests exist. |
| Public subpath placement is authorized. | Export-map review identifies the exact subpath. |
| Package semantics are characterized. | Focused package tests pass. |
| Product behavior remains live-owner evidence. | Product/adapter row remains assigned to its actual owner. |

### Select-action mapping for S-08

| Action/control | Capability rule | Missing or rejected capability | Subsequent action rule |
| --- | --- | --- | --- |
| `changeGrouping` grouping control | `selectOption` is required. | Missing capability throws immediately; rejection propagates. | No later action occurs. |
| `applyFilterChip` field control | `selectOption` is required and MUST be guarded before opening or mutating the filter workflow. | Missing capability throws immediately; rejection propagates. | Popover click, value entry, and apply MUST NOT occur. |
| `applyFilterChip` value control | The live owner permits a boolean `<select>` or text `<input>`. Attempt `selectOption` when available; absence or rejection selects the explicit `fill(value)` fallback. | If `fill` also rejects, that rejection propagates and apply does not occur. | Apply occurs exactly once only after one input path succeeds. |

`BrowserLocator.selectOption` remains optional. Public function signatures remain
unchanged. No new public error class is authorized.

## 5. Coupling and Boundary Findings

The first nine rows record the execution baseline and their durable resolution;
the remaining rows record boundaries intentionally preserved.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Five live unit titles were absent from exact owner accounting. | Baseline: 31 live cases; 26 titles in the sole active family row. | Focused evidence could omit behavior and could not safely baseline movement. | `must_fix` | Testing Harness authored family manifest | Resolved in S-01 before test movement; the current 40-case row is exact. |
| Group identification relied on a duplicated data literal and post-render adapter mutation. | Baseline helper, adapter toggle ref, and pinned vendor `GroupRow`; no authored semantic row selector. | Helper/adapter drift and an accidental private contract. | `must_fix` | Core 03/04, `package.grid_adapter`, `package.ui`, `package.test_utils` | Resolved in S-05 through adopted ARIA semantics, a shared selector, and zero-consumer legacy removal. |
| Marker assertion scrolled after invocation. | Baseline marker helper called target setup after assertion entry. | Evidence could manufacture the postcondition it claimed to observe. | `should_fix` | `package.test_utils` with collaboration/visual owners | Resolved in S-06 with a pure observer and atomic caller-order migration. |
| Required grouping and filter-field selection could silently no-op. | Baseline optional `selectOption` calls in required action paths. | Named actions could report completion without performing selection. | `must_fix` | `package.test_utils` | Resolved in S-08 with fail-closed capability checks and negative-order evidence. |
| Filter values are intentionally polymorphic. | Live UI renders boolean `<select>` or text `<input>` under one test ID. | Over-broad fail-closed logic would break text filters; silent omission would break fakes. | `must_fix` | Application control contract and `package.test_utils` | S-08 preserves the explicit select-or-fill mapping in Section 4. |
| One 1304-line suite contained four behavior families. | Baseline four `describe` blocks in `index.test.ts`. | Movement and evidence review were unnecessarily coupled. | `should_fix` | `package.test_utils` | Resolved in S-02 with one discovered entry and four registration modules. |
| Setup, actions, observers, and diagnostics shared one private implementation graph. | Baseline `grid.ts`, mixed helpers, `focus.ts`, and `marker.ts`. | Changes could unintentionally mix mutation and observation. | `should_fix` | `package.test_utils` | Resolved in S-03 with acyclic private seams behind unchanged facades. |
| Unused internals and a placeholder remained. | Baseline `resizeGridColumn`, `assertAnchorTestId`, and `.gitkeep`. | Misleading capability and maintenance noise. | `should_fix` | `package.test_utils` / repository hygiene | Resolved in S-04 after refreshed zero-consumer proof. |
| Guide language implied a broader accessibility API than the live package. | Baseline development guide versus exact export map. | Future agents could invent wrappers and ownership. | `should_fix` | Guide owner | Resolved in S-07 with exact subpaths and the binary admission rule. |
| Structural browser types avoid Playwright. | No Playwright dependency or import in the target. | Direct dependency would enlarge the boundary. | `intentional/no_action` | `package.test_utils` | Preserve structural typing. |
| Selector construction remains in UI contracts. | All target selector imports use `@cartulary/ui-contracts`. | Duplicated bytes would drift. | `intentional/no_action` | `package.ui` | Preserve shared ownership. |
| No vendor import crosses the adapter boundary. | Import policy and target scan. | Direct import would couple tests to RDG. | `intentional/no_action` | `package.grid_adapter` | Preserve boundary and static evidence. |
| HTTP, WebSocket, storage, authorization, revision, projection, saved-view, and domain behavior are absent. | Direct target scan. | Moving such behavior here would create a catch-all. | `intentional/no_action` | Applicable product owners | Keep outside every target slice. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Revalidate commit, worktree, authority, and write boundary. | Tracker only during planning. | `make lint-markdown` for tracker revisions. | Source fingerprint and planning boundary recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for every target path, caller, dependency, test, and artifact exclusion. | Target and direct consumers. | Read-only Git/search inspection. | Section 2 matches live repository. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05, WF-07 | Freeze facades/defaults and separate product from helper ownership. | Manifests, facades, owner docs, UI contracts, adapter evidence. | Typecheck and policy commands documented. | Section 4 contains no owner invention. |
| WF-03 | Characterization and accounting | chain | WF-02 | WF-05, WF-06 | Repair exact test accounting before movement and define behavior-family split. | Package tests and authored family manifest. | Focused owner slice plus drift/policy checks. | All 31 titles select exactly once. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Classify browser, grid, guide, product, and vendor coupling. | Target plus adjacent packages. | Import-boundary and architecture-policy rows. | Section 5 classifications remain evidence-backed. |
| WF-05 | Private facade/ownership redesign | chain | WF-02, WF-03, WF-04 | WF-06 | Define private action, observer, and diagnostic seams behind frozen facades. | Target source/tests. | Package slice, typecheck, frontend unit/static checks. | Public surface is byte-for-byte equivalent. |
| WF-06 | Slice sequencing and authorized corrections | chain | WF-05 | WF-08 | Execute structural slices, then separately authorized semantic/behavior slices. | Target and explicitly affected owners only. | Per-slice commands in Section 7. | Each slice has its own rollback and evidence. |
| WF-07 | Guide and harness planning | parallel | WF-02 | WF-08 | Correct guide placement and maintain exact evidence accounting. | Guide and authored harness inputs in later tasks. | Markdown, harness, drift, and policy checks. | No guide or generated output becomes behavioral authority. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Reconcile evidence, inspect scope, and record truthful results. | Files changed by authorized slices. | `make agent-finalize`, narrow gates, then `make check` when authorized. | Final handoff lists results and remaining gates. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | WF-03 inventory complete | Add the five existing virtual-targeting titles to the existing active selector row in ASCII order. Do not rename, move, or change a test or any row metadata. | `tools/test_families/package.test_utils.json`; Make-generated downstream artifacts only. | Selector overlap, stale generated topology, unintended row identity change. | Preserve all 31 tests exactly. | `make explain-test-owner OWNER=package.test_utils`; `make generate`; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make test-slice OWNER=package.test_utils`. | Restore the authored title array and regenerate; never hand-edit generated outputs. | All 31 current titles resolve exactly once, overlap no active row, and pass retained focused evidence. |
| S-02 | S-01 | Keep `index.test.ts` as the sole Vitest entry and family-row selector. Extract four non-`*.test.ts` registration modules under `src/test-suites/`: `selector-grouping.ts`, `continuity.ts`, `virtual-targeting.ts`, and `marker.ts`. The entry imports and invokes them without changing exact titles or behavior. | Target test entry and new test-only suite modules; the S-01 family row remains unchanged. | Accidental independent test discovery, registration-order drift, title loss, fake-timer leakage. | Preserve the exact 31-title set, current order where observable, hooks, timers, fakes, and behavior-family coverage. | `make test-slice OWNER=package.test_utils`; `make frontend-unit`; `make frontend-typecheck`. | Restore the monolithic entry; no catalog rollback is needed because selector identity remains fixed. | One cataloged entry registers four coherent suites; the exact row ID, selector file, titles, owner, verification, evidence, runtime, resource, fixture, tier, claim posture, and status remain unchanged. |
| S-03 | S-02 | Split private implementation into browser capabilities, setup/actions, observers, and diagnostics behind unchanged facades. | Authored target source/tests only. | Export omission, cycles, error or timing drift. | Direct facade shape plus all S-02 suites. | `make frontend-typecheck`; focused package slice; `make frontend-unit`; import-boundary check. | Revert one extraction at a time using facades as rollback seams. | Three public subpaths, signatures, defaults, errors, and behavior are unchanged; private graph is acyclic. |
| S-04 | S-03 | Remove `.gitkeep`, `resizeGridColumn`, and `assertAnchorTestId` after refreshed zero-consumer proof. | Target placeholder, private helper, tests. | Hidden deep import or accidental facade removal. | Preserve facade shape and direct shared-selector characterization. | Typecheck, focused package slice, import-boundary check. | Restore any item with a supported consumer. | Zero supported consumers and no public surface change. |
| S-05 | S-01 | Adopt Core 03 semantics and Core 04 binary evidence, add the shared semantic selector, characterize production binding, migrate test-utils from copied row-kind bytes, then remove zero-consumer legacy group-row mutations after the semantic cutover passes. | Core 03/04, `packages/ui-contracts`, `packages/grid-adapter`, target helper/tests, authored owner rows. | Duplicate ARIA state, vendor lifecycle, post-render mutation, selector drift. | Root/row ARIA, accessible label, one toggle, pointer/keyboard equivalence, record-row omission, non-record restrictions, helper production binding. | Owner slices for `package.ui`, `package.grid_adapter`, `package.test_utils`; typecheck, frontend unit/static, exact browser rows, `make browser-e2e-a11y`, and `make browser-e2e-visual`. | Retain legacy markers only through the initial semantic cutover; after zero-consumer proof remove the group-only marker/ref mutations in the same authorized slice. Roll back owner text and all projections atomically if the semantic binding fails. | Live DOM, shared selector, helper, owner text, and evidence agree; the unsupported legacy group-row data/ref mutations have zero consumers and are removed. |
| S-06 | S-03 | Prepare targets before marker-producing actions and make marker assertion read-only across all seven live invocations. | `marker.ts`, package tests, collaboration and visual callers, affected owner rows. | Offscreen failures, real continuity defects, screenshot changes, reordered collaboration timing. | Purity instrumentation of every mutating capability; row/cell mismatches; two-pixel geometry; remount retry; seven live call sites. | Focused package slice; exact affected owner rows; `make browser-e2e`; `make browser-e2e-visual`; a11y only if an a11y caller changes. | Restore assertion and all caller ordering together; do not update goldens to hide failure. | Assertion waits/reads only, every caller prepares before trigger, and live evidence passes without unauthorized golden changes. |
| S-07 | WF-02 | Correct guide wording to distinguish product accessibility, direct owner assertions, and admitted shared choreography; add the admission rule. Add no package API. | Development/frontend guides only in a later docs-authorized task. | Guide accidentally presented as behavior authority or API expansion. | Preserve facade shape and owner-row mapping. | `make lint-markdown`; typecheck/static only if executable references change. | Restore prose if an adopted owner contradicts it and record `BLOCKED: owner contradiction`. | Guide names exact current facades, admission is binary, and no generic keyboard/accessibility wrapper is required. |
| S-08 | S-03 | Guard required field/group selection and implement the explicit filter-value select-or-fill mapping from Section 4. | `browser.ts`, `grid-actions.ts`, package tests, direct consumers/fakes. | Partial UI action before failure, text-filter regression, swallowed rejection, changed diagnostics. | Required-capability failure order, no later action, select success, missing/rejected select fallback to fill, both-path failure, live filter/group flows. | Focused package slice; typecheck; frontend unit/static; exact affected browser rows. | Revert guard and choreography together; preserve public types/signatures. | Required selection fails before mutation; value entry follows exactly one successful input path; apply occurs only after success. |
| S-09 | S-04, S-07, and every authorized S-05/S-06/S-08 | Run cumulative validation, reconcile owner accounting, inspect scope, and append implementation handoff. | All files changed by prior authorized slices. | Stale generated output, skipped owner row, out-of-scope diff. | Preserve all owner-aligned characterization and live rows. | `make agent-finalize`, Section 8 gates, and `make check` when authorized. | Roll back the last failing slice to its checkpoint. | Required gates pass or have an explicit related failure record; generated drift and scope are clean; handoff is complete. |

### Slice execution status

| Slice | Status | Latest evidence | Next gate |
| --- | --- | --- | --- |
| S-01 | DONE | 31 exact titles; focused owner slice and JSON/generated checks pass. | S-02 |
| S-02 | DONE | One discovered entry registers four non-discovered suites; focused, typecheck, and 359-unit frontend evidence pass. | S-03 |
| S-03 | DONE | Private setup/action/observer/diagnostic seams and exact facade-shape evidence pass all required package, architecture, type, unit, and drift gates. | S-04 |
| S-04 | DONE | Fresh repository scan proves zero supported consumers; the placeholder and two misleading private helpers are removed without facade or title drift. | S-05 |
| S-05 | DONE | Core 03/04, the shared selector, adapter DOM, test-utils observer, package owners, relevant product rows, and broad a11y/visual evidence agree; legacy group-only mutations are removed. | S-06 |
| S-06 | DONE | Marker anchoring is a retrying read-only observer; all seven callers prepare on the observing page before marker-producing remote focus and pass focused, owner, functional, and visual evidence. | S-07 |
| S-07 | DONE | Both implementation guides name the exact three subpaths, delimit their current capabilities, preserve direct owner assertions, and enforce the binary admission rule. | S-08 |
| S-08 | DONE | Required field/group selection fails closed with ordered failure; value selection uses one successful select-or-fill path and applies exactly once; all five browser stages pass. | S-09 |
| S-09 | DONE | All cumulative generated, static, focused-owner, product-owner, five-stage browser, finalizer, 714-check, scope, and tracker gates pass. | Complete |

## 8. Validation Plan

Execution is complete through all cumulative code and browser gates. The table
records the final evidence posture; per-slice roots remain in Section 10.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Markdown, including this tracker | no | Completion checkpoint passes at `.cartulary/test-results/20260804T211307Z-p3598640`. |
| unit | `make test-slice OWNER=package.test_utils` | Exact package owner row | yes | PASS, 40/40 at `.cartulary/test-results/20260804T204713Z-p3165574`. |
| unit, broad frontend | `make frontend-unit` | Frontend and harness-node Vitest projects | no | PASS, 359/359 at `.cartulary/test-results/20260804T204713Z-p3165843`. |
| integration | Exact rows under `platform.viewquery`, `module.timeline`, `module.savedviews`, `module.collaboration`, and `module.workbook` | Product owners using changed helpers | no | PASS; exact retained roots are recorded in the final S-09 handoff row. |
| e2e/browser | Functional, webserver-backed, and stateful Make targets | Browser evidence | no | PASS, respectively 53/53, 62/62, and 36/36. |
| e2e/accessibility | `make browser-e2e-a11y` | Live accessibility evidence | no | PASS, 14/14; no accessibility artifact change. |
| e2e/visual | `make browser-e2e-visual` | Visual browser evidence | no | PASS, 14/14; no golden change. |
| e2e/measurement | `make browser-e2e-measurement` | Measurement evidence | no | Not run: no measurement consumer or Core 05 claim changed. |
| generated drift | `make generate-drift` | Generated outputs against authored owners | no | PASS at `.cartulary/test-results/20260804T204713Z-p3165215`. |
| generated policy | `make generated-artifact-policy-check` | Generated-root and authored-boundary policy | no | PASS at `.cartulary/test-results/20260804T204713Z-p3165257`. |
| JSON shape | `make json-shape-check` | Authored/generated JSON contracts | no | PASS at `.cartulary/test-results/20260804T204713Z-p3165359`. |
| import-boundary/static | `make frontend-import-boundary-check` | Vendor isolation, facades, test/runtime direction | yes | PASS at `.cartulary/test-results/20260804T204713Z-p3166007`. |
| route/payload policy | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131` | Route/payload exclusion and facade policy | yes | PASS at `.cartulary/test-results/20260804T204713Z-p3165554`. |
| typecheck | `make frontend-typecheck` | Workspace TypeScript graph | yes | PASS at `.cartulary/test-results/20260804T204713Z-p3165776`. |
| owner discovery | `make explain-test-owner OWNER=<owner-id>` | Narrow owner verification selection | yes | PASS for `package.test_utils`, `package.ui`, and `package.grid_adapter`; product rows were selected exactly. |
| finalizer | `make agent-finalize` | Harness-maintenance artifacts | no | PASS at `.cartulary/test-results/20260804T204635Z-p3160239`; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| full check | `make check` | Developer verification gate | no | PASS, 714/714 at `.cartulary/test-results/20260804T210306Z-p3470025`. |

## 9. Top-Level Work Tracker

Only `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED` are
valid statuses.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target, safe label, scope, authority, and normative statement classes | WF-00 | DONE | none | Section 1 | Planning authority and omission behavior are explicit. |
| T-002 | Inventory every tracked target and ignored artifact family | WF-01 | DONE | T-001 | Section 2 | Nineteen authored files and two exclusions are accounted. |
| T-003 | Map facades, 28 application importers, static policy consumer, dependencies, and owners | WF-02 | DONE | T-002 | Sections 2-4 | Every direct and indirect contract has an owner and test posture. |
| T-004 | Correct stale group-row repository findings | WF-04 | DONE | T-002 | Adapter, pinned vendor, target helper, Section 4 | Current data marker and vendor ARIA behavior are accurately distinguished from owner authority. |
| T-005 | Repair five missing package titles in owner accounting | WF-03 | DONE | T-003 | RB-001, S-01; run root `.cartulary/test-results/20260804T194247Z-p2467255` | All 31 tests select exactly once and pass. |
| T-006 | Split package tests by behavior family behind one cataloged entry | WF-03 | DONE | T-005 | S-02; focused root `.cartulary/test-results/20260804T194736Z-p2478155` | Four non-discovered suite modules preserve the exact entry file, row identity, titles, owner posture, and behavior. |
| T-007 | Separate private actions, observers, and diagnostics | WF-05 | DONE | T-006 | S-03; focused root `.cartulary/test-results/20260804T195522Z-p2526270` | Frozen facades remain exact and private graph is acyclic. |
| T-008 | Remove placeholder and unsupported dead internals | WF-06 | DONE | T-007 | S-04; focused root `.cartulary/test-results/20260804T200119Z-p2589240` | Refreshed zero-consumer proof and focused validation pass. |
| T-009 | Adopt and implement group-row semantic proposal | WF-06 | DONE | T-005 | RB-002, S-05; adapter owner root `.cartulary/test-results/20260804T200519Z-p2602152` | Core owner text, projections, live evidence, and zero-consumer legacy cleanup agree. |
| T-010 | Make marker assertion observer-only | WF-06 | DONE | T-007 | RB-003, S-06; collaboration root `.cartulary/test-results/20260804T201940Z-p2817587` | All seven call sites prepare before trigger and observer-only evidence passes. |
| T-011 | Correct guide responsibility language | WF-07 | DONE | T-003 | RB-004, S-07; Markdown root `.cartulary/test-results/20260804T202850Z-p2945180` | Guide states exact API and binary admission rule without inventing behavior. |
| T-012 | Implement exact select-capability behavior | WF-06 | DONE | T-007 | RB-005, S-08; focused root `.cartulary/test-results/20260804T203146Z-p2954401` | Action mapping passes negative and live evidence. |
| T-013 | Discover exact product owner rows for changed consumers | WF-08 | DONE | Any authorized behavior slice | Exact `platform.viewquery`, `module.timeline`, `module.savedviews`, `module.collaboration`, and `module.workbook` rows in Section 10 | Per-change selection is exact and owner-aligned. |
| T-014 | Run cumulative validation and implementation handoff | WF-08 | DONE | T-008, T-011, and every authorized correction | Retained run roots and Section 10 | Required gates and the final tracker checkpoint pass; scope is clean. |
| T-015 | Revise tracker in NLSpec voice and close planning decisions | WF-00 | DONE | T-001 through T-004 | This file | Stable decisions, interfaces, defaults, mappings, slices, and acceptance IDs are present. |

## 10. Session Handoff Log

Historical rows are retained even when a later freshness pass corrects their
repository interpretation.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | Scope, authority, branch/commit, safe label, and write boundary recorded. | Inspected framework, Core 00/Core 03, domain, Testing Harness NLSpec, frontend/dev guides; touched only this tracker. | `sed`, `git status`, `git rev-parse`, `date` | Tracker created from live evidence; no owner contradiction found. | None for tracker completion. | Obtain separate authorization before S-01 or any implementation. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | Normative classes, authority limits, explicit defaults, and proposal posture added. | Inspected NLSpec doctrine, analysis notes, Core 03/04, guides, research, repository source; touched only this tracker. | `git`, `stat`, `rg`, `sed`, `wc`, `jq`, `node`, `readlink`, `make explain-target`, `make lint-markdown` | Repository truth supersedes non-authoritative analysis; no owner contradiction found; Markdown passed at `.cartulary/test-results/20260804T191843Z-p2445899`. | Implementation remains separately authorized. | Begin later implementation with repository freshness and S-01. |
| 2026-08-04T15:40:13-04:00 | Codex / authorized remediation activation | S-01 through S-09 are authorized as serial workstreams; S-05 adoption and cleanup plus S-06/S-08 behavior changes are explicit. | Refreshed branch, commit, worktree, tracker decisions, statuses, and write boundary; touched only this tracker. | `date`, `git status`, `git rev-parse`, `apply_patch` | Source remains `main` at `df9500fc9dbf92d46410b24566dbb78f63b04269`; tracker is the sole untracked path. | None. | Validate tracker Markdown, then begin S-01. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | Backend behavior is not implemented by this target; routes, storage, auth, revisions, projections, and domain payloads remain out of scope. | Inspected target source and architecture policy; touched only this tracker. | `rg`, `sed` | No backend, SQL/storage, route, WebSocket, auth, revision, or projection implementation found. | None; any future discovery must be mapped to its product owner rather than moved here. | Keep backend modules excluded from behavior-preserving package slices. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | Backend exclusion is restated as a normative package boundary. | Rechecked target imports and indirect contract map; touched only this tracker. | `rg`, `sed` | No target-owned backend, transport, persistence, authorization, revision, or projection behavior found. | None. | Keep every future slice route-free and payload-agnostic. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | Legitimate route-free test facade with mixed setup/action/observer/diagnostic internals; no vendor import. | Inspected all target sources, grid-adapter/UI-contract evidence, 28 importers, representative callers; touched only this tracker. | `find`, `rg`, `sed`, `wc`, `stat`, `git check-ignore` | Public subpaths and private seam risks mapped. | Group-row semantic mismatch; marker observer correction needs authorization. | Implement S-01 first, then S-02/S-03; do not start S-05/S-06 without their gates. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | Mixed private boundary confirmed; stale group-row interpretation corrected. | Inspected target, adapter, UI contracts, pinned vendor source, live filter controls, 28 application importers, and static policy consumer; touched only this tracker. | `rg`, `sed`, `jq`, `node`, `readlink` | Adapter emits the data marker; vendor supplies row ARIA; filter values are select-or-fill; seven marker calls are mapped. | S-05 owner adoption; S-06/S-08 later authorization. | S-01, then structural slices; execute corrections only behind their gates. |
| 2026-08-04T16:29:03-04:00 | Codex / S-07 guide responsibility correction | The development guide names the exact `grid`, `accessibility`, and `visual` subpaths and their live capabilities without implying a generic accessibility API. The frontend implementation/testing guide distinguishes product-owner assertions from admitted shared choreography and makes all eight admission conditions binary. | Changed `docs/guides/cartulary-dev-guide.md`, `docs/guides/cartulary_frontend_implementation_testing_guide.md`, and this tracker only. | Responsibility-language `rg` inspection; `make lint-markdown`; `git diff --check` | Markdown passes at `.cartulary/test-results/20260804T202850Z-p2945180`; no API, implementation, generated, route, payload, authentication, dependency, lockfile, or product-behavior change occurred. | None. Compatibility impact is documentation-only. Rollback is restoration of the two guide sections; it requires no code generation or product migration. | Validate the tracker-inclusive Markdown checkpoint, then begin S-08. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | No generated contract is directly imported or changed; selector contracts are authored in UI contracts and harness topology is downstream of authored manifests. | Inspected package manifest, UI grid selectors, generated-artifact policy guidance, verification registry/owner contract; touched only this tracker. | `rg`, `sed`, `make help-all`, `make explain-target` | Generator/drift commands discovered; no generator run. | S-05 needs an owner-approved shared group semantic contract. | Update authored owners first and use Make generation only in later authorized work. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | Proposed group semantics, selector mapping, and compatibility posture are explicit but non-authoritative. | Inspected Core 03/04, UI selector exports/tests, adapter and vendor group rows; touched only this tracker. | `rg`, `sed`, `node`, `readlink` | Existing plural selector is frozen; proposed singular semantic selector is mapped; generated files remain untouched. | `BLOCKED:` Core owner has not adopted the proposal. | Obtain owner adoption before S-05; change authored inputs before Make generation. |
| 2026-08-04T17:10:29-04:00 | Codex / S-09 cumulative contract reconciliation | Core 03/04, UI-contract selectors, authored owner catalogs, and generated topology agree with the implementation. | Inspected all changed specification, selector, harness, generated, dependency, lockfile, golden, route, payload, authentication, and domain scope. | `make generate`; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; owner explanations; `git diff --check`; scope scans | Generate root `.cartulary/test-results/20260804T204655Z-p3162783`, JSON root `.cartulary/test-results/20260804T204713Z-p3165359`, drift root `.cartulary/test-results/20260804T204713Z-p3165215`, and policy root `.cartulary/test-results/20260804T204713Z-p3165257` pass. Only `tools/execution_topology_render_index.json` changed under generated outputs; no lockfile, dependency, golden, route, payload, authentication, or domain change occurred. | None. The additive selector and strengthened semantic DOM contract are the only supported contract migration; unsupported legacy group markers remain intentionally removed. Rollback requires reverting owner, projection, adapter, helper, test, and generated-topology changes atomically. | Complete the final tracker checkpoint. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | One package owner row accounts for 26 of 31 live tests; separate harness.browser policy protects route/payload and facade boundaries. | Inspected `index.test.ts`, family manifests, catalog/verification registry, Vitest config/runner, architecture policy; touched only this tracker. | `rg`, `sed`, `wc`, `make task-guide ROLE=module-author OWNER=package.test_utils`, `make explain-test-owner OWNER=package.test_utils`, `make explain-target` | Focused command is `make test-slice OWNER=package.test_utils`; no validation executed. | RB-001 owner-row accounting gap. | Implement S-01 without changing test behavior, then establish a focused baseline. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | RB-001 has a closed decision and exact binary evidence gate; S-02 preserves one selector entry and row identity. | Recounted tests and manifest titles; inspected exact row metadata, Vitest selector rules, and row-authoring guidance; touched only this tracker. | `rg`, `wc`, `jq`, `sed`, `make explain-target` | Live count remains 31; active exact-title count remains 26. No test or manifest changed. | S-01 blocks test movement. | Add only the five titles in S-01, then extract non-discovered suite modules behind the same entry. |
| 2026-08-04T15:43:00-04:00 | Codex / S-01 owner-accounting repair | All 31 live package titles are assigned to the existing active owner row. | Changed the authored package test-family manifest and Make-generated execution topology index; updated this tracker. | `make explain-test-owner`, `make generate`, `make json-shape-check`, `make generate-drift`, `make generated-artifact-policy-check`, `make test-slice OWNER=package.test_utils` | All commands passed; focused run root `.cartulary/test-results/20260804T194247Z-p2467255`; generated drift root `.cartulary/test-results/20260804T194236Z-p2464236`. | None. | Validate tracker Markdown, then begin S-02. |
| 2026-08-04T15:49:43-04:00 | Codex / S-02 suite split | The sole discovered entry now registers selector/grouping, continuity, virtual-targeting, and marker suites with shared fixtures. | Changed target tests and the catalog title resolver; updated this tracker. | `make format`, `make test-slice OWNER=package.test_utils`, `make frontend-typecheck`, `make frontend-unit` | Focused root `.cartulary/test-results/20260804T194736Z-p2478155`, typecheck root `.cartulary/test-results/20260804T194739Z-p2478647`, and 359-unit root `.cartulary/test-results/20260804T194750Z-p2479114` passed. The first focused run failed on a missed continuity-suite selector import; the related defect was fixed before the passing rerun. | None. | Validate tracker Markdown, then begin S-03. |
| 2026-08-04T15:59:49-04:00 | Codex / S-03 private seam split | Setup, actions, observers, and diagnostics are separate acyclic private modules; the three supported facades retain exact runtime and type surfaces, signatures, defaults, errors, and behavior. | Replaced `scrolling.ts`, `grid-editing.ts`, and `grouping.ts` with `grid-setup.ts`, `grid-actions.ts`, `grid-observers.ts`, and `grid-diagnostics.ts`; updated facades, focus/marker imports, package tests, title accounting, catalog registration discovery, generate-drift input discovery, and the Make-generated topology index. | `make format`; `make generate`; `make test-slice OWNER=package.test_utils`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; exact `harness.browser` architecture-policy slice; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; zero-deep-import `rg` scan | Focused root `.cartulary/test-results/20260804T195522Z-p2526270`, typecheck root `.cartulary/test-results/20260804T195522Z-p2526397`, 359-unit root `.cartulary/test-results/20260804T195522Z-p2526435`, import root `.cartulary/test-results/20260804T195522Z-p2526478`, architecture root `.cartulary/test-results/20260804T195717Z-p2566017`, final drift root `.cartulary/test-results/20260804T195915Z-p2578174`, JSON root `.cartulary/test-results/20260804T195939Z-p2580941`, and policy root `.cartulary/test-results/20260804T195939Z-p2580937` passed. Two related drift attempts failed because scratch generation did not yet copy registration-suite inputs and then reflected the newly changed harness input; dependency discovery and regeneration fixed them. | None. Compatibility impact is limited to unsupported deep imports, for which no consumers were found; rollback is atomic restoration of the three mixed modules plus registration/dependency-discovery changes followed by Make generation. | Validate tracker Markdown, then begin S-04. |
| 2026-08-04T16:01:34-04:00 | Codex / S-04 unsupported-internal cleanup | The empty placeholder, non-resizing `resizeGridColumn`, and test-only `assertAnchorTestId` are removed after a fresh full-repository consumer scan; the existing shared-selector title now exercises the supported `scrollGridCellIntoView` facade. | Removed `packages/test-utils/.gitkeep`; changed `grid-actions.ts` and the selector/grouping registration suite; refreshed the Make-generated topology index. | Full-repository `rg` consumer scans before and after removal; `make format`; `make generate`; `make test-slice OWNER=package.test_utils`; `make frontend-typecheck`; `make frontend-import-boundary-check`; exact `harness.browser` architecture-policy slice; `make generate-drift`; `make generated-artifact-policy-check` | Zero consumers remained after cleanup. Focused root `.cartulary/test-results/20260804T200119Z-p2589240`, typecheck root `.cartulary/test-results/20260804T200119Z-p2589472`, import root `.cartulary/test-results/20260804T200119Z-p2589535`, architecture root `.cartulary/test-results/20260804T200119Z-p2589326`, drift root `.cartulary/test-results/20260804T200119Z-p2589234`, and policy root `.cartulary/test-results/20260804T200119Z-p2589283` passed with no failures. | None. There is no supported compatibility impact; unsupported deep consumers would need to migrate to the public facade. Rollback is limited to restoring the removed private items and prior characterization body, then regenerating through Make. | Validate tracker Markdown, then begin S-05. |
| 2026-08-04T16:13:30-04:00 | Codex / S-05 group-row semantic adoption | Grouped grids now have adopted owner semantics: a treegrid root, level-one parent rows with row-level expansion parity, accessible bucket labels, one keyboard-operable toggle, and no expanded state on ordinary rows. Test-utils consumes the shared semantic selector; the legacy group-only ref mutations are removed after zero-consumer proof while structural styling is retained. | Changed Core 03 REQ-03-231/232, Core 04 AC-364, UI-contract selector/export/tests and owner titles, grid-adapter implementation/styles/tests, test-utils observer/tests, the generated topology index, and this tracker. | `make format`; `make generate`; focused slices for `package.ui`, `package.grid_adapter`, and `package.test_utils`; `make service-backed-test-slice OWNER=package.grid_adapter`; exact grouping rows for `platform.viewquery`, `module.timeline`, and `module.savedviews`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make browser-e2e-a11y`; `make browser-e2e-visual`; zero-consumer `rg` scans; `git diff --check` | UI root `.cartulary/test-results/20260804T200519Z-p2602145`, adapter root `.cartulary/test-results/20260804T200519Z-p2602152`, test-utils root `.cartulary/test-results/20260804T200519Z-p2602144`, service-backed root `.cartulary/test-results/20260804T200631Z-p2630676`, view-query root `.cartulary/test-results/20260804T200708Z-p2653491`, Timeline root `.cartulary/test-results/20260804T200742Z-p2675303`, saved-view root `.cartulary/test-results/20260804T200827Z-p2699338`, 359-unit root `.cartulary/test-results/20260804T200906Z-p2722558`, typecheck root `.cartulary/test-results/20260804T200906Z-p2722526`, drift root `.cartulary/test-results/20260804T200906Z-p2722294`, a11y root `.cartulary/test-results/20260804T201054Z-p2760981`, and visual root `.cartulary/test-results/20260804T201202Z-p2784298` passed with no failures or golden changes. | None. Compatibility is additive for the shared selector and intentionally removes unsupported legacy group markers; there is no persistence, route, payload, dependency, lockfile, or golden migration. Rollback must restore the owner text, selector, adapter binding, helper, tests, and bridge atomically, then regenerate through Make. | Validate tracker Markdown, then begin S-06. |
| 2026-08-04T16:27:09-04:00 | Codex / S-06 pure marker observer | Marker anchoring no longer imports or invokes grid setup. It retries read-only visibility, identity, and geometry observations every 50 ms for at most 3000 ms, reacquires locators after ordinary remounts, preserves row/cell identity and two-pixel tolerance, and emits bounded diagnostics. Shared collaboration setup prepares the observing page before remote focus; two post-trigger visual scrolls were removed, and the conflict-suite marker assertion now immediately follows the awaited presence consequence. | Changed `marker.ts`, its registration suite and exact owner titles, three collaboration/visual call-site regions, the generated topology index, and this tracker. | `make format`; `make generate`; `make test-slice OWNER=package.test_utils`; `make frontend-typecheck`; four exact `module.collaboration` rows; `make browser-e2e`; `make browser-e2e-visual`; `make frontend-unit`; `make frontend-import-boundary-check`; exact `harness.browser` architecture-policy slice; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; purity/order `rg` scans; `git diff --check` | All 36 package cases pass at `.cartulary/test-results/20260804T201907Z-p2816567`; typecheck passes at `.cartulary/test-results/20260804T201907Z-p2816632`; collaboration rows pass at `.cartulary/test-results/20260804T201940Z-p2817587`; the 53-unit functional browser graph passes at `.cartulary/test-results/20260804T202035Z-p2842838`; the 14-unit visual graph passes at `.cartulary/test-results/20260804T202344Z-p2876898`; 359 frontend units pass at `.cartulary/test-results/20260804T202512Z-p2901054`; drift passes at `.cartulary/test-results/20260804T202512Z-p2900813`. No failures or golden changes occurred. | None. The public signature is unchanged; unsupported callers that depended on assertion-owned scrolling must prepare explicitly. All supported callers are migrated. Rollback must restore the assertion and all caller ordering atomically; visual goldens remain untouched. | Validate tracker Markdown, then begin S-07. |
| 2026-08-04T16:45:21-04:00 | Codex / S-08 explicit selection capabilities | `changeGrouping` now requires `selectOption` and propagates rejection. `applyFilterChip` resolves and capability-checks the field selector before the popover, stops after field-selection rejection, tries value selection when present, falls back to fill only when selection is absent or rejects, and clicks Apply once only after successful value entry. `BrowserLocator.selectOption` remains structurally optional. | Changed `grid-actions.ts`, selector/action tests and exact package title accounting, the generated topology index, and this tracker. | `make format`; `make generate`; `make test-slice OWNER=package.test_utils`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; exact `harness.browser` architecture-policy slice; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make browser-e2e`; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make browser-e2e-a11y`; `make browser-e2e-visual`; optional-chaining and scope scans; `git diff --check` | All 40 package cases pass at `.cartulary/test-results/20260804T203146Z-p2954401`; typecheck passes at `.cartulary/test-results/20260804T203146Z-p2954480`; 359 frontend units pass at `.cartulary/test-results/20260804T203215Z-p2955926`; drift passes at `.cartulary/test-results/20260804T203215Z-p2955672`; functional 53/53 passes at `.cartulary/test-results/20260804T203404Z-p2997957`; webserver-backed 62/62 passes at `.cartulary/test-results/20260804T203716Z-p3032836`; stateful 36/36 passes at `.cartulary/test-results/20260804T204050Z-p3085266`; a11y 14/14 passes at `.cartulary/test-results/20260804T204247Z-p3110862`; visual 14/14 passes at `.cartulary/test-results/20260804T204353Z-p3134130`. No failures or golden changes occurred. | None. Signatures and error types are unchanged. Required-selection fakes must now implement `selectOption` or assert the intentional failure; text-value fakes may omit or reject it and provide fill. Rollback must restore action logic and negative-ordering tests together; no product, persistence, or generated-contract migration exists. | Validate tracker Markdown, then begin S-09. |
| 2026-08-04T17:10:29-04:00 | Codex / S-09 cumulative validation | The complete nine-slice implementation is owner-accounted, generated-clean, type-safe, boundary-clean, and green across focused product evidence and every applicable browser stage. | Reconciled all changed authored/generated files, all 40 package titles, three package owners, five product-owner selections, and final worktree scope. | `make agent-finalize`; `make generate`; JSON/drift/generated-policy/typecheck/unit/import/architecture gates; focused package and product owner slices; functional, webserver-backed, stateful, accessibility, and visual browser gates; `make check`; `make lint-markdown`; scope scans | Finalizer `.cartulary/test-results/20260804T204635Z-p3160239`; test-utils `.cartulary/test-results/20260804T204713Z-p3165574`; UI `.cartulary/test-results/20260804T204713Z-p3165604`; adapter `.cartulary/test-results/20260804T204737Z-p3170696`; view-query `.cartulary/test-results/20260804T204819Z-p3193462`; Timeline `.cartulary/test-results/20260804T204852Z-p3215052`; saved views `.cartulary/test-results/20260804T204937Z-p3238962`; collaboration `.cartulary/test-results/20260804T205027Z-p3264076`; workbook `.cartulary/test-results/20260804T205119Z-p3288429`; functional `.cartulary/test-results/20260804T205158Z-p3311234`; webserver-backed `.cartulary/test-results/20260804T205503Z-p3345367`; stateful `.cartulary/test-results/20260804T205837Z-p3397533`; a11y `.cartulary/test-results/20260804T210034Z-p3423062`; visual `.cartulary/test-results/20260804T210141Z-p3446321`; 714/714 check `.cartulary/test-results/20260804T210306Z-p3470025`; and tracker checkpoint `.cartulary/test-results/20260804T211307Z-p3598640` pass. | None. `RESULTS_DIR` was unset, so `make agent-finalize` truthfully skipped retained-run selection/maintenance with `results-dir-not-provided`. No measurement gate ran because no measurement consumer or Core 05 claim changed. No supported signature, timeout, tolerance, persistence, route, payload, auth, dependency, lockfile, or golden migration exists. Roll back only by slice in reverse dependency order. | Complete; no remaining remediation workstream. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | Target contains no authentication, authorization decision, route token, privileged control client, or domain payload. | Inspected target source, architecture policy, and import-boundary manifest; touched only this tracker. | `rg`, `sed` | Route/payload exclusion is an active harness.browser policy contract. | None. | Preserve route-free status and rerun the exact policy row after source/import movement. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | Security boundary remains unchanged; new diagnostics are constrained against raw HTML and incident payloads. | Rechecked target imports, capability types, and policy consumer; touched only this tracker. | `rg`, `sed` | No authentication, authorization, route, storage, or sensitive payload capability was added or proposed. | None. | Run the exact architecture-policy row after any source/import movement. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T14:30:34-04:00 | Codex / planning-only tracker creation | Tracker is decision-complete for behavior-preserving slices; behavior corrections remain separated. | Inspected all evidence named in Sections 1-5; touched only this tracker. | `git`, `find`, `rg`, `sed`, `wc`, `stat`, `make task-guide`, `make explain-test-owner`, `make help-all`, `make explain-target` | No validation suite, code generation, formatter, or production refactor was run. | RB-001 through RB-005 as classified below. | Start a later authorized session at S-01; refresh repository state before relying on this inventory. |
| 2026-08-04T15:12:31-04:00 | Codex / NLSpec-style tracker revision | All five RB decisions are explicit; only implementation/adoption gates remain. | Inspected all sources named in Sections 1-5; touched only this tracker. | `git`, `stat`, `rg`, `sed`, `wc`, `jq`, `node`, `readlink`, `make lint-markdown` | No production refactor, test, codegen, owner-spec, harness, or generated-artifact change occurred; Markdown passed at `.cartulary/test-results/20260804T191843Z-p2445899`. | RB-002 owner adoption; later authorization for RB-003/RB-005; S-01 evidence before movement. | Refresh repository state, implement S-01, then follow the dependency graph. |
| 2026-08-04T15:40:13-04:00 | Codex / authorized remediation activation | Execution authority is active and the prior adoption/behavior gates are closed decisions. | Updated Sections 1, 7, 9, 10, 11, and 12 in the tracker. | `date`, `git`, `apply_patch` | The tracker now requires a recorded, Markdown-clean checkpoint after every slice. | S-01 evidence remains pending. | Run tracker Markdown validation, then implement S-01 only. |
| 2026-08-04T17:10:29-04:00 | Codex / S-09 final risk review | All planned fixes and applicable validation are complete; no unresolved implementation or specification gap remains in scope. | Inspected the final diff, status, owner accounting, generated policy, lockfile/dependency/golden exclusions, and retained result roots. | `git diff --check`; `git status --short`; `git diff --name-only`; owner explanations; cumulative Make gates; final `make lint-markdown` | Scope contains only intended specification, guide, UI-contract, adapter, test-utils, E2E-ordering, harness-input, harness-script, and Make-generated topology changes; the tracker checkpoint passes. | None. | No remaining remediation workstream. |

## 11. Resolved Questions and Remaining Implementation Gates

The design question in each row is resolved. A resolved decision is not evidence
that its future implementation or owner adoption occurred.

| ID | Closure decision | Required implementation/evidence | Status |
| --- | --- | --- | --- |
| RB-001 | The five existing virtual-targeting titles belong in the existing active `package.test_utils` selector row. Row metadata, test files, exact titles, and behavior MUST remain unchanged in this repair. | Update the authored title array in ASCII order, regenerate only through Make, and retain proof that all 31 titles resolve exactly once and pass. | IMPLEMENTED; S-01 DONE. |
| RB-002 | The owner-defined treegrid contract and shared semantic selector in Section 4 are adopted by this remediation. The current attribute remains only through initial semantic validation, then zero-consumer legacy group mutations are removed before S-05 completes. | Align Core 03/04, UI contracts, adapter production binding, helper, accessibility evidence, and exact owner rows; prove zero supported legacy consumers before cleanup. | IMPLEMENTED; S-05 DONE. |
| RB-003 | Marker anchoring becomes observer-only after invocation. Public signature, identity checks, retry behavior, diagnostics boundary, and two-pixel tolerance remain stable. | Instrument purity, update all seven live calls atomically, and pass affected collaboration/visual evidence without unauthorized golden updates. | IMPLEMENTED; S-06 DONE. |
| RB-004 | Guide-listed keyboard, accessible-name, ARIA-state, contrast, and focus checks describe responsibility placement, not a required present package inventory. The exact current API remains three subpaths. | Correct guide wording, add the binary admission rule, preserve direct owner assertions, and add no new package export. | IMPLEMENTED; S-07 DONE. |
| RB-005 | Required grouping and filter-field selection fail closed. Filter-value entry preserves the explicit select-or-fill control mapping. `BrowserLocator.selectOption` remains optional and signatures remain stable. | Add failure-order, branch, rejection, fake, and live-consumer evidence; implement S-08 independently. | IMPLEMENTED; S-08 DONE. |

## 12. Binary Completion Criteria

Tracker completeness and implementation completion are distinct. Every
`TC-*` and `IA-*` row now passes.

### Tracker acceptance

| Acceptance ID | Binary criterion | Evidence | Current result |
| --- | --- | --- | --- |
| TC-001 | Every tracked target file and ignored artifact family is inventoried. | Section 2 lists nineteen authored files and two exclusions. | PASS |
| TC-002 | Every public interface, default, omission rule, and indirect product boundary has an owner and test posture. | Sections 3 and 4. | PASS |
| TC-003 | Current repository facts distinguish removed legacy adapter mutations, vendor-provided ARIA, and adopted owner semantics. | Sections 2, 4, 5, and RB-002. | PASS |
| TC-004 | Every resolved RB decision has aligned implementation status and retained evidence. | Sections 10 and 11. | PASS |
| TC-005 | Every workflow and slice has dependencies, intended change, risks, validation, rollback, and binary completion. | Sections 6 and 7. | PASS |
| TC-006 | Every behavior-changing slice has explicit authority; generated files were changed only through Make. | S-05, S-06, S-08, validation plan, and Section 10. | PASS |
| TC-007 | Exact impacted owners and commands are recorded from live catalog evidence rather than invented. | Sections 8 through 10. | PASS |
| TC-008 | Owner contradictions would be marked `BLOCKED: owner contradiction`; none is currently found. | Sections 1 and 7. | PASS |
| TC-009 | Repository/framework, repository/guide, and repository/analysis mismatches are explicit. | Sections 1, 4, and 5. | PASS |
| TC-010 | Prior handoff history is preserved and one current row exists in every required handoff table. | Section 10. | PASS |
| TC-011 | Phase maps and evidence rows are used only for verification accounting, not runtime architecture. | Sections 3, 4, and 5. | PASS |
| TC-012 | Markdown validation is truthfully recorded. | Section 8 and current handoff rows. | PASS |

### Future implementation acceptance

| Acceptance ID | Binary criterion | Required evidence | Current state |
| --- | --- | --- | --- |
| IA-001 | All 31 package tests resolve exactly once and pass under the package owner. | S-01 retained owner result `.cartulary/test-results/20260804T194247Z-p2467255` plus JSON/drift/policy checks. | DONE |
| IA-002 | One cataloged entry registers four non-discovered behavior-family suites while preserving exact row identity and behavior. | S-02 focused root `.cartulary/test-results/20260804T194736Z-p2478155` and broad root `.cartulary/test-results/20260804T194750Z-p2479114`. | DONE |
| IA-003 | Private setup/action/observer/diagnostic seams preserve every facade and default. | S-03 API-shape, typecheck, unit, import-boundary, architecture-policy, and generated-drift evidence recorded in Section 10. | DONE |
| IA-004 | Dead internals and placeholder have refreshed zero-consumer proof. | S-04 search, facade-shape, focused, typecheck, boundary, architecture, and drift gates recorded in Section 10. | DONE |
| IA-005 | Group-row semantics have adopted owner authority and aligned live evidence. | S-05 Core decision, owner slices, production binding, product rows, broad a11y/visual evidence, and zero-consumer legacy cleanup recorded in Section 10. | DONE |
| IA-006 | Marker assertion invokes no mutating capability and all seven callers use pre-trigger preparation. | S-06 purity, remount, timeout, identity, tolerance, collaboration, functional, and visual evidence recorded in Section 10. | DONE |
| IA-007 | Guide describes exact package API and binary helper admission without inventing API. | S-07 guide diff and Markdown evidence recorded in Section 10. | DONE |
| IA-008 | Required selection fails closed and value entry follows exactly one successful select-or-fill path. | S-08 negative ordering, branch, rejection, package, and five-stage live evidence recorded in Section 10. | DONE |
| IA-009 | Final generated drift, scope, owner evidence, and broad verification are clean or truthfully blocked. | S-09 retained results and implementation handoff. | DONE |

## 13. Iteration 2: Semantic Surface Hardening

Sections 1 through 12 are the completed Iteration 1 record and MUST remain
historical evidence. This section controls the authorized Iteration 2 execution.

### 13.1 Source posture and decisions

| Field | Iteration 2 value |
| --- | --- |
| Source fingerprint | Clean `main` at `12d57ef50bf7205c2572566dcfad20529f235e25` on 2026-08-04 |
| Execution authority | The 2026-08-04 implementation request authorizes S-10 through S-15 in serial order |
| Target state | Private, source-exported, version `0.0.0`; production-ready means a minimal enforceable internal contract, not a publishable artifact |
| Package identity | Keep `@cartulary/test-utils`; keep no root export and add no rename or compatibility package |
| Public subpaths | Remove `./accessibility` and `./visual` without shims; retain only `./grid` |
| Public API posture | Expose semantic grid choreography only; browser, network, locator, page, diagnostic, and scenario types or primitives remain private |
| Test posture | Preserve the sole discovered `index.test.ts` entry and all 40 exact active titles |
| Exclusions | No product specification, domain, route, payload, authentication, persistence, dependency, lockfile, generated contract, or visual-golden change |

The refreshed consumer scan found no `./accessibility` importer, one
`./visual` importer, one external consumer of the low-level browser/network
exports, and one consumer of `gridAnchorCommandScenarios`. The low-level
consumer is saved-view-owned support, while the scenario consumer is a single
Collaboration row. Neither creates continuing package-level compatibility
value.

### 13.2 Final public contract

The only supported package specifier after S-13 is
`@cartulary/test-utils/grid`. It has no public type exports and exactly these 17
runtime exports:

| Public runtime export | Responsibility |
| --- | --- |
| `applyFilterChip` | Apply one owner-backed grid filter through explicit field and value controls. |
| `assertActiveFilterChipVisible` | Observe the active filter chip. |
| `assertGridFocusContinuity` | Observe focus and viewport continuity. |
| `assertGroupRowPresentationOnly` | Observe the owner-backed presentation-only group-row contract. |
| `assertMarkerAnchoredToGridTarget` | Observe marker identity and geometry without mutation. |
| `assertMountedGridRowCountAtMost` | Observe the mounted saved-row bound. |
| `changeGrouping` | Select an owner-backed grouping field. |
| `collapseGridGroup` | Collapse one semantic group row. |
| `expandGridGroup` | Expand one semantic group row. |
| `isTestIdVisibleWithinGridViewport` | Observe target visibility within the owned grid scrollport. |
| `pasteGridMatrix` | Prepare one cell and dispatch exact TSV clipboard input. |
| `removeFilterChip` | Remove one active filter chip. |
| `scrollGridCellIntoView` | Prepare a record/field cell target. |
| `scrollGridTargetIntoView` | Prepare a stable test-ID target. |
| `scrollGridToBottom` | Scroll the owned grid scrollport to its bottom. |
| `scrollGridToOffset` | Scroll the owned grid scrollport to a vertical offset. |
| `sortByHeader` | Activate one owner-backed sort header. |

Signatures, timeout defaults, geometry tolerances, selector ownership, error
ordering, and Playwright-independent implementation remain unchanged except
where the slices below explicitly remove unsupported API or redundant
mechanics.

### 13.3 Serial slice plan

| Slice | Depends on | Remediation and ownership | Compatibility and migration | Risk if unresolved | Validation and exit criteria | Rollback posture |
| --- | --- | --- | --- | --- | --- | --- |
| S-10 | Tracker activation | Remove the duplicate `./accessibility` and `./visual` exports and facade files; migrate the sole visual importer to `./grid`; update guide and import-policy language plus exact facade characterization. Package, application E2E, guides, tests, and harness policy are in scope. | Intentional private-workspace API contraction. No accessibility consumer exists; the sole visual consumer migrates atomically. No shim or deprecation window. | Parallel import styles remain permanent compatibility burden and obscure the one cohesive package responsibility. | Zero alias imports; exact package facade, typecheck, frontend unit, import-boundary, architecture-policy, and Markdown gates pass. | Restore both aliases, facade files, consumer import, guide text, and policy wording together. |
| S-11 | S-10 | Replace saved-view imports of test-utils browser/network primitives with private saved-view structural types and owner-local guards. Then remove page evaluation, request, response, and interception capabilities from test-utils structural types. | Saved-view helper behavior, fake-driven signatures, request/response matching, error ordering, and visibility behavior remain unchanged. | Route/payload-capable types keep leaking into a route-free grid package and invite further app-support coupling. | Exact saved-view support row, `module.savedviews`, typecheck, frontend unit, import-boundary, architecture-policy, and Markdown gates pass. | Restore the imported primitives and structural members as one change; no product or wire rollback. |
| S-12 | S-11 | Move the Enter, Tab, blur, and single-cell-paste scenario table into its sole Collaboration consumer, typed with Playwright `Page` and `Locator`; remove the package scenario type/factory, unused `surface` parameter, and generic press/blur/dispatch guards. | Preserve scenario names, order, commands, and cataloged live behavior. Private fakes or deep imports receive no compatibility support. | A one-owner scenario and unused argument remain misleading reusable API and force unrelated capability members into the package. | Exact Collaboration service-backed row plus package, typecheck, boundary, architecture, and Markdown gates pass. | Restore scenario factory and consumer import atomically. |
| S-13 | S-12 | Remove all low-level runtime and type exports from `./grid`; prune unused structural members; remove the redundant paste `scrollIntoViewIfNeeded`; inline exact TSV formatting into the paste action and delete `matrix.ts`; characterize clipboard bytes through the public action. | The 17 semantic helpers remain supported. Exact TSV, focus-with-`preventScroll`, paste event, scrolling, defaults, and errors remain stable. Unsupported primitive consumers have been migrated in S-11/S-12. | The public API remains broader than its semantic purpose and redundant mutation can hide setup defects. | Runtime shape is exactly 17 symbols with no public types; zero external primitive imports; package, affected owner, type, unit, boundary, architecture, and Markdown gates pass. | Restore facade exports, structural members, formatter module, and paste call together. |
| S-14 | S-13 | Add negative compile evidence for root, removed aliases, private paths, and removed named exports; enforce the single subpath, Playwright/Node independence, acyclic runtime imports, and observer direction. Split registration and fixture support by facade, action, grouping, continuity, targeting, and marker seams while keeping one discovered entry. | No runtime or title compatibility impact. The same 40 titles remain assigned to the same row. Unsupported paths become explicitly enforced failures. | Accidental re-export or dependency reversal can silently recreate the removed legacy surface. Oversized test support will remain costly to extend. | All 40 titles resolve exactly once; generic dependency discovery handles every registration/fixture input; generation, JSON/drift/policy, package, type, frontend unit, script lint, import-boundary, architecture-policy, and Markdown gates pass. | Restore test modules and policy inputs together, then regenerate through Make. |
| S-15 | S-14 | Reconcile titles, generated topology, scope, and retained results; run cumulative validation and append final handoff. | No new behavior. Any failure is assigned to its originating slice; no golden update may hide drift. | Locally green slices can leave stale catalogs, generated topology, or an incomplete handoff. | Every command in Section 13.6 passes or has a truthful related failure record; final scope and tracker are clean. | Roll back only the responsible slice in reverse dependency order. |

### 13.4 Slice execution status

| Slice | Status | Latest evidence | Next gate |
| --- | --- | --- | --- |
| S-10 | DONE | Only `./grid` remains; zero supported alias consumers and all package, frontend, boundary, architecture, script, and tracker gates pass. | S-11 |
| S-11 | DONE | Saved-view support owns its structural browser/network seam; test-utils has no route, payload, request, response, or interception capability and all focused evidence passes. | S-12 |
| S-12 | DONE | Collaboration owns its four exact command scenarios; the package scenario factory, unused argument, and generic command guards are removed and all focused/live evidence passes. | S-13 |
| S-13 | DONE | `./grid` exposes exactly 17 semantic runtime helpers and no public types; paste uses one preparation path with exact TSV and residual primitive/formatter mechanics are removed. | S-14 |
| S-14 | DONE | Unsupported entrypoints and low-level exports fail at compile time; production imports are acyclic and Playwright/Node-independent; observer direction is enforced; one entry still registers 40 unique titles through cohesive support modules. | S-15 |
| S-15 | DONE | All generated, focused-owner, browser, lint, build, broad-check, scope, and handoff gates pass; no forbidden migration or golden drift occurred. | Iteration complete |

After each slice, this tracker MUST record files changed, behavior,
compatibility, commands, result roots, failures, and rollback posture. The
slice and corresponding task/acceptance rows become `DONE` or truthfully
`BLOCKED`, and `make lint-markdown` MUST pass before the next slice begins.

### 13.5 Iteration 2 task and acceptance tracker

| ID | Work item | Status | Depends on | Exit condition |
| --- | --- | --- | --- | --- |
| T-016 | Remove duplicate public subpaths and migrate all consumers | DONE | S-09 | S-10 exit criteria pass. |
| T-017 | Return saved-view browser/network support to its owner | DONE | T-016 | S-11 exit criteria pass. |
| T-018 | Return Collaboration command scenarios to their owner | DONE | T-017 | S-12 exit criteria pass. |
| T-019 | Contract the semantic grid facade and remove residual mechanics | DONE | T-018 | S-13 exit criteria pass. |
| T-020 | Enforce architectural boundaries and split characterization support | DONE | T-019 | S-14 exit criteria pass. |
| T-021 | Run cumulative validation and complete the handoff | DONE | T-020 | S-15 exit criteria and final tracker validation pass. |

| Acceptance ID | Binary criterion | Current state |
| --- | --- | --- |
| IA-010 | Only `./grid` remains and every alias consumer, guide, test, and policy reference is migrated without a shim. | DONE |
| IA-011 | Saved-view behavior owns its browser/network structural support and test-utils is route/payload incapable. | DONE |
| IA-012 | Collaboration owns its command scenarios and test-utils has no unused command capability layer. | DONE |
| IA-013 | `./grid` exposes exactly 17 semantic runtime helpers, no public types, and paste uses one preparation path with exact TSV. | DONE |
| IA-014 | Negative entrypoint, acyclic graph, Playwright/Node independence, observer direction, one-entry discovery, and 40-title accounting all pass. | DONE |
| IA-015 | Generated, focused-owner, browser, build, broad check, scope, and final tracker evidence pass. | DONE |

### 13.6 Validation and handoff requirements

The narrow gates named by each slice precede broader validation. S-15 runs:

- `make agent-finalize`
- `make generate`
- `make json-shape-check`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make frontend-typecheck`
- `make frontend-unit`
- `make frontend-import-boundary-check`
- `make test-slice OWNER=package.test_utils`
- `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.savedviews_suite_79e8036a1a,harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131`
- `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.browser.live_updates_never_retarget_pending_local_edits_11699e9b2d`
- focused `module.savedviews` evidence selected from the active owner catalog
- `make browser-e2e`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make browser-e2e-a11y`
- `make browser-e2e-visual`
- `make lint-biome`
- `make lint-scripts`
- `make build-web`
- `make check`
- `make lint-markdown`

Measurement is excluded because no measurement consumer behavior or Core 05
claim changes. Generated files are changed only through Make. `make
agent-finalize` receives `RESULTS_DIR` only when a successful full warm-check
root exists; otherwise the retained-run maintenance skip is recorded.

### 13.7 Iteration 2 handoff log

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Compatibility and rollback | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-04T17:24:28-04:00 | Codex / Iteration 2 activation | S-10 through S-15 are authorized serial workstreams; private-package, immediate alias removal, and semantic-only facade decisions are closed. | Refreshed package, consumers, guides, policies, owner catalogs, completed tracker history, branch, commit, and worktree; touched only this tracker. | `git status`; `git branch --show-current`; `git rev-parse HEAD`; repository `rg`/source inspection; owner explanations; `make lint-markdown` | Source is clean `main` at `12d57ef50bf7205c2572566dcfad20529f235e25`; live consumer counts and exact owner rows are recorded above; tracker activation passes at `.cartulary/test-results/20260804T212611Z-p3608106`. | No compatibility action yet. Remove no alias or primitive until its named slice migrates supported consumers. Roll back this activation by removing Section 13 only. | Validate the evidence-bearing tracker update, then begin S-10. |
| 2026-08-04T17:29:46-04:00 | Codex / S-10 duplicate-alias removal | `@cartulary/test-utils` now exposes only `./grid`; the zero-consumer accessibility alias and one-consumer visual alias are deleted without shims. | Changed the package export map, deleted both facade files, migrated the visual consumer, narrowed both guides, updated facade characterization, import-policy wording/fixture, and architecture-policy alias rejection. | Zero-consumer `rg` scans; `make format`; `make test-slice OWNER=package.test_utils`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; exact architecture-policy slice; `make lint-scripts`; `git diff --check` | Format `.cartulary/test-results/20260804T212743Z-p3611754`, package `.cartulary/test-results/20260804T212754Z-p3615089`, typecheck `.cartulary/test-results/20260804T212754Z-p3615331`, 359-unit `.cartulary/test-results/20260804T212754Z-p3615379`, import `.cartulary/test-results/20260804T212754Z-p3615414`, architecture `.cartulary/test-results/20260804T212754Z-p3615208`, and script-lint `.cartulary/test-results/20260804T212754Z-p3615537` pass with no failures. | The sole supported visual importer now uses the identical `./grid` export; accessibility had no importer. Removed aliases intentionally fail and receive no compatibility shim. Rollback restores the two facade files, export-map entries, consumer import, guide text, facade test, and policy wording together. | Pass tracker Markdown validation, then begin S-11. |
| 2026-08-04T17:35:07-04:00 | Codex / S-11 saved-view ownership restoration | Saved-view support now defines its private locator, page, request, response, visibility, evaluation, selection, and retry capabilities; test-utils retains only grid-required page capabilities. | Changed `apps/web/e2e/support/workbook/savedViews.ts`, `packages/test-utils/src/browser.ts`, `packages/test-utils/src/grid.ts`, and this tracker. | Route/network `rg` scans; `make format`; package slice; exact saved-view support row; focused `module.savedviews`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; exact architecture-policy slice; `git diff --check` | Format `.cartulary/test-results/20260804T213242Z-p3657922`, package `.cartulary/test-results/20260804T213251Z-p3661240`, saved-view support `.cartulary/test-results/20260804T213251Z-p3661210`, 34-unit saved-view owner `.cartulary/test-results/20260804T213251Z-p3661262`, typecheck `.cartulary/test-results/20260804T213251Z-p3661562`, 359-unit `.cartulary/test-results/20260804T213251Z-p3661611`, import `.cartulary/test-results/20260804T213251Z-p3661661`, and architecture `.cartulary/test-results/20260804T213251Z-p3661433` pass with no failures. | Runtime behavior, fake-driven signatures, route matching, request/response envelopes, visibility behavior, and failure ordering are unchanged. Compatibility is internal ownership only. Rollback restores the test-utils types/imports and removes the owner-local seam atomically. | Pass tracker Markdown validation, then begin S-12. |
| 2026-08-04T17:39:05-04:00 | Codex / S-12 Collaboration scenario ownership | The sole Collaboration consumer now owns the ordered Enter, Tab, blur, and single-cell-paste scenario table through native Playwright types; test-utils no longer exposes the scenario factory or generic command guards. | Changed the Collaboration spec, package actions/browser/facade modules, facade-shape characterization, and this tracker. | Zero-consumer `rg` scans; `make format`; package slice; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; exact architecture-policy row; exact Collaboration service-backed row; `git diff --check` | Final format `.cartulary/test-results/20260804T213847Z-p3797701`, typecheck `.cartulary/test-results/20260804T213851Z-p3800721`, package `.cartulary/test-results/20260804T213634Z-p3734977`, 359-unit `.cartulary/test-results/20260804T213634Z-p3735253`, import `.cartulary/test-results/20260804T213634Z-p3735299`, architecture `.cartulary/test-results/20260804T213634Z-p3735104`, and 12-unit Collaboration `.cartulary/test-results/20260804T213634Z-p3735158` pass. The first typecheck at `.cartulary/test-results/20260804T213634Z-p3735217` failed on one related unused type import; removing it produced the passing rerun. | Scenario names, order, commands, and live catalog title are unchanged. The removed package scenario had one consumer and receives no compatibility shim. Rollback restores the factory, command guards, facade exports, and consumer call together. | Pass tracker Markdown validation, then begin S-13. |
| 2026-08-04T17:46:08-04:00 | Codex / S-13 semantic facade contraction | The sole facade now has exactly 17 semantic runtime exports and no public types. Grid paste performs one target-preparation path, focuses with `preventScroll`, and dispatches exact TSV; the redundant locator scroll and standalone formatter module are removed. | Changed the grid facade, browser structural type, paste action, package characterization/fixture, both guides, deleted `matrix.ts`, and updated this tracker. | Exact facade and zero-consumer scans; `make format`; package slice; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; exact architecture-policy row; `make browser-e2e`; `make browser-e2e-a11y`; `git diff --check` | Final format `.cartulary/test-results/20260804T214543Z-p3905559`, typecheck `.cartulary/test-results/20260804T214547Z-p3908581`, package `.cartulary/test-results/20260804T214557Z-p3908986`, 359-unit `.cartulary/test-results/20260804T214109Z-p3807450`, import `.cartulary/test-results/20260804T214109Z-p3807475`, architecture `.cartulary/test-results/20260804T214109Z-p3807271`, functional 53/53 `.cartulary/test-results/20260804T214109Z-p3807606`, and a11y 14/14 `.cartulary/test-results/20260804T214110Z-p3807937` pass. The first typecheck at `.cartulary/test-results/20260804T214109Z-p3807522` rejected an overly narrow test-event cast; using the DOM `ClipboardEvent` contract produced the passing rerun. A direct Node source import was inapplicable because repository TypeScript modules use bundler resolution; the authoritative facade-shape package test passed. | Semantic helper signatures, defaults, errors, selectors, TSV bytes, focus, paste event, and grid behavior remain stable. Removed primitives had zero consumers after S-11/S-12 and receive no shim. Rollback restores the facade exports, structural member, formatter module, redundant call, characterization, and guide wording together. | Pass tracker Markdown validation, then begin S-14. |
| 2026-08-04T17:56:42-04:00 | Codex / S-14 architectural hardening | Unsupported root, removed aliases, a private deep path, and removed `delay` export now have compile-only negative evidence. Import policy enforces the sole declared subpath, an acyclic production graph, Playwright/Node independence, and observer-to-action/setup prohibition. One discovered entry registers the same 40 unique titles through facade, action, grouping, continuity, targeting, and marker-oriented modules. | Added the unsupported-entrypoint fixture; split the selector/grouping registration and shared fixture into cohesive package test-support modules; changed the sole test entry, import-boundary registry, static-analysis self-test, and this tracker; deleted the two superseded test-support files. No generated output changed. | `make format`; `make test-slice OWNER=package.test_utils`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-scripts`; `make generate`; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make frontend-unit`; exact architecture-policy row; owner/title/entrypoint and forbidden-import scans; `git diff --check`; `make lint-markdown` | Format `.cartulary/test-results/20260804T215336Z-p3913413`, package `.cartulary/test-results/20260804T215343Z-p3916639`, typecheck `.cartulary/test-results/20260804T215343Z-p3916758`, import policy `.cartulary/test-results/20260804T215343Z-p3916789`, script lint `.cartulary/test-results/20260804T215343Z-p3916838`, generation `.cartulary/test-results/20260804T215358Z-p3918151`, JSON `.cartulary/test-results/20260804T215410Z-p3920557`, drift `.cartulary/test-results/20260804T215410Z-p3920522`, generated policy `.cartulary/test-results/20260804T215410Z-p3920561`, 359-unit `.cartulary/test-results/20260804T215410Z-p3920916`, architecture `.cartulary/test-results/20260804T215410Z-p3920717`, and tracker Markdown `.cartulary/test-results/20260804T215714Z-p3963338` pass. The active package row contains 40 titles and 40 unique values; exactly one discovered test entry remains. No failures occurred. | Runtime exports and behavior are unchanged. Unsupported surfaces intentionally remain compile errors. Rollback restores the two consolidated support files and removes the negative fixture and three policy rules/self-tests together; regenerate only through Make. | Begin cumulative S-15 validation. |
| 2026-08-04T18:09:26-04:00 | Codex / S-15 cumulative validation and final handoff | Iteration 2 is complete. The private package has one semantic `./grid` facade with 17 runtime helpers, no public low-level types, compile-enforced unsupported surfaces, owner-local saved-view and Collaboration support, a directional acyclic implementation graph, cohesive characterization support, one discovered entry, and 40 uniquely owned titles. | Revalidated every authored/generated change from S-10 through S-14, added one lint-only use of the negative `delay` import, updated this tracker, and audited the final status. Changed scope is limited to test-utils, its two owner consumers, browser architecture/visual imports, two implementation guides, the import-boundary registry/self-test, and this handoff. | `make agent-finalize`; generation/JSON/drift/generated policy; frontend typecheck/unit/import policy; package, saved-view, Collaboration, and architecture focused rows; `make lint-biome`; `make lint-scripts`; `make build-web`; functional, webserver-backed, stateful, accessibility, and visual browser gates; `make check`; title/import/status/scope scans; `git diff --check`; `make lint-markdown` | Finalizer `.cartulary/test-results/20260804T215807Z-p3966621`, generation `.cartulary/test-results/20260804T215820Z-p3969123`, JSON `.cartulary/test-results/20260804T215837Z-p3971532`, drift `.cartulary/test-results/20260804T215837Z-p3971492`, generated policy `.cartulary/test-results/20260804T215837Z-p3971791`, unit `.cartulary/test-results/20260804T215838Z-p3972437`, harness rows `.cartulary/test-results/20260804T215838Z-p3972218`, saved views `.cartulary/test-results/20260804T215838Z-p3972324`, Collaboration `.cartulary/test-results/20260804T215838Z-p3972409`, script lint `.cartulary/test-results/20260804T215838Z-p3972726`, final typecheck `.cartulary/test-results/20260804T215958Z-p4035613`, import policy `.cartulary/test-results/20260804T215958Z-p4035622`, package `.cartulary/test-results/20260804T215957Z-p4035476`, Biome `.cartulary/test-results/20260804T215958Z-p4035685`, serial build `.cartulary/test-results/20260804T220012Z-p4037229`, functional `.cartulary/test-results/20260804T220033Z-p4042102`, webserver-backed `.cartulary/test-results/20260804T220033Z-p4042123`, stateful `.cartulary/test-results/20260804T220033Z-p4042153`, accessibility `.cartulary/test-results/20260804T220033Z-p4042399`, visual `.cartulary/test-results/20260804T220033Z-p4042223`, 714-unit check `.cartulary/test-results/20260804T220426Z-p5199`, and final tracker Markdown `.cartulary/test-results/20260804T221017Z-p163552` pass. Initial Biome `.cartulary/test-results/20260804T215838Z-p3973121` found the intentionally unresolved named import unused; `void delay` preserves the negative compile check and cleared the warning. A concurrent build `.cartulary/test-results/20260804T215838Z-p3973627` reported pass but emitted transient missing checksum inputs, so it was discarded in favor of the clean serial build. | No helper signature, timeout, tolerance, selector, route, payload, authentication, persistence, dependency, lockfile, product specification, domain, generated contract, or golden changed. Alias and primitive removals intentionally have no shim. Roll back in reverse slice order; S-14 policy/tests and S-13 facade contraction must remain paired. | Iteration 2 is complete; no implementation work remains. |
