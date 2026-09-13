# Web source layout

This guide describes implementation ownership under `apps/web/src`.
Start with the directory summaries below, then follow a child README for local
file responsibilities. Each guide inventories only its own files and immediate
subdirectories.

Directory placement does not define product behavior, workbook-surface identity,
route ownership, or domain vocabulary. Adopted specifications own behavior;
[domain.md](../../../docs/domain.md) owns vocabulary and owner navigation.

## Shared conventions

- Use direct relative imports. Adding path aliases or app-local barrel exports
  requires updating this convention and the import-boundary checks.
- Keep tests beside the module or surface they cover. Reusable fixtures belong
  in [testing](testing/README.md). Test names and selectors describe semantic
  owners and behavior; they are not runtime architecture boundaries.
- Generated artifacts and harness outputs remain owned by their source manifests
  and must not be hand-edited.

## Documentation maintenance

These READMEs are narrative implementation guides.
[The source ownership manifest](../../../tools/frontend_source_ownership.json)
accounts for exact `.ts`/`.tsx` ownership; the `web.architecture` policy tests
check that machine projection. Product tests, generators, runtime metadata,
conformance, and release evidence must not read, stat, hash, or parse READMEs
or other Markdown.

When adding, removing, or moving source files, update their local README and
the machine ownership manifest in the same change. Give every populated source
directory a README and link it from its parent. Keep file descriptions concise,
name the behavior tests cover, and place architectural detail at its narrowest
relevant directory. Documentation changes alone do not alter machine ownership.

## Subdirectories

| Directory | Responsibility |
| --- | --- |
| [app/](app/README.md) | Application shell, route entry surfaces, authentication, account settings, and incident/deployment administration. |
| [collaboration/](collaboration/README.md) | Incident-scoped browser WebSocket lifetime and typed collaboration event publication. |
| [extensions/](extensions/README.md) | Client extension discovery, availability coordination, and stable extension workspace identities. |
| [imports/](imports/README.md) | Workbook Import workflow ownership, captured requests, mapping, and job observation. |
| [measurement/](measurement/README.md) | Deterministic browser fixtures for Network Flow grid load measurement. |
| [networkFlow/](networkFlow/README.md) | Network Analysis extension presentation, analytical tables, imports, graph exploration, and retained operations. |
| [services/](services/README.md) | Browser transport, generated-contract adapters, bounded observation, and shared operation clients. |
| [shared/](shared/README.md) | Cross-feature validation, public errors, semantic contracts, and interaction helpers. |
| [testing/](testing/README.md) | Reusable frontend fixtures, Vitest setup, and source/import/selector policy tests. |
| [workbook/](workbook/README.md) | Workbook shell composition, cross-surface behavior tests, and navigation to workbook implementation owners. |

## Files

| File | Responsibility |
| --- | --- |
| [main.tsx](main.tsx) | Vite browser entry shim. Mounts the React application root and should stay thin. |
