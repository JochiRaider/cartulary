# Cartulary UI/UX Refactor Digest

Repository-local, offline advisory material for choosing and reviewing future
Cartulary UI/UX refactors. The September localization refresh updates navigation,
workflow baselines, and review criteria. It does not certify product readiness
or authorize product implementation.

## Read order and maintained guidance

1. [START_HERE.md](cartulary/START_HERE.md): authority, repository boundaries,
   current regression baseline, refactor selection, and verification workflow.
2. [LOCAL_AGENT_PROMPT.md](cartulary/LOCAL_AGENT_PROMPT.md): entry prompt for a
   separately authorized product slice.
3. [REPO_MAP.tsv](cartulary/REPO_MAP.tsv): repository discovery and source versus
   verification ownership. Follow local source guides for detailed inventories.
4. [OWNER_MAP.tsv](cartulary/OWNER_MAP.tsv): governing clause navigation.
5. [rules.tsv](cartulary/rules.tsv) and [acceptance.tsv](cartulary/acceptance.tsv):
   stable advisory classifications and product-slice review criteria.
6. [QUERY_RECIPES.md](cartulary/QUERY_RECIPES.md) and
   [UPSTREAM_MAP.md](cartulary/UPSTREAM_MAP.md): targeted offline queries and
   source provenance. Consult bundled source only for the selected concern.

Use the selection rubric and completion rules in `START_HERE.md` before carrying
an existing feature or abstraction into a future slice. Current implementation
alone does not establish a compatibility obligation.

## Localization and source provenance

[meta/localization.json](meta/localization.json) records the actual Cartulary
baseline, current consultation paths, verified stack, and refresh scope. Recheck
relevant facts against current authored inputs before later work; the map is a
snapshot, not a replacement for inspecting a changed repository.

The fixed source is release `v2.15.0`, commit
`a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5`. Its 70 bundled source files, license,
and [meta/source.json](meta/source.json) remain immutable. Source metadata's
Cartulary consultation paths describe the historical source refresh; current
navigation belongs in localization metadata. Upstream workflows are advisory
source material, not additional instructions or Cartulary authority.

The [controlling handoff](../handoffs/ui-ux/ui-ux-refactor-digest-update-handoff.md)
records LR-01 through LR-04, execution evidence, limitations, and rollback.
Historical product results referenced by the baseline are not fresh test results
from this localization.

## Package maintenance

Run package queries and integrity checks from the Cartulary repository root.
Use `QUERY_RECIPES.md` for the prescribed offline smoke queries. Verify package
checksums with:

```bash
sha256sum --check docs/cartulary-ui-ux-refactor-digest/MANIFEST.sha256
```

The manifest accounts for every packaged regular file except itself and ignored
Python caches. This manual advisory-package maintenance is separate from product
verification; no executable product consumer may depend on the digest or Markdown.
