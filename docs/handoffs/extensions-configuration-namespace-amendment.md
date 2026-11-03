# Extensions configuration namespace clarification

Status: adopted on 2026-11-03 under the maintainer's standing amendment
authorization, following concrete owner comparison, candidate Markdown lint
and preservation review (tracker E-413). This document is a manual handoff, never executable
input or product evidence.

The adopted EXT-REQ-207 distinguishes a profile ID from its explicitly declared
configuration namespace. Reference Pack currently declares `reference_packs`,
while its profile and claim key remain `reference_pack` and
`reference_pack.claimed`. Table 4-B's namespace scalar and Table 13-A's collision
predicate still equate the two identities. Applying those older tables rejects
the current authored Reference Pack contract and defeats the explicit namespace
design already adopted in version 0.12.0.

Amend Extensions document 0.14.7 as follows:

- Replace the configuration-namespace scalar row with: "The exact
  `configuration_namespace` declared by the current profile configuration
  contract under EXT-REQ-207; it is distinct from the profile ID and claim key."
- Replace Table 13-A's `configuration_namespace` predicate with: "Duplicate
  exact admitted `configuration_namespace` across configuration-contract
  declarations." Retain the existing duplicated-namespace token, grouped
  profiles/counts, and `extension_registry_conflict` reason family.
- Retain EXT-REQ-207 unchanged, including namespace-local keys, exact current
  names and absence of alias namespaces. Its local key-membership requirements
  apply before a namespace supplies a collision token. The configuration
  contract's schema and recognized profile do not alone establish that closure.
- Record the clarification in Core 00's current owner summary without workflow
  IDs or implementation details.

This resolves conflicting owner text; it introduces no new field, schema,
namespace, alias, stored-state migration or historical reader. Current Reference
Pack configuration remains valid. Producer/runtime collision validation must use
the declared namespace, reject invalid local key membership without exposing
unadmitted values, and preserve complete participating profiles and declaration
counts. Required implementation evidence belongs to the existing registry
diagnostic workstream, independently of this source-only clarification.

Review criteria: compare both table replacements with EXT-REQ-207 and the current
Core 04/Reference Pack namespace declarations; retain exact claim identity and
current-only policy; pass Markdown lint and the tracker/history preservation
review before recording adoption and editing dependent code.
