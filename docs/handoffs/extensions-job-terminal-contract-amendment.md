# Selected Job terminal contract amendment

Status: reviewed and adopted under the user's standing authorization on
2026-11-02 (tracker E-348). Implementation and validation remain separate.

## Problem and authority

Core 01 §3.3.9.1 requires registry-backed initiating-owner success codes.
Core 01 §17, its Evidence verification clause, and NF-REQ-170c define current
outcomes. Their required resource references are part of successful completion.
Current shared admission checks shape, identity and maximum counts, but accepts
arbitrary nonempty success codes and omitted required references. The current
job-kind projection cannot express those obligations.

## Concrete owner changes

- Extensions EXT-REQ-218 adopts the current-only
  `cartulary.extension_job_kind_contract.v3`. Add required
  `terminal_success_codes`, a nonempty array of at most 64 distinct tokens,
  sorted by ascending UTF-8 bytes. Each token follows the existing Table 4-B
  local-key grammar. The set must exactly equal the adopted initiating owner's
  permitted successful outcomes. `job_canceled` is forbidden in this set.
- Each existing `resource_ref_contracts` row adds required integer `min_refs`
  with `0 <= min_refs <= max_refs`. Existing maximum and aggregate bounds stay
  in force. A successful result must meet every declared minimum and maximum;
  undeclared kinds remain forbidden. An empty declaration allows no references.
- Core 01 §3.3.9.1 requires one immutable selected terminal contract for every
  recognized Job kind, including base Jobs. It combines the exact terminal
  schema, declared success codes and resource cardinalities. Both live terminal
  publication and retained-success/proof admission enforce it. Validation is
  pure; it does not resolve resources, authorize access or execute profile code.
  Cancellation remains the separate exact `job_canceled` outcome.
- The base Jobs projection becomes `cartulary.base_job_catalog.v2`, carrying
  those same terminal fields for each base kind. No default policy is inferred
  from a handler name or an absent extension policy.
- Report Composition REQ-RC-035a declares the current preview Job's successful
  code `composition_preview_rendered` with no resource references. The preview
  route's immutable admitted response remains unchanged.

## Current owner matrix

| Job kind | Permitted success codes | Required and optional references |
| --- | --- | --- |
| `evidence.verify_blob_v1` | `evidence_blob_verified` | None |
| `import.discovery_v1` | `import_session_discovered` | Exactly one `import_session` |
| `import.apply_v1` | `import_session_applied`, `import_session_partially_applied` | Exactly one `import_session`; zero to 1,023 `network_flow_table` |
| `incident_portability.export_v2` | `incident_bundle_exported` | Exactly one `incident_bundle` |
| `incident_portability.import_v2` | `incident_bundle_imported` | Exactly one `incident` |
| `reference_pack.import_v2` | `reference_pack_imported` | Exactly one `reference_pack_version` |
| `reference_pack.reverify_v2` | `reference_pack_reverified` | Exactly one `reference_pack_version` |
| `reference_pack.refresh_v2` | `reference_packs_refreshed` | Zero to 1,024 `reference_pack_version` refs; canonical route order |
| `network_flow_activity.graph_view_materialize_v1` | `network_flow_graph_view_materialized` | Exactly one `network_flow_graph_view` |
| `snapshot_reporting.composition_preview_v4` | `composition_preview_rendered` | None |
| `snapshot_reporting.snapshot_create_v4` | `snapshot_created` | Exactly one `snapshot` |
| `snapshot_reporting.release_create_v4` | `release_created` | Exactly one `release` |

Core's inline Reference Pack activation/disable routes do not gain dormant Job
kinds. Their declared future long-running outcomes are not added to the current
catalog without a continuing production consumer.

## Implementation and compatibility

Keep the common typed terminal policy on the immutable Jobs definition, with
Extensions supplying its owner projection through the existing composition
boundary. Extension-specific proof, idempotency and cancellation policy remain
separate. Use one pure selected-policy validation path for live publication and
retained/proof admission. Preserve immutable copies and exact terminal bytes.

Update authored inputs and regenerate through Make. Current descriptor and
dependent binding identities change coherently. Retired job-kind/base catalog
shapes and incompatible bindings reject; no aliases, historical readers,
converters, automatic resets or data deletion are introduced. Valid current
owner outcomes keep their existing public representation. Release qualification
and deployment remain separate.

## Review and validation exits

Review this matrix against adopted owner clauses and actual continuing writers,
including conditional import results and preview completion, before adoption.
No executable check reads this Markdown.

Prove unknown, wrong-owner and cancellation codes cannot publish success.
Prove omitted required kinds, below-minimum and above-maximum counts reject
before live mutation and before inactive reconciliation. Preserve optional
outputs and all declared current outcomes. Cover coherent proof rehashing,
transaction rollback, catalog copy isolation, current-only schema rejection,
generation fidelity and actual owner completion integrations. Record the exact
commands, source baselines and retained results in the controlling tracker.
