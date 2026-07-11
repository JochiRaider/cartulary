# Projection Provider Manifest Maintenance

`index.json` is an authored, validation-only projection of the code-backed
registry rooted at `internal/modules/projections/provider_registry.go`. The Go
registry is runtime authority; this manifest is not runtime configuration and
must not be loaded as one.

When provider descriptor content changes:

1. update and validate the code-backed descriptor first;
2. update `index.json` in the same change, preserving deterministic provider
   and descriptor-member order;
3. change the manifest/schema version only when the manifest shape changes;
4. run `make backend-unit` and `make json-shape-check`.

There is intentionally no manifest generator. Exact registry/manifest parity
is enforced by `internal/modules/projections/provider_manifest_test.go`, while
the JSON schema validates the authored contract shape.
