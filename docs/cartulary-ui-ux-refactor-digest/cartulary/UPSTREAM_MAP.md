# Bundled Upstream Material

All files under
`docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/` are exact copies
from release `v2.15.0` at pinned upstream commit
`a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5`. The copied Git subtree contains
70 tracked regular files and no symlinks. It is offline advisory material, not
Cartulary authority.

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
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/catalog-summary.json` and `data-provenance.json` | Upstream catalog counts and source provenance. | These describe upstream data maintenance, not Cartulary contracts. |

## On-demand files

| File or family | Use | Cartulary caveat |
| --- | --- | --- |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py` and `core.py` | Standard-library BM25 search. | Use explicit domains or `--stack react`; do not persist output. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/design_system.py` | Required import for the upstream CLI. | Do not invoke its generation or persistence paths. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/react-performance.csv` | React performance review questions. | Treat as advisory measurement prompts. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/app-interface.csv` | Native/mobile concerns. | Only generally transferable semantic and accessibility concerns apply. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/icons.csv` | Candidate icon concepts. | `docs/design.md` §3.11 owns semantic icon IDs; no standalone implementation registry exists. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/google-font-licenses.json` and `phosphor-icons-upstream.json` | Font and icon source provenance. | Provenance does not authorize new dependencies, fonts, or icon identities. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/reasoning_contract.py` | Upstream reasoning and relevance contract helpers. | Search ranking remains advisory and cannot resolve Cartulary owner conflicts. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/**` | Upstream data-quality, relevance, taxonomy, freshness, and layout tests plus fixtures. | The complete suite requires upstream repository-root maintenance scripts that are intentionally not bundled. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/pro-rules.md` | Native/mobile polish checklist. | Touch, safe-area, and mobile viewport assumptions do not apply directly. |
| Other bundled data | Source completeness and optional comparison. | Product, style, color, and landing recommendations remain advisory. |

## Validation topology

The bundled `validate_data.py` and search entry point operate offline from the
Cartulary repository root. Refresh-time validation runs the complete 153-test
suite from a temporary full checkout of the exact release because
`test_catalog_refresh.py` and `test_relevance_evaluator.py` depend on upstream
repository-root maintenance scripts outside the copied skill subtree. Those
scripts are not copied into Cartulary. Use `PYTHONDONTWRITEBYTECODE=1` together
with `python3 -B` so spawned Python processes cannot add caches to the snapshot.

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
