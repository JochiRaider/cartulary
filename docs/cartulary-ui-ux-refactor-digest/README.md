# Cartulary UI/UX Refactor Digest

Repository-local, offline advisory package for a future Cartulary browser UI
refactor. This localization is preparation only: it maps the portable digest to
the current repository without authorizing product behavior or interface
changes.

The package pins `nextlevelbuilder/ui-ux-pro-max-skill` release `v2.15.0` to
upstream commit `a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5`
(2026-08-14). The 70 tracked files below the copied upstream skill path, the
upstream license, and `meta/source.json` are preserved source material and
provenance.

All operational commands documented by the localized overlay run from the
Cartulary repository root. No localized command requires changing into the
digest directory.

## Read order

1. `docs/cartulary-ui-ux-refactor-digest/cartulary/START_HERE.md`
2. `docs/cartulary-ui-ux-refactor-digest/cartulary/LOCAL_AGENT_PROMPT.md`
3. `docs/cartulary-ui-ux-refactor-digest/cartulary/REPO_MAP.tsv`
4. Relevant rows from
   `docs/cartulary-ui-ux-refactor-digest/cartulary/rules.tsv` and
   `docs/cartulary-ui-ux-refactor-digest/cartulary/acceptance.tsv`
5. `docs/cartulary-ui-ux-refactor-digest/cartulary/OWNER_MAP.tsv`
6. `docs/cartulary-ui-ux-refactor-digest/cartulary/QUERY_RECIPES.md`
7. `docs/cartulary-ui-ux-refactor-digest/cartulary/UPSTREAM_MAP.md`
8. Bundled upstream files only when a targeted advisory query requires them

`REPO_MAP.tsv` is the compact repository discovery record. Future agents should
use it instead of repeating package, owner, test, and generated-artifact
discovery.

## Authority boundary

Cartulary's adopted owner documents govern behavior. The exact authority paths
and bounded subsystem status are recorded in `REPO_MAP.tsv`.
`docs/design.md` governs observable design behavior only inside its declared
scope. `docs/domain.md` governs repository vocabulary and owner navigation
inside its scope. Current code and tests are implementation evidence, not
automatic product authority.

The bundled upstream skill is advisory evidence only. It must not create
routes, schemas, authorization rules, product behavior, a second design system,
or a parallel source of truth. In particular, its Cybersecurity Platform
recommendations for Cyberpunk UI, Matrix green, deep black, threat displays,
and alert animation conflict with Cartulary's design contract and remain
classified `REJECT`.

## Verified repository stack

The current frontend is a pnpm 10.33 workspace using TypeScript 6, React 19,
Vite 8, Vitest with Testing Library and jsdom, Playwright, and Biome. The grid
vendor is `react-data-grid`, contained behind
`packages/grid-adapter`. Icons currently use `lucide-react`. Tailwind and shadcn
are not repository dependencies.

## Offline query

Run from the repository root:

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error feedback" --domain ux --json
```

Use
`docs/cartulary-ui-ux-refactor-digest/cartulary/QUERY_RECIPES.md` for scoped
queries. Do not use the upstream design-system generation or persistence
options against Cartulary.

## Package integrity

Run from the repository root:

```bash
sha256sum --check docs/cartulary-ui-ux-refactor-digest/MANIFEST.sha256
```

The manifest contains deterministic repository-relative entries for every
packaged file except itself and ignored Python caches. `meta/localization.json`
records the repository state used for this localization.

## Safe usage

- Read `AGENTS.md` and the exact owner documents before planning a later slice.
- Treat completed create, inspector, density, responsive, and visual
  remediation as regression baselines, and verify any future concern against
  current code and owner clauses before proposing product work.
- Query upstream material for review questions, not product decisions.
- Use stable semantic identifiers and owner-governed machine projections.
- Never bind tests or generators to specification prose, Markdown paths, line
  numbers, formatting, or document hashes.
