# Structural Recovery completion and deferred Graph reconstruction

Status: explicitly adopted by the repository maintainer on 2026-11-01 and
applied to the owner sources. Implementation and readiness validation remain
separate.

The [Extensions tracker](extensions-module-refactor-tracker.md) records F-050
and the required owner amendment. This handoff is review material, never an
executable requirements input.

## Conflict and recommendation

Adopted Core 01 REQ-01-424, Extensions EXT-REQ-235, Graph §9 and Network Flow
NF-REQ-201/203 defer extension-derived Graph reconstruction until successful
claim. Core 01 REQ-01-625A still requires that reconstruction and its successful
participant result before Recovery readiness. Current journal v6 nests the
target-generation identity inside Graph completion, and current retry and
control completion require that result. Structural restore cannot honestly
supply it. Simply removing those checks would lose generation authentication.

Use current journal v7 with an explicit structural generation tuple. Remove
Graph completion from Recovery execution and evidence. Retain the current
Graph algorithm and result for actual post-claim reconstruction. Bind journal
v7 into the required current Recovery codec registry so earlier completion
generations cannot pass startup or replay through an unchanged codec identity.

This changes no target-control columns, migration numbers or snapshot shape.
Control v1, snapshot v3, marker v5 and verification v5 remain current. The
verification artifact does not contain a Graph completion field; its execution
currently inherits the eager Graph call through restore.

## Concrete Core 01 amendments

Replace the text of REQ-01-625A, retaining its identifier, profiles and
verification references, with:

> Core 01's `RestoreProjectionRebuilder`,
> `restore_projection_rebuild_request_v1`,
> `restore_projection_rebuild_result_v1`, and provider registry govern workbook
> projection rebuild only. Graph MUST NOT be registered as a workbook provider
> or receive an unrestricted workbook source reference through those contracts.
>
> Recovery MUST authenticate the selected frozen catalog and current Graph
> algorithm, source-registry and implementation-binding identities before target
> mutation. A retired, unresolved or mismatched binding rejects selection or
> admission. Resolving these immutable definitions MUST NOT construct or invoke
> an extension source adapter, enumerate its declarations, or execute Graph
> reconstruction. The current identities remain
> `graphprojection.restore_rebuild.v4`, source registry v4 and implementation
> binding v5; no historical dispatcher, translator or fallback is permitted.
>
> Under REQ-01-424's extension-derived exception, structural Recovery leaves the
> four Graph-derived tables empty. It invokes only the independent required Core
> projection rebuilds, in ascending ASCII algorithm-ID order. Structural
> completion certifies preserved authoritative state and those Core obligations;
> it MUST NOT contain a Graph participant success result or certify extension
> availability. Recovery records the exact structural generation independently
> of any later Graph result.
>
> After successful claim compatibility, supported migration, final semantic
> validation and publication admission, application composition invokes the
> Graph-owned current reconstruction for the claimed source under Graph §9.
> Exact selected-result reconstruction, Network Flow job reconciliation,
> Reporting job/lease reconciliation and postcondition checks share the
> caller-controlled transaction. Dependent gates open only after its determinate
> success. Its failure leaves those gates closed; an indeterminate commit uses
> Core 04 fatal-integrity shutdown. Unrelated admitted non-graph operations
> remain available. This later result does not rewrite Recovery completion.

In REQ-01-578, immediately after its restore-verification outcome requirements,
insert:

> Restore verification certifies structural Recovery and its independent Core
> rebuild, invariant and Workbook obligations. Extension-derived reconstruction
> remains deferred under REQ-01-424/625A. A verified backup does not certify
> subsequent profile claim or Graph availability. The current
> `cartulary.restore_verification.v5` artifact shape remains unchanged.

In REQ-01-648's durable target-control completion paragraph, replace
“plaintext current v6 terminal Recovery journal payload” with
“plaintext current v7 terminal Recovery journal payload”. The exact-byte hash,
independent durable admission, full-record revision comparison, required
attestation, original serving lease and absence of a runtime Recovery-key
requirement remain unchanged.

In the portable restore completion paragraph, replace “target bindings and
participant completion” with “target bindings and structural generation
completion”, and replace “Intent and current v6 terminal payloads” with
“Intent and current v7 terminal payloads”. Preserve immutable intent v1,
authenticated context, size bounds, target-audit reconciliation and exact
replay without payload mutation.

## Concrete Core 04 amendments

In REQ-04-113, replace the current journal-format sentence with:

> Current Recovery admission and terminal journal records use
> `cartulary.operator_recovery_journal_payload.v7`. Version 6 and older payloads
> require their matching release and MUST NOT be read, converted, rewritten or
> accepted as current completion or replay evidence. Current verification
> remains `cartulary.restore_verification.v5`. No old marker, completion or
> verification basis establishes current target admission.

Change the remaining current journal schema reference in REQ-04-113 from v6
to v7. Replace its two paragraphs beginning “When an admitted restore includes
the Graph participant” and “The encrypted terminal evidence format” with:

> Structural restore MUST retain its admitted operation, attempt, backup,
> consistency point, target generation and target bindings without invoking an
> extension-owned Graph participant. It MUST NOT mint a second operation or
> generation identity. Exclusive serving-lease ownership remains required
> through mutation, independent Core validation, durable journal/attestation
> publication, control completion and any permitted reset.
>
> A current terminal record replaces `graph_projection_completion` with the
> required member `restore_generation`. Its value is null or a closed object
> containing exactly `target_generation_id`,
> `recovery_state_catalog_sha256` and `codec_registry_sha256`. The generation is
> a canonical nonzero UUID; the digests are lowercase 64-character SHA-256
> strings and must resolve to the selected current Recovery generation. Existing
> top-level operation, attempt, backup, consistency-point and
> `target_binding_digests` members retain their meanings.
>
> Every terminal outcome for an admitted per-target restore or verification
> attempt MUST carry this exact generation object and all four target-binding
> digests, including a failed outcome. Successful per-target completion requires
> the matching backup and consistency point. Scheduler-only aggregate/no-op and
> non-restore records carry null generation and target bindings. A pre-admission
> failure cannot manufacture an admitted generation. Admission-record fields
> are otherwise unchanged. Unknown fields, removed Graph completion, malformed
> identities, noncanonical bytes and inconsistent tuples are invalid.
>
> Completion and exact retry MUST authenticate the canonical current terminal
> payload and match its generation, catalog, codec, target bindings and enclosing
> operation/attempt/backup identity to the admitted control record. Failed,
> conflicting or absent terminal proof cannot authorize completion. Exact
> successful replay changes only the matching terminal/control metadata and
> MUST NOT repeat payload mutation or invoke Graph. Structural completion does
> not assert Graph availability. The safe administrative-audit summary remains
> unchanged and exposes none of these private digests or source values.

Add after those paragraphs:

> The selected Recovery generation's required codec registry MUST include the
> current `operator_recovery_journal_payload` role, resolved from the authored
> current schema registry. Omitted, retired or unresolved journal codecs are
> invalid. This format cutover changes the codec-registry digest; an older
> control/backup/verification pair does not become current merely because other
> physical storage formats are unchanged. There is no historical reader,
> implicit conversion, automatic reset or data deletion.

The typed terminal-record member description must include the closed structural
generation and target-binding members above. Canonical encoding, envelope,
encryption, append-only journaling, atomic audit and exact-byte hashing rules
continue to apply. No second encoding is introduced for control hashing.

## Concrete Graph clarification

In Graph §9, after the paragraph describing one reconstruction attempt per
successful process claim activation, insert:

> Reconstruction of a current restored installation uses the original operation
> and target-generation identities from the admitted database control record,
> together with the current catalog, source-registry and implementation-binding
> definitions and the successful source claim. It MUST NOT require a Recovery
> decryption key, an unexpired marker, source backup storage, or a fabricated
> consistency timestamp. Source enumeration and reconstruction use the same
> caller-controlled transaction over the admitted authoritative state. Result
> publication time is operational state and does not change logical result
> identity or canonical output digests. A new initialized installation follows
> ordinary Graph admission; it has no restored generation to reconstruct.

Graph result v4 continues to describe actual reconstruction only. No empty or
synthetic participant result may stand in for the deferred work. Existing
unavailable contracts, immutable Reporting inputs, pending-job gates,
completed-output verification and fatal handling remain authoritative.

## Impact and validation

- Replace v6 authored journal schema/positive fixture with v7 and regenerate
  through Make. Add the required codec role and regenerate current generation
  identities; reject v6 and earlier without positive historical readers.
- Remove Graph completion fields, optional adapters and eager source calls from
  Recovery. Preserve independent Core rebuilds and the Extensions contribution.
- Verify canonical v7 bytes, complete tuple matching, failed admitted outcomes,
  exact replay, conflicting/missing proof and completion after response loss.
- Requalify current capture/restore/verification, portable source-loss retry,
  durable quarantine, completed startup without a Recovery key, and rejection
  of the prior codec generation. No database or environment reset is authorized.
- Prove inactive malformed-but-representable source state restores without
  Graph source execution, fails later semantic claim without reconstruction,
  and valid claimed state reconstructs/reconciles before dependent gates open.
- Prove one attempt per activation, restart after interruption, independent Base
  availability, exact pending Reporting input/lease reconciliation and
  completed-output verification without a live Graph lease.

These are implementation exits following adoption, not claims established by
this document or by Markdown lint. Full module validation and final handoff
remain required under the controlling tracker.
