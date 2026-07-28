# Bundled Upstream Material

All files under
`docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/` are exact copies
from pinned upstream commit
`4857a2c5ef989794751a0f66b8545a4a49566286`. They are offline advisory
material, not Cartulary authority.

The original upstream `SKILL.md` contains commands for its native plugin
location and includes design-system generation/persistence examples. Those
commands are preserved for provenance and are not Cartulary repository
commands. Use the repository-root commands in
`docs/cartulary-ui-ux-refactor-digest/cartulary/QUERY_RECIPES.md`.

## Default-load files

| File | Use | Cartulary caveat |
| --- | --- | --- |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/SKILL.md` | Upstream workflow, domain routing, and zero-result behavior. | Design-system generation is forbidden for Cartulary. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/quick-reference.md` | Accessibility, interaction, performance, layout, forms, and navigation index. | Translate mobile/general rules through Cartulary's desktop workbook owners. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/ux-guidelines.csv` | Searchable general UX rows. | Tailwind examples are illustrative only. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/stacks/react.csv` | React-specific advisory rows for the verified stack. | React advice does not override Cartulary package or state ownership. |

## On-demand files

| File or family | Use | Cartulary caveat |
| --- | --- | --- |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py` and `core.py` | Standard-library BM25 search. | Use explicit domains or `--stack react`; do not persist output. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/design_system.py` | Required import for the upstream CLI. | Do not invoke its generation or persistence paths. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/react-performance.csv` | React performance review questions. | Treat as advisory measurement prompts. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/app-interface.csv` | Native/mobile concerns. | Only generally transferable semantic and accessibility concerns apply. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/icons.csv` | Candidate icon concepts. | `docs/design.md` §3.11 owns semantic icon IDs; no standalone implementation registry exists. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/pro-rules.md` | Native/mobile polish checklist. | Touch, safe-area, and mobile viewport assumptions do not apply directly. |
| Other bundled data | Source completeness and optional comparison. | Product, style, color, and landing recommendations remain advisory. |

## Known contradictory defaults

The preserved contradictory records are:

- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/ui-reasoning.csv`,
  product row 80, `Cybersecurity Platform`;
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/colors.csv`,
  product row 80, `Cybersecurity Platform`.

They recommend Cyberpunk UI, Matrix green/deep black, threat visualization, or
alert animation and remain classified `REJECT`.

## Integrity boundary

`docs/cartulary-ui-ux-refactor-digest/upstream/`,
`docs/cartulary-ui-ux-refactor-digest/upstream/LICENSE.ui-ux-pro-max.txt`, and
`docs/cartulary-ui-ux-refactor-digest/meta/source.json` must remain
byte-for-byte unchanged. Localization belongs only in the Cartulary overlay,
`meta/localization.json`, the package manifest, and repository guidance.
