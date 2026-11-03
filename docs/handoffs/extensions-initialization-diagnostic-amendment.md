# Extensions initialization diagnostic amendment

Status: explicitly adopted by the user and applied 2026-11-01 (tracker E-173).
The [Extensions tracker](extensions-module-refactor-tracker.md) controls
execution under F-040 / WS-37 and WS-38. This handoff is not an executable
input or behavioral authority. Dependent projections remain required under the adopted amendment.

## Problem and recommendation

[Extensions EXT-REQ-234](../extension-subsystem-nlspec.md#21-unclaim-reclaim-retirement-and-state-migration)
defines fresh initialization and its migration-step budget. The earlier tracker
citation to EXT-REQ-216 was imprecise: that requirement defines migration and
validation interfaces, while EXT-REQ-234 defines initialization. Table 26-A/B
has no initializer failure or initialization timeout. Its migration timeout
admits only a numbered step or a whole-profile timeout. Current fallbacks can
therefore name a final validator that never ran or report the wrong budget.

Add two initialization-specific diagnostics with small closed details. Preserve
the existing migration and final-validation diagnostics for their own
operations. Do not expand the migration-timeout union to hide initialization
behind a migration message and nullable migration identities. Retain the exact
origin of the selected deadline across nested initialization/final-validation
calls; an inherited deadline does not necessarily originate at the profile
budget.

This clarifies an existing admitted initialization feature without adding an
operation, callback, timeout setting, compatibility alias or persisted format.

## Adopted Extensions owner amendment

Under EXT-REQ-234, replace the sentence beginning “The deadline MUST use” with:

> The local initialization deadline MUST use
> `timeouts.extensions.migration_step_seconds` and the fresh-initialization
> start/end boundaries in Table 9-E. Its effective deadline MUST be the earlier
> of that local deadline and the whole-profile migration deadline under
> EXT-REQ-215, with equality selecting the inherited deadline. The context's
> `deadline_monotonic_ns` MUST be a process-local integer in
> `1..9223372036854775807` representing that effective deadline.

After EXT-REQ-234's paragraph beginning “Initialization runs inside”, insert:

> A returned error, panic, cancellation, forbidden scoped access, or result
> outside `cartulary.extension_state_initialization_result.v1` from the
> initialization algorithm MUST fail initialization with
> `extension_state_initialization_failed`, except when an expired selected
> deadline or an admitted fatal condition requires the outcome below. The
> diagnostic's `algorithm_id` MUST identify the admitted initializer that was
> invoked. The runtime MUST NOT invoke final validation after such an
> initializer failure or name an unexecuted final validator in its diagnostic.
> A malformed or missing packaged initializer definition or implementation
> remains a registry/binding admission failure before invocation.
>
> Expiry of the selected local initialization budget MUST use
> `extension_state_initialization_timeout` and the effective configured
> `timeouts.extensions.migration_step_seconds` value. This rule covers both
> initialization variants and all work within the initialization transaction's
> Table 9-E interval, including rollback or commit outcome determination.
> Expiry of a selected inherited whole-profile budget MUST instead use
> `extension_migration_timeout` with `operation_kind='profile_migration'`,
> null migration identity/version members, and the effective configured
> `timeouts.extensions.profile_migration_seconds` value. The selected budget's
> origin MUST survive further deadline nesting. If a final-state validator's
> own validation budget expires first, its failure MUST use
> `extension_state_validation_failed` with that validator's admitted identity;
> if its inherited initialization or profile deadline expires first or ties,
> the corresponding initialization or profile timeout above applies. Report
> the selected budget's configured value, not remaining time, an unrelated
> budget or an inferred default.
>
> A forbidden empty-state result retains `extension_state_incomplete`. A
> failure from an actually invoked final validator retains its final-state
> validation or validation-result diagnostic under EXT-REQ-217/225 unless the
> selected enclosing deadline requires the timeout above. None of these
> diagnostics may fabricate a migration ID, migration version transition,
> algorithm identity or raw failure detail. EXT-REQ-215 continues to govern
> sampled cancellation and commit proof, and EXT-REQ-193 continues to govern
> fatal outcomes. A proven successful commit is not reclassified as an
> initialization timeout; unknown mutation outcome remains fatal. Every
> noncommitted initialization failure rolls back pending state and metadata,
> writes no migration ledger row, and prevents claim publication.

In Table 26-A, insert these rows before `extension_migration_lock_timeout`:

| `reason_code` | Required use |
| --- | --- |
| `extension_state_initialization_failed` | The invoked fresh-state initializer failed or returned an invalid readiness result. |
| `extension_state_initialization_timeout` | Fresh state initialization exceeded its selected local initialization budget. |

In Table 26-B, insert the matching rows before
`extension_migration_lock_timeout`:

| `reason_code` | Exact message | Exact closed details |
| --- | --- | --- |
| `extension_state_initialization_failed` | `Extension state initialization failed.` | `profile_id`, `algorithm_id`; both are non-null admitted identities. |
| `extension_state_initialization_timeout` | `Extension state initialization exceeded its timeout.` | `profile_id`, `timeout_seconds`; the timeout is the effective configured `timeouts.extensions.migration_step_seconds` value. |

In the detail-rules paragraph following Table 26-B, append:

> Both initialization diagnostics use path `$` and the existing
> `migration_validation` startup phase. Their identifiers come only from the
> admitted profile and initialization definition. Neither permits a
> `migration_id`, migration version member, final-validator identity, raw
> result, or raw exception in details. The initialization-timeout finding
> applies to `kind='empty'` and `kind='algorithm'`; only an actually invoked
> algorithm can produce `extension_state_initialization_failed`.

Replace the EXT-AC-133 acceptance row with:

| ID | Acceptance criterion |
| --- | --- |
| `EXT-AC-133` | Fresh, already-current, fully migrated, and restored paths invoke initialization and final validation exactly as specified. Initializer invocation/result failures identify the invoked initializer; local initialization, inherited profile and local final-validation expiry select the exact owning diagnostic and configured budget, including equal and nested deadlines. Failed initialization rolls back state and metadata without a ledger row or claim publication; fatal/commit-proof precedence and secret-safe closed details remain exact. |

Table 26-A1's migration-timeout `operation_kind` remains exactly
`migration_step` or `profile_migration`. No initialization variant is added to
that union. No workstream identifier or adoption narrative enters the owner.

## Adopted Core companion amendment

In [Core 04 REQ-04-144](../spec/04_security_deployment_and_conformance.md#13-coordinated-extensions-subsystem-process-contracts),
after its first paragraph about condition registration and result precedence,
insert:

> Core imports `extension_state_initialization_failed` and
> `extension_state_initialization_timeout` from Extensions EXT-REQ-234 and
> Table 26-B, with their exact messages, closed details and selected-deadline
> rules. These are startup admission findings under
> `error.code='invalid_deployment_config'`; an ordinary reported startup
> failure opens no listener and exits `2`. Existing cancellation, proven
> commit and fatal process-lifecycle precedence still applies. Core MUST NOT
> translate these findings into a numbered migration or an unexecuted
> final-validator failure.

No other Core envelope, exit code, diagnostic member, lifecycle transition or
configuration domain changes.

## Impact, risks and validation

The change belongs in the two specifications, authored validation conditions
and schemas, generated projections, runtime state diagnostics, focused tests
and handoff documentation. Initializer errors previously reported through
incorrect generic fallbacks change deliberately. There is no historical error
alias, state conversion, backup reader, environment reset or deployment.
Registry/schema digests change when authored projections are regenerated.

The long-term benefit is accurate failure ownership for fresh state, with
separate initialization, migration and final-validation contracts. Retaining
the gap would make new initializer implementations difficult to diagnose and
could conceal which timeout is misconfigured. The main implementation risk is
losing deadline origin across two levels of inheritance or overriding a proven
commit with a late timeout; implementation must keep those facts typed.

Required implementation evidence after adoption:

- Returned error, panic, cancellation, wrong schema/status and scoped-access
  rejection identify the invoked initializer and never invoke final validation.
- Empty and algorithm initialization exercise local initialization expiry,
  earlier/equal inherited profile expiry, and nested final-validation expiry.
  Each exact reason/message/detail object uses the selected configured budget.
- Final validation still identifies its own failures when actually invoked;
  forbidden empty state and pre-invocation registry failures retain their
  separate classifications.
- PostgreSQL witnesses prove rollback of pending state/metadata, no ledger row,
  no publication, safe lock lifetime, and committed/absent/unknown outcome
  precedence. Zero/nonzero grace still follows the adopted fatal contract.
- Registered outward findings pass packaged schemas; extra/missing details and
  unregistered conditions reject. Harness-owned secrecy checks establish that
  callback payloads and raw exceptions are absent from diagnostics.
- Regeneration uses Make, with catalog/schema/generated/boundary checks and
  narrow current state/startup selections. Markdown never participates in
  executable admission, generation, testing or release evidence.

The explicit user decision supplies adoption. Lint and preservation review do
not establish implementation or module readiness.
