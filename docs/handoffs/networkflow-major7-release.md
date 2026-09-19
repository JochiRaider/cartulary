# Network Flow major 7 release handoff

Network Flow document 7.0.0 and public contract major 7 adopt consistent request
admission and define graph-query completion at audit commitment. The
[execution tracker](networkflow-module-refactor-tracker.md#17-f-24f-28--next-structural-refactoring-iteration)
contains the workstream and validation evidence; the
[Network Flow owner](../network-flow-activity-nlspec.md) defines behavior.

## User-visible changes

Claimed operations authenticate and check CSRF before current incident
visibility, role and lifecycle, resource-path syntax, request framing and
semantic validation. Malformed or hidden incidents use the same concealed
error. All operations reject undeclared URL parameters; bodyless reads reject
content. Invalid table and graph identifiers use resource error envelopes.
Core continues to reject unclaimed reserved families before module admission.

A graph execution completes when its required audit occurrence commits. A
later session-maintenance or response-delivery failure leaves that occurrence
intact. A newly executed retry creates another occurrence. Transaction commit
errors propagate without automatic replay; an indeterminate commit requires
operational investigation, not an assumption that the event was absent.

Saved reads propagate unexpected Jobs lookup failures safely. Missing job
history permits the documented retained-fact fallback. Read applications do
not infer a successful Jobs state from an existing selected result.

## Integration and compatibility

The standard browser supports major 7 only. Unsupported-major discovery hides
Network Analysis while preserving Base behavior. Ship server, browser assets
and generated registries together. Imports' active owner reference is
`network_flow_activity@7`; its facade and commit-protocol schemas are unchanged.
Go composition callers must handle the error returned by `Module.ImportOwner`.

Durable state 4, semantic query v2, materialization payload v1, source-profile
response v2, resource schemas, graph/digest identities, immutable mutation
receipts, import state, and exact Reporting leases remain unchanged. Existing
compatible retained data stays readable. There is no database migration,
receipt rewrite, historical deletion, or dual-major compatibility path.
The retained `saved_graph_cutover_v6` algorithm names an unchanged state
admission procedure, independently of the public major.

Three private applications own table queries, ephemeral graph queries and
saved-graph reads. Base construction retains one database/clock/dependency set,
Reporting provider and unobserved source verifier. Complete active composition
and single-assignment coordinator installation precede active publication;
inactive deployments retain their Reporting and Recovery contributions without
active key rings. Constructors perform no I/O or background work.

Private callers now use one typed `Project` operation returning
`graphprojection.ProjectionResultV2`, with an optional cancellation callback.
It replaces `ProjectEphemeral`/`ProjectSaved`; composer constructors no longer
accept an operational clock, and projection calls omit unused actor/time data.
Workers retain Jobs finalization and callers retain identity selection.
Removed surfaces include the unreachable non-v2 saved-result branch,
`graphEdgeAnnotations`, `graphComposition.SemanticSchemaID`, the route's unused
coordinator field, the production predicate wrapper, and
`semantic_http_test_bridge_test.go`. Their useful assertions now exercise live
semantic functions, current response metadata, the error mapper and actual HTTP.
Application guards inspect local declarations and references as well as imports.

## Rollout and rollback

1. Review the tracker's S-23 evidence and release candidate as one unit.
2. Quiesce the preceding release and its mutation/materialization writers.
3. Install the matching server, browser assets and generated registries.
4. Run the existing read-only retained-state admission before readiness. Resolve
   any incompatible-state finding through a separately adopted remediation;
   this release does not translate or erase it.
5. Verify claimed discovery advertises 7, Base remains available to unsupported
   clients, and current authorization, query audit and selected-result reads
   satisfy the adopted contracts. Record deployment-specific evidence.

Prefer a forward fix. Any rollback release must preserve the adopted admission,
audit, security and durable-state guarantees; the preceding binary is not
implicitly an approved rollback target. Repository verification does not prove
an operational deployment. Deployment and rollback execution remain external
and must have their own evidence.
