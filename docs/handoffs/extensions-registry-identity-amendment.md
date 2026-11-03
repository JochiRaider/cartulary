# Registry normalization and owner-fact identity amendment

Status: adopted/current on 2026-11-02 under the maintainer's standing
authorization, after source review, passing candidate Markdown lint and
preservation checks. This amendment addresses F-067 / WS-68 and creates no
new approval gate.
This handoff is not an executable input or substitute behavioral authority.

## Gap and decision

Extensions §11 references EXT-REQ-177 for normalized diagnostic indices, and
Table 13-A references EXT-REQ-205 for owner-fact collision hashes. Neither
requirement is defined in the current owner. The authored definition set names
`cartulary.extension_owner_fact_identity.v1`, but supplies no exact logical
members or canonical byte/digest implementation. Generator and runtime instead
duplicate a private string sort key.

Define normalization and a new current identity v2 explicitly. Identity names
the logical fact, not its fragment, owner-document location or complete value.
This keeps duplicate scalar declarations identical even when their values
differ. More specific collision rules still take precedence. Group facts by
profile and kind for understandable ordering, then use the exact canonical
identity and full-fact bytes to make invalid duplicate input deterministic.
Do not coalesce duplicates to manufacture a valid input.

Keep owner-fragment v3 and owner-input-registry v2 shapes. Replace the undefined
identity-v1 definition with v2 and add its closed authored schema. Use a pure
Extensions-owned primitive in both generator and runtime, without importing
generated artifacts into that primitive or adding a generic validator facade.
No configurable encoder, historical identity reader or compatibility alias is
needed.

## Concrete Extensions source amendment

Add the following after EXT-REQ-176's owner-fragment/input-registry definition:

> **EXT-REQ-177**
> Owner-input normalization MUST operate only on structurally admitted typed
> owner fragments and facts. It MUST preserve every declared scalar and fact;
> it MUST NOT trim, case-fold, Unicode-normalize, infer, repair, coalesce or
> discard values or duplicate declarations.
>
> For each fragment, sort facts by the tuple `(profile_id, fact_kind,
> canonical_identity_bytes, canonical_fact_bytes)`, comparing each component
> lexicographically as unsigned UTF-8 bytes. The identity bytes are defined by
> EXT-REQ-205. The final component is the complete admitted fact serialized
> under `extension_registry_canonical_json_v1`; it deterministically orders
> different values with the same logical identity. Byte-identical duplicate
> facts remain separate entries. Their relative order is immaterial because
> their values and diagnostic identities are equal.
>
> After fact normalization, sort fragments by `(owner_id, owner_fragment_id,
> canonical_fragment_bytes)` using the same byte comparison. The final
> component is the entire fragment with its normalized fact array, serialized
> under `extension_registry_canonical_json_v1`. Preserve duplicate fragments
> for collision classification. A fragment's declared identity MUST satisfy
> its owner-fragment scalar contract before it can supply a diagnostic ref.
>
> Authored input order MUST NOT affect normalized bytes or diagnostic indices.
> Published owner-input registries MUST already have this canonical order;
> runtime admission MUST verify it using the same rules and MUST NOT silently
> reorder a noncanonical package into an admitted one. Unknown-profile findings
> use the normalized zero-based fragment/fact positions assigned here. Neither
> normalization nor a computed fact identity establishes recognized-profile,
> owner-allocation or executable-contract admission.
>
> Profiles: base
> Verified by: EXT-AC-004, EXT-AC-005, EXT-AC-021, EXT-AC-076
>
> **EXT-REQ-205**
> Current logical owner-fact identity is the closed object
> `cartulary.extension_owner_fact_identity.v2` containing exactly `schema_id`,
> `profile_id`, `fact_kind` and `identity`. `schema_id` MUST be exactly that v2
> identifier. `profile_id` and `fact_kind` MUST be the unchanged values of the
> structurally admitted fact. `identity` MUST be the closed object assigned
> by Table 5-B below, with every named member required and no other member.
> Values MUST satisfy their existing owner scalars; an unknown fact variant
> has no identity and MUST be rejected, never mapped to an empty identity.
>
> Serialize the entire object under `extension_registry_canonical_json_v1`:
> UTF-8 key ordering, exact integer rules, no BOM and exactly one final LF.
> `identity_sha256` MUST be the lowercase SHA-256 of those exact bytes,
> including LF, with no additional prefix, separator or length framing.
> The schema member supplies the version/domain binding. A fallback
> `owner_fact_identity` collision token is exactly
> `<fact_kind>:sha256:<identity_sha256>`.
>
> The identity excludes fragment identity, source path, owner-document data,
> complete fact values and physical storage names. These fields cannot make
> duplicate logical declarations distinct. A narrower Table 13-A collision
> suppresses the fallback for those declarations. Missing/unknown owner-input
> failures retain EXT-REQ-062's separate safe references and formatters; they
> MUST NOT expose an offending value or its derived identity hash.
>
> Only identity v2 is current. Identity v1 is unsupported and absent from
> current definitions and consumers. This transient definition adds no stored
> state migration, historical reader, digest replacement or automatic deletion.
>
> Profiles: base
> Verified by: EXT-AC-021, EXT-AC-076, EXT-AC-079

Add the following normative table with EXT-REQ-205:

| Table 5-B: `fact_kind` | Exact `identity` members | Value source |
| --- | --- | --- |
| `recognized_profile`, `claim_configuration`, `state_ownership`, `admission_validation`, `egress`, `portability`, `snapshot_reporting` | None: exactly `{}` | The outer profile/kind identifies each scalar fact. |
| `runtime_dependency` | `dependency_profile_id` | Same-named fact member. |
| `route_family` | `route_family` | Same-named fact member. |
| `workspace` | `workspace_key` | Same-named fact member. |
| `public_schema` | `public_schema_id` | Same-named fact member. |
| `contribution` | `contribution_id` | `contribution.contribution_id`. |
| `migration_definition` | `migration_id` | `migration_definition.migration_id`. |
| `migration_ledger_definition` | `migration_id` | `migration_ledger_definition.migration_id`. |
| `worker_kind` | `worker_kind` | Same-named fact member. |
| `job_kind` | `job_kind` | `job_kind_contract.job_kind`. |

Change Table 13-A's owner-fact identity schema reference to v2. Retain the
collision class, existing public finding shape, counts, profile membership,
registered messages, safe formatters and narrower-class precedence. Restore
no prohibited capability fact or other unsupported variant just to exercise a
collision class; those declarations still fail their existing admission rules.

Extensions document advances to 0.14.6; public contract major 2 is unchanged.
Core 00's current summary records this exact typed identity/normalization
contract. No workstream ID or execution procedure belongs in normative text.

## Compatibility and implementation boundaries

The old transient identity schema is retired coherently. The current owner
fragment and registry shapes remain unchanged; current valid normalized
content may remain byte-identical, which must be measured rather than assumed.
Generated definition/schema/provenance and dependent integrity identities
change. No public diagnostic alias or persisted data conversion is introduced.

The shared implementation must preserve caller-owned data and duplicate entries,
validate the existing identity scalars, and return explicit errors for malformed
or unsupported variants. Byte admission continues through the existing strict
reader. Source values used for normalization remain internal; unknown-profile
diagnostics expose only admitted fragment refs and normalized indices.

Generators consume authored machine inputs and regenerate through Make. The
runtime validates current packaged projections and ordering. Neither reads this
handoff or any other Markdown. Build provenance, logical fact identity and
runtime diagnostic presentation remain distinct responsibilities.

## Validation and exit

- Review the exact member/variant table and sorting rules against current owner
  allocation, admitted fact variants and diagnostic rules before adoption.
- Use independent byte and digest literals for scalar and keyed identities,
  including the schema/version and final LF. Cover all admitted variants.
- Reverse and permute fragment/fact inputs, including equal identities with
  different values, and require identical normalization without duplicate loss
  or caller mutation. Cover empty registry input and multiple fragments.
- Reject malformed identity members, unsupported variants, missing/retired
  current schema definitions and noncanonical published ordering. Existing
  strict byte readers continue to reject invalid Unicode and duplicate keys.
- Exercise the actual generator and coordinator, regenerate via Make, and run
  applicable schema, catalog, generated-drift and boundary checks. Record exact
  current identities and preservation before resuming WS-36 collision work.

## Adoption record

Review confirmed all 16 admitted fact variants and their existing scalar
contracts. The current schema set has no exact v1 identity shape; its sole
authored occurrence is a definition registration. Identity excludes fact values
so conflicting scalar declarations remain the same logical fact. Full canonical
bytes provide the duplicate tie-break without changing identity or suppressing
a collision. Existing narrower collision rules remain authoritative.

Candidate Markdown passed with native exit 0 under
`extensions-ws68-identity-candidate-markdown-20261102`. Preservation review
confirmed all 714 original tracker lines, 164 original companion handoffs,
tracker metadata, twelve sections, unchanged HEAD and empty staged diff.
The adopted clauses are in Extensions §5 and Core 00 §1.7. Product implementation
and current projection validation follow this source adoption; these checks do
not themselves confer adoption.

## Implementation and focused validation

The shared `internal/modules/extensions/registryfacts` primitive now supplies
logical identity and duplicate-preserving order to generation and runtime.
Runtime requires the packaged v2 schema, while the authored definition set
remains generation-only. The generator admits typed fragment shapes before
normalization; runtime rejects noncanonical packaged order. Both private string
sort-key copies are removed. Existing strict readers retain byte admission.

E-406 records passing generator/runtime selections and catalog, schema,
generated-drift and backend-boundary checks at the same source digest. Tests
include sixteen independent exact byte/hash vectors, current schema rejection,
permuted duplicate declarations, empty input, nested copy isolation and invalid
Unicode rejection. The adopted full route grammar is enforced by one shared
parser and the identity schema, including exact byte/segment limits.
The runtime selection preserves graph and immutable-catalog
regressions. Final Markdown/preservation checkpoint closes this prerequisite;
full diagnostic collection and module readiness remain separate work.

| Current artifact | SHA-256 |
| --- | --- |
| Registry integrity | `e44b6717954f5b5f5630772a2ba31f764cc766aa9bb02319220f24c438369d3e` |
| Identity v2 schema | `251ddc6b168474df8e977b7a45a9dac2c69a48301a92d8e91157177bb492343d` |
| Generation contract definitions | `31a159717a91bb5847500e199801bfbb8af86180e7a775a1bd30baf549578de4` |
| Authored generated-schema source set | `72dde587ba37774054e27b469710549c96a345ffe2a1862948a1c0a2211466c1` |

Valid owner fragment, owner-input registry, profile descriptor and registry,
implementation binding, state/backup/participant registry bytes remain unchanged.
The generated Recovery artifacts also remain unchanged. This cutover introduces
no stored-data conversion, environment reset or historical identity reader.
