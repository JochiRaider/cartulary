# Cartulary UI/UX Refactor Overlay

## Purpose

Use this package to improve Cartulary's UI/UX without accessing the upstream
repository and without weakening Cartulary's contracts. The external material is
most useful as an audit workflow, searchable checklist, and source of
stack-specific implementation questions. It is not a visual-design authority.

## Authority order

Apply this order before every decision:

1. Adopted subsystem NLSpecs for their named subsystem.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. `docs/design.md` for observable design behavior inside its declared scope.
5. `domain.md`, developer guidance, and owner-approved implementation support.
6. Current code and tests as evidence of current state.
7. This Cartulary overlay.
8. Bundled upstream material as advisory evidence only.

If owner documents conflict, stop with `BLOCKED: owner contradiction`; do not
choose silently. Existing code is not automatically required behavior. Separate
an authorized normative correction from behavior-preserving structural movement.

## Product constraints

Cartulary is a workbook-native incident workspace. Preserve spreadsheet speed at
the view layer while retaining disciplined relational, authorization, history,
and evidence behavior underneath.

- The grid is the protagonist.
- Capture remains grid-first, compact, direct, keyboard-complete, and tolerant of
  incomplete facts.
- Conflict, evidence, validation, save, and recovery feedback appears where it
  changes the user's action.
- The inspector augments the grid; it does not replace it or become a detached
  form workflow.
- The visual language is dense graphite, calm, precise, inspectable, and durable.
- Warm accent is scarce and semantic.
- State never relies on color alone.
- Required system views stay reachable inside the workbook shell.
- Vendor coordinates, SQL/projection names, component names, CSS classes, and
  package-specific icon names never become product concepts or test authority.

## Classification rule

Classify every upstream recommendation before use:

- `ADOPT`: compatible principle that can be applied without changing a
  Cartulary-owned contract.
- `ADAPT`: useful concern whose upstream prescription must be translated through
  Cartulary's desktop workbook, token, interaction, or owner contracts.
- `REJECT`: conflicts with current authority, scope, density, behavior, or visual
  direction.

Record the classification in the change notes when it materially affects a
decision. The compact baseline is in `rules.tsv`.

## Priority implementation targets

These are known UI concerns, not permission to invent behavior. Verify each
against the current repository and owner clauses before editing.

1. **Density propagation**
   - Make the selector observably alter row height, cell padding, typography,
     editor geometry, and gutter rhythm through shared compact/default/
     comfortable tokens.
   - Preserve `Use surface default` as the clear/null owner-defined state.
   - Do not create per-surface density systems, custom row heights, or a theme
     selector.

2. **Row creation**
   - Restore an adjacent grid/view-bar creation affordance for Hosts and
     Identities when the active immutable view-schema contract permits creation.
   - Keep create in the workbook surface; do not introduce modal, wizard, or
     form-first capture.

3. **Responsive shell**
   - Choose shell chrome from inline size only.
   - Treat block-size state independently.
   - Vertical-only resizing must not hide or relocate top navigation, tabs,
     `System views`, account/application menu, or active query controls.
   - Grid and inspector scroll within the shell-owned work area; keep the status
     strip anchored and safe navigation reachable.

4. **Inspector configuration**
   - Derive sections and actions from active immutable `view_schema_id` and
     `inspector_config_v1`.
   - Keep the inspector closed by default unless an owner says otherwise.
   - Preserve grid visibility at the base viewport and maintain row, scroll,
     selection, and focus continuity.
   - Invalidate stale row-bound state after view or row changes.

5. **Client transaction recovery**
   - Generate each new `client_txn_id` with Web Crypto randomness.
   - Reuse an ID only for uncertain replay of the same logical mutation.
   - Expose a non-modal queued-edits recovery surface with
     `Retry with a new request ID` and `Discard blocked edit`.
   - Do not silently drop, duplicate, or overwrite pending work.

6. **Editing and conflict behavior**
   - Preserve single-click committed-cell editing, complete keyboard operation,
     paste, local draft retention, deterministic Escape handling, and field-local
     validation/conflict feedback.
   - Show the saved value and retained local draft separately.
   - Do not use a toast as the only unresolved-state location.

## Refactor workflow

### 1. Bootstrap

- Read `AGENTS.md` when present.
- Record branch, commit, dirty-tree state, allowed edit scope, and target seam.
- Identify the owner sections likely to govern the slice.

### 2. Inspect current state

- Detect the actual stack from repository manifests and source; never assume
  React, Tailwind, shadcn, or a grid vendor.
- Inventory shell/work-area ownership, grid adapter, virtualization, token use,
  density flow, inspector configuration, creation capabilities, focus/edit/paste
  behavior, transaction IDs, pending recovery, and current tests.
- Separate authored from generated files.
- Keep direct grid-vendor imports inside the grid-adapter owner.

### 3. Freeze contracts

- Map each observable behavior that could drift to its owner.
- Characterize risky existing behavior before moving code.
- Mark owner contradictions and missing authority explicitly.
- Do not promote visible labels or implementation structure into contracts.

### 4. Choose one slice

Use one coherent seam per slice: one package boundary, state family, component
family, or interaction defect. Do not mix a re-theme, schema change, route
change, harness redesign, and structural refactor in one slice.

### 5. Apply advisory guidance

- Query only the relevant upstream domain or detected stack.
- Translate each result through `rules.tsv`.
- Reuse Cartulary tokens, semantic icons, selectors, and fixture identities.
- Do not run `--design-system`, `--persist`, or `--force` against Cartulary.

### 6. Validate

Run the cheapest sufficient owner-aligned checks after each risky checkpoint:

- unit/state-machine tests for deterministic state selection;
- component tests for declared variants and compound states;
- keyboard/focus/paste tests for workbook interaction;
- accessibility checks for names, focus, live regions, contrast, non-color cues,
  and reduced motion;
- exact `D-VFIX-*` viewports for visual evidence;
- broader repository gates only after the slice-local checks pass.

Visual/accessibility evidence remains evidence of those qualities unless an owner
explicitly makes it implementation-conformance or publication evidence.

### 7. Handoff

For each changed behavior, record:

- governing owner clause or requirement;
- implementation location;
- verification command and artifact;
- upstream rule classification when material;
- known limitation or intentionally deferred issue;
- next smallest safe slice.

## Visual direction

Do:

- preserve `dark_graphite`, compact hairlines, neutral surfaces, stable row
  rhythm, semantic icons, and scarce warm accent;
- make focus, selection, conflict, evidence, status, and disabled/read-only
  states distinct with non-color cues;
- reserve space for async content and preserve authorized prior rows during
  background refresh;
- distinguish loading, refreshing, successful empty, filtered empty,
  unauthorized, unavailable, stale, closed/read-only, conflict, and replay
  recovery states.

Do not:

- use cyberpunk, Matrix green, neon, glow, cinematic gradients, decorative
  heatmaps, theatrical threat maps, alert animation, or dashboard-card
  dominance;
- turn routine creation into a modal, wizard, approval, or full-page flow;
- generate a second design-system master or raw component-local tokens;
- add skeleton records that could be mistaken for authoritative workbook rows;
- treat horizontal workbook-grid scrolling as shell overflow;
- apply mobile touch-size, mobile-first, or 16px-body guidance unconditionally
  to the current dense desktop profile;
- animate merely to satisfy a generic timing recommendation.

## Stable verification identifiers

Prefer:

- design fixture ID;
- `view_schema_id`;
- `record_id`;
- `field_key`;
- semantic icon ID;
- owner-defined capability or state enum.

Do not assert against:

- literal specification text, line numbers, hashes, formatting, or file layout;
- visible row number;
- incidental DOM hierarchy;
- CSS class or component implementation name;
- SQL/projection table name;
- grid-vendor coordinate;
- package-specific icon name.

Machine-testable facts derived from specifications belong in versioned,
owner-governed machine-readable artifacts outside documentation directories.

