# Current Reporting execution authority amendment

Status: adopted on 2026-11-02 under the user's standing amendment authorization.
Reviewed candidate SHA-256: `32df10cb4edda8095706b44f80574f1103414b76311a0f9a28c4553a2c0359b5`.
Core 01, Reporting and Report Composition owner text was applied before dependent
implementation. No user approval is pending.

## Problem and boundary

The current restored-job regression removes the initiating incident membership
before a deferred render becomes runnable. Reporting nevertheless publishes a
successful result. A current Jobs execution proves ownership and cancellation
state; it does not prove current actor authority.

Reporting must use the existing Core authorization owners through typed ports.
Auth owns current user activity; Incidents owns incident membership, role and
lifecycle. Jobs owns the immutable initiating identity and execution fence.
Reporting must not add an ACL, query another owner's tables directly, or ask an
inactive Graph source to authorize retained Reporting input.

## Concrete behavioral amendment

Add REQ-RPT-130 under Reporting §24.1, with coordinated references from Core 01
REQ-01-477/634 and Report Composition's preview authorization contract:

- Before processing private snapshot/render input on every attempt, derive the
  initiating actor, incident and job kind from the fenced, independently bound
  Jobs admission. Require an existing active user and current incident access.
  Snapshot and release creation require `editor`, `reviewer` or `admin`;
  composition preview permits every incident member, including `viewer`.
  These are the existing initiating-operation permissions. Deployment admin
  alone grants no incident access. Existing active and closed incident behavior
  remains allowed; a missing incident is not authorized.
- Repeat that check in the final success transaction, before owner effects,
  resource references, a success proof or terminal success can commit. Keep
  shared authority-row locks through commit so revocation and publication have
  a defined database order. Use Jobs transition lock, then incident,
  membership and Auth actor locks before owner publication. This preserves the
  existing Incidents and assessment authority order and avoids inversion with
  membership creation. Do not hold these locks during
  source materialization, participant calls or rendering subprocess work.
- Missing/inactive user, missing incident or membership, or insufficient role
  is a determinate nonretryable authority denial. Do not render further, publish
  output or create a success proof. Retain the initiating identity and exact
  input. Ordinary cancellation and failure cleanup remain permitted. A committed
  success cannot be rewritten because authority later changes; public replay
  and reads retain their current request-time authorization.
- Snapshot authority denial completes the common Job as failed with
  `error_summary.code='authorization_denied'`, message `Initiating actor is no
  longer authorized.`, and `retryable=false`. Release and preview denial use
  existing `release_render_failed` with `reason_code='initiating_actor_unavailable'`.
  Reporting's closed validation taxonomy adds failure code `authorization_denied`
  at stage `release_state` and reason code `initiating_actor_unavailable`, with
  no authority-specific safe-detail fields; existing lifecycle context markers
  remain permitted. Prepared/unprepared failed-attempt retention continues
  to follow the current owner lifecycle; no fabricated source/model facts.
- Storage/transport errors and cancellation are not authority denial. Preserve
  their owner classification, cancellation precedence, execution fencing and
  indeterminate-commit fatal handling.

## Projections and compatibility

Update the authored closed failure projections and regenerate through Make.
The inspected failure taxonomy is currently an authored Go projection; missing
packaged Reporting definitions remain part of F-030/WS-30 and must include the
new closed values before module readiness. This work must not invent a claim
that those schemas already exist. Core 01's unconditional failed-release-row
sentence must also be reconciled with the already adopted prepared/unprepared
Reporting lifecycle: only validated preparation can supply a failed release
row; otherwise retain truthful attempt evidence without fabricated facts.
Align Reporting Table 24-A and REQ-RPT-117c with that same distinction.
The current schema identities retain their existing shapes; their bound digests
and generated enums change where the new closed values are projected. No old
reader, alias, alternate writer or automatic data deletion is introduced.
This explicit authority-denial refinement supersedes the frozen-input amendment's
initial expectation that no additional reason code would be needed. It does not
reintroduce a live Graph or source-profile dependency.

## Review and validation

Review the exact roles against Core and Report Composition Table 6-A; review the
transaction order against existing Core owner locks and success finalization.
Check that authority denial can still commit a truthful failure without applying
the success-only permission guard to cancellation, cleanup or committed replay.

Require real current execution with early membership/role/user loss, revocation
while rendering before publication, successful authorized execution, allowed
closed-incident and preview-viewer cases, cancellation and exact committed
outcomes. Preserve current restore/source-unclaim evidence and input fidelity.
Assert no successful owner output or proof on denial and closed, secret-safe
failure details. Run focused owner portfolios, applicable generated/schema and
boundary checks, and final Markdown/preservation review. This document is manual
review evidence; executable requirements and tests must not consume Markdown.
