# Protocol-TS Public Type Compatibility Decision

> **Superseded by Protocol-TS Iteration 2.** This record is retained as
> historical evidence for the Iteration 1 compatibility posture. The private
> `0.0.0` package now exposes only its seven owner-declared family entrypoints;
> the root facade, `./core-http`, and the compatibility declarations described
> below were intentionally removed after repository consumer migration.

## Decision

The five supported package specifiers and every root export remain public. The
root facade aliases only `DensityMode` after S-00 and
`EvidenceHandleIssueRequest` after S-04 repairs its generated closed-empty
projection. The other fifteen handwritten declarations remain explicit root
compatibility types. Repository consumers migrate to generated,
operation-specific types without treating repository-local non-use as authority
to remove a public declaration.

This record describes the TypeScript compatibility decision. Core 01 and the
adopted subsystem owners continue to define wire behavior. Tests compile source
fixtures and runtime inputs directly; they do not read or otherwise depend on
this Markdown file.

## Effective compiler posture

The characterization runs inside `packages/protocol-ts/tsconfig.json` through
the workspace TypeScript build. It therefore retains `strict`,
`exactOptionalPropertyTypes`, `noUncheckedIndexedAccess`, readonly write
checking, and Bundler package-export resolution. The fixture imports all five
supported package specifiers and exercises assignment in both directions,
writes, omission, explicit `undefined`, excess members, arbitrary lookup,
header-value use, and generic parameter/return positions.

## Symbol matrix

`Legacy` means the current root declaration. `Generated` means the candidate
from `@cartulary/protocol-ts/core-http`. Assignment results include nested
member and index-signature behavior under the effective compiler posture.

| `public_symbol` | `generated_symbol` | `contract_owner` | `wire_position` | `property_variances` | `legacy_to_generated` | `generated_to_legacy` | `write_capability_equal` | `runtime_schema_equal` | `disposition` | `removal_gate` |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `ViewCell` | `ViewCell` | Core 01 §3.3.4, REQ-01-034 | `response` | Both are open; legacy is readonly and makes `value` optional, while generated is mutable and requires `value`. | No | Yes | No | No | `retain_compatibility` | A separately authorized versioned root-API migration plus required-value compatibility evidence. |
| `ViewRow` | `ViewRow` | Core 01 §3.3.4, REQ-01-034 | `response` | Both are open; legacy is recursively readonly, while generated is mutable, names optional `group_values`, and contains required-value generated cells. | No | Yes | No | No | `retain_compatibility` | A separately authorized versioned root-API migration after external-caller impact analysis. |
| `ViewMutationData` | `ViewMutationData` | Core 01 §§3.3.4 and 3.3.8 | `response` | Legacy is readonly, permits any string `view_schema_id`, and contains optional-value compatibility cells; generated is mutable, has a closed current-profile ID union, and contains required-value cells. | No | Yes | No | No | `retain_compatibility` | A separately authorized versioned root-API migration with external-caller impact analysis. |
| `DensityMode` | `DensityMode` | Core 01 §3.3.2.3 | `shared` | None; both are the exact scalar union `compact`, `default`, or `comfortable`. | Yes | Yes | Yes | Yes | `alias` | S-00 compile matrix, package runtime characterization, and S-03 consumer checks pass. |
| `AccountProfileResource` | `AccountProfileResource` | Core 01 §3.3.2.3 | `response` | Legacy properties are readonly; generated properties are writable. Members, requiredness, nullability, and scalars otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `AccountPreferencesResource` | `AccountPreferencesResource` | Core 01 §3.3.2.3 | `response` | Legacy properties are readonly; generated properties are writable. Members, requiredness, nullability, and scalars otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `AccountProfilePatchRequest` | `AccountProfilePatchRequest` | Core 01 §3.3.2.3 | `request` | Legacy properties are readonly; generated properties are writable. Closed members and scalars otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `AccountPreferencesPutRequest` | `AccountPreferencesPutRequest` | Core 01 §3.3.2.3 | `request` | Legacy properties are readonly; generated properties are writable. Required `density_mode` nullability and closed scalar vocabulary otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `AccountProfileEnvelope` | `AccountProfileEnvelope` | Core 01 §3.3.2.3 | `response` | Legacy envelope and nested resource are readonly; generated properties are writable. Shape otherwise matches. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `AccountPreferencesEnvelope` | `AccountPreferencesEnvelope` | Core 01 §3.3.2.3 | `response` | Legacy envelope and nested resource are readonly; generated properties are writable. Shape otherwise matches. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `ObjectBlobCreateRequest` | `ObjectBlobCreateRequest` | Core 01 §3.3.8, REQ-01-243 | `request` | Legacy properties are readonly; generated properties are writable. Optional nullable hints and required members otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `ObjectBlobUploadTarget` | `ObjectBlobUploadTarget` | Core 01 §3.3.8, REQ-01-244 | `response` | Legacy is readonly and generated is mutable; both require the same members and string-valued header map. | Yes | Yes | No | Yes | `retain_compatibility` | Preserve root compatibility until a separately authorized migration intentionally expands caller write capability. |
| `ObjectBlobCreateEnvelope` | `ObjectBlobCreateEnvelope` | Core 01 §3.3.8, REQ-01-244 | `response` | Legacy is recursively readonly and generated is mutable; members, requiredness, scalar values, and string-valued upload headers otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | Preserve root compatibility until a separately authorized migration intentionally expands caller write capability. |
| `EvidenceAttachBlobRequest` | `EvidenceAttachBlobRequest` | Core 01 §3.3.8, REQ-01-245 | `request` | Legacy properties are readonly; generated properties are writable. Closed members and scalars otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |
| `EvidenceAttachBlobEnvelope` | `EvidenceAttachBlobEnvelope` | Core 01 §3.3.8, REQ-01-245 | `response` | Both are open to additive row/cell members; legacy is recursively readonly, permits any string `view_schema_id`, and allows cells without `value`, while generated is mutable and narrower. | No | Yes | No | No | `retain_compatibility` | Preserve root compatibility indefinitely; use owner-correct generated operation responses internally. |
| `EvidenceHandleIssueRequest` | `EvidenceHandleIssueRequest` | Core 01 §16, REQ-01-459 | `request` | Both are the exact closed-empty type `Record<string, never>`. | Yes | Yes | Yes | Yes | `alias` | S-04 generic normalization and regeneration must continue to produce the exact closed-empty alias and reject member-bearing objects. |
| `EvidenceHandleEnvelope` | `EvidenceHandleEnvelope` | Core 01 §16 | `response` | Legacy envelope and nested data are readonly; generated properties are writable. Members, optional preview kind, nullability, and scalars otherwise match. | Yes | Yes | No | Yes | `retain_compatibility` | A separately authorized versioned root-API migration that intentionally expands caller write capability. |

## Additive row and cell projection repair

Core 01 REQ-01-034 requires clients to ignore unknown additive members inside
row and cell objects. S-04 repaired the authored OpenAPI projection so the
generated Core HTTP types and validators admit those members. That repair does
not authorize changing the more permissive root compatibility declarations or
making `ViewCell.value` optional in the canonical generated contract.

## Compatibility and migration impact

S-00 established the baseline and S-04 repaired owner projections without
removing root declarations. The compile fixture freezes the resulting external
caller acceptance, while repository consumer migrations use operation-specific
generated request and response types. Root declarations remain the stable
compatibility surface even after repository consumers stop using them.

Leaving these decisions implicit would allow apparently harmless structural
deduplication to change write capability, excess-property diagnostics, indexed
lookup, accepted header values, or closed-empty request behavior. Completion is
therefore binary: all seventeen rows exist, the compile fixture passes under the
real project, the runtime characterization passes, and no test reads this
record.
