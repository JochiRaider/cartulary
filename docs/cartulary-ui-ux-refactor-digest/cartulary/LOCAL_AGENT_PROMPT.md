# Local Agent Prompt

Refactor Cartulary's UI/UX as a contract-preserving remediation, not a re-theme
or product redesign.

Use the portable digest beside this prompt. Read `cartulary/START_HERE.md` first,
then load only relevant rows from `cartulary/rules.tsv` and
`cartulary/acceptance.tsv`. The bundled `upstream/` material is advisory and
offline-searchable; do not treat it as authority.

## Authority and scope

1. Treat adopted subsystem NLSpecs and Core 00 through Core 04 as the authorities
   for current product behavior.
2. Treat Core 05 as authority only for claim-bearing timed or fixture-sensitive
   publication.
3. Treat `docs/design.md` as the authority for observable design behavior inside
   its declared scope.
4. Treat the current repository as code truth, not automatic product authority.
5. If owner documents conflict, report `BLOCKED: owner contradiction`.
6. Do not create routes, schemas, authorization rules, lifecycle transitions,
   field membership, write behavior, or compatibility policy from generic UI
   advice.
7. Do not generate or persist a second design-system master. Preserve
   Cartulary's token registry, `dark_graphite` theme, density registry, semantic
   icon registry, surface contracts, and visual-fixture registry.
8. Do not infer behavior from labels, DOM hierarchy, CSS classes, component
   names, grid-vendor coordinates, SQL/projection names, or package icon names.

## First inspection

Read `AGENTS.md` if present. Record branch, commit, dirty state, scope, and
whether implementation changes are authorized. Detect the real frontend stack
from manifests and source. Inventory:

- shell and work-area ownership;
- grid adapter, direct vendor imports, and virtualization;
- token, semantic icon, and component usage;
- density selection and propagation;
- active `view_schema_id` and inspector configuration;
- row-creation capabilities and affordances;
- focus, keyboard, paste, commit, cancel, and conflict behavior;
- `client_txn_id` generation and pending-queue recovery;
- accessibility, component-state, browser, and visual-fixture coverage;
- authored versus generated artifacts.

Do not assume React, Tailwind, shadcn, or any grid implementation. Search by
symbols and contracts, then inspect every file proposed for editing.

## Work one coherent slice at a time

Map public behavior to owner clauses and characterize risky behavior before
movement. Choose the smallest coherent UI/package seam with a clear stop
condition. Separate normative corrections from behavior-preserving refactors.
Keep `/apps/web` responsible for workbook shell/application state, direct grid
vendor semantics inside `/packages/grid-adapter`, generated view adaptation in
`/packages/view-contracts`, stable selectors in `/packages/ui-contracts`, and
browser choreography in `/packages/test-utils`, unless current adopted owners
define a different boundary.

Prioritize verified instances of these defects:

1. Make density observably change row height, cell padding, typography, editor
   geometry, and gutter rhythm through shared compact/default/comfortable
   tokens. Preserve the owner-defined surface-default clear/null state.
2. Restore row creation for Hosts and Identities only when the active immutable
   view schema permits it. Keep creation in the grid or view bar.
3. Correct shell sizing so vertical-only resizing never hides or relocates top
   navigation, tabs, `System views`, account/application menu, or active query
   controls. Select chrome from inline size only; manage block size separately.
4. Derive inspector sections/actions from the active `view_schema_id` and
   `inspector_config_v1`; keep the inspector default-closed, grid-preserving, and
   continuity-safe.
5. Generate new `client_txn_id` values with Web Crypto randomness. Reuse an ID
   only for uncertain replay of the same logical mutation. Add non-modal queued
   edit recovery with `Retry with a new request ID` and
   `Discard blocked edit`.
6. Preserve single-click committed-cell editing, full keyboard operation, paste,
   local drafts, deterministic Escape behavior, and field-local validation and
   conflict feedback.

## Upstream advisory use

Classify every material upstream recommendation as `ADOPT`, `ADAPT`, or
`REJECT`. Use the bundled search tool only for a targeted concern:

```bash
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error feedback" --domain ux --json
```

If the actual stack is supported, query it explicitly. Do not use
`--design-system`, `--persist`, or `--force` for Cartulary. Reject the upstream
Cybersecurity Platform visual defaults: Cyberpunk UI, Matrix green, deep black,
threat-display, and alert animation.

## Visual direction

- Keep the grid as the protagonist.
- Preserve a dense graphite, calm, precise, inspectable workspace.
- Use warm accent sparingly for focus, selection, primary affirmative action,
  active grid handles, and limited brand emphasis.
- Use semantic state colors only with text, shape, marker, or accessible name.
- Use compact hairlines, neutral surfaces, stable rhythm, semantic icon IDs, and
  local feedback.
- Reject neon, glow, cinematic gradients, decorative threat maps, risk
  heatmaps, card-dashboard dominance, and decorative animation.
- Do not apply mobile-first layout, 44x44 touch targets, 16px body minimum, or
  no-horizontal-scroll guidance as unconditional desktop-grid rules.
- Do not create fake skeleton records inside the authoritative workbook grid.

## Quality gates

- Exercise every declared component state and compound-state precedence.
- Verify pointer and keyboard parity, visible focus, deterministic restoration,
  accessible names, live-region behavior, contrast, non-color cues, and reduced
  motion.
- Verify loading, refreshing, successful empty, filtered empty, unauthorized,
  unavailable, stale refresh, closed/read-only, conflict, replay recovery, and
  evidence lifecycle states.
- Preserve previous authorized rows during background refresh and show status
  separately.
- Run exact declared `D-VFIX-*` viewports and retain expected, actual, diff, and
  metadata when required.
- Use selectors based on fixture ID, `view_schema_id`, `record_id`, `field_key`,
  and semantic icon ID.
- Do not test literal specification prose, line positions, document hashes,
  formatting, or documentation structure. Use versioned owner-governed
  machine-readable artifacts for derived machine-testable facts.

## Required handoff

For every changed behavior, report:

- owner clause/requirement;
- before/after behavior;
- implementation paths;
- validation commands and results;
- evidence artifact;
- upstream classification if it affected the decision;
- intentionally deferred issue and next safe slice.

Do not claim completion until every applicable row in
`cartulary/acceptance.tsv` is `PASS`, `N/A` with rationale, or explicitly
`BLOCKED`.

