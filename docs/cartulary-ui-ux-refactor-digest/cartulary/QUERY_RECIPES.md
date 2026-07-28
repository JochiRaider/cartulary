# Offline Query Recipes

Run every command in this file from the Cartulary repository root. The search
engine uses Python's standard library and bundled CSV files; it performs no
network access. `-B` prevents Python bytecode caches.

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
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error messages" --domain ux -n 8 --json

# Async, empty, and recovery states
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "loading empty error recovery async feedback" --domain ux -n 8 --json

# Grid/list performance
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "virtualize list input latency reflow debounce" --domain ux -n 8 --json

# Responsive and layout stability
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "breakpoint layout shift fixed element overflow" --domain ux -n 8 --json

# Reduced motion
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "reduced motion interruptible layout shift" --domain ux -n 8 --json

# Candidate icon concepts; translate through docs/design.md §3.11
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "conflict evidence history warning retry discard" --domain icons -n 10 --json
```

## Verified React stack query

```bash
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "virtualized grid rerender focus async state" --stack react -n 8 --json

python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
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
python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/validate_data.py

python3 -B -m unittest discover \
  -s docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests \
  -p "test_*.py"
```

## Forbidden upstream operations

Do not use `--design-system`, `--persist`, or `--force` against Cartulary.
Those options can create a parallel `MASTER.md` design authority and conflict
with `docs/design.md`.
