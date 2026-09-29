# Built-in workbook grid entry handoff

## Boundary and authority

Base: clean `main` at `02377560f082767cc33d68538ba41f081335685a`.
Work branch: `codex/built-in-grid-entry`. Core 03 §2 requires explicit base
surface selection to focus the first visible authorized creation control,
otherwise the first eligible committed cell in navigation mode, otherwise the
grid container. Design §8.4 repeats that order and reserves invoking-control
focus return for menu dismissal. `docs/domain.md` supplies vocabulary and owner
navigation. No specification or executable documentation dependency changed.

The defect was at explicit top-bar selection: desktop tabs and compact Surfaces
items omitted the existing `focusFirstGridTarget` request. One shared callback
now passes that request for both representations and System views. The startup
controller still owns generation, cancellation, and acknowledgement; the
semantic focus hook and Grid Adapter still choose and focus targets. Menu
dismissal still restores its trigger. Initial load, saved views, extension
workspaces, retained drafts, queries, mutation and recovery owners are untouched.

## Distinct changes

- Pre-existing catalog repair: ASCII-sort three titles in the authored
  `web.workbook` generic inspector reference-summary row. Both task guides were
  blocked by that ordering before product edits. `make generate` updated its
  generated digest.
- Product behavior: `WorkbookShellTopBar.tsx` routes explicit built-in choices
  through the existing focus option. No new focus store, timer, selector, API,
  schema, authorization rule, or write path was added.
- Regression evidence: `WorkbookShell.surfaces.test.tsx` checks all five
  destinations, compact dismissal/selection, and absence of row writes.
  `incident-administration.spec.ts` checks editor desktop and compact pointer
  and keyboard selection, viewer cell and empty-grid fallback, supersession,
  and the System views control. New rows live in authored test families;
  generated browser routing and topology were refreshed through Make.

## Verification and review

- Expected red before the product edit: shell row failed on desktop and compact
  focus at `.cartulary/test-results/20260929T213009Z-p22604`; browser editor,
  viewer, and supersession scenarios failed at
  `.cartulary/test-results/20260929T213319Z-p28352`.
- Focused green: `make test-slice OWNER=web.workbook` for the new row passed at
  `.cartulary/test-results/20260929T213617Z-p63612`. The surrounding shell,
  startup, semantic-focus, and ordinary-authoring rows passed at
  `.cartulary/test-results/20260929T213800Z-p2179`. The four browser rows,
  including System views, passed through `make service-backed-test-slice` at
  `.cartulary/test-results/20260929T213633Z-p64245`. The supersession row
  passed again after it was strengthened to wait for the old query to settle at
  `.cartulary/test-results/20260929T215142Z-p34839`.
- `make format` and `make agent-finalize` passed. `RESULTS_DIR` was unset, so
  retained-run maintenance was skipped. Import boundaries, generation drift,
  generated-artifact policy, JSON shape, and Markdown lint checks passed.
  Formatter changes to two unrelated baseline files were restored.
- `make frontend-typecheck` failed at
  `.cartulary/test-results/20260929T215300Z-p71512`: an unchanged
  `workbook-inspector-edit.spec.ts` passes a `string` where
  `WorkbookInspectorPanelId` is required. `make lint-biome` failed at
  `.cartulary/test-results/20260929T215300Z-p71538` on unchanged formatting
  in that file and `GenericInspectorReferenceSummary.test.tsx`. These are
  unrelated to this selection seam. The smallest remaining broad verification
  step is to repair those baseline files separately, then rerun both targets.

Seeded private UI review used synthetic editor and viewer actors. Desktop
Evidence focused its Title creation field; compact Escape returned to Surfaces
and keyboard selection opened Timeline; viewer Evidence focused a committed
Title cell in navigation mode. Captures were inspected before the exact session
closed successfully and task-owned requests were removed. Review observations
are separate from product tests and accessibility certification; advisory axe
findings were outside this bounded focus correction.

## Acceptance and rollback

The applicable digest criteria for authority, scope, repository state,
creation, continuity, editing, keyboard access, selectors, test authority,
generated artifacts, compatibility, and handoff have focused evidence above.
Visual styling, tokens, inspector, recovery, and extension-specific criteria
are outside this change. There is no persisted-data migration. Revert the
top-bar, focused tests, and authored test rows together, then regenerate
derivatives. The independent catalog-order correction remains valid.
