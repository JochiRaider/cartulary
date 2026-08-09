# Projection Provider Manifest Maintenance

`index.json` is an authored, validation-only projection of the code-backed
descriptor registry assembled by `internal/app/projectionassembly/catalog.go`
from the eight required source-owner contributions. The assembled immutable Go
descriptor set is runtime authority; this manifest is not runtime
configuration and must not be loaded as one.

When provider descriptor content changes:

1. update the source-owner contribution and validate the assembled code-backed
   descriptor first;
2. update `index.json` in the same change, preserving deterministic provider
   and descriptor-member order;
3. change the manifest/schema version only when the manifest shape changes;
4. run `make test-slice OWNER=module.projections`,
   `make backend-module-boundary-check`, and `make json-shape-check`.

There is intentionally no manifest generator. Exact registry/manifest parity
is enforced by `internal/app/projectionassembly/catalog_manifest_test.go`,
while the JSON schema validates the authored contract shape.
