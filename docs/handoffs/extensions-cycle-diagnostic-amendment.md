# Extensions cycle diagnostic capacity amendment

Status: explicitly adopted by the user on 2026-11-01 (tracker E-133). Implementation and module
readiness are not established by this document. The controlling execution
artifact is [the Extensions tracker](extensions-module-refactor-tracker.md).

## Conflict and recommendation

The [Extensions owner](../extension-subsystem-nlspec.md) requires one finding
for each maximal strongly connected component, containing every internal edge
(EXT-REQ-062). Its registry permits 256 profiles, each with 16 dependencies
(EXT-REQ-041 and Table 7-F). A single component can therefore require 4096
distinct edge tokens. EXT-REQ-160/Table 26-B currently limits every
`conflicting_tokens[]` array to 256 items.

Raise the bound only for cycle findings. Preserve the complete component and
its edges. Restricting the supported graph, truncating diagnostics, or splitting
one component into several findings would contradict the existing graph rules
and make future profile dependencies harder to diagnose.

## Concrete owner amendment

In the paragraph following Table 26-B, replace this sentence:

> `conflicting_tokens[]` MUST contain `1..256` unique strings derived exactly by
> Table 13-A and sorted by ascending UTF-8 bytes.

with:

> `conflicting_tokens[]` MUST contain unique strings derived exactly by Table
> 13-A and sorted by ascending UTF-8 bytes. For
> `collision_class='dependency_cycle'`, it MUST contain `1..4096` items. For
> every other collision class, it MUST contain `1..256` items. A cycle finding
> MUST contain every distinct internal edge of its maximal strongly connected
> component; the implementation MUST NOT truncate, partition, or omit those
> edges to satisfy a smaller diagnostic bound.

The existing 256-profile, 16-dependency, finding-count, identifier, duplicate
count, ordering and secrecy rules remain authoritative. The 4096 bound is the
product of the admitted profile and per-profile dependency limits. It does not
increase the number of findings allowed in a result or phase.

No workflow identifiers or adoption history belong in the owner specification.
The adopted product contract above is applied to the owner; retain the
decision and execution evidence in the tracker and this handoff.

## Projection, compatibility and validation

Project the class-specific bounds into the authored diagnostic contract and
registered condition definitions used by the production classifier. Derive
repeated graph-limit facts from authored machine inputs and regenerate through
Make. No runtime, generator, test or evidence mechanism may read this document
or any other Markdown.

This is a coherent current-only, unpublished contract correction. Package the
updated definitions, generated identities and consumers together. Provide no
historical reader, alias, truncation fallback or automatic data change.

Validation must cover complete SCC findings with 256, 257 and 4096 internal
edges; separate components and outside dependents; exact membership and edge
ordering; duplicate exclusion; and maximum-length valid profile IDs. The
registered schema and formatter must accept complete valid cycle details and
reject excess or inconsistent tokens. Non-cycle classes must retain their
256-token bound. Finding-count overflow remains a separate classification.

Adoption closes the source ambiguity only. Production diagnostic integration,
packaged schema parity, secret-safe output, complete module validation and the
final handoff remain required.
