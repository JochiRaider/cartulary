# Offline Query Recipes

Run from the package root. Add `--json` for complete, machine-readable results.
The search engine is standard-library Python and performs no network access.

## Rules

1. Detect the actual Cartulary stack before a stack query.
2. Pass a domain explicitly; auto-detection can misroute overlapping terms.
3. Start with a narrow concern, not a product/style query.
4. If a query returns zero results, broaden once and disclose the fallback.
5. Classify results through `rules.tsv`.
6. Never persist generated design systems into Cartulary.

## High-value review queries

```bash
# Accessibility and interaction
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error messages" --domain ux -n 8 --json

# Async, empty, and recovery states
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "loading empty error recovery async feedback" --domain ux -n 8 --json

# Grid/list performance
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "virtualize list input latency reflow debounce" --domain ux -n 8 --json

# Responsive and layout stability
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "breakpoint layout shift fixed element overflow" --domain ux -n 8 --json

# Reduced motion
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "reduced motion interruptible layout shift" --domain ux -n 8 --json

# Semantic icon candidates; map results to Cartulary semantic icon IDs
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "conflict evidence history warning retry discard" --domain icons -n 10 --json
```

## Stack query

After detection, replace `<stack>` with one supported value shown by `--help`:

```bash
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "virtualized grid rerender focus async state" --stack <stack> -n 8 --json
```

React/Next-specific performance guidance is also searchable:

```bash
python3 upstream/ui-ux-pro-max/scripts/search.py \
  "rerender memo async waterfall bundle event handler" --domain react -n 10 --json
```

## Source-first inspection

For a full category scan without ranking:

- `upstream/ui-ux-pro-max/references/quick-reference.md`
- `upstream/ui-ux-pro-max/data/ux-guidelines.csv`
- `upstream/ui-ux-pro-max/data/stacks/<detected-stack>.csv`

`references/pro-rules.md` is native/mobile-oriented. Use it only for concerns
that genuinely transfer to desktop web, such as semantic controls, icon
consistency, contrast, and reduced motion. Do not import its touch/safe-area
assumptions wholesale.

## Forbidden Cartulary commands

Do not run any of the following against the project:

```text
--design-system
--persist
--force
```

They can create a parallel `MASTER.md` design authority and generate visual
recommendations that conflict with `docs/design.md`.

