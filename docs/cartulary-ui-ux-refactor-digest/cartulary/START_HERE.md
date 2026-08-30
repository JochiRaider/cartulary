# Cartulary UI/UX Refactor Overlay

## Purpose

This repository-local overlay preserves the reviewed UI/UX remediation baseline
and prepares later, separately authorized slices. It does not authorize
implementation, behavior changes, a re-theme, or a product redesign. The
bundled upstream material is useful as an offline audit workflow, searchable
checklist, and source of review questions. It is not Cartulary authority.

All paths and commands in this overlay are relative to the Cartulary repository
root. Use
`docs/cartulary-ui-ux-refactor-digest/cartulary/REPO_MAP.tsv` as the compact
current-repository inventory rather than repeating discovery.

## Authority order

Apply this order before every later design or implementation decision:

1. Adopted subsystem NLSpecs for their named scope, using the exact current
   paths and statuses in `REPO_MAP.tsv`.
2. `docs/spec/00_document_set_status_and_precedence.md` through
   `docs/spec/04_security_deployment_and_conformance.md` for current
   implementation-conformance behavior.
3. `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` only for
   claim-bearing timed or fixture-sensitive publication.
4. `docs/design.md` for observable design behavior inside its declared scope.
5. `docs/domain.md` for repository vocabulary and owner navigation inside its
   declared scope.
6. `docs/handoffs/cartulary_modular_refactor_planning_framework.md` and
   implementation guides as subordinate planning or implementation support.
7. Current code and tests as evidence of current state.
8. This localized overlay.
9. Bundled upstream material as advisory evidence only.

If owner documents conflict, stop with `BLOCKED: owner contradiction`; do not
choose silently. Existing code is not automatically required behavior.
Separate an authorized normative correction from behavior-preserving structural
movement.

No executable test, generator, conformance check, or release artifact may read,
stat, hash, or otherwise depend on documentation text. Typed projections and
test routing remain owned outside `docs/` as described by `AGENTS.md`.

## Verified repository boundaries

The current frontend is a pnpm workspace with `apps/web` as the top-level Vite
and React application:

- `apps/web/src/app` plus verification owner `web.application` contains
  application-shell routing, authentication, incident-directory, account, and
  administration state.
- `apps/web/src/workbook` plus verification owner `web.workbook` contains the
  workbook shell, surface controllers, browser state, mutation recovery,
  collaboration projection, and continuity behavior.
- `packages/grid-adapter` is the sole direct `react-data-grid` adapter and maps
  vendor coordinates and lifecycle into Cartulary identities.
- `contracts/view-schemas` contains authored machine projections of adopted
  view owners. `packages/protocol-ts/src/generated` is generated from contract
  inputs. `packages/view-contracts` is an authored TypeScript adapter over the
  generated contract facade, and `packages/view-contracts/src/generated` is the
  managed generated projection of authored view-schema inputs.
- `packages/ui-contracts` contains authored stable selector/test-ID builders and
  consumes generated design tokens.
- `packages/test-utils` contains reusable semantic test helpers. Product-specific
  browser choreography and fixtures remain under `apps/web/e2e` and
  `apps/web/e2e/support`.
- `contracts/design/tokens.v1.json` is the authored token/theme/density
  projection. `packages/ui-contracts/src/generated/design-tokens.ts` is
  generated through `make generate`.

The owner ID `package.ui` does not imply that `packages/ui` is a functional
package; that directory is absent. The owner routes to `packages/ui-contracts`.
`docs/design.md` §3.11 contains the design-owned semantic icon registry, but no
standalone implementation icon registry is present. Current components import
`lucide-react` directly where needed.

`harness.visual` owns the visual-fixture registry contract but is not an active
owner accepted by `make task-guide`. Use `make browser-e2e-visual` for the
public visual entry point and select product/package owners through
`REPO_MAP.tsv`. Visual reconciliation and maintained goldens are active
implementation support; they do not become Core authority. The completed
remediation record is implementation evidence at
`docs/handoffs/cartulary-ui-ux-remediation-handoff.md`.

## Product constraints

Cartulary is a workbook-native incident workspace. Preserve spreadsheet speed
at the view layer while retaining disciplined relational, authorization,
history, and evidence behavior underneath.

- The grid is the protagonist.
- Capture remains grid-first, compact, direct, keyboard-complete, and tolerant
  of incomplete facts.
- Conflict, evidence, validation, save, and recovery feedback appears where it
  changes the user's action.
- The inspector augments the grid; it does not replace it or become a detached
  form workflow.
- The visual language is dense graphite, calm, precise, inspectable, and
  durable.
- Warm accent is scarce and semantic.
- State never relies on color alone.
- Required system views stay reachable inside the workbook shell.
- Vendor coordinates, SQL/projection names, component names, CSS classes, and
  package-specific icon names never become product concepts or test authority.

## Classification rule

Classify every material upstream recommendation before use:

- `ADOPT`: compatible principle that can be applied without changing a
  Cartulary-owned contract.
- `ADAPT`: useful concern whose prescription must be translated through
  Cartulary's desktop workbook, token, interaction, or owner contracts.
- `REJECT`: conflicts with current authority, scope, density, behavior, or
  visual direction.

Do not weaken or bypass this classification. Record it in later change notes
when it materially affects a decision. The baseline matrix is
`docs/cartulary-ui-ux-refactor-digest/cartulary/rules.tsv`.

## Regression baseline for later verification

The completed remediation established the following current baselines. Future
slices must preserve them unless separately authorized owner changes require a
different result. This list is neither new product authority nor authorization
to edit product behavior.

1. **Density propagation**
   - Compact/default/comfortable selection reaches row and header height,
     block/inline padding, typography, gutters, saved/draft/read-only content,
     and full-cell editor geometry through shared tokens.
   - Preserve this complete box and the owner-defined clear/null
     surface-default state.

2. **Row creation**
   - Every current create entry point derives from `create_capable`, interaction
     authority, `inline_create`, and total `fields[].create_writable` discovery.
   - Payloads contain only declared create-writable fields/inputs and projected
     ordinary minima. Evidence and Indicators retain owner-specific validation.

3. **Responsive shell**
   - Chrome selection uses validated CSS-length accessors and token-backed
     inline-size thresholds; block-size state remains independent.
   - `innerWidth` and `innerHeight` are the fallback when `visualViewport` is
     absent. Inspector clamp geometry and ARIA state agree.

4. **Inspector configuration**
   - Current sections/actions use stable semantic dispatch from the active
     `view_schema_id`, group, route kind/owner, and action.
   - Every current declared group resolves exactly once. Role/state restrictions
     render disabled, confirmation invalidates stably, and unknown additive
     groups follow `unsupported_feature_behavior=omit_feature`.

5. **Client transaction recovery**
   - Preserve Web Crypto `client_txn_id` generation, uncertain replay reuse,
     shell-lifetime pending recovery, and blocked-edit retry/discard behavior.

6. **Editing and conflict behavior**
   - Preserve single-click committed-cell editing, keyboard operation, paste,
     draft retention, Escape handling, and field-local validation/conflict
     feedback.
   - Keep saved values and retained local drafts distinct where the owner
     requires it.

## Refactor workflow for a future authorized slice

### 1. Bootstrap

- Read `AGENTS.md`; no nested `AGENTS.md` currently applies.
- Record branch, commit, dirty state, allowed edit scope, and target seam.
- Revalidate `REPO_MAP.tsv` if the repository commit differs from
  `meta/localization.json`.
- Identify the exact owner sections likely to govern the slice.

### 2. Inspect current state

- Detect the stack from `package.json`, `pnpm-workspace.yaml`, package
  manifests, and source; do not rely only on this localization snapshot.
- Inventory the relevant shell/work area, grid adapter, token use, density
  flow, view contracts, inspector configuration, focus/edit/paste behavior,
  transaction recovery, and current tests.
- Separate authored from generated files using
  `tools/generated_artifact_policy.json`.
- Keep direct grid-vendor imports inside `packages/grid-adapter`.

### 3. Freeze contracts

- Map each observable behavior that could drift to `OWNER_MAP.tsv` and the
  exact owner document.
- Characterize risky existing behavior before structural movement.
- Mark contradictions and missing authority explicitly.
- Do not promote labels, DOM structure, or current package layout into
  normative authority.

### 4. Choose one slice

Use one coherent seam per slice: one package boundary, state family, component
family, or verified interaction defect. Do not mix re-theming, schema changes,
route changes, harness redesign, and structural movement.

### 5. Apply advisory guidance

- Query only the relevant upstream domain or verified React stack using
  `docs/cartulary-ui-ux-refactor-digest/cartulary/QUERY_RECIPES.md`.
- Translate every material result through `rules.tsv`.
- Reuse Cartulary tokens, selectors, and fixture identities.
- Do not use upstream `--design-system`, `--persist`, or `--force`.

### 6. Validate

Select the owner-specific loop first:

```bash
make task-guide ROLE=module-author OWNER=web.application
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=package.grid_adapter
make task-guide ROLE=module-author OWNER=package.view_contracts
make task-guide ROLE=module-author OWNER=package.ui
make task-guide ROLE=module-author OWNER=package.test_utils
make task-guide ROLE=module-author OWNER=web.design
make task-guide ROLE=module-author OWNER=harness.browser
```

Run the recommended `make test-slice OWNER=<verified-owner-id>` or
`make service-backed-test-slice OWNER=<verified-owner-id>` only for the selected
owner. Use public readiness targets when the slice requires them:

```bash
make frontend-typecheck
make frontend-unit
make frontend-import-boundary-check
make browser-e2e-webserver-backed
make browser-e2e-a11y
make browser-e2e-visual
make generate-drift
make generated-artifact-policy-check
```

Do not run every listed target mechanically. `make help`, `make help-all`, and
the applicable `make task-guide` output determine the current runnable surface.

Visual and accessibility results remain quality/readiness evidence unless an
owner explicitly makes them implementation-conformance or Core 05 publication
evidence.

### 7. Handoff

For each changed behavior, record:

- governing owner clause or requirement;
- implementation location;
- verification command and artifact;
- upstream classification when material;
- known limitation or intentionally deferred issue;
- next smallest safe slice.

## Visual direction

Do:

- preserve `dark_graphite`, compact hairlines, neutral surfaces, stable row
  rhythm, semantic identity, and scarce warm accent;
- distinguish focus, selection, conflict, evidence, status, disabled, and
  read-only states with non-color cues;
- reserve space for async content and preserve authorized prior rows during
  background refresh when the owner requires it;
- distinguish loading, refreshing, successful empty, filtered empty,
  unauthorized, unavailable, stale, closed/read-only, conflict, and replay
  recovery states.

Do not:

- use cyberpunk, Matrix green, neon, glow, cinematic gradients, decorative
  heatmaps, threat maps, alert animation, or dashboard-card dominance;
- turn routine creation into a modal, wizard, approval, or full-page flow;
- generate a second design-system master or raw component-local token registry;
- add skeleton records that could be mistaken for authoritative workbook rows;
- treat horizontal workbook-grid scrolling as shell overflow;
- apply mobile touch-size, mobile-first, or 16px-body guidance unconditionally
  to the dense desktop profile;
- animate merely to satisfy generic timing advice.

## Stable verification identifiers

Prefer owner-defined fixture IDs, `view_schema_id`, `record_id`, `field_key`,
semantic icon IDs, and owner-defined capability/state enums. Do not assert
against literal specification text, line numbers, hashes, formatting, document
layout, visible row numbers, incidental DOM hierarchy, CSS classes, component
names, SQL/projection names, vendor coordinates, or package-specific icon
names.

Machine-testable facts derived from specifications belong in versioned,
owner-governed machine-readable artifacts outside documentation directories.
