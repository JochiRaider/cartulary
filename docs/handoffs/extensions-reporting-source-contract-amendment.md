# Reporting source contract amendment

Status: adopted under standing maintainer authorization; owner review and
pre-implementation adoption are recorded as E-374 in the controlling
[tracker](extensions-module-refactor-tracker.md).

The current Network Flow specialization describes an artifact emitter. Its real
owner boundary is a transactional reader of immutable Graph output and typed
pre-redaction label candidates. Generic emit contexts, output references and
post-redaction output semantics do not describe this boundary.

## Concrete owner changes

In Extensions EXT-REQ-232 and EXT-REQ-221, distinguish a
`snapshot_reporting_participant` contribution whose contract reference is
`cartulary.extension_reporting_source_contract.v1`. It is a typed source reader,
not the generic `emit` operation in Table 23-D/F. Its canonical closed object has
exactly `schema_id`, `profile_id`, `participant_id`, `participant_kind` (exactly
`snapshot_reporting`), `projection_schema_id`, `label_schema_id`,
`read_algorithm_id`, `ordering_algorithm_id`, `state_family_ids`,
`max_vertices`, `max_edges`, and `max_label_bytes`. The contribution digest uses
`extension_participant_contract_sha256_v1`; references must resolve to current
packaged definitions and the actual typed implementation before publication.
State families remain an exact sorted unique subset of the authoritative
presence manifest. No physical names, callbacks or environment values enter the
contract. Other participant kinds retain their existing operation contracts.

A source reader is registered only for an admitted claim. It runs only for a
selected exact current result while the source is ready. Missing, inactive,
stale, malformed, over-limit or unavailable selections fail closed; they do not
return an invented empty graph. The caller owns authorization, transaction and
its deadline. The reader never commits, mutates state, opens another transaction,
performs egress, or applies presentation redaction. It respects cancellation and
returns typed input to Reporting, which owns disclosure and redaction before
freezing its immutable input. Structural Recovery never invokes this reader.

Network Flow NF-REQ-170f fixes the participant ID and read algorithm to
`network_flow_activity.reporting_graph_source_v1`, the ordering algorithm to
`network_flow_activity.reporting_graph_order_v1`, the projection schema to
`graph_projection.v2`, the label schema to
`cartulary.network_flow.reporting_graph_source.v2`, and the only state family to
`network_flow_activity.graph_views`. Maximum vertices and edges are 100,000 and
250,000, exactly the Graph owner limits. Maximum canonical label bytes are
67,108,864. The fixed-size selection consists of the incident UUID and exact
validated Graph result binding; it is a native typed input, not an artifact
emission context. Validate the selection before reading, lock the immutable
result before its declaration, require the active exact selected tuple, then
read its exact complete result. Label candidates preserve the Graph owner’s
strict `(sort_key, object_id)` order with one candidate per vertex/edge.

Label-byte accounting uses the complete logical label object defined by its
schema, encoded with `extension_registry_canonical_json_v1`, including its final
LF. Absent/nonmatching scalar values encode as null. Candidate field paths use
one-based indices padded to at least four decimal digits; larger admitted
indices use their ordinary decimal width. Check counts and bounds before
publishing output, with bounded incremental construction. These bytes define
accounting only; existing frozen Reporting input representation is unchanged.

Reporting’s source section imports this boundary and retains sole ownership of
classification enforcement, redaction, selected-result freezing and later render
validation. Source errors retain the existing selected-projection failure
mapping; no fallback emitter, legacy source contract or inactive callback exists.

## Cutover and review

Replace only Network Flow’s misleading emitter declaration with the dedicated
source contract. Keep current typed ports and Graph/Reporting ownership. Update
authored schemas, reference closure and bindings, then regenerate through Make.
Registry and binding digests change together; retained Graph, Reporting and
Recovery bytes do not change. Reject old or mismatched source declarations
before publication without historical readers or data deletion. The separate
renderer absent-state conflict remains explicitly assigned to its own workstream.

Review checks: exact owner boundaries, no new generic framework, explicit bounded
native input/output, current-only admission, preserved immutable byte formats,
no workflow labels in normative specs, and no Markdown execution dependency.
Validation must mutate every contract field, prove claimed-only installation,
exercise real selected-result reads and stale/inactive rejection, validate the
logical label byte form including large index widths, and retain current frozen
input/restart behavior.
