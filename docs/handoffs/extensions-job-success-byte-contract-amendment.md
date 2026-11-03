# Common Job success and commit-proof byte contract amendment

Status: adopted on 2026-11-02 under the maintainer's standing amendment
authorization, after concrete source review, passing candidate Markdown lint and
preservation review (tracker E-392). This amendment belongs to F-066 / WS-67.
Adoption establishes the owner contract; implementation and readiness remain
separate.

## Problem and decision

Core 01 §17 requires canonical terminal-success bytes, and Extensions
EXT-REQ-191 hashes those bytes and bounds the complete commit proof. Neither
contract assigns the exact encoding. The current implementation uses Go
`json.Marshal`, including struct declaration order and HTML escaping, for the
terminal digest and proof byte accounting. A different representation can
therefore change identity or admission without a defined behavioral change.

Use the existing shared RFC-8785 implementation for these runtime objects. The
new current schemas are `cartulary.common_job_terminal_success.v2` and
`cartulary.extension_job_commit_proof.v2`. Their member sets retain the existing
logical shapes. The new identities assign exact canonical bytes and distinguish
the current contract from the undefined predecessor. Registry/job-kind artifacts
continue to use their separately owned Extensions encoding, including its LF.

There is no configurable encoder selector, new generic registry or duplicate
canonicalization implementation. The supported terminal schema resolves to the
actual typed Job validator and canonical encoder; Extensions owns its proof
wrapper and uses the same platform primitive for complete-proof accounting.

## Concrete Core 01 amendment

Replace the common-job terminal-success definition in §17 with:

> `cartulary.common_job_terminal_success.v2` is the closed successful
> `result_summary` object from §3.3.9.1: required `code` and `message`, with
> optional `resource_refs[]` containing only
> `cartulary.common_job_resource_ref.v1`. Its canonical representation MUST
> retain the admitted code and message without Unicode normalization; reject
> duplicate resource identities; sort references by kind and then canonical
> resource-ID UTF-8 bytes; and omit `resource_refs` when it is empty. All
> selected-owner code, route and cardinality rules remain mandatory.
>
> The resulting closed object MUST be encoded as RFC-8785 canonical JSON in
> UTF-8, without a BOM or trailing LF. `terminal_result_sha256` MUST be the
> lowercase hexadecimal SHA-256 digest of exactly those bytes, with no prefix,
> delimiter or additional framing. A separate committed resource-reference
> array uses the same encoding and ordering, and is `[]` when empty.
>
> The canonical terminal object recorded in an Extensions commit proof MUST
> equal the canonical object committed to the public Job row. Physical JSONB
> representation and HTTP-envelope formatting are not digest inputs. A retained
> reader or replay path MAY reproduce the defined canonical bytes from that
> retained closed object; it MUST NOT reconstruct success from current resource
> state, replace its digest, or reinterpret it under a historical encoder.
>
> Current Job contracts MUST select v2. The prior terminal-success schema is
> unsupported in current bindings. This cutover does not alter the public
> result-summary member set, current cancellation semantics, or the separate
> live v1/v2 route-scoped idempotency identity contracts.

Update the immediately following Core-owned profile Job facts to select the
current terminal-success v2 schema. Their existing success-code and resource
policies remain authoritative.

## Concrete Extensions amendment

In EXT-REQ-191, replace the proof schema identity with
`cartulary.extension_job_commit_proof.v2`, including the required `schema_id`
value. Retain its twelve members, identity/nullability rules, atomic finalization,
correlation identities, immutable retention and nesting-depth ceiling.

Replace its canonical terminal/digest and proof-size assignment with:

> `terminal_result` MUST satisfy the selected current Core terminal-success
> v2 contract and contain the exact original canonical success value committed
> to the public Job. Its digest is the SHA-256 of the bytes assigned by Core 01
> §17. `resource_refs[]` MUST be the same ordered references, including the
> explicit empty array when the canonical terminal object omits that member.
>
> After validating every required member and selected identity contract, the
> complete commit proof MUST serialize as RFC-8785 canonical JSON in UTF-8,
> without a BOM or trailing LF. `max_proof_bytes` counts exactly this complete
> canonical encoding. Equality with the limit is valid; a larger encoding or
> nesting depth greater than 32 is invalid before final mutation. Physical
> database text, ordinary HTTP JSON, and a language-specific struct encoder
> MUST NOT determine that byte count. Canonicalization does not repair invalid
> members, types, identities, duplicate keys or Unicode.
>
> Only commit-proof v2 is supported by current bindings. A retained proof MUST
> still match the current selected Job contract and its original terminal
> digest. No historical terminal/proof encoder, compatibility fallback, stored
> digest replacement, conversion or automatic data deletion is introduced.

Advance the Extensions document patch version to `0.14.5`, retaining its public
contract major. Core 00 §1.7 records the current terminal/proof v2 byte assignment
and keeps implementation/readiness/release dispositions separate. No workstream
identifier or one-time review procedure is added to normative source text.

## Compatibility and implementation boundaries

- Current terminal/proof schema and dependent Job, implementation-binding and
  registry identities change together through authored inputs and Make
  generation. Unsupported bindings reject before their owner is invoked.
- The physical Job/proof table layout and public Job member sets do not change.
  Compatible retained logical values need no rewrite. A proof that fails its
  selected current contract or canonical digest remains rejected and retained.
- Backup generation, structural restore and later-claim admission retain their
  existing separation. Recompute actual dependent identities; do not invent a
  database migration or claim every older logical row is incompatible merely
  because it was written earlier.
- Revert the coherent unpublished change if necessary. Do not deploy, reset an
  environment, rewrite proofs or use old code to reinterpret new proof bytes.

## Required review and validation

- Review the exact Core/Extensions text above against the current typed writer,
  proof admission, retained reads and physical JSONB behavior before adoption.
- Freeze independent canonical byte/digest vectors covering reference member
  order, resource ordering, omitted/empty references, non-ASCII text, HTML
  characters and control escapes; prove no trailing LF enters either encoding.
- Exercise exact proof-byte limits and one-over rejection through the actual
  finalizer/admission path, with selected-owner policy and rollback preserved.
- Verify real PostgreSQL round trips, terminal/proof correspondence, retained
  replay, inactive reconciliation, strict malformed input and retired-schema
  rejection. Do not resolve current resource existence during shared validation.
- Regenerate projections through Make; run focused Jobs/Extensions and relevant
  collaborator checks, current capture/restore witnesses, schema/catalog/drift
  checks and preservation review. Record source identities, failures and passing
  evidence in the tracker before closing the workstream.

Passing tests establish implementation evidence; they do not adopt this
candidate or establish full module readiness.

## Current cutover identities

The following current projections changed. These are handoff observations of
Make-generated artifacts, not executable requirements or independent inputs.
The logical registry remains
`158c5849d6fde7053619099e80c7f3c1843ccd9c03571d546258a98ca1663999`;
the Extensions backup registry remains
`ba841d0b663f9c8f0036898572752d8e5505d8c657b0567ce28b9e0ec52e307b`.
Existing logical storage codecs and Recovery generations are unchanged.

| Generated artifact (under `contracts/extensions/generated/`) | SHA-256 |
| --- | --- |
| `dependency-snapshot.json` | `a4595efb45a68cda4e443ec23e6f94d336b8c63d4513db12156ee806970d438e` |
| `implementation-bindings/import.json` | `dcce043d8fb3132d5caf7cb09dbb52b54a6d8ba3a0ab4204a4f01ae768d065a7` |
| `implementation-bindings/incident_portability.json` | `7b644075ce2627b72822fa056037e93b9dcf29b724e40e29e929b8031f810a19` |
| `implementation-bindings/network_flow_activity.json` | `f7ac7f6d4289ebd764043a7b48c4259739b5e638937abedc91d1d29c9e894a4e` |
| `implementation-bindings/reference_pack.json` | `0705ff04b1ddb159519ba270afa2915307943b7ac737cf182e73d6deb33aa894` |
| `implementation-bindings/snapshot_reporting.json` | `ed379f3647d0793fb4cf8cb697b8fcff2ed48cdffe300d40511436a7cb3a5152` |
| `job-contracts/import/import.apply_v1.json` | `4e9f2e8b2b8e4a1c90c5bef67d66fc7490a0a82a8d4c95000ec6f40d1d690ee3` |
| `job-contracts/import/import.discovery_v1.json` | `dc26251196acf5128baa91ea3518f3634dfccdb752afe6bf1112df36e17915dd` |
| `job-contracts/incident_portability/incident_portability.export_v2.json` | `3c7bf50b03cc58c25091d6111c0ddedb70fe5e0e788ec8ff48b5b855c7613382` |
| `job-contracts/incident_portability/incident_portability.import_v2.json` | `79be2febde85e495991153e69eedde57e8fcfd6b8555788c933638f975eaf35b` |
| `job-contracts/network_flow_activity/network_flow_activity.graph_view_materialize_v1.json` | `68eb6b5b92c320cde60542332cbf80ad3b79b7f4f88b598b52b9a62676892250` |
| `job-contracts/reference_pack/reference_pack.import_v2.json` | `60b069092de4b5c17ed97d07ced16b371670f3bc72c1bb0c071c6eb415a04834` |
| `job-contracts/reference_pack/reference_pack.refresh_v2.json` | `4598f65b6a993ad0b6e7a21a6966b1cc46ba06c30501fd5eb4f87eed7d75a16f` |
| `job-contracts/reference_pack/reference_pack.reverify_v2.json` | `a661f2cace79f500747e02ec2d75c349eca51bbc1d02e12e4d90daab0966dcee` |
| `job-contracts/snapshot_reporting/snapshot_reporting.composition_preview_v4.json` | `a3a6274fe7caeba184b45213c092f7fb68f26f5c915a3e0a1ee25fd9f9af38da` |
| `job-contracts/snapshot_reporting/snapshot_reporting.release_create_v4.json` | `01f935fd97d039dc4ff4e69a294e105674ebc3fdf4c9aeebbcba3c3fbdfd9db2` |
| `job-contracts/snapshot_reporting/snapshot_reporting.snapshot_create_v4.json` | `bcceb9b03f629d7a4916c2f8e8bb1fc7ec295e7a907a343a191b30196cd26911` |
| `owner-input-registry.json` | `97b997440eafcbcae4a7ab882d1cf0e6969337900dc9216099e95430ca955e27` |
| `registry-integrity.json` | `b6dc8bee3f51b38ab83233e6a6b23b024514d6dd8ae0faad8a7d27fd949cb4ea` |
| `schemas/cartulary.extension_job_commit_proof.v2.schema.json` | `fc58dacc9788847ca60c709b6f431273ba3db1ef279d6d3d869a4d5d5e917e6b` |
| `state-registry.json` | `f2c00368a0c09732b936e309b80e8695294f52f88792c4f017466b6ac85d0b9f` |

The generated Core base Job contract, `contracts/jobs/base.v2.json`, has
SHA-256 `6ce0f1236a7238231ba13843a9eda05ad2ac590007098f4d2b3b72b8821f4391`.
