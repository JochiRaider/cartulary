# Offline Query Recipes

Run every command in this file from the Cartulary repository root. The search
engine uses Python's standard library and bundled CSV files; it performs no
network access. Use `PYTHONDONTWRITEBYTECODE=1` with `-B` so Python and any
spawned processes cannot create bytecode caches.

## Rules

1. Revalidate the selected stack against authored manifests; use
   [REPO_MAP.tsv](REPO_MAP.tsv) for navigation and
   [localization metadata](../meta/localization.json) for the dated snapshot.
2. Pass a domain or the verified `react` stack explicitly.
3. Start with a narrow concern, not a product or style query.
4. If a query returns zero results, broaden once and disclose the fallback.
5. Classify material results through
   `docs/cartulary-ui-ux-refactor-digest/cartulary/rules.tsv`.
6. Never generate or persist an upstream design system into Cartulary.

## UX localization smoke query

For the fixed unchanged source, this query returns three JSON results:

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "keyboard focus color only error feedback" --domain ux --json
```

The verified React smoke query below returns eight results. Inspect the JSON
result count and review material advice through existing rule IDs. These are
manual package-maintenance checks, not product acceptance tests.

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

## Verified React queries

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "virtualized grid rerender focus async state" --stack react -n 8 --json

PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "rerender memo async waterfall bundle event handler" --domain react -n 10 --json
```

The first command is the eight-result localization smoke query; the second is
an optional broader performance review query. Both are advisory. The inspected
Cartulary stack has no Next.js, Tailwind, or shadcn dependency; examples do not
establish implementation requirements.

### Material smoke-result dispositions

These mappings record review of the unchanged source for this localization.
Rules remain canonical in [rules.tsv](rules.tsv); the user's structural rubric
is separately attributed in [START_HERE.md](START_HERE.md).

| Bundled query advice | Existing rule disposition | Cartulary treatment |
| --- | --- | --- |
| UX Color Only | R003 `ADOPT` | Preserve owner-required non-color meaning. |
| UX Focus States and Focus Not Obscured (Enhanced); React Manage focus properly | R002 `ADOPT` | Apply declared visible focus and restoration behavior. Do not promote the source's enhanced AAA criterion into a new conformance requirement or infer a modal workflow. |
| React Handle async errors | R006/R008 `ADOPT` | Preserve owner-specific local failure and recovery; distinguish uncertain writes from failed reads after acknowledgement. |
| React Use Actions with async startTransition | R012 `ADAPT` | Review responsiveness within existing operation/admission ownership; a framework example cannot replace captured requests, authority fencing, or acknowledgement lifetime. |
| React Lift state up when needed; Use useState for local state; Avoid unnecessary state; Use useReducer for complex state | R012 `ADAPT` | Choose state placement from its semantic owner and lifetime. Derive redundant values where appropriate; do not move retained authoring or operations into short-lived component state because the source suggests a hook. |
| React Initialize state lazily | R012 `ADAPT` | Consider only for actual initialization cost within the chosen owner; no mandatory rewrite or speculative optimization follows. |

No fallback was needed for these two smoke queries. Future zero-result queries
may broaden once as described above and must disclose that fallback. The pinned
React catalog's verification date is upstream provenance, not a claim that this
localization independently validated every framework prescription.

## Source-first inspection

Exact bundled paths:

- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/quick-reference.md`
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/ux-guidelines.csv`
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/stacks/react.csv`
- `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/data/react-performance.csv`

`docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/references/pro-rules.md`
is native/mobile-oriented. Use only concerns that transfer to desktop web, such
as semantic controls, icon consistency, contrast, and reduced motion.

## Validation by changed boundary

The current localization leaves the upstream tree and source provenance unchanged.
Its gates are protected-source integrity, the two smoke queries, overlay/metadata
review, manifest reconciliation, and repository documentation finalization in the
[controlling handoff](../../handoffs/ui-ux/ui-ux-refactor-digest-update-handoff.md).
Do not run product suites or add a product dependency on these documents.

The complete 153-test full-checkout suite and bundled data validation were run
for the historical August source replacement. That suite is not a required rerun
for this localization. In a separately authorized future source replacement,
review validation requirements against that replacement: the full suite at this
pin needs repository-root maintenance scripts intentionally outside the bundle.
Never expand or patch the bundle to make that full-checkout suite run locally.

The bundled `validate_data.py` remains available for optional offline source/data
inspection with `PYTHONDONTWRITEBYTECODE=1 python3 -B`; the unchanged-source
refresh does not claim a new data-validation result.

## Forbidden upstream operations

Do not use `--design-system`, `--persist`, or `--force` against Cartulary.
Those options can create a parallel `MASTER.md` design authority and conflict
with `docs/design.md`.
