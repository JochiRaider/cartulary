# Bundled Upstream Material

All files under `upstream/ui-ux-pro-max/` are exact copies from the pinned
upstream commit. They are included for offline use and source traceability.

## Default-load files

| File | Use | Cartulary caveat |
|---|---|---|
| `SKILL.md` | Workflow, stack detection, domain routing, zero-result behavior. | Its design-system generation is not authoritative for Cartulary. |
| `references/quick-reference.md` | Scannable index of accessibility, interaction, performance, layout, forms, navigation, and data rules. | Translate mobile/general rules through Cartulary's desktop workbook profile. |
| `data/ux-guidelines.csv` | Searchable source rows for general UX rules. | Examples may name Tailwind classes; concepts, not classes, are transferable. |
| `data/stacks/<stack>.csv` | Stack-specific questions and examples. | Query only after detecting the real stack. |

## On-demand files

| File/family | Use | Cartulary caveat |
|---|---|---|
| `scripts/search.py`, `core.py` | Standard-library BM25 search over bundled data. | Use explicit domains; no persistence. |
| `scripts/design_system.py` | Required import for the upstream CLI. | Do not invoke its generation/persistence paths against Cartulary. |
| `data/react-performance.csv` | React/Next performance review if applicable. | Ignore if the repository is not React/Next. |
| `data/app-interface.csv` | Native/mobile app concerns. | Only semantic controls, feedback, a11y, contrast, and reduced motion transfer generally. |
| `data/icons.csv` | Candidate icon concepts. | Map behind Cartulary semantic icon IDs. |
| `references/pro-rules.md` | Native/mobile polish checklist. | Touch targets, safe areas, mobile viewport, and theme-pair assumptions do not apply 1:1. |
| Other domain data | Source completeness and optional comparative review. | Do not use product/style/color/landing recommendations as Cartulary authority. |

## Known contradictory upstream defaults

The exact upstream records are:

- `data/ui-reasoning.csv`, product row `80`, `Cybersecurity Platform`:
  recommends `Cyberpunk UI + Dark Mode (OLED)`, Matrix green/deep black,
  threat visualization, and alert animations.
- `data/colors.csv`, product row `80`, `Cybersecurity Platform`: defines
  Matrix green and alert red around black/dark surfaces.

These rows are preserved for provenance and must be classified `REJECT`.

## Why the full skill is included

The default Cartulary digest remains compact, while the complete search corpus
lets a local agent retrieve a few relevant rows without network access or
loading whole catalogs into context. Duplicate CLI distribution assets,
marketing pages, screenshots, examples, and unrelated premium design skills are
not included.

