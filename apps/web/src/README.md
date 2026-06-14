# apps/web/src Layout

This tree is organized by implementation ownership only. Directory placement
does not define product behavior, workbook-surface identity, route ownership,
or domain vocabulary.

- `main.tsx` stays at the source root as the Vite entry shim.
- `app/` owns the application shell, landing/auth/account surfaces, debug
  harness entrypoints, and app-shell tests.
- `services/` owns browser transport helpers and workbook API/evidence client
  helpers.
- `shared/` owns cross-feature helpers that are not workbook-specific.
- `testing/` owns Vitest setup files and shared test fixtures.
- `workbook/` owns the workbook shell, shell-level tests, and workbook
  subfolders.
- `workbook/components/`, `workbook/hooks/`, `workbook/models/`, and
  `workbook/utils/` hold shell-level presentation, runtime hooks, pure models,
  and reusable workbook helpers.
- `workbook/timeline/` owns Timeline-specific components, hooks, models, and
  services.

Use direct relative imports. Do not add path aliases or barrel exports for this
app without updating the local convention and import-boundary checks.

Keep tests beside the module or surface they cover. Shared fixtures belong in
`testing/`; generated artifacts and generated harness outputs are still owned
by their source manifests and must not be hand-edited.
