# Entity candidate discovery

This package owns the non-mutating source read for `listEntityCandidates`.
Core 01 REQ-01-676 defines eligibility, matching, ordering and pagination. Routes
own authenticated incident admission, list-query parsing and opaque cursors;
the application composition root injects this narrow reader. No workbook
projection, mutation store or inspector presentation dependency is required.

Reads use the entity-view text ordering, bounded authoritative source batches and
batch-local active aliases and eligible preserved identifiers. Sparse searches
continue to exhaustion or cancellation, never a synthetic scan cutoff. Only one
page plus a lookahead match is returned. Search does not resolve, reuse, create,
merge or emit revisions. Existing source indexes are used; no migration or
persistent search index is required.

The typed discovery projection is `contracts/entities/candidate-discovery.v1.json`.
API and integration coverage routes through `module.entities`; use the current
repository task guide. Ship the additive endpoint before the frontend consumer;
roll back the frontend before removing the endpoint.
