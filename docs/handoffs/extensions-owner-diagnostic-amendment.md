# Owner-input diagnostic contract amendment

Status: prepared 2026-11-01; explicitly adopted by the user and applied
2026-11-01 (tracker E-156). Implementation readiness remains unestablished.

Execution and evidence remain in the
[Extensions tracker](extensions-module-refactor-tracker.md), F-039 / WS-35.
This handoff is not an executable input or behavioral authority. The user explicitly adopted the concrete replacement below; its exact owner
edit is now applied.

## Conflict and recommendation

The paragraph following Table 11-A under EXT-REQ-062 in
[Extensions §11](../extension-subsystem-nlspec.md#11-required-runtime-dependencies)
requires bare owner IDs as `safe_ref` values. Owner IDs match
`^[a-z][a-z0-9_]{0,63}$`; they cannot satisfy EXT-REQ-233's mandatory
`<namespace>:<kind>:<opaque-id>` form. The same paragraph requires raw
`actual` strings that do not satisfy Table 26-D. Using only an owner reference
also leaves two profiles sharing that owner indistinguishable when the same
fact kind is missing.

Use a safe recognized-profile reference for a missing fact and a safe containing
fragment reference for an unrecognized fact. Retain the existing normalized
paths, generic reason/message and closed detail members. Use existing formatter
outputs; redact the unrecognized value. These changes remove the conflict and
identify each affected logical input without adding diagnostic fields, aliases,
or a new formatter language.

## Adopted owner replacement

Replace only the paragraph beginning “A missing required owner fact MUST use
path” under EXT-REQ-062 with the following three paragraphs:

> A missing required owner fact MUST use path `$.owner_fragments`, reason code
> `extension_registry_invalid`, and details containing
> `artifact_kind='owner_input_registry'`,
> `safe_ref='cartulary:profile:<profile_id>'`,
> `expected='exact:<fact_kind>'`, and `actual='missing'`. `<profile_id>` MUST be
> the exact recognized profile whose required fact is absent. `<fact_kind>`
> MUST be the exact closed fact-kind token required by its owner contract.
> The expected and actual values use `diagnostic_exact_token_v1` and
> `diagnostic_missing_v1`, respectively. Each missing `(profile_id, fact_kind)`
> identity MUST produce its own finding, including when multiple affected
> profiles have the same primary owner.
>
> An unrecognized owner fact MUST use the exact normalized path
> `$.owner_fragments[<fragment_index>].facts[<fact_index>].profile_id`, reason
> code `extension_registry_invalid`, and details containing
> `artifact_kind='owner_fragment'`,
> `safe_ref='cartulary:owner_fragment:<owner_fragment_id>'`,
> `expected='exact:recognized_profile_id'`, and `actual='redacted'`.
> `<owner_fragment_id>` MUST be the exact admitted identity of the containing
> fragment. The expected and actual values use `diagnostic_exact_token_v1` and
> `diagnostic_redacted_v1`, respectively. `<fragment_index>` and `<fact_index>`
> are the zero-based positions after the canonical sorting in EXT-REQ-177.
>
> These references are non-authorizing diagnostic identities that resolve only
> through the admitted recognized-profile or owner-fragment catalog. They MUST
> NOT derive from the offending value, a source-document locator or a physical
> storage identity. Both conditions MUST use the exact Table 26-B message and
> closed details object. They MUST NOT use `collision_class`,
> `conflicting_profile_ids[]`, `conflicting_tokens[]`, or `duplicate_count`.
> EXT-REQ-233 continues to govern malformed packaged bytes and their declared
> package-slot references.

No execution-workstream name or adoption narrative belongs in the owner text.

## Impact and validation

This is a current-only correction to incompatible diagnostic values. The
startup-finding object shape and generic reason code remain unchanged; consumers
must accept the adopted logical references and safe formatter outputs. No
stored-state migration, historical error alias or backup converter is introduced.
Authored projections and dependent digests change in the implementation slice.

Implementation requires current generator and runtime witnesses for:

- missing facts for two recognized profiles sharing an owner, with distinct
  profile references and deterministic finding order;
- an unrecognized fact at its exact normalized fragment/fact indices, with a
  valid containing-fragment reference and no offending value in diagnostics;
- closed details, exact message, registered formatters, safe-reference bounds
  and rejection of extra members;
- unchanged malformed-byte package-slot derivation and complete collision
  diagnostics under their separate adopted predicates.

Markdown lint and manual preservation review qualify this candidate for human
review. They do not adopt it or establish implementation readiness. Dependent
owner-fact projection changes are authorized and remain assigned to WS-36 after
its implementation prerequisites. Configuration, invocation lifetime and other
independent corrections retain their existing authority.
