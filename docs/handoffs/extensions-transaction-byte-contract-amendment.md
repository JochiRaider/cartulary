# Extensions transaction byte-contract amendment

Status: explicitly adopted by the user and applied 2026-11-01 (tracker E-206).
Dependent implementation and module readiness remain incomplete.

The [Extensions tracker](extensions-module-refactor-tracker.md) controls this
required correction under F-045 / WS-43 and WS-44. This handoff is not behavioral
authority or an executable input. Earlier source approvals remain effective;
the user explicitly adopted this concrete amendment.

## Conflict and recommendation

[EXT-REQ-219](../extension-subsystem-nlspec.md#19-extension-data-ownership)
bounds canonical participant inputs and prepare, validation and write results.
EXT-REQ-042 assigns its integer-only, UTF-8 canonical encoding and final LF only
to explicitly designated objects. It assigns the participant contract digest,
but the invocation objects have no corresponding encoding assignment. Current
input producers use ordinary JSON without LF; result measurement uses the
Extensions encoder. Both byte identity and boundary accounting are ambiguous.

Assign the existing `extension_registry_canonical_json_v1` algorithm to these
internal objects. It provides exact integers, deterministic member order,
Unicode handling and a single newline rule without adding another encoder,
compatibility path or generic schema framework. Keep typed callbacks and
process-local capabilities. Public request parsing and owner request-digest
algorithms retain their own contracts.

## Adopted Extensions owner amendment

Under EXT-REQ-219, immediately after the paragraph ending “There is no unbounded
or empty legacy mode.”, insert:

> Each participant input and each prepare, validation and write result MUST
> use `extension_registry_canonical_json_v1` from EXT-REQ-042 as its canonical
> byte form. The assignment covers the complete closed object and all nested
> JSON values. Each independently encoded input or result MUST end with
> exactly one LF, and every individual or aggregate byte limit in this
> protocol MUST count that LF. A nested value is embedded as its JSON value;
> it is not an independently encoded byte string and adds no separate LF.
>
> At a byte input boundary, the coordinator MUST strictly decode the original
> bytes, require exact equality with this canonical byte form, and validate
> the selected participant input schema before any prepare invocation or
> transaction creation. Malformed or noncanonical original bytes MUST fail
> input admission; the coordinator MUST NOT normalize them into an accepted
> input. Input construction MAY encode an already validated owner value into
> this byte form. These requirements apply to the internal participant input,
> not to the public HTTP request or another owner-defined transport that
> supplies values for its construction. Owner request-digest algorithms
> remain governed by their own contracts.
>
> Typed result adapters MUST construct exactly the declared closed result
> object for schema validation and canonical byte measurement. A process-local
> transaction capability or owner receipt MUST NOT be added to that object
> or serialized into this protocol. An owner receipt used by the shared
> finalizer remains outside the protocol result and continues to follow its
> owner's value and publication rules. This assignment does not require a
> serialized callback transport or permit capabilities in a byte artifact.
> The existing participant order, result-validation precedence, first-invalid
> stopping, rollback and commit-outcome rules remain in force.

In the current protocol declarations and result examples, replace these exact
schema identities throughout their normative references:

| Retired identity | Current identity |
| --- | --- |
| `cartulary.extension_transaction_participant_prepare_result.v1` | `cartulary.extension_transaction_participant_prepare_result.v2` |
| `cartulary.extension_transaction_participant_validation_result.v1` | `cartulary.extension_transaction_participant_validation_result.v2` |
| `cartulary.extension_transaction_participant_write_result.v1` | `cartulary.extension_transaction_participant_write_result.v2` |

Following the inserted encoding paragraphs, add:

> The current canonical input and result byte contracts MUST use distinct
> schema identities from their predecessors without an explicit byte-form
> assignment. Each participant contract MUST select its owner's current input
> schema and current prepare, validation and write algorithm identities.
> Previous input schemas, result schemas and algorithm bindings MUST reject;
> no implicit normalization, historical reader or compatibility alias is
> permitted. The encoding algorithm remains the existing
> `extension_registry_canonical_json_v1`; its definition and already assigned
> byte forms do not change.

The input and result byte ceilings, closed member shapes, numeric and
Unicode rules are otherwise unchanged. This assignment imports EXT-REQ-042's
existing JSON-integer representation; it introduces no floating-point encoding.
Input construction does not authorize repairing untrusted bytes that already
purport to be a canonical participant input.

## Coordinated Core clarification and adoption metadata

In [Core 01 REQ-01-629](../spec/01_architecture_storage_and_view_contracts.md),
after the sentence beginning “Each participant input and their aggregate”,
insert:

> Canonical byte identity and limit measurement follow EXT-REQ-219's encoding
> assignment, including the required final LF for each input and result.

Core continues to delegate the protocol to Extensions. Network Flow
NF-REQ-185 already contributes typed cross-owner participants; no additional
owner algorithm or public payload restriction is proposed there.

As adopted, advance the Extensions document version to `0.14.2`
with unchanged public contract major `2`. Update its current-version declaration
and [Core 00 §1.7](../spec/00_document_set_status_and_precedence.md#17-extensions-coordinated-owner-contracts--2026-10-31)
to record this follow-up adoption on 2026-11-01 while preserving the earlier
coordinated adoption. Add this statement to that adoption record:

> Extensions document `0.14.2` assigns the existing canonical encoder and
> LF-inclusive byte accounting to internal shared transaction inputs and
> results. The follow-up amendment was explicitly adopted on 2026-11-01;
> implementation, module readiness and deployment remain separate.

No workstream identifier, tracker reference or one-time execution instruction
belongs in the owner specifications. Their text states durable behavior only.

## Compatibility, implementation and validation

This is a deliberate current-only internal byte-contract correction. Existing
noncanonical participant byte inputs will reject; current producers must change
together. It adds no reader, alias, converter, database migration, historical
backup support or automatic deletion. Public HTTP JSON, request idempotency
identities and stored owner payloads retain their own contracts. Public contract major remains `2`: public request, discovery and response
semantics are not changed. Table 14-B's identity rule is respected by replacing
the affected internal input/result schemas and owner algorithm bindings instead
of silently giving their predecessors a newly specified byte contract. The
existing encoding algorithm itself is unchanged and reused directly.

The coordinated authored/code projection cutover is concrete:

| Owner | New current input schema | Current owner algorithm identities |
| --- | --- | --- |
| Network Flow indicator link | `cartulary.network_flow_activity.indicator_link_transaction_input.v2` | `network_flow_activity.indicator_link_prepare_v2`, `network_flow_activity.indicator_link_validate_v2`, `network_flow_activity.indicator_link_write_v2` |
| Network Flow analytical import | `cartulary.network_flow_activity.import_apply_transaction_input.v2` | `network_flow_activity.import_apply_prepare_v2`, `network_flow_activity.import_apply_validate_v2`, `network_flow_activity.import_apply_write_v2` |
| Core Incident Bundles import | `cartulary.incident_portability_import_transaction_input.v2` | `incident_portability.import_prepare_v2`, `incident_portability.import_validate_v2`, `incident_portability.import_write_v2` |

Use `cartulary.extension_shared_protocol_contract_set.v5` for the changed closed
support contract. Logical participant IDs, the v3 participant-contract member
shape, the process-local context and the already specified finding/detail
schemas retain their meanings. Participant contract digests and all dependent
bindings, registries and packaged identities must be regenerated from authored
inputs through Make. Old identity/digest pairs reject; no aliases survive.
These owner projection identifiers belong in authored contracts and this
execution handoff, not a new cross-owner identity list inside the Extensions
behavioral specification.

Implementation must reuse the existing strict Extensions decoder and encoder,
update every current participant producer and fixture, and retain typed owner
schema validation. The later analytical import implementation must obey the
same byte contract. Update the authored shared-protocol projection to declare
the selected encoding and admission rule, with exact runtime/generator parity;
Markdown must never become that machine input. No encoding-selector framework
or arbitrary callback schema resolver is needed.

Validation requires fixed vectors for exact UTF-8, integer and nested-object
encoding; one required LF; malformed/duplicate members; invalid Unicode;
BOM, reordered keys, alternate escapes, whitespace, missing/extra LF and
trailing-data rejection. Exercise real coordinator admission at individual and
aggregate byte ceilings, preserving first-invalid order and zero mutation on
rejection. Test current owner producers and closed result projections,
including process-local receipt exclusion. Public owner parsing continues to
accept its previously specified JSON spellings. Existing real PostgreSQL
rollback, cancellation and indeterminate-commit evidence remains required.

The user explicitly adopted this amendment after source review and Markdown
lint. Tests did not confer adoption. Implementation and full readiness still
require their separate validation and handoff exits.
