# Extensions Recovery boundary amendment

Status: explicitly adopted by the user on 2026-10-31 (tracker E-107). Adoption
does not establish implementation readiness. The controlling execution artifact
is [the Extensions tracker](extensions-module-refactor-tracker.md).

## Problem and selected boundary

Extensions' adopted Table 23-F already requires shared-owner enumeration and
shared structural restore validation with no profile-code invocation. Rebuilds
belong to the declared physical binding and run after claim admission. Yet
EXT-REQ-038, Tables 16-A/16-B, EXT-REQ-232 and NF-REQ-185 also require or describe
a separate profile backup participant with enumeration, validation and rebuild
operations. Network Flow's packaged generic v1 participant is therefore both
obsolete and redundant. Its dedicated timeout has no execution consumer.

Retire the profile backup participant rather than introduce a replacement
participant whose operations must not be invoked. Preserve Extensions' Recovery
source contribution, current codec definitions, physical bindings, exact binding
proofs, Recovery target lifecycle, and claimed-binding rebuild ownership.

## Concrete owner amendments

Apply these adopted amendments together. No workflow identifiers belong in
the product specifications.

In the Extensions NLSpec:

1. Replace the descriptor rule requiring exactly one
   `backup_restore_participant` for `extension_versioned` profiles with:

   > Every `extension_versioned` profile MUST provide the physical-state binding
   > and current backup codecs required by EXT-REQ-234 and EXT-REQ-235. No profile
   > backup/restore participant contribution is admitted. Extensions contributes
   > those bindings to Recovery through its existing source-owned contribution.

2. Remove `backup_restore_participant` from the contribution vocabulary, Tables
   16-A/16-B, and EXT-REQ-094. Remove `backup_restore` from EXT-REQ-232's
   specialization-kind vocabulary and its Table 23-D row. Specializations remain
   the closed `incident_portability` and `snapshot_reporting` contracts. Keep
   cross-owner transaction participants under their separate typed contract.

3. Retain Table 23-F's shared enumeration, shared restore validation and
   post-claim binding-rebuild rules. Replace EXT-REQ-221's backup participant
   context/result variants with this shared binding-validation contract:

   > The shared storage validator receives the exact admitted physical binding,
   > current codec, expected binding proof and a read-only accessor restricted
   > to that binding. These process-local inputs are not a profile callback or a
   > serialized participant context. They MUST expose no other binding or
   > profile capability. Its result MUST contain exactly `schema_id`,
   > `binding_id`, `status` and `findings`. `schema_id` MUST equal
   > `cartulary.extension_binding_restore_validation_result.v1`; `binding_id`
   > MUST equal the requested binding. `status` MUST be `valid` with an empty
   > findings array or `invalid` with 1..256
   > `cartulary.extension_transaction_participant_finding.v1` entries. Findings
   > obey EXT-REQ-219's path, owner mapping, secrecy, uniqueness and ordering
   > rules. Unknown members, duplicate members, wrong types and invalid null
   > are rejected. Result classification obeys EXT-REQ-225, including overflow
   > precedence; canonical result bytes are bounded at 67108864. The caller
   > derives the owner from the admitted binding, never from returned content.
   > A failed result stops advancement to the next binding and follows the
   > Recovery failure/quarantine contract. No profile code may execute.

   In Table 23-E's following paragraph, replace the old backup participant
   result identity with the new binding-validation result identity. Preserve all
   existing required shared validator IDs and validation scopes. Enumeration
   returns the admitted physical-binding catalog through the Recovery source
   port. Rebuild dispatch uses each excluded derived binding's declared typed
   algorithm after successful claim admission; neither requires a generic
   participant result or a second operation registry.

4. Remove `timeouts.extensions.backup_restore_participant_seconds`. Recovery
   operations retain their Recovery-owned deadlines and cancellation behavior;
   later claim/rebuild work retains the existing lifecycle deadline. The removed
   key is unknown configuration, with no alias or replacement Extensions key.

5. State the current-only contract cutover: generic
   `cartulary.extension_participant_contract.v1` and the retired profile backup
   participant context/result and specialization kind are not current. Reject
   them at generation/admission. No reader, converter, alias or historical
   support is added. The remaining current specialization uses its existing
   unpublished v3 identity and changed canonical digest; no old digest pair is
   accepted.

In Network Flow NF-REQ-185, remove the mandatory
`network_flow_activity.backup_restore_v1` participant and replace that bullet
with:

> One physical-state binding and the current codecs required by NF-REQ-184;
> Extensions supplies them to Recovery. Network Flow declares no profile backup
> participant. Its derived Graph binding rebuilds through the Graph owner only
> after successful claim admission.

Keep the transaction and Snapshot/Reporting participants and their ownership.
The Core Recovery amendment adds this explicit boundary to its Extensions
contribution description:

> Recovery consumes Extensions' physical-binding contribution and shared
> structural validators. It MUST NOT resolve or invoke a profile backup
> participant. Derived rebuild and profile-semantic validation occur through the
> subsequent claim lifecycle under their existing owners.

## Implementation, impact and acceptance

The implementation slice removes the authored generic backup participant,
contribution and digest bindings, its catalog/parser/generator support, unused
configuration field/default/validation, and positive tests of that obsolete
contract. It projects the new shared result contract and regenerates all affected
artifacts through Make. Static incompatible fixtures remain only for rejection.
No production backup data is deleted, converted or reset. Registry, descriptor
and dependent implementation/binding digests change in the coherent unpublished
cutover; incompatible backups still reject before mutation.

The slice passes when current participant admission rejects the retired kind,
exact current contributor/reference sets agree, removed configuration rejects,
existing transaction/reporting participants retain their typed contracts, and
focused catalog/schema/generated/boundary/configuration checks pass. The later
Recovery implementation slice must prove capture/restore, per-binding structural
validation, failure quarantine and subsequent claim/rebuild behavior through
current entry points. This amendment does not claim those later exits passed.
