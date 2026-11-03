# Common Job resource identity clarification

Status: reviewed and adopted on 2026-11-02 under the user’s standing
authorization; recorded in controlling tracker E-329.

Core 01 §3.3.9.1 and REQ-634 already require the closed current Job resource
vocabulary, canonical public kind-specific IDs and canonical GET read routes.
Network Flow NF-REQ-029 already defines table IDs as `nft_` plus 32 lowercase
hexadecimal characters. Its saved-graph declaration currently calls
`graph_view_id` a stable generated identifier without stating its lexical form.

Add the missing `graph_view_id` row to Table 6-B and refer to that identifier
contract from the saved-graph declaration table:

> Stable server-generated Network Flow resource ID: `nfgv_` followed by exactly
> 32 lowercase hexadecimal characters from CSPRNG output. No normalization or
> alternate lexical representation is accepted.

This gives the shared structural Job validator a concrete imported scalar
contract. The prefixed form distinguishes saved graphs from their source tables
and preserves all current production IDs. It does not introduce a historical
reader or permit inactive profile execution. No data migration is required.

Implementation must check IDs and exact kind-specific read routes without
requiring retained resources to exist or checking present authorization. Keep
the abstract common resource schema's optional route; actual current Job
families require their adopted read routes, and Reference Pack version IDs are
the complete required route. Preserve both continuing Core identity schemas.

Review criteria: the amendment is limited to the missing lexical definition,
uses the existing owner and stable public identity concept, introduces no
workflow labels into the specification and changes no storage layout. Validate
current and malformed public IDs through real terminal/proof consumers.
