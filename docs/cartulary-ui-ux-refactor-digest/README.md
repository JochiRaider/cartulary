# Cartulary UI/UX Refactor Digest

Portable, offline advisory package for refactoring Cartulary's browser UI.

This package pins `nextlevelbuilder/ui-ux-pro-max-skill` to commit
`4857a2c5ef989794751a0f66b8545a4a49566286` (2026-07-28) and adds a
Cartulary-specific authority overlay. It is self-contained: the local agent does
not need GitHub or another network source.

## Read order

1. `cartulary/START_HERE.md` — governing overlay and implementation workflow.
2. `cartulary/LOCAL_AGENT_PROMPT.md` — ready-to-use execution prompt.
3. `cartulary/rules.tsv` — compact ADOPT / ADAPT / REJECT rule matrix.
4. `cartulary/acceptance.tsv` — binary completion checks.
5. `cartulary/OWNER_MAP.tsv` — compact contract routing.
6. `cartulary/QUERY_RECIPES.md` — offline, targeted searches.
7. `cartulary/UPSTREAM_MAP.md` — bundled-file scope and known conflicts.
8. `upstream/` — exact advisory source material; open only on demand.

Default context cost stays small: load `START_HERE.md`, the prompt, and only the
TSV rows relevant to the current slice. Search the upstream CSVs instead of
loading entire catalogs.

## Non-negotiable authority boundary

Cartulary's adopted owner documents govern behavior. `docs/design.md` governs
observable design behavior inside its stated scope. The bundled upstream skill
is advisory evidence only. It must not create routes, schemas, authorization
rules, product behavior, a second design system, or a parallel source of truth.

In particular, do not adopt the upstream "Cybersecurity Platform" visual
recommendation. Its Cyberpunk UI, Matrix green, black background, threat-display,
and alert-animation defaults conflict directly with Cartulary's dense graphite,
calm, workbook-first design contract.

## Offline query

From this directory:

```bash
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error feedback" --domain ux --json
```

The tool uses only Python's standard library and bundled CSV files. Use
`cartulary/QUERY_RECIPES.md` for scoped examples.

## Package contents

- `cartulary/`: synthesized Cartulary instructions and acceptance evidence.
- `upstream/ui-ux-pro-max/`: exact upstream skill, references, scripts, and data.
- `upstream/LICENSE.ui-ux-pro-max.txt`: upstream MIT license.
- `meta/source.json`: repository and review provenance.
- `MANIFEST.sha256`: integrity hashes for every packaged file except itself.

## Safe usage

- Inspect the real Cartulary frontend stack and repository state before editing.
- Read relevant owner sections before changing observable behavior.
- Use upstream searches to find review questions, not to make product decisions.
- Characterize risky behavior before structural movement.
- Test behavior through stable semantic identifiers and production-relevant
  outcomes; never bind tests to literal specification prose or incidental DOM.
