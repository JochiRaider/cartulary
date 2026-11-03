# Durable Recovery target admission amendment

Status: explicitly adopted by the repository maintainer on 2026-11-01; applied
to the owner sources. Implementation and readiness validation remain separate.

The [Extensions tracker](extensions-module-refactor-tracker.md) controls this
required correction under F-048, WS-48 and WS-47. The amended owner sources
are behavioral authority; this handoff is not an executable input. The reviewed
candidate was explicitly adopted before dependent implementation.

## Problem and recommendation

Core 01 REQ-01-648 and Core 04 REQ-04-113 require an admitted failed or interrupted
restore target to remain non-serving. Production currently has no durable
startup gate after the operator releases its lease. A flag under backup storage
would lose authority when that root changes. Restored journal history must not
become the lifecycle authority for the current target.

Add one Recovery-owned database control record, independent of captured state.
Fence actual-state inspection and payload mutation, durably close target
admission before journaling or restoring data, and open it only after exact
durable completion. Application credentials can read this record but cannot
change it. Retain the current admission marker, authenticated completion
records, serving leases and subsequent profile claim lifecycle.

## Concrete Core 01 amendments

In REQ-01-661, after the sentence describing Network Flow's version 30 increment,
insert:

> Recovery appends version 35 after the already allocated owner increments
> 31–34. It adds `public.recovery_target_state`, owned by `module.recovery`,
> without rewriting versions 1–34 or deleting retained data. This table has
> state class `recovery_metadata`, backup inclusion
> `excluded_recovery_metadata`, restore action `ignore`, and null snapshot
> codec and restore-algorithm identities. It is installation-local admission
> authority, never restored source content.

In the same requirement, replace `exactly 147 table entries` with
`exactly 148 table entries`. Keep exactly 114 authoritative-required entries,
the synthetic Goose classification, existing ownership and all six object
families. Add after its public-behavior preservation sentence:

> The Recovery target admission required by REQ-04-113 is an expressly
> authorized startup lifecycle change. It adds no public product route or
> incident payload.

Under REQ-01-647, immediately before its object-family paragraph, insert:

> `recovery_target_state` MUST appear exactly once under `module.recovery` with
> the exclusion declared in REQ-01-661. Capture and restore MUST NOT export,
> import, truncate or replace this record as backup content. Ordinary restore
> and disposable verification reset preserve its identity and use the Recovery
> owner lifecycle to change its state. A snapshot containing a unit for this
> excluded table MUST reject before payload mutation.

Under REQ-01-648, after the paragraph ending “automatic target deletion or reset
is forbidden.”, insert:

> Recovery MUST maintain exactly one current database-owned target control
> record with logical schema `cartulary.recovery_target_state.v1`. Its exact
> members are `schema_id`, `state`, `revision`, `target_generation_id`,
> `purpose`, `operation`, `operation_id`, `attempt_id`, `backup_set_id`,
> `recovery_state_catalog_sha256`, `codec_registry_sha256`, `binding_digests`
> and `completion_evidence_sha256`. `revision` is a positive signed 64-bit
> integer and MUST increase on each transition without wrapping. UUID values
> are canonical non-zero UUIDs; SHA-256 values are lowercase hexadecimal.
> `binding_digests` uses the exact four non-secret identities of the current
> restore target marker. Missing, extra, malformed or unsupported control
> state MUST fail closed. Physical allocation remains machine-owned.
> `operation` is one of `restore_latest`, `restore_bundle`,
> `restore_verify_latest` or `restore_verify_due`. `attempt_id` follows the
> current terminal journal contract and is non-null only for
> `restore_verify_due`; it is required for each such admitted attempt. Restore
> operations use `restore_target`; verification operations use
> `restore_verification_target`.
>
> `state` is exactly `unproven`, `initialized`, `verification_ready`,
> `restoring`, `complete` or `resetting`. `unproven` and `initialized` have null operation,
> attempt, generation, purpose, binding, backup, catalog, codec and completion
> fields. `initialized` records admitted fresh initialization; it does not
> assert that the running deployment remains empty. A disposable verification
> reset may return to `verification_ready` while retaining only the exact
> verification-target generation, purpose and binding digests; other identity
> fields are null. `resetting` has those same bound verification fields and
> null operation, attempt, backup, catalog, codec and completion fields.
> `restoring` requires every identity field subject to the
> attempt rule above and has a null completion digest. `complete` requires the
> same identity fields and its exact successful completion digest. Purposes
> remain exactly `restore_target` and `restore_verification_target`. The
> migration owner creates and seeds the singleton. The application role has
> only `SELECT`; the Recovery role has only `SELECT` and `UPDATE`, with no
> insertion, deletion or truncation authority on this table.
>
> A new full database initialization, admitted as empty before DDL and not yet
> issued its application crypto-format identity, MUST establish initially
> `initialized` control state. Incremental adoption into an already initialized
> database without trustworthy control history MUST preserve its rows and
> establish `unproven` state. Absence of the new record, an old marker, restored
> journal history, empty-table heuristics or an application restart MUST NOT
> establish historical completion. No automatic reset, data deletion, old
> readiness reader, converter or administrative “mark healthy” bypass is
> introduced. Previously initialized databases needing retained-data
> continuity require separately authorized operational treatment; this
> contract does not silently bless them.
>
> After authenticating the complete selected input and the stopped-empty
> target, Recovery MUST acquire its existing exclusive target capability and
> fence the target data tables for the pending restore transaction. It MUST
> recheck actual freshness and bound mappings under that fence before admission
> journaling or payload mutation. Only `initialized` state, or the exact
> owner-admitted `verification_ready` generation, permits a new restore.
> Recovery MUST durably publish `restoring` with the exact operation, backup,
> generation, catalog, codec and target bindings before its admission journal
> and first payload mutation. This publication is installation-local admission
> metadata. A failure publishing or journaling admission MUST NOT allow payload
> mutation. A stale capability or concurrent writer MUST NOT bypass the
> inspection. Normal application writes do not change `initialized` state;
> a new restore still requires the independent actual stopped-empty inspection.
> The restoring-state commit is independent of the fenced payload transaction;
> payload rollback cannot reopen the target. Data fencing MUST NOT truncate or
> replace the protected control record.
> The data transaction and lease are mandatory owner capabilities,
> not optional notifications.
>
> An admitted failure, cancellation, lease loss, interruption or indeterminate
> outcome leaves control state non-complete across process death and restart.
> A preflight rejection before durable admission does not change an otherwise
> healthy target. Recovery may publish `complete` only after all structural
> restore obligations and independent Core checks pass, exact successful
> terminal journal/attestation evidence is durable, and exclusive ownership is
> still proved. `completion_evidence_sha256` is SHA-256 over the exact canonical
> plaintext current v6 terminal Recovery journal payload. It is not a digest of
> ciphertext or a newly defined encoding. Publication compares the admitted
> revision and exact identity tuple; mismatches fail closed. The application
> must not require a Recovery decryption key to read this control record.
>
> An exact successful retry may reconcile terminal metadata and complete the
> same control record under exclusive admission after authenticating the
> matching durable completion without conflicting terminal evidence. It MUST
> NOT restore payload again, renew the
> generation, or accept a merely started/failed operation. An indeterminate,
> cancelled or lease-lost target is not automatically reused. Only the existing
> explicitly disposable verification reset may clear verified disposable
> payload and return control to its bound `verification_ready` state, under
> exclusive admission and after reset postconditions pass. A failed reset
> remains non-serving. Ordinary targets are never reset automatically.
> Before any disposable reset mutation, Recovery MUST durably enter `resetting`
> and invalidate the previous operation/completion tuple. No prior completion
> can reconcile a resetting target. Reset may begin only for an exactly bound
> disposable verification target; an initially unbound initialized target also
> requires actual pristine admission. Only successful reset postconditions may
> produce `verification_ready`; an interrupted reset remains `resetting` and
> may continue only through the same explicitly disposable owner reset rules.
> A `verification_ready` generation may be renewed only under the existing
> marker-admission rules, under exclusive admission and after actual pristine
> revalidation with the same target bindings. Renewal updates the marker and
> control identity before a new attempt; it never promotes incomplete state.
>
> Capture MUST read this control state in its existing read-only repeatable-read
> snapshot and accept only `initialized` sources and exactly completed ordinary
> restore targets. A disposable verification target cannot become an application
> deployment or capture source merely by passing verification. This admission
> does not execute profile semantics. Structural Recovery completion does not
> establish a profile's compatibility, migration, validation, derived
> reconstruction or dependent-operation readiness; the subsequent claim
> lifecycle continues to own those obligations.

## Concrete Core 04 amendments

Under REQ-04-113, after its application shared-serving-lease paragraph, insert:

> After acquiring the application side of the serving lease and before profile
> initialization, route/worker activation or listener publication, the server
> MUST read the current database-owned Recovery target control state. Only
> `initialized` state or `complete` ordinary restore state with matching current target
> bindings and admitted current generation identities permits startup to
> continue. Disposable verification state, unproven/restoring/resetting state,
> missing or malformed state, and a completion mismatch MUST report
> `recovery_target_incomplete`, start no listener or worker, and exit `2`.
> This is a Core process-lifecycle admission condition, not invalid deployment
> configuration. Its path is `$`, its details are exactly `{}`, and its message
> is `Recovery target completion is not established; application startup is closed.`
> Changing a filesystem root, restarting the application, or replaying restored
> source history MUST NOT bypass this database-owned admission gate. Application
> credentials have read-only access; only the admitted Recovery owner may
> perform lifecycle transitions. No condition output may expose stored
> identities, raw storage paths, credentials or database error text.
> Startup compares the stored catalog/codec identities and target binding
> digests with the current installation. The admitted target generation is
> retained by this control record; startup does not require an unexpired
> restore-admission marker or access to the terminal payload's decryption key.

Replace REQ-04-113's sentence requiring every timed-out, cancelled or lease-lost
target to be reinitialized with:

> A timed-out, cancelled or lease-lost admitted restore or verification MUST
> remain non-serving unless the exact durable successful terminal proof permits
> completion reconciliation under REQ-01-648. A missing, conflicting or failed
> terminal proof never permits reconciliation and the target MUST be
> reinitialized before reuse. Reconciliation updates only terminal admission
> metadata under exclusive ownership; it never repeats payload mutation or
> automatically resets data.

In AC-544, replace `Recovery remains exactly 147/114` with
`Recovery remains exactly 148/114` and require explicit Recovery-only update
authority for target control state. Extend AC-428 and AC-534 with evidence that admitted
failed/interrupted restores remain non-serving across operator exit, application
restart and backup-root changes; exact completion/replay opens the gate only
after durable proof; and disposable reset preserves the control identity.

## Concrete Extensions companion amendments

In EXT-REQ-213's paragraph beginning “Before listener startup”, add:

> After acquiring the serving lease and before profile initialization or
> publication, application composition MUST enforce Core 04's database-owned
> Recovery target admission. `recovery_target_incomplete` starts no listener
> or worker and exits `2`; it does not run profile initialization to repair
> incomplete Recovery state.

Add `recovery_target_incomplete` alongside `recovery_serving_lease_active` in
§26's Core process-lifecycle envelope exception. Add it to Table 26-A with the
condition “Current database-owned Recovery completion is absent, invalid or
incomplete.” Add it to the formatter table with the exact Core message above
and no values; details are exactly `{}`. Extend EXT-AC-104 to require this
startup rejection before profile mutation or publication.

## Impact and validation

This is a coordinated current-only cutover. Authored migration 35, table and ACL
allocations, Recovery contribution/catalog, generation bindings, diagnostic
definitions, runtime admission, test initialization/reset support and generated
artifacts must change together through Make. Baseline versions 1–29 and existing
increments 30–34 remain byte-identical. The catalog changes from 147/114 to
148/114; earlier incompatible catalog/generation pairs reject. No deployment,
environment reset or data deletion is part of this effort.

An existing initialized database without trustworthy control evidence retains
its data but becomes unproven and cannot serve or create a new successful
capture. This is an explicit compatibility cost, not an implied migration or
readiness conversion. Fresh initialization and current restored deployments
gain a durable, role-protected admission authority independent of configuration
paths, captured data and profile semantics.

Validation must cover fresh initialization and data-preserving unproven upgrade;
runtime write denial; exact catalog/codec exclusion; a racing database writer;
journal failure; interruption before and after PostgreSQL commit; lease loss;
object/rebuild failure; stale or malformed state; changed backup roots; actual
startup rejection before profile mutation/listeners/workers; successful current
restore and exact replay; current disposable reset and failed reset; and later
claim admission. Positive and rejection paths must use the shipped entry points
with real PostgreSQL, storage and processes where those boundaries are involved.
