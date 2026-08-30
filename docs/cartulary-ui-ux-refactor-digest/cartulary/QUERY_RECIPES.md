# Offline Query Recipes

Run every command in this file from the Cartulary repository root. The search
engine uses Python's standard library and bundled CSV files; it performs no
network access. Use `PYTHONDONTWRITEBYTECODE=1` with `-B` so Python and any
spawned processes cannot create bytecode caches.

## Rules

1. Use the verified stack in
   `docs/cartulary-ui-ux-refactor-digest/cartulary/REPO_MAP.tsv`.
2. Pass a domain or the verified `react` stack explicitly.
3. Start with a narrow concern, not a product or style query.
4. If a query returns zero results, broaden once and disclose the fallback.
5. Classify material results through
   `docs/cartulary-ui-ux-refactor-digest/cartulary/rules.tsv`.
6. Never generate or persist an upstream design system into Cartulary.

## High-value review queries

```bash
# Accessibility and interaction
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error messages" --domain ux -n 8 --json

# Async, empty, and recovery states
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "loading empty error recovery async feedback" --domain ux -n 8 --json

# Grid/list performance
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "virtualize list input latency reflow debounce" --domain ux -n 8 --json

# Responsive and layout stability
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "breakpoint layout shift fixed element overflow" --domain ux -n 8 --json

# Reduced motion
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "reduced motion interruptible layout shift" --domain ux -n 8 --json

# Candidate icon concepts; translate through docs/design.md §3.11
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "conflict evidence history warning retry discard" --domain icons -n 10 --json
```

## Verified React stack query

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "virtualized grid rerender focus async state" --stack react -n 8 --json

PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "rerender memo async waterfall bundle event handler" --domain react -n 10 --json
```

The stack query is advisory for React. Cartulary is not Next.js, Tailwind, or
shadcn, and examples using those technologies do not establish dependencies or
implementation patterns.

## Source-first inspection

Exact bundled paths:

- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/quick-reference.md`
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/ux-guidelines.csv`
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/stacks/react.csv`
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/react-performance.csv`

`docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/pro-rules.md`
is native/mobile-oriented. Use only concerns that transfer to desktop web, such
as semantic controls, icon consistency, contrast, and reduced motion.

## Bundled validation

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B \
  docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/validate_data.py
```

The complete 153-test release suite is a refresh-time provenance check. Run it
from a temporary full checkout of the exact upstream release, not from this
copied subtree: two test modules depend on upstream repository-root maintenance
scripts that are intentionally outside the bundle. The bundle remains offline
and independently usable for data validation and targeted queries.

## Forbidden upstream operations

Do not use `--design-system`, `--persist`, or `--force` against Cartulary.
Those options can create a parallel `MASTER.md` design authority and conflict
with `docs/design.md`.
